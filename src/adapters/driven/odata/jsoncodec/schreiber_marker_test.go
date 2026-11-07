package jsoncodec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// mkMarker ist der Text, den der Leser als Marker liest, wenn sein Rohtext so lautet (SPEC-126 Nr. 10 f).
const mkMarker = "[REDACTED]"

// mkRune ist ein Zeichen jenseits von ASCII aus seinem Codepunkt (AGENTS.md §3.18).
func mkRune(cp rune) string { return string([]rune{cp}) }

// mkTexte erzeugt die Menge M1: das Element E (der Marker) und die Nachbarn N, die ihm ähneln und nicht verändert werden dürfen.
// Die Nachbarn sind 142 Einträge (Abweichung 90, Rest 36, Rest-Länge 4, Escape-Folge im Text 12); vier Einträge nennen einen Text,
// den ein anderer Eintrag auch nennt (Streichung an Stelle 0 gleich Suffix 9, an Stelle 9 gleich Präfix 9, `x` davor und dahinter
// als Rest gleich der Einfügung vor Stelle 0 und 10).
func mkTexte() []string {
	e := mkMarker
	n := []string{""}
	for k := 1; k <= 9; k++ {
		n = append(n, e[:k])
	}
	for k := 1; k <= 9; k++ {
		n = append(n, e[len(e)-k:])
	}
	for i := range 10 {
		for _, r := range []string{"x", mkRune(0x0415), mkRune(0xFF3B), mkRune(0x1F600)} {
			n = append(n, e[:i]+r+e[i+1:])
		}
	}
	for i := range 10 {
		n = append(n, e[:i]+e[i+1:])
	}
	for i := 0; i <= 10; i++ {
		n = append(n, e[:i]+"x"+e[i:])
	}
	for i := 1; i <= 8; i++ {
		n = append(n, e[:i]+strings.ToLower(e[i:i+1])+e[i+1:])
	}
	n = append(n, strings.ToLower(e), e+e)
	for _, z := range []string{"x", " ", "\t", "\n", "\r", mkRune(0xA0), "\x00", mkRune(0xFEFF), "[", "]", `"`, bs} {
		n = append(n, z+e, e+z, z+e+z)
	}
	for _, l := range []int{4096, 65537} {
		n = append(n, strings.Repeat("x", l)+e, e+strings.Repeat("x", l))
	}
	for i := range 10 {
		n = append(n, e[:i]+esc(fmt.Sprintf("%04x", e[i]))+e[i+1:])
	}
	return append([]string{e}, append(n, esc("005B")+e[1:], bs+e)...)
}

// mkEnc ist die Ausgabe eines Textes als String, wie SPEC-169 Nr. 1 und Nr. 4 sie nennen; der Orakel-Kodierer arbeitet byteweise
// und teilt keinen Code mit dem Schreiber.
func mkEnc(text string) string {
	if text == mkMarker {
		return `"` + esc("005b") + `REDACTED]"`
	}
	var b strings.Builder
	b.WriteByte('"')
	for i := range len(text) {
		switch c := text[i]; {
		case c == '"' || c == '\\':
			b.WriteString(bs + string(c))
		case c < 0x20:
			b.WriteString(esc(fmt.Sprintf("%04x", c)))
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// mkAdresse kodiert einen Schlüssel byteweise nach SPEC-169 Nr. 2.
func mkAdresse(schl string) string {
	var b strings.Builder
	for i := range len(schl) {
		switch c := schl[i]; {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', strings.IndexByte("-._~',=", c) >= 0:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// mkStrings zerlegt Bytes in die Rohtexte aller Strings: ein Anführungszeichen öffnet und schließt, ein Backslash überspringt das
// nächste Byte. Der Zerleger ist das Orakel für „kein String der Ausgabe ist roh der Marker“ und teilt keinen Code mit Schreiber und Leser.
func mkStrings(b string) []string {
	var out []string
	innen, start := false, 0
	for i := 0; i < len(b); i++ {
		switch c := b[i]; {
		case !innen && c == '"':
			innen, start = true, i+1
		case innen && c == '\\':
			i++
		case innen && c == '"':
			out = append(out, b[start:i])
			innen = false
		}
	}
	return out
}

// mkModus ist ein Modus der Ausgabe: Stufe und Zahlform aus Accept (SPEC-169 Nr. 3).
type mkModus struct {
	name   string
	accept []string
	keine  bool
}

func mkModi() []mkModus {
	return []mkModus{
		{"minimal Zahl", nil, false},
		{"minimal Zeichenkette", []string{"application/json;IEEE754Compatible=true"}, false},
		{"none Zahl", []string{"application/json;odata.metadata=none"}, true},
		{"none Zeichenkette", []string{"application/json;odata.metadata=none;IEEE754Compatible=true"}, true},
	}
}

// mkFeld ist eine Property einer Probe: ihr Wert und die Ausgabe des Wertes.
type mkFeld struct {
	name string
	art  model.WertArt
	text string
	json string
}

// mkProbe ist eine Entity mit der Ausgabe ihrer Properties.
type mkProbe struct {
	typ, schl string
	felder    []mkFeld
}

func (p mkProbe) entity() model.Entity {
	e := model.Entity{Typ: p.typ, Schluessel: p.schl}
	for _, f := range p.felder {
		e.Properties = append(e.Properties, pw(f.name, f.art, f.text))
	}
	return e
}

// gelesen ist die Entity, die der Leser zum geschriebenen Body liefern muss.
func (p mkProbe) gelesen() model.Entity {
	e := p.entity()
	e.Uebersprungen = []model.Uebersprungen{}
	return e
}

func mkStr(name, text, json string) mkFeld { return mkFeld{name, model.ArtString, text, json} }

func mkLiteral(text string) string { return "'" + strings.ReplaceAll(text, "'", "''") + "'" }

// mkStelle ist eine der sieben Stellen von M2, an denen der Text steht.
type mkStelle struct {
	name       string
	set, typ   string
	sammlung   bool
	schluessel bool
	proben     func(text, enc string) []mkProbe
}

func mkStellen() []mkStelle {
	alle := func(schl string, f ...mkFeld) mkProbe { return mkProbe{"Ns.Alle", schl, f} }
	k := mkStr("S", "k", mkEnc("k"))
	stellen := []mkStelle{
		{"Nichtschlüssel", "Alles", "Ns.Alle", false, false, func(text, enc string) []mkProbe {
			return []mkProbe{alle("'k'", k, mkStr("NS", text, enc))}
		}},
		{"Schlüssel", "Alles", "Ns.Alle", false, true, func(text, enc string) []mkProbe {
			return []mkProbe{alle(mkLiteral(text), mkStr("S", text, enc))}
		}},
		{"beide", "Alles", "Ns.Alle", false, true, func(text, enc string) []mkProbe {
			return []mkProbe{alle(mkLiteral(text), mkStr("S", text, enc), mkStr("NS", text, enc))}
		}},
		{"abgeleitet", "Produkte", "Ns.Product", false, false, func(text, enc string) []mkProbe {
			return []mkProbe{{"Ns.Sonder", "1", []mkFeld{{"Id", model.ArtInt32, "1", "1"}, mkStr("Extra", text, enc)}}}
		}},
	}
	for i := range 3 {
		stellen = append(stellen, mkStelle{fmt.Sprintf("Entities[%d]", i), "Alles", "Ns.Alle", true, false, func(text, enc string) []mkProbe {
			var ps []mkProbe
			for j, s := range []string{"a", "b", "c"} {
				ns := mkStr("NS", "k", mkEnc("k"))
				if j == i {
					ns = mkStr("NS", text, enc)
				}
				ps = append(ps, alle("'"+s+"'", mkStr("S", s, mkEnc(s)), ns))
			}
			return ps
		}})
	}
	return stellen
}

// objekt ist die Ausgabe einer Entity nach SPEC-169 Nr. 5; ktx ist der Kontext, leer bei einem Element der Collection.
func (s mkStelle) objekt(p mkProbe, m mkModus, ktx string) string {
	var teile []string
	if !m.keine {
		if ktx != "" {
			teile = append(teile, `"@odata.context":"`+ktx+`"`)
		}
		if p.typ != s.typ {
			teile = append(teile, `"@odata.type":"#`+p.typ+`"`)
		}
		teile = append(teile, `"@odata.id":"`+wurzel+"/"+s.set+"("+mkAdresse(p.schl)+`)"`)
	}
	for _, f := range p.felder {
		teile = append(teile, `"`+f.name+`":`+f.json)
	}
	return "{" + strings.Join(teile, ",") + "}"
}

// erwartet ist der Body der Probe in der Stufe und Zahlform des Modus (SPEC-169 Nr. 5 und 6).
func (s mkStelle) erwartet(m mkModus, ps []mkProbe) string {
	if !s.sammlung {
		ktx := ""
		if !m.keine {
			ktx = wurzel + "/$metadata#" + s.set + "/$entity"
		}
		return s.objekt(ps[0], m, ktx)
	}
	var es []string
	for _, p := range ps {
		es = append(es, s.objekt(p, m, ""))
	}
	kopf := ""
	if !m.keine {
		kopf = `"@odata.context":"` + wurzel + "/$metadata#" + s.set + `",`
	}
	return "{" + kopf + `"value":[` + strings.Join(es, ",") + "]}"
}

// schreibe schreibt die Proben als Entity oder als Collection; ein Fehler ist ErrUngueltig, und dann steht nichts in dst.
func (s mkStelle) schreibe(t *testing.T, was string, m mkModus, ps []mkProbe) (string, error) {
	t.Helper()
	z := neuesZiel()
	var err error
	if s.sammlung {
		var es []model.Entity
		for _, p := range ps {
			es = append(es, p.entity())
		}
		_, err = schreiber().SchreibeCollection(context.Background(), z, testModell(), model.CollectionAntwort{Wurzel: wurzel, Set: s.set, Entities: es, Accept: m.accept})
	} else {
		_, err = schreiber().SchreibeEntity(context.Background(), z, testModell(), model.EntityAntwort{Wurzel: wurzel, Set: s.set, Entity: ps[0].entity(), Accept: m.accept})
	}
	if err != nil {
		istUngueltig(t, err, was)
		z.leer(t, was)
		return "", err
	}
	z.nurEinWrite(t, was)
	return z.body(), nil
}

// mkLauf führt f für jedes Element texte × Stellen × Modi mit gültigem Body aus und liefert die Zahl der Elemente und die der Elemente,
// die ErrUngueltig sind. Ungültig ist genau das Element, dessen Schlüssel-Property U+0000 trägt (SPEC-168 Nr. 3 b).
func mkLauf(t *testing.T, texte []string, modi []mkModus, f func(was string, s mkStelle, m mkModus, ps []mkProbe, body string)) (n, ungueltig int) {
	t.Helper()
	for i, text := range texte {
		for _, s := range mkStellen() {
			for _, m := range modi {
				was := fmt.Sprintf("Text %d (%d Bytes), Stelle %s, %s", i, len(text), s.name, m.name)
				ps := s.proben(text, mkEnc(text))
				body, err := s.schreibe(t, was, m, ps)
				erwartetFehler := s.schluessel && strings.Contains(text, "\x00")
				n++
				switch {
				case erwartetFehler && err == nil:
					t.Errorf("%s: kein Fehler, will ErrUngueltig", was)
				case erwartetFehler:
					ungueltig++
				case err != nil:
					t.Errorf("%s: %v", was, err)
				default:
					f(was, s, m, ps, body)
				}
			}
		}
	}
	return n, ungueltig
}

// TestMarkerBytes hält Z01 (SPEC-169 Nr. 4): der Text `[REDACTED]` steht mit `[` als Escape an jeder Stelle und in jedem Modus.
func TestMarkerBytes(t *testing.T) {
	n, ungueltig := mkLauf(t, []string{mkMarker}, mkModi(), func(was string, s mkStelle, m mkModus, ps []mkProbe, body string) {
		if want := s.erwartet(m, ps); body != want {
			t.Errorf("%s:\n got %s\nwant %s", was, body, want)
		}
		if s.schluessel && !m.keine && !strings.Contains(body, `/Alles('%5BREDACTED%5D')"`) {
			t.Errorf("%s: Adresse nicht in der Form von Nr. 2: %s", was, body)
		}
	})
	if n != 28 || ungueltig != 0 {
		t.Errorf("Elemente %d, ungültig %d, will 28 und 0", n, ungueltig)
	}
}

// TestMarkerKeinRohtext hält Z02: kein String der Ausgabe ist roh der Marker, für jeden Text von M1 an jeder Stelle in jedem Modus.
func TestMarkerKeinRohtext(t *testing.T) {
	body := `{"a":"x` + bs + `"y` + bs + bs + `","b":"` + mkMarker + `"}`
	if got := mkStrings(body); !reflect.DeepEqual(got, []string{"a", "x" + bs + `"y` + bs + bs, "b", mkMarker}) {
		t.Fatalf("Zerleger: %q", got)
	}
	texte := mkTexte()
	n, ungueltig := mkLauf(t, texte, mkModi(), func(was string, _ mkStelle, _ mkModus, _ []mkProbe, body string) {
		for _, s := range mkStrings(body) {
			if s == mkMarker {
				t.Errorf("%s: ein String der Ausgabe ist roh der Marker", was)
			}
		}
	})
	if len(texte) != 143 || n != 4004 || ungueltig != 24 {
		t.Errorf("Texte %d, Elemente %d, ungültig %d, will 143, 4004 und 24", len(texte), n, ungueltig)
	}
}

// mkJSONWerte dekodiert den Body mit encoding/json und prüft, dass jede Property der Art string den Text der Probe trägt.
func mkJSONWerte(t *testing.T, was string, s mkStelle, ps []mkProbe, body string) {
	t.Helper()
	var wurzelObj map[string]any
	if err := json.Unmarshal([]byte(body), &wurzelObj); err != nil {
		t.Errorf("%s: kein JSON: %v", was, err)
		return
	}
	objekte := []any{wurzelObj}
	if s.sammlung {
		objekte, _ = wurzelObj["value"].([]any)
	}
	if len(objekte) != len(ps) {
		t.Errorf("%s: %d Objekte, will %d", was, len(objekte), len(ps))
		return
	}
	for i, p := range ps {
		o, _ := objekte[i].(map[string]any)
		for _, f := range p.felder {
			if got, ok := o[f.name].(string); f.art == model.ArtString && (!ok || got != f.text) {
				t.Errorf("%s: Property %s hat %d Bytes, will %d", was, f.name, len(got), len(f.text))
			}
		}
	}
}

// TestMarkerNachbarn hält Z03: jeder andere Text, auch jeder, der dem Marker gleicht, bleibt in der Form von SPEC-169 Nr. 1.
func TestMarkerNachbarn(t *testing.T) {
	n, ungueltig := mkLauf(t, mkTexte()[1:], mkModi(), func(was string, s mkStelle, m mkModus, ps []mkProbe, body string) {
		if want := s.erwartet(m, ps); body != want {
			t.Errorf("%s: Body weicht ab (%d Bytes, will %d)", was, len(body), len(want))
		}
		mkJSONWerte(t, was, s, ps, body)
	})
	if n != 3976 || ungueltig != 24 {
		t.Errorf("Elemente %d, ungültig %d, will 3976 und 24", n, ungueltig)
	}
}

// TestMarkerRundlauf hält Z04 (SPEC-169 Nr. 8): ODataCodec liest Art, Text, Schlüssel und Übersprungenes gleich, für jeden Text von M1
// an jeder Stelle in beiden Zahlformen der Stufe minimal.
func TestMarkerRundlauf(t *testing.T) {
	leser := jsoncodec.New()
	n, ungueltig := mkLauf(t, mkTexte(), mkModi()[:2], func(was string, s mkStelle, _ mkModus, ps []mkProbe, body string) {
		if !s.sammlung {
			got, err := leser.ParseEntity(context.Background(), strings.NewReader(body), testModell(), s.set, grenzeLesen)
			if want := ps[0].gelesen(); err != nil || !reflect.DeepEqual(got, want) {
				t.Errorf("%s Entity: %v, gleich: %v", was, err, reflect.DeepEqual(got, want))
			}
			return
		}
		want := model.EntitySeite{}
		for _, p := range ps {
			want.Entities = append(want.Entities, p.gelesen())
		}
		got, err := leser.ParseCollection(context.Background(), strings.NewReader(body), testModell(), s.set, grenzeLesen)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%s Collection: %v, gleich: %v", was, err, reflect.DeepEqual(got, want))
		}
	})
	if n != 2002 || ungueltig != 12 {
		t.Errorf("Elemente %d, ungültig %d, will 2002 und 12", n, ungueltig)
	}
}

// TestMarkerFehlerobjekt hält Z11: kein Feld des Fehlerobjekts eines Codes des Katalogs ist roh der Marker.
func TestMarkerFehlerobjekt(t *testing.T) {
	codes := model.AlleCodes()
	n := 0
	for _, c := range codes {
		f, err := model.NeuesFehlerobjekt(c)
		z := neuesZiel()
		if err == nil {
			_, err = schreiber().SchreibeFehler(context.Background(), z, f)
		}
		if err != nil {
			t.Errorf("Code %d: %v", int(c), err)
			continue
		}
		raw := mkStrings(z.body())
		if len(raw) != 7 {
			t.Errorf("Code %d: %d Strings, will 7", int(c), len(raw))
			continue
		}
		for i, feld := range []string{"code", "message", "target"} {
			if raw[2+2*i] == mkMarker {
				t.Errorf("Code %d, Feld %s: roh der Marker", int(c), feld)
			}
			n++
		}
	}
	if len(codes) == 0 || n != len(codes)*3 {
		t.Errorf("Elemente %d, will %d mal 3", n, len(codes))
	}
}

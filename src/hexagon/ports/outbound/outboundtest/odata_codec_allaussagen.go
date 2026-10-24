package outboundtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// jfGrund sind die Bodys von R (E1, C1) und die Fälle eines unbekannten Sets.
func jfGrund() []jfall {
	out := []jfall{basisR(true), basisR(false)}
	for _, einzel := range []bool{true, false} {
		for _, set := range []string{"Nope", "products", "Demo.Products", ""} {
			if set == "" {
				continue
			}
			b := basisR(einzel)
			b.name, b.set, b.ursache, b.ort = "A1 Set "+set, set, model.DatenEntitySet, set
			out = append(out, b)
		}
	}
	return out
}

// alleFaelle sind die Fälle aller Zeilen der Tabelle: gültige und fehlerhafte; A4, A5, H3, H4 und J1 laufen über sie.
func (s *jsuite) alleFaelle(*testing.T) []jfall {
	if s.faelle == nil {
		s.faelle = verbinde(jfGrund(), jfB1(), jfB2(), jfB3(), jfB4(), jfB5(), jfB6(), jfC1(), jfC2(), jfC3(),
			jfD1(), jfD2(), jfE1(), jfE2(), jfE3(), jfE4(), jfE5(), jfF1(), jfF2(), jfF3(), jfF4(),
			jfF5(), jfF6(), jfF7(), jfF8(), jfZ1(), jfF9(), jfF10(), jfF11(), jfG1(), jfG2(), jfG3(), h1Faelle(), h2Faelle(), jfI3(), jfK1())
	}
	return s.faelle
}

func jH3(t *testing.T, s *jsuite) {
	ursachen := map[model.DatenUrsache]int{}
	fehler := 0
	for _, f := range s.alleFaelle(t) {
		if f.ursache == "" {
			continue
		}
		r := s.lauf(f)
		var df *model.DatenFehler
		if !errors.As(r.err, &df) {
			t.Errorf("%s: Fehler %v ist kein *DatenFehler", f.name, r.err)
			continue
		}
		ursachen[df.Ursache]++
		fehler++
		switch {
		case !hatUrsache(df.Ursache):
			t.Errorf("%s: Ursache %q gehört nicht zu den fünfzehn", f.name, df.Ursache)
		case df.Code() != model.CodeE6101 || !errors.Is(r.err, model.ErrDatenUngueltig):
			t.Errorf("%s: Code %v, ErrDatenUngueltig %v", f.name, df.Code(), errors.Is(r.err, model.ErrDatenUngueltig))
		case !ortGueltig(df):
			t.Errorf("%s: Ort %q der Ursache %s ist kein gültiger Pointer, oder nicht leer im Lesen", f.name, df.Ort, df.Ursache)
		}
		r.nullwerte(t, f.name)
	}
	for _, u := range alleUrsachen() {
		if ursachen[u] == 0 {
			t.Errorf("H3: die Ursache %s kommt in keinem Fehlerfall vor", u)
		}
	}
	t.Logf("H3: %d Ursachen in %d Fehlerfällen", len(ursachen), fehler)
}

func alleUrsachen() []model.DatenUrsache {
	return []model.DatenUrsache{
		model.DatenSyntax, model.DatenKodierung, model.DatenTiefe, model.DatenZuGross, model.DatenDoppelterName, model.DatenStruktur,
		model.DatenKontext, model.DatenODataTyp, model.DatenUnbekannteProperty, model.DatenTyp, model.DatenWert, model.DatenNull,
		model.DatenZahlZuLang, model.DatenSchluessel, model.DatenEntitySet,
	}
}

func hatUrsache(u model.DatenUrsache) bool {
	for _, x := range alleUrsachen() {
		if x == u {
			return true
		}
	}
	return false
}

// ortGueltig prüft die Form des Orts: leer bei den Ursachen des Lesens, bei entity_set beliebig (der Name), sonst ein
// JSON-Pointer nach RFC 6901 (leer, oder `/…`, `~` nur als `~0` oder `~1`).
func ortGueltig(df *model.DatenFehler) bool {
	switch df.Ursache {
	case model.DatenSyntax, model.DatenKodierung, model.DatenTiefe, model.DatenZuGross, model.DatenDoppelterName:
		return df.Ort == ""
	case model.DatenEntitySet:
		return df.Ort != ""
	}
	if df.Ort == "" {
		return true
	}
	if df.Ort[0] != '/' {
		return false
	}
	for i := 0; i < len(df.Ort); i++ {
		if df.Ort[i] == '~' && (i+1 == len(df.Ort) || (df.Ort[i+1] != '0' && df.Ort[i+1] != '1')) {
			return false
		}
	}
	return true
}

// artZuTyp ist die Art, die ein Wert der Property haben darf (neben `null` und `redigiert`).
func artZuTyp(edm string) model.WertArt {
	return map[string]model.WertArt{
		"Edm.String": model.ArtString, "Edm.Boolean": model.ArtBoolean, "Edm.Int32": model.ArtInt32, "Edm.Int64": model.ArtInt64,
		"Edm.Decimal": model.ArtDecimal, "Edm.Guid": model.ArtGUID, "Edm.Date": model.ArtDate, "Edm.DateTimeOffset": model.ArtDateTimeOffset,
	}[edm]
}

func jH4(t *testing.T, s *jsuite) {
	n := 0
	for _, f := range s.alleFaelle(t) {
		if f.ursache != "" {
			continue
		}
		n++
		r := s.lauf(f)
		if r.err != nil {
			t.Errorf("%s: Fehler %v", f.name, r.err)
			continue
		}
		if !f.einzel && r.seite.Entities == nil {
			t.Errorf("%s: Entities ist nil", f.name)
		}
		for _, e := range r.entities(f) {
			s.pruefeEntity(t, f, e)
		}
	}
	t.Logf("H4: %d gültige Fälle", n)
}

// pruefeEntity prüft eine gelesene Entity gegen das Modell: Slices nicht nil, Typ und Schlüssel gesetzt, jede Property mit
// unterstütztem Typ und passender Art, nichts Nicht-Unterstütztes in Properties.
func (s *jsuite) pruefeEntity(t *testing.T, f jfall, e model.Entity) {
	t.Helper()
	if e.Properties == nil || e.Uebersprungen == nil || e.Typ == "" || e.Schluessel == "" {
		t.Errorf("%s: Entity unvollständig: %+v", f.name, e)
		return
	}
	typ, ok := s.modelle[f.m].EntityType(e.Typ)
	if !ok {
		t.Errorf("%s: Typ %q steht nicht im Modell", f.name, e.Typ)
		return
	}
	for _, w := range e.Properties {
		var prop model.Property
		for _, p := range typ.Properties {
			if p.Name == w.Name {
				prop = p
			}
		}
		art := artZuTyp(prop.EdmType)
		if prop.Name == "" || !prop.Unterstuetzt() || (w.Wert.Art != art && w.Wert.Art != model.ArtNull && w.Wert.Art != model.ArtRedigiert) {
			t.Errorf("%s: Property %q mit Art %q passt nicht zum Typ %q des Modells", f.name, w.Name, w.Wert.Art, prop.EdmType)
		}
	}
}

// i1Entities liefert die 64 Elemente von I1: Schlüssel in nicht sortierter Folge, je zwölf Properties in einer Folge, die
// von der des Modells abweicht und mit dem Element rotiert, mit den erwarteten Entities.
func i1Entities() (body []byte, soll []model.Entity) {
	var elems []string
	for i := range 64 {
		id := (i*37+11)%101 + 1000
		rot := []struct{ json, name, text, art string }{
			{`"RString":"s` + strconv.Itoa(i) + `"`, "RString", "s" + strconv.Itoa(i), "string"}, {`"RDecimal":1.50`, "RDecimal", "1.5", "decimal"},
			{`"Dbl":2.5`, "Dbl", "", ""}, {`"Id":` + strconv.Itoa(id), "Id", strconv.Itoa(id), "int32"}, {`"NBoolean":true`, "NBoolean", "true", "boolean"},
			{`"RInt64":"9007199254740993"`, "RInt64", "9007199254740993", "int64"}, {`"Tags":["a"]`, "Tags", "", ""},
			{`"NString":null`, "NString", "", "null"}, {`"RInt32":7`, "RInt32", "7", "int32"}, {`"value":"v"`, "value", "v", "string"},
			{`"Cx":{"a":1}`, "Cx", "", ""}, {`"NDecimal":3e-1`, "NDecimal", "0.3", "decimal"},
		}
		e := model.Entity{Typ: "T.W", Schluessel: strconv.Itoa(id), Properties: []model.PropertyWert{}, Uebersprungen: []model.Uebersprungen{}}
		var ms []string
		for k := range rot {
			r := rot[(k+i)%len(rot)]
			ms = append(ms, r.json)
			if r.art == "" {
				e.Uebersprungen = append(e.Uebersprungen, model.Uebersprungen{Name: r.name, Grund: model.UebersprungTyp})
			} else {
				e.Properties = append(e.Properties, pw(r.name, model.WertArt(r.art), r.text))
			}
		}
		elems = append(elems, "{"+strings.Join(ms, ",")+"}")
		soll = append(soll, e)
	}
	return jKol("Ws", elems...), soll
}

// jI1: gleiche Bytes ergeben gleiche Werte in der Reihenfolge des Bodys (SPEC-126 Nr. 15), zwanzigmal im Prozess.
func jI1(t *testing.T, s *jsuite) {
	body, soll := i1Entities()
	for lauf := range 20 {
		r := s.lauf(jfall{m: "W", set: "Ws", body: body})
		if r.err != nil || !reflect.DeepEqual(r.seite.Entities, soll) {
			t.Fatalf("Lauf %d: Fehler %v, Entities weichen von der Reihenfolge des Bodys ab", lauf, r.err)
		}
	}
}

func jfI3() []jfall {
	var out []jfall
	el := func(id, rest string) string { return `{"Id":` + id + rest + `}` }
	for _, f := range []struct {
		name  string
		elems []string
		keys  string
	}{
		{"gleiche Properties", []string{el("1", `,"NString":"a"`), el("1", `,"NString":"a"`)}, "1,1"},
		{"verschiedene Werte", []string{el("1", `,"NString":"a"`), el("1", `,"NString":"b"`)}, "1,1"},
		{"verschiedene Property-Mengen", []string{el("1", `,"NString":"a"`), el("1", `,"RInt32":3`)}, "1,1"},
		{"Schreibweise des Schlüssels", []string{el("1", ""), el("1.0", "")}, "1,1"},
		{"erstes und letztes Element", []string{el("1", ""), el("2", ""), el("3", ""), el("1", "")}, "1,2,3,1"},
	} {
		fall := jfall{name: "I3 " + f.name, m: "W", set: "Ws", body: jKol("Ws", f.elems...)}
		fall.nach = func(t *testing.T, r jergebnis) {
			var keys []string
			for _, e := range r.seite.Entities {
				keys = append(keys, e.Schluessel)
			}
			if strings.Join(keys, ",") != f.keys {
				t.Errorf("I3 %s: Schlüssel %v, erwartet %s: Dubletten werden weder entfernt noch gemeldet", f.name, keys, f.keys)
			}
		}
		out = append(out, fall)
	}
	return out
}

func jI3(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfI3()) }

// jI2: das Ergebnis gehört dem Aufrufer, leere Slices sind nie nil, Eingabe und Modell bleiben (SPEC-126 Nr. 12).
func jI2(t *testing.T, s *jsuite) {
	leer := s.lauf(jfall{m: "W", set: "Ws", body: jKol("Ws")})
	if b, err := json.Marshal(leer.seite.Entities); err != nil || string(b) != "[]" {
		t.Errorf("I2: leere Entities serialisieren als %s (%v), erwartet []", b, err)
	}
	body, soll := i1Entities()
	hashVor := sha256.Sum256(body)
	m := s.modelle["W"]
	typenVor, nsVor := m.EntityTypes(), m.Namespaces()
	r := s.lauf(jfall{m: "W", set: "Ws", body: body})
	if sha256.Sum256(body) != hashVor || !reflect.DeepEqual(m.EntityTypes(), typenVor) || !reflect.DeepEqual(m.Namespaces(), nsVor) {
		t.Error("I2: Eingabe oder Modell haben sich geändert")
	}
	zweite := s.lauf(jfall{m: "W", set: "Ws", body: body})
	for i := range zweite.seite.Entities { // das Ergebnis gehört dem Aufrufer: jede Änderung bleibt ohne Wirkung auf andere Ergebnisse und Lesungen
		e := &zweite.seite.Entities[i]
		e.Typ, e.Schluessel = "x", "x"
		for k := range e.Properties {
			e.Properties[k].Name, e.Properties[k].Wert.Text = "x", "x"
		}
		for k := range e.Uebersprungen {
			e.Uebersprungen[k].Name = "x"
		}
		e.Properties = append(e.Properties[:0], model.PropertyWert{})
	}
	zweite.seite.Entities = zweite.seite.Entities[:0]
	if !reflect.DeepEqual(r.seite.Entities, soll) {
		t.Error("I2: eine Änderung der zweiten Lesung wirkt auf das Ergebnis der ersten")
	}
	if again := s.lauf(jfall{m: "W", set: "Ws", body: body}); !reflect.DeepEqual(again.seite.Entities, soll) {
		t.Error("I2: eine Änderung des Ergebnisses wirkt auf die nächste Lesung")
	}
	ein := s.lauf(jfall{m: "W", set: "Ws", einzel: true, body: jEin("Ws", `"Id":1`)})
	if b, err := json.Marshal(ein.entity); err != nil || strings.Contains(string(b), "null") {
		t.Errorf("I2: Entity serialisiert als %s (%v)", b, err)
	}
}

// j1Faelle sind die Fälle von J1: die kleinen Bodys der Tabelle.
func j1Faelle(t *testing.T, s *jsuite) []jfall {
	var out []jfall
	for _, f := range s.alleFaelle(t) {
		if len(f.body) <= 4096 {
			out = append(out, f)
		}
	}
	return out
}

// jJ1: 16 Goroutinen hinter einer Startbarriere durchlaufen die Fälle, den Abbruch des Kontexts und den Fehler von src; der
// Detektor von `make test-race` meldet jeden Zugriff ohne Ordnung.
func jJ1(t *testing.T, s *jsuite) {
	faelle := j1Faelle(t, s)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i, f := range faelle {
				r := s.lauf(f)
				if (r.err == nil) != (f.ursache == "") {
					t.Errorf("%s: Fehler %v, erwartet Ursache %q", f.name, r.err, f.ursache)
				}
				if i%25 == 0 {
					s.ruf(abgebrochen(), strings.NewReader(string(f.body)), f)
					s.ruf(context.Background(), &endlichStrom{inner: &teilReader{teile: [][]byte{f.body[:len(f.body)/2]}, ende: errQuelle}}, f)
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

// jJ2: 16 Goroutinen lesen zugleich vier verschiedene Bodys je 50-mal; jedes Ergebnis ist gleich dem einzelnen.
func jJ2(t *testing.T, s *jsuite) {
	bodys := []jfall{
		basisR(false), {m: "K", set: "Lines", einzel: true, body: jEin("Lines", `"OrderID":5,"LineNo":2,"Note":"n"`)},
		{m: "Z", set: "Animals", body: jKol("Animals", `{"@odata.type":"#Zoo.Dog","Id":1,"Breed":"x","dyn":1}`, `{"Id":2}`)},
		{m: "R", set: "Products", body: jKol("Products", `{"ID":1,"Name":5}`), ursache: model.DatenTyp, ort: "/value/0/Name"},
	}
	soll := make([]jergebnis, len(bodys))
	for i, f := range bodys {
		soll[i] = s.lauf(f)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 50 {
				for i, f := range bodys {
					if got := s.lauf(f); !reflect.DeepEqual(got, soll[i]) {
						t.Errorf("Body %d: %+v, allein gelesen %+v", i, got, soll[i])
						return
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

func jfK1() []jfall {
	var out []jfall
	for c := range 0x20 {
		out = append(out, jfall{name: fmt.Sprintf("K1 Name mit %#02x", c), m: "R", set: "Products", einzel: true,
			body: jEin("Products", fmt.Sprintf(`"ID":1,"a\u%04x":1`, c)), ursache: model.DatenUnbekannteProperty, ort: "/a" + string([]byte{byte(c)})})
	}
	for _, n := range []struct{ json, ort string }{
		{`a/b`, "/a~1b"}, {`a~b`, "/a~0b"}, {`~1`, "/~01"}, {`a~/b`, "/a~0~1b"}, {``, "/"},
	} {
		out = append(out, beide("K1 Name "+n.json, "R", "Products", `"ID":1,"`+n.json+`":1`, model.DatenUnbekannteProperty, n.ort)...)
	}
	for _, i := range []int{10, 123} {
		elems := make([]string, i+1)
		for k := range elems {
			elems[k] = `{"ID":1}`
		}
		elems[i] = `{"ID":1,"Nope":1}`
		out = append(out, jfall{name: fmt.Sprintf("K1 Index %d", i), m: "R", set: "Products", body: jKol("Products", elems...),
			ursache: model.DatenUnbekannteProperty, ort: fmt.Sprintf("/value/%d/Nope", i)})
	}
	return append(out, jfK1Zeichen()...)
}

// k1Namen sind die Namen der Menge von K1 als Fälle der Suite (SPEC-126 Nr. 14: kein anderes Zeichen als `~` und `/` wird
// ersetzt): jedes Zeichen c von U+0000 bis U+007F als `c`, `~c` und `/c` (3 × 128 = 384) und je ein Zeichen aus U+0080,
// U+07FF, U+0800, U+FFFF, U+10000 und U+10FFFF (6) als Vertreter der Menge ab U+0080 für A4, A5, H3, H4 und J1: 390 Namen.
func k1Namen() []string {
	var out []string
	for c := range 0x80 {
		z := string(rune(c))
		out = append(out, z, "~"+z, "/"+z)
	}
	for _, r := range []rune{0x80, 0x7FF, 0x800, 0xFFFF, 0x10000, 0x10FFFF} {
		out = append(out, string(r))
	}
	return out
}

// k1ZahlASCII ist die Zahl der Namen aus U+0000 bis U+007F in k1Namen; k1ZahlOben die Zahl der Zeichen von U+0080 bis
// U+10FFFF ohne die Surrogate U+D800 bis U+DFFF, die k1Oben fährt; beide zusammen 1 112 320.
const (
	k1ZahlASCII = 3 * 128
	k1ZahlOben  = 0x110000 - 0x80 - (0xE000 - 0xD800)
	k1ZahlAlle  = k1ZahlASCII + k1ZahlOben
)

// k1Auswahl sagt, welche Zeichen ab U+0080 der Lauf fährt: alle; unter dem Detektor eine Probe (U+0080 bis U+00FF, die Ränder
// jeder UTF-8-Länge, U+07FF, U+0800, U+D7FF, U+E000, U+FFFF, U+10000 und U+10FFFF, und jedes 97. Zeichen).
func k1Auswahl(r rune) bool {
	if !unterDetektor {
		return true
	}
	switch r {
	case 0x7FF, 0x800, 0xD7FF, 0xE000, 0xFFFF, 0x10000, 0x10FFFF:
		return true
	}
	return r <= 0xFF || (r-0x80)%97 == 0
}

// k1Oben liefert die Körper der Menge ab U+0080: eine Entity von R, deren unbekannte Property das Zeichen als Namen trägt.
func k1Oben() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for r := rune(0x80); r <= 0x10FFFF; r++ {
			if istSurrogat(r) || !k1Auswahl(r) {
				continue
			}
			if !yield(jEin("Products", `"ID":1,"`+string(r)+`":1`)) {
				return
			}
		}
	}
}

// k1OrtPruefer prüft einen Körper von k1Oben, ein Mitglied einer Entity von R: `unbekannte_property`, Ort `/` und der Name
// (RFC 6901, Orakel k1Orakel).
func (s *jsuite) k1OrtPruefer(kopf, fuss string) gPruefer {
	return func(b []byte, _ int64) (klasse, abweichung string) {
		name := string(b[len(kopf) : len(b)-len(fuss)])
		r := s.aufruf(context.Background(), true, bytes.NewReader(b), s.modelle["R"], "Products", 1<<30)
		var df *model.DatenFehler
		if ort := k1Orakel(name); !errors.As(r.err, &df) || df.Ursache != model.DatenUnbekannteProperty || df.Ort != ort {
			return "unbekannte_property", fmt.Sprintf("Zeichen %U: Ergebnis %v, erwartet Ort %q", []rune(name)[0], r.err, ort)
		}
		return "unbekannte_property", ""
	}
}

// k1Orakel bildet den Ort eines Mitglieds zeichenweise nach RFC 6901: `~` wird `~0`, `/` wird `~1`, jedes andere Zeichen bleibt.
func k1Orakel(name string) string {
	var b strings.Builder
	b.WriteByte('/')
	for _, r := range name {
		switch r {
		case '~':
			b.WriteString("~0")
		case '/':
			b.WriteString("~1")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// k1JSON schreibt den Namen als JSON-String: ein Steuerzeichen als Escape, `"` und `\` mit Backslash, jedes andere Zeichen roh.
func k1JSON(name string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range name {
		switch {
		case r < 0x20:
			b.WriteString(esc(fmt.Sprintf("%04x", r)))
		case r == '"' || r == '\\':
			b.WriteString("\\" + string(r))
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// jfK1Zeichen sind die Fälle der Menge k1Namen je Methode: ein Name mit `@` ist `struktur`, sonst eine unbekannte Property, der
// Ort ist der Name nach dem Orakel.
func jfK1Zeichen() []jfall {
	var out []jfall
	for _, n := range k1Namen() {
		u := model.DatenUnbekannteProperty
		if strings.Contains(n, "@") {
			u = model.DatenStruktur
		}
		out = append(out, beide("K1 Zeichen "+strconv.QuoteToASCII(n), "R", "Products", `"ID":1,`+k1JSON(n)+`:1`, u, k1Orakel(n))...)
	}
	return out
}

// istSurrogat sagt, ob die Zahl ein Surrogat U+D800 bis U+DFFF ist, das als Name kein gültiges UTF-8 ist.
func istSurrogat(r rune) bool { return r >= 0xD800 && r <= 0xDFFF }

// jK1Oben fährt die Menge ab U+0080: jedes Zeichen bis U+10FFFF ohne Surrogate als Name; zusammen mit den 384 Namen aus U+0000
// bis U+007F sind es 1 112 320 (SPEC-126 Nr. 14: kein anderes Zeichen als `~` und `/` wird ersetzt). Unter dem Detektor läuft
// die Probe von k1Auswahl, und die Zahl ist die der gewählten Zeichen.
func jK1Oben(t *testing.T, s *jsuite) {
	kopf := `{` + kontextAn("Products", true) + `,"ID":1,"`
	fuss := `":1}`
	erwartet := int64(k1ZahlOben)
	if unterDetektor {
		erwartet = 0
		for r := rune(0x80); r <= 0x10FFFF; r++ {
			if !istSurrogat(r) && k1Auswahl(r) {
				erwartet++
			}
		}
	}
	gZaehle(t, "K1 Zeichen ab U+0080", gFahre(t, "K1 oben", k1Oben(), s.k1OrtPruefer(kopf, fuss)), erwartet)
	if !unterDetektor && k1ZahlASCII+erwartet != 1112320 {
		t.Errorf("K1: %d Namen der Menge, erwartet 1 112 320", k1ZahlASCII+erwartet)
	}
}

func jK1(t *testing.T, s *jsuite) {
	if n := len(k1Namen()); n != k1ZahlASCII+6 {
		t.Errorf("K1: %d Namen unter den Fällen, erwartet %d (3 × 128 für c, ~c, /c und 6 Zeichen ab U+0080)", n, k1ZahlASCII+6)
	}
	s.pruefeAlle(t, jfK1())
	jK1Oben(t, s)
	for _, p := range []struct {
		e    *model.DatenFehler
		soll string
	}{
		{&model.DatenFehler{Ursache: model.DatenSyntax}, "ODR-E6101: Baseline-Import ungültig oder inkonsistent: syntax"},
		{&model.DatenFehler{Ursache: model.DatenTyp, Ort: "/value/0/ID"}, `ODR-E6101: Baseline-Import ungültig oder inkonsistent: typ "/value/0/ID"`},
		{&model.DatenFehler{Ursache: model.DatenUnbekannteProperty, Ort: `/a"b`}, `ODR-E6101: Baseline-Import ungültig oder inkonsistent: unbekannte_property "/a\"b"`},
	} {
		if got := p.e.Error(); got != p.soll {
			t.Errorf("K1: Error() %q, erwartet %q", got, p.soll)
		}
		if p.e.Code() != model.CodeE6101 || p.e.Unwrap() != model.ErrDatenUngueltig || strings.Contains(model.CodeE6101.Message(), "/") {
			t.Errorf("K1: Code %v, Unwrap %v", p.e.Code(), p.e.Unwrap())
		}
	}
	for _, f := range jfK1()[:0x20] {
		r := s.lauf(f)
		for _, b := range []byte(r.err.Error()) {
			if b < 0x20 {
				t.Errorf("%s: Error() enthält das Byte %#x", f.name, b)
			}
		}
	}
}

func jK2(t *testing.T, s *jsuite) {
	r := s.lauf(jfall{m: "R", set: "Products", body: jKol("Products", `{"ID":"x"}`)})
	if r.err == nil {
		t.Fatal("K2: kein Fehler")
	}
	formen := map[string]error{"Fehler": r.err, "Errorf": fmt.Errorf("lesen: %w", r.err), "Join": errors.Join(errors.New("x"), r.err)}
	gegenstaende := []outbound.Gegenstand{
		outbound.GegenstandSonstiges, outbound.GegenstandSession, outbound.GegenstandWorkspace, outbound.GegenstandSnapshot,
		outbound.GegenstandSnapshotname, outbound.GegenstandImport, outbound.GegenstandExchange, outbound.GegenstandDienst,
	}
	for name, err := range formen {
		for _, g := range gegenstaende {
			if c, ok := outbound.Abbilden(err, g); !ok || c != model.CodeE6101 {
				t.Errorf("K2 %s / Gegenstand %d: %v %v, erwartet ODR-E6101", name, g, c, ok)
			}
		}
	}
	found := false
	for _, n := range outbound.EintraegeAbbildung() {
		found = found || n == "ErrDatenUngueltig"
	}
	if !found {
		t.Error("K2: die Tabelle der Abbildung nennt ErrDatenUngueltig nicht")
	}
	if _, ok := outbound.Abbilden(errors.Join(r.err, context.Canceled), outbound.GegenstandImport); ok {
		t.Error("K2: ein Kontextfehler neben dem Fehler bekommt einen Code")
	}
}

// jK3: kein Typ der Bibliothek steht in einer Signatur des Ports oder in einem Typ, den sie erreicht (SPEC-126, Kopf).
func jK3(t *testing.T, _ *jsuite) {
	erlaubt := map[string]bool{"": true, "context": true, "io": true, "github.com/pt9912/odata-recorder/src/hexagon/model": true}
	gesehen := map[reflect.Type]bool{}
	var geh func(reflect.Type)
	geh = func(ty reflect.Type) {
		if gesehen[ty] {
			return
		}
		gesehen[ty] = true
		if ty.Name() != "" && !erlaubt[ty.PkgPath()] {
			t.Errorf("K3: der Typ %s aus %s steht in einer Signatur", ty, ty.PkgPath())
		}
		if ty.PkgPath() == "context" || ty.PkgPath() == "io" {
			return
		}
		for _, k := range erreichbareTypen(ty) {
			geh(k)
		}
	}
	port := reflect.TypeFor[outbound.ODataCodec]()
	for i := range port.NumMethod() {
		geh(port.Method(i).Type)
	}
	if len(gesehen) < 12 {
		t.Errorf("K3: nur %d Typen erreicht, erwartet mindestens zwölf", len(gesehen))
	}
}

// erreichbareTypen sind die Typen, die ein Typ in einer Signatur oder in einem exportierten Feld nennt.
func erreichbareTypen(ty reflect.Type) []reflect.Type {
	var out []reflect.Type
	switch ty.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Chan:
		out = append(out, ty.Elem())
	case reflect.Map:
		out = append(out, ty.Key(), ty.Elem())
	case reflect.Struct:
		for i := range ty.NumField() {
			if ty.Field(i).IsExported() {
				out = append(out, ty.Field(i).Type)
			}
		}
	case reflect.Interface:
		for i := range ty.NumMethod() {
			out = append(out, ty.Method(i).Type)
		}
	case reflect.Func:
		for i := range ty.NumIn() {
			out = append(out, ty.In(i))
		}
		for i := range ty.NumOut() {
			out = append(out, ty.Out(i))
		}
	default:
	}
	return out
}

// jK4: der Leser hat kein Feld: gleichzeitige Aufrufe teilen nichts (SPEC-126 Nr. 15, die Typen der Suite sind die Schnittstelle).
func jK4(t *testing.T, s *jsuite) {
	ty := reflect.TypeOf(s.neu())
	if ty.Kind() == reflect.Pointer && ty.Elem().Kind() == reflect.Struct && ty.Elem().NumField() != 0 {
		t.Errorf("K4: der Leser %s hat %d Felder", ty, ty.Elem().NumField())
	}
}

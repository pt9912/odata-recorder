package urlparser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestKommentarformOperand (Zeile 11a): `--` ist am Operandenplatz `filter`, die übrigen fünf Formen sind dort `wert`
// (SPEC-164 Nr. 2, 4 b, 4 e, 5 c). Menge 24 = 6 Formen × 4 Stellungen (Textanfang, hinter Operator, hinter `(`, hinter `not`),
// jede mit dem Rest ` c` hinter dem Token.
func TestKommentarformOperand(t *testing.T) {
	n := 0
	for _, f := range []string{"--", "//", "/*", "#", ";", "\\"} {
		want := model.URLWert
		if f == "--" {
			want = model.URLFilter
		}
		for _, text := range []string{f + " c", "Id eq " + f + " c", "(" + f + " c)", "not " + f + " c"} {
			flFehler(t, text, want)
			n++
		}
	}
	if n != 24 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestMinusLaeufe (Zeile 11b): jeder Lauf mit `-` am Anfang, der kein Literal-Lauf ist, ist am Operandenplatz `filter`, am
// Operatorplatz `wert`. Menge 19 = 8 + 3 Literal-Gegenfälle + 8 am Operatorplatz.
func TestMinusLaeufe(t *testing.T) {
	minus := []string{"-", "--", "---", "-x", "-_", "--5", "-Id", "-(5)"}
	n := 0
	for _, m := range minus {
		flFehler(t, "Id eq "+m, model.URLFilter)
		flFehler(t, "Id "+m, model.URLWert)
		n += 2
	}
	flGut(t, "Id eq -5", bin("eq", eig("Id"), zahl("-5")))
	flFehler(t, "Id eq -.5", model.URLLiteral)
	flFehler(t, "Id eq -INF", model.URLLiteralTyp)
	if n+3 != 19 {
		t.Fatalf("%d Fälle", n+3)
	}
}

// flFuelle liefert anfang gefolgt von so vielen `a`, dass der Text n Bytes lang ist.
func flFuelle(anfang string, n int, ende string) string {
	return anfang + strings.Repeat("a", n-len(anfang)-len(ende)) + ende
}

// TestLaengeVorForm (Zeile 14a): über 512 Bytes ist es `zu_lang`, bei höchstens 512 die Form (SPEC-164 Nr. 4 a, 6). Menge 10 = 5
// Formen × beide Seiten der Grenze; wo die Form im Schlüssel vorkommt, ist die Ursache dort dieselbe.
func TestLaengeVorForm(t *testing.T) {
	formen := []struct {
		name       string
		bauen      func(n int) string
		schluessel bool // der Schlüssel liest die Form mit derselben Länge
	}{
		{"ungeschlossen", func(n int) string { return flFuelle("'", n, "") }, true},
		{"ungeschlossen mit Präfix", func(n int) string { return flFuelle("X'", n, "") }, true},
		{"geschlossen mit Rest", func(n int) string { return flFuelle("'", n-1, "'") + "x" }, false},
		{"Ziffernlauf", func(n int) string { return flFuelle("1x", n, "") }, true},
		{"unlesbares Präfix", func(n int) string { return flFuelle("a-b'", n, "'") }, true},
	}
	n := 0
	for _, f := range formen {
		for _, seite := range []struct {
			laenge int
			want   model.URLUrsache
		}{{512, model.URLLiteral}, {513, model.URLZuLang}} {
			lit := f.bauen(seite.laenge)
			if f.name == "geschlossen mit Rest" {
				// die Zeichenkette allein hat 512 bzw. 513 Bytes, der Rest ist ein Byte
				lit = flFuelle("'", seite.laenge, "'") + "x"
			}
			flFehler(t, "Id eq "+lit, seite.want)
			if f.schluessel {
				if _, u := flSchluessel(t, lit); u != seite.want {
					t.Errorf("%s %d: Schlüssel %q, Filter %q", f.name, seite.laenge, u, seite.want)
				}
			}
			n++
		}
	}
	if n != 10 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestRestZaehltNicht (Zeile 14b): der Rest hinter dem schließenden `'` zählt nicht zur Länge. Menge 3.
func TestRestZaehltNicht(t *testing.T) {
	flFehler(t, "Id eq 'a'"+strings.Repeat("x", 600), model.URLLiteral)
	flFehler(t, "Id eq "+flFuelle("'", 512, "'")+"x", model.URLLiteral)
	flFehler(t, "Id eq "+flFuelle("'", 513, "'")+"x", model.URLZuLang)
}

// flGUIDMit liefert die Guid der Form 8-4-4-4-12 mit h als erster Ziffer.
func flGUIDMit(h byte) string { return string([]byte{h}) + "0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" }

func flGUIDLit(g string) interface{ IstFilterAusdruck() } {
	return model.FilterLiteral{Literal: model.Literal{Art: model.LiteralGUID, Text: strings.ToLower(g)}}
}

// TestGuidAmAnfang (Zeile 15a): eine Guid der Form 8-4-4-4-12 ist an jedem Anfangszeichen ein Literal-Lauf, gleich dem Schlüssel
// (`/S(<Guid>)` und `/S(Id=<Guid>)`). Menge 22 = 10 Ziffern + 6 Buchstaben klein + 6 groß.
func TestGuidAmAnfang(t *testing.T) {
	n := 0
	for _, h := range []byte("0123456789abcdefABCDEF") {
		g := flGUIDMit(h)
		flGut(t, "Id eq "+g, bin("eq", eig("Id"), flGUIDLit(g)))
		benannt, u := flSchluessel(t, g)
		unbenannt, err := urlparser.New().ParseURL("/S("+g+")", "")
		if u != "" || err != nil || benannt.Text != strings.ToLower(g) || unbenannt.Ressource.Schluessel[0].Literal != benannt {
			t.Errorf("%q: Schlüssel %v %v %v", g, benannt, u, err)
		}
		n++
	}
	if n != 22 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestHexAnfangBleibtWort (Zeile 15b): Namen und Schlüsselwörter mit Hex-Buchstaben am Anfang bleiben Wortläufe. Menge 19 = 14
// Operanden + 5 Operatoren.
func TestHexAnfangBleibtWort(t *testing.T) {
	woerter := []string{"a", "abc", "add", "and", "eq", "ge", "le", "ne", "AND", "Eq", "abcdef12", "abcdef12_3", "d", "f1"}
	for _, w := range woerter {
		flGut(t, w+" eq 1", bin("eq", eig(w), zahl("1")))
	}
	flGut(t, "x and y", bin("and", eig("x"), eig("y")))
	flGut(t, "x eq y", bin("eq", eig("x"), eig("y")))
	for _, w := range []string{"add", "div", "divby"} {
		flFehler(t, "x "+w+" y", model.URLFilter)
	}
	if len(woerter)+5 != 19 {
		t.Fatal("Menge")
	}
}

// TestGuidRandformen (Zeile 15c): nur die genaue Form zählt. Je Anfang (Buchstabe `a`, Ziffer `0`): 59 Randformen mit `wert` bei
// Buchstaben und `literal` bei Ziffer am Anfang (35 und 37 Zeichen, vier Trenner × `_` `:` `+` ersetzt oder gestrichen oder um
// eins verschoben, 31 Nicht-Hex-Zeichen `g`, Umlaut und NBSP im Lauf), dazu die Schreibweisen und die Hex-Folge ohne Trenner.
// Menge 2 × 59 + 2 × 3 = 124.
func TestGuidRandformen(t *testing.T) {
	n := 0
	for _, anfang := range []byte{'a', '0'} {
		g := flGUIDMit(anfang)
		kurz := model.URLWert
		if anfang == '0' {
			kurz = model.URLLiteral
		}
		var nahe []string
		nahe = append(nahe, g[:35], g+"1")
		for _, p := range []int{8, 13, 18, 23} {
			for _, ersatz := range []string{"_", ":", "+"} {
				nahe = append(nahe, g[:p]+ersatz+g[p+1:])
			}
			nahe = append(nahe, g[:p]+g[p+1:], g[:p-1]+"-"+g[p-1:p]+g[p+1:], g[:p]+g[p+1:p+2]+"-"+g[p+2:])
		}
		for i := 1; i < len(g); i++ {
			if g[i] != '-' {
				nahe = append(nahe, g[:i]+"g"+g[i+1:])
			}
		}
		nahe = append(nahe, g[:5]+"\xc3\xa4"+g[6:], g[:5]+"\xc2\xa0"+g[6:])
		if len(nahe) != 59 {
			t.Fatalf("%d Randformen", len(nahe))
		}
		for _, s := range nahe {
			flFehler(t, "Id eq "+s, kurz)
			n++
		}
		flGut(t, "Id eq "+strings.ToUpper(g), bin("eq", eig("Id"), flGUIDLit(g)))
		flGut(t, "Id eq "+g[:9]+strings.ToUpper(g[9:18])+g[18:], bin("eq", eig("Id"), flGUIDLit(g)))
		ohne := strings.ReplaceAll(g, "-", "")
		if anfang == 'a' {
			flGut(t, "Id eq "+ohne, bin("eq", eig("Id"), eig(ohne)))
		} else {
			flFehler(t, "Id eq "+ohne, model.URLLiteral)
		}
		n += 3
	}
	if n != 124 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestGuidKontext (Zeile 15d): die Guid ist an acht Stellungen ein Literal-Lauf (Ende, vor `)`, vor Leerraum, links, gegen Guid) oder
// hat das Ergebnis des Wortlauts (vor `(` ein Literal und `(` am Operatorplatz `wert`, vor `'` `literal`, am Operator geklebt
// `wert`); je mit Buchstabe und Ziffer am Anfang. Menge 16.
func TestGuidKontext(t *testing.T) {
	n := 0
	for _, anfang := range []byte{'a', '0'} {
		g := flGUIDMit(anfang)
		l, id := flGUIDLit(g), eig("Id")
		flGut(t, "Id eq "+g, bin("eq", id, l))
		flGut(t, "(Id eq "+g+")", bin("eq", id, l))
		flGut(t, "Id eq "+g+" and b", bin("and", bin("eq", id, l), eig("b")))
		flGut(t, g+" eq Id", bin("eq", l, id))
		flGut(t, g+" eq "+g, bin("eq", l, l))
		flFehler(t, "Id eq "+g+"(1)", model.URLWert)
		flFehler(t, "Id eq "+g+"'x'", model.URLLiteral)
		flFehler(t, "Id eq"+g, model.URLWert)
		n += 8
	}
	if n != 16 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestKodierungPaare (Zeile 1, Paare am selben Text): ungültiges UTF-8 und U+0000 gehen dem Leerraum am Rand (`wert`) vor.
// Menge 6 = 3 Verletzungen × 2 Ränder.
func TestKodierungPaare(t *testing.T) {
	n := 0
	for _, bad := range []string{"\x00", "\xff", "\xc3\x28"} {
		flFehler(t, " Id eq 1"+bad, model.URLKodierung)
		flFehler(t, "Id"+bad+" ", model.URLKodierung)
		n += 2
	}
	if n != 6 {
		t.Fatal(n)
	}
}

// TestAufrufVorUngueltigemZeichen (Zeile 23, Paare am selben Token): ein Wortlauf unmittelbar vor `(` ist ein Aufruf (`filter`),
// auch wenn er ein Zeichen enthält, das in keinem Namen steht. Menge 64 = 27 druckbare Zeichen außer `'`, `(`, `)`, `,`
// (ohne Buchstabe, Ziffer, `_`) + 32 Steuerzeichen + 5 Formen (Name über 128 Zeichen, `a/1`, `a//b`, `a.`, `a..b`).
func TestAufrufVorUngueltigemZeichen(t *testing.T) {
	n := 0
	var zeichen []byte
	for c := byte(0x21); c < 0x7f; c++ {
		alnum := c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
		if !alnum && !strings.ContainsRune("'(),", rune(c)) {
			zeichen = append(zeichen, c)
		}
	}
	for c := byte(1); c < 0x20; c++ {
		zeichen = append(zeichen, c)
	}
	zeichen = append(zeichen, 0x7f)
	for _, c := range zeichen {
		flFehler(t, "a"+string([]byte{c})+"b(1)", model.URLFilter)
		n++
	}
	for _, s := range []string{strings.Repeat("a", 129), "a/1", "a//b", "a.", "a..b"} {
		flFehler(t, s+"(1)", model.URLFilter)
		n++
	}
	if n != 64 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestAufrufAnfaenge (Zeile 20, Lauf-Anfänge): jeder Wortlauf unmittelbar vor `(` ist ein Aufruf, gleich mit welchem Anfang
// (Buchstabe, `_`, `$`, `@`, Zeichen jenseits von ASCII) und welchem Rest. Menge 245 = 7 Anfänge × 7 Reste × 5 Formen.
func TestAufrufAnfaenge(t *testing.T) {
	anfaenge := []string{"a", "Z", "_", "$", "@", "\xc3\xa4", "\xe6\x97\xa5"}
	reste := []string{"", "x", "_1", "$it", "x.y", "x/y", strings.Repeat("b", 200)}
	n := 0
	for _, a := range anfaenge {
		for _, r := range reste {
			for _, form := range []string{"()", "(Id)", "(Id,1)", "(f(Id))", "(Id"} {
				flFehler(t, a+r+form, model.URLFilter)
				n++
			}
		}
	}
	if n != 245 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestTiefeNachGruppe (Zeile 22, Abbau): nach dem Abschluss einer inneren Gruppe oder eines inneren `not` baut sich die Tiefe ab,
// nicht auf null. Menge 512 = äußere Art (`(` offen, `not` mit abgeschlossenem Operand) × 32 Tiefen × spätere Art (`(`, `not`)
// × innere Art (`(a)`, `not a`) × 2 Seiten der Grenze.
func TestTiefeNachGruppe(t *testing.T) {
	n := 0
	for _, aussen := range []string{"(", "not "} {
		for d := range 32 {
			for _, k := range [][2]string{{"(", "(a)"}, {"(", "not a"}, {"not ", "(a)"}, {"not ", "not a"}} {
				spaeter, innen := k[0], k[1]
				// offen beim späteren Öffnen: bei `(` außen d, bei `not` außen keines (sein Operand ist gelesen)
				offen := d
				if aussen == "not " {
					offen = 0
				}
				for _, m := range []int{32 - offen, 33 - offen} {
					text := strings.Repeat(aussen, d) + innen + " and " + strings.Repeat(spaeter, m) + "a"
					if spaeter == "(" {
						text += strings.Repeat(")", m)
					}
					if aussen == "(" {
						text += strings.Repeat(")", d)
					}
					want := model.URLUrsache("")
					if offen+m > 32 {
						want = model.URLZuViele
					}
					flFehler(t, text, want)
					n++
				}
			}
		}
	}
	if n != 512 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestPfadGlieder (Zeile 19, Menge der Glieder): ein Pfad mit n Gliedern ist `wert_nicht_unterstuetzt`, ein fehlerhaftes Glied an
// jeder Stelle `wert`. Menge 64 + 6 = 70: n = 2 bis 64 (63), n = 4096 (die Länge des Textes lässt höchstens 4096 Glieder zu), 6
// Stellen des fehlerhaften Glieds bei sechs Gliedern. Prüfgrenze: n = 2 bis 64 und 4096, dazwischen nicht einzeln; Grund: der Leser
// zählt Glieder nicht, er liest sie in einer Schleife; Folge: eine Grenze bei n zwischen 65 und 4095 bliebe ungefangen.
func TestPfadGlieder(t *testing.T) {
	n := 0
	for g := 2; g <= 64; g++ {
		flFehler(t, strings.TrimSuffix(strings.Repeat("a/", g), "/")+" eq 1", model.URLWertNichtUnterstuetzt)
		n++
	}
	flFehler(t, strings.TrimSuffix(strings.Repeat("a/", 4096), "/"), model.URLWertNichtUnterstuetzt)
	n++
	for stelle := range 6 {
		glieder := []string{"a", "a", "a", "a", "a", "a"}
		glieder[stelle] = "b#"
		flFehler(t, strings.Join(glieder, "/"), model.URLWert)
		n++
	}
	if n != 70 {
		t.Fatalf("%d Fälle", n)
	}
}

// flWerteTyp prüft, dass der Baum nur aus Werten besteht: kein Zeiger, keine Slice, keine Map, kein Kanal, keine Funktion
// (SPEC-165 Nr. 2); Schnittstellen führen zu ihren Knoten.
func flWerteTyp(t *testing.T, v reflect.Value, pfad string) {
	t.Helper()
	switch v.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		t.Errorf("%s: %s im Baum", pfad, v.Kind())
	case reflect.Interface:
		if !v.IsNil() {
			flWerteTyp(t, v.Elem(), pfad)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			flWerteTyp(t, v.Field(i), pfad+"."+v.Type().Field(i).Name)
		}
	}
}

// TestFilterBaumNurWerte (Zeile 28, SPEC-165 Nr. 2): der Baum jedes Textes des Korpus (40 Gutfälle) besteht nur aus Werten, und
// das Ändern eines Knotens eines Ergebnisses verändert das zweite Ergebnis desselben Textes nicht. Menge 40.
func TestFilterBaumNurWerte(t *testing.T) {
	for _, f := range flKorpus() {
		a1, _ := flLies(f.text)
		a2, _ := flLies(f.text)
		flWerteTyp(t, reflect.ValueOf(&a1).Elem(), f.text)
		if v, ok := a1.(model.FilterVergleich); ok {
			geaendert := v
			geaendert.Links, geaendert.Op = eig("geaendert"), model.VergleichNe
			if gleich(geaendert, a2) {
				t.Errorf("%q: Änderung ohne Wirkung auf die Kopie", f.text)
			}
		}
		if !gleich(a1, f.want) || !gleich(a2, f.want) {
			t.Errorf("%q: Ergebnisse verändert", f.text)
		}
	}
}

// TestFilterReihenfolge (Zeile 28, Zustand außerhalb der Instanz): nach jedem der 24 Fehlerfälle (er lässt Zustand zurück, wenn
// ein Aufruf ihn behält) liefert jeder der 64 Texte das Ergebnis, das der Text allein hat: der Baum des Korpus bzw. `wert`.
// Das Orakel ist der Korpus, nie ein früherer Aufruf. Menge 24 × 64 = 1536.
func TestFilterReihenfolge(t *testing.T) {
	gut := map[string]interface{ IstFilterAusdruck() }{}
	for _, f := range flKorpus() {
		gut[f.text] = f.want
	}
	n := 0
	for _, vorher := range flFehlerKorpus() {
		_, _ = flLies(vorher)
		for _, text := range flWiederholungskorpus() {
			if want, ok := gut[text]; ok {
				flGut(t, text, want)
			} else {
				flFehler(t, text, model.URLWert)
			}
			n++
		}
	}
	if n != 1536 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestOperatorwortVorAnfuehrung (SPEC-164 Nr. 4, Schlusssatz): ein Operatorwort unmittelbar vor `'` ist ein Literal mit Präfix und
// am Platz eines Operators `wert`, nie ein Operator und nie `filter`. Menge 360 = 120 Schreibweisen der 16 Operatorwörter (eq ne gt
// ge lt le and or: 36; has in add sub mul div divby mod: 84) × 3 Stellungen (nach `Id`, nach `Id eq 1`, nach `(Id)` ohne Leerraum).
func TestOperatorwortVorAnfuehrung(t *testing.T) {
	woerter := strings.Fields("eq ne gt ge lt le and or has in add sub mul div divby mod")
	n := 0
	for _, w := range woerter {
		for _, s := range flSchreibweisen(w) {
			for _, text := range []string{"Id " + s + "'x'", "Id eq 1 " + s + "'x'", "(Id)" + s + "'x'"} {
				flFehler(t, text, model.URLWert)
				n++
			}
		}
	}
	if len(woerter) != 16 || n != 360 {
		t.Fatalf("%d Wörter, %d Fälle", len(woerter), n)
	}
}

// TestMinusLaufKlasse (Zeile 11b, Stichprobe der unbeschränkten Menge): jeder Lauf mit `-` am Anfang, der kein Literal-Lauf ist, ist
// am Operandenplatz `filter`. Menge 244 = 79 druckbare ASCII-Zeichen (ohne Ziffer, `.`, `'`, `(`, `)`, `,`) + 32 Steuerzeichen + 3
// Zeichen jenseits von ASCII (2, 3 und NBSP) hinter dem `-`, dazu `-` + `a`, `-` + 3-Byte-Zeichen mit 1 bis 64 Wiederholungen
// (128) und 600 und 4000 Bytes. Prüfgrenze: Zeichenklassen einzeln, Längen bis 64 und zwei lange; Grund: der Leser prüft nach dem
// `-` nur das zweite Byte; Folge: eine Regel, die erst bei einer Länge zwischen 65 und 599 oder 601 und 3999 greift, bliebe ungefangen.
func TestMinusLaufKlasse(t *testing.T) {
	n := 0
	pruefe := func(lauf string) {
		flFehler(t, "Id eq "+lauf, model.URLFilter)
		n++
	}
	for c := byte(0x21); c < 0x7f; c++ {
		if (c >= '0' && c <= '9') || strings.ContainsRune(".'(),", rune(c)) {
			continue
		}
		pruefe("-" + string([]byte{c}))
	}
	for c := byte(1); c < 0x20; c++ {
		pruefe("-" + string([]byte{c}))
	}
	pruefe("-\x7f")
	for _, z := range []string{"\xc3\xa4", "\xe6\x97\xa5", "\xc2\xa0"} {
		pruefe("-" + z)
	}
	for k := 1; k <= 64; k++ {
		pruefe("-" + strings.Repeat("a", k))
		pruefe("-" + strings.Repeat("\xe6\x97\xa5", k))
	}
	pruefe("-" + strings.Repeat("a", 599))
	pruefe("-" + strings.Repeat("a", 3999))
	if n != 244 {
		t.Fatalf("%d Fälle", n)
	}
}

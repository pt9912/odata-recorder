package urlparserv2_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestFlSchluesselwoerter (Zeile 6): die zwölf Schlüsselwörter in jeder ASCII-Schreibweise (108) ergeben den Baum der
// Kleinschreibung, die acht Operatorwörter von V4 (84 Schreibweisen) sind alle `filter`, Namen behalten ihre Schreibweise.
// Menge 196 = 108 + 84 + 4.
func TestFlSchluesselwoerter(t *testing.T) {
	n := 0
	a, b := flEig("a"), flEig("b")
	for _, w := range []string{"eq", "ne", "gt", "ge", "lt", "le", "and", "or"} {
		for _, s := range flSchreibweisen(w) {
			flGut(t, "a "+s+" b", flBin(w, a, b))
			n++
		}
	}
	rest := []struct {
		wort, text string
		want       flBaum
	}{
		{"not", "%s a", flNicht(a)}, {"true", "a eq %s", flBin("eq", a, flWahr("true"))},
		{"false", "a eq %s", flBin("eq", a, flWahr("false"))}, {"null", "a eq %s", flBin("eq", a, flNull())},
	}
	for _, r := range rest {
		for _, s := range flSchreibweisen(r.wort) {
			flGut(t, strings.Replace(r.text, "%s", s, 1), r.want)
			n++
		}
	}
	if n != 108 {
		t.Fatalf("%d Schreibweisen der Schlüsselwörter", n)
	}
	for _, w := range []string{"has", "in", "add", "sub", "mul", "div", "divby", "mod"} {
		for _, s := range flSchreibweisen(w) {
			flFehler(t, "Id "+s+" 1", flK(model.KompatFilter))
			n++
		}
	}
	if n != 192 {
		t.Fatalf("%d Schreibweisen", n)
	}
	for _, name := range []string{"Id", "id", "Nullx", "ANDx", "Truex"} {
		flGut(t, name+" eq 1", flBin("eq", flEig(name), flZahl("1")))
	}
}

// TestFlOperatorUnbekannt (Zeile 7): ein unbekanntes Token am Platz eines Operators ist `wert`, nie `filter`. Menge 12.
func TestFlOperatorUnbekannt(t *testing.T) {
	texte := []string{"Id foo 1", "Id equals 1", "Id = 1", "Id == 1", "Id != 1", "Id > 1", "Id <= 1", "Id && 1", "Id eqq 1", "Id e 1",
		"Id 1", "Id 'x'"}
	for _, s := range texte {
		flFehler(t, s, flU(model.URLWert))
	}
	if len(texte) != 12 {
		t.Fatal(len(texte))
	}
}

// TestFlOperatorwortAlsName (Zeile 8): am Platz eines Operanden sind Schlüsselwörter Namen, `not` ist Operator. Menge 19.
func TestFlOperatorwortAlsName(t *testing.T) {
	woerter := []string{"eq", "ne", "gt", "ge", "lt", "le", "and", "or", "has", "in", "add", "sub", "mul", "div", "divby", "mod"}
	for _, w := range woerter {
		flGut(t, w+" eq 1", flBin("eq", flEig(w), flZahl("1")))
	}
	for _, s := range []string{"not eq 1", "not", "not not"} {
		flFehler(t, s, flU(model.URLWert))
	}
	if len(woerter) != 16 {
		t.Fatal(len(woerter))
	}
}

// TestFlNamen (Zeile 9): Buchstabe und Ziffer nach Unicode `L` und `Nd` (Orakel `unicode.IsLetter`, `unicode.IsDigit`), 1 bis
// 128 Zeichen, ohne Normalisierung. Menge 66 = 29 Zeichen × 2 Stellen (58) + Länge 1, 128, 129 + zusammengesetzt und zerlegt +
// `Id`/`id` + Ziffer am Anfang.
func TestFlNamen(t *testing.T) {
	zeichen := []int{0x61, 0x5A, 0xE4, 0xDF, 0x65E5, 0x3A9, // L
		0x35, 0x663, 0x96B, 0xFF15, // Nd
		0x2163, 0x3007, 0x16EE, // Nl
		0xB2, 0xB3, 0xBD, 0x2460, // No
		0x301, 0x308, 0x903, // M
		0x5F, 0xB7, 0x2022, 0x20AC, 0xA9, 0x2603, 0x1F600, 0x203F, 0xD7} // sonstige
	n := 0
	for _, c := range zeichen {
		r := rune(c)
		erstes, spaeteres := unicode.IsLetter(r) || r == '_', unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
		for _, vorn := range []bool{true, false} {
			name, gueltig, ursache := "a"+string(r), spaeteres, model.URLWert
			if vorn {
				name, gueltig = string(r)+"a", erstes
				if r == '5' {
					ursache = model.URLLiteral // eine ASCII-Ziffer am Anfang ist ein Literal-Lauf
				}
			}
			if gueltig {
				flGut(t, name+" eq 1", flBin("eq", flEig(name), flZahl("1")))
			} else {
				flFehler(t, name+" eq 1", flU(ursache))
			}
			n++
		}
	}
	a128 := flWiederholt("a", 128)
	flGut(t, "a eq 1", flBin("eq", flEig("a"), flZahl("1")))
	flGut(t, a128+" eq 1", flBin("eq", flEig(a128), flZahl("1")))
	flFehler(t, flWiederholt("a", 129)+" eq 1", flU(model.URLWert))
	flGut(t, flRune(0xE9)+" eq 1", flBin("eq", flEig(flRune(0xE9)), flZahl("1")))
	flFehler(t, "e"+flRune(0x301)+" eq 1", flU(model.URLWert))
	flGut(t, "Id eq 1", flBin("eq", flEig("Id"), flZahl("1")))
	flGut(t, "id eq 1", flBin("eq", flEig("id"), flZahl("1")))
	flFehler(t, "5a eq 1", flU(model.URLLiteral))
	if n != 58 {
		t.Fatalf("%d Fälle", n)
	}
}

// flAusnahmen sind die Zeichen, die in Namen nicht zu `wert` führen: je Stelle (Anfang, Mitte, Ende) das erwartete Urteil. Die
// Tabelle ist das Orakel; sie steht im Test aus dem Wortlaut von SPEC-164 Nr. 4 mit SPEC-173 Nr. 5.
func flAusnahmen() map[byte][3]flV {
	w, f, l := flU(model.URLWert), flK(model.KompatFilter), flU(model.URLLiteral)
	return map[byte][3]flV{
		'$': {w, w, w}, '@': {f, w, w}, '+': {l, w, w}, '-': {f, w, w}, '.': {l, f, w}, '/': {w, f, w}, '\'': {l, l, l},
		'(': {w, f, f}, ')': {w, w, w}, ',': {w, w, w},
	}
}

// TestFlZeichenAussenGrammatik (Zeile 10): jedes druckbare ASCII-Zeichen außer Buchstabe, Ziffer, `_` und Leerzeichen (31) und
// jedes Steuerzeichen U+0001 bis U+001F und U+007F (32) am Anfang eines Laufs, in der Mitte und am Ende eines Namens. Menge 189.
func TestFlZeichenAussenGrammatik(t *testing.T) {
	ausnahmen := flAusnahmen()
	var zeichen []byte
	for c := byte(0x21); c < 0x7f; c++ {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && c != '_' {
			zeichen = append(zeichen, c)
		}
	}
	if len(zeichen) != 31 {
		t.Fatalf("%d druckbare Zeichen", len(zeichen))
	}
	for c := byte(1); c < 0x20; c++ {
		zeichen = append(zeichen, c)
	}
	zeichen = append(zeichen, 0x7f)
	n := 0
	for _, c := range zeichen {
		want := [3]flV{flU(model.URLWert), flU(model.URLWert), flU(model.URLWert)}
		if a, ok := ausnahmen[c]; ok {
			want = a
		}
		s := string([]byte{c})
		flFehler(t, s+"a eq 1", want[0])
		flFehler(t, "a"+s+"b eq 1", want[1])
		flFehler(t, "a"+s+" eq 1", want[2])
		n += 3
	}
	if n != 189 || len(ausnahmen) != 10 {
		t.Fatalf("%d Fälle, %d Ausnahmen", n, len(ausnahmen))
	}
}

// TestFlKommentarformen (Zeile 11): `//`, `/*`, `#`, `;` und der Rückstrich beginnen keinen Kommentar und lassen nichts aus: am
// Platz eines Operators sind sie `wert`, am Platz eines Operanden `wert`; `--` ist dort ein unäres Minus und `filter`. Menge 12.
func TestFlKommentarformen(t *testing.T) {
	n := 0
	for _, k := range []string{"-- c", "// c", "/* c */", "# c", "; c", "\\"} {
		flFehler(t, "Id eq 1 "+k, flU(model.URLWert))
		operand := flU(model.URLWert)
		if k == "-- c" {
			operand = flK(model.KompatFilter)
		}
		flFehler(t, "Id eq "+k, operand)
		n += 2
	}
	if n != 12 {
		t.Fatalf("%d Fälle", n)
	}
}

// flMinusOrakel ist die Regel von SPEC-173 Nr. 5 (*Literal-Lauf ohne Präfix*, *Weitere Konstrukte*) für einen Lauf `-` folge fort:
// nach `-` ein `.` ist ein Literal-Lauf ohne gültige Zahl (`literal`), nach `-` und Ziffer entscheidet die Form (Zahl, Exponent
// `literal_typ`, sonst `literal`), jede andere Folge ist ein unäres Minus (`filter`). Er liefert das Urteil und, bei einer Zahl,
// ihren Text.
func flMinusOrakel(folge, fort string) (flV, string) {
	switch {
	case folge == "1" && (fort == "" || fort == "1" || fort == "5L" || fort == "5M"):
		return flV{}, "-" + folge + strings.TrimRight(fort, "LM")
	case folge == "1" && fort == "1e3":
		return flK(model.KompatLiteralTyp), ""
	case folge == "1" || folge == ".":
		return flU(model.URLLiteral), ""
	}
	return flK(model.KompatFilter), ""
}

// TestFlMinusLaeufe (Zeile 12): `-` mit Ziffer oder `.` ist ein Literal-Lauf, genau `-INF` ist `literal_typ`, jeder andere Lauf mit
// `-` ist ein unäres Minus. Menge 35 = 5 Folgen nach `-` (Ziffer, `.`, Buchstabe, `-`, `(`) × 7 Fortsetzungen, dazu Goldens.
// Prüfgrenze: die Klasse der Läufe mit `-` ist unbeschränkt, die Auswahl ist ein akzeptiertes Negativ.
func TestFlMinusLaeufe(t *testing.T) {
	n := 0
	for _, folge := range []string{"1", ".", "a", "-", "("} {
		for _, fort := range []string{"", "1", "5L", "5M", "1e3", "INF", "INFO"} {
			v, zahl := flMinusOrakel(folge, fort)
			text := "Id eq -" + folge + fort
			if v == (flV{}) {
				flGut(t, text, flBin("eq", flEig("Id"), flZahl(zahl)))
			} else {
				flFehler(t, text, v)
			}
			n++
		}
	}
	if n != 35 {
		t.Fatalf("%d Fälle", n)
	}
	flFehler(t, "-INF", flK(model.KompatLiteralTyp))
	flFehler(t, "Id eq -INF", flK(model.KompatLiteralTyp))
	flFehler(t, "-inf", flK(model.KompatFilter))
	flFehler(t, "-INFO", flK(model.KompatFilter))
	flFehler(t, "-INF1", flK(model.KompatFilter))
}

// TestFlLeerzeichen (Zeile 13): die Zahl der Leerzeichen zwischen Token und um die Klammern ändert nichts, zwischen Läufen ist
// Leerraum Pflicht, am Rand ist er `wert`. Menge 25 = 3 + 12 + 5 + 5 (der Plan nennt 24; der fünfte Fehlerfall `Id eq1 1` ist ein Zusatzfall).
func TestFlLeerzeichen(t *testing.T) {
	want := flBin("or", flBin("and", flBin("eq", flEig("Id"), flZahl("1")), flBin("eq", flEig("Name"), flZahl("2"))), flEig("c"))
	n := 0
	for _, k := range []int{1, 2, 5} {
		sp := flWiederholt(" ", k)
		flGut(t, "Id"+sp+"eq"+sp+"1"+sp+"and"+sp+"Name"+sp+"eq"+sp+"2"+sp+"or"+sp+"c", want)
		n++
	}
	for stelle := range 4 {
		for _, k := range []int{0, 1, 3} {
			s := [4]string{" ", " ", " ", " "}
			s[stelle] = flWiederholt(" ", k)
			flGut(t, "Id eq 1 and"+s[0]+"("+s[1]+"Name eq 2"+s[2]+")"+s[3]+"or c", want)
			n++
		}
	}
	for _, f := range []struct {
		text string
		u    model.URLUrsache
	}{{"Id eq1", model.URLWert}, {"Ideq 1", model.URLWert}, {"Id eq 1and Name eq 2", model.URLLiteral}, {"Id eq 1 andName eq 2", model.URLWert},
		{"Id eq1 1", model.URLWert}} { // der Zusatzfall hat einen rechten Operanden, damit ein angenommenes `eq1` sichtbar wird
		flFehler(t, f.text, flU(f.u))
		n++
	}
	for _, s := range []string{" Id eq 1", "Id eq 1 ", " Id eq 1 ", " ", "   "} {
		flFehler(t, s, flU(model.URLWert))
		n++
	}
	if n != 25 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestFlAndererLeerraum (Zeile 14): jeder andere Leerraum und jedes unsichtbare Zeichen trennt nichts; die Bytes stehen als
// Werte, nicht als Zeichen im Quelltext (AGENTS.md §3.18). Menge 72 = 12 Zeichen × 5 Stellen (Fehler mit der Ursache der Form des
// Laufs) + 12 im Inhalt einer Zeichenkette.
func TestFlAndererLeerraum(t *testing.T) {
	zeichen := []string{"\t", "\n", "\r", "\v", "\f", "\xc2\x85", "\xc2\xa0", "\xe2\x80\x83", "\xe2\x80\xa8", "\xe3\x80\x80",
		"\xe2\x80\x8b", "\xef\xbb\xbf"}
	n := 0
	for _, c := range zeichen {
		flFehler(t, c+"Id eq 1", flU(model.URLWert))
		flFehler(t, "Id"+c+"eq 1", flU(model.URLWert))
		flFehler(t, "Id eq"+c+"1", flU(model.URLWert))
		flFehler(t, "("+c+"Id eq 1)", flU(model.URLWert))
		flFehler(t, "Id eq 1"+c, flU(model.URLLiteral))
		flGut(t, "Id eq 'a"+c+"b'", flBin("eq", flEig("Id"), flZk("'a"+c+"b'")))
		n += 6
	}
	if n != 72 {
		t.Fatalf("%d Fälle", n)
	}
}

// flLiteralBau ist eine Form eines Literals mit der Länge als Parameter; gut liefert den Baum bei 512 Bytes, nil bei einer Form,
// die dort einen Fehler ergibt.
type flLiteralBau struct {
	name string
	bau  func(laenge int) string
	gut  func(laenge int) flBaum
}

// TestFlLiteralLaenge (Zeile 15): die Länge geht der Form vor: 512 Bytes mit Anführungszeichen sind lesbar, 513 sind `zu_lang`.
// Menge 18 = 6 Formen × beide Seiten + V4-Form + Suffix `d` + Exponent, je mit Füllung auf beiden Seiten.
func TestFlLiteralLaenge(t *testing.T) {
	a := func(n int) string { return flWiederholt("a", n) }
	formen := []flLiteralBau{
		{"Zeichenkette", func(n int) string { return "'" + a(n-2) + "'" }, func(n int) flBaum { return flZk("'" + a(n-2) + "'") }},
		{"ungeschlossen", func(n int) string { return "'" + a(n-1) }, nil},
		{"X", func(n int) string { return "X'" + a(n-3) + "'" }, nil},
		{"guid", func(n int) string { return "guid'" + a(n-6) + "'" }, nil},
		{"datetime", func(n int) string { return "datetime'" + a(n-10) + "'" }, nil},
		{"Ziffern", func(n int) string { return flWiederholt("1", n) }, func(n int) flBaum { return flZahl(flWiederholt("1", n)) }},
		{"V4-Form", func(n int) string { return "2020-01-01" + flWiederholt("0", n-10) }, nil},
		{"Suffix d", func(n int) string { return flWiederholt("1", n-1) + "d" }, nil},
		{"Exponent", func(n int) string { return flWiederholt("1", n-3) + "e10" }, nil},
	}
	bei512 := map[string]flV{"ungeschlossen": flU(model.URLLiteral), "X": flK(model.KompatLiteralTyp), "guid": flU(model.URLLiteral),
		"datetime": flU(model.URLLiteral), "V4-Form": flU(model.URLWert), "Suffix d": flK(model.KompatLiteralTyp), "Exponent": flK(model.KompatLiteralTyp)}
	n := 0
	for _, f := range formen {
		for _, laenge := range []int{512, 513} {
			text := "Id eq " + f.bau(laenge)
			if len(f.bau(laenge)) != laenge {
				t.Fatalf("%s: Länge %d statt %d", f.name, len(f.bau(laenge)), laenge)
			}
			switch {
			case laenge == 513:
				flFehler(t, text, flU(model.URLZuLang))
			case f.gut != nil:
				flGut(t, text, flBin("eq", flEig("Id"), f.gut(laenge)))
			default:
				flFehler(t, text, bei512[f.name])
			}
			n++
		}
	}
	if n != 18 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestFlRestZaehltNicht (Zeile 16): der Rest hinter dem schließenden `'` zählt nicht zur Länge. Menge 3.
func TestFlRestZaehltNicht(t *testing.T) {
	flFehler(t, "Id eq 'a'"+flWiederholt("b", 600), flU(model.URLLiteral))
	flFehler(t, "Id eq '"+flWiederholt("a", 510)+"'x", flU(model.URLLiteral))
	flFehler(t, "Id eq '"+flWiederholt("a", 511)+"'x", flU(model.URLZuLang))
}

// TestFlLiteralEnde (Zeile 17): ein Literal endet an U+0020, `)`, `,` oder dem Textende. Menge 27 = 6 Arten × 3 Ender + 6 mit `,`
// + 3. Hinter `X'…'` meldet der Leser `literal_typ`, bevor er das `,` sieht (der erste Verstoß entscheidet).
func TestFlLiteralEnde(t *testing.T) {
	arten := []struct {
		text string
		want flBaum
	}{
		{"'a'", flZk("'a'")}, {"true", flWahr("true")}, {"42", flZahl("42")},
		{"guid'0a1b2c3d-0000-1111-2222-333344445555'", flLit(model.LiteralGUID, "0a1b2c3d-0000-1111-2222-333344445555")},
		{"datetime'2020-01-02T03:04:05'", flLit(model.LiteralZeitpunkt, "2020-01-02T03:04:05Z")}, {"X'a'", nil},
	}
	n := 0
	for _, k := range arten {
		ende := flU(model.URLWert) // das `,` danach
		if k.want == nil {
			ende = flK(model.KompatLiteralTyp)
		}
		for _, text := range []string{"Id eq " + k.text, "(Id eq " + k.text + ")", "Id eq " + k.text + " and b"} {
			if k.want == nil {
				flFehler(t, text, flK(model.KompatLiteralTyp))
			} else if strings.HasSuffix(text, " and b") {
				flGut(t, text, flBin("and", flBin("eq", flEig("Id"), k.want), flEig("b")))
			} else {
				flGut(t, text, flBin("eq", flEig("Id"), k.want))
			}
			n++
		}
		flFehler(t, "Id eq "+k.text+",", ende)
		n++
	}
	for _, s := range []string{"Id eq 'a'b", "Id eq 'a'1", "Id eq 'a''"} {
		flFehler(t, s, flU(model.URLLiteral))
		n++
	}
	if n != 27 {
		t.Fatalf("%d Fälle", n)
	}
}

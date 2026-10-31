package urlparser_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flRune schreibt ein Zeichen aus seinem Wert; Zeichen jenseits von ASCII stehen nie im Quelltext (AGENTS.md §3.18).
func flRune(r int) string { return string(rune(r)) }

// TestLeserArgument (Zeile 1): bei unmittelbarem Aufruf sind ungültiges UTF-8 und U+0000 `kodierung` (vor der Grammatik), der
// leere Text `wert`. Menge 12.
func TestLeserArgument(t *testing.T) {
	n := 1
	flFehler(t, "", model.URLWert)
	// Die Grammatik hat vor der Stelle schon einen Verstoß (`filter`, `wert`, `wert`); die Kodierung geht vor.
	vorne := []string{"", "foo(", "Id eq 1 and ("}
	for _, bad := range []string{"\x00", "\xff", "\xc3\x28"} {
		flFehler(t, bad+"Id eq 1", model.URLKodierung)
		flFehler(t, "foo("+bad+")", model.URLKodierung)
		flFehler(t, vorne[2]+bad, model.URLKodierung)
		n += 3
	}
	flGut(t, flRune(0xE4)+" eq 1", bin("eq", eig(flRune(0xE4)), zahl("1")))
	flGut(t, "Name eq '"+flRune(0xE4)+"'", bin("eq", eig("Name"), zk("'"+flRune(0xE4)+"'")))
	if n+2 != 12 {
		t.Fatalf("%d Fälle", n+2)
	}
}

// TestSchluesselwoerter (Zeile 7): die zwölf Schlüsselwörter in jeder ASCII-Schreibweise (108) ergeben den Baum der
// Kleinschreibung, die acht Wörter von Nr. 5 b (84 Schreibweisen) sind alle `filter`, Namen behalten ihre Schreibweise. Menge 196.
func TestSchluesselwoerter(t *testing.T) {
	n := 0
	for _, w := range []string{"eq", "ne", "gt", "ge", "lt", "le", "and", "or"} {
		for _, s := range flSchreibweisen(w) {
			flGut(t, "a "+s+" b", bin(w, eig("a"), eig("b")))
			n++
		}
	}
	for _, s := range flSchreibweisen("not") {
		flGut(t, s+" a", nicht(eig("a")))
		n++
	}
	for _, w := range []string{"true", "false"} {
		for _, s := range flSchreibweisen(w) {
			flGut(t, "a eq "+s, bin("eq", eig("a"), wahr(w)))
			n++
		}
	}
	for _, s := range flSchreibweisen("null") {
		flGut(t, "a eq "+s, bin("eq", eig("a"), flNull()))
		n++
	}
	if n != 108 {
		t.Fatalf("%d Schreibweisen der Schlüsselwörter", n)
	}
	for _, w := range []string{"has", "in", "add", "sub", "mul", "div", "divby", "mod"} {
		for _, s := range flSchreibweisen(w) {
			flFehler(t, "Id "+s+" 1", model.URLFilter)
			n++
		}
	}
	if n != 192 {
		t.Fatalf("%d Schreibweisen", n)
	}
	flGut(t, "Id eq 1", bin("eq", eig("Id"), zahl("1")))
	flGut(t, "id eq 1", bin("eq", eig("id"), zahl("1")))
	flGut(t, "Nullx eq 1", bin("eq", eig("Nullx"), zahl("1")))
	flGut(t, "ANDx eq 1", bin("eq", eig("ANDx"), zahl("1")))
}

// TestOperatorUnbekannt (Zeile 8): ein unbekanntes Token am Platz eines Operators ist `wert`, nie `filter`. Menge 12.
func TestOperatorUnbekannt(t *testing.T) {
	texte := []string{"Id foo 1", "Id equals 1", "Id = 1", "Id == 1", "Id != 1", "Id > 1", "Id <= 1", "Id && 1", "Id eqq 1", "Id e 1",
		"Id 1", "Id 'x'"}
	for _, s := range texte {
		flFehler(t, s, model.URLWert)
	}
	if len(texte) != 12 {
		t.Fatal(len(texte))
	}
}

// TestOperatorwortAlsName (Zeile 9): am Platz eines Operanden sind Schlüsselwörter Namen, `not` ist Operator. Menge 19.
func TestOperatorwortAlsName(t *testing.T) {
	woerter := []string{"eq", "ne", "gt", "ge", "lt", "le", "and", "or", "has", "in", "add", "sub", "mul", "div", "divby", "mod"}
	for _, w := range woerter {
		flGut(t, w+" eq 1", bin("eq", eig(w), zahl("1")))
	}
	for _, s := range []string{"not eq 1", "not", "not not"} {
		flFehler(t, s, model.URLWert)
	}
	if len(woerter) != 16 {
		t.Fatal(len(woerter))
	}
}

// TestNamen (Zeile 10): Buchstabe und Ziffer nach Unicode `L` und `Nd` (Orakel `unicode.IsLetter`, `unicode.IsDigit`), 1 bis
// 128 Zeichen, ohne Normalisierung. Menge 66 = 29 Zeichen × 2 Stellen + 8.
func TestNamen(t *testing.T) {
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
				flGut(t, name+" eq 1", bin("eq", eig(name), zahl("1")))
			} else {
				flFehler(t, name+" eq 1", ursache)
			}
			n++
		}
	}
	flGut(t, "a eq 1", bin("eq", eig("a"), zahl("1")))
	flGut(t, strings.Repeat("a", 128)+" eq 1", bin("eq", eig(strings.Repeat("a", 128)), zahl("1")))
	flFehler(t, strings.Repeat("a", 129)+" eq 1", model.URLWert)
	flGut(t, flRune(0xE9)+" eq 1", bin("eq", eig(flRune(0xE9)), zahl("1")))
	flFehler(t, "e"+flRune(0x301)+" eq 1", model.URLWert)
	flGut(t, "Id eq 1", bin("eq", eig("Id"), zahl("1")))
	flGut(t, "id eq 1", bin("eq", eig("id"), zahl("1")))
	flFehler(t, "5a eq 1", model.URLLiteral)
	if n != 58 {
		t.Fatalf("%d Fälle", n)
	}
}

// flAusnahmen sind die Zeichen, die in Namen nicht zu `wert` führen: je Stelle (Anfang, Mitte, Ende) die erwartete Ursache.
func flAusnahmen() map[byte][3]model.URLUrsache {
	return map[byte][3]model.URLUrsache{
		'$':  {model.URLWert, model.URLWert, model.URLWert},
		'@':  {model.URLFilter, model.URLWert, model.URLWert},
		'+':  {model.URLLiteral, model.URLWert, model.URLWert},
		'-':  {model.URLFilter, model.URLWert, model.URLWert},
		'.':  {model.URLLiteral, model.URLFilter, model.URLWert},
		'/':  {model.URLWert, model.URLWertNichtUnterstuetzt, model.URLWert},
		'\'': {model.URLLiteral, model.URLLiteral, model.URLLiteral},
		'(':  {model.URLWert, model.URLFilter, model.URLFilter},
		')':  {model.URLWert, model.URLWert, model.URLWert},
		',':  {model.URLWert, model.URLWert, model.URLWert},
	}
}

// TestZeichenAussenGrammatik (Zeile 11): jedes druckbare ASCII-Zeichen außer Buchstabe, Ziffer, `_` und Leerzeichen (31) und
// jedes Steuerzeichen (32) am Anfang eines Laufs, in der Mitte und am Ende eines Namens (189), dazu 6 Kommentarformen. Menge 195.
func TestZeichenAussenGrammatik(t *testing.T) {
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
		want := [3]model.URLUrsache{model.URLWert, model.URLWert, model.URLWert}
		if a, ok := ausnahmen[c]; ok {
			want = a
		}
		s := string([]byte{c})
		flFehler(t, s+"a eq 1", want[0])
		flFehler(t, "a"+s+"b eq 1", want[1])
		flFehler(t, "a"+s+" eq 1", want[2])
		n += 3
	}
	for _, k := range []string{"-- c", "// c", "/* c */", "# c", "; c", "\\"} {
		flFehler(t, "Id eq 1 "+k, model.URLWert)
		n++
	}
	if n != 195 || len(ausnahmen) != 10 {
		t.Fatalf("%d Fälle, %d Ausnahmen", n, len(ausnahmen))
	}
}

// TestLeerzeichen (Zeile 12): Zahl der Leerzeichen zwischen Token und um die Klammern ändert nichts, zwischen Läufen ist
// Leerraum Pflicht, am Rand ist er `wert`. Menge 24.
func TestLeerzeichen(t *testing.T) {
	want := bin("or", bin("and", bin("eq", eig("Id"), zahl("1")), bin("eq", eig("Name"), zahl("2"))), eig("c"))
	n := 0
	for _, k := range []int{1, 2, 5} {
		sp := strings.Repeat(" ", k)
		flGut(t, "Id"+sp+"eq"+sp+"1"+sp+"and"+sp+"Name"+sp+"eq"+sp+"2"+sp+"or"+sp+"c", want)
		n++
	}
	for stelle := range 4 {
		for _, k := range []int{0, 1, 3} {
			s := [4]string{" ", " ", " ", " "}
			s[stelle] = strings.Repeat(" ", k)
			flGut(t, "Id eq 1 and"+s[0]+"("+s[1]+"Name eq 2"+s[2]+")"+s[3]+"or c", want)
			n++
		}
	}
	flFehler(t, "Id eq1", model.URLWert)
	flFehler(t, "Ideq 1", model.URLWert)
	flFehler(t, "Id eq 1and Name eq 2", model.URLLiteral)
	flFehler(t, "Id eq 1 andName eq 2", model.URLWert)
	for _, s := range []string{" Id eq 1", "Id eq 1 ", " Id eq 1 ", " ", "   "} {
		flFehler(t, s, model.URLWert)
	}
	if n+4+5 != 24 {
		t.Fatalf("%d Fälle", n+9)
	}
}

// TestAndererLeerraum (Zeile 13): jeder andere Leerraum und jedes unsichtbare Zeichen trennt nichts; die Bytes stehen als
// Werte, nicht als Zeichen im Quelltext. Menge 72 = 12 × 5 Stellen + 12 im Inhalt einer Zeichenkette.
func TestAndererLeerraum(t *testing.T) {
	zeichen := []string{"\t", "\n", "\r", "\v", "\f", "\xc2\x85", "\xc2\xa0", "\xe2\x80\x83", "\xe2\x80\xa8", "\xe3\x80\x80",
		"\xe2\x80\x8b", "\xef\xbb\xbf"}
	n := 0
	for _, c := range zeichen {
		flFehler(t, c+"Id eq 1", model.URLWert)
		flFehler(t, "Id"+c+"eq 1", model.URLWert)
		flFehler(t, "Id eq"+c+"1", model.URLWert)
		flFehler(t, "("+c+"Id eq 1)", model.URLWert)
		flFehler(t, "Id eq 1"+c, model.URLLiteral)
		flGut(t, "Id eq 'a"+c+"b'", bin("eq", eig("Id"), zk("'a"+c+"b'")))
		n += 6
	}
	if n != 72 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestLiteralLaenge (Zeile 14): 512 Bytes gültig, 513 `zu_lang`. Menge 6.
func TestLiteralLaenge(t *testing.T) {
	a := func(n int) string { return strings.Repeat("a", n) }
	flGut(t, "Id eq '"+a(510)+"'", bin("eq", eig("Id"), zk("'"+a(510)+"'")))
	flFehler(t, "Id eq '"+a(511)+"'", model.URLZuLang)
	flFehler(t, "Id eq X'"+a(509)+"'", model.URLLiteralTyp)
	flFehler(t, "Id eq X'"+a(510)+"'", model.URLZuLang)
	flGut(t, "Id eq "+strings.Repeat("1", 512), bin("eq", eig("Id"), zahl(strings.Repeat("1", 512))))
	flFehler(t, "Id eq "+strings.Repeat("1", 513), model.URLZuLang)
}

// TestLiteralEnde (Zeile 16): ein Literal endet an U+0020, `)`, `,` oder dem Textende. Menge 27 = 6 Arten × 3 Ender + 6 mit
// `,` + 3.
func TestLiteralEnde(t *testing.T) {
	arten := []struct {
		text string
		node interface{ IstFilterAusdruck() }
	}{
		{"'a'", zk("'a'")}, {"true", wahr("true")}, {"42", zahl("42")},
		{"0a1b2c3d-0000-1111-2222-333344445555", model.FilterLiteral{Literal: model.Literal{Art: model.LiteralGUID, Text: "0a1b2c3d-0000-1111-2222-333344445555"}}},
		{"2020-01-02", model.FilterLiteral{Literal: model.Literal{Art: model.LiteralDatum, Text: "2020-01-02"}}},
		{"2020-01-02T03:04:05Z", model.FilterLiteral{Literal: model.Literal{Art: model.LiteralZeitpunkt, Text: "2020-01-02T03:04:05Z"}}},
	}
	for _, k := range arten {
		want := bin("eq", eig("Id"), k.node)
		flGut(t, "Id eq "+k.text, want)
		flGut(t, "(Id eq "+k.text+")", want)
		flGut(t, "Id eq "+k.text+" and b", bin("and", want, eig("b")))
		flFehler(t, "Id eq "+k.text+",", model.URLWert)
	}
	for _, s := range []string{"Id eq 'a'b", "Id eq 'a'1", "Id eq 'a''"} {
		flFehler(t, s, model.URLLiteral)
	}
}

// TestPfad (Zeile 19): ein Pfad ist `wert_nicht_unterstuetzt`, ein fehlerhafter `wert`. Menge 10.
func TestPfad(t *testing.T) {
	for _, s := range []string{"a/b", "a/b/c", "a/b/c/d", "a/b eq 1", "1 eq a/b"} {
		flFehler(t, s, model.URLWertNichtUnterstuetzt)
	}
	for _, s := range []string{"a/", "/a", "a//b", "a/1", "a/b.c"} {
		flFehler(t, s, model.URLWert)
	}
}

// TestAufruf (Zeile 20): jeder Name unmittelbar vor `(` ist ein Aufruf, `filter`, auch ein erfundener; mit Leerraum davor ist es
// keiner. Menge 5 × (40 + 5) + 3.
func TestAufruf(t *testing.T) {
	bibliothek := strings.Fields("contains startswith endswith length indexof substring tolower toupper trim concat year month day " +
		"hour minute second fractionalseconds date time totaloffsetminutes totalseconds now mindatetime maxdatetime round floor " +
		"ceiling geo.distance geo.length geo.intersects cast isof any all hassubset hassubsequence matchesPattern substringof " +
		"Items/any Ns.Typ.Funktion")
	namen := append(bibliothek, "foo", "Ns.foo", "a/any", "true", "x1")
	if len(bibliothek) != 40 {
		t.Fatalf("%d Namen der Bibliothek", len(bibliothek))
	}
	n := 0
	for _, name := range namen {
		for _, form := range []string{"()", "(Id)", "(Id,1)", "(f(Id))", "(Id"} {
			flFehler(t, name+form, model.URLFilter)
			n++
		}
	}
	flFehler(t, "Name (Id)", model.URLWert)
	flGut(t, "not(Id)", nicht(eig("Id")))
	flGut(t, "not (Id)", nicht(eig("Id")))
	if n != 225 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestNichtUnterstuetzt (Zeile 21): weitere gültige, nicht unterstützte Konstrukte. Menge 25.
func TestNichtUnterstuetzt(t *testing.T) {
	faelle := []struct {
		text string
		u    model.URLUrsache
	}{
		{"-Id", model.URLFilter}, {"- 5", model.URLFilter}, {"-(5)", model.URLFilter}, {"--5", model.URLFilter}, {"-INFO", model.URLFilter},
		{"$it", model.URLFilter}, {"$root", model.URLFilter}, {"$this", model.URLFilter}, {"$it/Id", model.URLFilter},
		{"$this/Id", model.URLFilter}, {"$x", model.URLWert},
		{"@p", model.URLFilter}, {"@", model.URLWert}, {"@1", model.URLWert},
		{"Ns.Typ", model.URLFilter}, {"a.b", model.URLFilter}, {"a.b.c", model.URLFilter}, {"a.", model.URLWert},
		{".a", model.URLLiteral}, {"a..b", model.URLWert},
		{"Id in (1,2)", model.URLFilter}, {"Id in [1]", model.URLFilter}, {"Id has 1", model.URLFilter},
		{"1 add 2 eq 3", model.URLFilter}, {"Id mod 2 eq 0", model.URLFilter},
	}
	for _, f := range faelle {
		flFehler(t, f.text, f.u)
	}
	if len(faelle) != 25 {
		t.Fatal(len(faelle))
	}
}

// TestStruktur (Zeile 27): fehlender oder überzähliger Operand, Operator oder Klammer, `,` sind `wert`. Menge 24.
func TestStruktur(t *testing.T) {
	k := flFehlerKorpus()
	for _, s := range k {
		flFehler(t, s, model.URLWert)
	}
	if len(k) != 24 {
		t.Fatal(len(k))
	}
}

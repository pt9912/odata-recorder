package urlparser_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// namenOrakel hält die Regeln des Wortlauts als Ausdrücke, unabhängig vom Parser: ein Name ist ein Buchstabe (Kategorie L) oder `_`,
// dann Buchstaben, Ziffern (Nd) und `_`, höchstens 128 Zeichen (SPEC-161 Nr. 4, SPEC-125 Nr. 8 b); das Typpräfix besteht aus
// Buchstaben, Ziffern und `_` in jeder Folge (SPEC-161 Nr. 6 g); eine herstellerspezifische Option besteht aus ASCII-Buchstaben,
// Ziffern, `_`, `-`, `.` mit einem Buchstaben oder `_` am Anfang und höchstens 128 Zeichen (SPEC-162 Nr. 3 e).
type namenOrakel struct{ name, praefix, hersteller *regexp.Regexp }

func neuerNamenOrakel() namenOrakel {
	return namenOrakel{
		regexp.MustCompile(`^[\p{L}_][\p{L}\p{Nd}_]*$`),
		regexp.MustCompile(`^[\p{L}\p{Nd}_]+$`),
		regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.\-]*$`),
	}
}

func (o namenOrakel) istName(m string) bool {
	return o.name.MatchString(m) && utf8.RuneCountInString(m) <= 128
}

func (o namenOrakel) istPraefix(m string) bool { return o.praefix.MatchString(m) }

// istHersteller gilt für den Namen einer Option; beginnt er mit `@`, ist es ein Alias mit der Regel des Namens.
func (o namenOrakel) istHersteller(m string) bool {
	if alias, ok := strings.CutPrefix(m, "@"); ok {
		return o.istName(alias)
	}
	return o.hersteller.MatchString(m) && len(m) <= 128
}

// achsenOrt ist ein Ort, an dem der Parser einen Namen liest, mit der Erwartung von Hand zu einem Muster.
type achsenOrt struct {
	name string
	bau  func(m string, o namenOrakel) fall
}

func nameOderNicht(ok bool, gut *model.ODataURL, schlecht model.URLUrsache, ort, pfad, query string) fall {
	if ok {
		return fall{pfad: pfad, query: query, want: gut}
	}
	return fall{pfad: pfad, query: query, u: schlecht, ort: ort}
}

func achsenOrte() []achsenOrt {
	pfadGlied := func(option string) func(string, namenOrakel) fall {
		return func(m string, o namenOrakel) fall {
			ursache := model.URLWert
			if o.istName(m) {
				ursache = model.URLWertNichtUnterstuetzt
			}
			return fall{pfad: "/S", query: option + "=p/" + m, u: ursache, ort: "query/1"}
		}
	}
	return []achsenOrt{
		{"Typpräfix", func(m string, o namenOrakel) fall {
			ursache := model.URLLiteral
			switch {
			case len(m)+3 > 512:
				ursache = model.URLZuLang
			case o.istPraefix(m):
				ursache = model.URLLiteralTyp
			}
			return fall{pfad: "/S(Id=" + m + "'x')", u: ursache, ort: "pfad/1"}
		}},
		{"Name der Menge", func(m string, o namenOrakel) fall {
			gut := &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: m}}
			return nameOderNicht(o.istName(m), gut, model.URLSegment, "pfad/1", "/"+m, "")
		}},
		{"Name eines Schlüsselteils", func(m string, o namenOrakel) fall {
			return nameOderNicht(o.istName(m), entity("S", teil(m, zahlLit("1"))), model.URLKey, "pfad/1", "/S("+m+"=1)", "")
		}},
		{"Alias", func(m string, o namenOrakel) fall {
			gut := mengeMitOptionen(model.Optionen{Sonstige: []model.SonstigeOption{{Name: "@" + m, Wert: "1"}}})
			return nameOderNicht(o.istName(m), gut, model.URLOptionName, "query/1", "/S", "@"+m+"=1")
		}},
		{"Element von $select", func(m string, o namenOrakel) fall {
			gut := mengeMitOptionen(model.Optionen{Select: &model.SelectOption{Namen: []string{m}}})
			return nameOderNicht(o.istName(m), gut, model.URLWert, "query/1", "/S", "$select="+m)
		}},
		{"Element von $orderby", func(m string, o namenOrakel) fall {
			gut := mengeMitOptionen(model.Optionen{OrderBy: []model.Sortierung{{Name: m}}})
			return nameOderNicht(o.istName(m), gut, model.URLWert, "query/1", "/S", "$orderby="+m)
		}},
		{"Pfad-Glied in $select", pfadGlied("$select")},
		{"Pfad-Glied in $orderby", pfadGlied("$orderby")},
		{"Glied eines Rest-Segments", func(m string, o namenOrakel) fall {
			ursache := model.URLSegment
			if o.istName(m) {
				ursache = model.URLPfadNichtUnterstuetzt
			}
			return fall{pfad: "/S/p." + m, u: ursache, ort: "pfad/2"}
		}},
		{"nacktes Wort im Schlüssel", nacktesWortFall},
	}
}

// nacktesWortFall: ein Name ohne Literal ist `key`; sonst ist der Teil ein Literal: eine Zahl aus Ziffern, mehr als 512 Bytes sind
// `zu_lang`, jede andere Form `literal` (SPEC-161 Nr. 5 c, 6).
func nacktesWortFall(m string, o namenOrakel) fall {
	pfad := "/S(" + m + ")"
	switch {
	case o.istName(m):
		return fall{pfad: pfad, u: model.URLKey, ort: "pfad/1"}
	case len(m) > 512:
		return fall{pfad: pfad, u: model.URLZuLang, ort: "pfad/1"}
	case strings.Trim(m, "0123456789") == "":
		return fall{pfad: pfad, want: entity("S", teil("", zahlLit(m)))}
	}
	return fall{pfad: pfad, u: model.URLLiteral, ort: "pfad/1"}
}

// achsenMuster bildet die Muster aus zwei Achsen: Zeichenanteil (ASCII-Buchstabe, Ziffer, `_`, `-` und die Nachbarn der Bereiche
// `@`, `[`, Gravis, `{`, `:` um A-Z, a-z, 0-9) × vier Stellungen (44), und Länge in Zeichen (1, 127, 128, 129 mit Zeichen von 1, 2, 3
// und 4 Bytes, 16) sowie `_` allein in 128 und 129 Zeichen (2): zusammen 62 Muster.
func achsenMuster() []string {
	var ms []string
	for _, c := range []string{"a", "Z", "0", "9", "_", "-", "@", "[", "`", "{", ":"} {
		for _, st := range []stellung{allein, erste, innere, letzte} {
			ms = append(ms, st.muster(c))
		}
	}
	for _, u := range []string{"a", "\xc3\xa4", "\xe3\x81\x82", "\xf0\x9d\x90\x80"} {
		for _, n := range []int{1, 127, 128, 129} {
			ms = append(ms, strings.Repeat(u, n))
		}
	}
	return append(ms, strings.Repeat("_", 128), strings.Repeat("_", 129))
}

// pruefeMitPanik prüft einen Fall; eine Panik des Parsers ist ein Befund dieses Elements.
func pruefeMitPanik(t *testing.T, beschreibung string, f fall) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: Panik %v", beschreibung, r)
		}
	}()
	pruefeFall(t, parserNeu(), f)
}

// TestNamenAchsen (Zeile 28, SPEC-161 Nr. 4, SPEC-125 Nr. 8 b): ein Name besteht aus Buchstaben (L), Ziffern (Nd) und `_`, beginnt
// nicht mit einer Ziffer, enthält kein `-` und hat 1 bis 128 Zeichen (Zeichen, nicht Bytes). Schleife über 62 Muster (Zeichenanteil ×
// Stellung, Länge in Zeichen mit Mehrbyte-Zeichen an der Grenze) × 10 Orte, an denen die Regel des Namens gilt, 620 Elemente; die
// Erwartung kommt aus Ausdrücken des Wortlauts, nie aus dem Parser.
func TestNamenAchsen(t *testing.T) {
	o := neuerNamenOrakel()
	n := 0
	for _, ort := range achsenOrte() {
		for _, m := range achsenMuster() {
			pruefeMitPanik(t, fmt.Sprintf("%s %.40q", ort.name, m), ort.bau(m, o))
			n++
		}
	}
	if len(achsenMuster()) != 62 || len(achsenOrte()) != 10 || n != 620 {
		t.Fatalf("Zählprobe: %d Muster, %d Orte, %d Elemente, erwartet 62, 10 und 620", len(achsenMuster()), len(achsenOrte()), n)
	}
}

// herstellerFall: der Name einer herstellerspezifischen Option; ein gültiger Name steht in `Sonstige`, sonst ist es `option_name`.
func herstellerFall(m string, o namenOrakel) fall {
	gut := mengeMitOptionen(model.Optionen{Sonstige: []model.SonstigeOption{{Name: m, Wert: "1"}}})
	return nameOderNicht(o.istHersteller(m), gut, model.URLOptionName, "query/1", "/S", m+"=1")
}

// TestHerstellerNamen (Zeile 46, SPEC-162 Nr. 3 d, e): der Name einer herstellerspezifischen Option besteht aus ASCII-Buchstaben,
// Ziffern, `_`, `-`, `.`, beginnt mit einem Buchstaben oder `_` und hat 1 bis 128 Zeichen. Schleife über 13 Zeichen (die Nachbarn der
// Bereiche `@`, `[`, Gravis, `{`, `:`, `/` und `-`, `.`, `_`, Ziffer, Buchstaben) × 4 Stellungen (52), die Längen 127, 128, 129 und
// der leere Name: 56 Elemente.
func TestHerstellerNamen(t *testing.T) {
	o := neuerNamenOrakel()
	var ms []string
	for _, c := range []string{"a", "Z", "0", "9", "_", "-", ".", "@", "[", "`", "{", ":", "/"} {
		for _, st := range []stellung{allein, erste, innere, letzte} {
			ms = append(ms, st.muster(c))
		}
	}
	ms = append(ms, strings.Repeat("a", 127), strings.Repeat("a", 128), strings.Repeat("a", 129), "")
	for _, m := range ms {
		pruefeMitPanik(t, fmt.Sprintf("Hersteller %.40q", m), herstellerFall(m, o))
	}
	if len(ms) != 56 {
		t.Fatalf("Zählprobe: %d Elemente, erwartet 56", len(ms))
	}
}

// ersterAb80 liefert das erste Zeichen ab U+0080 einer Tabelle, oder -1.
func ersterAb80(tab *unicode.RangeTable) rune {
	for _, r := range tab.R16 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			if c >= 0x80 {
				return c
			}
		}
	}
	for _, r := range tab.R32 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			if c >= 0x80 {
				return c
			}
		}
	}
	return -1
}

// kategorieMenge bildet je Unicode-Kategorie (29 kodierbare: Lu Ll Lt Lm Lo, Mn Mc Me, Nd Nl No, Pc Pd Ps Pe Pi Pf Po, Sm Sc Sk So,
// Zs Zl Zp, Cc Cf Co und Cn; Cs ist nicht kodierbar) ein Zeichen: das erste ab U+0080 der Tabelle von Go, für Cn das Zeichen U+0378,
// das in keiner Tabelle steht. Die Klasse folgt allein dem Namen der Kategorie (L: Buchstabe, Nd: Ziffer, sonst keines).
func kategorieMenge(t *testing.T) []unicodeZeichen {
	t.Helper()
	namen := []string{"Lu", "Ll", "Lt", "Lm", "Lo", "Mn", "Mc", "Me", "Nd", "Nl", "No", "Pc", "Pd", "Ps", "Pe", "Pi", "Pf", "Po", "Sm", "Sc", "Sk", "So",
		"Zs", "Zl", "Zp", "Cc", "Cf", "Co", "Cn"}
	var menge []unicodeZeichen
	for _, n := range namen {
		cp := rune(0x0378)
		if n != "Cn" {
			cp = ersterAb80(unicode.Categories[n])
		}
		for kat, tab := range unicode.Categories {
			if (kat == n) != unicode.Is(tab, cp) && len(kat) == 2 && kat != "LC" {
				t.Errorf("U+%04X liegt in %s: %v, erwartet %v", cp, kat, unicode.Is(tab, cp), kat == n)
			}
		}
		klasse := keines
		switch {
		case n[0] == 'L':
			klasse = buchstabe
		case n == "Nd":
			klasse = ziffer
		}
		menge = append(menge, unicodeZeichen{cp: cp, bytes: string(cp), klasse: klasse, name: fmt.Sprintf("U+%04X %s", cp, n)})
	}
	return menge
}

// TestUnicodeKategorien (Zeile 28, SPEC-161 Nr. 4): die Zeichenmenge deckt jede Unicode-Kategorie, nicht eine Auswahl. Der Generator
// nimmt je Kategorie das erste Zeichen ab U+0080 aus den Tabellen von Go, die Zählprüfung hält 29 Kategorien mit 29 verschiedenen
// Zeichen fest, jedes liegt in genau seiner Kategorie; Schleife über 29 Zeichen × 11 Orte × 4 Stellungen = 1276 Elemente.
func TestUnicodeKategorien(t *testing.T) {
	menge := kategorieMenge(t)
	gesehen := map[rune]bool{}
	for _, z := range menge {
		gesehen[z.cp] = true
	}
	if len(menge) != 29 || len(gesehen) != 29 {
		t.Fatalf("%d Kategorien, %d verschiedene Zeichen, erwartet 29 und 29", len(menge), len(gesehen))
	}
	els := unicodeElemente(menge)
	for _, e := range els {
		pruefeUnicodeElement(t, e)
	}
	if len(els) != 29*11*4 {
		t.Fatalf("Zählprobe: %d Elemente, erwartet 1276", len(els))
	}
}

// TestUtf8Rand (Zeile 25, SPEC-161 Nr. 2): gültiges UTF-8 am Rand jeder Länge ist kein Verstoß der Kodierung. Zehn Zeichen (U+0080,
// U+07FF, U+0800, U+D7FF, U+E000, U+FFFD, U+FFFE, U+FFFF, U+10000, U+10FFFF) × 2 Schreibweisen (roh, `%XX`) × 2 Stellen (Wert einer
// Option: gelesen; Name einer Option: `option_name`, nicht `kodierung`): 40 Elemente.
func TestUtf8Rand(t *testing.T) {
	zeichen := []struct {
		cp rune
		b  string
	}{
		{0x80, "\xc2\x80"}, {0x7FF, "\xdf\xbf"}, {0x800, "\xe0\xa0\x80"}, {0xD7FF, "\xed\x9f\xbf"}, {0xE000, "\xee\x80\x80"},
		{0xFFFD, "\xef\xbf\xbd"}, {0xFFFE, "\xef\xbf\xbe"}, {0xFFFF, "\xef\xbf\xbf"}, {0x10000, "\xf0\x90\x80\x80"}, {0x10FFFF, "\xf4\x8f\xbf\xbf"},
	}
	n := 0
	for _, z := range zeichen {
		if r, l := utf8.DecodeRuneInString(z.b); r != z.cp || l != len(z.b) {
			t.Fatalf("Bytes %q sind U+%04X, erwartet U+%04X", z.b, r, z.cp)
		}
		var escape strings.Builder
		for i := range len(z.b) {
			fmt.Fprintf(&escape, "%%%02X", z.b[i])
		}
		for _, form := range []string{z.b, escape.String()} {
			wert := mengeMitOptionen(model.Optionen{Sonstige: []model.SonstigeOption{{Name: "x", Wert: z.b}}})
			pruefeMitPanik(t, fmt.Sprintf("U+%04X Wert %q", z.cp, form), fall{pfad: "/S", query: "x=" + form, want: wert})
			pruefeMitPanik(t, fmt.Sprintf("U+%04X Name %q", z.cp, form), fall{pfad: "/S", query: form + "=1", u: model.URLOptionName, ort: "query/1"})
			n += 2
		}
	}
	if n != 40 {
		t.Fatalf("Zählprobe: %d Elemente, erwartet 40", n)
	}
}

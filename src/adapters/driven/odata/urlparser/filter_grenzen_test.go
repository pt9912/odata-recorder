package urlparser_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flNichtN baut `not not … a` mit n Präfixen.
func flNichtN(n int) interface{ IstFilterAusdruck() } {
	x := eig("a")
	for range n {
		x = nicht(x)
	}
	return x
}

// TestTiefe (Zeile 22): 32 offene `(` und `not` zusammen sind gültig, die 33. ist `zu_viele`; die Tiefe wird beim Schließen
// abgebaut. Menge 77 = 64 + 8 + 4 + 1.
func TestTiefe(t *testing.T) {
	n := 0
	for tiefe := 1; tiefe <= 32; tiefe++ {
		flGut(t, strings.Repeat("(", tiefe)+"a"+strings.Repeat(")", tiefe), eig("a"))
		flGut(t, strings.Repeat("not ", tiefe)+"a", flNichtN(tiefe))
		n += 2
	}
	for _, tiefe := range []int{33, 34, 100, 4000} {
		flFehler(t, strings.Repeat("(", tiefe)+"a"+strings.Repeat(")", tiefe), model.URLZuViele)
		flFehler(t, strings.Repeat("not ", tiefe)+"a", model.URLZuViele)
		n += 2
	}
	// gemischt: `(` und `not` abwechselnd; 16 Paare sind 32 offene, 17 Paare 34; die Klammern sind kein Knoten
	flGut(t, strings.Repeat("(not ", 16)+"a"+strings.Repeat(")", 16), flNichtN(16))
	flGut(t, strings.Repeat("(", 16)+strings.Repeat("not ", 16)+"a"+strings.Repeat(")", 16), flNichtN(16))
	flFehler(t, strings.Repeat("(not ", 17)+"a"+strings.Repeat(")", 17), model.URLZuViele)
	flFehler(t, strings.Repeat("(", 17)+strings.Repeat("not ", 16)+"a"+strings.Repeat(")", 17), model.URLZuViele)
	n += 4
	// die Tiefe wird beim Schließen und nach dem Operanden von `not` abgebaut: 40 Gruppen und 40 `not` nacheinander
	gruppen := strings.TrimSuffix(strings.Repeat("(a) and ", 40), " and ")
	if a, err := flLies(gruppen); err != nil || flZaehleKnoten(t, a) != 79 {
		t.Errorf("40 flache Gruppen: %v %v", a, err)
	}
	nots := strings.TrimSuffix(strings.Repeat("not a and ", 40), " and ")
	if a, err := flLies(nots); err != nil || flZaehleKnoten(t, a) != 119 {
		t.Errorf("40 flache not: %v %v", a, err)
	}
	n++
	if n != 77 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestRangImFilter (Zeile 23): der erste Verstoß in der Lesefolge entscheidet, gleich welcher Klasse. Menge 49 = 7 × 7 Paare
// (A vor B) der lesbaren Ursachen.
func TestRangImFilter(t *testing.T) {
	frag := []struct {
		u    model.URLUrsache
		text string
	}{
		{model.URLWert, ")"}, {model.URLLiteral, "1x"}, {model.URLZuLang, "'" + strings.Repeat("a", 512) + "'"},
		{model.URLZuViele, strings.Repeat("(", 33)}, {model.URLFilter, "foo(1)"}, {model.URLWertNichtUnterstuetzt, "a/b"},
		{model.URLLiteralTyp, "X'a'"},
	}
	n := 0
	for _, a := range frag {
		flFehler(t, "Id eq 1 and "+a.text, a.u) // jeder Baustein allein hat seine Ursache
		for _, b := range frag {
			flFehler(t, "Id eq 1 and "+a.text+" "+b.text, a.u)
			n++
		}
	}
	if n != 49 {
		t.Fatalf("%d Paare", n)
	}
}

// TestStopHinterNichtUnterstuetzt (Zeile 35): hinter dem ersten nicht unterstützten Konstrukt wird nicht gelesen; ein
// Syntaxfehler dahinter ändert die Ursache nicht. Menge 7.
func TestStopHinterNichtUnterstuetzt(t *testing.T) {
	faelle := []struct {
		text string
		u    model.URLUrsache
	}{
		{"foo(Id)", model.URLFilter}, {"Id has 1", model.URLFilter}, {"-Id", model.URLFilter}, {"@p", model.URLFilter},
		{"a.b", model.URLFilter}, {"a/b", model.URLWertNichtUnterstuetzt}, {"X'a'", model.URLLiteralTyp},
	}
	for _, f := range faelle {
		flFehler(t, f.text, f.u)
		flFehler(t, f.text+" and (", f.u)
	}
	if len(faelle) != 7 {
		t.Fatal(len(faelle))
	}
}

// flAlloc liefert die Zuteilung eines Aufrufs (runtime.MemStats.TotalAlloc, Differenz vor und nach).
func flAlloc(text string) uint64 {
	var vor, nach runtime.MemStats
	runtime.ReadMemStats(&vor)
	_, _ = flLies(text)
	runtime.ReadMemStats(&nach)
	return nach.TotalAlloc - vor.TotalAlloc
}

// TestFilterWaechter (Zeile 31): die Knoten des Baums sind höchstens die Länge des Textes, die Zuteilung eines Aufrufs
// höchstens 64 · n + 65536. Menge 40 = 8 Familien × n = 8192 und 4096 + 24 Eingaben aus Bruchstücken mit rohen Bytes; die
// Knoten werden an den 40 Gutfällen gezählt.
func TestFilterWaechter(t *testing.T) {
	familien := []struct{ einheit, ende string }{{"(", ""}, {"not ", "a"}, {"'", ""}, {"''", ""}, {"a and ", "a"}, {"a eq 1 or ", "a"},
		{"-", ""}, {"x'", ""}}
	var texte []string
	for _, f := range familien {
		for _, n := range []int{8192, 4096} {
			texte = append(texte, strings.Repeat(f.einheit, (n-len(f.ende))/len(f.einheit))+f.ende)
		}
	}
	stuecke := []string{"(", ")", "not ", "'", "''", " and ", " or ", "a", "1", "\xff", "%", "\x00", "-", ",", " eq ", "x'", "\xef\xbb\xbf"}
	for saat := uint64(1); saat <= 24; saat++ {
		r := rng(saat)
		var b strings.Builder
		for b.Len() < 8000 {
			b.WriteString(stuecke[r.IntN(len(stuecke))])
		}
		texte = append(texte, b.String())
	}
	if len(texte) != 40 {
		t.Fatalf("%d Eingaben", len(texte))
	}
	flAlloc("a") // Aufwärmen: erste Aufrufe teilen Tabellen zu
	for _, s := range texte {
		if got, grenze := flAlloc(s), uint64(64*len(s)+65536); got > grenze {
			t.Errorf("Länge %d, Anfang %.20q: Zuteilung %d über %d", len(s), s, got, grenze)
		}
		if a, err := flLies(s); err == nil && flZaehleKnoten(t, a) > len(s) {
			t.Errorf("Länge %d: mehr Knoten als Bytes", len(s))
		}
	}
	for _, f := range flKorpus() {
		a, _ := flLies(f.text)
		if k := flZaehleKnoten(t, a); k > len(f.text) {
			t.Errorf("%q: %d Knoten", f.text, k)
		}
	}
}

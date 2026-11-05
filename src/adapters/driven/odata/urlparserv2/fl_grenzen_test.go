package urlparserv2_test

import (
	"math/rand/v2"
	"runtime"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func flRng(saat uint64) *rand.Rand { return rand.New(rand.NewPCG(saat, 0x6f64)) }

// flNichtN baut `not not … a` mit n Präfixen.
func flNichtN(n int) interface{ IstFilterAusdruck() } {
	x := flEig("a")
	for range n {
		x = flNicht(x)
	}
	return x
}

// TestFlTiefe (Zeile 25): 32 offene `(` und `not` zusammen sind gültig, die 33. ist `zu_viele`; die Tiefe wird beim Schließen
// abgebaut. Menge 77 = 64 + 8 + 4 + 1.
func TestFlTiefe(t *testing.T) {
	zuViele := flU(model.URLZuViele)
	n := 0
	for tiefe := 1; tiefe <= 32; tiefe++ {
		flGut(t, flWiederholt("(", tiefe)+"a"+flWiederholt(")", tiefe), flEig("a"))
		flGut(t, flWiederholt("not ", tiefe)+"a", flNichtN(tiefe))
		n += 2
	}
	for _, tiefe := range []int{33, 34, 100, 4000} {
		flFehler(t, flWiederholt("(", tiefe)+"a"+flWiederholt(")", tiefe), zuViele)
		flFehler(t, flWiederholt("not ", tiefe)+"a", zuViele)
		n += 2
	}
	// gemischt: 16 Paare sind 32 offene, 17 Paare 34; die Klammern sind kein Knoten
	flGut(t, flWiederholt("(not ", 16)+"a"+flWiederholt(")", 16), flNichtN(16))
	flGut(t, flWiederholt("(", 16)+flWiederholt("not ", 16)+"a"+flWiederholt(")", 16), flNichtN(16))
	flFehler(t, flWiederholt("(not ", 17)+"a"+flWiederholt(")", 17), zuViele)
	flFehler(t, flWiederholt("(", 17)+flWiederholt("not ", 16)+"a"+flWiederholt(")", 17), zuViele)
	n += 4
	// die Tiefe wird beim Schließen und nach dem Operanden von `not` abgebaut: 40 Gruppen und 40 `not` nacheinander
	for _, f := range []struct {
		text   string
		knoten int
	}{{strings.TrimSuffix(flWiederholt("(a) and ", 40), " and "), 79}, {strings.TrimSuffix(flWiederholt("not a and ", 40), " and "), 119}} {
		if a, err := flLies(f.text); err != nil || flZaehleKnoten(t, a) != f.knoten {
			t.Errorf("40 flache Teile: %v %v", a, err)
		}
	}
	n++
	if n != 77 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestFlRangImAusdruck (Zeile 26): der erste Verstoß in der Lesefolge entscheidet, gleich welcher Klasse. Menge 49 = 7 × 7 Paare
// (A vor B) über die sieben Verstoßformen: 42 verschiedene und 7 gleiche.
func TestFlRangImAusdruck(t *testing.T) {
	frag := []struct {
		v    flV
		text string
	}{
		{flU(model.URLWert), ")"}, {flU(model.URLLiteral), "1x"}, {flU(model.URLZuLang), "'" + flWiederholt("a", 512) + "'"},
		{flU(model.URLZuViele), flWiederholt("(", 33)}, {flK(model.KompatFilter), "foo(1)"}, {flK(model.KompatFilter), "a/b"},
		{flK(model.KompatLiteralTyp), "X'a'"},
	}
	n := 0
	for _, a := range frag {
		flFehler(t, "Id eq 1 and "+a.text, a.v) // jeder Baustein allein hat sein Urteil
		for _, b := range frag {
			flFehler(t, "Id eq 1 and "+a.text+" "+b.text, a.v)
			n++
		}
	}
	if n != 49 {
		t.Fatalf("%d Paare", n)
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

// TestFlWaechter (Zeile 34): die Knoten des Baums sind höchstens die Länge des Textes, die Zuteilung eines Aufrufs höchstens
// 64 · n + 65536, auch für den Kompatibilitätsfehler. Menge 40 = 8 Familien × n = 8192 und 4096 + 24 Eingaben aus Bruchstücken
// mit rohen Bytes; die Knoten werden an den 40 Gutfällen gezählt.
func TestFlWaechter(t *testing.T) {
	familien := []struct{ einheit, ende string }{{"(", ""}, {"not ", "a"}, {"'", ""}, {"''", ""}, {"a and ", "a"}, {"a eq 1 or ", "a"},
		{"-", ""}, {"x'", ""}}
	var texte []string
	for _, f := range familien {
		for _, n := range []int{8192, 4096} {
			texte = append(texte, flWiederholt(f.einheit, (n-len(f.ende))/len(f.einheit))+f.ende)
		}
	}
	stuecke := []string{"(", ")", "not ", "'", "''", " and ", " or ", "a", "1", "\xff", "%", "\x00", "-", ",", " eq ", "x'", "\xef\xbb\xbf"}
	for saat := uint64(1); saat <= 24; saat++ {
		r := flRng(saat)
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
	// Kompatibilitätsfehler: Aufrufe, Pfade und Literale fremden Typs bis zur Obergrenze des Textes
	for _, s := range []string{flWiederholt("a/", 4096), "a and " + flWiederholt("x", 8000) + "(", "X'" + flWiederholt("a", 500) + "'"} {
		if got, grenze := flAlloc(s), uint64(64*len(s)+65536); got > grenze {
			t.Errorf("Länge %d: Zuteilung %d über %d", len(s), got, grenze)
		}
	}
}

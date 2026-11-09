package model_test

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Das Orakel der Grammatik (SPEC-283 Nr. 3) ist ein regulärer Ausdruck über dem Wert, in dem jedes Byte als eine Rune gleichen Werts
// steht (U+0000 bis U+00FF); es teilt keinen Code mit dem Leser des Modells.
const tagMuster = `(?:W/)?"[\x21\x23-\x7e\x{80}-\x{ff}]*"`

var (
	listeMuster = regexp.MustCompile(`^(?:[ \t]*,)*[ \t]*` + tagMuster + `[ \t]*(?:,[ \t]*(?:` + tagMuster + `[ \t]*)?)*$`)
	sternMuster = regexp.MustCompile(`^(?:[ \t]*,)*[ \t]*\*[ \t]*(?:,[ \t]*)*$`)
	tagFinder   = regexp.MustCompile(`"[\x21\x23-\x7e\x{80}-\x{ff}]*"`)
)

// alsRunen bildet jedes Byte auf eine Rune gleichen Werts ab.
func alsRunen(s string) string {
	r := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		r[i] = rune(s[i])
	}
	return string(r)
}

// befund ist, was das Orakel über die Zeilen eines Headers sagt.
type befund struct {
	gesetzt, lesbar, stern bool
	tags                   []string
}

// orakel liest die Zeilen: ohne Zeile nicht gesetzt, sonst die durch `,` verbundenen Zeilen als Liste.
func orakel(zeilen []string) befund {
	if len(zeilen) == 0 {
		return befund{}
	}
	s := alsRunen(strings.Join(zeilen, ","))
	b := befund{gesetzt: true}
	switch {
	case sternMuster.MatchString(s):
		b.lesbar, b.stern = true, true
	case listeMuster.MatchString(s):
		b.lesbar = true
		b.tags = tagFinder.FindAllString(s, -1)
	}
	return b
}

// sollErfuellt ist die Zusage von SPEC-283 Nr. 4 und 5 über dem Orakel: jedes Tag wird auf den Bytes von `"` bis `"` mit `"<v>"` verglichen.
func sollErfuellt(im, inm []string, v int64) bool {
	tag := `"` + strconv.FormatInt(v, 10) + `"`
	if b := orakel(im); b.gesetzt && (!b.lesbar || !b.stern && !slices.Contains(b.tags, tag)) {
		return false
	}
	if b := orakel(inm); b.gesetzt && (!b.lesbar || b.stern || slices.Contains(b.tags, tag)) {
		return false
	}
	return true
}

// TestVorbedingungGoldens (F1, F3, F10, F12): handgeschriebene Fälle des Vertrags neben den Generatoren.
func TestVorbedingungGoldens(t *testing.T) {
	faelle := []struct {
		name     string
		im       []string
		inm      []string
		v        int64
		erfuellt bool
	}{
		{"ohne Header", nil, nil, 7, true},
		{"leere Wertlisten", []string{}, []string{}, 7, true},
		{"If-Match schwach trifft", []string{`W/"7"`}, nil, 7, true},
		{"If-Match stark trifft", []string{`"7"`}, nil, 7, true},
		{"If-Match veraltet", []string{`W/"6"`}, nil, 7, false},
		{"If-Match Stern", []string{`*`}, nil, 7, true},
		{"If-Match Stern bei Version 0", []string{`*`}, nil, 0, true},
		{"If-Match leere Zeile", []string{``}, nil, 7, false},
		{"If-Match nur Leerraum", []string{` `}, nil, 7, false},
		{"If-Match nur Kommas", []string{` , ,`}, nil, 7, false},
		{"If-None-Match Stern", nil, []string{`*`}, 7, false},
		{"If-None-Match Treffer", nil, []string{`W/"7"`}, 7, false},
		{"If-None-Match ohne Treffer", nil, []string{`W/"6"`, `"8"`}, 7, true},
		{"If-None-Match leere Zeile", nil, []string{``}, 7, false},
		{"beide erfüllt", []string{`"7"`}, []string{`"8"`}, 7, true},
		{"If-Match erfüllt, If-None-Match nicht", []string{`"7"`}, []string{`"7"`}, 7, false},
		{"If-Match nicht, If-None-Match erfüllt", []string{`"8"`}, []string{`"9"`}, 7, false},
		{"beide nicht", []string{`"8"`}, []string{`"7"`}, 7, false},
		{"Stern neben Tag in einer anderen Zeile", []string{`*`, `"7"`}, nil, 7, false},
		{"leere Zeile neben Treffer", []string{``, `"7"`}, nil, 7, true},
		{"Komma im Tag und Treffer dahinter", []string{`"a,b","7"`}, nil, 7, true},
		{"Komma im Tag allein", []string{`"a,b"`}, nil, 7, false},
		{"07 trifft 7 nicht", []string{`"07"`}, nil, 7, false},
		{"+7 trifft 7 nicht", []string{`"+7"`}, nil, 7, false},
		{"70 trifft 7 nicht", []string{`"70"`}, nil, 7, false},
		{"1 trifft 10 nicht", []string{`"1"`}, nil, 10, false},
		{"10 trifft 10", []string{`"10"`}, nil, 10, true},
		{"w/ klein", []string{`w/"7"`}, nil, 7, false},
		{"Leerraum nach W/", []string{`W/ "7"`}, nil, 7, false},
		{"Doppelstern", []string{`**`}, nil, 7, false},
		{"Rest hinter dem Tag", []string{`"7"x`}, nil, 7, false},
		{"zwei Tags ohne Komma", []string{`"7" "8"`}, nil, 7, false},
	}
	for _, f := range faelle {
		if got := model.VorbedingungErfuellt(f.im, f.inm, f.v); got != f.erfuellt {
			t.Errorf("%s: %v, erwartet %v", f.name, got, f.erfuellt)
		}
		if got := sollErfuellt(f.im, f.inm, f.v); got != f.erfuellt {
			t.Errorf("%s: Orakel %v, Golden %v", f.name, got, f.erfuellt)
		}
	}
}

// TestVorbedingungVersionen (F1, F2, F6): die Klasse der Versionen aus dem Wortlaut (`etagKlasse`: 0 bis 300, Zehner- und
// Zweierpotenzen mit Nachbarn, Stichproben bis 2^63-1) × Form {`W/"<v>"`, `"<v>"`} × Nachbarn: v trifft, v-1 und v+1 treffen nicht;
// If-Match und If-None-Match sind spiegelbildlich.
func TestVorbedingungVersionen(t *testing.T) {
	n := 0
	for _, v := range etagKlasse() {
		for _, praefix := range []string{`W/`, ``} {
			tag := func(x int64) []string { return []string{praefix + `"` + strconv.FormatInt(x, 10) + `"`} }
			if !model.VorbedingungErfuellt(tag(v), nil, v) || model.VorbedingungErfuellt(nil, tag(v), v) {
				t.Errorf("Version %d Form %q: der eigene Tag trifft nicht", v, praefix)
			}
			if v > 0 && (model.VorbedingungErfuellt(tag(v-1), nil, v) || !model.VorbedingungErfuellt(nil, tag(v-1), v)) {
				t.Errorf("Version %d Form %q: der Tag der Version davor trifft", v, praefix)
			}
			if v < 1<<63-1 && (model.VorbedingungErfuellt(tag(v+1), nil, v) || !model.VorbedingungErfuellt(nil, tag(v+1), v)) {
				t.Errorf("Version %d Form %q: der Tag der Version danach trifft", v, praefix)
			}
			n++
		}
	}
	if n < 8000 {
		t.Errorf("%d Elemente, erwartet mindestens 8000", n)
	}
}

// TestVorbedingungSchreibweisen (F6): Generator über die Schreibweisen von v (führende Null, Vorzeichen, Ziffer vorn oder hinten, Leerzeichen,
// Hex, andere Zahl) × v aus {0, 1, 7, 10, 99, 100, 2^63-1} × Form; getroffen wird allein die Zeichenfolge der Dezimalzahl ohne führende Null.
func TestVorbedingungSchreibweisen(t *testing.T) {
	n, treffer := 0, 0
	for _, v := range []int64{0, 1, 7, 10, 99, 100, 1<<63 - 1} {
		d := strconv.FormatInt(v, 10)
		kandidaten := []string{d, "0" + d, "00" + d, "+" + d, "-" + d, d + " ", " " + d, d + "0", "1" + d, d + "1", d + ".0", d + "e0", "0x" + d,
			strconv.FormatInt(v, 16) + "h", "a" + d, d + "a", "", " ", "x"}
		if len(d) > 1 {
			kandidaten = append(kandidaten, d[:len(d)-1], d[1:])
		}
		for _, k := range kandidaten {
			for _, praefix := range []string{`W/`, ``} {
				n++
				wert := []string{praefix + `"` + k + `"`}
				want := k == d
				if want {
					treffer++
				}
				if got := model.VorbedingungErfuellt(wert, nil, v); got != want {
					t.Errorf("If-Match %q bei Version %d: %v, erwartet %v", wert[0], v, got, want)
				}
				// Ein Leerzeichen im Tag macht den Wert unlesbar, also auch für If-None-Match nicht erfüllt.
				wantINM := !want && !strings.Contains(k, " ")
				if got := model.VorbedingungErfuellt(nil, wert, v); got != wantINM || got != sollErfuellt(nil, wert, v) {
					t.Errorf("If-None-Match %q bei Version %d: %v, erwartet %v", wert[0], v, got, wantINM)
				}
			}
		}
	}
	if n < 280 || treffer != 14 {
		t.Errorf("%d Elemente, %d Treffer; erwartet mindestens 280 und 14", n, treffer)
	}
}

// grammatikBasen sind wohlgeformte Werte, an denen der Generator von TestVorbedingungGrammatik Byte für Byte ändert.
func grammatikBasen() []string {
	return []string{`"7"`, `W/"7"`, `*`, `"a,b","7"`, " \t\"7\" ,\tW/\"8\" ", `,"7",`, `"7","8"`, `W/"7" , "9"`}
}

// TestVorbedingungGrammatik (F7, F8): Generator über jede Position × jedes der 256 Bytes × {einfügen, ersetzen} und jede Position × {löschen}
// an acht wohlgeformten Werten; Orakel ist der reguläre Ausdruck der Grammatik, auf beide Header und die Version 7 angewandt. Die Zählprüfung
// nennt die Menge und beide Klassen (lesbar, unlesbar) müssen vorkommen.
func TestVorbedingungGrammatik(t *testing.T) {
	n, lesbar, unlesbar := 0, 0, 0
	pruefe := func(wert string) {
		n++
		im, inm := []string{wert}, []string{wert}
		if orakel(im).lesbar {
			lesbar++
		} else {
			unlesbar++
		}
		if got, want := model.VorbedingungErfuellt(im, nil, 7), sollErfuellt(im, nil, 7); got != want {
			t.Errorf("If-Match %q: %v, erwartet %v", wert, got, want)
		}
		if got, want := model.VorbedingungErfuellt(nil, inm, 7), sollErfuellt(nil, inm, 7); got != want {
			t.Errorf("If-None-Match %q: %v, erwartet %v", wert, got, want)
		}
	}
	soll := 0
	for _, basis := range grammatikBasen() {
		pruefe(basis)
		soll++
		for pos := 0; pos <= len(basis); pos++ {
			for b := 0; b < 256; b++ {
				pruefe(basis[:pos] + string([]byte{byte(b)}) + basis[pos:])
				soll++
				if pos < len(basis) {
					pruefe(basis[:pos] + string([]byte{byte(b)}) + basis[pos+1:])
					soll++
				}
			}
			if pos < len(basis) {
				pruefe(basis[:pos] + basis[pos+1:])
				soll++
			}
		}
	}
	if n != soll || n < 30000 || lesbar < 8000 || unlesbar < 20000 {
		t.Errorf("%d Elemente (erwartet %d, mindestens 30000), %d lesbar, %d unlesbar", n, soll, lesbar, unlesbar)
	}
}

// TestVorbedingungListen (F5, F7, F9): Generator über die Elemente {Treffer stark, Treffer schwach, `"a,b"`, leer, Nichttreffer} in jeder Reihenfolge
// der Länge 1 bis 3 × Leerraum {keiner, Leerzeichen, Tabulator, gemischt} an beiden Rändern jedes Elements × {eine Zeile, eine Zeile je
// Element}; die Erwartung folgt aus der Bauart (mindestens ein Element, Treffer irgendwo), nicht aus dem Orakel der Grammatik.
func TestVorbedingungListen(t *testing.T) {
	const (
		trefferStark = iota
		trefferSchwach
		komma
		leer
		nichtTreffer
	)
	text := []string{`"7"`, `W/"7"`, `"a,b"`, ``, `"8"`}
	raender := []string{"", " ", "\t", " \t \t"}
	n := 0
	var folgen [][]int
	for l := 1; l <= 3; l++ {
		var rec func(f []int)
		rec = func(f []int) {
			if len(f) == l {
				folgen = append(folgen, slices.Clone(f))
				return
			}
			for k := range text {
				rec(append(f, k))
			}
		}
		rec(nil)
	}
	for _, f := range folgen {
		hat := slices.ContainsFunc(f, func(k int) bool { return k == trefferStark || k == trefferSchwach })
		nichtLeer := slices.ContainsFunc(f, func(k int) bool { return k != leer })
		for _, vor := range raender {
			for _, nach := range raender {
				var elemente []string
				for _, k := range f {
					elemente = append(elemente, vor+text[k]+nach)
				}
				for _, zeilen := range [][]string{{strings.Join(elemente, ",")}, elemente} {
					n++
					wantIM, wantINM := nichtLeer && hat, nichtLeer && !hat
					if got := model.VorbedingungErfuellt(zeilen, nil, 7); got != wantIM {
						t.Errorf("If-Match %q: %v, erwartet %v", zeilen, got, wantIM)
					}
					if got := model.VorbedingungErfuellt(nil, zeilen, 7); got != wantINM {
						t.Errorf("If-None-Match %q: %v, erwartet %v", zeilen, got, wantINM)
					}
				}
			}
		}
	}
	if want := (5 + 25 + 125) * 16 * 2; n != want {
		t.Errorf("%d Elemente, erwartet %d", n, want)
	}
}

// TestVorbedingungNichtGesetzt (F10): kein Schlüssel ist hier nil, eine leere Wertliste ist die leere Liste; beide sind nicht gesetzt und
// erfüllt, eine leere Zeile ist gesetzt und nicht erfüllt, für jeden der beiden Header.
func TestVorbedingungNichtGesetzt(t *testing.T) {
	for i, nichtGesetzt := range [][]string{nil, {}, make([]string, 0, 3)} {
		if !model.VorbedingungErfuellt(nichtGesetzt, nil, 7) || !model.VorbedingungErfuellt(nil, nichtGesetzt, 7) ||
			!model.VorbedingungErfuellt(nichtGesetzt, nichtGesetzt, 7) {
			t.Errorf("Form %d: ein nicht gesetzter Header ist nicht erfüllt", i)
		}
	}
	if model.VorbedingungErfuellt([]string{""}, nil, 7) || model.VorbedingungErfuellt(nil, []string{""}, 7) {
		t.Error("eine leere Zeile ist erfüllt")
	}
}

// TestVorbedingungSternBeiJederVersion (F11, SPEC-283 Nr. 5): bei vorhandenem Eintrag ist If-None-Match: * nicht erfüllt und If-Match: * erfüllt, für
// jede Version der Klasse (`etagKlasse`: 0 bis 300, Potenzen mit Nachbarn, 2^63-1, Stichproben), auch Version 0 und die Grenze 0/1/2. Dazu `*` mit Leerraum
// und Kommas. Die Zählprüfung nennt die Menge.
func TestVorbedingungSternBeiJederVersion(t *testing.T) {
	n := 0
	klasse := etagKlasse()
	for _, v := range klasse {
		for _, s := range []string{"*", " * ", "\t*", ",*,", " , *\t,"} {
			n++
			if model.VorbedingungErfuellt(nil, []string{s}, v) || !model.VorbedingungErfuellt([]string{s}, nil, v) {
				t.Errorf("Version %d, Wert %q: If-None-Match erfüllt oder If-Match nicht erfüllt", v, s)
			}
		}
	}
	for v := int64(0); v <= 2; v++ {
		if model.VorbedingungErfuellt(nil, []string{"*"}, v) {
			t.Errorf("Version %d an der Grenze: If-None-Match: * erfüllt", v)
		}
	}
	if n != 5*len(klasse) || len(klasse) < 4000 {
		t.Errorf("%d Elemente bei %d Versionen, erwartet 5 je Version und mindestens 4000 Versionen", n, len(klasse))
	}
}

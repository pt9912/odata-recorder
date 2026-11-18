package urlparser_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestExpandOhneEinhaengung (Zeile 1, SPEC-262 Nr. 1): ohne `WithExpand` ist `$expand` mit jedem Wert `option_nicht_unterstuetzt`
// und der Wert bleibt ungelesen. Menge: 2 Parser (`New()`, `New(WithFilter)`) × 3 Wiederholungen (1 bis 3) × (5 Sonderwerte
// und 5219 Folgen der Länge 1 bis 3 über 17 Zeichen) = 31 344, dazu `$expand` ohne `=` je Parser.
func TestExpandOhneEinhaengung(t *testing.T) {
	parser := []*urlparser.Parser{urlparser.New(), urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser()))}
	werte := []string{"", "%G1", "%00", "(", "*"}
	for _, f := range exFolgen(3) {
		werte = append(werte, f.roh)
	}
	n := 0
	for _, p := range parser {
		for _, w := range werte {
			for wdh := 1; wdh <= 3; wdh++ {
				q := strings.TrimSuffix(strings.Repeat(exQ(w)+"&", wdh), "&")
				exPruefe(t, p, q, exFehl(model.URLOptionNichtUnterstuetzt, "query/1"))
				n++
			}
		}
		exPruefe(t, p, "$expand", exFehl(model.URLOptionNichtUnterstuetzt, "query/1"))
		n++
	}
	exZaehlt(t, "ohne Einhängung", n, 2*(3*5224+1))
}

// TestExpandListe (Zeile 2, SPEC-262 Nr. 5): die Namen stehen in der Reihenfolge der URL. Menge: Länge 1 bis 32 (je aufsteigend und
// absteigend, 64) und alle 24 Reihenfolgen der Länge 3 aus vier Namen.
func TestExpandListe(t *testing.T) {
	p := exParser()
	n := 0
	for laenge := 1; laenge <= 32; laenge++ {
		auf := exNamen("Nav", laenge)
		ab := make([]string, laenge)
		for i, s := range auf {
			ab[laenge-1-i] = s
		}
		for _, namen := range [][]string{auf, ab} {
			exPruefe(t, p, exQ(strings.Join(namen, ",")), exErg(namen...))
			n++
		}
	}
	vier := []string{"Category", "Supplier", "Ab", "Zz"}
	for i := range vier {
		for j := range vier {
			for k := range vier {
				if i == j || j == k || i == k {
					continue
				}
				namen := []string{vier[i], vier[j], vier[k]}
				exPruefe(t, p, exQ(strings.Join(namen, ",")), exErg(namen...))
				n++
			}
		}
	}
	exZaehlt(t, "Liste", n, 64+24)
}

// TestExpandDifferenz (Zeile 3, SPEC-262 Nr. 1, 5, 7): jede Query ohne `$expand` aus `korpus()` ergibt mit und ohne `WithExpand`
// dasselbe Ergebnis (tief gleich, auch der Fehler), und `Expand` ist `nil`; je Ressourcenart liest der Parser die Option.
func TestExpandDifferenz(t *testing.T) {
	ohne, mit := urlparser.New(), exParser()
	n := 0
	for _, f := range korpus() {
		if strings.Contains(strings.ToLower(f[1]), "expand") {
			continue
		}
		u1, e1 := ohne.ParseURL(f[0], f[1])
		u2, e2 := mit.ParseURL(f[0], f[1])
		if !gleich(u1, u2) || errText(e1) != errText(e2) {
			t.Errorf("%q %q:\n  ohne %+v %v\n  mit  %+v %v", f[0], f[1], u1, e1, u2, e2)
		}
		if u2.Optionen.Expand != nil {
			t.Errorf("%q %q: Expand ist nicht nil", f[0], f[1])
		}
		n++
	}
	if n < 50 {
		t.Fatalf("nur %d Queries verglichen", n)
	}
	for _, pfad := range []string{"/$metadata", "/S", "/S(1)", "/S/$count"} {
		u, err := mit.ParseURL(pfad, "$expand=Category,Supplier")
		if err != nil || u.Optionen.Expand == nil || len(u.Optionen.Expand.Elemente) != 2 {
			t.Errorf("%s: %+v %v", pfad, u.Optionen.Expand, err)
		}
	}
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// TestExpandDekodierung (Zeile 4, SPEC-262 Nr. 3, SPEC-162 Nr. 2): Name und Wert werden genau einmal dekodiert; die Schreibweise der
// Hex-Ziffern und ein Escape gegen das rohe Zeichen ändern nichts. Menge 24 Fälle (Fixture aus Bytewerten).
func TestExpandDekodierung(t *testing.T) {
	p := exParser()
	faelle := []struct {
		query string
		soll  exSoll
	}{
		{"$expand=Category", exErg("Category")},
		{"%24expand=Category", exErg("Category")},
		{"$expa%6Ed=Category", exErg("Category")},
		{"$expa%6ed=Category", exErg("Category")},
		{"%24%65xpand=Category", exErg("Category")},
		{"$expand=Category%2CSupplier", exErg("Category", "Supplier")},
		{"$expand=Category%2cSupplier", exErg("Category", "Supplier")},
		{"$expand=Categor%79", exErg("Category")},
		{"$expand=a" + exHex(exE, true), exErg("a" + exE)},
		{"$expand=a" + exHex(exE, false), exErg("a" + exE)},
		{"$expand=a" + exE, exErg("a" + exE)},
		{"$expand=Category%252CSupplier", exWert()},
		{"$expand=Category%2520", exWert()},
		{"%2524expand=Category", exFehl(model.URLOptionName, "query/1")},
		{"$Expand=Category", exFehl(model.URLOptionName, "query/1")},
		{"%24Expand=Category", exFehl(model.URLOptionName, "query/1")},
		{"$expand=Category%G1", exFehl(model.URLKodierung, "query/1")},
		{"$expand=Category%", exFehl(model.URLKodierung, "query/1")},
		{"$expand=Category%0", exFehl(model.URLKodierung, "query/1")},
		{"$expand=Cat%00", exFehl(model.URLKodierung, "query/1")},
		{"$expand=Cat%C3", exFehl(model.URLKodierung, "query/1")},
		{"$expand=Cat%FF", exFehl(model.URLKodierung, "query/1")},
		{"%G1=Category", exFehl(model.URLKodierung, "query/1")},
		{"$expand=Category+Supplier", exWert()},
	}
	for _, f := range faelle {
		exPruefe(t, p, f.query, f.soll)
	}
	exZaehlt(t, "Dekodierung", len(faelle), 24)
}

// TestExpandDoppelt (Zeile 5, SPEC-262 Nr. 2): ein zweites `$expand` ist `option_doppelt`, in jeder Schreibweise des Namens, mit
// gleichem, anderem und ungültigem Wert; der Wert des ersten wird gelesen (ist er ungültig, gilt dieser Verstoß), der des zweiten
// nicht. Menge: 3 Schreibweisen × 3 Schreibweisen × 4 erste Werte × 5 zweite Werte × 2 Stellen (ohne und mit `$top=1` davor).
func TestExpandDoppelt(t *testing.T) {
	p := exParser()
	namen := []string{"$expand", "%24expand", "$expa%6Ed"}
	erste := []string{"Category", "*", "a(x)", "1a"}
	zweite := []string{"Category", "Supplier", "1a", "%G1", "*"}
	n := 0
	for _, n1 := range namen {
		for _, n2 := range namen {
			for _, w1 := range erste {
				for _, w2 := range zweite {
					for _, davor := range []int{0, 1} {
						vor := strings.Repeat("$top=1&", davor)
						soll := exFehl(model.URLOptionDoppelt, "query/"+strconv.Itoa(davor+2))
						if w1 == "1a" {
							soll = exFehl(model.URLWert, "query/"+strconv.Itoa(davor+1))
						}
						exPruefe(t, p, vor+n1+"="+w1+"&"+n2+"="+w2, soll)
						n++
					}
				}
			}
		}
	}
	exZaehlt(t, "doppelt", n, 3*3*4*5*2)
}

// TestExpandLeer (Zeile 6, SPEC-262 Nr. 3 a): der leere Wert, mit und ohne `=`, ist `wert`.
func TestExpandLeer(t *testing.T) {
	p := exParser()
	for _, q := range []string{"$expand=", "$expand", "%24expand=", "$expand=&$top=1"} {
		exPruefe(t, p, q, exWert())
	}
	exPruefe(t, p, "$top=1&$expand=", exFehl(model.URLWert, "query/2"))
}

// TestExpandLeeresElement (Zeile 7, SPEC-262 Nr. 4 d): ein leeres Element an jeder Stelle ist `wert`. Menge: 1 bis 4 Namen × alle
// Stellen von vorn bis hinten (2 + 3 + 4 + 5 = 14), dazu `,`, `,,` und `a,,`.
func TestExpandLeeresElement(t *testing.T) {
	p := exParser()
	n := 0
	for laenge := 1; laenge <= 4; laenge++ {
		for stelle := 0; stelle <= laenge; stelle++ {
			namen := exNamen("n", laenge)
			elemente := append(append(append([]string{}, namen[:stelle]...), ""), namen[stelle:]...)
			exPruefe(t, p, exQ(strings.Join(elemente, ",")), exWert())
			n++
		}
	}
	for _, w := range []string{",", ",,", "a,,"} {
		exPruefe(t, p, exQ(w), exWert())
		n++
	}
	exZaehlt(t, "leeres Element", n, 14+3)
}

// exStellen sind die Stellen eines Zeichens im Namen `ab`: vorn, mitten, hinten.
func exStellen(c string) [3]string {
	return [3]string{c + "ab", "a" + c + "b", "ab" + c}
}

// exZeichenKlasse ist ein Zeichen jenseits von ASCII mit der Klasse des Orakels: Buchstabe, Ziffer oder keines von beiden.
type exZeichenKlasse struct {
	z      string
	klasse string
}

func exNamenzeichen() []exZeichenKlasse {
	return []exZeichenKlasse{{exE, "L"}, {exShe, "L"}, {exHan, "L"}, {exArab3, "N"}, {exNBSP, "-"}, {exBOM, "-"}, {exHoch2, "-"}}
}

// TestExpandNamenform (Zeile 8, SPEC-262 Nr. 4 a, d): Leerraum und jedes Zeichen außerhalb der Namensform an jeder Stelle des
// Namens ist `wert`; ein Buchstabe ist es nie, eine Ziffer nur am Anfang. Menge: die 125 ASCII-Zeichen außer `,`, `/` und `(`
// (eigene Zeilen), je 3 Stellen, roh (ohne Steuerzeichen, `&` und `%`) und als `%XX` in beiden Schreibweisen (ohne `%00`); dazu 7
// Zeichen jenseits von ASCII × 3 Stellen × 3 Schreibweisen (roh, `%XX` groß, klein).
func TestExpandNamenform(t *testing.T) {
	p := exParser()
	n := 0
	pruefe := func(roh, dek string) {
		t.Helper()
		exPruefe(t, p, exQ(roh), exOrakel(dek))
		n++
	}
	for b := range 128 {
		c := string([]byte{byte(b)})
		if c == "," || c == "/" || c == "(" {
			continue
		}
		rohOK := b >= 0x20 && b != 0x7f && c != "&" && c != "%"
		for i, dek := range exStellen(c) {
			if rohOK {
				pruefe(dek, dek)
			}
			if b != 0 {
				pruefe(exStellen(exHex(c, true))[i], dek)
				pruefe(exStellen(exHex(c, false))[i], dek)
			}
		}
	}
	for _, z := range exNamenzeichen() {
		for i, dek := range exStellen(z.z) {
			soll := exWert()
			if z.klasse == "L" || z.klasse == "N" && i > 0 {
				soll = exErg(dek)
			}
			for _, roh := range []string{exStellen(z.z)[i], exStellen(exHex(z.z, true))[i], exStellen(exHex(z.z, false))[i]} {
				exPruefe(t, p, exQ(roh), soll)
				n++
			}
		}
	}
	exZaehlt(t, "Namensform", n, 270+744+63)
}

// exOptSeg ist ein Segment der Query mit der Art der Option: `filter`, `expand` oder `andere`.
type exOptSeg struct{ text, art string }

// TestExpandOptionenFolgen (Zeile 3, SPEC-262 Nr. 1, SPEC-160 Nr. 8): die Einhängestellen von `New` ändern einander nicht, in jeder
// Folge. Die Folge bestimmt allein, ob `$filter` lesbar ist (die letzte `WithFilter` trägt einen Leser) und ob `$expand` lesbar
// ist (irgendeine `WithExpand`); ein nicht lesbares `$filter` ist `filter`, ein nicht lesbares `$expand`
// `option_nicht_unterstuetzt`, der erste Verstoß in der Lesefolge gilt. Menge: alle Folgen der Länge 0 bis 3 über {`WithFilter(Leser)`,
// `WithFilter(nil)`, `WithExpand()`} (1 + 3 + 9 + 27 = 40) × 7 Queries (`$filter` und `$expand` allein, in beiden Reihenfolgen, mit
// fremden Optionen dazwischen, und ohne beide) = 280 Fälle. Prüfgrenze: Länge 3 (es gibt drei Einhängungen; ein Zustand, der erst
// die vierte Option verlangt, bleibt ungefangen).
func TestExpandOptionenFolgen(t *testing.T) {
	const (
		opFilter = iota
		opFilterNil
		opExpand
	)
	leser := &zaehlLeser{ausdruck: testAusdruck{"ok"}}
	folgen := [][]int{{}}
	vorige := [][]int{{}}
	for range 3 {
		var naechste [][]int
		for _, v := range vorige {
			for _, o := range []int{opFilter, opFilterNil, opExpand} {
				naechste = append(naechste, append(append([]int{}, v...), o))
			}
		}
		folgen = append(folgen, naechste...)
		vorige = naechste
	}
	exZaehlt(t, "Folgen", len(folgen), 40)
	queries := [][]exOptSeg{
		{{"$filter=a", "filter"}},
		{{"$expand=Cat,Sup", "expand"}},
		{{"$filter=a", "filter"}, {"$expand=Cat", "expand"}},
		{{"$expand=Cat", "expand"}, {"$filter=a", "filter"}},
		{{"x=1", "andere"}, {"$expand=Cat", "expand"}, {"$top=1", "andere"}, {"$filter=a", "filter"}},
		{{"$filter=a", "filter"}, {"x=1", "andere"}, {"$expand=Cat", "expand"}},
		{{"$top=1", "andere"}, {"$skip=2", "andere"}, {"$count=true", "andere"}, {"$select=a,b", "andere"}, {"$orderby=a", "andere"}, {"x=1", "andere"}},
	}
	n := 0
	for _, folge := range folgen {
		var opts []urlparser.Option
		filterAn, expandAn := false, false
		for _, o := range folge {
			switch o {
			case opFilter:
				opts, filterAn = append(opts, urlparser.WithFilter(leser)), true
			case opFilterNil:
				opts, filterAn = append(opts, urlparser.WithFilter(nil)), false
			case opExpand:
				opts, expandAn = append(opts, urlparser.WithExpand()), true
			}
		}
		p := urlparser.New(opts...)
		for _, q := range queries {
			var texte, andere []string
			var verstoss *model.URLFehler
			for i, s := range q {
				texte = append(texte, s.text)
				switch {
				case s.art == "andere":
					andere = append(andere, s.text)
				case s.art == "filter" && !filterAn && verstoss == nil:
					verstoss = &model.URLFehler{Ursache: model.URLFilter, Ort: "query/" + strconv.Itoa(i+1)}
				case s.art == "expand" && !expandAn && verstoss == nil:
					verstoss = &model.URLFehler{Ursache: model.URLOptionNichtUnterstuetzt, Ort: "query/" + strconv.Itoa(i+1)}
				}
			}
			f := fall{pfad: "/S", query: strings.Join(texte, "&")}
			if verstoss != nil {
				f.u, f.ort = verstoss.Ursache, verstoss.Ort
			} else {
				ref, err := urlparser.New().ParseURL("/S", strings.Join(andere, "&"))
				if err != nil {
					t.Fatalf("Bezug %q: %v", andere, err)
				}
				for _, s := range q {
					switch s.art {
					case "filter":
						ref.Optionen.Filter = &model.FilterOption{Text: "a", Ausdruck: testAusdruck{"ok"}}
					case "expand":
						ref.Optionen.Expand = exKnoten(strings.Split(strings.TrimPrefix(s.text, "$expand="), ",")...).Optionen.Expand
					}
				}
				f.want = &ref
			}
			pruefeFall(t, p, f)
			n++
		}
	}
	exZaehlt(t, "Folgen × Queries", n, 40*7)
}

// TestExpandParserUnveraendert (Zeile 23, SPEC-262 Nr. 8, SPEC-160 Nr. 7): ein Aufruf von `ParseURL` schreibt kein Feld des
// Parsers, gleich welche Einhängung er trägt und welche Eingabe er liest. Menge: 4 Parser (ohne, mit Leser, mit `WithExpand`, mit
// beidem) × (die Eingaben von `korpus()`, 490 Eingaben von `exQueries`, 5219 Folgen der Länge 1 bis 3 über dem Alphabet, je mit und
// ohne vorangestelltes `$top=1`, Queries von 8191, 8192 und 8193 Bytes, eine Liste von 32 Namen aus 120 `N` und der Nummer und die
// gültigen Listen von 1 bis 32 Namen, je mit und ohne `$top=1`); der Zustand des
// Parsers wird nach jedem Aufruf gegen eine Kopie vor dem ersten gehalten. Prüfgrenze: ein Schreiben, das nur bei einer Eingabe
// außerhalb dieser Menge greift, bleibt ungefangen; die Menge der Eingaben ist unbeschränkt.
func TestExpandParserUnveraendert(t *testing.T) {
	leser := &zaehlLeser{ausdruck: testAusdruck{"ok"}}
	parser := []*urlparser.Parser{
		urlparser.New(), urlparser.New(urlparser.WithFilter(leser)), exParser(), urlparser.New(urlparser.WithFilter(leser), urlparser.WithExpand()),
	}
	var eingaben [][2]string
	eingaben = append(eingaben, korpus()...)
	for _, q := range exQueries() {
		eingaben = append(eingaben, [2]string{"/S", q})
	}
	for _, f := range exFolgen(3) {
		eingaben = append(eingaben, [2]string{"/S", exQ(f.roh)}, [2]string{"/S", "$top=1&" + exQ(f.roh)})
	}
	exZaehlt(t, "Alphabet bis Länge 3", len(exFolgen(3)), 17+17*17+17*17*17)
	for _, laenge := range []int{8191, 8192, 8193} {
		eingaben = append(eingaben, [2]string{"/S", "x=" + strings.Repeat("v", laenge-2)})
	}
	eingaben = append(eingaben, [2]string{"/S", exQ(strings.Join(exNamen(strings.Repeat("N", 120), 32), ","))})
	for l := 1; l <= 32; l++ {
		eingaben = append(eingaben, [2]string{"/S", exQ(strings.Join(exNamen("n", l), ","))}, [2]string{"/S", "$top=1&" + exQ(strings.Join(exNamen("n", l), ","))})
	}
	n := 0
	for i, p := range parser {
		vorher := *p
		for _, e := range eingaben {
			_, _ = p.ParseURL(e[0], e[1])
			if *p != vorher {
				t.Fatalf("Parser %d nach %q %q: %+v statt %+v", i, e[0], e[1], *p, vorher)
			}
			n++
		}
	}
	exZaehlt(t, "Parser × Eingaben", n, 4*len(eingaben))
}

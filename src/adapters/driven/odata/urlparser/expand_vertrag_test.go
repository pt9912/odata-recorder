package urlparser_test

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestExpandOrt (Zeile 18, SPEC-160 Nr. 4): der Ort ist `query/<n>`, `n` zählt die nichtleeren Segmente. Menge: Stelle der Option 1
// bis 4 × Zahl der leeren Segmente {0, 1, 3} an drei Orten (vorn, zwischen allen, hinten) = 36.
func TestExpandOrt(t *testing.T) {
	p := exParser()
	n := 0
	for stelle := 1; stelle <= 4; stelle++ {
		for _, leer := range []int{0, 1, 3} {
			trenner := strings.Repeat("&", leer+1)
			davor := optionenDavor(stelle - 1)
			davor = strings.ReplaceAll(davor, "&", trenner)
			soll := exFehl(model.URLWert, "query/"+strconv.Itoa(stelle))
			exPruefe(t, p, strings.Repeat("&", leer)+davor+"$expand=1a", soll)
			exPruefe(t, p, davor+"$expand=1a"+strings.Repeat("&", leer), soll)
			exPruefe(t, p, davor+"$expand=1a"+trenner+"zz=1", soll)
			n += 3
		}
	}
	exZaehlt(t, "Ort", n, 36)
}

// TestExpandFehlertext (Zeile 19, SPEC-160 Nr. 4): kein Teil der Eingabe steht im Fehler. Menge 14: jede Art von Verstoß aus Zeile
// 5 bis 16 mit einem Namen, der nur in der Eingabe vorkommt, und die Option ohne Einhängung.
func TestExpandFehlertext(t *testing.T) {
	const m = "Zqx9Kw"
	faelle := []string{
		"$expand=" + m + ",,x", "$expand=" + m + "%20", "$expand=1" + m, "$expand=" + m + ",*", "$expand=" + m + "/q",
		"$expand=" + m + "(x)", "$expand=" + m + "&$expand=" + m, "$expand=" + m + "," + m, "$expand=" + m + "%G1",
		"$expand=" + strings.Join(append([]string{m}, exNamen("N", 33)...), ","), "$expand=" + m + ",1a", "$expand=" + m + ".a/b.",
		"$expand=" + m + ";x", "$expand=a," + m + "/",
	}
	for _, q := range faelle {
		_, err := exParser().ParseURL("/S", q)
		var uf *model.URLFehler
		if !errors.As(err, &uf) || strings.Contains(err.Error(), m) {
			t.Errorf("%q: %v", q, err)
		}
	}
	_, err := urlparser.New().ParseURL("/S", "$expand="+m)
	if err == nil || strings.Contains(err.Error(), m) {
		t.Errorf("ohne Einhängung: %v", err)
	}
	exZaehlt(t, "Fehlertext", len(faelle), 14)
}

// TestExpandKopie (Zeile 20, SPEC-160 Nr. 3): das Ergebnis gehört dem Aufrufer; Änderung der Elemente, des Zeigers und Anhängen
// über die Kapazität ändern kein späteres Ergebnis. Menge: Länge 1 bis 3.
func TestExpandKopie(t *testing.T) {
	p := exParser()
	for laenge := 1; laenge <= 3; laenge++ {
		namen := exNamen("n", laenge)
		query := exQ(strings.Join(namen, ","))
		ref := exKnoten(namen...).Optionen.Expand
		r1, err := p.ParseURL("/S", query)
		r2, _ := p.ParseURL("/S", query)
		if err != nil || r1.Optionen.Expand == r2.Optionen.Expand || &r1.Optionen.Expand.Elemente[0] == &r2.Optionen.Expand.Elemente[0] {
			t.Fatalf("Länge %d: %v, der Zeiger auf Expand oder das Feld der Elemente ist geteilt", laenge, err)
		}
		for i := range r1.Optionen.Expand.Elemente {
			r1.Optionen.Expand.Elemente[i] = model.ExpandElement{Name: "Z"}
		}
		el := r1.Optionen.Expand.Elemente
		r1.Optionen.Expand.Elemente = append(el[:cap(el)], model.ExpandElement{Name: "Z"})
		r3, _ := p.ParseURL("/S", query)
		if !gleich(r2.Optionen.Expand, ref) || !gleich(r3.Optionen.Expand, ref) {
			t.Errorf("Länge %d: das Ergebnis hat sich durch Änderungen eines anderen geändert:\n  %+v\n  %+v", laenge, r2.Optionen.Expand, ref)
		}
	}
}

// exQueries sind die Eingaben der Zeilen 2 bis 16 für die Wiederholung und die Nebenläufigkeit: Listen der Länge 1 bis 4, die
// Verstöße je Art und die Paare der Rangfolge im Wert (432).
func exQueries() []string {
	var qs []string
	for laenge := 1; laenge <= 4; laenge++ {
		qs = append(qs, exQ(strings.Join(exNamen("Nav", laenge), ",")))
	}
	for _, w := range []string{"", ",a", "a,", "a,,b", "a,a", "*", "*,*", "a/b", "a/$ref", "a/*", "1a", "a b", "a)", "a(x)", "a%252C", "a%2Cb",
		strings.Join(exNamen("N", 33), ","), strings.Join(exNamen("N", 32), ",")} {
		qs = append(qs, exQ(w), "$top=1&"+exQ(w), "%24expand="+w+"&$expand="+w)
	}
	art := exVerstoesse()
	for _, a := range art {
		for _, b := range art {
			for i := 1; i <= 4; i++ {
				for j := 1; j <= 4; j++ {
					if i != j {
						el := []string{"n1", "f2", "f3", "f4", "f5"}
						el[i], el[j] = a.element, b.element
						qs = append(qs, exQ(strings.Join(el, ",")))
					}
				}
			}
		}
	}
	return qs
}

// TestExpandWiederholung (Zeile 21, SPEC-160 Nr. 3): dieselbe Eingabe ergibt auf derselben und auf einer anderen Instanz dasselbe
// Ergebnis und denselben Fehler; ein Fehler gibt den Nullwert. Menge: alle Eingaben von `exQueries`.
func TestExpandWiederholung(t *testing.T) {
	a, b := exParser(), exParser()
	qs := exQueries()
	for _, q := range qs {
		u1, e1 := a.ParseURL("/S", q)
		u2, e2 := a.ParseURL("/S", q)
		u3, e3 := b.ParseURL("/S", q)
		if !gleich(u1, u2) || !gleich(u1, u3) || errText(e1) != errText(e2) || errText(e1) != errText(e3) {
			t.Errorf("%q:\n  %+v %v\n  %+v %v\n  %+v %v", q, u1, e1, u2, e2, u3, e3)
		}
		if e1 != nil && !gleich(u1, model.ODataURL{}) {
			t.Errorf("%q: Fehler mit einem Ergebnis ungleich dem Nullwert", q)
		}
	}
	if len(qs) != 4+18*3+432 {
		t.Fatalf("%d Eingaben", len(qs))
	}
}

// TestExpandWaechter (Zeile 22, SPEC-160 Nr. 6 b, c): die Summe der Längen aller Texte im Ergebnis ist höchstens Eingabe plus 48
// und genau die Summe der Namen, die Zuteilung höchstens 64 · n + 65536 Bytes. Menge: 5 Zahlen (1, 2, 16, 31, 32) × 2 Längen (1,
// 128) × 2 Schreibweisen (roh, jedes Byte als `%XX`) = 20, dazu 4 Eingaben aus 32 × 128 Zeichen mit einer Query von genau 8191,
// 8192 und 8193 Bytes (Auffüllung durch eine fremde Option; 8193 ist `zu_lang`) und ein Name aus 128 Zeichen zu 2 Bytes.
func TestExpandWaechter(t *testing.T) {
	p := exParser()
	n := 0
	pruefe := func(query string, namen []string, fremd int) {
		t.Helper()
		var u model.ODataURL
		var err error
		z := gesamtZuteilung(func() { u, err = p.ParseURL("/S", query) })
		if len(query) > 8192 {
			var uf *model.URLFehler
			if !errors.As(err, &uf) || uf.Ursache != model.URLZuLang {
				t.Errorf("Query von %d Bytes: %v", len(query), err)
			}
			return
		}
		summe := len("S") + fremd
		for _, s := range namen {
			summe += len(s)
		}
		if err != nil || u.Optionen.Expand == nil || len(u.Optionen.Expand.Elemente) != len(namen) {
			t.Errorf("Query von %d Bytes: %v", len(query), err)
			return
		}
		if ist := textsumme(u); ist != summe || ist > len("/S")+len(query)+48 {
			t.Errorf("Query von %d Bytes: Textsumme %d, erwartet %d, Eingabe %d", len(query), ist, summe, len("/S")+len(query))
		}
		if grenze := uint64(64*(len("/S")+len(query)) + 65536); z > grenze {
			t.Errorf("Query von %d Bytes: Zuteilung %d über %d", len(query), z, grenze)
		}
	}
	buchstaben := "abcdefghijklmnopqrstuvwxyzABCDEF"
	bauen := func(zahl, laenge int, zeichen string) []string {
		namen := make([]string, zahl)
		for i := range namen {
			namen[i] = buchstaben[i:i+1] + strings.Repeat(zeichen, laenge-1)
		}
		return namen
	}
	for _, zahl := range []int{1, 2, 16, 31, 32} {
		for _, laenge := range []int{1, 128} {
			namen := bauen(zahl, laenge, "x")
			pruefe(exQ(strings.Join(namen, ",")), namen, 0)
			var roh []string
			for _, s := range namen {
				roh = append(roh, exHex(s, true))
			}
			pruefe(exQ(strings.Join(roh, ",")), namen, 0)
			n += 2
		}
	}
	namen := bauen(32, 128, "x")
	basis := exQ(strings.Join(namen, ","))
	for _, ziel := range []int{8191, 8192, 8193} {
		fuell := ziel - len(basis) - len("&p=")
		pruefe(basis+"&p="+strings.Repeat("v", fuell), namen, len("p")+fuell)
		n++
	}
	breit := []string{"a" + strings.Repeat(exE, 127)}
	pruefe(exQ(breit[0]), breit, 0)
	n++
	exZaehlt(t, "Wächter", n, 24)
}

// TestExpandGleichzeitig (Zeile 23, SPEC-160 Nr. 7, Wirkungs-Fänger): 4 Goroutinen × 50 Runden × alle Eingaben von `exQueries` auf
// einer gemeinsamen Instanz und auf einer eigenen; jedes Ergebnis gleicht dem eines einzelnen Aufrufs. Die Ergebnisse werden erst
// nach dem Ende aller Goroutinen verglichen, damit ein geteilter Speicher auch später noch auffällt.
func TestExpandGleichzeitig(t *testing.T) {
	qs := exQueries()
	ref := make([]model.ODataURL, len(qs))
	refFehler := make([]string, len(qs))
	einzel := exParser()
	for i, q := range qs {
		u, err := einzel.ParseURL("/S", q)
		if u.Optionen.Expand != nil { // eine eigene Kopie: der Vergleich teilt keinen Speicher mit den Ergebnissen der Goroutinen
			kopie := model.ExpandOption{Elemente: append([]model.ExpandElement(nil), u.Optionen.Expand.Elemente...)}
			u.Optionen.Expand = &kopie
		}
		ref[i], refFehler[i] = u, errText(err)
	}
	type ergebnis struct {
		i   int
		u   model.ODataURL
		err string
	}
	gemeinsam := exParser()
	var wg sync.WaitGroup
	start := make(chan struct{})
	alle := make([][]ergebnis, 4)
	for g := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eigen := exParser()
			<-start
			for runde := range 50 {
				p := gemeinsam
				if runde%2 == 1 {
					p = eigen
				}
				for i, q := range qs {
					u, err := p.ParseURL("/S", q)
					alle[g] = append(alle[g], ergebnis{i, u, errText(err)})
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	schlecht := 0
	for g := range alle {
		for _, e := range alle[g] {
			if !gleich(e.u, ref[e.i]) || e.err != refFehler[e.i] {
				if schlecht++; schlecht <= 5 {
					t.Errorf("Goroutine %d, Eingabe %q: %+v %q, erwartet %+v %q", g, qs[e.i], e.u, e.err, ref[e.i], refFehler[e.i])
				}
			}
		}
	}
	if schlecht > 0 {
		t.Errorf("%d abweichende Ergebnisse von %d", schlecht, 4*50*len(qs))
	}
}

// TestExpandAlphabet (Zeile 24, SPEC-262 Nr. 3, 4): jede Folge der Länge 1 bis 4 über den 17 Zeichen des Alphabets hat die Klasse
// des Orakels (Knoten mit den Namen, `wert`, `wert_nicht_unterstuetzt`). Menge 88 740 = 17 + 17² + 17³ + 17⁴. Prüfgrenze: längere
// Folgen nur als Familien der Zeilen 9, 14 und 22; eine Regel, die erst ab Länge 5 greift, bleibt ungefangen.
func TestExpandAlphabet(t *testing.T) {
	p := exParser()
	folgen := exFolgen(4)
	exZaehlt(t, "Alphabet", len(folgen), 17+17*17+17*17*17+17*17*17*17)
	klassen := map[model.URLUrsache]int{}
	for _, f := range folgen {
		soll := exOrakel(f.dek)
		klassen[soll.u]++
		exPruefe(t, p, exQ(f.roh), soll)
	}
	for _, u := range []model.URLUrsache{"", model.URLWert, model.URLWertNichtUnterstuetzt} {
		if klassen[u] < 100 {
			t.Errorf("Klasse %q: nur %d Folgen", u, klassen[u])
		}
	}
}

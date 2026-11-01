package urlparserv2_test

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

type ergebnis struct {
	url model.ODataURL
	err string
}

func rufe(p *urlparserv2.Parser, pfad, q string) ergebnis {
	u, err := p.ParseURL(pfad, q)
	if err != nil {
		return ergebnis{u, fmt.Sprintf("%T %v", err, err)}
	}
	return ergebnis{url: u}
}

// schwer ist ein Wert aus 300 Escapes desselben Zeichens: die Dekodierung braucht Zeit, und die Ergebnisse verschiedener Eingaben
// unterscheiden sich, so dass ein geteilter Puffer der Dekodierung sichtbar wird.
func schwer(c byte) string { return "x=" + strings.Repeat(fmt.Sprintf("%%%02X", c), 300) }

// eingaben40 sind 40 Eingaben: 14 Erfolge (sechs mit schwerer Dekodierung), 14 Verstöße der Klasse `ungueltig`, 12
// Kompatibilitätsfehler, verschränkt gerufen.
func eingaben40() [][2]string {
	return [][2]string{
		{"/S", schwer('A')}, {"/S", ""}, {"/S(1)", ""}, {"/S(A=1,B='x')", ""}, {"/S/$count", ""}, {"/S", "$top=1&$skip=2"},
		{"/S", schwer('B')}, {"/S", schwer('C')}, {"/S", schwer('D')}, {"/S", schwer('E')}, {"/S", "x=1&y=2"},
		{"/S(guid'AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE')", ""}, {"/S(datetime'2020-01-02T03:04')", ""}, {"/S", schwer('F')},
		{"/S%", ""}, {"/S", "$top=x"}, {"/S(1,,2)", ""}, {"/S", "$foo=1"}, {"/S", "%FF=1"}, {"/S/", ""}, {"/S(null)", ""},
		{"/S", "$top=1&$top=2"}, {"/$metadata/x", ""}, {"/S", "$select="}, {"/S", "$inlinecount=ALL"}, {"/S(Id=1,Id=2)", ""}, {"//", ""},
		{"/", ""}, {"/S/N", ""}, {"/S(1d)", ""}, {"/S", "$expand=1"}, {"/S", "$format=xml"}, {"/S", "$filter=a"}, {"/S(1)/$value", ""},
		{"/$batch", ""}, {"/S", "$orderby=tolower(N)"}, {"/S", "$select=A/B"}, {"/S", "$count=true"}, {"/S/N", "$top=x"},
		{"/S(1e3)", "$top=1"},
	}
}

// TestGleichzeitig (Zeile 17): gleichzeitige Aufrufe beeinflussen sich nicht, auf einer Instanz und auf zwei; jedes Ergebnis ist
// gleich dem eines einzelnen Aufrufs, der erste Kompatibilitätsfehler eines Laufs ist nie Zustand der Instanz. Menge: 16 Goroutinen
// × 40 Eingaben × 10 Runden, auf einer und auf zwei Instanzen. Der Fänger zählt die abweichenden Ergebnisse (Wirkung) und ist vom
// Detektor unabhängig; `make test-race` meldet das Fehlen einer Sperre.
func TestGleichzeitig(t *testing.T) {
	eingaben := eingaben40()
	if len(eingaben) != 40 {
		t.Fatalf("%d Eingaben, erwartet 40", len(eingaben))
	}
	bezug := urlparserv2.New()
	soll := make([]ergebnis, len(eingaben))
	arten := map[string]int{}
	for i, e := range eingaben {
		soll[i] = rufe(bezug, e[0], e[1])
		switch {
		case soll[i].err == "":
			arten["erfolg"]++
		case soll[i].err[:len("*model.URLFehler")] == "*model.URLFehler":
			arten["ungueltig"]++
		default:
			arten["kompat"]++
		}
	}
	if arten["erfolg"] != 14 || arten["ungueltig"] != 14 || arten["kompat"] != 12 {
		t.Fatalf("Arten der Eingaben %v, erwartet 14, 14, 12", arten)
	}
	for _, instanzen := range []int{1, 2} {
		ps := make([]*urlparserv2.Parser, instanzen)
		for i := range ps {
			ps[i] = urlparserv2.New()
		}
		var abweichend atomic.Int64
		var wg sync.WaitGroup
		start := make(chan struct{})
		for g := range 16 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p := ps[g%instanzen]
				<-start
				for range 10 {
					for k := range eingaben {
						i := (k + g*7) % len(eingaben)
						if got := rufe(p, eingaben[i][0], eingaben[i][1]); !reflect.DeepEqual(got, soll[i]) {
							abweichend.Add(1)
						}
					}
				}
			}()
		}
		close(start)
		wg.Wait()
		if abweichend.Load() != 0 {
			t.Errorf("%d Instanzen: %d von %d Ergebnissen weichen vom einzelnen Aufruf ab", instanzen, abweichend.Load(), 16*40*10)
		}
	}
}

// TestEinhaengungGleichzeitig (Zeile 18): der eingehängte FilterLeser wird aus mehreren Goroutinen gerufen, der Parser verändert
// nichts an ihm. Menge: 16 Goroutinen × 4 Eingaben (mit Wert, leer, doppelt, `ungueltig` davor) × 10 Runden; die Summe der
// Aufrufe des Zähl-Fakes ist gleich der Zahl der Eingaben, die ihn erreichen (320).
func TestEinhaengungGleichzeitig(t *testing.T) {
	l := &fakeFilterLeser{}
	p := urlparserv2.New(urlparserv2.WithFilter(l))
	eingaben := []struct {
		q        string
		erreicht bool
	}{{"$filter=a", true}, {"$filter=", false}, {"$filter=a&$filter=b", true}, {"$top=x&$filter=a", false}}
	soll := make([]ergebnis, len(eingaben))
	ref := urlparserv2.New(urlparserv2.WithFilter(&fakeFilterLeser{}))
	for i, e := range eingaben {
		soll[i] = rufe(ref, "/S", e.q)
	}
	var abweichend atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for g := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 10 {
				for k := range eingaben {
					i := (k + g) % len(eingaben)
					if got := rufe(p, "/S", eingaben[i].q); !reflect.DeepEqual(got, soll[i]) {
						abweichend.Add(1)
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	erreichend := 0
	for _, e := range eingaben {
		if e.erreicht {
			erreichend++
		}
	}
	if erwartet := int64(16 * 10 * erreichend); l.n.Load() != erwartet || int64(len(l.aufrufe())) != erwartet {
		t.Errorf("Aufrufe des Lesers %d (festgehalten %d), erwartet %d", l.n.Load(), len(l.aufrufe()), erwartet)
	}
	for _, a := range l.aufrufe() {
		if a != "a" {
			t.Fatalf("der Leser bekam %q statt a", a)
		}
	}
	if abweichend.Load() != 0 {
		t.Errorf("%d Ergebnisse weichen vom einzelnen Aufruf ab", abweichend.Load())
	}
}

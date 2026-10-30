package urlparser_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// korpus sind Gut- und Fehlerfälle aus allen Klassen: Pfade und Schlüssel, Optionen, die Ursachen der Verstöße und erzeugte URLs.
func korpus() [][2]string {
	k := [][2]string{
		{"/", ""}, {"/$metadata", ""}, {"/Set", "$top=0"}, {"/Set(1)", ""}, {"/Set(Id=42,Region='EU')", "$select=Id,Name&$orderby=Id%20desc,Name"},
		{"/Set/$count", "$filter=a"}, {"/Set(1)/$count", ""}, {"/Set(1)/Rest", ""}, {"/S('O''Hara')", "x=1&y=2&z=3&w=4&v=5&u=6"},
		{"/S(Id=0f8fad5b-d9cb-469f-a165-70867728950e)", "$count=true&$skip=007&$format=json"}, {"/S(D=2020-01-01,T=2020-01-01T00:00:00%2B01:00)", "@p=1&sap-client=100"},
		{"/S(X'AA')", ""}, {"/S(1x)", ""}, {"/S", "$top=1&$top=2"}, {"/S", "$expand=1&$foo=2"}, {"/S", "$select=a/b"}, {"/S", "$orderby=length(a)"},
		{"/S/%G1", ""}, {"/S", "x=%G1"}, {"/" + wiederhole("a", 129), ""}, {"/S(" + wiederhole("a", 129) + "=1)", ""}, {"/S-(1)", ""}, {"/S/a%20b", ""}, {"/S/Items(", ""}, {"/S", "$select=*,a,b,c&$orderby=a,b%20desc,c"}, {"/S", "b=1&a=2&c=3&d=4&e=5&f=6&g=7"},
	}
	for _, f := range ursachenFaelle() {
		k = append(k, [2]string{markiert(f.pfad), markiert(f.query)})
	}
	for saat := uint64(1); saat <= 25; saat++ {
		r := rng(saat)
		pfad, _, _ := genPfad(r, alleRessourcen()[saat%4])
		raw, _, _ := genQuery(r, genOptionen(r, optionKlassen()[saat%6]))
		k = append(k, [2]string{pfad, raw})
	}
	return k
}

// TestWiederholung (Zeile 13): gleiche Eingabe, gleiches Ergebnis und gleicher Fehler, auch auf einer anderen Instanz.
func TestWiederholung(t *testing.T) {
	a, b := parserNeu(), parserNeu()
	for _, f := range korpus() {
		u1, e1 := a.ParseURL(f[0], f[1])
		u2, e2 := a.ParseURL(f[0], f[1])
		u3, e3 := b.ParseURL(f[0], f[1])
		if !gleich(u1, u2) || !gleich(u1, u3) || fmt.Sprint(e1) != fmt.Sprint(e2) || fmt.Sprint(e1) != fmt.Sprint(e3) {
			t.Errorf("%q %q:\n  %+v %v\n  %+v %v\n  %+v %v", f[0], f[1], u1, e1, u2, e2, u3, e3)
		}
	}
}

// TestErgebnisTeiltNichts (Zeile 12): das Ergebnis teilt kein Slice und keinen Zeiger mit einem anderen; vier Felder
// (Schluessel, Select.Namen, OrderBy, Sonstige) und der Zeiger auf Select.
func TestErgebnisTeiltNichts(t *testing.T) {
	pfad, query := "/S(A=1,B='x',C=true)", "$select=a,b,c&$orderby=a,b%20desc,c&x=1&y=2&z=3"
	p := parserNeu()
	ref, _ := p.ParseURL(pfad, query)
	r1, _ := p.ParseURL(pfad, query)
	r2, _ := p.ParseURL(pfad, query)
	if r1.Optionen.Select == r2.Optionen.Select {
		t.Error("der Zeiger auf Select ist geteilt")
	}
	for i := range r1.Ressource.Schluessel {
		r1.Ressource.Schluessel[i] = model.SchluesselTeil{Name: "Z"}
	}
	r1.Ressource.Schluessel = append(r1.Ressource.Schluessel[:cap(r1.Ressource.Schluessel)], model.SchluesselTeil{Name: "Z"})
	for i := range r1.Optionen.Select.Namen {
		r1.Optionen.Select.Namen[i] = "Z"
	}
	r1.Optionen.Select.Alle = true
	r1.Optionen.Select.Namen = append(r1.Optionen.Select.Namen[:cap(r1.Optionen.Select.Namen)], "Z")
	for i := range r1.Optionen.OrderBy {
		r1.Optionen.OrderBy[i] = model.Sortierung{Name: "Z"}
	}
	r1.Optionen.OrderBy = append(r1.Optionen.OrderBy[:cap(r1.Optionen.OrderBy)], model.Sortierung{Name: "Z"})
	for i := range r1.Optionen.Sonstige {
		r1.Optionen.Sonstige[i] = model.SonstigeOption{Name: "Z"}
	}
	r1.Optionen.Sonstige = append(r1.Optionen.Sonstige[:cap(r1.Optionen.Sonstige)], model.SonstigeOption{Name: "Z"})
	r3, _ := p.ParseURL(pfad, query)
	if !gleich(r2, ref) || !gleich(r3, ref) {
		t.Errorf("das Ergebnis hat sich durch Änderungen eines anderen geändert:\n  %+v\n  %+v", r2, ref)
	}
}

// haltLeser hält den Aufruf, der ihn erreicht, an, bis der Test ihn weiterlaufen lässt.
type haltLeser struct{ erreicht, weiter chan struct{} }

func (h haltLeser) LeseFilter(string) (interface{ IstFilterAusdruck() }, error) {
	h.erreicht <- struct{}{}
	<-h.weiter
	return testAusdruck{"ok"}, nil
}

// TestVerschraenkt (Zeile 14): ein Aufruf bleibt im FilterLeser stehen, ein zweiter läuft vollständig durch, dann läuft der
// erste weiter; jeder Aufruf liefert, was er allein lieferte. Die Verschränkung ist erzwungen, kein Zufall: ein Zustand,
// den ein Aufruf vor dem Leser schreibt und danach liest, wird so in jedem Lauf gefunden. Menge 2: der haltende Aufruf hat
// einen Verstoß der Klasse `nicht_unterstuetzt` vor dem Leser und der andere keinen, und umgekehrt.
func TestVerschraenkt(t *testing.T) {
	faelle := []struct{ halt, frei [2]string }{
		{[2]string{"/$batch", "$filter=a"}, [2]string{"/S", "x=1"}},
		{[2]string{"/S", "$filter=a"}, [2]string{"/$batch", "x=1"}},
	}
	for _, f := range faelle {
		h := haltLeser{erreicht: make(chan struct{}), weiter: make(chan struct{})}
		p := urlparser.New(urlparser.WithFilter(h))
		sollHalt, sollHaltFehler := urlparser.New(urlparser.WithFilter(&zaehlLeser{ausdruck: testAusdruck{"ok"}})).ParseURL(f.halt[0], f.halt[1])
		sollFrei, sollFreiFehler := parse(f.frei[0], f.frei[1])
		type ergebnis struct {
			u   model.ODataURL
			err error
		}
		fertig := make(chan ergebnis)
		go func() {
			u, err := p.ParseURL(f.halt[0], f.halt[1])
			fertig <- ergebnis{u, err}
		}()
		<-h.erreicht
		frei, freiFehler := p.ParseURL(f.frei[0], f.frei[1])
		close(h.weiter)
		got := <-fertig
		if !gleich(got.u, sollHalt) || fmt.Sprint(got.err) != fmt.Sprint(sollHaltFehler) {
			t.Errorf("haltender Aufruf %q: %+v %v, allein %+v %v", f.halt, got.u, got.err, sollHalt, sollHaltFehler)
		}
		if !gleich(frei, sollFrei) || fmt.Sprint(freiFehler) != fmt.Sprint(sollFreiFehler) {
			t.Errorf("freier Aufruf %q: %+v %v, allein %+v %v", f.frei, frei, freiFehler, sollFrei, sollFreiFehler)
		}
	}
}

// TestGleichzeitig (Zeile 14): 16 Goroutinen rufen den Korpus auf einer Instanz; jedes Ergebnis ist das des einzelnen Aufrufs.
func TestGleichzeitig(t *testing.T) {
	p := parserNeu()
	faelle := korpus()
	soll := make([]model.ODataURL, len(faelle))
	sollFehler := make([]string, len(faelle))
	for i, f := range faelle {
		u, err := p.ParseURL(f[0], f[1])
		soll[i], sollFehler[i] = u, fmt.Sprint(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{}) // alle Goroutinen beginnen zugleich
	for g := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for round := range 10 {
				for i := range faelle {
					j := (i + g*7 + round) % len(faelle)
					u, err := p.ParseURL(faelle[j][0], faelle[j][1])
					if !gleich(u, soll[j]) || fmt.Sprint(err) != sollFehler[j] {
						t.Errorf("Goroutine %d: %q %q: %+v %v", g, faelle[j][0], faelle[j][1], u, err)
						return
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

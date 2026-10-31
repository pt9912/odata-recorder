package urlparser_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

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

// wartegrenze begrenzt jedes Warten auf einen Aufruf: ein Aufruf, der den Leser nicht erreicht oder nicht endet, macht den
// Test rot und lässt ihn nicht bis zur Zeitgrenze von `go test` hängen.
const wartegrenze = 10 * time.Second

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
		fertig := make(chan ergebnis, 1)
		go func() {
			u, err := p.ParseURL(f.halt[0], f.halt[1])
			fertig <- ergebnis{u, err}
		}()
		var gelassen sync.Once
		lassen := func() { gelassen.Do(func() { close(h.weiter) }) }
		defer lassen()
		select {
		case <-h.erreicht:
		case <-time.After(wartegrenze):
			t.Fatalf("haltender Aufruf %q erreicht den Leser nicht innerhalb von %v", f.halt, wartegrenze)
		}
		frei, freiFehler := p.ParseURL(f.frei[0], f.frei[1])
		lassen()
		var got ergebnis
		select {
		case got = <-fertig:
		case <-time.After(wartegrenze):
			t.Fatalf("haltender Aufruf %q endet nicht innerhalb von %v", f.halt, wartegrenze)
		}
		if !gleich(got.u, sollHalt) || fmt.Sprint(got.err) != fmt.Sprint(sollHaltFehler) {
			t.Errorf("haltender Aufruf %q: %+v %v, allein %+v %v", f.halt, got.u, got.err, sollHalt, sollHaltFehler)
		}
		if !gleich(frei, sollFrei) || fmt.Sprint(freiFehler) != fmt.Sprint(sollFreiFehler) {
			t.Errorf("freier Aufruf %q: %+v %v, allein %+v %v", f.frei, frei, freiFehler, sollFrei, sollFreiFehler)
		}
	}
}

// paarLeser lässt zwei Aufrufe gleichzeitig im Leser stehen und gibt sie in der Reihenfolge frei, die der Test wählt; er gibt den
// Text, den er bekam, als Ausdruck zurück, damit eine Vermischung der Aufrufe im Ergebnis sichtbar wird.
type paarLeser struct {
	erreicht chan string
	frei     map[string]chan struct{}
}

func (p paarLeser) LeseFilter(text string) (interface{ IstFilterAusdruck() }, error) {
	p.erreicht <- text
	<-p.frei[text]
	return testAusdruck{text}, nil
}

// pfadErgebnis ist ein Aufruf des Paares: ein Pfad mit der Ressource, die der Parser allein daraus liest (von Hand), und die
// Optionen, die in der Query vor `$filter` stehen, mit dem Wert, den sie allein ergeben (von Hand, ohne Filter).
type pfadErgebnis struct {
	pfad     string
	res      model.Ressource
	optionen string
	opt      model.Optionen
}

// query ist die Query des Aufrufs: die Optionen vor `$filter`, dann der eigene Text des Filters.
func (p pfadErgebnis) query(text string) string {
	if p.optionen == "" {
		return "$filter=" + text
	}
	return p.optionen + "&$filter=" + text
}

// verschraenktPfade sind die Pfade von A (Alpha) und von B (Beta): Menge, Entity, Anzahl und `$metadata`.
func verschraenktPfade() (a, b []pfadErgebnis) {
	for _, d := range []struct {
		name, schluessel string
		dst              *[]pfadErgebnis
	}{{"Alpha", "1", &a}, {"Beta", "2", &b}} {
		*d.dst = []pfadErgebnis{
			{pfad: "/" + d.name, res: model.Ressource{Art: model.RessourceMenge, EntitySet: d.name}},
			{pfad: "/" + d.name + "(" + d.schluessel + ")", res: model.Ressource{Art: model.RessourceEntity, EntitySet: d.name, Schluessel: []model.SchluesselTeil{teil("", zahlLit(d.schluessel))}}},
			{pfad: "/" + d.name + "/$count", res: model.Ressource{Art: model.RessourceAnzahl, EntitySet: d.name}},
			{pfad: "/$metadata", res: model.Ressource{Art: model.RessourceMetadaten}},
		}
	}
	return a, b
}

// verschraenktOptionen sind die sieben Arten von Option vor `$filter` (SPEC-162 Nr. 3 a, d, e): je Art zwei Aufrufe mit
// verschiedenem Wert, A auf `/Alpha` und B auf `/Beta`, jeweils mit dem Ergebnis von Hand.
func verschraenktOptionen() [][2]pfadErgebnis {
	menge := func(name string) model.Ressource { return model.Ressource{Art: model.RessourceMenge, EntitySet: name} }
	paar := func(a, b string, oa, ob model.Optionen) [2]pfadErgebnis {
		return [2]pfadErgebnis{
			{pfad: "/Alpha", res: menge("Alpha"), optionen: a, opt: oa},
			{pfad: "/Beta", res: menge("Beta"), optionen: b, opt: ob},
		}
	}
	zahl := func(v int) model.ZahlOption { return model.ZahlOption{Gesetzt: true, Wert: v} }
	wahr := func(v bool) model.WahrOption { return model.WahrOption{Gesetzt: true, Wert: v} }
	sonstige := func(name, wert string) model.Optionen {
		return model.Optionen{Sonstige: []model.SonstigeOption{{Name: name, Wert: wert}}}
	}
	return [][2]pfadErgebnis{
		paar("$select=Id", "$select=Name",
			model.Optionen{Select: &model.SelectOption{Namen: []string{"Id"}}}, model.Optionen{Select: &model.SelectOption{Namen: []string{"Name"}}}),
		paar("$orderby=Id", "$orderby=Name desc",
			model.Optionen{OrderBy: []model.Sortierung{{Name: "Id"}}}, model.Optionen{OrderBy: []model.Sortierung{{Name: "Name", Absteigend: true}}}),
		paar("$top=1", "$top=2", model.Optionen{Top: zahl(1)}, model.Optionen{Top: zahl(2)}),
		paar("$skip=3", "$skip=4", model.Optionen{Skip: zahl(3)}, model.Optionen{Skip: zahl(4)}),
		paar("$count=true", "$count=false", model.Optionen{Count: wahr(true)}, model.Optionen{Count: wahr(false)}),
		paar("@a=1", "@a=2", sonstige("@a", "1"), sonstige("@a", "2")),
		paar("x=1", "x=2", sonstige("x", "1"), sonstige("x", "2")),
	}
}

// TestVerschraenktLeser (Zeile 14, SPEC-160 Nr. 7 und 8): zwei Aufrufe mit verschiedenen Pfaden laufen beide durch den
// FilterLeser, stehen zugleich in ihm und werden in beiden Reihenfolgen freigegeben; der zuerst freigegebene läuft ganz durch,
// während der andere im Leser steht. Jeder Aufruf liefert, was er allein lieferte: die Ressource seines Pfads (von Hand) sowie
// Text und Ausdruck seines `$filter`. Menge 16 Paare: Pfad von A (`/Alpha`, `/Alpha(1)`, `/Alpha/$count`, `/$metadata`) × Pfad von B
// (`/Beta`, `/Beta(2)`, `/Beta/$count`, `/$metadata`), jedes in beiden Reihenfolgen der Freigabe, 32 Läufe; das Paar `/$metadata`
// gegen `/$metadata` ist der Kontrollfall mit gleicher Ressource. Ein Zustand der Instanz, den ein Aufruf nach dem Pfad schreibt
// und am Ende liest, auch unter einer Sperre und auch nur bei nichtleerem Set, wird in jedem Lauf gefunden; die Wirkung wird gezählt,
// kein Detektor. Dazu 14 Läufe für das Fenster „Option gelesen, vor dem Leser“: sieben Arten von Option (`$select`, `$orderby`,
// `$top`, `$skip`, `$count`, Alias, herstellerspezifisch) mit verschiedenem Wert in A und B, die Option steht in der Query vor
// `$filter`, jede Art in beiden Reihenfolgen der Freigabe; ein Zustand der Instanz, den eine Option vor dem Leser schreibt und
// am Ende liest, wird in jedem Lauf gefunden.
func TestVerschraenktLeser(t *testing.T) {
	pa, pb := verschraenktPfade()
	laeufe := 0
	for _, a := range pa {
		for _, b := range pb {
			laeufe += verschraenktPaar(t, a, b)
		}
	}
	if laeufe != 32 {
		t.Fatalf("%d Läufe der Pfade, erwartet 32 (16 Paare in zwei Reihenfolgen)", laeufe)
	}
	laeufe = 0
	for _, paar := range verschraenktOptionen() {
		laeufe += verschraenktPaar(t, paar[0], paar[1])
	}
	if laeufe != 14 {
		t.Fatalf("%d Läufe der Optionen, erwartet 14 (7 Arten in zwei Reihenfolgen)", laeufe)
	}
}

// verschraenktPaar fährt das Paar in beiden Reihenfolgen der Freigabe und liefert die Zahl der Läufe.
func verschraenktPaar(t *testing.T, ra, rb pfadErgebnis) int {
	t.Helper()
	pfade := map[string]pfadErgebnis{"a": ra, "b": rb}
	laeufe := 0
	for _, zuerst := range []string{"a", "b"} {
		laeufe++
		h := paarLeser{erreicht: make(chan string, 2), frei: map[string]chan struct{}{"a": make(chan struct{}), "b": make(chan struct{})}}
		p := urlparser.New(urlparser.WithFilter(h))
		type ergebnis struct {
			u   model.ODataURL
			err error
		}
		fertig := map[string]chan ergebnis{"a": make(chan ergebnis, 1), "b": make(chan ergebnis, 1)}
		for _, text := range []string{"a", "b"} {
			go func() {
				u, err := p.ParseURL(pfade[text].pfad, pfade[text].query(text))
				fertig[text] <- ergebnis{u, err}
			}()
		}
		for range 2 {
			select {
			case <-h.erreicht:
			case <-time.After(wartegrenze):
				t.Fatalf("nicht beide Aufrufe erreichen den Leser innerhalb von %v", wartegrenze)
			}
		}
		anderer := map[string]string{"a": "b", "b": "a"}[zuerst]
		for _, text := range []string{zuerst, anderer} {
			close(h.frei[text])
			select {
			case got := <-fertig[text]:
				optionen := pfade[text].opt
				optionen.Filter = &model.FilterOption{Text: text, Ausdruck: testAusdruck{text}}
				soll := model.ODataURL{Ressource: pfade[text].res, Optionen: optionen}
				if got.err != nil || !gleich(got.u, soll) {
					t.Errorf("Pfade %q und %q, Freigabe %q zuerst, Aufruf %q: %+v %v, allein %+v", ra.pfad, rb.pfad, zuerst, text, got.u, got.err, soll)
				}
			case <-time.After(wartegrenze):
				t.Fatalf("Aufruf %q endet nicht innerhalb von %v", text, wartegrenze)
			}
		}
	}
	return laeufe
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

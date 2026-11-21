package mock_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// TestTelemetrieImQuelltext hält Q2 und Q12 (SPEC-290 Nr. 1, 6): im Quelltext des Pakets ohne _test.go ruft der Dienst von Telemetrie und
// Span allein StartSpan (einmal: T1), Set (zweimal: T2 und T3) und End (einmal: T4); nie Log, CountError oder Snapshot. Die Zählung
// trägt die Stellen der Query; T5 und T6 (die Mutation) stehen in SPEC-290 Nr. 6 und gehören nicht zu ihr.
func TestTelemetrieImQuelltext(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	zaehler := map[string]int{}
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if ce, ok := n.(*ast.CallExpr); ok {
				if sel, ok := ce.Fun.(*ast.SelectorExpr); ok {
					zaehler[sel.Sel.Name]++
				}
			}
			return true
		})
	}
	soll := map[string]int{"StartSpan": 1, "Set": 2, "End": 1, "Log": 0, "CountError": 0, "Snapshot": 0}
	for name, n := range soll {
		if zaehler[name] != n {
			t.Errorf("%s im Quelltext %d-mal, erwartet %d", name, zaehler[name], n)
		}
	}
}

// Die Stellen mit Request-Wert der Query (SPEC-290 Nr. 8): Set, Schlüssel, Literal in $filter, Name in $filter, Name in $orderby, Name
// in $select, Header X-Recorder-Workspace und Accept; jede mit den Arten, die sie liest.
type stelleQuery struct {
	name string
	// aufbau trägt den Kanarienwert an der Stelle in den Request und liefert die Anfrage.
	aufbau func(u *umg, art model.RessourcenArt) inbound.Anfrage
	arten  []model.RessourcenArt
}

func stellenQuery() []stelleQuery {
	menge, entity, anzahl := model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl
	k := kanarienwert
	mitOpt := func(opt model.Optionen) func(u *umg, art model.RessourcenArt) inbound.Anfrage {
		return func(u *umg, art model.RessourcenArt) inbound.Anfrage {
			u.ressource(art, setProducts, opt, zahl("", "1"))
			return anfrage("GET", dienstRoot+"/x")
		}
	}
	return []stelleQuery{
		{"Set", func(u *umg, art model.RessourcenArt) inbound.Anfrage {
			u.ressource(art, k+"Set", model.Optionen{}, zahl("", "1"))
			return anfrage("GET", dienstRoot+"/x")
		}, []model.RessourcenArt{menge, entity, anzahl}},
		{"Schlüssel", func(u *umg, art model.RessourcenArt) inbound.Anfrage {
			u.ressource(art, setProducts, model.Optionen{}, model.SchluesselTeil{Name: k, Literal: model.Literal{Art: model.LiteralZeichenkette, Text: "'" + k + "'"}})
			return anfrage("GET", dienstRoot+"/x")
		}, []model.RessourcenArt{entity}},
		{"Literal in $filter", mitOpt(model.Optionen{Filter: filterOpt(nameGleich(k))}), []model.RessourcenArt{menge, anzahl}},
		{"Name in $filter", mitOpt(model.Optionen{Filter: filterOpt(vergleich(model.VergleichEq, k, model.Literal{Art: model.LiteralZahl, Text: "1"}))}),
			[]model.RessourcenArt{menge, anzahl}},
		{"Name in $orderby", mitOpt(model.Optionen{OrderBy: []model.Sortierung{{Name: k}}}), []model.RessourcenArt{menge}},
		{"Name in $select", mitOpt(model.Optionen{Select: &model.SelectOption{Namen: []string{k}}}), []model.RessourcenArt{menge, entity}},
		{"Header X-Recorder-Workspace", func(u *umg, art model.RessourcenArt) inbound.Anfrage {
			u.ws.setze(model.Workspace{ID: k, ServiceID: dienstID, BaselineSnapshotID: baselineID})
			u.ressource(art, setProducts, model.Optionen{}, zahl("", "1"))
			a := anfrage("GET", dienstRoot+"/x")
			a.Header[headerName] = []string{k}
			return a
		}, []model.RessourcenArt{menge, entity, anzahl}},
		{"Accept", func(u *umg, art model.RessourcenArt) inbound.Anfrage {
			u.ressource(art, setProducts, model.Optionen{}, zahl("", "1"))
			a := anfrage("GET", dienstRoot+"/x")
			a.Header["Accept"] = []string{k}
			return a
		}, []model.RessourcenArt{menge, entity, anzahl}},
	}
}

// TestKeinRequestWertImSpan hält Q14 und L15 (SPEC-290 Nr. 8): ein Kanarienwert an jeder der
// 8 Stellen des Requests, bei jeder Art, die sie liest, erreicht kein Attribut eines Spans und keinen Aufruf von Log. Jede Anfrage
// besteht die Zuordnung (auch das unbekannte Set trägt den Span mit dem Platzhalter) und verlangt genau einen StartSpan.
func TestKeinRequestWertImSpan(t *testing.T) {
	stellen := stellenQuery()
	if len(stellen) != 8 {
		t.Fatalf("%d Stellen, erwartet 8 (SPEC-290 Nr. 8: die Hälfte der 14 Stellen liest die Query)", len(stellen))
	}
	paare := 0
	for _, s := range stellen {
		for _, art := range s.arten {
			paare++
			u := leseUmg(t)
			a := s.aufbau(u, art)
			_, _ = u.rufe(a)
			if starts := u.tel.Anzahl(supporttest.OpStartSpan); starts != 1 {
				t.Errorf("%s Art %d: StartSpan %d-mal, erwartet 1", s.name, art, starts)
			}
			for _, x := range alleAttribute(u.tel.FakeTelemetry) {
				if strings.Contains(x.Text, kanarienwert) {
					t.Errorf("%s Art %d: Attribut %s trägt den Kanarienwert", s.name, art, x.Key)
				}
			}
			if u.tel.Anzahl(supporttest.OpLog) != 0 {
				t.Errorf("%s Art %d: Log gerufen", s.name, art)
			}
		}
	}
	if paare != 3+1+2+2+1+2+3+3 {
		t.Errorf("%d Paare, erwartet 17", paare)
	}
}

// TestKeinFehlertextImSpan hält SPEC-290 Nr. 8 (der Text eines Fehlers): trägt der Fehler eines Ports an einer der Stellen S3 bis S11
// den Kanarienwert, erreicht er kein Attribut des Spans; der Span beginnt, wo der Fehler entsteht, und endet einmal.
func TestKeinFehlertextImSpan(t *testing.T) {
	n := 0
	for _, s := range rueckwege() {
		if s.op == "" {
			continue
		}
		n++
		u := leseUmg(t)
		s.bau(u)
		u.p.setzeFehler(s.op, errors.New("roh "+kanarienwert))
		_, _ = u.leseGET()
		erwarteEinEnde(t, s.stelle, u)
		for _, x := range alleAttribute(u.tel.FakeTelemetry) {
			if strings.Contains(x.Text, kanarienwert) {
				t.Errorf("%s: Attribut %s trägt den Text des Fehlers", s.stelle, x.Key)
			}
		}
	}
	if n != 9 {
		t.Errorf("%d Stellen, erwartet 9 (S3 bis S11)", n)
	}
}

// zyklus ist eine Schranke, die n Goroutinen je Runde zusammenführt; wer nicht in der Frist ankommt, scheitert.
type zyklus struct {
	n      int
	mu     sync.Mutex
	kommen int
	tor    chan struct{}
}

func neuerZyklus(n int) *zyklus { return &zyklus{n: n, tor: make(chan struct{})} }

func (z *zyklus) warte(frist time.Duration) bool {
	z.mu.Lock()
	tor := z.tor
	z.kommen++
	if z.kommen == z.n {
		z.kommen = 0
		z.tor = make(chan struct{})
		close(tor)
		z.mu.Unlock()
		return true
	}
	z.mu.Unlock()
	select {
	case <-tor:
		return true
	case <-time.After(frist):
		return false
	}
}

// modell16 ist ein Modell mit 16 Sets S0 bis S15 des Typs Product.
func modell16() model.ServiceModel {
	sets := make([]model.EntitySet, 16)
	for i := range sets {
		sets[i] = model.EntitySet{Name: "S" + strconv.Itoa(i), EntityType: typProduct}
	}
	typ := model.EntityType{Name: typProduct, Keys: []string{"Id"}, Properties: produktProperties()}
	return model.NeuesServiceModel("4.0", []string{"N"}, []model.EntityType{typ}, sets, nil)
}

// TestSpanGleichzeitig hält Q13 (SPEC-290 Nr. 9 a): n gleichzeitige Anfragen ergeben je Span-Name n StartSpan und n End, jeder Span
// endet genau einmal und trägt die Attribute seiner Anfrage. 16 Goroutinen × 6 Operationen (GET und HEAD an Collection, Entity,
// Anzahl) × 8 Läufe = 768 Spans, je Goroutine ein eigenes Set und ein eigener Workspace mit eigenem Status; eine Schranke an jedem Aufruf
// eines Ports hinter dem Beginn des Spans (4 bis 5 je Anfrage) lässt alle 16 Spans zugleich offen sein, bevor einer endet, und führt die
// Goroutinen je Stufe zusammen (Wirkungs-Fänger: das Buch je Span, kein Datenrennen nötig; auch ein Attribut aus einer geteilten Stelle
// unter einer Sperre zeigt die Werte anderer Anfragen).
func TestSpanGleichzeitig(t *testing.T) {
	const goroutinen, laeufe = 16, 8
	const runden = 6 * laeufe
	u := leseUmg(t)
	u.c.mc.modell = modell16()
	status := model.AlleCoverageStatus()
	for i := range goroutinen {
		set := "S" + strconv.Itoa(i)
		u.co.setze("b"+strconv.Itoa(i), set, status[i%len(status)])
		es := []model.EntityState{produkt(1, "a"), produkt(2, "b")}
		for j := range es {
			es[j].EntitySet = set
		}
		u.en.setze(set, es...)
	}
	u.ws.ersatz = func(id string) (model.Workspace, error) {
		return model.Workspace{ID: id, ServiceID: dienstID, BaselineSnapshotID: "b" + strings.TrimPrefix(id, "w")}, nil
	}
	u.c.url.f = func(pfad, _ string) (model.ODataURL, error) {
		teile := strings.Split(pfad, "/") // "", Set, Art
		art := map[string]model.RessourcenArt{"c": model.RessourceMenge, "e": model.RessourceEntity, "a": model.RessourceAnzahl}[teile[2]]
		return model.ODataURL{Ressource: model.Ressource{Art: art, EntitySet: teile[1], Schluessel: []model.SchluesselTeil{zahl("", "1")}}}, nil
	}
	// Die Schranke steht an jedem Aufruf eines Ports hinter dem Beginn des Spans: alle 16 Goroutinen führen dieselbe Folge von Anfragen
	// und rufen je Anfrage dieselben Ports, so dass sie in jeder Stufe zugleich ankommen und jeder Wert, den ein Span zwischen zwei
	// Stufen aus einer geteilten Stelle liest, die Werte aller 16 Anfragen sah.
	schranke := neuerZyklus(goroutinen)
	var verpasst sync.Map
	var gescheitert atomic.Bool
	hinterBeginn := []string{"Blobs.Open", "MetadatenCodec.ParseMetadata", "Coverage.Get", "Entities.ListSet", "Entities.Get",
		"Schreiber.SchreibeCollection", "Schreiber.SchreibeEntity", "Schreiber.SchreibeAnzahl"}
	u.p.haken = func(op string) {
		if slices.Contains(hinterBeginn, op) && !gescheitert.Load() && !schranke.warte(10*time.Second) {
			gescheitert.Store(true)
			verpasst.Store("Schranke nicht erreicht", true)
		}
	}
	m := u.mocker()
	var wg sync.WaitGroup
	for i := range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range laeufe {
				for _, methode := range []string{"GET", "HEAD"} {
					for _, art := range []string{"c", "e", "a"} {
						a := anfrage(methode, dienstRoot+"/S"+strconv.Itoa(i)+"/"+art)
						a.Header[headerName] = []string{"w" + strconv.Itoa(i)}
						if antwort, err := m.Mock(u.ctx, a); err != nil {
							verpasst.Store(fmt.Sprintf("Goroutine %d %s %s: %v", i, methode, art, err), true)
						} else {
							_ = antwort.Body.Close()
						}
					}
				}
			}
		}()
	}
	wg.Wait()
	verpasst.Range(func(k, _ any) bool { t.Errorf("%v", k); return true })
	f := u.tel.FakeTelemetry
	if got := len(buecher(f)); got != goroutinen*runden {
		t.Fatalf("%d Spans, erwartet %d", got, goroutinen*runden)
	}
	jeSet := map[string][]spanBuch{}
	for _, b := range buecher(f) {
		set, _ := b.wert(attrSet)
		jeSet[set] = append(jeSet[set], b)
		if b.name != "odata.mock.query" || b.enden != 1 {
			t.Errorf("Span %+v: Name oder Zahl der End", b)
		}
	}
	for i := range goroutinen {
		set := "S" + strconv.Itoa(i)
		gelesen := 0
		for _, b := range jeSet[set] {
			if st, _ := b.wert(attrCov); st == string(status[i%len(status)]) {
				gelesen++
			} else if st != "nicht_gelesen" {
				t.Errorf("Set %s: Status %q", set, st)
			}
		}
		if len(jeSet[set]) != runden || gelesen != 4*laeufe {
			t.Errorf("Set %s: %d Spans, davon %d mit gelesenem Status; erwartet %d und %d", set, len(jeSet[set]), gelesen, runden, 4*laeufe)
		}
	}
	if len(jeSet) != goroutinen || !slices.Equal(status, model.AlleCoverageStatus()) {
		t.Errorf("%d Sets im Buch, erwartet 16", len(jeSet))
	}
}

package mock_test

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// zweig ist ein Request des Tests der Gleichzeitigkeit: der Dienst (Pfad), der Workspace (Dienst und Zustand), die Methode, der
// Modus von STRICT_COVERAGE und der Pfad der Ressource; die URL liest ein Fake nach dem Pfad.
type zweig struct {
	name    string
	dienst  string
	zustand model.CoverageStatus
	methode string
	strikt  bool
	pfad    string
}

// artPfad ist der Pfad der Ressource je Art (der Fake des URL-Lesers kennt sie).
var artPfad = map[model.RessourcenArt]string{model.RessourceMetadaten: "m", model.RessourceMenge: "c", model.RessourceEntity: "e", model.RessourceAnzahl: "a"}

// zweigeM14 sind die 48 Läufe von M14, gebildet aus den Modell-Mengen: Art 4 × Zustand 3 × Methode 2 × Modus 2.
func zweigeM14() []zweig {
	var r []zweig
	for _, art := range model.AlleRessourcenArten() {
		for _, z := range model.AlleCoverageStatus() {
			for _, m := range []string{"GET", "HEAD"} {
				for _, strikt := range []bool{false, true} {
					r = append(r, zweig{fmt.Sprintf("Art %d %s %s strikt=%v", art, z, m, strikt), "svc", z, m, strikt, artPfad[art]})
				}
			}
		}
	}
	return r
}

// zweigeM19 sind die 12 seltenen Zweige von M19: die 8 Codes des Lesepfads, die 3 Arten mit Erfolg (alle außer $metadata) und
// HEAD; jeder Request erreicht seinen Zweig je Aufruf einmal.
func zweigeM19() []zweig {
	c := model.CoverageComplete
	return []zweig{
		{"ODR-E3001", "kaputt", c, "GET", false, "c"}, {"ODR-E6001", "svc", c, "GET", false, "e-filter"}, {"ODR-E6003", "svc", model.CoveragePartial, "GET", true, "c"},
		{"ODR-E6004", "pm", c, "GET", false, "c"}, {"ODR-E6005", "svc", c, "GET", false, "e-key"}, {"ODR-E6008", "svc", c, "GET", false, "c-unbekannt"},
		{"ODR-E7001", "svc", c, "GET", false, "c-sks"}, {"ODR-E9001", "svc", c, "GET", false, "c-pairs"},
		{"Menge", "svc", c, "GET", false, "c"}, {"Entity", "svc", c, "GET", false, "e"}, {"Anzahl", "svc", c, "GET", false, "a"}, {"HEAD", "svc", c, "HEAD", false, "c"},
	}
}

// gleichzeitig baut den Aufbau des Tests: drei Dienste (svc, kaputt ohne Blob, pm mit Fehler des Metadaten-Lesers), drei
// Zustände der Coverage je Workspace (Kennung `<Dienst>|<Zustand>`) und zwei Dienste des Modus (strikt und nicht strikt).
func gleichzeitig(t *testing.T) (*umg, *mock.Mocker, *mock.Mocker) {
	t.Helper()
	u := leseUmg(t)
	pm := neueCodecs(neuesProt(), model.Kopfzeile{Name: "X-Pm", Wert: "p"})
	pm.p.setzeFehler("MetadatenCodec.ParseMetadata", model.ErrMetadatenUngueltig)
	u.c.url.f = urlGleichzeitig
	pm.url.f = urlGleichzeitig
	u.k.Dienste = []mock.Dienst{{ID: dienstID, Root: dienstRoot}, {ID: "kaputt", Root: "/kaputt"}, {ID: "pm", Root: "/pm", Codecs: pm.satz(false, "")}}
	u.md.setze(model.Service{ID: "kaputt", MetadataBlob: "sha256:fehlt"})
	u.md.setze(model.Service{ID: "pm", MetadataBlob: metaRef})
	for _, z := range model.AlleCoverageStatus() {
		for _, set := range []string{setProducts, "Sks", "Pairs", "Nirgends"} {
			u.co.setze("b"+string(z), set, z)
		}
	}
	u.ws.ersatz = func(id string) (model.Workspace, error) {
		dienst, zustand, _ := strings.Cut(id, "|")
		return model.Workspace{ID: id, ServiceID: dienst, BaselineSnapshotID: "b" + zustand}, nil
	}
	u.en.ersatz = func(set string) ([]model.EntityState, error) {
		switch set {
		case "Sks":
			return nil, outbound.ErrNichtGefunden
		case "Pairs":
			return []model.EntityState{eintrag("Pairs", "'x'", `{"A":`)}, nil
		}
		return []model.EntityState{produkt(1, "a"), produkt(2, "b"), produkt(3, "c")}, nil
	}
	nein := u.k
	nein.StrictCoverage = false
	ja := u.k
	ja.StrictCoverage = true
	mn, err := mock.New(nein)
	if err != nil {
		t.Fatal(err)
	}
	mj, err := mock.New(ja)
	if err != nil {
		t.Fatal(err)
	}
	return u, mn, mj
}

// urlGleichzeitig liest die Pfade des Tests: m, c, e, a und die Sonderfälle der seltenen Zweige.
func urlGleichzeitig(pfad, _ string) (model.ODataURL, error) {
	r := func(art model.RessourcenArt, set string, opt model.Optionen, teile ...model.SchluesselTeil) (model.ODataURL, error) {
		return model.ODataURL{Ressource: model.Ressource{Art: art, EntitySet: set, Schluessel: teile}, Optionen: opt}, nil
	}
	switch strings.TrimPrefix(pfad, "/") {
	case "m":
		return r(model.RessourceMetadaten, "", model.Optionen{})
	case "c":
		return r(model.RessourceMenge, setProducts, model.Optionen{})
	case "e":
		return r(model.RessourceEntity, setProducts, model.Optionen{}, zahl("", "1"))
	case "a":
		return r(model.RessourceAnzahl, setProducts, model.Optionen{})
	case "e-filter":
		return r(model.RessourceEntity, setProducts, model.Optionen{Filter: idGroesser("0")}, zahl("", "1"))
	case "e-key":
		return r(model.RessourceEntity, setProducts, model.Optionen{}, zahl("Z", "1"))
	case "c-unbekannt":
		return r(model.RessourceMenge, "Nirgends", model.Optionen{})
	case "c-sks":
		return r(model.RessourceMenge, "Sks", model.Optionen{})
	case "c-pairs":
		return r(model.RessourceMenge, "Pairs", model.Optionen{})
	}
	return model.ODataURL{}, &model.URLFehler{Ursache: model.URLSegment}
}

// TestZweigeGleichzeitig hält L36 (M14 und M19, SPEC-195 Nr. 12): jeder Zweig des Lesens läuft unter dem Detektor von `make
// test-race`: die 48 Läufe von M14 (Art × Zustand × Methode × Modus, aus den Modell-Mengen) und die 12 seltenen Zweige von M19
// (8 Codes des Lesepfads, 3 Arten, HEAD), jeder Request in 4 Runden je Goroutine bei 64 Goroutinen, also 256 Aufrufe je Zweig
// und Prozess; die Antwort ist gleich der des Einzelaufrufs. Zählprüfung gegen model.AlleRessourcenArten(),
// model.AlleCoverageStatus() und die Aufteilung von model.AlleCodes() (L37).
func TestZweigeGleichzeitig(t *testing.T) {
	const goroutinen, runden = 64, 4
	u, mn, mj := gleichzeitig(t)
	m14, m19 := zweigeM14(), zweigeM19()
	if len(m14) != len(model.AlleRessourcenArten())*len(model.AlleCoverageStatus())*2*2 || len(m14) != 48 {
		t.Fatalf("%d Läufe in M14", len(m14))
	}
	erreichbar, _ := codeListen()
	zuordnung := []string{"ODR-E5003", "ODR-E5004", "ODR-E5005", "ODR-E6002", "ODR-E6006"}
	var lesepfad []string
	for _, c := range erreichbar {
		if !slices.Contains(zuordnung, c.String()) {
			lesepfad = append(lesepfad, c.String())
		}
	}
	if len(m19) != len(lesepfad)+(len(model.AlleRessourcenArten())-1)+1 || len(lesepfad) != 8 || len(m19) != 12 {
		t.Fatalf("%d seltene Zweige, %d Codes des Lesepfads", len(m19), len(lesepfad))
	}
	alle := append(slices.Clone(m14), m19...)
	anfrageZu := func(z zweig) (inbound.Anfrage, *mock.Mocker) {
		root := "/" + z.dienst
		if z.dienst == dienstID {
			root = dienstRoot
		}
		a := anfrage(z.methode, root+"/"+z.pfad)
		a.Header[headerName] = []string{z.dienst + "|" + string(z.zustand)}
		if z.strikt {
			return a, mj
		}
		return a, mn
	}
	soll := make([]string, len(alle))
	codes := map[string]bool{}
	for i, z := range alle {
		a, m := anfrageZu(z)
		antwort, err := m.Mock(u.ctx, a)
		soll[i] = ergebnis(antwort, err)
		if i >= len(m14) && strings.HasPrefix(z.name, "ODR-") {
			if fehlercode(err).String() != z.name {
				t.Errorf("%s: Code %s", z.name, fehlercode(err))
			}
			codes[z.name] = true
		}
	}
	for _, c := range lesepfad {
		if !codes[c] {
			t.Errorf("%s hat keinen seltenen Zweig", c)
		}
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var abweichungen []string
	start := make(chan struct{})
	for g := 0; g < goroutinen; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for r := 0; r < runden; r++ {
				for i, z := range alle {
					a, m := anfrageZu(z)
					antwort, err := m.Mock(u.ctx, a)
					if got := ergebnis(antwort, err); got != soll[i] {
						mu.Lock()
						abweichungen = append(abweichungen, fmt.Sprintf("g%d %s: %q, erwartet %q", g, z.name, got, soll[i]))
						mu.Unlock()
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	if len(abweichungen) != 0 {
		t.Errorf("%d Abweichungen, erste: %s", len(abweichungen), abweichungen[0])
	}
}

// TestLesesicht hält L38 (SPEC-196 Nr. 10): wechselt der Zustand eines Sets atomar zwischen A (40 Einträge) und B (40 andere),
// ist jede Antwort der Collection und der Anzahl gleich A oder B vollständig, nie gemischt, und ListSet wird je Antwort genau
// einmal gerufen. 8 Leser × 200 Antworten, ein Schreiber, der den Zustand wechselt.
func TestLesesicht(t *testing.T) {
	const leser, antworten = 8, 200
	u := leseUmg(t)
	u.c.sch.zeigeIds = true
	var a, b []model.EntityState
	var idsA, idsB []string
	for i := 1; i <= 40; i++ {
		a = append(a, produkt(i, "a"))
		b = append(b, produkt(100+i, "b"))
		idsA = append(idsA, fmt.Sprint(i))
		idsB = append(idsB, fmt.Sprint(100+i))
	}
	sollA, sollB := strings.Join(idsA, ",")+",", strings.Join(idsB, ",")+","
	var phase atomic.Int32
	u.en.ersatz = func(string) ([]model.EntityState, error) {
		if phase.Load() == 0 {
			return a, nil
		}
		return b, nil
	}
	u.c.url.f = func(pfad, _ string) (model.ODataURL, error) {
		if strings.HasSuffix(pfad, "/a") {
			return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceAnzahl, EntitySet: setProducts}}, nil
		}
		return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: setProducts}}, nil
	}
	m := u.mocker()
	var stop atomic.Bool
	var schreiber sync.WaitGroup
	schreiber.Add(1)
	go func() {
		defer schreiber.Done()
		for !stop.Load() {
			phase.Store(1 - phase.Load())
		}
	}()
	var wg sync.WaitGroup
	var mu sync.Mutex
	var fehler []string
	for l := 0; l < leser; l++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < antworten; i++ {
				req := anfrage("GET", dienstRoot+"/c")
				if i%2 == 1 {
					req.Pfad = dienstRoot + "/a"
				}
				antwort, err := m.Mock(u.ctx, req)
				mu.Lock()
				if err != nil {
					fehler = append(fehler, err.Error())
				} else if body := string(lies(t, antwort)); i%2 == 0 {
					_, rest, _ := strings.Cut(body, "|")
					if rest+"," != sollA && rest+"," != sollB {
						fehler = append(fehler, "gemischt: "+rest)
					}
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	stop.Store(true)
	schreiber.Wait()
	if len(fehler) != 0 {
		t.Errorf("%d fehlerhafte Antworten, erste: %s", len(fehler), fehler[0])
	}
	if n := u.p.anzahl("Entities.ListSet"); n != leser*antworten {
		t.Errorf("%d Aufrufe von ListSet für %d Antworten", n, leser*antworten)
	}
}

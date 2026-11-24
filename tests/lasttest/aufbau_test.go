package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// kleinesProfil ist das Profil des Tests: zwei Sets zu 20 Entities in Seiten zu 10, 12 Lookups; Exchanges 1 + 4 + 12 gleich 17.
func kleinesProfil() Profil {
	return Profil{
		Name: "klein", Exchanges: 17, Sets: 2, Entities: 20, SeitenGroesse: 10, Lookups: 12, Rate: 20,
		Warmup: 3 * time.Second, Messphase: 10 * time.Second, Laeufe: 1, FristAufbau: 5 * time.Minute, Gesamtfrist: 15 * time.Minute,
	}
}

// gebaut zählt die Aufrufe von Herunterfahren eines Handle des Bauers.
type gebaut struct {
	herunter atomic.Int32
}

// sessionsErsatz ersetzt List der Sessions des Handle.
type sessionsErsatz struct {
	inbound.SessionUseCase
	liste func() ([]model.Session, error)
}

func (s sessionsErsatz) List(context.Context) ([]model.Session, error) { return s.liste() }

// workspacesErsatz lässt Create mit einem Fehler enden.
type workspacesErsatz struct {
	inbound.WorkspaceUseCase
	fehler error
}

func (w workspacesErsatz) Create(context.Context, string, string) (model.Workspace, error) {
	return model.Workspace{}, w.fehler
}

// bauer baut die Handles des Aufbaus über baueEcht, zählt je Handle die Aufrufe von Herunterfahren (der Aufruf geht unverändert an das
// Handle) und merkt sich die Zähler; baueFehler gibt für den n-ten Aufruf (ab 0) einen Fehler vor, ausruesten ersetzt Funktionen des n-ten Handle.
type bauer struct {
	mu         sync.Mutex
	aufrufe    int
	handles    []*gebaut
	baueFehler func(n int) error
	ausruesten func(n int, h *handle)
}

func (b *bauer) baue(ctx context.Context, k bootstrap.Konfiguration) (*handle, error) {
	b.mu.Lock()
	n := b.aufrufe
	b.aufrufe++
	b.mu.Unlock()
	if b.baueFehler != nil {
		if err := b.baueFehler(n); err != nil {
			return nil, err
		}
	}
	h, err := baueEcht(ctx, k)
	if err != nil {
		return nil, err
	}
	g, echt := &gebaut{}, h.Herunterfahren
	h.Herunterfahren = func(ctx context.Context) error {
		g.herunter.Add(1)
		return echt(ctx)
	}
	if b.ausruesten != nil {
		b.ausruesten(n, h)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handles = append(b.handles, g)
	return h, nil
}

// aendere ersetzt den Handler des Referenzdienstes durch f, das den ursprünglichen Handler als inner bekommt.
func aendere(f func(w http.ResponseWriter, r *http.Request, inner http.Handler)) func(http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f(w, r, inner) })
	}
}

// umgeschrieben ruft inner und gibt Status und Header weiter, den Body nach f.
func umgeschrieben(w http.ResponseWriter, r *http.Request, inner http.Handler, f func(body string) string) {
	rec := httptest.NewRecorder()
	inner.ServeHTTP(rec, r)
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.Header().Del("Content-Length")
	w.WriteHeader(rec.Code)
	_, _ = w.Write([]byte(f(rec.Body.String())))
}

// ablageLesen liest die Ablage unabhängig vom Aufbau: Status der Session, Zahl der Exchanges, Coverage und Entities je Set.
func ablageLesen(t *testing.T, dir string, a aufbauErgebnis, sets int) (status model.SessionStatus, exchanges int, coverage []string, entities []int) {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s, err := db.Sessions().Get(ctx, a.Session)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := db.Recording().ListBySession(ctx, a.Session)
	if err != nil {
		t.Fatal(err)
	}
	cov, err := db.Coverage().List(ctx, a.Baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cov {
		coverage = append(coverage, string(c.Status))
	}
	for set := 1; set <= sets; set++ {
		es, err := db.Entities().ListSet(ctx, a.Workspace, fmt.Sprintf("Set%d", set))
		if err != nil {
			t.Fatal(err)
		}
		entities = append(entities, len(es))
	}
	return s.Status, len(ex), coverage, entities
}

// TestAufbau: der Aufbau liefert genau so viele Exchanges, Entities und eine Baseline complete, wie das Profil sagt; ein Status ungleich 200,
// ein falscher MetadataHash, eine Seite weniger, eine Baseline partial und eine Entity weniger sind je ein Fehler des Aufbaus mit dem Grund
// aufbau an ihrer Stelle und ohne Workspace (SPEC-216 Nr. 6 e, 6 f).
func TestAufbau(t *testing.T) {
	wache(t, 120*time.Second)
	nextLink := regexp.MustCompile(`"@odata.nextLink":"[^"]*",`)
	// fehlt ist die erste Entity von Set1, die kein Lookup holt und die nicht die letzte ihrer Seite ist: ohne sie bleibt die Kette der Seiten
	// belegt (die Coverage ist complete), und der Workspace trägt eine Entity weniger.
	geholt := map[int]bool{}
	for _, s := range paare(kleinesProfil(), 1)[:kleinesProfil().Lookups] {
		if s[0] == 1 {
			geholt[s[1]] = true
		}
	}
	fehlt := 0
	for id := 1; id < 20 && fehlt == 0; id++ {
		if id%10 != 0 && !geholt[id] {
			fehlt = id
		}
	}
	fehlend := regexp.MustCompile(fmt.Sprintf(`\{"Id":%d,"Name":"Set1-%d",[^}]*\},?`, fehlt, fehlt))
	faelle := []struct {
		name    string
		aendern func(http.Handler) http.Handler
		stelle  string
		ursache string
	}{
		{"Gutfall", nil, "", ""},
		{"Seite 2 liefert 500", aendere(func(w http.ResponseWriter, r *http.Request, inner http.Handler) {
			if r.URL.Path == "/odata/Set1" && r.URL.RawQuery == "$skiptoken=10" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			inner.ServeHTTP(w, r)
		}), stelleStatus, "Status 500"},
		{"eine Seite weniger", aendere(func(w http.ResponseWriter, r *http.Request, inner http.Handler) {
			if r.URL.Path == "/odata/Set2" && r.URL.RawQuery == "" {
				umgeschrieben(w, r, inner, func(b string) string { return nextLink.ReplaceAllString(b, "") })
				return
			}
			inner.ServeHTTP(w, r)
		}), stelleZahl, "Exchanges 16, erwartet 17"},
		{"falscher MetadataHash", aendere(func(w http.ResponseWriter, r *http.Request, inner http.Handler) {
			if r.URL.Path == "/odata/$metadata" {
				umgeschrieben(w, r, inner, func(b string) string { return b + "\n" })
				return
			}
			inner.ServeHTTP(w, r)
		}), stelleImport, "ODR-E6004"},
		{"Baseline partial", aendere(func(w http.ResponseWriter, r *http.Request, inner http.Handler) {
			if r.URL.Path == "/odata/Set1" && r.URL.RawQuery == "" {
				umgeschrieben(w, r, inner, func(b string) string {
					return strings.Replace(b, `"@odata.nextLink":"http://127.0.0.1:`, `"@odata.nextLink":"http://127.0.0.2:`, 1)
				})
				return
			}
			inner.ServeHTTP(w, r)
		}), stelleCoverage, "Coverage nicht complete"},
		{"eine Entity weniger in der Seite", aendere(func(w http.ResponseWriter, r *http.Request, inner http.Handler) {
			if r.URL.Path == "/odata/Set1" {
				umgeschrieben(w, r, inner, func(b string) string { return fehlend.ReplaceAllString(b, "") })
				return
			}
			inner.ServeHTTP(w, r)
		}), stelleEntities, "Set1: Entities 19, erwartet 20"},
	}
	if len(faelle) != 6 {
		t.Fatalf("%d Fälle, erwartet 6", len(faelle))
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			dir := t.TempDir()
			e := aufbauEingabe{Profil: kleinesProfil(), Startwert: 1, Uhr: systemuhr{}, DataDir: dir, Umhuellen: f.aendern}
			erg, err := e.aufbauen(context.Background())
			if f.stelle != "" {
				var af *aufbauFehler
				if !errors.As(err, &af) || af.grund != grundAufbau || af.stelle != f.stelle || erg != (aufbauErgebnis{}) {
					t.Fatalf("Fehler %v, Ergebnis %+v, erwartet Grund aufbau an der Stelle %s ohne Ergebnis", err, erg, f.stelle)
				}
				if af.Error() != "Aufbau gescheitert: aufbau" || !strings.Contains(af.ursache.Error(), f.ursache) {
					t.Errorf("Meldung %q, Ursache %q, erwartet die Ursache %q", af.Error(), af.ursache, f.ursache)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			status, exchanges, coverage, entities := ablageLesen(t, dir, erg, 2)
			if status != model.SessionComplete || exchanges != 17 || fmt.Sprint(coverage) != "[complete complete]" || fmt.Sprint(entities) != "[20 20]" {
				t.Errorf("Ablage: Status %s, Exchanges %d, Coverage %v, Entities %v, erwartet complete, 17, [complete complete], [20 20]",
					status, exchanges, coverage, entities)
			}
			if erg.Exchanges != 17 || erg.Entities != 40 || erg.Session == "" || erg.Baseline == "" || erg.Workspace == "" {
				t.Errorf("Ergebnis %+v, erwartet 17 Exchanges und 40 Entities", erg)
			}
		})
	}
}

// TestAufbauDatenverzeichnis: DATA_DIR darf fehlen oder leer sein; ein Verzeichnis mit einer Datei und eine Datei an seiner Stelle sind
// `DATA_DIR nicht leer`, und der Aufbau lässt die Daten unberührt (SPEC-216 Nr. 6 e).
func TestAufbauDatenverzeichnis(t *testing.T) {
	wache(t, 120*time.Second)
	inhalt := []byte("Inhalt der Datei\x00\xff")
	faelle := []struct {
		name  string
		lege  func(t *testing.T) (dir string, datei string)
		fehlt bool
	}{
		{"fehlt", func(t *testing.T) (string, string) { return filepath.Join(t.TempDir(), "neu"), "" }, false},
		{"leer", func(t *testing.T) (string, string) { return t.TempDir(), "" }, false},
		{"mit einer Datei", func(t *testing.T) (string, string) {
			dir := t.TempDir()
			return dir, filepath.Join(dir, "alt.dat")
		}, true},
		{"ist eine Datei", func(t *testing.T) (string, string) {
			d := filepath.Join(t.TempDir(), "datei")
			return d, d
		}, true},
	}
	if len(faelle) != 4 {
		t.Fatalf("%d Fälle, erwartet 4", len(faelle))
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			dir, datei := f.lege(t)
			if datei != "" {
				if err := os.WriteFile(datei, inhalt, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			e := aufbauEingabe{Profil: kleinesProfil(), Startwert: 1, Uhr: systemuhr{}, DataDir: dir}
			erg, err := e.aufbauen(context.Background())
			if !f.fehlt {
				if err != nil || erg.Workspace == "" {
					t.Fatalf("Aufbau: %v, %+v", err, erg)
				}
				return
			}
			var af *aufbauFehler
			if !errors.As(err, &af) || af.stelle != stelleDaten || af.Error() != "Aufbau: DATA_DIR nicht leer" || erg != (aufbauErgebnis{}) {
				t.Fatalf("Fehler %v, Ergebnis %+v, erwartet DATA_DIR nicht leer", err, erg)
			}
			if got, rerr := os.ReadFile(datei); rerr != nil || string(got) != string(inhalt) {
				t.Errorf("Datei nach dem Fehler: %q, %v", got, rerr)
			}
			if fi, _ := os.Stat(dir); fi.IsDir() {
				es, _ := os.ReadDir(dir)
				if len(es) != 1 {
					t.Errorf("DATA_DIR trägt %d Einträge, erwartet die eine Datei", len(es))
				}
			}
		})
	}
}

// halte ersetzt den Referenzdienst so, dass er die zweite Seite des ersten Sets hält, bis frei geschlossen ist oder die Anfrage endet;
// gehalten meldet, dass sie gehalten wird.
func halte(gehalten chan<- struct{}, frei <-chan struct{}) func(http.Handler) http.Handler {
	return aendere(func(w http.ResponseWriter, r *http.Request, inner http.Handler) {
		if r.URL.Path == "/odata/Set1" && r.URL.RawQuery == "$skiptoken=10" {
			gehalten <- struct{}{}
			select {
			case <-frei:
			case <-r.Context().Done():
				return
			}
		}
		inner.ServeHTTP(w, r)
	})
}

// TestAufbauFrist: die Fake-Uhr geht über die Frist des Profils, während der Referenzdienst eine Antwort hält: der Aufbau endet mit dem Grund
// frist, und sein Handle ist genau einmal heruntergefahren. Eine Minute unter der Frist endet er nicht. Beide Profile tragen ihre Frist
// (30 min und 5 min), keines die des anderen.
func TestAufbauFrist(t *testing.T) {
	wache(t, 180*time.Second)
	for _, p := range []Profil{abnahme(), rauch()} {
		for _, ueber := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s ueber=%v", p.Name, ueber), func(t *testing.T) {
				k := kleinesProfil()
				k.FristAufbau = p.FristAufbau
				uhr := neueFakeUhr()
				gehalten, frei := make(chan struct{}, 1), make(chan struct{})
				var einmal sync.Once
				freigeben := func() { einmal.Do(func() { close(frei) }) }
				t.Cleanup(freigeben)
				b := &bauer{}
				e := aufbauEingabe{Profil: k, Startwert: 1, Uhr: uhr, DataDir: t.TempDir(), Baue: b.baue, Umhuellen: halte(gehalten, frei)}
				ergebnis := make(chan error, 1)
				go func() {
					_, err := e.aufbauen(context.Background())
					ergebnis <- err
				}()
				warteAuf(t, gehalten, 15*time.Second)
				if ueber {
					uhr.Vorstellen(p.FristAufbau + time.Minute)
				} else {
					uhr.Vorstellen(p.FristAufbau - time.Minute)
					freigeben()
				}
				err := warteAuf(t, ergebnis, 15*time.Second)
				var af *aufbauFehler
				if ueber {
					if !errors.As(err, &af) || af.grund != grundFrist || err.Error() != "Aufbau gescheitert: frist" {
						t.Fatalf("Fehler %v, erwartet Grund frist", err)
					}
					if len(b.handles) != 1 || b.handles[0].herunter.Load() != 1 {
						t.Errorf("Handles %d, Herunterfahren %d, erwartet 1 und 1", len(b.handles), b.handles[0].herunter.Load())
					}
					return
				}
				if err != nil {
					t.Fatalf("Aufbau eine Minute unter der Frist: %v", err)
				}
				for i, h := range b.handles {
					if h.herunter.Load() != 1 {
						t.Errorf("Handle %d: Herunterfahren %d, erwartet 1", i, h.herunter.Load())
					}
				}
			})
		}
	}
}

// TestAufbauAbbruch: ein Kontext, der vor dem Start oder während einer gehaltenen Antwort endet, ist ein Fehler des Aufbaus mit dem Kontextfehler
// als Ursache und ohne den Grund frist; jedes gebaute Handle ist genau einmal heruntergefahren (SPEC-216 Nr. 4 d).
func TestAufbauAbbruch(t *testing.T) {
	wache(t, 120*time.Second)
	for _, nach := range []string{"vor dem Start", "bei gehaltener Antwort"} {
		t.Run(nach, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gehalten, frei := make(chan struct{}, 1), make(chan struct{})
			t.Cleanup(func() { close(frei) })
			b := &bauer{}
			e := aufbauEingabe{Profil: kleinesProfil(), Startwert: 1, Uhr: neueFakeUhr(), DataDir: t.TempDir(), Baue: b.baue, Umhuellen: halte(gehalten, frei)}
			if nach == "vor dem Start" {
				cancel()
			} else {
				go func() {
					<-gehalten
					cancel()
				}()
			}
			erg, err := e.aufbauen(ctx)
			var af *aufbauFehler
			if !errors.As(err, &af) || af.grund != grundAufbau || !errors.Is(err, context.Canceled) || erg != (aufbauErgebnis{}) {
				t.Fatalf("Fehler %v, Ergebnis %+v, erwartet Grund aufbau mit der Ursache context.Canceled", err, erg)
			}
			for i, h := range b.handles {
				if h.herunter.Load() != 1 {
					t.Errorf("Handle %d: Herunterfahren %d, erwartet 1", i, h.herunter.Load())
				}
			}
		})
	}
}

// TestAufbauZweimal: ein zweiter Aufbau auf denselben Daten ist `DATA_DIR nicht leer` und lässt die Daten des ersten unberührt (SPEC-216 Nr. 6 e).
func TestAufbauZweimal(t *testing.T) {
	wache(t, 120*time.Second)
	dir := t.TempDir()
	e := aufbauEingabe{Profil: kleinesProfil(), Startwert: 1, Uhr: systemuhr{}, DataDir: dir}
	if _, err := e.aufbauen(context.Background()); err != nil {
		t.Fatal(err)
	}
	vorher, _ := os.ReadDir(dir)
	erg, err := e.aufbauen(context.Background())
	var af *aufbauFehler
	if !errors.As(err, &af) || af.stelle != stelleDaten || erg != (aufbauErgebnis{}) {
		t.Fatalf("zweiter Aufbau: %v, %+v", err, erg)
	}
	if nachher, _ := os.ReadDir(dir); len(vorher) == 0 || len(nachher) != len(vorher) {
		t.Errorf("Einträge vorher %d, nachher %d", len(vorher), len(nachher))
	}
}

// TestAufbauRauch: der Aufbau des Profils rauch trägt echte Größen (Nutzlast bis 60 KiB, 5 Sets zu 60 Entities): 300 Exchanges, 300 Entities,
// Coverage complete je Set, die Exchanges der Aufnahme sind genau die Menge von aufbauAnfragen mit dem Header Accept (Orakel: die Ziele
// aus erwarteteMenge), und die Listen replay und mock treffen in einem Handle der Betriebsart (listenTreffen).
func TestAufbauRauch(t *testing.T) {
	wache(t, 180*time.Second)
	dir := t.TempDir()
	e := aufbauEingabe{Profil: rauch(), Startwert: 1, Uhr: systemuhr{}, DataDir: dir}
	erg, err := e.aufbauen(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	status, exchanges, coverage, entities := ablageLesen(t, dir, erg, 5)
	if status != model.SessionComplete || exchanges != 300 || fmt.Sprint(coverage) != "[complete complete complete complete complete]" ||
		fmt.Sprint(entities) != "[60 60 60 60 60]" || erg.Exchanges != 300 || erg.Entities != 300 {
		t.Errorf("Status %s, Exchanges %d, Coverage %v, Entities %v, Ergebnis %+v", status, exchanges, coverage, entities, erg)
	}
	db, err := sqlite.Open(context.Background(), filepath.Join(dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	ex, err := db.Recording().ListBySession(context.Background(), erg.Session)
	if err != nil {
		t.Fatal(err)
	}
	menge := erwarteteMenge(rauch(), 1)
	for _, x := range ex {
		ziel := wurzel + x.Path
		if x.RawQuery != "" {
			ziel += "?" + x.RawQuery
		}
		menge[ziel]--
		if x.Method != "GET" || fmt.Sprint(x.RequestHeaders["Accept"]) != "[application/json]" {
			t.Errorf("Exchange %d: %s, Accept %v", x.SequenceNo, x.Method, x.RequestHeaders["Accept"])
		}
	}
	for ziel, n := range menge {
		if n != 0 {
			t.Errorf("%s: %d Exchanges zu viel oder zu wenig gegenüber der Liste", ziel, n)
		}
	}
	listenTreffen(t, dir, erg)
}

// fehlerlauf fährt den Aufbau des kleinen Profils mit dem Bauer und liefert die Stelle und die Ursache des Fehlers.
func fehlerlauf(t *testing.T, b *bauer, dir string, umhuellen func(http.Handler) http.Handler) (stelle string, ursache error) {
	t.Helper()
	e := aufbauEingabe{Profil: kleinesProfil(), Startwert: 1, Uhr: systemuhr{}, DataDir: dir, Baue: b.baue, Umhuellen: umhuellen}
	erg, err := e.aufbauen(context.Background())
	var af *aufbauFehler
	if !errors.As(err, &af) || af.grund != grundAufbau || erg != (aufbauErgebnis{}) {
		t.Fatalf("Fehler %v, Ergebnis %+v, erwartet Grund aufbau ohne Ergebnis", err, erg)
	}
	for i, h := range b.handles {
		if h.herunter.Load() != 1 {
			t.Errorf("Handle %d: Herunterfahren %d, erwartet 1", i, h.herunter.Load())
		}
	}
	return af.stelle, af.ursache
}

// TestAufbauFehlerJeSchritt: ein Fehler beim Bauen, Hochfahren und Lesen der Sessions, beim Bauen des zweiten Handle und beim Anlegen des
// Workspace ist je ein Fehler des Aufbaus an der Stelle des Schritts und mit dem Fehler des Schritts als Ursache (zwei Sessions statt einer hat
// keinen), ohne Ergebnis; jedes gebaute Handle ist genau einmal heruntergefahren.
func TestAufbauFehlerJeSchritt(t *testing.T) {
	wache(t, 180*time.Second)
	boese := errors.New("Fehler des Ersatzes")
	zwei := func() ([]model.Session, error) { return []model.Session{{ID: "a"}, {ID: "b"}}, nil }
	nichts := func() ([]model.Session, error) { return nil, boese }
	faelle := []struct {
		name    string
		b       *bauer
		stelle  string
		ursache bool
	}{
		{"Bauen des ersten Handle", &bauer{baueFehler: func(n int) error {
			if n == 0 {
				return boese
			}
			return nil
		}}, stelleAufnahme, true},
		{"Hochfahren", &bauer{ausruesten: func(n int, h *handle) {
			h.Hochfahren = func(context.Context) error { return boese }
		}}, stelleAufnahme, true},
		{"Sessions lesen", &bauer{ausruesten: func(n int, h *handle) {
			echt := h.Sessions
			h.Sessions = func() inbound.SessionUseCase { return sessionsErsatz{SessionUseCase: echt(), liste: nichts} }
		}}, stelleSession, true},
		{"zwei Sessions", &bauer{ausruesten: func(n int, h *handle) {
			echt := h.Sessions
			h.Sessions = func() inbound.SessionUseCase { return sessionsErsatz{SessionUseCase: echt(), liste: zwei} }
		}}, stelleSession, false},
		{"Bauen des zweiten Handle", &bauer{baueFehler: func(n int) error {
			if n == 1 {
				return boese
			}
			return nil
		}}, stelleImport, true},
		{"Workspace anlegen", &bauer{ausruesten: func(n int, h *handle) {
			if n == 1 {
				echt := h.Workspaces
				h.Workspaces = func() inbound.WorkspaceUseCase { return workspacesErsatz{WorkspaceUseCase: echt(), fehler: boese} }
			}
		}}, stelleImport, true},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			stelle, ursache := fehlerlauf(t, f.b, t.TempDir(), nil)
			if stelle != f.stelle || errors.Is(ursache, boese) != f.ursache {
				t.Errorf("Stelle %s, Ursache %v, erwartet %s und den Fehler des Schritts als Ursache: %v", stelle, ursache, f.stelle, f.ursache)
			}
		})
	}
}

// TestAufbauSeiteUnlesbar: eine Seite, die kein JSON ist, und ein `@odata.nextLink`, der keine Adresse ist, sind je ein Fehler an der Stelle
// der Aufnahme.
func TestAufbauSeiteUnlesbar(t *testing.T) {
	wache(t, 120*time.Second)
	for name, body := range map[string]string{
		"kein JSON":      "keine Seite",
		"Link ohne Ziel": `{"@odata.nextLink":"http://[::1","value":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			umhuellen := aendere(func(w http.ResponseWriter, r *http.Request, inner http.Handler) {
				if r.URL.Path == "/odata/Set1" && r.URL.RawQuery == "" {
					w.Header().Set("Content-Type", jsonTyp)
					_, _ = w.Write([]byte(body))
					return
				}
				inner.ServeHTTP(w, r)
			})
			if got, _ := fehlerlauf(t, &bauer{}, t.TempDir(), umhuellen); got != stelleAufnahme {
				t.Errorf("Stelle %s, erwartet %s", got, stelleAufnahme)
			}
		})
	}
}

// TestAufbauDienstNichtStartbar: ein DATA_DIR, das die Konfiguration nicht anlegen kann, und eine Adresse des Referenzdienstes, die nicht bindet,
// sind je ein Fehler des Aufbaus an der Stelle dienst, ohne Ergebnis und ohne ein gebautes Handle.
func TestAufbauDienstNichtStartbar(t *testing.T) {
	wache(t, 60*time.Second)
	for name, e := range map[string]aufbauEingabe{
		"DATA_DIR nicht anlegbar": {DataDir: "/proc/lasttest/daten"},
		"Adresse nicht bindbar":   {DataDir: t.TempDir(), Adresse: "127.0.0.1:99999"},
	} {
		t.Run(name, func(t *testing.T) {
			b := &bauer{}
			e.Profil, e.Startwert, e.Uhr, e.Baue = kleinesProfil(), 1, systemuhr{}, b.baue
			erg, err := e.aufbauen(context.Background())
			var af *aufbauFehler
			if !errors.As(err, &af) || af.stelle != stelleDienst || af.grund != grundAufbau || erg != (aufbauErgebnis{}) || len(b.handles) != 0 {
				t.Errorf("Fehler %v, Ergebnis %+v, Handles %d, erwartet die Stelle dienst", err, erg, len(b.handles))
			}
		})
	}
}

// TestPruefe: die Zählprüfung nach dem Aufbau meldet je Abweichung die Stelle: eine unbekannte und eine offene Session, eine Ablage ohne
// Datenbank, eine andere Zahl der Exchanges, mehr Sets in der Coverage als die Baseline trägt und mehr Entities je Set als der Workspace; die
// Ablage des Gutfalls besteht.
func TestPruefe(t *testing.T) {
	wache(t, 120*time.Second)
	dir := t.TempDir()
	e := aufbauEingabe{Profil: kleinesProfil(), Startwert: 1, Uhr: systemuhr{}, DataDir: dir}
	erg, err := e.aufbauen(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	k, err := bootstrap.LadeKonfiguration([]string{
		"MODE=record", "LISTEN=127.0.0.1:8080", "DATA_DIR=" + dir, "SERVICE_ID=" + dienstID, "SERVICE_ROOT_PATH=" + wurzel, "UPSTREAM_URL=http://127.0.0.1:9" + wurzel,
	})
	if err != nil {
		t.Fatal(err)
	}
	h, err := baueEcht(context.Background(), k)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Sessions().Create(context.Background(), dienstID, "offen"); err != nil {
		t.Fatal(err)
	}
	if err := h.Herunterfahren(context.Background()); err != nil {
		t.Fatal(err)
	}
	faelle := []struct {
		name   string
		aendre func(e *aufbauEingabe, a *aufbauErgebnis)
		stelle string
	}{
		{"Gutfall", func(*aufbauEingabe, *aufbauErgebnis) {}, ""},
		{"unbekannte Session", func(_ *aufbauEingabe, a *aufbauErgebnis) { a.Session = "gibt-es-nicht" }, stelleSession},
		{"Session offen", func(_ *aufbauEingabe, a *aufbauErgebnis) { a.Session = "offen" }, stelleSession},
		{"Ablage ohne Datenbank", func(e *aufbauEingabe, _ *aufbauErgebnis) { e.DataDir = filepath.Join(dir, "nichts", "da") }, stelleSession},
		{"Exchanges", func(e *aufbauEingabe, _ *aufbauErgebnis) { e.Profil.Exchanges++ }, stelleZahl},
		{"Sets der Coverage", func(e *aufbauEingabe, _ *aufbauErgebnis) { e.Profil.Sets++ }, stelleCoverage},
		{"Entities je Set", func(e *aufbauEingabe, _ *aufbauErgebnis) { e.Profil.Entities++ }, stelleEntities},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			e, a := e, erg
			f.aendre(&e, &a)
			exchanges, entities, err := e.pruefe(context.Background(), a)
			var af *aufbauFehler
			if f.stelle == "" {
				if err != nil || exchanges != 17 || entities != 40 {
					t.Errorf("Gutfall: %d, %d, %v", exchanges, entities, err)
				}
				return
			}
			if !errors.As(err, &af) || af.stelle != f.stelle || exchanges != 0 || entities != 0 {
				t.Errorf("Fehler %v, Zahlen %d und %d, erwartet die Stelle %s", err, exchanges, entities, f.stelle)
			}
		})
	}
}

// TestBaueEcht: ein Fehler von bootstrap.Build liefert kein Handle.
func TestBaueEcht(t *testing.T) {
	wache(t, 30*time.Second)
	if h, err := baueEcht(context.Background(), bootstrap.Konfiguration{}); err == nil || h != nil {
		t.Errorf("Handle %v, Fehler %v, erwartet nil und einen Fehler", h, err)
	}
}

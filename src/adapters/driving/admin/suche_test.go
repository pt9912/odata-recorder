package admin_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// specNach liefert die Zeile der Tabelle von SPEC-109 Nr. 2 und den Code des Modells zum Text des Codes.
func specNach(t *testing.T, code string) (specZeile, model.Code) {
	t.Helper()
	for i, z := range specCodes() {
		if z.code == code {
			return z, modellCode(i)
		}
	}
	t.Fatalf("Code %s steht nicht in der Tabelle", code)
	return specZeile{}, 0
}

// fehlerMit liefert einen gültigen *inbound.Fehler mit dem Code (Status 404, den jeder der hier genutzten Codes trägt).
func fehlerMit(t *testing.T, code string) *inbound.Fehler {
	t.Helper()
	_, c := specNach(t, code)
	return &inbound.Fehler{Code: c, Status: 404}
}

// bodyZu liefert das Fehlerobjekt des Codes als Literal des Vertrags.
func bodyZu(t *testing.T, code string) string {
	t.Helper()
	z, c := specNach(t, code)
	return fehlerBody(z, c)
}

// rufeMitPanik sendet den Request an den Server und liefert die Antwort, den Wert einer Panik, die ServeHTTP verlässt,
// und ob es eine gab.
func (u *umgebung) rufeMitPanik(r *http.Request) (rec *httptest.ResponseRecorder, wert any, panikt bool) {
	rec = httptest.NewRecorder()
	defer func() {
		if p := recover(); p != nil {
			wert, panikt = p, true
		}
	}()
	u.srv.ServeHTTP(rec, r)
	return rec, nil, false
}

// methodenFehler ist ein Fehler mit den Methoden Is und As: Is(ziel) ist wahr genau für istZiel; As setzt bei einem
// Ziel `**inbound.Fehler` den Wert alsFehler (nil mit alsNil: wahr mit einem nil-Zeiger) und ist sonst falsch.
type methodenFehler struct {
	istZiel   error
	alsFehler *inbound.Fehler
	alsNil    bool
}

func (*methodenFehler) Error() string { return "Fehler mit Methoden Is und As" }

func (e *methodenFehler) Is(ziel error) bool { return e.istZiel != nil && ziel == e.istZiel }

func (e *methodenFehler) As(ziel any) bool {
	p, ok := ziel.(**inbound.Fehler)
	switch {
	case !ok:
		return false
	case e.alsNil:
		*p = nil
		return true
	case e.alsFehler != nil:
		*p = e.alsFehler
		return true
	}
	return false
}

// TestIsUndAsMethoden hält F-1 (SPEC-238 Nr. 6 a, „folgt errors.Is und errors.As“): die Suche ruft die Methode Is eines
// Fehlers für die Ziele ErrUngueltig, Canceled und DeadlineExceeded und die Methode As für das Ziel *inbound.Fehler,
// in jeder Position (Menge M-Verpackung: Wurzel, hinter %w einfach und dreifach, Ast eines Join vorn und hinten);
// 9 Formen von Fehlern × 5 Verpackungen = 45 Fälle.
func TestIsUndAsMethoden(t *testing.T) {
	e7003 := fehlerMit(t, "ODR-E7003")
	ungueltig := &inbound.Fehler{Code: 0, Status: 500}
	const (
		sagt400   = "400"
		sagt404   = "404"
		sagt500   = "500"
		sagtAbruf = "Abbruch"
	)
	formen := []struct {
		name    string
		err     error
		erwarte string
	}{
		{"Is(ErrUngueltig)", &methodenFehler{istZiel: model.ErrUngueltig}, sagt400},
		{"Is(Canceled)", &methodenFehler{istZiel: context.Canceled}, sagtAbruf},
		{"Is(DeadlineExceeded)", &methodenFehler{istZiel: context.DeadlineExceeded}, sagtAbruf},
		{"As(*inbound.Fehler) gültig", &methodenFehler{alsFehler: e7003}, sagt404},
		{"As(*inbound.Fehler) ungültig", &methodenFehler{alsFehler: ungueltig}, sagt500},
		{"As wahr mit nil-Zeiger", &methodenFehler{alsNil: true}, sagt500},
		{"As wahr mit nil-Zeiger und Is(ErrUngueltig)", &methodenFehler{alsNil: true, istZiel: model.ErrUngueltig}, sagt400},
		{"Is(anderes Ziel)", &methodenFehler{istZiel: errors.ErrUnsupported}, sagt500},
		{"Is(ErrUngueltig) und As(*inbound.Fehler)", &methodenFehler{istZiel: model.ErrUngueltig, alsFehler: e7003}, sagt404},
	}
	faelle := 0
	for _, f := range formen {
		for _, v := range verpackungen() {
			faelle++
			name := f.name + ", " + v.name
			switch f.erwarte {
			case sagtAbruf:
				u, route := einRoute(t)
				route.err = v.packe(f.err)
				rec, wert, panikt := u.rufeMitPanik(anfrageMit("GET", "/r", nil, nil))
				if !panikt || wert != http.ErrAbortHandler || rec.Body.Len() != 0 {
					t.Errorf("%s: Panik %v (%v), Body %q, erwartet den Abbruch mit http.ErrAbortHandler ohne Body", name, panikt, wert, rec.Body.String())
				}
				u.keinePortAufrufe(t, name)
			case sagt400:
				pruefeFehlerDerRoute(t, name, v.packe(f.err), 400, objekt400, "ODR-E2102")
			case sagt404:
				pruefeFehlerDerRoute(t, name, v.packe(f.err), 404, objektE7003, "ODR-E7003")
			default:
				pruefeFehlerDerRoute(t, name, v.packe(f.err), 500, objekt500, "ODR-E9001")
			}
		}
	}
	gleichAuf(t, "Zahl der Fälle", faelle, 9*5)
}

// TestReihenfolgeDerSuche hält F-2 (SPEC-238 Nr. 6 a, „Wurzel zuerst, Unwrap() error und Unwrap() []error“): enthält ein
// Fehler zwei *inbound.Fehler, gilt der erste der Suche (Tiefensuche, Wurzel zuerst, die Elemente eines Join von vorn nach
// hinten). Menge: Position im Join (4 Elemente) für einen allein (8 Fälle) und für jedes Paar in beiden Besetzungen
// (6 Paare × 2 = 12), die Wurzel mit dem zweiten als Ursache (2), tief verpackte Paare (4), Paare aus `fmt.Errorf` mit zwei
// `%w` (2) und ein ungültiger neben einem gültigen in beiden Reihenfolgen (2): 30 Fälle.
func TestReihenfolgeDerSuche(t *testing.T) {
	a, b := fehlerMit(t, "ODR-E7003"), fehlerMit(t, "ODR-E7001")
	codeVon := map[*inbound.Fehler]string{a: "ODR-E7003", b: "ODR-E7001"}
	plain := func(i int) error { return fmt.Errorf("plain %d", i) }
	faelle := 0
	pruefe := func(name string, err error, erwartet *inbound.Fehler) {
		faelle++
		pruefeFehlerDerRoute(t, name, err, 404, bodyZu(t, codeVon[erwartet]), codeVon[erwartet])
	}
	join4 := func(besetzung map[int]error) error {
		var teile []error
		for i := range 4 {
			if e, ok := besetzung[i]; ok {
				teile = append(teile, e)
			} else {
				teile = append(teile, plain(i))
			}
		}
		return errors.Join(teile...)
	}
	for i := range 4 {
		pruefe(fmt.Sprintf("a allein an Stelle %d", i), join4(map[int]error{i: a}), a)
		pruefe(fmt.Sprintf("b allein an Stelle %d", i), join4(map[int]error{i: b}), b)
		for j := i + 1; j < 4; j++ {
			pruefe(fmt.Sprintf("a an %d vor b an %d", i, j), join4(map[int]error{i: a, j: b}), a)
			pruefe(fmt.Sprintf("b an %d vor a an %d", i, j), join4(map[int]error{i: b, j: a}), b)
		}
	}
	pruefe("Wurzel a mit Ursache b", &inbound.Fehler{Code: a.Code, Status: 404, Ursache: b}, a)
	pruefe("Wurzel b mit Ursache a", &inbound.Fehler{Code: b.Code, Status: 404, Ursache: a}, b)
	tief := func(e error) error { return fmt.Errorf("x: %w", errors.Join(plain(0), fmt.Errorf("y: %w", e))) }
	pruefe("tief a, dann b", errors.Join(tief(a), b), a)
	pruefe("b, dann tief a", errors.Join(b, tief(a)), b)
	pruefe("tief a, dann tief b", fmt.Errorf("z: %w", errors.Join(tief(a), tief(b))), a)
	pruefe("tief b, dann tief a", fmt.Errorf("z: %w", errors.Join(tief(b), tief(a))), b)
	pruefe("zwei %w, a vorn", fmt.Errorf("%w und %w", a, b), a)
	pruefe("zwei %w, b vorn", fmt.Errorf("%w und %w", b, a), b)
	ungueltig := &inbound.Fehler{Code: 0, Status: 500}
	faelle++
	pruefeFehlerDerRoute(t, "ungültiger vor gültigem", errors.Join(ungueltig, a), 500, objekt500, "ODR-E9001")
	pruefe("gültiger vor ungültigem", errors.Join(a, ungueltig), a)
	gleichAuf(t, "Zahl der Fälle", faelle, 30)
}

// zyklusFehler ist ein Fehler, dessen Unwrap auf einen weiteren zeigt; ein Zyklus entsteht durch das Setzen von weiter.
type zyklusFehler struct{ weiter error }

func (*zyklusFehler) Error() string { return "zyklischer Fehler" }

func (e *zyklusFehler) Unwrap() error { return e.weiter }

// zyklusJoin ist ein Fehler mit mehreren Kindern, der sich selbst enthalten kann.
type zyklusJoin struct{ kinder []error }

func (*zyklusJoin) Error() string { return "zyklischer Join" }

func (e *zyklusJoin) Unwrap() []error { return e.kinder }

// TestZyklusInDerSuche hält F-8: ein Zeiger, der auf dem Weg von der Wurzel schon vorkam, wird nicht noch einmal
// betreten; ein Fehler, dessen Unwrap auf sich oder einen Vorfahren zeigt (Unwrap() error und Unwrap() []error), beendet
// die Suche ohne Treffer statt den Stapel zu füllen, und was neben dem Zyklus steht, gilt. 6 Formen. Der Stapel ist auf
// 16 MiB begrenzt, damit ein Zyklus ohne Schutz schnell endet.
func TestZyklusInDerSuche(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(16 << 20))
	selbst := &zyklusFehler{}
	selbst.weiter = selbst
	a1, a2 := &zyklusFehler{}, &zyklusFehler{}
	a1.weiter, a2.weiter = a2, a1
	j := &zyklusJoin{}
	j.kinder = []error{fmt.Errorf("plain"), j}
	jUrsache := &zyklusFehler{}
	jUrsache.weiter = &zyklusJoin{kinder: []error{jUrsache}}
	gueltig := fehlerMit(t, "ODR-E7003")
	pruefeFehlerDerRoute(t, "Fehler zeigt auf sich", selbst, 500, objekt500, "ODR-E9001")
	pruefeFehlerDerRoute(t, "zwei Fehler im Kreis", a1, 500, objekt500, "ODR-E9001")
	pruefeFehlerDerRoute(t, "Join enthält sich", j, 500, objekt500, "ODR-E9001")
	pruefeFehlerDerRoute(t, "Kreis über Fehler und Join", jUrsache, 500, objekt500, "ODR-E9001")
	pruefeFehlerDerRoute(t, "Zyklus vor einem gültigen Fehler", errors.Join(selbst, gueltig), 404, objektE7003, "ODR-E7003")
	pruefeFehlerDerRoute(t, "Zyklus vor ErrUngueltig", errors.Join(a1, model.ErrUngueltig), 400, objekt400, "ODR-E2102")
}

// syncPuffer ist ein Puffer, in den mehrere Goroutinen schreiben dürfen.
type syncPuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (p *syncPuffer) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.Write(b)
}

func (p *syncPuffer) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.String()
}

// TestAbbruchMitErrAbortHandler hält F-3 (SPEC-238 Nr. 9): ein Kontextfehler der Route bricht die Verbindung mit
// `panic(http.ErrAbortHandler)` ab, ohne Fehlerobjekt und ohne Antwort. Zwei Beobachtungen je Fehler (Canceled,
// DeadlineExceeded): an ServeHTTP ist der Wert der Panik http.ErrAbortHandler und der Recorder ohne Body, ohne Header und ohne
// Zählen; auf einem laufenden Server unterdrückt net/http für diesen Wert jede Meldung (kein `http: panic serving` an
// einen ErrorLog, den der Test an den Server hängt), der Client sieht einen Transportfehler und kein Status.
func TestAbbruchMitErrAbortHandler(t *testing.T) {
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		u, route := einRoute(t)
		route.err = e
		rec, wert, panikt := u.rufeMitPanik(anfrageMit("GET", "/r", nil, nil))
		if !panikt || wert != http.ErrAbortHandler {
			t.Errorf("%v: Panik %v mit dem Wert %v, erwartet http.ErrAbortHandler", e, panikt, wert)
		}
		if rec.Body.Len() != 0 || len(rec.Header().Values("Content-Type")) != 0 {
			t.Errorf("%v: Body %q und Content-Type %q, erwartet keine Antwort", e, rec.Body.String(), rec.Header().Values("Content-Type"))
		}
		u.keinePortAufrufe(t, e.Error())

		u2, route2 := einRoute(t)
		route2.err = e
		var meldungen syncPuffer
		ts := httptest.NewUnstartedServer(u2.srv)
		ts.Config.ErrorLog = log.New(&meldungen, "", 0)
		ts.Start()
		req, err := http.NewRequest("GET", ts.URL+"/r", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+string(tokenGueltig()))
		resp, err := (&http.Client{Transport: &http.Transport{DisableKeepAlives: true}}).Do(req)
		if err == nil {
			_ = resp.Body.Close()
			t.Errorf("%v: der Client sah den Status %d, erwartet einen Transportfehler", e, resp.StatusCode)
		}
		ts.Close()
		if got := meldungen.String(); got != "" {
			t.Errorf("%v: net/http meldete %q, erwartet keine Meldung für http.ErrAbortHandler", e, got)
		}
	}
}

// pfadeMitUnwrapAufZeiger liefert die Stellen `paket.Typ` der Methoden Unwrap mit Zeiger als Empfänger im Quelltext der
// Verzeichnisse (ohne _test.go).
func pfadeMitUnwrapAufZeiger(t *testing.T, verzeichnisse ...string) []string {
	t.Helper()
	var out []string
	for _, dir := range verzeichnisse {
		err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return err
			}
			f, perr := parser.ParseFile(token.NewFileSet(), p, nil, 0)
			if perr != nil {
				return perr
			}
			for _, decl := range f.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Name.Name != "Unwrap" || fd.Recv == nil || len(fd.Recv.List) != 1 {
					continue
				}
				if _, ok := fd.Recv.List[0].Type.(*ast.StarExpr); ok {
					out = append(out, p)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return out
}

// TestTypenMitUnwrapAufZeiger hält das Maß des akzeptierten Negativs „nil-Zeiger anderer Fehlertypen“ gegen die Quelle:
// im Quelltext von src/hexagon und src/adapters/driven haben 13 Typen eine Methode Unwrap mit Zeiger als Empfänger; die
// Suche behandelt einen nil-Zeiger allein bei *inbound.Fehler. Wächst die Zahl, ist das Maß im Plan nachzuziehen.
func TestTypenMitUnwrapAufZeiger(t *testing.T) {
	stellen := pfadeMitUnwrapAufZeiger(t, filepath.Join("..", "..", "..", "hexagon"), filepath.Join("..", "..", "driven"))
	gleichAuf(t, "Zahl der Typen mit Unwrap auf einem Zeiger", len(stellen), 13)
}

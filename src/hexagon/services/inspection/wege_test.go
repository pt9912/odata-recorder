package inspection_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

// weg ist ein Weg des Dienstes (Modell-Menge der Wege des Plans): Name, Art und die Zeile der Tabelle
// Zusage · Fänger, die ihn fängt.
type weg struct{ name, art, zeile string }

const (
	artPort      = "Aufruf eines Ports über ein Feld von Konfig"
	artAbbildung = "Aufruf von mitCode"
	artUngueltig = "Stelle, die ErrUngueltig bildet"
	artCtxNil    = "Aufruf von ctxNil"
	artCtxErr    = "Aufruf von Err auf dem ctx"
)

// wege ist die Tabelle der Wege, die der Quelltext des Pakets tragen muss (17 Wege in fünf Arten).
func wege() []weg {
	return []weg{
		{"ctxNil", artUngueltig, "A1, A2"},
		{"New: Exchanges nil", artUngueltig, "K1"},
		{"New: Coverage nil", artUngueltig, "K2"},
		{"New: Zaehler nil", artUngueltig, "K2"},
		{"New: Modus", artUngueltig, "K3"},
		{"Exchanges: Nach unter 0", artUngueltig, "X1"},
		{"Exchanges: Limit unter 0", artUngueltig, "X1"},
		{"Coverage: Kennung leer", artUngueltig, "C1"},
		{"Exchanges: ctxNil", artCtxNil, "A1"},
		{"Coverage: ctxNil", artCtxNil, "A2"},
		{"Diagnose: ctxNil", artCtxNil, "A2"},
		{"Exchanges: ListBySession", artPort, "X7"},
		{"Coverage: List", artPort, "C1, C5"},
		{"Diagnose: Snapshot", artPort, "D2"},
		{"Exchanges: Fehler von ListBySession", artAbbildung, "X9"},
		{"Coverage: Fehler von List", artAbbildung, "C3"},
		{"Diagnose: ctx.Err", artCtxErr, "D1"},
	}
}

// Das Modul des Dienstes: die Importe des Pakets liegen darunter oder in der Standardbibliothek.
const hexagon = "github.com/pt9912/odata-recorder/src/hexagon/"

// importErlaubt prüft einen Importpfad des Dienstes (SPEC-236 Kopf): model, Inbound-, Outbound- und
// Support-Ports oder die Standardbibliothek ohne Umgebung, Netz, Log und Datenbank.
func importErlaubt(p string) bool {
	for _, ok := range []string{"model", "ports/inbound", "ports/outbound", "ports/support"} {
		if p == hexagon+ok || strings.HasPrefix(p, hexagon+ok+"/") {
			return true
		}
	}
	erstes, _, _ := strings.Cut(p, "/")
	if strings.Contains(erstes, ".") {
		return false
	}
	for _, verboten := range []string{"os", "net", "log", "syscall"} {
		if p == verboten || strings.HasPrefix(p, verboten+"/") {
			return false
		}
	}
	return p != "database/sql" && !strings.HasPrefix(p, "database/sql/")
}

// TestWegeImQuelltext (W1, W3): die Aufrufe der Ports über Felder von Konfig, die Aufrufe von mitCode und
// ctxNil, die Aufrufe von Err auf dem ctx und die Stellen, die ErrUngueltig bilden, zählt der Test im
// Quelltext des Pakets (ohne _test) und hält sie gleich den Wegen der Tabelle; ein weiterer Aufruf oder
// eine weitere Stelle ohne Zeile macht ihn rot. Dazu liegt jeder Import des Pakets in der erlaubten
// Menge. Grenze: ein Aufruf über eine lokale Variable und ein Aufruf von Abbilden ohne mitCode zählen
// nicht, ebenso keine Verzweigung ohne Aufruf; Get und Append fängt zur Laufzeit TestLesenSchreibtNichts.
func TestWegeImQuelltext(t *testing.T) {
	felderKonfig := map[string]bool{}
	for _, f := range reflect.VisibleFields(reflect.TypeFor[inspection.Konfig]()) {
		felderKonfig[f.Name] = true
	}
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	gezaehlt, dateien := map[string]int{}, 0
	for _, e := range eintraege {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		dateien++
		datei, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range datei.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil || !importErlaubt(p) {
				t.Errorf("%s importiert %s", e.Name(), imp.Path.Value)
			}
		}
		ast.Inspect(datei, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "mitCode" {
					gezaehlt[artAbbildung]++
				}
				if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "ctxNil" {
					gezaehlt[artCtxNil]++
				}
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
					if feld, ok := sel.X.(*ast.SelectorExpr); ok && felderKonfig[feld.Sel.Name] {
						gezaehlt[artPort]++
					}
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "ctx" && sel.Sel.Name == "Err" {
						gezaehlt[artCtxErr]++
					}
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "ErrUngueltig" {
					gezaehlt[artUngueltig]++
				}
			}
			return true
		})
	}
	if dateien == 0 {
		t.Fatal("keine Quelldatei des Pakets gefunden")
	}
	soll := map[string]int{}
	for _, w := range wege() {
		soll[w.art]++
	}
	if !reflect.DeepEqual(gezaehlt, soll) {
		t.Errorf("im Quelltext %v, Wege der Tabelle %v", gezaehlt, soll)
	}
}

// wegeExchanges ist die Modell-Menge der Wege von Exchanges (Plan §3, Tabelle *Wege*) ohne die Felder
// der Allowlist (X11): jeder Verwurf nach Grund, jede Seitengrenze, jede Fehlerabbildung, Erfolg.
func wegeExchanges() []string {
	return []string{
		"ctx nil", "Nach < 0", "Limit < 0", "Limit 0", "Limit über der Grenze", "Limit sonst", "Kontextfehler",
		"Fehler mit Code", "Erfolg", "Filter", "Schnitt", "Weiter gesetzt", "Weiter 0",
	}
}

// wegeCoverage und wegeDiagnose sind die Modell-Mengen der Wege von Coverage und Diagnose.
func wegeCoverage() []string {
	return []string{
		"Coverage: ctx nil", "Coverage: Kennung leer", "Coverage: Kontextfehler", "Coverage: Fehler mit Code",
		"Coverage: Erfolg", "Coverage: leere Liste",
	}
}

func wegeDiagnose() []string {
	return []string{
		"Diagnose: ctx nil", "Diagnose: ctx beendet", "Diagnose: Snapshot", "Diagnose: Eintrag gesetzt",
		"Diagnose: Eintrag fehlt", "Diagnose: Eintrag außerhalb", "Diagnose: Schlüssel erlaubt",
		"Diagnose: Schlüssel verworfen", "Diagnose: Schlüssel fehlt", "Diagnose: Modus",
	}
}

// zweig ist ein Fall der Menge W4 für Exchanges: die Wege, die er ausführt, und die Erwartung (Art:
// seite, ungueltig, kontext oder code).
type zweig struct {
	name    string
	wege    []string
	art     string
	n       int
	seite   inbound.Seite
	ohneCtx bool
	fehler  error
	code    model.Code
}

// zweigeExchanges erzeugt die Zweige von Exchanges: jeder Weg, die Fehler des Ports je Eintrag der
// Abbildung.
func zweigeExchanges() []zweig {
	z := []zweig{
		{name: "ctx nil", wege: []string{"ctx nil"}, art: "ungueltig", n: 5, ohneCtx: true},
		{name: "Nach unter 0", wege: []string{"Nach < 0"}, art: "ungueltig", n: 5, seite: inbound.Seite{Nach: -1}},
		{name: "Limit unter 0", wege: []string{"Limit < 0"}, art: "ungueltig", n: 5, seite: inbound.Seite{Limit: -1}},
		{name: "Limit 0, Schnitt", wege: []string{"Limit 0", "Schnitt", "Filter", "Weiter gesetzt", "Erfolg"}, art: "seite", n: 150},
		{name: "Limit über der Grenze", wege: []string{"Limit über der Grenze", "Schnitt", "Weiter gesetzt", "Erfolg"}, art: "seite", n: 1100, seite: inbound.Seite{Limit: math.MaxInt}},
		{name: "Limit sonst, Filter", wege: []string{"Limit sonst", "Filter", "Schnitt", "Weiter gesetzt", "Erfolg"}, art: "seite", n: 5, seite: inbound.Seite{Limit: 2, Nach: 1}},
		{name: "Weiter 0", wege: []string{"Weiter 0", "Erfolg"}, art: "seite", n: 5, seite: inbound.Seite{Limit: 10}},
		{name: "leere Session", wege: []string{"Weiter 0", "Erfolg"}, art: "seite", n: 0},
	}
	for _, k := range kontextfehler() {
		z = append(z, zweig{name: "Kontextfehler " + k.name, wege: []string{"Kontextfehler"}, art: "kontext", n: 1, fehler: k.err})
	}
	for _, c := range tabelleSession() {
		z = append(z, zweig{name: "Fehler " + c.name, wege: []string{"Fehler mit Code"}, art: "code", n: 1, fehler: fmt.Errorf("lesen: %w", c.err), code: c.code})
	}
	return z
}

// pruefe vergleicht das Ergebnis eines Aufrufs von Exchanges mit der Erwartung des Zweigs.
func (zw zweig) pruefe(t *testing.T, was string, z inbound.ExchangeSeite, err error) {
	t.Helper()
	var fe *inbound.Fehler
	switch zw.art {
	case "seite":
		if err != nil {
			t.Errorf("%s: Fehler %v", was, err)
			return
		}
		gleichSeite(t, was, z, zw.n, zw.seite.Nach, zw.seite.Limit)
	case "ungueltig":
		ohneCode(t, was, err)
	case "kontext":
		if err != zw.fehler || errors.As(err, &fe) {
			t.Errorf("%s: Fehler %v", was, err)
		}
	default:
		if !errors.As(err, &fe) || fe.Code != zw.code {
			t.Errorf("%s: Fehler %v, erwartet Code %v", was, err, zw.code)
		}
	}
}

// lauf ist ein Zweig der Menge W4 einer Operation: bau legt Fakes und einen Dienst an und liefert die
// Funktion, die jede Goroutine in jeder Runde ruft; sie prüft das Ergebnis gegen die Erwartung des Zweigs
// und schreibt danach in ihr Ergebnis (ein geteilter Speicher bricht unter dem Detektor).
type lauf struct {
	op, name, art string
	wege          []string
	bau           func(t *testing.T) func(was string)
}

// laufExchanges bildet den Zweig eines Exchanges-Falls.
func laufExchanges(zw zweig) lauf {
	return lauf{op: "Exchanges", name: zw.name, art: zw.art, wege: zw.wege, bau: func(t *testing.T) func(string) {
		f := neuesFake()
		f.legeAn(sid, zw.n)
		f.fehler = zw.fehler
		d := neuerDienst(t, f)
		return func(was string) {
			ctx := t.Context()
			if zw.ohneCtx {
				ctx = nilCtx()
			}
			z, err := d.Exchanges(ctx, sid, zw.seite)
			zw.pruefe(t, was, z, err)
		}
	}}
}

// laufCoverage bildet den Zweig eines Falls von Coverage: setup richtet die Fakes ein, ctx liefert den
// Kontext, soll prüft das Ergebnis.
func laufCoverage(name, art string, wege []string, id string, setup func(u *umgebung), ctx func(t *testing.T) context.Context,
	soll func(t *testing.T, was string, erg []inbound.CoverageEintrag, err error)) lauf {
	return lauf{op: "Coverage", name: name, art: art, wege: wege, bau: func(t *testing.T) func(string) {
		u := neueUmgebung()
		setup(u)
		d := u.dienst(t)
		return func(was string) {
			erg, err := d.Coverage(ctx(t), id)
			soll(t, was, erg, err)
			if len(erg) > 0 {
				erg[0].Evidenz += "geaendert"
				erg[0].EntitySet += "x"
			}
		}
	}}
}

func lebend(t *testing.T) context.Context { return t.Context() }

// pruefeFehlerCode verlangt einen *Fehler mit dem Code und das Ergebnis nil.
func pruefeFehlerCode(code model.Code) func(*testing.T, string, []inbound.CoverageEintrag, error) {
	return func(t *testing.T, was string, erg []inbound.CoverageEintrag, err error) {
		t.Helper()
		var fe *inbound.Fehler
		if !errors.As(err, &fe) || fe.Code != code || erg != nil {
			t.Errorf("%s: Fehler %v, Ergebnis nil: %v, erwartet Code %v", was, err, erg == nil, code)
		}
	}
}

// zweigeCoverage erzeugt die Zweige von Coverage: ctx nil, Kennung leer, Erfolg mit nil, leerer Liste
// und 101 Einträgen, die fünf Formen des Kontextfehlers und je Eintrag der Abbildung ein Fehler mit Code
// und ein anderer.
func zweigeCoverage() []lauf {
	keine := func(*umgebung) {}
	erfolg := func(m int) func(*testing.T, string, []inbound.CoverageEintrag, error) {
		return func(t *testing.T, was string, erg []inbound.CoverageEintrag, err error) {
			t.Helper()
			if err != nil || erg == nil || len(erg) != m {
				t.Errorf("%s: Fehler %v, Ergebnis nil: %v, %d Einträge, erwartet %d", was, err, erg == nil, len(erg), m)
				return
			}
			for i, e := range erg {
				if w := eintrag(m, i, 0, 0); e != w {
					t.Errorf("%s: Eintrag %d ist %+v, erwartet %+v", was, i, e, w)
				}
			}
		}
	}
	ungueltig := func(t *testing.T, was string, erg []inbound.CoverageEintrag, err error) {
		t.Helper()
		ohneCode(t, was, err)
		if erg != nil {
			t.Errorf("%s: Ergebnis %v statt nil", was, erg)
		}
	}
	l := []lauf{
		laufCoverage("ctx nil", "ungueltig", []string{"Coverage: ctx nil"}, "b1", keine, func(*testing.T) context.Context { return nilCtx() }, ungueltig),
		laufCoverage("Kennung leer", "ungueltig", []string{"Coverage: Kennung leer"}, "", keine, lebend, ungueltig),
		laufCoverage("Erfolg, nil vom Fake", "erfolg", []string{"Coverage: Erfolg", "Coverage: leere Liste"}, "b1",
			func(u *umgebung) { u.cov.setzeListe(nil) }, lebend, erfolg(0)),
		laufCoverage("Erfolg, leere Liste", "erfolg", []string{"Coverage: Erfolg", "Coverage: leere Liste"}, "b1",
			func(u *umgebung) { u.cov.setzeListe([]model.Coverage{}) }, lebend, erfolg(0)),
		laufCoverage("Erfolg, 101 Einträge", "erfolg", []string{"Coverage: Erfolg"}, "b1", func(u *umgebung) {
			fix := make([]model.Coverage, 101)
			for i := range fix {
				e := eintrag(101, i, 0, 0)
				fix[i] = model.Coverage{EntitySet: e.EntitySet, Status: e.Status, EvidenceJSON: e.Evidenz}
			}
			u.cov.setzeListe(fix)
		}, lebend, erfolg(101)),
	}
	for _, k := range kontextfehler() {
		l = append(l, laufCoverage("Kontextfehler "+k.name, "kontext", []string{"Coverage: Kontextfehler"}, "b1",
			func(u *umgebung) {
				u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}})
				u.cov.fehler = k.err
			}, lebend, func(t *testing.T, was string, erg []inbound.CoverageEintrag, err error) {
				t.Helper()
				var fe *inbound.Fehler
				if err != k.err || errors.As(err, &fe) || erg != nil {
					t.Errorf("%s: Fehler %v, Ergebnis nil: %v", was, err, erg == nil)
				}
			}))
	}
	for _, z := range tabelleSnapshot() {
		l = append(l, laufCoverage("Fehler "+z.name, "code", []string{"Coverage: Fehler mit Code"}, "b1",
			func(u *umgebung) {
				u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}})
				u.cov.fehler = fmt.Errorf("lesen: %w", z.err)
			}, lebend, pruefeFehlerCode(z.code)))
	}
	return l
}

// sollDiagnose prüft eine Diagnose gegen das Orakel: Modus, Einstellungen und Zähler je Code in der
// Reihenfolge von model.AlleCodes().
func sollDiagnose(t *testing.T, was string, d inbound.Diagnose, err error, modus string, einst map[string]string, zaehler map[model.Code]uint64) {
	t.Helper()
	codes := model.AlleCodes()
	if err != nil || d.Modus != modus || d.Einstellungen == nil || !reflect.DeepEqual(d.Einstellungen, einst) || len(d.Fehler) != len(codes) {
		t.Errorf("%s: Fehler %v, Modus %q, Einstellungen %v, %d Zähler", was, err, d.Modus, d.Einstellungen, len(d.Fehler))
		return
	}
	for i, c := range codes {
		if d.Fehler[i].Code != c || d.Fehler[i].Anzahl != zaehler[c] {
			t.Errorf("%s: Zähler %d ist %+v, erwartet %v %d", was, i, d.Fehler[i], c, zaehler[c])
		}
	}
}

// laufDiagnose bildet den Zweig eines Falls von Diagnose.
func laufDiagnose(name, art string, wege []string, setup func(u *umgebung), ctx func(t *testing.T) context.Context,
	soll func(t *testing.T, was string, d inbound.Diagnose, err error)) lauf {
	return lauf{op: "Diagnose", name: name, art: art, wege: wege, bau: func(t *testing.T) func(string) {
		u := neueUmgebung()
		setup(u)
		d := u.dienst(t)
		return func(was string) {
			erg, err := d.Diagnose(ctx(t))
			soll(t, was, erg, err)
			if erg.Einstellungen != nil {
				erg.Einstellungen["LINK_REWRITE"] = "geaendert"
				erg.Einstellungen["ZZ"] = "neu"
				delete(erg.Einstellungen, "MATCH_BODY_MODE")
			}
			if len(erg.Fehler) > 0 {
				erg.Fehler[0].Anzahl++
			}
		}
	}}
}

// zweigeDiagnose erzeugt die zehn Zweige von Diagnose: ctx nil, ctx beendet (zwei Formen), Errors nil,
// Errors mit Eintrag je Code und einem außerhalb, Einstellungen nil, Einstellungen mit erlaubtem,
// verworfenem und fehlendem Schlüssel in je einem der vier Modi.
func zweigeDiagnose() []lauf {
	keine := func(*umgebung) {}
	abgelaufen, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	cancel()
	zero := func(fehler error) func(*testing.T, string, inbound.Diagnose, error) {
		return func(t *testing.T, was string, d inbound.Diagnose, err error) {
			t.Helper()
			if fehler == nil {
				ohneCode(t, was, err)
			} else if err != fehler {
				t.Errorf("%s: Fehler %v", was, err)
			}
			if !reflect.DeepEqual(d, inbound.Diagnose{}) {
				t.Errorf("%s: Ergebnis %+v statt Nullwert", was, d)
			}
		}
	}
	nullen := map[model.Code]uint64{}
	l := []lauf{
		laufDiagnose("ctx nil", "ungueltig", []string{"Diagnose: ctx nil"}, keine, func(*testing.T) context.Context { return nilCtx() }, zero(nil)),
		laufDiagnose("ctx Canceled", "kontext", []string{"Diagnose: ctx beendet"}, keine, func(*testing.T) context.Context { return ctxBeendet() }, zero(context.Canceled)),
		laufDiagnose("ctx DeadlineExceeded", "kontext", []string{"Diagnose: ctx beendet"}, keine, func(*testing.T) context.Context { return abgelaufen }, zero(context.DeadlineExceeded)),
		laufDiagnose("Errors nil", "erfolg", []string{"Diagnose: Snapshot", "Diagnose: Eintrag fehlt", "Diagnose: Modus"},
			func(u *umgebung) { u.tel.SetzeStand(support.Snapshot{}) }, lebend,
			func(t *testing.T, was string, d inbound.Diagnose, err error) {
				t.Helper()
				sollDiagnose(t, was, d, err, "record", map[string]string{}, nullen)
			}),
	}
	errs, _ := standMCodes(1)
	errs[model.Code(math.MaxInt)] = 3
	l = append(l, laufDiagnose("Errors je Code und außerhalb", "erfolg",
		[]string{"Diagnose: Snapshot", "Diagnose: Eintrag gesetzt", "Diagnose: Eintrag fehlt", "Diagnose: Eintrag außerhalb", "Diagnose: Modus"},
		func(u *umgebung) { u.tel.SetzeStand(support.Snapshot{Errors: errs}) }, lebend,
		func(t *testing.T, was string, d inbound.Diagnose, err error) {
			t.Helper()
			sollDiagnose(t, was, d, err, "record", map[string]string{}, errs)
		}))
	l = append(l, laufDiagnose("Einstellungen nil", "erfolg", []string{"Diagnose: Snapshot", "Diagnose: Schlüssel fehlt", "Diagnose: Modus"},
		func(u *umgebung) { u.k.Einstellungen = nil }, lebend,
		func(t *testing.T, was string, d inbound.Diagnose, err error) {
			t.Helper()
			sollDiagnose(t, was, d, err, "record", map[string]string{}, nullen)
		}))
	for _, modus := range modiV() {
		l = append(l, laufDiagnose("Einstellungen, Modus "+modus, "erfolg",
			[]string{"Diagnose: Snapshot", "Diagnose: Schlüssel erlaubt", "Diagnose: Schlüssel verworfen", "Diagnose: Schlüssel fehlt", "Diagnose: Modus"},
			func(u *umgebung) {
				u.k.Modus = modus
				u.k.Einstellungen = map[string]string{"LINK_REWRITE": "off", "DATA_DIR": "/geheim"}
			}, lebend, func(t *testing.T, was string, d inbound.Diagnose, err error) {
				t.Helper()
				sollDiagnose(t, was, d, err, modus, map[string]string{"LINK_REWRITE": "off"}, nullen)
			}))
	}
	return l
}

// zweige ist die Menge W4: die Zweige aller drei Operationen.
func zweige() []lauf {
	var l []lauf
	for _, zw := range zweigeExchanges() {
		l = append(l, laufExchanges(zw))
	}
	l = append(l, zweigeCoverage()...)
	return append(l, zweigeDiagnose()...)
}

// TestZweigeSindDieWege (W4, Zählprüfung): die Zweige von TestGleichzeitig führen genau die Wege der
// Modell-Mengen aller Operationen, die Fehler mit Code einen je Eintrag von outbound.EintraegeAbbildung()
// und einen anderen, die Kontextfehler fünf je Operation mit Port, Diagnose zwei beendete Kontexte.
func TestZweigeSindDieWege(t *testing.T) {
	gefuehrt := map[string]bool{}
	code, kontext := map[string]int{}, map[string]int{}
	je := map[string]int{}
	for _, zw := range zweige() {
		je[zw.op]++
		for _, w := range zw.wege {
			gefuehrt[w] = true
		}
		switch zw.art {
		case "code":
			code[zw.op]++
		case "kontext":
			kontext[zw.op]++
		}
	}
	soll := slices.Concat(wegeExchanges(), wegeCoverage(), wegeDiagnose())
	for _, w := range soll {
		if !gefuehrt[w] {
			t.Errorf("der Weg %q hat keinen Zweig", w)
		}
	}
	if len(gefuehrt) != len(soll) {
		t.Errorf("Zweige führen %d Wege, die Modell-Mengen haben %d", len(gefuehrt), len(soll))
	}
	n := len(outbound.EintraegeAbbildung())
	for _, op := range []string{"Exchanges", "Coverage"} {
		if code[op] != n+1 || kontext[op] != 5 {
			t.Errorf("%s: Fehler mit Code %d, Quelle %d + 1; Kontextfehler %d", op, code[op], n, kontext[op])
		}
	}
	if kontext["Diagnose"] != 2 || code["Diagnose"] != 0 {
		t.Errorf("Diagnose: beendete Kontexte %d, Fehler mit Code %d", kontext["Diagnose"], code["Diagnose"])
	}
	if want := 8 + 5 + n + 1; je["Exchanges"] != want {
		t.Errorf("Exchanges: %d Zweige, erwartet %d", je["Exchanges"], want)
	}
	if want := 5 + 5 + n + 1; je["Coverage"] != want {
		t.Errorf("Coverage: %d Zweige, erwartet %d", je["Coverage"], want)
	}
	if je["Diagnose"] != 1+2+1+1+1+len(modiV()) {
		t.Errorf("Diagnose: %d Zweige", je["Diagnose"])
	}
}

// TestGleichzeitig (W4): je Zweig der Menge acht Goroutinen mit je 20 Runden gegen einen Dienst; jedes
// Ergebnis gleicht der Erwartung des Zweigs. Der Dienst hält keinen Zustand; ein Zustand in ihm ist ein
// Datenrennen, das make test-race je Prozess meldet, aber nur in den Zweigen, die der Test ausführt,
// deshalb läuft jeder Zweig der Menge gleichzeitig.
func TestGleichzeitig(t *testing.T) {
	for _, zw := range zweige() {
		run := zw.bau(t)
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Go(func() {
				for i := range 20 {
					run(fmt.Sprintf("%s: %s, Goroutine %d Runde %d", zw.op, zw.name, g, i))
				}
			})
		}
		wg.Wait()
	}
}

// nummerierteTelemetrie liefert Snapshot mit einer fortlaufenden Nummer im Zähler des ersten Codes.
func nummerierteTelemetrie(u *umgebung) {
	var n atomic.Uint64
	u.k.Zaehler = &hakenTelemetrie{FakeTelemetry: u.tel, haken: func(s *support.Snapshot) {
		if s.Errors == nil {
			s.Errors = map[model.Code]uint64{}
		}
		s.Errors[model.AlleCodes()[0]] = n.Add(1)
	}}
}

// pruefeNummern verlangt die Nummern 1 bis n lückenlos und paarweise verschieden.
func pruefeNummern(t *testing.T, was string, nums []int, n int) {
	t.Helper()
	slices.Sort(nums)
	for i, x := range nums {
		if x != i+1 {
			t.Errorf("%s: Nummern nicht lückenlos 1 bis %d, an Stelle %d steht %d", was, n, i, x)
			return
		}
	}
	if len(nums) != n {
		t.Errorf("%s: %d Antworten, erwartet %d", was, len(nums), n)
	}
}

// TestGleichzeitigWirkung (W4, Wirkungs-Fänger): bei N gleichzeitigen Aufrufen sind die N Antworten
// von Coverage (Nummer in der Evidenz des ersten Eintrags) und von Diagnose (Nummer im Zähler des ersten
// Codes) lückenlos 1 bis N und paarweise verschieden, und jeder Aufruf erreicht seinen Port genau einmal;
// ein Zwischenspeicher liefert eine Nummer doppelt, auch wenn der Detektor schweigt.
func TestGleichzeitigWirkung(t *testing.T) {
	const goroutinen, runden = 8, 20
	n := goroutinen * runden
	u := neueUmgebung()
	u.cov.nummeriere = true
	u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}, {EntitySet: "B", Status: "partial"}})
	nummerierteTelemetrie(u)
	d := u.dienst(t)
	var mu sync.Mutex
	var cov, dia []int
	var wg sync.WaitGroup
	for range goroutinen {
		wg.Go(func() {
			for range runden {
				erg, err := d.Coverage(t.Context(), "b1")
				diag, err3 := d.Diagnose(t.Context())
				if err != nil || err3 != nil || len(erg) != 2 {
					t.Errorf("Fehler %v %v, %d Einträge", err, err3, len(erg))
					return
				}
				x, err2 := strconv.Atoi(erg[0].Evidenz)
				if err2 != nil {
					t.Errorf("Evidenz %q: %v", erg[0].Evidenz, err2)
					return
				}
				mu.Lock()
				cov, dia = append(cov, x), append(dia, int(diag.Fehler[0].Anzahl))
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	pruefeNummern(t, "Coverage", cov, n)
	pruefeNummern(t, "Diagnose", dia, n)
	if u.cov.anzahl("List") != n || u.tel.Anzahl(supporttest.OpSnapshot) != n {
		t.Errorf("List %d, Snapshot %d, erwartet je %d", u.cov.anzahl("List"), u.tel.Anzahl(supporttest.OpSnapshot), n)
	}
}

// TestGleichzeitigGemischt (W4): dieselben acht Goroutinen rufen je Runde alle drei Operationen in
// wechselnder Reihenfolge gegen einen Dienst; jedes Ergebnis gleicht dem sequentiellen, die Nummern der
// Antworten von Coverage und Diagnose sind lückenlos 1 bis N, und jeder Port wird N-mal gerufen.
func TestGleichzeitigGemischt(t *testing.T) {
	const goroutinen, runden = 8, 20
	n := goroutinen * runden
	u := neueUmgebung()
	u.rec.legeAn(sid, 5)
	u.cov.nummeriere = true
	u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}})
	nummerierteTelemetrie(u)
	d := u.dienst(t)
	folgen := [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	var mu sync.Mutex
	var cov, dia []int
	var wg sync.WaitGroup
	for g := range goroutinen {
		wg.Go(func() {
			for i := range runden {
				for _, op := range folgen[(g+i)%len(folgen)] {
					was := fmt.Sprintf("Goroutine %d Runde %d Operation %d", g, i, op)
					switch op {
					case 0:
						z, err := d.Exchanges(t.Context(), sid, inbound.Seite{})
						if err != nil {
							t.Errorf("%s: %v", was, err)
							continue
						}
						gleichSeite(t, was, z, 5, 0, 0)
					case 1:
						erg, err := d.Coverage(t.Context(), "b1")
						x, err2 := 0, error(nil)
						if len(erg) == 1 {
							x, err2 = strconv.Atoi(erg[0].Evidenz)
						}
						if err != nil || err2 != nil || len(erg) != 1 || erg[0].EntitySet != "A" {
							t.Errorf("%s: %v %v %v", was, err, err2, erg)
							continue
						}
						mu.Lock()
						cov = append(cov, x)
						mu.Unlock()
					default:
						diag, err := d.Diagnose(t.Context())
						if err != nil || len(diag.Fehler) != len(model.AlleCodes()) {
							t.Errorf("%s: %v, %d Zähler", was, err, len(diag.Fehler))
							continue
						}
						mu.Lock()
						dia = append(dia, int(diag.Fehler[0].Anzahl))
						mu.Unlock()
					}
				}
			}
		})
	}
	wg.Wait()
	pruefeNummern(t, "Coverage", cov, n)
	pruefeNummern(t, "Diagnose", dia, n)
	if u.rec.anzahl("ListBySession") != n || u.cov.anzahl("List") != n || u.tel.Anzahl(supporttest.OpSnapshot) != n {
		t.Errorf("ListBySession %d, List %d, Snapshot %d, erwartet je %d", u.rec.anzahl("ListBySession"), u.cov.anzahl("List"), u.tel.Anzahl(supporttest.OpSnapshot), n)
	}
}

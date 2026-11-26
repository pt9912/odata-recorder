package bootstrap_test

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// verwendungen zählt in einer Quelle je `paket.Name` der Pakete admin und audit die Nennungen und hält je Aufruf die Argumente.
func verwendungen(t *testing.T, datei string, quelle any, nennungen map[string]int, aufrufe map[string][]string) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), datei, quelle, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, im := range f.Imports {
		if p, _ := strconv.Unquote(im.Path.Value); strings.HasSuffix(p, "/driving/admin") && im.Name != nil {
			t.Errorf("%s importiert admin unter dem Namen %s", datei, im.Name)
		}
	}
	name := func(e ast.Expr) string {
		s, ok := e.(*ast.SelectorExpr)
		if !ok {
			return ""
		}
		if id, ok := s.X.(*ast.Ident); ok && (id.Name == "admin" || id.Name == "audit") {
			return id.Name + "." + s.Sel.Name
		}
		return ""
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			var args []string
			for _, a := range x.Args {
				args = append(args, types.ExprString(a))
			}
			aufrufe[name(x.Fun)] = append(aufrufe[name(x.Fun)], strings.Join(args, ", "))
		case *ast.SelectorExpr:
			nennungen[name(x)]++
		}
		return true
	})
}

// TestAdminKonstruktorenEinmal fängt A07 und den Teil „die eine Instanz“ von A02 bis A05: im Quelltext von bootstrap ohne
// _test.go steht jeder Konstruktor der Routen genau einmal, als Aufruf mit den Feldern des Handle, die seine Zugänge liefern, und
// audit.New genau einmal; keine Zeichenkette eines Tests des Pakets admin enthält bootstrap.
func TestAdminKonstruktorenEinmal(t *testing.T) {
	nennungen, aufrufe := map[string]int{}, map[string][]string{}
	verwendungen(t, "probe.go", "package x\nfunc f() { admin.SessionRouten(a, b); admin.SessionRouten(a, b) }", nennungen, aufrufe)
	if nennungen["admin.SessionRouten"] != 2 || len(aufrufe["admin.SessionRouten"]) != 2 {
		t.Fatalf("Probe: %d Nennungen, %v", nennungen["admin.SessionRouten"], aufrufe["admin.SessionRouten"])
	}
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) < 10 {
		t.Fatalf("%d Dateien, %v", len(dateien), err)
	}
	nennungen, aufrufe = map[string]int{}, map[string][]string{}
	for _, d := range dateien {
		if !strings.HasSuffix(d, "_test.go") {
			verwendungen(t, d, nil, nennungen, aufrufe)
		}
	}
	soll := map[string]string{
		"admin.SessionRouten": "h.sessions, h.audit", "admin.InspektionRouten": "h.inspektion",
		"admin.ImportRouten": "h.importer, h.audit", "admin.WorkspaceRouten": "h.workspaces, h.snapshots, h.audit",
	}
	for name, args := range soll {
		if nennungen[name] != 1 || len(aufrufe[name]) != 1 || aufrufe[name][0] != args {
			t.Errorf("%s: %d Nennungen, Aufrufe %q; erwartet einen mit %q", name, nennungen[name], aufrufe[name], args)
		}
	}
	if nennungen["audit.New"] != 1 {
		t.Errorf("audit.New %d-mal genannt", nennungen["audit.New"])
	}
	fremde, err := filepath.Glob(filepath.Join("..", "adapters", "driving", "admin", "*_test.go"))
	if err != nil || len(fremde) < 10 {
		t.Fatalf("%d Tests von admin, %v", len(fremde), err)
	}
	for _, d := range fremde {
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if l, ok := n.(*ast.BasicLit); ok && l.Kind == token.STRING && strings.Contains(l.Value, "bootstrap") {
				t.Errorf("%s nennt %s", d, l.Value)
			}
			return true
		})
	}
}

type zaehlSchliesser struct {
	c io.Closer
	n *int
}

func (s zaehlSchliesser) Close() error { *s.n++; return s.c.Close() }

// TestAdminBuildFehler fängt A08: ein Fehler eines Konstruktors der Dienste und Routen ist der Fehler von Build, ohne Handle und
// mit einmal geschlossener Ablage (21 Zellen, dazu das nil von Snapshot- und Record-Dienst).
func TestAdminBuildFehler(t *testing.T) {
	scheitert := errors.New("Haken scheitert")
	sitzung := func(err error) func(*bootstrap.Bauteile) {
		return func(b *bootstrap.Bauteile) {
			b.NeueSession = func(session.Konfig) (inbound.SessionUseCase, error) { return nil, err }
		}
	}
	inspektion := func(err error) func(*bootstrap.Bauteile) {
		return func(b *bootstrap.Bauteile) {
			b.NeueInspektion = func(inspection.Konfig) (inbound.InspectionUseCase, error) { return nil, err }
		}
	}
	type fall struct {
		name, modus string
		setze       func(*bootstrap.Bauteile)
		soll        error
	}
	var faelle []fall
	for _, modus := range adminModi() {
		faelle = append(faelle,
			fall{"NeueSession scheitert", modus, sitzung(scheitert), scheitert},
			fall{"NeueInspektion scheitert", modus, inspektion(scheitert), scheitert},
			fall{"NeueSession nil", modus, sitzung(nil), model.ErrUngueltig},
			fall{"NeueInspektion nil", modus, inspektion(nil), model.ErrUngueltig},
			fall{"NeueWorkspaces nil", modus, func(b *bootstrap.Bauteile) {
				b.NeueWorkspaces = func(workspace.Konfig) (inbound.WorkspaceUseCase, error) { return nil, nil }
			}, model.ErrUngueltig})
	}
	faelle = append(faelle, fall{"NeuerImport nil", modusRecord, func(b *bootstrap.Bauteile) {
		b.NeuerImport = func(baselineimport.Konfig) (inbound.ImportUseCase, error) { return nil, nil }
	}, model.ErrUngueltig})
	if len(faelle) != 21 {
		t.Errorf("%d Zellen", len(faelle))
	}
	// Außerhalb der 21 Zellen: die zwei übrigen Konstruktoren, deren nil ein späterer meldet.
	for _, modus := range adminModi() {
		faelle = append(faelle, fall{"NeuerSnapshot nil", modus, func(b *bootstrap.Bauteile) {
			b.NeuerSnapshot = func(snapshot.Konfig) (inbound.SnapshotUseCase, error) { return nil, nil }
		}, model.ErrUngueltig})
	}
	faelle = append(faelle, fall{"NeuerRecorder nil", modusRecord, func(b *bootstrap.Bauteile) {
		b.NeuerRecorder = func(record.Konfig) (inbound.RecordUseCase, error) { return nil, nil }
	}, model.ErrUngueltig})
	for _, f := range faelle {
		b := bootstrap.StandardBauteile()
		geschlossen := 0
		b.Schliesser = func(c io.Closer) io.Closer { return zaehlSchliesser{c, &geschlossen} }
		f.setze(&b)
		h, err := bootstrap.BuildMit(context.Background(), routenKonfig(t.TempDir(), f.modus, "http://u1.test/o"), nil, b)
		if h != nil || !errors.Is(err, f.soll) || geschlossen != 1 {
			t.Errorf("%s/%s: Handle %v, Fehler %v, %d-mal geschlossen", f.name, f.modus, h != nil, err, geschlossen)
		}
		if h != nil {
			_ = h.Herunterfahren(context.Background())
		}
	}
}

package snapshot_test

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// loescheWorkspace löscht ws-1 samt State aus den Fakes.
func (w *welt) loescheWorkspace() {
	w.ws.loeschen("ws-1")
	w.ent.mu.Lock()
	delete(w.ent.ent, "ws-1")
	w.ent.mu.Unlock()
}

// TestFensterLesenAnlegen (T67): ein Ereignis nach Entities.List (Änderung, Tombstone, Restore des State) ändert den
// Snapshot nicht: das Manifest zeigt den State der Liste (Haltepunkt im Fake, in jedem Lauf).
func TestFensterLesenAnlegen(t *testing.T) {
	_, soll, _ := erzeugeMitState(t, standard("ws-1"))
	ereignisse := map[string]func(es []model.EntityState) []model.EntityState{
		"Änderung": func(es []model.EntityState) []model.EntityState {
			es[1].Payload, es[1].Version = `{"Id":"1","Name":"Spaeter"}`, 99
			return es
		},
		"Tombstone": func(es []model.EntityState) []model.EntityState { es[2].Deleted, es[2].Version = true, 50; return es },
		"Restore": func(es []model.EntityState) []model.EntityState {
			return []model.EntityState{{WorkspaceID: "ws-1", EntitySet: "Neu", EntityKey: "K", Payload: `{}`, Completeness: model.EntityComplete}}
		},
	}
	for name, f := range ereignisse {
		w := neu(t)
		w.p.nach = func(op string) {
			if op == "Entities.List" {
				w.p.nach = nil
				w.ent.setze("ws-1", f(standard("ws-1")))
			}
		}
		_, b, err := w.erzeuge("n")
		if err != nil || string(b) != string(soll) {
			t.Errorf("%s: %v\n%s", name, err, b)
		}
		if w.p.anzahl("Entities.List") != 1 {
			t.Errorf("%s: Entities.List %d mal", name, w.p.anzahl("Entities.List"))
		}
	}
	if len(ereignisse) != 3 {
		t.Error("M13")
	}
}

// TestCreateGegenLoeschen (T69): ein Löschen des Workspace zwischen N1 und N10 und nach Commit ist ODR-E7001, es entsteht
// kein Snapshot, und der Blob bleibt ohne Referenz (nach Commit) bzw. entsteht nicht (nach N1); Abort wird nie gerufen.
func TestCreateGegenLoeschen(t *testing.T) {
	for _, punkt := range []string{"Workspaces.Get", "Writer.Commit"} {
		w := neu(t)
		w.p.nach = func(op string) {
			if op == punkt {
				w.p.nach = nil
				w.loescheWorkspace()
			}
		}
		_, err := w.d.Create(context.Background(), "ws-1", "n")
		pruefeCode(t, err, "ODR-E7001", outbound.ErrNichtGefunden)
		blobs := len(w.blobs.inhalt)
		wantBlobs := map[string]int{"Workspaces.Get": 0, "Writer.Commit": 1}[punkt]
		if len(w.snaps.alle()) != 0 || blobs != wantBlobs || w.p.anzahl("Writer.Abort") != 0 {
			t.Errorf("%s: Snapshots %d, Blobs %d, Abort %d", punkt, len(w.snaps.alle()), blobs, w.p.anzahl("Writer.Abort"))
		}
	}
}

// zweig ist ein Zweig der Z-Liste von Create und List: der Aufbau, die Operation und die erwartete Antwort.
type zweig struct {
	id   string
	pre  func(w *welt, abbruch context.CancelFunc)
	op   string // "Create" oder "List"
	ws   string
	name string
	soll string
}

func zweigeSnapshot() []zweig {
	inj := func(op string, err error) func(w *welt, a context.CancelFunc) {
		return func(w *welt, _ context.CancelFunc) { w.p.setzeFehler(op, err) }
	}
	return []zweig{
		{"Z01", nil, "Create", "ws-1", "", "ungueltig"},
		{"Z02", inj("Workspaces.Get", errPlatte), "Create", "ws-1", "n", "ODR-E9001"},
		{"Z03", inj("Snapshots.Get", errPlatte), "Create", "ws-1", "n", "ODR-E9001"},
		{"Z04", inj("Entities.List", errPlatte), "Create", "ws-1", "n", "ODR-E9001"},
		{"Z05", inj("Blobs.Begin", errPlatte), "Create", "ws-1", "n", "ODR-E9001"},
		{"Z06", inj("Writer.Write", errPlatte), "Create", "ws-1", "n", "ODR-E9001"},
		{"Z07", inj("Writer.Commit", errPlatte), "Create", "ws-1", "n", "ODR-E9001"},
		{"Z08", func(w *welt, _ context.CancelFunc) {
			w.p.setzeFehler("Writer.Write", errPlatte)
			w.blobs.abortFehler = errors.New("Abort scheitert")
		}, "Create", "ws-1", "n", "ODR-E9001"},
		{"Z09", func(w *welt, _ context.CancelFunc) { _, _, _ = w.erzeuge("n") }, "Create", "ws-1", "n", "ODR-E7006"},
		{"Z10", func(w *welt, _ context.CancelFunc) {
			w.p.nach = func(op string) {
				if op == "Writer.Commit" {
					w.p.nach = nil
					w.loescheWorkspace()
				}
			}
		}, "Create", "ws-1", "n", "ODR-E7001"},
		{"Z11", func(w *welt, abbruch context.CancelFunc) {
			w.p.nach = func(op string) {
				if op == "Writer.Commit" {
					abbruch()
				}
			}
		}, "Create", "ws-1", "n", "kontext"},
		{"Z12", inj("Snapshots.CreateWorkspace", errPlatte), "Create", "ws-1", "n", "ODR-E9001"},
		{"Z13", nil, "List", "nope", "", "ODR-E7001"},
		{"Z14", inj("Snapshots.ListWorkspace", errPlatte), "List", "ws-1", "", "ODR-E9001"},
		{"Z15", func(w *welt, _ context.CancelFunc) { w.ent.setze("ws-1", []model.EntityState{}) }, "Create", "ws-1", "n", "ok"},
	}
}

func (z zweig) antwort(ctx context.Context, w *welt) string {
	var err error
	if z.op == "List" {
		_, err = w.d.List(ctx, z.ws)
	} else {
		_, err = w.d.Create(ctx, z.ws, z.name)
	}
	switch f, ok := codeVon(err); {
	case err == nil:
		return "ok"
	case ok:
		return f.Code.String()
	case errors.Is(err, context.Canceled):
		return "kontext"
	case errors.Is(err, model.ErrUngueltig):
		return "ungueltig"
	}
	return err.Error()
}

// TestZweigeSnapshot (T72): jeder Zweig der Z-Liste (Z01 bis Z15) läuft in vier Runden zu 64 Goroutinen gleichzeitig; die
// Antwort gleicht der des Einzelaufrufs und dem Literal der Liste. Die Welten entstehen vor dem Start.
func TestZweigeSnapshot(t *testing.T) {
	zs := zweigeSnapshot()
	if len(zs) != 15 {
		t.Fatalf("Zählprüfung: %d Zweige", len(zs))
	}
	neueWelt := func(z zweig) (*welt, context.Context) {
		w := neu(t)
		ctx, abbruch := context.WithCancel(context.Background())
		t.Cleanup(abbruch)
		if z.pre != nil {
			z.pre(w, abbruch)
		}
		return w, ctx
	}
	for _, z := range zs {
		w, ctx := neueWelt(z)
		einzel := z.antwort(ctx, w)
		if einzel != z.soll {
			t.Errorf("%s: Einzelaufruf %q, erwartet %q", z.id, einzel, z.soll)
		}
		for runde := 0; runde < 4; runde++ {
			welten := make([]*welt, 64)
			ctxs := make([]context.Context, 64)
			for i := range welten {
				welten[i], ctxs[i] = neueWelt(z)
			}
			antworten := make([]string, 64)
			gleichzeitig(64, func(i int) { antworten[i] = z.antwort(ctxs[i], welten[i]) })
			for i, a := range antworten {
				if a != einzel {
					t.Errorf("%s Runde %d Lauf %d: %q, Einzelaufruf %q", z.id, runde, i, a, einzel)
				}
			}
		}
	}
	// Z-Liste gegen die Quelle: dienst.go hat 13 Rückgaben mit Fehlerwert: 10 tragen die Zweige Z01 bis Z14 (Name; Get, Get,
	// List; Begin, Write, Commit; CreateWorkspace für Z09, Z10, Z12; ListWorkspace für Z13, Z14), 3 gehören zu New und zu
	// ctx nil (TestKonfigSnapshot, TestCtxNilSnapshot).
	// Die übrigen Quelldateien des Pakets: fehler.go bildet den Fehler (mitCode, ctxNil, nameUngueltig) und hat 4 Rückgaben
	// mit Fehlerwert, manifest.go schreibt Bytes (2 Rückgaben von Aufrufen, kein Fehlerwert; kein verstoss, kein errors).
	for datei, soll := range map[string]int{"dienst.go": 13, "fehler.go": 4, "manifest.go": 2} {
		if n := fehlerRueckgaben(t, datei); n != soll {
			t.Errorf("%d Rückgaben mit Fehlerwert in %s, erwartet %d", n, datei, soll)
		}
	}
}

// fehlerRueckgaben zählt in der Datei die return-Anweisungen, die einen Fehlerwert ungleich nil liefern (die Rückgabe `, nil`
// und `nil, ctxNil()`-freie Formen ausgenommen): jede Rückgabe, deren letztes Ergebnis ein Aufruf, ein Adresswert oder err ist.
func fehlerRueckgaben(t *testing.T, datei string) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ast.Inspect(f, func(x ast.Node) bool {
		if r, ok := x.(*ast.ReturnStmt); ok && len(r.Results) > 0 {
			switch l := r.Results[len(r.Results)-1].(type) {
			case *ast.CallExpr, *ast.UnaryExpr:
				n++
			case *ast.Ident:
				if l.Name == "err" {
					n++
				}
			}
		}
		return true
	})
	return n
}

// TestPaketOhneZustandSnapshot (T73): das Paket hält keine Sperre, keine Goroutine, keinen Zustand und keinen Typ mit
// Methoden außer Dienst; der Dienst hat als einziges Feld die Konfig; die Konfig wird nicht geschrieben. Der Quelltext
// (drei Dateien ohne _test.go) wird per go/parser gelesen.
func TestPaketOhneZustandSnapshot(t *testing.T) {
	erlaubt := map[string]bool{"bytes": true, "context": true, "errors": true, "math": true, "slices": true, "strconv": true, "unicode/utf8": true}
	pfad := "github.com/pt9912/odata-recorder/src/hexagon/"
	for _, p := range []string{"model", "ports/inbound", "ports/outbound", "ports/support"} {
		erlaubt[pfad+p] = true
	}
	dateien := []string{"dienst.go", "fehler.go", "manifest.go"}
	fset := token.NewFileSet()
	for _, d := range dateien {
		f, err := parser.ParseFile(fset, d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			if p, _ := strconv.Unquote(imp.Path.Value); !erlaubt[p] {
				t.Errorf("%s: Import %s", d, p)
			}
		}
		for _, decl := range f.Decls {
			switch x := decl.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "init" {
					t.Errorf("%s: init", d)
				}
				if x.Recv != nil {
					typ := x.Recv.List[0].Type
					if s, ok := typ.(*ast.StarExpr); ok {
						typ = s.X
					}
					if id, ok := typ.(*ast.Ident); !ok || id.Name != "Dienst" {
						t.Errorf("%s: Methode %s an anderem Typ", d, x.Name.Name)
					}
				}
			case *ast.GenDecl:
				for _, s := range x.Specs {
					switch v := s.(type) {
					case *ast.TypeSpec:
						if v.Name.Name != "Dienst" && v.Name.Name != "Konfig" {
							t.Errorf("%s: Typ %s", d, v.Name.Name)
						}
						if v.Name.Name == "Dienst" {
							st := v.Type.(*ast.StructType)
							if len(st.Fields.List) != 1 || st.Fields.List[0].Names[0].Name != "k" {
								t.Errorf("Dienst hat andere Felder als k")
							}
						}
					case *ast.ValueSpec:
						if x.Tok == token.VAR && (len(v.Names) != 1 || v.Names[0].Name != "_") {
							t.Errorf("%s: Variable %s", d, v.Names[0].Name)
						}
					}
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.GoStmt:
				t.Errorf("%s: go", d)
			case *ast.AssignStmt:
				for _, l := range x.Lhs {
					if s := strings.ReplaceAll(exprText(l), " ", ""); strings.HasPrefix(s, "d.k") {
						t.Errorf("%s: Zuweisung an %s", d, s)
					}
				}
			}
			return true
		})
	}
	if len(dateien) != 3 || !slices.Contains(dateien, "dienst.go") {
		t.Error("Zählprüfung")
	}
}

func exprText(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.SelectorExpr:
		return exprText(x.X) + "." + x.Sel.Name
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + exprText(x.X)
	case *ast.IndexExpr:
		return exprText(x.X) + "[]"
	}
	return "?"
}

package outboundtest

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Art nennt, welche Art Statement die Operation als erste führt.
type Art string

// Die Arten der Statements einer Operation: schreibend, Liste, einzelne Zeile.
const (
	ArtExec     Art = "Exec"
	ArtQuery    Art = "Query"
	ArtQueryRow Art = "QueryRow"
)

// Operation ist eine Operation der vier Repositories von WorkspaceStores mit dem Aufruf, der
// im Zustand von OperationenBasis gelingt.
type Operation struct {
	Name       string
	Art        Art
	Schreibend bool
	Lauf       func(context.Context, outbound.WorkspaceStores) error
	// Fehler ist der Fehler der Datenbank, den LaufFehler auslöst; nil, wenn die Operation keinen hat.
	Fehler     error
	LaufFehler func(context.Context, outbound.WorkspaceStores) error
}

// Gesamtzustand fasst zusammen, was die Operationen ändern können.
type Gesamtzustand struct {
	Workspaces []model.Workspace
	Entities   []model.EntityState
	Cursors    []model.ReplayCursor
	Snapshots  []model.WorkspaceSnapshot
}

// LiesGesamtzustand liest den Zustand der Workspaces des Service d und die Einträge,
// Cursors und Snapshots des Workspace w; ein gelöschter Workspace hat keine.
func LiesGesamtzustand(t *testing.T, v Verbindung) Gesamtzustand {
	t.Helper()
	ctx := context.Background()
	var z Gesamtzustand
	var err error
	z.Workspaces, err = v.Workspaces.List(ctx, "d")
	ok(t, err)
	if _, err = v.Workspaces.Get(ctx, "w"); errors.Is(err, outbound.ErrNichtGefunden) {
		return z
	}
	ok(t, err)
	z.Entities, err = v.Entities.List(ctx, "w")
	ok(t, err)
	z.Cursors, err = v.Cursors.List(ctx, "w")
	ok(t, err)
	z.Snapshots, err = v.Snapshots.ListWorkspace(ctx, "w")
	ok(t, err)
	return z
}

// OperationenBasis liefert die Verbindung mit dem Workspace w (Revision 0), einem Eintrag
// Products/alt, dem Cursor r/s1 und dem Snapshot sn.
func OperationenBasis(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w")
	ok(t, v.Entities.Put(ctx, entity("w", "Products", "alt")))
	ok(t, v.Cursors.Put(ctx, cursor("w", "r", "s1")))
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn", "d", "w", "n", 0)))
	return v
}

func ohneWert[T any](f func(context.Context, outbound.WorkspaceStores) (T, error)) func(context.Context, outbound.WorkspaceStores) error {
	return func(ctx context.Context, s outbound.WorkspaceStores) error {
		_, err := f(ctx, s)
		return err
	}
}

// Operationen sind die 19 Operationen der vier Repositories: 8 schreibend, 6 Listen, 5 einzelne Zeilen.
func Operationen() []Operation {
	return append(append(workspaceOperationen(), entityOperationen()...), cursorUndSnapshotOperationen()...)
}

func workspaceOperationen() []Operation {
	return []Operation{
		{"Workspaces.Create", ArtExec, true,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Workspaces.Create(ctx, arbeitsraum("wn", 5))
			},
			outbound.ErrBereitsVorhanden,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Workspaces.Create(ctx, arbeitsraum("w", 5))
			}},
		{"Workspaces.Get", ArtQueryRow, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.Workspace, error) {
				return s.Workspaces.Get(ctx, "w")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.Workspace, error) {
				return s.Workspaces.Get(ctx, "x")
			})},
		{"Workspaces.List", ArtQuery, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.Workspace, error) {
				return s.Workspaces.List(ctx, "d")
			}),
			nil, nil},
		{"Workspaces.Advance", ArtExec, true,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (int64, error) {
				return s.Workspaces.Advance(ctx, "w", 0)
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (int64, error) {
				return s.Workspaces.Advance(ctx, "x", 0)
			})},
		{"Workspaces.Delete", ArtExec, true,
			func(ctx context.Context, s outbound.WorkspaceStores) error { return s.Workspaces.Delete(ctx, "w") },
			outbound.ErrNichtGefunden,
			func(ctx context.Context, s outbound.WorkspaceStores) error { return s.Workspaces.Delete(ctx, "x") }},
	}
}

func entityOperationen() []Operation {
	neu := func(ws string) []model.EntityState { return []model.EntityState{entity(ws, "Products", "neu")} }
	return []Operation{
		{"Entities.Put", ArtExec, true,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Entities.Put(ctx, entity("w", "Products", "neu"))
			},
			outbound.ErrNichtGefunden,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Entities.Put(ctx, entity("x", "Products", "neu"))
			}},
		{"Entities.Get", ArtQueryRow, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.EntityState, error) {
				return s.Entities.Get(ctx, "w", "Products", "alt")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.EntityState, error) {
				return s.Entities.Get(ctx, "x", "Products", "alt")
			})},
		{"Entities.List", ArtQuery, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.EntityState, error) {
				return s.Entities.List(ctx, "w")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.EntityState, error) {
				return s.Entities.List(ctx, "x")
			})},
		{"Entities.ListSet", ArtQuery, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.EntityState, error) {
				return s.Entities.ListSet(ctx, "w", "Products")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.EntityState, error) {
				return s.Entities.ListSet(ctx, "x", "Products")
			})},
		{"Entities.ReplaceAll", ArtExec, true,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Entities.ReplaceAll(ctx, "w", neu("w"))
			},
			outbound.ErrNichtGefunden,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Entities.ReplaceAll(ctx, "x", neu("x"))
			}},
	}
}

func cursorUndSnapshotOperationen() []Operation {
	return []Operation{
		{"Cursors.Put", ArtExec, true,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Cursors.Put(ctx, cursor("w", "r", "s2"))
			},
			outbound.ErrNichtGefunden,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Cursors.Put(ctx, cursor("x", "r", "s2"))
			}},
		{"Cursors.Get", ArtQueryRow, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.ReplayCursor, error) {
				return s.Cursors.Get(ctx, "w", "r", "s1")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.ReplayCursor, error) {
				return s.Cursors.Get(ctx, "w", "r", "s9")
			})},
		{"Cursors.List", ArtQuery, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.ReplayCursor, error) {
				return s.Cursors.List(ctx, "w")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.ReplayCursor, error) {
				return s.Cursors.List(ctx, "x")
			})},
		{"Cursors.ResetAll", ArtExec, true,
			func(ctx context.Context, s outbound.WorkspaceStores) error { return s.Cursors.ResetAll(ctx, "w") },
			outbound.ErrNichtGefunden,
			func(ctx context.Context, s outbound.WorkspaceStores) error { return s.Cursors.ResetAll(ctx, "x") }},
		{"Snapshots.Get", ArtQueryRow, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.BaselineSnapshot, error) {
				return s.Snapshots.Get(ctx, "snap-1")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.BaselineSnapshot, error) {
				return s.Snapshots.Get(ctx, "x")
			})},
		{"Snapshots.List", ArtQuery, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.BaselineSnapshot, error) {
				return s.Snapshots.List(ctx, "d")
			}),
			nil, nil},
		{"Snapshots.CreateWorkspace", ArtExec, true,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn2", "d", "w", "n2", 1))
			},
			outbound.ErrBereitsVorhanden,
			func(ctx context.Context, s outbound.WorkspaceStores) error {
				return s.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn", "d", "w", "n2", 1))
			}},
		{"Snapshots.GetWorkspace", ArtQueryRow, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.WorkspaceSnapshot, error) {
				return s.Snapshots.GetWorkspace(ctx, "w", "sn")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) (model.WorkspaceSnapshot, error) {
				return s.Snapshots.GetWorkspace(ctx, "w", "x")
			})},
		{"Snapshots.ListWorkspace", ArtQuery, false,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.WorkspaceSnapshot, error) {
				return s.Snapshots.ListWorkspace(ctx, "w")
			}),
			outbound.ErrNichtGefunden,
			ohneWert(func(ctx context.Context, s outbound.WorkspaceStores) ([]model.WorkspaceSnapshot, error) {
				return s.Snapshots.ListWorkspace(ctx, "x")
			})},
	}
}

// KontextArt ist eine Art des Kontextfehlers (SPEC-108 Nr. 1 a): neu liefert einen lebenden
// Kontext und die Funktion, die ihn beendet; Soll ist der Fehler, den er nach dem Ende trägt.
type KontextArt struct {
	Name string
	Soll error
	neu  func() (context.Context, func())
}

// KontextArten sind die drei Arten des Kontextfehlers: Abbruch, Frist, Abbruch mit Ursache.
func KontextArten() []KontextArt {
	return []KontextArt{
		{"Canceled", context.Canceled, func() (context.Context, func()) {
			ctx, stop := context.WithCancel(context.Background())
			return ctx, stop
		}},
		{"DeadlineExceeded", context.DeadlineExceeded, func() (context.Context, func()) {
			ctx, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
			return ctx, func() { <-ctx.Done(); stop() }
		}},
		{"Canceled mit Ursache", context.Canceled, func() (context.Context, func()) {
			ctx, stop := context.WithCancelCause(context.Background())
			return ctx, func() { stop(errors.New("Ursache")) }
		}},
	}
}

// Beendet liefert den Kontext der Art im Zustand beendet.
func (a KontextArt) Beendet() context.Context {
	ctx, beende := a.neu()
	beende()
	return ctx
}

// UngueltigeAufrufe sind je schreibender Operation, die ein Argument vor ihrem ersten Statement
// als ungültig erkennt, ein Aufruf mit einem solchen Argument (Ergebnis ErrUngueltig, bevor die
// Datenbank gefragt wird; SPEC-108 Nr. 1 f, Reihenfolge der Prüfungen). Die übrigen schreibenden
// Operationen (Workspaces.Delete, Cursors.ResetAll) nehmen nur eine Kennung entgegen und
// erkennen vor dem Savepoint keinen ungültigen Wert: eine unbekannte Kennung ist ein Fehler der
// Datenbank (ErrNichtGefunden), kein ErrUngueltig.
func UngueltigeAufrufe() map[string]func(context.Context, outbound.WorkspaceStores) error {
	return map[string]func(context.Context, outbound.WorkspaceStores) error{
		"Workspaces.Create": func(ctx context.Context, s outbound.WorkspaceStores) error {
			return s.Workspaces.Create(ctx, arbeitsraum("", 5))
		},
		"Workspaces.Advance": func(ctx context.Context, s outbound.WorkspaceStores) error {
			_, err := s.Workspaces.Advance(ctx, "w", -1)
			return err
		},
		"Entities.Put": func(ctx context.Context, s outbound.WorkspaceStores) error {
			return s.Entities.Put(ctx, entity("", "Products", "neu"))
		},
		"Entities.ReplaceAll": func(ctx context.Context, s outbound.WorkspaceStores) error {
			return s.Entities.ReplaceAll(ctx, "w", []model.EntityState{entity("x", "Products", "neu")})
		},
		"Cursors.Put": func(ctx context.Context, s outbound.WorkspaceStores) error {
			return s.Cursors.Put(ctx, cursor("", "r", "s2"))
		},
		"Snapshots.CreateWorkspace": func(ctx context.Context, s outbound.WorkspaceStores) error {
			return s.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("", "d", "w", "n2", 1))
		},
	}
}

// OhneWorkspace nimmt den Workspace id aus dem Zustand.
func OhneWorkspace(z Gesamtzustand, id string) Gesamtzustand {
	var rest []model.Workspace
	for _, w := range z.Workspaces {
		if w.ID != id {
			rest = append(rest, w)
		}
	}
	z.Workspaces = rest
	return z
}

// inTransaktionMitW2 führt f in einer Transaktion aus, in der zuvor Workspaces.Create für w2
// gelang, und prüft, dass die Transaktion bestätigte, w2 also erhalten ist.
func inTransaktionMitW2(t *testing.T, v Verbindung, f func(outbound.WorkspaceStores)) {
	t.Helper()
	bg := context.Background()
	ok(t, v.Transaktion.WithWorkspaceTransaction(bg, func(s outbound.WorkspaceStores) error {
		ok(t, s.Workspaces.Create(bg, arbeitsraum("w2", 9)))
		f(s)
		return nil
	}))
	if _, err := v.Workspaces.Get(bg, "w2"); err != nil {
		t.Fatalf("die gelungene Operation derselben Transaktion ging verloren: %v", err)
	}
}

func gleichGesamt(t *testing.T, was string, got, soll Gesamtzustand) {
	t.Helper()
	if !reflect.DeepEqual(got, soll) {
		t.Fatalf("%s:\n%+v\n%+v", was, got, soll)
	}
}

// transaktionKontextVorOperation prüft SPEC-108 Nr. 1 a für jede Operation und jede Art des
// Kontextfehlers: der Fehler des Kontexts, keine Änderung.
func transaktionKontextVorOperation(t *testing.T, neu Neu) {
	for _, op := range Operationen() {
		t.Run(op.Name, func(t *testing.T) {
			v := OperationenBasis(t, neu)
			vorher := LiesGesamtzustand(t, v)
			inTransaktionMitW2(t, v, func(s outbound.WorkspaceStores) {
				for _, a := range KontextArten() {
					if err := op.Lauf(a.Beendet(), s); !errors.Is(err, a.Soll) {
						t.Errorf("%s: Fehler %v, erwartet %v", a.Name, err, a.Soll)
					}
				}
			})
			gleichGesamt(t, "Zustand nach beendetem Kontext", OhneWorkspace(LiesGesamtzustand(t, v), "w2"), vorher)
		})
	}
}

// transaktionKontextVorDatenbankfehler prüft den Vorrang aus SPEC-108 Nr. 1 a: ein beendeter
// Kontext geht dem Fehler der Datenbank vor, den derselbe Aufruf mit lebendem Kontext liefert.
func transaktionKontextVorDatenbankfehler(t *testing.T, neu Neu) {
	n := 0
	for _, op := range Operationen() {
		if op.Fehler == nil {
			continue
		}
		n++
		t.Run(op.Name, func(t *testing.T) {
			v := OperationenBasis(t, neu)
			inTransaktionMitW2(t, v, func(s outbound.WorkspaceStores) {
				if err := op.LaufFehler(context.Background(), s); !errors.Is(err, op.Fehler) {
					t.Fatalf("lebender Kontext: Fehler %v, erwartet %v", err, op.Fehler)
				}
				for _, a := range KontextArten() {
					err := op.LaufFehler(a.Beendet(), s)
					if !errors.Is(err, a.Soll) || errors.Is(err, op.Fehler) {
						t.Errorf("%s: Fehler %v, erwartet %v vor %v", a.Name, err, a.Soll, op.Fehler)
					}
				}
			})
		})
	}
	if n != 17 {
		t.Fatalf("%d Operationen mit Fehler der Datenbank, erwartet 17", n)
	}
}

// transaktionOperationenMenge hält fest, dass Operationen() die Methoden der vier Repositories
// von WorkspaceStores sind und nur diese.
func transaktionOperationenMenge(t *testing.T, _ Neu) {
	soll := map[string]bool{}
	ty := reflect.TypeOf(outbound.WorkspaceStores{})
	for i := range ty.NumField() {
		f := ty.Field(i)
		for j := range f.Type.NumMethod() {
			soll[f.Name+"."+f.Type.Method(j).Name] = true
		}
	}
	art := map[Art]int{}
	schreibend := 0
	for _, op := range Operationen() {
		if !soll[op.Name] {
			t.Errorf("Operation %s gehört keinem der Repositories", op.Name)
		}
		delete(soll, op.Name)
		art[op.Art]++
		if op.Schreibend {
			schreibend++
		}
	}
	if len(soll) != 0 {
		t.Errorf("Methoden ohne Eintrag in Operationen(): %v", soll)
	}
	if art[ArtExec] != 8 || art[ArtQuery] != 6 || art[ArtQueryRow] != 5 || schreibend != 8 {
		t.Errorf("Arten %v, schreibend %d, erwartet 8 Exec, 6 Query, 5 QueryRow", art, schreibend)
	}
}

// transaktionVerschluckteFehler: zwei verschluckte Fehler derselben Transaktion (Kontext,
// Datenbank) ändern nichts; die gelungene Operation bleibt (SPEC-108 Nr. 1 b, d).
func transaktionVerschluckteFehler(t *testing.T, neu Neu) {
	v := OperationenBasis(t, neu)
	vorher := LiesGesamtzustand(t, v)
	inTransaktionMitW2(t, v, func(s outbound.WorkspaceStores) {
		erwarte(t, s.Workspaces.Delete(abgebrochen(), "w"), context.Canceled)
		erwarte(t, s.Workspaces.Create(context.Background(), arbeitsraum("w", 3)), outbound.ErrBereitsVorhanden)
	})
	gleichGesamt(t, "Zustand nach zwei verschluckten Fehlern", OhneWorkspace(LiesGesamtzustand(t, v), "w2"), vorher)
}

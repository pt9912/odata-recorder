package sqlite_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

// haltepunkt ist ein Kontext, dessen Err-Aufrufe gezählt werden und dessen bei-ter Aufruf ihn
// beendet: danach liefert Err nil und beendet ihn erst nach der Antwort (der Abbruch trifft den
// Schritt nach seiner Prüfung), sonst beendet er vor der Antwort. Der Test legt so den Abbruch
// auf einen Schritt fest, ohne Zeitfenster und ohne Goroutine. vorher läuft im bei-ten Aufruf,
// bevor Err antwortet, mit den Stores der Transaktion (inTransaktion setzt sie). Done ist offen,
// bis der Kontext endet; ein Statement mit diesem Kontext ruft also den Interrupt des Treibers.
type haltepunkt struct {
	bei    int64
	danach bool
	vorher func(outbound.WorkspaceStores)
	stores outbound.WorkspaceStores
	n      atomic.Int64
	ende   atomic.Bool
	done   chan struct{}
}

func neuerHaltepunkt(bei int64, danach bool, vorher func(outbound.WorkspaceStores)) *haltepunkt {
	return &haltepunkt{bei: bei, danach: danach, vorher: vorher, done: make(chan struct{})}
}

func (h *haltepunkt) beende() {
	if h.ende.CompareAndSwap(false, true) {
		close(h.done)
	}
}

func (h *haltepunkt) Deadline() (time.Time, bool) { return time.Time{}, false }
func (h *haltepunkt) Done() <-chan struct{}       { return h.done }
func (h *haltepunkt) Value(any) any               { return nil }

func (h *haltepunkt) Err() error {
	if h.n.Add(1) == h.bei {
		if h.vorher != nil {
			h.vorher(h.stores)
		}
		if h.danach {
			h.beende()
			return nil
		}
		h.beende()
	}
	if h.ende.Load() {
		return context.Canceled
	}
	return nil
}

// aufrufe ist die Zahl der Err-Aufrufe bisher.
func (h *haltepunkt) aufrufe() int64 { return h.n.Load() }

// pruefungsZahl ist je Operation die Zahl der Prüfungen des Kontexts in einem gelungenen Lauf
// im Zustand von OperationenBasis: ein Statement des Adapters je Prüfung, der Savepoint der
// schreibenden Operation eingeschlossen. Die Zahl gehört zum Adapter, nicht zum Vertrag; der
// Probelauf in haltepunktPosition und haltepunktLesen bestätigt sie.
func pruefungsZahl() map[string]int {
	return map[string]int{
		"Workspaces.Create": 5, "Workspaces.Get": 1, "Workspaces.List": 1, "Workspaces.Advance": 3, "Workspaces.Delete": 4,
		"Entities.Put": 3, "Entities.Get": 1, "Entities.List": 1, "Entities.ListSet": 1, "Entities.ReplaceAll": 4,
		"Cursors.Put": 4, "Cursors.Get": 1, "Cursors.List": 1, "Cursors.ResetAll": 3,
		"Snapshots.Get": 1, "Snapshots.List": 1, "Snapshots.CreateWorkspace": 5, "Snapshots.GetWorkspace": 1, "Snapshots.ListWorkspace": 1,
	}
}

func pruefungen(t *testing.T, op outboundtest.Operation) int {
	t.Helper()
	zahlen := pruefungsZahl()
	n, ok := zahlen[op.Name]
	if !ok || len(zahlen) != 19 {
		t.Fatalf("Zahl der Prüfungen von %s unbekannt (%d Einträge, erwartet 19)", op.Name, len(zahlen))
	}
	return n
}

func operationen(t *testing.T, filter func(outboundtest.Operation) bool) []outboundtest.Operation {
	t.Helper()
	var out []outboundtest.Operation
	for _, op := range outboundtest.Operationen() {
		if filter(op) {
			out = append(out, op)
		}
	}
	return out
}

func operation(t *testing.T, name string) outboundtest.Operation {
	t.Helper()
	for _, op := range outboundtest.Operationen() {
		if op.Name == name {
			return op
		}
	}
	t.Fatalf("keine Operation %s", name)
	return outboundtest.Operation{}
}

// laufeMitW2 führt op in einer Transaktion aus, in der w2 vorher gelang, mit dem Kontext ctx;
// f läuft nach der Operation. Es liefert den Fehler der Operation; die Transaktion bestätigt.
func laufeMitW2(ctx context.Context, t *testing.T, v outboundtest.Verbindung, op outboundtest.Operation) error {
	t.Helper()
	bg := context.WithoutCancel(ctx)
	var opErr error
	err := v.Transaktion.WithWorkspaceTransaction(bg, func(s outbound.WorkspaceStores) error {
		w2 := model.Workspace{ID: "w2", ServiceID: "d", BaselineSnapshotID: "snap-1", CreatedAt: time.Date(2026, 10, 19, 10, 9, 0, 0, time.UTC)}
		if err := s.Workspaces.Create(bg, w2); err != nil {
			t.Fatal(err)
		}
		opErr = op.Lauf(ctx, s)
		return nil
	})
	if err != nil {
		t.Fatalf("%s: die Transaktion bestätigte nicht: %v", op.Name, err)
	}
	return opErr
}

// TestTransaktionAbbruchHaltepunkt fängt den Abbruch des Kontexts an einem festen Schritt:
// nach der Prüfung eines Statements einer Liste, einer einzelnen Zeile und eines schreibenden
// Statements an jeder Position einer Operation, und vor wie nach der Prüfung beim Bestätigen.
// Die Haltepunkte ersetzen das Zeitfenster der Reihenfolge Interrupt zuerst, das
// TestTransaktionAbbruchImStatement offen lässt; es läuft keine Goroutine, die Last des Hosts
// entscheidet nichts. Offen bleibt, was der Haltepunkt nicht abbildet: ein Abbruch durch einen
// anderen Auslöser als den Err-Aufruf (etwa ein Timer) trifft denselben Schritt, ohne dass ein
// Test ihn festlegt.
func TestTransaktionAbbruchHaltepunkt(t *testing.T) {
	for _, art := range []outboundtest.Art{outboundtest.ArtQuery, outboundtest.ArtQueryRow} {
		t.Run(strings.ToLower(string(art)), func(t *testing.T) { haltepunktLesen(t, art) })
	}
	t.Run("position", haltepunktPosition)
	t.Run("bestaetigen-vorher", func(t *testing.T) { haltepunktBestaetigen(t, false) })
	t.Run("bestaetigen-nachher", func(t *testing.T) { haltepunktBestaetigen(t, true) })
}

// haltepunktLesen: der Abbruch nach der Prüfung des Statements ändert eine lesende Operation
// nicht (SPEC-108 Nr. 1 c); jede Liste und jede einzelne Zeile ist ein Element der Schleife.
func haltepunktLesen(t *testing.T, art outboundtest.Art) {
	ops := operationen(t, func(op outboundtest.Operation) bool { return op.Art == art })
	if len(ops) != map[outboundtest.Art]int{outboundtest.ArtQuery: 6, outboundtest.ArtQueryRow: 5}[art] {
		t.Fatalf("%d Operationen der Art %s", len(ops), art)
	}
	for _, op := range ops {
		t.Run(op.Name, func(t *testing.T) {
			v := outboundtest.OperationenBasis(t, neueVerbindung)
			h := neuerHaltepunkt(1, true, nil)
			if err := laufeMitW2(h, t, v, op); err != nil {
				t.Fatalf("Fehler %v nach dem Abbruch im Statement", err)
			}
			if h.aufrufe() != int64(pruefungen(t, op)) || !h.ende.Load() {
				t.Fatalf("%d Prüfungen (erwartet %d), Kontext beendet: %v", h.aufrufe(), pruefungen(t, op), h.ende.Load())
			}
		})
	}
}

// haltepunktPosition: der Abbruch nach der Prüfung des k-ten Statements einer schreibenden
// Operation, für jedes k von 1 bis zur Zahl ihrer Prüfungen (SPEC-108 Nr. 1 b und c): vor dem
// letzten Statement Fehler des Kontexts und keine Änderung, beim letzten gelungen samt Wirkung.
func haltepunktPosition(t *testing.T) {
	ops := operationen(t, func(op outboundtest.Operation) bool { return op.Schreibend })
	if len(ops) != 8 {
		t.Fatalf("%d schreibende Operationen, erwartet 8", len(ops))
	}
	for _, op := range ops {
		t.Run(op.Name, func(t *testing.T) {
			// Probelauf: die Zahl der Prüfungen einer gelungenen Operation.
			v := outboundtest.OperationenBasis(t, neueVerbindung)
			probe := neuerHaltepunkt(0, false, nil)
			n := pruefungen(t, op)
			if err := laufeMitW2(probe, t, v, op); err != nil || probe.aufrufe() != int64(n) {
				t.Fatalf("Probelauf: Fehler %v, %d Prüfungen, erwartet %d", err, probe.aufrufe(), n)
			}
			for k := 1; k <= n; k++ {
				v := outboundtest.OperationenBasis(t, neueVerbindung)
				vorher := outboundtest.LiesGesamtzustand(t, v)
				err := laufeMitW2(neuerHaltepunkt(int64(k), true, nil), t, v, op)
				nachher := outboundtest.OhneWorkspace(outboundtest.LiesGesamtzustand(t, v), "w2")
				switch {
				case k < n && !errors.Is(err, context.Canceled):
					t.Errorf("k=%d: Fehler %v, erwartet context.Canceled", k, err)
				case k < n && !reflect.DeepEqual(nachher, vorher):
					t.Errorf("k=%d: Zustand nach dem Fehler geändert:\n%+v\n%+v", k, nachher, vorher)
				case k == n && err != nil:
					t.Errorf("k=%d (letztes Statement): Fehler %v, erwartet gelungen", k, err)
				case k == n && reflect.DeepEqual(nachher, vorher):
					t.Errorf("k=%d (letztes Statement): gelungen ohne Wirkung", k)
				}
			}
		})
	}
}

// haltepunktBestaetigen: der Kontext der Transaktion endet beim Err-Aufruf der Prüfung vor
// COMMIT (SPEC-108 Nr. 1 d). Endet er vor der Antwort, wird zurückgerollt und sein Fehler
// geliefert; endet er danach, wird bestätigt und nil geliefert.
func haltepunktBestaetigen(t *testing.T, danach bool) {
	bg := context.Background()
	v := outboundtest.OperationenBasis(t, neueVerbindung)
	laufe := func(ctx context.Context, id string) error {
		return v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
			w := model.Workspace{ID: id, ServiceID: "d", BaselineSnapshotID: "snap-1", CreatedAt: time.Date(2026, 10, 19, 10, 9, 0, 0, time.UTC)}
			return s.Workspaces.Create(bg, w)
		})
	}
	// Probelauf: die Zahl der Err-Aufrufe der Transaktion; der letzte ist die Prüfung vor COMMIT.
	probe := neuerHaltepunkt(0, false, nil)
	if err := laufe(probe, "wp"); err != nil || probe.aufrufe() < 1 {
		t.Fatalf("Probelauf: Fehler %v, %d Err-Aufrufe", err, probe.aufrufe())
	}
	err := laufe(neuerHaltepunkt(probe.aufrufe(), danach, nil), "wh")
	_, errGet := v.Workspaces.Get(bg, "wh")
	if danach && (err != nil || errGet != nil) {
		t.Fatalf("Abbruch nach der Prüfung: Fehler %v, Get %v, erwartet bestätigt", err, errGet)
	}
	if !danach && (!errors.Is(err, context.Canceled) || !errors.Is(errGet, outbound.ErrNichtGefunden)) {
		t.Fatalf("Abbruch vor der Prüfung: Fehler %v, Get %v, erwartet zurückgerollt mit context.Canceled", err, errGet)
	}
	if err := laufe(bg, "wn"); err != nil {
		t.Fatalf("die nächste Transaktion ist nicht nutzbar: %v", err)
	}
}

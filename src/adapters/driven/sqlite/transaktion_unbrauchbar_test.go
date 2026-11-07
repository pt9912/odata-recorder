package sqlite_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

// Scheitern das Zurücknehmen einer Operation oder der Abschluss ihres Savepoints, ist die
// Transaktion unbrauchbar (SPEC-108 Nr. 1 f). Die Tests lösen den Fehler der Datenbank aus,
// indem sie zwischen zwei Statements der Operation den Savepoint aufheben: der Eingriff eines
// Haltepunkts führt RELEASE operation auf der Verbindung aus.

const keinSavepoint = "no such savepoint"

// entityPut ist die Operation Entities.Put: Savepoint, Prüfung des Workspace und Eintrag sind
// ihre drei Statements, jedes mit einer Prüfung des Kontexts.
func entityPut(t *testing.T) outboundtest.Operation { return operation(t, "Entities.Put") }

// hebeSavepointAuf ist der Eingriff eines Haltepunkts: er hebt den Savepoint der Operation auf
// der Verbindung der Stores auf. Gelingt er nicht, bleibt der Savepoint, und die Tests melden,
// dass kein Fehler kam.
func hebeSavepointAuf(s outbound.WorkspaceStores) {
	_ = sqlite.StatementAufVerbindung(context.Background(), s, "RELEASE operation")
}

// inTransaktion führt f in einer Transaktion aus, die ohne den Abbruch von ctx läuft; f erhält
// ctx als Kontext der Operationen. Ist ctx ein Haltepunkt, kennt er danach die Stores der
// Transaktion für seinen Eingriff.
func inTransaktion(ctx context.Context, v outboundtest.Verbindung, f func(context.Context, outbound.WorkspaceStores) error) error {
	return v.Transaktion.WithWorkspaceTransaction(context.WithoutCancel(ctx), func(s outbound.WorkspaceStores) error {
		if h, ok := ctx.(*haltepunkt); ok {
			h.stores = s
		}
		return f(ctx, s)
	})
}

func w2(id string) model.Workspace {
	return model.Workspace{ID: id, ServiceID: "d", BaselineSnapshotID: "snap-1", CreatedAt: time.Date(2026, 10, 19, 10, 9, 0, 0, time.UTC)}
}

func TestTransaktionUnbrauchbar(t *testing.T) {
	t.Run("operation", unbrauchbarOperation)
	t.Run("weitere", unbrauchbarWeitere)
	t.Run("transaktion", unbrauchbarTransaktion)
	t.Run("release", unbrauchbarRelease)
}

// unbrauchbarOperation: scheitert ROLLBACK TO, liefert die Operation diesen Fehler und nicht den
// Fehler des Kontexts, der sie scheitern ließ.
func unbrauchbarOperation(t *testing.T) {
	v := outboundtest.OperationenBasis(t, neueVerbindung)
	var opErr error
	_ = inTransaktion(neuerHaltepunkt(2, false, hebeSavepointAuf), v, func(ctx context.Context, s outbound.WorkspaceStores) error {
		opErr = entityPut(t).Lauf(ctx, s)
		return nil
	})
	if opErr == nil || !strings.Contains(opErr.Error(), keinSavepoint) || errors.Is(opErr, context.Canceled) {
		t.Fatalf("Fehler %v, erwartet der Fehler von ROLLBACK TO (%s) statt context.Canceled", opErr, keinSavepoint)
	}
}

// unbrauchbarWeitere: danach endet jede der 8 schreibenden Operationen mit demselben Fehler,
// jede der 11 lesenden liefert weiter.
func unbrauchbarWeitere(t *testing.T) {
	v := outboundtest.OperationenBasis(t, neueVerbindung)
	var schreibend, lesend int
	_ = inTransaktion(neuerHaltepunkt(2, false, hebeSavepointAuf), v, func(ctx context.Context, s outbound.WorkspaceStores) error {
		erster := entityPut(t).Lauf(ctx, s)
		bg := context.WithoutCancel(ctx)
		for _, op := range outboundtest.Operationen() {
			err := op.Lauf(bg, s)
			switch {
			case op.Schreibend:
				schreibend++
				if erster == nil || err != erster {
					t.Errorf("%s: Fehler %v, erwartet der Fehler der unbrauchbaren Transaktion %v", op.Name, err, erster)
				}
			default:
				lesend++
				if err != nil {
					t.Errorf("%s: lesende Operation lieferte %v", op.Name, err)
				}
			}
		}
		return nil
	})
	if schreibend != 8 || lesend != 11 {
		t.Fatalf("%d schreibende und %d lesende Operationen, erwartet 8 und 11", schreibend, lesend)
	}
}

// unbrauchbarTransaktion: die Transaktion liefert den Fehler und rollt zurück, auch wenn f nil
// liefert; ein Fehler von f gilt unverändert; die nächste Transaktion ist nutzbar.
func unbrauchbarTransaktion(t *testing.T) {
	bg := context.Background()
	v := outboundtest.OperationenBasis(t, neueVerbindung)
	vorher := outboundtest.LiesGesamtzustand(t, v)
	macheUnbrauchbar := func(ctx context.Context, s outbound.WorkspaceStores) {
		if err := s.Workspaces.Create(context.WithoutCancel(ctx), w2("w2")); err != nil {
			t.Fatal(err)
		}
		_ = entityPut(t).Lauf(ctx, s)
	}
	err := inTransaktion(neuerHaltepunkt(2, false, hebeSavepointAuf), v, func(ctx context.Context, s outbound.WorkspaceStores) error {
		macheUnbrauchbar(ctx, s)
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), keinSavepoint) {
		t.Fatalf("Fehler %v, erwartet der Fehler des Savepoints (%s)", err, keinSavepoint)
	}
	if got := outboundtest.LiesGesamtzustand(t, v); !reflect.DeepEqual(got, vorher) {
		t.Fatalf("die unbrauchbare Transaktion bestätigte:\n%+v\n%+v", got, vorher)
	}
	eigener := errors.New("eigener Fehler von f")
	err = inTransaktion(neuerHaltepunkt(2, false, hebeSavepointAuf), v, func(ctx context.Context, s outbound.WorkspaceStores) error {
		macheUnbrauchbar(ctx, s)
		return eigener
	})
	if err != eigener {
		t.Fatalf("Fehler %v, erwartet der Fehler von f unverändert", err)
	}
	err = v.Transaktion.WithWorkspaceTransaction(bg, func(s outbound.WorkspaceStores) error { return s.Workspaces.Create(bg, w2("w3")) })
	if err != nil {
		t.Fatalf("die nächste Transaktion ist nicht nutzbar: %v", err)
	}
}

// unbrauchbarRelease: scheitert RELEASE nach gelungener Operation, ist die Transaktion ebenso
// unbrauchbar; der Eingriff liegt nach der Prüfung des letzten Statements.
func unbrauchbarRelease(t *testing.T) {
	v := outboundtest.OperationenBasis(t, neueVerbindung)
	op := entityPut(t)
	var opErr error
	err := inTransaktion(neuerHaltepunkt(int64(op.Pruefungen), true, hebeSavepointAuf), v, func(ctx context.Context, s outbound.WorkspaceStores) error {
		opErr = op.Lauf(ctx, s)
		return nil
	})
	for _, e := range []error{opErr, err} {
		if e == nil || !strings.Contains(e.Error(), keinSavepoint) {
			t.Fatalf("Fehler der Operation %v, der Transaktion %v, erwartet je der Fehler von RELEASE (%s)", opErr, err, keinSavepoint)
		}
	}
}

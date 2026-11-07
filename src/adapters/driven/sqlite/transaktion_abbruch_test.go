package sqlite_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	sqlitedriver "modernc.org/sqlite"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Abbruch des Kontexts im Statement einer offenen Transaktion (SPEC-108 Nr. 1, Kontext in der
// Transaktion). Der Treiber ruft bei ctx.Done in einer eigenen Goroutine sqlite3_interrupt und
// meldet nicht, wann der Interrupt angewendet ist; SQLite rollt bei SQLITE_INTERRUPT in einem
// schreibenden Statement die ganze Transaktion zurück. Der Adapter führt die Statements der
// Transaktion ohne den Abbruch des Aufrufs aus, so dass die Reihenfolge von Statement und
// Interrupt nichts entscheidet.

const (
	modusZuerst = "interrupt-zuerst"
	modusDanach = "statement-zuerst"

	// Grenze des Wartens auf den Interrupt in modusZuerst: der Treiber liefert kein Signal.
	wartezeitInterrupt = 500 * time.Millisecond
)

// abbruchFunktion meldet eine SQL-Funktion unter eigenem Namen an; sie gilt nur für
// Verbindungen, die danach öffnen. Sie ruft stop und wartet im Modus modusZuerst, bis der
// Interrupt des Treibers angewendet ist.
func abbruchFunktion(modus string, stop context.CancelFunc) string {
	name := fmt.Sprintf("odr_abbruch_%x", rand.Uint64())
	sqlitedriver.MustRegisterScalarFunction(name, 0, func(*sqlitedriver.FunctionContext, []driver.Value) (driver.Value, error) {
		stop()
		if modus == modusZuerst {
			time.Sleep(wartezeitInterrupt)
		}
		return int64(0), nil
	})
	return name
}

// abbruchErgebnis ist, was ein Durchlauf von abbruchDurchlauf beobachtet.
type abbruchErgebnis struct {
	opErr      error
	aktiv      bool // Transaktion vor dem COMMIT offen
	w2         bool // die gelungene Operation derselben Transaktion ist sichtbar
	commitErr  error
	snapshotDa bool
	stateAlt   bool
}

// abweichung nennt, was vom Vertrag abweicht; leer, wenn er hält.
func (r abbruchErgebnis) abweichung() string {
	switch {
	case !errors.Is(r.opErr, context.Canceled):
		return fmt.Sprintf("Operation lieferte nicht context.Canceled: %v", r.opErr)
	case !r.aktiv:
		return "Transaktion vor dem COMMIT nicht mehr offen"
	case r.commitErr != nil:
		return fmt.Sprintf("COMMIT scheiterte: %v", r.commitErr)
	case !r.w2:
		return "w2 ging verloren"
	case !r.snapshotDa || !r.stateAlt:
		return "Teilwirkung nicht zurückgenommen"
	}
	return ""
}

func (r abbruchErgebnis) String() string {
	return fmt.Sprintf("opErr=%v aktivVorCommit=%v w2=%v commitErr=%v snapshotDa=%v stateAlt=%v",
		r.opErr, r.aktiv, r.w2, r.commitErr, r.snapshotDa, r.stateAlt)
}

func abbruchEntity(ws, schluessel string) model.EntityState {
	return model.EntityState{WorkspaceID: ws, EntitySet: "Products", EntityKey: schluessel,
		Payload: `{"ID":1}`, KnownProperties: []string{"ID"}, Completeness: model.EntityComplete, Version: 1}
}

// abbruchBasis öffnet eine Datenbank mit Service d, Baseline snap-1, Workspace w, dem Eintrag
// alt und dem Workspace-Snapshot sn.
func abbruchBasis(t *testing.T) *sqlite.DB {
	t.Helper()
	bg := context.Background()
	d, err := sqlite.Open(bg, filepath.Join(t.TempDir(), "r.db"), sqlite.Options{})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	t0 := time.Date(2026, 10, 19, 10, 0, 0, 123456789, time.UTC)
	ref := "sha256:" + strings.Repeat("a", 64)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err = d.SQL().Exec(`INSERT INTO services(id, service_root, odata_version) VALUES ('d', '/', '4.0')`)
	must(err)
	must(d.Sessions().Create(bg, model.Session{ID: "s1", ServiceID: "d", CreatedAt: t0, Status: model.SessionOpen}))
	_, _, err = d.Imports().Publish(bg, model.Publication{
		Import:           model.Import{ID: "imp-1", RequestHash: "req-1", BaselineSnapshotID: "snap-1", ServiceID: "d", CreatedAt: t0, ReportJSON: `{"ok":true}`},
		Snapshot:         model.BaselineSnapshot{ID: "snap-1", ServiceID: "d", CreatedAt: t0, StateBlob: ref, StateHash: "sh-1", MetadataHash: "mh"},
		SourceSessionIDs: []string{"s1"},
		Coverage:         []model.Coverage{{EntitySet: "Products", Status: model.CoverageComplete, EvidenceJSON: `{"pages":3}`}},
	})
	must(err)
	must(d.Workspaces().Create(bg, model.Workspace{ID: "w", ServiceID: "d", BaselineSnapshotID: "snap-1", CreatedAt: t0}))
	must(d.Entities().Put(bg, abbruchEntity("w", "alt")))
	must(d.Snapshots().CreateWorkspace(bg, model.WorkspaceSnapshot{ID: "sn", ServiceID: "d", WorkspaceID: "w", Name: "n", CreatedAt: t0,
		StateBlob: ref, StateHash: "sh-sn", MetadataHash: "mh"}))
	return d
}

// abbruchDurchlauf führt den Aufbau von transaktionTeilwirkung (outboundtest) in der Reihenfolge
// des Modus: ein Trigger nach dem Löschen ruft den Abbruch des Kontexts der Operation.
func abbruchDurchlauf(t *testing.T, modus, fall string) abbruchErgebnis {
	t.Helper()
	bg := context.Background()
	ctx, stop := context.WithCancel(bg)
	defer stop()
	funktion := abbruchFunktion(modus, stop)
	d := abbruchBasis(t)
	t0 := time.Date(2026, 10, 19, 10, 1, 0, 0, time.UTC)
	ops := map[string]struct {
		tabelle string
		op      func(context.Context, outbound.WorkspaceStores) error
	}{
		"Delete": {"snapshots", func(ctx context.Context, s outbound.WorkspaceStores) error { return s.Workspaces.Delete(ctx, "w") }},
		"ReplaceAll": {"entities", func(ctx context.Context, s outbound.WorkspaceStores) error {
			return s.Entities.ReplaceAll(ctx, "w", []model.EntityState{abbruchEntity("w", "neu")})
		}},
	}
	f := ops[fall]
	if _, err := d.SQL().Exec(`CREATE TRIGGER nach_loeschen AFTER DELETE ON ` + f.tabelle + ` BEGIN SELECT ` + funktion + `(); END`); err != nil {
		t.Fatal(err)
	}
	if modus == modusDanach {
		// Eine P: die vom Abbruch geweckte Goroutine des Treibers läuft erst, wenn das Statement fertig ist.
		defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	}
	var r abbruchErgebnis
	r.commitErr = d.WithWorkspaceTransaction(bg, func(s outbound.WorkspaceStores) error {
		if err := s.Workspaces.Create(bg, model.Workspace{ID: "w2", ServiceID: "d", BaselineSnapshotID: "snap-1", CreatedAt: t0}); err != nil {
			t.Fatal(err)
		}
		r.opErr = f.op(ctx, s)
		r.aktiv = sqlite.TransaktionAktiv(bg, s)
		_, err := s.Workspaces.Get(bg, "w2")
		r.w2 = err == nil
		return nil
	})
	_, errSn := d.Snapshots().GetWorkspace(bg, "w", "sn")
	r.snapshotDa = errSn == nil
	liste, errL := d.Entities().List(bg, "w")
	r.stateAlt = errL == nil && len(liste) == 1 && liste[0].EntityKey == "alt"
	return r
}

// TestTransaktionAbbruchImStatement fängt ein Statement der offenen Transaktion, das der
// Abbruch des Aufrufs unterbricht: kommt der Interrupt vor dem Ende des Statements, rollt SQLite
// die ganze Transaktion zurück und COMMIT scheitert. Beide Reihenfolgen müssen dasselbe
// ergeben: die Operation liefert context.Canceled, die Transaktion bleibt offen und bestätigt,
// die gelungene Operation bleibt, die Teilwirkung der gescheiterten ist zurückgenommen. Offen
// bleibt das Warten von 500 ms in der Reihenfolge interrupt-zuerst: der Treiber meldet nicht,
// wann der Interrupt angewendet ist, und unter extremer Last könnte er später kommen;
// TestTransaktionAbbruchHaltepunkt fängt dieselbe Klasse ohne Zeitfenster.
func TestTransaktionAbbruchImStatement(t *testing.T) {
	for _, modus := range []string{modusZuerst, modusDanach} {
		t.Run(modus, func(t *testing.T) {
			for _, fall := range []string{"Delete", "ReplaceAll"} {
				t.Run(fall, func(t *testing.T) {
					r := abbruchDurchlauf(t, modus, fall)
					if a := r.abweichung(); a != "" {
						t.Fatalf("%s: %s\n%s", modus, a, r)
					}
				})
			}
		})
	}
}

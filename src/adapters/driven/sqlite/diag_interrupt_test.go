package sqlite_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	sqlitedriver "modernc.org/sqlite"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Diagnose der Reihenfolge von Statement und Interrupt (Flake von TestTransaktionVertrag, Fall
// "eine gescheiterte Operation nimmt ihre Teilwirkung zurück"). Der Treiber ruft bei
// ctx.Done in einer eigenen Goroutine sqlite3_interrupt und meldet nicht, wann der Interrupt
// angewendet ist; SQLite rollt bei SQLITE_INTERRUPT in einem schreibenden Statement die ganze
// Transaktion zurück.

const (
	modusZuerst     = "interrupt-zuerst"
	modusDanach     = "statement-zuerst"
	modusNatuerlich = "natuerlich"

	// Grenze des Wartens auf den Interrupt in modusZuerst: der Treiber liefert kein Signal.
	wartezeitInterrupt = 500 * time.Millisecond
)

type ereignis struct {
	ns   int64
	was  string
	wert int64
}

// protokoll hält Ereignisse im Speicher: vorab angelegter Puffer, atomarer Zähler, kein Log im
// heißen Pfad.
type protokoll struct {
	n     atomic.Int64
	start time.Time
	e     []ereignis
}

func neuesProtokoll() *protokoll { return &protokoll{start: time.Now(), e: make([]ereignis, 64)} }

func (p *protokoll) add(was string, wert int64) {
	i := p.n.Add(1) - 1
	if int(i) < len(p.e) {
		p.e[i] = ereignis{ns: time.Since(p.start).Nanoseconds(), was: was, wert: wert}
	}
}

func (p *protokoll) String() string {
	var b strings.Builder
	for i := 0; i < int(min(p.n.Load(), int64(len(p.e)))); i++ {
		fmt.Fprintf(&b, "%10d ns  %-28s %d\n", p.e[i].ns, p.e[i].was, p.e[i].wert)
	}
	return b.String()
}

type diagLauf struct {
	modus string
	stop  context.CancelFunc
	prot  *protokoll
}

// registriereDiagFunktion meldet eine Funktion unter eigenem Namen an; sie gilt nur für
// Verbindungen, die danach öffnen. Sie ruft den Abbruch und wartet im Modus modusZuerst.
func registriereDiagFunktion(l *diagLauf) string {
	name := fmt.Sprintf("odr_diag_%x", rand.Uint64())
	sqlitedriver.MustRegisterScalarFunction(name, 0, func(*sqlitedriver.FunctionContext, []driver.Value) (driver.Value, error) {
		l.prot.add("funktion-start", 0)
		l.stop()
		l.prot.add("stop-gerufen", 0)
		if l.modus == modusZuerst {
			time.Sleep(wartezeitInterrupt)
			l.prot.add("gewartet", 0)
		}
		return int64(0), nil
	})
	return name
}

type diagErgebnis struct {
	opErr        error
	aktiv        bool // Transaktion vor dem COMMIT offen
	w2           bool
	commitErr    error
	snapshotDa   bool
	stateAlt     bool
	prot         *protokoll
	modus, fall  string
	unerwartetAb string // Abweichung vom Vertrag, leer wenn der Vertrag hält
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func diagBasis(t *testing.T) *sqlite.DB {
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
	must(d.Entities().Put(bg, diagEntity("w", "alt")))
	must(d.Snapshots().CreateWorkspace(bg, model.WorkspaceSnapshot{ID: "sn", ServiceID: "d", WorkspaceID: "w", Name: "n", CreatedAt: t0,
		StateBlob: ref, StateHash: "sh-sn", MetadataHash: "mh"}))
	return d
}

func diagEntity(ws, schluessel string) model.EntityState {
	return model.EntityState{WorkspaceID: ws, EntitySet: "Products", EntityKey: schluessel,
		Payload: `{"ID":1}`, KnownProperties: []string{"ID"}, Completeness: model.EntityComplete, Version: 1}
}

// diagLauf führt den Aufbau von transaktionTeilwirkung (outboundtest) in der Reihenfolge des Modus.
func diagDurchlauf(t *testing.T, modus, fall string) diagErgebnis {
	t.Helper()
	bg := context.Background()
	ctx, stop := context.WithCancel(bg)
	defer stop()
	l := &diagLauf{modus: modus, stop: stop, prot: neuesProtokoll()}
	funktion := registriereDiagFunktion(l)
	d := diagBasis(t)
	t0 := time.Date(2026, 10, 19, 10, 1, 0, 0, time.UTC)
	ops := map[string]struct {
		tabelle string
		op      func(context.Context, outbound.WorkspaceStores) error
	}{
		"Delete": {"snapshots", func(ctx context.Context, s outbound.WorkspaceStores) error { return s.Workspaces.Delete(ctx, "w") }},
		"ReplaceAll": {"entities", func(ctx context.Context, s outbound.WorkspaceStores) error {
			return s.Entities.ReplaceAll(ctx, "w", []model.EntityState{diagEntity("w", "neu")})
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
	r := diagErgebnis{prot: l.prot, modus: modus, fall: fall}
	l.prot.add("transaktion-start", 0)
	r.commitErr = d.WithWorkspaceTransaction(bg, func(s outbound.WorkspaceStores) error {
		if err := s.Workspaces.Create(bg, model.Workspace{ID: "w2", ServiceID: "d", BaselineSnapshotID: "snap-1", CreatedAt: t0}); err != nil {
			t.Fatal(err)
		}
		l.prot.add("op-start", 0)
		r.opErr = f.op(ctx, s)
		l.prot.add("op-ende", b2i(errors.Is(r.opErr, context.Canceled)))
		r.aktiv = sqlite.DiagTransaktionAktiv(bg, s)
		l.prot.add("transaktion-aktiv", b2i(r.aktiv))
		_, err := s.Workspaces.Get(bg, "w2")
		r.w2 = err == nil
		l.prot.add("w2-da", b2i(r.w2))
		return nil
	})
	l.prot.add("commit-ende", b2i(r.commitErr == nil))
	_, errSn := d.Snapshots().GetWorkspace(bg, "w", "sn")
	r.snapshotDa = errSn == nil
	liste, errL := d.Entities().List(bg, "w")
	r.stateAlt = errL == nil && len(liste) == 1 && liste[0].EntityKey == "alt"
	switch {
	case !errors.Is(r.opErr, context.Canceled):
		r.unerwartetAb = "Operation lieferte nicht context.Canceled: " + fmt.Sprint(r.opErr)
	case r.commitErr != nil:
		r.unerwartetAb = "COMMIT scheiterte: " + r.commitErr.Error()
	case !r.w2:
		r.unerwartetAb = "w2 ging verloren"
	case !r.snapshotDa || !r.stateAlt:
		r.unerwartetAb = "Teilwirkung nicht zurückgenommen"
	}
	return r
}

func diagBericht(r diagErgebnis) string {
	return fmt.Sprintf("modus=%s fall=%s opErr=%v aktivVorCommit=%v w2=%v commitErr=%v snapshotDa=%v stateAlt=%v abweichung=%q\n%s",
		r.modus, r.fall, r.opErr, r.aktiv, r.w2, r.commitErr, r.snapshotDa, r.stateAlt, r.unerwartetAb, r.prot)
}

// TestDiagInterruptReihenfolge belegt den Mechanismus: kommt der Interrupt vor dem Ende des
// Statements, rollt SQLite die ganze Transaktion zurück und COMMIT scheitert; endet das Statement
// zuerst, bleibt die Transaktion bestehen. Der Test hält den Ist-Zustand des Adapters fest und wird
// rot, sobald der Adapter die Reihenfolge verträgt.
func TestDiagInterruptReihenfolge(t *testing.T) {
	for _, fall := range []string{"Delete", "ReplaceAll"} {
		t.Run(fall, func(t *testing.T) {
			const lauefe = 3
			for i := range lauefe {
				r := diagDurchlauf(t, modusZuerst, fall)
				if i == 0 {
					t.Logf("Protokoll interrupt-zuerst:\n%s", diagBericht(r))
				}
				if r.aktiv || r.commitErr == nil || !strings.Contains(r.commitErr.Error(), "no transaction is active") {
					t.Fatalf("interrupt-zuerst Lauf %d: erwartet Rollback der ganzen Transaktion\n%s", i, diagBericht(r))
				}
				if r.w2 {
					t.Fatalf("interrupt-zuerst Lauf %d: w2 überlebte\n%s", i, diagBericht(r))
				}
			}
			for i := range lauefe {
				r := diagDurchlauf(t, modusDanach, fall)
				if i == 0 {
					t.Logf("Protokoll statement-zuerst:\n%s", diagBericht(r))
				}
				if r.unerwartetAb != "" || !r.aktiv {
					t.Fatalf("statement-zuerst Lauf %d: erwartet Vertrag erfüllt\n%s", i, diagBericht(r))
				}
			}
		})
	}
}

// TestDiagInterruptNatuerlich fährt den ursprünglichen Ablauf (Abbruch ohne Warten, alle P) so oft,
// wie DIAG_N sagt (Voreinstellung: kein Lauf). Rote Läufe gibt der Test mit Protokoll auf stdout aus
// und zählt sie; er scheitert nicht.
func TestDiagInterruptNatuerlich(t *testing.T) {
	n, _ := strconv.Atoi(os.Getenv("DIAG_N"))
	rot := map[string]int{}
	for i := range n {
		for _, fall := range []string{"Delete", "ReplaceAll"} {
			r := diagDurchlauf(t, modusNatuerlich, fall)
			if r.unerwartetAb != "" {
				rot[fall]++
				t.Logf("=== ROT lauf=%d ===\n%s", i, diagBericht(r))
			}
		}
	}
	t.Logf("DIAG natuerlich: n=%d rot=%v", n, rot)
}

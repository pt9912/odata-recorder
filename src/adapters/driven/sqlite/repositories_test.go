package sqlite_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"sync/atomic"
	"testing"

	sqlitedriver "modernc.org/sqlite"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

func neueVerbindung(t *testing.T) outboundtest.Verbindung {
	t.Helper()
	// Eine Funktion gilt nur für Verbindungen, die nach ihrer Anmeldung öffnen; sie ruft den
	// Abbruch, den BrichNachImportAb setzt.
	var abbruch atomic.Pointer[context.CancelFunc]
	name := fmt.Sprintf("odr_abbruch_%x", rand.Uint64())
	sqlitedriver.MustRegisterScalarFunction(name, 0, func(*sqlitedriver.FunctionContext, []driver.Value) (driver.Value, error) {
		if f := abbruch.Load(); f != nil {
			(*f)()
		}
		return int64(0), nil
	})
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	return outboundtest.Verbindung{
		Sessions:  d.Sessions(),
		Recording: d.Recording(),
		Metadata:  d.Metadata(),
		Snapshots: d.Snapshots(),
		Imports:   d.Imports(),
		Coverage:  d.Coverage(),
		FuegeDienstHinzu: func(t *testing.T, id string) {
			t.Helper()
			if _, err := d.SQL().Exec(`INSERT INTO services(id, service_root, odata_version) VALUES (?, '/', '4.0')`, id); err != nil {
				t.Fatal(err)
			}
		},
		MacheZurBaselineQuelle: func(t *testing.T, sessionID string) {
			t.Helper()
			if _, err := d.SQL().Exec(`INSERT INTO snapshots(id, service_id, created_at, state_blob, state_hash, metadata_hash, snapshot_kind)
				SELECT 'b-' || id, service_id, created_at, 'x', 'h', 'm', 'baseline' FROM sessions WHERE id = ?`, sessionID); err != nil {
				t.Fatal(err)
			}
			if _, err := d.SQL().Exec(`INSERT INTO baseline_sources(baseline_snapshot_id, baseline_kind, session_id) VALUES ('b-' || ?, 'baseline', ?)`, sessionID, sessionID); err != nil {
				t.Fatal(err)
			}
		},
		FuegeWorkspaceSnapshotHinzu: func(t *testing.T, id, serviceID string) {
			t.Helper()
			if _, err := d.SQL().Exec(`INSERT INTO snapshots(id, service_id, created_at, state_blob, state_hash, metadata_hash, snapshot_kind, name, workspace_id)
				VALUES (?, ?, '2026-10-19T10:00:00.000000000Z', 'x', 'h', 'm', 'workspace', 'n', 'w')`, id, serviceID); err != nil {
				t.Fatal(err)
			}
		},
		BrichNachImportAb: func(t *testing.T, stop context.CancelFunc) {
			t.Helper()
			abbruch.Store(&stop)
			if _, err := d.SQL().Exec(`CREATE TRIGGER nach_import AFTER INSERT ON imports BEGIN SELECT ` + name + `(); END`); err != nil {
				t.Fatal(err)
			}
		},
	}
}

func TestSessionVertrag(t *testing.T)   { outboundtest.RunSessionVertrag(t, neueVerbindung) }
func TestRecordingVertrag(t *testing.T) { outboundtest.RunRecordingVertrag(t, neueVerbindung) }
func TestServiceVertrag(t *testing.T)   { outboundtest.RunServiceVertrag(t, neueVerbindung) }
func TestBaselineVertrag(t *testing.T)  { outboundtest.RunBaselineVertrag(t, neueVerbindung) }

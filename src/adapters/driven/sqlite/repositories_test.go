package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

func neueVerbindung(t *testing.T) outboundtest.Verbindung {
	t.Helper()
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	return outboundtest.Verbindung{
		Sessions:  d.Sessions(),
		Recording: d.Recording(),
		FuegeDienstHinzu: func(t *testing.T, id string) {
			t.Helper()
			if _, err := d.SQL().Exec(`INSERT INTO services(id, service_root, odata_version) VALUES (?, '/', '4.0')`, id); err != nil {
				t.Fatal(err)
			}
		},
		MacheZurBaselineQuelle: func(t *testing.T, sessionID string) {
			t.Helper()
			stmts := []string{
				`INSERT INTO snapshots(id, service_id, created_at, state_blob, state_hash, metadata_hash, snapshot_kind)
				 SELECT 'b-' || id, service_id, created_at, 'x', 'h', 'm', 'baseline' FROM sessions WHERE id = ?`,
				`INSERT INTO baseline_sources(baseline_snapshot_id, baseline_kind, session_id) VALUES ('b-' || ?, 'baseline', ?)`,
			}
			for _, s := range stmts {
				args := []any{sessionID}
				if s[:27] == stmts[1][:27] {
					args = []any{sessionID, sessionID}
				}
				if _, err := d.SQL().Exec(s, args...); err != nil {
					t.Fatal(err)
				}
			}
		},
	}
}

func TestSessionVertrag(t *testing.T)   { outboundtest.RunSessionVertrag(t, neueVerbindung) }
func TestRecordingVertrag(t *testing.T) { outboundtest.RunRecordingVertrag(t, neueVerbindung) }

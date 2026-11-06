package audit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/audit"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// frist ist die Zeit, nach der ein Aufruf als hängend gilt.
const frist = 5 * time.Second

// stand ist die Zeit der festen Uhr der Tests.
func stand() time.Time { return time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC) }

// testUhr zählt jeden Aufruf von Now und liefert die Zeit aus quelle(n) für den n-ten Aufruf.
type testUhr struct {
	mu      sync.Mutex
	aufrufe int
	quelle  func(n int) time.Time
}

func (u *testUhr) Now() time.Time {
	u.mu.Lock()
	u.aufrufe++
	n := u.aufrufe
	u.mu.Unlock()
	return u.quelle(n)
}

func (u *testUhr) Aufrufe() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.aufrufe
}

// zaehlIDs trägt alles, was FestIDs trägt, und zählt jede der fünf Operationen, auch die, die der
// Adapter nie rufen darf.
type zaehlIDs struct {
	inner *supporttest.FestIDs
	mu    sync.Mutex
	zahl  map[string]int
}

func neueZaehlIDs() *zaehlIDs {
	return &zaehlIDs{inner: supporttest.NewFestIDs(), zahl: map[string]int{}}
}

func (g *zaehlIDs) zaehle(op string) {
	g.mu.Lock()
	g.zahl[op]++
	g.mu.Unlock()
}

func (g *zaehlIDs) SessionID() string      { g.zaehle("session"); return g.inner.SessionID() }
func (g *zaehlIDs) ExchangeID() string     { g.zaehle("exchange"); return g.inner.ExchangeID() }
func (g *zaehlIDs) KorrelationsID() string { g.zaehle("korrelation"); return g.inner.KorrelationsID() }
func (g *zaehlIDs) WorkspaceID() string    { g.zaehle("workspace"); return g.inner.WorkspaceID() }
func (g *zaehlIDs) SnapshotID() string     { g.zaehle("snapshot"); return g.inner.SnapshotID() }

func (g *zaehlIDs) Korrelation() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.zahl["korrelation"]
}

func (g *zaehlIDs) Gesamt() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := 0
	for _, z := range g.zahl {
		n += z
	}
	return n
}

// aufbau ist ein Adapter mit seinen Fakes.
type aufbau struct {
	pfad string
	uhr  *testUhr
	ids  *zaehlIDs
	log  *bytes.Buffer
	d    *audit.Datei
}

func festeZeit(int) time.Time { return stand() }

func neu(t *testing.T, quelle func(n int) time.Time) *aufbau {
	t.Helper()
	return neuAn(t, filepath.Join(t.TempDir(), audit.DateiName), quelle)
}

func neuAn(t *testing.T, pfad string, quelle func(n int) time.Time) *aufbau {
	t.Helper()
	a := &aufbau{pfad: pfad, uhr: &testUhr{quelle: quelle}, ids: neueZaehlIDs(), log: &bytes.Buffer{}}
	logger := slog.New(slog.NewJSONHandler(a.log, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, at slog.Attr) slog.Attr {
			if at.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return at
		},
	}))
	d, err := audit.New(audit.Config{Pfad: pfad, Uhr: a.uhr, IDs: a.ids, Log: logger})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a.d = d
	return a
}

// inhalt liest die Datei; fehlt sie, ist das Ergebnis der leere Text.
func (a *aufbau) inhalt(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(a.pfad)
	if errors.Is(err, os.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatalf("Datei lesen: %v", err)
	}
	return string(b)
}

func (a *aufbau) dateiFehlt() bool {
	_, err := os.Stat(a.pfad)
	return errors.Is(err, os.ErrNotExist)
}

// logZeilen liest die Zeilen des Logs als Objekte.
func (a *aufbau) logZeilen(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, z := range strings.Split(strings.TrimSuffix(a.log.String(), "\n"), "\n") {
		if z == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(z), &m); err != nil {
			t.Fatalf("Log-Zeile %q ist kein JSON: %v", z, err)
		}
		out = append(out, m)
	}
	return out
}

// nichtsGerufen prüft, dass weder Uhr noch Generator gerufen wurden und nichts geschrieben ist.
func (a *aufbau) nichtsGerufen(t *testing.T, was string) {
	t.Helper()
	if n := a.uhr.Aufrufe(); n != 0 {
		t.Errorf("%s: Uhr %d-mal gerufen, erwartet 0", was, n)
	}
	if n := a.ids.Gesamt(); n != 0 {
		t.Errorf("%s: Generator %d-mal gerufen, erwartet 0", was, n)
	}
	if !a.dateiFehlt() {
		t.Errorf("%s: Datei angelegt: %q", was, a.inhalt(t))
	}
}

// mitFrist ruft f und bricht den Test ab, wenn f nicht innerhalb der Frist zurückkehrt.
func mitFrist(t *testing.T, was string, f func() error) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() { ch <- f() }()
	select {
	case err := <-ch:
		return err
	case <-time.After(frist):
		t.Fatalf("%s: keine Rückkehr innerhalb von %v", was, frist)
		return nil
	}
}

// orakelPflicht ist die Tabelle von SPEC-234 Nr. 2 aus dem Wortlaut, unabhängig vom Code:
// je Art die Pflichtfelder in der Folge service_id, session_id, workspace_id, import_id, error_code.
func orakelPflicht() map[support.AuditArt][5]bool {
	return map[support.AuditArt][5]bool{
		support.ArtSessionGeloescht:     {true, true, false, false, false},
		support.ArtSessionAufbewahrung:  {true, true, false, false, false},
		support.ArtWorkspaceGeloescht:   {false, false, true, false, false},
		support.ArtImportFehlgeschlagen: {true, false, false, true, true},
	}
}

func orakelFelder() [5]support.Feld {
	return [5]support.Feld{support.FeldServiceID, support.FeldSessionID, support.FeldWorkspaceID, support.FeldImportID, support.FeldErrorCode}
}

// wertVon liefert den gültigen Wert eines Felds.
func wertVon(f support.Feld) string { return "w-" + string(f) }

// setze setzt oder leert ein Feld des Eintrags.
func setze(e *support.Eintrag, f support.Feld, an bool) {
	w := ""
	if an {
		w = wertVon(f)
	}
	switch f {
	case support.FeldServiceID:
		e.ServiceID = w
	case support.FeldSessionID:
		e.SessionID = w
	case support.FeldWorkspaceID:
		e.WorkspaceID = w
	case support.FeldImportID:
		e.ImportID = w
	case support.FeldErrorCode:
		e.ErrorCode = 0
		if an {
			e.ErrorCode = model.CodeE6101
		}
	}
}

// gueltig bildet den Eintrag der Art mit genau ihren Pflichtfeldern.
func gueltig(art support.AuditArt) support.Eintrag {
	e := support.Eintrag{Art: art}
	soll := orakelPflicht()[art]
	for i, f := range orakelFelder() {
		setze(&e, f, soll[i])
	}
	return e
}

func slogNull() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }

// mkdirFuer legt das Verzeichnis "fehlt" unter wurzel an.
func mkdirFuer(wurzel string) error { return os.Mkdir(filepath.Join(wurzel, "fehlt"), 0o700) }

func corID(n int) string { return "cor-" + strings.Repeat("0", 31) + string(rune('0'+n)) }

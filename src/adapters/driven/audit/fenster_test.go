package audit_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/audit"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// schnellVerz liefert ein Verzeichnis im Arbeitsspeicher (/dev/shm), wo fsync nichts kostet;
// fehlt es, das Verzeichnis des Tests.
func schnellVerz(t *testing.T) string {
	t.Helper()
	if info, err := os.Stat("/dev/shm"); err == nil && info.IsDir() {
		if d, err := os.MkdirTemp("/dev/shm", "audit-"); err == nil {
			t.Cleanup(func() { _ = os.RemoveAll(d) })
			return d
		}
	}
	return t.TempDir()
}

// M1: SPEC-235 Nr. 2: der Adapter schließt die Datei wieder und hält kein Handle. Nach 100
// weiteren Append steht die Zahl der offenen Dateien des Prozesses so wie vor ihnen.
func TestKeinHandleBleibt(t *testing.T) {
	const n = 100
	a := neuAn(t, filepath.Join(schnellVerz(t), audit.DateiName), festeZeit)
	offen := func() int {
		es, err := os.ReadDir("/proc/self/fd")
		if err != nil {
			t.Fatalf("/proc/self/fd: %v", err)
		}
		return len(es)
	}
	if err := a.d.Append(gueltig(support.ArtSessionGeloescht)); err != nil {
		t.Fatal(err)
	}
	vorher := offen()
	for range n {
		if err := a.d.Append(gueltig(support.ArtSessionGeloescht)); err != nil {
			t.Fatal(err)
		}
	}
	if nachher := offen(); nachher != vorher {
		t.Errorf("offene Dateien: %d vor, %d nach %d Append", vorher, nachher, n)
	}
	// Auch der Fehlerpfad nach dem Öffnen (Schreiben auf /dev/full) schließt die Datei.
	voll := neuAn(t, "/dev/full", festeZeit)
	_ = voll.d.Append(gueltig(support.ArtSessionGeloescht))
	vorher = offen()
	for range n {
		_ = voll.d.Append(gueltig(support.ArtSessionGeloescht))
	}
	if nachher := offen(); nachher != vorher {
		t.Errorf("offene Dateien nach Fehlern: %d vor, %d nach %d Append", vorher, nachher, n)
	}
}

// ersteNull liefert im ersten Aufruf die Nullzeit und danach die feste Zeit.
func ersteNull(n int) time.Time {
	if n == 1 {
		return time.Time{}
	}
	return stand()
}

// hakenHandler ruft bei jedem Log-Eintrag f.
type hakenHandler struct{ f func() }

func (h hakenHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (h hakenHandler) Handle(context.Context, slog.Record) error { h.f(); return nil }
func (h hakenHandler) WithAttrs([]slog.Attr) slog.Handler        { return h }
func (h hakenHandler) WithGroup(string) slog.Handler             { return h }

// M2: SPEC-235 Nr. 5: der Log-Eintrag entsteht nach der Freigabe der Sperre. Der Handler ruft
// ein zweites Append desselben Adapters; es kehrt zurück, solange der Handler läuft. Je Stelle
// des Scheiterns ein Fall (Zeit, Öffnen, Schreiben).
func TestLogNachFreigabeDerSperre(t *testing.T) {
	stellen := map[string]struct {
		pfad string
		uhr  func(int) time.Time
	}{
		"Zeit":      {filepath.Join(schnellVerz(t), audit.DateiName), ersteNull},
		"Öffnen":    {filepath.Join(t.TempDir(), "fehlt", audit.DateiName), festeZeit},
		"Schreiben": {"/dev/full", festeZeit},
	}
	if len(stellen) != 3 {
		t.Fatalf("%d Stellen, erwartet 3", len(stellen))
	}
	for name, s := range stellen {
		var d *audit.Datei
		var blockiert, schonGerufen atomic.Bool
		h := hakenHandler{f: func() {
			if !schonGerufen.CompareAndSwap(false, true) {
				return
			}
			fertig := make(chan struct{})
			go func() { _ = d.Append(gueltig(support.ArtSessionGeloescht)); close(fertig) }()
			select {
			case <-fertig:
			case <-time.After(time.Second):
				blockiert.Store(true)
			}
		}}
		var err error
		d, err = audit.New(audit.Config{Pfad: s.pfad, Uhr: &testUhr{quelle: s.uhr}, IDs: neueZaehlIDs(), Log: slog.New(h)})
		if err != nil {
			t.Fatal(err)
		}
		if err := d.Append(gueltig(support.ArtSessionGeloescht)); err == nil {
			t.Errorf("%s: Append gelang", name)
		}
		if !schonGerufen.Load() {
			t.Errorf("%s: der Handler wurde nicht gerufen", name)
		}
		if blockiert.Load() {
			t.Errorf("%s: der Log-Handler lief unter der Sperre", name)
		}
	}
}

// M3: Fenster der Sperre zwischen Uhr.Now() und dem Schreiben, ohne Haken in der Datei: 40 Runden
// mit je 64 gleichzeitigen Append auf eine frische Datei (leer, und mit der halben Zeile "halb");
// die Zeitpunkte nehmen nicht ab und keine Zeile ist leer. Eine Mutation, die die Sperre im Fenster
// löst und sofort wieder nimmt, lässt in einer Runde ein Append dazwischen.
func TestFensterZwischenUhrUndSchreiben(t *testing.T) {
	const runden, n = 40, 64
	verz := schnellVerz(t)
	for _, vorher := range []string{"", "halb"} {
		for r := range runden {
			a := neuAn(t, filepath.Join(verz, fmt.Sprintf("a-%d-%d.jsonl", len(vorher), r)), func(k int) time.Time {
				return stand().Add(time.Duration(k) * time.Nanosecond)
			})
			if vorher != "" {
				if err := os.WriteFile(a.pfad, []byte(vorher), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			start := make(chan struct{})
			var wg sync.WaitGroup
			for range n {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					_ = a.d.Append(gueltig(support.ArtWorkspaceGeloescht))
				}()
			}
			close(start)
			wg.Wait()
			zeilen := strings.Split(strings.TrimSuffix(a.inhalt(t), "\n"), "\n")
			if len(zeilen) != n+min(1, len(vorher)) {
				t.Fatalf("Vorher %q, Runde %d: %d Zeilen, erwartet %d", vorher, r, len(zeilen), n+min(1, len(vorher)))
			}
			letzte := ""
			for i, z := range zeilen {
				if z == "" {
					t.Fatalf("Vorher %q, Runde %d: Zeile %d ist leer", vorher, r, i)
				}
				if !strings.HasPrefix(z, "{") {
					continue
				}
				if z[:44] < letzte {
					t.Fatalf("Vorher %q, Runde %d: Zeile %d nimmt in der Zeit ab: %s gegen %s", vorher, r, i, z[:44], letzte)
				}
				letzte = z[:44]
			}
		}
	}
}

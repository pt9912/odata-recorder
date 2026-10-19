package sqlite_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
)

func erwarteteTabellen() []string {
	return []string{
		"baseline_sources", "coverage", "entities", "exchanges", "imports",
		"replay_cursors", "services", "sessions", "snapshots", "workspaces",
	}
}

func open(t *testing.T, pfad string) *sqlite.DB {
	t.Helper()
	d, err := sqlite.Open(context.Background(), pfad, sqlite.Options{})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func tabellen(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var namen []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		namen = append(namen, n)
	}
	sort.Strings(namen)
	return namen
}

func skalar(t *testing.T, db *sql.DB, q string) string {
	t.Helper()
	var s string
	if err := db.QueryRow(q).Scan(&s); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return s
}

func TestNeueDateiWirdMigriert(t *testing.T) {
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	if got := tabellen(t, d.SQL()); strings.Join(got, ",") != strings.Join(erwarteteTabellen(), ",") {
		t.Fatalf("Tabellen %v, erwartet %v", got, erwarteteTabellen())
	}
	if v := skalar(t, d.SQL(), "PRAGMA user_version"); v != "1" {
		t.Fatalf("user_version %s", v)
	}
	if v := skalar(t, d.SQL(), "PRAGMA integrity_check"); v != "ok" {
		t.Fatalf("integrity_check %s", v)
	}
	rows, err := d.SQL().Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		t.Fatal("foreign_key_check meldet Zeilen")
	}
}

func TestZweitesOeffnenVeraendertNichts(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	d := open(t, pfad)
	if _, err := d.SQL().Exec(`INSERT INTO services(id, service_root, odata_version) VALUES ('s', '/', '4.0')`); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	d2 := open(t, pfad)
	if n := skalar(t, d2.SQL(), "SELECT count(*) FROM services"); n != "1" {
		t.Fatalf("Zeilen nach dem Wiederöffnen: %s", n)
	}
	if got := tabellen(t, d2.SQL()); len(got) != len(erwarteteTabellen()) {
		t.Fatalf("Tabellen %v", got)
	}
}

func TestPragmasJederVerbindung(t *testing.T) {
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	ctx := context.Background()
	c1, err := d.SQL().Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c1.Close() }()
	c2, err := d.SQL().Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c2.Close() }()
	for i, c := range []*sql.Conn{c1, c2} {
		for pragma, want := range map[string]string{"foreign_keys": "1", "busy_timeout": "5000", "journal_mode": "wal"} {
			var got string
			if err := c.QueryRowContext(ctx, "PRAGMA "+pragma).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("Verbindung %d: %s = %s, erwartet %s", i, pragma, got, want)
			}
		}
	}
	if _, err := c2.ExecContext(ctx, `INSERT INTO sessions(id, service_id, created_at, status) VALUES ('x', 'unbekannt', 't', 'open')`); err == nil {
		t.Fatal("Fremdschlüsselverletzung wurde angenommen")
	}
}

func TestBusyTimeoutAusDerOption(t *testing.T) {
	d, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "r.db"), sqlite.Options{BusyTimeout: 1234 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	if v := skalar(t, d.SQL(), "PRAGMA busy_timeout"); v != "1234" {
		t.Fatalf("busy_timeout %s", v)
	}
}

func roh(t *testing.T, pfad, stmt string) {
	t.Helper()
	db, err := sql.Open("sqlite", pfad)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(stmt); err != nil {
		t.Fatal(err)
	}
}

// zustand liefert den Hash der Datei und die Namen aller Dateien neben ihr.
func zustand(t *testing.T, pfad string) (string, []string) {
	t.Helper()
	b, err := os.ReadFile(pfad)
	if err != nil {
		t.Fatal(err)
	}
	eintraege, err := os.ReadDir(filepath.Dir(pfad))
	if err != nil {
		t.Fatal(err)
	}
	var namen []string
	for _, e := range eintraege {
		namen = append(namen, e.Name())
	}
	h := sha256.Sum256(b)
	return string(h[:]), namen
}

func TestNeuereSchemaversionWirdVerweigert(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	roh(t, pfad, "PRAGMA user_version = 2")
	vorher, dateienVorher := zustand(t, pfad)
	defer func() {
		nachher, dateien := zustand(t, pfad)
		if nachher != vorher || strings.Join(dateien, ",") != strings.Join(dateienVorher, ",") {
			t.Error("die abgelehnte Datei wurde verändert")
		}
	}()
	_, err := sqlite.Open(context.Background(), pfad, sqlite.Options{})
	if !errors.Is(err, sqlite.ErrSchemaNeuer) || !strings.Contains(err.Error(), "2") || !strings.Contains(err.Error(), "1") {
		t.Fatalf("Fehler %v", err)
	}
	roh(t, pfad, "SELECT 1")
	db, _ := sql.Open("sqlite", pfad)
	defer func() { _ = db.Close() }()
	if v := skalar(t, db, "PRAGMA user_version"); v != "2" {
		t.Fatalf("user_version verändert: %s", v)
	}
	if got := tabellen(t, db); len(got) != 0 {
		t.Fatalf("Tabellen angelegt: %v", got)
	}
}

func TestFremdeDatenbankWirdVerweigert(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	roh(t, pfad, "CREATE TABLE fremd (a INTEGER)")
	vorher, dateienVorher := zustand(t, pfad)
	defer func() {
		nachher, dateien := zustand(t, pfad)
		if nachher != vorher || strings.Join(dateien, ",") != strings.Join(dateienVorher, ",") {
			t.Error("die abgelehnte Datei wurde verändert")
		}
	}()
	_, err := sqlite.Open(context.Background(), pfad, sqlite.Options{})
	if !errors.Is(err, sqlite.ErrFremdeDatenbank) {
		t.Fatalf("Fehler %v", err)
	}
	db, _ := sql.Open("sqlite", pfad)
	defer func() { _ = db.Close() }()
	if got := tabellen(t, db); strings.Join(got, ",") != "fremd" {
		t.Fatalf("Tabellen %v", got)
	}
}

func TestKeineSQLiteDateiIstBeschaedigt(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	inhalt := []byte(strings.Repeat("kein SQLite, nur Text. ", 100))
	if err := os.WriteFile(pfad, inhalt, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := sqlite.Open(context.Background(), pfad, sqlite.Options{})
	if !errors.Is(err, sqlite.ErrBeschaedigt) {
		t.Fatalf("Fehler %v", err)
	}
	got, rerr := os.ReadFile(pfad)
	if rerr != nil || string(got) != string(inhalt) {
		t.Fatal("Datei wurde verändert")
	}
	if _, dateien := zustand(t, pfad); len(dateien) != 1 {
		t.Fatalf("Dateien neben der abgelehnten: %v", dateien)
	}
}

func TestAbgeschnittenerRumpfIstBeschaedigt(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	d := open(t, pfad)
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	voll, err := os.ReadFile(pfad)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pfad, voll[:100], 0o600); err != nil {
		t.Fatal(err)
	}
	vorher, dateienVorher := zustand(t, pfad)
	_, err = sqlite.Open(context.Background(), pfad, sqlite.Options{})
	if !errors.Is(err, sqlite.ErrBeschaedigt) {
		t.Fatalf("Fehler %v", err)
	}
	nachher, dateien := zustand(t, pfad)
	if nachher != vorher || strings.Join(dateien, ",") != strings.Join(dateienVorher, ",") {
		t.Fatal("die abgelehnte Datei wurde verändert")
	}
}

func TestUngueltigePfade(t *testing.T) {
	dir := t.TempDir()
	fehlend := filepath.Join(dir, "gibt-es-nicht", "r.db")
	for name, pfad := range map[string]string{"leer": "", "memory": ":memory:", "verzeichnis": dir, "verzeichnis fehlt": fehlend} {
		t.Run(name, func(t *testing.T) {
			if _, err := sqlite.Open(context.Background(), pfad, sqlite.Options{}); !errors.Is(err, sqlite.ErrPfad) {
				t.Fatalf("Fehler %v", err)
			}
		})
	}
	if _, err := os.Stat(filepath.Dir(fehlend)); err == nil {
		t.Fatal("Open hat ein Verzeichnis angelegt")
	}
}

func TestGleichzeitigesOeffnenMigrierteinmal(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	var wg sync.WaitGroup
	fehler := make([]error, 8)
	dbs := make([]*sqlite.DB, 8)
	for i := range fehler {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dbs[i], fehler[i] = sqlite.Open(context.Background(), pfad, sqlite.Options{})
		}()
	}
	wg.Wait()
	for i, err := range fehler {
		if err != nil {
			t.Fatalf("Öffnung %d: %v", i, err)
		}
		defer func() { _ = dbs[i].Close() }()
	}
	if got := tabellen(t, dbs[0].SQL()); len(got) != len(erwarteteTabellen()) {
		t.Fatalf("Tabellen %v", got)
	}
}

func TestCloseIstIdempotent(t *testing.T) {
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatalf("zweites Close: %v", err)
	}
}

func TestMigrationLaeuftInEinerTransaktion(t *testing.T) {
	o := open(t, filepath.Join(t.TempDir(), "leer.db"))
	if _, err := o.SQL().Exec("DROP TABLE imports; DROP TABLE coverage; DROP TABLE baseline_sources; DROP TABLE replay_cursors; DROP TABLE entities; DROP TABLE workspaces; DROP TABLE snapshots; DROP TABLE exchanges; DROP TABLE sessions; DROP TABLE services; PRAGMA user_version = 0"); err != nil {
		t.Fatal(err)
	}
	err := o.MigrateWith(context.Background(), "CREATE TABLE a (x INTEGER); CREATE TABLE a (y INTEGER);")
	if err == nil {
		t.Fatal("fehlerhaftes DDL wurde angenommen")
	}
	if got := tabellen(t, o.SQL()); len(got) != 0 {
		t.Fatalf("Teile der Migration bleiben: %v", got)
	}
	if v := skalar(t, o.SQL(), "PRAGMA user_version"); v != "0" {
		t.Fatalf("user_version %s", v)
	}
}

func TestAndereFehlerBleibenUnklassifiziert(t *testing.T) {
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	_, err := d.SQL().Exec(`INSERT INTO sessions(id, service_id, created_at, status) VALUES ('x', 'unbekannt', 't', 'open')`)
	if err == nil {
		t.Fatal("Fremdschlüsselverletzung wurde angenommen")
	}
	if got := sqlite.Klassifiziere(err); errors.Is(got, sqlite.ErrBeschaedigt) || got != err {
		t.Fatalf("Fehler wurde verändert: %v", got)
	}
	if got := sqlite.Klassifiziere(errors.New("anderer Fehler")); errors.Is(got, sqlite.ErrBeschaedigt) {
		t.Fatal("fremder Fehler wurde klassifiziert")
	}
}

func journalMode(t *testing.T, pfad string) string {
	t.Helper()
	db, err := sql.Open("sqlite", pfad)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	return skalar(t, db, "PRAGMA journal_mode")
}

func TestOeffnenStelltWalWiederHer(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	d := open(t, pfad)
	if _, err := d.SQL().Exec("PRAGMA journal_mode = delete"); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if got := journalMode(t, pfad); got != "delete" {
		t.Fatalf("Aufbau: Modus %s", got)
	}
	d2 := open(t, pfad)
	if err := d2.Close(); err != nil {
		t.Fatal(err)
	}
	if got := journalMode(t, pfad); got != "wal" {
		t.Fatalf("Modus nach dem Öffnen: %s", got)
	}
}

func TestWalUmstellenScheitertNachDerFristUndWirdNachgeholt(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "r.db")
	d := open(t, pfad)
	if _, err := d.SQL().Exec("PRAGMA journal_mode = delete"); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	leser, err := sql.Open("sqlite", pfad)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = leser.Close() }()
	tx, err := leser.Begin()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := tx.QueryRow("SELECT count(*) FROM services").Scan(&n); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err = sqlite.Open(context.Background(), pfad, sqlite.Options{BusyTimeout: 150 * time.Millisecond})
	if err == nil {
		t.Fatal("Öffnen gelang trotz gehaltener Lesesperre")
	}
	if d := time.Since(start); d < 100*time.Millisecond || d > 3*time.Second {
		t.Fatalf("Dauer %v passt nicht zur Frist von 150 ms", d)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	d2 := open(t, pfad)
	if err := d2.Close(); err != nil {
		t.Fatal(err)
	}
	if got := journalMode(t, pfad); got != "wal" {
		t.Fatalf("Modus nach dem zweiten Öffnen: %s", got)
	}
}

// belegtFehler liefert einen echten SQLITE_BUSY-Fehler: zwei Verbindungen ohne Wartezeit.
func belegtFehler(t *testing.T) error {
	t.Helper()
	pfad := filepath.Join(t.TempDir(), "b.db")
	a, err := sql.Open("sqlite", pfad)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()
	ca, err := a.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ca.Close() }()
	if _, err := ca.ExecContext(context.Background(), "CREATE TABLE t (x); BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	b, err := sql.Open("sqlite", pfad)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = b.Close() }()
	_, err = b.Exec("BEGIN IMMEDIATE")
	if err == nil {
		t.Fatal("zweite Schreibsperre gelang")
	}
	return err
}

func TestWalUmstellenWiederholtBeiBelegt(t *testing.T) {
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	belegt := belegtFehler(t)
	n := 0
	err := d.StelleWalSicherMit(context.Background(), 2*time.Second, func(context.Context) error {
		n++
		if n < 4 {
			return belegt
		}
		return nil
	})
	if err != nil || n != 4 {
		t.Fatalf("Versuche %d, Fehler %v", n, err)
	}
}

func TestWalUmstellenGibtNachDerFristAuf(t *testing.T) {
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	belegt := belegtFehler(t)
	start := time.Now()
	err := d.StelleWalSicherMit(context.Background(), 60*time.Millisecond, func(context.Context) error { return belegt })
	if !errors.Is(err, belegt) {
		t.Fatalf("Fehler %v", err)
	}
	if dauer := time.Since(start); dauer < 50*time.Millisecond || dauer > 2*time.Second {
		t.Fatalf("Dauer %v", dauer)
	}
}

func TestWalUmstellenWiederholtNurBeiBelegt(t *testing.T) {
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	n := 0
	_, fehler := d.SQL().Exec(`INSERT INTO sessions(id, service_id, created_at, status) VALUES ('x', 'unbekannt', 't', 'open')`)
	if fehler == nil {
		t.Fatal("Fremdschlüsselverletzung wurde angenommen")
	}
	err := d.StelleWalSicherMit(context.Background(), 2*time.Second, func(context.Context) error { n++; return fehler })
	if !errors.Is(err, fehler) || n != 1 {
		t.Fatalf("Versuche %d, Fehler %v", n, err)
	}
}

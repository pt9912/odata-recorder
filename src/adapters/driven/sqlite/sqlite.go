// Package sqlite ist der Driven Adapter der Persistenz: Er öffnet die SQLite-Datei und
// bringt sie auf die Schemaversion des Programms (SPEC-103). Dieses Paket bindet die
// SQLite-Bibliothek ein; die tech-Regel in .a-check.yml hält sie aus den übrigen
// Adaptern und dem Kern, die Composition Root nimmt a-check aus.
package sqlite

import (
	"context"
	"database/sql"
	_ "embed" // bettet die Migration ein
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	sqlitedriver "modernc.org/sqlite"
)

// Schemaversion ist die Version, auf die Open eine Datei bringt.
const Schemaversion = 1

// Fehler von Open. ErrBeschaedigt ist der Fall von ODR-E3001 (SPEC-018).
var (
	ErrPfad            = errors.New("sqlite: ungültiger Pfad")
	ErrBeschaedigt     = errors.New("sqlite: Datei ist keine lesbare SQLite-Datenbank")
	ErrSchemaNeuer     = errors.New("sqlite: Schemaversion der Datei ist neuer als bekannt")
	ErrFremdeDatenbank = errors.New("sqlite: Datei enthält eine fremde Datenbank")
)

//go:embed migrations/0001_init.sql
var migration0001 string

// Options steuert Open. Ein Busy-Timeout von 0 bedeutet SQLITE_BUSY_TIMEOUT (SPEC-014).
type Options struct {
	BusyTimeout time.Duration
}

const standardBusyTimeout = 5000 * time.Millisecond

// DB ist eine geöffnete, migrierte Datenbank.
type DB struct {
	sql *sql.DB
}

// Open öffnet die Datei unter path und migriert sie (SPEC-103).
func Open(ctx context.Context, path string, opt Options) (*DB, error) {
	if err := pruefePfad(path); err != nil {
		return nil, err
	}
	busy := opt.BusyTimeout
	if busy <= 0 {
		busy = standardBusyTimeout
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPfad, err)
	}
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busy.Milliseconds()))
	dsn := (&url.URL{Scheme: "file", Path: abs, RawQuery: q.Encode()}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: öffnen: %w", err)
	}
	d := &DB{sql: db}
	if err := d.migrate(ctx, migration0001); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := stelleWalSicher(ctx, busy, d.setzeWal); err != nil {
		_ = db.Close()
		return nil, err
	}
	return d, nil
}

// stelleWalSicher stellt journal_mode=WAL her. SQLite meldet SQLITE_BUSY beim Umstellen
// auch dann sofort, wenn der busy_timeout läuft, sobald ein anderer Öffner dieselbe
// Sperre hält; deshalb wiederholt sich das Umstellen bis zum Ablauf der Frist.
func stelleWalSicher(ctx context.Context, frist time.Duration, versuch func(context.Context) error) error {
	ende := time.Now().Add(frist)
	for {
		err := versuch(ctx)
		if err == nil || !istBelegt(err) || time.Now().After(ende) {
			return klassifiziere(err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (d *DB) setzeWal(ctx context.Context) error {
	var modus string
	return d.sql.QueryRowContext(ctx, "PRAGMA journal_mode = wal").Scan(&modus)
}

func istBelegt(err error) bool {
	var se *sqlitedriver.Error
	if !errors.As(err, &se) {
		return false
	}
	const busy, locked = 5, 6
	c := se.Code() & 0xff
	return c == busy || c == locked
}

// Close schließt die Datenbank; ein zweites Close ist kein Fehler.
func (d *DB) Close() error {
	return d.sql.Close()
}

func pruefePfad(path string) error {
	if path == "" || strings.HasPrefix(path, ":memory:") {
		return fmt.Errorf("%w: %q", ErrPfad, path)
	}
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return fmt.Errorf("%w: %q ist ein Verzeichnis", ErrPfad, path)
	}
	if fi, err := os.Stat(filepath.Dir(path)); err != nil || !fi.IsDir() {
		return fmt.Errorf("%w: Verzeichnis von %q fehlt", ErrPfad, path)
	}
	return nil
}

// klassifiziere macht aus einem Treiberfehler ErrBeschaedigt, wenn die Datei keine
// SQLite-Datenbank ist.
func klassifiziere(err error) error {
	var se *sqlitedriver.Error
	if errors.As(err, &se) {
		const notADB, corrupt = 26, 11
		if c := se.Code() & 0xff; c == notADB || c == corrupt {
			return fmt.Errorf("%w: %v", ErrBeschaedigt, err)
		}
	}
	return err
}

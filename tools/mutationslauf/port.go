package main

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// Bau beschreibt einen Aufruf von `go test -c` (SPEC-210 Nr. 5 b).
type Bau struct {
	Dir      string // Wurzel des Moduls der Kopie
	Paket    string // Paket relativ zu Dir, mit führendem "./"
	Ziel     string // Ablage des Test-Binary
	Detektor bool   // mit -race und cgo
}

// Start beschreibt den Start eines Test-Binary (SPEC-210 Nr. 5 c).
type Start struct {
	Dir       string   // Arbeitsverzeichnis: das Paketverzeichnis der Kopie
	Binary    string   // das Test-Binary
	Args      []string // Argumente des Binary
	Detektor  bool     // das Binary stammt aus einem Bau mit -race
	Nullprobe bool     // Start der Nullprobe (zählt nicht in `prozesse`)
	Frist     time.Duration
}

// Ausgang ist das Ergebnis eines Baus oder Starts.
type Ausgang struct {
	Code        int    // Exit-Code; ungleich 0 bei Signal
	Text        string // Anfang der Ausgabe (stdout und stderr)
	Abgebrochen bool   // das Werkzeug selbst beendete den Prozess
	KeinBinary  bool   // der Bau gelang, aber es entstand kein Binary (Paket ohne Testdatei)
}

// Prozesse ist der Port zu Go-Werkzeug, Test-Binaries und Cache.
type Prozesse interface {
	Bauen(ctx context.Context, b Bau) Ausgang
	Starten(ctx context.Context, s Start) Ausgang
	CacheGroesse() int64
	CacheLeeren(ctx context.Context) error
}

// Modul benennt ein Go-Modul des Baums; Dir ist relativ zur Wurzel, "" ist das Wurzelmodul.
type Modul struct{ Dir string }

// Dateien ist der Port zu Wegwerfkopien und ihrem Entfernen (SPEC-210 Nr. 4).
type Dateien interface {
	Kopieren(baum string, m Modul, ziel string) error
	Entfernen(pfad string) error
}

// vorhanden sagt, ob ein Pfad existiert.
func vorhanden(pfad string) bool {
	_, err := os.Lstat(pfad)
	return err == nil
}

// unter verbindet eine Wurzel mit einem Pfad in Schrägstrich-Schreibweise.
func unter(wurzel, rel string) string { return filepath.Join(wurzel, filepath.FromSlash(rel)) }

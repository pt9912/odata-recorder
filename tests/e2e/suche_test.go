package e2e_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// fund ist eine Stelle, an der die Bytes eines Wertes stehen.
type fund struct{ ort, wert string }

// durchsuche liest jede Datei unter dir (rekursiv, auch -wal, -shm und Blobs, ungeachtet der Größe) auf
// Rohbytes und nennt die Stellen der Werte (SPEC-037 Nr. 9); die Rückgabe nennt auch die gelesenen Dateien.
func durchsuche(t *testing.T, dir string, werte []string) (dateien []string, funde []fund) {
	t.Helper()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		dateien = append(dateien, rel)
		funde = append(funde, inBytes(rel, b, werte)...)
		return nil
	})
	if err != nil {
		t.Fatalf("Suche unter %s: %v", dir, err)
	}
	return dateien, funde
}

// inBytes nennt die Werte, die b trägt, mit dem Ort ort.
func inBytes(ort string, b []byte, werte []string) (funde []fund) {
	for _, w := range werte {
		if bytes.Contains(b, []byte(w)) {
			funde = append(funde, fund{ort, w})
		}
	}
	return funde
}

// sucheStroeme sucht in stdout und stderr des Binary.
func sucheStroeme(p *prozess, werte []string) []fund {
	return append(inBytes("stdout", p.aus.bytes(), werte), inBytes("stderr", p.fehl.bytes(), werte)...)
}

// hat sagt, ob unter den Funden der Wert am Ort steht (Ort als Teil des Namens).
func hat(funde []fund, ort, wert string) bool {
	for _, f := range funde {
		if f.wert == wert && bytes.Contains([]byte(f.ort), []byte(ort)) {
			return true
		}
	}
	return false
}

// kontrollDienst ist SERVICE_ID der Stufe Schutz und zugleich der Kontrollwert auf stderr: jeder Eintrag
// eines Fehlers trägt `service_id=<Wert>`.
const kontrollDienst = "KONTROLLE-STDERR-6b1d"

// kontrollDateien sind die Kontrollwerte unter DATA_DIR (SPEC-037 Nr. 9): je Ort der Datei ein Marker;
// `-wal`, `-shm`, ein Unterverzeichnis und eine Datei von 3 MiB mit dem Marker am Ende gehören zur Menge.
func kontrollDateien() map[string]string {
	return map[string]string{
		"kontrolle.txt": "KONTROLLE-DATEI-1a", filepath.Join("tief", "unten", "kontrolle.bin"): "KONTROLLE-TIEF-2b",
		"kontrolle.db-wal": "KONTROLLE-WAL-3c", "kontrolle.db-shm": "KONTROLLE-SHM-4d", "gross.bin": "KONTROLLE-GROSS-5e",
	}
}

// pflanzeKontrolle legt die Kontrolldateien unter dir ab.
func pflanzeKontrolle(t *testing.T, dir string) {
	t.Helper()
	for rel, marker := range kontrollDateien() {
		inhalt := []byte(marker)
		if rel == "gross.bin" {
			inhalt = append(make([]byte, 3<<20), marker...)
		}
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, inhalt, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// werteKontrolle sind die Marker der Dateien und der Kontrollwert von stderr.
func werteKontrolle() []string {
	w := []string{kontrollDienst}
	for _, m := range kontrollDateien() {
		w = append(w, m)
	}
	return w
}

// pruefeKontrolle meldet jeden Kontrollwert, den die Suche nicht fand: ohne ihn ist ihr Leerstand wertlos.
func pruefeKontrolle(t *testing.T, name string, dateien []fund, stderr []fund) {
	t.Helper()
	for rel, marker := range kontrollDateien() {
		if !hat(dateien, rel, marker) {
			t.Errorf("%s: die Suche findet den Kontrollwert %s in %s nicht", name, marker, rel)
		}
	}
	if !hat(stderr, "stderr", kontrollDienst) {
		t.Errorf("%s: die Suche findet den Kontrollwert %s auf stderr nicht", name, kontrollDienst)
	}
}

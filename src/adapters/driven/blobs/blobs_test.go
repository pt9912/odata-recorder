package blobs_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/blobs"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

func dateienUnter(t *testing.T, wurzel string) []string {
	t.Helper()
	var namen []string
	err := filepath.WalkDir(wurzel, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			r, _ := filepath.Rel(wurzel, p)
			namen = append(namen, r)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(namen)
	return namen
}

func neueAblage(t *testing.T) outboundtest.BlobAblage {
	t.Helper()
	dir := t.TempDir()
	s, err := blobs.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	pfad := func(ref string) string { return filepath.Join(dir, "sha256", ref[len("sha256:"):]) }
	return outboundtest.BlobAblage{
		Blobs:   s,
		Dateien: func(t *testing.T) []string { return dateienUnter(t, dir) },
		Ueberschreibe: func(t *testing.T, ref string, inhalt []byte) {
			t.Helper()
			if err := os.WriteFile(pfad(ref), inhalt, 0o600); err != nil {
				t.Fatal(err)
			}
		},
		ErsetzeDurchVerzeichnis: func(t *testing.T, ref string) {
			t.Helper()
			if err := os.Remove(pfad(ref)); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(pfad(ref), 0o700); err != nil {
				t.Fatal(err)
			}
		},
		LegeFremdesAb: func(t *testing.T) {
			t.Helper()
			if err := os.WriteFile(filepath.Join(dir, "sha256", "fremd.txt"), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
	}
}

func TestBlobVertrag(t *testing.T) { outboundtest.RunBlobVertrag(t, neueAblage) }

func TestRechteUnabhaengigVonUmask(t *testing.T) {
	alt := syscall.Umask(0)
	defer syscall.Umask(alt)
	dir := t.TempDir()
	s, err := blobs.New(filepath.Join(dir))
	if err != nil {
		t.Fatal(err)
	}
	w, err := s.Begin(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("a"))
	// Der offene Writer liegt als temporäre Datei vor.
	tmp := dateienUnter(t, filepath.Join(dir, "tmp"))
	if len(tmp) != 1 {
		t.Fatalf("temporäre Dateien %v", tmp)
	}
	if m := modus(t, filepath.Join(dir, "tmp", tmp[0])); m != 0o600 {
		t.Fatalf("temporäre Datei %o", m)
	}
	ref, _, err := w.Commit()
	if err != nil {
		t.Fatal(err)
	}
	if m := modus(t, filepath.Join(dir, "sha256", ref[len("sha256:"):])); m != 0o600 {
		t.Fatalf("Blob %o", m)
	}
	for _, d := range []string{"sha256", "tmp"} {
		if m := modus(t, filepath.Join(dir, d)); m != 0o700 {
			t.Fatalf("Verzeichnis %s %o", d, m)
		}
	}
}

func modus(t *testing.T, p string) os.FileMode {
	t.Helper()
	i, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return i.Mode().Perm()
}

func TestErzeugenLoeschtReste(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "tmp"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tmp", "rest.tmp"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := blobs.New(dir); err != nil {
		t.Fatal(err)
	}
	if d := dateienUnter(t, dir); len(d) != 0 {
		t.Fatalf("Dateien %v", d)
	}
}

func TestErzeugenMitUngueltigemVerzeichnis(t *testing.T) {
	dir := t.TempDir()
	datei := filepath.Join(dir, "datei")
	if err := os.WriteFile(datei, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	blockiert := filepath.Join(dir, "blockiert")
	if err := os.Mkdir(blockiert, 0o700); err != nil {
		t.Fatal(err)
	}
	// Eine Datei an der Stelle des Unterverzeichnisses lässt es nicht anlegen.
	if err := os.WriteFile(filepath.Join(blockiert, "sha256"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, pfad := range map[string]string{
		"leer": "", "fehlt": filepath.Join(dir, "gibt-es-nicht"), "Datei": datei, "Unterverzeichnis blockiert": blockiert,
	} {
		if _, err := blobs.New(pfad); err == nil {
			t.Fatalf("%s: New gelang", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "gibt-es-nicht")); err == nil {
		t.Fatal("New legte das Verzeichnis an")
	}
	if d := dateienUnter(t, blockiert); len(d) != 1 {
		t.Fatalf("Dateien %v", d)
	}
}

package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// echteDateien legt Wegwerfkopien im Dateisystem an (SPEC-210 Nr. 4).
type echteDateien struct{}

// Kopieren legt die Kopie eines Moduls an: beim Wurzelmodul go.mod, go.sum, src und tests, sonst das
// ganze Verzeichnis des Moduls; verschachtelte Module sind nicht Teil der Kopie (SPEC-210 Nr. 4 a).
func (echteDateien) Kopieren(baum string, m Modul, ziel string) error {
	quelle := unter(baum, m.Dir)
	if err := os.MkdirAll(ziel, 0o755); err != nil {
		return err
	}
	if m.Dir != "" {
		return kopiereBaum(quelle, ziel, quelle)
	}
	for _, f := range []string{"go.mod", "go.sum"} {
		if !vorhanden(filepath.Join(quelle, f)) {
			continue
		}
		if err := kopiereBaum(filepath.Join(quelle, f), filepath.Join(ziel, f), quelle); err != nil {
			return err
		}
	}
	for _, d := range []string{"src", "tests"} {
		if err := kopiereBaum(filepath.Join(quelle, d), filepath.Join(ziel, d), quelle); err != nil {
			return err
		}
	}
	return nil
}

// kopiereBaum kopiert Datei oder Verzeichnis; ein Verzeichnis mit eigener go.mod außer dem Modul der
// Wurzel `modul` bleibt draußen.
func kopiereBaum(von, nach, modul string) error {
	return filepath.WalkDir(von, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && p == von {
				return nil
			}
			return err
		}
		rel, _ := filepath.Rel(von, p)
		z := filepath.Join(nach, rel)
		switch {
		case d.IsDir():
			if p != modul && vorhanden(filepath.Join(p, "go.mod")) {
				return fs.SkipDir
			}
			return os.MkdirAll(z, 0o755)
		case d.Type()&fs.ModeSymlink != 0:
			l, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(l, z)
		case d.Type().IsRegular():
			return kopiereDatei(p, z)
		}
		return nil
	})
}

func kopiereDatei(von, nach string) error {
	i, err := os.Stat(von)
	if err != nil {
		return err
	}
	in, err := os.Open(von)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(nach, os.O_CREATE|os.O_WRONLY|os.O_EXCL, i.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// Entfernen löscht eine Kopie oder ein Binary samt Inhalt; ein fehlender Pfad ist kein Fehler.
func (echteDateien) Entfernen(pfad string) error { return os.RemoveAll(pfad) }

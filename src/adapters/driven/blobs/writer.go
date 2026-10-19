package blobs

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"os"
	"path/filepath"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

type zustand int

const (
	offen zustand = iota
	committet
	abgebrochen
)

// writer schreibt in eine temporäre Datei und macht sie mit Commit unter dem Hash sichtbar.
type writer struct {
	s *Store
	// ctxFehler fragt den Kontext von Begin ab, an den der Writer gebunden ist (SPEC-106 Nr. 1).
	ctxFehler func() error
	f         *os.File
	h         hash.Hash
	n         int64
	grenze    int64
	z         zustand
}

func (s *Store) neuerWriter(ctx context.Context, grenze int64) (*writer, error) {
	var zufall [16]byte
	if _, err := rand.Read(zufall[:]); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(s.tmp, hex.EncodeToString(zufall[:])+".tmp"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, dateiMode)
	if err != nil {
		return nil, err
	}
	if err := f.Chmod(dateiMode); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	return &writer{s: s, ctxFehler: ctx.Err, f: f, h: sha256.New(), grenze: grenze}, nil
}

// verwerfe schließt und löscht die temporäre Datei.
func (w *writer) verwerfe() {
	w.z = abgebrochen
	_ = w.f.Close()
	_ = os.Remove(w.f.Name())
}

func (w *writer) Write(p []byte) (int, error) {
	if w.z != offen {
		return 0, outbound.ErrAbgeschlossen
	}
	if err := w.ctxFehler(); err != nil {
		w.verwerfe()
		return 0, err
	}
	if w.n+int64(len(p)) > w.grenze {
		w.verwerfe()
		return 0, outbound.ErrZuGross
	}
	n, err := w.f.Write(p)
	w.h.Write(p[:n])
	w.n += int64(n)
	if err != nil {
		w.verwerfe()
		return n, err
	}
	return n, nil
}

func (w *writer) Commit() (string, int64, error) {
	if w.z != offen {
		return "", 0, outbound.ErrAbgeschlossen
	}
	if err := w.ctxFehler(); err != nil {
		w.verwerfe()
		return "", 0, err
	}
	h := hex.EncodeToString(w.h.Sum(nil))
	if err := w.sichere(filepath.Join(w.s.blobs, h)); err != nil {
		w.verwerfe()
		return "", 0, err
	}
	w.z = committet
	return praefix + h, w.n, nil
}

// sichere schreibt die Datei auf den Datenträger und benennt sie atomar um; eine vorhandene
// Datei gleichen Namens wird dabei durch die geprüften Bytes ersetzt.
func (w *writer) sichere(ziel string) error {
	if err := w.f.Sync(); err != nil {
		return err
	}
	if err := w.f.Close(); err != nil {
		return err
	}
	if i, err := os.Lstat(ziel); err == nil && !i.Mode().IsRegular() {
		if err := os.RemoveAll(ziel); err != nil {
			return err
		}
	}
	if err := os.Rename(w.f.Name(), ziel); err != nil {
		return err
	}
	d, err := os.Open(filepath.Dir(ziel))
	if err != nil {
		return err
	}
	return errors.Join(d.Sync(), d.Close())
}

func (w *writer) Abort() error {
	if w.z == offen {
		w.verwerfe()
	}
	return nil
}

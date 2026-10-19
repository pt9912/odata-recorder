// Package blobs legt Bodies content-adressiert als Dateien ab (SPEC-106, ARC-010).
package blobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

const (
	praefix   = "sha256:"
	hexLaenge = 64
	dateiMode = 0o600
	dirMode   = 0o700
	// Unterverzeichnisse des Datenverzeichnisses; die temporären Dateien liegen im selben
	// Dateisystem wie die Blobs, damit das Umbenennen atomar ist.
	blobDir = "sha256"
	tmpDir  = "tmp"
)

// Store ist das BlobRepository auf einem Verzeichnis. Ein Verzeichnis gehört zu einem Store
// zur Zeit.
type Store struct {
	blobs string
	tmp   string
}

var _ outbound.BlobRepository = (*Store)(nil)

// New öffnet die Ablage in dir. Das Verzeichnis legt der Aufrufer an; New legt nur seine
// Unterverzeichnisse an und löscht übrig gebliebene temporäre Dateien.
func New(dir string) (*Store, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("blobs: %w", err)
	}
	s := &Store{blobs: filepath.Join(dir, blobDir), tmp: filepath.Join(dir, tmpDir)}
	for _, d := range []string{s.blobs, s.tmp} {
		if err := os.MkdirAll(d, dirMode); err != nil {
			return nil, fmt.Errorf("blobs: %w", err)
		}
		if err := os.Chmod(d, dirMode); err != nil {
			return nil, fmt.Errorf("blobs: %w", err)
		}
	}
	if err := s.raeumeRestAuf(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) raeumeRestAuf() error {
	eintraege, err := os.ReadDir(s.tmp)
	if err != nil {
		return fmt.Errorf("blobs: %w", err)
	}
	for _, e := range eintraege {
		if err := os.Remove(filepath.Join(s.tmp, e.Name())); err != nil {
			return fmt.Errorf("blobs: %w", err)
		}
	}
	return nil
}

// hexVon liefert die Hexziffern einer gültigen Referenz.
func hexVon(ref string) (string, bool) {
	if !strings.HasPrefix(ref, praefix) || len(ref) != len(praefix)+hexLaenge {
		return "", false
	}
	h := ref[len(praefix):]
	for i := 0; i < len(h); i++ {
		if c := h[i]; (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return "", false
		}
	}
	return h, true
}

// pfadVon prüft die Referenz vor dem Kontext (SPEC-106) und liefert den Dateipfad.
func (s *Store) pfadVon(ctx context.Context, ref string) (string, error) {
	h, ok := hexVon(ref)
	if !ok {
		return "", errors.Join(model.ErrUngueltig, fmt.Errorf("Blob-Referenz %q", ref))
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return filepath.Join(s.blobs, h), nil
}

// Begin eröffnet einen Writer.
func (s *Store) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	if grenze < 1 {
		return nil, errors.Join(model.ErrUngueltig, fmt.Errorf("Grenze %d", grenze))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.neuerWriter(ctx, grenze)
}

// Open liefert den Leser einer Datei, deren Hash vollständig geprüft ist.
func (s *Store) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	pfad, err := s.pfadVon(ctx, ref)
	if err != nil {
		return nil, err
	}
	return pruefeDatei(pfad, ref[len(praefix):])
}

// Verify prüft die Datei wie Open.
func (s *Store) Verify(ctx context.Context, ref string) error {
	r, err := s.Open(ctx, ref)
	if err != nil {
		return err
	}
	return r.Close()
}

// pruefeDatei öffnet die Datei, vergleicht ihren Hash und setzt den Leser auf den Anfang;
// der Leser ist dieselbe geöffnete Datei, die geprüft wurde.
func pruefeDatei(pfad, soll string) (io.ReadCloser, error) {
	info, err := os.Lstat(pfad)
	if errors.Is(err, os.ErrNotExist) {
		return nil, outbound.ErrBlobFehlt
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.Join(outbound.ErrBlobBeschaedigt, errors.New("keine Datei"))
	}
	f, err := os.Open(pfad)
	if errors.Is(err, os.ErrNotExist) {
		return nil, outbound.ErrBlobFehlt
	}
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		_ = f.Close()
		return nil, err
	}
	if hex.EncodeToString(h.Sum(nil)) != soll {
		_ = f.Close()
		return nil, errors.Join(outbound.ErrBlobBeschaedigt, errors.New("Hash weicht ab"))
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}

// List liefert die Referenzen der Blob-Dateien; os.ReadDir sortiert nach Dateiname.
func (s *Store) List(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	eintraege, err := os.ReadDir(s.blobs)
	if err != nil {
		return nil, err
	}
	refs := []string{}
	for _, e := range eintraege {
		if _, ok := hexVon(praefix + e.Name()); ok && e.Type().IsRegular() {
			refs = append(refs, praefix+e.Name())
		}
	}
	return refs, nil
}

// Remove löscht die Datei einer Referenz.
func (s *Store) Remove(ctx context.Context, ref string) error {
	pfad, err := s.pfadVon(ctx, ref)
	if err != nil {
		return err
	}
	if err := os.Remove(pfad); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

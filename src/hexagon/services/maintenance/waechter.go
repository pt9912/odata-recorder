package maintenance

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// StandardKarenz ist der Wert, den der Composition Root dem Wächter übergibt (SPEC-123 Nr. 1, 8b).
const StandardKarenz = 10 * time.Minute

var _ outbound.BlobRepository = (*BlobWaechter)(nil)

// BlobWaechter ist ein BlobRepository vor dem Adapter: er reicht jeden Aufruf unverändert durch,
// ordnet Commit (Leser) und den Abgleich der Wartung (Schreiber) über eine gemeinsame Sperre und
// merkt die Zeit jedes gelungenen Commit (SPEC-123 Nr. 9). Er schützt nur, was durch ihn läuft.
type BlobWaechter struct {
	unten   outbound.BlobRepository
	uhr     support.Clock
	karenz  time.Duration
	sperre  sync.RWMutex
	mu      sync.Mutex
	commits map[string]time.Time
}

// NewBlobWaechter prüft die Argumente (SPEC-123 Nr. 1): nil-Werte und eine Karenz kleiner oder
// gleich 0 sind ErrUngueltig.
func NewBlobWaechter(blobs outbound.BlobRepository, uhr support.Clock, karenz time.Duration) (*BlobWaechter, error) {
	if blobs == nil || uhr == nil || karenz <= 0 {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Wächter ohne Blobs, Uhr oder mit Karenz kleiner oder gleich 0"))
	}
	return &BlobWaechter{unten: blobs, uhr: uhr, karenz: karenz, commits: map[string]time.Time{}}, nil
}

// Begin reicht den Aufruf durch und liefert einen Writer, dessen Commit die Lesersperre hält.
func (w *BlobWaechter) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	inner, err := w.unten.Begin(ctx, grenze)
	if err != nil {
		return inner, err
	}
	return &waechterWriter{CaptureWriter: inner, w: w}, nil
}

// Open reicht den Aufruf unverändert durch.
func (w *BlobWaechter) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	return w.unten.Open(ctx, ref)
}

// Verify reicht den Aufruf unverändert durch.
func (w *BlobWaechter) Verify(ctx context.Context, ref string) error { return w.unten.Verify(ctx, ref) }

// List reicht den Aufruf unverändert durch.
func (w *BlobWaechter) List(ctx context.Context) ([]string, error) { return w.unten.List(ctx) }

// Remove reicht den Aufruf unverändert durch.
func (w *BlobWaechter) Remove(ctx context.Context, ref string) error { return w.unten.Remove(ctx, ref) }

type waechterWriter struct {
	outbound.CaptureWriter
	w *BlobWaechter
}

// Commit hält die Lesersperre von vor dem Commit des Writers darunter bis nach dem Merken der Zeit.
func (c *waechterWriter) Commit() (string, int64, error) {
	c.w.sperre.RLock()
	defer c.w.sperre.RUnlock()
	ref, laenge, err := c.CaptureWriter.Commit()
	if err == nil {
		c.w.merke(ref)
	}
	return ref, laenge, err
}

// merke hält die Zeit des Commit und entfernt die Referenzen, deren Karenz zu dieser Zeit abgelaufen ist.
func (w *BlobWaechter) merke(ref string) {
	jetzt := w.uhr.Now()
	w.mu.Lock()
	defer w.mu.Unlock()
	w.commits[ref] = jetzt
	for r, z := range w.commits {
		if !jetzt.Before(z.Add(w.karenz)) {
			delete(w.commits, r)
		}
	}
}

// geschuetzt sagt, ob jetzt echt vor der Zeit des Commit von ref plus Karenz liegt.
func (w *BlobWaechter) geschuetzt(ref string, jetzt time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	z, ok := w.commits[ref]
	return ok && jetzt.Before(z.Add(w.karenz))
}

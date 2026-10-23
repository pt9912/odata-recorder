package maintenance

import "time"

// Hilfen für die externen Tests: die Sperre des Wächters und die gemerkte Zeit je Referenz.

// TryRLock nimmt die Lesersperre, wenn sie frei ist.
func (w *BlobWaechter) TryRLock() bool { return w.sperre.TryRLock() }

// RUnlock gibt eine mit TryRLock genommene Lesersperre frei.
func (w *BlobWaechter) RUnlock() { w.sperre.RUnlock() }

// TryLock nimmt die Schreibersperre, wenn sie frei ist.
func (w *BlobWaechter) TryLock() bool { return w.sperre.TryLock() }

// Unlock gibt eine mit TryLock genommene Schreibersperre frei.
func (w *BlobWaechter) Unlock() { w.sperre.Unlock() }

// Gemerkt liefert eine Kopie der gemerkten Zeit je Referenz.
func (w *BlobWaechter) Gemerkt() map[string]time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make(map[string]time.Time, len(w.commits))
	for ref, z := range w.commits {
		out[ref] = z
	}
	return out
}

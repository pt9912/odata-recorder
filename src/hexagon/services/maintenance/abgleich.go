package maintenance

import (
	"context"
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// blobAbgleich ist Stufe 4 (SPEC-123 Nr. 7): unter der Sperre des Wächters zuerst die
// referenzierten Dateien prüfen, dann die Dateien ohne Referenz entfernen. Der Start kennt keine
// Karenz, denn beim Start läuft kein Schreiber.
func (w *Wartung) blobAbgleich(ctx context.Context, b *inbound.Wartungsbericht) error {
	w.k.Blobs.sperre.Lock()
	defer w.k.Blobs.sperre.Unlock()
	if err := w.pruefeBlobs(ctx, b); err != nil {
		return err
	}
	n, err := w.abgleich(ctx, nil)
	b.BlobsEntfernt = n
	return err
}

// pruefeBlobs prüft jede Referenz mit Verify (SPEC-123 Nr. 7a): eine fehlende oder beschädigte
// Datei wird gemeldet und nicht entfernt, jeder andere Fehler beendet die Stufe.
func (w *Wartung) pruefeBlobs(ctx context.Context, b *inbound.Wartungsbericht) error {
	refs, err := w.k.Speicher.ReferencedBlobs(ctx)
	if err != nil {
		return mitCode(err)
	}
	for _, ref := range refs {
		err := w.k.Blobs.Verify(ctx, ref)
		switch {
		case err == nil:
		case kontextFehler(err):
			return err
		case errors.Is(err, outbound.ErrBlobFehlt):
			b.BlobsFehlend = append(b.BlobsFehlend, ref)
			w.protokolliereBlob(ctx, "blob_fehlt", ref)
		case errors.Is(err, outbound.ErrBlobBeschaedigt):
			b.BlobsBeschaedigt = append(b.BlobsBeschaedigt, ref)
			w.protokolliereBlob(ctx, "blob_beschaedigt", ref)
		default:
			return mitCode(err)
		}
	}
	return nil
}

// abgleich entfernt die Dateien ohne Referenz (SPEC-123 Nr. 7b) und liefert die Zahl der
// entfernten, auch bei einem Fehler. geschuetzt, wenn gesetzt, nimmt Kandidaten aus. Der Aufrufer
// hält die Sperre des Wächters.
func (w *Wartung) abgleich(ctx context.Context, geschuetzt func(string) bool) (int, error) {
	vorhanden, err := w.k.Blobs.List(ctx)
	if err != nil {
		return 0, mitCode(err)
	}
	refs, err := w.k.Speicher.ReferencedBlobs(ctx)
	if err != nil {
		return 0, mitCode(err)
	}
	referenziert := menge(refs)
	kandidaten := ohne(vorhanden, func(ref string) bool {
		return referenziert[ref] || geschuetzt != nil && geschuetzt(ref)
	})
	neu, err := w.k.Speicher.ReferencedBlobs(ctx)
	if err != nil {
		return 0, mitCode(err)
	}
	jetzt := menge(neu)
	entfernt := 0
	for _, ref := range ohne(kandidaten, func(ref string) bool { return jetzt[ref] }) {
		if err := w.k.Blobs.Remove(ctx, ref); err != nil {
			return entfernt, mitCode(err)
		}
		entfernt++
	}
	return entfernt, nil
}

func menge(liste []string) map[string]bool {
	m := make(map[string]bool, len(liste))
	for _, e := range liste {
		m[e] = true
	}
	return m
}

// ohne liefert die Elemente von liste in ihrer Folge, auf die weg nicht zutrifft.
func ohne(liste []string, weg func(string) bool) []string {
	var rest []string
	for _, e := range liste {
		if !weg(e) {
			rest = append(rest, e)
		}
	}
	return rest
}

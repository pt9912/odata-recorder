package outbound

import (
	"context"
	"io"
)

// BlobRepository legt Body-Bytes content-adressiert ab (SPEC-106). Eine Referenz ist
// "sha256:" und 64 Kleinbuchstaben-Hexziffern des SHA-256 der Bytes.
type BlobRepository interface {
	// Begin eröffnet einen Writer für höchstens grenze Bytes; er ist an ctx gebunden.
	Begin(ctx context.Context, grenze int64) (CaptureWriter, error)
	// Open prüft den Hash der Datei vollständig und liefert dann ihren Leser.
	Open(ctx context.Context, ref string) (io.ReadCloser, error)
	// Verify prüft wie Open, ohne einen Leser zu liefern.
	Verify(ctx context.Context, ref string) error
	// List liefert die Referenzen aller vorhandenen Blobs aufsteigend.
	List(ctx context.Context) ([]string, error)
	// Remove löscht einen Blob; eine unbekannte Referenz ist kein Fehler.
	Remove(ctx context.Context, ref string) error
}

// CaptureWriter nimmt die Bytes eines Body entgegen und hat genau einen Abschluss.
type CaptureWriter interface {
	io.Writer
	// Commit macht den Blob sichtbar und liefert Referenz und Länge.
	Commit() (ref string, laenge int64, err error)
	// Abort verwirft die Bytes; nach einem Abschluss ist es wirkungslos.
	Abort() error
}

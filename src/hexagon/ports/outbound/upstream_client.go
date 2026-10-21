package outbound

import (
	"context"
	"io"
)

// UpstreamClient sendet einen Request an den Upstream eines Service und liefert dessen Antwort
// als Strom (SPEC-120). Der Port trägt keinen Typ einer HTTP-Bibliothek; welche Header gehen,
// entscheidet der Aufrufer.
type UpstreamClient interface {
	// Senden kehrt mit dem Kopf der Antwort zurück; den Body schließt der Aufrufer. Die Fehler und
	// ihre Rangfolge stehen in SPEC-120 Nr. 8.
	Senden(ctx context.Context, anfrage UpstreamAnfrage) (UpstreamAntwort, error)
}

// UpstreamAnfrage ist ein Request an den Upstream (SPEC-120): Ziel ist die UpstreamURL des
// Service, Pfad der OData-Pfad, RawQuery die Query ohne `?`; Laenge -1 heißt unbekannt, 0 kein Body.
type UpstreamAnfrage struct {
	Ziel     string
	Methode  string
	Pfad     string
	RawQuery string
	Header   map[string][]string
	Body     io.Reader
	Laenge   int64
}

// UpstreamAntwort ist der Kopf einer Antwort des Upstreams und ihr Body als Strom (SPEC-120 Nr. 5).
type UpstreamAntwort struct {
	Status int
	Header map[string][]string
	Body   io.ReadCloser
}

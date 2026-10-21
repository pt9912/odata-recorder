package httpupstream

import (
	"context"
	"io"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// antwort liefert Status, Header und Body der Antwort, wie die Bibliothek sie liest (SPEC-120 Nr. 5).
func antwort(ctx context.Context, res *http.Response, ende context.CancelFunc) outbound.UpstreamAntwort {
	return outbound.UpstreamAntwort{
		Status: res.StatusCode,
		Header: res.Header,
		Body:   &antwortBody{body: res.Body, ctxFehler: ctx.Err, ende: ende},
	}
}

// antwortBody ist der Body der Antwort. Nach einem Abbruch von ctx endet jedes Read mit einem
// Fehler, der nicht io.EOF ist; ein laufendes Read beendet die Bibliothek, weil der Request am
// Kontext des Austauschs hängt. Close beendet diesen Kontext und mit ihm das Lesen aus dem
// Request-Body (SPEC-120 Nr. 4, 5).
type antwortBody struct {
	body      io.ReadCloser
	ctxFehler func() error
	ende      context.CancelFunc
}

func (b *antwortBody) Read(p []byte) (int, error) {
	if err := b.ctxFehler(); err != nil {
		return 0, err
	}
	return b.body.Read(p)
}

func (b *antwortBody) Close() error {
	err := b.body.Close()
	b.ende()
	return err
}

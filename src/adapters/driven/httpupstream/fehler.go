package httpupstream

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Kein Fehler dieses Pakets nennt `Methode`, `Pfad`, `RawQuery`, einen Header-Namen oder -Wert, ein
// Byte eines Bodys oder der Antwort (Status-Text, Header-Werte, Body) oder von `Ziel` mehr als Host
// und Port, und keiner trägt den Text eines Fehlers der Bibliothek oder des Request-Bodys (SPEC-120
// Nr. 8).

// ungueltig ist model.ErrUngueltig mit dem Namen des Felds (SPEC-120 Nr. 2).
func ungueltig(feld string) error {
	return fmt.Errorf("httpupstream: %w: %s", model.ErrUngueltig, feld)
}

// kontextFehler verpackt den Fehler von ctx; errors.Is erreicht ihn.
func kontextFehler(err error) error {
	return fmt.Errorf("httpupstream: %w", err)
}

// bodyFehler ist ein Fehler, den der Request-Body selbst lieferte (SPEC-120 Nr. 4); errors.Is
// erreicht ihn, sein Text erscheint nicht.
type bodyFehler struct {
	err error
}

func (*bodyFehler) Error() string { return "httpupstream: Fehler beim Lesen des Request-Bodys" }

func (e *bodyFehler) Unwrap() error { return e.err }

// ergebnis wendet die Rangfolge aus SPEC-120 Nr. 8 an. Ohne Kopf: rang. Mit Kopf zählt erst ctx,
// dann ob das Beenden der Frist gelang (Nr. 6), dann 101 (Nr. 5); in diesen Fällen schließt der
// Client den Body. Jeder Fehler beendet den Kontext des Austauschs.
func ergebnis(ctx context.Context, res *http.Response, err error, gestoppt bool, l *leser,
	ende context.CancelFunc,
) (outbound.UpstreamAntwort, error) {
	if err != nil {
		ende()
		return outbound.UpstreamAntwort{}, rang(ctx, l, gestoppt)
	}
	var fehler error
	if cerr := ctx.Err(); cerr != nil {
		fehler = kontextFehler(cerr)
	} else if !gestoppt {
		fehler = portFehler(outbound.ErrUpstreamZeit)
	} else if res.StatusCode == http.StatusSwitchingProtocols {
		fehler = portFehler(outbound.ErrUpstreamNichtErreichbar)
	}
	if fehler != nil {
		_ = res.Body.Close()
		ende()
		return outbound.UpstreamAntwort{}, fehler
	}
	return antwort(ctx, res, ende), nil
}

// rang ordnet einen Fehler vor dem Kopf: ctx, dann ein Fehler des Request-Bodys, dann der Ablauf der
// Frist, sonst nicht erreichbar.
func rang(ctx context.Context, l *leser, gestoppt bool) error {
	if err := ctx.Err(); err != nil {
		return kontextFehler(err)
	}
	if err := l.fehler(); err != nil {
		return &bodyFehler{err: err}
	}
	if !gestoppt {
		return portFehler(outbound.ErrUpstreamZeit)
	}
	return portFehler(outbound.ErrUpstreamNichtErreichbar)
}

func portFehler(err error) error {
	return fmt.Errorf("httpupstream: %w", err)
}

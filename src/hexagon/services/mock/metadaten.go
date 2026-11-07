package mock

import (
	"context"
	"io"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// metadaten liefert $metadata (SPEC-198 Nr. 4, 5): den Strom des Blobs unverändert mit dem Content-Type
// application/xml und dem Kennzeichen der Darstellung eines verworfenen Aufrufs von SchreibeAnzahl; bei HEAD ist der
// Body ein leerer Strom. Bei jedem Fehler nach dem Öffnen schließt es den Strom genau einmal.
func (m *Mocker) metadaten(ctx context.Context, head bool, d Dienst, s model.Service) (inbound.Antwort, error) {
	strom, err := m.k.Blobs.Open(ctx, s.MetadataBlob)
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	dar, err := d.Codecs.Schreiber.SchreibeAnzahl(ctx, io.Discard, 0)
	if err != nil {
		_ = strom.Close()
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSonstiges)
	}
	if dar.Kennzeichen.Name == "" || dar.Kennzeichen.Wert == "" {
		_ = strom.Close()
		return inbound.Antwort{}, mitCode(model.CodeE9001, nil)
	}
	body := strom
	if head {
		_ = strom.Close()
		body = io.NopCloser(strings.NewReader(""))
	}
	return inbound.Antwort{
		Status: 200,
		Header: map[string][]string{
			"Content-Type":       {"application/xml"},
			dar.Kennzeichen.Name: {dar.Kennzeichen.Wert},
		},
		Body: body,
	}, nil
}

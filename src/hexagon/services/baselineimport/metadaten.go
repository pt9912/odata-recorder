package baselineimport

import (
	"context"
	"io"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// dokument ist das $metadata der Session (SPEC-127 Nr. 7): die Blob-Referenz, der Exchange, der
// maßgeblich ist, und was der Leser daraus las.
type dokument struct {
	ref     string
	ex      model.Exchange
	modell  model.ServiceModel
	bericht model.ValidierungsBericht
}

// waehleDokument wählt das Dokument aus den Exchanges (Nr. 7 a bis e): ohne Kandidat
// ErrKeineMetadaten, bei mehr als einer Referenz ErrMetadatenAbweichend, bei einem Hash der
// Anfrage ungleich der Referenz ErrMetadatenHash; maßgeblich ist der Kandidat mit der kleinsten
// sequence_no. exs ist nach sequence_no geordnet.
func waehleDokument(exs []model.Exchange, hash string) (model.Exchange, error) {
	var kandidaten []model.Exchange
	for _, e := range exs {
		if u := klassifiziere(e, nil); u.spur == spurMetadaten && u.grund == "" {
			kandidaten = append(kandidaten, e)
		}
	}
	switch {
	case len(kandidaten) == 0:
		return model.Exchange{}, sonst(ErrKeineMetadaten)
	case slices.ContainsFunc(kandidaten, func(e model.Exchange) bool { return e.ResponseBlob != kandidaten[0].ResponseBlob }):
		return model.Exchange{}, sonst(ErrMetadatenAbweichend)
	case hash != "" && hash != kandidaten[0].ResponseBlob:
		return model.Exchange{}, sonst(ErrMetadatenHash)
	}
	return kandidaten[0], nil
}

// oeffne öffnet den Blob, ruft f mit dem Strom und schließt ihn genau einmal, auch bei einem
// Fehler; Fehler von Open und von f werden nach SPEC-109 Nr. 5 mit dem Gegenstand Sonstiges
// abgebildet.
func (i *Importer) oeffne(ctx context.Context, ref string, f func(src io.Reader) error) error {
	src, err := i.k.Blobs.Open(ctx, ref)
	if err != nil {
		return sonst(err)
	}
	err = f(src)
	_ = src.Close()
	if err != nil {
		return sonst(err)
	}
	return nil
}

// liesDokument liest das $metadata der Session mit dem Leser (Nr. 7 f).
func (i *Importer) liesDokument(ctx context.Context, ex model.Exchange) (dokument, error) {
	d := dokument{ref: ex.ResponseBlob, ex: ex}
	err := i.oeffne(ctx, d.ref, func(src io.Reader) error {
		var err error
		d.modell, d.bericht, err = i.k.MetadatenCodec.ParseMetadata(ctx, src, i.k.MaxCaptureBody)
		return err
	})
	return d, err
}

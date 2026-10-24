// Package csdl liest EDMX/CSDL V4 in das Modell des OData-Dienstes (SPEC-125); es ist der Driven
// Adapter des Ports outbound.MetadataCodec und nutzt allein encoding/xml.
package csdl

import (
	"context"
	"errors"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Leser setzt MetadataCodec um. Er hat keinen Zustand: gleichzeitige Aufrufe beeinflussen sich nicht.
type Leser struct{}

// New liefert den Leser.
func New() *Leser { return &Leser{} }

// ParseMetadata liest src bis io.EOF, höchstens grenze+1 Bytes (SPEC-125 Nr. 1 bis 14, grenze ist `max`).
func (*Leser) ParseMetadata(ctx context.Context, src io.Reader, grenze int64) (model.ServiceModel, model.ValidierungsBericht, error) {
	var leer model.ServiceModel
	switch {
	case ctx == nil:
		return leer, model.ValidierungsBericht{}, errors.Join(model.ErrUngueltig, errors.New("ctx fehlt"))
	case src == nil:
		return leer, model.ValidierungsBericht{}, errors.Join(model.ErrUngueltig, errors.New("src fehlt"))
	case grenze < 1:
		return leer, model.ValidierungsBericht{}, errors.Join(model.ErrUngueltig, errors.New("max kleiner 1"))
	}
	if err := ctx.Err(); err != nil {
		return leer, model.ValidierungsBericht{}, err
	}
	l := neuerLauf(ctx, src, grenze)
	if err := l.lies(); err != nil {
		return leer, model.ValidierungsBericht{}, err
	}
	if err := ctx.Err(); err != nil {
		return leer, model.ValidierungsBericht{}, err
	}
	return l.loese()
}

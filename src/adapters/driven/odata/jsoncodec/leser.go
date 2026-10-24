// Package jsoncodec liest OData-V4-JSON-Antworten auf Einzel-Entity- und Collection-Abfragen in typisierte Entities
// (SPEC-126); es ist der Driven Adapter des Ports outbound.ODataCodec und nutzt allein die Standardbibliothek.
package jsoncodec

import (
	"context"
	"errors"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Leser setzt ODataCodec um. Er hat keinen Zustand: gleichzeitige Aufrufe beeinflussen sich nicht.
type Leser struct{}

// New liefert den Leser.
func New() *Leser { return &Leser{} }

// ParseEntity liest die Antwort auf eine Einzel-Entity-Abfrage (SPEC-126).
func (*Leser) ParseEntity(ctx context.Context, src io.Reader, m model.ServiceModel, entitySet string, grenze int64) (model.Entity, error) {
	d, wurzel, err := vorbereiten(ctx, src, m, entitySet, grenze)
	if err != nil {
		return model.Entity{}, err
	}
	e, err := d.einzel(&wurzel)
	if err != nil {
		return model.Entity{}, err
	}
	return e, nil
}

// ParseCollection liest die Antwort auf eine Collection-Abfrage (SPEC-126).
func (*Leser) ParseCollection(ctx context.Context, src io.Reader, m model.ServiceModel, entitySet string, grenze int64) (model.EntitySeite, error) {
	d, wurzel, err := vorbereiten(ctx, src, m, entitySet, grenze)
	if err != nil {
		return model.EntitySeite{}, err
	}
	s, err := d.sammlung(&wurzel)
	if err != nil {
		return model.EntitySeite{}, err
	}
	return s, nil
}

// vorbereiten prüft die Argumente (SPEC-126 Nr. 1), liest den Strom und liefert den Baum samt dem Deuter.
func vorbereiten(ctx context.Context, src io.Reader, m model.ServiceModel, entitySet string, grenze int64) (*deuter, knoten, error) {
	var leer knoten
	switch {
	case ctx == nil:
		return nil, leer, errors.Join(model.ErrUngueltig, errors.New("ctx fehlt"))
	case src == nil:
		return nil, leer, errors.Join(model.ErrUngueltig, errors.New("src fehlt"))
	case grenze < 1:
		return nil, leer, errors.Join(model.ErrUngueltig, errors.New("grenze kleiner 1"))
	case entitySet == "":
		return nil, leer, errors.Join(model.ErrUngueltig, errors.New("entitySet fehlt"))
	}
	if err := ctx.Err(); err != nil {
		return nil, leer, err
	}
	d, ok := neuerDeuter(m, entitySet)
	if !ok {
		return nil, leer, &model.DatenFehler{Ursache: model.DatenEntitySet, Ort: entitySet}
	}
	wurzel, err := lies(ctx, src, grenze)
	if err != nil {
		return nil, leer, err
	}
	if err := ctx.Err(); err != nil {
		return nil, leer, err
	}
	return d, wurzel, nil
}

// lies liest src bis io.EOF, höchstens grenze+1 Bytes, und liefert den Baum (SPEC-126 Nr. 2 bis 5). Der Kontext
// wird vor jedem Read geprüft; die Bytes eines Reads werden verarbeitet, bevor der Fehler dieses Reads gilt.
func lies(ctx context.Context, src io.Reader, grenze int64) (knoten, error) {
	s := &scanner{}
	puffer := make([]byte, 32<<10)
	var gelesen int64
	for {
		if err := ctx.Err(); err != nil {
			return knoten{}, err
		}
		rest := grenze - gelesen
		want := len(puffer)
		if rest < int64(want) {
			want = int(rest) + 1
		}
		n, err := src.Read(puffer[:want])
		chunk := puffer[:n]
		ueber := int64(n) > rest
		if ueber {
			chunk = chunk[:rest]
		}
		gelesen += int64(len(chunk))
		if u := s.fuettere(chunk); u != "" {
			return knoten{}, &model.DatenFehler{Ursache: u}
		}
		if ueber {
			return knoten{}, &model.DatenFehler{Ursache: model.DatenZuGross}
		}
		switch {
		case err == io.EOF:
			if u := s.ende(); u != "" {
				return knoten{}, &model.DatenFehler{Ursache: u}
			}
			return s.wurzel, nil
		case err != nil:
			return knoten{}, err
		}
	}
}

// Package batchcodec liest den Body eines `$batch`-Requests im Format multipart/mixed (SPEC-182, SPEC-183). Er liest Bytes und
// entscheidet nichts über Zustand, Methode, Ziel oder Body der Einzelrequests.
package batchcodec

import (
	"context"
	"errors"
	"io"
	"math"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.ODataBatch = (*Leser)(nil)

// lesepuffer ist die größte Länge, die ein Read anfordert.
const lesepuffer = 32 * 1024

// Leser setzt ODataBatch um. Er hat kein Feld: gleichzeitige Aufrufe beeinflussen sich nicht.
type Leser struct{}

// NewLeser liefert den Leser der Version 4.
func NewLeser() *Leser { return &Leser{} }

// LeseBatch liest src in der Folge von SPEC-182 Nr. 5: Argumente, Kontext, Content-Type der Hülle, Strom, Kontext, Zerlegen.
func (*Leser) LeseBatch(ctx context.Context, src io.Reader, contentType []string, grenze int64) (model.Batch, error) {
	if ctx == nil || src == nil || grenze < 1 {
		return model.Batch{}, errors.Join(model.ErrUngueltig, errors.New("batchcodec: ctx, src oder grenze ungültig"))
	}
	if err := ctx.Err(); err != nil {
		return model.Batch{}, err
	}
	boundary, err := huellenGrenze(contentType)
	if err != nil {
		return model.Batch{}, err
	}
	daten, err := lies(ctx, src, grenze)
	if err != nil {
		return model.Batch{}, err
	}
	if err := ctx.Err(); err != nil {
		return model.Batch{}, err
	}
	return zerlege(daten, boundary)
}

// lies liest src allein mit Read bis io.EOF, höchstens grenze+1 Bytes, und prüft den Kontext vor jedem Read (SPEC-182 Nr. 2). Ein
// Fehler von src bleibt unverändert; das Byte grenze+1 geht ihm vor.
func lies(ctx context.Context, src io.Reader, grenze int64) ([]byte, error) {
	puffer := make([]byte, lesepuffer)
	var daten []byte
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, err := src.Read(puffer[:anforderung(grenze, int64(len(daten)))])
		daten = append(daten, puffer[:n]...)
		switch {
		case int64(len(daten)) > grenze:
			return nil, fehler(model.BatchZuGross, "")
		case err == io.EOF:
			return daten, nil
		case err != nil:
			return nil, err
		}
	}
}

// anforderung ist die Länge des nächsten Reads: höchstens der Puffer und höchstens grenze+1 abzüglich der gelieferten Bytes
// (geliefert ist nie größer als grenze).
func anforderung(grenze, geliefert int64) int64 {
	if grenze == math.MaxInt64 {
		return lesepuffer
	}
	return min(lesepuffer, grenze+1-geliefert)
}

// fehler ist der BatchFehler mit Ursache und Ort.
func fehler(u model.BatchUrsache, ort string) error { return &model.BatchFehler{Ursache: u, Ort: ort} }

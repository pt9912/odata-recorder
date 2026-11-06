package jsoncodec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Schreiber setzt ODataSchreiber für die Version 4.0 um (SPEC-168, SPEC-169). Er hat keinen Zustand und keinen Cache:
// gleichzeitige Aufrufe beeinflussen sich nicht.
type Schreiber struct{}

// NewSchreiber liefert den Schreiber.
func NewSchreiber() *Schreiber { return &Schreiber{} }

// version ist die Version, die der Schreiber setzt; das Kennzeichen jeder Ausgabe nennt sie (SPEC-169 Nr. 7).
const version = "4.0"

func kennzeichen() model.Kopfzeile { return model.Kopfzeile{Name: "OData-Version", Wert: version} }

// ungueltig ist ein Programmfehler des Aufrufers (SPEC-168 Nr. 1): model.ErrUngueltig ohne Code und ohne DatenFehler.
func ungueltig(format string, args ...any) error {
	return errors.Join(model.ErrUngueltig, fmt.Errorf(format, args...))
}

// SchreibeEntity schreibt die Entity als OData-JSON (SPEC-169 Nr. 5).
func (*Schreiber) SchreibeEntity(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.EntityAntwort) (model.Darstellung, error) {
	if err := pruefeZiele(ctx, dst); err != nil {
		return model.Darstellung{}, err
	}
	body, d, err := baueEntity(m, a)
	if err != nil {
		return model.Darstellung{}, err
	}
	return ausgeben(ctx, dst, body, d)
}

// SchreibeCollection schreibt die Collection als OData-JSON (SPEC-169 Nr. 6).
func (*Schreiber) SchreibeCollection(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.CollectionAntwort) (model.Darstellung, error) {
	if err := pruefeZiele(ctx, dst); err != nil {
		return model.Darstellung{}, err
	}
	body, d, err := baueCollection(m, a)
	if err != nil {
		return model.Darstellung{}, err
	}
	return ausgeben(ctx, dst, body, d)
}

// SchreibeAnzahl schreibt die Ziffern von n (SPEC-169 Nr. 7).
func (*Schreiber) SchreibeAnzahl(ctx context.Context, dst io.Writer, n int64) (model.Darstellung, error) {
	if err := pruefeZiele(ctx, dst); err != nil {
		return model.Darstellung{}, err
	}
	if n < 0 {
		return model.Darstellung{}, ungueltig("Anzahl negativ")
	}
	return ausgeben(ctx, dst, strconv.AppendInt(nil, n, 10), model.Darstellung{ContentType: "text/plain;charset=utf-8", Kennzeichen: kennzeichen()})
}

// SchreibeFehler schreibt das Fehlerobjekt zu einem Code der Tabelle (SPEC-168 Nr. 5 f, SPEC-169 Nr. 7).
func (*Schreiber) SchreibeFehler(ctx context.Context, dst io.Writer, f model.Fehlerobjekt) (model.Darstellung, error) {
	if err := pruefeZiele(ctx, dst); err != nil {
		return model.Darstellung{}, err
	}
	c, ok := model.ParseCode(f.Error.Code)
	if !ok {
		return model.Darstellung{}, ungueltig("Code des Fehlerobjekts unbekannt")
	}
	soll, err := model.NeuesFehlerobjekt(c)
	if err != nil || soll != f {
		return model.Darstellung{}, ungueltig("Fehlerobjekt weicht von der Tabelle ab")
	}
	var b ausgabe
	b.roh(`{"error":{"code":`)
	b.text(f.Error.Code)
	b.roh(`,"message":`)
	b.text(f.Error.Message)
	b.roh(`,"target":`)
	b.text(f.Error.Target)
	b.roh(`}}`)
	if b.err != nil {
		return model.Darstellung{}, b.err
	}
	return ausgeben(ctx, dst, b.b, model.Darstellung{ContentType: "application/json;odata.metadata=minimal;charset=utf-8", Kennzeichen: kennzeichen()})
}

// Adresse liefert `<Wurzel>/<Set>(<Schlüssel>)` (SPEC-169 Nr. 2).
func (*Schreiber) Adresse(m model.ServiceModel, wurzel, entitySet, schluessel string) (string, error) {
	if _, err := setTyp(m, entitySet); err != nil {
		return "", err
	}
	return adresse(wurzel, entitySet, schluessel)
}

// pruefeZiele prüft ctx und dst auf nil (SPEC-168 Nr. 1 a).
func pruefeZiele(ctx context.Context, dst io.Writer) error {
	if ctx == nil {
		return ungueltig("ctx fehlt")
	}
	if dst == nil {
		return ungueltig("dst fehlt")
	}
	return nil
}

// ausgeben fragt den Kontext genau einmal und schreibt body in genau einem Write (SPEC-168 Nr. 1, 6).
func ausgeben(ctx context.Context, dst io.Writer, body []byte, d model.Darstellung) (model.Darstellung, error) {
	if err := ctx.Err(); err != nil {
		return model.Darstellung{}, err
	}
	n, err := dst.Write(body)
	if err != nil {
		return model.Darstellung{}, err
	}
	if n < len(body) {
		return model.Darstellung{}, io.ErrShortWrite
	}
	return d, nil
}

package jsoncodecv2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Schreiber setzt ODataSchreiber für die Version 2.0 um (SPEC-168, SPEC-172 Nr. 6). Er hat keinen Zustand und keinen Cache:
// gleichzeitige Aufrufe beeinflussen sich nicht.
type Schreiber struct{}

// NewSchreiber liefert den Schreiber.
func NewSchreiber() *Schreiber { return &Schreiber{} }

// versionSchreiber ist die Version, die der Schreiber setzt; das Kennzeichen jeder Ausgabe nennt sie (SPEC-172 Nr. 6 g).
const versionSchreiber = "2.0"

func kennzeichen() model.Kopfzeile {
	return model.Kopfzeile{Name: "DataServiceVersion", Wert: versionSchreiber}
}

// ungueltig ist ein Programmfehler des Aufrufers (SPEC-168 Nr. 1): model.ErrUngueltig ohne Code und ohne DatenFehler.
func ungueltig(format string, args ...any) error {
	return errors.Join(model.ErrUngueltig, fmt.Errorf(format, args...))
}

// SchreibeEntity schreibt die Entity in der Hülle `d` (SPEC-172 Nr. 6 a).
func (*Schreiber) SchreibeEntity(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.EntityAntwort) (model.Darstellung, error) {
	if err := pruefeZiele(ctx, dst); err != nil {
		return model.Darstellung{}, err
	}
	body, err := baueEntity(m, a)
	if err != nil {
		return model.Darstellung{}, err
	}
	return ausgeben(ctx, dst, body, jsonDarstellung())
}

// SchreibeCollection schreibt die Collection in der Hülle `d` (SPEC-172 Nr. 6 b, c).
func (*Schreiber) SchreibeCollection(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.CollectionAntwort) (model.Darstellung, error) {
	if err := pruefeZiele(ctx, dst); err != nil {
		return model.Darstellung{}, err
	}
	body, err := baueCollection(m, a)
	if err != nil {
		return model.Darstellung{}, err
	}
	return ausgeben(ctx, dst, body, jsonDarstellung())
}

// SchreibeAnzahl schreibt die Ziffern von n (SPEC-172 Nr. 6 d).
func (*Schreiber) SchreibeAnzahl(ctx context.Context, dst io.Writer, n int64) (model.Darstellung, error) {
	if err := pruefeZiele(ctx, dst); err != nil {
		return model.Darstellung{}, err
	}
	if n < 0 {
		return model.Darstellung{}, ungueltig("Anzahl negativ")
	}
	return ausgeben(ctx, dst, strconv.AppendInt(nil, n, 10), model.Darstellung{ContentType: "text/plain;charset=utf-8", Kennzeichen: kennzeichen()})
}

// SchreibeFehler schreibt das Fehlerobjekt zu einem Code der Tabelle ohne `target` (SPEC-168 Nr. 5 f, SPEC-172 Nr. 6 e).
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
	// code und message sind die Texte der Tabelle und gültiges UTF-8: der Aufbau kann nicht scheitern (SPEC-168 Nr. 5 f).
	var b ausgabe
	b.roh(`{"error":{"code":`)
	b.text(f.Error.Code)
	b.roh(`,"message":{"lang":"de","value":`)
	b.text(f.Error.Message)
	b.roh(`}}}`)
	return ausgeben(ctx, dst, b.b, jsonDarstellung())
}

// Adresse liefert `<Wurzel>/<Set>(<Schlüssel>)` (SPEC-172 Nr. 6 h).
func (*Schreiber) Adresse(m model.ServiceModel, wurzel, entitySet, schluessel string) (string, error) {
	if _, err := setTyp(m, entitySet); err != nil {
		return "", err
	}
	return adresse(wurzel, entitySet, schluessel)
}

// jsonDarstellung ist der Kopf jeder Ausgabe außer der Anzahl (SPEC-172 Nr. 6 g).
func jsonDarstellung() model.Darstellung {
	return model.Darstellung{ContentType: "application/json;charset=utf-8", Kennzeichen: kennzeichen()}
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

package jsoncodec

import (
	"context"
	"errors"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// EingabeLeser setzt ODataEingabe um (SPEC-180, SPEC-181). Er hat keinen Zustand und kein Feld: gleichzeitige Aufrufe
// beeinflussen sich nicht.
type EingabeLeser struct{}

// NewEingabe liefert den Leser der Client-Bodies der Version 4.
func NewEingabe() *EingabeLeser { return &EingabeLeser{} }

// LeseEntity liest den Body eines POST, PATCH oder PUT in der Folge von SPEC-181 Nr. 8: Argumente, Kontext, Content-Type, Lesen,
// Kontext, Deuten.
func (*EingabeLeser) LeseEntity(ctx context.Context, src io.Reader, contentType []string, m model.ServiceModel, entitySet string, art model.EingabeArt, grenze int64) (model.EntityEingabe, error) {
	d, wurzel, err := eingabeVorbereiten(ctx, src, contentType, m, entitySet, art, grenze)
	if err != nil {
		return eingabeFehlerErgebnis(err)
	}
	e, err := d.eingabe(&wurzel, art)
	if err != nil {
		return eingabeFehlerErgebnis(err)
	}
	return e, nil
}

// eingabeFehlerErgebnis ist das Ergebnis jedes Fehlerfalls: der Nullwert, nie ein Teilergebnis (SPEC-180 Nr. 2).
func eingabeFehlerErgebnis(err error) (model.EntityEingabe, error) {
	return model.EntityEingabe{}, err
}

// programmfehler ist model.ErrUngueltig ohne Code und ohne EingabeFehler (SPEC-180 Nr. 1 a).
func programmfehler(was string) error { return errors.Join(model.ErrUngueltig, errors.New(was)) }

// eingabeVorbereiten prüft die Argumente, den Kontext und den Content-Type, liest den Strom und liefert den Baum samt dem
// Deuter.
func eingabeVorbereiten(ctx context.Context, src io.Reader, contentType []string, m model.ServiceModel, entitySet string, art model.EingabeArt, grenze int64) (*deuter, knoten, error) {
	var leer knoten
	switch {
	case ctx == nil:
		return nil, leer, programmfehler("ctx fehlt")
	case src == nil:
		return nil, leer, programmfehler("src fehlt")
	case grenze < 1:
		return nil, leer, programmfehler("grenze kleiner 1")
	case entitySet == "":
		return nil, leer, programmfehler("entitySet fehlt")
	case !art.Gueltig():
		return nil, leer, programmfehler("art unbekannt")
	}
	d, ok := neuerDeuter(m, entitySet)
	if !ok {
		return nil, leer, programmfehler("entitySet unbekannt")
	}
	if err := ctx.Err(); err != nil {
		return nil, leer, err
	}
	if !contentTypeGueltig(contentType) {
		return nil, leer, &model.EingabeFehler{Ursache: model.EingabeContentType}
	}
	wurzel, err := lies(ctx, &stromLeser{src: src}, grenze)
	if err != nil {
		return nil, leer, eingabeLesefehler(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, leer, err
	}
	return d, wurzel, nil
}

// srcFehler trägt einen Fehler von src durch lies, damit er vom Fehler des Lesens zu unterscheiden bleibt.
type srcFehler struct{ err error }

func (e *srcFehler) Error() string { return e.err.Error() }

// stromLeser reicht allein Read an src weiter und kennzeichnet jeden Fehler außer io.EOF.
type stromLeser struct{ src io.Reader }

func (s *stromLeser) Read(p []byte) (int, error) {
	n, err := s.src.Read(p)
	if err != nil && err != io.EOF {
		return n, &srcFehler{err: err}
	}
	return n, err
}

// eingabeLesefehler bildet den Fehler des Lesens ab: ein Verstoß des Scanners ist ein EingabeFehler gleichen Namens, ein Fehler
// von src und der Fehler des Kontexts bleiben unverändert.
func eingabeLesefehler(err error) error {
	if f, ok := err.(*srcFehler); ok {
		return f.err
	}
	if f, ok := err.(*model.DatenFehler); ok {
		return &model.EingabeFehler{Ursache: model.EingabeUrsache(f.Ursache), Ort: f.Ort}
	}
	return err
}

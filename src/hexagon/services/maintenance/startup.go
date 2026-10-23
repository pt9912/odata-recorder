package maintenance

import (
	"context"
	"errors"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// integritaet ist Stufe 1 (SPEC-123 Nr. 4): jedes Element der Liste von Check ist ein Befund, und
// ein Befund verhindert den Start.
func (w *Wartung) integritaet(ctx context.Context, b *inbound.Wartungsbericht) error {
	befunde, err := w.k.Speicher.Check(ctx)
	if err != nil {
		return mitCode(err)
	}
	if len(befunde) == 0 {
		return nil
	}
	b.Befunde = slices.Clone(befunde)
	w.protokolliereIntegritaet(ctx, len(befunde))
	return &inbound.Fehler{Code: model.CodeE3001, Status: 500, Ursache: outbound.ErrBeschaedigt}
}

// offeneSessions ist Stufe 2 (SPEC-123 Nr. 5): jede offene Session wird incomplete.
func (w *Wartung) offeneSessions(ctx context.Context, b *inbound.Wartungsbericht) error {
	offen, err := w.k.Sessions.FindOpen(ctx)
	if err != nil {
		return mitCode(err)
	}
	for _, s := range offen {
		if err := w.k.Sessions.SetStatus(ctx, s.ID, model.SessionIncomplete); err != nil {
			return mitCode(err)
		}
		b.OffeneBeendet++
	}
	return nil
}

// aufbewahrung ist Stufe 3 (SPEC-123 Nr. 6): eine Session, deren Beginn echt vor jetzt minus
// MaxAge liegt, wird gelöscht; der Vergleich geht über die Grenze, nicht über die Differenz.
func (w *Wartung) aufbewahrung(ctx context.Context, b *inbound.Wartungsbericht) error {
	if w.k.MaxAge == 0 {
		return nil
	}
	grenze := w.k.Uhr.Now().Add(-w.k.MaxAge)
	dienste, err := w.k.Dienste.List(ctx)
	if err != nil {
		return mitCode(err)
	}
	for _, d := range dienste {
		sessions, err := w.k.Sessions.List(ctx, d.ID)
		if err != nil {
			return mitCode(err)
		}
		for _, s := range sessions {
			if !s.CreatedAt.Before(grenze) {
				continue
			}
			if err := w.loesche(ctx, b, d.ID, s.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// loesche löscht eine alte Session. Eine Quelle einer Baseline wird übersprungen und gezählt, eine
// schon entfernte Session übergangen, ein Kontextfehler und jeder andere Fehler beenden die Stufe.
func (w *Wartung) loesche(ctx context.Context, b *inbound.Wartungsbericht, dienstID, sessionID string) error {
	err := w.k.Sessions.Delete(ctx, sessionID)
	switch {
	case err == nil:
		b.SessionsGeloescht++
	case kontextFehler(err):
		return err
	case errors.Is(err, outbound.ErrReferenziert):
		b.SessionsUebersprungen++
		w.protokolliereUebersprungen(ctx, dienstID, sessionID)
	case errors.Is(err, outbound.ErrNichtGefunden):
	default:
		return mitCode(err)
	}
	return nil
}

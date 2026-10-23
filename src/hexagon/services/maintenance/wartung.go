// Package maintenance ist der Wartungsdienst: er setzt MaintenanceUseCase um (SPEC-123). Er
// importiert Domain und Ports, keinen anderen Use Case und keinen Adapter.
package maintenance

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var _ inbound.MaintenanceUseCase = (*Wartung)(nil)

// Konfig ist die Konfiguration des Dienstes (SPEC-123 Nr. 1). MaxAge 0 schaltet die Aufbewahrung aus.
type Konfig struct {
	Sessions outbound.SessionRepository
	Dienste  outbound.MetadataRepository
	Speicher outbound.StorageMaintenance
	Blobs    *BlobWaechter
	Uhr      support.Clock
	MaxAge   time.Duration
	Log      *slog.Logger
}

// Wartung setzt MaintenanceUseCase um. Sie hält keinen Zustand außer ihrer Konfiguration.
type Wartung struct {
	k Konfig
}

// New prüft die Konfiguration (SPEC-123 Nr. 1).
func New(k Konfig) (*Wartung, error) {
	if k.Sessions == nil || k.Dienste == nil || k.Speicher == nil || k.Blobs == nil || k.Uhr == nil || k.Log == nil || k.MaxAge < 0 {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Port, Wächter, Uhr oder Log fehlt, oder MaxAge ist negativ"))
	}
	return &Wartung{k: k}, nil
}

func neuerBericht() inbound.Wartungsbericht {
	return inbound.Wartungsbericht{Befunde: []string{}, BlobsFehlend: []string{}, BlobsBeschaedigt: []string{}}
}

func ctxNil() error {
	return errors.Join(model.ErrUngueltig, errors.New("Kontext fehlt"))
}

// Startup führt die vier Stufen von SPEC-123 Nr. 3 in ihrer Reihenfolge aus; der erste Fehler
// beendet den Aufruf, und der Bericht trägt, was bis dahin geschah.
func (w *Wartung) Startup(ctx context.Context) (inbound.Wartungsbericht, error) {
	b := neuerBericht()
	if ctx == nil {
		return b, ctxNil()
	}
	stufen := []func(context.Context, *inbound.Wartungsbericht) error{
		w.integritaet, w.offeneSessions, w.aufbewahrung, w.blobAbgleich,
	}
	for _, stufe := range stufen {
		if err := stufe(ctx, &b); err != nil {
			return b, err
		}
	}
	return b, nil
}

// CleanupBlobs stellt sich bei einer offenen Session zurück und gleicht sonst die Blobs ab, ohne
// zu schonen, was der Wächter innerhalb der Karenz sah (SPEC-123 Nr. 8). Es ändert keine Session
// und schreibt kein Protokoll.
func (w *Wartung) CleanupBlobs(ctx context.Context) (inbound.Loeschbericht, error) {
	if ctx == nil {
		return inbound.Loeschbericht{}, ctxNil()
	}
	w.k.Blobs.sperre.Lock()
	defer w.k.Blobs.sperre.Unlock()
	offen, err := w.k.Sessions.FindOpen(ctx)
	if err != nil {
		return inbound.Loeschbericht{}, mitCode(err)
	}
	if len(offen) > 0 {
		return inbound.Loeschbericht{Zurueckgestellt: true}, nil
	}
	jetzt := w.k.Uhr.Now()
	n, err := w.abgleich(ctx, func(ref string) bool { return w.k.Blobs.geschuetzt(ref, jetzt) })
	return inbound.Loeschbericht{Entfernt: n}, err
}

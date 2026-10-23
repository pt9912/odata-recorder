// Package session ist der Sessiondienst: er verwaltet die Recording-Sessions und setzt
// SessionUseCase um (SPEC-124). Er importiert Domain und Ports, keinen anderen Use Case und
// keinen Adapter.
package session

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

// MaxIDLaenge ist die größte Länge einer vom Aufrufer gelieferten Kennung in Bytes (SPEC-124 Nr. 4).
const MaxIDLaenge = 64

// Konfig ist die Konfiguration des Dienstes (SPEC-124 Nr. 1). Dienste dient allein List, Wartung
// allein der Bereinigung nach dem Löschen. Record, Exchanges und Frist dienen allein dem Beenden:
// Record ist der RecordUseCase, den auch der HTTP-Adapter ruft, Frist der Wert von
// SHUTDOWN_TIMEOUT und die frist jedes Aufrufs von Drain.
type Konfig struct {
	Sessions  outbound.SessionRepository
	Dienste   outbound.MetadataRepository
	Wartung   inbound.MaintenanceUseCase
	Uhr       support.Clock
	IDs       support.IDGenerator
	Log       *slog.Logger
	Record    inbound.RecordUseCase
	Exchanges outbound.RecordingRepository
	Frist     time.Duration
}

// Verwaltung setzt die Operationen von SessionUseCase um. Sie hält keinen Zustand außer ihrer
// Konfiguration (SPEC-124 Nr. 10).
type Verwaltung struct {
	k Konfig
}

// New prüft die Konfiguration (SPEC-124 Nr. 1).
func New(k Konfig) (*Verwaltung, error) {
	if k.Sessions == nil || k.Dienste == nil || k.Wartung == nil || k.Uhr == nil || k.IDs == nil || k.Log == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Port, Uhr, Generator oder Log fehlt"))
	}
	if k.Record == nil || k.Exchanges == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Record oder Exchanges fehlt"))
	}
	if k.Frist < 0 {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Frist ist negativ"))
	}
	return &Verwaltung{k: k}, nil
}

func ctxNil() error {
	return errors.Join(model.ErrUngueltig, errors.New("Kontext fehlt"))
}

// idGueltig sagt, ob id der Form aus SPEC-124 Nr. 4 genügt: 1 bis MaxIDLaenge Bytes aus 0-9, a-z, -.
func idGueltig(id string) bool {
	if id == "" || len(id) > MaxIDLaenge {
		return false
	}
	for i := 0; i < len(id); i++ {
		b := id[i]
		if (b < '0' || b > '9') && (b < 'a' || b > 'z') && b != '-' {
			return false
		}
	}
	return true
}

// Create prüft die Argumente, bezieht Kennung (nur bei leerer id) und Zeit und legt die Session
// mit genau einem Aufruf von Sessions.Create an (SPEC-124 Nr. 3).
func (v *Verwaltung) Create(ctx context.Context, serviceID, id string) (model.Session, error) {
	if ctx == nil {
		return model.Session{}, ctxNil()
	}
	if serviceID == "" {
		return model.Session{}, errors.Join(model.ErrUngueltig, errors.New("serviceID fehlt"))
	}
	if id != "" && !idGueltig(id) {
		return model.Session{}, errors.Join(model.ErrUngueltig, errors.New("id außerhalb der Form"))
	}
	if id == "" {
		id = v.k.IDs.SessionID()
	}
	s := model.Session{ID: id, ServiceID: serviceID, CreatedAt: v.k.Uhr.Now(), Status: model.SessionOpen}
	if err := v.k.Sessions.Create(ctx, s); err != nil {
		return model.Session{}, mitCode(err, outbound.GegenstandSession)
	}
	return s, nil
}

// OpenInitialSession ist Create mit leerer id (SPEC-124 Nr. 5); den Modus prüft der Aufrufer.
func (v *Verwaltung) OpenInitialSession(ctx context.Context, serviceID string) (model.Session, error) {
	return v.Create(ctx, serviceID, "")
}

// Get reicht die Kennung unverändert an Sessions.Get (SPEC-124 Nr. 6).
func (v *Verwaltung) Get(ctx context.Context, id string) (model.Session, error) {
	if ctx == nil {
		return model.Session{}, ctxNil()
	}
	s, err := v.k.Sessions.Get(ctx, id)
	if err != nil {
		return model.Session{}, mitCode(err, outbound.GegenstandSession)
	}
	return s, nil
}

// List liefert die Sessions aller Services in der Reihenfolge der Services; der erste Fehler
// beendet den Aufruf ohne Teilliste (SPEC-124 Nr. 6).
func (v *Verwaltung) List(ctx context.Context) ([]model.Session, error) {
	if ctx == nil {
		return nil, ctxNil()
	}
	dienste, err := v.k.Dienste.List(ctx)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSonstiges)
	}
	alle := []model.Session{}
	for _, d := range dienste {
		ss, err := v.k.Sessions.List(ctx, d.ID)
		if err != nil {
			return nil, mitCode(err, outbound.GegenstandSonstiges)
		}
		alle = append(alle, ss...)
	}
	return alle, nil
}

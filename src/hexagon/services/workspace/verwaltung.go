// Package workspace ist der Workspace-Dienst: er legt Workspaces aus Baselines an, liest, löscht und
// setzt sie zurück und setzt WorkspaceUseCase um (SPEC-129). Er importiert Domain und Ports, keinen
// anderen Use Case und keinen Adapter.
package workspace

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// MaxIDLaenge ist die größte Länge einer vom Aufrufer gelieferten Kennung in Bytes (SPEC-129 Nr. 3).
const MaxIDLaenge = 64

// Konfig ist die Konfiguration des Dienstes (SPEC-129 Nr. 1). Workspaces dient Get, Delete, List und
// dem Lesen vor dem Reset, Transaktion der Anlage und dem Reset; Snapshots und Blobs dienen allein dem
// Lesen der Baseline, Dienste allein List, Wartung allein der Bereinigung nach dem Löschen, Uhr und IDs
// allein Create.
type Konfig struct {
	Workspaces  outbound.WorkspaceRepository
	Transaktion outbound.WorkspaceTransactor
	Snapshots   outbound.SnapshotRepository
	Blobs       outbound.BlobRepository
	Dienste     outbound.MetadataRepository
	Wartung     inbound.MaintenanceUseCase
	Uhr         support.Clock
	IDs         support.IDGenerator
	Log         *slog.Logger
}

// Verwaltung setzt die Operationen von WorkspaceUseCase um. Sie hält keinen Zustand außer ihrer
// Konfiguration und keine Sperre (SPEC-129 Nr. 14).
type Verwaltung struct {
	k Konfig
}

// New prüft die Konfiguration und ruft keinen Port (SPEC-129 Nr. 1).
func New(k Konfig) (*Verwaltung, error) {
	if k.Workspaces == nil || k.Transaktion == nil || k.Snapshots == nil || k.Blobs == nil || k.Dienste == nil ||
		k.Wartung == nil || k.Uhr == nil || k.IDs == nil || k.Log == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Port, Wartung, Uhr, Generator oder Log fehlt"))
	}
	return &Verwaltung{k: k}, nil
}

// idGueltig sagt, ob die nichtleere id der Form aus SPEC-129 Nr. 3 genügt: höchstens MaxIDLaenge Bytes aus 0-9, a-z, -.
// Die leere id fragt der Aufrufer nicht; sie heißt „vergeben“.
func idGueltig(id string) bool {
	if len(id) > MaxIDLaenge {
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

// Get reicht die Kennung unverändert an Workspaces.Get (SPEC-129 Nr. 8).
func (v *Verwaltung) Get(ctx context.Context, id string) (model.Workspace, error) {
	if ctx == nil {
		return model.Workspace{}, ctxNil()
	}
	w, err := v.k.Workspaces.Get(ctx, id)
	if err != nil {
		return model.Workspace{}, mitCode(err, outbound.GegenstandWorkspace)
	}
	return w, nil
}

// List liefert die Workspaces aller Services in der Reihenfolge der Services; der erste Fehler
// beendet den Aufruf ohne Teilliste (SPEC-129 Nr. 8).
func (v *Verwaltung) List(ctx context.Context) ([]model.Workspace, error) {
	if ctx == nil {
		return nil, ctxNil()
	}
	dienste, err := v.k.Dienste.List(ctx)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSonstiges)
	}
	alle := []model.Workspace{}
	for _, d := range dienste {
		ws, err := v.k.Workspaces.List(ctx, d.ID)
		if err != nil {
			return nil, mitCode(err, outbound.GegenstandSonstiges)
		}
		alle = append(alle, ws...)
	}
	return alle, nil
}

// transaktion führt f in genau einer Transaktion aus. Der Fehler von f kommt unverändert zurück, der
// Fehler der Transaktion selbst hat den Gegenstand Sonstiges (SPEC-129 Nr. 12); eine Panik reicht sie
// weiter (Nr. 15).
func (v *Verwaltung) transaktion(ctx context.Context, f func(outbound.WorkspaceStores) error) error {
	var fe error
	err := v.k.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		fe = f(s)
		return fe
	})
	switch {
	case err == nil:
		return nil
	case fe != nil && errors.Is(err, fe):
		return fe
	}
	return mitCode(err, outbound.GegenstandSonstiges)
}

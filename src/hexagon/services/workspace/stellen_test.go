package workspace_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// operation ist eine der Operationen von WorkspaceUseCase mit dem Zustand, den sie braucht.
type operation struct {
	name string
	vor  func(w *welt)
	lauf func(w *welt, ctx context.Context) error
}

func operationen() []operation {
	mitWorkspace := func(w *welt) { w.vorbelegen("ws-1", 0) }
	return []operation{
		{"Create", func(*welt) {}, func(w *welt, ctx context.Context) error { _, err := w.v.Create(ctx, "bsl-1", ""); return err }},
		{"Get", mitWorkspace, func(w *welt, ctx context.Context) error { _, err := w.v.Get(ctx, "ws-1"); return err }},
		{"List", mitWorkspace, func(w *welt, ctx context.Context) error { _, err := w.v.List(ctx); return err }},
		{"Delete", mitWorkspace, func(w *welt, ctx context.Context) error { _, err := w.v.Delete(ctx, "ws-1"); return err }},
		{"Reset", func(w *welt) { w.vorbelegen("ws-1", 0); w.veraendern("ws-1") }, func(w *welt, ctx context.Context) error { _, err := w.v.Reset(ctx, "ws-1"); return err }},
		{"Restore", func(w *welt) {
			w.vorbelegen("ws-1", 0)
			w.veraendern("ws-1")
			w.snapshotAblegen("ws-1", "snp-1", snapshotState("ws-1"))
		},
			func(w *welt, ctx context.Context) error { _, err := w.v.Restore(ctx, "ws-1", "snp-1"); return err }},
	}
}

func operationNamens(name string) operation {
	for _, o := range operationen() {
		if o.name == name {
			return o
		}
	}
	panic("Operation " + name)
}

// stelle ist eine Aufrufstelle des Dienstes mit Fehlerrückgabe (SPEC-129 Nr. 15): die Operation, der Name
// der Operation des Fakes und der Gegenstand der Abbildung.
type stelle struct {
	nr, op, label string
	g             outbound.Gegenstand
}

const (
	lblTx = "Transaktion.WithWorkspaceTransaction"
)

// stellen sind die Paare aus Stelle und Operation mit Fehlerrückgabe (S3 bis S5, S7 bis S16, S19, S20); 28 Paare,
// mit S17 (die Bereinigung gibt ihren Fehler nicht zurück, Z36 und Z37) die 29 aus Nr. 15. S20 liegt allein auf
// dem Weg einer Baseline-Kennung; vorbereitet ruft Restore dann mit ihr.
func stellen() []stelle {
	return []stelle{
		{"S10", "Restore", "Workspaces.Get", outbound.GegenstandWorkspace},
		{"S19", "Restore", "Snapshots.GetWorkspace", outbound.GegenstandSnapshot},
		{"S20", "Restore", "Snapshots.Get", outbound.GegenstandSnapshot},
		{"S4", "Restore", "Blobs.Open", outbound.GegenstandSnapshot},
		{"S5", "Restore", "Strom.Read", outbound.GegenstandSnapshot},
		{"S7", "Restore", lblTx, outbound.GegenstandSonstiges},
		{"S11", "Restore", "tx Workspaces.Get", outbound.GegenstandWorkspace},
		{"S9", "Restore", "tx Entities.ReplaceAll", outbound.GegenstandWorkspace},
		{"S13", "Restore", "tx Workspaces.Advance", outbound.GegenstandWorkspace},
		{"S3", "Create", "Snapshots.Get", outbound.GegenstandSnapshot},
		{"S4", "Create", "Blobs.Open", outbound.GegenstandSnapshot},
		{"S5", "Create", "Strom.Read", outbound.GegenstandSnapshot},
		{"S7", "Create", lblTx, outbound.GegenstandSonstiges},
		{"S8", "Create", "tx Workspaces.Create", outbound.GegenstandWorkspace},
		{"S9", "Create", "tx Entities.ReplaceAll", outbound.GegenstandWorkspace},
		{"S10", "Reset", "Workspaces.Get", outbound.GegenstandWorkspace},
		{"S3", "Reset", "Snapshots.Get", outbound.GegenstandSnapshot},
		{"S4", "Reset", "Blobs.Open", outbound.GegenstandSnapshot},
		{"S5", "Reset", "Strom.Read", outbound.GegenstandSnapshot},
		{"S7", "Reset", lblTx, outbound.GegenstandSonstiges},
		{"S11", "Reset", "tx Workspaces.Get", outbound.GegenstandWorkspace},
		{"S9", "Reset", "tx Entities.ReplaceAll", outbound.GegenstandWorkspace},
		{"S12", "Reset", "tx Cursors.ResetAll", outbound.GegenstandWorkspace},
		{"S13", "Reset", "tx Workspaces.Advance", outbound.GegenstandWorkspace},
		{"S10", "Get", "Workspaces.Get", outbound.GegenstandWorkspace},
		{"S14", "List", "Dienste.List", outbound.GegenstandSonstiges},
		{"S15", "List", "Workspaces.List", outbound.GegenstandSonstiges},
		{"S16", "Delete", "Workspaces.Delete", outbound.GegenstandWorkspace},
	}
}

// gegenstandVon ist der Gegenstand der Stelle für den Fehler err: bei S8 steht ErrNichtGefunden für die
// Baseline (SPEC-129 Nr. 15).
func (s stelle) gegenstandVon(err error) outbound.Gegenstand {
	if s.nr == "S8" && errors.Is(err, outbound.ErrNichtGefunden) {
		return outbound.GegenstandSnapshot
	}
	return s.g
}

func (s stelle) String() string { return fmt.Sprintf("%s/%s", s.op, s.nr) }

// vorbereitet liefert eine Welt, in der die Operation der Stelle lauffähig ist.
func (s stelle) vorbereitet(t *testing.T) (*welt, operation) {
	w := neu(t)
	o := operationNamens(s.op)
	o.vor(w)
	if s.nr == "S20" {
		// S20 liegt allein auf dem Weg einer Baseline-Kennung (SPEC-129 Nr. 16 d).
		o.lauf = func(w *welt, ctx context.Context) error { _, err := w.v.Restore(ctx, "ws-1", "bsl-1"); return err }
	}
	return w, o
}

func TestStellenMenge(t *testing.T) {
	st := stellen()
	if len(st) != 28 {
		t.Errorf("Stellen %d", len(st))
	}
	paare := map[string]bool{}
	for _, s := range st {
		paare[s.String()] = true
	}
	if len(paare) != len(st) {
		t.Error("doppelte Paare")
	}
	if !slices.Contains(outbound.EintraegeAbbildung(), "ErrNichtGefunden") {
		t.Error("Tabelle ohne ErrNichtGefunden")
	}
}

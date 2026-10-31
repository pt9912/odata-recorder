package workspace_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// umhuellt liefert das gültige Manifest mit z vor (seite -1) oder hinter (seite 1) dem Wert.
func umhuellt(z []byte, seite int) []byte {
	g := gueltig()
	if seite < 0 {
		return append(append([]byte{}, z...), g...)
	}
	return append(g, z...)
}

// TestManifestLeerraum (Z20, SPEC-129 Nr. 6 a): Leerraum vor und hinter dem Wert sind genau Leerzeichen, Tab, LF und
// CR (RFC 8259); Seitenvorschub, VT, NBSP, BOM und NUL sind keiner. Die Werte entstehen aus Bytewerten (AGENTS.md
// §3.18); die Mengen sind gezählt, damit der Test bricht, wenn sie schrumpfen.
func TestManifestLeerraum(t *testing.T) {
	leerraum := [][]byte{{0x20}, {0x09}, {0x0A}, {0x0D}}
	keinLeerraum := [][]byte{{0x0C}, {0x0B}, {0xC2, 0xA0}, {0xEF, 0xBB, 0xBF}, {0x00}}
	if len(leerraum) != 4 || len(keinLeerraum) != 5 {
		t.Fatalf("Mengen %d und %d, erwartet 4 und 5", len(leerraum), len(keinLeerraum))
	}
	alle := []byte{0x20, 0x09, 0x0A, 0x0D}
	angenommen, abgewiesen := 0, 0
	lauf := func(name string, m []byte, soll bool) {
		w := neu(t)
		w.baselineBytes("bsl-1", "svc", m)
		_, err := w.v.Create(context.Background(), "bsl-1", "ws-neu")
		if soll {
			angenommen++
			if err != nil {
				t.Errorf("%s: %v", name, err)
			}
			return
		}
		abgewiesen++
		pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
		if w.p.anzahl(lblTx) != 0 {
			t.Errorf("%s: Transaktion gelaufen", name)
		}
	}
	for _, seite := range []int{-1, 1} {
		for _, z := range leerraum {
			lauf(fmt.Sprintf("Leerraum %x Seite %d", z, seite), umhuellt(z, seite), true)
		}
		lauf(fmt.Sprintf("alle vier Seite %d", seite), umhuellt(alle, seite), true)
		for _, z := range keinLeerraum {
			lauf(fmt.Sprintf("kein Leerraum %x Seite %d", z, seite), umhuellt(z, seite), false)
		}
	}
	if angenommen != 10 || abgewiesen != 10 {
		t.Errorf("angenommen %d, abgewiesen %d, erwartet je 10", angenommen, abgewiesen)
	}
}

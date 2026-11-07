package model_test

import (
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestRessourceMengen (SPEC-196 Nr. 9): die beiden Mengen sind genau die Konstanten in deren Folge; ein Wert mehr, ein Wert
// weniger oder eine andere Folge macht den Test rot, und mit ihm die Zählprüfungen des Dienstes.
func TestRessourceMengen(t *testing.T) {
	arten := []model.RessourcenArt{model.RessourceMetadaten, model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl}
	if got := model.AlleRessourcenArten(); !slices.Equal(got, arten) || len(got) != 4 {
		t.Errorf("AlleRessourcenArten = %v, erwartet %v", got, arten)
	}
	status := []model.CoverageStatus{"complete", "partial", "unknown"}
	if got := model.AlleCoverageStatus(); !slices.Equal(got, status) || len(got) != 3 {
		t.Errorf("AlleCoverageStatus = %v, erwartet %v", got, status)
	}
	// Die Art 0 und die auf die letzte folgende sind keine Art.
	if slices.Contains(model.AlleRessourcenArten(), 0) || slices.Contains(model.AlleRessourcenArten(), model.RessourceAnzahl+1) {
		t.Error("AlleRessourcenArten nennt einen Wert außerhalb der Konstanten")
	}
}

// TestRessourceMengenKopie: das Ergebnis gehört dem Aufrufer; eine Änderung wirkt nicht auf den nächsten Aufruf.
func TestRessourceMengenKopie(t *testing.T) {
	a := model.AlleRessourcenArten()
	a[0] = 99
	if model.AlleRessourcenArten()[0] != model.RessourceMetadaten {
		t.Error("AlleRessourcenArten teilt seinen Speicher")
	}
	s := model.AlleCoverageStatus()
	s[0] = "x"
	if model.AlleCoverageStatus()[0] != model.CoverageComplete {
		t.Error("AlleCoverageStatus teilt seinen Speicher")
	}
}

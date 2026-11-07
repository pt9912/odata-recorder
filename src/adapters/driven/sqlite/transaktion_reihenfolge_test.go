package sqlite_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

// Reihenfolge der Prüfungen einer schreibenden Operation in der unbrauchbaren Transaktion
// (SPEC-108 Nr. 1 f): Argumente, dann Zustand unbrauchbar, dann Kontext, dann Datenbank. Die
// Tests rufen die Operationen mit Kontexten ohne Haltepunkt; sie zählen keine Prüfungen.

// inUnbrauchbarer führt f in einer Transaktion aus, die durch Entities.Put unbrauchbar wurde
// (ROLLBACK TO scheitert mit keinSavepoint); f erhält den ersten Fehler der Transaktion.
func inUnbrauchbarer(t *testing.T, f func(erster error, s outbound.WorkspaceStores)) {
	t.Helper()
	v := outboundtest.OperationenBasis(t, neueVerbindung)
	_ = inTransaktion(neuerHaltepunkt(2, false, hebeSavepointAuf), v, func(ctx context.Context, s outbound.WorkspaceStores) error {
		erster := entityPut(t).Lauf(ctx, s)
		if erster == nil || !strings.Contains(erster.Error(), keinSavepoint) {
			t.Fatalf("Aufbau: Fehler %v, erwartet der Fehler von ROLLBACK TO (%s)", erster, keinSavepoint)
		}
		f(erster, s)
		return nil
	})
}

// benannte Negative: diese zwei schreibenden Operationen nehmen nur eine Kennung entgegen und
// erkennen vor dem Savepoint keinen ungültigen Wert (outboundtest.UngueltigeAufrufe).
func ohneUngueltigesArgument() map[string]bool {
	return map[string]bool{"Workspaces.Delete": true, "Cursors.ResetAll": true}
}

// unbrauchbarArgument: Schleife über die 8 schreibenden Operationen. Bei den 6 mit erkennbar
// ungültigem Argument liefert die Operation in der unbrauchbaren Transaktion ErrUngueltig und
// nicht deren Fehler, und der Zustand bleibt gesetzt; mit lebender Transaktion liefert derselbe
// Aufruf ErrUngueltig ebenso (der Aufruf ist wirklich ungültig).
func unbrauchbarArgument(t *testing.T) {
	ungueltig := outboundtest.UngueltigeAufrufe()
	ohne := ohneUngueltigesArgument()
	ops := operationen(t, func(op outboundtest.Operation) bool { return op.Schreibend })
	if len(ops) != 8 || len(ungueltig) != 6 || len(ohne) != 2 {
		t.Fatalf("%d schreibende Operationen, %d ungültige Aufrufe, %d benannte Negative, erwartet 8, 6 und 2", len(ops), len(ungueltig), len(ohne))
	}
	for _, op := range ops {
		lauf, hat := ungueltig[op.Name]
		if hat == ohne[op.Name] {
			t.Fatalf("%s: genau eines von ungültigem Aufruf und benanntem Negativ erwartet", op.Name)
		}
		if !hat {
			continue
		}
		t.Run(op.Name, func(t *testing.T) {
			bg := context.Background()
			v := outboundtest.OperationenBasis(t, neueVerbindung)
			_ = v.Transaktion.WithWorkspaceTransaction(bg, func(s outbound.WorkspaceStores) error {
				if err := lauf(bg, s); !errors.Is(err, model.ErrUngueltig) {
					t.Fatalf("lebende Transaktion: Fehler %v, erwartet ErrUngueltig", err)
				}
				return nil
			})
			inUnbrauchbarer(t, func(erster error, s outbound.WorkspaceStores) {
				if err := lauf(bg, s); !errors.Is(err, model.ErrUngueltig) || errors.Is(err, erster) {
					t.Errorf("Fehler %v, erwartet ErrUngueltig und nicht der Fehler der Transaktion %v", err, erster)
				}
				if err := op.Lauf(bg, s); err != erster {
					t.Errorf("danach Fehler %v, erwartet der Zustand unverändert (%v)", err, erster)
				}
			})
		})
	}
}

// unbrauchbarKontext: Schleife über die 8 schreibenden Operationen und die 3 Arten des
// Kontextfehlers (24 Elemente). Eine weitere schreibende Operation mit gültigem Argument und
// beendetem Kontext liefert den Fehler der Transaktion und nicht den des Kontexts.
func unbrauchbarKontext(t *testing.T) {
	ops := operationen(t, func(op outboundtest.Operation) bool { return op.Schreibend })
	arten := outboundtest.KontextArten()
	if len(ops) != 8 || len(arten) != 3 {
		t.Fatalf("%d schreibende Operationen, %d Arten, erwartet 8 und 3", len(ops), len(arten))
	}
	for _, op := range ops {
		t.Run(op.Name, func(t *testing.T) {
			inUnbrauchbarer(t, func(erster error, s outbound.WorkspaceStores) {
				for _, a := range arten {
					err := op.Lauf(a.Beendet(), s)
					if err != erster || errors.Is(err, a.Soll) {
						t.Errorf("%s: Fehler %v, erwartet der Fehler der Transaktion %v", a.Name, err, erster)
					}
				}
			})
		})
	}
}

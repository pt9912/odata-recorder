package sqlite_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

// Die Steuerstatements ROLLBACK TO und RELEASE einer schreibenden Operation scheitern hier
// einzeln und gemeinsam (SPEC-108 Nr. 1 f): ein Wrapper der Verbindung der Transaktion lässt
// genau das benannte Statement mit einem eigenen Fehler scheitern, die Datenbank sieht es nicht.

const (
	rollbackTo = "ROLLBACK TO operation"
	releaseOp  = "RELEASE operation"
)

// steuerFall ist eine Kombination gescheiterter Steuerstatements. erfolg sagt, ob die Operation
// gelingt (nur RELEASE läuft) oder mit ihrem Fehler der Datenbank endet (ROLLBACK TO und RELEASE
// laufen); soll ist der Fehler der Operation und der Transaktion, nil wenn keines scheitert;
// nicht ist ein Fehler, den die Operation nicht liefert (der erste Fehler bestimmt).
type steuerFall struct {
	name      string
	scheitern map[string]error
	erfolg    bool
	soll      error
	nicht     error
}

func steuerFaelle() []steuerFall {
	eR, eL := errors.New("Fake: ROLLBACK TO scheitert"), errors.New("Fake: RELEASE scheitert")
	return []steuerFall{
		{"keines", nil, false, nil, nil},
		{"Fehlerpfad nur ROLLBACK TO", map[string]error{rollbackTo: eR}, false, eR, nil},
		{"Fehlerpfad nur RELEASE", map[string]error{releaseOp: eL}, false, eL, nil},
		{"Fehlerpfad beide", map[string]error{rollbackTo: eR, releaseOp: eL}, false, eR, eL},
		{"Erfolgspfad nur RELEASE", map[string]error{releaseOp: eL}, true, eL, nil},
	}
}

// unbrauchbarSteuerung: Schleife über die 8 schreibenden Operationen und die 5 Fälle (40
// Elemente). Scheitert ein Steuerstatement, liefern die Operation, jede weitere schreibende
// Operation und die Transaktion denselben Fehler, und nichts wird bestätigt.
func unbrauchbarSteuerung(t *testing.T) {
	ops := operationen(t, func(op outboundtest.Operation) bool { return op.Schreibend })
	if len(ops) != 8 || len(steuerFaelle()) != 5 {
		t.Fatalf("%d schreibende Operationen, %d Fälle, erwartet 8 und 5", len(ops), len(steuerFaelle()))
	}
	for _, op := range ops {
		if op.Fehler == nil {
			t.Fatalf("%s: kein Aufruf mit Fehler der Datenbank", op.Name)
		}
		for _, f := range steuerFaelle() {
			t.Run(op.Name+"/"+f.name, func(t *testing.T) { steuerLauf(t, op, f) })
		}
	}
}

func steuerLauf(t *testing.T, op outboundtest.Operation, f steuerFall) {
	bg := context.Background()
	v := outboundtest.OperationenBasis(t, neueVerbindung)
	vorher := outboundtest.LiesGesamtzustand(t, v)
	var opErr, weitere error
	err := v.Transaktion.WithWorkspaceTransaction(bg, func(s outbound.WorkspaceStores) error {
		sqlite.ScheitereBei(s, f.scheitern)
		lauf := op.LaufFehler
		if f.erfolg {
			lauf = op.Lauf
		}
		opErr = lauf(bg, s)
		weitere = entityPut(t).Lauf(bg, s)
		return nil
	})
	if f.soll == nil {
		if !errors.Is(opErr, op.Fehler) || err != nil || weitere != nil {
			t.Fatalf("Operation %v (erwartet %v), Transaktion %v, weitere %v, erwartet gelungen", opErr, op.Fehler, err, weitere)
		}
		return
	}
	for was, got := range map[string]error{"Operation": opErr, "weitere Operation": weitere, "Transaktion": err} {
		if !errors.Is(got, f.soll) || (f.nicht != nil && errors.Is(got, f.nicht)) {
			t.Errorf("%s: Fehler %v, erwartet %v", was, got, f.soll)
		}
	}
	if got := outboundtest.LiesGesamtzustand(t, v); !reflect.DeepEqual(got, vorher) {
		t.Errorf("die unbrauchbare Transaktion bestätigte:\n%+v\n%+v", got, vorher)
	}
}

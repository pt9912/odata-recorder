package baselineimport_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// TestAuditGelungeneWiederholung (V-1, SPEC-234 Nr. 4 d und 10): ein zweiter Aufruf mit gleichem Hash liefert
// den gespeicherten Import (Neu falsch) und schreibt keinen Eintrag; der erste Aufruf auch nicht.
func TestAuditGelungeneWiederholung(t *testing.T) {
	w := neu(t)
	w.sessionR()
	erst := mustImport(t, w, anfrage())
	zweit := mustImport(t, w, anfrage())
	if !erst.Neu || zweit.Neu {
		t.Fatalf("Neu: %v, %v", erst.Neu, zweit.Neu)
	}
	if n := w.audit.Aufrufe(); n != 0 {
		t.Errorf("%d Aufrufe des Audits", n)
	}
}

// kontextFormen sind die sechs Formen eines Kontextfehlers: Art (Canceled, DeadlineExceeded) × Verpackung
// (unverpackt, mit %w, in errors.Join neben einem Fehler eines Ports).
func kontextFormen() []error {
	var out []error
	for _, c := range []error{context.Canceled, context.DeadlineExceeded} {
		out = append(out, c, fmt.Errorf("Schritt: %w", c), errors.Join(c, errPlatte))
	}
	return out
}

// TestAuditKontextfehlerSechsFormen (V-2, M4): an jeder der 15 Stellen des Wegs „neu“ und der Wiederholung
// schreibt keine der sechs Formen einen Eintrag: 90 Läufe.
func TestAuditKontextfehlerSechsFormen(t *testing.T) {
	formen := kontextFormen()
	if len(formen) != 6 {
		t.Fatalf("%d Formen", len(formen))
	}
	n := 0
	for _, s := range stellen() {
		if s.weg == "get" {
			continue
		}
		for i, k := range formen {
			w, _, err := s.lauf(t, k)
			if err == nil {
				t.Errorf("%s, Form %d: kein Fehler", s.name, i)
			}
			if a := w.audit.Aufrufe(); a != 0 {
				t.Errorf("%s, Form %d: %d Aufrufe des Audits", s.name, i, a)
			}
			n++
		}
	}
	if n != 90 {
		t.Errorf("Läufe %d, erwartet 90", n)
	}
}

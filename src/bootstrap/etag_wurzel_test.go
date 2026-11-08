package bootstrap_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// TestKonfigEtagEnabled (E3, SPEC-110 Nr. 15, SPEC-282): ungesetzt, leer und false ergeben aus, true an; 1, True, TRUE, yes und 0 sind der Verstoß
// ETAG_ENABLED ohne den Wert in der Meldung, in jedem der drei Modi; die Konfiguration trägt den Wert, der Nullwert ist aus.
func TestKonfigEtagEnabled(t *testing.T) {
	const form = "ODR-E2001 ETAG_ENABLED: Wahrheitswert muss true oder false sein"
	faelle := []struct {
		env   []string
		wert  bool
		zeile []string
	}{
		{nil, false, nil}, {[]string{"ETAG_ENABLED="}, false, nil}, {[]string{"ETAG_ENABLED=  "}, false, nil}, {[]string{"ETAG_ENABLED=false"}, false, nil},
		{[]string{"ETAG_ENABLED=true"}, true, nil},
		{[]string{"ETAG_ENABLED=1"}, false, []string{form}}, {[]string{"ETAG_ENABLED=True"}, false, []string{form}},
		{[]string{"ETAG_ENABLED=TRUE"}, false, []string{form}}, {[]string{"ETAG_ENABLED=yes"}, false, []string{form}},
		{[]string{"ETAG_ENABLED=0"}, false, []string{form}},
	}
	n := 0
	for _, modus := range dreiModi() {
		for _, f := range faelle {
			n++
			k, zeilen := lade(t, append(append(basis(t), "MODE="+modus), f.env...)...)
			if !slices.Equal(zeilen, f.zeile) || (zeilen == nil && k.EtagEnabled != f.wert) {
				t.Errorf("%s %v: EtagEnabled %v, Zeilen %v; erwartet %v und %v", modus, f.env, k.EtagEnabled, zeilen, f.wert, f.zeile)
			}
			for _, z := range zeilen {
				for _, w := range f.env {
					if _, wert, _ := strings.Cut(w, "="); wert != "" && strings.TrimSpace(wert) != "" && strings.Contains(strings.TrimPrefix(z, "ODR-E2001 ETAG_ENABLED"), wert) {
						t.Errorf("%s %v: die Meldung nennt den Wert: %q", modus, f.env, z)
					}
				}
			}
		}
	}
	if n != 30 {
		t.Errorf("%d Fälle, erwartet 30", n)
	}
	// ETAG_ENABLED ungültig zugleich mit MODE ungültig: zwei Zeilen.
	_, zeilen := lade(t, append(basis(t), "MODE=Mock", "ETAG_ENABLED=1")...)
	if s := schluessel(t, zeilen); !slices.Equal(s, []string{"ETAG_ENABLED", "MODE"}) {
		t.Errorf("zwei Verstöße: %v", zeilen)
	}
	if (bootstrap.Konfiguration{}).EtagEnabled || (mock.Dienst{}).Etag {
		t.Error("der Nullwert ist nicht aus")
	}
}

// TestEtagFuerSatz (E4, SPEC-280 Nr. 2): Version des Satzes (4.0, 2.0, leer, andere) × (aus, an): der Dienst liefert ETags genau bei an und 4.0.
func TestEtagFuerSatz(t *testing.T) {
	n := 0
	for _, v := range []string{"4.0", "2.0", "", "4.01", "3.0", "4"} {
		for _, an := range []bool{false, true} {
			n++
			satz := outbound.CodecSatz{Version: v}
			if got, want := bootstrap.EtagFuerSatz(satz, an), an && v == "4.0"; got != want {
				t.Errorf("Version %q an=%v: %v, erwartet %v", v, an, got, want)
			}
		}
	}
	if s := bootstrap.SatzVersion4(); s.Version != "4.0" || !bootstrap.EtagFuerSatz(s, true) || bootstrap.EtagFuerSatz(s, false) {
		t.Errorf("der Satz des Mock hat die Version %q", s.Version)
	}
	if n != 12 {
		t.Errorf("%d Fälle, erwartet 12", n)
	}
}

// TestEtagVerdrahtung (E4): der Modus mock reicht ETAG_ENABLED an jeden Dienst des Satzes 4.0 weiter, an wie aus.
func TestEtagVerdrahtung(t *testing.T) {
	ctx := context.Background()
	for _, an := range []bool{false, true} {
		m := neuM3Haken()
		k := mockKonfig(t.TempDir())
		k.EtagEnabled = an
		hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
		if err != nil {
			t.Fatal(err)
		}
		if len(m.mock) != 1 || len(m.mock[0].Dienste) != 2 {
			t.Fatalf("an=%v: %d Mock, Dienste %+v", an, len(m.mock), m.mock)
		}
		for i, d := range m.mock[0].Dienste {
			if d.Etag != an {
				t.Errorf("an=%v Dienst %d: Etag %v", an, i, d.Etag)
			}
		}
		_ = hd.Herunterfahren(ctx)
	}
}

package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestUmgebung: die Umgebung des Binary hat je Betriebsart genau die fünf Schlüssel MODE, LISTEN, DATA_DIR, SERVICE_ID und SERVICE_ROOT_PATH
// mit den Werten des Vertrags (Orakel: Literale), MODE je Betriebsart verschieden, kein REPLAY_STRATEGY und keine UPSTREAM_URL (SPEC-216 Nr. 9 a).
func TestUmgebung(t *testing.T) {
	wache(t, 30*time.Second)
	erwartet := map[Betriebsart][]string{
		betriebReplay: {"MODE=replay", "LISTEN=127.0.0.1:34567", "DATA_DIR=/daten", "SERVICE_ID=lasttest", "SERVICE_ROOT_PATH=/odata"},
		betriebMock:   {"MODE=mock", "LISTEN=127.0.0.1:34567", "DATA_DIR=/daten", "SERVICE_ID=lasttest", "SERVICE_ROOT_PATH=/odata"},
	}
	n := 0
	for b, soll := range erwartet {
		ist := umgebung(b, "127.0.0.1:34567")
		if fmt.Sprint(ist) != fmt.Sprint(soll) {
			t.Errorf("Betriebsart %s: %v, erwartet %v", b, ist, soll)
		}
		for _, e := range ist {
			n++
			schluessel, _, _ := strings.Cut(e, "=")
			if schluessel == "REPLAY_STRATEGY" || schluessel == "UPSTREAM_URL" {
				t.Errorf("Betriebsart %s: Schlüssel %s", b, schluessel)
			}
		}
	}
	if len(erwartet) != 2 || n != 10 {
		t.Errorf("%d Betriebsarten, %d Schlüssel, erwartet 2 mal 5", len(erwartet), n)
	}
}

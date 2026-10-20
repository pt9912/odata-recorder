package supporttest

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

const (
	goroutinen = 64
	aufrufe    = 500
)

// UhrVertrag prüft SPEC-114 Nr. 1 und 2 für jede Clock: UTC, ohne monotone Ablesung, nicht
// fallend, auch bei gleichzeitigem Aufruf.
func UhrVertrag(t *testing.T, uhr support.Clock) {
	t.Helper()
	pruefeForm(t, uhr.Now())
	letzte := uhr.Now()
	for range 100 {
		jetzt := uhr.Now()
		pruefeForm(t, jetzt)
		if jetzt.Before(letzte) {
			t.Fatalf("Now fällt von %v auf %v", letzte, jetzt)
		}
		letzte = jetzt
	}
	var wg sync.WaitGroup
	fehler := make(chan string, goroutinen)
	for range goroutinen {
		wg.Go(func() {
			vorher := uhr.Now()
			for range 100 {
				jetzt := uhr.Now()
				if jetzt.Before(vorher) {
					fehler <- fmt.Sprintf("Now fällt gleichzeitig von %v auf %v", vorher, jetzt)
					return
				}
				vorher = jetzt
			}
		})
	}
	wg.Wait()
	close(fehler)
	for f := range fehler {
		t.Error(f)
	}
}

func pruefeForm(t *testing.T, z time.Time) {
	t.Helper()
	if z.Location() != time.UTC {
		t.Fatalf("Now liefert Ort %v, erwartet UTC", z.Location())
	}
	if z != z.Round(0) {
		t.Fatalf("Now liefert eine monotone Ablesung: %v", z)
	}
}

// praefixe nennt je Operation von IDGenerator das Präfix (SPEC-114 Nr. 6).
func praefixe() map[string]string {
	return map[string]string{
		"SessionID": "ses", "ExchangeID": "exc", "KorrelationsID": "cor",
		"WorkspaceID": "wsp", "SnapshotID": "snp",
	}
}

// IDVertrag prüft SPEC-114 Nr. 6 und 8 für jeden IDGenerator: Form und Präfix jeder Operation
// des Ports, paarweise verschiedene Kennungen auch bei gleichzeitigem Aufruf.
func IDVertrag(t *testing.T, gen support.IDGenerator) {
	t.Helper()
	typ := reflect.TypeFor[support.IDGenerator]()
	var namen []string
	for i := range typ.NumMethod() {
		namen = append(namen, typ.Method(i).Name)
	}
	sort.Strings(namen)
	var tabelle []string
	for n := range praefixe() {
		tabelle = append(tabelle, n)
	}
	sort.Strings(tabelle)
	if !reflect.DeepEqual(namen, tabelle) {
		t.Fatalf("Methoden des Ports %v und Zeilen der Präfix-Tabelle %v decken sich nicht", namen, tabelle)
	}
	form := regexp.MustCompile(`^[a-z]{3}-[0-9a-f]{32}$`)
	for _, name := range namen {
		erzeuge := reflect.ValueOf(gen).MethodByName(name).Interface().(func() string)
		id := erzeuge()
		if !form.MatchString(id) || id[:3] != praefixe()[name] {
			t.Errorf("%s liefert %q, erwartet %s-<32 × 0-9a-f>", name, id, praefixe()[name])
		}
		eindeutig(t, name, erzeuge)
	}
}

// eindeutig ruft erzeuge gleichzeitig und verlangt paarweise verschiedene Kennungen.
func eindeutig(t *testing.T, name string, erzeuge func() string) {
	t.Helper()
	ergebnis := make([][]string, goroutinen)
	var wg sync.WaitGroup
	for g := range goroutinen {
		wg.Go(func() {
			for range aufrufe {
				ergebnis[g] = append(ergebnis[g], erzeuge())
			}
		})
	}
	wg.Wait()
	gesehen := make(map[string]bool, goroutinen*aufrufe)
	for _, ids := range ergebnis {
		for _, id := range ids {
			if gesehen[id] {
				t.Fatalf("%s liefert %q doppelt", name, id)
			}
			gesehen[id] = true
		}
	}
}

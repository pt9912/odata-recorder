package httpodata_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestIfMatchHeaderUnveraendert (F27, SPEC-136 Nr. 4): If-Match und If-None-Match erreichen den Dienst mit jeder Zeile, in der Reihenfolge der
// Zeilen und unter dem kanonischen Namen, unverändert: Methode (PATCH, PUT, DELETE) × Bindung × Form des Headers (7).
func TestIfMatchHeaderUnveraendert(t *testing.T) {
	formen := []struct {
		name string
		kopf []string
		soll map[string][]string
	}{
		{"schwach", []string{`If-Match: W/"3"`}, map[string][]string{"If-Match": {`W/"3"`}}},
		{"stark", []string{`If-Match: "3"`}, map[string][]string{"If-Match": {`"3"`}}},
		{"Stern", []string{"If-None-Match: *"}, map[string][]string{"If-None-Match": {"*"}}},
		{"zwei Zeilen", []string{`If-Match: "3"`, `If-Match: "4"`}, map[string][]string{"If-Match": {`"3"`, `"4"`}}},
		{"Schreibweisen", []string{`if-match: "3"`, `IF-MATCH: "4"`}, map[string][]string{"If-Match": {`"3"`, `"4"`}}},
		{"Liste mit Komma im Tag", []string{`If-None-Match: "a,b", W/"3"`}, map[string][]string{"If-None-Match": {`"a,b", W/"3"`}}},
		{"leere Zeile", []string{"If-Match:"}, map[string][]string{"If-Match": {""}}},
	}
	n := 0
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for _, m := range []string{"PATCH", "PUT", "DELETE"} {
			for _, f := range formen {
				u.roh(t, zeile(m, "/p", f.kopf...))
				alle := u.d.alle()
				a := alle[len(alle)-1].a
				n++
				gleich(t, fmt.Sprintf("%s %s %s Header", b, m, f.name), a.Header, mitConnection(f.soll))
				gleich(t, fmt.Sprintf("%s %s %s Methode", b, m, f.name), a.Methode, m)
			}
		}
	}
	if n != 2*3*len(formen) {
		t.Errorf("%d Fälle, erwartet %d", n, 2*3*len(formen))
	}
}

// TestIfMatchFehler412 (F27, SPEC-136 Nr. 7 a): ODR-E6010 mit Status 412 geht mit 412 und dem Fehlerobjekt des Codes hinaus, ohne ETag; Methode 3 × Bindung 2.
func TestIfMatchFehler412(t *testing.T) {
	for _, b := range bindungen() {
		for _, m := range []string{"PATCH", "PUT", "DELETE"} {
			u := neuUmgebung(t, b, fehlerDienst(&inbound.Fehler{Code: model.CodeE6010, Status: 412}))
			resp, body := hole(t, u, m, "/p")
			if resp.StatusCode != 412 || !bytes.Equal(body, fehlerobjekt(t, model.CodeE6010)) || len(resp.Header.Values("ETag")) != 0 {
				t.Errorf("%s %s: Antwort %d %q %v, erwartet 412 mit dem Fehlerobjekt von ODR-E6010 und ohne ETag", b, m, resp.StatusCode, body, resp.Header)
			}
		}
	}
}

package httpodata_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestPreferHeaderUnveraendert (Z25, SPEC-136 Nr. 4): jede Kopfzeile `Prefer` erreicht den Dienst, in der Reihenfolge der Zeilen und unter dem kanonischen Namen,
// gleich in welcher Schreibweise der Name kam: Methode (POST, PATCH, PUT) × Bindung × Zahl der Zeilen 0 bis 3 × Schreibweise je Zeile (`Prefer`, `prefer`,
// `PREFER`; 1 + 3 + 9 + 27 Folgen).
func TestPreferHeaderUnveraendert(t *testing.T) {
	namen := []string{"Prefer", "prefer", "PREFER"}
	werte := []string{"return=minimal", "return=representation, respond-async", `x="a, return=minimal"`}
	n := 0
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for _, m := range []string{"POST", "PATCH", "PUT"} {
			for zahl := 0; zahl <= 3; zahl++ {
				folgen := 1
				for i := 0; i < zahl; i++ {
					folgen *= len(namen)
				}
				for f := 0; f < folgen; f++ {
					var kopf []string
					var soll []string
					for i, k := 0, f; i < zahl; i, k = i+1, k/len(namen) {
						kopf = append(kopf, namen[k%len(namen)]+": "+werte[i])
						soll = append(soll, werte[i])
					}
					u.roh(t, zeile(m, "/p", kopf...))
					alle := u.d.alle()
					a := alle[len(alle)-1].a
					n++
					want := mitConnection(nil)
					if soll != nil {
						want["Prefer"] = soll
					}
					gleich(t, fmt.Sprintf("%s %s %q Header", b, m, kopf), a.Header, want)
					gleich(t, fmt.Sprintf("%s %s %q Methode", b, m, kopf), a.Methode, m)
				}
			}
		}
	}
	if n != 2*3*(1+3+9+27) {
		t.Errorf("%d Fälle, erwartet %d", n, 2*3*(1+3+9+27))
	}
}

// TestPreferAntwortUnveraendert (Z25, SPEC-136 Nr. 5): Status 200, 201 und 204 und die Werte der Header `Preference-Applied`, `OData-EntityId` und `Location` des Dienstes
// gehen byte-gleich aus, unter dem kanonischen Namen des Clients (`Odata-Entityid`); die Schreibweise des Namens prüft der Test nicht. Ein `204` hat keinen Body.
// Methode (POST, PATCH, PUT) × Bindung × Status 3 × Wert von `Preference-Applied` 2.
func TestPreferAntwortUnveraendert(t *testing.T) {
	n := 0
	for _, b := range bindungen() {
		for _, m := range []string{"POST", "PATCH", "PUT"} {
			for _, status := range []int{200, 201, 204} {
				for _, applied := range []string{"return=minimal", "return=representation"} {
					h := map[string][]string{
						"Preference-Applied": {applied}, "OData-EntityId": {"http://ort.test/Products(7)?x=1"}, "Location": {"http://ort.test/Products(7)?x=1"},
						"Content-Type": {"application/json"},
					}
					body := []byte(`{"Id":7}`)
					if status == 204 {
						body = nil
					}
					u := neuUmgebung(t, b, liefert(inbound.Antwort{Status: status, Header: h, Body: ausBytes(body)}))
					resp, got := hole(t, u, m, "/p")
					n++
					kontext := fmt.Sprintf("%s %s %d %s", b, m, status, applied)
					if resp.StatusCode != status || !slices.Equal(got, body) && (len(got) != 0 || len(body) != 0) {
						t.Errorf("%s: Status %d Body %q, erwartet %d %q", kontext, resp.StatusCode, got, status, body)
					}
					for name, w := range map[string][]string{"Preference-Applied": {applied}, "Odata-Entityid": h["OData-EntityId"], "Location": h["Location"]} {
						if v := resp.Header[name]; !slices.Equal(v, w) {
							t.Errorf("%s: %s %q, erwartet %q", kontext, name, v, w)
						}
					}
				}
			}
		}
	}
	if n != 2*3*3*2 {
		t.Errorf("%d Fälle, erwartet %d", n, 2*3*3*2)
	}
}

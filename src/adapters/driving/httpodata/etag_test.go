package httpodata_test

import (
	"net/http"
	"slices"
	"sort"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestAntwortEtag (E7, E14, SPEC-281 Nr. 1): die Kopfzeile ETag des Dienstes geht unverändert aus (Wert und Zahl), und sie ist der einzige Unterschied
// der Header-Namen zur Antwort ohne sie: Methode (GET, HEAD) × Status (200, 201, 204) × Wert × Schreibweise des Namens × Bindung.
func TestAntwortEtag(t *testing.T) {
	namen := func(h http.Header) []string {
		var r []string
		for n := range h {
			if n != "Date" {
				r = append(r, n)
			}
		}
		sort.Strings(r)
		return r
	}
	n := 0
	for _, b := range bindungen() {
		for _, methode := range []string{"GET", "HEAD"} {
			for _, status := range []int{200, 201, 204} {
				for _, etag := range []string{`W/"0"`, `W/"7"`, `W/"9223372036854775807"`} {
					for _, name := range []string{"ETag", "Etag", "etag"} {
						basis := map[string][]string{"Content-Type": {"application/json"}, "OData-Version": {"4.0"}}
						mit := map[string][]string{"Content-Type": {"application/json"}, "OData-Version": {"4.0"}, name: {etag}}
						u0 := neuUmgebung(t, b, liefert(inbound.Antwort{Status: status, Header: basis, Body: leer()}))
						ohne, _ := hole(t, u0, methode, "/p")
						u1 := neuUmgebung(t, b, liefert(inbound.Antwort{Status: status, Header: mit, Body: leer()}))
						resp, _ := hole(t, u1, methode, "/p")
						n++
						if got := resp.Header.Values("ETag"); !slices.Equal(got, []string{etag}) {
							t.Errorf("%s %s %d %s %s: ETag %v", b, methode, status, name, etag, got)
						}
						if got, want := namen(resp.Header), append(namen(ohne.Header), "Etag"); !slices.Equal(got, sortiert(want)) {
							t.Errorf("%s %s %d %s: Header-Namen %v, erwartet %v", b, methode, status, name, got, sortiert(want))
						}
					}
				}
			}
		}
	}
	if n != len(bindungen())*2*3*3*3 {
		t.Errorf("%d Fälle", n)
	}
}

func sortiert(l []string) []string {
	r := slices.Clone(l)
	sort.Strings(r)
	return r
}

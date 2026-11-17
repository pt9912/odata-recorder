package hybrid_test

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
)

// TestGleichzeitigUpstream hält Z31 für den Weg U (SPEC-284 Nr. 13): 64 gleichzeitige Aufrufe über zwei Dienste, je ein
// POST mit eigenem Body, eigener Query und eigenem Pfad, laufen durch H3 und U. Jeder Aufruf sendet an das Ziel seines
// Dienstes mit seinem OData-Pfad, seiner Query und seinen Body-Bytes, und die Antwort gehört ihm. Der Fänger zählt die
// Wirkung (Ziel je Aufruf): ein Dienst, der die Liste der Dienste im Weg U beschreibt, sendet an das Ziel des anderen.
// Grenze: ein Datenrennen ohne Wirkung meldet nur der Detektor von make test-race, je Prozess.
func TestGleichzeitigUpstream(t *testing.T) {
	dienste := []hybrid.Dienst{
		{Root: "/odata/a", UpstreamURL: "http://a"},
		{Root: "/odata/b", UpstreamURL: "http://b"},
	}
	g := neuesRigMit(t, uPolicy(t, []string{"a", "b"}, []string{"POST /**"}), 64, dienste)
	g.mockE6003(t)
	g.u.fn = func(_ int, _ context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
		h := map[string][]string{"X-Ziel": {a.Ziel}, "X-Pfad": {a.Pfad}, "X-Query": {a.RawQuery}}
		return outbound.UpstreamAntwort{Status: 200, Header: h, Body: neuerStrom(nil)}, nil
	}
	root := func(id int) string { return []string{"a", "b"}[id%2] }
	antworten := make([]inbound.Antwort, aufrufe)
	fehler := make([]error, aufrufe)
	gleichzeitig(aufrufe, func(id int) {
		a := mitBody("POST", neuerStrom(koerper(id, 0)), -1)
		a.Pfad = fmt.Sprintf("/odata/%s/Item%d", root(id), id)
		a.RawQuery = fmt.Sprintf("id=%d", id)
		a.KorrelationsID = fmt.Sprint(id)
		antworten[id], fehler[id] = g.v.Hybrid(context.Background(), a)
	})
	for id := range aufrufe {
		h := antworten[id].Header
		wantPfad := fmt.Sprintf("/Item%d", id)
		if fehler[id] != nil || h["X-Ziel"][0] != "http://"+root(id) || h["X-Pfad"][0] != wantPfad ||
			h["X-Query"][0] != fmt.Sprintf("id=%d", id) {
			t.Errorf("Aufruf %d: Fehler %v, Header %v", id, fehler[id], h)
		}
	}
	if g.u.zahl() != aufrufe {
		t.Fatalf("U %d-mal gerufen, erwartet %d", g.u.zahl(), aufrufe)
	}
	for _, s := range g.u.aufrufe {
		q, err := url.ParseQuery(s.a.RawQuery)
		if err != nil {
			t.Fatal(err)
		}
		var id int
		if _, err := fmt.Sscan(q.Get("id"), &id); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(s.gelesen, koerper(id, 0)) || s.a.Laenge != int64(len(koerper(id, 0))) {
			t.Errorf("Aufruf %d: Body %v, Laenge %d", id, s.gelesen, s.a.Laenge)
		}
	}
}

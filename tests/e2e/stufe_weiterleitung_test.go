package e2e_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/tests/e2e/fixture"
)

// TestE2EWeiterleitung ist die Stufe von LH-FA-01 (SPEC-037 Nr. 7): jede Antwortart des Referenzdienstes
// erreicht den Client unverändert, leere Antworten bleiben leer, ein gestoppter Referenzdienst ist 502.
func TestE2EWeiterleitung(t *testing.T) {
	binaryPfad(t)
	ref := neueReferenz(t)
	p := starte(t, konfig{modus: "record", dir: t.TempDir(), upstream: ref.url()})
	p.bereit()
	const jsonTyp = "application/json;odata.metadata=minimal"
	ctx := ref.url() + "/odata/$metadata#Products"
	arten := []struct {
		name, pfad string
		body       string
		header     map[string]string
	}{
		{"$metadata", "/odata/$metadata", fixture.CSDL, map[string]string{"Content-Type": "application/xml"}},
		{"Seite", "/odata/Products", fmt.Sprintf(`{"@odata.context":%q,"@odata.nextLink":%q,"value":[{"Id":1,"Name":"Anvil"}]}`, ctx, ref.url()+"/odata/Products?$skiptoken=1"),
			map[string]string{"Content-Type": jsonTyp, "ETag": `W/"p1"`}},
		{"Entität", "/odata/Products(1)", fmt.Sprintf(`{"@odata.context":%q,"Id":1,"Name":"Anvil"}`, ref.url()+"/odata/$metadata#Products/$entity"),
			map[string]string{"Content-Type": jsonTyp, "ETag": `W/"e1"`}},
		{"$count", "/odata/Products/$count", "7", map[string]string{"Content-Type": "text/plain"}},
		{"$value", "/odata/Products(1)/Name/$value", "Anvil", map[string]string{"Content-Type": "text/plain"}},
	}
	for _, a := range arten {
		t.Run(a.name, func(t *testing.T) {
			w := holen(t, p.adresse, anfrage{pfad: a.pfad})
			if w.status != 200 || string(w.body) != a.body {
				t.Errorf("%d %q, erwartet 200 %q", w.status, w.body, a.body)
			}
			soll := map[string][]string{"Odata-Version": {"4.0"}, "Set-Cookie": {"sid=" + geheimC}}
			for k, v := range a.header {
				soll[http.CanonicalHeaderKey(k)] = []string{v}
			}
			ist := map[string][]string{}
			for k, v := range w.header {
				if k != "Date" && k != "Content-Length" {
					ist[k] = v
				}
			}
			if fmt.Sprint(ist) != fmt.Sprint(soll) {
				t.Errorf("Header %v, erwartet %v", ist, soll)
			}
		})
	}
	t.Run("leer", func(t *testing.T) {
		for _, f := range []struct {
			pfad   string
			status int
		}{{"/odata/Leer", 200}, {"/odata/Nichts", 204}} {
			if w := holen(t, p.adresse, anfrage{pfad: f.pfad}); w.status != f.status || len(w.body) != 0 {
				t.Errorf("%s: %d %q, erwartet %d ohne Body", f.pfad, w.status, w.body, f.status)
			}
		}
	})
	t.Run("Upstream gestoppt", func(t *testing.T) {
		ref.stoppe()
		w := holen(t, p.adresse, anfrage{pfad: "/odata/Products(1)"})
		pruefeFehler(t, "gestoppter Referenzdienst", w, 502, "ODR-E4001")
		z := p.warteAufZeile("ereignis=upstream", "grund=nicht_erreichbar")
		if !strings.Contains(z, "error_code=ODR-E4001") {
			t.Errorf("Eintrag ohne error_code=ODR-E4001: %s", z)
		}
	})
}

package admin_test

import (
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// importAuthVerstoss ist ein Verstoß gegen SPEC-238 Nr. 4 von M-Auth; gruppe ist a (kein oder mehrfacher Header), b (anderes
// Schema, leerer Wert) oder c (falscher Kandidat).
type importAuthVerstoss struct {
	name, gruppe string
	kopf         map[string][]string
}

// importAuthVerstoesse sind die 13 Verstöße von M-Auth (SPEC-297 Nr. 8, SPEC-238 Nr. 4 a bis c): 3 der Gruppe a, 6 der Gruppe
// b und 4 der Gruppe c; k ist das gültige Token.
func importAuthVerstoesse(k string) []importAuthVerstoss {
	a := "Authorization"
	gueltig := "Bearer " + k
	h := func(werte ...string) map[string][]string { return map[string][]string{a: werte} }
	return []importAuthVerstoss{
		{"kein Header", "a", nil},
		{"zwei Zeilen gleichlautend gültig", "a", h(gueltig, gueltig)},
		{"eine gültige und eine falsche", "a", h(gueltig, "Bearer "+strings.Repeat("x", len(k)))},
		{"leerer Wert", "b", h("")},
		{"Basic", "b", h("Basic " + k)},
		{"Token", "b", h("Token " + k)},
		{"Bearer ohne Leerzeichen", "b", h("Bearer" + k)},
		{"Bearer allein", "b", h("Bearer")},
		{"Bearer mit Tabulator", "b", h("Bearer\t" + k)},
		{"falsches Token gleicher Länge", "c", h("Bearer " + strings.Repeat("x", len(k)))},
		{"Präfix des Tokens", "c", h("Bearer " + k[:len(k)-1])},
		{"Token plus ein Zeichen", "c", h("Bearer " + k + "t")},
		{"Token in anderer Schreibweise", "c", h("Bearer " + strings.ToUpper(k))},
	}
}

// TestIM14ImportAutorisierung hält IM14 (SPEC-297 Nr. 8; LH-FA-18, SPEC-238 Nr. 3 d, 4, 8) über `admin.New` mit den echten
// Routen von ImportRouten (Rahmen und Routen zusammen): ein Request ohne gültiges Token ist 401 `ODR-E2101` mit
// `WWW-Authenticate: Bearer`, `Import` und `Get` 0 Aufrufe, der Body wird nicht gelesen (0 Lesezugriffe), genau ein Eintrag
// `admin_abgewiesen` vom Rahmen und keiner einer Art `import_*`; ein ungültiger Body ändert das nicht. M-Auth: 13 Verstöße ×
// (`POST` mit gültigem Body, `POST` mit ungültigem Body `{`, `GET`) = 39 Fälle; Kontrolle: mit gültigem Token 201 (`POST`)
// und 200 (`GET`), je 1 Aufruf.
func TestIM14ImportAutorisierung(t *testing.T) {
	k := string(tokenGueltig())
	verstoesse := importAuthVerstoesse(k)
	je := map[string]int{}
	for _, v := range verstoesse {
		je[v.gruppe]++
	}
	gleichAuf(t, "Verstöße je Gruppe gleich 3 + 6 + 4", []int{len(verstoesse), je["a"], je["b"], je["c"]}, []int{13, 3, 6, 4})
	anfragen := []struct {
		name, methode, ziel, body string
	}{
		{"POST gültiger Body", "POST", importsZiel, gutBody},
		{"POST ungültiger Body", "POST", importsZiel, "{"},
		{"GET", "GET", importsZiel + "/imp-pfad", gutBody},
	}
	faelle := 0
	for _, v := range verstoesse {
		for _, a := range anfragen {
			faelle++
			name := v.name + ", " + a.name
			u, fake, _ := importsUmgebung(t)
			leser := &bodyLeser{daten: []byte(a.body)}
			r := httptest.NewRequest(a.methode, a.ziel, leser)
			for n, w := range v.kopf {
				r.Header[n] = w
			}
			rec := u.rufe(r)
			pruefe401(t, name, rec)
			if fake.gesamt() != 0 || leser.n.Load() != 0 {
				t.Errorf("%s: %d Aufrufe des Dienstes, %d Lesezugriffe auf den Body, erwartet 0 und 0", name, fake.gesamt(), leser.n.Load())
			}
			got := u.audit.Eintraege()
			id := rec.Header().Get("X-Correlation-Id")
			if len(got) != 1 || got[0].Art != support.ArtAdminAbgewiesen || got[0].KorrelationsID != id || id == "" {
				t.Errorf("%s: Einträge %+v, erwartet genau einen Eintrag admin_abgewiesen mit der Korrelations-ID der Antwort", name, got)
			}
			for _, e := range got {
				if strings.HasPrefix(string(e.Art), "import_") {
					t.Errorf("%s: Eintrag der Art %s", name, e.Art)
				}
			}
			if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2101"}) || andere != 0 {
				t.Errorf("%s: CountError %q, andere Aufrufe %d", name, codes, andere)
			}
		}
	}
	gleichAuf(t, "Fälle", faelle, 39)

	u, fake, _ := importsUmgebung(t)
	if rec := u.rufe(anfrageMit("POST", importsZiel, nil, strings.NewReader(gutBody))); rec.Code != 201 || fake.zahlen() != [impAnzahl]int{1, 0} {
		t.Errorf("Kontrolle POST: Status %d, Zähler %v, erwartet 201 und einen Aufruf von Import", rec.Code, fake.zahlen())
	}
	u, fake, _ = importsUmgebung(t)
	if rec := u.rufe(importGet("imp-pfad")); rec.Code != 200 || fake.zahlen() != [impAnzahl]int{0, 1} {
		t.Errorf("Kontrolle GET: Status %d, Zähler %v, erwartet 200 und einen Aufruf von Get", rec.Code, fake.zahlen())
	}
}

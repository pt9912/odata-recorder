package admin_test

import (
	"io"
	"slices"
	"strings"
	"testing"
	"time"
)

// wsFrist setzt dem Test eine harte Frist: läuft sie ab, beendet ein Zeitgeber den Prozess mit einer Panik, die den Test
// nennt (ein Handler, der nicht zurückkehrt, ließe sich aus dem Test heraus nicht beenden). Die Tests dieser Dateien warten auf
// nichts; die Frist von einer Minute liegt weit über ihrer Laufzeit unter -race und Last (unter 10 s).
func wsFrist(t *testing.T) {
	t.Helper()
	z := time.AfterFunc(time.Minute, func() { panic("Frist von 1 Minute abgelaufen: " + t.Name()) })
	t.Cleanup(func() { z.Stop() })
}

// wsAuthVerstoesse sind die 13 Verstöße von M-Auth gegen SPEC-238 Nr. 4: a (3: kein Header, zwei gleichlautende gültige
// Zeilen, eine gültige und eine falsche), b (6: leerer Wert, `Basic`, `Token`, `Bearer` ohne Leerzeichen, `Bearer` allein,
// `Bearer` mit Tabulator) und c (4: falsches Token gleicher Länge, Präfix, Token plus ein Zeichen, andere Schreibweise); der
// Wert ist die Liste der Werte des Headers `Authorization`, nil heißt: kein Header.
func wsAuthVerstoesse() []struct {
	name  string
	werte []string
} {
	tok := string(tokenGueltig())
	return []struct {
		name  string
		werte []string
	}{
		{"kein Header", nil}, {"zwei gültige Zeilen", []string{"Bearer " + tok, "Bearer " + tok}},
		{"gültige und falsche Zeile", []string{"Bearer " + tok, "Bearer " + strings.Repeat("x", 32)}},
		{"leerer Wert", []string{""}}, {"Basic", []string{"Basic " + tok}}, {"Token", []string{"Token " + tok}},
		{"Bearer ohne Leerzeichen", []string{"Bearer" + tok}}, {"Bearer allein", []string{"Bearer"}},
		{"Bearer mit Tabulator", []string{"Bearer\t" + tok}}, {"falsches Token gleicher Länge", []string{"Bearer " + strings.Repeat("x", 32)}},
		{"Präfix des Tokens", []string{"Bearer " + tok[:31]}}, {"Token plus ein Zeichen", []string{"Bearer " + tok + "t"}},
		{"andere Schreibweise", []string{"Bearer " + strings.ToUpper(tok)}},
	}
}

// TestWS14Authentisierung hält WS14 (SPEC-298 Nr. 1; SPEC-238 Nr. 3 d, 4, 8) über den Server mit den echten Routen: M-Auth, 13
// Verstöße × 11 Fälle (die 8 Routen, bei den drei `POST` mit Body zusätzlich ein ungültiger Body) = 143: 401 `ODR-E2101` mit
// `WWW-Authenticate: Bearer`, keine der acht Operationen gerufen, der Body nie gelesen, genau ein Eintrag `admin_abgewiesen`,
// kein Eintrag einer Art `workspace_*` oder `snapshot_*`; mit gültigem Token ist der Status der Route der Erfolg, je ein Aufruf.
func TestWS14Authentisierung(t *testing.T) {
	wsFrist(t)
	type fall struct {
		name, methode, pfad, body string
		r                         wsRoute
	}
	var faelle []fall
	for i, r := range wsRouten() {
		faelle = append(faelle, fall{r.name, r.methode, r.pfad(wsPfadID), r.body, r})
		if i == rCreate || i == rSnCreate || i == rRestore {
			faelle = append(faelle, fall{r.name + " mit ungültigem Body", r.methode, r.pfad(wsPfadID), "{", r})
		}
	}
	gleichAuf(t, "Fälle", len(faelle), 11)
	gleichAuf(t, "Verstöße gleich 3 + 6 + 4", len(wsAuthVerstoesse()), 3+6+4)
	n := 0
	for _, v := range wsAuthVerstoesse() {
		for _, f := range faelle {
			n++
			name := v.name + " bei " + f.name
			e := wsNeu(t)
			var kopf map[string][]string
			if v.werte != nil {
				kopf = map[string][]string{"Authorization": v.werte}
			}
			l := wsNeuerLeser([]byte(f.body), nil)
			req := ohneAuth(f.methode, f.pfad, kopf)
			req.Body = io.NopCloser(l)
			rec := e.rufe(req)
			pruefeFehlerAntwort(t, name, rec, 401, objekt401, "")
			if l.n.Load() != 0 || e.p.gesamt() != 0 {
				t.Errorf("%s: %d Lesezugriffe auf den Body, %d Aufrufe der Dienste, erwartet 0 und 0", name, l.n.Load(), e.p.gesamt())
			}
			if codes, _ := zaehlungenVon(e.tele); !slices.Equal(codes, []string{"ODR-E2101"}) {
				t.Errorf("%s: CountError %q", name, codes)
			}
			if aufrufe := e.audit.Aufrufe(); aufrufe != 1 {
				t.Errorf("%s: %d Aufrufe von Audit, erwartet 1", name, aufrufe)
			}
			if es := e.audit.Eintraege(); len(es) != 1 || string(es[0].Art) != "admin_abgewiesen" || es[0].WorkspaceID != "" {
				t.Errorf("%s: Einträge %+v, erwartet genau einen `admin_abgewiesen`", name, es)
			}
		}
	}
	gleichAuf(t, "Fälle", n, 143)
	for _, r := range wsRouten() {
		e := wsNeu(t)
		rec := e.rufe(r.gut())
		wsPruefeErfolg(t, "gültiges Token "+r.name, e, r, rec.Code)
	}
}

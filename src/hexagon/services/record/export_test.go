package record

import "github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"

// Hilfen für die externen Tests: die reinen Funktionen der Wahl des Dienstes, der
// weitergeleiteten Header, von Accept-Encoding und des Zustands ohne Antwort.

// DienstFuer ist dienstFuer (SPEC-121 Nr. 2b).
func DienstFuer(dienste []Dienst, pfad string) (Dienst, string, bool) {
	return dienstFuer(dienste, pfad)
}

// Weiterzuleiten ist weiterzuleiten (SPEC-121 Nr. 4).
func Weiterzuleiten(h map[string][]string) map[string][]string { return weiterzuleiten(h) }

// AcceptEncoding ist acceptEncoding (SPEC-121 Nr. 4).
func AcceptEncoding(h map[string][]string) (string, bool) { return acceptEncoding(h) }

// OhneAntwort ist ohneAntwort (SPEC-121 Nr. 11).
func OhneAntwort(req capture.Ergebnis) capture.Zustand { return ohneAntwort(req) }

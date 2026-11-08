package admin

import "net/http"

// LadeMitHaltepunkten ruft lade mit Haltepunkten: nachAnlegen läuft nach dem Anlegen der Hilfsdatei, vorLink
// vor dem Link (SPEC-239 Nr. 3); beide erhalten den Pfad der Hilfsdatei, nil heißt: kein Halt.
func LadeMitHaltepunkten(pfad string, nachAnlegen, vorLink func(hilfsdatei string)) ([]byte, error) {
	if nachAnlegen == nil {
		nachAnlegen = keinHalt
	}
	if vorLink == nil {
		vorLink = keinHalt
	}
	return lade(pfad, haltepunkte{nachAnlegen: nachAnlegen, vorLink: vorLink})
}

// BildeAnfrage führt anfrage heraus: die Anfrage, die die Route zu r, id und werten erhält (SPEC-238 Nr. 2).
func BildeAnfrage(r *http.Request, id string, werte map[string]string) Anfrage {
	return anfrage(r, id, werte)
}

// Vergleiche führt vergleiche heraus (SPEC-238 Nr. 4 e).
func Vergleiche(erwartet, gegeben []byte) bool { return vergleiche(erwartet, gegeben) }

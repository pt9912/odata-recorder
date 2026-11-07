package admin

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

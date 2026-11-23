package admin

// LadeMitStatHalt ruft lade mit allen drei Haltepunkten: nachStat läuft nach dem fehlenden Stat vor dem Lstat,
// nachAnlegen nach dem Anlegen der Hilfsdatei und vorLink vor dem Link (SPEC-239 Nr. 3, 4); die beiden letzten
// erhalten den Pfad der Hilfsdatei, nil heißt bei jedem: kein Halt.
func LadeMitStatHalt(pfad string, nachStat func(), nachAnlegen, vorLink func(hilfsdatei string)) ([]byte, error) {
	if nachAnlegen == nil {
		nachAnlegen = keinHalt
	}
	if vorLink == nil {
		vorLink = keinHalt
	}
	return lade(pfad, haltepunkte{nachAnlegen: nachAnlegen, vorLink: vorLink, nachStat: nachStat})
}

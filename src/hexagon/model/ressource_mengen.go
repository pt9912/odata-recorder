package model

// AlleRessourcenArten liefert die vier Arten der Ressource in der Folge der Konstanten (SPEC-196 Nr. 9).
func AlleRessourcenArten() []RessourcenArt {
	return []RessourcenArt{RessourceMetadaten, RessourceMenge, RessourceEntity, RessourceAnzahl}
}

// AlleCoverageStatus liefert die drei Zustände der Coverage in der Folge der Konstanten (SPEC-196 Nr. 9).
func AlleCoverageStatus() []CoverageStatus {
	return []CoverageStatus{CoverageComplete, CoveragePartial, CoverageUnknown}
}

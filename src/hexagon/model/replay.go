package model

// Replayfaehig setzt die Fähigkeitsmatrix aus SPEC-045 um (SPEC-117): wahr genau dann, wenn
// der Exchange complete ist, seine Integrität verified oder absent ist und die Redaction der
// Response nicht unsafe ist; applied zählt nur bei redigierteBodies. Ein unbekannter oder
// leerer Wert schließt aus; die Funktion prüft den Exchange nicht und liefert nie einen Fehler.
func Replayfaehig(e Exchange, redigierteBodies bool) bool {
	if e.Capture != CaptureComplete {
		return false
	}
	if e.Integrity != IntegrityVerified && e.Integrity != IntegrityAbsent {
		return false
	}
	switch e.ResponseRedact {
	case RedactionNone:
		return true
	case RedactionApplied:
		return redigierteBodies
	}
	return false
}

package main

// Ergebnisse des Urteils einer Betriebsart (SPEC-216 Nr. 12 d); urteilKein steht für eine Betriebsart, die nicht gefahren wurde.
const (
	urteilBestanden     = "bestanden"
	urteilDurchgefallen = "durchgefallen"
	urteilUngueltig     = "ungueltig"
	urteilKein          = ""
)

// Die Exits des Programms (SPEC-216 Nr. 12 e).
const (
	exitBestanden     = 0
	exitDurchgefallen = 1
	exitAbbruch       = 2
	exitUngueltig     = 3
)

// abnahmewertMs ist der vorläufige Wert von LH-QA-01, gegen den eine Abnahme gemessen wird (SPEC-216 Nr. 12 b).
const abnahmewertMs = 100

// bestanden vergleicht in Mikrosekunden: p95 ≤ Zielwert × 1000 besteht (SPEC-216 Nr. 12 a).
func bestanden(p95, zielMs int64) bool { return p95 <= zielMs*1000 }

// toleranz ist das Größere aus einem Zehntel des kleineren p95 (abgerundet) und 2000 µs (SPEC-216 Nr. 12 c).
func toleranz(a, b int64) int64 { return max(min(a, b)/10, 2000) }

// urteil ist die Zeile `Urteil` einer Betriebsart; Dauern stehen in ganzen Mikrosekunden. P95 ist das größere der gültigen Läufe, Streuung
// und Toleranz sind ohne zwei gültige Läufe 0.
type urteil struct {
	Betriebsart string
	Laeufe      int
	ZielMs      int64
	P95         int64
	Streuung    int64
	Toleranz    int64
	Ergebnis    string
	Grund       string
}

// urteilBetriebsart urteilt über die Läufe einer Betriebsart (SPEC-216 Nr. 12 c, d): ein ungültiger Lauf macht das Urteil ungültig mit
// seinem Grund; zwei gültige Läufe mit einer Differenz der p95 über der Toleranz sind ungültig mit dem Grund streuung; sonst gilt das größere p95.
func urteilBetriebsart(betriebsart string, laeufe []Messung, zielMs int64) urteil {
	u := urteil{Betriebsart: betriebsart, Laeufe: len(laeufe), ZielMs: zielMs, Grund: grundOK}
	for _, l := range laeufe {
		if !l.Gueltig {
			if u.Ergebnis == urteilKein {
				u.Ergebnis, u.Grund = urteilUngueltig, l.Grund
			}
			continue
		}
		u.P95 = max(u.P95, l.P95)
	}
	if u.Ergebnis != urteilKein {
		return u
	}
	if len(laeufe) == 2 {
		a, b := laeufe[0].P95, laeufe[1].P95
		u.Streuung, u.Toleranz = max(a, b)-min(a, b), toleranz(a, b)
		if u.Streuung > u.Toleranz {
			u.Ergebnis, u.Grund = urteilUngueltig, grundStreuung
			return u
		}
	}
	u.Ergebnis = urteilDurchgefallen
	if bestanden(u.P95, zielMs) {
		u.Ergebnis = urteilBestanden
	}
	return u
}

// abnahmeGilt liefert, ob die Abnahme gilt (SPEC-216 Nr. 12 b): das Profil abnahme, der Zielwert abnahmewertMs, die Betriebsart beide und zwei
// Urteile bestanden; sonst nein mit dem ersten Grund in der Folge profil, ziel, betriebsart, ergebnis.
func abnahmeGilt(profil string, zielMs int64, betriebsart string, urteile ...string) (bool, string) {
	switch {
	case profil != "abnahme":
		return false, grundProfil
	case zielMs != abnahmewertMs:
		return false, grundZiel
	case betriebsart != "beide":
		return false, grundBetriebsart
	case len(urteile) != 2 || urteile[0] != urteilBestanden || urteile[1] != urteilBestanden:
		return false, grundErgebnis
	}
	return true, grundOK
}

// ausgang ist, womit das Programm endet: die Urteile von replay und mock (urteilKein für eine nicht gefahrene Betriebsart) und die Gründe für
// Exit 3 und Exit 2.
type ausgang struct {
	Replay, Mock           string
	Aufbau, Frist, Abbruch bool
}

// exitCode ist der Exit des Programms mit dem Vorrang 2 vor 3 vor 1 vor 0 (SPEC-216 Nr. 12 e).
func exitCode(a ausgang) int {
	switch {
	case a.Abbruch:
		return exitAbbruch
	case a.Aufbau || a.Frist || a.Replay == urteilUngueltig || a.Mock == urteilUngueltig:
		return exitUngueltig
	case a.Replay == urteilDurchgefallen || a.Mock == urteilDurchgefallen:
		return exitDurchgefallen
	}
	return exitBestanden
}

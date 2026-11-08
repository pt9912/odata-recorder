package policy

import "github.com/pt9912/odata-recorder/src/hexagon/model"

// Entscheidung ist das Ergebnis von Entscheide; der Nullwert ist Bleiben (SPEC-261 Nr. 4).
type Entscheidung int

// Die drei Entscheidungen: die Antwort des Zweigs bleibt, der Replay-Fallback ist erlaubt, der
// Upstream-Fallback ist erlaubt.
const (
	Bleiben Entscheidung = iota
	Replay
	Upstream
)

// Anfrage trägt genau die Eingaben der Entscheidung: Methode und Pfad des Rohrequests, das Ziel
// (die UPSTREAM_URL des Service) und den ODR-Code des letzten Ergebnisses. Status und Header sind
// keine Eingabe (SPEC-261 Nr. 2 d, Nr. 3).
type Anfrage struct {
	Methode string
	Pfad    string
	Ziel    string
	Code    model.Code
}

// methodenKlasse ist die Klasse einer Methode (SPEC-261 Nr. 2); verglichen wird byte-gleich.
type methodenKlasse int

const (
	klasseNie methodenKlasse = iota
	klasseLesend
	klasseMutierend
)

func klasseVon(methode string) methodenKlasse {
	switch methode {
	case "GET", "HEAD":
		return klasseLesend
	case "POST", "PUT", "PATCH", "DELETE", "MERGE":
		return klasseMutierend
	}
	return klasseNie
}

// ausloeser ist die Art, in der ein Code einen Fallback auslöst (SPEC-261 Nr. 3).
type ausloeser int

const (
	keinAusloeser ausloeser = iota
	mockUnsupported
	mockCoverage
	replayMiss
)

func ausloeserVon(c model.Code) ausloeser {
	switch c {
	case model.CodeE6001, model.CodeE6002:
		return mockUnsupported
	case model.CodeE6003:
		return mockCoverage
	case model.CodeE5001:
		return replayMiss
	}
	return keinAusloeser
}

// Entscheide legt fest, ob die Antwort bleibt oder ein Replay- oder Upstream-Fallback erlaubt ist.
// Die Funktion ist total, hängt nur von p und a ab und ändert nichts (SPEC-261 Nr. 4, 9); die
// erste zutreffende Regel entscheidet, und der Replay geht dem Upstream vor.
func (p *Policy) Entscheide(a Anfrage) Entscheidung {
	if p == nil {
		return Bleiben
	}
	klasse := klasseVon(a.Methode)
	art := ausloeserVon(a.Code)
	if klasse == klasseNie || art == keinAusloeser {
		return Bleiben
	}
	if art == mockUnsupported && p.replayUnsupported || art == mockCoverage && p.replayCoverage {
		return Replay
	}
	if !p.upstream || klasse == klasseMutierend && art == replayMiss || !p.hostFrei(a.Ziel) {
		return Bleiben
	}
	if klasse == klasseLesend || p.routeFrei(a.Methode, a.Pfad) {
		return Upstream
	}
	return Bleiben
}

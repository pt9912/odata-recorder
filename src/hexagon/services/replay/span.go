package replay

import (
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Schlüssel und Platzhalter des Spans `odata.replay.match` (SPEC-296 Nr. 4): die Platzhalter stehen bei StartSpan und bleiben, wo
// der Cursor nicht gelesen wurde bzw. die Auswahl nicht lief.
const (
	attrMatchStrategy     = "recorder.match_strategy"
	attrMatchResult       = "recorder.match_result"
	strategieNichtGelesen = "nicht_gelesen"
	ergebnisNichtGewaehlt = "nicht_gewaehlt"
)

// Ergebnisse der Auswahl (SPEC-296 Nr. 4 b).
const (
	ergebnisTreffer           = "treffer"
	ergebnisKeinTreffer       = "kein_treffer"
	ergebnisErschoepft        = "erschoepft"
	ergebnisAndererKey        = "anderer_key"
	ergebnisNichtReplayfaehig = "nicht_replayfaehig"
)

// ergebnisVon bildet das Ergebnis aus dem Fehler von waehle (SPEC-296 Nr. 4 b): kein Fehler ist ein Treffer, jeder Fehler der Auswahl
// trägt einen der vier Codes ODR-E5001, ODR-E5002, ODR-E5007 und ODR-E5008.
func ergebnisVon(err error) string {
	var f *inbound.Fehler
	if !errors.As(err, &f) {
		return ergebnisTreffer
	}
	switch f.Code {
	case model.CodeE5001:
		return ergebnisKeinTreffer
	case model.CodeE5002:
		return ergebnisErschoepft
	case model.CodeE5007:
		return ergebnisAndererKey
	}
	return ergebnisNichtReplayfaehig
}

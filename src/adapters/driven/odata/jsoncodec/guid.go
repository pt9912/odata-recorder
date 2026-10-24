package jsoncodec

import (
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// guidText ist die kanonische Schreibweise einer Edm.Guid: 8-4-4-4-12 Hex-Ziffern in Kleinbuchstaben (SPEC-126 Nr. 10 c).
func guidText(text string) (string, bool) {
	if len(text) != 36 {
		return "", false
	}
	for i := range len(text) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if text[i] != '-' {
				return "", false
			}
		} else if hexWert(text[i]) < 0 {
			return "", false
		}
	}
	return strings.ToLower(text), true
}

// lesText liest einen Wert, der als String steht und dessen Text kanon in die kanonische Schreibweise bringt (SPEC-126
// Nr. 10 c, d): der Marker geht der Deutung vor, ein anderer JSON-Typ ist `typ`, ein ungültiger Text `wert`.
func lesText(k *knoten, art model.WertArt, kanon func(string) (string, bool)) (model.Wert, model.DatenUrsache) {
	switch {
	case istMarker(k):
		return wertRedigiert(), ""
	case k.art != artString:
		return model.Wert{}, model.DatenTyp
	}
	text, ok := kanon(k.text)
	if !ok {
		return model.Wert{}, model.DatenWert
	}
	return model.Wert{Art: art, Text: text}, ""
}

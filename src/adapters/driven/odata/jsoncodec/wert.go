package jsoncodec

import "github.com/pt9912/odata-recorder/src/hexagon/model"

// lesWert liest den Wert k nach dem EDM-Typ der Property p (SPEC-126 Nr. 10); eine leere Ursache heißt: gelesen.
func lesWert(p model.Property, k *knoten) (model.Wert, model.DatenUrsache) {
	if k.art == artNull {
		if p.Nullable {
			return model.Wert{Art: model.ArtNull}, ""
		}
		return model.Wert{}, model.DatenNull
	}
	switch p.EdmType {
	case "Edm.String":
		return lesString(k)
	case "Edm.Boolean":
		return lesBoolean(k)
	case "Edm.Int32":
		return lesGanz(k, 32, model.ArtInt32, false)
	case "Edm.Int64":
		return lesGanz(k, 64, model.ArtInt64, true)
	case "Edm.Decimal":
		return lesDezimal(k)
	case "Edm.Guid":
		return lesText(k, model.ArtGUID, guidText)
	case "Edm.Date":
		return lesText(k, model.ArtDate, datumText)
	case "Edm.DateTimeOffset":
		return lesText(k, model.ArtDateTimeOffset, zeitpunktText)
	}
	return model.Wert{}, model.DatenTyp
}

// istMarker ist wahr für einen JSON-String, dessen Rohtext genau `[REDACTED]` ist (SPEC-126 Nr. 10 f).
func istMarker(k *knoten) bool { return k.art == artString && k.marker }

func wertRedigiert() model.Wert { return model.Wert{Art: model.ArtRedigiert} }

func lesString(k *knoten) (model.Wert, model.DatenUrsache) {
	switch {
	case istMarker(k):
		return wertRedigiert(), ""
	case k.art == artString:
		return model.Wert{Art: model.ArtString, Text: k.text}, ""
	}
	return model.Wert{}, model.DatenTyp
}

func lesBoolean(k *knoten) (model.Wert, model.DatenUrsache) {
	switch k.art {
	case artWahr:
		return model.Wert{Art: model.ArtBoolean, Text: "true"}, ""
	case artFalsch:
		return model.Wert{Art: model.ArtBoolean, Text: "false"}, ""
	}
	return model.Wert{}, model.DatenTyp
}

// zahlVon liest k als Zahl: eine JSON-Zahl, bei alsString auch einen String, dessen ganzer Inhalt eine ist.
func zahlVon(k *knoten, alsString bool) (zahl, model.DatenUrsache) {
	switch {
	case k.art == artZahl:
		z, _ := zerlege(k.text)
		return z, ""
	case k.art == artString && alsString:
		if z, ok := zerlege(k.text); ok {
			return z, ""
		}
		return zahl{}, model.DatenWert
	}
	return zahl{}, model.DatenTyp
}

func lesGanz(k *knoten, bits int, art model.WertArt, alsString bool) (model.Wert, model.DatenUrsache) {
	if alsString && istMarker(k) {
		return wertRedigiert(), ""
	}
	z, u := zahlVon(k, alsString)
	if u != "" {
		return model.Wert{}, u
	}
	text, ok := z.ganzText(bits)
	if !ok {
		return model.Wert{}, model.DatenWert
	}
	return model.Wert{Art: art, Text: text}, ""
}

func lesDezimal(k *knoten) (model.Wert, model.DatenUrsache) {
	if istMarker(k) {
		return wertRedigiert(), ""
	}
	z, u := zahlVon(k, true)
	if u != "" {
		return model.Wert{}, u
	}
	text, ok := z.dezimalText()
	if !ok {
		return model.Wert{}, model.DatenZahlZuLang
	}
	return model.Wert{Art: model.ArtDecimal, Text: text}, ""
}

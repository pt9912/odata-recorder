package mock

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// jsonText schreibt einen String in der Kanonform (SPEC-127 Nr. 14): `\"`, `\\` und `\u00XX` mit kleinen Hexziffern unter U+0020,
// sonst unverändert als UTF-8.
func jsonText(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"':
			b.WriteString(`\"`)
		case c == '\\':
			b.WriteString(`\\`)
		case c < 0x20:
			fmt.Fprintf(&b, `\u%04x`, c)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// wertJSON ist der Rohtext eines Wertes im Payload (SPEC-127 Nr. 11 a): boolean und int32 als Literal, null als `null`, jede
// andere Art als String; die Art `redigiert` kommt aus dem Leser nie.
func wertJSON(w model.Wert) (string, bool) {
	switch w.Art {
	case model.ArtNull:
		return "null", true
	case model.ArtBoolean, model.ArtInt32:
		return w.Text, true
	case model.ArtString, model.ArtInt64, model.ArtDecimal, model.ArtGUID, model.ArtDate, model.ArtDateTimeOffset:
		return jsonText(w.Text), true
	}
	return "", false
}

// payloadAus schreibt die Mitglieder in der Kanonform: aufsteigend nach den Bytes der Namen, ohne Leerraum.
func payloadAus(mitglieder []mitglied) string {
	slices.SortFunc(mitglieder, func(a, b mitglied) int { return strings.Compare(a.name, b.name) })
	var b strings.Builder
	b.WriteByte('{')
	for i, m := range mitglieder {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(jsonText(m.name))
		b.WriteByte(':')
		b.Write(m.roh)
	}
	b.WriteByte('}')
	return b.String()
}

// typMitglied ist das Mitglied `@odata.type` für einen Typ, der vom Typ des Sets abweicht (SPEC-193 Nr. 3 a).
func typMitglied(setTyp, typ string) []mitglied {
	if typ == setTyp {
		return nil
	}
	return []mitglied{{name: "@odata.type", roh: []byte(jsonText("#" + typ))}}
}

// legeAn bildet den vollständigen Zustand für POST und PUT (SPEC-193 Nr. 3 c): jede wirksame unterstützte Property des Typs der
// Eingabe, die genannten mit ihrem Wert, die nicht genannten nullbaren als `null`, alle bekannt, `complete`.
func legeAn(mu mutation, version int64) (model.EntityState, error) {
	typ, ok := mu.modell.EntityType(mu.eingabe.Typ)
	if !ok {
		return model.EntityState{}, mitCode(model.CodeE9001, nil)
	}
	mitglieder := typMitglied(mu.typ.Name, typ.Name)
	var bekannt []string
	genannt := 0
	for _, p := range typ.Properties {
		if !p.Unterstuetzt() {
			continue
		}
		roh := "null"
		if i := slices.IndexFunc(mu.eingabe.Properties, func(x model.PropertyWert) bool { return x.Name == p.Name }); i >= 0 {
			if roh, ok = wertJSON(mu.eingabe.Properties[i].Wert); !ok {
				return model.EntityState{}, mitCode(model.CodeE9001, nil)
			}
			genannt++
		} else if !p.Nullable {
			return model.EntityState{}, mitCode(model.CodeE9001, nil)
		}
		mitglieder = append(mitglieder, mitglied{name: p.Name, roh: []byte(roh)})
		bekannt = append(bekannt, p.Name)
	}
	if genannt != len(mu.eingabe.Properties) {
		return model.EntityState{}, mitCode(model.CodeE9001, nil)
	}
	slices.Sort(bekannt)
	return model.EntityState{
		WorkspaceID: mu.w.ID, EntitySet: mu.set, EntityKey: mu.schluessel, Payload: payloadAus(mitglieder),
		KnownProperties: bekannt, Completeness: model.EntityComplete, Version: version,
	}, nil
}

// aendere bildet den Zustand für PATCH (SPEC-193 Nr. 3 d): die Mitglieder, die der Body nicht nennt, bleiben mit ihrem Rohtext
// byte-gleich, die genannten tragen den neuen Wert; ein Body ohne Property lässt den Payload unverändert.
func aendere(mu mutation, es model.EntityState, mitglieder []mitglied) (model.EntityState, error) {
	if len(mu.eingabe.Properties) == 0 {
		return es, nil
	}
	bekannt := slices.Clone(es.KnownProperties)
	for _, p := range mu.eingabe.Properties {
		roh, ok := wertJSON(p.Wert)
		if !ok {
			return model.EntityState{}, mitCode(model.CodeE9001, nil)
		}
		if i := slices.IndexFunc(mitglieder, func(m mitglied) bool { return m.name == p.Name }); i >= 0 {
			mitglieder[i].roh = []byte(roh)
		} else {
			mitglieder = append(mitglieder, mitglied{name: p.Name, roh: []byte(roh)})
		}
		bekannt = append(bekannt, p.Name)
	}
	slices.Sort(bekannt)
	es.Payload = payloadAus(mitglieder)
	es.KnownProperties = slices.Compact(bekannt)
	return es, nil
}

// liesEintrag zerlegt den Payload des Eintrags in seine Mitglieder und nennt den Typ des Eintrags: den von `@odata.type`
// genannten, sonst den des Sets (SPEC-193 Nr. 2 a); ok ist falsch bei jedem Verstoß der Form.
func liesEintrag(payload, setTyp string) ([]mitglied, string, bool) {
	mitglieder, ok := zerlege(payload)
	if !ok {
		return nil, "", false
	}
	i := slices.IndexFunc(mitglieder, func(m mitglied) bool { return m.name == "@odata.type" })
	if i < 0 {
		return mitglieder, setTyp, true
	}
	var name string
	if err := json.Unmarshal(mitglieder[i].roh, &name); err != nil || !strings.HasPrefix(name, "#") {
		return nil, "", false
	}
	return mitglieder, name[1:], true
}

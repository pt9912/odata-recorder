package mock

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// artTyp ist die Art des Wertes zu einem unterstützten EDM-Typ (SPEC-196 Nr. 3 b): drei Typen stehen nicht als String.
func artTyp(edm string) model.WertArt {
	return map[string]model.WertArt{
		"Edm.String": model.ArtString, "Edm.Boolean": model.ArtBoolean, "Edm.Int32": model.ArtInt32, "Edm.Int64": model.ArtInt64,
		"Edm.Decimal": model.ArtDecimal, "Edm.Guid": model.ArtGUID, "Edm.Date": model.ArtDate, "Edm.DateTimeOffset": model.ArtDateTimeOffset,
	}[edm]
}

// mitglied ist ein Mitglied des Payloads: sein Name und der rohe Wert.
type mitglied struct {
	name string
	roh  []byte
}

// deuteListe deutet die Einträge ohne Tombstones (SPEC-196 Nr. 3) und prüft die Pflichten 3 e und 3 f (Nr. 4): jede Entity
// trägt ihre Key-Properties mit einem Wert, `Schluessel` ist nichtleer und je Entity eindeutig; jeder Verstoß ist ODR-E9001.
func deuteListe(l lauf, liste []model.EntityState) ([]model.Entity, error) {
	entities := []model.Entity{}
	gesehen := map[string]bool{}
	for _, es := range liste {
		if es.Deleted {
			continue
		}
		e, ok := deute(l.modell, l.typ, es)
		if !ok || es.EntityKey == "" || gesehen[es.EntityKey] || !schluesselProperties(l.modell, e) {
			return nil, mitCode(model.CodeE9001, nil)
		}
		gesehen[es.EntityKey] = true
		entities = append(entities, e)
	}
	return entities, nil
}

// schluesselProperties sagt, ob e alle Keys seines Typs mit einem Wert außer `null` und `redigiert` trägt (SPEC-196 Nr. 4 c).
func schluesselProperties(modell model.ServiceModel, e model.Entity) bool {
	t, _ := modell.EntityType(e.Typ)
	return !slices.ContainsFunc(t.Keys, func(k string) bool {
		i := slices.IndexFunc(e.Properties, func(p model.PropertyWert) bool { return p.Name == k })
		return i < 0 || e.Properties[i].Wert.Art == model.ArtNull || e.Properties[i].Wert.Art == model.ArtRedigiert
	})
}

// deute bildet die Entity aus dem Payload eines Eintrags (SPEC-196 Nr. 3); ok ist falsch bei jedem Verstoß der Form.
func deute(modell model.ServiceModel, setTyp model.EntityType, es model.EntityState) (model.Entity, bool) {
	mitglieder, ok := zerlege(es.Payload)
	if !ok {
		return model.Entity{}, false
	}
	typ := setTyp
	if i := slices.IndexFunc(mitglieder, func(m mitglied) bool { return m.name == "@odata.type" }); i >= 0 {
		var name string
		if err := json.Unmarshal(mitglieder[i].roh, &name); err != nil || !strings.HasPrefix(name, "#") {
			return model.Entity{}, false
		}
		if typ, ok = abgeleitet(modell, setTyp, name[1:]); !ok {
			return model.Entity{}, false
		}
		mitglieder = slices.Delete(mitglieder, i, i+1)
	}
	for _, m := range mitglieder {
		if !slices.ContainsFunc(typ.Properties, func(p model.Property) bool { return p.Name == m.name && p.Unterstuetzt() }) {
			return model.Entity{}, false
		}
	}
	props := []model.PropertyWert{}
	for _, p := range typ.Properties {
		i := slices.IndexFunc(mitglieder, func(m mitglied) bool { return m.name == p.Name })
		if i < 0 || !p.Unterstuetzt() {
			continue
		}
		w, ok := leseWert(p, mitglieder[i].roh)
		if !ok {
			return model.Entity{}, false
		}
		props = append(props, model.PropertyWert{Name: p.Name, Wert: w})
	}
	return model.Entity{Typ: typ.Name, Schluessel: es.EntityKey, Properties: props, Uebersprungen: []model.Uebersprungen{}}, true
}

// zerlege liest das eine JSON-Objekt des Payloads in seine Mitglieder; ein Mitglied zweimal ist ein Verstoß.
func zerlege(payload string) ([]mitglied, bool) {
	dec := json.NewDecoder(strings.NewReader(payload))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, false
	}
	var r []mitglied
	for dec.More() {
		t, err := dec.Token()
		name, istName := t.(string)
		var roh json.RawMessage
		if err != nil || !istName || dec.Decode(&roh) != nil || slices.ContainsFunc(r, func(m mitglied) bool { return m.name == name }) {
			return nil, false
		}
		r = append(r, mitglied{name: name, roh: bytes.TrimSpace(roh)})
	}
	if t, err := dec.Token(); err != nil || t != json.Delim('}') {
		return nil, false
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, false
	}
	return r, true
}

// abgeleitet löst name auf: ein nicht abstrakter Typ des Modells, der Typ des Sets oder über BaseType von ihm abgeleitet
// (die Kette je Typ höchstens einmal).
func abgeleitet(modell model.ServiceModel, setTyp model.EntityType, name string) (model.EntityType, bool) {
	t, ok := modell.EntityType(name)
	if !ok || t.Abstract {
		return model.EntityType{}, false
	}
	var besucht []string
	for c := t; ; {
		if c.Name == setTyp.Name {
			return t, true
		}
		if c.BaseType == "" || slices.Contains(besucht, c.Name) {
			return model.EntityType{}, false
		}
		besucht = append(besucht, c.Name)
		if c, ok = modell.EntityType(c.BaseType); !ok {
			return model.EntityType{}, false
		}
	}
}

// leseWert liest den rohen Wert einer Property nach ihrem EDM-Typ (SPEC-196 Nr. 3 b), nie über float64.
func leseWert(p model.Property, roh []byte) (model.Wert, bool) {
	art := artTyp(p.EdmType)
	switch {
	case string(roh) == "null":
		return model.Wert{Art: model.ArtNull}, p.Nullable
	case art == model.ArtBoolean:
		ok := string(roh) == "true" || string(roh) == "false"
		return model.Wert{Art: art, Text: string(roh)}, ok
	case art == model.ArtInt32:
		ok := len(roh) > 0 && (roh[0] == '-' || roh[0] >= '0' && roh[0] <= '9')
		return model.Wert{Art: art, Text: string(roh)}, ok
	}
	var s string
	if len(roh) == 0 || roh[0] != '"' || json.Unmarshal(roh, &s) != nil {
		return model.Wert{}, false
	}
	if s == model.Marker {
		return model.Wert{Art: model.ArtRedigiert}, true
	}
	return model.Wert{Art: art, Text: s}, true
}

// maske ist die Property-Maske `beobachtet` (SPEC-196 Nr. 4 a): die Vereinigung der KnownProperties aller Einträge der
// Antwort, Tombstones eingeschlossen; ohne Eintrag nil bei `complete` (alles beobachtet), sonst die leere Maske.
func maske(liste []model.EntityState, status model.CoverageStatus) func(string) bool {
	if len(liste) == 0 {
		if status == model.CoverageComplete {
			return nil
		}
		return func(string) bool { return false }
	}
	bekannt := map[string]bool{}
	for _, es := range liste {
		for _, n := range es.KnownProperties {
			bekannt[n] = true
		}
	}
	return func(name string) bool { return bekannt[name] }
}

// redigiertFrei sagt, ob keine Entity in einer Property der Optionen `$filter` und `$orderby` einen Wert der Art `redigiert`
// trägt (SPEC-196 Nr. 4 e).
func redigiertFrei(entities []model.Entity, opt model.Optionen) bool {
	var namen []string
	if opt.Filter != nil {
		namen = filterNamen(opt.Filter.Ausdruck, namen)
	}
	for _, o := range opt.OrderBy {
		namen = append(namen, o.Name)
	}
	return !slices.ContainsFunc(entities, func(e model.Entity) bool {
		return slices.ContainsFunc(e.Properties, func(p model.PropertyWert) bool {
			return p.Wert.Art == model.ArtRedigiert && slices.Contains(namen, p.Name)
		})
	})
}

// filterNamen sammelt die Namen der Eigenschaften an jeder Stelle des Baums.
func filterNamen(a model.FilterAusdruck, namen []string) []string {
	switch x := a.(type) {
	case model.FilterEigenschaft:
		return append(namen, x.Name)
	case model.FilterNicht:
		return filterNamen(x.Operand, namen)
	case model.FilterVerknuepfung:
		return filterNamen(x.Rechts, filterNamen(x.Links, namen))
	case model.FilterVergleich:
		return filterNamen(x.Rechts, filterNamen(x.Links, namen))
	}
	return namen
}

// ohneRedigiert entfernt die Werte der Art `redigiert` aus den Entities des Ergebnisses (SPEC-196 Nr. 4 e).
func ohneRedigiert(entities []model.Entity) []model.Entity {
	r := make([]model.Entity, len(entities))
	for i, e := range entities {
		e.Properties = slices.DeleteFunc(slices.Clone(e.Properties), func(p model.PropertyWert) bool { return p.Wert.Art == model.ArtRedigiert })
		r[i] = e
	}
	return r
}

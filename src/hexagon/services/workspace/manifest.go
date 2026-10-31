package workspace

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// eintrag ist ein geprüfter Eintrag von entities im Manifest (SPEC-129 Nr. 6 c); payload sind die
// Bytes des Gliedes, wie das Manifest sie trägt.
type eintrag struct {
	set, key     string
	payload      string
	props        []string
	completeness model.Completeness
}

// verstoss ist ein Verstoß des Manifests: ODR-E3001 mit der Ursache ErrBeschaedigt; ort nennt Glied
// und Stelle, nie einen Wert (SPEC-129 Nr. 6).
func verstoss(ort string, a ...any) error {
	return fmt.Errorf("%w: Manifest: %s", outbound.ErrBeschaedigt, fmt.Sprintf(ort, a...))
}

// objekt zerlegt raw in die Glieder eines JSON-Objekts; null und jede andere Art sind keines.
func objekt(raw json.RawMessage, ort string) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if len(raw) == 0 || raw[0] != '{' || json.Unmarshal(raw, &m) != nil {
		return nil, verstoss("%s ist kein Objekt", ort)
	}
	return m, nil
}

// zeichenkette liest das Glied name von m als JSON-String.
func zeichenkette(m map[string]json.RawMessage, name, ort string) (string, error) {
	raw := m[name]
	var s string
	if len(raw) == 0 || raw[0] != '"' || json.Unmarshal(raw, &s) != nil {
		return "", verstoss("%s.%s fehlt oder ist keine Zeichenkette", ort, name)
	}
	return s, nil
}

// ganzzahl liest raw als Ganzzahl in der Form der Kanonform: Ziffern ohne Vorzeichen, Bruch, Exponent
// und führende Null (die 0 gilt).
func ganzzahl(raw json.RawMessage) (int, bool) {
	if len(raw) == 0 || (len(raw) > 1 && raw[0] == '0') {
		return 0, false
	}
	for _, b := range raw {
		if b < '0' || b > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(string(raw))
	return n, err == nil
}

// leseManifest prüft das Manifest b der Baseline snap und liefert seine Einträge in der Folge des
// Manifests (SPEC-129 Nr. 6). Glieder, die Nr. 6 nicht nennt, bleiben ungelesen.
func leseManifest(b []byte, snap model.BaselineSnapshot) ([]eintrag, error) {
	if !utf8.Valid(b) || !json.Valid(b) {
		return nil, verstoss("kein JSON-Wert in gültigem UTF-8")
	}
	kopf, err := objekt(b[firstNonSpace(b):], "Manifest")
	if err != nil {
		return nil, err
	}
	elemente, err := leseKopf(kopf, snap)
	if err != nil {
		return nil, err
	}
	ergebnis := make([]eintrag, 0, len(elemente))
	gesehen := make(map[[2]string]bool, len(elemente))
	for i, raw := range elemente {
		e, err := leseEintrag(raw, i)
		if err != nil {
			return nil, err
		}
		k := [2]string{e.set, e.key}
		if gesehen[k] {
			return nil, verstoss("entities[%d] doppelt", i)
		}
		gesehen[k] = true
		ergebnis = append(ergebnis, e)
	}
	return ergebnis, nil
}

// leseKopf prüft den Kopf des Manifests (SPEC-129 Nr. 6 b) und liefert die Elemente von entities.
func leseKopf(kopf map[string]json.RawMessage, snap model.BaselineSnapshot) ([]json.RawMessage, error) {
	if string(kopf["manifest_version"]) != "1" {
		return nil, verstoss("manifest_version ist nicht die Ganzzahl 1")
	}
	for _, g := range [][2]string{{"snapshot_kind", "baseline"}, {"service_id", snap.ServiceID}, {"baseline_snapshot_id", snap.ID}} {
		s, err := zeichenkette(kopf, g[0], "Manifest")
		if err != nil {
			return nil, err
		}
		if s != g[1] {
			return nil, verstoss("%s passt nicht zur Baseline", g[0])
		}
	}
	rawEntities := kopf["entities"]
	var elemente []json.RawMessage
	if len(rawEntities) == 0 || rawEntities[0] != '[' || json.Unmarshal(rawEntities, &elemente) != nil {
		return nil, verstoss("entities fehlt oder ist kein Array")
	}
	if n, ok := ganzzahl(kopf["entity_count"]); !ok || n != len(elemente) {
		return nil, verstoss("entity_count ist keine Ganzzahl oder ungleich der Zahl der Elemente")
	}
	return elemente, nil
}

func firstNonSpace(b []byte) int {
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\t' || b[i] == '\r' || b[i] == '\n') {
		i++
	}
	return i
}

// leseEintrag prüft das Element i von entities (SPEC-129 Nr. 6 c).
func leseEintrag(raw json.RawMessage, i int) (eintrag, error) {
	ort := fmt.Sprintf("entities[%d]", i)
	m, err := objekt(raw, ort)
	if err != nil {
		return eintrag{}, err
	}
	var e eintrag
	if e.set, err = zeichenkette(m, "entity_set", ort); err != nil {
		return eintrag{}, err
	}
	if e.key, err = zeichenkette(m, "typed_key", ort); err != nil {
		return eintrag{}, err
	}
	if e.set == "" || e.key == "" {
		return eintrag{}, verstoss("%s: entity_set oder typed_key ist leer", ort)
	}
	c, err := zeichenkette(m, "completeness", ort)
	if err != nil {
		return eintrag{}, err
	}
	if e.completeness = model.Completeness(c); e.completeness != model.EntityComplete && e.completeness != model.EntityPartial {
		return eintrag{}, verstoss("%s.completeness ist unbekannt", ort)
	}
	if p := m["payload"]; len(p) == 0 || p[0] != '{' {
		return eintrag{}, verstoss("%s.payload fehlt oder ist kein Objekt", ort)
	}
	e.payload = string(m["payload"])
	if e.props, err = leseProperties(m["selected_properties"], ort); err != nil {
		return eintrag{}, err
	}
	return e, nil
}

// leseProperties liest selected_properties: ein Array nichtleerer, verschiedener Zeichenketten.
func leseProperties(raw json.RawMessage, ort string) ([]string, error) {
	var elemente []json.RawMessage
	if len(raw) == 0 || raw[0] != '[' || json.Unmarshal(raw, &elemente) != nil {
		return nil, verstoss("%s.selected_properties fehlt oder ist kein Array", ort)
	}
	props := make([]string, 0, len(elemente))
	gesehen := make(map[string]bool, len(elemente))
	for _, el := range elemente {
		var s string
		if len(el) == 0 || el[0] != '"' || json.Unmarshal(el, &s) != nil || s == "" || gesehen[s] {
			return nil, verstoss("%s.selected_properties enthält ein Element, das leer, doppelt oder keine Zeichenkette ist", ort)
		}
		gesehen[s] = true
		props = append(props, s)
	}
	return props, nil
}

// abbilden bildet die Einträge auf den Entity-State des Workspace id ab (SPEC-129 Nr. 6): Version 0,
// nicht gelöscht, in der Folge des Manifests; die Liste ist nie nil.
func abbilden(id string, es []eintrag) []model.EntityState {
	st := make([]model.EntityState, 0, len(es))
	for _, e := range es {
		st = append(st, model.EntityState{
			WorkspaceID: id, EntitySet: e.set, EntityKey: e.key, Payload: e.payload,
			KnownProperties: e.props, Completeness: e.completeness, Version: 0, Deleted: false,
		})
	}
	return st
}

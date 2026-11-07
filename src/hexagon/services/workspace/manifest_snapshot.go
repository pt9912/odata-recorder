package workspace

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// leseSnapshotManifest prüft das Manifest b des Workspace-Snapshots snap und liefert seine Einträge in der
// Folge des Manifests (SPEC-129 Nr. 20): Syntax, Kopf, Einträge, Eindeutigkeit und die Invarianten von
// EntityState. Glieder, die Nr. 20 nicht nennt, bleiben ungelesen; jeder Verstoß nennt den Ort, nie einen Wert.
func leseSnapshotManifest(b []byte, snap model.WorkspaceSnapshot, workspaceID string) ([]eintrag, error) {
	if !utf8.Valid(b) || !json.Valid(b) {
		return nil, verstoss("kein JSON-Wert in gültigem UTF-8")
	}
	kopf, err := objekt(b[firstNonSpace(b):], "Manifest")
	if err != nil {
		return nil, err
	}
	elemente, err := leseSnapshotKopf(kopf, snap)
	if err != nil {
		return nil, err
	}
	ergebnis := make([]eintrag, 0, len(elemente))
	gesehen := make(map[[2]string]bool, len(elemente))
	for i, raw := range elemente {
		e, err := leseSnapshotEintrag(raw, i)
		if err != nil {
			return nil, err
		}
		k := [2]string{e.set, e.key}
		if gesehen[k] {
			return nil, verstoss("entities[%d] doppelt", i)
		}
		gesehen[k] = true
		if err := abbildenEintrag(workspaceID, e).Validate(); err != nil {
			return nil, verstoss("entities[%d] verletzt die Invarianten des Eintrags", i)
		}
		ergebnis = append(ergebnis, e)
	}
	return ergebnis, nil
}

// leseSnapshotKopf prüft den Kopf (SPEC-129 Nr. 20 b) und liefert die Elemente von entities.
func leseSnapshotKopf(kopf map[string]json.RawMessage, snap model.WorkspaceSnapshot) ([]json.RawMessage, error) {
	if string(kopf["manifest_version"]) != "1" {
		return nil, verstoss("manifest_version ist nicht die Ganzzahl 1")
	}
	for _, g := range [][2]string{{"snapshot_kind", "workspace"}, {"service_id", snap.ServiceID}, {"metadata_hash", snap.MetadataHash}} {
		if err := gliedGleich(kopf, g[0], g[1]); err != nil {
			return nil, err
		}
	}
	return leseEntities(kopf)
}

// gliedGleich prüft, dass das Glied name des Kopfes die Zeichenkette soll ist.
func gliedGleich(kopf map[string]json.RawMessage, name, soll string) error {
	s, err := zeichenkette(kopf, name, "Manifest")
	if err != nil {
		return err
	}
	if s != soll {
		return verstoss("%s passt nicht zum Snapshot", name)
	}
	return nil
}

// leseDeleted liest deleted als Wahrheitswert; leseVersion liest version als Ganzzahl im Bereich von int64.
func leseDeleted(m map[string]json.RawMessage, ort string) (bool, error) {
	switch string(m["deleted"]) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, verstoss("%s.deleted fehlt oder ist kein Wahrheitswert", ort)
}

func leseVersion(m map[string]json.RawMessage, ort string) (int64, error) {
	v := m["version"]
	if len(v) == 0 || v[0] == '"' {
		return 0, verstoss("%s.version fehlt oder ist keine Ganzzahl", ort)
	}
	n, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		return 0, verstoss("%s.version ist keine Ganzzahl im Bereich von int64", ort)
	}
	return n, nil
}

// leseSnapshotEintrag prüft das Element i von entities (SPEC-129 Nr. 20 c).
func leseSnapshotEintrag(raw json.RawMessage, i int) (eintrag, error) {
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
	if e.deleted, err = leseDeleted(m, ort); err != nil {
		return eintrag{}, err
	}
	if e.version, err = leseVersion(m, ort); err != nil {
		return eintrag{}, err
	}
	if len(m["payload"]) == 0 {
		return eintrag{}, verstoss("%s.payload fehlt", ort)
	}
	e.payload = string(m["payload"])
	if e.props, err = leseProperties(m["selected_properties"], ort); err != nil {
		return eintrag{}, err
	}
	return e, nil
}

// abbildenEintrag bildet einen Eintrag auf den EntityState des Workspace id ab (SPEC-129 Nr. 20 e).
func abbildenEintrag(id string, e eintrag) model.EntityState {
	return model.EntityState{
		WorkspaceID: id, EntitySet: e.set, EntityKey: e.key, Payload: e.payload,
		KnownProperties: e.props, Completeness: e.completeness, Version: e.version, Deleted: e.deleted,
	}
}

// abbildenSnapshot bildet die Einträge des Manifests in dessen Folge auf den Entity-State des Workspace id
// ab, mit Version und Tombstone des Manifests; die Liste ist nie nil.
func abbildenSnapshot(id string, es []eintrag) []model.EntityState {
	st := make([]model.EntityState, 0, len(es))
	for _, e := range es {
		st = append(st, abbildenEintrag(id, e))
	}
	return st
}

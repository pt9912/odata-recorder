package model

import (
	"encoding/json"
	"slices"
	"time"
	"unicode/utf8"
)

// Workspace ist ein veränderlicher Arbeitsstand über einer Baseline. Revision zählt ab 0
// (SPEC-108).
type Workspace struct {
	ID                 string
	ServiceID          string
	BaselineSnapshotID string
	Revision           int64
	CreatedAt          time.Time
}

// Validate prüft die Invarianten eines neuen Workspace: Kennungen gesetzt, Zeit gültig,
// Revision 0.
func (w Workspace) Validate() error {
	switch {
	case w.ID == "" || w.ServiceID == "" || w.BaselineSnapshotID == "":
		return ungueltig("Workspace ohne Kennung, Service oder Baseline")
	case !zeitGueltig(w.CreatedAt):
		return ungueltig("Workspace ohne gültige Erzeugungszeit")
	case w.Revision != 0:
		return ungueltig("neuer Workspace muss Revision 0 tragen, nicht %d", w.Revision)
	}
	return nil
}

// Completeness sagt, ob ein Entity-Eintrag alle Properties der Entity trägt (SPEC-044).
type Completeness string

// Zustände der Vollständigkeit.
const (
	EntityComplete Completeness = "complete"
	EntityPartial  Completeness = "partial"
)

// EntityState ist ein Eintrag des Entity-State eines Workspace; Deleted kennzeichnet einen
// Tombstone. Payload ist frei geformtes JSON und wird byte-gleich gespeichert; der Schlüssel
// ist ein vom Aufrufer kanonisch gebildeter Text.
type EntityState struct {
	WorkspaceID     string
	EntitySet       string
	EntityKey       string
	Payload         string
	KnownProperties []string
	Completeness    Completeness
	Version         int64
	Deleted         bool
}

// Validate prüft die Invarianten eines Eintrags (SPEC-108 Nr. 6).
func (e EntityState) Validate() error {
	switch {
	case e.WorkspaceID == "" || e.EntitySet == "" || e.EntityKey == "":
		return ungueltig("Entity ohne Workspace, entity_set oder entity_key")
	case !jsonGueltig(e.Payload):
		return ungueltig("payload von %s/%s ist kein gültiges JSON in UTF-8", e.EntitySet, e.EntityKey)
	case !eines(e.Completeness, EntityComplete, EntityPartial):
		return ungueltig("unbekannte Vollständigkeit %q", e.Completeness)
	case e.Version < 0:
		return ungueltig("negative Version %d", e.Version)
	}
	return e.pruefeProperties()
}

func (e EntityState) pruefeProperties() error {
	gesehen := map[string]bool{}
	for _, p := range e.KnownProperties {
		switch {
		case p == "":
			return ungueltig("known_properties enthält eine leere Zeichenkette")
		case !utf8.ValidString(p):
			return ungueltig("known_properties enthält ungültiges UTF-8")
		case gesehen[p]:
			return ungueltig("known_properties enthält %q doppelt", p)
		}
		gesehen[p] = true
	}
	return nil
}

// KnownPropertiesJSON liefert known_properties als JSON-Array, aufsteigend sortiert; ohne
// Properties ist es "[]". Der Eintrag selbst bleibt unverändert.
func (e EntityState) KnownPropertiesJSON() string {
	sortiert := slices.Sorted(slices.Values(e.KnownProperties))
	if sortiert == nil {
		sortiert = []string{}
	}
	b, _ := json.Marshal(sortiert)
	return string(b)
}

// ReplayStrategy ist die Matching-Strategie eines Cursors.
type ReplayStrategy string

// Strategien des Replay.
const (
	StrategySequential ReplayStrategy = "sequential"
	StrategyStrict     ReplayStrategy = "strict"
	StrategyFirst      ReplayStrategy = "first"
	StrategyLast       ReplayStrategy = "last"
)

// ReplayCursor ist die Position einer Replay-Session in einer Recording-Session eines
// Workspace. Position ist die Anzahl der verbrauchten sequence_no in Consumed.
type ReplayCursor struct {
	WorkspaceID        string
	ReplaySessionID    string
	RecordingSessionID string
	ServiceID          string
	Strategy           ReplayStrategy
	Position           int64
	Consumed           []int64
}

// Validate prüft die Invarianten eines Cursors (SPEC-108 Nr. 10).
func (c ReplayCursor) Validate() error {
	switch {
	case c.WorkspaceID == "" || c.ReplaySessionID == "" || c.RecordingSessionID == "":
		return ungueltig("Cursor ohne Workspace, Replay-Session oder Recording-Session")
	case c.ServiceID == "":
		return ungueltig("Cursor ohne Service")
	case !eines(c.Strategy, StrategySequential, StrategyStrict, StrategyFirst, StrategyLast):
		return ungueltig("unbekannte Strategie %q", c.Strategy)
	case c.Position != int64(len(c.Consumed)):
		return ungueltig("position %d ist nicht die Anzahl der verbrauchten Exchanges (%d)", c.Position, len(c.Consumed))
	}
	gesehen := map[int64]bool{}
	for _, n := range c.Consumed {
		switch {
		case n <= 0:
			return ungueltig("consumed enthält %d", n)
		case gesehen[n]:
			return ungueltig("consumed enthält %d doppelt", n)
		}
		gesehen[n] = true
	}
	return nil
}

// ConsumedJSON liefert consumed als JSON-Array, aufsteigend sortiert; ohne Einträge "[]".
// Der Cursor selbst bleibt unverändert.
func (c ReplayCursor) ConsumedJSON() string {
	sortiert := slices.Sorted(slices.Values(c.Consumed))
	if sortiert == nil {
		sortiert = []int64{}
	}
	b, _ := json.Marshal(sortiert)
	return string(b)
}

// WorkspaceSnapshot ist ein unveränderlicher Snapshot der Art workspace.
type WorkspaceSnapshot struct {
	ID           string
	ServiceID    string
	WorkspaceID  string
	Name         string
	CreatedAt    time.Time
	StateBlob    string
	StateHash    string
	MetadataHash string
}

// Validate prüft die Invarianten eines Workspace-Snapshots (SPEC-108 Nr. 13).
func (s WorkspaceSnapshot) Validate() error {
	switch {
	case s.ID == "" || s.ServiceID == "" || s.WorkspaceID == "" || s.Name == "":
		return ungueltig("Snapshot ohne Kennung, Service, Workspace oder Name")
	case !zeitGueltig(s.CreatedAt):
		return ungueltig("Snapshot ohne gültige Erzeugungszeit")
	case !BlobReferenzGueltig(s.StateBlob):
		return ungueltig("state_blob %q ist keine Blob-Referenz", s.StateBlob)
	case s.StateHash == "" || s.MetadataHash == "":
		return ungueltig("Snapshot ohne state_hash oder metadata_hash")
	}
	return nil
}

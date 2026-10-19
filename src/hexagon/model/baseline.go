package model

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

// Service ist ein beobachteter OData-Dienst. MetadataBlob und MetadataHash sind beide
// gesetzt oder beide leer (SPEC-107).
type Service struct {
	ID           string
	ServiceRoot  string
	ODataVersion string
	MetadataBlob string
	MetadataHash string
}

// BlobReferenzGueltig sagt, ob ref aus "sha256:" und 64 Kleinbuchstaben-Hexziffern besteht
// (SPEC-106). Das Vorhandensein der Datei prüft sie nicht.
func BlobReferenzGueltig(ref string) bool {
	hex, ok := strings.CutPrefix(ref, "sha256:")
	if !ok || len(hex) != 64 {
		return false
	}
	for _, r := range hex {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// MetadatenGueltig prüft das Paar aus Blob-Referenz und Hash: beide leer oder beide
// gesetzt, die Referenz gültig.
func MetadatenGueltig(blob, hash string) error {
	switch {
	case (blob == "") != (hash == ""):
		return ungueltig("metadata_blob und metadata_hash sind beide gesetzt oder beide leer")
	case blob != "" && !BlobReferenzGueltig(blob):
		return ungueltig("metadata_blob %q ist keine Blob-Referenz", blob)
	}
	return nil
}

// Validate prüft die Invarianten eines neuen Service (SPEC-107 Nr. 1).
func (s Service) Validate() error {
	switch {
	case s.ID == "":
		return ungueltig("Service ohne Kennung")
	case s.ServiceRoot == "":
		return ungueltig("Service ohne service_root")
	case s.ODataVersion == "":
		return ungueltig("Service ohne odata_version")
	}
	return MetadatenGueltig(s.MetadataBlob, s.MetadataHash)
}

// BaselineSnapshot ist ein unveränderlicher Snapshot der Art baseline.
type BaselineSnapshot struct {
	ID           string
	ServiceID    string
	CreatedAt    time.Time
	StateBlob    string
	StateHash    string
	MetadataHash string
}

// Validate prüft die Invarianten eines Baseline-Snapshots.
func (b BaselineSnapshot) Validate() error {
	switch {
	case b.ID == "" || b.ServiceID == "":
		return ungueltig("Snapshot ohne Kennung oder Service")
	case !zeitGueltig(b.CreatedAt):
		return ungueltig("Snapshot ohne gültige Erzeugungszeit")
	case !BlobReferenzGueltig(b.StateBlob):
		return ungueltig("state_blob %q ist keine Blob-Referenz", b.StateBlob)
	case b.StateHash == "" || b.MetadataHash == "":
		return ungueltig("Snapshot ohne state_hash oder metadata_hash")
	}
	return nil
}

// Import ist der Eintrag eines erfolgreichen Imports; BaselineSnapshotID nennt die Baseline,
// die er erzeugte.
type Import struct {
	ID                 string
	RequestHash        string
	BaselineSnapshotID string
	ServiceID          string
	CreatedAt          time.Time
	ReportJSON         string
}

// jsonGueltig sagt, ob s gültiges JSON in gültigem UTF-8 ist; json.Valid lässt ungültiges
// UTF-8 in Zeichenketten zu.
func jsonGueltig(s string) bool { return utf8.ValidString(s) && json.Valid([]byte(s)) }

// Validate prüft die Invarianten eines Import-Eintrags.
func (i Import) Validate() error {
	switch {
	case i.ID == "" || i.ServiceID == "" || i.BaselineSnapshotID == "":
		return ungueltig("Import ohne Kennung, Service oder Baseline")
	case i.RequestHash == "":
		return ungueltig("Import ohne request_hash")
	case !jsonGueltig(i.ReportJSON):
		return ungueltig("report_json ist kein gültiges JSON in UTF-8")
	case !zeitGueltig(i.CreatedAt):
		return ungueltig("Import ohne gültige Erzeugungszeit")
	}
	return nil
}

// CoverageStatus sagt, wie vollständig ein Entity-Set in der Baseline belegt ist (SPEC-044).
type CoverageStatus string

// Zustände der Coverage.
const (
	CoverageComplete CoverageStatus = "complete"
	CoveragePartial  CoverageStatus = "partial"
	CoverageUnknown  CoverageStatus = "unknown"
)

// Coverage ist der Eintrag eines Entity-Set einer Baseline. EvidenceJSON ist frei
// geformtes JSON und wird byte-gleich gespeichert.
type Coverage struct {
	EntitySet    string
	Status       CoverageStatus
	EvidenceJSON string
}

// Validate prüft die Invarianten eines Coverage-Eintrags.
func (c Coverage) Validate() error {
	switch {
	case c.EntitySet == "":
		return ungueltig("Coverage ohne entity_set")
	case !eines(c.Status, CoverageComplete, CoveragePartial, CoverageUnknown):
		return ungueltig("unbekannter Coverage-Status %q", c.Status)
	case !jsonGueltig(c.EvidenceJSON):
		return ungueltig("Evidenz von %s ist kein gültiges JSON in UTF-8", c.EntitySet)
	}
	return nil
}

// Publication ist alles, was ein Import in einer Transaktion veröffentlicht (SPEC-107).
type Publication struct {
	Import           Import
	Snapshot         BaselineSnapshot
	SourceSessionIDs []string
	Coverage         []Coverage
}

// Validate prüft die Gültigkeit der Publikation (SPEC-107 Nr. 6).
func (p Publication) Validate() error {
	for _, pruefung := range []func() error{p.Snapshot.Validate, p.Import.Validate, p.pruefeBezug, p.pruefeQuellen, p.pruefeCoverage} {
		if err := pruefung(); err != nil {
			return err
		}
	}
	return nil
}

func (p Publication) pruefeBezug() error {
	if p.Import.ServiceID != p.Snapshot.ServiceID || p.Import.BaselineSnapshotID != p.Snapshot.ID {
		return ungueltig("Import und Snapshot gehören nicht zusammen")
	}
	return nil
}

func (p Publication) pruefeQuellen() error {
	if len(p.SourceSessionIDs) == 0 {
		return ungueltig("Import ohne Quell-Session")
	}
	gesehen := map[string]bool{}
	for _, id := range p.SourceSessionIDs {
		if gesehen[id] {
			return ungueltig("Quell-Session %s doppelt", id)
		}
		gesehen[id] = true
	}
	return nil
}

func (p Publication) pruefeCoverage() error {
	gesehen := map[string]bool{}
	for _, c := range p.Coverage {
		if err := c.Validate(); err != nil {
			return err
		}
		if gesehen[c.EntitySet] {
			return ungueltig("Coverage für %s doppelt", c.EntitySet)
		}
		gesehen[c.EntitySet] = true
	}
	return nil
}

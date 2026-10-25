package baselineimport

import (
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// MaxImportIDLaenge ist die größte Länge einer import_id in Bytes (SPEC-127 Nr. 3 a).
const MaxImportIDLaenge = 64

const (
	richtlinieReject = "reject"
	zielNeueBaseline = "new_baseline"
)

// idGueltig sagt, ob id aus 1 bis MaxImportIDLaenge Bytes aus 0-9, a-z und - besteht.
func idGueltig(id string) bool {
	if id == "" || len(id) > MaxImportIDLaenge {
		return false
	}
	for i := 0; i < len(id); i++ {
		b := id[i]
		if (b < '0' || b > '9') && (b < 'a' || b > 'z') && b != '-' {
			return false
		}
	}
	return true
}

// pruefeAnfrage prüft die Anfrage nach SPEC-127 Nr. 3 a bis e; ein Verstoß ist model.ErrUngueltig.
func pruefeAnfrage(a inbound.ImportAnfrage) error {
	switch {
	case !idGueltig(a.ImportID):
		return ungueltig("import_id außerhalb der Form")
	case a.ServiceID == "":
		return ungueltig("service_id fehlt")
	case len(a.SourceSessionIDs) != 1 || a.SourceSessionIDs[0] == "":
		return ungueltig("source_session_ids nennt nicht genau eine Session")
	case a.MetadataHash != "" && !model.BlobReferenzGueltig(a.MetadataHash):
		return ungueltig("metadata_hash ist keine Blob-Referenz")
	case a.ConflictPolicy != "" && a.ConflictPolicy != richtlinieReject:
		return ungueltig("conflict_policy ist nicht reject")
	case a.Target != "" && a.Target != zielNeueBaseline:
		return ungueltig("target ist nicht new_baseline")
	}
	return nil
}

// anfrageHash ist der Hash der Anfrage (SPEC-127 Nr. 3 f): der SHA-256 der Kanonform der sechs
// Mitglieder nach den Standardwerten.
func anfrageHash(a inbound.ImportAnfrage) string {
	richtlinie, ziel := a.ConflictPolicy, a.Target
	if richtlinie == "" {
		richtlinie = richtlinieReject
	}
	if ziel == "" {
		ziel = zielNeueBaseline
	}
	return sha256Hex(kanon(objekt{
		"conflict_policy":    richtlinie,
		"import_id":          a.ImportID,
		"metadata_hash":      a.MetadataHash,
		"service_id":         a.ServiceID,
		"source_session_ids": []any{a.SourceSessionIDs[0]},
		"target":             ziel,
	}))
}

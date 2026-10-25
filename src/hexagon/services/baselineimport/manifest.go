package baselineimport

import (
	"slices"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	zeitFormat       = "2006-01-02T15:04:05.000000000Z"
	manifestVersion  = 1
	snapshotKindName = "baseline"
)

// quelle ist ein gelesener Exchange in der Liste der Quellen (SPEC-127 Nr. 14 a).
type quelle struct {
	elemente   int
	set        string
	exchangeID string
	art        string
	seq        int
}

// ergebnis trägt alles, woraus Manifest und Antwortobjekt entstehen.
type ergebnis struct {
	importID   string
	serviceID  string
	sessionID  string
	snapshotID string
	zeit       time.Time
	metaRef    string
	entities   []fertig
	coverage   []model.Coverage
	quellen    []quelle
	// Report
	gelesen        int
	gesamt         int
	uebersprungen  map[string]int
	beobachtungen  int
	maskiert       int
	nichtUnterst   []any
	metadatenFunde []model.Fund
}

func (r ergebnis) coverageListe() []any {
	return liste(r.coverage, func(c model.Coverage) any {
		return objekt{"entity_set": c.EntitySet, "evidence": roh(c.EvidenceJSON), "status": string(c.Status)}
	})
}

func (r ergebnis) quellenListe() []any {
	return liste(r.quellen, func(q quelle) any {
		return objekt{"elements": q.elemente, "entity_set": q.set, "exchange_id": q.exchangeID, "kind": q.art, "sequence_no": q.seq, "session_id": r.sessionID}
	})
}

// manifest ist die Kanonform des Manifests des Snapshots (SPEC-127 Nr. 14 a).
func (r ergebnis) manifest() []byte {
	return kanon(objekt{
		"baseline_snapshot_id": r.snapshotID,
		"coverage":             r.coverageListe(),
		"created_at":           r.zeit.UTC().Format(zeitFormat),
		"entities":             liste(r.entities, func(f fertig) any { return f.json }),
		"entity_count":         len(r.entities),
		"import_id":            r.importID,
		"manifest_version":     manifestVersion,
		"metadata_hash":        r.metaRef,
		"provenance":           r.quellenListe(),
		"service_id":           r.serviceID,
		"snapshot_kind":        snapshotKindName,
		"source_session_ids":   []any{r.sessionID},
	})
}

// antwortObjekt ist die Kanonform des Antwortobjekts, der ReportJSON des Imports (Nr. 14 b).
func (r ergebnis) antwortObjekt() string {
	// uebersprungen trägt nur Gründe, die ein Exchange traf: jeder Zähler ist über 0.
	gruende := make([]string, 0, len(r.uebersprungen))
	for g := range r.uebersprungen {
		gruende = append(gruende, g)
	}
	slices.Sort(gruende)
	funde := liste(r.metadatenFunde, func(f model.Fund) any {
		return objekt{"kind": string(f.Art), "location": f.Ort, "name": f.Name}
	})
	return string(kanon(objekt{
		"baseline_snapshot_id": r.snapshotID,
		"coverage":             r.coverageListe(),
		"entity_count":         len(r.entities),
		"import_id":            r.importID,
		"report": objekt{
			"entities": objekt{"distinct": len(r.entities), "observations": r.beobachtungen},
			"exchanges": objekt{
				"read": r.gelesen, "total": r.gesamt,
				"skipped": liste(gruende, func(g string) any { return objekt{"count": r.uebersprungen[g], "reason": g} }),
			},
			"masked_properties":  r.maskiert,
			"metadata_hash":      r.metaRef,
			"provenance":         r.quellenListe(),
			"service_id":         r.serviceID,
			"source_session_ids": []any{r.sessionID},
			"unsupported_features": objekt{
				"entities": r.nichtUnterst,
				"metadata": funde,
			},
		},
	}))
}

// sortiereQuellen ordnet die Quellen aufsteigend nach sequence_no; das $metadata steht nicht
// zwingend am Anfang der Session.
func sortiereQuellen(q []quelle) {
	slices.SortFunc(q, func(a, b quelle) int { return a.seq - b.seq })
}

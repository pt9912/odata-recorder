package baselineimport

import (
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Gründe der Coverage (reason) und die Enden einer Kette (end), SPEC-127 Nr. 12, 13.
const (
	grundKeineBeobachtung  = "keine_beobachtung"
	grundKeineUngefilterte = "keine_ungefilterte_abfrage"
	grundKetteNichtBelegt  = "kette_nicht_belegt"
	grundKettenAbweichend  = "ketten_abweichend"
	grundKetteBelegt       = "kette_belegt"
	endeBelegt             = "proven"
	evidenzVersion         = 1
)

// antwort ist ein gelesener Exchange der Spur Collection oder Entity (SPEC-127 Nr. 9).
type antwort struct {
	ex          model.Exchange
	spur        spur
	set         string
	elemente    int
	seite       model.EntitySeite
	schluessel  map[string]struct{}
	ungefiltert bool
}

// seitenBeleg ist eine Seite, die eine Kette erreichte und las (Nr. 13).
type seitenBeleg struct {
	elemente   int
	exchangeID string
	seq        int
}

// anzahlBeleg ist @odata.count der ersten Seite gegen die verschiedenen Schlüssel der Kette.
type anzahlBeleg struct {
	deklariert int64
	beobachtet int64
}

// kette ist der Lauf von der ersten Seite einer ungefilterten Abfrage bis zu ihrem Ende (Nr. 12 b).
type kette struct {
	start      *antwort
	seiten     []seitenBeleg
	ende       string
	count      *anzahlBeleg
	schluessel map[string]struct{}
}

// coverageStatus bildet Status und Grund aus dem Nachweis (Nr. 12 g).
func coverageStatus(gelesen int, hatStart bool, ketten []kette) (model.CoverageStatus, string) {
	switch {
	case gelesen == 0:
		return model.CoverageUnknown, grundKeineBeobachtung
	case !hatStart:
		return model.CoveragePartial, grundKeineUngefilterte
	case len(ketten) == 0 || slices.ContainsFunc(ketten, func(k kette) bool { return k.ende != endeBelegt }):
		return model.CoveragePartial, grundKetteNichtBelegt
	case slices.ContainsFunc(ketten, func(k kette) bool { return !gleicheMenge(k.schluessel, ketten[0].schluessel) }):
		return model.CoveragePartial, grundKettenAbweichend
	}
	return model.CoverageComplete, grundKetteBelegt
}

func gleicheMenge(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func kettenObjekt(k kette) any {
	o := objekt{
		"start": objekt{"exchange_id": k.start.ex.ID, "sequence_no": k.start.ex.SequenceNo},
		"pages": liste(k.seiten, func(s seitenBeleg) any {
			return objekt{"elements": s.elemente, "exchange_id": s.exchangeID, "sequence_no": s.seq}
		}),
		"end": k.ende,
	}
	if k.count != nil {
		o["count"] = objekt{"declared": k.count.deklariert, "observed": k.count.beobachtet}
	}
	return o
}

// bewerteSet bildet Status und Evidenz eines Sets aus seinen gelesenen Antworten und der Zahl
// seiner verschiedenen Schlüssel (Nr. 12, 13).
func bewerteSet(k kettenlauf, set string, gelesen []*antwort, entities int) (model.CoverageStatus, objekt) {
	ketten := k.laufeKetten(set, gelesen)
	collection, entity := 0, 0
	hatStart := false
	for _, a := range gelesen {
		if a.spur == spurCollection {
			collection++
		} else {
			entity++
		}
		hatStart = hatStart || a.ungefiltert
	}
	status, grund := coverageStatus(len(gelesen), hatStart, ketten)
	return status, objekt{
		"version":   evidenzVersion,
		"reason":    grund,
		"responses": objekt{"collection": collection, "entity": entity},
		"entities":  entities,
		"chains":    liste(ketten, kettenObjekt),
	}
}

// coverageEintraege bildet je Set des Modells einen Eintrag, aufsteigend nach entity_set (Nr. 12).
func coverageEintraege(m model.ServiceModel, ketten kettenlauf, gelesen []*antwort, z *zusammen) []model.Coverage {
	nachSet := map[string][]*antwort{}
	for _, a := range gelesen {
		nachSet[a.set] = append(nachSet[a.set], a)
	}
	zahl := map[string]int{}
	for k := range z.entities {
		zahl[k.set]++
	}
	sets := m.EntitySets()
	slices.SortFunc(sets, func(a, b model.EntitySet) int { return strings.Compare(a.Name, b.Name) })
	out := make([]model.Coverage, 0, len(sets))
	for _, s := range sets {
		status, evidenz := bewerteSet(ketten, s.Name, nachSet[s.Name], zahl[s.Name])
		out = append(out, model.Coverage{EntitySet: s.Name, Status: status, EvidenceJSON: string(kanon(evidenz))})
	}
	return out
}

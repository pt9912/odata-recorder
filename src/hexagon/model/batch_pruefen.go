package model

import (
	"strconv"
	"strings"
)

// PruefeBatch prüft die Struktur eines gelesenen Batch (SPEC-184): die Methode je Ebene, die Content-ID je Changeset eindeutig, ein
// Verweis nur auf einen früheren Request desselben Changeset. Der erste Verstoß in der Folge der Teile und Requests ist ein
// *BatchFehler mit Ursache und Ort, sonst nil; b bleibt unverändert.
func PruefeBatch(b Batch) error {
	for i, teil := range b.Teile {
		if err := pruefeBatchTeil(teil, strconv.Itoa(i+1)); err != nil {
			return err
		}
	}
	return nil
}

// pruefeBatchTeil beurteilt jeden Request des Teils nach Changeset; ort ist die Nummer des Teils.
func pruefeBatchTeil(teil BatchTeil, ort string) error {
	// frueher hält die Content-IDs der früheren Requests eines Changeset; außerhalb bleibt es leer.
	frueher := map[string]bool{}
	for i, r := range teil.Requests {
		if ursache := batchVerstoss(r, teil.Changeset, frueher); ursache != "" {
			if teil.Changeset {
				ort += "/" + strconv.Itoa(i+1)
			}
			return &BatchFehler{Ursache: ursache, Ort: ort}
		}
		if teil.Changeset && r.ContentID != "" {
			frueher[r.ContentID] = true
		}
	}
	return nil
}

// batchVerstoss liefert die erste Ursache am Request in der Folge Methode, Eindeutigkeit, Auflösbarkeit, ohne Verstoß "".
func batchVerstoss(r BatchRequest, imChangeset bool, frueher map[string]bool) BatchUrsache {
	if ursache := batchMethodenVerstoss(r.Methode, imChangeset); ursache != "" {
		return ursache
	}
	if frueher[r.ContentID] {
		return BatchContentIDDoppelt
	}
	if id := r.Verweis(); id != "" && !frueher[id] {
		return BatchReferenz
	}
	return ""
}

// batchMethodenVerstoss vergleicht bytegleich (SPEC-184 Nr. 1).
func batchMethodenVerstoss(methode string, imChangeset bool) BatchUrsache {
	switch methode {
	case "POST", "PATCH", "PUT", "DELETE":
		return ""
	case "GET":
		if imChangeset {
			return BatchMethode
		}
		return ""
	}
	return BatchMethodeNichtUnterstuetzt
}

// Verweis liefert <id>, wenn Ziel mit `$` beginnt und der Text dahinter bis vor das erste `/` oder `?` die Form einer Content-ID hat
// und keines der reservierten Wörter ist, sonst "" (SPEC-184 Nr. 3). Dekodiert wird nicht.
func (r BatchRequest) Verweis() string {
	id, mitDollar := strings.CutPrefix(r.Ziel, "$")
	if !mitDollar {
		return ""
	}
	if ende := strings.IndexAny(id, "/?"); ende >= 0 {
		id = id[:ende]
	}
	switch id {
	case "metadata", "batch", "entity", "all", "crossjoin":
		return ""
	}
	if !ContentIDGueltig(id) {
		return ""
	}
	return id
}

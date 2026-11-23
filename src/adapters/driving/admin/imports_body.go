package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Die Schlüssel des Bodys von `POST /__admin/imports` (SPEC-297 Nr. 5 c); `service_id` steht in keyServiceID.
const (
	keyImportID       = "import_id"
	keySessionIDs     = "source_session_ids"
	keyMetadataHash   = "metadata_hash"
	keyConflictPolicy = "conflict_policy"
	keyTarget         = "target"
)

// ungueltigerImportBody ist der Fehler eines Verstoßes gegen SPEC-297 Nr. 5 a bis d; der Text nennt keinen Wert.
func ungueltigerImportBody() error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: Body von POST /__admin/imports ungültig"))
}

// leseImport liest den Body von `POST /__admin/imports` (SPEC-297 Nr. 5): ein Kontextfehler beim Lesen bleibt der Fehler;
// jeder andere Lesefehler und jeder Verstoß gegen Form, Schlüssel, Pflicht und Typen ist model.ErrUngueltig. Die Werte gehen
// dekodiert und unverändert in die Anfrage; ein fehlender optionaler Schlüssel ist leer.
func leseImport(r io.Reader) (inbound.ImportAnfrage, error) {
	roh, err := io.ReadAll(r)
	if err != nil {
		if finde(err, istKontextfehler) {
			return inbound.ImportAnfrage{}, err
		}
		return inbound.ImportAnfrage{}, ungueltigerImportBody()
	}
	if !utf8.Valid(roh) || !json.Valid(roh) {
		return inbound.ImportAnfrage{}, ungueltigerImportBody()
	}
	a, ok := leseImportObjekt(json.NewDecoder(bytes.NewReader(roh)))
	if !ok {
		return inbound.ImportAnfrage{}, ungueltigerImportBody()
	}
	return a, nil
}

// leseImportObjekt liest das eine Objekt eines Bodys, den json.Valid angenommen hat: die sechs Schlüssel, jeder höchstens
// einmal (dekodiert gezählt) und die drei Pflichtschlüssel genau einmal, jeder Wert eine Zeichenkette, `source_session_ids` ein
// Array von Zeichenketten. Jeder andere Schlüssel und jeder andere Typ ist ungültig.
func leseImportObjekt(dec *json.Decoder) (a inbound.ImportAnfrage, ok bool) {
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return a, false
	}
	gesehen := map[string]bool{}
	for dec.More() {
		k, err := dec.Token()
		schluessel, istText := k.(string)
		if err != nil || !istText || gesehen[schluessel] || !leseGlied(dec, &a, schluessel) {
			return a, false
		}
		gesehen[schluessel] = true
	}
	return a, gesehen[keyImportID] && gesehen[keyServiceID] && gesehen[keySessionIDs]
}

// leseGlied liest den Wert zum Schlüssel in die Anfrage und meldet, ob der Schlüssel einer der sechs und der Wert vom Typ des
// Feldes ist.
func leseGlied(dec *json.Decoder, a *inbound.ImportAnfrage, schluessel string) bool {
	var ok bool
	switch schluessel {
	case keyImportID:
		a.ImportID, ok = leseText(dec)
	case keyServiceID:
		a.ServiceID, ok = leseText(dec)
	case keySessionIDs:
		a.SourceSessionIDs, ok = leseTextListe(dec)
	case keyMetadataHash:
		a.MetadataHash, ok = leseText(dec)
	case keyConflictPolicy:
		a.ConflictPolicy, ok = leseText(dec)
	case keyTarget:
		a.Target, ok = leseText(dec)
	}
	return ok
}

// leseText liest den nächsten Wert und meldet, ob er eine Zeichenkette ist.
func leseText(dec *json.Decoder) (string, bool) {
	t, err := dec.Token()
	s, istText := t.(string)
	return s, err == nil && istText
}

// leseTextListe liest den nächsten Wert und meldet, ob er ein Array aus Zeichenketten ist; das Ergebnis führt jeden Eintrag in
// seiner Folge, auch leere und gleiche.
func leseTextListe(dec *json.Decoder) ([]string, bool) {
	if t, err := dec.Token(); err != nil || t != json.Delim('[') {
		return nil, false
	}
	liste := []string{}
	for dec.More() {
		s, ok := leseText(dec)
		if !ok {
			return nil, false
		}
		liste = append(liste, s)
	}
	_, _ = dec.Token()
	return liste, true
}

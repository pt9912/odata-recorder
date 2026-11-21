package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// keyServiceID und keyID sind die beiden Schlüssel des Bodys von `POST /__admin/sessions` (SPEC-292 Nr. 5 c).
const (
	keyServiceID = "service_id"
	keyID        = "id"
)

// ungueltigerBody ist der Fehler eines Verstoßes gegen SPEC-292 Nr. 5 a bis d; der Text nennt keinen Wert.
func ungueltigerBody() error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: Body von POST /__admin/sessions ungültig"))
}

// leseAnlegen liest den Body von `POST /__admin/sessions` (SPEC-292 Nr. 5): ein Kontextfehler beim Lesen bleibt der
// Fehler; jeder andere Lesefehler und jeder Verstoß gegen Form, Schlüssel und Werte ist model.ErrUngueltig. serviceID ist
// nichtleer, id ist leer, wenn der Schlüssel fehlt oder leer ist.
func leseAnlegen(r io.Reader) (serviceID, id string, err error) {
	roh, err := io.ReadAll(r)
	if err != nil {
		if finde(err, istKontextfehler) {
			return "", "", err
		}
		return "", "", ungueltigerBody()
	}
	if !utf8.Valid(roh) {
		return "", "", ungueltigerBody()
	}
	werte, ok := leseObjekt(json.NewDecoder(bytes.NewReader(roh)))
	if !ok || werte[keyServiceID] == "" {
		return "", "", ungueltigerBody()
	}
	return werte[keyServiceID], werte[keyID], nil
}

// leseObjekt liest genau ein Objekt aus dem Strom und danach das Ende: die Schlüssel sind service_id und id, jeder höchstens
// einmal (dekodiert gezählt), jeder Wert eine Zeichenkette. Das Ergebnis führt nur die gelesenen Schlüssel. Nach der
// Schleife steht `}` oder ein Fehler des Decoders (er lässt in einem Objekt kein anderes Ende zu, und ein fehlendes Ende
// ist ein *json.SyntaxError): der erste Aufruf von Token liest dieses Ende, der zweite endet nur mit io.EOF, wenn nichts
// außer Leerraum folgt.
func leseObjekt(dec *json.Decoder) (map[string]string, bool) {
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, false
	}
	werte := map[string]string{}
	for dec.More() {
		k, err := dec.Token()
		schluessel, istText := k.(string)
		if err != nil || !istText || (schluessel != keyServiceID && schluessel != keyID) {
			return nil, false
		}
		v, err := dec.Token()
		text, istText := v.(string)
		if _, doppelt := werte[schluessel]; err != nil || !istText || doppelt {
			return nil, false
		}
		werte[schluessel] = text
	}
	_, _ = dec.Token()
	if _, err := dec.Token(); err != io.EOF {
		return nil, false
	}
	return werte, true
}

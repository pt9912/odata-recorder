package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Schlüssel der drei Bodys (SPEC-298 Nr. 5); `id` von `POST /__admin/workspaces` ist keyID.
const (
	keyBaselineSnapshotID = "baseline_snapshot_id"
	keyName               = "name"
	keySnapshotID         = "snapshot_id"
)

// ungueltigerWorkspaceBody ist der Fehler eines Verstoßes gegen SPEC-298 Nr. 5 a bis d; der Text nennt keinen Wert.
func ungueltigerWorkspaceBody() error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: Body der Workspace-Route ungültig"))
}

// leseAnlegenWorkspace liest den Body von `POST /__admin/workspaces`: baseline_snapshot_id ist Pflicht, id ist leer, wenn
// der Schlüssel fehlt.
func leseAnlegenWorkspace(r io.Reader) (baselineID, id string, err error) {
	werte, err := leseWerte(r, []string{keyBaselineSnapshotID, keyID}, keyBaselineSnapshotID)
	if err != nil {
		return "", "", err
	}
	return werte[keyBaselineSnapshotID], werte[keyID], nil
}

// leseSnapshotName liest den Body von `POST …/snapshots`: name ist Pflicht.
func leseSnapshotName(r io.Reader) (string, error) {
	werte, err := leseWerte(r, []string{keyName}, keyName)
	return werte[keyName], err
}

// leseRestore liest den Body von `POST …/restore`: snapshot_id ist Pflicht.
func leseRestore(r io.Reader) (string, error) {
	werte, err := leseWerte(r, []string{keySnapshotID}, keySnapshotID)
	return werte[keySnapshotID], err
}

// leseWerte liest den Body bis zum Ende (SPEC-298 Nr. 5): ein Kontextfehler beim Lesen bleibt der Fehler, jeder andere
// Lesefehler und jeder Verstoß gegen Form, Schlüssel und Pflicht ist model.ErrUngueltig. Die Werte stehen dekodiert und
// ungeprüft im Ergebnis, das nur die gelesenen Schlüssel führt.
func leseWerte(r io.Reader, erlaubt []string, pflicht string) (map[string]string, error) {
	roh, err := io.ReadAll(r)
	if err != nil {
		if finde(err, istKontextfehler) {
			return nil, err
		}
		return nil, ungueltigerWorkspaceBody()
	}
	if !utf8.Valid(roh) {
		return nil, ungueltigerWorkspaceBody()
	}
	werte, ok := leseObjektMit(json.NewDecoder(bytes.NewReader(roh)), erlaubt)
	if _, da := werte[pflicht]; !ok || !da {
		return nil, ungueltigerWorkspaceBody()
	}
	return werte, nil
}

// leseObjektMit liest genau ein Objekt aus dem Strom und danach das Ende: erlaubt sind allein die Schlüssel in erlaubt, jeder
// höchstens einmal (dekodiert gezählt), jeder Wert eine Zeichenkette. Nach der Schleife steht `}` oder ein Fehler des
// Decoders (er lässt in einem Objekt kein anderes Ende zu): der erste Aufruf von Token liest dieses Ende, der zweite endet
// nur mit io.EOF, wenn nichts außer Leerraum folgt.
func leseObjektMit(dec *json.Decoder, erlaubt []string) (map[string]string, bool) {
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, false
	}
	werte := map[string]string{}
	for dec.More() {
		k, err := dec.Token()
		schluessel, istText := k.(string)
		if err != nil || !istText || !slices.Contains(erlaubt, schluessel) {
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

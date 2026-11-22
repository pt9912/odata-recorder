package admin

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Schlüssel der Query (SPEC-293 Nr. 3): limit und after bei `exchanges`, baseline_id bei `coverage`.
const (
	inspKeyLimit      = "limit"
	inspKeyAfter      = "after"
	inspKeyBaselineID = "baseline_id"
)

// inspQueryUngueltig ist der Fehler eines Verstoßes gegen SPEC-293 Nr. 3 bis 5; der Text nennt keinen Wert.
func inspQueryUngueltig() error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: Query der Inspektion ungültig"))
}

// leseQuery zerlegt die ganze Query mit url.ParseQuery (SPEC-293 Nr. 3): eine ungültige Prozentfolge und ein Semikolon
// als Trenner sind ein Verstoß, auch wenn nur ein Schlüssel betroffen ist, den die Route nicht liest. Das Ergebnis führt
// den dekodierten Wert jedes bekannten Schlüssels, der vorkommt; ein bekannter Schlüssel, der öfter als einmal oder mit
// leerem Wert vorkommt, ist ein Verstoß. Jeder unbekannte Schlüssel ist ohne Wirkung.
func leseQuery(roh string, bekannt ...string) (map[string]string, error) {
	alle, err := url.ParseQuery(roh)
	if err != nil {
		return nil, inspQueryUngueltig()
	}
	werte := make(map[string]string, len(bekannt))
	for _, k := range bekannt {
		vs, da := alle[k]
		if !da {
			continue
		}
		if len(vs) != 1 || vs[0] == "" {
			return nil, inspQueryUngueltig()
		}
		werte[k] = vs[0]
	}
	return werte, nil
}

// leseZahl liest eine Zahl nach SPEC-293 Nr. 4: eine oder mehr ASCII-Ziffern und nichts sonst, im Bereich von int (kein
// Kappen); führende Nullen sind erlaubt. Atoi weist den leeren Text ab, leseQuery liefert ihn nie.
func leseZahl(text string) (int, error) {
	for i := range len(text) {
		if text[i] < '0' || text[i] > '9' {
			return 0, inspQueryUngueltig()
		}
	}
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0, inspQueryUngueltig()
	}
	return n, nil
}

// leseSeite liest limit und after von `exchanges` (SPEC-293 Nr. 3, 4); ein fehlender Schlüssel ist 0.
func leseSeite(roh string) (inbound.Seite, error) {
	werte, err := leseQuery(roh, inspKeyLimit, inspKeyAfter)
	if err != nil {
		return inbound.Seite{}, err
	}
	var seite inbound.Seite
	if t, da := werte[inspKeyLimit]; da {
		if seite.Limit, err = leseZahl(t); err != nil {
			return inbound.Seite{}, err
		}
	}
	if t, da := werte[inspKeyAfter]; da {
		if seite.Nach, err = leseZahl(t); err != nil {
			return inbound.Seite{}, err
		}
	}
	return seite, nil
}

// leseBaselineID liest baseline_id von `coverage` (SPEC-293 Nr. 3, 5): Pflicht, höchstens einmal und nichtleer; jeder
// andere dekodierte Wert gilt byte-gleich.
func leseBaselineID(roh string) (string, error) {
	werte, err := leseQuery(roh, inspKeyBaselineID)
	if err != nil {
		return "", err
	}
	id, da := werte[inspKeyBaselineID]
	if !da {
		return "", inspQueryUngueltig()
	}
	return id, nil
}

package admin

import (
	"encoding/json"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// zeitLayout ist RFC 3339 mit genau neun Nachkommastellen; das `Z` der UTC-Zeit hängt zeitJSON an (SPEC-292 Nr. 7 c).
const zeitLayout = "2006-01-02T15:04:05.000000000"

// zeitJSON ist eine Zeit der Antwort: RFC 3339 in UTC mit neun Nachkommastellen, die Nullzeit ist `null`
// (SPEC-292 Nr. 7 c).
type zeitJSON time.Time

// MarshalJSON bildet die Zeit ab: ein anderer Ort wird nach UTC umgerechnet, der Zeitpunkt bleibt.
func (z zeitJSON) MarshalJSON() ([]byte, error) {
	t := time.Time(z)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.UTC().Format(zeitLayout) + `Z"`), nil
}

// sessionJSON ist die Session der Antwort: die vier Felder von model.Session in der Folge des Vertrags; der Status steht
// als Text unverändert (SPEC-292 Nr. 7 b).
type sessionJSON struct {
	ID        string   `json:"id"`
	ServiceID string   `json:"service_id"`
	CreatedAt zeitJSON `json:"created_at"`
	Status    string   `json:"status"`
}

// sessionListeJSON ist der Body von `GET /__admin/sessions`; die Liste ist nie `null` (SPEC-292 Nr. 7 d).
type sessionListeJSON struct {
	Sessions []sessionJSON `json:"sessions"`
}

// sessionAus bildet die Session des Dienstes ab, ohne ein Feld zu prüfen oder zu ändern.
func sessionAus(s model.Session) sessionJSON {
	return sessionJSON{ID: s.ID, ServiceID: s.ServiceID, CreatedAt: zeitJSON(s.CreatedAt), Status: string(s.Status)}
}

// schreibeJSON liefert den JSON-Body ohne Leerraum und ohne Zeilenende. Das Marshalling einer Struktur aus
// Zeichenketten und Zeiten scheitert nicht; ungültiges UTF-8 einer Zeichenkette wird U+FFFD (SPEC-292 Nr. 7 a).
func schreibeJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

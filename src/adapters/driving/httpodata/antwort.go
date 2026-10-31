package httpodata

import (
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// kopierPuffer ist die Größe eines Stücks beim Kopieren des Bodys.
const kopierPuffer = 32 * 1024

// hopByHop sind die Namen, die der Adapter nicht weitergibt (SPEC-136 Nr. 5c), klein geschrieben.
func hopByHop() []string {
	return []string{
		"connection", "proxy-connection", "keep-alive", "proxy-authenticate",
		"proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade",
	}
}

// schreibeAntwort liefert die Antwort des Dienstes aus (SPEC-136 Nr. 5, 6): ein Bruch des
// Vertrags ist ODR-E9001, sonst gehen Status, Header und Bytes unverändert aus; Close läuft genau
// einmal, auch beim Abbruch, und vor der Rückkehr des Handlers.
func schreibeAntwort(w http.ResponseWriter, r *http.Request, id string, a inbound.Antwort) {
	if a.Body != nil {
		defer func() { _ = a.Body.Close() }()
	}
	if a.Body == nil || a.Status < 200 || a.Status > 599 {
		schreibeFehlerobjekt(w, r, id, model.CodeE9001, http.StatusInternalServerError)
		return
	}
	h := w.Header()
	setzeHeader(h, a.Header)
	if _, da := h["Content-Type"]; !da {
		h["Content-Type"] = nil
	}
	_, mitLaenge := h["Content-Length"]
	// Der Request-Body bleibt bis Close lesbar (SPEC-136 Nr. 3, 4h); ohne Vollduplex verwirft
	// net/http beim ersten Schreiben den ungelesenen Rest. HTTP/1 kennt den Fehlerfall nicht.
	_ = http.NewResponseController(w).EnableFullDuplex()
	w.WriteHeader(a.Status)
	kopiere(w, a.Body, !mitLaenge)
}

// setzeHeader übernimmt alle Header außer den Hop-by-Hop-Namen und den Namen, die ein Wert von
// Connection nennt; die Namen laufen sortiert, damit Namen gleicher Schreibweise stabil münden.
func setzeHeader(ziel http.Header, quelle map[string][]string) {
	entfallen := hopByHop()
	for name, werte := range quelle {
		if strings.EqualFold(name, "Connection") {
			for _, wert := range werte {
				for _, teil := range strings.Split(wert, ",") {
					entfallen = append(entfallen, strings.ToLower(strings.Trim(teil, " \t")))
				}
			}
		}
	}
	namen := make([]string, 0, len(quelle))
	for name := range quelle {
		namen = append(namen, name)
	}
	slices.Sort(namen)
	for _, name := range namen {
		if slices.Contains(entfallen, strings.ToLower(name)) {
			continue
		}
		for _, wert := range quelle[name] {
			ziel.Add(name, wert)
		}
	}
}

// kopiere schreibt jedes gelesene Stück, bevor das nächste gelesen wird; mit leeren trägt jedes
// Stück bis zum Client durch (Flush). Ein Schreibfehler beendet das Kopieren, ein Lesefehler
// bricht die Verbindung ab, nachdem die bereits gelesenen Bytes ausgegangen sind; vor dem ersten
// Byte geht auch der Kopf nicht aus.
func kopiere(w http.ResponseWriter, body io.Reader, leeren bool) {
	rc := http.NewResponseController(w)
	puffer := make([]byte, kopierPuffer)
	geschrieben := false
	for {
		n, err := body.Read(puffer)
		if n > 0 {
			if _, werr := w.Write(puffer[:n]); werr != nil {
				return
			}
			geschrieben = true
			if leeren && rc.Flush() != nil {
				return
			}
		}
		switch err {
		case nil:
		case io.EOF:
			return
		default:
			if geschrieben {
				_ = rc.Flush()
			}
			brichAb()
		}
	}
}

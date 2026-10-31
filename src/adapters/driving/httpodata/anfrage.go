package httpodata

import (
	"net/http"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// anfrageAus bildet die Anfrage unverändert aus dem Request-Target und dem Body-Strom
// (SPEC-136 Nr. 4): Pfad und RawQuery aus RequestURI, nicht aus r.URL; Header ist r.Header.
func anfrageAus(r *http.Request, id string) inbound.Anfrage {
	pfad, query, _ := strings.Cut(r.RequestURI, "?")
	a := inbound.Anfrage{
		Methode:        r.Method,
		Pfad:           pfad,
		RawQuery:       query,
		KorrelationsID: id,
		Header:         r.Header,
	}
	if r.Body != http.NoBody {
		a.Body = r.Body
		a.Laenge = r.ContentLength
	}
	return a
}

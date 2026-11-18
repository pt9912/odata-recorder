package hybrid

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// waehle liefert Ziel und OData-Pfad des Dienstes, dessen Root gleich pfad ist oder ihm an einer Segmentgrenze
// vorausgeht, byteweise und bei mehreren den mit dem längsten Root (SPEC-284 Nr. 4); ohne Dienst sind beide leer.
func (v *Verteiler) waehle(pfad string) (ziel, odataPfad string) {
	gewaehlt := -1
	rest := ""
	for i, d := range v.dienste {
		r, ok := unterRoot(d.Root, pfad)
		if ok && (gewaehlt < 0 || len(d.Root) > len(v.dienste[gewaehlt].Root)) {
			gewaehlt, rest = i, r
		}
	}
	if gewaehlt < 0 {
		return "", ""
	}
	if rest == "" {
		rest = "/"
	}
	return v.dienste[gewaehlt].UpstreamURL, rest
}

// unterRoot liefert den Rest von pfad hinter dem Root und ob der Root pfad gleicht oder an einer Segmentgrenze
// vorausgeht; der Root / gilt für jeden Pfad, der mit / beginnt, und lässt den Pfad unverändert.
func unterRoot(root, pfad string) (string, bool) {
	switch {
	case root == "/":
		return pfad, strings.HasPrefix(pfad, "/")
	case pfad == root:
		return "", true
	case strings.HasPrefix(pfad, root+"/"):
		return pfad[len(root):], true
	}
	return "", false
}

// upstreamFallback sendet über U und liefert die Antwort mit der Kennzeichnung upstream oder den abgebildeten Fehler
// (SPEC-284 Nr. 8 bis 11); davor prüft er ctx (Nr. 3 e). Es gibt kein Wiederholen und keinen Rückfall (Nr. 8).
func (v *Verteiler) upstreamFallback(ctx context.Context, a, u inbound.Anfrage, ziel, odataPfad string) (inbound.Antwort, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return inbound.Antwort{}, ctxErr
	}
	antwort, err := v.upstream.Senden(ctx, outbound.UpstreamAnfrage{
		Ziel:     ziel,
		Methode:  a.Methode,
		Pfad:     odataPfad,
		RawQuery: a.RawQuery,
		Header:   upstreamHeader(a.Header),
		Body:     u.Body,
		Laenge:   u.Laenge,
	})
	if err != nil {
		verwirf(inbound.Antwort{Body: antwort.Body})
		return inbound.Antwort{}, upstreamFehler(err)
	}
	return gekennzeichnet(inbound.Antwort{Status: antwort.Status, Header: ohneHopByHop(antwort.Header), Body: antwort.Body}, wertUpstream), nil
}

// upstreamFehler bildet den Fehler von Senden (SPEC-284 Nr. 11): ein Kontextfehler bleibt ohne Code, ErrUpstreamZeit
// ist ODR-E4001 mit 504, ErrUpstreamNichtErreichbar ODR-E4001 mit 502, jeder andere Fehler ODR-E9001 mit 500.
func upstreamFehler(err error) error {
	if kontextFehler(err) {
		return err
	}
	switch {
	case errors.Is(err, outbound.ErrUpstreamZeit):
		return &inbound.Fehler{Code: model.CodeE4001, Status: 504, Ursache: err}
	case errors.Is(err, outbound.ErrUpstreamNichtErreichbar):
		return &inbound.Fehler{Code: model.CodeE4001, Status: 502, Ursache: err}
	}
	return &inbound.Fehler{Code: model.CodeE9001, Status: 500, Ursache: err}
}

// hopByHop sagt, ob klein (ein Name in ASCII-Kleinbuchstaben) zur festen Liste der Hop-by-Hop-Header gehört
// (SPEC-284 Nr. 8 c).
func hopByHop(klein string) bool {
	return slices.Contains([]string{
		"connection", "proxy-connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailer", "transfer-encoding", "upgrade",
	}, klein)
}

// ohneHopByHop liefert eine Kopie von h (neue Map, neue Wertlisten) ohne die Hop-by-Hop-Header und ohne jeden Namen, den
// ein Wert von Connection nennt: an Kommas geteilt, Leerzeichen und Tabulatoren am Rand entfernt, leere Teile ohne
// Wirkung (SPEC-284 Nr. 8 c, 9).
func ohneHopByHop(h map[string][]string) map[string][]string {
	return gefiltert(h, func(string) bool { return false })
}

// gefiltert kopiert h ohne die Hop-by-Hop-Header, ohne die Namen aus Connection und ohne jeden Namen, für den extra
// wahr ist; extra erhält den Namen in ASCII-Kleinbuchstaben.
func gefiltert(h map[string][]string, extra func(klein string) bool) map[string][]string {
	genannt := map[string]bool{}
	for n, werte := range h {
		if asciiKlein(n) != "connection" {
			continue
		}
		for _, w := range werte {
			for _, t := range strings.Split(w, ",") {
				if t = strings.Trim(t, " \t"); t != "" {
					genannt[asciiKlein(t)] = true
				}
			}
		}
	}
	out := make(map[string][]string, len(h))
	for n, werte := range h {
		if k := asciiKlein(n); !hopByHop(k) && !genannt[k] && !extra(k) {
			out[n] = slices.Clone(werte)
		}
	}
	return out
}

// upstreamHeader liefert die Header zum Upstream (SPEC-284 Nr. 8 c): eine Kopie der Header der Anfrage ohne
// Hop-by-Hop-Header, ohne jeden Namen mit dem Präfix X-Recorder-, ohne Expect, Host und Content-Length und ohne die drei
// Header zum Überschreiben der Methode; Accept-Encoding geht unverändert.
func upstreamHeader(h map[string][]string) map[string][]string {
	return gefiltert(h, func(klein string) bool {
		switch klein {
		case "expect", "host", "content-length", "x-http-method", "x-http-method-override", "x-method-override":
			return true
		}
		return strings.HasPrefix(klein, "x-recorder-")
	})
}

// Package httpupstream ist der Driven Adapter des Outbound Ports UpstreamClient (ARC-008): er sendet
// einen Request mit net/http der Standardbibliothek an den Upstream eines Service und liefert dessen
// Antwort als Strom (SPEC-120).
package httpupstream

import (
	"context"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.UpstreamClient = (*Client)(nil)

// Konfig ist die Konfiguration des Clients (SPEC-120 Nr. 1): Frist bis zum Kopf der Antwort,
// Zertifikatsprüfung und zusätzliche Vertrauensanker als PEM.
type Konfig struct {
	Frist   time.Duration
	Pruefen bool
	CAPEM   []byte
}

// Client setzt UpstreamClient um. Senden hält keinen Zustand am Client; gleichzeitige Aufrufe teilen
// nur die Verbindungen des Transports (SPEC-120 Nr. 9).
type Client struct {
	frist time.Duration
	// transport ist im Betrieb der *http.Transport aus New. Test-Haken: export_test.go setzt einen
	// Fake, der die Folge gleichzeitiger Ursachen bestimmt.
	transport http.RoundTripper
	// neueFrist startet den Timer der Frist und liefert seine Stop-Funktion; nil heißt
	// time.AfterFunc. Test-Haken: export_test.go setzt einen Timer, den der Test auslöst.
	neueFrist func(d time.Duration, f func()) (stop func() bool)
}

// New liefert einen Client; eine Frist unter 1 ns ist model.ErrUngueltig (SPEC-120 Nr. 1). Der
// Transport nutzt keinen Proxy und fordert keine Kompression an.
func New(k Konfig) (*Client, error) {
	if k.Frist < time.Nanosecond {
		return nil, ungueltig("Frist")
	}
	tr := &http.Transport{
		Proxy:              nil,
		DisableCompression: true,
	}
	return &Client{frist: k.Frist, transport: tr}, nil
}

// Senden prüft die Anfrage (SPEC-120 Nr. 2), sendet sie (Nr. 3, 4) und kehrt mit dem Kopf der
// Antwort zurück (Nr. 5). Der Kontext des Austauschs endet mit ctx, mit dem Ablauf der Frist, mit
// einem Fehler von Senden und mit dem Schließen des Antwort-Bodys; mit ihm endet das Lesen aus
// dem Request-Body. Die Rangfolge der Fehler steht in ergebnis (Nr. 8).
func (c *Client) Senden(ctx context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	z, err := pruefe(ctx, a)
	if err != nil {
		return outbound.UpstreamAntwort{}, err
	}
	if err := ctx.Err(); err != nil {
		return outbound.UpstreamAntwort{}, kontextFehler(err)
	}
	xctx, ende := context.WithCancel(ctx)
	stop := c.starteFrist(ende)
	var l *leser
	if a.Laenge != 0 {
		l = neuerLeser(xctx.Done(), a.Body)
	}
	res, err := c.transport.RoundTrip(baueRequest(xctx, z, a, l))
	gestoppt := stop()
	return ergebnis(ctx, res, err, gestoppt, l, ende)
}

// ziel ist ein geprüftes Ziel: Schema, Host samt Port und Pfad, je wie in Ziel geschrieben.
type ziel struct {
	schema, host, pfad string
}

// pruefe weist ungültige Argumente ab, bevor eine Verbindung entsteht oder Body gelesen wird
// (SPEC-120 Nr. 2); der Fehler nennt das Feld, nie seinen Wert.
func pruefe(ctx context.Context, a outbound.UpstreamAnfrage) (ziel, error) {
	z, zielOK := leseZiel(a.Ziel)
	switch {
	case ctx == nil:
		return ziel{}, ungueltig("ctx")
	case !zielOK:
		return ziel{}, ungueltig("Ziel")
	case !istToken(a.Methode) || a.Methode == http.MethodConnect:
		return ziel{}, ungueltig("Methode")
	case !strings.HasPrefix(a.Pfad, "/") || !ohneTrenner(a.Pfad, "?#"):
		return ziel{}, ungueltig("Pfad")
	case !ohneTrenner(a.RawQuery, "#"):
		return ziel{}, ungueltig("RawQuery")
	case !headerGueltig(a.Header):
		return ziel{}, ungueltig("Header")
	case a.Laenge < -1:
		return ziel{}, ungueltig("Laenge")
	case a.Body == nil && a.Laenge != 0:
		return ziel{}, ungueltig("Body")
	}
	return z, nil
}

// leseZiel nimmt eine absolute http- oder https-URL mit Host, ohne Zugangsdaten, Query und Fragment,
// mit einem Port von 1 bis 65535, falls einer steht, und ohne Byte unter 0x21 oder 0x7F im Pfad.
// Der Pfad wird aus dem Text geschnitten, nicht aus url.URL: dort steht er unescaped oder neu
// escaped.
func leseZiel(s string) (ziel, bool) {
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" ||
		u.User != nil || strings.ContainsAny(s, "?#") || !portGueltig(u) {
		return ziel{}, false
	}
	_, nachSchema, _ := strings.Cut(s, "://")
	pfad := ""
	if i := strings.IndexByte(nachSchema, '/'); i >= 0 {
		pfad = nachSchema[i:]
	}
	if !ohneTrenner(pfad, "") {
		return ziel{}, false
	}
	return ziel{schema: u.Scheme, host: u.Host, pfad: pfad}, true
}

// portGueltig verlangt für einen genannten Port 1 bis 65535; `host:` ohne Ziffern ist ungültig.
func portGueltig(u *url.URL) bool {
	if strings.HasSuffix(u.Host, ":") {
		return false
	}
	port := u.Port()
	if port == "" {
		return true
	}
	n, err := strconv.Atoi(port)
	return err == nil && n >= 1 && n <= 65535
}

// ohneTrenner meldet, ob s weder ein Byte unter 0x21 noch 0x7F noch eines aus verboten enthält.
func ohneTrenner(s, verboten string) bool {
	for i := range len(s) {
		if s[i] < 0x21 || s[i] == 0x7F || strings.IndexByte(verboten, s[i]) >= 0 {
			return false
		}
	}
	return true
}

// istToken meldet, ob s ein HTTP-Token ist (RFC 9110 §5.6.2); der leere Text ist keiner.
func istToken(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		b := s[i]
		alnum := b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
		if !alnum && strings.IndexByte("!#$%&'*+-.^_`|~", b) < 0 {
			return false
		}
	}
	return true
}

// headerGueltig verlangt Token als Namen und Werte ohne Steuerzeichen außer dem Tabulator.
func headerGueltig(h map[string][]string) bool {
	for name, werte := range h {
		if !istToken(name) {
			return false
		}
		for _, w := range werte {
			for i := range len(w) {
				if (w[i] < 0x20 && w[i] != '\t') || w[i] == 0x7F {
					return false
				}
			}
		}
	}
	return true
}

// baueRequest setzt die Request-Zeile und die Header nach SPEC-120 Nr. 3 und die Rahmung nach Nr. 4.
// Das Ziel der Request-Zeile steht in URL.Opaque, byte-gleich; beginnt es mit `//`, steht es in
// absoluter Form, denn RequestURI setzt nur das Schema davor.
func baueRequest(ctx context.Context, z ziel, a outbound.UpstreamAnfrage, l *leser) *http.Request {
	opaque := z.pfad + a.Pfad
	if strings.HasPrefix(opaque, "//") {
		opaque = "//" + z.host + opaque
	}
	req := &http.Request{
		Method:     a.Methode,
		URL:        &url.URL{Scheme: z.schema, Host: z.host, Opaque: opaque, RawQuery: a.RawQuery},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     baueHeader(a.Header),
		Host:       z.host,
	}
	rahme(req, a.Laenge, l)
	return req.WithContext(ctx)
}

// ausgenommen sind die Namen, die der Client nicht aus Header sendet (SPEC-120 Nr. 3); verglichen
// ohne Beachtung der ASCII-Schreibweise.
func ausgenommen(name string) bool {
	for _, a := range [...]string{"Host", "Content-Length", "Transfer-Encoding", "Trailer"} {
		if strings.EqualFold(name, a) {
			return true
		}
	}
	return false
}

// baueHeader übernimmt jeden Namen in seiner Schreibweise mit seinen Werten, außer den
// ausgenommenen und den Namen mit leerer Wertliste. User-Agent steht immer unter dem kanonischen
// Namen: mit einem Wert oder mit leerer Liste, dann sendet die Bibliothek keinen.
func baueHeader(h map[string][]string) http.Header {
	out := make(http.Header, len(h)+1)
	var uaNamen []string
	for name, werte := range h {
		switch {
		case ausgenommen(name):
		case strings.EqualFold(name, "User-Agent"):
			uaNamen = append(uaNamen, name)
		case len(werte) > 0:
			out[name] = werte
		}
	}
	out["User-Agent"] = userAgent(h, uaNamen)
	return out
}

// userAgent nimmt unter den Schreibweisen von User-Agent in der Bytefolge der Namen die erste mit
// nichtleerer Wertliste und davon den ersten Wert ohne Leerzeichen und Tabulatoren am Rand; ist er
// leer oder fehlt er, ist die Liste leer.
func userAgent(h map[string][]string, namen []string) []string {
	slices.Sort(namen)
	for _, n := range namen {
		if werte := h[n]; len(werte) > 0 {
			if w := strings.Trim(werte[0], " \t"); w != "" {
				return []string{w}
			}
			return []string{}
		}
	}
	return []string{}
}

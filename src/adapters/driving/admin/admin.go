package admin

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// healthzBody ist der Body der eingebauten Route `GET /healthz` (SPEC-238 Nr. 3 c).
const healthzBody = `{"status":"ok"}`

// healthzPfad ist der Pfad der Route, die ohne Authentisierung bedient wird (SPEC-238 Nr. 3 c).
const healthzPfad = "/healthz"

// schemaKlein ist das Schema der Authentisierung in Kleinbuchstaben; die ASCII-Schreibweise im Header ist
// gleichgültig (SPEC-238 Nr. 4 b).
const schemaKlein = "bearer"

// Handler bearbeitet einen Request einer Route (SPEC-238 Nr. 2).
type Handler func(ctx context.Context, a Anfrage) (Antwort, error)

// Route ordnet Methode und Muster einem Handler zu (SPEC-238 Nr. 2 a bis d).
type Route struct {
	Methode, Muster string
	Handler         Handler
}

// Anfrage ist, was eine Route vom Request erhält: das Token erreicht sie nie (SPEC-238 Nr. 2).
type Anfrage struct {
	Methode, Pfad, RawQuery, KorrelationsID string
	Pfadwerte                               map[string]string
	Header                                  map[string][]string
	Body                                    io.Reader
	Laenge                                  int64
}

// Antwort ist die Erfolgsantwort einer Route (SPEC-238 Nr. 2).
type Antwort struct {
	Status int
	Body   []byte
}

// Konfig trägt die Eingaben von New (SPEC-238 Nr. 1).
type Konfig struct {
	Listener   net.Listener
	Token      []byte
	IDs        support.IDGenerator
	Audit      support.AuditLog
	Telemetrie support.Telemetry
	Routen     []Route
	// Vergleich vergleicht das erwartete Token mit dem Kandidaten; nil heißt vergleiche (SPEC-238 Nr. 1, 4 e).
	Vergleich func(erwartet, gegeben []byte) bool
}

// Server authentisiert und ordnet Requests den Routen zu (SPEC-238); außer dem Zustand der Zulassung (SPEC-238 Nr. 11)
// ist sein Zustand nach New schreibgeschützt.
type Server struct {
	listener   net.Listener
	token      []byte
	ids        support.IDGenerator
	audit      support.AuditLog
	telemetrie support.Telemetry
	routen     []route
	vergleich  func(erwartet, gegeben []byte) bool
	http       *http.Server
	abbruch    context.CancelFunc

	// mu schützt die folgenden Felder; die Zulassung prüft angehalten und zählt unter einer Sperre ohne Freigabe
	// dazwischen, und keine Sperre liegt über einem Aufruf von Handler oder Vergleich (SPEC-238 Nr. 12).
	mu            sync.Mutex
	angehalten    bool
	durchAnhalten bool
	acceptFehler  error
	geschlossen   bool
	laufend       int
	leer          chan struct{}
	leerZu        bool
	verb          map[net.Conn]*verbindung
}

// vergleiche liefert, ob beide Werte gleich sind: ConstantTimeCompare auf den SHA-256-Summen. Der Vergleich der
// Summen hängt weder vom Inhalt noch von der Länge ab; die Länge des Gegebenen wirkt nur auf das Hashen und ist durch
// die Header-Grenze des bedienenden Servers begrenzt (SPEC-238 Nr. 4 e).
func vergleiche(erwartet, gegeben []byte) bool {
	a, b := sha256.Sum256(erwartet), sha256.Sum256(gegeben)
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

// ungueltig bildet den Fehler von New: model.ErrUngueltig mit einem festen Text, der keinen Wert nennt.
func ungueltig(was string) error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: "+was))
}

// healthz ist der Handler der eingebauten Route `GET /healthz` (SPEC-238 Nr. 3 c).
func healthz(context.Context, Anfrage) (Antwort, error) {
	return Antwort{Status: http.StatusOK, Body: []byte(healthzBody)}, nil
}

// baueRouten prüft die Routen nach SPEC-238 Nr. 2 a bis d und liefert sie zerlegt, die eingebaute Route zuletzt.
func baueRouten(routen []Route) ([]route, error) {
	eingebaut := route{methode: http.MethodGet, segs: []segment{{literal: "healthz"}}, handler: healthz}
	out := make([]route, 0, len(routen)+1)
	for _, r := range routen {
		segs, ok := parseMuster(r.Muster)
		switch {
		case !gueltigeMethode(r.Methode):
			return nil, ungueltig("Route mit ungültiger Methode")
		case !ok:
			return nil, ungueltig("Route mit ungültigem Muster")
		case r.Handler == nil:
			return nil, ungueltig("Route ohne Handler")
		}
		neu := route{methode: r.Methode, segs: segs, handler: r.Handler}
		if ueberlappt(neu, eingebaut) || slices.ContainsFunc(out, func(v route) bool { return ueberlappt(neu, v) }) {
			return nil, ungueltig("Routen überlappen")
		}
		out = append(out, neu)
	}
	return append(out, eingebaut), nil
}

// New prüft die Konfiguration (SPEC-238 Nr. 1) und kopiert Token und Routen. New ruft keinen Port, keine Methode
// des Listeners, öffnet keinen Socket und startet keine Goroutine.
func New(k Konfig) (*Server, error) {
	switch {
	case k.Listener == nil:
		return nil, ungueltig("Listener fehlt")
	case k.IDs == nil:
		return nil, ungueltig("IDs fehlt")
	case k.Audit == nil:
		return nil, ungueltig("Audit fehlt")
	case k.Telemetrie == nil:
		return nil, ungueltig("Telemetrie fehlt")
	case len(k.Token) < tokenMinBytes:
		return nil, ungueltig("Token zu kurz")
	}
	routen, err := baueRouten(k.Routen)
	if err != nil {
		return nil, err
	}
	vergleich := k.Vergleich
	if vergleich == nil {
		vergleich = vergleiche
	}
	basis, abbruch := context.WithCancel(context.Background())
	s := &Server{
		listener: k.Listener, token: slices.Clone(k.Token), ids: k.IDs, audit: k.Audit, telemetrie: k.Telemetrie,
		routen: routen, vergleich: vergleich, abbruch: abbruch,
		leer: make(chan struct{}), verb: map[net.Conn]*verbindung{},
	}
	s.http = s.baueHTTP(basis)
	return s, nil
}

// ServeHTTP vergibt die Korrelations-ID vor allem anderen und weist auf einem angehaltenen Server jeden Request mit 503
// ab; sonst bedient es `GET /healthz` ohne Token, authentisiert jeden anderen Request vor der Zuordnung und ruft dann
// den Handler der einen Route oder antwortet 404 oder 405 mit ODR-E2102 (SPEC-238 Nr. 3, 4, 5, 11).
// `EnableFullDuplex` erlaubt dem Handler, Lesen und Schreiben zu mischen; kein Test fängt das Entfernen (N-13).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = http.NewResponseController(w).EnableFullDuplex()
	s.markiereHandler(r)
	id := s.ids.KorrelationsID()
	w.Header().Set("X-Correlation-Id", id)
	if !s.zulassen() {
		abweisen(w)
		return
	}
	defer s.verlasse()
	s.bediene(w, r, id)
}

// bediene ist der Ablauf eines zugelassenen Requests ab Nr. 3 c.
func (s *Server) bediene(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method == http.MethodGet && r.URL.Path == healthzPfad {
		antwort, _ := healthz(r.Context(), Anfrage{})
		schreibeErfolg(w, antwort)
		return
	}
	if !s.authentisiert(r.Header.Values("Authorization")) {
		schreibeFehler(w, model.CodeE2101, http.StatusUnauthorized, "", "Bearer")
		return
	}
	t := ordne(s.routen, r.Method, r.URL.EscapedPath())
	switch {
	case t.route != nil:
		rufe(w, r, id, t)
	case len(t.allow) == 0:
		schreibeFehler(w, model.CodeE2102, http.StatusNotFound, "", "")
	default:
		schreibeFehler(w, model.CodeE2102, http.StatusMethodNotAllowed, strings.Join(t.allow, ", "), "")
	}
}

// authentisiert prüft die Zeilen des Headers Authorization (SPEC-238 Nr. 4 a bis c): genau eine Zeile mit dem Schema
// und einem nichtleeren Kandidaten; erst dann ruft sie Vergleich genau einmal, und allein sein Ergebnis gilt.
func (s *Server) authentisiert(zeilen []string) bool {
	if len(zeilen) != 1 {
		return false
	}
	rest, ok := hinterSchema(zeilen[0])
	if !ok || rest == "" {
		return false
	}
	return s.vergleich(s.token, []byte(rest))
}

// hinterSchema liefert den Rest des Wertes hinter dem Schema und den führenden Leerzeichen, Byte für Byte; ok ist
// falsch, wenn der Wert nicht mit dem Schema und mindestens einem Leerzeichen (0x20) beginnt (SPEC-238 Nr. 4 b).
func hinterSchema(wert string) (rest string, ok bool) {
	if len(wert) <= len(schemaKlein) || wert[len(schemaKlein)] != ' ' {
		return "", false
	}
	for i := 0; i < len(schemaKlein); i++ {
		c := wert[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != schemaKlein[i] {
			return "", false
		}
	}
	return strings.TrimLeft(wert[len(schemaKlein):], " "), true
}

// anfrage bildet die Anfrage der Route: der Header ist eine Kopie ohne Authorization (SPEC-238 Nr. 2).
func anfrage(r *http.Request, id string, werte map[string]string) Anfrage {
	header := make(map[string][]string, len(r.Header))
	for k, v := range r.Header {
		if !strings.EqualFold(k, "Authorization") {
			header[k] = slices.Clone(v)
		}
	}
	return Anfrage{
		Methode: r.Method, Pfad: r.URL.Path, RawQuery: r.URL.RawQuery, KorrelationsID: id,
		Pfadwerte: werte, Header: header, Body: r.Body, Laenge: r.ContentLength,
	}
}

// rufe ruft den Handler der Route genau einmal; Status und Body gehen durch, ein Fehler ist ODR-E9001 mit 500. Der Body
// ist auf 1 MiB begrenzt: das 1048577. Byte ergibt *http.MaxBytesError (SPEC-238 Nr. 10).
func rufe(w http.ResponseWriter, r *http.Request, id string, t treffer) {
	r.Body = http.MaxBytesReader(w, r.Body, bodyBytes)
	antwort, err := t.route.handler(r.Context(), anfrage(r, id, t.werte))
	if err != nil {
		schreibeFehler(w, model.CodeE9001, http.StatusInternalServerError, "", "")
		return
	}
	schreibeErfolg(w, antwort)
}

// schreibeErfolg schreibt die Erfolgsantwort einer Route: den Status der Route und, wenn der Body nicht leer ist,
// Content-Type, Content-Length und den Body (SPEC-238 Nr. 5).
func schreibeErfolg(w http.ResponseWriter, a Antwort) {
	if len(a.Body) > 0 {
		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Content-Length", strconv.Itoa(len(a.Body)))
	}
	w.WriteHeader(a.Status)
	if len(a.Body) > 0 {
		_, _ = w.Write(a.Body)
	}
}

// schreibeFehler ist die eine Stelle, die das Fehlerobjekt bildet und schreibt (SPEC-238 Nr. 5): drei Felder,
// Content-Type, Content-Length, bei 405 Allow (allow nichtleer) und bei 401 WWW-Authenticate (www nichtleer).
// Der Code ist eine Konstante der Tabelle, das Marshalling einer Struktur aus Zeichenketten scheitert nicht.
func schreibeFehler(w http.ResponseWriter, code model.Code, status int, allow, www string) {
	obj, _ := model.NeuesFehlerobjekt(code)
	body, _ := json.Marshal(obj)
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	if allow != "" {
		h.Set("Allow", allow)
	}
	if www != "" {
		h.Set("WWW-Authenticate", www)
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

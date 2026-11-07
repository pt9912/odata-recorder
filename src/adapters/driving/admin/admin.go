package admin

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// healthzBody ist der Body der eingebauten Route `GET /healthz` (SPEC-238 Nr. 3 c).
const healthzBody = `{"status":"ok"}`

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
}

// Server ordnet Requests den Routen zu (SPEC-238); sein Zustand ist nach New schreibgeschützt.
type Server struct {
	listener   net.Listener
	token      []byte
	ids        support.IDGenerator
	audit      support.AuditLog
	telemetrie support.Telemetry
	routen     []route
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
	return &Server{listener: k.Listener, token: slices.Clone(k.Token), ids: k.IDs, audit: k.Audit, telemetrie: k.Telemetrie, routen: routen}, nil
}

// ServeHTTP vergibt die Korrelations-ID vor allem anderen, ordnet den Request zu und ruft den Handler der einen
// Route oder antwortet 404 oder 405 mit ODR-E2102 (SPEC-238 Nr. 2, 3).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := s.ids.KorrelationsID()
	t := ordne(s.routen, r.Method, r.URL.EscapedPath())
	switch {
	case t.route != nil:
		s.rufe(w, r, id, t)
	case len(t.allow) == 0:
		schreibeFehler(w, model.CodeE2102, http.StatusNotFound, "")
	default:
		schreibeFehler(w, model.CodeE2102, http.StatusMethodNotAllowed, strings.Join(t.allow, ", "))
	}
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

// rufe ruft den Handler der Route genau einmal; Status und Body gehen durch, ein Fehler ist ODR-E9001 mit 500.
func (s *Server) rufe(w http.ResponseWriter, r *http.Request, id string, t treffer) {
	antwort, err := t.route.handler(r.Context(), anfrage(r, id, t.werte))
	if err != nil {
		schreibeFehler(w, model.CodeE9001, http.StatusInternalServerError, "")
		return
	}
	w.WriteHeader(antwort.Status)
	if len(antwort.Body) > 0 {
		_, _ = w.Write(antwort.Body)
	}
}

// schreibeFehler ist die eine Stelle, die das Fehlerobjekt bildet und schreibt (SPEC-238 Nr. 5): drei Felder,
// Content-Type, Content-Length und bei 405 Allow. Der Code ist eine Konstante der Tabelle, das Marshalling
// einer Struktur aus Zeichenketten scheitert nicht.
func schreibeFehler(w http.ResponseWriter, code model.Code, status int, allow string) {
	obj, _ := model.NeuesFehlerobjekt(code)
	body, _ := json.Marshal(obj)
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	if allow != "" {
		h.Set("Allow", allow)
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

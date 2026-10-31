// Package httpodata ist der Driving Adapter des OData-Listeners: Er bildet jeden HTTP-Request
// unverändert auf eine inbound.Anfrage ab, ruft den Dienst des Betriebsmodus als Funktionswert
// und schreibt dessen Antwort oder Fehler als HTTP-Antwort (SPEC-136, LH-FA-01, LH-FA-02.b).
package httpodata

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Verarbeiter ist der Dienst des Betriebsmodus; der Composition Root bindet Record oder Replay
// (SPEC-136 Nr. 2).
type Verarbeiter = func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error)

// Konfig trägt die Abhängigkeiten des Servers; alle drei Felder sind Pflicht (SPEC-136 Nr. 1).
type Konfig struct {
	Verarbeiten Verarbeiter
	IDs         support.IDGenerator
	Listener    net.Listener
}

// Server bedient den Listener; er hält nach New keinen Zustand zwischen Requests (SPEC-136 Nr. 10).
type Server struct {
	verarbeiten Verarbeiter
	ids         support.IDGenerator
	listener    net.Listener
	http        *http.Server
}

// New prüft die Konfiguration; es ruft keinen Port, keine Methode des Listeners und startet
// keine Goroutine (SPEC-136 Nr. 1).
func New(k Konfig) (*Server, error) {
	if k.Verarbeiten == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("httpodata: Verarbeiten fehlt"))
	}
	if k.IDs == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("httpodata: IDs fehlt"))
	}
	if k.Listener == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("httpodata: Listener fehlt"))
	}
	s := &Server{verarbeiten: k.Verarbeiten, ids: k.IDs, listener: k.Listener}
	s.http = &http.Server{
		Handler:  http.HandlerFunc(s.bearbeite),
		ErrorLog: log.New(io.Discard, "", 0),
		// Jeder Request geht an Verarbeiten, auch `OPTIONS *` (SPEC-136 Nr. 2, 4c); net/http
		// beantwortete ihn sonst selbst.
		DisableGeneralOptionsHandler: true,
	}
	return s, nil
}

// Serve bedient den Listener und blockiert; die Rückgabe ist der Fehler des Accept, nie nil
// (SPEC-136 Nr. 12).
func (s *Server) Serve() error {
	err := s.http.Serve(s.listener)
	if err == nil {
		return errors.New("httpodata: Serve endete ohne Fehler")
	}
	return err
}

// bearbeite ist der Ablauf je Request (SPEC-136 Nr. 3): Korrelations-ID, Anfrage, Dienst, Antwort
// oder Fehler.
func (s *Server) bearbeite(w http.ResponseWriter, r *http.Request) {
	id := s.ids.KorrelationsID()
	antwort, err := s.verarbeiten(r.Context(), anfrageAus(r, id))
	if err != nil {
		schreibeFehler(w, r, id, antwort, err)
		return
	}
	schreibeAntwort(w, r, id, antwort)
}

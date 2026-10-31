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
	"sync"
	"sync/atomic"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Die festen Grenzen des Servers (SPEC-136 Nr. 11); ein ReadTimeout und ein WriteTimeout gibt es nicht.
const (
	kopfFrist     = 10 * time.Second
	leerlaufFrist = 120 * time.Second
	kopfBytes     = 65536
)

// ErrFristAbgelaufen meldet, dass Schliessen die Verbindungen schließen musste, weil Handler die
// Frist überschritten (SPEC-136 Nr. 14).
var ErrFristAbgelaufen = errors.New("httpodata: Frist von Schliessen abgelaufen")

// Verarbeiter ist der Dienst des Betriebsmodus; der Composition Root bindet Record oder Replay
// (SPEC-136 Nr. 2).
type Verarbeiter = func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error)

// Konfig trägt die Abhängigkeiten des Servers; alle drei Felder sind Pflicht (SPEC-136 Nr. 1).
type Konfig struct {
	Verarbeiten Verarbeiter
	IDs         support.IDGenerator
	Listener    net.Listener
}

// Server bedient den Listener. Außer dem Zustand der Zulassung (SPEC-136 Nr. 13, 14) hält er
// nach New keinen Zustand zwischen Requests (SPEC-136 Nr. 10).
type Server struct {
	verarbeiten Verarbeiter
	ids         support.IDGenerator
	listener    net.Listener
	lis         *beobachteterListener
	http        *http.Server
	abbruch     context.CancelFunc

	// mu schützt die folgenden Felder; die Zulassung prüft angehalten und zählt unter einer Sperre
	// ohne Freigabe dazwischen.
	mu          sync.Mutex
	angehalten  bool
	geschlossen bool
	laufend     int
	leer        chan struct{}
	leerZu      bool
}

// beobachteterListener merkt, ob ein Accept nach Anhalten scheiterte: nur so ist der Fehler von
// Serve von einem Fehler des Accept davor zu trennen (SPEC-136 Nr. 12).
type beobachteterListener struct {
	net.Listener
	schliesst   atomic.Bool
	nachDemEnde atomic.Bool
}

func (l *beobachteterListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil && l.schliesst.Load() {
		l.nachDemEnde.Store(true)
	}
	return c, err
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
	basis, abbruch := context.WithCancel(context.Background())
	s := &Server{
		verarbeiten: k.Verarbeiten, ids: k.IDs, listener: k.Listener,
		lis:     &beobachteterListener{Listener: k.Listener},
		abbruch: abbruch,
		leer:    make(chan struct{}),
	}
	s.http = &http.Server{
		Handler:           http.HandlerFunc(s.bearbeite),
		ReadHeaderTimeout: kopfFrist,
		IdleTimeout:       leerlaufFrist,
		MaxHeaderBytes:    kopfBytes,
		// Der Kontext jedes Requests stammt von basis; Schliessen beendet ihn beim Ablauf der
		// Frist, auch für einen Request mit ungelesenem Body (SPEC-136 Nr. 4g, 14).
		BaseContext: func(net.Listener) context.Context { return basis },
		ErrorLog:    log.New(io.Discard, "", 0),
		// Jeder Request geht an Verarbeiten, auch `OPTIONS *` (SPEC-136 Nr. 2, 4c): der Server
		// ruft dafür den Handler.
		DisableGeneralOptionsHandler: true,
	}
	return s, nil
}

// Serve bedient den Listener und blockiert. Die Rückgabe ist nil nach Anhalten oder Schliessen,
// auch wenn sie davor lagen, sonst der Fehler des Accept (SPEC-136 Nr. 12).
func (s *Server) Serve() error {
	err := s.http.Serve(s.lis)
	if errors.Is(err, http.ErrServerClosed) || s.lis.nachDemEnde.Load() {
		return nil
	}
	if err == nil {
		return errors.New("httpodata: Serve endete ohne Fehler")
	}
	return err
}

// Anhalten schließt den Listener und wartet auf keinen Request (SPEC-136 Nr. 13). Der Aufruf ist
// wirkungslos, wenn schon angehalten ist; nach seiner Rückkehr ist der Listener zu.
func (s *Server) Anhalten() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.haltAnLocked()
}

// haltAnLocked setzt angehalten und schließt den Listener unter der Sperre des Aufrufers: ein
// zweiter Aufruf kehrt erst zurück, wenn der Listener zu ist (SPEC-136 Nr. 13d).
func (s *Server) haltAnLocked() {
	if s.angehalten {
		return
	}
	s.angehalten = true
	s.lis.schliesst.Store(true)
	defer s.meldeLeerLocked()
	_ = s.listener.Close()
}

// meldeLeerLocked öffnet leer, sobald angehalten ist und kein Handler mehr läuft.
func (s *Server) meldeLeerLocked() {
	if s.angehalten && s.laufend == 0 && !s.leerZu {
		s.leerZu = true
		close(s.leer)
	}
}

// Schliessen hält an, schließt Verbindungen ohne laufenden Request und wartet höchstens frist auf
// die zugelassenen Handler; danach schließt es alle Verbindungen, beendet den Kontext der Requests
// und wartet ohne Frist auf die Rückkehr jedes Handlers (SPEC-136 Nr. 14).
func (s *Server) Schliessen(frist time.Duration) error {
	if frist < 0 {
		return errors.Join(model.ErrUngueltig, errors.New("httpodata: frist ist negativ"))
	}
	s.mu.Lock()
	s.haltAnLocked()
	schon := s.geschlossen
	s.geschlossen = true
	s.mu.Unlock()
	if schon {
		return nil
	}
	defer s.abbruch()

	// Mit einem beendeten Kontext schließt Shutdown die Verbindungen ohne laufenden Request und
	// wartet auf keine weitere.
	beendet, beende := context.WithCancel(context.Background())
	beende()
	_ = s.http.Shutdown(beendet)

	select {
	case <-s.leer:
		return nil
	default:
	}
	timer := time.NewTimer(frist)
	defer timer.Stop()
	select {
	case <-s.leer:
		return nil
	case <-timer.C:
	}
	s.abbruch()
	_ = s.http.Close()
	<-s.leer
	return ErrFristAbgelaufen
}

// zulassen prüft angehalten und zählt den Handler in einem Abschnitt unter der Sperre; false heißt:
// der Request ist abzuweisen und zählt nicht als laufend (SPEC-136 Nr. 13b, 13c).
func (s *Server) zulassen() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.angehalten {
		return false
	}
	s.laufend++
	return true
}

// verlasse nimmt einen zugelassenen Handler aus dem Zähler.
func (s *Server) verlasse() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.laufend--
	s.meldeLeerLocked()
}

// bearbeite ist der Ablauf je Request (SPEC-136 Nr. 3): Korrelations-ID, Zulassung, Anfrage, Dienst,
// Antwort oder Fehler.
func (s *Server) bearbeite(w http.ResponseWriter, r *http.Request) {
	// Der Request-Body bleibt auf jedem Pfad bis Close lesbar (SPEC-136 Nr. 3, 4h); mit Vollduplex
	// verwirft net/http den ungelesenen Rest beim ersten Schreiben nicht. HTTP/1 kennt keinen Fehlerfall.
	_ = http.NewResponseController(w).EnableFullDuplex()
	id := s.ids.KorrelationsID()
	if !s.zulassen() {
		w.Header().Set("Connection", "close")
		schreibeFehlerobjekt(w, r, id, model.CodeE5010, http.StatusServiceUnavailable)
		return
	}
	defer s.verlasse()
	antwort, err := s.verarbeiten(r.Context(), anfrageAus(r, id))
	if err != nil {
		schreibeFehler(w, r, id, antwort, err)
		return
	}
	schreibeAntwort(w, r, id, antwort)
}

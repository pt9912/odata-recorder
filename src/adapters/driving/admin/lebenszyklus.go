package admin

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die festen Grenzen des Servers (SPEC-238 Nr. 10); ein WriteTimeout gibt es nicht (SPEC-124 Ungeprüft).
const (
	kopfFrist     = 10 * time.Second
	lesenFrist    = 30 * time.Second
	leerlaufFrist = 120 * time.Second
	kopfBytes     = 65536
	bodyBytes     = 1 << 20
)

// ErrFristAbgelaufen meldet, dass Schliessen die Verbindungen schließen musste, weil Handler die Frist überschritten
// (SPEC-238 Nr. 11, SPEC-136 Nr. 14).
var ErrFristAbgelaufen = errors.New("admin: Frist von Schliessen abgelaufen")

// verbindung hält, was Schliessen von einer Verbindung wissen muss: ob für die laufende Anfrage ein Handler begonnen
// hat (bis die Verbindung wieder im Leerlauf ist).
type verbindung struct {
	handler bool
}

// verbindungsSchluessel trägt die Verbindung im Kontext des Requests.
type verbindungsSchluessel struct{}

// baueHTTP baut den http.Server mit den festen Werten aus Nr. 10; der Kontext jedes Requests stammt von basis, den
// Schliessen beim Ablauf der Frist beendet, auch für einen Request mit ungelesenem Body (SPEC-238 Nr. 9).
func (s *Server) baueHTTP(basis context.Context) *http.Server {
	return &http.Server{
		Handler:           s,
		ReadHeaderTimeout: kopfFrist,
		ReadTimeout:       lesenFrist,
		IdleTimeout:       leerlaufFrist,
		MaxHeaderBytes:    kopfBytes,
		BaseContext:       func(net.Listener) context.Context { return basis },
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, verbindungsSchluessel{}, c)
		},
		ConnState: s.merkeZustand,
		ErrorLog:  log.New(io.Discard, "", 0),
		// `OPTIONS *` geht durch die Authentisierung (SPEC-238 Nr. 10, 3 d): der Server ruft dafür den Handler.
		DisableGeneralOptionsHandler: true,
	}
}

// merkeZustand führt die Verbindungen: jede ist von der Annahme bis zum Ende bekannt, und ein Handler gilt nur, bis
// die Verbindung wieder im Leerlauf ist.
func (s *Server) merkeZustand(c net.Conn, st http.ConnState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch st {
	case http.StateNew:
		s.verb[c] = &verbindung{}
	case http.StateIdle:
		if v := s.verb[c]; v != nil {
			v.handler = false
		}
	case http.StateHijacked, http.StateClosed:
		delete(s.verb, c)
	case http.StateActive:
	}
}

// markiereHandler vermerkt, dass für die Verbindung des Requests ein Handler begonnen hat.
func (s *Server) markiereHandler(r *http.Request) {
	c, ok := r.Context().Value(verbindungsSchluessel{}).(net.Conn)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if v := s.verb[c]; v != nil {
		v.handler = true
	}
}

// merkListener vermerkt das Ergebnis jedes Accept, das der Server vor dem Ende von Serve sieht.
type merkListener struct {
	net.Listener
	s *Server
}

// Accept reicht den inneren Accept durch. Ein Fehler gilt als vom Schließen verursacht, wenn Anhalten den Listener
// selbst schloss (dessen Close gelang) und der Fehler net.ErrClosed ist; jeder andere Fehler bleibt vermerkt, auch wenn
// Anhalten oder Schliessen vor der Rückkehr lief. Der letzte Fehler gilt; auf einen Fehler, den net/http wiederholt,
// folgt ein weiteres Accept.
func (l merkListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err == nil {
		return c, nil
	}
	l.s.mu.Lock()
	defer l.s.mu.Unlock()
	if l.s.durchAnhalten && errors.Is(err, net.ErrClosed) {
		l.s.acceptFehler = nil
	} else {
		l.s.acceptFehler = err
	}
	return c, err
}

// Serve bedient den Listener und blockiert. Die Rückgabe ist der letzte Fehler des Accept, außer er stammt vom Schließen
// durch Anhalten oder Schliessen selbst; dann, und wenn Serve erst danach begann, ist sie nil (SPEC-136 Nr. 12). Ob
// Anhalten den Listener schloss, entscheidet dessen Close: schlug es fehl, weil der Listener schon zu war, ging der
// Fehler voraus.
func (s *Server) Serve() error {
	_ = s.http.Serve(merkListener{Listener: s.listener, s: s})
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.acceptFehler
}

// Anhalten schließt den Listener und wartet auf keinen Request (SPEC-136 Nr. 13). Der Aufruf ist wirkungslos, wenn schon
// angehalten ist; nach seiner Rückkehr ist der Listener zu.
func (s *Server) Anhalten() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.haltAnLocked()
}

// haltAnLocked setzt angehalten und schließt den Listener unter der Sperre des Aufrufers: ein zweiter Aufruf kehrt erst
// zurück, wenn der Listener zu ist (SPEC-136 Nr. 13 d).
func (s *Server) haltAnLocked() {
	if s.angehalten {
		return
	}
	s.angehalten = true
	defer s.meldeLeerLocked()
	s.durchAnhalten = !errors.Is(s.listener.Close(), net.ErrClosed)
}

// meldeLeerLocked öffnet leer, sobald angehalten ist und kein Handler mehr läuft.
func (s *Server) meldeLeerLocked() {
	if s.angehalten && s.laufend == 0 && !s.leerZu {
		s.leerZu = true
		close(s.leer)
	}
}

// Schliessen hält an, schließt Verbindungen ohne laufenden Request und wartet höchstens frist auf die zugelassenen
// Handler; danach schließt es alle Verbindungen, beendet den Kontext der Requests und wartet ohne Frist auf die Rückkehr
// jedes Handlers (SPEC-136 Nr. 14).
func (s *Server) Schliessen(frist time.Duration) error {
	if frist < 0 {
		return ungueltig("frist ist negativ")
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

	// Mit einem beendeten Kontext schließt Shutdown die Verbindungen im Leerlauf und wartet auf keine weitere; danach
	// schließt der Server jede Verbindung nach ihrer Antwort.
	beendet, beende := context.WithCancel(context.Background())
	beende()
	_ = s.http.Shutdown(beendet)
	s.schliesseOhneHandler()

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

// schliesseOhneHandler schließt jede Verbindung, für deren Anfrage kein Handler begonnen hat: neue ohne ein Byte, mit
// halbem Kopf und im Leerlauf.
func (s *Server) schliesseOhneHandler() {
	s.mu.Lock()
	var zu []net.Conn
	for c, v := range s.verb {
		if !v.handler {
			zu = append(zu, c)
		}
	}
	s.mu.Unlock()
	for _, c := range zu {
		_ = c.Close()
	}
}

// zulassen prüft angehalten und zählt den Handler in einem Abschnitt unter der Sperre; false heißt: der Request ist
// abzuweisen und zählt nicht als laufend (SPEC-136 Nr. 13 b, 13 c).
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

// abweisen schreibt die Antwort eines angehaltenen Servers: 503 ODR-E5010 mit Connection: close (SPEC-238 Nr. 3 b, 5).
func abweisen(w http.ResponseWriter) {
	w.Header().Set("Connection", "close")
	schreibeFehler(w, model.CodeE5010, http.StatusServiceUnavailable, "", "")
}

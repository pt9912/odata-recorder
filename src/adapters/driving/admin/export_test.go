package admin

import (
	"log"
	"net/http"
	"time"
)

// Grenzen sind die Felder des http.Server, die SPEC-238 Nr. 10 festlegt, in der Reihenfolge der Vertragszeile.
type Grenzen struct {
	ReadHeaderTimeout, ReadTimeout, IdleTimeout time.Duration
	MaxHeaderBytes                              int
	WriteTimeout                                time.Duration
	ErrorLog                                    *log.Logger
	DisableGeneralOptionsHandler                bool
}

// GrenzenFuerTest liefert die Felder des http.Server (SPEC-238 Nr. 10).
func (s *Server) GrenzenFuerTest() Grenzen {
	h := s.http
	return Grenzen{
		ReadHeaderTimeout: h.ReadHeaderTimeout, ReadTimeout: h.ReadTimeout, IdleTimeout: h.IdleTimeout,
		MaxHeaderBytes: h.MaxHeaderBytes, WriteTimeout: h.WriteTimeout, ErrorLog: h.ErrorLog,
		DisableGeneralOptionsHandler: h.DisableGeneralOptionsHandler,
	}
}

// LadeMitHaltepunkten ruft lade mit Haltepunkten: nachAnlegen läuft nach dem Anlegen der Hilfsdatei, vorLink
// vor dem Link (SPEC-239 Nr. 3); beide erhalten den Pfad der Hilfsdatei, nil heißt: kein Halt.
func LadeMitHaltepunkten(pfad string, nachAnlegen, vorLink func(hilfsdatei string)) ([]byte, error) {
	if nachAnlegen == nil {
		nachAnlegen = keinHalt
	}
	if vorLink == nil {
		vorLink = keinHalt
	}
	return lade(pfad, haltepunkte{nachAnlegen: nachAnlegen, vorLink: vorLink})
}

// BildeAnfrage führt anfrage heraus: die Anfrage, die die Route zu r, id und werten erhält (SPEC-238 Nr. 2).
func BildeAnfrage(r *http.Request, id string, werte map[string]string) Anfrage {
	return anfrage(r, id, werte)
}

// Vergleiche führt vergleiche heraus (SPEC-238 Nr. 4 e).
func Vergleiche(erwartet, gegeben []byte) bool { return vergleiche(erwartet, gegeben) }

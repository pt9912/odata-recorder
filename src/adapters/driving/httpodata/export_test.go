package httpodata

import "time"

// GrenzenFuerTest liefert die Felder des http.Server, die SPEC-136 Nr. 11 festlegt.
func (s *Server) GrenzenFuerTest() (kopf, leerlauf time.Duration, kopfGroesse int, lesen, schreiben time.Duration) {
	return s.http.ReadHeaderTimeout, s.http.IdleTimeout, s.http.MaxHeaderBytes, s.http.ReadTimeout, s.http.WriteTimeout
}

package capture

import (
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// schreiber ist das dst des Redaktors: er eröffnet den CaptureWriter erst beim ersten Byte
// der Ausgabe, damit ein Body ohne Ausgabe keine temporäre Datei anlegt (SPEC-119 Nr. 8).
type schreiber struct {
	begin func() (outbound.CaptureWriter, error)
	w     outbound.CaptureWriter
	// offen ist wahr, solange ein Writer, den Begin lieferte, weder Commit noch Abort hatte.
	offen       bool
	geschrieben int64
	// behalte hält die Bytes, die der Writer angenommen hat (SPEC-119 Nr. 13).
	behalte   bool
	redigiert []byte
}

func (s *schreiber) Write(p []byte) (int, error) {
	if s.w == nil {
		w, err := s.begin()
		if err != nil {
			return 0, err
		}
		s.w, s.offen = w, true
	}
	n, err := s.w.Write(p)
	s.geschrieben += int64(n)
	if s.behalte {
		s.redigiert = append(s.redigiert, p[:n]...)
	}
	return n, err
}

// verwirf ruft Abort genau einmal, falls ein Writer offen ist; ein Fehler von Abort ändert
// das Ergebnis nicht (SPEC-119 Nr. 8).
func (s *schreiber) verwirf() {
	if s.offen {
		s.offen = false
		_ = s.w.Abort()
	}
}

package capture

import (
	"bufio"
	"compress/gzip"
	"compress/zlib"
	"io"
	"strings"
)

// maxGlieder ist die Zahl der Glieder, die eine Kette höchstens hat; identity zählt mit (SPEC-119 Nr. 6).
const maxGlieder = 2

// parseKette liest die Werte von Content-Encoding: geteilt an Kommas, Leerzeichen und
// Tabulatoren am Rand entfernt, ohne Beachtung der ASCII-Schreibweise. ok ist falsch bei einem
// unbekannten Namen, einem leeren Teil und mehr als maxGlieder Gliedern (SPEC-119 Nr. 6).
func parseKette(werte []string) (kette []string, ok bool) {
	for _, w := range werte {
		for _, teil := range strings.Split(w, ",") {
			name := asciiKlein(strings.Trim(teil, " \t"))
			if !bekannt(name) {
				return nil, false
			}
			kette = append(kette, name)
			if len(kette) > maxGlieder {
				return nil, false
			}
		}
	}
	return kette, true
}

func bekannt(name string) bool {
	return name == "gzip" || name == "deflate" || name == "identity"
}

func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// stufe entpackt ein Glied der Kette als Strom. Sie baut den Dekompressor beim ersten Read, so
// dass jeder Fehler beim Lesen des Kopfes als Fehler des Stroms beim Redaktor ankommt. Ein
// Eingang ohne ein Byte ist ein leerer Strom (io.EOF), kein Fehler des Kopfes (SPEC-119 Nr. 8).
type stufe struct {
	art    string
	quelle io.Reader
	br     *bufio.Reader
	leser  io.Reader
	fehler error
}

func (s *stufe) Read(p []byte) (int, error) {
	if s.br == nil {
		s.baue()
	}
	if s.fehler != nil {
		return 0, s.fehler
	}
	return s.leser.Read(p)
}

func (s *stufe) baue() {
	s.br = bufio.NewReader(s.quelle)
	if _, err := s.br.Peek(1); err != nil {
		s.fehler = err
		return
	}
	if s.art == "gzip" {
		s.leser, s.fehler = gzip.NewReader(s.br)
		return
	}
	s.leser, s.fehler = zlib.NewReader(s.br)
}

// uebrig prüft nach dem Ende des entpackten Stroms, ob die Eingabe der Stufe erschöpft ist:
// nil heißt, ein Byte steht noch an (Nachlauf), io.EOF, die Eingabe ist zu Ende.
func (s *stufe) uebrig() error {
	if s.br == nil {
		return io.EOF
	}
	_, err := s.br.Peek(1)
	return err
}

// baueKette entpackt in umgekehrter Reihenfolge der Kette: das letzte Glied liegt auf den
// Rohbytes (SPEC-119 Nr. 6). identity entpackt nichts. Die Stufen stehen von den Rohbytes aus.
func baueKette(kette []string, roh io.Reader) (oben io.Reader, stufen []*stufe) {
	oben = roh
	for i := len(kette) - 1; i >= 0; i-- {
		if kette[i] == "identity" {
			continue
		}
		s := &stufe{art: kette[i], quelle: oben}
		stufen = append(stufen, s)
		oben = s
	}
	return oben, stufen
}

// verhaeltnisOK meldet entpackt ≤ grenze × roh, ohne zu multiplizieren (SPEC-119 Nr. 7).
func verhaeltnisOK(roh, entpackt, grenze int64) bool {
	if roh < 1 {
		return entpackt == 0
	}
	q := entpackt / roh
	return q < grenze || (q == grenze && entpackt%roh == 0)
}

// zaehlLeser zählt die Bytes am Eingang des Redaktors.
type zaehlLeser struct {
	r io.Reader
	n int64
}

func (z *zaehlLeser) Read(p []byte) (int, error) {
	n, err := z.r.Read(p)
	z.n += int64(n)
	return n, err
}

package capture

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

// Ursachen, die der Leser der Goroutine als Fehler sieht. errAusserhalb heißt: der Kontext der
// Aufnahme ist beendet (Überlauf, Frist, Abbrechen, Abbruch von ctx) oder die Queue ist
// geschlossen; die Ursache liefert dann context.Cause.
var (
	errAusserhalb = errors.New("capture: Aufnahme von außen beendet")
	errRohgroesse = errors.New("capture: Rohgröße über MaxBytes")
)

// srcFehler trägt einen Fehler von src ungleich io.EOF durch Dekompressor und Redaktor.
type srcFehler struct{ err error }

func (*srcFehler) Error() string   { return "capture: Fehler von src" }
func (e *srcFehler) Unwrap() error { return e.err }

type queueEnde uint8

const (
	endeOffen queueEnde = iota
	endeEOF
	endeQuelle
	endeRoh
)

// queue ist die begrenzte Warteschlange zwischen Tee und Goroutine der Aufnahme (SPEC-119
// Nr. 4, 5). Einlegen blockiert nie; die Queue hält höchstens limit Rohbytes.
type queue struct {
	mu          sync.Mutex
	stuecke     [][]byte
	ausstehend  int64
	hoechst     int64
	roh         int64
	limit       int64
	maxRoh      int64
	zu          bool
	ende        queueEnde
	quellFehler error
	// ueberlauf beendet die Aufnahme mit der Ursache Überlauf; er läuft unter mu, damit der
	// Leser nach einer geschlossenen Queue die Ursache im Kontext findet.
	ueberlauf func()
	wecken    chan struct{}
}

func neueQueue(limit, maxRoh int64, ueberlauf func()) *queue {
	return &queue{limit: limit, maxRoh: maxRoh, ueberlauf: ueberlauf, wecken: make(chan struct{}, 1)}
}

func (q *queue) melde() {
	select {
	case q.wecken <- struct{}{}:
	default:
	}
}

// lege legt eine Kopie von p ein, solange die Queue offen ist; err ist der Fehler desselben Reads.
func (q *queue) lege(p []byte, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.zu || q.ende != endeOffen {
		return
	}
	stueck, roh := q.begrenze(p)
	if len(stueck) > 0 {
		if q.ausstehend+int64(len(stueck)) > q.limit {
			q.verwerfe()
			q.ueberlauf()
			return
		}
		q.stuecke = append(q.stuecke, bytes.Clone(stueck))
		q.ausstehend += int64(len(stueck))
		q.hoechst = max(q.hoechst, q.ausstehend)
	}
	switch {
	case roh:
		q.ende = endeRoh
	case err == io.EOF:
		q.ende = endeEOF
	case err != nil:
		q.ende = endeQuelle
		q.quellFehler = err
	}
	q.melde()
}

// begrenzt zählt die Rohbytes aller Reads und kürzt p auf die Bytes bis MaxBytes (SPEC-119 Nr. 5);
// roh meldet, dass mehr als MaxBytes geliefert wurden.
func (q *queue) begrenze(p []byte) (stueck []byte, roh bool) {
	q.roh += int64(len(p))
	if q.roh <= q.maxRoh {
		return p, false
	}
	return p[:int64(len(p))-(q.roh-q.maxRoh)], true
}

// verwerfe leert die Queue und schließt sie; die Bytes darin sind verloren (SPEC-119 Nr. 9).
func (q *queue) verwerfe() {
	q.zu = true
	q.stuecke = nil
	q.ausstehend = 0
	q.melde()
}

// schliesse beendet die Queue, wenn das Ergebnis bestimmt ist: sie ist leer, und Read kopiert nichts mehr.
func (q *queue) schliesse() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.verwerfe()
}

// entnimm kopiert bis zu len(p) Bytes aus der Queue. warte heißt: weder Bytes noch Ende da.
// Ist die Queue danach leer und mit io.EOF beendet, kommt das Ende mit den Bytes zurück.
func (q *queue) entnimm(p []byte) (n int, warte bool, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.stuecke) > 0 {
		n = copy(p, q.stuecke[0])
		q.ausstehend -= int64(n)
		if n == len(q.stuecke[0]) {
			q.stuecke[0] = nil
			q.stuecke = q.stuecke[1:]
		} else {
			q.stuecke[0] = q.stuecke[0][n:]
		}
		if len(q.stuecke) == 0 && q.ende == endeEOF {
			return n, false, io.EOF
		}
		return n, false, nil
	}
	if q.zu {
		return 0, false, errAusserhalb
	}
	switch q.ende {
	case endeEOF:
		return 0, false, io.EOF
	case endeQuelle:
		return 0, false, &srcFehler{err: q.quellFehler}
	case endeRoh:
		return 0, false, errRohgroesse
	}
	return 0, true, nil
}

// zustand liefert Wasserstand und Höchststand der Queue.
func (q *queue) zustand() (ausstehend, hoechst int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.ausstehend, q.hoechst
}

// queueLeser ist der Leser der Goroutine auf der Queue. Er wartet auf Bytes oder Ende und kehrt
// sofort zurück, sobald der Kontext der Aufnahme beendet ist (SPEC-119 Nr. 9).
type queueLeser struct {
	q    *queue
	done <-chan struct{}
	// roh zählt die entnommenen Rohbytes.
	roh int64
}

func (r *queueLeser) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for {
		select {
		case <-r.done:
			return 0, errAusserhalb
		default:
		}
		n, warte, err := r.q.entnimm(p)
		if !warte {
			r.roh += int64(n)
			return n, err
		}
		select {
		case <-r.q.wecken:
		case <-r.done:
		}
	}
}

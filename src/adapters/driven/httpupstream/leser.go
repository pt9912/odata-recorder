package httpupstream

import (
	"errors"
	"io"
	"sync"
)

// maxStueck begrenzt ein Stück, das die Goroutine des Lesers auf eine Anforderung liest.
const maxStueck = 32 << 10

// errBeendet ist die eigene Unterbrechung des Lesens mit dem Ende des Austauschs; sie ist kein
// Fehler des Bodys (SPEC-120 Nr. 4).
var errBeendet = errors.New("httpupstream: Lesen des Request-Bodys beendet")

// leser ist der Request-Body, den die Bibliothek liest (SPEC-120 Nr. 4). Eine Goroutine je Anfrage
// ruft Read der Quelle nur, wenn die Bibliothek ein Stück anfordert, und in einen eigenen Puffer.
// Endet der Austausch (ende), kehrt Read sofort zurück, auch wenn ein Read der Quelle noch läuft;
// dessen Ergebnis wird verworfen, und ein weiteres Read der Quelle gibt es nicht.
type leser struct {
	quelle io.Reader
	ende   <-chan struct{}
	anf    chan int
	erg    chan stueck

	mu           sync.Mutex
	fehlerQuelle error

	// letzter ist das Ergebnis nach io.EOF oder einem Fehler; nur Read liest und schreibt es.
	letzter error
}

type stueck struct {
	daten []byte
	err   error
}

func neuerLeser(ende <-chan struct{}, quelle io.Reader) *leser {
	l := &leser{quelle: quelle, ende: ende, anf: make(chan int), erg: make(chan stueck)}
	go l.lauf()
	return l
}

// Read fordert ein Stück an und wartet auf sein Ergebnis oder auf das Ende des Austauschs.
func (l *leser) Read(p []byte) (int, error) {
	if l.letzter != nil {
		return 0, l.letzter
	}
	if len(p) == 0 {
		return 0, nil
	}
	select {
	case l.anf <- min(len(p), maxStueck):
	case <-l.ende:
		return 0, errBeendet
	}
	select {
	case s := <-l.erg:
		l.letzter = s.err
		return copy(p, s.daten), s.err
	case <-l.ende:
		return 0, errBeendet
	}
}

// Close ist ohne Wirkung: das Lesen endet mit dem Austausch.
func (*leser) Close() error { return nil }

// fehler liefert den Fehler, den die Quelle selbst lieferte, außer io.EOF; nil ohne Leser.
func (l *leser) fehler() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.fehlerQuelle
}

func (l *leser) beendet() bool {
	select {
	case <-l.ende:
		return true
	default:
		return false
	}
}

// lauf liest je Anforderung ein Stück. Ein Ergebnis, das nach dem Ende kommt, verwirft er; nach
// io.EOF oder einem Fehler endet er.
func (l *leser) lauf() {
	var puffer []byte
	for {
		var n int
		select {
		case n = <-l.anf:
		case <-l.ende:
			return
		}
		// Grenze: fällt das Ende mit der Anforderung zusammen, wählt select zufällig; diese Prüfung
		// ruft dann kein Read. Ein Test dafür fehlt, er müsste die Wahl von select steuern.
		if l.beendet() {
			return
		}
		if cap(puffer) < n {
			puffer = make([]byte, n)
		}
		k, err := l.quelle.Read(puffer[:n])
		if l.beendet() {
			return
		}
		if err != nil && err != io.EOF {
			l.mu.Lock()
			l.fehlerQuelle = err
			l.mu.Unlock()
		}
		select {
		case l.erg <- stueck{daten: puffer[:k], err: err}:
		case <-l.ende:
			return
		}
		if err != nil {
			return
		}
	}
}

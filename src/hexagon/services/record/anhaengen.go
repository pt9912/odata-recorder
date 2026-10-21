package record

import (
	"context"
	"errors"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Ursachen, mit denen der Kontext eines Exchange endet: errDrain, wenn Drain ihn beendet,
// errEnde nach dem Anhängen und bei einem abgewiesenen Request.
var (
	errDrain = errors.New("record: Drain beendet den Exchange")
	errEnde  = errors.New("record: Exchange beendet")
)

// lauf ist ein laufender Exchange, wie Drain ihn sieht: er hält die Funktion, die seinen Kontext
// beendet, und die Ursache, mit der der Kontext endete.
type lauf struct {
	abbrechen context.CancelCauseFunc
	ursache   func() error
}

// sessionZustand ist der Zustand je Session in diesem Dienst (SPEC-121 Nr. 10, 13).
type sessionZustand struct {
	// unter Recorder.mu
	drain   bool
	laufend map[*lauf]struct{}
	// leer schließt, wenn laufend leer wird; ein neuer Kanal entsteht, wenn der erste Exchange beginnt.
	leer    chan struct{}
	bericht inbound.DrainBericht

	// anhaengen ordnet die Anhänge der Session nacheinander und schützt zahl und zahlBekannt.
	anhaengen   sync.Mutex
	zahl        int
	zahlBekannt bool
}

// zustand liefert den Zustand einer Session und legt ihn an; der Aufrufer hält r.mu.
func (r *Recorder) zustand(id string) *sessionZustand {
	sz := r.sessions[id]
	if sz == nil {
		sz = &sessionZustand{laufend: map[*lauf]struct{}{}}
		r.sessions[id] = sz
	}
	return sz
}

// haengeAn hängt den Exchange mit der Zahl der Session plus eins an (SPEC-121 Nr. 10). Hat der
// Dienst keine Zahl für die Session, liest er sie mit ListBySession, sonst nicht. Keine Zahl hat
// er, bis ein Lesen gelingt, und nach ErrUngueltigeSequenz aus Append, das sie verwirft; nach
// jedem anderen Fehler von Append bleibt sie. Sie steigt nur, wenn Append gelingt.
func (r *Recorder) haengeAn(ctx context.Context, sz *sessionZustand, e model.Exchange) error {
	sz.anhaengen.Lock()
	defer sz.anhaengen.Unlock()
	if !sz.zahlBekannt {
		liste, err := r.k.Exchanges.ListBySession(ctx, e.SessionID)
		if err != nil {
			return err
		}
		sz.zahl, sz.zahlBekannt = len(liste), true
	}
	e.SequenceNo = sz.zahl + 1
	if err := r.k.Exchanges.Append(ctx, e); err != nil {
		if errors.Is(err, outbound.ErrUngueltigeSequenz) {
			sz.zahlBekannt = false
		}
		return err
	}
	sz.zahl++
	return nil
}

// endeLauf beendet den Lauf eines Exchange: er zählt ihn im Bericht, wenn er nicht angehängt ist
// und wenn Drain seinen Kontext beendete, und nimmt ihn aus den laufenden (SPEC-121 Nr. 9, 13).
func (r *Recorder) endeLauf(sz *sessionZustand, l *lauf, nichtAngehaengt bool) {
	r.mu.Lock()
	if nichtAngehaengt {
		sz.bericht.NichtAngehaengt++
	}
	if errors.Is(l.ursache(), errDrain) {
		sz.bericht.Abgebrochen++
	}
	delete(sz.laufend, l)
	if len(sz.laufend) == 0 {
		close(sz.leer)
	}
	r.mu.Unlock()
	l.abbrechen(errEnde)
}

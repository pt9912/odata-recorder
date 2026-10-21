package record

import (
	"context"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Drain nimmt für die Session ab dem Aufruf keinen Exchange mehr an, wartet höchstens frist auf die
// laufenden, beendet danach (oder bei beendetem ctx) den Kontext jedes laufenden und wartet ohne
// Frist auf sein Ende (SPEC-121 Nr. 13). Drain liest und ändert die Ablage nicht.
func (r *Recorder) Drain(ctx context.Context, sessionID string, frist time.Duration) (inbound.DrainBericht, error) {
	if ctx == nil || sessionID == "" || frist < 0 {
		return inbound.DrainBericht{}, ungueltig("Drain ohne Kontext, ohne Session oder mit negativer Frist")
	}
	if leer := r.sperre(sessionID); leer != nil && !warte(ctx, leer, frist) {
		<-r.brichAb(sessionID)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sessions[sessionID].bericht, nil
}

// sperre setzt die Session dauerhaft in den Drain und liefert den Kanal, der schließt, wenn kein
// Exchange mehr läuft; nil, wenn keiner läuft.
func (r *Recorder) sperre(id string) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	sz := r.zustand(id)
	sz.drain = true
	if len(sz.laufend) == 0 {
		return nil
	}
	return sz.leer
}

// warte meldet, ob leer schloss, bevor frist ablief oder ctx endete.
func warte(ctx context.Context, leer <-chan struct{}, frist time.Duration) bool {
	wctx, stop := context.WithTimeout(ctx, frist)
	defer stop()
	select {
	case <-leer:
		return true
	case <-wctx.Done():
		return false
	}
}

// brichAb beendet den Kontext jedes laufenden Exchange der Session; der Abbruch wirkt nur über
// diesen Kontext.
func (r *Recorder) brichAb(id string) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	sz := r.sessions[id]
	for l := range sz.laufend {
		l.abbrechen(errDrain)
	}
	return sz.leer
}

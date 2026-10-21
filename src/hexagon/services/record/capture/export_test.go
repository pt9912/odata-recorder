package capture

import "math"

// Hilfen für die externen Tests: Wasserstand und Höchststand der Queue, die Queue ohne Aufnahme,
// der Zähler am Eingang des Redaktors, das Anhalten der Goroutine am Start und die reine
// Verhältnisprüfung.

// QueueAllein ist die Queue einer Aufnahme ohne Goroutine: der Test legt ein und entnimmt in
// Teilen, die er wählt.
type QueueAllein struct {
	q         *queue
	ueberlauf bool
}

// NeueQueueAllein liefert eine Queue mit der Grenze limit und ohne Grenze der Rohgröße.
func NeueQueueAllein(limit int64) *QueueAllein {
	qa := &QueueAllein{}
	qa.q = neueQueue(limit, math.MaxInt64, func() { qa.ueberlauf = true })
	return qa
}

// Lege legt p ein wie ein Read ohne Fehler und meldet, ob die Queue übergelaufen ist.
func (qa *QueueAllein) Lege(p []byte) bool {
	qa.q.lege(p, nil)
	return qa.ueberlauf
}

// Entnimm entnimmt höchstens n Bytes, wie der Leser der Goroutine mit einem Puffer der Länge n.
func (qa *QueueAllein) Entnimm(n int) []byte {
	p := make([]byte, n)
	k, _, _ := qa.q.entnimm(p)
	return p[:k]
}

// Ausstehend ist die Zahl der eingelegten und noch nicht entnommenen Bytes.
func (qa *QueueAllein) Ausstehend() int64 {
	n, _ := qa.q.zustand()
	return n
}

// Ausstehend ist die Zahl der eingelegten und noch nicht entnommenen Rohbytes.
func (a *Aufnahme) Ausstehend() int64 {
	n, _ := a.q.zustand()
	return n
}

// Hoechststand ist der größte Wasserstand der Queue seit Tee.
func (a *Aufnahme) Hoechststand() int64 {
	_, h := a.q.zustand()
	return h
}

// Entpackt wartet auf das Ergebnis und liefert die Bytes, die der Redaktor gelesen hat.
func (a *Aufnahme) Entpackt() int64 {
	<-a.fertig
	return a.entpackt
}

// AnhaltenAmStart hält die Goroutine jeder Aufnahme dieser Pipeline am Start an, bis der
// Aufruf von freigeben sie loslässt. Er gilt nur vor dem ersten Tee.
func (p *Pipeline) AnhaltenAmStart() (freigeben func()) {
	c := make(chan struct{})
	p.start = c
	return func() { close(c) }
}

// Verhaeltnis ist die Prüfung entpackt ≤ grenze × roh.
func Verhaeltnis(roh, entpackt, grenze int64) bool { return verhaeltnisOK(roh, entpackt, grenze) }

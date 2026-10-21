package capture

// Hilfen für die externen Tests: Wasserstand und Höchststand der Queue, der Zähler am Eingang
// des Redaktors, das Anhalten der Goroutine am Start und die reine Verhältnisprüfung.

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

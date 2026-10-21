package capture

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Ursachen, die der Kontext der Aufnahme als Cause trägt, und die Ursachen der Prüfungen nach dem Lesen.
var (
	errFrist     = fmt.Errorf("capture: Frist der Pipeline abgelaufen: %w", context.DeadlineExceeded)
	errUeberlauf = errors.New("capture: Queue übergelaufen")
	errEncoding  = errors.New("capture: Content-Encoding nicht entpackbar")
	errNachlauf  = errors.New("capture: Nachlauf hinter dem entpackten Strom")
	errBombe     = errors.New("capture: Verhältnis von entpackt zu roh überschritten")
	errLaenge    = errors.New("capture: Länge des Commit weicht von den geschriebenen Bytes ab")
	errBeendet   = errors.New("capture: Aufnahme beendet")
)

// Aufnahme ist ein Body im Tee: Read reicht den Strom unverändert durch, eine Goroutine nimmt
// ihn auf (SPEC-119). Read ruft nur eine Goroutine auf; Abbrechen und Ergebnis jede.
type Aufnahme struct {
	p        *Pipeline
	src      io.Reader
	q        *queue
	richtung Richtung
	eingang  model.BodyEingang
	kette    []string
	abbruch  context.CancelCauseFunc
	// vorab ist das Ergebnis, das schon in Tee feststeht (SPEC-119 Nr. 6); nil sonst.
	vorab *Ergebnis
	// fertig ist geschlossen, sobald ergebnis und entpackt gesetzt sind.
	fertig   chan struct{}
	ergebnis Ergebnis
	// entpackt zählt die Bytes am Eingang des Redaktors; nur die Tests lesen es (Entpackt in export_test.go).
	entpackt int64
}

// Read reicht (n, err) von src unverändert durch, wartet nie auf die Aufnahme und behält p nicht
// (SPEC-119 Nr. 3, 4).
func (a *Aufnahme) Read(p []byte) (int, error) {
	n, err := a.src.Read(p)
	if n > 0 || err != nil {
		a.q.lege(p[:n], err)
	}
	return n, err
}

// Abbrechen beendet die Aufnahme mit dem Ausgang Abbruch, wartet auf nichts und lässt Read
// unberührt; nach dem Ergebnis ist es wirkungslos (SPEC-119 Nr. 14).
func (a *Aufnahme) Abbrechen() { a.abbruch(nil) }

// Ergebnis wartet auf die Goroutine der Aufnahme und liefert bei jedem Aufruf denselben Wert
// (SPEC-119 Nr. 11).
func (a *Aufnahme) Ergebnis() Ergebnis {
	<-a.fertig
	return a.ergebnis
}

// arbeite ist die Goroutine der Aufnahme; sie ist die einzige, die Dekompressor, Redaktor und
// Writer benutzt, und endet in jedem Fall.
func (a *Aufnahme) arbeite(ctx context.Context) {
	if a.p.start != nil {
		<-a.p.start
	}
	e := a.verarbeite(ctx)
	a.q.schliesse()
	a.ergebnis = e
	a.abbruch(errBeendet)
	close(a.fertig)
}

func (a *Aufnahme) verarbeite(ctx context.Context) Ergebnis {
	if a.vorab != nil {
		return *a.vorab
	}
	roh := &queueLeser{q: a.q, done: ctx.Done()}
	oben, stufen := baueKette(a.kette, roh)
	zl := &zaehlLeser{r: oben}
	w := &schreiber{
		begin:   func() (outbound.CaptureWriter, error) { return a.p.konfig.Blobs.Begin(ctx, a.p.konfig.MaxBytes) },
		behalte: a.richtung == Request,
	}
	regeln := model.BodyRegeln{Pfade: a.p.konfig.Pfade, MaxBytes: a.p.konfig.MaxBytes}
	status, err := model.RedigiereBody(w, zl, a.eingang, regeln)
	a.entpackt = zl.n
	if err != nil {
		return a.verwirf(w, befundRedaktor(ctx, err))
	}
	if b := a.pruefeNachLesen(ctx, stufen, roh.roh, zl.n); b != nil {
		return a.verwirf(w, *b)
	}
	if ctx.Err() != nil {
		return a.verwirf(w, befundAusserhalb(ctx))
	}
	return a.schliesse(ctx, w, status)
}

func (a *Aufnahme) verwirf(w *schreiber, b befund) Ergebnis {
	w.verwirf()
	return verworfen(a.richtung, b)
}

// schliesse ruft Commit, wenn ein Writer offen ist; ein gelungener Commit wird nie zurückgenommen
// (SPEC-119 Nr. 8).
func (a *Aufnahme) schliesse(ctx context.Context, w *schreiber, status model.RedactionStatus) Ergebnis {
	if w.w == nil {
		return ohneBody(a.richtung)
	}
	ref, laenge, err := w.w.Commit()
	if err != nil {
		return a.verwirf(w, befundWriter(ctx, err))
	}
	w.offen = false
	if laenge != w.geschrieben {
		return verworfen(a.richtung, befund{ausgang: Verstoss, grund: GrundAblage, ursache: errLaenge})
	}
	e := Ergebnis{
		Richtung:  a.richtung,
		Ausgang:   Vollstaendig,
		Body:      model.BodyPresent,
		Blob:      ref,
		Redaction: status,
		Integrity: model.IntegrityVerified,
	}
	if a.richtung == Request {
		e.Redigiert = w.redigiert
	}
	return e
}

// pruefeNachLesen prüft in der Reihenfolge der Nr. 7: erst die Erschöpfung der Queue, dann das Verhältnis.
func (a *Aufnahme) pruefeNachLesen(ctx context.Context, stufen []*stufe, roh, entpackt int64) *befund {
	for _, s := range stufen {
		err := s.uebrig()
		if err == nil {
			return &befund{ausgang: Verstoss, grund: GrundDekodierung, ursache: errNachlauf}
		}
		if err != io.EOF {
			b := befundEingabe(ctx, err)
			return &b
		}
	}
	if !verhaeltnisOK(roh, entpackt, a.p.konfig.MaxVerhaeltnis) {
		return &befund{ausgang: Verstoss, grund: GrundBombe, ursache: errBombe}
	}
	return nil
}

// befundAusserhalb ordnet die Ursache zu, die den Kontext der Aufnahme beendete (SPEC-119 Nr. 9, 10):
// eine Deadline, die ctx des Aufrufers trägt, ist ein Abbruch und nicht die Frist der Pipeline.
func befundAusserhalb(ctx context.Context) befund {
	ursache := context.Cause(ctx)
	switch {
	case errors.Is(ursache, errFrist):
		return befund{ausgang: Verstoss, grund: GrundFrist, ursache: ursache}
	case errors.Is(ursache, errUeberlauf):
		return befund{ausgang: Verstoss, grund: GrundUeberlauf, ursache: ursache}
	}
	return befund{ausgang: Abbruch, grund: GrundAbbruch, ursache: ursache}
}

// befundRedaktor ordnet einen Fehler von RedigiereBody zu (SPEC-119 Nr. 10). Ein Fehler, der
// kein *model.UnsicherFehler ist, wird als ablage gemeldet; kein Test fängt diesen Zweig
// (SPEC-119 Ungeprüft).
func befundRedaktor(ctx context.Context, err error) befund {
	var uf *model.UnsicherFehler
	switch {
	case !errors.As(err, &uf):
		return befund{ausgang: Verstoss, grund: GrundAblage, ursache: err}
	case uf.Grund == model.GrundEingabe:
		return befundEingabe(ctx, err)
	case uf.Grund == model.GrundAusgabe:
		return befundWriter(ctx, err)
	}
	return befund{ausgang: Verstoss, grund: Grund(uf.Grund), ursache: err}
}

// befundEingabe ordnet einen Fehler des Stroms vor dem Redaktor zu: Rohgröße, Fehler von src,
// Ende von außen oder, sonst, ein Fehler der Dekompression.
func befundEingabe(ctx context.Context, err error) befund {
	var sf *srcFehler
	switch {
	case errors.Is(err, errRohgroesse):
		return befund{ausgang: Verstoss, grund: GrundZuGross, ursache: err}
	case errors.As(err, &sf):
		return befund{ausgang: Abbruch, grund: GrundAbbruch, ursache: sf.err}
	case errors.Is(err, errAusserhalb):
		return befundAusserhalb(ctx)
	}
	return befund{ausgang: Verstoss, grund: GrundDekodierung, ursache: err}
}

// befundWriter ordnet einen Fehler des Writers zu: ein Fehler des Kontexts zählt als die
// Ursache, die den Kontext beendete.
func befundWriter(ctx context.Context, err error) befund {
	switch {
	case errors.Is(err, outbound.ErrZuGross):
		return befund{ausgang: Verstoss, grund: GrundAusgabeZuGross, ursache: err}
	case ctx.Err() != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)):
		return befundAusserhalb(ctx)
	}
	return befund{ausgang: Verstoss, grund: GrundAblage, ursache: err}
}

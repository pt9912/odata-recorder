package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// probenIntervall ist der Abstand der Proben der Last während eines Laufs (SPEC-216 Nr. 11).
const probenIntervall = 30 * time.Second

// errLast ist der Fehler eines Aufbaus, den eine Probe über der Schwelle verhindert (Exit 3, Grund last).
var errLast = errors.New("Last über der Schwelle")

// lastQuelle liefert den Inhalt von /proc/loadavg; der Test ersetzt sie (SPEC-216 Nr. 14 c).
type lastQuelle interface {
	Lies() ([]byte, error)
}

// dateiLast liest die Last aus einer Datei.
type dateiLast struct{ pfad string }

func (d dateiLast) Lies() ([]byte, error) { return os.ReadFile(d.pfad) }

// probe ist eine Probe der Last in Hundertsteln; Zulaessig ist 2 × Last ≤ Kerne. Eine Probe, die sich nicht lesen lässt, trägt den kleinsten
// Wert über der Schwelle und ist nicht zulässig (SPEC-216 Nr. 11).
type probe struct {
	Hundertstel int
	Zulaessig   bool
}

// schwelle ist die Hälfte der Kerne in Hundertsteln der Last.
func schwelle(kerne int) int { return 50 * kerne }

// leseLast liest die erste Zahl des Inhalts als Dezimalzahl mit höchstens zwei Nachkommastellen in Hundertsteln.
func leseLast(inhalt []byte) (int, bool) {
	felder := bytes.Fields(inhalt)
	if len(felder) == 0 {
		return 0, false
	}
	ganz, rest, punkt := strings.Cut(string(felder[0]), ".")
	if punkt && (rest == "" || len(rest) > 2) {
		return 0, false
	}
	if ganz == "" || strings.Trim(ganz+rest, "0123456789") != "" {
		return 0, false
	}
	n, err := strconv.Atoi(ganz + (rest + "00")[:2])
	return n, err == nil
}

// nimmProbe liest die Quelle einmal.
func nimmProbe(q lastQuelle, kerne int) probe {
	inhalt, err := q.Lies()
	if n, ok := leseLast(inhalt); err == nil && ok {
		return probe{Hundertstel: n, Zulaessig: n <= schwelle(kerne)}
	}
	return probe{Hundertstel: schwelle(kerne) + 1}
}

// vorAufbau nimmt die Probe vor dem Aufbau; über der Schwelle läuft aufbau nicht und der Fehler ist errLast (SPEC-216 Nr. 11).
func vorAufbau(ctx context.Context, q lastQuelle, kerne int, aufbau func(context.Context) error) error {
	if !nimmProbe(q, kerne).Zulaessig {
		return errLast
	}
	return aufbau(ctx)
}

// lastProben sind die Proben eines Laufs mit ihren Zeitpunkten seit dem Start.
type lastProben struct {
	q      lastQuelle
	kerne  int
	uhr    Uhr
	start  time.Time
	cancel context.CancelFunc
	fertig chan struct{}

	mu     sync.Mutex
	proben []probe
	zeiten []time.Duration
}

// lastErgebnis ist die Zeile `Last` eines Laufs: Zahl der Proben, kleinste und größte in Hundertsteln, die Schwelle, und ob alle zulässig sind.
type lastErgebnis struct {
	Proben    int
	Min       int
	Max       int
	Schwelle  int
	Zulaessig bool
}

func (p *lastProben) nimm(zeit time.Time) {
	pr := nimmProbe(p.q, p.kerne)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proben = append(p.proben, pr)
	p.zeiten = append(p.zeiten, zeit.Sub(p.start))
}

// starteProben nimmt die Probe vor dem Warm-up und startet in einer eigenen Goroutine je probenIntervall der Uhr eine weitere, bis dauer
// seit dem Start vergangen ist (Gleichheit eingeschlossen), und kehrt zurück, ohne auf diese zu warten (SPEC-216 Nr. 11).
func starteProben(ctx context.Context, uhr Uhr, q lastQuelle, kerne int, dauer time.Duration) *lastProben {
	ctx, cancel := context.WithCancel(ctx)
	p := &lastProben{q: q, kerne: kerne, uhr: uhr, start: uhr.Jetzt(), cancel: cancel, fertig: make(chan struct{})}
	p.nimm(p.start)
	tick, stop := uhr.Ticker(probenIntervall)
	go func() {
		defer close(p.fertig)
		defer stop()
		for {
			select {
			case t := <-tick:
				p.nimm(t)
				if t.Sub(p.start) >= dauer {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return p
}

// beende beendet die Proben, nimmt die Probe nach dem Nachlauf und fasst alle zusammen.
func (p *lastProben) beende() lastErgebnis {
	p.cancel()
	<-p.fertig
	p.nimm(p.uhr.Jetzt())
	p.mu.Lock()
	defer p.mu.Unlock()
	e := lastErgebnis{Proben: len(p.proben), Schwelle: schwelle(p.kerne), Min: p.proben[0].Hundertstel, Zulaessig: true}
	for _, pr := range p.proben {
		e.Min, e.Max = min(e.Min, pr.Hundertstel), max(e.Max, pr.Hundertstel)
		e.Zulaessig = e.Zulaessig && pr.Zulaessig
	}
	return e
}

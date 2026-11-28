package batchcodec_test

import (
	"context"
	"errors"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/batchcodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func beendet() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TestBatchProgrammfehler hält Zeile 1: jede Kombination der Verstöße ctx nil, src nil, grenze unter 1 mit beiden Zuständen des
// Kontexts und jedem Wert aus {0, -1, MinInt64}, mit gültigem und ungültigem Content-Type: ErrUngueltig ohne BatchFehler und ohne
// Code, src ungelesen.
func TestBatchProgrammfehler(t *testing.T) {
	n := 0
	for verstoss := 1; verstoss < 8; verstoss++ {
		ctxNil, srcNil, grenzeSchlecht := verstoss&1 != 0, verstoss&2 != 0, verstoss&4 != 0
		grenzen := []int64{100}
		if grenzeSchlecht {
			grenzen = []int64{0, -1, math.MinInt64}
		}
		kontexte := []context.Context{context.Background(), beendet()}
		if ctxNil {
			var kein context.Context
			kontexte = []context.Context{kein}
		}
		for _, ctx := range kontexte {
			for _, grenze := range grenzen {
				for _, ct := range [][]string{{ctBatch}, {"text/plain"}, nil} {
					q := neueQuelle(huelle())
					var src io.Reader = q
					if srcNil {
						src = nil
					}
					got, err := batchcodec.NewLeser().LeseBatch(ctx, src, ct, grenze)
					var bf *model.BatchFehler
					var mitCode interface{ Code() model.Code }
					if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &bf) || errors.As(err, &mitCode) || !reflect.DeepEqual(got, model.Batch{}) {
						t.Errorf("Verstoß %d, grenze %d: %v", verstoss, grenze, err)
					}
					if q.reads != 0 || len(q.andere) != 0 {
						t.Errorf("Verstoß %d: src berührt: %d Reads %v", verstoss, q.reads, q.andere)
					}
					n++
				}
			}
		}
	}
	// ctx nil: 4 Kombinationen mit 1 Kontext (davon 2 mit 3 Grenzen), sonst 3 Kombinationen mit 2 Kontexten; je 3 Content-Types.
	if n != ((1+1+3+3)+(2+6+6))*3 {
		t.Errorf("%d Fälle", n)
	}
	got, err := leseMit(strings.NewReader(""), []string{ctBatch}, 1)
	wantFehler(t, "grenze 1 ist kein Programmfehler", got, err, model.BatchRahmen, "")
}

// ctxMethoden sind die Methoden eines Kontexts, den ein anderer innen trägt.
type ctxMethoden interface {
	Deadline() (time.Time, bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

// schaltbar ist ein Kontext, der erst auf Zuruf beendet ist: ist an gesetzt oder wird Err() zum abN-ten Mal gefragt, meldet er den
// Fehler des inneren, schon beendeten Kontexts, davor nil. context.Cause liefert die Ursache des inneren.
type schaltbar struct {
	innen       ctxMethoden
	an          bool
	fragen, abN int
}

func (c *schaltbar) Deadline() (time.Time, bool) { return c.innen.Deadline() }

func (c *schaltbar) Value(key any) any { return c.innen.Value(key) }

func (c *schaltbar) Err() error {
	c.fragen++
	if c.abN > 0 && c.fragen >= c.abN {
		c.an = true
	}
	if !c.an {
		return nil
	}
	return c.innen.Err()
}

func (c *schaltbar) Done() <-chan struct{} {
	if !c.an {
		return nil
	}
	return c.innen.Done()
}

// kontextArt ist eine Art, in der ein Kontext endet, mit dem Fehler, den Err() dann liefert.
type kontextArt struct {
	name string
	ctx  ctxMethoden
	soll error
}

// kontextArten sind Abbruch und abgelaufene Frist, je ohne und mit eigener Ursache.
func kontextArten(t *testing.T) []kontextArt {
	t.Helper()
	ursache := errors.New("Ursache")
	ctxUrsache, cancelUrsache := context.WithCancelCause(context.Background())
	cancelUrsache(ursache)
	ctxFrist, cancelFrist := context.WithDeadline(context.Background(), time.Unix(0, 0))
	t.Cleanup(cancelFrist)
	ctxFristUrsache, cancelFristUrsache := context.WithDeadlineCause(context.Background(), time.Unix(0, 0), ursache)
	t.Cleanup(cancelFristUrsache)
	return []kontextArt{
		{"Abbruch", beendet(), context.Canceled},
		{"Abbruch mit Ursache", ctxUrsache, context.Canceled},
		{"Frist", ctxFrist, context.DeadlineExceeded},
		{"Frist mit Ursache", ctxFristUrsache, context.DeadlineExceeded},
	}
}

// TestBatchAbbruch hält Zeile 2: der Fehler des Kontexts unverändert (==) und ohne Code an jeder Prüfstelle (vor dem Content-Type,
// vor dem ersten Read, vor jedem weiteren und nach dem letzten) für jede Art, in der ein Kontext endet, auch bei ungültigem Inhalt;
// er geht dem Content-Type vor.
func TestBatchAbbruch(t *testing.T) {
	arten := kontextArten(t)
	bodies := []string{huelle(httpTeil("GET x HTTP/1.1")), huelle("kaputt"), "kein Rahmen", ""}
	n, summe := 0, 0
	for _, art := range arten {
		for _, ct := range [][]string{{ctBatch}, {"text/plain"}, {"application/json"}, nil} {
			q := neueQuelle(huelle())
			got, err := batchcodec.NewLeser().LeseBatch(art.ctx, q, ct, 100)
			if err != art.soll || !reflect.DeepEqual(got, model.Batch{}) || q.reads != 0 {
				t.Errorf("%s vor dem Content-Type, %v: %v, %d Reads", art.name, ct, err, q.reads)
			}
		}
		for _, body := range bodies {
			// Ab der zweiten Abfrage beendet: die erste ist die vor dem Content-Type, vor dem ersten Read folgt eine weitere.
			q := neueQuelle(body)
			got, err := batchcodec.NewLeser().LeseBatch(&schaltbar{innen: art.ctx, abN: 2}, q, []string{ctBatch}, grenzeAll)
			if err != art.soll || !reflect.DeepEqual(got, model.Batch{}) || q.reads != 0 {
				t.Errorf("%s vor dem ersten Read, %q: %v, %d Reads", art.name, body, err, q.reads)
			}
			for _, mitEOF := range []bool{false, true} {
				probe := neueQuelle(body)
				probe.hoechstens, probe.eofMitDaten = 7, mitEOF
				if _, err := leseMit(probe, []string{ctBatch}, grenzeAll); errors.Is(err, art.soll) {
					t.Fatalf("ohne Abbruch: %v", err)
				}
				summe += probe.reads
				// k kleiner als probe.reads trifft die Prüfung vor dem nächsten Read, k gleich probe.reads die nach dem letzten.
				for k := 1; k <= probe.reads; k++ {
					ctx := &schaltbar{innen: art.ctx}
					q := neueQuelle(body)
					q.hoechstens, q.eofMitDaten = 7, mitEOF
					q.beim = func(read int) {
						if read == k {
							ctx.an = true
						}
					}
					got, err := batchcodec.NewLeser().LeseBatch(ctx, q, []string{ctBatch}, grenzeAll)
					if err != art.soll || !reflect.DeepEqual(got, model.Batch{}) || q.reads != k {
						t.Errorf("%s %q in Read %d von %d: %v, %d Reads", art.name, body, k, probe.reads, err, q.reads)
					}
					n++
				}
			}
		}
		abbruchRang(t, art)
	}
	if len(arten) != 4 || n != summe || n == 0 {
		t.Errorf("%d Arten, %d Fälle bei %d Reads", len(arten), n, summe)
	}
}

// abbruchRang: das Lesen geht dem Kontext nach dem letzten Lesen vor (SPEC-182 Nr. 5 d, e).
func abbruchRang(t *testing.T, art kontextArt) {
	t.Helper()
	ctx := &schaltbar{innen: art.ctx}
	q := neueQuelle("zehn Bytes")
	q.beim = func(int) { ctx.an = true }
	got, err := batchcodec.NewLeser().LeseBatch(ctx, q, []string{ctBatch}, 3)
	wantFehler(t, art.name+": zu_gross im selben Read", got, err, model.BatchZuGross, "")
	eigener := errors.New("Platte")
	ctx = &schaltbar{innen: art.ctx}
	q = neueQuelle("zehn Bytes")
	q.fehler, q.fehlerBei, q.mitDaten = eigener, 10, true
	q.beim = func(int) { ctx.an = true }
	if _, err := batchcodec.NewLeser().LeseBatch(ctx, q, []string{ctBatch}, grenzeAll); err != eigener {
		t.Errorf("%s: Fehler von src im selben Read: %v", art.name, err)
	}
}

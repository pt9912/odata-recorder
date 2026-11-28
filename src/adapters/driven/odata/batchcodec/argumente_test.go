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

// TestBatchAbbruch hält Zeile 2: der Fehler des Kontexts unverändert (==) und ohne Code, vor dem ersten Read, vor jedem weiteren
// und nach dem letzten, auch bei ungültigem Inhalt; er geht dem Content-Type vor.
func TestBatchAbbruch(t *testing.T) {
	ursache := errors.New("Ursache")
	ctxUrsache, cancelUrsache := context.WithCancelCause(context.Background())
	cancelUrsache(ursache)
	ctxFrist, cancelFrist := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancelFrist()
	soll := []error{context.Canceled, context.Canceled, context.DeadlineExceeded}
	for i, beendeter := range []context.Context{beendet(), ctxUrsache, ctxFrist} {
		for _, ct := range [][]string{{ctBatch}, {"text/plain"}, {"application/json"}, nil} {
			q := neueQuelle(huelle())
			got, err := batchcodec.NewLeser().LeseBatch(beendeter, q, ct, 100)
			if err != soll[i] || !reflect.DeepEqual(got, model.Batch{}) || q.reads != 0 {
				t.Errorf("vor dem ersten Read, %v: %v, %d Reads", ct, err, q.reads)
			}
		}
	}
	for _, body := range []string{huelle(httpTeil("GET x HTTP/1.1")), huelle("kaputt"), "kein Rahmen", ""} {
		for _, mitEOF := range []bool{false, true} {
			probe := neueQuelle(body)
			probe.hoechstens, probe.eofMitDaten = 7, mitEOF
			if _, err := leseMit(probe, []string{ctBatch}, grenzeAll); errors.Is(err, context.Canceled) {
				t.Fatalf("ohne Abbruch: %v", err)
			}
			for k := 1; k <= probe.reads; k++ {
				ctx, cancel := context.WithCancel(context.Background())
				q := neueQuelle(body)
				q.hoechstens, q.eofMitDaten = 7, mitEOF
				q.beim = func(read int) {
					if read == k {
						cancel()
					}
				}
				got, err := batchcodec.NewLeser().LeseBatch(ctx, q, []string{ctBatch}, grenzeAll)
				if err != context.Canceled || !reflect.DeepEqual(got, model.Batch{}) || q.reads != k {
					t.Errorf("%q Abbruch in Read %d von %d: %v, %d Reads", body, k, probe.reads, err, q.reads)
				}
				cancel()
			}
		}
	}
	// Das Lesen geht dem Kontext nach dem letzten Lesen vor (SPEC-182 Nr. 5 d, e).
	ctx, cancel := context.WithCancel(context.Background())
	q := neueQuelle("zehn Bytes")
	q.beim = func(int) { cancel() }
	got, err := batchcodec.NewLeser().LeseBatch(ctx, q, []string{ctBatch}, 3)
	wantFehler(t, "zu_gross im selben Read", got, err, model.BatchZuGross, "")
}

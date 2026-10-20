package clock_test

import (
	"crypto/rand"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/clock"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

func zeit(s string) time.Time {
	z, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return z
}

// folge liefert die Werte der Reihe nach, den letzten dauerhaft.
func folge(werte ...time.Time) func() time.Time {
	i := 0
	return func() time.Time {
		w := werte[min(i, len(werte)-1)]
		i++
		return w
	}
}

func aufrufe(u interface{ Now() time.Time }, n int) []time.Time {
	var r []time.Time
	for range n {
		r = append(r, u.Now())
	}
	return r
}

func gleich(t *testing.T, got, want []time.Time) {
	t.Helper()
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("Aufruf %d: %v, erwartet %v", i, got[i], want[i])
		}
	}
}

func TestVertrag(t *testing.T) {
	t.Run("Systemuhr", func(t *testing.T) { supporttest.UhrVertrag(t, clock.NewSystemuhr()) })
	t.Run("Quelle in +02:00", func(t *testing.T) {
		z := time.FixedZone("plus2", 7200)
		supporttest.UhrVertrag(t, clock.NeuMit(func() time.Time { return time.Now().In(z) }))
	})
	t.Run("ZufallsIDs", func(t *testing.T) { supporttest.IDVertrag(t, clock.NewZufallsIDs()) })
}

func TestSystemuhrOhneRundung(t *testing.T) {
	n := time.Now()
	if n.Nanosecond() == 0 {
		n = n.Add(1)
	}
	got := clock.NeuMit(func() time.Time { return n }).Now()
	if got != got.Round(0) {
		t.Errorf("monotone Ablesung geblieben: %v", got)
	}
	if got.Nanosecond() != n.Nanosecond() || !got.Equal(n) {
		t.Errorf("Nanosekunden verändert: %v, Quelle %v", got, n)
	}
}

func TestSystemuhrRuecksprung(t *testing.T) {
	t0 := zeit("2026-10-20T10:00:00Z")
	u := clock.NeuMit(folge(t0, t0.Add(-time.Hour), t0.Add(-time.Hour), t0.Add(time.Second), t0))
	gleich(t, aufrufe(u, 5), []time.Time{t0, t0, t0, t0.Add(time.Second), t0.Add(time.Second)})
}

func TestSystemuhrNebenlaeufig(t *testing.T) {
	var n atomic.Int64
	basis := zeit("2026-10-20T10:00:00Z")
	u := clock.NeuMit(func() time.Time { return basis.Add(-time.Duration(n.Add(1))) })
	var wg sync.WaitGroup
	for range 64 {
		wg.Go(func() {
			vorher := u.Now()
			for range 1000 {
				jetzt := u.Now()
				if jetzt.Before(vorher) {
					t.Errorf("Now fällt von %v auf %v", vorher, jetzt)
					return
				}
				vorher = jetzt
			}
		})
	}
	wg.Wait()
}

// TestSystemuhrQuelleSerialisiert zählt die gleichzeitig laufenden Aufrufe der Quelle: Now liest
// sie nur unter der Sperre.
func TestSystemuhrQuelleSerialisiert(t *testing.T) {
	var aktiv, hoechst atomic.Int64
	basis := zeit("2026-10-20T10:00:00Z")
	u := clock.NeuMit(func() time.Time {
		h := aktiv.Add(1)
		for {
			alt := hoechst.Load()
			if h <= alt || hoechst.CompareAndSwap(alt, h) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		aktiv.Add(-1)
		return basis
	})
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 20 {
				u.Now()
			}
		})
	}
	wg.Wait()
	if h := hoechst.Load(); h != 1 {
		t.Errorf("Quelle lief mit %d gleichzeitigen Aufrufen, erwartet 1", h)
	}
}

func TestSystemuhrBereichUngeprueft(t *testing.T) {
	t0 := zeit("2026-10-20T10:00:00Z")
	null, jahr10000 := time.Time{}, time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	gleich(t, aufrufe(clock.NeuMit(folge(null)), 1), []time.Time{null})
	jahr0 := time.Date(0, 6, 1, 0, 0, 0, 0, time.UTC)
	gleich(t, aufrufe(clock.NeuMit(folge(jahr0)), 1), []time.Time{jahr0})
	gleich(t, aufrufe(clock.NeuMit(folge(t0, null)), 2), []time.Time{t0, t0})
	gleich(t, aufrufe(clock.NeuMit(folge(t0, jahr10000, t0)), 3), []time.Time{t0, jahr10000, jahr10000})
}

func TestZufallsIDsUnabhaengig(t *testing.T) {
	a, b := clock.NewZufallsIDs(), clock.NewZufallsIDs()
	for name, p := range map[string][2]string{
		"SessionID":      {a.SessionID(), b.SessionID()},
		"ExchangeID":     {a.ExchangeID(), b.ExchangeID()},
		"KorrelationsID": {a.KorrelationsID(), b.KorrelationsID()},
		"WorkspaceID":    {a.WorkspaceID(), b.WorkspaceID()},
		"SnapshotID":     {a.SnapshotID(), b.SnapshotID()},
	} {
		if p[0] == p[1] {
			t.Errorf("%s: zwei Instanzen liefern %q", name, p[0])
		}
	}
}

// zaehlerLeser liefert die Bytes 0x00, 0x01, … aufsteigend.
type zaehlerLeser struct{ n byte }

func (z *zaehlerLeser) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = z.n
		z.n++
	}
	return len(p), nil
}

// TestZufallsIDsLiestCryptoRand ersetzt crypto/rand.Reader; kein t.Parallel, der Reader ist paketweit.
func TestZufallsIDsLiestCryptoRand(t *testing.T) {
	alt := rand.Reader
	t.Cleanup(func() { rand.Reader = alt })
	rand.Reader = &zaehlerLeser{}
	g := clock.NewZufallsIDs()
	ops := []struct {
		praefix string
		ruf     func() string
	}{
		{"ses", g.SessionID}, {"exc", g.ExchangeID}, {"cor", g.KorrelationsID},
		{"wsp", g.WorkspaceID}, {"snp", g.SnapshotID},
	}
	for i, op := range ops {
		var hex string
		for b := range 16 {
			hex += fmt.Sprintf("%02x", i*16+b)
		}
		if got, want := op.ruf(), op.praefix+"-"+hex; got != want {
			t.Errorf("%s: %q, erwartet %q", op.praefix, got, want)
		}
	}
}

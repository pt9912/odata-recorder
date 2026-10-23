package maintenance_test

import (
	"context"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
)

// TestZurueckstellen: bei einer offenen Session entfernt CleanupBlobs nichts, FindOpen kommt vor
// jeder Lesung, und die Session bleibt, wie sie ist (W15).
func TestZurueckstellen(t *testing.T) {
	for _, offene := range [][]string{{"s1"}, {"s1", "s2"}} {
		u := neu(t)
		u.blobs.lege(r("a"))
		for i, id := range offene {
			u.sessions.fuege(id, []string{dienstA, dienstB}[i], t0().Add(-time.Hour), model.SessionOpen)
		}
		u.sessions.fuege("s3", dienstA, t0().Add(-time.Hour), model.SessionComplete)
		c, err := u.cleanup(context.Background())
		if err != nil || c != (inbound.Loeschbericht{Zurueckgestellt: true}) || !slices.Equal(u.p.liste(), []string{"FindOpen"}) {
			t.Errorf("%v: %v %+v, Aufrufe %v", offene, err, c, u.p.liste())
		}
		if !u.blobs.hat(r("a")) || u.st.anzahl("SetStatus") != 0 || u.st.anzahl("Delete") != 0 {
			t.Errorf("%v: Datei, SetStatus oder Delete", offene)
		}
	}

	u := neu(t)
	u.blobs.lege(r("a"))
	u.sessions.fuege("s3", dienstA, t0().Add(-time.Hour), model.SessionComplete)
	c, err := u.cleanup(context.Background())
	soll := []string{"FindOpen", "Blobs.List", "ReferencedBlobs", "ReferencedBlobs", ev("Remove", r("a"))}
	if err != nil || c != (inbound.Loeschbericht{Entfernt: 1}) || !slices.Equal(u.p.liste(), soll) {
		t.Errorf("ohne offene Session: %v %+v, Aufrufe %v", err, c, u.p.liste())
	}
	if st, _ := u.sessions.status("s3"); st != model.SessionComplete || u.st.anzahl("SetStatus") != 0 || u.st.anzahl("Delete") != 0 {
		t.Errorf("ohne offene Session: s3 %s", st)
	}

	v := neu(t)
	v.blobs.lege(r("a"))
	v.st.scheitere("FindOpen", 1, errDisk)
	if _, err := v.cleanup(context.Background()); err == nil || v.st.anzahl("Remove") != 0 {
		t.Errorf("FindOpen scheitert: %v, Remove %d-mal", err, v.st.anzahl("Remove"))
	} else if code, status := codeVon(t, err); code != "ODR-E9001" || status != 500 {
		t.Errorf("FindOpen scheitert: %s, %d", code, status)
	}
}

// TestKarenz: die Karenz schützt, was der Wächter innerhalb von karenz committete, knapp unter der
// Grenze, und nur das; die letzte Zeit gilt (W16).
func TestKarenz(t *testing.T) {
	T := t0()
	karenz := maintenance.StandardKarenz

	// Grenze: Commit bei T - karenz - 1 ns, T - karenz und T - karenz + 1 ns; ohne Commit des Wächters.
	u := neuAb(t, T.Add(-karenz-time.Nanosecond), karenz)
	rc := u.commit("c")
	u.uhr.Advance(time.Nanosecond)
	rb := u.commit("b")
	u.uhr.Advance(time.Nanosecond)
	ra := u.commit("a")
	rd := refVon([]byte("d"))
	u.blobs.lege(rd)
	u.uhr.Advance(karenz - time.Nanosecond)
	if got := u.uhr.Now(); !got.Equal(T) {
		t.Fatalf("Uhr %v, erwartet %v", got, T)
	}
	c, err := u.cleanup(context.Background())
	if err != nil || c.Entfernt != 3 {
		t.Errorf("Grenze: %v %+v", err, c)
	}
	if got := u.blobs.refs(); !slices.Equal(got, []string{ra}) {
		t.Errorf("Grenze: Dateien %v, erwartet nur ra %v (rb %v, rc %v, rd %v entfernt)", got, ra, rb, rc, rd)
	}

	// Der letzte Commit gilt: bei T - 1 h und bei T - 1 ns.
	v := neuAb(t, T.Add(-time.Hour), karenz)
	re := v.commit("e")
	v.uhr.Advance(time.Hour - time.Nanosecond)
	v.commit("e")
	v.uhr.Advance(time.Nanosecond)
	if c, err := v.cleanup(context.Background()); err != nil || c.Entfernt != 0 || !v.blobs.hat(re) {
		t.Errorf("zwei Commits: %v %+v, Datei vorhanden %v", err, c, v.blobs.hat(re))
	}

	// Karenz 1 ns: geschützt ist der Commit des gleichen Augenblicks.
	w := neuAb(t, T, time.Nanosecond)
	w.commit("f")
	first, err := w.cleanup(context.Background())
	w.uhr.Advance(time.Nanosecond)
	second, err2 := w.cleanup(context.Background())
	if err != nil || err2 != nil || first.Entfernt != 0 || second.Entfernt != 1 {
		t.Errorf("Karenz 1 ns: %v %+v, %v %+v", err, first, err2, second)
	}

	// Die größte Karenz schützt, was der Wächter sah.
	x := neuAb(t, T, math.MaxInt64)
	rg := x.commit("g")
	x.uhr.Advance(1000 * tag)
	if c, err := x.cleanup(context.Background()); err != nil || c.Entfernt != 0 || !x.blobs.hat(rg) {
		t.Errorf("größte Karenz: %v %+v", err, c)
	}
}

// ergebnis ist der Rückgabewert eines CleanupBlobs in einer Goroutine.
type ergebnis struct {
	b   inbound.Loeschbericht
	err error
}

func (u *welt) cleanupAsync() <-chan ergebnis {
	ch := make(chan ergebnis, 1)
	go func() {
		b, err := u.w.CleanupBlobs(context.Background())
		ch <- ergebnis{b, err}
	}()
	return ch
}

// sperreHalten prüft, dass die Lesersperre des Wächters nicht zu nehmen ist.
func sperreHalten(t *testing.T, u *welt, wo string) {
	t.Helper()
	if u.waechter.TryRLock() {
		u.waechter.RUnlock()
		t.Errorf("%s: die Lesersperre des Wächters ist frei", wo)
	}
}

// TestSperreAbgleich: der Abgleich hält die Sperre des Wächters von vor FindOpen bis nach dem
// letzten Remove, und ein Commit gleicher Bytes gegen ein Remove endet in einer bestehenden
// Datei (W24).
func TestSperreAbgleich(t *testing.T) {
	for _, f := range []struct {
		name string
		op   string
		n    int
	}{{"FindOpen", "FindOpen", 1}, {"zweite Lesung", "ReferencedBlobs", 2}, {"Remove", "Remove", 1}} {
		u := neu(t)
		u.blobs.lege(r("a"))
		h := u.st.halte(f.op, f.n)
		ch := u.cleanupAsync()
		warte(t, h.erreicht, f.name)
		sperreHalten(t, u, f.name)
		h.gib()
		if e := <-ch; e.err != nil {
			t.Errorf("%s: %v", f.name, e.err)
		}
	}

	// Wirkung: ein Commit gleicher Bytes, während Remove der Datei hält.
	u := neu(t)
	rk := refVon([]byte("K"))
	u.blobs.lege(rk)
	h := u.st.halte("Remove", 1)
	called := make(chan struct{})
	u.st.bei("Commit", 1, func() { close(called) })
	ch := u.cleanupAsync()
	warte(t, h.erreicht, "Remove")
	committed := make(chan string, 1)
	go func() {
		w, err := u.waechter.Begin(context.Background(), 16)
		if err != nil {
			t.Error(err)
			committed <- ""
			return
		}
		if _, err := w.Write([]byte("K")); err != nil {
			t.Error(err)
		}
		ref, _, err := w.Commit()
		if err != nil {
			t.Error(err)
		}
		committed <- ref
	}()
	select {
	case <-called:
		t.Error("Commit gleicher Bytes lief, während Remove hielt")
	case <-time.After(100 * time.Millisecond):
	}
	h.gib()
	if e := <-ch; e.err != nil || e.b.Entfernt != 1 {
		t.Errorf("Remove: %v %+v", e.err, e.b)
	}
	if ref := <-committed; ref != rk {
		t.Errorf("Commit lieferte %q, erwartet %q", ref, rk)
	}
	if err := u.waechter.Verify(context.Background(), rk); err != nil {
		t.Errorf("Datei nach Commit und Remove: %v", err)
	}
}

// setzeHoechst hebt a auf n, wenn n größer ist.
func setzeHoechst(a *atomic.Int32, n int32) {
	for {
		m := a.Load()
		if n <= m || a.CompareAndSwap(m, n) {
			return
		}
	}
}

// TestAbgleicheNacheinander: gleichzeitige Abgleiche laufen nacheinander (W26).
func TestAbgleicheNacheinander(t *testing.T) {
	u := neu(t)
	u.blobs.lege(r("a"), r("b"), r("c"))
	var gleichzeitig, hoechst, summe atomic.Int32
	u.st.bei("Blobs.List", 0, func() {
		setzeHoechst(&hoechst, gleichzeitig.Add(1))
		time.Sleep(time.Millisecond)
		gleichzeitig.Add(-1)
	})
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			b, err := u.w.CleanupBlobs(context.Background())
			if err != nil {
				t.Error(err)
			}
			summe.Add(int32(b.Entfernt))
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		b, err := u.w.Startup(context.Background())
		if err != nil {
			t.Error(err)
		}
		summe.Add(int32(b.BlobsEntfernt))
	}()
	close(start)
	wg.Wait()
	if hoechst.Load() != 1 || summe.Load() != 3 {
		t.Errorf("höchstens %d gleichzeitige List, %d Entfernte; erwartet 1 und 3", hoechst.Load(), summe.Load())
	}
}

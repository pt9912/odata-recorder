package maintenance_test

import (
	"context"
	"errors"
	"fmt"
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

// sperreFrei prüft, dass nach der Rückkehr einer Operation die Schreiber- und die Lesersperre des
// Wächters zu nehmen sind: eine behaltene Schreibersperre macht beide falsch, eine behaltene
// Lesersperre die Schreibersperre.
func sperreFrei(t *testing.T, u *welt, wo string) {
	t.Helper()
	if u.waechter.TryLock() {
		u.waechter.Unlock()
	} else {
		t.Errorf("%s: die Schreibersperre des Wächters ist nach der Rückkehr nicht zu nehmen", wo)
	}
	if u.waechter.TryRLock() {
		u.waechter.RUnlock()
	} else {
		t.Errorf("%s: die Lesersperre des Wächters ist nach der Rückkehr nicht zu nehmen", wo)
	}
}

// TestSperreFrei: jeder Rückkehrpfad einer Operation, die die Sperre des Wächters nimmt, gibt sie
// frei; die Aufrufe laufen nacheinander und ohne Haltepunkt (W28). Je Fall steht der Ausgang, damit
// der Fall den Pfad nimmt, den sein Name nennt.
func TestSperreFrei(t *testing.T) {
	cleanup := func(u *welt) (string, error) {
		b, err := u.w.CleanupBlobs(context.Background())
		return fmt.Sprintf("entfernt %d, zurückgestellt %v", b.Entfernt, b.Zurueckgestellt), err
	}
	startup := func(u *welt) (string, error) {
		b, err := u.w.Startup(context.Background())
		return fmt.Sprintf("entfernt %d", b.BlobsEntfernt), err
	}
	commit := func(u *welt) (string, error) {
		w, err := u.waechter.Begin(context.Background(), 16)
		if err != nil {
			return "Begin", err
		}
		if _, err := w.Write([]byte("K")); err != nil {
			return "Write", err
		}
		ref, _, err := w.Commit()
		return fmt.Sprintf("Referenz von K %v", ref == refVon([]byte("K"))), err
	}
	mitBestand := func(u *welt) { bestandBereinigung(u) }
	faelle := []struct {
		name    string
		setze   func(u *welt)
		ruf     func(u *welt) (string, error)
		ausgang string
		fehler  error
	}{
		{"CleanupBlobs gelungen, zwei Kandidaten", mitBestand, cleanup, "entfernt 2, zurückgestellt false", nil},
		{"CleanupBlobs zurückgestellt", func(u *welt) {
			mitBestand(u)
			u.sessions.fuege("s1", dienstA, t0().Add(-time.Hour), model.SessionOpen)
		}, cleanup, "entfernt 0, zurückgestellt true", nil},
		{"CleanupBlobs, FindOpen scheitert", func(u *welt) {
			mitBestand(u)
			u.st.scheitere("FindOpen", 1, errDisk)
		}, cleanup, "entfernt 0, zurückgestellt false", errDisk},
		{"CleanupBlobs, erste Lesung scheitert", func(u *welt) {
			mitBestand(u)
			u.st.scheitere("ReferencedBlobs", 1, errDisk)
		}, cleanup, "entfernt 0, zurückgestellt false", errDisk},
		{"CleanupBlobs, Remove des zweiten Kandidaten scheitert", func(u *welt) {
			mitBestand(u)
			u.st.scheitere("Remove", 2, errDisk)
		}, cleanup, "entfernt 1, zurückgestellt false", errDisk},
		{"Startup gelungen", mitBestand, startup, "entfernt 2", nil},
		{"Startup, Verify der Prüfstufe scheitert", func(u *welt) {
			mitBestand(u)
			u.st.scheitere("Verify", 1, errDisk)
		}, startup, "entfernt 0", errDisk},
		{"Startup, Lesung der Prüfstufe scheitert", func(u *welt) {
			mitBestand(u)
			u.st.scheitere("ReferencedBlobs", 1, errDisk)
		}, startup, "entfernt 0", errDisk},
		{"Startup, Remove des zweiten Kandidaten scheitert", func(u *welt) {
			mitBestand(u)
			u.st.scheitere("Remove", 2, errDisk)
		}, startup, "entfernt 1", errDisk},
		{"Commit gelungen", func(*welt) {}, commit, "Referenz von K true", nil},
		{"Commit, der Writer darunter scheitert", func(u *welt) {
			u.st.scheitere("Commit", 1, errDisk)
		}, commit, "Referenz von K false", errDisk},
	}
	for _, f := range faelle {
		u := neu(t)
		f.setze(u)
		ausgang, err := f.ruf(u)
		if ausgang != f.ausgang || (f.fehler == nil && err != nil) || (f.fehler != nil && !errors.Is(err, f.fehler)) {
			t.Errorf("%s: Ausgang %q (%v), erwartet %q (%v)", f.name, ausgang, err, f.ausgang, f.fehler)
		}
		sperreFrei(t, u, f.name)
	}
}

// startupAsync ruft Startup in einer Goroutine.
func (u *welt) startupAsync() <-chan error {
	ch := make(chan error, 1)
	go func() {
		_, err := u.w.Startup(context.Background())
		ch <- err
	}()
	return ch
}

// TestSperreStartup: die Stufe 4 von Startup hält die Sperre des Wächters von ihrem Beginn bis zu
// ihrem Ende, im Prüfen und im Abgleich: in jedem Halt ist die Lesersperre nicht zu nehmen (W29).
func TestSperreStartup(t *testing.T) {
	for _, f := range []struct {
		name string
		op   string
	}{
		{"Lesung der Prüfstufe", "ReferencedBlobs"},
		{"Verify der ersten Referenz", "Verify"},
		{"Blobs.List", "Blobs.List"},
		{"Remove des ersten Kandidaten", "Remove"},
	} {
		u := neu(t)
		u.speicher.fest = []string{r("v1")}
		u.blobs.lege(r("v1"), r("u1"))
		h := u.st.halte(f.op, 1)
		t.Cleanup(h.gib)
		ch := u.startupAsync()
		warte(t, h.erreicht, f.name)
		sperreHalten(t, u, f.name)
		h.gib()
		if err := <-ch; err != nil {
			t.Errorf("%s: %v", f.name, err)
		}
	}
}

package session_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestBereinigungNachCommit: CleanupBlobs genau einmal, erst nach der Rückkehr von Sessions.Delete
// mit nil, nie nach einem Fehler; Startup nie (V12). Die Fälle sind deterministisch: der Haltepunkt
// ordnet mit Kanälen.
func TestBereinigungNachCommit(t *testing.T) {
	// (a) gelungenes Delete: genau eine Bereinigung.
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	if _, err := u.del(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	if n := u.st.anzahl("CleanupBlobs"); n != 1 {
		t.Errorf("(a) Bereinigung %d-mal", n)
	}
	if n := u.st.anzahl("Startup"); n != 0 {
		t.Errorf("(d) Startup %d-mal", n)
	}

	// (b) Im Haltepunkt von Sessions.Delete ist noch keine Bereinigung gerufen.
	u = neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	h := u.st.halte("Sessions.Delete", 1)
	defer h.gib()
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		_, _ = u.v.Delete(context.Background(), "s1")
	}()
	warte(t, h.erreicht, "Sessions.Delete betreten")
	if n := u.st.anzahl("CleanupBlobs"); n != 0 {
		t.Errorf("(b) Bereinigung %d-mal vor der Rückkehr von Sessions.Delete", n)
	}
	h.gib()
	warte(t, fertig, "Rückkehr von Delete")
	if n := u.st.anzahl("CleanupBlobs"); n != 1 {
		t.Errorf("(b) Bereinigung %d-mal nach der Rückkehr", n)
	}

	// (c) Nach einem Fehler des Löschens ruft der Dienst sie nie.
	fehlerfaelle := map[string]func(*welt){
		"Get unbekannt":     func(u *welt) { u.st.scheitere("Sessions.Get", 1, verpackt(outbound.ErrNichtGefunden)) },
		"offen":             func(u *welt) { u.sessions.fuege("s1", dienstA, t0(), model.SessionOpen) },
		"referenziert":      func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrReferenziert)) },
		"Delete unbekannt":  func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrNichtGefunden)) },
		"Get beschädigt":    func(u *welt) { u.st.scheitere("Sessions.Get", 1, verpackt(outbound.ErrBeschaedigt)) },
		"Delete beschädigt": func(u *welt) { u.st.scheitere("Sessions.Delete", 1, verpackt(outbound.ErrBeschaedigt)) },
		"Get Platte":        func(u *welt) { u.st.scheitere("Sessions.Get", 1, errDisk) },
		"Delete Platte":     func(u *welt) { u.st.scheitere("Sessions.Delete", 1, errDisk) },
		"Get Canceled":      func(u *welt) { u.st.scheitere("Sessions.Get", 1, context.Canceled) },
		"Delete Canceled":   func(u *welt) { u.st.scheitere("Sessions.Delete", 1, context.Canceled) },
		"Delete Deadline":   func(u *welt) { u.st.scheitere("Sessions.Delete", 1, context.DeadlineExceeded) },
	}
	for name, setup := range fehlerfaelle {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		setup(u)
		if _, err := u.v.Delete(context.Background(), "s1"); err == nil {
			t.Errorf("(c) %s: kein Fehler", name)
		}
		if n := u.st.anzahl("CleanupBlobs"); n != 0 {
			t.Errorf("(c) %s: Bereinigung %d-mal", name, n)
		}
	}
}

// TestBereinigungKontext: CleanupBlobs erhält den ctx des Aufrufs unverändert und läuft synchron
// (V13). Das Warten auf die Haltepunkte ordnet synctest.Wait, ohne Dauer.
func TestBereinigungKontext(t *testing.T) {
	grund := context.WithValue(context.Background(), ctxSchluessel{}, "w1")
	neuMitSession := func(t *testing.T) *welt {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		return u
	}

	// (a) lebender ctx ohne Frist: Wert da, nicht beendet, keine Frist.
	u := neuMitSession(t)
	if _, err := u.del(grund, "s1"); err != nil {
		t.Fatal(err)
	}
	if g := u.wartung.gesehen(); len(g) != 1 || g[0].wert != "w1" || g[0].fehler != nil || g[0].hatFrist {
		t.Errorf("(a) %+v", g)
	}

	// (b) ctx mit Frist: dieselbe Frist.
	mitFrist, abbruch := context.WithTimeout(grund, time.Hour)
	defer abbruch()
	frist, _ := mitFrist.Deadline()
	u = neuMitSession(t)
	if _, err := u.del(mitFrist, "s1"); err != nil {
		t.Fatal(err)
	}
	if g := u.wartung.gesehen(); len(g) != 1 || g[0].wert != "w1" || !g[0].hatFrist || !g[0].frist.Equal(frist) {
		t.Errorf("(b) %+v, erwartet Frist %v", g, frist)
	}

	// (c) Sessions.Delete beendet den ctx nach seinem Erfolg: die Bereinigung läuft trotzdem, mit
	// dem beendeten ctx, und Delete kehrt mit nil zurück.
	endet, beende := context.WithCancel(grund)
	defer beende()
	u = neuMitSession(t)
	u.sessions.nachDelete = beende
	b, err := u.del(endet, "s1")
	if g := u.wartung.gesehen(); err != nil || b.Entfernt != 2 || len(g) != 1 || g[0].wert != "w1" || g[0].fehler == nil {
		t.Errorf("(c) %+v, %+v, %v", g, b, err)
	}

	// (d) Delete kehrt nicht zurück, solange die Bereinigung läuft.
	synctest.Test(t, func(t *testing.T) {
		u := neuMitSession(t)
		h := u.st.halte("CleanupBlobs", 1)
		fertig := make(chan struct{})
		go func() {
			defer close(fertig)
			_, _ = u.v.Delete(grund, "s1")
		}()
		<-h.erreicht
		synctest.Wait()
		select {
		case <-fertig:
			t.Error("(d) Delete kehrte vor dem Ende der Bereinigung zurück")
		default:
		}
		h.gib()
		<-fertig
	})
}

// parallel startet n Aufrufe von f und wartet auf alle.
func parallel(n int, f func(i int)) {
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f(i)
		}()
	}
	wg.Wait()
}

// TestAnlegenGleichzeitig: zwei Create derselben id; Sessions.Create hält beide, bis beide es
// betreten haben, und arbeitet dann atomar: genau eines gelingt, das andere ist ODR-E7010, der
// Dienst liest vorher nichts (V17).
func TestAnlegenGleichzeitig(t *testing.T) {
	u := neu(t)
	b := neueBarriere(2)
	u.st.bei("Sessions.Create", 0, b.an)
	fehler := make([]error, 2)
	parallel(2, func(i int) { _, fehler[i] = u.v.Create(context.Background(), dienstA, "x1") })
	if b.lief() {
		t.Fatal("die Barriere lief ab: ein Aufruf erreichte Sessions.Create nicht")
	}
	ok, doppelt := 0, 0
	for _, err := range fehler {
		if err == nil {
			ok++
		} else if c, _ := codeVon(t, err); c == "ODR-E7010" {
			doppelt++
		}
	}
	if ok != 1 || doppelt != 1 {
		t.Errorf("gelungen %d, ODR-E7010 %d; Fehler %v", ok, doppelt, fehler)
	}
	if _, da := u.sessions.hat("x1"); !da || u.sessions.anzahl() != 1 {
		t.Errorf("Sessions im Fake: %d", u.sessions.anzahl())
	}
	for _, e := range u.p.liste() {
		if e == "Sessions.Get(x1)" {
			t.Error("der Dienst las vor dem Anlegen")
		}
	}
	if n := u.st.anzahl("Sessions.Create"); n != 2 {
		t.Errorf("Sessions.Create %d-mal", n)
	}
}

// TestLoeschenGleichzeitig: zwei Delete derselben Session; Sessions.Delete hält beide, bis beide
// es betreten haben (beide haben Get passiert): genau eines gelingt und ruft die Bereinigung, das
// andere ist ODR-E7003 ohne Bereinigung (V18).
func TestLoeschenGleichzeitig(t *testing.T) {
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	b := neueBarriere(2)
	u.st.bei("Sessions.Delete", 0, b.an)
	fehler := make([]error, 2)
	parallel(2, func(i int) { _, fehler[i] = u.v.Delete(context.Background(), "s1") })
	if b.lief() {
		t.Fatal("die Barriere lief ab: ein Aufruf erreichte Sessions.Delete nicht")
	}
	ok, unbekannt := 0, 0
	for _, err := range fehler {
		if err == nil {
			ok++
		} else if c, _ := codeVon(t, err); c == "ODR-E7003" {
			unbekannt++
		}
	}
	if ok != 1 || unbekannt != 1 {
		t.Errorf("gelungen %d, ODR-E7003 %d; Fehler %v", ok, unbekannt, fehler)
	}
	if n := u.st.anzahl("CleanupBlobs"); n != 1 {
		t.Errorf("Bereinigung %d-mal", n)
	}
	if g, d := u.st.anzahl("Sessions.Get"), u.st.anzahl("Sessions.Delete"); g != 2 || d != 2 {
		t.Errorf("Get %d-mal, Delete %d-mal", g, d)
	}
}

// TestGleichzeitig: der Dienst hält nach New keinen veränderlichen Zustand; 16 Goroutinen rufen je
// 50-mal Create, Get, List und Delete (V19). Unter -race meldet der Detektor einen Zustand, den
// zwei Goroutinen berühren.
func TestGleichzeitig(t *testing.T) {
	const goroutinen, runden = 16, 50
	u := neu(t)
	u.dienst(dienstA)
	var mu sync.Mutex
	vergeben := map[string]bool{}
	var geloescht atomic.Int64
	ctx := context.Background()
	parallel(goroutinen, func(int) {
		for range runden {
			s, err := u.v.Create(ctx, dienstA, "")
			if err != nil {
				t.Errorf("Create: %v", err)
				continue
			}
			mu.Lock()
			if vergeben[s.ID] {
				t.Errorf("Kennung %s doppelt vergeben", s.ID)
			}
			vergeben[s.ID] = true
			mu.Unlock()
			if _, err := u.v.Get(ctx, s.ID); err != nil {
				t.Errorf("Get: %v", err)
			}
			if _, err := u.v.List(ctx); err != nil {
				t.Errorf("List: %v", err)
			}
			if err := u.sessions.SetStatus(ctx, s.ID, model.SessionComplete); err != nil {
				t.Errorf("SetStatus: %v", err)
			}
			if b, err := u.v.Delete(ctx, s.ID); err != nil || b != (inbound.Loeschbericht{Entfernt: 2}) {
				t.Errorf("Delete: %+v, %v", b, err)
			} else {
				geloescht.Add(1)
			}
		}
	})
	if n := len(vergeben); n != goroutinen*runden {
		t.Errorf("%d Kennungen vergeben, erwartet %d", n, goroutinen*runden)
	}
	if g, c := geloescht.Load(), int64(u.st.anzahl("CleanupBlobs")); g != goroutinen*runden || c != g {
		t.Errorf("gelöscht %d, Bereinigungen %d", g, c)
	}
}

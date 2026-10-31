package replay_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// res ist das Ergebnis eines Replay in einer Goroutine.
type res struct {
	seq int
	err error
}

func lauf(u *umg, a inbound.Anfrage) res {
	resp, err := u.replayer().Replay(context.Background(), a)
	if err != nil {
		return res{err: err}
	}
	s := seqVon(u.t, resp)
	_ = resp.Body.Close()
	return res{seq: s}
}

// warte wartet, bis f wahr ist; die Reihenfolge der Ereignisse bestimmt der Aufbau, nicht diese Frist.
func warte(t *testing.T, was string, f func() bool) {
	t.Helper()
	ende := time.Now().Add(10 * time.Second)
	for !f() {
		if time.Now().After(ende) {
			t.Fatalf("Zeitüberschreitung: %s", was)
		}
		time.Sleep(200 * time.Microsecond)
	}
}

// halte reiht eine Transaktion ein, die den Kopf der Schlange hält, bis frei gerufen wird.
func (u *umg) halte() (frei func(), fertig <-chan error) {
	gate := make(chan struct{})
	ch := u.st.eindringling(context.Background(), func(outbound.WorkspaceStores) error { <-gate; return nil })
	return func() { close(gate) }, ch
}

// dreimalA legt n replayfähige Exchanges unter dem Key A an.
func dreimalA(u *umg, n int) {
	var es []model.Exchange
	for i := 1; i <= n; i++ {
		es = append(es, u.austausch(i, u.keyA(), true))
	}
	u.lege(es...)
}

// TestGleichzeitigGleicherCursor hält Z46: Menge 3 (n = 2, 3, 4); jeder Aufruf erhält einen anderen Exchange,
// consumed enthält alle, position gleich n.
func TestGleichzeitigGleicherCursor(t *testing.T) {
	for n := 2; n <= 4; n++ {
		u := neu(t)
		dreimalA(u, n)
		frei, kopf := u.halte()
		ergebnisse := make(chan res, n)
		for i := 0; i < n; i++ {
			go func() { ergebnisse <- lauf(u, anf()) }()
			warte(t, "Schlange", func() bool { return u.st.wartend() == i+2 })
		}
		frei()
		<-kopf
		var seqs []int
		for i := 0; i < n; i++ {
			r := <-ergebnisse
			if r.err != nil {
				t.Errorf("n=%d: %v", n, r.err)
			}
			seqs = append(seqs, r.seq)
		}
		slices.Sort(seqs)
		want := make([]int, n)
		for i := range want {
			want[i] = i + 1
		}
		c, _ := u.cursorStand()
		consumed := slices.Clone(c.Consumed)
		slices.Sort(consumed)
		if !reflect.DeepEqual(seqs, want) || c.Position != int64(n) || len(consumed) != n {
			t.Errorf("n=%d: Exchanges %v, Cursor %+v", n, seqs, c)
		}
	}
}

// verschraenkungen liefert die sechs Anordnungen von zwei Requests je Session (C(4,2)).
func verschraenkungen() [][]string {
	var r [][]string
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			o := []string{"rep-b", "rep-b", "rep-b", "rep-b"}
			o[i], o[j] = "rep-a", "rep-a"
			r = append(r, o)
		}
	}
	return r
}

// TestZweiReplaySessions hält Z47: Menge 6; zwei Replay-Sessions desselben Workspace und derselben
// Recording-Session erhalten jede 1, dann 3.
func TestZweiReplaySessions(t *testing.T) {
	if len(verschraenkungen()) != 6 {
		t.Fatal("keine sechs Verschränkungen")
	}
	for _, o := range verschraenkungen() {
		u := neu(t)
		u.lege(u.austausch(1, u.keyA(), true), u.austausch(2, u.keyB(), true), u.austausch(3, u.keyA(), true))
		frei, kopf := u.halte()
		kanaele := make([]chan res, len(o))
		for i, rep := range o {
			a := anf()
			a.Header["X-Recorder-Replay-Session"] = []string{rep}
			kanaele[i] = make(chan res, 1)
			go func() { kanaele[i] <- lauf(u, a) }()
			warte(t, "Schlange", func() bool { return u.st.wartend() == i+2 })
		}
		frei()
		<-kopf
		folgen := map[string][]int{}
		for i, rep := range o {
			r := <-kanaele[i]
			if r.err != nil {
				t.Errorf("%v: %v", o, r.err)
			}
			folgen[rep] = append(folgen[rep], r.seq)
		}
		if !reflect.DeepEqual(folgen, map[string][]int{"rep-a": {1, 3}, "rep-b": {1, 3}}) {
			t.Errorf("%v: %v", o, folgen)
		}
	}
}

func reset(st outbound.WorkspaceStores) error { return st.Cursors.ResetAll(context.Background(), wsID) }
func loescheWorkspace(st outbound.WorkspaceStores) error {
	return st.Workspaces.Delete(context.Background(), wsID)
}

func (u *umg) loescheSession(outbound.WorkspaceStores) error {
	u.sess.weg(sesID)
	u.st.mu.Lock()
	defer u.st.mu.Unlock()
	for ws, cs := range u.st.z.cur {
		u.st.z.cur[ws] = slices.DeleteFunc(cs, func(c model.ReplayCursor) bool { return c.RecordingSessionID == sesID })
	}
	return nil
}

// mitEindringling reiht beim ersten Aufruf von op nach seiner Rückkehr den Eindringling f ein und liefert
// die Ergebnisse von Replay und Eindringling.
func mitEindringling(u *umg, op string, f func(outbound.WorkspaceStores) error) (res, error) {
	var ch <-chan error
	var einmal atomic.Bool
	u.p.setzeNach(func(o string) {
		if o == op && einmal.CompareAndSwap(false, true) {
			ch = u.st.eindringling(context.Background(), f)
		}
	})
	r := lauf(u, anf())
	if ch == nil {
		u.t.Fatalf("Haltepunkt %s nicht erreicht", op)
	}
	return r, <-ch
}

// vorbedingung1 legt drei Exchanges und einen Cursor mit position 1 an.
func vorbedingung1(u *umg, strat model.ReplayStrategy) {
	dreimalA(u, 3)
	u.cursor(strat, 1)
}

// TestReplayGegenReset hält Z48: Menge 2; das Fortschreiben liegt vor dem Reset (danach position 0) oder nach
// ihm (danach 1, der Replay beginnt bei 0), nie position 2 aus altem Wert plus 1.
func TestReplayGegenReset(t *testing.T) {
	faelle := []struct {
		name     string
		hook     string
		seq      int
		position int64
	}{
		{"Reset nach dem Fortschreiben", "tx Cursors.Get", 2, 0},
		{"Reset vor dem Fortschreiben", "Exchanges.ListBySession", 1, 1},
	}
	for _, f := range faelle {
		u := neu(t)
		vorbedingung1(u, model.StrategySequential)
		r, err := mitEindringling(u, f.hook, reset)
		c, _ := u.cursorStand()
		if r.err != nil || err != nil || r.seq != f.seq || c.Position != f.position || int64(len(c.Consumed)) != f.position {
			t.Errorf("%s: Replay %+v, Reset %v, Cursor %+v", f.name, r, err, c)
		}
	}
}

// TestFensterDerTransaktion hält Z49 bis Z52: der Eindringling (Reset) wird an jedem Fenster nach einer
// Operation eingereiht. Vor der Transaktion (Sessions.Get, ListBySession) liegt der Reset vor dem Replay (position 1),
// in der Transaktion (nach Workspaces.Get, Cursors.Get, Blobs.Open) nach ihm (position 0).
func TestFensterDerTransaktion(t *testing.T) {
	faelle := []struct {
		name     string
		hook     string
		absent   bool
		seq      int
		position int64
	}{
		{"Z49 nach Sessions.Get", "Sessions.Get", false, 1, 1},
		{"Z49 nach ListBySession", "Exchanges.ListBySession", false, 1, 1},
		{"Z50 nach Workspaces.Get", "tx Workspaces.Get", false, 2, 0},
		{"Z51 nach Cursors.Get mit Body", "tx Cursors.Get", false, 2, 0},
		{"Z51 nach Cursors.Get ohne Body", "tx Cursors.Get", true, 2, 0},
		{"Z52 nach Blobs.Open", "Blobs.Open", false, 2, 0},
	}
	for _, f := range faelle {
		u := neu(t)
		vorbedingung1(u, model.StrategySequential)
		if f.absent {
			es := u.ex.liste[sesID]
			for i := range es {
				es[i] = ohneBody(es[i])
			}
		}
		r, err := mitEindringling(u, f.hook, reset)
		c, _ := u.cursorStand()
		if r.err != nil || err != nil || r.seq != f.seq || c.Position != f.position || int64(len(c.Consumed)) != f.position {
			t.Errorf("%s: Replay %+v, Reset %v, Cursor %+v", f.name, r, err, c)
		}
	}
}

// TestFensterDerTransaktionLoeschen hält Z50 mit dem Löschen des Workspace zwischen Workspaces.Get und Cursors.Get.
func TestFensterLoeschen(t *testing.T) {
	u := neu(t)
	vorbedingung1(u, model.StrategySequential)
	r, err := mitEindringling(u, "tx Workspaces.Get", loescheWorkspace)
	z := u.st.stand()
	if r.err != nil || err != nil || r.seq != 2 || len(z.cur[wsID]) != 0 || len(z.ws) != 0 {
		t.Errorf("Replay %+v, Löschen %v, Zustand %+v", r, err, z)
	}
}

// TestReplayGegenLoeschen hält Z53: Menge 4 plus der Haltepunkt nach Workspaces.Get; nie E7003 für den gelöschten
// Workspace, nie ein Cursor ohne Workspace, die gelöschte Session ist E7003 ohne Cursor.
func TestReplayGegenLoeschen(t *testing.T) {
	type fall struct {
		name  string
		hook  string
		wsDel bool
		ok    bool
		code  model.Code
	}
	faelle := []fall{
		{"Workspace vor", "Exchanges.ListBySession", true, false, model.CodeE7001},
		{"Workspace nach", "tx Cursors.Get", true, true, 0},
		{"Session vor", "Exchanges.ListBySession", false, false, model.CodeE7003},
		{"Session nach", "tx Cursors.Get", false, true, 0},
		{"Workspace nach Workspaces.Get", "tx Workspaces.Get", true, true, 0},
	}
	for _, f := range faelle {
		u := neu(t)
		dreimalA(u, 3)
		del := u.loescheSession
		if f.wsDel {
			del = func(st outbound.WorkspaceStores) error { return loescheWorkspace(st) }
		}
		r, err := mitEindringling(u, f.hook, del)
		z := u.st.stand()
		switch {
		case err != nil:
			t.Errorf("%s: Eindringling %v", f.name, err)
		case f.ok && (r.err != nil || r.seq != 1):
			t.Errorf("%s: Replay %+v", f.name, r)
		case !f.ok:
			erwarteCode(t, f.name, r.err, f.code)
		}
		if len(z.cur[wsID]) != 0 {
			t.Errorf("%s: Cursor ohne Workspace oder Session: %+v", f.name, z.cur)
		}
	}
}

// TestNurDieAchtOperationen hält Z57: Menge 155 (31 verbotene Operationen × 5 Läufe: Erfolg je Strategie und
// ein Fehlerlauf); keine der 27 Operationen der Ports und keine der vier Telemetrie-Operationen außer Log.
func TestNurDieAchtOperationen(t *testing.T) {
	erlaubt := []string{"Sessions.Get", "Exchanges.ListBySession", "Transaktion.WithWorkspaceTransaction",
		"tx Workspaces.Get", "tx Cursors.Get", "tx Cursors.Put", "Blobs.Open"}
	var verboten []string
	for _, op := range alleOps() {
		if !slices.Contains(erlaubt, op) {
			verboten = append(verboten, op)
		}
	}
	laeufe := []struct {
		name   string
		strat  model.ReplayStrategy
		fehler bool
	}{
		{"sequential", model.StrategySequential, false}, {"strict", model.StrategyStrict, false},
		{"first", model.StrategyFirst, false}, {"last", model.StrategyLast, false}, {"Fehlerlauf", model.StrategySequential, true},
	}
	n := 0
	for _, l := range laeufe {
		u := neu(t)
		u.k.Strategie = l.strat
		a := erfolgsLauf(u)
		if l.fehler {
			u.p.setzeFehler("tx Cursors.Put", errors.New("put"))
		}
		resp, err := u.replay(context.Background(), a)
		if (err != nil) != l.fehler {
			t.Errorf("%s: Fehler %v", l.name, err)
		} else if err == nil {
			_ = resp.Body.Close()
		}
		for _, op := range verboten {
			n++
			if u.p.anzahl(op) != 0 {
				t.Errorf("%s: %s gerufen", l.name, op)
			}
		}
		for _, op := range telOps()[1:] {
			n++
			if u.tel.Anzahl(op) != 0 {
				t.Errorf("%s: Telemetrie-Operation %d gerufen", l.name, op)
			}
		}
		nurErlaubt(t, l.name, u.p)
	}
	if len(verboten) != 27 || n != 155 {
		t.Errorf("Menge: %d verbotene Ports, %d Prüfungen; erwartet 27 und 155", len(verboten), n)
	}
}

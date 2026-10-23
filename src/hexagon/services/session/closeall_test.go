package session_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// neuMehrere liefert eine Umgebung mit den offenen Sessions o1 bis on, abwechselnd in dienstA und
// dienstB; o1, o3, … tragen einen vollständigen Exchange, o2, o4, … einen unvollständigen.
func neuMehrere(t *testing.T, n int) *welt {
	t.Helper()
	u := neu(t)
	u.dienst(dienstA, dienstB)
	for i := 1; i <= n; i++ {
		dienst := dienstA
		if i%2 == 0 {
			dienst = dienstB
		}
		u.offene(fmt.Sprintf("o%d", i), dienst, i)
		c := model.CaptureComplete
		if i%2 == 0 {
			c = model.CaptureIncomplete
		}
		u.exchange.haenge(fmt.Sprintf("o%d", i), c)
	}
	return u
}

// sollStatus ist der Status, den die Statusregel der Session oi von neuMehrere gibt.
func sollStatus(i int) model.SessionStatus {
	if i%2 == 0 {
		return model.SessionIncomplete
	}
	return model.SessionComplete
}

func oid(i int) string { return fmt.Sprintf("o%d", i) }

// zahl zählt die Aufrufe von op für die Kennung id aus dem Ereignisprotokoll.
func (u *welt) zahl(op, id string) int {
	n := 0
	for _, e := range u.p.liste() {
		if e == op+"("+id+")" {
			n++
		}
	}
	return n
}

// ctxMitWert ist ein ctx mit dem Wert w1, den die Fakes festhalten.
func ctxMitWert() context.Context {
	return context.WithValue(context.Background(), ctxSchluessel{}, "w1")
}

// startCloseAll führt CloseAll in einer Goroutine aus.
func (u *welt) startCloseAll(ctx context.Context) <-chan closeErgebnis {
	ch := make(chan closeErgebnis, 1)
	go func() {
		b, err := u.closeAll(ctx)
		ch <- closeErgebnis{b, err}
	}()
	return ch
}

type closeErgebnis struct {
	b   inbound.CloseBericht
	err error
}

// TestBeendenGleichzeitig: zwei gleichzeitige Beenden derselben Session rufen beide Drain, genau eines
// gelingt, das andere ist ODR-E7011 mit der Ursache, ohne Sperre im Dienst (B16, SPEC-124 Nr. 13).
// Drain und SetStatus halten jeden Aufruf, bis beide ihn betreten haben; synctest lässt die Frist der
// Barriere ohne Wartezeit ablaufen.
func TestBeendenGleichzeitig(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		u := neuBeenden(t, model.CaptureComplete)
		drain, setze := neueBarriere(2), neueBarriere(2)
		u.st.bei("Drain", 0, drain.an)
		u.sessions.beiSetStatus = func(context.Context, string) { setze.an() }
		var res [2]ergebnis
		var wg sync.WaitGroup
		for i := range res {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s, err := u.beenden(context.Background(), sid)
				res[i] = ergebnis{s, err}
			}()
		}
		wg.Wait()
		if drain.lief() || setze.lief() {
			t.Errorf("die Barriere lief ab: Drain %v, SetStatus %v", drain.lief(), setze.lief())
		}
		gewinner, verlierer := res[0], res[1]
		if gewinner.err != nil {
			gewinner, verlierer = verlierer, gewinner
		}
		if gewinner.err != nil || gewinner.s.Status != model.SessionComplete || gewinner.s.ID != sid {
			t.Errorf("Gewinner: %+v, %v", gewinner.s, gewinner.err)
		}
		code, status := codeVon(t, verlierer.err)
		if code != "ODR-E7011" || status != 409 || !errors.Is(verlierer.err, outbound.ErrUngueltigerUebergang) || verlierer.s != (model.Session{}) {
			t.Errorf("Verlierer: %s, %d, %v, %+v", code, status, verlierer.err, verlierer.s)
		}
		if u.st.anzahl("Drain") != 2 || u.st.anzahl("ListBySession") != 2 {
			t.Errorf("Drain %d-mal, ListBySession %d-mal", u.st.anzahl("Drain"), u.st.anzahl("ListBySession"))
		}
		if n, ok, st := u.sessions.versuche(sid); n != 2 || ok != 1 || st[0] != st[1] || st[0] != model.SessionComplete {
			t.Errorf("SetStatus %d-mal, %d gelungen, %v", n, ok, st)
		}
		if s := u.sessions.statusVon(sid); s != model.SessionComplete {
			t.Errorf("Status %q", s)
		}
		if n := len(u.log.alle()); n != 0 {
			t.Errorf("%d Einträge im Protokoll", n)
		}
	})
}

// TestBeendenGegenBeenden: ein Beenden, das open las, während ein anderes die Session beendet, liefert
// das Ergebnis eines Aufrufs danach (B17, SPEC-124 Nr. 13). Der Haltepunkt in Drain ordnet.
func TestBeendenGegenBeenden(t *testing.T) {
	u := neuBeenden(t, model.CaptureComplete)
	h := u.st.halte("Drain", 1)
	defer h.gib()
	a := u.startBeenden(context.Background(), sid)
	warte(t, h.erreicht, "Drain von A betreten")
	if b := abwarten(t, u.startBeenden(context.Background(), sid), "B"); b.err != nil || b.s.Status != model.SessionComplete {
		t.Fatalf("B: %+v, %v", b.s, b.err)
	}
	h.gib()
	r := abwarten(t, a, "Rückkehr von A")
	code, status := codeVon(t, r.err)
	if code != "ODR-E7011" || status != 409 || !errors.Is(r.err, outbound.ErrUngueltigerUebergang) || r.s != (model.Session{}) {
		t.Errorf("A: %s, %d, %v, %+v", code, status, r.err, r.s)
	}
	if u.st.anzahl("Drain") != 2 || u.st.anzahl("ListBySession") != 2 {
		t.Errorf("Drain %d-mal, ListBySession %d-mal", u.st.anzahl("Drain"), u.st.anzahl("ListBySession"))
	}
	if n, ok, _ := u.sessions.versuche(sid); n != 2 || ok != 1 {
		t.Errorf("SetStatus %d-mal, %d gelungen", n, ok)
	}
	if s := u.sessions.statusVon(sid); s != model.SessionComplete {
		t.Errorf("Status %q", s)
	}
}

// TestBeendenGegenLoeschen: ist die Session nach dem Beenden durch ein anderes Beenden gelöscht, liefert
// das wartende Beenden ODR-E7003 aus ListBySession und setzt keinen Status (B17, SPEC-124 Nr. 13).
func TestBeendenGegenLoeschen(t *testing.T) {
	u := neuBeenden(t, model.CaptureComplete)
	h := u.st.halte("Drain", 1)
	defer h.gib()
	a := u.startBeenden(context.Background(), sid)
	warte(t, h.erreicht, "Drain von A betreten")
	if b := abwarten(t, u.startBeenden(context.Background(), sid), "B"); b.err != nil {
		t.Fatal(b.err)
	}
	if _, err := u.del(context.Background(), sid); err != nil {
		t.Fatalf("Delete der beendeten Session: %v", err)
	}
	h.gib()
	r := abwarten(t, a, "Rückkehr von A")
	if code, _ := codeVon(t, r.err); code != "ODR-E7003" || r.s != (model.Session{}) {
		t.Errorf("A: %v, %+v", r.err, r.s)
	}
	if n, ok, _ := u.sessions.versuche(sid); n != 1 || ok != 1 {
		t.Errorf("SetStatus %d-mal, %d gelungen: A darf keinen Status setzen", n, ok)
	}
}

// TestCloseAllJede: CloseAll beendet jede offene Session aller Services und nur diese, je Element ein
// Fall, die gesuchte an erster, mittlerer und letzter Stelle (B18, SPEC-124 Nr. 14 b, c).
func TestCloseAllJede(t *testing.T) {
	for n := 1; n <= 4; n++ {
		u := neuMehrere(t, n)
		u.sessions.fuege("c1", dienstA, erzeugt(0), model.SessionComplete)
		u.sessions.fuege("i1", dienstB, erzeugt(n+1), model.SessionIncomplete)
		ctx := ctxMitWert()
		b, err := u.closeAll(ctx)
		if err != nil || b != (inbound.CloseBericht{Beendet: n}) {
			t.Errorf("%d offene: Bericht %+v, %v", n, b, err)
		}
		for i := 1; i <= n; i++ {
			id := oid(i)
			for _, op := range []string{"Sessions.Get", "Drain", "ListBySession", "SetStatus"} {
				if z := u.zahl(op, id); z != 1 {
					t.Errorf("%d offene, %s: %s %d-mal", n, id, op, z)
				}
			}
			if s := u.sessions.statusVon(id); s != sollStatus(i) {
				t.Errorf("%d offene, %s: Status %q, erwartet %q", n, id, s, sollStatus(i))
			}
		}
		for _, id := range []string{"c1", "i1"} {
			for _, op := range []string{"Sessions.Get", "Drain", "ListBySession", "SetStatus"} {
				if z := u.zahl(op, id); z != 0 {
					t.Errorf("%d offene, nicht offen %s: %s %d-mal", n, id, op, z)
				}
			}
		}
		if s := u.sessions.statusVon("c1"); s != model.SessionComplete {
			t.Errorf("c1: %q", s)
		}
		if s := u.sessions.statusVon("i1"); s != model.SessionIncomplete {
			t.Errorf("i1: %q", s)
		}
		for _, a := range u.record.aufrufe() {
			if a.frist != fristTest {
				t.Errorf("%d offene: Drain von %s mit frist %v", n, a.id, a.frist)
			}
		}
		if u.st.anzahl("FindOpen") != 1 || u.st.anzahl("Sessions.List") != 0 || u.st.anzahl("Dienste.List") != 0 {
			t.Errorf("%d offene: FindOpen %d-mal, List %d-mal", n, u.st.anzahl("FindOpen"), u.st.anzahl("Sessions.List"))
		}
		if g := u.st.gesehen("FindOpen"); len(g) != 1 || g[0] != ctx {
			t.Errorf("%d offene: FindOpen erhielt nicht den ctx des Aufrufs", n)
		}
	}
}

// TestCloseAllZugleich: CloseAll ruft Drain für mehrere Sessions zugleich mit derselben Frist und
// kehrt erst zurück, wenn jedes SetStatus geschehen ist (B19, SPEC-124 Nr. 14 c). Die Barriere läuft
// unter synctest ohne Wartezeit ab; ihr Ablaufen ist Rot.
func TestCloseAllZugleich(t *testing.T) {
	t.Run("Drain zugleich", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			u := neuMehrere(t, 3)
			b := neueBarriere(3)
			u.st.bei("Drain", 0, b.an)
			bericht, err := u.closeAll(context.Background())
			if b.lief() {
				t.Error("die Barriere lief ab: Drain nicht zugleich")
			}
			if err != nil || bericht != (inbound.CloseBericht{Beendet: 3}) {
				t.Errorf("Bericht %+v, %v", bericht, err)
			}
			a := u.record.aufrufe()
			if len(a) != 3 {
				t.Fatalf("Drain %d-mal", len(a))
			}
			for _, x := range a {
				if x.frist != fristTest {
					t.Errorf("Drain von %s mit frist %v", x.id, x.frist)
				}
			}
		})
	})
	t.Run("kehrt erst nach dem letzten SetStatus zurück", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			u := neuMehrere(t, 3)
			h := u.st.halteFuer("SetStatus", "o3")
			fertig := u.startCloseAll(context.Background())
			<-h.erreicht
			synctest.Wait()
			select {
			case r := <-fertig:
				t.Errorf("CloseAll kehrte vor dem letzten SetStatus zurück: %+v, %v", r.b, r.err)
				return
			default:
			}
			h.gib()
			r := <-fertig
			if r.err != nil || r.b != (inbound.CloseBericht{Beendet: 3}) {
				t.Errorf("Bericht %+v, %v", r.b, r.err)
			}
			if _, ok, _ := u.sessions.versuche("o3"); ok != 1 {
				t.Errorf("o3: %d gelungene SetStatus", ok)
			}
		})
	})
}

// TestCloseAllFehler: der Fehler einer Session hält die übrigen nicht auf und beendet sie nicht; die
// gescheiterte bleibt open (B20, SPEC-124 Nr. 14 d, e). Schleife über Position × Stelle × Art.
func TestCloseAllFehler(t *testing.T) {
	arten := []struct {
		art  error
		code string
	}{{verpackt(outbound.ErrBeschaedigt), "ODR-E3001"}, {errDisk, "ODR-E9001"}, {context.Canceled, ""}}
	for pos := 1; pos <= 3; pos++ {
		for i, st := range stellenBeenden() {
			for _, a := range arten {
				name := fmt.Sprintf("Position %d, %s, %v", pos, st.name, a.code)
				u := neuMehrere(t, 3)
				u.st.scheitereFuer(st.op, oid(pos), a.art)
				ctx := ctxMitWert()
				b, err := u.closeAll(ctx)
				if b != (inbound.CloseBericht{Beendet: 2, Fehlgeschlagen: 1}) {
					t.Errorf("%s: Bericht %+v", name, b)
				}
				switch {
				case a.code == "" || st.name == "Drain":
					if err != a.art {
						t.Errorf("%s: Fehler %v, erwartet den Fehler unverändert", name, err)
					}
					if a.code == "" && !ohneCode(err) {
						t.Errorf("%s: Kontextfehler mit Code", name)
					}
				default:
					if c, _ := codeVon(t, err); c != a.code || !errors.Is(err, a.art) {
						t.Errorf("%s: Fehler %v", name, err)
					}
				}
				for j := 1; j <= 3; j++ {
					for k, op := range []string{"Sessions.Get", "Drain", "ListBySession", "SetStatus"} {
						want := 1
						if j == pos && k > i {
							want = 0
						}
						if z := u.zahl(op, oid(j)); z != want {
							t.Errorf("%s: %s von %s %d-mal, erwartet %d", name, op, oid(j), z, want)
						}
					}
					want := sollStatus(j)
					if j == pos {
						want = model.SessionOpen
					}
					if s := u.sessions.statusVon(oid(j)); s != want {
						t.Errorf("%s: %s hat den Status %q, erwartet %q", name, oid(j), s, want)
					}
				}
				for k, op := range []string{"Sessions.Get", "Drain"} {
					g := u.st.gesehen(op)
					want := 3
					if i < k {
						want = 2
					}
					if len(g) != want {
						t.Errorf("%s: %s %d-mal gesehen, erwartet %d", name, op, len(g), want)
					}
					for _, c := range g {
						if c != ctx || c.Err() != nil {
							t.Errorf("%s: %s erhielt nicht den ctx des Aufrufs oder einen beendeten", name, op)
						}
					}
				}
				wantLog := 1
				if a.code == "" {
					wantLog = 0
				}
				if e := u.log.alle(); len(e) != wantLog || (wantLog == 1 && e[0].attrs["session_id"] != oid(pos)) {
					t.Errorf("%s: Protokoll %+v", name, e)
				}
			}
		}
	}
}

// TestCloseAllErsterFehler: der Fehler von CloseAll ist der der ersten gescheiterten Session in der
// Reihenfolge von FindOpen, nicht der zuerst zurückgekehrten (B21, SPEC-124 Nr. 14 e). Der Haltepunkt
// am Get der niedrigeren Position lässt die höhere zuerst zurückkehren.
func TestCloseAllErsterFehler(t *testing.T) {
	for _, p := range [][2]int{{1, 3}, {1, 2}, {2, 3}} {
		t.Run(fmt.Sprintf("Positionen %d und %d", p[0], p[1]), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				u := neuMehrere(t, 3)
				niedrig, hoch := verpackt(outbound.ErrBeschaedigt), errDisk
				u.st.scheitereFuer("Sessions.Get", oid(p[0]), niedrig)
				u.st.scheitereFuer("Sessions.Get", oid(p[1]), hoch)
				h := u.st.halteFuer("Sessions.Get", oid(p[0]))
				fertig := u.startCloseAll(context.Background())
				<-h.erreicht
				synctest.Wait()
				h.gib()
				r := <-fertig
				f, ok := r.err.(*inbound.Fehler)
				if !ok || f.Code != model.CodeE3001 || !errors.Is(r.err, niedrig) || errors.Is(r.err, hoch) {
					t.Errorf("Fehler %v", r.err)
				}
				if r.b != (inbound.CloseBericht{Beendet: 1, Fehlgeschlagen: 2}) {
					t.Errorf("Bericht %+v", r.b)
				}
			})
		})
	}
}

// TestCloseAllBereitsBeendet: ODR-E7011 einer Session in CloseAll ist kein Fehler, nicht gezählt und
// ohne Eintrag (B22, SPEC-124 Nr. 14 d).
func TestCloseAllBereitsBeendet(t *testing.T) {
	t.Run("nicht mehr offen beim Lesen", func(t *testing.T) {
		u := neuMehrere(t, 3)
		liste, _ := u.sessions.hat("o1")
		o2, _ := u.sessions.hat("o2")
		o3, _ := u.sessions.hat("o3")
		u.sessions.setzeOffene([]model.Session{liste, o2, o3})
		u.sessions.fuege("o2", dienstB, erzeugt(2), model.SessionComplete)
		b, err := u.closeAll(context.Background())
		if err != nil || b != (inbound.CloseBericht{Beendet: 2}) || len(u.log.alle()) != 0 {
			t.Errorf("Bericht %+v, %v, Protokoll %+v", b, err, u.log.alle())
		}
		if u.st.anzahl("Sessions.Get") != 3 || u.st.anzahl("Drain") != 2 {
			t.Errorf("Get %d-mal, Drain %d-mal", u.st.anzahl("Sessions.Get"), u.st.anzahl("Drain"))
		}
	})
	t.Run("ein anderes Beenden kam im Statuswechsel zuvor", func(t *testing.T) {
		u := neuMehrere(t, 3)
		u.st.scheitereFuer("SetStatus", "o2", verpackt(outbound.ErrUngueltigerUebergang))
		b, err := u.closeAll(context.Background())
		if err != nil || b != (inbound.CloseBericht{Beendet: 2}) || len(u.log.alle()) != 0 {
			t.Errorf("Bericht %+v, %v, Protokoll %+v", b, err, u.log.alle())
		}
		if u.st.anzahl("Drain") != 3 {
			t.Errorf("Drain %d-mal", u.st.anzahl("Drain"))
		}
	})
}

// TestCloseAllLeer: ohne offene Session ist das Ergebnis der Bericht aus Nullen, kein Drain und kein
// weiterer Port (B23, SPEC-124 Nr. 14 b).
func TestCloseAllLeer(t *testing.T) {
	faelle := map[string]func(*welt){
		"nil":         func(u *welt) { u.sessions.setzeOffene(nil) },
		"leere Liste": func(u *welt) { u.sessions.setzeOffene([]model.Session{}) },
		"nur nicht offene": func(u *welt) {
			u.sessions.fuege("c1", dienstA, erzeugt(1), model.SessionComplete)
			u.sessions.fuege("i1", dienstB, erzeugt(2), model.SessionIncomplete)
		},
	}
	for name, setup := range faelle {
		u := neu(t)
		u.dienst(dienstA, dienstB)
		setup(u)
		b, err := u.closeAll(context.Background())
		if err != nil || b != (inbound.CloseBericht{}) {
			t.Errorf("%s: Bericht %+v, %v", name, b, err)
		}
		for _, op := range alleOps() {
			want := 0
			if op == "FindOpen" {
				want = 1
			}
			if n := u.st.anzahl(op); n != want {
				t.Errorf("%s: %s %d-mal, erwartet %d", name, op, n, want)
			}
		}
		if n := len(u.log.alle()); n != 0 {
			t.Errorf("%s: %d Einträge im Protokoll", name, n)
		}
	}
}

// TestCloseAllFindOpen: ein Fehler von FindOpen ist ein Fehler mit dem Gegenstand Sonstiges, ein
// Kontextfehler bleibt unverändert; der Bericht ist aus Nullen, es folgt kein Drain (B24, SPEC-124
// Nr. 14 b).
func TestCloseAllFindOpen(t *testing.T) {
	mitCode := map[string]struct {
		art  error
		code string
	}{
		"ErrBeschaedigt":   {verpackt(outbound.ErrBeschaedigt), "ODR-E3001"},
		"Platte":           {errDisk, "ODR-E9001"},
		"ErrNichtGefunden": {verpackt(outbound.ErrNichtGefunden), "ODR-E9001"},
	}
	for name, f := range mitCode {
		u := neuMehrere(t, 2)
		u.st.scheitere("FindOpen", 1, f.art)
		ctx := ctxMitWert()
		b, err := u.closeAll(ctx)
		if c, _ := codeVon(t, err); c != f.code || !errors.Is(err, f.art) || b != (inbound.CloseBericht{}) {
			t.Errorf("%s: %v, Bericht %+v", name, err, b)
		}
		pruefeNichtsDanach(ctx, t, name, u)
	}
	ohne := map[string]error{
		"Canceled":                      context.Canceled,
		"DeadlineExceeded":              context.DeadlineExceeded,
		"Canceled neben ErrBeschaedigt": errors.Join(context.Canceled, verpackt(outbound.ErrBeschaedigt)),
	}
	for name, art := range ohne {
		u := neuMehrere(t, 2)
		u.st.scheitere("FindOpen", 1, art)
		ctx := ctxMitWert()
		b, err := u.closeAll(ctx)
		if err != art || !ohneCode(err) || b != (inbound.CloseBericht{}) {
			t.Errorf("%s: %v, Bericht %+v", name, err, b)
		}
		pruefeNichtsDanach(ctx, t, name, u)
	}
}

func pruefeNichtsDanach(ctx context.Context, t *testing.T, name string, u *welt) {
	t.Helper()
	for _, op := range []string{"Sessions.Get", "Drain", "ListBySession", "SetStatus"} {
		if n := u.st.anzahl(op); n != 0 {
			t.Errorf("%s: %s %d-mal", name, op, n)
		}
	}
	if g := u.st.gesehen("FindOpen"); u.st.anzahl("FindOpen") != 1 || len(g) != 1 || g[0] != ctx {
		t.Errorf("%s: FindOpen %d-mal, ctx unverändert %v", name, u.st.anzahl("FindOpen"), len(g) == 1 && g[0] == ctx)
	}
	if n := len(u.log.alle()); n != 0 {
		t.Errorf("%s: %d Einträge im Protokoll", name, n)
	}
}

// TestCloseAllKontext: ein vorab beendeter ctx lässt CloseAll nichts beenden; endet er in Drain, setzt
// CloseAll alle Status, denn ListBySession und SetStatus laufen ohne seinen Abbruch (B25, SPEC-124
// Nr. 11 e, 14 b, c).
func TestCloseAllKontext(t *testing.T) {
	ende, abbruch := context.WithCancel(context.Background())
	abbruch()
	vorbei, abbruch2 := context.WithDeadline(context.Background(), time.Unix(1, 0))
	defer abbruch2()
	for name, ctx := range map[string]context.Context{"Canceled": ende, "DeadlineExceeded": vorbei} {
		u := neuMehrere(t, 3)
		b, err := u.closeAll(ctx)
		if err != ctx.Err() || b != (inbound.CloseBericht{}) {
			t.Errorf("%s vorab: %v, Bericht %+v", name, err, b)
		}
		for _, op := range []string{"Sessions.Get", "Drain", "ListBySession", "SetStatus"} {
			if n := u.st.anzahl(op); n != 0 {
				t.Errorf("%s vorab: %s %d-mal", name, op, n)
			}
		}
	}

	nachher := func(t *testing.T, u *welt, b inbound.CloseBericht, err error) {
		t.Helper()
		if err != nil || b != (inbound.CloseBericht{Beendet: 3}) {
			t.Errorf("Bericht %+v, %v", b, err)
		}
		for i := 1; i <= 3; i++ {
			if s := u.sessions.statusVon(oid(i)); s != sollStatus(i) {
				t.Errorf("%s: Status %q, erwartet %q", oid(i), s, sollStatus(i))
			}
		}
		for _, op := range []string{"ListBySession", "SetStatus"} {
			g := u.st.gesehen(op)
			if len(g) != 3 {
				t.Errorf("%s: %d Aufrufe", op, len(g))
			}
			for _, c := range g {
				if _, hat := c.Deadline(); c.Value(ctxSchluessel{}) != "w1" || c.Err() != nil || c.Done() != nil || hat {
					t.Errorf("%s: Wert %v, Err %v, Done %v, Frist %v", op, c.Value(ctxSchluessel{}), c.Err(), c.Done(), hat)
				}
			}
		}
	}
	t.Run("Canceled in Drain", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			u := neuMehrere(t, 3)
			ctx, stopp := context.WithCancel(ctxMitWert())
			defer stopp()
			b := neueBarriere(3)
			u.st.bei("Drain", 0, func() { b.an(); stopp() })
			bericht, err := u.closeAll(ctx)
			if b.lief() || ctx.Err() == nil {
				t.Errorf("Barriere abgelaufen %v, ctx.Err %v", b.lief(), ctx.Err())
			}
			nachher(t, u, bericht, err)
		})
	})
	t.Run("DeadlineExceeded in Drain", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			u := neuMehrere(t, 3)
			ctx := &verfallCtx{wert: "w1"}
			b := neueBarriere(3)
			u.st.bei("Drain", 0, func() { b.an(); ctx.verfalle() })
			bericht, err := u.closeAll(ctx)
			if b.lief() || ctx.Err() != context.DeadlineExceeded {
				t.Errorf("Barriere abgelaufen %v, ctx.Err %v", b.lief(), ctx.Err())
			}
			nachher(t, u, bericht, err)
		})
	})
}

// pruefeEintrag hält einen Eintrag gegen SPEC-124 Nr. 9: Stufe, Text und genau die vier Attribute.
func pruefeEintrag(t *testing.T, name string, e eintrag, service, session, code string) {
	t.Helper()
	soll := map[string]string{"ereignis": "beenden_fehlgeschlagen", "service_id": service, "session_id": session, "error_code": code}
	if e.stufe != slog.LevelError || e.text != "session: Beenden beim Herunterfahren fehlgeschlagen" {
		t.Errorf("%s: Stufe %v, Text %q", name, e.stufe, e.text)
	}
	if len(e.attrs) != len(soll) {
		t.Errorf("%s: Attribute %v", name, e.attrs)
	}
	for k, v := range soll {
		if e.attrs[k] != v {
			t.Errorf("%s: Attribut %s = %q, erwartet %q", name, k, e.attrs[k], v)
		}
	}
	if strings.Contains(e.text, marker) {
		t.Errorf("%s: Text trägt die Marke", name)
	}
	for k, v := range e.attrs {
		if strings.Contains(v, marker) {
			t.Errorf("%s: Attribut %s trägt die Marke: %q", name, k, v)
		}
	}
}

// TestCloseAllProtokoll: je gescheiterter Session ohne Kontextfehler genau ein Eintrag mit Text, Stufe
// und den Attributen aus SPEC-124 Nr. 9, in der Reihenfolge der Sessions (B26).
func TestCloseAllProtokoll(t *testing.T) {
	t.Run("Reihenfolge der Sessions, nicht der Rückkehr", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			u := neuMehrere(t, 3)
			u.offene("o3", dienstB, 3)
			u.st.scheitereFuer("Sessions.Get", "o1", verpackt(outbound.ErrBeschaedigt))
			u.st.scheitereFuer("Sessions.Get", "o3", errDisk)
			h := u.st.halteFuer("Sessions.Get", "o1")
			fertig := u.startCloseAll(context.Background())
			<-h.erreicht
			synctest.Wait()
			h.gib()
			r := <-fertig
			if r.b != (inbound.CloseBericht{Beendet: 1, Fehlgeschlagen: 2}) {
				t.Errorf("Bericht %+v", r.b)
			}
			e := u.log.alle()
			if len(e) != 2 {
				t.Fatalf("%d Einträge: %+v", len(e), e)
			}
			pruefeEintrag(t, "erster", e[0], dienstA, "o1", "ODR-E3001")
			pruefeEintrag(t, "zweiter", e[1], dienstB, "o3", "ODR-E9001")
		})
	})
	t.Run("Dienst und Kennung der Session", func(t *testing.T) {
		u := neuMehrere(t, 3)
		u.st.scheitereFuer("Sessions.Get", "o2", errDisk)
		_, _ = u.closeAll(context.Background())
		e := u.log.alle()
		if len(e) != 1 {
			t.Fatalf("%d Einträge", len(e))
		}
		pruefeEintrag(t, "o2", e[0], dienstB, "o2", "ODR-E9001")
	})

	faelle := []struct {
		name  string
		op    string
		art   error
		codes []string
	}{
		{"Fehler von Drain mit Code", "Drain", e4001(), []string{"ODR-E4001"}},
		{"ErrUngueltig von Drain", "Drain", model.ErrUngueltig, []string{"ODR-E9001"}},
		{"Platte in Drain", "Drain", errDisk, []string{"ODR-E9001"}},
		{"ErrBeschaedigt in ListBySession", "ListBySession", verpackt(outbound.ErrBeschaedigt), []string{"ODR-E3001"}},
		{"Platte in SetStatus", "SetStatus", errDisk, []string{"ODR-E9001"}},
		{"Canceled", "Sessions.Get", context.Canceled, nil},
		{"DeadlineExceeded", "Sessions.Get", context.DeadlineExceeded, nil},
		{"Canceled in Drain", "Drain", context.Canceled, nil},
		{"Canceled neben Platte", "SetStatus", errors.Join(context.Canceled, errDisk), nil},
		{"ODR-E7011", "SetStatus", verpackt(outbound.ErrUngueltigerUebergang), nil},
		{"Erfolg", "Sessions.Get", nil, nil},
	}
	for _, f := range faelle {
		u := neuMehrere(t, 3)
		if f.art != nil {
			u.st.scheitereFuer(f.op, "o2", f.art)
		}
		_, _ = u.closeAll(context.Background())
		e := u.log.alle()
		if len(e) != len(f.codes) {
			t.Errorf("%s: %d Einträge: %+v", f.name, len(e), e)
			continue
		}
		for i, c := range f.codes {
			pruefeEintrag(t, f.name, e[i], dienstB, "o2", c)
		}
	}
}

// szenario ist ein Weg durch CloseAll mit der Zahl der offenen Sessions und der Zahl der Einträge.
type szenario struct {
	name    string
	offene  int
	setup   func(*welt)
	ctx     func() context.Context
	eintrag int
}

func closeAllSzenarien() []szenario {
	bg := func() context.Context { return context.Background() }
	scheitere := func(op string, art error) func(*welt) {
		return func(u *welt) { u.st.scheitereFuer(op, "o2", art) }
	}
	ss := []szenario{
		{"Gutfall", 3, nil, bg, 0},
		{"ohne offene Session", 0, nil, bg, 0},
		{"FindOpen Fehler", 0, func(u *welt) { u.st.scheitere("FindOpen", 1, errDisk) }, bg, 0},
		{"ctx vorab beendet", 0, nil, func() context.Context {
			c, abbruch := context.WithCancel(context.Background())
			abbruch()
			return c
		}, 0},
		{"nil-ctx", 0, nil, func() context.Context { return nil }, 0},
		{"ODR-E7011", 3, scheitere("SetStatus", verpackt(outbound.ErrUngueltigerUebergang)), bg, 0},
		{"Fehler in Drain mit Code", 3, scheitere("Drain", e4001()), bg, 1},
	}
	for _, st := range stellenBeenden() {
		ss = append(ss, szenario{"Fehler in " + st.name, 3, scheitere(st.op, errDisk), bg, 1})
		ss = append(ss, szenario{"Kontextfehler in " + st.name, 3, scheitere(st.op, context.Canceled), bg, 0})
	}
	return ss
}

// TestCloseAllNurGenannte: CloseAll ruft FindOpen genau einmal (bei nil-ctx nicht) und je Session die
// Operationen von Beenden, nichts von Dienste und Wartung, weder Now noch SessionID, und schreibt nur
// die Einträge von SPEC-124 Nr. 9 (B27, SPEC-124 Nr. 14).
func TestCloseAllNurGenannte(t *testing.T) {
	for _, s := range closeAllSzenarien() {
		u := neuMehrere(t, s.offene)
		if s.setup != nil {
			s.setup(u)
		}
		_, _ = u.closeAll(s.ctx())
		erlaubt := map[string]int{"Sessions.Get": s.offene, "Drain": s.offene, "ListBySession": s.offene, "SetStatus": s.offene}
		if s.name != "nil-ctx" {
			erlaubt["FindOpen"] = 1
		}
		for _, op := range alleOps() {
			n := u.st.anzahl(op)
			if n > erlaubt[op] || (op == "FindOpen" && n != erlaubt[op]) {
				t.Errorf("%s: %s %d-mal, erlaubt %d", s.name, op, n, erlaubt[op])
			}
		}
		if s.name == "Gutfall" {
			for op, soll := range erlaubt {
				if n := u.st.anzahl(op); n != soll {
					t.Errorf("%s: %s %d-mal, erwartet %d", s.name, op, n, soll)
				}
			}
		}
		e := u.log.alle()
		if len(e) != s.eintrag {
			t.Errorf("%s: %d Einträge, erwartet %d", s.name, len(e), s.eintrag)
		}
		for _, x := range e {
			if x.attrs["ereignis"] != "beenden_fehlgeschlagen" {
				t.Errorf("%s: Eintrag %+v", s.name, x)
			}
		}
	}
}

// TestCloseAllGegenBeenden: das Beenden einer Session in CloseAll gegen ein einzelnes Beenden derselben
// Session: genau ein Statuswechsel, CloseAll meldet keinen Fehler (B28, SPEC-124 Nr. 14 c, d). Die
// Haltepunkte ordnen mit Kanälen.
func TestCloseAllGegenBeenden(t *testing.T) {
	t.Run("CloseAll liest nach dem einzelnen Beenden", func(t *testing.T) {
		u := neuMehrere(t, 1)
		h := u.st.halte("Sessions.Get", 1)
		defer h.gib()
		fertig := u.startCloseAll(context.Background())
		warte(t, h.erreicht, "Get von CloseAll betreten")
		if s, err := u.beenden(context.Background(), "o1"); err != nil || s.Status != model.SessionComplete {
			t.Fatalf("einzeln: %+v, %v", s, err)
		}
		h.gib()
		select {
		case r := <-fertig:
			if r.err != nil || r.b != (inbound.CloseBericht{}) {
				t.Errorf("CloseAll: %+v, %v", r.b, r.err)
			}
		case <-time.After(zeitgrenze):
			t.Fatal("CloseAll kehrte nicht zurück")
		}
		if n := u.st.anzahl("Drain"); n != 1 {
			t.Errorf("Drain %d-mal, erwartet 1 (das einzelne Beenden)", n)
		}
		if n := len(u.log.alle()); n != 0 {
			t.Errorf("%d Einträge im Protokoll", n)
		}
	})

	for _, alleGewinnt := range []bool{true, false} {
		t.Run(fmt.Sprintf("beide halten im Statuswechsel, CloseAll gewinnt: %v", alleGewinnt), func(t *testing.T) {
			u := neuMehrere(t, 1)
			hAlle := &halt{erreicht: make(chan struct{}), frei: make(chan struct{})}
			hEinzeln := &halt{erreicht: make(chan struct{}), frei: make(chan struct{})}
			defer hAlle.gib()
			defer hEinzeln.gib()
			u.sessions.beiSetStatus = func(ctx context.Context, _ string) {
				switch ctx.Value(ctxSchluessel{}) {
				case "alle":
					close(hAlle.erreicht)
					<-hAlle.frei
				case "einzeln":
					close(hEinzeln.erreicht)
					<-hEinzeln.frei
				}
			}
			alle := u.startCloseAll(context.WithValue(context.Background(), ctxSchluessel{}, "alle"))
			einzeln := u.startBeenden(context.WithValue(context.Background(), ctxSchluessel{}, "einzeln"), "o1")
			warte(t, hAlle.erreicht, "SetStatus von CloseAll betreten")
			warte(t, hEinzeln.erreicht, "SetStatus des einzelnen Beenden betreten")
			erstes, zweites := hAlle, hEinzeln
			if !alleGewinnt {
				erstes, zweites = hEinzeln, hAlle
			}
			var ra closeErgebnis
			var re ergebnis
			warteAuf := func(gewinner bool) {
				if alleGewinnt == gewinner {
					select {
					case ra = <-alle:
					case <-time.After(zeitgrenze):
						t.Fatal("CloseAll kehrte nicht zurück")
					}
				} else {
					re = abwarten(t, einzeln, "Beenden")
				}
			}
			erstes.gib()
			warteAuf(true)
			zweites.gib()
			warteAuf(false)
			if ra.err != nil {
				t.Errorf("CloseAll: %v", ra.err)
			}
			if alleGewinnt {
				if ra.b != (inbound.CloseBericht{Beendet: 1}) {
					t.Errorf("CloseAll gewinnt: Bericht %+v", ra.b)
				}
				if code, _ := codeVon(t, re.err); code != "ODR-E7011" {
					t.Errorf("einzeln verliert: %v", re.err)
				}
			} else {
				if ra.b != (inbound.CloseBericht{}) {
					t.Errorf("einzeln gewinnt: Bericht von CloseAll %+v", ra.b)
				}
				if re.err != nil || re.s.Status != model.SessionComplete {
					t.Errorf("einzeln gewinnt: %+v, %v", re.s, re.err)
				}
			}
			if n, ok, _ := u.sessions.versuche("o1"); n != 2 || ok != 1 {
				t.Errorf("SetStatus %d-mal, %d gelungen", n, ok)
			}
			if n := len(u.log.alle()); n != 0 {
				t.Errorf("%d Einträge im Protokoll", n)
			}
		})
	}
}

// TestCloseAllGegenAnlegen: CloseAll ist ein Durchgang: eine Session, die nach FindOpen angelegt
// wird, bleibt open und unberührt (B29, SPEC-124 Nr. 14 f).
func TestCloseAllGegenAnlegen(t *testing.T) {
	u := neuMehrere(t, 1)
	h := u.st.halteFuer("Drain", "o1")
	defer h.gib()
	fertig := u.startCloseAll(context.Background())
	warte(t, h.erreicht, "Drain von o1 betreten")
	if _, err := u.create(context.Background(), dienstA, "x1"); err != nil {
		t.Fatal(err)
	}
	h.gib()
	select {
	case r := <-fertig:
		if r.err != nil || r.b != (inbound.CloseBericht{Beendet: 1}) {
			t.Errorf("Bericht %+v, %v", r.b, r.err)
		}
	case <-time.After(zeitgrenze):
		t.Fatal("CloseAll kehrte nicht zurück")
	}
	if s := u.sessions.statusVon("x1"); s != model.SessionOpen {
		t.Errorf("x1: Status %q", s)
	}
	for _, op := range []string{"Sessions.Get", "Drain", "ListBySession", "SetStatus"} {
		if z := u.zahl(op, "x1"); z != 0 {
			t.Errorf("x1: %s %d-mal", op, z)
		}
	}
	if n := u.st.anzahl("FindOpen"); n != 1 {
		t.Errorf("FindOpen %d-mal", n)
	}
}

// TestBeendenRace: der Dienst hält nach New keinen veränderlichen Zustand, auch nicht in Beenden und
// CloseAll; unter -race meldet der Detektor jedes ungeordnete Paar (B30, SPEC-124 Nr. 10, 13). Jede
// Session endet mit genau einem gelungenen SetStatus, jedes ODR-E7011 gehört einem Verlierer.
func TestBeendenRace(t *testing.T) {
	const gor, runden = 16, 50
	u := neu(t)
	u.dienst(dienstA, dienstB)
	for j := 0; j < runden; j++ {
		for g := 0; g < gor/2; g++ {
			u.offene(fmt.Sprintf("gem-%d-%d", g, j), dienstA, j)
		}
	}
	var mu sync.Mutex
	var fehlerTexte []string
	var gewinne = map[string]int{}
	var verloren = map[string]int{}
	melde := func(id string, err error) {
		mu.Lock()
		defer mu.Unlock()
		var f *inbound.Fehler
		if err == nil {
			gewinne[id]++
		} else if errors.As(err, &f) && f.Code == model.CodeE7011 {
			verloren[id]++
		} else {
			fehlerTexte = append(fehlerTexte, fmt.Sprintf("%s: %v", id, err))
		}
	}
	ctx := context.Background()
	parallel(gor, func(g int) {
		for j := 0; j < runden; j++ {
			id := fmt.Sprintf("s-%d-%d", g, j)
			u.exchange.haenge(id, model.CaptureComplete)
			if _, err := u.v.Create(ctx, dienstB, id); err != nil {
				t.Errorf("Create %s: %v", id, err)
				return
			}
			_, err := u.v.Beenden(ctx, id)
			melde(id, err)
			gem := fmt.Sprintf("gem-%d-%d", g/2, j)
			_, err = u.v.Beenden(ctx, gem)
			melde(gem, err)
			if _, err := u.v.Get(ctx, id); err != nil {
				t.Errorf("Get %s: %v", id, err)
			}
			if _, err := u.v.List(ctx); err != nil {
				t.Errorf("List: %v", err)
			}
			if b, err := u.v.CloseAll(ctx); err != nil || b.Fehlgeschlagen != 0 {
				t.Errorf("CloseAll: %+v, %v", b, err)
			}
		}
	})
	if len(fehlerTexte) != 0 {
		t.Fatalf("Fehler außer ODR-E7011: %v", fehlerTexte)
	}
	ids := make([]string, 0, len(gewinne))
	for id := range gewinne {
		ids = append(ids, id)
	}
	for id := range verloren {
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		if _, ok, _ := u.sessions.versuche(id); ok != 1 {
			t.Errorf("%s: %d gelungene SetStatus, erwartet 1", id, ok)
		}
		if gewinne[id] > 1 {
			t.Errorf("%s: %d Gewinner", id, gewinne[id])
		}
		if verloren[id] > 0 && u.sessions.statusVon(id) == model.SessionOpen {
			t.Errorf("%s: ODR-E7011 bei offener Session", id)
		}
	}
	for g := 0; g < gor; g++ {
		for j := 0; j < runden; j++ {
			if s := u.sessions.statusVon(fmt.Sprintf("s-%d-%d", g, j)); s == model.SessionOpen {
				t.Errorf("s-%d-%d bleibt open", g, j)
			}
		}
	}
}

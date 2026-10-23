package maintenance_test

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// altBestand legt die fünf Sessions von W8 an; der Name trägt den Abstand der Grenze.
func altBestand(u *welt) {
	u.dienst(dienstA)
	grenze := t0().Add(-30 * tag)
	u.sessions.fuege("a", dienstA, grenze.Add(-time.Nanosecond), model.SessionComplete)
	u.sessions.fuege("b", dienstA, grenze, model.SessionComplete)
	u.sessions.fuege("c", dienstA, grenze.Add(time.Nanosecond), model.SessionComplete)
	u.sessions.fuege("d", dienstA, t0(), model.SessionComplete)
	u.sessions.fuege("e", dienstA, t0().Add(time.Hour), model.SessionComplete)
}

// TestAlterGrenze: eine Session ist älter, wenn ihr Beginn echt vor jetzt minus MaxAge liegt; die
// Zeit wird einmal bezogen (W8).
func TestAlterGrenze(t *testing.T) {
	u := neu(t, mitMaxAge(30*tag))
	altBestand(u)
	b, err := u.startup(context.Background())
	if err != nil || b.SessionsGeloescht != 1 || !slices.Equal(u.p.mit("Delete"), []string{"Delete(a)"}) {
		t.Errorf("%v %+v %v", err, b, u.p.mit("Delete"))
	}
	for _, id := range []string{"b", "c", "d", "e"} {
		if _, da := u.sessions.status(id); !da {
			t.Errorf("Session %s gelöscht", id)
		}
	}

	v := neu(t, mitMaxAge(30*tag))
	altBestand(v)
	v.st.bei("Now", 0, func() { v.uhr.Advance(time.Nanosecond) })
	b, err = v.startup(context.Background())
	if err != nil || b.SessionsGeloescht != 1 || v.st.anzahl("Now") != 1 {
		t.Errorf("rückende Uhr: %v %+v, Now() %d-mal", err, b, v.st.anzahl("Now"))
	}
	if _, da := v.sessions.status("b"); !da {
		t.Error("rückende Uhr: Session b gelöscht")
	}
}

// TestAlterAus: MaxAge 0 schaltet die Aufbewahrung aus und ruft keinen Port der Stufe (W9).
func TestAlterAus(t *testing.T) {
	u := neu(t)
	u.dienst(dienstA)
	u.sessions.fuege("s1", dienstA, time.Date(1, 1, 1, 0, 0, 1, 0, time.UTC), model.SessionComplete)
	b, err := u.startup(context.Background())
	if err != nil || b.SessionsGeloescht != 0 || !reflect.DeepEqual(b, leerBericht()) {
		t.Errorf("%v %+v", err, b)
	}
	for _, op := range []string{"Dienste.List", "Sessions.List", "Delete", "Now"} {
		if n := u.st.anzahl(op); n != 0 {
			t.Errorf("%s %d-mal", op, n)
		}
	}
}

// TestAlterSehrGross: der Vergleich geht über die Grenze, nicht über die Differenz; bei dem
// größten MaxAge liegt die Grenze im Jahr 1734 (W9).
func TestAlterSehrGross(t *testing.T) {
	u := neu(t, mitMaxAge(math.MaxInt64))
	u.dienst(dienstA)
	u.sessions.fuege("jahr1", dienstA, time.Date(1, 1, 1, 0, 0, 1, 0, time.UTC), model.SessionComplete)
	u.sessions.fuege("jahr1800", dienstA, time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC), model.SessionComplete)
	b, err := u.startup(context.Background())
	if err != nil || b.SessionsGeloescht != 1 || !slices.Equal(u.p.mit("Delete"), []string{"Delete(jahr1)"}) {
		t.Errorf("%v %+v %v", err, b, u.p.mit("Delete"))
	}
	if _, da := u.sessions.status("jahr1800"); !da {
		t.Error("Session von 1800 gelöscht")
	}
}

// TestAlterAlleServices: die Aufbewahrung läuft über alle Services in der Reihenfolge der Ablage,
// auch über einen ohne Session (W10).
func TestAlterAlleServices(t *testing.T) {
	u := neu(t, mitMaxAge(30*tag))
	u.dienst("A", "B", "C")
	u.sessions.fuege("a1", "A", t0().Add(-50*tag), model.SessionComplete)
	u.sessions.fuege("a2", "A", t0().Add(-49*tag), model.SessionComplete)
	u.sessions.fuege("c1", "C", t0().Add(-48*tag), model.SessionComplete)
	u.sessions.fuege("c2", "C", t0().Add(-tag), model.SessionComplete)
	b, err := u.startup(context.Background())
	if err != nil || b.SessionsGeloescht != 3 {
		t.Errorf("%v %+v", err, b)
	}
	if got := u.p.mit("Delete"); !slices.Equal(got, []string{"Delete(a1)", "Delete(a2)", "Delete(c1)"}) {
		t.Errorf("Delete %v", got)
	}
	if u.st.anzahl("Dienste.List") != 1 || !slices.Equal(u.sessions.gelistet, []string{"A", "B", "C"}) {
		t.Errorf("Dienste.List %d-mal, Sessions.List für %v", u.st.anzahl("Dienste.List"), u.sessions.gelistet)
	}
}

// TestAlterDelete: ErrReferenziert überspringt mit Warnung und zählt, ErrNichtGefunden wird
// übergangen; die Stufe läuft weiter (W11).
func TestAlterDelete(t *testing.T) {
	bestand := func() *welt {
		u := neu(t, mitMaxAge(30*tag))
		u.dienst(dienstA)
		for i, id := range []string{"q1", "q2", "q3"} {
			u.sessions.fuege(id, dienstA, t0().Add(-time.Duration(50-i)*tag), model.SessionComplete)
		}
		return u
	}
	u := bestand()
	u.st.scheitere("Delete", 1, fmt.Errorf("x: %w", outbound.ErrReferenziert))
	b, err := u.startup(context.Background())
	if err != nil || b.SessionsUebersprungen != 1 || b.SessionsGeloescht != 2 {
		t.Errorf("(a) %v %+v", err, b)
	}
	if _, da := u.sessions.status("q1"); !da {
		t.Error("(a) Quelle einer Baseline gelöscht")
	}
	soll := []eintrag{{stufe: slog.LevelWarn, text: "wartung: Session als Quelle einer Baseline nicht gelöscht", attrs: map[string]string{
		"ereignis": "aufbewahrung_uebersprungen", "error_code": "ODR-E7004", "service_id": dienstA, "session_id": "q1"}}}
	if got := u.log.alle(); !reflect.DeepEqual(got, soll) {
		t.Errorf("(a) Protokoll %v, erwartet %v", got, soll)
	}

	v := bestand()
	v.st.scheitere("Delete", 2, fmt.Errorf("x: %w", outbound.ErrNichtGefunden))
	b, err = v.startup(context.Background())
	if err != nil || b.SessionsUebersprungen != 0 || b.SessionsGeloescht != 2 || len(v.log.alle()) != 0 {
		t.Errorf("(b) %v %+v, Protokoll %v", err, b, v.log.alle())
	}
	if v.st.anzahl("Delete") != 3 {
		t.Errorf("(b) Delete %d-mal", v.st.anzahl("Delete"))
	}
}

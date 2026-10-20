package supporttest_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

func plus2() *time.Location { return time.FixedZone("plus2", 2*3600) }

func zeit(s string) time.Time {
	z, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return z
}

func panikt(f func()) (p bool) {
	defer func() { p = recover() != nil }()
	f()
	return false
}

func TestFestVertrag(t *testing.T) {
	t.Run("FestUhr UTC", func(t *testing.T) {
		supporttest.UhrVertrag(t, supporttest.NewFestUhr(zeit("2026-10-20T10:00:00Z")))
	})
	t.Run("FestUhr Start in +02:00", func(t *testing.T) {
		supporttest.UhrVertrag(t, supporttest.NewFestUhr(time.Date(2026, 10, 20, 10, 0, 0, 5, plus2())))
	})
	t.Run("FestIDs", func(t *testing.T) { supporttest.IDVertrag(t, supporttest.NewFestIDs()) })
}

func TestFestUhrStart(t *testing.T) {
	for _, s := range []string{"0001-01-01T00:00:00.000000001Z", "9999-12-31T23:59:59.999999999Z"} {
		if got := supporttest.NewFestUhr(zeit(s)).Now(); !got.Equal(zeit(s)) {
			t.Errorf("Start %s: Now = %v", s, got)
		}
	}
	if got := supporttest.NewFestUhr(zeit("2026-10-20T10:00:00+02:00")).Now(); !got.Equal(zeit("2026-10-20T08:00:00Z")) {
		t.Errorf("Start in +02:00: Now = %v, erwartet 08:00 UTC", got)
	}
	for name, start := range map[string]time.Time{
		"Nullzeit":            time.Time{},
		"Nullzeit in +02:00":  time.Time{}.In(plus2()),
		"Jahr 10000":          time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
		"Jahr 10000 nach UTC": time.Date(9999, 12, 31, 23, 0, 0, 0, time.FixedZone("minus2", -2*3600)),
		"Jahr 0 nach UTC":     time.Date(1, 1, 1, 1, 0, 0, 0, plus2()),
	} {
		if !panikt(func() { supporttest.NewFestUhr(start) }) {
			t.Errorf("%s: keine Panik", name)
		}
	}
}

func TestFestUhrAdvance(t *testing.T) {
	start := zeit("2026-10-20T10:00:00Z")
	u := supporttest.NewFestUhr(start)
	if !u.Now().Equal(u.Now()) || !u.Now().Equal(start) {
		t.Fatal("Now ändert den Stand")
	}
	u.Advance(time.Nanosecond)
	if !u.Now().Equal(start.Add(1)) {
		t.Fatalf("Advance(1ns): %v", u.Now())
	}
	u.Advance(0)
	if !u.Now().Equal(start.Add(1)) {
		t.Fatalf("Advance(0) ändert den Stand: %v", u.Now())
	}
	if !panikt(func() { u.Advance(-1) }) || !u.Now().Equal(start.Add(1)) {
		t.Errorf("Advance(-1): keine Panik oder Stand geändert: %v", u.Now())
	}
	ende := supporttest.NewFestUhr(zeit("9999-12-31T23:59:59.999999999Z"))
	if !panikt(func() { ende.Advance(1) }) || !ende.Now().Equal(zeit("9999-12-31T23:59:59.999999999Z")) {
		t.Errorf("Advance über Jahr 9999: keine Panik oder Stand geändert: %v", ende.Now())
	}
	ende.Advance(0)
}

func TestFestUhrNebenlaeufig(t *testing.T) {
	start := zeit("2026-10-20T10:00:00Z")
	u := supporttest.NewFestUhr(start)
	var wg sync.WaitGroup
	for range 64 {
		wg.Go(func() {
			for range 1000 {
				u.Advance(time.Nanosecond)
			}
		})
	}
	wg.Wait()
	if want := start.Add(64000); !u.Now().Equal(want) {
		t.Errorf("Stand %v, erwartet %v", u.Now(), want)
	}
}

func ersteDrei(g interface{ SessionID() string }) []string {
	return []string{g.SessionID(), g.SessionID(), g.SessionID()}
}

func TestFestIDsFolge(t *testing.T) {
	g := supporttest.NewFestIDs()
	ops := map[string]func() string{
		"ses": g.SessionID, "exc": g.ExchangeID, "cor": g.KorrelationsID,
		"wsp": g.WorkspaceID, "snp": g.SnapshotID,
	}
	for praefix, op := range ops {
		for i, ende := range []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"} {
			if got, want := op(), praefix+"-000000000000000000000000000000"+ende; got != want {
				t.Errorf("%s Aufruf %d: %q, erwartet %q", praefix, i+1, got, want)
			}
		}
	}
	if got := supporttest.NewFestIDs().SessionID(); got != "ses-00000000000000000000000000000001" {
		t.Errorf("zweite Instanz beginnt mit %q", got)
	}
	if got := ersteDrei(supporttest.NewFestIDs()); got[2] != "ses-00000000000000000000000000000003" {
		t.Errorf("Folge %v", got)
	}
}

func TestFestIDsNebenlaeufig(t *testing.T) {
	g := supporttest.NewFestIDs()
	var mu sync.Mutex
	gesehen := map[string]bool{}
	var wg sync.WaitGroup
	for range 64 {
		wg.Go(func() {
			for range 500 {
				id := g.ExchangeID()
				mu.Lock()
				gesehen[id] = true
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	for i := 1; i <= 32000; i++ {
		if !gesehen[fmt.Sprintf("exc-%032d", i)] {
			t.Fatalf("Nummer %d fehlt", i)
		}
	}
	if len(gesehen) != 32000 {
		t.Errorf("%d verschiedene Kennungen, erwartet 32000", len(gesehen))
	}
}

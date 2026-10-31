package cli_test

import (
	"bytes"
	"io"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/cli"
)

// lauf-Szenario: ein Aufruf von `start` mit der Quelle Fake-Kanal oder echtes Signal.
type szenario struct {
	res    <-chan int
	start  *FakeStart
	sig    *FakeSignale
	senden func(syscall.Signal)
	out    *bytes.Buffer
	errw   *bytes.Buffer
}

// neuesSzenario baut den Aufruf; er läuft noch nicht. Echte Signale melden vorher einen eigenen Kanal an.
func neuesSzenario(t *testing.T, echt bool, start *FakeStart) *szenario {
	t.Helper()
	s := &szenario{start: start, out: &bytes.Buffer{}, errw: &bytes.Buffer{}}
	if echt {
		echteSignale(t)
		s.senden = func(g syscall.Signal) { sendeEcht(t, g) }
		return s
	}
	s.sig = NewFakeSignale()
	start.Signale = s.sig
	s.senden = func(g syscall.Signal) { s.sig.Senden(g) }
	return s
}

func (s *szenario) los() {
	k := cli.Konfig{Version: "1.2.3", Start: s.start.Func(), Stdout: s.out, Stderr: s.errw}
	if s.sig != nil {
		k.Signale = s.sig.Func()
	}
	s.res = lauf([]string{"start"}, k)
}

func quellenName(echt bool) string {
	if echt {
		return "echt"
	}
	return "Fake"
}

func zweiSignale() []syscall.Signal { return []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} }

func erwartetNachErzwingen(zweites syscall.Signal) int {
	if zweites == syscall.SIGINT {
		return 130
	}
	return 143
}

// C7: Die Anmeldung geschieht genau einmal je start, vor dem Aufruf von Start, und sonst nie.
func TestAnmeldung(t *testing.T) {
	pfade := []struct {
		name     string
		args     []string
		startNil bool
		want     int
	}{
		{"start", []string{"start"}, false, 1},
		{"ohne Argument", nil, false, 0},
		{"version", []string{"version"}, false, 0},
		{"Verstoß", []string{"foo"}, false, 0},
		{"start bei nil", []string{"start"}, true, 0},
	}
	if len(pfade) != 5 {
		t.Fatalf("Menge hat %d Elemente, erwartet 5", len(pfade))
	}
	for _, p := range pfade {
		t.Run(p.name, func(t *testing.T) {
			start := NewFakeStart(t, 0)
			sig := NewFakeSignale()
			sig.Verzoegerung = fenster
			start.Signale = sig
			k := konfig("1.2.3", start, sig, nil, nil)
			if p.startNil {
				k.Start = nil
			}
			cli.Run(p.args, k)
			if got := sig.Anmeldungen(); got != p.want {
				t.Errorf("Anmeldungen %d, erwartet %d", got, p.want)
			}
			if z := start.Zustand(); p.want == 1 && (z.Aufrufe != 1 || z.AnmBeiAufruf[0] != 1) {
				t.Errorf("Start sah %v Anmeldungen bei %d Aufrufen, erwartet 1 bei 1", z.AnmBeiAufruf, z.Aufrufe)
			}
		})
	}
}

// C8: Das erste Signal bricht den Kontext ab, und nicht davor; Start kehrt mit 0 zurück, Run liefert 0.
// Menge: Quelle × Signal × Zeitpunkt, Fake {SIGTERM, SIGINT, SIGHUP als fremdes} × 2 = 6 und
// echt {SIGTERM, SIGINT} nur im Lauf = 2; zusammen 8.
func TestErstesSignal(t *testing.T) {
	n := 0
	for _, echt := range []bool{false, true} {
		for _, g := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP} {
			for _, vorbelegt := range []bool{true, false} {
				if echt && (vorbelegt || g == syscall.SIGHUP) {
					continue
				}
				n++
				name := quellenName(echt) + "/" + g.String()
				if vorbelegt {
					name += "/vorbelegt"
				}
				t.Run(name, func(t *testing.T) {
					start := NewFakeStart(t, 0)
					start.WarteKontext = true
					s := neuesSzenario(t, echt, start)
					if vorbelegt {
						s.senden(g)
						s.los()
						erwarte(t, start.Abgebrochen, "Kontext abgebrochen")
						if z := start.Zustand(); z.Aufrufe != 1 || z.ErrBeiAufruf[0] == nil {
							t.Errorf("Start sah beim Aufruf Err %v, erwartet den Abbruch", z.ErrBeiAufruf)
						}
					} else {
						s.los()
						erwarte(t, start.Bereit, "Start bereit")
						z := start.Zustand()
						if z.ErrBeiAufruf[0] != nil || z.Ctx[0].Err() != nil {
							t.Fatalf("Kontext vor dem Signal abgebrochen: %v, %v", z.ErrBeiAufruf[0], z.Ctx[0].Err())
						}
						bleibtAus(t, start.Abgebrochen, "Kontext ohne Signal")
						bleibtAus(t, s.res, "Run ohne Signal")
						s.senden(g)
						erwarte(t, start.Abgebrochen, "Kontext abgebrochen")
					}
					if code := erwarte(t, s.res, "Run kehrt zurück"); code != 0 {
						t.Errorf("Code %d, erwartet 0", code)
					}
					if s.out.Len() != 0 || s.errw.Len() != 0 {
						t.Errorf("Ausgabe %q %q, erwartet leer", s.out.String(), s.errw.String())
					}
				})
			}
		}
	}
	if n != 8 {
		t.Fatalf("Menge hat %d Elemente, erwartet 8", n)
	}
}

// C9: Nach dem ersten Signal wartet Run auf Start und liefert dessen Code. Menge: Rückgabe {0, 1, 2} × Signal = 6.
func TestWartetAufStart(t *testing.T) {
	n := 0
	for _, r := range []int{0, 1, 2} {
		for _, g := range zweiSignale() {
			n++
			t.Run(g.String()+"/"+string(rune('0'+r)), func(t *testing.T) {
				start := NewFakeStart(t, r)
				start.WarteKontext, start.Halten = true, true
				s := neuesSzenario(t, false, start)
				s.los()
				erwarte(t, start.Bereit, "Start bereit")
				s.senden(g)
				erwarte(t, start.Abgebrochen, "Kontext abgebrochen")
				bleibtAus(t, s.res, "Run kehrt vor der Rückkehr von Start zurück")
				if z := start.Zustand(); z.Rueckkehren != 0 {
					t.Fatalf("Start schon zurückgekehrt: %d", z.Rueckkehren)
				}
				start.Frei()
				want := 1
				if r == 0 {
					want = 0
				}
				if code := erwarte(t, s.res, "Run kehrt zurück"); code != want {
					t.Errorf("Code %d, erwartet %d", code, want)
				}
			})
		}
	}
	if n != 6 {
		t.Fatalf("Menge hat %d Elemente, erwartet 6", n)
	}
}

// C10: Das zweite Signal erzwingt das Ende mit 130 (SIGINT) und sonst 143, ohne Warten und ohne Ausgabe.
// Menge: erstes × zweites {SIGTERM, SIGINT, SIGHUP als fremdes} × Quelle: Fake 6, echt (ohne SIGHUP) 4 = 10.
func TestZweitesSignal(t *testing.T) {
	n := 0
	for _, echt := range []bool{false, true} {
		for _, erstes := range zweiSignale() {
			for _, zweites := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP} {
				if echt && zweites == syscall.SIGHUP {
					continue
				}
				n++
				t.Run(quellenName(echt)+"/"+erstes.String()+"/"+zweites.String(), func(t *testing.T) {
					start := NewFakeStart(t, 0)
					start.WarteKontext, start.Halten = true, true
					s := neuesSzenario(t, echt, start)
					s.los()
					erwarte(t, start.Bereit, "Start bereit")
					s.senden(erstes)
					erwarte(t, start.Abgebrochen, "Kontext abgebrochen")
					s.senden(zweites)
					if code, want := erwarte(t, s.res, "Run erzwingt das Ende"), erwartetNachErzwingen(zweites); code != want {
						t.Errorf("Code %d, erwartet %d", code, want)
					}
					if z := start.Zustand(); z.Rueckkehren != 0 {
						t.Errorf("Run wartete auf Start: Rückkehren %d", z.Rueckkehren)
					}
					if s.out.Len() != 0 || s.errw.Len() != 0 {
						t.Errorf("Ausgabe %q %q, erwartet leer", s.out.String(), s.errw.String())
					}
				})
			}
		}
	}
	if n != 10 {
		t.Fatalf("Menge hat %d Elemente, erwartet 10", n)
	}
}

// C11: Die Abmeldung geschieht genau einmal auf jedem Pfad nach der Anmeldung und nie vor der Rückkehr
// von Start. Menge: 6 Pfade mit einer Abmeldung, dazu 3 ohne Anmeldung = 9.
func TestAbmeldung(t *testing.T) {
	pruefe := func(t *testing.T, s *szenario, start *FakeStart) {
		t.Helper()
		<-time.After(fenster)
		if got := s.sig.Abmeldungen(); got != 0 {
			t.Fatalf("Abmeldungen %d vor der Rückkehr von Start, erwartet 0", got)
		}
		start.Frei()
	}
	nachher := func(t *testing.T, s *szenario, start *FakeStart, erzwungen bool) {
		t.Helper()
		if s.sig.Anmeldungen() != 1 || s.sig.Abmeldungen() != 1 {
			t.Errorf("Anmeldungen %d, Abmeldungen %d, erwartet 1, 1", s.sig.Anmeldungen(), s.sig.Abmeldungen())
		}
		z := start.Zustand()
		if z.Aufrufe != 1 || z.AbmBeiAufruf[0] != 0 {
			t.Errorf("Start sah Abmeldungen %v bei Aufrufen %d, erwartet 0 bei 1", z.AbmBeiAufruf, z.Aufrufe)
		}
		if !erzwungen && (z.Rueckkehren != 1 || z.AbmBeiRueckkehr[0] != 0) {
			t.Errorf("Start sah Abmeldungen %v bei Rückkehren %d, erwartet 0 bei 1", z.AbmBeiRueckkehr, z.Rueckkehren)
		}
	}
	n := 0
	for _, r := range []int{0, 1, 2} {
		n++
		t.Run("Start liefert "+string(rune('0'+r)), func(t *testing.T) {
			start := NewFakeStart(t, r)
			start.Halten = true
			s := neuesSzenario(t, false, start)
			s.los()
			erwarte(t, start.Bereit, "Start bereit")
			pruefe(t, s, start)
			erwarte(t, s.res, "Run kehrt zurück")
			nachher(t, s, start, false)
		})
	}
	for _, r := range []int{0, 1} {
		n++
		t.Run("erstes Signal, dann "+string(rune('0'+r)), func(t *testing.T) {
			start := NewFakeStart(t, r)
			start.WarteKontext, start.Halten = true, true
			s := neuesSzenario(t, false, start)
			s.los()
			erwarte(t, start.Bereit, "Start bereit")
			s.senden(syscall.SIGTERM)
			erwarte(t, start.Abgebrochen, "Kontext abgebrochen")
			pruefe(t, s, start)
			erwarte(t, s.res, "Run kehrt zurück")
			nachher(t, s, start, false)
		})
	}
	n++
	t.Run("Erzwingen", func(t *testing.T) {
		start := NewFakeStart(t, 0)
		start.WarteKontext, start.Halten = true, true
		s := neuesSzenario(t, false, start)
		s.los()
		erwarte(t, start.Bereit, "Start bereit")
		s.senden(syscall.SIGTERM)
		erwarte(t, start.Abgebrochen, "Kontext abgebrochen")
		s.senden(syscall.SIGINT)
		erwarte(t, s.res, "Run erzwingt das Ende")
		nachher(t, s, start, true)
	})
	ohne := []struct {
		name     string
		args     []string
		startNil bool
	}{{"version", []string{"version"}, false}, {"Verstoß", []string{"foo"}, false}, {"start bei nil", []string{"start"}, true}}
	for _, p := range ohne {
		n++
		t.Run(p.name, func(t *testing.T) {
			sig := NewFakeSignale()
			k := konfig("1.2.3", NewFakeStart(t, 0), sig, nil, nil)
			if p.startNil {
				k.Start = nil
			}
			cli.Run(p.args, k)
			if sig.Anmeldungen() != 0 || sig.Abmeldungen() != 0 {
				t.Errorf("Anmeldungen %d, Abmeldungen %d, erwartet 0, 0", sig.Anmeldungen(), sig.Abmeldungen())
			}
		})
	}
	if n != 9 {
		t.Fatalf("Menge hat %d Elemente, erwartet 9", n)
	}
}

// lauf-Arten für C12; jede liefert den Code und prüft den Zustand des einen Aufrufs.
func fahreNormal(t *testing.T) {
	t.Helper()
	start := NewFakeStart(t, 1)
	sig := NewFakeSignale()
	start.Signale = sig
	code := cli.Run([]string{"start"}, konfig("1.2.3", start, sig, io.Discard, io.Discard))
	z := start.Zustand()
	if code != 1 || z.Aufrufe != 1 || z.ErrBeiAufruf[0] != nil || sig.Anmeldungen() != 1 || sig.Abmeldungen() != 1 {
		t.Errorf("normaler Lauf: Code %d, Aufrufe %d, Err %v, Anmeldungen %d, Abmeldungen %d; erwartet 1, 1, nil, 1, 1",
			code, z.Aufrufe, z.ErrBeiAufruf, sig.Anmeldungen(), sig.Abmeldungen())
	}
}

func fahreSignal(t *testing.T) {
	t.Helper()
	start := NewFakeStart(t, 0)
	start.WarteKontext = true
	s := neuesSzenario(t, false, start)
	s.los()
	erwarte(t, start.Bereit, "Start bereit")
	if z := start.Zustand(); z.ErrBeiAufruf[0] != nil || z.Ctx[0].Err() != nil || s.sig.Anmeldungen() != 1 {
		t.Errorf("Signal-Lauf: Err %v, Anmeldungen %d; erwartet nil, 1", z.ErrBeiAufruf, s.sig.Anmeldungen())
	}
	s.senden(syscall.SIGTERM)
	if code := erwarte(t, s.res, "Run kehrt zurück"); code != 0 {
		t.Errorf("Signal-Lauf: Code %d, erwartet 0", code)
	}
}

func fahreErzwingen(t *testing.T) {
	t.Helper()
	start := NewFakeStart(t, 0)
	start.WarteKontext, start.Halten = true, true
	s := neuesSzenario(t, false, start)
	s.los()
	erwarte(t, start.Bereit, "Start bereit")
	if z := start.Zustand(); z.ErrBeiAufruf[0] != nil || z.Ctx[0].Err() != nil || s.sig.Anmeldungen() != 1 {
		t.Errorf("Erzwingen-Lauf: Err %v, Anmeldungen %d; erwartet nil, 1", z.ErrBeiAufruf, s.sig.Anmeldungen())
	}
	s.senden(syscall.SIGTERM)
	erwarte(t, start.Abgebrochen, "Kontext abgebrochen")
	s.senden(syscall.SIGINT)
	if code := erwarte(t, s.res, "Run erzwingt das Ende"); code != 130 {
		t.Errorf("Erzwingen-Lauf: Code %d, erwartet 130", code)
	}
}

// C12: Kein Zustand zwischen Aufrufen. Menge: vier Paare nacheinander im selben Prozess.
func TestZweiAufrufe(t *testing.T) {
	paare := []struct {
		name            string
		erster, zweiter func(*testing.T)
	}{
		{"Erzwingen, normal", fahreErzwingen, fahreNormal},
		{"Signal, normal", fahreSignal, fahreNormal},
		{"normal, Signal", fahreNormal, fahreSignal},
		{"Erzwingen, Erzwingen", fahreErzwingen, fahreErzwingen},
	}
	if len(paare) != 4 {
		t.Fatalf("Menge hat %d Elemente, erwartet 4", len(paare))
	}
	for _, p := range paare {
		t.Run(p.name, func(t *testing.T) {
			p.erster(t)
			p.zweiter(t)
		})
	}
}

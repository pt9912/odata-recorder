package hybrid_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// mitBody liefert eine Anfrage mit dem Leser s als Body.
func mitBody(methode string, s io.Reader, laenge int64) inbound.Anfrage {
	a := anfrage(methode)
	a.Body, a.Laenge = s, laenge
	return a
}

// TestOhnePufferung hält Z12: der Body bleibt ungelesen und geht als derselbe Leser mit derselben Laenge an M, wenn
// die Anfrage nicht fallback-fähig ist (alle Schlüssel leer, nie freigegebene Methode, nur eine Host-Freigabe ohne
// Replay-Kategorie), keinen Body hat (nil, Laenge 0 mit Leser) oder Laenge größer MaxBytes hat.
func TestOhnePufferung(t *testing.T) {
	hostOhneKategorie := neuePolicy(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"POST /**"}})
	faelle := []struct {
		name    string
		pol     func() *policy.Policy
		methode string
		laenge  int64
		inhalt  int
	}{
		{"alle Schlüssel leer, POST", func() *policy.Policy { return mitKategorien(t) }, "POST", 4, 4},
		{"alle Schlüssel leer, GET", func() *policy.Policy { return mitKategorien(t) }, "GET", -1, 4},
		{"nie freigegebene Methode OPTIONS", func() *policy.Policy { return offenste(t) }, "OPTIONS", 4, 4},
		{"nie freigegebene Methode get", func() *policy.Policy { return offenste(t) }, "get", -1, 4},
		{"nur Host-Freigabe, POST", func() *policy.Policy { return hostOhneKategorie }, "POST", 4, 4},
		{"nur Host-Freigabe, GET", func() *policy.Policy { return hostOhneKategorie }, "GET", -1, 4},
		{"Laenge 0 mit Leser", func() *policy.Policy { return offenste(t) }, "POST", 0, 4},
		{"Laenge größer MaxBytes", func() *policy.Policy { return offenste(t) }, "POST", standardMax + 1, standardMax + 1},
		{"Laenge sehr groß", func() *policy.Policy { return offenste(t) }, "POST", math.MaxInt64, 4},
	}
	for _, f := range faelle {
		g := neuesRig(t, f.pol(), standardMax)
		g.mockFehler(code(codeNamed(t, "ODR-E6003")))
		s := neuerStrom(muster(f.inhalt, 3))
		_, _ = g.v.Hybrid(context.Background(), mitBody(f.methode, s, f.laenge))
		got := g.m.erster(t).a
		if got.Body != io.Reader(s) || got.Laenge != f.laenge || s.reads.Load() != 0 {
			t.Errorf("%s: Body unverändert erwartet, Reads %d, Laenge %d", f.name, s.reads.Load(), got.Laenge)
		}
	}
	g := neuesRig(t, offenste(t), standardMax)
	g.mockFehler(code(codeNamed(t, "ODR-E6003")))
	g.replayAntwort(antwortVon(200, nil))
	for _, l := range []int64{-1, 0} {
		a := anfrage("POST")
		a.Laenge = l
		if _, err := g.v.Hybrid(context.Background(), a); err != nil {
			t.Errorf("ohne Body, Laenge %d: %v", l, err)
		}
	}
	for i := 0; i < 2; i++ {
		if g.m.aufrufe[i].a.Body != nil {
			t.Errorf("ohne Body: Body nicht nil")
		}
	}
}

// pufferPolicies sind die Policies, unter denen eine Anfrage fallback-fähig ist (jede Menge der Kategorien außer der
// leeren).
func pufferPolicies(t *testing.T) map[string]*policy.Policy {
	return map[string]*policy.Policy{
		"unsupported": mitKategorien(t, katUnsupported),
		"coverage":    mitKategorien(t, katCoverage),
		"beide":       mitKategorien(t, katUnsupported, katCoverage),
		"offenste":    offenste(t),
	}
}

// TestPufferung hält Z13: M und R erhalten je einen eigenen vollständigen Leser über dieselben Bytes mit Laenge gleich
// ihrer Länge, auch wenn M nichts, ein Byte oder alles las, bei deklarierter Laenge und bei -1, für jede Policy mit
// Fähigkeit und mehrere Längen.
func TestPufferung(t *testing.T) {
	gezaehlt := 0
	for pname, pol := range pufferPolicies(t) {
		for _, lies := range []int{liesNichts, liesEinByte, liesAlles} {
			for _, deklariert := range []bool{true, false} {
				for _, n := range []int{1, 2, 5, standardMax} {
					gezaehlt++
					name := pname
					g := neuesRig(t, pol, standardMax)
					g.m.lies = lies
					if pname == "coverage" {
						g.mockFehler(code(codeNamed(t, "ODR-E6003")))
					} else {
						g.mockFehler(code(codeNamed(t, "ODR-E6002")))
					}
					g.r.lies = liesAlles
					g.replayAntwort(antwortVon(200, nil))
					soll := muster(n, n)
					laenge := int64(-1)
					if deklariert {
						laenge = int64(n)
					}
					orig := neuerStrom(bytes.Clone(soll))
					if _, err := g.v.Hybrid(context.Background(), mitBody("POST", orig, laenge)); err != nil {
						t.Errorf("%s lies %d n %d: %v", name, lies, n, err)
						continue
					}
					pruefeStufeBody(t, name, "M", g.m.erster(t), orig, soll, lies)
					pruefeStufeBody(t, name, "R", g.r.erster(t), orig, soll, liesAlles)
					if g.m.erster(t).a.Body == g.r.erster(t).a.Body {
						t.Errorf("%s: M und R teilen den Leser", name)
					}
				}
			}
		}
	}
	if gezaehlt != 4*3*2*4 {
		t.Errorf("Zählprüfung: %d Fälle", gezaehlt)
	}
}

func pruefeStufeBody(t *testing.T, pname, stufe string, a aufruf, orig *strom, soll []byte, lies int) {
	t.Helper()
	erwartet := soll
	switch lies {
	case liesNichts:
		erwartet = nil
	case liesEinByte:
		erwartet = soll[:1]
	}
	if !bytes.Equal(a.gelesen, erwartet) || a.leseErr != nil {
		t.Errorf("%s %s: gelesen %v (%v), erwartet %v", pname, stufe, a.gelesen, a.leseErr, erwartet)
	}
	if a.a.Laenge != int64(len(soll)) || a.a.Body == io.Reader(orig) {
		t.Errorf("%s %s: Laenge %d, eigener Leser %v", pname, stufe, a.a.Laenge, a.a.Body != io.Reader(orig))
	}
	if orig.closes.Load() != 0 {
		t.Errorf("%s %s: Body der Anfrage geschlossen", pname, stufe)
	}
}

// TestPufferungVollstaendig hält Z13 für M: der Leser von M liefert, wenn er ganz gelesen wird, alle Bytes in der
// Reihenfolge, und der von R danach ebenfalls; beide Leser sind unabhängig.
func TestPufferungVollstaendig(t *testing.T) {
	g := neuesRig(t, offenste(t), standardMax)
	g.m.lies = liesAlles
	g.r.lies = liesAlles
	g.mockFehler(code(codeNamed(t, "ODR-E6003")))
	g.replayAntwort(antwortVon(200, nil))
	soll := muster(standardMax, 9)
	if _, err := g.v.Hybrid(context.Background(), mitBody("PUT", neuerStrom(bytes.Clone(soll)), -1)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(g.m.erster(t).gelesen, soll) || !bytes.Equal(g.r.erster(t).gelesen, soll) {
		t.Errorf("Bytes: M %v, R %v, erwartet %v", g.m.erster(t).gelesen, g.r.erster(t).gelesen, soll)
	}
}

// TestGrenze hält Z14: genau MaxBytes Bytes werden gepuffert (deklariert und -1) und fallen zurück; MaxBytes + 1 und
// mehr, auch gelogen deklariert, geben keinen Fallback: M erhält alle Bytes in Reihenfolge, die Laenge der Anfrage
// und sein Ergebnis kehrt zurück, R läuft nicht. MaxBytes 1 und math.MaxInt64 eingeschlossen.
func TestGrenze(t *testing.T) {
	for _, max := range []int64{1, 2, standardMax, math.MaxInt64} {
		laengen := []int64{1, max - 1, max, max + 1, max + 2, 2*max + 3}
		if max == math.MaxInt64 {
			laengen = []int64{1, 5, 40}
		}
		for _, n := range laengen {
			if n < 1 {
				continue
			}
			for _, modus := range []string{"deklariert", "-1", "gelogen"} {
				laenge := n
				switch modus {
				case "-1":
					laenge = -1
				case "gelogen":
					laenge = min(n, max)
				}
				g := neuesRig(t, offenste(t), max)
				g.m.lies = liesAlles
				g.r.lies = liesAlles
				fehler := code(codeNamed(t, "ODR-E6003"))
				g.mockFehler(fehler)
				g.replayAntwort(antwortVon(200, nil))
				soll := muster(int(n), int(n))
				_, err := g.v.Hybrid(context.Background(), mitBody("POST", bytes.NewReader(bytes.Clone(soll)), laenge))
				m := g.m.erster(t)
				if !bytes.Equal(m.gelesen, soll) || m.leseErr != nil {
					t.Errorf("max %d n %d %s: M erhielt %d Bytes, erwartet %d", max, n, modus, len(m.gelesen), n)
				}
				if n <= max {
					if err != nil || g.r.zahl() != 1 || !bytes.Equal(g.r.erster(t).gelesen, soll) || m.a.Laenge != n {
						t.Errorf("max %d n %d %s: Fallback erwartet, Fehler %v, Replay %d", max, n, modus, err, g.r.zahl())
					}
					continue
				}
				if err != error(fehler) || g.r.zahl() != 0 {
					t.Errorf("max %d n %d %s: kein Fallback erwartet, Fehler %v, Replay %d", max, n, modus, err, g.r.zahl())
				}
				if wantL := map[string]int64{"deklariert": n, "-1": -1, "gelogen": max}[modus]; m.a.Laenge != wantL {
					t.Errorf("max %d n %d %s: Laenge %d, erwartet %d", max, n, modus, m.a.Laenge, wantL)
				}
			}
		}
	}
}

// TestLesefehler hält Z15: ein Lesefehler beim Puffern (nicht io.EOF; auch ein Kontextfehler) nach 0, 3 oder MaxBytes
// Bytes: kein Fallback, M erhält die gelesenen Bytes und dann denselben Fehler, sein Ergebnis kehrt unverändert zurück.
func TestLesefehler(t *testing.T) {
	fehlerArten := []error{errors.New("kaputt"), context.Canceled, io.ErrUnexpectedEOF, io.ErrClosedPipe}
	for _, le := range fehlerArten {
		for _, k := range []int{0, 3, standardMax} {
			for _, laenge := range []int64{-1, standardMax} {
				g := neuesRig(t, offenste(t), standardMax)
				g.m.lies = liesAlles
				fehler := code(codeNamed(t, "ODR-E6003"))
				g.mockFehler(fehler)
				g.replayAntwort(antwortVon(200, nil))
				soll := muster(k, 5)
				body := &fehlerNach{rest: bytes.Clone(soll), err: le}
				_, err := g.v.Hybrid(context.Background(), mitBody("POST", body, laenge))
				m := g.m.erster(t)
				if err != error(fehler) || g.r.zahl() != 0 || !bytes.Equal(m.gelesen, soll) || m.leseErr != le {
					t.Errorf("Fehler %v nach %d Bytes, Laenge %d: Ergebnis %v, Replay %d, M las %d Bytes mit %v",
						le, k, laenge, err, g.r.zahl(), len(m.gelesen), m.leseErr)
				}
				if m.a.Laenge != laenge {
					t.Errorf("Laenge %d, erwartet %d", m.a.Laenge, laenge)
				}
			}
		}
	}
}

// TestBodyNieGeschlossen hält Z16: der Dienst schließt den Body der Anfrage nie, in keinem Pfad (Erfolg, Replay,
// Fehler von R, Fehler ohne Fallback, Abbruch, Überlauf, Lesefehler), auch nicht per Typumwandlung zu io.Closer.
func TestBodyNieGeschlossen(t *testing.T) {
	e6003 := code(codeNamed(t, "ODR-E6003"))
	pfade := map[string]func(g *rig) context.Context{
		"Erfolg": func(g *rig) context.Context {
			g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
				return inbound.Antwort{Status: 200}, nil
			}
			return context.Background()
		},
		"Replay": func(g *rig) context.Context {
			g.mockFehler(e6003)
			g.replayAntwort(antwortVon(200, nil))
			return context.Background()
		},
		"Fehler von R": func(g *rig) context.Context {
			g.mockFehler(e6003)
			g.replayFehler(code(codeNamed(t, "ODR-E5001")))
			return context.Background()
		},
		"Fehler ohne Fallback": func(g *rig) context.Context {
			g.mockFehler(code(codeNamed(t, "ODR-E6005")))
			return context.Background()
		},
		"Abbruch": func(g *rig) context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
				cancel()
				return inbound.Antwort{}, e6003
			}
			return ctx
		},
	}
	for name, einrichten := range pfade {
		for _, laenge := range []int64{4, -1} {
			g := neuesRig(t, offenste(t), standardMax)
			ctx := einrichten(g)
			s := neuerStrom(muster(4, 2))
			_, _ = g.v.Hybrid(ctx, mitBody("POST", s, laenge))
			if s.closes.Load() != 0 {
				t.Errorf("%s (Laenge %d): Body der Anfrage %d-mal geschlossen", name, laenge, s.closes.Load())
			}
		}
	}
	g := neuesRig(t, offenste(t), standardMax)
	g.mockFehler(e6003)
	s := neuerStrom(muster(standardMax+3, 1))
	_, _ = g.v.Hybrid(context.Background(), mitBody("POST", s, -1))
	if s.closes.Load() != 0 {
		t.Errorf("Überlauf: Body der Anfrage geschlossen")
	}
}

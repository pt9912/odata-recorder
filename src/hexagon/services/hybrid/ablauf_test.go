package hybrid_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// ausloeser ist das Orakel der Auslöser (SPEC-261 Nr. 3 a) mit den Namen aus der Spezifikation, unabhängig von der
// Policy.
func ausloeserKategorie(name string) string {
	switch name {
	case "ODR-E6001", "ODR-E6002":
		return katUnsupported
	case "ODR-E6003":
		return katCoverage
	}
	return ""
}

func antwortVon(status int, body []byte) inbound.Antwort {
	return inbound.Antwort{Status: status, Header: map[string][]string{"Content-Type": {"application/json"}}, Body: neuerStrom(body)}
}

// TestErfolgUnveraendert hält Z03 (Erfolg): ein Erfolg des Mock geht ohne Kennzeichnung und als derselbe Wert aus,
// auch mit Status 404 und 500; weder Replay noch Close.
func TestErfolgUnveraendert(t *testing.T) {
	for _, status := range []int{200, 201, 204, 404, 409, 412, 500, 501} {
		g := neuesRig(t, offenste(t), standardMax)
		want := antwortVon(status, []byte("x"))
		g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return want, nil }
		got, err := g.v.Hybrid(context.Background(), anfrage("GET"))
		if err != nil || got.Status != status || got.Body != want.Body || zeiger(got.Header) != zeiger(want.Header) {
			t.Errorf("Status %d: Antwort %+v, Fehler %v", status, got, err)
		}
		if k := kennzeichen(got); len(k) != 0 {
			t.Errorf("Status %d: Kennzeichnung %v bei Erfolg", status, k)
		}
		g.zaehlen(t, "Erfolg", 1, 0)
		if c := want.Body.(*strom).closes.Load(); c != 0 {
			t.Errorf("Status %d: ausgelieferte Antwort %d-mal geschlossen", status, c)
		}
	}
}

// TestCodesSchleife hält Z03 (Codes): von jedem Code in model.AlleCodes(), als Fehler des Mock und mit jedem seiner
// Status und drei fremden, lösen unter der offensten Policy genau ODR-E6001, ODR-E6002 und ODR-E6003 den Replay aus;
// jeder andere Fehler kehrt als derselbe Wert zurück. Die Zählprüfung nennt die Menge.
func TestCodesSchleife(t *testing.T) {
	alle := model.AlleCodes()
	if len(alle) < 40 {
		t.Fatalf("model.AlleCodes() liefert %d Codes, erwartet mindestens 40", len(alle))
	}
	for _, n := range []string{"ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E5001", "ODR-E6006"} {
		codeNamed(t, n)
	}
	gezaehlt, replays, soll, sollReplays := 0, 0, 0, 0
	for _, c := range alle {
		stati := slices.Concat(c.Stati(), []int{200, 404, 500})
		soll += len(stati)
		if ausloeserKategorie(c.String()) != "" {
			sollReplays += len(stati)
		}
		for _, st := range stati {
			gezaehlt++
			g := neuesRig(t, offenste(t), standardMax)
			fehler := &inbound.Fehler{Code: c, Status: st}
			g.mockFehler(fehler)
			g.replayAntwort(antwortVon(200, nil))
			got, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			if ausloeserKategorie(c.String()) != "" {
				replays++
				if err != nil || g.r.zahl() != 1 || g.m.zahl() != 1 || !reflect.DeepEqual(kennzeichen(got), []string{"replay"}) {
					t.Errorf("%s/%d: Replay erwartet, Fehler %v, Mock %d, Replay %d", c, st, err, g.m.zahl(), g.r.zahl())
				}
				continue
			}
			if err != error(fehler) || g.r.zahl() != 0 || g.m.zahl() != 1 {
				t.Errorf("%s/%d: derselbe Fehler erwartet, Fehler %v, Mock %d, Replay %d", c, st, err, g.m.zahl(), g.r.zahl())
			}
		}
	}
	if gezaehlt != soll || replays != sollReplays || sollReplays < 9 {
		t.Errorf("Zählprüfung: %d von %d Fällen, %d von %d Replays", gezaehlt, soll, replays, sollReplays)
	}
}

// TestKategorieOrakel hält Z06: ODR-E6001 und ODR-E6002 (unsupported) und ODR-E6003 (coverage_insufficient) × jede
// Menge von Kategorien (keine, eine, die andere, beide) lösen den Replay genau dann aus, wenn die Kategorie ihres
// Codes gesetzt ist; die Antwort von R geht mit Kennzeichnung replay und unverändertem Status und Body aus, R erhält
// die Anfrage unverändert.
func TestKategorieOrakel(t *testing.T) {
	mengen := [][]string{{}, {katUnsupported}, {katCoverage}, {katUnsupported, katCoverage}}
	gezaehlt := 0
	for _, name := range []string{"ODR-E6001", "ODR-E6002", "ODR-E6003"} {
		for _, menge := range mengen {
			gezaehlt++
			g := neuesRig(t, mitKategorien(t, menge...), standardMax)
			fehler := code(codeNamed(t, name))
			g.mockFehler(fehler)
			want := antwortVon(206, []byte("aufgezeichnet"))
			g.replayAntwort(want)
			a := anfrage("GET")
			got, err := g.v.Hybrid(context.Background(), a)
			soll := false
			for _, k := range menge {
				soll = soll || k == ausloeserKategorie(name)
			}
			if !soll {
				if err != error(fehler) || g.r.zahl() != 0 {
					t.Errorf("%s %v: Fehler des Mock erwartet, Fehler %v, Replay %d", name, menge, err, g.r.zahl())
				}
				continue
			}
			if err != nil || g.r.zahl() != 1 || g.m.zahl() != 1 || got.Status != 206 || got.Body != want.Body {
				t.Errorf("%s %v: Replay erwartet, Fehler %v, Antwort %+v", name, menge, err, got)
				continue
			}
			if !reflect.DeepEqual(kennzeichen(got), []string{"replay"}) {
				t.Errorf("%s %v: Kennzeichnung %v", name, menge, kennzeichen(got))
			}
			r := g.r.erster(t).a
			if r.Methode != a.Methode || r.Pfad != a.Pfad || r.RawQuery != a.RawQuery || r.KorrelationsID != a.KorrelationsID ||
				r.Laenge != a.Laenge || zeiger(r.Header) != zeiger(a.Header) {
				t.Errorf("%s %v: Replay erhielt eine andere Anfrage: %+v", name, menge, r)
			}
		}
	}
	if gezaehlt != 12 {
		t.Errorf("Zählprüfung: %d Fälle, erwartet 3 × 4", gezaehlt)
	}
}

// TestFehlerOhneCode hält Z04: ein Fehler des Mock ohne Code oder mit Kontextfehler kehrt unverändert zurück, R läuft
// nicht, auch nicht unter der offensten Policy.
func TestFehlerOhneCode(t *testing.T) {
	e6003 := code(codeNamed(t, "ODR-E6003"))
	formen := map[string]error{
		"errors.New":               errors.New("irgendwas"),
		"ErrUngueltig":             model.ErrUngueltig,
		"Fehler Nullwert":          &inbound.Fehler{},
		"Canceled":                 context.Canceled,
		"DeadlineExceeded":         context.DeadlineExceeded,
		"Canceled verpackt":        fmt.Errorf("mock: %w", context.Canceled),
		"Join mit E6003":           errors.Join(e6003, context.Canceled),
		"Join Deadline mit E6003":  errors.Join(errors.New("x"), e6003, context.DeadlineExceeded),
		"E6003 mit Ursache Cancel": &inbound.Fehler{Code: codeNamed(t, "ODR-E6003"), Status: 501, Ursache: context.Canceled},
		"E6003 mit Ursache Zeit":   &inbound.Fehler{Code: codeNamed(t, "ODR-E6003"), Status: 501, Ursache: fmt.Errorf("w: %w", context.DeadlineExceeded)},
	}
	for name, fehler := range formen {
		g := neuesRig(t, offenste(t), standardMax)
		g.mockFehler(fehler)
		g.replayAntwort(antwortVon(200, nil))
		_, err := g.v.Hybrid(context.Background(), anfrage("GET"))
		if !sameErr(err, fehler) || g.r.zahl() != 0 || g.m.zahl() != 1 {
			t.Errorf("%s: Fehler gleich? %v, Mock %d, Replay %d", name, sameErr(err, fehler), g.m.zahl(), g.r.zahl())
		}
	}
	if len(formen) != 10 {
		t.Errorf("Zählprüfung: %d Formen", len(formen))
	}
}

// sameErr vergleicht zwei Fehler als Werte, ohne sie zu formatieren.
func sameErr(a, b error) bool { return a == b }

// TestEntscheideArgumente hält Z05: die Entscheidung hängt an Methode und Code. Methode: bekannte Methoden lösen den
// Replay aus, unbekannte (auch klein geschrieben) nie; Code: eine Policy mit nur coverage_insufficient lässt
// ODR-E6002 bleiben, eine mit nur unsupported ODR-E6003. Der Pfad erreicht die Entscheidung in diesem Teil nicht:
// der Replay-Zweig liest ihn nie (Plan, Zeile Z05).
func TestEntscheideArgumente(t *testing.T) {
	bekannt := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE"}
	unbekannt := []string{"OPTIONS", "TRACE", "get", "Post", "PATCHX", "CONNECT"}
	for _, m := range append(append([]string{}, bekannt...), unbekannt...) {
		g := neuesRig(t, mitKategorien(t, katUnsupported), standardMax)
		fehler := code(codeNamed(t, "ODR-E6002"))
		g.mockFehler(fehler)
		g.replayAntwort(antwortVon(200, nil))
		_, err := g.v.Hybrid(context.Background(), anfrage(m))
		istBekannt := false
		for _, b := range bekannt {
			istBekannt = istBekannt || b == m
		}
		if istBekannt != (g.r.zahl() == 1) || istBekannt != (err == nil) {
			t.Errorf("Methode %q: Replay %d, Fehler %v", m, g.r.zahl(), err)
		}
	}
	for _, c := range []struct{ kat, name string }{
		{katCoverage, "ODR-E6001"}, {katCoverage, "ODR-E6002"}, {katUnsupported, "ODR-E6003"},
	} {
		g := neuesRig(t, mitKategorien(t, c.kat), standardMax)
		g.mockFehler(code(codeNamed(t, c.name)))
		g.replayAntwort(antwortVon(200, nil))
		if _, err := g.v.Hybrid(context.Background(), anfrage("GET")); err == nil || g.r.zahl() != 0 {
			t.Errorf("Kategorie %s mit %s: Replay %d, Fehler %v", c.kat, c.name, g.r.zahl(), err)
		}
	}
}

// TestReplayMiss hält Z08: liefert R ODR-E5001, kehrt es als derselbe Wert zurück, für jede Methode und unter einer
// Policy ohne und mit Host-Freigabe; M und R genau einmal.
func TestReplayMiss(t *testing.T) {
	methoden := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE", "OPTIONS", "TRACE", "get"}
	for _, m := range methoden {
		for _, mitHost := range []bool{false, true} {
			p, name := mitKategorien(t, katUnsupported, katCoverage), "ohne Host"
			if mitHost {
				p, name = offenste(t), "mit Host"
			}
			g := neuesRig(t, p, standardMax)
			g.mockFehler(code(codeNamed(t, "ODR-E6003")))
			miss := code(codeNamed(t, "ODR-E5001"))
			g.replayFehler(miss)
			_, err := g.v.Hybrid(context.Background(), anfrage(m))
			if m == "OPTIONS" || m == "TRACE" || m == "get" {
				g.zaehlen(t, m+" "+name, 1, 0)
				continue
			}
			if err != error(miss) {
				t.Errorf("%s %s: Fehler %v, erwartet derselbe Wert", m, name, err)
			}
			g.zaehlen(t, m+" "+name, 1, 1)
		}
	}
}

// TestReplayFehlerSchleife hält Z09: jeder Code von model.AlleCodes() als Fehler von R, unter der offensten Policy
// und für GET, kehrt als derselbe Wert zurück; M und R genau einmal.
func TestReplayFehlerSchleife(t *testing.T) {
	alle := model.AlleCodes()
	n := 0
	for _, c := range alle {
		n++
		g := neuesRig(t, offenste(t), standardMax)
		g.mockFehler(code(codeNamed(t, "ODR-E6003")))
		fehler := code(c)
		g.replayFehler(fehler)
		_, err := g.v.Hybrid(context.Background(), anfrage("GET"))
		if err != error(fehler) {
			t.Errorf("%s: Fehler %v, erwartet derselbe Wert", c, err)
		}
		g.zaehlen(t, c.String(), 1, 1)
	}
	if n != len(model.AlleCodes()) || n < 40 {
		t.Errorf("Zählprüfung: %d Codes", n)
	}
}

// TestAbbruchZwischenStufen hält Z10: ist ctx nach M beendet, läuft R nicht und das Ergebnis ist ctx.Err() ohne Code;
// ein beendetes ctx ohne Auslöser ändert den Fehler des Mock nicht.
func TestAbbruchZwischenStufen(t *testing.T) {
	e6003 := code(codeNamed(t, "ODR-E6003"))
	t.Run("Abbruch in M", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		g := neuesRig(t, offenste(t), standardMax)
		g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			cancel()
			return inbound.Antwort{}, e6003
		}
		g.replayAntwort(antwortVon(200, nil))
		_, err := g.v.Hybrid(ctx, anfrage("GET"))
		var fe *inbound.Fehler
		if err != context.Canceled || errors.As(err, &fe) {
			t.Errorf("Fehler %v, erwartet context.Canceled ohne Code", err)
		}
		g.zaehlen(t, "Abbruch in M", 1, 0)
	})
	t.Run("Frist in M", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		g := neuesRig(t, offenste(t), standardMax)
		g.m.fn = func(_ int, c context.Context, _ inbound.Anfrage) (inbound.Antwort, error) {
			<-c.Done()
			return inbound.Antwort{}, e6003
		}
		g.replayAntwort(antwortVon(200, nil))
		_, err := g.v.Hybrid(ctx, anfrage("GET"))
		if err != context.DeadlineExceeded {
			t.Errorf("Fehler %v, erwartet context.DeadlineExceeded", err)
		}
		g.zaehlen(t, "Frist in M", 1, 0)
	})
	t.Run("schon beendet vor M", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		g := neuesRig(t, offenste(t), standardMax)
		g.mockFehler(e6003)
		g.replayAntwort(antwortVon(200, nil))
		_, err := g.v.Hybrid(ctx, anfrage("GET"))
		if err != context.Canceled {
			t.Errorf("Fehler %v, erwartet context.Canceled", err)
		}
		g.zaehlen(t, "schon beendet", 1, 0)
	})
	t.Run("ohne Auslöser", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		g := neuesRig(t, offenste(t), standardMax)
		fehler := code(codeNamed(t, "ODR-E6005"))
		g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			cancel()
			return inbound.Antwort{}, fehler
		}
		_, err := g.v.Hybrid(ctx, anfrage("GET"))
		if err != error(fehler) {
			t.Errorf("Fehler %v, erwartet der Fehler des Mock", err)
		}
	})
}

type schluessel struct{}

// TestKontextAnStellen hält Z11: H2 und H4 erhalten das ctx des Aufrufs unverändert (Identität, Wert, Done), und ein
// Kontextfehler dort, auch verpackt, kehrt unverändert ohne Code zurück.
func TestKontextAnStellen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), schluessel{}, "w"))
	defer cancel()
	g := neuesRig(t, offenste(t), standardMax)
	g.mockFehler(code(codeNamed(t, "ODR-E6003")))
	g.replayAntwort(antwortVon(200, nil))
	if _, err := g.v.Hybrid(ctx, anfrage("GET")); err != nil {
		t.Fatal(err)
	}
	for stelle, s := range map[string]*stufe{"H2 Mock": g.m, "H4 Replay": g.r} {
		got := s.erster(t).ctx
		if got != ctx || got.Value(schluessel{}) != "w" || got.Done() != ctx.Done() {
			t.Errorf("%s: ctx nicht unverändert", stelle)
		}
	}
	kontextFehler := []error{context.Canceled, context.DeadlineExceeded, fmt.Errorf("r: %w", context.Canceled), errors.Join(errors.New("a"), context.DeadlineExceeded)}
	for _, fehler := range kontextFehler {
		for stelle := range 2 {
			g := neuesRig(t, offenste(t), standardMax)
			if stelle == 0 {
				g.mockFehler(fehler)
			} else {
				g.mockFehler(code(codeNamed(t, "ODR-E6003")))
				g.replayFehler(fehler)
			}
			_, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			if !sameErr(err, fehler) {
				t.Errorf("Stelle H%d, Fehler %v: nicht unverändert zurückgekehrt: %v", 2+2*stelle, fehler, err)
			}
		}
	}
}

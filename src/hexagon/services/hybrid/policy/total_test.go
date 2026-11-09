package policy_test

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// listenWerte sind die Werte für P01: zwei Hosts, drei Freigaben (jede einer anderen Methode und
// eines anderen Musters), Kategorien nach kat.
func listenWerte(kat string) policy.Werte {
	return policy.Werte{
		Upstream:         true,
		Hosts:            []string{"a.h", "b.h"},
		Schreibfreigaben: []string{"POST /a", "PUT /b", "DELETE /c"},
		ReplayKategorien: kategorienListe(kat),
	}
}

type eingriff struct {
	name string
	vor  func(w *policy.Werte)
	nach func(w *policy.Werte)
}

// eingriffe sind die drei Eingriffe von P01: Reihenfolge umgekehrt, jeder Eintrag verdoppelt (beides
// vor New), und die Slices nach New überschrieben und gekürzt.
func eingriffe(liste string) []eingriff {
	jede := func(w *policy.Werte, f func(*[]string)) {
		switch liste {
		case policy.SchluesselHosts:
			f(&w.Hosts)
		case policy.SchluesselSchreibfreigaben:
			f(&w.Schreibfreigaben)
		default:
			f(&w.ReplayKategorien)
		}
	}
	return []eingriff{
		{name: "umgekehrt", vor: func(w *policy.Werte) { jede(w, func(l *[]string) { slices.Reverse(*l) }) }},
		{name: "verdoppelt", vor: func(w *policy.Werte) {
			jede(w, func(l *[]string) {
				var doppelt []string
				for _, e := range *l {
					doppelt = append(doppelt, e, e)
				}
				*l = doppelt
			})
		}},
		{name: "ueberschrieben", nach: func(w *policy.Werte) {
			for _, l := range [][]string{w.Hosts, w.Schreibfreigaben, w.ReplayKategorien} {
				for i := range l {
					l[i] = "!"
				}
			}
			w.Hosts, w.Schreibfreigaben, w.ReplayKategorien = w.Hosts[:0], w.Schreibfreigaben[:0], w.ReplayKategorien[:0]
		}},
	}
}

// entscheidungen fährt 3 Codes x 5 Methoden x 2 Ziele x 3 Pfade (90 Anfragen) und liefert die
// Entscheidungen in fester Reihenfolge.
func entscheidungen(p *policy.Policy) []policy.Entscheidung {
	var aus []policy.Entscheidung
	for _, c := range []model.Code{e6001, e6003, e5001} {
		for _, m := range []string{"GET", "POST", "PUT", "DELETE", "HEAD"} {
			for _, z := range []string{"https://a.h", "https://b.h"} {
				for _, pfad := range []string{"/a", "/b", "/c"} {
					aus = append(aus, ent(p, m, pfad, z, c))
				}
			}
		}
	}
	return aus
}

// TestListenSindMengen (P01): Liste {Hosts, Schreibfreigaben, ReplayKategorien} x Eingriff
// {umgekehrt, verdoppelt, nach New überschrieben und gekürzt} = 9 Fälle: dieselben 90 Entscheidungen
// wie ohne Eingriff, und New meldet keinen Fehler.
func TestListenSindMengen(t *testing.T) {
	faelle := 0
	for _, liste := range []string{policy.SchluesselHosts, policy.SchluesselSchreibfreigaben, policy.SchluesselReplayKategorien} {
		kat := kategorien0
		if liste == policy.SchluesselReplayKategorien {
			kat = kategorien2
		}
		basis := entscheidungen(neu(t, listenWerte(kat)))
		var anzahl [3]int
		for _, e := range basis {
			anzahl[e]++
		}
		if len(basis) != 90 || anzahl[policy.Bleiben] == 0 || anzahl[policy.Upstream]+anzahl[policy.Replay] == 0 {
			t.Fatalf("%s: Grundlage %d Entscheidungen, Verteilung %v unterscheidet nichts", liste, len(basis), anzahl)
		}
		for _, in := range eingriffe(liste) {
			faelle++
			w := listenWerte(kat)
			if in.vor != nil {
				in.vor(&w)
			}
			p := neu(t, w)
			if in.nach != nil {
				in.nach(&w)
			}
			if got := entscheidungen(p); !slices.Equal(got, basis) {
				t.Errorf("%s, %s: Entscheidungen weichen ab", liste, in.name)
			}
		}
	}
	if faelle != 9 {
		t.Fatalf("%d Fälle, erwartet 9", faelle)
	}
}

type eingabe struct{ methode, pfad, ziel string }

// grenzEingaben sind die Grenzformen von P22: leer, NUL, ungültiges UTF-8, ein Pfad von 1 MiB, ein
// `%` am Ende, 10000 Segmente.
func grenzEingaben() []eingabe {
	nul := string([]byte{0})
	kaputt := string([]byte{0xff, 0xfe, 0x80})
	mib := 1 << 20
	gross := "/" + strings.Repeat("a", mib-1)
	var aus []eingabe
	for _, m := range []string{"", "GET", "POST", nul, kaputt, "PUT"} {
		for _, pfad := range []string{"", "/", nul, kaputt, "/%", "/svc/Orders/%", "/svc/Orders/%2", "//", "/svc//Orders"} {
			for _, z := range []string{"", zielPassend, nul, kaputt, "https://h:", "https://[::1]:8443"} {
				aus = append(aus, eingabe{m, pfad, z})
			}
		}
	}
	for _, m := range []string{"GET", "POST", "PUT"} {
		aus = append(aus,
			eingabe{m, gross, zielPassend},
			eingabe{m, strings.Repeat("/a", 10000), zielPassend},
			eingabe{m, "/svc/Orders" + strings.Repeat("/x", 10000), zielPassend},
			eingabe{m, "/svc/Orders/" + strings.Repeat("x", mib/2), "https://" + strings.Repeat("h", mib/2)},
			eingabe{strings.Repeat("P", mib), "/svc/Orders/x", zielPassend},
		)
	}
	aus = append(aus, eingabe{"POST", strings.Repeat("/", mib), zielPassend})
	return aus
}

// zufallsEingaben füllt grenzEingaben mit Zufall (fester Seed) bis n Eingaben auf.
func zufallsEingaben(n int) []eingabe {
	r := rand.New(rand.NewPCG(1, 2))
	aus := grenzEingaben()
	methoden := append(methoden13(), "POST", "PUT", "DELETE")
	ziele := []string{zielPassend, "http://h", "https://[::1]:8443", "", "h", "http://h:99999", "https://h/%", "https://H", "https://x"}
	segmente := []string{"svc", "Orders", "x", "a", "*", "..", ".", "%2e", "%2F", "", "Orders(1)", "xay", "\x00", "\xff", "%", "\\"}
	for len(aus) < n {
		m := methoden[r.IntN(len(methoden))]
		if r.IntN(5) == 0 {
			m = string([]byte{byte(r.IntN(256)), byte(r.IntN(256))})
		}
		z := ziele[r.IntN(len(ziele))]
		if r.IntN(8) == 0 {
			z = string([]byte{byte(r.IntN(256)), byte(r.IntN(256)), byte(r.IntN(256))})
		}
		var pfad strings.Builder
		if r.IntN(3) == 0 {
			pfad.WriteString("/svc/Orders")
		}
		for i := r.IntN(6); i > 0; i-- {
			pfad.WriteString("/" + segmente[r.IntN(len(segmente))])
		}
		aus = append(aus, eingabe{m, pfad.String(), z})
	}
	return aus
}

// TestTotalitaet (P22): 3 Policies (nil, Policy{}, voll offen) x 6000 Eingaben x C (44 Codes):
// keine Panik, nil und Policy{} stets Bleiben; jede Eingabe zweimal, dazwischen 1000 andere Aufrufe,
// mit gleicher Entscheidung. Die 10000 Segmente und der Pfad von 1 MiB laufen in der Zeitgrenze von
// go test.
func TestTotalitaet(t *testing.T) {
	voll := neu(t, policy.Werte{
		Upstream: true,
		Hosts:    []string{"h", "[::1]:8443"},
		Schreibfreigaben: []string{
			"POST /svc/Orders/**", "PUT /svc/*/x*y/*", "DELETE /a/*/**", "PATCH /%2f/x", "MERGE /*", "POST /a",
			"POST /*", "PUT /svc/Orders/**",
		},
		ReplayKategorien: []string{policy.KategorieCoverage},
	})
	policies := []*policy.Policy{nil, {}, voll}
	eingaben := zufallsEingaben(6000)
	codes := codesC(t)
	if len(eingaben) != 6000 {
		t.Fatalf("%d Eingaben", len(eingaben))
	}
	lauf := func() []policy.Entscheidung {
		var aus []policy.Entscheidung
		for pi, p := range policies {
			for _, in := range eingaben {
				for _, c := range codes {
					e := p.Entscheide(policy.Anfrage{Methode: in.methode, Pfad: in.pfad, Ziel: in.ziel, Code: c})
					if pi < 2 && e != policy.Bleiben {
						t.Fatalf("Policy %d: %s", pi, name(e))
					}
					aus = append(aus, e)
				}
			}
		}
		return aus
	}
	erste := lauf()
	if len(erste) != 3*6000*44 {
		t.Fatalf("%d Entscheidungen", len(erste))
	}
	var anzahl [3]int
	for _, e := range erste {
		anzahl[e]++
	}
	if anzahl[policy.Replay] == 0 || anzahl[policy.Upstream] == 0 {
		t.Fatalf("Verteilung %v: die Eingaben erreichen nicht alle Entscheidungen", anzahl)
	}
	for i := 0; i < 1000; i++ {
		in := eingaben[(i*7)%len(eingaben)]
		voll.Entscheide(policy.Anfrage{Methode: in.methode, Pfad: in.pfad, Ziel: in.ziel, Code: codes[i%len(codes)]})
	}
	if zweite := lauf(); !slices.Equal(zweite, erste) {
		t.Fatalf("die zweite Runde weicht von der ersten ab")
	}
}

package hybrid_test

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// weg ist ein Fall des Fallbacks mit dem Verhalten der drei Stufen, der erwarteten Folge aller Aufrufe (Literal) und dem
// Ausgang: ident (der Aufruf liefert genau diesen Fehler), sonst code mit status (ein *inbound.Fehler), sonst eine Antwort
// mit dem Status. ergebnis ist das Ergebnis des ersten Spans, gezaehlt die Folge der Zählungen (leer: nicht geprüft).
type weg struct {
	name     string
	methode  string
	pol      func(*testing.T) *policy.Policy
	vorher   error
	m, r, u  verhalten
	folge    string
	ergebnis string
	gezaehlt string
	ident    error
	code     model.Code
	status   int
}

// ausfuehren baut den Dienst, ruft ihn und hält Folge, Buch und Zählungen gegen das Literal.
func (w weg) ausfuehren(t *testing.T) (umgebung, inbound.Antwort, error) {
	t.Helper()
	u := neueUmgebung(t, w.pol(t), zielH, w.m, w.r, w.u)
	u.c.beende(w.vorher)
	a, err := u.rufe(cmp.Or(w.methode, "GET"))
	if got := u.j.text(); w.folge != "" && got != w.folge {
		t.Errorf("%s: Folge\n  %s\nerwartet\n  %s", w.name, got, w.folge)
	}
	if got := strings.Join(u.j.gezaehlt(), " "); w.gezaehlt != "" && got != w.gezaehlt {
		t.Errorf("%s: Zählungen %q, erwartet %q", w.name, got, w.gezaehlt)
	}
	pruefeBuch(t, w.name, u.j)
	if sp := u.j.spanListe(); w.ergebnis != "" && len(sp) > 0 {
		if _, _, e := sp[0].buch(); e != w.ergebnis {
			t.Errorf("%s: Ergebnis des ersten Spans %q, erwartet %q", w.name, e, w.ergebnis)
		}
	}
	return u, a, err
}

// lauf führt den Fall aus und prüft den Ausgang.
func (w weg) lauf(t *testing.T) umgebung {
	t.Helper()
	u, a, err := w.ausfuehren(t)
	var f *inbound.Fehler
	switch {
	case w.ident != nil:
		if err != w.ident {
			t.Errorf("%s: Fehler %v, erwartet derselbe Wert %v", w.name, err, w.ident)
		}
	case w.code != 0:
		if !errors.As(err, &f) || f.Code != w.code || f.Status != w.status {
			t.Errorf("%s: Fehler %v, erwartet Code %s mit Status %d", w.name, err, w.code, w.status)
		}
	case err != nil || a.Status != w.status:
		t.Errorf("%s: Status %d und Fehler %v, erwartet Antwort mit Status %d", w.name, a.Status, err, w.status)
	}
	return u
}

// Die Folgen bis zur Stufe: Mock mit ODR-E6003 und Span replay, Mock mit ODR-E6001 und Span upstream.
const (
	mR = "Mock Count(ODR-E6003) Start(replay)"
	mU = "Mock Count(ODR-E6001) Start(upstream)"
)

// fehlerCodes liefert die Fehler der Stufen der Wege als Werte, deren Identität der Test vergleicht.
func fehlerCodes(t *testing.T, namen ...string) []error {
	t.Helper()
	var out []error
	for _, n := range namen {
		out = append(out, code(codeNamed(t, n)))
	}
	return out
}

// fuenfzehnWege sind die Rückwege nach dem Beginn (SPEC-321 Nr. 7 a): acht der Stufe replay und sieben der Stufe upstream.
func fuenfzehnWege(t *testing.T) []weg {
	f := fehlerCodes(t, "ODR-E6001", "ODR-E6003", "ODR-E5003", "ODR-E5001")
	e6001, e6003, e5003, miss := f[0], f[1], f[2], f[3]
	ohneCode := errors.New("ohne Code")
	canc := context.Canceled
	return []weg{
		{name: "R1 ctx vor R", pol: polR, m: verhalten{err: e6003, beende: canc}, ergebnis: "abgebrochen", ident: canc,
			folge: mR + " Set(replay=abgebrochen) End(replay)"},
		{name: "R2 Antwort", pol: polR, m: verhalten{err: e6003}, r: verhalten{status: 200}, ergebnis: "antwort", status: 200,
			folge: mR + " Replay Set(replay=antwort) End(replay)"},
		{name: "R3 Antwort 404", pol: polR, m: verhalten{err: e6003}, r: verhalten{status: 404}, ergebnis: "antwort", status: 404,
			folge: mR + " Replay Set(replay=antwort) End(replay)"},
		{name: "R4 Kontextfehler von R", pol: polR, m: verhalten{err: e6003}, r: verhalten{err: canc, body: true, beende: canc},
			ergebnis: "abgebrochen", ident: canc, folge: mR + " Replay Close Set(replay=abgebrochen) End(replay)"},
		{name: "R5 Fehler mit Code", pol: polR, m: verhalten{err: e6003}, r: verhalten{err: e5003, body: true},
			ergebnis: "fehler", ident: e5003, folge: mR + " Replay Close Set(replay=fehler) End(replay)"},
		{name: "R6 Fehler ohne Code", pol: polR, m: verhalten{err: e6003}, r: verhalten{err: ohneCode, body: true},
			ergebnis: "fehler", ident: ohneCode, folge: mR + " Replay Close Set(replay=fehler) End(replay)"},
		{name: "R7 Miss ohne Upstream", pol: polR, m: verhalten{err: e6003}, r: verhalten{err: miss, body: true},
			ergebnis: "fehler", ident: miss, folge: mR + " Replay Close Set(replay=fehler) End(replay)"},
		{name: "R8 Miss, dann Upstream", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: miss, body: true},
			u: verhalten{status: 200}, ergebnis: "fehler", status: 200,
			folge: mR + " Replay Close Set(replay=fehler) End(replay) Count(ODR-E5001) Start(upstream) Senden Set(upstream=antwort) End(upstream)"},
		{name: "U1 ctx vor U", pol: polU, m: verhalten{err: e6001, beende: canc}, ergebnis: "abgebrochen", ident: canc,
			folge: mU + " Set(upstream=abgebrochen) End(upstream)"},
		{name: "U2 Antwort", pol: polU, m: verhalten{err: e6001}, u: verhalten{status: 200}, ergebnis: "antwort", status: 200,
			folge: mU + " Senden Set(upstream=antwort) End(upstream)"},
		{name: "U3 Antwort 502", pol: polU, m: verhalten{err: e6001}, u: verhalten{status: 502}, ergebnis: "antwort", status: 502,
			folge: mU + " Senden Set(upstream=antwort) End(upstream)"},
		{name: "U4 Kontextfehler von Senden", pol: polU, m: verhalten{err: e6001}, u: verhalten{err: canc, body: true, beende: canc},
			ergebnis: "abgebrochen", ident: canc, folge: mU + " Senden Close Set(upstream=abgebrochen) End(upstream)"},
		{name: "U5 ErrUpstreamZeit", pol: polU, m: verhalten{err: e6001}, u: verhalten{err: outbound.ErrUpstreamZeit, body: true},
			ergebnis: "fehler", code: model.CodeE4001, status: 504, folge: mU + " Senden Close Set(upstream=fehler) End(upstream)"},
		{name: "U6 ErrUpstreamNichtErreichbar", pol: polU, m: verhalten{err: e6001},
			u: verhalten{err: outbound.ErrUpstreamNichtErreichbar, body: true}, ergebnis: "fehler", code: model.CodeE4001, status: 502,
			folge: mU + " Senden Close Set(upstream=fehler) End(upstream)"},
		{name: "U7 anderer Fehler", pol: polU, m: verhalten{err: e6001}, u: verhalten{err: ohneCode, body: true},
			ergebnis: "fehler", code: model.CodeE9001, status: 500, folge: mU + " Senden Close Set(upstream=fehler) End(upstream)"},
	}
}

// ketten sind die zwei weiteren Ausgänge der Kette Replay-Miss, dann Upstream (SPEC-321 Nr. 2): der Upstream scheitert, und
// der ctx endet in R, so dass die Stufe upstream nicht läuft.
func ketten(t *testing.T) []weg {
	f := fehlerCodes(t, "ODR-E6003", "ODR-E5001")
	e6003, miss := f[0], f[1]
	return []weg{
		{name: "K1 Kette, Upstream scheitert", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: miss, body: true},
			u: verhalten{err: outbound.ErrUpstreamNichtErreichbar, body: true}, ergebnis: "fehler", code: model.CodeE4001, status: 502,
			folge: mR + " Replay Close Set(replay=fehler) End(replay) Count(ODR-E5001) Start(upstream) Senden Close Set(upstream=fehler) End(upstream)"},
		{name: "K2 Kette, ctx endet in R", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: miss, body: true, beende: context.Canceled},
			ergebnis: "fehler", ident: context.Canceled,
			folge: mR + " Replay Close Set(replay=fehler) End(replay) Count(ODR-E5001) Start(upstream) Set(upstream=abgebrochen) End(upstream)"},
	}
}

// TestWege hält D3 bis D6: die 15 Wege nach dem Beginn und die zwei weiteren der Kette, je mit der Folge aller Aufrufe als
// Literal (Entscheidung, Zählung, Beginn, Stufe, Schließen des Bodys, Set, End), je einem Set und einem End, dem Ergebnis
// und der Art. Zählprüfung 15 (8 und 7) gleich SPEC-321 Nr. 7 a; Ergebnisse 4, 4 und 7; Art: 2 Einzelfälle und 3 Ketten.
func TestWege(t *testing.T) {
	haupt, zaehlung := fuenfzehnWege(t), map[string]int{}
	replay, upstream := 0, 0
	for _, w := range haupt {
		w.lauf(t)
		zaehlung[w.ergebnis]++
		if strings.HasPrefix(w.name, "R") {
			replay++
		} else {
			upstream++
		}
	}
	if len(haupt) != 15 || replay != 8 || upstream != 7 {
		t.Errorf("Zählprüfung: %d Wege (%d replay, %d upstream), erwartet 15 (8 und 7)", len(haupt), replay, upstream)
	}
	if zaehlung["abgebrochen"] != 4 || zaehlung["antwort"] != 4 || zaehlung["fehler"] != 7 {
		t.Errorf("Ergebnisse %v, erwartet 4 abgebrochen, 4 antwort, 7 fehler", zaehlung)
	}
	folge := 0
	for _, w := range haupt {
		if w.name == "R1 ctx vor R" || w.name == "R2 Antwort" || w.name == "U2 Antwort" {
			folge++
		}
	}
	if folge != 3 {
		t.Errorf("Zählprüfung der Folgen von D3 (ctx beendet, Replay, Upstream): %d, erwartet 3", folge)
	}
	arten := 0
	for _, w := range append(ketten(t), haupt[1], haupt[7], haupt[9]) {
		w.lauf(t)
		arten++
	}
	if arten != 5 {
		t.Errorf("Zählprüfung der Arten: %d, erwartet 5", arten)
	}
}

// TestAbbruch hält D11: fünf Stellen mal zwei Kontextfehler und ein Kontextfehler neben ODR-E5001 von R gleich 11; je Fall
// ein Set mit abgebrochen, ein End, die Stufe nicht gerufen, wo der ctx vorher endete, und derselbe Fehler als Ergebnis.
// Der Kontextfehler neben ODR-E5001 ist kein Miss: kein Upstream.
func TestAbbruch(t *testing.T) {
	f := fehlerCodes(t, "ODR-E6001", "ODR-E6003", "ODR-E5001")
	e6001, e6003, miss := f[0], f[1], f[2]
	n := 0
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		faelle := []weg{
			{name: "S1 ctx vor R", pol: polR, m: verhalten{err: e6003, beende: e}, ident: e,
				folge: mR + " Set(replay=abgebrochen) End(replay)"},
			{name: "S2 ctx vor U nach M", pol: polU, m: verhalten{err: e6001, beende: e}, ident: e,
				folge: mU + " Set(upstream=abgebrochen) End(upstream)"},
			{name: "S3 ctx vor U nach dem Miss", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: miss, body: true, beende: e}, ident: e,
				folge: mR + " Replay Close Set(replay=fehler) End(replay) Count(ODR-E5001) Start(upstream) Set(upstream=abgebrochen) End(upstream)"},
			{name: "S4 Kontextfehler von R", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: e, body: true}, ident: e,
				folge: mR + " Replay Close Set(replay=abgebrochen) End(replay)"},
			{name: "S5 Kontextfehler von Senden", pol: polU, m: verhalten{err: e6001}, u: verhalten{err: e, body: true}, ident: e,
				folge: mU + " Senden Close Set(upstream=abgebrochen) End(upstream)"},
		}
		for _, w := range faelle {
			w.name += " " + e.Error()
			w.lauf(t)
			n++
		}
	}
	gemischt := errors.Join(miss, context.Canceled)
	w := weg{name: "S6 Kontextfehler neben ODR-E5001", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: gemischt, body: true}, ident: gemischt,
		folge: mR + " Replay Close Set(replay=abgebrochen) End(replay)"}
	w.lauf(t)
	if n+1 != 11 {
		t.Errorf("Zählprüfung: %d Fälle, erwartet 11", n+1)
	}
}

// verworfenWege sind die acht Fälle von M-Verworfen (SPEC-321 Nr. 5): ODR-E6001 und ODR-E6003 je mit Replay, Upstream und
// Kette, ODR-E6002 (mutierend, POST) mit Replay und Upstream; die Folge der Zählungen ist das Literal.
func verworfenWege(t *testing.T) []weg {
	f := fehlerCodes(t, "ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E5001")
	e6001, e6002, e6003, miss := f[0], f[1], f[2], f[3]
	polKetteU := func(t *testing.T) *policy.Policy { return uPolicy(t, []string{"h"}, nil, katUnsupported) }
	polPost := func(t *testing.T) *policy.Policy { return uPolicy(t, []string{"h"}, []string{"POST /**"}) }
	kette := " Replay Close Set(replay=fehler) End(replay) Count(ODR-E5001) Start(upstream) Senden Set(upstream=antwort) End(upstream)"
	return []weg{
		{name: "E6001 Replay", pol: polR, m: verhalten{err: e6001}, r: verhalten{status: 200}, gezaehlt: "ODR-E6001",
			folge: "Mock Count(ODR-E6001) Start(replay) Replay Set(replay=antwort) End(replay)", status: 200},
		{name: "E6001 Upstream", pol: polU, m: verhalten{err: e6001}, u: verhalten{status: 200}, gezaehlt: "ODR-E6001",
			folge: mU + " Senden Set(upstream=antwort) End(upstream)", status: 200},
		{name: "E6001 Kette", pol: polKetteU, m: verhalten{err: e6001}, r: verhalten{err: miss, body: true}, u: verhalten{status: 200},
			gezaehlt: "ODR-E6001 ODR-E5001", folge: "Mock Count(ODR-E6001) Start(replay)" + kette, status: 200},
		{name: "E6003 Replay", pol: polR, m: verhalten{err: e6003}, r: verhalten{status: 200}, gezaehlt: "ODR-E6003",
			folge: mR + " Replay Set(replay=antwort) End(replay)", status: 200},
		{name: "E6003 Upstream", pol: polU, m: verhalten{err: e6003}, u: verhalten{status: 200}, gezaehlt: "ODR-E6003",
			folge: "Mock Count(ODR-E6003) Start(upstream) Senden Set(upstream=antwort) End(upstream)", status: 200},
		{name: "E6003 Kette", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: miss, body: true}, u: verhalten{status: 200},
			gezaehlt: "ODR-E6003 ODR-E5001", folge: mR + kette, status: 200},
		{name: "E6002 Replay", methode: "POST", pol: polR, m: verhalten{err: e6002}, r: verhalten{status: 200}, gezaehlt: "ODR-E6002",
			folge: "Mock Count(ODR-E6002) Start(replay) Replay Set(replay=antwort) End(replay)", status: 200},
		{name: "E6002 Upstream", methode: "POST", pol: polPost, m: verhalten{err: e6002}, u: verhalten{status: 200}, gezaehlt: "ODR-E6002",
			folge: "Mock Count(ODR-E6002) Start(upstream) Senden Set(upstream=antwort) End(upstream)", status: 200},
	}
}

// TestVerworfen hält D7: die Menge M-Verworfen mit 8 Fällen und 10 Zählungen; jeder Versuch zählt genau einmal mit dem Code
// des Fehlers von M, in der Kette danach ODR-E5001; alle vier Codes von support.VerworfeneCodes() werden erreicht.
func TestVerworfen(t *testing.T) {
	erreicht, zaehlungen, faelle := map[string]bool{}, 0, verworfenWege(t)
	for _, w := range faelle {
		u := w.lauf(t)
		for _, c := range u.j.gezaehlt() {
			erreicht[c] = true
			zaehlungen++
		}
	}
	var soll []string
	for _, c := range support.VerworfeneCodes() {
		soll = append(soll, c.String())
	}
	var got []string
	for c := range erreicht {
		got = append(got, c)
	}
	slices.Sort(got)
	if !slices.Equal(got, soll) || len(soll) != 4 {
		t.Errorf("erreichte Codes %v, erwartet die vier von VerworfeneCodes() %v", got, soll)
	}
	if len(faelle) != 8 || zaehlungen != 10 {
		t.Errorf("Zählprüfung: %d Fälle und %d Zählungen, erwartet 8 und 10", len(faelle), zaehlungen)
	}
}

// TestNichtGezaehlt hält D8: was unverändert zurückkehrt und der Fehler der letzten Stufe zählt nicht; die Menge M-NichtGezählt
// hat 20 Fälle, je mit genau einer Zählung, dem Code von M: ODR-E5001 mit Bleiben und der mutierende Replay-Miss, jeder Fehler von
// R außer ODR-E5001 (12 Codes von SPEC-296 Nr. 5 und ein Fehler ohne Code), drei Fehler von U und eine Antwort je Stufe.
func TestNichtGezaehlt(t *testing.T) {
	f := fehlerCodes(t, "ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E5001")
	e6001, e6002, e6003, miss := f[0], f[1], f[2], f[3]
	faelle := []weg{
		{name: "E5001 mit Bleiben", pol: polR, m: verhalten{err: e6003}, r: verhalten{err: miss, body: true}, gezaehlt: "ODR-E6003"},
		{name: "mutierender Replay-Miss", methode: "POST", pol: offenste, m: verhalten{err: e6002}, r: verhalten{err: miss, body: true},
			gezaehlt: "ODR-E6002"},
		{name: "Antwort von R", pol: polK, m: verhalten{err: e6003}, r: verhalten{status: 200}, gezaehlt: "ODR-E6003"},
		{name: "Antwort von U", pol: polU, m: verhalten{err: e6001}, u: verhalten{status: 200}, gezaehlt: "ODR-E6001"},
		{name: "Fehler von R ohne Code", pol: polK, m: verhalten{err: e6003}, r: verhalten{err: errors.New("ohne Code"), body: true},
			gezaehlt: "ODR-E6003"},
	}
	for _, fe := range []error{outbound.ErrUpstreamZeit, outbound.ErrUpstreamNichtErreichbar, errors.New("sonst")} {
		faelle = append(faelle, weg{name: "Fehler von U " + fe.Error(), pol: polU, m: verhalten{err: e6001}, u: verhalten{err: fe, body: true},
			gezaehlt: "ODR-E6001"})
	}
	codesR := []string{"ODR-E3001", "ODR-E5002", "ODR-E5003", "ODR-E5004", "ODR-E5005", "ODR-E5006", "ODR-E5007", "ODR-E5008", "ODR-E7001",
		"ODR-E7003", "ODR-E7009", "ODR-E9001"}
	for _, n := range codesR {
		faelle = append(faelle, weg{name: "Fehler von R " + n, pol: polK, m: verhalten{err: e6003},
			r: verhalten{err: code(codeNamed(t, n)), body: true}, gezaehlt: "ODR-E6003"})
	}
	for _, w := range faelle {
		_, _, _ = w.ausfuehren(t)
	}
	if len(faelle) != 20 || len(codesR) != 12 {
		t.Errorf("Zählprüfung: %d Fälle und %d Codes von R, erwartet 20 und 12", len(faelle), len(codesR))
	}
}

// TestMethoden hält D15: fünf mutierende Methoden mit ODR-E6002 tragen je den Span der Stufe (Replay mit der Kategorie
// unsupported, Upstream mit der Routenfreigabe) und zählen ODR-E6002; der Replay-Miss einer mutierenden Methode erreicht den
// Upstream nie (je Methode), und bei beiden offenen Wegen entscheidet der Replay. 5 Methoden mal 3 Fälle (Replay, Upstream, Miss)
// und 1 weiterer gleich 16.
func TestMethoden(t *testing.T) {
	f := fehlerCodes(t, "ODR-E6002", "ODR-E6003", "ODR-E5001")
	e6002, e6003, miss := f[0], f[1], f[2]
	methoden, n := []string{"POST", "PUT", "PATCH", "DELETE", "MERGE"}, 0
	for _, m := range methoden {
		route := func(t *testing.T) *policy.Policy { return uPolicy(t, []string{"h"}, []string{m + " /**"}) }
		for _, w := range []weg{
			{name: m + " Replay", methode: m, pol: polR, m: verhalten{err: e6002}, r: verhalten{status: 200}, status: 200, gezaehlt: "ODR-E6002",
				folge: "Mock Count(ODR-E6002) Start(replay) Replay Set(replay=antwort) End(replay)"},
			{name: m + " Upstream", methode: m, pol: route, m: verhalten{err: e6002}, u: verhalten{status: 200}, status: 200, gezaehlt: "ODR-E6002",
				folge: "Mock Count(ODR-E6002) Start(upstream) Senden Set(upstream=antwort) End(upstream)"},
			{name: m + " Miss", methode: m, pol: offenste, m: verhalten{err: e6002}, r: verhalten{err: miss, body: true}, ident: miss,
				gezaehlt: "ODR-E6002", folge: "Mock Count(ODR-E6002) Start(replay) Replay Close Set(replay=fehler) End(replay)"},
		} {
			w.lauf(t)
			n++
		}
	}
	w := weg{name: "GET, beide Wege offen", pol: offenste, m: verhalten{err: e6003}, r: verhalten{status: 200}, status: 200, gezaehlt: "ODR-E6003",
		folge: mR + " Replay Set(replay=antwort) End(replay)"}
	w.lauf(t)
	n++
	if len(methoden) != 5 || n != 16 {
		t.Errorf("Zählprüfung: %d Methoden und %d Fälle, erwartet 5 und 16", len(methoden), n)
	}
}

// TestWiederholungTelemetrie hält D14: zwei gleiche Anfragen an denselben Dienst tragen je ihre Spans und Zählungen (Replay, Upstream, Kette):
// 6 Aufrufe, 8 Spans mit verschiedener ID und 8 Zählungen, die Folge der zweiten gleich der der ersten.
func TestWiederholungTelemetrie(t *testing.T) {
	w := fuenfzehnWege(t)
	spans, zaehlungen, aufrufe := 0, 0, 0
	for _, f := range []weg{w[1], w[9], w[7]} {
		u := neueUmgebung(t, f.pol(t), zielH, f.m, f.r, f.u)
		for range 2 {
			_, _ = u.rufe("GET")
			aufrufe++
		}
		if got, soll := u.j.text(), f.folge+" "+f.folge; got != soll {
			t.Errorf("%s: Folge\n  %s\nerwartet\n  %s", f.name, got, soll)
		}
		pruefeBuch(t, f.name, u.j)
		ids := map[int]bool{}
		for _, s := range u.j.spanListe() {
			ids[s.id] = true
			spans++
		}
		zaehlungen += len(u.j.gezaehlt())
		if len(ids) != len(u.j.spanListe()) {
			t.Errorf("%s: die Spans tragen nicht je eine eigene ID", f.name)
		}
	}
	if aufrufe != 6 || spans != 8 || zaehlungen != 8 {
		t.Errorf("Zählprüfung: %d Aufrufe, %d Spans, %d Zählungen; erwartet 6, 8 und 8", aufrufe, spans, zaehlungen)
	}
}

// TestKontext hält D10: StartSpan erhält das ctx des Aufrufs unverändert, und keine Stufe erhält das ctx von StartSpan: in der
// Kette sehen M (zur Kontrolle), R und U die Identität des ctx des Aufrufs und keinen Wert aus StartSpan; beide Spans
// entstehen mit der Identität des ctx des Aufrufs. 3 Stellen, 2 hinter dem Beginn.
func TestKontext(t *testing.T) {
	f := fehlerCodes(t, "ODR-E6003", "ODR-E5001")
	u := neueUmgebung(t, polK(t), zielH, verhalten{err: f[0]}, verhalten{err: f[1], body: true}, verhalten{status: 200})
	if _, err := u.rufe("GET"); err != nil {
		t.Fatalf("Kette: %v", err)
	}
	stellen := map[string]ctxMethoden{
		"Mock": u.g.m.erster(t).ctx, "Replay": u.g.r.erster(t).ctx, "Senden": u.g.u.erster(t).ctx,
	}
	for name, c := range stellen {
		if c != ctxMethoden(u.c) || c.Value(jSchluessel{}) != nil {
			t.Errorf("%s: erhielt nicht das ctx des Aufrufs (Wert aus StartSpan: %v)", name, c.Value(jSchluessel{}))
		}
	}
	spans := u.j.spanListe()
	for _, s := range spans {
		if s.ctx != ctxMethoden(u.c) {
			t.Errorf("Span %d: StartSpan erhielt nicht das ctx des Aufrufs", s.id)
		}
	}
	if len(stellen) != 3 || len(spans) != 2 {
		t.Errorf("Zählprüfung: %d Stellen und %d Spans, erwartet 3 und 2", len(stellen), len(spans))
	}
}

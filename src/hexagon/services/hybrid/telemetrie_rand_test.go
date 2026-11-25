package hybrid_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// keinSpan hält, dass der Dienst keine Telemetrie rief und keine Stufe des Fallbacks: die Folge enthält höchstens den Mock.
func keinSpan(t *testing.T, was string, j *journal) {
	t.Helper()
	for _, e := range strings.Fields(j.text()) {
		if e != "Mock" {
			t.Errorf("%s: Aufruf %q; erwartet allein Mock, kein Span, keine Zählung, keine Stufe", was, e)
		}
	}
	pruefeBuch(t, was, j)
}

// TestOhneSpan hält D2 (SPEC-321 Nr. 2, 7 b): wo kein Fallback beginnt, ruft der Dienst weder StartSpan noch CountDiscarded
// noch sonst die Telemetrie. Menge M-OhneSpan mit 62 Fällen: Erfolg von M 2; Fehler ohne Code 6; jeder Code der Tabelle außer den vier
// Auslösern 37; nie freigegebene Methode 4; Policy ohne Freigabe 3; Pfad ohne Dienst bei einziger Upstream 1; Host nicht
// freigegeben und disabled 2; Body über MaxBytes 2 und Lesefehler beim Puffern 1; Argumentfehler 4.
func TestOhneSpan(t *testing.T) {
	f := fehlerCodes(t, "ODR-E6001", "ODR-E6002", "ODR-E6003")
	e6001, e6002, e6003 := f[0], f[1], f[2]
	n, ok := 0, verhalten{status: 200}
	fall := func(was string, pol func(*testing.T) *policy.Policy, m verhalten, aend func(*inbound.Anfrage)) {
		t.Helper()
		u := neueUmgebung(t, pol(t), zielH, m, ok, ok)
		a := anfrage("GET")
		if aend != nil {
			aend(&a)
		}
		_, _ = u.g.v.Hybrid(u.c, a)
		keinSpan(t, was, u.j)
		n++
	}
	fall("Erfolg 200", offenste, verhalten{status: 200}, nil)
	fall("Erfolg 404", offenste, verhalten{status: 404}, nil)
	ohneCode := []error{model.ErrUngueltig, errors.New("beliebig"), context.Canceled, context.DeadlineExceeded, &inbound.Fehler{},
		errors.Join(e6001, context.Canceled)}
	for i, e := range ohneCode {
		fall(fmt.Sprintf("Fehler ohne Code %d", i), offenste, verhalten{err: e}, nil)
	}
	ausloeser, andere := support.VerworfeneCodes(), 0
	for _, c := range model.AlleCodes() {
		if !slices.Contains(ausloeser, c) {
			fall("Code "+c.String(), offenste, verhalten{err: code(c)}, nil)
			andere++
		}
	}
	if andere != 37 || len(model.AlleCodes())-len(ausloeser) != 37 {
		t.Errorf("Zählprüfung: %d Codes außer den Auslösern, erwartet 37 (gleich len(AlleCodes()) minus 4)", andere)
	}
	for _, m := range []string{"OPTIONS", "TRACE", "CONNECT", "get"} {
		fall("Methode "+m, offenste, verhalten{err: e6003}, func(a *inbound.Anfrage) { a.Methode = m })
	}
	leer := func(t *testing.T) *policy.Policy { return neuePolicy(t, policy.Werte{}) }
	for i, e := range []error{e6001, e6002, e6003} {
		fall(fmt.Sprintf("Policy ohne Freigabe %d", i), leer, verhalten{err: e}, nil)
	}
	fall("Pfad ohne Dienst", polU, verhalten{err: e6001}, func(a *inbound.Anfrage) { a.Pfad = "/andere/Products" })
	fall("Host nicht freigegeben", func(t *testing.T) *policy.Policy { return uPolicy(t, []string{"anderer"}, nil) },
		verhalten{err: e6001}, nil)
	fall("disabled", func(t *testing.T) *policy.Policy { return neuePolicy(t, policy.Werte{Hosts: []string{"h"}}) }, verhalten{err: e6001}, nil)
	mitBody := func(laenge int64, body io.Reader) func(*inbound.Anfrage) {
		return func(a *inbound.Anfrage) { a.Methode, a.Body, a.Laenge = "POST", body, laenge }
	}
	fall("Body über MaxBytes, Laenge", offenste, verhalten{err: e6003}, mitBody(standardMax+1, neuerStrom(muster(standardMax+1, 1))))
	fall("Body über MaxBytes, -1", offenste, verhalten{err: e6003}, mitBody(-1, neuerStrom(muster(standardMax+5, 1))))
	fall("Lesefehler beim Puffern", offenste, verhalten{err: e6003}, mitBody(-1, &fehlerNach{rest: muster(3, 1), err: errors.New("lesefehler")}))
	n += argumentFehler(t, offenste)
	if n != 62 {
		t.Errorf("Zählprüfung: %d Fälle, erwartet 62", n)
	}
}

// argumentFehler ruft den Dienst mit den vier Argumentfehlern (ctx nil, Methode leer, Laenge -2, Body nil bei Laenge 1) und
// hält, dass keine Telemetrie gerufen wird; es liefert die Zahl der Fälle.
func argumentFehler(t *testing.T, pol func(*testing.T) *policy.Policy) int {
	t.Helper()
	nilCtx := func() context.Context { return nil }
	mods := []func(*inbound.Anfrage){
		func(*inbound.Anfrage) {},
		func(a *inbound.Anfrage) { a.Methode = "" },
		func(a *inbound.Anfrage) { a.Laenge = -2 },
		func(a *inbound.Anfrage) { a.Laenge, a.Body = 1, nil },
	}
	for i, mod := range mods {
		u := neueUmgebung(t, pol(t), zielH, verhalten{err: code(codeNamed(t, "ODR-E6003"))}, verhalten{status: 200}, verhalten{status: 200})
		a := anfrage("GET")
		mod(&a)
		var ctx context.Context = u.c
		if i == 0 {
			ctx = nilCtx()
		}
		if _, err := u.g.v.Hybrid(ctx, a); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("Argumentfehler %d: Fehler %v, erwartet model.ErrUngueltig", i, err)
		}
		if got := u.j.text(); got != "" || len(u.j.spanListe()) != 0 {
			t.Errorf("Argumentfehler %d: Folge %q; erwartet leer", i, got)
		}
		pruefeBuch(t, "Argumentfehler", u.j)
	}
	return len(mods)
}

// operationen zählt in f die Selektoren, die eine Operation von Telemetry, FallbackTelemetry oder Span nennen.
func operationen(f *ast.File) map[string]int {
	namen := []string{"Log", "StartSpan", "CountError", "CountDiscarded", "Set", "End"}
	z := map[string]int{}
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok && slices.Contains(namen, s.Sel.Name) {
			z[s.Sel.Name]++
		}
		return true
	})
	return z
}

// TestOperationenZaehlung fängt die Zählung selbst: jede der sechs Operationen wird gefunden, ein anderer Name nicht.
func TestOperationenZaehlung(t *testing.T) {
	quelle := "package p\nfunc f() { a.Log(); a.StartSpan(); a.CountError(); a.CountDiscarded(); s.Set(); s.Set(); s.End(); a.Other(); Set() }\n"
	f, err := parser.ParseFile(token.NewFileSet(), "x.go", quelle, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"Log": 1, "StartSpan": 1, "CountError": 1, "CountDiscarded": 1, "Set": 2, "End": 1}
	if got := operationen(f); !maps.Equal(got, want) {
		t.Errorf("Zählung %v, erwartet %v", got, want)
	}
}

// TestQuelltextTelemetrie hält D9 (SPEC-321 Nr. 1 und 6): der Quelltext des Pakets nennt StartSpan, CountDiscarded, Set und End je
// einmal und weder Log noch CountError.
func TestQuelltextTelemetrie(t *testing.T) {
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	summe, dateien := map[string]int{}, 0
	for _, e := range eintraege {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		dateien++
		for k, v := range operationen(f) {
			summe[k] += v
		}
	}
	want := map[string]int{"StartSpan": 1, "CountDiscarded": 1, "Set": 1, "End": 1}
	if !maps.Equal(summe, want) || dateien < 5 {
		t.Errorf("Operationen im Quelltext %v in %d Dateien, erwartet je einmal %v", summe, dateien, want)
	}
}

// kanarie ist der Wert aus Kleinbuchstaben und Ziffern, den ein Attribut annähme (Token); er ist so lang wie MaxBytes des
// Dienstes im Test (standardMax), damit ein Body mit ihm gepuffert wird.
const kanarie = "kanari7q"

// inhalte sind die Werte, die ein Fall des Dienstes aus Anfrage, Upstream und Konfiguration kennt; ohneUpstream sagt, dass die Policy
// das Ziel nie freigibt (ein Ziel mit Benutzerteil, SPEC-261 Nr. 5 c).
type inhalte struct {
	ziel             string
	hosts            []string
	ohneUpstream     bool
	a                inbound.Anfrage
	rHeader, uHeader map[string][]string
	rBody, uBody     string
	mText, rText     string
	uText            string
}

func neutral() inhalte {
	a := anfrage("GET")
	return inhalte{ziel: zielH, hosts: []string{"h"}, a: a, rHeader: map[string][]string{"X-R": {"1"}}, uHeader: map[string][]string{"X-U": {"1"}},
		rBody: "r", uBody: "u", mText: "m", rText: "r", uText: "u"}
}

// stelle setzt den Canary an eine der Stellen von SPEC-321 Nr. 8.
type stelle struct {
	name  string
	setze func(*inhalte)
}

// stellen sind die 14 Stellen von SPEC-321 Nr. 8.
func stellen() []stelle {
	return []stelle{
		{"Host des Ziels", func(i *inhalte) { i.ziel, i.hosts = "http://"+kanarie, []string{kanarie} }},
		{"Benutzerteil des Ziels", func(i *inhalte) { i.ziel, i.ohneUpstream = "http://"+kanarie+"@h", true }},
		{"Query des Ziels", func(i *inhalte) { i.ziel = "http://h?q=" + kanarie }},
		{"Pfad", func(i *inhalte) { i.a.Pfad = "/odata/svc/" + kanarie }},
		{"Name in der Query", func(i *inhalte) { i.a.RawQuery = kanarie + "=1" }},
		{"Wert in der Query", func(i *inhalte) { i.a.RawQuery = "q=" + kanarie }},
		{"Name eines Headers der Anfrage", func(i *inhalte) { i.a.Header["X-"+kanarie] = []string{"v"} }},
		{"Wert eines Headers der Anfrage", func(i *inhalte) { i.a.Header["X-Wert"] = []string{kanarie} }},
		{"Body der Anfrage", func(i *inhalte) { i.a.Body, i.a.Laenge = strings.NewReader(kanarie), int64(len(kanarie)) }},
		{"Name eines Headers der Antwort", func(i *inhalte) {
			i.rHeader["X-"+kanarie], i.uHeader["X-"+kanarie] = []string{"v"}, []string{"v"}
		}},
		{"Wert eines Headers der Antwort", func(i *inhalte) { i.rHeader["X-R"], i.uHeader["X-U"] = []string{kanarie}, []string{kanarie} }},
		{"Body der Antwort", func(i *inhalte) { i.rBody, i.uBody = kanarie, kanarie }},
		{"Text der Ursache eines Fehlers von Senden", func(i *inhalte) { i.uText = kanarie }},
		{"Text der Ursache eines Fehlers von Mock oder Replay", func(i *inhalte) { i.mText, i.rText = kanarie, kanarie }},
	}
}

// spans ist die Zahl der Spans, die die Art des Fallbacks erzeugt: 1, 1, 2 und 1; ohne Upstream (die Policy gibt das Ziel nie frei)
// entfällt jeder Span der Stufe upstream.
func (i inhalte) spans(art string) int {
	if i.ohneUpstream {
		return map[string]int{"Antwort des Replay": 1, "Kette": 1}[art]
	}
	return map[string]int{"Antwort des Replay": 1, "Antwort des Upstream": 1, "Kette": 2, "Fehler des Upstream": 1}[art]
}

// fall baut den Dienst für eine der vier Arten des Fallbacks.
func (i inhalte) fall(t *testing.T, art string) umgebung {
	t.Helper()
	fehler := func(c model.Code, text string) error {
		return &inbound.Fehler{Code: c, Status: c.Stati()[0], Ursache: errors.New(text)}
	}
	e6001, e6003, miss := model.CodeE6001, model.CodeE6003, model.CodeE5001
	r := verhalten{status: 200, header: i.rHeader, inhalt: i.rBody}
	u := verhalten{status: 200, header: i.uHeader, inhalt: i.uBody}
	switch art {
	case "Antwort des Replay":
		return neueUmgebung(t, polR(t), i.ziel, verhalten{err: fehler(e6003, i.mText)}, r, u)
	case "Antwort des Upstream":
		return neueUmgebung(t, uPolicy(t, i.hosts, nil), i.ziel, verhalten{err: fehler(e6001, i.mText)}, r, u)
	case "Kette":
		return neueUmgebung(t, uPolicy(t, i.hosts, nil, katCoverage), i.ziel, verhalten{err: fehler(e6003, i.mText)},
			verhalten{err: fehler(miss, i.rText), body: true, header: i.rHeader, inhalt: i.rBody}, u)
	}
	u.err, u.body = errors.Join(outbound.ErrUpstreamNichtErreichbar, errors.New(i.uText)), true
	return neueUmgebung(t, uPolicy(t, i.hosts, nil), i.ziel, verhalten{err: fehler(e6001, i.mText)}, r, u)
}

// TestKeinWertInDerTelemetrie hält D13 (SPEC-321 Nr. 8, LH-QA-04.a): Menge M-Stellen-Hybrid, 14 Stellen mal 4 Arten gleich 56; mit
// einem Canary an einer Stelle trägt weder ein Attribut noch ein Argument von CountDiscarded noch die Folge der Aufrufe den
// Canary, und die Spans entstehen (1, 1, 2 und 1 je Art).
func TestKeinWertInDerTelemetrie(t *testing.T) {
	arten := []string{"Antwort des Replay", "Antwort des Upstream", "Kette", "Fehler des Upstream"}
	faelle := 0
	for _, s := range stellen() {
		for _, art := range arten {
			i := neutral()
			s.setze(&i)
			u, erwartet := i.fall(t, art), i.spans(art)
			_, _ = u.g.v.Hybrid(u.c, i.a)
			was := s.name + " / " + art
			text := u.j.text()
			spans := u.j.spanListe()
			for _, sp := range spans {
				text += " " + sp.art()
				for _, a := range sp.start {
					text += " " + a.Key + "=" + a.Text
				}
				for _, satz := range sp.sets {
					for _, a := range satz {
						text += " " + a.Key + "=" + a.Text
					}
				}
			}
			if strings.Contains(strings.ToLower(text), kanarie) {
				t.Errorf("%s: die Telemetrie trägt den Canary: %s", was, text)
			}
			if len(spans) != erwartet {
				t.Errorf("%s: %d Spans, erwartet %d", was, len(spans), erwartet)
			}
			pruefeBuch(t, was, u.j)
			faelle++
		}
	}
	if len(stellen()) != 14 || len(arten) != 4 || faelle != 56 {
		t.Errorf("Zählprüfung: %d Stellen, %d Arten, %d Fälle; erwartet 14, 4 und 56", len(stellen()), len(arten), faelle)
	}
}

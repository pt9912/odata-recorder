package inspection_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

// weg ist ein Weg des Dienstes (Modell-Menge der Wege des Plans): Name, Art und die Zeile der Tabelle
// Zusage · Fänger, die ihn fängt.
type weg struct{ name, art, zeile string }

const (
	artPort      = "Aufruf eines Ports über ein Feld von Konfig"
	artAbbildung = "Aufruf von mitCode"
	artUngueltig = "Stelle, die ErrUngueltig bildet"
)

// wege ist die Tabelle der Wege, die der Quelltext des Pakets tragen muss.
func wege() []weg {
	return []weg{
		{"New: Exchanges nil", artUngueltig, "K1"},
		{"Exchanges: ctx nil", artUngueltig, "A1"},
		{"Exchanges: Nach unter 0", artUngueltig, "X1"},
		{"Exchanges: Limit unter 0", artUngueltig, "X1"},
		{"Exchanges: ListBySession", artPort, "X7"},
		{"Exchanges: Fehler von ListBySession", artAbbildung, "X9"},
	}
}

// TestWegeImQuelltext (W1): die Aufrufe der Ports über Felder von Konfig, die Aufrufe von mitCode und
// die Stellen, die ErrUngueltig bilden, zählt der Test im Quelltext des Pakets (ohne _test) und hält sie
// gleich den Wegen der Tabelle; ein weiterer Aufruf oder eine weitere Stelle ohne Zeile macht ihn rot.
// Grenze: ein Aufruf über eine lokale Variable und ein Aufruf von Abbilden ohne mitCode zählen nicht;
// Get und Append fängt zur Laufzeit TestLesenSchreibtNichts.
func TestWegeImQuelltext(t *testing.T) {
	felderKonfig := map[string]bool{}
	for _, f := range reflect.VisibleFields(reflect.TypeFor[inspection.Konfig]()) {
		felderKonfig[f.Name] = true
	}
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	gezaehlt, dateien := map[string]int{}, 0
	for _, e := range eintraege {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		dateien++
		datei, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(datei, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "mitCode" {
					gezaehlt[artAbbildung]++
				}
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
					if feld, ok := sel.X.(*ast.SelectorExpr); ok && felderKonfig[feld.Sel.Name] {
						gezaehlt[artPort]++
					}
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "ErrUngueltig" {
					gezaehlt[artUngueltig]++
				}
			}
			return true
		})
	}
	if dateien == 0 {
		t.Fatal("keine Quelldatei des Pakets gefunden")
	}
	soll := map[string]int{}
	for _, w := range wege() {
		soll[w.art]++
	}
	if !reflect.DeepEqual(gezaehlt, soll) {
		t.Errorf("im Quelltext %v, Wege der Tabelle %v", gezaehlt, soll)
	}
}

// wegeExchanges ist die Modell-Menge der Wege von Exchanges (Plan §3, Tabelle *Wege*) ohne die Felder
// der Allowlist (X11): jeder Verwurf nach Grund, jede Seitengrenze, jede Fehlerabbildung, Erfolg.
func wegeExchanges() []string {
	return []string{
		"ctx nil", "Nach < 0", "Limit < 0", "Limit 0", "Limit über der Grenze", "Limit sonst", "Kontextfehler",
		"Fehler mit Code", "Erfolg", "Filter", "Schnitt", "Weiter gesetzt", "Weiter 0",
	}
}

// zweig ist ein Fall der Menge W2: die Wege, die er ausführt, und die Erwartung (Art: seite, ungueltig,
// kontext oder code).
type zweig struct {
	name    string
	wege    []string
	art     string
	n       int
	seite   inbound.Seite
	ohneCtx bool
	fehler  error
	code    model.Code
}

// zweige erzeugt die Menge W2: jeder Weg von Exchanges, die Fehler des Ports je Eintrag der Abbildung.
func zweige() []zweig {
	z := []zweig{
		{name: "ctx nil", wege: []string{"ctx nil"}, art: "ungueltig", n: 5, ohneCtx: true},
		{name: "Nach unter 0", wege: []string{"Nach < 0"}, art: "ungueltig", n: 5, seite: inbound.Seite{Nach: -1}},
		{name: "Limit unter 0", wege: []string{"Limit < 0"}, art: "ungueltig", n: 5, seite: inbound.Seite{Limit: -1}},
		{name: "Limit 0, Schnitt", wege: []string{"Limit 0", "Schnitt", "Filter", "Weiter gesetzt", "Erfolg"}, art: "seite", n: 150},
		{name: "Limit über der Grenze", wege: []string{"Limit über der Grenze", "Schnitt", "Weiter gesetzt", "Erfolg"}, art: "seite", n: 1100, seite: inbound.Seite{Limit: math.MaxInt}},
		{name: "Limit sonst, Filter", wege: []string{"Limit sonst", "Filter", "Schnitt", "Weiter gesetzt", "Erfolg"}, art: "seite", n: 5, seite: inbound.Seite{Limit: 2, Nach: 1}},
		{name: "Weiter 0", wege: []string{"Weiter 0", "Erfolg"}, art: "seite", n: 5, seite: inbound.Seite{Limit: 10}},
		{name: "leere Session", wege: []string{"Weiter 0", "Erfolg"}, art: "seite", n: 0},
	}
	for name, e := range map[string]error{
		"Canceled": context.Canceled, "DeadlineExceeded": context.DeadlineExceeded, "verpackt": fmt.Errorf("lesen: %w", context.Canceled),
		"Join vorn": errors.Join(outbound.ErrNichtGefunden, context.Canceled), "Join hinten": errors.Join(context.DeadlineExceeded, outbound.ErrNichtGefunden),
	} {
		z = append(z, zweig{name: "Kontextfehler " + name, wege: []string{"Kontextfehler"}, art: "kontext", n: 1, fehler: e})
	}
	for _, c := range tabelleSession() {
		z = append(z, zweig{name: "Fehler " + c.name, wege: []string{"Fehler mit Code"}, art: "code", n: 1, fehler: fmt.Errorf("lesen: %w", c.err), code: c.code})
	}
	return z
}

// pruefe vergleicht das Ergebnis eines Aufrufs mit der Erwartung des Zweigs.
func (zw zweig) pruefe(t *testing.T, was string, z inbound.ExchangeSeite, err error) {
	t.Helper()
	var fe *inbound.Fehler
	switch zw.art {
	case "seite":
		if err != nil {
			t.Errorf("%s: Fehler %v", was, err)
			return
		}
		gleichSeite(t, was, z, zw.n, zw.seite.Nach, zw.seite.Limit)
	case "ungueltig":
		ohneCode(t, was, err)
	case "kontext":
		if err != zw.fehler || errors.As(err, &fe) {
			t.Errorf("%s: Fehler %v", was, err)
		}
	default:
		if !errors.As(err, &fe) || fe.Code != zw.code {
			t.Errorf("%s: Fehler %v, erwartet Code %v", was, err, zw.code)
		}
	}
}

// TestZweigeSindDieWege (W2, Zählprüfung): die Zweige von TestGleichzeitig führen genau die Wege der
// Modell-Menge, die Fehler mit Code einen je Eintrag von outbound.EintraegeAbbildung() und einen anderen.
func TestZweigeSindDieWege(t *testing.T) {
	gefuehrt, code, kontext := map[string]bool{}, 0, 0
	for _, zw := range zweige() {
		for _, w := range zw.wege {
			gefuehrt[w] = true
		}
		switch zw.art {
		case "code":
			code++
		case "kontext":
			kontext++
		}
	}
	for _, w := range wegeExchanges() {
		if !gefuehrt[w] {
			t.Errorf("der Weg %q hat keinen Zweig", w)
		}
	}
	if len(gefuehrt) != len(wegeExchanges()) {
		t.Errorf("Zweige führen %d Wege, die Modell-Menge hat %d", len(gefuehrt), len(wegeExchanges()))
	}
	if code != len(outbound.EintraegeAbbildung())+1 || kontext != 5 {
		t.Errorf("Fehler mit Code %d, Quelle %d + 1; Kontextfehler %d", code, len(outbound.EintraegeAbbildung()), kontext)
	}
}

// TestGleichzeitig (W2): je Zweig der Menge W2 acht Goroutinen mit je 20 Runden gegen einen Dienst;
// jedes Ergebnis gleicht der Erwartung des Zweigs. Der Dienst hält keinen Zustand; ein Zustand in
// ihm ist ein Datenrennen, das make test-race je Prozess meldet, aber nur in den Zweigen, die der
// Test ausführt, deshalb läuft jeder Zweig der Menge gleichzeitig.
func TestGleichzeitig(t *testing.T) {
	for _, zw := range zweige() {
		f := neuesFake()
		f.legeAn(sid, zw.n)
		f.fehler = zw.fehler
		d := neuerDienst(t, f)
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Go(func() {
				for i := range 20 {
					ctx := t.Context()
					if zw.ohneCtx {
						ctx = nilCtx()
					}
					z, err := d.Exchanges(ctx, sid, zw.seite)
					zw.pruefe(t, fmt.Sprintf("%s, Goroutine %d Runde %d", zw.name, g, i), z, err)
				}
			})
		}
		wg.Wait()
	}
}

package inspection_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
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

// TestGleichzeitig (W2): acht Goroutinen mit je 50 Aufrufen verschiedener Seiten zweier Sessions gegen
// einen Dienst; jedes Ergebnis gleicht dem Modell der Session. Der Dienst hält keinen Zustand; ein
// Zustand in ihm ist ein Datenrennen, das make test-race je Prozess meldet, und ein Zustand mit
// Wirkung (eine Liste, die von einer Session in die andere gerät) verfälscht das Ergebnis.
func TestGleichzeitig(t *testing.T) {
	groessen := map[string]int{sid: 25, "s2": 13}
	f := neuesFake()
	for id, n := range groessen {
		f.legeAn(id, n)
	}
	d := neuerDienst(t, f)
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Go(func() {
			for i := range 50 {
				id := []string{sid, "s2"}[(g+i)%2]
				s := inbound.Seite{Limit: 1 + (g+i)%7, Nach: (g * i) % 20}
				z, err := d.Exchanges(t.Context(), id, s)
				if err != nil {
					t.Errorf("Goroutine %d Aufruf %d: %v", g, i, err)
					continue
				}
				gleichSeite(t, fmt.Sprintf("Goroutine %d Aufruf %d", g, i), z, groessen[id], s.Nach, s.Limit)
			}
		})
	}
	wg.Wait()
}

package hybrid_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
)

// TestLaengeGegenBytes hält Z13 für die Klasse „deklarierte Laenge gegen die gelesenen Bytes“ (SPEC-284 Nr. 5 c): bei
// einem Body mit n Bytes erhalten M und R bis MaxBytes die Laenge n und alle Bytes, wie auch immer die Anfrage Laenge
// deklarierte: -1, jeder Wert 1 bis MaxBytes (kleiner, gleich und größer als n). Oberhalb der Grenze deklariert
// (MaxBytes+1 und MaxBytes+5) gibt es keinen Fallback, und M erhält die deklarierte Laenge. Menge: MaxBytes {1, 2, 3, 5, 8}
// × n 1 bis MaxBytes × Deklaration.
func TestLaengeGegenBytes(t *testing.T) {
	gezaehlt, ueber := 0, 0
	for _, grenze := range []int64{1, 2, 3, 5, 8} {
		for n := int64(1); n <= grenze; n++ {
			deklarationen := []int64{-1}
			for d := int64(1); d <= grenze; d++ {
				deklarationen = append(deklarationen, d)
			}
			for _, d := range deklarationen {
				gezaehlt++
				g := neuesRig(t, offenste(t), grenze)
				g.m.lies, g.r.lies = liesAlles, liesAlles
				g.mockFehler(code(codeNamed(t, "ODR-E6003")))
				g.replayAntwort(antwortVon(200, nil))
				soll := muster(int(n), 11)
				if _, err := g.v.Hybrid(context.Background(), mitBody("POST", bytes.NewReader(bytes.Clone(soll)), d)); err != nil {
					t.Errorf("max %d n %d Laenge %d: %v", grenze, n, d, err)
					continue
				}
				for name, s := range map[string]*stufe{"M": g.m, "R": g.r} {
					a := s.erster(t)
					if a.a.Laenge != n || !bytes.Equal(a.gelesen, soll) {
						t.Errorf("max %d n %d Laenge %d: %s erhielt Laenge %d und %d Bytes", grenze, n, d, name, a.a.Laenge, len(a.gelesen))
					}
				}
			}
			for _, d := range []int64{grenze + 1, grenze + 5} {
				ueber++
				g := neuesRig(t, offenste(t), grenze)
				g.m.lies = liesAlles
				fehler := code(codeNamed(t, "ODR-E6003"))
				g.mockFehler(fehler)
				soll := muster(int(n), 11)
				_, err := g.v.Hybrid(context.Background(), mitBody("POST", bytes.NewReader(bytes.Clone(soll)), d))
				if m := g.m.erster(t); err != error(fehler) || g.r.zahl() != 0 || m.a.Laenge != d || !bytes.Equal(m.gelesen, soll) {
					t.Errorf("max %d n %d Laenge %d: kein Fallback erwartet, Fehler %v, Laenge an M %d", grenze, n, d, err, m.a.Laenge)
				}
			}
		}
	}
	if gezaehlt != 122 || ueber != 38 {
		t.Errorf("Zählprüfung: %d Fälle unterhalb, %d oberhalb der Grenze", gezaehlt, ueber)
	}
}

// TestRootJederPosition hält Z01 für die Klasse „leerer Root an einer Position“: ein leerer Root an jeder Position
// einer Liste der Länge 1 bis 5 weist New ab; dieselben Listen mit lauter belegten Roots sind gültig.
func TestRootJederPosition(t *testing.T) {
	gezaehlt := 0
	for laenge := 1; laenge <= 5; laenge++ {
		for pos := -1; pos < laenge; pos++ {
			var dienste []hybrid.Dienst
			for i := range laenge {
				dienste = append(dienste, hybrid.Dienst{Root: fmt.Sprintf("/r%d", i)})
			}
			if pos >= 0 {
				dienste[pos].Root = ""
				gezaehlt++
			}
			_, err := hybrid.New(hybrid.Konfig{
				Dienste: dienste, Mock: fakeMock{&stufe{}}, Replay: fakeReplay{&stufe{}}, Policy: mitKategorien(t), MaxBytes: 1,
			})
			if (pos >= 0) != errors.Is(err, model.ErrUngueltig) || (pos < 0 && err != nil) {
				t.Errorf("Länge %d, leer an %d: Fehler %v", laenge, pos, err)
			}
		}
	}
	if gezaehlt != 15 {
		t.Errorf("Zählprüfung: %d Listen mit leerem Root", gezaehlt)
	}
}

// TestVerpackteFehler hält Z03 für die Klasse „Verpackung des Fehlers der Stufe“ (SPEC-284 Nr. 3 b): ein *Fehler, direkt,
// mit fmt.Errorf (%w, einfach und doppelt) und in einem errors.Join (vorn, hinten, verschachtelt), löst für jeden Code
// aus model.AlleCodes() den Replay genau dann aus, wenn der Code ein Auslöser ist; sonst kehrt derselbe Wert zurück.
func TestVerpackteFehler(t *testing.T) {
	arten := map[string]func(f error) error{
		"direkt":          func(f error) error { return f },
		"Errorf":          func(f error) error { return fmt.Errorf("mock: %w", f) },
		"Errorf zweifach": func(f error) error { return fmt.Errorf("a: %w", fmt.Errorf("b: %w", f)) },
		"Join vorn":       func(f error) error { return errors.Join(f, errors.New("x")) },
		"Join hinten":     func(f error) error { return errors.Join(errors.New("x"), f) },
		"Errorf Join":     func(f error) error { return fmt.Errorf("c: %w", errors.Join(errors.New("x"), f)) },
		"Join Join":       func(f error) error { return errors.Join(errors.Join(f), errors.New("x")) },
	}
	alle := model.AlleCodes()
	gezaehlt := 0
	for name, verpacke := range arten {
		for _, c := range alle {
			gezaehlt++
			g := neuesRig(t, offenste(t), standardMax)
			fehler := verpacke(code(c))
			g.mockFehler(fehler)
			g.replayAntwort(antwortVon(200, nil))
			_, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			if ausloeserKategorie(c.String()) != "" {
				if err != nil || g.r.zahl() != 1 {
					t.Errorf("%s %s: Replay erwartet, Fehler %v, Replay %d", name, c, err, g.r.zahl())
				}
				continue
			}
			if err != fehler || g.r.zahl() != 0 {
				t.Errorf("%s %s: derselbe Fehler erwartet, Replay %d", name, c, g.r.zahl())
			}
		}
	}
	if gezaehlt != len(arten)*len(alle) || len(arten) != 7 {
		t.Errorf("Zählprüfung: %d Fälle, %d Arten", gezaehlt, len(arten))
	}
}

// TestVerteilerFelder hält Z31 für die Klasse „Zustand im Verteiler“ (SPEC-284 Nr. 13): der Typ Verteiler hat genau die
// Felder der Konfig (mock, replay, policy, maxBytes) mit deren Typen, kein veränderliches Feld, keinen Kanal, keine Map
// und keine Sperre. Grenze: gelesen wird der Typ in verteiler.go; ein Feld in einer anderen Datei ist ein anderer Typ.
func TestVerteilerFelder(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "verteiler.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	var felder []string
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, s := range gd.Specs {
			ts := s.(*ast.TypeSpec)
			st, ok := ts.Type.(*ast.StructType)
			if ts.Name.Name != "Verteiler" || !ok {
				continue
			}
			for _, fl := range st.Fields.List {
				for _, n := range fl.Names {
					felder = append(felder, n.Name+" "+types.ExprString(fl.Type))
				}
				if len(fl.Names) == 0 {
					felder = append(felder, "eingebettet "+types.ExprString(fl.Type))
				}
			}
		}
	}
	want := []string{"mock inbound.MockUseCase", "replay inbound.ReplayUseCase", "policy *policy.Policy", "maxBytes int64"}
	slices.Sort(felder)
	slices.Sort(want)
	if !slices.Equal(felder, want) {
		t.Errorf("Felder des Verteiler %v, erwartet %v", felder, want)
	}
}

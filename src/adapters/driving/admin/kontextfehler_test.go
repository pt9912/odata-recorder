package admin_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestKontextfehlerOhneAntwort hält X1 (Nr. 9) auf einem laufenden Server: ein Kontextfehler der Route (context.Canceled,
// context.DeadlineExceeded) ergibt keine Antwort (der Client sieht einen Transportfehler), kein Zählen und keinen Eintrag,
// in jeder Form: direkt, verpackt, im Join, neben einem *inbound.Fehler, neben ErrUngueltig (M-Verpackung von X1) und als
// Ursache eines *inbound.Fehler und neben einem nil-Zeiger (zwei weitere Formen); 2 Fehler × 7 Formen.
func TestKontextfehlerOhneAntwort(t *testing.T) {
	var nil1 *inbound.Fehler
	formen := []verpackung{
		{"direkt", func(e error) error { return e }},
		{"%w", func(e error) error { return fmt.Errorf("verpackt: %w", e) }},
		{"Join", func(e error) error { return errors.Join(errors.New("anderer"), e) }},
		{"neben *inbound.Fehler", func(e error) error {
			return errors.Join(&inbound.Fehler{Code: model.CodeE7003, Status: 404}, e)
		}},
		{"neben ErrUngueltig", func(e error) error { return errors.Join(model.ErrUngueltig, e) }},
		{"Ursache eines *inbound.Fehler", func(e error) error {
			return &inbound.Fehler{Code: model.CodeE7003, Status: 404, Ursache: e}
		}},
		{"neben nil-Zeiger", func(e error) error { return errors.Join(nil1, e) }},
	}
	faelle := 0
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, f := range formen {
			faelle++
			name := e.Error() + ", " + f.name
			route := neueRoute()
			route.err = f.packe(e)
			u := neuLive(t, route)
			if x := erwarte(t, u.async("GET", "/p"), name); x.err == nil {
				t.Errorf("%s: der Client sah eine Antwort mit Status %d, erwartet einen Transportfehler", name, x.status)
			}
			if codes, andere := zaehlungenVon(u.tele); len(codes) != 0 || andere != 0 || u.audit.Aufrufe() != 0 {
				t.Errorf("%s: CountError %q, andere Aufrufe der Telemetrie %d, Audit %d, erwartet 0", name, codes, andere, u.audit.Aufrufe())
			}
			gleichAuf(t, name+": Aufrufe der Route", route.aufrufe(), 1)
		}
	}
	gleichAuf(t, "Zahl der Fälle", faelle, 14)
}

// zaehleSelektoren zählt in einer Quelle die Ausdrücke `paket.name` (paket leer: jeder Ausdruck mit diesem Namen).
func zaehleSelektoren(t *testing.T, name string, quelle any, paket, sel string) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, 0)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	n := 0
	ast.Inspect(f, func(x ast.Node) bool {
		if s, ok := x.(*ast.SelectorExpr); ok && s.Sel.Name == sel {
			if id, ok := s.X.(*ast.Ident); paket == "" || ok && id.Name == paket {
				n++
			}
		}
		return true
	})
	return n
}

// TestEineStelleSchreibtUndZaehlt hält C3 (Nr. 7): im Paket ohne _test.go gibt es genau eine Stelle, die
// `model.NeuesFehlerobjekt` ruft oder nennt, und genau eine, die `CountError` ruft oder nennt. Die Probe hält den Zähler
// gegen eine Quelle mit zwei Stellen und mit einer Namensgleichheit (`x.CountError` zählt, `model.Anderes` nicht).
func TestEineStelleSchreibtUndZaehlt(t *testing.T) {
	probe := "package x\nfunc f() { model.NeuesFehlerobjekt(1); _ = model.NeuesFehlerobjekt; model.Anderes(); s.CountError(1); t.CountError(2) }"
	gleichAuf(t, "Probe: Stellen von NeuesFehlerobjekt", zaehleSelektoren(t, "probe.go", probe, "model", "NeuesFehlerobjekt"), 2)
	gleichAuf(t, "Probe: Stellen von CountError", zaehleSelektoren(t, "probe.go", probe, "", "CountError"), 2)

	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	dateien, objekte, zaehler := 0, 0, 0
	for _, e := range eintraege {
		if n := e.Name(); strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			src, err := os.ReadFile(n)
			if err != nil {
				t.Fatal(err)
			}
			dateien++
			objekte += zaehleSelektoren(t, n, src, "model", "NeuesFehlerobjekt")
			zaehler += zaehleSelektoren(t, n, src, "", "CountError")
		}
	}
	if dateien < 4 || objekte != 1 || zaehler != 1 {
		t.Errorf("%d Dateien ohne _test.go: %d Stellen von NeuesFehlerobjekt und %d von CountError, erwartet mindestens 4 Dateien und je 1", dateien, objekte, zaehler)
	}
}

// pfadeDerTypen liefert die Paketpfade der benannten Typen, aus denen t besteht (Elemente von Zeigern, Folgen und
// Abbildungen eingeschlossen).
func pfadeDerTypen(t reflect.Type) []string {
	var out []string
	if t.PkgPath() != "" {
		out = append(out, t.PkgPath())
	}
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		out = append(out, pfadeDerTypen(t.Elem())...)
	case reflect.Map:
		out = append(out, pfadeDerTypen(t.Key())...)
		out = append(out, pfadeDerTypen(t.Elem())...)
	}
	return out
}

// TestRouteOhnePort hält C4 (Nr. 7, 15): `Handler` ist `func(context.Context, Anfrage) (Antwort, error)` (die Zuweisung
// übersetzt nur mit dieser Signatur), die Felder von `Anfrage` und `Antwort` sind genau die des Vertrags, und kein Typ
// darin stammt aus dem Paket der Querschnitts-Ports.
func TestRouteOhnePort(t *testing.T) {
	var _ func(context.Context, admin.Anfrage) (admin.Antwort, error) = admin.Handler(nil)
	h := reflect.TypeFor[admin.Handler]()
	if h.NumIn() != 2 || h.NumOut() != 2 || h.In(0) != reflect.TypeFor[context.Context]() || h.In(1) != reflect.TypeFor[admin.Anfrage]() ||
		h.Out(0) != reflect.TypeFor[admin.Antwort]() || h.Out(1) != reflect.TypeFor[error]() {
		t.Errorf("Signatur von Handler: %v", h)
	}
	soll := map[reflect.Type]map[string]reflect.Type{
		reflect.TypeFor[admin.Anfrage](): {
			"Methode": reflect.TypeFor[string](), "Pfad": reflect.TypeFor[string](), "RawQuery": reflect.TypeFor[string](),
			"KorrelationsID": reflect.TypeFor[string](), "Pfadwerte": reflect.TypeFor[map[string]string](),
			"Header": reflect.TypeFor[map[string][]string](), "Body": reflect.TypeFor[io.Reader](), "Laenge": reflect.TypeFor[int64](),
		},
		reflect.TypeFor[admin.Antwort](): {"Status": reflect.TypeFor[int](), "Body": reflect.TypeFor[[]byte]()},
	}
	for typ, felder := range soll {
		gleichAuf(t, typ.Name()+": Zahl der Felder", typ.NumField(), len(felder))
		for i := range typ.NumField() {
			f := typ.Field(i)
			if felder[f.Name] != f.Type {
				t.Errorf("%s.%s hat den Typ %v, erwartet %v", typ.Name(), f.Name, f.Type, felder[f.Name])
			}
			for _, p := range pfadeDerTypen(f.Type) {
				if strings.Contains(p, "/ports/support") {
					t.Errorf("%s.%s trägt einen Typ aus %s", typ.Name(), f.Name, p)
				}
			}
		}
	}
}

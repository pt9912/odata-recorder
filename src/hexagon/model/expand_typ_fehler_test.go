package model_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// exvUrsachenAusQuelle liest aus expand_typ_fehler.go die Werte aller Konstanten vom Typ ExpandUrsache (jede Konstante
// nennt ihren Typ selbst).
func exvUrsachenAusQuelle(t *testing.T) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "expand_typ_fehler.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var werte []string
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.CONST {
			continue
		}
		for _, s := range g.Specs {
			v := s.(*ast.ValueSpec)
			if id, ok := v.Type.(*ast.Ident); !ok || id.Name != "ExpandUrsache" {
				continue
			}
			for _, e := range v.Values {
				w, err := strconv.Unquote(e.(*ast.BasicLit).Value)
				if err != nil {
					t.Fatal(err)
				}
				werte = append(werte, w)
			}
		}
	}
	return werte
}

type exvZeile struct {
	ursache  model.ExpandUrsache
	sentinel error
	code     model.Code
	praefix  string
}

// exvTabelle ist die Tabelle von SPEC-263 Nr. 3.
func exvTabelle() []exvZeile {
	return []exvZeile{
		{model.ExpandNavigationUnbekannt, model.ErrURLUngueltig, model.CodeE6005, "ODR-E6005: Expand ungültig: "},
		{model.ExpandKeineNavigation, model.ErrURLUngueltig, model.CodeE6005, "ODR-E6005: Expand ungültig: "},
		{model.ExpandNavigationNichtUnterstuetzt, model.ErrURLNichtUnterstuetzt, model.CodeE6001, "ODR-E6001: Expand nicht unterstützt: "},
	}
}

func exvSentinels() []error {
	return []error{model.ErrURLUngueltig, model.ErrURLNichtUnterstuetzt, model.ErrFilterNichtBeobachtet, model.ErrUngueltig}
}

// exvPruefeFehler prüft einen Fehler gegen eine Zeile der Tabelle: errors.Is trifft genau die Sentinel der Klasse (auch
// verpackt und verbunden), Code und Text stehen fest, errors.As findet den ExpandFehler.
func exvPruefeFehler(t *testing.T, was string, err error, z exvZeile) {
	t.Helper()
	var ef *model.ExpandFehler
	if !errors.As(err, &ef) || ef.Ursache != z.ursache {
		t.Errorf("%s: kein ExpandFehler der Ursache %s: %v", was, z.ursache, err)
		return
	}
	if ef.Code() != z.code || err.Error() != z.praefix+string(z.ursache) {
		t.Errorf("%s: %v %q", was, ef.Code(), err.Error())
	}
	for _, s := range exvSentinels() {
		for form, e := range map[string]error{"roh": err, "verpackt": fmt.Errorf("x: %w", err), "verbunden": errors.Join(errors.New("x"), err)} {
			if errors.Is(e, s) != (s == z.sentinel) {
				t.Errorf("%s: %s: errors.Is(%v) = %v", was, form, s, s != z.sentinel)
			}
		}
	}
}

// exvAusFunktion liefert zu einer Ursache einen Typ und einen Knoten, bei denen PruefeExpand genau sie meldet.
func exvAusFunktion(u model.ExpandUrsache) (model.EntityType, *model.ExpandOption) {
	typ := model.EntityType{Name: "Ns.T", Properties: []model.Property{exvP("P", "Edm.String")}, NavigationProperties: []model.NavigationProperty{exvN("N", "Ns.T", false)}}
	switch u {
	case model.ExpandKeineNavigation:
		return typ, exvO("N", "P")
	case model.ExpandNavigationNichtUnterstuetzt:
		typ.Open = true
	}
	return typ, exvO("N", "Fremd")
}

// TestExpandFehler (Zeile 10): jede der drei Ursachen trägt Klasse, Code und Text der Tabelle von SPEC-263 Nr. 3, von Hand
// gebaut und aus PruefeExpand; errors.Is trifft genau die Sentinel der Klasse und nie ErrUngueltig. Zählprüfung gegen die
// Quelle: expand_typ_fehler.go führt genau die Ursachen der Tabelle.
func TestExpandFehler(t *testing.T) {
	tabelle := exvTabelle()
	soll := make([]string, 0, len(tabelle))
	for _, z := range tabelle {
		soll = append(soll, string(z.ursache))
	}
	ist := exvUrsachenAusQuelle(t)
	slices.Sort(soll)
	slices.Sort(ist)
	if !slices.Equal(ist, soll) {
		t.Errorf("Ursachen der Quelle %v, der Tabelle %v", ist, soll)
	}
	for _, z := range tabelle {
		exvPruefeFehler(t, "von Hand "+string(z.ursache), &model.ExpandFehler{Ursache: z.ursache}, z)
		typ, o := exvAusFunktion(z.ursache)
		exvPruefeFehler(t, "aus PruefeExpand "+string(z.ursache), model.PruefeExpand(typ, o), z)
	}
}

// TestExpandFehlerFremdeUrsache (Zeile 11): eine Ursache außerhalb der drei hat die Klasse der ersten Zeile, auch die einer
// der drei in anderer Schreibweise.
func TestExpandFehlerFremdeUrsache(t *testing.T) {
	ersteZeile := exvTabelle()[0]
	fremde := []string{"", "x", "navigation_unbekannt "}
	for _, z := range exvTabelle() {
		u := string(z.ursache)
		fremde = append(fremde, strings.ToUpper(u), strings.ToUpper(u[:1])+u[1:], u+"x")
	}
	for _, f := range fremde {
		ersteZeile.ursache = model.ExpandUrsache(f)
		exvPruefeFehler(t, fmt.Sprintf("Ursache %q", f), &model.ExpandFehler{Ursache: ersteZeile.ursache}, ersteZeile)
	}
}

// TestPruefeExpandFehlertext (Zeile 12): weder der Name noch die Stelle noch der Name des Typs steht im Fehler: jeder Verstoß
// an Stelle 0 bis 2, im offenen und im nicht offenen Typ, hat auf allen drei Wegen der Ausgabe denselben Text, den der Tabelle.
func TestPruefeExpandFehlertext(t *testing.T) {
	geheim := "Geheim" + string([]byte{0xc3, 0xa4})
	typ := model.EntityType{Name: "Ns.Typ" + geheim, Properties: []model.Property{exvP("Prop"+geheim, "Edm.String")}}
	for i := range 3 {
		typ.NavigationProperties = append(typ.NavigationProperties, exvN(fmt.Sprintf("Nav%s%d", geheim, i), "Ns.Typ"+geheim, false))
	}
	n := 0
	for _, z := range exvTabelle() {
		for _, offen := range []bool{false, true} {
			tt := typ
			tt.Open = offen
			if offen != (z.ursache == model.ExpandNavigationNichtUnterstuetzt) && z.ursache != model.ExpandKeineNavigation {
				continue
			}
			verstoss := "Name" + geheim
			if z.ursache == model.ExpandKeineNavigation {
				verstoss = "Prop" + geheim
			}
			for stelle := range 3 {
				namen := make([]string, 0, 3)
				for i := range stelle {
					namen = append(namen, tt.NavigationProperties[i].Name)
				}
				err := model.PruefeExpand(tt, exvO(append(namen, verstoss)...))
				exvPruefeFehler(t, fmt.Sprintf("%s Stelle %d offen=%v", z.ursache, stelle, offen), err, z)
				for _, s := range []string{err.Error(), fmt.Sprintf("%v", err), fmt.Sprintf("%+v", err)} {
					if strings.Contains(s, geheim) || strings.Contains(s, tt.Name) || s != z.praefix+string(z.ursache) {
						t.Errorf("%s Stelle %d: Text %q", z.ursache, stelle, s)
					}
				}
				n++
			}
		}
	}
	exvZaehlt(t, "Fehlertext", n, 3*(len(exvTabelle())+1))
}

package integration_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// liesVergleich liest `$filter` mit dem Leser des Filters und liefert den Vergleich der Wurzel.
func liesVergleich(t *testing.T, filter string) model.FilterVergleich {
	t.Helper()
	baum, err := urlparser.NewFilterLeser().LeseFilter(filter)
	if err != nil {
		t.Fatalf("%q: %v", filter, err)
	}
	v, ok := baum.(model.FilterVergleich)
	if !ok {
		t.Fatalf("%q: Wurzel %T", filter, baum)
	}
	return v
}

func wertVonRechts(t *testing.T, v model.FilterVergleich) model.Wert {
	t.Helper()
	l, ok := v.Rechts.(model.FilterLiteral)
	if !ok {
		t.Fatalf("Rechts ist %T", v.Rechts)
	}
	w, err := model.WertAusLiteral(l.Literal)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// TestInstantVerschiedenerVersatz (SPEC-166 Nr. 3 e, SPEC-161 Nr. 6 f): Zeitpunkte, die der Leser des Filters aus Schreibweisen mit
// verschiedenem Versatz liest, vergleichen sich nach dem Instant; die drei Schreibweisen des Beispiels und `00:30+01:00` gegen
// `23:30Z` des Vortags sind gleich, dazu alle Paare aus vier Instants in sieben Versätzen (mit und ohne Bruchteil) gegen den Vergleich
// von time.Time.
func TestInstantVerschiedenerVersatz(t *testing.T) {
	gleich := []string{"2020-01-01T10:00:00+02:00", "2020-01-01T08:00:00Z", "2020-01-01T07:00:00-01:00"}
	for _, a := range gleich {
		for _, b := range gleich {
			va, vb := wertVonRechts(t, liesVergleich(t, "At eq "+a)), wertVonRechts(t, liesVergleich(t, "At eq "+b))
			if got, err := model.Vergleiche(model.VergleichEq, va, vb); err != nil || got != model.WahrheitWahr {
				t.Errorf("%s eq %s: %d %v", a, b, got, err)
			}
		}
	}
	va := wertVonRechts(t, liesVergleich(t, "At eq 2020-01-01T00:30:00+01:00"))
	vb := wertVonRechts(t, liesVergleich(t, "At eq 2019-12-31T23:30:00Z"))
	if got, err := model.Vergleiche(model.VergleichEq, va, vb); err != nil || got != model.WahrheitWahr {
		t.Errorf("über das Jahr: %d %v", got, err)
	}

	instants := []string{"2019-12-31T23:30:00", "2020-01-01T10:00:00", "2020-02-29T23:59:59", "2020-12-31T23:30:00.25"}
	versaetze := []int{0, 2 * 60, -60, 5*60 + 30, -(23*60 + 59), 23*60 + 59, 60}
	type schreibweise struct {
		text    string
		instant time.Time
	}
	var alle []schreibweise
	for _, s := range instants {
		utc, err := time.Parse("2006-01-02T15:04:05.999999999", s)
		if err != nil {
			t.Fatal(err)
		}
		for _, min := range versaetze {
			lokal := utc.In(time.FixedZone("", min*60))
			vorzeichen, abs := '+', min
			if min < 0 {
				vorzeichen, abs = '-', -min
			}
			text := lokal.Format("2006-01-02T15:04:05.999999999") + fmt.Sprintf("%c%02d:%02d", vorzeichen, abs/60, abs%60)
			alle = append(alle, schreibweise{text, utc})
		}
	}
	ops := []model.VergleichOp{"eq", "ne", "gt", "ge", "lt", "le"}
	n := 0
	for _, a := range alle {
		wa := wertVonRechts(t, liesVergleich(t, "At eq "+a.text))
		for _, b := range alle {
			wb := wertVonRechts(t, liesVergleich(t, "At eq "+b.text))
			c := a.instant.Compare(b.instant)
			for _, op := range ops {
				want := map[model.VergleichOp]bool{"eq": c == 0, "ne": c != 0, "gt": c > 0, "ge": c >= 0, "lt": c < 0, "le": c <= 0}[op]
				if got, err := model.Vergleiche(op, wa, wb); err != nil || (got == model.WahrheitWahr) != want {
					t.Errorf("%s %s %s: %d %v, erwartet %v", a.text, op, b.text, got, err, want)
				}
				n++
			}
		}
	}
	if n != 28*28*6 {
		t.Errorf("%d Vergleiche", n)
	}
}

// TestFilterAusLeser (SPEC-167 Nr. 3 und 4): Bäume, die der Leser des Filters liefert, laufen durch die Typprüfung; die Randformen, die
// der Leser nicht prüft (SPEC-164 Nr. 10), ergeben den Code ihrer Ursache.
func TestFilterAusLeser(t *testing.T) {
	typ := model.EntityType{Name: "Ns.Product", Properties: []model.Property{
		{Name: "Id", EdmType: "Edm.Int32"}, {Name: "Price", EdmType: "Edm.Decimal"}, {Name: "Name", EdmType: "Edm.String"},
		{Name: "Active", EdmType: "Edm.Boolean"}, {Name: "Ref", EdmType: "Edm.Guid"}, {Name: "Bin", EdmType: "Edm.Binary"},
	}}
	for _, c := range []struct {
		filter  string
		ursache model.FilterTypUrsache
	}{
		{"Id eq 1", ""}, {"Id eq 1.0", ""}, {"Id eq 007", ""}, {"Id eq -0", ""}, {"Id eq 2147483647", ""}, {"Price eq 99999999999999999999.5", ""},
		{"Id eq 2147483648", model.FilterLiteralBereich}, {"Id eq 1.5", model.FilterLiteralBereich},
		{"not Price gt 5", model.FilterNichtBoolesch}, {"Id", model.FilterNichtBoolesch}, {"Active", ""}, {"true", ""}, {"null", ""},
		{"1 eq 1", ""}, {"1 eq 'a'", model.FilterTypUnvertraeglich}, {"Id eq 'x'", model.FilterTypUnvertraeglich},
		{"Id eq null", ""}, {"null eq null", ""}, {"Name eq 'a''b'", ""}, {"Foo eq 1", model.FilterEigenschaftUnbekannt},
		{"Bin eq null", model.FilterTypNichtUnterstuetzt},
		{"Ref gt 00000000-0000-0000-0000-000000000000", model.FilterOperatorNichtUnterstuetzt},
		{"Ref eq 00000000-0000-0000-0000-000000000000", ""},
		{"Active and Id gt 1 or not Active", ""}, {"Active and Id", model.FilterNichtBoolesch},
	} {
		baum, err := urlparser.NewFilterLeser().LeseFilter(c.filter)
		if err != nil {
			t.Errorf("%q: Leser: %v", c.filter, err)
			continue
		}
		err = model.PruefeFilter(typ, baum, nil)
		var ft *model.FilterTypFehler
		switch {
		case c.ursache == "" && err != nil:
			t.Errorf("%q: %v", c.filter, err)
		case c.ursache != "" && (!errors.As(err, &ft) || ft.Ursache != c.ursache):
			t.Errorf("%q: %v, erwartet %s", c.filter, err, c.ursache)
		}
	}
}

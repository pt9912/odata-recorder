package mock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestRedigiertAnJederPosition hält SPEC-196 Nr. 4 e (V-1): die Werte `redigiert` entfernt der Dienst aus jeder Entity der
// Collection. Menge: Position 4 (erste, zweite, dritte, vierte) und alle Positionen zugleich = 5 Läufe.
func TestRedigiertAnJederPosition(t *testing.T) {
	const n = 4
	for _, p := range []int{0, 1, 2, 3, -1} {
		var es []model.EntityState
		for i := 0; i < n; i++ {
			name := fmt.Sprintf(`"n%d"`, i)
			if p == i || p == -1 {
				name = `"` + model.Marker + `"`
			}
			es = append(es, eintrag(setProducts, fmt.Sprint(i+1), fmt.Sprintf(`{"Id":%d,"Name":%s}`, i+1, name), alleNamen()...))
		}
		u := leseUmg(t)
		u.en.setze(setProducts, es...)
		u.menge(model.Optionen{})
		if _, err := u.leseGET(); err != nil {
			t.Fatal(err)
		}
		c := u.letzteCollection()
		if len(c.Entities) != n {
			t.Fatalf("%d Entities", len(c.Entities))
		}
		for i, e := range c.Entities {
			_, da := wertVon(e, "Name")
			if soll := p != i && p != -1; da != soll {
				t.Errorf("Position %d, redigiert an %d: Name da=%v, erwartet %v", i, p, da, soll)
			}
		}
	}
}

// TestSelectAlleMitNamen hält SPEC-196 Nr. 7 e (V-2): bei `Alle` wahr gelten alle unterstützten Properties als angefordert, auch
// wenn `$select` Namen nennt. Menge: Form 4 (`*`, `*,Name`, `*,Id,Name`, `*,Unbekannt`) × Eintrag 2 (kennt nur Id und Name;
// vollständig) = 8 Läufe am Lookup.
func TestSelectAlleMitNamen(t *testing.T) {
	formen := []*model.SelectOption{{Alle: true}, {Alle: true, Namen: []string{"Name"}}, {Alle: true, Namen: []string{"Id", "Name"}}, {Alle: true, Namen: []string{"Unbekannt"}}}
	eintraege := []struct {
		name  string
		known []string
		soll  [4]model.Code
	}{
		{"partial", []string{"Id", "Name"}, [4]model.Code{model.CodeE6003, model.CodeE6003, model.CodeE6003, model.CodeE6005}},
		{"vollständig", alleNamen(), [4]model.Code{0, 0, 0, model.CodeE6005}},
	}
	for _, e := range eintraege {
		for i, f := range formen {
			u := leseUmg(t)
			u.en.setze(setProducts, eintrag(setProducts, "1", `{"Id":1,"Name":"a"}`, e.known...))
			u.entity(model.Optionen{Select: f}, "1")
			_, err := u.leseGET()
			if fehlercode(err) != e.soll[i] {
				t.Errorf("%s, Form %d: %v, erwartet %s", e.name, i, err, e.soll[i])
			}
		}
	}
}

// literalZu liefert zu jeder der sechs Literal-Arten ein Literal, das der Parser liefern kann.
func literalZu(a model.LiteralArt) model.Literal {
	texte := map[model.LiteralArt]string{
		model.LiteralZeichenkette: "'a'", model.LiteralWahrheitswert: "true", model.LiteralZahl: "1",
		model.LiteralGUID: "00000000-0000-0000-0000-000000000000", model.LiteralDatum: "2026-11-07", model.LiteralZeitpunkt: "2026-11-07T00:00:00Z",
	}
	return model.Literal{Art: a, Text: texte[a]}
}

// TestLiteralMatrix hält SPEC-196 Nr. 7 a (V-3): jede Literal-Art zu jedem EDM-Typ des Schlüssels; nur die acht Paare der
// Spezifikation passen (Orakel von Hand). Menge: 6 Literal-Arten × 8 Typen = 48 Läufe.
func TestLiteralMatrix(t *testing.T) {
	arten := []model.LiteralArt{model.LiteralZeichenkette, model.LiteralWahrheitswert, model.LiteralZahl, model.LiteralGUID, model.LiteralDatum, model.LiteralZeitpunkt}
	passt := map[string]model.LiteralArt{
		"Edm.String": model.LiteralZeichenkette, "Edm.Boolean": model.LiteralWahrheitswert, "Edm.Int32": model.LiteralZahl, "Edm.Int64": model.LiteralZahl,
		"Edm.Decimal": model.LiteralZahl, "Edm.Guid": model.LiteralGUID, "Edm.Date": model.LiteralDatum, "Edm.DateTimeOffset": model.LiteralZeitpunkt,
	}
	n := 0
	for i, edm := range achtTypen() {
		for _, a := range arten {
			n++
			u := leseUmg(t)
			u.c.mc.modell = schluesselModell()
			u.ressource(model.RessourceEntity, fmt.Sprintf("K%d", i), model.Optionen{}, model.SchluesselTeil{Name: "K", Literal: literalZu(a)})
			_, _ = u.leseGET()
			if (u.p.anzahl("Entities.Get") == 1) != (passt[edm] == a) {
				t.Errorf("%s mit Literal-Art %d: Get %d Mal, passt=%v", edm, a, u.p.anzahl("Entities.Get"), passt[edm] == a)
			}
		}
	}
	if n != 48 {
		t.Errorf("%d Läufe, erwartet 48", n)
	}
}

// TestZahlGrenzen hält SPEC-196 Nr. 7 a (V-3): `Edm.Int32` und `Edm.Int64` nehmen eine ganze Zahl im Bereich des Typs; beide Seiten
// jeder Grenze. Menge: 2 Typen × (Minimum, Minimum − 1, Maximum, Maximum + 1) = 8 Läufe.
func TestZahlGrenzen(t *testing.T) {
	faelle := []struct {
		set, text string
		ok        bool
	}{
		{"K2", "-2147483648", true}, {"K2", "-2147483649", false}, {"K2", "2147483647", true}, {"K2", "2147483648", false},
		{"K3", "-9223372036854775808", true}, {"K3", "-9223372036854775809", false}, {"K3", "9223372036854775807", true}, {"K3", "9223372036854775808", false},
	}
	for _, f := range faelle {
		u := leseUmg(t)
		u.c.mc.modell = schluesselModell()
		u.ressource(model.RessourceEntity, f.set, model.Optionen{}, zahl("K", f.text))
		_, err := u.leseGET()
		if f.ok != (u.p.anzahl("Entities.Get") == 1) || !f.ok && fehlercode(err) != model.CodeE6005 {
			t.Errorf("%s %s: Get %d Mal, Fehler %v, ok=%v", f.set, f.text, u.p.anzahl("Entities.Get"), err, f.ok)
		}
	}
}

// TestUnbenannterTeilBeiZweiKeys hält SPEC-196 Nr. 7 a (V-3): ein unbenannter Teil gilt nur bei einem Schlüsselteil; ein
// unbenannter neben einem benannten Teil bei zwei Keys ist ODR-E6005. Menge: Stelle des unbenannten Teils 2 = 2 Läufe.
func TestUnbenannterTeilBeiZweiKeys(t *testing.T) {
	str := model.Literal{Art: model.LiteralZeichenkette, Text: "'a'"}
	zahl1 := model.Literal{Art: model.LiteralZahl, Text: "1"}
	teile := [][]model.SchluesselTeil{
		{{Literal: zahl1}, {Name: "S", Literal: str}},
		{{Name: "I", Literal: zahl1}, {Literal: str}},
	}
	for i, tl := range teile {
		u := leseUmg(t)
		u.c.mc.modell = schluesselModell()
		u.ressource(model.RessourceEntity, "T2", model.Optionen{}, tl...)
		_, err := u.leseGET()
		_ = erwarteFehler(t, fmt.Sprintf("Form %d", i), err, model.CodeE6005)
	}
}

// TestTypOhneRaute hält SPEC-196 Nr. 3 a (V-4): `@odata.type` beginnt mit `#`; ein anderes erstes Zeichen vor einem gültigen Typnamen
// ist ODR-E9001. Menge: erstes Zeichen 7 (x, $, @, Leerzeichen, 0, `.`, `/`) = 7 Läufe (der Name danach ist `N.Special`).
func TestTypOhneRaute(t *testing.T) {
	for _, z := range []string{"x", "$", "@", " ", "0", ".", "/"} {
		u := leseUmg(t)
		_, err := u.einEintrag(`{"@odata.type":"` + z + `N.Special","Id":1}`)
		_ = erwarteFehler(t, "erstes Zeichen "+z, err, model.CodeE9001)
	}
}

// TestJSONTypenMatrix hält SPEC-196 Nr. 3 b (V-6): jeder EDM-Typ nimmt genau die JSON-Art der Spezifikation. Menge: 8 Typen × 5
// JSON-Arten (Zahl, String, Boolean, Array, Objekt) = 40 Läufe, 8 gültige Zellen.
func TestJSONTypenMatrix(t *testing.T) {
	arten := []struct{ name, json string }{{"Zahl", `1`}, {"String", `"x"`}, {"Boolean", `true`}, {"Array", `[]`}, {"Objekt", `{}`}}
	gueltig := map[string]string{
		"Edm.String": "String", "Edm.Boolean": "Boolean", "Edm.Int32": "Zahl", "Edm.Int64": "String", "Edm.Decimal": "String",
		"Edm.Guid": "String", "Edm.Date": "String", "Edm.DateTimeOffset": "String",
	}
	n := 0
	for _, edm := range achtTypen() {
		for _, a := range arten {
			n++
			u := leseUmg(t)
			_, err := u.einEintrag(`{"Id":1,"` + propJeTyp(edm) + `":` + a.json + `}`)
			if ok := gueltig[edm] == a.name; ok && err != nil || !ok && fehlercode(err) != model.CodeE9001 {
				t.Errorf("%s aus %s: %v, gültig=%v", edm, a.name, err, ok)
			}
		}
	}
	if n != 40 {
		t.Errorf("%d Läufe, erwartet 40", n)
	}
}

// TestKontextNebenNichtGefunden hält SPEC-195 Nr. 2 und SPEC-109 Nr. 6 (V-7): ein Kontextfehler neben ErrNichtGefunden kehrt ohne
// Code zurück, die Coverage (S6) wird nicht `unknown`, der Lookup (S8) nicht `ODR-E6008`. Menge: Stelle 2 × Kontextfehler 2
// (Canceled, DeadlineExceeded) × Reihenfolge 2 = 8 Läufe.
func TestKontextNebenNichtGefunden(t *testing.T) {
	for _, s := range stellenLesen() {
		if s.name != "S6" && s.name != "S8" {
			continue
		}
		for _, ctxErr := range []error{context.Canceled, context.DeadlineExceeded} {
			for _, vorn := range []bool{false, true} {
				u := leseUmg(t)
				s.bau(u)
				err := errors.Join(outbound.ErrNichtGefunden, ctxErr)
				if vorn {
					err = errors.Join(ctxErr, outbound.ErrNichtGefunden)
				}
				u.p.setzeFehler(s.op, err)
				_, got := u.leseGET()
				if got == nil || fehlercode(got) != 0 || !errors.Is(got, ctxErr) {
					t.Errorf("%s %v vorn=%v: %v", s.name, ctxErr, vorn, got)
				}
			}
		}
	}
}

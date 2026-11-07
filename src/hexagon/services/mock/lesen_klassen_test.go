package mock_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// dreiPfade ruft Collection, Entity (Schlüssel 1) und Anzahl auf dem Set mit einem Eintrag (Coverage complete) und gibt f
// den Aufbau, den Namen des Pfads und den Fehler.
func dreiPfade(t *testing.T, set string, es model.EntityState, f func(u *umg, pfad string, err error)) {
	t.Helper()
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl} {
		u := leseUmg(t)
		u.co.setze(baselineID, set, model.CoverageComplete)
		u.en.setze(set, es)
		u.ressource(art, set, model.Optionen{}, zahl("", es.EntityKey))
		a, err := u.leseGET()
		if err == nil {
			lies(t, a)
		}
		f(u, fmt.Sprintf("Art %d", art), err)
	}
}

// TestTypenDesSets hält SPEC-196 Nr. 3 a (F-1): `@odata.type` muss ein nicht abstrakter Typ sein, der Typ des Sets oder über
// BaseType von ihm abgeleitet. Menge: sieben Typen (konkret 1, abgeleitet 2 davon eine Stufe tiefer, abstrakt 2 mit BaseType-
// Beziehung zum Set, fremd 1, unbekannt 1) × 3 Pfade = 21.
func TestTypenDesSets(t *testing.T) {
	faelle := []struct {
		typ string
		ok  bool
	}{
		{typProduct, true}, {typSpecial, true}, {"N.Tief", true}, {"N.Abstrakt", false}, {"N.SpezAbstrakt", false}, {"N.Fremd", false}, {"N.GibtEsNicht", false},
	}
	n := 0
	for _, f := range faelle {
		es := eintrag(setProducts, "1", `{"@odata.type":"#`+f.typ+`","Id":1}`, append(alleNamen(), "Extra")...)
		dreiPfade(t, setProducts, es, func(_ *umg, pfad string, err error) {
			n++
			if f.ok && err != nil || !f.ok && fehlercode(err) != model.CodeE9001 {
				t.Errorf("%s %s: %v, erwartet ok=%v", f.typ, pfad, err, f.ok)
			}
		})
	}
	if n != 21 {
		t.Errorf("%d Läufe, erwartet 21", n)
	}
}

// TestNichtUnterstuetzteMenge zählt die Menge von F-2: jeder Typ der Liste ist nicht unterstützt und keiner der acht.
func TestNichtUnterstuetzteMenge(t *testing.T) {
	liste := nichtUnterstuetzteTypen()
	for _, e := range liste {
		if (model.Property{EdmType: e}).Unterstuetzt() || slices.Contains(model.UnterstuetzteEDMTypen(), e) {
			t.Errorf("%s ist unterstützt", e)
		}
	}
	if len(liste) != 12 {
		t.Errorf("%d Typen, erwartet 12", len(liste))
	}
}

// TestNichtUnterstuetzteProperty hält SPEC-196 Nr. 3 b und d (F-2): eine Property im Payload, die der Typ kennt und nicht
// unterstützt, ist ODR-E9001. Menge: 12 nicht unterstützte Typen × 3 Pfade = 36.
func TestNichtUnterstuetzteProperty(t *testing.T) {
	n := 0
	for _, edm := range nichtUnterstuetzteTypen() {
		es := eintrag("Viele", "1", `{"Id":1,"`+vielProp(edm)+`":1}`, "Id", vielProp(edm))
		dreiPfade(t, "Viele", es, func(_ *umg, pfad string, err error) {
			n++
			if fehlercode(err) != model.CodeE9001 {
				t.Errorf("%s %s: %v", edm, pfad, err)
			}
		})
	}
	if n != 36 {
		t.Errorf("%d Läufe, erwartet 36", n)
	}
}

// TestLookupNichtUnterstuetztNichtAngefordert hält SPEC-196 Nr. 7 e (F-6): ohne `$select` und mit `*` zählen allein die
// unterstützten Properties als angefordert; der Eintrag kennt die nicht unterstützte Property nicht und ist vollständig.
// Menge: 12 nicht unterstützte Typen von Viele × {ohne, `*`} = 24, dazu Seltsam × 2.
func TestLookupNichtUnterstuetztNichtAngefordert(t *testing.T) {
	n := 0
	lauf := func(set, payloadName string) {
		for _, sel := range []*model.SelectOption{nil, {Alle: true}} {
			u := leseUmg(t)
			u.en.setze(set, eintrag(set, "1", `{"Id":1}`, "Id"))
			u.ressource(model.RessourceEntity, set, model.Optionen{Select: sel}, zahl("", "1"))
			n++
			if _, err := u.leseGET(); err != nil {
				t.Errorf("%s %s: %v", set, payloadName, err)
			}
		}
	}
	for _, edm := range nichtUnterstuetzteTypen() {
		lauf("Viele", edm)
	}
	lauf("Seltsam", "Edm.Binary")
	if n != 26 {
		t.Errorf("%d Läufe, erwartet 26", n)
	}
}

// TestTypNichtAufgeloest hält SPEC-196 Nr. 2 (F-3): ein Set, dessen Typ sich nicht auflösen lässt, ist ODR-E9001 vor Coverage
// und Entities. Menge: 3 Pfade.
func TestTypNichtAufgeloest(t *testing.T) {
	n := 0
	dreiPfade(t, "Waise", eintrag("Waise", "1", `{"Id":1}`, "Id"), func(u *umg, pfad string, err error) {
		n++
		fe := erwarteFehler(t, "Waise "+pfad, err, model.CodeE9001)
		if fe != nil && fe.Ursache != nil {
			t.Errorf("%s: Ursache %v", pfad, fe.Ursache)
		}
		for _, op := range []string{"Coverage.Get", "Entities.Get", "Entities.ListSet"} {
			if u.p.anzahl(op) != 0 {
				t.Errorf("%s: %s gerufen", pfad, op)
			}
		}
	})
	if n != 3 {
		t.Errorf("%d Läufe", n)
	}
}

// TestPortFehlerBlob hält SPEC-195 Nr. 11 S4 im Lesepfad (F-4, M18): Blobs.Open mit dem Gegenstand `Sonstiges`. Menge: Art 3 ×
// Fehler 6 × Form 2 = 36; Orakel von Hand (SPEC-109 Nr. 5).
func TestPortFehlerBlob(t *testing.T) {
	fehler := []error{outbound.ErrNichtGefunden, outbound.ErrBereitsVorhanden, outbound.ErrBeschaedigt, outbound.ErrBlobFehlt, model.ErrUngueltig, errors.New("roh " + kanarienwert)}
	soll := [6]model.Code{model.CodeE9001, model.CodeE9001, model.CodeE3001, model.CodeE3001, model.CodeE9001, model.CodeE9001}
	n := 0
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl} {
		for i, f := range fehler {
			for _, verpackt := range []bool{false, true} {
				n++
				u := leseUmg(t)
				u.standard(art, "1")
				err := f
				if verpackt {
					err = fmt.Errorf("blob: %w", f)
				}
				u.p.setzeFehler("Blobs.Open", err)
				_, got := u.leseGET()
				if fe := erwarteFehler(t, fmt.Sprintf("S4 Art %d %v", art, f), got, soll[i]); fe != nil && !errors.Is(fe, f) {
					t.Errorf("S4 Art %d %v: Ursache nicht erreichbar", art, f)
				}
				if u.p.anzahl("MetadatenCodec.ParseMetadata") != 0 {
					t.Errorf("S4: ParseMetadata gerufen")
				}
			}
		}
	}
	if n != 36 {
		t.Errorf("%d Läufe, erwartet 36", n)
	}
}

// TestSchluesselNameDoppelt hält SPEC-196 Nr. 7 a (F-5): jeder Name der Keys steht höchstens einmal; zwei Teile mit demselben
// Namen sind ODR-E6005, auch wenn ihre Zahl die der Keys ist. Menge: Typ T2 mit zwei Keys × Name 2 (I, S) = 2.
func TestSchluesselNameDoppelt(t *testing.T) {
	teile := [][]model.SchluesselTeil{
		{zahl("I", "1"), zahl("I", "2")},
		{{Name: "S", Literal: model.Literal{Art: model.LiteralZeichenkette, Text: "'a'"}}, {Name: "S", Literal: model.Literal{Art: model.LiteralZeichenkette, Text: "'b'"}}},
	}
	for _, tl := range teile {
		u := leseUmg(t)
		u.c.mc.modell = schluesselModell()
		u.ressource(model.RessourceEntity, "T2", model.Optionen{}, tl...)
		_, err := u.leseGET()
		_ = erwarteFehler(t, "Name doppelt "+tl[0].Name, err, model.CodeE6005)
		if u.p.anzahl("Entities.Get") != 0 {
			t.Error("Get gerufen")
		}
	}
}

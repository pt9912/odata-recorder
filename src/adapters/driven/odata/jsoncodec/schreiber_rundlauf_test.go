package jsoncodec_test

import (
	"context"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const grenzeLesen = 1 << 20

func formenZahl() map[string][]string {
	return map[string][]string{"Zahl": nil, "Zeichenkette": {"application/json;IEEE754Compatible=true"}}
}

// gelesen ist die Entity, die der Leser zu einer Entity mit dem Schlüssel S "k" und den Properties props liefern muss.
func gelesen(typ, schluessel string, props []model.PropertyWert) model.Entity {
	return model.Entity{Typ: typ, Schluessel: schluessel, Properties: props, Uebersprungen: []model.Uebersprungen{}}
}

// rundlauf schreibt die Entity als Entity und als Collection mit drei Elementen und liest beides mit dem Leser zurück.
func rundlauf(t *testing.T, was, set string, accept []string, e model.Entity, anzahl model.Anzahl, auswahl []string) {
	t.Helper()
	leser := jsoncodec.New()
	ea := model.EntityAntwort{Wurzel: wurzel, Set: set, Entity: e, Accept: accept, Auswahl: auswahl}
	body := entityBody(t, was, ea)
	got, err := leser.ParseEntity(context.Background(), strings.NewReader(body), testModell(), set, grenzeLesen)
	if err != nil || !reflect.DeepEqual(got, gelesenVon(e)) {
		t.Errorf("%s Entity: %v\n got %+v\nwant %+v\nBody %s", was, err, got, gelesenVon(e), body)
	}
	ca := model.CollectionAntwort{Wurzel: wurzel, Set: set, Entities: []model.Entity{e, e, e}, Accept: accept, Anzahl: anzahl, Auswahl: auswahl}
	body = collectionBody(t, was, ca)
	seite, err := leser.ParseCollection(context.Background(), strings.NewReader(body), testModell(), set, grenzeLesen)
	want := model.EntitySeite{Entities: []model.Entity{gelesenVon(e), gelesenVon(e), gelesenVon(e)}, Anzahl: anzahl}
	if !anzahl.Vorhanden {
		want.Anzahl = model.Anzahl{}
	}
	if err != nil || !reflect.DeepEqual(seite, want) {
		t.Errorf("%s Collection: %v\n got %+v\nwant %+v\nBody %s", was, err, seite, want, body)
	}
}

func gelesenVon(e model.Entity) model.Entity { return gelesen(e.Typ, e.Schluessel, e.Properties) }

// TestRundlauf hält Zeile 31: was der Schreiber der Stufe minimal schreibt, liest ODataCodec in dieselbe Entity.
func TestRundlauf(t *testing.T) {
	n := 0
	for _, m := range artMengen() {
		for _, g := range m.gueltig {
			for form, accept := range formenZahl() {
				e := model.Entity{Typ: "Ns.Alle", Schluessel: "'k'", Properties: []model.PropertyWert{pw("S", model.ArtString, "k"), pw("N"+m.prop, m.art, g.text)}}
				rundlauf(t, m.prop+" "+form, "Alles", accept, e, model.Anzahl{}, nil)
				n += 2
			}
		}
	}
	if n != 132 {
		t.Errorf("Wert-Zellen %d, will 132", n)
	}
	for _, art := range model.UnterstuetzteEDMTypen() {
		name := map[string]string{"Edm.String": "NS", "Edm.Boolean": "NB", "Edm.Int32": "NI32", "Edm.Int64": "NI64", "Edm.Decimal": "NDec", "Edm.Guid": "NG", "Edm.Date": "ND", "Edm.DateTimeOffset": "NT"}[art]
		e := model.Entity{Typ: "Ns.Alle", Schluessel: "'k'", Properties: []model.PropertyWert{pw("S", model.ArtString, "k"), pw(name, model.ArtNull, "")}}
		rundlauf(t, "null "+art, "Alles", nil, e, model.Anzahl{}, nil)
		n += 2
	}
	for _, typ := range []string{"Ns.Sonder", "Ns.Sonder2"} {
		e := model.Entity{Typ: typ, Schluessel: "1", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "A"), pw("Extra", model.ArtString, "e")}}
		rundlauf(t, typ, "Produkte", nil, e, model.Anzahl{}, nil)
		n += 2
	}
	leser := jsoncodec.New()
	for form, accept := range formenZahl() {
		c := gutCollection()
		c.Entities, c.Accept = nil, accept
		seite, err := leser.ParseCollection(context.Background(), strings.NewReader(collectionBody(t, "leer", c)), testModell(), "Produkte", grenzeLesen)
		if err != nil || len(seite.Entities) != 0 || seite.Entities == nil || seite.Anzahl.Vorhanden || seite.FolgeLink != (model.FolgeLink{}) {
			t.Errorf("leere Collection %s: %+v, %v", form, seite, err)
		}
		n++
		for _, w := range []int64{0, 1, math.MaxInt64} {
			c.Entities, c.Anzahl = []model.Entity{produkt("1", "A")}, model.Anzahl{Vorhanden: true, Wert: w}
			seite, err := leser.ParseCollection(context.Background(), strings.NewReader(collectionBody(t, "Anzahl", c)), testModell(), "Produkte", grenzeLesen)
			if err != nil || seite.Anzahl != c.Anzahl {
				t.Errorf("Anzahl %d %s: %+v, %v", w, form, seite.Anzahl, err)
			}
			n++
		}
	}
	if n != 132+16+4+2+6 {
		t.Errorf("Zellen %d, will 160", n)
	}
}

// TestRundlaufAuswahl hält Zeile 32: der Kontext der Auswahl ist für den Leser gültig.
func TestRundlaufAuswahl(t *testing.T) {
	n := 0
	for _, auswahl := range [][]string{{"Id", "Name"}, {"*"}, {"Name", "Name"}, {"Kategorie"}, nil} {
		for form, accept := range formenZahl() {
			if form == "Zeichenkette" {
				continue
			}
			rundlauf(t, "Auswahl "+strings.Join(auswahl, ","), "Produkte", accept, produkt("1", "A"), model.Anzahl{}, auswahl)
			n += 2
		}
	}
	if n != 10 {
		t.Errorf("Zellen %d, will 10", n)
	}
}

package jsoncodecv2_test

import (
	"context"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodecv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const grenzeLesen = 1 << 20

// gelesenVon ist die Entity, die der Leser zu e liefern muss (SPEC-172 Nr. 7): dieselben Felder und kein übersprungenes Mitglied.
func gelesenVon(e model.Entity) model.Entity {
	return model.Entity{Typ: e.Typ, Schluessel: e.Schluessel, Properties: e.Properties, Uebersprungen: []model.Uebersprungen{}}
}

// rundlauf schreibt die Entity als Entity und als Collection mit drei Elementen und liest beides mit dem Leser zurück.
func rundlauf(t *testing.T, was, set string, e model.Entity, anzahl model.Anzahl) {
	t.Helper()
	leser := jsoncodecv2.New()
	ea := model.EntityAntwort{Wurzel: wurzel, Set: set, Entity: e}
	body := entityBody(t, was, ea)
	got, err := leser.ParseEntity(context.Background(), strings.NewReader(body), testModell(), set, grenzeLesen)
	if err != nil || !reflect.DeepEqual(got, gelesenVon(e)) {
		t.Errorf("%s Entity: %v\n got %+v\nwant %+v\nBody %s", was, err, got, gelesenVon(e), body)
	}
	ca := model.CollectionAntwort{Wurzel: wurzel, Set: set, Entities: []model.Entity{e, e, e}, Anzahl: anzahl}
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

// alle ist eine Entity des Typs Ns.Alle mit dem Schlüssel S gleich `k` und den Properties props.
func alle(props ...model.PropertyWert) model.Entity {
	return model.Entity{Typ: "Ns.Alle", Schluessel: "'k'", Properties: append([]model.PropertyWert{pw("S", model.ArtString, "k")}, props...)}
}

// TestRundlauf hält Zeile 35: was der Schreiber schreibt, liest jsoncodecv2.New() mit demselben Modell in dieselbe Entity oder Collection.
func TestRundlauf(t *testing.T) {
	n := 0
	for _, m := range artMengen() {
		for _, g := range m.gueltig {
			rundlauf(t, m.prop+" "+g.text, "Alles", alle(pw("N"+m.prop, m.art, g.text)), model.Anzahl{})
			n += 2
		}
	}
	if n != 62 {
		t.Errorf("Wert-Zellen %d, will 62", n)
	}
	_, namen := siebenTypen()
	for _, name := range namen {
		rundlauf(t, "null N"+name, "Alles", alle(pw("N"+name, model.ArtNull, "")), model.Anzahl{})
		n += 2
	}
	for _, typ := range []string{"Ns.Sonder", "Ns.Sonder2"} {
		e := model.Entity{Typ: typ, Schluessel: "1", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "A"), pw("Extra", model.ArtString, "e")}}
		rundlauf(t, typ, "Produkte", e, model.Anzahl{})
		n += 2
	}
	schluessel := map[string]model.PropertyWert{
		"S": pw("K", model.ArtString, "o'k"), "B": pw("K", model.ArtBoolean, "true"), "I32": pw("K", model.ArtInt32, "42"), "I64": pw("K", model.ArtInt64, "-1"),
		"Dec": pw("K", model.ArtDecimal, "1.5"), "G": pw("K", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e"), "T": pw("K", model.ArtDateTimeOffset, "2020-01-01T00:00:00.5Z"),
	}
	literal := map[string]string{"S": "'o''k'", "B": "true", "I32": "42", "I64": "-1", "Dec": "1.5", "G": "0f8fad5b-d9cb-469f-a165-70867728950e", "T": "2020-01-01T00:00:00.5Z"}
	for _, name := range namen {
		e := model.Entity{Typ: "Ns.K" + name, Schluessel: literal[name], Properties: []model.PropertyWert{schluessel[name]}}
		rundlauf(t, "Schlüssel "+name, "K"+name, e, model.Anzahl{})
		n += 2
	}
	leser := jsoncodecv2.New()
	for _, anzahl := range []model.Anzahl{{}, {Vorhanden: true}, {Vorhanden: true, Wert: 17}, {Vorhanden: true, Wert: math.MaxInt64}} {
		c := gutCollection()
		c.Entities, c.Anzahl = nil, anzahl
		seite, err := leser.ParseCollection(context.Background(), strings.NewReader(collectionBody(t, "leer", c)), testModell(), "Produkte", grenzeLesen)
		if err != nil || len(seite.Entities) != 0 || seite.Entities == nil || seite.Anzahl != anzahl || seite.FolgeLink != (model.FolgeLink{}) {
			t.Errorf("leere Collection %+v: %+v, %v", anzahl, seite, err)
		}
		n++
	}
	texte := 0
	for c := range 128 {
		rundlauf(t, "Zeichen", "Alles", alle(pw("NS", model.ArtString, string(rune(c)))), model.Anzahl{})
		texte++
	}
	for _, r := range []rune{0xE9, 0x800, 0x10000, 0x2028, 0x2029, 0xFEFF, 0xFFFF, 0x10FFFF} {
		rundlauf(t, "Mehrbyte-Zeichen", "Alles", alle(pw("NS", model.ArtString, string(r))), model.Anzahl{})
		texte++
	}
	if texte != 136 {
		t.Errorf("Texte %d, will 136", texte)
	}
	seite, err := leser.ParseEntity(context.Background(), strings.NewReader(entityBody(t, "Marker", alleAntwort(pw("S", model.ArtString, "k"), pw("NS", model.ArtString, "[REDACTED]")))),
		testModell(), "Alles", grenzeLesen)
	if want := gelesenVon(alle(pw("NS", model.ArtString, "[REDACTED]"))); err != nil || !reflect.DeepEqual(seite, want) {
		t.Errorf("Text [REDACTED]: %+v, %v; will %+v", seite, err, want)
	}
	n += texte + 1
	if n != 235 {
		t.Errorf("Zellen %d, will 235", n)
	}
}

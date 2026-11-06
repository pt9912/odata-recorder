package jsoncodec_test

import (
	"context"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// mitFrist bricht den Test ab, wenn f nicht endet (eine Kette ohne Besuchszähler im Ring läuft endlos).
func mitFrist(t *testing.T, was string, f func()) {
	t.Helper()
	fertig := make(chan struct{})
	go func() { defer close(fertig); f() }()
	select {
	case <-fertig:
	case <-time.After(20 * time.Second):
		t.Fatalf("%s: keine Rückkehr", was)
	}
}

func kopie(e model.Entity) model.Entity {
	e.Properties = append([]model.PropertyWert{}, e.Properties...)
	e.Uebersprungen = append([]model.Uebersprungen{}, e.Uebersprungen...)
	return e
}

// stellen sind die vier Stellen, an denen eine Entity steht: SchreibeEntity und das erste, mittlere, letzte von drei Elementen.
func stellenNamen() []string { return []string{"Entity", "Element 1", "Element 2", "Element 3"} }

// ueberStellen setzt mut auf die Entity an jeder der vier Stellen; bei fehler muss der Aufruf scheitern und dst leer sein,
// sonst gelingt er und pruef bekommt den Body und die Stelle. Es liefert die Zahl der Zellen.
func ueberStellen(t *testing.T, was string, mut func(e *model.Entity), fehler bool, pruef func(stelle, body string)) int {
	t.Helper()
	for i, stelle := range stellenNamen() {
		name := was + " / " + stelle
		var body string
		if i == 0 {
			a := gutEntity()
			a.Entity = kopie(a.Entity)
			mut(&a.Entity)
			if fehler {
				entityScheitert(t, name, a)
				continue
			}
			body = entityBody(t, name, a)
		} else {
			a := gutCollection()
			a.Entities = append([]model.Entity{}, a.Entities...)
			a.Entities[i-1] = kopie(a.Entities[i-1])
			mut(&a.Entities[i-1])
			if fehler {
				collectionScheitert(t, name, a)
				continue
			}
			body = collectionBody(t, name, a)
		}
		if pruef != nil {
			pruef(stelle, body)
		}
	}
	return len(stellenNamen())
}

// TestEntityTyp hält Zeile 7: der Typ der Entity ist nicht abstrakt und der Typ des Sets oder von ihm abgeleitet.
func TestEntityTyp(t *testing.T) {
	verstoesse := []string{"", "Ns.Unbekannt", "Ns.Abstrakt", "Ns.Fremd", "Ns.R1", "Ns.R2a", "Ns.R3a", "Ns.R4a"}
	n := 0
	mitFrist(t, "Ringe", func() {
		for _, typ := range verstoesse {
			n += ueberStellen(t, "Typ "+typ, func(e *model.Entity) { e.Typ = typ }, true, nil)
		}
	})
	for _, typ := range []string{"Ns.Product", "Ns.Sonder", "Ns.Sonder2"} {
		n += ueberStellen(t, "Typ "+typ, func(e *model.Entity) { e.Typ = typ }, false, func(stelle, body string) {
			if got := strings.Contains(body, `"@odata.type":"#`+typ+`"`); got != (typ != "Ns.Product") {
				t.Errorf("Typ %s / %s: @odata.type %v in %s", typ, stelle, got, body)
			}
		})
	}
	if n != 44 {
		t.Errorf("Zellen %d, will 44 (8 Verstöße + 3 Gegenfälle, je 4 Stellen)", n)
	}
}

// TestSchluessel hält Zeile 8: nicht leer, gültiges UTF-8, ohne U+0000, ohne Längengrenze.
func TestSchluessel(t *testing.T) {
	n := 0
	for was, k := range map[string]string{
		"leer": "", "ungültiges UTF-8": "\xff", "NUL am Anfang": "\x001", "NUL in der Mitte": "1\x002", "NUL am Ende": "12\x00",
	} {
		n += ueberStellen(t, "Schlüssel "+was, func(e *model.Entity) { e.Schluessel = k }, true, nil)
	}
	for was, k := range map[string]string{
		"Anführungszeichen": "'O''Hara'", "jenseits von ASCII": "\xc3\xa9", "100000 Zeichen": strings.Repeat("k", 100000),
	} {
		n += ueberStellen(t, "Schlüssel "+was, func(e *model.Entity) { e.Schluessel = k }, false, func(stelle, body string) {
			if !strings.Contains(body, `"@odata.id":"`+wurzel+`/Produkte(`) {
				t.Errorf("Schlüssel %s / %s: ohne Adresse", was, stelle)
			}
		})
	}
	if n != 32 {
		t.Errorf("Zellen %d, will 32 (8 Formen, je 4 Stellen)", n)
	}
}

// TestPropertyNamen hält Zeile 9: Namen wirksamer Properties unterstützten Typs, Zeichen für Zeichen, höchstens einmal.
func TestPropertyNamen(t *testing.T) {
	id, name := pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "A")
	verstoesse := map[string][]model.PropertyWert{
		"unbekannt":                 {id, pw("Nope", model.ArtString, "x")},
		"Navigation-Property":       {id, pw("Kategorie", model.ArtString, "x")},
		"Typ nicht unterstützt":     {id, pw("Blob", model.ArtString, "x")},
		"leerer Name":               {id, pw("", model.ArtString, "x")},
		"andere Schreibweise":       {id, pw("name", model.ArtString, "A")},
		"zweimal, gleicher Wert":    {id, name, name},
		"zweimal, anderer Wert":     {id, name, pw("Name", model.ArtString, "B")},
		"zweimal, nicht benachbart": {name, id, name},
		"zweimal, erste und letzte": {id, name, id},
		"zweimal, gleiche Property": {id, id},
	}
	n := 0
	for was, props := range verstoesse {
		n += ueberStellen(t, was, func(e *model.Entity) { e.Properties = append([]model.PropertyWert{}, props...) }, true, nil)
	}
	for was, props := range map[string][]model.PropertyWert{"nil": nil, "leer": {}} {
		n += ueberStellen(t, was, func(e *model.Entity) { e.Properties = props }, false, func(stelle, body string) {
			k := "1"
			if stelle != "Entity" {
				k = stelle[len(stelle)-1:]
			}
			if !strings.Contains(body, `Produkte(`+k+`)"}`) {
				t.Errorf("Properties %s / %s: Entity ohne Properties erwartet: %s", was, stelle, body)
			}
		})
	}
	if n != 48 {
		t.Errorf("Zellen %d, will 48 (10 Verstöße + 2 Gegenfälle, je 4 Stellen)", n)
	}
}

// TestAuswahl hält Zeile 12: jedes Element ist `*` oder eine Property oder Navigation-Property des Typs des Sets, in beiden Stufen.
func TestAuswahl(t *testing.T) {
	ungueltig := map[string]string{"unbekannt": "Nope", "leer": "", "Schreibweise": "name", "nur abgeleiteter Typ": "Extra", "Leerzeichen": "Name "}
	n := 0
	for was, el := range ungueltig {
		for _, listenstelle := range []string{"erste", "letzte"} {
			liste := []string{el, "Name"}
			if listenstelle == "letzte" {
				liste = []string{"Name", el}
			}
			for _, accept := range [][]string{nil, {"application/json;odata.metadata=none"}} {
				e, c := gutEntity(), gutCollection()
				e.Auswahl, c.Auswahl, e.Accept, c.Accept = liste, liste, accept, accept
				name := "Auswahl " + was + " " + listenstelle + " " + strings.Join(accept, "")
				entityScheitert(t, name, e)
				collectionScheitert(t, name, c)
				n += 2
			}
		}
	}
	if n != 40 {
		t.Errorf("Zellen %d, will 40", n)
	}
	gueltig := []struct {
		auswahl []string
		kontext string
	}{
		{[]string{"*"}, "Produkte(*)"}, {[]string{"Kategorie"}, "Produkte(Kategorie)"}, {[]string{"Name", "Name"}, "Produkte(Name,Name)"},
		{nil, "Produkte"}, {[]string{}, "Produkte"}, {[]string{"Name"}, "Produkte(Name)"},
	}
	for _, g := range gueltig {
		for _, accept := range [][]string{nil, {"application/json;odata.metadata=none"}} {
			e, c := gutEntity(), gutCollection()
			e.Auswahl, c.Auswahl, e.Accept, c.Accept = g.auswahl, g.auswahl, accept, accept
			eb, cb := entityBody(t, "Auswahl gültig", e), collectionBody(t, "Auswahl gültig", c)
			keinKontext := accept != nil
			if got := strings.Contains(eb, `"@odata.context":"`+wurzel+`/$metadata#`+g.kontext+`/$entity"`); got == keinKontext {
				t.Errorf("Auswahl %v %v: Kontext der Entity: %s", g.auswahl, accept, eb)
			}
			if got := strings.HasPrefix(cb, `{"@odata.context":"`+wurzel+`/$metadata#`+g.kontext+`","value"`); got == keinKontext {
				t.Errorf("Auswahl %v %v: Kontext der Collection: %s", g.auswahl, accept, cb)
			}
		}
	}
}

// TestAnzahlGrenzen hält Zeile 13: Anzahl und n ab 0, Wert ohne Wirkung bei Vorhanden falsch.
func TestAnzahlGrenzen(t *testing.T) {
	n := 0
	for _, w := range []int64{-1, 0, 1, math.MaxInt64} {
		a := gutCollection()
		a.Anzahl = model.Anzahl{Vorhanden: true, Wert: w}
		n++
		if w < 0 {
			collectionScheitert(t, "Anzahl "+strconv.FormatInt(w, 10), a)
			continue
		}
		if body := collectionBody(t, "Anzahl", a); !strings.Contains(body, `"@odata.count":`+strconv.FormatInt(w, 10)+`,"value"`) {
			t.Errorf("Anzahl %d: %s", w, body)
		}
	}
	for _, w := range []int64{0, 5, -1} {
		a := gutCollection()
		a.Anzahl = model.Anzahl{Wert: w}
		n++
		if body := collectionBody(t, "Anzahl nicht vorhanden", a); strings.Contains(body, "@odata.count") {
			t.Errorf("Anzahl nicht vorhanden %d: %s", w, body)
		}
	}
	for _, w := range []int64{-1, math.MinInt64, 0, 1, math.MaxInt64} {
		n++
		z := neuesZiel()
		_, err := schreiber().SchreibeAnzahl(context.Background(), z, w)
		if w < 0 {
			istUngueltig(t, err, "n negativ")
			z.leer(t, "n negativ")
		} else if err != nil || z.body() != strconv.FormatInt(w, 10) {
			t.Errorf("n %d: %q, %v", w, z.body(), err)
		}
	}
	if n != 12 {
		t.Errorf("Menge %d, will 12", n)
	}
}

// TestNichtGenannt hält Zeile 28: eine Property, die Properties nicht nennt, fehlt; Uebersprungen wird nie geschrieben.
func TestNichtGenannt(t *testing.T) {
	adr := `"@odata.id":"` + wurzel + `/Produkte(1)"`
	faelle := map[string]struct {
		props  []model.PropertyWert
		uebers []model.Uebersprungen
		want   string // die Mitglieder hinter @odata.id
	}{
		"nullbare fehlt":       {[]model.PropertyWert{pw("Id", model.ArtInt32, "1")}, nil, `"Id":1`},
		"nicht nullbare fehlt": {[]model.PropertyWert{pw("Name", model.ArtString, "A")}, nil, `"Name":"A"`},
		"Grund typ":            {[]model.PropertyWert{pw("Id", model.ArtInt32, "1")}, []model.Uebersprungen{{Name: "Blob", Grund: model.UebersprungTyp}}, `"Id":1`},
		"Grund navigation":     {[]model.PropertyWert{pw("Id", model.ArtInt32, "1")}, []model.Uebersprungen{{Name: "Kategorie", Grund: model.UebersprungNavigation}}, `"Id":1`},
		"Grund offen":          {[]model.PropertyWert{pw("Id", model.ArtInt32, "1")}, []model.Uebersprungen{{Name: "Offen", Grund: model.UebersprungOffen}}, `"Id":1`},
	}
	n := 0
	for was, f := range faelle {
		e := gutEntity()
		e.Entity.Properties, e.Entity.Uebersprungen = f.props, f.uebers
		c := gutCollection()
		c.Entities = []model.Entity{e.Entity}
		wantE := `{"@odata.context":"` + wurzel + `/$metadata#Produkte/$entity",` + adr + `,` + f.want + `}`
		wantC := `{"@odata.context":"` + wurzel + `/$metadata#Produkte","value":[{` + adr + `,` + f.want + `}]}`
		if got := entityBody(t, was, e); got != wantE {
			t.Errorf("%s Entity: %s", was, got)
		}
		if got := collectionBody(t, was, c); got != wantC {
			t.Errorf("%s Collection: %s", was, got)
		}
		n += 2
	}
	if n != 10 {
		t.Errorf("Menge %d, will 10", n)
	}
}

// TestArgumenteUnveraendert hält Zeile 35: der Schreiber ändert kein Argument.
func TestArgumenteUnveraendert(t *testing.T) {
	e := gutEntity()
	e.Entity.Properties = []model.PropertyWert{pw("Name", model.ArtString, "A"), pw("Id", model.ArtInt32, "1")}
	e.Entity.Uebersprungen = []model.Uebersprungen{{Name: "Kategorie", Grund: model.UebersprungNavigation}, {Name: "Blob", Grund: model.UebersprungTyp}}
	e.Auswahl = []string{"Name", "Id", "*"}
	e.Accept = []string{"text/html", "application/json;odata.metadata=minimal"}
	c := gutCollection()
	c.Entities = []model.Entity{kopie(e.Entity), kopie(e.Entity)}
	c.Auswahl, c.Accept = []string{"Name", "Id", "*"}, []string{"text/html", "application/json;odata.metadata=minimal"}
	vorE := model.EntityAntwort{Wurzel: e.Wurzel, Set: e.Set, Entity: kopie(e.Entity), Auswahl: append([]string{}, e.Auswahl...), Accept: append([]string{}, e.Accept...)}
	vorC := model.CollectionAntwort{Wurzel: c.Wurzel, Set: c.Set, Entities: []model.Entity{kopie(c.Entities[0]), kopie(c.Entities[1])},
		Auswahl: append([]string{}, c.Auswahl...), Accept: append([]string{}, c.Accept...)}
	entityBody(t, "Entity", e)
	collectionBody(t, "Collection", c)
	for _, f := range []struct {
		feld      string
		vor, nach any
	}{
		{"Entity.Properties", vorE.Entity.Properties, e.Entity.Properties}, {"Entity.Uebersprungen", vorE.Entity.Uebersprungen, e.Entity.Uebersprungen},
		{"Entity.Auswahl", vorE.Auswahl, e.Auswahl}, {"Entity.Accept", vorE.Accept, e.Accept},
		{"Collection.Entities", vorC.Entities, c.Entities}, {"Collection.Auswahl", vorC.Auswahl, c.Auswahl},
		{"Collection.Accept", vorC.Accept, c.Accept}, {"Collection.Anzahl", vorC.Anzahl, c.Anzahl},
	} {
		if !reflect.DeepEqual(f.vor, f.nach) {
			t.Errorf("%s verändert: %v -> %v", f.feld, f.vor, f.nach)
		}
	}
}

var _ = io.EOF

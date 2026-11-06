package jsoncodec_test

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	ktxEntity = `"@odata.context":"` + wurzel + `/$metadata#Produkte`
	idPrefix  = `"@odata.id":"` + wurzel + `/Produkte(`
)

// TestEntityBytes hält Zeile 24: Bytes der Entity als Literal, Stufe × Fall.
func TestEntityBytes(t *testing.T) {
	sonder := gutEntity()
	sonder.Entity.Typ = "Ns.Sonder"
	sonder.Entity.Properties = []model.PropertyWert{pw("Extra", model.ArtString, "e"), pw("Id", model.ArtInt32, "1")}
	ohne := gutEntity()
	ohne.Entity.Properties = nil
	mitNull := gutEntity()
	mitNull.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtNull, "")}
	idName := gutEntity()
	idName.Auswahl = []string{"Id", "Name"}
	stern := gutEntity()
	stern.Auswahl = []string{"*"}
	none := []string{"application/json;odata.metadata=none"}
	faelle := []struct {
		was  string
		a    model.EntityAntwort
		want string
	}{
		{"ohne Auswahl", gutEntity(), `{` + ktxEntity + `/$entity",` + idPrefix + `1)","Id":1,"Name":"A"}`},
		{"Auswahl Id,Name", idName, `{` + ktxEntity + `(Id,Name)/$entity",` + idPrefix + `1)","Id":1,"Name":"A"}`},
		{"Auswahl *", stern, `{` + ktxEntity + `(*)/$entity",` + idPrefix + `1)","Id":1,"Name":"A"}`},
		{"abgeleiteter Typ", sonder, `{` + ktxEntity + `/$entity","@odata.type":"#Ns.Sonder",` + idPrefix + `1)","Id":1,"Extra":"e"}`},
		{"Properties nil", ohne, `{` + ktxEntity + `/$entity",` + idPrefix + `1)"}`},
		{"null-Wert", mitNull, `{` + ktxEntity + `/$entity",` + idPrefix + `1)","Id":1,"Name":null}`},
	}
	n := 0
	for _, f := range faelle {
		if got := entityBody(t, f.was, f.a); got != f.want {
			t.Errorf("minimal %s:\n got %s\nwant %s", f.was, got, f.want)
		}
		f.a.Accept = none
		wantNone := map[string]string{
			"ohne Auswahl": `{"Id":1,"Name":"A"}`, "Auswahl Id,Name": `{"Id":1,"Name":"A"}`, "Auswahl *": `{"Id":1,"Name":"A"}`,
			"abgeleiteter Typ": `{"Id":1,"Extra":"e"}`, "Properties nil": `{}`, "null-Wert": `{"Id":1,"Name":null}`,
		}[f.was]
		if got := entityBody(t, f.was, f.a); got != wantNone {
			t.Errorf("none %s: %s, will %s", f.was, got, wantNone)
		}
		n += 2
	}
	if n != 12 {
		t.Errorf("Menge %d, will 12", n)
	}
}

// TestCollectionBytes hält Zeile 25: value immer vorhanden, @odata.count vor value; Stufe × Entities × Anzahl.
func TestCollectionBytes(t *testing.T) {
	el := func(k, name string) string {
		return `{` + idPrefix + k + `)","Id":` + k + `,"Name":"` + name + `"}`
	}
	entitiesNach := map[string]struct {
		es   []model.Entity
		want string
	}{
		"nil":  {nil, `[]`},
		"leer": {[]model.Entity{}, `[]`},
		"eins": {[]model.Entity{produkt("1", "A")}, `[` + el("1", "A") + `]`},
		"drei": {[]model.Entity{produkt("1", "A"), produkt("2", "B"), produkt("3", "C")}, `[` + el("1", "A") + `,` + el("2", "B") + `,` + el("3", "C") + `]`},
	}
	anzahl := map[string]struct {
		a    model.Anzahl
		want string
	}{
		"fehlt": {model.Anzahl{Vorhanden: false, Wert: 5}, ``},
		"0":     {model.Anzahl{Vorhanden: true, Wert: 0}, `"@odata.count":0,`},
		"42":    {model.Anzahl{Vorhanden: true, Wert: 42}, `"@odata.count":42,`},
	}
	n := 0
	for _, stufe := range []string{"minimal", "none"} {
		for en, e := range entitiesNach {
			for an, a := range anzahl {
				c := model.CollectionAntwort{Wurzel: wurzel, Set: "Produkte", Entities: e.es, Anzahl: a.a}
				want := `{"@odata.context":"` + wurzel + `/$metadata#Produkte",` + a.want + `"value":` + e.want + `}`
				if stufe == "none" {
					c.Accept = []string{"application/json;odata.metadata=none"}
					want = noneErwartet(a.want, e.want)
				}
				if got := collectionBody(t, stufe+" "+en+" "+an, c); got != want {
					t.Errorf("%s Entities %s Anzahl %s:\n got %s\nwant %s", stufe, en, an, got, want)
				}
				n++
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, will 24", n)
	}
}

// noneErwartet ist die Collection der Stufe none als Literal: Elemente ohne @odata.id.
func noneErwartet(count, value string) string {
	value = strings.NewReplacer(
		`{`+idPrefix+`1)",`, `{`, `{`+idPrefix+`2)",`, `{`, `{`+idPrefix+`3)",`, `{`,
	).Replace(value)
	return `{` + count + `"value":` + value + `}`
}

// permutationen liefert alle Anordnungen von xs.
func permutationen(xs []model.PropertyWert) [][]model.PropertyWert {
	if len(xs) <= 1 {
		return [][]model.PropertyWert{append([]model.PropertyWert{}, xs...)}
	}
	var out [][]model.PropertyWert
	for i := range xs {
		rest := append(append([]model.PropertyWert{}, xs[:i]...), xs[i+1:]...)
		for _, p := range permutationen(rest) {
			out = append(out, append([]model.PropertyWert{xs[i]}, p...))
		}
	}
	return out
}

// TestPropertyReihenfolge hält Zeile 26: die Reihenfolge des Modells, unabhängig von der Übergabe; je 5 Aufrufe.
func TestPropertyReihenfolge(t *testing.T) {
	n := 0
	for _, typ := range []string{"Ns.Product", "Ns.Sonder"} {
		props := []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "A"), pw("Preis", model.ArtDecimal, "2.5")}
		want := `"Id":1,"Name":"A","Preis":2.5`
		if typ == "Ns.Sonder" {
			props = append(props, pw("Extra", model.ArtString, "e"))
			want += `,"Extra":"e"`
		} else {
			props = append(props, pw("Aktiv", model.ArtBoolean, "true"))
			want = `"Id":1,"Name":"A","Preis":2.5,"Aktiv":true`
		}
		perms := permutationen(props)
		if len(perms) != 24 {
			t.Fatalf("Permutationen %d", len(perms))
		}
		for _, p := range perms {
			a := gutEntity()
			a.Entity.Typ, a.Entity.Properties = typ, p
			for range 5 {
				if got := entityBody(t, typ, a); !strings.HasSuffix(got, want+"}") {
					t.Errorf("%s %v: %s, will Ende %s", typ, p, got, want)
				}
			}
			n++
		}
	}
	if n != 48 {
		t.Errorf("Menge %d, will 48", n)
	}
}

// TestFehlerBytes hält Zeile 29: code, message, target in dieser Reihenfolge, ohne Leerraum; ein Fall je Code.
func TestFehlerBytes(t *testing.T) {
	codes := model.AlleCodes()
	for _, c := range codes {
		f, _ := model.NeuesFehlerobjekt(c)
		for _, s := range []string{f.Error.Code, f.Error.Message, f.Error.Target} {
			if strings.ContainsAny(s, `"`+bs+"\x00\x1f") || strings.Contains(s, "\n") {
				t.Fatalf("%s: Text %q braucht ein Escape, das Literal des Tests nicht kennt", c, s)
			}
		}
		want := `{"error":{"code":"` + f.Error.Code + `","message":"` + f.Error.Message + `","target":"` + f.Error.Target + `"}}`
		z := neuesZiel()
		if _, err := sch.SchreibeFehler(context.Background(), z, f); err != nil || z.body() != want {
			t.Errorf("%s: %q, %v; will %q", c, z.body(), err, want)
		}
	}
	if len(codes) < 36 {
		t.Errorf("Menge %d, will mindestens 36", len(codes))
	}
}

// TestAnzahlBytes hält Zeile 30: die Ziffern ohne Zeilenende.
func TestAnzahlBytes(t *testing.T) {
	for n, want := range map[int64]string{0: "0", 1: "1", 9: "9", 10: "10", math.MaxInt64: "9223372036854775807"} {
		z := neuesZiel()
		if _, err := sch.SchreibeAnzahl(context.Background(), z, n); err != nil || z.body() != want {
			t.Errorf("n %d: %q, %v; will %q", n, z.body(), err, want)
		}
	}
}

// TestStufeNone hält Zeile 33: der Body der Stufe none hat keinen Kontext, keinen Typ, keine Id; der Leser lehnt ihn ab.
func TestStufeNone(t *testing.T) {
	none := []string{"application/json;odata.metadata=none"}
	leser := jsoncodec.New()
	sonder := gutEntity()
	sonder.Entity.Typ = "Ns.Sonder"
	// nur Properties des Typs des Sets: der Leser prüft die Mitglieder in der Reihenfolge des Dokuments, und eine
	// Property nur des abgeleiteten Typs wäre ihm ohne @odata.type `unbekannte_property`, bevor er den Kontext vermisst
	sonder.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "A")}
	sonder.Accept = none
	e := gutEntity()
	e.Accept = none
	c := gutCollection()
	c.Accept = none
	cz := gutCollection()
	cz.Accept, cz.Anzahl = none, model.Anzahl{Vorhanden: true, Wert: 3}
	for was, a := range map[string]model.EntityAntwort{"Entity": e, "abgeleiteter Typ": sonder} {
		body := entityBody(t, was, a)
		for _, verboten := range []string{"@odata.context", "@odata.type", "@odata.id"} {
			if strings.Contains(body, verboten) {
				t.Errorf("%s: %s im Body %s", was, verboten, body)
			}
		}
		_, err := leser.ParseEntity(context.Background(), strings.NewReader(body), modell, "Produkte", 1<<20)
		var df *model.DatenFehler
		if !errors.As(err, &df) || df.Ursache != model.DatenKontext || df.Ort != "/@odata.context" {
			t.Errorf("%s: Leser: %v", was, err)
		}
	}
	for was, a := range map[string]model.CollectionAntwort{"Collection ohne Anzahl": c, "Collection mit Anzahl": cz} {
		body := collectionBody(t, was, a)
		if strings.Contains(body, "@odata.context") || strings.Contains(body, "@odata.id") || strings.Contains(body, "@odata.type") {
			t.Errorf("%s: Annotation im Body %s", was, body)
		}
		if got := strings.Contains(body, `"@odata.count":3`); got != (a.Anzahl.Vorhanden) {
			t.Errorf("%s: @odata.count %v: %s", was, got, body)
		}
		_, err := leser.ParseCollection(context.Background(), strings.NewReader(body), modell, "Produkte", 1<<20)
		var df *model.DatenFehler
		if !errors.As(err, &df) || df.Ursache != model.DatenKontext || df.Ort != "/@odata.context" {
			t.Errorf("%s: Leser: %v", was, err)
		}
	}
}

// ergebnis ist die Ausgabe eines Aufrufs: Bytes, Darstellung, Fehlertext.
type ergebnis struct {
	body string
	d    model.Darstellung
	err  string
}

// aufruf ruft die Operation mit Index i für den Lauf g: die Argumente hängen von g ab, nie von einem anderen Lauf.
func aufruf(i, g int) ergebnis {
	z := neuesZiel()
	name := "n" + strconv.Itoa(g)
	var d model.Darstellung
	var err error
	switch i {
	case 0:
		a := gutEntity()
		a.Entity = produkt(strconv.Itoa(g+1), name)
		d, err = sch.SchreibeEntity(context.Background(), z, modell, a)
	case 1:
		a := gutCollection()
		a.Entities = []model.Entity{produkt(strconv.Itoa(g+1), name), produkt(strconv.Itoa(g+2), name)}
		a.Accept = []string{"application/json;IEEE754Compatible=" + strconv.FormatBool(g%2 == 0)}
		d, err = sch.SchreibeCollection(context.Background(), z, modell, a)
	case 2:
		d, err = sch.SchreibeAnzahl(context.Background(), z, int64(g))
	default:
		f, _ := model.NeuesFehlerobjekt(model.AlleCodes()[g%len(model.AlleCodes())])
		d, err = sch.SchreibeFehler(context.Background(), z, f)
	}
	r := ergebnis{body: z.body(), d: d}
	if err != nil {
		r.err = err.Error()
	}
	return r
}

// TestWiederholung hält Zeile 34: gleiche Argumente ergeben gleiche Bytes, geänderte geänderte; kein Cache.
func TestWiederholung(t *testing.T) {
	n := 0
	for i := range 4 {
		erst := aufruf(i, 3)
		if erst.err != "" || erst.body == "" {
			t.Fatalf("Operation %d: %+v", i, erst)
		}
		for k := range 100 {
			if r := aufruf(i, 3); r != erst {
				t.Fatalf("Operation %d, Aufruf %d: %+v, will %+v", i, k, r, erst)
			}
		}
		if r := aufruf(i, 4); r == erst {
			t.Errorf("Operation %d: geändertes Argument, gleiche Ausgabe %q", i, r.body)
		}
		n += 2
	}
	a := gutEntity()
	basis := entityBody(t, "Basis", a)
	for was, f := range map[string]func(a *model.EntityAntwort){
		"Wert": func(a *model.EntityAntwort) {
			a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "B")}
		},
		"Schluessel": func(a *model.EntityAntwort) { a.Entity.Schluessel = "2" },
		"Set":        func(a *model.EntityAntwort) { a.Set = "Produkte2" },
		"Accept":     func(a *model.EntityAntwort) { a.Accept = []string{"application/json;odata.metadata=none"} },
	} {
		g := gutEntity()
		f(&g)
		if got := entityBody(t, was, g); got == basis {
			t.Errorf("%s geändert, gleiche Ausgabe %s", was, got)
		}
		n++
	}
	if n != 12 {
		t.Errorf("Menge %d, will 12", n)
	}
}

// TestGleichzeitig hält Zeile 36 und 37 (Wirkungs-Fänger und Datenrennen): 16 Goroutinen × 4 Operationen × 50 Läufe, jede
// Ausgabe gleich der des einzelnen Aufrufs; der Detektor von make test-race meldet das Rennen.
func TestGleichzeitig(t *testing.T) {
	const gor, laeufe = 16, 50
	var erwartet [gor][4]ergebnis
	for g := range gor {
		for i := range 4 {
			erwartet[g][i] = aufruf(i, g)
		}
	}
	var wg sync.WaitGroup
	abweichungen := make([]int, gor)
	for g := range gor {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range laeufe {
				for i := range 4 {
					if aufruf(i, g) != erwartet[g][i] {
						abweichungen[g]++
					}
				}
			}
		}()
	}
	wg.Wait()
	for g, a := range abweichungen {
		if a != 0 {
			t.Errorf("Goroutine %d: %d abweichende Ausgaben von %d", g, a, laeufe*4)
		}
	}
}

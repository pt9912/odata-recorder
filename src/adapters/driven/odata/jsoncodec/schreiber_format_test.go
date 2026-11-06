package jsoncodec_test

import (
	"context"
	"errors"
	"io"
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
		if _, err := schreiber().SchreibeFehler(context.Background(), z, f); err != nil || z.body() != want {
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
		if _, err := schreiber().SchreibeAnzahl(context.Background(), z, n); err != nil || z.body() != want {
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
		_, err := leser.ParseEntity(context.Background(), strings.NewReader(body), testModell(), "Produkte", 1<<20)
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
		_, err := leser.ParseCollection(context.Background(), strings.NewReader(body), testModell(), "Produkte", 1<<20)
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
		d, err = schreiber().SchreibeEntity(context.Background(), z, testModell(), a)
	case 1:
		a := gutCollection()
		a.Entities = []model.Entity{produkt(strconv.Itoa(g+1), name), produkt(strconv.Itoa(g+2), name)}
		a.Accept = []string{"application/json;IEEE754Compatible=" + strconv.FormatBool(g%2 == 0)}
		d, err = schreiber().SchreibeCollection(context.Background(), z, testModell(), a)
	case 2:
		d, err = schreiber().SchreibeAnzahl(context.Background(), z, int64(g))
	case 3:
		f, _ := model.NeuesFehlerobjekt(model.AlleCodes()[g%len(model.AlleCodes())])
		d, err = schreiber().SchreibeFehler(context.Background(), z, f)
	case 4: // alle acht Typen und null, beide Zahlformen
		props := []model.PropertyWert{pw("S", model.ArtString, "k"+name), pw("B", model.ArtBoolean, "true"), pw("I32", model.ArtInt32, strconv.Itoa(g)),
			pw("I64", model.ArtInt64, "9223372036854775807"), pw("Dec", model.ArtDecimal, "-"+strconv.Itoa(g+1)+".5"),
			pw("G", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e"), pw("D", model.ArtDate, "2024-02-29"),
			pw("T", model.ArtDateTimeOffset, "2020-01-01T00:00:00.5Z"), pw("NS", model.ArtNull, "")}
		a := alleAntwort(props...)
		a.Accept = []string{"application/json;IEEE754Compatible=" + strconv.FormatBool(g%2 == 1)}
		d, err = schreiber().SchreibeEntity(context.Background(), z, testModell(), a)
	case 5: // abgeleitete Typen, Auswahl, Stufe none, Anzahl
		a := gutCollection()
		a.Entities = []model.Entity{{Typ: "Ns.Sonder", Schluessel: strconv.Itoa(g), Properties: []model.PropertyWert{pw("Extra", model.ArtString, name)}},
			{Typ: "Ns.Sonder2", Schluessel: "7", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "7")}}}
		a.Auswahl, a.Anzahl = []string{"Id", "*"}, model.Anzahl{Vorhanden: true, Wert: int64(g)}
		a.Accept = []string{"text/html", "application/json;odata.metadata=" + []string{"none", "minimal"}[g%2]}
		d, err = schreiber().SchreibeCollection(context.Background(), z, testModell(), a)
	case 6: // ungültige Argumente an jeder Operation
		a := gutEntity()
		a.Entity.Typ = "Ns.Fremd" + name
		_, e1 := schreiber().SchreibeEntity(context.Background(), z, testModell(), a)
		b := gutCollection()
		b.Wurzel = "ftp://" + name
		_, e2 := schreiber().SchreibeCollection(context.Background(), z, testModell(), b)
		_, e3 := schreiber().SchreibeAnzahl(context.Background(), z, -int64(g)-1)
		_, e4 := schreiber().SchreibeFehler(context.Background(), z, model.Fehlerobjekt{})
		_, e5 := schreiber().Adresse(testModell(), wurzel, "Produkte", "")
		fehler := []error{e1, e2, e3, e4, e5}
		var nilCtx context.Context
		var nilDst io.Writer
		_, e6 := schreiber().SchreibeAnzahl(nilCtx, z, 1)
		_, e7 := schreiber().SchreibeAnzahl(context.Background(), nilDst, 1)
		falsch, _ := model.NeuesFehlerobjekt(model.AlleCodes()[0])
		falsch.Error.Message += name
		_, e8 := schreiber().SchreibeFehler(context.Background(), z, falsch)
		_, e9 := schreiber().SchreibeFehler(context.Background(), z, model.Fehlerobjekt{Error: model.FehlerInhalt{Code: "ODR-E0000"}})
		fehler = append(fehler, e6, e7, e8, e9)
		// jede Zeile des Raster-Verstoßes: Version, Set, Typ, Name, Duplikat, Art, Text, Auswahl, Schlüssel, Anzahl, ungültiges UTF-8
		for k, mut := range []func(a *model.EntityAntwort){
			func(a *model.EntityAntwort) { a.Set = "Unbekannt" },
			func(a *model.EntityAntwort) { a.Set = "OhneTyp" },
			func(a *model.EntityAntwort) { a.Entity.Typ = "Ns.Abstrakt" },
			func(a *model.EntityAntwort) { a.Entity.Typ = "Ns.R2a" },
			func(a *model.EntityAntwort) {
				a.Entity.Properties = []model.PropertyWert{pw("Nope", model.ArtString, "x")}
			},
			func(a *model.EntityAntwort) {
				a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Id", model.ArtInt32, "1")}
			},
			func(a *model.EntityAntwort) { a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtNull, "")} },
			func(a *model.EntityAntwort) {
				a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtString, "1")}
			},
			func(a *model.EntityAntwort) {
				a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "01")}
			},
			func(a *model.EntityAntwort) {
				a.Entity.Properties = []model.PropertyWert{pw("Name", model.ArtString, "\xff")}
			},
			func(a *model.EntityAntwort) { a.Auswahl = []string{"Nope"} },
			func(a *model.EntityAntwort) { a.Entity.Schluessel = "" },
			func(a *model.EntityAntwort) { a.Wurzel = "" },
		} {
			x := gutEntity()
			mut(&x)
			_, e := schreiber().SchreibeEntity(context.Background(), z, testModell(), x)
			_, e2 := schreiber().SchreibeEntity(context.Background(), z, modellVersion("2.0"), x)
			fehler = append(fehler, e, e2)
			y := gutCollection()
			y.Entities = []model.Entity{x.Entity}
			y.Set, y.Wurzel, y.Auswahl = x.Set, x.Wurzel, x.Auswahl
			y.Anzahl = model.Anzahl{Vorhanden: true, Wert: int64(-k - 1)}
			_, e3 := schreiber().SchreibeCollection(context.Background(), z, testModell(), y)
			fehler = append(fehler, e3)
		}
		err = errors.Join(fehler...)
	case 7: // beendeter Kontext und Fehler von dst
		abgebrochen, cancel := context.WithCancel(context.Background())
		cancel()
		_, e1 := schreiber().SchreibeEntity(abgebrochen, z, testModell(), gutEntity())
		z2 := &ziel{log: &[]string{}, kurz: -2}
		_, e2 := schreiber().SchreibeCollection(context.Background(), z2, testModell(), gutCollection())
		z3 := &ziel{log: &[]string{}, kurz: -1, err: errors.New("dst " + name)}
		_, e3 := schreiber().SchreibeAnzahl(context.Background(), z3, 1)
		err = errors.Join(e1, e2, e3)
	case 8: // Adresse mit Schlüsseln, die kodiert werden
		var s string
		s, err = schreiber().Adresse(testModell(), wurzel, "Produkte", "'a b'/"+name+"+%")
		r := ergebnis{body: s}
		if err != nil {
			r.err = err.Error()
		}
		return r
	default: // Wurzeln und Typen im Ring
		a := gutEntity()
		a.Wurzel = "https://[::1]:" + strconv.Itoa(g+1) + "/a/%C3%A4"
		a.Entity.Typ = "Ns.R3a"
		_, e1 := schreiber().SchreibeEntity(context.Background(), z, testModell(), a)
		a.Entity.Typ = "Ns.Product"
		d, e2 := schreiber().SchreibeEntity(context.Background(), z, testModell(), a)
		err = errors.Join(e1, e2)
		return ergebnis{body: z.body(), d: d, err: fehlerText(err)}
	}
	return ergebnis{body: z.body(), d: d, err: fehlerText(err)}
}

func fehlerText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// varianten ist die Zahl der Aufrufe von aufruf.
const varianten = 10

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

// TestGleichzeitig hält Zeile 36 und 37 (Wirkungs-Fänger und Datenrennen): 16 Goroutinen × 10 Aufrufarten (vier Operationen und Adresse, Fehler, alle acht Typen) × 50 Läufe, jede
// Ausgabe gleich der des einzelnen Aufrufs; der Detektor von make test-race meldet das Rennen.
func TestGleichzeitig(t *testing.T) {
	const gor, laeufe = 16, 50
	var erwartet [gor][varianten]ergebnis
	for g := range gor {
		for i := range varianten {
			erwartet[g][i] = aufruf(i, g)
			if r := erwartet[g][i]; r.body == "" && r.err == "" {
				t.Fatalf("Variante %d, Lauf %d: weder Ausgabe noch Fehler", i, g)
			}
		}
	}
	var wg sync.WaitGroup
	abweichungen := make([]int, gor)
	for g := range gor {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range laeufe {
				for i := range varianten {
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
			t.Errorf("Goroutine %d: %d abweichende Ausgaben von %d", g, a, laeufe*varianten)
		}
	}
}

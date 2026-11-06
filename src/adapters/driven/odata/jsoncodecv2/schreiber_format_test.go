package jsoncodecv2_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodecv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const kopfProdukt = `{"d":{"__metadata":{"uri":"` + wurzel + `/Produkte(1)","type":"Ns.Product"}`

// TestEntityBytes hält Zeile 27: Bytes der Entity als Literal.
func TestEntityBytes(t *testing.T) {
	alle := alleAntwort(pw("S", model.ArtString, "s"), pw("B", model.ArtBoolean, "true"), pw("I32", model.ArtInt32, "1"), pw("I64", model.ArtInt64, "2"),
		pw("Dec", model.ArtDecimal, "3.5"), pw("G", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e"), pw("T", model.ArtDateTimeOffset, "1970-01-01T00:00:00Z"))
	sonder := gutEntity()
	sonder.Entity.Typ = "Ns.Sonder"
	sonder.Entity.Properties = []model.PropertyWert{pw("Extra", model.ArtString, "e"), pw("Id", model.ArtInt32, "1")}
	ohne, leer, mitNull, zwei := gutEntity(), gutEntity(), gutEntity(), gutEntity()
	ohne.Entity.Properties, leer.Entity.Properties = nil, []model.PropertyWert{}
	mitNull.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtNull, "")}
	zwei.Entity.Schluessel = "Id=1,Region='EU'"
	faelle := []struct {
		was  string
		m    model.ServiceModel
		a    model.EntityAntwort
		want string
	}{
		{"Typ des Sets", testModell(), gutEntity(), kopfProdukt + `,"Id":1,"Name":"A"}}`},
		{"abgeleiteter Typ", testModell(), sonder, `{"d":{"__metadata":{"uri":"` + wurzel + `/Produkte(1)","type":"Ns.Sonder"},"Id":1,"Extra":"e"}}`},
		{"Properties nil", testModell(), ohne, kopfProdukt + `}}`},
		{"Properties leer", testModell(), leer, kopfProdukt + `}}`},
		{"null-Wert", testModell(), mitNull, kopfProdukt + `,"Id":1,"Name":null}}`},
		{"alle sieben Typen", testModell(), alle, `{"d":{"__metadata":{"uri":"` + wurzel + `/Alles(k)","type":"Ns.Alle"},"S":"s","B":true,"I32":1,"I64":"2","Dec":"3.5",` +
			`"G":"0f8fad5b-d9cb-469f-a165-70867728950e","T":` + dt("0") + `}}`},
		{"Schlüssel aus zwei Properties", testModell(), zwei, `{"d":{"__metadata":{"uri":"` + wurzel + `/Produkte(Id=1,Region='EU')","type":"Ns.Product"},"Id":1,"Name":"A"}}`},
		{"Set jenseits von ASCII", setModell("Prod\xc3\xbckte"), model.EntityAntwort{Wurzel: wurzel, Set: "Prod\xc3\xbckte",
			Entity: model.Entity{Typ: "Ns.T", Schluessel: "1", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1")}}},
			`{"d":{"__metadata":{"uri":"` + wurzel + `/Prod%C3%BCkte(1)","type":"Ns.T"},"Id":1}}`},
	}
	for _, f := range faelle {
		z := neuesZiel()
		if _, err := schreiber().SchreibeEntity(context.Background(), z, f.m, f.a); err != nil || z.body() != f.want {
			t.Errorf("%s:\n got %s (%v)\nwant %s", f.was, z.body(), err, f.want)
		}
	}
	if len(faelle) != 8 {
		t.Errorf("Menge %d, will 8", len(faelle))
	}
}

// TestCollectionBytes hält Zeile 28: results immer vorhanden, __count als String vor results; Entities × Anzahl.
func TestCollectionBytes(t *testing.T) {
	el := func(k, name string) string {
		return `{"__metadata":{"uri":"` + wurzel + `/Produkte(` + k + `)","type":"Ns.Product"},"Id":` + k + `,"Name":"` + name + `"}`
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
		"fehlt":    {model.Anzahl{Vorhanden: false, Wert: 5}, ``},
		"0":        {model.Anzahl{Vorhanden: true, Wert: 0}, `"__count":"0",`},
		"42":       {model.Anzahl{Vorhanden: true, Wert: 42}, `"__count":"42",`},
		"MaxInt64": {model.Anzahl{Vorhanden: true, Wert: math.MaxInt64}, `"__count":"9223372036854775807",`},
	}
	n := 0
	for en, e := range entitiesNach {
		for an, a := range anzahl {
			c := model.CollectionAntwort{Wurzel: wurzel, Set: "Produkte", Entities: e.es, Anzahl: a.a}
			want := `{"d":{` + a.want + `"results":` + e.want + `}}`
			if got := collectionBody(t, en+" "+an, c); got != want {
				t.Errorf("Entities %s Anzahl %s:\n got %s\nwant %s", en, an, got, want)
			}
			n++
		}
	}
	if n != 16 {
		t.Errorf("Menge %d, will 16", n)
	}
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

// TestPropertyReihenfolge hält Zeile 29: die Reihenfolge des Modells, unabhängig von der Übergabe; je 5 Aufrufe.
func TestPropertyReihenfolge(t *testing.T) {
	n := 0
	for _, typ := range []string{"Ns.Product", "Ns.Sonder"} {
		props := []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "A"), pw("Preis", model.ArtDecimal, "2.5")}
		var want string
		if typ == "Ns.Sonder" {
			props = append(props, pw("Extra", model.ArtString, "e"))
			want = `"Id":1,"Name":"A","Preis":"2.5","Extra":"e"`
		} else {
			props = append(props, pw("Aktiv", model.ArtBoolean, "true"))
			want = `"Id":1,"Name":"A","Preis":"2.5","Aktiv":true`
		}
		perms := permutationen(props)
		if len(perms) != 24 {
			t.Fatalf("Permutationen %d", len(perms))
		}
		for _, p := range perms {
			a := gutEntity()
			a.Entity.Typ, a.Entity.Properties = typ, p
			for range 5 {
				if got := entityBody(t, typ, a); !strings.HasSuffix(got, "},"+want+"}}") {
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

// TestFehlerBytes hält Zeile 33: code und message als Objekt mit lang und value, ohne target, ohne Leerraum; ein Fall je Code.
func TestFehlerBytes(t *testing.T) {
	codes := model.AlleCodes()
	for _, c := range codes {
		f, _ := model.NeuesFehlerobjekt(c)
		for _, s := range []string{f.Error.Code, f.Error.Message} {
			if strings.ContainsAny(s, `"`+bs+"\x00\x1f") || strings.Contains(s, "\n") {
				t.Fatalf("%s: Text %q braucht ein Escape, das Literal des Tests nicht kennt", c, s)
			}
		}
		want := `{"error":{"code":"` + f.Error.Code + `","message":{"lang":"de","value":"` + f.Error.Message + `"}}}`
		z := neuesZiel()
		if _, err := schreiber().SchreibeFehler(context.Background(), z, f); err != nil || z.body() != want {
			t.Errorf("%s: %q, %v; will %q", c, z.body(), err, want)
		}
		var dek struct {
			Error struct {
				Code    string
				Message struct{ Lang, Value string }
				Target  *string
			}
		}
		if err := json.Unmarshal([]byte(z.body()), &dek); err != nil || dek.Error.Code != f.Error.Code || dek.Error.Message.Value != f.Error.Message ||
			dek.Error.Message.Lang != "de" || dek.Error.Target != nil {
			t.Errorf("%s: Dekodierung %+v, %v", c, dek, err)
		}
	}
	if len(codes) < 36 {
		t.Errorf("Menge %d, will mindestens 36", len(codes))
	}
}

// TestAnzahlBytes hält Zeile 34: die Ziffern ohne Zeilenende.
func TestAnzahlBytes(t *testing.T) {
	n := 0
	for v, want := range map[int64]string{0: "0", 1: "1", 9: "9", 10: "10", math.MaxInt64: "9223372036854775807"} {
		z := neuesZiel()
		if _, err := schreiber().SchreibeAnzahl(context.Background(), z, v); err != nil || z.body() != want {
			t.Errorf("n %d: %q, %v; will %q", v, z.body(), err, want)
		}
		n++
	}
	if n != 5 {
		t.Errorf("Menge %d, will 5", n)
	}
}

// TestWerteBytes hält Zeile 32: Bytes der Werte je Typ.
func TestWerteBytes(t *testing.T) {
	faelle := []struct {
		prop string
		art  model.WertArt
		text string
		want string
	}{
		{"S", model.ArtString, "a", `"a"`}, {"B", model.ArtBoolean, "true", `true`}, {"I32", model.ArtInt32, "1", `1`}, {"I64", model.ArtInt64, "1", `"1"`},
		{"Dec", model.ArtDecimal, "1.5", `"1.5"`}, {"G", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e", `"0f8fad5b-d9cb-469f-a165-70867728950e"`},
		{"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00Z", dt("1577836800000")},
		{"I32", model.ArtInt32, "-1", `-1`}, {"I64", model.ArtInt64, "-1", `"-1"`}, {"Dec", model.ArtDecimal, "-1", `"-1"`},
		{"I64", model.ArtInt64, "9223372036854775807", `"9223372036854775807"`}, {"I64", model.ArtInt64, "-9223372036854775808", `"-9223372036854775808"`},
		{"Dec", model.ArtDecimal, "0.1", `"0.1"`}, {"Dec", model.ArtDecimal, "123456789012345678901234567890.123456789", `"123456789012345678901234567890.123456789"`},
		{"NS", model.ArtNull, "", `null`}, {"NI32", model.ArtNull, "", `null`}, {"NDec", model.ArtNull, "", `null`},
	}
	for _, f := range faelle {
		if body := entityBody(t, f.prop, alleAntwort(pw(f.prop, f.art, f.text))); !strings.HasSuffix(body, `"`+f.prop+`":`+f.want+"}}") {
			t.Errorf("%s %q: %s, will Ende %s", f.prop, f.text, body, f.want)
		}
	}
	if len(faelle) != 17 {
		t.Errorf("Menge %d, will 17", len(faelle))
	}
}

// zeitFaelle sind die Zeitpunkte der Zeile 15: Text und, wo benannt, die Millisekunden als Literal; das Orakel der übrigen ist time.Parse.
func zeitFaelle() [][2]string {
	return [][2]string{
		{"1970-01-01T00:00:00Z", "0"}, {"1970-01-01T00:00:00.001Z", "1"}, {"1970-01-01T00:00:00.999Z", "999"}, {"1970-01-01T00:00:01Z", "1000"},
		{"1969-12-31T23:59:59.999Z", "-1"}, {"1969-12-31T23:59:59Z", "-1000"}, {"1969-12-31T23:59:59.001Z", "-999"},
		{"0001-01-01T00:00:00Z", "-62135596800000"}, {"9999-12-31T23:59:59.999Z", "253402300799999"}, {"2024-02-29T12:00:00Z", "1709208000000"},
		// je eine Zehnerpotenz des Betrags 4 bis 12 (in Millisekunden), beide Vorzeichen
		{"1970-01-01T00:00:10Z", ""}, {"1970-01-01T00:01:40Z", ""}, {"1970-01-01T00:16:40Z", ""}, {"1970-01-01T02:46:40Z", ""},
		{"1970-01-02T03:46:40Z", ""}, {"1970-01-12T13:46:40Z", ""}, {"1970-04-26T17:46:40Z", ""}, {"1973-03-03T09:46:40Z", ""}, {"2001-09-09T01:46:40Z", ""},
		{"1969-12-31T23:59:50Z", ""}, {"1969-12-31T23:58:20Z", ""}, {"1969-12-31T23:43:20Z", ""}, {"1969-12-31T21:13:20Z", ""},
		{"1969-12-30T20:13:20Z", ""}, {"1969-12-20T10:13:20Z", ""}, {"1969-09-07T06:13:20Z", ""}, {"1966-10-31T14:13:20Z", ""}, {"1938-04-24T22:13:20Z", ""},
	}
}

// TestZeitpunktBytes hält Zeile 15: `\/Date(<ms>)\/` mit den Millisekunden seit 1970, `-` bei negativem Wert, ohne führende Null.
func TestZeitpunktBytes(t *testing.T) {
	n := 0
	for _, f := range zeitFaelle() {
		parsed, err := time.Parse(time.RFC3339Nano, f[0])
		if err != nil {
			t.Fatalf("%s: %v", f[0], err)
		}
		ms := strconv.FormatInt(parsed.UnixMilli(), 10)
		if f[1] != "" && f[1] != ms {
			t.Fatalf("%s: Literal %s, time.Parse %s", f[0], f[1], ms)
		}
		want := bs + "/Date(" + ms + ")" + bs + "/"
		if body := entityBody(t, f[0], alleAntwort(pw("T", model.ArtDateTimeOffset, f[0]))); !strings.HasSuffix(body, `"T":"`+want+`"}}`) {
			t.Errorf("%s: %s, will %s", f[0], body, want)
		}
		n++
	}
	if n != 28 {
		t.Errorf("Menge %d, will 28", n)
	}
}

// ergebnis ist die Ausgabe eines Aufrufs: Bytes, Darstellung, Fehlertext.
type ergebnis struct {
	body string
	d    model.Darstellung
	err  string
}

func fehlerText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// aufruf ruft die Operation mit Index i für den Lauf g: die Argumente hängen von g ab, nie von einem anderen Lauf.
func aufruf(s *jsoncodecv2.Schreiber, i, g int) ergebnis {
	z := neuesZiel()
	name := "n" + strconv.Itoa(g)
	var d model.Darstellung
	var err error
	switch i {
	case 0:
		a := gutEntity()
		a.Entity = produkt(strconv.Itoa(g+1), name)
		d, err = s.SchreibeEntity(context.Background(), z, testModell(), a)
	case 1:
		a := gutCollection()
		a.Entities = []model.Entity{produkt(strconv.Itoa(g+1), name), produkt(strconv.Itoa(g+2), name)}
		a.Anzahl = model.Anzahl{Vorhanden: g%2 == 0, Wert: int64(g)}
		d, err = s.SchreibeCollection(context.Background(), z, testModell(), a)
	case 2:
		d, err = s.SchreibeAnzahl(context.Background(), z, int64(g))
	case 3:
		f, _ := model.NeuesFehlerobjekt(model.AlleCodes()[g%len(model.AlleCodes())])
		d, err = s.SchreibeFehler(context.Background(), z, f)
	case 4: // alle sieben Typen und null
		props := []model.PropertyWert{pw("S", model.ArtString, "k"+name+"\"\x5c\x01"+marker(g)), pw("NDec", model.ArtDecimal, strconv.Itoa(g)), pw("B", model.ArtBoolean, "true"),
			pw("I32", model.ArtInt32, strconv.Itoa(g)), pw("I64", model.ArtInt64, "9223372036854775807"), pw("Dec", model.ArtDecimal, "-"+strconv.Itoa(g+1)+".5"),
			pw("G", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e"), pw("T", model.ArtDateTimeOffset, "2020-01-01T00:00:00.5Z"), pw("NS", model.ArtNull, "")}
		d, err = s.SchreibeEntity(context.Background(), z, testModell(), alleAntwort(props...))
	case 5: // abgeleitete Typen, Auswahl, Anzahl
		a := gutCollection()
		a.Entities = []model.Entity{{Typ: "Ns.Sonder", Schluessel: strconv.Itoa(g), Properties: []model.PropertyWert{pw("Extra", model.ArtString, name)}},
			{Typ: "Ns.Sonder2", Schluessel: "7", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "7")}}}
		a.Auswahl, a.Anzahl = []string{"Id", "*"}, model.Anzahl{Vorhanden: true, Wert: int64(g)}
		a.Accept = []string{"text/html", "application/json;odata.metadata=" + []string{"none", "minimal"}[g%2]}
		d, err = s.SchreibeCollection(context.Background(), z, testModell(), a)
	case 6:
		err = errors.Join(ungueltigeArgumente(s, z, g, name)...)
	case 7: // beendeter Kontext und Fehler von dst
		abgebrochen, cancel := context.WithCancel(context.Background())
		cancel()
		_, e1 := s.SchreibeEntity(abgebrochen, z, testModell(), gutEntity())
		z2 := &ziel{log: &[]string{}, kurz: -2}
		_, e2 := s.SchreibeCollection(context.Background(), z2, testModell(), gutCollection())
		z3 := &ziel{log: &[]string{}, kurz: -1, err: errors.New("dst " + name)}
		_, e3 := s.SchreibeAnzahl(context.Background(), z3, 1)
		err = errors.Join(e1, e2, e3)
	case 8: // Adresse mit Schlüsseln, die kodiert werden
		adr, e := s.Adresse(testModell(), wurzel, "Produkte", "'a b'/"+name+"+%")
		return ergebnis{body: adr, err: fehlerText(e)}
	case 9: // Wurzeln und Typen im Ring
		a := gutEntity()
		a.Wurzel = "https://[::1]:" + strconv.Itoa(g+1) + "/a/%C3%A4"
		a.Entity.Typ = "Ns.R3a"
		_, e1 := s.SchreibeEntity(context.Background(), z, testModell(), a)
		a.Entity.Typ = "Ns.Product"
		d, e2 := s.SchreibeEntity(context.Background(), z, testModell(), a)
		return ergebnis{body: z.body(), d: d, err: fehlerText(errors.Join(e1, e2))}
	default: // die Zweige der Ablehnung in den Prüfungen der Wurzel, der Art und des Textes
		err = errors.Join(ablehnungen(s, z)...)
	}
	return ergebnis{body: z.body(), d: d, err: fehlerText(err)}
}

// marker liefert den Text des Markers, jeden zweiten Lauf.
func marker(g int) string {
	if g%2 == 0 {
		return ""
	}
	return "[REDACTED]"
}

// ungueltigeArgumente ruft jede Operation mit einem ungültigen Argument und liefert die Fehler.
func ungueltigeArgumente(s *jsoncodecv2.Schreiber, z *ziel, g int, name string) []error {
	ctx := context.Background()
	a := gutEntity()
	a.Entity.Typ = "Ns.Fremd" + name
	_, e1 := s.SchreibeEntity(ctx, z, testModell(), a)
	b := gutCollection()
	b.Wurzel = "ftp://" + name
	_, e2 := s.SchreibeCollection(ctx, z, testModell(), b)
	_, e3 := s.SchreibeAnzahl(ctx, z, -int64(g)-1)
	_, e4 := s.SchreibeFehler(ctx, z, model.Fehlerobjekt{})
	_, e5 := s.Adresse(testModell(), wurzel, "Produkte", "")
	fehler := []error{e1, e2, e3, e4, e5}
	var nilCtx context.Context
	var nilDst io.Writer
	_, e6 := s.SchreibeAnzahl(nilCtx, z, 1)
	_, e7 := s.SchreibeAnzahl(ctx, nilDst, 1)
	falsch, _ := model.NeuesFehlerobjekt(model.AlleCodes()[0])
	falsch.Error.Message += name
	_, e8 := s.SchreibeFehler(ctx, z, falsch)
	_, e9 := s.SchreibeFehler(ctx, z, model.Fehlerobjekt{Error: model.FehlerInhalt{Code: "ODR-E0000"}})
	fehler = append(fehler, e6, e7, e8, e9)
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
		func(a *model.EntityAntwort) {
			a.Entity.Properties = []model.PropertyWert{pw("__metadata", model.ArtString, "x")}
		},
		func(a *model.EntityAntwort) {
			a.Entity.Properties = []model.PropertyWert{pw("Tag", model.ArtDate, "2024-01-05")}
		},
		func(a *model.EntityAntwort) { a.Auswahl = []string{"Nope"} },
		func(a *model.EntityAntwort) { a.Entity.Schluessel = "" },
		func(a *model.EntityAntwort) { a.Wurzel = "" },
	} {
		x := gutEntity()
		mut(&x)
		_, e := s.SchreibeEntity(ctx, z, testModell(), x)
		_, e2 := s.SchreibeEntity(ctx, z, modellVersion("4.0"), x)
		fehler = append(fehler, e, e2)
		y := gutCollection()
		y.Entities = []model.Entity{x.Entity}
		y.Set, y.Wurzel, y.Auswahl = x.Set, x.Wurzel, x.Auswahl
		y.Anzahl = model.Anzahl{Vorhanden: true, Wert: int64(-k - 1)}
		_, e3 := s.SchreibeCollection(ctx, z, testModell(), y)
		fehler = append(fehler, e3)
	}
	// ein Name des Modells, der kein gültiges UTF-8 ist: der Fehler kommt aus dem Escape, nicht aus der Prüfung des Wertes
	for _, st := range stringStellen()[3:] {
		m, x, _ := st.bau("\xff", "")
		_, e := s.SchreibeEntity(ctx, z, m, x)
		fehler = append(fehler, e)
	}
	return fehler
}

// ablehnungen ruft den Schreiber mit je einem Wert, der einen Zweig der Ablehnung ausführt (nil-Ziele, Wurzel, Art, Text), und liefert die Fehler.
func ablehnungen(s *jsoncodecv2.Schreiber, z *ziel) []error {
	var nilCtx context.Context
	var nilDst io.Writer
	ctx := context.Background()
	var fehler []error
	sammle := func(_ model.Darstellung, e error) { fehler = append(fehler, e) }
	sammle(s.SchreibeEntity(nilCtx, z, testModell(), gutEntity()))
	sammle(s.SchreibeEntity(ctx, nilDst, testModell(), gutEntity()))
	sammle(s.SchreibeCollection(ctx, nilDst, testModell(), gutCollection()))
	sammle(s.SchreibeFehler(ctx, nilDst, fehlerCode()))
	_, e := s.Adresse(testModell(), wurzel, "Unbekannt", "1")
	fehler = append(fehler, e)
	bad := gutCollection()
	bad.Entities = []model.Entity{produkt("1", "A"), {Typ: "Ns.Unbekannt", Schluessel: "2"}}
	sammle(s.SchreibeCollection(ctx, z, testModell(), bad))
	for _, w := range []string{
		"http://" + strings.Repeat("a", 2048), "http://", "http://h:", "http://h:80x", "http://h//a", "http://h/a b", "http://[]", "http://[g]",
		"http://h:123456", "http://[::1", "http://h/%zz", "http://h/..", "https://h", "http://h:80/a",
	} {
		_, e := s.Adresse(testModell(), w, "Produkte", "1")
		fehler = append(fehler, e)
	}
	for _, f := range []struct {
		prop string
		art  model.WertArt
		text string
	}{
		{"NS", "unbekannt", "x"}, {"I32", model.ArtInt32, "-"}, {"I32", model.ArtInt32, "x"}, {"I32", model.ArtInt32, "-5"}, {"I32", model.ArtInt32, "5x"},
		{"I32", model.ArtInt32, "99999999999"}, {"I64", model.ArtInt64, "-0"}, {"I64", model.ArtInt64, "01"},
		{"Dec", model.ArtDecimal, strings.Repeat("1", 1025)}, {"Dec", model.ArtDecimal, "x"}, {"Dec", model.ArtDecimal, "5x"}, {"Dec", model.ArtDecimal, "1.50"},
		{"Dec", model.ArtDecimal, "-0"}, {"Dec", model.ArtDecimal, "1.5.5"}, {"Dec", model.ArtDecimal, "-0.5"}, {"Dec", model.ArtDecimal, "1."}, {"Dec", model.ArtDecimal, "1.5"},
		{"G", model.ArtGUID, "0f8fad5b"}, {"G", model.ArtGUID, "0f8fad5bxd9cb-469f-a165-70867728950e"}, {"G", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950g"},
		{"B", model.ArtBoolean, "True"}, {"B", model.ArtBoolean, "false"},
		{"T", model.ArtDateTimeOffset, "2020"}, {"T", model.ArtDateTimeOffset, "2023-02-29T00:00:00Z"}, {"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00Z"},
		{"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00.1234Z"}, {"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00.120Z"}, {"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00,1Z"},
		{"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00.12aZ"}, {"T", model.ArtDateTimeOffset, "2020-13-01T00:00:00Z"}, {"T", model.ArtDateTimeOffset, "2020-01-01T24:00:00Z"},
		{"T", model.ArtDateTimeOffset, "2020-01-01T00:60:00Z"}, {"T", model.ArtDateTimeOffset, "2020-01-01T00:00:60Z"}, {"T", model.ArtDateTimeOffset, "0000-01-01T00:00:00Z"},
		{"T", model.ArtDateTimeOffset, "2020-0a-01T00:00:00Z"}, {"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00.1Z"}, {"T", model.ArtDateTimeOffset, "2020-01-01T00:00:00.12Z"},
		{"D", model.ArtDate, "2024-01-05"}, {"ND", model.ArtNull, ""}, {"S", model.ArtString, "[REDACTED]"},
	} {
		sammle(s.SchreibeEntity(ctx, z, testModell(), alleAntwort(pw(f.prop, f.art, f.text))))
	}
	return fehler
}

// varianten ist die Zahl der Aufrufe von aufruf.
const varianten = 11

// TestWiederholung hält Zeile 36: gleiche Argumente ergeben gleiche Bytes, geänderte geänderte; kein Cache.
func TestWiederholung(t *testing.T) {
	n := 0
	for i := range 4 {
		erst := aufruf(schreiber(), i, 3)
		if erst.err != "" || erst.body == "" {
			t.Fatalf("Operation %d: %+v", i, erst)
		}
		for k := range 100 {
			if r := aufruf(schreiber(), i, 3); r != erst {
				t.Fatalf("Operation %d, Aufruf %d: %+v, will %+v", i, k, r, erst)
			}
		}
		if r := aufruf(schreiber(), i, 4); r == erst {
			t.Errorf("Operation %d: geändertes Argument, gleiche Ausgabe %q", i, r.body)
		}
		n += 2
	}
	basis := entityBody(t, "Basis", gutEntity())
	for was, f := range map[string]func(a *model.EntityAntwort){
		"Wert": func(a *model.EntityAntwort) {
			a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "B")}
		},
		"Schluessel": func(a *model.EntityAntwort) { a.Entity.Schluessel = "2" },
		"Set":        func(a *model.EntityAntwort) { a.Set = "Produkte2" },
		"Wurzel":     func(a *model.EntityAntwort) { a.Wurzel = "http://anders" },
	} {
		g := gutEntity()
		f(&g)
		if got := entityBody(t, was, g); got == basis {
			t.Errorf("%s geändert, gleiche Ausgabe %s", was, got)
		}
		n++
	}
	g := gutEntity()
	g.Accept = []string{"application/json;odata.metadata=none"}
	if got := entityBody(t, "Accept", g); got != basis {
		t.Errorf("Accept geändert, andere Ausgabe %s", got)
	}
	n++
	if n != 13 {
		t.Errorf("Menge %d, will 13", n)
	}
}

// TestGleichzeitig hält Zeile 37 und 38: 16 Goroutinen rufen nach gemeinsamem Start einen gemeinsamen Schreiber mit 11 Aufrufarten in 50 Läufen;
// jede Ausgabe gleicht der des einzelnen Aufrufs, den der Test danach berechnet. Ein Datenrennen meldet allein der Detektor von make test-race,
// und nur in einem Zweig, den diese Aufrufe ausführen.
func TestGleichzeitig(t *testing.T) {
	const gor, laeufe = 16, 50
	geteilt := jsoncodecv2.NewSchreiber()
	start := make(chan struct{})
	gelaufen := make([][]ergebnis, gor)
	var wg sync.WaitGroup
	for g := range gor {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = jsoncodecv2.NewSchreiber()
			for range laeufe {
				for i := range varianten {
					gelaufen[g] = append(gelaufen[g], aufruf(geteilt, i, g))
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	for g := range gor {
		for i := range varianten {
			erwartet := aufruf(schreiber(), i, g)
			if erwartet.body == "" && erwartet.err == "" {
				t.Fatalf("Aufrufart %d, Goroutine %d: weder Ausgabe noch Fehler", i, g)
			}
			abweichend := 0
			for k := range laeufe {
				if gelaufen[g][k*varianten+i] != erwartet {
					abweichend++
				}
			}
			if abweichend != 0 {
				t.Errorf("Aufrufart %d, Goroutine %d: %d abweichende Ausgaben von %d", i, g, abweichend, laeufe)
			}
		}
	}
}

// TestSchreiberOhneZustand hält Zeile 39: der Schreiber hat kein Feld.
func TestSchreiberOhneZustand(t *testing.T) {
	if n := reflect.TypeFor[jsoncodecv2.Schreiber]().NumField(); n != 0 {
		t.Errorf("Schreiber hat %d Felder", n)
	}
}

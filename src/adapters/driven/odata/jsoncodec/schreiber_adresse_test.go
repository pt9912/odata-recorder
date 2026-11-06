package jsoncodec_test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// unveraendert sind die Bytes, die die Adresse nicht kodiert (SPEC-169 Nr. 2): die Tabelle des Orakels.
const unveraendert = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~',="

// kodiertErwartet ist das Orakel der Kodierung: jedes Byte außerhalb der Tabelle als `%` und zwei Hex-Ziffern in Großbuchstaben.
func kodiertErwartet(s string) string {
	var b strings.Builder
	for i := range len(s) {
		if strings.IndexByte(unveraendert, s[i]) >= 0 {
			b.WriteByte(s[i])
		} else {
			fmt.Fprintf(&b, "%%%02X", s[i])
		}
	}
	return b.String()
}

// setModell hat ein Set mit dem Namen set, dessen Typ die Property Id hat.
func setModell(set string) model.ServiceModel {
	typ := model.EntityType{Name: "Ns.T", Keys: []string{"Id"}, Properties: []model.Property{prop("Id", "Edm.Int32", false)}}
	return model.NeuesServiceModel("4.0", []string{"Ns"}, []model.EntityType{typ}, []model.EntitySet{{Name: set, EntityType: "Ns.T"}}, nil)
}

// TestAdresseKodierung hält Zeile 19: 130 Zeichen an zwei Stellen (Schlüssel, Set), Dekodierung mit der Standardbibliothek.
func TestAdresseKodierung(t *testing.T) {
	var zeichen []string
	for c := 1; c <= 0x7f; c++ {
		zeichen = append(zeichen, string(rune(c)))
	}
	zeichen = append(zeichen, "\xc3\xa9", "\xe2\x82\xac", "\xf0\x90\x80\x80")
	n := 0
	for _, z := range zeichen {
		for _, stelle := range []string{"Schlüssel", "Set"} {
			set, key := "Produkte", "k"+z
			m := testModell()
			if stelle == "Set" {
				set, key = "S"+z, "1"
				m = setModell(set)
			}
			want := wurzel + "/" + kodiertErwartet(set) + "(" + kodiertErwartet(key) + ")"
			got, err := schreiber().Adresse(m, wurzel, set, key)
			if err != nil || got != want {
				t.Errorf("%s %q: %q, %v; will %q", stelle, z, got, err, want)
			}
			if dek, err := url.PathUnescape(got); err != nil || dek != wurzel+"/"+set+"("+key+")" {
				t.Errorf("%s %q: Dekodierung %q, %v", stelle, z, dek, err)
			}
			n++
		}
	}
	if len(zeichen) != 130 || n != 260 {
		t.Errorf("Zeichen %d, Zellen %d; will 130, 260", len(zeichen), n)
	}
}

// idVon liest den Wert von `@odata.id` des ersten Vorkommens ab pos.
func idVon(body string, pos int) (string, int) {
	const k = `"@odata.id":"`
	i := strings.Index(body[pos:], k)
	if i < 0 {
		return "", -1
	}
	start := pos + i + len(k)
	ende := strings.IndexByte(body[start:], '"')
	return body[start : start+ende], start + ende
}

// TestAdresseGleich hält Zeile 20: Adresse, @odata.id der Entity und @odata.id des Elements sind dieselbe Zeichenfolge.
func TestAdresseGleich(t *testing.T) {
	schluessel := []string{"42", "'O''Hara'", "Id=1,Region='EU'", "\xc3\xa4", "a b", "a/b", "a?b", "a#b", "a%b", "a+b"}
	n := 0
	pruefe := func(m model.ServiceModel, set, key, kontextSet string) {
		adr, err := schreiber().Adresse(m, wurzel, set, key)
		if err != nil {
			t.Fatalf("%q: %v", key, err)
		}
		if want := wurzel + "/" + kodiertErwartet(set) + "(" + kodiertErwartet(key) + ")"; adr != want {
			t.Errorf("Adresse %q: %q, will %q", key, adr, want)
		}
		typ := "Ns.Product"
		if set != "Produkte" {
			typ = "Ns.T"
		}
		e := model.Entity{Typ: typ, Schluessel: key, Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1")}}
		zl := neuesZiel()
		if _, err := schreiber().SchreibeEntity(context.Background(), zl, m, model.EntityAntwort{Wurzel: wurzel, Set: set, Entity: e}); err != nil {
			t.Fatal(err)
		}
		if id, _ := idVon(zl.body(), 0); id != adr {
			t.Errorf("Entity %q: @odata.id %q, will %q", key, id, adr)
		}
		if !strings.Contains(zl.body(), `"@odata.context":"`+wurzel+`/$metadata#`+kontextSet+`/$entity"`) {
			t.Errorf("Kontext trägt das Set nicht roh: %s", zl.body())
		}
		zc := neuesZiel()
		if _, err := schreiber().SchreibeCollection(context.Background(), zc, m, model.CollectionAntwort{Wurzel: wurzel, Set: set, Entities: []model.Entity{e}}); err != nil {
			t.Fatal(err)
		}
		if id, _ := idVon(zc.body(), 0); id != adr {
			t.Errorf("Element %q: @odata.id %q, will %q", key, id, adr)
		}
		if !strings.Contains(zc.body(), `"@odata.context":"`+wurzel+`/$metadata#`+kontextSet+`","value"`) {
			t.Errorf("Kontext der Collection trägt das Set nicht roh: %s", zc.body())
		}
		n += 3
	}
	for _, k := range schluessel {
		pruefe(testModell(), "Produkte", k, "Produkte")
	}
	pruefe(setModell("Prod\xc3\xbckte"), "Prod\xc3\xbckte", "1", "Prod\xc3\xbckte")
	if n != 33 {
		t.Errorf("Zellen %d, will 33", n)
	}
}

// TestWurzel hält Zeile 21: die Form von Wurzel an Entity, Collection und Adresse.
func TestWurzel(t *testing.T) {
	gueltig := []string{
		"http://h", "https://h", "http://h:8080", "http://[::1]:8080", "http://h/a", "http://h/a/b", "http://h/%C3%A4", "http://127.0.0.1:1",
		"http://h/a.b", "http://h/...", "http://h/" + strings.Repeat("a", 2048-len("http://h/")),
		// gültige Seite der Zeichen und Ziffern: Host mit - _ ~ und Großbuchstaben, Port mit fünf Ziffern, Port mit Pfad, IPv6 mit Punkt und Großbuchstaben,
		// Hex-Kleinbuchstaben und Steuerzeichen im %-Teil, %2F im Segment
		"http://a-b", "http://a_b", "http://a~b", "http://A.B", "http://h:65535", "http://h:80/a/b", "http://[::FFFF:1.2.3.4]:80", "http://h/%c3%a4",
		"http://h/a-b/c_d/e~f/G", "http://h/%00", "http://h/%1F", "http://h/%2F",
	}
	ungueltig := []string{
		"", "ftp://h", "HTTP://h", "Https://h", "//h", "h", "http://", "http:///a", "http://h/", "http://h//a", "http://h/a/", "http://u@h", "http://u:p@h",
		"http://h?x=1", "http://h#f", "http://h/a?b", "http://h/a#b", "http://h/a" + bs + "b", "http://h/a b", " http://h", "http://h/\xc3\xa4", "http://h\xc3\xa4",
		"http://h/%", "http://h/%1", "http://h/%GG", "http://h/.", "http://h/..", "http://h/a/./b", "http://h:", "http://h:123456", "http://h:80a", "http://h:80:90",
		"http://[::1", "http://[]", "http://[g::1]", "http://h%41", "http://h\n", "http://h\r\nX: y", "http://h/a\x00", "http://h/a\x7f",
		"http://h/" + strings.Repeat("a", 2049-len("http://h/")), "http://h/%2E", "http://h/%2e%2e", "http://h/.%2E",
	}
	if len(gueltig) != 23 || len(ungueltig) != 44 {
		t.Fatalf("Menge gültig %d, ungültig %d; will 23, 44", len(gueltig), len(ungueltig))
	}
	n := 0
	for _, w := range gueltig {
		e, c := gutEntity(), gutCollection()
		e.Wurzel, c.Wurzel = w, w
		if body := entityBody(t, w, e); !strings.Contains(body, `"@odata.id":"`+w+`/Produkte(1)"`) {
			t.Errorf("Entity %q: %s", w, body)
		}
		collectionBody(t, w, c)
		if a, err := schreiber().Adresse(testModell(), w, "Produkte", "1"); err != nil || a != w+"/Produkte(1)" {
			t.Errorf("Adresse %q: %q, %v", w, a, err)
		}
		n += 3
	}
	for _, w := range ungueltig {
		e, c := gutEntity(), gutCollection()
		e.Wurzel, c.Wurzel = w, w
		entityScheitert(t, fmt.Sprintf("Entity %q", w), e)
		collectionScheitert(t, fmt.Sprintf("Collection %q", w), c)
		leer := model.CollectionAntwort{Wurzel: w, Set: "Produkte"}
		collectionScheitert(t, fmt.Sprintf("Collection ohne Entities %q", w), leer)
		a, err := schreiber().Adresse(testModell(), w, "Produkte", "1")
		if a != "" {
			t.Errorf("Adresse %q: %q", w, a)
		}
		istUngueltig(t, err, fmt.Sprintf("Adresse %q", w))
		n += 3
	}
	if n != 201 {
		t.Errorf("Zellen %d, will 201", n)
	}
}

// aushandlung ruft Entity und Collection mit accept und prüft Content-Type, Kontext und Zahlform gegen die Erwartung.
func aushandlung(t *testing.T, was string, accept []string, none, ieee bool) {
	t.Helper()
	stufe := "minimal"
	if none {
		stufe = "none"
	}
	want := "application/json;odata.metadata=" + stufe
	if ieee {
		want += ";IEEE754Compatible=true"
	}
	want += ";charset=utf-8"
	c := gutCollection()
	c.Accept, c.Anzahl = accept, model.Anzahl{Vorhanden: true, Wert: 3}
	e := alleAntwort(pw("I64", model.ArtInt64, "7"))
	e.Accept = accept
	for _, op := range []struct {
		name string
		f    func(z io.Writer) (model.Darstellung, error)
	}{
		{"Collection", func(z io.Writer) (model.Darstellung, error) {
			return schreiber().SchreibeCollection(context.Background(), z, testModell(), c)
		}},
		{"Entity", func(z io.Writer) (model.Darstellung, error) {
			return schreiber().SchreibeEntity(context.Background(), z, testModell(), e)
		}},
	} {
		z := neuesZiel()
		d, err := op.f(z)
		if err != nil || d.ContentType != want {
			t.Errorf("%s %s %q: Content-Type %q (%v), will %q", was, op.name, accept, d.ContentType, err, want)
		}
		body := z.body()
		if got := strings.Contains(body, "@odata.context"); got == none {
			t.Errorf("%s %s %q: Kontext %v bei none=%v: %s", was, op.name, accept, got, none, body)
		}
		zahl := `"I64":7`
		if op.name == "Collection" {
			zahl = `"@odata.count":3`
		}
		if ieee {
			zahl = strings.Replace(zahl, ":7", `:"7"`, 1)
			zahl = strings.Replace(zahl, ":3", `:"3"`, 1)
		}
		if !strings.Contains(body, zahl) {
			t.Errorf("%s %s %q: Zahlform %s fehlt in %s", was, op.name, accept, zahl, body)
		}
	}
}

// TestAushandlung hält Zeile 22: Stufe und Zahlform aus Accept, nie ein Fehler; Schleife über die Achsen.
func TestAushandlung(t *testing.T) {
	n := 0
	metadata := []string{"none", "NONE", "None", "minimal", "full", "foo", "", "<fehlt>"}
	ieee := []string{"true", "TRUE", "True", "false", "1", `"true"`, "", "<fehlt>"}
	for _, m := range metadata {
		for _, i := range ieee {
			accept := "application/json"
			if m != "<fehlt>" {
				accept += ";odata.metadata=" + m
			}
			if i != "<fehlt>" {
				accept += ";IEEE754Compatible=" + i
			}
			aushandlung(t, "Achsen", []string{accept}, strings.EqualFold(m, "none"), strings.EqualFold(i, "true"))
			n++
		}
	}
	faelle := []struct {
		was     string
		accept  []string
		none, i bool
	}{
		{"Stelle: zweiter Bereich", []string{"application/json;odata.metadata=minimal, application/json;odata.metadata=none"}, false, false},
		{"Stelle: zweiter Bereich IEEE", []string{"application/json, application/json;IEEE754Compatible=true"}, false, false},
		{"Stelle: anderer Typ davor", []string{"text/html, application/json;odata.metadata=none"}, true, false},
		{"Stelle: zweite Kopfzeile", []string{"text/html", "application/json;odata.metadata=none;IEEE754Compatible=true"}, true, true},
		{"Rand: Leerzeichen", []string{"application/json ; odata.metadata = none"}, true, false},
		{"Rand: Tabulator", []string{"application/json;\todata.metadata\t=\tnone\t"}, true, false},
		{"Rand: vor dem Medientyp", []string{" application/json ;IEEE754Compatible=true"}, false, true},
		{"Rand: mehrere Leerzeichen", []string{"application/json;  odata.metadata=none"}, true, false},
		{"Rand: nur Leerzeichen und Tabulator, nicht VT vor dem Namen", []string{"application/json;\x0bodata.metadata=none"}, false, false},
		{"Rand: nicht CR hinter dem Medientyp", []string{"application/json\r;odata.metadata=none"}, false, false},
		{"Rand: nicht NBSP vor dem Namen", []string{"application/json;\xc2\xa0odata.metadata=none"}, false, false},
		{"Länge: Medientyp mit Zusatz hinten", []string{"application/jsonx;odata.metadata=none"}, false, false},
		{"Länge: Medientyp mit Zusatz vorn", []string{"xapplication/json;odata.metadata=none"}, false, false},
		{"Länge: Medientyp mit Teil davor und dahinter", []string{"application/x-ndjson;odata.metadata=none"}, false, false},
		{"Schreibweise: Medientyp", []string{"Application/JSON;odata.metadata=none"}, true, false},
		{"Schreibweise: Name", []string{"application/json;ODATA.METADATA=none"}, true, false},
		{"Schreibweise: IEEE", []string{"application/json;ieee754compatible=true"}, false, true},
		{"Sonderform: nil", nil, false, false},
		{"Sonderform: leeres Slice", []string{}, false, false},
		{"Sonderform: leerer String", []string{""}, false, false},
		{"Sonderform: */*", []string{"*/*;odata.metadata=none"}, false, false},
		{"Sonderform: application/*", []string{"application/*;odata.metadata=none;IEEE754Compatible=true"}, false, false},
		{"Sonderform: text/html", []string{"text/html"}, false, false},
		{"doppelter Parameter: none zuerst", []string{"application/json;odata.metadata=none;odata.metadata=minimal"}, true, false},
		{"doppelter Parameter: minimal zuerst", []string{"application/json;odata.metadata=minimal;odata.metadata=none"}, false, false},
		{"Stelle: zwei Kopfzeilen, die erste gilt", []string{"application/json;odata.metadata=none", "application/json;odata.metadata=minimal"}, true, false},
		{"Länge: Name mit Zusatz", []string{"application/json;odata.metadata2=none;IEEE754Compatible2=true"}, false, false},
		{"Länge: Wert mit Zusatz", []string{"application/json;odata.metadata=nonex;IEEE754Compatible=truex"}, false, false},
		{"Länge: Wert mit Vorsatz", []string{"application/json;odata.metadata=xnone;IEEE754Compatible=xtrue"}, false, false},
		{"doppelter Parameter IEEE", []string{"application/json;IEEE754Compatible=false;IEEE754Compatible=true"}, false, false},
	}
	for _, f := range faelle {
		aushandlung(t, f.was, f.accept, f.none, f.i)
		n++
	}
	if n != 94 {
		t.Errorf("Menge %d, will 94 (64 + 30)", n)
	}
}

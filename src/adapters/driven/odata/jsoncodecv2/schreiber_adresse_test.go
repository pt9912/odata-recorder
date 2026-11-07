package jsoncodecv2_test

import (
	"context"
	"fmt"
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
	return model.NeuesServiceModel("2.0", []string{"Ns"}, []model.EntityType{typ}, []model.EntitySet{{Name: set, EntityType: "Ns.T"}}, nil)
}

// TestAdresseKodierung hält Zeile 23: 130 Zeichen an zwei Stellen (Schlüssel, Set), Dekodierung mit der Standardbibliothek.
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

// uriVon liest den Wert von `uri` des ersten Vorkommens.
func uriVon(body string) string {
	const k = `"uri":"`
	i := strings.Index(body, k)
	if i < 0 {
		return ""
	}
	start := i + len(k)
	return body[start : start+strings.IndexByte(body[start:], '"')]
}

// TestAdresseGleich hält Zeile 24: Adresse, __metadata.uri der Entity und __metadata.uri des Elements sind dieselbe Zeichenfolge.
func TestAdresseGleich(t *testing.T) {
	schluessel := []string{"42", "'O''Hara'", "Id=1,Region='EU'", "\xc3\xa4", "a b", "a/b", "a?b", "a#b", "a%b", "a+b"}
	n := 0
	pruefe := func(m model.ServiceModel, set, key string) {
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
		if uri := uriVon(zl.body()); uri != adr {
			t.Errorf("Entity %q: uri %q, will %q", key, uri, adr)
		}
		zc := neuesZiel()
		if _, err := schreiber().SchreibeCollection(context.Background(), zc, m, model.CollectionAntwort{Wurzel: wurzel, Set: set, Entities: []model.Entity{e}}); err != nil {
			t.Fatal(err)
		}
		if uri := uriVon(zc.body()); uri != adr {
			t.Errorf("Element %q: uri %q, will %q", key, uri, adr)
		}
		n += 3
	}
	for _, k := range schluessel {
		pruefe(testModell(), "Produkte", k)
	}
	pruefe(setModell("Prod\xc3\xbckte"), "Prod\xc3\xbckte", "1")
	if n != 33 {
		t.Errorf("Zellen %d, will 33", n)
	}
}

// wurzelGueltig und wurzelUngueltig sind die Mengen von Zeile 25 (zweite Tabelle des Plans).
func wurzelGueltig() []string {
	return []string{
		"http://h", "https://h", "http://h:8080", "http://[::1]:8080", "http://h/a", "http://h/a/b", "http://h/%C3%A4", "http://127.0.0.1:1",
		"http://h/a.b", "http://h/...", "http://h/" + strings.Repeat("a", 2048-len("http://h/")),
		// gültige Seite der Zeichen und Ziffern: Host mit - _ ~ und Großbuchstaben, Port mit fünf Ziffern, Port mit Pfad, IPv6 mit Punkt und Großbuchstaben,
		// Hex-Kleinbuchstaben und Steuerzeichen im %-Teil, %2F im Segment
		"http://a-b", "http://a_b", "http://a~b", "http://A.B", "http://h:65535", "http://h:80/a/b", "http://[::FFFF:1.2.3.4]:80", "http://h/%c3%a4",
		"http://h/a-b/c_d/e~f/G", "http://h/%00", "http://h/%1F", "http://h/%2F",
	}
}

func wurzelUngueltig() []string {
	return []string{
		"", "ftp://h", "HTTP://h", "Https://h", "//h", "h", "http://", "http:///a", "http://h/", "http://h//a", "http://h/a/", "http://u@h", "http://u:p@h",
		"http://h?x=1", "http://h#f", "http://h/a?b", "http://h/a#b", "http://h/a" + bs + "b", "http://h/a b", " http://h", "http://h/\xc3\xa4", "http://h\xc3\xa4",
		"http://h/%", "http://h/%1", "http://h/%GG", "http://h/.", "http://h/..", "http://h/a/./b", "http://h:", "http://h:123456", "http://h:80a", "http://h:80:90",
		"http://[::1", "http://[]", "http://[g::1]", "http://h%41", "http://h\n", "http://h\r\nX: y", "http://h/a\x00", "http://h/a\x7f",
		"http://h/" + strings.Repeat("a", 2049-len("http://h/")), "http://h/%2E", "http://h/%2e%2e", "http://h/.%2E",
	}
}

// TestWurzel hält Zeile 25: die Form von Wurzel an Entity, Collection, leerer Collection und Adresse.
func TestWurzel(t *testing.T) {
	gueltig, ungueltig := wurzelGueltig(), wurzelUngueltig()
	if len(gueltig) != 23 || len(ungueltig) != 44 {
		t.Fatalf("Menge gültig %d, ungültig %d; will 23, 44", len(gueltig), len(ungueltig))
	}
	for _, w := range gueltig {
		wurzelGilt(t, w, true)
	}
	for _, w := range ungueltig {
		wurzelGilt(t, w, false)
	}
}

// TestWurzelPort hält Zeile 25 für den Port (SPEC-172 Nr. 6 h: ein bis fünf Ziffern, der Wert wird nicht geprüft): jede Länge 1 bis 5 mit
// der Ziffer 1, dazu Werte über 65535 und Nullen; sechs Ziffern, ein Buchstabe und kein Port sind ungültig. Menge 5 + 5 + 3, je 4 Operationen.
func TestWurzelPort(t *testing.T) {
	var gueltig []string
	for l := 1; l <= 5; l++ {
		gueltig = append(gueltig, strings.Repeat("1", l))
	}
	gueltig = append(gueltig, "99999", "65536", "00000", "0", "00")
	ungueltig := []string{"111111", "1a", ""}
	for _, p := range gueltig {
		wurzelGilt(t, "http://h:"+p, true)
		wurzelGilt(t, "http://h:"+p+"/a", true)
	}
	for _, p := range ungueltig {
		wurzelGilt(t, "http://h:"+p, false)
	}
	if len(gueltig) != 10 || len(ungueltig) != 3 {
		t.Errorf("Menge gültig %d, ungültig %d; will 10, 3", len(gueltig), len(ungueltig))
	}
}

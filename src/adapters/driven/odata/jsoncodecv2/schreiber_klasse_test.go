package jsoncodecv2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Tests dieser Datei halten die Klassen der Zusagen als Schleifen über die ganze Menge (AGENTS.md §3.17): jedes Byte 0x00 bis 0xFF
// an jeder Stelle, die der Code verschieden behandelt, mit der erlaubten Menge als Orakel im Test.

func istZiffer(c byte) bool { return c >= '0' && c <= '9' }

func istHex(c byte) bool { return istZiffer(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' }

// istHostZeichen ist die Menge aus SPEC-169 Nr. 2 für Host und Segment: ASCII-Buchstaben, ASCII-Ziffern und `-._~`.
func istHostZeichen(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || istZiffer(c) || strings.IndexByte("-._~", c) >= 0
}

// wurzelGilt ruft Adresse, Entity, Collection mit einem Element und leere Collection mit w und prüft, dass alle vier dasselbe urteilen; 4 Zellen.
func wurzelGilt(t *testing.T, w string, erwartet bool) {
	t.Helper()
	was := fmt.Sprintf("Wurzel %q", w)
	e, c, leer := gutEntity(), gutCollection(), model.CollectionAntwort{Set: "Produkte"}
	c.Entities = c.Entities[:1]
	e.Wurzel, c.Wurzel, leer.Wurzel = w, w, w
	if erwartet {
		if body := entityBody(t, was, e); !strings.Contains(body, `"uri":"`+w+`/Produkte(1)"`) {
			t.Errorf("%s: Entity %s", was, body)
		}
		if body := collectionBody(t, was, c); !strings.Contains(body, `"uri":"`+w+`/Produkte(1)"`) {
			t.Errorf("%s: Collection %s", was, body)
		}
		if body := collectionBody(t, was+" leer", leer); body != `{"d":{"results":[]}}` {
			t.Errorf("%s: leere Collection %s", was, body)
		}
		if a, err := schreiber().Adresse(testModell(), w, "Produkte", "1"); err != nil || a != w+"/Produkte(1)" {
			t.Errorf("%s: Adresse %q, %v", was, a, err)
		}
		return
	}
	entityScheitert(t, was, e)
	collectionScheitert(t, was, c)
	collectionScheitert(t, was+" leer", leer)
	a, err := schreiber().Adresse(testModell(), w, "Produkte", "1")
	if a != "" {
		t.Errorf("%s: Adresse %q", was, a)
	}
	istUngueltig(t, err, was+" Adresse")
}

// TestWurzelZeichenMenge hält Zeile 26 als Klasse: je Stelle der Wurzel (Schema, Host, Port, Segment, `%`-Ziffern, IPv6) jedes der 256 Bytes,
// an Entity, Collection, leerer Collection und Adresse.
func TestWurzelZeichenMenge(t *testing.T) {
	stellen := 0
	zellen := 0
	je := func(bau func(c byte) string, gilt func(c byte) bool) {
		stellen++
		for i := range 256 {
			c := byte(i)
			wurzelGilt(t, bau(c), gilt(c))
			zellen += 4
		}
	}
	const kopf = "http://h"
	for i := range len(kopf) - 1 { // Schema und `://`: die sieben Stellen von `http://`
		orig := kopf[i]
		je(func(c byte) string { return kopf[:i] + string([]byte{c}) + kopf[i+1:] },
			func(c byte) bool { return c == orig })
	}
	je(func(c byte) string { return "http://" + string([]byte{c}) + "b" }, istHostZeichen)
	je(func(c byte) string { return "http://a" + string([]byte{c}) + "b" }, func(c byte) bool { return istHostZeichen(c) || c == '/' })
	je(func(c byte) string { return "http://a" + string([]byte{c}) }, istHostZeichen)
	je(func(c byte) string { return "http://h:8" + string([]byte{c}) + "0" }, func(c byte) bool { return istZiffer(c) || c == '/' })
	je(func(c byte) string { return "http://h/" + string([]byte{c}) + "b" }, istHostZeichen)
	je(func(c byte) string { return "http://h/a" + string([]byte{c}) + "b" }, func(c byte) bool { return istHostZeichen(c) || c == '/' })
	je(func(c byte) string { return "http://h/a" + string([]byte{c}) }, istHostZeichen)
	je(func(c byte) string { return "http://h/%" + string([]byte{c}) + "0" }, istHex)
	je(func(c byte) string { return "http://h/%0" + string([]byte{c}) }, istHex)
	je(func(c byte) string { return "http://[" + string([]byte{c}) + "]" }, func(c byte) bool { return istHex(c) || c == ':' || c == '.' })
	je(func(c byte) string { return "http://[::" + string([]byte{c}) + "]" }, func(c byte) bool { return istHex(c) || c == ':' || c == '.' })
	if stellen != 18 || zellen != 18*256*4 {
		t.Errorf("Stellen %d, Zellen %d; will 18, %d", stellen, zellen, 18*256*4)
	}
}

// textGilt schreibt eine Entity mit dem Text in der nicht nullbaren Property der Art und liefert, ob der Schreiber ihn annimmt;
// jeder andere Fehler als ErrUngueltig ist ein Fehlschlag.
func textGilt(t *testing.T, art model.WertArt, prop, text string) bool {
	t.Helper()
	z := neuesZiel()
	_, err := schreiber().SchreibeEntity(context.Background(), z, testModell(), alleAntwort(pw(prop, art, text)))
	if err != nil && !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("%s %q: %v", art, text, err)
	}
	return err == nil
}

// TestGuidStellen hält Zeile 13: jede der 36 Stellen mit jedem der 256 Bytes und jede Länge von 0 bis 40.
func TestGuidStellen(t *testing.T) {
	const basis = "0f8fad5b-d9cb-469f-a165-70867728950e"
	n := 0
	for i := range len(basis) {
		for b := range 256 {
			c := byte(b)
			text := basis[:i] + string([]byte{c}) + basis[i+1:]
			want := c >= '0' && c <= '9' || c >= 'a' && c <= 'f'
			if i == 8 || i == 13 || i == 18 || i == 23 {
				want = c == '-'
			}
			if got := textGilt(t, model.ArtGUID, "G", text); got != want {
				t.Errorf("Guid Stelle %d Byte %#x: angenommen %v, will %v", i, c, got, want)
			}
			n++
		}
	}
	for l := range 41 {
		text := basis[:min(l, len(basis))] + strings.Repeat("0", max(0, l-len(basis)))
		if got := textGilt(t, model.ArtGUID, "G", text); got != (l == 36) {
			t.Errorf("Guid Länge %d: angenommen %v", l, got)
		}
		n++
	}
	if n != 36*256+41 {
		t.Errorf("Zellen %d, will %d", n, 36*256+41)
	}
}

// ziffernFall ist die Klasse einer Zahlart (Zeile 12): ein Basis-Text, an dessen Stellen jedes Byte steht, dazu Längen und Randfälle; das
// Orakel urteilt über jeden Text.
type ziffernFall struct {
	art     model.WertArt
	prop    string
	basis   string
	laengen []string
	rand    []string
	orakel  func(s string) bool
}

func einsen(ls ...int) []string {
	var out []string
	for _, l := range ls {
		out = append(out, strings.Repeat("1", l))
	}
	return out
}

func bis(n int) []int {
	var out []int
	for i := range n {
		out = append(out, i)
	}
	return out
}

// ganzOrakel: strconv.ParseInt mit Rückprobe auf gleichen Text.
func ganzOrakel(bits int) func(string) bool {
	return func(s string) bool {
		v, err := strconv.ParseInt(s, 10, bits)
		return err == nil && strconv.FormatInt(v, 10) == s
	}
}

// ziffernFaelle sind die drei Zahlarten der Klasse „Ziffern“ mit 14 + 8, 23 + 8 und 9 + 12 Längen- und Randfällen.
func ziffernFaelle() []ziffernFall {
	dezimal := regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]*[1-9])?$`)
	return []ziffernFall{
		{model.ArtInt32, "I32", "-2147483648", einsen(bis(14)...),
			[]string{"2147483647", "2147483648", "-2147483648", "-2147483649", "0", "-0", "-1", "00"}, ganzOrakel(32)},
		{model.ArtInt64, "I64", "-9223372036854775808", einsen(bis(23)...),
			[]string{"9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809", "0", "-0", "-1", "00"}, ganzOrakel(64)},
		{model.ArtDecimal, "Dec", "-10.25", einsen(0, 1, 2, 1022, 1023, 1024, 1025, 1026, 2048),
			[]string{"0", "-0", "0.5", "-0.5", "0.50", "1.", "10", "-10.25", "0.0", "00", "1.5.5", "12345678901234567890.12345678901234567890"},
			func(s string) bool { return len(s) <= 1024 && s != "-0" && dezimal.MatchString(s) }},
	}
}

// TestZifferKlasse hält Zeile 12: jede Stelle des Basis-Textes mit jedem der 256 Bytes, dazu Längen und Randfälle, je gegen das Orakel.
func TestZifferKlasse(t *testing.T) {
	n := 0
	for _, f := range ziffernFaelle() {
		pruefe := func(text string) {
			n++
			if got := textGilt(t, f.art, f.prop, text); got != f.orakel(text) {
				t.Errorf("%s %.40q (Länge %d): angenommen %v, will %v", f.art, text, len(text), got, f.orakel(text))
			}
		}
		for i := range len(f.basis) {
			for b := range 256 {
				pruefe(f.basis[:i] + string([]byte{byte(b)}) + f.basis[i+1:])
			}
		}
		for _, s := range append(append([]string{}, f.laengen...), f.rand...) {
			pruefe(s)
		}
	}
	if n != 9546 {
		t.Errorf("Zellen %d, will 9546", n)
	}
}

// zeitOrakel urteilt über den Text eines Zeitpunkts: ein regulärer Ausdruck für die Form (Bruchteil von 0 bis 3 Ziffern, die letzte nicht
// `0`) und time.Date für Kalender und Bereich.
func zeitOrakel(re *regexp.Regexp, s string) bool {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	z := make([]int, 6)
	for i := range z {
		z[i], _ = strconv.Atoi(m[i+1])
	}
	t := time.Date(z[0], time.Month(z[1]), z[2], 0, 0, 0, 0, time.UTC)
	return z[0] >= 1 && t.Year() == z[0] && int(t.Month()) == z[1] && t.Day() == z[2] && z[3] <= 23 && z[4] <= 59 && z[5] <= 59
}

// TestZeitpunktKlasse hält Zeile 14: 24 Stellen × 256 Bytes, Bruchteile der Länge 0 bis 13 in drei Formen, jede Länge 1 bis 30 und die Feldgrenzen.
func TestZeitpunktKlasse(t *testing.T) {
	re := regexp.MustCompile(`^([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2}):([0-9]{2})(?:\.[0-9]{0,2}[1-9])?Z$`)
	const basis = "0001-01-01T00:00:00.999Z"
	n := 0
	pruefe := func(text string) {
		n++
		if got, want := textGilt(t, model.ArtDateTimeOffset, "T", text), zeitOrakel(re, text); got != want {
			t.Errorf("%q: angenommen %v, will %v", text, got, want)
		}
	}
	for i := range len(basis) {
		for b := range 256 {
			pruefe(basis[:i] + string([]byte{byte(b)}) + basis[i+1:])
		}
	}
	const kopf = "2020-01-01T00:00:00"
	for l := range 14 {
		pruefe(kopf + "." + strings.Repeat("1", l) + "Z")
		pruefe(kopf + "." + strings.Repeat("1", max(0, l-1)) + "0Z")
		pruefe(kopf + "." + strings.Repeat("0", l) + "Z")
	}
	lang := kopf + ".123456789012345678"
	for l := 1; l <= 30; l++ {
		pruefe(lang[:min(l-1, len(lang))] + "Z")
	}
	for _, f := range []string{
		"2020-00-01", "2020-01-01", "2020-12-01", "2020-13-01", // Monat
		"2020-01-00", "2020-01-01", "2020-01-31", "2020-01-32", "2020-02-29", "2023-02-29", "1900-02-29", "2000-02-29", "2020-04-31", "2020-02-30", "2020-04-30", "2020-02-28", "2019-02-29", // Tag
		"0000-01-01", "0001-01-01", "9999-12-31", "10000-01-01", // Jahr
	} {
		pruefe(f + "T00:00:00Z")
	}
	for _, f := range []string{"23:59:59", "24:00:00", "00:60:00", "00:00:60", "00:00:59", "12:34:56"} {
		pruefe("2020-01-01T" + f + "Z")
	}
	if n != 6243 {
		t.Errorf("Zellen %d, will 6243", n)
	}
}

// TestAuswahlStellen hält Zeile 16 für die Stelle: jedes der fünf Elemente einer Auswahl, eines nach dem anderen ungültig, an Entity und Collection.
func TestAuswahlStellen(t *testing.T) {
	ungueltig := []string{"Nope", "", "name", "Extra", "Name "}
	n := 0
	for _, el := range ungueltig {
		for i := range 5 {
			liste := []string{"Id", "Name", "Kategorie", "Preis", "*"}
			liste[i] = el
			e, c := gutEntity(), gutCollection()
			e.Auswahl, c.Auswahl = liste, liste
			name := fmt.Sprintf("Auswahl %q an Stelle %d", el, i)
			entityScheitert(t, name, e)
			collectionScheitert(t, name, c)
			n += 2
		}
	}
	if n != 50 {
		t.Errorf("Zellen %d, will 50", n)
	}
}

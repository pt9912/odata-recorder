package jsoncodec_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

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

// wurzelGilt ruft Adresse, Entity und Collection mit w und prüft, dass alle drei dasselbe urteilen; 3 Zellen.
func wurzelGilt(t *testing.T, w string, erwartet bool) {
	t.Helper()
	was := fmt.Sprintf("Wurzel %q", w)
	if erwartet {
		e, c := gutEntity(), gutCollection()
		e.Wurzel, c.Wurzel = w, w
		if body := entityBody(t, was, e); !strings.Contains(body, `"@odata.id":"`+w+`/Produkte(1)"`) {
			t.Errorf("%s: Entity %s", was, body)
		}
		collectionBody(t, was, c)
		if a, err := schreiber().Adresse(testModell(), w, "Produkte", "1"); err != nil || a != w+"/Produkte(1)" {
			t.Errorf("%s: Adresse %q, %v", was, a, err)
		}
		return
	}
	e, c := gutEntity(), gutCollection()
	e.Wurzel, c.Wurzel = w, w
	entityScheitert(t, was, e)
	collectionScheitert(t, was, c)
	a, err := schreiber().Adresse(testModell(), w, "Produkte", "1")
	if a != "" {
		t.Errorf("%s: Adresse %q", was, a)
	}
	istUngueltig(t, err, was+" Adresse")
}

// TestWurzelZeichenMenge hält Zeile 21 als Klasse: je Stelle der Wurzel (Schema, Host, Port, Segment, `%`-Ziffern, IPv6) jedes der 256 Bytes.
func TestWurzelZeichenMenge(t *testing.T) {
	stellen := 0
	zellen := 0
	je := func(bau func(c byte) string, gilt func(c byte) bool) {
		stellen++
		for i := range 256 {
			c := byte(i)
			wurzelGilt(t, bau(c), gilt(c))
			zellen += 3
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
	if stellen != 18 || zellen != 18*256*3 {
		t.Errorf("Stellen %d, Zellen %d; will 18, %d", stellen, zellen, 18*256*3)
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

// TestGuidStellen hält Zeile 11 für Guid als Klasse: jede der 36 Stellen mit jedem der 256 Bytes und jede Länge von 0 bis 40.
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

// TestZeitpunktKlasse hält Zeile 11 für DateTimeOffset als Klasse: jede Länge des Bruchteils von 0 bis 14, jedes Byte an jeder der zwölf Stellen
// des Bruchteils, jede Länge des Textes unter 20 (die Wache gegen den Schnitt) und jedes Byte an jeder der 20 Stellen des Textes.
func TestZeitpunktKlasse(t *testing.T) {
	const kopf = "2020-01-01T00:00:00"
	n := 0
	for l := range 15 {
		for _, f := range []struct {
			was  string
			text string
			gilt bool
		}{
			{"Punkt, Ziffern 1", kopf + "." + strings.Repeat("1", l) + "Z", l >= 1 && l <= 12},
			{"ohne Punkt", kopf + strings.Repeat("1", l) + "Z", l == 0},
			{"letzte Ziffer 0", kopf + "." + strings.Repeat("1", max(0, l-1)) + "0Z", false},
		} {
			if l == 0 && f.was == "letzte Ziffer 0" {
				continue
			}
			if got := textGilt(t, model.ArtDateTimeOffset, "T", f.text); got != f.gilt {
				t.Errorf("%s, Länge %d (%q): angenommen %v, will %v", f.was, l, f.text, got, f.gilt)
			}
			n++
		}
	}
	for p := range 12 {
		for b := range 256 {
			c := byte(b)
			frac := []byte(strings.Repeat("1", 12))
			frac[p] = c
			want := istZiffer(c) && !(p == 11 && c == '0')
			if got := textGilt(t, model.ArtDateTimeOffset, "T", kopf+"."+string(frac)+"Z"); got != want {
				t.Errorf("Bruchteil Stelle %d Byte %#x: angenommen %v, will %v", p, c, got, want)
			}
			n++
		}
	}
	// das Zeichen hinter den Sekunden, wenn eine Ziffer und das `Z` folgen: nur der Punkt leitet einen Bruchteil ein
	for b := range 256 {
		c := byte(b)
		if got := textGilt(t, model.ArtDateTimeOffset, "T", kopf+string([]byte{c})+"1Z"); got != (c == '.') {
			t.Errorf("Trennzeichen Byte %#x vor dem Bruchteil: angenommen %v", c, got)
		}
		n++
	}
	// Länge unter 20: jedes Präfix von `2020-01-01T00:00:00` mit `Z`; 18 und 19 haben T, die beiden Doppelpunkte und das Z an den festen Stellen.
	for l := 1; l <= 20; l++ {
		text := kopf[:l-1] + "Z"
		if got := textGilt(t, model.ArtDateTimeOffset, "T", text); got != (l == 20) {
			t.Errorf("Länge %d (%q): angenommen %v", l, text, got)
		}
		n++
	}
	const ganz = kopf + "Z"
	enger := map[int]func(byte) bool{
		5: func(c byte) bool { return c <= '1' }, 6: func(c byte) bool { return c != '0' }, 8: func(c byte) bool { return c <= '3' },
		9: func(c byte) bool { return c != '0' }, 11: func(c byte) bool { return c <= '2' }, 14: func(c byte) bool { return c <= '5' },
		17: func(c byte) bool { return c <= '5' },
	}
	for p := range len(ganz) {
		for b := range 256 {
			c := byte(b)
			var want bool
			switch p {
			case 4, 7, 10, 13, 16, 19:
				want = c == ganz[p]
			default:
				want = istZiffer(c)
				if f, ok := enger[p]; ok {
					want = want && f(c)
				}
			}
			text := ganz[:p] + string([]byte{c}) + ganz[p+1:]
			if got := textGilt(t, model.ArtDateTimeOffset, "T", text); got != want {
				t.Errorf("Stelle %d Byte %#x (%q): angenommen %v, will %v", p, c, text, got, want)
			}
			n++
		}
	}
	if want := 15*3 - 1 + 12*256 + 256 + 20 + 20*256; n != want {
		t.Errorf("Zellen %d, will %d", n, want)
	}
}

// TestAuswahlStellen hält Zeile 12 für die Stelle: jedes der fünf Elemente einer Auswahl, eines nach dem anderen ungültig, in beiden Stufen.
func TestAuswahlStellen(t *testing.T) {
	ungueltig := []string{"Nope", "", "name", "Extra", "Name "}
	n := 0
	for _, el := range ungueltig {
		for i := range 5 {
			liste := []string{"Id", "Name", "Kategorie", "Preis", "*"}
			liste[i] = el
			for _, accept := range [][]string{nil, {"application/json;odata.metadata=none"}} {
				e, c := gutEntity(), gutCollection()
				e.Auswahl, c.Auswahl, e.Accept, c.Accept = liste, liste, accept, accept
				name := fmt.Sprintf("Auswahl %q an Stelle %d %v", el, i, accept)
				entityScheitert(t, name, e)
				collectionScheitert(t, name, c)
				n += 2
			}
		}
	}
	if n != 100 {
		t.Errorf("Zellen %d, will 100", n)
	}
}

// TestAuswahlReihenfolge hält SPEC-169 Nr. 5: die Elemente der Auswahl stehen im Kontext in der Reihenfolge der Übergabe; alle 24 Anordnungen
// von vier Elementen, je dreimal im Prozess, Entity und Collection.
func TestAuswahlReihenfolge(t *testing.T) {
	var anordnungen func(xs []string) [][]string
	anordnungen = func(xs []string) [][]string {
		if len(xs) <= 1 {
			return [][]string{append([]string{}, xs...)}
		}
		var out [][]string
		for i := range xs {
			rest := append(append([]string{}, xs[:i]...), xs[i+1:]...)
			for _, p := range anordnungen(rest) {
				out = append(out, append([]string{xs[i]}, p...))
			}
		}
		return out
	}
	alle := anordnungen([]string{"Id", "Name", "Kategorie", "*"})
	n := 0
	for wdh := range 3 {
		for _, liste := range alle {
			e, c := gutEntity(), gutCollection()
			e.Auswahl, c.Auswahl = liste, liste
			kontext := wurzel + "/$metadata#Produkte(" + strings.Join(liste, ",") + ")"
			if body := entityBody(t, "Auswahl", e); !strings.Contains(body, `"@odata.context":"`+kontext+`/$entity"`) {
				t.Errorf("Entity %v (Wiederholung %d): %s", liste, wdh, body)
			}
			if body := collectionBody(t, "Auswahl", c); !strings.HasPrefix(body, `{"@odata.context":"`+kontext+`","value"`) {
				t.Errorf("Collection %v (Wiederholung %d): %s", liste, wdh, body)
			}
			n += 2
		}
	}
	if len(alle) != 24 || n != 144 {
		t.Errorf("Anordnungen %d, Zellen %d; will 24, 144", len(alle), n)
	}
}

// TestAcceptStellen hält Zeile 22 für die Stelle des Bereichs `application/json`: jede von fünf Stellen in einem Wert (Bereiche mit Komma)
// und jede von fünf Kopfzeilen; der erste Bereich dieses Typs gilt, ein späterer ändert nichts.
func TestAcceptStellen(t *testing.T) {
	n := 0
	for i := range 5 {
		bereiche := []string{"text/html", "text/plain", "image/png", "application/xml", "*/*"}
		bereiche[i] = "application/json;odata.metadata=none;IEEE754Compatible=true"
		spaeter := append(append([]string{}, bereiche...), "application/json;odata.metadata=minimal")
		aushandlung(t, fmt.Sprintf("Bereich %d eines Wertes", i), []string{strings.Join(bereiche, ", ")}, true, true)
		aushandlung(t, fmt.Sprintf("Kopfzeile %d", i), bereiche, true, true)
		aushandlung(t, fmt.Sprintf("Bereich %d, späterer Bereich ändert nichts", i), []string{strings.Join(spaeter, ",")}, true, true)
		aushandlung(t, fmt.Sprintf("Kopfzeile %d, spätere Zeile ändert nichts", i), spaeter, true, true)
		n += 4
	}
	if n != 20 {
		t.Errorf("Zellen %d, will 20", n)
	}
}

package urlparser_test

import (
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// formenGruppe sind die Elemente einer Untermenge von TestLiteralFormen; erwartet ist ihre Zahl aus der Rechnung von Hand.
type formenGruppe struct {
	name     string
	erwartet int
	els      []stellenElement
}

// formenBau sammelt die Elemente einer Gruppe mit der Erwartung, die der Aufrufer von Hand nennt.
type formenBau struct {
	gruppe string
	els    []stellenElement
}

func (b *formenBau) beschreibung(form string) string {
	return fmt.Sprintf("%s %d (%q)", b.gruppe, len(b.els)+1, form)
}

// typ erwartet `literal_typ`, lit erwartet `literal`, gelesen ein gelesenes Literal mit dem kanonischen Text.
func (b *formenBau) typ(formen ...string) {
	for _, form := range formen {
		b.els = append(b.els, stellenElement{b.beschreibung(form), form, model.URLLiteralTyp, 0, ""})
	}
}

func (b *formenBau) lit(formen ...string) {
	for _, form := range formen {
		b.els = append(b.els, stellenElement{b.beschreibung(form), form, model.URLLiteral, 0, ""})
	}
}

func (b *formenBau) gelesen(form string, art model.LiteralArt, text string) {
	b.els = append(b.els, stellenElement{b.beschreibung(form), form, "", art, text})
}

// schreibweisen liefert alle 2^n Schreibweisen eines kleingeschriebenen ASCII-Worts der Länge n: Bit i der Maske macht den
// Buchstaben i groß.
func schreibweisen(klein string) []string {
	var r []string
	for maske := range 1 << len(klein) {
		b := []byte(klein)
		for i := range b {
			if maske>>i&1 == 1 {
				b[i] -= 'a' - 'A'
			}
		}
		r = append(r, string(b))
	}
	return r
}

// ohne entfernt s aus der Liste.
func ohne(liste []string, s string) []string {
	var r []string
	for _, e := range liste {
		if e != s {
			r = append(r, e)
		}
	}
	return r
}

// uhrzeitFormen ist die Gruppe U (68 Elemente): Anfänge von `12:30:45.123456789012` der Länge 1 bis 21 (21), 13 Ziffern Bruchteil
// (1), Uhrzeit mit Rest (24), Bereiche (8), Ziffern jenseits von ASCII (10: 7 mit längerer Uhrzeit, 3 mit Länge 8), vier falsche Trennungen (4). In `klassen` steht je
// Länge von Hand: Z eine Zahl (Länge 1 und 2), L ein `literal` (Länge 3, 4, 6, 7, 9), T ein `literal_typ` (Länge 5, 8, 10 bis 21).
func uhrzeitFormen() []stellenElement {
	b := &formenBau{gruppe: "U1"}
	const voll, klassen = "12:30:45.123456789012", "ZZLLTLLTLTTTTTTTTTTTT"
	for i := 1; i <= len(voll); i++ {
		switch klassen[i-1] {
		case 'Z':
			b.gelesen(voll[:i], model.LiteralZahl, voll[:i])
		case 'T':
			b.typ(voll[:i])
		default:
			b.lit(voll[:i])
		}
	}
	b.gruppe = "U2"
	b.lit("12:30:45.1234567890123")
	b.gruppe = "U3"
	for _, basis := range []string{"12:30", "12:30:45", "12:30:45.5"} {
		for _, suffix := range []string{"Z", "z", "+01:00", "-01:00", "+01", "x", ":", "."} {
			b.lit(basis + suffix)
		}
	}
	b.gruppe = "U4"
	b.typ("00:00", "23:59", "23:59:59", "00:00:00.000000000000")
	b.lit("24:00", "23:60", "12:30:60", "99:99")
	b.gruppe = "U5"
	// U+0663 (arabisch-indische Ziffer drei) an jeder der sechs Ziffernstellen von `12:30:45`, U+FF11 als erste Ziffer
	const uhr, arabisch = "12:30:45", "\xd9\xa3"
	for _, i := range []int{0, 1, 3, 4, 6, 7} {
		b.lit(uhr[:i] + arabisch + uhr[i+1:])
	}
	b.lit("\xef\xbc\x912:30:45")
	// U+0663 (zwei Bytes) statt zweier ASCII-Ziffern: die Länge 8 bleibt, die Ziffernstelle (Stunde, Minute, Sekunde) wird erreicht
	b.lit(arabisch + ":30:45")
	b.lit("12:" + arabisch + ":45")
	b.lit("12:30:" + arabisch)
	b.gruppe = "U6"
	b.lit("12-30", "1:30", "123:30", "12::30")
	return b.els
}

// exponentFormen ist die Gruppe E (51 Elemente): 24 gültige Exponenten (Mantisse × `e`/`E` × Vorzeichen × Ziffern), 21 Formen mit
// falschem oder fehlendem Exponent, Rest dahinter, Leerraum, falscher Mantisse oder Ziffern jenseits von ASCII, und sechs Anfänge
// von `-1.5E-2`.
func exponentFormen() []stellenElement {
	b := &formenBau{gruppe: "E1"}
	for _, mantisse := range []string{"1", "-1.5"} {
		for _, e := range []string{"e", "E"} {
			for _, vorzeichen := range []string{"", "+", "-"} {
				for _, ziffern := range []string{"2", "30"} {
					b.typ(mantisse + e + vorzeichen + ziffern)
				}
			}
		}
	}
	b.gruppe = "E2"
	// kein oder falscher Exponent (7), Rest hinter dem Exponenten (6), Leerraum und Hex (2), falsche Mantisse (4), U+FF11 als
	// Mantisse und U+FF13 als Exponent (2)
	b.lit("e3", "-e3", "1e", "1e+", "1e-", "1ee3", "1e+-3")
	b.lit("1e3e", "1e3e3", "1e3x", "1e3L", "1e3.5", "1e3-")
	b.lit("1e 3", "0x1e3")
	b.lit(".5e3", "5.e3", "+1e3", "1.5.5e3")
	b.lit("\xef\xbc\x91e3", "1e\xef\xbc\x93")
	b.gruppe = "E3"
	b.lit("-", "-1.", "-1.5E", "-1.5E-")
	b.gelesen("-1", model.LiteralZahl, "-1")
	b.gelesen("-1.5", model.LiteralZahl, "-1.5")
	return b.els
}

// sonderwerte ist die Gruppe S (43 Elemente): `INF`, `-INF`, `NaN` genau so (3), jede andere Groß- und Kleinschreibung der
// drei (21), Anfänge (7), Zusätze davor und dahinter (10) und zwei Zeichen jenseits von ASCII, die ASCII-Großbuchstaben ähneln (2).
func sonderwerte() []stellenElement {
	b := &formenBau{gruppe: "S1"}
	b.typ("INF", "-INF", "NaN")
	b.gruppe = "S2"
	b.lit(ohne(schreibweisen("inf"), "INF")...)
	b.lit(ohne(schreibweisen("nan"), "NaN")...)
	for _, f := range ohne(schreibweisen("inf"), "INF") {
		b.lit("-" + f)
	}
	b.gruppe = "S3"
	b.lit("I", "IN", "N", "Na", "-", "-I", "-IN")
	b.gruppe = "S4"
	b.lit("+INF", "+NaN", "-NaN", "--INF", "INF-", "INFx", "NaNx", "-INFx", "INF ", "- INF")
	b.gruppe = "S5"
	// U+0131 und U+0130
	b.lit("\xc4\xb1nf", "\xc4\xb0NF")
	return b.els
}

// wahrheitswerte ist die Gruppe W (64 Elemente): alle 16 Schreibweisen von `true` und 32 von `false` (Text klein), dazu 16
// Formen, die `literal` bleiben: Anfänge (7), Rest (3), die Faltung U+017F zu `s` (2), Homoglyphen U+0435 und U+FF54 (2), Leerraum (2).
func wahrheitswerte() []stellenElement {
	b := &formenBau{gruppe: "W1"}
	for _, wort := range []string{"true", "false"} {
		for _, f := range schreibweisen(wort) {
			b.gelesen(f, model.LiteralWahrheitswert, wort)
		}
	}
	b.gruppe = "W2"
	b.lit("t", "tr", "tru", "f", "fa", "fal", "fals")
	b.lit("truex", "falsex", "true1")
	b.lit("fal\xc5\xbfe", "FAL\xc5\xbfE")
	b.lit("tru\xd0\xb5", "\xef\xbd\x94rue")
	b.lit(" true", "true ")
	return b.els
}

// apostrophFormen sind sieben Formen mit einem Präfix und `'` (Gruppe A): zwei `literal_typ`, fünf `literal`.
func apostrophFormen() []stellenElement {
	b := &formenBau{gruppe: "A"}
	b.typ("X''", "X'a''b'")
	b.lit("X'a", "X'a'b", "X'a''", "X 'a'", "'a'X")
	return b.els
}

// TestLiteralFormen (Zeile 38,SPEC-161 Nr. 6 g, Klasse „jede Form hinter jedem Anfang“): Schleife über die Menge der Uhrzeit, des
// Exponenten, der Sonderwerte, des Wahrheitswerts und der Apostroph-Anfänge, 233 Elemente (68 + 51 + 43 + 64 und 7). Jede
// Erwartung steht von Hand im Test, das Orakel ist nie der Parser; jedes Element wird einzeln gemeldet, auch eine Panik.
func TestLiteralFormen(t *testing.T) {
	gruppen := []formenGruppe{
		{"Uhrzeit", 68, uhrzeitFormen()},
		{"Exponent", 51, exponentFormen()},
		{"Sonderwerte", 43, sonderwerte()},
		{"Wahrheitswert", 64, wahrheitswerte()},
		{"Apostroph", 7, apostrophFormen()},
	}
	summe := 0
	for _, g := range gruppen {
		if len(g.els) != g.erwartet {
			t.Errorf("%s: %d Elemente, erwartet %d", g.name, len(g.els), g.erwartet)
		}
		for _, e := range g.els {
			pruefeStellenElement(t, e)
		}
		summe += len(g.els)
	}
	if summe != 233 {
		t.Fatalf("Zählprobe: %d Elemente, erwartet 233", summe)
	}
}

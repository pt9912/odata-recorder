package urlparserv2_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Tests dieser Datei sind Schleifen über die Mengen der Zusagen der Zeilen 6, 19 und 23; die Zeichen jenseits von ASCII
// entstehen aus Werten (flRune, AGENTS.md §3.18).

// flAufrufZeichen sind die Zeichen in der Mitte eines Wortlaufs vor `(`: U+0001 bis U+007F außer U+0020, `(`, `)`, `,` und `'`
// (die Läufe beenden), dazu sechs Zeichen jenseits von ASCII. Menge 128 = 122 + 6.
func flAufrufZeichen() []string {
	var zeichen []string
	for c := byte(1); c < 0x80; c++ {
		if c != ' ' && c != '(' && c != ')' && c != ',' && c != '\'' {
			zeichen = append(zeichen, string([]byte{c}))
		}
	}
	for _, r := range []int{0x85, 0xA0, 0x2003, 0x200B, 0x3000, 0xFEFF} {
		zeichen = append(zeichen, flRune(r))
	}
	return zeichen
}

// TestFlAufrufWortlaeufe (Zeile 23, SPEC-164 Nr. 4 c 1): ein Wortlauf unmittelbar vor `(` ist ein Aufruf, `filter`, und der Aufruf
// geht der Prüfung des Namens, von `$` und `@` voraus: auch mit jedem Zeichen in der Mitte, das kein Name erlaubt, mit `$` und `@`
// am Anfang, mit einem Pfad oder qualifizierten Namen mit ungültigem Glied und mit einem Namen über 128 Zeichen.
// Menge 267 = 128 Zeichen × 2 Stellen (256) + 6 Anfänge mit `$` oder `@` + 1 Name über 128 Zeichen + 4 Pfade.
func TestFlAufrufWortlaeufe(t *testing.T) {
	n := 0
	flAufruf := func(text string) {
		t.Helper()
		flFehler(t, text, flK(model.KompatFilter))
		n++
	}
	zeichen := flAufrufZeichen()
	if len(zeichen) != 128 {
		t.Fatalf("%d Zeichen, erwartet 128", len(zeichen))
	}
	for _, c := range zeichen {
		flAufruf("a" + c + "b(1)")
		flAufruf("a" + c + "(1)")
	}
	for _, s := range []string{"$x(1)", "@p(1)", "$(1)", "@(1)", "$it(1)", "$root/a(1)"} {
		flAufruf(s)
	}
	flAufruf(flWiederholt("a", 129) + "(1)")
	for _, s := range []string{"a//b(1)", "a/1(1)", "a..b(1)", "a/b.c(1)"} {
		flAufruf(s)
	}
	if n != 267 {
		t.Errorf("%d Fälle, erwartet 267", n)
	}
}

// flV4Form ist eine Form von V4 ohne Präfix: text, davon die ersten praefix Zeichen sind die Form, der Rest bleibt ungeprüft.
type flV4Form struct {
	text    string
	praefix int
}

// flV4Formen sind neun Formen: drei Datum und Zeitpunkt, drei Uhrzeit, drei Guid (Buchstabe am Anfang, groß, Ziffer am Anfang).
// Länge 10, 20, 25, 5, 8, 10, 36, 36, 36 (Summe 186 Stellen).
func flV4Formen() []flV4Form {
	return []flV4Form{
		{"2020-01-01", 10}, {"2020-01-01T00:00:00Z", 10}, {"2020-01-01T00:00:00+02:00", 10},
		{"12:30", 5}, {"12:30:00", 5}, {"12:30:00.5", 5},
		{guidBasis, 36}, {strings.ToUpper(guidBasis), 36}, {"00" + guidBasis[2:], 36},
	}
}

// flV4Urteil ist das Urteil über einen Lauf, der nicht mehr die Form von V4 hat: beginnt er mit einer ASCII-Ziffer, ist er ein
// Literal-Lauf und `literal`, sonst ein Wortlauf mit einem Zeichen außerhalb der Namen und `wert` (SPEC-173 Nr. 5, Absatz
// *Literal-Lauf ohne Präfix* d; SPEC-164 Nr. 4 b und c 4).
func flV4Urteil(s string) flV {
	if s[0] >= '0' && s[0] <= '9' {
		return flU(model.URLLiteral)
	}
	return flU(model.URLWert)
}

// flBeide prüft s als linken und als rechten Operanden.
func flBeide(t *testing.T, s string, want flV) {
	t.Helper()
	flFehler(t, "Id eq "+s, want)
	flFehler(t, s+" eq 1", want)
}

// TestFlV4FormPositionen (Zeile 19, SPEC-173 Nr. 5 *Literal-Lauf ohne Präfix* c und d): die Form von V4 gilt Position für Position:
// ist an einer Stelle der Form ein `x`, ist der Lauf `literal` (Ziffer am Anfang) bzw. `wert` (`x` am Anfang), hinter der Form
// bleibt ein `x` ungeprüft (`wert`). Dazu die Guid-Form in der Länge: jede Löschung (35 Zeichen), jedes Einfügen (37 Zeichen),
// 38 bis 40 Zeichen und jede Kürzung auf 30 bis 34 Zeichen sind keine Guid-Form.
// Menge 350 = 186 Stellen + 72 Löschungen + 74 Einfügungen + 6 Verlängerungen + 12 Kürzungen.
func TestFlV4FormPositionen(t *testing.T) {
	n := 0
	stellen := 0
	for _, f := range flV4Formen() {
		stellen += len(f.text)
		for p := range len(f.text) {
			s := f.text[:p] + "x" + f.text[p+1:]
			want := flU(model.URLWert)
			if p < f.praefix {
				want = flV4Urteil(s)
			}
			flBeide(t, s, want)
			n++
		}
	}
	if stellen != 186 {
		t.Fatalf("%d Stellen, erwartet 186", stellen)
	}
	loesch, einfuegen, verlaengert, gekuerzt := 0, 0, 0, 0
	for _, g := range []string{guidBasis, "00" + guidBasis[2:]} {
		for p := range len(g) {
			s := g[:p] + g[p+1:]
			flBeide(t, s, flV4Urteil(s))
			loesch++
		}
		for p := 0; p <= len(g); p++ {
			s := g[:p] + "a" + g[p:]
			flBeide(t, s, flV4Urteil(s))
			einfuegen++
		}
		for k := 2; k <= 4; k++ {
			s := g + flWiederholt("a", k)
			flBeide(t, s, flV4Urteil(s))
			verlaengert++
		}
		for k := 1; k <= 6; k++ {
			s := g[:len(g)-k]
			flBeide(t, s, flV4Urteil(s))
			gekuerzt++
		}
	}
	if loesch != 72 || einfuegen != 74 || verlaengert != 6 || gekuerzt != 12 {
		t.Errorf("%d, %d, %d, %d Fälle, erwartet 72, 74, 6, 12", loesch, einfuegen, verlaengert, gekuerzt)
	}
	if n+loesch+einfuegen+verlaengert+gekuerzt != 350 {
		t.Errorf("%d Fälle, erwartet 350", n+loesch+einfuegen+verlaengert+gekuerzt)
	}
}

// TestFlInfNanSchreibweise (Zeile 19, SPEC-164 Nr. 4 c 7): `INF` und `NaN` sind genau in dieser Schreibweise `literal_typ`; jede
// andere ASCII-Schreibweise von `inf` und `nan` ist ein Name, und ein Name mit Anhang ebenso.
// Menge 21 = 16 Schreibweisen (2 genau, 14 Namen) + 5 Namen mit Anhang oder Vorsatz, je links und rechts.
func TestFlInfNanSchreibweise(t *testing.T) {
	n, genau := 0, 0
	for _, w := range []string{"inf", "nan"} {
		for _, s := range flSchreibweisen(w) {
			if s == "INF" || s == "NaN" {
				flBeide(t, s, flK(model.KompatLiteralTyp))
				genau++
			} else {
				flGut(t, "Id eq "+s, flBin("eq", flEig("Id"), flEig(s)))
				flGut(t, s+" eq 1", flBin("eq", flEig(s), flZahl("1")))
			}
			n++
		}
	}
	for _, s := range []string{"INFx", "INF1", "xINF", "NaNN", "_NaN"} {
		flGut(t, "Id eq "+s, flBin("eq", flEig("Id"), flEig(s)))
		flGut(t, s+" eq 1", flBin("eq", flEig(s), flZahl("1")))
		n++
	}
	if n != 21 || genau != 2 {
		t.Errorf("%d Schreibweisen, %d genau; erwartet 21 und 2", n, genau)
	}
}

// TestFlSchreibweiseNichtASCII (Zeile 6, SPEC-164 Nr. 3): Schlüsselwörter gelten ohne Beachtung der ASCII-Schreibweise, und ein
// Zeichen jenseits von ASCII, das eine Kleinschreibung oder Faltung zu einem ASCII-Buchstaben hat (U+0130, U+0131, U+017F,
// U+212A), ist kein Buchstabe des Wortes: an jeder Stelle jedes der zwanzig Wörter ist der Lauf am Platz eines Operators `wert`
// und am Platz eines Operanden ein Name. Menge 464 = 58 Buchstaben × 4 Zeichen × 2 Plätze.
func TestFlSchreibweiseNichtASCII(t *testing.T) {
	woerter := []string{"eq", "ne", "gt", "ge", "lt", "le", "and", "or", "not", "true", "false", "null",
		"has", "in", "add", "sub", "mul", "div", "divby", "mod"}
	n, buchstaben := 0, 0
	for _, w := range woerter {
		buchstaben += len(w)
		for p := range len(w) {
			for _, r := range []int{0x130, 0x131, 0x17F, 0x212A} {
				s := w[:p] + flRune(r) + w[p+1:]
				flFehler(t, "Id "+s+" 1", flU(model.URLWert))
				flGut(t, s+" eq 1", flBin("eq", flEig(s), flZahl("1")))
				n += 2
			}
		}
	}
	if len(woerter) != 20 || buchstaben != 58 || n != 464 {
		t.Errorf("%d Wörter, %d Buchstaben, %d Fälle; erwartet 20, 58, 464", len(woerter), buchstaben, n)
	}
}

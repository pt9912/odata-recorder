package urlparser_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flLies liest einen Filter mit dem Leser der Grammatik.
func flLies(text string) (interface{ IstFilterAusdruck() }, error) {
	return urlparser.NewFilterLeser().LeseFilter(text)
}

// flUrsache liest text und liefert die Ursache des Fehlers (leer: gelesen). Der Fehler ist ein *URLFehler ohne Ort mit genau
// der Sentinel seiner Klasse, das Ergebnis bei Fehler nil (SPEC-164 Nr. 6).
func flUrsache(t *testing.T, text string) model.URLUrsache {
	t.Helper()
	a, err := flLies(text)
	if err == nil {
		if a == nil {
			t.Errorf("%q: weder Ausdruck noch Fehler", text)
		}
		return ""
	}
	var uf *model.URLFehler
	if !errors.As(err, &uf) || uf == nil || uf.Ort != "" || a != nil {
		t.Errorf("%q: Fehler %v, Ergebnis %v", text, err, a)
		return "?"
	}
	if errors.Is(err, model.ErrURLNichtUnterstuetzt) != klasse501(uf.Ursache) || errors.Is(err, model.ErrURLUngueltig) == klasse501(uf.Ursache) {
		t.Errorf("%q: Klasse von %s", text, uf.Ursache)
	}
	return uf.Ursache
}

// flFehler fordert die Ursache want.
func flFehler(t *testing.T, text string, want model.URLUrsache) {
	t.Helper()
	if got := flUrsache(t, text); got != want {
		t.Errorf("%q: Ursache %q, erwartet %q", text, got, want)
	}
}

// flGut fordert den Baum want.
func flGut(t *testing.T, text string, want interface{ IstFilterAusdruck() }) {
	t.Helper()
	a, err := flLies(text)
	if err != nil {
		t.Errorf("%q: Fehler %v", text, err)
		return
	}
	if !reflect.DeepEqual(a, want) {
		t.Errorf("%q:\n  ist   %#v\n  soll  %#v", text, a, want)
	}
}

// Knoten des erwarteten Baums, unabhängig vom Leser gebaut.
func eig(n string) interface{ IstFilterAusdruck() } { return model.FilterEigenschaft{Name: n} }
func zahl(s string) interface{ IstFilterAusdruck() } {
	return model.FilterLiteral{Literal: model.Literal{Art: model.LiteralZahl, Text: s}}
}
func zk(s string) interface{ IstFilterAusdruck() } {
	return model.FilterLiteral{Literal: model.Literal{Art: model.LiteralZeichenkette, Text: s}}
}
func wahr(s string) interface{ IstFilterAusdruck() } {
	return model.FilterLiteral{Literal: model.Literal{Art: model.LiteralWahrheitswert, Text: s}}
}
func nicht(x interface{ IstFilterAusdruck() }) interface{ IstFilterAusdruck() } {
	return model.FilterNicht{Operand: x}
}

func flNull() interface{ IstFilterAusdruck() } { return model.FilterNull{} }

// bin baut den Knoten eines der acht binären Operatoren.
func bin(op string, l, r interface{ IstFilterAusdruck() }) interface{ IstFilterAusdruck() } {
	switch op {
	case "and":
		return model.FilterVerknuepfung{Op: model.VerknuepfungUnd, Links: l, Rechts: r}
	case "or":
		return model.FilterVerknuepfung{Op: model.VerknuepfungOder, Links: l, Rechts: r}
	}
	return model.FilterVergleich{Op: model.VergleichOp(op), Links: l, Rechts: r}
}

// flStufe ist das Orakel der Rangfolge, 0 (schwach) bis 3.
func flStufe(op string) int {
	return map[string]int{"or": 0, "and": 1, "eq": 2, "ne": 2, "gt": 3, "ge": 3, "lt": 3, "le": 3}[op]
}

func flOps() []string { return []string{"eq", "ne", "gt", "ge", "lt", "le", "and", "or"} }

// flSchreibweisen liefert alle Schreibweisen von w, die sich nur in der ASCII-Groß- und Kleinschreibung unterscheiden.
func flSchreibweisen(w string) []string {
	var out []string
	for m := 0; m < 1<<len(w); m++ {
		b := []byte(w)
		for i := range b {
			if m>>i&1 == 1 {
				b[i] = b[i] &^ 0x20
			}
		}
		out = append(out, string(b))
	}
	return out
}

// flKorpus ist der Korpus der 40 Gutfälle (je Gruppe der Zeilen 2 bis 6, 17 und 18 mindestens fünf) mit dem erwarteten Baum.
func flKorpus() []struct {
	text string
	want interface{ IstFilterAusdruck() }
} {
	a, b, c, d := eig("a"), eig("b"), eig("c"), eig("d")
	flNull := flNull()
	return []struct {
		text string
		want interface{ IstFilterAusdruck() }
	}{
		{"Id eq 5", bin("eq", eig("Id"), zahl("5"))}, {"5 gt Id", bin("gt", zahl("5"), eig("Id"))},
		{"Name ne null", bin("ne", eig("Name"), flNull)}, {"null le Price", bin("le", flNull, eig("Price"))},
		{"Id ge Price", bin("ge", eig("Id"), eig("Price"))},
		{"a and b", bin("and", a, b)}, {"a or b", bin("or", a, b)}, {"not a", nicht(a)},
		{"not (a eq 1)", nicht(bin("eq", a, zahl("1")))},
		{"(a eq 1) and (b ne 2)", bin("and", bin("eq", a, zahl("1")), bin("ne", b, zahl("2")))},
		{"a or b and c", bin("or", a, bin("and", b, c))}, {"a and b eq c", bin("and", a, bin("eq", b, c))},
		{"a eq b gt c", bin("eq", a, bin("gt", b, c))}, {"not a gt 5", bin("gt", nicht(a), zahl("5"))},
		{"a or b or c and d", bin("or", bin("or", a, b), bin("and", c, d))},
		{"a and b and c", bin("and", bin("and", a, b), c)}, {"a or b or c or d", bin("or", bin("or", bin("or", a, b), c), d)},
		{"a eq 1 ne 2", bin("ne", bin("eq", a, zahl("1")), zahl("2"))}, {"a gt 1 le 2", bin("le", bin("gt", a, zahl("1")), zahl("2"))},
		{"a lt b lt c", bin("lt", bin("lt", a, b), c)},
		{"(a)", a}, {"((a))", a}, {"(a and b) and c", bin("and", bin("and", a, b), c)},
		{"a and (b and c)", bin("and", a, bin("and", b, c))}, {"((a or b)) and c", bin("and", bin("or", a, b), c)},
		{"null eq null", bin("eq", flNull, flNull)}, {"a eq NULL", bin("eq", a, flNull)}, {"Null ne a", bin("ne", flNull, a)},
		{"nullx eq 1", bin("eq", eig("nullx"), zahl("1"))}, {"_null eq xnull", bin("eq", eig("_null"), eig("xnull"))},
		{"Id", eig("Id")}, {"5", zahl("5")}, {"'x'", zk("'x'")}, {"null", flNull}, {"true", wahr("true")},
		{"not 5", nicht(zahl("5"))}, {"1 eq 'a'", bin("eq", zahl("1"), zk("'a'"))}, {"'a' eq 1", bin("eq", zk("'a'"), zahl("1"))},
		{"Id eq 1 and Name eq 'O''Hara'", bin("and", bin("eq", eig("Id"), zahl("1")), bin("eq", eig("Name"), zk("'O''Hara'")))},
		{"x eq TRUE", bin("eq", eig("x"), wahr("true"))},
	}
}

// flFehlerKorpus sind die 24 Fehlerfälle der Zeile 27 (Struktur) mit der Ursache `wert`.
func flFehlerKorpus() []string {
	return []string{"()", "(", ")", "(Id", "Id)", "((Id)", "(Id))", "Id eq", "Id eq 1 and", "Id eq 1 or", "not", "Id eq 1 not",
		"Id 1", "Id Id", "Id eq 1 2", "1 2", "(Id) (Id)", "Id,", ",Id", "Id eq 1,", "(Id,Id)", "Id eq ()", "not ()", "Id eq not"}
}

// flZaehleKnoten zählt die Knoten des Baums und prüft die Gestalt (SPEC-165 Nr. 1, 2): nur die sechs Wertetypen, kein Kind nil.
func flZaehleKnoten(t *testing.T, a interface{ IstFilterAusdruck() }) int {
	t.Helper()
	switch n := a.(type) {
	case model.FilterVergleich:
		return 1 + flZaehleKnoten(t, n.Links) + flZaehleKnoten(t, n.Rechts)
	case model.FilterVerknuepfung:
		return 1 + flZaehleKnoten(t, n.Links) + flZaehleKnoten(t, n.Rechts)
	case model.FilterNicht:
		return 1 + flZaehleKnoten(t, n.Operand)
	case model.FilterEigenschaft, model.FilterLiteral, model.FilterNull:
		return 1
	}
	t.Errorf("Knoten %#v ist kein Wertetyp des Filter-AST (nil oder Zeiger)", a)
	return 0
}

// flKodiere schreibt jedes Byte außer ASCII-Buchstaben und -Ziffern als %XX.
func flKodiere(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&15])
		}
	}
	return b.String()
}

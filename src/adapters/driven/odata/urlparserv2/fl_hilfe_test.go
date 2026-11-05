package urlparserv2_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flBaum ist ein Knoten des Filter-AST (model.FilterAusdruck); Rückgabewerte schreiben den Typ aus, weil der Linter `ireturn` einen
// benannten Alias meldet.
type flBaum = interface{ IstFilterAusdruck() }

// flV ist das Urteil über einen Text: gelesen (Nullwert), ein URLFehler der Klasse `ungueltig` mit seiner Ursache oder ein
// KompatFehler mit seiner Ursache.
type flV struct {
	u model.URLUrsache
	k model.KompatUrsache
}

func flU(u model.URLUrsache) flV    { return flV{u: u} }
func flK(k model.KompatUrsache) flV { return flV{k: k} }

func (v flV) String() string {
	switch {
	case v.u != "":
		return "URLFehler " + string(v.u)
	case v.k != "":
		return "KompatFehler " + string(v.k)
	}
	return "gelesen"
}

// flLies liest einen Filter mit dem Leser.
func flLies(text string) (interface{ IstFilterAusdruck() }, error) {
	return urlparserv2.NewFilterLeser().LeseFilter(text)
}

// flUrteil liest text und liefert Baum und Urteil. Es prüft die Form des Fehlers (SPEC-173 Nr. 5, *Der Leser*): genau der Typ
// `*URLFehler` mit Sentinel der Klasse `ungueltig` oder `*KompatFehler` mit `Code()` `ODR-E6006` und `ErrKompatibilitaet`, nie
// beide, `Ort` leer, das Ergebnis bei einem Fehler `nil`, bei keinem Fehler nicht `nil`.
func flUrteil(t *testing.T, text string) (interface{ IstFilterAusdruck() }, flV) {
	t.Helper()
	a, err := flLies(text)
	if err == nil {
		if a == nil {
			t.Errorf("%q: weder Ausdruck noch Fehler", text)
		}
		return a, flV{}
	}
	if a != nil {
		t.Errorf("%q: Ergebnis %v neben dem Fehler %v", text, a, err)
	}
	gestalt := func(ok bool, was string) {
		if !ok {
			t.Errorf("%q: Fehler %v (%T) ist %s", text, err, err, was)
		}
	}
	switch e := err.(type) {
	case *model.URLFehler:
		gestalt(e != nil && e.Ort == "" && e.Ursache.Gueltig() && !e.Ursache.NichtUnterstuetzt(), "keine Ursache der Klasse ungueltig ohne Ort")
		gestalt(errors.Is(err, model.ErrURLUngueltig) && !errors.Is(err, model.ErrURLNichtUnterstuetzt) && !errors.Is(err, model.ErrKompatibilitaet),
			"ohne genau den Sentinel der Klasse ungueltig")
		return a, flU(e.Ursache)
	case *model.KompatFehler:
		gestalt(e != nil && e.Ort == "" && (e.Ursache == model.KompatFilter || e.Ursache == model.KompatLiteralTyp), "keine Ursache filter oder literal_typ ohne Ort")
		gestalt(e.Code() == model.CodeE6006 && errors.Is(err, model.ErrKompatibilitaet) && !errors.Is(err, model.ErrURLUngueltig) &&
			!errors.Is(err, model.ErrURLNichtUnterstuetzt), "ohne ODR-E6006 und genau ErrKompatibilitaet")
		return a, flK(e.Ursache)
	}
	t.Errorf("%q: Fehler %v hat den Typ %T", text, err, err)
	return a, flV{u: "?"}
}

// flFehler fordert das Urteil want.
func flFehler(t *testing.T, text string, want flV) {
	t.Helper()
	if _, got := flUrteil(t, text); got != want {
		t.Errorf("%q: %v, erwartet %v", text, got, want)
	}
}

// flGut fordert den Baum want.
func flGut(t *testing.T, text string, want flBaum) {
	t.Helper()
	a, got := flUrteil(t, text)
	if got != (flV{}) {
		t.Errorf("%q: %v, erwartet den Baum %#v", text, got, want)
		return
	}
	if !reflect.DeepEqual(a, want) {
		t.Errorf("%q:\n  ist   %#v\n  soll  %#v", text, a, want)
	}
}

// Knoten des erwarteten Baums, unabhängig vom Leser gebaut.
func flEig(n string) interface{ IstFilterAusdruck() } { return model.FilterEigenschaft{Name: n} }
func flLit(a model.LiteralArt, text string) interface{ IstFilterAusdruck() } {
	return model.FilterLiteral{Literal: model.Literal{Art: a, Text: text}}
}
func flZahl(s string) interface{ IstFilterAusdruck() } { return flLit(model.LiteralZahl, s) }
func flZk(s string) interface{ IstFilterAusdruck() }   { return flLit(model.LiteralZeichenkette, s) }
func flWahr(s string) interface{ IstFilterAusdruck() } { return flLit(model.LiteralWahrheitswert, s) }
func flNicht(x flBaum) interface{ IstFilterAusdruck() } {
	return model.FilterNicht{Operand: x}
}
func flNull() interface{ IstFilterAusdruck() } { return model.FilterNull{} }

// flBin baut den Knoten eines der acht binären Operatoren.
func flBin(op string, l, r flBaum) interface{ IstFilterAusdruck() } {
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

// flStufenOps sind die vier binären Stufen von schwach nach stark; die fünfte Stufe ist `not`.
func flStufenOps() [][]string {
	return [][]string{{"or"}, {"and"}, {"eq", "ne"}, {"gt", "ge", "lt", "le"}}
}

// flSchreibweisen liefert alle Schreibweisen von w (Kleinbuchstaben), die sich nur in der ASCII-Schreibweise unterscheiden:
// eine je Bitmaske.
func flSchreibweisen(w string) []string {
	out := make([]string, 0, 1<<len(w))
	for m := range 1 << len(w) {
		out = append(out, schreibweise(w, m))
	}
	return out
}

// flRune schreibt ein Zeichen aus seinem Wert; Zeichen jenseits von ASCII stehen nie im Quelltext (AGENTS.md §3.18).
func flRune(r int) string { return string(rune(r)) }

type flFall struct {
	text string
	want flBaum
}

// flKorpus ist der Korpus der 40 Gutfälle (aus den Zeilen 1 bis 5, 20 und 21, je Gruppe mindestens fünf) mit dem erwarteten Baum.
func flKorpus() []flFall {
	a, b, c, d := flEig("a"), flEig("b"), flEig("c"), flEig("d")
	n := flNull()
	return []flFall{
		{"Id eq 5", flBin("eq", flEig("Id"), flZahl("5"))}, {"5 gt Id", flBin("gt", flZahl("5"), flEig("Id"))},
		{"Name ne null", flBin("ne", flEig("Name"), n)}, {"null le Price", flBin("le", n, flEig("Price"))},
		{"Id ge Price", flBin("ge", flEig("Id"), flEig("Price"))},
		{"a and b", flBin("and", a, b)}, {"a or b", flBin("or", a, b)}, {"not a", flNicht(a)},
		{"not (a eq 1)", flNicht(flBin("eq", a, flZahl("1")))},
		{"(a eq 1) and (b ne 2)", flBin("and", flBin("eq", a, flZahl("1")), flBin("ne", b, flZahl("2")))},
		{"a or b and c", flBin("or", a, flBin("and", b, c))}, {"a and b eq c", flBin("and", a, flBin("eq", b, c))},
		{"a eq b gt c", flBin("eq", a, flBin("gt", b, c))}, {"not a gt 5", flBin("gt", flNicht(a), flZahl("5"))},
		{"a or b or c and d", flBin("or", flBin("or", a, b), flBin("and", c, d))},
		{"a and b and c", flBin("and", flBin("and", a, b), c)}, {"a or b or c or d", flBin("or", flBin("or", flBin("or", a, b), c), d)},
		{"a eq 1 ne 2", flBin("ne", flBin("eq", a, flZahl("1")), flZahl("2"))}, {"a gt 1 le 2", flBin("le", flBin("gt", a, flZahl("1")), flZahl("2"))},
		{"a lt b lt c", flBin("lt", flBin("lt", a, b), c)},
		{"(a)", a}, {"((a))", a}, {"(a and b) and c", flBin("and", flBin("and", a, b), c)},
		{"a and (b and c)", flBin("and", a, flBin("and", b, c))}, {"((a or b)) and c", flBin("and", flBin("or", a, b), c)},
		{"null eq null", flBin("eq", n, n)}, {"a eq NULL", flBin("eq", a, n)}, {"Null ne a", flBin("ne", n, a)},
		{"nullx eq 1", flBin("eq", flEig("nullx"), flZahl("1"))}, {"_null eq xnull", flBin("eq", flEig("_null"), flEig("xnull"))},
		{"Id", flEig("Id")}, {"5", flZahl("5")}, {"'x'", flZk("'x'")}, {"null", n}, {"true", flWahr("true")},
		{"not 5", flNicht(flZahl("5"))}, {"1 eq 'a'", flBin("eq", flZahl("1"), flZk("'a'"))}, {"'a' eq 1", flBin("eq", flZk("'a'"), flZahl("1"))},
		{"Id eq 1 and Name eq 'O''Hara'", flBin("and", flBin("eq", flEig("Id"), flZahl("1")), flBin("eq", flEig("Name"), flZk("'O''Hara'")))},
		{"x eq TRUE", flBin("eq", flEig("x"), flWahr("true"))},
	}
}

// flFehlerKorpus sind die 24 Fehlerfälle der Zeile 30 (Struktur) mit der Ursache `wert`.
func flFehlerKorpus() []string {
	return []string{"()", "(", ")", "(Id", "Id)", "((Id)", "(Id))", "Id eq", "Id eq 1 and", "Id eq 1 or", "not", "Id eq 1 not",
		"Id 1", "Id Id", "Id eq 1 2", "1 2", "(Id) (Id)", "Id,", ",Id", "Id eq 1,", "(Id,Id)", "Id eq ()", "not ()", "Id eq not"}
}

// flZaehleKnoten zählt die Knoten des Baums und prüft die Gestalt (SPEC-165 Nr. 1, 2): nur die sechs Wertetypen, kein Kind nil.
func flZaehleKnoten(t *testing.T, a flBaum) int {
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

// flWiederholt bildet n Mal einheit.
func flWiederholt(einheit string, n int) string { return strings.Repeat(einheit, n) }

package urlparserv2_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flMitLeser ist ein Parser mit dem Leser von `$filter`.
func flMitLeser() *urlparserv2.Parser {
	return urlparserv2.New(urlparserv2.WithFilter(urlparserv2.NewFilterLeser()))
}

// TestFlPfad (Zeile 22): ein Pfad ist `filter` (Kompatibilität), kein `wert_nicht_unterstuetzt` und nie ein URLFehler; ein
// fehlerhafter Pfad ist `wert`; über `ParseURL` bleibt die Ursache `filter`, in `$select` ist derselbe Pfad `option`. Menge 10.
func TestFlPfad(t *testing.T) {
	for _, s := range []string{"a/b", "a/b/c", "a/b/c/d", "a/b eq 1", "1 eq a/b"} {
		flFehler(t, s, flK(model.KompatFilter))
	}
	for _, s := range []string{"a/", "/a", "a//b", "a/1", "a/b.c"} {
		flFehler(t, s, flU(model.URLWert))
	}
	p := flMitLeser()
	pruefe(t, p, "/S", "$filter=a/b", kmp(model.KompatFilter, "query/1"))
	pruefe(t, p, "/S", "$select=a/b", kmp(model.KompatOption, "query/1"))
	pruefe(t, p, "/S", "$orderby=a/b", kmp(model.KompatOption, "query/1"))
}

// flBibliothek sind die 22 Funktionen der Menge außerhalb des Umfangs (SPEC-170 Nr. 3), aus dem Wortlaut der Spezifikation.
func flBibliothek() []string {
	return strings.Fields("substringof endswith startswith length indexof replace substring tolower toupper trim concat year month " +
		"day hour minute second round floor ceiling isof cast")
}

// flV4Funktionen sind die fünf Namen von V4 und `geo.distance`, die V2 nicht kennt (SPEC-173 Nr. 5).
func flV4Funktionen() []string {
	return []string{"contains", "matchesPattern", "in", "any", "all", "geo.distance"}
}

// TestFlAufruf (Zeile 23): jeder Name unmittelbar vor `(` ist ein Aufruf, `filter`, auch ein erfundener; mit Leerraum davor ist es
// keiner. Menge 168 = 33 Namen × 5 Formen + 3.
func TestFlAufruf(t *testing.T) {
	namen := append(append(flBibliothek(), flV4Funktionen()...), "foo", "Ns.foo", "a/any", "true", "x1")
	if len(flBibliothek()) != 22 || len(namen) != 33 {
		t.Fatalf("%d Funktionen, %d Namen", len(flBibliothek()), len(namen))
	}
	n := 0
	for _, name := range namen {
		for _, form := range []string{"()", "(Id)", "(Id,1)", "(f(Id))", "(Id"} {
			flFehler(t, name+form, flK(model.KompatFilter))
			n++
		}
	}
	flFehler(t, "Name (Id)", flU(model.URLWert))
	flGut(t, "not(Id)", flNicht(flEig("Id")))
	flGut(t, "not (Id)", flNicht(flEig("Id")))
	if n != 165 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestFlNichtUnterstuetzt (Zeile 24): Operatoren und weitere Konstrukte, die V2 nicht kennt, sind `filter`. Menge 49 = 24
// (8 Operatorwörter × 3 Einbettungen) + 25.
func TestFlNichtUnterstuetzt(t *testing.T) {
	f, w, l := flK(model.KompatFilter), flU(model.URLWert), flU(model.URLLiteral)
	n := 0
	for _, op := range []string{"add", "sub", "mul", "div", "mod", "has", "in", "divby"} {
		for _, text := range []string{"1 " + op + " 2 eq 3", "Id " + op + " 1", "(Id " + op + " 1)"} {
			flFehler(t, text, f)
			n++
		}
	}
	faelle := []struct {
		text string
		v    flV
	}{
		{"-Id", f}, {"- 5", f}, {"-(5)", f}, {"--5", f}, {"-INFO", f},
		{"$it", f}, {"$root", f}, {"$this", f}, {"$it/Id", f}, {"$this/Id", f}, {"$root/Id", f}, {"$x", w},
		{"@p", f}, {"@", w}, {"@1", w},
		{"Ns.Typ", f}, {"a.b", f}, {"a.b.c", f}, {"a.", w}, {".a", l}, {"a..b", w},
		{"Id in (1,2)", f}, {"Id in [1]", f}, {"Items/any(x: x eq 1)", f}, {"isof('Ns.T')", f},
	}
	for _, c := range faelle {
		flFehler(t, c.text, c.v)
		n++
	}
	if n != 49 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestFlAusserhalbAllaussage (Zeile 35): jeder Operator und jede Funktion außerhalb des Umfangs ist `filter`; die Listen stehen
// im Test aus dem Wortlaut von SPEC-170 Nr. 3, nicht aus dem Code. Menge 99 = 33 Namen (5 Operatoren, 28 Funktionen) × 3
// Einbettungen (Wurzel, Operand eines Vergleichs, hinter `and`).
func TestFlAusserhalbAllaussage(t *testing.T) {
	operatoren := []string{"add", "sub", "mul", "div", "mod"}
	funktionen := append(flBibliothek(), flV4Funktionen()...)
	if len(operatoren) != 5 || len(funktionen) != 28 {
		t.Fatalf("%d Operatoren, %d Funktionen", len(operatoren), len(funktionen))
	}
	n := 0
	for _, op := range operatoren {
		for _, text := range []string{"Id " + op + " 1", "(Id " + op + " 1) eq 2", "Id eq 1 and Id " + op + " 2"} {
			flFehler(t, text, flK(model.KompatFilter))
			n++
		}
	}
	for _, f := range funktionen {
		for _, text := range []string{f + "(Id)", "Id eq " + f + "(Id)", "Id eq 1 and " + f + "(Id)"} {
			flFehler(t, text, flK(model.KompatFilter))
			n++
		}
	}
	if n != 99 {
		t.Fatalf("%d Fälle", n)
	}
}

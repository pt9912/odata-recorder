package urlparser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const flGUID = "0a1b2c3d-0000-1111-2222-333344445555"

// flLiteralGruppen sind die Formen der Literale in sieben Gruppen (Zeichenkette, Wahrheitswert, Zahl, Guid, Datum, Zeitpunkt,
// fremder Typ), nach den Zeilen 32 bis 38 des Pfad-Parsers gebildet.
func flLiteralGruppen() [][]string {
	zeichenketten := []string{"''", "'a'", "'O''Hara'", "''''", "'a,b'", "'a)b'", "'a(b'", "'a=b'", "'a%2Fb'", "'\\'",
		"'" + flRune(0xE4) + "'", "'%C3%A4'", "'%27%27'", "'x\ny'", "'a b'", "'a", "'a'b", "'", "a'", "'a''", "'a'x'"}
	wahrheit := append(flSchreibweisen("true"), flSchreibweisen("false")...)
	wahrheit = append(wahrheit, "1", "0", "t", "tru", "truee", "yes", "null", "NULL", "nUll", "inf", "nan")
	zahlen := []string{"0", "-0", "00", "007", "42", "-42", "1.5", "1.50", "0.0", "-0.0", "0.10", strings.Repeat("9", 38),
		"+1", ".5", "5.", "-", "1.2.3", "42L", "1.5M", "1d", "0x10", flRune(0xFF15), flRune(0x663), "--1", "(1)",
		"1e3", "1E3", "1.5e-2", "INF", "-INF", "NaN"}
	guids := []string{flGUID, strings.ToUpper(flGUID), "0a1B2c3D-0000-1111-2222-333344445555", "{" + flGUID + "}", strings.ReplaceAll(flGUID, "-", ""),
		"0a1b2c3-0000-1111-2222-3333444455556", "0a1b2c3d-0000-1111-2222-33334444555", "0a1b2c3d-0000-1111-2222-33334444555g",
		"0a1b2c3d-0000-1111-2222", "urn:uuid:" + flGUID, "a" + flGUID[1:], "A" + flGUID[1:], "f" + strings.ToUpper(flGUID)[1:]}
	daten := []string{"0000-01-01", "0001-01-01", "9999-12-31", "10000-01-01", "2020-00-01", "2020-01-01", "2020-12-01", "2020-13-01",
		"2023-02-28", "2023-02-29", "2024-02-29", "2024-02-30", "1900-02-29", "2000-02-29", "2024-04-30", "2024-04-31", "2024-01-31",
		"2024-01-32", "2024-01-00", "2024-1-5", "20240105", "+2024-01-05"}
	zeitpunkte := []string{"2020-01-01T23:59:59Z", "2020-01-01T24:00:00Z", "2020-01-01T10:59:00Z", "2020-01-01T10:60:00Z",
		"2020-01-01T10:00:59Z", "2020-01-01T10:00:60Z", "2020-01-01T10:00:00+23:59", "2020-01-01T10:00:00+24:00",
		"2020-01-01T10:00:00+00:60", "2020-01-01T10:00:00.1Z", "2020-01-01T10:00:00.123456789012Z", "2020-01-01T10:00:00.1234567890123Z",
		"2020-01-01T10:00:00.000Z", "2020-01-01T10:00Z", "2020-01-01T10:00:00Z", "2020-01-01T10:00:00+00:00", "2020-01-01T10:00:00+01:00",
		"2020-01-01t10:00:00Z", "2020-01-01T10:00:00z", "2020-01-01 10:00:00Z", "0001-01-01T00:00:00+01:00",
		"9999-12-31T23:59:59-01:00", "2020-01-01T10:00+02:00", "2020-01-01T10:00:00-02:30"}
	fremd := []string{"binary'AA=='", "X'AA'", "duration'P1D'", "geography'POINT(1 2)'", "geometry'POINT(1 2)'", "Ns.Farbe'Rot'",
		"datetime'2020-01-01T00:00:00'", "12:30", "12:30:00", "12:30:00.5", "binary'AA"}
	return [][]string{zeichenketten, wahrheit, zahlen, guids, daten, zeitpunkte, fremd}
}

// flAusnahmen15 sind die Formen, die der Filter anders liest als den Schlüssel, mit dem erwarteten Ergebnis des Filters.
// Grund je Gruppe: ein nacktes Wort ist im Filter ein Name oder `null`, `-` und `--1` beginnen ein Minus, Zeichen jenseits von
// ASCII und `{`, `urn:` beginnen einen Lauf ohne Form eines Literals, ein Leerzeichen trennt Token, `(1)` ist eine Gruppe, und
// ein Lauf mit einem Hex-Buchstaben am Anfang ist ein Wortlauf (SPEC-164 Nr. 4 b, c).
func flAusnahmen15() (map[string]model.URLUrsache, map[string]interface{ IstFilterAusdruck() }) {
	fehler := map[string]model.URLUrsache{"-": model.URLFilter, "--1": model.URLFilter, flRune(0xFF15): model.URLWert,
		flRune(0x663): model.URLWert, "{" + flGUID + "}": model.URLWert, "urn:uuid:" + flGUID: model.URLWert,
		"a" + flGUID[1:]: model.URLWert, "A" + flGUID[1:]: model.URLWert, "f" + strings.ToUpper(flGUID)[1:]: model.URLWert,
		"2020-01-01 10:00:00Z": model.URLWert}
	baum := map[string]interface{ IstFilterAusdruck() }{"(1)": zahl("1"), "null": flNull(), "NULL": flNull(), "nUll": flNull()}
	for _, w := range []string{"t", "tru", "truee", "yes", "inf", "nan"} {
		baum[w] = eig(w)
	}
	return fehler, baum
}

// flSchluessel liest lit als Schlüsselliteral `Id=<lit>` in `/S(…)` und liefert Literal oder Ursache.
func flSchluessel(t *testing.T, lit string) (model.Literal, model.URLUrsache) {
	t.Helper()
	u, err := urlparser.New().ParseURL("/S(Id="+flKodiere(lit)+")", "")
	if err == nil {
		return u.Ressource.Schluessel[0].Literal, ""
	}
	var uf *model.URLFehler
	if !errors.As(err, &uf) {
		t.Fatalf("%q: %v", lit, err)
	}
	return model.Literal{}, uf.Ursache
}

// TestLiteraleWieSchluessel (Zeile 15): als `Id eq <Literal>` ergibt jede Form dieselbe Art, denselben Text oder dieselbe
// Ursache wie in `/S(Id=<Literal>)`; die Ausnahmen (20) stehen mit erwartetem Ergebnis in flAusnahmen15. Menge 181 = 21 + 59 +
// 31 + 13 + 22 + 24 + 11.
func TestLiteraleWieSchluessel(t *testing.T) {
	ausFehler, ausBaum := flAusnahmen15()
	n, ausn := 0, 0
	for _, g := range flLiteralGruppen() {
		for _, lit := range g {
			n++
			if u, ok := ausFehler[lit]; ok {
				ausn++
				flFehler(t, "Id eq "+lit, u)
				continue
			}
			if b, ok := ausBaum[lit]; ok {
				ausn++
				flGut(t, "Id eq "+lit, bin("eq", eig("Id"), b))
				continue
			}
			key, u := flSchluessel(t, lit)
			if u != "" {
				flFehler(t, "Id eq "+lit, u)
			} else {
				flGut(t, "Id eq "+lit, bin("eq", eig("Id"), model.FilterLiteral{Literal: key}))
			}
		}
	}
	if n != 181 || ausn != len(ausFehler)+len(ausBaum) {
		t.Fatalf("%d Formen, %d Ausnahmen", n, ausn)
	}
}

// TestFilterDekodierung (Zeile 26): der Leser sieht den einmal dekodierten Text. Menge 22 = 6 Zeichen × 3 Schreibweisen (roh,
// %XX, %xx) + `%2527` + 3 (`+` bleibt `+`).
func TestFilterDekodierung(t *testing.T) {
	p := urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser()))
	n := 0
	for _, c := range []byte{'\'', '(', ')', ' ', ',', '/'} {
		var want interface{ IstFilterAusdruck() }
		roh, gross, klein := string([]byte{c}), "%"+strings.ToUpper(flHex(c)), "%"+strings.ToLower(flHex(c))
		inhalt := string([]byte{c})
		if c == '\'' {
			inhalt, roh, gross, klein = "''", "''", gross+gross, klein+klein
		}
		want = bin("eq", eig("Name"), zk("'a"+inhalt+"b'"))
		for _, schreibweise := range []string{roh, gross, klein} {
			u, err := p.ParseURL("/S", "$filter=Name eq 'a"+schreibweise+"b'")
			if err != nil || u.Optionen.Filter == nil || !gleich(u.Optionen.Filter.Ausdruck, want) || u.Optionen.Filter.Text != "Name eq 'a"+inhalt+"b'" {
				t.Errorf("%q: %+v %v", schreibweise, u.Optionen.Filter, err)
			}
			n++
		}
	}
	u, err := p.ParseURL("/S", "$filter=Name eq 'a%2527b'")
	if err != nil || !gleich(u.Optionen.Filter.Ausdruck, bin("eq", eig("Name"), zk("'a%27b'"))) {
		t.Errorf("%%2527: %+v %v", u.Optionen.Filter, err)
	}
	_, err = p.ParseURL("/S", "$filter=Id+eq+1")
	var uf *model.URLFehler
	if !errors.As(err, &uf) || uf.Ursache != model.URLWert {
		t.Errorf("Id+eq+1: %v", err)
	}
	u, err = p.ParseURL("/S", "$filter=Name eq 'a+b'")
	if err != nil || !gleich(u.Optionen.Filter.Ausdruck, bin("eq", eig("Name"), zk("'a+b'"))) {
		t.Errorf("'a+b': %+v %v", u.Optionen.Filter, err)
	}
	zeit := bin("eq", eig("T"), model.FilterLiteral{Literal: model.Literal{Art: model.LiteralZeitpunkt, Text: "2020-01-01T08:00:00Z"}})
	for _, w := range []string{"+", "%2B"} {
		u, err = p.ParseURL("/S", "$filter=T eq 2020-01-01T10:00"+w+"02:00")
		if err != nil || !gleich(u.Optionen.Filter.Ausdruck, zeit) {
			t.Errorf("Versatz %q: %+v %v", w, u.Optionen.Filter, err)
		}
	}
	if n != 18 {
		t.Fatalf("%d Fälle", n)
	}
}

func flHex(c byte) string {
	const hex = "0123456789abcdef"
	return string([]byte{hex[c>>4], hex[c&15]})
}

package urlparser_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Hilfen der Tests von `$expand` (SPEC-262). Jeder Wert jenseits von ASCII entsteht aus Bytewerten (`AGENTS.md` §3.18); das
// Orakel liest keinen Code des Lesers, es ist eine eigene Tabelle aus dem Wortlaut von SPEC-262 Nr. 3 und 4.

func exParser() *urlparser.Parser { return urlparser.New(urlparser.WithExpand()) }

// Zeichen jenseits von ASCII, als Bytewerte in der Schreibweise `\x`.
const (
	exE     = "\xc3\xa9"     // é: Buchstabe (Lu/Ll)
	exNBSP  = "\xc2\xa0"     // NBSP: kein Buchstabe, keine Ziffer
	exBOM   = "\xef\xbb\xbf" // BOM: weder noch
	exShe   = "\xd0\x96"     // Ж: Buchstabe
	exHan   = "\xe4\xb8\xad" // 中: Buchstabe (Lo)
	exArab3 = "\xd9\xa3"     // ٣: Ziffer (Nd)
	exHoch2 = "\xc2\xb2"     // ²: kein Nd (No)
)

// exBuchstaben und exZiffern sind die Zeichenklassen des Orakels (SPEC-125 Nr. 8 b), als endliche Tabelle der Zeichen, die die
// Generatoren verwenden.
const (
	exBuchstaben = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_" + exE + exShe + exHan
	exZiffern    = "0123456789" + exArab3
)

// exSoll ist die erwartete Antwort: der Knoten mit den Namen (u leer) oder die Ursache mit dem Ort.
type exSoll struct {
	namen []string
	u     model.URLUrsache
	ort   string
}

func exErg(namen ...string) exSoll                 { return exSoll{namen: namen} }
func exFehl(u model.URLUrsache, ort string) exSoll { return exSoll{u: u, ort: ort} }
func exQ(roh string) string                        { return "$expand=" + roh }
func exWert() exSoll                               { return exFehl(model.URLWert, "query/1") }
func exNichtUnterstuetzt() exSoll                  { return exFehl(model.URLWertNichtUnterstuetzt, "query/1") }
func exZuViele() exSoll                            { return exFehl(model.URLZuViele, "query/1") }

// exNamen sind n verschiedene Namen `<praefix>0` bis `<praefix>n-1`.
func exNamen(praefix string, n int) []string {
	namen := make([]string, n)
	for i := range namen {
		namen[i] = fmt.Sprintf("%s%d", praefix, i)
	}
	return namen
}

// exKnoten ist das erwartete Ergebnis von `/S` mit dem Knoten.
func exKnoten(namen ...string) *model.ODataURL {
	el := make([]model.ExpandElement, len(namen))
	for i, n := range namen {
		el[i] = model.ExpandElement{Name: n}
	}
	return &model.ODataURL{
		Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: "S"},
		Optionen:  model.Optionen{Expand: &model.ExpandOption{Elemente: el}},
	}
}

// exPruefe liest `/S` mit der Query und hält das Ergebnis gegen die Erwartung.
func exPruefe(t *testing.T, p *urlparser.Parser, query string, s exSoll) {
	t.Helper()
	f := fall{pfad: "/S", query: query, u: s.u, ort: s.ort}
	if s.u == "" {
		f.want = exKnoten(s.namen...)
	}
	pruefeFall(t, p, f)
}

// exSymbol ist ein Zeichen des Alphabets der Grammatik: die rohe Schreibweise in der URL und der dekodierte Text.
type exSymbol struct{ roh, dek string }

// exAlphabet sind die 17 Zeichen der Prüfgrenze (SPEC-262 *Ungeprüft*): Buchstabe, Ziffer, `_`, die Zeichen der Grammatik
// (`,`, `/`, `*`, `(`, `)`, `$`, `.`, `;`, `=`), zwei Escapes (`%20`, `%2C`) und zwei Zeichen jenseits von ASCII.
func exAlphabet() []exSymbol {
	s := []exSymbol{{"a", "a"}, {"Z", "Z"}, {"_", "_"}, {"0", "0"}, {",", ","}, {"/", "/"}, {"*", "*"}, {"(", "("}, {")", ")"},
		{"$", "$"}, {".", "."}, {";", ";"}, {"=", "="}, {"%20", " "}, {"%2C", ","}, {exE, exE}, {exNBSP, exNBSP}}
	return s
}

// exFolgen sind alle Folgen der Länge 1 bis n über dem Alphabet, nach Länge geordnet.
func exFolgen(n int) []exSymbol {
	alpha := exAlphabet()
	alle := []exSymbol{}
	vorige := []exSymbol{{}}
	for range n {
		var naechste []exSymbol
		for _, v := range vorige {
			for _, a := range alpha {
				naechste = append(naechste, exSymbol{v.roh + a.roh, v.dek + a.dek})
			}
		}
		alle = append(alle, naechste...)
		vorige = naechste
	}
	return alle
}

// exIstName ist das Orakel eines einfachen Namens: 1 bis 128 Zeichen, Buchstabe oder `_`, dann Buchstaben, Ziffern, `_`.
func exIstName(e string) bool {
	zeichen := []rune(e)
	if len(zeichen) < 1 || len(zeichen) > 128 {
		return false
	}
	for i, r := range zeichen {
		switch {
		case strings.ContainsRune(exBuchstaben, r):
		case i > 0 && strings.ContainsRune(exZiffern, r):
		default:
			return false
		}
	}
	return true
}

// exSegmentOrakel meldet ein zulässiges Segment eines Pfads: Name, qualifizierter Name oder `$ref`, `$count`, `$value`.
func exSegmentOrakel(s string) bool {
	if s == "$ref" || s == "$count" || s == "$value" {
		return true
	}
	for _, teil := range strings.Split(s, ".") {
		if !exIstName(teil) {
			return false
		}
	}
	return true
}

// exIstPfad ist das Orakel eines Pfads: mindestens ein `/` und lauter zulässige Segmente.
func exIstPfad(e string) bool {
	if !strings.Contains(e, "/") {
		return false
	}
	for _, s := range strings.Split(e, "/") {
		if !exSegmentOrakel(s) {
			return false
		}
	}
	return true
}

// exOrakel ist die erwartete Antwort auf den dekodierten Wert, nach SPEC-262 Nr. 3 und 4; Ort `query/1`.
func exOrakel(wert string) exSoll {
	switch {
	case wert == "":
		return exWert()
	case strings.Contains(wert, "("):
		return exNichtUnterstuetzt()
	case strings.Count(wert, ",") >= 32:
		return exZuViele()
	}
	var namen []string
	ungueltig, nichtUnterstuetzt := false, false
	elemente := strings.Split(wert, ",")
	for i, e := range elemente {
		switch {
		case slices.Contains(elemente[:i], e):
			ungueltig = true
		case exIstName(e):
			namen = append(namen, e)
		case e == "*" || exIstPfad(e):
			nichtUnterstuetzt = true
		default:
			ungueltig = true
		}
	}
	switch {
	case ungueltig:
		return exWert()
	case nichtUnterstuetzt:
		return exNichtUnterstuetzt()
	}
	return exErg(namen...)
}

// exHex schreibt jedes Byte von s als `%XX`, in Groß- oder Kleinschreibung der Hex-Ziffern.
func exHex(s string, gross bool) string {
	var b strings.Builder
	for i := range len(s) {
		if gross {
			fmt.Fprintf(&b, "%%%02X", s[i])
		} else {
			fmt.Fprintf(&b, "%%%02x", s[i])
		}
	}
	return b.String()
}

// exZaehlt hält die Zahl der Elemente einer Menge gegen die Zahl, die der Plan nennt.
func exZaehlt(t *testing.T, name string, ist, soll int) {
	t.Helper()
	if ist != soll {
		t.Fatalf("%s: Menge von %d Elementen, erwartet %d", name, ist, soll)
	}
}

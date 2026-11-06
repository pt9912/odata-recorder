package jsoncodec_test

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eChar liefert das Zeichen mit dem Wert b (0 bis 255) als Text: das Byte für b < 0x80, sonst das Zeichen U+00bb in UTF-8.
func eChar(b int) string { return string(rune(b)) }

// eJSONZeichen schreibt das Zeichen c als Inhalt eines JSON-Strings: eine Escape-Folge aus Bytewerten (AGENTS.md §3.18), Hex-Ziffern in
// der Schreibweise, die der Index wählt (klein, groß, gemischt).
func eJSONZeichen(c string, i int) string {
	r := []rune(c)[0]
	h := fmt.Sprintf("%04x", r)
	switch i % 3 {
	case 1:
		h = strings.ToUpper(h)
	case 2:
		h = h[:2] + strings.ToUpper(h[2:])
	}
	return esc(h)
}

// eErsetzt liefert die Texte, die entstehen, wenn das Zeichen an der Stelle pos von s durch jedes der 256 Zeichen ersetzt wird.
func eErsetzt(s string, pos int) []string {
	out := make([]string, 256)
	for b := range 256 {
		out[b] = s[:pos] + eChar(b) + s[pos+1:]
	}
	return out
}

// eErsetztRoh wie eErsetzt, aber mit dem rohen Byte b an der Stelle pos (für den Inhalt außerhalb eines Strings).
func eErsetztRoh(s string, pos int) []string {
	out := make([]string, 256)
	for b := range 256 {
		out[b] = s[:pos] + string([]byte{byte(b)}) + s[pos+1:]
	}
	return out
}

// eJSONText schreibt s als JSON-String: Ziffern, Buchstaben und `-:.+` roh, jedes andere Zeichen als Escape-Folge.
func eJSONText(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i, r := range []rune(s) {
		if r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == '-' || r == ':' || r == '.' || r == '+' {
			b.WriteRune(r)
		} else {
			b.WriteString(eJSONZeichen(string(r), i))
		}
	}
	b.WriteByte('"')
	return b.String()
}

// ePruefeTexte liest jeden Text als String-Wert der Property name von W (Art PATCH) und vergleicht mit dem Orakel: soll liefert den
// kanonischen Text oder "" für `wert`. Gibt die Zahl der Texte zurück.
func ePruefeTexte(t *testing.T, name string, texte []string, soll func(string) (string, bool)) int {
	t.Helper()
	for _, s := range texte {
		kanon, gueltig := soll(s)
		f := bad(eJSONText(s), model.EingabeWert)
		if gueltig {
			f = ok(eJSONText(s), kanon)
		}
		eWertPruefen(t, name, f, model.EingabePatch)
	}
	return len(texte)
}

var eGuidRE = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// TestEingabeGuidStellen hält Zeile 26 für Guid: jedes Zeichen an jeder der 36 Stellen und jede Länge von 0 bis 40. Menge 9257.
func TestEingabeGuidStellen(t *testing.T) {
	const basis = "0f8fad5b-d9cb-469f-a165-70867728950e"
	var texte []string
	for pos := range len(basis) {
		texte = append(texte, eErsetzt(basis, pos)...)
	}
	for l := 0; l <= 40; l++ {
		texte = append(texte, (basis + "0000")[:l])
	}
	n := ePruefeTexte(t, "G", texte, func(s string) (string, bool) { return strings.ToLower(s), eGuidRE.MatchString(s) })
	if n != 9257 {
		t.Errorf("Menge %d, erwartet 9257", n)
	}
}

var eDatumRE = regexp.MustCompile(`^([0-9]{4})-([0-9]{2})-([0-9]{2})$`)

// eKalender prüft Jahr 1 bis 9999, Monat und Tag nach dem gregorianischen Kalender.
func eKalender(j, m, d int) bool {
	if j < 1 || j > 9999 || m < 1 || m > 12 || d < 1 {
		return false
	}
	return d <= time.Date(j, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func eAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// TestEingabeDatumStellen hält Zeile 26 für Datum: Menge 2573 = 10 Stellen × 256 + Längen 0 bis 12.
func TestEingabeDatumStellen(t *testing.T) {
	// Die Basis liegt an den Grenzen (Schalttag, Monat 2), damit ein Zeichen allein Jahr 2021, Monat 13 oder Tag 39 erreicht.
	const basis = "2020-02-29"
	var texte []string
	for pos := range len(basis) {
		texte = append(texte, eErsetzt(basis, pos)...)
	}
	for l := 0; l <= 12; l++ {
		texte = append(texte, (basis + "01")[:l])
	}
	n := ePruefeTexte(t, "D", texte, func(s string) (string, bool) {
		m := eDatumRE.FindStringSubmatch(s)
		return s, m != nil && eKalender(eAtoi(m[1]), eAtoi(m[2]), eAtoi(m[3]))
	})
	if n != 2573 {
		t.Errorf("Menge %d, erwartet 2573", n)
	}
}

var eZeitRE = regexp.MustCompile(`^([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2})(?::([0-9]{2})(?:\.([0-9]{1,12}))?)?(?:Z|([+-])([0-9]{2}):([0-9]{2}))$`)

// eZeitOrakel rechnet den Zeitpunkt nach SPEC-126 Nr. 10 d in UTC; der Bruchteil bleibt Text und verliert die Nullen am Ende.
func eZeitOrakel(s string) (string, bool) {
	m := eZeitRE.FindStringSubmatch(s)
	if m == nil || !eKalender(eAtoi(m[1]), eAtoi(m[2]), eAtoi(m[3])) {
		return "", false
	}
	h, mi, sek := eAtoi(m[4]), eAtoi(m[5]), eAtoi(m[6])
	if h > 23 || mi > 59 || sek > 59 {
		return "", false
	}
	versatz := 0
	if m[8] != "" {
		oh, om := eAtoi(m[9]), eAtoi(m[10])
		if oh > 23 || om > 59 {
			return "", false
		}
		versatz = oh*60 + om
		if m[8] == "-" {
			versatz = -versatz
		}
	}
	t := time.Date(eAtoi(m[1]), time.Month(eAtoi(m[2])), eAtoi(m[3]), h, mi, sek, 0, time.UTC).Add(-time.Duration(versatz) * time.Minute)
	if t.Year() < 1 || t.Year() > 9999 {
		return "", false
	}
	text := t.Format("2006-01-02T15:04:05")
	if bruch := strings.TrimRight(m[7], "0"); bruch != "" {
		text += "." + bruch
	}
	return text + "Z", true
}

// TestEingabeZeitpunktStellen hält Zeile 26 für Zeitpunkte: Menge 9758 = Grundform 5120 + Bruchteil 3102 + Versatz 1536.
func TestEingabeZeitpunktStellen(t *testing.T) {
	// Die Basis liegt an den Grenzen (23:59:59, Schalttag), damit ein Zeichen allein Stunde 24, Minute 60 oder Sekunde 60 erreicht.
	grund := "2020-02-29T23:59:59Z"
	var texte []string
	for pos := range len(grund) {
		texte = append(texte, eErsetzt(grund, pos)...)
	}
	if len(texte) != 5120 {
		t.Fatalf("Grundform %d", len(texte))
	}
	for l := 0; l <= 14; l++ {
		texte = append(texte, "2020-01-01T00:00:00."+strings.Repeat("1", l)+"Z", "2020-01-01T00:00:00."+strings.Repeat("0", l)+"Z")
	}
	bruch := "2020-01-01T00:00:00.123456789012Z"
	for pos := 20; pos < 32; pos++ {
		texte = append(texte, eErsetzt(bruch, pos)...)
	}
	if len(texte) != 5120+3102 {
		t.Fatalf("Bruchteil %d", len(texte)-5120)
	}
	versatz := "2020-02-29T23:59:59+23:59" // ein Zeichen allein erreicht die Versatz-Stunde 24 und die Minute 60
	for pos := 19; pos < 25; pos++ {
		texte = append(texte, eErsetzt(versatz, pos)...)
	}
	n := ePruefeTexte(t, "T", texte, eZeitOrakel)
	if n != 9758 {
		t.Errorf("Menge %d, erwartet 9758", n)
	}
}

var eZahlRE = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// eZahlOrakel bewertet den Text einer Zahl mit math/big: der kanonische Text für den Typ oder false.
func eZahlOrakel(typ, s string) (string, bool) {
	if !eZahlRE.MatchString(s) {
		return "", false
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return "", false
	}
	if typ == "Dec" {
		text := r.FloatString(30)
		if strings.Contains(text, ".") {
			text = strings.TrimRight(strings.TrimRight(text, "0"), ".")
		}
		if text == "-0" {
			text = "0"
		}
		return text, true
	}
	if !r.IsInt() {
		return "", false
	}
	grenze := map[string]int{"I32": 32, "I64": 64}[typ]
	lo := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(grenze-1)))
	hi := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(grenze-1)), big.NewInt(1))
	if r.Num().Cmp(lo) < 0 || r.Num().Cmp(hi) > 0 {
		return "", false
	}
	return r.Num().String(), true
}

// TestEingabeZahlStellen hält Zeile 26 für Zahlen: Schablone `-1.5e+2`, jedes Zeichen an jeder der 7 Stellen. Als JSON-Zahl für Int32, Int64 und
// Decimal (ein Text, der keine Zahl nach RFC 8259 ist, ist ein Verstoß des Lesens), als Zeichenkette für Int64 und Decimal. Menge 5376 + 3584.
func TestEingabeZahlStellen(t *testing.T) {
	const schablone = "-1.5e+2"
	var texte, roh []string
	for pos := range len(schablone) {
		texte = append(texte, eErsetzt(schablone, pos)...)
		roh = append(roh, eErsetztRoh(schablone, pos)...)
	}
	n := 0
	for _, typ := range []string{"I32", "I64", "Dec"} {
		for _, s := range roh {
			// Leerraum davor ist Leerraum des Dokuments: die Zahl ist der Rest
			kanon, gueltig := eZahlOrakel(typ, strings.Trim(s, " \t\n\r"))
			got, err := eA("Ws", model.EingabePatch, `{"Id":1,"`+typ+`":`+s+`}`).los()
			was := typ + " " + strconv.Quote(s)
			switch {
			case !eZahlRE.MatchString(strings.Trim(s, " \t\n\r")):
				eFehler(t, was, got, err, model.EingabeSyntax, "")
			case gueltig:
				eOK(t, was, got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW(typ, eWArt[typ], kanon)}})
			default:
				eFehler(t, was, got, err, model.EingabeWert, "/"+typ)
			}
			n++
		}
	}
	if n != 5376 {
		t.Errorf("Zahlen %d, erwartet 5376", n)
	}
	for _, typ := range []string{"I64", "Dec"} {
		n += ePruefeTexte(t, typ, texte, func(s string) (string, bool) { return eZahlOrakel(typ, s) })
	}
	if n != 5376+3584 {
		t.Errorf("Menge %d, erwartet 8960", n)
	}
}

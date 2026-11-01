package urlparserv2_test

import (
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func keyOk(art model.LiteralArt, text string) erw { return okR(entityR("S", teil("", art, text))) }

// schreibweise setzt Buchstaben von wort groß, wo das Bit der Maske gesetzt ist (alle Schreibweisen durch Bitmasken).
func schreibweise(wort string, maske int) string {
	b := []byte(wort)
	for i := range b {
		if maske>>i&1 == 1 {
			b[i] = b[i] - 'a' + 'A'
		}
	}
	return string(b)
}

// TestSchluesselLiterale (Zeile 28): Zeichenkette, Wahrheitswert, ganze Zahl, `L`, Dezimalzahl mit oder ohne `M`: Art und kanonischer
// Text. Menge 35.
func TestSchluesselLiterale(t *testing.T) {
	zk, wv, zl := model.LiteralZeichenkette, model.LiteralWahrheitswert, model.LiteralZahl
	tabelle := []struct {
		eingabe string
		art     model.LiteralArt
		text    string
	}{
		{"''", zk, "''"}, {"'a'", zk, "'a'"}, {"'O''Hara'", zk, "'O''Hara'"}, {"'a,b'", zk, "'a,b'"}, {"'(x)'", zk, "'(x)'"},
		{"'a=b'", zk, "'a=b'"}, {"'%2F'", zk, "'/'"}, {"'%5C'", zk, "'\\'"},
		{"true", wv, "true"}, {"false", wv, "false"}, {"TRUE", wv, "true"}, {"False", wv, "false"}, {"tRuE", wv, "true"}, {"fALSE", wv, "false"},
		{"0", zl, "0"}, {"-0", zl, "0"}, {"007", zl, "7"}, {"42", zl, "42"}, {"-42", zl, "-42"}, {"0.0", zl, "0"}, {"1.50", zl, "1.5"},
		{"-0.50", zl, "-0.5"},
		{"42L", zl, "42"}, {"42l", zl, "42"}, {"-7L", zl, "-7"}, {"007L", zl, "7"}, {"0L", zl, "0"}, {"-0L", zl, "0"},
		{"1.5M", zl, "1.5"}, {"1.50m", zl, "1.5"}, {"0M", zl, "0"}, {"-0.0M", zl, "0"}, {"7M", zl, "7"}, {"7m", zl, "7"},
		{"-1.25M", zl, "-1.25"},
	}
	p := urlparserv2.New()
	for _, f := range tabelle {
		pruefe(t, p, "/S("+f.eingabe+")", "", keyOk(f.art, f.text))
		pruefe(t, p, "/S(K="+f.eingabe+")", "", okR(entityR("S", teil("K", f.art, f.text))))
	}
	if len(tabelle) != 35 {
		t.Errorf("%d Fälle, erwartet 35", len(tabelle))
	}
}

const guidBasis = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

// TestGuid (Zeile 29): `guid'…'`: das Präfix in jeder ASCII-Schreibweise, 8-4-4-4-12 Hex-Ziffern in jeder Schreibweise, der Text
// kleingeschrieben; jede Abweichung der Form ist `literal`. Menge 42: 16 Schreibweisen von `guid`, 6 der Ziffern, 20 Randformen.
func TestGuid(t *testing.T) {
	p := urlparserv2.New()
	n := 0
	for m := range 16 {
		pruefe(t, p, "/S("+schreibweise("guid", m)+"'"+guidBasis+"')", "", keyOk(model.LiteralGUID, guidBasis))
		n++
	}
	ziffern := func(f func(i int) bool) string {
		b := []byte(guidBasis)
		for i, c := range b {
			if c >= 'a' && c <= 'f' && f(i) {
				b[i] = c - 'a' + 'A'
			}
		}
		return string(b)
	}
	for _, g := range []string{
		ziffern(func(int) bool { return false }), ziffern(func(int) bool { return true }),
		ziffern(func(i int) bool { return i%2 == 0 }), ziffern(func(i int) bool { return i%2 == 1 }),
		ziffern(func(i int) bool { return i < 18 }), ziffern(func(i int) bool { return i >= 18 }),
	} {
		pruefe(t, p, "/S(guid'"+g+"')", "", keyOk(model.LiteralGUID, guidBasis))
		n++
	}
	gruppen := strings.Split(guidBasis, "-")
	var abweichend []string
	abweichend = append(abweichend, guidBasis[:35], guidBasis+"0")
	for _, pos := range []int{8, 13, 18, 23} {
		abweichend = append(abweichend, guidBasis[:pos]+guidBasis[pos+1:])
	}
	for gi := range gruppen {
		mehr, weniger := append([]string(nil), gruppen...), append([]string(nil), gruppen...)
		mehr[gi] += "0"
		weniger[gi] = weniger[gi][:len(weniger[gi])-1]
		abweichend = append(abweichend, strings.Join(mehr, "-"), strings.Join(weniger, "-"))
	}
	abweichend = append(abweichend, "g"+guidBasis[1:])
	for _, g := range abweichend {
		pruefe(t, p, "/S(guid'"+g+"')", "", ung(model.URLLiteral, "pfad/1"))
		n++
	}
	for _, roh := range []string{"guid" + guidBasis, "guid'{" + guidBasis + "}'", "guid'" + guidBasis} {
		pruefe(t, p, "/S("+roh+")", "", ung(model.URLLiteral, "pfad/1"))
		n++
	}
	if n != 42 {
		t.Errorf("%d Fälle, erwartet 42", n)
	}
}

// TestDatetime (Zeile 30): `datetime'JJJJ-MM-TTThh:mm'` mit `:ss` und `.f` (1 bis 7 Ziffern), ohne Zone, Text als UTC mit
// Bruchteil ohne Nullen am Ende. Menge 298: 256 Schreibweisen von `datetime` und 42 Formen. Das Orakel ist `time.Date` des Tests.
func TestDatetime(t *testing.T) {
	p := urlparserv2.New()
	n := 0
	soll := func(y, mo, d, h, mi, s int, bruch string) string {
		text := time.Date(y, time.Month(mo), d, h, mi, s, 0, time.UTC).Format("2006-01-02T15:04:05")
		if bruch != "" {
			text += "." + bruch
		}
		return text + "Z"
	}
	for m := range 256 {
		pruefe(t, p, "/S("+schreibweise("datetime", m)+"'2020-01-02T03:04')", "", keyOk(model.LiteralZeitpunkt, soll(2020, 1, 2, 3, 4, 0, "")))
		n++
	}
	gut := []struct {
		eingabe string
		text    string
	}{
		{"2020-01-02T03:04", soll(2020, 1, 2, 3, 4, 0, "")}, {"2020-01-02T03:04:05", soll(2020, 1, 2, 3, 4, 5, "")},
		{"2020-01-02T03:04:00", soll(2020, 1, 2, 3, 4, 0, "")},
		{"2020-01-02T03:04:05.1", soll(2020, 1, 2, 3, 4, 5, "1")}, {"2020-01-02T03:04:05.12", soll(2020, 1, 2, 3, 4, 5, "12")},
		{"2020-01-02T03:04:05.123", soll(2020, 1, 2, 3, 4, 5, "123")}, {"2020-01-02T03:04:05.1234", soll(2020, 1, 2, 3, 4, 5, "1234")},
		{"2020-01-02T03:04:05.12345", soll(2020, 1, 2, 3, 4, 5, "12345")}, {"2020-01-02T03:04:05.123456", soll(2020, 1, 2, 3, 4, 5, "123456")},
		{"2020-01-02T03:04:05.1234567", soll(2020, 1, 2, 3, 4, 5, "1234567")},
		{"2020-01-02T03:04:05.100", soll(2020, 1, 2, 3, 4, 5, "1")}, {"2020-01-02T03:04:05.000", soll(2020, 1, 2, 3, 4, 5, "")},
		{"2020-01-02T03:04:05.0010", soll(2020, 1, 2, 3, 4, 5, "001")},
		{"0001-01-01T00:00", soll(1, 1, 1, 0, 0, 0, "")}, {"9999-12-31T23:59", soll(9999, 12, 31, 23, 59, 0, "")},
		{"2020-12-01T00:00", soll(2020, 12, 1, 0, 0, 0, "")}, {"2020-01-31T00:00", soll(2020, 1, 31, 0, 0, 0, "")},
		{"2024-02-29T00:00", soll(2024, 2, 29, 0, 0, 0, "")}, {"2020-01-02T23:00", soll(2020, 1, 2, 23, 0, 0, "")},
		{"2020-01-02T03:59", soll(2020, 1, 2, 3, 59, 0, "")}, {"2020-01-02T03:04:59", soll(2020, 1, 2, 3, 4, 59, "")},
	}
	for _, f := range gut {
		pruefe(t, p, "/S(datetime'"+f.eingabe+"')", "", keyOk(model.LiteralZeitpunkt, f.text))
		n++
	}
	schlecht := []string{
		"0000-01-01T00:00", "2020-13-01T00:00", "2020-00-01T00:00", "2020-01-32T00:00", "2020-01-00T00:00", "2023-02-29T00:00",
		"2020-02-30T00:00", "2020-01-02T24:00", "2020-01-02T03:60", "2020-01-02T03:04:60",
		"2020-01-02T03:04:05Z", "2020-01-02T03:04:05+01:00", "2020-01-02T03:04:05-01:00",
		"2020-01-02", "2020-01-02t03:04", "2020-01-02 03:04", "2020-01-02T03:04:05.12345678",
		" 2020-01-02T03:04", "2020-01-02T03:04 ", "2020-01-02T03:04:05.", "2020-01-02T10",
	}
	for _, s := range schlecht {
		pruefe(t, p, "/S(datetime'"+s+"')", "", ung(model.URLLiteral, "pfad/1"))
		n++
	}
	if n != 256+len(gut)+len(schlecht) || len(gut) != 21 || len(schlecht) != 21 {
		t.Errorf("%d Fälle (%d gut, %d schlecht), erwartet 298", n, len(gut), len(schlecht))
	}
}

// TestLiteralTyp (Zeile 31): Literale außerhalb des Umfangs sind Kompatibilität `literal_typ`, nie ein URLFehler. Menge 75: 25 Formen
// × 3 Stellen (unbenannt, benannt, zweiter Teil).
func TestLiteralTyp(t *testing.T) {
	p := urlparserv2.New()
	n := 0
	for _, f := range ausserhalbLiterale {
		for _, pfad := range []string{"/S(" + f + ")", "/S(K=" + f + ")", "/S(A=1,B=" + f + ")"} {
			pruefe(t, p, pfad, "", kmp(model.KompatLiteralTyp, "pfad/1"))
			n++
		}
	}
	if n != 75 {
		t.Errorf("%d Fälle, erwartet 75", n)
	}
}

// TestLiteralVerstoss (Zeile 32): jede andere Form ist `literal`, nicht `literal_typ` und nicht `key`. Menge 31.
func TestLiteralVerstoss(t *testing.T) {
	formen := []string{"null", "NULL", "Null", "+1", ".5", "5.", "42LL", "42Lx", "1.5LM", "0x10", "1_000", guidBasis, "2020-01-02",
		"2020-01-02T03:04:05Z", "{" + guidBasis + "}", "guid'x'", "guid'", "datetime'2020-01-01'", "datetime'2020-01-02T03:04:05Z'",
		"datetime'2020-01-02T03:04:05+01:00'", "'a'b", "'a", "'a''", "1e", "1E+", "-", "--1", "1 2", "%20", "true1", "yes"}
	p := urlparserv2.New()
	for _, f := range formen {
		pruefe(t, p, "/S(K="+f+")", "", ung(model.URLLiteral, "pfad/1"))
	}
	if len(formen) != 31 {
		t.Errorf("%d Formen, erwartet 31", len(formen))
	}
}

// TestLiteralLaenge (Zeile 33): die Länge geht der Form vor: 512 Bytes mit Anführungszeichen sind lesbar, 513 sind `zu_lang`. Menge
// 12: sechs Formen × beide Seiten.
func TestLiteralLaenge(t *testing.T) {
	a := func(n int) string { return strings.Repeat("a", n) }
	formen := []struct {
		name string
		bau  func(laenge int) string
		want erw // bei 512 Bytes
	}{
		{"Zeichenkette", func(n int) string { return "'" + a(n-2) + "'" }, okR(entityR("S", teil("K", model.LiteralZeichenkette, "'"+a(510)+"'")))},
		{"ungeschlossen", func(n int) string { return "'" + a(n-1) }, ung(model.URLLiteral, "pfad/1")},
		{"X", func(n int) string { return "X'" + a(n-3) + "'" }, kmp(model.KompatLiteralTyp, "pfad/1")},
		{"guid", func(n int) string { return "guid'" + a(n-6) + "'" }, ung(model.URLLiteral, "pfad/1")},
		{"Ziffern", func(n int) string { return strings.Repeat("1", n) }, okR(entityR("S", teil("K", model.LiteralZahl, strings.Repeat("1", 512))))},
		{"datetime", func(n int) string { return "datetime'" + a(n-10) + "'" }, ung(model.URLLiteral, "pfad/1")},
	}
	p := urlparserv2.New()
	for _, f := range formen {
		for _, laenge := range []int{512, 513} {
			e := f.want
			if laenge == 513 {
				e = ung(model.URLZuLang, "pfad/1")
			}
			if l := f.bau(laenge); len(l) != laenge {
				t.Fatalf("%s: Länge %d statt %d", f.name, len(l), laenge)
			}
			pruefe(t, p, "/S(K="+f.bau(laenge)+")", "", e)
		}
	}
}

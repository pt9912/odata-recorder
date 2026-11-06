package model_test

import (
	"errors"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

var alleOps = []model.VergleichOp{model.VergleichEq, model.VergleichNe, model.VergleichGt, model.VergleichGe, model.VergleichLt, model.VergleichLe}

// erwartet ist die Wahrheit eines Vergleichs bei dem Vorzeichen c der Ordnung (links gegen rechts); die Tabelle steht im Test.
func erwartet(op model.VergleichOp, c int) model.Wahrheit {
	var b bool
	switch op {
	case "eq":
		b = c == 0
	case "ne":
		b = c != 0
	case "gt":
		b = c > 0
	case "ge":
		b = c >= 0
	case "lt":
		b = c < 0
	case "le":
		b = c <= 0
	}
	if b {
		return wW
	}
	return wF
}

func vorzeichen(c int) int {
	switch {
	case c < 0:
		return -1
	case c > 0:
		return 1
	}
	return 0
}

func werte(art model.WertArt, texte ...string) []model.Wert {
	var ws []model.Wert
	for _, t := range texte {
		ws = append(ws, model.Wert{Art: art, Text: t})
	}
	return ws
}

func istUngueltig(err error) bool {
	var ft *model.FilterTypFehler
	return errors.Is(err, model.ErrUngueltig) && !errors.As(err, &ft)
}

// Rohbytes (AGENTS.md §3.18): die Werte entstehen aus Bytewerten, nie aus Escape-Folgen im Quelltext.
func nbsp() string { return string([]byte{0xc2, 0xa0}) }
func bom() string  { return string([]byte{0xef, 0xbb, 0xbf}) }

// --- Orakel, unabhängig vom Code ---

func ratVon(t *testing.T, s string) *big.Rat {
	t.Helper()
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		t.Fatalf("Orakel liest %q nicht", s)
	}
	return r
}

func orakelZahl(t *testing.T) func(a, b string) int {
	return func(a, b string) int { return ratVon(t, a).Cmp(ratVon(t, b)) }
}

// instant ist der Zeitpunkt in Sekunden als Bruch: Sekunden seit der Epoche (time) plus der Bruchteil.
func instant(t *testing.T, s string) *big.Rat {
	t.Helper()
	z, err := time.Parse("2006-01-02T15:04:05", s[:19])
	if err != nil {
		t.Fatalf("Orakel liest %q nicht: %v", s, err)
	}
	r := new(big.Rat).SetInt64(z.Unix())
	if rest := strings.TrimSuffix(s[19:], "Z"); rest != "" {
		ziffern := rest[1:]
		zaehler, _ := new(big.Int).SetString(ziffern, 10)
		nenner := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(ziffern))), nil)
		r.Add(r, new(big.Rat).SetFrac(zaehler, nenner))
	}
	return r
}

func orakelZeit(t *testing.T) func(a, b string) int {
	return func(a, b string) int { return instant(t, a).Cmp(instant(t, b)) }
}

func orakelText(a, b string) int { return slices.Compare([]rune(a), []rune(b)) }

func orakelWahrheit(a, b string) int {
	rang := map[string]int{"false": 0, "true": 1}
	return rang[a] - rang[b]
}

func orakelDatum(a, b string) int {
	var ta, tb [3]int
	for i, p := range strings.Split(a, "-") {
		ta[i], _ = strconv.Atoi(p)
	}
	for i, p := range strings.Split(b, "-") {
		tb[i], _ = strconv.Atoi(p)
	}
	return slices.Compare(ta[:], tb[:])
}

// formOrakel hält die Form je Art als reguläre Ausdrücke (SPEC-166 Nr. 7); `\z` und `[0-9]` sind ASCII und ohne Zeilenende-Sonderfall.
var formOrakel = map[model.WertArt]*regexp.Regexp{
	model.ArtBoolean:        regexp.MustCompile(`\A(true|false)\z`),
	model.ArtInt32:          zahlForm,
	model.ArtInt64:          zahlForm,
	model.ArtDecimal:        zahlForm,
	model.ArtGUID:           regexp.MustCompile(`\A[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\z`),
	model.ArtDate:           regexp.MustCompile(`\A[0-9]{4}-[0-9]{2}-[0-9]{2}\z`),
	model.ArtDateTimeOffset: regexp.MustCompile(`\A[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{0,11}[1-9])?Z\z`),
}

var zahlForm = regexp.MustCompile(`\A(0|-?[1-9][0-9]*(\.[0-9]*[1-9])?|-?0\.[0-9]*[1-9])\z`)

func formGueltig(art model.WertArt, text string) bool {
	re, ok := formOrakel[art]
	if !ok {
		return true // ArtString nimmt jeden Text, ArtNull liest ihn nicht
	}
	if (art == model.ArtInt32 || art == model.ArtInt64 || art == model.ArtDecimal) && len(text) > 1024 {
		return false
	}
	return re.MatchString(text)
}

// --- Generator der Randformen ---

// randformen bildet aus einem gültigen Text alle Randformen der Bytes: Ränder mit Leerraum, Steuer- und
// Sonderzeichen, Wechsel der Schreibweise je Stelle, ein Zeichen je Stelle ersetzt, gestrichen und eingefügt,
// jedes Präfix und jedes Suffix, verdoppelt. Das Ergebnis ist die Menge ohne den Text selbst.
func randformen(seed string) []string {
	set := map[string]bool{}
	raender := []string{" ", "\n", "\t", "\r", "\v", "\f", "\x00", nbsp(), bom()}
	ersatz := []string{"x", "X", " ", "/", ":", "@", "G", "`", "g", "\n", "-", "0", "9", "a", ".", "e", "t", nbsp()}
	einfuege := []string{"0", "-", ".", " ", "z"}
	for _, r := range raender {
		set[r+seed], set[seed+r] = true, true
	}
	for i := 0; i < len(seed); i++ {
		c := seed[i]
		switch {
		case c >= 'a' && c <= 'z':
			set[seed[:i]+string(c-32)+seed[i+1:]] = true
		case c >= 'A' && c <= 'Z':
			set[seed[:i]+string(c+32)+seed[i+1:]] = true
		}
		for _, z := range ersatz {
			set[seed[:i]+z+seed[i+1:]] = true
		}
		set[seed[:i]+seed[i+1:]] = true
	}
	for i := 0; i <= len(seed); i++ {
		for _, z := range einfuege {
			set[seed[:i]+z+seed[i:]] = true
		}
		if i < len(seed) {
			set[seed[:i]] = true
			set[seed[i+1:]] = true
		}
	}
	set[seed+seed] = true
	set[seed+seed[len(seed)-1:]] = true
	delete(set, seed)
	var out []string
	for k := range set {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// --- Klassenlisten ---

type klassenListe struct {
	name   string
	werte  []model.Wert
	orakel func(a, b string) int
}

func lang(vorn string, n int, zeichen string, ende string) string {
	return vorn + strings.Repeat(zeichen, n) + ende
}

func textListe() []model.Wert {
	return werte(model.ArtString, "", "\x00", " a", "A", "AAA", "AB", "B", "a", "a ", "ab", "abc", "e"+string([]byte{0xcc, 0x81}), nbsp(),
		string([]byte{0xc3, 0xa9}), string([]byte{0xc3, 0xa9}), string([]byte{0xef, 0xbf, 0xbf}), string([]byte{0xf0, 0x90, 0x80, 0x80}))
}

func dezimalListe() []model.Wert {
	return werte(model.ArtDecimal, "0.1", "0.10000000000000001", "-1.5", "-1.25", "-0.5", "-0.25", "0", "9", "10", "99", "100", "0.3",
		"0.30000000000000004", "-10", "-9", "-100", "0.1234567890123456789012345678901234567891", "0.100000000000000000000001", "0.100000000000000000000002",
		lang("1", 500, "0", "1")+lang("", 400, "0", ""), lang("1", 500, "0", "2")+lang("", 400, "0", ""), lang("0.", 1021, "0", "1"),
		lang("1", 1023, "0", ""), lang("1", 511, "0", "")+"."+lang("", 510, "0", "1"))
}

func zeitListe() []model.Wert {
	return werte(model.ArtDateTimeOffset, "0001-01-01T00:00:00Z", "2020-01-01T00:00:00Z", "2020-01-01T00:00:00.000000000001Z",
		"2020-01-01T00:00:00.000000000002Z", "2020-01-01T00:00:00.49Z", "2020-01-01T00:00:00.5Z", "2020-01-01T00:00:00.51Z", "2020-01-01T00:00:01Z", "2020-01-01T23:59:59Z",
		"2020-01-02T00:00:00Z", "2020-12-31T23:59:59.999999999998Z", "2020-12-31T23:59:59.999999999999Z", "2021-01-01T00:00:00Z", "9999-12-31T23:59:59.999999999999Z",
		"2020-01-01T00:00:00.5Z")
}

func klassenListen(t *testing.T) []klassenListe {
	return []klassenListe{
		{"zahl", werte(model.ArtInt64, "9007199254740991", "9007199254740992", "9007199254740993", "9007199254740994", "-9007199254740993",
			"9223372036854775807", "-9223372036854775808"), orakelZahl(t)},
		{"dezimal", dezimalListe(), orakelZahl(t)},
		{"text", textListe(), orakelText},
		{"wahrheit", werte(model.ArtBoolean, "false", "true"), orakelWahrheit},
		{"datum", werte(model.ArtDate, "0001-01-01", "2024-02-28", "2024-02-29", "2024-03-01", "2025-01-01", "9999-12-31", "2024-02-29"), orakelDatum},
		{"zeitpunkt", zeitListe(), orakelZeit(t)},
	}
}

// pruefeKlasse vergleicht alle geordneten Paare der Liste in allen sechs Operatoren und mit OrdneWerte gegen das
// Orakel; es liefert die Zahl der geprüften Vergleiche.
func pruefeKlasse(t *testing.T, name string, ws []model.Wert, orakel func(a, b string) int) int {
	t.Helper()
	n := 0
	for _, a := range ws {
		for _, b := range ws {
			c := vorzeichen(orakel(a.Text, b.Text))
			for _, op := range alleOps {
				got, err := model.Vergleiche(op, a, b)
				if err != nil || got != erwartet(op, c) {
					t.Errorf("%s: %q %s %q = %d, %v, erwartet %d", name, a.Text, op, b.Text, got, err, erwartet(op, c))
				}
				n++
			}
		}
	}
	return n
}

// TestZahlExakt (Zeile 1): Int64 und Decimal um 2^53 und an den Grenzen von Int64: alle Paare, sechs Operatoren, gegen math/big.
func TestZahlExakt(t *testing.T) {
	l := klassenListen(t)[0]
	n := 0
	for _, art := range []model.WertArt{model.ArtInt64, model.ArtDecimal} {
		ws := werte(art, "9007199254740991", "9007199254740992", "9007199254740993", "9007199254740994", "-9007199254740993",
			"9223372036854775807", "-9223372036854775808")
		n += pruefeKlasse(t, string(art), ws, l.orakel)
	}
	if n != 2*7*7*6 {
		t.Errorf("%d Vergleiche", n)
	}
}

// TestZahlGleichUeberArten (Zeile 2): gleiche Werte über int32, int64 und decimal in allen neun Paaren der Arten.
func TestZahlGleichUeberArten(t *testing.T) {
	texte := []string{"0", "1", "-1", "5", "2147483647", "-2147483648"}
	arten := []model.WertArt{model.ArtInt32, model.ArtInt64, model.ArtDecimal}
	n := 0
	for _, la := range arten {
		for _, ra := range arten {
			for _, a := range texte {
				for _, b := range texte {
					c := vorzeichen(ratVon(t, a).Cmp(ratVon(t, b)))
					for _, op := range alleOps {
						got, err := model.Vergleiche(op, model.Wert{Art: la, Text: a}, model.Wert{Art: ra, Text: b})
						if err != nil || got != erwartet(op, c) {
							t.Errorf("%s %s %s %s %s: %d, %v", la, a, op, ra, b, got, err)
						}
						n++
					}
				}
			}
		}
	}
	if n != 9*36*6 {
		t.Errorf("%d Vergleiche", n)
	}
}

// TestDecimalExakt (Zeile 3): Decimal ohne Binär-Float und ohne Facetten, alle Paare gegen math/big; 1024 Zeichen gültig, 1025 nicht.
func TestDecimalExakt(t *testing.T) {
	l := klassenListen(t)[1]
	if n := pruefeKlasse(t, l.name, l.werte, l.orakel); n != 24*24*6 {
		t.Errorf("%d Vergleiche", n)
	}
	for _, w := range l.werte {
		if len(w.Text) > 1024 {
			t.Errorf("Element der Liste mit %d Zeichen", len(w.Text))
		}
	}
	lang1024 := lang("0.", 1021, "0", "1")
	lang1025 := lang("0.", 1022, "0", "1")
	ganz1024, ganz1025 := lang("1", 1023, "0", ""), lang("1", 1024, "0", "")
	for _, c := range []struct {
		text string
		ok   bool
	}{{lang1024, true}, {lang1025, false}, {ganz1024, true}, {ganz1025, false}} {
		for _, art := range []model.WertArt{model.ArtDecimal, model.ArtInt32, model.ArtInt64} {
			w := model.Wert{Art: art, Text: c.text}
			_, e1 := model.Vergleiche(model.VergleichEq, w, w)
			_, e2 := model.Vergleiche(model.VergleichNe, w, model.Wert{Art: model.ArtNull})
			_, e3 := model.OrdneWerte(model.Wert{Art: model.ArtDecimal, Text: "0"}, w)
			if (e1 == nil) != c.ok || (e2 == nil) != c.ok || (e3 == nil) != c.ok {
				t.Errorf("%s mit %d Zeichen: %v %v %v, erwartet gültig=%v", art, len(c.text), e1, e2, e3, c.ok)
			}
		}
	}
	if len(lang1024) != 1024 || len(lang1025) != 1025 || len(ganz1024) != 1024 || len(ganz1025) != 1025 {
		t.Error("Länge der Prüfwerte")
	}
}

// TestTextOrdinal (Zeile 4): Text nach Codepunkten, ohne Locale, Normalisierung und Faltung; alle Paare, sechs Operatoren.
func TestTextOrdinal(t *testing.T) {
	l := klassenListen(t)[2]
	if n := pruefeKlasse(t, l.name, l.werte, l.orakel); n != 17*17*6 {
		t.Errorf("%d Vergleiche", n)
	}
	// die Liste steigt nach Codepunkten (das Orakel zählt Runen, nicht Bytes)
	for i := 1; i < len(l.werte); i++ {
		if l.orakel(l.werte[i-1].Text, l.werte[i].Text) > 0 {
			t.Errorf("Liste fällt bei %d", i)
		}
	}
	gleich, _ := model.Vergleiche(model.VergleichEq, model.Wert{Art: model.ArtString, Text: "a"}, model.Wert{Art: model.ArtString, Text: "A"})
	if gleich != wF {
		t.Error("a eq A ist wahr")
	}
}

// TestWahrheitOrdnung (Zeile 5): false vor true, alle vier Paare, sechs Operatoren.
func TestWahrheitOrdnung(t *testing.T) {
	ws := werte(model.ArtBoolean, "false", "true")
	if n := pruefeKlasse(t, "wahrheit", ws, orakelWahrheit); n != 4*6 {
		t.Errorf("%d Vergleiche", n)
	}
}

// TestDatum (Zeile 6): Date wie der Kalender, gleiche Daten sind gleich.
func TestDatum(t *testing.T) {
	l := klassenListen(t)[4]
	if n := pruefeKlasse(t, l.name, l.werte, l.orakel); n != 7*7*6 {
		t.Errorf("%d Vergleiche", n)
	}
}

// TestZeitpunktOrdnung (Zeile 7): DateTimeOffset nach Instant, Sekunde und Bruchteil, alle Paare gegen math/big.
func TestZeitpunktOrdnung(t *testing.T) {
	l := klassenListen(t)[5]
	if n := pruefeKlasse(t, l.name, l.werte, l.orakel); n != 15*15*6 {
		t.Errorf("%d Vergleiche", n)
	}
	a := model.Wert{Art: model.ArtDateTimeOffset, Text: "2020-01-01T00:00:00Z"}
	b := model.Wert{Art: model.ArtDateTimeOffset, Text: "2020-01-01T00:00:00.5Z"}
	if got, _ := model.Vergleiche(model.VergleichLt, a, b); got != wW || a.Text < b.Text {
		t.Error("der Text allein ordnet nicht: Z steht hinter dem Punkt, der Vergleich muss den Bruchteil lesen")
	}
}

// TestKlassenUnvergleichbar (Zeile 9): jede Art gegen jede Art; verschiedene Klassen sind ErrUngueltig, dieselbe Klasse
// vergleicht, ArtNull vergleicht mit jeder Art.
func TestKlassenUnvergleichbar(t *testing.T) {
	klasse := map[model.WertArt]string{model.ArtString: "text", model.ArtBoolean: "wahrheit", model.ArtInt32: "zahl", model.ArtInt64: "zahl",
		model.ArtDecimal: "zahl", model.ArtGUID: "guid", model.ArtDate: "datum", model.ArtDateTimeOffset: "zeitpunkt"}
	probe := map[model.WertArt]string{model.ArtString: "a", model.ArtBoolean: "true", model.ArtInt32: "1", model.ArtInt64: "1", model.ArtDecimal: "1",
		model.ArtGUID: "00000000-0000-0000-0000-000000000000", model.ArtDate: "2020-01-01", model.ArtDateTimeOffset: "2020-01-01T00:00:00Z"}
	ordnung := func(op model.VergleichOp) bool { return op != "eq" && op != "ne" }
	n, verschieden := 0, 0
	for la, lk := range klasse {
		for ra, rk := range klasse {
			l, r := model.Wert{Art: la, Text: probe[la]}, model.Wert{Art: ra, Text: probe[ra]}
			for _, op := range alleOps {
				got, err := model.Vergleiche(op, l, r)
				erwartetFehler := lk != rk || (lk == "guid" && ordnung(op))
				if erwartetFehler != (err != nil) || (err != nil && !istUngueltig(err)) || (err == nil && got != erwartet(op, 0)) {
					t.Errorf("%s %s %s: %d, %v", la, op, ra, got, err)
				}
				if _, err := model.OrdneWerte(l, r); (lk != rk || lk == "guid") != (err != nil) {
					t.Errorf("OrdneWerte %s %s: %v", la, ra, err)
				}
				n++
				if lk != rk {
					verschieden++
				}
			}
		}
		// ArtNull ist mit jeder Art vergleichbar, in beiden Richtungen
		for _, op := range alleOps {
			null, w := model.Wert{Art: model.ArtNull}, model.Wert{Art: la, Text: probe[la]}
			for _, p := range [][2]model.Wert{{null, w}, {w, null}} {
				_, err := model.Vergleiche(op, p[0], p[1])
				if (lk == "guid" && ordnung(op)) != (err != nil) {
					t.Errorf("null und %s %s: %v", la, op, err)
				}
				n++
			}
		}
	}
	if n != 64*6+8*6*2 || verschieden != 50*6 {
		t.Errorf("%d Fälle, %d verschiedener Klassen", n, verschieden)
	}
	// die 30 geordneten Paare verschiedener Klassen, eine Art je Klasse, Date gegen DateTimeOffset ausdrücklich
	for _, op := range alleOps {
		_, e1 := model.Vergleiche(op, model.Wert{Art: model.ArtDate, Text: "2020-01-01"}, model.Wert{Art: model.ArtDateTimeOffset, Text: "2020-01-01T00:00:00Z"})
		_, e2 := model.Vergleiche(op, model.Wert{Art: model.ArtDateTimeOffset, Text: "2020-01-01T00:00:00Z"}, model.Wert{Art: model.ArtDate, Text: "2020-01-01"})
		_, e3 := model.Vergleiche(op, model.Wert{Art: model.ArtString, Text: "00000000-0000-0000-0000-000000000000"}, model.Wert{Art: model.ArtGUID, Text: "00000000-0000-0000-0000-000000000000"})
		if !istUngueltig(e1) || !istUngueltig(e2) || !istUngueltig(e3) {
			t.Errorf("%s: %v %v %v", op, e1, e2, e3)
		}
	}
}

// TestGuid (Zeile 10): nur Gleichheit; Ordnungsoperatoren und OrdneWerte sind ErrUngueltig, auch gegen ArtNull; HatOrdnung.
func TestGuid(t *testing.T) {
	g1 := model.Wert{Art: model.ArtGUID, Text: "00000000-0000-0000-0000-000000000001"}
	g2 := model.Wert{Art: model.ArtGUID, Text: "00000000-0000-0000-0000-000000000002"}
	null := model.Wert{Art: model.ArtNull}
	for _, p := range []struct {
		l, r  model.Wert
		eq, n model.Wahrheit
	}{{g1, g1, wW, wF}, {g1, g2, wF, wW}, {g2, g1, wF, wW}} {
		if got, err := model.Vergleiche(model.VergleichEq, p.l, p.r); err != nil || got != p.eq {
			t.Errorf("eq: %d %v", got, err)
		}
		if got, err := model.Vergleiche(model.VergleichNe, p.l, p.r); err != nil || got != p.n {
			t.Errorf("ne: %d %v", got, err)
		}
	}
	for _, op := range []model.VergleichOp{model.VergleichGt, model.VergleichGe, model.VergleichLt, model.VergleichLe} {
		for _, p := range [][2]model.Wert{{g1, g1}, {g1, g2}, {g2, g1}, {g1, null}, {null, g1}} {
			if got, err := model.Vergleiche(op, p[0], p[1]); !istUngueltig(err) || got != wU {
				t.Errorf("%s: %d %v", op, got, err)
			}
		}
	}
	for _, p := range [][2]model.Wert{{g1, g1}, {g1, g2}, {g1, null}, {null, g1}} {
		if _, err := model.OrdneWerte(p[0], p[1]); !istUngueltig(err) {
			t.Errorf("OrdneWerte: %v", err)
		}
	}
	namen := model.UnterstuetzteEDMTypen()
	if len(namen) != 8 {
		t.Fatalf("%d Typen", len(namen))
	}
	n := 0
	for _, name := range namen {
		if model.HatOrdnung(name) != (name != "Edm.Guid") {
			t.Errorf("HatOrdnung(%q) = %v", name, !(name != "Edm.Guid"))
		}
		for _, m := range randformen(name) {
			if soll := slices.Contains(namen, m) && m != "Edm.Guid"; model.HatOrdnung(m) != soll {
				t.Errorf("HatOrdnung(%q) ist %v", m, !soll)
			}
			n++
		}
	}
	for _, name := range append([]string{"", "edm.string", "Collection(Edm.String)", "Ns.Color"}, typenAusserhalb()...) {
		if model.HatOrdnung(name) {
			t.Errorf("HatOrdnung(%q) ist wahr", name)
		}
	}
	if n != randformenZahlTypnamen {
		t.Errorf("%d Randformen der Typnamen", n)
	}
}

// TestNullVergleiche (Zeile 11): null in den sechs Vergleichen, die Paare (null, null), (null, x), (x, null) und neun
// Vertreter; der Text eines ArtNull-Wertes zählt nicht.
func TestNullVergleiche(t *testing.T) {
	vertreter := []model.Wert{{Art: model.ArtString, Text: ""}, {Art: model.ArtInt32, Text: "0"}, {Art: model.ArtBoolean, Text: "false"},
		{Art: model.ArtDate, Text: "2020-01-01"}, {Art: model.ArtDateTimeOffset, Text: "2020-01-01T00:00:00Z"}, {Art: model.ArtDecimal, Text: "1.5"},
		{Art: model.ArtString, Text: "x"}, {Art: model.ArtString, Text: "null"}, {Art: model.ArtGUID, Text: "00000000-0000-0000-0000-000000000000"}}
	nulls := []model.Wert{{Art: model.ArtNull}, {Art: model.ArtNull, Text: "x"}}
	n := 0
	for _, op := range alleOps {
		beideNull := op == "eq" || op == "ge" || op == "le"
		for _, a := range nulls {
			for _, b := range nulls {
				want := wF
				if beideNull {
					want = wW
				}
				if got, err := model.Vergleiche(op, a, b); err != nil || got != want {
					t.Errorf("null %s null: %d %v", op, got, err)
				}
				n++
			}
		}
		for _, x := range vertreter {
			if x.Art == model.ArtGUID && op != "eq" && op != "ne" {
				continue
			}
			want := wF
			if op == "ne" {
				want = wW
			}
			for _, nl := range nulls {
				for _, p := range [][2]model.Wert{{nl, x}, {x, nl}} {
					if got, err := model.Vergleiche(op, p[0], p[1]); err != nil || got != want {
						t.Errorf("%v %s %v: %d %v", p[0], op, p[1], got, err)
					}
					n++
				}
			}
		}
	}
	if n != 6*4+(6*9-4)*4 {
		t.Errorf("%d Fälle", n)
	}
}

// TestOrdneWerteGleichVergleiche (Zeile 16): OrdneWerte gleich dem Orakel und gleich lt/eq/gt von Vergleiche; null kommt zuerst.
func TestOrdneWerteGleichVergleiche(t *testing.T) {
	n := 0
	for _, l := range klassenListen(t) {
		for _, a := range l.werte {
			for _, b := range l.werte {
				c := vorzeichen(l.orakel(a.Text, b.Text))
				got, err := model.OrdneWerte(a, b)
				if err != nil || got != c {
					t.Errorf("%s: OrdneWerte(%q, %q) = %d %v, erwartet %d", l.name, a.Text, b.Text, got, err, c)
				}
				for op, soll := range map[model.VergleichOp]int{"lt": -1, "eq": 0, "gt": 1} {
					v, _ := model.Vergleiche(op, a, b)
					if (v == wW) != (c == soll) {
						t.Errorf("%s: %q %s %q gegen OrdneWerte %d", l.name, a.Text, op, b.Text, c)
					}
				}
				n++
			}
			null := model.Wert{Art: model.ArtNull}
			if g, err := model.OrdneWerte(null, a); err != nil || g != -1 {
				t.Errorf("%s: null gegen %q: %d %v", l.name, a.Text, g, err)
			}
			if g, err := model.OrdneWerte(a, null); err != nil || g != 1 {
				t.Errorf("%s: %q gegen null: %d %v", l.name, a.Text, g, err)
			}
			if lt, _ := model.Vergleiche(model.VergleichLt, null, a); lt != wF {
				t.Errorf("%s: null lt %q ist wahr", l.name, a.Text)
			}
			n += 2
		}
	}
	if g, err := model.OrdneWerte(model.Wert{Art: model.ArtNull}, model.Wert{Art: model.ArtNull, Text: "x"}); err != nil || g != 0 {
		t.Errorf("null gegen null: %d %v", g, err)
	}
	if n != 1336 {
		t.Errorf("%d Fälle", n)
	}
}

// TestOrdnungTotal (Zeile 17): über alle Tripel jeder Liste samt null: Reflexivität, Antisymmetrie, Transitivität.
func TestOrdnungTotal(t *testing.T) {
	n := 0
	for _, l := range klassenListen(t) {
		ws := append([]model.Wert{{Art: model.ArtNull}}, l.werte...)
		ord := func(a, b model.Wert) int {
			c, err := model.OrdneWerte(a, b)
			if err != nil {
				t.Fatalf("%s: %v", l.name, err)
			}
			return c
		}
		for _, a := range ws {
			if ord(a, a) != 0 {
				t.Errorf("%s: %q nicht gleich sich selbst", l.name, a.Text)
			}
			for _, b := range ws {
				if ord(a, b) != -ord(b, a) {
					t.Errorf("%s: nicht antisymmetrisch %q %q", l.name, a.Text, b.Text)
				}
				for _, c := range ws {
					ab, bc, ac := ord(a, b), ord(b, c), ord(a, c)
					if ab <= 0 && bc <= 0 && (ac > 0 || ((ab < 0 || bc < 0) && ac >= 0)) {
						t.Errorf("%s: nicht transitiv %q %q %q", l.name, a.Text, b.Text, c.Text)
					}
					n++
				}
			}
		}
	}
	if n != 26604 {
		t.Errorf("%d Tripel statt 26604", n)
	}
}

// TestWertAusLiteral (Zeile 18): jede der sechs Arten; Zeichenketten als Klasse: alle Inhalte über {', a} bis zur Länge 5
// kodiert (jedes ' verdoppelt) kehren zurück; alle Texte über {', a} bis zur Länge 7 gegen den regulären Ausdruck der
// Kodierung: gültig genau dann, wenn er passt.
func TestWertAusLiteral(t *testing.T) {
	for _, c := range []struct {
		l    model.Literal
		want model.Wert
	}{
		{model.Literal{Art: model.LiteralWahrheitswert, Text: "true"}, model.Wert{Art: model.ArtBoolean, Text: "true"}},
		{model.Literal{Art: model.LiteralZahl, Text: "1.5"}, model.Wert{Art: model.ArtDecimal, Text: "1.5"}},
		{model.Literal{Art: model.LiteralGUID, Text: "00000000-0000-0000-0000-000000000000"}, model.Wert{Art: model.ArtGUID, Text: "00000000-0000-0000-0000-000000000000"}},
		{model.Literal{Art: model.LiteralDatum, Text: "2020-01-01"}, model.Wert{Art: model.ArtDate, Text: "2020-01-01"}},
		{model.Literal{Art: model.LiteralZeitpunkt, Text: "2020-01-01T00:00:00Z"}, model.Wert{Art: model.ArtDateTimeOffset, Text: "2020-01-01T00:00:00Z"}},
		{model.Literal{Art: model.LiteralZahl, Text: ""}, model.Wert{Art: model.ArtDecimal, Text: ""}},
		{model.Literal{Art: model.LiteralDatum, Text: "kein Datum"}, model.Wert{Art: model.ArtDate, Text: "kein Datum"}},
		{model.Literal{Art: model.LiteralZeichenkette, Text: "'a'"}, model.Wert{Art: model.ArtString, Text: "a"}},
		{model.Literal{Art: model.LiteralZeichenkette, Text: "''"}, model.Wert{Art: model.ArtString, Text: ""}},
		{model.Literal{Art: model.LiteralZeichenkette, Text: "''''"}, model.Wert{Art: model.ArtString, Text: "'"}},
		{model.Literal{Art: model.LiteralZeichenkette, Text: "'O''Hara'"}, model.Wert{Art: model.ArtString, Text: "O'Hara"}},
		{model.Literal{Art: model.LiteralZeichenkette, Text: "'a''''b'"}, model.Wert{Art: model.ArtString, Text: "a''b"}},
		{model.Literal{Art: model.LiteralZeichenkette, Text: "'a\nb" + nbsp() + "'"}, model.Wert{Art: model.ArtString, Text: "a\nb" + nbsp()}},
	} {
		if got, err := model.WertAusLiteral(c.l); err != nil || got != c.want {
			t.Errorf("%v: %v %v, erwartet %v", c.l, got, err, c.want)
		}
	}
	for _, l := range []model.Literal{{}, {Art: 7, Text: "'a'"}, {Art: -1}, {Art: 0, Text: "true"}} {
		if got, err := model.WertAusLiteral(l); !istUngueltig(err) || got != (model.Wert{}) {
			t.Errorf("%v: %v %v", l, got, err)
		}
	}
	kodierung := regexp.MustCompile(`\A'([^']|'')*'\z`)
	var alle func(n int) []string
	alle = func(n int) []string {
		if n == 0 {
			return []string{""}
		}
		var out []string
		for _, s := range alle(n - 1) {
			out = append(out, s+"'", s+"a")
		}
		return out
	}
	inhalte, kandidaten := 0, 0
	for n := 0; n <= 5; n++ {
		for _, inhalt := range alle(n) {
			kodiert := "'" + strings.ReplaceAll(inhalt, "'", "''") + "'"
			if got, err := model.WertAusLiteral(model.Literal{Art: model.LiteralZeichenkette, Text: kodiert}); err != nil || got.Text != inhalt || got.Art != model.ArtString {
				t.Errorf("%q: %q %v, erwartet %q", kodiert, got.Text, err, inhalt)
			}
			inhalte++
		}
	}
	gueltig := 0
	for n := 0; n <= 7; n++ {
		for _, s := range alle(n) {
			_, err := model.WertAusLiteral(model.Literal{Art: model.LiteralZeichenkette, Text: s})
			if (err == nil) != kodierung.MatchString(s) {
				t.Errorf("%q: Fehler %v, Orakel %v", s, err, kodierung.MatchString(s))
			}
			if err == nil {
				gueltig++
			}
			kandidaten++
		}
	}
	if inhalte != 63 || kandidaten != 255 || gueltig == 0 || gueltig == kandidaten {
		t.Errorf("%d Inhalte, %d Kandidaten, %d gültig", inhalte, kandidaten, gueltig)
	}
	nRand := 0
	for _, seed := range []string{"'a'", "'O''Hara'", "''", "''''", "' a b'"} {
		for _, s := range randformen(seed) {
			got, err := model.WertAusLiteral(model.Literal{Art: model.LiteralZeichenkette, Text: s})
			if kodierung.MatchString(s) {
				soll := strings.ReplaceAll(s[1:len(s)-1], "''", "'")
				if err != nil || got.Text != soll {
					t.Errorf("%q: %q %v, erwartet %q", s, got.Text, err, soll)
				}
			} else if !istUngueltig(err) {
				t.Errorf("%q: %v", s, err)
			}
			nRand++
		}
	}
	if nRand != randformenLiteral {
		t.Errorf("%d Randformen der Zeichenketten", nRand)
	}
	for _, s := range []string{"abc", "'a", "a'", "'", "", "\"a\"", "'a'b'", " 'a'", "'a' ", "'a'\n"} {
		if _, err := model.WertAusLiteral(model.Literal{Art: model.LiteralZeichenkette, Text: s}); !istUngueltig(err) {
			t.Errorf("%q: %v", s, err)
		}
	}
}

var gueltigeSamen = map[model.WertArt][]string{
	model.ArtBoolean: {"true", "false"},
	model.ArtInt32:   {"0", "5", "-1", "2147483647", "-2147483648"},
	model.ArtInt64:   {"9007199254740993", "-9223372036854775808", "10"},
	model.ArtDecimal: {"0.1", "-0.5", "1.25", "10", "0.10000000000000001", "-1.5"},
	model.ArtGUID:    {"0a1b2c3d-4e5f-6789-abcd-ef0123456789", "00000000-0000-0000-0000-000000000000"},
	model.ArtDate:    {"2024-02-29", "0001-01-01"},
	model.ArtDateTimeOffset: {"2020-01-01T10:00:00Z", "2020-01-01T10:00:00.5Z", "2020-01-01T10:00:00.000000000001Z",
		"9999-12-31T23:59:59.123456789012Z"},
}

// TestFormverstoesse (Zeile 19): die Randformen der Bytes jedes gültigen Samens (Generator randformen) in beiden
// Richtungen von Vergleiche und OrdneWerte, gegen null und für WahrheitVon: ErrUngueltig genau dann, wenn der reguläre
// Ausdruck der Art nicht passt. ArtRedigiert, unbekannte Arten und unbekannte Operatoren.
func TestFormverstoesse(t *testing.T) {
	ungueltig, gueltigMutanten, samen := 0, 0, 0
	for art, liste := range gueltigeSamen {
		for _, seed := range liste {
			samen++
			if !formGueltig(art, seed) {
				t.Fatalf("Samen %q ist nach dem Orakel ungültig", seed)
			}
			gut := model.Wert{Art: art, Text: seed}
			null := model.Wert{Art: model.ArtNull}
			for _, m := range append([]string{seed}, randformen(seed)...) {
				w := model.Wert{Art: art, Text: m}
				ok := formGueltig(art, m)
				if ok && m != seed {
					gueltigMutanten++
				}
				if !ok {
					ungueltig++
				}
				for name, f := range map[string]func() (model.Wahrheit, error){
					"links":     func() (model.Wahrheit, error) { return model.Vergleiche(model.VergleichEq, w, gut) },
					"rechts":    func() (model.Wahrheit, error) { return model.Vergleiche(model.VergleichEq, gut, w) },
					"null":      func() (model.Wahrheit, error) { return model.Vergleiche(model.VergleichNe, w, null) },
					"nullLinks": func() (model.Wahrheit, error) { return model.Vergleiche(model.VergleichNe, null, w) },
					"ordnungL": func() (model.Wahrheit, error) {
						_, err := model.OrdneWerte(w, gut)
						return wU, err
					},
					"ordnungR": func() (model.Wahrheit, error) {
						_, err := model.OrdneWerte(gut, w)
						return wU, err
					},
				} {
					if art == model.ArtGUID && strings.HasPrefix(name, "ordnung") {
						continue // Guid hat keine Ordnung (Zeile 10)
					}
					got, err := f()
					if ok && err != nil {
						t.Errorf("%s %q %s: %v", art, m, name, err)
					}
					if !ok && (!istUngueltig(err) || got != wU || strings.Contains(err.Error(), m) && len(m) > 3) {
						t.Errorf("%s %q %s: %d %v", art, m, name, got, err)
					}
				}
				if art == model.ArtBoolean {
					got, err := model.WahrheitVon(w)
					if ok != (err == nil) || (ok && got == wU) {
						t.Errorf("WahrheitVon(%q): %d %v", m, got, err)
					}
				}
			}
		}
	}
	if samen != 24 || ungueltig != randformenUngueltig || gueltigMutanten != randformenGueltig {
		t.Errorf("%d Samen, %d ungültige und %d gültige Randformen", samen, ungueltig, gueltigMutanten)
	}
	// ArtString nimmt jeden Text, ArtNull liest ihn nicht
	for _, m := range append([]string{"x"}, randformen("a b")...) {
		s, n := model.Wert{Art: model.ArtString, Text: m}, model.Wert{Art: model.ArtNull, Text: m}
		if got, err := model.Vergleiche(model.VergleichEq, s, s); err != nil || got != wW {
			t.Errorf("string %q: %d %v", m, got, err)
		}
		if got, err := model.Vergleiche(model.VergleichEq, n, model.Wert{Art: model.ArtNull}); err != nil || got != wW {
			t.Errorf("null %q: %d %v", m, got, err)
		}
	}
	gut := model.Wert{Art: model.ArtString, Text: "a"}
	for _, art := range []model.WertArt{model.ArtRedigiert, "", "bogus", "String", "NULL", "int16", " string"} {
		for _, p := range [][2]model.Wert{{{Art: art}, gut}, {gut, {Art: art}}, {{Art: art}, {Art: model.ArtNull}}, {{Art: model.ArtNull}, {Art: art}}} {
			if _, err := model.Vergleiche(model.VergleichEq, p[0], p[1]); !istUngueltig(err) {
				t.Errorf("Art %q: %v", art, err)
			}
			if _, err := model.OrdneWerte(p[0], p[1]); !istUngueltig(err) {
				t.Errorf("OrdneWerte Art %q: %v", art, err)
			}
		}
	}
	// der Operator: nur die sechs; jede Randform eines Operators ist ErrUngueltig
	gueltigOps := map[string]bool{"eq": true, "ne": true, "gt": true, "ge": true, "lt": true, "le": true}
	nOps := 0
	for _, op := range alleOps {
		for _, m := range append([]string{"", "EQ", "like", "and", "not"}, randformen(string(op))...) {
			got, err := model.Vergleiche(model.VergleichOp(m), gut, gut)
			if gueltigOps[m] {
				continue
			}
			if !istUngueltig(err) || got != wU {
				t.Errorf("Operator %q: %d %v", m, got, err)
			}
			nOps++
		}
	}
	if nOps != randformenOperatoren {
		t.Errorf("%d Operator-Randformen", nOps)
	}
	// kein Wert im Fehlertext
	geheim := model.Wert{Art: model.ArtDecimal, Text: "GEHEIM9"}
	_, e1 := model.Vergleiche(model.VergleichEq, geheim, geheim)
	_, e2 := model.OrdneWerte(geheim, geheim)
	_, e3 := model.WahrheitVon(model.Wert{Art: model.ArtBoolean, Text: "GEHEIM9"})
	_, e4 := model.WertAusLiteral(model.Literal{Art: model.LiteralZeichenkette, Text: "'GEHEIM9"})
	for _, e := range []error{e1, e2, e3, e4} {
		if e == nil || strings.Contains(e.Error(), "GEHEIM9") {
			t.Errorf("Fehlertext: %v", e)
		}
	}
}

// Zählprüfung der Mengen: Die Konstanten sind die gemessene Größe der Generatoren.
const (
	randformenZahlTypnamen = 2369
	randformenUngueltig    = 7261
	randformenGueltig      = 862
	randformenOperatoren   = 457
)

// Zählprüfung der Randformen der Zeichenketten in TestWertAusLiteral.
const randformenLiteral = 714

package telemetry_test

import (
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// zeile ist eine Zeile der Tabelle in SPEC-230 Nr. 3; menge ist "log" oder der Name des Spans.
type zeile struct {
	key   string
	menge string
	art   support.Art
	form  string
}

// katalog ist die Tabelle in SPEC-230 Nr. 3, Zeile für Zeile aus dem Wortlaut (24 Schlüssel).
func katalog() []zeile {
	s, i, f := support.ArtString, support.ArtInt, support.ArtFloat
	return []zeile{
		{"mode", "log", s, "token"}, {"ereignis", "log", s, "token"}, {"richtung", "log", s, "token"}, {"grund", "log", s, "token"},
		{"service_id", "log", s, "kennung"}, {"session_id", "log", s, "kennung"}, {"exchange_id", "log", s, "kennung"},
		{"workspace_id", "log", s, "kennung"}, {"correlation_id", "log", s, "kennung"},
		{"error_code", "log", s, "fehlercode"}, {"duration_ms", "log", f, "dauer"}, {"blob", "log", s, "blob"}, {"befunde", "log", i, "zahl"},
		{"http.request.method", "odata.gateway.request", s, "methode"},
		{"http.response.status_code", "odata.gateway.request", i, "status"},
		{"recorder.mode", "odata.gateway.request", s, "token"},
		{"recorder.session_id", "odata.record.capture", s, "kennung"},
		{"recorder.capture_status", "odata.record.capture", s, "token"},
		{"recorder.match_strategy", "odata.replay.match", s, "token"},
		{"recorder.match_result", "odata.replay.match", s, "token"},
		{"recorder.entity_set", "odata.mock.query", s, "name"},
		{"recorder.coverage_status", "odata.mock.query", s, "token"},
		{"recorder.workspace_id", "odata.workspace.mutation", s, "kennung"},
		{"recorder.mutation_type", "odata.workspace.mutation", s, "token"},
	}
}

func spanNamen() []string {
	return []string{
		"odata.gateway.request", "odata.record.capture", "odata.replay.match",
		"odata.mock.query", "odata.workspace.mutation",
	}
}

// alleCodes sind die 33 Texte der Tabelle in SPEC-109 Nr. 2.
func alleCodes() []string {
	return []string{
		"ODR-E2001", "ODR-E2101", "ODR-E2401", "ODR-E3001", "ODR-E4001",
		"ODR-E5001", "ODR-E5002", "ODR-E5003", "ODR-E5004", "ODR-E5005", "ODR-E5006", "ODR-E5007", "ODR-E5008", "ODR-E5009",
		"ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E6004", "ODR-E6101", "ODR-E6102",
		"ODR-E7001", "ODR-E7002", "ODR-E7003", "ODR-E7004", "ODR-E7005", "ODR-E7006",
		"ODR-E7007", "ODR-E7008", "ODR-E7009", "ODR-E7010", "ODR-E7011", "ODR-E7012", "ODR-E9001",
	}
}

func formen() []string {
	return []string{"token", "kennung", "name", "methode", "status", "zahl", "fehlercode", "dauer", "blob"}
}

// schluesselZuForm liefert einen Schlüssel der Tabelle mit dieser Form samt seiner Menge und Art.
func schluesselZuForm(t *testing.T, form string) zeile {
	t.Helper()
	for _, z := range katalog() {
		if z.form == form {
			return z
		}
	}
	t.Fatalf("kein Schlüssel mit der Form %s", form)
	return zeile{}
}

func menge(z zeile) string {
	if z.menge == "log" {
		return ""
	}
	return z.menge
}

// formRegexp ist die Beschreibung der Textformen aus dem Wortlaut von SPEC-230 Nr. 3.
func formRegexp(form string) *regexp.Regexp {
	muster := map[string]string{
		"token":   `^[a-z][a-z0-9_]{0,31}$`,
		"kennung": `^[A-Za-z0-9._-]{1,64}$`,
		"name":    `^[A-Za-z_][A-Za-z0-9_.]{0,127}$`,
		"methode": `^[A-Z]{1,16}$`,
		"blob":    `^sha256:[0-9a-f]{64}$`,
	}
	return regexp.MustCompile(muster[form])
}

// orakel entscheidet aus der Beschreibung der Form, nicht aus dem Code des Adapters.
func orakel(z zeile, a support.Attr) bool {
	if a.Art != z.art {
		return false
	}
	switch z.form {
	case "status":
		return a.Ganz >= 100 && a.Ganz <= 599
	case "zahl":
		return a.Ganz >= 0
	case "dauer":
		return !math.IsNaN(a.Gleit) && !math.IsInf(a.Gleit, 0) && a.Gleit >= 0 && !math.Signbit(a.Gleit)
	case "fehlercode":
		for _, c := range alleCodes() {
			if a.Text == c {
				return true
			}
		}
		return false
	}
	return formRegexp(z.form).MatchString(a.Text)
}

func verwuerfe(op string, a support.Attr) []support.Verwurf {
	_, v := telemetry.Filtern(op, []support.Attr{a})
	return v
}

func pruefeFall(t *testing.T, z zeile, a support.Attr, want bool, was string) {
	t.Helper()
	angenommen, v := telemetry.Filtern(menge(z), []support.Attr{a})
	if want {
		if len(v) != 0 || len(angenommen) != 1 || angenommen[0] != a {
			t.Errorf("%s %s: angenommen %v, verworfen %v, erwartet angenommen", z.form, was, angenommen, v)
		}
		return
	}
	if len(angenommen) != 0 || len(v) != 1 || v[0] != support.VerwurfWert {
		t.Errorf("%s %s: angenommen %v, verworfen %v, erwartet genau %q", z.form, was, angenommen, v, support.VerwurfWert)
	}
}

func attrZuArt(z zeile, art support.Art, text string, ganz int64, gleit float64) support.Attr {
	return support.Attr{Key: z.key, Art: art, Text: text, Ganz: ganz, Gleit: gleit}
}

// P02: die Mengen der Schlüssel sind geschlossen.
func TestSchluesselmengen(t *testing.T) {
	logIn, spanIn, paare := 0, 0, 0
	for _, z := range katalog() {
		gueltig := gueltigerWert(z)
		// Menge Log
		_, v := telemetry.Filtern("", []support.Attr{gueltig})
		paare++
		if z.menge == "log" {
			logIn++
			if len(v) != 0 {
				t.Errorf("%s in Log: verworfen %v", z.key, v)
			}
		} else if len(v) != 1 || v[0] != support.VerwurfSchluessel {
			t.Errorf("%s in Log: verworfen %v, erwartet schluessel", z.key, v)
		}
		// Menge Span: je Span-Name
		for _, name := range spanNamen() {
			_, v := telemetry.Filtern(name, []support.Attr{gueltig})
			paare++
			if z.menge == name {
				spanIn++
				if len(v) != 0 {
					t.Errorf("%s in %s: verworfen %v", z.key, name, v)
				}
			} else if len(v) != 1 || v[0] != support.VerwurfSchluessel {
				t.Errorf("%s in %s: verworfen %v, erwartet schluessel", z.key, name, v)
			}
		}
	}
	if len(katalog()) != 24 || paare != 24+24*5 || logIn != 13 || spanIn != 11 {
		t.Errorf("Zählprüfung: %d Schlüssel, %d Paare, %d in Log, %d in Spans; erwartet 24, 144, 13, 11", len(katalog()), paare, logIn, spanIn)
	}
	fremd := []string{"timestamp", "level", "msg", "Mode", "", "mode "}
	for _, k := range fremd {
		v := verwuerfe("", support.Str(k, "token"))
		if len(v) != 1 || v[0] != support.VerwurfSchluessel {
			t.Errorf("Schlüssel %q: verworfen %v, erwartet schluessel", k, v)
		}
	}
	if len(fremd) != 6 {
		t.Errorf("%d fremde Schlüssel, erwartet 6", len(fremd))
	}
	for _, name := range []string{"odata.gateway.request ", "ODATA.GATEWAY.REQUEST", "odata.unbekannt"} {
		if v := verwuerfe(name, support.Str("recorder.mode", "record")); len(v) != 1 || v[0] != support.VerwurfSchluessel {
			t.Errorf("Span %q: verworfen %v, erwartet schluessel (Name außerhalb der Menge)", name, v)
		}
	}
}

// gueltigerWert liefert einen Wert, den die Form von z annimmt.
func gueltigerWert(z zeile) support.Attr {
	switch z.form {
	case "token":
		return support.Str(z.key, "record")
	case "kennung":
		return support.Str(z.key, "ses-0001.a_B")
	case "name":
		return support.Str(z.key, "Orders.Item_1")
	case "methode":
		return support.Str(z.key, "GET")
	case "status":
		return support.Int(z.key, 200)
	case "zahl":
		return support.Int(z.key, 0)
	case "fehlercode":
		return support.Str(z.key, "ODR-E2001")
	case "dauer":
		return support.Millis(z.key, 1500*time.Microsecond)
	}
	return support.Str(z.key, "sha256:"+strings.Repeat("a", 64))
}

func textKandidaten(z zeile) []string {
	grenzen := map[string]int{"token": 32, "kennung": 64, "name": 128, "methode": 16}
	fuell := map[string]string{"token": "a", "kennung": "A", "name": "a", "methode": "A"}
	ersteZeichen := map[string][]string{
		"token":   {"a", "1", "_", "A"},
		"kennung": {"A", ".", "_", "-", "9"},
		"name":    {"a", "A", "_", "1", "."},
		"methode": {"A"},
	}
	var out []string
	for _, erst := range ersteZeichen[z.form] {
		for _, n := range []int{0, 1, 2, grenzen[z.form] - 1, grenzen[z.form], grenzen[z.form] + 1} {
			if n == 0 {
				out = append(out, "")
				continue
			}
			out = append(out, erst+strings.Repeat(fuell[z.form], n-1))
		}
	}
	return out
}

// P03: die neun Wertformen nehmen an und weisen ab an beiden Seiten jeder Grenze, bei leerem Wert
// und bei falscher Art.
func TestWertformGrenzen(t *testing.T) {
	gesehen := 0
	for _, form := range formen() {
		gesehen++
		z := schluesselZuForm(t, form)
		var faelle []support.Attr
		switch form {
		case "token", "kennung", "name", "methode":
			for _, s := range textKandidaten(z) {
				faelle = append(faelle, support.Str(z.key, s))
			}
		case "status":
			for _, v := range []int64{99, 100, 101, 598, 599, 600, 0, -1, math.MaxInt64, math.MinInt64} {
				faelle = append(faelle, support.Int(z.key, v))
			}
		case "zahl":
			for _, v := range []int64{-1, 0, 1, math.MaxInt64, math.MinInt64} {
				faelle = append(faelle, support.Int(z.key, v))
			}
		case "dauer":
			for _, v := range []float64{math.Copysign(0, -1), -1e-6, 0, 1e-6, 12.345, math.MaxFloat64, math.NaN(), math.Inf(1), math.Inf(-1), -1} {
				faelle = append(faelle, attrZuArt(z, support.ArtFloat, "", 0, v))
			}
			faelle = append(faelle, support.Millis(z.key, -time.Nanosecond), support.Millis(z.key, 0), support.Millis(z.key, time.Nanosecond))
		case "fehlercode":
			faelle = append(faelle, support.Str(z.key, ""), support.Str(z.key, "ODR-E2001"), support.Str(z.key, "ODR-E9001"), support.Str(z.key, "ODR-E0000"))
		case "blob":
			for _, n := range []int{0, 1, 63, 64, 65} {
				faelle = append(faelle, support.Str(z.key, "sha256:"+strings.Repeat("0", n)))
			}
			faelle = append(faelle, support.Str(z.key, strings.Repeat("0", 71)), support.Str(z.key, "sha256:"+strings.Repeat("A", 64)),
				support.Str(z.key, "sha256:"+strings.Repeat("g", 64)), support.Str(z.key, "SHA256:"+strings.Repeat("a", 64)))
		}
		for _, a := range faelle {
			pruefeFall(t, z, a, orakel(z, a), "Wert")
		}
		// falsche Art: jede der anderen Arten und zwei ungültige Arten, mit Werten, die in ihrer Art plausibel sind
		for _, art := range []support.Art{support.ArtString, support.ArtInt, support.ArtFloat, 0, 99} {
			if art == z.art {
				continue
			}
			pruefeFall(t, z, attrZuArt(z, art, "a", 200, 5), false, "falsche Art")
		}
		// leer
		pruefeFall(t, z, attrZuArt(z, z.art, "", 0, 0), form == "zahl" || form == "dauer", "leer")
	}
	if gesehen != 9 {
		t.Errorf("%d Formen geprüft, erwartet 9", gesehen)
	}
}

// Die Grenzen aus dem Wortlaut, von Hand: beide Seiten jeder Grenze.
func TestWertformGrenzenVonHand(t *testing.T) {
	a := func(n int) string { return strings.Repeat("a", n) }
	faelle := []struct {
		form string
		wert support.Attr
		want bool
	}{
		{"token", support.Str("", a(32)), true}, {"token", support.Str("", a(33)), false},
		{"kennung", support.Str("", a(64)), true}, {"kennung", support.Str("", a(65)), false},
		{"name", support.Str("", a(128)), true}, {"name", support.Str("", a(129)), false},
		{"methode", support.Str("", strings.Repeat("G", 16)), true}, {"methode", support.Str("", strings.Repeat("G", 17)), false},
		{"status", support.Int("", 99), false}, {"status", support.Int("", 100), true},
		{"status", support.Int("", 599), true}, {"status", support.Int("", 600), false},
		{"zahl", support.Int("", -1), false}, {"zahl", support.Int("", 0), true},
		{"dauer", support.Millis("", -time.Nanosecond), false}, {"dauer", support.Millis("", 0), true},
		{"dauer", support.Attr{Art: support.ArtFloat, Gleit: math.Copysign(0, -1)}, false},
		{"dauer", support.Attr{Art: support.ArtFloat, Gleit: 0}, true},
		{"dauer", support.Attr{Art: support.ArtFloat, Gleit: math.NaN()}, false},
		{"dauer", support.Attr{Art: support.ArtFloat, Gleit: math.Inf(1)}, false},
		{"dauer", support.Attr{Art: support.ArtFloat, Gleit: math.Inf(-1)}, false},
		{"blob", support.Str("", "sha256:"+strings.Repeat("0", 63)), false}, {"blob", support.Str("", "sha256:"+strings.Repeat("0", 64)), true},
		{"blob", support.Str("", "sha256:"+strings.Repeat("0", 65)), false},
	}
	for _, f := range faelle {
		z := schluesselZuForm(t, f.form)
		w := f.wert
		w.Key = z.key
		pruefeFall(t, z, w, f.want, "von Hand")
	}
}

// P04: die sechs Formen mit Text nehmen jedes Byte außerhalb ihrer Zeichen nicht an: Schleife über
// alle 256 Bytewerte × 6 Formen, je Wert an jeder Stelle eines gültigen Grundwerts ersetzt, dazu
// vorangestellt und angehängt; das Orakel ist die Beschreibung der Form (regulärer Ausdruck,
// Menge der Codes), nicht der Code des Adapters.
func TestWertformZeichen(t *testing.T) {
	basis := map[string]string{
		"token": "ab12", "kennung": "Ab.1_-9", "name": "Ab.1_x", "methode": "GET",
		"fehlercode": "ODR-E2001", "blob": "sha256:" + strings.Repeat("a", 64),
	}
	werte, abgewiesen, angenommen := 0, 0, 0
	for form, b := range basis {
		z := schluesselZuForm(t, form)
		pruefeFall(t, z, support.Str(z.key, b), true, "Grundwert")
		for v := range 256 {
			c := string([]byte{byte(v)})
			varianten := []string{c + b, b + c}
			for i := range len(b) {
				varianten = append(varianten, b[:i]+c+b[i+1:])
			}
			for _, w := range varianten {
				a := support.Str(z.key, w)
				want := orakel(z, a)
				werte++
				if want {
					angenommen++
				} else {
					abgewiesen++
				}
				pruefeFall(t, z, a, want, "Byte "+strings.ToValidUTF8(c, "?"))
			}
		}
	}
	// 6 Formen × 256 Bytes × (Länge des Grundwerts + 2 Stellen): 4+7+6+3+9+71 = 100, + 12.
	if len(basis) != 6 || werte != 256*(100+12) || abgewiesen == 0 || angenommen == 0 {
		t.Errorf("Zählprüfung: %d Formen, %d Werte (%d angenommen, %d abgewiesen); erwartet 6 und %d Werte", len(basis), werte, angenommen, abgewiesen, 256*(100+12))
	}
}

// P05: error_code nimmt genau die 33 Texte der Tabelle an.
func TestFehlercodeForm(t *testing.T) {
	z := schluesselZuForm(t, "fehlercode")
	codes := alleCodes()
	if len(codes) != 33 {
		t.Fatalf("%d Texte, erwartet 33", len(codes))
	}
	alle := model.AlleCodes()
	if len(alle) != len(codes) {
		t.Fatalf("model.AlleCodes() hat %d Codes, die Tabelle %d", len(alle), len(codes))
	}
	for i, c := range alle {
		if c.String() != codes[i] {
			t.Errorf("Code %d: %q, Tabelle %q", i, c.String(), codes[i])
		}
	}
	for _, c := range codes {
		pruefeFall(t, z, support.Str(z.key, c), true, "Code "+c)
		// Nachbarn: jede Veränderung eines gültigen Textes ist ungültig
		for _, n := range []string{
			strings.ToLower(c), c + "1", c + " ", " " + c, c[:len(c)-1], c[:5], c[:len(c)-1] + "x", "X" + c[1:],
		} {
			pruefeFall(t, z, support.Str(z.key, n), false, "Nachbar "+n)
		}
	}
	for _, f := range []string{"ODR-E0000", "odr-e2001", "ODR-E2001 ", "", "ODR-E20011", "ODR-E", "ODR-E9002", "ODR-E1001"} {
		pruefeFall(t, z, support.Str(z.key, f), false, "fremd "+f)
	}
}

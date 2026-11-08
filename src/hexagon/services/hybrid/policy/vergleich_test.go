package policy_test

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// hostEintraege sind die zwölf Einträge des Rasters P12.
func hostEintraege() []string {
	return []string{
		"h", "H", "h:443", "h:80", "h:8443", "[::1]", "[0:0:0:0:0:0:0:1]", "[::1]:8443", "x",
		"[::ffff:102:304]", "[::ffff:1.2.3.4]", "1.2.3.4",
	}
}

// zielSoll nennt je gültigem Ziel die Indizes der Einträge, die es freigeben: von Hand aus Nr. 5 b
// bis d gelesen, nicht berechnet.
type zielZeile struct {
	ziel    string
	treffer []int
}

func zielSoll() []zielZeile {
	return []zielZeile{
		{"http://h", []int{0, 1, 3}},
		{"https://h", []int{0, 1, 2}},
		{"https://h:443", []int{0, 1, 2}},
		{"https://h:8443", []int{4}},
		{"http://h:80", []int{0, 1, 3}},
		{"http://H", []int{0, 1, 3}},
		{"https://[::1]", []int{5, 6}},
		{"https://[::1]:8443", []int{7}},
		{"https://h/pfad", []int{0, 1, 2}},
		{"https://h?q", []int{0, 1, 2}},
		{"http://h:", []int{0, 1, 3}},
		{"https://[::ffff:1.2.3.4]", []int{9, 10}},
		{"https://1.2.3.4", []int{11}},
		// über das Raster des Plans hinaus: Ziele in nicht kanonischer Form, ein Port mit führender
		// Null (eine Dezimalzahl, Nr. 5 c)
		{"https://[0:0:0:0:0:0:0:1]", []int{5, 6}},
		{"https://[::ffff:102:304]", []int{9, 10}},
		{"https://[::FFFF:1.2.3.4]", []int{9, 10}},
		{"https://h:0443", []int{0, 1, 2}},
		{"HTTP://h", []int{0, 1, 3}},
	}
}

func hostPolicy(t *testing.T, eintrag string) *policy.Policy {
	t.Helper()
	return neu(t, policy.Werte{Upstream: true, Hosts: []string{eintrag}})
}

// TestHostVergleich (P12): 18 gültige Ziele x 12 Einträge mit dem Soll von Hand je Paar.
func TestHostVergleich(t *testing.T) {
	paare, treffer := 0, 0
	for _, z := range zielSoll() {
		for i, e := range hostEintraege() {
			paare++
			want := policy.Bleiben
			if slices.Contains(z.treffer, i) {
				want = policy.Upstream
				treffer++
			}
			if got := ent(hostPolicy(t, e), "GET", "/", z.ziel, e6001); got != want {
				t.Errorf("Ziel %q, Eintrag %q: %s, erwartet %s", z.ziel, e, name(got), name(want))
			}
		}
	}
	if len(hostEintraege()) != 12 || len(zielSoll()) != 18 || paare != 18*12 || treffer != 31+6+3+3 {
		t.Fatalf("%d Einträge, %d Paare, %d Treffer; erwartet 12, 216, 43", len(hostEintraege()), paare, treffer)
	}
}

// nichtLesbar sind die 15 Ziele des Plans, die nichts freigeben, und weitere Formen, die Nr. 5 c
// ausschließt.
func nichtLesbar() []string {
	byte80 := string([]byte{0x80})
	umlaut := string([]byte{0xc3, 0xa4})
	nul := string([]byte{0})
	z := []string{
		"", "h", "ftp://h", "http://", "http://u@h", "http://h:0", "http://h:99999", "http://[::1",
		"http://[::1%25eth0]", "http://h" + byte80, "http://h%41", "http://h.", "http://xh", "http://a.h", "http://h.example",
	}
	return append(z,
		"http://@h", "http://::1", "http://[1.2.3.4]", "http://[h]", "//h", "http:h", "http://h"+umlaut, "http://h"+nul,
		"http://h:80x", "http://h:-1", "http://h:+80", "http://h:100000", "http://[]", " http://h", "http://h ",
		"https://:443", "http://[::1]x", "http://[::1%25eth0]:80",
		// weitere Formen des Benutzerteils (Nr. 5 c: kein Benutzerteil): mit Passwort, mit leerem
		// Passwort, ohne Namen, vor Port und vor IPv6, mit Prozent-Kodierung, mit zweitem `@`
		"http://u:p@h", "http://:p@h", "http://u:@h", "http://u:p@h:443", "https://u:p@[::1]",
		"http://u%40v:p@h", "http://a@b@h", "http://:@h",
	)
}

// TestHostNichtLesbar (P12): ein nicht lesbares oder fremdes Ziel gibt nichts frei, gegen jeden
// einzelnen Eintrag und gegen die Liste aller zwölf.
func TestHostNichtLesbar(t *testing.T) {
	ziele := nichtLesbar()
	alle := neu(t, policy.Werte{Upstream: true, Hosts: hostEintraege()})
	paare := 0
	for _, z := range ziele {
		if got := ent(alle, "GET", "/", z, e6001); got != policy.Bleiben {
			t.Errorf("Ziel %q gegen alle Einträge: %s", z, name(got))
		}
		for _, e := range hostEintraege() {
			paare++
			if got := ent(hostPolicy(t, e), "GET", "/", z, e6001); got != policy.Bleiben {
				t.Errorf("Ziel %q, Eintrag %q: %s", z, e, name(got))
			}
		}
	}
	if len(ziele) != 15+18+8 || paare != 41*12 || !slices.Contains(ziele, "http://h.") {
		t.Fatalf("%d Ziele, %d Paare", len(ziele), paare)
	}
}

// TestHostListeAlsMenge (P12, P01): ein Ziel ist frei, wenn irgendein Eintrag der Liste passt, an
// jeder Stelle der Liste.
func TestHostListeAlsMenge(t *testing.T) {
	for i := range hostEintraege() {
		liste := hostEintraege()
		liste[0], liste[i] = liste[i], liste[0]
		p := neu(t, policy.Werte{Upstream: true, Hosts: liste})
		for _, z := range zielSoll() {
			want := policy.Bleiben
			if len(z.treffer) > 0 {
				want = policy.Upstream
			}
			if got := ent(p, "GET", "/", z.ziel, e6001); got != want {
				t.Errorf("Eintrag %q vorn, Ziel %q: %s", hostEintraege()[i], z.ziel, name(got))
			}
		}
	}
}

// wKlasse ist die Zeichenklasse W (Nr. 6 c) als Klasse eines regulären Ausdrucks, aufgezählt Byte
// für Byte.
func wKlasse() string {
	var b strings.Builder
	b.WriteString("[")
	for c := 0x21; c <= 0x7e; c++ {
		if strings.ContainsRune("/%\\?#", rune(c)) {
			continue
		}
		fmt.Fprintf(&b, `\x%02x`, c)
	}
	b.WriteString("]")
	return b.String()
}

// musterOrakel ist das Orakel zu P15: das Muster als regulärer Ausdruck mit der Klasse W; was ein
// `*` oder `**` ganz deckt, darf nicht `.` oder `..` sein.
type musterOrakel struct {
	re    *regexp.Regexp
	ganze []int // Gruppen, die ein Segment ganz decken (`*` allein oder `**`)
}

func neuesOrakel(muster string) musterOrakel {
	w := wKlasse()
	var o musterOrakel
	var b strings.Builder
	b.WriteString("^")
	gruppe := 0
	for _, s := range strings.Split(muster[1:], "/") {
		b.WriteString("/")
		switch {
		case s == "**":
			gruppe++
			o.ganze = append(o.ganze, gruppe)
			b.WriteString("(" + w + "+(?:/" + w + "+)*)")
		case strings.Contains(s, "*"):
			vor, nach, _ := strings.Cut(s, "*")
			gruppe++
			if vor == "" && nach == "" {
				o.ganze = append(o.ganze, gruppe)
			}
			b.WriteString(regexp.QuoteMeta(vor) + "(" + w + "+)" + regexp.QuoteMeta(nach))
		default:
			b.WriteString(regexp.QuoteMeta(s))
		}
	}
	b.WriteString("$")
	o.re = regexp.MustCompile(b.String())
	return o
}

func (o musterOrakel) passt(pfad string) bool {
	m := o.re.FindStringSubmatch(pfad)
	if m == nil {
		return false
	}
	for _, g := range o.ganze {
		for _, s := range strings.Split(m[g], "/") {
			if s == "." || s == ".." {
				return false
			}
		}
	}
	return true
}

// musterMenge erzeugt die 456 Muster des Plans: Folgen der Länge 1 bis 3 aus sieben Segmenten, das
// letzte aus diesen sieben oder `**`.
func musterMenge() []string {
	sieben := []string{"Orders", "Orders(1)", "*", "Orders(*)", "a*", "*b", "a*a"}
	letzte := append(slices.Clone(sieben), "**")
	var aus []string
	var rec func(pfx string, rest int)
	rec = func(pfx string, rest int) {
		if rest == 1 {
			for _, l := range letzte {
				aus = append(aus, pfx+"/"+l)
			}
			return
		}
		for _, s := range sieben {
			rec(pfx+"/"+s, rest-1)
		}
	}
	for n := 1; n <= 3; n++ {
		rec("", n)
	}
	return aus
}

// pfadMenge erzeugt die 2961 Pfade: Folgen der Länge 0 bis 3 aus vierzehn Segmenten mit führendem
// `/` (1 + 14 + 196 + 2744 = 2955), dazu sechs Pfade ohne führendes `/`. Die zehn Segmente des
// Plans sind um `aa` und `aba` (sonst deckt kein Pfad ein Segment von `a*a`) sowie `xa` und
// `XOrders(1)` ergänzt: sie enthalten den Anfang eines Musters (`a`, `Orders(`), beginnen aber
// nicht damit.
func pfadMenge() []string {
	zehn := []string{"Orders", "Orders(1)", "Orders(1,2)", "ab", "xb", "a", "b", "", ".", "..", "aa", "aba", "xa", "XOrders(1)"}
	aus := []string{"/"}
	var rec func(folge []string, rest int)
	rec = func(folge []string, rest int) {
		if rest == 0 {
			aus = append(aus, "/"+strings.Join(folge, "/"))
			return
		}
		for _, s := range zehn {
			rec(append(slices.Clone(folge), s), rest-1)
		}
	}
	for n := 1; n <= 3; n++ {
		rec(nil, n)
	}
	return append(aus, "", "Orders", "Orders/Orders", "a/b", "ab", "x/y/z")
}

// TestMusterVergleich (P15): 456 Muster x 2961 Pfade gegen das Orakel mit der Klasse W.
func TestMusterVergleich(t *testing.T) {
	muster, pfade := musterMenge(), pfadMenge()
	if len(muster) != 456 || len(pfade) != 2961 {
		t.Fatalf("%d Muster, %d Pfade; erwartet 456 und 2961", len(muster), len(pfade))
	}
	paare, treffer := 0, 0
	for _, m := range muster {
		o := neuesOrakel(m)
		p := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"POST " + m}})
		hits := 0
		for _, pfad := range pfade {
			paare++
			want := o.passt(pfad)
			got := ent(p, "POST", pfad, zielPassend, e6001) == policy.Upstream
			if got != want {
				t.Errorf("Muster %q, Pfad %q: %v, erwartet %v", m, pfad, got, want)
			}
			if want {
				hits++
			}
		}
		if hits == 0 || hits == len(pfade) {
			t.Errorf("Muster %q: %d Treffer von %d, das Orakel unterscheidet nichts", m, hits, len(pfade))
		}
		treffer += hits
	}
	if paare != 456*2961 {
		t.Fatalf("%d Paare", paare)
	}
	if treffer != trefferP15 {
		t.Errorf("%d Treffer, erwartet %d", treffer, trefferP15)
	}
}

// trefferP15 ist die Zahl der Treffer des Orakels über die 1350216 Paare; der Wert steht hier, damit
// ein Orakel, das mit dem Code zugleich bricht, auffällt.
const trefferP15 = 18809

// endungen sind die 17 Pfadendungen von P16.
func endungen() []string {
	return []string{
		"%2e%2e/Geheim", "%2E%2E", "%2e", "..", ".", "a%2Fb", "a%5Cb", "a" + string(rune(92)) + "b", "a?b", "a#b",
		"a//b", "a/", string([]byte{0x80}), string([]byte{0}), string([]byte{0xc3, 0xa4}), "a%zz", "Orders('a%20b')",
	}
}

// TestPfadOhneDekodierung (P16): kein Platzhalter deckt eine der 17 Endungen unter /svc/Orders/;
// ein Muster ohne Platzhalter gleicht byte-gleich.
func TestPfadOhneDekodierung(t *testing.T) {
	faelle := 0
	if len(endungen()) != 17 {
		t.Fatalf("%d Endungen", len(endungen()))
	}
	for _, m := range []string{"/svc/Orders/*", "/svc/Orders/**", "/svc/Orders/Orders(*)"} {
		p := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"PUT " + m}})
		for _, e := range endungen() {
			faelle++
			if got := ent(p, "PUT", "/svc/Orders/"+e, zielPassend, e6002); got != policy.Bleiben {
				t.Errorf("Muster %q, Endung %q: %s", m, e, name(got))
			}
		}
		// Gegenprobe: dasselbe Muster deckt eine gewöhnliche Endung
		gut := "a"
		if strings.HasSuffix(m, "(*)") {
			gut = "Orders(x)"
		}
		if got := ent(p, "PUT", "/svc/Orders/"+gut, zielPassend, e6002); got != policy.Upstream {
			t.Errorf("Muster %q, Endung %q: %s", m, gut, name(got))
		}
	}
	lit := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"PUT /svc/Orders/a%2Fb"}})
	faelle += 2
	if got := ent(lit, "PUT", "/svc/Orders/a%2Fb", zielPassend, e6002); got != policy.Upstream {
		t.Errorf("a%%2Fb: %s", name(got))
	}
	if got := ent(lit, "PUT", "/svc/Orders/a%2fb", zielPassend, e6002); got != policy.Bleiben {
		t.Errorf("a%%2fb: %s", name(got))
	}
	if faelle != 53 {
		t.Fatalf("%d Fälle, erwartet 53", faelle)
	}
}

// TestPfadOhneDekodierungWeitere (P16): weitere Formen mit Prozent-Kodierung treffen weder einen
// Platzhalter noch ein Muster ohne Platzhalter, dessen Text sie dekodiert wären.
func TestPfadOhneDekodierungWeitere(t *testing.T) {
	weitere := []string{"a%2eb", "a%2Eb", "%2e%2e%2e", "a%2fb", "%41", "%2e.", ".%2e", "a%25b"}
	faelle := 0
	for _, m := range []string{"/svc/Orders/*", "/svc/Orders/**", "/svc/Orders/a*b", "/svc/Orders/a.b", "/svc/Orders/A", "/svc/Orders/a/b", "/svc/Orders/..."} {
		p := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"PUT " + m}})
		for _, e := range weitere {
			faelle++
			if got := ent(p, "PUT", "/svc/Orders/"+e, zielPassend, e6002); got != policy.Bleiben {
				t.Errorf("Muster %q, Endung %q: %s", m, e, name(got))
			}
		}
	}
	if faelle != len(weitere)*7 {
		t.Fatalf("%d Fälle", faelle)
	}
}

// TestWMenge (P16, Nr. 6 c): je Byte 0 bis 255 genau dann gedeckt, wenn es zur Menge W gehört
// (0x21 bis 0x7E außer `/`, `%`, `\`, `?`, `#`); `**` deckt dazu mehrere Segmente. 5 Muster x 256 Bytes.
func TestWMenge(t *testing.T) {
	faelle, gedeckt := 0, 0
	for _, m := range []string{"/svc/*", "/svc/x*", "/svc/*y", "/svc/x*y", "/svc/**"} {
		p := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"PUT " + m}})
		for b := 0; b < 256; b++ {
			faelle++
			c := byte(b)
			inW := c >= 0x21 && c <= 0x7e && !strings.ContainsRune("/%\\?#", rune(c))
			want := inW || m == "/svc/**" && c == '/'
			got := ent(p, "PUT", "/svc/x"+string([]byte{c})+"y", zielPassend, e6002) == policy.Upstream
			if got != want {
				t.Errorf("Muster %q, Byte 0x%02x: gedeckt %v, erwartet %v", m, b, got, want)
			}
			gedeckt += b2i(got)
		}
	}
	if faelle != 1280 || gedeckt != 4*89+90 {
		t.Fatalf("%d Fälle, %d gedeckt; erwartet 1280 und 446", faelle, gedeckt)
	}
}

// TestPunktSegmente (P16, Nr. 6 c): ein von `*` oder `**` ganz gedecktes Segment ist weder `.` noch
// `..`; ein Segment mit mehr als dem Punkt ist es nie.
func TestPunktSegmente(t *testing.T) {
	stern := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"PUT /svc/*"}})
	rest := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"PUT /svc/**"}})
	for _, s := range []string{".", ".."} {
		for _, p := range []*policy.Policy{stern, rest} {
			if got := ent(p, "PUT", "/svc/"+s, zielPassend, e6002); got != policy.Bleiben {
				t.Errorf("Segment %q: %s", s, name(got))
			}
		}
	}
	for _, s := range []string{"...", ".a", "a.", "a..", "..a"} {
		for _, p := range []*policy.Policy{stern, rest} {
			if got := ent(p, "PUT", "/svc/"+s, zielPassend, e6002); got != policy.Upstream {
				t.Errorf("Segment %q: %s", s, name(got))
			}
		}
	}
	// ein nur teilweise gedecktes Segment ist nie `.` oder `..`: das Segment mit dem Punkt im
	// gedeckten Teil passt
	teil := []struct{ muster, pfad string }{
		{"/svc/*b", "/svc/..b"}, {"/svc/*b", "/svc/.b"}, {"/svc/a*", "/svc/a.."}, {"/svc/a*", "/svc/a."}, {"/svc/a*b", "/svc/a.b"}, {"/svc/a*b", "/svc/a..b"},
	}
	for _, tp := range teil {
		p := neu(t, policy.Werte{Upstream: true, Hosts: []string{"h"}, Schreibfreigaben: []string{"PUT " + tp.muster}})
		if got := ent(p, "PUT", tp.pfad, zielPassend, e6002); got != policy.Upstream {
			t.Errorf("Muster %q, Pfad %q: %s", tp.muster, tp.pfad, name(got))
		}
	}
	if got := ent(rest, "PUT", "/svc/a/../b", zielPassend, e6002); got != policy.Bleiben {
		t.Errorf("a/../b: %s", name(got))
	}
	if got := ent(rest, "PUT", "/svc/a/b/c", zielPassend, e6002); got != policy.Upstream {
		t.Errorf("a/b/c: %s", name(got))
	}
}

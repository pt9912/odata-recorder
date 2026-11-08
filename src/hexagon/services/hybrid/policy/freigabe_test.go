package policy_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// Das Orakel für einen Eintrag von Schreibfreigaben ist unabhängig von parseFreigabe gebaut:
// reguläre Ausdrücke über die Menge der Bytes, danach die Regeln je Segment.
var (
	musterZeichen = regexp.MustCompile(`^(?:/(?:[!"$&-+\-.0->@-\[\]-~]|%[0-9A-Fa-f]{2})+)+$`)
	punktFolge    = regexp.MustCompile(`^(?:\.|%2[eE]){1,2}$`)
	punktEin      = regexp.MustCompile(`^(?:\.|%2[eE])$`)
)

func orakelMuster(m string) bool {
	if !musterZeichen.MatchString(m) {
		return false
	}
	teile := strings.Split(m, "/")[1:]
	for i, s := range teile {
		sterne := strings.Count(s, "*")
		if s == "**" {
			if i != len(teile)-1 {
				return false
			}
			continue
		}
		if sterne > 1 {
			return false
		}
		if sterne == 0 && punktFolge.MatchString(s) {
			return false
		}
		if sterne == 1 && punktEin.MatchString(strings.Replace(s, "*", "", 1)) {
			return false
		}
	}
	return true
}

func orakelFreigabe(e string) string {
	if e == "" {
		return policy.UrsacheLeer
	}
	teile := strings.SplitN(e, " ", 2)
	if len(teile) == 1 {
		return policy.UrsacheTrenner
	}
	if !slices.Contains([]string{"POST", "PUT", "PATCH", "DELETE", "MERGE"}, teile[0]) {
		return policy.UrsacheMethode
	}
	if !orakelMuster(teile[1]) {
		return policy.UrsacheMuster
	}
	return ""
}

func TestFreigabeEintraege(t *testing.T) {
	var gut []string
	for _, m := range []string{"POST", "PUT", "PATCH", "DELETE", "MERGE"} {
		gut = append(gut, m+" /svc/Orders")
	}
	schlecht := []eintragUrsache{
		{"", policy.UrsacheLeer},
		{"POST", policy.UrsacheTrenner}, {"/svc/x", policy.UrsacheTrenner}, {"GET", policy.UrsacheTrenner},
		{"post /svc/x", policy.UrsacheMethode}, {"GET /svc/x", policy.UrsacheMethode}, {"HEAD /svc/x", policy.UrsacheMethode},
		{"OPTIONS /svc/x", policy.UrsacheMethode}, {" /svc/x", policy.UrsacheMethode},
		{"get x", policy.UrsacheMethode}, // Rangfolge: methode vor muster
		{"POST  /svc/x", policy.UrsacheMuster}, {"POST /svc/x ", policy.UrsacheMuster}, {"POST x", policy.UrsacheMuster},
		{"POST ", policy.UrsacheMuster},
	}
	if len(gut) != 5 || len(schlecht) != 14 {
		t.Fatalf("Menge der Tabelle: %d Gutfälle und %d Verstöße, erwartet 5 und 14", len(gut), len(schlecht))
	}
	for _, e := range gut {
		methode, segs, u := policy.ParseFreigabe(e)
		if u != "" || methode != e[:strings.IndexByte(e, ' ')] || len(segs) != 2 {
			t.Errorf("Gutfall %q: %q %v %q", e, methode, segs, u)
		}
		if u := einzel(t, policy.SchluesselSchreibfreigaben, e); u != "" {
			t.Errorf("Gutfall %q: New meldet %q", e, u)
		}
	}
	for _, s := range schlecht {
		if _, _, u := policy.ParseFreigabe(s.eintrag); u != s.ursache {
			t.Errorf("Verstoß %q: Ursache %q, erwartet %q", s.eintrag, u, s.ursache)
		}
		if u := einzel(t, policy.SchluesselSchreibfreigaben, s.eintrag); u != s.ursache {
			t.Errorf("Verstoß %q: New meldet %q, erwartet %q", s.eintrag, u, s.ursache)
		}
	}
}

func TestMusterEintraege(t *testing.T) {
	lit := func(s string) policy.Segment { return policy.Segment{Art: policy.SegLiteral, Anfang: s} }
	stern := func(a, e string) policy.Segment { return policy.Segment{Art: policy.SegStern, Anfang: a, Ende: e} }
	rest := policy.Segment{Art: policy.SegRest}
	gut := []struct {
		muster string
		segs   []policy.Segment
	}{
		{"/a", []policy.Segment{lit("a")}},
		{"/a/b", []policy.Segment{lit("a"), lit("b")}},
		{"/a/*", []policy.Segment{lit("a"), stern("", "")}},
		{"/a/**", []policy.Segment{lit("a"), rest}},
		{"/a/Orders(*)", []policy.Segment{lit("a"), stern("Orders(", ")")}},
		{"/a/*b", []policy.Segment{lit("a"), stern("", "b")}},
		{"/a/a*", []policy.Segment{lit("a"), stern("a", "")}},
		{"/a/x%2Fy", []policy.Segment{lit("a"), lit("x%2Fy")}},
		{"/a/%2e%2ex", []policy.Segment{lit("a"), lit("%2e%2ex")}},
	}
	schlecht := []string{
		"a/b", "/", "//", "/a//b", "/a/", "/a?", "/a#", "/a\\b", "/a,b", "/a/\x1f", "/a/\x7f", "/a/\x80",
		"/a/%", "/a/%2", "/a/%G0", "/a/%2G", "/a/*b*", "/a/**b", "/a/***", "/**/a", "/a/x**",
		"/a/.", "/a/..", "/a/%2e", "/a/%2E", "/a/.%2e", "/a/%2e%2E", "/a/%2e./b",
	}
	if len(gut) != 9 || len(schlecht) != 28 {
		t.Fatalf("Menge der Tabelle: %d Gutfälle und %d Verstöße, erwartet 9 und 28", len(gut), len(schlecht))
	}
	// Leerzeichen gehört nicht zur Menge (0x20 liegt unter 0x21)
	schlecht = append(schlecht, "/a b", "/a/ ", "/a/b ", "/ a")
	for _, g := range gut {
		segs, ok := policy.ParseMuster(g.muster)
		if !ok || !slices.Equal(segs, g.segs) {
			t.Errorf("Gutfall %q: %v %v, erwartet %v", g.muster, segs, ok, g.segs)
		}
		if u := einzel(t, policy.SchluesselSchreibfreigaben, "POST "+g.muster); u != "" {
			t.Errorf("Gutfall %q: New meldet %q", g.muster, u)
		}
	}
	for _, m := range schlecht {
		if _, ok := policy.ParseMuster(m); ok {
			t.Errorf("Verstoß %q angenommen", m)
		}
		if u := einzel(t, policy.SchluesselSchreibfreigaben, "POST "+m); u != policy.UrsacheMuster {
			t.Errorf("Verstoß %q: New meldet %q, erwartet muster", m, u)
		}
	}
}

// TestMusterStern führt die Klasse der Segmente `A*B` (A und B je null bis zwei Teile aus `.`,
// `%2e`, `%2E`: 13 mal 13 = 169) als Segment in `/a/<S>`, `/<S>/b` und `/a/<S>/**`. Das Orakel
// liest SPEC-261 Nr. 6 b, e: ungültig ist ein Segment genau dann, wenn A und B nicht beide leer
// sind und eine Einsetzung des `*` durch einen oder zwei Punkte nach dem Ersetzen von `%2e` und
// `%2E` `.` oder `..` ergibt (das `*` könnte ein Punkt-Segment decken).
func TestMusterStern(t *testing.T) {
	teile := []string{".", "%2e", "%2E"}
	seiten := []string{""}
	for _, a := range teile {
		seiten = append(seiten, a)
		for _, b := range teile {
			seiten = append(seiten, a+b)
		}
	}
	if len(seiten) != 13 {
		t.Fatalf("Seiten: %d, erwartet 13", len(seiten))
	}
	ersetze := strings.NewReplacer("%2e", ".", "%2E", ".")
	zahl, schlecht := 0, 0
	for _, a := range seiten {
		for _, b := range seiten {
			s := a + "*" + b
			zahl++
			ungueltig := false
			if a != "" || b != "" {
				for _, p := range []string{".", ".."} {
					if r := ersetze.Replace(a + p + b); r == "." || r == ".." {
						ungueltig = true
					}
				}
			}
			if ungueltig {
				schlecht++
			}
			for _, m := range []string{"/a/" + s, "/" + s + "/b", "/a/" + s + "/**"} {
				_, ok := policy.ParseMuster(m)
				if ok == ungueltig {
					t.Errorf("Muster %q: angenommen %v, erwartet %v", m, ok, !ungueltig)
				}
				soll := ""
				if ungueltig {
					soll = policy.UrsacheMuster
				}
				if u := einzel(t, policy.SchluesselSchreibfreigaben, "POST "+m); u != soll {
					t.Errorf("Muster %q: New meldet %q, erwartet %q", m, u, soll)
				}
			}
		}
	}
	if zahl != 169 || schlecht != 6 {
		t.Fatalf("Zählprüfung: %d Segmente, %d Verstöße, erwartet 169 und 6", zahl, schlecht)
	}
	// Handfälle: die sechs Verstöße und fünf Gutfälle
	for _, s := range []string{".*", "*.", "%2e*", "*%2e", "%2E*", "*%2E"} {
		if _, ok := policy.ParseMuster("/svc/" + s); ok {
			t.Errorf("/svc/%s angenommen", s)
		}
	}
	for _, s := range []string{"*", "..*", ".*.", "%2e*%2e", "a.*"} {
		if _, ok := policy.ParseMuster("/svc/" + s); !ok {
			t.Errorf("/svc/%s abgewiesen", s)
		}
	}
}

// TestFreigabeGenerator führt zu fünf Grundformen jedes Byte an jeder Stelle des ganzen Eintrags
// (einfügen, ersetzen, löschen) und vergleicht die Ursache mit dem Orakel.
func TestFreigabeGenerator(t *testing.T) {
	basen := []string{"POST /svc/Orders(*)/**", "DELETE /a/*b/x%2Fy", "MERGE /a/%2e%2ex/*", "PATCH /a/b", "PUT /a/.x"}
	erwartet, gezaehlt, gut, schlecht := 0, 0, 0, 0
	ursachen := map[string]int{}
	for _, b := range basen {
		erwartet += zahlVarianten(b)
		for _, e := range varianten(b) {
			gezaehlt++
			soll := orakelFreigabe(e)
			methode, segs, u := policy.ParseFreigabe(e)
			if u != soll {
				t.Errorf("%q: Ursache %q, erwartet %q", e, u, soll)
				continue
			}
			ursachen[u]++
			if u == "" {
				gut++
				if methode == "" || len(segs) == 0 {
					t.Errorf("%q: Gutfall ohne Wert", e)
				}
			} else {
				schlecht++
				if methode != "" || segs != nil {
					t.Errorf("%q: Verstoß mit Wert", e)
				}
			}
		}
	}
	if gezaehlt != erwartet {
		t.Fatalf("Zahl der Elemente %d, erwartet %d", gezaehlt, erwartet)
	}
	for _, u := range []string{"", policy.UrsacheTrenner, policy.UrsacheMethode, policy.UrsacheMuster} {
		if ursachen[u] < 50 {
			t.Errorf("Ursache %q nur %d Mal erzeugt", u, ursachen[u])
		}
	}
	if gut == 0 || schlecht == 0 {
		t.Fatal("Generator einseitig")
	}
}

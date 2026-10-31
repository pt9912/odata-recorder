package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"testing"
)

// ersteAbweichung meldet die erste Zeile, in der ist von soll abweicht, statt beide ganz auszugeben.
func ersteAbweichung(t *testing.T, was, ist, soll string) {
	t.Helper()
	if ist == soll {
		return
	}
	a, c := strings.Split(ist, "\n"), strings.Split(soll, "\n")
	for i := 0; i < len(a) || i < len(c); i++ {
		var x, y string
		if i < len(a) {
			x = a[i]
		}
		if i < len(c) {
			y = c[i]
		}
		if x != y {
			t.Errorf("%s: Zeile %d: %q statt %q (%d gegen %d Zeilen)", was, i+1, x, y, len(a), len(c))
			return
		}
	}
}

// ausnahmeNamen sind die Namen, die als Pfadkomponente auf Ausnahme geprüft werden (SPEC-233 Nr. 5 a, f):
// die Verzeichnisnamen des Repos, die Namen gängiger Ausnahmen, alle Zeichen [A-Za-z0-9-_] als Name und die
// Abwandlungen der Komponenten von `.harness/baseline` (Präfix, Suffix, Groß-/Kleinschreibung, Abschnitte).
func ausnahmeNamen() []string {
	menge := map[string]bool{}
	for _, n := range strings.Fields(`adapters adr agents baselineimport beispiele blobs blocked bootstrap capture
		carveouts .claude clock commands conventions csdl docs done driven evidence gegenproben .githooks
		groessenmarke .harness harness hexagon hooks httpupstream inbound in-progress jsoncodec main
		maintenance migrations mk model next observations odata open outbound outboundtest plan planning ports
		pruefung record reviews rules schema sensors services session skills spec sqlite src stempel support
		supporttest telemetry testdata tools user vertraege welle-fundament welle-persistenz workspace zeichen
		BEO-REPO baseline .git .github vendor node_modules .idea tmp .cache build dist bin out`) {
		menge[n] = true
	}
	for c := 'a'; c <= 'z'; c++ {
		menge[string(c)], menge[strings.ToUpper(string(c))] = true, true
	}
	for c := '0'; c <= '9'; c++ {
		menge[string(c)] = true
	}
	menge["-"], menge["_"] = true, true
	for _, c := range []string{".harness", "baseline"} {
		for _, v := range []string{c + "x", "x" + c, c + "2", strings.ToUpper(c), strings.ToUpper(c[:2]) + c[2:],
			strings.ToUpper(c[:1]) + c[1:], c[:len(c)-1], c[1:], c + ".", "." + c, c + "~", c[:len(c)/2], c[len(c)/2:]} {
			menge[v] = true
		}
	}
	var l []string
	for n := range menge {
		if n != "" && n != "." && n != ".." {
			l = append(l, n)
		}
	}
	sort.Strings(l)
	return l
}

// TestKeineWeitereAusnahme: außer dem Teilbaum `.harness/baseline` unmittelbar unter der Wurzel wird kein Name
// ausgenommen (SPEC-233 Nr. 5 a, f). Menge: Name (ausnahmeNamen) mal Stelle im Pfad (zehn Eltern-Pfade) mal
// Dateiart (Datei im gleichnamigen Verzeichnis, Datei mit dem Namen selbst). Die Erwartung folgt allein dem
// Pfad: ausgenommen ist, was mit `.harness/baseline/` beginnt. Prüfgrenze: ein Name außerhalb der Menge ist
// nicht geprüft (unbeschränkte Menge); die Menge nennt die Namen des Repos, die gängigen Ausnahmen und jedes
// Ein-Zeichen-Verzeichnis.
func TestKeineWeitereAusnahme(t *testing.T) {
	namen := ausnahmeNamen()
	eltern := []string{"", "a/b/c/", ".harness/", "x/.harness/", "x.harness/", ".harnessx/", ".Harness/",
		".harness/baseline/", "x/.harness/baseline/", ".harness/baseline/.harness/baseline/"}
	if len(namen) != 169 || len(eltern) != 10 {
		t.Fatalf("%d Namen und %d Eltern-Pfade statt 169 und 10", len(namen), len(eltern))
	}
	laeufe, ausgenommen := 0, 0
	for _, e := range eltern {
		for _, datei := range []bool{false, true} {
			pfade := make([]string, 0, len(namen))
			for _, n := range namen {
				if datei {
					pfade = append(pfade, e+n)
				} else {
					pfade = append(pfade, e+n+"/x.md")
				}
			}
			dateien := map[string][]byte{"z.md": b("a\n")}
			var soll []string
			for _, p := range pfade {
				dateien[p] = cat(b("x"), zb(0x00A0))
				if strings.HasPrefix(p, ".harness/baseline/") {
					ausgenommen++
				} else {
					soll = append(soll, p)
				}
			}
			sort.Strings(soll)
			erwartet := ""
			for _, p := range soll {
				erwartet += orakelZeile(p, 1, 2, 0x00A0)
			}
			r := fahre(baum(t, dateien))
			laeufe++
			was := fmt.Sprintf("Eltern %q, Datei %v", e, datei)
			if len(soll) == 0 {
				gleich(t, was, r.out, okZeile(1, 0, 0))
				continue
			}
			ersteAbweichung(t, was, r.err, erwartet+summe(len(soll), len(soll)))
		}
	}
	// Ausgenommen: zwei Eltern-Pfade mit dem Präfix mal zwei Dateiarten mal alle Namen, dazu `.harness/` mit
	// dem Namen `baseline` als Verzeichnis.
	if laeufe != 20 || ausgenommen != 2*2*len(namen)+1 {
		t.Errorf("%d Läufe und %d ausgenommene Pfade statt 20 und %d", laeufe, ausgenommen, 2*2*len(namen)+1)
	}
}

// TestNurAndereEintraege: eine Wurzel ohne Textdatei ist `keine Textdatei geprüft`, Exit 2 (SPEC-233 Nr. 11 b),
// gleich welche Einträge sie trägt. Menge: jede nichtleere Teilmenge der sechs Arten Eintrag (63) mal
// Vielfachheit 1 und 3; mit einer Textdatei daneben stimmen die Zähler *b* und *s*.
func TestNurAndereEintraege(t *testing.T) {
	type art struct {
		name           string
		legen          func(t *testing.T, dir, name string)
		binaer, andere int
	}
	arten := []art{
		{"symlink-haengend", func(t *testing.T, dir, n string) { link(t, "gibt-es-nicht", dir+"/"+n) }, 0, 1},
		{"symlink-auf-verzeichnis", func(t *testing.T, dir, n string) { link(t, ".", dir+"/"+n) }, 0, 1},
		{"pipe", func(t *testing.T, dir, n string) {
			if err := syscall.Mkfifo(dir+"/"+n, 0o600); err != nil {
				t.Fatal(err)
			}
		}, 0, 1},
		{"binaer", func(t *testing.T, dir, n string) { schreibe(t, dir+"/"+n+".bin", []byte{0x00, 0x01}) }, 1, 0},
		{"leeres-verzeichnis", func(t *testing.T, dir, n string) {
			if err := os.Mkdir(dir+"/"+n, 0o755); err != nil {
				t.Fatal(err)
			}
		}, 0, 0},
		{"basis-teilbaum", func(t *testing.T, dir, n string) {
			if err := os.MkdirAll(dir+"/.harness/baseline", 0o755); err != nil {
				t.Fatal(err)
			}
			schreibe(t, dir+"/.harness/baseline/"+n+".md", cat(b("a"), zb(0x00A0)))
		}, 0, 0},
	}
	if len(arten) != 6 {
		t.Fatalf("%d Arten statt 6", len(arten))
	}
	laeufe := 0
	for maske := 1; maske < 1<<len(arten); maske++ {
		for _, viele := range []int{1, 3} {
			for _, mitText := range []bool{false, true} {
				dir := t.TempDir()
				bin, andere, was := 0, 0, ""
				for i, a := range arten {
					if maske&(1<<i) == 0 {
						continue
					}
					was += a.name + " "
					for k := 0; k < viele; k++ {
						a.legen(t, dir, fmt.Sprintf("%s-%d", a.name, k))
					}
					bin, andere = bin+viele*a.binaer, andere+viele*a.andere
				}
				if mitText {
					schreibe(t, dir+"/t.md", b("a\n"))
				}
				r := fahre(dir)
				laeufe++
				was = fmt.Sprintf("%sx%d, Text %v", was, viele, mitText)
				if mitText {
					gleich(t, was, r.out, okZeile(1, bin, andere))
					gleich(t, was+" stderr", r.err, "")
					continue
				}
				if r.code != 2 || r.out != "" || r.err != "zeichen: Eingabe: keine Textdatei geprüft\n" {
					t.Errorf("%s: Exit %d, stdout %q, stderr %q", was, r.code, r.out, r.err)
				}
			}
		}
	}
	if laeufe != 63*2*2 {
		t.Errorf("%d Läufe statt %d", laeufe, 63*2*2)
	}
}

func schreibe(t *testing.T, pfad string, daten []byte) {
	t.Helper()
	if err := os.WriteFile(pfad, daten, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestSpalteHinterZeichen: die Spalte zählt jedes gültig dekodierte Zeichen und jedes ungültige Byte als 1,
// gleich wie viele Bytes es hat (SPEC-233 Nr. 9). Menge: vorangehendes Zeichen (17: Tabulator, Leerzeichen,
// ASCII, zwei 2-Byte-, vier 3-Byte-, zwei 4-Byte-Zeichen, U+FFFD, fünf ungültige Folgen) mal Anzahl (1, 2, 5)
// mal Zeichen von Z (253) mal Leser (4).
func TestSpalteHinterZeichen(t *testing.T) {
	type vor struct {
		name      string
		daten     []byte
		spalten   int
		ungueltig []byte
	}
	elemente := []vor{
		{"tabulator", b("\t"), 1, nil}, {"leerzeichen", b(" "), 1, nil}, {"ascii", b("a"), 1, nil},
		{"umlaut", b("\xC3\xA4"), 1, nil}, {"U+034F", zb(0x034F), 1, nil},
		{"strich", zb(0x2013), 1, nil}, {"euro", zb(0x20AC), 1, nil}, {"cjk", zb(0x4E2D), 1, nil}, {"U+FE0F", zb(0xFE0F), 1, nil},
		{"U+10000", zb(0x10000), 1, nil}, {"emoji", zb(0x1F600), 1, nil},
		{"U+FFFD", zb(0xFFFD), 1, nil},
		{"ungueltig-80", []byte{0x80}, 1, []byte{0x80}}, {"ungueltig-FF", []byte{0xFF}, 1, []byte{0xFF}},
		{"abgebrochen-C2", []byte{0xC2}, 1, []byte{0xC2}},
		{"abgebrochen-E2-82", []byte{0xE2, 0x82}, 2, []byte{0xE2, 0x82}},
		{"abgebrochen-F0-9F-98", []byte{0xF0, 0x9F, 0x98}, 3, []byte{0xF0, 0x9F, 0x98}},
	}
	anzahlen := []int{1, 2, 5}
	if len(elemente) != 17 || len(anzahlen) != 3 || len(zElemente()) != 253 || len(leserFormen()) != 4 {
		t.Fatalf("Mengen: %d, %d, %d, %d", len(elemente), len(anzahlen), len(zElemente()), len(leserFormen()))
	}
	faelle := 0
	for _, e := range elemente {
		for _, k := range anzahlen {
			for _, z := range zElemente() {
				var daten []byte
				soll := ""
				for j := 0; j < k; j++ {
					daten = cat(daten, e.daten)
					for i, c := range e.ungueltig {
						soll += fmt.Sprintf("zeichen: d:1:%d: ungültiges UTF-8 (Byte 0x%02X)\n", 1+j*e.spalten+i, c)
					}
				}
				daten = cat(daten, zb(z))
				soll += orakelZeile("d", 1, 1+k*e.spalten, z)
				for _, l := range leserFormen() {
					faelle++
					if _, got := scanne(l.neu(daten)); got != soll {
						t.Fatalf("%s x%d vor U+%04X, %s: %q statt %q", e.name, k, z, l.name, got, soll)
					}
				}
			}
		}
	}
	if faelle != 17*3*253*4 {
		t.Errorf("%d Fälle statt %d", faelle, 17*3*253*4)
	}
}

// TestSpalten: die Spalte eines Befunds in einer Datei hinter Vorgängern verschiedener Art (SPEC-233 Nr. 9):
// 14 Vorgänger mal Anzahl 1, 2, 3 mal Ziel U+00A0 und U+E0001 = 84 Dateien. Die Datei trägt `q`, LF, dann
// Vorgänger mal Anzahl und das Ziel; Zeichen von Z und das ungültige Byte unter den Vorgängern sind selbst Befunde.
func TestSpalten(t *testing.T) {
	type vor struct {
		daten []byte
		soll  func(spalte int) string // Befund-Zeile des Vorgängers in Zeile 2 oder ""
	}
	z := func(r rune) func(int) string { return func(s int) string { return orakelZeile("@", 2, s, r) } }
	keiner := func(int) string { return "" }
	vorgaenger := []vor{
		{b("a"), keiner}, {b("\t"), keiner}, {b(" "), keiner}, {zb(0xE4), keiner}, {zb(0x2013), keiner}, {zb(0x20AC), keiner},
		{zb(0x1F600), keiner}, {zb(0xFFFD), keiner},
		{[]byte{0xFF}, func(s int) string { return fmt.Sprintf("zeichen: @:2:%d: ungültiges UTF-8 (Byte 0xFF)\n", s) }},
		{zb(0x0D), z(0x0D)}, {zb(0x00A0), z(0x00A0)}, {zb(0xE0001), z(0xE0001)},
		{zb(0x0301), keiner}, {zb(0x4E2D), keiner},
	}
	ziele := []rune{0x00A0, 0xE0001}
	if len(vorgaenger) != 14 {
		t.Fatalf("%d Vorgänger statt 14", len(vorgaenger))
	}
	dateien := map[string][]byte{}
	erwartet := map[string]string{}
	var namen []string
	k, d := 0, 0
	for i, v := range vorgaenger {
		for n := 1; n <= 3; n++ {
			for _, ziel := range ziele {
				name := fmt.Sprintf("v%02d-%d-%X.md", i, n, ziel)
				daten, soll := b("q\n"), ""
				for j := 0; j < n; j++ {
					daten = cat(daten, v.daten)
					soll += strings.ReplaceAll(v.soll(1+j), "@", name)
				}
				daten = cat(daten, zb(ziel))
				soll += orakelZeile(name, 2, 1+n, ziel)
				dateien[name], erwartet[name] = daten, soll
				namen = append(namen, name)
				k, d = k+strings.Count(soll, "\n"), d+1
			}
		}
	}
	if len(namen) != 84 {
		t.Fatalf("%d Dateien statt 84", len(namen))
	}
	sort.Strings(namen)
	soll := ""
	for _, n := range namen {
		soll += erwartet[n]
	}
	ersteAbweichung(t, "Spalten", fahre(baum(t, dateien)).err, soll+summe(k, d))
}

// maskiertOrakel ist der Pfad nach SPEC-233 Nr. 10 b, unabhängig von maskiere gebaut.
func maskiertOrakel(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\\':
			sb.WriteString(`\\`)
		case c < 0x20 || c > 0x7E:
			fmt.Fprintf(&sb, `\x%02X`, c)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

// TestSortiergrenze: die Folge der Pfade ist bytegleich; `/` ist das Byte 0x2F (SPEC-233 Nr. 12 a). Menge: jedes
// Byte x außer NUL und `/` (254) als Zeichen hinter `a` mal ohne und mit Verzeichnis-Präfix (2). Je Lauf drei
// Pfade `a<x>k.md`, `a/m.md`, `a<x>z.md`: `a/m.md` steht vor beiden, wenn x größer als `/` ist, hinter beiden,
// wenn x kleiner ist; jede andere Einordnung von `/` setzt den Pfad dazwischen.
func TestSortiergrenze(t *testing.T) {
	laeufe := 0
	for x := 1; x < 256; x++ {
		if x == '/' {
			continue
		}
		for _, pre := range []string{"", "d/"} {
			c := string([]byte{byte(x)})
			klein, mitte, gross := pre+"a"+c+"k.md", pre+"a/m.md", pre+"a"+c+"z.md"
			reihe := []string{mitte, klein, gross}
			if x < '/' {
				reihe = []string{klein, gross, mitte}
			}
			soll := ""
			for _, p := range reihe {
				soll += orakelZeile(maskiertOrakel(p), 1, 2, 0x00A0)
			}
			dateien := map[string][]byte{}
			for _, p := range reihe {
				dateien[p] = cat(b("x"), zb(0x00A0))
			}
			r := fahre(baum(t, dateien))
			laeufe++
			gleich(t, fmt.Sprintf("Byte 0x%02X, Präfix %q", x, pre), r.err, soll+summe(3, 3))
		}
	}
	if laeufe != 254*2 {
		t.Errorf("%d Läufe statt %d", laeufe, 254*2)
	}
}

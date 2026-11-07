package telemetry_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// L01: Menge M1 (Folge, Stelle): alle 2^13 Teilmengen der 13 Felder × 2 Folgen des Aufrufs
// (Katalogfolge, umgekehrt), jeder Aufruf zweimal. Orakel: der Token-Leser (Schlüssel in
// Dateifolge) gegen das Modell; ein nicht gesetztes Feld fehlt.
func TestZeilenform(t *testing.T) {
	folge := logSchluessel(t)
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	var faelle [][]string
	for maske := range 1 << len(folge) {
		for _, umgekehrt := range []bool{false, true} {
			var attrs []support.Attr
			var keys []string
			for i, k := range folge {
				if maske>>i&1 == 1 {
					attrs = append(attrs, gueltigFuer(k))
					keys = append(keys, k)
				}
			}
			if umgekehrt {
				slices.Reverse(attrs)
			}
			a.Log(context.Background(), slog.LevelInfo, "m", attrs...)
			a.Log(context.Background(), slog.LevelInfo, "m", attrs...)
			faelle = append(faelle, keys)
		}
	}
	zeilen := senke.zeilen()
	if len(faelle) != 2<<len(folge) || len(zeilen) != 2*len(faelle) {
		t.Fatalf("%d Fälle, %d Zeilen; erwartet %d und %d", len(faelle), len(zeilen), 2<<len(folge), 4<<len(folge))
	}
	for i, keys := range faelle {
		if !bytes.Equal(zeilen[2*i], zeilen[2*i+1]) {
			t.Fatalf("Fall %d: zwei gleiche Einträge ergeben zwei verschiedene Zeilen", i)
		}
		g := lies(t, zeilen[2*i])
		if want := append([]string{"timestamp", "level", "msg"}, keys...); !slices.Equal(g.schluessel, want) {
			t.Fatalf("Fall %d: Schlüssel %v, erwartet %v", i, g.schluessel, want)
		}
		if g.text["timestamp"] != festerStempel || g.text["level"] != "INFO" || g.text["msg"] != "m" {
			t.Fatalf("Fall %d: Kopf %v", i, g.text)
		}
		for _, k := range keys {
			if g.text[k] != erwarteterText(gueltigFuer(k)) {
				t.Fatalf("Fall %d: %s = %q, erwartet %q", i, k, g.text[k], erwarteterText(gueltigFuer(k)))
			}
		}
	}
	for g, n := range a.Snapshot().Dropped {
		if n != 0 {
			t.Errorf("Grund %s zählt %d bei lauter gültigen Werten", g, n)
		}
	}
}

// L02: Menge M5 (8 Zustände der Zeit): neun Nachkommastellen in UTC, Now einmal je Zeile. Die
// Erwartung steht je Fall als Zeichenkette und als time.Format.
func TestTimestamp(t *testing.T) {
	plus := time.FixedZone("plus", 5*3600+30*60)
	minus := time.FixedZone("minus", -8*3600)
	d := func(n int) time.Time { return time.Date(2026, 10, 30, 12, 34, 56, n, time.UTC) }
	faelle := []struct {
		t    time.Time
		want string
	}{
		{d(0), "2026-10-30T12:34:56.000000000Z"}, {d(1), "2026-10-30T12:34:56.000000001Z"},
		{d(10), "2026-10-30T12:34:56.000000010Z"}, {d(123456789), "2026-10-30T12:34:56.123456789Z"},
		{d(999999999), "2026-10-30T12:34:56.999999999Z"},
		{time.Date(2026, 10, 30, 12, 0, 0, 500000000, plus), "2026-10-30T06:30:00.500000000Z"},
		{time.Date(2026, 12, 31, 20, 0, 0, 0, minus), "2027-01-01T04:00:00.000000000Z"},
		{time.Time{}, "0001-01-01T00:00:00.000000000Z"},
	}
	if len(faelle) != 8 {
		t.Fatalf("%d Fälle, erwartet 8", len(faelle))
	}
	for _, f := range faelle {
		senke, uhr := &FakeSenke{}, &zaehlUhr{t: f.t}
		a := neuMit(t, senke, uhr, slog.LevelInfo)
		a.Log(context.Background(), slog.LevelInfo, "m")
		z := senke.zeilen()
		if len(z) != 1 || uhr.n.Load() != 1 {
			t.Fatalf("%v: %d Zeilen, %d Aufrufe von Now; erwartet je 1", f.t, len(z), uhr.n.Load())
		}
		if got := lies(t, z[0]).text["timestamp"]; got != f.want || got != f.t.UTC().Format("2006-01-02T15:04:05.000000000Z") {
			t.Errorf("%v: timestamp %q, erwartet %q", f.t, got, f.want)
		}
	}
}

func stufenWerte() []int { return []int{-8, -4, -1, 0, 3, 4, 7, 8, 12} }

// stufenGrenzen prüft die Zählprüfung der Menge M4: an jeder Grenze stehen g-1 und g in der Liste.
func stufenGrenzen(t *testing.T) []int {
	t.Helper()
	werte := stufenWerte()
	for _, g := range []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		if !slices.Contains(werte, int(g)-1) || !slices.Contains(werte, int(g)) {
			t.Fatalf("die Liste der Stufen nennt %d oder %d nicht", g-1, g)
		}
	}
	return werte
}

// L03: Menge M4 (neun Stufen, beide Seiten jeder der drei Grenzen): der Name der Stufe nach der
// Tabelle der Spezifikation.
func TestStufen(t *testing.T) {
	for _, l := range stufenGrenzen(t) {
		senke := &FakeSenke{}
		a := neuMit(t, senke, festeUhr(), slog.Level(-8))
		a.Log(context.Background(), slog.Level(l), "m")
		z := senke.zeilen()
		if len(z) != 1 || lies(t, z[0]).text["level"] != modellStufe(l) {
			t.Errorf("Stufe %d: %q, erwartet eine Zeile mit %s", l, z, modellStufe(l))
		}
	}
}

// L04: Menge M4 (81 Paare Stufe × MinLevel), jedes mit einem gültigen und einem ungültigen Attribut:
// eine Zeile genau dann, wenn Stufe >= Minimum; Now einmal oder nie; der Verwurf zählt immer. Dazu
// MinLevel 1 mit den Stufen 0 bis 3 (verglichen wird der Wert, nicht der Name).
func TestMinLevel(t *testing.T) {
	werte := stufenGrenzen(t)
	paare := 0
	pruefe := func(stufe, mindest int) {
		paare++
		senke, uhr := &FakeSenke{}, festeUhr()
		a := neuMit(t, senke, uhr, slog.Level(mindest))
		a.Log(context.Background(), slog.Level(stufe), "m", support.Str("session_id", "s1"), support.Str("mode", "a b"))
		geschrieben := int64(0)
		if stufe >= mindest {
			geschrieben = 1
		}
		if int64(len(senke.zeilen())) != geschrieben || uhr.n.Load() != geschrieben {
			t.Errorf("Stufe %d, MinLevel %d: %d Zeilen, %d Aufrufe von Now; erwartet je %d", stufe, mindest, len(senke.zeilen()), uhr.n.Load(), geschrieben)
		}
		if w := a.Snapshot().Dropped[support.VerwurfWert]; w != 1 {
			t.Errorf("Stufe %d, MinLevel %d: wert %d, erwartet 1", stufe, mindest, w)
		}
	}
	for _, s := range werte {
		for _, m := range werte {
			pruefe(s, m)
		}
	}
	if paare != len(werte)*len(werte) {
		t.Fatalf("%d Paare, erwartet %d", paare, len(werte)*len(werte))
	}
	for s := range 4 {
		pruefe(s, 1)
	}
}

// msgFaelle schreibt je Element roh eine Zeile mit msg gleich roh und prüft sie: Rohbytes der Zeile
// (ein einziges Zeilenende, kein Byte unter 0x20 sonst, keine Folge der Zeilentrenner, gültiges UTF-8,
// die Zeichen <, >, & und DEL so oft wie im Modell) und gelesener msg gleich dem Modell.
func msgFaelle(t *testing.T, menge []string) {
	t.Helper()
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	for _, roh := range menge {
		a.Log(context.Background(), slog.LevelInfo, roh)
	}
	zeilen := senke.zeilen()
	if len(zeilen) != len(menge) {
		t.Fatalf("%d Zeilen, erwartet %d", len(zeilen), len(menge))
	}
	for i, roh := range menge {
		z, want := zeilen[i], modellMsg(roh)
		if !utf8.Valid(z) || z[len(z)-1] != '\n' {
			t.Fatalf("Element %d (%x): Zeile ohne gültiges UTF-8 oder Zeilenende: %q", i, roh, z)
		}
		for _, b := range z[:len(z)-1] {
			if b < 0x20 {
				t.Fatalf("Element %d (%x): Byte %#x in der Zeile", i, roh, b)
			}
		}
		if bytes.Contains(z, []byte{0xE2, 0x80, 0xA8}) || bytes.Contains(z, []byte{0xE2, 0x80, 0xA9}) {
			t.Fatalf("Element %d (%x): Zeilentrenner unmaskiert", i, roh)
		}
		for _, c := range []byte{'<', '>', '&', 0x7F} {
			if bytes.Count(z, []byte{c}) != strings.Count(want, string([]byte{c})) {
				t.Fatalf("Element %d (%x): Byte %#x steht %d-mal, im Text %d-mal", i, roh, c, bytes.Count(z, []byte{c}), strings.Count(want, string([]byte{c})))
			}
		}
		if got := lies(t, z).text["msg"]; got != want {
			t.Fatalf("Element %d (%x): msg %x, erwartet %x", i, roh, got, want)
		}
	}
}

// L05: Menge M2 (Länge, Zeichen): ASCII von 0 bis 300 Bytes; ein Zeichen der Breite 1 bis 4 an
// Offset 250 bis 257, danach kein oder ein Zeichen der Breite 1 bis 4; ein Byte 0x80 an Offset 250
// bis 257. Orakel: das Modell. 256 ist die Konstante der Spezifikation.
func TestMsgLaenge(t *testing.T) {
	const grenze = 256
	var menge []string
	for n := 0; n <= 300; n++ {
		menge = append(menge, strings.Repeat("a", n))
	}
	breiten := []rune{'x', 0xE9, 0x20AC, 0x1F600}
	folgen := append([]string{""}, "x", string(rune(0xE9)), string(rune(0x20AC)), string(rune(0x1F600)))
	for _, r := range breiten {
		for off := grenze - 6; off <= grenze+1; off++ {
			for _, danach := range folgen {
				menge = append(menge, strings.Repeat("a", off)+string(r)+danach)
			}
		}
	}
	for off := grenze - 6; off <= grenze+1; off++ {
		menge = append(menge, strings.Repeat("a", off)+string([]byte{0x80}))
	}
	if want := 301 + len(breiten)*8*len(folgen) + 8; len(menge) != 469 || len(menge) != want {
		t.Fatalf("%d Elemente, erwartet 469", len(menge))
	}
	msgFaelle(t, menge)
}

// mengeZeichen bildet M3 aus Bytewerten.
func mengeZeichen(t *testing.T) []string {
	t.Helper()
	b := []byte{0x00, 0x22, 0x5C, 0x61, 0x7F, 0x80, 0x8F, 0x90, 0x9F, 0xA0, 0xBF, 0xC0, 0xC2, 0xF5, 0xFF}
	f := []byte{0x80, 0x8F, 0x90, 0xBF}
	var menge []string
	einzeln := 0
	for v := range 256 {
		menge = append(menge, string([]byte{byte(v)}))
		einzeln++
	}
	for e := range 256 {
		for _, z := range b {
			menge = append(menge, string([]byte{byte(e), z}))
		}
	}
	for e := 0xE0; e <= 0xEF; e++ {
		for _, z1 := range b {
			for _, z2 := range b {
				menge = append(menge, string([]byte{byte(e), z1, z2}))
			}
		}
	}
	for e := 0xF0; e <= 0xF4; e++ {
		for _, z1 := range b {
			for _, z2 := range f {
				for _, z3 := range f {
					menge = append(menge, string([]byte{byte(e), z1, z2, z3}))
				}
			}
		}
	}
	benannt := []rune{0x80, 0x7FF, 0x800, 0xD7FF, 0xE000, 0xFFFF, 0x10000, 0x10FFFF, 0x2028, 0x2029, 0xFEFF, 0xA0, 0xFFFD}
	for _, r := range benannt {
		menge = append(menge, string(r))
	}
	if want := einzeln + 256*len(b) + 16*len(b)*len(b) + 5*len(b)*len(f)*len(f) + len(benannt); einzeln != 256 || len(b) != 15 || len(menge) != want || want != 8909 {
		t.Fatalf("%d Elemente, erwartet %d", len(menge), want)
	}
	return menge
}

// L06: Menge M3 (8909 Folgen von Bytes als msg): Maskierung allein von ", Backslash, Zeichen unter
// U+0020, U+2028 und U+2029, je ungültigem Byte ein U+FFFD. Fixtures aus Bytewerten.
func TestMsgZeichen(t *testing.T) { msgFaelle(t, mengeZeichen(t)) }

// L07: Menge M6 (14 Werte): duration_ms wie FormatFloat('f', 3), ohne Exponent; befunde als Ganzzahl.
func TestZahlen(t *testing.T) {
	dauern := []struct {
		v    float64
		want string
	}{
		{0, "0.000"}, {5e-324, "0.000"}, {0.0005, "0.001"}, {0.9995, "1.000"}, {1.0005, "1.000"},
		{12.345, "12.345"}, {5000, "5000.000"}, {1e21, "1" + strings.Repeat("0", 21) + ".000"}, {math.MaxFloat64, ""},
	}
	ganze := []int64{0, 1, 9, 10, math.MaxInt64}
	if len(dauern)+len(ganze) != 14 {
		t.Fatalf("%d Werte, erwartet 14", len(dauern)+len(ganze))
	}
	dezimal, ganz := regexp.MustCompile(`^[0-9]+\.[0-9]{3}$`), regexp.MustCompile(`^(0|[1-9][0-9]*)$`)
	zeile := func(a support.Attr) gelesen {
		senke := &FakeSenke{}
		neuMit(t, senke, festeUhr(), slog.LevelInfo).Log(context.Background(), slog.LevelInfo, "m", a)
		return lies(t, senke.zeilen()[0])
	}
	for _, d := range dauern {
		g := zeile(support.Attr{Key: "duration_ms", Art: support.ArtFloat, Gleit: d.v})
		want := strconv.FormatFloat(d.v, 'f', 3, 64)
		if d.want != "" && d.want != want {
			t.Fatalf("Erwartung von Hand %q, FormatFloat %q", d.want, want)
		}
		if got := g.text["duration_ms"]; got != want || !g.zahl["duration_ms"] || !dezimal.MatchString(got) {
			t.Errorf("duration_ms %v: %q (Zahl %t), erwartet %q", d.v, got, g.zahl["duration_ms"], want)
		}
	}
	for _, n := range ganze {
		g := zeile(support.Int("befunde", n))
		if got := g.text["befunde"]; got != strconv.FormatInt(n, 10) || !g.zahl["befunde"] || !ganz.MatchString(got) {
			t.Errorf("befunde %d: %q (Zahl %t)", n, got, g.zahl["befunde"])
		}
	}
}

// L08: Menge M7 (Folge, Stelle): 13 Schlüssel × 14 Positionen des zweiten Werts; beide gültig und
// verschieden. Orakel: der Wert, der in der Liste des Aufrufs zuerst steht; schluessel zählt 1.
func TestDoppelterSchluessel(t *testing.T) {
	folge := logSchluessel(t)
	n := 0
	for _, k := range folge {
		for pos := 0; pos <= len(folge); pos++ {
			var z zeile
			for _, kz := range katalog() {
				if kz.key == k {
					z = kz
				}
			}
			var attrs []support.Attr
			for _, f := range folge {
				attrs = append(attrs, gueltigFuer(f))
			}
			zweiter := andererWert(z)
			if erwarteterText(zweiter) == erwarteterText(gueltigFuer(k)) {
				t.Fatalf("%s: beide Werte gleich", k)
			}
			attrs = slices.Insert(attrs, pos, zweiter)
			n++
			senke := &FakeSenke{}
			a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
			a.Log(context.Background(), slog.LevelInfo, "m", attrs...)
			g := lies(t, senke.zeilen()[0])
			erster := slices.IndexFunc(attrs, func(x support.Attr) bool { return x.Key == k })
			if got, want := g.text[k], erwarteterText(attrs[erster]); got != want {
				t.Errorf("%s, zweiter Wert an Stelle %d: %q, erwartet %q", k, pos, got, want)
			}
			if len(g.schluessel) != 3+len(folge) {
				t.Errorf("%s, Stelle %d: Schlüssel %v", k, pos, g.schluessel)
			}
			if s := a.Snapshot().Dropped; s[support.VerwurfSchluessel] != 1 || s[support.VerwurfWert] != 0 {
				t.Errorf("%s, Stelle %d: schluessel %d, wert %d; erwartet 1 und 0", k, pos, s[support.VerwurfSchluessel], s[support.VerwurfWert])
			}
		}
	}
	if n != 13*14 {
		t.Fatalf("%d Fälle, erwartet %d", n, 13*14)
	}
}

// L09: Menge M8 (Abbruch, Argumentart): 3 ctx (nil, beendet, abgelaufene Frist) × 6 Arten. Orakel:
// eine Zeile mit timestamp, level, msg und den gültigen Feldern; Verwürfe gleich dem Modell; eine
// Panik färbt den Fall.
func TestLogArgumente(t *testing.T) {
	var kein context.Context
	beendet, abbruch := context.WithCancel(context.Background())
	abbruch()
	frist, ende := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer ende()
	ctxe := map[string]context.Context{"nil": kein, "beendet": beendet, "abgelaufen": frist}
	gut := support.Str("session_id", "s1")
	arten := []struct {
		name       string
		attrs      []support.Attr
		felder     int
		schluessel uint64
		wert       uint64
	}{
		{"nil-Slice", nil, 0, 0, 0},
		{"leerer Schlüssel", []support.Attr{gut, support.Str("", "x")}, 1, 1, 0},
		{"NaN", []support.Attr{gut, {Key: "duration_ms", Art: support.ArtFloat, Gleit: math.NaN()}}, 1, 0, 1},
		{"Art 3", []support.Attr{gut, {Key: "mode", Art: 3, Text: "record"}}, 1, 0, 1},
		{"Art -1", []support.Attr{gut, {Key: "mode", Art: -1, Text: "record"}}, 1, 0, 1},
		{"ungültiges UTF-8", []support.Attr{gut, support.Str("mode", string([]byte{0x72, 0xFF}))}, 1, 0, 1},
	}
	faelle := 0
	for cname, ctx := range ctxe {
		for _, art := range arten {
			faelle++
			senke := &FakeSenke{}
			a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
			wo := "ctx " + cname + ", " + art.name
			if !binnen(t, 2*time.Second, wo, func() { a.Log(ctx, slog.LevelInfo, "m", art.attrs...) }) {
				t.Fatalf("%s: Log kehrt nicht zurück", wo)
			}
			z := senke.zeilen()
			if len(z) != 1 {
				t.Errorf("%s: %d Zeilen, erwartet 1", wo, len(z))
				continue
			}
			if g := lies(t, z[0]); len(g.schluessel) != 3+art.felder {
				t.Errorf("%s: Schlüssel %v", wo, g.schluessel)
			}
			if d := a.Snapshot().Dropped; d[support.VerwurfSchluessel] != art.schluessel || d[support.VerwurfWert] != art.wert {
				t.Errorf("%s: schluessel %d, wert %d; erwartet %d und %d", wo, d[support.VerwurfSchluessel], d[support.VerwurfWert], art.schluessel, art.wert)
			}
		}
	}
	if faelle != 18 {
		t.Fatalf("%d Fälle, erwartet 18", faelle)
	}
}

// formenWrite sind die sieben Formen der Rückgabe von Write (Menge M9) mit dem Wert von senke je Zeile.
func formenWrite() []struct {
	name  string
	f     func(p []byte) (int, error)
	senke uint64
} {
	fehler := errors.New("Senke defekt")
	return []struct {
		name  string
		f     func(p []byte) (int, error)
		senke uint64
	}{
		{"0 Bytes, Fehler", func([]byte) (int, error) { return 0, fehler }, 1},
		{"weniger Bytes, Fehler", func(p []byte) (int, error) { return len(p) / 2, fehler }, 1},
		{"alle Bytes, Fehler", func(p []byte) (int, error) { return len(p), fehler }, 1},
		{"0 Bytes, kein Fehler", func([]byte) (int, error) { return 0, nil }, 1},
		{"weniger Bytes, kein Fehler", func(p []byte) (int, error) { return len(p) - 1, nil }, 1},
		{"mehr Bytes, kein Fehler", func(p []byte) (int, error) { return len(p) + 5, nil }, 0},
		{"alle Bytes, kein Fehler", func(p []byte) (int, error) { return len(p), nil }, 0},
	}
}

// zeilenVarianten sind die zwei Zeilen der Menge M9: nur Pflichtfelder, und alle 13 Felder.
func zeilenVarianten() [][]support.Attr {
	var alle []support.Attr
	for _, k := range orakelFolge() {
		alle = append(alle, gueltigFuer(k))
	}
	return [][]support.Attr{nil, alle}
}

// L10: Menge M9 (7 Formen × 2 Zeilen): senke gleich dem Modell, ein Write je Zeile (kein Wiederholen,
// kein Nachschreiben), die nächste Zeile ganz in der Senke; ein Verwurf des Attributs zählt davon
// unabhängig.
func TestSenkeAusfall(t *testing.T) {
	formen := formenWrite()
	if len(formen) != 7 {
		t.Fatalf("%d Formen, erwartet 7", len(formen))
	}
	for _, f := range formen {
		for i, attrs := range zeilenVarianten() {
			senke := &FakeSenke{antwort: f.f}
			a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
			for range 2 {
				a.Log(context.Background(), slog.LevelInfo, "m", attrs...)
			}
			z := senke.zeilen()
			if len(z) != 2 || senke.anzahl() != 2 || !bytes.Equal(z[0], z[1]) || len(lies(t, z[1]).schluessel) != 3+len(attrs) {
				t.Errorf("%s, Zeile %d: %d Write; die zweite Zeile nicht ganz in der Senke", f.name, i, senke.anzahl())
			}
			if got := a.Snapshot().Dropped[support.VerwurfSenke]; got != 2*f.senke {
				t.Errorf("%s, Zeile %d: senke %d, erwartet %d", f.name, i, got, 2*f.senke)
			}
		}
	}
	senke := &FakeSenke{antwort: formen[0].f}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	a.Log(context.Background(), slog.LevelInfo, "m", support.Str("session_id", "s1"), support.Str("mode", "a b"))
	if d := a.Snapshot().Dropped; d[support.VerwurfWert] != 1 || d[support.VerwurfSenke] != 1 {
		t.Errorf("wert %d, senke %d; erwartet je 1", d[support.VerwurfWert], d[support.VerwurfSenke])
	}
}

// L12: ein Write je Zeile mit der ganzen Zeile samt Zeilenende: drei verschiedene Zeilen
// (Pflichtfelder, alle 13 Felder, msg mit Maskierung). Die Verweise der Senke zeigen nach der
// Rückkehr dieselben Bytes wie die Kopien: der Adapter schreibt in kein übergebenes Argument zurück.
func TestEinWrite(t *testing.T) {
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	varianten := zeilenVarianten()
	a.Log(context.Background(), slog.LevelInfo, "eins", varianten[0]...)
	a.Log(context.Background(), slog.LevelWarn, "zwei", varianten[1]...)
	a.Log(context.Background(), slog.LevelError, "drei \" \n <", varianten[0]...)
	z := senke.zeilen()
	if senke.anzahl() != 3 || len(z) != 3 {
		t.Fatalf("%d Write, erwartet 3", senke.anzahl())
	}
	for i, zeile := range z {
		g := lies(t, zeile)
		if !bytes.Equal(zeile, senke.verweise[i]) {
			t.Errorf("Zeile %d: das Argument von Write wurde nach der Rückkehr geändert", i)
		}
		for j := range i {
			if bytes.Equal(zeile, z[j]) {
				t.Errorf("Zeilen %d und %d gleich", j, i)
			}
		}
		if want := []string{"eins", "zwei", "drei \" \n <"}[i]; g.text["msg"] != want {
			t.Errorf("Zeile %d: msg %q, erwartet %q", i, g.text["msg"], want)
		}
	}
}

// L13: Ersatz des Teils „Log schreibt keine Zeile“ von TestKernSchreibtNichts: die drei Aufrufe des
// alten Tests gegen MinLevel INFO ergeben zwei Zeilen in dieser Folge, der Eintrag DEBUG keine.
func TestLogSchreibtZeile(t *testing.T) {
	a, puffer := neu(t)
	a.Log(context.Background(), slog.LevelError, "m", support.Str("mode", "record"))
	a.Log(context.TODO(), slog.LevelDebug, "m")
	var kein context.Context
	a.Log(kein, slog.LevelInfo, "m")
	zeilen := strings.SplitAfter(puffer.String(), "\n")
	if len(zeilen) != 3 || zeilen[2] != "" {
		t.Fatalf("Zeilen %q, erwartet zwei", zeilen)
	}
	e, i := lies(t, []byte(zeilen[0])), lies(t, []byte(zeilen[1]))
	if e.text["level"] != "ERROR" || e.text["mode"] != "record" || i.text["level"] != "INFO" || len(i.schluessel) != 3 {
		t.Errorf("Zeilen %q", zeilen)
	}
}

// L14: Menge M10 (Stelle, Weg, Ausgabe): 13 Felder + 1 fremder Schlüssel + 1 Fehlertext der Senke
// = 15 Stellen × 4 Wege (Header, Body, Query-Literal, Konfigurationswert) × 2 Ausgaben (Bytes der
// Senke, %+v des Stands): weder das Canary noch sein Präfix von 4 Bytes. Die Kontrolle zeigt, dass
// derselbe Wert in einem gültigen Feld die Prüfung rot färbt.
func TestCanary(t *testing.T) {
	const canary = "kZq7Xw3VbN"
	wege := []struct{ name, wert string }{
		{"Header", "Bearer " + canary},
		{"Body", `{"password":"` + canary + `"}`},
		{"Query", "$filter=Name eq '" + canary + "'"},
		{"Konfiguration", "https://user:" + canary + "@host"},
	}
	stellen := append(logSchluessel(t), "fremd", "Senke")
	enthaelt := func(s []byte) bool { return bytes.Contains(s, []byte(canary[:4])) }
	pruefungen := 0
	for _, stelle := range stellen {
		for _, weg := range wege {
			senke := &FakeSenke{}
			a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
			switch stelle {
			case "Senke":
				senke.antwort = func([]byte) (int, error) { return 0, errors.New("Schreiben: " + weg.wert) }
				a.Log(context.Background(), slog.LevelInfo, "m", support.Str("session_id", "s1"))
			default:
				a.Log(context.Background(), slog.LevelInfo, "m", support.Str(stelle, weg.wert))
			}
			ausgaben := map[string][]byte{"Senke": senke.bytesAlle(), "Stand": fmt.Appendf(nil, "%+v", a.Snapshot())}
			for name, o := range ausgaben {
				pruefungen++
				if enthaelt(o) {
					t.Errorf("Stelle %s, Weg %s, Ausgabe %s: das Canary oder sein Präfix steht in der Ausgabe", stelle, weg.name, name)
				}
			}
		}
	}
	if pruefungen != 15*4*2 {
		t.Fatalf("%d Prüfungen, erwartet %d", pruefungen, 15*4*2)
	}
	senke := &FakeSenke{}
	neuMit(t, senke, festeUhr(), slog.LevelInfo).Log(context.Background(), slog.LevelInfo, "m", support.Str("session_id", canary))
	if !enthaelt(senke.bytesAlle()) {
		t.Error("Kontrolle: ein Canary in einem gültigen Feld steht in der Zeile; die Prüfung sieht es nicht")
	}
	if enthaelt(festeUhrZeile(t)) {
		t.Error("Kontrolle: das Präfix steht in der Zeile eines Aufrufs mit gültigen Werten")
	}
}

// rohWert ist der Text des Werts von msg in den Rohbytes einer Zeile ohne weitere Felder.
func rohWert(t *testing.T, z []byte) []byte {
	t.Helper()
	vor, nach := `{"timestamp":"`+festerStempel+`","level":"INFO","msg":"`, "\"}\n"
	if !bytes.HasPrefix(z, []byte(vor)) || !bytes.HasSuffix(z, []byte(nach)) || len(z) < len(vor)+len(nach) {
		t.Fatalf("Zeile %q hat nicht die Form einer Zeile ohne Felder", z)
	}
	return z[len(vor) : len(z)-len(nach)]
}

// F-1 (SPEC-231 Nr. 1): Menge Zeichen der Eingabe × maskiert/unverändert, Orakel auf den Rohbytes
// der Zeile und nicht über den JSON-Leser. Elemente: die 256 Bytewerte einzeln (0x80 bis 0xFF
// allein sind ungültig und ergeben ein U+FFFD), jedes gültige Zeichen von U+0000 bis U+FFFF ohne
// die 2048 Ersatzzeichen (63488) und drei Zeichen oberhalb (U+10000, U+10FFFF, U+1F600), zusammen
// 63747. Maskiert werden allein ", Backslash, die Zeichen unter U+0020, U+2028 und U+2029: dort
// beginnt der Wert mit einem Backslash, enthält das Zeichen nicht roh und liest sich als dasselbe
// Zeichen; jedes andere Zeichen steht mit genau seinen Bytes da. Prüfgrenze: Zeichen oberhalb von
// U+FFFF sind mit drei Vertretern geprüft (der Code behandelt sie nicht anders als U+0800 bis U+FFFF).
func TestMaskierungRohbytes(t *testing.T) {
	var menge [][2]string // roh, erwarteter Text von msg
	for v := range 256 {
		if v < 0x80 {
			menge = append(menge, [2]string{string([]byte{byte(v)}), string([]byte{byte(v)})})
		} else {
			menge = append(menge, [2]string{string([]byte{byte(v)}), string(rune(0xFFFD))})
		}
	}
	einzeln := len(menge)
	for r := rune(0); r <= 0xFFFF; r++ {
		if r < 0xD800 || r > 0xDFFF {
			menge = append(menge, [2]string{string(r), string(r)})
		}
	}
	bmp := len(menge) - einzeln
	for _, r := range []rune{0x10000, 0x10FFFF, 0x1F600} {
		menge = append(menge, [2]string{string(r), string(r)})
	}
	if einzeln != 256 || bmp != 0x10000-2048 || len(menge) != 256+0x10000-2048+3 {
		t.Fatalf("%d einzeln, %d BMP, %d Elemente", einzeln, bmp, len(menge))
	}
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	for _, e := range menge {
		a.Log(context.Background(), slog.LevelInfo, e[0])
	}
	zeilen := senke.zeilen()
	if len(zeilen) != len(menge) {
		t.Fatalf("%d Zeilen, erwartet %d", len(zeilen), len(menge))
	}
	maskiert := 0
	for i, e := range menge {
		roh, text := rohWert(t, zeilen[i]), e[1]
		c, _ := utf8.DecodeRuneInString(text)
		if c == '"' || c == '\\' || c < ' ' || c == 0x2028 || c == 0x2029 {
			maskiert++
			if len(roh) < 2 || roh[0] != '\\' || (c != '"' && c != '\\' && bytes.Contains(roh, []byte(text))) {
				t.Fatalf("Element %d (%x): Zeichen nicht maskiert: roh %q", i, e[0], roh)
			}
			if got := lies(t, zeilen[i]).text["msg"]; got != text {
				t.Fatalf("Element %d (%x): gelesen %q, erwartet %q", i, e[0], got, text)
			}
			continue
		}
		if string(roh) != text {
			t.Fatalf("Element %d (%x): roh %x, erwartet unverändert %x", i, e[0], roh, text)
		}
	}
	// 34 aus den Bytewerten (32 Steuerzeichen, ", Backslash), 36 aus dem Bereich (dieselben und U+2028, U+2029)
	if maskiert != 34+36 {
		t.Errorf("%d maskierte Elemente, erwartet %d", maskiert, 34+36)
	}
}

// Log ändert den Stand nicht über die Verwürfe hinaus: nach einem Aufruf mit allen gültigen Feldern
// und einem ungültigen Attribut sind Fehlerzähler, Verteilungen und Spans gleich, und von den
// Verwürfen wächst allein wert.
func TestLogAendertNurVerwuerfe(t *testing.T) {
	a := neuMit(t, &FakeSenke{}, festeUhr(), slog.LevelInfo)
	vorher := a.Snapshot()
	attrs := []support.Attr{support.Str("mode", "a b")}
	for _, k := range orakelFolge() {
		attrs = append(attrs, gueltigFuer(k))
	}
	a.Log(context.Background(), slog.LevelError, "m", attrs...)
	nachher := a.Snapshot()
	if !reflect.DeepEqual(vorher.Errors, nachher.Errors) || !reflect.DeepEqual(vorher.Durations, nachher.Durations) || len(nachher.Spans) != 0 {
		t.Errorf("Log änderte Fehlerzähler, Verteilungen oder Spans: %+v", nachher)
	}
	for g, n := range nachher.Dropped {
		if want := vorher.Dropped[g]; (g == support.VerwurfWert && n != want+1) || (g != support.VerwurfWert && n != want) {
			t.Errorf("Grund %s: %d, vorher %d", g, n, want)
		}
	}
}

func festeUhrZeile(t *testing.T) []byte {
	t.Helper()
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	var attrs []support.Attr
	for _, k := range orakelFolge() {
		attrs = append(attrs, gueltigFuer(k))
	}
	a.Log(context.Background(), slog.LevelInfo, "m", attrs...)
	return senke.bytesAlle()
}

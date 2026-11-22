package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// Die Gründe der Verstöße, unabhängig vom Code aus SPEC-110 Nr. 9 und 17 abgeschrieben.
const (
	grundVersionAufzaehlung = "Wert nicht in der Aufzählung: 4.0, 2.0"
	grundVersionHybrid      = "Modus hybrid für Version 2.0 nicht unterstützt"
	grundDoppeltGesetzt     = "zugleich mit SERVICES_FILE gesetzt"
	grundNurZeichenkette    = "muss eine Zeichenkette sein"
)

// versionW ist die Menge W: die zwei Versionen (SPEC-110 Nr. 17 a).
func versionW() []string { return []string{"4.0", "2.0"} }

// versionFall ist eine Eingabe der Menge R und der Wert, den die Zusage dazu nennt.
type versionFall struct{ eingabe, soll string }

// versionR ist die Menge R: Randleerraum um jeden Wert von W, 4 Zeichen (Leerzeichen, Tabulator, LF, CR) × 3 Seiten × 2 Werte = 24.
func versionR() []versionFall {
	var aus []versionFall
	for _, w := range versionW() {
		for _, z := range []string{" ", "\t", "\n", "\r"} {
			aus = append(aus, versionFall{z + w, w}, versionFall{w + z, w}, versionFall{z + w + z, w})
		}
	}
	return aus
}

// nbsp ist U+00A0 aus Bytewerten (AGENTS.md §3.18); der Rand zählt nicht als Leerraum (SPEC-110 Nr. 1).
func nbsp() string { return string([]byte{0xC2, 0xA0}) }

// versionF ist die Menge F der fremden Werte, 27: je Wert von W neun Abwandlungen (Präfix v, V und 0; Suffix 0, .0 und x; Komma
// statt Punkt; Punkt entfernt; nur der Major), NBSP vor und nach jedem Wert von W, dazu fünf Nachbarn.
func versionF() []string {
	var f []string
	for _, w := range versionW() {
		f = append(f, "v"+w, "V"+w, "0"+w, w+"0", w+".0", w+"x", strings.Replace(w, ".", ",", 1), strings.Replace(w, ".", "", 1), w[:1])
		f = append(f, nbsp()+w, w+nbsp())
	}
	return append(f, "1.0", "2.1", "3.0", "4.01", "5.0")
}

// versionT sind die rohen JSON-Werte falschen Typs eines Feldes der Datei: Zahl, Wahrheitswert, Array, Objekt.
func versionT() []string { return []string{`1`, `true`, `[]`, `{}`} }

// belegungen sind die Folgen aus W der Länge n (2^n).
func belegungen(n int) [][]string {
	var aus [][]string
	for maske := 0; maske < 1<<n; maske++ {
		b := make([]string, n)
		for i := range b {
			b[i] = versionW()[(maske>>(n-1-i))&1]
		}
		aus = append(aus, b)
	}
	return aus
}

// alleBelegungen sind die Belegungen für N = 1, 2, 3 (2 + 4 + 8 = 14).
func alleBelegungen() [][]string {
	var aus [][]string
	for n := 1; n <= 3; n++ {
		aus = append(aus, belegungen(n)...)
	}
	return aus
}

func mitV2(b []string) bool { return slices.Contains(b, "2.0") }

// versionDatei schreibt eine SERVICES_FILE mit einem Eintrag je Element von roh: id und service_root aus dem Index, dazu das Feld
// odata_version mit dem rohen JSON-Wert, wenn er nicht leer ist.
func versionDatei(t *testing.T, roh []string) string {
	t.Helper()
	var eintraege []string
	for i, v := range roh {
		e := fmt.Sprintf(`{"id":"s%d","service_root":"/s%d"`, i, i)
		if v != "" {
			e += `,"odata_version":` + v
		}
		eintraege = append(eintraege, e+"}")
	}
	return schreibe(t, "["+strings.Join(eintraege, ",")+"]")
}

// versionJSON ist der Wert als JSON-Zeichenkette.
func versionJSON(t *testing.T, s string) string {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// zaehle verlangt die Zahl der Elemente einer Menge, damit eine Schleife nicht still leer läuft.
func zaehle(t *testing.T, name string, ist, soll int) {
	t.Helper()
	if ist != soll {
		t.Fatalf("Menge %s hat %d Elemente, erwartet %d", name, ist, soll)
	}
}

// versionEinzel ist eine gültige Umgebung im Modus record mit Einzelschlüsseln.
func versionEinzel(t *testing.T, extra ...string) []string { return append(basis(t), extra...) }

// keinDatenverzeichnis verlangt, dass DATA_DIR der Umgebung nicht angelegt ist (SPEC-110 Nr. 11).
func keinDatenverzeichnis(t *testing.T, env []string) {
	t.Helper()
	for i := len(env) - 1; i >= 0; i-- {
		if dir, ok := strings.CutPrefix(env[i], "DATA_DIR="); ok {
			if _, err := os.Stat(dir); err == nil {
				t.Errorf("DATA_DIR %s ist angelegt", dir)
			}
			return
		}
	}
	t.Fatal("kein DATA_DIR in der Umgebung")
}

// TestVersionEinzelGueltig fängt A01: der Einzelschlüssel mit 4.0 oder 2.0, auch mit Randleerraum, erreicht Dienste[0].Version ohne
// den Rand (SPEC-110 Nr. 17 a). Menge W und R, 26 Eingaben, Erwartung aus der Menge.
func TestVersionEinzelGueltig(t *testing.T) {
	var faelle []versionFall
	for _, w := range versionW() {
		faelle = append(faelle, versionFall{w, w})
	}
	zaehle(t, "W", len(faelle), 2)
	zaehle(t, "R", len(versionR()), 24)
	faelle = append(faelle, versionR()...)
	for _, f := range faelle {
		k, zeilen := lade(t, versionEinzel(t, "SERVICE_ODATA_VERSION="+f.eingabe)...)
		if len(zeilen) != 0 || len(k.Dienste) != 1 || k.Dienste[0].Version != f.soll {
			t.Errorf("SERVICE_ODATA_VERSION=%q: Zeilen %q, Dienste %+v, erwartet Version %q", f.eingabe, zeilen, k.Dienste, f.soll)
		}
	}
}

// TestVersionEinzelFehlt fängt A02: fehlend, leer und nur Leerraum gelten 4.0, die Konfiguration trägt nie einen leeren Wert.
func TestVersionEinzelFehlt(t *testing.T) {
	eingaben := [][]string{nil, {"SERVICE_ODATA_VERSION="}}
	for _, z := range []string{" ", "\t", "\n", "\r"} {
		eingaben = append(eingaben, []string{"SERVICE_ODATA_VERSION=" + z})
	}
	zaehle(t, "fehlend, leer, Leerraum", len(eingaben), 6)
	for _, e := range eingaben {
		k, zeilen := lade(t, versionEinzel(t, e...)...)
		if len(zeilen) != 0 || len(k.Dienste) != 1 || k.Dienste[0].Version != "4.0" {
			t.Errorf("Eingabe %q: Zeilen %q, Dienste %+v, erwartet Version 4.0", e, zeilen, k.Dienste)
		}
	}
}

// TestVersionEinzelFremd fängt A03: jeder fremde Wert ist genau eine Zeile ohne den Wert, und DATA_DIR bleibt unangelegt
// (SPEC-110 Nr. 11, 17 a und c). Menge F, 27, disjunkt zu W und R.
func TestVersionEinzelFremd(t *testing.T) {
	f := versionF()
	zaehle(t, "F", len(f), 27)
	bekannt := map[string]bool{}
	for _, w := range versionW() {
		bekannt[w] = true
	}
	for _, r := range versionR() {
		bekannt[r.eingabe] = true
	}
	gesehen := map[string]bool{}
	for _, w := range f {
		if bekannt[w] || gesehen[w] {
			t.Fatalf("F enthält %q doppelt oder aus W und R", w)
		}
		gesehen[w] = true
	}
	for _, w := range f {
		env := versionEinzel(t, "SERVICE_ODATA_VERSION="+w)
		_, zeilen := lade(t, env...)
		if want := []string{zeile("SERVICE_ODATA_VERSION", grundVersionAufzaehlung)}; !slices.Equal(zeilen, want) {
			t.Errorf("SERVICE_ODATA_VERSION=%q: Zeilen %q, erwartet %q", w, zeilen, want)
		}
		keinDatenverzeichnis(t, env)
	}
}

// TestVersionDatei fängt A04: odata_version je Eintrag wirkt für diesen Eintrag und für keinen anderen (SPEC-110 Nr. 9, 17 a). Die Datei
// hat 1, 2 oder 3 Einträge (6 Positionen) und je Position fünf Wertklassen (30); die Nachbarn tragen den anderen Wert.
func TestVersionDatei(t *testing.T) {
	klassen := []struct{ name, roh, soll string }{
		{"2.0", `"2.0"`, "2.0"},
		{"2.0 mit Randleerraum", versionJSON(t, " \t2.0\r\n "), "2.0"},
		{"fehlend", ``, "4.0"},
		{"null", `null`, "4.0"},
		{"leer", `""`, "4.0"},
	}
	zaehle(t, "Wertklassen", len(klassen), 5)
	positionen, zellen := 0, 0
	for n := 1; n <= 3; n++ {
		for pos := 0; pos < n; pos++ {
			positionen++
			for _, kl := range klassen {
				zellen++
				nachbarRoh, nachbarSoll := `"2.0"`, "2.0"
				if kl.soll == "2.0" {
					nachbarRoh, nachbarSoll = `"4.0"`, "4.0"
				}
				roh := make([]string, n)
				soll := make([]string, n)
				for i := range roh {
					roh[i], soll[i] = nachbarRoh, nachbarSoll
				}
				roh[pos], soll[pos] = kl.roh, kl.soll
				env := []string{"MODE=replay", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), "SERVICES_FILE=" + versionDatei(t, roh)}
				k, zeilen := lade(t, env...)
				got := make([]string, len(k.Dienste))
				for i, d := range k.Dienste {
					got[i] = d.Version
				}
				if len(zeilen) != 0 || !slices.Equal(got, soll) {
					t.Errorf("Klasse %q an Position %d von %d: Zeilen %q, Versionen %q, erwartet %q", kl.name, pos, n, zeilen, got, soll)
				}
			}
		}
	}
	zaehle(t, "Positionen", positionen, 6)
	zaehle(t, "Zellen", zellen, 30)
}

// TestVersionDateiFremd fängt A05: ein fremder Wert ist die Zeile mit dem Index des Eintrags, ein Feld falschen Typs allein „muss eine
// Zeichenkette sein“, keine zweite Zeile (SPEC-110 Nr. 9, 17 a). F und T × 6 Positionen, 186 Zellen.
func TestVersionDateiFremd(t *testing.T) {
	f, typen := versionF(), versionT()
	zaehle(t, "F", len(f), 27)
	zaehle(t, "T", len(typen), 4)
	var faelle []struct{ roh, grund string }
	for _, w := range f {
		faelle = append(faelle, struct{ roh, grund string }{versionJSON(t, w), grundVersionAufzaehlung})
	}
	for _, roh := range typen {
		faelle = append(faelle, struct{ roh, grund string }{roh, grundNurZeichenkette})
	}
	positionen := 0
	for n := 1; n <= 3; n++ {
		for pos := 0; pos < n; pos++ {
			positionen++
			for _, fall := range faelle {
				roh := make([]string, n)
				roh[pos] = fall.roh
				env := []string{"MODE=replay", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), "SERVICES_FILE=" + versionDatei(t, roh)}
				_, zeilen := lade(t, env...)
				want := []string{zeile(fmt.Sprintf("SERVICES_FILE[%d].odata_version", pos), fall.grund)}
				if !slices.Equal(zeilen, want) {
					t.Errorf("Eintrag %d von %d mit %s: Zeilen %q, erwartet %q", pos, n, fall.roh, zeilen, want)
				}
			}
		}
	}
	zaehle(t, "Positionen", positionen, 6)
	zaehle(t, "Zellen", positionen*len(faelle), 186)
}

// TestDoppelkonfiguration fängt A06: jeder der vier Einzelschlüssel zugleich mit SERVICES_FILE ist allein die Zeile der
// Doppelkonfiguration, auch SERVICE_ODATA_VERSION mit fremdem Wert (SPEC-110 Nr. 9, 17 a). 4 Schlüssel × {gültig, fremd}.
func TestDoppelkonfiguration(t *testing.T) {
	schluessel := []string{"SERVICE_ID", "SERVICE_ROOT_PATH", "UPSTREAM_URL", "SERVICE_ODATA_VERSION"}
	werte := map[string][2]string{
		"SERVICE_ID":            {"x", "a b"},
		"SERVICE_ROOT_PATH":     {"/x", "x"},
		"UPSTREAM_URL":          {"https://h.test/odata", "ftp://h"},
		"SERVICE_ODATA_VERSION": {"2.0", "v2.0"},
	}
	zaehle(t, "Einzelschlüssel", len(schluessel), 4)
	zellen := 0
	for _, key := range schluessel {
		zaehle(t, "Werte von "+key, len(werte[key]), 2)
		for _, w := range werte[key] {
			zellen++
			env := []string{"MODE=replay", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), "SERVICES_FILE=" + versionDatei(t, []string{""}), key + "=" + w}
			_, zeilen := lade(t, env...)
			if want := []string{zeile(key, grundDoppeltGesetzt)}; !slices.Equal(zeilen, want) {
				t.Errorf("%s=%q: Zeilen %q, erwartet %q", key, w, zeilen, want)
			}
		}
	}
	zaehle(t, "Zellen", zellen, 8)
}

// TestHybridMitV2 fängt A07: hybrid mit mindestens einem Service 2.0 ist genau eine Zeile am Schlüssel MODE, gleich wie viele und an
// welcher Stelle, und DATA_DIR bleibt unangelegt; Start endet mit Exit 1 und derselben Zeile (SPEC-110 Nr. 11, 17 b und c).
// Menge 12: der Einzelschlüssel (1) und die Datei mit N = 1, 2, 3 Einträgen, jede Belegung mit mindestens einem 2.0 (1 + 3 + 7).
func TestHybridMitV2(t *testing.T) {
	type konfig struct {
		name string
		env  []string
	}
	var faelle []konfig
	dir := func() string { return "DATA_DIR=" + filepath.Join(t.TempDir(), "d") }
	faelle = append(faelle, konfig{"Einzelschlüssel", []string{"MODE=hybrid", "SERVICE_ODATA_VERSION=2.0", dir()}})
	je := map[int]int{}
	for n := 1; n <= 3; n++ {
		for _, b := range belegungen(n) {
			if !mitV2(b) {
				continue
			}
			je[n]++
			roh := make([]string, n)
			for i, v := range b {
				roh[i] = versionJSON(t, v)
			}
			faelle = append(faelle, konfig{fmt.Sprint("Datei ", b), []string{"MODE=hybrid", dir(), "SERVICES_FILE=" + versionDatei(t, roh)}})
		}
	}
	if je[1] != 1 || je[2] != 3 || je[3] != 7 || len(faelle) != 12 {
		t.Fatalf("Menge: %v, %d Konfigurationen, erwartet 1, 3, 7 und 12", je, len(faelle))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, f := range faelle {
		want := zeile("MODE", grundVersionHybrid)
		_, zeilen := lade(t, f.env...)
		if !slices.Equal(zeilen, []string{want}) {
			t.Errorf("%s: Zeilen %q, erwartet allein %q", f.name, zeilen, want)
		}
		keinDatenverzeichnis(t, f.env)
		var stderr syncPuffer
		if exit := bootstrap.StartMit(ctx, f.env, &stderr, bootstrap.StandardBauteile()); exit != bootstrap.ExitKonfiguration || stderr.String() != want+"\n" {
			t.Errorf("%s: Start endet mit %d und %q, erwartet %d und %q", f.name, exit, stderr.String(), bootstrap.ExitKonfiguration, want+"\n")
		}
		keinDatenverzeichnis(t, f.env)
	}
}

// TestVersionModus fängt A08: allein hybrid mit 2.0 verletzt; record, replay und mock nehmen jede Version; bei ungültigem MODE
// meldet allein MODE seine Form (SPEC-110 Nr. 8, 17 b). Tabelle 5 Modi × 2 Versionen.
func TestVersionModus(t *testing.T) {
	modi := []string{"record", "replay", "mock", "hybrid", "Mock"}
	zaehle(t, "Modi", len(modi), 5)
	zaehle(t, "Versionen", len(versionW()), 2)
	for _, modus := range modi {
		for _, v := range versionW() {
			env := []string{"MODE=" + modus, "SERVICE_ODATA_VERSION=" + v, "UPSTREAM_URL=https://h.test/odata", "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}
			var want []string
			switch {
			case modus == "Mock":
				want = []string{zeile("MODE", "Wert nicht in der Aufzählung: record, replay, mock, hybrid")}
			case modus == "hybrid" && v == "2.0":
				want = []string{zeile("MODE", grundVersionHybrid)}
			}
			if _, zeilen := lade(t, env...); !slices.Equal(zeilen, want) {
				t.Errorf("MODE=%s, Version %s: Zeilen %q, erwartet %q", modus, v, zeilen, want)
			}
		}
	}
}

// TestVersionFremdHybrid fängt A09: ein fremder Wert zählt weder als 2.0 noch als 4.0; mit hybrid meldet er allein seine Form
// (SPEC-110 Nr. 17 b). Menge F, 27.
func TestVersionFremdHybrid(t *testing.T) {
	f := versionF()
	zaehle(t, "F", len(f), 27)
	for _, w := range f {
		env := []string{"MODE=hybrid", "SERVICE_ODATA_VERSION=" + w, "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}
		_, zeilen := lade(t, env...)
		if want := []string{zeile("SERVICE_ODATA_VERSION", grundVersionAufzaehlung)}; !slices.Equal(zeilen, want) {
			t.Errorf("hybrid mit SERVICE_ODATA_VERSION=%q: Zeilen %q, erwartet %q", w, zeilen, want)
		}
	}
}

// TestVersionMeldung fängt A10: alle Verstöße stehen in einer Meldung, nach Schlüssel in Bytefolge, MODE vor SERVICE_ODATA_VERSION und
// vor SERVICES_FILE[…] (SPEC-110 Nr. 12, 17 c). Vier Begleitverstöße zu hybrid mit 2.0, deren Schlüssel vor MODE stehen oder (die
// Datei) nach ihm; zwei weitere Schlüssel hinter MODE halten die Folge auch dort.
func TestVersionMeldung(t *testing.T) {
	hybridV2 := func(extra ...string) []string {
		return append([]string{"MODE=hybrid", "SERVICE_ODATA_VERSION=2.0", "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}, extra...)
	}
	datei := func() []string {
		return []string{"MODE=hybrid", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), "SERVICES_FILE=" + versionDatei(t, []string{`"2.0"`, versionJSON(t, "v2.0")})}
	}
	modus := zeile("MODE", grundVersionHybrid)
	vorher := []struct {
		name string
		env  []string
		want []string
	}{
		{"LISTEN", hybridV2("LISTEN=kaputt"), []string{zeile("LISTEN", "Adresse muss host:port sein, Port 1 bis 65535"), modus}},
		{"HYBRID_FALLBACK", hybridV2("HYBRID_FALLBACK=vielleicht"), []string{zeile("HYBRID_FALLBACK", "Wert nicht in der Aufzählung: disabled, enabled"), modus}},
		{"SERVICE_ID", hybridV2("SERVICE_ID=a b"), []string{modus, zeile("SERVICE_ID", "1 bis 64 Zeichen aus Buchstaben, Ziffern, ., _ und -")}},
		{"fremde Version eines zweiten Eintrags", datei(), []string{modus, zeile("SERVICES_FILE[1].odata_version", grundVersionAufzaehlung)}},
	}
	zaehle(t, "Begleitverstöße", len(vorher), 4)
	nachher := []struct {
		name string
		env  []string
		want []string
	}{
		{"SERVICE_ROOT_PATH", hybridV2("SERVICE_ROOT_PATH=x"), []string{modus, zeile("SERVICE_ROOT_PATH", "muss / sein oder Segmente /a/b ohne abschließenden Schrägstrich")}},
		{"UPSTREAM_URL", hybridV2("UPSTREAM_URL=ftp://h"), []string{modus, zeile("UPSTREAM_URL", "keine absolute http- oder https-URL ohne Zugangsdaten, Query und Fragment")}},
	}
	zaehle(t, "Begleitverstöße hinter MODE", len(nachher), 2)
	for _, f := range append(vorher, nachher...) {
		if _, zeilen := lade(t, f.env...); !slices.Equal(zeilen, f.want) {
			t.Errorf("%s: Zeilen %q, erwartet %q", f.name, zeilen, f.want)
		}
		keinDatenverzeichnis(t, f.env)
	}
}

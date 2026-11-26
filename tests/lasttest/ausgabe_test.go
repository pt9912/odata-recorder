package main

import (
	"bytes"
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// zeileVon liefert die Zeile von z und verlangt, dass sie keinen Fehler trägt.
func zeileVon(t *testing.T, z zeilenart) string {
	t.Helper()
	s, err := z.zeile()
	if err != nil {
		t.Fatalf("%+v: %v", z, err)
	}
	return s
}

// TestAusgabe: Golden-Text je Urteil (3), je Hinweis (2, dazu ein zweiter Exit) und je Zeilenart (6: Start, Aufbau, Messung, Last, Urteil, Abnahme, mit allen Schlüsseln
// in der Folge des Vertrags; 14 Zeilen, `gueltig` und `grund` von einander unabhängig geführt: `Aufbau` mit verschiedenen `exchanges` und `entities` und einer Dauer mit 999.999 ns über der
// Millisekunde, abgeschnitten statt gerundet, `Messung` mit `gueltig=nein` und `gueltig=ja`, dazu Dauern unter einer Millisekunde); Zahlenformen aus ganzen Zahlen (0 gleich 0.000, 1 gleich 0.001, 999999 gleich 999.999, 1000000 gleich 1000.000, dazu
// eine Zahl über 2^53); die 13 Tokens und je 5 Werte außerhalb der Menge in den drei Zeilen mit grund (15); schreibe schreibt nichts bei einem Fehler.
func TestAusgabe(t *testing.T) {
	wache(t, 30*time.Second)
	golden := []struct {
		z    zeilenart
		want string
	}{
		{urteil{"replay", 2, 100, 99999, 123, 9999, urteilBestanden, grundOK},
			"lasttest: Urteil betriebsart=replay laeufe=2 ziel_ms=100 p95_ms=99.999 streuung_ms=0.123 toleranz_ms=9.999 ergebnis=bestanden grund=ok"},
		{urteil{"mock", 1, 80, 100001, 0, 0, urteilDurchgefallen, grundOK},
			"lasttest: Urteil betriebsart=mock laeufe=1 ziel_ms=80 p95_ms=100.001 streuung_ms=0.000 toleranz_ms=0.000 ergebnis=durchgefallen grund=ok"},
		{urteil{"mock", 2, 100, 45000, 20001, 4500, urteilUngueltig, grundStreuung},
			"lasttest: Urteil betriebsart=mock laeufe=2 ziel_ms=100 p95_ms=45.000 streuung_ms=20.001 toleranz_ms=4.500 ergebnis=ungueltig grund=streuung"},
		{hinweis{Betriebsart: "replay", Exit: 1}, "lasttest: Hinweis betriebsart=replay binary_exit=1"},
		{hinweis{Betriebsart: "mock", Kill: true}, "lasttest: Hinweis betriebsart=mock binary_kill=ja"},
		{hinweis{Betriebsart: "mock", Exit: 137}, "lasttest: Hinweis betriebsart=mock binary_exit=137"},
		{startZeile{"abnahme", "beide", 100, 999999999999999999, 20},
			"lasttest: Start profil=abnahme betriebsart=beide ziel_ms=100 startwert=999999999999999999 kerne=20"},
		{aufbauZeile{10000, 9999, 61, 95*time.Second + 250*time.Millisecond + 999999*time.Nanosecond},
			"lasttest: Aufbau exchanges=10000 entities=9999 daten_mib=61 dauer_s=95.250"},
		{Messung{"replay", 2, 13200, 13199, 12000, 3, 7, 98765, 123456, 11999, 20001, 805, 4, 5, 6, 3500, 987654321, 1500*time.Millisecond + 999999*time.Nanosecond, false, grundFehler},
			"lasttest: Messung betriebsart=replay lauf=2 anfragen=13200 spans=13199 messphase=12000 fehler=3 fremde_zeilen=7 p95_ms=98.765 " +
				"max_ms=123.456 kind=11999 kind_p95_ms=20.001 kind_anteil=80.5 ohne_kind=4 kind_ohne_eltern=5 gleichzeitig_max=6 " +
				"verspaetung_max_ms=3.500 bytes=987654321 treiber_cpu_s=1.500 gueltig=nein grund=fehler"},
		{Messung{"mock", 1, 260, 260, 200, 0, 0, 999, 1000, 0, 0, 0, 0, 0, 1, 1, 7, 0, true, grundOK},
			"lasttest: Messung betriebsart=mock lauf=1 anfragen=260 spans=260 messphase=200 fehler=0 fremde_zeilen=0 p95_ms=0.999 " +
				"max_ms=1.000 kind=0 kind_p95_ms=0.000 kind_anteil=0.0 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=1 " +
				"verspaetung_max_ms=0.001 bytes=7 treiber_cpu_s=0.000 gueltig=ja grund=ok"},
		{Messung{Betriebsart: "mock", Lauf: 2, Gueltig: false, Grund: grundOK},
			"lasttest: Messung betriebsart=mock lauf=2 anfragen=0 spans=0 messphase=0 fehler=0 fremde_zeilen=0 p95_ms=0.000 " +
				"max_ms=0.000 kind=0 kind_p95_ms=0.000 kind_anteil=0.0 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=0 " +
				"verspaetung_max_ms=0.000 bytes=0 treiber_cpu_s=0.000 gueltig=nein grund=ok"},
		{lastZeile{"mock", 1, lastErgebnis{24, 100, 1200, 1000, false}}, "lasttest: Last betriebsart=mock lauf=1 proben=24 min=1.00 max=12.00 schwelle=10.00"},
		{abnahmeZeile{100, true, grundOK}, "lasttest: Abnahme ziel_ms=100 abnahmewert_ms=100 abnahme=ja grund=ok"},
		{abnahmeZeile{80, false, grundZiel}, "lasttest: Abnahme ziel_ms=80 abnahmewert_ms=100 abnahme=nein grund=ziel"},
	}
	if len(golden) != 14 {
		t.Fatalf("%d Golden-Zeilen, erwartet 14", len(golden))
	}
	for _, g := range golden {
		if got := zeileVon(t, g.z); got != g.want {
			t.Errorf("Zeile\n%s\nerwartet\n%s", got, g.want)
		}
	}
	zahlen := map[int64]string{0: "0.000", 1: "0.001", 999: "0.999", 1000: "1.000", 999999: "999.999", 1000000: "1000.000", 9007199254740993: "9007199254740.993"}
	if len(zahlen) != 7 {
		t.Fatalf("%d Zahlenformen, erwartet 7", len(zahlen))
	}
	for n, want := range zahlen {
		if got := tausendstel(n); got != want {
			t.Errorf("tausendstel(%d) = %s, erwartet %s", n, got, want)
		}
	}
	for n, want := range map[int]string{0: "0.0", 5: "0.5", 333: "33.3", 666: "66.6", 1000: "100.0"} {
		if got := zehntel(n); got != want {
			t.Errorf("zehntel(%d) = %s, erwartet %s", n, got, want)
		}
	}
	for n, want := range map[int]string{0: "0.00", 5: "0.05", 100: "1.00", 1000: "10.00", 1001: "10.01"} {
		if got := hundertstel(n); got != want {
			t.Errorf("hundertstel(%d) = %s, erwartet %s", n, got, want)
		}
	}
	formen(t)
	tokenLiteral := []string{"ok", "fehler", "zaehlung", "modus", "last", "binary", "streuung", "aufbau", "frist", "profil", "ziel", "betriebsart", "ergebnis"}
	if !slices.Equal(tokens(), tokenLiteral) {
		t.Errorf("Tokens %v, erwartet %v", tokens(), tokenLiteral)
	}
	mitGrund := func(g string) []zeilenart {
		return []zeilenart{Messung{Grund: g}, urteil{Grund: g}, abnahmeZeile{Grund: g}}
	}
	for _, g := range tokenLiteral {
		for _, z := range mitGrund(g) {
			if s, err := z.zeile(); err != nil || !strings.HasSuffix(s, " grund="+g) {
				t.Errorf("Token %s: Zeile %q, Fehler %v", g, s, err)
			}
		}
	}
	fremde := 0
	for _, g := range []string{"", "OK", "frei text", "ok ", "nein"} {
		for _, z := range mitGrund(g) {
			var puffer bytes.Buffer
			if err := schreibe(&puffer, z); !errors.Is(err, errToken) || puffer.Len() != 0 {
				t.Errorf("grund %q in %T: Fehler %v, geschrieben %q; erwartet errToken und nichts", g, z, err, puffer.String())
			}
			fremde++
		}
	}
	if fremde != 15 {
		t.Errorf("%d fremde Werte, erwartet 15", fremde)
	}
	var puffer bytes.Buffer
	if err := schreibe(&puffer, hinweis{Betriebsart: "mock", Kill: true}); err != nil || puffer.String() != "lasttest: Hinweis betriebsart=mock binary_kill=ja\n" {
		t.Errorf("schreibe: %q, %v", puffer.String(), err)
	}
	if err := schreibe(&schlechterSchreiber{}, hinweis{}); err == nil {
		t.Error("schreibe verschweigt den Fehler des Schreibers")
	}
}

// formen führt die Zahlenformen als Ziffer × Größenordnung: n gleich d · 10^k und d · 10^k + 1 für d = 1 bis 9 und k = 0 bis 17 (324 Werte je
// Form, darunter Werte über 2^53, an denen float64 die letzte Stelle verliert); die Erwartung schneidet die Dezimaldarstellung, rechnet nicht.
func formen(t *testing.T) {
	t.Helper()
	schneide := func(n int64, nachkomma int) string {
		s := strconv.FormatInt(n, 10)
		for len(s) <= nachkomma {
			s = "0" + s
		}
		return s[:len(s)-nachkomma] + "." + s[len(s)-nachkomma:]
	}
	werte := 0
	for ziffer := int64(1); ziffer <= 9; ziffer++ {
		n := ziffer
		for k := 0; k <= 17; k++ {
			for _, wert := range []int64{n, n + 1} {
				if got, want := tausendstel(wert), schneide(wert, 3); got != want {
					t.Errorf("tausendstel(%d) = %s, erwartet %s", wert, got, want)
				}
				if got, want := zehntel(int(wert)), schneide(wert, 1); got != want {
					t.Errorf("zehntel(%d) = %s, erwartet %s", wert, got, want)
				}
				if got, want := hundertstel(int(wert)), schneide(wert, 2); got != want {
					t.Errorf("hundertstel(%d) = %s, erwartet %s", wert, got, want)
				}
				werte++
			}
			n *= 10
		}
	}
	if werte != 324 {
		t.Errorf("%d Werte, erwartet 324", werte)
	}
}

type schlechterSchreiber struct{}

func (*schlechterSchreiber) Write([]byte) (int, error) { return 0, errors.New("Schreiben gescheitert") }

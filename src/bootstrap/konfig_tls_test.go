package bootstrap_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

const grundAbgelehnt = "wird von dieser Version nicht unterstützt"

// tlsUmgebung ist eine gültige Umgebung im Modus modus; ein leerer Pfad setzt den Schlüssel nicht.
func tlsUmgebung(t *testing.T, modus, cert, key string, extra ...string) []string {
	t.Helper()
	env := hybridBasis(t, modus, extra...)
	if cert != "" {
		env = append(env, listenCert+"="+cert)
	}
	if key != "" {
		env = append(env, listenKey+"="+key)
	}
	return env
}

// pem2 liefert die ersten 24 Zeichen des Base64-Körpers eines PEM-Blocks: ein Stück, das in keiner Ausgabe der Konfiguration stehen darf.
func pem2(pem []byte) string { return strings.Split(string(pem), "\n")[1][:24] }

// TestKonfigListenTLSWerte (B01): die Konfiguration trägt die Pfade, nie den Inhalt der Dateien (SPEC-291 Nr. 7 a, SPEC-110 Nr. 1).
// Menge W, 6 Elemente; dazu die Klassen „Leerraum am Rand“ (2 Schlüssel × 4 Zeichen × 3 Seiten = 24) und „Zeichen außerhalb des Leerraums am
// Rand“ (2 × 6 × 3 = 36, bleiben im Pfad und machen die Datei unlesbar).
func TestKonfigListenTLSWerte(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relCert, err1 := filepath.Rel(wd, cert)
	relKey, err2 := filepath.Rel(wd, key)
	if err1 != nil || err2 != nil {
		t.Fatalf("relative Pfade: %v, %v", err1, err2)
	}
	nbsp := "\xc2\xa0"
	faelle := []struct {
		name, cWert, kWert, wantCert, wantKey string
		zeilen                                []string
	}{
		{"Pfad", cert, key, cert, key, nil},
		{"Leerraum am Rand", " \t\n\r" + cert + "\r\n\t ", "\t" + key + "\n", cert, key, nil},
		{"NBSP am Rand", nbsp + cert, key + nbsp, "", "", []string{zeile(listenCert, gNichtLesbar), zeile(listenKey, gNichtLesbar)}},
		{"leer", "", "", "", "", nil},
		{"nur Leerraum", " \t\r\n", "\n\t ", "", "", nil},
		{"relativer Pfad", relCert, relKey, relCert, relKey, nil},
	}
	if len(faelle) != 6 {
		t.Fatalf("Menge W hat %d Elemente, erwartet 6", len(faelle))
	}
	for _, f := range faelle {
		k, zeilen := lade(t, append(hybridBasis(t, "record"), listenCert+"="+f.cWert, listenKey+"="+f.kWert)...)
		if !slices.Equal(zeilen, f.zeilen) || k.ListenTLSCert != f.wantCert || k.ListenTLSKey != f.wantKey {
			t.Errorf("%s: Zeilen %q, Pfade %q und %q; erwartet %q, %q und %q", f.name, zeilen, k.ListenTLSCert, k.ListenTLSKey, f.zeilen, f.wantCert, f.wantKey)
		}
	}
	paare := func(zeichen string, vor, nach bool) (cs, ks [2]string) {
		return [2]string{mitRand([]string{cert}, zeichen, vor, nach)[0], key}, [2]string{cert, mitRand([]string{key}, zeichen, vor, nach)[0]}
	}
	rand, fremd := 0, 0
	for _, zeichen := range []string{" ", "\t", "\n", "\r"} {
		for _, s := range seitenPaare() {
			c, k := paare(zeichen, s.vor, s.nach)
			for _, p := range [][2]string{c, k} {
				rand++
				got := erwarteGueltig(t, tlsUmgebung(t, "record", p[0], p[1])...)
				if got.ListenTLSCert != cert || got.ListenTLSKey != key {
					t.Errorf("Rand %q, Seiten %+v: Pfade %q und %q; erwartet die Pfade ohne Leerraum", zeichen, s, got.ListenTLSCert, got.ListenTLSKey)
				}
			}
		}
	}
	for _, zeichen := range fremdeZeichen() {
		for _, s := range seitenPaare() {
			c, k := paare(zeichen, s.vor, s.nach)
			for i, p := range [][2]string{c, k} {
				fremd++
				name := []string{listenCert, listenKey}[i]
				if _, zeilen := lade(t, tlsUmgebung(t, "record", p[0], p[1])...); !slices.Equal(zeilen, []string{zeile(name, gNichtLesbar)}) {
					t.Errorf("Zeichen %q an %s, Seiten %+v: Zeilen %q, erwartet %q", zeichen, name, s, zeilen, gNichtLesbar)
				}
			}
		}
	}
	if rand != 24 || fremd != 36 {
		t.Errorf("Klassen haben %d und %d Elemente, erwartet 24 und 36", rand, fremd)
	}
	k := erwarteGueltig(t, tlsUmgebung(t, "record", cert, key)...)
	for _, s := range []string{fmt.Sprintf("%+v", k), fmt.Sprintf("%#v", k)} {
		if !strings.Contains(s, cert) || strings.Contains(s, pem2(w.schluessel(w.gut))) || strings.Contains(s, pem2(w.gut.ZertifikatPEM())) {
			t.Errorf("Ausgabe der Konfiguration trägt den Pfad nicht oder einen Teil einer Datei: %.200s", s)
		}
	}
}

// vierPlusEins sind die vier Modi und ein ungültiger Wert von MODE (Menge Modi, 5 Elemente).
func vierPlusEins() []string { return append(vierModi(), "Record") }

// TestKonfigListenTLSVerstoesse (B02): ein Verstoß des Ladens ist die Zeile `ODR-E2001 <Schlüssel>: <Grund>` mit dem Namen des Schlüssels,
// sortiert, Exit 1, ohne Pfad und Inhalt, und DATA_DIR entsteht nicht (SPEC-291 Nr. 7 b, c; SPEC-110 Nr. 11, 12). Menge V: 8 Elemente (die
// Ränge 1 bis 4 je Schlüssel), dazu 3 Paare, deren Meldung beide Schlüssel in der Folge CERT, KEY trägt.
func TestKonfigListenTLSVerstoesse(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	menge := []tlsFall{
		w.paar(n, zLeer, zGueltig), w.paar(n, zGueltig, zLeer), w.paar(n, zNichtLesbar, zGueltig), w.paar(n, zGueltig, zNichtLesbar),
		w.paar(n, zKeinPEM, zGueltig), w.paar(n, zGueltig, zKeinPEM), w.paar(n, zBesonders, zGueltig), w.paar(n, zGueltig, zBesonders),
	}
	paare := []tlsFall{w.paar(n, zBesonders, zBesonders), w.paar(n, zNichtLesbar, zNichtLesbar), w.paar(n, zKeinPEM, zLeer)}
	if len(menge) != 8 || len(paare) != 3 {
		t.Fatalf("Mengen V und Paare haben %d und %d Elemente, erwartet 8 und 3", len(menge), len(paare))
	}
	verboten := []string{"geheim-pfad-4711", "geheim-cn-4711", w.dir, pem2(w.schluessel(w.gut)), "tls:", "x509:"}
	for _, f := range append(menge, paare...) {
		daten := filepath.Join(t.TempDir(), "d")
		env := []string{"MODE=record", "UPSTREAM_URL=https://example.test/odata", "DATA_DIR=" + daten}
		if f.cert != "" {
			env = append(env, listenCert+"="+f.cert)
		}
		if f.key != "" {
			env = append(env, listenKey+"="+f.key)
		}
		var stderr bytes.Buffer
		k, code := bootstrap.LadeOderMelde(env, &stderr)
		var soll []string
		for _, v := range f.erwartet {
			soll = append(soll, zeile(v.Schluessel, v.Grund))
		}
		if got := strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n"); code != 1 || !slices.Equal(got, soll) || k.Modus != "" {
			t.Errorf("%s: Exit %d, Zeilen %q; erwartet Exit 1 und %q", f.name, code, got, soll)
		}
		if _, err := os.Stat(daten); err == nil {
			t.Errorf("%s: DATA_DIR wurde angelegt", f.name)
		}
		for _, v := range verboten {
			if strings.Contains(stderr.String(), v) {
				t.Errorf("%s: die Meldung nennt %q", f.name, v)
			}
		}
	}
}

// TestKonfigListenTLSModi (B03): dieselben Verstöße in jedem Modus, auch bei ungültigem MODE (SPEC-291 Nr. 7 a, e; SPEC-110 Nr. 8). Menge
// Modi, 5 Elemente, je mit zwei Verstößen (Zertifikat ohne PEM, Schlüssel ungesetzt).
func TestKonfigListenTLSModi(t *testing.T) {
	w := neueWerkstatt(t)
	f := w.paar(tlsNamen{listenCert, listenKey}, zKeinPEM, zLeer)
	tlsZeilen := []string{zeile(listenCert, gKeinZert), zeile(listenKey, gErforderlich(listenCert))}
	if len(vierPlusEins()) != 5 {
		t.Fatalf("Menge Modi hat %d Elemente, erwartet 5", len(vierPlusEins()))
	}
	for _, modus := range vierPlusEins() {
		soll := slices.Clone(tlsZeilen)
		if modus == "Record" {
			soll = append(soll, zeile("MODE", "Wert nicht in der Aufzählung: record, replay, mock, hybrid"))
		}
		if _, zeilen := lade(t, tlsUmgebung(t, modus, f.cert, f.key)...); !slices.Equal(zeilen, soll) {
			t.Errorf("Modus %q: Zeilen %q, erwartet %q", modus, zeilen, soll)
		}
	}
}

// TestKonfigListenTLSKombinationen (B02, B03): die Klasse „jede Kombination der beiden Schlüssel gesetzt, leer oder fehlend, in jedem Modus“ mit dem Orakel der
// Tabelle von SPEC-291 Nr. 2 (erwartetPaar, ohne den Code des Bausteins). Die Zustände je Schlüssel sind die fünf der Menge Q (ungesetzt, nicht lesbar, ohne PEM,
// gültig, besonders), und `ungesetzt` hat drei Formen: fehlende Variable, leerer Wert, nur Leerraum (7 Darstellungen je Schlüssel, 7 × 7 = 49) × 5 Modi = 245.
func TestKonfigListenTLSKombinationen(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	type darstellung struct {
		setze bool
		wert  string
	}
	darstellungen := func(zustand int, pfad string) []darstellung {
		if zustand != zLeer {
			return []darstellung{{true, pfad}}
		}
		return []darstellung{{false, ""}, {true, ""}, {true, " \t\r\n"}}
	}
	zahl := 0
	for c := range zAnzahl {
		for k := range zAnzahl {
			f := w.paar(n, c, k)
			var soll []string
			for _, v := range f.erwartet {
				soll = append(soll, zeile(v.Schluessel, v.Grund))
			}
			for _, dc := range darstellungen(c, f.cert) {
				for _, dk := range darstellungen(k, f.key) {
					for _, modus := range vierPlusEins() {
						zahl++
						env := hybridBasis(t, modus)
						if dc.setze {
							env = append(env, listenCert+"="+dc.wert)
						}
						if dk.setze {
							env = append(env, listenKey+"="+dk.wert)
						}
						want := slices.Clone(soll)
						if modus == "Record" {
							want = append(want, zeile("MODE", "Wert nicht in der Aufzählung: record, replay, mock, hybrid"))
						}
						if _, zeilen := lade(t, env...); !slices.Equal(zeilen, want) {
							t.Errorf("Zertifikat %d (%+v), Schlüssel %d (%+v), Modus %q: Zeilen %q, erwartet %q", c, dc, k, dk, modus, zeilen, want)
						}
					}
				}
			}
		}
	}
	if zahl != 245 {
		t.Errorf("Klasse hat %d Elemente, erwartet 245 (49 Darstellungen × 5 Modi)", zahl)
	}
}

// TestKonfigListenTLSRest (B04): jeder andere gesetzte Schlüssel mit dem Anfang `LISTEN_TLS_` ist `ODR-E2001` („nicht unterstützt“); die
// Schreibweise zählt, ein leerer Wert zählt nicht (SPEC-110 Nr. 2). Menge R: 5 abgelehnte und 3 ignorierte Namen, je in den vier Modi; dazu
// die Klasse jedes Schlüssels aus Anfang und einem Rest von ein oder zwei Zeichen aus A bis Z, 0 bis 9 und `_` (37 + 37² = 1406) und
// zehn Namen in der Nähe der zwei gelieferten Schlüssel.
func TestKonfigListenTLSRest(t *testing.T) {
	const anfang = "LISTEN_TLS_"
	if got := bootstrap.ListenTLSPraefix(); got != anfang || !strings.HasPrefix(listenKey, got) {
		t.Errorf("Anfang der Schlüssel %q, erwartet %q", got, anfang)
	}
	abgelehnt := []string{anfang + "X", anfang + "CLIENT_CA", anfang, anfang + "CERTS", anfang + "KEY_FILE"}
	ignoriert := []struct{ name, wert string }{{anfang + "X", ""}, {anfang + "X", " \t\r\n"}, {"listen_tls_x", "irgendwas"}}
	if len(abgelehnt) != 5 || len(ignoriert) != 3 {
		t.Fatalf("Menge R hat %d und %d Elemente, erwartet 5 und 3", len(abgelehnt), len(ignoriert))
	}
	for _, modus := range vierModi() {
		for _, name := range abgelehnt {
			_, zeilen := lade(t, hybridBasis(t, modus, name+"="+geheimes)...)
			if soll := []string{zeile(name, grundAbgelehnt)}; !slices.Equal(zeilen, soll) {
				t.Errorf("%s %q: Zeilen %q, erwartet %q", modus, name, zeilen, soll)
			}
		}
		for _, f := range ignoriert {
			erwarteGueltig(t, hybridBasis(t, modus, f.name+"="+f.wert)...)
		}
	}
	env := hybridBasis(t, "mock")
	// Namen, die zur Familie nicht gehören (ohne den Unterstrich nach TLS, Anfang von LISTEN), sind jede Variable der Umgebung und bleiben ohne Wirkung.
	for _, name := range []string{"LISTEN_TLS", "LISTEN_TLSX", "LISTEN_TL", "LISTEN_", "LISTEN_X"} {
		erwarteGueltig(t, append(slices.Clone(env), name+"=x")...)
	}
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	namen := []string{"CER", "CERTS", "CERT_", "KE", "KEYS", "KEY_", "CERTKEY", "_CERT", "TLS_CERT", "KEY\x00"}
	for i := range len(alphabet) {
		namen = append(namen, alphabet[i:i+1])
		for j := range len(alphabet) {
			namen = append(namen, alphabet[i:i+1]+alphabet[j:j+1])
		}
	}
	if len(namen) != 10+1406 {
		t.Fatalf("Klasse hat %d Elemente, erwartet 1416", len(namen))
	}
	for _, rest := range namen {
		name := anfang + rest
		_, zeilen := lade(t, append(slices.Clone(env), name+"=x")...)
		if soll := []string{zeile(name, grundAbgelehnt)}; !slices.Equal(zeilen, soll) {
			t.Errorf("Name %q: Zeilen %q, erwartet %q", name, zeilen, soll)
		}
	}
}

// TestKonfigListenTLSNachbarn (B04): die zwei gelieferten Schlüssel sind keine abgelehnten; gesetzt melden sie allein den Grund der Datei.
func TestKonfigListenTLSNachbarn(t *testing.T) {
	for _, name := range []string{listenCert, listenKey} {
		_, zeilen := lade(t, hybridBasis(t, "record", name+"="+filepath.Join(t.TempDir(), "fehlt.pem"))...)
		for _, z := range zeilen {
			if strings.Contains(z, grundAbgelehnt) {
				t.Errorf("%s gesetzt: Zeile %q", name, z)
			}
		}
		if len(zeilen) != 2 {
			t.Errorf("%s gesetzt: Zeilen %q, erwartet je eine für den Schlüssel und den anderen", name, zeilen)
		}
	}
}

// loopbackHosts sind die drei Hosts, aus denen die Konfiguration die Basis ableitet (SPEC-110 Nr. 14 c).
func loopbackHosts() []string { return []string{"127.0.0.1", "[::1]", "localhost"} }

// TestKonfigBasisSchema (B06): die abgeleitete PUBLIC_BASE_URL trägt `https`, wenn beide Schlüssel gesetzt sind, gleich ob die Dateien gültig
// sind, sonst `http`; in record und replay bleibt sie leer (SPEC-291 Nr. 7 d, SPEC-110 Nr. 14 c). Menge H: 6 Elemente in mock; dazu dieselben
// 6 in hybrid, die Fälle mit ungültigen Dateien und mit einem einzelnen Schlüssel (gemessen an der Länge der Wurzel, denn die Konfiguration
// einer Meldung bleibt leer) und zwei Fälle mit TLS in record und replay.
func TestKonfigBasisSchema(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	for _, modus := range []string{"mock", "hybrid"} {
		n := 0
		for _, host := range loopbackHosts() {
			for _, mitTLS := range []bool{true, false} {
				n++
				c, k, schema := "", "", "http://"
				if mitTLS {
					c, k, schema = cert, key, "https://"
				}
				got := erwarteGueltig(t, tlsUmgebung(t, modus, c, k, "LISTEN="+host+":8080")...)
				if soll := schema + host + ":8080"; got.PublicBaseURL != soll {
					t.Errorf("%s %s mit TLS %v: Basis %q, erwartet %q", modus, host, mitTLS, got.PublicBaseURL, soll)
				}
			}
		}
		if n != 6 {
			t.Errorf("Menge H hat %d Elemente, erwartet 6", n)
		}
	}
	for _, modus := range []string{"record", "replay"} {
		if got := erwarteGueltig(t, tlsUmgebung(t, modus, cert, key)...); got.PublicBaseURL != "" {
			t.Errorf("%s mit TLS: Basis %q, erwartet leer", modus, got.PublicBaseURL)
		}
	}
	// Das Schema zählt in der Länge der Wurzel: `localhost:65535` mit Root von 2026 Bytes hat bei `https` 2049 Bytes (Verstoß), bei `http` 2048.
	laenge := zeile("PUBLIC_BASE_URL", "mit dem Root eines Service länger als 2048 Bytes")
	root := "SERVICE_ROOT_PATH=/" + strings.Repeat("a", 2025)
	text, fehlt := w.datei([]byte("kein PEM\n")), w.fehlt()
	for _, f := range []struct {
		name, cert, key string
		https           bool
	}{
		{"gültige Dateien", cert, key, true}, {"beide Dateien fehlen", fehlt, fehlt, true}, {"beide ohne PEM", text, text, true},
		{"Zertifikat allein", cert, "", false}, {"Schlüssel allein", "", key, false}, {"Zertifikat allein, Datei fehlt", fehlt, "", false},
	} {
		_, zeilen := lade(t, tlsUmgebung(t, "mock", f.cert, f.key, "LISTEN=localhost:65535", root)...)
		if got := slices.Contains(zeilen, laenge); got != f.https {
			t.Errorf("%s: Zeilen %q; das Schema https ist %v, erwartet %v", f.name, zeilen, got, f.https)
		}
	}
}

// TestKonfigBasisLaenge (B07): die Länge der Wurzel zählt das Schema: bei TLS ist ein Root, der die Basis auf 2048 Bytes bringt, gültig, einer
// mit einem Byte mehr nicht; ohne TLS liegt die Grenze ein Byte höher (SPEC-110 Nr. 14 e). Menge B, 4 Elemente, Host `localhost:65535`.
func TestKonfigBasisLaenge(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	laenge := zeile("PUBLIC_BASE_URL", "mit dem Root eines Service länger als 2048 Bytes")
	faelle := []struct {
		mitTLS bool
		root   int
		zeilen []string
	}{{true, 2025, nil}, {true, 2026, []string{laenge}}, {false, 2026, nil}, {false, 2027, []string{laenge}}}
	if len(faelle) != 4 {
		t.Fatalf("Menge B hat %d Elemente, erwartet 4", len(faelle))
	}
	for _, f := range faelle {
		c, k := "", ""
		if f.mitTLS {
			c, k = cert, key
		}
		root := "/" + strings.Repeat("a", f.root-1)
		got, zeilen := lade(t, tlsUmgebung(t, "mock", c, k, "LISTEN=localhost:65535", "SERVICE_ROOT_PATH="+root)...)
		if !slices.Equal(zeilen, f.zeilen) || (f.zeilen == nil && len(got.PublicBaseURL)+len(root) != 2048) {
			t.Errorf("TLS %v, Root von %d Bytes: Zeilen %q, erwartet %q", f.mitTLS, f.root, zeilen, f.zeilen)
		}
	}
}

// TestKonfigBasisGesetzt (B08): eine gesetzte PUBLIC_BASE_URL wird nicht gegen TLS abgeglichen (SPEC-291 Nr. 7 d). Menge A, 2 Elemente.
func TestKonfigBasisGesetzt(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	faelle := []struct{ cert, key, basis string }{{cert, key, "http://h.example"}, {"", "", "https://h.example"}}
	if len(faelle) != 2 {
		t.Fatalf("Menge A hat %d Elemente, erwartet 2", len(faelle))
	}
	for _, f := range faelle {
		got := erwarteGueltig(t, tlsUmgebung(t, "mock", f.cert, f.key, "PUBLIC_BASE_URL="+f.basis)...)
		if got.PublicBaseURL != f.basis {
			t.Errorf("TLS %v: Basis %q, erwartet %q unverändert", f.cert != "", got.PublicBaseURL, f.basis)
		}
	}
}

// TestKonfigListenTLSZeit (B09): das Laden prüft die Gültigkeit zur Zeit des Starts: dieselbe Umgebung ergibt mit früherer und späterer Zeit
// verschiedene Ergebnisse (SPEC-291 Nr. 3, SPEC-110 Nr. 13). Der interne Einstieg läuft mit fester Zeit (beide Grenzen, je eine Nanosekunde
// daneben, die Mitte, die Zeit null), LadeKonfiguration liest die Systemuhr.
func TestKonfigListenTLSZeit(t *testing.T) {
	w := neueWerkstatt(t)
	g := w.wurzel.Blatt(t, tlstest.Optionen{})
	von, bis := g.Zertifikat.NotBefore, g.Zertifikat.NotAfter
	env := tlsUmgebung(t, "record", w.datei(g.ZertifikatPEM()), w.datei(w.schluessel(g)))
	aussen := []string{zeile(listenCert, gAusserhalb)}
	zeiten := []struct {
		name  string
		jetzt time.Time
		soll  []string
	}{
		{"NotBefore minus 1 ns", von.Add(-time.Nanosecond), aussen}, {"NotBefore", von, nil}, {"Mitte", von.Add(bis.Sub(von) / 2), nil},
		{"NotAfter", bis, nil}, {"NotAfter plus 1 ns", bis.Add(time.Nanosecond), aussen}, {"Zeit null", time.Time{}, aussen},
	}
	var ergebnisse []string
	for _, z := range zeiten {
		_, err := bootstrap.LadeKonfigurationZu(env, z.jetzt)
		var zeilen []string
		if err != nil {
			zeilen = strings.Split(err.Error(), "\n")
		}
		ergebnisse = append(ergebnisse, strings.Join(zeilen, "|"))
		if !slices.Equal(zeilen, z.soll) {
			t.Errorf("%s: Zeilen %q, erwartet %q", z.name, zeilen, z.soll)
		}
	}
	if ergebnisse[0] == ergebnisse[1] || ergebnisse[1] == ergebnisse[4] {
		t.Errorf("früher, in und später ergeben nicht verschiedene Ergebnisse: %q", ergebnisse)
	}
	jetzt := time.Now()
	for _, f := range []struct {
		name     string
		von, bis time.Time
		soll     []string
	}{
		{"eine Stunde zurück abgelaufen", jetzt.Add(-3 * time.Hour), jetzt.Add(-time.Hour), aussen},
		{"eine Stunde voraus gültig", jetzt.Add(-time.Hour), jetzt.Add(time.Hour), nil},
		{"erst in einer Stunde gültig", jetzt.Add(time.Hour), jetzt.Add(3 * time.Hour), aussen},
	} {
		b := w.wurzel.Blatt(t, tlstest.Optionen{Von: f.von, Bis: f.bis})
		if _, zeilen := lade(t, tlsUmgebung(t, "record", w.datei(b.ZertifikatPEM()), w.datei(w.schluessel(b)))...); !slices.Equal(zeilen, f.soll) {
			t.Errorf("Systemuhr, %s: Zeilen %q, erwartet %q", f.name, zeilen, f.soll)
		}
	}
}

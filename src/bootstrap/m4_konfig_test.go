package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// vierModi sind die lieferbaren Werte von MODE (SPEC-110 Nr. 8).
func vierModi() []string { return []string{"record", "replay", "mock", "hybrid"} }

// hybridBasis ist eine gültige Umgebung im Modus modus mit Upstream und Allowlist; spätere Einträge überschreiben frühere.
func hybridBasis(t *testing.T, modus string, extra ...string) []string {
	t.Helper()
	return append([]string{"MODE=" + modus, "UPSTREAM_URL=https://h.test/odata", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"),
		"FALLBACK_ALLOW_HOSTS=a.example"}, extra...)
}

func zeile(schluessel, grund string) string { return "ODR-E2001 " + schluessel + ": " + grund }

// Texte der Gründe, unabhängig vom Code aus SPEC-110 Nr. 16 c abgeschrieben.
const (
	grundLeer      = "leerer Eintrag"
	grundTrenner   = "Methode und Muster nicht durch ein Leerzeichen getrennt"
	grundMethode   = "Methode ist keine mutierende Methode"
	grundMuster    = "Pfadmuster ungültig"
	grundHost      = "Host ungültig"
	grundKategorie = "Kategorie unbekannt; erlaubt sind unsupported und coverage_insufficient"
	grundOhneHosts = "enabled verlangt mindestens einen Eintrag in FALLBACK_ALLOW_HOSTS"
	grundFormFB    = "Wert nicht in der Aufzählung: disabled, enabled"
	grundPflicht   = "Pflichtfeld fehlt" // SPEC-110 Nr. 9, 16 f: derselbe Text in den Modi record und hybrid
)

const (
	kHosts   = "FALLBACK_ALLOW_HOSTS"
	kWrite   = "FALLBACK_WRITE_ALLOW"
	kUpstr   = "HYBRID_FALLBACK"
	kReplay  = "HYBRID_REPLAY_FALLBACK"
	geheimes = "geheim"
)

// bytefolge ist die Reihenfolge der Meldung (SPEC-110 Nr. 16 g), nicht die des Join von policy.New.
func bytefolge() []string { return []string{kHosts, kWrite, kUpstr, kReplay} }

// TestHybridKonfigFehlerJeUrsache fängt M04: jede der sieben Ursachen von SPEC-261 (neun Fälle, leer an drei Listen) × die vier
// Modi; je Fall genau eine Zeile `ODR-E2001 <Schlüssel>: Eintrag <n>: <Text>`, n ab 1 mit Doppelten vor dem Eintrag mitgezählt,
// ohne den Wert (SPEC-110 Nr. 12, 16 c).
func TestHybridKonfigFehlerJeUrsache(t *testing.T) {
	faelle := []struct{ ursache, schluessel, wert, grund string }{
		{"leer", kHosts, "a.example,a.example,", "Eintrag 3: " + grundLeer},
		{"leer", kWrite, "POST /a,POST /a,", "Eintrag 3: " + grundLeer},
		{"leer", kReplay, "unsupported,unsupported,", "Eintrag 3: " + grundLeer},
		{"trenner", kWrite, "POST /a,POST /a," + geheimes, "Eintrag 3: " + grundTrenner},
		{"methode", kWrite, "POST /a,POST /a,GET /" + geheimes, "Eintrag 3: " + grundMethode},
		{"muster", kWrite, "POST /a,POST /a,POST " + geheimes, "Eintrag 3: " + grundMuster},
		{"host", kHosts, "a.example,a.example,http://" + geheimes, "Eintrag 3: " + grundHost},
		{"kategorie", kReplay, "unsupported,unsupported," + geheimes, "Eintrag 3: " + grundKategorie},
		{"ohne-hosts", kUpstr, "enabled", grundOhneHosts},
	}
	ursachen := map[string]bool{}
	n := 0
	for _, f := range faelle {
		ursachen[f.ursache] = true
		for _, modus := range vierModi() {
			n++
			extra := []string{f.schluessel + "=" + f.wert}
			if f.ursache == "ohne-hosts" {
				extra = []string{kUpstr + "=enabled", kHosts + "="}
			}
			_, zeilen := lade(t, hybridBasis(t, modus, extra...)...)
			if soll := []string{zeile(f.schluessel, f.grund)}; !slices.Equal(zeilen, soll) {
				t.Errorf("%s/%s: Zeilen %q, erwartet %q", modus, f.ursache, zeilen, soll)
			}
		}
	}
	if len(faelle) != 9 || len(ursachen) != 7 || n != 36 {
		t.Errorf("Menge: %d Fälle, %d Ursachen, %d Läufe; erwartet 9, 7, 36", len(faelle), len(ursachen), n)
	}
	for _, u := range []string{policy.UrsacheLeer, policy.UrsacheTrenner, policy.UrsacheMethode, policy.UrsacheMuster,
		policy.UrsacheHost, policy.UrsacheKategorie, policy.UrsacheOhneHosts} {
		if !ursachen[u] {
			t.Errorf("Ursache %q der Policy hat keinen Fall", u)
		}
	}
}

// TestHybridKonfigFehlerErsterJeSchluessel fängt M04: zwei Verstöße am selben Schlüssel in beiden Folgen melden einen Grund, den
// ersten der Liste (SPEC-110 Nr. 16 c).
func TestHybridKonfigFehlerErsterJeSchluessel(t *testing.T) {
	for _, f := range [][2]string{{"a.example,,http://" + geheimes, "Eintrag 2: " + grundLeer}, {"http://" + geheimes + ",,a.example", "Eintrag 1: " + grundHost}} {
		if _, zeilen := lade(t, hybridBasis(t, "hybrid", kHosts+"="+f[0])...); !slices.Equal(zeilen, []string{zeile(kHosts, f[1])}) {
			t.Errorf("%q: Zeilen %q", f[0], zeilen)
		}
	}
}

// TestHybridKonfigFehlerReihenfolge fängt M04: jedes der sechs Paare von Verstößen an zwei Schlüsseln steht in der Bytefolge der
// Schlüssel (die Umgebung ist eine Menge, eine Folge der Eingabe gibt es nicht); ein Verstoß des Hauptschalters ist die Form (der Zusammenhang `ohne-hosts` schließt einen
// ungültigen Eintrag der Hosts aus). Dazu die zwei Fälle, in denen der Join von policy.New (Hosts, Freigaben, Kategorien,
// Zusammenhang) eine andere Folge hat als die Bytefolge.
func TestHybridKonfigFehlerReihenfolge(t *testing.T) {
	verstoss := map[string][2]string{ // Schlüssel -> Wert, Grund
		kHosts: {"http://" + geheimes, "Eintrag 1: " + grundHost}, kWrite: {geheimes, "Eintrag 1: " + grundTrenner},
		kUpstr: {"x", grundFormFB}, kReplay: {geheimes, "Eintrag 1: " + grundKategorie},
	}
	pruefe := func(name string, schluessel ...string) {
		var env, soll []string
		for _, s := range schluessel {
			env = append(env, s+"="+verstoss[s][0])
		}
		for _, s := range bytefolge() {
			if slices.Contains(schluessel, s) {
				soll = append(soll, zeile(s, verstoss[s][1]))
			}
		}
		if _, zeilen := lade(t, hybridBasis(t, "hybrid", env...)...); !slices.Equal(zeilen, soll) {
			t.Errorf("%s: Zeilen %q, erwartet %q", name, zeilen, soll)
		}
	}
	paare := 0
	for i, a := range bytefolge() {
		for _, b := range bytefolge()[i+1:] {
			paare++
			pruefe(a+"+"+b, a, b)
		}
	}
	if paare != 6 {
		t.Errorf("%d Paare, erwartet 6", paare)
	}
	for _, schluessel := range []string{kWrite, kReplay} {
		_, zeilen := lade(t, hybridBasis(t, "hybrid", kUpstr+"=enabled", kHosts+"=", schluessel+"="+verstoss[schluessel][0])...)
		soll := []string{zeile(kUpstr, grundOhneHosts), zeile(schluessel, verstoss[schluessel][1])}
		if schluessel == kWrite {
			soll[0], soll[1] = soll[1], soll[0]
		}
		if !slices.Equal(zeilen, soll) {
			t.Errorf("enabled ohne Hosts mit %s: Zeilen %q, erwartet %q", schluessel, zeilen, soll)
		}
	}
}

// TestHybridKonfigFehlerStart fängt M04: ein Verstoß an einem der vier Schlüssel beendet den Start mit Exit 1 und der Zeile, ohne Ereignis
// des Hochfahrens und ohne Listener, in jedem der beiden Modi mit Haken; der Wert steht nicht in der Meldung.
func TestHybridKonfigFehlerStart(t *testing.T) {
	for _, modus := range []string{modusRecord, modusReplay} {
		ctx, cancel := context.WithCancel(context.Background())
		h := neuHaken()
		code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modus, 1, kHosts+"=http://"+geheimes), h)
		cancel()
		if soll := []string{zeile(kHosts, "Eintrag 1: "+grundHost)}; code != 1 || !slices.Equal(meldungsZeilen(stderr), soll) {
			t.Errorf("%s: Rückgabe %d, Meldungen %q", modus, code, meldungsZeilen(stderr))
		}
		if ev := h.p.liste(); len(ev) != 0 || h.offeneListener() != 0 || strings.Contains(stderr, geheimes) {
			t.Errorf("%s: Ereignisse %v, Listener %d, stderr %q", modus, ev, h.offeneListener(), stderr)
		}
	}
}

// TestHybridUpstreamPflicht fängt M05: Aufbau {ein Service ohne und mit URL, Datei mit zwei Einträgen in vier Belegungen} ×
// HYBRID_FALLBACK {disabled, enabled}; nur bei enabled meldet jeder Service ohne URL `Pflichtfeld fehlt`, am Schlüssel der URL und
// nie am Schlüssel des Hauptschalters (SPEC-110 Nr. 16 f).
func TestHybridUpstreamPflicht(t *testing.T) {
	const url = `"upstream_url":"https://h.test"`
	aufbauten := []struct {
		name   string
		datei  string
		fehlen []string
	}{
		{"einzeln ohne", "", []string{"UPSTREAM_URL"}},
		{"einzeln mit", "-", nil},
		{"Datei, nur [0] ohne", `[{"id":"a","service_root":"/a"},{"id":"b","service_root":"/b",` + url + `}]`, []string{"SERVICES_FILE[0].upstream_url"}},
		{"Datei, nur [1] ohne", `[{"id":"a","service_root":"/a",` + url + `},{"id":"b","service_root":"/b"}]`, []string{"SERVICES_FILE[1].upstream_url"}},
		{"Datei, beide ohne", `[{"id":"a","service_root":"/a"},{"id":"b","service_root":"/b","upstream_url":null}]`,
			[]string{"SERVICES_FILE[0].upstream_url", "SERVICES_FILE[1].upstream_url"}},
		{"Datei, beide mit", `[{"id":"a","service_root":"/a",` + url + `},{"id":"b","service_root":"/b",` + url + `}]`, nil},
	}
	n := 0
	for _, a := range aufbauten {
		for _, schalter := range []string{"disabled", "enabled"} {
			n++
			env := []string{"MODE=hybrid", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), kHosts + "=a.example", kUpstr + "=" + schalter}
			switch {
			case a.datei == "-":
				env = append(env, "UPSTREAM_URL=https://h.test")
			case a.datei != "":
				env = append(env, "SERVICES_FILE="+schreibe(t, a.datei))
			}
			var soll []string
			for _, s := range a.fehlen {
				if schalter == "enabled" {
					soll = append(soll, zeile(s, "Pflichtfeld fehlt"))
				}
			}
			if _, zeilen := lade(t, env...); !slices.Equal(zeilen, soll) {
				t.Errorf("%s/%s: Zeilen %q, erwartet %q", a.name, schalter, zeilen, soll)
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge hat %d Elemente, erwartet 12", n)
	}
}

// TestHybridUpstreamPflichtModus fängt M05: Modus {record, replay, mock, hybrid, ungültig} × HYBRID_FALLBACK {gültig enabled, ungültig}
// ohne UPSTREAM_URL; die Pflicht aus hybrid gilt nur im gültigen Modus hybrid bei gültigem enabled, in record gilt die eigene (SPEC-110
// Nr. 9, 16 e, f). Dazu die Paare mit der leeren Allowlist und mit einem ungültigen Eintrag, der `ohne-hosts` ausschließt.
func TestHybridUpstreamPflichtModus(t *testing.T) {
	n := 0
	for _, modus := range []string{"record", "replay", "mock", "hybrid", "Hybrid"} {
		for _, schalter := range []string{"enabled", "Enabled"} {
			n++
			var soll []string
			if schalter == "Enabled" {
				soll = append(soll, zeile(kUpstr, grundFormFB))
			}
			if modus == "Hybrid" {
				soll = append(soll, zeile("MODE", "Wert nicht in der Aufzählung: record, replay, mock, hybrid"))
			}
			if modus == "record" || modus == "hybrid" && schalter == "enabled" {
				soll = append(soll, zeile("UPSTREAM_URL", grundPflicht))
			}
			env := []string{"MODE=" + modus, "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), kHosts + "=a.example", kUpstr + "=" + schalter}
			if _, zeilen := lade(t, env...); !slices.Equal(zeilen, soll) {
				t.Errorf("%s/%s: Zeilen %q, erwartet %q", modus, schalter, zeilen, soll)
			}
		}
	}
	if n != 10 {
		t.Errorf("Menge hat %d Elemente, erwartet 10", n)
	}
	for name, tc := range map[string]struct {
		hosts string
		soll  []string
	}{
		"leere Allowlist":      {"", []string{zeile(kUpstr, grundOhneHosts), zeile("UPSTREAM_URL", "Pflichtfeld fehlt")}},
		"ungültiger Eintrag":   {"http://" + geheimes, []string{zeile(kHosts, "Eintrag 1: "+grundHost), zeile("UPSTREAM_URL", "Pflichtfeld fehlt")}},
		"gültiger Eintrag":     {"a.example", []string{zeile("UPSTREAM_URL", "Pflichtfeld fehlt")}},
		"nur Leerraum (leer)":  {"  ", []string{zeile(kUpstr, grundOhneHosts), zeile("UPSTREAM_URL", "Pflichtfeld fehlt")}},
		"Name none (ein Host)": {"none", []string{zeile("UPSTREAM_URL", "Pflichtfeld fehlt")}},
	} {
		env := []string{"MODE=hybrid", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), kUpstr + "=enabled", kHosts + "=" + tc.hosts}
		if _, zeilen := lade(t, env...); !slices.Equal(zeilen, tc.soll) {
			t.Errorf("%s: Zeilen %q, erwartet %q", name, zeilen, tc.soll)
		}
	}
}

// TestHybridKonfigSyntax fängt M06: Randform × vier Schlüssel (SPEC-110 Nr. 1, 3, 6, 16 b, d); die Menge nennt je Schlüssel die Klassen
// leer, Leerraum, leerer Eintrag, unbekannt, none und Doppelte, und der Zähler hält sie gegen die Tabelle des Rasters.
func TestHybridKonfigSyntax(t *testing.T) {
	type fall struct{ klasse, schluessel, wert, grund, getragen string }
	f := func(klasse, schluessel, wert, grund, getragen string) fall {
		return fall{klasse, schluessel, wert, grund, getragen}
	}
	e := func(n int, grund string) string { return fmt.Sprintf("Eintrag %d: %s", n, grund) }
	faelle := []fall{
		f("leer", kUpstr, "", "", "false"), f("leer", kUpstr, " \t", "", "false"), f("Leerraum", kUpstr, " enabled ", "", "true"),
		f("Leerraum", kUpstr, "\tdisabled ", "", "false"), f("unbekannt", kUpstr, "Enabled", grundFormFB, "false"),
		f("unbekannt", kUpstr, "ENABLED", grundFormFB, "false"), f("unbekannt", kUpstr, "true", grundFormFB, "false"),
		f("unbekannt", kUpstr, "1", grundFormFB, "false"), f("none", kUpstr, "none", grundFormFB, "false"),

		f("leer", kHosts, "", "", ""), f("leer", kHosts, "  ", "", ""), f("Leerraum", kHosts, " a.example , b.example ", "", "a.example|b.example"),
		f("leerer Eintrag", kHosts, "a.example,,b.example", e(2, grundLeer), ""), f("leerer Eintrag", kHosts, ",", e(1, grundLeer), ""),
		f("leerer Eintrag", kHosts, "a.example,", e(2, grundLeer), ""), f("unbekannt", kHosts, "http://h", e(1, grundHost), ""),
		f("unbekannt", kHosts, "a.example b.example", e(1, grundHost), ""), f("none", kHosts, "none", "", "none"),
		f("unbekannt", kHosts, "a.example;b.example", e(1, grundHost), ""),
		f("Doppelte", kHosts, "a.example,a.example", "", "a.example|a.example"),

		f("leer", kWrite, "", "", ""), f("leer", kWrite, "\t", "", ""), f("Leerraum", kWrite, " POST /a , PUT /b/* ", "", "POST /a|PUT /b/*"),
		f("leerer Eintrag", kWrite, "POST /a,,PUT /b", e(2, grundLeer), ""), f("leerer Eintrag", kWrite, "POST /a,", e(2, grundLeer), ""),
		f("unbekannt", kWrite, "GET /a", e(1, grundMethode), ""), f("unbekannt", kWrite, "post /a", e(1, grundMethode), ""),
		f("unbekannt", kWrite, "POST  /a", e(1, grundMuster), ""), f("none", kWrite, "none", e(1, grundTrenner), ""),
		f("unbekannt", kWrite, "POST /a;PUT /b", e(1, grundMuster), ""),
		f("Doppelte", kWrite, "POST /a,POST /a", "", "POST /a|POST /a"),

		f("leer", kReplay, "", "", ""), f("leer", kReplay, " ", "", ""), f("Leerraum", kReplay, " unsupported , coverage_insufficient ", "", "unsupported|coverage_insufficient"),
		f("leerer Eintrag", kReplay, "unsupported,,coverage_insufficient", e(2, grundLeer), ""), f("leerer Eintrag", kReplay, "unsupported,", e(2, grundLeer), ""),
		f("unbekannt", kReplay, "Unsupported", e(1, grundKategorie), ""), f("unbekannt", kReplay, "ODR-E6003", e(1, grundKategorie), ""),
		f("none", kReplay, "none", e(1, grundKategorie), ""), f("Doppelte", kReplay, "unsupported,unsupported", "", "unsupported|unsupported"),
		f("unbekannt", kReplay, "unsupported;coverage_insufficient", e(1, grundKategorie), ""),
	}
	klassen := map[string]map[string]bool{}
	for _, c := range faelle {
		if klassen[c.schluessel] == nil {
			klassen[c.schluessel] = map[string]bool{}
		}
		klassen[c.schluessel][c.klasse] = true
		k, zeilen := lade(t, hybridBasis(t, "hybrid", c.schluessel+"="+c.wert)...)
		var soll []string
		if c.grund != "" {
			soll = []string{zeile(c.schluessel, c.grund)}
		}
		if !slices.Equal(zeilen, soll) {
			t.Errorf("%s=%q: Zeilen %q, erwartet %q", c.schluessel, c.wert, zeilen, soll)
			continue
		}
		if c.grund != "" {
			continue
		}
		got := map[string]string{kUpstr: fmt.Sprint(k.Fallback.Upstream), kHosts: strings.Join(k.Fallback.Hosts, "|"),
			kWrite: strings.Join(k.Fallback.Schreibfreigaben, "|"), kReplay: strings.Join(k.Fallback.ReplayKategorien, "|")}[c.schluessel]
		if got != c.getragen {
			t.Errorf("%s=%q: Konfiguration trägt %q, erwartet %q", c.schluessel, c.wert, got, c.getragen)
		}
	}
	listen := []string{"leer", "Leerraum", "leerer Eintrag", "unbekannt", "none", "Doppelte"}
	for schluessel, soll := range map[string][]string{kUpstr: {"leer", "Leerraum", "unbekannt", "none"}, kHosts: listen, kWrite: listen, kReplay: listen} {
		if len(klassen[schluessel]) != len(soll) {
			t.Errorf("%s: %d Klassen, erwartet %d", schluessel, len(klassen[schluessel]), len(soll))
		}
		for _, k := range soll {
			if !klassen[schluessel][k] {
				t.Errorf("%s: Klasse %q ohne Fall", schluessel, k)
			}
		}
	}
	if len(faelle) != 41 {
		t.Errorf("Menge hat %d Elemente, erwartet 41", len(faelle))
	}
}

// TestHybridKonfigJederModus fängt M07: die vier Schlüssel gelten in jedem der vier Modi gleich: gültig sind sie kein Verstoß und die
// Konfiguration trägt die gelesenen Werte, ungültig melden sie ihre Zeile (SPEC-110 Nr. 16 a).
func TestHybridKonfigJederModus(t *testing.T) {
	gueltig := policy.Werte{Upstream: true, Hosts: []string{"a.example"}, Schreibfreigaben: []string{"POST /a"}, ReplayKategorien: []string{"unsupported"}}
	ungueltig := map[string][2]string{kUpstr: {"x", grundFormFB}, kHosts: {"http://" + geheimes, "Eintrag 1: " + grundHost},
		kWrite: {geheimes, "Eintrag 1: " + grundTrenner}, kReplay: {geheimes, "Eintrag 1: " + grundKategorie}}
	n := 0
	for _, modus := range vierModi() {
		k := erwarteGueltig(t, hybridBasis(t, modus, kUpstr+"=enabled", kWrite+"=POST /a", kReplay+"=unsupported")...)
		if !slices.Equal(k.Fallback.Hosts, gueltig.Hosts) || k.Fallback.Upstream != gueltig.Upstream ||
			!slices.Equal(k.Fallback.Schreibfreigaben, gueltig.Schreibfreigaben) || !slices.Equal(k.Fallback.ReplayKategorien, gueltig.ReplayKategorien) {
			t.Errorf("%s: Konfiguration trägt %+v, erwartet %+v", modus, k.Fallback, gueltig)
		}
		for _, schluessel := range bytefolge() {
			n++
			v := ungueltig[schluessel]
			if _, zeilen := lade(t, hybridBasis(t, modus, schluessel+"="+v[0])...); !slices.Equal(zeilen, []string{zeile(schluessel, v[1])}) {
				t.Errorf("%s/%s: Zeilen %q", modus, schluessel, zeilen)
			}
		}
	}
	if n != 16 {
		t.Errorf("Menge hat %d Elemente, erwartet 16", n)
	}
}

// TestHybridKonfigMockSchluessel fängt M07: PUBLIC_BASE_URL, STRICT_COVERAGE und ETAG_ENABLED gelten im Modus hybrid wie im Modus mock
// (SPEC-110 Nr. 14, 15): Ableitung aus Loopback, Pflicht bei fremdem Host, Länge der Wurzel, wirksame Werte.
func TestHybridKonfigMockSchluessel(t *testing.T) {
	for _, modus := range []string{"mock", "hybrid"} {
		k := erwarteGueltig(t, hybridBasis(t, modus, "LISTEN=localhost:08081")...)
		if k.PublicBaseURL != "http://localhost:8081" || !k.StrictCoverage || k.EtagEnabled {
			t.Errorf("%s: Basis %q, StrictCoverage %v, EtagEnabled %v", modus, k.PublicBaseURL, k.StrictCoverage, k.EtagEnabled)
		}
		k = erwarteGueltig(t, hybridBasis(t, modus, "STRICT_COVERAGE=false", "ETAG_ENABLED=true", "LISTEN=0.0.0.0:80", "PUBLIC_BASE_URL=https://r.example")...)
		if k.PublicBaseURL != "https://r.example" || k.StrictCoverage || !k.EtagEnabled {
			t.Errorf("%s gesetzt: Basis %q, StrictCoverage %v, EtagEnabled %v", modus, k.PublicBaseURL, k.StrictCoverage, k.EtagEnabled)
		}
		if _, zeilen := lade(t, hybridBasis(t, modus, "LISTEN=0.0.0.0:80")...); !slices.Equal(zeilen,
			[]string{zeile("PUBLIC_BASE_URL", "im Modus mock oder hybrid bei diesem Host erforderlich")}) {
			t.Errorf("%s fremder Host: Zeilen %q", modus, zeilen)
		}
	}
	// Die Wurzel aus Basis (http://h, 8 Bytes) und Root: 2048 Bytes gelten, 2049 nicht, nur in den Modi mock und hybrid.
	for _, modus := range vierModi() {
		for _, root := range []int{2040, 2041} {
			_, zeilen := lade(t, hybridBasis(t, modus, "PUBLIC_BASE_URL=http://h", "SERVICE_ROOT_PATH=/"+strings.Repeat("a", root-1))...)
			var soll []string
			if root == 2041 && (modus == "mock" || modus == "hybrid") {
				soll = []string{zeile("PUBLIC_BASE_URL", "mit dem Root eines Service länger als 2048 Bytes")}
			}
			if !slices.Equal(zeilen, soll) {
				t.Errorf("%s mit Root von %d Bytes: Zeilen %q, erwartet %q", modus, root, zeilen, soll)
			}
		}
	}
}

// TestHybridKonfigFamilien fängt M08: Modus {record, replay, mock, hybrid} × Familie (ADMIN_*, LISTEN_TLS_*, die vier Schlüssel des
// Modus hybrid) = 24 Zellen; `ADMIN_*` und `LISTEN_TLS_*` bleiben in jedem Modus `nicht unterstützt` ohne den Wert, die vier Schlüssel
// sind es nicht mehr (SPEC-110 Nr. 2).
func TestHybridKonfigFamilien(t *testing.T) {
	familien := []struct {
		schluessel, wert string
		abgelehnt        bool
	}{
		{"ADMIN_LISTEN", geheimes, true}, {"LISTEN_TLS_CERT", geheimes, true}, {kUpstr, "disabled", false}, {kHosts, "a.example", false},
		{kWrite, "POST /a", false}, {kReplay, "unsupported", false},
	}
	n := 0
	for _, modus := range vierModi() {
		for _, f := range familien {
			n++
			_, zeilen := lade(t, hybridBasis(t, modus, f.schluessel+"="+f.wert)...)
			var soll []string
			if f.abgelehnt {
				soll = []string{zeile(f.schluessel, "wird von dieser Version nicht unterstützt")}
			}
			if !slices.Equal(zeilen, soll) || strings.Contains(strings.Join(zeilen, "\n"), geheimes) {
				t.Errorf("%s/%s: Zeilen %q, erwartet %q", modus, f.schluessel, zeilen, soll)
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge hat %d Elemente, erwartet 24", n)
	}
}

// TestHybridKonfigModus fängt M09: `hybrid` ist ein gültiger Wert von MODE und die Konfiguration trägt ihn; ein ungültiger Wert nennt die vier
// lieferbaren Werte, und die Syntax der übrigen Schlüssel wird mitgemeldet (SPEC-110 Nr. 8).
func TestHybridKonfigModus(t *testing.T) {
	if k := erwarteGueltig(t, "MODE=hybrid", "DATA_DIR="+filepath.Join(t.TempDir(), "d")); k.Modus != "hybrid" || k.Fallback.Upstream {
		t.Errorf("Modus %q, Hauptschalter %v", k.Modus, k.Fallback.Upstream)
	}
	_, zeilen := lade(t, "MODE=Hybrid", kUpstr+"=x", "DATA_DIR="+filepath.Join(t.TempDir(), "d"))
	soll := []string{zeile(kUpstr, grundFormFB), zeile("MODE", "Wert nicht in der Aufzählung: record, replay, mock, hybrid")}
	if !slices.Equal(zeilen, soll) {
		t.Errorf("Zeilen %q, erwartet %q", zeilen, soll)
	}
}

// TestHybridNeuePolicy fängt M04: NeuePolicy baut aus den gelesenen Werten die Policy und reicht jeden der vier Werte weiter; der Nullwert
// bleibt nicht stehen. Je Wert eine Entscheidung, die ohne ihn `Bleiben` wäre (SPEC-261 Nr. 4).
func TestHybridNeuePolicy(t *testing.T) {
	k := erwarteGueltig(t, hybridBasis(t, "hybrid", kUpstr+"=enabled", kHosts+"=h.test", kWrite+"=POST /svc/Orders", kReplay+"=coverage_insufficient")...)
	pol, err := k.NeuePolicy()
	if err != nil || pol == nil {
		t.Fatalf("Policy %v, Fehler %v", pol, err)
	}
	const ziel = "https://h.test"
	for name, tc := range map[string]struct {
		a    policy.Anfrage
		soll policy.Entscheidung
	}{
		"Upstream: Hauptschalter und Host":    {policy.Anfrage{Methode: "GET", Pfad: "/svc/x", Ziel: ziel, Code: model.CodeE5001}, policy.Upstream},
		"fremder Host bleibt":                 {policy.Anfrage{Methode: "GET", Pfad: "/svc/x", Ziel: "https://fremd.test", Code: model.CodeE5001}, policy.Bleiben},
		"Replay: Kategorie":                   {policy.Anfrage{Methode: "GET", Pfad: "/svc/x", Ziel: ziel, Code: model.CodeE6003}, policy.Replay},
		"Schreibfreigabe: Methode und Muster": {policy.Anfrage{Methode: "POST", Pfad: "/svc/Orders", Ziel: ziel, Code: model.CodeE6001}, policy.Upstream},
		"andere Route bleibt":                 {policy.Anfrage{Methode: "POST", Pfad: "/svc/Other", Ziel: ziel, Code: model.CodeE6001}, policy.Bleiben},
	} {
		if got := pol.Entscheide(tc.a); got != tc.soll {
			t.Errorf("%s: Entscheidung %v, erwartet %v", name, got, tc.soll)
		}
	}
	// Der Hauptschalter aus: der Replay-Fallback gilt, kein Upstream.
	k = erwarteGueltig(t, hybridBasis(t, "hybrid", kUpstr+"=disabled", kHosts+"=h.test", kReplay+"=coverage_insufficient")...)
	if pol, err = k.NeuePolicy(); err != nil || pol.Entscheide(policy.Anfrage{Methode: "GET", Pfad: "/", Ziel: ziel, Code: model.CodeE5001}) != policy.Bleiben ||
		pol.Entscheide(policy.Anfrage{Methode: "GET", Pfad: "/", Ziel: ziel, Code: model.CodeE6003}) != policy.Replay {
		t.Errorf("disabled: Fehler %v", err)
	}
}

// TestHybridNeuePolicyFehler fängt M04: ein Fehler von policy.New ist ein *KonfigFehler mit ODR-E2001, in der Bytefolge und ohne Wert, und
// die Policy ist nil (SPEC-110 Nr. 16 c, g).
func TestHybridNeuePolicyFehler(t *testing.T) {
	k := bootstrap.Konfiguration{Fallback: policy.Werte{Upstream: true, Hosts: nil, Schreibfreigaben: []string{geheimes}, ReplayKategorien: []string{"x", ""}}}
	pol, err := k.NeuePolicy()
	var kf *bootstrap.KonfigFehler
	if pol != nil || !errors.As(err, &kf) || kf.Code() != model.CodeE2001 {
		t.Fatalf("Policy %v, Fehler %v", pol, err)
	}
	soll := []string{zeile(kWrite, "Eintrag 1: "+grundTrenner), zeile(kUpstr, grundOhneHosts), zeile(kReplay, "Eintrag 1: "+grundKategorie)}
	if !slices.Equal(kf.Zeilen, soll) {
		t.Errorf("Zeilen %q, erwartet %q", kf.Zeilen, soll)
	}
}

// TestUpstreamPflichtGrundGleich fängt M05: fehlt die UPSTREAM_URL, ist der Grund in den Modi record und hybrid (bei gültigem enabled) derselbe Text
// `Pflichtfeld fehlt` an demselben Schlüssel, ohne Modus im Text (SPEC-110 Nr. 9, 16 f); 2 Modi × 5 Aufbauten (einzeln, Datei mit [0], [1],
// beiden und `null` ohne URL); beide Modi gleichen derselben erwarteten Zeile, also einander.
func TestUpstreamPflichtGrundGleich(t *testing.T) {
	const url = `"upstream_url":"https://h.test"`
	aufbauten := []struct {
		datei  string
		fehlen []string
	}{
		{"", []string{"UPSTREAM_URL"}},
		{`[{"id":"a","service_root":"/a"},{"id":"b","service_root":"/b",` + url + `}]`, []string{"SERVICES_FILE[0].upstream_url"}},
		{`[{"id":"a","service_root":"/a",` + url + `},{"id":"b","service_root":"/b"}]`, []string{"SERVICES_FILE[1].upstream_url"}},
		{`[{"id":"a","service_root":"/a"},{"id":"b","service_root":"/b"}]`, []string{"SERVICES_FILE[0].upstream_url", "SERVICES_FILE[1].upstream_url"}},
		{`[{"id":"a","service_root":"/a","upstream_url":null},{"id":"b","service_root":"/b",` + url + `}]`, []string{"SERVICES_FILE[0].upstream_url"}},
	}
	n := 0
	for _, a := range aufbauten {
		for _, modus := range []string{"record", "hybrid"} {
			n++
			env := []string{"MODE=" + modus, "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), kHosts + "=a.example", kUpstr + "=enabled"}
			if a.datei != "" {
				env = append(env, "SERVICES_FILE="+schreibe(t, a.datei))
			}
			_, zeilen := lade(t, env...)
			var soll []string
			for _, s := range a.fehlen {
				soll = append(soll, zeile(s, grundPflicht))
			}
			if !slices.Equal(zeilen, soll) {
				t.Errorf("%s, Aufbau %q: Zeilen %q, erwartet %q", modus, a.fehlen, zeilen, soll)
			}
		}
	}
	if n != 10 {
		t.Errorf("Menge hat %d Elemente, erwartet 10", n)
	}
}

// TestHybridUngueltigerSchalterOhneFolgen fängt M04 und M05: ein ungültiges HYBRID_FALLBACK gilt weder als enabled noch als disabled; es meldet allein
// seine Form, `ohne-hosts` und die Pflicht von UPSTREAM_URL entfallen (SPEC-110 Nr. 16 e). 4 Modi × 8 ungültige Werte × 3 Belegungen der
// Allowlist (nicht gesetzt, leer, nur Leerraum), ohne UPSTREAM_URL außer im Modus record, dessen eigene Pflicht gilt.
func TestHybridUngueltigerSchalterOhneFolgen(t *testing.T) {
	werte := []string{"Enabled", "ENABLED", "eNabled", "true", "1", "none", "yes", "enabled,disabled"}
	hosts := [][]string{nil, {kHosts + "="}, {kHosts + "=  "}}
	n := 0
	for _, modus := range vierModi() {
		for _, w := range werte {
			for _, h := range hosts {
				n++
				env := append([]string{"MODE=" + modus, "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), kUpstr + "=" + w}, h...)
				if modus == "record" {
					env = append(env, "UPSTREAM_URL=https://h.test/odata")
				}
				if _, zeilen := lade(t, env...); !slices.Equal(zeilen, []string{zeile(kUpstr, grundFormFB)}) {
					t.Errorf("%s/%s/%q: Zeilen %q", modus, w, h, zeilen)
				}
			}
		}
	}
	if n != 96 {
		t.Errorf("Menge hat %d Elemente, erwartet 96", n)
	}
}

// TestHybridDisabledOhneHostKeinVerstoss fängt M04: bei disabled (gesetzt und als Default) ist eine Freigabe oder eine Kategorie ohne Host in FALLBACK_ALLOW_HOSTS
// kein Verstoß, und die Konfiguration trägt die Werte (SPEC-110 Nr. 16 e). 4 Modi × 2 Schalter × 4 Belegungen von Freigabe und Kategorie × 2 Belegungen der
// Allowlist (nicht gesetzt, leer); der Fall mit enabled und leerer Allowlist meldet dagegen `ohne-hosts` (TestHybridKonfigFehlerReihenfolge).
func TestHybridDisabledOhneHostKeinVerstoss(t *testing.T) {
	schalter := [][]string{{kUpstr + "=disabled"}, nil}
	belegungen := [][]string{nil, {kWrite + "=POST /a"}, {kReplay + "=unsupported"}, {kWrite + "=POST /a", kReplay + "=coverage_insufficient"}}
	hosts := [][]string{nil, {kHosts + "="}}
	n := 0
	for _, modus := range vierModi() {
		for _, s := range schalter {
			for _, b := range belegungen {
				for _, h := range hosts {
					n++
					env := append(append(append([]string{"MODE=" + modus, "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}, s...), b...), h...)
					if modus == "record" {
						env = append(env, "UPSTREAM_URL=https://h.test/odata")
					}
					k, zeilen := lade(t, env...)
					if len(zeilen) != 0 {
						t.Errorf("%s/%q/%q/%q: Zeilen %q, erwartet keine", modus, s, b, h, zeilen)
						continue
					}
					if len(k.Fallback.Hosts) != 0 || k.Fallback.Upstream || len(k.Fallback.Schreibfreigaben) != countPrefix(b, kWrite) ||
						len(k.Fallback.ReplayKategorien) != countPrefix(b, kReplay) {
						t.Errorf("%s/%q/%q: Konfiguration trägt %+v", modus, s, b, k.Fallback)
					}
				}
			}
		}
	}
	if n != 64 {
		t.Errorf("Menge hat %d Elemente, erwartet 64", n)
	}
}

// countPrefix zählt die Zuweisungen in env, die mit dem Schlüssel schluessel beginnen.
func countPrefix(env []string, schluessel string) int {
	n := 0
	for _, e := range env {
		if strings.HasPrefix(e, schluessel+"=") {
			n++
		}
	}
	return n
}

// TestHybridListeLeerraumUmEintrag fängt M06: Leerraum um einen Eintrag der drei Listen wird entfernt (SPEC-110 Nr. 6, 16 b). Die Menge: Zeichen {Leerzeichen,
// Tabulator, LF, CR} × Seite {vor jedem Eintrag, nach jedem Eintrag, beide} × drei Listen = 36 Fälle mit je drei Einträgen; die Seiten stehen
// auch an der Grenze der Liste, die Nr. 1 für den ganzen Wert schon entfernt, und zwischen den Kommas, wo allein Nr. 6 entfernt. Die Zählprüfung hält
// die vier Zeichen gegen die Konstante des Codes; die Zeichen außerhalb prüft TestHybridListeFremderLeerraum.
func TestHybridListeLeerraumUmEintrag(t *testing.T) {
	if got := []rune(bootstrap.Leerraum); len(got) != 4 || !slices.Equal(got, []rune{' ', '\t', '\n', '\r'}) {
		t.Errorf("Leerraum des Codes %q, erwartet Leerzeichen, Tabulator, LF, CR", bootstrap.Leerraum)
	}
	listen := []struct {
		schluessel string
		eintraege  []string
		trage      func(k bootstrap.Konfiguration) []string
	}{
		{kHosts, []string{"a.example", "b.example", "c.example"}, func(k bootstrap.Konfiguration) []string { return k.Fallback.Hosts }},
		{kWrite, []string{"POST /a", "PUT /b", "DELETE /c"}, func(k bootstrap.Konfiguration) []string { return k.Fallback.Schreibfreigaben }},
		{kReplay, []string{"unsupported", "coverage_insufficient", "unsupported"}, func(k bootstrap.Konfiguration) []string { return k.Fallback.ReplayKategorien }},
	}
	n := 0
	for _, l := range listen {
		for _, w := range []string{" ", "\t", "\n", "\r"} {
			for _, s := range seitenPaare() {
				n++
				teile := mitRand(l.eintraege, w, s.vor, s.nach)
				k, zeilen := lade(t, hybridBasis(t, "hybrid", l.schluessel+"="+strings.Join(teile, ","))...)
				if len(zeilen) != 0 || !slices.Equal(l.trage(k), l.eintraege) {
					t.Errorf("%s mit %q, Seiten %+v: Zeilen %q, Konfiguration trägt %q", l.schluessel, w, s, zeilen, l.trage(k))
				}
			}
		}
	}
	if n != 36 {
		t.Errorf("Menge hat %d Elemente, erwartet 36", n)
	}
}

// TestKonfigListenFremderLeerraum fängt M06 an den fünf Listen außerhalb des Modus hybrid (SPEC-110 Nr. 6): ein Zeichen außerhalb des Leerraums am Rand
// eines Eintrags bleibt im Eintrag und macht ihn ungültig. Die Menge: fünf Listen × sechs Zeichen × Seite {vor, nach, beide} = 90 Fälle mit je drei Einträgen;
// erwartet ist je Fall genau ein Verstoß am Schlüssel der Liste (Header-Namen, Query-Parameter-Namen und JSON-Pfade lassen das Zeichen nicht zu).
func TestKonfigListenFremderLeerraum(t *testing.T) {
	n := 0
	for _, l := range listenAusserHybrid() {
		for _, w := range fremdeZeichen() {
			for _, s := range seitenPaare() {
				n++
				teile := mitRand(l.eintraege, w, s.vor, s.nach)
				_, zeilen := lade(t, hybridBasis(t, "record", l.schluessel+"="+strings.Join(teile, ","))...)
				if got := schluessel(t, zeilen); !slices.Equal(got, []string{l.schluessel}) {
					t.Errorf("%s mit %q, Seiten %+v: Verstöße %q, erwartet einen am Schlüssel", l.schluessel, w, s, zeilen)
				}
			}
		}
	}
	if n != 90 {
		t.Errorf("Menge hat %d Elemente, erwartet 90", n)
	}
}

// listenAusserHybrid sind die fünf Listen, die in jedem Modus gelten und im Modus hybrid keine Policy haben, je mit drei gültigen Einträgen.
func listenAusserHybrid() []struct {
	schluessel string
	eintraege  []string
	trage      func(k bootstrap.Konfiguration) []string
} {
	return []struct {
		schluessel string
		eintraege  []string
		trage      func(k bootstrap.Konfiguration) []string
	}{
		{"MATCH_VARY_HEADERS", []string{"X-A", "X-B", "X-C"}, func(k bootstrap.Konfiguration) []string { return k.MatchVaryHeaders }},
		{"VOLATILE_HEADERS", []string{"X-A", "X-B", "X-C"}, func(k bootstrap.Konfiguration) []string { return k.VolatileHeaders }},
		{"REDACT_HEADERS", []string{"X-A", "X-B", "X-C"}, func(k bootstrap.Konfiguration) []string { return k.RedactHeaders }},
		{"REDACT_QUERY_PARAMS", []string{"p_a", "p_b", "p_c"}, func(k bootstrap.Konfiguration) []string { return k.RedactQueryParams }},
		{"REDACT_JSON_PATHS", []string{"$..a", "$..b", "$..c"}, func(k bootstrap.Konfiguration) []string { return k.RedactJSONPaths }},
	}
}

// sauberDienst sind die Felder eines gültigen Dienstes der Services-Datei.
func sauberDienst() map[string]string {
	return map[string]string{"id": "x", "service_root": "/a", "upstream_url": "https://h.test/o"}
}

// dienstMitRand lädt eine Services-Datei mit einem Dienst, dessen Feld feld auf der Seite seite (vor oder nach) das Zeichen w trägt.
func dienstMitRand(t *testing.T, feld, seite, w string) (bootstrap.Konfiguration, []string) {
	t.Helper()
	werte := sauberDienst()
	if seite == "vor" {
		werte[feld] = w + werte[feld]
	} else {
		werte[feld] += w
	}
	roh, err := json.Marshal([]map[string]string{werte})
	if err != nil {
		t.Fatal(err)
	}
	return lade(t, "MODE=record", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, string(roh)))
}

// seitenPaare sind die Seiten, auf denen ein Zeichen an einem Eintrag steht: vor, nach, beide.
func seitenPaare() []struct{ vor, nach bool } {
	return []struct{ vor, nach bool }{{true, false}, {false, true}, {true, true}}
}

// mitRand setzt w vor und oder nach jeden Eintrag.
func mitRand(eintraege []string, w string, vor, nach bool) []string {
	out := make([]string, len(eintraege))
	for i, e := range eintraege {
		if vor {
			e = w + e
		}
		if nach {
			e += w
		}
		out[i] = e
	}
	return out
}

// fremdeZeichen sind Zeichen außerhalb des Leerraums von SPEC-110 Nr. 1, aus Bytewerten gebaut (AGENTS.md §3.18): VT, FF, NEL, NBSP,
// Figure Space und BOM. Sie bleiben im Wert; die 19 Zeichen mit White_Space außerhalb von ASCII führt weissraumNichtASCII.
func fremdeZeichen() []string {
	return []string{"\x0b", "\x0c", "\xc2\x85", "\xc2\xa0", "\xe2\x80\x87", "\xef\xbb\xbf"}
}

// TestHybridListeFremderLeerraum fängt M06: ein Zeichen außerhalb des Leerraums am Rand eines Eintrags bleibt im Eintrag und macht ihn ungültig
// (SPEC-110 Nr. 1, 6, 16 b). Die Menge: sechs Zeichen × Seite {vor, nach, beide} × drei Listen = 54 Fälle mit je drei Einträgen; erwartet ist je Fall genau
// ein Verstoß am Schlüssel mit der Ursache der Policy am ersten Eintrag, und die Konfiguration trägt weder den Eintrag noch seine bereinigte Fassung.
func TestHybridListeFremderLeerraum(t *testing.T) {
	listen := []struct {
		schluessel string
		eintraege  []string
		grund      func(vor bool) string
		trage      func(k bootstrap.Konfiguration) []string
	}{
		{kHosts, []string{"a.example", "b.example", "c.example"}, func(bool) string { return grundHost },
			func(k bootstrap.Konfiguration) []string { return k.Fallback.Hosts }},
		{kWrite, []string{"POST /a", "PUT /b", "DELETE /c"}, func(vor bool) string {
			if vor {
				return grundMethode
			}
			return grundMuster
		}, func(k bootstrap.Konfiguration) []string { return k.Fallback.Schreibfreigaben }},
		{kReplay, []string{"unsupported", "coverage_insufficient", "unsupported"}, func(bool) string { return grundKategorie },
			func(k bootstrap.Konfiguration) []string { return k.Fallback.ReplayKategorien }},
	}
	n := 0
	for _, l := range listen {
		for _, w := range fremdeZeichen() {
			for _, s := range seitenPaare() {
				n++
				teile := mitRand(l.eintraege, w, s.vor, s.nach)
				k, zeilen := lade(t, hybridBasis(t, "hybrid", l.schluessel+"="+strings.Join(teile, ","))...)
				if soll := []string{zeile(l.schluessel, "Eintrag 1: "+l.grund(s.vor))}; !slices.Equal(zeilen, soll) {
					t.Errorf("%s mit %q, Seiten %+v: Zeilen %q, erwartet %q", l.schluessel, w, s, zeilen, soll)
				}
				for _, got := range l.trage(k) {
					if slices.Contains(l.eintraege, got) || slices.Contains(teile, got) {
						t.Errorf("%s mit %q, Seiten %+v: Konfiguration trägt den Eintrag %q", l.schluessel, w, s, got)
					}
				}
			}
		}
	}
	if n != 54 {
		t.Errorf("Menge hat %d Elemente, erwartet 54", n)
	}
	for _, w := range fremdeZeichen() {
		if strings.ContainsAny(w, bootstrap.Leerraum) {
			t.Errorf("Zeichen %q liegt im Leerraum des Codes", w)
		}
	}
}

// TestKonfigFremderLeerraumAnderenStellen fängt M06 an den Stellen außerhalb der Listen, die SPEC-110 Nr. 1 kappt: der Wert eines Schlüssels (MODE), die Prüfung
// von ADMIN_* auf einen gesetzten Wert, UPSTREAM_URL und die Felder der Services-Datei (id, service_root, upstream_url, je vor und nach). Die Menge:
// 6 Zeichen × (1 MODE + 1 ADMIN_X + 2 UPSTREAM_URL + 6 Felder der Datei) = 60 Fälle, dazu zwei Gegenfälle mit den vier Zeichen des Leerraums.
// Erwartet ist ein Verstoß am Schlüssel oder ein Wert, der das Zeichen trägt (ein Pfad der URL darf es führen).
func TestKonfigFremderLeerraumAnderenStellen(t *testing.T) {
	n := 0
	for _, w := range fremdeZeichen() {
		n++
		_, zeilen := lade(t, hybridBasis(t, "record", "MODE="+w+"record")...)
		if soll := []string{zeile("MODE", "Wert nicht in der Aufzählung: record, replay, mock, hybrid")}; !slices.Equal(zeilen, soll) {
			t.Errorf("MODE mit %q: Zeilen %q, erwartet %q", w, zeilen, soll)
		}
		n++
		_, zeilen = lade(t, hybridBasis(t, "record", "ADMIN_X="+w)...)
		if soll := []string{zeile("ADMIN_X", "wird von dieser Version nicht unterstützt")}; !slices.Equal(zeilen, soll) {
			t.Errorf("ADMIN_X mit %q: Zeilen %q, erwartet %q", w, zeilen, soll)
		}
		for _, seite := range []string{"vor", "nach"} {
			n++
			sauber := "https://h.test/odata"
			wert := w + sauber
			if seite == "nach" {
				wert = sauber + w
			}
			k, zeilen := lade(t, hybridBasis(t, "record", "UPSTREAM_URL="+wert)...)
			switch {
			case len(zeilen) > 0:
				if got := schluessel(t, zeilen); !slices.Equal(got, []string{"UPSTREAM_URL"}) {
					t.Errorf("UPSTREAM_URL %s, Zeichen %q: Verstöße %q", seite, w, zeilen)
				}
			case len(k.Dienste) != 1 || k.Dienste[0].UpstreamURL == sauber || !strings.Contains(k.Dienste[0].UpstreamURL, w):
				t.Errorf("UPSTREAM_URL %s, Zeichen %q: Dienste %+v, erwartet den Wert ohne Bereinigung", seite, w, k.Dienste)
			}
		}
		for _, feld := range []string{"id", "service_root", "upstream_url"} {
			for _, seite := range []string{"vor", "nach"} {
				n++
				sauber := sauberDienst()
				k, zeilen := dienstMitRand(t, feld, seite, w)
				// Das Zeichen bleibt im Feld: entweder ist das Feld ungültig, oder der Wert trägt das Zeichen (ein Pfad der URL darf es führen).
				if len(zeilen) > 0 {
					if got := schluessel(t, zeilen); !slices.Equal(got, []string{"SERVICES_FILE[0]." + feld}) {
						t.Errorf("Feld %s, %s, Zeichen %q: Verstöße %q", feld, seite, w, zeilen)
					}
					continue
				}
				if len(k.Dienste) != 1 {
					t.Fatalf("Feld %s, %s, Zeichen %q: Dienste %+v", feld, seite, w, k.Dienste)
				}
				d := k.Dienste[0]
				if got := map[string]string{"id": d.ID, "service_root": d.Root, "upstream_url": d.UpstreamURL}[feld]; got == sauber[feld] || !strings.Contains(got, w) {
					t.Errorf("Feld %s, %s, Zeichen %q: Wert %q ist bereinigt", feld, seite, w, got)
				}
			}
		}
	}
	if n != 60 {
		t.Errorf("Menge hat %d Elemente, erwartet 60", n)
	}
	if _, zeilen := lade(t, hybridBasis(t, "record", "ADMIN_X= \t\r\n")...); len(zeilen) != 0 {
		t.Errorf("ADMIN_X mit Leerraum allein: Zeilen %q, erwartet keine", zeilen)
	}
	if _, zeilen := lade(t, hybridBasis(t, "record", "MODE=\t record\r\n")...); len(zeilen) != 0 {
		t.Errorf("MODE mit Leerraum am Rand: Zeilen %q, erwartet keine", zeilen)
	}
}

// TestKonfigLeerraumJedeStelle fängt M06 an jeder Stelle, die kappt (SPEC-110 Nr. 1, 6): jedes der vier Zeichen des Leerraums wird auf jeder Seite entfernt.
// Die Menge: Zeichen {Leerzeichen, Tabulator, LF, CR} × Stelle {Wert (MODE) vor und nach, ADMIN_X allein, fünf Listen außerhalb von hybrid je vor und nach,
// UPSTREAM_URL vor und nach, drei Felder der Services-Datei je vor und nach} = 4 × 21 = 84 Fälle; die Einträge der drei Listen von hybrid führt
// TestHybridListeLeerraumUmEintrag (36). Erwartet ist je Fall kein Verstoß und der Wert ohne das Zeichen.
func TestKonfigLeerraumJedeStelle(t *testing.T) {
	const url = "https://h.test/odata"
	n := 0
	for _, w := range []string{" ", "\t", "\n", "\r"} {
		for _, wert := range []string{w + "record", "record" + w} {
			n++
			if k, zeilen := lade(t, hybridBasis(t, "record", "MODE="+wert)...); len(zeilen) != 0 || k.Modus != "record" {
				t.Errorf("MODE=%q: Zeilen %q, Modus %q", wert, zeilen, k.Modus)
			}
		}
		n++
		if _, zeilen := lade(t, hybridBasis(t, "record", "ADMIN_X="+w)...); len(zeilen) != 0 {
			t.Errorf("ADMIN_X=%q: Zeilen %q, erwartet keine", w, zeilen)
		}
		for _, l := range listenAusserHybrid() {
			for _, s := range seitenPaare()[:2] {
				n++
				teile := mitRand(l.eintraege, w, s.vor, s.nach)
				k, zeilen := lade(t, hybridBasis(t, "record", l.schluessel+"="+strings.Join(teile, ","))...)
				if len(zeilen) != 0 {
					t.Errorf("%s mit %q, Seiten %+v: Zeilen %q, erwartet keine", l.schluessel, w, s, zeilen)
				}
				for _, e := range l.eintraege {
					if !slices.Contains(l.trage(k), e) {
						t.Errorf("%s mit %q, Seiten %+v: Konfiguration trägt %q, erwartet %q", l.schluessel, w, s, l.trage(k), e)
					}
				}
			}
		}
		for _, wert := range []string{w + url, url + w} {
			n++
			if k, zeilen := lade(t, hybridBasis(t, "record", "UPSTREAM_URL="+wert)...); len(zeilen) != 0 || len(k.Dienste) != 1 || k.Dienste[0].UpstreamURL != url {
				t.Errorf("UPSTREAM_URL=%q: Zeilen %q, Dienste %+v, erwartet %q", wert, zeilen, k.Dienste, url)
			}
		}
		for _, feld := range []string{"id", "service_root", "upstream_url"} {
			for _, seite := range []string{"vor", "nach"} {
				n++
				k, zeilen := dienstMitRand(t, feld, seite, w)
				sauber := sauberDienst()
				if len(zeilen) != 0 || len(k.Dienste) != 1 || k.Dienste[0].ID != sauber["id"] || k.Dienste[0].Root != sauber["service_root"] || k.Dienste[0].UpstreamURL != sauber["upstream_url"] {
					t.Errorf("Feld %s, %s, Zeichen %q: Zeilen %q, Dienste %+v, erwartet %v", feld, seite, w, zeilen, k.Dienste, sauber)
				}
			}
		}
	}
	if n != 84 {
		t.Errorf("Menge hat %d Elemente, erwartet 84", n)
	}
}

// weissraumNichtASCII sind die 19 Zeichen mit der Eigenschaft White_Space außerhalb von ASCII (U+0085, U+00A0, U+1680, U+2000 bis U+200A, U+2028, U+2029,
// U+202F, U+205F, U+3000), aus Zahlenwerten gebaut (AGENTS.md §3.18). Mit den sechs in ASCII sind es 25 (SPEC-201 Nr. 2 e); 21 liegen außerhalb des Leerraums.
func weissraumNichtASCII() []string {
	var out []string
	for _, r := range []rune{0x85, 0xa0, 0x1680, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000} {
		out = append(out, string(r))
	}
	for r := rune(0x2000); r <= 0x200a; r++ {
		out = append(out, string(r))
	}
	return out
}

// TestKonfigWeissraumAusserhalbBleibt fängt M06 für jedes der 19 Zeichen von weissraumNichtASCII: es bleibt am Rand eines Wertes im Wert und macht ihn
// ungültig (SPEC-110 Nr. 1, 6). Die Menge: 19 Zeichen × Stelle {MODE, REDACT_HEADERS} × Seite {vor, nach} = 76 Fälle; erwartet ist je Fall genau ein Verstoß
// am Schlüssel. Das fängt jede Erweiterung von kappe um eines dieser Zeichen, auch bei unveränderter Konstante leerraum.
func TestKonfigWeissraumAusserhalbBleibt(t *testing.T) {
	if got := len(weissraumNichtASCII()); got != 19 {
		t.Fatalf("Menge hat %d Zeichen, erwartet 19 (25 White_Space abzüglich 6 in ASCII)", got)
	}
	n := 0
	for _, w := range weissraumNichtASCII() {
		for _, c := range []struct{ schluessel, wert string }{{"MODE", "record"}, {"REDACT_HEADERS", "X-A"}} {
			for _, wert := range []string{w + c.wert, c.wert + w} {
				n++
				_, zeilen := lade(t, hybridBasis(t, "record", c.schluessel+"="+wert)...)
				if got := schluessel(t, zeilen); !slices.Equal(got, []string{c.schluessel}) {
					t.Errorf("%s=%q: Verstöße %q, erwartet einen am Schlüssel", c.schluessel, wert, zeilen)
				}
			}
		}
	}
	if n != 76 {
		t.Errorf("Menge hat %d Elemente, erwartet 76", n)
	}
}

// zeichenMengeZ ist die Menge Z der Zeichen, die ein Eintrag von REDACT_QUERY_PARAMS tragen könnte (SPEC-110 Nr. 6), aus Bytewerten gebaut
// (AGENTS.md §3.18): die 128 ASCII-Bytes, die 19 Zeichen von weissraumNichtASCII und vier weitere (BOM, Nullbreitenzeichen, der Buchstabe U+00FC
// und das Byte 0xFF, das kein UTF-8 ist).
func zeichenMengeZ() (ascii, weissraum, weitere []string) {
	for b := byte(0); b < 128; b++ {
		ascii = append(ascii, string([]byte{b}))
	}
	return ascii, weissraumNichtASCII(), []string{string(rune(0xfeff)), string(rune(0x200b)), string(rune(0xfc)), string([]byte{0xff})}
}

// queryNameErlaubt ist das Orakel für ein Zeichen aus Z: eines der 90 ASCII-Zeichen von U+0021 bis U+007E außer `=`, `&`, `#`, `,` (SPEC-110 Nr. 6).
func queryNameErlaubt(z string) bool {
	return len(z) == 1 && z[0] >= 0x21 && z[0] <= 0x7e && !strings.Contains("=&#,", z)
}

// TestKonfigQueryNameZeichen fängt F-1 und F-3 auf der Ebene der Prüfung (SPEC-110 Nr. 6): jedes der 151 Zeichen von Z allein und als `a<Zeichen>b`; erlaubt sind
// genau die 90, die 61 anderen (38 in ASCII, 19 mit White_Space, 4 weitere) machen den Namen ungültig; dazu die Grenzen U+0020/U+0021 und U+007E/U+007F
// und der leere Name.
func TestKonfigQueryNameZeichen(t *testing.T) {
	ascii, weissraum, weitere := zeichenMengeZ()
	alle := slices.Concat(ascii, weissraum, weitere)
	if len(ascii) != 128 || len(weissraum) != 19 || len(weitere) != 4 || len(alle) != 151 {
		t.Fatalf("Z hat %d + %d + %d = %d Zeichen, erwartet 128 + 19 + 4 = 151", len(ascii), len(weissraum), len(weitere), len(alle))
	}
	erlaubt, asciiVerboten, andere := 0, 0, 0
	for _, z := range alle {
		soll := queryNameErlaubt(z)
		switch {
		case soll:
			erlaubt++
		case len(z) == 1 && z[0] < 128:
			asciiVerboten++
			andere++
		default:
			andere++
		}
		for _, s := range []string{z, "a" + z + "b"} {
			if got := bootstrap.IstQueryName(s); got != soll {
				t.Errorf("Name %q: gültig %v, erwartet %v", s, got, soll)
			}
		}
	}
	if erlaubt != 90 || asciiVerboten != 38 || erlaubt+asciiVerboten != 128 || andere != 61 {
		t.Errorf("erlaubt %d, in ASCII verboten %d, andere %d; erwartet 90, 38 (90 + 38 = 128) und 61", erlaubt, asciiVerboten, andere)
	}
	for s, soll := range map[string]bool{"": false, "\x20": false, "\x21": true, "\x7e": true, "\x7f": false, "a b": false, "access_token": true} {
		if got := bootstrap.IstQueryName(s); got != soll {
			t.Errorf("Name %q: gültig %v, erwartet %v", s, got, soll)
		}
	}
}

// TestKonfigQueryNameInListe fängt F-1 und F-3 über die Umgebung (REDACT_QUERY_PARAMS, SPEC-110 Nr. 1, 6): die 150 Zeichen von Z ohne U+0000 (die Umgebung trägt kein
// NUL) × Stelle im Namen {Anfang, Mitte, Ende} mit je drei Einträgen = 450 Fälle. Erwartet: ein Zeichen des Leerraums an Anfang und Ende wird entfernt; ein Komma
// in der Mitte trennt zwei Einträge; ein erlaubtes Zeichen steht im Namen; jedes andere Zeichen ist genau ein Verstoß am Schlüssel und nie ein Eintrag der Konfiguration.
func TestKonfigQueryNameInListe(t *testing.T) {
	ascii, weissraum, weitere := zeichenMengeZ()
	z150 := slices.Concat(ascii[1:], weissraum, weitere)
	if len(z150) != 150 {
		t.Fatalf("Z ohne U+0000 hat %d Zeichen, erwartet 150", len(z150))
	}
	basis := hybridBasis(t, "record")
	standard := erwarteGueltig(t, basis...).RedactQueryParams
	namen := []string{"p_a", "p_b", "p_c"}
	gueltig, verstoss := 0, 0
	for _, z := range z150 {
		for _, pos := range []string{"anfang", "mitte", "ende"} {
			teile := make([]string, len(namen))
			for i, n := range namen {
				switch pos {
				case "anfang":
					teile[i] = z + n
				case "mitte":
					teile[i] = n[:1] + z + n[1:]
				default:
					teile[i] = n + z
				}
			}
			var soll []string
			switch {
			case strings.Contains(" \t\n\r", z) && pos != "mitte":
				soll = namen
			case z == "," && pos == "mitte":
				soll = []string{"p", "_a", "_b", "_c"}
			case queryNameErlaubt(z):
				soll = teile
			}
			k, zeilen := lade(t, append(slices.Clone(basis), "REDACT_QUERY_PARAMS="+strings.Join(teile, ","))...)
			if soll == nil {
				verstoss++
				if got := schluessel(t, zeilen); !slices.Equal(got, []string{"REDACT_QUERY_PARAMS"}) {
					t.Errorf("Zeichen %q, Stelle %s: Zeilen %q, erwartet einen Verstoß am Schlüssel", z, pos, zeilen)
				}
				continue
			}
			gueltig++
			eigene := slices.DeleteFunc(slices.Clone(k.RedactQueryParams), func(x string) bool { return slices.Contains(standard, x) })
			soll = slices.Compact(slices.Sorted(slices.Values(soll)))
			if len(zeilen) != 0 || !slices.Equal(eigene, soll) {
				t.Errorf("Zeichen %q, Stelle %s: Zeilen %q, eigene Namen %q, erwartet %q", z, pos, zeilen, eigene, soll)
			}
		}
	}
	if gueltig != 4*2+1+90*3 || verstoss != 150*3-gueltig || gueltig+verstoss != 450 {
		t.Errorf("gültig %d, Verstoß %d; erwartet 279 (8 am Rand, 1 Komma in der Mitte, 270 erlaubte) und 171 von 450", gueltig, verstoss)
	}
}

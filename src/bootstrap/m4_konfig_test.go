package bootstrap_test

import (
	"context"
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
			switch {
			case modus == "record":
				soll = append(soll, zeile("UPSTREAM_URL", "im Modus record erforderlich"))
			case modus == "hybrid" && schalter == "enabled":
				soll = append(soll, zeile("UPSTREAM_URL", "Pflichtfeld fehlt"))
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

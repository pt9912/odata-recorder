package bootstrap_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeichen der Klassen dieser Datei entstehen aus Bytewerten (AGENTS.md §3.18); `%` ist 0x25.

// queryVerboten sind die fünf ASCII-Zeichen im Bereich U+0021 bis U+007E, die kein Zeichen eines Query-Parameter-Namens sind: `=` `&` `#` `,` `%` (SPEC-110 Nr. 6).
func queryVerboten() []byte { return []byte{0x3d, 0x26, 0x23, 0x2c, 0x25} }

// queryNameZeichen sind die 89 Zeichen eines Query-Parameter-Namens (SPEC-110 Nr. 6): U+0021 bis U+007E ohne die fünf verbotenen.
func queryNameZeichen() []byte {
	var out []byte
	for c := byte(0x21); c <= 0x7e; c++ {
		if !slices.Contains(queryVerboten(), c) {
			out = append(out, c)
		}
	}
	return out
}

// headerTokenZeichen sind die 77 Zeichen eines HTTP-Tokens (RFC 9110): 52 Buchstaben, 10 Ziffern und die 15 Zeichen `!#$%&'*+-.^_`|~`.
func headerTokenZeichen() []byte {
	var out []byte
	for c := byte('a'); c <= 'z'; c++ {
		out = append(out, c, c-'a'+'A')
	}
	for c := byte('0'); c <= '9'; c++ {
		out = append(out, c)
	}
	return append(out, 0x21, 0x23, 0x24, 0x25, 0x26, 0x27, 0x2a, 0x2b, 0x2d, 0x2e, 0x5e, 0x5f, 0x60, 0x7c, 0x7e)
}

// jsonNameZeichen sind die 64 Zeichen des Namens eines JSON-Pfads `$..name`: 52 Buchstaben, 10 Ziffern, `_` und `-` (SPEC-110 Nr. 6).
func jsonNameZeichen() []byte {
	out := slices.DeleteFunc(headerTokenZeichen(), func(c byte) bool { return c < '0' || (c > '9' && c < 'A') || (c > 'Z' && c < 'a') || c > 'z' })
	return append(out, '_', '-')
}

// basisOhneModus ist eine Umgebung ohne MODE.
func basisOhneModus(t *testing.T, extra ...string) []string {
	t.Helper()
	return append([]string{"UPSTREAM_URL=https://h.test/odata", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), "FALLBACK_ALLOW_HOSTS=a.example"}, extra...)
}

// TestKonfigRedactionTrifftJedesZeichen fängt F-1 über die Konfiguration und den Redaktor (SPEC-110 Nr. 6, SPEC-115 Nr. 2, 4, 6; LH-FA-14): jedes gültige Zeichen
// jedes Namens jedes Pfades trifft den Parameter, den ein Client so schreibt. Die Menge: 89 Zeichen eines Query-Namens (`a<c>b` trifft `a<c>b=GEHEIM` in RedigiereQuery),
// 77 Zeichen eines Header-Tokens (`X<c>y` trifft `X<c>y` und `x<c>Y` in RedigiereHeader), 64 Zeichen eines JSON-Namens (`$..a<c>b` trifft das Mitglied `a<c>b` und `A<c>B`) = 230 Elemente.
// Die Gegenseite: die fünf verbotenen Zeichen machen den Namen ungültig, und ein Name mit Kodierung (`a%41`, `my%20token`, `tok%2Ben`, `a%zz`) ist ein Verstoß am Schlüssel.
func TestKonfigRedactionTrifftJedesZeichen(t *testing.T) {
	const marker = "[REDACTED]"
	basis := hybridBasis(t, "record")
	n := 0
	if got := len(queryNameZeichen()); got != 89 {
		t.Fatalf("Query-Namen haben %d Zeichen, erwartet 89", got)
	}
	for _, c := range queryNameZeichen() {
		n++
		name := "a" + string([]byte{c}) + "b"
		k := erwarteGueltig(t, append(slices.Clone(basis), "REDACT_QUERY_PARAMS="+name)...)
		got, geaendert := model.RedigiereQuery(name+"=GEHEIM&x=1", k.RedactQueryParams)
		if want := name + "=" + marker + "&x=1"; got != want || !geaendert {
			t.Errorf("Query-Name %q: RedigiereQuery liefert %q (geändert %v), erwartet %q", name, got, geaendert, want)
		}
	}
	if got := len(headerTokenZeichen()); got != 77 {
		t.Fatalf("Header-Token haben %d Zeichen, erwartet 77", got)
	}
	for _, c := range headerTokenZeichen() {
		n++
		eintrag := "X" + string([]byte{c}) + "y"
		k := erwarteGueltig(t, append(slices.Clone(basis), "REDACT_HEADERS="+eintrag)...)
		for _, name := range []string{"X" + string([]byte{c}) + "y", "x" + string([]byte{c}) + "Y"} {
			out := model.RedigiereHeader(map[string][]string{name: {"GEHEIM"}}, k.RedactHeaders, nil)
			if len(out) != 1 {
				t.Errorf("Header-Zeichen %q, Name %q: %d Header im Ergebnis, erwartet 1", string([]byte{c}), name, len(out))
			}
			for h, v := range out {
				if !slices.Equal(v, []string{marker}) {
					t.Errorf("Header-Zeichen %q, Name %q: %q trägt %q, erwartet %q", string([]byte{c}), name, h, v, marker)
				}
			}
		}
	}
	if got := len(jsonNameZeichen()); got != 64 {
		t.Fatalf("JSON-Namen haben %d Zeichen, erwartet 64", got)
	}
	for _, c := range jsonNameZeichen() {
		n++
		pfad := "$..a" + string([]byte{c}) + "b"
		k := erwarteGueltig(t, append(slices.Clone(basis), "REDACT_JSON_PATHS="+pfad)...)
		if !slices.Contains(k.RedactJSONPaths, pfad) {
			t.Errorf("JSON-Pfad %q fehlt in der Konfiguration %q", pfad, k.RedactJSONPaths)
			continue
		}
		p, err := model.ParseJSONPfad(pfad)
		if err != nil {
			t.Errorf("JSON-Pfad %q: ParseJSONPfad meldet %v", pfad, err)
			continue
		}
		for _, mitglied := range []string{"a" + string([]byte{c}) + "b", "A" + string([]byte{c}) + "B"} {
			if !p.Passt(mitglied) {
				t.Errorf("JSON-Pfad %q trifft das Mitglied %q nicht", pfad, mitglied)
			}
		}
	}
	if n != 230 {
		t.Errorf("Menge hat %d Elemente, erwartet 230 (89 + 77 + 64)", n)
	}
	gegenseite := 0
	for _, c := range queryVerboten() {
		gegenseite++
		name := "a" + string([]byte{c}) + "b"
		if bootstrap.IstQueryName(name) {
			t.Errorf("Name %q gilt als gültig", name)
		}
		if c == 0x2c {
			continue
		}
		_, zeilen := lade(t, append(slices.Clone(basis), "REDACT_QUERY_PARAMS="+name)...)
		if got := schluessel(t, zeilen); !slices.Equal(got, []string{"REDACT_QUERY_PARAMS"}) {
			t.Errorf("Name %q: Verstöße %q, erwartet einen am Schlüssel", name, zeilen)
		}
	}
	for _, name := range []string{"a%41", "my%20token", "tok%2Ben", "a%zz"} {
		gegenseite++
		_, zeilen := lade(t, append(slices.Clone(basis), "REDACT_QUERY_PARAMS="+name)...)
		if got := schluessel(t, zeilen); !slices.Equal(got, []string{"REDACT_QUERY_PARAMS"}) {
			t.Errorf("Name %q: Verstöße %q, erwartet einen am Schlüssel", name, zeilen)
		}
	}
	if gegenseite != 9 {
		t.Errorf("Gegenseite hat %d Elemente, erwartet 9 (5 verbotene Zeichen und 4 Kodierungen)", gegenseite)
	}
}

// TestKonfigListeFehlerAnJederPosition fängt F-2 (SPEC-110 Nr. 6): ein ungültiger Eintrag macht die Liste ungültig, an welcher Position er steht. Die Menge: 13 ungültige Einträge
// (MATCH_VARY_HEADERS 1, VOLATILE_HEADERS 1, REDACT_HEADERS 2, REDACT_QUERY_PARAMS 6, REDACT_JSON_PATHS 3) × Position {1, 2, 3} von drei Einträgen, die anderen beiden gültig = 39 Fälle;
// erwartet ist je Fall genau ein Verstoß am Schlüssel.
func TestKonfigListeFehlerAnJederPosition(t *testing.T) {
	faelle := []struct {
		schluessel string
		gueltig    []string
		ungueltig  []string
	}{
		{"MATCH_VARY_HEADERS", []string{"X-A", "X-B"}, []string{"bad name"}},
		{"VOLATILE_HEADERS", []string{"X-A", "X-B"}, []string{"bad name"}},
		{"REDACT_HEADERS", []string{"X-A", "X-B"}, []string{"bad name", "X:Y"}},
		{"REDACT_QUERY_PARAMS", []string{"token", "p_b"}, []string{"bad name", "p=q", "p&q", "p#q", "p%41", "p\x7fq"}},
		{"REDACT_JSON_PATHS", []string{"$..a", "$..b"}, []string{"$..bad name", "$.a", "a"}},
	}
	n := 0
	for _, f := range faelle {
		for _, u := range f.ungueltig {
			for pos := 0; pos < 3; pos++ {
				n++
				e := slices.Insert(slices.Clone(f.gueltig), pos, u)
				_, zeilen := lade(t, hybridBasis(t, "record", f.schluessel+"="+strings.Join(e, ","))...)
				if got := schluessel(t, zeilen); !slices.Equal(got, []string{f.schluessel}) {
					t.Errorf("%s=%q (Position %d): Verstöße %q, erwartet einen am Schlüssel", f.schluessel, e, pos+1, zeilen)
				}
			}
		}
	}
	if n != 39 {
		t.Errorf("Menge hat %d Elemente, erwartet 39", n)
	}
}

// TestKonfigListenSyntaxJederModus fängt F-3 (SPEC-110 Nr. 8): die Syntax aller Listen-Schlüssel gilt in jedem Modus. Die Menge: Modus {record, replay, mock, hybrid, ungesetzt} (5) ×
// Schlüssel {MATCH_VARY_HEADERS, VOLATILE_HEADERS, REDACT_HEADERS, REDACT_QUERY_PARAMS, REDACT_JSON_PATHS, FALLBACK_ALLOW_HOSTS, FALLBACK_WRITE_ALLOW, HYBRID_REPLAY_FALLBACK} (8) = 40 Fälle mit
// je einem ungültigen Eintrag (die Header-Token-Prüfung bei replay ist einer davon); erwartet ist der Verstoß am Schlüssel, bei gesetztem Modus der einzige.
func TestKonfigListenSyntaxJederModus(t *testing.T) {
	ungueltig := []struct{ schluessel, wert string }{
		{"MATCH_VARY_HEADERS", "bad name"}, {"VOLATILE_HEADERS", "bad name"}, {"REDACT_HEADERS", "bad name"}, {"REDACT_QUERY_PARAMS", "bad name"},
		{"REDACT_JSON_PATHS", "$..bad name"}, {kHosts, "http://x"}, {kWrite, "x"}, {kReplay, "x"},
	}
	n := 0
	for _, modus := range append(vierModi(), "ungesetzt") {
		for _, u := range ungueltig {
			n++
			env := basisOhneModus(t, u.schluessel+"="+u.wert)
			if modus != "ungesetzt" {
				env = hybridBasis(t, modus, u.schluessel+"="+u.wert)
			}
			_, zeilen := lade(t, env...)
			got := schluessel(t, zeilen)
			if modus == "ungesetzt" {
				if !slices.Contains(got, u.schluessel) {
					t.Errorf("Modus %s, %s=%q: Verstöße %q, erwartet einen am Schlüssel", modus, u.schluessel, u.wert, zeilen)
				}
			} else if !slices.Equal(got, []string{u.schluessel}) {
				t.Errorf("Modus %s, %s=%q: Verstöße %q, erwartet einen am Schlüssel", modus, u.schluessel, u.wert, zeilen)
			}
		}
	}
	if n != 40 {
		t.Errorf("Menge hat %d Elemente, erwartet 40", n)
	}
}

// TestKonfigQueryNameLaengeUndStelle fängt F-4 (SPEC-110 Nr. 6): jedes Byte des Namens wird geprüft, in jeder Länge. Die Menge: ein Name der Länge 64 mit einem der fünf verbotenen Zeichen
// an jeder der 64 Stellen (320 Fälle) und der Name aus 64 gültigen Zeichen; ein Name aus 4096 Bytes ist gültig, mit `%` an jeder der 4096 Stellen ungültig (4096 Fälle), und je eines der fünf verbotenen Zeichen an
// erster und letzter Stelle (10 Fälle); dazu über die Umgebung der Name aus 4096 Bytes gültig und derselbe mit `%` an letzter Stelle ein Verstoß. Prüfgrenze: die Schleife läuft über len(s) ohne
// Obergrenze; die Tests führen Namen bis 4096 Bytes, ein längerer ist nicht geprüft (akzeptiertes Negativ im Plan).
func TestKonfigQueryNameLaengeUndStelle(t *testing.T) {
	verboten := queryVerboten()
	n := 0
	if !bootstrap.IstQueryName(strings.Repeat("a", 64)) {
		t.Error("Name aus 64 gültigen Zeichen gilt als ungültig")
	}
	for _, c := range verboten {
		for i := 0; i < 64; i++ {
			n++
			b := []byte(strings.Repeat("a", 64))
			b[i] = c
			if bootstrap.IstQueryName(string(b)) {
				t.Errorf("Name der Länge 64 mit %q an Stelle %d gilt als gültig", string([]byte{c}), i)
			}
		}
	}
	gross := strings.Repeat("a", 4096)
	if !bootstrap.IstQueryName(gross) {
		t.Error("Name aus 4096 gültigen Zeichen gilt als ungültig")
	}
	for i := 0; i < 4096; i++ {
		n++
		b := []byte(gross)
		b[i] = 0x25
		if bootstrap.IstQueryName(string(b)) {
			t.Errorf("Name der Länge 4096 mit `%%` an Stelle %d gilt als gültig", i)
		}
	}
	for _, c := range verboten {
		for _, i := range []int{0, 4095} {
			n++
			b := []byte(gross)
			b[i] = c
			if bootstrap.IstQueryName(string(b)) {
				t.Errorf("Name der Länge 4096 mit %q an Stelle %d gilt als gültig", string([]byte{c}), i)
			}
		}
	}
	if n != 320+4096+10 {
		t.Errorf("Menge hat %d Elemente, erwartet 4426", n)
	}
	basis := hybridBasis(t, "record")
	k := erwarteGueltig(t, append(slices.Clone(basis), "REDACT_QUERY_PARAMS="+gross)...)
	if !slices.Contains(k.RedactQueryParams, gross) {
		t.Error("Name aus 4096 Bytes steht nicht in der Konfiguration")
	}
	schlecht := gross[:4095] + string([]byte{0x25})
	_, zeilen := lade(t, append(slices.Clone(basis), "REDACT_QUERY_PARAMS="+schlecht)...)
	if got := schluessel(t, zeilen); !slices.Equal(got, []string{"REDACT_QUERY_PARAMS"}) {
		t.Errorf("Name aus 4096 Bytes mit `%%` am Ende: Verstöße %d, erwartet einen am Schlüssel", len(zeilen))
	}
}

// TestKappeLaesstJedesAndereZeichenStehen fängt F-5 (SPEC-110 Nr. 1): kappe entfernt nur Leerzeichen, Tabulator, LF und CR. Die Menge: jeder Codepunkt von U+0000 bis U+10FFFF außer den vier
// des Leerraums und den 2048 Surrogaten (1112060 Codepunkte), je als `<Zeichen>x` und `x<Zeichen>` unverändert; dazu die 128 Rohbytes 0x80 bis 0xFF je allein, vor und nach `x`. Prüfgrenze: mehrbytige
// ungültige Folgen sind nicht einzeln geprüft (leerraum ist ASCII, und kein Zeichen außer den vier kann den Rand einer Folge treffen); die Folge im Betrieb: keine.
func TestKappeLaesstJedesAndereZeichenStehen(t *testing.T) {
	n := 0
	for r := rune(0); r <= 0x10ffff; r++ {
		if r >= 0xd800 && r <= 0xdfff || strings.ContainsRune(" \t\n\r", r) {
			continue
		}
		n++
		for _, s := range []string{string(r) + "x", "x" + string(r)} {
			if got := bootstrap.Kappe(s); got != s {
				t.Errorf("Kappe(%q) liefert %q, erwartet unverändert (U+%04X)", s, got, r)
			}
		}
	}
	if n != 0x110000-2048-4 {
		t.Errorf("Menge hat %d Codepunkte, erwartet 1112060", n)
	}
	roh := 0
	for b := 0x80; b <= 0xff; b++ {
		roh++
		for _, s := range []string{string([]byte{byte(b)}), string([]byte{byte(b), 'x'}), string([]byte{'x', byte(b)})} {
			if got := bootstrap.Kappe(s); got != s {
				t.Errorf("Kappe(%q) liefert %q, erwartet unverändert (Byte 0x%02X)", s, got, b)
			}
		}
	}
	if roh != 128 {
		t.Errorf("Menge hat %d Rohbytes, erwartet 128", roh)
	}
	for _, w := range []string{" ", "\t", "\n", "\r"} {
		if got := bootstrap.Kappe(w + "x" + w); got != "x" {
			t.Errorf("Kappe(%q) liefert %q, erwartet x", w+"x"+w, got)
		}
	}
}

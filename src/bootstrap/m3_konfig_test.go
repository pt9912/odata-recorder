package bootstrap_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func dreiModi() []string { return []string{"record", "replay", "mock"} }

// TestKonfigUpstreamNurImRecordPflicht fängt A02: Modus × Wert {fehlt, gültig, ungültig} × Quelle {Einzelschlüssel, Datei};
// nur im Modus record ist ein fehlender Wert ein Verstoß, ein ungültiger in jedem (SPEC-110 Nr. 9).
func TestKonfigUpstreamNurImRecordPflicht(t *testing.T) {
	n := 0
	for _, modus := range dreiModi() {
		for _, art := range []string{"fehlt", "gültig", "ungültig"} {
			for _, quelle := range []string{"Einzelschlüssel", "Datei"} {
				n++
				wert := map[string]string{"fehlt": "", "gültig": "https://h.test/odata", "ungültig": "ftp://x"}[art]
				env := []string{"MODE=" + modus, "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}
				key := "UPSTREAM_URL"
				if quelle == "Einzelschlüssel" && wert != "" {
					env = append(env, "UPSTREAM_URL="+wert)
				}
				if quelle == "Datei" {
					key = "SERVICES_FILE[0].upstream_url"
					feld := ""
					if wert != "" {
						feld = `,"upstream_url":"` + wert + `"`
					}
					env = append(env, "SERVICES_FILE="+schreibe(t, `[{"id":"a","service_root":"/a"`+feld+`}]`))
				}
				var soll []string
				switch {
				case art == "ungültig":
					soll = []string{"ODR-E2001 " + key + ": keine absolute http- oder https-URL ohne Zugangsdaten, Query und Fragment"}
				case art == "fehlt" && modus == "record":
					soll = []string{"ODR-E2001 " + key + ": im Modus record erforderlich"}
				}
				if _, zeilen := lade(t, env...); !slices.Equal(zeilen, soll) {
					t.Errorf("%s/%s/%s: Zeilen %v, erwartet %v", modus, art, quelle, zeilen, soll)
				}
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge hat %d Elemente, erwartet 18", n)
	}
}

// TestKonfigStrictCoverage fängt A03: Wert × Modus; die Konfiguration trägt den wirksamen Wert, nie den Nullwert (SPEC-110 Nr. 14 a).
func TestKonfigStrictCoverage(t *testing.T) {
	const form = "ODR-E2001 STRICT_COVERAGE: Wahrheitswert muss true oder false sein"
	faelle := []struct {
		env   []string
		wert  bool
		zeile []string
	}{
		{nil, true, nil}, {[]string{"STRICT_COVERAGE="}, true, nil}, {[]string{"STRICT_COVERAGE=true"}, true, nil},
		{[]string{"STRICT_COVERAGE=false"}, false, nil}, {[]string{"STRICT_COVERAGE=TRUE"}, false, []string{form}},
		{[]string{"STRICT_COVERAGE=1"}, false, []string{form}}, {[]string{"STRICT_COVERAGE=yes"}, false, []string{form}},
	}
	for _, modus := range dreiModi() {
		for _, f := range faelle {
			k, zeilen := lade(t, append(append(basis(t), "MODE="+modus), f.env...)...)
			if !slices.Equal(zeilen, f.zeile) || (zeilen == nil && k.StrictCoverage != f.wert) {
				t.Errorf("%s %v: StrictCoverage %v, Zeilen %v; erwartet %v und %v", modus, f.env, k.StrictCoverage, zeilen, f.wert, f.zeile)
			}
		}
	}
	// STRICT_COVERAGE ungültig zugleich mit MODE ungültig: zwei Zeilen.
	_, zeilen := lade(t, append(basis(t), "MODE=Mock", "STRICT_COVERAGE=1")...)
	if s := schluessel(t, zeilen); !slices.Equal(s, []string{"MODE", "STRICT_COVERAGE"}) {
		t.Errorf("zwei Verstöße: %v", zeilen)
	}
}

// schreiberNimmtAn ist das Orakel von PUBLIC_BASE_URL: der Schreiber der Version 4 nimmt w als Wurzel an (SPEC-110 Nr. 14 b).
func schreiberNimmtAn(w string) bool {
	typ := model.EntityType{Name: "N.T", Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}}}
	m := model.NeuesServiceModel("4.0", []string{"N"}, []model.EntityType{typ}, []model.EntitySet{{Name: "S", EntityType: "N.T"}}, nil)
	_, err := jsoncodec.NewSchreiber().Adresse(m, w, "S", "1")
	return err == nil
}

// TestKonfigPublicBaseURLForm fängt A04: jede Zeichenfolge der Listen × Modi {mock, record}; die Konfiguration nimmt sie genau an,
// wenn der Schreiber sie als Wurzel annimmt, und die Listen nennen die Erwartung vorab (SPEC-110 Nr. 14 b, d).
func TestKonfigPublicBaseURLForm(t *testing.T) {
	gueltig := []string{"http://localhost:8080", "https://recorder.example", "https://h.example/api", "http://[::1]:80", "http://h:99999",
		"http://h/a/b", "http://h/a%20b", "http://127.0.0.1:1", "http://h/a.b_c~d-e", "http://[fe80::1]:9", "http://h/%41"}
	ungueltig := []string{"http://h/", "http://h//x", "HTTP://h", "ftp://h", "http://u@h", "http://h?x", "http://h#x", "http://",
		`http://h\x`, "http://h x", "http://h\tx", "http://h\x01", "http://h\xc3\xa4", "http://h/\xc3\xa4", "http://h:123456", "http://h:",
		"http://h:8080x", "http://h/.", "http://h/..", "http://h/%2E", "http://h/%2e%2e", "http://h/%", "http://h/%4", "http://h/%zz",
		"http://:80", "http://[]", "http://[::1", "http://[g]", "http://h/a b", "http://h/a/", "/h", "h", "https:/h", "http://h/a//b",
		"http://h/" + strings.Repeat("a", 2048), "http://h" + strings.Repeat("/a", 1024) + "/"}
	n := 0
	for _, modus := range []string{"mock", "record"} {
		for erwartet, liste := range map[bool][]string{true: gueltig, false: ungueltig} {
			for _, w := range liste {
				n++
				if schreiberNimmtAn(w) != erwartet {
					t.Fatalf("Orakel nimmt %q nicht mit %v an", w, erwartet)
				}
				_, zeilen := lade(t, append(basis(t), "MODE="+modus, "PUBLIC_BASE_URL="+w)...)
				if erwartet != (len(zeilen) == 0) || (!erwartet && (len(zeilen) != 1 || !strings.HasPrefix(zeilen[0], "ODR-E2001 PUBLIC_BASE_URL: Form ungültig"))) {
					t.Errorf("%s %q: Zeilen %v, gültig erwartet: %v", modus, w, zeilen, erwartet)
				}
			}
		}
	}
	if n != 2*(len(gueltig)+len(ungueltig)) {
		t.Errorf("Menge hat %d Elemente", n)
	}
}

// TestKonfigPublicBaseURLZeichen fängt A04 als Generator: jedes der 256 Bytes an fünf Stellen (Host, Segment, Port, Prozent-Kodierung,
// IPv6-Klammer), Orakel ist der Schreiber der Version 4; die Konfiguration nimmt die Zeichenfolge genau an, wenn er sie annimmt.
func TestKonfigPublicBaseURLZeichen(t *testing.T) {
	stellen := []string{"http://h#x", "http://h/s#x", "http://h:8#0", "http://h/%#1", "http://[a#]"}
	n, angenommen := 0, 0
	for _, st := range stellen {
		for b := range 256 {
			w := strings.Replace(st, "#", string([]byte{byte(b)}), 1)
			n++
			soll := schreiberNimmtAn(w)
			if soll {
				angenommen++
			}
			_, zeilen := lade(t, append(basis(t), "MODE=mock", "PUBLIC_BASE_URL="+w)...)
			if soll != (len(zeilen) == 0) {
				t.Errorf("%q: Schreiber nimmt an: %v, Zeilen der Konfiguration %v", w, soll, zeilen)
			}
		}
	}
	if n != 1280 || angenommen < 100 || angenommen > 1100 {
		t.Errorf("Menge hat %d Elemente, davon %d angenommen", n, angenommen)
	}
}

// TestKonfigPublicBaseURLLaenge fängt A04 (Länge): die Wurzel aus Basis und längstem Root hat im Modus mock die Grenze 2048 Bytes;
// 2048 sind gültig, 2049 ein Verstoß, der Root `/` zählt nichts, bei zwei Diensten zählt der längste, in record und replay gilt keine
// Grenze der Wurzel; die Länge der Basis allein ist Teil der Form (SPEC-110 Nr. 14 b, d).
func TestKonfigPublicBaseURLLaenge(t *testing.T) {
	const basisURL = "http://h"
	const zeile = "ODR-E2001 PUBLIC_BASE_URL: mit dem Root eines Service länger als 2048 Bytes"
	root := func(n int) string { return "/" + strings.Repeat("a", n-1) }
	faelle := []struct {
		name  string
		dazu  []string
		modus string
		zeile []string
	}{
		{"2048 gültig", []string{"SERVICE_ROOT_PATH=" + root(2048-len(basisURL))}, "mock", nil},
		{"2049 Verstoß", []string{"SERVICE_ROOT_PATH=" + root(2049-len(basisURL))}, "mock", []string{zeile}},
		{"Root / zählt nichts", []string{"SERVICE_ROOT_PATH=/"}, "mock", nil},
		{"zwei Dienste, der längste zählt", []string{"SERVICES_FILE=" + schreibe(t,
			`[{"id":"a","service_root":"/a"},{"id":"b","service_root":"`+root(2049-len(basisURL))+`"},{"id":"c","service_root":"/c"}]`)}, "mock", []string{zeile}},
		{"zwei Dienste, der längste passt", []string{"SERVICES_FILE=" + schreibe(t,
			`[{"id":"a","service_root":"/a"},{"id":"b","service_root":"`+root(2048-len(basisURL))+`"}]`)}, "mock", nil},
		{"record ohne Grenze", []string{"SERVICE_ROOT_PATH=" + root(3000)}, "record", nil},
		{"replay ohne Grenze", []string{"SERVICE_ROOT_PATH=" + root(3000)}, "replay", nil},
	}
	for _, f := range faelle {
		env := append(basis(t), "MODE="+f.modus, "PUBLIC_BASE_URL="+basisURL)
		if strings.HasPrefix(f.dazu[0], "SERVICES_FILE=") {
			env = append(env, "UPSTREAM_URL=") // der Einzelschlüssel steht neben der Datei nicht
		}
		if _, zeilen := lade(t, append(env, f.dazu...)...); !slices.Equal(zeilen, f.zeile) {
			t.Errorf("%s: Zeilen %v, erwartet %v", f.name, zeilen, f.zeile)
		}
	}
	// Eine Basis von 2048 Bytes ist mit dem Root `/` gültig und mit jedem weiteren Segment ein Verstoß.
	lang := basisURL + "/" + strings.Repeat("a", 2048-len(basisURL)-1)
	if _, zeilen := lade(t, append(basis(t), "MODE=mock", "PUBLIC_BASE_URL="+lang)...); len(zeilen) != 0 {
		t.Errorf("Basis von 2048 Bytes mit dem Root /: %v", zeilen)
	}
	if _, zeilen := lade(t, append(basis(t), "MODE=mock", "PUBLIC_BASE_URL="+lang, "SERVICE_ROOT_PATH=/x")...); !slices.Equal(zeilen, []string{zeile}) {
		t.Errorf("Basis von 2048 Bytes mit dem Root /x: %v", zeilen)
	}
	// Form und Länge zugleich verletzt: nur die Form.
	_, zeilen := lade(t, append(basis(t), "MODE=mock", "PUBLIC_BASE_URL=http://u@h", "SERVICE_ROOT_PATH="+root(3000))...)
	if len(zeilen) != 1 || !strings.Contains(zeilen[0], "Form ungültig") {
		t.Errorf("Form und Länge: %v", zeilen)
	}
}

// TestKonfigPublicBaseURLAbleitung fängt A05: Host × Port × Modus; im Modus mock leitet die Konfiguration die Basis aus LISTEN ab,
// wenn der Host Loopback ist (Port ohne führende Nullen), sonst ist es ein Verstoß; in den anderen Modi und bei ungültigem MODE
// gibt es keinen Verstoß und der Wert ist leer (SPEC-110 Nr. 14 c).
func TestKonfigPublicBaseURLAbleitung(t *testing.T) {
	hosts := map[string]bool{"127.0.0.1": true, "[::1]": true, "localhost": true, "0.0.0.0": false, "[::]": false, "LOCALHOST": false,
		"127.0.0.2": false, "example.org": false, "[0:0:0:0:0:0:0:1]": false}
	ports := map[string]string{"8080": "8080", "08080": "8080", "000080": "80"}
	const pflicht = "ODR-E2001 PUBLIC_BASE_URL: im Modus mock bei diesem Host erforderlich"
	n := 0
	for host, loopback := range hosts {
		for port, ohneNull := range ports {
			for _, modus := range []string{"mock", "record", "replay", "hybrid", "Mock"} {
				n++
				k, zeilen := lade(t, append(basis(t), "MODE="+modus, "LISTEN="+host+":"+port)...)
				switch {
				case modus == "mock" && loopback:
					if len(zeilen) != 0 || k.PublicBaseURL != "http://"+host+":"+ohneNull {
						t.Errorf("%s %s:%s: %q, %v", modus, host, port, k.PublicBaseURL, zeilen)
					}
				case modus == "mock":
					if !slices.Equal(zeilen, []string{pflicht}) {
						t.Errorf("%s %s:%s: Zeilen %v", modus, host, port, zeilen)
					}
				case modus == "record" || modus == "replay":
					if len(zeilen) != 0 || k.PublicBaseURL != "" {
						t.Errorf("%s %s:%s: %q, %v", modus, host, port, k.PublicBaseURL, zeilen)
					}
				default: // hybrid und ungültig: allein die Zeile MODE
					if len(zeilen) != 1 || !strings.HasPrefix(zeilen[0], "ODR-E2001 MODE: ") {
						t.Errorf("%s %s:%s: Zeilen %v", modus, host, port, zeilen)
					}
				}
			}
		}
	}
	if want := len(hosts) * len(ports) * 5; n != want {
		t.Errorf("Menge hat %d Elemente, erwartet %d", n, want)
	}
	// Ein gesetzter Wert geht der Ableitung vor und erreicht die Konfiguration unverändert.
	if k := erwarteGueltig(t, append(basis(t), "MODE=mock", "LISTEN=0.0.0.0:8080", "PUBLIC_BASE_URL= https://r.example/api ")...); k.PublicBaseURL != "https://r.example/api" {
		t.Errorf("PublicBaseURL %q", k.PublicBaseURL)
	}
}

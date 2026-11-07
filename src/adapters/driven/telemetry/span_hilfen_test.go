package telemetry_test

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// spanZeile ist eine Zeile der Tabelle in spezifikation.md §5: Name und Pflicht-Attribute in der
// Reihenfolge der Tabelle.
type spanZeile struct {
	name    string
	pflicht []string
}

// spanTabelle ist das Orakel der Namen und Pflicht-Attribute als Literal aus dem Vertrag; die
// Zählprüfung vergleicht es mit support.Spans() (gleiche Namen, Schlüssel und Reihenfolge, zusammen elf).
func spanTabelle(t *testing.T) []spanZeile {
	t.Helper()
	tab := []spanZeile{
		{"odata.gateway.request", []string{"http.request.method", "http.response.status_code", "recorder.mode"}},
		{"odata.record.capture", []string{"recorder.session_id", "recorder.capture_status"}},
		{"odata.replay.match", []string{"recorder.match_strategy", "recorder.match_result"}},
		{"odata.mock.query", []string{"recorder.entity_set", "recorder.coverage_status"}},
		{"odata.workspace.mutation", []string{"recorder.workspace_id", "recorder.mutation_type"}},
	}
	quelle, summe := support.Spans(), 0
	if len(quelle) != len(tab) {
		t.Fatalf("support.Spans() hat %d Namen, die Tabelle %d", len(quelle), len(tab))
	}
	for i, q := range quelle {
		summe += len(tab[i].pflicht)
		if string(q.Name) != tab[i].name || !slices.Equal(q.Pflicht, tab[i].pflicht) {
			t.Fatalf("support.Spans()[%d] = %s %v, die Tabelle %s %v", i, q.Name, q.Pflicht, tab[i].name, tab[i].pflicht)
		}
	}
	if summe != 11 {
		t.Fatalf("%d Pflicht-Attribute, erwartet 11", summe)
	}
	return tab
}

// eimerGrenzen ist das Orakel der Grenzen als Literal (SPEC-230 Nr. 8); die Zählprüfung vergleicht es
// mit support.Eimer().
func eimerGrenzen(t *testing.T) []time.Duration {
	t.Helper()
	lit := []time.Duration{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 5000}
	for i := range lit {
		lit[i] *= time.Millisecond
	}
	if !slices.Equal(lit, support.Eimer()) {
		t.Fatalf("support.Eimer() = %v, die Tabelle %v", support.Eimer(), lit)
	}
	return lit
}

// gueltigSpan, anderesSpan und ungueltigSpan liefern zu einem Pflicht-Attribut einen gültigen, einen
// zweiten gültigen und einen ungültigen Wert (SPEC-230 Nr. 3).
func gueltigSpan(key string) support.Attr {
	switch key {
	case "http.request.method":
		return support.Str(key, "GET")
	case "http.response.status_code":
		return support.Int(key, 200)
	case "recorder.session_id":
		return support.Str(key, "ses-0001")
	case "recorder.workspace_id":
		return support.Str(key, "ws-0001")
	case "recorder.entity_set":
		return support.Str(key, "Products")
	}
	return support.Str(key, "record_a")
}

func anderesSpan(key string) support.Attr {
	switch key {
	case "http.request.method":
		return support.Str(key, "POST")
	case "http.response.status_code":
		return support.Int(key, 404)
	case "recorder.session_id":
		return support.Str(key, "ses-0002")
	case "recorder.workspace_id":
		return support.Str(key, "ws-0002")
	case "recorder.entity_set":
		return support.Str(key, "Orders")
	}
	return support.Str(key, "replay_b")
}

func ungueltigSpan(key string) support.Attr {
	if key == "http.response.status_code" {
		return support.Int(key, 99)
	}
	return support.Str(key, "a b")
}

// werteVon bildet zu den Schlüsseln die Attribute nach f.
func werteVon(keys []string, f func(string) support.Attr) []support.Attr {
	out := make([]support.Attr, 0, len(keys))
	for _, k := range keys {
		out = append(out, f(k))
	}
	return out
}

// alleSpanKeys sind die elf Pflicht-Attribute aller Spans.
func alleSpanKeys(t *testing.T) []string {
	var out []string
	for _, z := range spanTabelle(t) {
		out = append(out, z.pflicht...)
	}
	return out
}

// orakelID ist die ID des n-ten Spans eines Adapters: spn- und 16 Dezimalziffern (SPEC-230 Nr. 6).
func orakelID(n int) string {
	z := strconv.Itoa(n)
	return "spn-" + strings.Repeat("0", 16-len(z)) + z
}

// umgebung ist ein Adapter mit FakeSenke und zählender Uhr.
type umgebung struct {
	a     *telemetry.Adapter
	senke *FakeSenke
	uhr   *zaehlUhr
}

func neuUmgebung(t *testing.T, mindest slog.Level, maxSpans int) umgebung {
	t.Helper()
	u := umgebung{senke: &FakeSenke{}, uhr: festeUhr()}
	a, err := telemetry.New(telemetry.Config{Writer: u.senke, Clock: u.uhr, MinLevel: mindest, MaxSpans: maxSpans})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	u.a = a
	return u
}

// neuMitMax baut einen Adapter mit Uhr, Senke und MaxSpans.
func neuMitMax(t *testing.T, w io.Writer, uhr support.Clock, mindest slog.Level, maxSpans int) *telemetry.Adapter {
	t.Helper()
	a, err := telemetry.New(telemetry.Config{Writer: w, Clock: uhr, MinLevel: mindest, MaxSpans: maxSpans})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

// pruefeVektor vergleicht alle sieben Gründe von Dropped mit erwartet (fehlende Einträge gelten 0).
func pruefeVektor(t *testing.T, wo string, s support.Snapshot, erwartet map[support.Verwurf]uint64) {
	t.Helper()
	if len(s.Dropped) != 7 || len(support.Verwuerfe()) != 7 {
		t.Errorf("%s: %d Gründe im Stand, %d in der Quelle; erwartet 7", wo, len(s.Dropped), len(support.Verwuerfe()))
	}
	for _, g := range support.Verwuerfe() {
		if s.Dropped[g] != erwartet[g] {
			t.Errorf("%s: %s zählt %d, erwartet %d", wo, g, s.Dropped[g], erwartet[g])
		}
	}
	for g := range erwartet {
		if !slices.Contains(support.Verwuerfe(), g) {
			t.Fatalf("%s: erwarteter Grund %q gehört nicht zu support.Verwuerfe()", wo, g)
		}
	}
}

// startEnde beginnt einen Span mit attrs und beendet ihn.
func startEnde(a *telemetry.Adapter, name string, attrs ...support.Attr) {
	_, sp := a.StartSpan(context.Background(), support.SpanName(name), attrs...)
	sp.End()
}

// attrsGleich vergleicht zwei Listen von Attributen auf Länge und Elemente; nil und leer sind gleich.
func attrsGleich(x, y []support.Attr) bool { return slices.Equal(x, y) }

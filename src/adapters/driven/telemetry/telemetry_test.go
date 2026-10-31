package telemetry_test

import (
	"bytes"
	"context"
	"errors"
	"go/parser"
	"go/token"
	"log/slog"
	"math"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

func neu(t *testing.T) (*telemetry.Adapter, *bytes.Buffer) {
	t.Helper()
	var puffer bytes.Buffer
	a, err := telemetry.New(telemetry.Config{Writer: &puffer, Clock: supporttest.NewFestUhr(time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC))})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a, &puffer
}

func summe(m map[model.Code]uint64) (n uint64) {
	for _, v := range m {
		n += v
	}
	return n
}

// reihenFest prüft SPEC-230 Nr. 8 an einem Stand nach Aufrufen: die Zahl der Reihen ist fest (33
// Codes, 7 Gründe, 5 Span-Namen mit je 12 Eimern), kein Wert eines Aufrufers bildet einen Schlüssel.
func reihenFest(t *testing.T, wo string, s support.Snapshot) {
	t.Helper()
	if len(s.Errors) != 33 || len(s.Dropped) != 7 || len(s.Durations) != 5 {
		t.Errorf("%s: Errors %d, Dropped %d, Durations %d; erwartet 33, 7, 5", wo, len(s.Errors), len(s.Dropped), len(s.Durations))
	}
	for n, v := range s.Durations {
		if len(v.Counts) != 12 {
			t.Errorf("%s: %s hat %d Eimer, erwartet 12", wo, n, len(v.Counts))
		}
	}
}

// P07, P17: nach Aufrufen mit Werten, die kein Schlüssel sein dürfen, bleibt die Zahl der Reihen
// fest, geprüft nach jedem Aufruf. Menge: Aufrufstelle × Wert: CountError mit 7 Codes außerhalb der
// Tabelle, Log mit den 12 fremden und den 13 gültigen Schlüsseln × 4 Werten (darunter ungültige), und
// StartSpan mit 8 Namen (die 5 gültigen, der leere, ein fremder, einer mit Leerzeichen), danach Set
// und End.
func TestReihenFest(t *testing.T) {
	a, _ := neu(t)
	ctx := context.Background()
	aufrufe := 0
	pruefe := func(wo string) {
		aufrufe++
		reihenFest(t, wo, a.Snapshot())
	}
	for _, c := range []model.Code{0, -1, model.CodeE9001 + 1, math.MaxInt, math.MinInt, 1, 99999} {
		a.CountError(c)
		pruefe("CountError " + strconv.Itoa(int(c)))
	}
	schluessel := []string{"", "x", "timestamp", "level", "msg", "Mode", "mode ", "a b", "ODR-E2001", "senke", "wert", "fehlercode"}
	for _, z := range katalog() {
		if z.menge == "log" {
			schluessel = append(schluessel, z.key)
		}
	}
	for _, k := range schluessel {
		for _, v := range []string{"", "a b", "ok", "x=y"} {
			a.Log(ctx, slog.LevelInfo, "m", support.Str(k, v), support.Int(k, 7))
			pruefe("Log " + k + " " + v)
		}
	}
	namen := append(spanNamen(), "", "unbekannt", "odata.mock.query ")
	for _, n := range namen {
		_, sp := a.StartSpan(ctx, support.SpanName(n), support.Str("x", "y"))
		sp.Set(support.Str("x", "y"))
		sp.End()
		pruefe("StartSpan " + n)
	}
	if aufrufe != 7+25*4+8 || len(namen) != 8 {
		t.Errorf("%d Prüfungen, %d Namen; erwartet %d und 8", aufrufe, len(namen), 7+25*4+8)
	}
}

// P17: jedes verletzende Attribut zählt einen Verwurf, nicht nur die ersten (SPEC-230 Nr. 4). Menge:
// Zahl der Attribute n von 0 bis 40 × 3 Arten (n fremde Schlüssel; n-mal derselbe gültige Schlüssel,
// davon n-1 doppelt; n ungültige Werte reihum über die 13 gültigen Schlüssel) = 123 Aufrufe. Über 40
// Attribute ist die Menge unbeschränkt; der Code zählt in einer Schleife ohne Grenze (Negativ im
// Plan).
func TestVerwurfJeAttribut(t *testing.T) {
	var logKeys []string
	for _, z := range katalog() {
		if z.menge == "log" {
			logKeys = append(logKeys, z.key)
		}
	}
	aufrufe := 0
	for n := 0; n <= 40; n++ {
		fremd, doppelt, ungueltig := []support.Attr{}, []support.Attr{}, []support.Attr{}
		for i := range n {
			fremd = append(fremd, support.Str("fremd"+strconv.Itoa(i), "ok"))
			doppelt = append(doppelt, support.Str("session_id", "ok"))
			ungueltig = append(ungueltig, support.Str(logKeys[i%len(logKeys)], "a b"))
		}
		faelle := []struct {
			art              string
			attrs            []support.Attr
			schluessel, wert uint64
		}{
			{"fremde Schlüssel", fremd, uint64(n), 0},
			{"derselbe Schlüssel", doppelt, uint64(max(n-1, 0)), 0},
			{"ungültige Werte", ungueltig, 0, uint64(n)},
		}
		for _, f := range faelle {
			aufrufe++
			a, _ := neu(t)
			a.Log(context.Background(), slog.LevelInfo, "m", f.attrs...)
			s := a.Snapshot()
			if s.Dropped[support.VerwurfSchluessel] != f.schluessel || s.Dropped[support.VerwurfWert] != f.wert {
				t.Errorf("n=%d, %s: schluessel %d, wert %d; erwartet %d und %d", n, f.art, s.Dropped[support.VerwurfSchluessel], s.Dropped[support.VerwurfWert], f.schluessel, f.wert)
			}
		}
	}
	if aufrufe != 41*3 || len(logKeys) != 13 {
		t.Errorf("%d Aufrufe, %d Schlüssel; erwartet 123 und 13", aufrufe, len(logKeys))
	}
}

// P13 (Seite des Adapters): Adapter und Fake erfüllen beide Ports.
func TestAdapterErfuelltPorts(t *testing.T) {
	a, _ := neu(t)
	f := supporttest.NewFakeTelemetry()
	var (
		_ support.Telemetry       = a
		_ support.TelemetryReader = a
		_ support.Telemetry       = f
		_ support.TelemetryReader = f
	)
}

// P06: CountError(c) erhöht genau den Zähler von c um 1.
func TestCountErrorJeCode(t *testing.T) {
	codes := model.AlleCodes()
	if len(codes) != 33 {
		t.Fatalf("%d Codes, erwartet 33", len(codes))
	}
	for _, c := range codes {
		a, _ := neu(t)
		a.CountError(c)
		s := a.Snapshot()
		if s.Errors[c] != 1 || summe(s.Errors) != 1 {
			t.Errorf("%s: Zähler %d, Summe %d, erwartet 1 und 1", c, s.Errors[c], summe(s.Errors))
		}
		a.CountError(c)
		if s := a.Snapshot(); s.Errors[c] != 2 || summe(s.Errors) != 2 {
			t.Errorf("%s: nach dem zweiten Aufruf Zähler %d, Summe %d, erwartet 2 und 2", c, s.Errors[c], summe(s.Errors))
		}
	}
}

// P07: ein Code außerhalb der Tabelle erhöht keinen Zähler und zählt fehlercode.
func TestCountErrorUnbekannt(t *testing.T) {
	codes := []model.Code{0, -1, model.CodeE9001 + 1, math.MaxInt, math.MinInt}
	for _, c := range codes {
		a, _ := neu(t)
		a.CountError(c)
		s := a.Snapshot()
		reihenFest(t, "Code "+strconv.Itoa(int(c)), s)
		if summe(s.Errors) != 0 || s.Dropped[support.VerwurfFehlercode] != 1 {
			t.Errorf("Code %d: Summe %d, fehlercode %d, erwartet 0 und 1", int(c), summe(s.Errors), s.Dropped[support.VerwurfFehlercode])
		}
		for g, n := range s.Dropped {
			if g != support.VerwurfFehlercode && n != 0 {
				t.Errorf("Code %d: Grund %s zählt %d", int(c), g, n)
			}
		}
	}
	if len(codes) < 4 {
		t.Fatal("weniger als 4 Fälle")
	}
}

// P08: Snapshot trägt jeden Schlüssel, auch mit 0.
func TestSnapshotVollstaendig(t *testing.T) {
	a, _ := neu(t)
	s := a.Snapshot()
	if len(s.Errors) != 33 || len(s.Durations) != 5 || len(s.Dropped) != 7 {
		t.Fatalf("Errors %d, Durations %d, Dropped %d; erwartet 33, 5, 7", len(s.Errors), len(s.Durations), len(s.Dropped))
	}
	for _, c := range model.AlleCodes() {
		if n, ok := s.Errors[c]; !ok || n != 0 {
			t.Errorf("Errors[%s] = %d, vorhanden %t", c, n, ok)
		}
	}
	for _, name := range spanNamen() {
		v, ok := s.Durations[support.SpanName(name)]
		if !ok || len(v.Counts) != 12 || v.Count != 0 || v.Sum != 0 {
			t.Errorf("Durations[%s] = %+v, vorhanden %t; erwartet 12 Counts, Count und Sum 0", name, v, ok)
		}
		for _, n := range v.Counts {
			if n != 0 {
				t.Errorf("Durations[%s]: Eimer ungleich 0", name)
			}
		}
	}
	for _, g := range []string{"senke", "schluessel", "wert", "pflicht", "spanname", "zustand", "fehlercode"} {
		if n, ok := s.Dropped[support.Verwurf(g)]; !ok || n != 0 {
			t.Errorf("Dropped[%s] = %d, vorhanden %t", g, n, ok)
		}
	}
	if s.Spans == nil || len(s.Spans) != 0 {
		t.Errorf("Spans = %v, erwartet leer", s.Spans)
	}
}

// P09: Snapshot ist eine Kopie in beide Richtungen.
func TestSnapshotKopie(t *testing.T) {
	a, _ := neu(t)
	a.CountError(model.CodeE2001)
	a.Log(context.Background(), slog.LevelInfo, "m", support.Str("x", "y"))
	s := a.Snapshot()
	s.Errors[model.CodeE2001] = 99
	s.Dropped[support.VerwurfSchluessel] = 99
	s.Durations[support.SpanMockQuery].Counts[0] = 99
	s.Durations[support.SpanGatewayRequest] = support.Verteilung{Count: 99}
	n := a.Snapshot()
	if n.Errors[model.CodeE2001] != 1 {
		t.Errorf("Änderung an Errors wirkt zurück: %d", n.Errors[model.CodeE2001])
	}
	if n.Dropped[support.VerwurfSchluessel] != 1 {
		t.Errorf("Änderung an Dropped wirkt zurück: %d", n.Dropped[support.VerwurfSchluessel])
	}
	if n.Durations[support.SpanMockQuery].Counts[0] != 0 {
		t.Error("Änderung an Counts wirkt zurück")
	}
	if n.Durations[support.SpanGatewayRequest].Count != 0 {
		t.Error("Änderung an Durations wirkt zurück")
	}
	// ein späterer Aufruf ändert den früheren Stand nicht
	frueher := a.Snapshot()
	a.CountError(model.CodeE2001)
	a.Log(context.Background(), slog.LevelInfo, "m", support.Str("x", "y"))
	if frueher.Errors[model.CodeE2001] != 1 || frueher.Dropped[support.VerwurfSchluessel] != 1 {
		t.Errorf("ein späterer Aufruf ändert den früheren Stand: %+v", frueher)
	}
	// Lesen ändert und zählt nichts
	vor := a.Snapshot()
	nach := a.Snapshot()
	if !reflect.DeepEqual(vor, nach) {
		t.Error("Snapshot ändert den Stand")
	}
}

// P10: New weist ungültige Werte ab.
func TestNewConfig(t *testing.T) {
	var w bytes.Buffer
	uhr := supporttest.NewFestUhr(time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC))
	faelle := []struct {
		name  string
		cfg   telemetry.Config
		ok    bool
		spans int
	}{
		{"nil-Writer", telemetry.Config{Clock: uhr}, false, 0},
		{"nil-Clock", telemetry.Config{Writer: &w}, false, 0},
		{"MaxSpans -1", telemetry.Config{Writer: &w, Clock: uhr, MaxSpans: -1}, false, 0},
		{"MaxSpans 0", telemetry.Config{Writer: &w, Clock: uhr, MaxSpans: 0}, true, 1024},
		{"MaxSpans 1", telemetry.Config{Writer: &w, Clock: uhr, MaxSpans: 1}, true, 1},
		{"MaxSpans 65536", telemetry.Config{Writer: &w, Clock: uhr, MaxSpans: 65536}, true, 65536},
		{"MaxSpans 65537", telemetry.Config{Writer: &w, Clock: uhr, MaxSpans: 65537}, false, 0},
		{"gültig", telemetry.Config{Writer: &w, Clock: uhr, MaxSpans: 10, MinLevel: slog.LevelDebug}, true, 10},
		{"MinLevel über Error", telemetry.Config{Writer: &w, Clock: uhr, MinLevel: slog.LevelError + 1}, true, 1024},
	}
	if len(faelle) != 9 {
		t.Fatalf("%d Fälle, erwartet 9", len(faelle))
	}
	for _, f := range faelle {
		a, err := telemetry.New(f.cfg)
		if f.ok {
			if err != nil || a == nil || a.MaxSpans() != f.spans {
				t.Errorf("%s: Fehler %v, Adapter %v; erwartet gültig mit MaxSpans %d", f.name, err, a, f.spans)
			}
			continue
		}
		if a != nil || !errors.Is(err, model.ErrUngueltig) || err.Error() != model.ErrUngueltig.Error() {
			t.Errorf("%s: Adapter %v, Fehler %v; erwartet nil und model.ErrUngueltig", f.name, a, err)
		}
	}
}

// P17: ein Attribut zählt höchstens einen Verwurf, Schlüssel vor Wert; der erste angenommene
// Wert eines Schlüssels gilt.
func TestVerwurfRangfolge(t *testing.T) {
	leer := support.Str("session_id", "a b")
	faelle := []struct {
		name       string
		attrs      []support.Attr
		schluessel uint64
		wert       uint64
		gilt       []support.Attr
	}{
		{"unbekannter Schlüssel mit ungültigem Wert", []support.Attr{support.Str("timestamp", "a b")}, 1, 0, nil},
		{"doppelter Schlüssel, zweiter Wert ungültig", []support.Attr{support.Str("session_id", "eins"), leer}, 1, 0, []support.Attr{support.Str("session_id", "eins")}},
		{"doppelter Schlüssel, erster Wert ungültig", []support.Attr{leer, support.Str("session_id", "zwei")}, 0, 1, []support.Attr{support.Str("session_id", "zwei")}},
		{"doppelter Schlüssel, beide gültig", []support.Attr{support.Str("session_id", "eins"), support.Str("session_id", "zwei")}, 1, 0, []support.Attr{support.Str("session_id", "eins")}},
		{"fremder Schlüssel doppelt", []support.Attr{support.Str("x", "a"), support.Str("x", "b")}, 2, 0, nil},
	}
	for _, f := range faelle {
		a, _ := neu(t)
		a.Log(context.Background(), slog.LevelInfo, "m", f.attrs...)
		s := a.Snapshot()
		reihenFest(t, f.name, s)
		if s.Dropped[support.VerwurfSchluessel] != f.schluessel || s.Dropped[support.VerwurfWert] != f.wert {
			t.Errorf("%s: schluessel %d, wert %d; erwartet %d und %d", f.name, s.Dropped[support.VerwurfSchluessel], s.Dropped[support.VerwurfWert], f.schluessel, f.wert)
		}
		for g, n := range s.Dropped {
			if g != support.VerwurfSchluessel && g != support.VerwurfWert && n != 0 {
				t.Errorf("%s: Grund %s zählt %d", f.name, g, n)
			}
		}
		angenommen, _ := telemetry.Filtern("", f.attrs)
		if !reflect.DeepEqual(angenommen, f.gilt) {
			t.Errorf("%s: angenommen %+v, erwartet %+v", f.name, angenommen, f.gilt)
		}
	}
}

// Log schreibt keine Zeile und ändert nichts außer den Verwürfen; StartSpan ist ohne Wirkung.
func TestKernSchreibtNichts(t *testing.T) {
	a, puffer := neu(t)
	a.Log(context.Background(), slog.LevelError, "m", support.Str("mode", "record"))
	a.Log(context.TODO(), slog.LevelDebug, "m")
	var kein context.Context
	a.Log(kein, slog.LevelInfo, "m")
	ctx := context.WithValue(context.Background(), ctxSchluessel{}, 1)
	ctxK, sp := a.StartSpan(ctx, support.SpanMockQuery, support.Str("recorder.entity_set", "x"))
	sp.Set(support.Str("recorder.coverage_status", "voll"))
	sp.End()
	sp.End()
	ctxN, spN := a.StartSpan(kein, "unbekannt")
	spN.End()
	if puffer.Len() != 0 {
		t.Errorf("der Kern schrieb %d Bytes", puffer.Len())
	}
	if ctxK != ctx || ctxN == nil || sp == nil || spN == nil {
		t.Error("StartSpan liefert nicht den übergebenen ctx und einen Span")
	}
	s := a.Snapshot()
	reihenFest(t, "Kern", s)
	for g, n := range s.Dropped {
		if n != 0 {
			t.Errorf("Grund %s zählt %d", g, n)
		}
	}
	if len(s.Spans) != 0 || s.Durations[support.SpanMockQuery].Count != 0 {
		t.Errorf("StartSpan hat gewirkt: %+v", s)
	}
}

type ctxSchluessel struct{}

// P15: CountError verliert keinen Zuwachs, in beiden Zweigen; 64 Goroutinen mit je 500 Aufrufen,
// Codes reihum über die 33 der Tabelle und 4 außerhalb.
func TestCountErrorGleichzeitig(t *testing.T) {
	const goroutinen, aufrufe = 64, 500
	bekannt := model.AlleCodes()
	codes := append(slices.Clone(bekannt), 0, -1, model.CodeE9001+1, math.MaxInt)
	a, _ := neu(t)
	var wg sync.WaitGroup
	for g := range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range aufrufe {
				a.CountError(codes[(g*aufrufe+i)%len(codes)])
			}
		}()
	}
	wg.Wait()
	erwartet := map[model.Code]uint64{}
	unbekannt := uint64(0)
	for g := range goroutinen {
		for i := range aufrufe {
			c := codes[(g*aufrufe+i)%len(codes)]
			if slices.Contains(bekannt, c) {
				erwartet[c]++
			} else {
				unbekannt++
			}
		}
	}
	s := a.Snapshot()
	reihenFest(t, "CountError gleichzeitig", s)
	if len(erwartet) != 33 || len(codes) != 37 || unbekannt == 0 || summe(s.Errors)+s.Dropped[support.VerwurfFehlercode] != goroutinen*aufrufe {
		t.Errorf("Summe %d plus %d, erwartet %d bei %d Codes", summe(s.Errors), s.Dropped[support.VerwurfFehlercode], goroutinen*aufrufe, len(erwartet))
	}
	for c, n := range erwartet {
		if s.Errors[c] != n {
			t.Errorf("%s: %d, erwartet %d", c, s.Errors[c], n)
		}
	}
	if s.Dropped[support.VerwurfFehlercode] != unbekannt {
		t.Errorf("fehlercode %d, erwartet %d", s.Dropped[support.VerwurfFehlercode], unbekannt)
	}
}

// P15: Log, CountError und Snapshot gleichzeitig. 16 Schreiber starten zugleich (je 2000 Runden mit
// einem bekannten Code, einem unbekannten und einem Log mit zwei Verwürfen), 16 Leser lesen je 400
// Stände und prüfen, dass kein Zähler fällt; am Ende stehen die genauen Summen.
func TestAdapterGleichzeitig(t *testing.T) {
	const schreiber, runden, leser, staende = 16, 2000, 16, 400
	a, _ := neu(t)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range schreiber {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range runden {
				a.CountError(model.CodeE2001)
				a.CountError(0)
				a.Log(context.Background(), slog.LevelInfo, "m", support.Str("x", "y"), support.Str("mode", "a b"))
			}
		}()
	}
	fallend := make(chan string, leser*staende)
	for range leser {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			var e, k, w, f uint64
			for range staende {
				s := a.Snapshot()
				ne, nk, nw, nf := s.Errors[model.CodeE2001], s.Dropped[support.VerwurfSchluessel], s.Dropped[support.VerwurfWert], s.Dropped[support.VerwurfFehlercode]
				if ne < e || nk < k || nw < w || nf < f {
					fallend <- "ein Zähler fiel zwischen zwei Ständen"
				}
				e, k, w, f = ne, nk, nw, nf
			}
		}()
	}
	close(start)
	wg.Wait()
	close(fallend)
	for m := range fallend {
		t.Error(m)
	}
	s := a.Snapshot()
	reihenFest(t, "Adapter gleichzeitig", s)
	const n = schreiber * runden
	if s.Errors[model.CodeE2001] != n || s.Dropped[support.VerwurfFehlercode] != n ||
		s.Dropped[support.VerwurfSchluessel] != n || s.Dropped[support.VerwurfWert] != n {
		t.Errorf("E2001 %d, fehlercode %d, schluessel %d, wert %d; erwartet je %d", s.Errors[model.CodeE2001],
			s.Dropped[support.VerwurfFehlercode], s.Dropped[support.VerwurfSchluessel], s.Dropped[support.VerwurfWert], n)
	}
}

// P15: Snapshot liefert einen Stand zu einem Zeitpunkt (SPEC-230 Nr. 9). Die Menge sind alle
// Zähler, die ein Aufruf dieses Teils ändert: die 33 Codes von Errors in der Reihenfolge der
// Tabelle, dann fehlercode, schluessel und wert von Dropped (36 Zähler, 35 Fenster der Sperre
// zwischen je zwei Nachbarn). Schreiber erhöhen in jeder Runde diese Zähler in genau dieser
// Reihenfolge: alle 33 Codes nacheinander, CountError(0) und ein Log mit zwei Verwürfen
// (schluessel, wert). Leser lesen Stände, solange die Schreiber laufen. Das Orakel steht im
// Test, nicht im Adapter: in jedem Stand gilt für jedes Paar Nachbarn k[i+1] <= k[i] <= k[i+1] +
// Zahl der Schreiber, für das Paar schluessel und wert die Gleichheit (ein Log zählt beide unter
// einer Sperre). Liest ein Stand zwei Nachbarn zu verschiedenen Zeiten, bricht das Paar eine der
// beiden Seiten; ein Log, das seine Verwürfe in getrennten Abschnitten zählt, bricht die
// Gleichheit. Der Detektor meldet das nicht, weil jeder Zugriff unter einer Sperre liegt. Durations und Spans sowie die Gründe
// senke, pflicht, spanname und zustand ändert kein Aufruf dieses Teils, ein Fenster um sie sieht
// der Test nicht (Negativ N-1 im Plan).
func TestSnapshotZeitpunkt(t *testing.T) {
	const schreiber, leser, runden = 8, 8, 1000
	codes := model.AlleCodes()
	if len(codes) != 33 {
		t.Fatalf("%d Codes, erwartet 33", len(codes))
	}
	kette := func(s support.Snapshot) []uint64 {
		k := make([]uint64, 0, len(codes)+3)
		for _, c := range codes {
			k = append(k, s.Errors[c])
		}
		return append(k, s.Dropped[support.VerwurfFehlercode], s.Dropped[support.VerwurfSchluessel], s.Dropped[support.VerwurfWert])
	}
	namen := make([]string, 0, len(codes)+3)
	for _, c := range codes {
		namen = append(namen, c.String())
	}
	namen = append(namen, "fehlercode", "schluessel", "wert")
	laenge := len(namen)
	fenster := laenge - 1
	if laenge != 36 || fenster != 35 {
		t.Fatalf("%d Zähler und %d Fenster, erwartet 36 und 35", laenge, fenster)
	}

	// ein Log zählt schluessel und wert unter einer Sperre: dieses Paar ist gleich, die übrigen liegen
	// um höchstens die Zahl der Schreiber auseinander
	luft := func(i int) uint64 {
		if i == fenster-1 {
			return 0
		}
		return schreiber
	}

	a, _ := neu(t)
	start := make(chan struct{})
	var schreibend, lesend sync.WaitGroup
	var fertig atomic.Bool
	var gelesen atomic.Int64
	vorn := make([]atomic.Int64, fenster) // k[i+1] > k[i]: der hintere Zähler zu spät gelesen
	hinten := make([]atomic.Int64, fenster)
	for range schreiber {
		schreibend.Add(1)
		go func() {
			defer schreibend.Done()
			<-start
			for range runden {
				for _, c := range codes {
					a.CountError(c)
				}
				a.CountError(0)
				a.Log(context.Background(), slog.LevelInfo, "m", support.Str("x", "y"), support.Str("mode", "a b"))
			}
		}()
	}
	for range leser {
		lesend.Add(1)
		go func() {
			defer lesend.Done()
			<-start
			for {
				k := kette(a.Snapshot())
				for i := range fenster {
					if k[i+1] > k[i] {
						vorn[i].Add(1)
					}
					if k[i] > k[i+1]+luft(i) {
						hinten[i].Add(1)
					}
				}
				gelesen.Add(1)
				if fertig.Load() {
					return
				}
			}
		}()
	}
	close(start)
	schreibend.Wait()
	fertig.Store(true)
	lesend.Wait()
	for i := range fenster {
		if n := vorn[i].Load(); n != 0 {
			t.Errorf("Fenster %s | %s: %d Stände mit %s über %s: der hintere Zähler später gelesen", namen[i], namen[i+1], n, namen[i+1], namen[i])
		}
		if n := hinten[i].Load(); n != 0 {
			t.Errorf("Fenster %s | %s: %d Stände mit %s mehr als %d über %s: der vordere Zähler später gelesen", namen[i], namen[i+1], n, namen[i], luft(i), namen[i+1])
		}
	}
	if gelesen.Load() < leser {
		t.Errorf("%d Stände gelesen, erwartet mindestens %d", gelesen.Load(), leser)
	}
	const n = schreiber * runden
	for i, v := range kette(a.Snapshot()) {
		if v != n {
			t.Errorf("%s zählt %d, erwartet %d", namen[i], v, n)
		}
	}
}

// P17: Log wendet MinLevel nicht an: bei jeder Stufe des Aufrufs und jedem MinLevel zählt ein
// Verwurf (4 Stufen × 4 MinLevel).
func TestLogOhneMinLevel(t *testing.T) {
	stufen := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	minima := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelError, slog.LevelError + 1}
	for _, mindest := range minima {
		for _, stufe := range stufen {
			var w bytes.Buffer
			a, err := telemetry.New(telemetry.Config{Writer: &w, Clock: supporttest.NewFestUhr(time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)), MinLevel: mindest})
			if err != nil {
				t.Fatal(err)
			}
			a.Log(context.Background(), stufe, "m", support.Str("x", "y"))
			if n := a.Snapshot().Dropped[support.VerwurfSchluessel]; n != 1 {
				t.Errorf("MinLevel %v, Stufe %v: schluessel %d, erwartet 1", mindest, stufe, n)
			}
		}
	}
}

// P10: jeder Aufruf von New liefert einen eigenen Adapter; dieselbe Config gilt mehrfach.
func TestNewUnabhaengig(t *testing.T) {
	var w bytes.Buffer
	cfg := telemetry.Config{Writer: &w, Clock: supporttest.NewFestUhr(time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC))}
	a, err := telemetry.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := telemetry.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a.CountError(model.CodeE2001)
	a.CountError(0)
	a.Log(context.Background(), slog.LevelInfo, "m", support.Str("x", "y"), support.Str("mode", "a b"))
	sa, sb := a.Snapshot(), b.Snapshot()
	if sa.Errors[model.CodeE2001] != 1 || sa.Dropped[support.VerwurfFehlercode] != 1 || sa.Dropped[support.VerwurfSchluessel] != 1 || sa.Dropped[support.VerwurfWert] != 1 {
		t.Errorf("A zeigt die Zuwächse nicht: %+v", sa)
	}
	if summe(sb.Errors) != 0 {
		t.Errorf("B zählt Fehler von A: %+v", sb.Errors)
	}
	for g, n := range sb.Dropped {
		if n != 0 {
			t.Errorf("B zählt den Grund %s von A: %d", g, n)
		}
	}
	b.CountError(model.CodeE2101)
	if a.Snapshot().Errors[model.CodeE2101] != 0 {
		t.Error("A zählt den Fehler von B")
	}
}

// P14: der Adapter importiert nur model, ports/support und die Standardbibliothek. a-check prüft
// die Richtung der Schichten und erlaubt dem Adapter auch ports/outbound; dieser Test prüft die
// Menge der Importe jeder Datei des Verzeichnisses, auch der Dateien _test.go (a-check scannt sie
// ebenso): in den Dateien _test.go sind zusätzlich der Adapter selbst und supporttest erlaubt.
func TestAdapterImporte(t *testing.T) {
	const modul = "github.com/pt9912/odata-recorder/src/hexagon/"
	erlaubt := []string{modul + "model", modul + "ports/support"}
	nurTests := []string{modul + "ports/support/supporttest", "github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"}
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) < 2 {
		t.Fatalf("Dateien des Pakets: %v, %v", dateien, err)
	}
	gesehen := 0
	for _, d := range dateien {
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			pfad, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			gesehen++
			stdlib := !strings.Contains(strings.SplitN(pfad, "/", 2)[0], ".")
			if strings.HasSuffix(d, "_test.go") && slices.Contains(nurTests, pfad) {
				continue
			}
			if !stdlib && !slices.Contains(erlaubt, pfad) {
				t.Errorf("%s importiert %s", d, pfad)
			}
		}
	}
	if gesehen == 0 {
		t.Error("keine Importe gelesen")
	}
}

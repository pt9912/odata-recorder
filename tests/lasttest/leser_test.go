package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
)

const (
	jsonSpan = `"span"`
	jsonName = `"odata.gateway.request"`
	jsonID   = `"spn-1"`
)

// jsonZeile baut ein Objekt aus den fünf Feldern in Rohform des JSON; ein leerer Wert lässt das Feld weg.
func jsonZeile(ereignis, msg, spanID, parent, dauer string) string {
	var felder []string
	for _, f := range [][2]string{{"ereignis", ereignis}, {"msg", msg}, {"span_id", spanID}, {"parent_id", parent}, {"duration_ms", dauer}} {
		if f[1] != "" {
			felder = append(felder, `"`+f[0]+`":`+f[1])
		}
	}
	return "{" + strings.Join(felder, ",") + "}"
}

// gepolstert ist eine gültige Zeile eines Gesamtspans mit der Kennung id und genau n Bytes Länge ohne Zeilenende.
func gepolstert(id string, n int) string {
	kopf := `{"ereignis":"span","msg":"odata.gateway.request","span_id":"` + id + `","duration_ms":1.000,"z":"`
	return kopf + strings.Repeat("z", n-len(kopf)-2) + `"}`
}

// leseAlles liest r und liefert die Span-Zeilen in der Folge, die Zahl der fremden Zeilen und den Fehler.
func leseAlles(r io.Reader) ([]spanZeile, int, error) {
	var zeilen []spanZeile
	fremd, err := lies(r, func(z spanZeile) { zeilen = append(zeilen, z) })
	return zeilen, fremd, err
}

// zeitfolge ist die Uhr des Adapters: jede Lesung liefert die nächste Zeit der Liste.
type zeitfolge struct {
	mu     sync.Mutex
	zeiten []time.Time
}

func (z *zeitfolge) Now() time.Time {
	z.mu.Lock()
	defer z.mu.Unlock()
	t := z.zeiten[0]
	z.zeiten = z.zeiten[1:]
	return t
}

// TestSpanZeilen: Menge der Formen: 10 gültige Zeilen (die 3 des Plans, eine mit Leerraum vor dem Objekt, drei Gesamtspans ohne, mit null und mit
// einer Zahl als recorder.mode, die als Zeile gelesen werden und einen leeren Modus tragen, und je eine mit leerer parent_id, leerer span_id und
// leerem msg) und 47 fremde (die 13 Formen des Plans, eine ohne ereignis, dazu 18 weitere: Feld falscher Art oder null je Pflichtfeld und parent_id, JSON ohne Objekt,
// Dauer ohne Punkt mit drei Ziffern, mit Exponent, ohne Ziffer vor dem Punkt, mit Buchstaben, mit Vorzeichen, über int64; dazu 14 Formen nahe
// span: der Großbuchstabe je Stelle, die Anfänge, ein angehängtes und ein vorangestelltes Zeichen, und 01.000); Dauern als ganze Mikrosekunden
// aus dem Text (0.000 bis 9223372036854775.807, eins mehr ist fremd); die Zeilen des echten Adapters (telemetry.New mit vorgegebener Uhr,
// Dauer 12,345 ms) stimmen mit Snapshot().Spans überein.
func TestSpanZeilen(t *testing.T) {
	wache(t, 30*time.Second)
	gueltig := []struct {
		zeile string
		want  spanZeile
	}{
		{`{"timestamp":"t","level":"INFO","msg":"odata.gateway.request","ereignis":"span","span_id":"spn-1","duration_ms":12.345,"recorder.mode":"replay"}`,
			spanZeile{Name: "odata.gateway.request", SpanID: "spn-1", Modus: "replay", Mikros: 12345}},
		{`{"msg":"odata.replay.match","ereignis":"span","span_id":"spn-2","parent_id":"spn-1","duration_ms":0.500}`,
			spanZeile{Name: "odata.replay.match", SpanID: "spn-2", ParentID: "spn-1", Mikros: 500}},
		{`{"ereignis":"span","msg":"x","span_id":"s","duration_ms":7.000}`, spanZeile{Name: "x", SpanID: "s", Mikros: 7000}},
		{" \t" + `{"ereignis":"span","msg":"x","span_id":"s","duration_ms":7.000}`, spanZeile{Name: "x", SpanID: "s", Mikros: 7000}},
		{`{"ereignis":"span","msg":"odata.gateway.request","span_id":"s","duration_ms":7.000}`, spanZeile{Name: "odata.gateway.request", SpanID: "s", Mikros: 7000}},
		{`{"ereignis":"span","msg":"odata.gateway.request","span_id":"s","duration_ms":7.000,"recorder.mode":null}`,
			spanZeile{Name: "odata.gateway.request", SpanID: "s", Mikros: 7000}},
		{`{"ereignis":"span","msg":"odata.gateway.request","span_id":"s","duration_ms":7.000,"recorder.mode":1}`,
			spanZeile{Name: "odata.gateway.request", SpanID: "s", Mikros: 7000}},
		{`{"ereignis":"span","msg":"odata.replay.match","span_id":"s","parent_id":"","duration_ms":7.000}`, spanZeile{Name: "odata.replay.match", SpanID: "s", Mikros: 7000}},
		{`{"ereignis":"span","msg":"x","span_id":"","duration_ms":7.000}`, spanZeile{Name: "x", Mikros: 7000}},
		{`{"ereignis":"span","msg":"","span_id":"s","duration_ms":7.000}`, spanZeile{SpanID: "s", Mikros: 7000}},
	}
	if len(gueltig) != 10 {
		t.Fatalf("%d gültige Formen, erwartet 10", len(gueltig))
	}
	for _, g := range gueltig {
		z, fremd, err := leseAlles(strings.NewReader(g.zeile + "\n"))
		if err != nil || fremd != 0 || len(z) != 1 || z[0] != g.want {
			t.Errorf("%s: Zeilen %+v, fremd %d, Fehler %v; erwartet %+v", g.zeile, z, fremd, err, g.want)
		}
	}
	plan := []string{
		"kein JSON", "[1,2]", jsonZeile(`"log"`, jsonName, jsonID, "", "1.000"), jsonZeile(jsonSpan, "", jsonID, "", "1.000"),
		jsonZeile(jsonSpan, jsonName, "", "", "1.000"), jsonZeile(jsonSpan, jsonName, jsonID, "", ""), jsonZeile(jsonSpan, jsonName, jsonID, "", `"1.000"`),
		jsonZeile(jsonSpan, jsonName, jsonID, "", "1"), jsonZeile(jsonSpan, jsonName, jsonID, "", "1.00"),
		jsonZeile(jsonSpan, jsonName, jsonID, "", "1.0000"), jsonZeile(jsonSpan, jsonName, jsonID, "", "-1.000"),
		jsonZeile(jsonSpan, jsonName, jsonID, "", "1e3"), "",
	}
	weitere := []string{
		"null", "{}", jsonZeile("1", jsonName, jsonID, "", "1.000"), jsonZeile(jsonSpan, "1", jsonID, "", "1.000"),
		jsonZeile(jsonSpan, jsonName, "1", "", "1.000"), jsonZeile(jsonSpan, jsonName, jsonID, "1", "1.000"),
		jsonZeile(jsonSpan, jsonName, jsonID, "null", "1.000"), jsonZeile(jsonSpan, jsonName, jsonID, "", "null"),
		jsonZeile("null", jsonName, jsonID, "", "1.000"), jsonZeile(jsonSpan, "null", jsonID, "", "1.000"),
		jsonZeile(jsonSpan, jsonName, "null", "", "1.000"), jsonZeile(jsonSpan, jsonName, jsonID, "", "123"),
		jsonZeile(jsonSpan, jsonName, jsonID, "", "1.5e1"),
		jsonZeile(jsonSpan, jsonName, jsonID, "", ".345"), jsonZeile(jsonSpan, jsonName, jsonID, "", "1.00a"),
		jsonZeile(jsonSpan, jsonName, jsonID, "", "+1.000"), jsonZeile(jsonSpan, jsonName, jsonID, "", "9223372036854775.808"),
		jsonZeile(jsonSpan, jsonName, jsonID, "", "99999999999999999999.000"),
	}
	nbsp := string([]byte{0xc2, 0xa0})
	ereignisse := []string{
		`"Span"`, `"sPan"`, `"spAn"`, `"spaN"`, `"s"`, `"sp"`, `"spa"`, `""`,
		`"span "`, `"span\t"`, `"spans"`, `"span` + nbsp + `"`, `" span"`, `"sspan"`,
	}
	for _, e := range ereignisse {
		weitere = append(weitere, jsonZeile(e, jsonName, jsonID, "", "1.000"))
	}
	// jedes Pflichtfeld einzeln fehlend: msg, span_id und duration_ms stehen im Plan, ereignis hier.
	weitere = append(weitere, jsonZeile(jsonSpan, jsonName, jsonID, "", "01.000"), jsonZeile("", jsonName, jsonID, "", "1.000"))
	if len(plan) != 13 || len(ereignisse) != 14 || len(weitere) != 34 {
		t.Fatalf("%d, %d und %d fremde Formen, erwartet 13, 14 und 34 weitere", len(plan), len(ereignisse), len(weitere))
	}
	for _, f := range append(plan, weitere...) {
		z, fremd, err := leseAlles(strings.NewReader(f + "\n"))
		if err != nil || fremd != 1 || len(z) != 0 {
			t.Errorf("%q: Zeilen %+v, fremd %d, Fehler %v; erwartet keine Zeile und fremd 1", f, z, fremd, err)
		}
	}
	dauern := map[string]int64{"0.000": 0, "0.001": 1, "0.999": 999, "1.000": 1000, "123456.789": 123456789, "9223372036854775.807": 9223372036854775807}
	if len(dauern) != 6 {
		t.Fatalf("%d Dauern, erwartet 6", len(dauern))
	}
	for text, mikros := range dauern {
		z, fremd, err := leseAlles(strings.NewReader(jsonZeile(jsonSpan, jsonName, jsonID, "", text) + "\n"))
		if err != nil || fremd != 0 || len(z) != 1 || z[0].Mikros != mikros {
			t.Errorf("Dauer %s: Zeilen %+v, fremd %d, Fehler %v; erwartet %d µs", text, z, fremd, err, mikros)
		}
	}
	dauerKlassen(t)
	adapterZeilen(t)
}

// dauerLesen liest die Dauer `text` aus der Zeile eines Spans; ok sagt, dass die Zeile eine Span-Zeile ist.
func dauerLesen(text string) (mikros int64, ok bool) {
	z, fremd, err := leseAlles(strings.NewReader(jsonZeile(jsonSpan, jsonName, jsonID, "", text) + "\n"))
	if err != nil || fremd != 0 || len(z) != 1 {
		return 0, false
	}
	return z[0].Mikros, true
}

// dauerKlassen führt die Menge der Dauern aus dem Wortlaut (Ziffern, Punkt, genau drei Nachkommastellen) als Ziffer × Stelle (49 Elemente: die
// Ziffern 0 bis 9 an den vier Stellen `11.111` hinter der ersten, die Ziffern 1 bis 9 an der ersten) und als Länge des Teils vor dem Punkt (1 bis
// 20 Ziffern: bis 15 eine Dauer, darüber der Überlauf von int64 und eine fremde Zeile); die Erwartung bildet die Stellenwerte, nicht den Leser.
func dauerKlassen(t *testing.T) {
	t.Helper()
	gewicht := []int64{10000, 1000, 100, 10, 1}
	n := 0
	for stelle, w := range gewicht {
		for ziffer := int64(0); ziffer <= 9; ziffer++ {
			if stelle == 0 && ziffer == 0 {
				continue
			}
			ziffern := []byte("11111")
			ziffern[stelle] = byte('0' + ziffer)
			text := string(ziffern[:2]) + "." + string(ziffern[2:])
			if got, ok := dauerLesen(text); !ok || got != 11111+(ziffer-1)*w {
				t.Errorf("Dauer %s: %d, %v; erwartet %d", text, got, ok, 11111+(ziffer-1)*w)
			}
			n++
		}
	}
	if n != 49 {
		t.Errorf("%d Elemente Ziffer × Stelle, erwartet 49", n)
	}
	potenz := int64(10000)
	for laenge := 1; laenge <= 20; laenge++ {
		text := strings.Repeat("9", laenge) + ".999"
		got, ok := dauerLesen(text)
		if laenge <= 15 && (!ok || got != potenz-1) || laenge > 15 && ok {
			t.Errorf("Dauer %s: %d, %v; erwartet %d bis 15 Ziffern, darüber eine fremde Zeile", text, got, ok, potenz-1)
		}
		if laenge < 15 {
			potenz *= 10
		}
	}
}

// adapterZeilen schreibt mit telemetry.New zwei Zeilen (Kind und Gesamtspan, Dauer des Gesamtspans 12,345 ms) und liest sie.
func adapterZeilen(t *testing.T) {
	t.Helper()
	t0 := time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC)
	var puffer bytes.Buffer
	a, err := telemetry.New(telemetry.Config{Writer: &puffer, Clock: &zeitfolge{zeiten: []time.Time{
		t0, t0.Add(time.Millisecond), t0.Add(2 * time.Millisecond), t0.Add(12345 * time.Microsecond)}}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, gesamt := a.StartSpan(context.Background(), spanGesamt)
	_, kind := a.StartSpan(ctx, spanReplay)
	kind.End()
	gesamt.End()
	spans := a.Snapshot().Spans
	zeilen, fremd, err := leseAlles(&puffer)
	if err != nil || fremd != 0 || len(zeilen) != 2 || len(spans) != 2 {
		t.Fatalf("Zeilen %+v, fremd %d, Fehler %v, Snapshot %d Spans; erwartet 2, 0 und 2", zeilen, fremd, err, len(spans))
	}
	for i, z := range zeilen {
		if z.Name != string(spans[i].Name) || z.SpanID != spans[i].ID || z.ParentID != spans[i].ParentID || z.Mikros != spans[i].Duration.Microseconds() {
			t.Errorf("Zeile %d %+v gleich Snapshot %+v erwartet", i, z, spans[i])
		}
	}
	if zeilen[0].ParentID != zeilen[1].SpanID || zeilen[1].Mikros != 12345 || zeilen[1].ParentID != "" || zeilen[0].Mikros != time.Millisecond.Microseconds() {
		t.Errorf("Kind %+v, Gesamtspan %+v: erwartet parent_id des Kindes gleich span_id des Gesamtspans, 12345 µs und kein parent_id am Gesamtspan", zeilen[0], zeilen[1])
	}
}

// pipeLauf schreibt in eine Pipe zuerst eine Zeile von 3 MiB, dann mib MiB in Zeilen zu 200 Byte, jede zwanzigste eine gültige Zeile, und
// verlangt, dass der Schreiber nach höchstens einer Sekunde fertig ist, alle gültigen Zeilen vom Sammler eingeordnet und alle anderen als
// fremd gezählt sind.
func pipeLauf(t *testing.T, mib int) {
	t.Helper()
	pr, pw := io.Pipe()
	schreiberFertig := make(chan struct{})
	go func() {
		defer close(schreiberFertig)
		_, _ = pw.Write([]byte(strings.Repeat("x", 3<<20) + "\n"))
		fremdZeile, spanZeilen := strings.Repeat("x", 199)+"\n", gepolstert("spn-1", 199)+"\n"
		block := strings.Repeat(strings.Repeat(fremdZeile, 19)+spanZeilen, 16)
		for n := mib << 20 / 200; n >= 320; n -= 320 {
			_, _ = pw.Write([]byte(block))
		}
		_ = pw.Close()
	}()
	ergebnis := make(chan [2]int, 1)
	go func() {
		s := neuerSammler("replay")
		fremd, _ := lies(pr, s.nimm)
		ergebnis <- [2]int{s.anzahl(), fremd}
	}()
	warteAuf(t, schreiberFertig, time.Second)
	zeilen := mib << 20 / 200 / 320 * 320
	if got := warteAuf(t, ergebnis, time.Second); got != [2]int{zeilen / 20, 1 + zeilen - zeilen/20} {
		t.Errorf("%d MiB: Spans %d, fremd %d; erwartet %d und %d", mib, got[0], got[1], zeilen/20, 1+zeilen-zeilen/20)
	}
}

// TestLeserHaeltNieAn (Wirkungs-Fänger, 20 Läufe): der Leser entleert die Pipe ohne Unterbrechung, auch hinter einer Zeile über 1 MiB und
// bei einer Span-Zeile in jeder zwanzigsten; ein Schreiber mit 20 MiB endet nach höchstens einer Sekunde. Eine Zeile von 1 MiB wird gelesen,
// eine von 1 MiB + 1 B zwischen drei Span-Zeilen verworfen und als fremd gezählt; der Speicher wächst nicht mit der Länge (64 MiB in einer
// Zeile allozieren unter 16 MiB); eine letzte Zeile ohne Zeilenende ist eine Zeile; ein Fehler des Stroms endet das Lesen.
func TestLeserHaeltNieAn(t *testing.T) {
	wache(t, 60*time.Second)
	pipeLauf(t, 20)
	for i := 1; i < 20; i++ {
		pipeLauf(t, 1)
	}
	eingabe := gepolstert("s1", 150) + "\n" + gepolstert("l1", maxZeile) + "\n" + gepolstert("s2", 150) + "\n" +
		gepolstert("l2", maxZeile+1) + "\n" + gepolstert("s3", 150)
	zeilen, fremd, err := leseAlles(strings.NewReader(eingabe))
	var ids []string
	for _, z := range zeilen {
		ids = append(ids, z.SpanID)
	}
	if err != nil || fremd != 1 || strings.Join(ids, " ") != "s1 l1 s2 s3" {
		t.Errorf("Zeilen %v, fremd %d, Fehler %v; erwartet s1 l1 s2 s3 und fremd 1 (die Zeile von 1 MiB + 1 B)", ids, fremd, err)
	}
	var grenzen strings.Builder
	var erwartet []string
	for _, n := range []int{65535, 65536, 65537, 131071, 131072, 131073, maxZeile - 1} {
		erwartet = append(erwartet, "c"+strconv.Itoa(n))
		grenzen.WriteString(gepolstert(erwartet[len(erwartet)-1], n) + "\n")
	}
	zeilen, fremd, err = leseAlles(strings.NewReader(grenzen.String()))
	ids = nil
	for _, z := range zeilen {
		ids = append(ids, z.SpanID)
	}
	if err != nil || fremd != 0 || !slices.Equal(ids, erwartet) {
		t.Errorf("Zeilen an den Grenzen des Puffers (64 KiB, 128 KiB, ±1): %v, fremd %d, Fehler %v; erwartet %v", ids, fremd, err, erwartet)
	}
	var vorher, nachher runtime.MemStats
	runtime.ReadMemStats(&vorher)
	fremd, err = lies(&endlos{rest: 64 << 20}, func(spanZeile) {})
	runtime.ReadMemStats(&nachher)
	if err != nil || fremd != 1 || nachher.TotalAlloc-vorher.TotalAlloc > 16<<20 {
		t.Errorf("64 MiB in einer Zeile: fremd %d, Fehler %v, %d Byte alloziert; erwartet fremd 1 und unter 16 MiB", fremd, err, nachher.TotalAlloc-vorher.TotalAlloc)
	}
	boom := errors.New("Strom gerissen")
	zeilen, fremd, err = leseAlles(io.MultiReader(strings.NewReader("x\nabc"), &fehlerLeser{boom}))
	if !errors.Is(err, boom) || fremd != 2 || len(zeilen) != 0 {
		t.Errorf("Fehler des Stroms: Zeilen %v, fremd %d, Fehler %v; erwartet fremd 2 und den Fehler des Stroms", zeilen, fremd, err)
	}
}

// endlos liefert rest Bytes `x` ohne Zeilenende.
type endlos struct{ rest int }

func (e *endlos) Read(p []byte) (int, error) {
	if e.rest == 0 {
		return 0, io.EOF
	}
	n := min(len(p), e.rest)
	for i := range p[:n] {
		p[i] = 'x'
	}
	e.rest -= n
	return n, nil
}

type fehlerLeser struct{ err error }

func (f *fehlerLeser) Read([]byte) (int, error) { return 0, f.err }

package telemetry_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// S07 (M5): die Dauer ist Ende minus Anfang, bei Ende vor Anfang 0; erster Eimer mit Dauer <= Grenze.
// 26 Läufe: 0, jede der 11 Grenzen und jede um 1 ns darüber, 1 Stunde, Ende 1 ns und 1 s vor dem Anfang.
func TestEimer(t *testing.T) {
	tab := spanTabelle(t)
	grenzen := eimerGrenzen(t)
	type fall struct {
		name  string
		dauer time.Duration
		eimer int
	}
	faelle := []fall{{"0", 0, 0}}
	for i, g := range grenzen {
		faelle = append(faelle, fall{"Grenze " + g.String(), g, i}, fall{"Grenze " + g.String() + " + 1 ns", g + 1, i + 1})
	}
	faelle = append(faelle, fall{"1 Stunde", time.Hour, len(grenzen)}, fall{"1 ns vor dem Anfang", -1, 0}, fall{"1 s vor dem Anfang", -time.Second, 0})
	if len(faelle) != 26 {
		t.Fatalf("%d Fälle, erwartet 26", len(faelle))
	}
	start := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	for i, f := range faelle {
		z := tab[i%len(tab)]
		a := neuMitMax(t, io.Discard, &folgeUhr{zeiten: []time.Time{start, start.Add(f.dauer)}}, slog.LevelInfo, 0)
		startEnde(a, z.name)
		s := a.Snapshot()
		v := s.Durations[support.SpanName(z.name)]
		erwartet := max(f.dauer, 0)
		counts := make([]uint64, len(grenzen)+1)
		counts[f.eimer] = 1
		if v.Count != 1 || v.Sum != erwartet || !slices.Equal(v.Counts, counts) || len(s.Spans) != 1 || s.Spans[0].Duration != erwartet {
			t.Errorf("%s (%s): Verteilung %+v, Eintrag %+v; erwartet Eimer %d und Dauer %v", f.name, z.name, v, s.Spans, f.eimer, erwartet)
		}
	}
}

// S08: Count ist die Summe von Counts, Sum die Summe der Dauern, je Name getrennt; die Zahl der Einträge
// eines Namens im Ring ist höchstens Count. 5 Namen × 3 Enden mit festen Dauern.
func TestVerteilung(t *testing.T) {
	tab := spanTabelle(t)
	dauern := []time.Duration{3 * time.Millisecond, 40 * time.Millisecond, 700 * time.Millisecond}
	start := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	var zeiten []time.Time
	for range tab {
		for _, d := range dauern {
			zeiten = append(zeiten, start, start.Add(d))
		}
	}
	a := neuMitMax(t, io.Discard, &folgeUhr{zeiten: zeiten}, slog.LevelInfo, 0)
	count, sum := map[string]uint64{}, map[string]time.Duration{}
	for _, z := range tab {
		for _, d := range dauern {
			startEnde(a, z.name)
			count[z.name]++
			sum[z.name] += d
			s := a.Snapshot()
			for _, y := range tab {
				v, eintraege := s.Durations[support.SpanName(y.name)], 0
				var imRing uint64
				for _, r := range s.Spans {
					if string(r.Name) == y.name {
						imRing++
					}
				}
				for _, c := range v.Counts {
					eintraege += int(c)
				}
				if v.Count != count[y.name] || v.Sum != sum[y.name] || uint64(eintraege) != v.Count || imRing > v.Count {
					t.Errorf("nach %s: %s hat Count %d Sum %v Counts %d Ring %d; erwartet Count %d Sum %v", z.name, y.name, v.Count, v.Sum, eintraege, imRing, count[y.name], sum[y.name])
				}
			}
		}
	}
}

// S12 (M6): der Ring hält die zuletzt beendeten MaxSpans, den ältesten voran, mit den Attributen in
// der Reihenfolge der Tabelle und nie nil. 13 Läufe: MaxSpans 1 (Enden 0 bis 3), 3 (0, 2, 3, 4, 7),
// 0 als Default 1024 (1024, 1025) und 65536 (65536, 65537).
func TestRing(t *testing.T) {
	tab := spanTabelle(t)
	laeufe := []struct {
		max   int
		enden []int
	}{{1, []int{0, 1, 2, 3}}, {3, []int{0, 2, 3, 4, 7}}, {0, []int{1024, 1025}}, {65536, []int{65536, 65537}}}
	groesstes := 0
	for _, z := range tab {
		groesstes = max(groesstes, len(z.pflicht))
	}
	basis := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	gezaehlt := 0
	for _, l := range laeufe {
		for _, n := range l.enden {
			gezaehlt++
			wirksam := l.max
			if wirksam == 0 {
				wirksam = 1024
			}
			var zeiten []time.Time
			for i := range n {
				s := basis.Add(time.Duration(i) * time.Second)
				zeiten = append(zeiten, s, s.Add(time.Duration(i%7+1)*time.Millisecond))
			}
			a := neuMitMax(t, io.Discard, &folgeUhr{zeiten: zeiten}, slog.LevelError+1, l.max)
			if a.MaxSpans() != wirksam {
				t.Fatalf("MaxSpans %d hält %d, erwartet %d", l.max, a.MaxSpans(), wirksam)
			}
			count := map[string]uint64{}
			for i := range n {
				z := tab[i%len(tab)]
				startEnde(a, z.name, werteVon(z.pflicht, gueltigSpan)...)
				count[z.name]++
			}
			s := a.Snapshot()
			if len(s.Spans) != min(n, wirksam) || s.Spans == nil {
				t.Fatalf("MaxSpans %d, %d Enden: %d Einträge, erwartet %d", l.max, n, len(s.Spans), min(n, wirksam))
			}
			imRing := map[string]uint64{}
			for j, r := range s.Spans {
				i := n - len(s.Spans) + j
				z := tab[i%len(tab)]
				imRing[z.name]++
				if r.ID != orakelID(i+1) || r.ParentID != "" || string(r.Name) != z.name || !r.Start.Equal(zeiten[2*i]) ||
					r.Duration != time.Duration(i%7+1)*time.Millisecond || !attrsGleich(r.Attrs, werteVon(z.pflicht, gueltigSpan)) ||
					len(r.Attrs) > groesstes {
					t.Fatalf("MaxSpans %d, %d Enden, Platz %d: Eintrag %+v, erwartet ID %s von %s", l.max, n, j, r, orakelID(i+1), z.name)
				}
			}
			for _, z := range tab {
				v := s.Durations[support.SpanName(z.name)]
				if v.Count != count[z.name] || imRing[z.name] > v.Count || (n <= wirksam && imRing[z.name] != v.Count) {
					t.Errorf("MaxSpans %d, %d Enden, %s: Count %d, Ring %d, erwartet Count %d", l.max, n, z.name, v.Count, imRing[z.name], count[z.name])
				}
			}
		}
	}
	if gezaehlt != 13 {
		t.Errorf("%d Läufe, erwartet 13", gezaehlt)
	}
}

// S12: Spans und Attrs des Stands sind Kopien; Snapshot ändert nichts; Attrs eines Spans ohne Attribute
// ist leer und nicht nil.
func TestRingKopie(t *testing.T) {
	tab := spanTabelle(t)
	a := neuMitMax(t, io.Discard, festeUhr(), slog.LevelInfo, 3)
	for i := range 5 {
		startEnde(a, tab[i%len(tab)].name, werteVon(tab[i%len(tab)].pflicht, gueltigSpan)...)
	}
	grund := a.Snapshot().Kopie()
	erster := a.Snapshot()
	erster.Spans[0].Attrs[0] = support.Str("x", "y")
	erster.Spans[0].ID = "geändert"
	erster.Spans = erster.Spans[:1]
	zweiter, dritter := a.Snapshot(), a.Snapshot()
	if !reflect.DeepEqual(zweiter, grund) || !reflect.DeepEqual(zweiter, dritter) {
		t.Errorf("eine Änderung am Stand wirkte zurück oder Snapshot ändert: %+v gegen %+v", zweiter.Spans, grund.Spans)
	}
	startEnde(a, tab[0].name)
	if !reflect.DeepEqual(zweiter.Spans, grund.Spans) {
		t.Errorf("ein späteres End änderte einen früheren Stand")
	}
	letzter := a.Snapshot().Spans[2]
	if letzter.Attrs == nil || len(letzter.Attrs) != 0 {
		t.Errorf("Attrs eines Spans ohne Attribute: %#v", letzter.Attrs)
	}
	if s := neuMitMax(t, io.Discard, festeUhr(), slog.LevelInfo, 0).Snapshot(); s.Spans == nil || len(s.Spans) != 0 {
		t.Errorf("Spans eines frischen Adapters: %#v", s.Spans)
	}
}

// S13 (M10): jedes End schreibt eine Zeile der Stufe INFO; unter MinLevel über INFO entfällt sie, die
// Aufzeichnung bleibt; bei defekter Senke zählt senke je Zeile einmal und der Span verhält sich wie ohne
// Ausfall. 5 Spans × 9 Modi, je ein Span mit einem Kind.
func TestSpanModi(t *testing.T) {
	tab := spanTabelle(t)
	fehler := errors.New("defekt")
	modi := []struct {
		name    string
		mindest slog.Level
		antwort func(p []byte) (int, error)
		zeilen  int
		senke   uint64
	}{
		{"MinLevel -4", -4, nil, 2, 0}, {"MinLevel 0", 0, nil, 2, 0}, {"MinLevel 1", 1, nil, 0, 0},
		{"MinLevel 4", 4, nil, 0, 0}, {"MinLevel 8", 8, nil, 0, 0}, {"MinLevel 9", 9, nil, 0, 0},
		{"Fehler ohne Bytes", 0, func([]byte) (int, error) { return 0, fehler }, 2, 2},
		{"weniger Bytes ohne Fehler", 0, func(p []byte) (int, error) { return len(p) - 1, nil }, 2, 2},
		{"alle Bytes mit Fehler", 0, func(p []byte) (int, error) { return len(p), fehler }, 2, 2},
	}
	laeufe := 0
	for i, z := range tab {
		kind := tab[(i+1)%len(tab)]
		var grund *support.Snapshot
		for _, m := range modi {
			laeufe++
			u := neuUmgebung(t, m.mindest, 0)
			u.senke.antwort = m.antwort
			ctx, wurzel := u.a.StartSpan(context.Background(), support.SpanName(z.name), werteVon(z.pflicht, gueltigSpan)...)
			_, k := u.a.StartSpan(ctx, support.SpanName(kind.name), werteVon(kind.pflicht, gueltigSpan)...)
			k.End()
			wurzel.End()
			s := u.a.Snapshot()
			wo := z.name + " / " + m.name
			pruefeVektor(t, wo, s, map[support.Verwurf]uint64{support.VerwurfSenke: m.senke})
			zeilen := u.senke.zeilen()
			if len(zeilen) != m.zeilen {
				t.Fatalf("%s: %d Write, erwartet %d", wo, len(zeilen), m.zeilen)
			}
			for j, w := range zeilen {
				g := lies(t, w)
				if g.text["level"] != "INFO" || g.text["msg"] != []string{kind.name, z.name}[j] {
					t.Errorf("%s: Zeile %d: %s %s", wo, j, g.text["level"], g.text["msg"])
				}
			}
			if len(s.Spans) != 2 || s.Spans[0].ParentID != orakelID(1) || s.Spans[1].ParentID != "" || u.uhr.n.Load() != 4 {
				t.Errorf("%s: Einträge %+v, %d Ablesungen", wo, s.Spans, u.uhr.n.Load())
			}
			if grund == nil {
				grund = &s
			} else if !reflect.DeepEqual(s.Spans, grund.Spans) || !reflect.DeepEqual(s.Durations, grund.Durations) {
				t.Errorf("%s: Aufzeichnung weicht vom ersten Modus ab", wo)
			}
		}
	}
	if laeufe != 45 {
		t.Errorf("%d Läufe, erwartet 45", laeufe)
	}
}

// S21 (M9): die Span-Zeile trägt die Felder in der Reihenfolge von SPEC-231 Nr. 6, parent_id nur mit
// Eltern, duration_ms mit drei Nachkommastellen; sie geht in einem Write hinaus, endet auf einem
// Zeilenende, und der Adapter behält das Argument nicht. 6 Dauern × 5 Spans × Wurzel und Kind = 60 Läufe.
func TestSpanZeileInhalt(t *testing.T) {
	dauern := []struct {
		d    time.Duration
		text string
	}{{0, "0.000"}, {1, "0.000"}, {499999, "0.500"}, {1234567, "1.235"}, {5000000001, "5000.000"}, {time.Hour, "3600000.000"}}
	start := time.Date(2026, 11, 7, 11, 22, 33, 123456789, time.UTC)
	laeufe := 0
	for _, z := range spanTabelle(t) {
		for _, d := range dauern {
			for _, kind := range []bool{false, true} {
				laeufe++
				ende := start.Add(d.d)
				zeiten := []time.Time{start, ende}
				if kind {
					zeiten = []time.Time{start, start, ende, ende}
				}
				senke := &FakeSenke{}
				a := neuMitMax(t, senke, &folgeUhr{zeiten: zeiten}, slog.LevelInfo, 0)
				attrs := werteVon(z.pflicht, gueltigSpan)
				ctx, wurzel := a.StartSpan(context.Background(), support.SpanName(z.name), attrs...)
				erwartetFolge := []string{"timestamp", "level", "msg", "ereignis", "span_id"}
				spanID, elternID := orakelID(1), ""
				if kind {
					_, k := a.StartSpan(ctx, support.SpanName(z.name), attrs...)
					k.End()
					spanID, elternID = orakelID(2), orakelID(1)
					erwartetFolge = append(erwartetFolge, "parent_id")
				}
				erwartetFolge = append(erwartetFolge, "duration_ms")
				erwartetFolge = append(erwartetFolge, z.pflicht...)
				wurzel.End()
				zeilen := senke.zeilen()
				if len(zeilen) != 1+boolZahl(kind) {
					t.Fatalf("%s: %d Zeilen", z.name, len(zeilen))
				}
				for range 3 {
					startEnde(a, z.name)
				}
				for i, w := range zeilen {
					if !bytes.Equal(w, senke.verweise[i]) || bytes.IndexFunc(w[:len(w)-1], func(r rune) bool { return r < 0x20 }) >= 0 {
						t.Errorf("%s: der Adapter hat das Argument von Write %d geändert oder die Zeile hat ein Steuerzeichen", z.name, i)
					}
				}
				g := lies(t, zeilen[0])
				wo := z.name + " " + d.text
				if !slices.Equal(g.schluessel, erwartetFolge) {
					t.Fatalf("%s: Felder %v, erwartet %v", wo, g.schluessel, erwartetFolge)
				}
				if g.text["timestamp"] != ende.Format("2006-01-02T15:04:05.000000000Z") || g.text["level"] != "INFO" || g.text["msg"] != z.name ||
					g.text["ereignis"] != "span" || g.text["span_id"] != spanID || g.text["parent_id"] != elternID ||
					g.text["duration_ms"] != d.text || !g.zahl["duration_ms"] {
					t.Errorf("%s: Zeile %v", wo, g.text)
				}
				for _, at := range attrs {
					if g.text[at.Key] != erwarteterText(at) || g.zahl[at.Key] != (at.Art == support.ArtInt) {
						t.Errorf("%s: %s = %q (Zahl %v)", wo, at.Key, g.text[at.Key], g.zahl[at.Key])
					}
				}
			}
		}
	}
	if laeufe != 60 {
		t.Errorf("%d Läufe, erwartet 60", laeufe)
	}
}

func boolZahl(b bool) int {
	if b {
		return 1
	}
	return 0
}

// S21 (LH-QA-01.a): ein Leser findet den Anteil von odata.replay.match und odata.mock.query am Gesamtspan
// über span_id und parent_id.
func TestSpanZeileGesamtspan(t *testing.T) {
	t0 := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	ms := func(n int) time.Time { return t0.Add(time.Duration(n) * time.Millisecond) }
	senke := &FakeSenke{}
	uhr := &folgeUhr{zeiten: []time.Time{t0, t0, ms(3), t0, ms(4), ms(10)}}
	a := neuMitMax(t, senke, uhr, slog.LevelInfo, 0)
	gesamt := werteVon([]string{"http.request.method", "http.response.status_code", "recorder.mode"}, gueltigSpan)
	ctx, wurzel := a.StartSpan(context.Background(), support.SpanGatewayRequest, gesamt...)
	_, match := a.StartSpan(ctx, support.SpanReplayMatch)
	match.End()
	_, anfrage := a.StartSpan(ctx, support.SpanMockQuery)
	anfrage.End()
	wurzel.End()
	var ids, dauern = map[string]string{}, map[string]string{}
	var eltern = map[string]string{}
	for _, w := range senke.zeilen() {
		g := lies(t, w)
		ids[g.text["span_id"]], dauern[g.text["msg"]], eltern[g.text["msg"]] = g.text["msg"], g.text["duration_ms"], g.text["parent_id"]
	}
	wurzelID := orakelID(1)
	if ids[wurzelID] != "odata.gateway.request" || eltern["odata.replay.match"] != wurzelID || eltern["odata.mock.query"] != wurzelID ||
		eltern["odata.gateway.request"] != "" || dauern["odata.gateway.request"] != "10.000" ||
		dauern["odata.replay.match"] != "3.000" || dauern["odata.mock.query"] != "4.000" {
		t.Errorf("Gesamtspan nicht lesbar: IDs %v, Eltern %v, Dauern %v", ids, eltern, dauern)
	}
}

// S12, S21 (F1): die Attribute stehen in Eintrag und Zeile in der Reihenfolge der Tabelle in §5, nicht in
// der Folge des Setzens. Menge: jede Anordnung der Pflicht-Attribute jedes Spans (3! + 4 × 2! = 14, aus
// der Länge von Pflicht gezählt) × 3 Wege des Setzens (alle in StartSpan, alle in einem Set, je eines in
// einem Set) = 42 Läufe; je Span ein Adapter, so dass spätere Läufe nicht der erste Span sind. Orakel:
// der Eintrag im Ring, die Folge der Schlüssel der Zeile und die Lage der Schlüssel in den Rohbytes.
func TestAttributFolge(t *testing.T) {
	laeufe, erwartet := 0, 0
	for _, z := range spanTabelle(t) {
		fak := 1
		for i := 2; i <= len(z.pflicht); i++ {
			fak *= i
		}
		perms := permutationen(z.pflicht)
		if len(perms) != fak {
			t.Fatalf("%s: %d Anordnungen, erwartet %d", z.name, len(perms), fak)
		}
		erwartet += 3 * fak
		senke := &FakeSenke{}
		a := neuMitMax(t, senke, festeUhr(), slog.LevelInfo, 0)
		soll := werteVon(z.pflicht, gueltigSpan)
		folge := append([]string{"timestamp", "level", "msg", "ereignis", "span_id", "duration_ms"}, z.pflicht...)
		for _, p := range perms {
			for _, weg := range []string{"StartSpan", "Set", "einzeln"} {
				laeufe++
				attrs := werteVon(p, gueltigSpan)
				var sp support.Span
				if weg == "StartSpan" {
					_, sp = a.StartSpan(context.Background(), support.SpanName(z.name), attrs...)
				} else {
					_, sp = a.StartSpan(context.Background(), support.SpanName(z.name))
					if weg == "Set" {
						sp.Set(attrs...)
					} else {
						for _, at := range attrs {
							sp.Set(at)
						}
					}
				}
				sp.End()
				wo := z.name + " " + weg + " " + strings.Join(p, ",")
				spans := a.Snapshot().Spans
				if r := spans[len(spans)-1]; !attrsGleich(r.Attrs, soll) {
					t.Errorf("%s: Eintrag %+v, erwartet %+v", wo, r.Attrs, soll)
				}
				zeilen := senke.zeilen()
				zeile := zeilen[len(zeilen)-1]
				if g := lies(t, zeile); !slices.Equal(g.schluessel, folge) {
					t.Errorf("%s: Felder %v, erwartet %v", wo, g.schluessel, folge)
				}
				letzte := -1
				for _, k := range z.pflicht {
					i := bytes.Index(zeile, []byte(`"`+k+`":`))
					if i <= letzte {
						t.Errorf("%s: %q steht in den Rohbytes nicht hinter seinem Vorgänger: %q", wo, k, zeile)
					}
					letzte = i
				}
			}
		}
	}
	if laeufe != erwartet || laeufe != 42 {
		t.Errorf("%d Läufe, erwartet %d und 42", laeufe, erwartet)
	}
}

package telemetry_test

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

type ctxWert struct{}

// nachID ordnet die Einträge eines Stands nach ihrer ID.
func nachID(s support.Snapshot) map[string]support.SpanRecord {
	m := map[string]support.SpanRecord{}
	for _, r := range s.Spans {
		m[r.ID] = r
	}
	return m
}

// S01 (M3): ein Name außerhalb der Menge liefert den übergebenen ctx und einen Span ohne Wirkung und
// zählt genau einen spanname; 8 Namen × 2 Attributlisten × 2 ctx = 32 Läufe, dazu die 5 Namen der
// Quelle als Gegenprobe.
func TestSpanNamen(t *testing.T) {
	tab := spanTabelle(t)
	fremd := []string{
		"", "ODATA.GATEWAY.REQUEST", "odata.gateway.request ", "odata.gateway", "odata.gateway.requests",
		"odata.mock", "odata.foo.bar", "Bearer ZQ9K-name",
	}
	for _, n := range fremd {
		for _, z := range tab {
			if n == z.name {
				t.Fatalf("der fremde Name %q liegt in support.Spans()", n)
			}
		}
	}
	var kein context.Context
	ctxs := []context.Context{context.WithValue(context.Background(), ctxWert{}, 1), kein}
	attrFormen := [][]support.Attr{nil, {support.Str("mode", "record")}}
	laeufe := 0
	for _, n := range fremd {
		for _, at := range attrFormen {
			for _, c := range ctxs {
				laeufe++
				wo := "Name " + n
				u := neuUmgebung(t, slog.LevelInfo, 0)
				c2, sp := u.a.StartSpan(c, support.SpanName(n), at...)
				if c2 == nil || sp == nil || (c != nil && c2 != c) {
					t.Errorf("%s: ctx %v Span %v; erwartet der übergebene ctx (nil: ein nicht-nil ctx) und ein Span", wo, c2, sp)
				}
				sp.Set(gueltigSpan("recorder.mode"), support.Str("x", "y"))
				sp.End()
				sp.End()
				s := u.a.Snapshot()
				pruefeVektor(t, wo, s, map[support.Verwurf]uint64{support.VerwurfSpanname: 1})
				if len(s.Durations) != len(tab) || len(s.Spans) != 0 || u.senke.anzahl() != 0 || u.uhr.n.Load() != 0 {
					t.Errorf("%s: %d Verteilungen, %d Einträge, %d Write, %d Ablesungen", wo, len(s.Durations), len(s.Spans), u.senke.anzahl(), u.uhr.n.Load())
				}
				for _, z := range tab {
					if v, ok := s.Durations[support.SpanName(z.name)]; !ok || v.Count != 0 {
						t.Errorf("%s: Verteilung von %s: %+v", wo, z.name, v)
					}
				}
			}
		}
	}
	if laeufe != 32 || len(fremd) != 8 {
		t.Errorf("%d Läufe, %d Namen; erwartet 32 und 8", laeufe, len(fremd))
	}
	for _, z := range tab {
		u := neuUmgebung(t, slog.LevelInfo, 0)
		_, sp := u.a.StartSpan(context.Background(), support.SpanName(z.name))
		pruefeVektor(t, "Gegenprobe "+z.name, u.a.Snapshot(), nil)
		if u.uhr.n.Load() != 1 || sp == nil {
			t.Errorf("Gegenprobe %s: %d Ablesungen", z.name, u.uhr.n.Load())
		}
	}
}

// S02 (M4): fehlt beim End ein Pflicht-Attribut, wird aufgezeichnet und pflicht zählt je fehlendes;
// ein ungültiger Wert zählt wert beim Setzen und pflicht beim Ende. Je Span jede Teilmenge seiner
// Pflicht-Attribute × 2 Wege (alle in StartSpan, alle über Set); dazu je Pflicht-Attribut ein
// ungültiger Wert × 2 Wege.
func TestPflichtAttribute(t *testing.T) {
	tab := spanTabelle(t)
	teil, ungueltig, erwartetTeil := 0, 0, 0
	for _, s := range support.Spans() {
		erwartetTeil += 2 * (1 << len(s.Pflicht))
	}
	setzen := func(u umgebung, name string, weg string, attrs []support.Attr) {
		var sp support.Span
		if weg == "start" {
			_, sp = u.a.StartSpan(context.Background(), support.SpanName(name), attrs...)
		} else {
			_, sp = u.a.StartSpan(context.Background(), support.SpanName(name))
			sp.Set(attrs...)
		}
		sp.End()
	}
	for _, z := range tab {
		for maske := range 1 << len(z.pflicht) {
			var keys []string
			for i, k := range z.pflicht {
				if maske>>i&1 == 1 {
					keys = append(keys, k)
				}
			}
			for _, weg := range []string{"start", "set"} {
				teil++
				u := neuUmgebung(t, slog.LevelInfo, 0)
				attrs := werteVon(keys, gueltigSpan)
				setzen(u, z.name, weg, attrs)
				s := u.a.Snapshot()
				wo := z.name + " " + weg
				pruefeVektor(t, wo, s, map[support.Verwurf]uint64{support.VerwurfPflicht: uint64(len(z.pflicht) - len(keys))})
				if len(s.Spans) != 1 || !attrsGleich(s.Spans[0].Attrs, attrs) || s.Durations[support.SpanName(z.name)].Count != 1 {
					t.Errorf("%s %v: Einträge %+v", wo, keys, s.Spans)
				}
			}
		}
		for _, key := range z.pflicht {
			for _, weg := range []string{"start", "set"} {
				ungueltig++
				u := neuUmgebung(t, slog.LevelInfo, 0)
				var attrs, gilt []support.Attr
				for _, k := range z.pflicht {
					if k == key {
						attrs = append(attrs, ungueltigSpan(k))
						continue
					}
					attrs = append(attrs, gueltigSpan(k))
					gilt = append(gilt, gueltigSpan(k))
				}
				setzen(u, z.name, weg, attrs)
				s := u.a.Snapshot()
				pruefeVektor(t, key+" "+weg, s, map[support.Verwurf]uint64{support.VerwurfWert: 1, support.VerwurfPflicht: 1})
				if len(s.Spans) != 1 || !attrsGleich(s.Spans[0].Attrs, gilt) {
					t.Errorf("%s %s: Eintrag %+v, erwartet %+v", key, weg, s.Spans, gilt)
				}
			}
		}
	}
	if teil != erwartetTeil || ungueltig != 2*len(alleSpanKeys(t)) || teil != 48 {
		t.Errorf("%d Teilmengen-Läufe (erwartet %d, 48), %d ungültige Läufe", teil, erwartetTeil, ungueltig)
	}
}

// S03 (M2): an einem Span sind allein seine Pflicht-Attribute zulässig; alles andere zählt genau
// ein schluessel, auch mit leerem Wert. 5 Spans × 32 Schlüssel = 160 Paare (11 zulässige, 149 fremde)
// × 2 Aufrufstellen; die fremden zusätzlich mit leerem Wert.
func TestSpanSchluessel(t *testing.T) {
	tab := spanTabelle(t)
	keys := slices.Clone(alleSpanKeys(t))
	keys = append(keys, logSchluessel(t)...)
	keys = append(keys, "timestamp", "level", "span_id", "parent_id", "", "x", "Recorder.mode", "recorder.mode ")
	if len(keys) != 11+13+4+4 || len(tab)*len(keys) != 160 {
		t.Fatalf("%d Schlüssel, %d Paare; erwartet 32 und 160", len(keys), len(tab)*len(keys))
	}
	zulaessig, fremd, pruefungen := 0, 0, 0
	for _, z := range tab {
		for _, k := range keys {
			ok := slices.Contains(z.pflicht, k)
			werte := []support.Attr{gueltigFuer(k)}
			if ok {
				werte[0] = gueltigSpan(k)
				zulaessig++
			} else {
				werte = append(werte, support.Str(k, ""))
				fremd++
			}
			for _, wert := range werte {
				for _, weg := range []string{"start", "set"} {
					pruefungen++
					u := neuUmgebung(t, slog.LevelInfo, 0)
					var sp support.Span
					if weg == "start" {
						_, sp = u.a.StartSpan(context.Background(), support.SpanName(z.name), wert)
					} else {
						_, sp = u.a.StartSpan(context.Background(), support.SpanName(z.name))
						sp.Set(wert)
					}
					sp.End()
					s := u.a.Snapshot()
					wollen, gilt := uint64(1), []support.Attr(nil)
					if ok {
						wollen, gilt = 0, []support.Attr{wert}
					}
					if s.Dropped[support.VerwurfSchluessel] != wollen || s.Dropped[support.VerwurfWert] != 0 ||
						len(s.Spans) != 1 || !attrsGleich(s.Spans[0].Attrs, gilt) {
						t.Errorf("%s %q %s: schluessel %d wert %d, Einträge %+v; erwartet schluessel %d", z.name, k, weg,
							s.Dropped[support.VerwurfSchluessel], s.Dropped[support.VerwurfWert], s.Spans, wollen)
					}
				}
			}
		}
	}
	if zulaessig != 11 || fremd != 149 || pruefungen != 2*(160+149) {
		t.Errorf("%d zulässige, %d fremde Paare, %d Prüfungen; erwartet 11, 149 und %d", zulaessig, fremd, pruefungen, 2*(160+149))
	}
	for _, k := range []string{"span_id", "parent_id"} {
		u := neuUmgebung(t, slog.LevelInfo, 0)
		u.a.Log(context.Background(), slog.LevelInfo, "m", support.Str(k, "spn-0000000000000001"))
		s := u.a.Snapshot()
		if s.Dropped[support.VerwurfSchluessel] != 1 || strings.Contains(string(u.senke.bytesAlle()), k) {
			t.Errorf("Log mit %s: schluessel %d, Zeile %q", k, s.Dropped[support.VerwurfSchluessel], u.senke.bytesAlle())
		}
	}
}

// S04: Set ersetzt einen gültigen Wert; ein ungültiger lässt den früheren stehen; derselbe Schlüssel
// zweimal im selben Aufruf: der erste gilt; Set ohne Attribut ändert und zählt nichts. 11 Schlüssel
// × 5 Fälle.
func TestSetErsetzt(t *testing.T) {
	läufe := 0
	for _, z := range spanTabelle(t) {
		for _, key := range z.pflicht {
			v1, v2, schlecht := gueltigSpan(key), anderesSpan(key), ungueltigSpan(key)
			if v1 == v2 {
				t.Fatalf("%s: die beiden gültigen Werte sind gleich", key)
			}
			faelle := []struct {
				name   string
				lauf   func(a *telemetry.Adapter) support.Span
				gilt   support.Attr
				vektor map[support.Verwurf]uint64
			}{
				{"gültig nach gültig", func(a *telemetry.Adapter) support.Span {
					_, sp := a.StartSpan(context.Background(), support.SpanName(z.name), v1)
					sp.Set(v2)
					return sp
				}, v2, nil},
				{"ungültig nach gültig", func(a *telemetry.Adapter) support.Span {
					_, sp := a.StartSpan(context.Background(), support.SpanName(z.name), v1)
					sp.Set(schlecht)
					return sp
				}, v1, map[support.Verwurf]uint64{support.VerwurfWert: 1}},
				{"doppelt in Set", func(a *telemetry.Adapter) support.Span {
					_, sp := a.StartSpan(context.Background(), support.SpanName(z.name))
					sp.Set(v1, v2)
					return sp
				}, v1, map[support.Verwurf]uint64{support.VerwurfSchluessel: 1}},
				{"Set ohne Attribut", func(a *telemetry.Adapter) support.Span {
					_, sp := a.StartSpan(context.Background(), support.SpanName(z.name), v1)
					sp.Set()
					return sp
				}, v1, nil},
				{"doppelt in StartSpan", func(a *telemetry.Adapter) support.Span {
					_, sp := a.StartSpan(context.Background(), support.SpanName(z.name), v1, v2)
					return sp
				}, v1, map[support.Verwurf]uint64{support.VerwurfSchluessel: 1}},
			}
			for _, f := range faelle {
				läufe++
				u := neuUmgebung(t, slog.LevelInfo, 0)
				f.lauf(u.a).End()
				s := u.a.Snapshot()
				erwartet := map[support.Verwurf]uint64{support.VerwurfPflicht: uint64(len(z.pflicht) - 1)}
				for g, n := range f.vektor {
					erwartet[g] = n
				}
				pruefeVektor(t, key+" "+f.name, s, erwartet)
				if len(s.Spans) != 1 || !attrsGleich(s.Spans[0].Attrs, []support.Attr{f.gilt}) {
					t.Errorf("%s %s: Eintrag %+v, erwartet %+v", key, f.name, s.Spans, f.gilt)
				}
			}
		}
	}
	if läufe != 11*5 {
		t.Errorf("%d Läufe, erwartet 55", läufe)
	}
}

// S06 (M8): der Rückgabe-ctx leitet den übergebenen ab und trägt den Span; bei fremdem Namen ist er der
// übergebene; ein nil-ctx gilt als Background. 8 Formen des ctx × 2 Namen, dazu zwei Geschwister.
func TestSpanCtx(t *testing.T) {
	tab := spanTabelle(t)
	abgebrochen, abbruch := context.WithCancel(context.Background())
	abbruch()
	frist, fristEnde := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer fristEnde()
	wert := context.WithValue(context.Background(), ctxWert{}, "w")
	fremderAdapter := neuUmgebung(t, slog.LevelInfo, 0).a
	ctxAnderer, _ := fremderAdapter.StartSpan(context.Background(), support.SpanName(tab[1].name))
	ctxFake, _ := supporttest.NewFakeTelemetry().StartSpan(context.Background(), support.SpanName(tab[1].name))
	var kein context.Context
	formen := []struct {
		name   string
		ctx    func(a *telemetry.Adapter) context.Context
		eltern bool
	}{
		{"nil", func(*telemetry.Adapter) context.Context { return kein }, false},
		{"Background", func(*telemetry.Adapter) context.Context { return context.Background() }, false},
		{"abgebrochen", func(*telemetry.Adapter) context.Context { return abgebrochen }, false},
		{"Frist abgelaufen", func(*telemetry.Adapter) context.Context { return frist }, false},
		{"Wert", func(*telemetry.Adapter) context.Context { return wert }, false},
		{"Span desselben Adapters", func(a *telemetry.Adapter) context.Context {
			c, _ := a.StartSpan(wert, support.SpanName(tab[1].name))
			return c
		}, true},
		{"Span eines anderen Adapters", func(*telemetry.Adapter) context.Context { return ctxAnderer }, false},
		{"Span einer FakeTelemetry", func(*telemetry.Adapter) context.Context { return ctxFake }, false},
	}
	laeufe := 0
	for _, f := range formen {
		for _, name := range []string{tab[2].name, "odata.fremd"} {
			laeufe++
			wo := f.name + " / " + name
			u := neuUmgebung(t, slog.LevelInfo, 0)
			c := f.ctx(u.a)
			c2, sp := u.a.StartSpan(c, support.SpanName(name))
			if c2 == nil || sp == nil {
				t.Fatalf("%s: ctx %v Span %v", wo, c2, sp)
			}
			if name == "odata.fremd" {
				if c != nil && c2 != c {
					t.Errorf("%s: fremder Name liefert nicht den übergebenen ctx", wo)
				}
				continue
			}
			if c == nil {
				c = context.Background()
			}
			d1, ok1 := c.Deadline()
			d2, ok2 := c2.Deadline()
			if c2 == c || c2.Err() != c.Err() || d1 != d2 || ok1 != ok2 || c2.Value(ctxWert{}) != c.Value(ctxWert{}) {
				t.Errorf("%s: der Rückgabe-ctx leitet nicht ab (gleich %v, Err %v, Frist %v, Wert %v)", wo, c2 == c, c2.Err(), ok2, c2.Value(ctxWert{}))
			}
			select {
			case <-c2.Done():
				if c.Err() == nil {
					t.Errorf("%s: Rückgabe-ctx ist beendet, der übergebene nicht", wo)
				}
			default:
				if c.Err() != nil {
					t.Errorf("%s: Rückgabe-ctx ist nicht beendet", wo)
				}
			}
			_, kind := u.a.StartSpan(c2, support.SpanName(tab[3].name))
			kind.End()
			sp.End()
			elternID := ""
			if f.eltern {
				elternID = orakelID(1)
			}
			selbst, kindID := orakelID(1), orakelID(2)
			if f.eltern {
				selbst, kindID = orakelID(2), orakelID(3)
			}
			recs := nachID(u.a.Snapshot())
			if recs[selbst].ParentID != elternID || recs[kindID].ParentID != selbst || recs[selbst].Name != support.SpanName(name) {
				t.Errorf("%s: Einträge %+v, erwartet Eltern %q und Kind-Eltern %q", wo, recs, elternID, selbst)
			}
		}
	}
	if laeufe != 16 {
		t.Errorf("%d Läufe, erwartet 16", laeufe)
	}
	// Geschwister aus einem ctx: der Abbruch des einen wirkt nicht auf das andere
	u := neuUmgebung(t, slog.LevelInfo, 0)
	ctxA, abbruchA := context.WithCancel(wert)
	ctxB, abbruchB := context.WithCancel(wert)
	defer abbruchB()
	a1, spA := u.a.StartSpan(ctxA, support.SpanName(tab[0].name))
	b1, spB := u.a.StartSpan(ctxB, support.SpanName(tab[0].name))
	abbruchA()
	if a1.Err() == nil || b1.Err() != nil {
		t.Errorf("Geschwister: Err %v und %v", a1.Err(), b1.Err())
	}
	_, ka := u.a.StartSpan(a1, support.SpanName(tab[1].name))
	_, kb := u.a.StartSpan(b1, support.SpanName(tab[1].name))
	for _, sp := range []support.Span{ka, kb, spA, spB} {
		sp.End()
	}
	recs := nachID(u.a.Snapshot())
	if recs[orakelID(3)].ParentID != orakelID(1) || recs[orakelID(4)].ParentID != orakelID(2) || recs[orakelID(1)].ParentID != "" {
		t.Errorf("Geschwister teilen einen Span: %+v", recs)
	}
}

// S05 (M7): ParentID ist die ID des Spans desselben Adapters im ctx, sonst leer; für jeden Namen und
// jede Tiefe. Ketten: 25 Paare (Tiefe 1), 15 Ketten der Tiefe 2 bis 4 mit Namen reihum, je End von
// innen und von außen = 80 Läufe; dazu ein Eltern-Span, der vor dem Kind endete, und die drei Formen
// ohne Eltern (anderer Adapter, FakeTelemetry, fremder Wert) je Name.
func TestElternKind(t *testing.T) {
	tab := spanTabelle(t)
	n := len(tab)
	var ketten [][]int
	for e := range n {
		for k := range n {
			ketten = append(ketten, []int{e, k})
		}
	}
	for tiefe := 2; tiefe <= 4; tiefe++ {
		for start := range n {
			k := []int{}
			for i := 0; i <= tiefe; i++ {
				k = append(k, (start+i)%n)
			}
			ketten = append(ketten, k)
		}
	}
	laeufe := 0
	for _, k := range ketten {
		for _, innen := range []bool{true, false} {
			laeufe++
			u := neuUmgebung(t, slog.LevelInfo, 0)
			ctxs, spans := []context.Context{context.Background()}, []support.Span{}
			for _, i := range k {
				c, sp := u.a.StartSpan(ctxs[len(ctxs)-1], support.SpanName(tab[i].name))
				ctxs, spans = append(ctxs, c), append(spans, sp)
			}
			for i := range spans {
				if innen {
					spans[len(spans)-1-i].End()
				} else {
					spans[i].End()
				}
			}
			recs := nachID(u.a.Snapshot())
			for pos, i := range k {
				eltern := ""
				if pos > 0 {
					eltern = orakelID(pos)
				}
				r := recs[orakelID(pos+1)]
				if r.ParentID != eltern || r.Name != support.SpanName(tab[i].name) {
					t.Errorf("Kette %v (innen %v) Stelle %d: Eintrag %+v, erwartet Eltern %q", k, innen, pos, r, eltern)
				}
			}
		}
	}
	if laeufe != (n*n+3*n)*2 {
		t.Errorf("%d Läufe, erwartet %d", laeufe, (n*n+3*n)*2)
	}
	fremderAdapter := neuUmgebung(t, slog.LevelInfo, 0).a
	for i, z := range tab {
		u := neuUmgebung(t, slog.LevelInfo, 0)
		ctx, eltern := u.a.StartSpan(context.Background(), support.SpanName(z.name))
		eltern.End()
		_, kind := u.a.StartSpan(ctx, support.SpanName(tab[(i+1)%n].name))
		kind.End()
		if r := nachID(u.a.Snapshot())[orakelID(2)]; r.ParentID != orakelID(1) {
			t.Errorf("Eltern %s vor dem Kind beendet: ParentID %q", z.name, r.ParentID)
		}
		anderer, _ := fremderAdapter.StartSpan(context.Background(), support.SpanName(z.name))
		fake, _ := supporttest.NewFakeTelemetry().StartSpan(context.Background(), support.SpanName(z.name))
		for form, c := range map[string]context.Context{
			"anderer Adapter": anderer, "FakeTelemetry": fake, "fremder Wert": context.WithValue(context.Background(), ctxWert{}, 1),
		} {
			v := neuUmgebung(t, slog.LevelInfo, 0)
			_, sp := v.a.StartSpan(c, support.SpanName(z.name))
			sp.End()
			if r := v.a.Snapshot().Spans[0]; r.ParentID != "" {
				t.Errorf("%s / %s: ParentID %q, erwartet leer", form, z.name, r.ParentID)
			}
		}
	}
}

// S09: ein zweites End und jedes Set nach dem End sind ohne Wirkung und zählen zustand einmal je
// Aufruf; sie lesen die Uhr nicht und schreiben keine Zeile. 5 Spans × 5 Fälle.
func TestEndDoppelt(t *testing.T) {
	läufe := 0
	for _, z := range spanTabelle(t) {
		for fall := range 5 {
			läufe++
			u := neuUmgebung(t, slog.LevelInfo, 0)
			_, sp := u.a.StartSpan(context.Background(), support.SpanName(z.name), werteVon(z.pflicht, gueltigSpan)...)
			sp.End()
			vor, zeilen, ablesungen := u.a.Snapshot(), u.senke.anzahl(), u.uhr.n.Load()
			switch fall {
			case 0:
				sp.End()
			case 1:
				sp.Set(werteVon(z.pflicht, anderesSpan)...)
			case 2:
				sp.Set(support.Str("x", "y"))
			case 3:
				sp.Set(ungueltigSpan(z.pflicht[0]))
			case 4:
				sp.Set()
			}
			nach := u.a.Snapshot()
			erwartet := map[support.Verwurf]uint64{support.VerwurfZustand: 1}
			pruefeVektor(t, z.name, nach, erwartet)
			if !reflect.DeepEqual(nach.Spans, vor.Spans) || !reflect.DeepEqual(nach.Durations, vor.Durations) ||
				u.senke.anzahl() != zeilen || u.uhr.n.Load() != ablesungen {
				t.Errorf("%s Fall %d: Stand, Zeilen (%d, vorher %d) oder Ablesungen (%d, vorher %d) geändert", z.name, fall, u.senke.anzahl(), zeilen, u.uhr.n.Load(), ablesungen)
			}
		}
	}
	if läufe != 25 {
		t.Errorf("%d Läufe, erwartet 25", läufe)
	}
}

// S10: End zeichnet auch bei beendetem ctx und nach dem Ende des Eltern-Span auf; Set nimmt Attribute
// am Span eines abgebrochenen ctx an; StartSpan zeichnet bei abgebrochenem ctx auf. 5 Fälle × 5 Spans.
func TestEndNachAbbruch(t *testing.T) {
	läufe := 0
	for _, z := range spanTabelle(t) {
		attrs := werteVon(z.pflicht, gueltigSpan)
		faelle := []struct {
			name string
			lauf func(a *telemetry.Adapter)
			zahl int
		}{
			{"ctx beim Start abgebrochen", func(a *telemetry.Adapter) {
				c, abbruch := context.WithCancel(context.Background())
				abbruch()
				_, sp := a.StartSpan(c, support.SpanName(z.name), attrs...)
				sp.End()
			}, 1},
			{"ctx nach dem Start abgebrochen", func(a *telemetry.Adapter) {
				c, abbruch := context.WithCancel(context.Background())
				_, sp := a.StartSpan(c, support.SpanName(z.name), attrs...)
				abbruch()
				sp.End()
			}, 1},
			{"Frist abgelaufen", func(a *telemetry.Adapter) {
				c, ende := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
				defer ende()
				_, sp := a.StartSpan(c, support.SpanName(z.name), attrs...)
				sp.End()
			}, 1},
			{"Eltern vor Kind beendet", func(a *telemetry.Adapter) {
				c, eltern := a.StartSpan(context.Background(), support.SpanName(z.name), attrs...)
				_, kind := a.StartSpan(c, support.SpanName(z.name), attrs...)
				eltern.End()
				kind.End()
			}, 2},
			{"Set nach dem Abbruch", func(a *telemetry.Adapter) {
				c, abbruch := context.WithCancel(context.Background())
				_, sp := a.StartSpan(c, support.SpanName(z.name))
				abbruch()
				sp.Set(attrs...)
				sp.End()
			}, 1},
		}
		for _, f := range faelle {
			läufe++
			u := neuUmgebung(t, slog.LevelInfo, 0)
			f.lauf(u.a)
			s := u.a.Snapshot()
			if len(s.Spans) != f.zahl {
				t.Fatalf("%s / %s: %d Einträge, erwartet %d", z.name, f.name, len(s.Spans), f.zahl)
			}
			for _, r := range s.Spans {
				if !attrsGleich(r.Attrs, attrs) {
					t.Errorf("%s / %s: Attribute %+v, erwartet %+v", z.name, f.name, r.Attrs, attrs)
				}
			}
			pruefeVektor(t, z.name+" / "+f.name, s, nil)
		}
	}
	if läufe != 25 {
		t.Errorf("%d Läufe, erwartet 25", läufe)
	}
}

// S11: ein Span ohne End wird nicht aufgezeichnet, und der Adapter hält ihn nicht: nach dem Verwerfen
// aller Verweise des Tests und runtime.GC() läuft der Finalizer jedes der 1000 Spans; ein Span, den der
// Test hält, läuft ihn nicht (Kontrolle).
func TestSpanOhneEnde(t *testing.T) {
	u := neuUmgebung(t, slog.LevelInfo, 0)
	tab := spanTabelle(t)
	var frei atomic.Int64
	beobachten := func(sp support.Span) { runtime.SetFinalizer(sp, func(support.Span) { frei.Add(1) }) }
	starten := func() {
		for i := range 1000 {
			_, sp := u.a.StartSpan(context.Background(), support.SpanName(tab[i%len(tab)].name), gueltigSpan(tab[i%len(tab)].pflicht[0]))
			beobachten(sp)
		}
	}
	starten()
	_, gehalten := u.a.StartSpan(context.Background(), support.SpanName(tab[0].name))
	beobachten(gehalten)
	for i := 0; i < 500 && frei.Load() < 1000; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
	s := u.a.Snapshot()
	if frei.Load() != 1000 || len(s.Spans) != 0 || u.senke.anzahl() != 0 {
		t.Errorf("%d von 1000 Spans freigegeben (der gehaltene zählt nicht); %d Einträge, %d Zeilen", frei.Load(), len(s.Spans), u.senke.anzahl())
	}
	for _, d := range s.Durations {
		if d.Count != 0 {
			t.Errorf("Verteilung zählt einen Span ohne End: %+v", d)
		}
	}
	runtime.KeepAlive(gehalten)
}

// S14 (M14): keine Operation der Spans gerät in Panik. 6 Stellen (StartSpan, Set, End, Snapshot, Set und
// End eines Spans ohne Wirkung) × Argumentarten; Set und End auch nach dem Ende.
func TestSpanKeinePanik(t *testing.T) {
	stellen := 2*reflect.TypeFor[support.Span]().NumMethod() + 1 + reflect.TypeFor[support.TelemetryReader]().NumMethod()
	if stellen != 6 || reflect.TypeFor[support.Telemetry]().NumMethod() != 3 {
		t.Fatalf("%d Stellen, %d Methoden von Telemetry; erwartet 6 und 3", stellen, reflect.TypeFor[support.Telemetry]().NumMethod())
	}
	arten := map[string][]support.Attr{
		"nil-Slice":        nil,
		"leerer Schlüssel": {support.Str("", "x")},
		"Art 0":            {{Key: "recorder.mode"}},
		"Art 99":           {{Key: "recorder.mode", Art: 99, Text: "x"}},
		"NaN":              {{Key: "http.response.status_code", Art: support.ArtFloat, Gleit: math.NaN()}},
		"129 Bytes":        {support.Str("recorder.mode", strings.Repeat("a", 129))},
	}
	ohne := func(wo string, f func()) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("%s: Panik %v", wo, r)
			}
		}()
		f()
	}
	var kein context.Context
	aufrufe := 0
	for art, attrs := range arten {
		for _, c := range []context.Context{kein, context.Background()} {
			u := neuUmgebung(t, slog.LevelInfo, 0)
			var sp, fremd support.Span
			ohne("StartSpan "+art, func() { _, sp = u.a.StartSpan(c, support.SpanGatewayRequest, attrs...); aufrufe++ })
			ohne("StartSpan fremd "+art, func() { _, fremd = u.a.StartSpan(c, "odata.fremd", attrs...); aufrufe++ })
			for _, wo := range []string{"vor dem Ende", "nach dem Ende"} {
				ohne("Set "+wo+" "+art, func() { sp.Set(attrs...); aufrufe++ })
				ohne("Snapshot "+wo+" "+art, func() { u.a.Snapshot(); aufrufe++ })
				ohne("Set ohne Wirkung "+wo+" "+art, func() { fremd.Set(attrs...); aufrufe++ })
				ohne("End "+wo+" "+art, func() { sp.End(); aufrufe++ })
				ohne("End ohne Wirkung "+wo+" "+art, func() { fremd.End(); aufrufe++ })
			}
		}
	}
	if aufrufe != len(arten)*2*12 {
		t.Errorf("%d Aufrufe, erwartet %d", aufrufe, len(arten)*2*12)
	}
}

// S19: jeder Adapter hat seinen eigenen Zustand (Ring, Verteilung, Gründe, Zähler der IDs); dieselbe
// Config gilt mehrfach und wird nicht festgehalten.
func TestAdapterEigen(t *testing.T) {
	tab := spanTabelle(t)
	senke, uhr := &FakeSenke{}, festeUhr()
	cfg := telemetry.Config{Writer: senke, Clock: uhr, MaxSpans: 3}
	a, errA := telemetry.New(cfg)
	b, errB := telemetry.New(cfg)
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}
	for i := range 4 {
		startEnde(a, tab[i%len(tab)].name)
	}
	startEnde(a, "odata.fremd")
	sb := b.Snapshot()
	if len(sb.Spans) != 0 || summe(sb.Errors) != 0 {
		t.Errorf("B zeigt Spans von A: %+v", sb.Spans)
	}
	pruefeVektor(t, "B", sb, nil)
	for _, d := range sb.Durations {
		if d.Count != 0 {
			t.Errorf("B zählt Enden von A: %+v", d)
		}
	}
	startEnde(b, tab[0].name)
	if id := b.Snapshot().Spans[0].ID; id != orakelID(1) {
		t.Errorf("B beginnt bei %s, erwartet %s", id, orakelID(1))
	}
	if s := a.Snapshot(); len(s.Spans) != 3 || s.Spans[2].ID != orakelID(4) || s.Dropped[support.VerwurfSpanname] != 1 {
		t.Errorf("A hat sich durch B geändert: %+v", s.Spans)
	}
	// die Config wird nicht festgehalten: eine Änderung danach wirkt nicht
	spaeter := &FakeSenke{}
	cfg.MaxSpans, cfg.Writer = 1, spaeter
	for i := range 3 {
		startEnde(a, tab[i].name)
	}
	if s := a.Snapshot(); len(s.Spans) != 3 || a.MaxSpans() != 3 {
		t.Errorf("A hält MaxSpans nicht: %d Einträge, MaxSpans %d", len(s.Spans), a.MaxSpans())
	}
	if senke.anzahl() != 4+1+3 || spaeter.anzahl() != 0 {
		t.Errorf("Senke von A und B: %d Zeilen, spätere Senke %d; erwartet 8 und 0", senke.anzahl(), spaeter.anzahl())
	}
}

// S20 (M15): die Uhr wird je Span zweimal gelesen (StartSpan und erstes End, auch wenn die Zeile
// entfällt), nie von Set, vom zweiten End, von Snapshot oder von einem Span ohne Wirkung; der timestamp
// der Zeile ist die Ablesung des Endes. 5 Spans × 3 Modi × 5 Aufrufstellen.
func TestUhrAblesungen(t *testing.T) {
	modi := []struct {
		name    string
		mindest slog.Level
		defekt  bool
	}{{"Zeile", slog.LevelInfo, false}, {"MinLevel 1", 1, false}, {"Senke defekt", slog.LevelInfo, true}}
	start := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)
	ende := start.Add(7 * time.Millisecond)
	laeufe := 0
	for _, z := range spanTabelle(t) {
		for _, m := range modi {
			laeufe++
			senke := &FakeSenke{}
			if m.defekt {
				senke.antwort = func([]byte) (int, error) { return 0, errors.New("defekt") }
			}
			uhr := &folgeUhr{zeiten: []time.Time{start, ende}}
			a := neuMitMax(t, senke, uhr, m.mindest, 0)
			var sp support.Span
			stellen := []struct {
				name string
				tue  func()
				n    int64
			}{
				{"StartSpan", func() { _, sp = a.StartSpan(context.Background(), support.SpanName(z.name)) }, 1},
				{"Set", func() { sp.Set(gueltigSpan(z.pflicht[0])) }, 1},
				{"erstes End", func() { sp.End() }, 2},
				{"zweites End", func() { sp.End() }, 2},
				{"Snapshot", func() { a.Snapshot() }, 2},
			}
			for _, st := range stellen {
				st.tue()
				if got := uhr.n.Load(); got != st.n {
					t.Errorf("%s / %s / %s: %d Ablesungen, erwartet %d", z.name, m.name, st.name, got, st.n)
				}
			}
			if m.name == "Zeile" {
				zeilen := senke.zeilen()
				if len(zeilen) != 1 {
					t.Fatalf("%s: %d Zeilen", z.name, len(zeilen))
				}
				g := lies(t, zeilen[0])
				if g.text["timestamp"] != "2026-11-07T10:00:00.007000000Z" || g.text["duration_ms"] != "7.000" {
					t.Errorf("%s: timestamp %s, duration_ms %s", z.name, g.text["timestamp"], g.text["duration_ms"])
				}
			}
			ohne := festeUhr()
			b := neuMitMax(t, &FakeSenke{}, ohne, slog.LevelInfo, 0)
			_, w := b.StartSpan(context.Background(), "odata.fremd")
			w.Set(gueltigSpan(z.pflicht[0]))
			w.End()
			if ohne.n.Load() != 0 {
				t.Errorf("%s: ein Span ohne Wirkung las die Uhr %d-mal", z.name, ohne.n.Load())
			}
		}
	}
	if laeufe != 15 {
		t.Errorf("%d Läufe, erwartet 15", laeufe)
	}
}

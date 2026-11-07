package telemetry_test

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

const (
	masseSchreiber = 16
	masseEnden     = 500
	masseLeser     = 8
)

// offenerSpanCtx beginnt einen Span, der nie endet, und liefert den ctx, der ihn trägt.
func offenerSpanCtx(a *telemetry.Adapter, name string) context.Context {
	c, _ := a.StartSpan(context.Background(), support.SpanName(name))
	return c
}

// masseLauf fährt 16 Schreiber mit je einem Anker-Span (nie beendet, sein Kind-Verhältnis trennt die
// Zeilen je Schreiber) und 500 Enden reihum über die Namen, nach jedem End ein CountError und ein
// Log mit einem fremden Schlüssel, dazu 8 Leser, die pruefe auf jedem Stand rufen. Der Ring fasst alle
// 8000 Spans.
func masseLauf(t *testing.T, pruefe func(s support.Snapshot) string) umgebung {
	t.Helper()
	tab := spanTabelle(t)
	u := neuUmgebung(t, slog.LevelInfo, 65536)
	anker := make([]context.Context, masseSchreiber)
	for g := range anker {
		anker[g] = offenerSpanCtx(u.a, tab[0].name)
	}
	start := make(chan struct{})
	melden := make(chan string, 4)
	var schreibend, lesend sync.WaitGroup
	var fertig atomic.Bool
	for g := range masseSchreiber {
		schreibend.Add(1)
		go func() {
			defer schreibend.Done()
			<-start
			for i := range masseEnden {
				_, sp := u.a.StartSpan(anker[g], support.SpanName(tab[(g+i)%len(tab)].name))
				sp.End()
				u.a.CountError(model.CodeE2001)
				u.a.Log(context.Background(), slog.LevelInfo, "m", support.Str("x", "y"))
			}
		}()
	}
	for range masseLeser {
		lesend.Add(1)
		go func() {
			defer lesend.Done()
			<-start
			for {
				if m := pruefe(u.a.Snapshot()); m != "" {
					select {
					case melden <- m:
					default:
					}
				}
				if fertig.Load() {
					return
				}
				runtime.Gosched()
			}
		}()
	}
	close(start)
	schreibend.Wait()
	fertig.Store(true)
	lesend.Wait()
	close(melden)
	for m := range melden {
		t.Error(m)
	}
	return u
}

// verteilungKaputt prüft SPEC-230 Nr. 9 an einem Stand: Count gleich Summe von Counts und die Zahl der
// Einträge je Name gleich Count (der Ring fasst alle Enden); es liefert die Abweichung oder "".
func verteilungKaputt(tab []spanZeile, s support.Snapshot) string {
	imRing := map[string]uint64{}
	for _, r := range s.Spans {
		imRing[string(r.Name)]++
	}
	for _, z := range tab {
		v, summe := s.Durations[support.SpanName(z.name)], uint64(0)
		for _, c := range v.Counts {
			summe += c
		}
		if v.Count != summe || imRing[z.name] != v.Count {
			return fmt.Sprintf("%s: Count %d, Summe der Counts %d, Einträge im Ring %d", z.name, v.Count, summe, imRing[z.name])
		}
	}
	return ""
}

// S15 (M12): die Aufzeichnung eines Endes (Ring und Verteilung) liegt in einem Abschnitt unter der Sperre;
// die Senke sieht nie zwei gleichzeitige Write, 8000 Span-Zeilen, je Schreiber die span_id aufsteigend.
// Wirkungs-Fänger mit Rate: unter der Sperre liegt kein Argument, an dem ein Test anhalten könnte.
func TestEndSnapshotGleichzeitig(t *testing.T) {
	tab := spanTabelle(t)
	u := masseLauf(t, func(s support.Snapshot) string { return verteilungKaputt(tab, s) })
	s := u.a.Snapshot()
	if m := verteilungKaputt(tab, s); m != "" {
		t.Errorf("Endstand: %s", m)
	}
	var enden uint64
	for _, v := range s.Durations {
		enden += v.Count
	}
	if enden != masseSchreiber*masseEnden || u.senke.hoechstwert() != 1 {
		t.Errorf("%d Enden (erwartet %d), höchstens %d gleichzeitige Write", enden, masseSchreiber*masseEnden, u.senke.hoechstwert())
	}
	letzte, zeilen := map[string]string{}, 0
	for _, w := range u.senke.zeilen() {
		g := lies(t, w)
		if g.text["ereignis"] != "span" {
			continue
		}
		zeilen++
		if g.text["span_id"] <= letzte[g.text["parent_id"]] {
			t.Errorf("Schreiber %s: %s folgt auf %s", g.text["parent_id"], g.text["span_id"], letzte[g.text["parent_id"]])
		}
		letzte[g.text["parent_id"]] = g.text["span_id"]
	}
	if zeilen != masseSchreiber*masseEnden || len(letzte) != masseSchreiber {
		t.Errorf("%d Span-Zeilen von %d Schreibern; erwartet %d von %d", zeilen, len(letzte), masseSchreiber*masseEnden, masseSchreiber)
	}
}

// S16 (M12): die Kopie im Snapshot liegt in einem Abschnitt unter der Sperre, und die vier Sammlungen sind
// ein Stand. Je Runde eines Schreibers steigen nacheinander die Enden, Errors und schluessel von
// Dropped; pflicht steigt mit dem End. Orakel je Stand: wie S15, pflicht gleich der Summe der fehlenden
// Pflicht-Attribute der Einträge, und 0 <= Enden - Errors <= 16, 0 <= Errors - schluessel <= 16.
func TestSnapshotEinStand(t *testing.T) {
	tab := spanTabelle(t)
	fehlend := map[string]uint64{}
	for _, z := range tab {
		fehlend[z.name] = uint64(len(z.pflicht))
	}
	kette := func(s support.Snapshot) string {
		if m := verteilungKaputt(tab, s); m != "" {
			return m
		}
		var enden, pflicht uint64
		for _, r := range s.Spans {
			enden++
			pflicht += fehlend[string(r.Name)]
		}
		e, l := s.Errors[model.CodeE2001], s.Dropped[support.VerwurfSchluessel]
		if s.Dropped[support.VerwurfPflicht] != pflicht || enden < e || e < l || enden-e > masseSchreiber || e-l > masseSchreiber {
			return fmt.Sprintf("Kette: Enden %d, Errors %d, schluessel %d, pflicht %d (erwartet %d)", enden, e, l, s.Dropped[support.VerwurfPflicht], pflicht)
		}
		return ""
	}
	u := masseLauf(t, kette)
	s := u.a.Snapshot()
	if m := kette(s); m != "" || s.Errors[model.CodeE2001] != masseSchreiber*masseEnden || s.Dropped[support.VerwurfSchluessel] != masseSchreiber*masseEnden {
		t.Errorf("Endstand: %s (Errors %d)", m, s.Errors[model.CodeE2001])
	}
}

// S17 (M12): die IDs sind paarweise verschieden, haben die Form spn- und 16 Dezimalziffern und zählen
// ab 1 lückenlos: 32 Goroutinen × 500 Spans gleichzeitig. Detektor und Wirkung je Prozess.
func TestSpanIDs(t *testing.T) {
	const goroutinen, je = 32, 500
	tab := spanTabelle(t)
	a := neuMitMax(t, &FakeSenke{}, festeUhr(), slog.LevelError+1, 65536)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := range je {
				startEnde(a, tab[(g+i)%len(tab)].name)
			}
		}()
	}
	close(start)
	wg.Wait()
	s := a.Snapshot()
	gesehen := map[string]int{}
	for _, r := range s.Spans {
		gesehen[r.ID]++
	}
	if len(s.Spans) != goroutinen*je || len(gesehen) != goroutinen*je {
		t.Fatalf("%d Einträge, %d verschiedene IDs; erwartet %d", len(s.Spans), len(gesehen), goroutinen*je)
	}
	for n := 1; n <= goroutinen*je; n++ {
		if gesehen[orakelID(n)] != 1 {
			t.Fatalf("ID %s kommt %d-mal vor, erwartet 1", orakelID(n), gesehen[orakelID(n)])
		}
	}
}

// S18 (M13): keine Sperre des Adapters liegt über Clock.Now(). Eine Uhr, die in Now() CountError, Snapshot,
// Log, StartSpan, Set oder End des eigenen Spans ruft, blockiert den Aufruf nicht: StartSpan × 4 Rückrufe,
// End × 6 Rückrufe = 10 Fälle, je mit einer Frist von 2 s.
func TestUhrRuftAdapter(t *testing.T) {
	tab := spanTabelle(t)
	type rueckruf struct {
		name    string
		tue     func(a *telemetry.Adapter, eigen support.Span)
		nurEnde bool
	}
	rueckrufe := []rueckruf{
		{"CountError", func(a *telemetry.Adapter, _ support.Span) { a.CountError(model.CodeE2001) }, false},
		{"Snapshot", func(a *telemetry.Adapter, _ support.Span) { a.Snapshot() }, false},
		{"Log", func(a *telemetry.Adapter, _ support.Span) {
			a.Log(context.Background(), slog.LevelInfo, "m", support.Str("mode", "record"))
		}, false},
		{"StartSpan", func(a *telemetry.Adapter, _ support.Span) {
			_, sp := a.StartSpan(context.Background(), support.SpanName(tab[1].name))
			sp.End()
		}, false},
		{"Set des Spans", func(_ *telemetry.Adapter, eigen support.Span) { eigen.Set(gueltigSpan(tab[0].pflicht[0])) }, true},
		{"End des Spans", func(_ *telemetry.Adapter, eigen support.Span) { eigen.End() }, true},
	}
	faelle := 0
	for _, r := range rueckrufe {
		for _, ende := range []bool{false, true} {
			if r.nurEnde && !ende {
				continue
			}
			faelle++
			wo := "StartSpan"
			if ende {
				wo = "End"
			}
			wo += " mit Rückruf " + r.name
			senke := &FakeSenke{}
			uhr := &RueckrufUhr{t: time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)}
			a := neuMitMax(t, senke, uhr, slog.LevelInfo, 0)
			var sp support.Span
			uhr.rueckruf = func() { r.tue(a, sp) }
			if !ende {
				if !binnen(t, 2*time.Second, wo, func() { _, sp = a.StartSpan(context.Background(), support.SpanName(tab[0].name)) }) {
					t.Errorf("%s: kehrt nicht binnen 2 s zurück", wo)
				}
				continue
			}
			uhr.gerufen.Store(true)
			_, sp = a.StartSpan(context.Background(), support.SpanName(tab[0].name), werteVon(tab[0].pflicht, gueltigSpan)...)
			uhr.gerufen.Store(false)
			if !binnen(t, 2*time.Second, wo, sp.End) {
				t.Errorf("%s: kehrt nicht binnen 2 s zurück", wo)
				continue
			}
			s := a.Snapshot()
			zustand := uint64(0)
			if r.nurEnde {
				zustand = 1
			}
			zeilen := 1
			if r.name == "Log" || r.name == "StartSpan" {
				zeilen = 2 // die Zeile des Rückrufs und die des Endes
			}
			if s.Dropped[support.VerwurfZustand] != zustand || senke.anzahl() != zeilen || len(s.Spans) < 1 {
				t.Errorf("%s: zustand %d (erwartet %d), %d Zeilen (erwartet %d), %d Einträge", wo, s.Dropped[support.VerwurfZustand], zustand, senke.anzahl(), zeilen, len(s.Spans))
			}
		}
	}
	if faelle != 10 {
		t.Errorf("%d Fälle, erwartet 10", faelle)
	}
}

// S23: von gleichzeitigen End desselben Spans zeichnet genau eines auf, die übrigen zählen zustand und
// lesen die Uhr nicht; Set und End eines Spans sind gleichzeitig sicher. (1) Haltepunkt an der Uhr: das
// erste End hält in Now(), das zweite kehrt zurück, bevor das erste fortfährt; (2) Masse: 16 End und
// 8 Set desselben Spans, 50 Spans.
func TestEndGleichzeitig(t *testing.T) {
	tab := spanTabelle(t)
	t.Run("Haltepunkt an der Uhr", func(t *testing.T) {
		senke := &FakeSenke{}
		uhr := &RueckrufUhr{t: time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)}
		a := neuMitMax(t, senke, uhr, slog.LevelInfo, 0)
		uhr.gerufen.Store(true)
		_, sp := a.StartSpan(context.Background(), support.SpanName(tab[0].name), werteVon(tab[0].pflicht, gueltigSpan)...)
		eintritt, frei, erstesFertig := make(chan struct{}), make(chan struct{}), make(chan struct{})
		uhr.rueckruf = func() {
			eintritt <- struct{}{}
			<-frei
		}
		uhr.gerufen.Store(false)
		go func() {
			sp.End()
			close(erstesFertig)
		}()
		select {
		case <-eintritt:
		case <-time.After(5 * time.Second):
			t.Fatal("das erste End erreicht die Uhr nicht")
		}
		if !binnen(t, 2*time.Second, "zweites End", sp.End) {
			t.Fatal("das zweite End kehrt nicht zurück, während das erste in Now() hält")
		}
		s := a.Snapshot()
		if s.Dropped[support.VerwurfZustand] != 1 || len(s.Spans) != 0 || senke.anzahl() != 0 {
			t.Errorf("während das erste hält: zustand %d, %d Einträge, %d Zeilen; erwartet 1, 0, 0", s.Dropped[support.VerwurfZustand], len(s.Spans), senke.anzahl())
		}
		close(frei)
		select {
		case <-erstesFertig:
		case <-time.After(5 * time.Second):
			t.Fatal("das erste End kehrt nach der Freigabe nicht zurück")
		}
		s = a.Snapshot()
		if len(s.Spans) != 1 || s.Dropped[support.VerwurfZustand] != 1 || senke.anzahl() != 1 {
			t.Errorf("nach der Freigabe: %d Einträge, zustand %d, %d Zeilen; erwartet 1, 1, 1", len(s.Spans), s.Dropped[support.VerwurfZustand], senke.anzahl())
		}
	})
	t.Run("Masse", func(t *testing.T) {
		const enden, setzer, spans = 16, 8, 50
		u := neuUmgebung(t, slog.LevelError+1, 0)
		for n := 1; n <= spans; n++ {
			z := tab[n%len(tab)]
			_, sp := u.a.StartSpan(context.Background(), support.SpanName(z.name), werteVon(z.pflicht, gueltigSpan)...)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for range enden {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					sp.End()
				}()
			}
			for range setzer {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					sp.Set(werteVon(z.pflicht, anderesSpan)...)
				}()
			}
			close(start)
			wg.Wait()
			s := u.a.Snapshot()
			if len(s.Spans) != n {
				t.Fatalf("Span %d: %d Einträge, erwartet %d", n, len(s.Spans), n)
			}
			for i, a := range s.Spans[n-1].Attrs {
				if a != gueltigSpan(z.pflicht[i]) && a != anderesSpan(z.pflicht[i]) {
					t.Errorf("Span %d: Attribut %+v ist keines der gesetzten", n, a)
				}
			}
			if len(s.Spans[n-1].Attrs) != len(z.pflicht) {
				t.Errorf("Span %d: %d Attribute, erwartet %d", n, len(s.Spans[n-1].Attrs), len(z.pflicht))
			}
		}
		if z := u.a.Snapshot().Dropped[support.VerwurfZustand]; z < spans*(enden-1) || z > spans*(enden-1+setzer) {
			t.Errorf("zustand %d, erwartet von %d bis %d", z, spans*(enden-1), spans*(enden-1+setzer))
		}
	})
}

// S24: die Span-Zeile geht durch denselben Schreiber wie Log: ein Write unter der einen Sperre der Zeile,
// im Aufruf von End; die Aufzeichnung geht der Zeile voraus, und Snapshot, CountError, StartSpan und Set
// warten nicht auf die Senke. Haltepunkt an der Senke.
func TestSpanZeileSperre(t *testing.T) {
	tab := spanTabelle(t)
	neu := func() (*telemetry.Adapter, *FakeSenke) {
		senke := &FakeSenke{}
		senke.blockieren()
		return neuMitMax(t, senke, festeUhr(), slog.LevelInfo, 0), senke
	}
	laeuft := func(f func()) chan struct{} {
		fertig := make(chan struct{})
		go func() {
			defer close(fertig)
			f()
		}()
		return fertig
	}
	bleibt := func(c chan struct{}, wo string) {
		t.Helper()
		select {
		case <-c:
			t.Errorf("%s: kehrt zurück, obwohl das Write hängt", wo)
		case <-time.After(200 * time.Millisecond):
		}
	}
	kehrt := func(c chan struct{}, wo string) {
		t.Helper()
		select {
		case <-c:
		case <-time.After(5 * time.Second):
			t.Errorf("%s: kehrt nach der Freigabe nicht zurück", wo)
		}
	}
	t.Run("ein Log hält das Write", func(t *testing.T) {
		a, senke := neu()
		defer senke.freigeben()
		logFertig := laeuft(func() { a.Log(context.Background(), slog.LevelInfo, "m") })
		if !senke.warteEintritt(5 * time.Second) {
			t.Fatal("das Log erreicht die Senke nicht")
		}
		_, sp := a.StartSpan(context.Background(), support.SpanName(tab[0].name))
		endFertig := laeuft(sp.End)
		bleibt(endFertig, "End")
		if senke.anzahl() != 1 || senke.hoechstwert() != 1 {
			t.Errorf("%d Write, höchstens %d gleichzeitig; erwartet 1 und 1", senke.anzahl(), senke.hoechstwert())
		}
		senke.freigeben()
		kehrt(logFertig, "Log")
		kehrt(endFertig, "End")
		if senke.anzahl() != 2 || senke.hoechstwert() != 1 {
			t.Errorf("nach der Freigabe %d Write, höchstens %d gleichzeitig; erwartet 2 und 1", senke.anzahl(), senke.hoechstwert())
		}
	})
	t.Run("ein End hält das Write", func(t *testing.T) {
		a, senke := neu()
		defer senke.freigeben()
		_, sp := a.StartSpan(context.Background(), support.SpanName(tab[0].name), werteVon(tab[0].pflicht, gueltigSpan)...)
		erstes := laeuft(sp.End)
		if !senke.warteEintritt(5 * time.Second) {
			t.Fatal("das End erreicht die Senke nicht")
		}
		var zweites support.Span
		if !binnen(t, 2*time.Second, "Operationen neben dem hängenden Write", func() {
			if s := a.Snapshot(); len(s.Spans) != 1 {
				t.Errorf("Snapshot ohne die Aufzeichnung des Endes: %d Einträge", len(s.Spans))
			}
			a.CountError(model.CodeE2001)
			var c context.Context
			c, zweites = a.StartSpan(context.Background(), support.SpanName(tab[1].name))
			zweites.Set(gueltigSpan(tab[1].pflicht[0]))
			if c == nil {
				t.Error("kein ctx")
			}
		}) {
			t.Fatal("Snapshot, CountError, StartSpan oder Set warten auf die Senke")
		}
		bleibt(erstes, "das erste End")
		weiteres := laeuft(zweites.End)
		bleibt(weiteres, "ein weiteres End")
		senke.freigeben()
		kehrt(erstes, "das erste End")
		kehrt(weiteres, "ein weiteres End")
	})
	t.Run("End kehrt nach Write zurück", func(t *testing.T) {
		a, senke := neu()
		defer senke.freigeben()
		_, sp := a.StartSpan(context.Background(), support.SpanName(tab[2].name))
		fertig := laeuft(sp.End)
		bleibt(fertig, "End")
		senke.freigeben()
		kehrt(fertig, "End")
		if zeilen := senke.zeilen(); len(zeilen) != 1 || lies(t, zeilen[0]).text["msg"] != tab[2].name {
			t.Errorf("Zeilen nach der Rückkehr von End: %q", zeilen)
		}
	})
}

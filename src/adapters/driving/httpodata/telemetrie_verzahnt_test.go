package httpodata_test

import (
	"context"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// treffpunkt lässt zwei Teilnehmer im selben Augenblick weiterlaufen: der zweite, der ankommt, setzt einen Zeitpunkt 2 µs voraus; beide
// warten im Dauerlauf bis dahin und laufen, ohne sich zu begegnen, mit einem Abstand von wenigen Nanosekunden los. start ist der Nullpunkt
// seiner Zeitmarken.
type treffpunkt struct {
	start time.Time
	da    atomic.Int32
	ziel  atomic.Int64
}

func neuerTreffpunkt() *treffpunkt { return &treffpunkt{start: time.Now()} }

func (tp *treffpunkt) jetzt() int64 { return int64(time.Since(tp.start)) }

// treffe wartet auf den anderen Teilnehmer; sie kehrt ohne ihn zurück, wenn dieser fertig ist (fertig > 0) oder nach 2 s (false).
func (tp *treffpunkt) treffe(fertig *atomic.Int32) bool {
	if tp.da.Add(1) == 2 {
		tp.ziel.Store(tp.jetzt() + 2000)
	}
	frist := tp.jetzt() + int64(2*time.Second)
	for i := 0; tp.ziel.Load() == 0; i++ {
		if fertig.Load() > 0 {
			return true
		}
		if i%1024 == 1023 {
			runtime.Gosched()
			if tp.jetzt() > frist {
				return false
			}
		}
	}
	ziel := tp.ziel.Load()
	for i := 0; tp.jetzt() < ziel; i++ {
		if i%1024 == 1023 {
			runtime.Gosched()
		}
	}
	return true
}

// vRunde ist eine Runde mit zwei Teilnehmern: je Operation, die beide Wege ausführen, ein Treffpunkt.
type vRunde struct {
	tp     map[string]*treffpunkt
	fertig atomic.Int32
	zeit   *atomic.Bool
}

func (r *vRunde) treffe(op string) {
	if tp := r.tp[op]; tp != nil && !tp.treffe(&r.fertig) {
		r.zeit.Store(true)
	}
}

// vWeg ist ein Weg der Verzahnung: was der Dienst für die Methode liefert, was der Client und die Telemetrie sehen müssen, und die
// Operationen in der Folge des Handlers, hinter denen die Teilnehmer aufeinander warten (Start: vor dem Aufruf des Handlers; Verarbeiten:
// nach der Rückkehr des Dienstes; die übrigen: nach der Rückkehr der Operation von Telemetry).
type vWeg struct {
	name     string
	methode  string
	tu       httpodata.Verarbeiter
	anhalten bool
	status   int
	code     model.Code
	abbruch  bool
	ops      []string
}

// verzahnteWege liefert die 8 Wege; Methode, Status und Code sind je Weg verschieden (der Status 0 ist der Abbruch, der Code 0 kein Zähler).
func verzahnteWege() []vWeg {
	ant := []string{"Start", "StartSpan", "Verarbeiten", "Set", "End"}
	fehler := []string{"Start", "StartSpan", "Verarbeiten", "CountError", "Set", "End"}
	return []vWeg{
		{"Antwort 201", "GET", antwortMit(201, ausBytes([]byte("abc"))), false, 201, 0, false, ant},
		{"Antwort 304", "HEAD", antwortMit(304, leer()), false, 304, 0, false, ant},
		{"Fehler 409", "POST", fehlerMit(&inbound.Fehler{Code: model.CodeE6009, Status: 409}), false, 409, model.CodeE6009, false, fehler},
		{"Fehler 412", "PUT", fehlerMit(&inbound.Fehler{Code: model.CodeE6010, Status: 412}), false, 412, model.CodeE6010, false, fehler},
		{"Fehler 404", "DELETE", fehlerMit(&inbound.Fehler{Code: model.CodeE6008, Status: 404}), false, 404, model.CodeE6008, false, fehler},
		{"Bruch", "PATCH", bruchOhneBody(), false, 500, model.CodeE9001, false, fehler},
		{"Abbruch", "OPTIONS", fehlerMit(context.Canceled), false, 0, 0, true, []string{"Start", "StartSpan", "Verarbeiten", "End"}},
		{"503", "TRACE", nil, true, 503, model.CodeE5010, false, []string{"Start", "StartSpan", "CountError", "Set", "End"}},
	}
}

// vTeil ist ein Teilnehmer einer Runde.
type vTeil struct {
	w       vWeg
	modus   string
	p       *probe
	jw      *jWriter
	abbruch bool
}

// laufeRunde fährt zwei Wege gleichzeitig durch den Handler, mit demselben Server (derselbe) oder je einem eigenen mit eigenem Fake und mit
// den Modi `record` und `hybrid`; die Teilnehmer treffen sich vor dem Handler, hinter dem Dienst und hinter jeder Operation von Telemetry, die
// beide Wege ausführen. Danach prüft sie je Teilnehmer: Status am Client, Abbruch, den Span seiner Methode (Modus, Status, ein End) und, je
// Server, die Codes des Zählers als Menge. Es liefert die Meldungen.
func laufeRunde(t *testing.T, zeit *atomic.Bool, a, b vWeg, derselbe bool) []string {
	t.Helper()
	r := &vRunde{tp: map[string]*treffpunkt{}, zeit: zeit}
	for _, op := range a.ops {
		if slices.Contains(b.ops, op) {
			r.tp[op] = neuerTreffpunkt()
		}
	}
	ws := []vWeg{a, b}
	dienst := func(ctx context.Context, an inbound.Anfrage) (inbound.Antwort, error) {
		var antwort inbound.Antwort
		var err error
		for _, w := range ws {
			if w.methode == an.Methode && w.tu != nil {
				antwort, err = w.tu(ctx, an)
			}
		}
		r.treffe("Verarbeiten")
		return antwort, err
	}
	var teile []*vTeil
	var gemeinsam *probe
	if derselbe {
		gemeinsam = neuProbe(t, "record", dienst)
	}
	for i, w := range ws {
		tl := &vTeil{w: w, modus: "record", p: gemeinsam}
		if !derselbe {
			tl.modus = []string{"record", "hybrid"}[i]
			tl.p = neuProbe(t, tl.modus, dienst)
		}
		if w.anhalten {
			tl.p.srv.Anhalten()
		}
		tl.jw = neuJWriter(tl.p.j)
		teile = append(teile, tl)
	}
	for _, tl := range teile {
		tl.p.tel.setzeNach(r.treffe)
	}
	var wg sync.WaitGroup
	for _, tl := range teile {
		req := jRequest(context.Background(), tl.w.methode, "/p")
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer r.fertig.Add(1)
			r.treffe("Start")
			tl.abbruch = tl.p.ruf(tl.jw, req)
		}()
	}
	wg.Wait()
	var meldungen []string
	server := "je ein Server"
	if derselbe {
		server = "derselbe Server"
	}
	melde := func(s string) { meldungen = append(meldungen, a.name+" und "+b.name+" ("+server+"): "+s) }
	gesehen := map[*probe]bool{}
	sollCodes := map[*probe][]string{}
	for _, tl := range teile {
		if tl.abbruch != tl.w.abbruch || tl.jw.status != tl.w.status {
			melde(tl.w.name + ": Abbruch/Status des Clients falsch")
		}
		var mein *spanBuch
		for _, s := range buecher(tl.p.tel.FakeTelemetry) {
			if m, _ := attrText(s.start, "http.request.method"); m == tl.w.methode {
				mein = s
			}
		}
		if mein == nil {
			melde(tl.w.name + ": kein Span")
			continue
		}
		if modus, _ := attrText(mein.start, "recorder.mode"); modus != tl.modus {
			melde(tl.w.name + ": Modus " + modus + " statt " + tl.modus)
		}
		if mein.ends != 1 {
			melde(tl.w.name + ": End falsch")
		}
		switch {
		case tl.w.status == 0 && len(mein.sets) != 0:
			melde(tl.w.name + ": Set ohne Antwort")
		case tl.w.status != 0 && (len(mein.sets) != 1 || len(mein.sets[0]) != 1 || mein.sets[0][0].Ganz != int64(tl.w.status)):
			melde(tl.w.name + ": Status im Span falsch")
		}
		if tl.w.code != 0 {
			sollCodes[tl.p] = append(sollCodes[tl.p], tl.w.code.String())
		}
	}
	for _, tl := range teile {
		if gesehen[tl.p] {
			continue
		}
		gesehen[tl.p] = true
		var got []string
		for _, c := range zaehlerFolge(tl.p.tel.FakeTelemetry) {
			got = append(got, c.String())
		}
		slices.Sort(got)
		slices.Sort(sollCodes[tl.p])
		if strings.Join(got, ",") != strings.Join(sollCodes[tl.p], ",") {
			melde("Zähler " + strings.Join(got, ",") + " statt " + strings.Join(sollCodes[tl.p], ","))
		}
	}
	return meldungen
}

// TestGesamtspanVerzahnt hält G14 (SPEC-294 Nr. 9 a) für die Klasse der geordnet geteilten Variable: ein Zustand außerhalb der Anfrage (Feld des
// Servers, Variable des Pakets, auch unter einer Sperre oder atomar) für Status, Code, Methode, Modus oder Span, den eine Anfrage schreibt und
// später liest, vertauscht die Werte zweier Anfragen, sobald diese zwischen Schreiben und Lesen laufen. Zwei Teilnehmer mit verschiedener Methode,
// verschiedenem Status und Code (8 Wege, 28 Paare) laufen gleichzeitig durch den Handler, mit demselben Server (21 Paare ohne den Weg 503, der
// den Server anhält) und mit je einem eigenen Server mit eigenem Fake und anderem Modus (28 Paare), 49 Kombinationen, je 500 Runden (24500); sie treffen
// sich vor dem Handler, hinter dem Dienst und hinter jeder Operation von Telemetry, die beide Wege ausführen (Treffpunkt in jedem Fenster, das
// einen Aufruf des Ports umschließt, und im Dauerlauf bis auf wenige Nanosekunden an den Fenstern ohne Aufruf). Das Orakel sind Literale je Weg:
// Status am Client, Abbruch, im Span der Methode Modus, genau ein Set mit dem Status (keines beim Abbruch) und ein End, je Server die Menge der
// gezählten Codes. Harte Frist 120 s und 2 s je Treffpunkt.
func TestGesamtspanVerzahnt(t *testing.T) {
	const runden = 500
	frist := time.Now().Add(120 * time.Second)
	ws := verzahnteWege()
	if len(ws) != 8 {
		t.Fatalf("%d Wege, erwartet 8", len(ws))
	}
	var zeit atomic.Bool
	kombinationen, rot := 0, 0
	for _, derselbe := range []bool{true, false} {
		for i := range ws {
			for j := i + 1; j < len(ws); j++ {
				if derselbe && (ws[i].anhalten || ws[j].anhalten) {
					continue
				}
				kombinationen++
				for range runden {
					vs := verzahnteWege()
					for _, m := range laufeRunde(t, &zeit, vs[i], vs[j], derselbe) {
						rot++
						t.Errorf("%s", m)
					}
					if rot >= 20 || zeit.Load() || time.Now().After(frist) {
						t.Fatalf("%d Meldungen, Treffpunkt abgelaufen: %v, Frist abgelaufen: %v", rot, zeit.Load(), time.Now().After(frist))
					}
				}
			}
		}
	}
	if kombinationen != 21+28 {
		t.Errorf("%d Kombinationen, erwartet 49", kombinationen)
	}
}

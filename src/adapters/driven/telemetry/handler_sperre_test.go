package telemetry_test

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// blockierendH liefert einen Adapter über einer blockierenden Senke; die Freigabe läuft beim Ende des
// Tests.
func blockierendH(t *testing.T) (*FakeSenke, *telemetry.Adapter) {
	t.Helper()
	senke := &FakeSenke{}
	senke.blockieren()
	t.Cleanup(senke.freigeben)
	return senke, neuMit(t, senke, festeUhr(), slog.LevelInfo)
}

// handleInGoroutine ruft Handle in einer Goroutine; fertig wird geschlossen, wenn Handle zurückkehrt.
func handleInGoroutine(h slog.Handler, msg string) <-chan struct{} {
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		_ = h.Handle(context.Background(), rec(slog.LevelInfo, msg))
	}()
	return fertig
}

// ableitungen sind alle Handler, die aus der Wurzel von a durch höchstens drei Ableitungen WithAttrs
// (A) und WithGroup (G) entstehen, mit ihrem Modell: die Arten von Handlern der Menge M17 in jeder
// Tiefe (1 + 2 + 4 + 8 = 15). Der Name ist die Folge der Ableitungen, "W" die Wurzel.
func ableitungen(t *testing.T, a *telemetry.Adapter) []kette {
	t.Helper()
	wurzel := kette{name: "W", h: a.Handler()}
	alle, stufe := []kette{wurzel}, []kette{wurzel}
	for tiefe := 1; tiefe <= 3; tiefe++ {
		var naechste []kette
		for _, k := range stufe {
			x := s("mode", "record")
			mitA := k.mitAttrs(k.h.WithAttrs([]slog.Attr{x}), x)
			mitA.name = strings.TrimPrefix(k.name, "W") + "A"
			mitG := k.mitGruppe(k.h.WithGroup("g"), "g")
			mitG.name = strings.TrimPrefix(k.name, "W") + "G"
			naechste = append(naechste, mitA, mitG)
		}
		alle, stufe = append(alle, naechste...), naechste
	}
	if len(alle) != 1+2+4+8 {
		t.Fatalf("%d Handler, erwartet 15", len(alle))
	}
	return alle
}

// H10 (1): Haltepunkt am Writer: A schreibt über die Wurzel und hängt im Write; jeder der 14 anderen
// Handler aus ableitungen (Menge M17 in jeder Tiefe bis 3) schreibt zugleich; innerhalb 200 ms tritt
// kein zweites Write ein, nach der Freigabe stehen 15 Zeilen da, nie zwei Write zugleich.
func TestHandlerKeineZweiWriteHaltepunkt(t *testing.T) {
	senke, a := blockierendH(t)
	alle := ableitungen(t, a)
	fertige := []<-chan struct{}{handleInGoroutine(alle[0].h, "W")}
	if !senke.warteEintritt(5 * time.Second) {
		t.Fatal("A erreicht Write nicht")
	}
	for _, k := range alle[1:] {
		fertige = append(fertige, handleInGoroutine(k.h, k.name))
	}
	time.Sleep(200 * time.Millisecond)
	if z := senke.zeilen(); len(z) != 1 {
		var namen []string
		for _, b := range z[1:] {
			namen = append(namen, lies(t, b).text["msg"])
		}
		t.Errorf("ein zweites Write trat ein, während das erste hing: Handler %v", namen)
	}
	senke.freigeben()
	for _, f := range fertige {
		if !zurueck(f, 2*time.Second) {
			t.Fatal("ein Handle kehrt nach der Freigabe nicht zurück")
		}
	}
	if len(senke.zeilen()) != len(alle) || senke.hoechstwert() != 1 {
		t.Errorf("%d Zeilen, Höchstwert gleichzeitiger Write %d; erwartet %d und 1", len(senke.zeilen()), senke.hoechstwert(), len(alle))
	}
}

// H10 (3): Handle kehrt nicht zurück, solange Write blockiert (100 ms Beobachtung), danach sofort; das
// Write läuft im Aufruf von Handle.
func TestHandleKehrtNachWriteZurueck(t *testing.T) {
	senke, a := blockierendH(t)
	fertig := handleInGoroutine(a.Handler(), "a")
	if !senke.warteEintritt(5 * time.Second) {
		t.Fatal("kein Write")
	}
	if zurueck(fertig, 100*time.Millisecond) {
		t.Error("Handle kehrte zurück, während Write blockierte")
	}
	senke.freigeben()
	if !zurueck(fertig, 2*time.Second) {
		t.Error("Handle kehrt nach der Freigabe nicht zurück")
	}
}

// H10 (2): Masse: 32 Goroutinen × 200 Einträge, verteilt auf die 15 Handler aus ableitungen (jeder
// wird von mindestens einer Goroutine geschrieben) und slog.New: 6400 gültige Zeilen, nie zwei Write
// zugleich, je Goroutine streng aufsteigend (die Nummer steht im Text).
func TestHandlerKeineZweiWriteMasse(t *testing.T) {
	const goroutinen, runden = 32, 200
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	alle := ableitungen(t, a)
	logger := slog.New(alle[0].h)
	var schreiber []func(msg string)
	for _, k := range alle {
		schreiber = append(schreiber, func(m string) { _ = k.h.Handle(context.Background(), rec(slog.LevelInfo, m)) })
	}
	schreiber = append(schreiber, func(m string) { logger.Info(m) })
	if goroutinen < len(schreiber) {
		t.Fatalf("%d Goroutinen für %d Schreiber", goroutinen, len(schreiber))
	}
	var gruppe sync.WaitGroup
	start := make(chan struct{})
	for g := range goroutinen {
		gruppe.Add(1)
		go func() {
			defer gruppe.Done()
			<-start
			for n := 1; n <= runden; n++ {
				schreiber[g%len(schreiber)]("g" + strconv.Itoa(g) + ":" + strconv.Itoa(n))
			}
		}()
	}
	close(start)
	gruppe.Wait()
	zeilen := senke.zeilen()
	if len(zeilen) != goroutinen*runden {
		t.Fatalf("%d Zeilen, erwartet %d", len(zeilen), goroutinen*runden)
	}
	letzte := map[string]int{}
	for _, z := range zeilen {
		g, n, ok := strings.Cut(lies(t, z).text["msg"], ":")
		nr, err := strconv.Atoi(n)
		if !ok || err != nil || nr != letzte[g]+1 {
			t.Fatalf("Zeile %q: Nummer %q nach %d", z, n, letzte[g])
		}
		letzte[g] = nr
	}
	if len(letzte) != goroutinen || senke.hoechstwert() != 1 {
		t.Errorf("%d Goroutinen, Höchstwert gleichzeitiger Write %d; erwartet %d und 1", len(letzte), senke.hoechstwert(), goroutinen)
	}
}

// H10 (4): Masse unter dem Detektor: 32 Goroutinen leiten auf demselben Elternhandler ab (WithAttrs mit
// befunde gleich der Nummer, WithGroup) und schreiben mit dem Ergebnis; der Elternhandler hat eine
// Kette aus 0 bis 8 vorgelegten Attributen (jede Kapazität, die das Wachstum der Liste ergibt). Jede
// Zeile trägt die Felder des Elternhandlers, die gerade Nummer dazu befunde gleich der eigenen.
func TestHandlerAbleitenMasse(t *testing.T) {
	const goroutinen, runden = 32, 20
	folge := logSchluessel(t)
	for laenge := 0; laenge <= 8; laenge++ {
		senke := &FakeSenke{}
		a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
		k := kette{h: a.Handler()}
		for _, key := range folge[:laenge] {
			x := slogAttr(gueltigFuer(key))
			k = k.mitAttrs(k.h.WithAttrs([]slog.Attr{x}), x)
		}
		var gruppe sync.WaitGroup
		start := make(chan struct{})
		for g := range goroutinen {
			gruppe.Add(1)
			go func() {
				defer gruppe.Done()
				<-start
				for range runden {
					msg := "g" + strconv.Itoa(g)
					if g%2 == 0 {
						_ = k.h.WithAttrs([]slog.Attr{slog.Int64("befunde", int64(g))}).Handle(context.Background(), rec(slog.LevelInfo, msg))
					} else {
						_ = k.h.WithGroup("g").Handle(context.Background(), rec(slog.LevelInfo, msg, s("service_id", "svc-9")))
					}
				}
			}()
		}
		close(start)
		gruppe.Wait()
		zeilen := senke.zeilen()
		if len(zeilen) != goroutinen*runden {
			t.Fatalf("Länge %d: %d Zeilen, erwartet %d", laenge, len(zeilen), goroutinen*runden)
		}
		eltern := k.erwartet()
		for _, z := range zeilen {
			g := lies(t, z)
			nr, _ := strconv.Atoi(strings.TrimPrefix(g.text["msg"], "g"))
			e := erwartung{felder: map[string]string{}}
			for key, v := range eltern.felder {
				e.felder[key] = v
			}
			if nr%2 == 0 {
				e.felder["befunde"] = strconv.Itoa(nr)
			}
			pruefeZeile(t, "Länge "+strconv.Itoa(laenge)+", "+g.text["msg"], z, "INFO", g.text["msg"], e)
		}
	}
}

// H11: keine Sperre liegt über Clock.Now(), Seite Handle. Menge M16: die Uhr ruft beim ersten Now
// CountError, Snapshot, Log oder Handle eines abgeleiteten Handlers; der Aufruf von Handle kehrt
// innerhalb 2 s zurück, bei Log und Handle stehen zwei Zeilen da, die innere zuerst.
func TestUhrRuftHandle(t *testing.T) {
	rueckrufe := []struct {
		name   string
		f      func(a *telemetry.Adapter)
		zeilen int
	}{
		{"CountError", func(a *telemetry.Adapter) { a.CountError(model.CodeE2001) }, 1},
		{"Snapshot", func(a *telemetry.Adapter) { _ = a.Snapshot() }, 1},
		{"Log", func(a *telemetry.Adapter) { a.Log(context.Background(), slog.LevelInfo, "innen") }, 2},
		{"Handle", func(a *telemetry.Adapter) {
			_ = a.Handler().WithGroup("g").Handle(context.Background(), rec(slog.LevelInfo, "innen"))
		}, 2},
	}
	if len(rueckrufe) != 4 {
		t.Fatalf("%d Rückrufe, erwartet 4", len(rueckrufe))
	}
	for _, r := range rueckrufe {
		senke := &FakeSenke{}
		uhr := &RueckrufUhr{t: festeUhr().t}
		a := neuMit(t, senke, uhr, slog.LevelInfo)
		uhr.rueckruf = func() { r.f(a) }
		h := a.Handler().WithAttrs([]slog.Attr{s("session_id", "v1")})
		if !binnen(t, 2*time.Second, r.name, func() { _ = h.Handle(context.Background(), rec(slog.LevelInfo, "aussen")) }) {
			t.Errorf("Rückruf %s in Now: Handle kehrt nicht zurück", r.name)
			continue
		}
		z := senke.zeilen()
		if len(z) != r.zeilen {
			t.Errorf("Rückruf %s: %d Zeilen, erwartet %d", r.name, len(z), r.zeilen)
		}
		if r.zeilen == 2 && (lies(t, z[0]).text["msg"] != "innen" || lies(t, z[1]).text["msg"] != "aussen") {
			t.Errorf("Rückruf %s: Zeilen %q, erwartet innen, dann aussen", r.name, z)
		}
	}
}

// H13: Zähler, Snapshot und das Zählen von Verwürfen warten nicht auf eine blockierte Senke, Seite
// Handle. Haltepunkt am Writer: A hängt in Handle der Wurzel; B ruft CountError, Snapshot und Handle
// jedes der 15 Handler aus ableitungen (Menge M17 in jeder Tiefe bis 3) unter MinLevel mit einem
// ungültigen Attribut des Aufrufs (und den ungültigen vorgelegten Attributen der Kette); jede kehrt
// innerhalb 2 s zurück, die Verwürfe wachsen um die Summe des Modells der Ketten.
func TestBlockierteSenkeHandle(t *testing.T) {
	senke, a := blockierendH(t)
	alle := ableitungen(t, a)
	fertigA := handleInGoroutine(alle[0].h, "a")
	if !senke.warteEintritt(5 * time.Second) {
		t.Fatal("A erreicht Write nicht")
	}
	ops := []struct {
		name string
		f    func()
	}{
		{"CountError", func() { a.CountError(model.CodeE2001) }},
		{"Snapshot", func() { _ = a.Snapshot() }},
	}
	var schluessel, wert uint64
	for _, k := range alle {
		ops = append(ops, struct {
			name string
			f    func()
		}{"Handle unter MinLevel, Handler " + k.name, func() {
			_ = k.h.Handle(context.Background(), rec(slog.LevelDebug, "m", s("mode", "a b")))
		}})
		e := k.erwartet(s("mode", "a b"))
		schluessel, wert = schluessel+e.schluessel, wert+e.wert
	}
	vorher := a.Snapshot()
	for _, o := range ops {
		if !binnen(t, 2*time.Second, o.name, o.f) {
			t.Fatalf("%s wartet auf die blockierte Senke", o.name)
		}
	}
	stand := a.Snapshot()
	if stand.Errors[model.CodeE2001] != 1 || stand.Dropped[support.VerwurfWert]-vorher.Dropped[support.VerwurfWert] != wert ||
		stand.Dropped[support.VerwurfSchluessel]-vorher.Dropped[support.VerwurfSchluessel] != schluessel {
		t.Errorf("CountError %d, wert %d, schluessel %d; erwartet 1, %d und %d", stand.Errors[model.CodeE2001],
			stand.Dropped[support.VerwurfWert], stand.Dropped[support.VerwurfSchluessel], wert, schluessel)
	}
	if zurueck(fertigA, 100*time.Millisecond) {
		t.Fatal("A kehrt vor der Freigabe zurück")
	}
	senke.freigeben()
	if !zurueck(fertigA, 2*time.Second) || len(senke.zeilen()) != 1 {
		t.Errorf("A nach der Freigabe: %d Zeilen, erwartet 1", len(senke.zeilen()))
	}
}

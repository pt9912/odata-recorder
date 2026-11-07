package telemetry_test

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// blockierend liefert einen Adapter über einer blockierenden Senke; die Freigabe läuft beim Ende des Tests.
func blockierend(t *testing.T) (*FakeSenke, support.Telemetry, func() support.Snapshot) {
	t.Helper()
	senke := &FakeSenke{}
	senke.blockieren()
	t.Cleanup(senke.freigeben)
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	return senke, a, a.Snapshot
}

// logInGoroutine ruft Log in einer Goroutine; fertig wird geschlossen, wenn Log zurückkehrt.
func logInGoroutine(a support.Telemetry, id string) <-chan struct{} {
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		a.Log(context.Background(), slog.LevelInfo, "m", support.Str("session_id", id))
	}()
	return fertig
}

func zurueck(c <-chan struct{}, d time.Duration) bool {
	select {
	case <-c:
		return true
	case <-time.After(d):
		return false
	}
}

// L11: Zähler, Snapshot und CountError warten nicht auf eine blockierte Senke (SPEC-232 Nr. 5 g).
// Haltepunkt am Writer: A hängt im Write; B ruft CountError, Snapshot und ein Log unter MinLevel mit
// ungültigem Attribut, jedes kehrt innerhalb 2 s zurück; nach der Freigabe kehrt A zurück, eine Zeile.
func TestBlockierteSenke(t *testing.T) {
	senke, a, snapshot := blockierend(t)
	fertigA := logInGoroutine(a, "a1")
	if !senke.warteEintritt(5 * time.Second) {
		t.Fatal("A erreicht Write nicht")
	}
	ops := []struct {
		name string
		f    func()
	}{
		{"CountError", func() { a.CountError(model.CodeE2001) }},
		{"Snapshot", func() { _ = snapshot() }},
		{"Log unter MinLevel mit ungültigem Attribut", func() {
			a.Log(context.Background(), slog.LevelDebug, "m", support.Str("mode", "a b"))
		}},
	}
	for _, o := range ops {
		if !binnen(t, 2*time.Second, o.name, o.f) {
			t.Fatalf("%s wartet auf die blockierte Senke", o.name)
		}
	}
	stand := snapshot()
	if stand.Errors[model.CodeE2001] != 1 || stand.Dropped[support.VerwurfWert] != 1 {
		t.Errorf("CountError %d, wert %d; erwartet je 1", stand.Errors[model.CodeE2001], stand.Dropped[support.VerwurfWert])
	}
	if zurueck(fertigA, 100*time.Millisecond) {
		t.Fatal("A kehrt vor der Freigabe zurück")
	}
	senke.freigeben()
	if !zurueck(fertigA, 2*time.Second) || len(senke.zeilen()) != 1 {
		t.Errorf("A nach der Freigabe: %d Zeilen, erwartet 1", len(senke.zeilen()))
	}
}

// L17 (1): Haltepunkt: A hängt im Write, B ruft Log; das zweite Write darf innerhalb 200 ms nicht
// eintreten, nach der Freigabe stehen beide Zeilen da, nie zwei Write zugleich.
func TestKeineZweiWriteHaltepunkt(t *testing.T) {
	senke, a, _ := blockierend(t)
	fertigA := logInGoroutine(a, "a1")
	if !senke.warteEintritt(5 * time.Second) {
		t.Fatal("A erreicht Write nicht")
	}
	fertigB := logInGoroutine(a, "b1")
	if senke.warteEintritt(200 * time.Millisecond) {
		t.Error("ein zweites Write trat ein, während das erste hing")
	}
	senke.freigeben()
	if !zurueck(fertigA, 2*time.Second) || !zurueck(fertigB, 2*time.Second) {
		t.Fatal("ein Log kehrt nach der Freigabe nicht zurück")
	}
	if len(senke.zeilen()) != 2 || senke.hoechstwert() != 1 {
		t.Errorf("%d Zeilen, Höchstwert gleichzeitiger Write %d; erwartet 2 und 1", len(senke.zeilen()), senke.hoechstwert())
	}
}

// masse lässt goroutinen Goroutinen je runden Aufrufe von Log mit eigener session_id und aufsteigendem
// befunde schreiben und liefert die Senke.
func masse(t *testing.T, goroutinen, runden int) *FakeSenke {
	t.Helper()
	senke := &FakeSenke{}
	a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for g := range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for n := 1; n <= runden; n++ {
				a.Log(context.Background(), slog.LevelInfo, "m", support.Str("session_id", "g"+strconv.Itoa(g)), support.Int("befunde", int64(n)))
			}
		}()
	}
	close(start)
	wg.Wait()
	return senke
}

// L17 (2): Masse: 32 Goroutinen × 200 Einträge ergeben 6400 gültige Zeilen, nie zwei Write zugleich.
func TestKeineZweiWriteMasse(t *testing.T) {
	const goroutinen, runden = 32, 200
	senke := masse(t, goroutinen, runden)
	zeilen := senke.zeilen()
	if len(zeilen) != goroutinen*runden {
		t.Fatalf("%d Zeilen, erwartet %d", len(zeilen), goroutinen*runden)
	}
	for _, z := range zeilen {
		if g := lies(t, z); len(g.schluessel) != 5 {
			t.Fatalf("Zeile %q vermischt oder unvollständig", z)
		}
	}
	if senke.hoechstwert() != 1 {
		t.Errorf("Höchstwert gleichzeitiger Write %d, erwartet 1", senke.hoechstwert())
	}
}

// L18 (1): Log kehrt nicht zurück, solange Write blockiert (100 ms Beobachtung), danach sofort.
func TestLogKehrtNachWriteZurueck(t *testing.T) {
	senke, a, _ := blockierend(t)
	fertig := logInGoroutine(a, "a1")
	if !senke.warteEintritt(5 * time.Second) {
		t.Fatal("kein Write")
	}
	if zurueck(fertig, 100*time.Millisecond) {
		t.Error("Log kehrte zurück, während Write blockierte")
	}
	senke.freigeben()
	if !zurueck(fertig, 2*time.Second) {
		t.Error("Log kehrt nach der Freigabe nicht zurück")
	}
}

// L18 (2): Masse: je session_id stehen befunde 1 bis 200 in der Senke streng aufsteigend.
func TestZeilenReihenfolgeMasse(t *testing.T) {
	const goroutinen, runden = 32, 200
	letzte := map[string]int{}
	for _, z := range masse(t, goroutinen, runden).zeilen() {
		g := lies(t, z)
		n, err := strconv.Atoi(g.text["befunde"])
		if err != nil || n != letzte[g.text["session_id"]]+1 {
			t.Fatalf("session_id %s: befunde %q nach %d", g.text["session_id"], g.text["befunde"], letzte[g.text["session_id"]])
		}
		letzte[g.text["session_id"]] = n
	}
	if len(letzte) != goroutinen {
		t.Errorf("%d session_id, erwartet %d", len(letzte), goroutinen)
	}
	for s, n := range letzte {
		if n != runden {
			t.Errorf("%s endet bei %d, erwartet %d", s, n, runden)
		}
	}
}

// uhrMitRueckruf liefert einen Adapter, dessen Uhr beim ersten Now rueckruf(a) ruft.
func uhrMitRueckruf(t *testing.T, rueckruf func(a support.Telemetry, ts func() support.Snapshot)) (*FakeSenke, support.Telemetry) {
	t.Helper()
	senke := &FakeSenke{}
	uhr := &RueckrufUhr{t: festeUhr().t}
	a := neuMit(t, senke, uhr, slog.LevelInfo)
	uhr.rueckruf = func() { rueckruf(a, a.Snapshot) }
	return senke, a
}

// L19: Menge M16 (CountError, Snapshot): ruft die Uhr in Now den Zustand, kehrt Log innerhalb 2 s
// zurück; keine Sperre des Zustands liegt über Now.
func TestUhrRuftZustand(t *testing.T) {
	rueckrufe := map[string]func(a support.Telemetry, ts func() support.Snapshot){
		"CountError": func(a support.Telemetry, _ func() support.Snapshot) { a.CountError(model.CodeE2001) },
		"Snapshot":   func(_ support.Telemetry, ts func() support.Snapshot) { _ = ts() },
	}
	for name, r := range rueckrufe {
		senke, a := uhrMitRueckruf(t, r)
		if !binnen(t, 2*time.Second, name, func() { a.Log(context.Background(), slog.LevelInfo, "m") }) {
			t.Errorf("Rückruf %s in Now: Log kehrt nicht zurück", name)
			continue
		}
		if len(senke.zeilen()) != 1 {
			t.Errorf("Rückruf %s: %d Zeilen, erwartet 1", name, len(senke.zeilen()))
		}
	}
	if len(rueckrufe) != 2 {
		t.Fatalf("%d Rückrufe, erwartet 2", len(rueckrufe))
	}
}

// L20: Menge M16 (Log): ruft die Uhr in Now selbst Log, kehrt der Aufruf innerhalb 2 s zurück und es
// stehen zwei Zeilen da, die innere zuerst; keine Sperre der Zeile liegt über Now.
func TestUhrRuftZeile(t *testing.T) {
	senke, a := uhrMitRueckruf(t, func(a support.Telemetry, _ func() support.Snapshot) {
		a.Log(context.Background(), slog.LevelInfo, "innen")
	})
	if !binnen(t, 2*time.Second, "Rückruf Log", func() { a.Log(context.Background(), slog.LevelInfo, "aussen") }) {
		t.Fatal("Log kehrt nicht zurück, wenn Now Log ruft")
	}
	z := senke.zeilen()
	if len(z) != 2 || lies(t, z[0]).text["msg"] != "innen" || lies(t, z[1]).text["msg"] != "aussen" {
		t.Errorf("Zeilen %q, erwartet innen, dann aussen", z)
	}
}

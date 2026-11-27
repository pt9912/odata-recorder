package admin_test

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Dieser Teil hält die nebenläufigen Zusagen von SPEC-301 Nr. 2 und 5 d, 5 e: die Zulassung ist ein Schritt unter einer
// Sperre, und die Sperre liegt nicht über Append.

// abweisenZugleich sendet n Abweisungen aus n Goroutinen hinter einem Startsignal und wartet höchstens 10 s auf sie.
func abweisenZugleich(t *testing.T, u *umgebung, n int) {
	t.Helper()
	start, fertig := make(chan struct{}), make(chan struct{})
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			<-start
			u.srv.ServeHTTP(rec, ohneAuth("GET", "/ok", nil))
			if rec.Code != 401 {
				t.Errorf("Status %d, erwartet 401", rec.Code)
			}
		}()
	}
	close(start)
	go func() { wg.Wait(); close(fertig) }()
	select {
	case <-fertig:
	case <-time.After(10 * time.Second):
		t.Fatalf("Frist von 10 s abgelaufen: %d gleichzeitige Abweisungen", n)
	}
}

// TestFensterGleichzeitig hält W10 (Nr. 5 d): in jedem von 1200 Fenstern ergeben 64 gleichzeitige Abweisungen genau 20
// Einträge, und CountError läuft je Abweisung. Die Zahl der Fenster hält die Wirkung auch ohne Detektor in jedem Prozess.
func TestFensterGleichzeitig(t *testing.T) {
	const goroutinen, fenster = 64, 1200
	uhr := &stellUhr{stand: fensterT0()}
	u, _ := fensterServer(t, uhr)
	for f := range fenster {
		uhr.vorruecken(60 * time.Second)
		vorher := zahlEintraege(u)
		abweisenZugleich(t, u, goroutinen)
		if n := zahlEintraege(u) - vorher; n != 20 {
			t.Fatalf("Fenster %d: %d Einträge, erwartet 20", f+1, n)
		}
	}
	if codes, _ := zaehlungenVon(u.tele); len(codes) != goroutinen*fenster {
		t.Errorf("CountError %d-mal, erwartet %d", len(codes), goroutinen*fenster)
	}
}

// TestFensterZulassungAtomar hält W11 (Nr. 2, 5 d, 6): solange der erste Request in Now() hält, ruft ein zweiter die Uhr
// nicht und schreibt keinen Eintrag; nach der Freigabe sind von 64 Abweisungen genau 20 zugelassen.
func TestFensterZulassungAtomar(t *testing.T) {
	haelt, frei := make(chan struct{}), make(chan struct{})
	var einmal sync.Once
	gibFrei := func() { einmal.Do(func() { close(frei) }) }
	t.Cleanup(gibFrei)
	weitere := make(chan struct{}, 128)
	uhr := &stellUhr{stand: fensterT0()}
	uhr.vor = func(n int) {
		if n == 1 {
			close(haelt)
			<-frei
			return
		}
		select {
		case weitere <- struct{}{}:
		default:
		}
	}
	u, _ := fensterServer(t, uhr)
	fertig := make(chan struct{}, 2)
	sende := func() {
		rec := httptest.NewRecorder()
		u.srv.ServeHTTP(rec, ohneAuth("GET", "/ok", nil))
		fertig <- struct{}{}
	}
	go sende()
	select {
	case <-haelt:
	case <-time.After(5 * time.Second):
		t.Fatal("der erste Request erreichte die Uhr nicht")
	}
	go sende()
	ende := time.Now().Add(5 * time.Second)
	for k, _ := u.ids.zahlen(); k < 2; k, _ = u.ids.zahlen() {
		if time.Now().After(ende) {
			t.Fatal("der zweite Request begann nicht")
		}
		time.Sleep(time.Millisecond)
	}
	select {
	case <-weitere:
		t.Error("der zweite Request rief die Uhr, solange der erste in Now() hält")
	case <-time.After(200 * time.Millisecond):
	}
	if n := u.audit.Aufrufe(); n != 0 {
		t.Errorf("%d Einträge, solange der erste in Now() hält, erwartet 0", n)
	}
	gibFrei()
	for range 2 {
		select {
		case <-fertig:
		case <-time.After(5 * time.Second):
			t.Fatal("ein Request endete nach der Freigabe nicht")
		}
	}
	abweisenZugleich(t, u, 62)
	gleichAuf(t, "Einträge nach 64 Abweisungen", zahlEintraege(u), 20)
}

// TestFensterOhneSperreUeberAppend hält W12 (Nr. 2, 5 e): solange Append der ersten Abweisung hängt, wird eine zweite
// zugelassen, ihr Append läuft, und sie ist beantwortet.
func TestFensterOhneSperreUeberAppend(t *testing.T) {
	begonnen, frei := make(chan struct{}), make(chan struct{})
	var einmal, erster sync.Once
	gibFrei := func() { einmal.Do(func() { close(frei) }) }
	t.Cleanup(gibFrei)
	u, _ := fensterServer(t, &stellUhr{stand: fensterT0()})
	u.audit.Vor = func(support.Eintrag) {
		haelt := false
		erster.Do(func() { haelt = true })
		if haelt {
			close(begonnen)
			<-frei
		}
	}
	antwort := func() <-chan *httptest.ResponseRecorder {
		c := make(chan *httptest.ResponseRecorder, 1)
		go func() {
			rec := httptest.NewRecorder()
			u.srv.ServeHTTP(rec, ohneAuth("GET", "/ok", nil))
			c <- rec
		}()
		return c
	}
	a := antwort()
	select {
	case <-begonnen:
	case <-time.After(5 * time.Second):
		t.Fatal("Append der ersten Abweisung begann nicht")
	}
	select {
	case rec := <-antwort():
		pruefe401OhneKennung(t, "zweite Abweisung", rec)
		gleichAuf(t, "Aufrufe von Append, solange der erste hängt", u.audit.Aufrufe(), 2)
	case <-time.After(5 * time.Second):
		t.Error("die zweite Abweisung wartet auf Append der ersten")
	}
	gibFrei()
	select {
	case rec := <-a:
		pruefe401OhneKennung(t, "erste Abweisung", rec)
	case <-time.After(5 * time.Second):
		t.Fatal("die erste Abweisung endete nach der Freigabe nicht")
	}
}

package hybrid_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

const aufrufe = 64

// gleichzeitig startet n Aufrufe auf einmal und wartet auf alle.
func gleichzeitig(n int, f func(id int)) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	for id := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			f(id)
		}()
	}
	close(start)
	wg.Wait()
}

func idVon(a inbound.Anfrage) int {
	id, _ := strconv.Atoi(a.KorrelationsID)
	return id
}

// koerper liefert den Body des Aufrufs id: Länge und Bytes sind je Aufruf verschieden.
func koerper(id int, kind int) []byte {
	if kind == 4 {
		return muster(standardMax+1+id%5, id)
	}
	return muster(1+id%standardMax, id)
}

// TestGleichzeitig hält Z31: 64 gleichzeitige Aufrufe, gemischt über Erfolg, Replay mit Puffer, Replay-Fehler, Fehler
// ohne Fallback und Überlauf, erhalten je ihre eigene Antwort und ihre eigenen Bytes; die Summen der Aufrufe der
// Stufen stimmen mit der Erwartung.
func TestGleichzeitig(t *testing.T) {
	g := neuesRig(t, offenste(t), standardMax)
	e6003 := codeNamed(t, "ODR-E6003")
	e6005 := codeNamed(t, "ODR-E6005")
	miss := codeNamed(t, "ODR-E5001")
	g.m.fn = func(_ int, _ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		id := idVon(a)
		kind := id % 5
		if a.Body != nil {
			if got, _ := io.ReadAll(a.Body); !bytes.Equal(got, koerper(id, kind)) {
				t.Errorf("Aufruf %d: M erhielt fremde Bytes", id)
			}
		}
		switch kind {
		case 0:
			return inbound.Antwort{Status: 200, Body: neuerStrom([]byte("m" + strconv.Itoa(id)))}, nil
		case 3:
			return inbound.Antwort{}, &inbound.Fehler{Code: e6005, Status: id}
		}
		return inbound.Antwort{}, &inbound.Fehler{Code: e6003, Status: id}
	}
	g.r.fn = func(_ int, _ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		id := idVon(a)
		if id%5 == 2 {
			return inbound.Antwort{}, &inbound.Fehler{Code: miss, Status: id}
		}
		echo, _ := io.ReadAll(a.Body)
		return inbound.Antwort{Status: 200, Body: neuerStrom(echo)}, nil
	}
	gleichzeitig(aufrufe, func(id int) {
		kind := id % 5
		a := anfrage("POST")
		a.KorrelationsID = strconv.Itoa(id)
		b := koerper(id, kind)
		a.Body, a.Laenge = bytes.NewReader(b), int64(len(b))
		if id%2 == 1 {
			a.Laenge = -1
		}
		if kind == 0 {
			a.Body, a.Laenge = nil, 0
		}
		got, err := g.v.Hybrid(context.Background(), a)
		var fe *inbound.Fehler
		switch kind {
		case 0:
			if err != nil || readAll(got) != "m"+strconv.Itoa(id) || len(kennzeichen(got)) != 0 {
				t.Errorf("Aufruf %d: Erfolg verfälscht: %v", id, err)
			}
		case 1, 4:
			if kind == 1 && (err != nil || !bytes.Equal([]byte(readAll(got)), b) || len(kennzeichen(got)) != 1) {
				t.Errorf("Aufruf %d: Replay verfälscht: %v", id, err)
			}
			if kind == 4 && (!errors.As(err, &fe) || fe.Status != id) {
				t.Errorf("Aufruf %d: Überlauf: Fehler %v", id, err)
			}
		default:
			if !errors.As(err, &fe) || fe.Status != id {
				t.Errorf("Aufruf %d: Fehler einer anderen Anfrage: %v", id, err)
			}
		}
	})
	// Erwartung: M je Aufruf einmal; R bei den Arten 1 und 2.
	wantR := 0
	for id := range aufrufe {
		if k := id % 5; k == 1 || k == 2 {
			wantR++
		}
	}
	g.zaehlen(t, "gleichzeitig", aufrufe, wantR)
}

func readAll(a inbound.Antwort) string {
	if a.Body == nil {
		return ""
	}
	b, _ := io.ReadAll(a.Body)
	return string(b)
}

// TestGleichzeitigerAbbruch hält Z32: 64 Aufrufe, davon die Hälfte mit beendetem ctx mitten in M oder R oder nach M:
// jedes Ergebnis gehört zum eigenen Aufruf, der Strom einer verworfenen Antwort wird genau einmal geschlossen, der einer
// ausgelieferten nie, und R läuft nur, wo der Ablauf es verlangt.
func TestGleichzeitigerAbbruch(t *testing.T) {
	g := neuesRig(t, offenste(t), standardMax)
	e6003 := codeNamed(t, "ODR-E6003")
	cancels := make([]context.CancelFunc, aufrufe)
	sm, sr := make([]*strom, aufrufe), make([]*strom, aufrufe)
	for id := range aufrufe {
		sm[id], sr[id] = neuerStrom([]byte("m")), neuerStrom([]byte("r"))
	}
	g.m.fn = func(_ int, ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		id := idVon(a)
		switch id % 4 {
		case 0:
			cancels[id]()
			return inbound.Antwort{Status: 500, Body: sm[id]}, errors.Join(&inbound.Fehler{Code: e6003, Status: id}, ctx.Err())
		case 2:
			cancels[id]()
			return inbound.Antwort{Status: 501, Body: sm[id]}, &inbound.Fehler{Code: e6003, Status: id}
		}
		return inbound.Antwort{}, &inbound.Fehler{Code: e6003, Status: id}
	}
	g.r.fn = func(_ int, ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		id := idVon(a)
		if id%4 == 1 {
			cancels[id]()
			return inbound.Antwort{Status: 500, Body: sr[id]}, fmt.Errorf("r%d: %w", id, ctx.Err())
		}
		return inbound.Antwort{Status: 200, Body: sr[id]}, nil
	}
	gleichzeitig(aufrufe, func(id int) {
		a := anfrage("GET")
		a.KorrelationsID = strconv.Itoa(id)
		ctx, cancel := abbruchKontext()
		defer cancel()
		cancels[id] = cancel
		got, err := g.v.Hybrid(ctx, a)
		var fe *inbound.Fehler
		switch id % 4 {
		case 0:
			if !errors.Is(err, context.Canceled) || !errors.As(err, &fe) || fe.Status != id {
				t.Errorf("Aufruf %d: Fehler %v gehört nicht zum Aufruf", id, err)
			}
		case 1:
			if want := fmt.Sprintf("r%d: context canceled", id); err == nil || err.Error() != want {
				t.Errorf("Aufruf %d: Fehler %v, erwartet %q", id, err, want)
			}
		case 2:
			if err != context.Canceled {
				t.Errorf("Aufruf %d: Fehler %v, erwartet context.Canceled", id, err)
			}
		case 3:
			if err != nil || got.Body != io.ReadCloser(sr[id]) {
				t.Errorf("Aufruf %d: Antwort %+v, Fehler %v", id, got, err)
			}
		}
	})
	verworfenM, verworfenR := 0, 0
	for id := range aufrufe {
		wantM, wantR := 0, 0
		switch id % 4 {
		case 0, 2:
			wantM = 1
		case 1:
			wantR = 1
		}
		if id%4 == 0 || id%4 == 2 {
			verworfenM++
		}
		if id%4 == 1 {
			verworfenR++
		}
		if int(sm[id].closes.Load()) != wantM || int(sr[id].closes.Load()) != wantR {
			t.Errorf("Aufruf %d: Close von M %d (erwartet %d), von R %d (erwartet %d)", id, sm[id].closes.Load(), wantM, sr[id].closes.Load(), wantR)
		}
	}
	wantReplays := 0
	for id := range aufrufe {
		if id%4 == 1 || id%4 == 3 {
			wantReplays++
		}
	}
	g.zaehlen(t, "Abbruch", aufrufe, wantReplays)
	if verworfenM != 32 || verworfenR != 16 {
		t.Errorf("Zählprüfung: %d verworfene Antworten von M, %d von R", verworfenM, verworfenR)
	}
}

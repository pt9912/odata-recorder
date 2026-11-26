package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"
)

// anfrageFrist ist die Frist je Anfrage der Last, Body eingeschlossen (SPEC-216 Nr. 8 c).
const anfrageFrist = 10 * time.Second

// taktStand ist, was der Takt eines Laufs zählt (SPEC-216 Nr. 8 b): die gestarteten Anfragen, die größte Zahl gleichzeitig offener und die
// größte Verspätung eines Starts gegen seine Fälligkeit.
type taktStand struct {
	Gestartet       int
	GleichzeitigMax int
	VerspaetungMax  time.Duration
}

// fahreTakt startet anzahl Aufrufe von sende, jeden in einer eigenen Goroutine: der k-te (ab 0) ist zur Zeit t0 + k × 1 s ÷ rate fällig,
// gerechnet aus k und t0 und nie aus dem Ende eines früheren Aufrufs (SPEC-216 Nr. 8 a). Er wartet auf keinen früheren Aufruf, kehrt erst
// zurück, wenn jeder gestartete zurückgekehrt ist, und startet nach dem Ende von ctx keinen mehr.
func fahreTakt(ctx context.Context, uhr Uhr, rate, anzahl int, sende func(ctx context.Context, k int)) taktStand {
	t0 := uhr.Jetzt()
	var (
		mu    sync.Mutex
		offen int
		stand taktStand
		wg    sync.WaitGroup
	)
	for k := range anzahl {
		faellig := t0.Add(time.Duration(k) * time.Second / time.Duration(rate))
		if uhr.WarteBis(ctx, faellig) != nil || ctx.Err() != nil {
			break
		}
		stand.Gestartet++
		wg.Add(1)
		go func() {
			defer wg.Done()
			spaet := uhr.Jetzt().Sub(faellig)
			mu.Lock()
			offen++
			stand.GleichzeitigMax = max(stand.GleichzeitigMax, offen)
			stand.VerspaetungMax = max(stand.VerspaetungMax, spaet)
			mu.Unlock()
			sende(ctx, k)
			mu.Lock()
			offen--
			mu.Unlock()
		}()
	}
	wg.Wait()
	return stand
}

// lastClient ist der Client der Last (SPEC-216 Nr. 8 b, d): Keep-Alive, keine Obergrenze der Verbindungen und der ruhenden Verbindungen,
// keine Weiterleitung. rt nil ist ein eigener Transport.
func lastClient(rt http.RoundTripper) *http.Client {
	if rt == nil {
		rt = &http.Transport{MaxIdleConnsPerHost: math.MaxInt32}
	}
	return &http.Client{Transport: rt, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

// leererBody ist ein leerer Body ohne GetBody: der Transport sendet damit dieselben Bytes wie ohne Body und wiederholt die Anfrage nach
// einem Fehler der Verbindung nicht (SPEC-216 Nr. 8 d).
type leererBody struct{}

func (leererBody) Read([]byte) (int, error) { return 0, io.EOF }

func (leererBody) Close() error { return nil }

// holeLast sendet eine Anfrage der Last und liest den Body vollständig (SPEC-216 Nr. 8 c): gelesen sind die Bytes des Bodys, auch bei einem
// Fehler. Ein Fehler ist ein Fehler des Transports, das Ende der anfrageFrist, ein Lesefehler des Bodys oder ein Status ungleich 200.
func holeLast(ctx context.Context, uhr Uhr, c *http.Client, ziel string, header [][2]string) (gelesen int64, err error) {
	ctx, cancel := mitFrist(ctx, uhr, anfrageFrist)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ziel, nil)
	if err != nil {
		return 0, err
	}
	for _, h := range header {
		req.Header.Set(h[0], h[1])
	}
	req.Body = leererBody{}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	gelesen, err = io.Copy(io.Discard, resp.Body)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Status %d", resp.StatusCode)
	}
	return gelesen, err
}

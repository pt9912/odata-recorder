package main

import (
	"context"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// starteBinaryUmgebung startet ein Handle mit der Umgebung des Binary (umgebung) über dir und liefert seine Adresse; Cleanup fährt es herunter.
// Der Port 0 ist für LISTEN kein gültiger Wert der Konfiguration, die Prüfung bekommt 8080, das Handle einen freien Port.
func starteBinaryUmgebung(t *testing.T, b Betriebsart, dir string) string {
	t.Helper()
	ctx := context.Background()
	env := umgebung(b, "127.0.0.1:8080")
	env[slices.Index(env, "DATA_DIR=/daten")] = "DATA_DIR=" + dir
	k, err := bootstrap.LadeKonfiguration(env)
	if err != nil {
		t.Fatalf("Umgebung des Binary für %s: %v", b, err)
	}
	k.Listen = "127.0.0.1:0"
	h, err := baueEcht(ctx, k)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(ctx) })
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatal(err)
	}
	return h.Adresse()
}

// listenTreffen prüft, dass jede Anfrage der Listen replay und mock des Profils rauch in einem Handle der Betriebsart über die Daten des Aufbaus
// und mit der Umgebung des Binary Status 200 trifft und ein Body höchstens 64 KiB trägt (SPEC-216 Nr. 6 b, 7 und 9 a); die Listen sind damit gegen den Dienst geprüft, nicht gegen
// ein Abbild. Der Aufruf gehört zu TestAufbauRauch, der den Aufbau liefert.
func listenTreffen(t *testing.T, dir string, erg aufbauErgebnis) {
	t.Helper()
	listen := map[Betriebsart][]Anfrage{
		betriebReplay: listeReplay(rauch(), 1, 1, erg.Workspace, erg.Session),
		betriebMock:   listeMock(rauch(), 1, erg.Workspace),
	}
	for _, b := range []Betriebsart{betriebReplay, betriebMock} {
		t.Run(string(b), func(t *testing.T) {
			adresse := starteBinaryUmgebung(t, b, dir)
			falsch := 0
			for i, a := range listen[b] {
				req, err := http.NewRequest(http.MethodGet, "http://"+adresse+a.Ziel(), nil)
				if err != nil {
					t.Fatal(err)
				}
				for _, h := range a.Header {
					req.Header.Set(h[0], h[1])
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				laenge, _ := io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode != http.StatusOK || laenge > 65536 {
					if falsch++; falsch <= 5 {
						t.Errorf("Anfrage %d %s: Status %d, %d Bytes, erwartet 200 und höchstens 65536", i, a.Ziel(), resp.StatusCode, laenge)
					}
				}
			}
			if falsch > 0 || len(listen[b]) != 260 {
				t.Errorf("%d von %d Anfragen nicht 200", falsch, len(listen[b]))
			}
		})
	}
}

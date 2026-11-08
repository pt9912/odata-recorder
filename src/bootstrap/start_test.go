package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// umgebungFuer liefert eine gültige Umgebung mit n Services (a, b, …) im Modus modus; DATA_DIR liegt
// unter dem Verzeichnis des Tests.
func umgebungFuer(t *testing.T, modus string, n int, extra ...string) []string {
	t.Helper()
	dir := t.TempDir()
	eintraege := make([]string, n)
	for i := range eintraege {
		id := string(rune('a' + i))
		eintraege[i] = fmt.Sprintf(`{"id":%q,"service_root":"/%s","upstream_url":"http://127.0.0.1:9/%s"}`, id, id, id)
	}
	datei := filepath.Join(dir, "dienste.json")
	if err := os.WriteFile(datei, []byte("["+strings.Join(eintraege, ",")+"]"), 0o600); err != nil {
		t.Fatal(err)
	}
	return append([]string{
		"MODE=" + modus, "LISTEN=127.0.0.1:8080", "DATA_DIR=" + filepath.Join(dir, "d"),
		"SERVICES_FILE=" + datei, "SHUTDOWN_TIMEOUT=2s",
	}, extra...)
}

// fahre ruft Start mit den Haken; sobald "start: bereit" im Log steht, bricht es ctx ab. Es liefert
// Rückgabe und stderr.
func fahre(ctx context.Context, t *testing.T, cancel context.CancelFunc, env []string, h *haken) (int, string) {
	t.Helper()
	puffer := &syncPuffer{}
	ergebnis := make(chan int, 1)
	go func() { ergebnis <- bootstrap.StartMit(ctx, env, puffer, h.bauteile()) }()
	frist := time.After(20 * time.Second)
	for {
		select {
		case code := <-ergebnis:
			return code, puffer.String()
		case <-frist:
			t.Fatalf("Start kehrt nicht zurück; stderr:\n%s", puffer.String())
		case <-time.After(2 * time.Millisecond):
			if strings.Contains(puffer.String(), "start: bereit") {
				cancel()
			}
		}
	}
}

// nurStart lässt aus dem Protokoll die Ereignisse des Herunterfahrens weg.
func nurStart(ev []string) []string {
	var aus []string
	for _, e := range ev {
		if !slices.Contains([]string{"anhalten", pCloseAll, pSchliessen, "serve-ende", pAblageZu}, e) {
			aus = append(aus, e)
		}
	}
	return aus
}

func meldungsZeilen(stderr string) []string {
	var z []string
	for _, l := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(l, "ODR-") {
			z = append(z, l)
		}
	}
	return z
}

// TestStartReihenfolge fängt B1: die Ereignisse des Starts folgen in beiden Modi der Reihenfolge von
// SPEC-139 Nr. 6; die Services stehen vor den ersten Sessions, je in der Reihenfolge der Konfiguration.
func TestStartReihenfolge(t *testing.T) {
	erwartet := map[string][]string{
		modusRecord: {pAblage, pWartung, "dienst:a", "dienst:b", "session:a", "session:b", pListen, pServe},
		modusReplay: {pAblage, pWartung, "dienst:a", "dienst:b", pListen, pServe},
	}
	for _, modus := range modi() {
		t.Run(modus, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			h := neuHaken()
			code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modus, 2), h)
			if code != 0 {
				t.Fatalf("Rückgabe %d, stderr:\n%s", code, stderr)
			}
			if got := nurStart(h.p.liste()); !slices.Equal(got, erwartet[modus]) {
				t.Errorf("Reihenfolge %v, erwartet %v", got, erwartet[modus])
			}
			// Der Eintrag „bereit“ (SPEC-139 Nr. 8) trägt Modus und die Adresse des gebundenen Listeners.
			soll := map[string]string{"level": "INFO", "msg": "start: bereit", "mode": modus, "listen": h.listener[0].Addr().String()}
			var bereit []map[string]string
			for _, z := range logZeilen(t, stderr) {
				if z["msg"] == "start: bereit" {
					bereit = append(bereit, z)
				}
			}
			if len(bereit) != 1 || !maps.Equal(bereit[0], soll) {
				t.Errorf("Logeinträge „bereit“ %v, erwartet %v", bereit, soll)
			}
		})
	}
}

// TestAbbruchInJederPhase fängt S2: Menge Schritt {Ablage, Wartung, Service anlegen, erste Session, Listener}
// × Ort {vor dem Schritt, im Schritt, im Schritt mit umhülltem Kontextfehler} × Modus (erste Session nur
// in record) = 10 + 8 + 9 = 27; der umhüllte Kontextfehler (errors.Is findet ihn, == nicht) hat nur einen
// Sinn im Schritt, vor dem Schritt liefert kein Aufruf einen Fehler.
func TestAbbruchInJederPhase(t *testing.T) {
	schritte := map[string][]string{
		modusRecord: {pAblage, pWartung, pDienst, pSession, pListen},
		modusReplay: {pAblage, pWartung, pDienst, pListen},
	}
	ereignis := map[string]string{pAblage: pAblage, pWartung: pWartung, pDienst: "dienst:a", pSession: "session:a", pListen: pListen}
	n := 0
	for _, modus := range modi() {
		for i, schritt := range schritte[modus] {
			for _, ortForm := range []string{"vor", "in", "in/umhuellt"} {
				n++
				ort, umhuellt := strings.CutSuffix(ortForm, "/umhuellt")
				t.Run(modus+"/"+schritt+"/"+ortForm, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					h := neuHaken()
					h.abbrechen, h.umhuellt = cancel, umhuellt
					var erwartet []string
					for j, s := range schritte[modus] {
						if j < i || (ort == "in" && j == i) {
							erwartet = append(erwartet, ereignis[s])
						}
					}
					switch {
					case ort == "in":
						h.phase, h.wirkung = schritt, blockiere
					case i == 0:
						cancel()
					default:
						h.phase, h.wirkung = schritte[modus][i-1], brecheAb
					}
					code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modus, 1), h)
					if code != 0 || len(meldungsZeilen(stderr)) != 0 {
						t.Errorf("Rückgabe %d, Meldungen %v", code, meldungsZeilen(stderr))
					}
					if got := nurStart(h.p.liste()); !slices.Equal(got, erwartet) {
						t.Errorf("Ereignisse %v, erwartet %v", got, erwartet)
					}
					geoeffnet := i > 0
					begonnen := modus == modusRecord && ((ort == "in" && i >= 3) || (ort == "vor" && i >= 4))
					for name, soll := range map[string]bool{pAblageZu: geoeffnet, pCloseAll: begonnen, "anhalten": false} {
						if got := h.p.zaehle(name) > 0; got != soll {
							t.Errorf("Ereignis %s: %v, erwartet %v", name, got, soll)
						}
					}
					if h.offeneListener() != 0 {
						t.Error("ein Listener ist offen")
					}
				})
			}
		}
	}
	if n != 27 {
		t.Errorf("Menge hat %d Elemente, erwartet 27", n)
	}
}

// TestKontextFehlerUmhuellt fängt SPEC-138 Nr. 8 f an den Stellen, die einen Fehler gegen den Kontextfehler
// halten und nicht in TestAbbruchInJederPhase liegen. Menge 1: Quelle im Herunterfahren {CloseAll nur in
// record, Schliessen, Ablage schließen} × Modus × Form {blank, umhüllt}, ctx abgebrochen = 10: Rückgabe 0,
// keine Meldung (abschluss vergleicht mit errors.Is). Menge 2: Fehler des Listeners {Canceled,
// DeadlineExceeded} × Form {blank, umhüllt}, ctx offen = 4: der Fehler ist nicht „Adresse nicht bindbar“
// (kontextFehler vergleicht mit errors.Is und kennt beide), Rückgabe 1 mit der Zeile ohne Abbildung.
func TestKontextFehlerUmhuellt(t *testing.T) {
	formen := map[string]func(error) error{
		"blank":    func(e error) error { return e },
		"umhuellt": func(e error) error { return fmt.Errorf("schritt: %w", e) },
	}
	quellen := map[string][]string{modusRecord: {pCloseAll, pSchliessen, pAblageZu}, modusReplay: {pSchliessen, pAblageZu}}
	n := 0
	for _, modus := range modi() {
		for _, quelle := range quellen[modus] {
			for form, bilde := range formen {
				n++
				t.Run("Herunterfahren/"+modus+"/"+quelle+"/"+form, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					h := neuHaken()
					h.fehlerIn = map[string]error{quelle: bilde(context.Canceled)}
					code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modus, 1), h)
					if zeilen := meldungsZeilen(stderr); code != 0 || len(zeilen) != 0 {
						t.Errorf("Rückgabe %d, Meldungen %q", code, zeilen)
					}
					if h.p.zaehle(pAblageZu) != 1 {
						t.Errorf("Ablage %d-mal geschlossen", h.p.zaehle(pAblageZu))
					}
				})
			}
		}
	}
	for name, fehler := range map[string]error{"Canceled": context.Canceled, "DeadlineExceeded": context.DeadlineExceeded} {
		for form, bilde := range formen {
			n++
			t.Run("Listener/"+name+"/"+form, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				h := neuHaken()
				h.fehlerIn = map[string]error{pListen: bilde(fehler)}
				code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modusRecord, 1), h)
				if zeilen := meldungsZeilen(stderr); code != 1 || !slices.Equal(zeilen, []string{"ODR-E9001: Internal error"}) {
					t.Errorf("Rückgabe %d, Meldungen %q", code, zeilen)
				}
			})
		}
	}
	if n != 14 {
		t.Errorf("Menge hat %d Elemente, erwartet 14", n)
	}
}

// TestKonfigFehlerBeiAbbruch fängt S3: Kontext {offen, abgebrochen} × Verstoß {MODE=hybrid, ungültiger Schlüssel}.
func TestKonfigFehlerBeiAbbruch(t *testing.T) {
	for _, abgebrochen := range []bool{false, true} {
		for _, verstoss := range []string{"MODE=hybrid", "MAX_CAPTURE_BODY=abc"} {
			t.Run(fmt.Sprintf("abgebrochen=%v/%s", abgebrochen, verstoss), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if abgebrochen {
					cancel()
				}
				var stderr syncPuffer
				code := bootstrap.StartMit(ctx, umgebungFuer(t, modusRecord, 1, verstoss), &stderr, bootstrap.StandardBauteile())
				if code != 1 || !strings.Contains(stderr.String(), "ODR-E2001 ") {
					t.Errorf("Rückgabe %d, stderr %q", code, stderr.String())
				}
			})
		}
	}
}

// TestStartLiestUmgebung fängt, dass Start os.Environ liest und stderr statt stdout beschreibt.
func TestStartLiestUmgebung(t *testing.T) {
	t.Setenv("MAX_CAPTURE_BODY", "abc")
	var stdout, stderr syncPuffer
	if code := bootstrap.Start(context.Background(), &stdout, &stderr); code != 1 {
		t.Errorf("Rückgabe %d", code)
	}
	if !strings.Contains(stderr.String(), "ODR-E2001 MAX_CAPTURE_BODY: ") || stdout.String() != "" {
		t.Errorf("stderr %q, stdout %q", stderr.String(), stdout.String())
	}
}

// TestStartFehler fängt S4: Phase {Ablage öffnen, Wartung, Service anlegen, erste Session, Listener,
// Serve, CloseAll, Schliessen, Ablage schließen} × Kontext {offen, nach Beginn der Phase abgebrochen};
// Ablage öffnen nur mit offenem Kontext; replay ohne erste Session und CloseAll: 17 + 13 = 30.
func TestStartFehler(t *testing.T) {
	phasen := []string{pAblage, pWartung, pDienst, pSession, pListen, pServe, pCloseAll, pSchliessen, pAblageZu}
	nurRecord := map[string]bool{pSession: true, pCloseAll: true}
	gebunden := map[string]bool{pServe: true, pCloseAll: true, pSchliessen: true, pAblageZu: true}
	sessionBegonnen := map[string]bool{pSession: true, pListen: true, pServe: true, pCloseAll: true, pSchliessen: true, pAblageZu: true}
	zeile := regexp.MustCompile(`^ODR-E[0-9]{4}( |: )`)
	n := 0
	for _, modus := range modi() {
		for _, phase := range phasen {
			for _, abgebrochen := range []bool{false, true} {
				if (modus == modusReplay && nurRecord[phase]) || (phase == pAblage && abgebrochen) {
					continue
				}
				n++
				t.Run(fmt.Sprintf("%s/%s/abgebrochen=%v", modus, phase, abgebrochen), func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					h := neuHaken()
					h.abbrechen, h.phase, h.wirkung = cancel, phase, scheitere
					if abgebrochen {
						h.wirkung = scheitereNachAbbruch
					}
					code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modus, 2), h)
					zeilen := meldungsZeilen(stderr)
					erwartetZeile := "ODR-E9001: Internal error"
					if phase == pListen {
						erwartetZeile = "ODR-E2001 LISTEN: Adresse nicht bindbar"
					}
					if code != 1 || len(zeilen) != 1 || !zeile.MatchString(zeilen[0]) || zeilen[0] != erwartetZeile {
						t.Errorf("Rückgabe %d, Meldungen %q, erwartet %q", code, zeilen, erwartetZeile)
					}
					if strings.Contains(stderr, "GEHEIMER-WERT") || strings.Contains(strings.Join(zeilen, "\n"), "127.0.0.1") {
						t.Errorf("die Meldung trägt einen Wert:\n%s", stderr)
					}
					soll := map[string]int{pAblageZu: 1, "anhalten": 0, pSchliessen: 0, pCloseAll: 0}
					if phase == pAblage {
						soll[pAblageZu] = 0
					}
					if gebunden[phase] {
						soll["anhalten"], soll[pSchliessen] = 1, 1
					}
					if modus == modusRecord && sessionBegonnen[phase] {
						soll[pCloseAll] = 1
					}
					for name, anzahl := range soll {
						if got := h.p.zaehle(name); got != anzahl {
							t.Errorf("Ereignis %s %d-mal, erwartet %d", name, got, anzahl)
						}
					}
					if h.offeneListener() != 0 {
						t.Error("ein Listener ist offen")
					}
				})
			}
		}
	}
	if n != 30 {
		t.Errorf("Menge hat %d Elemente, erwartet 30", n)
	}
}

// TestStartMehrereFehler fängt S4 für zwei Fehler im selben Lauf: jeder bekommt eine Zeile, der Rest
// des Herunterfahrens läuft.
func TestStartMehrereFehler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h := neuHaken()
	h.abbrechen, h.phase, h.auch, h.wirkung = cancel, pCloseAll, pAblageZu, scheitere
	code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modusRecord, 1), h)
	if zeilen := meldungsZeilen(stderr); code != 1 || !slices.Equal(zeilen, []string{"ODR-E9001: Internal error", "ODR-E9001: Internal error"}) {
		t.Errorf("Rückgabe %d, Meldungen %q", code, zeilen)
	}
	if h.p.zaehle(pAblageZu) != 1 || h.p.zaehle(pSchliessen) != 1 {
		t.Errorf("Ereignisse %v", h.p.liste())
	}
}

// TestServeAusfall fängt SPEC-139 Nr. 6, 8 und SPEC-138 Nr. 8 (e): Modus {record, replay} × Ausgang von
// Serve ohne Anhalten und ohne Abbruch von ctx {Fehler, nil} = 4. Start fährt von sich aus herunter,
// liefert 1 und schreibt genau ODR-E9001; die Testbremse ist ein ctx mit Frist von 5 s, kein Schlaf.
func TestServeAusfall(t *testing.T) {
	ausgaenge := []struct {
		name    string
		wirkung wirkung
	}{{"Fehler", scheitere}, {"nil ohne Anhalten", beendeOhneFehler}}
	n := 0
	for _, modus := range modi() {
		for _, a := range ausgaenge {
			n++
			t.Run(modus+"/"+a.name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				h := neuHaken()
				h.phase, h.wirkung = pServe, a.wirkung
				var stderr syncPuffer
				code := bootstrap.StartMit(ctx, umgebungFuer(t, modus, 1), &stderr, h.bauteile())
				if ctx.Err() != nil {
					t.Fatalf("Start fährt nach dem Ausfall von Serve nicht von sich aus herunter; stderr:\n%s", stderr.String())
				}
				if zeilen := meldungsZeilen(stderr.String()); code != 1 || !slices.Equal(zeilen, []string{"ODR-E9001: Internal error"}) {
					t.Errorf("Rückgabe %d, Meldungen %q", code, zeilen)
				}
				soll := map[string]int{"anhalten": 1, pSchliessen: 1, pAblageZu: 1, pCloseAll: 0}
				if modus == modusRecord {
					soll[pCloseAll] = 1
				}
				for name, anzahl := range soll {
					if got := h.p.zaehle(name); got != anzahl {
						t.Errorf("Ereignis %s %d-mal, erwartet %d", name, got, anzahl)
					}
				}
				if h.offeneListener() != 0 {
					t.Error("ein Listener ist offen")
				}
			})
		}
	}
	if n != 4 {
		t.Errorf("Menge hat %d Elemente, erwartet 4", n)
	}
	t.Run("ctx vor dem Schließen von Fehler abgebrochen", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		h := neuHaken()
		h.phase, h.wirkung, h.abbruchServe = pServe, wartetAufAbbruch, ctx.Done()
		code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modusRecord, 1), h)
		if zeilen := meldungsZeilen(stderr); code != 0 || len(zeilen) != 0 {
			t.Errorf("Rückgabe %d, Meldungen %q", code, zeilen)
		}
	})
}

// TestWarte fängt den Teil von SPEC-139 Nr. 8, den ein Lauf nicht ordnen kann: sind ctx abgebrochen und
// Fehler() ohne Wert geschlossen, wählt select zufällig; jede Wahl ergibt dasselbe (Abbruch, 0).
// Die Schleife von 200 Aufrufen färbt eine Mutation, die nur den Zweig des Schließens trifft.
func TestWarte(t *testing.T) {
	zu := func() <-chan error { c := make(chan error); close(c); return c }
	mitWert := func() <-chan error { c := make(chan error, 1); c <- errPhase; return c }
	abgebrochen := func() context.Context { c, cancel := context.WithCancel(context.Background()); cancel(); return c }
	for range 200 {
		if err := bootstrap.Warte(abgebrochen(), zu()); err != nil {
			t.Fatalf("ctx abgebrochen, Fehler() ohne Wert geschlossen: %v, erwartet nil", err)
		}
	}
	if err := bootstrap.Warte(context.Background(), zu()); !errors.Is(err, bootstrap.ErrServeBeendet) {
		t.Errorf("ctx offen, Fehler() ohne Wert geschlossen: %v", err)
	}
	if err := bootstrap.Warte(context.Background(), mitWert()); !errors.Is(err, errPhase) {
		t.Errorf("ctx offen, Fehler() mit Wert: %v", err)
	}
	if err := bootstrap.Warte(abgebrochen(), make(chan error)); err != nil {
		t.Errorf("ctx abgebrochen, Fehler() offen: %v", err)
	}
}

// TestStartFehlerPaare fängt SPEC-139 Nr. 7 und 8 (B10): jedes mögliche Paar aus den Quellen {Fehler des
// Hochfahrens, Ausfall von Serve, CloseAll, Fehlgeschlagen, Schliessen, Ablage schließen}, die Zeilen
// verschieden je Quelle (Serve und Fehlgeschlagen tragen beide ODR-E9001, das Paar zählt die Zeilen).
// Hochfahren schlägt am Listener fehl; mit Serve und mit Schliessen ist es nicht kombinierbar, weil
// dann kein Server gebunden ist. Die Zeilen stehen in der Reihenfolge der Schritte, die Zeile des Ausfalls
// von Serve zuletzt, hinter allen Zeilen des Herunterfahrens (SPEC-139 Nr. 8); beim Paar von Serve und
// Fehlgeschlagen sind beide Zeilen gleich, die Reihenfolge dort nicht zu unterscheiden.
func TestStartFehlerPaare(t *testing.T) {
	zeileZu := func(c model.Code) string { return c.String() + ": " + c.Message() }
	mitCode := func(c model.Code) error { return &inbound.Fehler{Code: c, Status: 500, Ursache: errPhase} }
	quellen := []struct {
		name  string
		setze func(h *haken)
		zeile string
	}{
		{"Hochfahren", func(h *haken) { h.fehlerIn[pListen] = errPhase }, "ODR-E2001 LISTEN: Adresse nicht bindbar"},
		{"Serve", func(h *haken) { h.fehlerIn[pServe] = errPhase }, "ODR-E9001: Internal error"},
		{"CloseAll", func(h *haken) { h.fehlerIn[pCloseAll] = mitCode(model.CodeE5005) }, zeileZu(model.CodeE5005)},
		{"Fehlgeschlagen", func(h *haken) { h.bericht = inbound.CloseBericht{Fehlgeschlagen: 1} }, "ODR-E9001: Internal error"},
		{"Schliessen", func(h *haken) { h.fehlerIn[pSchliessen] = mitCode(model.CodeE5009) }, zeileZu(model.CodeE5009)},
		{"Ablage", func(h *haken) { h.fehlerIn[pAblageZu] = mitCode(model.CodeE3001) }, zeileZu(model.CodeE3001)},
	}
	nicht := func(a, b string) bool {
		return (a == "Hochfahren" && (b == "Serve" || b == "Schliessen")) || (b == "Hochfahren" && (a == "Serve" || a == "Schliessen"))
	}
	n := 0
	for i, a := range quellen {
		for _, b := range quellen[i+1:] {
			if nicht(a.name, b.name) {
				continue
			}
			n++
			t.Run(a.name+"+"+b.name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				h := neuHaken()
				h.fehlerIn = map[string]error{}
				h.abbrechen = cancel
				a.setze(h)
				b.setze(h)
				code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modusRecord, 1), h)
				erwartet := []string{a.zeile, b.zeile}
				if a.name == "Serve" {
					erwartet = []string{b.zeile, a.zeile}
				}
				zeilen := meldungsZeilen(stderr)
				if code != 1 || !slices.Equal(zeilen, erwartet) {
					t.Errorf("Rückgabe %d, Meldungen %q, erwartet %q", code, zeilen, erwartet)
				}
				if h.p.zaehle(pAblageZu) != 1 {
					t.Errorf("Ablage %d-mal geschlossen", h.p.zaehle(pAblageZu))
				}
			})
		}
	}
	if n != 13 {
		t.Errorf("Menge hat %d Paare, erwartet 13", n)
	}
}

// TestStartMeldungMitCode fängt SPEC-139 Nr. 8 (a): ein Fehler mit Code erscheint als Error() seines Fehlerobjekts.
func TestStartMeldungMitCode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h := neuHaken()
	h.sessionFehler = &inbound.Fehler{Code: model.CodeE5005, Status: 404, Ursache: errPhase}
	code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modusRecord, 1), h)
	soll := "ODR-E5005: " + model.CodeE5005.Message()
	if zeilen := meldungsZeilen(stderr); code != 1 || !slices.Equal(zeilen, []string{soll}) {
		t.Errorf("Rückgabe %d, Meldungen %q, erwartet %q", code, zeilen, soll)
	}
}

// TestPruefenUndWarnung fängt B4: UPSTREAM_TLS_INSECURE_SKIP_VERIFY {nicht gesetzt, true} × Modus.
func TestPruefenUndWarnung(t *testing.T) {
	for _, modus := range modi() {
		for _, skip := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/skip=%v", modus, skip), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				h := neuHaken()
				var extra []string
				if skip {
					extra = append(extra, "UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true")
				}
				code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modus, 1, extra...), h)
				if code != 0 {
					t.Fatalf("Rückgabe %d, stderr:\n%s", code, stderr)
				}
				if (modus == modusRecord) != (h.upstream != nil) || (h.upstream != nil && h.upstream.Pruefen == skip) {
					t.Errorf("upstream %+v bei skip=%v im Modus %s", h.upstream, skip, modus)
				}
				// Jeder Eintrag der Stufe WARN und jeder mit einem Ereignis ist genau der feste Eintrag (SPEC-139
				// Nr. 9), ganz, mit allen Attributen, und er steht nur im einen Fall.
				var auffaellig []map[string]string
				for _, z := range logZeilen(t, stderr) {
					if z["level"] == "WARN" || z["ereignis"] != "" {
						auffaellig = append(auffaellig, z)
					}
				}
				var soll []map[string]string
				if skip && modus == modusRecord {
					soll = append(soll, map[string]string{
						"level": "WARN", "msg": "start: Zertifikatsprüfung zum Upstream abgeschaltet", "ereignis": "tls_pruefung_aus",
					})
				}
				if !slices.EqualFunc(auffaellig, soll, maps.Equal) {
					t.Errorf("Einträge %v, erwartet %v:\n%s", auffaellig, soll, stderr)
				}
			})
		}
	}
}

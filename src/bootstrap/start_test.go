package bootstrap_test

import (
	"context"
	"fmt"
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
		})
	}
}

// TestAbbruchInJederPhase fängt S2: Menge Schritt {Ablage, Wartung, Service anlegen, erste Session, Listener}
// × Ort {vor dem Schritt, im Schritt} × Modus (erste Session nur in record) = 10 + 8 = 18.
func TestAbbruchInJederPhase(t *testing.T) {
	schritte := map[string][]string{
		modusRecord: {pAblage, pWartung, pDienst, pSession, pListen},
		modusReplay: {pAblage, pWartung, pDienst, pListen},
	}
	ereignis := map[string]string{pAblage: pAblage, pWartung: pWartung, pDienst: "dienst:a", pSession: "session:a", pListen: pListen}
	n := 0
	for _, modus := range modi() {
		for i, schritt := range schritte[modus] {
			for _, ort := range []string{"vor", "in"} {
				n++
				t.Run(modus+"/"+schritt+"/"+ort, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					h := neuHaken()
					h.abbrechen = cancel
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
	if n != 18 {
		t.Errorf("Menge hat %d Elemente, erwartet 18", n)
	}
}

// TestKonfigFehlerBeiAbbruch fängt S3: Kontext {offen, abgebrochen} × Verstoß {MODE=mock, ungültiger Schlüssel}.
func TestKonfigFehlerBeiAbbruch(t *testing.T) {
	for _, abgebrochen := range []bool{false, true} {
		for _, verstoss := range []string{"MODE=mock", "MAX_CAPTURE_BODY=abc"} {
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

// TestServeAusfall fängt SPEC-138 Nr. 8 (e): fällt Serve im Lauf aus, fährt Start von sich aus herunter
// und liefert 1, ohne dass ctx abgebrochen wird.
func TestServeAusfall(t *testing.T) {
	h := neuHaken()
	h.phase, h.wirkung = pServe, scheitere
	var stderr syncPuffer
	ergebnis := make(chan int, 1)
	go func() {
		ergebnis <- bootstrap.StartMit(context.Background(), umgebungFuer(t, modusRecord, 1), &stderr, h.bauteile())
	}()
	select {
	case code := <-ergebnis:
		if zeilen := meldungsZeilen(stderr.String()); code != 1 || !slices.Equal(zeilen, []string{"ODR-E9001: Internal error"}) {
			t.Errorf("Rückgabe %d, Meldungen %q", code, zeilen)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("Start fährt nach dem Ausfall von Serve nicht herunter; stderr:\n%s", stderr.String())
	}
	if h.p.zaehle(pAblageZu) != 1 || h.p.zaehle(pSchliessen) != 1 {
		t.Errorf("Ereignisse %v", h.p.liste())
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
				warnungen := 0
				for _, l := range strings.Split(stderr, "\n") {
					if strings.Contains(l, "ereignis=tls_pruefung_aus") && strings.Contains(l, "level=WARN") &&
						strings.Contains(l, `msg="start: Zertifikatsprüfung zum Upstream abgeschaltet"`) {
						warnungen++
					}
				}
				soll := 0
				if skip && modus == modusRecord {
					soll = 1
				}
				if warnungen != soll || strings.Count(stderr, "tls_pruefung_aus") != soll {
					t.Errorf("%d Warnungen, erwartet %d:\n%s", warnungen, soll, stderr)
				}
			})
		}
	}
}

package admin_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// halt setzt den Haltepunkt h der Operation op im Fake, der sie trägt: vor dem Ergebnis, mit nach gleich wahr danach.
func (e *wsEnv) halt(op wsOp, nach bool, h func(context.Context)) {
	vor, danach := &e.wf.vor, &e.wf.nach
	if op == snList || op == snCreate {
		vor, danach = &e.sf.vor, &e.sf.nach
	}
	if nach {
		danach[op] = h
		return
	}
	vor[op] = h
}

// TestWS17Warten hält WS17 (SPEC-298 Nr. 9 a, b) als Wirkungs-Fänger mit Kanälen (M-Warten: 8 Routen × 2 Haltepunkte, im Aufruf
// und nach der Rückkehr der Operation, 16 Zellen): während die Operation blockiert, ist ServeHTTP nicht zurückgekehrt, nichts steht
// am ResponseWriter (kein WriteHeader, kein Write), und der `ctx` der Operation trägt keine Frist; nach der Freigabe ist die Antwort
// die ohne Wartezeit (Status, Body und Header). Das fehlende `WriteTimeout` fängt `grenzen_test.go`.
func TestWS17Warten(t *testing.T) {
	wsFrist(t)
	zellen, routen := 0, 0
	for _, r := range wsRouten() {
		routen++
		ohne := wsNeu(t).rufe(r.gut())
		for _, stelle := range []string{"im Aufruf", "nach der Rückkehr"} {
			zellen++
			name := r.name + " " + stelle
			e := wsNeu(t)
			drin, frei, fertig := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var mitFrist atomic.Bool
			e.halt(r.op, stelle == "nach der Rückkehr", func(ctx context.Context) {
				if _, ok := ctx.Deadline(); ok {
					mitFrist.Store(true)
				}
				close(drin)
				select {
				case <-frei:
				case <-time.After(10 * time.Second):
				}
			})
			var ev ereignisse
			rec := httptest.NewRecorder()
			go func() {
				defer close(fertig)
				e.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &ev}, r.gut())
			}()
			select {
			case <-drin:
			case <-time.After(10 * time.Second):
				t.Fatalf("%s: die Operation wurde nicht gerufen", name)
			}
			select {
			case <-fertig:
				t.Errorf("%s: ServeHTTP kehrte zurück, während die Operation blockierte", name)
			case <-time.After(20 * time.Millisecond): // Grenze: ein ServeHTTP, das nach 20 ms nicht zurückgekehrt ist, wartet auf die Operation
			}
			if evs := ev.lies(); len(evs) != 0 || rec.Body.Len() != 0 {
				t.Errorf("%s: vor der Rückkehr der Operation geschrieben: Ereignisse %q, Body %d Bytes", name, evs, rec.Body.Len())
			}
			close(frei)
			select {
			case <-fertig:
			case <-time.After(10 * time.Second):
				t.Fatalf("%s: ServeHTTP kehrte nach der Freigabe nicht zurück", name)
			}
			if mitFrist.Load() || rec.Code != r.status || rec.Body.String() != ohne.Body.String() || !reflect.DeepEqual(rec.Result().Header, ohne.Result().Header) {
				t.Errorf("%s: Frist im ctx der Operation %v, Antwort %d %q %v, erwartet keine Frist und %d %q %v", name, mitFrist.Load(), rec.Code,
					rec.Body.String(), rec.Result().Header, ohne.Code, ohne.Body.String(), ohne.Result().Header)
			}
			e.p.nurDiagonale(t, name, r.op)
		}
	}
	gleichAuf(t, "Routen, Zellen", []int{routen, zellen}, []int{8, 16})
}

// wsEnde ist ein Ende des `ctx`: Art `cancel` oder `Frist`, Zeitpunkt vor dem Aufruf, im Aufruf oder nach der Rückkehr der
// Operation (bei den Routen mit Eintrag vor Append).
type wsEnde struct{ art, zeit string }

func wsEnden() []wsEnde {
	var out []wsEnde
	for _, art := range []string{"cancel", "Frist"} {
		for _, zeit := range []string{"vor dem Aufruf", "im Aufruf", "nach der Rückkehr"} {
			out = append(out, wsEnde{art, zeit})
		}
	}
	return out
}

// TestWS18Kontext hält WS18 (SPEC-298 Nr. 9 c, d) als Menge M-Kontext: 8 Routen × 6 Enden des `ctx` (Art × Zeitpunkt) × 5 Ausgänge
// des Dienstes (Erfolg, Canceled und DeadlineExceeded je allein und neben einem `*Fehler E7001` im `errors.Join`) = 240 Zellen;
// darin liegen die 140 des Plans (4 Routen mit Eintrag × 4 Enden × 5 = 80, 4 Routen ohne Eintrag × 3 Enden × 5 = 60). Orakel: bei
// nil läuft die Route zu Ende, antwortet wie ohne Ende des `ctx` und schreibt bei den vier Routen mit Eintrag genau einen (Literal),
// bei den anderen keinen; bei einem Kontextfehler bricht die Verbindung ab, ohne Antwort, Zählen und Eintrag. In jeder Zelle ruft
// die Route die Operation genau einmal, und der Fake sieht das Ende des `ctx` (im Aufruf vor der Rückkehr) und den Wert im `ctx`.
func TestWS18Kontext(t *testing.T) {
	wsFrist(t)
	ausgaenge := []struct {
		name string
		err  error
	}{
		{"Erfolg", nil}, {"Canceled", context.Canceled}, {"DeadlineExceeded", context.DeadlineExceeded},
		{"Canceled neben E7001", errors.Join(context.Canceled, wsE7001())}, {"DeadlineExceeded neben E7001", errors.Join(context.DeadlineExceeded, wsE7001())},
	}
	zellen, plan := 0, map[bool]int{}
	for i, r := range wsRouten() {
		mitEintrag := wsEintragArt(i) != ""
		ohne := wsNeu(t)
		wsDienst(ohne)
		ref := ohne.rufe(r.gut())
		for _, ende := range wsEnden() {
			for _, a := range ausgaenge {
				zellen++
				if ende.art == "cancel" || (ende.zeit == "vor dem Aufruf" && mitEintrag) {
					plan[mitEintrag]++
				}
				name := fmt.Sprintf("%s, Ende %s %s, Ausgang %s", r.name, ende.art, ende.zeit, a.name)
				e := wsNeu(t)
				wsDienst(e)
				e.wf.err, e.sf.err = a.err, a.err
				ctx, beende, frei, soll := kontextEnde(ende.art, ende.zeit)
				sah := false
				switch ende.zeit {
				case "vor dem Aufruf":
					beende()
				case "im Aufruf":
					e.halt(r.op, false, func(c context.Context) {
						beende()
						select {
						case <-c.Done():
							sah = true
						case <-time.After(5 * time.Second):
						}
					})
				default:
					e.halt(r.op, true, func(context.Context) { beende() })
				}
				rec := httptest.NewRecorder()
				abbruch := abbruchVon(func() { e.srv.ServeHTTP(rec, r.gut().WithContext(ctx)) })
				frei()
				e.p.nurDiagonale(t, name, r.op)
				if c := e.p.ctxListe()[0]; c.Value(ctxSchluessel{}) != "wert" || c.Err() != soll || (ende.zeit == "im Aufruf" && !sah) {
					t.Errorf("%s: Wert im ctx %v, Ende des ctx gesehen (Err %v, erwartet %v, im Aufruf %v)", name, c.Value(ctxSchluessel{}), c.Err(), soll, sah)
				}
				if codes, andere := zaehlungenVon(e.tele); len(codes) != 0 || andere != 0 {
					t.Errorf("%s: Zählungen %q, andere Aufrufe %d", name, codes, andere)
				}
				if a.err != nil {
					if abbruch != http.ErrAbortHandler || rec.Body.Len() != 0 || e.audit.Aufrufe() != 0 {
						t.Errorf("%s: Panik %v, Body %q, %d Einträge; erwartet den Abbruch ohne Antwort und ohne Eintrag", name, abbruch, rec.Body.String(), e.audit.Aufrufe())
					}
					continue
				}
				if abbruch != nil || rec.Code != r.status || rec.Body.String() != ref.Body.String() || ref.Body.Len() == 0 {
					t.Errorf("%s: Panik %v, Antwort %d %q, erwartet %d %q", name, abbruch, rec.Code, rec.Body.String(), r.status, ref.Body.String())
				}
				wollen := []support.Eintrag{}
				if mitEintrag {
					wollen = append(wollen, wsDienstEintrag(i, rec.Header().Get("X-Correlation-Id")))
				}
				gleichAuf(t, name+": Einträge", append([]support.Eintrag{}, e.audit.Eintraege()...), wollen)
			}
		}
	}
	gleichAuf(t, "Zellen gleich 8 × 6 × 5, Zellen des Plans (80 + 60)", []int{zellen, plan[true], plan[false]}, []int{240, 80, 60})
}

// TestWS18Kontextfehler hält WS18 (SPEC-298 Nr. 9 c; SPEC-238 Nr. 6 a 1) für jede Form eines Kontextfehlers beim Aufruf des
// Dienstes: 24 Formen (kontextfehlerFormen: Art × 12 Verpackungen) × 2 Wertangaben (der Dienst liefert zum Fehler einen Wert oder
// keinen) × 3 Zustände des `ctx` (aktiv, `cancel` oder Frist vor dem Aufruf) × 8 Routen = 1152 Zellen. Orakel: der Abbruch ohne
// Antwort, Zählen und Eintrag, und genau ein Aufruf der Operation der Route.
func TestWS18Kontextfehler(t *testing.T) {
	wsFrist(t)
	formen := kontextfehlerFormen()
	gleichAuf(t, "Formen", len(formen), 2*12)
	zustaende := []string{"aktiv", "cancel vor dem Aufruf", "Frist vor dem Aufruf"}
	zellen := 0
	for _, r := range wsRouten() {
		for _, f := range formen {
			for _, mitWert := range []bool{false, true} {
				for _, zustand := range zustaende {
					zellen++
					name := fmt.Sprintf("%s, %s, mit Wert %v, ctx %s", r.name, f.name, mitWert, zustand)
					e := wsNeu(t)
					e.wf.err, e.sf.err = f.err, f.err
					if !mitWert {
						e.wf.ws, e.sf.snap = model.Workspace{}, model.WorkspaceSnapshot{}
					}
					art, zeit := "cancel", ""
					if zustand == "Frist vor dem Aufruf" {
						art, zeit = "Frist", "vor dem Aufruf"
					}
					ctx, beende, frei, _ := kontextEnde(art, zeit)
					if zustand == "cancel vor dem Aufruf" {
						beende()
					}
					var ev ereignisse
					rec := httptest.NewRecorder()
					abbruch := abbruchVon(func() {
						e.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &ev}, r.gut().WithContext(ctx))
					})
					frei()
					e.p.nurDiagonale(t, name, r.op)
					if evs := ev.lies(); abbruch != http.ErrAbortHandler || len(evs) != 0 || rec.Body.Len() != 0 || e.audit.Aufrufe() != 0 {
						t.Errorf("%s: Panik %v, Ereignisse %q, Body %q, %d Einträge; erwartet den Abbruch ohne Antwort und ohne Eintrag", name, abbruch, evs, rec.Body.String(), e.audit.Aufrufe())
					}
					if codes, andere := zaehlungenVon(e.tele); len(codes) != 0 || andere != 0 {
						t.Errorf("%s: Zählungen %q, andere Aufrufe %d", name, codes, andere)
					}
				}
			}
		}
	}
	gleichAuf(t, "Zellen gleich 24 × 2 × 3 × 8", zellen, 1152)
}

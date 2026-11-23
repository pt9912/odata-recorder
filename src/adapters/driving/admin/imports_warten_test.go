package admin_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// TestIM18Warten hält IM18 (SPEC-297 Nr. 9 a, b): die Antwort entsteht erst nach der Rückkehr von Import und Get, ohne Frist.
// Wirkungs-Fänger mit Kanälen (2 Routen × 2 Haltepunkte, im Aufruf und nach der Rückkehr des Fakes, 4 Zellen): während der Dienst
// blockiert, ist ServeHTTP nicht zurückgekehrt, es steht nichts am ResponseWriter (kein WriteHeader, kein Write) und der Body ist
// leer, der `ctx` des Dienstes trägt keine Frist; nach der Freigabe ist die Antwort die ohne Wartezeit (201 bzw. 200 mit dem
// Report, die Header der Erfolgsantwort). Das fehlende `WriteTimeout` fängt `grenzen_test.go`; die Frist von 10 s je Wartepunkt
// ist nur der Abbruch im Fehlschlag.
func TestIM18Warten(t *testing.T) {
	zellen := 0
	for _, route := range []string{"POST", "GET"} {
		op, status := map[string]impOp{"POST": impImport, "GET": impGet}[route], map[string]int{"POST": 201, "GET": 200}[route]
		for _, stelle := range []string{"im Aufruf", "nach der Rückkehr"} {
			zellen++
			name := route + " " + stelle
			u, fake, _ := importsUmgebung(t)
			drin, frei, fertig := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var mitFrist atomic.Bool
			halt := func(ctx context.Context) {
				if _, ok := ctx.Deadline(); ok {
					mitFrist.Store(true)
				}
				close(drin)
				select {
				case <-frei:
				case <-time.After(10 * time.Second):
				}
			}
			if stelle == "im Aufruf" {
				fake.vor[op] = halt
			} else {
				fake.nach[op] = halt
			}
			var e ereignisse
			rec := httptest.NewRecorder()
			go func() {
				defer close(fertig)
				u.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &e}, importRequestFuer(route))
			}()
			select {
			case <-drin:
			case <-time.After(10 * time.Second):
				t.Fatalf("%s: der Dienst wurde nicht gerufen", name)
			}
			select {
			case <-fertig:
				t.Errorf("%s: ServeHTTP kehrte zurück, während der Dienst blockierte", name)
			default:
			}
			if ev := e.lies(); len(ev) != 0 || rec.Body.Len() != 0 {
				t.Errorf("%s: vor der Rückkehr des Dienstes geschrieben: Ereignisse %q, Body %d Bytes", name, ev, rec.Body.Len())
			}
			close(frei)
			select {
			case <-fertig:
			case <-time.After(10 * time.Second):
				t.Fatalf("%s: ServeHTTP kehrte nach der Freigabe nicht zurück", name)
			}
			if mitFrist.Load() || rec.Code != status || rec.Body.String() != reportA {
				t.Errorf("%s: Frist im ctx des Dienstes %v, Antwort %d %q, erwartet keine Frist und %d %q", name, mitFrist.Load(), rec.Code, rec.Body.String(), status, reportA)
			}
			pruefeErfolgsHeader(t, name, rec)
			fake.nurDiagonale(t, name, op)
		}
	}
	gleichAuf(t, "Zellen", zellen, 4)
}

// fristKontext ist ein Kontext, dessen Frist der Test ablaufen lässt: `Done` schließt mit ablaufen, `Err` liefert danach
// context.DeadlineExceeded (wie ein Kontext, dessen Frist endet), `Deadline` nennt vor dem Ablauf eine Frist in einer Stunde und
// danach eine vor einer Stunde, `Value` liefert den Wert unter ctxSchluessel{}. So endet die Frist zu einem Zeitpunkt, den der
// Test setzt (im Aufruf, nach der Rückkehr), ohne Uhr und ohne Wettlauf mit ihr.
type fristKontext struct {
	wert   any
	fertig chan struct{}
	einmal sync.Once
}

func neuerFristKontext(wert any) *fristKontext {
	return &fristKontext{wert: wert, fertig: make(chan struct{})}
}

func (f *fristKontext) Done() <-chan struct{} { return f.fertig }

func (f *fristKontext) Value(key any) any {
	if key == (ctxSchluessel{}) {
		return f.wert
	}
	return nil
}

func (f *fristKontext) Err() error {
	select {
	case <-f.fertig:
		return context.DeadlineExceeded
	default:
		return nil
	}
}

func (f *fristKontext) Deadline() (time.Time, bool) {
	if f.Err() != nil {
		return time.Now().Add(-time.Hour), true
	}
	return time.Now().Add(time.Hour), true
}

func (f *fristKontext) ablaufen() { f.einmal.Do(func() { close(f.fertig) }) }

// kontextEnde liefert einen ctx mit einem Wert und seiner Art des Endes (art `cancel` oder `Frist`): beende endet ihn, frei gibt
// seine Mittel frei, soll ist der Fehler, den `Err` nach dem Ende liefert. Die Frist vor dem Aufruf ist eine des Pakets context
// (bei der Anlage abgelaufen, beende ist dort ohne Wirkung); jede andere Frist ist ein fristKontext.
func kontextEnde(art, zeit string) (ctx context.Context, beende, frei func(), soll error) {
	eltern := context.WithValue(context.Background(), ctxSchluessel{}, "wert")
	switch {
	case art == "cancel":
		c, cancel := context.WithCancel(eltern)
		return c, cancel, cancel, context.Canceled
	case zeit == "vor dem Aufruf":
		c, cancel := context.WithDeadline(eltern, time.Now().Add(-time.Hour))
		return c, func() {}, cancel, context.DeadlineExceeded
	}
	f := neuerFristKontext("wert")
	return f, f.ablaufen, func() {}, context.DeadlineExceeded
}

// TestIM19Kontext hält IM19 (SPEC-297 Nr. 9 c, d; SPEC-238 Nr. 9) als Menge M-Kontext: Route × Ende des `ctx` (Art `cancel` und
// `Frist` × Zeitpunkt vor dem Aufruf, im Aufruf und nach der Rückkehr von Import oder Get, bei `POST` vor Append: 6 Enden je
// Route) × Ausgang des Dienstes; `POST` 6 Ausgänge (Erfolg mit `Neu` wahr und mit `Neu` falsch, Canceled und DeadlineExceeded je
// allein und mit einem Wert zugleich) = 36 Zellen, `GET` 5 Ausgänge (der Erfolg ohne die Achse `Neu`, die nur `POST` hat) = 30
// Zellen, zusammen 66. Orakel: liefert der Dienst Wert und nil, läuft die Route zu Ende, antwortet 201 (`Neu`
// wahr) bzw. 200 (`POST` mit `Neu` falsch, `GET`) mit dem Report und schreibt bei `POST` in jeder dieser Zellen genau einen
// Eintrag, bei 201 und bei 200 und bei jedem Ende des `ctx`, bei `GET` keinen; liefert er einen Kontextfehler, bricht die
// Verbindung ab, ohne Antwort, ohne Zählen und ohne Eintrag. In jeder Zelle ruft die Route den Dienst genau einmal (auch bei
// beendetem `ctx` beim Eintritt), der Dienst sieht das Ende des `ctx` mit dem Fehler der Art (im Aufruf: vor der Rückkehr, sonst
// am `ctx`, den er erhielt) und den Wert im `ctx`.
func TestIM19Kontext(t *testing.T) {
	var enden []struct{ art, zeit string }
	for _, art := range []string{"cancel", "Frist"} {
		for _, zeit := range []string{"vor dem Aufruf", "im Aufruf", "nach der Rückkehr"} {
			enden = append(enden, struct{ art, zeit string }{art, zeit})
		}
	}
	type ausgang struct {
		name    string
		err     error
		mitWert bool
		neu     bool
	}
	ausgaenge := map[string][]ausgang{
		"POST": {
			{"Erfolg mit Neu wahr", nil, true, true}, {"Erfolg mit Neu falsch", nil, true, false},
			{"Canceled", context.Canceled, false, true}, {"Canceled mit Wert", context.Canceled, true, true},
			{"DeadlineExceeded", context.DeadlineExceeded, false, true}, {"DeadlineExceeded mit Wert", context.DeadlineExceeded, true, true},
		},
		"GET": {
			{"Erfolg", nil, true, false},
			{"Canceled", context.Canceled, false, false}, {"Canceled mit Wert", context.Canceled, true, false},
			{"DeadlineExceeded", context.DeadlineExceeded, false, false}, {"DeadlineExceeded mit Wert", context.DeadlineExceeded, true, false},
		},
	}
	zellen := 0
	for _, route := range []string{"POST", "GET"} {
		op := map[string]impOp{"POST": impImport, "GET": impGet}[route]
		for _, ende := range enden {
			for _, a := range ausgaenge[route] {
				zellen++
				name := route + ", Ende " + ende.art + " " + ende.zeit + ", Ausgang " + a.name
				status := 200
				if route == "POST" && a.neu {
					status = 201
				}
				u, fake, _ := importsUmgebung(t)
				fake.err, fake.ergebnis.Neu = a.err, a.neu
				if !a.mitWert {
					fake.ergebnis, fake.imp = inbound.ImportErgebnis{}, model.Import{}
				}
				ctx, beende, frei, soll := kontextEnde(ende.art, ende.zeit)
				sah := false
				switch ende.zeit {
				case "vor dem Aufruf":
					beende()
				case "im Aufruf":
					fake.vor[op] = func(c context.Context) {
						beende()
						select {
						case <-c.Done():
							sah = true
						case <-time.After(5 * time.Second):
						}
					}
				default:
					fake.nach[op] = func(context.Context) { beende() }
				}
				rec := httptest.NewRecorder()
				abbruch := abbruchVon(func() { u.srv.ServeHTTP(rec, importRequestFuer(route).WithContext(ctx)) })
				frei()
				fake.nurDiagonale(t, name, op)
				if c := fake.ctxListe()[0]; c.Value(ctxSchluessel{}) != "wert" || c.Err() != soll || (ende.zeit == "im Aufruf" && !sah) {
					t.Errorf("%s: Wert im ctx %v, Ende des ctx gesehen (Err %v, erwartet %v, im Aufruf %v)", name, c.Value(ctxSchluessel{}), c.Err(), soll, sah)
				}
				if codes, andere := zaehlungenVon(u.tele); len(codes) != 0 || andere != 0 {
					t.Errorf("%s: Zählungen %q, andere Aufrufe %d", name, codes, andere)
				}
				if a.err != nil {
					if abbruch != http.ErrAbortHandler || rec.Body.Len() != 0 || u.audit.Aufrufe() != 0 {
						t.Errorf("%s: Panik %v, Body %q, %d Einträge; erwartet den Abbruch ohne Antwort und ohne Eintrag", name, abbruch, rec.Body.String(), u.audit.Aufrufe())
					}
					continue
				}
				if abbruch != nil || rec.Code != status || rec.Body.String() != reportA {
					t.Errorf("%s: Panik %v, Antwort %d %q, erwartet %d %q", name, abbruch, rec.Code, rec.Body.String(), status, reportA)
				}
				wollen := []support.Eintrag{}
				if route == "POST" {
					wollen = append(wollen, eintragImport(rec.Header().Get("X-Correlation-Id")))
				}
				gleichAuf(t, name+": Einträge", append([]support.Eintrag{}, u.audit.Eintraege()...), wollen)
			}
		}
	}
	gleichAuf(t, "Zellen gleich 36 + 30, Enden, Ausgänge je Route", []int{zellen, len(enden) * len(ausgaenge["POST"]), len(enden) * len(ausgaenge["GET"]), len(enden), len(ausgaenge["POST"]), len(ausgaenge["GET"])},
		[]int{66, 36, 30, 6, 6, 5})
}

// kontextfehlerFormen sind die Formen eines Kontextfehlers, den ein Aufruf des Dienstes liefern kann (SPEC-238 Nr. 6 a 1: der
// Fehler enthält Canceled oder DeadlineExceeded, auch verpackt, in einem `errors.Join`, neben anderen Fehlern, als Ursache eines
// `*inbound.Fehler`): Art (Canceled, DeadlineExceeded) × Verpackung (12: bloß, mit `%w`, mit `%w` zweimal verschachtelt, Join mit
// dem Kontextfehler zuerst, Join mit ihm zuletzt, Join mit nil zuerst, ein Errorf mit zwei `%w`, ein Join in einem `%w`, Ursache
// eines `*inbound.Fehler` mit Code und Status, neben einem gültigen `*inbound.Fehler` in beiden Folgen, neben einem nil-Zeiger
// `*inbound.Fehler`) = 24 Formen.
func kontextfehlerFormen() []fehlerFall {
	var out []fehlerFall
	for _, k := range []struct {
		name string
		err  error
	}{{"Canceled", context.Canceled}, {"DeadlineExceeded", context.DeadlineExceeded}} {
		andere, e7003 := errors.New("andere"), fehlerE7003()
		for _, v := range []struct {
			name string
			err  error
		}{
			{"bloß", k.err},
			{"mit %w", fmt.Errorf("x: %w", k.err)},
			{"mit %w zweimal verschachtelt", fmt.Errorf("a: %w", fmt.Errorf("b: %w", k.err))},
			{"Join, Kontextfehler zuerst", errors.Join(k.err, andere)},
			{"Join, Kontextfehler zuletzt", errors.Join(andere, k.err)},
			{"Join mit nil zuerst", errors.Join(nil, k.err)},
			{"Errorf mit zwei %w", fmt.Errorf("%w und %w", andere, k.err)},
			{"Join in %w", fmt.Errorf("x: %w", errors.Join(k.err, andere))},
			{"Ursache eines *inbound.Fehler", &inbound.Fehler{Code: model.CodeE7003, Status: 404, Ursache: k.err}},
			{"hinter einem *inbound.Fehler", errors.Join(e7003, k.err)},
			{"vor einem *inbound.Fehler", errors.Join(k.err, e7003)},
			{"neben einem nil-Zeiger *inbound.Fehler", errors.Join((*inbound.Fehler)(nil), k.err)},
		} {
			out = append(out, fehlerFall{name: k.name + " " + v.name, err: v.err, abbruch: true})
		}
	}
	return out
}

// TestIM19Kontextfehler hält IM19 (SPEC-297 Nr. 9 c; SPEC-238 Nr. 6 a 1, 9) für jede Form eines Kontextfehlers beim Aufruf des
// Dienstes als Schleife über die Menge: 24 Formen (kontextfehlerFormen) × 2 Wertangaben (der Dienst liefert zum Fehler einen Wert
// oder keinen) × 3 Zustände des `ctx` beim Aufruf (aktiv, `cancel` vor dem Aufruf, Frist vor dem Aufruf abgelaufen) × 2 Routen
// (`POST` mit `Import`, `GET` mit `Get`) = 288 Zellen. Orakel in jeder Zelle: die Verbindung bricht ab, nichts ist an den
// ResponseWriter geschrieben (kein WriteHeader, kein Write), kein Zählen, kein Eintrag, und der Dienst wurde genau einmal gerufen
// (die andere Operation nie).
func TestIM19Kontextfehler(t *testing.T) {
	formen := kontextfehlerFormen()
	gleichAuf(t, "Formen", len(formen), 2*12)
	zustaende := []string{"aktiv", "cancel vor dem Aufruf", "Frist vor dem Aufruf"}
	zellen := 0
	for _, route := range []string{"POST", "GET"} {
		op := map[string]impOp{"POST": impImport, "GET": impGet}[route]
		for _, f := range formen {
			for _, mitWert := range []bool{false, true} {
				for _, zustand := range zustaende {
					zellen++
					name := fmt.Sprintf("%s, %s, mit Wert %v, ctx %s", route, f.name, mitWert, zustand)
					u, fake, _ := importsUmgebung(t)
					fake.err = f.err
					if !mitWert {
						fake.ergebnis, fake.imp = inbound.ImportErgebnis{}, model.Import{}
					}
					var ctx context.Context
					var frei func()
					switch zustand {
					case "aktiv":
						ctx, _, frei, _ = kontextEnde("cancel", "")
					case "cancel vor dem Aufruf":
						var beende func()
						ctx, beende, frei, _ = kontextEnde("cancel", "")
						beende()
					default:
						ctx, _, frei, _ = kontextEnde("Frist", "vor dem Aufruf")
					}
					var e ereignisse
					rec := httptest.NewRecorder()
					abbruch := abbruchVon(func() {
						u.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &e}, importRequestFuer(route).WithContext(ctx))
					})
					frei()
					fake.nurDiagonale(t, name, op)
					if ev := e.lies(); abbruch != http.ErrAbortHandler || len(ev) != 0 || rec.Body.Len() != 0 || u.audit.Aufrufe() != 0 {
						t.Errorf("%s: Panik %v, Ereignisse %q, Body %q, %d Einträge; erwartet den Abbruch ohne Antwort und ohne Eintrag", name, abbruch, ev, rec.Body.String(), u.audit.Aufrufe())
					}
					if codes, andere := zaehlungenVon(u.tele); len(codes) != 0 || andere != 0 {
						t.Errorf("%s: Zählungen %q, andere Aufrufe %d", name, codes, andere)
					}
				}
			}
		}
	}
	gleichAuf(t, "Zellen gleich 24 × 2 × 3 × 2", zellen, len(formen)*2*len(zustaende)*2)
	gleichAuf(t, "Zellen", zellen, 288)
}

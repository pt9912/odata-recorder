package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// TestIM19Kontext hält IM19 (SPEC-297 Nr. 9 c, d; SPEC-238 Nr. 9) als Menge M-Kontext: `POST` × Ende des `ctx` (`cancel` vor dem
// Aufruf, abgelaufene Frist vor dem Aufruf, im Aufruf, nach der Rückkehr von Import und vor Append: 4) × Ausgang des Dienstes
// (Erfolg; Canceled und DeadlineExceeded je allein und mit einem Wert zugleich: 5) = 20, `GET` × 3 Enden (vor dem Aufruf, im
// Aufruf, nach der Rückkehr) × 5 = 15, zusammen 35, dazu 5 Zellen, die der Plan nicht nennt: `GET` mit abgelaufener Frist vor dem
// Aufruf (Ende wie bei `POST`), zusammen 40. Orakel: liefert der Dienst Wert und nil, läuft die Route zu Ende, antwortet
// 201 bzw. 200 mit dem Report und schreibt bei `POST` genau einen Eintrag, bei `GET` keinen; liefert er einen Kontextfehler,
// bricht die Verbindung ab, ohne Antwort, ohne Zählen und ohne Eintrag. In jeder Zelle ruft die Route den Dienst genau einmal
// (auch bei beendetem `ctx` beim Eintritt), der Dienst sieht das Ende des `ctx` (im Aufruf: vor der Rückkehr, sonst am `ctx`, den
// er erhielt) und den Wert im `ctx`.
func TestIM19Kontext(t *testing.T) {
	enden := map[string][]string{
		"POST": {"cancel vor dem Aufruf", "Frist vor dem Aufruf", "im Aufruf", "nach der Rückkehr, vor Append"},
		"GET":  {"cancel vor dem Aufruf", "Frist vor dem Aufruf", "im Aufruf", "nach der Rückkehr"},
	}
	ausgaenge := []struct {
		name    string
		err     error
		mitWert bool
	}{
		{"Erfolg", nil, true},
		{"Canceled", context.Canceled, false}, {"Canceled mit Wert", context.Canceled, true},
		{"DeadlineExceeded", context.DeadlineExceeded, false}, {"DeadlineExceeded mit Wert", context.DeadlineExceeded, true},
	}
	zellen := 0
	for _, route := range []string{"POST", "GET"} {
		op, status := map[string]impOp{"POST": impImport, "GET": impGet}[route], map[string]int{"POST": 201, "GET": 200}[route]
		for _, ende := range enden[route] {
			for _, a := range ausgaenge {
				zellen++
				name := route + ", Ende " + ende + ", Ausgang " + a.name
				u, fake, _ := importsUmgebung(t)
				fake.err = a.err
				if !a.mitWert {
					fake.ergebnis, fake.imp = inbound.ImportErgebnis{}, model.Import{}
				}
				ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxSchluessel{}, "wert"))
				sah := false
				switch ende {
				case "cancel vor dem Aufruf":
					cancel()
				case "Frist vor dem Aufruf":
					ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Hour))
				case "im Aufruf":
					fake.vor[op] = func(c context.Context) {
						cancel()
						select {
						case <-c.Done():
							sah = true
						case <-time.After(5 * time.Second):
						}
					}
				default:
					fake.nach[op] = func(context.Context) { cancel() }
				}
				rec := httptest.NewRecorder()
				abbruch := abbruchVon(func() { u.srv.ServeHTTP(rec, importRequestFuer(route).WithContext(ctx)) })
				cancel()
				fake.nurDiagonale(t, name, op)
				if c := fake.ctxListe()[0]; c.Value(ctxSchluessel{}) != "wert" || c.Err() == nil || (ende == "im Aufruf" && !sah) {
					t.Errorf("%s: Wert im ctx %v, Ende des ctx gesehen (Err %v, im Aufruf %v)", name, c.Value(ctxSchluessel{}), c.Err(), sah)
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
	gleichAuf(t, "Zellen gleich 20 + 15 + 5 Zusatzzellen", []int{zellen, len(enden["POST"]) * len(ausgaenge), len(enden["GET"]) * len(ausgaenge)}, []int{40, 20, 20})
}

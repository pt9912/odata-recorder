package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// inspTreffen hält jeden Aufruf von halt, bis ziel Aufrufe eingetreten sind (Wirkungs-Fänger gegen eine Sperre über dem Aufruf des
// Dienstes: unter einer Sperre tritt nie ein zweiter Aufruf ein); nach 5 s ohne den letzten Aufruf endet das Warten, der Treffpunkt
// ist verpasst, und jeder spätere Aufruf kehrt sofort zurück.
type inspTreffen struct {
	n        atomic.Int64
	ziel     int64
	zu       chan struct{}
	einmal   sync.Once
	verpasst atomic.Bool
}

func neuesInspTreffen(ziel int) *inspTreffen {
	return &inspTreffen{ziel: int64(ziel), zu: make(chan struct{})}
}

func (p *inspTreffen) halt() {
	if p.verpasst.Load() {
		return
	}
	if p.n.Add(1) == p.ziel {
		p.einmal.Do(func() { close(p.zu) })
	}
	select {
	case <-p.zu:
	case <-time.After(5 * time.Second):
		p.verpasst.Store(true)
	}
}

// inspGleichzeitig sendet n Requests zugleich (nach einer Schranke) und liefert die Antworten in der Folge der Requests; nach der
// Frist ohne Rückkehr aller bricht der Test mit Fatalf ab, statt an das Zeitlimit des Pakets zu geraten.
func inspGleichzeitig(t *testing.T, u *umgebung, n int, bau func(i int) *http.Request) []*httptest.ResponseRecorder {
	t.Helper()
	out := make([]*httptest.ResponseRecorder, n)
	schranke := neueSchranke(n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := bau(i)
			out[i] = httptest.NewRecorder()
			schranke.warte()
			u.srv.ServeHTTP(out[i], r)
		}()
	}
	fertig := make(chan struct{})
	go func() { wg.Wait(); close(fertig) }()
	select {
	case <-fertig:
	case <-time.After(30 * time.Second):
		t.Fatalf("Frist von 30 s abgelaufen: die %d gleichzeitigen Requests kehrten nicht zurück", n)
	}
	return out
}

// TestN5GleichzeitigeRequests hält N5 (SPEC-293 Nr. 10): in 150 Runden je 16 gleichzeitige Requests je Route (das Fenster eines
// Zustands zwischen Requests trifft in einer Runde nur selten), jeder mit eigener Eingabe, deren Antwort aus dem Argument des Dienstes
// entsteht; `exchanges` trägt je vier Requests dieselbe Session mit verschiedenen Seiten. Jede Antwort ist die zu ihrem Request
// (Orakel: das Literal aus der eigenen Eingabe). Wirkungs-Fänger: alle 16 Requests einer Route sind zugleich im Aufruf des Dienstes
// (Treffpunkt), keine Sperre liegt über dem Aufruf; der Zustand zwischen den Requests (vor und nach dem Aufruf) zeigt sich als
// fremde Antwort und, nach dem Treffpunkt ungeordnet, dem Detektor von `make test-race`. Der Test hat eine eigene Frist.
func TestN5GleichzeitigeRequests(t *testing.T) {
	const n, runden = 16, 150
	u, fake, _ := inspUmgebung(t)
	wert := func(ctx context.Context) int { return ctx.Value(ctxSchluessel{}).(int) }
	fake.exchangesFn = func(_ context.Context, id string, s inbound.Seite) (inbound.ExchangeSeite, error) {
		z, _ := inspZeile(1, map[string]inspLit{"ID": {id, `"` + id + `"`}, "SequenceNo": {s.Nach + 1, strconv.Itoa(s.Nach + 1)}})
		return inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{z}, Limit: s.Limit, Gesamt: len(id), Weiter: s.Nach + s.Limit}, nil
	}
	fake.coverageFn = func(_ context.Context, id string) ([]inbound.CoverageEintrag, error) {
		return []inbound.CoverageEintrag{{EntitySet: id, Status: model.CoverageComplete, Evidenz: `{"id":"` + id + `"}`}}, nil
	}
	fake.diagnoseFn = func(ctx context.Context) (inbound.Diagnose, error) {
		k := strconv.Itoa(wert(ctx))
		return inbound.Diagnose{Modus: "m" + k, Einstellungen: map[string]string{"k" + k: "v" + k}, Fehler: []inbound.FehlerZaehler{{Code: model.CodeE2102, Anzahl: uint64(wert(ctx))}}}, nil
	}
	mitKennung := func(r *http.Request, k int) *http.Request {
		return r.WithContext(context.WithValue(r.Context(), ctxSchluessel{}, k))
	}
	routen := []struct {
		name string
		op   inspOp
		bau  func(k int) *http.Request
		soll func(k int) string
	}{
		{"exchanges", inspOpExchanges, func(k int) *http.Request {
			return inspReq(exchangesZiel("ses-"+strconv.Itoa(k%4)), "limit="+strconv.Itoa(k%7+1)+"&after="+strconv.Itoa(k*10))
		}, func(k int) string {
			id, nach, limit := "ses-"+strconv.Itoa(k%4), k*10, k%7+1
			_, js := inspZeile(1, map[string]inspLit{"ID": {id, `"` + id + `"`}, "SequenceNo": {nach + 1, strconv.Itoa(nach + 1)}})
			return inspSeiteJSON([]string{js}, limit, len(id), strconv.Itoa(nach+limit))
		}},
		{"coverage", inspOpCoverage, func(k int) *http.Request {
			return inspReq(inspCoverage, "baseline_id=bl-"+strconv.Itoa(k))
		}, func(k int) string {
			id := "bl-" + strconv.Itoa(k)
			return `{"entries":[{"entity_set":"` + id + `","status":"complete","evidence":{"id":"` + id + `"}}]}`
		}},
		{"diagnostics", inspOpDiagnose, func(int) *http.Request { return inspReq(inspDiagnostics, "") }, func(k int) string {
			s := strconv.Itoa(k)
			return `{"mode":"m` + s + `","settings":{"k` + s + `":"v` + s + `"},"errors":[{"code":"ODR-E2102","count":` + s + `}]}`
		}},
	}
	beginn := time.Now()
	for r := range runden {
		for _, route := range routen {
			p := neuesInspTreffen(n)
			fake.vor[route.op] = func(context.Context) { p.halt() }
			antworten := inspGleichzeitig(t, u, n, func(i int) *http.Request { return mitKennung(route.bau(r*n+i), r*n+i) })
			if p.verpasst.Load() || p.n.Load() != n {
				t.Fatalf("Runde %d %s: nicht alle %d Requests waren zugleich im Aufruf des Dienstes (%d Aufrufe)", r, route.name, n, p.n.Load())
			}
			for i, rec := range antworten {
				if rec.Code != 200 || rec.Body.String() != route.soll(r*n+i) {
					t.Fatalf("Runde %d %s %d: %d %q, erwartet %q", r, route.name, i, rec.Code, rec.Body.String(), route.soll(r*n+i))
				}
			}
			if time.Since(beginn) > 120*time.Second {
				t.Fatalf("Frist von 120 s abgelaufen in Runde %d: die Runden von N5 endeten nicht", r)
			}
		}
	}
	gleichAuf(t, "Aufrufe des Dienstes", fake.zahlen(), [inspOpAnzahl]int{runden * n, runden * n, runden * n})
	if u.audit.Aufrufe() != 0 {
		t.Errorf("%d Aufrufe von Audit, erwartet 0", u.audit.Aufrufe())
	}
}

package admin_test

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Pfade der drei Routen der Inspektion (SPEC-293 Nr. 1) als Literale des Vertrags; exchangesZiel setzt den kodierten
// Pfadwert ein.
const (
	inspCoverage    = "/__admin/coverage"
	inspDiagnostics = "/__admin/diagnostics"
)

func exchangesZiel(id string) string { return "/__admin/sessions/" + id + "/exchanges" }

// Die Fehlerobjekte der Codes, die ein Dienst der Inspektion liefert, als Literale des Vertrags (SPEC-109 Nr. 2, SPEC-240).
const objektE7002 = `{"error":{"code":"ODR-E7002","message":"Snapshot unknown","target":"workspace"}}`

// inspOp nennt die drei Operationen von inbound.InspectionUseCase in der Folge des Ports.
type inspOp int

const (
	inspOpExchanges inspOp = iota
	inspOpCoverage
	inspOpDiagnose
	inspOpAnzahl
)

// inspAufruf hält einen Aufruf fest: Operation und jedes Argument außer ctx (id und Seite bei Exchanges, id die Baseline bei
// Coverage); den ctx hält der Fake in einer eigenen Liste.
type inspAufruf struct {
	op    inspOp
	id    string
	seite inbound.Seite
}

// inspektionFake erfüllt inbound.InspectionUseCase: jede der drei Operationen zählt ihren Aufruf und hält ctx und jedes
// Argument fest (auch Seite); Ergebnis und Fehler sind einstellbar (seite, eintraege, diagnose, err; die Funktionen
// ersetzen sie je Operation), vor ist je Operation ein Haltepunkt, der nach dem Zählen ohne die Sperre des Fakes läuft.
// Der Fake gerät nie in Panik und ist für gleichzeitigen Aufruf sicher; die Einstellungen setzt der Test vor dem ersten
// Aufruf.
type inspektionFake struct {
	mu        sync.Mutex
	aufrufe   []inspAufruf
	ctxs      []context.Context
	zaehler   [inspOpAnzahl]int
	seite     inbound.ExchangeSeite
	eintraege []inbound.CoverageEintrag
	diagnose  inbound.Diagnose
	err       error

	exchangesFn func(ctx context.Context, id string, seite inbound.Seite) (inbound.ExchangeSeite, error)
	coverageFn  func(ctx context.Context, id string) ([]inbound.CoverageEintrag, error)
	diagnoseFn  func(ctx context.Context) (inbound.Diagnose, error)
	vor         [inspOpAnzahl]func(ctx context.Context)
}

func (f *inspektionFake) rufe(ctx context.Context, a inspAufruf) {
	f.mu.Lock()
	f.aufrufe = append(f.aufrufe, a)
	f.ctxs = append(f.ctxs, ctx)
	f.zaehler[a.op]++
	vor := f.vor[a.op]
	f.mu.Unlock()
	if vor != nil {
		vor(ctx)
	}
}

func (f *inspektionFake) Exchanges(ctx context.Context, sessionID string, seite inbound.Seite) (inbound.ExchangeSeite, error) {
	f.rufe(ctx, inspAufruf{op: inspOpExchanges, id: sessionID, seite: seite})
	if f.exchangesFn != nil {
		return f.exchangesFn(ctx, sessionID, seite)
	}
	return f.seite, f.err
}

func (f *inspektionFake) Coverage(ctx context.Context, baselineID string) ([]inbound.CoverageEintrag, error) {
	f.rufe(ctx, inspAufruf{op: inspOpCoverage, id: baselineID})
	if f.coverageFn != nil {
		return f.coverageFn(ctx, baselineID)
	}
	return f.eintraege, f.err
}

func (f *inspektionFake) Diagnose(ctx context.Context) (inbound.Diagnose, error) {
	f.rufe(ctx, inspAufruf{op: inspOpDiagnose})
	if f.diagnoseFn != nil {
		return f.diagnoseFn(ctx)
	}
	return f.diagnose, f.err
}

// zahlen liefert die Zahl der Aufrufe je Operation.
func (f *inspektionFake) zahlen() [inspOpAnzahl]int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.zaehler
}

// gesamt liefert die Zahl aller Aufrufe.
func (f *inspektionFake) gesamt() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.aufrufe)
}

// argumente liefert die Aufrufe mit ihren Argumenten als Kopie.
func (f *inspektionFake) argumente() []inspAufruf {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.aufrufe)
}

// ctxListe liefert den ctx jedes Aufrufs als Kopie.
func (f *inspektionFake) ctxListe() []context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxs)
}

// nurDiagonale meldet jede Abweichung der Zähler vom Orakel: genau die Operation op einmal, jede andere nie (op gleich
// inspOpAnzahl: keine).
func (f *inspektionFake) nurDiagonale(t *testing.T, name string, op inspOp) {
	t.Helper()
	for o, n := range f.zahlen() {
		soll := 0
		if inspOp(o) == op {
			soll = 1
		}
		if n != soll {
			t.Errorf("%s: Operation %d wurde %d-mal gerufen, erwartet %d", name, o, n, soll)
		}
	}
}

// inspUmgebung baut einen Server mit den Routen der Inspektion über dem Fake und dem FakeAudit der Umgebung.
func inspUmgebung(t *testing.T) (*umgebung, *inspektionFake, []admin.Route) {
	t.Helper()
	k, u := konfigMit()
	f := &inspektionFake{}
	routen, err := admin.InspektionRouten(f)
	if err != nil || len(routen) != 3 {
		t.Fatalf("InspektionRouten: %d Routen, %v", len(routen), err)
	}
	k.Routen = routen
	return neuMitKonfig(t, k, u), f, routen
}

// inspReq bildet `GET <pfad>?<query>`; ein leerer query lässt das `?` weg.
func inspReq(pfad, query string) *http.Request {
	if query != "" {
		pfad += "?" + query
	}
	return anfrageMit("GET", pfad, nil, nil)
}

// inspE7003 und inspE7002 sind die Fehler mit Code, die der Dienst für eine unbekannte Session und Baseline liefert.
func inspE7003() error { return &inbound.Fehler{Code: model.CodeE7003, Status: 404} }
func inspE7002() error { return &inbound.Fehler{Code: model.CodeE7002, Status: 404} }

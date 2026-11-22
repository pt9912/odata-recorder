package admin_test

import (
	"context"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// sessionsZiel ist der Pfad der Routen für Sessions (SPEC-292 Nr. 1).
const sessionsZiel = "/__admin/sessions"

// sessionsOp nennt die sieben Operationen von inbound.SessionUseCase in der Folge des Ports.
type sessionsOp int

const (
	opCreate sessionsOp = iota
	opOpenInitial
	opGet
	opList
	opDelete
	opBeenden
	opCloseAll
	opAnzahl
)

// sessionsFake erfüllt inbound.SessionUseCase: jede der sieben Operationen zählt ihren Aufruf und hält ctx und jedes
// Argument fest (Operation, ctx und die zwei Zeichenketten je Aufruf, in der Folge der Aufrufe); Ergebnis und Fehler sind
// einstellbar (session, sessionen, bericht, err; erzeuge, liste, beende und loesche ersetzen sie für Create, List, Beenden und
// Delete), vor ist je Operation ein Haltepunkt, der nach dem Zählen ohne die Sperre des Fakes läuft. Der Fake gerät nie in
// Panik und ist für gleichzeitigen Aufruf sicher; die Einstellungen setzt der Test vor dem ersten Aufruf.
type sessionsFake struct {
	mu      sync.Mutex
	ops     []sessionsOp
	ctxs    []context.Context
	args    [][2]string
	zaehler [opAnzahl]int

	session   model.Session
	sessionen []model.Session
	bericht   inbound.Loeschbericht
	err       error
	erzeuge   func(ctx context.Context, serviceID, id string) (model.Session, error)
	liste     func(ctx context.Context) ([]model.Session, error)
	beende    func(ctx context.Context, id string) (model.Session, error)
	loesche   func(ctx context.Context, id string) (inbound.Loeschbericht, error)
	vor       [opAnzahl]func(ctx context.Context)
}

func (f *sessionsFake) rufe(ctx context.Context, op sessionsOp, a, b string) {
	f.mu.Lock()
	f.ops = append(f.ops, op)
	f.ctxs = append(f.ctxs, ctx)
	f.args = append(f.args, [2]string{a, b})
	f.zaehler[op]++
	vor := f.vor[op]
	f.mu.Unlock()
	if vor != nil {
		vor(ctx)
	}
}

func (f *sessionsFake) Create(ctx context.Context, serviceID, id string) (model.Session, error) {
	f.rufe(ctx, opCreate, serviceID, id)
	if f.erzeuge != nil {
		return f.erzeuge(ctx, serviceID, id)
	}
	return f.session, f.err
}

func (f *sessionsFake) OpenInitialSession(ctx context.Context, serviceID string) (model.Session, error) {
	f.rufe(ctx, opOpenInitial, serviceID, "")
	return f.session, f.err
}

func (f *sessionsFake) Get(ctx context.Context, id string) (model.Session, error) {
	f.rufe(ctx, opGet, id, "")
	return f.session, f.err
}

func (f *sessionsFake) List(ctx context.Context) ([]model.Session, error) {
	f.rufe(ctx, opList, "", "")
	if f.liste != nil {
		return f.liste(ctx)
	}
	return f.sessionen, f.err
}

func (f *sessionsFake) Delete(ctx context.Context, id string) (inbound.Loeschbericht, error) {
	f.rufe(ctx, opDelete, id, "")
	if f.loesche != nil {
		return f.loesche(ctx, id)
	}
	return f.bericht, f.err
}

func (f *sessionsFake) Beenden(ctx context.Context, id string) (model.Session, error) {
	f.rufe(ctx, opBeenden, id, "")
	if f.beende != nil {
		return f.beende(ctx, id)
	}
	return f.session, f.err
}

func (f *sessionsFake) CloseAll(ctx context.Context) (inbound.CloseBericht, error) {
	f.rufe(ctx, opCloseAll, "", "")
	return inbound.CloseBericht{}, f.err
}

// zahlen liefert die Zahl der Aufrufe je Operation.
func (f *sessionsFake) zahlen() [opAnzahl]int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.zaehler
}

// gesamt liefert die Zahl aller Aufrufe.
func (f *sessionsFake) gesamt() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.ops)
}

// argumente liefert die Argumente der Aufrufe (die zwei Zeichenketten, nicht benutzte leer) als Kopie.
func (f *sessionsFake) argumente() [][2]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.args)
}

// ctxListe liefert den ctx jedes Aufrufs als Kopie.
func (f *sessionsFake) ctxListe() []context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxs)
}

// nurDiagonale meldet jede Abweichung der Zähler von dem Orakel: genau die Operation op einmal, jede andere nie.
func (f *sessionsFake) nurDiagonale(t *testing.T, name string, op sessionsOp) {
	t.Helper()
	f.nurNullAusser(t, name, op)
}

// nurNullAusser meldet jeden Zähler ungleich 1 bei den Operationen ops und ungleich 0 bei allen anderen.
func (f *sessionsFake) nurNullAusser(t *testing.T, name string, ops ...sessionsOp) {
	t.Helper()
	for o, n := range f.zahlen() {
		soll := 0
		if slices.Contains(ops, sessionsOp(o)) {
			soll = 1
		}
		if n != soll {
			t.Errorf("%s: Operation %d wurde %d-mal gerufen, erwartet %d", name, o, n, soll)
		}
	}
}

// sessionsUmgebung baut einen Server mit den Routen der Sessions über dem Fake und dem FakeAudit der Umgebung.
func sessionsUmgebung(t *testing.T) (*umgebung, *sessionsFake, []admin.Route) {
	t.Helper()
	k, u := konfigMit()
	f := &sessionsFake{}
	routen, err := admin.SessionRouten(f, u.audit)
	if err != nil || len(routen) != 4 {
		t.Fatalf("SessionRouten: %d Routen, %v", len(routen), err)
	}
	k.Routen = routen
	return neuMitKonfig(t, k, u), f, routen
}

// sessionsReq bildet einen Request an die Routen der Sessions; query steht hinter dem `?` (leer: ohne), kopf setzt Header.
func sessionsReq(methode, query string, kopf map[string][]string, body io.Reader) *http.Request {
	ziel := sessionsZiel
	if query != "" {
		ziel += "?" + query
	}
	return anfrageMit(methode, ziel, kopf, body)
}

// postReq bildet `POST /__admin/sessions` mit dem Body.
func postReq(body []byte) *http.Request {
	return sessionsReq("POST", "", nil, strings.NewReader(string(body)))
}

// closeReq bildet `POST /__admin/sessions/{id}/close` und deleteReq `DELETE /__admin/sessions/{id}`; id steht unverändert im
// Pfad, also kodiert.
func closeReq(id string) *http.Request {
	return anfrageMit("POST", sessionsZiel+"/"+id+"/close", nil, nil)
}

func deleteReq(id string) *http.Request { return anfrageMit("DELETE", sessionsZiel+"/"+id, nil, nil) }

// vierRouten nennt die vier Routen in der Folge von SPEC-292 Nr. 1 mit der Operation, die sie rufen.
func vierRouten() []struct {
	name string
	op   sessionsOp
} {
	return []struct {
		name string
		op   sessionsOp
	}{{"GET", opList}, {"POST", opCreate}, {"close", opBeenden}, {"DELETE", opDelete}}
}

// reqFuer bildet den Request der Route name (eine der vier von vierRouten): body gilt für `POST`, die Kennung der zwei
// Routen mit Pfadwert ist `ses-pfad`.
func reqFuer(name, body string) *http.Request {
	switch name {
	case "close":
		return closeReq("ses-pfad")
	case "DELETE":
		return deleteReq("ses-pfad")
	}
	return sessionsReq(name, "", nil, strings.NewReader(body))
}

// zeitA ist eine Zeit in UTC mit neun Nachkommastellen ungleich null.
func zeitA() time.Time { return time.Date(2026, 11, 6, 10, 0, 0, 123456789, time.UTC) }

// sessionA ist eine Session des Dienstes; sessionAJSON ist ihre Antwort als Literal des Vertrags.
func sessionA() model.Session {
	return model.Session{ID: "ses-1", ServiceID: "svc", CreatedAt: zeitA(), Status: model.SessionOpen}
}

const sessionAJSON = `{"id":"ses-1","service_id":"svc","created_at":"2026-11-06T10:00:00.123456789Z","status":"open"}`

// objekt als Literale des Vertrags (SPEC-109 Nr. 2, SPEC-240) für die Codes, die ein Dienst der Sessions liefert.
const (
	objektE7010 = `{"error":{"code":"ODR-E7010","message":"Session id already taken","target":"session"}}`
	objektE5005 = `{"error":{"code":"ODR-E5005","message":"No service configured for the path","target":"service"}}`
	objektE3001 = `{"error":{"code":"ODR-E3001","message":"Stored data damaged or not loadable","target":"storage"}}`
	objektE7011 = `{"error":{"code":"ODR-E7011","message":"Session is not open","target":"session"}}`
	objektE7009 = `{"error":{"code":"ODR-E7009","message":"Session is still open","target":"session"}}`
	objektE7004 = `{"error":{"code":"ODR-E7004","message":"Session is referenced by a baseline","target":"session"}}`
)

package admin_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// TestS1Konstruktor hält S1 und B1 (SPEC-292 Nr. 1): der Gutfall liefert genau die vier Routen in der Folge des Vertrags, ein
// nil-Port ist model.ErrUngueltig ohne Routen und ohne Wert im Text (der andere Port trägt einen Canary, den ein
// Formatieren des Arguments in den Text brächte), und die Funktion ruft keinen Port; die Zahl der Goroutinen fängt eine
// Goroutine, die bleibt, eine kurzlebige fängt allein die Quelle (TestS1S16QuelleDerRouten).
func TestS1Konstruktor(t *testing.T) {
	audit := supporttest.NewFakeAudit()
	fake := &sessionsFake{}
	vorher := runtime.NumGoroutine()
	var routen []admin.Route
	var err error
	for range 100 {
		routen, err = admin.SessionRouten(fake, audit)
	}
	if nachher := runtime.NumGoroutine(); err != nil || nachher > vorher+10 {
		t.Errorf("SessionRouten: %v, Goroutinen vorher %d, nachher %d", err, vorher, nachher)
	}
	got := [][2]string{}
	for _, r := range routen {
		if r.Handler == nil {
			t.Errorf("Route %s %s ohne Handler", r.Methode, r.Muster)
		}
		got = append(got, [2]string{r.Methode, r.Muster})
	}
	gleichAuf(t, "Routen in der Folge des Vertrags", got, [][2]string{
		{"GET", "/__admin/sessions"}, {"POST", "/__admin/sessions"},
		{"POST", "/__admin/sessions/{id}/close"}, {"DELETE", "/__admin/sessions/{id}"},
	})
	if fake.gesamt() != 0 || audit.Aufrufe() != 0 {
		t.Errorf("SessionRouten rief %d-mal den Dienst und %d-mal Audit, erwartet 0 und 0", fake.gesamt(), audit.Aufrufe())
	}

	mitCanary := supporttest.NewFakeAudit()
	_ = mitCanary.Append(support.Eintrag{ServiceID: canary()})
	faelle := []struct {
		name     string
		sessions inbound.SessionUseCase
		audit    support.AuditLog
	}{
		{"sessions nil", nil, mitCanary},
		{"audit nil", &sessionsFake{session: model.Session{ID: canary()}}, nil},
	}
	for _, f := range faelle {
		routen, err := admin.SessionRouten(f.sessions, f.audit)
		if routen != nil || !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v Routen, Fehler %v, erwartet keine und model.ErrUngueltig", f.name, routen, err)
			continue
		}
		if text := fmt.Sprintf("%v %+v %#v", err, err, err); strings.Contains(text, canary()) {
			t.Errorf("%s: der Text des Fehlers nennt einen Wert: %q", f.name, text)
		}
	}
	gleichAuf(t, "Zahl der Felder", len(faelle), 2)
	gleichAuf(t, "Aufrufe von Audit nach den nil-Fällen", mitCanary.Aufrufe(), 1)
}

// TestS2AufrufeUndKontext hält S2 (SPEC-292 Nr. 2): jede der zwei Routen ruft genau eine Operation von sessions (2 Routen ×
// 7 Zähler = 14 Zellen, 1 auf der Diagonale, sonst 0) mit den Argumenten unverändert; der Wert im ctx des Requests kommt
// an, und der Handler gibt ctx ohne Ableitung weiter (Gleichheit der Identität); von audit läuft allein Append.
func TestS2AufrufeUndKontext(t *testing.T) {
	faelle := []struct {
		name, methode, body string
		op                  sessionsOp
		args                [2]string
		appends             int
	}{
		{"GET", "GET", "", opList, [2]string{"", ""}, 0},
		{"POST", "POST", `{"service_id":"svc-x","id":"id-y"}`, opCreate, [2]string{"svc-x", "id-y"}, 1},
	}
	zellen := 0
	for _, f := range faelle {
		u, fake, routen := sessionsUmgebung(t)
		fake.session = sessionA()
		r := sessionsReq(f.methode, "", nil, strings.NewReader(f.body))
		r = r.WithContext(context.WithValue(r.Context(), ctxSchluessel{}, "wert-"+f.name))
		if rec := u.rufe(r); rec.Code/100 != 2 {
			t.Fatalf("%s: Status %d", f.name, rec.Code)
		}
		zellen += len(fake.zahlen())
		fake.nurDiagonale(t, f.name, f.op)
		gleichAuf(t, f.name+": Argumente", fake.argumente(), [][2]string{f.args})
		if got := fake.ctxListe()[0].Value(ctxSchluessel{}); got != "wert-"+f.name {
			t.Errorf("%s: der Wert im ctx des Requests kam als %v an", f.name, got)
		}
		if n := u.audit.Aufrufe(); n != f.appends {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet %d", f.name, n, f.appends)
		}

		ctx := context.WithValue(context.Background(), ctxSchluessel{}, "direkt")
		i := map[string]int{"GET": 0, "POST": 1}[f.name]
		_, _ = routen[i].Handler(ctx, admin.Anfrage{Methode: f.methode, Body: strings.NewReader(f.body), KorrelationsID: "cor-1"})
		if got := fake.ctxListe(); len(got) != 2 || got[1] != ctx {
			t.Errorf("%s: der Handler reichte einen anderen ctx weiter als den, den er erhielt", f.name)
		}
	}
	gleichAuf(t, "Zellen (Routen × Zähler)", zellen, 2*int(opAnzahl))
	gleichAuf(t, "Zähler gleich Zahl der Operationen des Ports", int(opAnzahl), 7)
}

// TestS3Zuordnung hält S3 und B1 (SPEC-292 Nr. 1, SPEC-238 Nr. 2): M-Route, drei Pfadformen mal sieben Methoden, 21 Fälle:
// `GET` und `POST /__admin/sessions`, `POST …/{id}/close` und `DELETE …/{id}` treffen (4), die 17 anderen sind 405 mit dem
// `Allow` der Pfadform (`GET, POST`; `DELETE`; `POST`, als Literale) und rufen den Dienst nie. Zählprüfung: die Pfadformen
// sind die verschiedenen Muster der gelieferten Routen (3) und die Routen die vier der Liste.
func TestS3Zuordnung(t *testing.T) {
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	formen := []struct {
		pfad, allow string
		treffer     map[string]sessionsOp
	}{
		{"/__admin/sessions", "GET, POST", map[string]sessionsOp{"GET": opList, "POST": opCreate}},
		{"/__admin/sessions/ses-pfad", "DELETE", map[string]sessionsOp{"DELETE": opDelete}},
		{"/__admin/sessions/ses-pfad/close", "POST", map[string]sessionsOp{"POST": opBeenden}},
	}
	faelle, treffer := 0, 0
	for _, f := range formen {
		for _, m := range methoden {
			faelle++
			u, fake, _ := sessionsUmgebung(t)
			fake.session = sessionA()
			rec := u.rufe(anfrageMit(m, f.pfad, nil, strings.NewReader(`{"service_id":"svc"}`)))
			if op, ok := f.treffer[m]; ok {
				treffer++
				if rec.Code/100 != 2 {
					t.Errorf("%s %s: Status %d", m, f.pfad, rec.Code)
				}
				fake.nurDiagonale(t, m+" "+f.pfad, op)
				continue
			}
			pruefeFehlerAntwort(t, m+" "+f.pfad, rec, 405, objekt404, f.allow)
			if codes, _ := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || fake.gesamt() != 0 || u.audit.Aufrufe() != 0 {
				t.Errorf("%s %s: CountError %q, %d Aufrufe des Dienstes, %d von Audit", m, f.pfad, codes, fake.gesamt(), u.audit.Aufrufe())
			}
		}
	}
	_, _, routen := sessionsUmgebung(t)
	muster := map[string]bool{}
	for _, r := range routen {
		muster[r.Muster] = true
	}
	gleichAuf(t, "Fälle, Treffer, Pfadformen, Routen", []int{faelle, treffer, len(muster), len(routen)}, []int{21, 4, len(formen), 4})
	gleichAuf(t, "Pfadformen", len(muster), 3)
}

// zaehlLeser liest aus Bytes und zählt jeden Aufruf von Read.
type zaehlLeser struct {
	*bytes.Reader
	n atomic.Int64
}

func (z *zaehlLeser) Read(p []byte) (int, error) {
	z.n.Add(1)
	return z.Reader.Read(p)
}

// TestS4QueryUndBody hält S4 (SPEC-292 Nr. 3): keine der zwei Routen liest die Query, jede Form (6 Queries × 2 Routen = 12
// Fälle) ist ohne Wirkung gegenüber dem Request ohne Query; `GET` liest den Body nie (3 ungültige Bodies, 0 Lesezugriffe).
func TestS4QueryUndBody(t *testing.T) {
	queries := []string{"", "limit=x", "a=1&a=2", "%zz", "%", "a=1;b=2"}
	faelle := 0
	for _, q := range queries {
		u, fake, _ := sessionsUmgebung(t)
		fake.session, fake.sessionen = sessionA(), []model.Session{sessionA()}
		rec := u.rufe(sessionsReq("GET", q, nil, nil))
		if rec.Code != 200 || rec.Body.String() != `{"sessions":[`+sessionAJSON+`]}` {
			t.Errorf("GET mit Query %q: %d %q", q, rec.Code, rec.Body.String())
		}
		rec = u.rufe(sessionsReq("POST", q, nil, strings.NewReader(`{"service_id":"svc","id":"i"}`)))
		if rec.Code != 201 || rec.Body.String() != sessionAJSON {
			t.Errorf("POST mit Query %q: %d %q", q, rec.Code, rec.Body.String())
		}
		gleichAuf(t, "Argumente mit Query "+q, fake.argumente(), [][2]string{{"", ""}, {"svc", "i"}})
		fake.nurNullAusser(t, "Query "+q, opList, opCreate)
		faelle += 2
	}
	bodies := [][]byte{[]byte("{"), []byte("[1,2"), {0xFF, 0xFE, 0x00}}
	for i, b := range bodies {
		u, fake, _ := sessionsUmgebung(t)
		leser := &zaehlLeser{Reader: bytes.NewReader(b)}
		var body io.Reader = leser
		if rec := u.rufe(sessionsReq("GET", "", nil, body)); rec.Code != 200 || rec.Body.String() != `{"sessions":[]}` || leser.n.Load() != 0 || fake.gesamt() != 1 {
			t.Errorf("GET mit ungültigem Body Nr. %d: Status %d, %d Lesezugriffe, %d Aufrufe", i, rec.Code, leser.n.Load(), fake.gesamt())
		}
		faelle++
	}
	gleichAuf(t, "Fälle", faelle, len(queries)*2+len(bodies))
	gleichAuf(t, "Fälle laut Plan", faelle, 15)
}

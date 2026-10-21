package record_test

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Die Fälle des Dienstes tragen Geheimnisse an jeder redigierten Stelle (Authorization, Cookie,
// Query token, Body $..password, Antwort Set-Cookie und $..password) und Marker an Stellen, die
// nur das Protokoll nie trägt (Pfad, Query q, Header X-Info).
const (
	reqJSON = `{"password":"` + geheim + `","a":1}`
	resJSON = `{"password":"` + geheim + `"}`
	reqRed  = `{"password":"[REDACTED]","a":1}`
	resRed  = `{"password":"[REDACTED]"}`
)

func geheimeAnfrage(typ, body string) inbound.Anfrage {
	return inbound.Anfrage{
		Methode: "POST", Pfad: "/a/PFAD-7f", RawQuery: "token=" + geheim + "&q=Q-7f", KorrelationsID: "cor-1",
		Header: map[string][]string{"Authorization": {"Bearer " + geheim}, "Cookie": {"sid=" + geheim},
			"X-Info": {"H-7f"}, "Content-Type": {typ}},
		Body: strings.NewReader(body), Laenge: int64(len(body)),
	}
}

func geheimeAntwort(b *fakeBody) *plan {
	return antwort(200, map[string][]string{"Set-Cookie": {"sid=" + geheim}, "Content-Type": {"application/json"}}, b)
}

// soll sind die erwarteten Felder eines angehängten Exchange; req und res sind die Inhalte der
// Blobs, leer ohne Blob.
type soll struct {
	capture              model.CaptureStatus
	reqBody, resBody     model.BodyState
	reqRedact, resRedact model.RedactionStatus
	integrity            model.BodyIntegrity
	equivalence          model.ReplayEquivalence
	status               int
	req, res             string
	key                  bool
}

type dienstFall struct {
	name string
	lauf func(t *testing.T, u *umgebung)
	// soll ist nil, wenn der Exchange nicht angehängt ist.
	soll *soll
}

func normal(t *testing.T, u *umgebung, a inbound.Anfrage) {
	t.Helper()
	ant, err := u.record(context.Background(), a)
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	liesUndSchliesse(t, ant.Body)
}

// ohneKopf führt einen Request aus, dessen Senden auf das Ende von ctx wartet, und beendet ihn mit
// beende; liest der Fake den Body ganz, wartet es vorher auf den Commit des Requests.
func ohneKopf(t *testing.T, u *umgebung, lies int, beende func(abbrechen func())) {
	t.Helper()
	u.upstream.plane(&plan{lies: lies, warteAufCtx: true})
	ctx, abbrechen := context.WithCancel(context.Background())
	defer abbrechen()
	res := u.recordNebenbei(ctx, geheimeAnfrage("application/json", reqJSON))
	empfange(t, "Senden", u.upstream.imSenden)
	if lies < 0 {
		empfange(t, "Commit des Requests", u.blobs.commits)
	}
	beende(abbrechen)
	if r := empfange(t, "Record", res); r.err == nil {
		t.Error("Record ohne Fehler")
	}
}

// drainNachKopf beendet einen Exchange mit Antwort über Drain(s1, 0); vorher liest der Test, was
// lies vorsieht, beim JSON-Request nach dem Commit des Requests.
func drainNachKopf(t *testing.T, u *umgebung, typ, body string, b *fakeBody, lies func(io.Reader)) {
	t.Helper()
	u.upstream.plane(geheimeAntwort(b))
	ant, err := u.record(context.Background(), geheimeAnfrage(typ, body))
	if err != nil {
		t.Fatal(err)
	}
	if typ == "application/json" {
		empfange(t, "Commit des Requests", u.blobs.commits)
	}
	lies(ant.Body)
	d := u.drainNebenbei(context.Background(), "s1", 0)
	empfange(t, "Ende des Kontexts", u.upstream.gesehen()[0].done)
	schliesse(t, ant.Body)
	empfange(t, "Drain", d)
}

func ohneAntwortSoll(c model.CaptureStatus) *soll {
	s := &soll{capture: c, reqBody: model.BodyPresent, resBody: model.BodyAbsent, reqRedact: model.RedactionApplied,
		resRedact: model.RedactionNone, integrity: model.IntegrityVerified, equivalence: model.EquivalenceNone,
		req: reqRed, key: true}
	if c == model.CaptureFailed {
		s.reqBody, s.reqRedact, s.integrity, s.req, s.key = model.BodyDiscarded, model.RedactionUnsafe, model.IntegrityAbsent, "", false
	}
	return s
}

func requestVerworfen() *soll {
	return &soll{capture: model.CaptureFailed, reqBody: model.BodyDiscarded, resBody: model.BodyPresent,
		reqRedact: model.RedactionUnsafe, resRedact: model.RedactionApplied, integrity: model.IntegrityVerified,
		equivalence: model.EquivalenceNone, status: 200, res: resRed}
}

func dienstFaelle() []dienstFall {
	gut := &soll{capture: model.CaptureComplete, reqBody: model.BodyPresent, resBody: model.BodyPresent,
		reqRedact: model.RedactionApplied, resRedact: model.RedactionApplied, integrity: model.IntegrityVerified,
		equivalence: model.EquivalenceSemantic, status: 200, req: reqRed, res: resRed, key: true}
	responseAbbruch := &soll{capture: model.CaptureIncomplete, reqBody: model.BodyPresent, resBody: model.BodyDiscarded,
		reqRedact: model.RedactionApplied, resRedact: model.RedactionUnsafe, integrity: model.IntegrityVerified,
		equivalence: model.EquivalenceNone, status: 200, req: reqRed, key: true}
	halb := func() *fakeBody { return koerper(`{"password":"`).dann(make(chan struct{}), geheim+`"}`, nil) }
	lies5 := func(r io.Reader) { _, _ = r.Read(make([]byte, 5)) }
	return []dienstFall{
		{"Gutfall", func(t *testing.T, u *umgebung) {
			u.upstream.plane(geheimeAntwort(koerper(resJSON)))
			normal(t, u, geheimeAnfrage("application/json", reqJSON))
		}, gut},
		{"typ", func(t *testing.T, u *umgebung) {
			u.upstream.plane(geheimeAntwort(koerper(resJSON)))
			normal(t, u, geheimeAnfrage("text/plain", "password="+geheim))
		}, requestVerworfen()},
		{"syntax", func(t *testing.T, u *umgebung) {
			u.upstream.plane(geheimeAntwort(koerper(resJSON)))
			normal(t, u, geheimeAnfrage("application/json", `{"password":"`+geheim+`",}`))
		}, requestVerworfen()},
		{"zu_gross", func(t *testing.T, u *umgebung) {
			u.upstream.plane(geheimeAntwort(koerper(resJSON)))
			normal(t, u, geheimeAnfrage("application/json", `{"password":"`+geheim+`","x":"`+strings.Repeat("x", 300)+`"}`))
		}, requestVerworfen()},
		{"Abbruch im Request", func(t *testing.T, u *umgebung) {
			p := geheimeAntwort(koerper(resJSON))
			p.lies = 5
			u.upstream.plane(p)
			normal(t, u, geheimeAnfrage("application/json", reqJSON))
		}, requestVerworfen()},
		{"Abbruch in der Response", func(t *testing.T, u *umgebung) {
			u.upstream.plane(geheimeAntwort(halb()))
			ant, err := u.record(context.Background(), geheimeAnfrage("application/json", reqJSON))
			if err != nil {
				t.Fatal(err)
			}
			lies5(ant.Body)
			schliesse(t, ant.Body)
		}, responseAbbruch},
		{"ohne Antwort, nicht erreichbar", func(t *testing.T, u *umgebung) {
			u.upstream.plane(&plan{lies: -1, fehler: outbound.ErrUpstreamNichtErreichbar})
			_, _ = u.record(context.Background(), geheimeAnfrage("application/json", reqJSON))
		}, ohneAntwortSoll(model.CaptureIncomplete)},
		{"ohne Antwort, Zeit", func(t *testing.T, u *umgebung) {
			u.upstream.plane(&plan{lies: -1, fehler: outbound.ErrUpstreamZeit})
			_, _ = u.record(context.Background(), geheimeAnfrage("application/json", reqJSON))
		}, ohneAntwortSoll(model.CaptureIncomplete)},
		{"ohne Antwort, anderer Fehler mit Geheimnis im Text", func(t *testing.T, u *umgebung) {
			u.upstream.plane(&plan{lies: 0, fehler: errDisk})
			_, _ = u.record(context.Background(), geheimeAnfrage("application/json", reqJSON))
		}, ohneAntwortSoll(model.CaptureFailed)},
		{"ohne Antwort, Client-Abbruch", func(t *testing.T, u *umgebung) {
			ohneKopf(t, u, -1, func(abbrechen func()) { abbrechen() })
		}, ohneAntwortSoll(model.CaptureIncomplete)},
		{"ohne Antwort, Drain", func(t *testing.T, u *umgebung) {
			ohneKopf(t, u, -1, func(func()) { u.drainNebenbei(context.Background(), "s1", 0) })
		}, ohneAntwortSoll(model.CaptureIncomplete)},
		{"Drain, Request läuft", func(t *testing.T, u *umgebung) {
			ohneKopf(t, u, 5, func(func()) { u.drainNebenbei(context.Background(), "s1", 0) })
		}, ohneAntwortSoll(model.CaptureFailed)},
		{"Drain, Response läuft", func(t *testing.T, u *umgebung) {
			drainNachKopf(t, u, "application/json", reqJSON, halb(), lies5)
		}, responseAbbruch},
		{"Drain, beide Ergebnisse", func(t *testing.T, u *umgebung) {
			drainNachKopf(t, u, "application/json", reqJSON, koerper(resJSON), func(r io.Reader) {
				_, _ = io.ReadAll(r)
				empfange(t, "Commit der Response", u.blobs.commits)
			})
		}, gut},
		{"Drain, Request verworfen, Response läuft", func(t *testing.T, u *umgebung) {
			drainNachKopf(t, u, "text/plain", "password="+geheim, halb(), lies5)
		}, &soll{capture: model.CaptureFailed, reqBody: model.BodyDiscarded, resBody: model.BodyDiscarded,
			reqRedact: model.RedactionUnsafe, resRedact: model.RedactionUnsafe, integrity: model.IntegrityAbsent,
			equivalence: model.EquivalenceNone, status: 200}},
		{"nicht angehängt", func(t *testing.T, u *umgebung) {
			u.exchanges.appendFehler[0] = errDisk
			u.upstream.plane(geheimeAntwort(koerper(resJSON)))
			normal(t, u, geheimeAnfrage("application/json", reqJSON))
		}, nil},
	}
}

func pruefeSoll(t *testing.T, u *umgebung, name string, s *soll) {
	t.Helper()
	l := u.exchanges.angehaengt()
	if s == nil {
		if len(l) != 0 {
			t.Errorf("%s: %d angehängt", name, len(l))
		}
		return
	}
	if len(l) != 1 {
		t.Fatalf("%s: %d angehängt", name, len(l))
	}
	e := l[0]
	req, _ := u.blobs.inhalt(e.RequestBlob)
	res, _ := u.blobs.inhalt(e.ResponseBlob)
	got := soll{e.Capture, e.RequestBody, e.ResponseBody, e.RequestRedact, e.ResponseRedact, e.Integrity,
		e.Equivalence, e.ResponseStatus, req, res, e.MatchKey != ""}
	if got != *s {
		t.Errorf("%s:\n%+v\nerwartet\n%+v", name, got, *s)
	}
}

// TestAnhaengenNachCommit: in jedem Fall steht jedes Append nach dem Commit jedes Blobs, auf den
// der Exchange verweist, und danach kommt kein Aufruf für dieselbe Kennung; jeder Exchange trägt die
// von Hand geschriebenen Felder seines Falls, darunter die Blobs (B17).
func TestAnhaengenNachCommit(t *testing.T) {
	for _, f := range dienstFaelle() {
		u := neu(t)
		f.lauf(t, u)
		pruefeSoll(t, u, f.name, f.soll)
		ereignisse := u.prot.liste()
		for _, e := range u.exchanges.angehaengt() {
			app := slices.Index(ereignisse, "Append "+e.ID)
			for _, ref := range []string{e.RequestBlob, e.ResponseBlob} {
				if ref == "" {
					continue
				}
				c := slices.Index(ereignisse, "Commit "+ref)
				if _, da := u.blobs.inhalt(ref); !da || c < 0 || c > app {
					t.Errorf("%s: Commit %d nach Append %d (%v)", f.name, c, app, ereignisse)
				}
			}
		}
		if len(u.exchanges.unerwartet) > 0 || len(u.sessions.unerwartet) > 0 {
			t.Errorf("%s: unerwartete Aufrufe %v %v", f.name, u.exchanges.unerwartet, u.sessions.unerwartet)
		}
	}
}

// TestKeinGeheimnis: in keinem Fall erreicht ein Geheimnis an redigierter Stelle Blob, Exchange
// oder Protokoll, und das Protokoll trägt nie Pfad, Query, Header oder einen Fehlertext: jeder
// Eintrag hält die geschlossenen Mengen von pruefeWerte (B24).
func TestKeinGeheimnis(t *testing.T) {
	for _, f := range dienstFaelle() {
		u := neu(t)
		f.lauf(t, u)
		if strings.Contains(u.blobs.alleAngeboten(), geheim) {
			t.Errorf("%s: Geheimnis im Blob-Fake", f.name)
		}
		for _, e := range u.exchanges.angehaengt() {
			if strings.Contains(fmt.Sprintf("%+v", e), geheim) {
				t.Errorf("%s: Geheimnis im Exchange %+v", f.name, e)
			}
		}
		for _, e := range u.log.alle() {
			text := fmt.Sprintf("%s %v", e.text, e.attrs)
			for _, m := range []string{geheim, "PFAD-7f", "Q-7f", "H-7f"} {
				if strings.Contains(text, m) {
					t.Errorf("%s: %s im Protokoll: %s", f.name, m, text)
				}
			}
		}
		pruefeWerte(t, f.name, u.log.alle())
		if g := u.upstream.gesehen(); len(g) > 0 && !strings.Contains(fmt.Sprint(g[0].a.Header, g[0].a.RawQuery, g[0].body), geheim) {
			t.Errorf("%s: Gegenprobe, der Upstream sah das Geheimnis nicht", f.name)
		}
	}
}

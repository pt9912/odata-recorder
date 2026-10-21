package record_test

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
)

// TestDienstFuer: Root gleich oder an der Segmentgrenze, byteweise, der längste; leerer Rest ist / (B3).
func TestDienstFuer(t *testing.T) {
	ab := []record.Dienst{{ID: "ab", Root: "/a/b"}, {ID: "abc", Root: "/a/bc"}}
	wurzel := []record.Dienst{{ID: "w", Root: "/"}}
	beide := []record.Dienst{{ID: "w", Root: "/"}, {ID: "a", Root: "/a"}}
	faelle := []struct {
		dienste      []record.Dienst
		pfad, id, od string
	}{
		{ab, "/a/b", "ab", "/"},
		{ab, "/a/b/", "ab", "/"},
		{ab, "/a/b/c", "ab", "/c"},
		{ab, "/a/bc/x", "abc", "/x"},
		{ab, "/a/bcd", "", ""},
		{ab, "/A/b", "", ""},
		{ab, "/%61/b", "", ""},
		{ab, "/a", "", ""},
		{ab, "", "", ""},
		{ab, "*", "", ""},
		{wurzel, "/x", "w", "/x"},
		{wurzel, "/", "w", "/"},
		{wurzel, "*", "", ""},
		{beide, "/a/x", "a", "/x"},
		{beide, "/ab", "w", "/ab"},
	}
	for _, f := range faelle {
		d, od, ok := record.DienstFuer(f.dienste, f.pfad)
		if ok != (f.id != "") || d.ID != f.id || od != f.od {
			t.Errorf("%q: %q %q %v, erwartet %q %q", f.pfad, d.ID, od, ok, f.id, f.od)
		}
	}
	dreh := []record.Dienst{beide[1], beide[0]}
	if d, od, _ := record.DienstFuer(dreh, "/a/x"); d.ID != "a" || od != "/x" {
		t.Errorf("Reihenfolge der Dienste: %q %q", d.ID, od)
	}
}

// TestDienstUnbekannt: ODR-E5005 ohne Ablage und Upstream (B3).
func TestDienstUnbekannt(t *testing.T) {
	u := neu(t)
	_, err := u.record(context.Background(), anfrage("GET", "/c/x"))
	pruefeCode(t, err, model.CodeE5005, 404)
	if u.sessions.anzahlAufrufe() != 0 || len(u.upstream.gesehen()) != 0 {
		t.Error("Ablage oder Upstream gerufen")
	}
}

func reject(k *record.Konfig) { k.Regeln.InvalidEncoding = model.KodierungReject }

// TestKodierung: nach dem Service und vor der Session, ODR-E5006 bei reject, literal leitet weiter;
// ErrUngueltig ist ODR-E9001 (B4).
func TestKodierung(t *testing.T) {
	query := anfrage("GET", "/a/x")
	query.RawQuery = "a=%zz"
	pfad := anfrage("GET", "/a/p%zz")
	for name, a := range map[string]inbound.Anfrage{"Query": query, "Pfad": pfad} {
		u := neu(t, reject)
		_, err := u.record(context.Background(), a)
		pruefeCode(t, err, model.CodeE5006, 400)
		if u.sessions.anzahlAufrufe() != 0 || len(u.upstream.gesehen()) != 0 || u.ids.ExchangeID() != kennung("exc", 1) {
			t.Errorf("%s: Ablage, Upstream oder Kennung berührt", name)
		}
		l := neu(t)
		l.rundlauf(a)
		if e := l.letzter(); e.MatchKey == "" || len(l.upstream.gesehen()) != 1 {
			t.Errorf("%s literal: Key %q", name, e.MatchKey)
		}
	}
	u := neu(t, reject)
	ohneDienst := anfrage("GET", "/c/x")
	ohneDienst.RawQuery = "a=%zz"
	_, err := u.record(context.Background(), ohneDienst)
	pruefeCode(t, err, model.CodeE5005, 404)
	u.sessions.setze("s1", "svc-a", model.SessionComplete)
	_, err = u.record(context.Background(), query)
	pruefeCode(t, err, model.CodeE5006, 400)
	w := neu(t, func(k *record.Konfig) { k.Dienste = []record.Dienst{{ID: "svc-a", Root: "/", UpstreamURL: "http://a"}} })
	_, err = w.record(context.Background(), anfrage("GET", "/a?b"))
	pruefeCode(t, err, model.CodeE9001, 500)
}

func mitSession(a inbound.Anfrage, name string, werte ...string) inbound.Anfrage {
	a.Header = map[string][]string{name: werte}
	return a
}

const sessionHeader = "X-Recorder-Recording-Session"

// TestSessionMitHeader: genau ein nichtleerer Wert über alle Schreibweisen; Rangfolge unbekannt,
// fremd, geschlossen oder im Drain; der Header wählt immer (B5).
func TestSessionMitHeader(t *testing.T) {
	sessions := func(u *umgebung) {
		u.sessions.setze("s2", "svc-a", model.SessionOpen)
		u.sessions.setze("s3", "svc-a", model.SessionComplete)
		u.sessions.setze("s4", "svc-b", model.SessionOpen)
		u.sessions.setze("s5", "svc-b", model.SessionComplete)
	}
	get := anfrage("GET", "/a/x")
	faelle := []struct {
		name    string
		a       inbound.Anfrage
		nurS1   bool
		session string
		code    model.Code
	}{
		{"s2", mitSession(get, sessionHeader, "s2"), false, "s2", 0},
		{"nur s1 offen, Header s3", mitSession(get, sessionHeader, "s3"), true, "", model.CodeE5009},
		{"klein geschrieben", mitSession(get, "x-recorder-recording-session", "s2"), false, "s2", 0},
		{"zwei Werte", mitSession(get, sessionHeader, "s1", "s2"), false, "", model.CodeE5003},
		{"leerer Wert", mitSession(get, sessionHeader, ""), false, "", model.CodeE5003},
		{"leere Wertliste", mitSession(get, sessionHeader), true, "s1", 0},
		{"unbekannt", mitSession(get, sessionHeader, "unbekannt"), false, "", model.CodeE7003},
		{"s4 fremd", mitSession(get, sessionHeader, "s4"), false, "", model.CodeE5004},
		{"s5 fremd und geschlossen", mitSession(get, sessionHeader, "s5"), false, "", model.CodeE5004},
		{"s3 geschlossen", mitSession(get, sessionHeader, "s3"), false, "", model.CodeE5009},
	}
	zweiSchreibweisen := get
	zweiSchreibweisen.Header = map[string][]string{sessionHeader: {"s1"}, "x-recorder-recording-session": {"s2"}}
	faelle = append(faelle, struct {
		name    string
		a       inbound.Anfrage
		nurS1   bool
		session string
		code    model.Code
	}{"zwei Schreibweisen", zweiSchreibweisen, false, "", model.CodeE5003})
	for _, f := range faelle {
		u := neu(t)
		if !f.nurS1 {
			sessions(u)
		} else {
			u.sessions.setze("s3", "svc-a", model.SessionComplete)
		}
		pruefeWahl(t, u, f.name, f.a, f.session, f.code)
	}
	u := neu(t)
	if _, err := u.r.Drain(context.Background(), "s1", 0); err != nil {
		t.Fatal(err)
	}
	pruefeWahl(t, u, "s1 nach Drain", mitSession(get, sessionHeader, "s1"), "", model.CodeE5009)
}

// pruefeWahl führt a aus: mit session wird der Exchange dort angehängt, mit code ist es der
// Fehler ohne Weiterleitung und ohne Exchange.
func pruefeWahl(t *testing.T, u *umgebung, name string, a inbound.Anfrage, session string, code model.Code) {
	t.Helper()
	if session != "" {
		u.rundlauf(a)
		if e := u.letzter(); e.SessionID != session {
			t.Errorf("%s: Session %q, erwartet %q", name, e.SessionID, session)
		}
		return
	}
	_, err := u.record(context.Background(), a)
	var f *inbound.Fehler
	if !errors.As(err, &f) || f.Code != code {
		t.Errorf("%s: %v, erwartet %v", name, err, code)
	}
	if len(u.upstream.gesehen()) != 0 || len(u.exchanges.angehaengt()) != 0 {
		t.Errorf("%s: weitergeleitet oder angehängt", name)
	}
}

// TestSessionOhneHeader: die eine offene des Service, die nicht im Drain ist (B6).
func TestSessionOhneHeader(t *testing.T) {
	get := anfrage("GET", "/a/x")
	faelle := []struct {
		name    string
		vorher  func(u *umgebung)
		session string
		code    model.Code
	}{
		{"keine offene", func(u *umgebung) { u.sessions.setze("s1", "svc-a", model.SessionComplete) }, "", model.CodeE5009},
		{"s1 offen", func(*umgebung) {}, "s1", 0},
		{"s1 und s2 offen", func(u *umgebung) { u.sessions.setze("s2", "svc-a", model.SessionOpen) }, "", model.CodeE5003},
		{"s2 im Drain", func(u *umgebung) {
			u.sessions.setze("s2", "svc-a", model.SessionOpen)
			if _, err := u.r.Drain(context.Background(), "s2", 0); err != nil {
				t.Fatal(err)
			}
		}, "s1", 0},
		{"s3 geschlossen", func(u *umgebung) { u.sessions.setze("s3", "svc-a", model.SessionComplete) }, "s1", 0},
		{"nur s4 offen", func(u *umgebung) {
			u.sessions.setze("s1", "svc-a", model.SessionComplete)
			u.sessions.setze("s4", "svc-b", model.SessionOpen)
		}, "", model.CodeE5009},
	}
	for _, f := range faelle {
		u := neu(t)
		f.vorher(u)
		pruefeWahl(t, u, f.name, get, f.session, f.code)
	}
}

// TestSessionFehler: Fehler der Ablage als Code mit Gegenstand Session, Kontextfehler unverändert (B7).
func TestSessionFehler(t *testing.T) {
	for _, mitHeader := range []bool{true, false} {
		for _, f := range []struct {
			err    error
			code   model.Code
			status int
		}{{outbound.ErrBeschaedigt, model.CodeE3001, 500}, {errDisk, model.CodeE9001, 500}, {context.Canceled, 0, 0}} {
			u := neu(t)
			a := anfrage("GET", "/a/x")
			if mitHeader {
				a = mitSession(a, sessionHeader, "s1")
				u.sessions.getFehler = f.err
			} else {
				u.sessions.listFehler = f.err
			}
			_, err := u.record(context.Background(), a)
			if f.code == 0 {
				if err != f.err {
					t.Errorf("Header %v: %v, erwartet denselben Fehlerwert", mitHeader, err)
				}
				continue
			}
			pruefeCode(t, err, f.code, f.status)
			if !errors.Is(err, f.err) {
				t.Errorf("Header %v: Ursache fehlt in %v", mitHeader, err)
			}
		}
	}
}

// TestAnnahmeGegenDrain: Drain und Annahme unter einer Sperre; Kennung und Zeit erst nach der
// Annahme (B8).
func TestAnnahmeGegenDrain(t *testing.T) {
	u := neu(t)
	u.sessions.haltList = true
	res := u.recordNebenbei(context.Background(), anfrage("GET", "/a/x"))
	empfange(t, "List", u.sessions.imList)
	if _, err := u.r.Drain(context.Background(), "s1", 0); err != nil {
		t.Fatal(err)
	}
	close(u.sessions.freiList)
	pruefeCode(t, empfange(t, "Record", res).err, model.CodeE5009, 409)
	if len(u.upstream.gesehen()) != 0 {
		t.Error("nach Drain weitergeleitet")
	}

	synctest.Test(t, func(t *testing.T) {
		u := neu(t)
		halt := make(chan struct{})
		u.upstream.plane(&plan{lies: -1, halt: halt, status: 200, body: koerper("x")})
		res := u.recordNebenbei(context.Background(), anfrage("GET", "/a/x"))
		<-u.upstream.imSenden
		d := u.drainNebenbei(context.Background(), "s1", time.Hour)
		synctest.Wait()
		close(halt)
		r := <-res
		synctest.Wait()
		select {
		case <-d:
			t.Fatal("Drain kehrte vor Close zurück")
		default:
		}
		_, _ = r.ant.Body.Read(make([]byte, 8))
		if err := r.ant.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if b := <-d; b.err != nil || b.b.Abgebrochen != 0 {
			t.Errorf("Drain: %+v", b)
		}
	})

	k := neu(t)
	k.sessions.beimLesen = func() { k.uhr.Advance(time.Second) }
	for _, p := range []string{"/c/x", "/a/x", "/a/x"} {
		a := anfrage("GET", p)
		if p == "/a/x" {
			a = mitSession(a, sessionHeader, "unbekannt")
		}
		if _, err := k.record(context.Background(), a); err == nil {
			t.Fatal("abgewiesener Request angenommen")
		}
	}
	k.rundlauf(anfrage("GET", "/a/x"))
	e := k.letzter()
	if e.ID != kennung("exc", 1) || !e.StartedAt.Equal(t0().Add(3*time.Second)) {
		t.Errorf("Kennung %s, StartedAt %v", e.ID, e.StartedAt)
	}
}

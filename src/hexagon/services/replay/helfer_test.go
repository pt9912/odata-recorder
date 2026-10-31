package replay_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// Kennungen der Fixtures: paarweise verschieden, damit eine Vertauschung den Wert ändert.
const (
	wsID  = "wsp-a"
	repID = "rep-a"
	sesID = "ses-a"
	svcID = "svc-a"
	root  = "/svc"
)

// flüchtig sind die sieben Default-Namen von VOLATILE_HEADERS (SPEC-063).
func fluechtig() []string {
	return []string{"Date", "Server", "Via", "Age", "X-Request-Id", "X-Correlation-Id", "Traceparent"}
}

// umg ist eine Testumgebung: die Fakes, ihr gemeinsames Protokoll und die Konfiguration.
type umg struct {
	t    *testing.T
	p    *prot
	sess *sessions
	ex   *exchanges
	bl   *blobs
	st   *store
	tel  *supporttest.FakeTelemetry
	k    replay.Konfig
}

// neu baut eine Umgebung mit einem Dienst svc-a unter /svc, einer vollständigen Session ses-a und dem Workspace
// wsp-a desselben Dienstes.
func neu(t *testing.T) *umg {
	t.Helper()
	p := neuesProt(t)
	sess := &sessions{p: p, m: map[string]model.Session{}}
	st := &store{p: p, sess: sess, z: leererZustand()}
	u := &umg{t: t, p: p, sess: sess, st: st, tel: supporttest.NewFakeTelemetry(),
		ex: &exchanges{p: p, liste: map[string][]model.Exchange{}}, bl: &blobs{p: p, inhalt: map[string][]byte{}}}
	u.k = replay.Konfig{
		Dienste: []replay.Dienst{{ID: svcID, Root: root}}, Sessions: sess, Exchanges: u.ex, Blobs: u.bl, Transaktion: st,
		Regeln:   model.SchluesselRegeln{VolatileHeaders: fluechtig(), InvalidEncoding: model.KodierungLiteral, BodyMode: model.BodyBytes},
		MaxBytes: 1 << 20, Strategie: model.StrategySequential, Telemetrie: u.tel,
	}
	u.sess.setze(model.Session{ID: sesID, ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionComplete})
	u.st.z.ws[wsID] = model.Workspace{ID: wsID, ServiceID: svcID, BaselineSnapshotID: "snap-a", CreatedAt: zeit()}
	return u
}

func zeit() time.Time { return time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC) }

// replayer baut den Dienst aus der aktuellen Konfiguration.
func (u *umg) replayer() *replay.Replayer {
	u.t.Helper()
	r, err := replay.New(u.k)
	if err != nil {
		u.t.Fatalf("New: %v", err)
	}
	return r
}

// replay ruft den Dienst mit einem frischen Replayer.
func (u *umg) replay(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	u.t.Helper()
	return u.replayer().Replay(ctx, a)
}

// anf ist der Standard-Request: GET /svc/Products mit den drei Steuerheadern.
func anf() inbound.Anfrage {
	return inbound.Anfrage{
		Methode: "GET", Pfad: root + "/Products", KorrelationsID: "cor-a", Laenge: -1,
		Header: map[string][]string{
			"X-Recorder-Workspace":         {wsID},
			"X-Recorder-Recording-Session": {sesID},
			"X-Recorder-Replay-Session":    {repID},
		},
	}
}

// key bildet den Key, unter dem eine Aufzeichnung des Requests a (ohne Body) mit dem Dienst d und dem
// OData-Pfad odata steht, unabhängig vom Dienst.
func (u *umg) keyFuer(a inbound.Anfrage, dienst, odata string) string {
	u.t.Helper()
	s, err := model.BildeSchluessel(model.SchluesselEingabe{
		Method: a.Methode, ServiceID: dienst, Path: odata, RawQuery: a.RawQuery, Header: a.Header,
		RequestRedaction: model.RedactionNone,
	}, u.k.Regeln)
	if err != nil {
		u.t.Fatalf("BildeSchluessel: %v", err)
	}
	return s.MatchKey
}

// keyA ist der Key des Standard-Requests; keyB und keyC unterscheiden sich im Pfad.
func (u *umg) keyA() string { return u.keyFuer(anf(), svcID, "/Products") }
func (u *umg) keyB() string { return u.keyFuer(anf(), svcID, "/Orders") }

// anfB und anfC fragen die Keys B und C.
func anfB() inbound.Anfrage { a := anf(); a.Pfad = root + "/Orders"; return a }
func anfC() inbound.Anfrage { a := anf(); a.Pfad = root + "/Never"; return a }

// blobRef ist die Referenz des Blobs mit den Bytes d.
func blobRef(d []byte) string {
	s := sha256.Sum256(d)
	return "sha256:" + hex.EncodeToString(s[:])
}

func bodyVon(seq int) []byte { return []byte("body-" + strconv.Itoa(seq)) }

// austausch baut den Exchange mit SequenceNo seq und dem Key key; ok sagt, ob er replayfähig ist (sonst
// capture_status failed). Der Header X-Seq trägt seq, der Blob die Bytes body-<seq>; der Blob liegt im Fake.
func (u *umg) austausch(seq int, key string, ok bool) model.Exchange {
	d := bodyVon(seq)
	ref := blobRef(d)
	u.bl.mu.Lock()
	u.bl.inhalt[ref] = d
	u.bl.mu.Unlock()
	e := model.Exchange{
		ID: "exc-" + strconv.Itoa(seq), SessionID: sesID, SequenceNo: seq, Method: "GET", Path: "/Products",
		MatchKey: key, StartedAt: zeit(), ResponseStatus: 200,
		ResponseHeaders: map[string][]string{"X-Seq": {strconv.Itoa(seq)}, "Content-Type": {"text/plain"}},
		RequestBody:     model.BodyAbsent, ResponseBody: model.BodyPresent, ResponseBlob: ref,
		Capture: model.CaptureComplete, RequestRedact: model.RedactionNone, ResponseRedact: model.RedactionNone,
		Integrity: model.IntegrityVerified, Equivalence: model.EquivalenceByte,
	}
	if !ok {
		e.Capture = model.CaptureFailed
	}
	if err := e.Validate(); err != nil {
		u.t.Fatalf("Fixture ungültig: %v", err)
	}
	return e
}

// lege setzt die Exchanges der Session ses-a.
func (u *umg) lege(es ...model.Exchange) { u.ex.setze(sesID, es) }

// cursor legt einen Cursor des Workspace wsp-a für ses-a und rep-a in den Store.
func (u *umg) cursor(s model.ReplayStrategy, consumed ...int64) {
	z := u.st.stand()
	z.cur[wsID] = append(z.cur[wsID], model.ReplayCursor{WorkspaceID: wsID, ReplaySessionID: repID, RecordingSessionID: sesID,
		ServiceID: svcID, Strategy: s, Position: int64(len(consumed)), Consumed: consumed})
	u.st.setze(z)
}

// cursorStand liefert den Cursor (wsp-a, rep-a, ses-a) des Stores.
func (u *umg) cursorStand() (model.ReplayCursor, bool) {
	for _, c := range u.st.stand().cur[wsID] {
		if c.ReplaySessionID == repID && c.RecordingSessionID == sesID {
			return c, true
		}
	}
	return model.ReplayCursor{}, false
}

// seqVon liest X-Seq aus der Antwort.
func seqVon(t *testing.T, a inbound.Antwort) int {
	t.Helper()
	w := a.Header["X-Seq"]
	if len(w) != 1 {
		t.Fatalf("X-Seq fehlt in %v", a.Header)
	}
	n, err := strconv.Atoi(w[0])
	if err != nil {
		t.Fatalf("X-Seq: %v", err)
	}
	return n
}

// lies liest den Body der Antwort vollständig.
func lies(t *testing.T, a inbound.Antwort) []byte {
	t.Helper()
	if a.Body == nil {
		t.Fatal("Body ist nil")
	}
	b, err := io.ReadAll(a.Body)
	if err != nil {
		t.Fatalf("Body lesen: %v", err)
	}
	return b
}

// fehlerCode liefert Code und Status eines *inbound.Fehler; ohne ihn ist ok falsch.
func fehlerCode(err error) (model.Code, int, bool) {
	var f *inbound.Fehler
	if errors.As(err, &f) {
		return f.Code, f.Status, true
	}
	return 0, 0, false
}

// erwarteCode prüft, dass err ein *inbound.Fehler mit dem Code c und dem einen Status des Codes ist.
func erwarteCode(t *testing.T, was string, err error, c model.Code) {
	t.Helper()
	got, status, ok := fehlerCode(err)
	if !ok || got != c {
		t.Errorf("%s: Fehler %v, erwartet %s", was, err, c)
		return
	}
	if status != c.Stati()[0] {
		t.Errorf("%s: Status %d, erwartet %d", was, status, c.Stati()[0])
	}
}

// erwarteOhneCode prüft, dass err kein *inbound.Fehler ist und die Ursache cause enthält.
func erwarteOhneCode(t *testing.T, was string, err, cause error) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: kein Fehler", was)
		return
	}
	if _, _, ok := fehlerCode(err); ok {
		t.Errorf("%s: Fehler mit Code: %v", was, err)
	}
	if !errors.Is(err, cause) {
		t.Errorf("%s: Fehler %v enthält %v nicht", was, err, cause)
	}
}

// nurErlaubt prüft, dass das Protokoll allein Operationen der acht Stellen enthält, die der Dienst ruft
// (SPEC-135 Nr. 11, 18).
func nurErlaubt(t *testing.T, was string, p *prot) {
	t.Helper()
	erlaubt := []string{"Sessions.Get", "Exchanges.ListBySession", "Transaktion.WithWorkspaceTransaction",
		"tx Workspaces.Get", "tx Cursors.Get", "tx Cursors.Put", "Blobs.Open"}
	for _, op := range p.ops() {
		found := false
		for _, e := range erlaubt {
			found = found || e == op
		}
		if !found {
			t.Errorf("%s: verbotene Operation %s", was, op)
		}
	}
}

func blobFehlt() error       { return outbound.ErrBlobFehlt }
func blobBeschaedigt() error { return outbound.ErrBlobBeschaedigt }

// mitBody ersetzt den Blob des Exchange durch d; der Blob liegt im Fake.
func (u *umg) mitBody(e model.Exchange, d []byte) model.Exchange {
	ref := blobRef(d)
	u.bl.mu.Lock()
	u.bl.inhalt[ref] = d
	u.bl.mu.Unlock()
	e.ResponseBody, e.ResponseBlob = model.BodyPresent, ref
	return e
}

// ohneBody macht den Response-Body des Exchange absent.
func ohneBody(e model.Exchange) model.Exchange {
	e.ResponseBody, e.ResponseBlob = model.BodyAbsent, ""
	return e
}

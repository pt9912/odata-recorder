package admin_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// wsPruefeErfolg hält die Erwartung eines gelungenen Requests: der Status der Route, genau ein Aufruf, und zwar der ihrer
// Operation, mit den Argumenten der Route als Literalen.
func wsPruefeErfolg(t *testing.T, name string, e *wsEnv, r wsRoute, status int) {
	t.Helper()
	if status != r.status {
		t.Errorf("%s: Status %d, erwartet %d", name, status, r.status)
	}
	e.p.nurDiagonale(t, name, r.op)
	if got := e.p.liste(); len(got) != 1 || got[0].a != r.args[0] || got[0].b != r.args[1] {
		t.Errorf("%s: Aufrufe %+v, erwartet die Argumente %q", name, got, r.args)
	}
}

// TestWS1Konstruktor hält WS1 (SPEC-298 Nr. 1): der Gutfall liefert genau die acht Routen in der Folge des Vertrags, ein
// nil-Port ist model.ErrUngueltig ohne Routen und ohne Wert im Text (die anderen Ports tragen einen Canary), die Funktion ruft
// keinen Port (9 Zähler: acht Operationen und Audit) und `New` nimmt die Routen der Sessions, der Inspektion und der Workspaces
// zusammen an (4 + 3 + 8 = 15, keine Überlappung); die Zahl der Goroutinen fängt eine, die bleibt, eine kurzlebige fängt allein
// die Quelle (TestWS15Quelltext).
func TestWS1Konstruktor(t *testing.T) {
	audit := supporttest.NewFakeAudit()
	p := &wsProtokoll{}
	wf, sf := &workspaceFake{p: p}, &snapshotFake{p: p}
	vorher := runtime.NumGoroutine()
	var routen []admin.Route
	var err error
	for range 100 {
		routen, err = admin.WorkspaceRouten(wf, sf, audit)
	}
	if nachher := runtime.NumGoroutine(); err != nil || nachher > vorher+10 {
		t.Errorf("WorkspaceRouten: %v, Goroutinen vorher %d, nachher %d", err, vorher, nachher)
	}
	got := [][2]string{}
	for _, r := range routen {
		if r.Handler == nil {
			t.Errorf("Route %s %s ohne Handler", r.Methode, r.Muster)
		}
		got = append(got, [2]string{r.Methode, r.Muster})
	}
	gleichAuf(t, "Routen in der Folge des Vertrags", got, [][2]string{
		{"GET", "/__admin/workspaces"}, {"POST", "/__admin/workspaces"}, {"GET", "/__admin/workspaces/{id}"},
		{"DELETE", "/__admin/workspaces/{id}"}, {"POST", "/__admin/workspaces/{id}/reset"},
		{"GET", "/__admin/workspaces/{id}/snapshots"}, {"POST", "/__admin/workspaces/{id}/snapshots"},
		{"POST", "/__admin/workspaces/{id}/restore"},
	})
	if p.gesamt() != 0 || audit.Aufrufe() != 0 {
		t.Errorf("WorkspaceRouten rief %d-mal einen Dienst und %d-mal Audit, erwartet 0 und 0", p.gesamt(), audit.Aufrufe())
	}
	gleichAuf(t, "Zähler (Operationen und Audit)", int(wsOpAnzahl)+1, 9)

	faelle := []struct {
		name string
		mach func(c support.AuditLog) ([]admin.Route, error)
	}{
		{"workspaces nil", func(c support.AuditLog) ([]admin.Route, error) {
			return admin.WorkspaceRouten(nil, &snapshotFake{p: p, snap: model.WorkspaceSnapshot{ID: canary()}}, c)
		}},
		{"snapshots nil", func(c support.AuditLog) ([]admin.Route, error) {
			return admin.WorkspaceRouten(&workspaceFake{p: p, ws: model.Workspace{ID: canary()}}, nil, c)
		}},
		{"audit nil", func(support.AuditLog) ([]admin.Route, error) {
			return admin.WorkspaceRouten(&workspaceFake{p: p, ws: model.Workspace{ID: canary()}}, &snapshotFake{p: p}, nil)
		}},
	}
	for _, f := range faelle {
		mitCanary := supporttest.NewFakeAudit()
		_ = mitCanary.Append(support.Eintrag{ServiceID: canary()})
		rs, err := f.mach(mitCanary)
		if rs != nil || !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v Routen, Fehler %v, erwartet keine und model.ErrUngueltig", f.name, rs, err)
			continue
		}
		if text := fmt.Sprintf("%v %+v %#v", err, err, err); strings.Contains(text, canary()) {
			t.Errorf("%s: der Text des Fehlers nennt einen Wert: %q", f.name, text)
		}
		if mitCanary.Aufrufe() != 1 {
			t.Errorf("%s: Audit sah %d Aufrufe, erwartet den einen des Tests", f.name, mitCanary.Aufrufe())
		}
	}
	gleichAuf(t, "Zahl der Felder", len(faelle), 3)
	gleichAuf(t, "Zahl der Routen", len(routen), 8)

	sess, err := admin.SessionRouten(&sessionsFake{}, audit)
	insp, err2 := admin.InspektionRouten(&inspektionFake{})
	k, _ := konfigMit(slices.Concat(sess, insp, routen)...)
	if srv, errNeu := admin.New(k); err != nil || err2 != nil || errNeu != nil || srv == nil {
		t.Errorf("New mit %d Routen: %v, %v, %v", len(sess)+len(insp)+len(routen), err, err2, errNeu)
	}
	gleichAuf(t, "Routen zusammen", len(sess)+len(insp)+len(routen), 15)
}

// TestWS2Zuordnung hält WS2 (SPEC-298 Nr. 1, SPEC-238 Nr. 2): M-Route, fünf Pfadformen mal sieben Methoden, 35 Fälle: die
// acht Routen treffen, die 27 anderen sind 405 mit dem `Allow` der Pfadform (als Literale) und rufen keinen Dienst.
func TestWS2Zuordnung(t *testing.T) {
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	formen := []struct{ pfad, allow string }{
		{"/__admin/workspaces", "GET, POST"}, {"/__admin/workspaces/wsp-1", "DELETE, GET"}, {"/__admin/workspaces/wsp-1/reset", "POST"},
		{"/__admin/workspaces/wsp-1/snapshots", "GET, POST"}, {"/__admin/workspaces/wsp-1/restore", "POST"},
	}
	rs := wsRouten()
	faelle, treffer := 0, 0
	for _, f := range formen {
		for _, m := range methoden {
			faelle++
			e := wsNeu(t)
			var hit *wsRoute
			for i := range rs {
				if rs[i].methode == m && strings.Replace(rs[i].muster, "{id}", "wsp-1", 1) == f.pfad {
					hit = &rs[i]
				}
			}
			body := `{"baseline_snapshot_id":"b"}`
			if hit != nil {
				body = hit.body
			}
			rec := e.rufe(anfrageMit(m, f.pfad, nil, strings.NewReader(body)))
			if hit != nil {
				treffer++
				if rec.Code != hit.status {
					t.Errorf("%s %s: Status %d, erwartet %d", m, f.pfad, rec.Code, hit.status)
				}
				e.p.nurDiagonale(t, m+" "+f.pfad, hit.op)
				continue
			}
			pruefeFehlerAntwort(t, m+" "+f.pfad, rec, 405, objekt404, f.allow)
			if codes, _ := zaehlungenVon(e.tele); !slices.Equal(codes, []string{"ODR-E2102"}) || e.p.gesamt() != 0 || e.audit.Aufrufe() != 0 {
				t.Errorf("%s %s: CountError %q, %d Aufrufe des Dienstes, %d von Audit", m, f.pfad, codes, e.p.gesamt(), e.audit.Aufrufe())
			}
		}
	}
	muster := map[string]bool{}
	for _, r := range wsNeu(t).routen {
		muster[r.Muster] = true
	}
	gleichAuf(t, "Fälle, Treffer, Pfadformen, Routen", []int{faelle, treffer, len(muster), len(wsNeu(t).routen)}, []int{35, 8, 5, 8})
}

// TestWS3AufrufeUndKontext hält WS3 (SPEC-298 Nr. 2): jede Route ruft genau eine Operation (8 Routen × 8 Operationen = 64
// Zellen, 1 auf der Diagonale, sonst 0) mit den Argumenten als Literalen; der Wert im ctx des Requests kommt an, der Handler
// gibt ctx ohne Ableitung weiter (Gleichheit der Identität), und in diesem Teil ruft keine Route Audit.
func TestWS3AufrufeUndKontext(t *testing.T) {
	gleichAuf(t, "Operationen von WorkspaceUseCase", reflect.TypeFor[inbound.WorkspaceUseCase]().NumMethod(), 6)
	gleichAuf(t, "Operationen von SnapshotUseCase", reflect.TypeFor[inbound.SnapshotUseCase]().NumMethod(), 2)
	gleichAuf(t, "Zähler gleich 6 + 2", int(wsOpAnzahl), 8)
	zellen := 0
	for i, r := range wsRouten() {
		e := wsNeu(t)
		req := r.gut()
		req = req.WithContext(context.WithValue(req.Context(), ctxSchluessel{}, "wert-"+r.name))
		rec := e.rufe(req)
		zellen += len(e.p.zahlen())
		wsPruefeErfolg(t, r.name, e, r, rec.Code)
		if got := e.p.ctxListe()[0].Value(ctxSchluessel{}); got != "wert-"+r.name {
			t.Errorf("%s: der Wert im ctx des Requests kam als %v an", r.name, got)
		}
		if n := e.audit.Aufrufe(); n != 0 {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet 0 in diesem Teil", r.name, n)
		}

		ctx := context.WithValue(context.Background(), ctxSchluessel{}, "direkt")
		_, _ = e.routen[i].Handler(ctx, admin.Anfrage{
			Methode: r.methode, Body: strings.NewReader(r.body), KorrelationsID: "cor-1", Pfadwerte: map[string]string{"id": wsPfadID},
		})
		if got := e.p.ctxListe(); len(got) != 2 || got[1] != ctx {
			t.Errorf("%s: der Handler reichte einen anderen ctx weiter als den, den er erhielt", r.name)
		}
	}
	gleichAuf(t, "Zellen (Routen × Operationen)", zellen, 64)
}

// wsKoepfe sind die 26 Paare aus Name und Wert von M-Header (SPEC-298 Nr. 3): Content-Type mit 7 Werten und 19 weitere Namen
// mit je einem Wert.
func wsKoepfe() [][2]string {
	out := [][2]string{}
	for _, ct := range []string{
		"application/json", "application/json; charset=utf-8", "application/json; charset=latin1", "text/plain",
		"application/x-www-form-urlencoded", "multipart/form-data; boundary=x", "",
	} {
		out = append(out, [2]string{"Content-Type", ct})
	}
	for _, n := range []string{
		"Content-Encoding", "Content-Language", "Content-Disposition", "Accept", "Accept-Encoding", "Accept-Language", "Transfer-Encoding",
		"If-Match", "If-None-Match", "If-Modified-Since", "Cookie", "Origin", "Referer", "User-Agent", "X-Correlation-Id", "X-Forwarded-For",
		"X-Recorder-Recording-Session", "Prefer", "Expect",
	} {
		out = append(out, [2]string{n, "x"})
	}
	return out
}

// TestWS4QueryHeaderBody hält WS4 (SPEC-298 Nr. 3): keine Route liest Query oder Header oder prüft `Laenge`, und `GET`,
// `DELETE` und `reset` lesen den Body nie. M-Query 6 × 8 = 48, M-Body 3 × 5 = 15 (0 Lesezugriffe), M-Header (26 Paare + 7
// Längen) × 8 = 264; Summe 327. Die Antwort und die Argumente sind die des Requests ohne die Eigenschaft.
func TestWS4QueryHeaderBody(t *testing.T) {
	faelle := 0
	for _, r := range wsRouten() {
		for _, q := range []string{"", "limit=x", "a=1&a=2", "%zz", "%", "a=1;b=2"} {
			e := wsNeu(t)
			rec := e.rufe(r.request(wsPfadID, q, nil, strings.NewReader(r.body)))
			wsPruefeErfolg(t, r.name+" Query "+q, e, r, rec.Code)
			faelle++
		}
	}
	gleichAuf(t, "M-Query", faelle, 48)
	faelle = 0
	for _, r := range wsRouten() {
		if r.body != "" {
			continue
		}
		for _, b := range []string{"{", "null", "x"} {
			e, l := wsNeu(t), wsNeuerLeser([]byte(b), nil)
			rec := e.rufe(r.request(wsPfadID, "", nil, l))
			wsPruefeErfolg(t, r.name+" Body "+b, e, r, rec.Code)
			if n := l.n.Load(); n != 0 {
				t.Errorf("%s: %d Lesezugriffe auf den Body, erwartet 0", r.name, n)
			}
			faelle++
		}
	}
	gleichAuf(t, "M-Body", faelle, 15)
	faelle = 0
	laengen := []int64{-1, 0, 1, 20, 1 << 20, 1<<20 + 1, 1<<63 - 1}
	for _, r := range wsRouten() {
		for _, kv := range wsKoepfe() {
			e := wsNeu(t)
			rec := e.rufe(r.request(wsPfadID, "", map[string][]string{kv[0]: {kv[1]}}, strings.NewReader(r.body)))
			wsPruefeErfolg(t, r.name+" Header "+kv[0]+": "+kv[1], e, r, rec.Code)
			faelle++
		}
		for _, n := range laengen {
			e := wsNeu(t)
			req := r.gut()
			req.ContentLength = n
			rec := e.rufe(req)
			wsPruefeErfolg(t, fmt.Sprintf("%s Länge %d", r.name, n), e, r, rec.Code)
			faelle++
		}
	}
	gleichAuf(t, "M-Header", faelle, 264)
	gleichAuf(t, "Header, Längen", []int{len(wsKoepfe()), len(laengen)}, []int{26, 7})
}

// wsPfadWerte sind die 3993 Pfadwerte von M-Pfadwert als Paar aus der kodierten Form im Pfad und dem Wert, den der Dienst
// bekommt (Literale aus Bytewerten, nie aus einer Dekodierung): 13 Werte, 512 Bytes (jedes Byte allein und zwischen zwei
// Buchstaben) und 3468 Prozentfolgen (867 Folgen in 4 Stellungen).
func wsPfadWerte() [][2]string {
	b := func(c ...byte) string { return string(c) }
	w := [][2]string{
		{"x", "x"}, {"wsp-1", "wsp-1"}, {"WSP-1", "WSP-1"}, {"a.b", "a.b"}, {".", "."}, {"..", ".."}, {"a%2Fb", "a/b"},
		{"%C3%A4", b(0xC3, 0xA4)}, {"%FF", b(0xFF)}, {"%00", b(0)}, {strings.Repeat("a", 65), strings.Repeat("a", 65)},
		{"a+b", "a+b"}, {"%20", " "},
	}
	for c := range 256 {
		h := fmt.Sprintf("%%%02X", c)
		w = append(w, [2]string{h, b(byte(c))}, [2]string{"a" + h + "z", "a" + b(byte(c)) + "z"})
	}
	alphabet25, alphabet6 := "0123456789abcdefABCDEFg%+", "0Fag%+"
	folgen := []string{"%"}
	for _, c1 := range alphabet25 {
		folgen = append(folgen, "%"+string(c1))
		for _, c2 := range alphabet25 {
			folgen = append(folgen, "%"+string(c1)+string(c2))
		}
	}
	for _, c1 := range alphabet6 {
		for _, c2 := range alphabet6 {
			for _, c3 := range alphabet6 {
				folgen = append(folgen, "%"+string(c1)+string(c2)+string(c3))
			}
		}
	}
	for _, f := range folgen {
		for _, s := range [][2]string{{"", ""}, {"a", ""}, {"", "z"}, {"a", "z"}} {
			dekodiert := s[0] + f + s[1]
			w = append(w, [2]string{strings.ReplaceAll(dekodiert, "%", "%25"), dekodiert})
		}
	}
	return w
}

// TestWS5Pfadwert hält WS5 (SPEC-298 Nr. 4): `{id}` geht byte-gleich an den Dienst, ohne Prüfung und ohne Änderung, bei den
// sechs Routen mit Pfadwert je 3993 Werte (23958 Fälle); die Route setzt den Wert in keine Antwort (ein Canary als Pfadwert
// steht in keinem Byte der Antwort, auch nicht in der eines Fehlers des Dienstes).
func TestWS5Pfadwert(t *testing.T) {
	werte := wsPfadWerte()
	gleichAuf(t, "Werte je Route", len(werte), 13+512+3468)
	kodiert := map[string]bool{}
	for _, w := range werte {
		kodiert[w[0]] = true
	}
	// Fünf kodierte Formen kommen in zwei Teilmengen vor (`%FF`, `%00`, `%20` unter den 13 Werten und den Bytes, `%25` und
	// `a%25z` unter den Bytes und den Prozentfolgen); der Plan zählt die Summe der Teilmengen.
	gleichAuf(t, "verschiedene kodierte Formen", len(kodiert), 3993-5)
	gleichAuf(t, "Alphabet, Paare, Tripel", []int{len("0123456789abcdefABCDEFg%+"), 25 * 25, 6 * 6 * 6}, []int{25, 625, 216})
	routen, faelle := 0, 0
	for _, r := range wsRouten() {
		if !strings.Contains(r.muster, "{id}") {
			continue
		}
		routen++
		e := wsNeu(t)
		fehler := 0
		for _, w := range werte {
			e.p.zuruecksetzen()
			faelle++
			rec := e.rufe(r.request(w[0], "", nil, strings.NewReader(r.body)))
			got := e.p.liste()
			if rec.Code != r.status || len(got) != 1 || got[0].op != r.op || got[0].a != w[1] || got[0].b != r.args[1] {
				if fehler++; fehler <= 5 {
					t.Errorf("%s Pfad %q: Status %d, Aufrufe %+v, erwartet %d und den Wert %q", r.name, w[0], rec.Code, got, r.status, w[1])
				}
			}
		}
		if fehler > 5 {
			t.Errorf("%s: %d Fälle fehlgeschlagen", r.name, fehler)
		}
		for _, err := range []error{nil, wsE7001()} {
			e.wf.err, e.sf.err = err, err
			rec := e.rufe(r.request(canary(), "", nil, strings.NewReader(r.body)))
			if strings.Contains(antwortText(rec), canary()) {
				t.Errorf("%s: der Pfadwert steht in der Antwort (Fehler %v)", r.name, err)
			}
		}
	}
	gleichAuf(t, "Routen mit Pfadwert und Fälle", []int{routen, faelle}, []int{6, 23958})
}

package httpodata_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// A1: New weist jedes fehlende Feld ab, ohne Wert im Text und ohne Aufruf des Listeners.
type zaehlListener struct{ aufrufe int }

func (l *zaehlListener) Accept() (net.Conn, error) { l.aufrufe++; return nil, errors.New("zu") }
func (l *zaehlListener) Close() error              { l.aufrufe++; return nil }
func (l *zaehlListener) Addr() net.Addr            { l.aufrufe++; return &net.UnixAddr{Name: "CANARY-ADDR"} }

func TestNewUngueltig(t *testing.T) {
	d := &dienst{}
	ids := supporttest.NewFestIDs()
	for _, f := range []string{"Verarbeiten", "IDs", "Listener"} {
		lis := &zaehlListener{}
		k := httpodata.Konfig{Verarbeiten: d.Record, IDs: ids, Listener: lis}
		switch f {
		case "Verarbeiten":
			k.Verarbeiten = nil
		case "IDs":
			k.IDs = nil
		case "Listener":
			k.Listener = nil
		}
		srv, err := httpodata.New(k)
		if !errors.Is(err, model.ErrUngueltig) || srv != nil {
			t.Errorf("%s nil: srv=%v err=%v, erwartet nil und ErrUngueltig", f, srv, err)
		}
		if err != nil && strings.Contains(err.Error(), "CANARY-ADDR") {
			t.Errorf("%s nil: Fehlertext nennt einen Wert: %v", f, err)
		}
		if lis.aufrufe != 0 {
			t.Errorf("%s nil: New rief den Listener %d-mal", f, lis.aufrufe)
		}
	}
	lis := &zaehlListener{}
	srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: d.Record, IDs: ids, Listener: lis})
	if err != nil || srv == nil || lis.aufrufe != 0 {
		t.Errorf("alle gesetzt: srv=%v err=%v aufrufe=%d", srv, err, lis.aufrufe)
	}
}

// A2: Serve blockiert und endet mit einem Fehler, wenn der Listener schließt.
func TestServeEndet(t *testing.T) {
	d := &dienst{}
	neu := func(t *testing.T) (*httpodata.Server, net.Listener) {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: d.Record, IDs: supporttest.NewFestIDs(), Listener: lis})
		if err != nil {
			t.Fatal(err)
		}
		return srv, lis
	}
	ergebnis := func(srv *httpodata.Server) chan error {
		c := make(chan error, 1)
		go func() { c <- srv.Serve() }()
		return c
	}
	t.Run("von außen geschlossen", func(t *testing.T) {
		srv, lis := neu(t)
		c := ergebnis(srv)
		select {
		case err := <-c:
			t.Fatalf("Serve endete vor dem Schließen: %v", err)
		case <-time.After(100 * time.Millisecond):
		}
		_ = lis.Close()
		select {
		case err := <-c:
			if err == nil {
				t.Error("Serve lieferte nil")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Serve endete nicht in 5 s")
		}
	})
	t.Run("geschlossen übergeben", func(t *testing.T) {
		srv, lis := neu(t)
		_ = lis.Close()
		select {
		case err := <-ergebnis(srv):
			if err == nil {
				t.Error("Serve lieferte nil")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Serve endete nicht in 5 s")
		}
	})
}

// A3: Die Methode geht byte-gleich an den Dienst.
func TestAnfrageMethode(t *testing.T) {
	methoden := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "TRACE", "PROPFIND", "get"}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for i, m := range methoden {
			u.roh(t, zeile(m, "/p", "Content-Length: 0"))
			if got := u.d.alle()[i].a.Methode; got != m {
				t.Errorf("%s: Methode %q kam als %q an", b, m, got)
			}
		}
		if n := len(u.d.alle()); n != len(methoden) {
			t.Errorf("%s: %d Aufrufe, erwartet %d", b, n, len(methoden))
		}
	}
}

// A4: Pfad ist das Target bis zum ersten ?, byte-gleich.
func TestAnfragePfad(t *testing.T) {
	formen := []string{"/", "/a", "/a/", "//a", "/a/./b", "/a/../b", "/a%2Fb", "/a%2fb", "/%41", "/a%20b", "/a;x=1", "/a#b", "/\xC3\xA4"}
	if len(formen) != 13 {
		t.Fatalf("Zahl der Pfad-Formen = %d, erwartet 13", len(formen))
	}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for i, p := range formen {
			u.roh(t, zeile("GET", p))
			a := u.d.alle()[i].a
			if a.Pfad != p || a.RawQuery != "" {
				t.Errorf("%s: Pfad %q kam als (%q, %q) an", b, p, a.Pfad, a.RawQuery)
			}
		}
	}
	u := neuUmgebung(t, "record", nil)
	zeilen, _ := teile(u.roh(t, zeile("GET", "/a%zz")))
	_, _, _, gesamt := u.d.zaehler()
	if len(zeilen) == 0 || !strings.HasPrefix(zeilen[0], "HTTP/1.1 400") || gesamt != 0 || u.ids.zahlen() != [5]int64{} {
		t.Errorf("/a%%zz: Statuszeile %v, Aufrufe %d, IDs %v; erwartet 400 von net/http ohne Aufruf", zeilen, gesamt, u.ids.zahlen())
	}
}

// A5: RawQuery ist der Rest hinter dem ersten ?.
func TestAnfrageQuery(t *testing.T) {
	formen := []struct{ target, query string }{
		{"/p", ""}, {"/p?", ""}, {"/p?a=1", "a=1"}, {"/p?a=1&a=2", "a=1&a=2"},
		{"/p?$filter=Name%20eq%20%27x%27", "$filter=Name%20eq%20%27x%27"}, {"/p?%zz", "%zz"}, {"/p?a=%", "a=%"},
		{"/p?a?b", "a?b"}, {"/p?a=b;c", "a=b;c"}, {"/p?\xC3\xA4", "\xC3\xA4"}, {"/p?&&", "&&"}, {"/p?=", "="},
	}
	if len(formen) != 12 {
		t.Fatalf("Zahl der Query-Formen = %d, erwartet 12", len(formen))
	}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for i, f := range formen {
			u.roh(t, zeile("GET", f.target))
			a := u.d.alle()[i].a
			if a.Pfad != "/p" || a.RawQuery != f.query {
				t.Errorf("%s: %q kam als (%q, %q) an, erwartet (/p, %q)", b, f.target, a.Pfad, a.RawQuery, f.query)
			}
		}
	}
}

// A6: Targets ohne führenden / gehen als Pfad an den Dienst; der Dienst antwortet E5005.
func TestAnfrageFremdesTarget(t *testing.T) {
	faelle := []struct{ methode, target, pfad, query string }{
		{"GET", "http://h/OData/x?q=1", "http://h/OData/x", "q=1"},
		{"OPTIONS", "*", "*", ""},
		{"CONNECT", "h:443", "h:443", ""},
	}
	for _, f := range faelle {
		u := neuUmgebung(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			return inbound.Antwort{}, &inbound.Fehler{Code: model.CodeE5005, Status: 404}
		})
		zeilen, body := teile(u.roh(t, zeile(f.methode, f.target)))
		al := u.d.alle()
		if len(al) != 1 {
			t.Errorf("%s %s: %d Aufrufe, Antwort %q", f.methode, f.target, len(al), zeilen)
			continue
		}
		a := al[0].a
		if a.Pfad != f.pfad || a.RawQuery != f.query {
			t.Errorf("%s %s: (%q, %q), erwartet (%q, %q)", f.methode, f.target, a.Pfad, a.RawQuery, f.pfad, f.query)
		}
		if !strings.HasPrefix(zeilen[0], "HTTP/1.1 404") || !strings.Contains(string(body), "ODR-E5005") {
			t.Errorf("%s %s: Antwort %v %q, erwartet 404 mit ODR-E5005", f.methode, f.target, zeilen, body)
		}
	}
}

func mitConnection(extra map[string][]string) map[string][]string {
	out := map[string][]string{"Connection": {"close"}}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

// A7: Header trägt alle Header, Werte und Reihenfolge, kanonisch, ohne Host und Transfer-Encoding.
func TestAnfrageHeader(t *testing.T) {
	faelle := []struct {
		name    string
		methode string
		kopf    []string
		body    string
		soll    map[string][]string
		laenge  int64
	}{
		{"ein Wert", "GET", []string{"X-A: 1"}, "", map[string][]string{"X-A": {"1"}}, 0},
		{"zwei Zeilen", "GET", []string{"X-A: 1", "X-A: 2"}, "", map[string][]string{"X-A": {"1", "2"}}, 0},
		{"Schreibweisen", "GET", []string{"x-a: 1", "X-A: 2"}, "", map[string][]string{"X-A": {"1", "2"}}, 0},
		{"leerer Wert", "GET", []string{"X-E:"}, "", map[string][]string{"X-E": {""}}, 0},
		{"X-Recorder", "GET", []string{"X-Recorder-Workspace: w", "X-Recorder-Recording-Session: s", "X-Recorder-Replay-Session: r"}, "",
			map[string][]string{"X-Recorder-Workspace": {"w"}, "X-Recorder-Recording-Session": {"s"}, "X-Recorder-Replay-Session": {"r"}}, 0},
		{"Expect", "GET", []string{"Expect: 100-continue"}, "", map[string][]string{"Expect": {"100-continue"}}, 0},
		{"Upgrade", "GET", []string{"Upgrade: websocket"}, "", map[string][]string{"Upgrade": {"websocket"}}, 0},
		{"Host", "GET", nil, "", map[string][]string{}, 0},
		{"chunked", "POST", []string{"Transfer-Encoding: chunked"}, "3\r\nabc\r\n0\r\n\r\n", map[string][]string{}, -1},
		{"Content-Length", "POST", []string{"Content-Length: 3"}, "abc", map[string][]string{"Content-Length": {"3"}}, 3},
	}
	if len(faelle) != 10 {
		t.Fatalf("Zahl der Fälle = %d, erwartet 10", len(faelle))
	}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for i, f := range faelle {
			u.roh(t, append(zeile(f.methode, "/p", f.kopf...), f.body...))
			a := u.d.alle()[i].a
			if f.name == "Host" {
				if _, da := a.Header["Host"]; da {
					t.Errorf("%s: Host steht in Header: %v", b, a.Header)
				}
				continue
			}
			gleich(t, fmt.Sprintf("%s %s Header", b, f.name), a.Header, mitConnection(f.soll))
			gleich(t, fmt.Sprintf("%s %s Laenge", b, f.name), a.Laenge, f.laenge)
		}
	}
}

// A8: Body nil und Laenge 0 ohne Body, sonst Strom und angekündigte Länge.
func TestAnfrageBodyForm(t *testing.T) {
	faelle := []struct {
		name    string
		methode string
		kopf    []string
		body    string
		istNil  bool
		laenge  int64
	}{
		{"GET ohne Body", "GET", nil, "", true, 0},
		{"Content-Length 0", "POST", []string{"Content-Length: 0"}, "", true, 0},
		{"Content-Length 1", "POST", []string{"Content-Length: 1"}, "a", false, 1},
		{"chunked in zwei Chunks", "POST", []string{"Transfer-Encoding: chunked"}, "2\r\nab\r\n3\r\ncde\r\n0\r\n\r\n", false, -1},
		{"chunked leer", "POST", []string{"Transfer-Encoding: chunked"}, "0\r\n\r\n", false, -1},
	}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for i, f := range faelle {
			u.roh(t, append(zeile(f.methode, "/p", f.kopf...), f.body...))
			a := u.d.alle()[i].a
			if (a.Body == nil) != f.istNil || a.Laenge != f.laenge {
				t.Errorf("%s %s: Body nil=%v Laenge=%d, erwartet nil=%v Laenge=%d", b, f.name, a.Body == nil, a.Laenge, f.istNil, f.laenge)
			}
		}
	}
}

// A9: Der Request-Body kommt ungelesen, unverändert und ohne Grenze an.
func TestAnfrageBodyBytes(t *testing.T) {
	groessen := []int{1, 4095, 4096, 4097, 65535, 65536, 65537, 8*1024*1024 + 1}
	for _, b := range bindungen() {
		var mu sync.Mutex
		var gelesen []byte
		var fehler error
		u := neuUmgebung(t, b, func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			got, err := io.ReadAll(a.Body)
			mu.Lock()
			gelesen, fehler = got, err
			mu.Unlock()
			return inbound.Antwort{Status: 200, Body: leer()}, nil
		})
		for i, n := range groessen {
			soll := muster(n, uint64(i+1))
			req, _ := http.NewRequest("POST", "http://"+u.addr+"/p", bytes.NewReader(soll))
			resp, err := u.client().Do(req)
			if err != nil {
				t.Fatalf("%s %d: %v", b, n, err)
			}
			_ = resp.Body.Close()
			mu.Lock()
			if fehler != nil || !bytes.Equal(gelesen, soll) {
				t.Errorf("%s %d Bytes: gelesen %d Bytes, Fehler %v; erwartet gleiche Bytes", b, n, len(gelesen), fehler)
			}
			mu.Unlock()
		}
	}
}

// A10: KorrelationsID kommt einmal je Request aus IDs.
func TestKorrelationsID(t *testing.T) {
	u := neuUmgebung(t, "record", nil)
	for i := 0; i < 3; i++ {
		u.roh(t, zeile("GET", "/p"))
	}
	for i, a := range u.d.alle() {
		soll := fmt.Sprintf("cor-%032d", i+1)
		if a.a.KorrelationsID != soll {
			t.Errorf("Request %d: KorrelationsID %q, erwartet %q", i+1, a.a.KorrelationsID, soll)
		}
	}
	if got := u.ids.zahlen(); got != [5]int64{0, 0, 3, 0, 0} {
		t.Errorf("IDs-Zähler = %v, erwartet [0 0 3 0 0]", got)
	}
}

// A11: Verarbeiten genau einmal je Request, nur der Dienst der Bindung, Drain nie.
func TestEinAufruf(t *testing.T) {
	// Erfolg, *Fehler, Fehler ohne Code und Kontextfehler: jede Antwortart ruft den Dienst einmal.
	tu := func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		switch a.Pfad {
		case "/fehler":
			return inbound.Antwort{}, &inbound.Fehler{Code: model.CodeE5001, Status: 404}
		case "/ohnecode":
			return inbound.Antwort{}, errors.New("x")
		case "/kontext":
			return inbound.Antwort{}, context.Canceled
		}
		return inbound.Antwort{Status: 200, Body: leer()}, nil
	}
	pfade := []string{"/ok", "/fehler", "/ohnecode", "/kontext"}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, tu)
		for _, p := range pfade {
			u.roh(t, zeile("GET", p))
		}
		eigen, fremd := u.d.eigen(b)
		_, _, drain, _ := u.d.zaehler()
		if eigen != len(pfade) || fremd != 0 || drain != 0 {
			t.Errorf("%s: eigen=%d fremd=%d drain=%d, erwartet %d 0 0", b, eigen, fremd, drain, len(pfade))
		}
	}
}

// A12: ctx ist r.Context() ohne Frist und endet, wenn der Client abbricht.
func TestKontextClientAbbruch(t *testing.T) {
	type ergebnis struct {
		err   error
		frist bool
	}
	for _, mitBody := range []bool{false, true} {
		erg := make(chan ergebnis, 1)
		gestartet := make(chan struct{})
		u := neuUmgebung(t, "record", func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			if a.Body != nil {
				if _, err := io.ReadAll(a.Body); err != nil {
					erg <- ergebnis{err: err}
					return inbound.Antwort{}, err
				}
			}
			close(gestartet)
			select {
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
			}
			_, frist := ctx.Deadline()
			erg <- ergebnis{err: ctx.Err(), frist: frist}
			return inbound.Antwort{}, ctx.Err()
		})
		c, err := net.Dial("tcp", u.addr)
		if err != nil {
			t.Fatal(err)
		}
		req := zeile("GET", "/p")
		if mitBody {
			req = append(zeile("POST", "/p", "Content-Length: 3"), "abc"...)
		}
		if _, err := c.Write(req); err != nil {
			t.Fatal(err)
		}
		<-gestartet
		_ = c.Close()
		select {
		case r := <-erg:
			if !errors.Is(r.err, context.Canceled) || r.frist {
				t.Errorf("mitBody=%v: ctx.Err()=%v Frist=%v, erwartet Canceled ohne Frist", mitBody, r.err, r.frist)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("mitBody=%v: ctx endete nicht in 5 s", mitBody)
		}
	}
}

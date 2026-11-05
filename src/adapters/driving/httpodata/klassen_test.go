package httpodata_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// A13 über die Klasse: jeder Status von 200 bis 599 geht unverändert aus (400 Elemente je Bindung).
func TestAntwortStatusAlle(t *testing.T) {
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			s, _ := strconv.Atoi(a.Pfad[1:])
			return inbound.Antwort{Status: s, Body: leer()}, nil
		})
		n := 0
		for s := 200; s <= 599; s++ {
			resp, _ := hole(t, u, "GET", "/"+strconv.Itoa(s))
			if resp.StatusCode != s {
				t.Errorf("%s: Status %d kam als %d an", b, s, resp.StatusCode)
			}
			n++
		}
		if n != 400 {
			t.Fatalf("Zahl der Status = %d, erwartet 400", n)
		}
	}
}

// A32 über die Klasse: bei HEAD tragen alle 35 Paare Status, Content-Type, Content-Length und
// X-Correlation-Id, ohne Body.
func TestFehlerHeadAlleCodes(t *testing.T) {
	codes, stati := codeStatusPaare()
	for _, b := range bindungen() {
		var n atomic.Int64
		u := neuUmgebung(t, b, func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			j := n.Add(1) - 1
			return inbound.Antwort{}, &inbound.Fehler{Code: codes[j], Status: stati[j]}
		})
		for i := range codes {
			resp, body := hole(t, u, "HEAD", "/p")
			id := u.d.alle()[i].a.KorrelationsID
			soll := fehlerobjekt(t, codes[i])
			if resp.StatusCode != stati[i] || len(body) != 0 ||
				resp.Header.Get("Content-Length") != fmt.Sprint(len(soll)) ||
				resp.Header.Get("Content-Type") != "application/json" ||
				resp.Header.Get("X-Correlation-Id") != id {
				t.Errorf("%s %v/%d: HEAD Status %d %d Body-Bytes %v, ID %s", b, codes[i], stati[i], resp.StatusCode, len(body), resp.Header, id)
			}
		}
	}
}

// A15 über die Klasse: gleiche Werte desselben Namens gehen alle aus (4 Namen × 4 Wertfolgen × 2 Bindungen).
func TestAntwortHeaderGleicheWerte(t *testing.T) {
	hundert := make([]string, 100)
	for i := range hundert {
		hundert[i] = []string{"a", "b"}[i%2]
	}
	folgen := [][]string{{"a", "a"}, {"a", "b", "a"}, {"a", "a", "a"}, hundert, {"", ""}, {"a", "", "b"}, {"", "a"}}
	namen := []string{"X-A", "Set-Cookie", "Vary", "Via"}
	for _, b := range bindungen() {
		for _, name := range namen {
			for i, werte := range folgen {
				u := neuUmgebung(t, b, liefert(inbound.Antwort{Status: 200, Header: map[string][]string{name: werte}, Body: leer()}))
				resp, _ := hole(t, u, "GET", "/p")
				gleich(t, fmt.Sprintf("%s %s Folge %d", b, name, i+1), resp.Header[name], werte)
			}
		}
	}
}

// A16 über die Klasse: die Connection-Liste in jeder Schreibweise des Namens, jeder Form des Wertes
// (bis zu fünf Zeilen und fünf Namen) und jeder Schreibweise der genannten Namen (3 × 4 × 2 = 24 je
// Bindung). Prüfgrenze: mehr als fünf Zeilen oder Namen sind nicht geführt.
func TestAntwortConnectionListe(t *testing.T) {
	namen := []string{"Connection", "connection", "CONNECTION"}
	formen := [][]string{
		{"X-A, X-B, X-C, X-D, X-E"}, {"X-A", "X-B", "X-C", "X-D", "X-E"},
		{" x-a ,,\tX-B", "X-C ,", ",X-D", "X-E"}, {"X-A,X-B,X-C,X-D,X-E"},
	}
	schreibweisen := [][]string{{"X-A", "X-B", "X-C", "X-D", "X-E"}, {"x-a", "x-b", "x-c", "x-d", "x-e"}}
	for _, b := range bindungen() {
		for _, cn := range namen {
			for _, f := range formen {
				for _, s := range schreibweisen {
					h := map[string][]string{cn: f, "X-Keep": {"3"}}
					for _, n := range s {
						h[n] = []string{"1"}
					}
					u := neuUmgebung(t, b, liefert(inbound.Antwort{Status: 200, Header: h, Body: ausBytes([]byte("ok"))}))
					kopf, _ := kopfZeilen(u.rohOffen(t, "GET", "/p"))
					keep := false
					for _, k := range kopf {
						keep = keep || k == "X-Keep"
						for _, n := range append([]string{"Connection"}, s...) {
							if strings.EqualFold(k, n) {
								t.Errorf("%s %s %q %v: %s steht im Kopf %v", b, cn, f, s, k, kopf)
							}
						}
					}
					if !keep {
						t.Errorf("%s %s %q: X-Keep fehlt in %v", b, cn, f, kopf)
					}
				}
			}
		}
	}
}

// rohBytes liefert jedes Byte 0x21 bis 0xFF außer 0x7F, ohne die Bytes in aus.
func rohBytes(aus ...byte) []byte {
	var out []byte
	for c := 0x21; c <= 0xFF; c++ {
		if c != 0x7F && !bytes.Contains(aus, []byte{byte(c)}) {
			out = append(out, byte(c))
		}
	}
	return out
}

// A4 über die Klasse: jedes Rohbyte und jedes Escape %00 bis %FF (groß und klein) im Pfad bleiben
// byte-gleich (Rohbyte innen, am Ende und am Anfang: 3 × 220, dazu 512 Escapes = 1172 Formen je Bindung).
func TestAnfragePfadKlassen(t *testing.T) {
	var formen []string
	for _, c := range rohBytes('%', '?') {
		formen = append(formen, "/a"+string([]byte{c})+"b", "/a"+string([]byte{c}), "/"+string([]byte{c})+"b")
	}
	for c := 0; c < 256; c++ {
		formen = append(formen, fmt.Sprintf("/p%%%02X", c), fmt.Sprintf("/p%%%02x", c))
	}
	if len(formen) != 220*3+512 {
		t.Fatalf("Zahl der Formen = %d, erwartet 1172", len(formen))
	}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for i, p := range formen {
			zeilen, _ := teile(u.roh(t, zeile("GET", p)))
			al := u.d.alle()
			if len(al) != i+1 {
				t.Errorf("%s: Pfad %q erreichte den Dienst nicht: %v", b, p, zeilen[0])
				return
			}
			if a := al[i].a; a.Pfad != p || a.RawQuery != "" {
				t.Errorf("%s: Pfad %q kam als (%q, %q) an", b, p, a.Pfad, a.RawQuery)
			}
		}
	}
}

// A5 über die Klasse: jedes Rohbyte und jedes Escape in der Query und die Formen mit ? am Ende
// (Rohbyte innen, am Ende und am Anfang: 3 × 222, dazu 512 Escapes und 4 Randformen = 1182 Formen je Bindung).
func TestAnfrageQueryKlassen(t *testing.T) {
	var formen []string
	for _, c := range rohBytes() {
		formen = append(formen, "a"+string([]byte{c})+"b", "a"+string([]byte{c}), string([]byte{c})+"b")
	}
	for c := 0; c < 256; c++ {
		formen = append(formen, fmt.Sprintf("q=%%%02X", c), fmt.Sprintf("q=%%%02x", c))
	}
	formen = append(formen, "a?", "?", "a=1&", "&")
	if len(formen) != 222*3+512+4 {
		t.Fatalf("Zahl der Formen = %d, erwartet 1182", len(formen))
	}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		for i, q := range formen {
			zeilen, _ := teile(u.roh(t, zeile("GET", "/p?"+q)))
			al := u.d.alle()
			if len(al) != i+1 {
				t.Errorf("%s: Query %q erreichte den Dienst nicht: %v", b, q, zeilen[0])
				return
			}
			if a := al[i].a; a.Pfad != "/p" || a.RawQuery != q {
				t.Errorf("%s: Query %q kam als (%q, %q) an", b, q, a.Pfad, a.RawQuery)
			}
		}
	}
}

// A7 über die Klasse: 55 weitergereichte Header-Namen, jeder mit eigenem Wert (Prüfgrenze: die Menge
// der Namen ist unbeschränkt).
func TestAnfrageHeaderNamen(t *testing.T) {
	namen := []string{
		"Authorization", "Cookie", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "User-Agent",
		"Cache-Control", "If-None-Match", "If-Match", "If-Modified-Since", "Range", "Origin", "Referer",
		"Content-Encoding", "Content-Language", "Keep-Alive", "Proxy-Authorization", "Proxy-Connection", "TE",
		"Pragma", "Prefer", "OData-Version", "OData-MaxVersion",
		"X-Forwarded-For", "X-Forwarded-Proto", "X-Forwarded-Host", "Forwarded", "Via", "If-Range", "If-Unmodified-Since",
		"Accept-Charset", "Date", "From", "Max-Forwards", "Warning", "DNT", "Sec-Fetch-Mode", "Upgrade-Insecure-Requests",
		"Content-MD5", "Content-Location", "Slug", "OData-Isolation", "OData-EntityId", "Upgrade", "Allow", "Last-Modified",
		"ETag", "Expires", "Age", "Retry-After", "Server", "Location", "WWW-Authenticate", "Proxy-Authenticate",
	}
	if len(namen) != 55 {
		t.Fatalf("Zahl der Namen = %d, erwartet 55", len(namen))
	}
	soll := map[string][]string{"Connection": {"close"}}
	var kopf []string
	for _, n := range namen {
		kopf = append(kopf, n+": v-"+n)
		soll[http.CanonicalHeaderKey(n)] = []string{"v-" + n}
	}
	for _, b := range bindungen() {
		u := neuUmgebung(t, b, nil)
		u.roh(t, zeile("GET", "/p", kopf...))
		gleich(t, string(b)+" Header", u.d.alle()[0].a.Header, soll)
	}
}

// A9: der Body kommt in fünf Formen (Länge 1, 3, 100000, chunked, chunked leer) als r.Body selbst an (nicht umhüllt) und ungelesen: net/http sendet 100 Continue
// erst beim ersten Lesen durch den Dienst.
func TestAnfrageBodyUngelesen(t *testing.T) {
	t.Run("nicht umhüllt", func(t *testing.T) {
		typ := make(chan string, 1)
		u := neuUmgebung(t, "record", func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			typ <- reflect.TypeOf(a.Body).String()
			return inbound.Antwort{Status: 200, Body: leer()}, nil
		})
		formen := [][]byte{
			append(zeile("POST", "/p", "Content-Length: 1"), "a"...),
			append(zeile("POST", "/p", "Content-Length: 3"), "abc"...),
			append(zeile("POST", "/p", "Content-Length: 100000"), muster(100000, 5)...),
			append(zeile("POST", "/p", "Transfer-Encoding: chunked"), "3\r\nabc\r\n0\r\n\r\n"...),
			append(zeile("POST", "/p", "Transfer-Encoding: chunked"), "0\r\n\r\n"...),
		}
		for i, f := range formen {
			u.roh(t, f)
			if got := <-typ; got != "*http.body" {
				t.Errorf("Form %d: Typ des Bodys = %s, erwartet *http.body", i+1, got)
			}
		}
	})
	t.Run("ungelesen", func(t *testing.T) {
		gestartet, weiter := make(chan struct{}), make(chan struct{})
		u := neuUmgebung(t, "record", func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			close(gestartet)
			<-weiter
			_, _ = io.ReadAll(a.Body)
			return inbound.Antwort{Status: 200, Body: leer()}, nil
		})
		c, err := net.Dial("tcp", u.addr)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = c.Close() }()
		_, _ = c.Write([]byte("POST /p HTTP/1.1\r\nHost: h\r\nConnection: close\r\nContent-Length: 3\r\nExpect: 100-continue\r\n\r\n"))
		select {
		case <-gestartet:
		case <-time.After(2 * time.Second):
			_ = c.SetReadDeadline(time.Now().Add(time.Second))
			n, _ := c.Read(make([]byte, 64))
			t.Fatalf("der Dienst wurde nicht gerufen; der Adapter hat den Body gelesen (%d Bytes vom Server)", n)
		}
		_ = c.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		if n, _ := c.Read(make([]byte, 64)); n != 0 {
			t.Fatalf("%d Bytes vor dem ersten Lesen des Dienstes: der Adapter hat den Body gelesen", n)
		}
		close(weiter)
		_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
		buf := make([]byte, 64)
		n, err := c.Read(buf)
		if err != nil || !strings.HasPrefix(string(buf[:n]), "HTTP/1.1 100") {
			t.Fatalf("nach dem Lesen: %q, %v; erwartet 100 Continue", buf[:n], err)
		}
		_, _ = c.Write([]byte("abc"))
		rest, _ := io.ReadAll(c)
		if !strings.Contains(string(rest), "200 OK") {
			t.Errorf("Antwort %q", rest)
		}
	})
}

// A23 über die Klasse der Pfade: auf jedem Pfad bleiben Header und Body der Anfrage bis zur
// Rückkehr von Close gültig (6 Pfade).
func TestLeiheJederPfad(t *testing.T) {
	pfade := []struct {
		name   string
		status int
		err    error
	}{
		{"Erfolg", 200, nil}, {"Statusverstoß", 99, nil},
		{"Fehler mit Code", 200, &inbound.Fehler{Code: model.CodeE5001, Status: 404}},
		{"Fehler ohne Code", 200, errors.New("x")}, {"ErrUngueltig", 200, model.ErrUngueltig},
		{"Kontextfehler", 200, context.Canceled},
	}
	soll := muster(1000, 11)
	for _, p := range pfade {
		type bericht struct {
			gelesen []byte
			fehler  error
			header  bool
		}
		gemeldet := make(chan bericht, 1)
		k := ausBytes([]byte("abc"))
		var mitschrift aufruf
		k.beiClose = func() {
			var b bericht
			b.gelesen, b.fehler = io.ReadAll(mitschrift.a.Body)
			b.header = fmt.Sprint(mitschrift.a.Header) == fmt.Sprint(mitschrift.abbild)
			gemeldet <- b
		}
		u := neuUmgebung(t, "record", func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			mitschrift = aufruf{a: a, abbild: kopiereMap(a.Header)}
			return inbound.Antwort{Status: p.status, Body: k}, p.err
		})
		kunde := &http.Client{Transport: &http.Transport{}, Timeout: 10 * time.Second}
		req, _ := http.NewRequest("POST", "http://"+u.addr+"/p", bytes.NewReader(soll))
		req.Header.Set("X-Recorder-Workspace", "w")
		if resp, err := kunde.Do(req); err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		select {
		case b := <-gemeldet:
			if b.fehler != nil || !bytes.Equal(b.gelesen, soll) || !b.header {
				t.Errorf("%s: in Close Body gleich=%v Fehler=%v Header unverändert=%v", p.name, bytes.Equal(b.gelesen, soll), b.fehler, b.header)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("%s: Close wurde nicht gerufen", p.name)
		}
		kunde.CloseIdleConnections()
	}
}

// A21b (SPEC-136 Nr. 6c): der Fehler von Close bleibt ohne Wirkung, an jeder Stelle, an der Close
// gerufen wird (7 Pfade): die Beobachtung des Clients ist mit und ohne Fehler von Close dieselbe.
func TestCloseFehlerOhneWirkung(t *testing.T) {
	boese := errors.New("Lesefehler")
	pfade := []struct {
		name    string
		methode string
		status  int
		k       func() *koerper
		err     error
	}{
		{"Erfolg", "GET", 200, func() *koerper { return ausBytes([]byte("abc")) }, nil},
		{"Erfolg HEAD", "HEAD", 200, func() *koerper { return ausBytes(nil) }, nil},
		{"Lesefehler", "GET", 200, func() *koerper { return skript([]schritt{{"ab", nil}, {"", boese}}) }, nil},
		{"Statusverstoß", "GET", 99, func() *koerper { return ausBytes([]byte("abc")) }, nil},
		{"204 mit Bytes", "GET", 204, func() *koerper { return ausBytes([]byte("abc")) }, nil},
		{"Fehler und Antwort", "GET", 200, func() *koerper { return ausBytes([]byte("abc")) }, &inbound.Fehler{Code: model.CodeE5001, Status: 404}},
		{"Kontextfehler", "GET", 200, func() *koerper { return ausBytes([]byte("abc")) }, context.Canceled},
	}
	beobachte := func(p int, closeErr error) string {
		f := pfade[p]
		k := f.k()
		k.closeErr = closeErr
		u := neuUmgebung(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			return inbound.Antwort{Status: f.status, Body: k}, f.err
		})
		req, _ := http.NewRequest(f.methode, "http://"+u.addr+"/p", nil)
		resp, err := u.client().Do(req)
		if err != nil {
			warte(t, f.name+": Close", func() bool { c, _, _, _ := k.stand(); return c == 1 })
			return "kein Kopf"
		}
		body, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		warte(t, f.name+": Close", func() bool { c, _, _, _ := k.stand(); return c == 1 })
		return fmt.Sprintf("%d %q %v", resp.StatusCode, body, rerr)
	}
	for i := range pfade {
		ohne, mit := beobachte(i, nil), beobachte(i, errors.New("Close schlägt fehl"))
		if ohne != mit {
			t.Errorf("%s: Client sieht ohne Fehler von Close %s, mit Fehler %s", pfade[i].name, ohne, mit)
		}
	}
}

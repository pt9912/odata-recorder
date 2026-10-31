package httpodata_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// liefert ist der Dienst, der immer dieselbe Antwort ausgibt.
func liefert(a inbound.Antwort) func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
	return func(context.Context, inbound.Anfrage) (inbound.Antwort, error) { return a, nil }
}

// hole sendet eine Anfrage mit dem Client von net/http und liest den Body ganz.
func hole(t *testing.T, u *umgebung, methode, pfad string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(methode, "http://"+u.addr+pfad, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := u.client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", methode, pfad, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Body lesen: %v", err)
	}
	return resp, body
}

func fehlerobjekt(t *testing.T, c model.Code) []byte {
	t.Helper()
	o, err := model.NeuesFehlerobjekt(c)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// A13: Der Status geht unverändert aus.
func TestAntwortStatus(t *testing.T) {
	for _, b := range bindungen() {
		for _, s := range []int{200, 201, 204, 301, 304, 400, 404, 500, 599} {
			u := neuUmgebung(t, b, liefert(inbound.Antwort{Status: s, Body: leer()}))
			resp, _ := hole(t, u, "GET", "/p")
			if resp.StatusCode != s {
				t.Errorf("%s: Status %d kam als %d an", b, s, resp.StatusCode)
			}
		}
	}
}

// A14: Ein Bruch des Dienstes ergibt ODR-E9001 mit 500, ohne Header der Antwort.
func TestAntwortVerstoss(t *testing.T) {
	type fall struct {
		status int
		body   bool
	}
	faelle := []fall{{200, false}, {0, true}, {100, true}, {101, true}, {199, true}, {600, true}, {1000, true}, {0, false}}
	for _, f := range faelle {
		var k *koerper
		a := inbound.Antwort{Status: f.status, Header: map[string][]string{"X-Marker": {"m"}}}
		if f.body {
			k = leer()
			a.Body = k
		}
		u := neuUmgebung(t, "record", liefert(a))
		resp, body := hole(t, u, "GET", "/p")
		if resp.StatusCode != 500 || !bytes.Equal(body, fehlerobjekt(t, model.CodeE9001)) || resp.Header.Get("X-Marker") != "" {
			t.Errorf("Status %d body=%v: Antwort %d %q %v, erwartet 500 mit ODR-E9001 ohne X-Marker", f.status, f.body, resp.StatusCode, body, resp.Header)
		}
		if k != nil {
			if c, _, _, _ := k.stand(); c != 1 {
				t.Errorf("Status %d: Close %d-mal, erwartet 1", f.status, c)
			}
		}
	}
}

// A15: Antwort-Header unverändert, alle Werte in Reihenfolge.
func TestAntwortHeader(t *testing.T) {
	viele := map[string][]string{}
	for i := 0; i < 64; i++ {
		viele[fmt.Sprintf("X-H%02d", i)] = []string{fmt.Sprintf("v%02d", i)}
	}
	faelle := []map[string][]string{
		{"X-A": {"1"}}, {"X-A": {"1", "2"}}, {"Set-Cookie": {"a=1", "b=2"}}, {"X-E": {""}}, {"x-lower": {"v"}}, viele,
	}
	for _, b := range bindungen() {
		for i, h := range faelle {
			u := neuUmgebung(t, b, liefert(inbound.Antwort{Status: 200, Header: h, Body: leer()}))
			resp, _ := hole(t, u, "GET", "/p")
			for name, werte := range h {
				gleich(t, fmt.Sprintf("%s Fall %d %s", b, i+1, name), resp.Header[http.CanonicalHeaderKey(name)], werte)
			}
		}
	}
}

// kopfZeilen liefert die Zeilen des Kopfes einer rohen Antwort als Name und Wert.
func kopfZeilen(roh []byte) (namen []string, alle string) {
	zeilen, _ := teile(roh)
	for _, z := range zeilen[1:] {
		n, _, _ := strings.Cut(z, ":")
		namen = append(namen, n)
	}
	return namen, strings.Join(zeilen, "\n")
}

// A16: Hop-by-Hop-Header entfallen samt der Namen, die Connection nennt; die Anfrage bleibt offen,
// denn net/http ersetzt Connection selbst, sobald der Client schließt.
func TestAntwortHopByHop(t *testing.T) {
	namen := []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "TE", "Trailer", "Transfer-Encoding", "Upgrade"}
	roh := func(t *testing.T, b bindung, h map[string][]string) ([]string, string) {
		u := neuUmgebung(t, b, liefert(inbound.Antwort{Status: 200, Header: h, Body: ausBytes([]byte("ok"))}))
		return kopfZeilen(u.rohOffen(t, "GET", "/p"))
	}
	for _, b := range bindungen() {
		for _, n := range namen {
			for _, schreibweise := range []string{n, strings.ToLower(n)} {
				marker := "MARK-" + schreibweise
				_, alle := roh(t, b, map[string][]string{schreibweise: {marker}, "X-Keep": {"k"}})
				if strings.Contains(alle, marker) || !strings.Contains(alle, "X-Keep: k") {
					t.Errorf("%s: %s: Kopf enthält den Marker oder verliert X-Keep:\n%s", b, schreibweise, alle)
				}
			}
		}
		listen := []struct {
			wert   string
			namen  []string
			bleibt string
		}{
			{"X-Foo, x-bar", []string{"X-Foo", "X-Bar"}, "X-Keep"},
			{" X-A ,,\tX-B", []string{"X-A", "X-B"}, "X-Keep"},
		}
		for _, l := range listen {
			h := map[string][]string{"Connection": {l.wert}, "X-Keep": {"k"}}
			for _, n := range l.namen {
				h[n] = []string{"v"}
			}
			kopf, _ := roh(t, b, h)
			for _, n := range l.namen {
				for _, k := range kopf {
					if strings.EqualFold(k, n) {
						t.Errorf("%s: Connection %q: %s steht im Kopf %v", b, l.wert, n, kopf)
					}
				}
			}
			da := false
			for _, k := range kopf {
				da = da || k == l.bleibt
			}
			if !da {
				t.Errorf("%s: Connection %q: %s fehlt im Kopf %v", b, l.wert, l.bleibt, kopf)
			}
		}
	}
}

// A17: Kein Sniffing, kein ergänzter Content-Type.
func TestAntwortContentType(t *testing.T) {
	faelle := []struct {
		name string
		h    map[string][]string
		soll []string
	}{
		{"fehlt", nil, nil},
		{"text/plain", map[string][]string{"Content-Type": {"text/plain"}}, []string{"text/plain"}},
		{"klein", map[string][]string{"content-type": {"application/json"}}, []string{"application/json"}},
	}
	for _, f := range faelle {
		u := neuUmgebung(t, "record", liefert(inbound.Antwort{Status: 200, Header: f.h, Body: ausBytes([]byte("<html>x</html>"))}))
		resp, _ := hole(t, u, "GET", "/p")
		gleich(t, f.name+" Content-Type", resp.Header["Content-Type"], f.soll)
	}
}

// A18: Die Bytes der Antwort gehen unverändert aus, über jede Puffergrenze.
func TestAntwortBytes(t *testing.T) {
	groessen := []int{0, 1, 4095, 4096, 4097, 32767, 32768, 32769, 1048577, 8*1024*1024 + 1}
	for _, b := range bindungen() {
		for i, n := range groessen {
			soll := muster(n, uint64(100+i))
			u := neuUmgebung(t, b, liefert(inbound.Antwort{Status: 200, Body: ausBytes(soll)}))
			_, body := hole(t, u, "GET", "/p")
			if !bytes.Equal(body, soll) {
				t.Errorf("%s %d Bytes: %d Bytes kamen an, nicht gleich", b, n, len(body))
			}
		}
	}
}

// schritt ist ein Aufruf von Read.
type schritt struct {
	daten string
	err   error
}

func skript(s []schritt) *koerper {
	i := 0
	return &koerper{lies: func(p []byte) (int, error) {
		x := s[i]
		i++
		return copy(p, x.daten), x.err
	}}
}

// A19: Die Formen von Read gehen vollständig aus.
func TestAntwortReadFormen(t *testing.T) {
	formen := map[string][]schritt{
		"n und nil":        {{"ab", nil}, {"cd", nil}, {"", io.EOF}},
		"n und EOF":        {{"ab", nil}, {"cd", io.EOF}},
		"0 und nil mitten": {{"ab", nil}, {"", nil}, {"", nil}, {"cd", nil}, {"", nil}, {"", io.EOF}},
	}
	for name, s := range formen {
		k := skript(s)
		u := neuUmgebung(t, "replay", liefert(inbound.Antwort{Status: 200, Body: k}))
		_, body := hole(t, u, "GET", "/p")
		if string(body) != "abcd" {
			t.Errorf("%s: Body %q, erwartet abcd", name, body)
		}
		if c, _, _, _ := k.stand(); c != 1 {
			t.Errorf("%s: Close %d-mal", name, c)
		}
	}
}

// A20: HEAD, 204 und 304 ohne Body; eine gesetzte Content-Length bleibt.
func TestAntwortOhneBody(t *testing.T) {
	faelle := []struct {
		methode string
		status  int
		laenge  string
	}{{"HEAD", 200, "7"}, {"GET", 204, ""}, {"GET", 304, ""}}
	for _, f := range faelle {
		h := map[string][]string{}
		if f.laenge != "" {
			h["Content-Length"] = []string{f.laenge}
		}
		u := neuUmgebung(t, "record", liefert(inbound.Antwort{Status: f.status, Header: h, Body: leer()}))
		resp, body := hole(t, u, f.methode, "/p")
		if resp.StatusCode != f.status || len(body) != 0 || resp.Header.Get("Content-Length") != f.laenge {
			t.Errorf("%s %d: Status %d, %d Body-Bytes, Content-Length %q; erwartet Länge %q", f.methode, f.status, resp.StatusCode, len(body), resp.Header.Get("Content-Length"), f.laenge)
		}
	}
}

// A21: Close genau einmal auf jedem Pfad, nie vor dem letzten Read, nie ein Read danach.
func TestCloseEinmal(t *testing.T) {
	boese := fmt.Errorf("Lesefehler")
	faelle := []struct {
		name    string
		methode string
		status  int
		k       *koerper
		err     error
	}{
		{"Erfolg", "GET", 200, ausBytes([]byte("abc")), nil},
		{"Erfolg HEAD", "HEAD", 200, ausBytes(nil), nil},
		{"Lesefehler", "GET", 200, skript([]schritt{{"ab", nil}, {"", boese}}), nil},
		{"Statusverstoß", "GET", 99, ausBytes([]byte("abc")), nil},
		{"204 mit Bytes", "GET", 204, ausBytes([]byte("abc")), nil},
		{"Fehler und Antwort", "GET", 200, ausBytes([]byte("abc")), &inbound.Fehler{Code: model.CodeE5001, Status: 404}},
	}
	for _, f := range faelle {
		u := neuUmgebung(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			return inbound.Antwort{Status: f.status, Body: f.k}, f.err
		})
		req, _ := http.NewRequest(f.methode, "http://"+u.addr+"/p", nil)
		if resp, err := u.client().Do(req); err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		warte(t, f.name+": Close", func() bool { c, _, _, _ := f.k.stand(); return c >= 1 })
		time.Sleep(50 * time.Millisecond)
		c, nachClose, nachEnde, _ := f.k.stand()
		if c != 1 || nachClose != 0 || nachEnde != 0 {
			t.Errorf("%s: Close=%d Read-nach-Close=%d Read-nach-Ende=%d, erwartet 1 0 0", f.name, c, nachClose, nachEnde)
		}
	}
}

// A22: Ein Schreibfehler beendet das Kopieren, Close einmal; mit Content-Length leert der Adapter
// nicht, dort ist der Schreibfehler der einzige Ausgang.
func TestSchreibfehler(t *testing.T) {
	for _, mitLaenge := range []bool{false, true} {
		k := &koerper{lies: func(p []byte) (int, error) {
			n := min(len(p), 1024)
			for i := 0; i < n; i++ {
				p[i] = 'x'
			}
			return n, nil
		}}
		h := map[string][]string{}
		if mitLaenge {
			h["Content-Length"] = []string{"1099511627776"}
		}
		u := neuUmgebung(t, "record", liefert(inbound.Antwort{Status: 200, Header: h, Body: k}))
		c, err := net.Dial("tcp", u.addr)
		if err != nil {
			t.Fatal(err)
		}
		_ = c.SetDeadline(time.Now().Add(5 * time.Second))
		if _, err := c.Write(zeile("GET", "/p")); err != nil {
			t.Fatal(err)
		}
		if n, err := io.ReadAtLeast(c, make([]byte, 512), 512); err != nil {
			t.Fatalf("mitLaenge=%v: erstes Stück: %d Bytes, %v", mitLaenge, n, err)
		}
		_ = c.Close()
		warte(t, "Close nach dem Schreibfehler", func() bool { n, _, _, _ := k.stand(); return n == 1 })
		_, _, _, a := k.stand()
		time.Sleep(200 * time.Millisecond)
		_, _, _, b := k.stand()
		if a != b {
			t.Errorf("mitLaenge=%v: Read-Zähler wächst weiter: %d, dann %d", mitLaenge, a, b)
		}
		if n, nachClose, _, _ := k.stand(); n != 1 || nachClose != 0 {
			t.Errorf("mitLaenge=%v: Close=%d Read-nach-Close=%d, erwartet 1 0", mitLaenge, n, nachClose)
		}
	}
}

// A23: Der Handler kehrt erst nach Close zurück; bis dahin sind Header und Body der Anfrage gültig.
func TestLeihe(t *testing.T) {
	type bericht struct {
		gelesen []byte
		fehler  error
		header  bool
	}
	soll := muster(100000, 9)
	freigabe := make(chan struct{})
	gemeldet := make(chan bericht, 1)
	var mitschrift aufruf
	k := ausBytes([]byte("abc"))
	k.beiClose = func() {
		<-freigabe
		var b bericht
		d := mitschrift
		b.gelesen, b.fehler = io.ReadAll(d.a.Body)
		b.header = fmt.Sprint(d.a.Header) == fmt.Sprint(d.abbild)
		gemeldet <- b
	}
	u := neuUmgebung(t, "record", func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		mitschrift = aufruf{a: a, abbild: kopiereMap(a.Header)}
		return inbound.Antwort{Status: 200, Body: k}, nil
	})
	req, _ := http.NewRequest("POST", "http://"+u.addr+"/p", bytes.NewReader(soll))
	req.Header.Set("X-Recorder-Workspace", "w")
	// Eine Verbindung mit Keep-Alive: bei Connection: close verwirft net/http den Rest nicht.
	kunde := &http.Client{Transport: &http.Transport{}, Timeout: 10 * time.Second}
	defer kunde.CloseIdleConnections()
	resp, err := kunde.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if _, err := io.ReadFull(resp.Body, make([]byte, 3)); err != nil {
		t.Fatalf("erstes Stück: %v", err)
	}
	ende := make(chan error, 1)
	go func() { _, err := io.ReadAll(resp.Body); ende <- err }()
	select {
	case <-ende:
		t.Fatal("das Ende der Antwort kam, bevor Close zurückkehrte")
	case <-time.After(300 * time.Millisecond):
	}
	close(freigabe)
	select {
	case err := <-ende:
		if err != nil {
			t.Errorf("Ende der Antwort: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("das Ende der Antwort kam nach der Freigabe nicht")
	}
	b := <-gemeldet
	if b.fehler != nil || !bytes.Equal(b.gelesen, soll) || !b.header {
		t.Errorf("in Close: Body gleich=%v Fehler=%v Header unverändert=%v", bytes.Equal(b.gelesen, soll), b.fehler, b.header)
	}
}

// A24: Ohne Content-Length erreicht ein Stück den Client, bevor das nächste gelesen wird.
func TestStreaming(t *testing.T) {
	weiter := make(chan struct{})
	n := 0
	k := &koerper{lies: func(p []byte) (int, error) {
		n++
		if n == 1 {
			return copy(p, "abc"), nil
		}
		<-weiter
		return 0, io.EOF
	}}
	u := neuUmgebung(t, "replay", liefert(inbound.Antwort{Status: 200, Body: k}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://"+u.addr+"/p", nil)
	resp, err := u.client().Do(req)
	if err != nil {
		close(weiter)
		t.Fatalf("Kopf der Antwort kam nicht vor dem zweiten Stück: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	got := make([]byte, 3)
	_, err = io.ReadFull(resp.Body, got)
	close(weiter)
	if err != nil || string(got) != "abc" {
		t.Errorf("erstes Stück: %q, %v; erwartet abc", got, err)
	}
	rest, _ := io.ReadAll(resp.Body)
	if len(rest) != 0 {
		t.Errorf("Rest %q", rest)
	}
}

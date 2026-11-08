package admin_test

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// leitungStatus sendet den Text roh über TCP an den Server und liefert den Status der Antwort.
func leitungStatus(t *testing.T, adresse, text string) int {
	t.Helper()
	conn, err := net.Dial("tcp", adresse)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(conn, text); err != nil {
		t.Fatal(err)
	}
	res, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	return res.StatusCode
}

// TestLeitungAnnahmenZuNetHTTP hält U10: Leerraum am Rand des Wertes entfernt net/http, zwei Authorization-Zeilen auf der
// Leitung kommen als zwei Werte an, und ein nichtkanonischer Schlüssel auf der Leitung wird kanonisiert. Der Server
// läuft auf Loopback, der Client schreibt die Zeilen selbst.
func TestLeitungAnnahmenZuNetHTTP(t *testing.T) {
	tok := string(tokenGueltig())
	faelle := []struct {
		name    string
		kopf    string
		status  int
		aufrufe int
	}{
		{"Leerzeichen und Tab dahinter", "Authorization: Bearer " + tok + " \t\r\n", 200, 1},
		{"zwei gültige Zeilen", "Authorization: Bearer " + tok + "\r\nAuthorization: Bearer " + tok + "\r\n", 401, 0},
		{"Schlüssel in Kleinschreibung", "authorization: Bearer " + tok + "\r\n", 200, 1},
	}
	for _, f := range faelle {
		route := neueRoute()
		u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		ts := httptest.NewServer(u.srv)
		text := "GET /a HTTP/1.1\r\nHost: loopback\r\n" + f.kopf + "Connection: close\r\n\r\n"
		status := leitungStatus(t, strings.TrimPrefix(ts.URL, "http://"), text)
		ts.Close()
		if status != f.status || route.aufrufe() != f.aufrufe {
			t.Errorf("%s: Status %d, Route %d Aufrufe, erwartet %d und %d", f.name, status, route.aufrufe(), f.status, f.aufrufe)
		}
	}
}

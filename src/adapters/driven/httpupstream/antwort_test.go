package httpupstream_test

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// antwortServer liest einen Request, schreibt antwort und schließt die Verbindung; jeder Request
// kommt auf an.
func antwortServer(t *testing.T, antwort string, an chan<- empfangen) *roherServer {
	t.Helper()
	return starteServer(t, func(c net.Conn, r *bufio.Reader) {
		e, err := leseAnfrage(r)
		if err != nil {
			return
		}
		if an != nil {
			an <- e
		}
		_, _ = io.WriteString(c, antwort)
	})
}

func gzipt(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := io.WriteString(w, s); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// pruefeHeader verlangt genau die Einträge aus soll, keinen weiteren; ohne Einträge gleichen sich
// eine leere Map und nil.
func pruefeHeader(t *testing.T, ist, soll map[string][]string) {
	t.Helper()
	if len(ist) != len(soll) || (len(soll) > 0 && !reflect.DeepEqual(ist, soll)) {
		t.Errorf("Header %q, erwartet genau %q", ist, soll)
	}
}

// TestAntwort: Status, Header und Body kommen, wie der Upstream sie sendet, nach SPEC-120 Nr. 5;
// Header trägt genau die Namen des Kopfs (A7).
func TestAntwort(t *testing.T) {
	t.Run("a Header und gzip", func(t *testing.T) {
		gz := gzipt(t, "hallo")
		s := antwortServer(t, fmt.Sprintf("HTTP/1.1 201 Created\r\nset-cookie: a=1\r\nSet-Cookie: b=2\r\nbad name: x\r\n"+
			"X-Rand:  r\t\r\nX-Fold: a\r\n b\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(gz), gz), nil)
		ant, err := sende(context.Background(), t, neuerClient(t, zeitgrenze), anfrage(s.ziel(), "GET", "/"))
		if err != nil {
			t.Fatal(err)
		}
		body, err := liesAlles(t, ant.Body)
		if ant.Status != 201 || err != nil || !bytes.Equal(body, gz) {
			t.Errorf("Status %d, Body %s, %v", ant.Status, text(body), err)
		}
		pruefeHeader(t, ant.Header, map[string][]string{
			"Set-Cookie": {"a=1", "b=2"}, "bad name": {"x"}, "X-Rand": {"r"}, "X-Fold": {"a b"}, "Content-Encoding": {"gzip"},
			"Content-Length": {fmt.Sprint(len(gz))},
		})
	})
	t.Run("b 302", func(t *testing.T) {
		an := make(chan empfangen, 16)
		s := antwortServer(t, "HTTP/1.1 302 Found\r\nLocation: /x\r\nContent-Length: 0\r\n\r\n", an)
		ant, err := sende(context.Background(), t, neuerClient(t, zeitgrenze), anfrage(s.ziel(), "GET", "/"))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = liesAlles(t, ant.Body)
		if ant.Status != 302 || len(an) != 1 {
			t.Errorf("Status %d, %d Requests", ant.Status, len(an))
		}
		pruefeHeader(t, ant.Header, map[string][]string{"Location": {"/x"}, "Content-Length": {"0"}})
	})
	t.Run("c Zwischenantworten", func(t *testing.T) {
		s := antwortServer(t, "HTTP/1.1 100 Continue\r\n\r\nHTTP/1.1 103 Early Hints\r\nLink: </a>\r\n\r\n"+
			"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok", nil)
		ant, err := sende(context.Background(), t, neuerClient(t, zeitgrenze), anfrage(s.ziel(), "GET", "/"))
		if err != nil {
			t.Fatal(err)
		}
		if body, err := liesAlles(t, ant.Body); ant.Status != 200 || err != nil || string(body) != "ok" {
			t.Errorf("Status %d, Body %q, %v", ant.Status, body, err)
		}
		pruefeHeader(t, ant.Header, map[string][]string{"Content-Length": {"2"}})
	})
	t.Run("f Kopf ohne Header-Zeile", func(t *testing.T) {
		s := antwortServer(t, "HTTP/1.1 200 OK\r\n\r\nab", nil)
		ant, err := sende(context.Background(), t, neuerClient(t, zeitgrenze), anfrage(s.ziel(), "GET", "/"))
		if err != nil {
			t.Fatal(err)
		}
		if body, err := liesAlles(t, ant.Body); ant.Status != 200 || err != nil || string(body) != "ab" {
			t.Errorf("Status %d, Body %q, %v", ant.Status, body, err)
		}
		pruefeHeader(t, ant.Header, map[string][]string{})
	})
	for _, f := range []struct{ methode, antwort string }{
		{"GET", "HTTP/1.1 204 No Content\r\nContent-Length: 3\r\n\r\n"},
		{"GET", "HTTP/1.1 304 Not Modified\r\nContent-Length: 3\r\n\r\n"},
		{"HEAD", "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\n"},
	} {
		t.Run("d "+f.methode+" "+f.antwort[9:12], func(t *testing.T) {
			s := antwortServer(t, f.antwort, nil)
			ant, err := sende(context.Background(), t, neuerClient(t, zeitgrenze), anfrage(s.ziel(), f.methode, "/"))
			if err != nil {
				t.Fatal(err)
			}
			if ant.Body == nil {
				t.Fatal("Body nil")
			}
			if n, err := ant.Body.Read(make([]byte, 8)); n != 0 || !errors.Is(err, io.EOF) {
				t.Errorf("erstes Read %d, %v", n, err)
			}
			_ = ant.Body.Close()
		})
	}
	for _, f := range []struct {
		name, antwort, daten string
		eof                  bool
	}{
		{"Länge abgeschnitten", "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nab", "ab", false},
		{"chunked abgeschnitten", "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n2\r\nab\r\n", "ab", false},
		{"ohne Länge", "HTTP/1.1 200 OK\r\n\r\nab", "ab", true},
	} {
		t.Run("e "+f.name, func(t *testing.T) {
			s := antwortServer(t, f.antwort, nil)
			ant, err := sende(context.Background(), t, neuerClient(t, zeitgrenze), anfrage(s.ziel(), "GET", "/"))
			if err != nil {
				t.Fatal(err)
			}
			body, err := liesAlles(t, ant.Body)
			if string(body) != f.daten || (err == nil) != f.eof {
				t.Errorf("Body %q, %v", body, err)
			}
		})
	}
}

// nichtErreichbarFaelle: jeder Fehler vor dem Kopf ist ErrUpstreamNichtErreichbar, ebenso 101; ein
// Fall je Element der Aufzählung in SPEC-120 Nr. 8 (A9). Frist 1 h.
func nichtErreichbarFaelle() []fehlerFall {
	f := []fehlerFall{
		{name: "geschlossener Port", art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			adr := ln.Addr().String()
			_ = ln.Close()
			_, err = sende(context.Background(), t, neuerClient(t, time.Hour), mitGeheim(anfrage("http://"+adr, "GET", "")))
			return err
		}},
		{name: "Host nicht auflösbar", art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
			_, err := sende(context.Background(), t, neuerClient(t, time.Hour), mitGeheim(anfrage("http://host.invalid", "GET", "")))
			return err
		}},
		{name: "geschlossen ohne Antwort", art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
			s := antwortServer(t, "", nil)
			_, err := sende(context.Background(), t, neuerClient(t, time.Hour), mitGeheim(anfrage(s.ziel(), "GET", "")))
			return err
		}},
	}
	for _, na := range [][2]string{
		{"kein HTTP", geheim + "-r-kein-http\r\n\r\n"},
		{"Name kein Token", "HTTP/1.1 200 OK\r\nX(" + geheim + "-k: v\r\nContent-Length: 0\r\n\r\n"},
		{"Steuerzeichen im Wert", "HTTP/1.1 200 OK\r\nX-A: " + geheim + "-w\x01\r\nContent-Length: 0\r\n\r\n"},
		{"Status nicht dreistellig", "HTTP/1.1 2" + geheim + " OK\r\nContent-Length: 0\r\n\r\n"},
		{"Kopfzeile ohne Doppelpunkt", "HTTP/1.1 200 OK\r\n" + geheim + "-z\r\n\r\n"},
	} {
		name, antwort := na[0], na[1]
		f = append(f, fehlerFall{name: name, art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
			s := antwortServer(t, antwort, nil)
			_, err := sende(context.Background(), t, neuerClient(t, time.Hour), mitGeheim(anfrage(s.ziel(), "GET", "")))
			return err
		}})
	}
	return append(f, fehlerFall{name: "101", art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
		zu := make(chan struct{})
		s := starteServer(t, func(c net.Conn, r *bufio.Reader) {
			if _, err := leseAnfrage(r); err != nil {
				return
			}
			// Status-Text, Header-Wert und Wert von Upgrade tragen den Marker, dazu die Bytes nach dem Kopf.
			_, _ = io.WriteString(c, "HTTP/1.1 101 "+geheim+"-s\r\nUpgrade: "+geheim+"-u\r\nConnection: Upgrade\r\nX-K: "+
				geheim+"-h\r\n\r\n"+geheim+"-t")
			_, _ = io.Copy(io.Discard, r)
			close(zu)
		})
		_, err := sende(context.Background(), t, neuerClient(t, time.Hour), mitGeheim(anfrage(s.ziel(), "GET", "")))
		warte(t, zu)
		return err
	}})
}

// TestNichtErreichbar: Fehler vor dem Kopf und 101 sind ErrUpstreamNichtErreichbar, nicht
// ErrUpstreamZeit (A9, SPEC-120 Nr. 5, 8).
func TestNichtErreichbar(t *testing.T) {
	fahre(t, nichtErreichbarFaelle())
}

package batchcodec_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func anfragezeilenFall(t *testing.T, zeile string, ok bool, methode, ziel string) {
	t.Helper()
	got, err := lese(huelle(block([]string{zeileCTH}, zeile)))
	if !ok {
		wantFehler(t, strconv.Quote(zeile), got, err, model.BatchAnfragezeile, "1")
		return
	}
	wantOK(t, strconv.Quote(zeile), got, err, ergebnis(einzel(req(methode, ziel))))
}

// TestBatchAnfragezeile hält Zeile 22: `Methode SP Ziel SP HTTP/1.1` mit null bis drei SP an jeder der vier Stellen, HTAB, andere
// Versionen, leere Zeile davor und alle 256 Bytes in Methode und Ziel; Methode und Ziel wie geschrieben.
func TestBatchAnfragezeile(t *testing.T) {
	n := 0
	for a := range 4 {
		for b := range 4 {
			for c := range 4 {
				for d := range 4 {
					zeile := strings.Repeat(" ", a) + "GET" + strings.Repeat(" ", b) + "x" + strings.Repeat(" ", c) + "HTTP/1.1" + strings.Repeat(" ", d)
					anfragezeilenFall(t, zeile, a == 0 && b == 1 && c == 1 && d == 0, "GET", "x")
					n++
				}
			}
		}
	}
	for _, z := range []string{
		"", "GET", "GET x", "GET HTTP/1.1", "GET x HTTP/1.0", "GET x HTTP/2", "GET x http/1.1", "GET x HTTP/1.1x", "GET x HTTP/1.",
		"GET" + ht + "x HTTP/1.1", "GET x" + ht + "HTTP/1.1", "GET x HTTP/1.1" + ht, crlf + "GET x HTTP/1.1",
		"GET x HTTP/1.1" + lf + "Host: a", "GET x HTTP/1.1" + cr + "Host: a", "GET x y HTTP/1.1", "HTTP/1.1 GET x", geheim,
	} {
		anfragezeilenFall(t, z, false, "", "")
	}
	for b := range 256 {
		c := string([]byte{byte(b)})
		for _, m := range []string{c + "ET", "G" + c + "T", "GE" + c} {
			anfragezeilenFall(t, m+" x HTTP/1.1", strings.IndexByte(tchar, byte(b)) >= 0, m, "x")
			n++
		}
		for _, z := range []string{c + "ac", "a" + c + "c", "ac" + c} {
			anfragezeilenFall(t, "GET "+z+" HTTP/1.1", b >= 0x21 && b != 0x7F, "GET", z)
			n++
		}
	}
	if n != 256+256*6 {
		t.Errorf("%d Fälle", n)
	}
	for _, z := range []struct{ m, ziel string }{
		{"get", "Products"}, {"MERGE", "/Products(1)"}, {"X-Y", "$1"}, {"OPTIONS", "http://anderer.host:8/svc/Products?$top=1"},
		{"HEAD", "https://[::1]/a%20b"}, {"GET", "Products?$filter=Name%20eq%20'x'"},
	} {
		anfragezeilenFall(t, z.m+" "+z.ziel+" HTTP/1.1", true, z.m, z.ziel)
	}
	for _, kopf := range []string{zeileCTH, zeileCTH + crlf + crlf, zeileCTH + crlf + crlf + crlf} {
		got, err := lese(huelle(kopf))
		wantFehler(t, "Teil ohne Request "+strconv.Quote(kopf), got, err, model.BatchAnfragezeile, "1")
	}
}

// requestMit liest einen Request mit den Kopfzeilen und dem Body.
func requestMit(kopf []string, body string) (model.Batch, error) {
	return lese(huelle(block([]string{zeileCTH}, anfrage("POST x HTTP/1.1", kopf, body))))
}

// TestBatchRequest hält Zeile 23: Header in Reihenfolge mit dem Namen wie geschrieben und doppelt, Body jedes Byte hinter der
// leeren Zeile, sonst nil; Content-Length ohne Wirkung.
func TestBatchRequest(t *testing.T) {
	kopf := []string{"b: 2", "A: 1", "host:  h ", "Host: i", "a: " + esB(0xE4, 0x80), "Content-Type:" + ht + "t" + ht}
	var want []model.BatchKopf
	for _, p := range [][2]string{{"b", "2"}, {"A", "1"}, {"host", "h"}, {"Host", "i"}, {"a", esB(0xE4, 0x80)}, {"Content-Type", "t"}} {
		want = append(want, model.BatchKopf{Name: p[0], Wert: p[1]})
	}
	got, err := requestMit(kopf, "")
	r := req("POST", "x")
	r.Header = want
	wantOK(t, "Header", got, err, ergebnis(einzel(r)))
	bodies := []string{
		"abc", " ", crlf, crlf + crlf, crlf + crlf + "a" + crlf + crlf, "a" + crlf + crlf + "b", "--bx", " --b", "--b- -" + crlf + "x",
		"x" + lf + "--b--" + lf, "x" + cr + "--b", nul + esB(0xFF, 0x00, 0x80), "Transfer-Encoding: chunked" + crlf, strings.Repeat("z", 100000),
	}
	for _, body := range bodies {
		for _, laenge := range []string{"", "0", "1", strconv.Itoa(len(body)), "999999", "-1", "x"} {
			var kl []string
			if laenge != "" {
				kl = []string{"Content-Length: " + laenge}
			}
			got, err := requestMit(kl, body)
			r := req("POST", "x")
			if kl != nil {
				r.Header = []model.BatchKopf{{Name: "Content-Length", Wert: laenge}}
			}
			r.Body = []byte(body)
			wantOK(t, "Body mit Content-Length "+laenge, got, err, ergebnis(einzel(r)))
		}
	}
	for _, rest := range []string{"POST x HTTP/1.1", "POST x HTTP/1.1" + crlf, "POST x HTTP/1.1" + crlf + crlf} {
		got, err := lese(huelle(block([]string{zeileCTH}, rest)))
		wantOK(t, "ohne Body "+strconv.Quote(rest), got, err, ergebnis(einzel(req("POST", "x"))))
		if err == nil && got.Teile[0].Requests[0].Body != nil {
			t.Errorf("%q: Body ist nicht nil", rest)
		}
	}
}

// TestBatchRequestTransferEncoding hält Zeile 24: Transfer-Encoding in jeder Schreibweise, mit jedem Wert und an jeder Stelle ist
// `transfer_kodierung`, nach dem ganzen Block.
func TestBatchRequestTransferEncoding(t *testing.T) {
	n := jedeSchreibweise("transfer-encoding", func(name string) {
		got, err := requestMit([]string{name + ": chunked"}, "")
		if f, ok := err.(*model.BatchFehler); !ok || f.Ursache != model.BatchTransferKodierung || got.Teile != nil {
			t.Errorf("%q: %v", name, err)
		}
	})
	if n != 1<<16 {
		t.Errorf("%d Schreibweisen", n)
	}
	for _, wert := range []string{"", "chunked", "identity", "binary", "gzip, chunked", geheim} {
		for stelle := range 3 {
			kopf := []string{"A: 1", "B: 2"}
			kopf = append(kopf[:stelle], append([]string{"Transfer-Encoding: " + wert}, kopf[stelle:]...)...)
			got, err := requestMit(kopf, "body")
			wantFehler(t, "Wert "+wert+" an Stelle "+strconv.Itoa(stelle), got, err, model.BatchTransferKodierung, "1")
		}
	}
	// Der ganze Block geht vor: eine spätere falsche Zeile und die 101. Zeile gewinnen.
	got, err := requestMit([]string{"Transfer-Encoding: chunked", "ohne Doppelpunkt"}, "")
	wantFehler(t, "spätere kopfzeile", got, err, model.BatchKopfzeile, "1")
	got, err = requestMit(append([]string{"Transfer-Encoding: chunked"}, zeilen(100)...), "")
	wantFehler(t, "101. Zeile", got, err, model.BatchZuViele, "1")
	for _, name := range []string{"Content-Transfer-Encoding", "X-Transfer-Encoding", "Transfer-Encoding-X", "TE"} {
		got, err := requestMit([]string{name + ": chunked"}, "")
		if err != nil || len(got.Teile[0].Requests[0].Header) != 1 {
			t.Errorf("%s: %v", name, err)
		}
	}
	got, err = lese(huelle(block([]string{zeileCTH, "Transfer-Encoding: chunked"}, anfrage("GET x HTTP/1.1", nil, ""))))
	wantOK(t, "im Block des Teils", got, err, ergebnis(einzel(req("GET", "x"))))
}

package httpupstream_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestNewArgumente: eine Frist unter 1 ns ist model.ErrUngueltig ohne Client; 1 ns ist gültig (A1,
// SPEC-120 Nr. 1).
func TestNewArgumente(t *testing.T) {
	for _, frist := range []time.Duration{0, -time.Nanosecond} {
		if c, err := httpupstream.New(httpupstream.Konfig{Frist: frist, Pruefen: true}); !errors.Is(err, model.ErrUngueltig) || c != nil {
			t.Errorf("Frist %v: %v, Client %v", frist, err, c != nil)
		}
	}
	if c, err := httpupstream.New(httpupstream.Konfig{Frist: time.Nanosecond, Pruefen: true}); err != nil || c == nil {
		t.Errorf("Frist 1 ns: %v", err)
	}
}

// TestTransport: der Transport nutzt keinen Proxy aus der Umgebung; der Fänger ist das Feld, weil
// die Bibliothek die Umgebung je Prozess einmal liest (A11, SPEC-120 Nr. 7).
func TestTransport(t *testing.T) {
	if httpupstream.Transport(neuerClient(t, time.Hour)).Proxy != nil {
		t.Error("Proxy gesetzt")
	}
}

// TestGleichzeitig: 20 Aufrufe zugleich mit je eigenem Pfad, Header und Body; jede Antwort trägt
// ihre Werte. Ohne Race-Detektor fängt der Test ein geteiltes Feld nur mit der Rate aus §3 des Plans
// (A15, SPEC-120 Nr. 9).
func TestGleichzeitig(t *testing.T) {
	s := starteServer(t, func(c net.Conn, r *bufio.Reader) {
		for {
			e, err := leseAnfrage(r)
			if err != nil {
				return
			}
			pfad := strings.Fields(e.zeile)[1]
			body := pfad + "|" + strings.Join(e.werte("X-Nr"), ",") + "|" + string(e.body)
			if _, err := fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n%s", len(body), body); err != nil {
				return
			}
		}
	})
	c := neuerClient(t, zeitgrenze)
	const anzahl = 20
	ergebnisse := make(chan error, anzahl)
	start := make(chan struct{})
	for i := range anzahl {
		go func() {
			<-start
			a := anfrage(s.ziel(), "POST", fmt.Sprintf("/p%d", i))
			a.Header = map[string][]string{"X-Nr": {fmt.Sprint(i)}}
			body := fmt.Sprintf("body-%d", i)
			a.Body, a.Laenge = strings.NewReader(body), int64(len(body))
			ant, err := c.Senden(context.Background(), a)
			if err != nil {
				ergebnisse <- err
				return
			}
			b, err := io.ReadAll(ant.Body)
			_ = ant.Body.Close()
			if soll := fmt.Sprintf("/p%d|%d|body-%d", i, i, i); err != nil || string(b) != soll {
				err = fmt.Errorf("Antwort %q, erwartet %q, %v", b, soll, err)
			}
			ergebnisse <- err
		}()
	}
	close(start)
	for range anzahl {
		if err := warte(t, ergebnisse); err != nil {
			t.Error(err)
		}
	}
}

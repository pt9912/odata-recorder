package httpodata_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

func fehlerDienst(err error) func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
	return func(context.Context, inbound.Anfrage) (inbound.Antwort, error) { return inbound.Antwort{}, err }
}

// A25: Ein Lesefehler bricht die Verbindung ab, ohne Fehlerobjekt, zu jedem Zeitpunkt.
func TestLesefehler(t *testing.T) {
	boese := io.ErrUnexpectedEOF
	zeitpunkte := []struct {
		name    string
		schritt []schritt
		bytes   string
		antwort bool
	}{
		{"erstes Read", []schritt{{"", boese}}, "", false},
		{"nach dem ersten Stück", []schritt{{"abc", nil}, {"", boese}}, "abc", true},
		{"n und Fehler zugleich", []schritt{{"abc", boese}}, "abc", true},
	}
	for _, z := range zeitpunkte {
		for _, mitLaenge := range []bool{false, true} {
			k := skript(z.schritt)
			h := map[string][]string{}
			if mitLaenge {
				h["Content-Length"] = []string{"6"}
			}
			u := neuUmgebung(t, "record", liefert(inbound.Antwort{Status: 200, Header: h, Body: k}))
			req, _ := http.NewRequest("GET", "http://"+u.addr+"/p", nil)
			resp, err := u.client().Do(req)
			name := fmt.Sprintf("%s mitLaenge=%v", z.name, mitLaenge)
			if !z.antwort {
				if err == nil {
					_ = resp.Body.Close()
					t.Errorf("%s: der Client erhielt eine Antwort %d", name, resp.StatusCode)
				}
			} else if err != nil {
				t.Errorf("%s: Do: %v", name, err)
			} else {
				got, rerr := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if rerr == nil || string(got) != z.bytes || resp.StatusCode != 200 {
					t.Errorf("%s: Status %d, Body %q, Lesefehler %v; erwartet 200, %q und einen Fehler", name, resp.StatusCode, got, rerr, z.bytes)
				}
			}
			warte(t, name+": Close", func() bool { c, _, _, _ := k.stand(); return c == 1 })
		}
	}
}

// codeStatusPaare liefert alle Paare mit Status ungleich 0 aus der Codetabelle.
func codeStatusPaare() (codes []model.Code, stati []int) {
	for _, c := range model.AlleCodes() {
		for _, s := range c.Stati() {
			if s != 0 {
				codes, stati = append(codes, c), append(stati, s)
			}
		}
	}
	return codes, stati
}

// A26: Jeder *Fehler mit gültigem Code und Status geht mit diesem Status und dem Fehlerobjekt aus.
func TestFehlerobjektJeCode(t *testing.T) {
	codes, stati := codeStatusPaare()
	if len(codes) != 33 {
		t.Fatalf("Zahl der Paare = %d, erwartet 33", len(codes))
	}
	for _, b := range bindungen() {
		var zaehler atomic.Int64
		u := neuUmgebung(t, b, func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			j := zaehler.Add(1) - 1
			return inbound.Antwort{}, &inbound.Fehler{Code: codes[j], Status: stati[j]}
		})
		for i := range codes {
			resp, body := hole(t, u, "GET", "/p")
			id := u.d.alle()[i].a.KorrelationsID
			var obj map[string]map[string]string
			if err := json.Unmarshal(body, &obj); err != nil || len(obj["error"]) != 3 {
				t.Errorf("%s %v: Body %q hat nicht genau drei Felder (%v)", b, codes[i], body, err)
			}
			if resp.StatusCode != stati[i] || !bytes.Equal(body, fehlerobjekt(t, codes[i])) ||
				resp.Header.Get("Content-Type") != "application/json" ||
				resp.Header.Get("Content-Length") != fmt.Sprint(len(body)) ||
				resp.Header.Get("X-Correlation-Id") != id {
				t.Errorf("%s %v/%d: Antwort %d %q %v, ID %s", b, codes[i], stati[i], resp.StatusCode, body, resp.Header, id)
			}
		}
	}
}

// A27: Ein *Fehler mit unzulässigem Paar ergibt ODR-E9001 mit 500.
func TestFehlerUnzulaessig(t *testing.T) {
	paare := []inbound.Fehler{
		{Code: 0, Status: 500}, {Code: 9999, Status: 404}, {Code: model.CodeE2001, Status: 0},
		{Code: model.CodeE5001, Status: 200}, {Code: model.CodeE5001, Status: 0}, {Code: model.CodeE5001, Status: 600},
		{Code: model.CodeE4001, Status: 500},
	}
	for _, p := range paare {
		u := neuUmgebung(t, "record", fehlerDienst(&inbound.Fehler{Code: p.Code, Status: p.Status}))
		resp, body := hole(t, u, "GET", "/p")
		if resp.StatusCode != 500 || !bytes.Equal(body, fehlerobjekt(t, model.CodeE9001)) {
			t.Errorf("(%d, %d): Antwort %d %q, erwartet 500 mit ODR-E9001", p.Code, p.Status, resp.StatusCode, body)
		}
	}
}

// A28: Ein *Fehler wird auch verpackt und in errors.Join gefunden.
func TestFehlerVerpackt(t *testing.T) {
	f := &inbound.Fehler{Code: model.CodeE5003, Status: 400}
	formen := []error{
		fmt.Errorf("außen: %w", f), errors.Join(errors.New("x"), f), fmt.Errorf("a: %w", fmt.Errorf("b: %w", f)),
	}
	for i, e := range formen {
		u := neuUmgebung(t, "record", fehlerDienst(e))
		resp, body := hole(t, u, "GET", "/p")
		if resp.StatusCode != 400 || !bytes.Equal(body, fehlerobjekt(t, model.CodeE5003)) {
			t.Errorf("Form %d: Antwort %d %q, erwartet 400 mit ODR-E5003", i+1, resp.StatusCode, body)
		}
	}
}

// A29: Ein Fehler ohne Code ergibt ODR-E9001 mit 500, auch model.ErrUngueltig.
func TestFehlerOhneCode(t *testing.T) {
	for _, e := range []error{model.ErrUngueltig, errors.New("x-text")} {
		u := neuUmgebung(t, "record", fehlerDienst(e))
		resp, body := hole(t, u, "GET", "/p")
		if resp.StatusCode != 500 || !bytes.Equal(body, fehlerobjekt(t, model.CodeE9001)) {
			t.Errorf("%v: Antwort %d %q, erwartet 500 mit ODR-E9001", e, resp.StatusCode, body)
		}
	}
}

// A30: Ein Kontextfehler ohne Code bekommt keine Antwort; er gewinnt gegen einen *Fehler.
func TestKontextfehler(t *testing.T) {
	fehler := map[string]error{"Canceled": context.Canceled, "DeadlineExceeded": context.DeadlineExceeded}
	for name, ce := range fehler {
		huellen := []error{
			ce, fmt.Errorf("h: %w", ce),
			errors.Join(ce, &inbound.Fehler{Code: model.CodeE5001, Status: 404}),
			errors.Join(model.ErrUngueltig, ce),
		}
		for i, e := range huellen {
			k := ausBytes([]byte("abc"))
			u := neuUmgebung(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
				return inbound.Antwort{Status: 200, Body: k}, e
			})
			req, _ := http.NewRequest("GET", "http://"+u.addr+"/p", nil)
			resp, err := u.client().Do(req)
			if err == nil {
				body, _ := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				t.Errorf("%s Hülle %d: Antwort %d %q, erwartet Abbruch", name, i+1, resp.StatusCode, body)
			}
			warte(t, "Close", func() bool { c, _, _, _ := k.stand(); return c == 1 })
		}
	}
}

// A31: Das Fehlerobjekt enthält keinen Wert des Requests und keine Ursache.
func TestFehlerObjektKeineWerte(t *testing.T) {
	const canary = "CANARY-7f3a"
	stellen := []string{"Pfad", "Query", "Header", "Body", "Ursache"}
	arten := []string{"Fehler mit Ursache", "ErrUngueltig"}
	for _, stelle := range stellen {
		for _, art := range arten {
			u := neuUmgebung(t, "record", func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
				wert := canary
				switch stelle {
				case "Pfad":
					wert = a.Pfad
				case "Query":
					wert = a.RawQuery
				case "Header":
					wert = erster(a.Header["X-C"])
				case "Body":
					b, _ := io.ReadAll(a.Body)
					wert = string(b)
				}
				ursache := errors.New("Ursache " + wert)
				if art == "ErrUngueltig" {
					return inbound.Antwort{}, errors.Join(model.ErrUngueltig, ursache)
				}
				return inbound.Antwort{}, &inbound.Fehler{Code: model.CodeE5001, Status: 404, Ursache: ursache}
			})
			pfad, query, kopf, body := "/p", "", "", ""
			switch stelle {
			case "Pfad":
				pfad = "/" + canary
			case "Query":
				query = "?q=" + canary
			case "Header":
				kopf = canary
			case "Body":
				body = canary
			}
			req, _ := http.NewRequest("POST", "http://"+u.addr+pfad+query, strings.NewReader(body))
			req.Header.Set("X-C", kopf)
			resp, err := u.client().Do(req)
			if err != nil {
				t.Fatalf("%s/%s: %v", stelle, art, err)
			}
			rest, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if strings.Contains(string(rest), canary) || strings.Contains(fmt.Sprint(resp.Header), canary) {
				t.Errorf("%s/%s: die Antwort enthält den Canary: %q %v", stelle, art, rest, resp.Header)
			}
		}
	}
}

// A32: Bei HEAD entfällt der Body des Fehlerobjekts, die Header bleiben.
func TestFehlerHead(t *testing.T) {
	u := neuUmgebung(t, "record", fehlerDienst(&inbound.Fehler{Code: model.CodeE5001, Status: 404}))
	resp, body := hole(t, u, "HEAD", "/p")
	soll := len(fehlerobjekt(t, model.CodeE5001))
	if resp.StatusCode != 404 || len(body) != 0 || resp.Header.Get("Content-Length") != fmt.Sprint(soll) {
		t.Errorf("HEAD: Status %d, %d Body-Bytes, Content-Length %q; erwartet 404, 0, %d", resp.StatusCode, len(body), resp.Header.Get("Content-Length"), soll)
	}
}

// A33: X-Correlation-Id steht nur auf Fehlerantworten.
func TestKorrelationsHeader(t *testing.T) {
	eigen := neuUmgebung(t, "record", liefert(inbound.Antwort{Status: 200, Header: map[string][]string{"X-Correlation-Id": {"fremd"}}, Body: leer()}))
	resp, _ := hole(t, eigen, "GET", "/p")
	gleich(t, "Erfolg mit eigenem Header", resp.Header["X-Correlation-Id"], []string{"fremd"})
	ohne := neuUmgebung(t, "record", nil)
	resp, _ = hole(t, ohne, "GET", "/p")
	gleich(t, "Erfolg ohne Header", resp.Header["X-Correlation-Id"], []string(nil))
	fehler := neuUmgebung(t, "record", fehlerDienst(&inbound.Fehler{Code: model.CodeE5001, Status: 404}))
	resp, _ = hole(t, fehler, "GET", "/p")
	gleich(t, "Fehler", resp.Header["X-Correlation-Id"], []string{fmt.Sprintf("cor-%032d", 1)})
}

// A34: Gleichzeitige Requests teilen nichts.
func TestGleichzeitig(t *testing.T) {
	const n = 64
	var mu sync.Mutex
	angekommen := 0
	alle := make(chan struct{})
	u := neuUmgebung(t, "record", func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		body, _ := io.ReadAll(a.Body)
		mu.Lock()
		angekommen++
		if angekommen == n {
			close(alle)
		}
		mu.Unlock()
		select {
		case <-alle:
		case <-time.After(5 * time.Second):
		}
		echo := fmt.Sprintf("%s|%s|%s|%s", a.Pfad, a.RawQuery, erster(a.Header["X-N"]), body)
		return inbound.Antwort{Status: 200, Header: map[string][]string{"X-Echo": {erster(a.Header["X-N"])}}, Body: ausBytes([]byte(echo))}, nil
	})
	var wg sync.WaitGroup
	var ids []string
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/p%d?q=%d", u.addr, i, i), strings.NewReader(fmt.Sprintf("body%d", i)))
			req.Header.Set("X-N", fmt.Sprint(i))
			resp, err := u.client().Do(req)
			if err != nil {
				t.Errorf("Request %d: %v", i, err)
				return
			}
			got, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			soll := fmt.Sprintf("/p%d|q=%d|%d|body%d", i, i, i, i)
			if string(got) != soll || resp.Header.Get("X-Echo") != fmt.Sprint(i) {
				t.Errorf("Request %d: Antwort %q (X-Echo %q), erwartet %q", i, got, resp.Header.Get("X-Echo"), soll)
			}
		}()
	}
	wg.Wait()
	for _, a := range u.d.alle() {
		ids = append(ids, a.a.KorrelationsID)
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if seen[id] {
			t.Errorf("KorrelationsID %q doppelt", id)
		}
		seen[id] = true
	}
	if len(seen) != n {
		t.Errorf("%d verschiedene KorrelationsIDs, erwartet %d", len(seen), n)
	}
}

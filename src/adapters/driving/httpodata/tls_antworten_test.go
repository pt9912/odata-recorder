package httpodata_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// art ist eine Antwortart des Adapters: was der Dienst für eine Anfrage liefert.
type art struct {
	name, methode string
	tu            func() (inbound.Antwort, error)
}

// antwortarten ist die Menge jeder Antwortart des Adapters (SPEC-136 Nr. 5 bis 9): Erfolg je Status, Fehler je Paar aus Code und Status (GET und HEAD), Fehler außerhalb der
// Tabelle, Antwort gegen den Vertrag, Abbruch vor und nach dem ersten Byte. Die Antwort nach Anhalten (503) führt der Test getrennt.
func antwortarten() []art {
	var out []art
	for _, s := range []int{200, 201, 204, 301, 304, 400, 404, 500, 599} {
		out = append(out, art{fmt.Sprintf("Erfolg %d", s), "GET", func() (inbound.Antwort, error) {
			return inbound.Antwort{Status: s, Header: map[string][]string{"X-Art": {"erfolg"}, "Content-Type": {"text/plain"}}, Body: ausBytes([]byte("ok"))}, nil
		}})
	}
	out = append(out, art{"Erfolg HEAD", "HEAD", func() (inbound.Antwort, error) {
		return inbound.Antwort{Status: 200, Header: map[string][]string{"Content-Length": {"2"}}, Body: ausBytes([]byte("ok"))}, nil
	}})
	codes, stati := codeStatusPaare()
	for i := range codes {
		for _, m := range []string{"GET", "HEAD"} {
			out = append(out, art{fmt.Sprintf("Fehler %v/%d %s", codes[i], stati[i], m), m, func() (inbound.Antwort, error) {
				return inbound.Antwort{}, &inbound.Fehler{Code: codes[i], Status: stati[i]}
			}})
		}
	}
	out = append(out,
		art{"Fehler mit Status außerhalb der Tabelle", "GET", func() (inbound.Antwort, error) {
			return inbound.Antwort{}, &inbound.Fehler{Code: model.CodeE5005, Status: 500}
		}},
		art{"Fehler ohne Code", "GET", func() (inbound.Antwort, error) { return inbound.Antwort{}, &inbound.Fehler{} }},
		art{"fremder Fehler", "GET", func() (inbound.Antwort, error) { return inbound.Antwort{}, errors.New("fremd") }},
		art{"Antwort ohne Body", "GET", func() (inbound.Antwort, error) { return inbound.Antwort{Status: 200}, nil }},
		art{"Status 199", "GET", func() (inbound.Antwort, error) { return inbound.Antwort{Status: 199, Body: leer()}, nil }},
		art{"Status 600", "GET", func() (inbound.Antwort, error) { return inbound.Antwort{Status: 600, Body: leer()}, nil }},
		art{"Abbruch vor dem ersten Byte", "GET", func() (inbound.Antwort, error) { return inbound.Antwort{}, context.Canceled }},
		art{"Abbruch nach dem ersten Stück", "GET", func() (inbound.Antwort, error) {
			n := 0
			return inbound.Antwort{Status: 200, Body: &koerper{lies: func(p []byte) (int, error) {
				if n++; n == 1 {
					return copy(p, "abc"), nil
				}
				return 0, errors.New("Lesefehler")
			}}}, nil
		}},
	)
	return out
}

// beob ist, was der Client von einer Antwort sah (Verbindungsdaten wie `Date` ausgenommen); Fehler heißt, die Anfrage oder das Lesen des Bodys scheiterte.
type beob struct {
	Status int
	Header map[string][]string
	Body   string
	Fehler bool
}

func beobachteArt(c *http.Client, url, methode string) beob {
	req, err := http.NewRequest(methode, url, nil)
	if err != nil {
		return beob{Fehler: true}
	}
	resp, err := c.Do(req)
	if err != nil {
		return beob{Fehler: true}
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	h := resp.Header.Clone()
	h.Del("Date")
	return beob{Status: resp.StatusCode, Header: h, Body: string(b), Fehler: err != nil}
}

// TestAntwortartenMitTLS (B13): die Antwort ist ohne und mit TLS dieselbe, für jede Antwortart des Adapters (SPEC-291 Nr. 9 d, SPEC-136 Nr. 5 bis 9): Status, Header und Body
// byte-gleich, die Anfrage scheitert in beiden oder in keinem. Menge: jede Art aus antwortarten (9 + 1 + 2 × Paare + 8) und die Antwort nach Anhalten, je ohne und mit TLS.
func TestAntwortartenMitTLS(t *testing.T) {
	arten := antwortarten()
	codes, _ := codeStatusPaare()
	if soll := 9 + 1 + 2*len(codes) + 8; len(arten) != soll {
		t.Fatalf("Menge hat %d Arten, erwartet %d", len(arten), soll)
	}
	lauf := func(mitTLS bool) []beob {
		tu := func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			i, err := strconv.Atoi(strings.TrimPrefix(a.Pfad, "/art/"))
			if err != nil {
				return inbound.Antwort{Status: 200, Body: ausBytes([]byte("ok"))}, nil
			}
			return arten[i].tu()
		}
		var url string
		var client, haelt *http.Client
		var anhalten func()
		if mitTLS {
			u := neuTLSUmgebung(t, tu, nil)
			url, client, anhalten = "https://"+u.addr, u.httpClient(), u.srv.Anhalten
			haelt = &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: u.klient(), ForceAttemptHTTP2: true}}
		} else {
			u := neuUmgebung(t, "record", tu)
			url, client, anhalten = "http://"+u.addr, u.client(), u.srv.Anhalten
			haelt = &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{}}
		}
		var aus []beob
		for i, a := range arten {
			aus = append(aus, beobachteArt(client, url+"/art/"+strconv.Itoa(i), a.methode))
		}
		if o := beobachteArt(haelt, url+"/p", "GET"); o.Status != 200 {
			t.Fatalf("mitTLS %v: Request vor Anhalten: %+v", mitTLS, o)
		}
		anhalten()
		return append(aus, beobachteArt(haelt, url+"/p", "GET"))
	}
	ohne, mit := lauf(false), lauf(true)
	if len(ohne) != len(arten)+1 || len(mit) != len(ohne) {
		t.Fatalf("Beobachtungen: %d und %d, erwartet %d", len(ohne), len(mit), len(arten)+1)
	}
	for i := range ohne {
		name := "Antwort nach Anhalten"
		if i < len(arten) {
			name = arten[i].name
		}
		if fmt.Sprintf("%+v", ohne[i]) != fmt.Sprintf("%+v", mit[i]) {
			t.Errorf("%s:\nohne TLS %+v\nmit TLS  %+v", name, ohne[i], mit[i])
		}
	}
	if got := ohne[len(ohne)-1]; got.Status != 503 || !strings.Contains(got.Body, "ODR-E5010") {
		t.Errorf("Antwort nach Anhalten ohne TLS: %+v, erwartet 503 ODR-E5010", got)
	}
	if got := ohne[len(arten)-1]; !got.Fehler || got.Body != "abc" {
		t.Errorf("Abbruch nach dem ersten Stück: %+v, erwartet Body abc und einen Fehler", got)
	}
}

// TestSchreibstellenDesAdapters (B13): die Zählprüfung der Menge „jede Stelle, an der der Adapter eine Antwort schreibt“ gegen die Quelle (go/parser): genau diese Funktionen der
// Dateien des Pakets rufen Header, WriteHeader, Write oder Flush; eine neue Stelle färbt den Test, bis TestAntwortartenMitTLS sie führt. Erreicht werden sie so:
// schreibeAntwort und kopiere durch die Arten „Erfolg“, schreibeFehlerobjekt durch jede Fehlerart, bearbeite (Connection: close) durch die Antwort nach Anhalten.
func TestSchreibstellenDesAdapters(t *testing.T) {
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	gefunden := map[string][]string{}
	fset := token.NewFileSet()
	for _, e := range eintraege {
		if n := e.Name(); !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					if s, ok := c.Fun.(*ast.SelectorExpr); ok && slices.Contains([]string{"Header", "WriteHeader", "Write", "Flush"}, s.Sel.Name) {
						k := e.Name() + ":" + fn.Name.Name
						if !slices.Contains(gefunden[k], s.Sel.Name) {
							gefunden[k] = append(gefunden[k], s.Sel.Name)
						}
					}
				}
				return true
			})
		}
	}
	for _, v := range gefunden {
		slices.Sort(v)
	}
	soll := map[string][]string{
		"antwort.go:kopiere":             {"Flush", "Write"},
		"antwort.go:schreibeAntwort":     {"Header", "WriteHeader"},
		"fehler.go:schreibeFehlerobjekt": {"Header", "Write", "WriteHeader"},
		"server.go:bearbeite":            {"Header"},
	}
	if fmt.Sprint(gefunden) != fmt.Sprint(soll) {
		t.Errorf("Schreibstellen %v, erwartet %v", gefunden, soll)
	}
}

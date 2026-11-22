package replay_test

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// TestTelemetrieImQuelltext hält W11 (SPEC-296 Nr. 1, 6): im Quelltext des Pakets ohne _test.go ruft der Dienst von Telemetrie und Span
// allein StartSpan, Set, End und Log, je einmal; nie CountError oder Snapshot.
func TestTelemetrieImQuelltext(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	zaehler := map[string]int{}
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if ce, ok := n.(*ast.CallExpr); ok {
				if sel, ok := ce.Fun.(*ast.SelectorExpr); ok {
					zaehler[sel.Sel.Name]++
				}
			}
			return true
		})
	}
	for name, n := range map[string]int{"StartSpan": 1, "Set": 1, "End": 1, "Log": 1, "CountError": 0, "Snapshot": 0} {
		if zaehler[name] != n {
			t.Errorf("%s im Quelltext %d-mal, erwartet %d", name, zaehler[name], n)
		}
	}
}

// TestKeinZaehlerImDienst hält W11 (SPEC-296 Nr. 5): der Dienst ruft CountError nie; für jeden der 13 Codes von SPEC-135 Nr. 15 ein Szenario
// und für jede der 2 Arten des Kontextfehlers an jeder der 8 Stellen (S1, S2 über den Body, S3 bis S8) ein Lauf: 0 Aufrufe. Zählprüfung 13, 8
// und 2.
func TestKeinZaehlerImDienst(t *testing.T) {
	szen := szenarien()
	for _, sz := range szen {
		u := neu(t)
		u.mitJournal()
		resp, err := u.replay(context.Background(), sz.baue(u, model.StrategySequential))
		if err == nil {
			_ = resp.Body.Close()
		}
		if n := u.tel.Anzahl(supporttest.OpCountError); n != 0 {
			t.Errorf("%s: CountError %d-mal gerufen", sz.code, n)
		}
	}
	stellen8 := []string{"Sessions.Get", "", "Exchanges.ListBySession", "Transaktion.WithWorkspaceTransaction", "tx Workspaces.Get", "tx Cursors.Get",
		"Blobs.Open", "tx Cursors.Put"}
	for _, k := range kontextfehler() {
		for _, op := range stellen8 {
			u := neu(t)
			u.mitJournal()
			_, err := u.replay(context.Background(), kontextAn(op, k)(u))
			erwarteOhneCode(t, "Kontext "+op, err, k)
			if n := u.tel.Anzahl(supporttest.OpCountError); n != 0 {
				t.Errorf("Kontext %s: CountError %d-mal gerufen", op, n)
			}
		}
	}
	if len(szen) != 13 || len(stellen8) != 8 || len(kontextfehler()) != 2 {
		t.Errorf("Menge: %d Codes, %d Stellen und %d Kontextfehler, erwartet 13, 8 und 2", len(szen), len(stellen8), len(kontextfehler()))
	}
}

// kanarie ist der Wert des Requests in der Form Token, aus Bytewerten gebildet (AGENTS.md §3.18).
func kanarie() string { return string([]byte{0x63, 0x61, 0x6e, 0x61, 0x72, 0x79}) }

// legeFuer legt, wenn treffer, den Exchange mit dem Key des Requests a (und seines Bodys) an; ohne ihn endet die Auswahl mit ODR-E5001.
func (u *umg) legeFuer(a inbound.Anfrage, body []byte, treffer bool) {
	u.t.Helper()
	if !treffer {
		return
	}
	rr := model.RedactionNone
	if body != nil {
		body, rr = redigiert(u.t, body, u.k.Pfade)
	}
	s, err := model.BildeSchluessel(model.SchluesselEingabe{Method: a.Methode, ServiceID: svcID, Path: strings.TrimPrefix(a.Pfad, root),
		RawQuery: a.RawQuery, Header: a.Header, Body: body, RequestRedaction: rr}, u.k.Regeln)
	if err != nil {
		u.t.Fatalf("BildeSchluessel: %v", err)
	}
	e := u.austausch(1, s.MatchKey, true)
	e.KeyRedacted = true
	u.lege(e)
}

// stellenKanarie sind die acht Stellen des Requests (SPEC-296 Nr. 8): Pfad, Name und Wert der Query, Name und Wert eines Headers, Body, Wert des
// Headers X-Recorder-Replay-Session und der Text der Ursache eines Fehlers eines Ports.
func stellenKanarie() map[string]func(u *umg, k string, treffer bool) inbound.Anfrage {
	mit := func(setze func(a *inbound.Anfrage, k string)) func(u *umg, k string, treffer bool) inbound.Anfrage {
		return func(u *umg, k string, treffer bool) inbound.Anfrage {
			a := anf()
			setze(&a, k)
			u.legeFuer(a, nil, treffer)
			return a
		}
	}
	stellen := map[string]func(u *umg, k string, treffer bool) inbound.Anfrage{
		"Pfad":          mit(func(a *inbound.Anfrage, k string) { a.Pfad += "/" + k }),
		"Query-Name":    mit(func(a *inbound.Anfrage, k string) { a.RawQuery = k + "=1" }),
		"Query-Wert":    mit(func(a *inbound.Anfrage, k string) { a.RawQuery = "q=" + k }),
		"Header-Name":   mit(func(a *inbound.Anfrage, k string) { a.Header["X-"+k] = []string{"v"} }),
		"Header-Wert":   mit(func(a *inbound.Anfrage, k string) { a.Header["X-Wert"] = []string{k} }),
		"Replay-Header": mit(func(a *inbound.Anfrage, k string) { a.Header["X-Recorder-Replay-Session"] = []string{k} }),
		"Body": func(u *umg, k string, treffer bool) inbound.Anfrage {
			a := anf()
			body := `"` + k + `"`
			a.Header["Content-Type"] = []string{"application/json"}
			a.Body, a.Laenge = strings.NewReader(body), int64(len(body))
			u.legeFuer(a, []byte(body), treffer)
			return a
		},
	}
	// Text der Ursache: je Port mit Fehlerrückgabe (S1 bis S8, S2 über den Body) in zwei Formen, der Fehler mit Code (der Dienst bildet ihn und
	// hält die Ursache über Unwrap) und der Kontextfehler ohne Code (er kehrt unverändert zurück); die Art Treffer ist der fehlende Cursor,
	// dessen Fehler den Text trägt und kein Fehler ist.
	for _, p := range []struct{ name, op string }{{"S1", "Sessions.Get"}, {"S2", ""}, {"S3", "Exchanges.ListBySession"}, {"S4", "Transaktion.WithWorkspaceTransaction"},
		{"S5", "tx Workspaces.Get"}, {"S6", "tx Cursors.Get"}, {"S7", "Blobs.Open"}, {"S8", "tx Cursors.Put"}} {
		ursache := error(outbound.ErrBeschaedigt)
		switch p.op {
		case "":
			ursache = io.ErrUnexpectedEOF
		case "Blobs.Open":
			ursache = outbound.ErrBlobFehlt
		}
		for form, cause := range map[string]error{"Code": ursache, "Kontext": context.Canceled} {
			stellen["Ursache "+p.name+" "+form] = func(u *umg, k string, treffer bool) inbound.Anfrage {
				if treffer {
					a := erfolgsLauf(u)
					u.p.setzeFehler("tx Cursors.Get", errors.Join(outbound.ErrNichtGefunden, errors.New(k)))
					return a
				}
				return kontextAn(p.op, errors.Join(cause, errors.New(k)))(u)
			}
		}
	}
	return stellen
}

// TestKeinRequestWertImSpan hält W12 (SPEC-296 Nr. 8): Menge M-Stellen-Replay, 7 Stellen des Requests und die achte, der Text der Ursache, an
// jedem der 8 Ports in 2 Formen (16) gleich 23, × 2 Arten (Treffer, Fehler) gleich 46; ein Canary in der Form Token an einer Stelle erreicht
// keinen Schlüssel und keinen Wert eines Aufrufs des Fakes (StartSpan, Set, Log, Text des Eintrags). Die Senke des echten Adapters prüft der
// Test des Composition Root. Zählprüfung 8 Stellen der Spezifikation (7 und die achte), 8 Ports.
func TestKeinRequestWertImSpan(t *testing.T) {
	k := kanarie()
	stellen := stellenKanarie()
	n := 0
	for name, baue := range stellen {
		for _, treffer := range []bool{true, false} {
			n++
			was := name + " treffer=" + map[bool]string{true: "ja", false: "nein"}[treffer]
			u := neu(t)
			u.mitJournal()
			resp, err := u.replay(context.Background(), baue(u, k, treffer))
			if (err == nil) != treffer {
				t.Errorf("%s: Fehler %v", was, err)
			}
			if err == nil {
				_ = resp.Body.Close()
			}
			if len(buecher(u.tel)) != 1 {
				t.Errorf("%s: kein Span", was)
			}
			for _, a := range u.tel.Aufrufe() {
				for _, at := range a.Attrs {
					if strings.Contains(at.Key, k) || strings.Contains(at.Text, k) {
						t.Errorf("%s: Attribut %+v nennt den Canary", was, at)
					}
				}
				if strings.Contains(a.Msg, k) {
					t.Errorf("%s: Text %q nennt den Canary", was, a.Msg)
				}
			}
		}
	}
	if len(stellen) != 7+8*2 || n != 2*(7+8*2) {
		t.Errorf("Menge: %d Stellen (7 des Requests und die Ursache an 8 Ports in 2 Formen), %d Elemente, erwartet 23 und 46", len(stellen), n)
	}
}

package mock_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// TestPaketOhneZustand hält Z22 in jedem Lauf (SPEC-195 Nr. 12): im Quelltext des Pakets steht keine Paketvariable außer
// der Prüfung der Schnittstelle, keine Sperre, keine Goroutine, kein Kanal, keine Uhr, kein Zufall, kein Typ mit Methoden
// außer Mocker; der Dienst importiert keinen Adapter, keinen anderen Dienst und kein Paket außerhalb von Domain und Ports.
func TestPaketOhneZustand(t *testing.T) {
	verboten := []string{"sync", "sync/atomic", "time", "math/rand", "math/rand/v2", "crypto/rand", "os", "net", "net/http", "database/sql"}
	fset := token.NewFileSet()
	for _, d := range quelltext(t) {
		f, err := parser.ParseFile(fset, d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if i := strings.Index(p, "/src/"); i >= 0 {
				rest := p[i+len("/src/"):]
				if !strings.HasPrefix(rest, "hexagon/model") && !strings.HasPrefix(rest, "hexagon/ports/") {
					t.Errorf("%s importiert %s", d, p)
				}
			}
			for _, v := range verboten {
				if p == v {
					t.Errorf("%s importiert %s", d, p)
				}
			}
		}
		for _, decl := range f.Decls {
			switch x := decl.(type) {
			case *ast.GenDecl:
				for _, s := range x.Specs {
					if v, ok := s.(*ast.ValueSpec); ok && x.Tok == token.VAR {
						for _, n := range v.Names {
							if n.Name != "_" {
								t.Errorf("%s: Paketvariable %s", d, n.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if x.Recv == nil {
					continue
				}
				typ := x.Recv.List[0].Type
				if st, ok := typ.(*ast.StarExpr); ok {
					typ = st.X
				}
				if id, ok := typ.(*ast.Ident); !ok || id.Name != "Mocker" {
					t.Errorf("%s: Methode %s an einem Typ außer Mocker", d, x.Name.Name)
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.GoStmt:
				t.Errorf("%s: Goroutine", d)
			case *ast.ChanType, *ast.SendStmt:
				t.Errorf("%s: Kanal", d)
			}
			return true
		})
	}
}

// ergebnis fasst Antwort oder Fehler in einen Text: der Code des Fehlers, sonst Status, Header und Body.
func ergebnis(a inbound.Antwort, err error) string {
	if err != nil {
		var f *inbound.Fehler
		if errors.As(err, &f) {
			return "fehler " + f.Code.String()
		}
		return "roh " + err.Error()
	}
	b, rerr := io.ReadAll(a.Body)
	if rerr != nil {
		return "lesefehler " + rerr.Error()
	}
	_ = a.Body.Close()
	return fmt.Sprintf("%d %v %q", a.Status, a.Header, b)
}

// urlFuerGleichzeitig liest die drei Pfade, die TestGleichzeitig stellt: $metadata, ein Set und alles andere (Segment).
func urlFuerGleichzeitig(p, _ string) (model.ODataURL, error) {
	switch p {
	case "/$metadata":
		return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMetadaten}}, nil
	case "/Set":
		return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: "Set"}}, nil
	case "/art":
		return model.ODataURL{}, nil
	}
	return model.ODataURL{}, &model.URLFehler{Ursache: model.URLSegment}
}

// TestGleichzeitig hält Z22 für gleichzeitige Aufrufe: 64 Goroutinen × 18 Requests gegen einen Dienst mit sieben Services
// (Standard, leere Metadaten, ohne Eintrag, fehlender Blob, Fehler der Probe, nur lesend, Root `/`), in drei Klassen von
// Workspaces (unbekannt, fremd, eigen) und mit einem Kontextfehler des Ports, geben je Request die Antwort des Einzelaufrufs;
// unter `make test-race` meldet der Detektor einen ungeordneten Zugriff des Dienstes. Jeder Zweig von `Mock` und seiner
// Helfer läuft darin, außer New und der Prüfung der Konfig (kein gleichzeitiger Aufruf).
func TestGleichzeitig(t *testing.T) {
	const goroutinen, requests = 64, 18
	u := neueUmg(t)
	u.c.url.f = urlFuerGleichzeitig
	probe := neueCodecs(neuesProt(), model.Kopfzeile{Name: "X-Probe", Wert: "p"})
	probe.url.f = urlFuerGleichzeitig
	probe.p.setzeFehler("Schreiber.SchreibeAnzahl", errors.New("roh "+kanarienwert))
	lesend := neueCodecs(neuesProt(), model.Kopfzeile{Name: "X-Lesend", Wert: "l"})
	lesend.url.f = urlFuerGleichzeitig
	u.k.Dienste = append(u.k.Dienste,
		mock.Dienst{ID: "leer", Root: "/leer"}, mock.Dienst{ID: "fehlt", Root: "/fehlt"}, mock.Dienst{ID: "kaputt", Root: "/kaputt"},
		mock.Dienst{ID: "probe", Root: "/probe", Codecs: probe.satz(false, "")}, mock.Dienst{ID: "lesend", Root: "/lesend", Codecs: lesend.satz(true, "")},
		mock.Dienst{ID: "wurzel", Root: "/"})
	u.md.setze(model.Service{ID: "leer"})
	u.md.setze(model.Service{ID: "kaputt", MetadataBlob: "sha256:fehlt"})
	u.md.setze(model.Service{ID: "probe", MetadataBlob: metaRef})
	u.md.setze(model.Service{ID: "lesend", MetadataBlob: metaRef})
	u.ws.ersatz = func(id string) (model.Workspace, error) {
		name, dienst, _ := strings.Cut(id, "|")
		if strings.HasPrefix(name, "k") {
			return model.Workspace{}, context.Canceled
		}
		g, _ := strconv.Atoi(strings.TrimPrefix(name, "w"))
		switch g % 3 {
		case 0:
			return model.Workspace{}, outbound.ErrNichtGefunden
		case 1:
			return model.Workspace{ID: id, ServiceID: "anderer"}, nil
		}
		return model.Workspace{ID: id, ServiceID: dienst, BaselineSnapshotID: "b"}, nil
	}
	anfragen := func(g, i int) inbound.Anfrage {
		dienst, root := dienstID, dienstRoot
		anders := map[int][2]string{8: {"leer", "/leer"}, 9: {"fehlt", "/fehlt"}, 10: {"kaputt", "/kaputt"}, 11: {"probe", "/probe"},
			12: {"lesend", "/lesend"}, 17: {"wurzel", "/x"}}
		if r, da := anders[i]; da {
			dienst, root = r[0], r[1]
		}
		a := anfrage("GET", root+"/$metadata")
		a.Header[headerName] = []string{"w" + strconv.Itoa(g) + "|" + dienst}
		switch i {
		case 1:
			a.Methode = "HEAD"
		case 2, 12:
			a.Methode = "POST"
		case 3:
			a.Pfad = ""
		case 4:
			delete(a.Header, headerName)
		case 13:
			a.Laenge = -2
		case 14:
			a.Pfad = dienstRoot + "/art"
		case 15:
			a.Header[headerName] = []string{"k" + strconv.Itoa(g) + "|" + dienst}
		case 16:
			a.Pfad = dienstRoot
		case 5:
			a.Pfad = dienstRoot + "/Set"
		case 6:
			a.Pfad = dienstRoot + "/bad"
		case 7:
			a.RawQuery = "x=" + strconv.Itoa(g)
		}
		return a
	}
	m := u.mocker()
	var soll [goroutinen][requests]string
	for g := range soll {
		for i := range soll[g] {
			soll[g][i] = ergebnis(m.Mock(u.ctx, anfragen(g, i)))
		}
	}
	if soll[2][0] != soll[5][0] || soll[2][0] == soll[1][0] || soll[2][0] == soll[0][0] {
		t.Fatalf("Klassen der Workspaces nicht unterscheidbar: %q %q %q", soll[0][0], soll[1][0], soll[2][0])
	}
	verschieden := map[string]bool{}
	for _, s := range soll[2] {
		verschieden[s] = true
	}
	if len(verschieden) < 13 {
		t.Errorf("nur %d verschiedene Ergebnisse unter %d Requests: %q", len(verschieden), requests, soll[2])
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var abweichungen []string
	start := make(chan struct{})
	for g := 0; g < goroutinen; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < requests; i++ {
				if got := ergebnis(m.Mock(u.ctx, anfragen(g, i))); got != soll[g][i] {
					mu.Lock()
					abweichungen = append(abweichungen, fmt.Sprintf("g%d i%d: %q, erwartet %q", g, i, got, soll[g][i]))
					mu.Unlock()
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	if len(abweichungen) != 0 {
		t.Errorf("%d Abweichungen, erste: %s", len(abweichungen), abweichungen[0])
	}
}

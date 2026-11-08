package admin_test

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestNewNilFelder hält K1: Listener, IDs, Audit und Telemetrie je nil sind model.ErrUngueltig ohne Server; der Gutfall steht daneben.
func TestNewNilFelder(t *testing.T) {
	loescher := map[string]func(k *admin.Konfig){
		"Listener":   func(k *admin.Konfig) { k.Listener = nil },
		"IDs":        func(k *admin.Konfig) { k.IDs = nil },
		"Audit":      func(k *admin.Konfig) { k.Audit = nil },
		"Telemetrie": func(k *admin.Konfig) { k.Telemetrie = nil },
	}
	if len(loescher) != 4 {
		t.Fatalf("%d Felder, erwartet 4", len(loescher))
	}
	for name, loesche := range loescher {
		k, _ := konfigMit()
		loesche(&k)
		srv, err := admin.New(k)
		if !errors.Is(err, model.ErrUngueltig) || srv != nil {
			t.Errorf("%s nil: Fehler %v, Server %v", name, err, srv)
		}
	}
	k, _ := konfigMit()
	if srv, err := admin.New(k); err != nil || srv == nil {
		t.Errorf("Gutfall: Fehler %v, Server %v", err, srv)
	}
}

// TestNewTokenLaenge hält K2: unter 32 Bytes ist ungültig, 32 sind gültig; der Text nennt das Token nicht.
func TestNewTokenLaenge(t *testing.T) {
	for _, n := range []int{0, 1, 31, 32, 33} {
		k, _ := konfigMit()
		k.Token = bytes.Repeat([]byte(canary()), 3)[:n]
		srv, err := admin.New(k)
		if gueltig := n >= 32; gueltig != (err == nil && srv != nil) {
			t.Errorf("Token von %d Bytes: Fehler %v, Server %v", n, err, srv)
		}
		if n < 32 && (!errors.Is(err, model.ErrUngueltig) || srv != nil) {
			t.Errorf("Token von %d Bytes: Fehler %v, Server %v, erwartet ErrUngueltig ohne Server", n, err, srv)
		}
		if err != nil && n > 0 && (strings.Contains(err.Error(), string(k.Token)) || strings.Contains(err.Error(), string(k.Token[:min(n, 4)]))) {
			t.Errorf("Token von %d Bytes: der Text nennt das Token: %q", n, err)
		}
	}
}

// stabileGoroutinen liefert die Zahl der Goroutinen, sobald sie zweimal im Abstand von 10 ms gleich ist.
func stabileGoroutinen() int {
	letzte := runtime.NumGoroutine()
	for range 100 {
		time.Sleep(10 * time.Millisecond)
		jetzt := runtime.NumGoroutine()
		if jetzt == letzte {
			break
		}
		letzte = jetzt
	}
	return letzte
}

// TestNewRuftNichts hält K3: New ruft keinen Port und keine Methode des Listeners und startet keine Goroutine.
func TestNewRuftNichts(t *testing.T) {
	k, u := konfigMit(admin.Route{Methode: "GET", Muster: "/a", Handler: neueRoute().handle})
	vorher := stabileGoroutinen()
	srv, err := admin.New(k)
	nachher := stabileGoroutinen()
	if err != nil || srv == nil {
		t.Fatalf("New: %v", err)
	}
	if vorher != nachher {
		t.Errorf("Goroutinen vor New %d, danach %d", vorher, nachher)
	}
	if n := u.ids.alle(); n != 0 {
		t.Errorf("IDs: %d Aufrufe, erwartet 0", n)
	}
	u.keinePortAufrufe(t, "nach New")
}

// listenNamen sind die Namen, die der Adapter nirgends rufen darf (SPEC-238 Nr. 1: keine Standardadresse, kein Socket).
func listenNamen() []string {
	return []string{"Listen", "ListenTCP", "ListenAndServe", "ListenAndServeTLS"}
}

// zaehleListenAufrufe zählt die Aufrufe der Namen aus listenNamen in einer Quelldatei.
func zaehleListenAufrufe(t *testing.T, name string, quelle any) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, 0)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	n := 0
	ast.Inspect(f, func(nd ast.Node) bool {
		c, ok := nd.(*ast.CallExpr)
		if !ok {
			return true
		}
		var ident string
		switch fn := c.Fun.(type) {
		case *ast.Ident:
			ident = fn.Name
		case *ast.SelectorExpr:
			ident = fn.Sel.Name
		}
		for _, l := range listenNamen() {
			if ident == l {
				n++
			}
		}
		return true
	})
	return n
}

// zaehleGoInFunktion zählt die go-Anweisungen im Rumpf der Funktion name einer Quelle.
func zaehleGoInFunktion(t *testing.T, quelle, name string) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "quelle.go", quelle, 0)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name {
			ast.Inspect(fd.Body, func(nd ast.Node) bool {
				if _, ok := nd.(*ast.GoStmt); ok {
					n++
				}
				return true
			})
		}
	}
	return n
}

// TestNewStartetKeineGoroutine hält K3 gegen eine Goroutine, die vor der Messung endet: in der Funktion New steht
// keine go-Anweisung. Der Zähler wird an einer Quelle mit zwei go-Anweisungen in New und einer in einer anderen Funktion
// gehalten. Grenze: eine Goroutine in einer Funktion, die New ruft, fängt allein TestNewRuftNichts, und nur solange sie läuft.
func TestNewStartetKeineGoroutine(t *testing.T) {
	probe := "package x\nfunc New() { go f(); go func() {}() }\nfunc andere() { go g() }"
	if n := zaehleGoInFunktion(t, probe, "New"); n != 2 {
		t.Fatalf("Zähler zählt %d go-Anweisungen in New der Probe, erwartet 2", n)
	}
	src, err := os.ReadFile("admin.go")
	if err != nil {
		t.Fatal(err)
	}
	if n := zaehleGoInFunktion(t, string(src), "New"); n != 0 {
		t.Errorf("%d go-Anweisungen in New", n)
	}
}

// TestKeinListenImPaket hält K3: im Paket ohne _test.go steht kein Aufruf von Listen, ListenTCP, ListenAndServe oder
// ListenAndServeTLS. Der Zähler wird an einer Quelle mit allen vier Namen gegen die Menge gehalten.
func TestKeinListenImPaket(t *testing.T) {
	probe := `package x
func f() { net.Listen("tcp", ":1"); net.ListenTCP("tcp", nil); srv.ListenAndServe(); srv.ListenAndServeTLS("", ""); Listen() }`
	if n := zaehleListenAufrufe(t, "probe.go", probe); n != len(listenNamen())+1 || len(listenNamen()) != 4 {
		t.Fatalf("Zähler zählt %d Aufrufe in der Probe, erwartet %d (vier Namen und ein bloßes Listen)", n, len(listenNamen())+1)
	}
	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	geprueft := 0
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		geprueft++
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		if n := zaehleListenAufrufe(t, d, src); n != 0 {
			t.Errorf("%s: %d Aufrufe von Listen-Funktionen", d, n)
		}
	}
	if geprueft < 3 {
		t.Errorf("%d Dateien des Pakets geprüft, erwartet mindestens admin.go, route.go und tokendatei.go", geprueft)
	}
}

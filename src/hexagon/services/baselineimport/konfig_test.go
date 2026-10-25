package baselineimport_test

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

func ohneCode(err error) bool {
	var f *inbound.Fehler
	return !errors.As(err, &f)
}

// TestZeileA1KonfigUngueltig: jede Prüfung von SPEC-127 Nr. 1 je für sich (Menge 23) und die
// Gutfälle der Grenze von MaxCaptureBody (2).
func TestZeileA1KonfigUngueltig(t *testing.T) {
	w := neu(t)
	faelle := map[string]func(*baselineimport.Konfig){
		"Sessions nil":       func(k *baselineimport.Konfig) { k.Sessions = nil },
		"Exchanges nil":      func(k *baselineimport.Konfig) { k.Exchanges = nil },
		"Blobs nil":          func(k *baselineimport.Konfig) { k.Blobs = nil },
		"Metadaten nil":      func(k *baselineimport.Konfig) { k.Metadaten = nil },
		"Snapshots nil":      func(k *baselineimport.Konfig) { k.Snapshots = nil },
		"Imports nil":        func(k *baselineimport.Konfig) { k.Imports = nil },
		"MetadatenCodec nil": func(k *baselineimport.Konfig) { k.MetadatenCodec = nil },
		"DatenCodec nil":     func(k *baselineimport.Konfig) { k.DatenCodec = nil },
		"Uhr nil":            func(k *baselineimport.Konfig) { k.Uhr = nil },
		"IDs nil":            func(k *baselineimport.Konfig) { k.IDs = nil },
		"Dienste nil":        func(k *baselineimport.Konfig) { k.Dienste = nil },
		"Dienst ohne ID":     func(k *baselineimport.Konfig) { k.Dienste = []baselineimport.Dienst{{UpstreamURL: upstream}} },
		"Dienst doppelt": func(k *baselineimport.Konfig) {
			k.Dienste = []baselineimport.Dienst{{ID: "a", UpstreamURL: upstream}, {ID: "a", UpstreamURL: upstream}}
		},
		"URL leer":             func(k *baselineimport.Konfig) { k.Dienste = []baselineimport.Dienst{{ID: "a"}} },
		"URL ohne Schema":      func(k *baselineimport.Konfig) { k.Dienste = []baselineimport.Dienst{{ID: "a", UpstreamURL: "/x"}} },
		"URL ftp":              func(k *baselineimport.Konfig) { k.Dienste = []baselineimport.Dienst{{ID: "a", UpstreamURL: "ftp://h"}} },
		"MaxCaptureBody 0":     func(k *baselineimport.Konfig) { k.MaxCaptureBody = 0 },
		"MaxCaptureBody -1":    func(k *baselineimport.Konfig) { k.MaxCaptureBody = -1 },
		"InvalidEncoding leer": func(k *baselineimport.Konfig) { k.Regeln.InvalidEncoding = "" },
		"InvalidEncoding x":    func(k *baselineimport.Konfig) { k.Regeln.InvalidEncoding = "x" },
		"BodyMode leer":        func(k *baselineimport.Konfig) { k.Regeln.BodyMode = "" },
		"BodyMode x":           func(k *baselineimport.Konfig) { k.Regeln.BodyMode = "x" },
	}
	// Eine URL ohne Namen des Hosts (SPEC-127 Nr. 1): ohne Autorität, nur Port, nur Zugangsdaten, nur Port mit Schema, leerer Port, leere Klammern.
	for _, u := range []string{"http:///x", "http://:8080/x", "http://u@/x", "https://:443", "http://:/x", "http://[]/x"} {
		faelle["URL ohne Namen des Hosts "+u] = func(k *baselineimport.Konfig) {
			k.Dienste = []baselineimport.Dienst{{ID: "a", UpstreamURL: u}}
		}
	}
	if len(faelle) != 28 {
		t.Fatalf("Menge %d, erwartet 28", len(faelle))
	}
	for name, aendere := range faelle {
		k := w.konfig
		aendere(&k)
		imp, err := baselineimport.New(k)
		if !errors.Is(err, model.ErrUngueltig) || imp != nil || !ohneCode(err) {
			t.Errorf("%s: %v, Dienst %v", name, err, imp != nil)
		}
	}
	k := w.konfig
	k.Dienste = []baselineimport.Dienst{}
	if imp, err := baselineimport.New(k); !errors.Is(err, model.ErrUngueltig) || imp != nil {
		t.Errorf("Dienste leer: %v", err)
	}
	for _, r := range []model.SchluesselRegeln{
		{InvalidEncoding: model.KodierungReject, BodyMode: model.BodyBytes},
		{InvalidEncoding: model.KodierungLiteral, BodyMode: model.BodyJSONCanonical},
	} {
		k := w.konfig
		k.Regeln = r
		if imp, err := baselineimport.New(k); err != nil || imp == nil {
			t.Errorf("Regeln %+v: %v", r, err)
		}
	}
	for _, grenze := range []int64{1, math.MaxInt64} {
		k := w.konfig
		k.MaxCaptureBody = grenze
		if imp, err := baselineimport.New(k); err != nil || imp == nil {
			t.Errorf("MaxCaptureBody %d: %v", grenze, err)
		}
	}
}

// TestZeileA1UpstreamNormiert: die UpstreamURL gilt ohne einen abschließenden `/` ihres Pfades, auch bei einem Pfad
// mit `%2F` (SPEC-127 Nr. 1). Sechs Pfade; beobachtet an der Auflösung des Links: der absolute Link mit dem
// Pfad ohne den Schrägstrich ist die nächste Seite, ein Link, der das `%2F` dekodiert, nicht.
func TestZeileA1UpstreamNormiert(t *testing.T) {
	n := 0
	for _, pfad := range []string{"/a%2Fb/", "/a%2Fb", "/a/", "/", "", "/a%2F/"} {
		n++
		up := "https://up.example" + pfad
		norm := strings.TrimSuffix(pfad, "/")
		for _, f := range []struct {
			link string
			end  string
		}{
			{"https://up.example" + norm + "/Products?$skiptoken=1", "proven"},
			{"https://up.example" + strings.ReplaceAll(norm, "%2F", "/") + "/Products?$skiptoken=1", map[bool]string{true: "link_nicht_aufloesbar", false: "proven"}[strings.Contains(norm, "%2F")]},
		} {
			w := neu(t)
			w.konfig.Dienste = []baselineimport.Dienst{{ID: dienstID, UpstreamURL: up}}
			b := w.standard()
			b.collection("Products", "", mitLink(seite(produkt(1, "A", "1")), f.link))
			b.collection("Products", "$skiptoken=1", seite(produkt(2, "A", "1")))
			mustImport(t, w, anfrage())
			if _, _, ketten := ketteVon(t, w); endeVon(ketten[0]) != f.end {
				t.Errorf("Upstream %q, Link %q: Ende %s, erwartet %s", up, f.link, endeVon(ketten[0]), f.end)
			}
		}
	}
	if n != 6 {
		t.Errorf("Menge %d, erwartet 6", n)
	}
}

// --- I10: Typen, Importe, Zustand ----------------------------------------------------------

func verboteneImporte(t *testing.T, dateien map[string]string) []string {
	t.Helper()
	erlaubt := []string{"model", "ports/inbound", "ports/outbound", "ports/support"}
	var verboten []string
	for name, text := range dateien {
		f, err := parser.ParseFile(token.NewFileSet(), name, text, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			erst, _, _ := strings.Cut(p, "/")
			switch {
			case p == "net" || p == "net/http" || p == "os" || strings.HasPrefix(p, "os/") || p == "crypto/tls" || p == "database/sql":
				verboten = append(verboten, p)
			case strings.Contains(erst, "."):
				rest, ok := strings.CutPrefix(p, "github.com/pt9912/odata-recorder/src/hexagon/")
				if !ok || !slices.Contains(erlaubt, rest) {
					verboten = append(verboten, p)
				}
			}
		}
	}
	return verboten
}

// TestZeileI10Importe: das Paket importiert keinen Adapter, keinen Dienst, kein net, net/http,
// os, os/exec, crypto/tls und database/sql; die Gegenprobe führt ein Element je Eintrag der Menge (8).
func TestZeileI10Importe(t *testing.T) {
	dateien := dienstDateien(t)
	if len(dateien) < 8 {
		t.Fatalf("nur %d Dateien gelesen", len(dateien))
	}
	if v := verboteneImporte(t, dateien); len(v) > 0 {
		t.Errorf("verbotene Importe: %v", v)
	}
	pfad := "github.com/pt9912/odata-recorder/src/"
	gegen := map[string]string{
		"a.go": "package a\nimport \"" + pfad + "adapters/driven/sqlite\"\n",
		"b.go": "package a\nimport \"" + pfad + "hexagon/services/session\"\n",
		"c.go": "package a\nimport \"net\"\n",
		"d.go": "package a\nimport \"net/http\"\n",
		"e.go": "package a\nimport \"os\"\n",
		"f.go": "package a\nimport \"os/exec\"\n",
		"g.go": "package a\nimport \"crypto/tls\"\n",
		"h.go": "package a\nimport \"database/sql\"\n",
	}
	if v := verboteneImporte(t, gegen); len(v) != 8 {
		t.Errorf("Gegenprobe meldet %d von 8: %v", len(v), v)
	}
	erlaubt := map[string]string{"i.go": "package a\nimport (\n\t\"net/url\"\n\t\"" + pfad + "hexagon/model\"\n\t\"" + pfad + "hexagon/ports/inbound\"\n)\n"}
	if v := verboteneImporte(t, erlaubt); len(v) != 0 {
		t.Errorf("Gegenprobe meldet erlaubte Importe: %v", v)
	}
}

// goroutinenIn liefert die Stellen in text, an denen Code eine Goroutine startet: die Anweisung go und die Aufrufe
// von WaitGroup.Go (jeder Aufruf einer Methode Go), time.AfterFunc und context.AfterFunc (SPEC-127 Nr. 18).
func goroutinenIn(t *testing.T, name, text string) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, text, 0)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	pakete := map[string]string{}
	for _, imp := range f.Imports {
		p, _ := strconv.Unquote(imp.Path.Value)
		lokal := p[strings.LastIndex(p, "/")+1:]
		if imp.Name != nil {
			lokal = imp.Name.Name
		}
		pakete[lokal] = p
	}
	var funde []string
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GoStmt:
			funde = append(funde, "go")
		case *ast.CallExpr:
			sel, ok := x.Fun.(*ast.SelectorExpr)
			if !ok {
				break
			}
			if sel.Sel.Name == "Go" {
				funde = append(funde, "WaitGroup.Go")
			}
			if id, ok := sel.X.(*ast.Ident); ok && sel.Sel.Name == "AfterFunc" {
				p := pakete[id.Name]
				if p == "" {
					p = id.Name
				}
				if p == "time" || p == "context" {
					funde = append(funde, p+".AfterFunc")
				}
			}
		}
		return true
	})
	return funde
}

// dienstDateien liest die Quelltexte des Pakets ohne die Tests.
func dienstDateien(t *testing.T) map[string]string {
	t.Helper()
	dateien := map[string]string{}
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range eintraege {
		if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			b, err := os.ReadFile(filepath.Join(".", e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			dateien[e.Name()] = string(b)
		}
	}
	return dateien
}

// TestZeileI10Goroutinen: der Dienst startet keine Goroutine (SPEC-127 Nr. 18): in keiner Datei des Pakets steht die
// Anweisung go oder ein Aufruf von WaitGroup.Go, time.AfterFunc oder context.AfterFunc. Die Prüfung ist
// deterministisch; die Gegenprobe fügt jede der vier Formen in jede Datei ein (Menge: Dateien mal 4).
func TestZeileI10Goroutinen(t *testing.T) {
	dateien := dienstDateien(t)
	if len(dateien) < 11 {
		t.Fatalf("nur %d Dateien gelesen, erwartet mindestens 11", len(dateien))
	}
	formen := []struct{ name, code string }{
		{"go", "func zzEingefuegt() { go func() {}() }"},
		{"WaitGroup.Go", "func zzEingefuegt() { var wg sync.WaitGroup; wg.Go(func() {}) }"},
		{"time.AfterFunc", "func zzEingefuegt() { time.AfterFunc(0, func() {}) }"},
		{"context.AfterFunc", "func zzEingefuegt() { context.AfterFunc(nil, func() {}) }"},
	}
	n := 0
	for name, text := range dateien {
		if funde := goroutinenIn(t, name, text); len(funde) > 0 {
			t.Errorf("%s startet eine Goroutine: %v", name, funde)
		}
		for _, f := range formen {
			n++
			if funde := goroutinenIn(t, name, text+"\n"+f.code+"\n"); !slices.Equal(funde, []string{f.name}) {
				t.Errorf("%s mit eingefügtem %s: Funde %v", name, f.name, funde)
			}
		}
	}
	if n != 4*len(dateien) {
		t.Errorf("Menge %d", n)
	}
	// Ein umbenannter Import von time zählt.
	if funde := goroutinenIn(t, "a.go", "package a\nimport z \"time\"\nfunc f() { z.AfterFunc(0, nil) }\n"); !slices.Equal(funde, []string{"time.AfterFunc"}) {
		t.Errorf("umbenannter Import: %v", funde)
	}
}

// TestZeileI10Typen: der Port nennt nur Typen der Domain, des Ports und von context; der Dienst
// erfüllt ihn; seine Felder sind genau die der Konfig (kein veränderlicher Zustand).
func TestZeileI10Typen(t *testing.T) {
	var _ inbound.ImportUseCase = (*baselineimport.Importer)(nil)
	ok := []string{"", "context", "github.com/pt9912/odata-recorder/src/hexagon/model", "github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"}
	it := reflect.TypeOf((*inbound.ImportUseCase)(nil)).Elem()
	if it.NumMethod() != 2 {
		t.Errorf("Methoden %d, erwartet 2", it.NumMethod())
	}
	for i := 0; i < it.NumMethod(); i++ {
		m := it.Method(i).Type
		var ts []reflect.Type
		for j := 0; j < m.NumIn(); j++ {
			ts = append(ts, m.In(j))
		}
		for j := 0; j < m.NumOut(); j++ {
			ts = append(ts, m.Out(j))
		}
		for _, x := range ts {
			if !slices.Contains(ok, x.PkgPath()) {
				t.Errorf("%s: Typ %v aus %q", it.Method(i).Name, x, x.PkgPath())
			}
		}
	}
	imp := reflect.TypeOf(baselineimport.Importer{})
	if imp.NumField() != 1 || imp.Field(0).Type != reflect.TypeOf(baselineimport.Konfig{}) {
		t.Errorf("Felder des Dienstes: %d, erwartet genau die Konfig", imp.NumField())
	}
}

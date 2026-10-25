package baselineimport_test

import (
	"errors"
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
		"URL leer":        func(k *baselineimport.Konfig) { k.Dienste = []baselineimport.Dienst{{ID: "a"}} },
		"URL ohne Schema": func(k *baselineimport.Konfig) { k.Dienste = []baselineimport.Dienst{{ID: "a", UpstreamURL: "/x"}} },
		"URL ftp":         func(k *baselineimport.Konfig) { k.Dienste = []baselineimport.Dienst{{ID: "a", UpstreamURL: "ftp://h"}} },
		"URL ohne Host": func(k *baselineimport.Konfig) {
			k.Dienste = []baselineimport.Dienst{{ID: "a", UpstreamURL: "http:///x"}}
		},
		"MaxCaptureBody 0":     func(k *baselineimport.Konfig) { k.MaxCaptureBody = 0 },
		"MaxCaptureBody -1":    func(k *baselineimport.Konfig) { k.MaxCaptureBody = -1 },
		"InvalidEncoding leer": func(k *baselineimport.Konfig) { k.Regeln.InvalidEncoding = "" },
		"InvalidEncoding x":    func(k *baselineimport.Konfig) { k.Regeln.InvalidEncoding = "x" },
		"BodyMode leer":        func(k *baselineimport.Konfig) { k.Regeln.BodyMode = "" },
		"BodyMode x":           func(k *baselineimport.Konfig) { k.Regeln.BodyMode = "x" },
	}
	if len(faelle) != 23 {
		t.Fatalf("Menge %d, erwartet 23", len(faelle))
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

package record_test

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
)

// TestNewArgumente: jede Prüfung von SPEC-121 Nr. 1 je für sich; überlappende Roots sind gültig (B1).
func TestNewArgumente(t *testing.T) {
	u := neu(t)
	faelle := map[string]func(*record.Konfig){
		"Dienste nil":             func(k *record.Konfig) { k.Dienste = nil },
		"Dienste leer":            func(k *record.Konfig) { k.Dienste = []record.Dienst{} },
		"Dienst ohne ID":          func(k *record.Konfig) { k.Dienste[1].ID = "" },
		"Dienst ohne Root":        func(k *record.Konfig) { k.Dienste[1].Root = "" },
		"Dienst ohne UpstreamURL": func(k *record.Konfig) { k.Dienste[1].UpstreamURL = "" },
		"Sessions nil":            func(k *record.Konfig) { k.Sessions = nil },
		"Exchanges nil":           func(k *record.Konfig) { k.Exchanges = nil },
		"Upstream nil":            func(k *record.Konfig) { k.Upstream = nil },
		"Pipeline nil":            func(k *record.Konfig) { k.Pipeline = nil },
		"Uhr nil":                 func(k *record.Konfig) { k.Uhr = nil },
		"IDs nil":                 func(k *record.Konfig) { k.IDs = nil },
		"Log nil":                 func(k *record.Konfig) { k.Log = nil },
		"InvalidEncoding leer":    func(k *record.Konfig) { k.Regeln.InvalidEncoding = "" },
		"InvalidEncoding x":       func(k *record.Konfig) { k.Regeln.InvalidEncoding = "x" },
		"BodyMode leer":           func(k *record.Konfig) { k.Regeln.BodyMode = "" },
		"BodyMode x":              func(k *record.Konfig) { k.Regeln.BodyMode = "x" },
	}
	for name, aendere := range faelle {
		k := u.konfig
		k.Dienste = append([]record.Dienst(nil), u.konfig.Dienste...)
		aendere(&k)
		r, err := record.New(k)
		if !errors.Is(err, model.ErrUngueltig) || r != nil {
			t.Errorf("%s: %v, Dienst %v", name, err, r != nil)
		}
	}
	gut := u.konfig
	gut.Dienste = []record.Dienst{{ID: "w", Root: "/", UpstreamURL: "http://w"}, {ID: "a", Root: "/a", UpstreamURL: "http://a"}}
	gut.Regeln.InvalidEncoding, gut.Regeln.BodyMode = model.KodierungReject, model.BodyJSONCanonical
	if r, err := record.New(gut); err != nil || r == nil {
		t.Errorf("überlappende Roots: %v", err)
	}
}

// TestAnfrageArgumente: eine Anfrage, die kein Adapter liefert, ist ErrUngueltig ohne *Fehler,
// ohne Weiterleitung, ohne Lesen der Ablage und ohne Kennung (B2).
func TestAnfrageArgumente(t *testing.T) {
	u := neu(t)
	var keinCtx context.Context
	faelle := map[string]func() error{
		"ctx nil": func() error { _, err := u.r.Record(keinCtx, anfrage("GET", "/a/x")); return err },
		"Methode leer": func() error {
			_, err := u.r.Record(context.Background(), anfrage("", "/a/x"))
			return err
		},
		"Laenge -2": func() error {
			a := mitBody(anfrage("POST", "/a/x"), "application/json", "{}")
			a.Laenge = -2
			_, err := u.r.Record(context.Background(), a)
			return err
		},
		"Body nil bei Laenge 5": func() error {
			a := anfrage("POST", "/a/x")
			a.Laenge = 5
			_, err := u.r.Record(context.Background(), a)
			return err
		},
	}
	for name, f := range faelle {
		err := f()
		var fe *inbound.Fehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &fe) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if n := len(u.upstream.gesehen()); n != 0 {
		t.Errorf("Upstream %d Aufrufe", n)
	}
	if n := u.sessions.anzahlAufrufe(); n != 0 {
		t.Errorf("Session-Fake %d Aufrufe", n)
	}
	if id := u.ids.ExchangeID(); id != kennung("exc", 1) {
		t.Errorf("Kennung bezogen: nächste ist %s", id)
	}
	for _, laenge := range []int64{-1, 0} {
		a := anfrage("POST", "/a/x")
		a.Laenge = laenge
		u.rundlauf(a)
	}
	for i, g := range u.upstream.gesehen() {
		if g.a.Body != nil || g.a.Laenge != 0 {
			t.Errorf("Gutfall %d: Body %v, Laenge %d", i, g.a.Body != nil, g.a.Laenge)
		}
	}
	if n := len(u.exchanges.angehaengt()); n != 2 {
		t.Errorf("%d Exchanges", n)
	}
}

// verboteneImporte liefert die Importe der Dateien, die weder Standardbibliothek ohne Netz, os
// und TLS noch Modell, Ports oder capture sind.
func verboteneImporte(t *testing.T, dateien map[string]string) []string {
	t.Helper()
	erlaubt := []string{"model", "ports/inbound", "ports/outbound", "ports/support", "services/record/capture"}
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
			case p == "net" || strings.HasPrefix(p, "net/") || p == "os" || strings.HasPrefix(p, "os/") || p == "crypto/tls":
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

// TestImporte: record und inbound importieren keine HTTP-Bibliothek und keinen Adapter (B31).
func TestImporte(t *testing.T) {
	dateien := map[string]string{}
	for _, dir := range []string{".", "../../ports/inbound"} {
		eintraege, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range eintraege {
			if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
				b, err := os.ReadFile(filepath.Join(dir, e.Name()))
				if err != nil {
					t.Fatal(err)
				}
				dateien[filepath.Join(dir, e.Name())] = string(b)
			}
		}
	}
	if len(dateien) < 9 {
		t.Fatalf("nur %d Dateien gelesen", len(dateien))
	}
	if v := verboteneImporte(t, dateien); len(v) > 0 {
		t.Errorf("verbotene Importe: %v", v)
	}
	gegen := map[string]string{
		"a.go": "package a\nimport \"net/http\"\n",
		"b.go": "package a\nimport \"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite\"\n",
		"c.go": "package a\nimport \"os\"\n",
	}
	if v := verboteneImporte(t, gegen); len(v) != 3 {
		t.Errorf("Gegenprobe meldet %v", v)
	}
}

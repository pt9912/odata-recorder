package maintenance_test

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
)

// TestNewArgumente: jede Prüfung von SPEC-123 Nr. 1 je für sich; MaxAge 0, 1 ns und der größte
// Wert sind gültig (W1).
func TestNewArgumente(t *testing.T) {
	u := neu(t)
	faelle := map[string]func(*maintenance.Konfig){
		"Sessions nil": func(k *maintenance.Konfig) { k.Sessions = nil },
		"Dienste nil":  func(k *maintenance.Konfig) { k.Dienste = nil },
		"Speicher nil": func(k *maintenance.Konfig) { k.Speicher = nil },
		"Blobs nil":    func(k *maintenance.Konfig) { k.Blobs = nil },
		"Uhr nil":      func(k *maintenance.Konfig) { k.Uhr = nil },
		"Log nil":      func(k *maintenance.Konfig) { k.Log = nil },
		"MaxAge -1 ns": func(k *maintenance.Konfig) { k.MaxAge = -time.Nanosecond },
	}
	for name, aendere := range faelle {
		k := u.konfig
		aendere(&k)
		w, err := maintenance.New(k)
		if !errors.Is(err, model.ErrUngueltig) || w != nil {
			t.Errorf("%s: %v, Wartung %v", name, err, w != nil)
		}
	}
	for _, d := range []time.Duration{0, time.Nanosecond, math.MaxInt64} {
		k := u.konfig
		k.MaxAge = d
		if w, err := maintenance.New(k); err != nil || w == nil {
			t.Errorf("MaxAge %v: %v", d, err)
		}
	}
}

// TestNewWaechterArgumente: jede Prüfung von NewBlobWaechter je für sich; 1 ns und der größte
// Wert sind gültige Karenzen (W1).
func TestNewWaechterArgumente(t *testing.T) {
	u := neu(t)
	faelle := map[string]func() (*maintenance.BlobWaechter, error){
		"blobs nil": func() (*maintenance.BlobWaechter, error) { return maintenance.NewBlobWaechter(nil, u.uhr, time.Minute) },
		"uhr nil": func() (*maintenance.BlobWaechter, error) {
			return maintenance.NewBlobWaechter(u.blobs, nil, time.Minute)
		},
		"karenz 0": func() (*maintenance.BlobWaechter, error) { return maintenance.NewBlobWaechter(u.blobs, u.uhr, 0) },
		"karenz -1": func() (*maintenance.BlobWaechter, error) {
			return maintenance.NewBlobWaechter(u.blobs, u.uhr, -time.Nanosecond)
		},
	}
	for name, f := range faelle {
		if w, err := f(); !errors.Is(err, model.ErrUngueltig) || w != nil {
			t.Errorf("%s: %v, Wächter %v", name, err, w != nil)
		}
	}
	for _, d := range []time.Duration{time.Nanosecond, math.MaxInt64} {
		if w, err := maintenance.NewBlobWaechter(u.blobs, u.uhr, d); err != nil || w == nil {
			t.Errorf("karenz %v: %v", d, err)
		}
	}
	if maintenance.StandardKarenz != 10*time.Minute {
		t.Errorf("StandardKarenz = %v", maintenance.StandardKarenz)
	}
}

// TestCtxNil: ctx nil ist ErrUngueltig ohne *Fehler und ohne Aufruf eines Ports (W2). Fänger eines
// Port-Aufrufs ist die Panik des Fakes: jeder Fake wertet den ctx aus, bevor er zählt, und ein nil
// als ctx löst sie aus; der Zähler gesamt() wird für sich nie rot.
func TestCtxNil(t *testing.T) {
	var keinCtx context.Context
	u := neu(t, mitMaxAge(30*tag))
	u.dienst(dienstA)
	u.sessions.fuege("s1", dienstA, t0().Add(-40*tag), model.SessionOpen)
	u.blobs.lege(r("a"))
	faelle := map[string]func() error{
		"Startup":      func() error { _, err := u.w.Startup(keinCtx); return err },
		"CleanupBlobs": func() error { _, err := u.w.CleanupBlobs(keinCtx); return err },
	}
	for name, f := range faelle {
		err := f()
		var fe *inbound.Fehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &fe) {
			t.Errorf("%s: %v", name, err)
		}
		if n := u.st.gesamt(); n != 0 {
			t.Errorf("%s: %d Aufrufe der Ports", name, n)
		}
	}
	if b, _ := u.w.Startup(keinCtx); b.Befunde == nil || b.BlobsFehlend == nil || b.BlobsBeschaedigt == nil {
		t.Errorf("Listen nil: %+v", b)
	}
}

// verboteneImporte liefert die Importe, die maintenance nicht haben darf: die Standardbibliothek
// ohne net, os und crypto/tls, Domain und Ports; kein Adapter und kein Paket unter services.
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

// TestImporte: maintenance importiert keinen Adapter, keine HTTP-Bibliothek und keinen anderen
// Dienst (W23).
func TestImporte(t *testing.T) {
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
	if len(dateien) < 4 {
		t.Fatalf("nur %d Dateien gelesen", len(dateien))
	}
	if v := verboteneImporte(t, dateien); len(v) > 0 {
		t.Errorf("verbotene Importe: %v", v)
	}
	gegen := map[string]string{
		"a.go": "package a\nimport \"net/http\"\n",
		"b.go": "package a\nimport \"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite\"\n",
		"c.go": "package a\nimport \"os\"\n",
		"d.go": "package a\nimport \"github.com/pt9912/odata-recorder/src/hexagon/services/record\"\n",
	}
	if v := verboteneImporte(t, gegen); len(v) != 4 {
		t.Errorf("Gegenprobe meldet %v", v)
	}
}

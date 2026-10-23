package session_test

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
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// TestNewArgumente: jede Prüfung von SPEC-124 Nr. 1 je für sich; die Konfig des Tests setzt genau
// die sechs Felder und keinen Modus (V1).
func TestNewArgumente(t *testing.T) {
	u := neu(t)
	faelle := map[string]func(*session.Konfig){
		"Sessions nil": func(k *session.Konfig) { k.Sessions = nil },
		"Dienste nil":  func(k *session.Konfig) { k.Dienste = nil },
		"Wartung nil":  func(k *session.Konfig) { k.Wartung = nil },
		"Uhr nil":      func(k *session.Konfig) { k.Uhr = nil },
		"IDs nil":      func(k *session.Konfig) { k.IDs = nil },
		"Log nil":      func(k *session.Konfig) { k.Log = nil },
	}
	for name, aendere := range faelle {
		k := u.konfig
		aendere(&k)
		v, err := session.New(k)
		if !errors.Is(err, model.ErrUngueltig) || v != nil || !ohneCode(err) {
			t.Errorf("%s: %v, Dienst %v", name, err, v != nil)
		}
	}
	if v, err := session.New(u.konfig); err != nil || v == nil {
		t.Errorf("alle sechs Felder gesetzt: %v", err)
	}
}

// TestCtxNil: ein nil-ctx ist bei jeder Operation ErrUngueltig ohne Code, ohne Aufruf eines Ports,
// ohne Now und SessionID (V2). Ein Port-Aufruf panict auf dem nil-ctx und macht den Fall rot.
func TestCtxNil(t *testing.T) {
	var keinCtx context.Context
	faelle := map[string]func(*welt) error{
		"Create mit leerer id": func(u *welt) error { _, err := u.v.Create(keinCtx, dienstA, ""); return err },
		"Create mit id":        func(u *welt) error { _, err := u.v.Create(keinCtx, dienstA, "x1"); return err },
		"OpenInitialSession":   func(u *welt) error { _, err := u.v.OpenInitialSession(keinCtx, dienstA); return err },
		"Get":                  func(u *welt) error { _, err := u.v.Get(keinCtx, "s1"); return err },
		"List":                 func(u *welt) error { _, err := u.v.List(keinCtx); return err },
		"Delete":               func(u *welt) error { _, err := u.v.Delete(keinCtx, "s1"); return err },
	}
	for name, f := range faelle {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: Panik %v (ein Port wurde gerufen)", name, r)
				}
			}()
			if err := f(u); !errors.Is(err, model.ErrUngueltig) || !ohneCode(err) {
				t.Errorf("%s: %v", name, err)
			}
		}()
		if n := u.st.gesamt(); n != 0 {
			t.Errorf("%s: %d Aufrufe von Ports, Now oder SessionID", name, n)
		}
		if id := u.folgeID(); id != sesID(1) {
			t.Errorf("%s: nächste Kennung %s, erwartet %s", name, id, sesID(1))
		}
		if _, ok := u.sessions.hat("s1"); !ok {
			t.Errorf("%s: die Session wurde gelöscht", name)
		}
	}
}

// portsDerFakes nennt jede Operation der Fakes unter dem Namen, unter dem sie gezählt wird.
func portsDerFakes() []string {
	return []string{
		"Dienste.List", "Dienste.Get", "Dienste.Create", "Dienste.SetMetadata",
		"Sessions.Create", "Sessions.Get", "SetStatus", "Sessions.List", "FindOpen", "Sessions.Delete",
		"CleanupBlobs", "Startup",
	}
}

// TestDienstDientAllein: jede Operation ruft genau die Ports, die ihr Vertrag nennt, und sonst
// keine Operation eines Fakes; Dienste dient allein List und nur der Auflistung (V22). Im Gutfall
// zählt jede Operation genau, wenn alle Ports scheitern, höchstens.
func TestDienstDientAllein(t *testing.T) {
	scheitern := func(u *welt) {
		for _, p := range []string{"Sessions.Create", "Sessions.Get", "Sessions.List", "Sessions.Delete", "Dienste.List"} {
			u.st.scheitere(p, 0, errDisk)
		}
	}
	for i, op := range operationen() {
		erlaubt := map[string]int{}
		for _, s := range portStellen()[i] {
			erlaubt[s.port] = s.n
		}
		for _, fall := range []struct {
			name  string
			setup func(*welt)
		}{{"Gutfall", nil}, {"Ports scheitern", scheitern}} {
			u := neu(t)
			u.bestand()
			if fall.setup != nil {
				fall.setup(u)
			}
			_, err := op.f(context.Background(), u)
			if (err != nil) != (fall.setup != nil) {
				t.Errorf("%s, %s: %v", op.name, fall.name, err)
			}
			for _, p := range portsDerFakes() {
				got := u.st.anzahl(p)
				if got > erlaubt[p] || (fall.setup == nil && got != erlaubt[p]) {
					t.Errorf("%s, %s: %s %d-mal, erlaubt %d", op.name, fall.name, p, got, erlaubt[p])
				}
			}
		}
	}
}

// verboteneImporte liefert die Importe, die session nicht haben darf: die Standardbibliothek
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

// TestImporte: session importiert keinen Adapter, keine HTTP-Bibliothek und keinen anderen
// Dienst, auch nicht maintenance (V20).
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
	if len(dateien) < 3 {
		t.Fatalf("nur %d Dateien gelesen", len(dateien))
	}
	if v := verboteneImporte(t, dateien); len(v) > 0 {
		t.Errorf("verbotene Importe: %v", v)
	}
	// Gegenprobe: ein Fall je Element des Matchers (net genau, net/ als Präfix, os genau, os/ als
	// Präfix, crypto/tls, Pfad mit Punkt außerhalb der erlaubten Ports) und die erlaubten Importe.
	pfad := "github.com/pt9912/odata-recorder/src/"
	gegen := map[string]string{
		"a.go": "package a\nimport \"net/http\"\n",
		"b.go": "package a\nimport \"" + pfad + "adapters/driven/sqlite\"\n",
		"c.go": "package a\nimport \"os\"\n",
		"d.go": "package a\nimport \"" + pfad + "hexagon/services/maintenance\"\n",
		"e.go": "package a\nimport \"net\"\n",
		"f.go": "package a\nimport \"crypto/tls\"\n",
		"g.go": "package a\nimport \"os/exec\"\n",
	}
	soll := []string{"crypto/tls", "net", "net/http", "os", "os/exec", pfad + "adapters/driven/sqlite", pfad + "hexagon/services/maintenance"}
	v := verboteneImporte(t, gegen)
	slices.Sort(v)
	slices.Sort(soll)
	if !slices.Equal(v, soll) {
		t.Errorf("Gegenprobe meldet %v, erwartet %v", v, soll)
	}
	erlaubt := map[string]string{"h.go": "package a\nimport (\n\t\"context\"\n\t\"crypto/sha256\"\n\t\"log/slog\"\n\t\"" + pfad +
		"hexagon/model\"\n\t\"" + pfad + "hexagon/ports/inbound\"\n\t\"" + pfad + "hexagon/ports/outbound\"\n\t\"" + pfad + "hexagon/ports/support\"\n)\n"}
	if v := verboteneImporte(t, erlaubt); len(v) != 0 {
		t.Errorf("Gegenprobe meldet erlaubte Importe: %v", v)
	}
}

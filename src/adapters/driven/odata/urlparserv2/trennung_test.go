package urlparserv2_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const odataPraefix = "github.com/pt9912/odata-recorder/src/adapters/driven/odata/"

// verboteneImporte liest die Importe jeder .go-Datei in dir und nennt die Pfade eines Pakets unter odata/, dessen Name nicht auf
// `v2` endet (ADR-0011: kein V2-Paket importiert ein V4-Paket); ein eigenes Unterpaket von urlparserv2 ist erlaubt.
func verboteneImporte(t *testing.T, dir string) (dateien int, verboten []string) {
	t.Helper()
	namen, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range namen {
		f, err := parser.ParseFile(token.NewFileSet(), n, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		dateien++
		for _, im := range f.Imports {
			pfad, _ := strconv.Unquote(im.Path.Value)
			if rest, ok := strings.CutPrefix(pfad, odataPraefix); ok {
				if erstes, _, _ := strings.Cut(rest, "/"); !strings.HasSuffix(erstes, "v2") {
					verboten = append(verboten, filepath.Base(n)+": "+pfad)
				}
			}
		}
	}
	return dateien, verboten
}

// TestKeinV4Import (Zeile 16): kein Import eines Pakets unter odata/ ohne den Namen v2 aus urlparserv2, weder im Produktcode noch
// in den Tests; vier Gegenproben im Wegwerfbaum fangen den Import von urlparser, csdl und jsoncodec, nicht den eines eigenen
// Unterpakets.
func TestKeinV4Import(t *testing.T) {
	dateien, verboten := verboteneImporte(t, ".")
	if dateien < 7 || len(verboten) != 0 {
		t.Errorf("%d Dateien gelesen (erwartet mindestens 7), verbotene Importe %v", dateien, verboten)
	}
	for _, g := range []struct {
		pfad     string
		verboten int
	}{{"urlparser", 1}, {"csdl", 1}, {"jsoncodec", 1}, {"urlparserv2/hilfe", 0}} {
		dir := t.TempDir()
		quelle := "package x\n\nimport _ \"" + odataPraefix + g.pfad + "\"\n"
		if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(quelle), 0o600); err != nil {
			t.Fatal(err)
		}
		if n, v := verboteneImporte(t, dir); n != 1 || len(v) != g.verboten {
			t.Errorf("Gegenprobe %s: %d Dateien, verbotene Importe %v, erwartet %d", g.pfad, n, v, g.verboten)
		}
	}
}

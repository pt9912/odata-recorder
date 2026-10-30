package urlparser_test

import (
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestImporte (Zeile 59): der Parser importiert nur das Modell, die Ports und die Standardbibliothek (`ARC-011`), kein Paket
// eines anderen Adapters. `make a-check` fängt den Import einer Schicht, die der Adapter nicht nennen darf (Dienste), nicht
// aber den eines Nachbar-Adapters derselben Schicht; diesen Fall fängt der Test.
func TestImporte(t *testing.T) {
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	const modul = "github.com/pt9912/odata-recorder/src/"
	gelesen := 0
	for _, e := range eintraege {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		gelesen++
		for _, imp := range f.Imports {
			pfad, _ := strconv.Unquote(imp.Path.Value)
			stdlib := !strings.Contains(strings.SplitN(pfad, "/", 2)[0], ".")
			erlaubt := strings.HasPrefix(pfad, modul+"hexagon/model") || strings.HasPrefix(pfad, modul+"hexagon/ports/")
			if !stdlib && !erlaubt {
				t.Errorf("%s importiert %s", e.Name(), pfad)
			}
		}
	}
	if gelesen < 6 {
		t.Fatalf("nur %d Quelltexte gelesen", gelesen)
	}
}

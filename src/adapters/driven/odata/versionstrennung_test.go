package odata_test

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdlv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

const (
	edmxV2 = "http://schemas.microsoft.com/ado/2007/06/edmx"
	edmxV4 = "http://docs.oasis-open.org/odata/ns/edmx"
	edmV2  = "http://schemas.microsoft.com/ado/2008/09/edm"
	edmV4  = "http://docs.oasis-open.org/odata/ns/edm"
	meta   = "http://schemas.microsoft.com/ado/2007/08/dataservices/metadata"
)

// dokument ist ein Dokument mit Wurzel name im Namensraum ns, Version v und dem Inhalt inhalt.
func dokument(ns, name, v, inhalt string) string {
	return `<edmx:` + name + ` xmlns:edmx="` + ns + `" Version="` + v + `">` + inhalt + `</edmx:` + name + `>`
}

// TestVersionstrennungDokumente hält J2: ein V4-Dokument am V2-Leser und ein V2-Dokument am V4-Leser sind `version`,
// ohne Modell; ein Dokument der eigenen Version wird gelesen. Schleife über die Menge der Dokumente mal beide Leser.
func TestVersionstrennungDokumente(t *testing.T) {
	ds2 := `<edmx:DataServices><Schema xmlns="` + edmV2 + `" Namespace="Ns"><EntityType Name="T"><Key><PropertyRef Name="Id"/></Key><Property Name="Id" Type="Edm.Int32"/></EntityType></Schema></edmx:DataServices>`
	ds4 := `<edmx:DataServices><Schema xmlns="` + edmV4 + `" Namespace="Ns"><EntityType Name="T"><Key><PropertyRef Name="Id"/></Key><Property Name="Id" Type="Edm.Int32"/></EntityType></Schema></edmx:DataServices>`
	ok, ver, wur := model.MetadatenUrsache(""), model.UrsacheVersion, model.UrsacheWurzel
	for _, d := range []struct {
		name   string
		doc    string
		v2, v4 model.MetadatenUrsache
	}{
		{"Referenzdokument V2", dokument(edmxV2, "Edmx", "1.0", ds2), ok, ver},
		{"Referenzdokument V4", dokument(edmxV4, "Edmx", "4.0", ds4), ver, ok},
		{"minimale Wurzel V2", dokument(edmxV2, "Edmx", "1.0", `<edmx:DataServices/>`), ok, ver},
		{"minimale Wurzel V4", dokument(edmxV4, "Edmx", "4.0", `<edmx:DataServices/>`), ver, ok},
		{"V4-Wurzel mit Version 1.0", dokument(edmxV4, "Edmx", "1.0", `<edmx:DataServices/>`), ver, ver},
		{"V2-Wurzel mit Version 4.0", dokument(edmxV2, "Edmx", "4.0", `<edmx:DataServices/>`), ver, ver},
		{"V2-Wurzel mit Version 3.0", dokument(edmxV2, "Edmx", "3.0", `<edmx:DataServices/>`), ver, ver},
		{"V2-Wurzel mit DataServiceVersion 3.0", dokument(edmxV2, "Edmx", "1.0", `<edmx:DataServices xmlns:m="`+meta+`" m:DataServiceVersion="3.0"/>`), ver, ver},
		{"V2-Wurzel mit DataServiceVersion 2.0", dokument(edmxV2, "Edmx", "1.0", `<edmx:DataServices xmlns:m="`+meta+`" m:DataServiceVersion="2.0"/>`), ok, ver},
		{"Wurzel Foo im Namensraum von V2", dokument(edmxV2, "Foo", "1.0", `<edmx:DataServices/>`), wur, wur},
		{"Wurzel Foo im Namensraum von V4", dokument(edmxV4, "Foo", "4.0", `<edmx:DataServices/>`), wur, wur},
	} {
		for _, l := range []struct {
			name  string
			leser outbound.MetadataCodec
			soll  model.MetadatenUrsache
			ver   string
		}{{"V2", csdlv2.New(), d.v2, "2.0"}, {"V4", csdl.New(), d.v4, "4.0"}} {
			m, b, err := l.leser.ParseMetadata(context.Background(), strings.NewReader(d.doc), 1<<20)
			var mf *model.MetadatenFehler
			switch {
			case l.soll == "":
				if err != nil || m.Version() != l.ver || b.Funde == nil {
					t.Errorf("%s am Leser %s: %v, Version %q", d.name, l.name, err, m.Version())
				}
			case !errors.As(err, &mf) || mf.Ursache != l.soll || mf.Ort != "" || m.Version() != "" || len(m.EntityTypes()) != 0 || b.Funde != nil:
				t.Errorf("%s am Leser %s: Fehler %v, Version %q, erwartet %s ohne Modell", d.name, l.name, err, m.Version(), l.soll)
			}
		}
	}
}

// importeVerstoesse liefert die Verstöße gegen die Trennung der Versionen unter root: ein Paket der Version 2 (Name auf
// `v2`) importiert ein Paket der Version 4 (jedes andere außer den versionsneutralen) oder umgekehrt. prefix ist der
// Importpfad von root mit abschließendem Schrägstrich.
func importeVerstoesse(root, prefix string, neutral []string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(pfad string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
			return err
		}
		von := paket(root, filepath.Dir(pfad))
		if von == "" || slices.Contains(neutral, von) {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), pfad, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			s, _ := strconv.Unquote(imp.Path.Value)
			rest, ok := strings.CutPrefix(s, prefix)
			if !ok {
				continue
			}
			nach, _, _ := strings.Cut(rest, "/")
			if !slices.Contains(neutral, nach) && strings.HasSuffix(von, "v2") != strings.HasSuffix(nach, "v2") {
				out = append(out, pfad+" importiert "+s)
			}
		}
		return nil
	})
	return out, err
}

// paket ist der Name des Pakets unter root, zu dem das Verzeichnis dir gehört (das erste Glied unter root).
func paket(root, dir string) string {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." {
		return ""
	}
	erst, _, _ := strings.Cut(filepath.ToSlash(rel), "/")
	return erst
}

// TestKeinPaketImportiertEineAndereVersion hält J3: kein V2-Paket importiert ein V4-Paket und umgekehrt; Gemeinsames
// liegt in einem Paket ohne Versionsnamen, das hier als neutral eingetragen wird (zurzeit keines).
func TestKeinPaketImportiertEineAndereVersion(t *testing.T) {
	const prefix = "github.com/pt9912/odata-recorder/src/adapters/driven/odata/"
	verstoesse, err := importeVerstoesse(".", prefix, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(verstoesse) != 0 {
		t.Errorf("Verstöße gegen die Trennung der Versionen: %v", verstoesse)
	}
	pakete := 0
	entries, _ := os.ReadDir(".")
	for _, e := range entries {
		if e.IsDir() {
			pakete++
		}
	}
	if pakete < 4 {
		t.Errorf("%d Pakete unter dem Verzeichnis, erwartet mindestens 4 (csdl, csdlv2, jsoncodec, urlparser)", pakete)
	}
}

// TestImportScannerErkenntVerstoss ist die Gegenprobe zu J3: ein Wegwerfbaum mit einem Import in jede Richtung ist rot,
// einer mit einem versionsneutralen Paket und den eigenen Paketen grün.
func TestImportScannerErkenntVerstoss(t *testing.T) {
	const prefix = "x/odata/"
	schreibe := func(root, paket, imp string) {
		dir := filepath.Join(root, paket)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		src := "package " + paket + "\n"
		if imp != "" {
			src += "import _ \"" + prefix + imp + "\"\n"
		}
		if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []struct {
		name   string
		pakete map[string]string
		soll   int
	}{
		{"V2 importiert V4", map[string]string{"csdl": "", "csdlv2": "csdl"}, 1},
		{"V4 importiert V2", map[string]string{"csdl": "csdlv2", "csdlv2": ""}, 1},
		{"V2 importiert ein Unterpaket von V4", map[string]string{"csdl": "", "csdlv2": "csdl/innen"}, 1},
		{"beide Richtungen", map[string]string{"csdl": "csdlv2", "csdlv2": "csdl"}, 2},
		{"jedes Paket importiert sein eigenes", map[string]string{"csdl": "csdl/innen", "csdlv2": "csdlv2/innen"}, 0},
		{"beide importieren ein neutrales Paket", map[string]string{"csdl": "rahmen", "csdlv2": "rahmen", "rahmen": ""}, 0},
		{"Pakete ohne Importe", map[string]string{"csdl": "", "csdlv2": ""}, 0},
	} {
		root := t.TempDir()
		for p, imp := range f.pakete {
			schreibe(root, p, imp)
		}
		got, err := importeVerstoesse(root, prefix, []string{"rahmen"})
		if err != nil || len(got) != f.soll {
			t.Errorf("%s: %d Verstöße %v (%v), erwartet %d", f.name, len(got), got, err, f.soll)
		}
	}
	root := t.TempDir()
	schreibe(root, "csdl", "")
	if err := os.WriteFile(filepath.Join(root, "csdl", "b.go"), []byte("package csdl\nimport _ \"fmt\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := importeVerstoesse(root, prefix, nil); err != nil || len(got) != 0 {
		t.Errorf("Standardbibliothek: %v %v", got, err)
	}
}

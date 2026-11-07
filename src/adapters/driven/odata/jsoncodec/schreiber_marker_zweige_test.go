package jsoncodec_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// mkEntity ist die Entity der Stelle Nichtschlüssel mit dem Text in der Property NS.
func mkEntity(text string) model.Entity {
	return mkProbe{"Ns.Alle", "'k'", []mkFeld{mkStr("S", "k", ""), mkStr("NS", text, "")}}.entity()
}

// mkOperation schreibt eine Entity (Entity oder Collection mit einem Element) mit dem Text und einer Auswahl.
type mkOperation struct {
	name     string
	schreibe func(ctx context.Context, dst io.Writer, e model.Entity, auswahl []string) (model.Darstellung, error)
}

func mkOperationen() []mkOperation {
	return []mkOperation{
		{"Entity", func(ctx context.Context, dst io.Writer, e model.Entity, auswahl []string) (model.Darstellung, error) {
			return schreiber().SchreibeEntity(ctx, dst, testModell(), model.EntityAntwort{Wurzel: wurzel, Set: "Alles", Entity: e, Auswahl: auswahl})
		}},
		{"Collection", func(ctx context.Context, dst io.Writer, e model.Entity, auswahl []string) (model.Darstellung, error) {
			return schreiber().SchreibeCollection(ctx, dst, testModell(), model.CollectionAntwort{Wurzel: wurzel, Set: "Alles", Entities: []model.Entity{e}, Auswahl: auswahl})
		}},
	}
}

// TestMarkerAndereArten hält Z05 (M4, 16 Elemente): der Text `[REDACTED]` bei jeder anderen Art und an jedem anderen Typ ist ErrUngueltig.
func TestMarkerAndereArten(t *testing.T) {
	props := func(name string, art model.WertArt) []model.PropertyWert {
		return []model.PropertyWert{pw("S", model.ArtString, "k"), pw(name, art, mkMarker)}
	}
	fremde := []struct {
		prop string
		art  model.WertArt
	}{
		{"B", model.ArtBoolean}, {"I32", model.ArtInt32}, {"I64", model.ArtInt64}, {"Dec", model.ArtDecimal},
		{"G", model.ArtGUID}, {"D", model.ArtDate}, {"T", model.ArtDateTimeOffset},
	}
	n := 0
	fall := func(was string, ps []model.PropertyWert) {
		entityScheitert(t, was, model.EntityAntwort{Wurzel: wurzel, Set: "Alles", Entity: model.Entity{Typ: "Ns.Alle", Schluessel: "'k'", Properties: ps}})
		n++
	}
	for _, f := range fremde {
		fall("Art "+string(f.art)+" an "+f.prop, props(f.prop, f.art))
		fall("Art string an "+f.prop, props(f.prop, model.ArtString))
	}
	fall("Art redigiert", props("NS", model.ArtRedigiert))
	fall("Art null mit Text", props("NS", model.ArtNull))
	if n != 16 {
		t.Errorf("Elemente %d, will 16", n)
	}
}

// TestMarkerUTF8 hält Z06 (M5, 16 Elemente): ungültiges UTF-8 vor oder hinter dem Marker ist ErrUngueltig.
func TestMarkerUTF8(t *testing.T) {
	n := 0
	for _, op := range mkOperationen() {
		for _, b := range []string{"\xff", "\xc0\x80", "\xed\xa0\x80", "\xe2\x82"} {
			for _, text := range []string{b + mkMarker, mkMarker + b} {
				scheitert(t, fmt.Sprintf("%s %q", op.name, text), func(ctx context.Context, dst io.Writer) (model.Darstellung, error) {
					return op.schreibe(ctx, dst, mkEntity(text), nil)
				})
				n++
			}
		}
	}
	if n != 16 {
		t.Errorf("Elemente %d, will 16", n)
	}
}

// TestMarkerDst hält Z07 (M6, Fälle 1 bis 3): ein Fehler von dst bleibt unverändert, ein kurzer Write ist io.ErrShortWrite; ein Write.
func TestMarkerDst(t *testing.T) {
	n := 0
	errDst := errors.New("dst kaputt")
	for _, op := range mkOperationen() {
		z0 := neuesZiel()
		if _, err := op.schreibe(context.Background(), z0, mkEntity(mkMarker), nil); err != nil {
			t.Fatalf("%s: %v", op.name, err)
		}
		laenge := len(z0.body())
		for _, f := range []struct {
			name string
			z    *ziel
			want error
		}{
			{"Write meldet einen Fehler", &ziel{log: &[]string{}, kurz: -1, err: errDst}, errDst},
			{"Write meldet Länge-1", &ziel{log: &[]string{}, kurz: -2}, io.ErrShortWrite},
			{"Write meldet Länge-5", &ziel{log: &[]string{}, kurz: laenge - 5}, io.ErrShortWrite},
		} {
			d, err := op.schreibe(context.Background(), f.z, mkEntity(mkMarker), nil)
			if err != f.want || d != (model.Darstellung{}) {
				t.Errorf("%s %s: Fehler %v, Darstellung %+v", op.name, f.name, err, d)
			}
			f.z.nurEinWrite(t, op.name+" "+f.name)
			n++
		}
	}
	if n != 6 {
		t.Errorf("Elemente %d, will 6", n)
	}
}

// TestMarkerKontext hält Z08 (M6, Fälle 4 und 5): ein beendetes ctx ist der Fehler des Kontexts und lässt dst leer; ein ungültiges
// Argument geht dem Kontext vor.
func TestMarkerKontext(t *testing.T) {
	n := 0
	for _, op := range mkOperationen() {
		z := neuesZiel()
		k := neuerKontext(z, context.Canceled)
		d, err := op.schreibe(k, z, mkEntity(mkMarker), nil)
		if err != context.Canceled || d != (model.Darstellung{}) || k.errN != 1 {
			t.Errorf("%s ctx beendet: Fehler %v, Darstellung %+v, Err %d mal", op.name, err, d, k.errN)
		}
		z.leer(t, op.name+" ctx beendet")
		z = neuesZiel()
		k = neuerKontext(z, context.Canceled)
		d, err = op.schreibe(k, z, mkEntity(mkMarker), []string{"Unbekannt"})
		istUngueltig(t, err, op.name+" ungültiges Argument")
		if d != (model.Darstellung{}) || k.errN != 0 {
			t.Errorf("%s ungültiges Argument: Darstellung %+v, Err %d mal", op.name, d, k.errN)
		}
		z.leer(t, op.name+" ungültiges Argument")
		n += 2
	}
	if n != 4 {
		t.Errorf("Elemente %d, will 4", n)
	}
}

// TestMarkerArgumente hält Z09 (M7, 286 Elemente): der Schreiber verändert kein Argument und liefert bei gleichen Argumenten gleiche Bytes.
func TestMarkerArgumente(t *testing.T) {
	n := 0
	for i, text := range mkTexte() {
		for _, op := range mkOperationen() {
			erste, vorher := mkEntity(text), mkEntity(text)
			z1, z2 := neuesZiel(), neuesZiel()
			_, err1 := op.schreibe(context.Background(), z1, erste, nil)
			_, err2 := op.schreibe(context.Background(), z2, mkEntity(text), nil)
			if err1 != nil || err2 != nil || z1.body() != z2.body() {
				t.Errorf("Text %d %s: Fehler %v, %v, gleiche Bytes: %v", i, op.name, err1, err2, z1.body() == z2.body())
			}
			if !reflect.DeepEqual(erste, vorher) {
				t.Errorf("Text %d %s: das Argument hat sich geändert", i, op.name)
			}
			n++
		}
	}
	if n != 286 {
		t.Errorf("Elemente %d, will 286", n)
	}
}

// mkZweig ist eine Verzweigungsstelle des Schreibpfades eines Edm.String-Textes (Zweig-Menge W) und die Tests, die sie ausführen.
type mkZweig struct {
	nr       string
	funktion string
	stelle   string
	faenger  []string
}

// mkZweige ist die Zweig-Menge W: die Zeilen in der Reihenfolge des Quelltextes je Funktion.
func mkZweige() []mkZweig {
	z := []string{"TestMarkerBytes", "TestMarkerKeinRohtext", "TestMarkerNachbarn", "TestMarkerRundlauf"}
	return []mkZweig{
		{"W01", "ausgabe.wert", "case model.ArtNull", []string{"TestRundlauf"}},
		{"W02", "ausgabe.wert", "case model.ArtString", z},
		{"W03", "ausgabe.wert", "case model.ArtBoolean, model.ArtInt32", []string{"TestRundlauf"}},
		{"W04", "ausgabe.wert", "case model.ArtInt64, model.ArtDecimal", []string{"TestRundlauf", "TestZahlform"}},
		{"W05", "ausgabe.wert", "default", []string{"TestRundlauf", "TestMarkerAndereArten"}},
		{"W06", "ausgabe.stringWert", "if s == markerText", []string{"TestMarkerBytes", "TestMarkerNachbarn"}},
		{"W07", "ausgabe.text", "if !utf8.ValidString(s)", []string{"TestUngueltigesUTF8AnVierStellen", "TestMarkerUTF8"}},
		{"W08", "ausgabe.text", "range len(s)", []string{"TestEscapes", "TestMarkerNachbarn"}},
		{"W09", "ausgabe.text", `case c == '"' || c == '\\'`, []string{"TestEscapes", "TestMarkerNachbarn"}},
		{"W10", "ausgabe.text", "case c < 0x20", []string{"TestEscapes", "TestMarkerNachbarn"}},
		{"W11", "ausgabe.text", "default", []string{"TestMarkerNachbarn"}},
		{"W12", "pruefeWert", "if w.Art == model.ArtNull", []string{"TestMarkerAndereArten"}},
		{"W13", "pruefeWert", `if !p.Nullable || w.Text != ""`, []string{"TestMarkerAndereArten"}},
		{"W14", "pruefeWert", "if !bekannt || edm != p.EdmType", []string{"TestMarkerAndereArten", "TestArtGegenTyp"}},
		{"W15", "pruefeWert", "if !kanon(w.Text)", []string{"TestMarkerAndereArten", "TestMarkerUTF8", "TestTextKanonisch"}},
		{"W16", "ausgeben", "if err := ctx.Err(); err != nil", []string{"TestMarkerKontext"}},
		{"W17", "ausgeben", "if err != nil", []string{"TestMarkerDst"}},
		{"W18", "ausgeben", "if n < len(body)", []string{"TestMarkerDst"}},
	}
}

// mkQuelle parst die Dateien des Pakets, deren Namen auf Ende (`.go` ohne `_test`, bzw. `_test.go`) passen.
func mkQuelle(t *testing.T, test bool) (*token.FileSet, []*ast.File) {
	t.Helper()
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Paketverzeichnis: %v", err)
	}
	fset := token.NewFileSet()
	var dateien []*ast.File
	for _, e := range eintraege {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") != test {
			continue
		}
		f, err := parser.ParseFile(fset, e.Name(), nil, 0)
		if err != nil {
			t.Fatalf("%s: %v", e.Name(), err)
		}
		dateien = append(dateien, f)
	}
	return fset, dateien
}

// mkAusdruck ist der Quelltext eines Ausdrucks oder einer Anweisung.
func mkAusdruck(fset *token.FileSet, n any) string {
	var b bytes.Buffer
	if err := printer.Fprint(&b, fset, n); err != nil {
		return "?"
	}
	return b.String()
}

// mkFunktionsname ist `Empfänger.Name` bei einer Methode, sonst `Name`.
func mkFunktionsname(d *ast.FuncDecl) string {
	if d.Recv == nil || len(d.Recv.List) != 1 {
		return d.Name.Name
	}
	typ := d.Recv.List[0].Type
	if s, ok := typ.(*ast.StarExpr); ok {
		typ = s.X
	}
	if id, ok := typ.(*ast.Ident); ok {
		return id.Name + "." + d.Name.Name
	}
	return d.Name.Name
}

// mkVerzweigungen beschreibt die Knoten IfStmt, CaseClause, CommClause, ForStmt und RangeStmt einer Funktion in der Reihenfolge des Quelltextes.
func mkVerzweigungen(fset *token.FileSet, d *ast.FuncDecl) []string {
	var out []string
	ast.Inspect(d.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt:
			s := "if "
			if x.Init != nil {
				s += mkAusdruck(fset, x.Init) + "; "
			}
			out = append(out, s+mkAusdruck(fset, x.Cond))
		case *ast.CaseClause:
			if x.List == nil {
				out = append(out, "default")
				break
			}
			var es []string
			for _, e := range x.List {
				es = append(es, mkAusdruck(fset, e))
			}
			out = append(out, "case "+strings.Join(es, ", "))
		case *ast.CommClause:
			out = append(out, "comm")
		case *ast.ForStmt:
			out = append(out, "for")
		case *ast.RangeStmt:
			out = append(out, "range "+mkAusdruck(fset, x.X))
		}
		return true
	})
	return out
}

// TestMarkerZweige hält Z10: die Verzweigungsstellen der Funktionen des Schreibpfades eines Edm.String-Textes sind genau die 18 Zeilen
// von W, in der Reihenfolge des Quelltextes, und jede Zeile nennt Tests, die es im Paket gibt.
func TestMarkerZweige(t *testing.T) {
	fset, dateien := mkQuelle(t, false)
	gefunden := map[string][]string{}
	for _, f := range dateien {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
				gefunden[mkFunktionsname(fd)] = mkVerzweigungen(fset, fd)
			}
		}
	}
	zweige := mkZweige()
	soll := map[string][]string{}
	for _, z := range zweige {
		soll[z.funktion] = append(soll[z.funktion], z.stelle)
	}
	n := 0
	for fn, stellen := range soll {
		got, ok := gefunden[fn]
		if !ok || !slices.Equal(got, stellen) {
			t.Errorf("Funktion %s: gefunden %q (vorhanden: %v), will %q", fn, got, ok, stellen)
		}
		n += len(got)
	}
	if len(zweige) != 18 || n != 18 || len(soll) != 5 {
		t.Errorf("Zeilen %d, Knoten %d, Funktionen %d, will 18, 18 und 5", len(zweige), n, len(soll))
	}
	_, tests := mkQuelle(t, true)
	vorhanden := map[string]bool{}
	for _, f := range tests {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
				vorhanden[fd.Name.Name] = true
			}
		}
	}
	for _, z := range zweige {
		if len(z.faenger) == 0 {
			t.Errorf("%s: kein Fänger", z.nr)
		}
		for _, f := range z.faenger {
			if !vorhanden[f] {
				t.Errorf("%s: Test %s fehlt im Paket", z.nr, f)
			}
		}
	}
}

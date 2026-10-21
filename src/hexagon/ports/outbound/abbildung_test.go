package outbound_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

const (
	gSonst = outbound.GegenstandSonstiges
	gSess  = outbound.GegenstandSession
	gWork  = outbound.GegenstandWorkspace
	gSnap  = outbound.GegenstandSnapshot
	gName  = outbound.GegenstandSnapshotname
	gImp   = outbound.GegenstandImport
	gExch  = outbound.GegenstandExchange
	gDien  = outbound.GegenstandDienst
)

func alleGegenstaende() []outbound.Gegenstand {
	return []outbound.Gegenstand{gSonst, gSess, gWork, gSnap, gName, gImp, gExch, gDien}
}

func jeder(c model.Code) map[outbound.Gegenstand]model.Code {
	m := map[outbound.Gegenstand]model.Code{}
	for _, g := range alleGegenstaende() {
		m[g] = c
	}
	return m
}

func mit(basis map[outbound.Gegenstand]model.Code, je map[outbound.Gegenstand]model.Code) map[outbound.Gegenstand]model.Code {
	for g := range basis {
		basis[g] = model.CodeE9001
	}
	for g, c := range je {
		basis[g] = c
	}
	return basis
}

// TestAbbildungTabelle hat eine Zeile je Paar Fehler × Gegenstand aus SPEC-109 Nr. 5.
func TestAbbildungTabelle(t *testing.T) {
	tests := []struct {
		fehler error
		soll   map[outbound.Gegenstand]model.Code
	}{
		{outbound.ErrBlobFehlt, jeder(model.CodeE3001)},
		{outbound.ErrBlobBeschaedigt, jeder(model.CodeE3001)},
		{outbound.ErrBeschaedigt, jeder(model.CodeE3001)},
		{outbound.ErrPfad, jeder(model.CodeE2001)},
		{outbound.ErrZuGross, jeder(model.CodeE2401)},
		{outbound.ErrUpstreamNichtErreichbar, jeder(model.CodeE4001)},
		{outbound.ErrUpstreamZeit, jeder(model.CodeE4001)},
		{outbound.ErrFremderService, jeder(model.CodeE5004)},
		{outbound.ErrDienstUnbekannt, jeder(model.CodeE5005)},
		{outbound.ErrFalscheArt, jeder(model.CodeE7008)},
		{outbound.ErrReferenziert, jeder(model.CodeE7004)},
		{outbound.ErrSessionGeschlossen, jeder(model.CodeE5009)},
		{outbound.ErrUngueltigerUebergang, jeder(model.CodeE7011)},
		{outbound.ErrNichtGefunden, mit(jeder(0), map[outbound.Gegenstand]model.Code{
			gSess: model.CodeE7003, gWork: model.CodeE7001, gSnap: model.CodeE7002,
			gImp: model.CodeE7005, gExch: model.CodeE7012, gDien: model.CodeE5005,
		})},
		{outbound.ErrBereitsVorhanden, mit(jeder(0), map[outbound.Gegenstand]model.Code{
			gSess: model.CodeE7010, gWork: model.CodeE7007, gName: model.CodeE7006, gImp: model.CodeE6102,
		})},
		{model.ErrUngueltig, mit(jeder(0), map[outbound.Gegenstand]model.Code{gImp: model.CodeE6101})},
		{outbound.ErrUngueltigeSequenz, jeder(model.CodeE9001)},
		{outbound.ErrAbgeschlossen, jeder(model.CodeE9001)},
		{outbound.ErrRevisionVeraltet, jeder(model.CodeE9001)},
		{errors.New("unbekannt"), jeder(model.CodeE9001)},
	}
	for _, tc := range tests {
		for _, g := range alleGegenstaende() {
			got, ok := outbound.Abbilden(tc.fehler, g)
			if !ok || got != tc.soll[g] {
				t.Errorf("%v / Gegenstand %d: %v (%v), erwartet %v", tc.fehler, g, got, ok, tc.soll[g])
			}
		}
	}
}

func TestAbbildungVerpackt(t *testing.T) {
	verpackt := fmt.Errorf("lesen: %w", fmt.Errorf("Blob: %w", outbound.ErrBlobFehlt))
	if got, _ := outbound.Abbilden(verpackt, gSonst); got != model.CodeE3001 {
		t.Errorf("verpackt: %v", got)
	}
	// erster Treffer der Tabelle gewinnt, nicht der erste Fehler im Join
	if got, _ := outbound.Abbilden(errors.Join(model.ErrUngueltig, outbound.ErrBlobFehlt), gImp); got != model.CodeE3001 {
		t.Errorf("Join(ErrUngueltig, ErrBlobFehlt): %v", got)
	}
	if got, _ := outbound.Abbilden(errors.Join(outbound.ErrBlobFehlt, model.ErrUngueltig), gImp); got != model.CodeE3001 {
		t.Errorf("Join(ErrBlobFehlt, ErrUngueltig): %v", got)
	}
	if got, _ := outbound.Abbilden(errors.Join(outbound.ErrNichtGefunden, outbound.ErrPfad), gSess); got != model.CodeE2001 {
		t.Errorf("Join(ErrNichtGefunden, ErrPfad): %v", got)
	}
}

func TestAbbildungKeinCode(t *testing.T) {
	for name, err := range map[string]error{
		"nil": nil, "Canceled": context.Canceled, "Deadline": context.DeadlineExceeded,
		"verpackt": fmt.Errorf("x: %w", context.Canceled), "Errorf mit %w": fmt.Errorf("%w", context.Canceled),
		"Join mit Blob": errors.Join(outbound.ErrBlobFehlt, context.Canceled), "Join Canceled zuerst": errors.Join(context.Canceled, outbound.ErrBlobFehlt),
		"Join Blob und Deadline": errors.Join(outbound.ErrBlobFehlt, context.DeadlineExceeded),
		"Wrap um Join":           fmt.Errorf("x: %w", errors.Join(outbound.ErrNichtGefunden, context.DeadlineExceeded)),
	} {
		if c, ok := outbound.Abbilden(err, gSonst); ok || c != 0 {
			t.Errorf("%s: %v %v", name, c, ok)
		}
	}
}

// fehlerVariablen liest die Namen der Variablen `Err…` auf Paketebene aus den Quelltexten
// des Verzeichnisses, ohne die Testdateien.
func fehlerVariablen(t *testing.T, verzeichnis string) []string {
	t.Helper()
	eintraege, err := os.ReadDir(verzeichnis)
	if err != nil {
		t.Fatal(err)
	}
	var namen []string
	fset := token.NewFileSet()
	for _, e := range eintraege {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(verzeichnis, e.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			gen, ok := d.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				for _, n := range spec.(*ast.ValueSpec).Names {
					if strings.HasPrefix(n.Name, "Err") {
						namen = append(namen, n.Name)
					}
				}
			}
		}
	}
	slices.Sort(namen)
	return namen
}

// TestAbbildungVollstaendig hält die Variablen `Err…` des Pakets gegen die Tabelle (SPEC-109 Nr. 8).
func TestAbbildungVollstaendig(t *testing.T) {
	vorhanden := fehlerVariablen(t, ".")
	if len(vorhanden) < 15 {
		t.Fatalf("nur %d Variablen gefunden: %v", len(vorhanden), vorhanden)
	}
	eintraege := outbound.EintraegeAbbildung()
	for _, n := range vorhanden {
		if !slices.Contains(eintraege, n) {
			t.Errorf("%s hat keinen Eintrag in der Abbildung", n)
		}
	}
	for _, n := range eintraege {
		if n != "ErrUngueltig" && !slices.Contains(vorhanden, n) {
			t.Errorf("Eintrag %s ohne Variable im Paket", n)
		}
	}
}

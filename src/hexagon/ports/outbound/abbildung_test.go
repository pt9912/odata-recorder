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
		{model.ErrURLUngueltig, jeder(model.CodeE6005)},
		{model.ErrURLNichtUnterstuetzt, jeder(model.CodeE6001)},
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
		modell := []string{"ErrUngueltig", "ErrMetadatenUngueltig", "ErrDatenUngueltig", "ErrURLUngueltig", "ErrURLNichtUnterstuetzt", "ErrFilterNichtBeobachtet", "ErrKompatibilitaet"}
		if !slices.Contains(modell, n) && !slices.Contains(vorhanden, n) {
			t.Errorf("Eintrag %s ohne Variable im Paket", n)
		}
	}
}

// TestAbbildungMetadaten (Zeile H2): der Fehler des Metadatenlesers bildet bei jedem Gegenstand auf ODR-E6004 ab.
func TestAbbildungMetadaten(t *testing.T) {
	e := &model.MetadatenFehler{Ursache: model.UrsacheKette, Ort: "Ns.T"}
	formen := map[string]error{"Fehler": e, "Errorf": fmt.Errorf("lesen: %w", e), "Join": errors.Join(errors.New("x"), e)}
	for name, err := range formen {
		for _, g := range alleGegenstaende() {
			if got, ok := outbound.Abbilden(err, g); !ok || got != model.CodeE6004 {
				t.Errorf("%s / Gegenstand %d: %v %v", name, g, got, ok)
			}
		}
	}
	if !slices.Contains(outbound.EintraegeAbbildung(), "ErrMetadatenUngueltig") {
		t.Error("die Tabelle nennt ErrMetadatenUngueltig nicht")
	}
	if _, ok := outbound.Abbilden(errors.Join(e, context.Canceled), gImp); ok {
		t.Error("ein Kontextfehler neben dem Fehler bekommt einen Code")
	}
}

// TestAbbildungDaten (Zeile K2): der Fehler des Lesers der JSON-Antworten bildet bei jedem Gegenstand auf
// ODR-E6101 ab, auch verpackt und verbunden; ein Kontextfehler daneben bekommt keinen Code.
func TestAbbildungDaten(t *testing.T) {
	e := &model.DatenFehler{Ursache: model.DatenSchluessel, Ort: "/value/0/Id"}
	formen := map[string]error{"Fehler": e, "Errorf": fmt.Errorf("lesen: %w", e), "Join": errors.Join(errors.New("x"), e)}
	for name, err := range formen {
		for _, g := range alleGegenstaende() {
			if got, ok := outbound.Abbilden(err, g); !ok || got != model.CodeE6101 {
				t.Errorf("%s / Gegenstand %d: %v %v", name, g, got, ok)
			}
		}
	}
	if !slices.Contains(outbound.EintraegeAbbildung(), "ErrDatenUngueltig") {
		t.Error("die Tabelle nennt ErrDatenUngueltig nicht")
	}
	if _, ok := outbound.Abbilden(errors.Join(e, context.Canceled), gImp); ok {
		t.Error("ein Kontextfehler neben dem Fehler bekommt einen Code")
	}
}

// TestAbbildungURL (Zeile 61): der Fehler des Lesers der OData-URL bildet bei jedem der acht Gegenstände auf den Code seiner Klasse
// ab (ODR-E6005 für `ungueltig`, ODR-E6001 für `nicht_unterstuetzt`), auch verpackt und verbunden; 16 Ursachen × 8 Gegenstände.
func TestAbbildungURL(t *testing.T) {
	n := 0
	for _, u := range model.AlleURLUrsachen() {
		soll := model.CodeE6005
		if u.NichtUnterstuetzt() {
			soll = model.CodeE6001
		}
		e := &model.URLFehler{Ursache: u, Ort: "query/1"}
		formen := map[string]error{"Fehler": e, "Errorf": fmt.Errorf("lesen: %w", e), "Join": errors.Join(errors.New("x"), e)}
		for name, err := range formen {
			for _, g := range alleGegenstaende() {
				if got, ok := outbound.Abbilden(err, g); !ok || got != soll {
					t.Errorf("%s %s / Gegenstand %d: %v %v, erwartet %v", u, name, g, got, ok, soll)
				}
				n++
			}
		}
		if _, ok := outbound.Abbilden(errors.Join(e, context.Canceled), gSonst); ok {
			t.Error("ein Kontextfehler neben dem Fehler bekommt einen Code")
		}
	}
	if n != 16*3*8 {
		t.Errorf("%d Fälle", n)
	}
	for _, name := range []string{"ErrURLUngueltig", "ErrURLNichtUnterstuetzt"} {
		if !slices.Contains(outbound.EintraegeAbbildung(), name) {
			t.Errorf("die Tabelle nennt %s nicht", name)
		}
	}
}

// TestAbbildungKompat (Zeile 13): jeder Fehler, der ErrKompatibilitaet trägt, bildet bei jedem der acht Gegenstände auf ODR-E6006
// ab; 6 Ursachen × 8 Formen (der Sentinel roh und verpackt, der Zeiger roh, einfach und doppelt verpackt, verbunden links und
// rechts, mit context.Canceled verbunden: kein Code) × 8 Gegenstände. Ein Fehler der Gruppe ODR-E9001 daneben ändert den Code
// nicht (die Zeile steht vor der Gruppe).
func TestAbbildungKompat(t *testing.T) {
	n := 0
	for _, u := range model.AlleKompatUrsachen() {
		e := &model.KompatFehler{Ursache: u, Ort: "query/1"}
		formen := map[string]error{
			"Sentinel": model.ErrKompatibilitaet, "SentinelVerpackt": fmt.Errorf("lesen: %w", model.ErrKompatibilitaet), "Zeiger": e,
			"Einfach": fmt.Errorf("lesen: %w", e), "Doppelt": fmt.Errorf("a: %w", fmt.Errorf("b: %w", e)),
			"JoinLinks": errors.Join(e, errors.New("x")), "JoinRechts": errors.Join(errors.New("x"), e),
			"Abbruch": errors.Join(e, context.Canceled),
		}
		for name, err := range formen {
			for _, g := range alleGegenstaende() {
				got, ok := outbound.Abbilden(err, g)
				if name == "Abbruch" {
					if ok {
						t.Errorf("%s %s / Gegenstand %d: Code %v neben context.Canceled", u, name, g, got)
					}
				} else if !ok || got != model.CodeE6006 {
					t.Errorf("%s %s / Gegenstand %d: %v %v, erwartet ODR-E6006", u, name, g, got, ok)
				}
				n++
			}
		}
		for _, daneben := range []error{outbound.ErrAbgeschlossen, outbound.ErrUngueltigeSequenz, outbound.ErrRevisionVeraltet} {
			for _, err := range []error{errors.Join(e, daneben), errors.Join(daneben, e)} {
				if got, ok := outbound.Abbilden(err, gSonst); !ok || got != model.CodeE6006 {
					t.Errorf("%s neben %v: %v %v, erwartet ODR-E6006", u, daneben, got, ok)
				}
			}
		}
	}
	if n != 6*8*8 {
		t.Errorf("%d Fälle", n)
	}
	if !slices.Contains(outbound.EintraegeAbbildung(), "ErrKompatibilitaet") {
		t.Error("die Tabelle nennt ErrKompatibilitaet nicht")
	}
}

// TestAbbildungAbfrage (Zeile 24 von slice-query-interpreter): der Fehler des Interpreters bildet bei jedem der acht Gegenstände auf
// den Code seiner Ursache ab (ODR-E6005 für eigenschaft_unbekannt, ODR-E6001 für typ_nicht_unterstuetzt und
// ordnung_nicht_unterstuetzt, ODR-E6003 für nicht_beobachtet), auch verpackt und verbunden. 4 Ursachen × 3 Optionen × 4 Formen × 8
// Gegenstände.
func TestAbbildungAbfrage(t *testing.T) {
	soll := map[model.AbfrageUrsache]model.Code{
		model.AbfrageEigenschaftUnbekannt: model.CodeE6005, model.AbfrageTypNichtUnterstuetzt: model.CodeE6001,
		model.AbfrageOrdnungNichtUnterstuetzt: model.CodeE6001, model.AbfrageNichtBeobachtet: model.CodeE6003,
	}
	n := 0
	for _, u := range model.AlleAbfrageUrsachen() {
		for _, o := range []model.AbfrageOption{model.AbfrageFilter, model.AbfrageOrderBy, model.AbfrageSelect} {
			e := &model.AbfrageFehler{Option: o, Ursache: u}
			formen := map[string]error{"Fehler": e, "Errorf": fmt.Errorf("ausführen: %w", e), "JoinLinks": errors.Join(e, errors.New("x")), "JoinRechts": errors.Join(errors.New("x"), e)}
			for name, err := range formen {
				for _, g := range alleGegenstaende() {
					if got, ok := outbound.Abbilden(err, g); !ok || got != soll[u] {
						t.Errorf("%s %s %s / Gegenstand %d: %v %v, erwartet %v", u, o, name, g, got, ok, soll[u])
					}
					n++
				}
			}
		}
	}
	if n != 4*3*4*8 {
		t.Errorf("%d Fälle", n)
	}
}

// TestAbbildungFilterTyp (Zeile 36): der Fehler der Typprüfung des Filters bildet bei jedem der acht Gegenstände auf den Code seiner
// Ursache ab (ODR-E6005 für vier Ursachen, ODR-E6001 für zwei, ODR-E6003 für nicht_beobachtet), auch verpackt und verbunden; ein
// Kontextfehler daneben bekommt keinen Code. 7 Ursachen × 4 Formen × 8 Gegenstände; die Tabelle nennt ErrFilterNichtBeobachtet.
func TestAbbildungFilterTyp(t *testing.T) {
	n := 0
	for _, u := range model.AlleFilterTypUrsachen() {
		soll := model.CodeE6005
		switch u {
		case model.FilterTypNichtUnterstuetzt, model.FilterOperatorNichtUnterstuetzt:
			soll = model.CodeE6001
		case model.FilterNichtBeobachtet:
			soll = model.CodeE6003
		}
		e := &model.FilterTypFehler{Ursache: u}
		formen := map[string]error{"Fehler": e, "Errorf": fmt.Errorf("prüfen: %w", e), "JoinLinks": errors.Join(e, errors.New("x")), "JoinRechts": errors.Join(errors.New("x"), e)}
		for name, err := range formen {
			for _, g := range alleGegenstaende() {
				if got, ok := outbound.Abbilden(err, g); !ok || got != soll {
					t.Errorf("%s %s / Gegenstand %d: %v %v, erwartet %v", u, name, g, got, ok, soll)
				}
				n++
			}
		}
		if _, ok := outbound.Abbilden(errors.Join(e, context.Canceled), gSonst); ok {
			t.Error("ein Kontextfehler neben dem Fehler bekommt einen Code")
		}
	}
	if n != 7*4*8 {
		t.Errorf("%d Fälle", n)
	}
	if !slices.Contains(outbound.EintraegeAbbildung(), "ErrFilterNichtBeobachtet") {
		t.Error("die Tabelle nennt ErrFilterNichtBeobachtet nicht")
	}
}

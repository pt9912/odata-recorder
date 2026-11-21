package mock_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Die Tests dieser Datei halten den Span `odata.workspace.mutation` (SPEC-290 Nr. 2 b, 3, 4 c und d, 6 T5 und T6, 8, 9; Zeilen M1 bis M10 des
// Plans). Die Schlüssel der Pflicht-Attribute und die Werte von `recorder.mutation_type` stehen als Literal im Test, nie als Konstante des Dienstes.
const (
	attrWsID = "recorder.workspace_id"
	attrArt  = "recorder.mutation_type"
)

// mutationsOp ist eine der vier Operationen mit Span (SPEC-290 Nr. 2 b): Methode, Wert von `recorder.mutation_type` und ein Schlüssel,
// an dem die Mutation gelingt (POST einen neuen, sonst einen vorhandenen).
type mutationsOp struct{ methode, art, id string }

func mutationsOps() []mutationsOp {
	return []mutationsOp{{"POST", "post", "7"}, {"PATCH", "patch", "2"}, {"PUT", "put", "2"}, {"DELETE", "delete", "2"}}
}

// pflichtMutation sind die Pflicht-Attribute von SpanWorkspaceMutation aus der Tabelle der Spans, nicht aus dem Code des Dienstes.
func pflichtMutation(t *testing.T) []string {
	t.Helper()
	for _, s := range support.Spans() {
		if s.Name == support.SpanWorkspaceMutation {
			return s.Pflicht
		}
	}
	t.Fatal("support.Spans() nennt odata.workspace.mutation nicht")
	return nil
}

// TestPflichtSchluesselMutation hält, dass die Literale des Tests die Pflicht-Attribute der Tabelle sind (SPEC-030).
func TestPflichtSchluesselMutation(t *testing.T) {
	if got := pflichtMutation(t); !slices.Equal(got, []string{attrWsID, attrArt}) {
		t.Errorf("Pflicht-Attribute %v, erwartet [%s %s]", got, attrWsID, attrArt)
	}
}

// erwarteMutationsSpan prüft nach einem Aufruf den Stand des Fakes (M1, M3, M5; SPEC-290 Nr. 3, 4, 6): ein Span `odata.workspace.mutation`
// mit einem StartSpan, einem End, ohne Set, Log, CountError und Snapshot; die Schlüssel des StartSpan sind genau die Pflicht-Attribute in
// ihrer Folge, ihre Werte die Kennung des gelesenen Workspace und die Art; kein Attribut trägt den Kanarienwert.
func erwarteMutationsSpan(t *testing.T, kontext string, u *umg, ws, art string) {
	t.Helper()
	f := u.tel.FakeTelemetry
	zaehler := map[supporttest.Operation]int{
		supporttest.OpStartSpan: 1, supporttest.OpEnd: 1, supporttest.OpSet: 0, supporttest.OpLog: 0, supporttest.OpCountError: 0,
		supporttest.OpSnapshot: 0,
	}
	for op, soll := range zaehler {
		if got := f.Anzahl(op); got != soll {
			t.Errorf("%s: Operation %d %d-mal, erwartet %d", kontext, op, got, soll)
		}
	}
	buch := buecher(f)
	if len(buch) != 1 || buch[0].name != support.SpanWorkspaceMutation || buch[0].enden != 1 || buch[0].sets != 0 {
		t.Errorf("%s: Bücher %+v", kontext, buch)
		return
	}
	pflicht := pflichtMutation(t)
	for _, a := range f.Aufrufe() {
		if a.Op == supporttest.OpStartSpan && (len(a.Attrs) != 2 || a.Attrs[0].Key != pflicht[0] || a.Attrs[1].Key != pflicht[1]) {
			t.Errorf("%s: StartSpan mit %v, erwartet allein %v", kontext, a.Attrs, pflicht)
		}
	}
	for _, a := range alleAttribute(f) {
		if !slices.Contains(pflicht, a.Key) || a.Art != support.ArtString || strings.Contains(a.Text, kanarienwert) {
			t.Errorf("%s: Attribut %+v außerhalb der Pflicht oder mit Kanarienwert", kontext, a)
		}
	}
	if got, _ := buch[0].wert(attrWsID); got != ws {
		t.Errorf("%s: %s %q, erwartet %q", kontext, attrWsID, got, ws)
	}
	if got, _ := buch[0].wert(attrArt); got != art {
		t.Errorf("%s: %s %q, erwartet %q", kontext, attrArt, got, art)
	}
}

// TestMutationSpanJeOperation hält M1 und M4 (SPEC-290 Nr. 2 b, 4 d, 6 T5 und T6): jede der vier Mutationen öffnet odata.workspace.mutation
// einmal und beendet ihn einmal, ohne Set; die Schlüssel sind genau die Pflicht-Attribute, der Wert von mutation_type ist die Methode in
// Kleinschreibung (Orakel Literal je Methode), die Kennung die des gelesenen Workspace. Methode 4.
func TestMutationSpanJeOperation(t *testing.T) {
	n := 0
	for _, op := range mutationsOps() {
		n++
		u := schreibUmg(t)
		a, err := u.mutiere(op.methode, op.id, eingabeIDName(op.id, "x"))
		erwarteStatus(t, op.methode, a, err, sollStatus(op.methode))
		erwarteMutationsSpan(t, op.methode, u, wsID, op.art)
	}
	if n != 4 || len(mutationsOps()) != 4 {
		t.Errorf("%d Operationen, erwartet 4 (SPEC-192 Nr. 3)", n)
	}
}

// TestMutationKleinschreibungKeinSpan hält M4 (SPEC-192 Nr. 3, SPEC-290 Nr. 7 a): die Methode zählt in der Schreibweise des Requests; eine
// Methode in Kleinschreibung oder gemischt ist ODR-E6002 und trägt keinen Span, schreibt nichts. Fünf Formen.
func TestMutationKleinschreibungKeinSpan(t *testing.T) {
	formen := []string{"post", "Post", "patch", "put", "delete"}
	for _, methode := range formen {
		u := schreibUmg(t)
		_, err := u.mutiere(methode, "2", eingabeIDName("2", "x"))
		pruefeFehler(t, methode, err, e6002)
		if aufrufe := u.tel.Aufrufe(); len(aufrufe) != 0 {
			t.Errorf("%s: %d Aufrufe von Telemetrie, erwartet 0", methode, len(aufrufe))
		}
		u.ohneSchreiben(methode)
	}
	if len(formen) != 5 {
		t.Errorf("%d Formen, erwartet 5", len(formen))
	}
}

// TestMutationSpanFolge hält M2 (SPEC-290 Nr. 3 a bis c): die Folge von Ports, Transaktion und Telemetrie je Methode als Literal; der Span
// beginnt nach Workspaces.Get und vor Metadaten.Get und endet hinter dem Ende der Transaktion; zwischen dem Eintrag der Transaktion und
// ihrem Ende steht kein Aufruf von Telemetrie. Vier Folgen.
func TestMutationSpanFolge(t *testing.T) {
	kopf := []string{"URL.ParseURL", "Workspaces.Get", "Telemetrie.StartSpan", "Metadaten.Get", "Blobs.Open", "MetadatenCodec.ParseMetadata"}
	transaktion := []string{"Transaktion.WithWorkspaceTransaction", "Workspaces.Get", "Entities.Get", "Entities.Put", "Workspaces.Advance"}
	ende := []string{"Transaktion.Ende", "Span.End"}
	faelle := []struct {
		methode, id string
		body        bool
		antwort     []string
	}{
		{"POST", "7", true, []string{"Schreiber.SchreibeEntity", "Schreiber.Adresse"}},
		{"PATCH", "2", true, []string{"Schreiber.SchreibeAnzahl"}},
		{"PUT", "2", true, []string{"Schreiber.SchreibeAnzahl"}},
		{"DELETE", "2", false, []string{"Schreiber.SchreibeAnzahl"}},
	}
	for _, f := range faelle {
		u := schreibUmg(t)
		u.tx.ende = true
		u.tel.setzeMitProt()
		a, err := u.mutiere(f.methode, f.id, eingabeIDName(f.id, "x"))
		erwarteStatus(t, f.methode, a, err, sollStatus(f.methode))
		want := slices.Clone(kopf)
		if f.body {
			want = append(want, "Eingabe.LeseEntity")
		}
		want = append(append(append(want, transaktion...), f.antwort...), ende...)
		got := u.p.ops()
		if !slices.Equal(got, want) {
			t.Errorf("%s: Folge\n%v\nerwartet\n%v", f.methode, got, want)
		}
		innen := false
		for _, op := range got {
			switch {
			case op == "Transaktion.WithWorkspaceTransaction":
				innen = true
			case op == "Transaktion.Ende":
				innen = false
			case innen && (strings.HasPrefix(op, "Telemetrie.") || strings.HasPrefix(op, "Span.")):
				t.Errorf("%s: %s innerhalb der Transaktion", f.methode, op)
			}
		}
	}
}

// mutRueckweg ist eine Aufrufstelle S3 bis S5 oder S17 bis S26 hinter dem Beginn der Mutation (SPEC-290 Nr. 6): die Operation, an der der
// Fehler eingespeist wird (nr: beim n-ten Aufruf, 0 bei jedem), die Methode, die sie erreicht, und der Schlüssel; S25 hat kein ctx.
type mutRueckweg struct {
	stelle, op, methode, id string
	nr                      int
	ohneCtx                 bool
}

func mutRueckwege() []mutRueckweg {
	return []mutRueckweg{
		{stelle: "S3", op: "Metadaten.Get", methode: "PATCH", id: "2"},
		{stelle: "S4", op: "Blobs.Open", methode: "PATCH", id: "2"},
		{stelle: "S5", op: "MetadatenCodec.ParseMetadata", methode: "PATCH", id: "2"},
		{stelle: "S17", op: "Eingabe.LeseEntity", methode: "PATCH", id: "2"},
		{stelle: "S18", op: "Transaktion.WithWorkspaceTransaction", methode: "PATCH", id: "2"},
		{stelle: "S19", op: "Workspaces.Get", methode: "PATCH", id: "2", nr: 2},
		{stelle: "S20", op: "Entities.Get", methode: "PATCH", id: "2"},
		{stelle: "S21", op: "Coverage.Get", methode: "PATCH", id: "99"},
		{stelle: "S22", op: "Entities.Put", methode: "PATCH", id: "2"},
		{stelle: "S23", op: "Workspaces.Advance", methode: "PATCH", id: "2"},
		{stelle: "S24", op: "Schreiber.SchreibeEntity", methode: "POST", id: "7"},
		{stelle: "S25", op: "Schreiber.Adresse", methode: "POST", id: "7", ohneCtx: true},
		{stelle: "S26", op: "Schreiber.SchreibeAnzahl", methode: "PATCH", id: "2"},
	}
}

// TestMutationEndeJederRueckweg hält M3 (SPEC-290 Nr. 3 b, 6, 7 b; Raster, Zusatztabelle): genau ein End auf jedem Rückweg der Mutation hinter
// dem Beginn, mit Fehler mit Code (13 Stellen), mit Kontextfehler ohne Code (12 Stellen, alle außer S25) und mit dem Paar aus Kontextfehler
// und Fehler der Tabelle (`errors.Join`, SPEC-109 Nr. 6; 12 Stellen); je Fall bleibt das Ergebnis der Code bzw. der Kontextfehler ohne Code,
// keine Antwort, Entity-State und Revision wie vor dem Aufruf (Rollback des Fakes der Transaktion), der Span trägt allein seine zwei Werte
// und nie den Text des Fehlers. Fälle: 13 + 12 + 12 = 37.
func TestMutationEndeJederRueckweg(t *testing.T) {
	stellen := mutRueckwege()
	art := map[string]string{"PATCH": "patch", "POST": "post"}
	formen := []struct {
		name   string
		err    error
		code   bool
		mitCtx bool
	}{
		{"Fehler mit Code", errors.New("roh " + kanarienwert), true, false},
		{"Kontextfehler", context.Canceled, false, true},
		{"Kontextfehler und Fehler der Tabelle", errors.Join(context.Canceled, outbound.ErrBeschaedigt), false, true},
	}
	mitCode, mitKontext, paar := 0, 0, 0
	for _, s := range stellen {
		for _, f := range formen {
			if f.mitCtx && s.ohneCtx {
				continue
			}
			kontext := s.stelle + " mit " + f.name
			u := schreibUmg(t)
			if s.nr > 0 {
				u.p.setzeFehlerBeiNr(s.op, s.nr, f.err)
			} else {
				u.p.setzeFehler(s.op, f.err)
			}
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			a, err := u.mutiere(s.methode, s.id, eingabeIDName(s.id, "x"))
			if f.code {
				mitCode++
				_ = erwarteFehler(t, kontext, err, model.CodeE9001)
			} else {
				if f.name == "Kontextfehler" {
					mitKontext++
				} else {
					paar++
				}
				if !errors.Is(err, context.Canceled) || fehlercode(err) != 0 {
					t.Errorf("%s: Fehler %v, erwartet Canceled ohne Code", kontext, err)
				}
			}
			if a.Body != nil || a.Status != 0 {
				t.Errorf("%s: Antwort %d trotz Fehler", kontext, a.Status)
			}
			erwarteMutationsSpan(t, kontext, u, wsID, art[s.methode])
			u.unveraendert(vorher, rev, kontext)
		}
	}
	if len(stellen) != 13 || mitCode != 13 || mitKontext != 12 || paar != 12 {
		t.Errorf("%d Stellen, %d mit Code, %d mit Kontextfehler, %d Paare; erwartet 13, 13, 12 und 12 (SPEC-195 Nr. 11)", len(stellen), mitCode, mitKontext, paar)
	}
}

// kette ist der Text eines Aufrufs als Kette der Namen (`m.k.Metadaten.Get`); ein anderer Ausdruck ergibt `?`.
func kette(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return kette(x.X) + "." + x.Sel.Name
	}
	return "?"
}

// mutMethoden sind die Methoden, die die Stelle erreichen: der Leser (S17) liest keinen Body bei DELETE, die Coverage in f (S21) und die
// Antwort `204` (S26) gibt es nicht bei POST, und die Antwort von POST (S24, S25) allein bei POST.
func mutMethoden(stelle string) []string {
	switch stelle {
	case "S17":
		return []string{"POST", "PATCH", "PUT"}
	case "S21", "S26":
		return []string{"PATCH", "PUT", "DELETE"}
	case "S24", "S25":
		return []string{"POST"}
	}
	return []string{"POST", "PATCH", "PUT", "DELETE"}
}

// rueckwegFall ruft die Mutation der Methode an der Stelle s, speist fehler an ihrer Operation ein (bau bereitet einen Ausgang ohne Fehler
// eines Ports vor) und prüft: Ergebnis der Code, 0 für den Kontextfehler ohne Code; keine Antwort; Entity-State und Revision wie vorher; ein
// Span mit Art und Kennung, einem End, ohne Zähler.
func rueckwegFall(t *testing.T, kontext string, s mutRueckweg, methode string, fehler error, bau func(u *umg), code model.Code) {
	t.Helper()
	id := "2"
	switch {
	case methode == "POST":
		id = "7"
	case s.stelle == "S21":
		id = "99"
	}
	u := schreibUmg(t)
	if bau != nil {
		bau(u)
	}
	switch {
	case fehler == nil:
	case s.nr > 0:
		u.p.setzeFehlerBeiNr(s.op, s.nr, fehler)
	default:
		u.p.setzeFehler(s.op, fehler)
	}
	vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
	a, err := u.mutiere(methode, id, eingabeIDName(id, "x"))
	if code != 0 {
		_ = erwarteFehler(t, kontext, err, code)
	} else if !errors.Is(err, context.Canceled) || fehlercode(err) != 0 {
		t.Errorf("%s: Fehler %v, erwartet Canceled ohne Code", kontext, err)
	}
	if a.Body != nil || a.Status != 0 {
		t.Errorf("%s: Antwort %d trotz Fehler", kontext, a.Status)
	}
	erwarteMutationsSpan(t, kontext, u, wsID, strings.ToLower(methode))
	u.unveraendert(vorher, rev, kontext)
}

// mutCode ist ein Fehler an einer Stelle hinter dem Beginn und der Code, den die Spezifikation für ihn an dieser Stelle nennt: der Gegenstand
// je Stelle steht in SPEC-195 Nr. 11, die Abbildung in SPEC-109 Nr. 5 (als Handkopie, nie aus dem Code). Ein Eintrag mit bau und ohne fehler
// ist ein Ausgang ohne Fehler eines Ports (der Service ohne Metadaten, SPEC-195 Nr. 9).
type mutCode struct {
	stelle string
	fehler error
	code   model.Code
	bau    func(u *umg)
}

func mutCodes() []mutCode {
	roh, nichtGefunden, beschaedigt := errors.New("roh"), outbound.ErrNichtGefunden, outbound.ErrBeschaedigt
	e3001, e5005, e6004, e6007, e6002, e7001, e9001 := model.CodeE3001, model.CodeE5005, model.CodeE6004, model.CodeE6007, model.CodeE6002,
		model.CodeE7001, model.CodeE9001
	c := func(stelle string, err error, code model.Code) mutCode {
		return mutCode{stelle: stelle, fehler: err, code: code}
	}
	eingabe := &model.EingabeFehler{Ursache: model.EingabeSyntax, Ort: kanarienwert}
	return []mutCode{
		c("S3", nichtGefunden, e5005), c("S3", beschaedigt, e3001), c("S3", roh, e9001),
		{stelle: "S3", code: e6004, bau: func(u *umg) { u.md.setze(model.Service{ID: dienstID}) }},
		c("S4", outbound.ErrBlobFehlt, e3001), c("S4", outbound.ErrBlobBeschaedigt, e3001), c("S4", nichtGefunden, e9001), c("S4", roh, e9001),
		c("S5", model.ErrMetadatenUngueltig, e6004), c("S5", roh, e9001),
		c("S17", eingabe, e6007), c("S17", model.ErrEingabeNichtUnterstuetzt, e6002), c("S17", roh, e9001),
		c("S18", beschaedigt, e3001), c("S18", roh, e9001),
		c("S19", nichtGefunden, e7001), c("S19", beschaedigt, e3001), c("S19", roh, e9001),
		c("S20", beschaedigt, e3001), c("S20", roh, e9001),
		c("S21", beschaedigt, e3001), c("S21", roh, e9001),
		c("S22", nichtGefunden, e7001), c("S22", beschaedigt, e3001), c("S22", roh, e9001),
		c("S23", outbound.ErrRevisionVeraltet, e9001), c("S23", nichtGefunden, e7001), c("S23", beschaedigt, e3001), c("S23", roh, e9001),
		c("S24", model.ErrUngueltig, e9001), c("S24", beschaedigt, e3001), c("S24", roh, e9001),
		c("S25", model.ErrUngueltig, e9001), c("S25", beschaedigt, e3001), c("S25", roh, e9001),
		c("S26", model.ErrUngueltig, e9001), c("S26", beschaedigt, e3001), c("S26", roh, e9001),
	}
}

// TestMutationCodeJederRueckweg hält M3 für die Codes (SPEC-290 Nr. 7 b, SPEC-195 Nr. 10 und 11): an jeder der 13 Stellen hinter dem Beginn
// endet der Span genau einmal mit jedem Code, den die Spezifikation für einen Fehler dieser Stelle nennt (ODR-E5005 aus S3, ODR-E6004 aus S3
// und S5, ODR-E3001 aus S3, S4 und den Stellen mit Zustand, ODR-E6007 und ODR-E6002 aus dem Leser, ODR-E7001 aus dem Workspace, ODR-E9001),
// bei jeder Methode, die die Stelle erreicht. Menge Fehler × Methode: 38 Fehler an 13 Stellen in 126 Fällen, 7 verschiedene Codes, alle
// unter den 16 von SPEC-195 Nr. 10; jede Stelle der Tabelle der Rückwege trägt Fälle.
func TestMutationCodeJederRueckweg(t *testing.T) {
	stellen := map[string]mutRueckweg{}
	for _, s := range mutRueckwege() {
		stellen[s.stelle] = s
	}
	codes := map[string]bool{}
	gesehen := map[string]bool{}
	faelle, n := mutCodes(), 0
	for _, c := range faelle {
		s, da := stellen[c.stelle]
		if !da {
			t.Fatalf("Stelle %s steht nicht in der Tabelle der Rückwege", c.stelle)
		}
		gesehen[c.stelle] = true
		codes[c.code.String()] = true
		for _, methode := range mutMethoden(c.stelle) {
			n++
			var err error
			if c.fehler != nil {
				err = fmt.Errorf("%s: %w", kanarienwert, c.fehler)
			}
			rueckwegFall(t, c.stelle+" "+methode+" "+c.code.String(), s, methode, err, c.bau, c.code)
		}
	}
	for code := range codes {
		if !slices.Contains(spec195Nr10(), code) {
			t.Errorf("%s gehört nicht zu den 16 erreichbaren Codes", code)
		}
	}
	if len(faelle) != 38 || n != 126 || len(gesehen) != len(mutRueckwege()) || len(codes) != 7 {
		t.Errorf("%d Fehler, %d Fälle, %d Stellen, %d Codes; erwartet 38, 126, 13 und 7", len(faelle), n, len(gesehen), len(codes))
	}
}

// TestMutationKontextJederRueckwegJedeMethode hält M3 für den Kontextfehler und das Paar bei jeder Methode, die die Stelle erreicht: die 12
// Stellen mit ctx (alle außer S25), Methoden je Stelle wie bei den Codes: 42 Fälle je Form, 84 Fälle.
func TestMutationKontextJederRueckwegJedeMethode(t *testing.T) {
	formen := []struct {
		name string
		err  error
	}{{"Kontextfehler", context.Canceled}, {"Kontextfehler und Fehler der Tabelle", errors.Join(context.Canceled, outbound.ErrBeschaedigt)}}
	n := 0
	for _, s := range mutRueckwege() {
		if s.ohneCtx {
			continue
		}
		for _, f := range formen {
			for _, methode := range mutMethoden(s.stelle) {
				n++
				rueckwegFall(t, s.stelle+" "+methode+" mit "+f.name, s, methode, f.err, nil, 0)
			}
		}
	}
	if n != 84 {
		t.Errorf("%d Fälle, erwartet 84", n)
	}
}

// entscheidung ist ein Ausgang, den der Dienst hinter dem Beginn ohne Fehler eines Ports bildet (SPEC-290 Nr. 7 b): die Operation, der
// Aufbau und der Code.
type entscheidung struct {
	name, methode string
	art           model.RessourcenArt
	set           string
	teile         []model.SchluesselTeil
	e             model.EntityEingabe
	bau           func(u *umg)
	code          model.Code
}

func entscheidungen() []entscheidung {
	reserviert := model.EntityEingabe{Typ: typProduct, Schluessel: "2", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "2"), pw("Name", model.ArtString, model.Marker)}}
	fremderName := zahl(kanarienwert, "1")
	nichts := func(*umg) {}
	var r []entscheidung
	for _, op := range mutationsOps() {
		art, teile := model.RessourceEntity, []model.SchluesselTeil{zahl("", op.id)}
		if op.methode == "POST" {
			art, teile = model.RessourceMenge, nil
		}
		e := eingabeIDName(op.id, "x")
		r = append(r,
			entscheidung{op.methode + " unbekanntes Set", op.methode, art, kanarienwert + "Set", teile, e, nichts, e6008},
			entscheidung{op.methode + " Set ohne Typ", op.methode, art, "Waise", teile, e, nichts, e9001},
		)
	}
	for _, op := range mutationsOps()[1:] {
		e := eingabeIDName(op.id, "x")
		tomb := func(u *umg) {
			x := voll(2, "t")
			x.Deleted = true
			u.en.setze(setProducts, x)
		}
		r = append(r,
			entscheidung{op.methode + " Schlüssel nicht bildbar", op.methode, model.RessourceEntity, setProducts, []model.SchluesselTeil{fremderName}, e, nichts, e6005},
			entscheidung{op.methode + " Tombstone", op.methode, model.RessourceEntity, setProducts, []model.SchluesselTeil{zahl("", "2")}, e, tomb, e6008},
			entscheidung{op.methode + " fehlend, Coverage complete", op.methode, model.RessourceEntity, setProducts, []model.SchluesselTeil{zahl("", "99")},
				eingabeIDName("99", "x"), nichts, e6008},
			entscheidung{op.methode + " fehlend, Coverage partial", op.methode, model.RessourceEntity, setProducts, []model.SchluesselTeil{zahl("", "99")},
				eingabeIDName("99", "x"), func(u *umg) { u.co.setze(baselineID, setProducts, model.CoveragePartial) }, e6003},
		)
	}
	for _, methode := range []string{"POST", "PATCH", "PUT"} {
		art, teile := model.RessourceEntity, []model.SchluesselTeil{zahl("", "2")}
		if methode == "POST" {
			art, teile = model.RessourceMenge, nil
		}
		r = append(r, entscheidung{methode + " reservierter Text", methode, art, setProducts, teile, reserviert, nichts, e6007})
	}
	for _, methode := range []string{"PATCH", "PUT"} {
		r = append(r, entscheidung{methode + " Schlüssel des Bodys widerspricht", methode, model.RessourceEntity, setProducts, []model.SchluesselTeil{zahl("", "2")},
			eingabeIDName("3", "x"), nichts, e6007})
	}
	r = append(r, entscheidung{"POST ohne Schlüssel im Body", "POST", model.RessourceMenge, setProducts, nil,
		model.EntityEingabe{Typ: typProduct, Properties: []model.PropertyWert{pw("Name", model.ArtString, "x")}}, nichts, e9001})
	r = append(r, entscheidung{"PATCH an einen Eintrag mit partial", "PATCH", model.RessourceEntity, setProducts, []model.SchluesselTeil{zahl("", "2")},
		eingabeIDName("2", "x"), func(u *umg) {
			x := voll(2, "p")
			x.Completeness = model.EntityPartial
			u.en.setze(setProducts, x)
		}, e6003})
	return r
}

// TestMutationEndeJedeEntscheidung hält M3 für die Ausgänge, die der Dienst nach dem Beginn selbst bildet (SPEC-290 Nr. 6, 7 b: jeder Code):
// ein End, Name und Art unverändert, Entity-State und Revision wie vorher. Die Codes ODR-E6009 (POST auf vorhandenen Schlüssel), ODR-E6010,
// ODR-E6002 und ODR-E6007 des Lesers führt codeSzenarien (TestKeinZaehlerImDienst: ein End je Szenario), ODR-E7001 hinter dem Beginn
// TestMutationWorkspaceKennung. 27 Fälle.
func TestMutationEndeJedeEntscheidung(t *testing.T) {
	fs := entscheidungen()
	for _, f := range fs {
		u := schreibUmg(t)
		f.bau(u)
		vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
		u.ressource(f.art, f.set, model.Optionen{}, f.teile...)
		u.ei.ergebnis = f.e
		a := anfrage(f.methode, dienstRoot+"/x")
		a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
		a.Header["Content-Type"] = []string{"application/json"}
		_, err := u.rufe(a)
		_ = erwarteFehler(t, f.name, err, f.code)
		erwarteMutationsSpan(t, f.name, u, wsID, strings.ToLower(f.methode))
		u.unveraendert(vorher, rev, f.name)
	}
	if len(fs) != 27 {
		t.Errorf("%d Fälle, erwartet 27", len(fs))
	}
}

// TestMutationPortAufrufeImQuelltext ist die Zählprüfung gegen die Quelle zu M3 und M7 (SPEC-195 Nr. 11): die Dateien des Pfads der Mutation
// nennen die Aufrufstellen S17 bis S20, S22 bis S24 und S26 genau einmal, S25 einmal; mit S3 bis S5 und S21 (geteilt mit der Query, in
// mock.go und lesen.go) sind das die 13 Stellen hinter dem Beginn. Ein weiterer Aufruf eines dieser Ports in den beiden Dateien macht den
// Test rot, bis die Tabelle der Rückwege eine Zeile dazu führt.
func TestMutationPortAufrufeImQuelltext(t *testing.T) {
	zaehler := map[string]int{}
	for _, d := range []string{"schreiben.go", "schreiben_antwort.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if ce, ok := n.(*ast.CallExpr); ok {
				if sel, ok := ce.Fun.(*ast.SelectorExpr); ok {
					zaehler[sel.Sel.Name]++
				}
			}
			return true
		})
	}
	soll := map[string]int{"LeseEntity": 1, "WithWorkspaceTransaction": 1, "Get": 2, "Put": 1, "Advance": 1, "SchreibeEntity": 1, "Adresse": 1, "SchreibeAnzahl": 1}
	summe := 0
	for name, n := range soll {
		summe += n
		if zaehler[name] != n {
			t.Errorf("%s im Quelltext %d-mal, erwartet %d", name, zaehler[name], n)
		}
	}
	if summe != 9 || summe+4 != len(mutRueckwege()) {
		t.Errorf("%d Aufrufe in den Dateien der Mutation, %d Stellen der Tabelle; erwartet 9 und 13", summe, len(mutRueckwege()))
	}
	// Die vier geteilten Stellen S3, S4, S5 und S21 stehen je einmal in mock.go und lesen.go (Kette des Aufrufs als Text).
	ketten := map[string]int{}
	for _, d := range []string{"mock.go", "lesen.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if ce, ok := n.(*ast.CallExpr); ok {
				ketten[kette(ce.Fun)]++
			}
			return true
		})
	}
	geteilt := map[string]string{"S3": "m.k.Metadaten.Get", "S4": "m.k.Blobs.Open", "S5": "d.Codecs.Metadaten.ParseMetadata", "S21": "m.k.Coverage.Get"}
	for stelle, k := range geteilt {
		if ketten[k] != 1 {
			t.Errorf("%s: %s im Quelltext %d-mal, erwartet 1", stelle, k, ketten[k])
		}
	}
}

// TestMutationWorkspaceKennung hält M5 (SPEC-290 Nr. 4 c, 7 a und d): recorder.workspace_id ist die Kennung des gelesenen Workspace, nie
// der Wert des Headers; der Header ist der Kanarienwert. Ein unbekannter und ein fremder Workspace tragen keinen Span und keinen Aufruf von
// Telemetrie (ODR-E7001, ODR-E5004); ein Workspace, der nach Schritt (f) verschwindet, trägt den Span mit seiner Kennung und endet mit
// ODR-E7001. Methode 4 × {gelesen ≠ Header, unbekannt, fremd, nach (f) gelöscht} = 16.
func TestMutationWorkspaceKennung(t *testing.T) {
	const gelesen = "ws-gelesen"
	n := 0
	for _, op := range mutationsOps() {
		kopf := map[string][]string{headerName: {kanarienwert}}
		anfr := func(u *umg) inbound.Anfrage {
			a := u.anfrageMitBody(op.methode, op.id, eingabeIDName(op.id, "x"), kopf)
			a.Header[headerName] = []string{kanarienwert}
			return a
		}
		{ // gelesen
			n++
			u := schreibUmg(t)
			u.workspaceZuKopf(kanarienwert, gelesen)
			a, err := u.rufe(anfr(u))
			erwarteStatus(t, op.methode+" gelesen", a, err, sollStatus(op.methode))
			erwarteMutationsSpan(t, op.methode+" gelesen", u, gelesen, op.art)
		}
		{ // unbekannt
			n++
			u := schreibUmg(t)
			_, err := u.rufe(anfr(u))
			pruefeFehler(t, op.methode+" unbekannt", err, e7001)
			if aufrufe := u.tel.Aufrufe(); len(aufrufe) != 0 {
				t.Errorf("%s unbekannt: %d Aufrufe von Telemetrie, erwartet 0", op.methode, len(aufrufe))
			}
		}
		{ // fremd
			n++
			u := schreibUmg(t)
			u.ws.setze(model.Workspace{ID: kanarienwert, ServiceID: "anderer", BaselineSnapshotID: baselineID})
			_, err := u.rufe(anfr(u))
			pruefeFehler(t, op.methode+" fremd", err, model.CodeE5004)
			if aufrufe := u.tel.Aufrufe(); len(aufrufe) != 0 {
				t.Errorf("%s fremd: %d Aufrufe von Telemetrie, erwartet 0", op.methode, len(aufrufe))
			}
		}
		{ // nach Schritt (f) gelöscht
			n++
			u := schreibUmg(t)
			u.ws.nach = func(k int) {
				if k == 1 {
					u.ws.entferne(wsID)
				}
			}
			_, err := u.mutiere(op.methode, op.id, eingabeIDName(op.id, "x"))
			pruefeFehler(t, op.methode+" gelöscht", err, e7001)
			erwarteMutationsSpan(t, op.methode+" gelöscht", u, wsID, op.art)
		}
	}
	if n != 16 {
		t.Errorf("%d Fälle, erwartet 16", n)
	}
}

// workspaceZuKopf lässt Workspaces.Get zum Header kopf den Workspace gelesen liefern (die Fixture ist ungleich dem Wert des Headers); die
// Transaktion und Advance kennen den Workspace unter gelesen.
func (u *umg) workspaceZuKopf(kopf, gelesen string) {
	u.ws.setze(model.Workspace{ID: gelesen, ServiceID: dienstID, BaselineSnapshotID: baselineID})
	u.ws.ersatz = func(id string) (model.Workspace, error) {
		if id == kopf {
			id = gelesen
		}
		x, da := u.ws.m[id]
		if !da {
			return model.Workspace{}, outbound.ErrNichtGefunden
		}
		return x, nil
	}
}

// putAbbruch ist der Fake von Put, der den Eintrag schreibt, den ctx des Aufrufs beendet und dessen Fehler liefert: der Abbruch mitten in der
// Transaktion (SPEC-290 Nr. 3 f, S22); die Transaktion rollt den geschriebenen Eintrag zurück.
type putAbbruch struct {
	outbound.EntityRepository
	cancel func(ws string) context.CancelFunc
}

func (p putAbbruch) Put(ctx context.Context, x model.EntityState) error {
	if err := p.EntityRepository.Put(ctx, x); err != nil {
		return err
	}
	p.cancel(x.WorkspaceID)()
	return ctx.Err()
}

// TestMutationAbbruchInDerTransaktion hält M6 deterministisch (SPEC-290 Nr. 3 f, 9 b, Raster Abbruch): beendet der Put den ctx, ist das Ergebnis
// der Kontextfehler ohne Code, der Eintrag und die Revision sind wie vorher (Rollback), Advance läuft nicht, und der Span endet genau einmal. Methode 4.
func TestMutationAbbruchInDerTransaktion(t *testing.T) {
	n := 0
	for _, op := range mutationsOps() {
		n++
		u := schreibUmg(t)
		ctx, cancel := context.WithCancel(u.ctx)
		u.tx.innen = func(s outbound.WorkspaceStores) outbound.WorkspaceStores {
			s.Entities = putAbbruch{EntityRepository: s.Entities, cancel: func(string) context.CancelFunc { return cancel }}
			return s
		}
		vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
		a := u.anfrageMitBody(op.methode, op.id, eingabeIDName(op.id, "x"), nil)
		antwort, err := u.mocker().Mock(ctx, a)
		cancel()
		if !errors.Is(err, context.Canceled) || fehlercode(err) != 0 || antwort.Body != nil {
			t.Errorf("%s: Fehler %v, erwartet Canceled ohne Code und keine Antwort", op.methode, err)
		}
		if got := u.p.anzahl("Entities.Put"); got != 1 || u.p.anzahl("Workspaces.Advance") != 0 {
			t.Errorf("%s: Put %d, Advance %d, erwartet 1 und 0", op.methode, got, u.p.anzahl("Workspaces.Advance"))
		}
		erwarteMutationsSpan(t, op.methode, u, wsID, op.art)
		u.unveraendert(vorher, rev, op.methode)
	}
	if n != 4 {
		t.Errorf("%d Operationen, erwartet 4", n)
	}
}

// TestMutationCtxUnveraendert hält M7 (SPEC-290 Nr. 3 d, SPEC-195 Nr. 2, 11): StartSpan und jeder Port der Mutation erhalten das ctx des
// Aufrufs, nie das des Spans (der Fake liefert ein abgeleitetes). Neun Stellen S17 bis S24 und S26 in drei Läufen (POST, PATCH an einen
// vorhandenen und an einen fehlenden Eintrag), dazu StartSpan; mit den elf Stellen der Query (S2 bis S12) sind das 20 Stellen mit ctx.
func TestMutationCtxUnveraendert(t *testing.T) {
	lauf := func(methode, id string) *umg {
		u := schreibUmg(t)
		_, _ = u.mutiere(methode, id, eingabeIDName(id, "x"))
		return u
	}
	aufbauten := map[string]*umg{"post": lauf("POST", "7"), "patch": lauf("PATCH", "2"), "fehlend": lauf("PATCH", "99")}
	stellen := []struct{ stelle, op, lauf string }{
		{"S17", "Eingabe.LeseEntity", "post"}, {"S18", "Transaktion.WithWorkspaceTransaction", "post"}, {"S19", "Workspaces.Get", "post"},
		{"S20", "Entities.Get", "post"}, {"S21", "Coverage.Get", "fehlend"}, {"S22", "Entities.Put", "post"}, {"S23", "Workspaces.Advance", "post"},
		{"S24", "Schreiber.SchreibeEntity", "post"}, {"S26", "Schreiber.SchreibeAnzahl", "patch"},
	}
	for _, s := range stellen {
		u := aufbauten[s.lauf]
		ev := u.p.von(s.op)
		if len(ev) == 0 {
			t.Errorf("%s: %s nicht gerufen", s.stelle, s.op)
		}
		for _, e := range ev {
			if e.ctx != u.ctx {
				t.Errorf("%s: %s erhielt nicht das ctx des Aufrufs", s.stelle, s.op)
			}
		}
	}
	for name, u := range aufbauten {
		starts := 0
		for _, a := range u.tel.Aufrufe() {
			if a.Op == supporttest.OpStartSpan {
				starts++
				if any(a.Ctx()) != any(u.ctx) {
					t.Errorf("%s: StartSpan erhielt nicht das ctx des Aufrufs", name)
				}
			}
		}
		if starts != 1 {
			t.Errorf("%s: %d StartSpan, erwartet 1", name, starts)
		}
	}
	if len(stellen) != 9 || len(stellen)+11 != 20 {
		t.Errorf("%d Stellen, erwartet 9 (mit den 11 der Query 20, SPEC-195 Nr. 11)", len(stellen))
	}
}

// stelleMutation ist eine Stelle mit Request-Wert der Mutation (SPEC-290 Nr. 8; Zeile M8): der Aufbau trägt den Kanarienwert an der Stelle in
// den Request, die Methoden sind die, die sie lesen.
type stelleMutation struct {
	name     string
	methoden []string
	aufbau   func(u *umg, methode string) inbound.Anfrage
}

func stellenMutation() []stelleMutation {
	k := kanarienwert
	alle := []string{"POST", "PATCH", "PUT", "DELETE"}
	mitBody := []string{"POST", "PATCH", "PUT"}
	ohnePost := []string{"PATCH", "PUT", "DELETE"}
	schluessel := func(methode string) string { return mutationsOpFuer(methode).id }
	standard := func(u *umg, methode string, kopf map[string][]string) inbound.Anfrage {
		return u.anfrageMitBody(methode, schluessel(methode), eingabeIDName(schluessel(methode), "x"), kopf)
	}
	mitEingabe := func(e func(id string) model.EntityEingabe) func(u *umg, methode string) inbound.Anfrage {
		return func(u *umg, methode string) inbound.Anfrage {
			return u.anfrageMitBody(methode, schluessel(methode), e(schluessel(methode)), nil)
		}
	}
	kopfStelle := func(name string, wert func() []string, methoden []string, etag bool) stelleMutation {
		return stelleMutation{name, methoden, func(u *umg, methode string) inbound.Anfrage {
			if etag {
				u.k.Dienste[0].Etag = true
			}
			return standard(u, methode, map[string][]string{name: wert()})
		}}
	}
	return []stelleMutation{
		{"Name einer Property im Body", mitBody, mitEingabe(func(id string) model.EntityEingabe {
			return model.EntityEingabe{Typ: typProduct, Schluessel: id, Properties: []model.PropertyWert{pw("Id", model.ArtInt32, id), pw(k+"Prop", model.ArtString, "x")}}
		})},
		{"Wert einer Property im Body", mitBody, mitEingabe(func(id string) model.EntityEingabe { return eingabeIDName(id, k) })},
		kopfStelle("Content-Type", func() []string { return []string{"application/json; x=" + k} }, mitBody, false),
		kopfStelle("Prefer", func() []string { return []string{"return=minimal, x-" + k} }, mitBody, false),
		kopfStelle("If-Match", func() []string { return []string{`W/"` + k + `"`} }, ohnePost, true),
		kopfStelle("If-None-Match", func() []string { return []string{`W/"` + k + `"`} }, ohnePost, true),
		{"Set", alle, func(u *umg, methode string) inbound.Anfrage {
			a := standard(u, methode, nil)
			art, teile := model.RessourceEntity, []model.SchluesselTeil{zahl("", schluessel(methode))}
			if methode == "POST" {
				art, teile = model.RessourceMenge, nil
			}
			u.ressource(art, k+"Set", model.Optionen{}, teile...)
			return a
		}},
		{"Schlüssel", ohnePost, func(u *umg, methode string) inbound.Anfrage {
			a := standard(u, methode, nil)
			u.ressource(model.RessourceEntity, setProducts, model.Optionen{},
				model.SchluesselTeil{Name: k, Literal: model.Literal{Art: model.LiteralZeichenkette, Text: "'" + k + "'"}})
			return a
		}},
		{"Header X-Recorder-Workspace", alle, func(u *umg, methode string) inbound.Anfrage {
			u.workspaceZuKopf(k, "ws-gelesen")
			a := standard(u, methode, nil)
			a.Header[headerName] = []string{k}
			return a
		}},
		kopfStelle("Accept", func() []string { return []string{k} }, alle, false),
	}
}

func mutationsOpFuer(methode string) mutationsOp {
	for _, op := range mutationsOps() {
		if op.methode == methode {
			return op
		}
	}
	return mutationsOp{}
}

// TestKeinRequestWertImMutationsSpan hält M8 und L15 (SPEC-290 Nr. 8): ein Kanarienwert an jeder der 10 Stellen des Requests, die eine
// Mutation liest (6 des Bodys und der Header, 4 der Query: Set, Schlüssel, Header des Workspace, Accept), bei jeder Methode, die sie liest,
// erreicht kein Attribut eines Spans und keine Meldung eines Log; die 14 Stellen von Nr. 8 sind diese 10 und die 8 der Query ohne die vier
// gemeinsamen. Jede Anfrage besteht die Zuordnung und verlangt genau einen StartSpan (ein Fehler hinter dem Beginn ändert das nicht).
func TestKeinRequestWertImMutationsSpan(t *testing.T) {
	stellen := stellenMutation()
	if len(stellen) != 10 {
		t.Fatalf("%d Stellen, erwartet 10 (6 + 4)", len(stellen))
	}
	paare := 0
	for _, s := range stellen {
		for _, methode := range s.methoden {
			paare++
			u := schreibUmg(t)
			a := s.aufbau(u, methode)
			_, _ = u.rufe(a)
			if starts := u.tel.Anzahl(supporttest.OpStartSpan); starts != 1 {
				t.Errorf("%s %s: StartSpan %d-mal, erwartet 1", s.name, methode, starts)
			}
			for _, x := range u.tel.Aufrufe() {
				for _, at := range x.Attrs {
					if strings.Contains(at.Text, kanarienwert) || strings.Contains(at.Key, kanarienwert) {
						t.Errorf("%s %s: Attribut %s trägt den Kanarienwert", s.name, methode, at.Key)
					}
				}
				if strings.Contains(x.Msg, kanarienwert) {
					t.Errorf("%s %s: Meldung trägt den Kanarienwert", s.name, methode)
				}
			}
			if u.tel.Anzahl(supporttest.OpLog) != 0 {
				t.Errorf("%s %s: Log gerufen", s.name, methode)
			}
		}
	}
	if paare != 3+3+3+3+3+3+4+3+4+4 {
		t.Errorf("%d Paare, erwartet 33", paare)
	}
	// Zählprüfung gegen SPEC-290 Nr. 8: die 8 Stellen der Query (Q14) und die 6 Stellen allein der Mutation sind die 14 Stellen; vier der
	// 10 Stellen dieser Tabelle (Set, Schlüssel, Header des Workspace, Accept) sind die der Query.
	if got := len(stellenQuery()) + len(stellen) - 4; got != 14 {
		t.Errorf("%d Stellen insgesamt, erwartet 14 (SPEC-290 Nr. 8)", got)
	}
}

// Die Läufe unter Last: 16 Goroutinen in 16 Workspaces führen je Runde die vier Methoden in einer Folge aus, die um die Nummer der
// Goroutine rotiert; so sind zu jeder Stufe der Schranke verschiedene Arten zugleich offen.
const mutGoroutinen = 16

func mutationsFolge(i int) []string {
	basis := []string{"POST", "PATCH", "PUT", "DELETE"}
	return append(slices.Clone(basis[i%4:]), basis[:i%4]...)
}

// mutationsSchluessel ist der Schlüssel der Methode in Runde r: POST einen neuen, DELETE einen vorhandenen je Runde, PATCH und PUT den einen
// Eintrag 10.
func mutationsSchluessel(methode string, r int) string {
	switch methode {
	case "POST":
		return strconv.Itoa(1000 + r)
	case "DELETE":
		return strconv.Itoa(100 + r)
	}
	return "10"
}

// setzeWs legt die Einträge eines Sets für einen Workspace ab (der Fake trennt den Zustand nach Workspace, wenn jeWs gilt).
func (e *entities) setzeWs(ws, set string, eintraege ...model.EntityState) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.m == nil {
		e.m = map[string][]model.EntityState{}
	}
	e.m[e.schl(ws, set)] = slices.Clone(eintraege)
}

// gleichzeitigUmg ist ein Aufbau mit 16 Workspaces w0 bis w15, je mit dem Eintrag 10 und den Einträgen 100 bis 100 + laeufe − 1.
func gleichzeitigUmg(t *testing.T, laeufe int) *umg {
	t.Helper()
	u := neueUmg(t)
	u.en.jeWs = true
	u.c.mc.modell = testModell()
	u.co.setze(baselineID, setProducts, model.CoverageComplete)
	u.dynamisch()
	for i := range mutGoroutinen {
		ws := "w" + strconv.Itoa(i)
		u.ws.setze(model.Workspace{ID: ws, ServiceID: dienstID, BaselineSnapshotID: baselineID})
		schluessel := []int{10}
		for r := range laeufe {
			schluessel = append(schluessel, 100+r)
		}
		var es []model.EntityState
		for _, k := range schluessel {
			x := voll(k, "a")
			x.WorkspaceID = ws
			es = append(es, x)
		}
		u.en.setzeWs(ws, setProducts, es...)
	}
	return u
}

// stufenWorkspaces ist das WorkspaceRepository der Konfig des Dienstes, dessen Get nach dem Lesen die Stufe der Schranke ruft: die Stelle S2
// vor dem Beginn des Spans; die Stores in der Transaktion (S19) bleiben der Fake des Aufbaus und rufen sie nicht.
type stufenWorkspaces struct {
	outbound.WorkspaceRepository
	stufe func()
}

func (s stufenWorkspaces) Get(ctx context.Context, id string) (model.Workspace, error) {
	w, err := s.WorkspaceRepository.Get(ctx, id)
	s.stufe()
	return w, err
}

// fahreMutationen führt die Läufe aus: je Goroutine laeufe Runden mit den vier Methoden; eine Schranke an Workspaces.Get (S2, vor dem Beginn
// des Spans), den Ports vor der Transaktion und am Eingang der Transaktion hält alle 16 Aufrufe zugleich an jeder dieser Stellen fest, bevor
// einer weitergeht; hinter dem Beginn sind so alle 16 Spans zugleich offen. Die Ports in der Transaktion sind keine Stufen: die Transaktion
// läuft hintereinander. Ist abbr nicht nil, bricht jeder Aufruf jeder zweiten Goroutine (i gerade) in S22 ab: die Goroutine hinterlegt die
// Funktion, die den ctx ihres Aufrufs beendet, in abbr. Nicht erreichte Schranke und unerwartete Ausgänge sind Fehler.
func fahreMutationen(t *testing.T, u *umg, laeufe int, abbr *abbrueche) {
	t.Helper()
	abbrechen := func(i int) bool { return abbr != nil && i%2 == 0 }
	schranke := neuerZyklus(mutGoroutinen)
	var gescheitert atomic.Bool
	var verpasst sync.Map
	stufe := func() {
		if !gescheitert.Load() && !schranke.warte(10*time.Second) {
			gescheitert.Store(true)
			verpasst.Store("Schranke nicht erreicht", true)
		}
	}
	stufen := []string{"Metadaten.Get", "Blobs.Open", "MetadatenCodec.ParseMetadata", "Transaktion.WithWorkspaceTransaction"}
	u.p.haken = func(op string) {
		if slices.Contains(stufen, op) {
			stufe()
		}
	}
	u.k.Workspaces = stufenWorkspaces{WorkspaceRepository: u.ws, stufe: stufe}
	m := u.mocker()
	var wg sync.WaitGroup
	for i := range mutGoroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ws := "w" + strconv.Itoa(i)
			for r := range laeufe {
				for _, methode := range mutationsFolge(i) {
					ctx, cancel := context.WithCancel(u.ctx)
					if abbrechen(i) {
						abbr.setze(ws, cancel)
					}
					antwort, err := m.Mock(ctx, dynAnfrage(methode, mutationsSchluessel(methode, r), "n", ws))
					cancel()
					switch {
					case abbrechen(i) && (!errors.Is(err, context.Canceled) || fehlercode(err) != 0):
						verpasst.Store("Goroutine "+strconv.Itoa(i)+" "+methode+": Fehler "+errText(err)+", erwartet Canceled ohne Code", true)
					case !abbrechen(i) && err != nil:
						verpasst.Store("Goroutine "+strconv.Itoa(i)+" "+methode+": "+errText(err), true)
					case err == nil:
						_ = antwort.Body.Close()
					}
				}
			}
		}()
	}
	wg.Wait()
	verpasst.Range(func(k, _ any) bool { t.Errorf("%v", k); return true })
}

func errText(err error) string {
	if err == nil {
		return "kein Fehler"
	}
	return err.Error()
}

// abbrueche hält je Workspace die Funktion, die den ctx des laufenden Aufrufs beendet; den Eintrag eines Workspace setzt die Goroutine
// vor ihrem Aufruf, und putAbbruch liest ihn in der Transaktion dieses Aufrufs.
type abbrueche struct {
	mu sync.Mutex
	m  map[string]context.CancelFunc
}

func (a *abbrueche) setze(ws string, f context.CancelFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.m[ws] = f
}

func (a *abbrueche) hole(ws string) context.CancelFunc {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.m[ws]
}

// TestMutationAbbruchGleichzeitig hält M6 als Wirkungs-Fänger (SPEC-290 Nr. 3 f, 9 b; AGENTS.md §3.10): 16 Goroutinen × 4 Methoden × 4 Runden,
// in jeder zweiten Goroutine (acht) bricht jede Mutation in S22 ab. Je Aufruf endet der Span genau einmal mit seiner Art und Kennung; der
// Entity-State und die Revision der abgebrochenen Workspaces sind unverändert, die der übrigen um 4 je Runde (Revision) und 2 je Runde
// (Version des Eintrags 10) gewachsen. Die Schranke von fahreMutationen hält alle 16 Spans zugleich offen.
func TestMutationAbbruchGleichzeitig(t *testing.T) {
	const laeufe = 4
	u := gleichzeitigUmg(t, laeufe)
	abbr := &abbrueche{m: map[string]context.CancelFunc{}}
	u.tx.innen = func(s outbound.WorkspaceStores) outbound.WorkspaceStores {
		s.Entities = putAbbruch{EntityRepository: s.Entities, cancel: func(ws string) context.CancelFunc {
			if f := abbr.hole(ws); f != nil {
				return f
			}
			return func() {}
		}}
		return s
	}
	vorher := map[string][]model.EntityState{}
	for i := range mutGoroutinen {
		ws := "w" + strconv.Itoa(i)
		vorher[ws] = u.en.ausWs(ws, setProducts)
	}
	fahreMutationen(t, u, laeufe, abbr)
	for i := range mutGoroutinen {
		ws := "w" + strconv.Itoa(i)
		nachher := u.en.ausWs(ws, setProducts)
		rev := u.ws.revision(ws)
		if i%2 == 0 {
			if rev != 0 || !slices.EqualFunc(vorher[ws], nachher, gleicherEintrag) {
				t.Errorf("%s (abgebrochen): Revision %d, Zustand verändert %v; erwartet 0 und unverändert", ws, rev, !slices.EqualFunc(vorher[ws], nachher, gleicherEintrag))
			}
			continue
		}
		var zehn model.EntityState
		for _, x := range nachher {
			if x.EntityKey == "10" {
				zehn = x
			}
		}
		if rev != 4*laeufe || zehn.Version != 1+2*laeufe {
			t.Errorf("%s: Revision %d, Version von 10 gleich %d; erwartet %d und %d", ws, rev, zehn.Version, 4*laeufe, 1+2*laeufe)
		}
	}
	erwarteBuchProAufruf(t, u, laeufe)
}

// gleicherEintrag vergleicht zwei Einträge in ihrem Zustand.
func gleicherEintrag(a, b model.EntityState) bool {
	return a.EntityKey == b.EntityKey && a.Payload == b.Payload && a.Version == b.Version && a.Deleted == b.Deleted &&
		a.Completeness == b.Completeness && slices.Equal(a.KnownProperties, b.KnownProperties)
}

// erwarteBuchProAufruf prüft das Buch des Fakes nach einem Lauf von fahreMutationen: je Aufruf ein Span, in der Folge der Aufrufe je Workspace
// die rotierte Folge der vier Arten laeufe-mal, jeder mit Kennung seiner Anfrage, genau einem End und ohne Set.
func erwarteBuchProAufruf(t *testing.T, u *umg, laeufe int) {
	t.Helper()
	buch := buecher(u.tel.FakeTelemetry)
	if len(buch) != mutGoroutinen*4*laeufe {
		t.Fatalf("%d Spans, erwartet %d", len(buch), mutGoroutinen*4*laeufe)
	}
	jeWs := map[string][]spanBuch{}
	for _, b := range buch {
		ws, _ := b.wert(attrWsID)
		jeWs[ws] = append(jeWs[ws], b)
		if b.name != support.SpanWorkspaceMutation || b.enden != 1 || b.sets != 0 || len(b.attrs) != 2 {
			t.Errorf("Span %+v: Name, Zahl der End oder der Attribute", b)
		}
	}
	art := map[string]string{"POST": "post", "PATCH": "patch", "PUT": "put", "DELETE": "delete"}
	for i := range mutGoroutinen {
		ws := "w" + strconv.Itoa(i)
		var soll, got []string
		for range laeufe {
			for _, m := range mutationsFolge(i) {
				soll = append(soll, art[m])
			}
		}
		for _, b := range jeWs[ws] {
			a, _ := b.wert(attrArt)
			got = append(got, a)
		}
		if !slices.Equal(got, soll) {
			t.Errorf("%s: Arten %v, erwartet %v", ws, got, soll)
		}
	}
	if len(jeWs) != mutGoroutinen {
		t.Errorf("%d Workspaces im Buch, erwartet %d", len(jeWs), mutGoroutinen)
	}
}

// TestMutationSpanGleichzeitig hält M10 (SPEC-290 Nr. 9 a; AGENTS.md §3.10): n gleichzeitige Mutationen ergeben n Spans, jeder mit Art und
// Kennung seiner Anfrage und genau einem End. 16 Goroutinen × 4 Methoden × 8 Runden = 512 Spans in 16 Workspaces; die Schranke von
// fahreMutationen (Stufen: Workspaces.Get, Metadaten.Get, Blobs.Open, ParseMetadata, Eingang der Transaktion; der Leser S17 ist keine Stufe)
// hält alle 16 Spans zugleich offen, und weil die Folge der Methoden je Goroutine rotiert, sind dabei alle vier Arten offen
// (Wirkungs-Fänger: das Buch je Span, kein Datenrennen nötig). Jede Mutation gelingt.
func TestMutationSpanGleichzeitig(t *testing.T) {
	const laeufe = 8
	u := gleichzeitigUmg(t, laeufe)
	fahreMutationen(t, u, laeufe, nil)
	erwarteBuchProAufruf(t, u, laeufe)
	if got := u.p.anzahl("Entities.Put"); got != mutGoroutinen*4*laeufe {
		t.Errorf("%d Put, erwartet %d", got, mutGoroutinen*4*laeufe)
	}
	if len(mutationsFolge(0)) != 4 || mutGoroutinen != 16 {
		t.Errorf("Folge %d, Goroutinen %d; erwartet 4 und 16", len(mutationsFolge(0)), mutGoroutinen)
	}
}

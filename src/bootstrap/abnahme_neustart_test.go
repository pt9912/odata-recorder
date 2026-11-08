package bootstrap_test

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Neustart (SPEC-049 Nr. 7, LH-QA-03): im Prozess des Tests werden Datenbank und Dienste geschlossen und neu gebaut; abgeschlossene Mutationen und
// Snapshots sind danach vorhanden, Fehlversuche haben nichts hinterlassen.

// abNeustartAnfragen sind die vier Anfragen, deren Antworten vor und nach dem Neustart byte-gleich sind.
func abNeustartAnfragen() []abAnfrage {
	return []abAnfrage{{Methode: "GET", Pfad: "/Products"}, {Methode: "GET", Pfad: "/Products(3)"}, {Methode: "GET", Pfad: "/Products(10)"},
		{Methode: "GET", Pfad: "/Products/$count"}}
}

func abNeustartAntworten(u *abUmgebung) []abIst {
	var l []abIst
	for _, a := range abNeustartAnfragen() {
		i := u.sende(a)
		i.Antwort = inbound.Antwort{}
		l = append(l, i)
	}
	return l
}

func abListe(t *testing.T, u *abUmgebung, ws string) []model.WorkspaceSnapshot {
	t.Helper()
	l, err := u.snap.List(context.Background(), ws)
	abMust(t, err)
	return l
}

// TestAbnahmeNeustart (T63 bis T70): die Schritte 1 bis 9 von Nr. 7 c, dann der Neustart und die Zusagen (1) bis (6) von Nr. 7 d.
func TestAbnahmeNeustart(t *testing.T) {
	ctx := context.Background()
	u := abFrisch(t)
	u.neuerWorkspace("ws-2")
	u.cursor("ws-1", 3)
	u.cursor("ws-2", 2)
	erwarte := func(was string, ist abIst, status int, code model.Code) {
		t.Helper()
		if code != 0 {
			if ist.F == nil || ist.F.Code != code || ist.F.Status != status {
				t.Fatalf("%s: %v (Status %d), erwartet %s mit %d", was, ist.Err, ist.Status, code, status)
			}
			return
		}
		if ist.F != nil || ist.Err != nil || ist.Status != status {
			t.Fatalf("%s: Status %d, Fehler %v, erwartet %d", was, ist.Status, ist.Err, status)
		}
	}
	erwarte("1 POST", u.sende(abAnfrage{Methode: "POST", Pfad: "/Products", Body: `{"Id":10,"Name":"Eyelet"}`}), 201, 0)
	erwarte("2 PATCH", u.sende(abAnfrage{Methode: "PATCH", Pfad: "/Products(1)", Body: `{"Name":"Anvil2"}`}), 204, 0)
	erwarte("3 PUT", u.sende(abAnfrage{Methode: "PUT", Pfad: "/Products(2)", Body: `{"Id":2,"Name":"Bolt2","Active":true}`}), 204, 0)
	erwarte("4 DELETE", u.sende(abAnfrage{Methode: "DELETE", Pfad: "/Products(3)"}), 204, 0)
	nach4 := u.stand("ws-1")
	// Fehlversuche (Nr. 7 b): Schritt 5 mit Fehler, Schritt 6 mit beendetem Kontext; sie hinterlassen nichts.
	erwarte("5 POST doppelt", u.sende(abAnfrage{Methode: "POST", Pfad: "/Products", Body: `{"Id":1}`}), 409, model.CodeE6009)
	beendet, abbruch := context.WithCancel(ctx)
	abbruch()
	ist6 := abSende(beendet, u.mock, "ws-1", abAnfrage{Methode: "PATCH", Pfad: "/Products(4)", Body: `{"Name":"X"}`})
	var f6 *inbound.Fehler
	if !errors.Is(ist6.Err, context.Canceled) || errors.As(ist6.Err, &f6) {
		t.Fatalf("6 PATCH mit beendetem Kontext: %v, erwartet ein Kontextfehler ohne Code", ist6.Err)
	}
	for _, d := range abStandGleich(u.stand("ws-1"), nach4) {
		t.Errorf("Fehlversuche haben den Stand verändert: %s", d)
	}
	if nach4.Revision != 4 {
		t.Fatalf("Revision nach Schritt 4: %d, erwartet 4", nach4.Revision)
	}
	vor, err := u.snap.Create(ctx, "ws-1", "vor")
	abMust(t, err)
	erwarte("8 PATCH", u.sende(abAnfrage{Methode: "PATCH", Pfad: "/Products(5)", Body: `{"Name":"Late"}`}), 204, 0)
	nach, err := u.snap.Create(ctx, "ws-1", "nach")
	abMust(t, err)
	if vor.Name != "vor" || nach.Name != "nach" {
		t.Fatalf("Snapshots %q und %q", vor.Name, nach.Name)
	}
	// Festhalten.
	stand1, stand2 := u.stand("ws-1"), u.stand("ws-2")
	liste1 := abListe(t, u, "ws-1")
	antworten := abNeustartAntworten(u)
	if stand1.Revision != 5 || stand2.Revision != 0 || len(liste1) != 2 {
		t.Fatalf("vor dem Neustart: Revision %d und %d, %d Snapshots; erwartet 5, 0, 2", stand1.Revision, stand2.Revision, len(liste1))
	}

	// Neustart (T63): neue Instanzen, die alten sind geschlossen und nicht mehr benutzbar.
	alt := u
	u = u.neustart()
	if u == alt || u.db == alt.db || u.ws == alt.ws || u.snap == alt.snap || u.mock == alt.mock {
		t.Error("nach dem Neustart steht eine Instanz der alten Umgebung")
	}
	if _, err := alt.db.Workspaces().Get(ctx, "ws-1"); err == nil {
		t.Error("die alte Datenbank ist nach dem Neustart benutzbar")
	}
	if _, err := abLeseStand(ctx, alt.db, "ws-1"); err == nil {
		t.Error("die alte Datenbank liefert einen Stand")
	}

	// (1) Stand gleich, Version, Revisionen.
	nStand1, nStand2 := u.stand("ws-1"), u.stand("ws-2")
	for _, d := range abStandGleich(nStand1, stand1) {
		t.Errorf("(1) ws-1 nach dem Neustart: %s", d)
	}
	for _, d := range abStandGleich(nStand2, stand2) {
		t.Errorf("(1) ws-2 nach dem Neustart: %s", d)
	}
	if p1 := eintrag(nStand1, "Products", "1"); p1.Version != 1 || nStand1.Revision != 5 || nStand2.Revision != 0 {
		t.Errorf("(1) Version von Products 1: %d (erwartet 1), Revisionen %d und %d (erwartet 5 und 0)", p1.Version, nStand1.Revision, nStand2.Revision)
	}
	// (2) Antworten byte-gleich.
	neueAntworten := abNeustartAntworten(u)
	if len(neueAntworten) != 4 || !reflect.DeepEqual(neueAntworten, antworten) {
		t.Errorf("(2) Antworten nach dem Neustart weichen ab:\n neu %+v\n alt %+v", neueAntworten, antworten)
	}
	if antworten[1].F == nil || antworten[1].F.Code != model.CodeE6008 || antworten[2].Status != 200 || antworten[3].Body != "5" {
		t.Errorf("(2) die festgehaltenen Antworten sind nicht die erwarteten: %+v", antworten)
	}
	// (3) Snapshots.
	if l := abListe(t, u, "ws-1"); !reflect.DeepEqual(l, liste1) {
		t.Errorf("(3) Snapshots von ws-1 nach dem Neustart:\n neu %+v\n alt %+v", l, liste1)
	}
	if l := abListe(t, u, "ws-2"); len(l) != 0 {
		t.Errorf("(3) ws-2 hat %d Snapshots, erwartet keinen", len(l))
	}
	// (4) Restore von vor.
	w, err := u.ws.Restore(ctx, "ws-1", vor.ID)
	abMust(t, err)
	r := u.stand("ws-1")
	if w.Revision != 6 || r.Revision != 6 {
		t.Errorf("(4) Revision nach Restore %d/%d, erwartet 6", w.Revision, r.Revision)
	}
	if !reflect.DeepEqual(r.Eintraege, nach4.Eintraege) {
		t.Errorf("(4) State nach Restore ungleich dem State nach Schritt 4")
	}
	if p5 := eintrag(r, "Products", "5"); p5.Payload != abPayload(5, nil) || eintrag(r, "Products", "10").EntityKey != "10" || !eintrag(r, "Products", "3").Deleted {
		t.Errorf("(4) Products 5 %q, Products 10 vorhanden: %v, Tombstone 3: %v", p5.Payload, eintrag(r, "Products", "10").EntityKey == "10", eintrag(r, "Products", "3").Deleted)
	}
	if !reflect.DeepEqual(r.Cursors, nStand1.Cursors) || len(r.Cursors) != 1 || r.Cursors[0].Position != 3 {
		t.Errorf("(4) Cursors nach Restore %+v, erwartet unverändert %+v", r.Cursors, nStand1.Cursors)
	}
	// (5) die Zähler laufen weiter.
	erwarte("(5) PATCH", u.sende(abAnfrage{Methode: "PATCH", Pfad: "/Products(1)", Body: `{"Name":"Anvil3"}`}), 204, 0)
	nach5 := u.stand("ws-1")
	if p1 := eintrag(nach5, "Products", "1"); p1.Version != 2 || nach5.Revision != 7 {
		t.Errorf("(5) Version %d, Revision %d; erwartet 2 und 7", p1.Version, nach5.Revision)
	}
	// (6) ein zweiter Neustart ohne Änderung.
	u = u.neustart()
	for _, d := range abStandGleich(u.stand("ws-1"), nach5) {
		t.Errorf("(6) ws-1 nach dem zweiten Neustart: %s", d)
	}
	if l := abListe(t, u, "ws-1"); !reflect.DeepEqual(l, liste1) {
		t.Errorf("(6) Snapshots nach dem zweiten Neustart:\n neu %+v\n alt %+v", l, liste1)
	}
}

func eintrag(s abStand, set, key string) model.EntityState {
	for _, e := range s.Eintraege {
		if e.EntitySet == set && e.EntityKey == key {
			return e
		}
	}
	return model.EntityState{}
}

// abSucheSkip nennt die Aufrufe von Skip, SkipNow, Skipf an einem Wert und von testing.Short im Quelltext src.
func abSucheSkip(dateiname string, src []byte) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), dateiname, src, 0)
	if err != nil {
		return nil, err
	}
	var funde []string
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			switch sel.Sel.Name {
			case "Skip", "SkipNow", "Skipf":
				funde = append(funde, sel.Sel.Name)
			case "Short":
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == "testing" {
					funde = append(funde, "testing.Short")
				}
			}
		}
		return true
	})
	return funde, nil
}

// TestAbnahmeKeinSkip (T43): kein t.Skip und kein testing.Short in den Dateien der Abnahme (SPEC-049 Nr. 9 a).
func TestAbnahmeKeinSkip(t *testing.T) {
	for _, quelle := range []string{"package x\nfunc f(t T) { t.Skip() }", "package x\nfunc f(t T) { t.SkipNow() }", "package x\nfunc f(t T) { t.Skipf(\"\") }",
		"package x\nfunc f() bool { return testing.Short() }"} {
		if funde, err := abSucheSkip("x.go", []byte(quelle)); err != nil || len(funde) != 1 {
			t.Errorf("Suche in %q: %v (%v), erwartet ein Fund", quelle, funde, err)
		}
	}
	dateien, err := filepath.Glob("abnahme_*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, soll := range []string{"abnahme_aufbau_test.go", "abnahme_matrix_test.go", "abnahme_determinismus_test.go", "abnahme_isolation_test.go",
		"abnahme_neustart_test.go", "abnahme_nebenlaeufig_test.go"} {
		if !slices.Contains(dateien, soll) {
			t.Errorf("%s fehlt unter den Dateien der Abnahme", soll)
		}
	}
	for _, d := range dateien {
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		if funde, err := abSucheSkip(d, src); err != nil || len(funde) != 0 {
			t.Errorf("%s: %v (%v)", d, funde, err)
		}
	}
}

package workspace_test

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestRestoreFensterLesen (T60): ein Schreiber, der zwischen Lesen und Transaktion committet (Mutation, Reset, zweites
// Restore, Fortschreiben eines Cursors), ändert nichts: der State ist der des Snapshots, die Revision die in der
// Transaktion gelesene plus 1; ein Löschen dazwischen ist ODR-E7001 und schreibt nichts. Das Ereignis läuft synchron
// im Haltepunkt nach Blobs.Open, den der Test dem Fake übergibt (Ereignis 5).
func TestRestoreFensterLesen(t *testing.T) {
	ctx := context.Background()
	n := 0
	for _, ev := range []string{"Mutation", "Reset", "Restore", "Cursor", "Loeschen"} {
		w := restoreWelt(t)
		andere := snapshotState("ws-1")
		andere[0].Payload = `{"Id":"1","Name":"Zweiter"}`
		w.snapshotAblegen("ws-1", "snp-2", andere)
		advance := map[string]int64{"Mutation": 1, "Reset": 1, "Restore": 1}[ev]
		var fehler error
		w.p.nach = func(op string) {
			if op != "Blobs.Open" {
				return
			}
			w.p.nach = nil
			switch ev {
			case "Mutation":
				fehler = entRepo{w.st, false}.Put(ctx, model.EntityState{WorkspaceID: "ws-1", EntitySet: "Products", EntityKey: "Id=77",
					Payload: `{"Id":"77"}`, Completeness: model.EntityComplete, Version: 1})
				if fehler == nil {
					_, fehler = wsRepo{w.st, false}.Advance(ctx, "ws-1", 5)
				}
			case "Reset":
				_, fehler = w.v.Reset(ctx, "ws-1")
			case "Restore":
				_, fehler = w.v.Restore(ctx, "ws-1", "snp-2")
			case "Cursor":
				fehler = curRepo{w.st, false}.Put(ctx, model.ReplayCursor{WorkspaceID: "ws-1", ReplaySessionID: "rep-1", RecordingSessionID: "ses-1",
					ServiceID: "svc", Strategy: model.StrategySequential, Position: 9, Consumed: []int64{1, 2, 3}})
			default:
				fehler = wsRepo{w.st, false}.Delete(ctx, "ws-1")
			}
		}
		cursorsVorher := slices.Clone(w.st.stand().cur["ws-1"])
		got, err := restoreLauf(w)
		if fehler != nil {
			t.Fatalf("%s: Ereignis: %v", ev, fehler)
		}
		z := w.st.stand()
		if ev == "Loeschen" {
			pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
			if len(w.p.von("tx Entities.ReplaceAll")) != 0 || len(z.ent["ws-1"]) != 0 {
				t.Errorf("Löschen: ReplaceAll oder State vorhanden")
			}
			n++
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", ev, err)
		}
		if !equalState(z.ent["ws-1"], snapshotState("ws-1")) {
			t.Errorf("%s: State %v", ev, z.ent["ws-1"])
		}
		if got.Revision != 5+advance+1 || z.ws["ws-1"].Revision != 5+advance+1 {
			t.Errorf("%s: Revision %d, erwartet %d", ev, got.Revision, 5+advance+1)
		}
		advances := w.p.von("tx Workspaces.Advance")
		if a := advances[len(advances)-1].args[1]; a != 5+advance {
			t.Errorf("%s: Advance mit %v, erwartet die in der Transaktion gelesene %d", ev, a, 5+advance)
		}
		if ev == "Cursor" && (len(z.cur["ws-1"]) != 2 || z.cur["ws-1"][0].Position != 9) {
			t.Errorf("Cursor: Stand nach Restore %v", z.cur["ws-1"])
		}
		if ev != "Cursor" && ev != "Reset" && !reflectGleich(cursorsVorher, z.cur["ws-1"]) {
			t.Errorf("%s: Cursors geändert", ev)
		}
		n++
	}
	if n != 5 {
		t.Errorf("M13: %d Ereignisse", n)
	}
}

func reflectGleich(a, b []model.ReplayCursor) bool {
	return len(a) == len(b) && slices.EqualFunc(a, b, func(x, y model.ReplayCursor) bool {
		return x.WorkspaceID == y.WorkspaceID && x.ReplaySessionID == y.ReplaySessionID && x.RecordingSessionID == y.RecordingSessionID &&
			x.Position == y.Position && slices.Equal(x.Consumed, y.Consumed) && x.Strategy == y.Strategy
	})
}

// TestRestoreFensterTransaktion (T61): kein anderer Schreiber läuft zwischen (1) und (2) und zwischen (2) und (3); es
// gibt genau eine Transaktion. Ein Eindringling, der im Haltepunkt eingereiht wird, tritt erst nach Advance ein (das
// Protokoll nennt seine Stelle), und in einem Wartefenster von 50 ms bleibt er draußen.
func TestRestoreFensterTransaktion(t *testing.T) {
	for _, punkt := range []string{"tx Workspaces.Get", "tx Entities.ReplaceAll"} {
		w := restoreWelt(t)
		var eingetreten atomic.Bool
		var ch <-chan error
		var imFenster bool
		w.p.nach = func(op string) {
			if op != punkt {
				return
			}
			w.p.nach = nil
			ch = w.st.eindringling(context.Background(), func(outbound.WorkspaceStores) error {
				eingetreten.Store(true)
				_ = w.p.beginn("Eindringling", nil)
				return nil
			})
			time.Sleep(50 * time.Millisecond)
			imFenster = eingetreten.Load()
		}
		if _, err := restoreLauf(w); err != nil {
			t.Fatal(err)
		}
		if err := <-ch; err != nil || !eingetreten.Load() {
			t.Fatalf("%s: Eindringling lief nicht: %v", punkt, err)
		}
		if imFenster {
			t.Errorf("%s: der Eindringling trat im Fenster ein", punkt)
		}
		if w.p.anzahl(lblTx) != 1 {
			t.Errorf("%s: %d Transaktionen", punkt, w.p.anzahl(lblTx))
		}
		ops := w.p.ops()
		if slices.Index(ops, "Eindringling") < slices.Index(ops, "tx Workspaces.Advance") {
			t.Errorf("%s: der Eindringling lief vor Advance: %v", punkt, ops)
		}
	}
}

// zweig ist ein Zweig der Z-Liste von Restore: der Aufbau, die Kennung des Snapshots und die erwartete Antwort.
type zweig struct {
	id    string
	pre   func(w *welt, abbruch context.CancelFunc)
	snap  string
	soll  string // Code, "ok", "kontext" oder "panik"
	close int
}

func zweigeRestore() []zweig {
	hashFalsch := func(w *welt) {
		s := w.snaps.wsSnaps[[2]string{"ws-1", "snp-1"}]
		s.StateHash = "sha256:" + strings.Repeat("e", 64)
		w.snaps.wsSnaps[[2]string{"ws-1", "snp-1"}] = s
	}
	inj := func(op string, err error) func(w *welt, a context.CancelFunc) {
		return func(w *welt, _ context.CancelFunc) { w.p.setzeFehler(op, err) }
	}
	return []zweig{
		{"Z16", inj("Workspaces.Get", errPlatte), "snp-1", "ODR-E9001", 0},
		{"Z17", nil, "unbekannt", "ODR-E7002", 0},
		{"Z18", inj("Snapshots.GetWorkspace", errPlatte), "snp-1", "ODR-E9001", 0},
		{"Z19", func(w *welt, _ context.CancelFunc) { w.baseline("bsl-x", "svc2", standardEnts()) }, "bsl-x", "ODR-E5004", 0},
		{"Z20", nil, "bsl-1", "ODR-E7008", 0},
		{"Z21", inj("Snapshots.Get", errPlatte), "bsl-1", "ODR-E9001", 0},
		{"Z22", func(w *welt, _ context.CancelFunc) { hashFalsch(w) }, "snp-1", "ODR-E3001", 0},
		{"Z23", inj("Blobs.Open", errPlatte), "snp-1", "ODR-E9001", 0},
		{"Z24", func(w *welt, _ context.CancelFunc) { w.blobs.fehlerBei, w.blobs.lesefehler = 10, errPlatte }, "snp-1", "ODR-E9001", 1},
		{"Z25", func(w *welt, _ context.CancelFunc) { w.blobs.closeFehler = errPlatte }, "snp-1", "ok", 1},
		{"Z26", func(w *welt, _ context.CancelFunc) { w.snapshotBytesAblegen("ws-1", "snp-1", []byte("kein json")) }, "snp-1", "ODR-E3001", 1},
		{"Z27", func(w *welt, _ context.CancelFunc) {
			w.p.nach = func(op string) {
				if op == "Blobs.Open" {
					w.p.nach = nil
					_ = wsRepo{w.st, false}.Delete(context.Background(), "ws-1")
				}
			}
		}, "snp-1", "ODR-E7001", 1},
		{"Z28", inj("tx Entities.ReplaceAll", errPlatte), "snp-1", "ODR-E9001", 1},
		{"Z29", inj("tx Workspaces.Advance", errPlatte), "snp-1", "ODR-E9001", 1},
		{"Z30", inj(lblTx, errPlatte), "snp-1", "ODR-E9001", 1},
		{"Z31", func(w *welt, _ context.CancelFunc) { w.p.setzePanik("tx Entities.ReplaceAll", "peng") }, "snp-1", "panik", 1},
		{"Z32", func(w *welt, abbruch context.CancelFunc) {
			w.p.nach = func(op string) {
				if op == "Blobs.Open" {
					w.p.nach = nil
					abbruch()
				}
			}
		}, "snp-1", "kontext", 1},
	}
}

// lauf führt Restore für den Zweig z in einer frischen Welt aus und liefert die Antwort als Text und die Zahl der Close.
func (z zweig) lauf(t *testing.T) (string, int) {
	w := restoreWelt(t)
	ctx, abbruch := context.WithCancel(context.Background())
	defer abbruch()
	if z.pre != nil {
		z.pre(w, abbruch)
	}
	return z.antwort(ctx, w)
}

func (z zweig) antwort(ctx context.Context, w *welt) (string, int) {
	var antwort string
	func() {
		defer func() {
			if r := recover(); r != nil {
				antwort = "panik"
			}
		}()
		_, err := w.v.Restore(ctx, "ws-1", z.snap)
		switch f, ok := codeVon(err); {
		case err == nil:
			antwort = "ok"
		case ok:
			antwort = f.Code.String()
		case errors.Is(err, context.Canceled):
			antwort = "kontext"
		default:
			antwort = err.Error()
		}
	}()
	closes := 0
	for _, s := range w.blobs.strome() {
		closes += s.geschlossen()
	}
	return antwort, closes
}

// TestZweigeRestore (T72): jeder Zweig der Z-Liste (Z16 bis Z32) läuft in vier Runden zu 64 Goroutinen gleichzeitig; die
// Antwort gleicht der des Einzelaufrufs und dem Literal der Liste, der Strom wird so oft geschlossen wie gesagt. Die
// Welten entstehen vor dem Start. Das Abdeckungsprofil unter dem Detektor liegt im Bericht.
func TestZweigeRestore(t *testing.T) {
	zs := zweigeRestore()
	if len(zs) != 17 {
		t.Fatalf("Zählprüfung: %d Zweige", len(zs))
	}
	// Z-Liste gegen die Quelle: wiederherstellen.go hat 16 Rückgaben mit Fehlerwert: Restore 9 (ctx nil, S10, S19, S20 über
	// die Hilfe, Manifest, f mit S11, S9, S13, Transaktion), baselineStattSnapshot 3 (S20, ODR-E5004, ODR-E7008) und
	// snapshotManifest 4 (Hash, S4, S5, Manifest). Z16 bis Z32 verteilen sich darauf; Z25 (Close) und Z31 (Panik) haben keine
	// Rückgabe, Z32 ist der Kontextfehler einer Stelle.
	if n := fehlerRueckgaben(t, "wiederherstellen.go"); n != 16 {
		t.Errorf("%d Rückgaben mit Fehlerwert in wiederherstellen.go, erwartet 16", n)
	}
	// Z26 (Verstoß des Manifests) gegen die Quelle: manifest_snapshot.go ruft verstoss an 11 Stellen und hat 16 Rückgaben mit
	// Fehlerwert; eine weitere Stelle ist eine weitere Regel von SPEC-129 Nr. 20 und braucht einen Fall in M6.
	if n, r := rufe(t, "manifest_snapshot.go", "verstoss"), fehlerRueckgaben(t, "manifest_snapshot.go"); n != 11 || r != 24 {
		t.Errorf("manifest_snapshot.go: %d Aufrufe von verstoss (erwartet 11), %d Rückgaben mit Fehlerwert (erwartet 24)", n, r)
	}
	for _, z := range zs {
		einzel, closes := z.lauf(t)
		if einzel != z.soll || closes != z.close {
			t.Errorf("%s: Einzelaufruf %q mit %d Close, erwartet %q mit %d", z.id, einzel, closes, z.soll, z.close)
		}
		for runde := 0; runde < 4; runde++ {
			welten := make([]*welt, 64)
			ctxs := make([]context.Context, 64)
			for i := range welten {
				w := restoreWelt(t)
				ctx, abbruch := context.WithCancel(context.Background())
				t.Cleanup(abbruch)
				if z.pre != nil {
					z.pre(w, abbruch)
				}
				welten[i], ctxs[i] = w, ctx
			}
			antworten := make([]string, 64)
			gleichzeitig(64, func(i int) { antworten[i], _ = z.antwort(ctxs[i], welten[i]) })
			for i, a := range antworten {
				if a != einzel {
					t.Errorf("%s Runde %d Lauf %d: %q, Einzelaufruf %q", z.id, runde, i, a, einzel)
				}
			}
		}
	}
}

// rufe zählt in der Datei die Aufrufe der Funktion name.
func rufe(t *testing.T, datei, name string) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ast.Inspect(f, func(x ast.Node) bool {
		if c, ok := x.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == name {
				n++
			}
		}
		return true
	})
	return n
}

// fehlerRueckgaben zählt in der Datei die return-Anweisungen, deren letztes Ergebnis ein Aufruf, ein Verweis auf err oder
// ein Adresswert ist (jede Rückgabe mit Fehlerwert).
func fehlerRueckgaben(t *testing.T, datei string) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ast.Inspect(f, func(x ast.Node) bool {
		if r, ok := x.(*ast.ReturnStmt); ok && len(r.Results) > 0 {
			switch l := r.Results[len(r.Results)-1].(type) {
			case *ast.CallExpr, *ast.UnaryExpr:
				n++
			case *ast.Ident:
				if l.Name == "err" {
					n++
				}
			}
		}
		return true
	})
	return n
}

// aufrufstellen zählt in der Datei datei die Aufrufe der Formen Wurzel.Port.Operation (v.k.Blobs.Open) und
// Wurzel.Operation (rc.Close, io.ReadAll, v.transaktion); die Wurzel s steht für die Stores der Transaktion und trägt
// das Präfix "tx ".
func aufrufstellen(t *testing.T, datei string) map[string]int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	r := map[string]int{}
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch x := sel.X.(type) {
		case *ast.SelectorExpr: // v.k.Blobs / s.Workspaces
			switch y := x.X.(type) {
			case *ast.SelectorExpr: // v.k
				r[x.Sel.Name+"."+sel.Sel.Name]++
			case *ast.Ident:
				if y.Name == "s" {
					r["tx "+x.Sel.Name+"."+sel.Sel.Name]++
				}
			}
		case *ast.Ident:
			r[x.Name+"."+sel.Sel.Name]++
		}
		return true
	})
	return r
}

// TestStellenRestore (T74): jede Aufrufstelle eines Ports in wiederherstellen.go steht mit Zahl in der Liste von
// SPEC-129 Nr. 15 (S4, S5, S6, S7, S9, S10, S11, S13, S19, S20); es gibt weder eine Stelle von Cursors noch ein
// Entities.Put, Snapshots.CreateWorkspace oder Delete.
func TestStellenRestore(t *testing.T) {
	// S10, S19, S20, S4, S5, S6, S7, S11, S9, S13 und die beiden Hilfen, die keinen Port rufen.
	soll := map[string]int{
		"Workspaces.Get": 1, "Snapshots.GetWorkspace": 1, "Snapshots.Get": 1,
		"Blobs.Open": 1, "io.ReadAll": 1, "rc.Close": 1,
		"v.transaktion":     1,
		"tx Workspaces.Get": 1, "tx Entities.ReplaceAll": 1, "tx Workspaces.Advance": 1,
		"v.baselineStattSnapshot": 1, "v.snapshotManifest": 1,
	}
	for _, datei := range []string{"wiederherstellen.go"} {
		got := aufrufstellen(t, datei)
		for k := range got {
			if strings.Contains(k, "Cursors") {
				t.Errorf("Aufrufstelle %s", k)
			}
		}
		for k, n := range soll {
			if got[k] != n {
				t.Errorf("%s: %d Aufrufstellen, erwartet %d", k, got[k], n)
			}
		}
		for _, verboten := range []string{"Entities.Put", "tx Entities.Put", "Snapshots.CreateWorkspace", "Workspaces.Delete", "Workspaces.Create", "tx Workspaces.Create"} {
			if got[verboten] != 0 {
				t.Errorf("verbotene Aufrufstelle %s", verboten)
			}
		}
	}
	// Die neun Paare mit Fehlerrückgabe von Restore (S4, S5, S7, S9, S10, S11, S13, S19, S20) stehen in stellen().
	if n := len(restoreStellen()); n != 9 {
		t.Errorf("Stellen von Restore %d, erwartet 9", n)
	}
}

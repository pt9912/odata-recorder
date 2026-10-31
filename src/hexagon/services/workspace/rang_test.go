package workspace_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// aufruf sind die Argumente eines Aufrufs von Create oder Reset.
type aufruf struct {
	ctx        ctxMethoden
	baselineID string
	id         string
}

// verstossFall ist ein Verstoß gegen den Ablauf von Create oder Reset (Zusatztabelle in §6 des Plans):
// rang ordnet ihn in der Folge der Schritte, apply stellt ihn her, pruefe hält fest, was ihn von den
// anderen unterscheidet (Code, Ursache, Zähler der Fakes).
type verstossFall struct {
	name   string
	rang   int
	apply  func(w *welt, a *aufruf)
	pruefe func(t *testing.T, w *welt, err error)
}

func ohnePort(t *testing.T, w *welt, err error) {
	t.Helper()
	pruefeOhneCode(t, err, model.ErrUngueltig)
	if w.p.gesamt() != 0 {
		t.Errorf("Aufrufe %v", w.p.ops())
	}
}

func zaehler(t *testing.T, w *welt, soll map[string]int) {
	t.Helper()
	for op, n := range soll {
		if got := w.p.anzahl(op); got != n {
			t.Errorf("%s: %d, erwartet %d (Folge %v)", op, got, n, w.p.ops())
		}
	}
}

func kaputtesManifest(w *welt) {
	w.baselineBytes("bsl-1", "svc", mitBaum(func(m obj) { delete(m, "entity_count") }))
}

// createVerstoesse sind V1 bis V10 und K bei Create (SPEC-129 Nr. 12).
func createVerstoesse() []verstossFall {
	return []verstossFall{
		{"V1 ctx nil", 1, func(w *welt, a *aufruf) { a.ctx = nil }, ohnePort},
		{"V2 baselineID leer", 2, func(w *welt, a *aufruf) { a.baselineID = "" }, ohnePort},
		{"V3 id ungültig", 2, func(w *welt, a *aufruf) { a.id = "Gross" }, ohnePort},
		{"K Kontext abgebrochen", 3, func(w *welt, a *aufruf) { a.ctx = abgebrochen() }, func(t *testing.T, w *welt, err error) {
			pruefeOhneCode(t, err, context.Canceled)
			if !slices.Equal(w.p.ops(), []string{"Snapshots.Get"}) {
				t.Errorf("Aufrufe %v", w.p.ops())
			}
		}},
		{"V4 Baseline unbekannt", 4, func(w *welt, a *aufruf) { a.baselineID = "unbekannt" }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE7002, outbound.ErrNichtGefunden)
			zaehler(t, w, map[string]int{"Blobs.Open": 0, lblTx: 0})
		}},
		{"V5 StateHash ungleich StateBlob", 5, func(w *welt, a *aufruf) {
			b := w.snaps.baselines["bsl-1"]
			b.StateHash = "sha256:" + strings.Repeat("0", 64)
			w.snaps.baselines["bsl-1"] = b
		}, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
			zaehler(t, w, map[string]int{"Snapshots.Get": 1, "Blobs.Open": 0, lblTx: 0})
		}},
		{"V6 Open", 6, func(w *welt, a *aufruf) { w.p.setzeFehler("Blobs.Open", outbound.ErrBlobFehlt) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBlobFehlt)
			zaehler(t, w, map[string]int{"Blobs.Open": 1, "Strom.Read": 0, "Strom.Close": 0, lblTx: 0})
		}},
		{"V7 Lesefehler", 7, func(w *welt, a *aufruf) { w.p.setzeFehler("Strom.Read", outbound.ErrBlobBeschaedigt) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBlobBeschaedigt)
			zaehler(t, w, map[string]int{"Blobs.Open": 1, "Strom.Close": 1, lblTx: 0})
		}},
		{"V8 Manifest ungültig", 8, func(w *welt, a *aufruf) { kaputtesManifest(w) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
			if errors.Is(err, outbound.ErrBlobBeschaedigt) {
				t.Error("Ursache des Lesefehlers statt der des Manifests")
			}
			zaehler(t, w, map[string]int{"Blobs.Open": 1, "Strom.Close": 1, lblTx: 0})
		}},
		{"V9 id vergeben", 9, func(w *welt, a *aufruf) { a.id = "ws-1"; w.vorbelegen("ws-1", 0) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE7007, outbound.ErrBereitsVorhanden)
			zaehler(t, w, map[string]int{lblTx: 1, "Blobs.Open": 1, "tx Entities.ReplaceAll": 0})
		}},
		{"V10 anderer Fehler der Transaktion", 10, func(w *welt, a *aufruf) { w.p.setzeFehler("tx Entities.ReplaceAll", errPlatte) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE9001, errPlatte)
			zaehler(t, w, map[string]int{lblTx: 1, "tx Workspaces.Create": 1, "tx Entities.ReplaceAll": 1})
		}},
	}
}

func abgebrochen() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// resetVerstoesse sind R1 bis R11 und K bei Reset.
func resetVerstoesse() []verstossFall {
	schritt := func(label string, rang int, name string, vorher ...string) verstossFall {
		return verstossFall{name, rang, func(w *welt, a *aufruf) { w.p.setzeFehler(label, errPlatte) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE9001, errPlatte)
			ops := w.p.ops()
			if ops[len(ops)-1] != label {
				t.Errorf("letzter Aufruf %s, erwartet %s", ops[len(ops)-1], label)
			}
			zaehler(t, w, map[string]int{lblTx: 1})
		}}
	}
	return []verstossFall{
		{"R1 ctx nil", 1, func(w *welt, a *aufruf) { a.ctx = nil }, ohnePort},
		{"K Kontext abgebrochen", 2, func(w *welt, a *aufruf) { a.ctx = abgebrochen() }, func(t *testing.T, w *welt, err error) {
			pruefeOhneCode(t, err, context.Canceled)
			if !slices.Equal(w.p.ops(), []string{"Workspaces.Get"}) {
				t.Errorf("Aufrufe %v", w.p.ops())
			}
		}},
		{"R2 Workspace unbekannt", 3, func(w *welt, a *aufruf) { a.id = "unbekannt" }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
			if !slices.Equal(w.p.ops(), []string{"Workspaces.Get"}) {
				t.Errorf("Aufrufe %v", w.p.ops())
			}
		}},
		{"R3 Baseline unbekannt", 4, func(w *welt, a *aufruf) { delete(w.snaps.baselines, "bsl-1") }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE7002, outbound.ErrNichtGefunden)
			zaehler(t, w, map[string]int{"Blobs.Open": 0, lblTx: 0})
		}},
		{"R4 StateHash", 5, func(w *welt, a *aufruf) {
			b := w.snaps.baselines["bsl-1"]
			b.StateHash = "sha256:" + strings.Repeat("0", 64)
			w.snaps.baselines["bsl-1"] = b
		}, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
			zaehler(t, w, map[string]int{"Blobs.Open": 0, lblTx: 0})
		}},
		{"R5 Open", 6, func(w *welt, a *aufruf) { w.p.setzeFehler("Blobs.Open", outbound.ErrBlobFehlt) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBlobFehlt)
			zaehler(t, w, map[string]int{"Blobs.Open": 1, "Strom.Read": 0, lblTx: 0})
		}},
		{"R6 Lesefehler", 7, func(w *welt, a *aufruf) { w.p.setzeFehler("Strom.Read", outbound.ErrBlobBeschaedigt) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBlobBeschaedigt)
			zaehler(t, w, map[string]int{"Strom.Close": 1, lblTx: 0})
		}},
		{"R7 Manifest", 8, func(w *welt, a *aufruf) { kaputtesManifest(w) }, func(t *testing.T, w *welt, err error) {
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
			if errors.Is(err, outbound.ErrBlobBeschaedigt) {
				t.Error("Ursache des Lesefehlers statt der des Manifests")
			}
			zaehler(t, w, map[string]int{"Blobs.Open": 1, "Strom.Close": 1, lblTx: 0})
		}},
		schritt("tx Workspaces.Get", 9, "R8 erster Schritt"),
		schritt("tx Entities.ReplaceAll", 10, "R9 zweiter Schritt"),
		schritt("tx Cursors.ResetAll", 11, "R10 dritter Schritt"),
		schritt("tx Workspaces.Advance", 12, "R11 vierter Schritt"),
	}
}

// paareNachRang bildet jedes Paar von Verstößen, in der Folge der Liste, mit dem frühen zuerst; ein Paar
// mit K und dem Verstoß ctx nil gibt es nicht (K gilt mit jedem Verstoß außer ctx nil).
func paareNachRang(vs []verstossFall) [][2]verstossFall {
	var r [][2]verstossFall
	for i := range vs {
		for j := i + 1; j < len(vs); j++ {
			a, b := vs[i], vs[j]
			if a.rang > b.rang {
				a, b = b, a
			}
			if (a.name[0] == 'K' && b.rang == 1) || (b.name[0] == 'K' && a.rang == 1) {
				continue
			}
			r = append(r, [2]verstossFall{a, b})
		}
	}
	return r
}

// TestRangfolgeCreate (Z32): jedes Paar von Verstößen bei Create ergibt den früheren; 54 Paare.
func TestRangfolgeCreate(t *testing.T) {
	paare := paareNachRang(createVerstoesse())
	if len(paare) != 54 {
		t.Fatalf("Paare %d, erwartet 54 (45 der Verstöße, 9 mit abgebrochenem Kontext)", len(paare))
	}
	for _, p := range paare {
		frueh, spaet := p[0], p[1]
		if frueh.rang == spaet.rang && frueh.name > spaet.name {
			frueh, spaet = spaet, frueh
		}
		w := neu(t)
		a := &aufruf{ctx: context.Background(), baselineID: "bsl-1", id: "ws-neu"}
		spaet.apply(w, a)
		frueh.apply(w, a)
		vorher := w.st.stand()
		w.p.leeren()
		_, err := w.v.Create(a.ctx, a.baselineID, a.id)
		t.Run(frueh.name+" vor "+spaet.name, func(t *testing.T) { frueh.pruefe(t, w, err) })
		if !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("%s/%s: Zustand geändert", frueh.name, spaet.name)
		}
	}
}

// TestRangfolgeReset (Z45): jedes Paar von Verstößen bei Reset ergibt den früheren; 65 Paare.
func TestRangfolgeReset(t *testing.T) {
	paare := paareNachRang(resetVerstoesse())
	if len(paare) != 65 {
		t.Fatalf("Paare %d, erwartet 65 (55 der Verstöße, 10 mit abgebrochenem Kontext)", len(paare))
	}
	for _, p := range paare {
		frueh, spaet := p[0], p[1]
		w := neu(t)
		w.vorbelegen("ws-1", 0)
		w.veraendern("ws-1")
		a := &aufruf{ctx: context.Background(), id: "ws-1"}
		spaet.apply(w, a)
		frueh.apply(w, a)
		vorher := w.st.stand()
		w.p.leeren()
		_, err := w.v.Reset(a.ctx, a.id)
		t.Run(frueh.name+" vor "+spaet.name, func(t *testing.T) { frueh.pruefe(t, w, err) })
		if !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("%s/%s: Zustand geändert", frueh.name, spaet.name)
		}
	}
}

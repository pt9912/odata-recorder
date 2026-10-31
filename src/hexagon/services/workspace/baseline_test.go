package workspace_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// beideOperationen führt f für Create und Reset aus; die Welt der Reset-Fälle trägt den Workspace ws-1
// mit Mutationen, so dass ein Zustand, der sich ändert, auffällt.
func beideOperationen(t *testing.T, f func(w *welt, op string, lauf func() error)) {
	t.Helper()
	for _, op := range []string{"Create", "Reset"} {
		w := neu(t)
		var lauf func() error
		switch op {
		case "Create":
			lauf = func() error { _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); return err }
		default:
			w.vorbelegen("ws-1", 0)
			w.veraendern("ws-1")
			lauf = func() error { _, err := w.v.Reset(context.Background(), "ws-1"); return err }
		}
		f(w, op, lauf)
	}
}

// TestBaselineGetEinmalVorOpen (Z17): Snapshots.Get genau einmal mit der Kennung unverändert, vor Open;
// ErrNichtGefunden, auch für die Kennung eines Workspace-Snapshots, ist ODR-E7002, danach weder Open noch
// Transaktion.
func TestBaselineGetEinmalVorOpen(t *testing.T) {
	beideOperationen(t, func(w *welt, op string, lauf func() error) {
		w.p.leeren()
		if err := lauf(); err != nil {
			t.Fatalf("%s: %v", op, err)
		}
		ops := w.p.ops()
		get, open := slicesIndex(ops, "Snapshots.Get"), slicesIndex(ops, "Blobs.Open")
		if w.p.anzahl("Snapshots.Get") != 1 || get > open || open < 0 {
			t.Errorf("%s: Get %d mal, Folge %v", op, w.p.anzahl("Snapshots.Get"), ops)
		}
		if a := w.p.von("Snapshots.Get")[0].args; len(a) != 1 || a[0] != "bsl-1" {
			t.Errorf("%s: Argument %v", op, a)
		}
	})
	for _, op := range []string{"Create", "Reset"} {
		w := neu(t)
		var err error
		if op == "Create" {
			_, err = w.v.Create(context.Background(), "snp-ws-1", "ws-neu")
		} else {
			w.vorbelegen("ws-1", 0)
			z := w.st.stand()
			z.ws["ws-1"] = model.Workspace{ID: "ws-1", ServiceID: "svc", BaselineSnapshotID: "snp-ws-1", CreatedAt: t0()}
			w.st.setze(z)
			_, err = w.v.Reset(context.Background(), "ws-1")
		}
		pruefeCode(t, err, model.CodeE7002, outbound.ErrNichtGefunden)
		zaehler(t, w, map[string]int{"Snapshots.Get": 1, "Blobs.Open": 0, lblTx: 0})
	}
}

func slicesIndex(s []string, x string) int {
	for i, e := range s {
		if e == x {
			return i
		}
	}
	return -1
}

// TestStateHashGegenStateBlob (Z18): gleich geht weiter; ein Unterschied an erster oder letzter Stelle,
// im Präfix oder in der Länge ist ODR-E3001 mit ErrBeschaedigt, ohne dass der Blob geöffnet wird.
func TestStateHashGegenStateBlob(t *testing.T) {
	zeichen := func(s string, i int) string {
		b := []byte(s)
		if b[i] == 'x' {
			b[i] = 'y'
		} else {
			b[i] = 'x'
		}
		return string(b)
	}
	abweichungen := map[string]func(ref string) string{
		"erste Stelle": func(r string) string { return zeichen(r, 0) }, "letzte Stelle": func(r string) string { return zeichen(r, len(r)-1) },
		"Präfix": func(r string) string { return "sha512" + r[len("sha256"):] }, "länger": func(r string) string { return r + "0" },
		"kürzer": func(r string) string { return r[:len(r)-1] }, "leer": func(string) string { return "" },
	}
	beideOperationen(t, func(w *welt, op string, lauf func() error) {
		b := w.snaps.baselines["bsl-1"]
		for name, f := range abweichungen {
			b2 := b
			b2.StateHash = f(b.StateBlob)
			w.snaps.baselines["bsl-1"] = b2
			w.p.leeren()
			pruefeCode(t, lauf(), model.CodeE3001, outbound.ErrBeschaedigt)
			zaehler(t, w, map[string]int{"Blobs.Open": 0, lblTx: 0, "Snapshots.Get": 1})
			if t.Failed() {
				t.Logf("%s %s", op, name)
			}
		}
		w.snaps.baselines["bsl-1"] = b
		w.p.leeren()
		if err := lauf(); err != nil {
			t.Errorf("%s: gleich geht nicht weiter: %v", op, err)
		}
	})
}

// TestOpenLesenSchliessen (Z19): Open genau einmal mit StateBlob; der Strom wird bis zum Ende gelesen und genau
// einmal geschlossen, auf jedem Weg nach gelungenem Open; ein Fehler von Close ändert nichts.
func TestOpenLesenSchliessen(t *testing.T) {
	laenge := len(gueltig())
	if laenge < 600 {
		t.Fatalf("das Manifest ist zu kurz für mehrere Read-Aufrufe: %d", laenge)
	}
	type weg struct {
		name   string
		setup  func(w *welt)
		erfolg bool
		// vollGelesen: der Strom ist bis zum Ende gelesen, wenn er geschlossen wird.
		vollGelesen bool
	}
	wege := []weg{
		{"Erfolg", func(w *welt) {}, true, true},
		{"Fehler von Close", func(w *welt) { w.blobs.closeFehler = errPlatte }, true, true},
		{"Lesefehler Offset 0", func(w *welt) { w.blobs.fehlerBei, w.blobs.lesefehler = 0, outbound.ErrBlobBeschaedigt }, false, false},
		{"Lesefehler Mitte", func(w *welt) { w.blobs.fehlerBei, w.blobs.lesefehler = laenge/2, outbound.ErrBlobBeschaedigt }, false, false},
		{"Lesefehler Ende", func(w *welt) { w.blobs.fehlerBei, w.blobs.lesefehler = laenge, outbound.ErrBlobBeschaedigt }, false, true},
		{"Kontextfehler beim Lesen", func(w *welt) { w.blobs.fehlerBei, w.blobs.lesefehler = laenge/3, context.Canceled }, false, false},
		{"Manifest Syntax", func(w *welt) { w.baselineBytes("bsl-1", "svc", gueltig()[:laenge-1]) }, false, true},
		{"Manifest Kopf", func(w *welt) { w.baselineBytes("bsl-1", "svc", mitBaum(func(m obj) { m["service_id"] = "x" })) }, false, true},
		{"Manifest Eintrag", func(w *welt) { w.baselineBytes("bsl-1", "svc", mitEintrag(3, 2, func(e obj) { e["payload"] = nil })) }, false, true},
		{"Manifest doppelt", func(w *welt) { w.baselineBytes("bsl-1", "svc", doppelte()[0].manifest) }, false, true},
	}
	for _, wg := range wege {
		beideOperationen(t, func(w *welt, op string, lauf func() error) {
			wg.setup(w)
			w.p.leeren()
			err := lauf()
			if (err == nil) != wg.erfolg {
				t.Errorf("%s/%s: Fehler %v", op, wg.name, err)
			}
			st := w.blobs.strome()
			if len(st) != 1 || w.p.anzahl("Blobs.Open") != 1 {
				t.Fatalf("%s/%s: %d Ströme, %d Open", op, wg.name, len(st), w.p.anzahl("Blobs.Open"))
			}
			if got := w.p.von("Blobs.Open")[0].args[0]; got != w.snaps.baselines["bsl-1"].StateBlob {
				t.Errorf("%s/%s: Open mit %v", op, wg.name, got)
			}
			if st[0].geschlossen() != 1 {
				t.Errorf("%s/%s: Close %d mal", op, wg.name, st[0].geschlossen())
			}
			if wg.vollGelesen && st[0].restBeiClose != 0 {
				t.Errorf("%s/%s: Rest %d Bytes beim Schließen", op, wg.name, st[0].restBeiClose)
			}
		})
	}
	// Die Wege sind eine Menge von zehn; der Strom einer Anlage wird vom Test nie nach dem Schließen gelesen.
	if len(wege) != 10 {
		t.Errorf("Wege %d", len(wege))
	}
}

// TestManifestVerstoesse (Z20 bis Z23): jeder Verstoß gegen Syntax, Kopf, Einträge und Eindeutigkeit ist
// ODR-E3001 mit ErrBeschaedigt, bei Create und bei Reset, ohne Transaktion; der Zustand bleibt.
func TestManifestVerstoesse(t *testing.T) {
	gruppen := []struct {
		name   string
		faelle []manifestFall
		zahl   int
	}{
		{"Syntax", syntaxVerstoesse(), len(gueltig()) + 10 + 4},
		{"Kopf", kopfVerstoesse(), 12 + 7 + 7 + 7 + 13 + 1},
		{"Einträge", eintragVerstoesse(), 4*(7+7+9+7+13) + 6 + 2 + 15},
		{"Doppelte", doppelte(), 3},
	}
	summe := 0
	for _, g := range gruppen {
		if len(g.faelle) != g.zahl {
			t.Errorf("%s: %d Fälle, erwartet %d", g.name, len(g.faelle), g.zahl)
		}
		summe += len(g.faelle)
		for _, f := range g.faelle {
			beideOperationen(t, func(w *welt, op string, lauf func() error) {
				w.baselineBytes("bsl-1", "svc", f.manifest)
				vorher := w.st.stand()
				w.p.leeren()
				manifestAbgewiesen(t, w, lauf(), g.name+"/"+f.name+"/"+op)
				if !zustaendeGleich(vorher, w.st.stand()) {
					t.Errorf("%s/%s/%s: Zustand geändert", g.name, f.name, op)
				}
			})
		}
	}
	if len(alleVerstoesse()) != summe {
		t.Errorf("alleVerstoesse %d, Summe der Gruppen %d", len(alleVerstoesse()), summe)
	}
	t.Logf("Verstöße des Manifests: Syntax %d, Kopf %d, Einträge %d, Doppelte %d, zusammen %d; Länge des gültigen Manifests %d Bytes",
		len(syntaxVerstoesse()), len(kopfVerstoesse()), len(eintragVerstoesse()), len(doppelte()), summe, len(gueltig()))
}

// TestManifestGueltigeFormen (Z20 bis Z23, Gegenseite): die Kontrolle und die Formen, die gelten, werden
// angelegt; der Leerraum dahinter gilt, entity_count gleich der Länge gilt.
func TestManifestGueltigeFormen(t *testing.T) {
	for name, f := range gueltigeFormen() {
		w := neu(t)
		w.baselineBytes("bsl-1", "svc", f.manifest)
		if _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if got := len(w.p.von("tx Entities.ReplaceAll")[0].args[1].([]model.EntityState)); got != f.n {
			t.Errorf("%s: %d Einträge, erwartet %d", name, got, f.n)
		}
	}
}

// verstoesseVorDerAnlage sind alle Szenarien, in denen Create oder Reset vor Nr. 4 e (Kennung und Zeit)
// endet: die Verstöße der Argumente und der Baseline.
func verstoesseVorDerAnlage() []verstossFall {
	var r []verstossFall
	for _, v := range createVerstoesse() {
		if v.rang <= 8 && v.name[0] != 'K' {
			r = append(r, v)
		}
	}
	for _, f := range alleVerstoesse() {
		r = append(r, verstossFall{"Manifest " + f.name, 8, func(w *welt, a *aufruf) { w.baselineBytes("bsl-1", "svc", f.manifest) }, nil})
	}
	return r
}

// TestVerstossVerbrauchtWederZeitNochKennung (Z16, Z24): ein Verstoß vor Nr. 4 e verbraucht weder Kennung noch
// Zeit, öffnet keine Transaktion und lässt den Zustand unverändert; die Menge sind V1 bis V8 und jeder
// Verstoß des Manifests.
func TestVerstossVerbrauchtWederZeitNochKennung(t *testing.T) {
	vs := verstoesseVorDerAnlage()
	if len(vs) != 8+len(alleVerstoesse()) {
		t.Fatalf("Szenarien %d", len(vs))
	}
	for _, v := range vs {
		w := neu(t)
		a := &aufruf{ctx: context.Background(), baselineID: "bsl-1", id: ""}
		v.apply(w, a)
		vorher := w.st.stand()
		w.p.leeren()
		_, err := w.v.Create(a.ctx, a.baselineID, a.id)
		if err == nil {
			t.Errorf("%s: kein Fehler", v.name)
		}
		zaehler(t, w, map[string]int{"Uhr.Now": 0, "IDs.WorkspaceID": 0, lblTx: 0})
		if w.p.anzahl("IDs.SessionID")+w.p.anzahl("IDs.SnapshotID") != 0 {
			t.Errorf("%s: Kennung verbraucht", v.name)
		}
		if !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("%s: Zustand geändert", v.name)
		}
	}
	// Reset: derselbe Satz ohne Zeit und Kennung, ohne Transaktion, Zustand unverändert.
	resetVs := []verstossFall{}
	for _, v := range resetVerstoesse() {
		if v.name[0] == 'R' && v.rang >= 4 && v.rang <= 8 {
			resetVs = append(resetVs, v)
		}
	}
	for _, f := range alleVerstoesse() {
		resetVs = append(resetVs, verstossFall{"Manifest " + f.name, 8, func(w *welt, a *aufruf) { w.baselineBytes("bsl-1", "svc", f.manifest) }, nil})
	}
	if len(resetVs) != 5+len(alleVerstoesse()) {
		t.Fatalf("Szenarien von Reset %d", len(resetVs))
	}
	for _, v := range resetVs {
		w := neu(t)
		w.vorbelegen("ws-1", 0)
		w.veraendern("ws-1")
		a := &aufruf{ctx: context.Background(), baselineID: "bsl-1", id: "ws-1"}
		v.apply(w, a)
		vorher := w.st.stand()
		w.p.leeren()
		if _, err := w.v.Reset(a.ctx, a.id); err == nil {
			t.Errorf("Reset %s: kein Fehler", v.name)
		}
		zaehler(t, w, map[string]int{"Uhr.Now": 0, "IDs.WorkspaceID": 0, lblTx: 0})
		if !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("Reset %s: Zustand geändert", v.name)
		}
	}
}

// TestZeitGenauEinmal (Z16): ein Create, das Nr. 4 e erreicht, bezieht die Zeit genau einmal, und CreatedAt
// ist dieser Wert; auch eine Doppelung bezieht sie, denn sie entsteht erst in der Transaktion.
func TestZeitGenauEinmal(t *testing.T) {
	w := neu(t)
	w.uhr.u.Advance(3 * time.Second)
	ws, err := w.v.Create(context.Background(), "bsl-1", "")
	if err != nil {
		t.Fatal(err)
	}
	zaehler(t, w, map[string]int{"Uhr.Now": 1, "IDs.WorkspaceID": 1})
	if !ws.CreatedAt.Equal(t0().Add(3 * time.Second)) {
		t.Errorf("CreatedAt %v", ws.CreatedAt)
	}
	if got := w.st.stand().ws[ws.ID].CreatedAt; !got.Equal(ws.CreatedAt) {
		t.Errorf("gespeichertes CreatedAt %v", got)
	}
	if _, err := w.v.Create(context.Background(), "bsl-1", ws.ID); err == nil {
		t.Error("Doppelung nicht abgewiesen")
	}
	zaehler(t, w, map[string]int{"Uhr.Now": 2, "IDs.WorkspaceID": 1})
}

// TestFehlerOhneWert: der Text eines Fehlers des Manifests nennt keinen Wert aus dem Manifest.
func TestFehlerOhneWert(t *testing.T) {
	w := neu(t)
	w.baselineBytes("bsl-1", "svc", mitEintrag(3, 0, func(e obj) { e["entity_set"] = "GeheimerWertXYZ"; e["completeness"] = "GeheimerWertXYZ" }))
	_, err := w.v.Create(context.Background(), "bsl-1", "ws-neu")
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
	u := err
	for u != nil {
		if strings.Contains(u.Error(), "GeheimerWertXYZ") {
			t.Errorf("Wert im Fehlertext: %v", u)
		}
		u = errors.Unwrap(u)
	}
}

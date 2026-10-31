package workspace_test

import (
	"context"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

var _ inbound.WorkspaceUseCase = (*workspace.Verwaltung)(nil)

// TestLoeschenRuftDieBereinigungNachDemLoeschen (Z35): Workspaces.Delete genau einmal mit der id unverändert,
// danach genau einmal CleanupBlobs mit dem ctx des Aufrufs unverändert; kein Lesen davor.
func TestLoeschenRuftDieBereinigungNachDemLoeschen(t *testing.T) {
	for _, id := range idFormen() {
		w := neu(t)
		z := w.st.stand()
		z.ws[id] = model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-1", CreatedAt: t0()}
		w.st.setze(z)
		w.wartung.setze(inbound.Loeschbericht{Entfernt: 3}, nil)
		w.p.leeren()
		ctx := neuerKontext(context.WithValue(context.Background(), ctxSchluessel{}, "wert"))
		bericht, err := w.v.Delete(ctx, id)
		if err != nil || bericht.Entfernt != 3 {
			t.Errorf("id %q: %+v, %v", id, bericht, err)
		}
		if ops := w.p.ops(); !slices.Equal(ops, []string{"Workspaces.Delete", "Wartung.CleanupBlobs"}) {
			t.Errorf("id %q: Folge %v", id, ops)
		}
		if a := w.p.von("Workspaces.Delete")[0]; a.args[0] != id || a.ctx != context.Context(ctx) {
			t.Errorf("id %q: Delete mit %v", id, a.args)
		}
		if c := w.p.von("Wartung.CleanupBlobs"); len(c) != 1 || c[0].ctx != context.Context(ctx) {
			t.Errorf("id %q: Bereinigung %v", id, c)
		}
		if _, da := w.st.stand().ws[id]; da {
			t.Errorf("id %q: nicht gelöscht", id)
		}
	}
}

// TestLoeschenFehlerRuftKeineBereinigung (Z35): nach jedem Fehler von Workspaces.Delete, in der Menge der Tabelle, wird die
// Bereinigung nie gerufen; der Fehler hat den Gegenstand Workspace; der Bericht ist der Nullwert.
func TestLoeschenFehlerRuftKeineBereinigung(t *testing.T) {
	for _, fe := range tabellenFehlerListe() {
		w := neu(t)
		w.vorbelegen("ws-1", 0)
		w.wartung.setze(inbound.Loeschbericht{Entfernt: 9}, nil)
		w.p.setzeFehler("Workspaces.Delete", fe.err)
		w.p.leeren()
		bericht, err := w.v.Delete(context.Background(), "ws-1")
		c, _ := outbound.Abbilden(fe.err, outbound.GegenstandWorkspace)
		pruefeCode(t, err, c, fe.err)
		if bericht != (inbound.Loeschbericht{}) || w.p.anzahl("Wartung.CleanupBlobs") != 0 || w.p.gesamt() != 1 {
			t.Errorf("%s: Bericht %+v, Folge %v", fe.name, bericht, w.p.ops())
		}
	}
	w := neu(t)
	_, err := w.v.Delete(context.Background(), "unbekannt")
	pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
	if w.p.anzahl("Wartung.CleanupBlobs") != 0 {
		t.Error("Bereinigung nach ODR-E7001")
	}
}

// bereinigungsFaelle sind die sechs Ergebnisse von CleanupBlobs (Z36), mit teilweise gefüllten Berichten.
func bereinigungsFaelle() []struct {
	name    string
	bericht inbound.Loeschbericht
	err     error
} {
	return []struct {
		name    string
		bericht inbound.Loeschbericht
		err     error
	}{
		{"Erfolg", inbound.Loeschbericht{Entfernt: 3}, nil},
		{"Zurückgestellt", inbound.Loeschbericht{Zurueckgestellt: true}, nil},
		{"Fehler mit Code", inbound.Loeschbericht{Entfernt: 2, Zurueckgestellt: true}, &inbound.Fehler{Code: model.CodeE3001, Status: 500}},
		{"Fehler ohne Code", inbound.Loeschbericht{Entfernt: 1}, errPlatte},
		{"Canceled", inbound.Loeschbericht{Entfernt: 5}, context.Canceled},
		{"DeadlineExceeded", inbound.Loeschbericht{Entfernt: 4, Zurueckgestellt: true}, context.DeadlineExceeded},
	}
}

// TestLoeschenErgebnisIstDerBericht (Z36): das Ergebnis ist der Bericht von CleanupBlobs unverändert und der Fehler nil, auch
// bei einem Fehler der Bereinigung, bei Zurueckgestellt und bei einem Kontextfehler.
func TestLoeschenErgebnisIstDerBericht(t *testing.T) {
	faelle := bereinigungsFaelle()
	if len(faelle) != 6 {
		t.Fatalf("Fälle %d", len(faelle))
	}
	for _, f := range faelle {
		w := neu(t)
		w.vorbelegen("ws-1", 0)
		w.wartung.setze(f.bericht, f.err)
		bericht, err := w.v.Delete(context.Background(), "ws-1")
		if err != nil || bericht != f.bericht {
			t.Errorf("%s: %+v, %v, erwartet %+v ohne Fehler", f.name, bericht, err, f.bericht)
		}
	}
}

// TestLoeschenNachBeendetemKontext (Z08): ein Delete, dessen Löschen committed ist, ruft die Bereinigung auch bei
// inzwischen beendetem Kontext, einmal und mit diesem ctx.
func TestLoeschenNachBeendetemKontext(t *testing.T) {
	w := neu(t)
	w.vorbelegen("ws-1", 0)
	w.wartung.setze(inbound.Loeschbericht{Entfernt: 2}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	w.p.nach = func(op string) {
		if op == "Workspaces.Delete" {
			cancel()
		}
	}
	bericht, err := w.v.Delete(ctx, "ws-1")
	if err != nil || bericht.Entfernt != 2 {
		t.Errorf("%+v, %v", bericht, err)
	}
	c := w.p.von("Wartung.CleanupBlobs")
	if len(c) != 1 || c[0].ctx != ctx || c[0].ctx.Err() == nil {
		t.Errorf("Bereinigung %v", c)
	}
	if _, da := w.st.stand().ws["ws-1"]; da {
		t.Error("nicht gelöscht")
	}
}

// drei legt die Workspaces ws-a, ws-b, ws-c mit State, Cursors und Snapshots an.
func drei(t *testing.T) *welt {
	w := neu(t)
	for _, id := range []string{"ws-a", "ws-b", "ws-c"} {
		w.vorbelegen(id, 1)
		w.veraendern(id)
	}
	w.p.leeren()
	return w
}

// TestLoeschenBerührtNurDieseID (Z38): Löschen des mittleren von drei Workspaces entfernt Workspace, State, Cursors und
// Snapshots dieser id, die anderen bleiben; ein zweites Löschen ist ODR-E7001 ohne Bereinigung.
func TestLoeschenBeruehrtNurDieseID(t *testing.T) {
	w := drei(t)
	vorher := w.st.stand()
	if _, err := w.v.Delete(context.Background(), "ws-b"); err != nil {
		t.Fatal(err)
	}
	nachher := w.st.stand()
	if _, da := nachher.ws["ws-b"]; da || len(nachher.ent["ws-b"]) != 0 || len(nachher.cur["ws-b"]) != 0 || len(nachher.snaps["ws-b"]) != 0 {
		t.Errorf("ws-b nicht vollständig gelöscht")
	}
	for _, id := range []string{"ws-a", "ws-c"} {
		if !zustaendeGleich(nurWorkspace(vorher, id), nurWorkspace(nachher, id)) {
			t.Errorf("%s geändert", id)
		}
	}
	_, err := w.v.Delete(context.Background(), "ws-b")
	pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
	if w.p.anzahl("Wartung.CleanupBlobs") != 1 || w.p.anzahl("Workspaces.Delete") != 2 || w.p.anzahl("Workspaces.List") != 0 {
		t.Errorf("Folge %v", w.p.ops())
	}
}

// TestGleichzeitigesLoeschen (Z55): zwei gleichzeitige Delete derselben Kennung: genau eines gelingt und ruft die
// Bereinigung, das andere ist ODR-E7001 und ruft sie nicht.
func TestGleichzeitigesLoeschen(t *testing.T) {
	for n := 2; n <= 4; n++ {
		w := neu(t)
		w.vorbelegen("ws-1", 0)
		w.wartung.setze(inbound.Loeschbericht{Entfernt: 1}, nil)
		fehler := make([]error, n)
		gleichzeitig(n, func(i int) { _, fehler[i] = w.v.Delete(context.Background(), "ws-1") })
		gelungen := 0
		for _, err := range fehler {
			if err == nil {
				gelungen++
			} else {
				pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
			}
		}
		if gelungen != 1 || w.p.anzahl("Wartung.CleanupBlobs") != 1 || w.p.anzahl("Workspaces.Delete") != n {
			t.Errorf("n=%d: gelungen %d, Bereinigung %d, Delete %d", n, gelungen, w.p.anzahl("Wartung.CleanupBlobs"), w.p.anzahl("Workspaces.Delete"))
		}
	}
}

// TestLoeschenGegenCreate (Z56): Delete gegen Create einer anderen id: beide gelingen unabhängig, das Löschen berührt den
// neuen Workspace nicht; in beiden Reihenfolgen und gleichzeitig.
func TestLoeschenGegenCreate(t *testing.T) {
	pruefe := func(name string, w *welt) {
		z := w.st.stand()
		if _, da := z.ws["ws-alt"]; da || len(z.ent["ws-alt"]) != 0 {
			t.Errorf("%s: ws-alt nicht gelöscht", name)
		}
		if _, da := z.ws["ws-neu"]; !da || !equalState(z.ent["ws-neu"], erwarteterState("ws-neu", standardEnts())) {
			t.Errorf("%s: ws-neu berührt", name)
		}
		if _, da := z.ws["ws-bleibt"]; !da {
			t.Errorf("%s: ws-bleibt berührt", name)
		}
	}
	lauf := map[string]func(w *welt){
		"Create dann Delete": func(w *welt) {
			if _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); err != nil {
				t.Error(err)
			}
			if _, err := w.v.Delete(context.Background(), "ws-alt"); err != nil {
				t.Error(err)
			}
		},
		"Delete dann Create": func(w *welt) {
			if _, err := w.v.Delete(context.Background(), "ws-alt"); err != nil {
				t.Error(err)
			}
			if _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); err != nil {
				t.Error(err)
			}
		},
		"gleichzeitig": func(w *welt) {
			gleichzeitig(2, func(i int) {
				var err error
				if i == 0 {
					_, err = w.v.Create(context.Background(), "bsl-1", "ws-neu")
				} else {
					_, err = w.v.Delete(context.Background(), "ws-alt")
				}
				if err != nil {
					t.Error(err)
				}
			})
		},
	}
	for name, f := range lauf {
		w := neu(t)
		w.vorbelegen("ws-alt", 0)
		w.vorbelegen("ws-bleibt", 0)
		f(w)
		pruefe(name, w)
	}
	if len(lauf) != 3 {
		t.Errorf("Läufe %d", len(lauf))
	}
}

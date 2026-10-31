package workspace_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// idFormen sind die Formen, die Get, Delete und Reset unverändert an den Port reichen (SPEC-129 Nr. 8): leer,
// 65 Bytes, Großbuchstaben, NUL, Byte ab 0x80.
func idFormen() []string {
	return []string{"", strings.Repeat("a", 65), "Gross", "a\x00b", "\x80x", "\xc3\xa4", "ws 1", "ws/1", "ws-1"}
}

// TestGetReichtIDDurch (Z33): Get reicht die id unverändert an Workspaces.Get, liefert den Workspace
// unverändert, ODR-E7001 bei Unbekanntem und Gelöschtem; kein anderer Port.
func TestGetReichtIDDurch(t *testing.T) {
	for _, id := range idFormen() {
		w := neu(t)
		z := w.st.stand()
		soll := model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-1", Revision: 3, CreatedAt: t0()}
		z.ws[id] = soll
		w.st.setze(z)
		w.p.leeren()
		got, err := w.v.Get(context.Background(), id)
		if err != nil || got != soll {
			t.Errorf("id %q: %+v, %v", id, got, err)
		}
		if ev := w.p.von("Workspaces.Get"); len(ev) != 1 || ev[0].args[0] != id || w.p.gesamt() != 1 {
			t.Errorf("id %q: Aufrufe %v", id, w.p.ops())
		}
		// gelöscht und unbekannt
		z = w.st.stand()
		delete(z.ws, id)
		w.st.setze(z)
		w.p.leeren()
		got, err = w.v.Get(context.Background(), id)
		pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
		if got != (model.Workspace{}) || w.p.gesamt() != 1 {
			t.Errorf("id %q: Nullwert erwartet, Aufrufe %v", id, w.p.ops())
		}
	}
}

// listeWelt legt n Services mit je k Workspaces an; die Zeiten steigen, die Kennungen fallen, so dass eine
// Neuordnung nach Kennung auffällt, und die Services stehen nicht in der Folge ihrer Kennung.
func listeWelt(t *testing.T, n, k int) (*welt, []model.Workspace) {
	w := neu(t)
	w.dienste.alle = nil
	z := w.st.stand()
	var soll []model.Workspace
	for i := 0; i < n; i++ {
		svc := fmt.Sprintf("svc-%c", 'z'-i)
		w.dienste.alle = append(w.dienste.alle, model.Service{ID: svc})
		for j := 0; j < k; j++ {
			ws := model.Workspace{ID: fmt.Sprintf("ws-%s-%c", svc, 'z'-j), ServiceID: svc, BaselineSnapshotID: "bsl-1", CreatedAt: t0().Add(time.Duration(j) * time.Second)}
			z.ws[ws.ID] = ws
			soll = append(soll, ws)
		}
	}
	w.st.setze(z)
	w.p.leeren()
	return w, soll
}

// TestListeJederService (Z34): Dienste.List einmal, Workspaces.List je Service einmal in dessen Reihenfolge,
// keine Neuordnung, nie nil.
func TestListeJederService(t *testing.T) {
	faelle := 0
	for _, n := range []int{0, 1, 3} {
		for _, k := range []int{0, 1, 2} {
			w, soll := listeWelt(t, n, k)
			got, err := w.v.List(context.Background())
			if err != nil || got == nil || !slices.Equal(got, soll) && (len(got) != 0 || len(soll) != 0) {
				t.Errorf("%d Services mit %d: %+v, %v", n, k, got, err)
			}
			zaehler(t, w, map[string]int{"Dienste.List": 1, "Workspaces.List": n})
			for i, ev := range w.p.von("Workspaces.List") {
				if ev.args[0] != w.dienste.alle[i].ID {
					t.Errorf("Aufruf %d mit %v", i, ev.args[0])
				}
			}
			faelle++
		}
	}
	if faelle != 9 {
		t.Errorf("Fälle %d", faelle)
	}
}

// TestListeErsterFehlerBeendet (Z34): ein Fehler an jedem der 1 plus n Aufrufe beendet List ohne Teilliste
// und ohne weiteren Aufruf.
func TestListeErsterFehlerBeendet(t *testing.T) {
	const n = 3
	for k := 1; k <= 1+n; k++ {
		w, _ := listeWelt(t, n, 2)
		label, nr := "Dienste.List", 1
		if k > 1 {
			label, nr = "Workspaces.List", k-1
		}
		w.p.setzeFehler(fmt.Sprintf("%s#%d", label, nr), errPlatte)
		got, err := w.v.List(context.Background())
		pruefeCode(t, err, model.CodeE9001, errPlatte)
		if got != nil {
			t.Errorf("Aufruf %d: Teilliste %v", k, got)
		}
		if w.p.gesamt() != k {
			t.Errorf("Aufruf %d: %d Aufrufe %v", k, w.p.gesamt(), w.p.ops())
		}
	}
}

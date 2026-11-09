package mock_test

import (
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// wettlauf fährt n Anfragen gleichzeitig gegen einen Aufbau mit Eintrag 2 in der Version 5; sperreAmEingang hält jede Transaktion am Eingang, bis
// alle n angekommen sind: was der Dienst vor der Transaktion liest, haben dann alle n gelesen, bevor einer schreibt. Der Fake zählt jede Operation.
func wettlauf(t *testing.T, n int, anfragen func(i int) inbound.Anfrage) (*umg, *ergebnisse) {
	t.Helper()
	u := etagUmg(t, true, 5, 5, 5)
	u.dynamisch()
	sperreAmEingang(u, int32(n))
	m := u.mocker()
	var r ergebnisse
	aufEinmal(n, func(i int) { r.nimm(m.Mock(u.ctx, anfragen(i))) })
	return u, &r
}

// mitKopf ist der dynamische Request einer Methode an den Schlüssel 2 mit der Kopfzeile If-Match.
func mitKopf(m string, i int, ifMatchWert string) inbound.Anfrage {
	a := dynAnfrage(m, "2", fmt.Sprint("g", i), wsID)
	a.Header["If-Match"] = []string{ifMatchWert}
	return a
}

// pruefeWettlauf prüft die Wirkung: genau ein Erfolg, die übrigen mit dem Code fehler, Put und Advance einmal, Version und Revision plus 1,
// eine Transaktion und ein Lesen des Eintrags je Anfrage.
func pruefeWettlauf(t *testing.T, kontext string, u *umg, r *ergebnisse, n int, fehler model.Code, geloescht bool) {
	t.Helper()
	got := u.zustand()["2"]
	if r.stati[204] != 1 || r.codes[fehler] != n-1 || len(r.stati) != 1 || len(r.codes) != 1 || r.anders != 0 {
		t.Errorf("%s: Status %v Codes %v anders %d; erwartet ein 204 und %d × %s", kontext, r.stati, r.codes, r.anders, n-1, fehler)
	}
	if u.p.anzahl("Entities.Put") != 1 || u.p.anzahl("Workspaces.Advance") != 1 || got.Version != 6 || got.Deleted != geloescht || u.ws.revision(wsID) != 1 {
		t.Errorf("%s: Put %d Advance %d Version %d gelöscht %v Revision %d; erwartet 1, 1, 6, %v, 1", kontext, u.p.anzahl("Entities.Put"),
			u.p.anzahl("Workspaces.Advance"), got.Version, got.Deleted, u.ws.revision(wsID), geloescht)
	}
	if u.p.anzahl("Transaktion.WithWorkspaceTransaction") != n || u.p.anzahl("Entities.Get") != n {
		t.Errorf("%s: Transaktion %d Entities.Get %d; erwartet je %d", kontext, u.p.anzahl("Transaktion.WithWorkspaceTransaction"), u.p.anzahl("Entities.Get"), n)
	}
}

// TestIfMatchWettlaufPatch (F15): von 20 gleichzeitigen PATCH mit demselben If-Match gelingt genau einer, die übrigen sind ODR-E6010; die
// Version wächst um 1, die Revision um 1, Put einmal. Fenster: Lesen der Version und Vergleich (SPEC-283 Nr. 8 a). 4 Runden.
func TestIfMatchWettlaufPatch(t *testing.T) {
	for runde := 0; runde < 4; runde++ {
		u, r := wettlauf(t, 20, func(i int) inbound.Anfrage { return mitKopf("PATCH", i, `W/"5"`) })
		pruefeWettlauf(t, fmt.Sprint("PATCH Runde ", runde), u, r, 20, model.CodeE6010, false)
	}
}

// TestIfMatchWettlaufPut (F16): dasselbe für 20 gleichzeitige PUT und für eine Mischung aus je 10 PATCH und PUT; Fenster: Vergleich und Schreiben
// (SPEC-283 Nr. 8 b). 2 Formen × 4 Runden.
func TestIfMatchWettlaufPut(t *testing.T) {
	formen := map[string]func(i int) string{
		"PUT":      func(int) string { return "PUT" },
		"Mischung": func(i int) string { return []string{"PATCH", "PUT"}[i%2] },
	}
	for name, methode := range formen {
		for runde := 0; runde < 4; runde++ {
			u, r := wettlauf(t, 20, func(i int) inbound.Anfrage { return mitKopf(methode(i), i, `"5"`) })
			pruefeWettlauf(t, fmt.Sprint(name, " Runde ", runde), u, r, 20, model.CodeE6010, false)
		}
	}
}

// TestIfMatchWettlaufFalsche (F17): 20 gleichzeitige Anfragen mit falschem Tag neben einer mit richtigem: ein 204, 20 × ODR-E6010; Put einmal
// und Advance einmal, Version und Revision plus 1 (Fenster: Entscheidung 412 und Ende der Transaktion, SPEC-283 Nr. 8 c). Methode 2 × 4 Runden.
func TestIfMatchWettlaufFalsche(t *testing.T) {
	for _, m := range []string{"PATCH", "PUT"} {
		for runde := 0; runde < 4; runde++ {
			u, r := wettlauf(t, 21, func(i int) inbound.Anfrage {
				if i == 0 {
					return mitKopf(m, i, `W/"5"`)
				}
				return mitKopf(m, i, `W/"4"`)
			})
			pruefeWettlauf(t, fmt.Sprint(m, " Runde ", runde), u, r, 21, model.CodeE6010, false)
		}
	}
}

// TestIfMatchWettlaufDelete (F18): von 20 gleichzeitigen DELETE mit demselben If-Match gelingt genau einer, die übrigen sind ODR-E6008 (der
// Tombstone), nie ODR-E6010 (SPEC-283 Nr. 8 d). 4 Runden.
func TestIfMatchWettlaufDelete(t *testing.T) {
	for runde := 0; runde < 4; runde++ {
		u, r := wettlauf(t, 20, func(i int) inbound.Anfrage { return mitKopf("DELETE", i, `W/"5"`) })
		pruefeWettlauf(t, fmt.Sprint("DELETE Runde ", runde), u, r, 20, model.CodeE6008, true)
		if r.codes[model.CodeE6010] != 0 {
			t.Errorf("Runde %d: %d × ODR-E6010", runde, r.codes[model.CodeE6010])
		}
	}
}

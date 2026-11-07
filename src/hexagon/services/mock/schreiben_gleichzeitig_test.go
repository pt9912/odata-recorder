package mock_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// dynamisch macht den URL-Leser und den Leser der Bodies abhängig vom Request, damit viele Goroutinen verschiedene Mutationen
// gegen einen Dienst fahren: der Pfad hinter dem Root ist /<Methode>/<Schlüssel>, der Body "<Schlüssel>|<Name>".
func (u *umg) dynamisch() {
	u.c.url.f = func(pfad, _ string) (model.ODataURL, error) {
		teile := strings.Split(strings.Trim(pfad, "/"), "/")
		m, key := teile[len(teile)-2], teile[len(teile)-1]
		if m == "POST" || m == "GET" {
			return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: setProducts}}, nil
		}
		return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: setProducts, Schluessel: []model.SchluesselTeil{zahl("", key)}}}, nil
	}
	u.ei.f = func(_ context.Context, src io.Reader, _ []string, _ string, _ model.EingabeArt) (model.EntityEingabe, error) {
		b, _ := io.ReadAll(src)
		teile := strings.SplitN(string(b), "|", 2)
		return eingabeIDName(teile[0], teile[1]), nil
	}
}

// dynAnfrage ist der Request an den dynamischen Aufbau.
func dynAnfrage(m, key, name, ws string) inbound.Anfrage {
	a := anfrage(m, dienstRoot+"/"+m+"/"+key)
	a.Header["X-Recorder-Workspace"] = []string{ws}
	body := key + "|" + name
	a.Body, a.Laenge = io.NopCloser(strings.NewReader(body)), int64(len(body))
	return a
}

// ergebnisse sammelt Code und Status der Antworten gleichzeitiger Aufrufe.
type ergebnisse struct {
	mu     sync.Mutex
	stati  map[int]int
	codes  map[model.Code]int
	anders int
}

func (e *ergebnisse) nimm(a inbound.Antwort, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.stati == nil {
		e.stati, e.codes = map[int]int{}, map[model.Code]int{}
	}
	if err != nil {
		if c := fehlercode(err); c != 0 {
			e.codes[c]++
		} else {
			e.anders++
		}
		return
	}
	e.stati[a.Status]++
}

// aufEinmal fährt n Aufrufe gleichzeitig los (ein Start für alle) und wartet auf alle.
func aufEinmal(n int, f func(i int)) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			f(i)
		}()
	}
	close(start)
	wg.Wait()
}

// TestPostWettlauf (W69): von n gleichzeitigen POST mit demselben Schlüssel gelingt genau einer, die übrigen sind ODR-E6009, Put
// läuft genau einmal. 32 Goroutinen × 4 Runden; der Wirkungs-Fänger zählt, was das Fenster zuließe.
func TestPostWettlauf(t *testing.T) {
	for runde := 0; runde < 4; runde++ {
		u := schreibUmg(t)
		u.dynamisch()
		m := u.mocker()
		var r ergebnisse
		aufEinmal(32, func(i int) {
			r.nimm(m.Mock(u.ctx, dynAnfrage("POST", "7", fmt.Sprint("g", i), wsID)))
		})
		if r.stati[201] != 1 || r.codes[e6009] != 31 || r.anders != 0 || u.p.anzahl("Entities.Put") != 1 {
			t.Errorf("Runde %d: 201 ×%d, ODR-E6009 ×%d, anders %d, Put %d; erwartet 1, 31, 0, 1", runde, r.stati[201], r.codes[e6009], r.anders, u.p.anzahl("Entities.Put"))
		}
	}
}

// TestDeleteWettlauf (W70): von n gleichzeitigen DELETE derselben Entity gelingt genau eines (204), die übrigen sind ODR-E6008,
// Put einmal, Version plus 1. 32 Goroutinen × 4 Runden.
func TestDeleteWettlauf(t *testing.T) {
	for runde := 0; runde < 4; runde++ {
		u := schreibUmg(t)
		u.dynamisch()
		m := u.mocker()
		var r ergebnisse
		aufEinmal(32, func(i int) {
			r.nimm(m.Mock(u.ctx, dynAnfrage("DELETE", "2", "x", wsID)))
		})
		if v := u.zustand()["2"]; r.stati[204] != 1 || r.codes[e6008] != 31 || r.anders != 0 || u.p.anzahl("Entities.Put") != 1 || v.Version != 2 || !v.Deleted {
			t.Errorf("Runde %d: 204 ×%d, ODR-E6008 ×%d, Put %d, Version %d", runde, r.stati[204], r.codes[e6008], u.p.anzahl("Entities.Put"), v.Version)
		}
	}
}

// TestRevisionWettlauf (W71): n gleichzeitige erfolgreiche Mutationen eines Workspace auf verschiedenen Schlüsseln erhöhen die
// Revision um genau n, keine endet mit ODR-E9001, die entity_version jeder Entity stimmt. 32 Goroutinen × Methode 4 × 4 Runden.
func TestRevisionWettlauf(t *testing.T) {
	for _, meth := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		for runde := 0; runde < 4; runde++ {
			u := schreibUmg(t)
			u.dynamisch()
			var es []model.EntityState
			for k := 11; k < 43; k++ {
				es = append(es, voll(k, "alt"))
			}
			u.en.setze(setProducts, es...)
			m := u.mocker()
			var r ergebnisse
			aufEinmal(32, func(i int) {
				key := fmt.Sprint(11 + i)
				if meth == "POST" {
					key = fmt.Sprint(101 + i)
				}
				r.nimm(m.Mock(u.ctx, dynAnfrage(meth, key, "neu", wsID)))
			})
			if r.stati[sollStatus(meth)] != 32 || len(r.codes) != 0 || r.anders != 0 || u.ws.revision(wsID) != 32 {
				t.Errorf("%s Runde %d: Status %v Codes %v Revision %d, erwartet 32 × %d und Revision 32", meth, runde, r.stati, r.codes, u.ws.revision(wsID), sollStatus(meth))
			}
			want := int64(2)
			if meth == "POST" {
				want = 1
			}
			for k, x := range u.zustand() {
				if x.Version != want {
					t.Errorf("%s: Version von %s ist %d, erwartet %d", meth, k, x.Version, want)
				}
			}
		}
	}
}

// TestFensterZuordnung (W72): wird der Workspace nach dem Lesen in Schritt (f) gelöscht, ist die Mutation ODR-E7001 und schreibt
// nichts; setzt ein Reset ihn zurück, wirkt sie auf den Zustand nach dem Reset und folgt dessen Revision. Haltepunkt im Fake von
// Workspaces.Get (der Hook läuft nach dem ersten Aufruf), Ereignis 2 × Methode 4 = 8.
func TestFensterZuordnung(t *testing.T) {
	n := 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		for _, ereignis := range []string{"Löschen", "Reset"} {
			n++
			u := schreibUmg(t)
			vorher := u.en.inhalt()
			reset := gefuellt("2")
			reset.Version = 4
			u.ws.nach = func(k int) {
				if k != 1 {
					return
				}
				if ereignis == "Löschen" {
					u.ws.entferne(wsID)
					return
				}
				u.ws.setze(model.Workspace{ID: wsID, ServiceID: dienstID, BaselineSnapshotID: baselineID, Revision: 5})
				u.en.setze(setProducts, reset)
			}
			id := schluesselFuer(m)
			a, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
			kontext := m + " " + ereignis
			if ereignis == "Löschen" {
				pruefeFehler(t, kontext, err, e7001)
				u.ohneSchreiben(kontext)
				for set, liste := range vorher {
					if len(u.en.inhalt()[set]) != len(liste) {
						t.Errorf("%s: Zustand verändert", kontext)
					}
				}
				continue
			}
			erwarteStatus(t, kontext, a, err, sollStatus(m))
			adv := u.p.von("Workspaces.Advance")
			if len(adv) != 1 || adv[0].args[1] != int64(5) || u.ws.revision(wsID) != 6 {
				t.Errorf("%s: Advance %v, Revision %d, erwartet erwartet=5 und 6", kontext, adv, u.ws.revision(wsID))
			}
			want := int64(5)
			if m == "POST" {
				want = 1
			}
			if got := u.zustand()[id]; got.Version != want {
				t.Errorf("%s: Version %d, erwartet %d (Zustand nach dem Reset)", kontext, got.Version, want)
			}
		}
	}
	if n != 8 {
		t.Errorf("Menge %d, erwartet 8", n)
	}
}

// TestIsolationWorkspaces (W73, LH-QA-03.a): 20 Workspaces mit gleichen Sets und Schlüsseln, je eine eigene Folge von 12
// Mutationen, gleichzeitig; jeder endet in seinem Soll, keiner beobachtet den Zustand eines anderen.
func TestIsolationWorkspaces(t *testing.T) {
	u := neueUmg(t)
	u.en.jeWs = true
	u.c.mc.modell = testModell()
	u.co.setze(baselineID, setProducts, model.CoverageComplete)
	u.dynamisch()
	const anzahl = 20
	for i := 0; i < anzahl; i++ {
		u.ws.setze(model.Workspace{ID: fmt.Sprint("w", i), ServiceID: dienstID, BaselineSnapshotID: baselineID})
	}
	folge := []string{"POST", "PATCH", "PUT", "DELETE", "POST", "PATCH", "PUT", "PATCH", "DELETE", "POST", "PATCH", "PUT"}
	m := u.mocker()
	var fehler atomic.Int64
	aufEinmal(anzahl, func(i int) {
		ws := fmt.Sprint("w", i)
		for s, meth := range folge {
			if _, err := m.Mock(u.ctx, dynAnfrage(meth, "7", fmt.Sprintf("%s-%d", ws, s), ws)); err != nil {
				fehler.Add(1)
			}
		}
	})
	if fehler.Load() != 0 {
		t.Errorf("%d Mutationen endeten mit einem Fehler", fehler.Load())
	}
	for i := 0; i < anzahl; i++ {
		ws := fmt.Sprint("w", i)
		liste := u.en.ausWs(ws, setProducts)
		want := erwPayload(map[string]string{"Id": "7", "Name": fmt.Sprintf(`"%s-%d"`, ws, 11)})
		if len(liste) != 1 || liste[0].Payload != want || liste[0].Version != 12 || liste[0].Deleted || liste[0].WorkspaceID != ws || u.ws.revision(ws) != 12 {
			t.Errorf("%s: %+v Revision %d, erwartet %s Version 12 Revision 12", ws, liste, u.ws.revision(ws), want)
		}
	}
	if u.p.anzahl("Entities.Put") != anzahl*12 || u.p.anzahl("Workspaces.Advance") != anzahl*12 {
		t.Errorf("Put %d Advance %d, erwartet je %d", u.p.anzahl("Entities.Put"), u.p.anzahl("Workspaces.Advance"), anzahl*12)
	}
}

// TestLesenGegenSchreiben (W74): jede Antwort der Collection zeigt den Zustand vor oder nach einer Mutation, nie eine Mischung;
// ein ListSet je Antwort. 8 Leser × 200 Antworten, ein Schreiber wechselt per PUT zwischen A und B.
func TestLesenGegenSchreiben(t *testing.T) {
	u := schreibUmg(t)
	u.dynamisch()
	var es []model.EntityState
	for k := 1; k <= 40; k++ {
		es = append(es, voll(k, "A"))
	}
	u.en.setze(setProducts, es...)
	serial := func(liste []model.EntityState) string {
		var sb strings.Builder
		for _, x := range liste {
			sb.WriteString(x.EntityKey + "=" + x.Payload + "|")
		}
		return sb.String()
	}
	erlaubt := map[string]bool{serial(u.en.inhalt()[setProducts]): true}
	b := voll(1, "B")
	neu := append([]model.EntityState{b}, es[1:]...)
	erlaubt[serial(neu)] = true
	var mu sync.Mutex
	gesehen := map[string]int{}
	u.en.beobachter = func(liste []model.EntityState) {
		mu.Lock()
		defer mu.Unlock()
		gesehen[serial(liste)]++
	}
	m := u.mocker()
	var fertig atomic.Bool
	var schreiber sync.WaitGroup
	schreiber.Add(1)
	start := make(chan struct{})
	go func() {
		defer schreiber.Done()
		<-start
		for i := 0; ; i++ {
			name := map[bool]string{true: "A", false: "B"}[i%2 == 1]
			if _, err := m.Mock(u.ctx, dynAnfrage("PUT", "1", name, wsID)); err != nil {
				t.Errorf("PUT: %v", err)
				return
			}
			if fertig.Load() {
				return
			}
		}
	}()
	var leser sync.WaitGroup
	for l := 0; l < 8; l++ {
		leser.Add(1)
		go func() {
			defer leser.Done()
			<-start
			for k := 0; k < 200; k++ {
				a, err := m.Mock(u.ctx, dynAnfrage("GET", "x", "", wsID))
				if err != nil {
					t.Errorf("GET: %v", err)
					return
				}
				_, _ = io.Copy(io.Discard, a.Body)
				_ = a.Body.Close()
			}
		}()
	}
	close(start)
	leser.Wait()
	fertig.Store(true)
	schreiber.Wait()
	gesamt := 0
	for s, k := range gesehen {
		gesamt += k
		if !erlaubt[s] {
			t.Errorf("%d Antworten zeigen einen Zustand, der vor und nach keiner Mutation besteht", k)
		}
	}
	if gesamt != 1600 || u.p.anzahl("Entities.ListSet") != 1600 {
		t.Errorf("%d beobachtete Antworten, %d Aufrufe von ListSet, erwartet je 1600", gesamt, u.p.anzahl("Entities.ListSet"))
	}
}

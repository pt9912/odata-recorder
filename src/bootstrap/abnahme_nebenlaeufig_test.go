package bootstrap_test

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Nebenläufigkeit: Fenster der Mutation mit Haltepunkt (SPEC-049 Nr. 8, SPEC-192 Nr. 8). Der Haltepunkt öffnet das Fenster in jedem Lauf; ein
// Wirkungs-Fänger zählt, was das Fenster zuließe.

const abHaltFrist = 10 * time.Second

// abHalt hält jeden Eintreffenden, bis N eingetroffen sind (N gleich 0: bis zur Freigabe von Hand) oder die Frist abläuft.
type abHalt struct {
	N     int
	Frist time.Duration

	mu           sync.Mutex
	eingetroffen int
	abgelaufen   int
	gegeben      bool
	frei         chan struct{}
}

func abNeuerHalt(n int) *abHalt {
	return &abHalt{N: n, Frist: abHaltFrist, frei: make(chan struct{})}
}

func (h *abHalt) geben() {
	if !h.gegeben {
		h.gegeben = true
		close(h.frei)
	}
}

// halte zählt den Eintreffenden und hält ihn; ein Eintreffender, dessen Frist abläuft, wird gezählt und kehrt zurück.
func (h *abHalt) halte() {
	h.mu.Lock()
	h.eingetroffen++
	if h.N > 0 && h.eingetroffen >= h.N {
		h.geben()
	}
	h.mu.Unlock()
	select {
	case <-h.frei:
	case <-time.After(h.Frist):
		h.mu.Lock()
		h.abgelaufen++
		h.mu.Unlock()
	}
}

// Freigabe gibt alle zugleich frei.
func (h *abHalt) Freigabe() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.geben()
}

// Eingetroffen ist die Zahl der Eintreffenden bisher.
func (h *abHalt) Eingetroffen() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.eingetroffen
}

// Warte wartet, bis n eingetroffen sind, höchstens die Frist des Haltepunkts.
func (h *abHalt) Warte(n int) bool {
	ende := time.Now().Add(h.Frist)
	for time.Now().Before(ende) {
		if h.Eingetroffen() >= n {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}

// Pruefe meldet Fatal „Haltepunkt nicht erreicht“, wenn ein Eintreffender die Frist erreichte.
func (h *abHalt) Pruefe(m abMelder) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.abgelaufen > 0 {
		m.Fatalf("Haltepunkt nicht erreicht: %d von %d eingetroffen, %d nach %v abgelaufen", h.eingetroffen, h.N, h.abgelaufen, h.Frist)
	}
}

// abHaltWorkspaces umhüllt das WorkspaceRepository der Konfig: Get (Stelle S2, vor der Transaktion) hält.
type abHaltWorkspaces struct {
	outbound.WorkspaceRepository
	h *abHalt
}

func (w abHaltWorkspaces) Get(ctx context.Context, id string) (model.Workspace, error) {
	w.h.halte()
	return w.WorkspaceRepository.Get(ctx, id)
}

// abHaltEntities umhüllt das EntityRepository der Konfig: jede Operation hält.
type abHaltEntities struct {
	outbound.EntityRepository
	h *abHalt
}

func (e abHaltEntities) Put(ctx context.Context, x model.EntityState) error {
	e.h.halte()
	return e.EntityRepository.Put(ctx, x)
}

func (e abHaltEntities) Get(ctx context.Context, ws, set, key string) (model.EntityState, error) {
	e.h.halte()
	return e.EntityRepository.Get(ctx, ws, set, key)
}

func (e abHaltEntities) List(ctx context.Context, ws string) ([]model.EntityState, error) {
	e.h.halte()
	return e.EntityRepository.List(ctx, ws)
}

func (e abHaltEntities) ListSet(ctx context.Context, ws, set string) ([]model.EntityState, error) {
	e.h.halte()
	return e.EntityRepository.ListSet(ctx, ws, set)
}

func (e abHaltEntities) ReplaceAll(ctx context.Context, ws string, es []model.EntityState) error {
	e.h.halte()
	return e.EntityRepository.ReplaceAll(ctx, ws, es)
}

// abFenster ist die Umgebung eines Fensters: der Prüfling mit Haltepunkt an Workspaces (n) und an Entities (n).
type abFenster struct {
	*abUmgebung
	wsH, entH *abHalt
}

func abNeuesFenster(t *testing.T, n int) *abFenster {
	t.Helper()
	u := abFrisch(t)
	f := &abFenster{abUmgebung: u, wsH: abNeuerHalt(n), entH: abNeuerHalt(n)}
	u.mock, u.strikt = u.baueMock(func(w outbound.WorkspaceRepository) outbound.WorkspaceRepository { return abHaltWorkspaces{w, f.wsH} },
		func(e outbound.EntityRepository) outbound.EntityRepository { return abHaltEntities{e, f.entH} })
	return f
}

// gleichzeitig führt die n Anfragen auf n Goroutinen aus und liefert die Ergebnisse in der Folge von a.
func (f *abFenster) gleichzeitig(a []abAnfrage) []abIst {
	ist := make([]abIst, len(a))
	p := &abParallel{N: len(a)}
	p.lauf(func(k int) { ist[k] = abSende(context.Background(), f.mock, "ws-1", a[k]) })
	return ist
}

// zaehle zählt die Ergebnisse nach Status (Antworten) und Code (Fehler).
func abZaehle(ist []abIst) (status map[int]int, codes map[model.Code]int) {
	status, codes = map[int]int{}, map[model.Code]int{}
	for _, i := range ist {
		if i.F != nil {
			codes[i.F.Code]++
		} else if i.Err != nil {
			codes[0]++
		} else {
			status[i.Status]++
		}
	}
	return
}

// pruefeHalt prüft Fatal-Bedingung und Wirkungszählung: n Eintreffende am Workspace-Haltepunkt, keiner am Haltepunkt der Entities.
func (f *abFenster) pruefeHalt(t *testing.T, n int) {
	t.Helper()
	f.wsH.Pruefe(t)
	f.entH.Pruefe(t)
	if got := f.wsH.Eingetroffen(); got != n {
		t.Errorf("am Haltepunkt der Workspaces trafen %d ein, erwartet %d", got, n)
	}
	if got := f.entH.Eingetroffen(); got != 0 {
		t.Errorf("am Haltepunkt der Entities trafen %d ein, erwartet keiner (alle Zugriffe laufen über die Stores der Transaktion)", got)
	}
}

// TestAbnahmeNebenlaeufigPost (T71): 20 gleichzeitige POST mit gleichem Schlüssel: eine 201, 19 ODR-E6009, Version 1, Revision plus 1.
func TestAbnahmeNebenlaeufigPost(t *testing.T) {
	f := abNeuesFenster(t, 20)
	vorher := f.stand("ws-1")
	var a []abAnfrage
	for j := 0; j < 20; j++ {
		a = append(a, abAnfrage{Methode: "POST", Pfad: "/Products", Body: `{"Id":50,"Name":"R` + strconv.Itoa(j) + `"}`})
	}
	ist := f.gleichzeitig(a)
	f.pruefeHalt(t, 20)
	status, codes := abZaehle(ist)
	if status[201] != 1 || codes[model.CodeE6009] != 19 || len(status) != 1 || len(codes) != 1 {
		t.Errorf("Status %v Codes %v, erwartet ein 201 und 19 mal ODR-E6009", status, codes)
	}
	gewinner := ""
	for j, i := range ist {
		if i.F == nil && i.Status == 201 {
			gewinner = `"Name":"R` + strconv.Itoa(j) + `"`
		}
	}
	nachher := f.stand("ws-1")
	e := eintrag(nachher, "Products", "50")
	if e.Version != 1 || gewinner == "" || !strings.Contains(e.Payload, gewinner) {
		t.Errorf("Products 50: Version %d, Payload %s; erwartet Version 1 mit %s", e.Version, e.Payload, gewinner)
	}
	if nachher.Revision != vorher.Revision+1 || len(nachher.Eintraege) != len(vorher.Eintraege)+1 {
		t.Errorf("Revision %d (vorher %d), Einträge %d (vorher %d); erwartet plus 1", nachher.Revision, vorher.Revision, len(nachher.Eintraege), len(vorher.Eintraege))
	}
}

// TestAbnahmeNebenlaeufigDelete (T72): 20 gleichzeitige DELETE: eine 204, 19 ODR-E6008, Version 1, Revision plus 1.
func TestAbnahmeNebenlaeufigDelete(t *testing.T) {
	f := abNeuesFenster(t, 20)
	vorher := f.stand("ws-1")
	a := slices.Repeat([]abAnfrage{{Methode: "DELETE", Pfad: "/Products(1)"}}, 20)
	ist := f.gleichzeitig(a)
	f.pruefeHalt(t, 20)
	status, codes := abZaehle(ist)
	if status[204] != 1 || codes[model.CodeE6008] != 19 || len(status) != 1 || len(codes) != 1 {
		t.Errorf("Status %v Codes %v, erwartet eine 204 und 19 mal ODR-E6008", status, codes)
	}
	nachher := f.stand("ws-1")
	if e := eintrag(nachher, "Products", "1"); e.Version != 1 || !e.Deleted || nachher.Revision != vorher.Revision+1 {
		t.Errorf("Products 1: Version %d, gelöscht %v, Revision %d (vorher %d); erwartet Version 1, gelöscht, plus 1", e.Version, e.Deleted, nachher.Revision, vorher.Revision)
	}
}

// TestAbnahmeNebenlaeufigPatch (T73): 20 gleichzeitige PATCH: 20 mal 204, Revision plus 20, Version 20, kein ODR-E9001.
func TestAbnahmeNebenlaeufigPatch(t *testing.T) {
	f := abNeuesFenster(t, 20)
	vorher := f.stand("ws-1")
	var a []abAnfrage
	for j := 0; j < 20; j++ {
		a = append(a, abAnfrage{Methode: "PATCH", Pfad: "/Products(2)", Body: `{"Name":"M` + strconv.Itoa(j) + `"}`})
	}
	ist := f.gleichzeitig(a)
	f.pruefeHalt(t, 20)
	status, codes := abZaehle(ist)
	if status[204] != 20 || len(status) != 1 || len(codes) != 0 {
		t.Errorf("Status %v Codes %v, erwartet 20 mal 204 und kein ODR-E9001", status, codes)
	}
	nachher := f.stand("ws-1")
	e := eintrag(nachher, "Products", "2")
	namen := map[string]bool{}
	for j := 0; j < 20; j++ {
		namen[`"Name":"M`+strconv.Itoa(j)+`"`] = true
	}
	trifft := false
	for n := range namen {
		trifft = trifft || strings.Contains(e.Payload, n)
	}
	if e.Version != 20 || nachher.Revision != vorher.Revision+20 || !trifft {
		t.Errorf("Products 2: Version %d, Payload %s, Revision %d (vorher %d); erwartet Version 20, einen der Namen M0 bis M19, plus 20", e.Version, e.Payload, nachher.Revision, vorher.Revision)
	}
}

// einzeln führt einen PATCH auf einer Goroutine aus, die am Haltepunkt der Workspaces hält; fn läuft, während sie dort steht, dann die Freigabe.
func (f *abFenster) einzeln(t *testing.T, fn func()) abIst {
	t.Helper()
	ergebnis := make(chan abIst, 1)
	go func() {
		ergebnis <- abSende(context.Background(), f.mock, "ws-1", abAnfrage{Methode: "PATCH", Pfad: "/Products(1)", Body: `{"Name":"Z"}`})
	}()
	if !f.wsH.Warte(1) {
		f.wsH.Freigabe()
		<-ergebnis
		t.Fatal("Haltepunkt nicht erreicht: der PATCH kam nicht an die Stelle S2")
	}
	fn()
	f.wsH.Freigabe()
	return <-ergebnis
}

// TestAbnahmeNebenlaeufigLoeschen (T74): der Workspace wird am Haltepunkt gelöscht: ODR-E7001, die Ablage trägt keinen Eintrag des Workspace.
func TestAbnahmeNebenlaeufigLoeschen(t *testing.T) {
	f := abNeuesFenster(t, 0)
	ist := f.einzeln(t, func() {
		if _, err := f.ws.Delete(context.Background(), "ws-1"); err != nil {
			t.Errorf("Delete: %v", err)
		}
	})
	if ist.F == nil || ist.F.Code != model.CodeE7001 {
		t.Errorf("PATCH nach dem Löschen: Status %d, Fehler %v; erwartet ODR-E7001", ist.Status, ist.Err)
	}
	for _, set := range []string{"Labels", "Orders", "Products"} {
		es, err := f.db.Entities().ListSet(context.Background(), "ws-1", set)
		if err != nil && !errors.Is(err, outbound.ErrNichtGefunden) || len(es) != 0 {
			t.Errorf("die Ablage trägt %d Einträge von %s des gelöschten Workspace (%v)", len(es), set, err)
		}
	}
	if _, err := f.db.Workspaces().Get(context.Background(), "ws-1"); !errors.Is(err, outbound.ErrNichtGefunden) {
		t.Errorf("der Workspace besteht nach dem PATCH: %v", err)
	}
	f.wsH.Pruefe(t)
	f.entH.Pruefe(t)
}

// TestAbnahmeNebenlaeufigReset (T75): der Workspace wird am Haltepunkt zurückgesetzt: 204, Version 1, Name Z, Revision des Reset plus 1.
func TestAbnahmeNebenlaeufigReset(t *testing.T) {
	f := abNeuesFenster(t, 0)
	var revReset int64
	ist := f.einzeln(t, func() {
		w, err := f.ws.Reset(context.Background(), "ws-1")
		if err != nil {
			t.Errorf("Reset: %v", err)
		}
		revReset = w.Revision
	})
	if ist.F != nil || ist.Err != nil || ist.Status != 204 {
		t.Errorf("PATCH nach dem Reset: Status %d, Fehler %v; erwartet 204", ist.Status, ist.Err)
	}
	s := f.stand("ws-1")
	if e := eintrag(s, "Products", "1"); e.Version != 1 || !strings.Contains(e.Payload, `"Name":"Z"`) || s.Revision != revReset+1 || revReset != 1 {
		t.Errorf("Products 1: Version %d, Payload %s; Revision %d, Revision des Reset %d; erwartet Version 1, Name Z, Reset 1, danach 2", e.Version, e.Payload, s.Revision, revReset)
	}
	f.wsH.Pruefe(t)
	f.entH.Pruefe(t)
}

// TestAbnahmeHaltepunkt (T76): der Haltepunkt hält bis n eingetroffen sind und gibt dann alle zugleich frei; die Freigabe von Hand gibt frei; die Frist
// ist 10 s, ihr Ablauf ist Fatal.
func TestAbnahmeHaltepunkt(t *testing.T) {
	if abHaltFrist != 10*time.Second || abNeuerHalt(3).Frist != 10*time.Second {
		t.Fatalf("Frist %v, erwartet 10 s", abNeuerHalt(3).Frist)
	}
	h := abNeuerHalt(3)
	fertig := make(chan int, 3)
	gehe := func() {
		go func() { h.halte(); fertig <- h.Eingetroffen() }()
	}
	gehe()
	gehe()
	if !h.Warte(2) {
		t.Fatal("zwei Eintreffende nicht gezählt")
	}
	select {
	case n := <-fertig:
		t.Fatalf("ein Eintreffender kehrte zurück, als erst zwei von drei eingetroffen waren (%d)", n)
	case <-time.After(100 * time.Millisecond):
	}
	gehe()
	for i := 0; i < 3; i++ {
		select {
		case n := <-fertig:
			if n < 3 {
				t.Errorf("ein Eintreffender kehrte zurück, als erst %d von 3 eingetroffen waren", n)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("die Eintreffenden wurden bei n nicht freigegeben")
		}
	}
	// Freigabe von Hand.
	m := abNeuerHalt(0)
	hand := make(chan struct{})
	go func() { m.halte(); close(hand) }()
	m.Warte(1)
	select {
	case <-hand:
		t.Fatal("der Haltepunkt ohne n hielt nicht")
	case <-time.After(50 * time.Millisecond):
	}
	m.Freigabe()
	select {
	case <-hand:
	case <-time.After(5 * time.Second):
		t.Fatal("die Freigabe von Hand gab nicht frei")
	}
	// Ablauf der Frist: Fatal „Haltepunkt nicht erreicht“.
	k := abNeuerHalt(5)
	k.Frist = 20 * time.Millisecond
	ende := make(chan struct{})
	go func() { k.halte(); close(ende) }()
	select {
	case <-ende:
	case <-time.After(5 * time.Second):
		t.Fatal("die Frist des Haltepunkts lief nicht ab")
	}
	melder := &abTestMelder{}
	k.Pruefe(melder)
	if len(melder.fatal) != 1 || !strings.Contains(melder.fatal[0], "Haltepunkt nicht erreicht") {
		t.Errorf("Ablauf der Frist: %+v, erwartet Fatal „Haltepunkt nicht erreicht“", melder)
	}
	ok := &abTestMelder{}
	h.Pruefe(ok)
	if len(ok.fatal) != 0 {
		t.Errorf("Haltepunkt ohne Ablauf meldet Fatal: %+v", ok)
	}
}

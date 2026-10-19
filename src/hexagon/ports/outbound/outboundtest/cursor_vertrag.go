package outboundtest

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// RunCursorVertrag prüft für CursorRepository SPEC-108 Nr. 10 bis 12 und Nr. 17.
func RunCursorVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Put und Get tragen alle Felder", cursorFall1},
		{"Put ersetzt und trennt die Schlüssel", cursorFall2},
		{"Put lehnt Ungültiges ab", cursorFall3},
		{"Prüfreihenfolge von Put", cursorFall4},
		{"Get und List bei Unbekanntem und Leerem", cursorFall5},
		{"List ordnet nach Replay- und Recording-Session", cursorFall6},
		{"ResetAll setzt zurück", cursorFall7},
		{"das Löschen einer Recording-Session löscht ihre Cursors", cursorFall8},
		{"abgebrochener Kontext beendet jede Operation ohne Änderung", cursorKontextFall},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func cursor(ws, replay, aufnahme string) model.ReplayCursor {
	return model.ReplayCursor{WorkspaceID: ws, ReplaySessionID: replay, RecordingSessionID: aufnahme, ServiceID: "d",
		Strategy: model.StrategySequential, Position: 3, Consumed: []int64{5, 1, 3}}
}

// cursorBasis legt die Workspaces w und w2 (Service d) an; die Sessions s1 und s2 gehören zu d,
// x1 zu e.
func cursorBasis(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w", "w2")
	return v
}

func cursorIDs(l []model.ReplayCursor) []string {
	out := []string{}
	for _, c := range l {
		out = append(out, c.WorkspaceID+"/"+c.ReplaySessionID+"/"+c.RecordingSessionID)
	}
	return out
}

func cursorFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	for i, s := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
		c := cursor("w", "r", "s1")
		c.ReplaySessionID, c.Strategy = string(rune('a'+i)), s
		eingabe := append([]int64(nil), c.Consumed...)
		ok(t, v.Cursors.Put(ctx, c))
		if !reflect.DeepEqual(c.Consumed, eingabe) {
			t.Fatalf("Put änderte die Eingabe: %v", c.Consumed)
		}
		got, err := v.Cursors.Get(ctx, "w", c.ReplaySessionID, "s1")
		ok(t, err)
		c.Consumed = []int64{1, 3, 5}
		if !reflect.DeepEqual(got, c) {
			t.Fatalf("Get %+v, erwartet %+v", got, c)
		}
	}
	leer := cursor("w", "leer", "s1")
	leer.Position, leer.Consumed = 0, nil
	ok(t, v.Cursors.Put(ctx, leer))
	got, err := v.Cursors.Get(ctx, "w", "leer", "s1")
	ok(t, err)
	leer.Consumed = []int64{}
	if !reflect.DeepEqual(got, leer) {
		t.Fatalf("Cursor ohne Verbrauch: %+v", got)
	}
}

func cursorFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	ok(t, v.Cursors.Put(ctx, cursor("w", "r", "s1")))
	ok(t, v.Cursors.Put(ctx, cursor("w", "r", "s2")))
	ok(t, v.Cursors.Put(ctx, cursor("w", "r2", "s1")))
	ok(t, v.Cursors.Put(ctx, cursor("w2", "r", "s1")))
	neuer := cursor("w", "r", "s1")
	neuer.Strategy, neuer.Position, neuer.Consumed = model.StrategyLast, 1, []int64{9}
	ok(t, v.Cursors.Put(ctx, neuer))
	got, err := v.Cursors.Get(ctx, "w", "r", "s1")
	ok(t, err)
	if !reflect.DeepEqual(got, neuer) {
		t.Fatalf("ersetzter Cursor %+v", got)
	}
	for _, k := range [][3]string{{"w", "r", "s2"}, {"w", "r2", "s1"}, {"w2", "r", "s1"}} {
		g, err := v.Cursors.Get(ctx, k[0], k[1], k[2])
		ok(t, err)
		if g.Position != 3 || g.Strategy != model.StrategySequential {
			t.Fatalf("Put änderte %v: %+v", k, g)
		}
	}
	if l, _ := v.Cursors.List(ctx, "w"); len(l) != 3 {
		t.Fatalf("%d Cursors in w, erwartet 3", len(l))
	}
}

func cursorFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	fehlerhaft := map[string]func(*model.ReplayCursor){
		"ohne Workspace":         func(c *model.ReplayCursor) { c.WorkspaceID = "" },
		"ohne Replay-Session":    func(c *model.ReplayCursor) { c.ReplaySessionID = "" },
		"ohne Recording-Session": func(c *model.ReplayCursor) { c.RecordingSessionID = "" },
		"ohne Service":           func(c *model.ReplayCursor) { c.ServiceID = "" },
		"Strategie leer":         func(c *model.ReplayCursor) { c.Strategy = "" },
		"Strategie unbekannt":    func(c *model.ReplayCursor) { c.Strategy = "random" },
		"Position zu klein":      func(c *model.ReplayCursor) { c.Position = 2 },
		"Position zu groß":       func(c *model.ReplayCursor) { c.Position = 4 },
		"negative Position":      func(c *model.ReplayCursor) { c.Position = -1 },
		"consumed mit 0":         func(c *model.ReplayCursor) { c.Consumed = []int64{0, 1, 3} },
		"consumed negativ":       func(c *model.ReplayCursor) { c.Consumed = []int64{-2, 1, 3} },
		"consumed mit Doppelten": func(c *model.ReplayCursor) { c.Consumed = []int64{1, 3, 1} },
		"Position ohne consumed": func(c *model.ReplayCursor) { c.Consumed = nil },
		"consumed ohne Position": func(c *model.ReplayCursor) { c.Position = 0 },
	}
	for name, mut := range fehlerhaft {
		c := cursor("w", "r", "s1")
		mut(&c)
		if err := v.Cursors.Put(ctx, c); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	if l, err := v.Cursors.List(ctx, "w"); err != nil || len(l) != 0 {
		t.Fatalf("Ungültiges wurde gespeichert: %v %v", l, err)
	}
}

func cursorFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	for _, f := range []struct {
		name                        string
		workspace, aufnahme, dienst string
		position                    int64
		soll                        error
	}{
		{"ungültig vor unbekanntem Workspace", "x", "s1", "d", -1, model.ErrUngueltig},
		{"unbekannter Workspace vor unbekannter Session", "x", "y", "d", 0, outbound.ErrNichtGefunden},
		{"unbekannter Workspace vor fremdem Service", "x", "s1", "e", 0, outbound.ErrNichtGefunden},
		{"unbekannte Session vor fremdem Service", "w", "y", "e", 0, outbound.ErrNichtGefunden},
		{"fremder Service des Cursors", "w", "s1", "e", 0, outbound.ErrFremderService},
		{"Session eines anderen Service", "w", "x1", "d", 0, outbound.ErrFremderService},
		{"Session und Cursor im anderen Service", "w", "x1", "e", 0, outbound.ErrFremderService},
	} {
		c := cursor(f.workspace, "r", f.aufnahme)
		c.ServiceID, c.Position, c.Consumed = f.dienst, f.position, nil
		if f.position < 0 {
			c.Consumed = []int64{1}
		}
		if err := v.Cursors.Put(ctx, c); !errors.Is(err, f.soll) {
			t.Fatalf("%s: Fehler %v, erwartet %v", f.name, err, f.soll)
		}
	}
	if l, err := v.Cursors.List(ctx, "w"); err != nil || len(l) != 0 {
		t.Fatalf("Abgelehntes wurde gespeichert: %v %v", l, err)
	}
}

func cursorFall5(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	for _, k := range [][3]string{{"w", "r", "s1"}, {"x", "r", "s1"}, {"", "r", "s1"}, {"w", "", "s1"}, {"w", "r", ""}} {
		_, err := v.Cursors.Get(ctx, k[0], k[1], k[2])
		erwarte(t, err, outbound.ErrNichtGefunden)
	}
	l, err := v.Cursors.List(ctx, "w")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("Workspace ohne Cursors: %v %v", l, err)
	}
	for _, id := range []string{"gibt-es-nicht", ""} {
		_, err = v.Cursors.List(ctx, id)
		erwarte(t, err, outbound.ErrNichtGefunden)
	}
}

func cursorFall6(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	for _, k := range [][2]string{{"b", "s1"}, {"a", "s2"}, {"b", "s2"}, {"a", "s1"}, {"B", "s1"}} {
		ok(t, v.Cursors.Put(ctx, cursor("w", k[0], k[1])))
	}
	ok(t, v.Cursors.Put(ctx, cursor("w2", "0", "s1")))
	l, err := v.Cursors.List(ctx, "w")
	ok(t, err)
	if want := []string{"w/B/s1", "w/a/s1", "w/a/s2", "w/b/s1", "w/b/s2"}; !reflect.DeepEqual(cursorIDs(l), want) {
		t.Fatalf("Reihenfolge %v, erwartet %v", cursorIDs(l), want)
	}
}

func cursorFall7(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	erwarte(t, v.Cursors.ResetAll(ctx, "gibt-es-nicht"), outbound.ErrNichtGefunden)
	erwarte(t, v.Cursors.ResetAll(ctx, ""), outbound.ErrNichtGefunden)
	ok(t, v.Cursors.ResetAll(ctx, "w"))
	strategien := map[string]model.ReplayStrategy{"a": model.StrategyStrict, "b": model.StrategyFirst}
	for r, s := range strategien {
		c := cursor("w", r, "s1")
		c.Strategy = s
		ok(t, v.Cursors.Put(ctx, c))
	}
	ok(t, v.Cursors.Put(ctx, cursor("w", "a", "s2")))
	ok(t, v.Cursors.Put(ctx, cursor("w2", "a", "s1")))
	ok(t, v.Cursors.ResetAll(ctx, "w"))
	l, err := v.Cursors.List(ctx, "w")
	ok(t, err)
	if len(l) != 3 {
		t.Fatalf("ResetAll änderte die Zahl der Cursors: %v", cursorIDs(l))
	}
	for _, c := range l {
		if c.Position != 0 || c.Consumed == nil || len(c.Consumed) != 0 || c.ServiceID != "d" {
			t.Fatalf("Cursor nach ResetAll: %+v", c)
		}
		if soll, ok := strategien[c.ReplaySessionID]; ok && c.RecordingSessionID == "s1" && c.Strategy != soll {
			t.Fatalf("ResetAll änderte die Strategie: %+v", c)
		}
	}
	if g, _ := v.Cursors.Get(ctx, "w2", "a", "s1"); g.Position != 3 || len(g.Consumed) != 3 {
		t.Fatalf("ResetAll setzte einen anderen Workspace zurück: %+v", g)
	}
}

func cursorFall8(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := cursorBasis(t, neu)
	ok(t, v.Cursors.Put(ctx, cursor("w", "r", "s2")))
	ok(t, v.Cursors.Put(ctx, cursor("w", "r", "s1")))
	ok(t, v.Cursors.Put(ctx, cursor("w2", "r", "s2")))
	// s2 ist keine Baseline-Quelle und damit löschbar.
	ok(t, v.Sessions.Delete(ctx, "s2"))
	l, err := v.Cursors.List(ctx, "w")
	ok(t, err)
	if !reflect.DeepEqual(cursorIDs(l), []string{"w/r/s1"}) {
		t.Fatalf("Cursors nach dem Löschen von s2: %v", cursorIDs(l))
	}
	if v.Zeilen(t, "replay_cursors") != 1 {
		t.Fatalf("%d Cursor-Zeilen", v.Zeilen(t, "replay_cursors"))
	}
	if _, err := v.Workspaces.Get(ctx, "w2"); err != nil {
		t.Fatalf("Workspace nach dem Löschen: %v", err)
	}
}

func cursorKontextFall(t *testing.T, neu Neu) {
	bg, ctx := context.Background(), abgebrochen()
	v := cursorBasis(t, neu)
	ok(t, v.Cursors.Put(bg, cursor("w", "r", "s1")))
	neuer := cursor("w", "r", "s1")
	neuer.Position, neuer.Consumed = 1, []int64{4}
	erwarte(t, v.Cursors.Put(ctx, neuer), context.Canceled)
	_, err := v.Cursors.Get(ctx, "w", "r", "s1")
	erwarte(t, err, context.Canceled)
	_, err = v.Cursors.List(ctx, "w")
	erwarte(t, err, context.Canceled)
	erwarte(t, v.Cursors.ResetAll(ctx, "w"), context.Canceled)
	erwarte(t, v.Cursors.Put(ctx, model.ReplayCursor{}), model.ErrUngueltig)
	if g, err := v.Cursors.Get(bg, "w", "r", "s1"); err != nil || g.Position != 3 {
		t.Fatalf("Änderung trotz Abbruch: %+v %v", g, err)
	}
}

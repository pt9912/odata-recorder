package outboundtest

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// RunEntityVertrag prüft für EntityRepository SPEC-108 Nr. 6 bis 9 und Nr. 17.
func RunEntityVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Put und Get tragen alle Felder, Payload byte-gleich", entityFall1},
		{"Put ersetzt und trennt Workspaces, Mengen und Schlüssel", entityFall2},
		{"Put lehnt Ungültiges ab", entityFall3},
		{"Get, List und ListSet bei Unbekanntem, Leerem und Tombstones", entityFall4},
		{"List und ListSet ordnen nach Menge und Schlüssel", entityFall5},
		{"ReplaceAll ersetzt den State", entityFall6},
		{"ReplaceAll ist ganz oder gar nicht", entityFall7},
		{"ReplaceAll prüft in der Reihenfolge der Spezifikation", entityFall8},
		{"ReplaceAll ist atomar", entityReplaceAtomar},
		{"ReplaceAll prüft jede Art von ungültig vor Kontext und Workspace", entityFall9},
		{"Put ersetzt auch einen Tombstone und senkt die Version", entityFall10},
		{"abgebrochener Kontext beendet jede Operation ohne Änderung", entityKontextFall},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func entity(ws, menge, schluessel string) model.EntityState {
	return model.EntityState{WorkspaceID: ws, EntitySet: menge, EntityKey: schluessel,
		Payload: "{ \"ID\" : 1,\n \"Name\":\"ä\" }\n", KnownProperties: []string{"Name", "ID"},
		Completeness: model.EntityComplete, Version: 1}
}

// entityBasis legt die Workspaces w und w2 in Service d an.
func entityBasis(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w", "w2")
	return v
}

func entityKeys(l []model.EntityState) []string {
	out := []string{}
	for _, e := range l {
		out = append(out, e.EntitySet+"/"+e.EntityKey)
	}
	return out
}

func entityFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	e := entity("w", "Products", "1")
	eingabe := append([]string(nil), e.KnownProperties...)
	ok(t, v.Entities.Put(ctx, e))
	if !reflect.DeepEqual(e.KnownProperties, eingabe) {
		t.Fatalf("Put änderte die Eingabe: %v", e.KnownProperties)
	}
	tomb := entity("w", "Products", "2")
	tomb.Deleted, tomb.Completeness, tomb.Version, tomb.KnownProperties = true, model.EntityPartial, 7, nil
	tomb.Payload = "null"
	ok(t, v.Entities.Put(ctx, tomb))
	want := e
	want.KnownProperties = []string{"ID", "Name"}
	got, err := v.Entities.Get(ctx, "w", "Products", "1")
	ok(t, err)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Get %+v, erwartet %+v", got, want)
	}
	tomb.KnownProperties = []string{}
	got, err = v.Entities.Get(ctx, "w", "Products", "2")
	ok(t, err)
	if !reflect.DeepEqual(got, tomb) {
		t.Fatalf("Tombstone %+v, erwartet %+v", got, tomb)
	}
	l, err := v.Entities.List(ctx, "w")
	ok(t, err)
	if !reflect.DeepEqual(l, []model.EntityState{want, tomb}) {
		t.Fatalf("List %+v", l)
	}
	l, err = v.Entities.ListSet(ctx, "w", "Products")
	ok(t, err)
	if !reflect.DeepEqual(l, []model.EntityState{want, tomb}) {
		t.Fatalf("ListSet %+v", l)
	}
}

func entityFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	ok(t, v.Entities.Put(ctx, entity("w", "Products", "1")))
	ok(t, v.Entities.Put(ctx, entity("w2", "Products", "1")))
	ok(t, v.Entities.Put(ctx, entity("w", "Orders", "1")))
	ok(t, v.Entities.Put(ctx, entity("w", "Products", "10")))
	neu1 := entity("w", "Products", "1")
	neu1.Payload, neu1.Version, neu1.Deleted, neu1.Completeness = `{"neu":true}`, 2, true, model.EntityPartial
	neu1.KnownProperties = []string{"z"}
	ok(t, v.Entities.Put(ctx, neu1))
	got, err := v.Entities.Get(ctx, "w", "Products", "1")
	ok(t, err)
	if !reflect.DeepEqual(got, neu1) {
		t.Fatalf("ersetzter Eintrag %+v", got)
	}
	for _, k := range [][3]string{{"w2", "Products", "1"}, {"w", "Orders", "1"}, {"w", "Products", "10"}} {
		g, err := v.Entities.Get(ctx, k[0], k[1], k[2])
		ok(t, err)
		if g.Version != 1 || g.Deleted || g.Payload != entity("", "", "").Payload {
			t.Fatalf("Put änderte %v: %+v", k, g)
		}
	}
	if l, _ := v.Entities.List(ctx, "w"); len(l) != 3 {
		t.Fatalf("%d Einträge in w, erwartet 3 (Ersetzen, nicht Anhängen)", len(l))
	}
}

func entityFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	fehlerhaft := map[string]func(*model.EntityState){
		"ohne Workspace":            func(e *model.EntityState) { e.WorkspaceID = "" },
		"ohne entity_set":           func(e *model.EntityState) { e.EntitySet = "" },
		"ohne entity_key":           func(e *model.EntityState) { e.EntityKey = "" },
		"Payload leer":              func(e *model.EntityState) { e.Payload = "" },
		"Payload kein JSON":         func(e *model.EntityState) { e.Payload = "{" },
		"Payload kein UTF-8":        func(e *model.EntityState) { e.Payload = "\"\xff\"" },
		"Vollständigkeit leer":      func(e *model.EntityState) { e.Completeness = "" },
		"Vollständigkeit unbekannt": func(e *model.EntityState) { e.Completeness = "voll" },
		"negative Version":          func(e *model.EntityState) { e.Version = -1 },
		"leere Property":            func(e *model.EntityState) { e.KnownProperties = []string{"a", ""} },
		"doppelte Property":         func(e *model.EntityState) { e.KnownProperties = []string{"a", "b", "a"} },
		"Property kein UTF-8":       func(e *model.EntityState) { e.KnownProperties = []string{"\xff"} },
	}
	for name, mut := range fehlerhaft {
		e := entity("w", "Products", "1")
		mut(&e)
		if err := v.Entities.Put(ctx, e); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	if l, err := v.Entities.List(ctx, "w"); err != nil || len(l) != 0 {
		t.Fatalf("Ungültiges wurde gespeichert: %v %v", l, err)
	}
	erwarte(t, v.Entities.Put(ctx, entity("gibt-es-nicht", "Products", "1")), outbound.ErrNichtGefunden)
	// Die Gültigkeit geht dem Workspace voraus.
	ungueltig := entity("gibt-es-nicht", "Products", "1")
	ungueltig.Version = -1
	erwarte(t, v.Entities.Put(ctx, ungueltig), model.ErrUngueltig)
	ok(t, v.Entities.Put(ctx, func() model.EntityState { e := entity("w", "Products", "1"); e.Version = 0; return e }()))
}

func entityFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	for _, k := range [][3]string{{"w", "Products", "x"}, {"gibt-es-nicht", "Products", "1"}, {"", "Products", "1"}, {"w", "", "1"}, {"w", "Products", ""}} {
		_, err := v.Entities.Get(ctx, k[0], k[1], k[2])
		erwarte(t, err, outbound.ErrNichtGefunden)
	}
	l, err := v.Entities.List(ctx, "w")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("leerer State: %v %v", l, err)
	}
	l, err = v.Entities.ListSet(ctx, "w", "Products")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("leere Menge: %v %v", l, err)
	}
	for _, id := range []string{"gibt-es-nicht", ""} {
		_, err = v.Entities.List(ctx, id)
		erwarte(t, err, outbound.ErrNichtGefunden)
		_, err = v.Entities.ListSet(ctx, id, "Products")
		erwarte(t, err, outbound.ErrNichtGefunden)
		_, err = v.Entities.ListSet(ctx, id, "")
		erwarte(t, err, model.ErrUngueltig)
	}
	_, err = v.Entities.ListSet(ctx, "w", "")
	erwarte(t, err, model.ErrUngueltig)
	ok(t, v.Entities.Put(ctx, entity("w", "Products", "1")))
	l, err = v.Entities.ListSet(ctx, "w", "Orders")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("Menge ohne Einträge: %v %v", l, err)
	}
	if l, err := v.Entities.ListSet(ctx, "w2", "Products"); err != nil || len(l) != 0 {
		t.Fatalf("ListSet sah einen anderen Workspace: %v %v", l, err)
	}
}

func entityFall5(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	for _, k := range [][2]string{{"b", "2"}, {"a", "z"}, {"b", "10"}, {"a", "B"}, {"B", "a"}, {"a", "a"}} {
		ok(t, v.Entities.Put(ctx, entity("w", k[0], k[1])))
	}
	l, err := v.Entities.List(ctx, "w")
	ok(t, err)
	if want := []string{"B/a", "a/B", "a/a", "a/z", "b/10", "b/2"}; !reflect.DeepEqual(entityKeys(l), want) {
		t.Fatalf("List %v, erwartet %v", entityKeys(l), want)
	}
	l, err = v.Entities.ListSet(ctx, "w", "a")
	ok(t, err)
	if want := []string{"a/B", "a/a", "a/z"}; !reflect.DeepEqual(entityKeys(l), want) {
		t.Fatalf("ListSet %v, erwartet %v", entityKeys(l), want)
	}
}

func entityFall6(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	ok(t, v.Entities.Put(ctx, entity("w", "Products", "alt")))
	ok(t, v.Entities.Put(ctx, entity("w", "Products", "bleibt")))
	ok(t, v.Entities.Put(ctx, entity("w2", "Products", "alt")))
	neuer := entity("w", "Products", "bleibt")
	neuer.Version, neuer.Payload = 9, `[1,  2]`
	tomb := entity("w", "Orders", "t")
	tomb.Deleted = true
	ok(t, v.Entities.ReplaceAll(ctx, "w", []model.EntityState{tomb, neuer}))
	l, err := v.Entities.List(ctx, "w")
	ok(t, err)
	neuer.KnownProperties = []string{"ID", "Name"}
	tomb.KnownProperties = []string{"ID", "Name"}
	if !reflect.DeepEqual(l, []model.EntityState{tomb, neuer}) {
		t.Fatalf("State nach ReplaceAll: %+v", l)
	}
	if l, err := v.Entities.List(ctx, "w2"); err != nil || len(l) != 1 {
		t.Fatalf("ReplaceAll änderte einen anderen Workspace: %v %v", l, err)
	}
	ok(t, v.Entities.ReplaceAll(ctx, "w", nil))
	if l, err := v.Entities.List(ctx, "w"); err != nil || l == nil || len(l) != 0 {
		t.Fatalf("leere Liste leert den State nicht: %v %v", l, err)
	}
	ok(t, v.Entities.ReplaceAll(ctx, "w", []model.EntityState{}))
}

func entityFall7(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	alt := entity("w", "Products", "alt")
	ok(t, v.Entities.Put(ctx, alt))
	gut := entity("w", "Products", "gut")
	ungueltig := entity("w", "Products", "u")
	ungueltig.Payload = "{"
	fremd := entity("w2", "Products", "f")
	doppelt := entity("w", "Products", "gut")
	doppelt.Version = 5
	for name, liste := range map[string][]model.EntityState{
		"ungültiger Eintrag":              {gut, ungueltig},
		"Eintrag eines anderen Workspace": {gut, fremd},
		"doppelter Schlüssel":             {gut, doppelt},
		"ungültiger zuerst":               {ungueltig, gut},
	} {
		if err := v.Entities.ReplaceAll(ctx, "w", liste); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
		l, err := v.Entities.List(ctx, "w")
		ok(t, err)
		if keys := entityKeys(l); !reflect.DeepEqual(keys, []string{"Products/alt"}) {
			t.Fatalf("%s änderte den State: %v", name, keys)
		}
	}
	// Gleicher Schlüssel in verschiedenen Mengen ist kein Duplikat.
	ok(t, v.Entities.ReplaceAll(ctx, "w", []model.EntityState{entity("w", "A", "1"), entity("w", "B", "1")}))
}

func entityFall8(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	ungueltig := entity("gibt-es-nicht", "Products", "u")
	ungueltig.Payload = "{"
	erwarte(t, v.Entities.ReplaceAll(abgebrochen(), "gibt-es-nicht", []model.EntityState{ungueltig}), model.ErrUngueltig)
	erwarte(t, v.Entities.ReplaceAll(ctx, "gibt-es-nicht", []model.EntityState{ungueltig}), model.ErrUngueltig)
	erwarte(t, v.Entities.ReplaceAll(ctx, "gibt-es-nicht", []model.EntityState{entity("gibt-es-nicht", "P", "1")}), outbound.ErrNichtGefunden)
	erwarte(t, v.Entities.ReplaceAll(ctx, "gibt-es-nicht", nil), outbound.ErrNichtGefunden)
	erwarte(t, v.Entities.ReplaceAll(ctx, "", nil), outbound.ErrNichtGefunden)
	erwarte(t, v.Entities.ReplaceAll(abgebrochen(), "gibt-es-nicht", nil), context.Canceled)
	erwarte(t, v.Entities.ReplaceAll(abgebrochen(), "w", nil), context.Canceled)
}

func entityReplaceAtomar(t *testing.T, neu Neu) {
	bg := context.Background()
	v := entityBasis(t, neu)
	ok(t, v.Entities.Put(bg, entity("w", "Products", "alt")))
	ctx, stop := context.WithCancel(bg)
	v.BrichNachLoeschenAb(t, "entities", stop)
	erwarte(t, v.Entities.ReplaceAll(ctx, "w", []model.EntityState{entity("w", "Products", "neu")}), context.Canceled)
	l, err := v.Entities.List(bg, "w")
	if err != nil || !reflect.DeepEqual(entityKeys(l), []string{"Products/alt"}) {
		t.Fatalf("State nach abgebrochenem ReplaceAll: %v %v", entityKeys(l), err)
	}
}

func entityKontextFall(t *testing.T, neu Neu) {
	bg, ctx := context.Background(), abgebrochen()
	v := entityBasis(t, neu)
	ok(t, v.Entities.Put(bg, entity("w", "Products", "1")))
	erwarte(t, v.Entities.Put(ctx, entity("w", "Products", "2")), context.Canceled)
	_, err := v.Entities.Get(ctx, "w", "Products", "1")
	erwarte(t, err, context.Canceled)
	_, err = v.Entities.List(ctx, "w")
	erwarte(t, err, context.Canceled)
	_, err = v.Entities.ListSet(ctx, "w", "Products")
	erwarte(t, err, context.Canceled)
	erwarte(t, v.Entities.ReplaceAll(ctx, "w", []model.EntityState{entity("w", "Products", "3")}), context.Canceled)
	// Das Argument geht dem Kontext voraus.
	_, err = v.Entities.ListSet(ctx, "w", "")
	erwarte(t, err, model.ErrUngueltig)
	erwarte(t, v.Entities.Put(ctx, model.EntityState{}), model.ErrUngueltig)
	l, err := v.Entities.List(bg, "w")
	if err != nil || !reflect.DeepEqual(entityKeys(l), []string{"Products/1"}) {
		t.Fatalf("Änderung trotz Abbruch: %v %v", entityKeys(l), err)
	}
}

func entityFall9(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	gut := entity("gibt-es-nicht", "P", "1")
	ungueltig := entity("gibt-es-nicht", "P", "2")
	ungueltig.Version = -1
	for name, liste := range map[string][]model.EntityState{
		"Eintrag eines anderen Workspace": {gut, entity("w", "P", "9")},
		"doppelter Schlüssel":             {gut, gut},
		"ungültiger Eintrag":              {gut, ungueltig},
	} {
		for kontext, c := range map[string]context.Context{"abgebrochen": abgebrochen(), "unbekannter Workspace": ctx} {
			if err := v.Entities.ReplaceAll(c, "gibt-es-nicht", liste); !errors.Is(err, model.ErrUngueltig) || errors.Is(err, context.Canceled) {
				t.Fatalf("%s bei %s: Fehler %v", name, kontext, err)
			}
		}
	}
}

func entityFall10(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := entityBasis(t, neu)
	tomb := entity("w", "Products", "1")
	tomb.Deleted, tomb.Version, tomb.Completeness = true, 7, model.EntityPartial
	ok(t, v.Entities.Put(ctx, tomb))
	lebend := entity("w", "Products", "1")
	lebend.Version, lebend.KnownProperties, lebend.Payload = 1, nil, `{"neu":1}`
	ok(t, v.Entities.Put(ctx, lebend))
	got, err := v.Entities.Get(ctx, "w", "Products", "1")
	ok(t, err)
	lebend.KnownProperties = []string{}
	if !reflect.DeepEqual(got, lebend) {
		t.Fatalf("wiederbelebter Eintrag %+v, erwartet %+v", got, lebend)
	}
}

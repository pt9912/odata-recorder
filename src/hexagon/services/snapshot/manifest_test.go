package snapshot_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// kopf umschließt die Einträge es (schon in der Kanonform, mit Komma getrennt) mit dem Kopf des Manifests von Hand.
func kopf(n int, es, svc string) string {
	return fmt.Sprintf(`{"entities":[%s],"entity_count":%d,"manifest_version":1,"metadata_hash":"%s","service_id":"%s","snapshot_kind":"workspace"}`,
		es, n, metadatenHash(), svc)
}

// eintrag ist ein Eintrag der Kanonform von Hand: sieben Glieder aufsteigend.
func eintrag(compl string, del bool, set, payload, props, key string, version int64) string {
	return fmt.Sprintf(`{"completeness":"%s","deleted":%t,"entity_set":"%s","payload":%s,"selected_properties":%s,"typed_key":"%s","version":%d}`,
		compl, del, set, payload, props, key, version)
}

func erzeugeMitState(t *testing.T, es []model.EntityState) (*welt, []byte, model.WorkspaceSnapshot) {
	t.Helper()
	w := neu(t)
	w.ent.setze("ws-1", es)
	s, b, err := w.erzeuge("n")
	if err != nil {
		t.Fatal(err)
	}
	return w, b, s
}

// TestManifestForm (T07): genau die Glieder entities, entity_count, manifest_version, metadata_hash, service_id und
// snapshot_kind in der Kanonform ohne Leerraum, Mitglieder aufsteigend; die Bytes stehen von Hand da (leer, ein Eintrag,
// drei Mengen).
func TestManifestForm(t *testing.T) {
	cat := model.EntityState{WorkspaceID: "ws-1", EntitySet: "Categories", EntityKey: "Id=1", Payload: `{"Id":"1"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 2}
	ord := model.EntityState{WorkspaceID: "ws-1", EntitySet: "Orders", EntityKey: "Id=5", Payload: `{"Id":"5","Sum":"3"}`, KnownProperties: []string{"Id", "Sum"}, Completeness: model.EntityPartial, Version: 9}
	prod := model.EntityState{WorkspaceID: "ws-1", EntitySet: "Products", EntityKey: "Id=1", Payload: `{"Id":"1"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 3}
	faelle := []struct {
		name string
		es   []model.EntityState
		soll string
	}{
		{"leer", nil, kopf(0, "", "svc")},
		{"ein Eintrag", []model.EntityState{prod}, kopf(1, eintrag("complete", false, "Products", `{"Id":"1"}`, `["Id"]`, "Id=1", 3), "svc")},
		{"drei Mengen", []model.EntityState{cat, ord, prod}, kopf(3, strings.Join([]string{
			eintrag("complete", false, "Categories", `{"Id":"1"}`, `["Id"]`, "Id=1", 2),
			eintrag("partial", false, "Orders", `{"Id":"5","Sum":"3"}`, `["Id","Sum"]`, "Id=5", 9),
			eintrag("complete", false, "Products", `{"Id":"1"}`, `["Id"]`, "Id=1", 3)}, ","), "svc")},
	}
	for _, f := range faelle {
		_, b, _ := erzeugeMitState(t, f.es)
		if string(b) != f.soll {
			t.Errorf("%s:\n%s\nerwartet\n%s", f.name, b, f.soll)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(b, &m); err != nil || len(m) != 6 {
			t.Errorf("%s: Glieder %d (%v)", f.name, len(m), err)
		}
	}
}

// TestManifestZeichen (T07, T08): Zeichenketten der Kanonform: `\"` und `\\`, `\u00xx` mit kleinen Hexziffern für jedes
// Byte unter 0x20, sonst unverändert (0x7F und Mehrbyte-Zeichen bleiben Rohbytes); gilt für entity_set, typed_key und
// jede Property. Die Bytes entstehen aus Bytewerten, die Escape-Folge mit esc.
func TestManifestZeichen(t *testing.T) {
	roh := "K\"" + `\` + string([]byte{0x1F, 0x0A, 0x7F, 0xC3, 0xA9})
	soll := `K\"\\` + esc("u001f") + esc("u000a") + string([]byte{0x7F, 0xC3, 0xA9})
	e := model.EntityState{WorkspaceID: "ws-1", EntitySet: roh, EntityKey: roh + "k", Payload: `{}`, KnownProperties: []string{roh}, Completeness: model.EntityComplete}
	_, b, _ := erzeugeMitState(t, []model.EntityState{e})
	erwartet := kopf(1, eintrag("complete", false, soll, `{}`, `["`+soll+`"]`, soll+"k", 0), "svc")
	if string(b) != erwartet {
		t.Errorf("%q\nerwartet\n%q", b, erwartet)
	}
	if bytes.Contains(b, []byte{0x1F}) || bytes.Contains(b, []byte{0x0A}) {
		t.Error("Steuerzeichen als Rohbyte im Manifest")
	}
}

// TestManifestEintraege (T08): M2, Art 4 × Version 4 × Maske 3 = 48; die Glieder jedes Eintrags tragen den Wert des
// Zustands (Orakel von Hand, Properties aufsteigend).
func TestManifestEintraege(t *testing.T) {
	arten := []struct {
		c   model.Completeness
		del bool
	}{{model.EntityComplete, false}, {model.EntityPartial, false}, {model.EntityComplete, true}, {model.EntityPartial, true}}
	versionen := []int64{0, 1, 7, 1 << 40}
	masken := []struct {
		in   []string
		soll string
	}{{[]string{}, `[]`}, {[]string{"Id"}, `["Id"]`}, {[]string{"Z", "M", "A"}, `["A","M","Z"]`}}
	n := 0
	for _, a := range arten {
		for _, v := range versionen {
			for _, m := range masken {
				e := model.EntityState{WorkspaceID: "ws-1", EntitySet: "S", EntityKey: "K", Payload: `{"x":1}`, KnownProperties: m.in, Completeness: a.c, Version: v, Deleted: a.del}
				_, b, _ := erzeugeMitState(t, []model.EntityState{e})
				soll := kopf(1, eintrag(string(a.c), a.del, "S", `{"x":1}`, m.soll, "K", v), "svc")
				if string(b) != soll {
					t.Errorf("%v %d %v:\n%s\nerwartet\n%s", a, v, m.in, b, soll)
				}
				n++
			}
		}
	}
	if n != 48 {
		t.Errorf("M2: %d Elemente", n)
	}
}

// payloads sind die sieben Arten von M3; die Bytes entstehen aus Bytewerten und Raw-Strings, die Escape-Folge mit esc.
func payloads() []string {
	return []string{
		`{"a":1}`,
		`{ "a" :  1 }`,
		`[1, 2]`,
		`"x"`,
		`1.50`,
		`null`,
		`{"s":"` + `\"` + `\\` + esc("u0001") + string([]byte{0xC3, 0xA9}) + `<>&"}`,
	}
}

// TestPayloadByteGleich (T09): payload steht byte-gleich, wie der Port ihn lieferte, jede JSON-Art, auch bei einem
// Tombstone (M3: Art 7 × Zustand 2 = 14).
func TestPayloadByteGleich(t *testing.T) {
	n := 0
	for _, p := range payloads() {
		for _, del := range []bool{false, true} {
			e := model.EntityState{WorkspaceID: "ws-1", EntitySet: "S", EntityKey: "K", Payload: p, Completeness: model.EntityComplete, Deleted: del}
			_, b, _ := erzeugeMitState(t, []model.EntityState{e})
			if !bytes.Contains(b, []byte(`"payload":`+p+`,"selected_properties"`)) {
				t.Errorf("Payload %q Tombstone %v: nicht byte-gleich in %q", p, del, b)
			}
			n++
		}
	}
	if n != 14 {
		t.Errorf("M3: %d Elemente", n)
	}
	// Die Rohbytes der Fixture: Gegenlesen der Escape-Folge als Zeichenfolge Backslash, u, 0001.
	if !strings.Contains(payloads()[6], "\\"+"u0001") || strings.ContainsRune(payloads()[6], 1) {
		t.Error("Fixture trägt das Steuerzeichen statt der Escape-Folge")
	}
}

// TestTombstoneManifest (T10): Tombstones stehen im Manifest mit deleted wahr und unveränderten payload,
// selected_properties, completeness und version; Position 3 (erster, mittlerer, letzter bei 5 Einträgen) × Completeness 2.
func TestTombstoneManifest(t *testing.T) {
	n := 0
	for _, pos := range []int{0, 2, 4} {
		for _, c := range []model.Completeness{model.EntityComplete, model.EntityPartial} {
			var es []model.EntityState
			var teile []string
			for i := 0; i < 5; i++ {
				key := fmt.Sprintf("K%d", i)
				e := model.EntityState{WorkspaceID: "ws-1", EntitySet: "S", EntityKey: key, Payload: fmt.Sprintf(`{"i":%d}`, i), KnownProperties: []string{"i"},
					Completeness: c, Version: int64(10 + i), Deleted: i == pos}
				es = append(es, e)
				teile = append(teile, eintrag(string(c), i == pos, "S", e.Payload, `["i"]`, key, e.Version))
			}
			_, b, _ := erzeugeMitState(t, es)
			if string(b) != kopf(5, strings.Join(teile, ","), "svc") {
				t.Errorf("Position %d %s:\n%s", pos, c, b)
			}
			n++
		}
	}
	if n != 6 {
		t.Errorf("%d Elemente", n)
	}
}

// TestManifestFolge (T11): die Folge der Einträge ist aufsteigend nach (entity_set, typed_key) in Bytes, die der
// Properties aufsteigend, unabhängig von der Folge des Ports (M4: Folge 6 × Properties 2 = 12); derselbe Schlüssel in
// drei Mengen sind drei Einträge.
func TestManifestFolge(t *testing.T) {
	sortiert := []struct{ set, key string }{{"Alpha", "1"}, {"Alpha", "10"}, {"Gamma", "1"}, {"Gamma", "10"}, {"beta", "1"}, {"beta", "10"}}
	var teile []string
	for _, s := range sortiert {
		teile = append(teile, eintrag("complete", false, s.set, `{}`, `["P","Q"]`, s.key, 1))
	}
	soll := kopf(6, strings.Join(teile, ","), "svc")
	folgen := [][]int{{0, 1, 2, 3, 4, 5}, {5, 4, 3, 2, 1, 0}, {3, 0, 5, 1, 4, 2}, {2, 4, 1, 5, 0, 3}, {1, 0, 3, 2, 5, 4}, {5, 3, 4, 1, 2, 0}}
	n := 0
	for fi, f := range folgen {
		for _, props := range [][]string{{"P", "Q"}, {"Q", "P"}} {
			var es []model.EntityState
			for _, i := range f {
				es = append(es, model.EntityState{WorkspaceID: "ws-1", EntitySet: sortiert[i].set, EntityKey: sortiert[i].key, Payload: `{}`,
					KnownProperties: slices.Clone(props), Completeness: model.EntityComplete, Version: 1})
			}
			_, b, _ := erzeugeMitState(t, es)
			if string(b) != soll {
				t.Errorf("Folge %d Properties %v:\n%s", fi, props, b)
			}
			n++
		}
	}
	if n != 12 {
		t.Errorf("M4: %d Elemente", n)
	}
}

// aenderungen sind die sieben Glieder eines Eintrags, jedes um einen Wert verändert (M4: Änderung eines der 7 Glieder).
func aenderungen() []func(e *model.EntityState) {
	return []func(e *model.EntityState){
		func(e *model.EntityState) {
			e.Completeness = map[model.Completeness]model.Completeness{model.EntityComplete: model.EntityPartial, model.EntityPartial: model.EntityComplete}[e.Completeness]
		},
		func(e *model.EntityState) { e.Deleted = !e.Deleted },
		func(e *model.EntityState) { e.EntitySet += "X" },
		func(e *model.EntityState) { e.Payload = `{"geaendert":true}` },
		func(e *model.EntityState) { e.KnownProperties = append(slices.Clone(e.KnownProperties), "Neu") },
		func(e *model.EntityState) { e.EntityKey += "X" },
		func(e *model.EntityState) { e.Version += 100 },
	}
}

func siebenEintraege() []model.EntityState {
	var es []model.EntityState
	for i := 0; i < 7; i++ {
		es = append(es, model.EntityState{WorkspaceID: "ws-1", EntitySet: []string{"A", "B", "C"}[i%3], EntityKey: fmt.Sprintf("K%d", i), Payload: fmt.Sprintf(`{"i":%d}`, i),
			KnownProperties: []string{"i"}, Completeness: model.EntityComplete, Version: int64(i)})
	}
	return es
}

// TestManifestDeterministisch (T12): derselbe State ergibt dieselben Bytes, dieselbe Referenz und denselben state_hash
// unabhängig von Revision, Name, Kennung des Snapshots, Zeit und Kennung des Workspace (M5, sechs Quellen); service_id und
// metadata_hash ändern die Referenz, ebenso die Änderung jedes der 7 Glieder an 3 Positionen (21 Läufe); 50 Wiederholungen.
func TestManifestDeterministisch(t *testing.T) {
	w0, b0, s0 := erzeugeMitState(t, siebenEintraege())
	gleich := map[string]func(w *welt){
		"Revision": func(w *welt) {
			w.ws.mu.Lock()
			x := w.ws.ws["ws-1"]
			x.Revision = 99
			w.ws.ws["ws-1"] = x
			w.ws.mu.Unlock()
		},
		"Zeit":    func(w *welt) { w.festUhr.Advance(48 * 3600 * 1e9) },
		"Kennung": func(w *welt) { w.festIDs.SnapshotID(); w.festIDs.SnapshotID() },
		"Erzeugung des Workspace": func(w *welt) {
			w.ws.mu.Lock()
			x := w.ws.ws["ws-1"]
			x.CreatedAt = x.CreatedAt.AddDate(0, 0, -3)
			w.ws.ws["ws-1"] = x
			w.ws.mu.Unlock()
		},
	}
	n := 0
	for name, f := range gleich {
		w := neu(t)
		w.ent.setze("ws-1", siebenEintraege())
		f(w)
		s, b, err := w.erzeuge("anderer Name")
		if err != nil || !bytes.Equal(b, b0) || s.StateBlob != s0.StateBlob || s.StateHash != s0.StateHash {
			t.Errorf("%s: %v ändert Bytes oder Referenz", name, err)
		}
		n++
	}
	// Kennung des Workspace: derselbe State unter ws-2.
	w := neu(t)
	es := siebenEintraege()
	for i := range es {
		es[i].WorkspaceID = "ws-2"
	}
	w.ent.setze("ws-2", es)
	s, err := w.d.Create(context.Background(), "ws-2", "n")
	if err != nil || !bytes.Equal(w.manifestBytes(s.StateBlob), b0) || s.StateBlob != s0.StateBlob {
		t.Errorf("Kennung des Workspace: %v", err)
	}
	n += 2
	// Name: in demselben Workspace unter 50 Namen.
	for i := 0; i < 50; i++ {
		s, err := w0.d.Create(context.Background(), "ws-1", fmt.Sprintf("name-%d", i))
		if err != nil || s.StateBlob != s0.StateBlob || s.StateHash != s0.StateHash || s.ID == s0.ID {
			t.Fatalf("Wiederholung %d: %v", i, err)
		}
	}
	if n != 6 {
		t.Errorf("M5: %d gleichbleibende Quellen, erwartet 6", n)
	}
	// service_id und metadata_hash ändern die Referenz.
	w = neu(t)
	w.ent.setze("ws-1", siebenEintraege())
	w.ws.mu.Lock()
	x := w.ws.ws["ws-1"]
	x.ServiceID = "svc2"
	w.ws.ws["ws-1"] = x
	w.ws.mu.Unlock()
	if s, _, err := w.erzeuge("n"); err != nil || s.StateBlob == s0.StateBlob {
		t.Errorf("service_id: %v", err)
	}
	w = neu(t)
	w.ent.setze("ws-1", siebenEintraege())
	w.snaps.mu.Lock()
	bl := w.snaps.baselines["bsl-1"]
	bl.MetadataHash = "sha256:" + strings.Repeat("c", 64)
	w.snaps.baselines["bsl-1"] = bl
	w.snaps.mu.Unlock()
	if s, _, err := w.erzeuge("n"); err != nil || s.StateBlob == s0.StateBlob || s.MetadataHash != bl.MetadataHash {
		t.Errorf("metadata_hash: %v", err)
	}
	// Änderung eines Gliedes an drei Positionen.
	laeufe := 0
	for _, pos := range []int{0, 3, 6} {
		for gi, f := range aenderungen() {
			es := siebenEintraege()
			f(&es[pos])
			if s, _, err := func() (model.WorkspaceSnapshot, []byte, error) {
				_, b, s := erzeugeMitState(t, es)
				return s, b, nil
			}(); err != nil || s.StateBlob == s0.StateBlob {
				t.Errorf("Position %d Glied %d: Referenz unverändert", pos, gi)
			}
			laeufe++
		}
	}
	if laeufe != 21 {
		t.Errorf("M4: %d Läufe", laeufe)
	}
}

// TestKeinCursorImManifest (T13): Cursors sind nie Bestandteil: das Manifest hat genau die sechs Glieder, die Konfig kein
// Feld für Cursors, und die Bytes tragen keine Kennung und keine Position eines Cursors (M15: 3 × 3 = 9 Stände neben dem
// State der Fixture).
func TestKeinCursorImManifest(t *testing.T) {
	_, soll, _ := erzeugeMitState(t, standard("ws-1"))
	n := 0
	for _, anzahl := range []int{0, 1, 3} {
		for _, stand := range []int{0, 5, 99} {
			var cursors []model.ReplayCursor
			for i := 0; i < anzahl; i++ {
				cursors = append(cursors, model.ReplayCursor{WorkspaceID: "ws-1", ReplaySessionID: fmt.Sprintf("rep-CANARYCUR%d", i), RecordingSessionID: "ses-CANARYREC",
					Strategy: model.StrategySequential, Position: int64(stand)})
			}
			_, b, _ := erzeugeMitState(t, standard("ws-1"))
			if !bytes.Equal(b, soll) || bytes.Contains(b, []byte("CANARY")) || bytes.Contains(bytes.ToLower(b), []byte("cursor")) {
				t.Errorf("Anzahl %d Stand %d (%d Cursors): Bytes weichen ab oder nennen einen Cursor", anzahl, stand, len(cursors))
			}
			n++
		}
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(soll, &m); err != nil || len(m) != 6 {
		t.Errorf("Glieder %d (%v)", len(m), err)
	}
	for i, typ := 0, reflect.TypeOf(neu(t).konfig); i < typ.NumField(); i++ {
		if _, hat := typ.Field(i).Type.MethodByName("ResetAll"); hat {
			t.Errorf("Feld %s kann Cursors", typ.Field(i).Name)
		}
	}
	if n != 9 {
		t.Errorf("M15: %d Elemente", n)
	}
}

// TestLeererWorkspace (T14): ein leerer Workspace ergibt einen leeren, gültigen Snapshot, ob der Port eine leere Liste
// oder nil liefert: entities ist [] (nie null), entity_count 0, kein Sonderfall.
func TestLeererWorkspace(t *testing.T) {
	for _, nilLeer := range []bool{false, true} {
		w := neu(t)
		w.ent.nilLeer = nilLeer
		w.ent.setze("ws-1", []model.EntityState{})
		s, b, err := w.erzeuge("leer")
		if err != nil {
			t.Fatalf("nil=%v: %v", nilLeer, err)
		}
		if string(b) != kopf(0, "", "svc") || s.Validate() != nil {
			t.Errorf("nil=%v: %s (%v)", nilLeer, b, s.Validate())
		}
	}
}

package workspace_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// genEnts erzeugt n Einträge in sets Mengen; die Wertarten wechseln je Eintrag: string, boolean, int32 /
// int64 über 2^53 und decimal (als JSON-String) / guid, date, datetimeoffset / null; jeder dritte trägt einen
// Typ, der vom Typ der Menge abweicht (@odata.type); complete und partial wechseln.
func genEnts(n, sets int) []ent {
	namen := []string{"Products", "Categories", "Orders"}[:sets]
	var r []ent
	for i := 0; i < n; i++ {
		var p obj
		switch i % 4 {
		case 0:
			p = obj{"Name": "text", "Flag": true, "Count": 7}
		case 1:
			p = obj{"Big": "9007199254740993", "Dec": "12.50", "Neg": "-9223372036854775808"}
		case 2:
			p = obj{"G": "01234567-89ab-cdef-0123-456789abcdef", "D": "2026-10-20", "T": "2026-10-20T10:00:00Z"}
		default:
			p = obj{"N": nil, "Name": "x"}
		}
		if i%3 == 0 {
			p["@odata.type"] = "#Model.Special"
		}
		var props []string
		for k := range p {
			if k != "@odata.type" {
				props = append(props, k)
			}
		}
		sort.Strings(props)
		compl := "complete"
		if i%2 == 1 {
			compl = "partial"
		}
		r = append(r, ent{namen[i%sets], fmt.Sprintf("Id=%d", i), schreibe(p), props, compl})
	}
	return r
}

// dekodiert liest den Payload mit einem eigenen Decoder (Zahlen als Text) und schreibt ihn in der Kanonform
// zurück: das Orakel für den Wert, unabhängig vom Dienst.
func dekodiert(t *testing.T, payload string) string {
	t.Helper()
	d := json.NewDecoder(strings.NewReader(payload))
	d.UseNumber()
	var m map[string]any
	if err := d.Decode(&m); err != nil {
		t.Fatalf("Payload %q: %v", payload, err)
	}
	var zahl func(v any) any
	zahl = func(v any) any {
		switch x := v.(type) {
		case json.Number:
			return roh(x.String())
		case map[string]any:
			o := obj{}
			for k, e := range x {
				o[k] = zahl(e)
			}
			return o
		case []any:
			l := make([]any, 0, len(x))
			for _, e := range x {
				l = append(l, zahl(e))
			}
			return l
		}
		return v
	}
	return schreibe(zahl(m))
}

// angelegterState liefert den Entity-State, den Create an ReplaceAll gab.
func angelegterState(t *testing.T, w *welt) (string, []model.EntityState) {
	t.Helper()
	ev := w.p.von("tx Entities.ReplaceAll")
	if len(ev) != 1 {
		t.Fatalf("ReplaceAll %d mal", len(ev))
	}
	return ev[0].args[0].(string), ev[0].args[1].([]model.EntityState)
}

// TestAbbildungGenerator (Z25): für 0, 1, 2 und 17 Einträge in 1 bis 3 Mengen mit allen Wertarten ist der
// Entity-State, den Create an ReplaceAll gibt, gleich dem erwarteten: Kennung, Menge, Schlüssel, Payload
// byte-gleich, Properties, Vollständigkeit, Version 0, nicht gelöscht, in der Folge des Manifests,
// Validate() erfüllt.
func TestAbbildungGenerator(t *testing.T) {
	zellen := [][2]int{{0, 1}, {1, 1}, {2, 1}, {2, 2}, {17, 1}, {17, 2}, {17, 3}}
	summe := 0
	arten := map[string]bool{}
	for _, z := range zellen {
		es := genEnts(z[0], z[1])
		summe += len(es)
		w := neu(t)
		w.baseline("bsl-1", "svc", es)
		ws, err := w.v.Create(context.Background(), "bsl-1", "ws-neu")
		if err != nil {
			t.Fatalf("%v: %v", z, err)
		}
		wid, got := angelegterState(t, w)
		want := erwarteterState("ws-neu", es)
		if wid != "ws-neu" || ws.ID != "ws-neu" || !reflect.DeepEqual(got, want) {
			t.Errorf("%v: State %+v, erwartet %+v", z, got, want)
		}
		for i, e := range got {
			if err := e.Validate(); err != nil {
				t.Errorf("%v[%d]: Validate: %v", z, i, err)
			}
			if e.WorkspaceID != "ws-neu" || e.Version != 0 || e.Deleted {
				t.Errorf("%v[%d]: %+v", z, i, e)
			}
			if dekodiert(t, e.Payload) != e.Payload {
				t.Errorf("%v[%d]: Payload nicht byte-gleich dem Orakel", z, i)
			}
			for _, a := range []string{"Name", "Flag", "Count", "Big", "Dec", "Neg", "G", "D", "T", "N", "@odata.type"} {
				if strings.Contains(e.Payload, `"`+a+`"`) {
					arten[a] = true
				}
			}
		}
		if ev := w.p.von("tx Workspaces.Create"); len(ev) != 1 || ev[0].args[0].(model.Workspace).ID != "ws-neu" {
			t.Errorf("%v: Create %v", z, ev)
		}
	}
	if summe != 0+1+2+2+17+17+17 {
		t.Errorf("Einträge %d", summe)
	}
	if len(arten) != 11 {
		t.Errorf("Wertarten %d von 11: %v", len(arten), arten)
	}
}

// TestPayloadByteGleich (Z25): die Payload-Bytes bleiben, wie das Manifest sie trägt: Leerraum, die Schreibweise
// von Zahlen und Zahlen über 2^53 ändern sich nicht (kein Umweg über float64 oder eine Neukodierung).
func TestPayloadByteGleich(t *testing.T) {
	payloads := []string{
		`{ "A" : 12345678901234567890 , "B" : 1.50 , "C" : 1e400 }`,
		`{"Z":1,"A":2}`,
		"{\n\t\"a\": [1, 2,\t3]\n}",
		`{"big":9007199254740993,"neg":-0.0}`,
	}
	es := make([]ent, 0, len(payloads))
	for i, p := range payloads {
		es = append(es, ent{"Products", fmt.Sprintf("Id=%d", i), p, []string{"a"}, "complete"})
	}
	w := neu(t)
	w.baseline("bsl-1", "svc", es)
	if _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); err != nil {
		t.Fatal(err)
	}
	_, got := angelegterState(t, w)
	for i, e := range got {
		if e.Payload != payloads[i] {
			t.Errorf("Payload %d: %q, erwartet %q", i, e.Payload, payloads[i])
		}
	}
}

// TestZeichenketten (Z26): Escape-Folge, U+FFFD, ein Zeichen aus vier Bytes und NBSP in Payload und Schlüsseln
// bleiben byte-gleich; die Werte entstehen aus Bytewerten (AGENTS.md §3.18).
func TestZeichenketten(t *testing.T) {
	fffd, vier, nbsp := "\xef\xbf\xbd", "\xf0\x9f\x98\x80", "\xc2\xa0"
	werte := []string{esc("u0000"), fffd, vier, nbsp, esc("n"), esc("\"") + esc("\\"), esc("ud83d") + esc("ude00")}
	var es []ent
	for i, v := range werte {
		payload := `{"k":"` + v + `","` + v + `":1}`
		key := "K" + v
		if i == 0 {
			key = "K\x00"
		}
		es = append(es, ent{"Set" + fffd, key + fmt.Sprint(i), payload, []string{"k", v}, "complete"})
	}
	w := neu(t)
	w.baseline("bsl-1", "svc", es)
	if _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); err != nil {
		t.Fatal(err)
	}
	_, got := angelegterState(t, w)
	if len(got) != len(werte) {
		t.Fatalf("Einträge %d", len(got))
	}
	for i, e := range got {
		if e.Payload != es[i].payload || e.EntityKey != es[i].key || e.EntitySet != "Set"+fffd {
			t.Errorf("Eintrag %d: Payload %q Schlüssel %q Set %q", i, e.Payload, e.EntityKey, e.EntitySet)
		}
		if !slices.Equal(e.KnownProperties, es[i].props) {
			t.Errorf("Eintrag %d: Properties %q", i, e.KnownProperties)
		}
	}
	if !bytes.Contains([]byte(got[0].Payload), []byte{'\\', 'u', '0', '0', '0', '0'}) {
		t.Error("die Escape-Folge ist nicht erhalten")
	}
}

// TestLeereUndTeilweiseBaseline (Z27): eine leere Baseline und Entities mit partial werden angelegt; ReplaceAll
// läuft genau einmal, auch mit leerer Liste; Vollständigkeit wird durchgereicht.
func TestLeereUndTeilweiseBaseline(t *testing.T) {
	voll := genEnts(6, 2)
	teil := slices.Clone(voll)
	for i := range teil {
		teil[i].compl = "partial"
	}
	for name, es := range map[string][]ent{"leer": nil, "nur partial": teil, "gemischt": voll} {
		w := neu(t)
		w.baseline("bsl-1", "svc", es)
		if _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		zaehler(t, w, map[string]int{"tx Entities.ReplaceAll": 1, lblTx: 1})
		_, got := angelegterState(t, w)
		if !reflect.DeepEqual(got, erwarteterState("ws-neu", es)) {
			t.Errorf("%s: State %+v", name, got)
		}
		if len(w.st.stand().ent["ws-neu"]) != len(es) {
			t.Errorf("%s: gespeichert %d", name, len(w.st.stand().ent["ws-neu"]))
		}
	}
}

// TestErgebnisVonCreate (Z28): das Ergebnis ist Workspace{ID, Service der Baseline, Baseline, Revision 0,
// CreatedAt}, gleich dem Wert, den Workspaces.Create erhielt; die Service-Kennung ist die der Baseline.
func TestErgebnisVonCreate(t *testing.T) {
	for _, id := range []string{"", "ws-eigen"} {
		w := neu(t)
		w.baseline("bsl-2", "svc-zwei", standardEnts())
		ws, err := w.v.Create(context.Background(), "bsl-2", id)
		if err != nil {
			t.Fatal(err)
		}
		erhalten := w.p.von("tx Workspaces.Create")[0].args[0].(model.Workspace)
		want := model.Workspace{ID: ws.ID, ServiceID: "svc-zwei", BaselineSnapshotID: "bsl-2", Revision: 0, CreatedAt: t0()}
		if id != "" && ws.ID != id {
			t.Errorf("ID %q", ws.ID)
		}
		if ws != want || erhalten != want {
			t.Errorf("Ergebnis %+v, Create erhielt %+v, erwartet %+v", ws, erhalten, want)
		}
		if got := w.st.stand().ws[ws.ID]; got != want {
			t.Errorf("gespeichert %+v", got)
		}
	}
}

// TestTransaktionDerAnlage (Z29): genau eine Transaktion; darin Create und danach ReplaceAll je genau
// einmal, sonst nichts; scheitert Create, läuft ReplaceAll nicht.
func TestTransaktionDerAnlage(t *testing.T) {
	w := neu(t)
	if _, err := w.v.Create(context.Background(), "bsl-1", ""); err != nil {
		t.Fatal(err)
	}
	var folge []string
	for _, op := range w.p.ops() {
		if !strings.HasPrefix(op, "Strom.") && op != "Uhr.Now" && !strings.HasPrefix(op, "IDs.") {
			folge = append(folge, op)
		}
	}
	want := []string{"Snapshots.Get", "Blobs.Open", lblTx, "tx Workspaces.Create", "tx Entities.ReplaceAll"}
	if !slices.Equal(folge, want) {
		t.Errorf("Folge %v, erwartet %v", folge, want)
	}
	// Create scheitert: kein ReplaceAll, aber die Transaktion bleibt eine.
	w2 := neu(t)
	w2.p.setzeFehler("tx Workspaces.Create", outbound.ErrBereitsVorhanden)
	_, _ = w2.v.Create(context.Background(), "bsl-1", "x")
	zaehler(t, w2, map[string]int{lblTx: 1, "tx Entities.ReplaceAll": 0, "tx Workspaces.Create": 1})
}

// TestFehlerInDerAnlage (Z30): ein Fehler an Create oder ReplaceAll und ein Abbruch des Kontexts nach
// jedem der beiden Schritte hinterlassen keinen Workspace und keinen State; der Fehler erreicht den Aufrufer
// abgebildet.
func TestFehlerInDerAnlage(t *testing.T) {
	for _, label := range []string{"tx Workspaces.Create", "tx Entities.ReplaceAll"} {
		w := neu(t)
		vorher := w.st.stand()
		w.p.setzeFehler(label, errPlatte)
		_, err := w.v.Create(context.Background(), "bsl-1", "ws-neu")
		pruefeCode(t, err, model.CodeE9001, errPlatte)
		if !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("Fehler an %s: Zustand geändert", label)
		}
		// Abbruch nach dem Schritt.
		w2 := neu(t)
		vorher = w2.st.stand()
		ctx, cancel := context.WithCancel(context.Background())
		w2.p.nach = func(op string) {
			if op == label {
				cancel()
			}
		}
		_, err = w2.v.Create(ctx, "bsl-1", "ws-neu")
		pruefeOhneCode(t, err, context.Canceled)
		if !zustaendeGleich(vorher, w2.st.stand()) {
			t.Errorf("Abbruch nach %s: Zustand geändert", label)
		}
	}
}

// TestVergebeneIDNachDerBaseline (Z31): eine vergebene id ist ODR-E7007 aus Create, der vorhandene Workspace
// bleibt, ReplaceAll läuft nicht, und die Baseline wurde zuvor gelesen (kein Vorab-Lesen).
func TestVergebeneIDNachDerBaseline(t *testing.T) {
	w := neu(t)
	w.baseline("bsl-2", "svc", genEnts(2, 1))
	w.vorbelegen("ws-1", 4)
	z := w.st.stand()
	ws := z.ws["ws-1"]
	ws.BaselineSnapshotID = "bsl-2"
	z.ws["ws-1"] = ws
	w.st.setze(z)
	vorher := w.st.stand()
	w.p.leeren()
	_, err := w.v.Create(context.Background(), "bsl-1", "ws-1")
	pruefeCode(t, err, model.CodeE7007, outbound.ErrBereitsVorhanden)
	zaehler(t, w, map[string]int{"Blobs.Open": 1, "Workspaces.Get": 0, "tx Workspaces.Get": 0, "tx Entities.ReplaceAll": 0, lblTx: 1})
	if !zustaendeGleich(vorher, w.st.stand()) {
		t.Error("der vorhandene Workspace wurde geändert")
	}
}

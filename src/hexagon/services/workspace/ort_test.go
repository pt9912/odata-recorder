package workspace_test

import (
	"bytes"
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Z71 (SPEC-129 Nr. 6): Der Fehlertext jedes Verstoßes des Manifests nennt den Ort (Glied und Stelle) und
// keinen Wert. Die Menge sind alle Stellen, an denen manifest.go einen Verstoß bildet (der Aufruf von verstoss),
// mal die Formen des Wertes an der Stelle, mal die Einträge 0 bis 2, mal Create und Reset. Jeder Wert des
// Manifests (Kennungen, Schlüssel, Eigenschaften, der verletzende Wert selbst) trägt die Marke geheimMarke
// oder ist die Zahl geheimZahl; kommt eine von beiden im Text irgendeiner Stufe der Ursachenkette vor, ist
// das ein Leck.

const (
	geheimMarke = "GEHEIM"
	geheimZahl  = "424242"
)

// ortFall ist ein Manifest, das an der Stelle stelle (Schlüssel aus ortStellen) verstößt; der Text nennt ort
// und das Merkmal der Stelle.
type ortFall struct {
	stelle, name, ort string
	manifest          []byte
}

// ortStellen sind die Schlüssel der Stellen: jeder ist ein Teilstring genau eines Formattextes eines Aufrufs
// von verstoss in manifest.go, und jeder dieser Formattexte enthält genau einen Schlüssel.
func ortStellen() []string {
	return []string{
		"kein JSON-Wert", "%s ist kein Objekt", "fehlt oder ist keine Zeichenkette", "] doppelt", "manifest_version",
		"passt nicht zur Baseline", "entities fehlt oder ist kein Array", "entity_count", "ist leer",
		"completeness ist unbekannt", "payload fehlt oder ist kein Objekt", "selected_properties fehlt oder ist kein Array",
		"selected_properties enth",
	}
}

func markerEnts() []ent {
	return []ent{
		{"GEHEIM-set-0", "GEHEIM-key-0", `{"GEHEIM-p":"GEHEIM-v0"}`, []string{"GEHEIM-a0", "GEHEIM-b0"}, "complete"},
		{"GEHEIM-set-1", "GEHEIM-key-1", `{"GEHEIM-p":"GEHEIM-v1"}`, []string{"GEHEIM-a1"}, "partial"},
		{"GEHEIM-set-2", "GEHEIM-key-2", `{"GEHEIM-p":"GEHEIM-v2"}`, []string{"GEHEIM-a2", "GEHEIM-b2"}, "complete"},
	}
}

func markerBaum() obj { return manifestBaum("GEHEIM-bsl", "GEHEIM-svc", markerEnts()) }

func ortFaelle() []ortFall {
	var r []ortFall
	add := func(stelle, name, ort string, b []byte) { r = append(r, ortFall{stelle, name, ort, b}) }
	// kopf setzt das Glied des Manifests auf jede Form vs.
	kopf := func(stelle, glied, ort string, vs ...any) {
		for j, v := range vs {
			m := markerBaum()
			setze(m, glied, v)
			add(stelle, glied+" Form "+strconv.Itoa(j), ort, []byte(schreibe(m)))
		}
	}
	// eintrag setzt das Glied des Eintrags i (0 bis 2) auf jede Form vs; ortGlied hängt sich an entities[i].
	eintrag := func(stelle, glied, ortGlied string, vs ...any) {
		for i := 0; i < 3; i++ {
			for j, v := range vs {
				m := markerBaum()
				setze(eintraege(m)[i].(obj), glied, v)
				add(stelle, "Eintrag "+strconv.Itoa(i)+" "+glied+" Form "+strconv.Itoa(j), "entities["+strconv.Itoa(i)+"]"+ortGlied, []byte(schreibe(m)))
			}
		}
	}
	zahl := roh(geheimZahl)
	feld := []any{geheimMarke}
	objGlied := obj{geheimMarke: geheimMarke}
	nichtString := []any{fehlt, zahl, feld, objGlied, nil, true}

	// kein JSON-Wert: Präfixe, Rest dahinter, ungültiges UTF-8 in einem Wert, BOM.
	g := []byte(schreibe(markerBaum()))
	for _, n := range []int{0, 1, len(g) / 2, len(g) - 1} {
		add("kein JSON-Wert", "Präfix "+strconv.Itoa(n), "kein JSON-Wert", g[:n])
	}
	add("kein JSON-Wert", "Rest", "kein JSON-Wert", append(append([]byte{}, g...), " GEHEIM"...))
	add("kein JSON-Wert", "UTF-8", "kein JSON-Wert", bytes.Replace(g, []byte("GEHEIM-set-0"), []byte("GEHEIM-\xff"), 1))
	add("kein JSON-Wert", "BOM", "kein JSON-Wert", append([]byte{0xEF, 0xBB, 0xBF}, g...))

	// kein Objekt: das Manifest und jedes Element von entities.
	for j, b := range []string{`"GEHEIM-s"`, `["GEHEIM"]`, geheimZahl, "null", "true"} {
		add("%s ist kein Objekt", "Manifest Form "+strconv.Itoa(j), "Manifest ist kein Objekt", []byte(b))
	}
	for i := 0; i < 3; i++ {
		for j, v := range []any{"GEHEIM-s", feld, zahl, nil, true} {
			m := markerBaum()
			eintraege(m)[i] = v
			add("%s ist kein Objekt", "Element "+strconv.Itoa(i)+" Form "+strconv.Itoa(j), "entities["+strconv.Itoa(i)+"] ist kein Objekt", []byte(schreibe(m)))
		}
	}

	// fehlt oder ist keine Zeichenkette: die drei Glieder des Kopfes und die drei des Eintrags.
	for _, gl := range []string{"snapshot_kind", "service_id", "baseline_snapshot_id"} {
		kopf("fehlt oder ist keine Zeichenkette", gl, "Manifest."+gl, nichtString...)
	}
	for _, gl := range []string{"entity_set", "typed_key", "completeness"} {
		eintrag("fehlt oder ist keine Zeichenkette", gl, "."+gl, nichtString...)
	}

	// doppelt: jedes Paar der drei Einträge; der spätere ist der Ort.
	for _, p := range [][2]int{{0, 1}, {0, 2}, {1, 2}} {
		m := markerBaum()
		a, b := eintraege(m)[p[0]].(obj), eintraege(m)[p[1]].(obj)
		b["entity_set"], b["typed_key"] = a["entity_set"], a["typed_key"]
		add("] doppelt", "Paar "+strconv.Itoa(p[0])+","+strconv.Itoa(p[1]), "entities["+strconv.Itoa(p[1])+"] doppelt", []byte(schreibe(m)))
	}

	kopf("manifest_version", "manifest_version", "manifest_version", fehlt, zahl, "GEHEIM-v", feld, nil)
	kopf("passt nicht zur Baseline", "snapshot_kind", "snapshot_kind passt", "GEHEIM-kind")
	kopf("passt nicht zur Baseline", "service_id", "service_id passt", "GEHEIM-other")
	kopf("passt nicht zur Baseline", "baseline_snapshot_id", "baseline_snapshot_id passt", "GEHEIM-other")
	kopf("entities fehlt oder ist kein Array", "entities", "entities fehlt", fehlt, nil, "GEHEIM-e", zahl, objGlied)
	kopf("entity_count", "entity_count", "entity_count", fehlt, zahl, "GEHEIM-c", nil, roh("3.0"), feld)

	// leer: entity_set leer bei geheimem Schlüssel und umgekehrt.
	eintrag("ist leer", "entity_set", ": entity_set", "")
	eintrag("ist leer", "typed_key", ": entity_set", "")
	eintrag("completeness ist unbekannt", "completeness", ".completeness", "GEHEIM-c")
	eintrag("payload fehlt oder ist kein Objekt", "payload", ".payload", fehlt, nil, "GEHEIM-p", zahl, feld)
	eintrag("selected_properties fehlt oder ist kein Array", "selected_properties", ".selected_properties", fehlt, nil, "GEHEIM-s", zahl, objGlied)
	eintrag("selected_properties enth", "selected_properties", ".selected_properties",
		[]any{"GEHEIM-a", "GEHEIM-a"}, []any{"GEHEIM-a", ""}, []any{zahl}, []any{nil, "GEHEIM-a"}, []any{objGlied})
	return r
}

// verstossFormate liest die Formattexte aller Aufrufe von verstoss in manifest.go (das Arbeitsverzeichnis
// eines Tests ist das Paketverzeichnis).
func verstossFormate(t *testing.T) []string {
	t.Helper()
	src, err := os.ReadFile("manifest.go")
	if err != nil {
		t.Fatal(err)
	}
	f, err := parser.ParseFile(token.NewFileSet(), "manifest.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var r []string
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "verstoss" && len(c.Args) > 0 {
			l, ok := c.Args[0].(*ast.BasicLit)
			if !ok || l.Kind != token.STRING {
				t.Fatalf("Aufruf von verstoss ohne Formattext als Literal bei %v", c.Pos())
			}
			s, err := strconv.Unquote(l.Value)
			if err != nil {
				t.Fatal(err)
			}
			r = append(r, s)
		}
		return true
	})
	return r
}

// TestVerstoesseNennenOrtOhneWert (Z71): je Fall der Menge nennt jede Stufe des Fehlertextes von Create und
// Reset nirgends eine Marke, und der Text insgesamt nennt Ort und Merkmal der Stelle.
func TestVerstoesseNennenOrtOhneWert(t *testing.T) {
	formate := verstossFormate(t)
	if len(formate) != 13 || len(ortStellen()) != 13 {
		t.Fatalf("%d Aufrufe von verstoss, %d Stellen in der Tabelle, erwartet je 13", len(formate), len(ortStellen()))
	}
	// Die Tabelle deckt die Stellen des Codes genau: jeder Formattext trifft genau einen Schlüssel, jeder
	// Schlüssel genau einen Formattext, und jede Stelle hat Fälle.
	for _, f := range formate {
		n := 0
		for _, s := range ortStellen() {
			if strings.Contains(f, s) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("Formattext %q trifft %d Schlüssel", f, n)
		}
	}
	faelle := ortFaelle()
	je := map[string]int{}
	for _, f := range faelle {
		je[f.stelle]++
	}
	for _, s := range ortStellen() {
		n := 0
		for _, f := range formate {
			if strings.Contains(f, s) {
				n++
			}
		}
		if n != 1 || je[s] == 0 {
			t.Errorf("Schlüssel %q trifft %d Formattexte und hat %d Fälle", s, n, je[s])
		}
	}
	for _, f := range faelle {
		merkmal := strings.TrimPrefix(f.stelle, "%s")
		for _, op := range []string{"Create", "Reset"} {
			w := neu(t)
			w.baselineBytes("GEHEIM-bsl", "GEHEIM-svc", f.manifest)
			var err error
			if op == "Create" {
				_, err = w.v.Create(context.Background(), "GEHEIM-bsl", "ws-neu")
			} else {
				w.vorbelegen("ws-1", 0)
				z := w.st.stand()
				x := z.ws["ws-1"]
				x.BaselineSnapshotID = "GEHEIM-bsl"
				z.ws["ws-1"] = x
				w.st.setze(z)
				_, err = w.v.Reset(context.Background(), "ws-1")
			}
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
			if err == nil {
				continue
			}
			// Der Ort steht in der Ursache (ErrBeschaedigt), nicht im Text des Codes: eine Stufe der Kette
			// nennt Ort und Merkmal zugleich.
			nennt := false
			for u := err; u != nil; u = errors.Unwrap(u) {
				s := u.Error()
				if strings.Contains(s, geheimMarke) || strings.Contains(s, geheimZahl) {
					t.Errorf("%s/%s/%s: Wert im Fehlertext: %q", f.stelle, f.name, op, s)
				}
				nennt = nennt || (strings.Contains(s, f.ort) && strings.Contains(s, merkmal))
			}
			if !nennt {
				t.Errorf("%s/%s/%s: keine Stufe von %q nennt Ort %q und Merkmal %q", f.stelle, f.name, op, err.Error(), f.ort, merkmal)
			}
		}
	}
	t.Logf("Fälle: %d an %d Stellen, je Create und Reset", len(faelle), len(ortStellen()))
}

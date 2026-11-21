package mock_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Die Schlüssel der Pflicht-Attribute von odata.mock.query stehen als Literal im Test (SPEC-029); pflichtQuery liest sie aus
// support.Spans(), und TestPflichtSchluesselQuery hält beide gleich.
const (
	attrSet = "recorder.entity_set"
	attrCov = "recorder.coverage_status"
)

// pflichtQuery sind die Pflicht-Attribute von SpanMockQuery aus der Tabelle der Spans, nicht aus dem Code des Dienstes.
func pflichtQuery(t *testing.T) []string {
	t.Helper()
	for _, s := range support.Spans() {
		if s.Name == support.SpanMockQuery {
			return s.Pflicht
		}
	}
	t.Fatal("support.Spans() nennt odata.mock.query nicht")
	return nil
}

// TestPflichtSchluesselQuery hält, dass die Literale des Tests die Pflicht-Attribute der Tabelle sind (SPEC-029).
func TestPflichtSchluesselQuery(t *testing.T) {
	if got := pflichtQuery(t); !slices.Equal(got, []string{attrSet, attrCov}) {
		t.Errorf("Pflicht-Attribute %v, erwartet [%s %s]", got, attrSet, attrCov)
	}
}

// erwarteSpan prüft nach einem Aufruf den Stand des Fakes (Q3, Q12; SPEC-290 Nr. 3, 4, 6): ein Span `odata.mock.query` mit einem
// StartSpan, einem End und `sets` Aufrufen von Set mit je einem Attribut, kein Log, kein CountError, die Endwerte von Set und Status,
// und die Schlüssel allein die der Pflicht-Attribute; kein Attribut trägt den Kanarienwert.
func erwarteSpan(t *testing.T, kontext string, u *umg, set, status string, sets int) {
	t.Helper()
	f := u.tel.FakeTelemetry
	zaehler := map[supporttest.Operation]int{
		supporttest.OpStartSpan: 1, supporttest.OpEnd: 1, supporttest.OpSet: sets, supporttest.OpLog: 0, supporttest.OpCountError: 0,
		supporttest.OpSnapshot: 0,
	}
	for op, soll := range zaehler {
		if got := f.Anzahl(op); got != soll {
			t.Errorf("%s: Operation %d %d-mal, erwartet %d", kontext, op, got, soll)
		}
	}
	buch := buecher(f)
	if len(buch) != 1 || buch[0].name != support.SpanMockQuery || buch[0].enden != 1 || buch[0].sets != sets {
		t.Errorf("%s: Bücher %+v", kontext, buch)
		return
	}
	pflicht := pflichtQuery(t)
	for _, a := range f.Aufrufe() {
		if a.Op == supporttest.OpSet && len(a.Attrs) != 1 {
			t.Errorf("%s: Set mit %d Attributen", kontext, len(a.Attrs))
		}
		if a.Op == supporttest.OpStartSpan && (len(a.Attrs) != 2 || a.Attrs[0].Key != pflicht[0] || a.Attrs[1].Key != pflicht[1]) {
			t.Errorf("%s: StartSpan mit %v, erwartet allein %v", kontext, a.Attrs, pflicht)
		}
	}
	for _, a := range alleAttribute(f) {
		if !slices.Contains(pflicht, a.Key) || a.Art != support.ArtString || strings.Contains(a.Text, kanarienwert) {
			t.Errorf("%s: Attribut %+v außerhalb der Pflicht oder mit Kanarienwert", kontext, a)
		}
	}
	if got, _ := buch[0].wert(attrSet); got != set {
		t.Errorf("%s: %s %q, erwartet %q", kontext, attrSet, got, set)
	}
	if got, _ := buch[0].wert(attrCov); got != status {
		t.Errorf("%s: %s %q, erwartet %q", kontext, attrCov, got, status)
	}
}

// queryFall ist ein Lauf einer Abfrage mit dem Ausgang des Dienstes und den Endwerten des Spans.
type queryFall struct {
	name   string
	bau    func(u *umg)
	code   model.Code
	ctxErr error
	set    string
	status string
	sets   int
}

// laufe ruft GET und prüft Ausgang und Span; der Ausgang ist der Code, der Kontextfehler ohne Code oder die Antwort 200.
func (f queryFall) laufe(t *testing.T) {
	t.Helper()
	u := leseUmg(t)
	f.bau(u)
	a, err := u.leseGET()
	switch {
	case f.ctxErr != nil:
		if !errors.Is(err, f.ctxErr) || fehlercode(err) != 0 {
			t.Errorf("%s: Fehler %v, erwartet %v ohne Code", f.name, err, f.ctxErr)
		}
	case f.code != 0:
		_ = erwarteFehler(t, f.name, err, f.code)
	case err != nil:
		t.Errorf("%s: Fehler %v", f.name, err)
	default:
		lies(t, a)
	}
	erwarteSpan(t, f.name, u, f.set, f.status, f.sets)
}

// TestSpanJeOperation hält Q3 und Q12 (SPEC-290 Nr. 2 a, 6 T1 bis T4): jede der sechs Operationen (GET und HEAD an Collection,
// Entity, Anzahl) öffnet odata.mock.query einmal und beendet ihn einmal; der Span trägt allein die zwei Pflicht-Schlüssel, und der
// Fake zählt jede Operation: Set je einmal für das Set und, außer beim Lookup mit Treffer, für die Coverage. Methode 2 × Art 3 = 6.
func TestSpanJeOperation(t *testing.T) {
	arten := []struct {
		name   string
		art    model.RessourcenArt
		status string
		sets   int
	}{
		{"Collection", model.RessourceMenge, "complete", 2},
		{"Entity", model.RessourceEntity, "nicht_gelesen", 1},
		{"Anzahl", model.RessourceAnzahl, "complete", 2},
	}
	n := 0
	for _, methode := range []string{"GET", "HEAD"} {
		for _, a := range arten {
			n++
			u := leseUmg(t)
			u.ressource(a.art, setProducts, model.Optionen{}, zahl("", "1"))
			antwort, err := u.rufe(anfrage(methode, dienstRoot+"/x"))
			if err != nil {
				t.Fatalf("%s %s: %v", methode, a.name, err)
			}
			lies(t, antwort)
			erwarteSpan(t, methode+" "+a.name, u, setProducts, a.status, a.sets)
		}
	}
	if n != 6 {
		t.Errorf("%d Operationen, erwartet 6", n)
	}
}

// TestSpanFolge hält Q5 (SPEC-290 Nr. 3 a und b, 6): die Folge von Ports und Telemetrie je Art als Literal; der Span beginnt nach
// Workspaces.Get und vor Metadaten.Get, Set(entity_set) folgt der Prüfung des Sets, Set(coverage_status) dem Lesen der Coverage, End
// dem letzten Aufruf des Schreibers. Vier Folgen.
func TestSpanFolge(t *testing.T) {
	kopf := []string{"URL.ParseURL", "Workspaces.Get", "Telemetrie.StartSpan", "Metadaten.Get", "Blobs.Open", "MetadatenCodec.ParseMetadata",
		"Span.Set " + attrSet}
	cov := "Span.Set " + attrCov
	faelle := []struct {
		name string
		bau  func(u *umg)
		code model.Code
		rest []string
	}{
		{"Collection", func(u *umg) { u.menge(model.Optionen{}) }, 0,
			[]string{"Coverage.Get", cov, "Entities.ListSet", "Schreiber.SchreibeCollection", "Span.End"}},
		{"Entity mit Treffer", func(u *umg) { u.entity(model.Optionen{}, "1") }, 0, []string{"Entities.Get", "Schreiber.SchreibeEntity", "Span.End"}},
		{"Entity ohne Treffer", func(u *umg) { u.entity(model.Optionen{}, "99") }, model.CodeE6008,
			[]string{"Entities.Get", "Coverage.Get", cov, "Span.End"}},
		{"Anzahl", func(u *umg) { u.anzahl(model.Optionen{}) }, 0,
			[]string{"Coverage.Get", cov, "Entities.ListSet", "Schreiber.SchreibeAnzahl", "Span.End"}},
	}
	for _, f := range faelle {
		u := leseUmg(t)
		u.tel.setzeMitProt()
		f.bau(u)
		a, err := u.leseGET()
		if f.code != 0 {
			_ = erwarteFehler(t, f.name, err, f.code)
		} else if err != nil {
			t.Fatalf("%s: %v", f.name, err)
		} else {
			lies(t, a)
		}
		if got, want := u.p.ops(), append(slices.Clone(kopf), f.rest...); !slices.Equal(got, want) {
			t.Errorf("%s: Folge\n%v\nerwartet\n%v", f.name, got, want)
		}
	}
}

// unbekannteNamen sind Namen eines Sets, die das Modell nicht nennt: mit `'`, Leerzeichen, `$`, `/`, einem Byte ab 0x80 (aus
// Bytewerten, AGENTS.md §3.18) und 300 Bytes; jeder trägt den Kanarienwert, damit ein Attribut ihn verrät.
func unbekannteNamen() []string {
	k := kanarienwert
	return []string{k + "'x", k + " y", k + "$top", k + "/z", k + string([]byte{0xc3, 0xbc}), k + strings.Repeat("x", 300-len(k))}
}

// TestSpanEntitySet hält Q9 (SPEC-290 Nr. 4 a): recorder.entity_set ist der Name aus dem Modell nach der Prüfung, sonst `_.`, nie der
// Text des Requests. Art 3 × {bekanntes Set, 6 unbekannte Namen, Modell nicht lesbar (S4)} = 24.
func TestSpanEntitySet(t *testing.T) {
	arten := []struct {
		name string
		art  model.RessourcenArt
	}{{"Collection", model.RessourceMenge}, {"Entity", model.RessourceEntity}, {"Anzahl", model.RessourceAnzahl}}
	n := 0
	for _, a := range arten {
		setze := func(set string) func(u *umg) {
			return func(u *umg) { u.ressource(a.art, set, model.Optionen{}, zahl("", "1")) }
		}
		var faelle []queryFall
		faelle = append(faelle, queryFall{name: a.name + " bekannt", bau: setze(setProducts), set: setProducts, status: "complete", sets: 2})
		if a.art == model.RessourceEntity {
			faelle[0].status, faelle[0].sets = "nicht_gelesen", 1
		}
		for i, name := range unbekannteNamen() {
			faelle = append(faelle, queryFall{name: a.name + " unbekannt " + string(rune('a'+i)), bau: setze(name), code: model.CodeE6008, set: "_.", status: "nicht_gelesen"})
		}
		faelle = append(faelle, queryFall{name: a.name + " Modell nicht lesbar", code: model.CodeE3001, set: "_.", status: "nicht_gelesen",
			bau: func(u *umg) {
				setze(setProducts)(u)
				u.p.setzeFehler("Blobs.Open", outbound.ErrBlobFehlt)
			}})
		for _, f := range faelle {
			n++
			f.laufe(t)
		}
	}
	if n != 24 || len(unbekannteNamen()) != 6 {
		t.Errorf("%d Fälle und %d unbekannte Namen, erwartet 24 und 6", n, len(unbekannteNamen()))
	}
}

// TestSpanCoverage hält Q10 (SPEC-290 Nr. 4 b, 7 f): bei Collection und Anzahl ist recorder.coverage_status der gelesene Status, auch bei
// ODR-E6003 des strikten Modus. Art 2 × Status 3 × STRICT_COVERAGE 2 = 12, dazu der Rang `ODR-E6003` vor `ODR-E6005` im strikten Modus.
func TestSpanCoverage(t *testing.T) {
	status := []model.CoverageStatus{"complete", "partial", "unknown"}
	if !slices.Equal(status, model.AlleCoverageStatus()) || len(status) != 3 {
		t.Fatalf("Status %v, Modell %v", status, model.AlleCoverageStatus())
	}
	n := 0
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceAnzahl} {
		for _, s := range status {
			for _, strikt := range []bool{false, true} {
				n++
				code := model.Code(0)
				if strikt && s != "complete" {
					code = model.CodeE6003
				}
				queryFall{name: string(s), code: code, set: setProducts, status: string(s), sets: 2, bau: func(u *umg) {
					u.co.setze(baselineID, setProducts, s)
					u.k.StrictCoverage = strikt
					u.ressource(art, setProducts, model.Optionen{})
				}}.laufe(t)
			}
		}
	}
	if n != 12 {
		t.Errorf("%d Fälle, erwartet 12", n)
	}
	unbekannt := filterOpt(vergleich(model.VergleichEq, "Unbekannt", model.Literal{Art: model.LiteralZahl, Text: "1"}))
	queryFall{name: "strikt partial mit unbekannter Property", code: model.CodeE6003, set: setProducts, status: "partial", sets: 2, bau: func(u *umg) {
		u.co.setze(baselineID, setProducts, model.CoveragePartial)
		u.k.StrictCoverage = true
		u.menge(model.Optionen{Filter: unbekannt})
	}}.laufe(t)
}

// TestSpanCoverageLookup hält Q11 (SPEC-290 Nr. 4 b): beim Lookup ohne Treffer ist der Wert der gelesene Status, wo die Coverage nicht
// gelesen wird `nicht_gelesen`. Ohne Treffer: Status 3 × STRICT_COVERAGE 2 = 6; nicht gelesen: 7 Fälle, dazu `$select` einer Property,
// die der Eintrag nicht kennt.
func TestSpanCoverageLookup(t *testing.T) {
	n := 0
	for _, s := range model.AlleCoverageStatus() {
		for _, strikt := range []bool{false, true} {
			n++
			code := model.CodeE6003
			if s == "complete" {
				code = model.CodeE6008
			}
			queryFall{name: "ohne Treffer " + string(s), code: code, set: setProducts, status: string(s), sets: 2, bau: func(u *umg) {
				u.co.setze(baselineID, setProducts, s)
				u.k.StrictCoverage = strikt
				u.entity(model.Optionen{}, "99")
			}}.laufe(t)
		}
	}
	tot := produkt(1, "t")
	tot.Deleted = true
	ohneHit := func(u *umg) { u.entity(model.Optionen{}, "99") }
	faelle := []queryFall{
		{name: "Treffer", bau: func(u *umg) { u.entity(model.Optionen{}, "1") }, set: setProducts, status: "nicht_gelesen", sets: 1},
		{name: "Tombstone", bau: func(u *umg) { u.en.setze(setProducts, tot); u.entity(model.Optionen{}, "1") }, code: model.CodeE6008, set: setProducts, status: "nicht_gelesen", sets: 1},
		{name: "Coverage fehlt (unknown, gelesen)", bau: func(u *umg) { u.ressource(model.RessourceEntity, "Kunden", model.Optionen{}, zahl("", "99")) }, code: model.CodeE6003,
			set: "Kunden", status: "unknown", sets: 2},
		{name: "Lesefehler", bau: func(u *umg) { ohneHit(u); u.p.setzeFehler("Coverage.Get", errors.New("roh")) }, code: model.CodeE9001, set: setProducts, status: "nicht_gelesen", sets: 1},
		{name: "Status außerhalb", bau: func(u *umg) { ohneHit(u); u.co.setze(baselineID, setProducts, "bogus") }, code: model.CodeE9001, set: setProducts, status: "nicht_gelesen", sets: 1},
		{name: "Kontextfehler", bau: func(u *umg) { ohneHit(u); u.p.setzeFehler("Coverage.Get", context.Canceled) }, ctxErr: context.Canceled, set: setProducts, status: "nicht_gelesen", sets: 1},
		{name: "unbekanntes Set", bau: func(u *umg) { u.ressource(model.RessourceEntity, "Unbekannt", model.Optionen{}, zahl("", "1")) }, code: model.CodeE6008, set: "_.", status: "nicht_gelesen"},
		{name: "Select einer unbekannten Property bei Treffer", code: model.CodeE6003, set: setProducts, status: "nicht_gelesen", sets: 1, bau: func(u *umg) {
			u.en.setze(setProducts, produkt(5, "e", "Id"))
			u.entity(model.Optionen{Select: &model.SelectOption{Namen: []string{"Name"}}}, "5")
		}},
	}
	for _, f := range faelle {
		n++
		f.laufe(t)
	}
	if n != 6+8 {
		t.Errorf("%d Fälle, erwartet 14", n)
	}
}

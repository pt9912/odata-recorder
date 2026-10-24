package outboundtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Der Body der Einzel-Entity E1 und der Collection C1 aus drei Elementen von R (Referenzdienst), ohne Zeilenumbruch am Ende.
const (
	jEinzelR = `{"@odata.context":"http://localhost/svc/$metadata#Products/$entity","ID":1,"Name":"Milk","Price":1.5,"Weight":0.25,"Category":{"ID":"11"}}`
	jKollR   = `{"@odata.context":"$metadata#Products","@odata.count":3,"value":[{"ID":1,"Name":"Milk","Price":1.5,"Weight":0.25},` +
		`{"ID":2,"Name":"Bread","Price":2,"Category":null},{"ID":3,"Name":null,"Price":"3.10","Weight":null}],"@odata.nextLink":"Products?$skiptoken=3"}`
)

// basisR ist der gültige Body von R für die Methode.
func basisR(einzel bool) jfall {
	if einzel {
		return jfall{name: "E1", m: "R", set: "Products", einzel: true, body: []byte(jEinzelR)}
	}
	return jfall{name: "C1", m: "R", set: "Products", body: []byte(jKollR)}
}

func methodeName(einzel bool) string {
	if einzel {
		return "ParseEntity"
	}
	return "ParseCollection"
}

// Ausgänge der Argument-Fälle von A1.
const (
	a1Ungueltig = "ErrUngueltig"
	a1SetFehler = "entity_set"
	a1Kontext   = "Kontext"
)

// a1Fall ist ein Aufruf mit ungültigen Argumenten oder unbekanntem Set und dem erwarteten Ausgang.
type a1Fall struct {
	name   string
	ctx    func() context.Context
	src    bool
	grenze int64
	set    string
	leer   bool
	soll   string
}

func a1Faelle() []a1Fall {
	hin := context.Background
	kein := func() context.Context { return nil }
	return []a1Fall{
		{"ctx nil", kein, true, 1000, "Products", false, a1Ungueltig}, {"src nil", hin, false, 1000, "Products", false, a1Ungueltig},
		{"grenze 0", hin, true, 0, "Products", false, a1Ungueltig}, {"grenze -1", hin, true, -1, "Products", false, a1Ungueltig},
		{"grenze MinInt64", hin, true, math.MinInt64, "Products", false, a1Ungueltig}, {"entitySet leer", hin, true, 1000, "", false, a1Ungueltig},
		{"alle vier zugleich", kein, false, 0, "", false, a1Ungueltig},
		{"grenze 0 und beendeter ctx (Argument vor Kontext)", abgebrochen, true, 0, "Products", false, a1Ungueltig},
		{"leerer entitySet und beendeter ctx", abgebrochen, true, 1000, "", false, a1Ungueltig},
		{"src nil und unbekanntes Set (Argument vor Set)", hin, false, 1000, "Nope", false, a1Ungueltig},
		{"beendeter ctx und unbekanntes Set (Kontext vor Set)", abgebrochen, true, 1000, "Nope", false, a1Kontext},
		{"unbekanntes Set", hin, true, 1000, "Nope", false, a1SetFehler}, {"andere Schreibweise", hin, true, 1000, "products", false, a1SetFehler},
		{"mit Namensraum", hin, true, 1000, "Demo.Products", false, a1SetFehler}, {"mit Container", hin, true, 1000, "Container/Products", false, a1SetFehler},
		{"Typ statt Set", hin, true, 1000, "Product", false, a1SetFehler}, {"Leerzeichen", hin, true, 1000, "Products ", false, a1SetFehler},
		{"Steuerzeichen im Namen", hin, true, 1000, "Pro\x01ducts/~", false, a1SetFehler},
		{"Nullwert-Modell", hin, true, 1000, "Products", true, a1SetFehler},
	}
}

// pruefeA1 fährt einen Fall von A1 für die Methode: der Ausgang, kein Read und die Nullwerte.
func (s *jsuite) pruefeA1(t *testing.T, einzel bool, f a1Fall, m model.ServiceModel) {
	t.Helper()
	rd := &teilReader{teile: [][]byte{[]byte("{")}}
	var src io.Reader = rd
	if !f.src {
		src = nil
	}
	if f.leer {
		m = model.ServiceModel{}
	}
	ctx := f.ctx()
	r := s.aufruf(ctx, einzel, src, m, f.set, f.grenze)
	name := methodeName(einzel) + " " + f.name
	switch f.soll {
	case a1Ungueltig:
		var df *model.DatenFehler
		if !errors.Is(r.err, model.ErrUngueltig) || errors.As(r.err, &df) {
			t.Errorf("%s: Fehler %v, erwartet model.ErrUngueltig ohne DatenFehler", name, r.err)
		}
	case a1SetFehler:
		datenFehler(t, name, r.err, model.DatenEntitySet, f.set)
	default:
		if r.err != ctx.Err() || r.err == nil {
			t.Errorf("%s: Fehler %v, erwartet den Fehler des Kontexts", name, r.err)
		}
	}
	if rd.reads != 0 {
		t.Errorf("%s: %d Reads, erwartet 0", name, rd.reads)
	}
	r.nullwerte(t, name)
}

// a1OhneTyp ist das Modell von Hand mit einem Set, dessen Typ das Modell nicht auflöst (SPEC-126 Nr. 1 f).
func a1OhneTyp() model.ServiceModel {
	return model.NeuesServiceModel("4.0", nil, nil, []model.EntitySet{{Name: "X", EntityType: "T.Fehlt"}}, nil)
}

// jA1: die Argumente und das Set in ihrer Reihenfolge (SPEC-126 Nr. 1 a bis f).
func jA1(t *testing.T, s *jsuite) {
	hin := context.Background
	for _, einzel := range []bool{true, false} {
		for _, f := range a1Faelle() {
			s.pruefeA1(t, einzel, f, s.modelle["R"])
		}
		// Ein Set, dessen Typ das Modell nicht auflöst: dieselbe Fehlerart wie ein unbekanntes Set, ohne Read.
		s.pruefeA1(t, einzel, a1Fall{name: "Set mit nicht aufgelöstem Typ", ctx: hin, src: true, grenze: 1000, set: "X", soll: a1SetFehler}, a1OhneTyp())
		// grenze 1 ist gültig: der Body `{` ist unvollständig, kein ungültiges Argument.
		r := s.aufruf(context.Background(), einzel, &teilReader{teile: [][]byte{[]byte("{")}}, s.modelle["R"], "Products", 1)
		datenFehler(t, methodeName(einzel)+" grenze 1", r.err, model.DatenSyntax, "")
	}
}

// a2Stelle ist eine Stelle, an der der Kontext endet: im Read bei (1 ist der erste), mit dem Ende des Stroms im letzten
// Read, und die Zahl der Reads, die der Leser bis dahin gemacht hat.
type a2Stelle struct {
	name    string
	body    []byte
	stuecke []int
	bei     int
	mitEnde bool
	reads   int
}

func a2Stellen(doc, typFehler []byte) []a2Stelle {
	cuts := []int{len(doc) / 3, 2 * len(doc) / 3}
	return []a2Stelle{
		{"Abbruch im ersten Read", doc, cuts, 1, false, 1}, {"Abbruch im mittleren Read", doc, cuts, 2, false, 2},
		{"Abbruch im letzten Read, der io.EOF liefert", doc, cuts, 3, true, 3},
		{"Abbruch im letzten Read, Verstoß des Deutens dahinter", typFehler, []int{len(typFehler) / 2}, 2, true, 2},
	}
}

// pruefeA2 endet den Kontext mit dem Fehler ce an der Stelle und prüft den Fehler, die Zahl der Reads und die Abbildung.
func (s *jsuite) pruefeA2(t *testing.T, f jfall, ce error, st a2Stelle) {
	t.Helper()
	ctx := neuBeendbar(nil)
	rd := teile(st.body, st.stuecke...)
	rd.mitEnde = st.mitEnde
	rd.nach = func(n int) {
		if n == st.bei {
			ctx.beende(ce)
		}
	}
	r := s.ruf(ctx, rd, f)
	name := methodeName(f.einzel) + " " + st.name + " (" + ce.Error() + ")"
	if r.err != ce || rd.reads != st.reads {
		t.Errorf("%s: Fehler %v, %d Reads; erwartet den Fehler des Kontexts und %d Reads", name, r.err, rd.reads, st.reads)
	}
	if _, ok := outbound.Abbilden(r.err, outbound.GegenstandImport); ok {
		t.Errorf("%s: der Fehler des Kontexts bekommt einen Code", name)
	}
	r.nullwerte(t, name)
}

// jA2: der Kontext wird vor jedem Read und einmal vor dem Deuten geprüft, der Fehler kommt unverändert zurück
// (SPEC-126 Nr. 1 e): Schleife über die Kontextfehler, die Stellen und beide Methoden.
func jA2(t *testing.T, s *jsuite) {
	for _, einzel := range []bool{true, false} {
		basis := basisR(einzel)
		typFehler := jfall{m: "R", set: "Products", einzel: einzel, body: jKol("Products", `{"ID":1,"Name":5}`)}
		if einzel {
			typFehler.body = jEin("Products", `"ID":1,"Name":5`)
		}
		for _, ce := range kontextFehler() {
			for _, st := range a2Stellen(basis.body, typFehler.body) {
				s.pruefeA2(t, typFehler, ce, st)
			}
			// Abbruch und Verstoß des Lesens in den Bytes desselben Reads: der Verstoß geht vor.
			ctx := neuBeendbar(nil)
			rd := teile([]byte(`{"a" 1}`))
			rd.nach = func(int) { ctx.beende(ce) }
			r := s.ruf(ctx, rd, typFehler)
			datenFehler(t, methodeName(einzel)+" Abbruch und Verstoß im selben Read ("+ce.Error()+")", r.err, model.DatenSyntax, "")
			rd = teile(basis.body)
			r = s.ruf(neuBeendbar(ce), rd, basis)
			if r.err != ce || rd.reads != 0 {
				t.Errorf("%s beendet vor dem Aufruf (%v): %v, %d Reads", methodeName(einzel), ce, r.err, rd.reads)
			}
		}
		ctx, abbruch := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
		if r := s.ruf(ctx, teile(basis.body), basis); r.err != context.DeadlineExceeded {
			t.Errorf("%s abgelaufene Frist: %v", methodeName(einzel), r.err)
		}
		abbruch()
		if r := s.ruf(abgebrochen(), teile(basis.body), basis); r.err != context.Canceled {
			t.Errorf("%s beendet über CancelFunc: %v", methodeName(einzel), r.err)
		}
	}
}

// jA3: ein Fehler von src ist kein DatenFehler und bleibt über errors.Is erreichbar (SPEC-126 Nr. 2); Daten und
// Fehler im selben Read: ein Verstoß des Lesens in den Daten geht vor, sonst gilt der Fehler.
func jA3(t *testing.T, s *jsuite) {
	for _, einzel := range []bool{true, false} {
		basis := basisR(einzel)
		doc := basis.body
		mitte := func(b string) int { return bytesIndex(doc, b) }
		for _, f := range []struct {
			name    string
			rd      *teilReader
			mitEnde bool
			syntax  bool
		}{
			{"vor dem ersten Byte", &teilReader{}, false, false}, {"nach dem ersten Byte", teile(doc[:1]), false, false},
			{"mitten in einem String", teile(doc[:mitte("Milk")+2]), false, false},
			{"mitten in einer Zahl", teile(doc[:mitte("1.5")+1]), false, false},
			{"nach dem vollständigen Body", teile(doc), false, false},
			{"Daten und Fehler im selben Read", teile(doc[:60]), true, false},
			{"Daten mit Verstoß des Lesens und Fehler im selben Read", teile([]byte(`{"a" 1}`)), true, true},
		} {
			f.rd.ende, f.rd.mitEnde = errQuelle, f.mitEnde
			r := s.ruf(context.Background(), &endlichStrom{inner: f.rd}, basis)
			name := methodeName(einzel) + " " + f.name
			if f.syntax {
				datenFehler(t, name, r.err, model.DatenSyntax, "")
			} else {
				var df *model.DatenFehler
				if !errors.Is(r.err, errQuelle) || errors.As(r.err, &df) {
					t.Errorf("%s: Fehler %v, erwartet errQuelle ohne DatenFehler", name, r.err)
				}
			}
			r.nullwerte(t, name)
		}
		rd := &teilReader{teile: [][]byte{doc}, mitEnde: true}
		if r := s.ruf(context.Background(), rd, basis); r.err != nil {
			t.Errorf("%s Daten mit io.EOF im letzten Read: %v", methodeName(einzel), r.err)
		}
	}
	jA3M2(t, s)
}

// a3M2Zahl ist die Zahl der Läufe der Menge M2: 5 Fehlerarten × (2L + 3) Zeitpunkte für E1 (L = 138, 279) und C1 (L = 257, 517).
const a3M2Zahl = 5*279 + 5*517

// a3MaxReads ist die Zahl der Reads nach dem Fehler, nach der fehlerStrom io.EOF liefert: ein Leser, der nach dem Fehler
// weiterliest, endet damit im Fall und nicht an der Zeitgrenze des Tests.
const a3MaxReads = 10000

// endlichStrom reicht Reads an inner weiter und liefert nach a3MaxReads Reads io.EOF: ein Leser, der den wiederholten Fehler
// von inner übergeht, endet damit im Fall und nicht an der Zeitgrenze des Tests.
type endlichStrom struct {
	inner io.Reader
	reads int
}

func (w *endlichStrom) Read(p []byte) (int, error) {
	if w.reads++; w.reads > a3MaxReads {
		return 0, io.EOF
	}
	return w.inner.Read(p)
}

// fsLese ist ein geplanter Read von fehlerStrom: die Bytes und der Fehler, mit dem er endet.
type fsLese struct {
	daten []byte
	err   error
}

// fehlerStrom liefert die geplanten Reads, der letzte trägt den Fehler. Ein Read danach zählt als nachFehler und liefert den
// Rest des Bodys, ein Byte je Read, dann io.EOF; nach a3MaxReads solchen Reads liefert er io.EOF.
type fehlerStrom struct {
	plan       []fsLese
	rest       []byte
	nachFehler int
}

func (r *fehlerStrom) Read(p []byte) (int, error) {
	if len(r.plan) > 0 {
		l := r.plan[0]
		r.plan = r.plan[1:]
		return copy(p, l.daten), l.err
	}
	if r.nachFehler++; r.nachFehler > a3MaxReads || len(r.rest) == 0 {
		return 0, io.EOF
	}
	n := copy(p[:1], r.rest)
	r.rest = r.rest[n:]
	return n, nil
}

// a3Fehlerarten sind die Fehler von src der Menge M2: jeder Fehler außer dem Wert io.EOF zählt (SPEC-126 Nr. 2), auch der
// umhüllte io.EOF, ein Kontextfehler von src bei lebendem ctx und io.ErrUnexpectedEOF.
func a3Fehlerarten() []error {
	return []error{errQuelle, io.ErrUnexpectedEOF, context.Canceled, context.DeadlineExceeded, fmt.Errorf("x: %w", io.EOF)}
}

// a3Zeitpunkte sind die Zeitpunkte des Fehlers in der Zerlegung eines Bodys der Länge L (2L + 3): ein Byte je Read und der
// Fehler allein nach k Bytes (k = 0 bis L), ein Byte je Read und der Fehler im Read des Bytes k + 1 (k = 0 bis L − 1), der
// ganze Body in einem Read mit dem Fehler, der ganze Body in einem Read und der Fehler im nächsten.
func a3Zeitpunkte(doc []byte, fehler error) []*fehlerStrom {
	bytewei := func(k int) []fsLese {
		var plan []fsLese
		for i := range k {
			plan = append(plan, fsLese{daten: doc[i : i+1]})
		}
		return plan
	}
	var out []*fehlerStrom
	for k := 0; k <= len(doc); k++ {
		out = append(out, &fehlerStrom{plan: append(bytewei(k), fsLese{err: fehler}), rest: doc[k:]})
	}
	for k := 0; k < len(doc); k++ {
		out = append(out, &fehlerStrom{plan: append(bytewei(k), fsLese{daten: doc[k : k+1], err: fehler}), rest: doc[k+1:]})
	}
	out = append(out, &fehlerStrom{plan: []fsLese{{daten: doc, err: fehler}}})
	return append(out, &fehlerStrom{plan: []fsLese{{daten: doc}, {err: fehler}}})
}

// jA3M2: jeder Fehler von src zu jedem Zeitpunkt endet das Lesen mit ihm, ohne DatenFehler, mit Nullwerten, und danach liest
// der Leser nicht weiter (SPEC-126 Nr. 2): Schleife über die Fehlerarten, die Zeitpunkte und beide Methoden.
func jA3M2(t *testing.T, s *jsuite) {
	var laeufe int
	for _, einzel := range []bool{true, false} {
		basis := basisR(einzel)
		for _, fehler := range a3Fehlerarten() {
			for i, rd := range a3Zeitpunkte(basis.body, fehler) {
				r := s.ruf(context.Background(), rd, basis)
				name := fmt.Sprintf("%s M2 Fehler %v Zeitpunkt %d", methodeName(einzel), fehler, i)
				var df *model.DatenFehler
				if !errors.Is(r.err, fehler) || errors.As(r.err, &df) {
					t.Errorf("%s: Fehler %v, erwartet den Fehler von src ohne DatenFehler", name, r.err)
				}
				if rd.nachFehler != 0 {
					t.Errorf("%s: %d Reads nach dem Fehler, erwartet 0", name, rd.nachFehler)
				}
				r.nullwerte(t, name)
				laeufe++
			}
		}
	}
	if laeufe != a3M2Zahl {
		t.Errorf("A3 M2: %d Läufe, erwartet %d (E1 hat %d Bytes, C1 %d)", laeufe, a3M2Zahl, len(jEinzelR), len(jKollR))
	}
	t.Logf("A3 M2: %d Läufe", laeufe)
}

func bytesIndex(doc []byte, teil string) int {
	for i := 0; i+len(teil) <= len(doc); i++ {
		if string(doc[i:i+len(teil)]) == teil {
			return i
		}
	}
	return -1
}

// Grenzen der Menge von A4: bis j4Voll Bytes ist jede Zweiteilung ein Lauf, bis j4Probe Bytes sind es die ersten und
// die letzten j4Rand Positionen und j4Streu verteilte, darüber nur ganz und Byte für Byte (akzeptiertes Negativ im Plan).
const (
	j4Voll   = 4096
	j4Probe  = 65536
	j4Rand   = 64
	j4Streu  = 16
	j4Detekt = 8
)

// j4Zweiteilungen sind die Stellen, an denen A4 einen Body der Länge l in zwei Reads teilt; unter dem Detektor immer
// die Probe (eine Zweiteilung hat keinen Zustand, den der Detektor sehen könnte; die Allaussage fährt `make test`).
func j4Zweiteilungen(l int) []int {
	rand := j4Rand
	if unterDetektor {
		rand = j4Detekt
	}
	var out []int
	switch {
	case l <= j4Voll && !unterDetektor:
		for s := 1; s < l; s++ {
			out = append(out, s)
		}
	case l <= j4Probe:
		stellen := map[int]bool{}
		for i := 1; i <= rand; i++ {
			stellen[i], stellen[l-i] = true, true
		}
		for i := 1; i <= j4Streu; i++ {
			stellen[l*i/(j4Streu+1)] = true
		}
		for s := 1; s < l; s++ {
			if stellen[s] {
				out = append(out, s)
			}
		}
	}
	return out
}

// jA4: das Ergebnis hängt nicht von der Zerlegung in Reads ab (SPEC-126 Nr. 2): Schleife über alle Fälle der Suite
// (gültige und fehlerhafte) und die Zerlegungen ganz, Byte für Byte, jede Zweiteilung bis j4Voll Bytes und darüber eine
// Probe, je mit und ohne leere Reads. Der Test zählt die Läufe der Bodys bis j4Voll Bytes gegen Bodys × (L−1) × 2.
func jA4(t *testing.T, s *jsuite) {
	var bodys, laeufe, erwartet atomic.Int64
	arbeit := make(chan jfall)
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), 8) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range arbeit {
				if len(f.body) <= j4Voll {
					bodys.Add(1)
					erwartet.Add(int64(max(len(f.body)-1, 0) * 2))
				}
				laeufe.Add(s.zerlegt(t, f))
			}
		}()
	}
	var gross []jfall
	for _, f := range s.alleFaelle(t) {
		if len(f.body) > j4Probe {
			gross = append(gross, f)
			continue
		}
		arbeit <- f
	}
	close(arbeit)
	wg.Wait()
	for _, f := range gross {
		s.zerlegt(t, f)
	}
	if (laeufe.Load() != erwartet.Load() && !unterDetektor) || bodys.Load() < 1000 {
		t.Errorf("%d Bodys bis %d Bytes, %d Läufe der Zweiteilungen, erwartet %d und mindestens 1000 Bodys",
			bodys.Load(), j4Voll, laeufe.Load(), erwartet.Load())
	}
	t.Logf("%d Bodys bis %d Bytes, %d Zweiteilungen", bodys.Load(), j4Voll, laeufe.Load())
}

// zerlegt fährt einen Fall in allen Zerlegungen von A4 gegen den Lauf ganz und liefert die Zahl der Läufe der
// Zweiteilungen (nur für einen Body bis j4Voll Bytes, sonst 0). Der erste Unterschied endet den Fall.
func (s *jsuite) zerlegt(t *testing.T, f jfall) (zweiteilungen int64) {
	lauf := func(schnitt int, leer bool) jergebnis {
		return s.ruf(context.Background(), &zerlegtReader{doc: f.body, schnitt: schnitt, leerVor: leer}, f)
	}
	soll := lauf(0, false)
	stimmt := func(schnitt int, leer bool) bool {
		if got := lauf(schnitt, leer); !reflect.DeepEqual(got, soll) {
			t.Errorf("%s (%d Bytes) zerlegt bei %d (leere Reads %v): %+v, ganz: %+v", f.name, len(f.body), schnitt, leer, got, soll)
			return false
		}
		return true
	}
	if !stimmt(-1, false) || (len(f.body) <= j4Probe && !stimmt(-1, true)) {
		return 0
	}
	for _, sch := range j4Zweiteilungen(len(f.body)) {
		for _, leer := range []bool{false, true} {
			if !stimmt(sch, leer) {
				return 0
			}
			if len(f.body) <= j4Voll {
				zweiteilungen++
			}
		}
	}
	return zweiteilungen
}

// jA5: der Leser schließt src nie, liest nach dem Ende nicht weiter und höchstens grenze+1 Bytes (SPEC-126 Nr. 2, 5).
func jA5(t *testing.T, s *jsuite) {
	var faelle int
	defer func() { t.Logf("A5: %d Fälle der Suite und 2 Fälle mit Deckel", faelle) }()
	for _, f := range s.alleFaelle(t) {
		if f.gross {
			continue
		}
		faelle++
		rd := teile(f.body)
		w := &a5Strom{inner: rd}
		s.ruf(context.Background(), w, f)
		if rd.closes != 0 || rd.nachEnde != 0 || w.leer != 0 || int64(w.geliefert)-1 > f.grenzeOderStandard() {
			t.Errorf("%s: Close %d, Reads nach dem Ende %d, Reads mit leerem Puffer %d, geliefert %d Bytes bei grenze %d", f.name,
				rd.closes, rd.nachEnde, w.leer, w.geliefert, f.grenzeOderStandard())
		}
	}
	for _, einzel := range []bool{true, false} {
		e := &deckelStrom{deckel: 1000 + a5Deckel}
		w := &a5Strom{inner: e}
		r := s.aufruf(context.Background(), einzel, w, s.modelle["R"], "Products", 1000)
		datenFehler(t, methodeName(einzel)+" endloser Strom", r.err, model.DatenZuGross, "")
		if e.geliefert != 1001 || e.closes != 0 || w.leer != 0 {
			t.Errorf("%s: geliefert %d Bytes, Close %d, Reads mit leerem Puffer %d; erwartet 1001, 0 und 0", methodeName(einzel), e.geliefert, e.closes, w.leer)
		}
	}
}

// a5Deckel ist die Zahl der Bytes über grenze, nach denen der Strom von A5 io.EOF liefert.
const a5Deckel = 4096

// a5MaxReads ist die Zahl der Reads, nach der a5Strom einen Fehler liefert: ein Leser, der Reads mit leerem Puffer wiederholt,
// endet damit im Fall und nicht an der Zeitgrenze des Tests.
const a5MaxReads = 10000

// errA5Ohnefortschritt beendet einen Leser, der a5MaxReads Reads macht.
var errA5Ohnefortschritt = errors.New("A5: der Leser liest ohne Fortschritt")

// a5Strom reicht Reads an inner weiter und zählt, was der Leser verlangt: einen Read mit leerem Puffer (er kennt dann keine
// Größengrenze mehr; der Strom liefert dafür nichts) und die gelieferten Bytes.
type a5Strom struct {
	inner interface {
		io.Reader
		io.Closer
	}
	reads, leer, geliefert int
}

func (w *a5Strom) Read(p []byte) (int, error) {
	if w.reads++; w.reads > a5MaxReads {
		return 0, errA5Ohnefortschritt
	}
	if len(p) == 0 {
		w.leer++
		return 0, nil
	}
	n, err := w.inner.Read(p)
	w.geliefert += n
	return n, err
}

func (w *a5Strom) Close() error { return w.inner.Close() }

// deckelStrom liefert Leerzeichen und nach deckel Bytes io.EOF. Der Leser verwirft Leerraum, sein Speicher wächst dabei nicht;
// der Deckel beendet den Strom auch für einen Leser, der die Größengrenze nicht durchsetzt. Der Fall wird dann aus dem
// benannten Grund rot: am Zähler der gelieferten Bytes (mehr als grenze + 1) und an `syntax` (leere Eingabe statt `zu_gross`).
type deckelStrom struct {
	deckel    int
	geliefert int
	closes    int
}

func (e *deckelStrom) Read(p []byte) (int, error) {
	n := min(len(p), e.deckel-e.geliefert)
	if n <= 0 {
		return 0, io.EOF
	}
	for i := range n {
		p[i] = ' '
	}
	e.geliefert += n
	return n, nil
}

func (e *deckelStrom) Close() error { e.closes++; return nil }

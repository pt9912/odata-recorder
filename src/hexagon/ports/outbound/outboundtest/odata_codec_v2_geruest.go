package outboundtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// v2G1: die Argumente (SPEC-172 Nr. 1, SPEC-126 Nr. 1 a bis d): ErrUngueltig ohne DatenFehler, src wird nicht gelesen. Schleife
// über die Verstöße und beide Methoden; grenze 1 ist gültig.
func v2G1(t *testing.T, s *jsuite) {
	n := 0
	for _, einzel := range []bool{true, false} {
		for _, f := range a1Faelle() {
			if f.soll == a1Ungueltig && !f.leer {
				s.pruefeA1(t, einzel, f, s.modelle["R"])
				n++
			}
		}
		r := s.aufruf(context.Background(), einzel, &teilReader{teile: [][]byte{[]byte("{")}}, s.modelle["R"], "Products", 1)
		datenFehler(t, methodeName(einzel)+" grenze 1", r.err, model.DatenSyntax, "")
	}
	if n != 2*10 {
		t.Errorf("G1: %d Läufe, erwartet 20 (10 Verstöße mal 2 Methoden)", n)
	}
}

// v2Versionen sind die Versionen, die der Leser ablehnt: leer, andere Versionen und Schreibweisen, die mit `2.` beginnen oder
// `2.0` enthalten.
func v2Versionen() []string {
	return []string{"", "1.0", "3.0", "4.0", "2.0 ", "2.00", "20", " 2.0", "2.", "v2.0"}
}

// v2G2: ein Modell mit der Version ungleich 2.0 ist ErrUngueltig, src wird nicht gelesen; 2.0 liest (SPEC-172 Nr. 1, E1).
func v2G2(t *testing.T, s *jsuite) {
	hin := context.Background
	for _, einzel := range []bool{true, false} {
		for _, v := range v2Versionen() {
			f := a1Fall{name: "Version " + fmt.Sprintf("%q", v), ctx: hin, src: true, grenze: 1000, set: "Products", soll: a1Ungueltig}
			s.pruefeA1(t, einzel, f, v2Version(s.modelle["R"], v))
		}
		f := a1Fall{name: "Nullwert-Modell", ctx: hin, src: true, grenze: 1000, set: "Products", leer: true, soll: a1Ungueltig}
		s.pruefeA1(t, einzel, f, s.modelle["R"])
		s.pruefe(t, v2Basis(einzel))
	}
}

// v2G3: die Prüffolge der Argumente (Zusatztabelle Z1): ein Fall je Paar, src wird nicht gelesen.
func v2G3(t *testing.T, s *jsuite) {
	hin, kein := context.Background, func() context.Context { return nil }
	falsch := v2Version(s.modelle["R"], "4.0")
	for _, einzel := range []bool{true, false} {
		for _, z := range []struct {
			f a1Fall
			m model.ServiceModel
		}{
			{a1Fall{"Version falsch und ctx beendet", abgebrochen, true, 1000, "Products", false, a1Ungueltig}, falsch},
			{a1Fall{"Version falsch und entitySet unbekannt", hin, true, 1000, "Nope", false, a1Ungueltig}, falsch},
			{a1Fall{"Nullwert-Modell und beliebiges Set", hin, true, 1000, "Products", true, a1Ungueltig}, s.modelle["R"]},
			{a1Fall{"ctx beendet und entitySet unbekannt", abgebrochen, true, 1000, "Nope", false, a1Kontext}, s.modelle["R"]},
			{a1Fall{"ctx nil und Version falsch", kein, true, 1000, "Products", false, a1Ungueltig}, falsch},
			{a1Fall{"Version falsch und src nil", hin, false, 1000, "Products", false, a1Ungueltig}, falsch},
		} {
			s.pruefeA1(t, einzel, z.f, z.m)
		}
		rd := &teilReader{teile: [][]byte{[]byte(v2EinR)}, ende: errQuelle}
		r := s.aufruf(context.Background(), einzel, rd, falsch, "Products", 0)
		if !errors.Is(r.err, model.ErrUngueltig) || errors.Is(r.err, errQuelle) || rd.reads != 0 {
			t.Errorf("%s Argumentverstoß und Fehler von src: %v, %d Reads", methodeName(einzel), r.err, rd.reads)
		}
	}
}

// v2G4: entitySet nennt kein Set des Modells, ein Set anderer Schreibweise, mit Container oder Namensraum, den Namen eines Typs,
// oder ein Set, dessen Typ das Modell nicht auflöst: entity_set, Ort der Name, src nicht gelesen (SPEC-126 Nr. 1 f).
func v2G4(t *testing.T, s *jsuite) {
	hin := context.Background
	namen := []string{"Nope", "products", "PRODUCTS", "Demo.Products", "Container/Products", "Product", "Demo.Product", "Products ",
		" Products", "Pro\x01ducts/~", "Products\x00", "Employee", "Ws"}
	for _, einzel := range []bool{true, false} {
		for _, n := range namen {
			s.pruefeA1(t, einzel, a1Fall{name: "Set " + fmt.Sprintf("%q", n), ctx: hin, src: true, grenze: 1000, set: n, soll: a1SetFehler}, s.modelle["R"])
		}
		ohneTyp := model.NeuesServiceModel("2.0", nil, nil, []model.EntitySet{{Name: "X", EntityType: "T.Fehlt"}}, nil)
		s.pruefeA1(t, einzel, a1Fall{name: "Set mit nicht aufgelöstem Typ", ctx: hin, src: true, grenze: 1000, set: "X", soll: a1SetFehler}, ohneTyp)
	}
}

// v2Belege sind die Körper der Zeilen G5 und G6: je Methode der gültige Body und ein Body mit einem Verstoß des Deutens.
func v2Belege(einzel bool) (basis, typFehler jfall) {
	basis = v2Basis(einzel)
	typFehler = jfall{m: "R", set: "Products", einzel: einzel, body: v2Kol(`{"ID":1,"Name":5}`)}
	if einzel {
		typFehler.body = v2Ein(`"ID":1,"Name":5`)
	}
	return basis, typFehler
}

// v2G5: der Kontext wird vor jedem Read und einmal vor dem Deuten geprüft, der Fehler kommt unverändert zurück (SPEC-126 Nr. 1 e):
// Schleife über die Kontextfehler, die Stellen und beide Methoden.
func v2G5(t *testing.T, s *jsuite) {
	for _, einzel := range []bool{true, false} {
		basis, typFehler := v2Belege(einzel)
		for _, ce := range kontextFehler() {
			for _, st := range a2Stellen(basis.body, typFehler.body) {
				s.pruefeA2(t, typFehler, ce, st)
			}
			ctx := neuBeendbar(nil)
			rd := teile([]byte(`{"a" 1}`))
			rd.nach = func(int) { ctx.beende(ce) }
			r := s.ruf(ctx, rd, typFehler)
			datenFehler(t, methodeName(einzel)+" Abbruch und Verstoß im selben Read ("+ce.Error()+")", r.err, model.DatenSyntax, "")
			rd = teile(basis.body)
			if r = s.ruf(neuBeendbar(ce), rd, basis); r.err != ce || rd.reads != 0 {
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

// v2G6: ein Fehler von src ist kein DatenFehler und bleibt über errors.Is erreichbar (SPEC-126 Nr. 2): Daten und Fehler im
// selben Read; ein Verstoß des Lesens in den Daten geht vor. Dazu jeder Fehler (12) zu jedem Zeitpunkt (2L + 3) beider Bodys.
func v2G6(t *testing.T, s *jsuite) {
	var laeufe, erwartet int
	for _, einzel := range []bool{true, false} {
		basis, _ := v2Belege(einzel)
		a3Quellfehler(t, s, einzel, basis, `"ID":1`, 5, 30)
		erwartet += a3FehlerZahl * (2*len(basis.body) + 3)
		laeufe += v2G6Zeitpunkte(t, s, einzel, basis)
	}
	if n := len(a3Fehlerarten()); n != a3FehlerZahl || laeufe != erwartet {
		t.Errorf("G6: %d Fehlerarten (erwartet %d), %d Läufe (erwartet %d)", n, a3FehlerZahl, laeufe, erwartet)
	}
	t.Logf("G6: %d Läufe", laeufe)
}

// v2G6Zeitpunkte fährt jeden Fehler von src zu jedem Zeitpunkt über basis und liefert die Zahl der Läufe.
func v2G6Zeitpunkte(t *testing.T, s *jsuite, einzel bool, basis jfall) int {
	laeufe := 0
	for _, fehler := range a3Fehlerarten() {
		for i, rd := range a3Zeitpunkte(basis.body, fehler) {
			r := s.ruf(context.Background(), rd, basis)
			name := fmt.Sprintf("%s Fehler %v Zeitpunkt %d", methodeName(einzel), fehler, i)
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
	return laeufe
}

// v2G7: Ergebnis und Fehler hängen nicht davon ab, wie src den Strom in Reads zerlegt (SPEC-126 Nr. 2). Die Zerlegungen von A4
// über alle Fälle der Suite (ganz, Byte für Byte, jede Zweiteilung, mit und ohne leere Reads) und `n > 0` zusammen mit io.EOF:
// der Body in einem Read mit dem Ende, in zwei Reads mit dem Ende im zweiten, Byte für Byte mit dem Ende im letzten.
func v2G7(t *testing.T, s *jsuite) {
	jA4(t, s)
	var laeufe, abweichungen int
	for _, f := range j1Faelle(t, s) {
		soll := s.lauf(f)
		einzeln := make([]int, 0, len(f.body))
		for i := 1; i < len(f.body); i++ {
			einzeln = append(einzeln, i)
		}
		for _, stuecke := range [][]int{nil, {len(f.body) / 2}, einzeln} {
			rd := teile(f.body, stuecke...)
			rd.mitEnde = true
			laeufe++
			if got := s.ruf(context.Background(), rd, f); !reflect.DeepEqual(got, soll) {
				if abweichungen++; abweichungen <= gMeldungen {
					t.Errorf("%s (%d Bytes) mit io.EOF im letzten Read, %d Stücke: %+v, ganz: %+v", f.name, len(f.body), len(stuecke)+1, got, soll)
				}
			}
		}
	}
	t.Logf("G7: %d Läufe mit Daten und io.EOF im selben Read, %d Abweichungen", laeufe, abweichungen)
}

// grenzStrom liefert doc in Reads (schnitt 0 ganz, -1 Byte für Byte, sonst zweiteilig) und zählt die Reads, die mehr anfordern,
// als grenze + 1 abzüglich der gelieferten Bytes zulässt (SPEC-126 Nr. 5: der Leser fordert nie mehr an).
type grenzStrom struct {
	doc       []byte
	schnitt   int
	grenze    int64
	geliefert int
	ueber     int
	closes    int
}

func (r *grenzStrom) Read(p []byte) (int, error) {
	if int64(len(p))-1 > r.grenze-int64(r.geliefert) {
		r.ueber++
	}
	if r.geliefert >= len(r.doc) {
		return 0, io.EOF
	}
	ende := len(r.doc)
	switch {
	case r.schnitt == -1:
		ende = r.geliefert + 1
	case r.schnitt > r.geliefert:
		ende = r.schnitt
	}
	n := copy(p, r.doc[r.geliefert:ende])
	r.geliefert += n
	return n, nil
}

func (r *grenzStrom) Close() error { r.closes++; return nil }

// v2FaelleG12 sind die Fälle der Größe als Fälle der Suite: ein Body mit Leerraum aufgefüllt bei verschiedenen Werten von
// grenze, und die Paare Verstoß davor und dahinter (Zusatztabelle Z2).
func v2FaelleG12() []jfall {
	var out []jfall
	for _, einzel := range []bool{true, false} {
		basis := v2Basis(einzel)
		body := string(basis.body) + "     "
		l := int64(len(body))
		for _, g := range []struct {
			name    string
			grenze  int64
			ursache model.DatenUrsache
		}{{"1", 1, model.DatenZuGross}, {"2", 2, model.DatenZuGross}, {"L-1", l - 1, model.DatenZuGross}, {"L", l, ""}, {"L+1", l + 1, ""}, {"MaxInt64", math.MaxInt64, ""}} {
			out = append(out, jfall{name: "G12 grenze " + g.name + " " + basis.name, m: "R", set: "Products", einzel: einzel, body: []byte(body), grenze: g.grenze, ursache: g.ursache})
		}
		for _, v := range v2Verstoesse() {
			davor := jfall{name: "G12 Verstoß " + v.name + " in Byte grenze " + basis.name, m: "R", set: "Products", einzel: einzel,
				body: []byte(v.body), grenze: int64(len(v.prefix)), ursache: v.ursache}
			dahinter := davor
			dahinter.name, dahinter.grenze, dahinter.ursache = "G12 Verstoß "+v.name+" in Byte grenze+1 "+basis.name, int64(len(v.prefix))-1, model.DatenZuGross
			out = append(out, davor, dahinter)
		}
	}
	return out
}

// v2Verstoss ist ein Body, dessen erster Verstoß des Lesens am letzten Byte von prefix festgestellt wird.
type v2Verstoss struct {
	name, prefix, body string
	ursache            model.DatenUrsache
}

// v2Verstoesse sind die vier Verstöße des Lesens mit dem Byte, an dem sie feststehen: syntax am ersten Byte, das die Grammatik
// nicht erlaubt, kodierung am ersten ungültigen Byte, doppelter_name am Ende des zweiten Namens (vor seinem Wert) und tiefe an
// der 513. offenen Klammer.
func v2Verstoesse() []v2Verstoss {
	mk := func(name, prefix string, u model.DatenUrsache) v2Verstoss {
		return v2Verstoss{name, prefix, prefix + "  \"x\":1}}   ", u}
	}
	var tief string
	for range 513 {
		tief += "["
	}
	return []v2Verstoss{
		mk("syntax", `{"d" 1`, model.DatenSyntax),
		mk("kodierung", "{\"d\":\"\xff", model.DatenKodierung),
		mk("doppelter_name", `{"d":1,"d"`, model.DatenDoppelterName),
		mk("tiefe", tief, model.DatenTiefe),
	}
}

// v2G12: höchstens grenze Bytes; das Byte grenze + 1 ist zu_gross, davor entscheidet ein Verstoß; MaxInt64 ist gültig; der
// Leser fordert nie mehr als grenze + 1 Bytes an (SPEC-126 Nr. 5). Schleife über grenze, die Zerlegung des Stroms und die Methode.
func v2G12(t *testing.T, s *jsuite) {
	var laeufe int
	for _, f := range v2FaelleG12() {
		for _, schnitt := range []int{0, -1, len(f.body) / 2} {
			rd := &grenzStrom{doc: f.body, schnitt: schnitt, grenze: f.grenzeOderStandard()}
			r := s.ruf(context.Background(), rd, f)
			name := fmt.Sprintf("%s zerlegt bei %d", f.name, schnitt)
			if f.ursache != "" {
				datenFehler(t, name, r.err, f.ursache, "")
				r.nullwerte(t, name)
			} else if r.err != nil {
				t.Errorf("%s: Fehler %v bei gültigem Body", name, r.err)
			}
			if rd.ueber != 0 || rd.closes != 0 {
				t.Errorf("%s: %d Reads fordern mehr als grenze+1 an, Close %d", name, rd.ueber, rd.closes)
			}
			laeufe++
		}
	}
	if laeufe != 3*len(v2FaelleG12()) {
		t.Errorf("G12: %d Läufe", laeufe)
	}
	v2G12Puffer(t, s)
	v2G12Gross(t, s)
	t.Logf("G12: %d Läufe der Zerlegung", laeufe)
}

// v2G12Puffer fährt die Menge P5 (die Größe des Puffers ist Sache des Adapters): grenze = m·P + d für jede Zweierpotenz P von
// 256 Bytes bis 128 KiB, m in {1, 2, 3} und d in {−1, 0, +1}, der Body der Länge grenze−1, grenze und grenze+1, ganz und in
// Reads zu je P Bytes, je Methode.
func v2G12Puffer(t *testing.T, s *jsuite) {
	const hoechst = 3<<17 + 1
	leerraum := make([]byte, hoechst)
	for i := range leerraum {
		leerraum[i] = ' '
	}
	laeufe := b5PLaeufe()
	for _, l := range laeufe {
		basis := `{"d":{"results":[]}}`
		if l.einzel {
			basis = `{"d":{"ID":1}}`
		}
		rd := &stueckStrom{doc: append([]byte(basis), leerraum[:l.laenge-int64(len(basis))]...), stueck: l.stueck}
		r := s.aufruf(context.Background(), l.einzel, rd, s.modelle["R"], "Products", l.grenze())
		name := fmt.Sprintf("G12 P5 %s P=%d m=%d d=%d Länge %d grenze %d Stück %d", methodeName(l.einzel), l.p, l.m, l.d, l.laenge, l.grenze(), l.stueck)
		if l.laenge > l.grenze() {
			datenFehler(t, name, r.err, model.DatenZuGross, "")
			r.nullwerte(t, name)
		} else if r.err != nil {
			t.Errorf("%s: Fehler %v bei gültigem Body", name, r.err)
		}
		if int64(rd.geliefert) > l.grenze()+1 {
			t.Errorf("%s: %d Bytes geliefert, höchstens grenze+1 = %d", name, rd.geliefert, l.grenze()+1)
		}
	}
	if len(laeufe) != b5PZahl {
		t.Errorf("G12 P5: %d Läufe, erwartet %d", len(laeufe), b5PZahl)
	}
}

// v2G12Gross fährt die Menge P6 (große Grenzen, grenze = 2^k + d): gültig bei der Länge grenze, zu_gross bei grenze + 1; der
// Strom erzeugt den Body und zählt, was der Leser anfordert.
func v2G12Gross(t *testing.T, s *jsuite) {
	laeufe := b6Laeufe()
	for _, l := range laeufe {
		basis := `{"d":{"results":[]}}`
		if l.einzel {
			basis = `{"d":{"ID":1}}`
		}
		laenge := l.grenze()
		if l.ueber {
			laenge++
		}
		rd := &erzeugerStrom{kopf: basis, laenge: laenge}
		if l.ueber {
			rd.laenge += b6Deckel
		}
		r := s.aufruf(context.Background(), l.einzel, rd, s.modelle["R"], "Products", l.grenze())
		name := fmt.Sprintf("G12 P6 %s k=%d d=%d Länge %d grenze %d", methodeName(l.einzel), l.k, l.d, laenge, l.grenze())
		if l.ueber {
			datenFehler(t, name, r.err, model.DatenZuGross, "")
			r.nullwerte(t, name)
		} else if r.err != nil {
			t.Errorf("%s: Fehler %v bei gültigem Body", name, r.err)
		}
		if rd.geliefert > l.grenze()+1 {
			t.Errorf("%s: %d Bytes geliefert, höchstens grenze+1 = %d", name, rd.geliefert, l.grenze()+1)
		}
	}
	if erwartet := map[bool]int{true: 24, false: 72}[unterDetektor]; len(laeufe) != erwartet {
		t.Errorf("G12 P6: %d Läufe, erwartet %d", len(laeufe), erwartet)
	}
}

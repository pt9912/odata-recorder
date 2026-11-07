package inspection_test

import (
	"context"
	"errors"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

// ohneCode verlangt einen Argumentverstoß: ErrUngueltig, kein *inbound.Fehler.
func ohneCode(t *testing.T, was string, err error) {
	t.Helper()
	var fe *inbound.Fehler
	if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &fe) {
		t.Errorf("%s: Fehler %v, erwartet ErrUngueltig ohne Code", was, err)
	}
}

// TestNewExchangesNil (K1): ein nil in Exchanges weist New ab, eine gültige Konfiguration nicht.
func TestNewExchangesNil(t *testing.T) {
	d, err := inspection.New(inspection.Konfig{Exchanges: nil})
	if d != nil {
		t.Error("New liefert einen Dienst ohne Exchanges")
	}
	ohneCode(t, "Exchanges nil", err)
	if d, err = inspection.New(inspection.Konfig{Exchanges: neuesFake()}); d == nil || err != nil {
		t.Errorf("gültige Konfiguration: %v, %v", d, err)
	}
}

// TestKontextNil (A1): ctx nil ist ErrUngueltig ohne Code und ruft keinen Port, auch zusammen mit
// einem ungültigen Seitenparameter.
func TestKontextNil(t *testing.T) {
	for _, s := range []inbound.Seite{{}, {Nach: -1}, {Limit: -1}} {
		f := neuesFake()
		f.legeAn(sid, 3)
		z, err := neuerDienst(t, f).Exchanges(nilCtx(), sid, s)
		ohneCode(t, "ctx nil", err)
		if !reflect.DeepEqual(z, inbound.ExchangeSeite{}) {
			t.Errorf("Seite %+v: Ergebnis %+v statt Nullwert", s, z)
		}
		if f.gesamt() != 0 || len(f.kontexte()) != 0 {
			t.Errorf("Seite %+v: ein Port wurde gerufen", s)
		}
	}
}

// TestSeitenparameterNegativ (X1): ein Nach oder Limit unter 0 ist ErrUngueltig ohne Code und ohne
// den Wert im Text, kein Port wird gerufen; 0 ist gültig.
func TestSeitenparameterNegativ(t *testing.T) {
	werte := []int{-1, -2, math.MinInt}
	for _, name := range []string{"Nach", "Limit"} {
		for _, v := range werte {
			s := inbound.Seite{}
			if name == "Nach" {
				s.Nach = v
			} else {
				s.Limit = v
			}
			f := neuesFake()
			f.legeAn(sid, 3)
			z, err := neuerDienst(t, f).Exchanges(context.Background(), sid, s)
			ohneCode(t, name+" "+strconv.Itoa(v), err)
			if err != nil && strings.Contains(err.Error(), strconv.Itoa(v)) {
				t.Errorf("%s %d: der Text nennt den Wert: %v", name, v, err)
			}
			if f.gesamt() != 0 || !reflect.DeepEqual(z, inbound.ExchangeSeite{}) {
				t.Errorf("%s %d: Port gerufen oder Ergebnis %+v", name, v, z)
			}
		}
	}
	f := neuesFake()
	f.legeAn(sid, 3)
	if z := lies(t, neuerDienst(t, f), sid, inbound.Seite{Nach: 0, Limit: 0}); len(z.Eintraege) != 3 {
		t.Errorf("Nach 0 und Limit 0: %d Einträge", len(z.Eintraege))
	}
}

// Achsen der Menge M-Seite (Plan §3, Pflicht-Mengen).
func limitAchse() []int {
	return []int{0, 1, vertragStandard, vertragMax - 1, vertragMax, vertragMax + 1, math.MaxInt}
}
func bestandAchse(l int) []int { return []int{0, 1, l - 1, l, l + 1, 2 * l, 2*l + 1} }
func nachAchse(n, l int) []int { return []int{0, 1, n - 1, n, n + 1, n + l} }

type seitenFall struct{ limit, n, nach int }

// seitenFaelle erzeugt die Menge M-Seite ohne ungültige Zeiger und ohne gleiche Fälle; roh ist die
// Zahl vor dem Entfernen, formel das Produkt der Längen der Achsen.
func seitenFaelle() (faelle []seitenFall, roh, formel int) {
	gesehen := map[seitenFall]bool{}
	for _, limit := range limitAchse() {
		l := wirksam(limit)
		for _, n := range bestandAchse(l) {
			for _, nach := range nachAchse(n, l) {
				roh++
				c := seitenFall{limit, n, nach}
				if nach < 0 || gesehen[c] {
					continue
				}
				gesehen[c] = true
				faelle = append(faelle, c)
			}
		}
	}
	return faelle, roh, len(limitAchse()) * len(bestandAchse(1)) * len(nachAchse(1, 1))
}

// TestMengeSeiteZaehlung: die Menge M-Seite hat die Zahl der Fälle aus den Achsen, enthält die
// Grenzen und der Nullfall (leere Session) ist darin.
func TestMengeSeiteZaehlung(t *testing.T) {
	faelle, roh, formel := seitenFaelle()
	if roh != formel || len(faelle) == 0 || len(faelle) > roh {
		t.Fatalf("roh %d, Formel %d, Fälle %d", roh, formel, len(faelle))
	}
	for _, l := range limitAchse() {
		if !slices.ContainsFunc(faelle, func(c seitenFall) bool { return c.limit == l && c.n == 0 && c.nach == 0 }) {
			t.Errorf("Limit %d: der Fall der leeren Session fehlt", l)
		}
	}
}

// TestSeitenModell (X2, X3, X4): jede Seite der Menge M-Seite gleicht dem Modell: Eintraege, Gesamt,
// Weiter, Limit (0: Standard, über der Obergrenze: gekappt), Reihenfolge, die Grenzen von Nach und
// Limit gegen Gesamt und die leere Liste.
func TestSeitenModell(t *testing.T) {
	faelle, _, _ := seitenFaelle()
	for _, c := range faelle {
		f := neuesFake()
		f.legeAn(sid, c.n)
		z := lies(t, neuerDienst(t, f), sid, inbound.Seite{Limit: c.limit, Nach: c.nach})
		gleichSeite(t, "Limit "+strconv.Itoa(c.limit)+" n "+strconv.Itoa(c.n)+" Nach "+strconv.Itoa(c.nach), z, c.n, c.nach, c.limit)
	}
}

// TestSeitenFolge (X5): die Folge der Seiten ab jedem Nach der Menge liefert jeden Exchange genau
// einmal in Reihenfolge, und gleiche Argumente ergeben das gleiche Ergebnis.
func TestSeitenFolge(t *testing.T) {
	faelle, _, _ := seitenFaelle()
	for _, c := range faelle {
		f := neuesFake()
		f.legeAn(sid, c.n)
		d := neuerDienst(t, f)
		seiten := folge(t, d, sid, c.limit, c.nach)
		if got, want := nummern(seiten...), bereich(c.nach+1, c.n); !slices.Equal(got, want) {
			t.Errorf("Limit %d n %d Nach %d: gelesen %v, erwartet %v", c.limit, c.n, c.nach, got, want)
		}
		s := inbound.Seite{Limit: c.limit, Nach: c.nach}
		if a, b := lies(t, d, sid, s), lies(t, d, sid, s); !reflect.DeepEqual(a, b) {
			t.Errorf("Limit %d n %d Nach %d: zwei Aufrufe, zwei Ergebnisse", c.limit, c.n, c.nach)
		}
	}
}

// TestSeitenGoldens: Werte als Zahlen, nicht aus dem Modell, an den Grenzen des Plans (X4).
func TestSeitenGoldens(t *testing.T) {
	fall := []struct {
		n, nach, limit     int
		nums               []int
		weiter, gesamt, lm int
	}{
		{3, 0, 3, []int{1, 2, 3}, 0, 3, 3},
		{3, 0, 2, []int{1, 2}, 2, 3, 2},
		{3, 0, 4, []int{1, 2, 3}, 0, 3, 4},
		{3, 2, 1, []int{3}, 0, 3, 1},
		{3, 3, 1, []int{}, 0, 3, 1},
		{3, 4, 1, []int{}, 0, 3, 1},
		{0, 0, 0, []int{}, 0, 0, 100},
		{3, 0, math.MaxInt, []int{1, 2, 3}, 0, 3, 1000},
	}
	for _, c := range fall {
		f := neuesFake()
		f.legeAn(sid, c.n)
		z := lies(t, neuerDienst(t, f), sid, inbound.Seite{Limit: c.limit, Nach: c.nach})
		if !slices.Equal(nummern(z), c.nums) || z.Weiter != c.weiter || z.Gesamt != c.gesamt || z.Limit != c.lm || z.Eintraege == nil {
			t.Errorf("%+v: erhalten %v Weiter %d Gesamt %d Limit %d", c, nummern(z), z.Weiter, z.Gesamt, z.Limit)
		}
	}
}

// TestGrenzenDesVertrags (X2): Standard 100 und Obergrenze 1000 aus SPEC-236 Nr. 3 als Literale des
// Tests, nicht aus dem Paket: die Menge der Limits an beiden Grenzen (je Grenze: darunter, genau,
// darüber, ein weiterer Wert darüber) bei einem Bestand über der Obergrenze.
func TestGrenzenDesVertrags(t *testing.T) {
	f := neuesFake()
	f.legeAn(sid, 1100)
	d := neuerDienst(t, f)
	fall := []struct{ limit, erwartet int }{
		{0, 100}, {1, 1}, {99, 99}, {100, 100}, {101, 101}, {500, 500},
		{999, 999}, {1000, 1000}, {1001, 1000}, {1002, 1000}, {1100, 1000}, {math.MaxInt, 1000},
	}
	for _, c := range fall {
		z := lies(t, d, sid, inbound.Seite{Limit: c.limit})
		if len(z.Eintraege) != c.erwartet || z.Limit != c.erwartet {
			t.Errorf("Limit %d: %d Einträge, Limit der Antwort %d, erwartet %d", c.limit, len(z.Eintraege), z.Limit, c.erwartet)
		}
	}
}

// TestSequenzMitLuecken (X3, X5): der Zeiger ist der Wert der sequence_no, nicht die Zahl der gelesenen
// Einträge: Sessions mit Lücken in den Nummern, das Modell über die Nummern, die Folge der Seiten ab
// jedem Zeiger. Der Port liefert keine Lücken (SPEC-105 Nr. 8); der Dienst filtert nach dem Wert.
func TestSequenzMitLuecken(t *testing.T) {
	listen := [][]int{{2, 5, 6, 10, 11}, {7}, {1, 3}, {100, 200, 300, 400}}
	faelle := 0
	for _, nums := range listen {
		zeiger := []int{0, nums[len(nums)-1] + 5}
		for _, q := range nums {
			zeiger = append(zeiger, q-1, q, q+1)
		}
		for _, limit := range []int{0, 1, 2, 3} {
			for _, nach := range zeiger {
				faelle++
				f := neuesFake()
				f.setzeNummern(sid, nums)
				d := neuerDienst(t, f)
				want, weiter := modellListe(nums, nach, wirksam(limit))
				z := lies(t, d, sid, inbound.Seite{Limit: limit, Nach: nach})
				if !slices.Equal(nummern(z), want) || z.Weiter != weiter || z.Gesamt != len(nums) {
					t.Errorf("%v Limit %d Nach %d: %v Weiter %d Gesamt %d, erwartet %v Weiter %d", nums, limit, nach, nummern(z), z.Weiter, z.Gesamt, want, weiter)
				}
				rest, _ := modellListe(nums, nach, len(nums))
				if got := nummern(folge(t, d, sid, limit, nach)...); !slices.Equal(got, rest) {
					t.Errorf("%v Limit %d Nach %d: Folge %v, erwartet %v", nums, limit, nach, got, rest)
				}
			}
		}
	}
	if faelle == 0 {
		t.Error("kein Fall gelaufen")
	}
}

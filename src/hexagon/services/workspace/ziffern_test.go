package workspace_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// zahlEnts sind n verschiedene Einträge.
func zahlEnts(n int) []ent {
	r := make([]ent, 0, n)
	for i := 0; i < n; i++ {
		s := strconv.Itoa(i)
		r = append(r, ent{"Products", "Id=" + s, `{"Id":"` + s + `"}`, []string{"Id"}, "complete"})
	}
	return r
}

// anzahlen ist die Menge der Anzahlen von Einträgen: 0 bis 130 und die Hunderter 200 bis 900 und 999. Sie
// führt jede der zehn Ziffern an der Einerstelle, jede der Ziffern 1 bis 9 an der Zehner- und der Hunderterstelle
// und die 0 an der Zehnerstelle (100 bis 109).
func anzahlen() []int {
	var r []int
	for n := 0; n <= 130; n++ {
		r = append(r, n)
	}
	for h := 2; h <= 9; h++ {
		r = append(r, h*100)
	}
	return append(r, 999)
}

// TestEntityCountZiffern (Z72, SPEC-129 Nr. 6 b „Ziffern“): ein Manifest mit entity_count gleich der Zahl der
// Einträge gilt für jede Anzahl der Menge, also mit jeder Ziffer an jeder Stelle; derselbe Zähler um 1, 10
// und 100 zu hoch (und um 1 zu niedrig) ist ODR-E3001, auch wenn er nur aus gültigen Ziffern besteht.
func TestEntityCountZiffern(t *testing.T) {
	ns := anzahlen()
	// Die Menge deckt die Ziffern an den Stellen: Stelle 0 und 1 mit 0 bis 9, Stelle 2 mit 1 bis 9.
	gedeckt := map[[2]int]bool{}
	for _, n := range ns {
		s := strconv.Itoa(n)
		for i := 0; i < len(s); i++ {
			gedeckt[[2]int{len(s) - 1 - i, int(s[i] - '0')}] = true
		}
	}
	for stelle := 0; stelle < 3; stelle++ {
		for d := 0; d <= 9; d++ {
			if stelle == 2 && d == 0 {
				continue
			}
			if !gedeckt[[2]int{stelle, d}] {
				t.Errorf("Ziffer %d an Stelle %d nicht in der Menge", d, stelle)
			}
		}
	}
	w := neu(t)
	gut, schlecht := 0, 0
	for _, n := range ns {
		baum := func(zaehler int) []byte {
			m := manifestBaum("bsl-1", "svc", zahlEnts(n))
			m["entity_count"] = roh(strconv.Itoa(zaehler))
			return []byte(schreibe(m))
		}
		w.baselineBytes("bsl-1", "svc", baum(n))
		if _, err := w.v.Create(context.Background(), "bsl-1", "ws-"+strconv.Itoa(n)); err != nil {
			t.Errorf("entity_count %d bei %d Einträgen: %v", n, n, err)
		}
		gut++
		falsch := []int{n + 1, n + 10, n + 100}
		if n > 0 {
			falsch = append(falsch, n-1)
		}
		for _, z := range falsch {
			w.baselineBytes("bsl-1", "svc", baum(z))
			_, err := w.v.Create(context.Background(), "bsl-1", "ws-falsch")
			pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
			schlecht++
		}
	}
	t.Logf("Anzahlen %d, gültig %d, abgewiesen %d", len(ns), gut, schlecht)
}

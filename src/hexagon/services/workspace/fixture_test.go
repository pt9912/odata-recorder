package workspace_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

// gleicheFelder nennt jedes Paar von Feldern des Struct v aus Zeichenketten oder Zeiten, die denselben Wert
// tragen, als "A=B".
func gleicheFelder(v any) []string {
	rv := reflect.ValueOf(v)
	var namen, werte []string
	for i := 0; i < rv.NumField(); i++ {
		if k := rv.Type().Field(i).Type; k.Kind() == reflect.String || k == reflect.TypeOf(time.Time{}) {
			namen = append(namen, rv.Type().Field(i).Name)
			werte = append(werte, fmt.Sprint(rv.Field(i).Interface()))
		}
	}
	var r []string
	for i := range werte {
		for j := i + 1; j < len(werte); j++ {
			if werte[i] == werte[j] {
				r = append(r, namen[i]+"="+namen[j])
			}
		}
	}
	sort.Strings(r)
	return r
}

// TestFixtureFelderVerschieden (Z73, M-3): in jeder Fixture der Struct-Arten, die der Dienst liest oder
// schreibt (Baseline, Workspace, Eintrag des Entity-States), tragen die Felder aus Zeichenketten und Zeiten
// paarweise verschiedene Werte, mit einer Ausnahme: StateHash gleich StateBlob verlangt SPEC-129 Nr. 5 b. Sonst
// bliebe ein Aufruf mit dem falschen Feld (Open mit MetadataHash) oder ein vertauschtes Feld grün.
func TestFixtureFelderVerschieden(t *testing.T) {
	w := neu(t)
	w.vorbelegen("ws-1", 0)
	fixtures := map[string]struct {
		v       any
		erlaubt []string
	}{
		"Baseline":  {w.snaps.baselines["bsl-1"], []string{"StateBlob=StateHash"}},
		"Workspace": {w.st.stand().ws["ws-1"], nil},
	}
	for i, e := range erwarteterState("ws-1", standardEnts()) {
		fixtures[fmt.Sprintf("Entity %d", i)] = struct {
			v       any
			erlaubt []string
		}{e, nil}
	}
	for name, f := range fixtures {
		got := gleicheFelder(f.v)
		if fmt.Sprint(got) != fmt.Sprint(f.erlaubt) {
			t.Errorf("%s: Felder gleichen Wertes %v, erlaubt %v", name, got, f.erlaubt)
		}
	}
	if len(fixtures) != 5 {
		t.Errorf("%d Fixtures", len(fixtures))
	}
}

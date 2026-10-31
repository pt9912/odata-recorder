package replay_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestResponseBodyZustand hält Z65: Menge 8; ein gewählter Exchange, dessen Response-Body weder present noch absent
// ist (Zustand discarded, unbekannter Wert x, leerer Wert; die Fixture setzt ihn nach Validate), ist ODR-E9001
// (500). Open und Put werden nicht gerufen, Cursors.Get genau einmal, der Store bleibt unverändert, der Strom ist
// nil, und Error() nennt keinen Wert. Die Gegenprobe liefert absent und present wie Z34 und Z35.
func TestResponseBodyZustand(t *testing.T) {
	type element struct {
		zustand model.BodyState
		strat   model.ReplayStrategy
	}
	var menge []element
	for _, z := range []model.BodyState{model.BodyDiscarded, "x", ""} {
		for _, s := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict} {
			menge = append(menge, element{z, s})
		}
	}
	menge = append(menge, element{model.BodyDiscarded, model.StrategyFirst}, element{model.BodyDiscarded, model.StrategyLast})
	for _, el := range menge {
		was := fmt.Sprintf("Z65 Zustand %q %s", el.zustand, el.strat)
		u := neu(t)
		u.k.Strategie = el.strat
		e := u.austausch(1, u.keyA(), true)
		e.ResponseBody = el.zustand
		u.lege(e)
		vorher := u.st.stand()
		resp, err := u.replay(context.Background(), anf())
		erwarteCode(t, was, err, model.CodeE9001)
		if _, status, ok := fehlerCode(err); ok && status != 500 {
			t.Errorf("%s: Status %d, erwartet 500", was, status)
		}
		if err != nil && err.Error() != model.CodeE9001.String()+": "+model.CodeE9001.Message() {
			t.Errorf("%s: Error() %q", was, err.Error())
		}
		if resp.Body != nil {
			t.Errorf("%s: Strom nicht nil", was)
		}
		for op, want := range map[string]int{"Blobs.Open": 0, "tx Cursors.Put": 0, "tx Cursors.Get": 1} {
			if got := u.p.anzahl(op); got != want {
				t.Errorf("%s: %s %d-mal gerufen, erwartet %d", was, op, got, want)
			}
		}
		if !reflect.DeepEqual(u.st.stand(), vorher) {
			t.Errorf("%s: Store geändert", was)
		}
	}
	if len(menge) != 8 {
		t.Errorf("Menge: %d, erwartet 8", len(menge))
	}
	for _, g := range []struct {
		zustand model.BodyState
		opens   int
	}{{model.BodyAbsent, 0}, {model.BodyPresent, 1}} {
		u := neu(t)
		e := u.austausch(1, u.keyA(), true)
		if g.zustand == model.BodyAbsent {
			e = ohneBody(e)
		}
		u.lege(e)
		resp, err := u.replay(context.Background(), anf())
		if err != nil {
			t.Errorf("Gegenprobe %s: %v", g.zustand, err)
			continue
		}
		_ = lies(t, resp)
		_ = resp.Body.Close()
		if got := u.p.anzahl("Blobs.Open"); got != g.opens {
			t.Errorf("Gegenprobe %s: Open %d-mal, erwartet %d", g.zustand, got, g.opens)
		}
	}
}

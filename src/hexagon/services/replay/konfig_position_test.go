package replay_test

import (
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// TestNewKonfigPosition hält Z1 für die Position in den Mengen: Menge 10 (3 Verstöße × 3 Positionen in einer Liste
// von drei Einträgen, dazu ein Gutfall mit drei gültigen Einträgen je Liste). New prüft jeden Eintrag von Dienste
// und Pfade, nicht nur den ersten oder den letzten.
func TestNewKonfigPosition(t *testing.T) {
	gueltigePfade := func() []model.JSONPfad {
		var r []model.JSONPfad
		for _, s := range []string{"$..password", "$..token", "$..secret"} {
			p, err := model.ParseJSONPfad(s)
			if err != nil {
				t.Fatal(err)
			}
			r = append(r, p)
		}
		return r
	}
	gueltigeDienste := func() []replay.Dienst {
		return []replay.Dienst{{ID: "d0", Root: "/d0"}, {ID: "d1", Root: "/d1"}, {ID: "d2", Root: "/d2"}}
	}
	verstoesse := []struct {
		name string
		mod  func(k *replay.Konfig, pos int)
	}{
		{"leere Kennung", func(k *replay.Konfig, pos int) { k.Dienste[pos].ID = "" }},
		{"leeres Root", func(k *replay.Konfig, pos int) { k.Dienste[pos].Root = "" }},
		{"Nullwert in Pfade", func(k *replay.Konfig, pos int) { k.Pfade[pos] = model.JSONPfad{} }},
	}
	n := 0
	for _, v := range verstoesse {
		for pos := 0; pos < 3; pos++ {
			n++
			u := neu(t)
			u.k.Dienste, u.k.Pfade = gueltigeDienste(), gueltigePfade()
			v.mod(&u.k, pos)
			r, err := replay.New(u.k)
			if r != nil || !errors.Is(err, model.ErrUngueltig) {
				t.Errorf("%s an Position %d: Dienst %v, Fehler %v", v.name, pos, r, err)
			}
		}
	}
	n++
	u := neu(t)
	u.k.Dienste, u.k.Pfade = gueltigeDienste(), gueltigePfade()
	if r, err := replay.New(u.k); r == nil || err != nil {
		t.Errorf("Gutfall abgewiesen: %v", err)
	}
	if n != 10 {
		t.Errorf("Menge: %d, erwartet 10", n)
	}
}

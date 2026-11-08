package mock_test

import (
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// etagDynamisch macht den URL-Leser abhängig vom Pfad hinter dem Root: /PATCH/<Schlüssel> ist eine Mutation der Entity, /GE/<Schlüssel> das Lesen
// der Entity, /GC/x das Lesen der Collection; der Leser der Bodies liefert Id und Name zum Schlüssel.
func (u *umg) etagDynamisch() {
	u.dynamisch()
	u.c.url.f = func(pfad, _ string) (model.ODataURL, error) {
		teile := strings.Split(strings.Trim(pfad, "/"), "/")
		m, key := teile[len(teile)-2], teile[len(teile)-1]
		if m == "GC" {
			return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: setProducts}}, nil
		}
		return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: setProducts, Schluessel: []model.SchluesselTeil{zahl("", key)}}}, nil
	}
}

// TestEtagGleichzeitigeAenderungen (E18): n gleichzeitige erfolgreiche PATCH derselben Entity mit Version v geben n verschiedene ETags, genau die
// der Versionen v+1 bis v+n; der Endstand ist der ETag von v+n. n = 2, 8, 32, je 20 Wiederholungen, je Wiederholung ein frischer Aufbau.
func TestEtagGleichzeitigeAenderungen(t *testing.T) {
	for _, n := range []int{2, 8, 32} {
		for wdh := 0; wdh < 20; wdh++ {
			const v = 5
			u := etagUmg(t, true, v, v, v)
			u.etagDynamisch()
			m := u.mocker()
			var mu sync.Mutex
			var etags []string
			aufEinmal(n, func(i int) {
				a := dynAnfrage("PATCH", "1", "n"+strconv.Itoa(i), wsID)
				ant, err := m.Mock(u.ctx, a)
				if err != nil || ant.Status != 204 {
					t.Errorf("n=%d: Status %d, Fehler %v", n, ant.Status, err)
					return
				}
				w, _ := etagWerte(ant)
				mu.Lock()
				etags = append(etags, w...)
				mu.Unlock()
			})
			slices.Sort(etags)
			var want []string
			for k := 1; k <= n; k++ {
				want = append(want, etagLit(int64(v+k)))
			}
			slices.Sort(want)
			if !slices.Equal(etags, want) {
				t.Fatalf("n=%d Wiederholung %d: ETags %v, erwartet %v", n, wdh, etags, want)
			}
			if got := u.etagVon("1"); got != etagLit(int64(v+n)) {
				t.Errorf("n=%d: Endstand %q, erwartet %q", n, got, etagLit(int64(v+n)))
			}
		}
	}
}

// TestEtagLesenWaehrendAenderung (E19): m Änderungen gleichzeitig mit l Lesern (Entity und Collection): in jeder Antwort ist der Header gleich dem
// Mitglied, die Version liegt zwischen v und v+m, und die Folge der Lesungen eines Lesers fällt nie.
func TestEtagLesenWaehrendAenderung(t *testing.T) {
	const m, l, lesungen, v = 8, 4, 40, 3
	for wdh := 0; wdh < 20; wdh++ {
		u := etagUmg(t, true, v, v, v)
		u.c.sch.zeigeEtag = true
		u.etagDynamisch()
		dienst := u.mocker()
		ver := func(etag string) int {
			x, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(etag, `W/"`), `"`))
			if err != nil {
				t.Errorf("ETag %q: %v", etag, err)
			}
			return x
		}
		var mu sync.Mutex
		bericht := func(f string, a ...any) {
			mu.Lock()
			defer mu.Unlock()
			t.Errorf(f, a...)
		}
		aufEinmal(m+l, func(i int) {
			if i < m {
				_, err := dienst.Mock(u.ctx, dynAnfrage("PATCH", "1", "n"+strconv.Itoa(i), wsID))
				if err != nil {
					bericht("Änderung %d: %v", i, err)
				}
				return
			}
			vorige := 0
			for k := 0; k < lesungen; k++ {
				pfad := "/GE/1"
				if i%2 == 0 {
					pfad = "/GC/x"
				}
				a := inbound.Anfrage{Methode: "GET", Pfad: dienstRoot + pfad, Header: map[string][]string{"X-Recorder-Workspace": {wsID}}, Laenge: -1}
				ant, err := dienst.Mock(u.ctx, a)
				if err != nil {
					bericht("Leser %d: %v", i, err)
					return
				}
				w, _ := etagWerte(ant)
				b, _ := io.ReadAll(ant.Body)
				_ = ant.Body.Close()
				text := string(b)
				var mitglied string
				if pfad == "/GE/1" {
					mitglied = text[strings.LastIndex(text, "|")+1:]
					if len(w) != 1 || w[0] != mitglied {
						bericht("Leser %d: Header %v, Mitglied %q", i, w, mitglied)
						return
					}
				} else {
					if len(w) != 0 {
						bericht("Leser %d: Collection trägt den ETag %v", i, w)
					}
					teile := strings.Split(text, "|")
					for _, teil := range teile {
						if strings.HasPrefix(teil, "1=") {
							mitglied = strings.TrimPrefix(teil, "1=")
						}
					}
				}
				x := ver(mitglied)
				if x < v || x > v+m || x < vorige {
					bericht("Leser %d Lesung %d: Version %d außerhalb %d bis %d oder unter der vorigen %d", i, k, x, v, v+m, vorige)
					return
				}
				vorige = x
			}
		})
	}
}

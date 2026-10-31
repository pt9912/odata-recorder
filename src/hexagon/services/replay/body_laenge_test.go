package replay_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// lesezaehler ist ein Body, der jeden Aufruf von Read und die gelesenen Bytes zählt; fehler ist das Ergebnis jedes
// Read.
type lesezaehler struct {
	daten   []byte
	fehler  error
	aufrufe int
	gelesen int
}

func (l *lesezaehler) Read(p []byte) (int, error) {
	l.aufrufe++
	if l.fehler != nil {
		return 0, l.fehler
	}
	if len(l.daten) == 0 {
		return 0, io.EOF
	}
	n := copy(p, l.daten)
	l.daten = l.daten[n:]
	l.gelesen += n
	return n, nil
}

// TestLaengeUndBody hält Z3 und Z17 für die Steuerung durch Body und Laenge (SPEC-135 Nr. 2, 7 a): Menge 96 = Laenge
// {-1, 0, 1, MaxBytes-1, MaxBytes, MaxBytes+1} × Body {nil, leer, nicht leer, nicht lesbar} × Strategie {4}. Nur Body
// nil und Laenge 0 heißen kein Body: er wird nicht gelesen (kein Read-Aufruf) und der Key ist der ohne Body; sonst wird
// er gelesen, und Laenge gegen die gelesenen Bytes bleibt ungeprüft. Body nil bei Laenge größer 0 ist model.ErrUngueltig
// ohne Port. Das Orakel entsteht aus dem Wortlaut der Spezifikation, nicht aus dem Code.
func TestLaengeUndBody(t *testing.T) {
	const grenze = 8
	laengen := []int64{-1, 0, 1, grenze - 1, grenze, grenze + 1}
	bodies := []struct {
		name string
		neu  func() *lesezaehler
	}{
		{"nil", nil},
		{"leer", func() *lesezaehler { return &lesezaehler{} }},
		{"nicht leer", func() *lesezaehler { return &lesezaehler{daten: []byte(`{"a":1}`)} }},
		{"nicht lesbar", func() *lesezaehler { return &lesezaehler{fehler: errors.New("nicht lesbar")} }},
	}
	n := 0
	for _, strat := range vierStrategien() {
		for _, laenge := range laengen {
			for _, b := range bodies {
				n++
				was := fmt.Sprintf("%s Laenge %d Body %s", strat, laenge, b.name)
				u := neu(t)
				u.k.Strategie = strat
				u.k.MaxBytes = grenze
				a := anf()
				a.Methode = "POST"
				a.Header["Content-Type"] = []string{"application/json"}
				a.Laenge = laenge
				var body *lesezaehler
				if b.neu != nil {
					body = b.neu()
					a.Body = body
				}
				u.lege(u.austausch(1, u.keyMitBody(a, nil, model.RedactionNone), true))
				resp, err := u.replay(context.Background(), a)
				liest := b.neu != nil && laenge != 0
				switch {
				case b.neu == nil && laenge > 0:
					erwarteOhneCode(t, was, err, model.ErrUngueltig)
					if u.p.gesamt() != 0 {
						t.Errorf("%s: %d Aufrufe eines Ports nach dem Programmfehler", was, u.p.gesamt())
					}
				case !liest || b.name == "leer":
					if err != nil {
						t.Errorf("%s: %v, erwartet den Key ohne Body", was, err)
						break
					}
					_ = resp.Body.Close()
				default:
					erwarteCode(t, was, err, model.CodeE5001)
					if b.name == "nicht lesbar" && u.p.anzahl("Exchanges.ListBySession") != 0 {
						t.Errorf("%s: Port nach dem Lesefehler gerufen", was)
					}
					if _, da := u.cursorStand(); da {
						t.Errorf("%s: Cursor nach dem Fehler", was)
					}
				}
				switch {
				case body == nil:
				case !liest && body.aufrufe != 0:
					t.Errorf("%s: %d Read-Aufrufe, %d Bytes gelesen; erwartet keiner", was, body.aufrufe, body.gelesen)
				case liest && body.aufrufe == 0:
					t.Errorf("%s: Body nicht gelesen", was)
				case liest && b.name == "nicht leer" && body.gelesen != 7:
					t.Errorf("%s: %d Bytes gelesen, erwartet 7", was, body.gelesen)
				}
			}
		}
	}
	if n != 96 {
		t.Errorf("Menge: %d, erwartet 96", n)
	}
}

package outbound_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestAbbildungEingabe (Zeile 44): beide Sentinel haben eine Zeile; `ErrEingabeNichtUnterstuetzt` ergibt ODR-E6002, `ErrEingabeUngueltig` ODR-E6007, bei
// jedem der acht Gegenstände, auch verpackt und verbunden; ein Kontextfehler daneben bekommt keinen Code. Menge 2 Sentinel × 8 Gegenstände, dazu
// 19 Ursachen × 8.
func TestAbbildungEingabe(t *testing.T) {
	n := 0
	for _, s := range []struct {
		err  error
		soll model.Code
	}{{model.ErrEingabeNichtUnterstuetzt, model.CodeE6002}, {model.ErrEingabeUngueltig, model.CodeE6007}} {
		for _, g := range alleGegenstaende() {
			if got, ok := outbound.Abbilden(s.err, g); !ok || got != s.soll {
				t.Errorf("%v / Gegenstand %d: %v %v, erwartet %v", s.err, g, got, ok, s.soll)
			}
			n++
		}
	}
	if n != 16 {
		t.Errorf("Menge %d, erwartet 16", n)
	}
	for _, u := range model.AlleEingabeUrsachen() {
		soll := model.CodeE6007
		if u.NichtUnterstuetzt() {
			soll = model.CodeE6002
		}
		e := &model.EingabeFehler{Ursache: u, Ort: "/x"}
		formen := map[string]error{"Fehler": e, "Errorf": fmt.Errorf("lesen: %w", e), "Join": errors.Join(errors.New("x"), e)}
		for name, err := range formen {
			for _, g := range alleGegenstaende() {
				if got, ok := outbound.Abbilden(err, g); !ok || got != soll {
					t.Errorf("%s %s / Gegenstand %d: %v %v, erwartet %v", u, name, g, got, ok, soll)
				}
			}
		}
		if _, ok := outbound.Abbilden(errors.Join(e, context.Canceled), gSonst); ok {
			t.Error("ein Kontextfehler neben dem Fehler bekommt einen Code")
		}
	}
	for _, name := range []string{"ErrEingabeUngueltig", "ErrEingabeNichtUnterstuetzt"} {
		if !slices.Contains(outbound.EintraegeAbbildung(), name) {
			t.Errorf("die Tabelle nennt %s nicht", name)
		}
	}
}

type eingabeFake struct{}

func (eingabeFake) LeseEntity(context.Context, io.Reader, []string, model.ServiceModel, string, model.EingabeArt, int64) (model.EntityEingabe, error) {
	return model.EntityEingabe{}, nil
}

// TestCodecSatzEingabe (Zeile 41): der Satz trägt den Leser im Feld `Eingabe` des Typs `ODataEingabe`; der Nullwert des Feldes ist nil.
func TestCodecSatzEingabe(t *testing.T) {
	f, ok := reflect.TypeFor[outbound.CodecSatz]().FieldByName("Eingabe")
	if !ok || f.Type != reflect.TypeFor[outbound.ODataEingabe]() {
		t.Fatalf("Feld Eingabe: %v %v", f, ok)
	}
	s := outbound.CodecSatz{Eingabe: eingabeFake{}}
	if s.Eingabe == nil || (outbound.CodecSatz{}).Eingabe != nil {
		t.Error("Feld Eingabe trägt den Leser nicht oder der Nullwert ist nicht nil")
	}
}

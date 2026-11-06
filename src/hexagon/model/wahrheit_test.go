package model_test

import (
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	wF = model.WahrheitFalsch
	wW = model.WahrheitWahr
	wU = model.WahrheitUnbekannt
)

// tafelEingaben sind die drei Konstanten und zwei Zahlenwerte außerhalb: sie gelten als unbekannt (SPEC-166 Nr. 5).
func tafelEingaben() []model.Wahrheit {
	return []model.Wahrheit{wF, wW, wU, model.Wahrheit(3), model.Wahrheit(-7)}
}

// wahrKanonisch bildet einen Wert außerhalb der drei Konstanten auf unbekannt ab; die Tafeln unten tragen nur die drei.
func wahrKanonisch(w model.Wahrheit) model.Wahrheit {
	if w == wF || w == wW {
		return w
	}
	return wU
}

// Die Tafeln stehen ausgeschrieben: Zeile = linker Operand, Spalte = rechter, Reihenfolge falsch, wahr, unbekannt.
var (
	tafelUnd = map[model.Wahrheit][3]model.Wahrheit{
		wF: {wF, wF, wF},
		wW: {wF, wW, wU},
		wU: {wF, wU, wU},
	}
	tafelOder = map[model.Wahrheit][3]model.Wahrheit{
		wF: {wF, wW, wU},
		wW: {wW, wW, wW},
		wU: {wU, wW, wU},
	}
	tafelNicht = map[model.Wahrheit]model.Wahrheit{wF: wW, wW: wF, wU: wU}
)

func spalte(w model.Wahrheit) int {
	switch w {
	case wF:
		return 0
	case wW:
		return 1
	}
	return 2
}

func pruefeTafel(t *testing.T, name string, f func(a, b model.Wahrheit) model.Wahrheit, tafel map[model.Wahrheit][3]model.Wahrheit) {
	t.Helper()
	n := 0
	for _, a := range tafelEingaben() {
		for _, b := range tafelEingaben() {
			soll := tafel[wahrKanonisch(a)][spalte(wahrKanonisch(b))]
			if got := f(a, b); got != soll {
				t.Errorf("%s(%d, %d) = %d, erwartet %d", name, a, b, got, soll)
			}
			n++
		}
	}
	if n != 25 {
		t.Errorf("%d Paare statt 25", n)
	}
}

// TestUndTafel (Zeile 12): alle Paare aus falsch, wahr, unbekannt und zwei Werten außerhalb der Konstanten in beiden
// Reihenfolgen gegen die ausgeschriebene Tafel.
func TestUndTafel(t *testing.T) { pruefeTafel(t, "Und", model.Und, tafelUnd) }

// TestOderTafel (Zeile 13): wie TestUndTafel für Oder.
func TestOderTafel(t *testing.T) { pruefeTafel(t, "Oder", model.Oder, tafelOder) }

// TestNichtTafel (Zeile 14): alle drei Werte und zwei außerhalb der Konstanten; das Ergebnis ist eine der drei Konstanten.
func TestNichtTafel(t *testing.T) {
	n := 0
	for _, a := range tafelEingaben() {
		if got, soll := model.Nicht(a), tafelNicht[wahrKanonisch(a)]; got != soll {
			t.Errorf("Nicht(%d) = %d, erwartet %d", a, got, soll)
		}
		n++
	}
	if n != 5 {
		t.Errorf("%d Eingaben statt 5", n)
	}
}

// TestWahrheitVonUndDurchlaesst (Zeile 15): WahrheitVon über alle zehn Arten und vier Texte gegen die Erwartung des
// Tests, Durchlaesst nur für wahr.
func TestWahrheitVonUndDurchlaesst(t *testing.T) {
	arten := []model.WertArt{model.ArtNull, model.ArtString, model.ArtBoolean, model.ArtInt32, model.ArtInt64, model.ArtDecimal,
		model.ArtGUID, model.ArtDate, model.ArtDateTimeOffset, model.ArtRedigiert}
	n := 0
	for _, art := range arten {
		for _, text := range []string{"true", "false", "", "x"} {
			got, err := model.WahrheitVon(model.Wert{Art: art, Text: text})
			var soll model.Wahrheit
			fehler := true
			switch {
			case art == model.ArtNull:
				soll, fehler = wU, false
			case art == model.ArtBoolean && text == "true":
				soll, fehler = wW, false
			case art == model.ArtBoolean && text == "false":
				soll, fehler = wF, false
			}
			if fehler {
				var ft *model.FilterTypFehler
				if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &ft) || got != wU {
					t.Errorf("%s %q: %d, %v", art, text, got, err)
				}
			} else if err != nil || got != soll {
				t.Errorf("%s %q: %d, %v, erwartet %d", art, text, got, err, soll)
			}
			n++
		}
	}
	if n != 40 {
		t.Errorf("%d Fälle statt 40", n)
	}
	for _, art := range []model.WertArt{"", "bogus"} {
		if _, err := model.WahrheitVon(model.Wert{Art: art, Text: "true"}); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("Art %q: %v", art, err)
		}
	}
	for _, w := range tafelEingaben() {
		if got := w.Durchlaesst(); got != (w == wW) {
			t.Errorf("Durchlaesst(%d) = %v", w, got)
		}
	}
}

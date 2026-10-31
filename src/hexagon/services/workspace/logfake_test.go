package workspace_test

import (
	"context"
	"log/slog"
	"reflect"
	"testing"
)

// TestLogHandlerLeitetAb (Z37, Z64): der Fake des Protokolls trägt jede Ableitung eines *slog.Logger mit; die
// Menge sind die Methoden von *slog.Logger, die einen Logger liefern (With, WithGroup). Ein Dienst, der den Logger
// ableitet, ändert die aufgezeichneten Namen oder fügt Attribute hinzu, und der Fake zeigt das.
func TestLogHandlerLeitetAb(t *testing.T) {
	typ := reflect.TypeOf(&slog.Logger{})
	abgeleitet := 0
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if m.Type.NumOut() != 1 || m.Type.Out(0) != typ {
			continue
		}
		abgeleitet++
		lh := neuerLogHandler()
		basis := reflect.ValueOf(slog.New(lh))
		var soll map[string]string
		var l *slog.Logger
		switch m.Name {
		case "With":
			l = basis.MethodByName("With").Call([]reflect.Value{reflect.ValueOf("vorab"), reflect.ValueOf("v")})[0].Interface().(*slog.Logger)
			soll = map[string]string{"vorab": "v", "ereignis": "e"}
		case "WithGroup":
			l = basis.MethodByName("WithGroup").Call([]reflect.Value{reflect.ValueOf("g")})[0].Interface().(*slog.Logger)
			soll = map[string]string{"g.ereignis": "e"}
		default:
			t.Errorf("Ableitung %s ist im Test nicht behandelt", m.Name)
			continue
		}
		l.LogAttrs(context.Background(), slog.LevelInfo, "text", slog.String("ereignis", "e"))
		got := lh.eintraege()
		if len(got) != 1 || !mapsGleich(got[0].attr, soll) {
			t.Errorf("%s: aufgezeichnet %+v, erwartet %v", m.Name, got, soll)
		}
	}
	if abgeleitet != 2 {
		t.Errorf("%d ableitende Methoden, erwartet 2", abgeleitet)
	}
	// Eine Gruppe nach einem Attribut qualifiziert nur die späteren Namen.
	lh := neuerLogHandler()
	slog.New(lh).With("a", 1).WithGroup("g").LogAttrs(context.Background(), slog.LevelInfo, "t", slog.Int("b", 2))
	if got := lh.eintraege(); len(got) != 1 || !mapsGleich(got[0].attr, map[string]string{"a": "1", "g.b": "2"}) {
		t.Errorf("With, dann WithGroup: %+v", got)
	}
}

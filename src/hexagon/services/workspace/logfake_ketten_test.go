package workspace_test

import (
	"context"
	"log/slog"
	"testing"
)

// TestLogHandlerKetten (Z37, Z64): der Fake des Protokolls qualifiziert die Namen wie ein Handler des Originals, wenn
// die Ableitungen aufeinander folgen: Attribute vor und nach einer Gruppe, verschachtelte Gruppen, eine leere Gruppe
// und eine Gruppe mit Attributen im Eintrag selbst. Die Menge ist die Folge aus With und WithGroup bis zur Länge
// drei über zwei Namen: 2 + 4 + 8 = 14 Ketten; der Orakel-Handler rechnet die Namen aus der Folge, unabhängig vom Fake.
func TestLogHandlerKetten(t *testing.T) {
	schritte := []string{"With", "WithGroup"}
	var ketten [][]string
	var bau func(k []string)
	bau = func(k []string) {
		if len(k) > 0 {
			ketten = append(ketten, append([]string{}, k...))
		}
		if len(k) == 3 {
			return
		}
		for _, s := range schritte {
			bau(append(k, s))
		}
	}
	bau(nil)
	if len(ketten) != 14 {
		t.Fatalf("%d Ketten, erwartet 14", len(ketten))
	}
	for _, kette := range ketten {
		lh := neuerLogHandler()
		l := slog.New(lh)
		praefix := ""
		soll := map[string]string{}
		for i, s := range kette {
			name := string(rune('a' + i))
			if s == "With" {
				l = l.With("w"+name, name)
				soll[praefix+"w"+name] = name
			} else {
				l = l.WithGroup("g" + name)
				praefix += "g" + name + "."
			}
		}
		l.LogAttrs(context.Background(), slog.LevelInfo, "t", slog.String("ereignis", "e"))
		soll[praefix+"ereignis"] = "e"
		got := lh.eintraege()
		if len(got) != 1 || !mapsGleich(got[0].attr, soll) {
			t.Errorf("Kette %v: aufgezeichnet %+v, erwartet %v", kette, got, soll)
		}
	}
	// Eine leere Gruppe ändert die Namen nicht.
	lh := neuerLogHandler()
	slog.New(lh).WithGroup("").LogAttrs(context.Background(), slog.LevelInfo, "t", slog.String("k", "v"))
	if got := lh.eintraege(); len(got) != 1 || !mapsGleich(got[0].attr, map[string]string{"k": "v"}) {
		t.Errorf("leere Gruppe: %+v", got)
	}
}

package telemetry_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

// textFehler, textStringer, panikText und lvText tragen einen Text, den ein Aufrufer in einen error, einen
// Stringer, die Panik eines LogValue oder das Ergebnis eines LogValue legen könnte.
type textFehler string

func (e textFehler) Error() string { return string(e) }

type textStringer string

func (s textStringer) String() string { return string(s) }

type panikText string

func (p panikText) LogValue() slog.Value { panic(string(p)) }

type lvText string

func (l lvText) LogValue() slog.Value { return slog.StringValue(string(l)) }

// H09: ein Canary erreicht über den Handler keine Zeile, keinen Stand und keinen Fehlerwert. Menge
// M15: 35 Stellen (13 Felder über den Aufruf, 13 über WithAttrs, 8 weitere Stellen, der Fehlertext der
// Senke) × 4 Wege (Header, Body, Query-Literal, Konfigurationswert) × 2 Ausgaben (Bytes der Senke, %+v
// des Stands) = 280 Prüfungen: weder das Canary noch sein Präfix von 4 Bytes. Die Kontrollen zeigen,
// dass derselbe Wert in einem gültigen Feld die Prüfung rot färbt und dass das Präfix in der Zeile
// gültiger Werte nicht steht.
func TestCanaryHandler(t *testing.T) {
	const canary = "kZq7Xw3VbN"
	wege := []struct{ name, wert string }{
		{"Header", "Bearer " + canary},
		{"Body", `{"password":"` + canary + `"}`},
		{"Query", "$filter=Name eq '" + canary + "'"},
		{"Konfiguration", "https://user:" + canary + "@host"},
	}
	felder := logSchluessel(t)
	var stellen []string
	for _, k := range felder {
		stellen = append(stellen, "Aufruf "+k)
	}
	for _, k := range felder {
		stellen = append(stellen, "WithAttrs "+k)
	}
	weitere := []string{
		"Text eines error", "Text eines Stringer", "Name einer Gruppe", "Name von WithGroup", "fremder Schlüssel",
		"Wert eines Elements einer Gruppe", "Text der Panik eines LogValue", "LogValuer mit dem Wert als String",
	}
	stellen = append(append(stellen, weitere...), "Fehlertext der Senke")
	if len(felder) != 13 || len(wege) != 4 || len(stellen) != 35 {
		t.Fatalf("%d Felder, %d Wege, %d Stellen; erwartet 13, 4 und 35", len(felder), len(wege), len(stellen))
	}
	enthaelt := func(b []byte) bool { return bytes.Contains(b, []byte(canary[:4])) }
	pruefungen := 0
	for _, stelle := range stellen {
		for _, weg := range wege {
			senke := &FakeSenke{}
			a := neuMit(t, senke, festeUhr(), slog.LevelInfo)
			h := a.Handler()
			var attr slog.Attr
			switch art, key, _ := strings.Cut(stelle, " "); {
			case art == "Aufruf" || art == "WithAttrs":
				attr = slog.String(key, weg.wert)
			case stelle == "Text eines error":
				attr = slog.Any("mode", textFehler(weg.wert))
			case stelle == "Text eines Stringer":
				attr = slog.Any("mode", textStringer(weg.wert))
			case stelle == "Name einer Gruppe":
				attr = slog.Group(weg.wert, slog.String("mode", "record"))
			case stelle == "Name von WithGroup":
				h, attr = h.WithGroup(weg.wert), slog.String("service_id", "svc-9")
			case stelle == "fremder Schlüssel":
				attr = slog.String(weg.wert, "x")
			case stelle == "Wert eines Elements einer Gruppe":
				attr = slog.Group("mode", slog.String("mode", weg.wert))
			case stelle == "Text der Panik eines LogValue":
				attr = slog.Any("mode", panikText(weg.wert))
			case stelle == "LogValuer mit dem Wert als String":
				attr = slog.Any("mode", lvText(weg.wert))
			default:
				attr = slog.String("session_id", "s1")
				senke.antwort = func([]byte) (int, error) { return 0, errors.New("Schreiben: " + weg.wert) }
			}
			if strings.HasPrefix(stelle, "WithAttrs ") {
				h = h.WithAttrs([]slog.Attr{attr})
				attr = slog.Attr{}
			}
			var aufruf []slog.Attr
			if attr.Key != "" {
				aufruf = []slog.Attr{attr}
			}
			if err := h.Handle(context.Background(), rec(slog.LevelInfo, "m", aufruf...)); err != nil {
				t.Errorf("Stelle %s, Weg %s: Handle liefert einen Fehlerwert %v", stelle, weg.name, err)
			}
			ausgaben := map[string][]byte{"Senke": senke.bytesAlle(), "Stand": fmt.Appendf(nil, "%+v", a.Snapshot())}
			for name, o := range ausgaben {
				pruefungen++
				if enthaelt(o) {
					t.Errorf("Stelle %s, Weg %s, Ausgabe %s: das Canary oder sein Präfix steht in der Ausgabe", stelle, weg.name, name)
				}
			}
		}
	}
	if pruefungen != 35*4*2 {
		t.Fatalf("%d Prüfungen, erwartet %d", pruefungen, 35*4*2)
	}
	senke := &FakeSenke{}
	_ = neuMit(t, senke, festeUhr(), slog.LevelInfo).Handler().Handle(context.Background(), rec(slog.LevelInfo, "m", s("session_id", canary)))
	if !enthaelt(senke.bytesAlle()) {
		t.Error("Kontrolle: ein Canary in einem gültigen Feld steht in der Zeile; die Prüfung sieht es nicht")
	}
	senke = &FakeSenke{}
	var gueltig []slog.Attr
	for _, k := range felder {
		gueltig = append(gueltig, slogAttr(gueltigFuer(k)))
	}
	_ = neuMit(t, senke, festeUhr(), slog.LevelInfo).Handler().Handle(context.Background(), rec(slog.LevelInfo, "m", gueltig...))
	if len(lies(t, senke.bytesAlle()).schluessel) != 3+len(felder) || enthaelt(senke.bytesAlle()) {
		t.Error("Kontrolle: das Präfix steht in der Zeile eines Aufrufs mit gültigen Werten, oder die Zeile ist unvollständig")
	}
}

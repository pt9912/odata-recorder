package telemetry_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// S22 (M11): ein Canary erreicht über keinen Weg der Spans eine Zeile, ein Attribut oder eine Metrik.
// 4 Einbringungswege (Header, Body-Fragment, Query-Literal, Konfigurationswert) × 49 Stellen (24
// Schlüssel an 2 Aufrufstellen und der Name des Spans) × 5 Ausgabewege = 980 Prüfungen. Schlüssel: 13
// von Log (Schlüssel i am Span i reihum) und 11 Pflicht-Attribute (an ihrem Span). Ausgabewege: Bytes der
// Senke aus dem End, aus einem Log danach, aus einem Eintrag über Handler() danach, %+v von Spans samt Attrs
// und %+v von Errors, Durations und Dropped. Die Eingabe über den Handler ist H09.
func TestCanarySpans(t *testing.T) {
	const token = "ZQ9K"
	tab := spanTabelle(t)
	wege := []struct{ name, wert string }{
		{"Header", "Bearer ZQ9K-header"},
		{"Body", `{"k":"ZQ9K-body"}`},
		{"Query", "$filter=Name eq 'ZQ9K-query'"},
		{"Konfiguration", "https://user:ZQ9K-konfig@host"},
	}
	type stelle struct {
		key     string
		span    int
		aufruf  string
		istName bool
	}
	var stellen []stelle
	for i, k := range logSchluessel(t) {
		for _, aufruf := range []string{"StartSpan", "Set"} {
			stellen = append(stellen, stelle{k, i % len(tab), aufruf, false})
		}
	}
	pflichtZahl := 0
	for i, z := range tab {
		for _, k := range z.pflicht {
			pflichtZahl++
			for _, aufruf := range []string{"StartSpan", "Set"} {
				stellen = append(stellen, stelle{k, i, aufruf, false})
			}
		}
	}
	stellen = append(stellen, stelle{istName: true, aufruf: "StartSpan"})
	if len(stellen) != 2*(13+pflichtZahl)+1 || len(stellen) != 49 || pflichtZahl != 11 {
		t.Fatalf("%d Stellen, %d Pflicht-Attribute; erwartet 49 und 11", len(stellen), pflichtZahl)
	}
	leckt := func(out []byte, wert string) bool {
		return bytes.Contains(out, []byte(wert)) || bytes.Contains(out, []byte(token))
	}
	pruefungen := 0
	for _, w := range wege {
		for _, st := range stellen {
			u := neuUmgebung(t, slog.LevelInfo, 0)
			ctx := context.Background()
			name := support.SpanName(tab[st.span].name)
			switch {
			case st.istName:
				_, sp := u.a.StartSpan(ctx, support.SpanName(w.wert))
				sp.Set()
				sp.End()
			case st.aufruf == "StartSpan":
				_, sp := u.a.StartSpan(ctx, name, support.Str(st.key, w.wert))
				sp.End()
			default:
				_, sp := u.a.StartSpan(ctx, name)
				sp.Set(support.Str(st.key, w.wert))
				sp.End()
			}
			aus1 := u.senke.bytesAlle()
			u.a.Log(ctx, slog.LevelInfo, "m", support.Str("mode", "record"))
			aus2 := u.senke.bytesAlle()[len(aus1):]
			slog.New(u.a.Handler()).Info("m", "mode", "record")
			aus3 := u.senke.bytesAlle()[len(aus1)+len(aus2):]
			s := u.a.Snapshot()
			aus4 := []byte(fmt.Sprintf("%+v", s.Spans))
			aus5 := []byte(fmt.Sprintf("%+v %+v %+v", s.Errors, s.Durations, s.Dropped))
			for i, aus := range [][]byte{aus1, aus2, aus3, aus4, aus5} {
				pruefungen++
				if leckt(aus, w.wert) {
					t.Errorf("Weg %s, Stelle %s %s %q, Ausgabeweg %d: der Canary steht in %q", w.name, st.aufruf, name, st.key, i+1, aus)
				}
			}
		}
	}
	if pruefungen != len(wege)*49*5 || pruefungen != 980 {
		t.Errorf("%d Prüfungen, erwartet 980", pruefungen)
	}
	// Kontrolle (a): ein gültiger Wert eines Pflicht-Attributs steht in der Span-Zeile und in Attrs
	for _, z := range tab {
		for _, k := range z.pflicht {
			u := neuUmgebung(t, slog.LevelInfo, 0)
			at := gueltigSpan(k)
			startEnde(u.a, z.name, at)
			s := u.a.Snapshot()
			wert := erwarteterText(at)
			if !bytes.Contains(u.senke.bytesAlle(), []byte(wert)) || !bytes.Contains([]byte(fmt.Sprintf("%+v", s.Spans)), []byte(wert)) {
				t.Errorf("Kontrolle (a): %s = %s steht nicht in Zeile und Attrs", k, wert)
			}
		}
	}
	// Kontrolle (b): ein Canary als msg von Log findet der Leser der Senke (er ist nicht blind)
	for _, w := range wege {
		u := neuUmgebung(t, slog.LevelInfo, 0)
		u.a.Log(context.Background(), slog.LevelInfo, w.wert)
		if !leckt(u.senke.bytesAlle(), w.wert) {
			t.Errorf("Kontrolle (b): der Leser findet %s als msg nicht", w.name)
		}
	}
}

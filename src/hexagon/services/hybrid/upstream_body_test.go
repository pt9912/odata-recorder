package hybrid_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestBodyNieGeschlossenUpstream hält Z16 für den Weg U (SPEC-284 Nr. 5 d, 10): der Dienst schließt den Body der Anfrage
// nie. Menge aus dem Wortlaut der Zusage: jeder Weg zu U (nach M, nach R) × jede Form des Bodys (gepuffert mit Laenge
// -1 und 4, Laenge 0 mit Leser, über der Grenze, Lesefehler) × jeder Ausgang (Erfolg, jeder Fehler der Tabelle von
// Senden, Kontextfehler von Senden, Abbruch vor U). Der Body zählt jedes Close; wo ein Fallback möglich ist und kein Abbruch
// vorliegt, läuft U genau einmal, sonst nie.
func TestBodyNieGeschlossenUpstream(t *testing.T) {
	type bodyForm struct {
		name   string
		neu    func() *strom
		laenge int64
		faehig bool
	}
	formen := []bodyForm{
		{"gepuffert, Laenge -1", func() *strom { return neuerStrom(muster(4, 2)) }, -1, true},
		{"gepuffert, Laenge 4", func() *strom { return neuerStrom(muster(4, 2)) }, 4, true},
		{"Laenge 0 mit Leser", func() *strom { return neuerStrom(muster(4, 2)) }, 0, true},
		{"über der Grenze", func() *strom { return neuerStrom(muster(standardMax+3, 2)) }, standardMax + 3, false},
		{"Lesefehler", func() *strom { return &strom{r: &fehlerNach{rest: muster(2, 2), err: errors.New("lesefehler")}} }, -1, false},
	}
	type ausgang struct {
		name    string
		senden  func(g *rig)
		abbruch bool
	}
	ausgaenge := []ausgang{{name: "Erfolg", senden: func(*rig) {}}}
	for _, f := range fehlerTabelle() {
		ausgaenge = append(ausgaenge, ausgang{name: "Fehler " + f.name, senden: func(g *rig) {
			g.u.liefert(outbound.UpstreamAntwort{Status: 200, Body: neuerStrom(nil)}, f.err)
		}})
	}
	ausgaenge = append(ausgaenge,
		ausgang{name: "Kontextfehler von Senden", senden: func(g *rig) { g.u.liefert(outbound.UpstreamAntwort{}, context.Canceled) }},
		ausgang{name: "Abbruch vor U", senden: func(*rig) {}, abbruch: true},
	)
	n := 0
	for weg, neu := range wege(t) {
		for _, form := range formen {
			for _, aus := range ausgaenge {
				n++
				g := neu()
				ctx, cancel := context.WithCancel(context.Background())
				if aus.abbruch {
					stufe := &g.m.fn
					fehler := error(code(codeNamed(t, mockCode)))
					if weg == "nach R" {
						stufe = &g.r.fn
						fehler = code(codeNamed(t, missCode))
					}
					*stufe = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
						cancel()
						return inbound.Antwort{}, fehler
					}
				}
				aus.senden(g)
				s := form.neu()
				_, _ = g.v.Hybrid(ctx, mitBody("GET", s, form.laenge))
				cancel()
				soll := 0
				if form.faehig && !aus.abbruch {
					soll = 1
				}
				if got := g.u.zahl(); got != soll || s.closes.Load() != 0 {
					t.Errorf("%s, %s, %s: U %d-mal (erwartet %d), Body %d-mal geschlossen", weg, form.name, aus.name, got, soll, s.closes.Load())
				}
			}
		}
	}
	if n != 2*5*(1+len(fehlerTabelle())+2) {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

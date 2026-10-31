package replay_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestAbbildungVerbuende hält Z6 für den Verbund mit ErrNichtGefunden: Menge 588 (7 Stellen × 21 zweite Fehler × 2
// Reihenfolgen × 2 Verpackungen). Die Tabelle von SPEC-109 Nr. 5 bildet den Verbund nach dem ersten zutreffenden
// Eintrag ab: steht der zweite Fehler vor der Zeile von ErrNichtGefunden (13 Fehler), gilt sein Code, steht er
// danach (7 Fehler und ein unbekannter), gilt ErrNichtGefunden mit dem Gegenstand der Stelle; an S6 ist das der
// fehlende Cursor und damit Erfolg. Die Handtabelle folgt dem Wortlaut von SPEC-109 Nr. 5, nicht dem Code.
func TestAbbildungVerbuende(t *testing.T) {
	ng := outbound.ErrNichtGefunden
	type zweiter struct {
		name string
		err  error
		vor  model.Code // Code, wenn die Zeile vor ErrNichtGefunden steht; sonst 0
	}
	zweite := []zweiter{
		{"ErrBlobFehlt", outbound.ErrBlobFehlt, model.CodeE3001}, {"ErrBlobBeschaedigt", outbound.ErrBlobBeschaedigt, model.CodeE3001},
		{"ErrBeschaedigt", outbound.ErrBeschaedigt, model.CodeE3001}, {"ErrPfad", outbound.ErrPfad, model.CodeE2001},
		{"ErrZuGross", outbound.ErrZuGross, model.CodeE2401}, {"ErrUpstreamNichtErreichbar", outbound.ErrUpstreamNichtErreichbar, model.CodeE4001},
		{"ErrUpstreamZeit", outbound.ErrUpstreamZeit, model.CodeE4001}, {"ErrFremderService", outbound.ErrFremderService, model.CodeE5004},
		{"ErrDienstUnbekannt", outbound.ErrDienstUnbekannt, model.CodeE5005}, {"ErrFalscheArt", outbound.ErrFalscheArt, model.CodeE7008},
		{"ErrReferenziert", outbound.ErrReferenziert, model.CodeE7004}, {"ErrSessionGeschlossen", outbound.ErrSessionGeschlossen, model.CodeE5009},
		{"ErrUngueltigerUebergang", outbound.ErrUngueltigerUebergang, model.CodeE7011},
		{"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden, 0}, {"model.ErrUngueltig", model.ErrUngueltig, 0},
		{"ErrMetadatenUngueltig", model.ErrMetadatenUngueltig, 0}, {"ErrDatenUngueltig", model.ErrDatenUngueltig, 0},
		{"ErrUngueltigeSequenz", outbound.ErrUngueltigeSequenz, 0}, {"ErrAbgeschlossen", outbound.ErrAbgeschlossen, 0},
		{"ErrRevisionVeraltet", outbound.ErrRevisionVeraltet, 0}, {"anderer", errors.New("anderer"), 0},
	}
	if len(zweite) != 21 {
		t.Fatalf("%d zweite Fehler, erwartet 21", len(zweite))
	}
	// Code von ErrNichtGefunden je Stelle; 0 ist der fehlende Cursor (Erfolg).
	ngCode := map[string]model.Code{
		"S1": model.CodeE7003, "S3": model.CodeE7003, "S4": model.CodeE9001, "S5": model.CodeE7001,
		"S6": 0, "S7": model.CodeE9001, "S8": model.CodeE7003,
	}
	n := 0
	for _, s := range stellen() {
		for _, z := range zweite {
			for _, ngZuerst := range []bool{true, false} {
				for _, verpackt := range []bool{false, true} {
					n++
					var err error
					if ngZuerst {
						err = errors.Join(ng, z.err)
					} else {
						err = errors.Join(z.err, ng)
					}
					if verpackt {
						err = fmt.Errorf("verpackt: %w", err)
					}
					was := fmt.Sprintf("%s Join(ng zuerst: %v, %s) verpackt: %v", s.name, ngZuerst, z.name, verpackt)
					u := neu(t)
					a := erfolgsLauf(u)
					u.p.setzeFehler(s.op, err)
					resp, got := u.replay(context.Background(), a)
					want := z.vor
					if want == 0 {
						want = ngCode[s.name]
					}
					if want == 0 {
						if got != nil {
							t.Errorf("%s: Fehler %v, erwartet Erfolg", was, got)
						} else {
							_ = resp.Body.Close()
						}
						continue
					}
					erwarteCode(t, was, got, want)
				}
			}
		}
	}
	if n != 588 {
		t.Errorf("Menge: %d Elemente, erwartet 588", n)
	}
}

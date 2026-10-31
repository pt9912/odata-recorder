package workspace_test

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// eintragsKopie hält fest, was ein Logeintrag trägt.
type eintragsKopie struct {
	stufe slog.Level
	text  string
	attr  map[string]string
}

// TestProtokollNachDemLoeschen (Z37): je Ereignis ein Eintrag mit festem Text, Stufe und genau den Attributen
// ereignis, workspace_id, error_code; kein Eintrag bei Kontextfehler und bei Erfolg; ein Fehler ohne *Fehler
// ergibt ODR-E9001; kein Fehlertext.
func TestProtokollNachDemLoeschen(t *testing.T) {
	const fehlgeschlagen = "workspace: Bereinigung nach dem Löschen fehlgeschlagen"
	const zurueckgestellt = "workspace: Bereinigung nach dem Löschen zurückgestellt"
	faelle := []struct {
		name    string
		bericht inbound.Loeschbericht
		err     error
		soll    *eintragsKopie
	}{
		{"Erfolg", inbound.Loeschbericht{Entfernt: 3}, nil, nil},
		{"Zurückgestellt", inbound.Loeschbericht{Zurueckgestellt: true}, nil,
			&eintragsKopie{slog.LevelInfo, zurueckgestellt, map[string]string{"ereignis": "bereinigung_zurueckgestellt", "workspace_id": "ws-1"}}},
		{"Fehler mit Code", inbound.Loeschbericht{}, &inbound.Fehler{Code: model.CodeE3001, Status: 500},
			&eintragsKopie{slog.LevelError, fehlgeschlagen, map[string]string{"ereignis": "bereinigung_fehlgeschlagen", "workspace_id": "ws-1", "error_code": "ODR-E3001"}}},
		{"Fehler mit Code eingepackt", inbound.Loeschbericht{Zurueckgestellt: true}, fmt.Errorf("außen: %w", &inbound.Fehler{Code: model.CodeE3001, Status: 500}),
			&eintragsKopie{slog.LevelError, fehlgeschlagen, map[string]string{"ereignis": "bereinigung_fehlgeschlagen", "workspace_id": "ws-1", "error_code": "ODR-E3001"}}},
		{"Fehler ohne Code", inbound.Loeschbericht{}, errPlatte,
			&eintragsKopie{slog.LevelError, fehlgeschlagen, map[string]string{"ereignis": "bereinigung_fehlgeschlagen", "workspace_id": "ws-1", "error_code": "ODR-E9001"}}},
		{"Canceled", inbound.Loeschbericht{}, context.Canceled, nil},
		{"DeadlineExceeded", inbound.Loeschbericht{Zurueckgestellt: true}, context.DeadlineExceeded, nil},
		{"Kontextfehler neben Fehler", inbound.Loeschbericht{}, joinMit(context.Canceled), nil},
	}
	for _, f := range faelle {
		w := neu(t)
		lh := neuerLogHandler()
		k := w.konfig
		k.Log = slog.New(lh)
		v, err := workspace.New(k)
		if err != nil {
			t.Fatal(err)
		}
		w.vorbelegen("ws-1", 0)
		w.wartung.setze(f.bericht, f.err)
		if _, err := v.Delete(context.Background(), "ws-1"); err != nil {
			t.Errorf("%s: %v", f.name, err)
		}
		got := lh.eintraege()
		switch {
		case f.soll == nil && len(got) != 0:
			t.Errorf("%s: Einträge %+v", f.name, got)
		case f.soll != nil && (len(got) != 1 || got[0].stufe != f.soll.stufe || got[0].text != f.soll.text || !mapsGleich(got[0].attr, f.soll.attr)):
			t.Errorf("%s: Einträge %+v, erwartet %+v", f.name, got, *f.soll)
		}
		for _, e := range got {
			for k, v := range e.attr {
				if strings.Contains(v, errPlatte.Error()) || strings.Contains(v, "Fehler:") {
					t.Errorf("%s: Fehlertext im Attribut %s=%q", f.name, k, v)
				}
			}
		}
	}
	if len(faelle) != 8 {
		t.Errorf("Fälle %d", len(faelle))
	}
}

func joinMit(ctxErr error) error {
	return fmt.Errorf("%w und %w", ctxErr, &inbound.Fehler{Code: model.CodeE3001, Status: 500})
}

func mapsGleich(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, da := b[k]; !da || w != v {
			return false
		}
	}
	return true
}

// TestAndereOperationenSchreibenNichts (Z37): Create, Get, List und Reset schreiben nichts ins Protokoll, im
// Erfolgsfall und im Fehlerfall.
func TestAndereOperationenSchreibenNichts(t *testing.T) {
	for _, o := range operationen() {
		if o.name == "Delete" {
			continue
		}
		for _, fehler := range []bool{false, true} {
			w := neu(t)
			lh := neuerLogHandler()
			k := w.konfig
			k.Log = slog.New(lh)
			v, err := workspace.New(k)
			if err != nil {
				t.Fatal(err)
			}
			w.v = v
			o.vor(w)
			if fehler {
				w.p.setzeFehler(stelleVon(o.name), errPlatte)
			}
			if got := o.lauf(w, context.Background()); (got != nil) != fehler {
				t.Errorf("%s fehler=%v: %v", o.name, fehler, got)
			}
			if e := lh.eintraege(); len(e) != 0 {
				t.Errorf("%s fehler=%v: Einträge %+v", o.name, fehler, e)
			}
		}
	}
}

// stelleVon nennt für eine Operation einen Aufruf, an dem ein Fehler sie scheitern lässt.
func stelleVon(op string) string {
	for _, s := range stellen() {
		if s.op == op {
			return s.label
		}
	}
	panic("keine Stelle für " + op)
}

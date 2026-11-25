package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

type ctxSchluessel struct{}

// spur zeichnet jeden Aufruf einer Stelle des Aufbaus auf (Name und Kontext) und lässt den Aufruf Nummer bei mit dem Kontextfehler enden: sie
// beendet den Kontext des Aufrufers und liefert den Fehler des Kontexts, den sie bekam.
type spur struct {
	mu      sync.Mutex
	namen   []string
	ctxs    []context.Context
	bei     int
	abbruch context.CancelFunc
}

func (s *spur) betrete(ctx context.Context, name string) error {
	s.mu.Lock()
	n := len(s.namen)
	s.namen = append(s.namen, name)
	s.ctxs = append(s.ctxs, ctx)
	s.mu.Unlock()
	if n != s.bei {
		return nil
	}
	s.abbruch()
	return ctx.Err()
}

type spurSessions struct {
	inbound.SessionUseCase
	s *spur
}

func (x spurSessions) List(ctx context.Context) ([]model.Session, error) {
	if err := x.s.betrete(ctx, "List"); err != nil {
		return nil, err
	}
	return x.SessionUseCase.List(ctx)
}

type spurImport struct {
	inbound.ImportUseCase
	s *spur
}

func (x spurImport) Import(ctx context.Context, a inbound.ImportAnfrage) (inbound.ImportErgebnis, error) {
	if err := x.s.betrete(ctx, "Import"); err != nil {
		return inbound.ImportErgebnis{}, err
	}
	return x.ImportUseCase.Import(ctx, a)
}

type spurWorkspaces struct {
	inbound.WorkspaceUseCase
	s *spur
}

func (x spurWorkspaces) Create(ctx context.Context, baselineID, id string) (model.Workspace, error) {
	if err := x.s.betrete(ctx, "Create"); err != nil {
		return model.Workspace{}, err
	}
	return x.WorkspaceUseCase.Create(ctx, baselineID, id)
}

// baue baut ein Handle über baueEcht, dessen Aufrufe mit Kontext (Hochfahren, Herunterfahren, List, Import, Create) die Spur führt; Herunterfahren
// fährt in jedem Fall herunter.
func (s *spur) baue(ctx context.Context, k bootstrap.Konfiguration) (*handle, error) {
	if err := s.betrete(ctx, "baue"); err != nil {
		return nil, err
	}
	h, err := baueEcht(ctx, k)
	if err != nil {
		return nil, err
	}
	hoch, herunter, sessions, imp, ws := h.Hochfahren, h.Herunterfahren, h.Sessions, h.Import, h.Workspaces
	h.Hochfahren = func(ctx context.Context) error {
		if err := s.betrete(ctx, "Hochfahren"); err != nil {
			return err
		}
		return hoch(ctx)
	}
	h.Herunterfahren = func(ctx context.Context) error { return errors.Join(s.betrete(ctx, "Herunterfahren"), herunter(ctx)) }
	h.Sessions = func() inbound.SessionUseCase { return spurSessions{sessions(), s} }
	h.Import = func() inbound.ImportUseCase { return spurImport{imp(), s} }
	h.Workspaces = func() inbound.WorkspaceUseCase { return spurWorkspaces{ws(), s} }
	return h, nil
}

// oeffne öffnet die Ablage über oeffneAblage, deren Lesezugriffe die Spur führt.
func (s *spur) oeffne(ctx context.Context, dir string) (*ablage, error) {
	if err := s.betrete(ctx, "Oeffne"); err != nil {
		return nil, err
	}
	a, err := oeffneAblage(ctx, dir)
	if err != nil {
		return nil, err
	}
	session, exchanges, coverage, entities := a.Session, a.Exchanges, a.Coverage, a.Entities
	a.Session = func(ctx context.Context, id string) (model.Session, error) {
		if err := s.betrete(ctx, "Session"); err != nil {
			return model.Session{}, err
		}
		return session(ctx, id)
	}
	a.Exchanges = func(ctx context.Context, id string) (int, error) {
		if err := s.betrete(ctx, "Exchanges"); err != nil {
			return 0, err
		}
		return exchanges(ctx, id)
	}
	a.Coverage = func(ctx context.Context, id string) ([]model.Coverage, error) {
		if err := s.betrete(ctx, "Coverage"); err != nil {
			return nil, err
		}
		return coverage(ctx, id)
	}
	a.Entities = func(ctx context.Context, ws, set string) (int, error) {
		if err := s.betrete(ctx, "Entities"); err != nil {
			return 0, err
		}
		return entities(ctx, ws, set)
	}
	return a, nil
}

// TestCtxJeAufrufstelle: jede der 14 Aufrufstellen des Aufbaus (Bauen, Hochfahren, Lesen der Sessions, Herunterfahren, zweites Bauen, Import,
// Anlegen des Workspace, zweites Herunterfahren, Öffnen der Ablage, Session, Exchanges, Coverage, Entities je Set) bekommt den Kontext des
// Aufrufs unverändert weitergereicht: er trägt den Wert des Aufrufers, keine Frist des Kontexts (die Frist des Aufbaus läuft über die Uhr) und
// ist der Kontext der Frist (nach dem Aufbau beendet, der des Aufrufers nicht); Herunterfahren bekommt ihn ohne den Abbruch
// (context.WithoutCancel, SPEC-216 Nr. 6 f): mit dem Wert, und auch nach dem Aufbau nicht beendet. Endet eine der 12 Stellen außer
// Herunterfahren mit dem Kontextfehler, ist der Aufbau ein Fehler mit dieser Ursache und ohne den Grund frist, keine spätere Stelle läuft, und
// jedes gebaute Handle ist genau einmal heruntergefahren; den Fehler von Herunterfahren führt TestAufbauHerunterfahrenFehler.
func TestCtxJeAufrufstelle(t *testing.T) {
	wache(t, 180*time.Second)
	erwartet := []string{
		"baue", "Hochfahren", "List", "Herunterfahren", "baue", "Import", "Create", "Herunterfahren",
		"Oeffne", "Session", "Exchanges", "Coverage", "Entities", "Entities",
	}
	fahre := func(bei int) (*spur, aufbauErgebnis, context.Context, error) {
		parent, abbruch := context.WithCancel(context.WithValue(context.Background(), ctxSchluessel{}, "wert"))
		t.Cleanup(abbruch)
		s := &spur{bei: bei, abbruch: abbruch}
		e := aufbauEingabe{Profil: kleinesProfil(), Startwert: 1, Uhr: neueFakeUhr(), DataDir: t.TempDir(), Baue: s.baue, Oeffne: s.oeffne}
		erg, err := e.aufbauen(parent)
		return s, erg, parent, err
	}
	s, _, parent, err := fahre(-1)
	if err != nil || fmt.Sprint(s.namen) != fmt.Sprint(erwartet) {
		t.Fatalf("Aufrufe %v, Fehler %v, erwartet %v ohne Fehler", s.namen, err, erwartet)
	}
	for i, ctx := range s.ctxs {
		_, hatFrist := ctx.Deadline()
		beendet := s.namen[i] != "Herunterfahren"
		if ctx.Value(ctxSchluessel{}) != "wert" || hatFrist || (ctx.Err() != nil) != beendet || parent.Err() != nil {
			t.Errorf("Stelle %d (%s): Wert %v, Frist des Kontexts %v, beendet nach dem Aufbau: %v, Kontext des Aufrufers beendet: %v; erwartet den Wert, keine Frist, beendet: %v, nicht beendet",
				i, s.namen[i], ctx.Value(ctxSchluessel{}), hatFrist, ctx.Err() != nil, parent.Err() != nil, beendet)
		}
	}
	for k := range erwartet {
		if erwartet[k] == "Herunterfahren" {
			continue
		}
		t.Run(fmt.Sprintf("%d %s", k, erwartet[k]), func(t *testing.T) {
			s, erg, _, err := fahre(k)
			var af *aufbauFehler
			if !errors.As(err, &af) || af.grund != grundAufbau || !errors.Is(err, context.Canceled) || erg != (aufbauErgebnis{}) {
				t.Fatalf("Fehler %v, Ergebnis %+v, erwartet die Ursache context.Canceled ohne den Grund frist und ohne Ergebnis", err, erg)
			}
			gebaut, herunter := 0, 0
			for i, n := range s.namen {
				if n == "baue" && i != k {
					gebaut++
				}
				if n == "Herunterfahren" {
					herunter++
				}
				if i > k && n != "Herunterfahren" {
					t.Errorf("Stelle %d (%s) läuft nach dem Kontextfehler an Stelle %d", i, n, k)
				}
			}
			if fmt.Sprint(s.namen[:k+1]) != fmt.Sprint(erwartet[:k+1]) || herunter != gebaut {
				t.Errorf("Aufrufe %v, Herunterfahren %d für %d gebaute Handles, erwartet je Handle eines und den Anfang %v", s.namen, herunter, gebaut, erwartet[:k+1])
			}
		})
	}
}

package mock_test

import (
	"context"
	"io"
	"maps"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// Die Fakes erfüllen die Ports zur Übersetzungszeit (AGENTS.md §3.15, Auftrag F); der Dienst erfüllt MockUseCase in mock.go.
var (
	_ outbound.WorkspaceTransactor = (*transaktion)(nil)
	_ outbound.ODataEingabe        = (*eingabeFake)(nil)
	_ outbound.WorkspaceRepository = (*workspaces)(nil)
	_ outbound.EntityRepository    = (*entities)(nil)
	_ outbound.CursorRepository    = (*cursors)(nil)
	_ outbound.SnapshotRepository  = (*snapshots)(nil)
	_ outbound.ODataSchreiber      = (*schreiber)(nil)
	_ outbound.CoverageRepository  = (*coverage)(nil)
	_ inbound.MockUseCase          = (*mock.Mocker)(nil)
)

// cursors ist der Fake von CursorRepository: jede Operation zählt, der Dienst ruft keine (SPEC-192 Nr. 6).
type cursors struct{ p *prot }

func (c *cursors) Put(ctx context.Context, x model.ReplayCursor) error {
	return c.p.beginn(ctx, "Cursors.Put", x)
}

func (c *cursors) Get(ctx context.Context, w, r, s string) (model.ReplayCursor, error) {
	return model.ReplayCursor{}, c.p.beginn(ctx, "Cursors.Get", w, r, s)
}

func (c *cursors) List(ctx context.Context, w string) ([]model.ReplayCursor, error) {
	return nil, c.p.beginn(ctx, "Cursors.List", w)
}

func (c *cursors) ResetAll(ctx context.Context, w string) error {
	return c.p.beginn(ctx, "Cursors.ResetAll", w)
}

// snapshots ist der Fake von SnapshotRepository: jede Operation zählt, der Dienst ruft keine.
type snapshots struct{ p *prot }

func (s *snapshots) Get(ctx context.Context, id string) (model.BaselineSnapshot, error) {
	return model.BaselineSnapshot{}, s.p.beginn(ctx, "Snapshots.Get", id)
}

func (s *snapshots) List(ctx context.Context, id string) ([]model.BaselineSnapshot, error) {
	return nil, s.p.beginn(ctx, "Snapshots.List", id)
}

func (s *snapshots) CreateWorkspace(ctx context.Context, x model.WorkspaceSnapshot) error {
	return s.p.beginn(ctx, "Snapshots.CreateWorkspace", x)
}

func (s *snapshots) GetWorkspace(ctx context.Context, w, id string) (model.WorkspaceSnapshot, error) {
	return model.WorkspaceSnapshot{}, s.p.beginn(ctx, "Snapshots.GetWorkspace", w, id)
}

func (s *snapshots) ListWorkspace(ctx context.Context, w string) ([]model.WorkspaceSnapshot, error) {
	return nil, s.p.beginn(ctx, "Snapshots.ListWorkspace", w)
}

// transaktion ist der Fake von WorkspaceTransactor: Transaktionen laufen hintereinander (eine Sperre), f erhält die Fakes
// des Aufbaus, und ein Fehler oder eine Panik von f stellt Entity-State und Revisionen auf den Stand vor f zurück.
type transaktion struct {
	p  *prot
	ws *workspaces
	en *entities
	cu *cursors
	sn *snapshots
	mu sync.Mutex
	// innen läuft für jede Transaktion mit den Stores, die f erhält; ein Test ersetzt damit Fakes oder hält an.
	innen func(outbound.WorkspaceStores) outbound.WorkspaceStores
	// nachF ist der Fehler der Bestätigung: ein Wert ungleich nil rollt zurück und wird anstelle von nil geliefert.
	nachF error
	// vor ist der Haltepunkt am Eingang: er läuft nach dem Aufruf und vor der Sperre, also zwischen allem, was der Dienst vor der
	// Transaktion liest, und dem Schreiben in ihr.
	vor func()
}

func (t *transaktion) WithWorkspaceTransaction(ctx context.Context, f func(outbound.WorkspaceStores) error) (err error) {
	if err := t.p.beginn(ctx, "Transaktion.WithWorkspaceTransaction", f); err != nil {
		return err
	}
	if t.vor != nil {
		t.vor()
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	entities := t.en.inhalt()
	t.ws.mu.Lock()
	workspaces := maps.Clone(t.ws.m)
	t.ws.mu.Unlock()
	fertig := false
	defer func() {
		if !fertig {
			t.en.stelleHer(entities)
			t.ws.mu.Lock()
			t.ws.m = workspaces
			t.ws.mu.Unlock()
		}
	}()
	s := outbound.WorkspaceStores{Workspaces: t.ws, Entities: t.en, Cursors: t.cu, Snapshots: t.sn}
	if t.innen != nil {
		s = t.innen(s)
	}
	err = f(s)
	if err == nil && t.nachF != nil {
		err = t.nachF
	}
	fertig = err == nil
	return err
}

// eingabeFake ist der Fake von ODataEingabe: LeseEntity hält jedes Argument fest und liefert f, sonst das feste Ergebnis.
type eingabeFake struct {
	p        *prot
	ergebnis model.EntityEingabe
	f        func(ctx context.Context, src io.Reader, contentType []string, entitySet string, art model.EingabeArt) (model.EntityEingabe, error)
}

func (e *eingabeFake) LeseEntity(ctx context.Context, src io.Reader, contentType []string, m model.ServiceModel, entitySet string, art model.EingabeArt, grenze int64) (model.EntityEingabe, error) {
	if err := e.p.beginn(ctx, "Eingabe.LeseEntity", src, contentType, m, entitySet, art, grenze); err != nil {
		return model.EntityEingabe{}, err
	}
	if e.f != nil {
		return e.f(ctx, src, contentType, entitySet, art)
	}
	return e.ergebnis, nil
}

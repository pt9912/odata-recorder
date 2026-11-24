package admin

import (
	"context"
	"errors"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// workspacesPfad ist das Muster der Routen für Workspaces (SPEC-298 Nr. 1).
const workspacesPfad = "/__admin/workspaces"

// workspaceHandler hält allein die drei Ports der Routen (SPEC-298 Nr. 1, 12); von audit ruft er nur Append, einmal nach dem
// Erfolg der vier POST-Routen mit Eintrag (Nr. 11).
type workspaceHandler struct {
	workspaces inbound.WorkspaceUseCase
	snapshots  inbound.SnapshotUseCase
	audit      support.AuditLog
}

// workspacePortFehlt bildet den Fehler von WorkspaceRouten bei einem nil-Port: model.ErrUngueltig, der Text nennt keinen Wert.
func workspacePortFehlt(was string) error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: WorkspaceRouten: "+was+" fehlt"))
}

// WorkspaceRouten liefert die acht Routen der Workspaces und Snapshots in der Folge von SPEC-298 Nr. 1. Die Funktion ruft
// keinen Port und startet keine Goroutine; sie liefert model.ErrUngueltig und keine Routen bei einem nil-Wert.
func WorkspaceRouten(workspaces inbound.WorkspaceUseCase, snapshots inbound.SnapshotUseCase, audit support.AuditLog) ([]Route, error) {
	switch {
	case workspaces == nil:
		return nil, workspacePortFehlt("workspaces")
	case snapshots == nil:
		return nil, workspacePortFehlt("snapshots")
	case audit == nil:
		return nil, workspacePortFehlt("audit")
	}
	h := workspaceHandler{workspaces: workspaces, snapshots: snapshots, audit: audit}
	je := workspacesPfad + "/{" + pfadID + "}"
	return []Route{
		{Methode: http.MethodGet, Muster: workspacesPfad, Handler: h.liste},
		{Methode: http.MethodPost, Muster: workspacesPfad, Handler: h.anlege},
		{Methode: http.MethodGet, Muster: je, Handler: h.lese},
		{Methode: http.MethodDelete, Muster: je, Handler: h.loesche},
		{Methode: http.MethodPost, Muster: je + "/reset", Handler: h.zuruecksetze},
		{Methode: http.MethodGet, Muster: je + "/snapshots", Handler: h.snapshotListe},
		{Methode: http.MethodPost, Muster: je + "/snapshots", Handler: h.snapshotErzeuge},
		{Methode: http.MethodPost, Muster: je + "/restore", Handler: h.stelleWiederHer},
	}, nil
}

// liste ruft List mit dem ctx des Handlers; Query und Body bleiben ungelesen (SPEC-298 Nr. 2, 3, 6, 7 d). Jeder Fehler des
// Dienstes geht unverändert und ohne Antwort zurück (Nr. 8).
func (h workspaceHandler) liste(ctx context.Context, _ Anfrage) (Antwort, error) {
	ws, err := h.workspaces.List(ctx)
	if err != nil {
		return Antwort{}, err
	}
	out := make([]workspaceJSON, len(ws))
	for i, w := range ws {
		out[i] = workspaceAus(w)
	}
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(workspaceListeJSON{Workspaces: out})}, nil
}

// anlege liest den Body (SPEC-298 Nr. 5) und ruft Create mit dem ctx des Handlers; die Antwort ist 201 mit dem Workspace. Nach
// dem Erfolg, auch bei beendetem ctx (Nr. 9 d), schreibt die Route genau einen Eintrag `workspace_angelegt` aus dem Ergebnis des
// Dienstes und der Korrelations-ID der Antwort; das Ergebnis von Append ändert nichts (Nr. 11).
func (h workspaceHandler) anlege(ctx context.Context, a Anfrage) (Antwort, error) {
	baselineID, id, err := leseAnlegenWorkspace(a.Body)
	if err != nil {
		return Antwort{}, err
	}
	w, err := h.workspaces.Create(ctx, baselineID, id)
	if err != nil {
		return Antwort{}, err
	}
	_ = h.audit.Append(support.Eintrag{Art: support.ArtWorkspaceAngelegt, ServiceID: w.ServiceID, WorkspaceID: w.ID, KorrelationsID: a.KorrelationsID})
	return Antwort{Status: http.StatusCreated, Body: schreibeJSON(workspaceAus(w))}, nil
}

// lese ruft Get mit dem Pfadwert unverändert (SPEC-298 Nr. 2, 4).
func (h workspaceHandler) lese(ctx context.Context, a Anfrage) (Antwort, error) {
	w, err := h.workspaces.Get(ctx, a.Pfadwerte[pfadID])
	if err != nil {
		return Antwort{}, err
	}
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(workspaceAus(w))}, nil
}

// loesche ruft Delete mit dem Pfadwert unverändert und antwortet mit dem Bericht des Dienstes (SPEC-298 Nr. 10).
func (h workspaceHandler) loesche(ctx context.Context, a Anfrage) (Antwort, error) {
	b, err := h.workspaces.Delete(ctx, a.Pfadwerte[pfadID])
	if err != nil {
		return Antwort{}, err
	}
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(loeschberichtJSON{BlobsRemoved: b.Entfernt, CleanupDeferred: b.Zurueckgestellt})}, nil
}

// zuruecksetze ruft Reset mit dem Pfadwert unverändert; der Body bleibt ungelesen (SPEC-298 Nr. 3). Nach dem Erfolg schreibt die
// Route einen Eintrag `workspace_zurueckgesetzt` wie anlege (Nr. 11).
func (h workspaceHandler) zuruecksetze(ctx context.Context, a Anfrage) (Antwort, error) {
	w, err := h.workspaces.Reset(ctx, a.Pfadwerte[pfadID])
	if err != nil {
		return Antwort{}, err
	}
	_ = h.audit.Append(support.Eintrag{Art: support.ArtWorkspaceZurueckgesetzt, ServiceID: w.ServiceID, WorkspaceID: w.ID, KorrelationsID: a.KorrelationsID})
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(workspaceAus(w))}, nil
}

// snapshotListe ruft snapshots.List mit dem Pfadwert unverändert (SPEC-298 Nr. 2, 6, 7 d).
func (h workspaceHandler) snapshotListe(ctx context.Context, a Anfrage) (Antwort, error) {
	ss, err := h.snapshots.List(ctx, a.Pfadwerte[pfadID])
	if err != nil {
		return Antwort{}, err
	}
	out := make([]snapshotJSON, len(ss))
	for i, s := range ss {
		out[i] = snapshotAus(s)
	}
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(snapshotListeJSON{Snapshots: out})}, nil
}

// snapshotErzeuge liest den Body (SPEC-298 Nr. 5) und ruft snapshots.Create; die Antwort ist 201 mit dem Snapshot. Nach dem Erfolg
// schreibt die Route einen Eintrag `snapshot_erzeugt` mit `ServiceID` und `WorkspaceID` des Snapshots (Nr. 11).
func (h workspaceHandler) snapshotErzeuge(ctx context.Context, a Anfrage) (Antwort, error) {
	name, err := leseSnapshotName(a.Body)
	if err != nil {
		return Antwort{}, err
	}
	s, err := h.snapshots.Create(ctx, a.Pfadwerte[pfadID], name)
	if err != nil {
		return Antwort{}, err
	}
	_ = h.audit.Append(support.Eintrag{Art: support.ArtSnapshotErzeugt, ServiceID: s.ServiceID, WorkspaceID: s.WorkspaceID, KorrelationsID: a.KorrelationsID})
	return Antwort{Status: http.StatusCreated, Body: schreibeJSON(snapshotAus(s))}, nil
}

// stelleWiederHer liest den Body (SPEC-298 Nr. 5) und ruft Restore; die Antwort ist 200 mit dem Workspace. Nach dem Erfolg
// schreibt die Route einen Eintrag `workspace_wiederhergestellt` wie anlege (Nr. 11).
func (h workspaceHandler) stelleWiederHer(ctx context.Context, a Anfrage) (Antwort, error) {
	snapshotID, err := leseRestore(a.Body)
	if err != nil {
		return Antwort{}, err
	}
	w, err := h.workspaces.Restore(ctx, a.Pfadwerte[pfadID], snapshotID)
	if err != nil {
		return Antwort{}, err
	}
	_ = h.audit.Append(support.Eintrag{Art: support.ArtWorkspaceWiederhergestellt, ServiceID: w.ServiceID, WorkspaceID: w.ID, KorrelationsID: a.KorrelationsID})
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(workspaceAus(w))}, nil
}

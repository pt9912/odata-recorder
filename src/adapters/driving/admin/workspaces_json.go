package admin

import (
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// workspaceJSON ist der Workspace der Antwort: die fünf Felder von model.Workspace in der Folge des Vertrags; kein Feld
// entfällt (SPEC-298 Nr. 7 a, c).
type workspaceJSON struct {
	ID                 string   `json:"id"`
	ServiceID          string   `json:"service_id"`
	BaselineSnapshotID string   `json:"baseline_snapshot_id"`
	Revision           int64    `json:"revision"`
	CreatedAt          zeitJSON `json:"created_at"`
}

// workspaceListeJSON ist der Body von `GET /__admin/workspaces`; die Liste ist nie `null` (SPEC-298 Nr. 7 d).
type workspaceListeJSON struct {
	Workspaces []workspaceJSON `json:"workspaces"`
}

// snapshotJSON ist der Snapshot der Antwort: vier der acht Felder von model.WorkspaceSnapshot; Dienst, Blob und beide Hashes
// stehen in keiner Antwort (SPEC-298 Nr. 7 b).
type snapshotJSON struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspace_id"`
	Name        string   `json:"name"`
	CreatedAt   zeitJSON `json:"created_at"`
}

// snapshotListeJSON ist der Body von `GET /__admin/workspaces/{id}/snapshots`; die Liste ist nie `null`.
type snapshotListeJSON struct {
	Snapshots []snapshotJSON `json:"snapshots"`
}

// workspaceAus bildet den Workspace des Dienstes ab, ohne ein Feld zu prüfen oder zu ändern.
func workspaceAus(w model.Workspace) workspaceJSON {
	return workspaceJSON{
		ID: w.ID, ServiceID: w.ServiceID, BaselineSnapshotID: w.BaselineSnapshotID, Revision: w.Revision, CreatedAt: zeitJSON(w.CreatedAt),
	}
}

// snapshotAus bildet den Snapshot des Dienstes ab, ohne ein Feld zu prüfen oder zu ändern.
func snapshotAus(s model.WorkspaceSnapshot) snapshotJSON {
	return snapshotJSON{ID: s.ID, WorkspaceID: s.WorkspaceID, Name: s.Name, CreatedAt: zeitJSON(s.CreatedAt)}
}

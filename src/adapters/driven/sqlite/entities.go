package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// EntityRepository erfüllt outbound.EntityRepository auf SQLite (SPEC-108).
type EntityRepository struct {
	z zugriff
}

var _ outbound.EntityRepository = (*EntityRepository)(nil)

// Entities liefert das Repository für den Entity-State.
func (d *DB) Entities() *EntityRepository { return &EntityRepository{z: zugriff{db: d}} }

const entitySpalten = "workspace_id, entity_set, entity_key, payload_json, known_properties_json, entity_completeness, entity_version, deleted"

const entityUpsert = "INSERT INTO entities(" + entitySpalten + ") VALUES (?, ?, ?, ?, ?, ?, ?, ?)" +
	" ON CONFLICT(workspace_id, entity_set, entity_key) DO UPDATE SET payload_json = excluded.payload_json," +
	" known_properties_json = excluded.known_properties_json, entity_completeness = excluded.entity_completeness," +
	" entity_version = excluded.entity_version, deleted = excluded.deleted"

func schreibeEntity(ctx context.Context, q querier, e model.EntityState) error {
	geloescht := 0
	if e.Deleted {
		geloescht = 1
	}
	_, err := q.ExecContext(ctx, entityUpsert, e.WorkspaceID, e.EntitySet, e.EntityKey, e.Payload,
		e.KnownPropertiesJSON(), string(e.Completeness), e.Version, geloescht)
	return klassifiziere(err)
}

// Put schreibt oder ersetzt einen Eintrag.
func (r *EntityRepository) Put(ctx context.Context, e model.EntityState) error {
	if err := e.Validate(); err != nil {
		return err
	}
	return r.z.schreibe(ctx, func(q querier) error {
		if err := pruefeWorkspace(ctx, q, e.WorkspaceID); err != nil {
			return err
		}
		return schreibeEntity(ctx, q, e)
	})
}

func leseEntity(row zeile) (model.EntityState, error) {
	var e model.EntityState
	var bekannt, vollstaendig string
	var geloescht int
	if err := row.Scan(&e.WorkspaceID, &e.EntitySet, &e.EntityKey, &e.Payload, &bekannt, &vollstaendig, &e.Version, &geloescht); err != nil {
		return model.EntityState{}, err
	}
	e.Completeness, e.Deleted = model.Completeness(vollstaendig), geloescht == 1
	if err := json.Unmarshal([]byte(bekannt), &e.KnownProperties); err != nil {
		return model.EntityState{}, fmt.Errorf("sqlite: known_properties von %s/%s: %w", e.EntitySet, e.EntityKey, err)
	}
	return e, nil
}

// Get liest einen Eintrag, auch einen Tombstone.
func (r *EntityRepository) Get(ctx context.Context, workspaceID, entitySet, entityKey string) (model.EntityState, error) {
	e, err := leseEntity(r.z.QueryRowContext(ctx, "SELECT "+entitySpalten+
		" FROM entities WHERE workspace_id = ? AND entity_set = ? AND entity_key = ?", workspaceID, entitySet, entityKey))
	if errors.Is(err, sql.ErrNoRows) {
		return model.EntityState{}, fmt.Errorf("%w: Entity %s/%s in Workspace %s", outbound.ErrNichtGefunden, entitySet, entityKey, workspaceID)
	}
	return e, klassifiziere(err)
}

// List liefert alle Einträge eines Workspace nach entity_set und entity_key.
func (r *EntityRepository) List(ctx context.Context, workspaceID string) ([]model.EntityState, error) {
	return r.liste(ctx, workspaceID, "", "")
}

// ListSet liefert die Einträge einer Menge nach entity_key.
func (r *EntityRepository) ListSet(ctx context.Context, workspaceID, entitySet string) ([]model.EntityState, error) {
	if entitySet == "" {
		return nil, ungueltigerWert("leeres entity_set")
	}
	return r.liste(ctx, workspaceID, " AND entity_set = ?", entitySet)
}

func (r *EntityRepository) liste(ctx context.Context, workspaceID, filter string, args ...any) ([]model.EntityState, error) {
	q := r.z
	rows, err := q.QueryContext(ctx, "SELECT "+entitySpalten+" FROM entities WHERE workspace_id = ?"+filter+
		" ORDER BY entity_set, entity_key", append([]any{workspaceID}, args...)...)
	if err != nil {
		return nil, klassifiziere(err)
	}
	out, err := sammle(rows, leseEntity)
	if err != nil {
		return nil, err
	}
	return leereListe(ctx, q, out, workspaceID)
}

func pruefeErsatz(workspaceID string, eintraege []model.EntityState) error {
	gesehen := map[[2]string]bool{}
	for _, e := range eintraege {
		if e.WorkspaceID != workspaceID {
			return ungueltigerWert("Entity %s/%s gehört zu Workspace %q, nicht zu %q", e.EntitySet, e.EntityKey, e.WorkspaceID, workspaceID)
		}
		if err := e.Validate(); err != nil {
			return err
		}
		k := [2]string{e.EntitySet, e.EntityKey}
		if gesehen[k] {
			return ungueltigerWert("Entity %s/%s doppelt", e.EntitySet, e.EntityKey)
		}
		gesehen[k] = true
	}
	return nil
}

// ReplaceAll ersetzt den gesamten State des Workspace.
func (r *EntityRepository) ReplaceAll(ctx context.Context, workspaceID string, eintraege []model.EntityState) error {
	if err := pruefeErsatz(workspaceID, eintraege); err != nil {
		return err
	}
	return r.z.schreibe(ctx, func(q querier) error {
		if err := pruefeWorkspace(ctx, q, workspaceID); err != nil {
			return err
		}
		if _, err := q.ExecContext(ctx, "DELETE FROM entities WHERE workspace_id = ?", workspaceID); err != nil {
			return klassifiziere(err)
		}
		for _, e := range eintraege {
			if err := schreibeEntity(ctx, q, e); err != nil {
				return err
			}
		}
		return nil
	})
}

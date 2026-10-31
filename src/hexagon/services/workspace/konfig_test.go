package workspace_test

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// TestNewNilFeld (Z01): ein nil in einem der neun Feldern, auch in jedem Paar von zweien, ergibt
// ErrUngueltig ohne Code und keinen Dienst; die anderen Felder sind gesetzt.
func TestNewNilFeld(t *testing.T) {
	w := neu(t)
	typ := reflect.TypeOf(workspace.Konfig{})
	if typ.NumField() != 9 {
		t.Fatalf("Felder %d", typ.NumField())
	}
	faelle := 0
	for i := 0; i < typ.NumField(); i++ {
		for j := i; j < typ.NumField(); j++ {
			k := w.konfig
			v := reflect.ValueOf(&k).Elem()
			v.Field(i).SetZero()
			v.Field(j).SetZero()
			d, err := workspace.New(k)
			if d != nil {
				t.Errorf("%s/%s: Dienst trotz nil", typ.Field(i).Name, typ.Field(j).Name)
			}
			pruefeOhneCode(t, err, model.ErrUngueltig)
			faelle++
		}
	}
	if faelle != 9+36 {
		t.Errorf("Fälle %d, erwartet 45 (neun einzelne, 36 Paare)", faelle)
	}
}

// TestNewRuftKeinenPort (Z02): New ruft keinen Port, und jede der Operationen im Erfolgsfall ruft nur die
// benannten Operationen.
func TestNewRuftKeinenPort(t *testing.T) {
	w := neu(t)
	if _, err := workspace.New(w.konfig); err != nil {
		t.Fatal(err)
	}
	if w.p.gesamt() != 0 {
		t.Errorf("New rief %v", w.p.ops())
	}
}

// erlaubt sind die Operationen der Fakes, die eine Operation im Erfolgsfall rufen darf (SPEC-129 Nr. 1).
func erlaubt() map[string][]string {
	baseline := []string{"Snapshots.Get", "Blobs.Open", "Strom.Read", "Strom.Close"}
	return map[string][]string{
		"Create": append([]string{"Uhr.Now", "IDs.WorkspaceID", lblTx, "tx Workspaces.Create", "tx Entities.ReplaceAll"}, baseline...),
		"Get":    {"Workspaces.Get"},
		"List":   {"Dienste.List", "Workspaces.List"},
		"Delete": {"Workspaces.Delete", "Wartung.CleanupBlobs"},
		"Reset": append([]string{"Workspaces.Get", lblTx, "tx Workspaces.Get", "tx Entities.ReplaceAll", "tx Cursors.ResetAll",
			"tx Workspaces.Advance"}, baseline...),
	}
}

func TestNurBenannteOperationen(t *testing.T) {
	for _, o := range operationen() {
		w := neu(t)
		o.vor(w)
		w.p.leeren()
		if err := o.lauf(w, context.Background()); err != nil {
			t.Fatalf("%s: %v", o.name, err)
		}
		for _, op := range w.p.ops() {
			if !slices.Contains(erlaubt()[o.name], op) {
				t.Errorf("%s ruft %s", o.name, op)
			}
		}
		if o.name != "Create" && (w.p.anzahl("Uhr.Now") != 0 || w.p.anzahl("IDs.WorkspaceID") != 0) {
			t.Errorf("%s bezieht Zeit oder Kennung", o.name)
		}
	}
}

// TestPortUndKonfigFestgeschrieben (Z03): Namen und Typen der Felder, Methoden und Signaturen des Ports,
// Signatur von New und MaxIDLaenge.
func TestPortUndKonfigFestgeschrieben(t *testing.T) {
	soll := []struct {
		name string
		typ  reflect.Type
	}{
		{"Workspaces", reflect.TypeOf((*outbound.WorkspaceRepository)(nil)).Elem()},
		{"Transaktion", reflect.TypeOf((*outbound.WorkspaceTransactor)(nil)).Elem()},
		{"Snapshots", reflect.TypeOf((*outbound.SnapshotRepository)(nil)).Elem()},
		{"Blobs", reflect.TypeOf((*outbound.BlobRepository)(nil)).Elem()},
		{"Dienste", reflect.TypeOf((*outbound.MetadataRepository)(nil)).Elem()},
		{"Wartung", reflect.TypeOf((*inbound.MaintenanceUseCase)(nil)).Elem()},
		{"Uhr", reflect.TypeOf((*support.Clock)(nil)).Elem()},
		{"IDs", reflect.TypeOf((*support.IDGenerator)(nil)).Elem()},
	}
	typ := reflect.TypeOf(workspace.Konfig{})
	if typ.NumField() != 9 {
		t.Fatalf("Felder %d", typ.NumField())
	}
	for i, s := range soll {
		if f := typ.Field(i); f.Name != s.name || f.Type != s.typ {
			t.Errorf("Feld %d: %s %v, erwartet %s %v", i, f.Name, f.Type, s.name, s.typ)
		}
	}
	if f := typ.Field(8); f.Name != "Log" || f.Type.String() != "*slog.Logger" {
		t.Errorf("Feld 8: %s %v", f.Name, f.Type)
	}
	port := reflect.TypeOf((*inbound.WorkspaceUseCase)(nil)).Elem()
	methoden := map[string]string{
		"Create": "func(context.Context, string, string) (model.Workspace, error)",
		"Delete": "func(context.Context, string) (inbound.Loeschbericht, error)",
		"Get":    "func(context.Context, string) (model.Workspace, error)",
		"List":   "func(context.Context) ([]model.Workspace, error)",
		"Reset":  "func(context.Context, string) (model.Workspace, error)",
	}
	if port.NumMethod() != len(methoden) {
		t.Errorf("Methoden %d", port.NumMethod())
	}
	for i := 0; i < port.NumMethod(); i++ {
		m := port.Method(i)
		if methoden[m.Name] != m.Type.String() {
			t.Errorf("%s: %s, erwartet %s", m.Name, m.Type, methoden[m.Name])
		}
	}
	if got := reflect.TypeOf(workspace.New).String(); got != "func(workspace.Konfig) (*workspace.Verwaltung, error)" {
		t.Errorf("New: %s", got)
	}
	if workspace.MaxIDLaenge != 64 {
		t.Errorf("MaxIDLaenge %d", workspace.MaxIDLaenge)
	}
}

// TestVerwaltungOhneZustand (Z04): die Verwaltung hat als einziges Feld die Konfig.
func TestVerwaltungOhneZustand(t *testing.T) {
	typ := reflect.TypeOf(workspace.Verwaltung{})
	if typ.NumField() != 1 || typ.Field(0).Type != reflect.TypeOf(workspace.Konfig{}) {
		t.Errorf("Felder der Verwaltung: %d", typ.NumField())
	}
}

// TestKontextNilJedeOperation (Z05): ctx nil ist bei jeder Operation ErrUngueltig ohne Code, und kein Port,
// keine Kennung und keine Zeit werden gerufen.
func TestKontextNilJedeOperation(t *testing.T) {
	var kein context.Context
	for _, o := range operationen() {
		w := neu(t)
		o.vor(w)
		w.p.leeren()
		pruefeOhneCode(t, o.lauf(w, kein), model.ErrUngueltig)
		if w.p.gesamt() != 0 {
			t.Errorf("%s: Aufrufe %v", o.name, w.p.ops())
		}
	}
}

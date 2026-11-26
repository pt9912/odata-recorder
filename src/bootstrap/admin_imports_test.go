package bootstrap_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// adminAufnahme fährt den Modus record gegen einen Upstream mit $metadata und Products hoch, zeichnet beide in der ersten Session
// auf und beendet sie; es liefert Handle, Ablage, Datenverzeichnis und die Kennung der Session.
func adminAufnahme(t *testing.T) (*bootstrap.Handle, *sqlite.DB, string, string) {
	t.Helper()
	ctx := context.Background()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/$metadata") {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = io.WriteString(w, msProductsXML())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"@odata.context":"$metadata#Products","value":[{"Id":1,"Name":"Milk"}]}`)
	}))
	t.Cleanup(up.Close)
	dir := t.TempDir()
	h, db := adminBau(t, routenKonfig(dir, modusRecord, up.URL+"/o"), nil, nil)
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatal(err)
	}
	offen, err := h.Sessions().List(ctx)
	if err != nil || len(offen) != 1 {
		t.Fatalf("Sessions %+v, %v", offen, err)
	}
	for _, pfad := range []string{"/r1/$metadata", "/r1/Products"} {
		a, err := h.Record().Record(ctx, inbound.Anfrage{Methode: "GET", Pfad: pfad, KorrelationsID: "cor-auf"})
		if err != nil || a.Status != 200 {
			t.Fatalf("Record %s: %d, %v", pfad, a.Status, err)
		}
		_, _ = io.Copy(io.Discard, a.Body)
		_ = a.Body.Close()
	}
	if _, err := h.Sessions().Beenden(ctx, offen[0].ID); err != nil {
		t.Fatal(err)
	}
	return h, db, dir, offen[0].ID
}

// TestAdminImportsVerdrahtet fängt A04: im Modus record die Handler von POST und GET gegen Ablage und audit.jsonl, in den drei
// anderen Modi ein Build mit Import() gleich nil ohne eine Route der Imports (2 + 3 Zellen).
func TestAdminImportsVerdrahtet(t *testing.T) {
	h, db, dir, quelle := adminAufnahme(t)
	zellen := 0
	report := rufe(t, h, 201, "POST", "/__admin/imports", "", "", `{"import_id":"imp-a","service_id":"x1","source_session_ids":["`+quelle+`"]}`)
	imp, err := db.Imports().Get(context.Background(), "imp-a")
	if err != nil || imp.BaselineSnapshotID == "" || imp.ReportJSON != report {
		t.Fatalf("Import %+v, %v; Antwort %s", imp, err, report)
	}
	if _, err := db.Snapshots().Get(context.Background(), imp.BaselineSnapshotID); err != nil {
		t.Errorf("Baseline: %v", err)
	}
	eintraege := auditEintraege(t, dir)
	letzter := eintraege[len(eintraege)-1]
	if letzter["art"] != "import_veroeffentlicht" || letzter["import_id"] != "imp-a" || letzter["correlation_id"] != "cor-test" {
		t.Errorf("Eintrag %v", letzter)
	}
	zellen++
	if got := rufe(t, h, 200, "GET", "/__admin/imports/{id}", "imp-a", "", ""); got != report {
		t.Errorf("GET %s, erwartet %s", got, report)
	}
	zellen++
	for _, modus := range []string{modusReplay, modusMock, modusHybrid} {
		h, _ := adminBau(t, routenKonfig(t.TempDir(), modus, "http://u1.test/o"), nil, nil)
		for _, r := range bootstrap.AdminRouten(h) {
			if strings.HasPrefix(r.Muster, "/__admin/imports") {
				t.Errorf("%s: Route %s %s", modus, r.Methode, r.Muster)
			}
		}
		if h.Import() != nil {
			t.Errorf("%s: Import() %v", modus, h.Import())
		}
		zellen++
	}
	if zellen != 5 {
		t.Errorf("%d Zellen", zellen)
	}
}

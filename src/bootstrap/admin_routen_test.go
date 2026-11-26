package bootstrap_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// Die Fänger der Verdrahtung der Admin-Endpunkte (SPEC-300 Nr. 5, 13 a, b) rufen die Routen und Handler, die Build im Handle
// ablegt; keiner öffnet einen Listener des Admin-Adapters.

// adminModi sind die vier Modi, die Achse M-Modus.
func adminModi() []string { return []string{modusRecord, modusReplay, modusMock, modusHybrid} }

// routenKonfig ist eine Konfiguration, die in jedem Modus baut: ein Dienst x1 mit Root /r1.
func routenKonfig(dir, modus, upstream string) bootstrap.Konfiguration {
	k := basisKonfig(dir, modus, bootstrap.Dienst{ID: "x1", Root: "/r1", UpstreamURL: upstream})
	k.PublicBaseURL = "http://recorder.test/r1"
	return k
}

// adminBau baut mit den Konstruktoren des Betriebs, die setze ändert, und liefert die Ablage, die Build öffnete.
func adminBau(t *testing.T, k bootstrap.Konfiguration, senke io.Writer, setze func(*bootstrap.Bauteile)) (*bootstrap.Handle, *sqlite.DB) {
	t.Helper()
	b := bootstrap.StandardBauteile()
	var db *sqlite.DB
	oeffne := b.Oeffne
	b.Oeffne = func(ctx context.Context, pfad string, opt sqlite.Options) (*sqlite.DB, error) {
		d, err := oeffne(ctx, pfad, opt)
		db = d
		return d, err
	}
	if setze != nil {
		setze(&b)
	}
	if senke == nil {
		senke = io.Discard
	}
	h, err := bootstrap.BuildMitSenke(context.Background(), k, nil, senke, b)
	if err != nil {
		t.Fatalf("Build im Modus %s: %v", k.Modus, err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(context.Background()) })
	return h, db
}

// rufe ruft den Handler der Route mit Methode und Muster und verlangt den Status; die Korrelations-ID ist cor-test.
func rufe(t *testing.T, h *bootstrap.Handle, status int, methode, muster, id, query, body string) string {
	t.Helper()
	for _, r := range bootstrap.AdminRouten(h) {
		if r.Methode != methode || r.Muster != muster {
			continue
		}
		a, err := r.Handler(context.Background(), admin.Anfrage{
			Methode: methode, RawQuery: query, KorrelationsID: "cor-test", Pfadwerte: map[string]string{"id": id}, Body: strings.NewReader(body),
		})
		if err != nil || a.Status != status {
			t.Fatalf("%s %s: Status %d, %v; erwartet %d", methode, muster, a.Status, err, status)
		}
		return string(a.Body)
	}
	t.Fatalf("keine Route %s %s", methode, muster)
	return ""
}

// auditEintraege liefert je Zeile von audit.jsonl die Felder.
func auditEintraege(t *testing.T, dir string) []map[string]string {
	t.Helper()
	roh, err := os.ReadFile(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var out []map[string]string
	for _, zeile := range strings.Split(strings.TrimSpace(string(roh)), "\n") {
		felder := map[string]string{}
		if err := json.Unmarshal([]byte(zeile), &felder); err != nil {
			t.Fatalf("Zeile %q: %v", zeile, err)
		}
		out = append(out, felder)
	}
	return out
}

// adminSoll sind die Zeilen von SPEC-005 ohne GET /healthz in der Folge der Konstruktoren: Sessions 4, Inspektion 3, Imports 2,
// Workspaces 8; die der Imports gelten nur im Modus record.
func adminSoll() []string {
	return []string{
		"GET /__admin/sessions", "POST /__admin/sessions", "POST /__admin/sessions/{id}/close", "DELETE /__admin/sessions/{id}",
		"GET /__admin/sessions/{id}/exchanges", "GET /__admin/coverage", "GET /__admin/diagnostics",
		"POST /__admin/imports", "GET /__admin/imports/{id}",
		"GET /__admin/workspaces", "POST /__admin/workspaces", "GET /__admin/workspaces/{id}", "DELETE /__admin/workspaces/{id}",
		"POST /__admin/workspaces/{id}/reset", "GET /__admin/workspaces/{id}/snapshots", "POST /__admin/workspaces/{id}/snapshots",
		"POST /__admin/workspaces/{id}/restore",
	}
}

// TestAdminRouten fängt A01: je Modus und je Zeile, ob die Route im Handle liegt (17 x 4 Zellen), die Folge der Routen und je
// Route ein Handler.
func TestAdminRouten(t *testing.T) {
	anzahl := map[string]int{modusRecord: 17, modusReplay: 15, modusMock: 15, modusHybrid: 15}
	zellen := 0
	for _, modus := range adminModi() {
		h, _ := adminBau(t, routenKonfig(t.TempDir(), modus, "http://u1.test/o"), nil, nil)
		var da, soll []string
		for _, r := range bootstrap.AdminRouten(h) {
			da = append(da, r.Methode+" "+r.Muster)
			if r.Handler == nil {
				t.Errorf("%s: %s %s ohne Handler", modus, r.Methode, r.Muster)
			}
		}
		for _, zeile := range adminSoll() {
			erwartet := modus == modusRecord || !strings.Contains(zeile, "/__admin/imports")
			if erwartet {
				soll = append(soll, zeile)
			}
			if slices.Contains(da, zeile) != erwartet {
				t.Errorf("%s: Route %s vorhanden %v, erwartet %v", modus, zeile, !erwartet, erwartet)
			}
			zellen++
		}
		if !slices.Equal(da, soll) || len(da) != anzahl[modus] {
			t.Errorf("%s: %d Routen %v, erwartet %d in der Folge %v", modus, len(da), da, anzahl[modus], soll)
		}
	}
	if zellen != 68 || len(adminSoll()) != 4+3+2+8 {
		t.Errorf("%d Zellen, %d Zeilen", zellen, len(adminSoll()))
	}
}

package bootstrap_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// TestAdminSessionsVerdrahtet fängt A02: je Modus die vier Handler gegen die Ablage und audit.jsonl (4 x 4 Zellen).
func TestAdminSessionsVerdrahtet(t *testing.T) {
	ctx := context.Background()
	const basis, je = "/__admin/sessions", "/__admin/sessions/{id}"
	zellen := 0
	for _, modus := range adminModi() {
		dir := t.TempDir()
		saeen(t, dir, func(d *wsDatei) { saeeDienst(d, "x1", "/r1") })
		h, db := adminBau(t, routenKonfig(dir, modus, "http://u1.test/o"), nil, nil)
		stand := func(soll model.SessionStatus) {
			t.Helper()
			if s, err := db.Sessions().Get(ctx, "ses-a"); err != nil || s.Status != soll || s.ServiceID != "x1" {
				t.Errorf("%s: Session %+v, %v; erwartet %s", modus, s, err, soll)
			}
		}
		rufe(t, h, 201, "POST", basis, "", "", `{"service_id":"x1","id":"ses-a"}`)
		stand(model.SessionOpen)
		if body := rufe(t, h, 200, "GET", basis, "", "", ""); !strings.Contains(body, `"ses-a"`) {
			t.Errorf("%s: Liste %s", modus, body)
		}
		rufe(t, h, 200, "POST", je+"/close", "ses-a", "", "")
		stand(model.SessionComplete)
		rufe(t, h, 200, "DELETE", je, "ses-a", "", "")
		if _, err := db.Sessions().Get(ctx, "ses-a"); !errors.Is(err, outbound.ErrNichtGefunden) {
			t.Errorf("%s: Session nach DELETE: %v", modus, err)
		}
		var arten []string
		for _, e := range auditEintraege(t, dir) {
			arten = append(arten, e["art"])
			if e["session_id"] != "ses-a" || e["art"] == "session_angelegt" && (e["service_id"] != "x1" || e["correlation_id"] != "cor-test") {
				t.Errorf("%s: Eintrag %v", modus, e)
			}
		}
		if got := strings.Join(arten, " "); got != "session_angelegt session_beendet session_geloescht" {
			t.Errorf("%s: Arten %q", modus, got)
		}
		zellen += 4
	}
	if zellen != 16 {
		t.Errorf("%d Zellen", zellen)
	}
}

// adminZaehlung hält, was die Haken eines Modus sahen.
type adminZaehlung struct {
	recorder, pipeline, upstream int
	session                      *session.Konfig
	wartung                      maintenance.Konfig
}

func (z *adminZaehlung) haken(b *bootstrap.Bauteile) {
	echt := bootstrap.StandardBauteile()
	b.NeuerRecorder = func(k record.Konfig) (inbound.RecordUseCase, error) { z.recorder++; return echt.NeuerRecorder(k) }
	b.NeuePipeline = func(k capture.Konfig) (*capture.Pipeline, error) { z.pipeline++; return echt.NeuePipeline(k) }
	b.NeuerUpstream = func(k httpupstream.Konfig) (outbound.UpstreamClient, error) {
		z.upstream++
		return echt.NeuerUpstream(k)
	}
	b.NeueSession = func(k session.Konfig) (inbound.SessionUseCase, error) { z.session = &k; return echt.NeueSession(k) }
	b.NeueWartung = func(k maintenance.Konfig) (inbound.MaintenanceUseCase, error) {
		z.wartung = k
		return echt.NeueWartung(k)
	}
}

// TestAdminSessionsAlleModi fängt A06: je Modus fünf Beobachtungen am Handle (Zugänge, List, Create, Beenden, Delete) und in den
// Modi ohne Recorder sieben an der session.Konfig (4 x 5 + 3 x 7 Zellen); dazu jedes Feld der Konfig und die Zahl der Bauten
// von Recorder, Pipeline und Upstream.
func TestAdminSessionsAlleModi(t *testing.T) {
	ctx := context.Background()
	abgebrochen, abbrechen := context.WithCancel(ctx)
	abbrechen()
	bauten := map[string][3]int{modusRecord: {1, 1, 1}, modusReplay: {}, modusMock: {}, modusHybrid: {0, 0, 1}}
	zellen := 0
	for _, modus := range adminModi() {
		z, senke := &adminZaehlung{}, &syncPuffer{}
		k := routenKonfig(t.TempDir(), modus, "http://u1.test/o")
		k.ShutdownTimeout = 7 * time.Second
		h, db := adminBau(t, k, senke, z.haken)
		if got := [3]int{z.recorder, z.pipeline, z.upstream}; got != bauten[modus] {
			t.Errorf("%s: Recorder, Pipeline, Upstream %v, erwartet %v", modus, got, bauten[modus])
		}
		s := h.Sessions()
		if s == nil || (h.Record() != nil) != (modus == modusRecord) {
			t.Fatalf("%s: Sessions %v, Record %v", modus, s, h.Record())
		}
		liste := func(soll int) {
			t.Helper()
			if got, err := s.List(ctx); err != nil || len(got) != soll {
				t.Errorf("%s: %d Sessions, %v; erwartet %d", modus, len(got), err, soll)
			}
		}
		liste(0)
		if err := h.Hochfahren(ctx); err != nil {
			t.Fatal(err)
		}
		erste := 0
		if modus == modusRecord {
			erste = 1
		}
		liste(erste)
		neu, err := s.Create(ctx, "x1", "ses-n")
		if err != nil || neu.Status != model.SessionOpen {
			t.Errorf("%s: Create %+v, %v", modus, neu, err)
		}
		if ende, err := s.Beenden(ctx, "ses-n"); err != nil || ende.Status != model.SessionComplete {
			t.Errorf("%s: Beenden %+v, %v", modus, ende, err)
		}
		if _, err := s.Delete(ctx, "ses-n"); err != nil {
			t.Errorf("%s: Delete %v", modus, err)
		}
		if _, err := db.Sessions().Get(ctx, "ses-n"); !errors.Is(err, outbound.ErrNichtGefunden) {
			t.Errorf("%s: Session nach Delete: %v", modus, err)
		}
		zellen += 5
		if modus == modusRecord {
			continue
		}
		got := z.session
		for _, c := range []context.Context{ctx, abgebrochen} {
			if bericht, err := got.Record.Drain(c, "ses-n", -time.Second); bericht != (inbound.DrainBericht{}) || err != nil {
				t.Errorf("%s: (1) Drain %+v, %v", modus, bericht, err)
			}
			a, err := got.Record.Record(c, inbound.Anfrage{Methode: "GET", Pfad: "/r1/Products"})
			if !errors.Is(err, model.ErrUngueltig) || !reflect.DeepEqual(a, inbound.Antwort{}) {
				t.Errorf("%s: (2) Record %+v, %v", modus, a, err)
			}
		}
		got.Log.Info("marke-der-sessions")
		_, istSessions := got.Sessions.(*sqlite.SessionRepository)
		_, istExchanges := got.Exchanges.(*sqlite.RecordingRepository)
		for nr, ok := range []bool{
			got.Frist == 7*time.Second, got.Uhr != nil && got.Uhr == z.wartung.Uhr, got.IDs != nil && got.IDs == bootstrap.IDsVon(h),
			got.Audit != nil && got.Audit == z.wartung.Audit, strings.Contains(senke.String(), "marke-der-sessions"),
		} {
			if !ok {
				t.Errorf("%s: Beobachtung (%d)", modus, nr+3)
			}
		}
		if !istSessions || !istExchanges || got.Dienste != z.wartung.Dienste || got.Wartung != h.Wartung() || reflect.TypeOf(*got).NumField() != 10 {
			t.Errorf("%s: übrige Felder %+v", modus, *got)
		}
		zellen += 7
	}
	if zellen != 41 {
		t.Errorf("%d Zellen", zellen)
	}
}

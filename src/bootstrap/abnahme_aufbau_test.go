package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/blobs"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// Aufbau, Ausgangsdaten und Hilfen der Abnahme des Mock (SPEC-049 Nr. 1 bis 3): der Prüfling ist Mocker.Mock mit den echten Adaptern
// über einer SQLite-Datei und einer Blob-Ablage; Workspace- und Snapshot-Dienst laufen über derselben Datei, Ablage und demselben Wächter.

// abW ist die Zeichenfolge W der Spezifikation: PublicBaseURL und Root des Dienstes.
const abW = "http://recorder.test/odata"

// abCSDL ist das CSDL der Ausgangsdaten (SPEC-049 Nr. 2): eine Zeile, ohne Zeilenende.
const abCSDL = `<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"><edmx:DataServices>` +
	`<Schema Namespace="N" xmlns="http://docs.oasis-open.org/odata/ns/edm"><EntityType Name="Product"><Key><PropertyRef Name="Id"/></Key>` +
	`<Property Name="Id" Type="Edm.Int32" Nullable="false"/><Property Name="Name" Type="Edm.String"/><Property Name="Active" Type="Edm.Boolean"/>` +
	`<Property Name="Stock" Type="Edm.Int64"/><Property Name="Price" Type="Edm.Decimal"/><Property Name="Ref" Type="Edm.Guid"/>` +
	`<Property Name="Released" Type="Edm.Date"/><Property Name="Updated" Type="Edm.DateTimeOffset"/></EntityType>` +
	`<EntityType Name="Order"><Key><PropertyRef Name="Id"/></Key><Property Name="Id" Type="Edm.Int32" Nullable="false"/><Property Name="Note" Type="Edm.String"/></EntityType>` +
	`<EntityType Name="Label"><Key><PropertyRef Name="Id"/></Key><Property Name="Id" Type="Edm.Int32" Nullable="false"/><Property Name="Text" Type="Edm.String"/></EntityType>` +
	`<EntityContainer Name="C"><EntitySet Name="Products" EntityType="N.Product"/><EntitySet Name="Orders" EntityType="N.Order"/>` +
	`<EntitySet Name="Labels" EntityType="N.Label"/></EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`

// abMitglieder sind die Mitglieder von Product in der Folge des Modells.
func abMitglieder() []string {
	return []string{"Id", "Name", "Active", "Stock", "Price", "Ref", "Released", "Updated"}
}

// abWerte sind die Werte der fünf Produkte der Ausgangsdaten als JSON-Text der Antwort (Stock und Price als Zahl).
func abWerte() map[int]map[string]string {
	return map[int]map[string]string{
		1: {"Id": "1", "Name": `"Anvil"`, "Active": "true", "Stock": "9007199254740993", "Price": "12.5", "Ref": `"00000000-0000-0000-0000-000000000001"`, "Released": `"2026-01-15"`, "Updated": `"2026-01-15T10:00:00Z"`},
		2: {"Id": "2", "Name": `"Bolt"`, "Active": "false", "Stock": "0", "Price": "0.25", "Ref": `"00000000-0000-0000-0000-000000000002"`, "Released": `"2025-12-31"`, "Updated": `"2025-12-31T23:59:59.5Z"`},
		3: {"Id": "3", "Name": `"Chain"`, "Active": "true", "Stock": "40", "Price": "100", "Ref": `"00000000-0000-0000-0000-000000000003"`, "Released": `"2026-03-01"`, "Updated": `"2026-03-01T08:30:00Z"`},
		4: {"Id": "4", "Name": "null", "Active": "true", "Stock": "-5", "Price": "-3.75", "Ref": `"00000000-0000-0000-0000-000000000004"`, "Released": `"2024-02-29"`, "Updated": `"2024-02-29T00:00:00Z"`},
		5: {"Id": "5", "Name": `"Drill"`, "Active": "null", "Stock": "7", "Price": "null", "Ref": `"00000000-0000-0000-0000-000000000005"`, "Released": `"2026-06-30"`, "Updated": `"2026-06-30T12:00:00.25Z"`},
	}
}

// abU setzt W in einen Text der Fixtures ein.
func abU(s string) string { return strings.ReplaceAll(s, "{W}", abW) }

// abPmit ist P(n) mit den Werten aus ueber an Stelle der Ausgangswerte: die Mitglieder in der Folge des Modells.
func abPmit(n int, ueber map[string]string) string {
	var teile []string
	for _, m := range abMitglieder() {
		w := abWerte()[n][m]
		if v, da := ueber[m]; da {
			w = v
		}
		teile = append(teile, `"`+m+`":`+w)
	}
	return strings.Join(teile, ",")
}

// abP ist P(n).
func abP(n int) string { return abPmit(n, nil) }

// abEL ist EL(n), das Element einer Collection.
func abEL(n int) string {
	return `{"@odata.id":"{W}/Products(` + strconv.Itoa(n) + `)",` + abP(n) + `}`
}

// abEN ist EN(n), der Body einer Entity.
func abEN(n int) string { return abENmit(n, nil) }

func abENmit(n int, ueber map[string]string) string {
	return abU(`{"@odata.context":"{W}/$metadata#Products/$entity","@odata.id":"{W}/Products(` + strconv.Itoa(n) + `)",` + abPmit(n, ueber) + `}`)
}

// abK ist K(…), der Body einer Collection.
func abK(elemente ...string) string {
	return abU(`{"@odata.context":"{W}/$metadata#Products","value":[` + abU(strings.Join(elemente, ",")) + `]}`)
}

// abAuswahl ist der Body einer Collection mit $select=sel und den Entities ids in dieser Folge: die Mitglieder der Auswahl in der Folge des Modells.
func abAuswahl(sel string, ids ...int) string {
	waehlt := map[string]bool{}
	for _, m := range strings.Split(sel, ",") {
		waehlt[m] = true
	}
	var el []string
	for _, n := range ids {
		el = append(el, `{"@odata.id":"{W}/Products(`+strconv.Itoa(n)+`)"`+abAuswahlMitglieder(n, waehlt)+`}`)
	}
	return abU(`{"@odata.context":"{W}/$metadata#Products(` + sel + `)","value":[` + strings.Join(el, ",") + `]}`)
}

// abAuswahlMitglieder sind die gewählten Mitglieder von Produkt n; das Produkt 10 (CRUD-01) trägt nur seinen Schlüssel.
func abAuswahlMitglieder(n int, waehlt map[string]bool) string {
	werte := abWerte()[n]
	if n == 10 {
		werte = map[string]string{"Id": "10"}
	}
	var s string
	for _, m := range abMitglieder() {
		if waehlt[m] {
			s += `,"` + m + `":` + werte[m]
		}
	}
	return s
}

// abPayload ist der Payload eines Produkts der Baseline in der Kanonform von SPEC-127 Nr. 14: die Mitglieder aufsteigend nach den Bytes
// ihrer Namen, Id als Zahl, Stock und Price als String, null als null.
func abPayload(n int, ueber map[string]string) string {
	namen := abMitglieder()
	sort.Strings(namen)
	var teile []string
	for _, m := range namen {
		w := abWerte()[n][m]
		if v, da := ueber[m]; da {
			w = v
		}
		if (m == "Stock" || m == "Price") && w != "null" && !strings.HasPrefix(w, `"`) {
			w = `"` + w + `"`
		}
		teile = append(teile, `"`+m+`":`+w)
	}
	return "{" + strings.Join(teile, ",") + "}"
}

// abAlleProperties sind die acht Namen von Product aufsteigend.
func abAlleProperties() []string {
	l := abMitglieder()
	sort.Strings(l)
	return l
}

// abAusgang ist der Entity-State der Baseline bsl-1 für den Workspace id: fünf Produkte, ein Eintrag in Orders (partial), Labels leer.
func abAusgang(id string) []model.EntityState {
	es := []model.EntityState{{WorkspaceID: id, EntitySet: "Orders", EntityKey: "1", Payload: `{"Id":1}`, KnownProperties: []string{"Id"}, Completeness: model.EntityPartial}}
	for n := 1; n <= 5; n++ {
		es = append(es, model.EntityState{WorkspaceID: id, EntitySet: "Products", EntityKey: strconv.Itoa(n), Payload: abPayload(n, nil),
			KnownProperties: abAlleProperties(), Completeness: model.EntityComplete})
	}
	return es
}

// abManifest ist das Manifest der Baseline bsl-1 (SPEC-127 Nr. 14, so weit SPEC-129 Nr. 6 es liest).
func abManifest() []byte {
	var l []any
	for _, e := range abAusgang("") {
		l = append(l, map[string]any{"completeness": string(e.Completeness), "entity_set": e.EntitySet, "payload": json.RawMessage(e.Payload),
			"provenance": []any{}, "selected_properties": e.KnownProperties, "typed_key": e.EntityKey})
	}
	b, err := json.Marshal(map[string]any{"baseline_snapshot_id": "bsl-1", "coverage": []any{}, "created_at": "2026-10-20T10:00:00.000000000Z",
		"entities": l, "entity_count": len(l), "import_id": "imp-bsl-1", "manifest_version": 1, "metadata_hash": "sha256:" + strings.Repeat("0", 64),
		"provenance": []any{}, "service_id": "svc", "snapshot_kind": "baseline", "source_session_ids": []string{"ses-1"}})
	if err != nil {
		panic(err)
	}
	return b
}

// abUmgebung ist der Aufbau: eine Datei, eine Ablage, die Dienste darüber und der Prüfling (lax und strikt).
type abUmgebung struct {
	t        *testing.T
	dbPfad   string
	blobPfad string
	db       *sqlite.DB
	store    *blobs.Store
	waechter *maintenance.BlobWaechter
	ws       *workspace.Verwaltung
	snap     *snapshot.Dienst
	uhr      *supporttest.FestUhr
	ids      *supporttest.FestIDs
	mock     *mock.Mocker
	strikt   *mock.Mocker
	zu       bool
}

func abMust(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// abOeffne öffnet Datei und Ablage und baut Wächter, Wartung, beide Dienste und den Prüfling; es legt keine Daten an (Aufbaufehler sind Fatal,
// SPEC-049 Nr. 9 a).
func abOeffne(t *testing.T, dbPfad, blobPfad string) *abUmgebung {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, dbPfad, sqlite.Options{})
	abMust(t, err)
	store, err := blobs.New(blobPfad)
	abMust(t, err)
	u := &abUmgebung{t: t, dbPfad: dbPfad, blobPfad: blobPfad, db: db, store: store, uhr: supporttest.NewFestUhr(wsT0()), ids: supporttest.NewFestIDs()}
	t.Cleanup(func() { u.schliesse() })
	u.waechter, err = maintenance.NewBlobWaechter(store, u.uhr, time.Nanosecond)
	abMust(t, err)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	audit := supporttest.NewFakeAudit()
	wartung, err := maintenance.New(maintenance.Konfig{Sessions: db.Sessions(), Dienste: db.Metadata(), Speicher: db.Maintenance(),
		Blobs: u.waechter, Uhr: u.uhr, Log: log, Audit: audit})
	abMust(t, err)
	u.ws, err = workspace.New(workspace.Konfig{Workspaces: db.Workspaces(), Transaktion: db, Snapshots: db.Snapshots(), Blobs: u.waechter,
		Dienste: db.Metadata(), Wartung: wartung, Uhr: u.uhr, IDs: u.ids, Log: log, Audit: audit})
	abMust(t, err)
	u.snap, err = snapshot.New(snapshot.Konfig{Workspaces: db.Workspaces(), Entities: db.Entities(), Snapshots: db.Snapshots(), Blobs: u.waechter,
		Uhr: u.uhr, IDs: u.ids})
	abMust(t, err)
	u.mock, u.strikt = u.baueMock(nil, nil)
	return u
}

// schliesse schließt die Datenbank genau einmal.
func (u *abUmgebung) schliesse() {
	if !u.zu {
		u.zu = true
		_ = u.db.Close()
	}
}

// baueMock baut den Prüfling lax und strikt über derselben Datei; wsDek und entDek umhüllen die Repositories der Konfig (Haltepunkt, SPEC-049 Nr. 8).
func (u *abUmgebung) baueMock(wsDek func(outbound.WorkspaceRepository) outbound.WorkspaceRepository, entDek func(outbound.EntityRepository) outbound.EntityRepository) (*mock.Mocker, *mock.Mocker) {
	u.t.Helper()
	var wr outbound.WorkspaceRepository = u.db.Workspaces()
	var er outbound.EntityRepository = u.db.Entities()
	if wsDek != nil {
		wr = wsDek(wr)
	}
	if entDek != nil {
		er = entDek(er)
	}
	bau := func(strikt bool) *mock.Mocker {
		m, err := mock.New(mock.Konfig{
			Dienste: []mock.Dienst{{ID: "svc", Root: "/odata"}}, Metadaten: u.db.Metadata(), Workspaces: wr, Entities: er, Coverage: u.db.Coverage(),
			Blobs: u.store, MetadatenCodec: csdl.New(), URLCodec: urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())),
			Schreiber: jsoncodec.NewSchreiber(), Eingabe: jsoncodec.NewEingabe(), Transaktion: u.db, PublicBaseURL: "http://recorder.test",
			StrictCoverage: strikt, MaxBytes: 8388608,
		})
		abMust(u.t, err)
		return m
	}
	return bau(false), bau(true)
}

// blob legt b in der Ablage ab, am Wächter vorbei, und liefert die Referenz.
func (u *abUmgebung) blob(b []byte) string {
	u.t.Helper()
	w, err := u.store.Begin(context.Background(), int64(len(b))+1)
	abMust(u.t, err)
	_, err = w.Write(b)
	abMust(u.t, err)
	ref, _, err := w.Commit()
	abMust(u.t, err)
	return ref
}

// abFrisch ist der Ausgangszustand von SPEC-049 Nr. 2 in einer frischen Datei und Ablage: Dienst svc mit den Metadaten, die Baseline bsl-1
// mit der Coverage, der Workspace ws-1 (Revision 0). Kein Zustand einer vorigen ID, kein Reset als Mittel des Aufbaus (Nr. 3 b).
func abFrisch(t *testing.T) *abUmgebung {
	t.Helper()
	u := abOeffne(t, filepath.Join(t.TempDir(), "r.db"), t.TempDir())
	u.ausgangsdaten()
	abMust(t, func() error { _, err := u.ws.Create(context.Background(), "bsl-1", "ws-1"); return err }())
	return u
}

// ausgangsdaten legt Dienst, Session, Metadaten und die Baseline an.
func (u *abUmgebung) ausgangsdaten() {
	u.t.Helper()
	ctx := context.Background()
	abMust(u.t, u.db.Metadata().Create(ctx, model.Service{ID: "svc", ServiceRoot: "/", ODataVersion: "4.0"}))
	abMust(u.t, u.db.Sessions().Create(ctx, model.Session{ID: "ses-1", ServiceID: "svc", CreatedAt: wsT0(), Status: model.SessionOpen}))
	abMust(u.t, u.db.Sessions().SetStatus(ctx, "ses-1", model.SessionComplete))
	csdlRef := u.blob([]byte(abCSDL))
	abMust(u.t, u.db.Metadata().SetMetadata(ctx, "svc", csdlRef, csdlRef))
	ref := u.blob(abManifest())
	_, neu, err := u.db.Imports().Publish(ctx, model.Publication{
		Import:           model.Import{ID: "imp-bsl-1", RequestHash: "req-bsl-1", BaselineSnapshotID: "bsl-1", ServiceID: "svc", CreatedAt: wsT0(), ReportJSON: "{}"},
		Snapshot:         wsBaselineSnapshot("bsl-1", ref),
		SourceSessionIDs: []string{"ses-1"},
		Coverage: []model.Coverage{{EntitySet: "Products", Status: model.CoverageComplete, EvidenceJSON: "{}"},
			{EntitySet: "Labels", Status: model.CoverageComplete, EvidenceJSON: "{}"}, {EntitySet: "Orders", Status: model.CoveragePartial, EvidenceJSON: "{}"}},
	})
	if err != nil || !neu {
		u.t.Fatalf("Publish: %v, neu=%v", err, neu)
	}
}

// neuerWorkspace legt den Workspace id aus bsl-1 an.
func (u *abUmgebung) neuerWorkspace(id string) {
	u.t.Helper()
	_, err := u.ws.Create(context.Background(), "bsl-1", id)
	abMust(u.t, err)
}

// cursor schreibt den Cursor rep-1/ses-1 des Workspace id auf Position pos mit consumed 1 bis pos (Service svc, Strategie sequential).
func (u *abUmgebung) cursor(id string, pos int64) {
	u.t.Helper()
	ctx := context.Background()
	var verbraucht []int64
	for i := int64(1); i <= pos; i++ {
		verbraucht = append(verbraucht, i)
	}
	abMust(u.t, u.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		return s.Cursors.Put(ctx, model.ReplayCursor{WorkspaceID: id, ReplaySessionID: "rep-1", RecordingSessionID: "ses-1", ServiceID: "svc",
			Strategy: model.StrategySequential, Position: pos, Consumed: verbraucht})
	}))
}

// neustart schließt Datenbank und gibt die Instanzen frei, öffnet dieselbe Datei und dasselbe Verzeichnis der Ablage neu und baut Wächter, Wartung,
// beide Dienste und den Prüfling neu (SPEC-049 Nr. 7 a); keine Instanz davor wird weiterbenutzt.
func (u *abUmgebung) neustart() *abUmgebung {
	u.t.Helper()
	u.schliesse()
	return abOeffne(u.t, u.dbPfad, u.blobPfad)
}

// abAnfrage ist ein Schritt: Methode, Pfad nach /odata, Query so, wie sie nach dem ? steht, weitere Header und Body.
type abAnfrage struct {
	Methode, Pfad, Query, Body string
	Header                     map[string][]string
}

// anfrage baut die inbound.Anfrage: Workspace ws, Content-Type bei einem Body, weitere Header.
func (a abAnfrage) anfrage(ws string) inbound.Anfrage {
	h := map[string][]string{"X-Recorder-Workspace": {ws}}
	if a.Body != "" {
		h["Content-Type"] = []string{"application/json"}
	}
	for k, v := range a.Header {
		h[k] = v
	}
	r := inbound.Anfrage{Methode: a.Methode, Pfad: "/odata" + a.Pfad, RawQuery: a.Query, Header: h, Laenge: -1}
	if a.Body != "" {
		r.Body, r.Laenge = strings.NewReader(a.Body), int64(len(a.Body))
	}
	return r
}

// abIst ist, was ein Schritt ergab: eine Antwort (Status, Header mit kanonischen Namen, Body) oder ein Fehler mit Code, dazu die Nullwert-Prüfung.
type abIst struct {
	Status  int
	Header  map[string][]string
	Body    string
	F       *inbound.Fehler
	Err     error
	Antwort inbound.Antwort
}

// abSende ruft den Prüfling und liest den Body bis zum Ende und schließt ihn (SPEC-049 Nr. 3 c). Ein Lesefehler des Bodys ist ein Fehler des Schritts.
func abSende(ctx context.Context, m *mock.Mocker, ws string, a abAnfrage) abIst {
	ant, err := m.Mock(ctx, a.anfrage(ws))
	if err != nil {
		var f *inbound.Fehler
		r := abIst{Err: err, Antwort: ant}
		if errors.As(err, &f) {
			r.F = f
		}
		return r
	}
	r := abIst{Status: ant.Status, Header: map[string][]string{}, Antwort: ant}
	for k, v := range ant.Header {
		n := http.CanonicalHeaderKey(k)
		r.Header[n] = append(r.Header[n], v...)
	}
	if ant.Body != nil {
		b, rerr := io.ReadAll(ant.Body)
		if rerr != nil {
			r.Err = rerr
		}
		if cerr := ant.Body.Close(); cerr != nil && r.Err == nil {
			r.Err = cerr
		}
		r.Body = string(b)
	}
	return r
}

// sende ruft den lax konfigurierten Prüfling für ws-1.
func (u *abUmgebung) sende(a abAnfrage) abIst {
	return abSende(context.Background(), u.mock, "ws-1", a)
}

// sendeStrikt ruft den Prüfling mit StrictCoverage für ws-1.
func (u *abUmgebung) sendeStrikt(a abAnfrage) abIst {
	return abSende(context.Background(), u.strikt, "ws-1", a)
}

// abStand ist die Liste aller Einträge der drei Sets eines Workspace, seine Cursors und die Revision (SPEC-049 Nr. 5 c).
type abStand struct {
	Eintraege []model.EntityState
	Cursors   []model.ReplayCursor
	Revision  int64
}

// stand liest den Stand des Workspace id über die Ports der Ablage (nur lesend).
func (u *abUmgebung) stand(id string) abStand {
	u.t.Helper()
	s, err := abLeseStand(context.Background(), u.db, id)
	abMust(u.t, err)
	return s
}

func abLeseStand(ctx context.Context, db *sqlite.DB, id string) (abStand, error) {
	var s abStand
	for _, set := range []string{"Labels", "Orders", "Products"} {
		es, err := db.Entities().ListSet(ctx, id, set)
		if err != nil {
			return s, err
		}
		sort.Slice(es, func(i, j int) bool { return es[i].EntityKey < es[j].EntityKey })
		for _, e := range es {
			e.KnownProperties = append([]string(nil), e.KnownProperties...)
			sort.Strings(e.KnownProperties)
			s.Eintraege = append(s.Eintraege, e)
		}
	}
	cs, err := db.Cursors().List(ctx, id)
	if err != nil {
		return s, err
	}
	for _, c := range cs {
		if len(c.Consumed) == 0 {
			c.Consumed = nil
		}
		s.Cursors = append(s.Cursors, c)
	}
	sort.Slice(s.Cursors, func(i, j int) bool {
		a, b := s.Cursors[i], s.Cursors[j]
		return a.ReplaySessionID < b.ReplaySessionID || a.ReplaySessionID == b.ReplaySessionID && a.RecordingSessionID < b.RecordingSessionID
	})
	w, err := db.Workspaces().Get(ctx, id)
	if err != nil {
		return s, err
	}
	s.Revision = w.Revision
	return s, nil
}

// abStandGleich nennt die Unterschiede zweier Stände Feld für Feld (leer: gleich).
func abStandGleich(a, b abStand) []string {
	var d []string
	if a.Revision != b.Revision {
		d = append(d, fmt.Sprintf("Revision %d gegen %d", a.Revision, b.Revision))
	}
	if len(a.Eintraege) != len(b.Eintraege) {
		return append(d, fmt.Sprintf("Zahl der Einträge %d gegen %d", len(a.Eintraege), len(b.Eintraege)))
	}
	for i := range a.Eintraege {
		if !reflect.DeepEqual(a.Eintraege[i], b.Eintraege[i]) {
			d = append(d, fmt.Sprintf("Eintrag %s/%s: %+v gegen %+v", a.Eintraege[i].EntitySet, a.Eintraege[i].EntityKey, a.Eintraege[i], b.Eintraege[i]))
		}
	}
	if !reflect.DeepEqual(a.Cursors, b.Cursors) {
		d = append(d, fmt.Sprintf("Cursors %+v gegen %+v", a.Cursors, b.Cursors))
	}
	return d
}

// abAusgangsstand ist der Stand von Workspace id im Ausgangszustand: Version 0, nichts gelöscht, Payload wie die Baseline, ohne Cursors.
func abAusgangsstand(id string, revision int64) abStand {
	es := abAusgang(id)
	sort.Slice(es, func(i, j int) bool {
		if es[i].EntitySet != es[j].EntitySet {
			return es[i].EntitySet < es[j].EntitySet
		}
		return es[i].EntityKey < es[j].EntityKey
	})
	return abStand{Eintraege: es, Revision: revision}
}

// abParallel führt arbeit(k) für k = 0 bis n-1 je auf einer Goroutine aus; alle n Goroutinen warten auf ein gemeinsames Startsignal, das erst
// fällt, wenn alle am Tor angekommen sind (SPEC-049 Nr. 6 a). Gestartet und Beendet zählen die Goroutinen. Die Goroutinen werden nacheinander
// erzeugt, jede erst, wenn die vorige am Tor angekommen ist.
type abParallel struct {
	N         int
	Bereit    atomic.Int32
	Gefallen  atomic.Bool
	Gestartet atomic.Int32
	Beendet   atomic.Int32
}

// lauf führt arbeit aus und kehrt zurück, wenn alle n Goroutinen beendet sind. arbeit sieht über p den Stand von Bereit und Gefallen.
func (p *abParallel) lauf(arbeit func(k int)) {
	start := make(chan struct{})
	fertig := make(chan struct{}, p.N)
	for k := 0; k < p.N; k++ {
		go func() {
			p.Gestartet.Add(1)
			p.Bereit.Add(1)
			<-start
			arbeit(k)
			p.Beendet.Add(1)
			fertig <- struct{}{}
		}()
		for int(p.Bereit.Load()) <= k {
			time.Sleep(10 * time.Microsecond)
		}
	}
	p.Gefallen.Store(true)
	close(start)
	for i := 0; i < p.N; i++ {
		<-fertig
	}
}

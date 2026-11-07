package snapshot_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
)

func t0() time.Time { return time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC) }

// metadatenHash ist der MetadataHash der Baseline der Fixture; er ist von jedem anderen Wert verschieden.
func metadatenHash() string { return "sha256:" + strings.Repeat("b", 64) }

// welt ist ein Dienst mit allen Fakes; der Workspace "ws-1" (Service "svc", Revision 5) trägt den State standard("ws-1"),
// "ws-2" einen leeren; die Baseline "bsl-1" trägt metadatenHash().
type welt struct {
	t       *testing.T
	p       *prot
	ws      *wsFake
	ent     *entFake
	snaps   *snapFake
	blobs   *blobFake
	festUhr *supporttest.FestUhr
	festIDs *supporttest.FestIDs
	konfig  snapshot.Konfig
	d       *snapshot.Dienst
}

func neu(t *testing.T) *welt {
	t.Helper()
	p := neuesProt(t)
	w := &welt{t: t, p: p, festUhr: supporttest.NewFestUhr(t0()), festIDs: supporttest.NewFestIDs()}
	t.Cleanup(func() {
		if n := p.verboten.Load(); n != 0 {
			t.Errorf("%d Operationen gerufen, die der Dienst nie rufen darf", n)
		}
	})
	w.ws = &wsFake{p: p, ws: map[string]model.Workspace{
		"ws-1": {ID: "ws-1", ServiceID: "svc", BaselineSnapshotID: "bsl-1", Revision: 5, CreatedAt: t0()},
		"ws-2": {ID: "ws-2", ServiceID: "svc", BaselineSnapshotID: "bsl-1", Revision: 0, CreatedAt: t0()},
	}}
	w.ent = &entFake{p: p, ent: map[string][]model.EntityState{"ws-1": standard("ws-1"), "ws-2": {}}}
	ref := "sha256:" + strings.Repeat("a", 64)
	w.snaps = &snapFake{p: p, ws: w.ws, baselines: map[string]model.BaselineSnapshot{
		"bsl-1": {ID: "bsl-1", ServiceID: "svc", CreatedAt: t0(), StateBlob: ref, StateHash: ref, MetadataHash: metadatenHash()},
	}}
	w.blobs = &blobFake{p: p, inhalt: map[string][]byte{}}
	w.konfig = snapshot.Konfig{Workspaces: w.ws, Entities: w.ent, Snapshots: w.snaps, Blobs: w.blobs,
		Uhr: zaehlUhr{p: p, u: w.festUhr}, IDs: zaehlIDs{p: p, g: w.festIDs}}
	d, err := snapshot.New(w.konfig)
	if err != nil {
		t.Fatal(err)
	}
	w.d = d
	return w
}

// standard ist der State der Fixture: drei Einträge in zwei Mengen, nicht nach Menge und Schlüssel geordnet, ein
// Tombstone, drei Versionen.
func standard(wid string) []model.EntityState {
	return []model.EntityState{
		{WorkspaceID: wid, EntitySet: "Products", EntityKey: "Id=2", Payload: `{"Id":"2"}`, KnownProperties: []string{"Id"},
			Completeness: model.EntityPartial, Version: 1, Deleted: true},
		{WorkspaceID: wid, EntitySet: "Products", EntityKey: "Id=1", Payload: `{"Id":"1","Name":"Milk"}`, KnownProperties: []string{"Name", "Id"},
			Completeness: model.EntityComplete, Version: 3},
		{WorkspaceID: wid, EntitySet: "Categories", EntityKey: "Id=1", Payload: `{"Id":"1","Title":"Food"}`, KnownProperties: []string{"Id", "Title"},
			Completeness: model.EntityComplete, Version: 0},
	}
}

// manifestBytes liefert die Bytes des Blobs mit der Referenz ref.
func (w *welt) manifestBytes(ref string) []byte {
	w.blobs.mu.Lock()
	defer w.blobs.mu.Unlock()
	return w.blobs.inhalt[ref]
}

// erzeuge ruft Create für ws-1 unter dem Namen name und liefert die Bytes des Manifests.
func (w *welt) erzeuge(name string) (model.WorkspaceSnapshot, []byte, error) {
	s, err := w.d.Create(context.Background(), "ws-1", name)
	if err != nil {
		return s, nil, err
	}
	return s, w.manifestBytes(s.StateBlob), nil
}

// ctxIdentitaet ist ein ctx, dessen Identität ein Fake festhalten kann; ein abgeleiteter Kontext ist ein anderer Wert.
type ctxIdentitaet struct{ innen ctxMethoden }

func (c *ctxIdentitaet) Deadline() (time.Time, bool) { return c.innen.Deadline() }
func (c *ctxIdentitaet) Done() <-chan struct{}       { return c.innen.Done() }
func (c *ctxIdentitaet) Err() error                  { return c.innen.Err() }
func (c *ctxIdentitaet) Value(key any) any           { return c.innen.Value(key) }

type ctxSchluessel struct{}

func neuerCtx() *ctxIdentitaet {
	return &ctxIdentitaet{context.WithValue(context.Background(), ctxSchluessel{}, "wert")}
}

func codeVon(err error) (*inbound.Fehler, bool) {
	var f *inbound.Fehler
	return f, errors.As(err, &f)
}

// pruefeCode: err ist ein *inbound.Fehler mit dem Code c, dem einen Status des Codes und, wenn u gesetzt ist, der Ursache u.
func pruefeCode(t *testing.T, err error, c string, u error) {
	t.Helper()
	f, ok := codeVon(err)
	if !ok {
		t.Errorf("kein *Fehler mit Code: %v", err)
		return
	}
	if f.Code.String() != c || f.Status != f.Code.Stati()[0] {
		t.Errorf("Code %v Status %d, erwartet %s", f.Code, f.Status, c)
	}
	if u != nil && !errors.Is(err, u) {
		t.Errorf("Ursache %v nicht erreichbar in %v", u, err)
	}
}

// pruefeOhneCode: err trägt ziel und keinen *inbound.Fehler.
func pruefeOhneCode(t *testing.T, err, ziel error) {
	t.Helper()
	if err == nil || !errors.Is(err, ziel) {
		t.Errorf("Fehler %v, erwartet %v", err, ziel)
		return
	}
	if _, ok := codeVon(err); ok {
		t.Errorf("%v trägt einen Code", err)
	}
}

var errPlatte = errors.New("Platte voll")

// gleichzeitig startet n Funktionen über einen gemeinsamen Kanal zugleich und wartet auf alle.
func gleichzeitig(n int, f func(i int)) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			f(i)
		}()
	}
	close(start)
	wg.Wait()
}

func esc(rest string) string { return "\\" + rest }

// tabellenFehler nennt jeden Fehler der Matrix M7: acht Fehler, davon einer ohne Zeile in der Tabelle.
type tabellenFehler struct {
	name string
	err  error
}

func m7Fehler() []tabellenFehler {
	return []tabellenFehler{{"ErrNichtGefunden", outbound.ErrNichtGefunden}, {"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden},
		{"ErrFremderService", outbound.ErrFremderService}, {"ErrFalscheArt", outbound.ErrFalscheArt}, {"ErrBlobFehlt", outbound.ErrBlobFehlt},
		{"ErrBeschaedigt", outbound.ErrBeschaedigt}, {"ErrUngueltig", model.ErrUngueltig}, {"roh", errPlatte}}
}

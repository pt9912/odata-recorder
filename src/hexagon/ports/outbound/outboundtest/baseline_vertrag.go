package outboundtest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// RunBaselineVertrag prüft SPEC-107 Nr. 4 bis 12 für SnapshotRepository, ImportRepository
// und CoverageRepository.
func RunBaselineVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Publish schreibt alles und die Leser lesen es", publikationFall1},
		{"ungültige Publikationen werden abgelehnt und schreiben nichts", publikationFall2},
		{"Publish ist atomar", publikationFall3},
		{"gleicher Anfrage-Hash ist idempotent", publikationFall4},
		{"gleichzeitige Publish mit gleichem Hash ordnen sich", publikationNebenlaeufig},
		{"Prüfreihenfolge von Publish", publikationFall5},
		{"Baselines sind unveränderlich", publikationFall6},
		{"Baseline-Quellen: BaselinesOf, Sources, Löschen der Session", provenienzFall},
		{"nur Baselines sind sichtbar, geordnet nach Zeit", snapshotFall},
		{"Coverage lesen, Evidenz byte-gleich", coverageFall},
		{"abgebrochener Kontext beendet jede Operation ohne Änderung", baselineKontextFall},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

// publikationBasis legt die Services d und e an, die Sessions s1 und s2 in d und x1 in e.
func publikationBasis(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	v := neu(t)
	ctx := context.Background()
	for _, id := range []string{"d", "e"} {
		v.FuegeDienstHinzu(t, id)
	}
	for i, s := range [][2]string{{"s1", "d"}, {"s2", "d"}, {"x1", "e"}} {
		ok(t, v.Sessions.Create(ctx, session(s[0], s[1], i)))
	}
	return v
}

func publikation(dienst, n string, quellen ...string) model.Publication {
	return model.Publication{
		Import: model.Import{ID: "imp-" + n, RequestHash: "req-" + n, BaselineSnapshotID: "snap-" + n, ServiceID: dienst,
			CreatedAt: t0(), ReportJSON: `{"ok":true}`},
		Snapshot: model.BaselineSnapshot{ID: "snap-" + n, ServiceID: dienst, CreatedAt: t0(), StateBlob: ref("a"),
			StateHash: "sh-" + n, MetadataHash: "mh"},
		SourceSessionIDs: quellen,
		Coverage:         []model.Coverage{{EntitySet: "Products", Status: model.CoverageComplete, EvidenceJSON: `{"pages":3}`}},
	}
}

func veroeffentliche(t *testing.T, v Verbindung, p model.Publication) {
	t.Helper()
	_, neu, err := v.Imports.Publish(context.Background(), p)
	ok(t, err)
	if !neu {
		t.Fatalf("Publish von %s war nicht neu", p.Import.ID)
	}
}

// nichtsGeschrieben prüft, dass die Publikation in keiner der vier Tabellen eine Zeile hinterließ.
func nichtsGeschrieben(t *testing.T, v Verbindung, p model.Publication) {
	t.Helper()
	ctx := context.Background()
	_, err := v.Snapshots.Get(ctx, p.Snapshot.ID)
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Imports.Get(ctx, p.Import.ID)
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Coverage.List(ctx, p.Snapshot.ID)
	erwarte(t, err, outbound.ErrNichtGefunden)
	for _, q := range p.SourceSessionIDs {
		if l, err := v.Imports.BaselinesOf(ctx, q); err == nil && slices.Contains(l, p.Snapshot.ID) {
			t.Fatalf("Quelle %s trägt die Baseline %s", q, p.Snapshot.ID)
		}
	}
}

func publikationFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	p := publikation("d", "1", "s1", "s2")
	p.Import.CreatedAt = t0().In(time.FixedZone("p5", 5*3600))
	p.Coverage = append(p.Coverage, model.Coverage{EntitySet: "Orders", Status: model.CoveragePartial, EvidenceJSON: `{}`})
	imp, istNeu, err := v.Imports.Publish(ctx, p)
	ok(t, err)
	if !istNeu || imp.ID != "imp-1" {
		t.Fatalf("Publish lieferte %+v, neu=%v", imp, istNeu)
	}
	pruefeLesen(t, v, p)
	leer := publikation("d", "leer", "s1")
	leer.Coverage = nil
	veroeffentliche(t, v, leer)
}

func pruefeLesen(t *testing.T, v Verbindung, p model.Publication) {
	t.Helper()
	ctx := context.Background()
	b, err := v.Snapshots.Get(ctx, "snap-1")
	ok(t, err)
	zeitGleich := b.CreatedAt.Equal(t0())
	b.CreatedAt = t0()
	if want := (model.BaselineSnapshot{ID: "snap-1", ServiceID: "d", CreatedAt: t0(), StateBlob: ref("a"), StateHash: "sh-1", MetadataHash: "mh"}); !zeitGleich || b != want {
		t.Fatalf("Baseline %+v", b)
	}
	i, err := v.Imports.Get(ctx, "imp-1")
	ok(t, err)
	zeitGleich = i.CreatedAt.Equal(t0())
	i.CreatedAt = t0()
	if want := (model.Import{ID: "imp-1", RequestHash: "req-1", BaselineSnapshotID: "snap-1", ServiceID: "d", CreatedAt: t0(), ReportJSON: `{"ok":true}`}); !zeitGleich || i != want {
		t.Fatalf("Import %+v", i)
	}
	if q, err := v.Imports.Sources(ctx, "snap-1"); err != nil || !reflect.DeepEqual(q, []string{"s1", "s2"}) {
		t.Fatalf("Quellen %v %v", q, err)
	}
	if c, err := v.Coverage.List(ctx, "snap-1"); err != nil || !reflect.DeepEqual(c, []model.Coverage{p.Coverage[1], p.Coverage[0]}) {
		t.Fatalf("Coverage %v %v", c, err)
	}
}

func publikationFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	cov := func(p *model.Publication, c model.Coverage) { p.Coverage = []model.Coverage{c} }
	fehlerhaft := map[string]func(*model.Publication){
		"Import ohne Kennung":                         func(p *model.Publication) { p.Import.ID = "" },
		"Snapshot ohne Kennung":                       func(p *model.Publication) { p.Snapshot.ID, p.Import.BaselineSnapshotID = "", "" },
		"ohne Service":                                func(p *model.Publication) { p.Snapshot.ServiceID, p.Import.ServiceID = "", "" },
		"Service von Import und Snapshot verschieden": func(p *model.Publication) { p.Import.ServiceID = "e" },
		"Import verweist auf andere Baseline":         func(p *model.Publication) { p.Import.BaselineSnapshotID = "andere" },
		"ohne request_hash":                           func(p *model.Publication) { p.Import.RequestHash = "" },
		"ohne state_hash":                             func(p *model.Publication) { p.Snapshot.StateHash = "" },
		"ohne metadata_hash":                          func(p *model.Publication) { p.Snapshot.MetadataHash = "" },
		"state_blob leer":                             func(p *model.Publication) { p.Snapshot.StateBlob = "" },
		"state_blob ungültig":                         func(p *model.Publication) { p.Snapshot.StateBlob = "sha256:abc" },
		"state_blob Großbuchstaben":                   func(p *model.Publication) { p.Snapshot.StateBlob = ref("F") },
		"report_json leer":                            func(p *model.Publication) { p.Import.ReportJSON = "" },
		"report_json kein JSON":                       func(p *model.Publication) { p.Import.ReportJSON = "{" },
		"keine Quell-Session":                         func(p *model.Publication) { p.SourceSessionIDs = nil },
		"leere Quellenliste":                          func(p *model.Publication) { p.SourceSessionIDs = []string{} },
		"doppelte Quell-Session":                      func(p *model.Publication) { p.SourceSessionIDs = []string{"s1", "s2", "s1"} },
		"Coverage ohne entity_set":                    func(p *model.Publication) { cov(p, model.Coverage{Status: model.CoverageComplete, EvidenceJSON: `{}`}) },
		"Coverage-Status unbekannt": func(p *model.Publication) {
			cov(p, model.Coverage{EntitySet: "P", Status: "voll", EvidenceJSON: `{}`})
		},
		"Coverage-Status leer": func(p *model.Publication) { cov(p, model.Coverage{EntitySet: "P", EvidenceJSON: `{}`}) },
		"Evidenz kein JSON": func(p *model.Publication) {
			cov(p, model.Coverage{EntitySet: "P", Status: model.CoverageUnknown, EvidenceJSON: "x"})
		},
		"Evidenz leer": func(p *model.Publication) { cov(p, model.Coverage{EntitySet: "P", Status: model.CoverageUnknown}) },
		"Evidenz kein UTF-8": func(p *model.Publication) {
			cov(p, model.Coverage{EntitySet: "P", Status: model.CoverageUnknown, EvidenceJSON: "\"\xff\""})
		},
		"report_json kein UTF-8": func(p *model.Publication) { p.Import.ReportJSON = "\"\xff\"" },
		"doppeltes entity_set": func(p *model.Publication) {
			c := model.Coverage{EntitySet: "P", Status: model.CoveragePartial, EvidenceJSON: `{}`}
			p.Coverage = []model.Coverage{c, c}
		},
		"Import ohne Zeit":      func(p *model.Publication) { p.Import.CreatedAt = time.Time{} },
		"Import nach Jahr 9999": func(p *model.Publication) { p.Import.CreatedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		"Snapshot ohne Zeit":    func(p *model.Publication) { p.Snapshot.CreatedAt = time.Time{} },
		"Snapshot vor Jahr 1":   func(p *model.Publication) { p.Snapshot.CreatedAt = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) },
	}
	for name, mut := range fehlerhaft {
		p := publikation("d", "x", "s1", "s2")
		mut(&p)
		if _, _, err := v.Imports.Publish(ctx, p); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	nichtsGeschrieben(t, v, publikation("d", "x", "s1"))
	if l, err := v.Snapshots.List(ctx, "d"); err != nil || len(l) != 0 {
		t.Fatalf("ungültige Publikationen wurden gespeichert: %v %v", l, err)
	}
}

func publikationFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	unbekannt := publikation("d", "1", "s1", "gibt-es-nicht")
	_, _, err := v.Imports.Publish(ctx, unbekannt)
	erwarte(t, err, outbound.ErrNichtGefunden)
	nichtsGeschrieben(t, v, unbekannt)
	fremd := publikation("d", "2", "s1", "x1")
	_, _, err = v.Imports.Publish(ctx, fremd)
	erwarte(t, err, model.ErrUngueltig)
	nichtsGeschrieben(t, v, fremd)
	// Nach dem Rücksetzen ist dieselbe Anfrage veröffentlichbar.
	veroeffentliche(t, v, publikation("d", "1", "s1"))

	abbruch := publikation("d", "3", "s1", "s2")
	c, stop := context.WithCancel(ctx)
	v.BrichNachImportAb(t, stop)
	_, _, err = v.Imports.Publish(c, abbruch)
	erwarte(t, err, context.Canceled)
	nichtsGeschrieben(t, v, abbruch)
}

func publikationFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	veroeffentliche(t, v, publikation("d", "1", "s1"))
	zweite := publikation("d", "2", "s2", "gibt-es-nicht")
	zweite.Import.RequestHash = "req-1"
	zweite.Coverage = nil
	imp, istNeu, err := v.Imports.Publish(ctx, zweite)
	ok(t, err)
	if istNeu || imp.ID != "imp-1" || imp.BaselineSnapshotID != "snap-1" {
		t.Fatalf("zweiter Aufruf lieferte %+v, neu=%v", imp, istNeu)
	}
	if l, err := v.Snapshots.List(ctx, "d"); err != nil || len(l) != 1 {
		t.Fatalf("zweite Baseline entstand: %v %v", l, err)
	}
	if c, err := v.Coverage.List(ctx, "snap-1"); err != nil || len(c) != 1 {
		t.Fatalf("Coverage der ersten Anfrage verändert: %v %v", c, err)
	}
	nichtsGeschrieben(t, v, zweite)
	// Derselbe Hash in einem anderen Service ist eine eigene Anfrage.
	andere := publikation("e", "3", "x1")
	andere.Import.RequestHash = "req-1"
	imp, istNeu, err = v.Imports.Publish(ctx, andere)
	ok(t, err)
	if !istNeu || imp.ID != "imp-3" {
		t.Fatalf("anderer Service lieferte %+v, neu=%v", imp, istNeu)
	}
}

func publikationNebenlaeufig(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	const n = 8
	importe := make([]model.Import, n)
	neue := make([]bool, n)
	fehler := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := publikation("d", fmt.Sprint(i), "s1")
			p.Import.RequestHash = "gleich"
			importe[i], neue[i], fehler[i] = v.Imports.Publish(ctx, p)
		}()
	}
	wg.Wait()
	anzahl := 0
	for i := range n {
		ok(t, fehler[i])
		if neue[i] {
			anzahl++
		}
		if importe[i].ID != importe[0].ID {
			t.Fatalf("Import %d: %s, Import 0: %s", i, importe[i].ID, importe[0].ID)
		}
	}
	if l, err := v.Snapshots.List(ctx, "d"); anzahl != 1 || err != nil || len(l) != 1 {
		t.Fatalf("%d neue Imports, %d Baselines (%v)", anzahl, len(l), err)
	}
}

func publikationFall5(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	veroeffentliche(t, v, publikation("d", "1", "s1"))

	ungueltig := publikation("d", "2", "s1")
	ungueltig.Import.RequestHash = ""
	_, _, err := v.Imports.Publish(abgebrochen(), ungueltig)
	erwarte(t, err, model.ErrUngueltig)
	if errors.Is(err, context.Canceled) {
		t.Fatalf("Gültigkeit geht dem Kontext voraus: %v", err)
	}
	_, _, err = v.Imports.Publish(abgebrochen(), publikation("d", "1", "s1"))
	erwarte(t, err, context.Canceled)

	bekannt := publikation("d", "1", "gibt-es-nicht")
	if imp, istNeu, err := v.Imports.Publish(ctx, bekannt); err != nil || istNeu || imp.ID != "imp-1" {
		t.Fatalf("Hash geht Kennung und Quellen voraus: %+v %v %v", imp, istNeu, err)
	}
	_, _, err = v.Imports.Publish(ctx, publikation("fremd", "1", "gibt-es-nicht"))
	erwarte(t, err, outbound.ErrDienstUnbekannt)
	doppelterSnapshot := publikation("d", "3", "gibt-es-nicht")
	doppelterSnapshot.Snapshot.ID, doppelterSnapshot.Import.BaselineSnapshotID = "snap-1", "snap-1"
	_, _, err = v.Imports.Publish(ctx, doppelterSnapshot)
	erwarte(t, err, outbound.ErrBereitsVorhanden)
	doppelterImport := publikation("d", "4", "gibt-es-nicht")
	doppelterImport.Import.ID = "imp-1"
	_, _, err = v.Imports.Publish(ctx, doppelterImport)
	erwarte(t, err, outbound.ErrBereitsVorhanden)
	_, _, err = v.Imports.Publish(ctx, publikation("d", "5", "x1", "gibt-es-nicht"))
	erwarte(t, err, model.ErrUngueltig)
	_, _, err = v.Imports.Publish(ctx, publikation("d", "6", "gibt-es-nicht", "x1"))
	erwarte(t, err, outbound.ErrNichtGefunden)
}

func publikationFall6(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	veroeffentliche(t, v, publikation("d", "1", "s1"))
	neuerInhalt := publikation("d", "2", "s2")
	neuerInhalt.Snapshot.ID, neuerInhalt.Import.BaselineSnapshotID = "snap-1", "snap-1"
	neuerInhalt.Snapshot.StateHash = "ueberschrieben"
	_, _, err := v.Imports.Publish(ctx, neuerInhalt)
	erwarte(t, err, outbound.ErrBereitsVorhanden)
	if b, err := v.Snapshots.Get(ctx, "snap-1"); err != nil || b.StateHash != "sh-1" {
		t.Fatalf("Baseline überschrieben: %+v %v", b, err)
	}
	if q, err := v.Imports.Sources(ctx, "snap-1"); err != nil || !reflect.DeepEqual(q, []string{"s1"}) {
		t.Fatalf("Quellen verändert: %v %v", q, err)
	}
	v.FuegeWorkspaceSnapshotHinzu(t, "w1", "d")
	kollision := publikation("d", "3", "s1")
	kollision.Snapshot.ID, kollision.Import.BaselineSnapshotID = "w1", "w1"
	_, _, err = v.Imports.Publish(ctx, kollision)
	erwarte(t, err, outbound.ErrBereitsVorhanden)
}

func provenienzFall(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	veroeffentliche(t, v, publikation("d", "b", "s2", "s1"))
	veroeffentliche(t, v, publikation("d", "a", "s1"))
	if l, err := v.Imports.BaselinesOf(ctx, "s1"); err != nil || !reflect.DeepEqual(l, []string{"snap-a", "snap-b"}) {
		t.Fatalf("BaselinesOf s1: %v %v", l, err)
	}
	if l, err := v.Imports.Sources(ctx, "snap-b"); err != nil || !reflect.DeepEqual(l, []string{"s1", "s2"}) {
		t.Fatalf("Sources: %v %v", l, err)
	}
	l, err := v.Imports.BaselinesOf(ctx, "x1")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("Session ohne Baselines: %v %v", l, err)
	}
	_, err = v.Imports.BaselinesOf(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Imports.Sources(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	v.FuegeWorkspaceSnapshotHinzu(t, "w1", "d")
	_, err = v.Imports.Sources(ctx, "w1")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Imports.Get(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Imports.Get(ctx, "")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Imports.Sources(ctx, "")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Imports.BaselinesOf(ctx, "")
	erwarte(t, err, outbound.ErrNichtGefunden)
	erwarte(t, v.Sessions.Delete(ctx, "s1"), outbound.ErrReferenziert)
	erwarte(t, v.Sessions.Delete(ctx, "s2"), outbound.ErrReferenziert)
	ok(t, v.Sessions.Delete(ctx, "x1"))
	if _, err := v.Snapshots.Get(ctx, "snap-a"); err != nil {
		t.Fatalf("Baseline nach abgelehntem Löschen: %v", err)
	}
}

func snapshotFall(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	zeiten := []struct {
		n string
		t time.Time
	}{
		{"b", t0().Add(500 * time.Millisecond)},
		{"a", t0()},
		{"c", t0().Add(100 * time.Millisecond)},
		// 09:00 in UTC-5 sind 14:00 UTC: die Uhrzeit des Texts läge vor 10:00, die Zeit danach.
		{"zone", time.Date(2026, 10, 19, 9, 0, 0, 0, time.FixedZone("m5", -5*3600))},
		{"gleich", t0()},
	}
	for _, z := range zeiten {
		p := publikation("d", z.n, "s1")
		p.Snapshot.CreatedAt = z.t
		veroeffentliche(t, v, p)
	}
	veroeffentliche(t, v, publikation("e", "e1", "x1"))
	v.FuegeWorkspaceSnapshotHinzu(t, "w1", "d")
	l, err := v.Snapshots.List(ctx, "d")
	ok(t, err)
	var ids []string
	for _, b := range l {
		ids = append(ids, b.ID)
	}
	if want := []string{"snap-a", "snap-gleich", "snap-c", "snap-b", "snap-zone"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("Reihenfolge %v, erwartet %v", ids, want)
	}
	_, err = v.Snapshots.Get(ctx, "w1")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Snapshots.Get(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Snapshots.Get(ctx, "")
	erwarte(t, err, outbound.ErrNichtGefunden)
	l, err = v.Snapshots.List(ctx, "unbekannt")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("unbekannter Service: %v %v", l, err)
	}
	_, err = v.Snapshots.List(ctx, "")
	erwarte(t, err, model.ErrUngueltig)
}

func coverageFall(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := publikationBasis(t, neu)
	p := publikation("d", "1", "s1")
	p.Coverage = []model.Coverage{
		{EntitySet: "Zeta", Status: model.CoverageUnknown, EvidenceJSON: "{ \"b\" : 1,\n  \"a\":\"ä\" }\n"},
		{EntitySet: "Alpha", Status: model.CoverageComplete, EvidenceJSON: `[1,  2]`},
		{EntitySet: "Mid", Status: model.CoveragePartial, EvidenceJSON: `null`},
	}
	veroeffentliche(t, v, p)
	veroeffentliche(t, v, publikation("d", "2", "s2"))
	for _, soll := range p.Coverage {
		got, err := v.Coverage.Get(ctx, "snap-1", soll.EntitySet)
		ok(t, err)
		if got != soll {
			t.Fatalf("gelesen %q, geschrieben %q", got, soll)
		}
	}
	l, err := v.Coverage.List(ctx, "snap-1")
	ok(t, err)
	if want := []model.Coverage{p.Coverage[1], p.Coverage[2], p.Coverage[0]}; !reflect.DeepEqual(l, want) {
		t.Fatalf("List %q, erwartet %q", l, want)
	}
	if l, err := v.Coverage.List(ctx, "snap-2"); err != nil || len(l) != 1 || l[0].EntitySet != "Products" {
		t.Fatalf("Coverage einer anderen Baseline: %v %v", l, err)
	}
	p3 := publikation("d", "3", "s1")
	p3.Coverage = nil
	veroeffentliche(t, v, p3)
	if l, err := v.Coverage.List(ctx, "snap-3"); err != nil || l == nil || len(l) != 0 {
		t.Fatalf("Baseline ohne Einträge: %v %v", l, err)
	}
	_, err = v.Coverage.List(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	v.FuegeWorkspaceSnapshotHinzu(t, "w1", "d")
	_, err = v.Coverage.List(ctx, "w1")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Coverage.Get(ctx, "gibt-es-nicht", "Alpha")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Coverage.Get(ctx, "snap-1", "Gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	for _, a := range [][2]string{{"", "Alpha"}, {"snap-1", ""}, {"", ""}} {
		_, err = v.Coverage.Get(ctx, a[0], a[1])
		erwarte(t, err, model.ErrUngueltig)
	}
	_, err = v.Coverage.List(ctx, "")
	erwarte(t, err, model.ErrUngueltig)
}

func baselineKontextFall(t *testing.T, neu Neu) {
	bg, ctx := context.Background(), abgebrochen()
	v := publikationBasis(t, neu)
	veroeffentliche(t, v, publikation("d", "1", "s1"))
	neue := publikation("d", "2", "s2")
	_, _, err := v.Imports.Publish(ctx, neue)
	erwarte(t, err, context.Canceled)
	nichtsGeschrieben(t, v, neue)
	_, err = v.Snapshots.Get(ctx, "snap-1")
	erwarte(t, err, context.Canceled)
	_, err = v.Snapshots.List(ctx, "d")
	erwarte(t, err, context.Canceled)
	_, err = v.Imports.Get(ctx, "imp-1")
	erwarte(t, err, context.Canceled)
	_, err = v.Imports.BaselinesOf(ctx, "s1")
	erwarte(t, err, context.Canceled)
	_, err = v.Imports.Sources(ctx, "snap-1")
	erwarte(t, err, context.Canceled)
	_, err = v.Coverage.Get(ctx, "snap-1", "Products")
	erwarte(t, err, context.Canceled)
	_, err = v.Coverage.List(ctx, "snap-1")
	erwarte(t, err, context.Canceled)
	// Das Argument geht dem Kontext voraus.
	_, err = v.Snapshots.List(ctx, "")
	erwarte(t, err, model.ErrUngueltig)
	_, err = v.Coverage.Get(ctx, "", "")
	erwarte(t, err, model.ErrUngueltig)
	_, err = v.Coverage.List(ctx, "")
	erwarte(t, err, model.ErrUngueltig)
	if _, err := v.Snapshots.Get(bg, "snap-1"); err != nil {
		t.Fatalf("Baseline fehlt: %v", err)
	}
}

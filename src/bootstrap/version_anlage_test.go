package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// versionDienste sind die Services, die ein Start mit der Belegung b anlegt: Kennung und Root aus dem Index, die Version des Dienstes.
func versionDienste(b []string) []model.Service {
	aus := make([]model.Service, len(b))
	for i, v := range b {
		aus[i] = model.Service{ID: fmt.Sprintf("s%d", i), ServiceRoot: fmt.Sprintf("/s%d", i), ODataVersion: v}
	}
	return aus
}

// TestAnlageVersion fängt A13: die Anlage ruft je Dienst Create(id, root, Version, "", "") einmal, mit der Version des Dienstes, in
// der Reihenfolge der Dienste (SPEC-139 Nr. 6, SPEC-107 Nr. 14). Menge: 14 Belegungen in den Modi record, replay und mock und die
// Belegungen aus 4.0 allein in hybrid; der Port-Fake hält jeden Service fest und zählt jede Operation.
func TestAnlageVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	belegt := alleBelegungen()
	zaehle(t, "Belegungen", len(belegt), 14)
	je := map[int]int{}
	for _, b := range belegt {
		je[len(b)]++
	}
	if je[1] != 2 || je[2] != 4 || je[3] != 8 {
		t.Fatalf("Belegungen je Länge: %v, erwartet 2, 4 und 8", je)
	}
	faelle := 0
	for _, b := range belegt {
		for _, modus := range versionModi(b) {
			faelle++
			v := neuVersionHaken()
			h, err := lauf(ctx, t, versionKonfig(t.TempDir(), modus, b), nil, v.bauteile())
			if err != nil {
				t.Errorf("Modus %s, Belegung %v: Hochfahren: %v", modus, b, err)
				continue
			}
			if want := versionDienste(b); !slices.Equal(v.dienstAnlagen, want) {
				t.Errorf("Modus %s, Belegung %v: Create erhielt %+v, erwartet %+v", modus, b, v.dienstAnlagen, want)
			}
			for op, soll := range map[string]int{"Create": len(b), "Get": 0, "List": 0, "SetMetadata": 0} {
				if got := v.dienstZaehler(op); got != soll {
					t.Errorf("Modus %s, Belegung %v: %d Aufrufe von %s, erwartet %d", modus, b, got, op, soll)
				}
			}
			if len(v.dienstGets) != 0 {
				t.Errorf("Modus %s, Belegung %v: Get für %v", modus, b, v.dienstGets)
			}
			_ = h.Herunterfahren(ctx)
		}
	}
	zaehle(t, "Belegungen × Modi", faelle, 14*3+3)
}

// TestAnlageBekannt fängt A14: ein bekannter Service mit gleichen Werten tut nichts (ein Create, ein Get, kein zweites Create); jede
// Abweichung verhindert den Start mit je einer Zeile `ODR-E2001 service_root: …` und `ODR-E2001 odata_version: …` in dieser
// Reihenfolge und je einem Logeintrag; die Version geht gegen die des Dienstes (SPEC-139 Nr. 6, 8 c). Menge 9: gespeichert {4.0, 2.0}
// × konfiguriert {4.0, 2.0, fehlend} (6) und, bei anderem Root, die Paare (4.0, 4.0), (4.0, 2.0), (2.0, 4.0) (3).
func TestAnlageBekannt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	wurzel := "ODR-E2001 service_root: gespeicherter Service gleicher id hat einen anderen service_root"
	version := "ODR-E2001 odata_version: gespeicherter Service gleicher id hat eine andere odata_version"
	type zelle struct {
		gespeichert, konfiguriert string
		andererRoot               bool
	}
	var zellen []zelle
	for _, g := range versionW() {
		for _, k := range []string{"4.0", "2.0", ""} {
			zellen = append(zellen, zelle{g, k, false})
		}
	}
	zaehle(t, "gleicher Root", len(zellen), 6)
	zellen = append(zellen, zelle{"4.0", "4.0", true}, zelle{"4.0", "2.0", true}, zelle{"2.0", "4.0", true})
	zaehle(t, "Zellen", len(zellen), 9)
	soll := map[string]string{
		"level": "ERROR", "msg": "start: gespeicherter Service weicht von der Konfiguration ab",
		"ereignis": "dienst_abweichend", "error_code": "ODR-E2001", "service_id": "s0",
	}
	for _, z := range zellen {
		dir := t.TempDir()
		root := "/s0"
		if z.andererRoot {
			root = "/x"
		}
		db := oeffneAblage(ctx, t, dir)
		if err := db.Metadata().Create(ctx, model.Service{ID: "s0", ServiceRoot: root, ODataVersion: z.gespeichert}); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		var zeilen []string
		if z.andererRoot {
			zeilen = append(zeilen, wurzel)
		}
		if wirksam := cmpLeer(z.konfiguriert); wirksam != z.gespeichert {
			zeilen = append(zeilen, version)
		}
		var logPuffer syncPuffer
		v := neuVersionHaken()
		h, err := lauf(ctx, t, versionKonfig(dir, "replay", []string{z.konfiguriert}), slog.New(slog.NewTextHandler(&logPuffer, nil)), v.bauteile())
		switch {
		case len(zeilen) == 0 && err != nil:
			t.Errorf("%+v: Hochfahren: %v", z, err)
		case len(zeilen) > 0 && (err == nil || bootstrap.Meldung(err) != strings.Join(zeilen, "\n")):
			t.Errorf("%+v: Hochfahren: %v, erwartet %q", z, err, zeilen)
		}
		var eintraege []map[string]string
		for _, l := range logZeilen(t, logPuffer.String()) {
			if l["level"] == "ERROR" || l["ereignis"] != "" {
				eintraege = append(eintraege, l)
			}
		}
		if len(eintraege) != len(zeilen) {
			t.Errorf("%+v: %d Logeinträge, erwartet %d:\n%s", z, len(eintraege), len(zeilen), logPuffer.String())
		}
		for _, e := range eintraege {
			if !maps.Equal(e, soll) {
				t.Errorf("%+v: Logeintrag %v, erwartet %v", z, e, soll)
			}
		}
		if got := v.dienstZaehler("Create"); got != 1 {
			t.Errorf("%+v: %d Aufrufe von Create, erwartet 1", z, got)
		}
		if !slices.Equal(v.dienstGets, []string{"s0"}) {
			t.Errorf("%+v: Get für %v, erwartet [s0]", z, v.dienstGets)
		}
		_ = h.Herunterfahren(ctx)
	}
}

// abbruchDienste bricht den Start an der Stelle ab: stelle ist die Zahl der Create, die vollständig laufen, bevor der Kontext endet
// (0: das erste Create bricht ab, ohne etwas zu speichern; 1: das zweite; 2: nach dem zweiten Create).
type abbruchDienste struct {
	outbound.MetadataRepository
	abbruch context.CancelFunc
	stelle  int
	n       int
}

func (d *abbruchDienste) Create(ctx context.Context, s model.Service) error {
	d.n++
	if d.stelle < 2 && d.n == d.stelle+1 {
		d.abbruch()
		return ctx.Err()
	}
	err := d.MetadataRepository.Create(ctx, s)
	if d.stelle == 2 && d.n == 2 {
		d.abbruch()
	}
	return err
}

// TestAnlageWiederholung fängt A15: bricht der Start vor, zwischen oder nach den zwei Anlagen ab, legt der nächste Start nur an, was
// fehlt, mit der Version des Dienstes, und ändert keine gespeicherte Version (SPEC-107 Nr. 12 und 14, SPEC-139 Nr. 6). Menge 12:
// 3 Abbruchstellen × die 4 Belegungen für N = 2.
func TestAnlageWiederholung(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	stellen := []int{0, 1, 2}
	belegt := belegungen(2)
	zaehle(t, "Abbruchstellen", len(stellen), 3)
	zaehle(t, "Belegungen", len(belegt), 4)
	for _, stelle := range stellen {
		for _, b := range belegt {
			dir := t.TempDir()
			ersterCtx, abbruch := context.WithCancel(ctx)
			bauteile := bootstrap.StandardBauteile()
			dienste := bauteile.Dienste
			bauteile.Dienste = func(r outbound.MetadataRepository) outbound.MetadataRepository {
				return &abbruchDienste{MetadataRepository: dienste(r), abbruch: abbruch, stelle: stelle}
			}
			h, err := bootstrap.BuildMit(ersterCtx, versionKonfig(dir, "replay", b), nil, bauteile)
			if err != nil {
				t.Fatal(err)
			}
			if err := h.Hochfahren(ersterCtx); !errors.Is(err, context.Canceled) {
				t.Fatalf("Abbruchstelle %d, Belegung %v: Hochfahren liefert %v, erwartet den Abbruch", stelle, b, err)
			}
			_ = h.Herunterfahren(ctx)
			abbruch()

			v := neuVersionHaken()
			h2, err := lauf(ctx, t, versionKonfig(dir, "replay", b), nil, v.bauteile())
			if err != nil {
				t.Errorf("Abbruchstelle %d, Belegung %v: zweiter Start: %v", stelle, b, err)
				continue
			}
			if got := v.dienstZaehler("Create"); got != len(b) {
				t.Errorf("Abbruchstelle %d, Belegung %v: zweiter Start ruft Create %d-mal, erwartet %d", stelle, b, got, len(b))
			}
			_ = h2.Herunterfahren(ctx)
			db := oeffneAblage(ctx, t, dir)
			got, err := db.Metadata().List(ctx)
			if err != nil || !slices.Equal(got, versionDienste(b)) {
				t.Errorf("Abbruchstelle %d, Belegung %v: gespeichert %+v (%v), erwartet %+v", stelle, b, got, err, versionDienste(b))
			}
			_ = db.Close()
		}
	}
}

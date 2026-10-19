package outboundtest

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

func ref(ziffer string) string { return "sha256:" + strings.Repeat(ziffer, 64) }

func dienstWert(id string) model.Service {
	return model.Service{ID: id, ServiceRoot: "https://example.test/odata", ODataVersion: "4.0", MetadataBlob: ref("a"), MetadataHash: "mh-" + id}
}

func abgebrochen() context.Context {
	ctx, abbruch := context.WithCancel(context.Background())
	abbruch()
	return ctx
}

// RunServiceVertrag prüft SPEC-107 Nr. 1 bis 3 und Nr. 12 für MetadataRepository.
func RunServiceVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Create und Get tragen alle Felder, auch ohne Metadaten", serviceFall1},
		{"Create lehnt Ungültiges und Doppeltes ab", serviceFall2},
		{"Get und List", serviceFall3},
		{"SetMetadata ersetzt, löscht und prüft", serviceFall4},
		{"abgebrochener Kontext beendet jede Operation ohne Änderung", serviceFall5},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func serviceFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	mit, ohne := dienstWert("mit"), dienstWert("ohne")
	ohne.MetadataBlob, ohne.MetadataHash = "", ""
	ok(t, v.Metadata.Create(ctx, mit))
	ok(t, v.Metadata.Create(ctx, ohne))
	for _, soll := range []model.Service{mit, ohne} {
		got, err := v.Metadata.Get(ctx, soll.ID)
		ok(t, err)
		if got != soll {
			t.Fatalf("gelesen %+v, geschrieben %+v", got, soll)
		}
	}
}

func serviceFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	for name, mut := range map[string]func(*model.Service){
		"ohne Kennung":               func(s *model.Service) { s.ID = "" },
		"ohne service_root":          func(s *model.Service) { s.ServiceRoot = "" },
		"ohne odata_version":         func(s *model.Service) { s.ODataVersion = "" },
		"Blob ohne Hash":             func(s *model.Service) { s.MetadataHash = "" },
		"Hash ohne Blob":             func(s *model.Service) { s.MetadataBlob = "" },
		"ungültige Referenz":         func(s *model.Service) { s.MetadataBlob = "sha256:abc" },
		"Referenz in Großbuchstaben": func(s *model.Service) { s.MetadataBlob = ref("A") },
	} {
		s := dienstWert("x")
		mut(&s)
		if err := v.Metadata.Create(ctx, s); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	if l, err := v.Metadata.List(ctx); err != nil || len(l) != 0 {
		t.Fatalf("Ungültiges wurde gespeichert: %v %v", l, err)
	}
	ok(t, v.Metadata.Create(ctx, dienstWert("x")))
	anders := dienstWert("x")
	anders.ServiceRoot = "https://anders.test"
	erwarte(t, v.Metadata.Create(ctx, anders), outbound.ErrBereitsVorhanden)
	if got, _ := v.Metadata.Get(ctx, "x"); got != dienstWert("x") {
		t.Fatalf("doppeltes Create änderte den Service: %+v", got)
	}
}

func serviceFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	_, err := v.Metadata.Get(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	l, err := v.Metadata.List(ctx)
	ok(t, err)
	if l == nil || len(l) != 0 {
		t.Fatalf("List ohne Services: %v", l)
	}
	for _, id := range []string{"z", "a", "m"} {
		ok(t, v.Metadata.Create(ctx, dienstWert(id)))
	}
	l, err = v.Metadata.List(ctx)
	ok(t, err)
	var ids []string
	for _, s := range l {
		ids = append(ids, s.ID)
	}
	if !reflect.DeepEqual(ids, []string{"a", "m", "z"}) {
		t.Fatalf("Reihenfolge %v", ids)
	}
}

func serviceFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	ok(t, v.Metadata.Create(ctx, dienstWert("d")))
	ok(t, v.Metadata.SetMetadata(ctx, "d", ref("b"), "neu"))
	if got, _ := v.Metadata.Get(ctx, "d"); got.MetadataBlob != ref("b") || got.MetadataHash != "neu" {
		t.Fatalf("nach Ersetzen: %+v", got)
	}
	for name, a := range map[string][2]string{"nur Blob": {ref("c"), ""}, "nur Hash": {"", "h"}, "ungültige Referenz": {"sha256:xyz", "h"}} {
		if err := v.Metadata.SetMetadata(ctx, "d", a[0], a[1]); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	if got, _ := v.Metadata.Get(ctx, "d"); got.MetadataBlob != ref("b") || got.MetadataHash != "neu" {
		t.Fatalf("ungültiges SetMetadata änderte den Service: %+v", got)
	}
	ok(t, v.Metadata.SetMetadata(ctx, "d", "", ""))
	if got, _ := v.Metadata.Get(ctx, "d"); got.MetadataBlob != "" || got.MetadataHash != "" {
		t.Fatalf("nach Löschen: %+v", got)
	}
	erwarte(t, v.Metadata.SetMetadata(ctx, "gibt-es-nicht", ref("b"), "h"), outbound.ErrNichtGefunden)
	// Die Argumente werden vor der Kennung geprüft.
	erwarte(t, v.Metadata.SetMetadata(ctx, "gibt-es-nicht", ref("b"), ""), model.ErrUngueltig)
	erwarte(t, v.Metadata.SetMetadata(ctx, "", ref("b"), "h"), outbound.ErrNichtGefunden)
	_, err := v.Metadata.Get(ctx, "")
	erwarte(t, err, outbound.ErrNichtGefunden)
}

func serviceFall5(t *testing.T, neu Neu) {
	bg, ctx := context.Background(), abgebrochen()
	v := neu(t)
	ok(t, v.Metadata.Create(bg, dienstWert("d")))
	erwarte(t, v.Metadata.Create(ctx, dienstWert("neu")), context.Canceled)
	_, err := v.Metadata.Get(ctx, "d")
	erwarte(t, err, context.Canceled)
	_, err = v.Metadata.List(ctx)
	erwarte(t, err, context.Canceled)
	erwarte(t, v.Metadata.SetMetadata(ctx, "d", "", ""), context.Canceled)
	erwarte(t, v.Metadata.Create(ctx, model.Service{}), model.ErrUngueltig)
	erwarte(t, v.Metadata.SetMetadata(ctx, "d", ref("b"), ""), model.ErrUngueltig)
	if _, err := v.Metadata.Get(bg, "neu"); !errors.Is(err, outbound.ErrNichtGefunden) {
		t.Fatalf("Create trotz Abbruch: %v", err)
	}
	if got, _ := v.Metadata.Get(bg, "d"); got != dienstWert("d") {
		t.Fatalf("SetMetadata trotz Abbruch: %+v", got)
	}
}

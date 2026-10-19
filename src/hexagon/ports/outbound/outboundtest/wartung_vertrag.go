package outboundtest

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// RunWartungVertrag prüft StorageMaintenance, SPEC-108 Nr. 15 bis 17, und dass die Listen nie
// nil sind.
func RunWartungVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Check einer sauberen Datenbank ist leer", wartungFall1},
		{"Check meldet eine Fremdschlüssel-Verletzung", wartungFall2},
		{"Check meldet eine Verletzung eines CHECK-Constraints", wartungFall3},
		{"ReferencedBlobs liefert jede Quelle einmal, aufsteigend", wartungFall4},
		{"ReferencedBlobs einer leeren Datenbank", wartungFall5},
		{"abgebrochener Kontext beendet jede Operation", wartungKontextFall},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func wartungFall1(t *testing.T, neu Neu) {
	v := wsMitSnapshots(t, neu)
	b, err := v.Wartung.Check(context.Background())
	if err != nil || b == nil || len(b) != 0 {
		t.Fatalf("Befunde %v %v", b, err)
	}
}

func wartungFall2(t *testing.T, neu Neu) {
	v := wsMitSnapshots(t, neu)
	v.VerletzeFremdschluessel(t)
	b, err := v.Wartung.Check(context.Background())
	ok(t, err)
	if len(b) != 1 || !strings.Contains(b[0], "foreign_key_check") || !strings.Contains(b[0], "sessions") {
		t.Fatalf("Befunde %v", b)
	}
}

func wartungFall3(t *testing.T, neu Neu) {
	v := wsMitSnapshots(t, neu)
	v.VerletzeCheck(t)
	b, err := v.Wartung.Check(context.Background())
	ok(t, err)
	if len(b) == 0 || !strings.Contains(strings.Join(b, "\n"), "integrity_check") {
		t.Fatalf("Befunde %v", b)
	}
	for _, s := range b {
		if strings.Contains(s, "foreign_key_check") {
			t.Fatalf("Fremdschlüssel-Befund ohne Verletzung: %v", b)
		}
	}
}

func wartungFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsMitSnapshots(t, neu)
	// Aufgebaut sind: snap-1 und snap-e tragen ref("a"), ws-d und ws-e ref("c").
	dienst := dienstWert("m")
	dienst.MetadataBlob = ref("9")
	ok(t, v.Metadata.Create(ctx, dienst))
	gleich := dienstWert("m2")
	gleich.MetadataBlob = ref("a")
	ok(t, v.Metadata.Create(ctx, gleich))
	ok(t, v.Metadata.Create(ctx, model.Service{ID: "ohne", ServiceRoot: "/", ODataVersion: "4.0"}))
	e := exchange("s2", "e1", 1)
	e.RequestBody, e.RequestBlob, e.ResponseBlob = model.BodyPresent, ref("5"), ref("e")
	ok(t, v.Recording.Append(ctx, e))
	ohne := exchange("s2", "e2", 2)
	ohne.ResponseBody, ohne.ResponseBlob = model.BodyAbsent, ""
	ok(t, v.Recording.Append(ctx, ohne))
	doppelt := exchange("s2", "e3", 3)
	doppelt.RequestBody, doppelt.RequestBlob, doppelt.ResponseBlob = model.BodyPresent, ref("6"), ref("5")
	ok(t, v.Recording.Append(ctx, doppelt))
	got, err := v.Wartung.ReferencedBlobs(ctx)
	ok(t, err)
	if want := []string{ref("5"), ref("6"), ref("9"), ref("a"), ref("c"), ref("e")}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Referenzen %v, erwartet %v", got, want)
	}
}

func wartungFall5(t *testing.T, neu Neu) {
	got, err := neu(t).Wartung.ReferencedBlobs(context.Background())
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("Referenzen %v %v", got, err)
	}
}

func wartungKontextFall(t *testing.T, neu Neu) {
	ctx := abgebrochen()
	v := wsMitSnapshots(t, neu)
	_, err := v.Wartung.Check(ctx)
	erwarte(t, err, context.Canceled)
	_, err = v.Wartung.ReferencedBlobs(ctx)
	erwarte(t, err, context.Canceled)
}

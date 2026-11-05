package bootstrap_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// konfigMitWerten trägt in jedem Feld, das Import oder Workspace-Dienst erreicht, einen Wert ungleich dem Default:
// zwei Dienste mit verschiedenen Kennungen und URLs, alle sechs Felder der Regelmenge, MaxCaptureBody.
func konfigMitWerten(dir, modus string) bootstrap.Konfiguration {
	k := basisKonfig(dir, modus,
		bootstrap.Dienst{ID: "x1", Root: "/r1", UpstreamURL: "http://u1.test/o"},
		bootstrap.Dienst{ID: "x2", Root: "/r2", UpstreamURL: "http://u2.test/o"})
	k.MatchVaryHeaders = []string{"x-vary"}
	k.VolatileHeaders = []string{"x-vol"}
	k.RedactHeaders = []string{"x-red"}
	k.RedactQueryParams = []string{"qred"}
	k.MatchInvalidEncoding = "reject"
	k.MatchBodyMode = "json-canonical"
	k.MaxCaptureBody = 12345
	return k
}

func baue(ctx context.Context, t *testing.T, k bootstrap.Konfiguration, h *haken) *bootstrap.Handle {
	t.Helper()
	return baueMitLog(ctx, t, k, h, nil)
}

// baueMitLog baut den Handle mit dem übergebenen Logger; nil lässt Build den eigenen wählen.
func baueMitLog(ctx context.Context, t *testing.T, k bootstrap.Konfiguration, h *haken, log *slog.Logger) *bootstrap.Handle {
	t.Helper()
	hd, err := bootstrap.BuildMit(ctx, k, log, h.bauteile())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = hd.Herunterfahren(ctx) })
	return hd
}

// TestImportErhaeltWerte fängt die Zeilen M5 und M6 von slice-m2-referenzdienst (Import): jedes Feld der Konfig des Imports gegen seinen
// Verbraucher; die Felder der Regelmenge als Schleife, die Menge der Felder von Konfig als Tabelle.
func TestImportErhaeltWerte(t *testing.T) {
	ctx := context.Background()
	h := neuHaken()
	k := konfigMitWerten(t.TempDir(), modusRecord)
	hd := baue(ctx, t, k, h)
	got := h.importKfg
	if got == nil {
		t.Fatal("der Konstruktor des Imports wurde nicht gerufen")
	}
	erwartet := model.SchluesselRegeln{
		VaryHeaders: []string{"x-vary"}, VolatileHeaders: []string{"x-vol"}, RedactHeaders: []string{"x-red"},
		RedactQueryParams: []string{"qred"}, InvalidEncoding: model.KodierungReject, BodyMode: model.BodyJSONCanonical,
	}
	rt, rg, re := reflect.TypeOf(erwartet), reflect.ValueOf(got.Regeln), reflect.ValueOf(erwartet)
	if rt.NumField() != 6 {
		t.Fatalf("SchluesselRegeln hat %d Felder, der Test kennt 6", rt.NumField())
	}
	for i := range rt.NumField() {
		if re.Field(i).IsZero() {
			t.Errorf("Regelfeld %s trägt im Test den Default", rt.Field(i).Name)
		}
		if !reflect.DeepEqual(rg.Field(i).Interface(), re.Field(i).Interface()) {
			t.Errorf("Regelfeld %s: %v, erwartet %v", rt.Field(i).Name, rg.Field(i), re.Field(i))
		}
		if !reflect.DeepEqual(rg.Field(i).Interface(), reflect.ValueOf(h.recorder.Regeln).Field(i).Interface()) {
			t.Errorf("Regelfeld %s ist an Import und Record nicht gleich", rt.Field(i).Name)
		}
	}
	_, istWaechter := got.Blobs.(*maintenance.BlobWaechter)
	_, istSessions := got.Sessions.(*sqlite.SessionRepository)
	_, istRecording := got.Exchanges.(*sqlite.RecordingRepository)
	_, istSnapshots := got.Snapshots.(*sqlite.SnapshotRepository)
	_, istImports := got.Imports.(*sqlite.ImportRepository)
	_, istCsdl := got.MetadatenCodec.(*csdl.Leser)
	_, istJSON := got.DatenCodec.(*jsoncodec.Leser)
	dieselbe := map[string]bool{
		"Dienste": reflect.DeepEqual(got.Dienste, []baselineimport.Dienst{
			{ID: "x1", UpstreamURL: "http://u1.test/o"}, {ID: "x2", UpstreamURL: "http://u2.test/o"}}),
		"Sessions":       istSessions,
		"Exchanges":      istRecording,
		"Blobs":          istWaechter && got.Blobs == h.pipeline.Blobs && got.Blobs == h.wartung.Blobs,
		"Metadaten":      got.Metadaten != nil && got.Metadaten == h.session.Dienste && got.Metadaten == h.wartung.Dienste,
		"Snapshots":      istSnapshots,
		"Imports":        istImports,
		"MetadatenCodec": istCsdl,
		"DatenCodec":     istJSON,
		"Regeln":         true,
		"MaxCaptureBody": got.MaxCaptureBody == 12345 && got.MaxCaptureBody == h.pipeline.MaxBytes,
		"Uhr":            got.Uhr != nil && got.Uhr == h.wartung.Uhr && got.Uhr == h.recorder.Uhr,
		"IDs":            got.IDs != nil && got.IDs == bootstrap.IDsVon(hd),
	}
	typ := reflect.TypeOf(*got)
	if len(dieselbe) != typ.NumField() {
		t.Errorf("Tabelle hat %d Zeilen, Konfig %d Felder", len(dieselbe), typ.NumField())
	}
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		ok, da := dieselbe[name]
		if !da {
			t.Errorf("Feld %s steht nicht in der Tabelle", name)
		} else if !ok {
			t.Errorf("Feld %s erreicht den Import nicht wie erwartet", name)
		}
	}
	if h.importe != 1 {
		t.Errorf("Konstruktor des Imports %d-mal gerufen", h.importe)
	}
}

// TestWorkspaceErhaeltWerte fängt die Zeile M6 von slice-m2-referenzdienst (Workspace-Dienst) in beiden Modi: Feld für Feld.
func TestWorkspaceErhaeltWerte(t *testing.T) {
	ctx := context.Background()
	for _, modus := range modi() {
		t.Run(modus, func(t *testing.T) {
			h := neuHaken()
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			hd := baueMitLog(ctx, t, konfigMitWerten(t.TempDir(), modus), h, log)
			got := h.workspace
			if got == nil {
				t.Fatal("der Konstruktor des Workspace-Dienstes wurde nicht gerufen")
			}
			_, istWaechter := got.Blobs.(*maintenance.BlobWaechter)
			_, istWorkspaces := got.Workspaces.(*sqlite.WorkspaceRepository)
			_, istSnapshots := got.Snapshots.(*sqlite.SnapshotRepository)
			verbraucher := h.pipeline.Blobs
			if modus == modusReplay {
				verbraucher = h.replay.Blobs
			}
			dieselbe := map[string]bool{
				"Workspaces":  istWorkspaces,
				"Transaktion": h.db != nil && got.Transaktion == any(h.db),
				"Snapshots":   istSnapshots,
				"Blobs":       istWaechter && got.Blobs == verbraucher && got.Blobs == h.wartung.Blobs,
				"Dienste":     got.Dienste != nil && got.Dienste == h.wartung.Dienste,
				"Wartung":     got.Wartung != nil && got.Wartung == hd.Wartung(),
				"Uhr":         got.Uhr != nil && got.Uhr == h.wartung.Uhr,
				"IDs":         got.IDs != nil && got.IDs == bootstrap.IDsVon(hd),
				"Log":         got.Log != nil && got.Log == log,
			}
			typ := reflect.TypeOf(*got)
			if len(dieselbe) != typ.NumField() {
				t.Errorf("Tabelle hat %d Zeilen, Konfig %d Felder", len(dieselbe), typ.NumField())
			}
			for i := range typ.NumField() {
				name := typ.Field(i).Name
				ok, da := dieselbe[name]
				if !da {
					t.Errorf("Feld %s steht nicht in der Tabelle", name)
				} else if !ok {
					t.Errorf("Feld %s erreicht den Workspace-Dienst nicht wie erwartet", name)
				}
			}
			if h.workspaces != 1 {
				t.Errorf("Konstruktor des Workspace-Dienstes %d-mal gerufen", h.workspaces)
			}
			if (h.importKfg != nil) != (modus == modusRecord) {
				t.Errorf("Konstruktor des Imports im Modus %s: %v", modus, h.importKfg != nil)
			}
		})
	}
}

// TestZugaengeJeModus fängt die Zeile M9b von slice-m2-referenzdienst: {Import, Workspaces} x {record, replay} und der gleichzeitige Aufruf.
func TestZugaengeJeModus(t *testing.T) {
	ctx := context.Background()
	zugaenge := []struct {
		name string
		hole func(*bootstrap.Handle) any
		// nilImReplay sagt, ob der Zugang im Modus replay die nil-Schnittstelle liefert.
		nilImReplay bool
	}{
		// Die Umwandlung einer nil-Schnittstelle in any bleibt nil; ein typisierter Nullzeiger bliebe es nicht.
		{"Import", func(h *bootstrap.Handle) any { return h.Import() }, true},
		{"Workspaces", func(h *bootstrap.Handle) any { return h.Workspaces() }, false},
	}
	for _, z := range zugaenge {
		for _, modus := range modi() {
			t.Run(z.name+"/"+modus, func(t *testing.T) {
				hd := baue(ctx, t, basisKonfig(t.TempDir(), modus, dienstA("http://127.0.0.1:9/a")), neuHaken())
				// Der erste Aufruf jedes Zugangs liegt in den Goroutinen, gleichzeitig: füllt ein Zugang sein Feld
				// beim ersten Aufruf, schreibt er dort, und der Detektor meldet es.
				const goroutinen, runden = 8, 200
				start := make(chan struct{})
				var wg sync.WaitGroup
				werte := make([][]any, goroutinen)
				for g := range goroutinen {
					wg.Add(1)
					go func() {
						defer wg.Done()
						<-start
						for range runden {
							werte[g] = append(werte[g], z.hole(hd))
						}
					}()
				}
				close(start)
				wg.Wait()
				erster := werte[0][0]
				if nilFall := z.nilImReplay && modus == modusReplay; nilFall != (erster == nil) {
					t.Fatalf("Zugang %s im Modus %s: %v (nil erwartet: %v)", z.name, modus, erster, nilFall)
				}
				if erster != nil {
					if v := reflect.ValueOf(erster); v.Kind() == reflect.Pointer && v.IsNil() {
						t.Fatalf("Zugang %s liefert einen typisierten Nullzeiger", z.name)
					}
				}
				for g, liste := range werte {
					if len(liste) != runden {
						t.Fatalf("Goroutine %d: %d Werte", g, len(liste))
					}
					for i, w := range liste {
						if w != erster {
							t.Fatalf("Goroutine %d, Aufruf %d: anderer Wert als der erste", g, i)
						}
					}
				}
			})
		}
	}
}

// TestZugaengeTypen fängt, dass die Zugänge die echten Dienste liefern (nicht einen Fake oder die falsche Art).
func TestZugaengeTypen(t *testing.T) {
	ctx := context.Background()
	for _, modus := range modi() {
		hd := baue(ctx, t, basisKonfig(t.TempDir(), modus, dienstA("http://127.0.0.1:9/a")), neuHaken())
		if _, ok := hd.Workspaces().(*workspace.Verwaltung); !ok {
			t.Errorf("Workspaces() im Modus %s: %T", modus, hd.Workspaces())
		}
		if i, ok := hd.Import().(*baselineimport.Importer); ok != (modus == modusRecord) {
			t.Errorf("Import() im Modus %s: %T", modus, i)
		}
	}
}

// TestKonstruktorFehlerSchliesstAblage fängt, dass ein Fehler eines der beiden Konstruktoren Build
// scheitern lässt, unverändert zurückkommt und die Ablage schließt.
func TestKonstruktorFehlerSchliesstAblage(t *testing.T) {
	ctx := context.Background()
	fehlerHier := errors.New("Konstruktor scheitert")
	faelle := []struct {
		name  string
		modus string
		setze func(*bootstrap.Bauteile)
	}{
		{"Import", modusRecord, func(b *bootstrap.Bauteile) {
			b.NeuerImport = func(baselineimport.Konfig) (inbound.ImportUseCase, error) { return nil, fehlerHier }
		}},
		{"Workspaces/record", modusRecord, func(b *bootstrap.Bauteile) {
			b.NeueWorkspaces = func(workspace.Konfig) (inbound.WorkspaceUseCase, error) { return nil, fehlerHier }
		}},
		{"Workspaces/replay", modusReplay, func(b *bootstrap.Bauteile) {
			b.NeueWorkspaces = func(workspace.Konfig) (inbound.WorkspaceUseCase, error) { return nil, fehlerHier }
		}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			h := neuHaken()
			b := h.bauteile()
			f.setze(&b)
			hd, err := bootstrap.BuildMit(ctx, basisKonfig(t.TempDir(), f.modus, dienstA("http://127.0.0.1:9/a")), nil, b)
			if !errors.Is(err, fehlerHier) || hd != nil {
				t.Fatalf("Build: %v, %v", hd, err)
			}
			if h.p.zaehle(pAblageZu) != 1 {
				t.Errorf("Ablage %d-mal geschlossen", h.p.zaehle(pAblageZu))
			}
		})
	}
}

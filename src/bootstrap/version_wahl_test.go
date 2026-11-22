package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// versionHaken ist der Haken von Mock, Import und Anlage, der zusätzlich die Konfig des Hybrid-Dienstes festhält.
type versionHaken struct {
	*m3Haken
	hybrid []hybrid.Konfig
}

func neuVersionHaken() *versionHaken { return &versionHaken{m3Haken: neuM3Haken()} }

func (v *versionHaken) bauteile() bootstrap.Bauteile {
	b := v.m3Haken.bauteile()
	echt := bootstrap.StandardBauteile()
	b.NeuerHybrid = func(k hybrid.Konfig) (inbound.HybridUseCase, error) {
		v.hybrid = append(v.hybrid, k)
		return echt.NeuerHybrid(k)
	}
	return b
}

// versionKonfig ist eine gültige Konfiguration im Modus modus mit einem Dienst je Version von b (Kennung und Root aus dem Index).
func versionKonfig(dir, modus string, b []string) bootstrap.Konfiguration {
	dienste := make([]bootstrap.Dienst, len(b))
	for i, v := range b {
		dienste[i] = bootstrap.Dienst{ID: fmt.Sprintf("s%d", i), Root: fmt.Sprintf("/s%d", i), UpstreamURL: fmt.Sprintf("http://127.0.0.1:9/s%d", i), Version: v}
	}
	k := basisKonfig(dir, modus, dienste...)
	k.PublicBaseURL, k.StrictCoverage, k.MaxCaptureBody = "http://127.0.0.1:8080/api", true, 12345
	return k
}

// versionModi sind die Modi, in denen die Belegung b baut: der Modus hybrid nimmt keinen Dienst 2.0 (SPEC-110 Nr. 17 b).
func versionModi(b []string) []string {
	if mitV2(b) {
		return []string{"record", "replay", "mock"}
	}
	return []string{"record", "replay", "mock", "hybrid"}
}

// versionBau baut im Modus modus mit der Belegung b und liefert Haken und Handle.
func versionBau(t *testing.T, modus string, b []string, etag bool) (*versionHaken, *bootstrap.Handle) {
	t.Helper()
	v := neuVersionHaken()
	k := versionKonfig(t.TempDir(), modus, b)
	k.EtagEnabled = etag
	h, err := bootstrap.BuildMit(context.Background(), k, nil, v.bauteile())
	if err != nil {
		t.Fatalf("Build im Modus %s mit %v: %v", modus, b, err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(context.Background()) })
	return v, h
}

// paketVon ist der Name des Pakets, aus dem der Typ des Werts stammt; nil ist leer.
func paketVon(x any) string {
	if x == nil {
		return ""
	}
	t := reflect.TypeOf(x)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return path.Base(t.PkgPath())
}

// satzOrakel ist der Satz je Version, die Pakete aus SPEC-139 Nr. 11 b abgeschrieben (`csdl`, `jsoncodec`, `urlparser` und die mit
// Endung `v2`); Eingabe bei 2.0 ist nil.
func satzOrakel() map[string]map[string]string {
	return map[string]map[string]string{
		"4.0": {"Version": "4.0", "Metadaten": "csdl", "Daten": "jsoncodec", "URL": "urlparser", "Schreiber": "jsoncodec", "Eingabe": "jsoncodec", "NurLesend": "false"},
		"2.0": {"Version": "2.0", "Metadaten": "csdlv2", "Daten": "jsoncodecv2", "URL": "urlparserv2", "Schreiber": "jsoncodecv2", "Eingabe": "", "NurLesend": "true"},
	}
}

// satzFelder liest die sieben Felder eines Satzes in die Form des Orakels.
func satzFelder(s outbound.CodecSatz) map[string]string {
	return map[string]string{
		"Version": s.Version, "Metadaten": paketVon(s.Metadaten), "Daten": paketVon(s.Daten), "URL": paketVon(s.URL),
		"Schreiber": paketVon(s.Schreiber), "Eingabe": paketVon(s.Eingabe), "NurLesend": strconv.FormatBool(s.NurLesend),
	}
}

// versionOrakelZaehlen hält die Tabelle 7 Felder × 2 Versionen fest.
func versionOrakelZaehlen(t *testing.T) {
	t.Helper()
	zaehle(t, "Versionen des Orakels", len(satzOrakel()), 2)
	for v, felder := range satzOrakel() {
		zaehle(t, "Felder des Satzes "+v, len(felder), 7)
	}
}

// vergleicheSatz meldet jedes Feld, in dem der Satz vom Orakel der Version abweicht.
func vergleicheSatz(t *testing.T, was string, s outbound.CodecSatz, version string) {
	t.Helper()
	got, want := satzFelder(s), satzOrakel()[version]
	for feld, w := range want {
		if got[feld] != w {
			t.Errorf("%s: Feld %s ist %q, erwartet %q (Version %s)", was, feld, got[feld], w, version)
		}
	}
}

// hoch ist b hoch n.
func hoch(b, n int) int {
	r := 1
	for i := 0; i < n; i++ {
		r *= b
	}
	return r
}

// folgenAus sind alle Folgen der Länge 1 bis nMax aus den Symbolen, die erste Stelle am langsamsten.
func folgenAus(symbole []string, nMax int) [][]string {
	var aus [][]string
	for n := 1; n <= nMax; n++ {
		for k := 0; k < hoch(len(symbole), n); k++ {
			f := make([]string, n)
			for i := range f {
				f[i] = symbole[(k/hoch(len(symbole), n-1-i))%len(symbole)]
			}
			aus = append(aus, f)
		}
	}
	return aus
}

// TestBuildVersionLeer fängt A11: eine leere Version gilt 4.0, an jeder Stelle der Liste; Create erhält 4.0, der Dienst trägt den Satz
// 4.0, und der Aufrufer behält seine Konfiguration (SPEC-139 Nr. 2, 6). Menge 39: die Folgen der Länge 1, 2 und 3 aus {leer, 4.0, 2.0}
// (3 + 9 + 27), im Modus mock mit Hochfahren; in den Modi record, replay und hybrid (ohne Dienst 2.0) baut Build jede Folge.
func TestBuildVersionLeer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	folgen := folgenAus([]string{"", "4.0", "2.0"}, 3)
	zaehle(t, "Folgen", len(folgen), 3+9+27)
	for _, b := range folgen {
		soll := make([]string, len(b))
		for i, v := range b {
			soll[i] = cmpLeer(v)
		}
		v := neuVersionHaken()
		k := versionKonfig(t.TempDir(), "mock", b)
		eingabe := slices.Clone(k.Dienste)
		h, err := bootstrap.BuildMit(ctx, k, nil, v.bauteile())
		if err != nil {
			t.Fatalf("Build mit %q: %v", b, err)
		}
		if !reflect.DeepEqual(k.Dienste, eingabe) {
			t.Errorf("Build mit %q ändert die Dienste des Aufrufers: %+v, vorher %+v", b, k.Dienste, eingabe)
		}
		if err := h.Hochfahren(ctx); err != nil {
			t.Fatalf("Hochfahren mit %q: %v", b, err)
		}
		var angelegt []string
		for _, s := range v.dienstAnlagen {
			angelegt = append(angelegt, s.ODataVersion)
		}
		if !slices.Equal(angelegt, soll) {
			t.Errorf("Folge %q: Create erhielt die Versionen %q, erwartet %q", b, angelegt, soll)
		}
		for i, d := range v.mock[0].Dienste {
			if d.Codecs == nil {
				t.Errorf("Folge %q, Dienst %d ohne Satz", b, i)
				continue
			}
			vergleicheSatz(t, fmt.Sprintf("Folge %q, Dienst %d", b, i), *d.Codecs, soll[i])
		}
		_ = h.Herunterfahren(ctx)
		for _, modus := range []string{"record", "replay", "hybrid"} {
			if modus == "hybrid" && mitV2(b) {
				continue
			}
			hm, err := bootstrap.BuildMit(ctx, versionKonfig(t.TempDir(), modus, b), nil, neuVersionHaken().bauteile())
			if err != nil {
				t.Errorf("Modus %s, Folge %q: Build: %v", modus, b, err)
				continue
			}
			_ = hm.Herunterfahren(ctx)
		}
	}
}

// cmpLeer ist die wirksame Version zu einer Angabe: leer gilt 4.0.
func cmpLeer(v string) string {
	if v == "" {
		return "4.0"
	}
	return v
}

// TestBuildVersionUngueltig fängt A12: eine unbekannte Version und hybrid mit 2.0 sind model.ErrUngueltig, ohne Handle und ohne die
// Ablage zu öffnen (SPEC-139 Nr. 2). Menge: F × {record, replay, mock} (81), (hybrid, 2.0) (1) und der Randleerraum R × dieselben
// Modi (72), den Build nicht kürzt.
func TestBuildVersionUngueltig(t *testing.T) {
	type fall struct{ modus, version string }
	var faelle []fall
	f := versionF()
	zaehle(t, "F", len(f), 27)
	modi := []string{"record", "replay", "mock"}
	zaehle(t, "Modi", len(modi), 3)
	for _, w := range f {
		for _, m := range modi {
			faelle = append(faelle, fall{m, w})
		}
	}
	zaehle(t, "F × Modi", len(faelle), 81)
	for _, r := range versionR() {
		for _, m := range modi {
			faelle = append(faelle, fall{m, r.eingabe})
		}
	}
	zaehle(t, "R × Modi", len(faelle)-81, 72)
	faelle = append(faelle, fall{"hybrid", "2.0"})
	for _, fl := range faelle {
		dir := t.TempDir()
		v := neuVersionHaken()
		h, err := bootstrap.BuildMit(context.Background(), versionKonfig(dir, fl.modus, []string{fl.version}), nil, v.bauteile())
		if err == nil || !errors.Is(err, model.ErrUngueltig) || h != nil {
			if h != nil {
				_ = h.Herunterfahren(context.Background())
			}
			t.Errorf("Modus %s, Version %q: Build liefert %v, erwartet model.ErrUngueltig und keinen Handle", fl.modus, fl.version, err)
		}
		if v.oeffnen != 0 || v.p.zaehle(pAblage) != 0 || v.p.zaehle(pAblageZu) != 0 {
			t.Errorf("Modus %s, Version %q: die Ablage wurde berührt", fl.modus, fl.version)
		}
		if _, err := os.Stat(filepath.Join(dir, dateiDB)); err == nil {
			t.Errorf("Modus %s, Version %q: %s ist angelegt", fl.modus, fl.version, dateiDB)
		}
		if _, err := os.Stat(filepath.Join(dir, verzeichnisB)); err == nil {
			t.Errorf("Modus %s, Version %q: %s ist angelegt", fl.modus, fl.version, verzeichnisB)
		}
	}
}

// TestBuildVersionStellen fängt A12 an jeder Stelle der Liste (SPEC-139 Nr. 2: die Version jedes Dienstes): Build weist jeden Dienst mit
// unbekannter Version und im Modus hybrid jeden Dienst 2.0 ab, gleich an welcher Stelle er steht, gleich wie viele Dienste die Liste
// führt und ob eine der beiden Verletzungen oder beide in verschiedenen Diensten vorliegen; alle anderen Konfigurationen baut es.
// Menge: die Folgen der Länge 1 bis 3 aus {4.0, 2.0, unbekannt} (39) × die vier Modi (156), davon 45 gültig; Orakel: abgewiesen genau
// bei einem unbekannten Dienst oder bei hybrid mit einem Dienst 2.0 (ein Wert ist nie beides, die Art „beides“ sind zwei Dienste).
func TestBuildVersionStellen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	const unbekannt = "v2.0"
	symbole := []string{"4.0", "2.0", unbekannt}
	modi := []string{"record", "replay", "mock", "hybrid"}
	folgen := folgenAus(symbole, 3)
	gesamt, gueltig := 0, 0
	for n := 1; n <= 3; n++ {
		gesamt += hoch(len(symbole), n) * len(modi)
		gueltig += hoch(len(versionW()), n)*(len(modi)-1) + 1
	}
	zaehle(t, "Folgen", len(folgen), 3+9+27)
	faelle, gebaut := 0, 0
	for _, b := range folgen {
		for _, modus := range modi {
			faelle++
			mitUnbekannt, mitHybridV2 := slices.Contains(b, unbekannt), modus == "hybrid" && slices.Contains(b, "2.0")
			dir := t.TempDir()
			v := neuVersionHaken()
			h, err := bootstrap.BuildMit(ctx, versionKonfig(dir, modus, b), nil, v.bauteile())
			if !mitUnbekannt && !mitHybridV2 {
				gebaut++
				if err != nil || h == nil {
					t.Errorf("Modus %s, Dienste %v: Build liefert %v, erwartet einen Handle", modus, b, err)
				}
				if h != nil {
					_ = h.Herunterfahren(ctx)
				}
				continue
			}
			art := "hybrid mit 2.0"
			switch {
			case mitUnbekannt && mitHybridV2:
				art = "beides in zwei Diensten"
			case mitUnbekannt:
				art = "unbekannte Version"
			}
			if h != nil {
				_ = h.Herunterfahren(ctx)
			}
			if err == nil || !errors.Is(err, model.ErrUngueltig) || h != nil {
				t.Errorf("Modus %s, Dienste %v (%s): Build liefert %v, erwartet model.ErrUngueltig und keinen Handle", modus, b, art, err)
			}
			if v.oeffnen != 0 || v.p.zaehle(pAblage) != 0 || v.p.zaehle(pAblageZu) != 0 {
				t.Errorf("Modus %s, Dienste %v (%s): die Ablage wurde berührt", modus, b, art)
			}
			for _, name := range []string{dateiDB, verzeichnisB} {
				if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
					t.Errorf("Modus %s, Dienste %v (%s): %s ist angelegt", modus, b, art, name)
				}
			}
		}
	}
	zaehle(t, "Konfigurationen", faelle, gesamt)
	zaehle(t, "gebaute Konfigurationen", gebaut, gueltig)
}

// TestWahlSatz fängt A16: der Satz jedes Dienstes des Mock ist der seiner Version, in jedem der sieben Felder; kein Dienst trägt den
// Satz eines anderen (SPEC-139 Nr. 11 b, SPEC-175 Nr. 3). Tabelle 7 × 2 und Schleife über 14 Belegungen in den Modi mock und hybrid.
func TestWahlSatz(t *testing.T) {
	versionOrakelZaehlen(t)
	tabelle := bootstrap.SatzTabelle()
	zaehle(t, "Einträge der Tabelle", len(tabelle), 2)
	for version := range satzOrakel() {
		satz, ok := tabelle[version]
		if !ok {
			t.Errorf("die Tabelle hat keinen Eintrag %s", version)
		}
		vergleicheSatz(t, "Tabelle, Eintrag "+version, satz, version)
	}
	belegt := alleBelegungen()
	zaehle(t, "Belegungen", len(belegt), 14)
	for _, b := range belegt {
		for _, modus := range []string{"mock", "hybrid"} {
			if modus == "hybrid" && mitV2(b) {
				continue
			}
			v, _ := versionBau(t, modus, b, false)
			dienste := v.mock[0].Dienste
			zaehle(t, "Dienste des Mock", len(dienste), len(b))
			for i, d := range dienste {
				if d.Codecs == nil {
					t.Fatalf("Modus %s, Belegung %v, Dienst %d ohne Satz", modus, b, i)
				}
				vergleicheSatz(t, fmt.Sprintf("Modus %s, Belegung %v, Dienst %d", modus, b, i), *d.Codecs, b[i])
				for j := 0; j < i; j++ {
					if dienste[j].Codecs == d.Codecs {
						t.Errorf("Modus %s, Belegung %v: Dienst %d und %d teilen einen Satz", modus, b, j, i)
					}
				}
			}
		}
	}
}

// TestWahlFilterLeser fängt A17: der URL-Leser beider Sätze trägt den Leser von `$filter` (SPEC-139 Nr. 11 b, SPEC-160 Nr. 8,
// SPEC-173 Nr. 5). Menge 4: (4.0, Id eq 1), (4.0, Name eq 'a' and Id gt 1), (2.0, Id eq 1), (2.0, substringof, außerhalb des Umfangs von V2).
func TestWahlFilterLeser(t *testing.T) {
	v, _ := versionBau(t, "mock", []string{"4.0", "2.0"}, false)
	dienste := v.mock[0].Dienste
	zaehle(t, "Dienste des Mock", len(dienste), 2)
	if dienste[0].Codecs == nil || dienste[1].Codecs == nil {
		t.Fatal("ein Dienst des Mock ohne Satz")
	}
	faelle := []struct {
		dienst int
		filter string
		kompat bool
	}{
		{0, "Id eq 1", false}, {0, "Name eq 'a' and Id gt 1", false}, {1, "Id eq 1", false}, {1, "substringof('a',Name)", true},
	}
	zaehle(t, "Fälle", len(faelle), 4)
	for _, f := range faelle {
		u, err := dienste[f.dienst].Codecs.URL.ParseURL("/Products", "$filter="+f.filter)
		var kf *model.KompatFehler
		switch {
		case f.kompat && (!errors.As(err, &kf) || kf.Ursache != model.KompatFilter):
			t.Errorf("Version %s, %s: Fehler %v, erwartet KompatFehler mit Ursache filter", dienste[f.dienst].Codecs.Version, f.filter, err)
		case !f.kompat && (err != nil || u.Optionen.Filter == nil):
			t.Errorf("Version %s, %s: Ergebnis %+v, Fehler %v, erwartet einen Ausdruck", dienste[f.dienst].Codecs.Version, f.filter, u, err)
		}
	}
}

// rueckfallOrakel sind die Felder der Konfig von Mock und Import: je Feld das Paket der Version 4.0 (SPEC-139 Nr. 4, 11 b).
func rueckfallOrakel() map[string]string {
	return map[string]string{
		"Mock.MetadatenCodec": "csdl", "Mock.URLCodec": "urlparser", "Mock.Schreiber": "jsoncodec",
		"Import.MetadatenCodec": "csdl", "Import.DatenCodec": "jsoncodec",
	}
}

// TestWahlRueckfall fängt A18: MetadatenCodec, URLCodec und Schreiber der Konfig des Mock und MetadatenCodec und DatenCodec der Konfig
// des Imports sind in jeder Belegung die Werte der Version 4.0, auch bei Diensten, die alle 2.0 sind (SPEC-139 Nr. 4, 11 b).
// 5 Felder × 14 Belegungen; der Mock in den Modi mock und hybrid, der Import im Modus record.
func TestWahlRueckfall(t *testing.T) {
	soll := rueckfallOrakel()
	zaehle(t, "Felder", len(soll), 5)
	belegt := alleBelegungen()
	zaehle(t, "Belegungen", len(belegt), 14)
	for _, b := range belegt {
		for _, modus := range versionModi(b) {
			if modus == "replay" {
				continue
			}
			v, _ := versionBau(t, modus, b, false)
			got := map[string]string{}
			if modus == "record" {
				k := *v.importKfg
				got["Import.MetadatenCodec"], got["Import.DatenCodec"] = paketVon(k.MetadatenCodec), paketVon(k.DatenCodec)
			} else {
				k := v.mock[0]
				got["Mock.MetadatenCodec"], got["Mock.URLCodec"], got["Mock.Schreiber"] = paketVon(k.MetadatenCodec), paketVon(k.URLCodec), paketVon(k.Schreiber)
			}
			for feld, paket := range got {
				if paket != soll[feld] {
					t.Errorf("Modus %s, Belegung %v: %s ist aus %q, erwartet %q", modus, b, feld, paket, soll[feld])
				}
			}
			zaehle(t, "Felder im Modus "+modus, len(got), map[string]int{"record": 2, "mock": 3, "hybrid": 3}[modus])
		}
	}
}

// dienstOhneVersion verlangt, dass der Dienst des Moduls keine Version kennt: kein Feld namens Version (SPEC-139 Nr. 4, SPEC-170 Nr. 6).
func dienstOhneVersion(t *testing.T, name string, typ reflect.Type) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		if strings.Contains(typ.Field(i).Name, "Version") {
			t.Errorf("%s hat das Feld %s: der Dienst der Aufnahme und des Abspielens kennt keine Version", name, typ.Field(i).Name)
		}
	}
}

// TestWahlImport fängt A19: jeder Dienst des Imports trägt in Codecs den Satz seiner Version, auch bei 4.0; record.Dienst, replay.Dienst
// und hybrid.Dienst tragen keine Version und gleichen für jede Belegung denen von V4; in replay, mock und hybrid ist Import() nil und
// baselineimport.New wird nicht gerufen (SPEC-139 Nr. 4, 11 b, SPEC-170 Nr. 6). 14 Belegungen × 4 Modi, in hybrid nur die drei Belegungen aus 4.0 allein.
func TestWahlImport(t *testing.T) {
	versionOrakelZaehlen(t)
	dienstOhneVersion(t, "record.Dienst", reflect.TypeOf(record.Dienst{}))
	dienstOhneVersion(t, "replay.Dienst", reflect.TypeOf(replay.Dienst{}))
	dienstOhneVersion(t, "hybrid.Dienst", reflect.TypeOf(hybrid.Dienst{}))
	belegt := alleBelegungen()
	zaehle(t, "Belegungen", len(belegt), 14)
	modi := 0
	for _, b := range belegt {
		var recDienste []record.Dienst
		var repDienste []replay.Dienst
		var hybDienste []hybrid.Dienst
		for i := range b {
			id, root, up := fmt.Sprintf("s%d", i), fmt.Sprintf("/s%d", i), fmt.Sprintf("http://127.0.0.1:9/s%d", i)
			recDienste = append(recDienste, record.Dienst{ID: id, Root: root, UpstreamURL: up})
			repDienste = append(repDienste, replay.Dienst{ID: id, Root: root})
			hybDienste = append(hybDienste, hybrid.Dienst{Root: root, UpstreamURL: up})
		}
		for _, modus := range []string{"record", "replay", "mock", "hybrid"} {
			if modus == "hybrid" && mitV2(b) {
				continue
			}
			modi++
			v, h := versionBau(t, modus, b, false)
			if modus != "record" {
				if h.Import() != nil || v.verboten["import"] != 0 || v.importe != 0 {
					t.Errorf("Modus %s, Belegung %v: der Import ist gebaut (%d Aufrufe von baselineimport.New)", modus, b, v.verboten["import"])
				}
			}
			switch modus {
			case "record":
				if v.verboten["import"] != 1 || h.Import() == nil {
					t.Fatalf("Belegung %v: baselineimport.New %d-mal gerufen", b, v.verboten["import"])
				}
				dienste := v.importKfg.Dienste
				zaehle(t, "Dienste des Imports", len(dienste), len(b))
				for i, d := range dienste {
					if d.Codecs == nil {
						t.Errorf("Belegung %v: Dienst %d des Imports ohne Satz", b, i)
						continue
					}
					vergleicheSatz(t, fmt.Sprintf("Import, Belegung %v, Dienst %d", b, i), *d.Codecs, b[i])
					if d.ID != recDienste[i].ID || d.UpstreamURL != recDienste[i].UpstreamURL {
						t.Errorf("Belegung %v: Dienst %d des Imports ist %q %q", b, i, d.ID, d.UpstreamURL)
					}
				}
				if !reflect.DeepEqual(v.recorder.Dienste, recDienste) {
					t.Errorf("Belegung %v: record.Dienste %+v, erwartet %+v", b, v.recorder.Dienste, recDienste)
				}
			case "replay":
				if !reflect.DeepEqual(v.replay.Dienste, repDienste) {
					t.Errorf("Belegung %v: replay.Dienste %+v, erwartet %+v", b, v.replay.Dienste, repDienste)
				}
			case "hybrid":
				if len(v.hybrid) != 1 || !reflect.DeepEqual(v.hybrid[0].Dienste, hybDienste) {
					t.Errorf("Belegung %v: hybrid.Dienste %+v, erwartet %+v", b, v.hybrid, hybDienste)
				}
			}
		}
	}
	zaehle(t, "Belegungen × Modi", modi, 14*4-11)
}

// TestWahlEtag fängt A20: `Etag` ist je Dienst wahr genau bei ETAG_ENABLED und dem Satz 4.0 (SPEC-139 Nr. 11 a, SPEC-280 Nr. 2).
// 14 Belegungen × {wahr, falsch} in den Modi mock und hybrid.
func TestWahlEtag(t *testing.T) {
	belegt := alleBelegungen()
	zaehle(t, "Belegungen", len(belegt), 14)
	zellen := 0
	for _, b := range belegt {
		for _, modus := range []string{"mock", "hybrid"} {
			if modus == "hybrid" && mitV2(b) {
				continue
			}
			for _, an := range []bool{true, false} {
				zellen++
				v, _ := versionBau(t, modus, b, an)
				dienste := v.mock[0].Dienste
				zaehle(t, "Dienste des Mock", len(dienste), len(b))
				for i, d := range dienste {
					if want := an && b[i] == "4.0"; d.Etag != want {
						t.Errorf("Modus %s, Belegung %v, ETAG_ENABLED %v, Dienst %d: Etag %v, erwartet %v", modus, b, an, i, d.Etag, want)
					}
				}
			}
		}
	}
	zaehle(t, "Zellen", zellen, (14+3)*2)
}

// stelleVersion ist eine Funktion einer Quelldatei, in der das Feld Version oder ODataVersion gelesen oder geschrieben wird.
type stelleVersion struct{ datei, funktion string }

// versionStellen sind die Funktionen des Pakets, die die Version eines Dienstes oder Services lesen oder schreiben, mit der Zahl der
// Zugriffe: die Prüfung der Konfiguration, die Wahl in version.go und die Anlage mit dem Vergleich in lebenszyklus.go (SPEC-139 Nr. 6,
// 11 b, SPEC-175 Nr. 3). Ein Zugriff an einer anderen Stelle oder ein weiterer in einer dieser Funktionen ändert die Tabelle.
func versionStellen() map[stelleVersion]int {
	return map[stelleVersion]int{
		{"konfig_version.go", "pruefeVersionModus"}: 1,
		{"konfig_dienste.go", "einzelDienst"}:       1,
		{"konfig_dienste.go", "dateiEintrag"}:       1,
		{"version.go", "etagFuerSatz"}:              1,
		{"version.go", "wirksameDienste"}:           2,
		{"version.go", "waehle"}:                    2,
		{"lebenszyklus.go", "legeDiensteAn"}:        1,
		{"lebenszyklus.go", "abweichungen"}:         2,
	}
}

// quellDateien sind die Quelldateien des Pakets ohne Tests, wie der Compiler sie sieht (go/build); das Verzeichnis führt keine weitere.
func quellDateien(t *testing.T) []string {
	t.Helper()
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatal(err)
	}
	dateien := append(slices.Clone(pkg.GoFiles), pkg.IgnoredGoFiles...)
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	im := 0
	for _, e := range eintraege {
		if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			im++
		}
	}
	if len(dateien) == 0 || len(dateien) != im {
		t.Fatalf("der Compiler führt %d Quelldateien (%v), das Verzeichnis %d", len(dateien), dateien, im)
	}
	return dateien
}

// zaehleVersionStellen zählt je Funktion die Selektoren Version und ODataVersion der geparsten Dateien.
func zaehleVersionStellen(asts map[string]*ast.File) map[stelleVersion]int {
	ist := map[stelleVersion]int{}
	for d, f := range asts {
		for _, decl := range f.Decls {
			name := ""
			if fd, ok := decl.(*ast.FuncDecl); ok {
				name = fd.Name.Name
			}
			ast.Inspect(decl, func(n ast.Node) bool {
				if sel, ok := n.(*ast.SelectorExpr); ok && (sel.Sel.Name == "Version" || sel.Sel.Name == "ODataVersion") {
					ist[stelleVersion{d, name}]++
				}
				return true
			})
		}
	}
	return ist
}

// TestVersionLiteraleAnEinerStelle fängt A21 in zwei Prüfungen über jede Quelldatei des Pakets (die Menge ist die des Compilers, gegen
// das Verzeichnis gezählt): (1) allein version.go und konfig_version.go nennen eine Version, als Literal 4.0 oder 2.0 oder als Name
// der zwei Konstanten, die ein Literal tragen; (2) jede Lesung und jede Zuweisung der Felder Version und ODataVersion steht in einer
// Funktion der Tabelle versionStellen, mit genau ihrer Zahl, so dass ein Zweig über die Version an einer anderen Stelle (if, switch,
// case, Index, Schlüssel einer Map) eine neue Zeile der Tabelle braucht (SPEC-139 Nr. 11 b, SPEC-175 Nr. 3). Nicht gefangen: ein Zweig
// über eine Kopie des Werts innerhalb einer Funktion der Tabelle und ein Zugriff über reflect.
func TestVersionLiteraleAnEinerStelle(t *testing.T) {
	dateien := quellDateien(t)
	istVersion := func(l *ast.BasicLit) bool {
		s, err := strconv.Unquote(l.Value)
		return l.Kind == token.STRING && err == nil && (s == "4.0" || s == "2.0")
	}
	asts := map[string]*ast.File{}
	konstanten := map[string]bool{}
	for _, d := range dateien {
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		asts[d] = f
		ast.Inspect(f, func(n ast.Node) bool {
			if vs, ok := n.(*ast.ValueSpec); ok {
				for i, w := range vs.Values {
					if l, ok := w.(*ast.BasicLit); ok && istVersion(l) && i < len(vs.Names) {
						konstanten[vs.Names[i].Name] = true
					}
				}
			}
			return true
		})
	}
	if len(konstanten) != 2 {
		t.Fatalf("Konstanten mit einem Literal der Versionen: %v, erwartet zwei", konstanten)
	}
	var nennen []string
	for _, d := range dateien {
		genannt := false
		ast.Inspect(asts[d], func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.BasicLit:
				genannt = genannt || istVersion(n)
			case *ast.Ident:
				genannt = genannt || konstanten[n.Name]
			}
			return true
		})
		if genannt {
			nennen = append(nennen, d)
		}
	}
	slices.Sort(nennen)
	if want := []string{"konfig_version.go", "version.go"}; !slices.Equal(nennen, want) {
		t.Errorf("Dateien, die eine Version nennen: %v, erwartet %v", nennen, want)
	}
	soll, ist := versionStellen(), zaehleVersionStellen(asts)
	if !maps.Equal(ist, soll) {
		for stelle, n := range ist {
			if soll[stelle] != n {
				t.Errorf("%s, Funktion %q: %d Zugriffe auf Version oder ODataVersion, die Tabelle nennt %d", stelle.datei, stelle.funktion, n, soll[stelle])
			}
		}
		for stelle, n := range soll {
			if _, da := ist[stelle]; !da {
				t.Errorf("%s, Funktion %q: kein Zugriff, die Tabelle nennt %d", stelle.datei, stelle.funktion, n)
			}
		}
	}
}

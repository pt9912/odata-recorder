package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"maps"
	"slices"
	"strconv"
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
		if h != nil {
			_ = h.Herunterfahren(ctx)
		}
	}
}

// anlageZeile ist eine Zeile der Meldung einer Abweichung des gespeicherten Services (SPEC-139 Nr. 6, 8 c): der Schlüssel und der Grund.
type anlageZeile struct{ schluessel, zeile string }

// anlageZeilen sind die Zeilen je Schlüssel, in der Reihenfolge `service_root`, `odata_version`, aus dem Wortlaut der Spezifikation.
func anlageZeilen() []anlageZeile {
	return []anlageZeile{
		{"service_root", "ODR-E2001 service_root: gespeicherter Service gleicher id hat einen anderen service_root"},
		{"odata_version", "ODR-E2001 odata_version: gespeicherter Service gleicher id hat eine andere odata_version"},
	}
}

// anlageSchluessel liest aus lebenszyklus.go die Schlüssel, die abweichungen an seine Liste hängt, in der Reihenfolge der Quelle.
func anlageSchluessel(t *testing.T) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "lebenszyklus.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var schluessel []string
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "abweichungen" {
			ast.Inspect(fd, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok && len(c.Args) > 1 {
					if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "append" {
						for _, a := range c.Args[1:] {
							if l, ok := a.(*ast.BasicLit); ok && l.Kind == token.STRING {
								s, _ := strconv.Unquote(l.Value)
								schluessel = append(schluessel, s)
							}
						}
					}
				}
				return true
			})
		}
	}
	return schluessel
}

// anlageFall ist ein Start gegen eine Ablage, in der jeder Dienst schon steht: b sind die konfigurierten Versionen, gespeichert die
// Services der Ablage, stelle der Dienst, dessen gespeicherter Service abweicht (-1: keiner), zeilen die erwartete Meldung.
type anlageFall struct {
	b           []string
	gespeichert []model.Service
	stelle      int
	zeilen      []string
}

// anlageFaelle erzeugt die Menge zu A14 über die Dienste: für N = 1, 2, 3 Dienste, jede Stelle des abweichenden Dienstes, jede Belegung
// der Versionen der anderen aus {4.0, 2.0}, jede nicht leere Teilmenge der Schlüssel, die abweichen, und jedes Paar aus gespeicherter
// Version {4.0, 2.0} und konfigurierter {4.0, 2.0, fehlend}, das zur Teilmenge passt (die Version weicht ab genau dann, wenn
// odata_version in der Teilmenge liegt): je Teilmenge 3 Paare. Dazu die 14 Belegungen ohne Abweichung.
func anlageFaelle(zeilen []anlageZeile) []anlageFall {
	var aus []anlageFall
	konfiguriert := []string{"4.0", "2.0", ""}
	bitRoot := slices.IndexFunc(zeilen, func(z anlageZeile) bool { return z.schluessel == "service_root" })
	bitVersion := slices.IndexFunc(zeilen, func(z anlageZeile) bool { return z.schluessel == "odata_version" })
	dienst := func(i int, root, version string) model.Service {
		return model.Service{ID: fmt.Sprintf("s%d", i), ServiceRoot: root, ODataVersion: version}
	}
	for n := 1; n <= 3; n++ {
		for _, b := range belegungen(n) {
			gespeichert := make([]model.Service, n)
			for i, v := range b {
				gespeichert[i] = dienst(i, fmt.Sprintf("/s%d", i), v)
			}
			aus = append(aus, anlageFall{b, gespeichert, -1, nil})
		}
		for stelle := 0; stelle < n; stelle++ {
			for _, nachbarn := range belegungen(n - 1) {
				for maske := 1; maske < 1<<len(zeilen); maske++ {
					for _, g := range versionW() {
						for _, konf := range konfiguriert {
							if (cmpLeer(konf) != g) != (maske&(1<<bitVersion) != 0) {
								continue
							}
							f := anlageFall{stelle: stelle}
							f.b = slices.Insert(slices.Clone(nachbarn), stelle, konf)
							for i := range f.b {
								v := cmpLeer(f.b[i])
								f.gespeichert = append(f.gespeichert, dienst(i, fmt.Sprintf("/s%d", i), v))
							}
							f.gespeichert[stelle] = dienst(stelle, fmt.Sprintf("/s%d", stelle), g)
							if maske&(1<<bitRoot) != 0 {
								f.gespeichert[stelle].ServiceRoot = fmt.Sprintf("/x%d", stelle)
							}
							for i, z := range zeilen {
								if maske&(1<<i) != 0 {
									f.zeilen = append(f.zeilen, z.zeile)
								}
							}
							aus = append(aus, f)
						}
					}
				}
			}
		}
	}
	return aus
}

// TestAnlageAbweichungJeDienst fängt A14 an jedem Dienst der Liste (SPEC-139 Nr. 6 und 8 c, SPEC-107 Nr. 14: für jeden Dienst; die
// Version geht gegen die des Dienstes): weicht der gespeicherte Service eines Dienstes ab, gleich an welcher Stelle der Liste er steht,
// wie viele Dienste die Liste führt und welche Versionen die anderen tragen, verhindert das den Start mit den Zeilen dieses Dienstes
// in der Reihenfolge der Schlüssel und je einem Logeintrag mit seiner Kennung; die Anlage ruft Create und Get je Dienst bis zu ihm,
// nicht darüber hinaus, und ändert die Ablage nicht; ohne Abweichung startet jede Belegung. Die Schlüssel der Abweichung liest der Test
// aus abweichungen in lebenszyklus.go und verlangt, dass jeder eine Zeile des Orakels hat. Menge: 2^k - 1 nicht leere Teilmengen der k
// Schlüssel × 3 Paare der Versionen × (Stelle, Belegung der anderen): N = 1: 1, N = 2: 2 × 2, N = 3: 3 × 4 = 17, dazu 14 ohne Abweichung,
// im Modus replay; die vier Modi führt TestAnlageAbweichungModi.
func TestAnlageAbweichungJeDienst(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	zeilen := anlageZeilen()
	quelle := anlageSchluessel(t)
	var orakel []string
	for _, z := range zeilen {
		orakel = append(orakel, z.schluessel)
	}
	if !slices.Equal(quelle, orakel) {
		t.Fatalf("abweichungen nennt die Schlüssel %v, das Orakel führt %v: jeder Schlüssel braucht eine Zeile und Fälle", quelle, orakel)
	}
	faelle := anlageFaelle(zeilen)
	mit, ohne := 0, 0
	for _, f := range faelle {
		if f.stelle < 0 {
			ohne++
		} else {
			mit++
		}
	}
	zaehle(t, "Fälle mit Abweichung", mit, (hoch(2, len(zeilen))-1)*3*(1+2*2+3*4))
	zaehle(t, "Fälle ohne Abweichung", ohne, 2+4+8)
	for _, f := range faelle {
		pruefeAnlageFall(ctx, t, f, "replay")
	}
}

// TestAnlageAbweichungModi fängt A14 in jedem Modus (SPEC-139 Nr. 6: die Anlage ist Schritt 2 in allen vier Modi): dieselben Prüfungen
// wie TestAnlageAbweichungJeDienst für die Fälle, in denen der letzte von drei Diensten abweicht und kein Dienst 2.0 konfiguriert
// ist (hybrid nimmt keinen). Menge: 6 Fälle (3 nicht leere Teilmengen der Schlüssel × 2 Paare der Versionen) × 4 Modi = 24.
func TestAnlageAbweichungModi(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	var faelle []anlageFall
	for _, f := range anlageFaelle(anlageZeilen()) {
		if len(f.b) == 3 && f.stelle == 2 && !mitV2(f.b) {
			faelle = append(faelle, f)
		}
	}
	modi := []string{"record", "replay", "mock", "hybrid"}
	zaehle(t, "Fälle", len(faelle), 6)
	zaehle(t, "Modi", len(modi), 4)
	for _, modus := range modi {
		for _, f := range faelle {
			pruefeAnlageFall(ctx, t, f, modus)
		}
	}
}

// pruefeAnlageFall führt einen Start im Modus modus gegen die Ablage des Falls und prüft Meldung, Logeinträge, die Aufrufe des
// Port-Fakes und die Ablage.
func pruefeAnlageFall(ctx context.Context, t *testing.T, f anlageFall, modus string) {
	t.Helper()
	dir := t.TempDir()
	db := oeffneAblage(ctx, t, dir)
	for _, s := range f.gespeichert {
		if err := db.Metadata().Create(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()
	var logPuffer syncPuffer
	v := neuVersionHaken()
	h, err := lauf(ctx, t, versionKonfig(dir, modus, f.b), slog.New(slog.NewTextHandler(&logPuffer, nil)), v.bauteile())
	was := fmt.Sprintf("Modus %s, Versionen %q, Abweichung an Stelle %d, gespeichert %+v", modus, f.b, f.stelle, f.gespeichert)
	switch {
	case len(f.zeilen) == 0 && err != nil:
		t.Errorf("%s: Hochfahren: %v", was, err)
	case len(f.zeilen) > 0 && (err == nil || bootstrap.Meldung(err) != strings.Join(f.zeilen, "\n")):
		t.Errorf("%s: Hochfahren: %v, erwartet %q", was, err, f.zeilen)
	}
	var eintraege []map[string]string
	for _, l := range logZeilen(t, logPuffer.String()) {
		if l["level"] == "ERROR" || l["ereignis"] != "" {
			eintraege = append(eintraege, l)
		}
	}
	if len(eintraege) != len(f.zeilen) {
		t.Errorf("%s: %d Logeinträge, erwartet %d:\n%s", was, len(eintraege), len(f.zeilen), logPuffer.String())
	}
	for _, e := range eintraege {
		soll := map[string]string{
			"level": "ERROR", "msg": "start: gespeicherter Service weicht von der Konfiguration ab",
			"ereignis": "dienst_abweichend", "error_code": "ODR-E2001", "service_id": fmt.Sprintf("s%d", f.stelle),
		}
		if !maps.Equal(e, soll) {
			t.Errorf("%s: Logeintrag %v, erwartet %v", was, e, soll)
		}
	}
	erreicht := len(f.b)
	if f.stelle >= 0 {
		erreicht = f.stelle + 1
	}
	var ids []string
	wirksam := make([]string, len(f.b))
	for i, s := range f.gespeichert {
		ids = append(ids, s.ID)
		wirksam[i] = cmpLeer(f.b[i])
	}
	ids = ids[:erreicht]
	if got := v.dienstZaehler("Create"); got != erreicht {
		t.Errorf("%s: %d Aufrufe von Create, erwartet %d", was, got, erreicht)
	}
	if want := versionDienste(wirksam)[:erreicht]; !slices.Equal(v.dienstAnlagen, want) {
		t.Errorf("%s: Create erhielt %+v, erwartet %+v", was, v.dienstAnlagen, want)
	}
	if !slices.Equal(v.dienstGets, ids) {
		t.Errorf("%s: Get für %v, erwartet %v", was, v.dienstGets, ids)
	}
	for _, op := range []string{"List", "SetMetadata"} {
		if got := v.dienstZaehler(op); got != 0 {
			t.Errorf("%s: %d Aufrufe von %s, erwartet 0", was, got, op)
		}
	}
	if h != nil {
		_ = h.Herunterfahren(ctx)
	}
	db = oeffneAblage(ctx, t, dir)
	if got, err := db.Metadata().List(ctx); err != nil || !slices.Equal(got, f.gespeichert) {
		t.Errorf("%s: die Ablage hält danach %+v (%v)", was, got, err)
	}
	_ = db.Close()
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

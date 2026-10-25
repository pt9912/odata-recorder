package baselineimport_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestZeileF10Phasen: die 111 herstellbaren Paare der 16 Verstöße: erwartet der Verstoß der früheren
// Phase, die Zähler der späteren Phasen sind 0; dazu jeder Verstoß allein.
func TestZeileF10Phasen(t *testing.T) {
	if len(verstoesse()) != 16 {
		t.Fatalf("Verstöße %d", len(verstoesse()))
	}
	paare := 0
	for i := 1; i <= 16; i++ {
		for j := i + 1; j <= 16; j++ {
			if nichtZugleich(i, j) {
				continue
			}
			paare++
			laufePaar(t, i, j)
		}
		laufePaar(t, i, 0)
	}
	if paare != 111 {
		t.Errorf("Paare %d, erwartet 111", paare)
	}
}

func laufePaar(t *testing.T, i, j int) {
	t.Helper()
	w := neu(t)
	w.sessionR()
	a := anfrage()
	vs := verstoesse()
	for _, v := range vs {
		if v.nr == i || v.nr == j {
			v.setup(w, &a)
		}
	}
	_, err := w.importiere(a)
	if err == nil || !vs[i-1].trifft(err) {
		t.Errorf("Paar (%d %s, %d): Fehler %v", i, vs[i-1].name, j, err)
	}
	for op, n := range w.stand() {
		erlaubt := 1 << 30
		switch {
		case op == "Blobs.Open" && i < 11:
			erlaubt = 0
		case op == "Blobs.Open" && i < 13:
			erlaubt = 1
		case op != "Blobs.Open" && phase(op) > i:
			erlaubt = 0
		}
		if n > erlaubt {
			t.Errorf("Paar (%d, %d): %s %d-mal, erlaubt %d", i, j, op, n, erlaubt)
		}
	}
}

// TestZeileF6NichtsSchreibendesVorDemLetztenLesen: zwölf Verstöße vor dem Schreiben (96 Zähler gleich 0)
// und fünf Schreibfälle (Menge 17).
func TestZeileF6NichtsSchreibendesVorDemLetztenLesen(t *testing.T) {
	type fall struct {
		name  string
		setup func(w *welt, a *inbound.ImportAnfrage)
	}
	var faelle []fall
	vs := verstoesse()
	for _, v := range vs[3:13] { // Session unbekannt bis Fehler einer Antwort
		faelle = append(faelle, fall{v.name, nil})
	}
	faelle = append(faelle,
		fall{"Open (Antwort) scheitert", func(w *welt, a *inbound.ImportAnfrage) { w.p.scheitere("Blobs.Open", 2, errPlatte) }},
		fall{"Konflikt", func(w *welt, a *inbound.ImportAnfrage) {
			w.daten.nachRef[w.exchanges.m[sitzungID][2].ResponseBlob] = datenVorgabe{entity: produkt(1, "ANDERS", "1.5")}
		}},
	)
	if len(faelle) != 12 {
		t.Fatalf("Verstöße %d", len(faelle))
	}
	nullen := []string{"Blobs.Begin", "Write", "Commit", "Abort", "Imports.Publish", "Metadaten.SetMetadata", "IDs.SnapshotID", "Uhr.Now"}
	for i, f := range faelle {
		w := neu(t)
		w.sessionR()
		a := anfrage()
		if i < 10 {
			vs := verstoesse()
			vs[i+3].setup(w, &a)
		} else {
			f.setup(w, &a)
		}
		if _, err := w.importiere(a); err == nil {
			t.Errorf("%s: kein Fehler", f.name)
		}
		for _, op := range nullen {
			if n := w.p.anzahl(op); n != 0 {
				t.Errorf("%s: %s %d-mal", f.name, op, n)
			}
		}
	}
	schreib := []struct {
		name        string
		op          string
		abort       int
		setMetadata int
		publish     int
		erfolg      bool
	}{
		{"Begin scheitert", "Blobs.Begin", 0, 0, 0, false},
		{"Write scheitert", "Write", 1, 0, 0, false},
		{"Commit scheitert", "Commit", 1, 0, 0, false},
		{"Publish scheitert", "Imports.Publish", 0, 0, 1, false},
		{"Commit gelingt", "", 0, 1, 1, true},
	}
	for _, s := range schreib {
		w := neu(t)
		w.sessionR()
		if s.op != "" {
			w.p.scheitere(s.op, 1, errPlatte)
		}
		_, err := w.importiere(anfrage())
		if (err == nil) != s.erfolg {
			t.Errorf("%s: %v", s.name, err)
		}
		if w.p.anzahl("Abort") != s.abort || w.p.anzahl("Metadaten.SetMetadata") != s.setMetadata || w.p.anzahl("Imports.Publish") != s.publish {
			t.Errorf("%s: Abort %d, SetMetadata %d, Publish %d", s.name, w.p.anzahl("Abort"), w.p.anzahl("Metadaten.SetMetadata"), w.p.anzahl("Imports.Publish"))
		}
		if s.erfolg && (w.p.anzahl("IDs.SnapshotID") != 1 || w.p.anzahl("Uhr.Now") != 1) {
			t.Errorf("%s: SnapshotID %d, Now %d", s.name, w.p.anzahl("IDs.SnapshotID"), w.p.anzahl("Uhr.Now"))
		}
	}
}

// TestZeileF9Reihenfolge: die Folge der Operationen des neuen Imports und der frühen Wiederholung.
func TestZeileF9Reihenfolge(t *testing.T) {
	w := neu(t)
	w.sessionR()
	mustImport(t, w, anfrage())
	neuFolge := []string{
		"Imports.Get", "Sessions.Get", "Exchanges.ListBySession",
		"Blobs.Open", "ParseMetadata", "Close",
		"Blobs.Open", "ParseCollection", "Close",
		"Blobs.Open", "ParseEntity", "Close",
		"IDs.SnapshotID", "Uhr.Now", "Blobs.Begin", "Write", "Commit", "Imports.Publish", "Metadaten.SetMetadata",
	}
	if got := w.p.folge(); !slices.Equal(got, neuFolge) {
		t.Errorf("neu: %v", got)
	}
	vor := len(w.p.folge())
	mustImport(t, w, anfrage())
	if got := w.p.folge()[vor:]; !slices.Equal(got, []string{"Imports.Get", "Snapshots.Get", "Metadaten.SetMetadata"}) {
		t.Errorf("Wiederholung: %v", got)
	}
}

// TestZeileI9Get: Get prüft ctx, reicht die Kennung unverändert und bildet Fehler mit dem Gegenstand
// Import (Menge 18).
func TestZeileI9Get(t *testing.T) {
	w := neu(t).start()
	var keinCtx context.Context
	if _, err := w.imp.Get(keinCtx, "x"); !errors.Is(err, model.ErrUngueltig) || !ohneCode(err) || w.gesamt() != 0 {
		t.Errorf("ctx nil: %v, Aufrufe %d", err, w.gesamt())
	}
	for _, id := range []string{"", "x", "A"} {
		w := neu(t).start()
		_, err := w.imp.Get(context.Background(), id)
		erwarteFehler(t, "Get "+id, err, model.CodeE7005, 404)
		if a := w.p.argsVon("Imports.Get"); len(a) != 1 || a[0][0] != id {
			t.Errorf("Get %q: Argumente %v", id, a)
		}
	}
	n := 0
	for _, f := range portFehler() {
		n++
		w := neu(t).start()
		w.p.scheitere("Imports.Get", 1, f.err)
		_, err := w.imp.Get(context.Background(), "imp-1")
		code, status := erwarteterCode("Import", f.name)
		erwarteFehler(t, "Get mit "+f.name, err, code, status)
		if !errors.Is(err, f.err) {
			t.Errorf("Get mit %s: Ursache nicht erreichbar", f.name)
		}
	}
	for _, k := range kontextFehler() {
		n++
		w := neu(t).start()
		w.p.scheitere("Imports.Get", 1, k)
		if _, err := w.imp.Get(context.Background(), "imp-1"); err != k || !ohneCode(err) {
			t.Errorf("Get mit %v: %v", k, err)
		}
	}
	if n != 14 {
		t.Errorf("Menge %d", n)
	}
	// Gutfall: der gespeicherte Import.
	w = neu(t)
	w.sessionR()
	r := mustImport(t, w, anfrage())
	got, err := w.imp.Get(context.Background(), "imp-1")
	if err != nil || got != r.Import {
		t.Errorf("Get: %v", err)
	}
}

// formG10 ist eine Form des Fehlers einer Antwort (SPEC-127 Nr. 4 f, 9, 10).
type formG10 struct {
	name  string
	setze func(w *welt, e model.Exchange, entity bool)
	ist   func(err error, w *welt, e model.Exchange) bool
}

func formenG10() []formG10 {
	df := &model.DatenFehler{Ursache: model.DatenWert, Ort: "/x"}
	return []formG10{
		{"DatenFehler", func(w *welt, e model.Exchange, _ bool) { w.daten.nachRef[e.ResponseBlob] = datenVorgabe{err: df} },
			func(err error, _ *welt, _ model.Exchange) bool {
				var g *model.DatenFehler
				return errors.As(err, &g) && g == df
			}},
		{"Konflikt wert", func(w *welt, e model.Exchange, entity bool) {
			w.daten.nachRef[e.ResponseBlob] = vorgabeFuer(entity, produkt(1, "ANDERS", "1.5"))
		}, func(err error, _ *welt, e model.Exchange) bool {
			var k *model.KonfliktFehler
			return errors.As(err, &k) && k.Ursache == model.KonfliktWert && k.Zweite == e.ID && k.Property == "Name"
		}},
		{"Konflikt typ", func(w *welt, e model.Exchange, entity bool) {
			g := produkt(1, "Milch", "1.5")
			g.Typ = typGeschenk
			w.daten.nachRef[e.ResponseBlob] = vorgabeFuer(entity, g)
		}, func(err error, _ *welt, e model.Exchange) bool {
			var k *model.KonfliktFehler
			return errors.As(err, &k) && k.Ursache == model.KonfliktTyp && k.Zweite == e.ID
		}},
		{"Blob fehlt", func(w *welt, e model.Exchange, _ bool) { delete(w.blobs.inhalt, e.ResponseBlob) },
			func(err error, _ *welt, _ model.Exchange) bool {
				return errors.Is(err, outbound.ErrBlobFehlt) && istCode(model.CodeE3001)(err)
			}},
		{"Fehler von src", func(w *welt, e model.Exchange, _ bool) {
			w.daten.nachRef[e.ResponseBlob] = datenVorgabe{err: outbound.ErrBlobBeschaedigt}
		}, func(err error, _ *welt, _ model.Exchange) bool {
			return errors.Is(err, outbound.ErrBlobBeschaedigt) && istCode(model.CodeE3001)(err)
		}},
		{"Kontextfehler", func(w *welt, e model.Exchange, _ bool) {
			w.daten.nachRef[e.ResponseBlob] = datenVorgabe{err: context.Canceled}
		},
			func(err error, _ *welt, _ model.Exchange) bool { return err == context.Canceled && ohneCode(err) }},
	}
}

func vorgabeFuer(entity bool, e model.Entity) datenVorgabe {
	if entity {
		return datenVorgabe{entity: e}
	}
	return datenVorgabe{seite: seite(e)}
}

// TestZeileG10FehlerZweierAntworten: 12 Formen (sechs mal Collection und Entity), jedes geordnete
// Paar (144): der erste Fehler der Folge entscheidet, spätere Exchanges werden nicht geöffnet.
func TestZeileG10FehlerZweierAntworten(t *testing.T) {
	formen := formenG10()
	n := 0
	for _, f1 := range formen {
		for _, e1 := range []bool{false, true} {
			for _, f2 := range formen {
				for _, e2 := range []bool{false, true} {
					n++
					w := neu(t)
					b := w.standard()
					b.collection("Products", "", seite(produkt(1, "Milch", "1.5")))
					mach := func(entity bool) model.Exchange {
						pfad := "/Products"
						if entity {
							pfad = "/Products(1)"
						}
						ex := b.roh(pfad, "")
						w.daten.nachRef[ex.ResponseBlob] = datenVorgabe{seite: seite(produkt(1, "Milch", "1.5")), entity: produkt(1, "Milch", "1.5")}
						return ex
					}
					x1, x2 := mach(e1), mach(e2)
					f1.setze(w, x1, e1)
					f2.setze(w, x2, e2)
					_, err := w.importiere(anfrage())
					name := fmt.Sprintf("%s/%v dann %s/%v", f1.name, e1, f2.name, e2)
					if err == nil || !f1.ist(err, w, x1) {
						t.Errorf("%s: %v", name, err)
					}
					if slices.ContainsFunc(w.p.argsVon("Blobs.Open"), func(a []any) bool { return a[0] == x2.ResponseBlob }) {
						t.Errorf("%s: der spätere Exchange wurde geöffnet", name)
					}
				}
			}
		}
	}
	if n != 144 {
		t.Errorf("Paare %d", n)
	}
}

// TestZeileF6WriteMitFalscherLaenge: ein Write, das ohne Fehler 0, eine negative oder eine zu große Länge
// meldet, beendet den Import mit ODR-E9001 und Abort, statt zu kreisen oder zu zerschneiden.
func TestZeileF6WriteMitFalscherLaenge(t *testing.T) {
	for name, laenge := range map[string]func(int) int{
		"nichts": func(int) int { return 0 }, "negativ": func(int) int { return -1 }, "zu groß": func(n int) int { return n + 1 },
	} {
		w := neu(t)
		w.sessionR()
		w.blobs.laenge = laenge
		fertig := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fertig <- fmt.Errorf("Panik: %v", r)
				}
			}()
			_, err := w.importiere(anfrage())
			fertig <- err
		}()
		select {
		case err := <-fertig:
			erwarteFehler(t, name, err, model.CodeE9001, 500)
			if w.p.anzahl("Abort") != 1 || w.p.anzahl("Imports.Publish") != 0 || w.p.anzahl("Write") != 1 {
				t.Errorf("%s: Abort %d, Publish %d, Write %d", name, w.p.anzahl("Abort"), w.p.anzahl("Imports.Publish"), w.p.anzahl("Write"))
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("%s: der Import kreist", name)
		}
	}
}

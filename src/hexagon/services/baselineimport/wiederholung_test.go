package baselineimport_test

import (
	"errors"
	"maps"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// sessionR ist die Session des Referenzfalls: $metadata, die Collection Products mit zwei Entities
// und die Einzel-Entity Products(1).
func (w *welt) sessionR() *bauer {
	b := w.standard()
	b.collection("Products", "", seite(produkt(1, "Milch", "1.5"), produkt(2, "Brot", "2")))
	b.entity("Products", "1", produkt(1, "Milch", "1.5"))
	return b
}

func (w *welt) stand() map[string]int {
	w.p.mu.Lock()
	defer w.p.mu.Unlock()
	return maps.Clone(w.p.zaehler)
}

// unterschied nennt die Operationen, deren Zähler seit vor gewachsen sind, mit dem Zuwachs.
func (w *welt) unterschied(vor map[string]int) map[string]int {
	d := map[string]int{}
	for op, n := range w.stand() {
		if n != vor[op] {
			d[op] = n - vor[op]
		}
	}
	return d
}

func mustImport(t *testing.T, w *welt, a inbound.ImportAnfrage) inbound.ImportErgebnis {
	t.Helper()
	r, err := w.importiere(a)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	return r
}

// TestZeileB2Wiederholung: gleiche import_id und gleicher Hash liefern den gespeicherten Import
// byte-gleich, mit Neu falsch, vor jedem Lesen, und setzen die Metadaten trotzdem.
func TestZeileB2Wiederholung(t *testing.T) {
	w := neu(t)
	w.sessionR()
	erst := mustImport(t, w, anfrage())
	if !erst.Neu {
		t.Fatal("der erste Aufruf ist nicht neu")
	}
	vor := w.stand()
	zweit := mustImport(t, w, anfrage())
	if zweit.Neu {
		t.Error("Neu bei der Wiederholung")
	}
	if zweit.Import.ReportJSON != erst.Import.ReportJSON || zweit.Import.ID != erst.Import.ID {
		t.Error("die Wiederholung liefert einen anderen Import")
	}
	d := w.unterschied(vor)
	want := map[string]int{"Imports.Get": 1, "Snapshots.Get": 1, "Metadaten.SetMetadata": 1}
	if !maps.Equal(d, want) {
		t.Errorf("Zähler der Wiederholung %v, erwartet %v", d, want)
	}
	snap := w.snapshots.m[erst.Import.BaselineSnapshotID]
	got := w.metadaten.gesetzt
	if len(got) != 2 || got[1] != (setzung{dienstID, snap.MetadataHash, snap.MetadataHash}) {
		t.Errorf("SetMetadata %v", got)
	}
	if w.p.verbotenGerufen() != 0 {
		t.Error("eine Operation, die der Dienst nie rufen darf")
	}
}

// TestZeileB3AndererHash: bei bekannter import_id ist jede Änderung der Anfrage ODR-E6102, ohne
// weiteren Aufruf außer Imports.Get (Menge 4).
func TestZeileB3AndererHash(t *testing.T) {
	andere := map[string]func(*inbound.ImportAnfrage){
		"ServiceID":      func(a *inbound.ImportAnfrage) { a.ServiceID = "zweiter" },
		"Session":        func(a *inbound.ImportAnfrage) { a.SourceSessionIDs = []string{"ses-2"} },
		"MetadataHash":   func(a *inbound.ImportAnfrage) { a.MetadataHash = "sha256:" + hexNullen(64) },
		"MetadataHash 2": func(a *inbound.ImportAnfrage) { a.MetadataHash = "sha256:" + hexNullen(63) + "1" },
	}
	for name, f := range andere {
		w := neu(t)
		w.konfig.Dienste = append(w.konfig.Dienste, baselineimport.Dienst{ID: "zweiter", UpstreamURL: upstream})
		w.sessionR()
		mustImport(t, w, anfrage())
		vor := w.stand()
		a := anfrage()
		f(&a)
		_, err := w.importiere(a)
		fe := fehlerMit(t, name, err, model.CodeE6102, 409)
		if fe != nil && !errors.Is(fe, outbound.ErrBereitsVorhanden) {
			t.Errorf("%s: Ursache %v", name, fe.Ursache)
		}
		if d := w.unterschied(vor); !maps.Equal(d, map[string]int{"Imports.Get": 1}) {
			t.Errorf("%s: Zähler %v", name, d)
		}
	}
}

// TestZeileB4ImportIDNachFehler: nach einem fehlgeschlagenen Import ist die import_id frei (Menge 6).
func TestZeileB4ImportIDNachFehler(t *testing.T) {
	type fall struct {
		name  string
		bruch func(w *welt, b *bauer) func()
	}
	faelle := []fall{
		{"Session unbekannt", func(w *welt, b *bauer) func() {
			s := w.sessions.m[sitzungID]
			w.sessions.loesche(sitzungID)
			return func() { w.sessions.m[sitzungID] = s }
		}},
		{"kein $metadata", func(w *welt, b *bauer) func() {
			w.exchanges.m[sitzungID] = w.exchanges.m[sitzungID][1:]
			return func() { b.metadaten("<edmx 1/>") }
		}},
		{"DatenFehler", func(w *welt, b *bauer) func() {
			ref := w.exchanges.m[sitzungID][1].ResponseBlob
			alt := w.daten.nachRef[ref]
			w.daten.nachRef[ref] = datenVorgabe{err: &model.DatenFehler{Ursache: model.DatenWert}}
			return func() { w.daten.nachRef[ref] = alt }
		}},
		{"Konflikt", func(w *welt, b *bauer) func() {
			ref := w.exchanges.m[sitzungID][2].ResponseBlob
			alt := w.daten.nachRef[ref]
			e := produkt(1, "ANDERS", "1.5")
			w.daten.nachRef[ref] = datenVorgabe{entity: e}
			return func() { w.daten.nachRef[ref] = alt }
		}},
		{"Begin", func(w *welt, b *bauer) func() {
			w.p.scheitere("Blobs.Begin", 1, errPlatte)
			return func() {}
		}},
		{"Publish", func(w *welt, b *bauer) func() {
			w.p.scheitere("Imports.Publish", 1, errPlatte)
			return func() {}
		}},
	}
	for _, f := range faelle {
		w := neu(t)
		b := w.sessionR()
		heile := f.bruch(w, b)
		if _, err := w.importiere(anfrage()); err == nil {
			t.Errorf("%s: kein Fehler", f.name)
			continue
		}
		if len(w.imports.m) != 0 {
			t.Errorf("%s: Imports hält %d Einträge", f.name, len(w.imports.m))
		}
		heile()
		r, err := w.importiere(anfrage())
		if err != nil || !r.Neu {
			t.Errorf("%s: zweiter Aufruf %v, Neu %v", f.name, err, r.Neu)
		}
	}
}

// TestZeileF7SetMetadataFehler: ein Fehler von SetMetadata ist der Fehler des Imports, die
// Baseline bleibt, die Wiederholung heilt (nach neuem und nach „nicht neu“).
func TestZeileF7SetMetadataFehler(t *testing.T) {
	// Nach neuem Publish: ErrNichtGefunden ist ODR-E5005.
	w := neu(t)
	w.sessionR()
	w.p.scheitere("Metadaten.SetMetadata", 1, outbound.ErrNichtGefunden)
	r, err := w.importiere(anfrage())
	erwarteFehler(t, "nach neuem Publish", err, model.CodeE5005, 404)
	if r != (inbound.ImportErgebnis{}) {
		t.Error("Ergebnis bei Fehler ist nicht der Nullwert")
	}
	if w.p.anzahl("Imports.Publish") != 1 || len(w.imports.m) != 1 {
		t.Fatalf("Publish %d, Imports %d", w.p.anzahl("Imports.Publish"), len(w.imports.m))
	}
	erstReport := w.imports.pubs[0].Import.ReportJSON
	vor := w.stand()
	r2, err := w.importiere(anfrage())
	if err != nil || r2.Neu || r2.Import.ReportJSON != erstReport {
		t.Fatalf("Heilung: %v, Neu %v", err, r2.Neu)
	}
	if d := w.unterschied(vor); d["Blobs.Begin"] != 0 || d["Imports.Publish"] != 0 || d["Metadaten.SetMetadata"] != 1 {
		t.Errorf("Zähler der Heilung %v", d)
	}
	if len(w.metadaten.gesetzt) != 1 {
		t.Errorf("gesetzt %v", w.metadaten.gesetzt)
	}
	// Nach „nicht neu“: ein fremder Fehler ist ODR-E9001.
	w = neu(t)
	w.sessionR()
	hash := baselineimport.AnfrageHash(anfrage())
	w.imports.m["andere-id"] = model.Import{ID: "andere-id", RequestHash: hash, BaselineSnapshotID: "snp-x", ServiceID: dienstID, CreatedAt: t0(), ReportJSON: "{}"}
	w.p.scheitere("Metadaten.SetMetadata", 1, errPlatte)
	_, err = w.importiere(anfrage())
	erwarteFehler(t, "nach nicht neu", err, model.CodeE9001, 500)
	if w.p.anzahl("Imports.Publish") != 1 {
		t.Errorf("Publish %d", w.p.anzahl("Imports.Publish"))
	}
	r3, err := w.importiere(anfrage())
	if err != nil || r3.Neu || len(w.metadaten.gesetzt) != 1 {
		t.Errorf("zweiter Aufruf: %v, Neu %v, gesetzt %v", err, r3.Neu, w.metadaten.gesetzt)
	}
}

// TestZeileF8SetMetadataArgumente: SetMetadata(Service, Referenz, Referenz) mit der Referenz des
// Dokuments auf allen drei Wegen (Menge 3), genau einmal.
func TestZeileF8SetMetadataArgumente(t *testing.T) {
	ref := func(w *welt) string { return w.exchanges.m[sitzungID][0].ResponseBlob }
	pruefe := func(name string, w *welt, ref string) {
		t.Helper()
		if len(w.metadaten.gesetzt) != 1 || w.metadaten.gesetzt[0] != (setzung{dienstID, ref, ref}) {
			t.Errorf("%s: SetMetadata %v", name, w.metadaten.gesetzt)
		}
	}
	// Publish neu. Die Referenz des Dokuments ist weder StateBlob noch RequestHash.
	w := neu(t)
	w.sessionR()
	r := mustImport(t, w, anfrage())
	pruefe("neu", w, ref(w))
	snap := w.snapshots.m[r.Import.BaselineSnapshotID]
	if ref(w) == snap.StateBlob || ref(w) == r.Import.RequestHash || ref(w) == baselineimport.AnfrageHash(anfrage()) {
		t.Fatal("Fixture: die Referenz des Dokuments gleicht einer anderen Referenz")
	}
	// Publish nicht neu.
	w = neu(t)
	w.sessionR()
	w.imports.m["andere-id"] = model.Import{ID: "andere-id", RequestHash: baselineimport.AnfrageHash(anfrage()), BaselineSnapshotID: "snp-x", ServiceID: dienstID, CreatedAt: t0(), ReportJSON: "{}"}
	if r := mustImport(t, w, anfrage()); r.Neu {
		t.Error("Neu bei nicht neu")
	}
	pruefe("nicht neu", w, ref(w))
	// Frühe Wiederholung.
	w = neu(t)
	w.sessionR()
	mustImport(t, w, anfrage())
	w.metadaten.gesetzt = nil
	mustImport(t, w, anfrage())
	pruefe("Wiederholung", w, ref(w))
}

// TestZeileI7ErgebnisIstPublish: das Ergebnis ist der Import, den Publish lieferte, und Neu sein
// Wahrheitswert (Menge 2).
func TestZeileI7ErgebnisIstPublish(t *testing.T) {
	fremd := model.Import{ID: "fremd", RequestHash: "h", BaselineSnapshotID: "snp-9", ServiceID: dienstID, CreatedAt: t0(), ReportJSON: `{"x":1}`}
	for _, neuWert := range []bool{true, false} {
		w := neu(t)
		w.sessionR()
		w.imports.uebersteuere = func(model.Publication) (model.Import, bool, error) { return fremd, neuWert, nil }
		r := mustImport(t, w, anfrage())
		if r.Import != fremd || r.Neu != neuWert {
			t.Errorf("Neu %v: Ergebnis %+v", neuWert, r)
		}
	}
}

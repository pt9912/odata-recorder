package baselineimport_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// bericht liest den ReportJSON des Imports als Objekt.
func bericht(t *testing.T, r inbound.ImportErgebnis) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(r.Import.ReportJSON), &m); err != nil {
		t.Fatalf("ReportJSON: %v", err)
	}
	return m
}

// uebersprungen liefert die Zählung der Gründe aus dem Report.
func uebersprungen(t *testing.T, r inbound.ImportErgebnis) map[string]int {
	t.Helper()
	ex := bericht(t, r)["report"].(map[string]any)["exchanges"].(map[string]any)
	out := map[string]int{}
	for _, g := range ex["skipped"].([]any) {
		z := g.(map[string]any)
		out[z["reason"].(string)] = int(z["count"].(float64))
	}
	return out
}

type kandidatFall struct {
	name     string
	kandidat bool
	mod      aenderung
}

// kandidatenFaelle ist die Menge von D1: die Basis, je eine Abweichung (26 ohne, 7 mit Kandidat).
func kandidatenFaelle() []kandidatFall {
	var out []kandidatFall
	nein := func(name string, a aenderung) { out = append(out, kandidatFall{name, false, a}) }
	ja := func(name string, a aenderung) { out = append(out, kandidatFall{name, true, a}) }
	for _, m := range []string{"POST", "HEAD", "get"} {
		nein("Methode "+m, func(e *model.Exchange) { e.Method = m })
	}
	for _, p := range []string{"/$metadata/", "/$Metadata", "/$metadata/x", "/x/$metadata", "/$metadata%20"} {
		nein("Pfad "+p, func(e *model.Exchange) { e.Path = p })
	}
	for _, c := range []model.CaptureStatus{model.CaptureIncomplete, model.CaptureFailed} {
		nein("capture "+string(c), func(e *model.Exchange) { e.Capture = c })
	}
	for _, s := range []int{201, 204, 304, 404, 500} {
		nein(fmt.Sprint("Status ", s), func(e *model.Exchange) { e.ResponseStatus = s })
	}
	for _, b := range []model.BodyState{model.BodyAbsent, model.BodyDiscarded} {
		nein("Body "+string(b), func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = b, "" })
	}
	nein("redaction applied", func(e *model.Exchange) { e.ResponseRedact = model.RedactionApplied })
	for _, q := range []string{"a=b", "$select=x", "$format=xml&a=b"} {
		nein("Query "+q, func(e *model.Exchange) { e.RawQuery = q })
	}
	ct := func(v ...string) aenderung {
		return func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{"Content-Type": v} }
	}
	nein("Content-Type fehlt", func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{} })
	for _, v := range []string{"application/xml;charset=iso-8859-1", "application/json", "text/plain", "application/xmlx"} {
		nein("Content-Type "+v, ct(v))
	}
	for _, q := range []string{"", "$format=xml", "$format=application%2Fxml"} {
		ja("Query "+q, func(e *model.Exchange) { e.RawQuery = q })
	}
	for _, v := range []string{"application/xml", "text/xml", "APPLICATION/XML", "application/xml;charset=utf-8"} {
		ja("Content-Type "+v, ct(v))
	}
	return out
}

// TestZeileD1Kandidat: das Prädikat des Kandidaten (Menge 33: 26 Abweichungen ohne, 7 mit Kandidat).
func TestZeileD1Kandidat(t *testing.T) {
	faelle := kandidatenFaelle()
	if len(faelle) != 33 {
		t.Fatalf("Menge %d, erwartet 33", len(faelle))
	}
	for _, f := range faelle {
		w := neu(t)
		w.sitzung().metadaten("<edmx/>", f.mod)
		_, err := w.importiere(anfrage())
		if f.kandidat != (err == nil) {
			t.Errorf("%s: Kandidat %v, Fehler %v", f.name, f.kandidat, err)
		}
		if !f.kandidat && !errors.Is(err, baselineimport.ErrKeineMetadaten) {
			t.Errorf("%s: %v", f.name, err)
		}
	}
}

// TestZeileD2OhneKandidat: ohne Kandidat ODR-E6004 mit ErrKeineMetadaten vor jedem Lesen (Menge 27).
func TestZeileD2OhneKandidat(t *testing.T) {
	pruefe := func(name string, w *welt) {
		t.Helper()
		_, err := w.importiere(anfrage())
		f := fehlerMit(t, name, err, model.CodeE6004, 422)
		if f != nil && (!errors.Is(f, baselineimport.ErrKeineMetadaten) || !errors.Is(f, model.ErrMetadatenUngueltig)) {
			t.Errorf("%s: Ursache %v", name, f.Ursache)
		}
		for _, op := range []string{"Blobs.Open", "ParseMetadata", "Blobs.Begin"} {
			if n := w.p.anzahl(op); n != 0 {
				t.Errorf("%s: %s %d-mal", name, op, n)
			}
		}
	}
	pruefe("leere Session", neu(t))
	n := 1
	for _, f := range kandidatenFaelle() {
		if f.kandidat {
			continue
		}
		w := neu(t)
		w.sitzung().metadaten("<edmx/>", f.mod)
		pruefe(f.name, w)
		n++
	}
	if n != 27 {
		t.Errorf("Menge %d, erwartet 27", n)
	}
}

// TestZeileD3MehrereGleiche: mehrere Kandidaten mit gleicher Referenz sind ein Dokument, das mit der
// kleinsten sequence_no (Menge 6).
func TestZeileD3MehrereGleiche(t *testing.T) {
	for _, n := range []int{2, 3} {
		for _, lage := range []string{"Anfang", "zwischen", "Ende"} {
			w := neu(t)
			b := w.sitzung()
			var ids []string
			meta := func() { ids = append(ids, b.metadaten("<edmx/>").ID) }
			entity := func() { b.collection("Products", "", seite(produkt(1, "Milch", "1.5"))) }
			switch lage {
			case "Anfang":
				for i := 0; i < n; i++ {
					meta()
				}
				entity()
			case "zwischen":
				entity()
				for i := 0; i < n; i++ {
					meta()
					entity()
				}
			case "Ende":
				entity()
				for i := 0; i < n; i++ {
					meta()
				}
			}
			r := mustImport(t, w, anfrage())
			name := fmt.Sprint(n, " ", lage)
			ref := w.exchanges.m[sitzungID][slices.IndexFunc(w.exchanges.m[sitzungID], func(e model.Exchange) bool { return e.ID == ids[0] })].ResponseBlob
			opens := 0
			for _, a := range w.p.argsVon("Blobs.Open") {
				if a[0] == ref {
					opens++
				}
			}
			if opens != 1 || w.p.anzahl("ParseMetadata") != 1 {
				t.Errorf("%s: Open %d, ParseMetadata %d", name, opens, w.p.anzahl("ParseMetadata"))
			}
			if got := uebersprungen(t, r)["metadata_doppelt"]; got != n-1 {
				t.Errorf("%s: metadata_doppelt %d, erwartet %d", name, got, n-1)
			}
			prov := bericht(t, r)["report"].(map[string]any)["provenance"].([]any)
			var meta0 []string
			for _, p := range prov {
				if p.(map[string]any)["kind"] == "metadata" {
					meta0 = append(meta0, p.(map[string]any)["exchange_id"].(string))
				}
			}
			if !slices.Equal(meta0, ids[:1]) {
				t.Errorf("%s: Quelle des Dokuments %v, erwartet %v", name, meta0, ids[:1])
			}
		}
	}
}

// TestZeileD4Verschiedene: verschiedene Referenzen sind ODR-E6004 mit ErrMetadatenAbweichend, in
// jeder Anordnung, ohne Lesen; der Hash wählt nicht aus (Menge 10).
func TestZeileD4Verschiedene(t *testing.T) {
	refA := refVon([]byte("A"))
	n := 0
	for _, ordnung := range []string{"AB", "BA", "AAB", "ABA", "BAA"} {
		for _, mitHash := range []bool{false, true} {
			w := neu(t)
			b := w.sitzung()
			for _, c := range ordnung {
				b.metadaten(string(c))
			}
			a := anfrage()
			if mitHash {
				a.MetadataHash = refA
			}
			_, err := w.importiere(a)
			f := fehlerMit(t, ordnung, err, model.CodeE6004, 422)
			if f != nil && !errors.Is(f, baselineimport.ErrMetadatenAbweichend) {
				t.Errorf("%s: Ursache %v", ordnung, f.Ursache)
			}
			if w.p.anzahl("Blobs.Open") != 0 {
				t.Errorf("%s mit Hash %v: Open %d-mal", ordnung, mitHash, w.p.anzahl("Blobs.Open"))
			}
			n++
		}
	}
	if n != 10 {
		t.Errorf("Menge %d", n)
	}
}

// TestZeileD5HashPrueft: der Hash der Anfrage prüft das Dokument vor dem Lesen (Menge 4).
func TestZeileD5HashPrueft(t *testing.T) {
	ref := refVon([]byte("<edmx/>"))
	anders := ref[:len(ref)-1] + "0"
	if ref[len(ref)-1] == '0' {
		anders = ref[:len(ref)-1] + "1"
	}
	faelle := []struct {
		name, hash string
		fehler     bool
	}{
		{"leer", "", false}, {"gleich", ref, false},
		{"andere gültige Referenz", refVon([]byte("anderes")), true}, {"ein anderes Zeichen", anders, true},
	}
	for _, f := range faelle {
		w := neu(t)
		w.sitzung().metadaten("<edmx/>")
		a := anfrage()
		a.MetadataHash = f.hash
		_, err := w.importiere(a)
		if !f.fehler {
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
			}
			continue
		}
		fe := fehlerMit(t, f.name, err, model.CodeE6004, 422)
		if fe != nil && !errors.Is(fe, baselineimport.ErrMetadatenHash) {
			t.Errorf("%s: Ursache %v", f.name, fe.Ursache)
		}
		if n := w.p.anzahl("Blobs.Open"); n != 0 {
			t.Errorf("%s: Open %d-mal", f.name, n)
		}
	}
}

// TestZeileD6LesenDesDokuments: Open einmal, ParseMetadata mit Kontext, dem Strom von Open und der
// Grenze, Close einmal vor der ersten Antwort (Menge 3).
func TestZeileD6LesenDesDokuments(t *testing.T) {
	for _, grenze := range []int64{1, 8 << 20, 1<<63 - 1} {
		w := neu(t)
		w.konfig.MaxCaptureBody = grenze
		b := w.sitzung()
		b.metadaten("<edmx/>")
		b.collection("Products", "", seite(produkt(1, "Milch", "1.5")))
		mustImport(t, w, anfrage())
		if len(w.meta.aufrufe) != 1 {
			t.Fatalf("Grenze %d: %d Aufrufe", grenze, len(w.meta.aufrufe))
		}
		a := w.meta.aufrufe[0]
		strome := w.blobs.strome()
		if a.grenze != grenze || a.src != io.Reader(strome[0]) {
			t.Errorf("Grenze %d: Aufruf %+v", grenze, a)
		}
		if got := strings.Join(w.p.folge()[3:8], ","); got != "Blobs.Open,ParseMetadata,Close,Blobs.Open,ParseCollection" {
			t.Errorf("Grenze %d: Folge %s", grenze, got)
		}
		for i, s := range strome {
			if s.geschlossen() != 1 {
				t.Errorf("Grenze %d: Strom %d %d-mal geschlossen", grenze, i, s.geschlossen())
			}
		}
		if c := w.p.ctxVon("ParseMetadata"); len(c) != 1 || c[0] == nil {
			t.Error("Kontext fehlt")
		}
	}
}

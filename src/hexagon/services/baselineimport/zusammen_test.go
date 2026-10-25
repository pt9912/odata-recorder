package baselineimport_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// manifest liest das Manifest, das der Import als Blob schrieb.
func manifest(t *testing.T, w *welt) map[string]any {
	t.Helper()
	if len(w.blobs.writers) == 0 {
		t.Fatal("kein Manifest geschrieben")
	}
	var m map[string]any
	if err := json.Unmarshal(w.blobs.writers[len(w.blobs.writers)-1].bytes(), &m); err != nil {
		t.Fatalf("Manifest: %v", err)
	}
	return m
}

func entitaeten(m map[string]any) []map[string]any {
	var out []map[string]any
	for _, e := range m["entities"].([]any) {
		out = append(out, e.(map[string]any))
	}
	return out
}

func schluessel(es []map[string]any) []string {
	var out []string
	for _, e := range es {
		out = append(out, e["entity_set"].(string)+"|"+e["typed_key"].(string))
	}
	return out
}

// eP ist eine Entity des Sets Products mit dem Schlüssel 1 und den Properties props nach Id.
func eP(props ...model.PropertyWert) model.Entity {
	return ent(typProdukt, "1", append([]model.PropertyWert{pw("Id", model.ArtInt32, "1")}, props...)...)
}

// beobachtung legt eine Antwort der Spur Entity oder Collection mit e an.
func (b *bauer) beobachtung(set string, collection bool, es ...model.Entity) model.Exchange {
	if collection {
		return b.collection(set, "", seite(es...))
	}
	return b.entity(set, "1", es[0])
}

// TestZeileG1Identitaet: die Identität ist (Set, Schlüssel) (Menge 8).
func TestZeileG1Identitaet(t *testing.T) {
	for _, k := range []string{"1", "'1'", "'a''b'", "Id=1,Region='EU'"} {
		w := neu(t)
		b := w.standard()
		e := ent(typProdukt, k, pw("Id", model.ArtInt32, "1"))
		b.collection("Products", "", seite(e))
		b.collection("Products", "$top=5", seite(e))
		mustImport(t, w, anfrage())
		if got := schluessel(entitaeten(manifest(t, w))); !slices.Equal(got, []string{"Products|" + k}) {
			t.Errorf("zweimal im selben Set, Schlüssel %q: %v", k, got)
		}
		w = neu(t)
		b = w.standard()
		b.collection("Products", "", seite(e))
		b.collection("Tags", "", seite(e))
		mustImport(t, w, anfrage())
		if got := schluessel(entitaeten(manifest(t, w))); !slices.Equal(got, []string{"Products|" + k, "Tags|" + k}) {
			t.Errorf("in zwei Sets, Schlüssel %q: %v", k, got)
		}
	}
}

// TestZeileG2Dubletten: identische Dubletten werden zusammengeführt, die Provenienz nennt jeden
// Exchange (Menge 6).
func TestZeileG2Dubletten(t *testing.T) {
	e := eP(pw("Name", model.ArtString, "Milch"))
	faelle := []struct {
		name      string
		bau       func(b *bauer) []model.Exchange
		eintraege int
		beob      int
	}{
		{"Collection und Collection", func(b *bauer) []model.Exchange {
			return []model.Exchange{b.beobachtung("Products", true, e), b.beobachtung("Products", true, e)}
		}, 2, 2},
		{"Collection und Entity", func(b *bauer) []model.Exchange {
			return []model.Exchange{b.beobachtung("Products", true, e), b.beobachtung("Products", false, e)}
		}, 2, 2},
		{"Entity und Collection", func(b *bauer) []model.Exchange {
			return []model.Exchange{b.beobachtung("Products", false, e), b.beobachtung("Products", true, e)}
		}, 2, 2},
		{"Entity und Entity", func(b *bauer) []model.Exchange {
			return []model.Exchange{b.beobachtung("Products", false, e), b.beobachtung("Products", false, e)}
		}, 2, 2},
		{"zwei Elemente einer Antwort", func(b *bauer) []model.Exchange {
			return []model.Exchange{b.beobachtung("Products", true, e, e)}
		}, 1, 2},
		{"drei Exchanges", func(b *bauer) []model.Exchange {
			return []model.Exchange{b.beobachtung("Products", true, e), b.beobachtung("Products", false, e), b.beobachtung("Products", true, e)}
		}, 3, 3},
	}
	for _, f := range faelle {
		w := neu(t)
		b := w.standard()
		exs := f.bau(b)
		r := mustImport(t, w, anfrage())
		es := entitaeten(manifest(t, w))
		if len(es) != 1 {
			t.Errorf("%s: %d Entities", f.name, len(es))
			continue
		}
		prov := es[0]["provenance"].([]any)
		if len(prov) != f.eintraege {
			t.Errorf("%s: %d Einträge der Provenienz, erwartet %d", f.name, len(prov), f.eintraege)
		}
		seen := map[string]bool{}
		for i, p := range prov {
			pm := p.(map[string]any)
			seen[pm["exchange_id"].(string)] = true
			if pm["session_id"] != sitzungID || fmt.Sprint(pm["projection"]) != "[Id Name]" || !strings.HasPrefix(pm["hash"].(string), "sha256:") {
				t.Errorf("%s: Eintrag %d: %v", f.name, i, pm)
			}
			if i > 0 && prov[i-1].(map[string]any)["exchange_id"].(string) >= pm["exchange_id"].(string) {
				t.Errorf("%s: Provenienz nicht aufsteigend", f.name)
			}
		}
		for _, x := range exs {
			if !seen[x.ID] {
				t.Errorf("%s: Exchange %s fehlt in der Provenienz", f.name, x.ID)
			}
		}
		rep := bericht(t, r)["report"].(map[string]any)["entities"].(map[string]any)
		if rep["distinct"].(float64) != 1 || int(rep["observations"].(float64)) != f.beob {
			t.Errorf("%s: report.entities %v", f.name, rep)
		}
	}
}

func arten() []model.WertArt {
	return []model.WertArt{model.ArtNull, model.ArtString, model.ArtBoolean, model.ArtInt32, model.ArtInt64, model.ArtDecimal, model.ArtGUID, model.ArtDate, model.ArtDateTimeOffset}
}

func textVon(a model.WertArt) string {
	if a == model.ArtNull {
		return ""
	}
	return "1"
}

// zweiBeobachtungen importiert zwei Entity-Antworten derselben Entity; es liefert den Fehler und die Welt.
func zweiBeobachtungen(t *testing.T, e1, e2 model.Entity) (*welt, error) {
	t.Helper()
	w := neu(t)
	b := w.standard()
	b.beobachtung("Products", false, e1)
	b.beobachtung("Products", false, e2)
	_, err := w.importiere(anfrage())
	return w, err
}

// TestZeileG3Gleichheit: zwei Werte sind gleich genau bei gleicher Art und gleichem Text (Menge 89).
func TestZeileG3Gleichheit(t *testing.T) {
	n, gleich := 0, 0
	pruefe := func(a, b model.PropertyWert, konflikt bool) {
		n++
		_, err := zweiBeobachtungen(t, eP(a), eP(b))
		var k *model.KonfliktFehler
		if konflikt != errors.As(err, &k) || (!konflikt && err != nil) {
			t.Errorf("%v gegen %v: %v, erwartet Konflikt %v", a.Wert, b.Wert, err, konflikt)
		}
		if konflikt && (k.Ursache != model.KonfliktWert || k.Property != "P") {
			t.Errorf("%v gegen %v: %+v", a.Wert, b.Wert, k)
		}
	}
	for _, a := range arten() {
		for _, b := range arten() {
			if a == b {
				gleich++
			}
			pruefe(pw("P", a, textVon(a)), pw("P", b, textVon(b)), a != b)
		}
	}
	for _, a := range arten() {
		if a == model.ArtNull {
			continue
		}
		pruefe(pw("P", a, "1"), pw("P", a, "2"), true)
	}
	if n != 89 || gleich != 9 {
		t.Errorf("Menge %d, gleiche Werte %d", n, gleich)
	}
}

// TestZeileG4Zustaende: die 16 Paare der Zustände einer Property und 256 Paare von Teilmengen (Menge 272).
func TestZeileG4Zustaende(t *testing.T) {
	zustaende := []string{"nicht genannt", "null", "Wert", "redigiert"}
	prop := func(z string) []model.PropertyWert {
		switch z {
		case "null":
			return []model.PropertyWert{pw("Name", model.ArtNull, "")}
		case "Wert":
			return []model.PropertyWert{pw("Name", model.ArtString, "x")}
		case "redigiert":
			return []model.PropertyWert{pw("Name", model.ArtRedigiert, "")}
		}
		return nil
	}
	n := 0
	for _, z1 := range zustaende {
		for _, z2 := range zustaende {
			n++
			w, err := zweiBeobachtungen(t, eP(prop(z1)...), eP(prop(z2)...))
			bekannt := func(z string) bool { return z == "null" || z == "Wert" }
			konflikt := bekannt(z1) && bekannt(z2) && z1 != z2
			name := z1 + " / " + z2
			if konflikt != (err != nil) {
				t.Errorf("%s: Fehler %v, erwartet Konflikt %v", name, err, konflikt)
				continue
			}
			if konflikt {
				continue
			}
			e := entitaeten(manifest(t, w))[0]
			payload := e["payload"].(map[string]any)
			sel := fmt.Sprint(e["selected_properties"])
			maskiert := 0
			switch {
			case bekannt(z1) || bekannt(z2):
				want := "x"
				var wantV any = want
				if z1 == "null" || z2 == "null" {
					wantV = nil
				}
				if v, ok := payload["Name"]; !ok || v != wantV || sel != "[Id Name]" {
					t.Errorf("%s: payload %v, selected %s", name, payload, sel)
				}
			default:
				if _, ok := payload["Name"]; ok || sel != "[Id]" {
					t.Errorf("%s: payload %v, selected %s", name, payload, sel)
				}
				if z1 == "redigiert" || z2 == "redigiert" {
					maskiert = 1
				}
			}
			for k, v := range payload {
				if strings.Contains(fmt.Sprint(k, v), "redigiert") {
					t.Errorf("%s: Marker im Payload %v", name, payload)
				}
			}
			var rep map[string]any
			if err := json.Unmarshal([]byte(w.imports.m["imp-1"].ReportJSON), &rep); err != nil {
				t.Fatal(err)
			}
			if int(rep["report"].(map[string]any)["masked_properties"].(float64)) != maskiert {
				t.Errorf("%s: masked_properties %v, erwartet %d", name, rep["report"].(map[string]any)["masked_properties"], maskiert)
			}
		}
	}
	namen := []string{"P1", "P2", "P3", "P4"}
	teil := func(maske int) []model.PropertyWert {
		var out []model.PropertyWert
		for i, nm := range namen {
			if maske&(1<<i) != 0 {
				out = append(out, pw(nm, model.ArtString, "v"+nm))
			}
		}
		return out
	}
	for m1 := 0; m1 < 16; m1++ {
		for m2 := 0; m2 < 16; m2++ {
			n++
			w, err := zweiBeobachtungen(t, eP(teil(m1)...), eP(teil(m2)...))
			if err != nil {
				t.Fatalf("Teilmengen %b %b: %v", m1, m2, err)
			}
			want := []string{"Id"}
			for i, nm := range namen {
				if (m1|m2)&(1<<i) != 0 {
					want = append(want, nm)
				}
			}
			e := entitaeten(manifest(t, w))[0]
			if fmt.Sprint(e["selected_properties"]) != fmt.Sprint(want) || len(e["payload"].(map[string]any)) != len(want) {
				t.Errorf("Teilmengen %b %b: %v, erwartet %v", m1, m2, e["selected_properties"], want)
			}
		}
	}
	if n != 272 {
		t.Errorf("Menge %d, erwartet 272", n)
	}
}

// TestZeileG2Teilprojektionen: die Provenienz nennt je Exchange seine Projektion und den Hash der
// Kanonform der Werte genau dieser Properties (SPEC-127 Nr. 11 d).
func TestZeileG2Teilprojektionen(t *testing.T) {
	w := neu(t)
	b := w.standard()
	b.beobachtung("Products", false, eP(pw("Name", model.ArtString, "Milch")))
	b.beobachtung("Products", false, eP(pw("Price", model.ArtDecimal, "1.5")))
	mustImport(t, w, anfrage())
	prov := entitaeten(manifest(t, w))[0]["provenance"].([]any)
	soll := []struct{ projektion, hash string }{
		{"[Id Name]", "sha256:559cc18a74a669ac9d17f143a441328a2b6d4ef71c49b8de5966ee3df07a20a3"},
		{"[Id Price]", "sha256:6dea862a107a92cf84601e25a1aaa32fc0cf92f66d6a17ab7418faf9d9974275"},
	}
	if len(prov) != 2 {
		t.Fatalf("Provenienz %v", prov)
	}
	for i, s := range soll {
		p := prov[i].(map[string]any)
		if fmt.Sprint(p["projection"]) != s.projektion || p["hash"] != s.hash {
			t.Errorf("Eintrag %d: %v, erwartet %v", i, p, s)
		}
	}
}

// TestZeileI1ProvenienzAufsteigend: die Quellen stehen aufsteigend nach sequence_no, auch wenn das
// $metadata nicht der erste Exchange der Session ist.
func TestZeileI1ProvenienzAufsteigend(t *testing.T) {
	w := neu(t)
	b := w.sitzung()
	b.collection("Products", "", seite(produkt(1, "Milch", "1.5")))
	b.metadaten("<edmx 1/>")
	b.entity("Products", "1", produkt(1, "Milch", "1.5"))
	r := mustImport(t, w, anfrage())
	var ids []string
	for _, p := range bericht(t, r)["report"].(map[string]any)["provenance"].([]any) {
		ids = append(ids, p.(map[string]any)["exchange_id"].(string))
	}
	if !slices.Equal(ids, []string{"exc-001", "exc-002", "exc-003"}) {
		t.Errorf("Quellen des Reports %v", ids)
	}
	ids = nil
	for _, p := range manifest(t, w)["provenance"].([]any) {
		ids = append(ids, p.(map[string]any)["exchange_id"].(string))
	}
	if !slices.Equal(ids, []string{"exc-001", "exc-002", "exc-003"}) {
		t.Errorf("Quellen des Manifests %v", ids)
	}
}

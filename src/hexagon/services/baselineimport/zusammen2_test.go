package baselineimport_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

func eMarker(a, b, c string) model.Entity {
	return ent(typProdukt, "'"+marker+"'", pw("Id", model.ArtInt32, "1"), pw("A", model.ArtString, a), pw("B", model.ArtString, b), pw("C", model.ArtString, c))
}

// TestZeileG5KonfliktWert: ODR-E6101 mit *KonfliktFehler, nichts geschrieben, kein Wert und kein
// Schlüssel im Fehler (Menge 12).
func TestZeileG5KonfliktWert(t *testing.T) {
	n := 0
	layouts := []string{"zwei Antworten", "zwei Elemente einer Antwort", "Collection dann Entity", "Entity dann Collection"}
	for pos, prop := range []string{"A", "B", "C"} {
		for _, layout := range layouts {
			n++
			w := neu(t)
			b := w.standard()
			e1 := eMarker(marker+"a", marker+"b", marker+"c")
			vals := []string{marker + "a", marker + "b", marker + "c"}
			vals[pos] = marker + "ANDERS"
			e2 := eMarker(vals[0], vals[1], vals[2])
			var x1, x2 model.Exchange
			switch layout {
			case "zwei Antworten":
				x1, x2 = b.beobachtung("Products", false, e1), b.beobachtung("Products", false, e2)
			case "zwei Elemente einer Antwort":
				x1 = b.beobachtung("Products", true, e1, e2)
				x2 = x1
			case "Collection dann Entity":
				x1, x2 = b.beobachtung("Products", true, e1), b.beobachtung("Products", false, e2)
			default:
				x1, x2 = b.beobachtung("Products", false, e1), b.beobachtung("Products", true, e2)
			}
			_, err := w.importiere(anfrage())
			name := prop + " " + layout
			fe := fehlerMit(t, name, err, model.CodeE6101, 422)
			var k *model.KonfliktFehler
			if !errors.As(err, &k) || k.Ursache != model.KonfliktWert || k.EntitySet != "Products" || k.Property != prop || k.Erste != x1.ID || k.Zweite != x2.ID {
				t.Errorf("%s: %+v", name, k)
			}
			if !errors.Is(err, model.ErrKonflikt) || !errors.Is(err, model.ErrDatenUngueltig) {
				t.Errorf("%s: errors.Is", name)
			}
			for _, text := range []string{err.Error(), fmt.Sprintf("%v", err), fmt.Sprintf("%v %+v %#v", k, k, k), fmt.Sprint(fe)} {
				if strings.Contains(text, marker) {
					t.Errorf("%s: der Fehlertext trägt einen Wert oder Schlüssel: %s", name, text)
				}
			}
			if w.p.anzahl("Blobs.Begin") != 0 || w.p.anzahl("Imports.Publish") != 0 {
				t.Errorf("%s: etwas geschrieben", name)
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge %d", n)
	}
}

// TestZeileG6KonfliktTyp: ein Konflikt des Typs geht dem des Werts vor (Menge 6).
func TestZeileG6KonfliktTyp(t *testing.T) {
	mit := func(typ, name string) model.Entity {
		e := eP(pw("Name", model.ArtString, name))
		e.Typ = typ
		return e
	}
	typen := []struct {
		name   string
		t1, t2 string
		typ    bool
	}{
		{"Typ des Sets gegen abgeleiteten", typProdukt, typGeschenk, true},
		{"abgeleiteter gegen abgeleiteten", typGeschenk, "Shop.Special", true},
		{"gleicher Typ", typProdukt, typProdukt, false},
	}
	for _, f := range typen {
		for _, verschieden := range []bool{false, true} {
			n2 := "Milch"
			if verschieden {
				n2 = "Brot"
			}
			_, err := zweiBeobachtungen(t, mit(f.t1, "Milch"), mit(f.t2, n2))
			name := fmt.Sprint(f.name, " verschieden=", verschieden)
			var k *model.KonfliktFehler
			switch {
			case f.typ:
				if !errors.As(err, &k) || k.Ursache != model.KonfliktTyp || k.Property != "" || k.EntitySet != "Products" || k.Erste != "exc-002" || k.Zweite != "exc-003" {
					t.Errorf("%s: %v %+v", name, err, k)
				}
			case verschieden:
				if !errors.As(err, &k) || k.Ursache != model.KonfliktWert || k.Property != "Name" {
					t.Errorf("%s: %v %+v", name, err, k)
				}
			default:
				if err != nil {
					t.Errorf("%s: %v", name, err)
				}
			}
		}
	}
}

// TestZeileG6TypnamenAehnlich: zwei Typnamen sind gleich genau bei gleichen Bytes (SPEC-127 Nr. 10 f). Sieben
// Namen, die einander ähneln, alle 49 geordneten Paare mal Werte {gleich, verschieden} (98): typ genau bei
// verschiedenem Namen, sonst wert oder kein Fehler.
func TestZeileG6TypnamenAehnlich(t *testing.T) {
	namen := []string{typProdukt, "Shop.ProductEx", "Shop.Produc", "shop.product", typProdukt + " ", "Other.Product", typGeschenk}
	n := 0
	for _, t1 := range namen {
		for _, t2 := range namen {
			for _, verschieden := range []bool{false, true} {
				n++
				e1, e2 := eP(pw("Name", model.ArtString, "Milch")), eP(pw("Name", model.ArtString, "Milch"))
				e1.Typ, e2.Typ = t1, t2
				if verschieden {
					e2 = eP(pw("Name", model.ArtString, "Brot"))
					e2.Typ = t2
				}
				_, err := zweiBeobachtungen(t, e1, e2)
				var k *model.KonfliktFehler
				switch {
				case t1 != t2:
					if !errors.As(err, &k) || k.Ursache != model.KonfliktTyp {
						t.Errorf("%q gegen %q (verschieden=%v): %v", t1, t2, verschieden, err)
					}
				case verschieden:
					if !errors.As(err, &k) || k.Ursache != model.KonfliktWert {
						t.Errorf("%q gegen %q (verschieden=%v): %v", t1, t2, verschieden, err)
					}
				case err != nil:
					t.Errorf("%q gegen %q (verschieden=%v): %v", t1, t2, verschieden, err)
				}
			}
		}
	}
	if n != 98 {
		t.Errorf("Menge %d, erwartet 98", n)
	}
}

func permutationen(n int) [][]int {
	if n == 0 {
		return [][]int{{}}
	}
	var out [][]int
	for _, p := range permutationen(n - 1) {
		for i := 0; i <= len(p); i++ {
			q := slices.Insert(slices.Clone(p), i, n-1)
			out = append(out, q)
		}
	}
	return out
}

func beobachtungenG7() []model.Entity {
	return []model.Entity{
		eP(pw("P1", model.ArtString, "a"), pw("P2", model.ArtString, "v2")),
		eP(pw("P2", model.ArtString, "v2"), pw("P3", model.ArtString, "c")),
		eP(pw("P3", model.ArtString, "c"), pw("P4", model.ArtString, "d")),
		eP(pw("P1", model.ArtString, "a"), pw("P4", model.ArtString, "d")),
		eP(pw("P2", model.ArtString, "x2")),
	}
}

// TestZeileG7Reihenfolge: das Ergebnis hängt von der Reihenfolge nicht ab; der gemeldete Konflikt
// folgt der sequence_no (24 und 120 Reihenfolgen, dazu drei Lieferreihenfolgen der Liste).
func TestZeileG7Reihenfolge(t *testing.T) {
	os := beobachtungenG7()
	kern := func(w *welt) string {
		e := entitaeten(manifest(t, w))[0]
		return fmt.Sprint(e["payload"], e["selected_properties"], e["completeness"])
	}
	var soll string
	for i, p := range permutationen(4) {
		w := neu(t)
		b := w.standard()
		for _, k := range p {
			b.beobachtung("Products", false, os[k])
		}
		mustImport(t, w, anfrage())
		if i == 0 {
			soll = kern(w)
		} else if got := kern(w); got != soll {
			t.Errorf("Reihenfolge %v: %s, erwartet %s", p, got, soll)
		}
	}
	if len(permutationen(4)) != 24 || len(permutationen(5)) != 120 {
		t.Fatal("Zahl der Reihenfolgen")
	}
	for _, p := range permutationen(5) {
		w := neu(t)
		b := w.standard()
		var ids []string
		for _, k := range p {
			ids = append(ids, b.beobachtung("Products", false, os[k]).ID)
		}
		_, err := w.importiere(anfrage())
		var k *model.KonfliktFehler
		if !errors.As(err, &k) {
			t.Errorf("Reihenfolge %v: kein Konflikt: %v", p, err)
			continue
		}
		// Orakel: in der Folge der sequence_no; Erste der erste Beitrag zu P2, Zweite der erste abweichende.
		erste, wert, zweite := "", "", ""
		for i, idx := range p {
			for _, pr := range os[idx].Properties {
				if pr.Name != "P2" {
					continue
				}
				switch {
				case erste == "":
					erste, wert = ids[i], pr.Wert.Text
				case pr.Wert.Text != wert && zweite == "":
					zweite = ids[i]
				}
			}
			if zweite != "" {
				break
			}
		}
		if k.Property != "P2" || k.Erste != erste || k.Zweite != zweite {
			t.Errorf("Reihenfolge %v: Konflikt %+v, erwartet %s und %s", p, k, erste, zweite)
		}
	}
	// Die Liste der Exchanges in drei anderen Reihenfolgen geliefert.
	manifestMit := func(umordnen func([]model.Exchange)) string {
		w := neu(t)
		b := w.standard()
		for _, k := range []int{0, 1, 2, 3} {
			b.beobachtung("Products", false, os[k])
		}
		umordnen(w.exchanges.m[sitzungID])
		mustImport(t, w, anfrage())
		return string(w.blobs.writers[0].bytes())
	}
	soll = manifestMit(func([]model.Exchange) {})
	for name, f := range map[string]func([]model.Exchange){
		"umgekehrt":  slices.Reverse[[]model.Exchange],
		"rotiert":    func(es []model.Exchange) { copy(es, append(slices.Clone(es[2:]), es[:2]...)) },
		"vertauscht": func(es []model.Exchange) { es[0], es[3], es[1], es[4] = es[3], es[0], es[4], es[1] },
	} {
		if got := manifestMit(f); got != soll {
			t.Errorf("%s: anderes Manifest", name)
		}
	}
}

// TestZeileG8Vollstaendigkeit: complete, wenn jede unterstützte Property bekannt ist; @odata.type
// nur bei abweichendem Typ (Menge 6).
func TestZeileG8Vollstaendigkeit(t *testing.T) {
	id, nm, pr := pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "Milch"), pw("Price", model.ArtDecimal, "1.5")
	faelle := []struct {
		name  string
		typ   string
		props []model.PropertyWert
		want  string
	}{
		{"Id", typProdukt, []model.PropertyWert{id}, "partial"},
		{"Id Name", typProdukt, []model.PropertyWert{id, nm}, "partial"},
		{"Id Price", typProdukt, []model.PropertyWert{id, pr}, "partial"},
		{"Id Name Price", typProdukt, []model.PropertyWert{id, nm, pr}, "complete"},
		{"Gift", typGeschenk, []model.PropertyWert{id, nm, pr}, "complete"},
	}
	for _, f := range faelle {
		w := neu(t)
		w.standard().beobachtung("Products", false, ent(f.typ, "1", f.props...))
		mustImport(t, w, anfrage())
		e := entitaeten(manifest(t, w))[0]
		if e["completeness"] != f.want {
			t.Errorf("%s: %v, erwartet %s", f.name, e["completeness"], f.want)
		}
		typ, hat := e["payload"].(map[string]any)["@odata.type"]
		if f.typ == typGeschenk && (!hat || typ != "#Shop.Gift") || f.typ == typProdukt && hat {
			t.Errorf("%s: @odata.type %v", f.name, typ)
		}
		if hat && slices.Contains(e["selected_properties"].([]any), any("@odata.type")) {
			t.Errorf("%s: @odata.type in selected_properties", f.name)
		}
	}
}

// TestZeileG9NichtUnterstuetzt: unsupported_features zählt die verschiedenen Entities je Mitglied
// und gibt die Funde des Metadatenlesers in dessen Reihenfolge aus (Menge 14).
func TestZeileG9NichtUnterstuetzt(t *testing.T) {
	n := 0
	mitGrund := func(key string, g model.UebersprungGrund) model.Entity {
		e := ent(typProdukt, key, pw("Id", model.ArtInt32, key))
		e.Uebersprungen = []model.Uebersprungen{{Name: "X", Grund: g}}
		return e
	}
	for _, g := range []model.UebersprungGrund{model.UebersprungTyp, model.UebersprungNavigation, model.UebersprungOffen} {
		for i, form := range []string{"eine Entity", "zwei Entities", "dieselbe Entity in zwei Exchanges"} {
			n++
			w := neu(t)
			b := w.standard()
			switch form {
			case "eine Entity":
				b.beobachtung("Products", true, mitGrund("1", g))
			case "zwei Entities":
				b.beobachtung("Products", true, mitGrund("1", g), mitGrund("2", g))
			default:
				b.beobachtung("Products", true, mitGrund("1", g))
				b.beobachtung("Products", false, mitGrund("1", g))
			}
			r := mustImport(t, w, anfrage())
			liste := bericht(t, r)["report"].(map[string]any)["unsupported_features"].(map[string]any)["entities"].([]any)
			want := []int{1, 2, 1}[i]
			if len(liste) != 1 {
				t.Errorf("%s %s: %v", g, form, liste)
				continue
			}
			z := liste[0].(map[string]any)
			if int(z["entities"].(float64)) != want || z["name"] != "X" || z["reason"] != string(g) || z["entity_set"] != "Products" {
				t.Errorf("%s %s: %v", g, form, z)
			}
		}
	}
	funde := []model.Fund{
		{Art: model.FundPropertyTyp, Ort: "Shop.Product/Rating", Name: "Edm.Double"}, {Art: model.FundAnnotation, Ort: "Shop", Name: "Core.Description"},
		{Art: model.FundKonstrukt, Ort: "Shop.Container", Name: "Singleton"}, {Art: model.FundErweiterung, Ort: "Shop", Name: "{urn:x}y"},
	}
	for i, reihenfolge := range [][]model.Fund{funde, {funde[3], funde[1], funde[0], funde[2]}} {
		n += []int{4, 1}[i] // vier Arten in der Reihenfolge des Lesers, dazu der Fall der vertauschten Reihenfolge
		w := neu(t)
		// Der Leser liefert seinen Bericht; die Erwartung ist eine Kopie, denn der Dienst darf ihn nicht ordnen.
		erwartet := slices.Clone(reihenfolge)
		w.meta.vorgabe.bericht = model.ValidierungsBericht{Funde: reihenfolge}
		w.sitzung().metadaten("<edmx 1/>")
		r := mustImport(t, w, anfrage())
		liste := bericht(t, r)["report"].(map[string]any)["unsupported_features"].(map[string]any)["metadata"].([]any)
		if len(liste) != 4 {
			t.Fatalf("Funde %v", liste)
		}
		for i, f := range erwartet {
			z := liste[i].(map[string]any)
			if z["kind"] != string(f.Art) || z["location"] != f.Ort || z["name"] != f.Name {
				t.Errorf("Fund %d: %v, erwartet %+v", i, z, f)
			}
		}
	}
	if n != 14 {
		t.Errorf("Menge %d", n)
	}
}

// TestZeileG9OrdnungDerZeilen: die Ordnung von unsupported_features.entities nach (entity_set, name, reason) in
// Bytes (SPEC-127 Nr. 14 b), an der Ordnungsfunktion: 12 Einträge {2 Sets mal 2 Namen mal 3 Gründe}, jedes der 66
// Paare in beiden Reihenfolgen (132 Aufrufe); Orakel die Ordnung der Zeichenketten (Set, dann Name, dann Grund).
func TestZeileG9OrdnungDerZeilen(t *testing.T) {
	var tupel [][3]string
	for _, set := range []string{"Products", "Orders"} {
		for _, name := range []string{"Beta", "Alpha"} {
			for _, grund := range []string{string(model.UebersprungTyp), string(model.UebersprungNavigation), string(model.UebersprungOffen)} {
				tupel = append(tupel, [3]string{set, name, grund})
			}
		}
	}
	if len(tupel) != 12 {
		t.Fatalf("Einträge %d", len(tupel))
	}
	vorzeichen := func(n int) int {
		switch {
		case n < 0:
			return -1
		case n > 0:
			return 1
		}
		return 0
	}
	soll := func(a, b [3]string) int {
		for i := range a {
			if c := strings.Compare(a[i], b[i]); c != 0 {
				return c
			}
		}
		return 0
	}
	aufrufe := 0
	for i := range tupel {
		for j := range tupel {
			if i == j {
				continue
			}
			aufrufe++
			if got, want := vorzeichen(baselineimport.VergleicheZeilen(tupel[i], tupel[j])), vorzeichen(soll(tupel[i], tupel[j])); got != want {
				t.Errorf("%v gegen %v: Vorzeichen %d, erwartet %d", tupel[i], tupel[j], got, want)
			}
		}
	}
	if aufrufe != 132 {
		t.Errorf("Aufrufe %d, erwartet 132", aufrufe)
	}
	if baselineimport.VergleicheZeilen(tupel[0], tupel[0]) != 0 {
		t.Error("derselbe Eintrag ist nicht gleich")
	}
	// Dieselben zwölf Einträge durch den Import: die Folge im Antwortobjekt ist die der Ordnung, in jedem von 20 Läufen.
	for lauf := 0; lauf < 20; lauf++ {
		w := neu(t)
		b := w.standard()
		for _, set := range []string{"Products", "Tags"} {
			e := ent(typProdukt, "1", pw("Id", model.ArtInt32, "1"))
			for _, name := range []string{"Beta", "Alpha"} {
				for _, g := range []model.UebersprungGrund{model.UebersprungTyp, model.UebersprungNavigation, model.UebersprungOffen} {
					e.Uebersprungen = append(e.Uebersprungen, model.Uebersprungen{Name: name, Grund: g})
				}
			}
			b.collection(set, "", seite(e))
		}
		r := mustImport(t, w, anfrage())
		var got []string
		for _, z := range bericht(t, r)["report"].(map[string]any)["unsupported_features"].(map[string]any)["entities"].([]any) {
			m := z.(map[string]any)
			got = append(got, fmt.Sprint(m["entity_set"], "|", m["name"], "|", m["reason"]))
		}
		if len(got) != 12 || !slices.IsSorted(got) {
			t.Fatalf("Lauf %d: %v", lauf, got)
		}
	}
}

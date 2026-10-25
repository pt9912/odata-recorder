package baselineimport_test

import (
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eintrag liefert Status und Evidenz des Sets aus dem Manifest.
func eintrag(t *testing.T, w *welt, set string) (string, map[string]any) {
	t.Helper()
	for _, c := range manifest(t, w)["coverage"].([]any) {
		m := c.(map[string]any)
		if m["entity_set"] == set {
			return m["status"].(string), m["evidence"].(map[string]any)
		}
	}
	t.Fatalf("kein Eintrag für %s", set)
	return "", nil
}

func grundVon(ev map[string]any) string { return ev["reason"].(string) }

// beginntKette sagt, ob die Evidenz eine ungefilterte Abfrage als Beginn einer Kette ausweist.
func beginntKette(ev map[string]any) bool {
	g := grundVon(ev)
	return g == "kette_nicht_belegt" || g == "kette_belegt" || g == "ketten_abweichend"
}

// TestZeileH2Positivliste: eine Query ist ungefiltert genau dann, wenn jede Option $format, $count,
// $select oder $orderby heißt (Menge 58: 8 positive und 50 negative).
func TestZeileH2Positivliste(t *testing.T) {
	alle := "$format=json&$count=true&$select=Id&$orderby=Id"
	positiv := []string{"", "$format=json", "$count=true", "$select=Id", "$orderby=Id", alle, "$orderby=Id&$select=Id&$count=true&$format=json", "%24count=true"}
	namen := []string{"$filter", "$top", "$skip", "$skiptoken", "$deltatoken", "$search", "$apply", "$expand", "$compute", "$levels", "$schemaversion", "$index",
		"$batch", "$entity", "$ref", "$value", "$id", "$Format", "$COUNT", "$Select", "$OrderBy", "sap-client", "@a", "x", ""}
	if len(positiv) != 8 || len(namen) != 25 {
		t.Fatalf("Mengen %d und %d", len(positiv), len(namen))
	}
	pruefe := func(q string, erwartet bool) {
		t.Helper()
		w := neu(t)
		w.standard().collection("Products", q, seite(produkt(1, "A", "1")))
		mustImport(t, w, anfrage())
		_, ev := eintrag(t, w, "Products")
		if beginntKette(ev) != erwartet {
			t.Errorf("Query %q: Grund %s, erwartet ungefiltert %v", q, grundVon(ev), erwartet)
		}
	}
	n := 0
	for _, q := range positiv {
		n++
		pruefe(q, true)
	}
	for _, name := range namen {
		opt := name + "=1"
		if name == "x" {
			opt = "x"
		}
		for _, q := range []string{opt, alle + "&" + opt} {
			n++
			pruefe(q, false)
		}
	}
	if n != 58 {
		t.Errorf("Menge %d, erwartet 58", n)
	}
}

// TestZeileH11PartialUnknown: unknown, partial und der Grund (Menge 11).
func TestZeileH11PartialUnknown(t *testing.T) {
	type fall struct {
		name   string
		bau    func(b *bauer)
		status string
		grund  string
	}
	coll := func(q string, s model.EntitySeite, as ...aenderung) func(b *bauer) {
		return func(b *bauer) { b.collection("Products", q, s, as...) }
	}
	e1 := produkt(1, "A", "1")
	faelle := []fall{
		{"keine Antwort", func(*bauer) {}, "unknown", "keine_beobachtung"},
		{"query übersprungen", coll("$apply=x", seite(e1)), "unknown", "keine_beobachtung"},
		{"status übersprungen", coll("", seite(e1), status(404)), "unknown", "keine_beobachtung"},
		{"Projektion ohne Schlüssel", coll("$select=Name", seite(e1)), "unknown", "keine_beobachtung"},
		{"Einzel-Entity", func(b *bauer) { b.entity("Products", "1", e1) }, "partial", "keine_ungefilterte_abfrage"},
		{"$top=10", coll("$top=10", seite(e1)), "partial", "keine_ungefilterte_abfrage"},
		{"$filter", coll("$filter=Id%20eq%201", seite(e1)), "partial", "keine_ungefilterte_abfrage"},
		{"$skip", coll("$skip=1", seite(e1)), "partial", "keine_ungefilterte_abfrage"},
		{"$top=0 mit leerem value", coll("$top=0", seite()), "partial", "keine_ungefilterte_abfrage"},
		{"nur die Folgeseite", coll("$skiptoken=1", seite(e1)), "partial", "keine_ungefilterte_abfrage"},
		{"erste Seite mit Link ohne Folgeseite", coll("", mitLink(seite(e1), "Products?$skiptoken=1")), "partial", "kette_nicht_belegt"},
	}
	if len(faelle) != 11 {
		t.Fatalf("Menge %d", len(faelle))
	}
	for _, f := range faelle {
		w := neu(t)
		f.bau(w.standard())
		mustImport(t, w, anfrage())
		status, ev := eintrag(t, w, "Products")
		if status != f.status || grundVon(ev) != f.grund {
			t.Errorf("%s: %s/%s, erwartet %s/%s", f.name, status, grundVon(ev), f.status, f.grund)
		}
	}
}

// TestZeileH12EintragJeSet: ein Eintrag je Set des Modells, aufsteigend nach entity_set (Menge 4).
func TestZeileH12EintragJeSet(t *testing.T) {
	for _, sets := range [][]string{nil, {"Only"}, {"Zebra", "Apple", "Mango"}} {
		w := neu(t)
		var es []model.EntitySet
		for _, s := range sets {
			es = append(es, model.EntitySet{Name: s, EntityType: typProdukt})
		}
		w.meta.vorgabe.modell = model.NeuesServiceModel("4.0", []string{"Shop"}, modellR().EntityTypes(), es, nil)
		w.sitzung().metadaten("<edmx/>")
		mustImport(t, w, anfrage())
		var namen []string
		for _, c := range manifest(t, w)["coverage"].([]any) {
			namen = append(namen, c.(map[string]any)["entity_set"].(string))
		}
		want := slices.Sorted(slices.Values(sets))
		if len(sets) == 0 {
			want = nil
		}
		if !slices.Equal(namen, want) {
			t.Errorf("Modell %v: Einträge %v, erwartet %v", sets, namen, want)
		}
	}
	// Eine Antwort für ein Set, das das Modell nicht kennt, ergibt keinen Eintrag und den Grund pfad.
	w := neu(t)
	b := w.standard()
	b.collection("Unknown", "", seite(produkt(1, "A", "1")))
	r := mustImport(t, w, anfrage())
	var namen []string
	for _, c := range manifest(t, w)["coverage"].([]any) {
		namen = append(namen, c.(map[string]any)["entity_set"].(string))
	}
	if !slices.Equal(namen, []string{"Orders", "Products", "Tags"}) || uebersprungen(t, r)["pfad"] != 1 {
		t.Errorf("Einträge %v, übersprungen %v", namen, uebersprungen(t, r))
	}
}

// TestZeileH13FormDerEvidenz: die Evidenz in der Kanonform für die drei Sets des Referenzfalls: kein Beleg,
// keine ungefilterte Abfrage und eine belegte Kette.
func TestZeileH13FormDerEvidenz(t *testing.T) {
	w := weltI1(t)
	mustImport(t, w, anfrage())
	soll := map[string]string{
		"Tags":     `{"chains":[],"entities":0,"reason":"keine_beobachtung","responses":{"collection":0,"entity":0},"version":1}`,
		"Orders":   `{"chains":[],"entities":1,"reason":"keine_ungefilterte_abfrage","responses":{"collection":0,"entity":1},"version":1}`,
		"Products": `{"chains":[{"end":"proven","pages":[{"elements":2,"exchange_id":"exc-002","sequence_no":2}],"start":{"exchange_id":"exc-002","sequence_no":2}}],"entities":2,"reason":"kette_belegt","responses":{"collection":1,"entity":0},"version":1}`,
	}
	for _, c := range w.imports.pubs[0].Coverage {
		if c.EvidenceJSON != soll[c.EntitySet] {
			t.Errorf("%s: %s", c.EntitySet, c.EvidenceJSON)
		}
	}
}

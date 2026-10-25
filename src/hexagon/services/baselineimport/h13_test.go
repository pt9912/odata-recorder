package baselineimport_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestZeileH13EvidenzMitKette: die Evidenz in der Kanonform für die zwei Gründe mit Kette und die
// acht Enden einer Kette, als Goldens von Hand (die drei Gründe ohne Kette stehen in H13FormDerEvidenz).
func TestZeileH13EvidenzMitKette(t *testing.T) {
	seite1 := `{"elements":1,"exchange_id":"exc-002","sequence_no":2}`
	seite2 := `{"elements":1,"exchange_id":"exc-003","sequence_no":3}`
	start := `"start":{"exchange_id":"exc-002","sequence_no":2}}`
	ev := func(ketten, entities, reason, collection string) string {
		return `{"chains":[` + ketten + `],"entities":` + entities + `,"reason":"` + reason + `","responses":{"collection":` + collection + `,"entity":0},"version":1}`
	}
	kette := func(a ...seitenSpec) func(*bauer) { return func(b *bauer) { b.kette(a...) } }
	faelle := []struct {
		name   string
		bau    func(b *bauer)
		soll   string
		status string
	}{
		{"proven", kette(seitenSpec{ids: []int{1}, link: linkZu(1)}, seitenSpec{ids: []int{2}}),
			ev(`{"end":"proven","pages":[`+seite1+`,`+seite2+`],`+start, "2", "kette_belegt", "2"), "complete"},
		{"proven mit Anzahl", kette(seitenSpec{ids: []int{1}, anzahl: int64p(1)}),
			ev(`{"count":{"declared":1,"observed":1},"end":"proven","pages":[`+seite1+`],`+start, "1", "kette_belegt", "1"), "complete"},
		{"seite_nicht_lesbar", kette(seitenSpec{ids: []int{1}, link: linkZu(1)}, seitenSpec{ids: []int{2}, aend: status(404)}),
			ev(`{"end":"seite_nicht_lesbar","pages":[`+seite1+`],`+start, "1", "kette_nicht_belegt", "1"), "partial"},
		{"seite_nicht_lesbar (anderes Set)", func(b *bauer) {
			schluessel := b.w.matchKey("/Products", queryVon(1))
			b.kette(seitenSpec{ids: []int{1}, link: linkZu(1)}, seitenSpec{ids: []int{2}, aend: func(e *model.Exchange) { e.Path, e.MatchKey = "/Orders", schluessel }})
		}, ev(`{"end":"seite_nicht_lesbar","pages":[`+seite1+`],`+start, "1", "kette_nicht_belegt", "1"), "partial"},
		{"key_redacted", kette(seitenSpec{ids: []int{1}, link: linkZu(1)}, seitenSpec{ids: []int{2}, aend: func(e *model.Exchange) { e.KeyRedacted = true }}),
			ev(`{"end":"key_redacted","pages":[`+seite1+`],`+start, "2", "kette_nicht_belegt", "2"), "partial"},
		{"link_redigiert", kette(seitenSpec{ids: []int{1}, redigiert: true}),
			ev(`{"end":"link_redigiert","pages":[`+seite1+`],`+start, "1", "kette_nicht_belegt", "1"), "partial"},
		{"link_nicht_aufloesbar", kette(seitenSpec{ids: []int{1}, link: "mailto:x@y"}),
			ev(`{"end":"link_nicht_aufloesbar","pages":[`+seite1+`],`+start, "1", "kette_nicht_belegt", "1"), "partial"},
		{"schleife", kette(seitenSpec{ids: []int{1}, link: linkZu(0)}),
			ev(`{"end":"schleife","pages":[`+seite1+`],`+start, "1", "kette_nicht_belegt", "1"), "partial"},
		{"seite_fehlt", kette(seitenSpec{ids: []int{1}, link: "Products?$skiptoken=77"}),
			ev(`{"end":"seite_fehlt","pages":[`+seite1+`],`+start, "1", "kette_nicht_belegt", "1"), "partial"},
		{"anzahl_abweichend", kette(seitenSpec{ids: []int{1}, anzahl: int64p(5)}),
			ev(`{"count":{"declared":5,"observed":1},"end":"anzahl_abweichend","pages":[`+seite1+`],`+start, "1", "kette_nicht_belegt", "1"), "partial"},
		{"bestand_ausserhalb_kette", func(b *bauer) {
			b.kette(seitenSpec{ids: []int{1}})
			b.entity("Products", "2", produkteVon([]int{2})[0])
		}, `{"chains":[{"end":"proven","pages":[` + seite1 + `],` + start + `],"entities":2,"reason":"bestand_ausserhalb_kette","responses":{"collection":1,"entity":1},"version":1}`, "partial"},
		{"ketten_abweichend", func(b *bauer) {
			b.collection("Products", "", seite(produkteVon([]int{1, 2})...))
			b.collection("Products", "$format=json", seite(produkteVon([]int{1, 3})...))
		}, ev(`{"end":"proven","pages":[{"elements":2,"exchange_id":"exc-002","sequence_no":2}],`+start+
			`,{"end":"proven","pages":[{"elements":2,"exchange_id":"exc-003","sequence_no":3}],"start":{"exchange_id":"exc-003","sequence_no":3}}`,
			"3", "ketten_abweichend", "2"), "partial"},
	}
	n := 0
	for _, f := range faelle {
		n++
		w := neu(t)
		f.bau(w.standard())
		mustImport(t, w, anfrage())
		var got string
		var status model.CoverageStatus
		for _, c := range w.imports.pubs[0].Coverage {
			if c.EntitySet == "Products" {
				got, status = c.EvidenceJSON, c.Status
			}
		}
		if got != f.soll || string(status) != f.status {
			t.Errorf("%s: %s\nerwartet: %s (%s)", f.name, got, f.soll, f.status)
		}
	}
	if n != 12 {
		t.Errorf("Menge %d, erwartet 12 (mit den drei Goldens von H13FormDerEvidenz 15)", n)
	}
}

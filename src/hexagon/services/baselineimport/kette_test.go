package baselineimport_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func mitAnzahl(s model.EntitySeite, n int64) model.EntitySeite {
	s.Anzahl = model.Anzahl{Vorhanden: true, Wert: n}
	return s
}

func redigierterLink(s model.EntitySeite) model.EntitySeite {
	s.FolgeLink = model.FolgeLink{Vorhanden: true, Redigiert: true}
	return s
}

// seitenSpec beschreibt eine Seite einer Kette: die Schlüssel der Elemente, den Link, die Anzahl und
// eine Änderung am Exchange.
type seitenSpec struct {
	ids       []int
	link      string
	redigiert bool
	anzahl    *int64
	aend      aenderung
}

func produkteVon(ids []int) []model.Entity {
	var es []model.Entity
	for _, id := range ids {
		es = append(es, produkt(id, fmt.Sprint("P", id), "1"))
	}
	return es
}

// linkZu ist der relative Link auf die Seite k einer Kette (k gleich 0: die erste Seite).
func linkZu(k int) string {
	if k == 0 {
		return "Products"
	}
	return fmt.Sprintf("Products?$skiptoken=%d", k)
}

func queryVon(k int) string {
	if k == 0 {
		return ""
	}
	return fmt.Sprintf("$skiptoken=%d", k)
}

// kette legt die Seiten einer Kette für Products an: die erste ohne Query, die k-te mit $skiptoken=k.
func (b *bauer) kette(seiten ...seitenSpec) []model.Exchange {
	var out []model.Exchange
	for k, s := range seiten {
		sp := seite(produkteVon(s.ids)...)
		if s.link != "" {
			sp = mitLink(sp, s.link)
		}
		if s.redigiert {
			sp = redigierterLink(sp)
		}
		if s.anzahl != nil {
			sp = mitAnzahl(sp, *s.anzahl)
		}
		out = append(out, b.collection("Products", queryVon(k), sp, s.aend))
	}
	return out
}

// standardSeiten liefert n Seiten mit je einem Element; die letzte hat keinen Link.
func standardSeiten(n int) []seitenSpec {
	var out []seitenSpec
	for k := 0; k < n; k++ {
		s := seitenSpec{ids: []int{k + 1}}
		if k < n-1 {
			s.link = linkZu(k + 1)
		}
		out = append(out, s)
	}
	return out
}

func int64p(n int64) *int64 { return &n }

// ketteVon liest Status, Grund und Ketten des Sets Products aus dem Manifest.
func ketteVon(t *testing.T, w *welt) (string, string, []map[string]any) {
	t.Helper()
	status, ev := eintrag(t, w, "Products")
	var ketten []map[string]any
	for _, k := range ev["chains"].([]any) {
		ketten = append(ketten, k.(map[string]any))
	}
	return status, grundVon(ev), ketten
}

func endeVon(k map[string]any) string { return k["end"].(string) }

func seitenNummern(k map[string]any) []int {
	var out []int
	for _, p := range k["pages"].([]any) {
		out = append(out, int(p.(map[string]any)["sequence_no"].(float64)))
	}
	return out
}

// matchKey bildet den Match-Key von GET pfad?query mit den Regeln der Welt.
func (w *welt) matchKey(pfad, query string) string {
	k, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: dienstID, Path: pfad, RawQuery: query, RequestRedaction: model.RedactionNone}, w.konfig.Regeln)
	if err != nil {
		w.t.Fatal(err)
	}
	return k.MatchKey
}

// TestZeileH1BelegteKette: eine belegte Kette ist complete (Menge 7).
func TestZeileH1BelegteKette(t *testing.T) {
	for n := 1; n <= 4; n++ {
		w := neu(t)
		w.standard().kette(standardSeiten(n)...)
		mustImport(t, w, anfrage())
		status, grund, ketten := ketteVon(t, w)
		if status != "complete" || grund != "kette_belegt" || len(ketten) != 1 || endeVon(ketten[0]) != "proven" || len(ketten[0]["pages"].([]any)) != n {
			t.Errorf("Länge %d: %s/%s %v", n, status, grund, ketten)
		}
	}
	w := neu(t)
	w.standard().kette(seitenSpec{})
	mustImport(t, w, anfrage())
	status, grund, ketten := ketteVon(t, w)
	_, ev := eintrag(t, w, "Products")
	if status != "complete" || grund != "kette_belegt" || len(ketten) != 1 || ev["entities"].(float64) != 0 {
		t.Errorf("leere Collection: %s/%s %v", status, grund, ketten)
	}
	w = neu(t)
	w.standard().collection("Products", "$skiptoken=1", seite(produkt(1, "A", "1")))
	mustImport(t, w, anfrage())
	if status, grund, ketten := ketteVon(t, w); status != "partial" || grund != "keine_ungefilterte_abfrage" || len(ketten) != 0 {
		t.Errorf("nur die Folgeseite: %s/%s %v", status, grund, ketten)
	}
	w = neu(t)
	w.standard().kette(seitenSpec{ids: []int{1}, link: linkZu(1)})
	mustImport(t, w, anfrage())
	status, grund, ketten = ketteVon(t, w)
	if status != "partial" || grund != "kette_nicht_belegt" || len(ketten) != 1 || endeVon(ketten[0]) != "seite_fehlt" {
		t.Errorf("Link ohne Folgeseite: %s/%s %v", status, grund, ketten)
	}
}

// TestZeileH3SeiteLesbar: jede Seite der Kette ist eine gelesene Antwort desselben Sets (Menge 30).
func TestZeileH3SeiteLesbar(t *testing.T) {
	zustaende := []struct {
		name string
		a    aenderung
	}{
		{"capture incomplete", func(e *model.Exchange) { e.Capture = model.CaptureIncomplete }},
		{"capture failed", func(e *model.Exchange) { e.Capture = model.CaptureFailed }},
		{"Status 404", status(404)},
		{"Status 500", status(500)},
		{"Body absent", func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = model.BodyAbsent, "" }},
		{"kein_json", ct("text/plain")},
		{"ohne_kontext", ct("application/json;odata.metadata=none")},
		{"Option $apply", func(e *model.Exchange) { e.RawQuery = "$apply=x&" + e.RawQuery }},
		{"$select=Name", func(e *model.Exchange) { e.RawQuery = "$select=Name&" + e.RawQuery }},
		{"anderes Set", func(e *model.Exchange) { e.Path = "/Orders" }},
	}
	n := 0
	for seiteNr := 2; seiteNr <= 4; seiteNr++ {
		for _, z := range zustaende {
			n++
			w := neu(t)
			spec := standardSeiten(4)
			schluessel := w.matchKey("/Products", queryVon(seiteNr-1))
			spec[seiteNr-1].aend = func(e *model.Exchange) {
				z.a(e)
				e.MatchKey = schluessel
			}
			w.standard().kette(spec...)
			mustImport(t, w, anfrage())
			status, grund, ketten := ketteVon(t, w)
			if status != "partial" || grund != "kette_nicht_belegt" || len(ketten) != 1 || endeVon(ketten[0]) != "seite_nicht_lesbar" ||
				len(ketten[0]["pages"].([]any)) != seiteNr-1 {
				t.Errorf("Seite %d, %s: %s/%s %v", seiteNr, z.name, status, grund, ketten)
			}
		}
	}
	if n != 30 {
		t.Errorf("Menge %d, erwartet 30", n)
	}
	// Ergänzung: eine gelesene Antwort der Spur Entity mit dem Key der erwarteten Seite ist keine Seite der Collection.
	w := neu(t)
	spec := standardSeiten(4)
	schluessel := w.matchKey("/Products", queryVon(1))
	spec[1].aend = func(e *model.Exchange) { e.Path, e.RawQuery, e.MatchKey = "/Products(2)", "", schluessel }
	exs := w.standard().kette(spec...)
	w.daten.nachRef[exs[1].ResponseBlob] = datenVorgabe{entity: produkt(2, "A", "1")}
	mustImport(t, w, anfrage())
	if _, _, ketten := ketteVon(t, w); len(ketten) != 1 || endeVon(ketten[0]) != "seite_nicht_lesbar" || len(ketten[0]["pages"].([]any)) != 1 {
		t.Errorf("Antwort der Spur Entity: %v", ketten)
	}
}

// TestZeileH4NaechsteSeite: die nächste Seite ist die mit der kleinsten größeren sequence_no (Menge 5).
func TestZeileH4NaechsteSeite(t *testing.T) {
	// Link auf einen Key, den die Session nicht hat.
	w := neu(t)
	w.standard().kette(seitenSpec{ids: []int{1}, link: "Products?$skiptoken=77"})
	mustImport(t, w, anfrage())
	if _, _, k := ketteVon(t, w); endeVon(k[0]) != "seite_fehlt" {
		t.Errorf("fremder Key: %v", k)
	}
	// Der einzige Treffer liegt vor der Seite.
	w = neu(t)
	b := w.standard()
	b.collection("Products", "$skiptoken=1", seite(produkt(9, "A", "1")))
	b.kette(seitenSpec{ids: []int{1}, link: linkZu(1)})
	mustImport(t, w, anfrage())
	if _, _, k := ketteVon(t, w); endeVon(k[0]) != "seite_fehlt" {
		t.Errorf("Treffer davor: %v", k)
	}
	// Zwei Treffer vor und zwei nach der Seite: der erste danach.
	w = neu(t)
	b = w.standard()
	b.collection("Products", "$skiptoken=1", seite(produkt(8, "A", "1")))
	b.collection("Products", "$skiptoken=1", seite(produkt(9, "A", "1")))
	start := b.kette(seitenSpec{ids: []int{1}, link: linkZu(1)})[0]
	erste := b.collection("Products", "$skiptoken=1", seite(produkt(2, "A", "1")))
	b.collection("Products", "$skiptoken=1", seite(produkt(3, "A", "1")))
	mustImport(t, w, anfrage())
	_, _, k := ketteVon(t, w)
	if endeVon(k[0]) != "proven" || !slices.Equal(seitenNummern(k[0]), []int{start.SequenceNo, erste.SequenceNo}) {
		t.Errorf("zwei Treffer vor und nach: %v", k)
	}
	// Link ohne Folgeseite.
	w = neu(t)
	w.standard().kette(seitenSpec{ids: []int{1}, link: linkZu(1)})
	mustImport(t, w, anfrage())
	if _, _, k := ketteVon(t, w); endeVon(k[0]) != "seite_fehlt" {
		t.Errorf("ohne Folgeseite: %v", k)
	}
	// Gleiche Query mit anderer Methode: anderer Key.
	w = neu(t)
	b = w.standard()
	b.kette(seitenSpec{ids: []int{1}, link: linkZu(1)})
	b.collection("Products", "$skiptoken=1", seite(produkt(2, "A", "1")), func(e *model.Exchange) { e.Method = "POST" })
	mustImport(t, w, anfrage())
	if _, _, k := ketteVon(t, w); endeVon(k[0]) != "seite_fehlt" {
		t.Errorf("andere Methode: %v", k)
	}
}

// TestZeileH5KeyUndLinkRedigiert: key_redacted und ein redigierter Link verhindern complete auf jeder
// Seite (Menge 7).
func TestZeileH5KeyUndLinkRedigiert(t *testing.T) {
	for seiteNr := 1; seiteNr <= 4; seiteNr++ {
		w := neu(t)
		spec := standardSeiten(4)
		spec[seiteNr-1].aend = func(e *model.Exchange) { e.KeyRedacted = true }
		w.standard().kette(spec...)
		mustImport(t, w, anfrage())
		status, grund, ketten := ketteVon(t, w)
		if status != "partial" || grund != "kette_nicht_belegt" || endeVon(ketten[0]) != "key_redacted" || len(ketten[0]["pages"].([]any)) != seiteNr-1 {
			t.Errorf("key_redacted auf Seite %d: %s/%s %v", seiteNr, status, grund, ketten)
		}
	}
	for seiteNr := 1; seiteNr <= 3; seiteNr++ {
		w := neu(t)
		spec := standardSeiten(4)
		spec[seiteNr-1].link, spec[seiteNr-1].redigiert = "", true
		w.standard().kette(spec...)
		mustImport(t, w, anfrage())
		status, grund, ketten := ketteVon(t, w)
		if status != "partial" || grund != "kette_nicht_belegt" || endeVon(ketten[0]) != "link_redigiert" || len(ketten[0]["pages"].([]any)) != seiteNr {
			t.Errorf("redigierter Link auf Seite %d: %s/%s %v", seiteNr, status, grund, ketten)
		}
	}
}

// TestZeileH7VaryHeader: die Vary-Header der Vorgängerseite bilden den Key (Menge 3).
func TestZeileH7VaryHeader(t *testing.T) {
	faelle := []struct {
		name string
		vary []string
		kopf string
	}{
		{"Accept", []string{"Accept"}, "Accept"}, {"Vary leer", nil, "Accept"}, {"andere Schreibweise", []string{"Accept"}, "accept"},
	}
	for _, f := range faelle {
		w := neu(t)
		w.konfig.Regeln.VaryHeaders = f.vary
		b := w.standard()
		// Seite k trägt den Header mit dem Wert "p<k>"; der Key der Seite k+1 ist mit den Headern der Seite k gebildet.
		var vorher map[string][]string
		for k := 0; k < 3; k++ {
			kopf := map[string][]string{f.kopf: {fmt.Sprint("p", k)}}
			vorherKopf := vorher
			sp := seite(produkt(k+1, "A", "1"))
			if k < 2 {
				sp = mitLink(sp, linkZu(k+1))
			}
			b.collection("Products", queryVon(k), sp, func(e *model.Exchange) {
				e.RequestHeaders = kopf
				if vorherKopf != nil {
					key, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: dienstID, Path: "/Products", RawQuery: queryVon(k), Header: vorherKopf, RequestRedaction: model.RedactionNone}, w.konfig.Regeln)
					if err != nil {
						t.Fatal(err)
					}
					e.MatchKey = key.MatchKey
				}
			})
			vorher = kopf
		}
		mustImport(t, w, anfrage())
		status, grund, ketten := ketteVon(t, w)
		if status != "complete" || grund != "kette_belegt" || len(ketten[0]["pages"].([]any)) != 3 {
			t.Errorf("%s: %s/%s %v", f.name, status, grund, ketten)
		}
	}
}

// TestZeileH8Schleife: ein Link auf den Key einer Seite der Kette ist eine Schleife (Menge 6).
func TestZeileH8Schleife(t *testing.T) {
	n := 0
	for p := 1; p <= 3; p++ {
		for q := 1; q <= p; q++ {
			n++
			w := neu(t)
			b := w.standard()
			spec := standardSeiten(4)
			spec[p-1].link = linkZu(q - 1)
			exs := b.kette(spec...)
			// Ein späterer Exchange mit dem Key der Seite q liegt vor.
			b.collection("Products", queryVon(q-1)+"&$format=json", seite(produkt(9, "A", "1")), func(e *model.Exchange) { e.MatchKey = exs[q-1].MatchKey })
			mustImport(t, w, anfrage())
			_, _, ketten := ketteVon(t, w)
			if endeVon(ketten[0]) != "schleife" || len(ketten[0]["pages"].([]any)) != p {
				t.Errorf("Seite %d auf Seite %d: %v", p, q, ketten[0])
			}
		}
	}
	if n != 6 {
		t.Errorf("Menge %d", n)
	}
}

// TestZeileH9Anzahl: @odata.count der ersten Seite gegen die verschiedenen Schlüssel (Menge 13).
func TestZeileH9Anzahl(t *testing.T) {
	n := 0
	for _, ids := range [][]int{{1, 2, 3}, {1, 2, 2}} {
		verschieden := map[int]bool{}
		for _, id := range ids {
			verschieden[id] = true
		}
		for _, declared := range []int64{1, 2, 3, 4, -1} {
			n++
			var a *int64
			if declared >= 0 {
				a = int64p(declared)
			}
			w := neu(t)
			w.standard().kette(seitenSpec{ids: ids, anzahl: a})
			mustImport(t, w, anfrage())
			status, grund, ketten := ketteVon(t, w)
			bewiesen := declared < 0 || declared == int64(len(verschieden))
			if bewiesen != (endeVon(ketten[0]) == "proven") || bewiesen != (status == "complete") {
				t.Errorf("Elemente %v, declared %d: %s/%s %v", ids, declared, status, grund, ketten[0])
			}
			if _, hat := ketten[0]["count"]; hat != (declared >= 0) {
				t.Errorf("Elemente %v, declared %d: count %v", ids, declared, ketten[0]["count"])
			}
			if declared >= 0 {
				c := ketten[0]["count"].(map[string]any)
				if int64(c["declared"].(float64)) != declared || int(c["observed"].(float64)) != len(verschieden) {
					t.Errorf("count %v", c)
				}
			}
		}
	}
	// Die Anzahl der zweiten Seite wird nicht gelesen.
	n++
	w := neu(t)
	spec := standardSeiten(2)
	spec[1].anzahl = int64p(99)
	w.standard().kette(spec...)
	mustImport(t, w, anfrage())
	if _, _, k := ketteVon(t, w); endeVon(k[0]) != "proven" {
		t.Errorf("Anzahl der zweiten Seite: %v", k[0])
	}
	// Leere Kette mit Anzahl 0 und 1.
	for _, declared := range []int64{0, 1} {
		n++
		w := neu(t)
		w.standard().kette(seitenSpec{anzahl: int64p(declared)})
		mustImport(t, w, anfrage())
		_, _, k := ketteVon(t, w)
		want := "proven"
		if declared == 1 {
			want = "anzahl_abweichend"
		}
		if endeVon(k[0]) != want {
			t.Errorf("leere Kette mit Anzahl %d: %v", declared, k[0])
		}
	}
	if n != 13 {
		t.Errorf("Menge %d, erwartet 13", n)
	}
}

// TestZeileH10MehrereKetten: alle belegt und gleiche Schlüsselmenge (Menge 10).
func TestZeileH10MehrereKetten(t *testing.T) {
	gleich := func(b *bauer, q string) { b.collection("Products", q, seite(produkteVon([]int{1, 2})...)) }
	anders := func(b *bauer, q string) { b.collection("Products", q, seite(produkteVon([]int{1, 3})...)) }
	gebrochen := func(b *bauer, q string) {
		b.collection("Products", q, mitLink(seite(produkteVon([]int{1, 2})...), "Products?$skiptoken=99"))
	}
	typen := map[string]func(*bauer, string){"gleich": gleich, "anders": anders, "gebrochen": gebrochen}
	faelle := []struct {
		starts []string
		grund  string
		status string
	}{
		{[]string{"gleich", "gleich"}, "kette_belegt", "complete"}, {[]string{"gleich", "gleich"}, "kette_belegt", "complete"},
		{[]string{"gleich", "anders"}, "ketten_abweichend", "partial"}, {[]string{"anders", "gleich"}, "ketten_abweichend", "partial"},
		{[]string{"gleich", "gebrochen"}, "kette_nicht_belegt", "partial"}, {[]string{"gebrochen", "gleich"}, "kette_nicht_belegt", "partial"},
		{[]string{"gebrochen", "gebrochen"}, "kette_nicht_belegt", "partial"},
		{[]string{"gleich", "gleich", "anders"}, "ketten_abweichend", "partial"}, {[]string{"gleich", "anders", "gleich"}, "ketten_abweichend", "partial"},
		{[]string{"anders", "gleich", "gleich"}, "ketten_abweichend", "partial"},
	}
	queries := []string{"", "$format=json", "$count=true"}
	n := 0
	for _, f := range faelle {
		n++
		w := neu(t)
		b := w.standard()
		for i, s := range f.starts {
			typen[s](b, queries[i])
		}
		mustImport(t, w, anfrage())
		status, grund, ketten := ketteVon(t, w)
		if status != f.status || grund != f.grund || len(ketten) != len(f.starts) {
			t.Errorf("%v: %s/%s mit %d Ketten", f.starts, status, grund, len(ketten))
		}
	}
	if n != 10 {
		t.Errorf("Menge %d", n)
	}
}

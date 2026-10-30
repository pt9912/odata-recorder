package urlparser_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// zuteilungDicht ist der Teil von TestWaechterZuteilung (Zeile 16, SPEC-160 Nr. 6 c, SPEC-162 Nr. 1) für Eingaben mit dichten Segmenten:
// jede Stelle, die den Text einmal durchläuft, bleibt unter 64 · n + 65536 Bytes. Menge 8, jedes Element endet mit `zu_viele`, weil
// die Zerlegung vor der Prüfung der Zahl läuft: (B) fünf Formen mit Segmenten in `rawQuery` bis 8192 Bytes (`a&`, `a=&`, `=a&`,
// `a=b&`, `&&a`, jeweils so oft wiederholt, wie sie in 8192 Bytes passen), (C) 1024 nichtleere Segmente im Pfad (2048 Bytes),
// (D) `$select` und `$orderby` mit 4088 Elementen (je 8184 Bytes). Die Zerlegung der Query allein, ohne den Parser, hält
// TestQuerySegmenteLinear im Paket model.
func zuteilungDicht(t *testing.T) {
	t.Helper()
	type element struct {
		name        string
		pfad, query string
		ursache     model.URLUrsache
		ort         string
	}
	var els []element
	for _, einheit := range []string{"a&", "a=&", "=a&", "a=b&", "&&a"} {
		els = append(els, element{"rawQuery " + einheit, "/S", wiederhole(einheit, 8192/len(einheit)), model.URLZuViele, "query"})
	}
	els = append(els, element{"Pfad /a", wiederhole("/a", 1024), "", model.URLZuViele, "pfad"})
	els = append(els, element{"$select", "/S", "$select=" + wiederhole("a,", 4088), model.URLZuViele, "query/1"})
	els = append(els, element{"$orderby", "/S", "$orderby=" + wiederhole("a,", 4088), model.URLZuViele, "query/1"})
	if len(els) != 8 {
		t.Fatalf("%d dichte Formen, erwartet 8", len(els))
	}
	for _, e := range els {
		z := gesamtZuteilung(func() { _, _ = parse(e.pfad, e.query) })
		if grenze := uint64(64*(len(e.pfad)+len(e.query)) + 65536); z > grenze {
			t.Errorf("%s: Zuteilung %d über %d", e.name, z, grenze)
		}
		pruefeFall(t, parserNeu(), fall{pfad: e.pfad, query: e.query, u: e.ursache, ort: e.ort})
	}
}

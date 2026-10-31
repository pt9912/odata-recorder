package urlparser_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// verschiedeneNamen bildet n verschiedene gültige Namen (ASCII): `n`, drei Ziffern der Nummer und `x` bis zur Länge laenge; der
// letzte hat die Länge letzte.
func verschiedeneNamen(n, laenge, letzte int) []string {
	namen := make([]string, n)
	for i := range namen {
		l := laenge
		if i == n-1 {
			l = letzte
		}
		namen[i] = fmt.Sprintf("n%03d", i) + wiederhole("x", l-4)
	}
	return namen
}

// zuteilungDichtSelect ist der Teil von TestWaechterZuteilung (Zeile 16, SPEC-160 Nr. 6 c, SPEC-162 Nr. 6) für ein gültiges, dichtes
// `$select`: die Zahl der Elemente und die Länge der Query liegen auf der Grenze, der Aufruf wird gelesen, und die Zuteilung bleibt
// unter 64 · n + 65536 Bytes, auch wenn ein Element den ganzen Wert kopierte. Menge 2: (Q5a) 256 Namen zu 30 Zeichen, `rawQuery` 7943
// Bytes; (Q5b) 255 Namen zu 31 und ein Name zu 24 Zeichen, `rawQuery` genau 8192 Bytes.
func zuteilungDichtSelect(t *testing.T) {
	t.Helper()
	// je Fall: Länge der Elemente, Länge des letzten Elements, Länge der Query von Hand
	for i, name := range []string{"Q5a", "Q5b"} {
		laenge, letzte, soll := [2]int{30, 31}[i], [2]int{30, 24}[i], [2]int{7943, 8192}[i]
		namen := verschiedeneNamen(256, laenge, letzte)
		query := "$select=" + strings.Join(namen, ",")
		if len(namen) != 256 || len(query) != soll {
			t.Fatalf("%s: %d Namen, rawQuery %d Bytes, erwartet 256 und %d", name, len(namen), len(query), soll)
		}
		z := gesamtZuteilung(func() { _, _ = parse("/S", query) })
		if grenze := uint64(64*(len("/S")+len(query)) + 65536); z > grenze {
			t.Errorf("%s: Zuteilung %d über %d", name, z, grenze)
		}
		want := &model.ODataURL{
			Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: "S"},
			Optionen:  model.Optionen{Select: &model.SelectOption{Namen: namen}},
		}
		pruefeFall(t, parserNeu(), fall{pfad: "/S", query: query, want: want})
	}
}

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

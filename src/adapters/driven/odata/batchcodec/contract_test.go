package batchcodec_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ersetze ersetzt das k-te Vorkommen (ab 0) von alt im Beispiel durch neu.
func ersetze(t *testing.T, alt, neu string, k int) string {
	t.Helper()
	s := beispiel()
	pos := 0
	for range k + 1 {
		i := strings.Index(s[pos:], alt)
		if i < 0 {
			t.Fatalf("%q steht nicht %d-mal im Beispiel", alt, k+1)
		}
		pos += i + 1
	}
	return s[:pos-1] + neu + s[pos-1+len(alt):]
}

// TestBatchContract ist der Parser-Contract-Test von LH-FA-16.a (Zeile 35): das Beispiel aus SPEC-183 Nr. 8 und jede Zeile seiner
// Tabelle; die Zeilen der Prüfung liest der Leser mit den geänderten Werten.
func TestBatchContract(t *testing.T) {
	got, err := leseMit(strings.NewReader(beispiel()), ctBeispiel(), grenzeAll)
	wantOK(t, "Beispiel", got, err, beispielErgebnis())
	got, err = leseMit(strings.NewReader("--batch_1--"), ctBeispiel(), grenzeAll)
	wantOK(t, "leerer Batch", got, err, ergebnis())
	got, err = leseMit(strings.NewReader(beispiel()), []string{"application/json"}, grenzeAll)
	wantFehler(t, "application/json an der Hülle", got, err, model.BatchFormat, "")
	for _, f := range []struct {
		name string
		body string
		u    model.BatchUrsache
		ort  string
	}{
		{"ohne --batch_1--", ersetze(t, "--batch_1--"+crlf, "", 0), model.BatchRahmen, ""},
		{"ohne --cs_1--", ersetze(t, "--cs_1--"+crlf, "", 0), model.BatchRahmen, "2"},
		{"Teil 1 ohne Content-Type", ersetze(t, zeileCTH+crlf, "", 0), model.BatchContentType, "1"},
		{"HTTP/1.0", ersetze(t, "Products(1) HTTP/1.1", "Products(1) HTTP/1.0", 0), model.BatchAnfragezeile, "1"},
		{"verschachtelt", ersetze(t, zeileCTH, "Content-Type: multipart/mixed; boundary=x", 1), model.BatchVerschachtelt, "2/1"},
		{"base64", ersetze(t, "Content-Transfer-Encoding: binary", "Content-Transfer-Encoding: base64", 0), model.BatchTransferKodierung, "1"},
	} {
		got, err := leseMit(strings.NewReader(f.body), ctBeispiel(), grenzeAll)
		wantFehler(t, f.name, got, err, f.u, f.ort)
	}
	for _, f := range []struct {
		name, alt, neu string
		aendere        func(r *model.BatchRequest)
	}{
		{"GET in 2/2", "PATCH $1", "GET $1", func(r *model.BatchRequest) { r.Methode = "GET" }},
		{"MERGE in 2/2", "PATCH $1", "MERGE $1", func(r *model.BatchRequest) { r.Methode = "MERGE" }},
		{"Content-ID 1 auch in 2/2", "Content-ID: 2", "Content-ID: 1", func(r *model.BatchRequest) { r.ContentID = "1" }},
		{"PATCH $3 in 2/2", "PATCH $1", "PATCH $3", func(r *model.BatchRequest) { r.Ziel = "$3" }},
	} {
		want := beispielErgebnis()
		f.aendere(&want.Teile[1].Requests[1])
		got, err := leseMit(strings.NewReader(ersetze(t, f.alt, f.neu, 0)), ctBeispiel(), grenzeAll)
		wantOK(t, f.name, got, err, want)
	}
}

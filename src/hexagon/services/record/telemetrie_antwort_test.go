package record_test

import (
	"context"
	"slices"
	"testing"
)

// antwortFall ist eine Antwort des Upstreams: Methode, Status, Header und Body, und ob sie nach HTTP einen Body hat (`SPEC-121` Nr. 5: nicht auf
// `HEAD`, nicht bei 1xx, 204 und 304; dann gibt es keine Aufnahme der Antwort).
type antwortFall struct {
	name, methode string
	status        int
	header        map[string][]string
	body          string
	hatBody       bool
}

// antwortFaelle ist die Menge M-Antwort aus dem Wortlaut von `SPEC-121` Nr. 5 und `SPEC-295` Nr. 7 a: je Statusklasse 1xx bis 5xx mindestens eine
// Antwort (101, 200, 204, 301, 304, 404, 500, 503), `HEAD` mit zwei Status, dazu der leere Body ohne Länge, mit `Content-Length` 0 und der Body mit
// gesetzter Länge, dazu beide Seiten jeder Grenze von `hatBody` (99 und 100, 199 und 200, 203 bis 206, 303 bis 305; Verifikation V-2): 20 Zellen, davon 7
// ohne Body nach HTTP. Der Status 99 ist kein gültiger Status eines Exchange (`SPEC-105` Nr. 10): der Exchange ist nicht angehängt, der Span endet
// wie bei jedem anderen Fehler des Anhängens.
func antwortFaelle() []antwortFall {
	json := map[string][]string{"Content-Type": {"application/json"}}
	mitLaenge := func(l string) map[string][]string {
		return map[string][]string{"Content-Type": {"application/json"}, "Content-Length": {l}}
	}
	return []antwortFall{
		{"HEAD 200", "HEAD", 200, mitLaenge("7"), "", false},
		{"HEAD 404", "HEAD", 404, mitLaenge("7"), "", false},
		{"99 mit Body, ungültiger Status", "GET", 99, json, `{"a":1}`, true},
		{"100 ohne Body", "GET", 100, nil, "", false},
		{"101", "GET", 101, nil, "", false},
		{"199 ohne Body", "GET", 199, nil, "", false},
		{"203 mit Body", "GET", 203, json, `{"a":1}`, true},
		{"204", "GET", 204, nil, "", false},
		{"205 mit Body", "GET", 205, json, `{"a":1}`, true},
		{"206 mit Body", "GET", 206, json, `{"a":1}`, true},
		{"303 mit Body", "GET", 303, json, `{"a":1}`, true},
		{"304", "GET", 304, nil, "", false},
		{"305 mit Body", "GET", 305, json, `{"a":1}`, true},
		{"200 leer ohne Länge", "GET", 200, json, "", true},
		{"200 leer mit Content-Length 0", "GET", 200, mitLaenge("0"), "", true},
		{"200 Body mit Content-Length", "GET", 200, mitLaenge("7"), `{"a":1}`, true},
		{"301 mit Body", "GET", 301, json, `{"a":1}`, true},
		{"404 mit Body", "GET", 404, json, `{"a":1}`, true},
		{"500 mit Body", "GET", 500, json, `{"a":1}`, true},
		{"503 mit Body", "GET", 503, json, `{"a":1}`, true},
	}
}

// TestEndeJedeAntwort hält F-2, F-12 und V-2 (SPEC-295 Nr. 3 b, 4 b, 7 a): Menge M-Antwort, 20 Zellen × 2 Wege (Close nach EOF, Close ohne EOF) gleich 40;
// je Zelle ein Span mit genau einem Set und einem End und dem Status der Aufnahme als Literal (`complete`; ohne EOF `incomplete` nur bei einer Antwort mit
// Body: eine Antwort ohne Body hat keine Aufnahme, die abbrechen könnte), unabhängig vom Status des Upstreams; der Exchange trägt den Status der Antwort.
func TestEndeJedeAntwort(t *testing.T) {
	faelle := antwortFaelle()
	ohneBody, n := 0, 0
	for _, f := range faelle {
		if !f.hatBody {
			ohneBody++
		}
		for _, weg := range []string{wegEOF, wegOhneEOF} {
			n++
			u, j := neuMitJournal(t)
			var koerperAntwort *fakeBody
			if f.body == "" {
				koerperAntwort = koerper()
			} else {
				koerperAntwort = koerper(f.body)
			}
			u.upstream.plane(antwort(f.status, f.header, koerperAntwort))
			a := anfrage(f.methode, "/a/x")
			soll := "complete"
			ant, err := u.record(context.Background(), a)
			if err != nil {
				t.Fatalf("%s: %v", f.name, err)
			}
			if weg == wegEOF {
				liesUndSchliesse(t, ant.Body)
			} else {
				schliesse(t, ant.Body)
				if f.hatBody {
					soll = "incomplete"
				}
			}
			was := f.name + ", " + weg
			if b := buecher(j.FakeTelemetry); len(b) != 1 {
				t.Errorf("%s: %d Bücher, erwartet 1", was, len(b))
			} else {
				pruefeBuch(t, was, b[0], "s1", soll)
			}
			want, angehaengt := []string{"Set", "ListBySession", "Append", "End"}, 1
			if f.status == 99 {
				want = []string{"Set", "ListBySession", "Append", "Log record: Exchange nicht angeh" + string([]byte{0xc3, 0xa4}) + "ngt", "End"}
				angehaengt = 0
			}
			if got := j.folge("Set", "ListBySession", "Append", want[3], "End"); !slices.Equal(got, want) {
				t.Errorf("%s: Folge %v, erwartet %v", was, got, want)
			}
			if l := u.exchanges.angehaengt(); len(l) != angehaengt || (angehaengt == 1 && l[0].ResponseStatus != f.status) {
				t.Errorf("%s: %d angehängt, erwartet %d mit dem Status %d der Antwort", was, len(l), angehaengt, f.status)
			}
		}
	}
	if len(faelle) != 20 || ohneBody != 7 || n != 40 {
		t.Errorf("%d Zellen, davon %d ohne Body, %d Läufe; erwartet 20, 7 und 40", len(faelle), ohneBody, n)
	}
}

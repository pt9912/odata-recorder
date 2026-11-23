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
// gesetzter Länge: 12 Zellen, davon 5 ohne Body nach HTTP.
func antwortFaelle() []antwortFall {
	json := map[string][]string{"Content-Type": {"application/json"}}
	mitLaenge := func(l string) map[string][]string {
		return map[string][]string{"Content-Type": {"application/json"}, "Content-Length": {l}}
	}
	return []antwortFall{
		{"HEAD 200", "HEAD", 200, mitLaenge("7"), "", false},
		{"HEAD 404", "HEAD", 404, mitLaenge("7"), "", false},
		{"101", "GET", 101, nil, "", false},
		{"204", "GET", 204, nil, "", false},
		{"304", "GET", 304, nil, "", false},
		{"200 leer ohne Länge", "GET", 200, json, "", true},
		{"200 leer mit Content-Length 0", "GET", 200, mitLaenge("0"), "", true},
		{"200 Body mit Content-Length", "GET", 200, mitLaenge("7"), `{"a":1}`, true},
		{"301 mit Body", "GET", 301, json, `{"a":1}`, true},
		{"404 mit Body", "GET", 404, json, `{"a":1}`, true},
		{"500 mit Body", "GET", 500, json, `{"a":1}`, true},
		{"503 mit Body", "GET", 503, json, `{"a":1}`, true},
	}
}

// TestEndeJedeAntwort hält F-2 und F-12 (SPEC-295 Nr. 3 b, 4 b, 7 a): Menge M-Antwort, 12 Zellen × 2 Wege (Close nach EOF, Close ohne EOF) gleich 24;
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
			if got, want := j.folge("Set", "ListBySession", "Append", "End"), []string{"Set", "ListBySession", "Append", "End"}; !slices.Equal(got, want) {
				t.Errorf("%s: Folge %v, erwartet %v", was, got, want)
			}
			if l := u.exchanges.angehaengt(); len(l) != 1 || l[0].ResponseStatus != f.status {
				t.Errorf("%s: angehängt %d, Status der Antwort im Exchange erwartet %d", was, len(l), f.status)
			}
		}
	}
	if len(faelle) != 12 || ohneBody != 5 || n != 24 {
		t.Errorf("%d Zellen, davon %d ohne Body, %d Läufe; erwartet 12, 5 und 24", len(faelle), ohneBody, n)
	}
}

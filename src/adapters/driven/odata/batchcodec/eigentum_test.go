package batchcodec_test

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/batchcodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ueberschreibe schreibt in die ganze Kapazität jedes Body, Header und Requests des Teils k.
func ueberschreibe(b model.Batch, k int) {
	r := b.Teile[k].Requests[:cap(b.Teile[k].Requests)]
	for i := range b.Teile[k].Requests {
		body := r[i].Body[:cap(r[i].Body)]
		for j := range body {
			body[j] = 'Z'
		}
		kopf := r[i].Header[:cap(r[i].Header)]
		for j := range kopf {
			kopf[j] = model.BatchKopf{Name: "Z", Wert: "Z"}
		}
	}
	for i := range r {
		r[i] = model.BatchRequest{Methode: "Z"}
	}
}

func eigentumBody() string {
	kopf := []string{"A: 1", "B: 2", "C: 3"}
	teil := func(body string) string { return block([]string{zeileCTH}, anfrage("POST x HTTP/1.1", kopf, body)) }
	return huelle(teil("eins"), changeset("c", teil("zwei"), teil("drei"), teil("vier")), changeset("d", teil("fuenf")), teil("sechs"))
}

// TestBatchEigentum hält Zeile 31: gleiche Eingabe, gleiches Ergebnis; das Ergebnis gehört dem Aufrufer, kein Wert teilt Speicher
// mit einem anderen.
func TestBatchEigentum(t *testing.T) {
	body := eigentumBody()
	frisch, err := lese(body)
	if err != nil {
		t.Fatal(err)
	}
	for k := range frisch.Teile {
		erst, err1 := lese(body)
		zweit, err2 := lese(body)
		if err1 != nil || err2 != nil || !reflect.DeepEqual(erst, frisch) || !reflect.DeepEqual(zweit, frisch) {
			t.Fatalf("Wiederholung: %v %v", err1, err2)
		}
		ueberschreibe(erst, k)
		for i := range erst.Teile {
			if i != k && !reflect.DeepEqual(erst.Teile[i], frisch.Teile[i]) {
				t.Errorf("Teil %d überschrieben: Teil %d geändert: %+v", k, i, erst.Teile[i])
			}
		}
		if !reflect.DeepEqual(zweit, frisch) {
			t.Errorf("Teil %d überschrieben: der zweite Aufruf ist geändert", k)
		}
	}
	// Innerhalb eines Teils: jeder Request für sich.
	for k := range 3 {
		erst, _ := lese(body)
		r := erst.Teile[1].Requests[k]
		b := r.Body[:cap(r.Body)]
		for j := range b {
			b[j] = 'Z'
		}
		h := r.Header[:cap(r.Header)]
		for j := range h {
			h[j] = model.BatchKopf{Name: "Z"}
		}
		for i := range 3 {
			if i != k && !reflect.DeepEqual(erst.Teile[1].Requests[i], frisch.Teile[1].Requests[i]) {
				t.Errorf("Request %d überschrieben: Request %d geändert", k, i)
			}
		}
	}
	// Der Strom behält seine Bytes nicht: eine Änderung an der Quelle nach dem Lesen ändert nichts.
	q := neueQuelle(body)
	got, err := leseMit(q, []string{ctBatch}, grenzeAll)
	for i := range q.data {
		q.data[i] = 'Z'
	}
	wantOK(t, "Quelle überschrieben", got, err, frisch)
}

// TestBatchGleichzeitig hält Zeile 32: gleichzeitige Aufrufe auf einer und auf mehreren Instanzen beeinflussen sich nicht; die vier
// Goroutinen lesen versetzt vier verschiedene Eingaben.
func TestBatchGleichzeitig(t *testing.T) {
	eingaben := []string{
		eigentumBody(),
		huelle(block([]string{zeileCTH}, anfrage("PUT y HTTP/1.1", nil, strings.Repeat("a", 40000)))),
		huelle(changeset("c", httpTeil("POST z HTTP/1.1")), httpTeil("GET x HTTP/1.0")),
		huelle(block([]string{zeileCTH}, anfrage("PUT y HTTP/1.1", nil, strings.Repeat("b", 50000)))),
	}
	type lauf struct {
		b   model.Batch
		err error
	}
	soll := make([]lauf, len(eingaben))
	for i, e := range eingaben {
		b, err := lese(e)
		soll[i] = lauf{b: b, err: err}
	}
	gemeinsam := batchcodec.NewLeser()
	var wg sync.WaitGroup
	for g := range 4 {
		wg.Go(func() {
			for i := range 40 {
				leser := gemeinsam
				if i%4 == 3 {
					leser = batchcodec.NewLeser()
				}
				k := (g + i/10) % len(eingaben)
				q := neueQuelle(eingaben[k])
				q.hoechstens = 1000
				b, err := leser.LeseBatch(t.Context(), q, []string{ctBatch}, grenzeAll)
				if !reflect.DeepEqual(b, soll[k].b) || !reflect.DeepEqual(err, soll[k].err) {
					t.Errorf("Goroutine %d, Eingabe %d: anderes Ergebnis (%v)", g, k, err)
					return
				}
			}
		})
	}
	wg.Wait()
}

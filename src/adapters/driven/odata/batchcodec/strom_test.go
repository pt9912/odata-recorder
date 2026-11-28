package batchcodec_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ursachenFall ist eine Fixture je Ursache des Lesers.
type ursachenFall struct {
	body, ct string
	u        model.BatchUrsache
	ort      string
	grenze   int64
}

func ursachenFaelle() []ursachenFall {
	viele := []string{zeileCTH}
	for range 100 {
		viele = append(viele, "X: "+geheim[:1])
	}
	return []ursachenFall{
		{huelle(block(nil, anfrage("GET "+geheim+" HTTP/1.1", nil, ""))), ctBatch, model.BatchContentType, "1", grenzeAll},
		{huelle(block([]string{"Content-Type: multipart/mixed; x=" + geheim}, "")), ctBatch, model.BatchBoundary, "1", grenzeAll},
		{huelle(httpTeil("GET " + geheim + " HTTP/1.1")), ctBatch, model.BatchZuGross, "", 20},
		{huelle(block(viele, anfrage("GET x HTTP/1.1", nil, ""))), ctBatch, model.BatchZuViele, "1", grenzeAll},
		{"--b" + crlf + geheim + crlf, ctBatch, model.BatchRahmen, "", grenzeAll},
		{huelle(block([]string{zeileCTH, geheim}, "")), ctBatch, model.BatchKopfzeile, "1", grenzeAll},
		{huelle(idTeil(geheim+"!", "GET x HTTP/1.1")), ctBatch, model.BatchContentID, "1", grenzeAll},
		{huelle(httpTeil("GET " + geheim + " HTTP/1.0")), ctBatch, model.BatchAnfragezeile, "1", grenzeAll},
		{huelle(), "application/json; x=" + geheim, model.BatchFormat, "", grenzeAll},
		{huelle(block([]string{"Content-Type: text/" + geheim}, "")), ctBatch, model.BatchTeilTyp, "1", grenzeAll},
		{huelle(changeset("c", changeset("d", httpTeil("GET "+geheim+" HTTP/1.1")))), ctBatch, model.BatchVerschachtelt, "1/1", grenzeAll},
		{huelle(block([]string{zeileCTH, "Content-Transfer-Encoding: " + geheim}, "")), ctBatch, model.BatchTransferKodierung, "1", grenzeAll},
	}
}

func gleichesErgebnis(a model.Batch, ae error, b model.Batch, be error) bool {
	return reflect.DeepEqual(a, b) && reflect.DeepEqual(ae, be)
}

// zerlegungen sind die Zerlegungen eines Stroms der Länge n.
func zerlegungen(data string) []func() *quelle {
	n := len(data)
	var out []func() *quelle
	mit := func(stueck []int, hoechstens int, eof bool) {
		out = append(out, func() *quelle {
			q := neueQuelle(data)
			q.stueck, q.hoechstens, q.eofMitDaten = append([]int(nil), stueck...), hoechstens, eof
			return q
		})
	}
	for i := range n + 1 {
		mit([]int{i}, 0, false)
		mit([]int{i}, 0, true)
		mit([]int{i, 0, 0}, 0, false)
	}
	mit(nil, 1, false)
	mit(nil, 1, true)
	var mitLeeren []int
	for range n {
		mitLeeren = append(mitLeeren, 0, 1)
	}
	mit(mitLeeren, 0, false)
	return out
}

// TestBatchZerlegung hält Zeile 6: gelesen wird allein mit Read bis io.EOF, nie Close; Ergebnis und Fehler hängen über das Beispiel
// und je eine Fixture je Ursache des Lesers nicht von der Zerlegung in Reads ab.
func TestBatchZerlegung(t *testing.T) {
	faelle := ursachenFaelle()
	if len(faelle) != 12 {
		t.Fatalf("%d Ursachen-Fixtures", len(faelle))
	}
	faelle = append(faelle, ursachenFall{body: beispiel(), ct: ctBeispiel()[0], grenze: grenzeAll})
	for _, f := range faelle {
		ganz := neueQuelle(f.body)
		want, wantErr := leseMit(ganz, []string{f.ct}, f.grenze)
		if f.u == "" {
			wantOK(t, "Beispiel", want, wantErr, beispielErgebnis())
		} else {
			wantFehler(t, string(f.u), want, wantErr, f.u, f.ort)
		}
		for i, bau := range zerlegungen(f.body) {
			q := bau()
			got, err := leseMit(q, []string{f.ct}, f.grenze)
			if !gleichesErgebnis(got, err, want, wantErr) {
				t.Fatalf("%s Zerlegung %d: %v %v", f.u, i, got, err)
			}
			if len(q.andere) != 0 || q.nachEnde != 0 {
				t.Fatalf("%s Zerlegung %d: andere Aufrufe %v, Reads nach dem Ende %d", f.u, i, q.andere, q.nachEnde)
			}
			if f.u == "" && !q.ende {
				t.Fatalf("Zerlegung %d: nicht bis zum Ende gelesen", i)
			}
		}
	}
}

// TestBatchStromfehler hält Zeile 7: ein Fehler von src an jeder Stelle, mit und ohne Bytes, bei gültigem und ungültigem Inhalt,
// ist kein BatchFehler und bleibt über errors.Is erreichbar.
func TestBatchStromfehler(t *testing.T) {
	eigener := errors.New("Platte")
	fehler := []error{eigener, io.ErrUnexpectedEOF, fmt.Errorf("umhüllt: %w", io.EOF), context.Canceled}
	for _, body := range []string{huelle(httpTeil("GET x HTTP/1.1")), huelle(block([]string{geheim}, "")), "kein Rahmen"} {
		for _, e := range fehler {
			for bei := range len(body) + 1 {
				for _, mitDaten := range []bool{false, true} {
					for _, hoechstens := range []int{0, 5} {
						q := neueQuelle(body)
						q.fehler, q.fehlerBei, q.mitDaten, q.hoechstens = e, bei, mitDaten, hoechstens
						got, err := leseMit(q, []string{ctBatch}, grenzeAll)
						var bf *model.BatchFehler
						if !errors.Is(err, e) || errors.As(err, &bf) || !reflect.DeepEqual(got, model.Batch{}) || q.nachEnde != 0 {
							t.Fatalf("%v bei %d (mitDaten %v): %v, Reads nach dem Fehler %d", e, bei, mitDaten, err, q.nachEnde)
						}
					}
				}
			}
		}
	}
}

// groesseFall liest body mit grenze in Stücken von höchstens stueck Bytes, mit eofMitDaten kommt das letzte Stück zusammen mit
// io.EOF, und prüft die angeforderten Längen.
func groesseFall(t *testing.T, body string, grenze int64, stueck int, eofMitDaten bool) (model.Batch, error) {
	t.Helper()
	q := neueQuelle(body)
	q.hoechstens, q.eofMitDaten = stueck, eofMitDaten
	got, err := leseMit(q, []string{ctBatch}, grenze)
	var geliefert int64
	for i, a := range q.angefordert {
		if a < 1 || (grenze != math.MaxInt64 && int64(a) > grenze+1-geliefert) {
			t.Fatalf("grenze %d, Read %d fordert %d bei %d gelieferten Bytes", grenze, i+1, a, geliefert)
		}
		geliefert += int64(q.geliefert[i])
	}
	if grenze < int64(len(body)) && (geliefert != grenze+1 || q.nachEnde != 0) {
		t.Fatalf("grenze %d: %d Bytes genommen, erwartet %d", grenze, geliefert, grenze+1)
	}
	return got, err
}

// TestBatchGroesse hält Zeile 8: grenze Bytes sind gültig, das nächste ist zu_gross, vor jedem Verstoß des Inhalts und vor dem
// Fehler von src im selben Read, ob das Ende des Stroms im selben oder in einem eigenen Read kommt; kein Read fordert mehr an als
// grenze+1 abzüglich der gelieferten Bytes.
func TestBatchGroesse(t *testing.T) {
	klein := huelle(httpTeil("GET x HTTP/1.1"))
	gross := huelle(block([]string{zeileCTH}, anfrage("POST x HTTP/1.1", nil, strings.Repeat("y", 70000))))
	n := 0
	for _, body := range []string{klein, gross} {
		l := int64(len(body))
		for _, grenze := range []int64{1, l - 2, l - 1, l, l + 1, math.MaxInt64} {
			for _, stueck := range []int{1, 7, len(body), 32768, 32769, 0} {
				for _, mitEOF := range []bool{false, true} {
					got, err := groesseFall(t, body, grenze, stueck, mitEOF)
					if grenze < l {
						wantFehler(t, fmt.Sprintf("grenze %d von %d, EOF mit Daten %v", grenze, l, mitEOF), got, err, model.BatchZuGross, "")
					} else if err != nil || len(got.Teile) != 1 {
						t.Errorf("grenze %d von %d, Stück %d, EOF mit Daten %v: %v", grenze, l, stueck, mitEOF, err)
					}
					n++
				}
			}
		}
	}
	if n != 2*6*6*2 {
		t.Errorf("%d Fälle", n)
	}
	groesseEnde(t)
	got, err := groesseFall(t, "x", 1, 0, false)
	wantFehler(t, "ein Byte bei grenze 1", got, err, model.BatchRahmen, "")
	for _, body := range []string{strings.Repeat("x", 11), "--b" + crlf + "kaputt" + crlf + "--b--"} {
		got, err := groesseFall(t, body, 10, 0, false)
		wantFehler(t, "Inhalt hinter der Grenze", got, err, model.BatchZuGross, "")
	}
	eigener := errors.New("Platte")
	for _, f := range []struct {
		bei    int
		grenze int64
		gross  bool
	}{{11, 10, true}, {20, 10, true}, {10, 10, false}, {9, 10, false}} {
		q := neueQuelle(strings.Repeat("x", 30))
		q.fehler, q.fehlerBei, q.mitDaten = eigener, f.bei, true
		got, err := leseMit(q, []string{ctBatch}, f.grenze)
		if f.gross {
			wantFehler(t, "Fehler von src mit dem Byte grenze+1", got, err, model.BatchZuGross, "")
		} else if err != eigener {
			t.Errorf("Fehler bei %d, grenze %d: %v", f.bei, f.grenze, err)
		}
	}
}

// groesseEnde führt beide Seiten der Grenze (Länge grenze-1 bis grenze+2) mit jedem Ende des Stroms (io.EOF oder ein anderer
// Fehler, mit den letzten Bytes oder in einem eigenen Read) und jeder Stückelung (am Stück, in 1000, in grenze Bytes, sodass das
// Byte grenze+1 allein kommt), auch an der Länge des Lesepuffers.
func groesseEnde(t *testing.T) {
	t.Helper()
	eigener := errors.New("Platte")
	n := 0
	for _, grenze := range []int{1, 2, 32767, 32768, 32769, 65536} {
		for l := grenze - 1; l <= grenze+2; l++ {
			for _, ende := range []error{io.EOF, eigener} {
				for _, mitDaten := range []bool{false, true} {
					for _, stueck := range []int{0, 1000, grenze} {
						q := neueQuelle(strings.Repeat("x", l))
						q.hoechstens = stueck
						if ende == io.EOF {
							q.eofMitDaten = mitDaten
						} else {
							q.fehler, q.fehlerBei, q.mitDaten = ende, l, mitDaten
						}
						got, err := leseMit(q, []string{ctBatch}, int64(grenze))
						was := fmt.Sprintf("grenze %d, Länge %d, Ende %v mit Daten %v, Stück %d", grenze, l, ende, mitDaten, stueck)
						switch {
						case l > grenze:
							wantFehler(t, was, got, err, model.BatchZuGross, "")
						case ende == io.EOF:
							wantFehler(t, was, got, err, model.BatchRahmen, "")
						case err != eigener || !reflect.DeepEqual(got, model.Batch{}):
							t.Errorf("%s: %v", was, err)
						}
						if q.nachEnde != 0 {
							t.Errorf("%s: %d Reads nach dem Ende", was, q.nachEnde)
						}
						n++
					}
				}
			}
		}
	}
	if n != 6*4*2*2*3 {
		t.Errorf("%d Fälle am Ende des Stroms", n)
	}
}

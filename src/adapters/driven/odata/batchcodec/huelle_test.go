package batchcodec_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ungelesen liest mit dem Content-Type ct und verlangt die Ursache u, ohne dass src berührt wird.
func ungelesen(t *testing.T, ct []string, u model.BatchUrsache) {
	t.Helper()
	q := neueQuelle(huelle())
	got, err := leseMit(q, ct, 100)
	wantFehler(t, strings.Join(ct, "|"), got, err, u, "")
	if q.reads != 0 || len(q.andere) != 0 {
		t.Errorf("%q: src gelesen (%d Reads, %v)", ct, q.reads, q.andere)
	}
}

// gelesen verlangt, dass der leere Batch mit der Grenzzeichenfolge g unter ct gelesen wird.
func gelesen(t *testing.T, ct, g string) {
	t.Helper()
	got, err := leseMit(strings.NewReader("--"+g+"--"), []string{ct}, grenzeAll)
	wantOK(t, ct, got, err, ergebnis())
}

// TestBatchHuelle hält Zeile 3: null bis drei Kopfzeilen, andere Medientypen und jede Schreibweise von multipart/mixed mit SP und
// HTAB am Rand; bei einem Verstoß wird src nicht gelesen.
func TestBatchHuelle(t *testing.T) {
	gelesen(t, ctBatch, "b")
	for _, ct := range [][]string{
		nil, {}, {ctBatch, ctBatch}, {ctBatch, "text/plain"}, {"text/plain", ctBatch}, {ctBatch, ctBatch, ctBatch}, {"", ctBatch},
		{ctBatch, ""}, {ctBatch, "application/json"}, {"application/json", "application/json"},
	} {
		ungelesen(t, ct, model.BatchContentType)
	}
	for _, typ := range []string{
		"", " ", ht, " " + ht + " ", "multipart/related", "text/plain", "multipart/mixedx", "xmultipart/mixed", "multipart /mixed",
		"multipart/mixed " + "x", "multipart/mixed" + lf, cr + "multipart/mixed", "multipart/mixed" + esB(0xC2, 0xA0),
		"multipart/mix" + esB(0xC4, 0xB1) + "d", "application/http", "boundary=b",
	} {
		ungelesen(t, []string{typ + "; boundary=b"}, model.BatchContentType)
		ungelesen(t, []string{typ}, model.BatchContentType)
	}
	n := jedeSchreibweise("multipart/mixed", func(s string) {
		got, err := leseMit(strings.NewReader("--b--"), []string{s + "; boundary=b"}, grenzeAll)
		if err != nil || got.Teile == nil || len(got.Teile) != 0 {
			t.Errorf("Schreibweise %q: %v", s, err)
		}
	})
	if n != 1<<14 {
		t.Errorf("%d Schreibweisen", n)
	}
	for _, rand := range []string{" ", ht, " " + ht, ht + " "} {
		gelesen(t, rand+"multipart/mixed"+rand+";"+rand+"boundary=b"+rand, "b")
	}
}

// TestBatchJSONFormat hält Zeile 4: application/json in jeder Schreibweise und mit Parametern ist `format`, src ungelesen.
func TestBatchJSONFormat(t *testing.T) {
	for _, param := range []string{"", ";", "; boundary=b", "; boundary=\"", "; charset=utf-8", ";boundary;boundary"} {
		for _, rand := range []string{"", " ", ht} {
			ungelesen(t, []string{rand + "application/json" + rand + param}, model.BatchFormat)
		}
	}
	n := jedeSchreibweise("application/json", func(s string) {
		q := neueQuelle("")
		got, err := leseMit(q, []string{s}, 100)
		if f, ok := err.(*model.BatchFehler); !ok || f.Ursache != model.BatchFormat || got.Teile != nil || q.reads != 0 {
			t.Errorf("Schreibweise %q: %v", s, err)
		}
	})
	if n != 1<<15 {
		t.Errorf("%d Schreibweisen", n)
	}
	// U+017F faltet nicht auf s.
	ungelesen(t, []string{"application/j" + esB(0xC5, 0xBF) + "on"}, model.BatchContentType)
}

// boundaryZeichen sind die Zeichen einer Grenzzeichenfolge aus SPEC-183 Nr. 1.
const boundaryZeichen = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'()+_,-./:=? "

func boundaryFall(t *testing.T, param, soll string) {
	t.Helper()
	ct := "multipart/mixed; " + param
	if soll == "" {
		ungelesen(t, []string{ct}, model.BatchBoundary)
		return
	}
	gelesen(t, ct, soll)
}

// TestBatchBoundary hält Zeile 5: alle 256 Bytes am Anfang, in der Mitte und am Ende, ohne und mit Anführungszeichen; die Längen
// 0, 1, 70, 71; fehlt, doppelt, ohne `=`, einzelnes Anführungszeichen; jede Schreibweise des Namens und fremde Parameter.
func TestBatchBoundary(t *testing.T) {
	n := 0
	for b := range 256 {
		c := string([]byte{byte(b)})
		erlaubt := strings.IndexByte(boundaryZeichen, byte(b)) >= 0
		for stelle, wert := range []string{c + "ab", "a" + c + "b", "ab" + c} {
			// in Anführungszeichen: genau die Zeichen der Menge, SP nicht am Ende
			soll := ""
			if erlaubt && (c != " " || stelle != 2) {
				soll = wert
			}
			boundaryFall(t, `boundary="`+wert+`"`, soll)
			// ohne Anführungszeichen: SP und HTAB am Rand entfallen, `;` beendet den Wert
			soll = ""
			switch {
			case (c == " " || c == ht) && stelle != 1:
				soll = "ab"
			case c == ";":
				soll = []string{"", "a", "ab"}[stelle]
			case erlaubt:
				soll = wert
			}
			boundaryFall(t, "boundary="+wert, soll)
			n += 2
		}
	}
	if n != 256*3*2 {
		t.Errorf("%d Fälle", n)
	}
	for _, l := range []int{0, 1, 70, 71} {
		w, soll := strings.Repeat("a", l), ""
		if l >= 1 && l <= 70 {
			soll = w
		}
		boundaryFall(t, "boundary="+w, soll)
		boundaryFall(t, `boundary="`+w+`"`, soll)
	}
	boundaryFall(t, `boundary=" ab"`, " ab")
	boundaryFall(t, `boundary="ab "`, "")
	boundaryFall(t, `boundary= "ab" `, "ab")
	boundaryFall(t, "boundary=a b", "a b")
	for _, param := range []string{
		"", "charset=utf-8", "xboundary=b", "boundaryx=b", "boundary", "boundary=b; boundary=b", "boundary=b; boundary=c",
		"boundary=c; boundary=b", "boundary; boundary=b", "boundary=b; boundary", `boundary="`, `boundary=b"`, `boundary="b`,
		`boundary=""`, "boundary b", "=b", "boundary=" + geheim + " " + ht + "x",
	} {
		boundaryFall(t, param, "")
	}
	if got := jedeSchreibweise("boundary", func(name string) { boundaryFall(t, name+"=b", "b") }); got != 1<<8 {
		t.Errorf("%d Schreibweisen", got)
	}
	for _, param := range []string{
		"charset=utf-8; boundary=b", "boundary=b; charset=utf-8", "; ;boundary=b;;", "xboundary=c; boundary=b; boundaryx=d",
		ht + "boundary" + ht + "=" + ht + "b", "a=b=c; boundary=b",
	} {
		boundaryFall(t, param, "b")
	}
}

package model_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// contentIDAlphabet sind die Zeichen einer Content-ID aus SPEC-183 Nr. 4 c im Wortlaut; das Orakel ist diese Liste, nie der
// Quelltext des Modells.
const contentIDAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"

// TestContentIDGueltig hält Zeile 21 am Modell: jedes der 256 Bytes am Anfang, in der Mitte und am Ende, die Längen 0, 1, 128 und
// 129. Menge 256 × 3 Stellen und 8 Längen (4 Längen × 2 Zeichen).
func TestContentIDGueltig(t *testing.T) {
	n := 0
	for b := range 256 {
		c := string([]byte{byte(b)})
		soll := strings.IndexByte(contentIDAlphabet, byte(b)) >= 0
		for _, wert := range []string{c + "ab", "a" + c + "b", "ab" + c} {
			if model.ContentIDGueltig(wert) != soll {
				t.Errorf("Byte %#02x in %q: erwartet %v", b, wert, soll)
			}
			n++
		}
		if model.ContentIDGueltig(c) != soll {
			t.Errorf("Byte %#02x allein: erwartet %v", b, soll)
		}
	}
	for _, l := range []struct {
		laenge int
		soll   bool
	}{{0, false}, {1, true}, {128, true}, {129, false}} {
		for _, z := range []string{"a", "~"} {
			if model.ContentIDGueltig(strings.Repeat(z, l.laenge)) != l.soll {
				t.Errorf("Länge %d aus %q: erwartet %v", l.laenge, z, l.soll)
			}
			n++
		}
	}
	if n != 256*3+8 {
		t.Errorf("Menge %d", n)
	}
}

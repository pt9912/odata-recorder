package workspace_test

import (
	"context"
	"fmt"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// keinLeerraum ist die Klasse „jedes andere Zeichen“ aus SPEC-129 Nr. 6 a als Menge von Bytefolgen, ohne Doppelte:
// jedes der 256 Bytes allein außer den vier Leerraum-Bytes (252; die Bytes ab 0x80 sind allein ungültiges UTF-8) und
// jede Rune der Tabellen unicode.Z, unicode.Cc und unicode.Cf als UTF-8 außer dem Leerzeichen und den Zeichen Tab, LF,
// CR. Die Folgen entstehen aus Werten, nicht aus Zeichen im Quelltext (AGENTS.md §3.18).
func keinLeerraum() [][]byte {
	gesehen := map[string]bool{}
	var r [][]byte
	hinzu := func(b []byte) {
		if !gesehen[string(b)] {
			gesehen[string(b)] = true
			r = append(r, b)
		}
	}
	ist := func(u rune) bool { return u == 0x20 || u == 0x09 || u == 0x0A || u == 0x0D }
	for b := 0; b <= 0xFF; b++ {
		if !ist(rune(b)) {
			hinzu([]byte{byte(b)})
		}
	}
	for _, tab := range []*unicode.RangeTable{unicode.Z, unicode.Cc, unicode.Cf} {
		for _, rg := range tab.R16 {
			for u := rune(rg.Lo); u <= rune(rg.Hi); u += rune(rg.Stride) {
				if !ist(u) {
					hinzu(utf8.AppendRune(nil, u))
				}
			}
		}
		for _, rg := range tab.R32 {
			for u := rune(rg.Lo); u <= rune(rg.Hi); u += rune(rg.Stride) {
				hinzu(utf8.AppendRune(nil, u))
			}
		}
	}
	return r
}

// leerraumFolge ist eine Folge aus n Leerraum-Zeichen (Leerzeichen, Tab, LF, CR in wechselnder Reihenfolge, versetzt
// um n); n = 0 ist die leere Folge.
func leerraumFolge(n int) []byte {
	z := []byte{0x20, 0x09, 0x0A, 0x0D}
	f := make([]byte, n)
	for i := range f {
		f[i] = z[(i+n)%4]
	}
	return f
}

// umhuelltBeidseitig liefert das gültige Manifest mit z vor und hinter dem Wert.
func umhuelltBeidseitig(z []byte) []byte {
	return append(append(append([]byte{}, z...), gueltig()...), z...)
}

// TestManifestLeerraumKlasse (Z20, SPEC-129 Nr. 6 a): Leerraum vor und hinter dem Wert sind genau Leerzeichen, Tab, LF
// und CR in beliebiger Länge und Folge; jedes andere Zeichen ist ODR-E3001. Mengen: die Längen 0 bis 130 an jeder der
// drei Stellen (davor, dahinter, beidseitig; 393 Fälle, angenommen) und die Klasse „anderes Zeichen“ (252 Bytes und
// die Runen von Z, Cc und Cf, ohne Doppelte) an denselben drei Stellen (abgewiesen). Prüfgrenze: die Länge des
// Leerraums ist unbeschränkt, geprüft sind 0 bis 130 (über der üblichen Größe eines Begrenzers von 4 bis 128). Grund:
// eine Zahl ohne Messpunkt wäre geraten. Folge: ein Leerraum über 130 Zeichen, den der Code begrenzte, bliebe
// unentdeckt; der Import schreibt das Manifest in der Kanonform ohne Leerraum (SPEC-127 Nr. 14), im Betrieb erreicht
// ihn nur ein fremder Schreiber.
func TestManifestLeerraumKlasse(t *testing.T) {
	anderes := keinLeerraum()
	if len(anderes) < 400 {
		t.Fatalf("Klasse %d Folgen, erwartet mindestens 400 (252 Bytes und die Runen von Z, Cc, Cf)", len(anderes))
	}
	t.Logf("Klasse: %d Folgen (252 Bytes, dazu Runen von Z, Cc, Cf ohne Doppelte)", len(anderes))
	angenommen, abgewiesen := 0, 0
	lauf := func(name string, m []byte, soll bool) {
		w := neu(t)
		w.baselineBytes("bsl-1", "svc", m)
		_, err := w.v.Create(context.Background(), "bsl-1", "ws-neu")
		if soll {
			angenommen++
			if err != nil {
				t.Errorf("%s: %v", name, err)
			}
			return
		}
		abgewiesen++
		pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
		if w.p.anzahl(lblTx) != 0 {
			t.Errorf("%s: Transaktion gelaufen", name)
		}
	}
	for n := 0; n <= 130; n++ {
		f := leerraumFolge(n)
		lauf(fmt.Sprintf("Leerraum %d davor", n), umhuellt(f, -1), true)
		lauf(fmt.Sprintf("Leerraum %d dahinter", n), umhuellt(f, 1), true)
		lauf(fmt.Sprintf("Leerraum %d beidseitig", n), umhuelltBeidseitig(f), true)
	}
	for _, z := range anderes {
		lauf(fmt.Sprintf("anderes %x davor", z), umhuellt(z, -1), false)
		lauf(fmt.Sprintf("anderes %x dahinter", z), umhuellt(z, 1), false)
		lauf(fmt.Sprintf("anderes %x beidseitig", z), umhuelltBeidseitig(z), false)
	}
	if angenommen != 131*3 || abgewiesen != 3*len(anderes) {
		t.Errorf("angenommen %d, abgewiesen %d, erwartet 393 und %d", angenommen, abgewiesen, 3*len(anderes))
	}
}

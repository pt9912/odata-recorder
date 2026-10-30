package workspace_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// maxFolgenLaenge ist die größte Länge, bis zu der TestManifestLeerraumKlasse jede Folge aus Leerzeichen, Tab, LF und
// CR führt (Prüfgrenze im Kopf des Tests).
const maxFolgenLaenge = 4

// keinLeerraum ist die Klasse „jedes andere Zeichen“ aus SPEC-129 Nr. 6 a als Menge von Bytefolgen, ohne Doppelte:
// (1) jedes der 256 Bytes allein außer den vier Leerraum-Bytes (252; die Bytes ab 0x80 sind allein ungültiges UTF-8),
// (2) jede Rune der Tabellen unicode.Z, unicode.Cc und unicode.Cf als UTF-8 außer dem Leerzeichen und den Zeichen Tab,
// LF, CR, (3) Lo und Hi jeder Range jeder Tabelle von unicode.Categories, unicode.Scripts und unicode.Properties,
// (4) zwölf einzelne Runen (Buchstabe, Han, Emoji, privat, kombinierend, nicht zugewiesen, römische Zahl, Mikrozeichen,
// Braille, Nichtzeichen, letzter Skalarwert, letztes Zeichen vor dem Ersatzbereich: U+00E9, U+4E2D, U+1F600, U+E000,
// U+0301, U+0378, U+2160, U+00B5, U+2800, U+FFFE, U+10FFFF, U+D7FF) und (5) sechs ungültige Bytefolgen
// (Ersatzzeichen-Kodierung, zwei überlange, über U+10FFFF, abgeschnitten, Fortsetzungsbyte ohne Kopf). Die Folgen
// entstehen aus Werten, nicht aus Zeichen im Quelltext (AGENTS.md §3.18).
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
	rune1 := func(u rune) {
		if utf8.ValidRune(u) && !ist(u) {
			hinzu(utf8.AppendRune(nil, u))
		}
	}
	for b := 0; b <= 0xFF; b++ {
		if !ist(rune(b)) {
			hinzu([]byte{byte(b)})
		}
	}
	alle := func(tab *unicode.RangeTable, f func(rune)) {
		for _, rg := range tab.R16 {
			for u := rune(rg.Lo); u <= rune(rg.Hi); u += rune(rg.Stride) {
				f(u)
			}
		}
		for _, rg := range tab.R32 {
			for u := rune(rg.Lo); u <= rune(rg.Hi); u += rune(rg.Stride) {
				f(u)
			}
		}
	}
	for _, tab := range []*unicode.RangeTable{unicode.Z, unicode.Cc, unicode.Cf} {
		alle(tab, rune1)
	}
	for _, m := range []map[string]*unicode.RangeTable{unicode.Categories, unicode.Scripts, unicode.Properties} {
		for _, tab := range m {
			for _, rg := range tab.R16 {
				rune1(rune(rg.Lo))
				rune1(rune(rg.Hi))
			}
			for _, rg := range tab.R32 {
				rune1(rune(rg.Lo))
				rune1(rune(rg.Hi))
			}
		}
	}
	for _, u := range []rune{0xE9, 0x4E2D, 0x1F600, 0xE000, 0x0301, 0x0378, 0x2160, 0xB5, 0x2800, 0xFFFE, 0x10FFFF, 0xD7FF} {
		rune1(u)
	}
	for _, b := range [][]byte{{0xED, 0xA0, 0x80}, {0xC0, 0x80}, {0xE0, 0x80, 0x80}, {0xF4, 0x90, 0x80, 0x80}, {0xF0, 0x9F, 0x98}, {0x80, 0x20}} {
		hinzu(b)
	}
	return r
}

// leerraumFolgen liefert alle Folgen aus Leerzeichen, Tab, LF und CR der Länge 0 bis maxLaenge, je Länge n die 4^n
// Zahlen von 0 bis 4^n - 1 zur Basis 4 (Ziffer 0 bis 3 für Leerzeichen, Tab, LF, CR), die höchste Ziffer vorn.
func leerraumFolgen(maxLaenge int) [][]byte {
	alphabet := []byte{0x20, 0x09, 0x0A, 0x0D}
	var r [][]byte
	for n := 0; n <= maxLaenge; n++ {
		anzahl := 1
		for i := 0; i < n; i++ {
			anzahl *= 4
		}
		for k := 0; k < anzahl; k++ {
			f := make([]byte, n)
			rest := k
			for i := n - 1; i >= 0; i-- {
				f[i] = alphabet[rest%4]
				rest /= 4
			}
			r = append(r, f)
		}
	}
	return r
}

// leerraumFolge ist eine Folge aus n Leerraum-Zeichen (Leerzeichen, Tab, LF, CR im Wechsel, versetzt um n); sie führt
// die Längen bis 130, die leerraumFolgen nicht erreicht.
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
// und CR; jedes andere Zeichen ist ODR-E3001. Mengen, je an drei Stellen (davor, dahinter, beidseitig):
//   - angenommen, Länge: der Wechsel aus Leerzeichen, Tab, LF, CR in den Längen 0 bis 130 (131 Folgen);
//   - angenommen, Folge: jede der 4^n Folgen aus den vier Zeichen der Längen 0 bis 4 (1 + 4 + 16 + 64 + 256 = 341); das
//     Orakel ist der Wortlaut von Nr. 6 a: jede Folge besteht nur aus den vier Zeichen, also wird sie angenommen;
//   - abgewiesen, Zeichen: die Klasse „anderes Zeichen“ aus keinLeerraum allein;
//   - abgewiesen, Mischung (nur davor und dahinter): jede Folge der Längen 0 bis 2 (21) mit einem Fremdzeichen aus
//     {NUL, VT, FF, NBSP, BOM} an jeder der n + 1 Stellen (57 Stellen mal 5 Zeichen).
//
// Prüfgrenzen (die Mengen sind unbeschränkt): Länge 130 (über der üblichen Größe eines Begrenzers von 4 bis 128) und
// Folgenlänge 4 (ein Muster aus Zeilenende und Leerzeile hat bis zu vier Zeichen: CR LF CR LF). Zeichen: der Test
// führt nicht alle 1 112 060 Skalarwerte (rund 80 s je Lauf, ein Vielfaches unter -race), sondern jedes Byte, die ganze
// Menge von Z, Cc und Cf und die Ränder jeder Range jeder Unicode-Tabelle, die Go nennt; jede Abweisung nach einer
// Klasse von Go (IsSpace, IsControl, IsPrint, IsGraphic, einer Kategorie oder eines Skripts) trifft damit ein Element.
// Folge: eine Abweisung, die genau ein Zeichen außerhalb dieser Mengen träfe, ein Leerraum über 130 Zeichen, den der
// Code begrenzte, und ein Muster aus mehr als vier Leerraum-Zeichen blieben unentdeckt; der Import schreibt das
// Manifest in der Kanonform ohne Leerraum (SPEC-127 Nr. 14), im Betrieb erreicht ihn nur ein fremder Schreiber.
func TestManifestLeerraumKlasse(t *testing.T) {
	anderes := keinLeerraum()
	if len(anderes) < 1000 {
		t.Fatalf("Klasse %d Folgen, erwartet mindestens 1000 (252 Bytes, Z, Cc, Cf, Ränder der Tabellen, 18 Stichproben)", len(anderes))
	}
	t.Logf("Klasse: %d Folgen", len(anderes))
	folgen := leerraumFolgen(maxFolgenLaenge)
	je := map[int]int{}
	verschieden := map[string]bool{}
	for _, f := range folgen {
		je[len(f)]++
		verschieden[string(f)] = true
		for _, b := range f {
			if b != 0x20 && b != 0x09 && b != 0x0A && b != 0x0D {
				t.Fatalf("Folge %x enthält das Byte %#x", f, b)
			}
		}
	}
	if len(folgen) != 341 || len(verschieden) != 341 || je[0] != 1 || je[1] != 4 || je[2] != 16 || je[3] != 64 || je[4] != 256 {
		t.Fatalf("%d Folgen, %d verschiedene, je Länge %v; erwartet 341 verschiedene, je Länge 1, 4, 16, 64, 256", len(folgen), len(verschieden), je)
	}
	for _, p := range [][]byte{{0x0D, 0x0A}, {0x0A, 0x0A}, {0x20, 0x20}, {0x09, 0x09}, {0x0D, 0x0D}, {0x0A, 0x0D}, {0x0D, 0x0A, 0x0D, 0x0A}} {
		if !verschieden[string(p)] {
			t.Fatalf("die Folge %x fehlt in der Menge", p)
		}
	}
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
		if err == nil {
			t.Errorf("%s: angenommen, erwartet ODR-E3001", name)
			return
		}
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
	for _, f := range folgen {
		lauf(fmt.Sprintf("Folge %x davor", f), umhuellt(f, -1), true)
		lauf(fmt.Sprintf("Folge %x dahinter", f), umhuellt(f, 1), true)
		lauf(fmt.Sprintf("Folge %x beidseitig", f), umhuelltBeidseitig(f), true)
	}
	for _, z := range anderes {
		lauf(fmt.Sprintf("anderes %x davor", z), umhuellt(z, -1), false)
		lauf(fmt.Sprintf("anderes %x dahinter", z), umhuellt(z, 1), false)
		lauf(fmt.Sprintf("anderes %x beidseitig", z), umhuelltBeidseitig(z), false)
	}
	fremd := [][]byte{{0x00}, {0x0B}, {0x0C}, {0xC2, 0xA0}, {0xEF, 0xBB, 0xBF}}
	mischung := 0
	for _, f := range leerraumFolgen(2) {
		for pos := 0; pos <= len(f); pos++ {
			for _, z := range fremd {
				m := bytes.Join([][]byte{f[:pos], z, f[pos:]}, nil)
				lauf(fmt.Sprintf("Mischung %x davor", m), umhuellt(m, -1), false)
				lauf(fmt.Sprintf("Mischung %x dahinter", m), umhuellt(m, 1), false)
				mischung += 2
			}
		}
	}
	if mischung != 570 {
		t.Errorf("%d Mischungen, erwartet 570 (57 Stellen mal 5 Zeichen mal 2 Seiten)", mischung)
	}
	if angenommen != 131*3+341*3 || abgewiesen != 3*len(anderes)+570 {
		t.Errorf("angenommen %d, abgewiesen %d, erwartet %d und %d", angenommen, abgewiesen, 131*3+341*3, 3*len(anderes)+570)
	}
}

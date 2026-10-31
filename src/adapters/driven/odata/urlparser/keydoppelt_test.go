package urlparser_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestKeyDoppelt (Zeile 30, SPEC-161 Nr. 5 c und Nr. 8): ein Name ist doppelt, wenn er nach der Dekodierung bytegleich ist; jede
// Abweichung von der Bytegleichheit ist ein verschiedener Name, auch die Schreibweise und die Unicode-Faltung. Menge 55: 9
// doppelte Namen (`key`, Ort `pfad/1`) und 46 Körper mit verschiedenen Namen, die gelesen werden, mit den Namen wie geschrieben in
// der Reihenfolge der URL; darunter sechs Paare, die sich nur an einer inneren Stelle unterscheiden und in Länge, erstem und
// letztem Byte gleichen (P1 bis P6). Die Zeichen jenseits von ASCII stehen als Bytewerte.
func TestKeyDoppelt(t *testing.T) {
	var gleiche []string
	gleiche = append(gleiche, "Id=1,Id=2", "Id=1,Id=1", "\xc3\xa4=1,\xc3\xa4=2", "\xc3\xa4=1,%C3%A4=2", "%49d=1,Id=2", "Id=1,I%64=2",
		"a=1,b=2,a=3", "a=1,b=2,c=3,b=4")
	var sechzehn []string
	for i := 1; i <= 15; i++ {
		sechzehn = append(sechzehn, fmt.Sprintf("n%d=1", i))
	}
	gleiche = append(gleiche, strings.Join(append(sechzehn, "n1=1"), ","))
	if len(gleiche) != 9 {
		t.Fatalf("%d doppelte Namen, erwartet 9", len(gleiche))
	}
	for _, k := range gleiche {
		pruefeFall(t, parserNeu(), fall{pfad: "/S(" + k + ")", u: keyU, ort: "pfad/1"})
	}

	// verschiedene Namen: je Körper die Namen in der Reihenfolge der URL; das Literal des i-ten Teils ist die Zahl i
	var verschieden [][]string
	ascii := []string{"Id", "id", "ID", "iD"}
	for _, a := range ascii {
		for _, b := range ascii {
			if a != b {
				verschieden = append(verschieden, []string{a, b})
			}
		}
	}
	if len(verschieden) != 12 {
		t.Fatalf("%d ASCII-Paare, erwartet 12", len(verschieden))
	}
	// Namen, die sich nur nach Unicode-Faltung gleichen: k, K und das Kelvin-Zeichen; s, S und das lange s; ä und Ä; die Gruppe
	// der Digraphen (klein, Titelform, groß); σ, ς und Σ
	faltung := [][2]string{
		{"k", "K"}, {"k", "\xe2\x84\xaa"}, {"K", "\xe2\x84\xaa"}, {"s", "\xc5\xbf"}, {"S", "\xc5\xbf"}, {"\xc3\xa4", "\xc3\x84"},
		{"\xc7\x86", "\xc7\x85"}, {"\xc7\x86", "\xc7\x84"}, {"\xc7\x85", "\xc7\x84"}, {"\xcf\x83", "\xcf\x82"}, {"\xcf\x83", "\xce\xa3"},
		{"\xcf\x82", "\xce\xa3"},
	}
	if len(faltung) != 12 {
		t.Fatalf("%d Faltungspaare, erwartet 12", len(faltung))
	}
	for _, p := range faltung {
		verschieden = append(verschieden, []string{p[0], p[1]}, []string{p[1], p[0]})
	}
	verschieden = append(verschieden,
		[]string{"Id", "Idx"}, []string{"Idx", "Id"}, // gemeinsamer Anfang, verschiedene Länge
		[]string{"a", "B", "A"}, []string{"Id", "id", "ID", "iD"}) // nicht benachbart
	// Paare mit einem inneren Unterschied: P1 Länge 3, Stelle 1; P2 bis P4 Länge 5, Stellen 1, 2, 3; P5 Länge 128, Stelle 64; P6 der
	// zweite Byte eines Zeichens jenseits von ASCII (ä gegen ö, gleiche Länge, gleiches erstes Byte der Mitte)
	for _, p := range [][2]string{
		{"aXb", "aYb"}, {"aXccb", "aYccb"}, {"acXcb", "acYcb"}, {"accXb", "accYb"},
		{wiederhole("a", 63) + "X" + wiederhole("a", 64), wiederhole("a", 63) + "Y" + wiederhole("a", 64)},
		{"a\xc3\xa4b", "a\xc3\xb6b"},
	} {
		verschieden = append(verschieden, []string{p[0], p[1]})
	}
	if len(verschieden) != 46 {
		t.Fatalf("%d Körper mit verschiedenen Namen, erwartet 46", len(verschieden))
	}
	for _, namen := range verschieden {
		var teile []string
		var soll []model.SchluesselTeil
		for i, n := range namen {
			teile = append(teile, fmt.Sprintf("%s=%d", n, i+1))
			soll = append(soll, teil(n, zahlLit(fmt.Sprint(i+1))))
		}
		pruefeFall(t, parserNeu(), fall{pfad: "/S(" + strings.Join(teile, ",") + ")", want: entity("S", soll...)})
	}
}

package urlparser_test

import (
	"strings"
	"testing"
)

// Die Zahlen im Pfad von `$expand` (SPEC-262 Nr. 4 c): die Zahl der Teile eines qualifizierten Namens ("mindestens zwei Namen"), die
// Zahl der Segmente eines Pfads und die Länge des Segments als Ganzes sind allein durch `rawQuery` (8192 Bytes, SPEC-160 Nr. 6)
// begrenzt. Das Orakel ist `exOrakel` (kein Code des Lesers); jede Eingabe steht roh und dekodiert, die Trenner `.` und `/` auch als
// `%2E`, `%2e` und `%2F`.

// exTeileMax ist die größte Zahl der Teile (und der Segmente), bei der die Query mit dem Präfix `N/N/` höchstens 8192 Bytes hat:
// 8 Bytes `$expand=`, 4 Bytes Präfix, 2k - 1 Bytes.
const exTeileMax = 4090

// exStufen sind die Zahlen von Teilen (und Segmenten), die geprüft werden: jede von 1 bis 300, darüber jede 97. (397 bis 4083) und
// die größte, 4090: 300 + 39 + 1 = 340 Zahlen. Der Aufwand je Eingabe wächst mit der Zahl; ein Schwellenwert (jede Zahl über N, jede
// Stelle ab dem N.), gleich wo er liegt, trifft die größte Zahl und ihre letzte Stelle.
func exStufen() []int {
	var s []int
	for k := 1; k <= 300; k++ {
		s = append(s, k)
	}
	for k := 397; k < exTeileMax; k += 97 {
		s = append(s, k)
	}
	return append(s, exTeileMax)
}

// exVoll ist die Zahl der Stufen, die vollständig (jede Stelle, jeder Inhalt) geprüft werden; darüber nur erste, mittlere und letzte.
const exVoll = 40

// exDefekt ist ein Inhalt an einer Stelle: roh, dekodiert und ob der Teil allein ein gültiger Name ist.
type exDefekt struct{ roh, dek string }

// exTeilDefekte sind 3 gültige und 7 ungültige Inhalte eines Teils eines qualifizierten Namens.
func exTeilDefekte() []exDefekt {
	return []exDefekt{{"b", "b"}, {exE, exE}, {"_0", "_0"}, {"", ""}, {"1", "1"}, {"-", "-"}, {exNBSP, exNBSP}, {"$ref", "$ref"},
		{"%20", " "}, {strings.Repeat("a", 129), strings.Repeat("a", 129)}}
}

// exStellenIdx sind die Stellen, die bei k Elementen geprüft werden: alle bis `exVoll`, darüber die erste, mittlere und letzte.
func exStellenIdx(k int) []int {
	if k <= exVoll {
		idx := make([]int, k)
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	return exErsteMitteLetzte(k)
}

// exErsteMitteLetzte sind die erste, mittlere und letzte Stelle von k Elementen ohne Doppelte (k=1: 1, k=2: 2, sonst 3).
func exErsteMitteLetzte(k int) []int {
	idx := []int{0}
	for _, i := range []int{k / 2, k - 1} {
		if i != idx[len(idx)-1] {
			idx = append(idx, i)
		}
	}
	return idx
}

// exTeileForm baut ein Segment aus k Teilen mit dem Trenner trenn; Teil idx trägt den Inhalt (idx < 0: keiner), jeder andere `a`.
func exTeileForm(k, idx int, inhalt, trenn string) string {
	teile := make([]string, k)
	for i := range teile {
		teile[i] = "a"
	}
	if idx >= 0 {
		teile[idx] = inhalt
	}
	return strings.Join(teile, trenn)
}

func exPruefeStellen(t *testing.T, roh, dek [4]string, n *int) {
	t.Helper()
	p := exParser()
	for s := range roh {
		exPruefe(t, p, exQ(roh[s]), exOrakel(dek[s]))
		*n++
	}
}

// TestExpandSegmentTeile (Zeile 12, SPEC-262 Nr. 4 c): die Zahl der Teile eines qualifizierten Namens ist unbeschränkt bis auf
// `rawQuery`; die Zahlen von `exStufen` (340, bis 4090, der größten, die in die Query passt) haben an jeder Stelle des Pfads (erstes,
// mittleres, letztes Segment, allein) das Ergebnis des Orakels. Achsen: Zahl der Teile (340) × Stelle des Teils mit dem Inhalt (bis 40
// Teile jede, darüber die erste, mittlere und letzte) × Inhalt (bis 40 Teile 10 Inhalte und 3 Schreibweisen des Trenners, darüber `b`
// und leer mit `.`) × 4 Stellen im Pfad; dazu je Zahl der Teile der Fall ohne Inhalt und bis 40 Teile verschiedene Namen. Menge:
// 820 × 10 × 3 × 4 + 300 × 3 × 2 × 4 + 340 × 4 + 40 × 3 × 4 = 98 400 + 7200 + 1360 + 480 = 107 440. Prüfgrenze: die Zahlen zwischen den
// Stufen über 300 sind nicht geprüft (Grund: der Aufwand wächst mit der Zahl, ein Lauf aller 4090 dauert mit dem Detektor Minuten);
// Folge: eine Regel, die nur eine Zahl dieser Lücken trifft und keine größere, bleibt ungefangen, eine Schwelle trifft die 4090.
func TestExpandSegmentTeile(t *testing.T) {
	trenner := [][2]string{{".", "."}, {"%2E", "."}, {"%2e", "."}}
	n := 0
	for _, k := range exStufen() {
		if t.Failed() {
			return
		}
		var roh, dek [4]string
		stellen := func(seg, segDek string) { roh, dek = exSegmentStellen(seg), exSegmentStellen(segDek) }
		stellen(exTeileForm(k, -1, "", "."), exTeileForm(k, -1, "", "."))
		exPruefeStellen(t, roh, dek, &n)
		if k <= exVoll {
			namen := make([]string, k)
			for i := range namen {
				namen[i] = "n" + strings.Repeat("x", i)
			}
			for _, tr := range trenner {
				stellen(strings.Join(namen, tr[0]), strings.Join(namen, tr[1]))
				exPruefeStellen(t, roh, dek, &n)
			}
		}
		defekte := exTeilDefekte()
		trennerK := trenner
		if k > exVoll {
			defekte = []exDefekt{{"b", "b"}, {"", ""}}
			trennerK = trenner[:1]
		}
		for _, idx := range exStellenIdx(k) {
			for _, d := range defekte {
				for _, tr := range trennerK {
					stellen(exTeileForm(k, idx, d.roh, tr[0]), exTeileForm(k, idx, d.dek, tr[1]))
					exPruefeStellen(t, roh, dek, &n)
				}
			}
		}
	}
	exZaehlt(t, "Teile × Stelle × Inhalt × Stellen im Pfad", n, 98400+7200+1360+480)
}

// exSegmentArten sind 6 zulässige und 5 unzulässige Segmente eines Pfads (SPEC-262 Nr. 4 c).
func exSegmentArten() []string {
	return []string{"b", "Ns.Typ", "a.b.c", "$ref", "$count", "$value", "", "*", "$levels", "a.", ".a"}
}

// TestExpandSegmentZahl (Zeile 11, 12, SPEC-262 Nr. 4 c): die Zahl der Segmente eines Pfads ist unbeschränkt bis auf `rawQuery`; jede
// Zahl von `exStufen` (340, bis 4090) hat mit einem Segment an jeder Stelle (bis 40 jede, darüber erste, mittlere, letzte) das Ergebnis
// des Orakels. Achsen: Zahl der Segmente × Stelle × Art des Segments (11; darüber `b`, leer und `*`) × Trenner `/` und `%2F` (bis 40);
// dazu alle Segmente von einer Art (40 × 11 × 2) und das Element als erstes, mittleres und letztes einer Liste (40 × 11 × 3).
// Menge: 820 × 11 × 2 + 300 × 3 × 3 + 880 + 1320 = 18 040 + 2700 + 880 + 1320 = 22 940. Prüfgrenze wie bei `TestExpandSegmentTeile`.
func TestExpandSegmentZahl(t *testing.T) {
	p := exParser()
	slash := [][2]string{{"/", "/"}, {"%2F", "/"}}
	pfad := func(m, idx int, art, trenn string) string {
		s := make([]string, m)
		for i := range s {
			s[i] = "N"
			if idx < 0 || i == idx {
				s[i] = art
			}
		}
		return strings.Join(s, trenn)
	}
	n := 0
	for _, m := range exStufen() {
		if t.Failed() {
			return
		}
		arten, trennerM := exSegmentArten(), slash
		if m > exVoll {
			arten, trennerM = []string{"b", "", "*"}, slash[:1]
		}
		for _, idx := range exStellenIdx(m) {
			for _, art := range arten {
				for _, tr := range trennerM {
					exPruefe(t, p, exQ(pfad(m, idx, art, tr[0])), exOrakel(pfad(m, idx, art, tr[1])))
					n++
				}
			}
		}
		if m > exVoll {
			continue
		}
		for _, art := range exSegmentArten() {
			for _, tr := range slash {
				exPruefe(t, p, exQ(pfad(m, -1, art, tr[0])), exOrakel(pfad(m, -1, art, tr[1])))
				n++
			}
			pf := pfad(m, m/2, art, "/")
			for _, liste := range []string{"p,q," + pf, "p," + pf + ",q", pf + ",p,q"} {
				exPruefe(t, p, exQ(liste), exOrakel(liste))
				n++
			}
		}
	}
	exZaehlt(t, "Segmente × Stelle × Art", n, 18040+2700+880+1320)
}

// TestExpandSegmentKreuz (Zeile 12, SPEC-262 Nr. 4 c): Zahl der Segmente und Zahl der Teile je Segment zusammen. Achsen: 1 bis 6
// Segmente × 1 bis 8 Teile je Segment (jedes Segment hat gleich viele) × Segment mit dem Inhalt × Teil mit dem Inhalt × Inhalt {`-`,
// `b`}; dazu die Fälle ohne Inhalt. Menge: 21 × 36 × 2 + 48 = 1560. Dazu große Paare aus {2, 5, 17, 64, 200} Segmenten × {2, 5, 17,
// 64, 200} Teilen mit höchstens 4000 Teilen insgesamt (21 Paare) × {kein Inhalt, `-` im ersten Teil des ersten Segments, im mittleren Teil
// des mittleren, im letzten Teil des letzten} = 84; die übrigen Paare aus Segment und Teil bleiben für diese Größen ungeprüft.
func TestExpandSegmentKreuz(t *testing.T) {
	p := exParser()
	n := 0
	bau := func(m, k, si, ti int, inhalt string) string {
		segs := make([]string, m)
		for i := range segs {
			if i == si {
				segs[i] = exTeileForm(k, ti, inhalt, ".")
			} else {
				segs[i] = exTeileForm(k, -1, "", ".")
			}
		}
		return strings.Join(segs, "/")
	}
	for m := 1; m <= 6; m++ {
		for k := 1; k <= 8; k++ {
			e := bau(m, k, -1, -1, "")
			exPruefe(t, p, exQ(e), exOrakel(e))
			n++
			for si := range m {
				for ti := range k {
					for _, inhalt := range []string{"-", "b"} {
						e := bau(m, k, si, ti, inhalt)
						exPruefe(t, p, exQ(e), exOrakel(e))
						n++
					}
				}
			}
		}
	}
	exZaehlt(t, "Segmente × Teile × Stelle × Inhalt", n, 21*36*2+48)
	grosse := []int{2, 5, 17, 64, 200}
	nGross := 0
	for _, m := range grosse {
		for _, k := range grosse {
			if m*k > 4000 {
				continue
			}
			e := bau(m, k, -1, -1, "")
			exPruefe(t, p, exQ(e), exOrakel(e))
			nGross++
			for _, pos := range [][2]int{{0, 0}, {m / 2, k / 2}, {m - 1, k - 1}} {
				e := bau(m, k, pos[0], pos[1], "-")
				exPruefe(t, p, exQ(e), exOrakel(e))
				nGross++
			}
		}
	}
	exZaehlt(t, "große Segmente × Teile", nGross, 21*4)
}

// TestExpandSegmentLaengeGesamt (Zeile 12, SPEC-262 Nr. 4 a, c): jeder Name hat 1 bis 128 Zeichen, gleich wie viele Teile das Segment
// trägt und wie lang es als Ganzes ist. Achsen: 1 bis 30 Teile × Breite {1, 2 Bytes} × {alle Teile 128 Zeichen, ein Teil mit 129
// Zeichen an erster, mittlerer, letzter Stelle} × 4 Stellen im Pfad. Menge: (30 + 87) × 2 × 4 = 936.
func TestExpandSegmentLaengeGesamt(t *testing.T) {
	n := 0
	for k := 1; k <= 30; k++ {
		for _, fuell := range []string{"a", exE} {
			name := func(l int) string { return strings.Repeat(fuell, l) }
			teile := func(idx int) string {
				s := make([]string, k)
				for i := range s {
					s[i] = name(128)
					if i == idx {
						s[i] = name(129)
					}
				}
				return strings.Join(s, ".")
			}
			segs := []string{teile(-1)}
			for _, idx := range exErsteMitteLetzte(k) {
				segs = append(segs, teile(idx))
			}
			for _, seg := range segs {
				st := exSegmentStellen(seg)
				exPruefeStellen(t, st, st, &n)
			}
		}
	}
	exZaehlt(t, "Teile × Breite × Länge × Stellen im Pfad", n, (30+87)*2*4)
}

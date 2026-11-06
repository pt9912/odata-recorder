package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// fehlerGoRoh schreibt die Codes als Raw-String-Literale.
func fehlerGoRoh(codes []string) string {
	var b strings.Builder
	b.WriteString("package model\n\nvar tabelle = []zeile{\n")
	for _, c := range codes {
		fmt.Fprintf(&b, "\t{`%s`, 1},\n", c)
	}
	b.WriteString("}\n")
	return b.String()
}

// mitLiteralen setzt rohe Zeilen vor den Schluss von fehler.go.
func mitLiteralen(codes []string, zeilen ...string) string {
	f := fehlerGo(codes)
	return strings.TrimSuffix(f, "}\n") + strings.Join(zeilen, "\n") + "\n}\n"
}

// formen sind die neun Schreibungen, die `^ODR-E[0-9]{4}$` verfehlen (aus Bytewerten gebaut).
func formen() []string {
	arabisch := string([]byte{0xD9, 0xA1, 0xD9, 0xA2, 0xD9, 0xA3, 0xD9, 0xA4})
	vollbreite := string([]byte{0xEF, 0xBC, 0x91, 0xEF, 0xBC, 0x92, 0xEF, 0xBC, 0x93, 0xEF, 0xBC, 0x94})
	return []string{
		"ODR-E123", "ODR-E12345", "ODR-e1234", "ODR-E1234 ", "ODR-E 1234", "ODR-",
		"ODR-E" + arabisch, "ODR-E" + vollbreite, "ODR-E1234\n",
	}
}

func formQuelle(quelle string, n int, x string) string {
	return fmt.Sprintf("Form: %s Zeile %d: %s hat nicht die Form", quelle, n, strconv.Quote(x))
}

// TestKatCLeser: jedes Literal `ODR-E` mit vier Ziffern in fehler.go ist ein Code, an jeder Stelle der Datei,
// in Anführungszeichen und als Raw-String. Menge 40 Codes × 2 Schreibungen = 80: je Code fehlt er in §4 und genau
// die Art 4 für ihn steht in der Ausgabe.
func TestKatCLeser(t *testing.T) {
	codes := codeListe()
	zaehle(t, "Codes", len(codes), 40)
	schreibungen := map[string]func([]string) string{"Anführungszeichen": fehlerGo, "Raw-String": fehlerGoRoh}
	zaehle(t, "Schreibungen", len(schreibungen), 2)
	for name, bau := range schreibungen {
		for _, e := range codes {
			m := grund(codes, nil)
			m[dFehler] = bau(codes)
			m[dSpez] = spezText(ohne(codes, e))
			erwarteBefunde(t, name+" "+e, m, fmt.Sprintf("%s in fehler.go (Zeile %d), nicht in §4", e, zeileVon(t, m[dFehler], e)))
		}
	}
	// Ein Name in zwei Schreibungen derselben Datei ist ein Code und doppelt.
	m := grund(codes, nil)
	m[dFehler] = mitLiteralen(codes, "\t{`"+codes[0]+"`, 1},")
	erwarteBefunde(t, "Anführungszeichen und Raw-String", m,
		fmt.Sprintf("%s doppelt in fehler.go, erstes Vorkommen Zeile %d", codes[0], zeileVon(t, m[dFehler], codes[0])))
}

// TestKatCIgnoriert: ein Literal, das nicht mit `ODR-` beginnt, und Kommentare sind keine Codes. Menge 6.
func TestKatCIgnoriert(t *testing.T) {
	codes := codeListe()[:3]
	zeilen := []string{
		"\t\"siehe ODR-E9001\",",
		"\t// \"ODR-E9002\"",
		"\t/* \"ODR-E9003\" \"ODR-E123\" */",
		"\t\"Invalid configuration\",",
		"\tODR_E9004: 1,",
		"\t`erste Zeile\nODR-E9005`,",
	}
	zaehle(t, "Formen", len(zeilen), 6)
	for _, f := range zeilen {
		m := grund(codes, nil)
		m[dFehler] = mitLiteralen(codes, f)
		erwarteOk(t, f, m)
	}
}

// TestKatCForm: ein Literal mit `ODR-` am Anfang, das die Form verfehlt, ist Befund Form je Vorkommen und kein
// Code. Menge 9, jede Form zweimal gesetzt: zwei Befunde Form, kein Befund doppelt.
func TestKatCForm(t *testing.T) {
	codes := codeListe()[:3]
	zaehle(t, "Formen", len(formen()), 9)
	for _, x := range formen() {
		lit := "\t" + strconv.QuoteToASCII(x) + ","
		m := grund(codes, nil)
		m[dFehler] = mitLiteralen(codes, lit, lit)
		zeilen := strings.Split(m[dFehler], "\n")
		var erste, zweite int
		for i, z := range zeilen {
			if z == lit && erste == 0 {
				erste = i + 1
			} else if z == lit {
				zweite = i + 1
			}
		}
		erwarteBefunde(t, strconv.Quote(x), m, formQuelle("fehler.go", erste, x), formQuelle("fehler.go", zweite, x))
	}
}

// TestKatCZiffern: die Ziffer jeder der vier Stellen ist ASCII 0 bis 9 und sonst nichts. Menge: 4 Stellen × 256
// Bytewerte = 1024; eine Ziffer ergibt einen Code (Art 4, er steht nicht in §4), jeder andere Byte Befund Form.
func TestKatCZiffern(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	for stelle := 0; stelle < 4; stelle++ {
		for b := 0; b < 256; b++ {
			wert := []byte("ODR-E1111")
			wert[5+stelle] = byte(b)
			quelle := "ODR-E1111"[:5+stelle] + fmt.Sprintf("\\x%02x", b) + "ODR-E1111"[6+stelle:]
			if b >= '0' && b <= '9' {
				quelle = string(wert)
			}
			m := grund(codes, nil)
			m[dFehler] = mitLiteralen(codes, "\t\""+quelle+"\",")
			zeile := zeileVon(t, m[dFehler], quelle)
			if b >= '0' && b <= '9' {
				erwarteBefunde(t, quelle, m, fmt.Sprintf("%s in fehler.go (Zeile %d), nicht in §4", wert, zeile), string(wert)+" fehlt im Handbuch")
			} else {
				erwarteBefunde(t, quelle, m, formQuelle("fehler.go", zeile, string(wert)))
			}
			n++
		}
	}
	zaehle(t, "Ziffern", n, 1024)
}

// TestKatCDoppelt: ein Code als Literal an mehreren Stellen ist Befund doppelt, einmal je Code, mit der Zeile des
// ersten Vorkommens. Menge 3 Orte des zweiten Vorkommens (direkt dahinter, Mitte, Ende) × 2 (zweimal, dreimal) = 6.
func TestKatCDoppelt(t *testing.T) {
	codes := codeListe()[:5]
	x := codes[2]
	orte := map[string]int{"dahinter": 3, "Mitte": 4, "Ende": 5}
	zaehle(t, "Orte", len(orte), 3)
	n := 0
	for ort, stelle := range orte {
		for _, mal := range []int{2, 3} {
			var liste []string
			liste = append(liste, codes[:stelle]...)
			for i := 1; i < mal; i++ {
				liste = append(liste, x)
			}
			liste = append(liste, codes[stelle:]...)
			m := grund(codes, nil)
			m[dFehler] = fehlerGo(liste)
			erwarteBefunde(t, fmt.Sprintf("%s ×%d", ort, mal), m,
				fmt.Sprintf("%s doppelt in fehler.go, erstes Vorkommen Zeile %d", x, zeileVon(t, m[dFehler], x)))
			n++
		}
	}
	zaehle(t, "Fälle", n, 6)
}

// TestKatCEingabe: Eingabefehler von fehler.go. Menge 5: Datei fehlt, Verzeichnis statt Datei, Literal ohne Ende,
// leere Datei, Datei nur mit Kommentaren.
func TestKatCEingabe(t *testing.T) {
	codes := codeListe()[:3]
	faelle := []struct {
		name  string
		bau   func(m map[string]string)
		exakt string // sonst nur der Anfang
		anf   string
	}{
		{"fehlt", func(m map[string]string) { delete(m, dFehler) }, "fehler.go fehlt: src/hexagon/model/fehler.go", ""},
		{"Verzeichnis", func(m map[string]string) { delete(m, dFehler); m[dFehler+"/x"] = "" }, "", "fehler.go nicht lesbar: "},
		{"ohne Ende", func(m map[string]string) { m[dFehler] = "package model\n\nvar x = \"ODR-E1234\n" }, "fehler.go nicht lesbar: kein Go (Zeile 3)", ""},
		{"leer", func(m map[string]string) { m[dFehler] = "" }, "fehler.go: kein Code gefunden", ""},
		{"Kommentare", func(m map[string]string) { m[dFehler] = "package model\n\n// \"ODR-E2001\"\n/* `ODR-E2002` */\n" }, "fehler.go: kein Code gefunden", ""},
	}
	zaehle(t, "Formen", len(faelle), 5)
	for _, f := range faelle {
		m := grund(codes, nil)
		f.bau(m)
		w := schreibe(t, m)
		c, o, e := fahre(w)
		erster, _, _ := strings.Cut(e, "\n")
		soll := "katalog: Eingabe: " + f.exakt
		ok := e == soll+"\n"
		if f.anf != "" {
			ok = strings.HasPrefix(e, "katalog: Eingabe: "+f.anf) && strings.Count(e, "\n") == 1
		}
		if c != 2 || o != "" || !ok {
			t.Errorf("%s: Exit %d stdout %q stderr %q (erste Zeile %q)", f.name, c, o, e, erster)
		}
	}
}

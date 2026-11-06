package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestKatAusnahmenLesen: je Zeile `<Code> | <Grund>`, getrennt am ersten Strich, beide Seiten getrimmt; leere
// Zeilen und die leere Datei gelten; Eingabefehler bei fehlender Datei und bei einer Zeile ohne Trennstrich.
// Menge 10: Datei fehlt, Zeile ohne Strich, leere Datei und sieben grüne Schreibungen (leere Zeile, mehrere leere
// Zeilen, Grund mit Strichen darin, Leerraum um beide Seiten, Zeile aus Leerraum, CRLF, ohne Zeilenumbruch am
// Ende). Die Ausnahme ist im grünen Fall wirksam: ihr Code hat keine Zeile im Handbuch.
func TestKatAusnahmenLesen(t *testing.T) {
	codes := codeListe()[:3]
	x := codes[1]
	gruen := func(ausnahmen string) map[string]string {
		m := grund(codes, []string{x})
		m[dAusnahmen] = ausnahmen
		return m
	}
	ok := map[string]string{
		"leere Zeile":          x + " | Grund\n\n",
		"Grund mit Strich":     x + " | Grund | mit Strich | darin\n",
		"Leerraum um Seiten":   " \t" + x + " \t|\t Grund \t\n",
		"Zeile aus Leerraum":   " \t \n" + x + " | Grund\n \t\n",
		"CRLF":                 x + " | Grund\r\n\r\n",
		"ohne Zeilenumbruch":   x + " | Grund",
		"mehrere leere Zeilen": "\n\n" + x + " | Grund\n\n\n",
	}
	for name, a := range ok {
		erwarteOk(t, name, gruen(a))
	}
	// Die leere Datei gilt: keine Ausnahme, also ist x Pflicht und fehlt im Handbuch nicht (grund() liefert die Zeile).
	m := grund(codes, nil)
	m[dAusnahmen] = ""
	erwarteOk(t, "leere Datei", m)
	m = grund(codes, nil)
	delete(m, dAusnahmen)
	erwarteEingabe(t, "Datei fehlt", m, "Ausnahmen fehlt: harness/katalog-ausnahmen.txt")
	m = gruen(x + " | Grund\n" + x + "\n\n" + codes[2] + " Grund\n")
	erwarteEingabe(t, "ohne Trennstrich", m, "Ausnahmen Zeile 2: Zeile ohne Trennstrich", "Ausnahmen Zeile 4: Zeile ohne Trennstrich")
	zaehle(t, "Formen", len(ok)+3, 10)
}

// TestKatAusnahmenPruefen: Form, leerer Grund, doppelt, Code nicht in fehler.go (Art 1, 5, 3, 6). Menge 8 Formen +
// 2 leere Gründe + 2 doppelt + 40 Codes, die in fehler.go fehlen und deren Ausnahme bleibt (der Code steht in §4).
func TestKatAusnahmenPruefen(t *testing.T) {
	codes := codeListe()
	basis := codes[:3]
	n := 0
	for _, x := range append(formen()[:3], formen()[4:6]...) {
		m := grund(basis, nil)
		m[dAusnahmen] = x + " | Grund\n"
		erwarteBefunde(t, "Form "+x, m, formQuelle("Ausnahmen", 1, x))
		n++
	}
	// Die Schreibung mit Leerraum hinten ist nach dem Trimmen ein Code und kein Befund Form; die übrigen drei Formen
	// trifft das Trimmen nicht.
	for _, x := range append(formen()[6:8], "ODR-E1234x") {
		m := grund(basis, nil)
		m[dAusnahmen] = x + " | Grund\n"
		erwarteBefunde(t, "Form "+x, m, formQuelle("Ausnahmen", 1, x))
		n++
	}
	for _, g := range []string{"", "  \t "} {
		m := grund(basis, []string{basis[1]})
		m[dAusnahmen] = basis[1] + " |" + g + "\n"
		erwarteBefunde(t, "leerer Grund", m, fmt.Sprintf("Ausnahme %s (Zeile 1) ohne Grund", basis[1]))
		n++
	}
	for _, mal := range []int{2, 3} {
		m := grund(basis, []string{basis[1]})
		m[dAusnahmen] = strings.Repeat(basis[1]+" | Grund\n", mal)
		erwarteBefunde(t, fmt.Sprintf("doppelt ×%d", mal), m, basis[1]+" doppelt in Ausnahmen, erstes Vorkommen Zeile 1")
		n++
	}
	for _, e := range codes {
		m := grund(codes, nil)
		m[dFehler] = fehlerGo(ohne(codes, e))
		m[dAusnahmen] = e + " | Grund\n"
		kuerze(m, dHandbuch, "| `"+e+"` | Die Bedeutung. |\n", "")
		erwarteBefunde(t, "nicht in fehler.go "+e, m, fmt.Sprintf("Ausnahme %s (Zeile 1) steht nicht in fehler.go", e))
		n++
	}
	zaehle(t, "Fälle", n, 8+2+2+40)
}

// TestKatAusnahmeGrundMitStrich: ein leerer Grund hinter einem Strich im Grund ist kein leerer Grund.
func TestKatAusnahmeGrundMitStrich(t *testing.T) {
	codes := codeListe()[:3]
	m := grund(codes, []string{codes[1]})
	m[dAusnahmen] = codes[1] + " | | zweiter\n"
	erwarteOk(t, "Grund beginnt mit Strich", m)
}

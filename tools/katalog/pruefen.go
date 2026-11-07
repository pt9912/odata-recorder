package main

import "strconv"

// pruefeFehlercodes liest die vier Quellen in ihrer Reihenfolge (fehler.go, Spezifikation, Ausnahmen,
// Handbuch) und vergleicht die Mengen. Eine Quelle, die fehlt, schließt nur ihre eigenen Prüfungen aus
// (SPEC-201 Nr. 8).
func pruefeFehlercodes(wurzel string) *bericht {
	b := &bericht{}
	var c, k, ausnahmen menge
	var h handbuch
	if t, ok := lies(wurzel, pfadFehler, "fehler.go", b); ok {
		c = liesC(t, b)
	}
	if t, ok := lies(wurzel, pfadSpez, "spezifikation.md", b); ok {
		k = liesK(t, b)
	}
	if t, ok := lies(wurzel, pfadAusnahmen, "Ausnahmen", b); ok {
		ausnahmen = liesAusnahmen(t, b)
	}
	if t, ok := lies(wurzel, pfadHandbuch, "Handbuch", b); ok {
		h = liesHandbuch(t, b)
	}
	if len(b.eingabe) == 0 {
		vergleiche(c, k, ausnahmen, h, b)
	}
	return b
}

// vergleiche prüft die Beziehungen der Mengen C, K, H und W (SPEC-201 Nr. 1, 7): jeder Code aus C steht in K
// (K ohne C ist kein Befund), eine Ausnahme steht in C, jeder Code aus H hat eine Zeile in der Tabelle des
// Handbuchs, und jede Nennung im Handbuch steht in C und nicht in den Ausnahmen. Die Reihenfolge der Ausgabe
// legt allein ausgeben fest.
func vergleiche(c, k, ausnahmen menge, h handbuch, b *bericht) {
	for name, zeile := range c {
		if _, da := k[name]; !da {
			b.befund(artOhneSpezifikation, name, zeile, name+" in fehler.go (Zeile "+strconv.Itoa(zeile)+"), nicht in §4")
		}
		_, ausnahme := ausnahmen[name]
		if _, tabelle := h.tabelle[name]; !ausnahme && !tabelle {
			b.befund(artFehlt, name, 0, name+" fehlt im Handbuch")
		}
	}
	for name, zeile := range ausnahmen {
		if _, da := c[name]; !da {
			b.befund(artAusnahmeOhneCode, name, zeile, "Ausnahme "+name+" (Zeile "+strconv.Itoa(zeile)+") steht nicht in fehler.go")
		}
	}
	for name, zeile := range h.nennung {
		z := strconv.Itoa(zeile)
		if _, da := c[name]; !da {
			b.befund(artUnbekannt, name, zeile, "Handbuch Zeile "+z+" nennt "+name+", der nicht in fehler.go steht")
		}
		if _, da := ausnahmen[name]; da {
			b.befund(artAusnahmeGenannt, name, zeile, "Handbuch Zeile "+z+" nennt "+name+", der in den Ausnahmen steht")
		}
	}
}

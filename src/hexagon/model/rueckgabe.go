package model

// Rueckgabe ist die Wahl des Clients im Header `Prefer` für die Antwort einer Mutation (SPEC-285); der Nullwert ist die Wahl ohne `Prefer`.
type Rueckgabe int

const (
	// RueckgabeKeine ist keine Wahl: kein Header, `return` mit anderem Wert oder nur andere Präferenzen (SPEC-285 Nr. 3).
	RueckgabeKeine Rueckgabe = iota
	// RueckgabeMinimal ist `return=minimal`.
	RueckgabeMinimal
	// RueckgabeEntity ist `return=representation`.
	RueckgabeEntity
)

// LiesRueckgabe liest die Wahl aus den Zeilen der Kopfzeilen `Prefer` (SPEC-285): ein Element je Kopfzeile, nil heißt kein Header. Das
// erste lesbare Element namens `return` entscheidet, auch wenn sein Wert keine Wahl ist; spätere Elemente haben keine Wirkung, eine Zeile
// reicht nie in die nächste. Die Funktion liest nichts außer zeilen, ändert nichts und liefert nie einen Fehler; `TestLiesRueckgabeZeit`
// hält ihre Zeit für sieben Formen langer Zeilen gegen lineares Wachstum, andere Formen bleiben ungeprüft (SPEC-285 Nr. 1).
func LiesRueckgabe(zeilen []string) Rueckgabe {
	for _, z := range zeilen {
		if r, entschieden := liesPreferZeile(z); entschieden {
			return r
		}
	}
	return RueckgabeKeine
}

// liesPreferZeile liest die Elemente einer Zeile (SPEC-285 Nr. 2) und sagt mit entschieden, ob ein lesbares Element `return` darin stand.
func liesPreferZeile(z string) (r Rueckgabe, entschieden bool) {
	i := 0
	for i < len(z) {
		if i = ueberLeerraum(z, i); i >= len(z) {
			break
		}
		if z[i] == ',' {
			i++
			continue
		}
		name := i
		for i < len(z) && istTokenByte(z[i]) {
			i++
		}
		if i > name && endeDesNamens(z, i) && asciiKlein(z[name:i]) == "return" {
			return liesReturnWert(z, i), true
		}
		i = elementEnde(z, i)
	}
	return RueckgabeKeine, false
}

// endeDesNamens sagt, ob ab i das Ende eines Namens steht: Leerraum, `=`, `;`, `,` oder das Ende der Zeile (SPEC-285 Nr. 2); jedes andere Byte
// macht das Element unlesbar.
func endeDesNamens(z string, i int) bool {
	return i >= len(z) || z[i] == ' ' || z[i] == '\t' || z[i] == '=' || z[i] == ';' || z[i] == ','
}

// liesReturnWert liest hinter dem Namen `return` das Wort und wählt (SPEC-285 Nr. 3): jede andere Form als `=` mit einem Token oder einem
// Wort in Anführungszeichen, gefolgt von Leerraum und Ende, `;` oder `,`, ist RueckgabeKeine; `Rest` hinter `;` bleibt ungelesen.
func liesReturnWert(z string, i int) Rueckgabe {
	if i = ueberLeerraum(z, i); i >= len(z) || z[i] != '=' {
		return RueckgabeKeine
	}
	i = ueberLeerraum(z, i+1)
	var wort string
	if i < len(z) && z[i] == '"' {
		w, naechste, ok := liesAnfuehrung(z, i)
		if !ok {
			return RueckgabeKeine
		}
		wort, i = w, naechste
	} else {
		start := i
		for i < len(z) && istTokenByte(z[i]) {
			i++
		}
		wort = z[start:i]
	}
	if i = ueberLeerraum(z, i); i < len(z) && z[i] != ';' && z[i] != ',' {
		return RueckgabeKeine
	}
	switch asciiKlein(wort) {
	case "minimal":
		return RueckgabeMinimal
	case "representation":
		return RueckgabeEntity
	}
	return RueckgabeKeine
}

// liesAnfuehrung liest ab dem `"` bei i ein Wort in Anführungszeichen: ein Backslash ersetzt sich durch das folgende Byte; ok ist falsch,
// wenn das Paar in der Zeile nicht schließt.
func liesAnfuehrung(z string, i int) (wort string, naechste int, ok bool) {
	var b []byte
	for j := i + 1; j < len(z); j++ {
		switch {
		case z[j] == '"':
			return string(b), j + 1, true
		case z[j] == '\\' && j+1 < len(z):
			j++
			b = append(b, z[j])
		case z[j] == '\\':
			return "", len(z), false
		default:
			b = append(b, z[j])
		}
	}
	return "", len(z), false
}

// elementEnde liefert die Stelle des ersten `,` ab i außerhalb eines Anführungszeichen-Paares oder das Ende der Zeile; ein Paar, das nicht
// schließt, reicht bis zum Ende der Zeile.
func elementEnde(z string, i int) int {
	for i < len(z) {
		switch z[i] {
		case ',':
			return i
		case '"':
			i++
			for i < len(z) && z[i] != '"' {
				if z[i] == '\\' {
					i++
				}
				i++
			}
			i++
		default:
			i++
		}
	}
	return len(z)
}

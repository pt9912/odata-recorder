package model

// Wahrheit ist ein dreiwertiger Wahrheitswert (SPEC-166 Nr. 5). Der Nullwert ist WahrheitUnbekannt.
type Wahrheit int

// Die drei Wahrheitswerte; jede andere Zahl gilt als unbekannt.
const (
	WahrheitUnbekannt Wahrheit = iota
	WahrheitFalsch
	WahrheitWahr
)

// Durchlaesst sagt, ob ein Datensatz mit diesem Ergebnis den Filter passiert: nur WahrheitWahr (LH-FA-08.b).
func (w Wahrheit) Durchlaesst() bool { return w == WahrheitWahr }

// Und verknüpft zwei Wahrheitswerte: falsch, wenn ein Operand falsch ist, wahr, wenn beide wahr sind, sonst unbekannt.
func Und(a, b Wahrheit) Wahrheit {
	switch {
	case a == WahrheitFalsch || b == WahrheitFalsch:
		return WahrheitFalsch
	case a == WahrheitWahr && b == WahrheitWahr:
		return WahrheitWahr
	}
	return WahrheitUnbekannt
}

// Oder verknüpft zwei Wahrheitswerte: wahr, wenn ein Operand wahr ist, falsch, wenn beide falsch sind, sonst unbekannt.
func Oder(a, b Wahrheit) Wahrheit {
	switch {
	case a == WahrheitWahr || b == WahrheitWahr:
		return WahrheitWahr
	case a == WahrheitFalsch && b == WahrheitFalsch:
		return WahrheitFalsch
	}
	return WahrheitUnbekannt
}

// Nicht vertauscht falsch und wahr und lässt unbekannt.
func Nicht(a Wahrheit) Wahrheit {
	switch a {
	case WahrheitFalsch:
		return WahrheitWahr
	case WahrheitWahr:
		return WahrheitFalsch
	}
	return WahrheitUnbekannt
}

// WahrheitVon deutet einen Wert als Wahrheit: `true` und `false` der Art boolean, ArtNull als unbekannt.
// Jede andere Art und jeder andere Text einer Art boolean ist ErrUngueltig (SPEC-166 Nr. 5, 7).
func WahrheitVon(w Wert) (Wahrheit, error) {
	switch w.Art {
	case ArtNull:
		return WahrheitUnbekannt, nil
	case ArtBoolean:
		switch w.Text {
		case "true":
			return WahrheitWahr, nil
		case "false":
			return WahrheitFalsch, nil
		}
		return WahrheitUnbekannt, formFehler("Text der Art boolean")
	}
	return WahrheitUnbekannt, formFehler("Art ohne Wahrheit")
}

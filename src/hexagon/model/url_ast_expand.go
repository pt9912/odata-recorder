package model

// ExpandElement ist ein gelesenes Element von `$expand`: der Name der Navigation, dekodiert und wie geschrieben (SPEC-262 Nr. 5).
type ExpandElement struct {
	Name string
}

// ExpandOption ist der gelesene Wert von `$expand`: mindestens ein Element in der Reihenfolge der URL, ohne doppelte
// (SPEC-262 Nr. 5). Das Feld `Optionen.Expand` ist `nil`, wenn die URL `$expand` nicht nennt.
type ExpandOption struct {
	Elemente []ExpandElement
}

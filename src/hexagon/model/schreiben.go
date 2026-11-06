package model

// Kopfzeile ist eine Kopfzeile der Antwort: Name und Wert (SPEC-168 Nr. 7).
type Kopfzeile struct {
	Name, Wert string
}

// Darstellung ist der Kopf, den ein Schreiber zu den Bytes liefert: der vollständige Wert von `Content-Type` und die
// Kopfzeile, die die Version der Darstellung nennt (SPEC-168 Nr. 7). Bei jedem Fehler ist sie der Nullwert.
type Darstellung struct {
	ContentType string
	Kennzeichen Kopfzeile
}

// EntityAntwort ist die Eingabe von SchreibeEntity (SPEC-168): Wurzel steht vor `/<Set>`, Auswahl sind die Properties
// von `$select`, Accept die Werte der Kopfzeilen `Accept` der Anfrage in ihrer Reihenfolge.
type EntityAntwort struct {
	Wurzel, Set string
	Entity      Entity
	Auswahl     []string
	Accept      []string
}

// CollectionAntwort ist die Eingabe von SchreibeCollection (SPEC-168); Anzahl ist die Gesamtzahl, die der Aufrufer meldet.
type CollectionAntwort struct {
	Wurzel, Set string
	Entities    []Entity
	Anzahl      Anzahl
	Auswahl     []string
	Accept      []string
}

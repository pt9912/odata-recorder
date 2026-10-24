package jsoncodec

// art ist die JSON-Art eines Knotens.
type art uint8

const (
	artObjekt art = iota + 1
	artArray
	artString
	artZahl
	artWahr
	artFalsch
	artNull
)

// knoten ist ein gelesener JSON-Wert. Ein Objekt hält die Namen in der Reihenfolge des Dokuments und die Werte
// parallel dazu; ein Array hält seine Elemente in kinder.
type knoten struct {
	art    art
	marker bool   // String: der Rohtext ist genau `[REDACTED]` (kein Escape)
	text   string // String: die dekodierten Zeichen; Zahl: der Rohtext
	kinder []knoten
	namen  []string
}

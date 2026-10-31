package model

// RessourcenArt ist die Art der Ressource einer OData-URL (SPEC-160 Nr. 2). Der Nullwert ist keine Art.
type RessourcenArt int

// Die vier Arten der Ressource, die der Mock liest (SPEC-161 Nr. 3, 7).
const (
	RessourceMetadaten RessourcenArt = iota + 1 // `/$metadata`
	RessourceMenge                              // `/Set`
	RessourceEntity                             // `/Set(Schlüssel)`
	RessourceAnzahl                             // `/Set/$count`
)

// LiteralArt ist die Art eines Schlüsselliterals (SPEC-161 Nr. 6). Der Nullwert ist keine Art.
type LiteralArt int

// Die sechs Literalarten.
const (
	LiteralZeichenkette LiteralArt = iota + 1
	LiteralWahrheitswert
	LiteralZahl
	LiteralGUID
	LiteralDatum
	LiteralZeitpunkt
)

// Literal ist ein Schlüsselliteral der URL; Text ist die Literalform des Schlüssels (SPEC-126 Nr. 11).
type Literal struct {
	Art  LiteralArt
	Text string
}

// SchluesselTeil ist ein Teil des Schlüssels in der Reihenfolge der URL; Name ist beim einen unbenannten Teil leer.
type SchluesselTeil struct {
	Name    string
	Literal Literal
}

// Ressource ist der Pfad der URL (SPEC-160 Nr. 2).
type Ressource struct {
	Art        RessourcenArt
	EntitySet  string
	Schluessel []SchluesselTeil
}

// SelectOption ist der gelesene Wert von `$select`; Alle steht für `*`.
type SelectOption struct {
	Alle  bool
	Namen []string
}

// Sortierung ist ein Element von `$orderby`.
type Sortierung struct {
	Name       string
	Absteigend bool
}

// ZahlOption ist `$top` oder `$skip`; Gesetzt unterscheidet „auf Null gesetzt“ von „nicht gesetzt“.
type ZahlOption struct {
	Gesetzt bool
	Wert    int
}

// WahrOption ist `$count`.
type WahrOption struct {
	Gesetzt bool
	Wert    bool
}

// FilterAusdruck ist der Ausdruck, den ein Leser von `$filter` liefert; die Typen des Filter-AST setzen ihn um. Er ist ein
// Alias der anonymen Schnittstelle: eine Umsetzung von urlparser.FilterLeser schreibt `interface{ IstFilterAusdruck() }` als
// Ergebnistyp, denn `ireturn` meldet eine benannte Schnittstelle als Ergebnis, eine anonyme nicht.
type FilterAusdruck = interface {
	IstFilterAusdruck()
}

// FilterOption trägt den dekodierten Wert von `$filter` und, wenn ein Leser eingehängt ist, den Ausdruck.
type FilterOption struct {
	Text     string
	Ausdruck FilterAusdruck
}

// SonstigeOption ist eine Option ohne `$` (herstellerspezifisch oder Parameter-Alias), dekodiert und ohne Wirkung.
type SonstigeOption struct {
	Name, Wert string
}

// Optionen trägt je Systemoption ein Feld; was die URL nicht nennt, ist der Nullwert (SPEC-160 Nr. 2).
type Optionen struct {
	Select   *SelectOption
	OrderBy  []Sortierung
	Top      ZahlOption
	Skip     ZahlOption
	Count    WahrOption
	Filter   *FilterOption
	Sonstige []SonstigeOption
}

// ODataURL ist das Ergebnis des Lesens einer OData-URL (SPEC-160): ein Wert aus Texten und Zahlen.
type ODataURL struct {
	Ressource Ressource
	Optionen  Optionen
}

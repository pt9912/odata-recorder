package model

// VergleichOp ist der Operator eines FilterVergleich; der Wert ist sein Text in Kleinbuchstaben (SPEC-165).
type VergleichOp string

// Die sechs Vergleichsoperatoren.
const (
	VergleichEq VergleichOp = "eq"
	VergleichNe VergleichOp = "ne"
	VergleichGt VergleichOp = "gt"
	VergleichGe VergleichOp = "ge"
	VergleichLt VergleichOp = "lt"
	VergleichLe VergleichOp = "le"
)

// VerknuepfungOp ist der Operator einer FilterVerknuepfung; der Wert ist sein Text in Kleinbuchstaben (SPEC-165).
type VerknuepfungOp string

// Die zwei Verknüpfungsoperatoren.
const (
	VerknuepfungUnd  VerknuepfungOp = "and"
	VerknuepfungOder VerknuepfungOp = "or"
)

// FilterVergleich ist `Links Op Rechts` (SPEC-165). Links steht im Text vor Rechts.
type FilterVergleich struct {
	Op     VergleichOp
	Links  FilterAusdruck
	Rechts FilterAusdruck
}

// FilterVerknuepfung ist `Links and Rechts` bzw. `Links or Rechts` (SPEC-165).
type FilterVerknuepfung struct {
	Op     VerknuepfungOp
	Links  FilterAusdruck
	Rechts FilterAusdruck
}

// FilterNicht ist `not Operand` (SPEC-165).
type FilterNicht struct {
	Operand FilterAusdruck
}

// FilterEigenschaft ist der Name einer Property, wie geschrieben (SPEC-165).
type FilterEigenschaft struct {
	Name string
}

// FilterLiteral ist ein Literal; Literal.Text ist die kanonische Literalform des Schlüssels (SPEC-165).
type FilterLiteral struct {
	Literal Literal
}

// FilterNull ist das Wort `null` (SPEC-165).
type FilterNull struct{}

// IstFilterAusdruck macht FilterVergleich zu einem FilterAusdruck.
func (FilterVergleich) IstFilterAusdruck() {}

// IstFilterAusdruck macht FilterVerknuepfung zu einem FilterAusdruck.
func (FilterVerknuepfung) IstFilterAusdruck() {}

// IstFilterAusdruck macht FilterNicht zu einem FilterAusdruck.
func (FilterNicht) IstFilterAusdruck() {}

// IstFilterAusdruck macht FilterEigenschaft zu einem FilterAusdruck.
func (FilterEigenschaft) IstFilterAusdruck() {}

// IstFilterAusdruck macht FilterLiteral zu einem FilterAusdruck.
func (FilterLiteral) IstFilterAusdruck() {}

// IstFilterAusdruck macht FilterNull zu einem FilterAusdruck.
func (FilterNull) IstFilterAusdruck() {}

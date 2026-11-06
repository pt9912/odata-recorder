package model

import "strconv"

// knotenTyp ist der Typ eines Knotens des Filter-AST (SPEC-167 Nr. 2): seine Klasse; für eine FilterEigenschaft
// zusätzlich der EDM-Typ, für ein LiteralZahl der Text.
type knotenTyp struct {
	klasse    wertKlasse
	edmTyp    string
	zahlText  string
	zahlLiteral bool
}

var wahrheitTyp = knotenTyp{klasse: klasseWahrheit}

// PruefeFilter prüft den Baum gegen die wirksamen Properties von typ (SPEC-167). beobachtet fragt die
// Property-Maske, nil heißt: alle beobachtet. Das Ergebnis ist nil, ein *FilterTypFehler oder ErrUngueltig für
// einen Programmfehler des Aufrufers; der Baum bleibt unverändert.
func PruefeFilter(typ EntityType, ausdruck FilterAusdruck, beobachtet func(name string) bool) error {
	p := filterPruefer{typ: typ, beobachtet: beobachtet}
	t, err := p.knoten(ausdruck)
	if err != nil {
		return err
	}
	return verlangeWahrheit(t)
}

type filterPruefer struct {
	typ        EntityType
	beobachtet func(name string) bool
}

func verstoss(u FilterTypUrsache) error { return &FilterTypFehler{Ursache: u} }

// verlangeWahrheit verlangt Wahrheit oder null (SPEC-167 Nr. 3 b, 3 d).
func verlangeWahrheit(t knotenTyp) error {
	if t.klasse != klasseWahrheit && t.klasse != klasseNull {
		return verstoss(FilterNichtBoolesch)
	}
	return nil
}

// knoten prüft einen Teilbaum von links nach rechts, die Kinder vor der Regel des Knotens; der erste Verstoß endet
// die Prüfung (SPEC-167 Nr. 3).
func (p filterPruefer) knoten(a FilterAusdruck) (knotenTyp, error) {
	switch n := a.(type) {
	case FilterEigenschaft:
		return p.eigenschaft(n.Name)
	case FilterLiteral:
		return literalTyp(n.Literal)
	case FilterNull:
		return knotenTyp{klasse: klasseNull}, nil
	case FilterNicht:
		t, err := p.knoten(n.Operand)
		if err != nil {
			return knotenTyp{}, err
		}
		return wahrheitTyp, verlangeWahrheit(t)
	case FilterVerknuepfung:
		return p.verknuepfung(n)
	case FilterVergleich:
		return p.vergleich(n)
	}
	return knotenTyp{}, formFehler("Knoten des Filters")
}

func (p filterPruefer) verknuepfung(n FilterVerknuepfung) (knotenTyp, error) {
	l, err := p.knoten(n.Links)
	if err != nil {
		return knotenTyp{}, err
	}
	if err := verlangeWahrheit(l); err != nil {
		return knotenTyp{}, err
	}
	r, err := p.knoten(n.Rechts)
	if err != nil {
		return knotenTyp{}, err
	}
	if err := verlangeWahrheit(r); err != nil {
		return knotenTyp{}, err
	}
	if n.Op != VerknuepfungUnd && n.Op != VerknuepfungOder {
		return knotenTyp{}, formFehler("Operator der Verknüpfung")
	}
	return wahrheitTyp, nil
}

// eigenschaft schlägt den Namen bytegleich nach (SPEC-167 Nr. 3 a, 5). Die Maske sieht nur Namen von Properties
// mit unterstütztem Typ.
func (p filterPruefer) eigenschaft(name string) (knotenTyp, error) {
	for _, prop := range p.typ.Properties {
		if prop.Name != name {
			continue
		}
		if !prop.Unterstuetzt() {
			return knotenTyp{}, verstoss(FilterTypNichtUnterstuetzt)
		}
		if p.beobachtet != nil && !p.beobachtet(name) {
			return knotenTyp{}, verstoss(FilterNichtBeobachtet)
		}
		k, _ := klasseVonEdmTyp(prop.EdmType)
		return knotenTyp{klasse: k, edmTyp: prop.EdmType}, nil
	}
	if p.typ.Open {
		return knotenTyp{}, verstoss(FilterTypNichtUnterstuetzt)
	}
	for _, nav := range p.typ.NavigationProperties {
		if nav.Name == name {
			return knotenTyp{}, verstoss(FilterTypNichtUnterstuetzt)
		}
	}
	return knotenTyp{}, verstoss(FilterEigenschaftUnbekannt)
}

// literalTyp liefert die Klasse der Art des Literals (SPEC-167 Nr. 2); eine unbekannte Art ist ein Programmfehler.
func literalTyp(l Literal) (knotenTyp, error) {
	switch l.Art {
	case LiteralZeichenkette:
		return knotenTyp{klasse: klasseText}, nil
	case LiteralWahrheitswert:
		return knotenTyp{klasse: klasseWahrheit}, nil
	case LiteralZahl:
		return knotenTyp{klasse: klasseZahl, zahlText: l.Text, zahlLiteral: true}, nil
	case LiteralGUID:
		return knotenTyp{klasse: klasseGUID}, nil
	case LiteralDatum:
		return knotenTyp{klasse: klasseDatum}, nil
	case LiteralZeitpunkt:
		return knotenTyp{klasse: klasseZeitpunkt}, nil
	}
	return knotenTyp{}, formFehler("Art des Literals")
}

// vergleich prüft die Kinder und dann in dieser Folge: Verträglichkeit, Bereich eines Zahlliterals gegen eine
// Property des Typs Edm.Int32 oder Edm.Int64, Ordnungsoperator auf Guid (SPEC-167 Nr. 3 c).
func (p filterPruefer) vergleich(n FilterVergleich) (knotenTyp, error) {
	l, err := p.knoten(n.Links)
	if err != nil {
		return knotenTyp{}, err
	}
	r, err := p.knoten(n.Rechts)
	if err != nil {
		return knotenTyp{}, err
	}
	switch n.Op {
	case VergleichEq, VergleichNe, VergleichGt, VergleichGe, VergleichLt, VergleichLe:
	default:
		return knotenTyp{}, formFehler("Operator des Vergleichs")
	}
	if l.klasse != r.klasse && l.klasse != klasseNull && r.klasse != klasseNull {
		return knotenTyp{}, verstoss(FilterTypUnvertraeglich)
	}
	if !bereichGueltig(l, r) || !bereichGueltig(r, l) {
		return knotenTyp{}, verstoss(FilterLiteralBereich)
	}
	if (l.klasse == klasseGUID || r.klasse == klasseGUID) && n.Op != VergleichEq && n.Op != VergleichNe {
		return knotenTyp{}, verstoss(FilterOperatorNichtUnterstuetzt)
	}
	return wahrheitTyp, nil
}

// bereichGueltig prüft ein Zahlliteral gegen den Typ der Property auf der anderen Seite: bei Edm.Int32 und
// Edm.Int64 ganzzahlig und im Bereich des Typs, sonst ohne Grenze.
func bereichGueltig(literal, gegen knotenTyp) bool {
	if !literal.zahlLiteral {
		return true
	}
	var bits int
	switch gegen.edmTyp {
	case "Edm.Int32":
		bits = 32
	case "Edm.Int64":
		bits = 64
	default:
		return true
	}
	if _, _, _, ok := zerlegeZahl(literal.zahlText); !ok {
		return false
	}
	_, err := strconv.ParseInt(literal.zahlText, 10, bits)
	return err == nil
}

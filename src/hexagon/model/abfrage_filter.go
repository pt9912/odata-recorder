package model

// filterNamen hängt die Namen aller FilterEigenschaft des Baums an namen an, links vor rechts und ohne Kurzschluss
// (SPEC-191 Nr. 2 b).
func filterNamen(a FilterAusdruck, namen []string) []string {
	switch n := a.(type) {
	case FilterEigenschaft:
		return append(namen, n.Name)
	case FilterNicht:
		return filterNamen(n.Operand, namen)
	case FilterVerknuepfung:
		return filterNamen(n.Rechts, filterNamen(n.Links, namen))
	case FilterVergleich:
		return filterNamen(n.Rechts, filterNamen(n.Links, namen))
	}
	return namen
}

// filtere liefert die Entities, deren Wurzel durchlässt (SPEC-191 Nr. 2 c); ohne Filter die Eingabe selbst, die der Aufrufer
// nicht verändert.
func filtere(entities []Entity, f *FilterOption) ([]Entity, error) {
	if f == nil {
		return entities, nil
	}
	treffer := make([]Entity, 0, len(entities))
	for _, e := range entities {
		w, err := filterWert(f.Ausdruck, e)
		if err != nil {
			return nil, err
		}
		wahrheit, err := WahrheitVon(w)
		if err != nil {
			return nil, err
		}
		if wahrheit.Durchlaesst() {
			treffer = append(treffer, e)
		}
	}
	return treffer, nil
}

// filterWert wertet einen Knoten auf einer Entity aus. Ein Vergleich oder eine Verknüpfung ist der Wert `boolean`, unbekannt
// ist `null` (SPEC-191 Nr. 2 c); beide Operanden werden immer ausgewertet.
func filterWert(a FilterAusdruck, e Entity) (Wert, error) {
	switch n := a.(type) {
	case FilterEigenschaft:
		w, ok := wertVon(e, n.Name)
		if !ok {
			return Wert{}, formFehler("Property der Entity fehlt")
		}
		return w, nil
	case FilterLiteral:
		return WertAusLiteral(n.Literal)
	case FilterNull:
		return Wert{Art: ArtNull}, nil
	case FilterNicht:
		o, err := filterWert(n.Operand, e)
		if err != nil {
			return Wert{}, err
		}
		w, err := WahrheitVon(o)
		return wahrheitAlsWert(Nicht(w)), err
	case FilterVerknuepfung:
		return verknuepfeWerte(n, e)
	case FilterVergleich:
		l, r, err := operanden2(n.Links, n.Rechts, e)
		if err != nil {
			return Wert{}, err
		}
		w, err := Vergleiche(n.Op, l, r)
		return wahrheitAlsWert(w), err
	}
	return Wert{}, formFehler("Knoten des Filters")
}

// operanden2 wertet beide Operanden aus; der Fehler des linken geht vor.
func operanden2(links, rechts FilterAusdruck, e Entity) (l, r Wert, err error) {
	l, errL := filterWert(links, e)
	r, errR := filterWert(rechts, e)
	if errL != nil {
		return Wert{}, Wert{}, errL
	}
	return l, r, errR
}

// verknuepfeWerte wertet `and` und `or` über die Wahrheit der Operanden aus.
func verknuepfeWerte(n FilterVerknuepfung, e Entity) (Wert, error) {
	l, r, err := operanden2(n.Links, n.Rechts, e)
	if err != nil {
		return Wert{}, err
	}
	wl, err := WahrheitVon(l)
	if err != nil {
		return Wert{}, err
	}
	wr, err := WahrheitVon(r)
	if err != nil {
		return Wert{}, err
	}
	switch n.Op {
	case VerknuepfungUnd:
		return wahrheitAlsWert(Und(wl, wr)), nil
	case VerknuepfungOder:
		return wahrheitAlsWert(Oder(wl, wr)), nil
	}
	return Wert{}, formFehler("Operator der Verknüpfung")
}

// wahrheitAlsWert deutet wahr als `boolean` `true`, falsch als `false` und unbekannt als `null`.
func wahrheitAlsWert(w Wahrheit) Wert {
	switch w {
	case WahrheitWahr:
		return Wert{Art: ArtBoolean, Text: "true"}
	case WahrheitFalsch:
		return Wert{Art: ArtBoolean, Text: "false"}
	}
	return Wert{Art: ArtNull}
}

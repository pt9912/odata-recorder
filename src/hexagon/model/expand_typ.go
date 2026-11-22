package model

import "slices"

// expandUrsache liefert die Ursache des Verstoßes für einen Namen, "" für eine Navigation des Typs (SPEC-263 Nr. 2).
func expandUrsache(typ EntityType, name string) ExpandUrsache {
	switch {
	case slices.ContainsFunc(typ.NavigationProperties, func(n NavigationProperty) bool { return n.Name == name }):
		return ""
	case slices.ContainsFunc(typ.Properties, func(p Property) bool { return p.Name == name }):
		return ExpandKeineNavigation
	case typ.Open:
		return ExpandNavigationNichtUnterstuetzt
	}
	return ExpandNavigationUnbekannt
}

// PruefeExpand prüft die Namen von o gegen die wirksamen Navigation-Properties von typ (SPEC-263): der erste Verstoß von
// links entscheidet. Das Ergebnis ist nil, ein *ExpandFehler oder ErrUngueltig für einen Programmfehler des Aufrufers;
// typ und o bleiben unverändert, und die Funktion liest weder Modell noch Daten.
func PruefeExpand(typ EntityType, o *ExpandOption) error {
	if err := expandProgrammfehler(o); err != nil {
		return err
	}
	for _, e := range o.Elemente {
		if u := expandUrsache(typ, e.Name); u != "" {
			return &ExpandFehler{Ursache: u}
		}
	}
	return nil
}

// expandProgrammfehler stellt für den ganzen Knoten fest, was der Parser nie liefert: kein Knoten, keine Elemente, ein
// leerer Name, `*` und ein Name, der bytegleich zweimal steht (SPEC-263 Nr. 4).
func expandProgrammfehler(o *ExpandOption) error {
	if o == nil {
		return formFehler("Expand ist nil")
	}
	if len(o.Elemente) == 0 {
		return formFehler("Expand ohne Element")
	}
	gesehen := make(map[string]struct{}, len(o.Elemente))
	for _, e := range o.Elemente {
		switch e.Name {
		case "":
			return formFehler("Expand mit leerem Namen")
		case "*":
			return formFehler("Expand mit `*` als Namen")
		}
		if _, doppelt := gesehen[e.Name]; doppelt {
			return formFehler("Expand mit doppeltem Namen")
		}
		gesehen[e.Name] = struct{}{}
	}
	return nil
}

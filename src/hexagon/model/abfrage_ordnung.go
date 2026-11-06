package model

import (
	"slices"
	"strings"
)

// pruefeName prüft einen Namen von `$orderby` oder `$select` in der Folge der Ursachen von SPEC-190 Nr. 4; `ordnung_nicht_unterstuetzt`
// gibt es nur an `$orderby`. Die Maske sieht nur Namen von Properties mit unterstütztem Typ (und Ordnung).
func pruefeName(typ EntityType, name string, o AbfrageOption, beobachtet func(name string) bool) error {
	i := slices.IndexFunc(typ.Properties, func(p Property) bool { return p.Name == name })
	if i < 0 {
		if typ.Open || slices.ContainsFunc(typ.NavigationProperties, func(n NavigationProperty) bool { return n.Name == name }) {
			return abfrageVerstoss(o, AbfrageTypNichtUnterstuetzt)
		}
		return abfrageVerstoss(o, AbfrageEigenschaftUnbekannt)
	}
	p := typ.Properties[i]
	switch {
	case !p.Unterstuetzt():
		return abfrageVerstoss(o, AbfrageTypNichtUnterstuetzt)
	case o == AbfrageOrderBy && !HatOrdnung(p.EdmType):
		return abfrageVerstoss(o, AbfrageOrdnungNichtUnterstuetzt)
	case beobachtet != nil && !beobachtet(name):
		return abfrageVerstoss(o, AbfrageNichtBeobachtet)
	}
	return nil
}

// spalte ist ein Schlüssel der Ordnung: der Wert jeder Entity, die Richtung und ob die Werte Guids sind (nur im Schlüssel).
type spalte struct {
	werte      []Wert
	absteigend bool
	guid       bool
}

// ordne liefert eine geordnete Kopie von treffer (SPEC-191 Nr. 3): die Elemente von order, dann die Key-Properties
// aufsteigend, zuletzt der Text von Schluessel für Eingaben, deren Schlüssel die Werte nicht eindeutig machen. Die Eingabe
// bleibt unverändert.
func ordne(typ EntityType, treffer []Entity, order []Sortierung) ([]Entity, error) {
	spalten := make([]spalte, 0, len(order)+len(typ.Keys))
	for _, s := range order {
		sp, err := baueSpalte(treffer, s.Name, s.Absteigend, false)
		if err != nil {
			return nil, err
		}
		spalten = append(spalten, sp)
	}
	for _, k := range typ.Keys {
		sp, err := baueSpalte(treffer, k, false, true)
		if err != nil {
			return nil, err
		}
		spalten = append(spalten, sp)
	}
	idx := make([]int, len(treffer))
	for i := range idx {
		idx[i] = i
	}
	slices.SortFunc(idx, func(a, b int) int {
		for _, sp := range spalten {
			if c := sp.vergleiche(a, b); c != 0 {
				return c
			}
		}
		return strings.Compare(treffer[a].Schluessel, treffer[b].Schluessel)
	})
	geordnet := make([]Entity, len(treffer))
	for k, i := range idx {
		geordnet[k] = treffer[i]
	}
	return geordnet, nil
}

// baueSpalte liest den Wert der Property aus jeder Entity und prüft, dass OrdneWerte sie ordnen kann: gültige Werte einer
// einzigen Klasse neben `null`, ohne Guid. Im Schlüssel sind Guids erlaubt und werden nach dem Text geordnet (SPEC-191 Nr. 3 d).
func baueSpalte(treffer []Entity, name string, absteigend, schluessel bool) (spalte, error) {
	sp := spalte{werte: make([]Wert, len(treffer)), absteigend: absteigend}
	var klasse wertKlasse
	for i, e := range treffer {
		w, _ := wertVon(e, name)
		k, err := pruefeWert(w)
		if err != nil {
			return spalte{}, err
		}
		sp.werte[i] = w
		if k == klasseNull {
			continue
		}
		if klasse != 0 && k != klasse {
			return spalte{}, formFehler("Klassen nicht vergleichbar")
		}
		klasse = k
	}
	if klasse == klasseGUID {
		if !schluessel {
			return spalte{}, formFehler("Ordnung auf Guid")
		}
		sp.guid = true
	}
	return sp, nil
}

// vergleiche ordnet die Entities a und b nach dieser Spalte; die Werte sind in baueSpalte geprüft.
func (s spalte) vergleiche(a, b int) int {
	var c int
	if s.guid {
		c = strings.Compare(s.werte[a].Text, s.werte[b].Text)
	} else {
		c, _ = OrdneWerte(s.werte[a], s.werte[b])
	}
	if s.absteigend {
		return -c
	}
	return c
}

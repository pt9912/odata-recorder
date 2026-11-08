package model

// Ergebnis ist das Ergebnis von FuehreAus (SPEC-190 Nr. 2): Entities ist nie nil; Anzahl ist die Zahl nach dem Filter und
// vor `$skip` und `$top`. Bei einem Fehler ist es der Nullwert.
type Ergebnis struct {
	Entities []Entity
	Anzahl   int
}

// FuehreAus führt die Pipeline von `LH-FA-08.a` auf entities aus: Filter, Zählung, Ordnung, `$skip`, `$top`, Projektion
// (SPEC-190, SPEC-191). beobachtet ist die Property-Maske, nil heißt: alle beobachtet. Der Fehler ist ErrUngueltig für einen
// Programmfehler des Aufrufers, ein *FilterTypFehler oder ein *AbfrageFehler für einen Verstoß des Clients; die Eingabe bleibt
// unverändert, das Ergebnis gehört dem Aufrufer.
func FuehreAus(typ EntityType, entities []Entity, opt Optionen, beobachtet func(name string) bool) (Ergebnis, error) {
	if err := pruefeArgumente(typ, entities, opt); err != nil {
		return Ergebnis{}, err
	}
	if err := pruefeNamen(typ, opt, beobachtet); err != nil {
		return Ergebnis{}, err
	}
	if err := pruefeVorhandene(entities, opt); err != nil {
		return Ergebnis{}, err
	}
	treffer, err := filtere(entities, opt.Filter)
	if err != nil {
		return Ergebnis{}, err
	}
	geordnet, err := ordne(typ, treffer, opt.OrderBy)
	if err != nil {
		return Ergebnis{}, err
	}
	von, bis := seite(len(geordnet), opt.Skip, opt.Top)
	return Ergebnis{Entities: projiziere(geordnet[von:bis], opt.Select), Anzahl: len(geordnet)}, nil
}

// pruefeNamen ist Stufe 2 bis 4 der Rangfolge (SPEC-190 Nr. 5): `$filter`, `$orderby`, `$select`.
func pruefeNamen(typ EntityType, opt Optionen, beobachtet func(name string) bool) error {
	if opt.Filter != nil {
		if err := PruefeFilter(typ, opt.Filter.Ausdruck, beobachtet); err != nil {
			return err
		}
	}
	for _, s := range opt.OrderBy {
		if err := pruefeName(typ, s.Name, AbfrageOrderBy, beobachtet); err != nil {
			return err
		}
	}
	if opt.Select != nil {
		for _, n := range opt.Select.Namen {
			if err := pruefeName(typ, n, AbfrageSelect, beobachtet); err != nil {
				return err
			}
		}
	}
	return nil
}

// pruefeVorhandene ist Stufe 5: jede Property von `$filter` und dann von `$orderby` steht in jeder Entity (SPEC-191 Nr. 2 b, 3 c).
func pruefeVorhandene(entities []Entity, opt Optionen) error {
	if opt.Filter != nil {
		if err := verlangeProperties(entities, filterNamen(opt.Filter.Ausdruck, nil), AbfrageFilter); err != nil {
			return err
		}
	}
	namen := make([]string, 0, len(opt.OrderBy))
	for _, s := range opt.OrderBy {
		namen = append(namen, s.Name)
	}
	return verlangeProperties(entities, namen, AbfrageOrderBy)
}

// verlangeProperties meldet `nicht_beobachtet`, wenn eine Entity einen der Namen nicht nennt.
func verlangeProperties(entities []Entity, namen []string, o AbfrageOption) error {
	for _, e := range entities {
		for _, n := range namen {
			if _, ok := wertVon(e, n); !ok {
				return abfrageVerstoss(o, AbfrageNichtBeobachtet)
			}
		}
	}
	return nil
}

// wertVon liefert den ersten Eintrag der Entity mit dem Namen (SPEC-190 *Ungeprüft*); false, wenn die Entity ihn nicht nennt.
func wertVon(e Entity, name string) (Wert, bool) {
	for _, p := range e.Properties {
		if p.Name == name {
			return p.Wert, true
		}
	}
	return Wert{}, false
}

// pruefeArgumente ist Stufe 1: die Programmfehler des Aufrufers (SPEC-190 Nr. 3 a bis f).
func pruefeArgumente(typ EntityType, entities []Entity, opt Optionen) error {
	if (opt.Top.Gesetzt && opt.Top.Wert < 0) || (opt.Skip.Gesetzt && opt.Skip.Wert < 0) {
		return formFehler("negativer Wert von $top oder $skip")
	}
	if opt.Filter != nil && opt.Filter.Ausdruck == nil {
		return formFehler("Filter ohne Ausdruck")
	}
	if s := opt.Select; s != nil {
		if (!s.Alle && len(s.Namen) == 0) || hatDoppelte(s.Namen) {
			return formFehler("Select ohne Namen oder mit doppeltem Namen")
		}
	}
	namen := make([]string, 0, len(opt.OrderBy))
	for _, s := range opt.OrderBy {
		namen = append(namen, s.Name)
	}
	if hatDoppelte(namen) {
		return formFehler("doppelter Name in $orderby")
	}
	return pruefeEntities(typ, entities)
}

// hatDoppelte sagt, ob ein Name zweimal vorkommt.
func hatDoppelte(namen []string) bool {
	gesehen := make(map[string]struct{}, len(namen))
	for _, n := range namen {
		if _, ok := gesehen[n]; ok {
			return true
		}
		gesehen[n] = struct{}{}
	}
	return false
}

// pruefeEntities verlangt je Entity einen nichtleeren, eindeutigen Schlüssel und Key-Properties mit Wert (SPEC-190 Nr. 3 e, f).
func pruefeEntities(typ EntityType, entities []Entity) error {
	gesehen := make(map[string]struct{}, len(entities))
	for _, e := range entities {
		if _, doppelt := gesehen[e.Schluessel]; e.Schluessel == "" || doppelt {
			return formFehler("Schlüssel leer oder doppelt")
		}
		gesehen[e.Schluessel] = struct{}{}
		for _, k := range typ.Keys {
			if w, ok := wertVon(e, k); !ok || w.Art == ArtNull || w.Art == ArtRedigiert {
				return formFehler("Key-Property fehlt, ist null oder redigiert")
			}
		}
	}
	return nil
}

// seite liefert den Ausschnitt [von, bis) von n geordneten Entities: zuerst `$skip`, dann `$top` (SPEC-191 Nr. 4). Die Rechnung
// bildet nie die Summe der beiden Werte.
func seite(n int, skip, top ZahlOption) (von, bis int) {
	if skip.Gesetzt {
		von = min(skip.Wert, n)
	}
	bis = n
	if top.Gesetzt && top.Wert < n-von {
		bis = von + top.Wert
	}
	return von, bis
}

// projiziere liefert eine Kopie jeder Entity mit den Properties, die `$select` nennt (SPEC-191 Nr. 5); die Liste ist nie nil.
func projiziere(entities []Entity, sel *SelectOption) []Entity {
	aus := make([]Entity, 0, len(entities))
	var namen map[string]struct{}
	if sel != nil && !sel.Alle {
		namen = make(map[string]struct{}, len(sel.Namen))
		for _, n := range sel.Namen {
			namen[n] = struct{}{}
		}
	}
	for _, e := range entities {
		props := make([]PropertyWert, 0, len(e.Properties))
		for _, p := range e.Properties {
			if _, ok := namen[p.Name]; namen == nil || ok {
				props = append(props, p)
			}
		}
		aus = append(aus, Entity{Typ: e.Typ, Schluessel: e.Schluessel, Properties: props, Etag: e.Etag})
	}
	return aus
}

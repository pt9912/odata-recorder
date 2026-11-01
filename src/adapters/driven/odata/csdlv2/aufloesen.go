package csdlv2

import (
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// loese führt SPEC-171 Nr. 7 aus, in den Schritten a bis g, und baut Modell und Bericht.
func (l *lauf) loese() (model.ServiceModel, model.ValidierungsBericht, error) {
	d := &l.d
	for _, schritt := range []func() error{
		d.basen, d.ketten, d.summe, d.vererbung, d.typenAufloesen, d.containerAufloesen, d.schluessel, d.ohneSchluessel,
	} {
		if err := schritt(); err != nil {
			return model.ServiceModel{}, model.ValidierungsBericht{}, err
		}
	}
	return d.modell(), model.ValidierungsBericht{Funde: l.funde()}, nil
}

func nichtAufloesbar(ort string) error { return fehler(model.UrsacheNichtAufloesbar, ort) }

// qualify löst `<Namensraum oder Alias>.<Name>` in den qualifizierten Namen auf.
func (d *dokument) qualify(s string) (string, bool) {
	i := strings.LastIndex(s, ".")
	if i <= 0 || i == len(s)-1 {
		return "", false
	}
	ns, ok := d.abbild[s[:i]]
	return ns + "." + s[i+1:], ok
}

func (d *dokument) entityRef(s string) (*rawType, bool) {
	q, ok := d.qualify(s)
	t := d.index[q]
	return t, ok && t != nil
}

func (d *dokument) assocRef(s string) (*rawAssoc, bool) {
	q, ok := d.qualify(s)
	a := d.assocs[q]
	return a, ok && a != nil
}

// basen ist Nr. 9 a von SPEC-125, erster Durchgang: jeder BaseType löst auf einen Entity-Typ auf.
func (d *dokument) basen() error {
	for _, t := range d.typen {
		if !t.hasBase {
			continue
		}
		b, ok := d.entityRef(t.baseRaw)
		if !ok {
			return nichtAufloesbar(t.qn + "@BaseType")
		}
		t.base = b
	}
	return nil
}

// ketten ist Nr. 9 a, zweiter Durchgang: jede Kette endet und hat höchstens 64 Typen.
func (d *dokument) ketten() error {
	for _, t := range d.typen {
		gesehen := []*rawType{t}
		for c := t.base; c != nil; c = c.base {
			switch {
			case slices.Contains(gesehen, c):
				return nichtAufloesbar(t.qn + "@BaseType")
			case len(gesehen) >= maxKette:
				return fehler(model.UrsacheKette, t.qn)
			}
			gesehen = append(gesehen, c)
		}
	}
	return nil
}

// kette liefert t und seine Basistypen, die Wurzel zuerst.
func kette(t *rawType) []*rawType {
	var k []*rawType
	for c := t; c != nil; c = c.base {
		k = append(k, c)
	}
	slices.Reverse(k)
	return k
}

// summe ist Nr. 9 b: die Summe der wirksamen Members aller Typen, Navigation-Properties eingeschlossen.
func (d *dokument) summe() error {
	total := 0
	for _, t := range d.typen {
		for _, c := range kette(t) {
			total += len(c.members)
		}
		if total > maxMembers {
			return fehler(model.UrsacheZuVieleMembers, t.qn)
		}
	}
	return nil
}

// vererbung ist Nr. 9 c: ein Schlüssel je Kette und kein geerbter Name.
func (d *dokument) vererbung() error {
	for _, t := range d.typen {
		ahnen := kette(t)
		ahnen = ahnen[:len(ahnen)-1]
		namen := map[string]bool{}
		hatSchluessel := false
		for _, a := range ahnen {
			hatSchluessel = hatSchluessel || len(a.refs) > 0
			for _, m := range a.members {
				namen[m.name] = true
			}
		}
		if hatSchluessel && len(t.refs) > 0 {
			return struktur(t.qn)
		}
		for _, m := range t.members {
			if namen[m.name] {
				return struktur(t.qn + "/" + m.name)
			}
		}
	}
	return nil
}

// typenAufloesen ist Nr. 7 d: die Typen der Properties nach Nr. 5, die Assoziationen der Navigation-Properties nach Nr. 6.
func (d *dokument) typenAufloesen() error {
	for _, t := range d.typen {
		for _, m := range t.members {
			switch {
			case m.nav && !d.navigation(m):
				return nichtAufloesbar(t.qn + "/" + m.name + "@Relationship")
			case !m.nav && !d.property(m):
				return nichtAufloesbar(t.qn + "/" + m.name + "@Type")
			}
		}
	}
	return nil
}

// navigation löst die Relationship auf eine Association, deren Rollen FromRole und ToRole verschieden sind, und
// den Typ des End der Rolle ToRole auf einen Entity-Typ (Nr. 6).
func (d *dokument) navigation(m *member) bool {
	a, ok := d.assocRef(m.rel)
	if !ok || m.von == m.nach {
		return false
	}
	if _, hatVon := a.end(m.von); !hatVon {
		return false
	}
	nach, ok := a.end(m.nach)
	if !ok {
		return false
	}
	t, ok := d.entityRef(nach.typ)
	if !ok {
		return false
	}
	m.assoc, m.edm, m.collection = a, t.qn, nach.mult == "*"
	return true
}

// property löst den Typ einer Property nach Nr. 5 auf: ein bekannter Name und ein deklarierter Typ bleiben
// im Modell, `Collection(…)` und jeder andere Name sind nicht auflösbar.
func (d *dokument) property(m *member) bool {
	s := m.typ
	if strings.HasPrefix(s, "Edm.") {
		if !slices.Contains(bekannteEDMTypen(), s) {
			return false
		}
		m.edm, m.fundName = edmTypImModell(s), s
		return true
	}
	q, ok := d.qualify(s)
	if !ok || (d.index[q] == nil && !d.deklariert[q]) {
		return false
	}
	m.edm, m.fundName = q, q
	return true
}

// containerAufloesen ist Nr. 7 e: die Kinder des Containers in ihrer Reihenfolge.
func (d *dokument) containerAufloesen() error {
	c := d.container
	if c == nil {
		return nil
	}
	for _, k := range c.reihe {
		if k.aset != nil {
			if err := d.assocSetAufloesen(k.aset); err != nil {
				return err
			}
			continue
		}
		t, ok := d.entityRef(k.set.typ)
		if !ok {
			return nichtAufloesbar(k.set.ort + "@EntityType")
		}
		k.set.typ = t.qn
	}
	return nil
}

// assocSetAufloesen ist Nr. 10 f: die Association, die EntitySet beider End und, sobald eine Navigation auf die
// Association verweist, ihre Rollen.
func (d *dokument) assocSetAufloesen(s *rawAssocSet) error {
	a, ok := d.assocRef(s.assoc)
	if !ok {
		return nichtAufloesbar(s.ort)
	}
	for _, e := range s.ends {
		if !d.container.setNamen[e.set] {
			return nichtAufloesbar(s.ort)
		}
	}
	for _, m := range d.navigationen(a) {
		_, hatVon := s.set(m.von)
		_, hatNach := s.set(m.nach)
		if !hatVon || !hatNach {
			return nichtAufloesbar(s.ort)
		}
	}
	s.ziel = a
	return nil
}

// navigationen liefert die Navigation-Properties aller Typen, die auf a verweisen, in der Reihenfolge des Dokuments.
func (d *dokument) navigationen(a *rawAssoc) []*member {
	var out []*member
	for _, t := range d.typen {
		for _, m := range t.members {
			if m.nav && m.assoc == a {
				out = append(out, m)
			}
		}
	}
	return out
}

// wirksameProperties liefert die Properties von t und seiner Basistypen nach Namen.
func wirksameProperties(t *rawType) map[string]*member {
	m := map[string]*member{}
	for _, c := range kette(t) {
		for _, p := range c.members {
			if !p.nav {
				m[p.name] = p
			}
		}
	}
	return m
}

func unterstuetzt(edm string) bool { return slices.Contains(model.UnterstuetzteEDMTypen(), edm) }

// schluessel ist Nr. 9 f: jede PropertyRef nennt eine Property mit unterstütztem Typ.
func (d *dokument) schluessel() error {
	for _, t := range d.typen {
		if len(t.refs) == 0 {
			continue
		}
		props := wirksameProperties(t)
		for _, ref := range t.refs {
			ort := t.qn + "/" + ref
			p, ok := props[ref]
			switch {
			case strings.Contains(ref, "/"):
				return fehler(model.UrsacheSchluesseltyp, ort)
			case !ok:
				return nichtAufloesbar(ort + "@Name")
			case !unterstuetzt(p.edm):
				return fehler(model.UrsacheSchluesseltyp, ort)
			}
		}
	}
	return nil
}

// ohneSchluessel ist Nr. 9 g.
func (d *dokument) ohneSchluessel() error {
	for _, t := range d.typen {
		if t.abstract {
			continue
		}
		if !slices.ContainsFunc(kette(t), func(c *rawType) bool { return len(c.refs) > 0 }) {
			return fehler(model.UrsacheOhneSchluessel, t.qn)
		}
	}
	return nil
}

func (d *dokument) modell() model.ServiceModel {
	var typen []model.EntityType
	for _, t := range d.typen {
		typen = append(typen, modellTyp(t))
	}
	var sets []model.EntitySet
	var bindings []model.NavigationBinding
	if c := d.container; c != nil {
		for _, k := range c.reihe {
			if k.set != nil {
				sets = append(sets, model.EntitySet{Name: k.set.name, EntityType: k.set.typ})
				continue
			}
			bindings = append(bindings, d.bindings(k.aset)...)
		}
	}
	return model.NeuesServiceModel("2.0", d.namespaces, typen, sets, bindings)
}

// bindings bildet zu einer AssociationSet und jeder Navigation-Property ihrer Association ein Binding (Nr. 6):
// Set aus dem End der Rolle FromRole, Pfad der Name der Navigation, Ziel aus dem End der Rolle ToRole.
func (d *dokument) bindings(s *rawAssocSet) []model.NavigationBinding {
	var out []model.NavigationBinding
	for _, m := range d.navigationen(s.ziel) {
		von, _ := s.set(m.von)
		nach, _ := s.set(m.nach)
		out = append(out, model.NavigationBinding{Set: von, Path: m.name, Target: d.container.qn + "/" + nach})
	}
	return out
}

func modellTyp(t *rawType) model.EntityType {
	et := model.EntityType{Name: t.qn, Abstract: t.abstract}
	if t.base != nil {
		et.BaseType = t.base.qn
	}
	for _, c := range kette(t) {
		if len(c.refs) > 0 {
			et.Keys = c.refs
		}
		for _, m := range c.members {
			if m.nav {
				et.NavigationProperties = append(et.NavigationProperties, model.NavigationProperty{Name: m.name, Target: m.edm, Collection: m.collection})
			} else {
				et.Properties = append(et.Properties, model.Property{Name: m.name, EdmType: m.edm, Nullable: m.nullable})
			}
		}
	}
	return et
}

// funde baut den Bericht: die Stelle einer Property trägt einen Fund, wenn ihr Typ nicht unterstützt ist.
func (l *lauf) funde() []model.Fund {
	funde := []model.Fund{}
	for _, s := range l.slots {
		switch {
		case s.mem == nil:
			funde = append(funde, s.fund)
		case !unterstuetzt(s.mem.edm):
			s.fund.Name = s.mem.fundName
			funde = append(funde, s.fund)
		}
	}
	return funde
}

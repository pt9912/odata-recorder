package csdl

import (
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const collectionPrefix = "Collection("

// loese führt Nr. 9 aus, in den Schritten a bis g, und baut Modell und Bericht.
func (l *lauf) loese() (model.ServiceModel, model.ValidierungsBericht, error) {
	d := &l.d
	for _, schritt := range []func() error{
		d.basen, d.ketten, d.summe, d.vererbung, d.typenAufloesen, d.setsAufloesen, d.schluessel, d.ohneSchluessel,
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

// basen ist Nr. 9 a, erster Durchgang: jeder BaseType löst auf einen Entity-Typ auf.
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

// summe ist Nr. 9 b: die Summe der wirksamen Members aller Typen.
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

// typenAufloesen ist Nr. 9 d: die Typen der eigenen Members jedes Typs.
func (d *dokument) typenAufloesen() error {
	for _, t := range d.typen {
		for _, m := range t.members {
			ok := false
			if m.nav {
				ok = d.navTyp(m)
			} else {
				m.edm, ok = d.propTyp(m.typ)
			}
			if !ok {
				return nichtAufloesbar(t.qn + "/" + m.name + "@Type")
			}
		}
	}
	return nil
}

func (d *dokument) navTyp(m *member) bool {
	s := m.typ
	if strings.HasPrefix(s, collectionPrefix) && strings.HasSuffix(s, ")") {
		s, m.collection = s[len(collectionPrefix):len(s)-1], true
	}
	t, ok := d.entityRef(s)
	if ok {
		m.edm = t.qn
	}
	return ok
}

// propTyp löst den Typ einer Property nach Nr. 10 auf.
func (d *dokument) propTyp(s string) (string, bool) {
	if strings.HasPrefix(s, collectionPrefix) && strings.HasSuffix(s, ")") {
		inner := s[len(collectionPrefix) : len(s)-1]
		if inner == "" || strings.HasPrefix(inner, collectionPrefix) {
			return "", false
		}
		q, ok := d.einfacherTyp(inner)
		return collectionPrefix + q + ")", ok
	}
	return d.einfacherTyp(s)
}

func (d *dokument) einfacherTyp(s string) (string, bool) {
	if strings.HasPrefix(s, "Edm.") {
		return s, slices.Contains(bekannteEDMTypen(), s)
	}
	q, ok := d.qualify(s)
	return q, ok && (d.index[q] != nil || d.deklariert[q])
}

// setsAufloesen ist Nr. 9 e: Entity-Sets und Bindings.
func (d *dokument) setsAufloesen() error {
	c := d.container
	if c == nil {
		return nil
	}
	for _, s := range c.sets {
		t, ok := d.entityRef(s.typ)
		if !ok {
			return nichtAufloesbar(s.ort + "@EntityType")
		}
		s.typ = t.qn
		for i, b := range s.bindings {
			z, ok := d.bindingZiel(b.target)
			if !ok {
				return nichtAufloesbar(s.ort + "@Target")
			}
			s.bindings[i].target = z
		}
	}
	return nil
}

// bindingZiel löst den Namen eines Sets oder Singletons des Containers auf, auch als `<Container>/<Name>`;
// das Ergebnis ist immer `<qualifizierter Container>/<Name>` (SPEC-125 Nr. 11).
func (d *dokument) bindingZiel(raw string) (string, bool) {
	c := d.container
	name := raw
	if i := strings.LastIndex(raw, "/"); i >= 0 {
		q, ok := d.qualify(raw[:i])
		if !ok || q != c.qn {
			return "", false
		}
		name = raw[i+1:]
	}
	return c.qn + "/" + name, c.setNamen[name] || c.singletons[name]
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
	if d.container != nil {
		for _, s := range d.container.sets {
			sets = append(sets, model.EntitySet{Name: s.name, EntityType: s.typ})
			for _, b := range s.bindings {
				bindings = append(bindings, model.NavigationBinding{Set: s.name, Path: b.path, Target: b.target})
			}
		}
	}
	return model.NeuesServiceModel("4.0", d.namespaces, typen, sets, bindings)
}

func modellTyp(t *rawType) model.EntityType {
	et := model.EntityType{Name: t.qn, Abstract: t.abstract}
	if t.base != nil {
		et.BaseType = t.base.qn
	}
	for _, c := range kette(t) {
		et.Open = et.Open || c.open
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
			s.fund.Name = s.mem.edm
			funde = append(funde, s.fund)
		}
	}
	return funde
}

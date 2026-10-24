package csdl

import (
	"encoding/xml"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ausgewertet liefert die Art des Kindelements, wenn der Leser es an dieser Stelle auswertet (Nr. 7).
func ausgewertet(eltern art, n xml.Name) (art, bool) {
	type paar struct {
		eltern art
		name   string
	}
	switch {
	case eltern == kEdmx && n.Space == nsEdmx && n.Local == "DataServices":
		return kDataServices, true
	case n.Space != nsEdm:
		return 0, false
	}
	switch (paar{eltern, n.Local}) {
	case paar{kDataServices, "Schema"}:
		return kSchema, true
	case paar{kSchema, "EntityType"}:
		return kEntityType, true
	case paar{kSchema, "EntityContainer"}:
		return kContainer, true
	case paar{kSchema, "Annotations"}:
		return kAnnotations, true
	case paar{kEntityType, "Key"}:
		return kKey, true
	case paar{kEntityType, "Property"}:
		return kProperty, true
	case paar{kEntityType, "NavigationProperty"}:
		return kNavProp, true
	case paar{kKey, "PropertyRef"}:
		return kPropertyRef, true
	case paar{kContainer, "EntitySet"}:
		return kSet, true
	case paar{kSet, "NavigationPropertyBinding"}:
		return kBinding, true
	}
	return 0, false
}

func (l *lauf) kind(p rahmen, se xml.StartElement) error {
	n := se.Name
	if a, ok := ausgewertet(p.art, n); ok {
		return l.werte(a, p, se)
	}
	switch {
	case n.Space == nsEdm && n.Local == "Annotation":
		return l.annotation(p, se)
	case n.Space == nsEdm || n.Space == nsEdmx:
		return l.konstrukt(p, se)
	}
	return l.erweiterung(p, se)
}

func (l *lauf) werte(a art, p rahmen, se xml.StartElement) error {
	switch a {
	case kDataServices:
		return l.dataServices(se)
	case kSchema:
		return l.schema(se)
	case kEntityType:
		return l.entityType(p, se)
	case kKey:
		return l.key(p, se)
	case kPropertyRef:
		return l.propertyRef(p, se)
	case kProperty, kNavProp:
		return l.member(p, se, a == kNavProp)
	case kContainer:
		return l.container(p, se)
	case kSet:
		return l.set(p, se)
	case kBinding:
		return l.binding(p, se)
	}
	return l.annotations(se)
}

func (l *lauf) annotations(se xml.StartElement) error {
	target, _ := attr(se, "Target")
	l.push(rahmen{art: kAnnotations, ort: target})
	l.attributFunde(se, target, nil)
	return nil
}

func (l *lauf) fund(a model.FundArt, ort, name string) {
	l.slots = append(l.slots, slot{fund: model.Fund{Art: a, Ort: ort, Name: name}})
}

// attributFunde meldet die Attribute mit Namensraum (außer xmlns und xml) und die genannten
// Attribute ohne Namensraum mit dem Wert `true` (Nr. 7).
func (l *lauf) attributFunde(se xml.StartElement, ort string, konstrukt []string) {
	for _, a := range se.Attr {
		switch {
		case a.Name.Space == "":
			if a.Value == "true" && slices.Contains(konstrukt, a.Name.Local) {
				l.fund(model.FundKonstrukt, ort, a.Name.Local)
			}
		case a.Name.Space != "xmlns" && a.Name.Space != nsXML:
			l.fund(model.FundErweiterung, ort, "{"+a.Name.Space+"}"+a.Name.Local)
		}
	}
}

// ortKind ist der Ort eines Kindes nach Nr. 7: mit Name unter Schema und Container, sonst der des Elternelements.
func ortKind(p rahmen, se xml.StartElement) string {
	n, _ := attr(se, "Name")
	switch {
	case n == "":
		return p.ort
	case p.art == kSchema:
		return p.ort + "." + n
	case p.art == kContainer:
		return p.ort + "/" + n
	}
	return p.ort
}

func (l *lauf) konstrukt(p rahmen, se xml.StartElement) error {
	ort := ortKind(p, se)
	l.fund(model.FundKonstrukt, ort, se.Name.Local)
	l.attributFunde(se, ort, nil)
	if n, _ := attr(se, "Name"); n != "" && se.Name.Space == nsEdm {
		switch {
		case p.art == kSchema && slices.Contains([]string{"ComplexType", "EnumType", "TypeDefinition"}, se.Name.Local):
			l.d.deklariert[p.ort+"."+n] = true
		case p.art == kContainer && se.Name.Local == "Singleton":
			l.d.container.singletons[n] = true
		}
	}
	l.push(rahmen{art: kOpak})
	return nil
}

func (l *lauf) annotation(p rahmen, se xml.StartElement) error {
	term, _ := attr(se, "Term")
	l.fund(model.FundAnnotation, p.ort, term)
	l.attributFunde(se, p.ort, nil)
	l.push(rahmen{art: kOpak})
	return nil
}

func (l *lauf) erweiterung(p rahmen, se xml.StartElement) error {
	l.fund(model.FundErweiterung, ortKind(p, se), "{"+se.Name.Space+"}"+se.Name.Local)
	l.push(rahmen{art: kOpak})
	return nil
}

func (l *lauf) dataServices(se xml.StartElement) error {
	l.dataServ++
	if l.dataServ > 1 {
		return fehler(model.UrsacheWurzel, "")
	}
	l.push(rahmen{art: kDataServices})
	l.attributFunde(se, "", nil)
	return nil
}

// nameAttr liest und prüft `Name` (Nr. 8 a, b); der Ort eines Verstoßes ist der des Elternelements.
func nameAttr(se xml.StartElement, eltern string) (string, error) {
	n, ok := attr(se, "Name")
	if !ok || !einfacherName(n) {
		return "", struktur(eltern + "@Name")
	}
	return n, nil
}

// boolAttr prüft einen Wahrheitswert (Nr. 8 c); gesetzt sagt, ob das Attribut da ist.
func boolAttr(se xml.StartElement, ort, name string) (wert, gesetzt bool, err error) {
	v, ok := attr(se, name)
	switch {
	case !ok:
		return false, false, nil
	case v == "true":
		return true, true, nil
	case v == "false":
		return false, true, nil
	}
	return false, true, struktur(ort + "@" + name)
}

func (l *lauf) schema(se xml.StartElement) error {
	ns, _ := attr(se, "Namespace")
	if !namensraumGueltig(ns) {
		return struktur("@Namespace")
	}
	if ns == "Edm" {
		return struktur(ns + "@Namespace")
	}
	bekannt := l.d.abbild[ns] == ns
	if err := l.d.zuordnen(ns, ns, "@Namespace", ns); err != nil {
		return err
	}
	if al, ok := attr(se, "Alias"); ok {
		if !einfacherName(al) || al == "Edm" {
			return struktur(ns + "@Alias")
		}
		if err := l.d.zuordnen(al, ns, "@Alias", ns); err != nil {
			return err
		}
	}
	if !bekannt {
		l.d.namespaces = append(l.d.namespaces, ns)
	}
	l.push(rahmen{art: kSchema, ort: ns})
	l.attributFunde(se, ns, nil)
	return nil
}

// zuordnen trägt name → namensraum ein (Nr. 8 d); ein Widerspruch ist ein Verstoß am Attribut.
func (d *dokument) zuordnen(name, namensraum, attribut, ns string) error {
	if alt, ok := d.abbild[name]; ok && alt != namensraum {
		return struktur(ns + attribut)
	}
	d.abbild[name] = namensraum
	return nil
}

func (l *lauf) entityType(p rahmen, se xml.StartElement) error {
	name, err := nameAttr(se, p.ort)
	if err != nil {
		return err
	}
	qn := p.ort + "." + name
	var b [3]bool
	for i, a := range []string{"Abstract", "OpenType", "HasStream"} {
		if b[i], _, err = boolAttr(se, qn, a); err != nil {
			return err
		}
	}
	if l.d.index[qn] != nil {
		return struktur(qn)
	}
	t := &rawType{namespace: p.ort, name: name, qn: qn, abstract: b[0], open: b[1], namen: map[string]bool{}}
	t.baseRaw, t.hasBase = attr(se, "BaseType")
	l.d.typen = append(l.d.typen, t)
	l.d.index[qn] = t
	l.attributFunde(se, qn, []string{"OpenType", "HasStream"})
	l.push(rahmen{art: kEntityType, ort: qn, typ: t})
	return nil
}

func (l *lauf) key(p rahmen, se xml.StartElement) error {
	t := p.typ
	t.keys++
	if t.keys > 1 {
		return struktur(t.qn)
	}
	l.push(rahmen{art: kKey, ort: t.qn, typ: t})
	l.attributFunde(se, t.qn, nil)
	return nil
}

func (l *lauf) propertyRef(p rahmen, se xml.StartElement) error {
	t := p.typ
	n, ok := attr(se, "Name")
	if !ok || n == "" {
		return struktur(t.qn + "@Name")
	}
	if t.refSet[n] {
		return struktur(t.qn)
	}
	if t.refSet == nil {
		t.refSet = map[string]bool{}
	}
	t.refSet[n] = true
	t.refs = append(t.refs, n)
	l.push(rahmen{art: kPropertyRef, ort: t.qn, typ: t})
	l.attributFunde(se, t.qn, nil)
	return nil
}

func (l *lauf) member(p rahmen, se xml.StartElement, nav bool) error {
	t := p.typ
	name, err := nameAttr(se, t.qn)
	if err != nil {
		return err
	}
	ort := t.qn + "/" + name
	typ, ok := attr(se, "Type")
	if !ok || typ == "" {
		return struktur(ort + "@Type")
	}
	m := &member{nav: nav, name: name, typ: typ, nullable: true}
	a, konstrukt := kProperty, []string(nil)
	if nav {
		a, konstrukt = kNavProp, []string{"ContainsTarget"}
		_, _, err = boolAttr(se, ort, "ContainsTarget")
	} else {
		var v, gesetzt bool
		v, gesetzt, err = boolAttr(se, ort, "Nullable")
		m.nullable = !gesetzt || v
	}
	switch {
	case err != nil:
		return err
	case t.namen[name]:
		return struktur(ort)
	}
	t.namen[name] = true
	t.members = append(t.members, m)
	if !nav {
		l.slots = append(l.slots, slot{fund: model.Fund{Art: model.FundPropertyTyp, Ort: ort}, mem: m})
	}
	l.attributFunde(se, ort, konstrukt)
	l.push(rahmen{art: a, ort: ort, typ: t})
	return nil
}

func (l *lauf) container(p rahmen, se xml.StartElement) error {
	name, err := nameAttr(se, p.ort)
	if err != nil {
		return err
	}
	qn := p.ort + "." + name
	if l.d.container != nil {
		return struktur(qn)
	}
	l.d.container = &rawContainer{qn: qn, singletons: map[string]bool{}, setNamen: map[string]bool{}}
	l.attributFunde(se, qn, nil)
	l.push(rahmen{art: kContainer, ort: qn})
	return nil
}

func (l *lauf) set(p rahmen, se xml.StartElement) error {
	name, err := nameAttr(se, p.ort)
	if err != nil {
		return err
	}
	ort := p.ort + "/" + name
	typ, ok := attr(se, "EntityType")
	if !ok || typ == "" {
		return struktur(ort + "@EntityType")
	}
	if l.d.container.setNamen[name] {
		return struktur(ort)
	}
	l.d.container.setNamen[name] = true
	s := &rawSet{name: name, typ: typ, ort: ort}
	l.d.container.sets = append(l.d.container.sets, s)
	l.attributFunde(se, ort, nil)
	l.push(rahmen{art: kSet, ort: ort, set: s})
	return nil
}

func (l *lauf) binding(p rahmen, se xml.StartElement) error {
	pfad, ok := attr(se, "Path")
	if !ok || pfad == "" {
		return struktur(p.ort + "@Path")
	}
	ziel, ok := attr(se, "Target")
	if !ok || ziel == "" {
		return struktur(p.ort + "@Target")
	}
	p.set.bindings = append(p.set.bindings, rawBinding{path: pfad, target: ziel})
	l.attributFunde(se, p.ort, nil)
	l.push(rahmen{art: kBinding, ort: p.ort, set: p.set})
	return nil
}

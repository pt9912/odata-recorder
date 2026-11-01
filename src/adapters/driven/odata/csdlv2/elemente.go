package csdlv2

import (
	"encoding/xml"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ausgewertet liefert die Art des Kindelements, wenn der Leser es an dieser Stelle auswertet (SPEC-171 Nr. 4).
func ausgewertet(eltern art, n xml.Name) (art, bool) {
	switch {
	case eltern == kEdmx && n.Space == nsEdmx && n.Local == "DataServices":
		return kDataServices, true
	case !istEdm(n.Space):
		return 0, false
	}
	a, ok := kindArten(eltern)[n.Local]
	return a, ok
}

// kindArten sind die Kindelemente im Namensraum edm, die der Leser unter eltern auswertet, nach lokalem Namen.
func kindArten(eltern art) map[string]art {
	switch eltern {
	case kDataServices:
		return map[string]art{"Schema": kSchema}
	case kSchema:
		return map[string]art{"EntityType": kEntityType, "Association": kAssoc, "EntityContainer": kContainer}
	case kEntityType:
		return map[string]art{"Key": kKey, "Property": kProperty, "NavigationProperty": kNavProp}
	case kKey:
		return map[string]art{"PropertyRef": kPropertyRef}
	case kAssoc:
		return map[string]art{"End": kAssocEnd}
	case kContainer:
		return map[string]art{"EntitySet": kSet, "AssociationSet": kAssocSet}
	case kAssocSet:
		return map[string]art{"End": kAssocSetEnd}
	}
	return nil
}

func (l *lauf) kind(p rahmen, se xml.StartElement) error {
	n := se.Name
	if p.art == kDataServices && n.Local == "Schema" && istSpaeteresEdm(n.Space) {
		return fehler(model.UrsacheVersion, "")
	}
	if a, ok := ausgewertet(p.art, n); ok {
		return l.werte(a, p, se)
	}
	if istEdm(n.Space) || n.Space == nsEdmx {
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
	case kAssoc:
		return l.association(p, se)
	case kAssocEnd:
		return l.assocEnd(p, se)
	case kContainer:
		return l.container(p, se)
	case kSet:
		return l.set(p, se)
	case kAssocSet:
		return l.assocSet(p, se)
	}
	return l.assocSetEnd(p, se)
}

func (l *lauf) fund(a model.FundArt, ort, name string) {
	l.slots = append(l.slots, slot{fund: model.Fund{Art: a, Ort: ort, Name: name}})
}

// attributFunde meldet die Attribute mit Namensraum außer xmlns, xml und den drei Attributen ohne Wirkung (Nr. 4).
func (l *lauf) attributFunde(se xml.StartElement, ort string) {
	for _, a := range se.Attr {
		switch {
		case a.Name.Space == "", a.Name.Space == "xmlns", a.Name.Space == nsXML:
		case a.Name.Space == nsMeta && metaAttributOhneWirkung(a.Name.Local):
		default:
			l.fund(model.FundErweiterung, ort, "{"+a.Name.Space+"}"+a.Name.Local)
		}
	}
}

// ortKind ist der Ort eines Kindes nach Nr. 4: mit Name unter Schema und Container, sonst der des Elternelements.
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
	l.attributFunde(se, ort)
	if n, _ := attr(se, "Name"); n != "" && p.art == kSchema && se.Name.Local == "ComplexType" && istEdm(se.Name.Space) {
		l.d.deklariert[p.ort+"."+n] = true
	}
	l.push(rahmen{art: kOpak})
	return nil
}

func (l *lauf) erweiterung(p rahmen, se xml.StartElement) error {
	l.fund(model.FundErweiterung, ortKind(p, se), "{"+se.Name.Space+"}"+se.Name.Local)
	l.push(rahmen{art: kOpak})
	return nil
}

// dataServices prüft Nr. 2 c: genau ein DataServices, DataServiceVersion fehlend, 1.0 oder 2.0.
func (l *lauf) dataServices(se xml.StartElement) error {
	l.dataServ++
	if l.dataServ > 1 {
		return struktur("")
	}
	for _, a := range se.Attr {
		if a.Name.Space == nsMeta && a.Name.Local == "DataServiceVersion" && a.Value != "1.0" && a.Value != "2.0" {
			return fehler(model.UrsacheVersion, "")
		}
	}
	l.push(rahmen{art: kDataServices})
	l.attributFunde(se, "")
	return nil
}

// nameAttr liest und prüft `Name` (Nr. 10 a, SPEC-125 Nr. 8 b); der Ort eines Verstoßes ist der des Elternelements.
func nameAttr(se xml.StartElement, eltern string) (string, error) {
	n, ok := attr(se, "Name")
	if !ok || !einfacherName(n) {
		return "", struktur(eltern + "@Name")
	}
	return n, nil
}

// pflicht liest ein Pflichtattribut (Nr. 10 a): fehlend oder leer ist ein Verstoß am Attribut.
func pflicht(se xml.StartElement, ort, name string) (string, error) {
	v, ok := attr(se, name)
	if !ok || v == "" {
		return "", struktur(ort + "@" + name)
	}
	return v, nil
}

// boolAttr prüft einen Wahrheitswert (SPEC-125 Nr. 8 c); gesetzt sagt, ob das Attribut da ist.
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
	if ns == "Edm" || strings.HasPrefix(ns, "Edm.") {
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
	l.attributFunde(se, ns)
	return nil
}

// zuordnen trägt name → namensraum ein (SPEC-125 Nr. 8 d); ein Widerspruch ist ein Verstoß am Attribut.
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
	abstract, _, err := boolAttr(se, qn, "Abstract")
	if err != nil {
		return err
	}
	if l.d.index[qn] != nil {
		return struktur(qn)
	}
	t := &rawType{namespace: p.ort, name: name, qn: qn, abstract: abstract, namen: map[string]bool{}}
	t.baseRaw, t.hasBase = attr(se, "BaseType")
	l.d.typen = append(l.d.typen, t)
	l.d.index[qn] = t
	l.attributFunde(se, qn)
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
	l.attributFunde(se, t.qn)
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
	l.attributFunde(se, t.qn)
	return nil
}

// member liest eine Property (Name, Type, Nullable) oder Navigation-Property (Name, Relationship, FromRole, ToRole).
func (l *lauf) member(p rahmen, se xml.StartElement, nav bool) error {
	t := p.typ
	name, err := nameAttr(se, t.qn)
	if err != nil {
		return err
	}
	ort := t.qn + "/" + name
	m := &member{nav: nav, name: name, nullable: true}
	a := kProperty
	if nav {
		a = kNavProp
		err = m.leseNavigation(se, ort)
	} else {
		err = m.leseProperty(se, ort)
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
	l.attributFunde(se, ort)
	l.push(rahmen{art: a, ort: ort, typ: t})
	return nil
}

func (m *member) leseProperty(se xml.StartElement, ort string) (err error) {
	if m.typ, err = pflicht(se, ort, "Type"); err != nil {
		return err
	}
	v, gesetzt, err := boolAttr(se, ort, "Nullable")
	m.nullable = !gesetzt || v
	return err
}

func (m *member) leseNavigation(se xml.StartElement, ort string) (err error) {
	if m.rel, err = pflicht(se, ort, "Relationship"); err != nil {
		return err
	}
	if m.von, err = pflicht(se, ort, "FromRole"); err != nil {
		return err
	}
	m.nach, err = pflicht(se, ort, "ToRole")
	return err
}

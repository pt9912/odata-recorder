package csdlv2

import (
	"encoding/xml"
	"slices"
)

// multiplizitaeten sind die Werte von `Multiplicity` (SPEC-171 Nr. 6, 10 d).
func multiplizitaeten() []string { return []string{"0..1", "1", "*"} }

// association liest eine Association: Name, doppelt im Namensraum (Nr. 10 e).
func (l *lauf) association(p rahmen, se xml.StartElement) error {
	name, err := nameAttr(se, p.ort)
	if err != nil {
		return err
	}
	qn := p.ort + "." + name
	if l.d.assocs[qn] != nil {
		return struktur(qn)
	}
	a := &rawAssoc{qn: qn}
	l.d.assocs[qn] = a
	l.attributFunde(se, qn)
	l.push(rahmen{art: kAssoc, ort: qn, assoc: a})
	return nil
}

// assocEnd liest ein End der Association: höchstens zwei, verschiedene Rollen, Multiplicity (Nr. 10 b, d).
func (l *lauf) assocEnd(p rahmen, se xml.StartElement) error {
	a := p.assoc
	if len(a.ends) >= 2 {
		return struktur(a.qn)
	}
	var e rawEnd
	var err error
	if e.typ, err = pflicht(se, a.qn, "Type"); err != nil {
		return err
	}
	if e.role, err = pflicht(se, a.qn, "Role"); err != nil {
		return err
	}
	if e.mult, err = pflicht(se, a.qn, "Multiplicity"); err != nil {
		return err
	}
	if !slices.Contains(multiplizitaeten(), e.mult) {
		return struktur(a.qn + "@Multiplicity")
	}
	if len(a.ends) == 1 && a.ends[0].role == e.role {
		return struktur(a.qn + "@Role")
	}
	a.ends = append(a.ends, e)
	l.attributFunde(se, a.qn)
	l.push(rahmen{art: kAssocEnd, ort: a.qn, assoc: a})
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
	l.d.container = &rawContainer{qn: qn, setNamen: map[string]bool{}, asetNamen: map[string]bool{}}
	l.attributFunde(se, qn)
	l.push(rahmen{art: kContainer, ort: qn})
	return nil
}

func (l *lauf) set(p rahmen, se xml.StartElement) error {
	name, err := nameAttr(se, p.ort)
	if err != nil {
		return err
	}
	ort := p.ort + "/" + name
	typ, err := pflicht(se, ort, "EntityType")
	if err != nil {
		return err
	}
	if l.d.container.setNamen[name] {
		return struktur(ort)
	}
	l.d.container.setNamen[name] = true
	s := &rawSet{name: name, typ: typ, ort: ort}
	l.d.container.reihe = append(l.d.container.reihe, kindRef{set: s})
	l.attributFunde(se, ort)
	l.push(rahmen{art: kSet, ort: ort})
	return nil
}

// assocSet liest eine AssociationSet: Name, Association, doppelt im Container (Nr. 10 e).
func (l *lauf) assocSet(p rahmen, se xml.StartElement) error {
	name, err := nameAttr(se, p.ort)
	if err != nil {
		return err
	}
	ort := p.ort + "/" + name
	assoc, err := pflicht(se, ort, "Association")
	if err != nil {
		return err
	}
	if l.d.container.asetNamen[name] {
		return struktur(ort)
	}
	l.d.container.asetNamen[name] = true
	s := &rawAssocSet{name: name, ort: ort, assoc: assoc}
	l.d.container.reihe = append(l.d.container.reihe, kindRef{aset: s})
	l.attributFunde(se, ort)
	l.push(rahmen{art: kAssocSet, ort: ort, aset: s})
	return nil
}

// assocSetEnd liest ein End der AssociationSet: höchstens zwei, verschiedene Rollen (Nr. 10 c).
func (l *lauf) assocSetEnd(p rahmen, se xml.StartElement) error {
	s := p.aset
	if len(s.ends) >= 2 {
		return struktur(s.ort)
	}
	var e rawSetEnd
	var err error
	if e.role, err = pflicht(se, s.ort, "Role"); err != nil {
		return err
	}
	if e.set, err = pflicht(se, s.ort, "EntitySet"); err != nil {
		return err
	}
	if len(s.ends) == 1 && s.ends[0].role == e.role {
		return struktur(s.ort + "@Role")
	}
	s.ends = append(s.ends, e)
	l.attributFunde(se, s.ort)
	l.push(rahmen{art: kAssocSetEnd, ort: s.ort, aset: s})
	return nil
}

package jsoncodecv2

import (
	"slices"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// metadataName ist der Name der Hülle einer Entity; eine Property dieses Namens ist nie ein Teil der Ausgabe (SPEC-172 Nr. 6 j).
const metadataName = "__metadata"

// typSicht ist ein geprüfter Typ mit dem Verzeichnis seiner wirksamen Properties.
type typSicht struct {
	typ model.EntityType
	idx map[string]int
}

// sicht gehört einem Aufruf: das Modell, das Set und die Typen, die der Aufruf schon geprüft hat.
type sicht struct {
	m     model.ServiceModel
	set   model.EntityType
	typen map[string]*typSicht
}

// setTyp prüft Modell, Version und Set (SPEC-168 Nr. 2) und liefert den Typ des Sets.
func setTyp(m model.ServiceModel, name string) (model.EntityType, error) {
	if m.Version() != versionSchreiber {
		return model.EntityType{}, ungueltig("Modell hat nicht die Version %s", versionSchreiber)
	}
	s, ok := m.EntitySet(name)
	if !ok || name == "" {
		return model.EntityType{}, ungueltig("Set ist dem Modell unbekannt")
	}
	t, ok := m.EntityType(s.EntityType)
	if !ok {
		return model.EntityType{}, ungueltig("Typ des Sets ist dem Modell unbekannt")
	}
	return t, nil
}

func neueSicht(m model.ServiceModel, set string) (*sicht, error) {
	t, err := setTyp(m, set)
	if err != nil {
		return nil, err
	}
	return &sicht{m: m, set: t, typen: map[string]*typSicht{}}, nil
}

// typ liefert den Typ mit dem Namen, wenn er nicht abstrakt ist und der Typ des Sets oder von ihm abgeleitet; die Kette
// wird je Typ höchstens einmal besucht (SPEC-168 Nr. 3 a).
func (s *sicht) typ(name string) (*typSicht, error) {
	if ts, ok := s.typen[name]; ok {
		return ts, nil
	}
	t, ok := s.m.EntityType(name)
	if !ok || t.Abstract {
		return nil, ungueltig("Typ der Entity ist unbekannt oder abstrakt")
	}
	besucht := map[string]bool{}
	for cur := t; cur.Name != s.set.Name; {
		besucht[cur.Name] = true
		next, ok := s.m.EntityType(cur.BaseType)
		if !ok || besucht[next.Name] {
			return nil, ungueltig("Typ der Entity ist nicht vom Typ des Sets abgeleitet")
		}
		cur = next
	}
	ts := &typSicht{typ: t, idx: map[string]int{}}
	for i, p := range t.Properties {
		if _, doppelt := ts.idx[p.Name]; !doppelt {
			ts.idx[p.Name] = i
		}
	}
	s.typen[name] = ts
	return ts, nil
}

// pruefeAuswahl prüft die Elemente gegen die Properties und Navigation-Properties des Typs des Sets (SPEC-168 Nr. 5 a); die Auswahl
// hat keine Wirkung auf die Ausgabe (SPEC-172 Nr. 6 m).
func (s *sicht) pruefeAuswahl(auswahl []string) error {
	for _, e := range auswahl {
		if e == "*" {
			continue
		}
		if slices.ContainsFunc(s.set.Properties, func(p model.Property) bool { return p.Name == e }) ||
			slices.ContainsFunc(s.set.NavigationProperties, func(n model.NavigationProperty) bool { return n.Name == e }) {
			continue
		}
		return ungueltig("Element der Auswahl ist unbekannt")
	}
	return nil
}

// werte prüft die Properties der Entity (SPEC-168 Nr. 3 c, 4; SPEC-172 Nr. 6 j, k) und liefert zu jedem Namen den Index in e.Properties.
func (ts *typSicht) werte(e model.Entity) (map[string]int, error) {
	werte := make(map[string]int, len(e.Properties))
	for i, pw := range e.Properties {
		pi, ok := ts.idx[pw.Name]
		if !ok || !ts.typ.Properties[pi].Unterstuetzt() {
			return nil, ungueltig("Name der Property ist unbekannt oder nicht unterstützt")
		}
		if pw.Name == metadataName || ts.typ.Properties[pi].EdmType == "Edm.Date" {
			return nil, ungueltig("Property ist in V2 nie ein Teil der Antwort")
		}
		if _, doppelt := werte[pw.Name]; doppelt {
			return nil, ungueltig("Property kommt doppelt vor")
		}
		if err := pruefeWert(ts.typ.Properties[pi], pw.Wert); err != nil {
			return nil, err
		}
		werte[pw.Name] = i
	}
	return werte, nil
}

// entity prüft die Entity (SPEC-168 Nr. 3, 4) und hängt sie ohne die Hülle `d` an (SPEC-172 Nr. 6 a).
func (s *sicht) entity(a *ausgabe, e model.Entity, wurzel, set string) error {
	ts, err := s.typ(e.Typ)
	if err != nil {
		return err
	}
	adr, err := adresse(wurzel, set, e.Schluessel)
	if err != nil {
		return err
	}
	werte, err := ts.werte(e)
	if err != nil {
		return err
	}
	a.roh(`{"__metadata":{"uri":`)
	a.text(adr)
	a.roh(`,"type":`)
	a.text(e.Typ)
	a.roh("}")
	for _, p := range ts.typ.Properties {
		i, ok := werte[p.Name]
		if !ok {
			continue
		}
		a.roh(",")
		a.text(p.Name)
		a.roh(":")
		a.wert(e.Properties[i].Wert)
	}
	a.roh("}")
	return nil
}

// wert hängt den geprüften Wert an (SPEC-172 Nr. 6 f).
func (a *ausgabe) wert(w model.Wert) {
	switch w.Art {
	case model.ArtNull:
		a.roh("null")
	case model.ArtString:
		a.stringWert(w.Text)
	case model.ArtBoolean, model.ArtInt32:
		a.roh(w.Text)
	case model.ArtDateTimeOffset:
		ms, _ := zeitpunktMillis(w.Text)
		a.roh(`"\/Date(` + strconv.FormatInt(ms, 10) + `)\/"`)
	default:
		a.roh(`"` + w.Text + `"`)
	}
}

// baueEntity prüft die Argumente und baut die Ausgabe von SchreibeEntity.
func baueEntity(m model.ServiceModel, an model.EntityAntwort) ([]byte, error) {
	s, err := neueSicht(m, an.Set)
	if err != nil {
		return nil, err
	}
	if err := s.pruefeAuswahl(an.Auswahl); err != nil {
		return nil, err
	}
	var a ausgabe
	a.roh(`{"d":`)
	if err := s.entity(&a, an.Entity, an.Wurzel, an.Set); err != nil {
		return nil, err
	}
	a.roh("}")
	return a.ergebnis()
}

// baueCollection prüft die Argumente und baut die Ausgabe von SchreibeCollection.
func baueCollection(m model.ServiceModel, an model.CollectionAntwort) ([]byte, error) {
	s, err := neueSicht(m, an.Set)
	if err != nil {
		return nil, err
	}
	if !wurzelGueltig(an.Wurzel) {
		return nil, ungueltig("Wurzel hat nicht die Form von SPEC-172 Nr. 6 h")
	}
	if err := s.pruefeAuswahl(an.Auswahl); err != nil {
		return nil, err
	}
	if an.Anzahl.Vorhanden && an.Anzahl.Wert < 0 {
		return nil, ungueltig("Anzahl negativ")
	}
	var a ausgabe
	a.roh(`{"d":{`)
	if an.Anzahl.Vorhanden {
		a.roh(`"__count":"` + strconv.FormatInt(an.Anzahl.Wert, 10) + `",`)
	}
	a.roh(`"results":[`)
	for i, e := range an.Entities {
		if i > 0 {
			a.roh(",")
		}
		if err := s.entity(&a, e, an.Wurzel, an.Set); err != nil {
			return nil, err
		}
	}
	a.roh("]}}")
	return a.ergebnis()
}

// ergebnis liefert die Bytes, wenn kein Text ungültiges UTF-8 war.
func (a *ausgabe) ergebnis() ([]byte, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.b, nil
}

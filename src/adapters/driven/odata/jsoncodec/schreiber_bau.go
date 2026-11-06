package jsoncodec

import (
	"slices"
	"strconv"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

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
	if m.Version() != version {
		return model.EntityType{}, ungueltig("Modell hat nicht die Version %s", version)
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

// pruefeAuswahl prüft die Elemente gegen die Properties und Navigation-Properties des Typs des Sets (SPEC-168 Nr. 5 a).
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

// kontext ist `<Wurzel>/$metadata#<Set>`, bei nichtleerer Auswahl mit `(<Auswahl>)` (SPEC-169 Nr. 5).
func kontext(wurzel, set string, auswahl []string) string {
	k := wurzel + "/$metadata#" + set
	if len(auswahl) > 0 {
		k += "(" + strings.Join(auswahl, ",") + ")"
	}
	return k
}

// entity prüft die Entity (SPEC-168 Nr. 3, 4) und hängt sie als Objekt an; ein nichtleerer ktx ist der Wert von `@odata.context`.
func (s *sicht) entity(a *ausgabe, e model.Entity, wurzel, set string, f form, ktx string) error {
	ts, err := s.typ(e.Typ)
	if err != nil {
		return err
	}
	adr, err := adresse(wurzel, set, e.Schluessel)
	if err != nil {
		return err
	}
	werte := make(map[string]int, len(e.Properties))
	for i, pw := range e.Properties {
		pi, ok := ts.idx[pw.Name]
		if !ok || !ts.typ.Properties[pi].Unterstuetzt() {
			return ungueltig("Name der Property ist unbekannt oder nicht unterstützt")
		}
		if _, doppelt := werte[pw.Name]; doppelt {
			return ungueltig("Property kommt doppelt vor")
		}
		if err := pruefeWert(ts.typ.Properties[pi], pw.Wert); err != nil {
			return err
		}
		werte[pw.Name] = i
	}
	a.roh("{")
	erstes := true
	if ktx != "" {
		a.mitglied(&erstes, "@odata.context")
		a.text(ktx)
	}
	if !f.keine {
		if e.Typ != s.set.Name {
			a.mitglied(&erstes, "@odata.type")
			a.text("#" + e.Typ)
		}
		a.mitglied(&erstes, "@odata.id")
		a.text(adr)
	}
	for _, p := range ts.typ.Properties {
		i, ok := werte[p.Name]
		if !ok {
			continue
		}
		a.mitglied(&erstes, p.Name)
		a.wert(e.Properties[i].Wert, f)
	}
	a.roh("}")
	return nil
}

// wert hängt den geprüften Wert an (SPEC-169 Nr. 4).
func (a *ausgabe) wert(w model.Wert, f form) {
	switch w.Art {
	case model.ArtNull:
		a.roh("null")
	case model.ArtString:
		a.text(w.Text)
	case model.ArtBoolean, model.ArtInt32:
		a.roh(w.Text)
	case model.ArtInt64, model.ArtDecimal:
		a.zahl(w.Text, f)
	default:
		a.roh(`"` + w.Text + `"`)
	}
}

// zahl hängt eine Zahl ziffernweise an, bei der Zahlform `Zeichenkette` in Anführungszeichen.
func (a *ausgabe) zahl(text string, f form) {
	if f.zeichenkette {
		text = `"` + text + `"`
	}
	a.roh(text)
}

// baueEntity prüft die Argumente und baut die Ausgabe von SchreibeEntity.
func baueEntity(m model.ServiceModel, an model.EntityAntwort) ([]byte, model.Darstellung, error) {
	s, err := neueSicht(m, an.Set)
	if err != nil {
		return nil, model.Darstellung{}, err
	}
	if err := s.pruefeAuswahl(an.Auswahl); err != nil {
		return nil, model.Darstellung{}, err
	}
	f := aushandeln(an.Accept)
	var a ausgabe
	ktx := ""
	if !f.keine {
		ktx = kontext(an.Wurzel, an.Set, an.Auswahl) + "/$entity"
	}
	if err := s.entity(&a, an.Entity, an.Wurzel, an.Set, f, ktx); err != nil {
		return nil, model.Darstellung{}, err
	}
	return a.ergebnis(f)
}

// baueCollection prüft die Argumente und baut die Ausgabe von SchreibeCollection.
func baueCollection(m model.ServiceModel, an model.CollectionAntwort) ([]byte, model.Darstellung, error) {
	s, err := neueSicht(m, an.Set)
	if err != nil {
		return nil, model.Darstellung{}, err
	}
	if !wurzelGueltig(an.Wurzel) {
		return nil, model.Darstellung{}, ungueltig("Wurzel hat nicht die Form von SPEC-169 Nr. 2")
	}
	if err := s.pruefeAuswahl(an.Auswahl); err != nil {
		return nil, model.Darstellung{}, err
	}
	if an.Anzahl.Vorhanden && an.Anzahl.Wert < 0 {
		return nil, model.Darstellung{}, ungueltig("Anzahl negativ")
	}
	f := aushandeln(an.Accept)
	var a ausgabe
	a.roh("{")
	erstes := true
	if !f.keine {
		a.mitglied(&erstes, "@odata.context")
		a.text(kontext(an.Wurzel, an.Set, an.Auswahl))
	}
	if an.Anzahl.Vorhanden {
		a.mitglied(&erstes, "@odata.count")
		a.zahl(strconv.FormatInt(an.Anzahl.Wert, 10), f)
	}
	a.mitglied(&erstes, "value")
	a.roh("[")
	for i, e := range an.Entities {
		if i > 0 {
			a.roh(",")
		}
		if err := s.entity(&a, e, an.Wurzel, an.Set, f, ""); err != nil {
			return nil, model.Darstellung{}, err
		}
	}
	a.roh("]}")
	return a.ergebnis(f)
}

// ergebnis liefert die Bytes und den Kopf, wenn kein Text ungültiges UTF-8 war.
func (a *ausgabe) ergebnis(f form) ([]byte, model.Darstellung, error) {
	if a.err != nil {
		return nil, model.Darstellung{}, a.err
	}
	return a.b, model.Darstellung{ContentType: f.contentType(), Kennzeichen: kennzeichen()}, nil
}

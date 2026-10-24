package jsoncodec

import (
	"strconv"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// typInfo ist die Sicht auf einen Entity-Typ mit Verzeichnissen seiner Members (SPEC-125 Nr. 11).
type typInfo struct {
	typ        model.EntityType
	props      map[string]int
	navs       map[string]struct{}
	schluessel map[string]struct{}
}

// deuter deutet den Baum gegen das Modell (SPEC-126 Nr. 6 bis 13). Er gehört einem Aufruf.
type deuter struct {
	m      model.ServiceModel
	set    string
	setTyp *typInfo
	typen  map[string]*typInfo
	pfad   []string
}

// neuerDeuter liefert den Deuter für das Entity-Set; ok ist falsch, wenn das Modell das Set oder seinen Typ nicht kennt.
func neuerDeuter(m model.ServiceModel, set string) (*deuter, bool) {
	s, ok := m.EntitySet(set)
	if !ok {
		return nil, false
	}
	d := &deuter{m: m, set: set, typen: map[string]*typInfo{}}
	d.setTyp, ok = d.info(s.EntityType)
	return d, ok
}

// info liefert die Sicht auf den Typ mit dem qualifizierten Namen.
func (d *deuter) info(name string) (*typInfo, bool) {
	if i, ok := d.typen[name]; ok {
		return i, i != nil
	}
	t, ok := d.m.EntityType(name)
	if !ok {
		d.typen[name] = nil
		return nil, false
	}
	i := &typInfo{typ: t, props: map[string]int{}, navs: map[string]struct{}{}, schluessel: map[string]struct{}{}}
	for n, p := range t.Properties {
		i.props[p.Name] = n
	}
	for _, n := range t.NavigationProperties {
		i.navs[n.Name] = struct{}{}
	}
	for _, k := range t.Keys {
		i.schluessel[k] = struct{}{}
	}
	d.typen[name] = i
	return i, true
}

// fehler bildet den DatenFehler mit dem Pointer auf pfad und die weiteren Segmente (SPEC-126 Nr. 14): in einem
// Segment steht `~0` für `~` und `~1` für `/`.
func (d *deuter) fehler(u model.DatenUrsache, weitere ...string) error {
	ersatz := strings.NewReplacer("~", "~0", "/", "~1")
	var b strings.Builder
	for _, s := range d.pfad {
		b.WriteByte('/')
		b.WriteString(ersatz.Replace(s))
	}
	for _, s := range weitere {
		b.WriteByte('/')
		b.WriteString(ersatz.Replace(s))
	}
	return &model.DatenFehler{Ursache: u, Ort: b.String()}
}

// annotationArt unterscheidet die Mitglieder nach ihrem Namen (SPEC-126 Nr. 6).
type annotationArt uint8

const (
	keineAnnotation annotationArt = iota
	annotationObjekt
	annotationProperty
	annotationUngueltig
)

func artDesNamens(name string) annotationArt {
	i := strings.IndexByte(name, '@')
	switch {
	case i < 0:
		return keineAnnotation
	case name[i+1:] == "" || strings.Contains(name[i+1:], "@"):
		return annotationUngueltig
	case i == 0:
		return annotationObjekt
	}
	return annotationProperty
}

// sammlung deutet die Wurzel einer Collection (SPEC-126 Nr. 6 a, 8, 13 c).
func (d *deuter) sammlung(w *knoten) (model.EntitySeite, error) {
	if w.art != artObjekt {
		return model.EntitySeite{}, d.fehler(model.DatenStruktur)
	}
	seite := model.EntitySeite{Entities: []model.Entity{}}
	var kontext, hatWert bool
	for i, name := range w.namen {
		k := &w.kinder[i]
		var err error
		switch name {
		case "@odata.context":
			kontext = true
			err = d.pruefeKontext(k, false)
		case "@odata.count":
			seite.Anzahl, err = d.anzahl(k)
		case "@odata.nextLink":
			seite.FolgeLink, err = d.folgeLink(k)
		case "value":
			hatWert = true
			seite.Entities, err = d.elemente(k)
		default:
			if artDesNamens(name) != annotationObjekt {
				err = d.fehler(model.DatenStruktur, name)
			}
		}
		if err != nil {
			return model.EntitySeite{}, err
		}
	}
	if !kontext {
		return model.EntitySeite{}, d.fehler(model.DatenKontext, "@odata.context")
	}
	if !hatWert {
		return model.EntitySeite{}, d.fehler(model.DatenStruktur, "value")
	}
	return seite, nil
}

// elemente deutet `value`: ein Array von Objekten (SPEC-126 Nr. 6 a).
func (d *deuter) elemente(k *knoten) ([]model.Entity, error) {
	if k.art != artArray {
		return nil, d.fehler(model.DatenStruktur, "value")
	}
	out := make([]model.Entity, 0, len(k.kinder))
	for i := range k.kinder {
		el := &k.kinder[i]
		index := strconv.Itoa(i)
		if el.art != artObjekt {
			return nil, d.fehler(model.DatenStruktur, "value", index)
		}
		d.pfad = append(d.pfad, "value", index)
		e, err := d.entity(el, false)
		d.pfad = d.pfad[:len(d.pfad)-2]
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// einzel deutet die Wurzel einer Einzel-Entity (SPEC-126 Nr. 6 b).
func (d *deuter) einzel(w *knoten) (model.Entity, error) {
	if w.art != artObjekt {
		return model.Entity{}, d.fehler(model.DatenStruktur)
	}
	return d.entity(w, true)
}

// anzahl deutet `@odata.count` (SPEC-126 Nr. 8 c).
func (d *deuter) anzahl(k *knoten) (model.Anzahl, error) {
	z, u := zahlVon(k, true)
	if u != "" {
		return model.Anzahl{}, d.fehler(u, "@odata.count")
	}
	text, ok := z.ganzText(64)
	if !ok {
		return model.Anzahl{}, d.fehler(model.DatenWert, "@odata.count")
	}
	v, _ := strconv.ParseInt(text, 10, 64)
	if v < 0 {
		return model.Anzahl{}, d.fehler(model.DatenWert, "@odata.count")
	}
	return model.Anzahl{Vorhanden: true, Wert: v}, nil
}

// folgeLink deutet `@odata.nextLink` (SPEC-126 Nr. 8 b).
func (d *deuter) folgeLink(k *knoten) (model.FolgeLink, error) {
	switch {
	case k.art != artString:
		return model.FolgeLink{}, d.fehler(model.DatenTyp, "@odata.nextLink")
	case k.text == "":
		return model.FolgeLink{}, d.fehler(model.DatenWert, "@odata.nextLink")
	case k.marker:
		return model.FolgeLink{Vorhanden: true, Redigiert: true}, nil
	}
	return model.FolgeLink{Vorhanden: true, Wert: k.text}, nil
}

package jsoncodecv2

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

// deuter deutet den Baum gegen das Modell (SPEC-172 Nr. 2 bis 5). Er gehört einem Aufruf.
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

// fehler bildet den DatenFehler mit dem Pointer auf pfad und die weiteren Segmente (SPEC-126 Nr. 14): in einem Segment steht
// `~0` für `~` und `~1` für `/`.
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

// huelle prüft die Wurzel (SPEC-172 Nr. 2 c, Schritte 1 bis 3): ein Objekt mit dem Mitglied `d`, sonst nichts. Es liefert `d`.
func (d *deuter) huelle(w *knoten) (*knoten, error) {
	if w.art != artObjekt {
		return nil, d.fehler(model.DatenStruktur)
	}
	var huelle *knoten
	for i, name := range w.namen {
		if name == "d" {
			huelle = &w.kinder[i]
		}
	}
	if huelle == nil {
		return nil, d.fehler(model.DatenStruktur, "d")
	}
	for _, name := range w.namen {
		if name != "d" {
			return nil, d.fehler(model.DatenStruktur, name)
		}
	}
	return huelle, nil
}

// sammlung deutet die Antwort auf eine Collection (SPEC-172 Nr. 2 a): `d` ist ein Array von Objekten oder ein Objekt mit `results`.
func (d *deuter) sammlung(w *knoten) (model.EntitySeite, error) {
	huelle, err := d.huelle(w)
	if err != nil {
		return model.EntitySeite{}, err
	}
	d.pfad = []string{"d"}
	switch huelle.art {
	case artArray:
		es, err := d.elemente(huelle)
		return model.EntitySeite{Entities: es}, err
	case artObjekt:
		return d.mitResults(huelle)
	}
	return model.EntitySeite{}, d.fehler(model.DatenStruktur)
}

// mitResults deutet das Objekt `d` einer Collection: `results` ist Pflicht, `__count` und `__next` sind möglich, jedes andere
// Mitglied ist `struktur`; die Mitglieder werden in der Reihenfolge des Dokuments gedeutet.
func (d *deuter) mitResults(huelle *knoten) (model.EntitySeite, error) {
	seite := model.EntitySeite{Entities: []model.Entity{}}
	var hatResults bool
	for i, name := range huelle.namen {
		k := &huelle.kinder[i]
		var err error
		switch name {
		case "results":
			hatResults = true
			seite.Entities, err = d.results(k)
		case "__count":
			seite.Anzahl, err = d.anzahl(k)
		case "__next":
			seite.FolgeLink, err = d.folgeLink(k)
		default:
			err = d.fehler(model.DatenStruktur, name)
		}
		if err != nil {
			return model.EntitySeite{}, err
		}
	}
	if !hatResults {
		return model.EntitySeite{}, d.fehler(model.DatenStruktur, "results")
	}
	return seite, nil
}

// results deutet `results`: ein Array von Objekten.
func (d *deuter) results(k *knoten) ([]model.Entity, error) {
	if k.art != artArray {
		return nil, d.fehler(model.DatenStruktur, "results")
	}
	d.pfad = append(d.pfad, "results")
	es, err := d.elemente(k)
	d.pfad = d.pfad[:len(d.pfad)-1]
	return es, err
}

// elemente deutet die Elemente eines Arrays von Entities; die Elemente werden der Reihe nach gedeutet, jedes von Anfang bis Ende.
func (d *deuter) elemente(k *knoten) ([]model.Entity, error) {
	out := make([]model.Entity, 0, len(k.kinder))
	for i := range k.kinder {
		el := &k.kinder[i]
		index := strconv.Itoa(i)
		if el.art != artObjekt {
			return nil, d.fehler(model.DatenStruktur, index)
		}
		d.pfad = append(d.pfad, index)
		e, err := d.entity(el)
		d.pfad = d.pfad[:len(d.pfad)-1]
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// einzel deutet die Antwort auf eine Einzel-Entity (SPEC-172 Nr. 2 b): `d` ist ein Objekt und die Entity.
func (d *deuter) einzel(w *knoten) (model.Entity, error) {
	huelle, err := d.huelle(w)
	if err != nil {
		return model.Entity{}, err
	}
	if huelle.art != artObjekt {
		return model.Entity{}, d.fehler(model.DatenStruktur, "d")
	}
	d.pfad = []string{"d"}
	return d.entity(huelle)
}

// anzahl deutet `__count` (SPEC-172 Nr. 2 a).
func (d *deuter) anzahl(k *knoten) (model.Anzahl, error) {
	z, u := zahlVon(k, true)
	if u != "" {
		return model.Anzahl{}, d.fehler(u, "__count")
	}
	text, ok := z.ganzText(64)
	if !ok {
		return model.Anzahl{}, d.fehler(model.DatenWert, "__count")
	}
	v, _ := strconv.ParseInt(text, 10, 64)
	if v < 0 {
		return model.Anzahl{}, d.fehler(model.DatenWert, "__count")
	}
	return model.Anzahl{Vorhanden: true, Wert: v}, nil
}

// folgeLink deutet `__next` (SPEC-172 Nr. 2 a mit SPEC-126 Nr. 8 b).
func (d *deuter) folgeLink(k *knoten) (model.FolgeLink, error) {
	switch {
	case k.art != artString:
		return model.FolgeLink{}, d.fehler(model.DatenTyp, "__next")
	case k.text == "":
		return model.FolgeLink{}, d.fehler(model.DatenWert, "__next")
	case k.marker:
		return model.FolgeLink{Vorhanden: true, Redigiert: true}, nil
	}
	return model.FolgeLink{Vorhanden: true, Wert: k.text}, nil
}

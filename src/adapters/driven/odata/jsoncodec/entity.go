package jsoncodec

import (
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// pruefeKontext prüft `@odata.context` der Wurzel (SPEC-126 Nr. 7); einzel verlangt die Form der Einzel-Entity.
func (d *deuter) pruefeKontext(k *knoten, einzel bool) error {
	if d.kontextGueltig(k, einzel) {
		return nil
	}
	return d.fehler(model.DatenKontext, "@odata.context")
}

// kontextGueltig prüft die Form `<Pfad>#<Set>[(<Liste>)][/$entity]` (SPEC-126 Nr. 7). Nur ein String trägt ein `#`; jeder
// andere Knoten hat keinen Text mit Pfad und Fragment und fällt an der ersten Prüfung oder am Set durch.
func (d *deuter) kontextGueltig(k *knoten, einzel bool) bool {
	pfad, fragment, _ := strings.Cut(k.text, "#")
	if pfad != "$metadata" && !strings.HasSuffix(pfad, "/$metadata") {
		return false
	}
	var ok bool
	if einzel {
		if fragment, ok = strings.CutSuffix(fragment, "/$entity"); !ok {
			return false
		}
	}
	set, liste, klammer := strings.Cut(fragment, "(")
	if set != d.set {
		return false
	}
	if !klammer {
		return true
	}
	liste, ok = strings.CutSuffix(liste, ")")
	if !ok {
		return false
	}
	for _, e := range strings.Split(liste, ",") {
		if e == "*" {
			continue
		}
		_, prop := d.setTyp.props[e]
		_, nav := d.setTyp.navs[e]
		if !prop && !nav {
			return false
		}
	}
	return true
}

// abgeleitet liefert den Typ, den `@odata.type` nennt (SPEC-126 Nr. 8 a): der Typ des Sets oder ein von ihm
// abgeleiteter, nicht abstrakter Typ des Modells.
func (d *deuter) abgeleitet(k *knoten) (*typInfo, bool) {
	name, ok := strings.CutPrefix(k.text, "#")
	if k.art != artString || !ok {
		return nil, false
	}
	info, ok := d.info(name)
	if !ok || info.typ.Abstract {
		return nil, false
	}
	besucht := map[string]bool{}
	for t := info; t != nil && !besucht[t.typ.Name]; t, _ = d.info(t.typ.BaseType) {
		if t == d.setTyp {
			return info, true
		}
		besucht[t.typ.Name] = true
	}
	return nil, false
}

// entity deutet ein Entity-Objekt (SPEC-126 Nr. 9, 11, 13 c): Element von `value` oder Wurzel von ParseEntity.
func (d *deuter) entity(o *knoten, wurzel bool) (model.Entity, error) {
	info := d.setTyp
	for i, name := range o.namen {
		if name != "@odata.type" {
			continue
		}
		var ok bool
		if info, ok = d.abgeleitet(&o.kinder[i]); !ok {
			return model.Entity{}, d.fehler(model.DatenODataTyp, name)
		}
	}
	e := model.Entity{Typ: info.typ.Name, Properties: []model.PropertyWert{}, Uebersprungen: []model.Uebersprungen{}}
	var kontext bool
	for i, name := range o.namen {
		k := &o.kinder[i]
		if err := d.mitglied(info, &e, name, k, wurzel); err != nil {
			return model.Entity{}, err
		}
		kontext = kontext || (wurzel && name == "@odata.context")
	}
	if wurzel && !kontext {
		return model.Entity{}, d.fehler(model.DatenKontext, "@odata.context")
	}
	return e, d.schluessel(info, &e)
}

// mitglied deutet ein Mitglied des Entity-Objekts: eine Annotation, eine Property, eine Navigation oder einen
// anderen Namen.
func (d *deuter) mitglied(info *typInfo, e *model.Entity, name string, k *knoten, wurzel bool) error {
	switch artDesNamens(name) {
	case annotationUngueltig:
		return d.fehler(model.DatenStruktur, name)
	case annotationObjekt:
		switch {
		case wurzel && name == "@odata.context":
			return d.pruefeKontext(k, true)
		case wurzel && (name == "@odata.count" || name == "@odata.nextLink"):
			return d.fehler(model.DatenStruktur, name)
		}
		return nil
	case annotationProperty:
		return nil
	}
	if n, ok := info.props[name]; ok {
		return d.property(info, e, info.typ.Properties[n], k)
	}
	if _, ok := info.navs[name]; ok {
		if k.art != artObjekt && k.art != artArray && k.art != artNull {
			return d.fehler(model.DatenTyp, name)
		}
		e.Uebersprungen = append(e.Uebersprungen, model.Uebersprungen{Name: name, Grund: model.UebersprungNavigation})
		return nil
	}
	if !info.typ.Open {
		return d.fehler(model.DatenUnbekannteProperty, name)
	}
	e.Uebersprungen = append(e.Uebersprungen, model.Uebersprungen{Name: name, Grund: model.UebersprungOffen})
	return nil
}

// property deutet eine Property des Typs: unterstützt gelesen, sonst übersprungen. Ein Schlüssel ist weder `null`
// noch der Marker noch ein String mit U+0000 (SPEC-126 Nr. 11).
func (d *deuter) property(info *typInfo, e *model.Entity, p model.Property, k *knoten) error {
	if !p.Unterstuetzt() {
		e.Uebersprungen = append(e.Uebersprungen, model.Uebersprungen{Name: p.Name, Grund: model.UebersprungTyp})
		return nil
	}
	_, istSchluessel := info.schluessel[p.Name]
	if istSchluessel && k.art == artNull {
		return d.fehler(model.DatenSchluessel, p.Name)
	}
	w, u := lesWert(p, k)
	if u != "" {
		return d.fehler(u, p.Name)
	}
	if istSchluessel && (w.Art == model.ArtRedigiert || (w.Art == model.ArtString && strings.Contains(w.Text, "\x00"))) {
		return d.fehler(model.DatenSchluessel, p.Name)
	}
	e.Properties = append(e.Properties, model.PropertyWert{Name: p.Name, Wert: w})
	return nil
}

// schluessel bildet die Literalform des Schlüssels in der Reihenfolge des Modells (SPEC-126 Nr. 11); der erste
// fehlende Teil ist der Fehler.
func (d *deuter) schluessel(info *typInfo, e *model.Entity) error {
	if len(info.typ.Keys) == 0 {
		return d.fehler(model.DatenSchluessel)
	}
	teile := make([]string, 0, len(info.typ.Keys))
	for _, name := range info.typ.Keys {
		var w *model.Wert
		for i := range e.Properties {
			if e.Properties[i].Name == name {
				w = &e.Properties[i].Wert
				break
			}
		}
		if w == nil {
			return d.fehler(model.DatenSchluessel, name)
		}
		lit := w.Text
		if w.Art == model.ArtString {
			lit = "'" + strings.ReplaceAll(lit, "'", "''") + "'"
		}
		if len(info.typ.Keys) > 1 {
			lit = name + "=" + lit
		}
		teile = append(teile, lit)
	}
	e.Schluessel = strings.Join(teile, ",")
	return nil
}

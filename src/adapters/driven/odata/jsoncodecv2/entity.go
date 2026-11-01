package jsoncodecv2

import (
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// metadata deutet `__metadata` (SPEC-172 Nr. 3): ein Objekt, dessen Mitglied `type` einen Entity-Typ nennt; jedes andere Mitglied
// ist ohne Wirkung. Es liefert den Typ der Entity.
func (d *deuter) metadata(k *knoten) (*typInfo, error) {
	if k.art != artObjekt {
		return nil, d.fehler(model.DatenStruktur, "__metadata")
	}
	for i, name := range k.namen {
		if name != "type" {
			continue
		}
		info, ok := d.abgeleitet(&k.kinder[i])
		if !ok {
			return nil, d.fehler(model.DatenODataTyp, "__metadata", "type")
		}
		return info, nil
	}
	return d.setTyp, nil
}

// abgeleitet liefert den Typ, den `type` nennt: ein String, der den qualifizierten Namen des Typs des Sets oder eines von ihm
// abgeleiteten, nicht abstrakten Typs des Modells trägt.
func (d *deuter) abgeleitet(k *knoten) (*typInfo, bool) {
	if k.art != artString {
		return nil, false
	}
	info, ok := d.info(k.text)
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

// entity deutet ein Entity-Objekt (SPEC-172 Nr. 2 c Schritt 5, Nr. 3, Nr. 5): zuerst `__metadata`, wo es im Objekt steht, dann
// die übrigen Mitglieder in der Reihenfolge des Dokuments, zuletzt der Schlüssel.
func (d *deuter) entity(o *knoten) (model.Entity, error) {
	info := d.setTyp
	for i, name := range o.namen {
		if name != "__metadata" {
			continue
		}
		var err error
		if info, err = d.metadata(&o.kinder[i]); err != nil {
			return model.Entity{}, err
		}
		break
	}
	e := model.Entity{Typ: info.typ.Name, Properties: []model.PropertyWert{}, Uebersprungen: []model.Uebersprungen{}}
	for i, name := range o.namen {
		if name == "__metadata" {
			continue
		}
		if err := d.mitglied(info, &e, name, &o.kinder[i]); err != nil {
			return model.Entity{}, err
		}
	}
	return e, d.schluessel(info, &e)
}

// mitglied deutet ein Mitglied des Entity-Objekts: eine Property, eine Navigation oder einen anderen Namen. V2 kennt keine
// Annotationen und keinen offenen Typ: jeder andere Name ist unbekannt.
func (d *deuter) mitglied(info *typInfo, e *model.Entity, name string, k *knoten) error {
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
	return d.fehler(model.DatenUnbekannteProperty, name)
}

// property deutet eine Property des Typs: Typen der sieben gelesen, sonst übersprungen, auch `Edm.Date`, den V2 nicht kennt. Ein
// Schlüssel ist weder `null` noch der Marker noch ein String mit U+0000 (SPEC-126 Nr. 11).
func (d *deuter) property(info *typInfo, e *model.Entity, p model.Property, k *knoten) error {
	if !p.Unterstuetzt() || p.EdmType == "Edm.Date" {
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

// schluessel bildet die Literalform des Schlüssels in der Reihenfolge des Modells (SPEC-126 Nr. 11); der erste fehlende Teil
// ist der Fehler.
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

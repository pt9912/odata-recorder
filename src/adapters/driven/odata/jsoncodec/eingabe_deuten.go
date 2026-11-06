package jsoncodec

import (
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eingabeFehler bildet den EingabeFehler ohne Ort.
func eingabeFehler(u model.EingabeUrsache) error { return &model.EingabeFehler{Ursache: u} }

// eingabeFehlerAn bildet den EingabeFehler am Mitglied name der Wurzel: Ort ist der JSON-Pointer (SPEC-180 Nr. 3), in dem `~0`
// für `~` und `~1` für `/` steht und sonst nichts ersetzt wird.
func eingabeFehlerAn(u model.EingabeUrsache, name string) error {
	return &model.EingabeFehler{Ursache: u, Ort: "/" + strings.NewReplacer("~", "~0", "/", "~1").Replace(name)}
}

// eingabe deutet die Wurzel gegen das Modell in der Folge von SPEC-181 Nr. 8 f: Wurzel, Typ, Mitglieder in der Reihenfolge des
// Dokuments, Pflicht am Ende.
func (d *deuter) eingabe(w *knoten, art model.EingabeArt) (model.EntityEingabe, error) {
	if w.art != artObjekt {
		return model.EntityEingabe{}, eingabeFehler(model.EingabeStruktur)
	}
	info, err := d.eingabeTyp(w)
	if err != nil {
		return model.EntityEingabe{}, err
	}
	out := model.EntityEingabe{Typ: info.typ.Name, Properties: make([]model.PropertyWert, 0, min(len(w.namen), len(info.typ.Properties)))}
	genannt := map[string]int{}
	for i, name := range w.namen {
		if err := eingabeMitglied(info, &out, genannt, name, &w.kinder[i]); err != nil {
			return model.EntityEingabe{}, err
		}
	}
	if art != model.EingabePatch {
		if name, fehlt := erforderlichFehlt(info, genannt); fehlt {
			return model.EntityEingabe{}, eingabeFehlerAn(model.EingabeFehlend, name)
		}
	}
	out.Schluessel = eingabeSchluessel(info, &out, genannt)
	return out, nil
}

// eingabeTyp liefert den Typ der Entity (SPEC-181 Nr. 3): `@odata.type` der Wurzel, sonst der Typ des Sets; ein abstrakter Typ legt
// keine Entity an.
func (d *deuter) eingabeTyp(w *knoten) (*typInfo, error) {
	for i, name := range w.namen {
		if name != "@odata.type" {
			continue
		}
		info, ok := d.abgeleitet(&w.kinder[i])
		if !ok {
			return nil, eingabeFehlerAn(model.EingabeODataTyp, name)
		}
		return info, nil
	}
	if d.setTyp.typ.Abstract {
		return nil, eingabeFehlerAn(model.EingabeODataTyp, "@odata.type")
	}
	return d.setTyp, nil
}

// eingabeMitglied deutet ein Mitglied der Wurzel (SPEC-181 Nr. 4 und 5).
func eingabeMitglied(info *typInfo, out *model.EntityEingabe, genannt map[string]int, name string, k *knoten) error {
	switch artDesNamens(name) {
	case annotationUngueltig:
		return eingabeFehlerAn(model.EingabeStruktur, name)
	case annotationObjekt:
		return nil
	case annotationProperty:
		if i := strings.IndexByte(name, '@'); name[i+1:] == "odata.bind" {
			return eingabeFehlerAn(model.EingabeBind, name)
		}
		return nil
	}
	if n, ok := info.props[name]; ok {
		p := info.typ.Properties[n]
		if !p.Unterstuetzt() {
			return eingabeFehlerAn(model.EingabePropertyTyp, name)
		}
		return eingabeProperty(info, out, genannt, p, k)
	}
	if _, ok := info.navs[name]; ok {
		return eingabeFehlerAn(model.EingabeNavigation, name)
	}
	if info.typ.Open {
		return eingabeFehlerAn(model.EingabeOffen, name)
	}
	return eingabeFehlerAn(model.EingabeUnbekannteProperty, name)
}

// eingabeProperty liest den Wert einer Property unterstützten Typs (SPEC-181 Nr. 6 und 7 b). Der Marker `[REDACTED]` ist ein
// gewöhnlicher String; ein Schlüssel ist weder `null` noch ein String mit U+0000.
func eingabeProperty(info *typInfo, out *model.EntityEingabe, genannt map[string]int, p model.Property, k *knoten) error {
	_, istSchluessel := info.schluessel[p.Name]
	if istSchluessel && k.art == artNull {
		return eingabeFehlerAn(model.EingabeNull, p.Name)
	}
	gewoehnlich := *k
	gewoehnlich.marker = false
	w, u := lesWert(p, &gewoehnlich)
	if u != "" {
		return eingabeFehlerAn(model.EingabeUrsache(u), p.Name)
	}
	if istSchluessel && w.Art == model.ArtString && strings.Contains(w.Text, "\x00") {
		return eingabeFehlerAn(model.EingabeSchluessel, p.Name)
	}
	genannt[p.Name] = len(out.Properties)
	out.Properties = append(out.Properties, model.PropertyWert{Name: p.Name, Wert: w})
	return nil
}

// erforderlichFehlt liefert die erste erforderliche Property in der Reihenfolge des Modells, die der Body nicht nennt (SPEC-181
// Nr. 7 a): jede Schlüssel-Property und jede nicht nullbare, jeweils unterstützten Typs.
func erforderlichFehlt(info *typInfo, genannt map[string]int) (string, bool) {
	for _, p := range info.typ.Properties {
		if !p.Unterstuetzt() {
			continue
		}
		if _, istSchluessel := info.schluessel[p.Name]; !istSchluessel && p.Nullable {
			continue
		}
		if _, ok := genannt[p.Name]; !ok {
			return p.Name, true
		}
	}
	return "", false
}

// eingabeSchluessel bildet die Literalform des Schlüssels in der Reihenfolge des Modells (SPEC-126 Nr. 11), wenn der Body alle
// Schlüssel-Properties nennt, sonst leer.
func eingabeSchluessel(info *typInfo, out *model.EntityEingabe, genannt map[string]int) string {
	if len(info.typ.Keys) == 0 {
		return ""
	}
	teile := make([]string, 0, len(info.typ.Keys))
	for _, name := range info.typ.Keys {
		i, ok := genannt[name]
		if !ok {
			return ""
		}
		w := out.Properties[i].Wert
		lit := w.Text
		if w.Art == model.ArtString {
			lit = "'" + strings.ReplaceAll(lit, "'", "''") + "'"
		}
		if len(info.typ.Keys) > 1 {
			lit = name + "=" + lit
		}
		teile = append(teile, lit)
	}
	return strings.Join(teile, ",")
}

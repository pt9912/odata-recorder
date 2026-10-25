package baselineimport

import (
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

type bekannterWert struct {
	wert  model.Wert
	erste string
}

// beitrag ist, was ein Exchange zu einer Entity beigetragen hat (SPEC-127 Nr. 11 d): die Namen
// und Werte der bekannten Properties, die er nannte.
type beitrag struct {
	exchangeID string
	typ        string
	werte      map[string]model.Wert
}

type entitaetSchluessel struct {
	set        string
	schluessel string
}

type entitaet struct {
	typ           string
	typErste      string
	props         map[string]bekannterWert
	maskiert      map[string]bool
	uebersprungen map[model.Uebersprungen]bool
	beitraege     []*beitrag
}

// zusammen führt die Beobachtungen einer Session zu Entities zusammen (SPEC-127 Nr. 10).
type zusammen struct {
	entities      map[entitaetSchluessel]*entitaet
	beobachtungen int
}

func neuZusammen() *zusammen {
	return &zusammen{entities: map[entitaetSchluessel]*entitaet{}}
}

// beobachte nimmt eine Beobachtung der Entity (set, e.Schluessel) aus dem Exchange auf; ein
// Widerspruch zu den bisherigen Beobachtungen ist ein *model.KonfliktFehler (Nr. 10 e, f, h).
func (z *zusammen) beobachte(exchangeID, set string, e model.Entity) error {
	z.beobachtungen++
	k := entitaetSchluessel{set, e.Schluessel}
	ent, vorhanden := z.entities[k]
	if !vorhanden {
		ent = &entitaet{
			typ: e.Typ, typErste: exchangeID,
			props: map[string]bekannterWert{}, maskiert: map[string]bool{}, uebersprungen: map[model.Uebersprungen]bool{},
		}
		z.entities[k] = ent
	}
	if e.Typ != ent.typ {
		return &model.KonfliktFehler{Ursache: model.KonfliktTyp, EntitySet: set, Erste: ent.typErste, Zweite: exchangeID}
	}
	b := &beitrag{exchangeID: exchangeID, typ: e.Typ, werte: map[string]model.Wert{}}
	if n := len(ent.beitraege); n > 0 && ent.beitraege[n-1].exchangeID == exchangeID {
		b = ent.beitraege[n-1]
	} else {
		ent.beitraege = append(ent.beitraege, b)
	}
	for _, p := range e.Properties {
		if p.Wert.Art == model.ArtRedigiert {
			ent.maskiert[p.Name] = true
			continue
		}
		if alt, ok := ent.props[p.Name]; ok {
			if alt.wert != p.Wert {
				return &model.KonfliktFehler{Ursache: model.KonfliktWert, EntitySet: set, Property: p.Name, Erste: alt.erste, Zweite: exchangeID}
			}
		} else {
			ent.props[p.Name] = bekannterWert{wert: p.Wert, erste: exchangeID}
		}
		b.werte[p.Name] = p.Wert
	}
	for _, u := range e.Uebersprungen {
		ent.uebersprungen[u] = true
	}
	return nil
}

// payloadWert ist der Wert einer Property im Payload nach seiner Art (Nr. 11 a).
func payloadWert(w model.Wert) any {
	switch w.Art {
	case model.ArtNull:
		return nil
	case model.ArtBoolean:
		return w.Text == "true"
	case model.ArtInt32:
		return roh(w.Text)
	}
	return w.Text
}

func payloadObjekt(werte map[string]model.Wert, typ, setTyp string) objekt {
	o := objekt{}
	for n, w := range werte {
		o[n] = payloadWert(w)
	}
	if typ != setTyp {
		o["@odata.type"] = "#" + typ
	}
	return o
}

// fertig ist eine Entity der Baseline in der Form des Manifests (Nr. 14 a).
type fertig struct {
	set        string
	schluessel string
	json       objekt
}

// fertige bildet die Entities der Baseline, aufsteigend nach Set und Schlüssel (Bytes).
func (z *zusammen) fertige(m model.ServiceModel, sessionID string) []fertig {
	keys := make([]entitaetSchluessel, 0, len(z.entities))
	for k := range z.entities {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b entitaetSchluessel) int {
		if c := strings.Compare(a.set, b.set); c != 0 {
			return c
		}
		return strings.Compare(a.schluessel, b.schluessel)
	})
	out := make([]fertig, 0, len(keys))
	for _, k := range keys {
		ent := z.entities[k]
		setTyp := ""
		if s, ok := m.EntitySet(k.set); ok {
			setTyp = s.EntityType
		}
		werte := map[string]model.Wert{}
		namen := make([]string, 0, len(ent.props))
		for n, p := range ent.props {
			werte[n] = p.wert
			namen = append(namen, n)
		}
		slices.Sort(namen)
		prov := make([]any, 0, len(ent.beitraege))
		for _, b := range ent.beitraege {
			pn := make([]string, 0, len(b.werte))
			for n := range b.werte {
				pn = append(pn, n)
			}
			slices.Sort(pn)
			prov = append(prov, objekt{
				"session_id":  sessionID,
				"exchange_id": b.exchangeID,
				"projection":  liste(pn, func(s string) any { return s }),
				"hash":        sha256Hex(kanon(payloadObjekt(b.werte, b.typ, setTyp))),
			})
		}
		out = append(out, fertig{set: k.set, schluessel: k.schluessel, json: objekt{
			"completeness":        vollstaendigkeit(m, ent),
			"entity_set":          k.set,
			"payload":             payloadObjekt(werte, ent.typ, setTyp),
			"provenance":          prov,
			"selected_properties": liste(namen, func(s string) any { return s }),
			"typed_key":           k.schluessel,
		}})
	}
	return out
}

// vollstaendigkeit ist complete, wenn jede unterstützte Property des Typs bekannt ist (Nr. 11 c).
func vollstaendigkeit(m model.ServiceModel, ent *entitaet) string {
	t, ok := m.EntityType(ent.typ)
	if !ok {
		return "partial"
	}
	for _, p := range t.Properties {
		if _, bekannt := ent.props[p.Name]; p.Unterstuetzt() && !bekannt {
			return "partial"
		}
	}
	return "complete"
}

// maskierte zählt die Paare (Entity, Property), die in einer Beobachtung redigiert und in keiner
// bekannt waren (Nr. 14 b).
func (z *zusammen) maskierte() int {
	n := 0
	for _, ent := range z.entities {
		for p := range ent.maskiert {
			if _, bekannt := ent.props[p]; !bekannt {
				n++
			}
		}
	}
	return n
}

// nichtUnterstuetzt zählt je (Set, Name, Grund) die verschiedenen Entities, die das Mitglied
// übersprangen, aufsteigend (Nr. 14 b).
func (z *zusammen) nichtUnterstuetzt() []any {
	type zeile struct{ set, name, grund string }
	zaehler := map[zeile]int{}
	for k, ent := range z.entities {
		for u := range ent.uebersprungen {
			zaehler[zeile{k.set, u.Name, string(u.Grund)}]++
		}
	}
	zeilen := make([]zeile, 0, len(zaehler))
	for ze := range zaehler {
		zeilen = append(zeilen, ze)
	}
	slices.SortFunc(zeilen, func(a, b zeile) int {
		if c := strings.Compare(a.set, b.set); c != 0 {
			return c
		}
		if c := strings.Compare(a.name, b.name); c != 0 {
			return c
		}
		return strings.Compare(a.grund, b.grund)
	})
	return liste(zeilen, func(ze zeile) any {
		return objekt{"entities": zaehler[ze], "entity_set": ze.set, "name": ze.name, "reason": ze.grund}
	})
}

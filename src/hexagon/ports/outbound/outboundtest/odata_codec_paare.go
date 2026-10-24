package outboundtest

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Plätze der Stellen S1 bis S15 (Zusatztabelle in SPEC-126 Nr. 13): ein Mitglied der Wurzel einer Collection, das Ende der
// Wurzel, ein Mitglied oder Element eines Entity-Objekts, das Ende eines Entity-Objekts.
const (
	platzR = iota
	platzRE
	platzE
	platzEE
)

// hStelle ist ein Verstoß des Deutens am Modell P.
type hStelle struct {
	id      int
	platz   int
	ursache model.DatenUrsache
	ort     string // der Pointer in der Entity oder an der Wurzel; bei S7 leer (das Element selbst)
}

func hStellen() []hStelle {
	return []hStelle{
		{1, platzR, model.DatenKontext, "/@odata.context"}, {2, platzRE, model.DatenKontext, "/@odata.context"},
		{3, platzR, model.DatenWert, "/@odata.count"}, {4, platzR, model.DatenTyp, "/@odata.nextLink"},
		{5, platzR, model.DatenStruktur, "/other"}, {6, platzRE, model.DatenStruktur, "/value"},
		{7, platzE, model.DatenStruktur, ""}, {8, platzE, model.DatenODataTyp, "/@odata.type"},
		{9, platzE, model.DatenUnbekannteProperty, "/Nope"}, {10, platzE, model.DatenTyp, "/Note"},
		{11, platzE, model.DatenWert, "/Shipped"}, {12, platzE, model.DatenNull, "/Req"},
		{13, platzE, model.DatenSchluessel, "/Part"}, {14, platzEE, model.DatenSchluessel, "/Seq"},
		{15, platzE, model.DatenZahlZuLang, "/Amount"},
	}
}

// paarWerte sind die JSON-Werte des Modells P: ein gültiger Teil Part des Schlüssels (ein Guid) und ein ungültiger Wert für
// Shipped (ein Datum, das der Kalender nicht kennt).
func paarWerte() (partGut, shippedSchlecht string) { return `"` + guidKlein + `"`, `"2023-02-29"` }

// hItem ist ein Mitglied eines Objekts im Body; stelle ist die Stelle, deren Verstoß es trägt (0: keiner).
type hItem struct {
	text   string
	stelle int
}

// hWert ist der Platzhalter des Mitglieds `value` in der Wurzel einer Collection.
const hWert = "\x00value"

// hElement ist ein Entity-Objekt: bei einer Entity die Wurzel, bei einer Collection ein Element von `value`.
type hElement struct {
	null      bool
	items     []hItem
	typStelle int
	endStelle int
}

// hDok ist ein Body aus Mitgliedern und Elementen mit den Stellen, die er verletzt.
type hDok struct {
	einzel   bool
	root     []hItem
	elemente []hElement
	ctxFehlt bool
	valFehlt bool
	junk     string
}

func hBasisElement() hElement {
	part, _ := paarWerte()
	return hElement{items: []hItem{{`"Part":` + part, 0}, {`"Seq":2`, 0}, {`"Req":3`, 0}}}
}

func hBasisKol() hDok {
	d := hDok{root: []hItem{{`"@odata.context":"$metadata#Ps"`, 0}, {hWert, 0}}}
	for range 3 {
		d.elemente = append(d.elemente, hBasisElement())
	}
	return d
}

func hBasisEin() hDok {
	e := hBasisElement()
	e.items = append([]hItem{{`"@odata.context":"$metadata#Ps/$entity"`, 0}}, e.items...)
	return hDok{einzel: true, elemente: []hElement{e}}
}

func (d hDok) klon() hDok {
	d.root = slices.Clone(d.root)
	d.elemente = slices.Clone(d.elemente)
	for i := range d.elemente {
		d.elemente[i].items = slices.Clone(d.elemente[i].items)
	}
	return d
}

// element liefert das Element e (bei der Entity das eine Objekt).
func (d *hDok) element(e int) *hElement {
	if d.einzel {
		return &d.elemente[0]
	}
	return &d.elemente[e]
}

// mitglieder liefert die Liste, in der der Kontext steht: die Wurzel der Collection oder das Objekt der Entity.
func (d *hDok) mitglieder() *[]hItem {
	if d.einzel {
		return &d.elemente[0].items
	}
	return &d.root
}

func indexVon(items []hItem, praefix string) int {
	return slices.IndexFunc(items, func(x hItem) bool { return strings.HasPrefix(x.text, praefix) })
}

// zusatzWurzel fügt ein Mitglied der Wurzel einer Collection ein: wo ist 0 vor dem Kontext, 1 zwischen Kontext und `value`,
// 2 hinter `value`.
func (d *hDok) zusatzWurzel(wo int, it hItem) {
	switch wo {
	case 0:
		d.root = slices.Insert(d.root, 0, it)
	case 1:
		d.root = slices.Insert(d.root, indexVon(d.root, hWert), it)
	default:
		d.root = append(d.root, it)
	}
}

// zusatz fügt ein Mitglied in das Objekt e ein: wo ist 0 ganz vorn, 1 vor den Mitgliedern des Basis-Objekts (bei der
// Entity hinter dem Kontext), 2 hinter ihnen.
func (d *hDok) zusatz(e, wo int, it hItem) {
	el := d.element(e)
	switch {
	case wo == 0:
		el.items = slices.Insert(el.items, 0, it)
	case wo == 1 && d.einzel:
		el.items = slices.Insert(el.items, 1, it)
	case wo == 1:
		el.items = slices.Insert(el.items, 0, it)
	default:
		el.items = append(el.items, it)
	}
}

// ersetze setzt das Mitglied mit dem Namen durch ein anderes; entferne löscht es.
func (d *hDok) ersetze(e int, name string, it hItem) {
	el := d.element(e)
	el.items[indexVon(el.items, `"`+name+`"`)] = it
}

func (d *hDok) entferne(e int, name string) {
	el := d.element(e)
	el.items = slices.Delete(el.items, indexVon(el.items, `"`+name+`"`), indexVon(el.items, `"`+name+`"`)+1)
}

// kontextVerletzen setzt den Kontext auf ein anderes Set (S1); wo 2 stellt ihn ans Ende seiner Liste.
func (d *hDok) kontextVerletzen(wo int) {
	items := d.mitglieder()
	i := indexVon(*items, `"@odata.context"`)
	neu := hItem{`"@odata.context":"$metadata#Other"`, 1}
	if d.einzel {
		neu.text = `"@odata.context":"$metadata#Other/$entity"`
	}
	(*items)[i] = neu
	if wo == 2 {
		*items = append(slices.Delete(*items, i, i+1), neu)
	}
}

// setzeStelle setzt den Verstoß der Stelle in den Body: e ist das Element, wo der Ort eines Zusatzes (zusatzWurzel, zusatz).
func (d *hDok) setzeStelle(id, e, wo int) {
	switch id {
	case 1:
		d.kontextVerletzen(wo)
	case 2:
		d.ctxFehlt = true
		items := d.mitglieder()
		*items = slices.Delete(*items, indexVon(*items, `"@odata.context"`), indexVon(*items, `"@odata.context"`)+1)
	case 3, 4, 5:
		d.zusatzWurzel(wo, hItem{hZusaetze()[id], id})
	case 6:
		d.valFehlt, d.elemente = true, nil
		d.root = slices.Delete(d.root, indexVon(d.root, hWert), indexVon(d.root, hWert)+1)
	case 7:
		d.elemente[e].null = true
	case 12:
		d.ersetze(e, "Req", hItem{`"Req":null`, 12})
	case 13:
		d.ersetze(e, "Part", hItem{`"Part":null`, 13})
	case 14:
		d.entferne(e, "Seq")
		d.element(e).endStelle = 14
	default:
		if id == 8 {
			d.element(e).typStelle = 8
		}
		d.zusatz(e, wo, hItem{hZusaetze()[id], id})
	}
}

// hZusaetze sind die Mitglieder, die eine Stelle zum Body hinzufügt.
func hZusaetze() map[int]string {
	_, shipped := paarWerte()
	return map[int]string{
		3: `"@odata.count":-1`, 4: `"@odata.nextLink":5`, 5: `"other":1`, 8: `"@odata.type":"#T.Nope"`, 9: `"Nope":1`, 10: `"Note":5`,
		11: `"Shipped":` + shipped, 15: `"Amount":1` + strings.Repeat("0", 1024),
	}
}

// text liefert den Body.
func (d hDok) text() []byte {
	obj := func(items []hItem) string {
		var ms []string
		for _, it := range items {
			ms = append(ms, it.text)
		}
		return strings.Join(ms, ",")
	}
	if d.einzel {
		return []byte("{" + obj(d.elemente[0].items) + "}" + d.junk)
	}
	var ms []string
	for _, it := range d.root {
		if it.text != hWert {
			ms = append(ms, it.text)
			continue
		}
		var es []string
		for _, e := range d.elemente {
			if e.null {
				es = append(es, "null")
			} else {
				es = append(es, "{"+obj(e.items)+"}")
			}
		}
		ms = append(ms, `"value":[`+strings.Join(es, ",")+`]`)
	}
	return []byte("{" + strings.Join(ms, ",") + "}" + d.junk)
}

// objektStelle ist die Stelle, die das Objekt vor seinem Ende trifft: das Element `null`, `@odata.type`, dann die
// Mitglieder in der Reihenfolge des Dokuments.
func objektStelle(el hElement) int {
	switch {
	case el.null:
		return 7
	case el.typStelle != 0:
		return el.typStelle
	}
	for _, it := range el.items {
		if it.stelle != 0 {
			return it.stelle
		}
	}
	return 0
}

// erwarteteStelle geht den Body in der Reihenfolge von SPEC-126 Nr. 13 c durch: die Mitglieder der Wurzel in der Reihenfolge
// des Dokuments, `value` an seiner Stelle (jedes Element von Anfang bis Ende), danach das Ende der Wurzel.
func (d hDok) erwarteteStelle() (stelle, element int) {
	if d.einzel {
		el := d.elemente[0]
		switch {
		case objektStelle(el) != 0:
			return objektStelle(el), 0
		case d.ctxFehlt:
			return 2, 0
		}
		return el.endStelle, 0
	}
	for _, it := range d.root {
		if it.text != hWert {
			if it.stelle != 0 {
				return it.stelle, 0
			}
			continue
		}
		for e, el := range d.elemente {
			if s := objektStelle(el); s != 0 {
				return s, e
			}
			if el.endStelle != 0 {
				return el.endStelle, e
			}
		}
	}
	switch {
	case d.ctxFehlt:
		return 2, 0
	case d.valFehlt:
		return 6, 0
	}
	return 0, 0
}

// hFall bildet den Fall zum Body: der Fehler ist die Stelle, die der Ablauf zuerst trifft.
func hFall(name string, d hDok) jfall {
	s, e := d.erwarteteStelle()
	st := hStellen()[s-1]
	ort := st.ort
	if !d.einzel && st.platz != platzR && s != 2 && s != 6 {
		ort = "/value/" + strconv.Itoa(e) + st.ort
	}
	return jfall{name: name, m: "P", set: "Ps", einzel: d.einzel, body: d.text(), ursache: st.ursache, ort: ort}
}

// hLage setzt die Stelle id an den Ort wo des Elements e.
type hLage struct{ id, e, wo int }

func hPaar(name string, basis hDok, lagen ...hLage) jfall {
	d := basis.klon()
	for _, l := range lagen {
		d.setzeStelle(l.id, l.e, l.wo)
	}
	return hFall(name, d)
}

func gehoertZu(id, platz int) bool { return hStellen()[id-1].platz == platz }

func idsMit(platz int) []int {
	var out []int
	for _, s := range hStellen() {
		if s.platz == platz {
			out = append(out, s.id)
		}
	}
	return out
}

// hMoeglich meldet, ob beide Stellen zugleich in einem Body stehen können: der Kontext ist nicht zugleich falsch und
// fehlend, und ohne `value` gibt es kein Element.
func hMoeglich(a, b int) bool {
	if (a == 1 && b == 2) || (a == 2 && b == 1) {
		return false
	}
	for _, p := range [][2]int{{a, b}, {b, a}} {
		if p[0] == 6 && (gehoertZu(p[1], platzE) || gehoertZu(p[1], platzEE)) {
			return false
		}
	}
	return true
}

// hPaareKol sind die Paare der Collection (Platzierungen nach der Tabelle in §6 des Plans).
func hPaareKol() []jfall {
	var out []jfall
	basis := hBasisKol()
	name := func(a, b int, s string) string { return fmt.Sprintf("H2 S%d S%d %s", a, b, s) }
	rs := idsMit(platzR)
	for i, a := range rs { // R–R
		for _, b := range rs[i+1:] {
			if a == 1 {
				out = append(out, hPaar(name(a, b, "b vor a"), basis, hLage{1, 0, 0}, hLage{b, 0, 0}), hPaar(name(a, b, "a vor b"), basis, hLage{1, 0, 0}, hLage{b, 0, 1}))
				continue
			}
			out = append(out, hPaar(name(a, b, "a vor b"), basis, hLage{a, 0, 1}, hLage{b, 0, 1}), hPaar(name(a, b, "b vor a"), basis, hLage{b, 0, 1}, hLage{a, 0, 1}))
		}
	}
	for _, r := range rs {
		for _, re := range idsMit(platzRE) { // R–RE
			if hMoeglich(r, re) {
				out = append(out, hPaar(name(r, re, "RE am Ende"), basis, hLage{r, 0, 1}, hLage{re, 0, 0}))
			}
		}
		for _, e := range append(idsMit(platzE), idsMit(platzEE)...) { // R–E und R–EE
			for wo := 1; wo <= 2; wo++ {
				out = append(out, hPaar(name(r, e, fmt.Sprintf("R an %d", wo)), basis, hLage{r, 0, wo}, hLage{e, 1, 2}))
			}
		}
	}
	out = append(out, hPaar(name(2, 6, "RE–RE"), basis, hLage{2, 0, 0}, hLage{6, 0, 0}))
	for _, e := range append(idsMit(platzE), idsMit(platzEE)...) { // RE–E und RE–EE: ohne `value` gibt es kein Element
		out = append(out, hPaar(name(2, e, "RE am Ende"), basis, hLage{e, 1, 2}, hLage{2, 0, 0}))
	}
	return append(out, hPaareElemente(basis, name)...)
}

// hPaareElemente sind die Paare E–E und E–EE: in zwei Elementen in beiden Reihenfolgen, im selben Element in den Reihenfolgen,
// die die Stellen zulassen.
func hPaareElemente(basis hDok, name func(a, b int, s string) string) []jfall {
	var out []jfall
	es := idsMit(platzE)
	zwei := func(a, b int) []jfall {
		return []jfall{
			hPaar(name(a, b, "a im ersten, b im letzten Element"), basis, hLage{a, 0, 2}, hLage{b, 2, 2}),
			hPaar(name(a, b, "b im ersten, a im letzten Element"), basis, hLage{b, 0, 2}, hLage{a, 2, 2}),
		}
	}
	for i, a := range es {
		for _, b := range es[i+1:] {
			out = append(out, zwei(a, b)...)
			switch a {
			case 7:
			case 8:
				out = append(out, hPaar(name(a, b, "selbes Element, S8 gilt zuerst"), basis, hLage{b, 1, 2}, hLage{8, 1, 2}))
			default:
				out = append(out, selbesElement(name, basis, a, b)...)
			}
		}
	}
	for _, e := range es { // E–EE
		out = append(out, zwei(e, 14)...)
		if e != 7 {
			out = append(out, hPaar(name(e, 14, "selbes Element, S14 gilt zuletzt"), basis, hLage{e, 1, 2}, hLage{14, 1, 2}))
		}
	}
	return out
}

// selbesElement liefert die Reihenfolgen zweier Stellen im Element 1: zwei Zusätze in beiden Reihenfolgen, ein Zusatz vor
// und hinter dem Basis-Objekt, wenn eine der Stellen ein Mitglied ersetzt (S12, S13), und die Basis umgekehrt für beide.
func selbesElement(name func(a, b int, s string) string, basis hDok, a, b int) []jfall {
	ersetzt := func(id int) bool { return id == 12 || id == 13 }
	switch {
	case ersetzt(a) && ersetzt(b):
		umgekehrt := basis.klon()
		slices.Reverse(umgekehrt.elemente[1].items)
		return []jfall{
			hPaar(name(a, b, "Basis wie im Modell"), basis, hLage{a, 1, 2}, hLage{b, 1, 2}),
			hPaar(name(a, b, "Basis umgekehrt"), umgekehrt, hLage{a, 1, 2}, hLage{b, 1, 2}),
		}
	case ersetzt(a) || ersetzt(b):
		r, z := a, b
		if ersetzt(b) {
			r, z = b, a
		}
		return []jfall{
			hPaar(name(a, b, "Zusatz vor dem Basis-Objekt"), basis, hLage{r, 1, 2}, hLage{z, 1, 1}),
			hPaar(name(a, b, "Zusatz hinter dem Basis-Objekt"), basis, hLage{r, 1, 2}, hLage{z, 1, 2}),
		}
	}
	return []jfall{
		hPaar(name(a, b, "a vor b"), basis, hLage{a, 1, 2}, hLage{b, 1, 2}),
		hPaar(name(a, b, "b vor a"), basis, hLage{b, 1, 2}, hLage{a, 1, 2}),
	}
}

// hPaareEin sind die Paare der Entity: M–M, M–N und N–N.
func hPaareEin() []jfall {
	var out []jfall
	basis := hBasisEin()
	name := func(a, b int, s string) string { return fmt.Sprintf("H2 Entity S%d S%d %s", a, b, s) }
	ms := []int{1, 8, 9, 10, 11, 12, 13, 15}
	for i, a := range ms {
		for _, b := range ms[i+1:] {
			out = append(out, hPaarMM(name, basis, a, b)...)
		}
	}
	for _, m := range ms {
		for _, n := range []int{2, 14} {
			if hMoeglich(m, n) {
				out = append(out, hPaar(name(m, n, "N am Ende"), basis, hLage{m, 0, 2}, hLage{n, 0, 0}))
			}
		}
	}
	return append(out, hPaar(name(2, 14, "N–N"), basis, hLage{2, 0, 0}, hLage{14, 0, 0}))
}

// hPaarMM sind die Platzierungen zweier Mitglieder der Entity: S8 steht zuletzt im Body (der Ablauf trifft es zuerst),
// der Kontext (S1) steht vorn oder anders als das andere Mitglied, sonst beide Reihenfolgen.
func hPaarMM(name func(a, b int, s string) string, basis hDok, a, b int) []jfall {
	ersetzt := func(id int) bool { return id == 1 || id == 12 || id == 13 }
	switch {
	case a == 8 || b == 8:
		x := a
		if a == 8 {
			x = b
		}
		return []jfall{hPaar(name(a, b, "S8 zuletzt im Body"), basis, hLage{x, 0, 2}, hLage{8, 0, 2})}
	case a == 1:
		lagen := [][]hLage{{{1, 0, 1}, {b, 0, 0}}, {{1, 0, 1}, {b, 0, 2}}}
		if ersetzt(b) {
			lagen = [][]hLage{{{1, 0, 1}, {b, 0, 1}}, {{1, 0, 2}, {b, 0, 1}}}
		}
		return []jfall{hPaar(name(a, b, "Kontext vorn"), basis, lagen[0]...), hPaar(name(a, b, "Kontext anders"), basis, lagen[1]...)}
	}
	return selbesElementEin(name, basis, a, b)
}

func selbesElementEin(name func(a, b int, s string) string, basis hDok, a, b int) []jfall {
	ersetzt := func(id int) bool { return id == 12 || id == 13 }
	switch {
	case ersetzt(a) && ersetzt(b):
		umgekehrt := basis.klon()
		slices.Reverse(umgekehrt.elemente[0].items[1:])
		return []jfall{
			hPaar(name(a, b, "Basis wie im Modell"), basis, hLage{a, 0, 2}, hLage{b, 0, 2}),
			hPaar(name(a, b, "Basis umgekehrt"), umgekehrt, hLage{a, 0, 2}, hLage{b, 0, 2}),
		}
	case ersetzt(a) || ersetzt(b):
		r, z := a, b
		if ersetzt(b) {
			r, z = b, a
		}
		return []jfall{
			hPaar(name(a, b, "Zusatz vor dem Basis-Objekt"), basis, hLage{r, 0, 2}, hLage{z, 0, 1}),
			hPaar(name(a, b, "Zusatz hinter dem Basis-Objekt"), basis, hLage{r, 0, 2}, hLage{z, 0, 2}),
		}
	}
	return []jfall{
		hPaar(name(a, b, "a vor b"), basis, hLage{a, 0, 2}, hLage{b, 0, 2}),
		hPaar(name(a, b, "b vor a"), basis, hLage{b, 0, 2}, hLage{a, 0, 2}),
	}
}

// h2Faelle sind die Fälle des Teils 2 der Zusatztabelle.
func h2Faelle() []jfall { return append(hPaareKol(), hPaareEin()...) }

func jH2(t *testing.T, s *jsuite) {
	fs := h2Faelle()
	if len(fs) != h2Zahl {
		t.Errorf("H2: %d Fälle, erwartet %d (226 der Collection und 66 der Entity abzüglich 11, die kein Body ausdrückt)", len(fs), h2Zahl)
	}
	s.pruefeAlle(t, fs)
}

// h2Zahl ist die Zahl der Fälle: 292 der Zusatztabelle, abzüglich der elf Paare, die kein Body ausdrücken kann (S1 und S2 des
// Kontexts, S6 ohne `value` mit einem Element oder dessen Ende: zehn in der Collection, eines in der Entity).
const h2Zahl = 281

// Fragmente je Verstoß des Lesens, eingesetzt als Wert eines Mitglieds: der Verstoß liegt innerhalb des Fragments.
func lesenFragmente() []struct {
	name string
	ur   model.DatenUrsache
	frag string
} {
	return []struct {
		name string
		ur   model.DatenUrsache
		frag string
	}{
		{"syntax", model.DatenSyntax, `[1 2]`}, {"kodierung", model.DatenKodierung, `"\uD800"`},
		{"tiefe", model.DatenTiefe, klammern('a', 600)}, {"doppelter_name", model.DatenDoppelterName, `{"a":1,"a":2}`},
	}
}

func faelleH1Paare() []jfall {
	var out []jfall
	fr := lesenFragmente()
	kol := func(a, b string) []byte {
		return []byte(`{"@odata.context":"$metadata#Products","value":[{"ID":1,"Weight":` + a + `},{"ID":2,"Weight":` + b + `}]}`)
	}
	for _, x := range fr {
		for _, y := range fr {
			if x.ur == y.ur {
				continue
			}
			out = append(out, jfall{name: fmt.Sprintf("H1 %s vor %s", x.name, y.name), m: "R", set: "Products", body: kol(x.frag, y.frag), ursache: x.ur})
		}
		einzel := []byte(`{"@odata.context":"$metadata#Products","value":[{"ID":1,"Weight":` + x.frag + `}]}`)
		vorn := strings.Index(string(einzel), x.frag)
		out = append(out,
			jfall{name: "H1 " + x.name + " vor dem Byte grenze+1", m: "R", set: "Products", body: einzel, grenze: int64(len(einzel)) - 1, ursache: x.ur},
			jfall{name: "H1 " + x.name + " hinter dem Byte grenze+1", m: "R", set: "Products", body: einzel, grenze: int64(vorn), ursache: model.DatenZuGross})
	}
	return out
}

// lesenInEinem setzt in den Body einer Stelle ein Fragment des Lesens: vor oder hinter der Stelle (lage 0 oder 1). Am Ende der
// Wurzel (RE) und am Ende der Entity (N) liegt „hinter ihr“ nach dem schließenden `}`; dort kann nur Müll stehen, der selbst `syntax` ist.
func lesenInEinem(id int, einzel bool, lage int, frag string) (hDok, bool) {
	d := hBasisKol()
	if einzel {
		d = hBasisEin()
	}
	st := hStellen()[id-1]
	wo := 2
	if id == 1 || st.platz == platzR {
		wo = 1
	}
	d.setzeStelle(id, 1, wo)
	extra := hItem{`"Weight":` + frag, 0}
	switch {
	case einzel && (id == 2 || id == 14), !einzel && st.platz == platzRE:
		if lage == 0 {
			if einzel {
				d.zusatz(0, 2, extra)
			} else {
				d.zusatzWurzel(2, extra)
			}
			return d, false
		}
		d.junk = " " + frag
		return d, true
	case einzel:
		d.zusatz(0, []int{0, 2}[lage], extra)
	case st.platz == platzR:
		d.zusatzWurzel([]int{0, 2}[lage], extra)
	default:
		e := []int{0, 2}[lage]
		if st.platz == platzEE && lage == 0 {
			d.zusatz(1, 2, extra)
			return d, false
		}
		d.elemente[e].items = []hItem{extra}
	}
	return d, false
}

// h1FaelleEiner sind die Fälle von H1 für eine Stelle: jeder Verstoß des Lesens vor und hinter ihr, und `zu_gross` mit dem
// Byte grenze+1 hinter und vor ihr.
func h1FaelleEiner(id int, einzel bool) []jfall {
	var out []jfall
	for _, f := range lesenFragmente() {
		for lage := range 2 {
			d, junk := lesenInEinem(id, einzel, lage, f.frag)
			fall := hFall(fmt.Sprintf("H1 %s mit S%d (einzel %v, Lage %d)", f.name, id, einzel, lage), d)
			fall.ursache, fall.ort = f.ur, ""
			if junk {
				fall.ursache = model.DatenSyntax
			}
			out = append(out, fall)
		}
	}
	ohne := hBasisKol()
	if einzel {
		ohne = hBasisEin()
	}
	ohne.setzeStelle(id, 1, 2)
	body := ohne.text()
	for _, g := range []int64{int64(len(body)) - 1, 5} {
		out = append(out, jfall{name: fmt.Sprintf("H1 zu_gross mit S%d (einzel %v, grenze %d)", id, einzel, g), m: "P", set: "Ps", einzel: einzel,
			body: body, grenze: g, ursache: model.DatenZuGross})
	}
	return out
}

func faelleH1Deuten() []jfall {
	var out []jfall
	for _, einzel := range []bool{false, true} {
		for _, st := range hStellen() {
			if einzel && st.id >= 3 && st.id <= 7 {
				continue
			}
			out = append(out, h1FaelleEiner(st.id, einzel)...)
		}
	}
	return out
}

// h1Zahl ist die Zahl der Fälle von H1: 20 im Lesen und 5 × (15 + 10) × 2 im Lesen gegen Deuten.
const h1Zahl = 270

func h1Faelle() []jfall { return append(faelleH1Paare(), faelleH1Deuten()...) }

func jH1(t *testing.T, s *jsuite) {
	fs := h1Faelle()
	if len(fs) != h1Zahl {
		t.Errorf("H1: %d Fälle, erwartet %d", len(fs), h1Zahl)
	}
	s.pruefeAlle(t, fs)
}

package model_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Das Fixture-Modell von PruefeExpand (SPEC-263): achtzehn Typen mit Basis-Ketten bis zur dritten Stufe, mindestens einem
// Typ zu jeder der acht Kombinationen aus offen, abstrakt und abgeleitet, Typen ohne Navigation und ohne Member, einer
// Property mit nicht unterstütztem Typ in jedem Zweig, je einem Typ mit 32 Navigationen (offen und nicht offen) und dem
// Zeichen `é` in einer Navigation und in einer Property. Die Listen eines Typs stehen mit den geerbten Membern zuerst und
// in einer Folge, die nie aufsteigend nach Namen sortiert ist; `Keys` nennt die erste Property und einen Namen, den keine
// Liste führt.

// Rohbytes (AGENTS.md §3.18): `é` in der Form NFC und NFD entsteht aus Bytewerten.
func exvEAkut() string    { return string([]byte{0xc3, 0xa9}) }
func exvEAkutNFD() string { return string([]byte{0x65, 0xcc, 0x81}) }

type exvSpez struct {
	name, basis     string
	abstrakt, offen bool
	props           []model.Property
	navs            []model.NavigationProperty
}

func exvP(name, edm string) model.Property {
	return model.Property{Name: name, EdmType: edm, Nullable: true}
}

func exvN(name, ziel string, sammlung bool) model.NavigationProperty {
	return model.NavigationProperty{Name: name, Target: ziel, Collection: sammlung}
}

func exvSpezen() []exvSpez {
	viele := make([]model.NavigationProperty, 0, 32)
	for i := 32; i >= 1; i-- {
		viele = append(viele, exvN(fmt.Sprintf("Nav%02d", i), "Ns.Kategorie", i%2 == 0))
	}
	vp := []model.Property{exvP("VId", "Edm.Int32"), exvP("VName", "Edm.String"), exvP("VWert", "Edm.Decimal"), exvP("VBlob", "Edm.Binary")}
	return []exvSpez{
		{name: "Ns.Basis",
			props: []model.Property{exvP("Id", "Edm.Int32"), exvP("Titel", "Edm.String"), exvP("Daten", "Edm.Binary")},
			navs:  []model.NavigationProperty{exvN("Positionen", "Ns.Position", true), exvN("Kategorie", "Ns.Kategorie", false)}},
		{name: "Ns.Abgeleitet", basis: "Ns.Basis",
			props: []model.Property{exvP("Zusatz", "Edm.String"), exvP("Dauer", "Edm.Duration")},
			navs:  []model.NavigationProperty{exvN("Lieferant", "Ns.Lieferant", false), exvN("Caf"+exvEAkut(), "Ns.Lieferant", true)}},
		{name: "Ns.Stufe3", basis: "Ns.Abgeleitet",
			props: []model.Property{exvP("Tiefe", "Edm.Int32")},
			navs:  []model.NavigationProperty{exvN("Archiv", "Ns.Kategorie", false)}},
		{name: "Ns.Abstrakt", abstrakt: true,
			props: []model.Property{exvP("AKey", "Edm.Guid"), exvP("ABeschr", "Edm.String")},
			navs:  []model.NavigationProperty{exvN("Besitzer", "Ns.Lieferant", false)}},
		{name: "Ns.Konkret", basis: "Ns.Abstrakt",
			props: []model.Property{exvP("KName", "Edm.String")},
			navs:  []model.NavigationProperty{exvN("AKNav", "Ns.Position", true)}},
		{name: "Ns.Offen", offen: true,
			props: []model.Property{exvP("OId", "Edm.Int32"), exvP("OText", "Edm.String"), exvP("OBlob", "Edm.Stream")},
			navs:  []model.NavigationProperty{exvN("ONav", "Ns.Kategorie", false), exvN("OMehr", "Ns.Position", true)}},
		{name: "Ns.OffenKind", basis: "Ns.Offen", offen: true,
			props: []model.Property{exvP("OKind", "Edm.Int32")},
			navs:  []model.NavigationProperty{exvN("OKindNav", "Ns.Basis", false), exvN("OKindNav2", "Ns.Basis", true)}},
		{name: "Ns.Kategorie",
			props: []model.Property{exvP("KatTitel", "Edm.String"), exvP("KatId", "Edm.Int32")},
			navs:  []model.NavigationProperty{exvN("Produkte", "Ns.Basis", true)}},
		{name: "Ns.AbstraktKind", basis: "Ns.Abstrakt", abstrakt: true,
			props: []model.Property{exvP("AKProp", "Edm.Int32")},
			navs:  []model.NavigationProperty{exvN("AKNav2", "Ns.Kategorie", false)}},
		{name: "Ns.OffenAbstrakt", abstrakt: true, offen: true,
			props: []model.Property{exvP("XText", "Edm.String"), exvP("XId", "Edm.Int32")},
			navs:  []model.NavigationProperty{exvN("XNav", "Ns.Kategorie", false), exvN("XMehr", "Ns.Position", true)}},
		{name: "Ns.OffenAbstraktKind", basis: "Ns.OffenAbstrakt", abstrakt: true, offen: true,
			props: []model.Property{exvP("XKind", "Edm.Int32")},
			navs:  []model.NavigationProperty{exvN("XKindNav", "Ns.Basis", false)}},
		{name: "Ns.Lieferant",
			props: []model.Property{exvP("LId", "Edm.Int32"), exvP("Firma", "Edm.String"), exvP("Pr"+exvEAkut()+"fer", "Edm.String")},
			navs:  []model.NavigationProperty{exvN("Lieferungen", "Ns.Position", true)}},
		{name: "Ns.Position",
			props: []model.Property{exvP("PosId", "Edm.Int32"), exvP("Menge", "Edm.Decimal")},
			navs:  []model.NavigationProperty{exvN("Artikel", "Ns.Basis", false)}},
		{name: "Ns.OhneNav", props: []model.Property{exvP("Zahl", "Edm.Int64"), exvP("Notiz", "Edm.String")}},
		{name: "Ns.Leer"},
		{name: "Ns.LeerOffen", offen: true},
		{name: "Ns.Viele", props: vp, navs: viele},
		{name: "Ns.VieleOffen", offen: true, props: vp, navs: viele},
	}
}

// exvT ist ein Typ des Fixture-Modells: typ mit den wirksamen Listen, navs und props sind die Namen daraus in derselben
// Folge (das Orakel der Tests).
type exvT struct {
	name, basis     string
	abstrakt, offen bool
	typ             model.EntityType
	navs, props     []string
}

// exvModell bildet die wirksamen Listen je Typ: die Member der Basis-Kette zuerst.
func exvModell() []exvT {
	spezen := exvSpezen()
	nach := map[string]exvSpez{}
	for _, s := range spezen {
		nach[s.name] = s
	}
	out := make([]exvT, 0, len(spezen))
	for _, s := range spezen {
		var kette []exvSpez
		for c, ok := s, true; ok; c, ok = nach[c.basis] {
			kette = append([]exvSpez{c}, kette...)
		}
		t := exvT{name: s.name, basis: s.basis, abstrakt: s.abstrakt, offen: s.offen}
		t.typ = model.EntityType{Name: s.name, BaseType: s.basis, Abstract: s.abstrakt, Open: s.offen}
		for _, c := range kette {
			for _, p := range c.props {
				t.typ.Properties = append(t.typ.Properties, p)
				t.props = append(t.props, p.Name)
			}
			for _, n := range c.navs {
				t.typ.NavigationProperties = append(t.typ.NavigationProperties, n)
				t.navs = append(t.navs, n.Name)
			}
		}
		t.typ.Keys = []string{exvSchluessel}
		if len(t.props) > 0 {
			t.typ.Keys = []string{t.props[0], exvSchluessel}
		}
		out = append(out, t)
	}
	return out
}

// exvSchluessel steht in `Keys` jedes Typs und in keiner Liste von Navigationen oder Properties.
const exvSchluessel = "Schluessel"

// exvUmgekehrt dreht die Listen eines Typs um (zweite Folge der Achse Reihenfolge der Listen).
func exvUmgekehrt(t exvT) exvT {
	t.typ.Properties = slices.Clone(t.typ.Properties)
	t.typ.NavigationProperties = slices.Clone(t.typ.NavigationProperties)
	t.navs, t.props = slices.Clone(t.navs), slices.Clone(t.props)
	slices.Reverse(t.typ.Properties)
	slices.Reverse(t.typ.NavigationProperties)
	slices.Reverse(t.navs)
	slices.Reverse(t.props)
	return t
}

// exvBeide liefert jeden Typ in der Folge der Spezifikation und mit umgekehrten Listen.
func exvBeide() []exvT {
	m := exvModell()
	for _, t := range exvModell() {
		m = append(m, exvUmgekehrt(t))
	}
	return m
}

// exvZahl zählt die Member eines Typs aus den eigenen Listen der Kette, nicht aus den wirksamen Listen.
type exvZahl struct{ navs, props, tiefe int }

func exvZaehlen() map[string]exvZahl {
	spezen := exvSpezen()
	nach := map[string]exvSpez{}
	for _, s := range spezen {
		nach[s.name] = s
	}
	out := map[string]exvZahl{}
	for _, s := range spezen {
		z := exvZahl{}
		for c, ok := s, true; ok; c, ok = nach[c.basis] {
			z.navs += len(c.navs)
			z.props += len(c.props)
			if c.basis != "" {
				z.tiefe++
			}
		}
		out[s.name] = z
	}
	return out
}

// exvNamen ist die Menge aller Namen, die die Tests als fremden Namen verwenden: jedes Member jedes Typs, der qualifizierte
// und der einfache Name jedes Typs und zwei Namen, die kein Typ führt.
func exvNamen(m []exvT) []string {
	namen := []string{"Unbekannt", "Fremd", exvSchluessel}
	for _, t := range m {
		namen = append(namen, t.name, t.name[strings.LastIndex(t.name, ".")+1:])
		namen = append(namen, t.navs...)
		namen = append(namen, t.props...)
	}
	slices.Sort(namen)
	return slices.Compact(namen)
}

// exvFremd sind die Namen aus exvNamen, die der Typ in keiner Liste führt.
func exvFremd(m []exvT, t exvT) []string {
	var out []string
	for _, n := range exvNamen(m) {
		if !slices.Contains(t.navs, n) && !slices.Contains(t.props, n) {
			out = append(out, n)
		}
	}
	return out
}

// exvVarianten sind die Abwandlungen eines Namens, die kein Vergleich mit gleicher Schreibweise findet: erster Buchstabe
// umgeschaltet, alles klein, alles groß, angehängtes Leerzeichen und, bei `é` in der Form NFC, dasselbe in der Form NFD.
func exvVarianten(name string) []string {
	erst := strings.ToUpper(name[:1])
	if erst == name[:1] {
		erst = strings.ToLower(name[:1])
	}
	v := []string{erst + name[1:], strings.ToLower(name), strings.ToUpper(name), name + " "}
	if strings.Contains(name, exvEAkut()) {
		v = append(v, strings.ReplaceAll(name, exvEAkut(), exvEAkutNFD()))
	}
	return v
}

// exvFehlt ist die Ursache für einen Namen, den der Typ in keiner Liste führt.
func exvFehlt(t exvT) string {
	if t.offen {
		return string(model.ExpandNavigationNichtUnterstuetzt)
	}
	return string(model.ExpandNavigationUnbekannt)
}

// exvUnbekannt ist der i-te Name, den kein Typ des Fixture-Modells führt.
func exvUnbekannt(i int) string { return fmt.Sprintf("Unbekannt%d", i) }

func exvO(namen ...string) *model.ExpandOption {
	o := &model.ExpandOption{Elemente: []model.ExpandElement{}}
	for _, n := range namen {
		o.Elemente = append(o.Elemente, model.ExpandElement{Name: n})
	}
	return o
}

// exvKlasse ist das Ergebnis eines Aufrufs als Wort: `gueltig`, die Ursache, `programmfehler` (ErrUngueltig ohne
// ExpandFehler und ohne eine Sentinel der URL-Klassen) oder die Beschreibung eines anderen Fehlers.
func exvKlasse(err error) string {
	var ef *model.ExpandFehler
	switch {
	case err == nil:
		return "gueltig"
	case errors.As(err, &ef):
		return string(ef.Ursache)
	case errors.Is(err, model.ErrURLUngueltig), errors.Is(err, model.ErrURLNichtUnterstuetzt):
		return "Programmfehler mit Sentinel der URL-Klassen: " + err.Error()
	case errors.Is(err, model.ErrUngueltig):
		return "programmfehler"
	}
	return "anderer Fehler: " + err.Error()
}

// exvFall ist eine Eingabe mit ihrer Erwartung (ein Wort aus exvKlasse).
type exvFall struct {
	art  string
	typ  model.EntityType
	o    *model.ExpandOption
	soll string
}

// exvLaufe fährt die Fälle und meldet die ersten fünf Abweichungen mit der Zahl aller.
func exvLaufe(t *testing.T, faelle []exvFall) {
	t.Helper()
	falsch := 0
	for _, f := range faelle {
		if got := exvKlasse(model.PruefeExpand(f.typ, f.o)); got != f.soll {
			if falsch++; falsch <= 5 {
				t.Errorf("%s: %s, erwartet %s", f.art, got, f.soll)
			}
		}
	}
	if falsch > 5 {
		t.Errorf("insgesamt %d von %d Fällen falsch", falsch, len(faelle))
	}
}

func exvEins(b bool) int {
	if b {
		return 1
	}
	return 0
}

// exvZaehlt prüft die Zahl der Fälle gegen die Zahl, die die Tabelle der Spezifikation der Fixture nennt.
func exvZaehlt(t *testing.T, was string, ist, soll int) {
	t.Helper()
	t.Logf("%s: %d Fälle", was, ist)
	if ist != soll {
		t.Errorf("%s: %d Fälle, erwartet %d", was, ist, soll)
	}
}

// TestPruefeExpandFixture prüft das Fixture-Modell selbst: wirksame Listen gleich der Kette, keine doppelten Namen, keine
// Liste aufsteigend sortiert, keine Abwandlung gleich einem Namen des Typs, und die Elemente der Mengen, die die Tests
// durchlaufen, sind vorhanden.
func TestPruefeExpandFixture(t *testing.T) {
	zahlen := exvZaehlen()
	m := exvBeide()
	var abstrakt, offen, stufe3, ohneNav, leer, viele, blobNichtOffen, blobOffen, geerbteBlob, sammlung, entitaet int
	for _, tt := range m {
		z := zahlen[tt.name]
		if len(tt.navs) != z.navs || len(tt.props) != z.props || len(tt.typ.NavigationProperties) != z.navs || len(tt.typ.Properties) != z.props {
			t.Errorf("%s: %d und %d Member, die Kette zählt %d und %d", tt.name, len(tt.navs), len(tt.props), z.navs, z.props)
		}
		alle := append(slices.Clone(tt.navs), tt.props...)
		slices.Sort(alle)
		if len(slices.Compact(slices.Clone(alle))) != len(alle) {
			t.Errorf("%s: ein Name steht doppelt", tt.name)
		}
		for _, n := range alle {
			if n == "" || strings.TrimSpace(n) != n {
				t.Errorf("%s: Name %q", tt.name, n)
			}
			for _, v := range exvVarianten(n) {
				if v == n || slices.Contains(alle, v) {
					t.Errorf("%s: Abwandlung %q von %q kollidiert", tt.name, v, n)
				}
			}
		}
		for _, f := range exvFremd(m, tt) {
			if f == exvUnbekannt(0) {
				t.Errorf("Name %q kommt im Modell vor", f)
			}
		}
		for _, p := range tt.typ.Properties {
			if !p.Unterstuetzt() {
				blobNichtOffen += exvEins(!tt.offen)
				blobOffen += exvEins(tt.offen)
				geerbteBlob += exvEins(slices.ContainsFunc(m, func(b exvT) bool { return b.name == tt.basis && slices.Contains(b.props, p.Name) }))
			}
		}
		for _, n := range tt.typ.NavigationProperties {
			if n.Collection {
				sammlung++
			} else {
				entitaet++
			}
		}
		abstrakt += exvEins(tt.abstrakt)
		offen += exvEins(tt.offen)
		stufe3 += exvEins(z.tiefe == 2)
		ohneNav += exvEins(z.navs == 0 && z.props > 0)
		leer += exvEins(z.navs == 0 && z.props == 0)
		viele += exvEins(z.navs == 32)
	}
	for was, n := range map[string]int{"abstrakt": abstrakt, "offen": offen, "dritte Stufe": stufe3, "ohne Navigation": ohneNav, "leer": leer,
		"32 Navigationen": viele, "Property mit nicht unterstütztem Typ im nicht offenen Typ": blobNichtOffen,
		"Property mit nicht unterstütztem Typ im offenen Typ": blobOffen, "geerbte Property mit nicht unterstütztem Typ": geerbteBlob,
		"Navigation auf eine Collection": sammlung, "Navigation auf eine Entity": entitaet} {
		if n == 0 {
			t.Errorf("das Fixture-Modell hat keinen Typ: %s", was)
		}
	}
	for _, tt := range exvModell() {
		for _, liste := range [][]string{tt.navs, tt.props} {
			if len(liste) >= 2 && slices.IsSorted(liste) {
				t.Errorf("%s: Liste %v ist in der Folge der Spezifikation aufsteigend sortiert", tt.name, liste)
			}
		}
	}
	if len(exvSpezen())*2 != len(m) {
		t.Errorf("%d Typen in beiden Folgen, %d Spezifikationen", len(m), len(exvSpezen()))
	}
	// die Achsen offen × abstrakt × abgeleitet: jede Kombination hat mindestens einen Typ
	achse := []bool{false, true}
	kombinationen := map[[3]bool]int{}
	for _, tt := range exvModell() {
		kombinationen[[3]bool{tt.offen, tt.abstrakt, tt.basis != ""}]++
	}
	n := 0
	for _, o := range achse {
		for _, a := range achse {
			for _, d := range achse {
				if kombinationen[[3]bool{o, a, d}] == 0 {
					t.Errorf("kein Typ mit offen=%v, abstrakt=%v, abgeleitet=%v", o, a, d)
				}
				n++
			}
		}
	}
	if n != len(achse)*len(achse)*len(achse) || len(kombinationen) != n {
		t.Errorf("%d Kombinationen der Achsen, %d im Fixture-Modell", n, len(kombinationen))
	}
	for _, tt := range exvModell() {
		if !slices.Contains(tt.typ.Keys, exvSchluessel) || (len(tt.props) > 0 && !slices.Contains(tt.typ.Keys, tt.props[0])) {
			t.Errorf("%s: Keys %v", tt.name, tt.typ.Keys)
		}
	}
	// `é` steht in einer Navigation und in einer Property
	nav, prop := 0, 0
	for _, tt := range exvModell() {
		nav += exvEAkutNamen(tt.navs)
		prop += exvEAkutNamen(tt.props)
	}
	if nav == 0 || prop == 0 {
		t.Errorf("é in %d Navigationen und %d Properties", nav, prop)
	}
}

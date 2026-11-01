package outboundtest

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	// siebenV2 sind die unterstützten V2-Typen (SPEC-170 Nr. 4), achtV2 die bekannten nicht unterstützten (SPEC-171 Nr. 5).
	siebenV2 = "Edm.String Edm.Boolean Edm.Int32 Edm.Int64 Edm.Decimal Edm.Guid Edm.DateTime"
	achtV2   = "Edm.Binary Edm.Byte Edm.DateTimeOffset Edm.Double Edm.Int16 Edm.SByte Edm.Single Edm.Time"
	// nurV4 sind die Namen von V4, die V2 nicht kennt.
	nurV4 = "Edm.Date Edm.Duration Edm.TimeOfDay Edm.Stream Edm.Untyped " +
		"Edm.Geography Edm.GeographyPoint Edm.GeographyLineString Edm.GeographyPolygon Edm.GeographyMultiPoint Edm.GeographyMultiLineString Edm.GeographyMultiPolygon Edm.GeographyCollection " +
		"Edm.Geometry Edm.GeometryPoint Edm.GeometryLineString Edm.GeometryPolygon Edm.GeometryMultiPoint Edm.GeometryMultiLineString Edm.GeometryMultiPolygon Edm.GeometryCollection"
)

// modellTypV2 ist der Typ im Modell zu einem bekannten Namen des Dokuments (SPEC-171 Nr. 5).
func modellTypV2(name string) string {
	switch name {
	case "Edm.DateTime":
		return "Edm.DateTimeOffset"
	case "Edm.DateTimeOffset":
		return "Edm.DateTimeOffset(V2)"
	}
	return name
}

// ketteGen sind die Typen T1 bis Tn, jeder die Ableitung des vorigen; abstrakt gilt für alle, schluessel sagt, wo der Key steht.
func ketteGen(n int, abstrakt bool, schluessel string) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		attrs := ""
		if abstrakt {
			attrs = ` Abstract="true"`
		}
		if i > 1 {
			attrs += fmt.Sprintf(` BaseType="Ns.T%d"`, i-1)
		}
		inhalt := ""
		if (schluessel == "geerbt" && i == 1) || (schluessel == "eigener" && i == n) {
			inhalt = schluesselID
		}
		fmt.Fprintf(&b, `<EntityType Name="T%d"%s>%s</EntityType>`, i, attrs, inhalt)
	}
	return b.String()
}

func faelleV2D1() []fall {
	abstrakt := `<EntityType Name="A" Abstract="true"/>`
	out := []fall{
		schlecht("erster Typ", eins2(typOhne("A", "")+typ("B", "")+typ("C", "")), model.UrsacheOhneSchluessel, "Ns.A"),
		schlecht("mittlerer Typ", eins2(typ("A", "")+typOhne("B", "")+typ("C", "")), model.UrsacheOhneSchluessel, "Ns.B"),
		schlecht("letzter Typ", eins2(typ("A", "")+typ("B", "")+typOhne("C", "")), model.UrsacheOhneSchluessel, "Ns.C"),
		schlecht("zweites Schema", []byte(edmx2(schema2("N1", typ("A", ""))+schema2("N2", typOhne("B", "")))), model.UrsacheOhneSchluessel, "N2.B"),
		schlecht("leerer Key", eins2(typ("A", "")+`<EntityType Name="B"><Key/>`+prop("Id", "Edm.Int32")+`</EntityType>`), model.UrsacheOhneSchluessel, "Ns.B"),
		schlecht("Key nur mit Kommentar", eins2(`<EntityType Name="A"><Key><!-- c --></Key></EntityType>`), model.UrsacheOhneSchluessel, "Ns.A"),
		schlecht("abgeleiteter Typ, Basis ohne Schlüssel", eins2(ableitung("D", "Ns.B", "")+typOhne("B", "")), model.UrsacheOhneSchluessel, "Ns.D"),
		schlecht("abstrakte Basis ohne Schlüssel, konkrete Ableitung ohne Schlüssel", eins2(abstrakt+ableitung("D", "Ns.A", "")), model.UrsacheOhneSchluessel, "Ns.D"),
		gut("abstrakter Typ ohne Schlüssel", eins2(abstrakt)),
		gut("abstrakte Basis, Ableitung definiert den Schlüssel", eins2(abstrakt+ableitung("D", "Ns.A", schluesselID))),
	}
	for _, abs := range []bool{false, true} {
		for _, form := range []string{"keiner", "eigener", "geerbt"} {
			for n := 1; n <= 3; n++ {
				if form == "geerbt" && n == 1 {
					continue
				}
				name := fmt.Sprintf("Schlüssel %s, Kette %d, abstrakt %v", form, n, abs)
				doc := eins2(ketteGen(n, abs, form))
				if abs || form == "geerbt" || (form == "eigener" && n == 1) {
					out = append(out, gut(name, doc))
				} else {
					out = append(out, schlecht(name, doc, model.UrsacheOhneSchluessel, "Ns.T1"))
				}
			}
		}
	}
	return out
}

func zeileV2D1(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2D1()) }

// d2docV2 ist ein Dokument mit Alias `Self`, einem ComplexType und einem Verweis auf ein externes Dokument.
func d2docV2(inhalt string) []byte {
	return []byte(`<edmx:Edmx xmlns:edmx="` + v2EdmxNs + `" Version="1.0"><edmx:Reference Uri="u"><edmx:Include Namespace="Ext"/></edmx:Reference>` +
		`<edmx:DataServices>` + schemaAlias2("Ns", "Self", `<ComplexType Name="Cx"/>`+inhalt) + `</edmx:DataServices></edmx:Edmx>`)
}

func faelleV2D2() []fall {
	set := func(ref string) string {
		return typ("T", "") + `<EntityContainer Name="C"><EntitySet Name="S" EntityType="` + ref + `"/></EntityContainer>`
	}
	arten := []struct {
		name, ort string
		inhalt    func(string) string
	}{
		{"BaseType", "Ns.D@BaseType", func(r string) string { return typ("T", "") + ableitung("D", r, "") }},
		{"Typ einer Property", "Ns.T/P@Type", func(r string) string { return typ("T", prop("P", r)) }},
		{"EntityType eines Entity-Sets", "Ns.C/S@EntityType", set},
	}
	var out []fall
	for _, a := range arten {
		for _, r := range []string{"Self.Nope", "Ns.Nope", "Zzz.T", "T", "Ext.Thing", "Self.Cx", "ns.T", "Ns.t", " Ns.T", "Ns.T ", " Self.T", "Self.T ", "Ns.", ".T"} {
			if r == "Self.Cx" && a.name == "Typ einer Property" {
				continue
			}
			out = append(out, schlecht(a.name+" "+r, d2docV2(a.inhalt(r)), model.UrsacheNichtAufloesbar, a.ort))
		}
	}
	for _, r := range []string{"Nope", "Zzz", "N", " Id", "Id ", "id", "Id2"} {
		doc := `<EntityType Name="T"><Key><PropertyRef Name="` + r + `"/></Key>` + prop("Id", "Edm.Int32") + `</EntityType>`
		out = append(out, schlecht("Name einer PropertyRef "+r, d2docV2(doc), model.UrsacheNichtAufloesbar, "Ns.T/"+r+"@Name"))
	}
	return append(out, gut("Self.Cx als Typ einer Property", d2docV2(typ("T", prop("P", "Self.Cx")))),
		gut("Alias als Namensraum der Referenzen", d2docV2(typ("T", "")+ableitung("D", "Self.T", "")+
			`<EntityContainer Name="C"><EntitySet Name="S" EntityType="Self.T"/></EntityContainer>`)))
}

func zeileV2D2(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2D2()) }

func kettenDocV2(n int, wurzelZuerst bool) []byte { return eins2(kettenTypen(n, wurzelZuerst)) }

// faelleV2D3Ringe sind die Ketten, die sich schließen, mit dem Orakel aus der Zahl der verschiedenen Typen der Kette
// des ersten Typs im Dokument: höchstens 64 ist nicht_aufloesbar, sonst kette (SPEC-125 Nr. 9 a).
func faelleV2D3Ringe() []fall {
	var out []fall
	for _, ring := range []int{1, 2, 63, 64, 65, 66} {
		for _, schwanz := range []int{0, 1} {
			for _, rueck := range []bool{false, true} {
				typen, erster := ringTypen(schwanz, ring, rueck)
				verschieden := ring
				if schwanz > 0 && !rueck {
					verschieden += schwanz
				}
				name := fmt.Sprintf("Ring %d, Schwanz %d, rückwärts %v", ring, schwanz, rueck)
				if verschieden <= 64 {
					out = append(out, schlecht(name, eins2(typen), model.UrsacheNichtAufloesbar, "Ns."+erster+"@BaseType"))
				} else {
					out = append(out, schlecht(name, eins2(typen), model.UrsacheKette, "Ns."+erster))
				}
			}
		}
	}
	return out
}

func faelleV2D3() []fall {
	ab := ableitung
	nA := model.UrsacheNichtAufloesbar
	out := append(faelleV2D3Ringe(),
		schlecht("Selbstbezug", eins2(ab("A", "Ns.A", "")), nA, "Ns.A@BaseType"),
		schlecht("Zyklus über zwei Typen", eins2(ab("A", "Ns.B", "")+ab("B", "Ns.A", "")), nA, "Ns.A@BaseType"),
		schlecht("Schwanz in einen Zyklus", eins2(ab("A", "Ns.B", "")+ab("B", "Ns.C", "")+ab("C", "Ns.B", "")), nA, "Ns.A@BaseType"),
		schlecht("Zyklus über zwei Schemas", []byte(edmx2(schema2("N1", ab("A", "N2.B", ""))+schema2("N2", ab("B", "N1.A", "")))), nA, "N1.A@BaseType"))
	for _, n := range []int{63, 64, 65, 66} {
		for _, zuerst := range []bool{true, false} {
			name := fmt.Sprintf("Kette von %d, Wurzel zuerst %v", n, zuerst)
			switch {
			case n <= 64:
				out = append(out, gut(name, kettenDocV2(n, zuerst)))
			case zuerst:
				out = append(out, schlecht(name, kettenDocV2(n, zuerst), model.UrsacheKette, "Ns.T65"))
			default:
				out = append(out, schlecht(name, kettenDocV2(n, zuerst), model.UrsacheKette, fmt.Sprintf("Ns.T%d", n)))
			}
		}
	}
	return out
}

func zeileV2D3(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2D3()) }

func faelleV2D4() []fall {
	var out []fall
	typen := append(strings.Fields(siebenV2+" "+achtV2), "Ns.Cx", "Ns.Other")
	for _, ty := range typen {
		doc := eins2(`<ComplexType Name="Cx"/>` + typ("Other", "") + `<EntityType Name="E"><Key><PropertyRef Name="K"/></Key>` + prop("K", ty) + `</EntityType>`)
		if slices.Contains(strings.Fields(siebenV2), ty) {
			out = append(out, gut("Schlüssel "+ty, doc))
		} else {
			out = append(out, schlecht("Schlüssel "+ty, doc, model.UrsacheSchluesseltyp, "Ns.E/K"))
		}
		out = append(out, gut("Property "+ty, eins2(`<ComplexType Name="Cx"/>`+typ("Other", "")+typ("E", prop("P", ty)))))
	}
	pfad := `<EntityType Name="E"><Key><PropertyRef Name="Addr/City"/></Key>` + prop("Addr", "Ns.Cx") + `</EntityType>`
	return append(out, schlecht("Schlüssel über einen Pfad", eins2(`<ComplexType Name="Cx"/>`+pfad), model.UrsacheSchluesseltyp, "Ns.E/Addr/City"))
}

// zeileV2D4: nur die sieben Typen sind als Schlüssel gültig, jeder andere ist schluesseltyp (SPEC-171 Nr. 5, 7 f).
func zeileV2D4(t *testing.T, neu CodecNeu) {
	if n := len(strings.Fields(siebenV2)) + len(strings.Fields(achtV2)); n != 15 {
		t.Fatalf("%d bekannte Namen, erwartet 15", n)
	}
	if got := strings.Join(model.UnterstuetzteEDMTypen(), " "); got != "Edm.String Edm.Boolean Edm.Int32 Edm.Int64 Edm.Decimal Edm.Guid Edm.Date Edm.DateTimeOffset" {
		t.Errorf("UnterstuetzteEDMTypen: %v", model.UnterstuetzteEDMTypen())
	}
	for _, f := range faelleV2D4() {
		m, _ := pruefeFall(t, neu, f)
		if f.ursache != "" {
			continue
		}
		name, ty, _ := strings.Cut(f.name, " ")
		typen := m.EntityTypes()
		e, ok := m.EntityType("Ns.E")
		if len(typen) < 1 || !ok || len(e.Properties) == 0 {
			t.Errorf("%s: %+v", f.name, typen)
			continue
		}
		p := e.Properties[len(e.Properties)-1]
		wantTyp := modellTypV2(ty)
		unterstuetzt := slices.Contains(strings.Fields(siebenV2), ty)
		if p.EdmType != wantTyp || p.Unterstuetzt() != unterstuetzt || (name == "Schlüssel" && !slices.Equal(e.Keys, []string{"K"})) {
			t.Errorf("%s: Property %+v, Keys %v, erwartet %s", f.name, p, e.Keys, wantTyp)
		}
	}
}

func faelleV2D5() []fall {
	doc := func(ty string) []byte {
		return []byte(edmx2(schemaAlias2("Ns", "Self", `<ComplexType Name="Cx"/>`+typ("Other", "")+typ("T", prop("P", ty)))))
	}
	var out []fall
	for _, ty := range append(strings.Fields(siebenV2+" "+achtV2), "Self.Cx", "Ns.Other", "Ns.T") {
		out = append(out, gut("Typ "+ty, doc(ty)))
	}
	nA := []string{"Collection(Edm.String)", "Collection(Self.Cx)", "Collection(Collection(Edm.String))", "Collection(Self.Nope)", "Collection()",
		"Collection(Edm.String", "Collection(Edm.StringX", "Collection(Edm.String)X", "Collection(Edm.String) ", "Collection(Ns.T", "Collection(Ns.TX",
		"Edm.Foo", "Edm.", "Edm", "edm.string", " Edm.String", "Edm.String ", "edm.DateTime", "Edm.datetime", "Self.Nope", "Ns.Nope", "Nope", "Zzz.T"}
	for _, ty := range append(nA, strings.Fields(nurV4)...) {
		out = append(out, schlecht("Typ "+ty, doc(ty), model.UrsacheNichtAufloesbar, "Ns.T/P@Type"))
	}
	return out
}

// d5EdmFallV2 ist ein Dokument, dessen Property `P` den ComplexType `Cx` eines Namensraums oder Alias nennt, der mit `Edm`
// beginnt oder `Edm` in anderer Schreibweise ist, nicht `Edm` und nicht `Edm.…` (SPEC-171 Nr. 5, 10 a); ns ist der Namensraum des Typs `T`, ctyp der
// qualifizierte Name des ComplexType.
type d5EdmFallV2 struct {
	fall
	ns, ctyp string
}

// schreibweisenEdm sind die sieben Schreibweisen von `Edm` in Groß- und Kleinbuchstaben außer `Edm` selbst (`edm`, `EDM`,
// `eDm`, …): die Schreibweise zählt, reserviert ist allein `Edm` (SPEC-171 Nr. 1 mit SPEC-125 Nr. 8).
func schreibweisenEdm() []string {
	var out []string
	for maske := range 8 {
		b := []byte("edm")
		for i := range b {
			if maske&(1<<i) != 0 {
				b[i] -= 'a' - 'A'
			}
		}
		if s := string(b); s != "Edm" {
			out = append(out, s)
		}
	}
	return out
}

// faelleV2D5Edm sind die Namensräume (Edm2, EdmX, Edmonton, Edm_, EdmX.Inner) und Aliase (die vier ohne Punkt) der Klasse
// und die sieben anderen Schreibweisen von `Edm`, je als Namensraum, als Namensraum `<Schreibweise>.Sub` und als Alias
// (21 Fälle).
func faelleV2D5Edm() []d5EdmFallV2 {
	var out []d5EdmFallV2
	for _, s := range schreibweisenEdm() {
		for _, ns := range []string{s, s + ".Sub"} {
			doc := edmx2(schema2(ns, `<ComplexType Name="Cx"/>`+typ("T", prop("P", ns+".Cx"))))
			out = append(out, d5EdmFallV2{gut("Namensraum "+ns, []byte(doc)), ns, ns + ".Cx"})
		}
		doc := edmx2(schemaAlias2("Ns", s, `<ComplexType Name="Cx"/>`+typ("T", prop("P", s+".Cx"))))
		out = append(out, d5EdmFallV2{gut("Alias "+s, []byte(doc)), "Ns", "Ns.Cx"})
	}
	for _, ns := range []string{"Edm2", "EdmX", "Edmonton", "Edm_", "EdmX.Inner"} {
		doc := edmx2(schema2(ns, `<ComplexType Name="Cx"/>`+typ("T", prop("P", ns+".Cx"))))
		out = append(out, d5EdmFallV2{gut("Namensraum "+ns, []byte(doc)), ns, ns + ".Cx"})
	}
	for _, al := range []string{"Edm2", "EdmX", "Edmonton", "Edm_"} {
		doc := edmx2(schemaAlias2("Ns", al, `<ComplexType Name="Cx"/>`+typ("T", prop("P", al+".Cx"))))
		out = append(out, d5EdmFallV2{gut("Alias "+al, []byte(doc)), "Ns", "Ns.Cx"})
	}
	return out
}

// zeileV2D5Edm: ein Typ aus einem Namensraum oder Alias, der mit `Edm` beginnt und nicht `Edm.` ist, löst auf und bleibt
// mit Fund property_typ im Modell.
func zeileV2D5Edm(t *testing.T, neu CodecNeu) {
	faelle := faelleV2D5Edm()
	if len(faelle) != 30 {
		t.Errorf("%d Fälle, erwartet 30", len(faelle))
	}
	for _, f := range faelle {
		m, b := pruefeFall(t, neu, f.fall)
		e, _ := m.EntityType(f.ns + ".T")
		funde := []model.Fund{{Art: model.FundKonstrukt, Ort: f.ctyp, Name: "ComplexType"}, {Art: model.FundPropertyTyp, Ort: f.ns + ".T/P", Name: f.ctyp}}
		if len(e.Properties) != 2 || e.Properties[1].EdmType != f.ctyp || e.Properties[1].Unterstuetzt() || !slices.Equal(b.Funde, funde) {
			t.Errorf("%s: Properties %+v, Funde %v, erwartet %s und %v", f.name, e.Properties, b.Funde, f.ctyp, funde)
		}
	}
}

// zeileV2D5: die Formen des Property-Typs (SPEC-171 Nr. 5): bekannte nicht unterstützte Namen und deklarierte Typen bleiben
// mit Unterstuetzt() falsch und einem Fund im Modell, alles andere ist nicht_aufloesbar.
func zeileV2D5(t *testing.T, neu CodecNeu) {
	zeileV2D5Edm(t, neu)
	if n := len(strings.Fields(nurV4)); n != 21 {
		t.Fatalf("%d Namen von V4, erwartet 21", n)
	}
	for _, f := range faelleV2D5() {
		m, b := pruefeFall(t, neu, f)
		if f.ursache != "" {
			continue
		}
		ty := strings.TrimPrefix(f.name, "Typ ")
		e, _ := m.EntityType("Ns.T")
		p := e.Properties[len(e.Properties)-1]
		wantTyp, fund := modellTypV2(ty), ty
		switch ty {
		case "Self.Cx", "Ns.Other", "Ns.T":
			wantTyp, fund = "Ns."+ty[strings.Index(ty, ".")+1:], "Ns."+ty[strings.Index(ty, ".")+1:]
		}
		unterstuetzt := slices.Contains(strings.Fields(siebenV2), ty)
		funde := []model.Fund{{Art: model.FundKonstrukt, Ort: "Ns.Cx", Name: "ComplexType"}}
		if !unterstuetzt {
			funde = append(funde, model.Fund{Art: model.FundPropertyTyp, Ort: "Ns.T/P", Name: fund})
		}
		if p.EdmType != wantTyp || p.Unterstuetzt() != unterstuetzt || !slices.Equal(b.Funde, funde) {
			t.Errorf("%s: Property %+v, Funde %v, erwartet %s und %v", f.name, p, b.Funde, wantTyp, funde)
		}
	}
}

// volldokV2 ist ein Dokument mit jedem Pflichtattribut von SPEC-171 Nr. 10 a; o ersetzt den Wert eines Attributs
// (Schlüssel `Element.Attribut`), "-" lässt es weg.
func volldokV2(o map[string]string) []byte {
	a := func(k, v string) string {
		if w, ok := o[k]; ok {
			v = w
		}
		if v == "-" {
			return ""
		}
		_, name, _ := strings.Cut(k, ".")
		return ` ` + name + `="` + v + `"`
	}
	return []byte(`<edmx:Edmx xmlns:edmx="` + v2EdmxNs + `" Version="1.0"><edmx:DataServices><Schema xmlns="` + v2EdmNs + `"` + a("Schema.Namespace", "Ns") + a("Schema.Alias", "-") + `>` +
		`<EntityType` + a("EntityType.Name", "T") + a("EntityType.Abstract", "-") + `><Key><PropertyRef` + a("PropertyRef.Name", "Id") + `/></Key>` +
		`<Property` + a("Property.Name", "Id") + a("Property.Type", "Edm.Int32") + a("Property.Nullable", "-") + `/>` +
		`<NavigationProperty` + a("NavigationProperty.Name", "N") + a("NavigationProperty.Relationship", "Ns.A") + a("NavigationProperty.FromRole", "P") +
		a("NavigationProperty.ToRole", "Q") + `/></EntityType>` +
		`<Association` + a("Association.Name", "A") + `><End` + a("End.Type", "Ns.T") + a("End.Role", "P") + a("End.Multiplicity", "*") + `/>` +
		`<End Type="Ns.T" Role="Q" Multiplicity="0..1"/></Association>` +
		`<EntityContainer` + a("EntityContainer.Name", "C") + `><EntitySet` + a("EntitySet.Name", "S") + a("EntitySet.EntityType", "Ns.T") + `/>` +
		`<AssociationSet` + a("AssociationSet.Name", "AS") + a("AssociationSet.Association", "Ns.A") + `><End` + a("SetEnd.Role", "P") + a("SetEnd.EntitySet", "S") + `/>` +
		`<End Role="Q" EntitySet="S"/></AssociationSet></EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`)
}

// pflichtV2 sind die Pflichtattribute von SPEC-171 Nr. 10 a mit dem Ort ihres Verstoßes.
func pflichtV2() []struct{ attr, ort string } {
	return []struct{ attr, ort string }{
		{"Schema.Namespace", "@Namespace"}, {"EntityType.Name", "Ns@Name"}, {"PropertyRef.Name", "Ns.T@Name"}, {"Property.Name", "Ns.T@Name"},
		{"Property.Type", "Ns.T/Id@Type"}, {"NavigationProperty.Name", "Ns.T@Name"}, {"NavigationProperty.Relationship", "Ns.T/N@Relationship"},
		{"NavigationProperty.FromRole", "Ns.T/N@FromRole"}, {"NavigationProperty.ToRole", "Ns.T/N@ToRole"}, {"Association.Name", "Ns@Name"},
		{"End.Type", "Ns.A@Type"}, {"End.Role", "Ns.A@Role"}, {"End.Multiplicity", "Ns.A@Multiplicity"}, {"EntityContainer.Name", "Ns@Name"},
		{"EntitySet.Name", "Ns.C@Name"}, {"EntitySet.EntityType", "Ns.C/S@EntityType"}, {"AssociationSet.Name", "Ns.C@Name"},
		{"AssociationSet.Association", "Ns.C/AS@Association"}, {"SetEnd.Role", "Ns.C/AS@Role"}, {"SetEnd.EntitySet", "Ns.C/AS@EntitySet"},
	}
}

func faelleV2D6() []fall {
	struk := model.UrsacheStruktur
	out := []fall{gut("volles Dokument", volldokV2(nil))}
	for _, p := range pflichtV2() {
		out = append(out, schlecht("Pflicht "+p.attr+" fehlt", volldokV2(map[string]string{p.attr: "-"}), struk, p.ort),
			schlecht("Pflicht "+p.attr+" leer", volldokV2(map[string]string{p.attr: ""}), struk, p.ort))
	}
	for _, w := range []string{"True", "1", "yes", "", " true"} {
		for _, p := range []struct{ attr, ort string }{{"EntityType.Abstract", "Ns.T@Abstract"}, {"Property.Nullable", "Ns.T/Id@Nullable"}} {
			out = append(out, schlecht(fmt.Sprintf("%s %q", p.attr, w), volldokV2(map[string]string{p.attr: w}), struk, p.ort))
		}
	}
	for _, p := range []string{"EntityType.Abstract", "Property.Nullable"} {
		for _, w := range []string{"true", "false"} {
			out = append(out, gut(p+" "+w, volldokV2(map[string]string{p: w})))
		}
	}
	return append(out, faelleV2D6Namen()...)
}

func faelleV2D6Namen() []fall {
	struk := model.UrsacheStruktur
	var out []fall
	for _, n := range []struct {
		name    string
		gueltig bool
	}{
		{"", false}, {"1abc", false}, {"a-b", false}, {"a b", false}, {"a.b", false}, {" a", false}, {"_x", true},
		{"Größe", true}, {repeat("ä", 128), true}, {repeat("ä", 129), false}, {repeat("a", 127), true}, {repeat("a", 128), true}, {repeat("a", 129), false},
		{"a²", false}, {"a٣", true}, {"a\xCC\x81", false}, {"٣a", false}, {"aäß٣_", true},
	} {
		docs := map[string][]byte{"Ns@Name": eins2(typ(n.name, "")), "Ns.T@Name": eins2(typ("T", prop(n.name, "Edm.String"))),
			"Ns@Alias": []byte(edmx2(schemaAlias2("Ns", n.name, "")))}
		for _, ort := range []string{"Ns@Name", "Ns.T@Name", "Ns@Alias"} {
			name := fmt.Sprintf("Name %.12q (%d Bytes) in %s", n.name, len(n.name), ort)
			if n.gueltig {
				out = append(out, gut(name, docs[ort]))
			} else {
				out = append(out, schlecht(name, docs[ort], struk, ort))
			}
		}
	}
	for _, n := range []struct {
		ns      string
		gueltig bool
	}{{repeat("a", 127) + "." + repeat("b", 127), true}, {repeat("a", 127) + "." + repeat("b", 128), true}, {repeat("a", 128) + "." + repeat("b", 128), false},
		{"a.b", true}, {"a..b", false}, {".a", false}, {"a.", false}, {"A1.B2", true}, {repeat("ä", 127) + "." + repeat("ä", 128), true},
		{repeat("ä", 128) + "." + repeat("ä", 128), false}, {repeat("ä", 127) + "." + repeat("ä", 129), false}, {"a٣.b", true}, {"a.b²", false}} {
		doc := []byte(edmx2(schema2(n.ns, "")))
		if n.gueltig {
			out = append(out, gut(fmt.Sprintf("Namensraum %.12q (%d)", n.ns, len(n.ns)), doc))
		} else {
			out = append(out, schlecht(fmt.Sprintf("Namensraum %.12q (%d)", n.ns, len(n.ns)), doc, struk, "@Namespace"))
		}
	}
	return append(out, faelleV2D6Aliase()...)
}

func faelleV2D6Aliase() []fall {
	struk := model.UrsacheStruktur
	two := func(a, b string) []byte { return []byte(edmx2(a + b)) }
	out := []fall{
		schlecht("Edm als Namespace", []byte(edmx2(schema2("Edm", ""))), struk, "Edm@Namespace"),
		schlecht("Edm als Alias", []byte(edmx2(schemaAlias2("Ns", "Edm", ""))), struk, "Ns@Alias"),
		schlecht("Edm.Sub als Namespace", []byte(edmx2(schema2("Edm.Sub", ""))), struk, "Edm.Sub@Namespace"),
		schlecht("Edm.Sub.Deep als Namespace", []byte(edmx2(schema2("Edm.Sub.Deep", ""))), struk, "Edm.Sub.Deep@Namespace"),
		schlecht("Edm.Sub im zweiten Schema", two(schema2("N1", ""), schema2("Edm.Sub", "")), struk, "Edm.Sub@Namespace"),
		schlecht("Alias doppelt für zwei Namensräume", two(schemaAlias2("N1", "X", ""), schemaAlias2("N2", "X", "")), struk, "N2@Alias"),
		schlecht("Alias gleich dem späteren Namensraum", two(schemaAlias2("N1", "N2", ""), schema2("N2", "")), struk, "N2@Namespace"),
		schlecht("Alias gleich dem früheren Namensraum", two(schema2("N2", ""), schemaAlias2("N1", "N2", "")), struk, "N1@Alias"),
		gut("zwei Aliase desselben Namensraums", two(schemaAlias2("N1", "A", ""), schemaAlias2("N1", "B", ""))),
		gut("Alias gleich dem eigenen Namensraum", []byte(edmx2(schemaAlias2("N1", "N1", "")))),
	}
	for _, n := range []string{"edm", "EdmX", "Edmonton", "Edm2"} {
		out = append(out, gut("Namespace "+n, []byte(edmx2(schema2(n, "")))), gut("Alias "+n, []byte(edmx2(schemaAlias2("Ns", n, "")))))
	}
	return append(out, gut("Namespace Ns.Edm", []byte(edmx2(schema2("Ns.Edm", "")))), gut("Alias Edm2 neben Namespace Edm2", []byte(edmx2(schemaAlias2("Edm2", "Edm2", "")))))
}

func zeileV2D6(t *testing.T, neu CodecNeu) {
	if n := len(pflichtV2()); n != 20 {
		t.Fatalf("%d Pflichtattribute, erwartet 20", n)
	}
	pruefeFaelle(t, neu, faelleV2D6())
}

// faelleV2D7 sind die Doppelungen von SPEC-171 Nr. 10 e und SPEC-125 Nr. 8 e: eine Zeile je Element der Menge.
func faelleV2D7() []fall {
	struk := model.UrsacheStruktur
	twoSchemas := func(a, b string) []byte { return []byte(edmx2(schema2("Ns", a) + schema2("Ns", b))) }
	tt, aa := typ("T", ""), `<Association Name="A"><End Type="Ns.T" Role="F" Multiplicity="*"/><End Type="Ns.T" Role="G" Multiplicity="*"/></Association>`
	set := `<EntitySet Name="S" EntityType="Ns.T"/>`
	asetS := func(n string) string {
		return `<AssociationSet Name="` + n + `" Association="Ns.A"><End Role="F" EntitySet="S"/><End Role="G" EntitySet="S"/></AssociationSet>`
	}
	return []fall{
		schlecht("Entity-Typ doppelt", eins2(tt+tt), struk, "Ns.T"), schlecht("Entity-Typ doppelt über zwei Schemas", twoSchemas(tt, tt), struk, "Ns.T"),
		schlecht("Property doppelt", eins2(typ("T", prop("P", "Edm.String")+prop("P", "Edm.String"))), struk, "Ns.T/P"),
		schlecht("Property und Navigation-Property gleichen Namens", eins2(typ("T", prop("P", "Edm.String")+nav2("P"))), struk, "Ns.T/P"),
		schlecht("Navigation-Property doppelt", eins2(typ("T", nav2("P")+nav2("P"))), struk, "Ns.T/P"),
		schlecht("EntitySet doppelt", eins2(tt+`<EntityContainer Name="C">`+set+set+`</EntityContainer>`), struk, "Ns.C/S"),
		schlecht("PropertyRef doppelt", eins2(`<EntityType Name="T"><Key><PropertyRef Name="Id"/><PropertyRef Name="Id"/></Key>`+prop("Id", "Edm.Int32")+`</EntityType>`), struk, "Ns.T"),
		schlecht("zwei Key", eins2(`<EntityType Name="T"><Key><PropertyRef Name="Id"/></Key><Key><PropertyRef Name="Other"/></Key>`+prop("Id", "Edm.Int32")+prop("Other", "Edm.Int32")+`</EntityType>`), struk, "Ns.T"),
		schlecht("zweiter EntityContainer im selben Schema", eins2(`<EntityContainer Name="C"/><EntityContainer Name="D"/>`), struk, "Ns.D"),
		schlecht("zweiter EntityContainer im zweiten Schema", []byte(edmx2(schema2("N1", `<EntityContainer Name="C"/>`)+schema2("N2", `<EntityContainer Name="D"/>`))), struk, "N2.D"),
		schlecht("zweiter EntityContainer gleichen Namens", eins2(`<EntityContainer Name="C"/><EntityContainer Name="C"/>`), struk, "Ns.C"),
		schlecht("Association doppelt", eins2(tt+aa+aa), struk, "Ns.A"), schlecht("Association doppelt über zwei Schemas", twoSchemas(tt+aa, aa), struk, "Ns.A"),
		schlecht("AssociationSet doppelt", eins2(tt+aa+`<EntityContainer Name="C">`+set+asetS("X")+asetS("X")+`</EntityContainer>`), struk, "Ns.C/X"),
		gut("Association gleichen Namens in zwei Namensräumen", []byte(edmx2(schema2("N1", aa)+schema2("N2", aa)))),
		gut("gleicher Name in zwei Namensräumen", []byte(edmx2(schema2("N1", tt)+schema2("N2", tt)))),
		gut("EntitySet und AssociationSet gleichen Namens", eins2(tt+aa+`<EntityContainer Name="C">`+set+asetS("S")+`</EntityContainer>`)),
		gut("Property gleichen Namens in zwei Typen", eins2(typ("T", prop("P", "Edm.String"))+typ("U", prop("P", "Edm.String")))),
	}
}

func zeileV2D7(t *testing.T, neu CodecNeu) {
	faelle := faelleV2D7()
	if n := len(faelle); n < 17 {
		t.Fatalf("%d Fälle", n)
	}
	pruefeFaelle(t, neu, faelle)
}

func faelleV2D8() []fall {
	struk := model.UrsacheStruktur
	keyAb := `<Key><PropertyRef Name="Id"/></Key>`
	return []fall{
		schlecht("Key in Basis und Ableitung", eins2(typ("B", "")+ableitung("D", "Ns.B", keyAb)), struk, "Ns.D"),
		gut("Key nur im abgeleiteten Typ", eins2(`<EntityType Name="B" Abstract="true">`+prop("Id", "Edm.Int32")+`</EntityType>`+ableitung("D", "Ns.B", keyAb))),
		schlecht("Key in Wurzel, zweiter und dritter Stufe", eins2(typ("W", "")+ableitung("S", "Ns.W", keyAb)+ableitung("T", "Ns.S", keyAb)), struk, "Ns.S"),
		schlecht("Key in Wurzel und dritter Stufe", eins2(typ("W", "")+ableitung("S", "Ns.W", "")+ableitung("T", "Ns.S", keyAb)), struk, "Ns.T"),
		schlecht("geerbter Name mit anderem Typ", eins2(typ("B", prop("P", "Edm.String"))+ableitung("D", "Ns.B", prop("P", "Edm.Int32"))), struk, "Ns.D/P"),
		schlecht("geerbter Name mit gleichem Typ", eins2(typ("B", prop("P", "Edm.String"))+ableitung("D", "Ns.B", prop("P", "Edm.String"))), struk, "Ns.D/P"),
		schlecht("geerbte Navigation-Property als Property", eins2(typ("B", nav2("P"))+ableitung("D", "Ns.B", prop("P", "Edm.String"))), struk, "Ns.D/P"),
		schlecht("geerbte Property gegen eigene Navigation-Property", eins2(typ("B", prop("P", "Edm.String"))+ableitung("D", "Ns.B", nav2("P"))), struk, "Ns.D/P"),
		schlecht("geerbte Navigation-Property gegen eigene", eins2(typ("B", nav2("P"))+ableitung("D", "Ns.B", nav2("P"))), struk, "Ns.D/P"),
		schlecht("geerbte Property in der dritten Stufe gegen Navigation-Property", eins2(typ("B", prop("P", "Edm.String"))+ableitung("D", "Ns.B", "")+ableitung("E", "Ns.D", nav2("P"))), struk, "Ns.E/P"),
		gut("leerer Key in der Ableitung bei Schlüssel in der Basis", eins2(typ("B", "")+ableitung("D", "Ns.B", "<Key/>"))),
		gut("Key in der mittleren Stufe, Wurzel ohne Schlüssel", eins2(`<EntityType Name="W" Abstract="true"/>`+ableitung("S", "Ns.W", schluesselID)+ableitung("T", "Ns.S", ""))),
		schlecht("Key in der mittleren und der dritten Stufe, Wurzel ohne Schlüssel", eins2(`<EntityType Name="W" Abstract="true"/>`+ableitung("S", "Ns.W", schluesselID)+ableitung("T", "Ns.S", keyAb)), struk, "Ns.T"),
		gut("gleicher Name in zwei Geschwistern", eins2(typ("B", "")+ableitung("D1", "Ns.B", prop("P", "Edm.String"))+ableitung("D2", "Ns.B", prop("P", "Edm.String")))),
	}
}

func zeileV2D8(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2D8()) }

// durchlaeufeV2 sind die acht Durchläufe des Auflösens (SPEC-171 Nr. 7), je mit einem Dokumentteil, der allein diesen
// Durchlauf verletzt, und dessen Ursache und Ort.
func durchlaeufeV2() []durchlauf {
	nA, struk := model.UrsacheNichtAufloesbar, model.UrsacheStruktur
	return []durchlauf{
		{"a1 BaseType", ableitung("X1", "Ns.Nope", ""), nA, "Ns.X1@BaseType"},
		{"a2 Kette", ableitung("Y1", "Ns.Y2", "") + ableitung("Y2", "Ns.Y1", ""), nA, "Ns.Y1@BaseType"},
		{"b Summe", typ("B", props(65536)), model.UrsacheZuVieleMembers, "Ns.B"},
		{"c Schlüssel in Ketten", typ("C1", "") + ableitung("C2", "Ns.C1", keyAbleitung), struk, "Ns.C2"},
		{"d Typen", typ("Q1", prop("P", "Ns.Nope")), nA, "Ns.Q1/P@Type"},
		{"e Sets", `<EntityContainer Name="Ct"><EntitySet Name="Sx" EntityType="Ns.Nope2"/></EntityContainer>`, nA, "Ns.Ct/Sx@EntityType"},
		{"f Schlüsseltyp", `<EntityType Name="F1"><Key><PropertyRef Name="K"/></Key>` + prop("K", "Edm.Double") + `</EntityType>`, model.UrsacheSchluesseltyp, "Ns.F1/K"},
		{"g ohne Schlüssel", typOhne("G1", ""), model.UrsacheOhneSchluessel, "Ns.G1"},
	}
}

// faelleV2D9 führt für jedes der 28 Paare (früherer, späterer Durchlauf) in beiden Dokumentreihenfolgen ein Dokument,
// das beide verletzt; gemeldet wird der Verstoß des früheren.
func faelleV2D9() []fall {
	var out []fall
	d := durchlaeufeV2()
	for i := range d {
		for j := i + 1; j < len(d); j++ {
			for _, rueck := range []bool{false, true} {
				doc, name := d[j].teil+d[i].teil, fmt.Sprintf("Paar %s vor %s, späterer zuerst", d[i].name, d[j].name)
				if rueck {
					doc, name = d[i].teil+d[j].teil, fmt.Sprintf("Paar %s vor %s, früherer zuerst", d[i].name, d[j].name)
				}
				f := schlecht(name, eins2(doc), d[i].ursache, d[i].ort)
				if i == 2 || j == 2 {
					f = gross(f)
				}
				out = append(out, f)
			}
		}
	}
	nA := model.UrsacheNichtAufloesbar
	return append(out,
		schlecht("Verweis ohne Ziel gewinnt gegen nicht unterstützten Typ", eins2(typ("A", prop("W", "Edm.Double"))+typ("B", prop("P", "Ns.Nope"))), nA, "Ns.B/P@Type"),
		schlecht("Property vor Navigation-Property desselben Typs", eins2(typ("T", prop("P", "Ns.Nope")+nav2("N"))), nA, "Ns.T/P@Type"),
		schlecht("Navigation-Property vor Property desselben Typs", eins2(typ("T", nav2("N")+prop("P", "Ns.Nope"))), nA, "Ns.T/N@Relationship"),
		schlecht("Typ des ersten vor Navigation des zweiten", eins2(typ("A", prop("P", "Ns.Nope"))+typ("B", nav2("N"))), nA, "Ns.A/P@Type"),
	)
}

func zeileV2D9(t *testing.T, neu CodecNeu) {
	if n, soll := len(durchlaeufeV2()), 8; n != soll || len(faelleV2D9()) < 2*(n*(n-1)/2) {
		t.Errorf("%d Durchläufe und %d Fälle", n, len(faelleV2D9()))
	}
	pruefeFaelle(t, neu, faelleV2D9())
	r := teile(eins2(typ("T", prop("P", "Ns.Nope"))), 60)
	ctx, abbruch := context.WithCancel(context.Background())
	r.nach = func(n int) {
		if n == 2 {
			abbruch()
		}
	}
	r.mitEnde = true
	if _, _, err := neu.aufruf(ctx, r, 1<<20); err != context.Canceled {
		t.Errorf("Kontext endet im letzten Read vor dem Auflösen: %v", err)
	}
	zeileV2D9Mehrere(t, neu)
}

// zeileV2D9Mehrere: mehrere struktur-Verstöße in einem Element melden struktur mit dem Ort eines der Verstöße; welcher,
// ist nicht zugesagt (SPEC-125 Nr. 8 f).
func zeileV2D9Mehrere(t *testing.T, neu CodecNeu) {
	for _, f := range []struct {
		name string
		doc  []byte
		orte []string
	}{
		{"Property mit Nullable x und ohne Type", eins2(typ("T", `<Property Name="p" Nullable="x"/>`)), []string{"Ns.T/p@Nullable", "Ns.T/p@Type"}},
		{"Property mit doppeltem Namen und Nullable x", eins2(typ("T", prop("p", "Edm.String")+`<Property Name="p" Type="Edm.String" Nullable="x"/>`)), []string{"Ns.T/p@Nullable", "Ns.T/p"}},
		{"Pflichtangabe fehlt und Doppelname im selben Element", eins2(typ("T", prop("P", "Edm.String")+`<Property Name="P"/>`)), []string{"Ns.T/P@Type", "Ns.T/P"}},
		{"Entity-Typ doppelt und Abstract x", eins2(typ("T", "") + `<EntityType Name="T" Abstract="x">` + schluesselID + `</EntityType>`), []string{"Ns.T", "Ns.T@Abstract"}},
		{"Navigation-Property ohne Relationship und ohne FromRole", eins2(typ("T", `<NavigationProperty Name="n" ToRole="G"/>`)), []string{"Ns.T/n@Relationship", "Ns.T/n@FromRole"}},
		{"End ohne Role und mit Multiplicity x", eins2(typ("T", "") + `<Association Name="R"><End Type="Ns.T" Multiplicity="x"/></Association>`), []string{"Ns.R@Role", "Ns.R@Multiplicity"}},
	} {
		_, _, err := lies(neu, gut(f.name, f.doc))
		var mf *model.MetadatenFehler
		if !errors.As(err, &mf) || mf.Ursache != model.UrsacheStruktur || !slices.Contains(f.orte, mf.Ort) {
			t.Errorf("%s: %v, erwartet struktur mit einem der Orte %q", f.name, err, f.orte)
		}
	}
}

// zeileV2D10: jeder Verstoß ist ODR-E6004 mit genau einer der 13 Ursachen (SPEC-125 Nr. 13).
func zeileV2D10(t *testing.T, neu CodecNeu) {
	gesehen := map[model.MetadatenUrsache]bool{}
	for _, f := range alleFaelleV2(t) {
		if f.ursache == "" {
			continue
		}
		gesehen[f.ursache] = true
		pruefeFall(t, neu, f)
	}
	for _, u := range strings.Fields("syntax kodierung doctype wurzel struktur version ohne_schluessel nicht_aufloesbar schluesseltyp zu_gross tiefe kette zu_viele_properties") {
		if !gesehen[model.MetadatenUrsache(u)] {
			t.Errorf("Ursache %s kommt in keinem Fehlerfall vor", u)
		}
	}
	if len(gesehen) != 13 {
		t.Errorf("%d Ursachen in den Fehlerfällen, erwartet 13", len(gesehen))
	}
}

// d11V2 sind die Parameter eines Dokuments der Klasse von D11: Schemas, Typen, Properties je Typ, Assoziation, Sets, fremde Teile.
type d11V2 struct {
	schemas, typen, props int
	assoc, sets, fremd    bool
}

func d11Alle() []d11V2 {
	var out []d11V2
	for i := range 3 * 4 * 3 * 8 {
		p := d11V2{schemas: i % 3, typen: i / 3 % 4, props: []int{0, 1, 7}[i/12%3], assoc: i/36%2 == 1, sets: i/72%2 == 1, fremd: i/144%2 == 1}
		if (p.schemas == 0 && p.typen > 0) || (p.typen == 0 && (p.props != 0 || p.assoc)) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// d11Soll ist die Erwartung zu einem Parametersatz, allein aus den Parametern gebildet.
type d11Soll struct {
	namespaces []string
	typen      []model.EntityType
	sets       []model.EntitySet
	bindings   []model.NavigationBinding
	funde      []model.Fund
}

// d11Dokument erzeugt das Dokument zu p und seine Erwartung.
func d11Dokument(p d11V2) ([]byte, d11Soll) {
	sieben := strings.Fields(siebenV2)
	var soll d11Soll
	schemas := make([]strings.Builder, p.schemas)
	var kopf strings.Builder
	for i := range p.schemas {
		soll.namespaces = append(soll.namespaces, fmt.Sprintf("N%d", i+1))
	}
	if p.fremd {
		fremd := `<f:Foo xmlns:f="urn:x"/>`
		if p.schemas == 0 {
			kopf.WriteString(fremd)
			soll.funde = append(soll.funde, model.Fund{Art: model.FundErweiterung, Name: "{urn:x}Foo"})
		} else {
			schemas[0].WriteString(fremd + `<ComplexType Name="Cx"/>`)
			soll.funde = append(soll.funde, model.Fund{Art: model.FundErweiterung, Ort: "N1", Name: "{urn:x}Foo"}, model.Fund{Art: model.FundKonstrukt, Ort: "N1.Cx", Name: "ComplexType"})
		}
	}
	qn := func(i int) string { return fmt.Sprintf("N%d.T%d", (i-1)%p.schemas+1, i) }
	imSchema := make([][]model.EntityType, p.schemas)
	for i := 1; i <= p.typen; i++ {
		et := model.EntityType{Name: qn(i), Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32", Nullable: true}}}
		inhalt := schluesselID
		for k := range p.props {
			ty := sieben[(k+i)%7]
			inhalt += prop(fmt.Sprintf("P%d", k), ty)
			et.Properties = append(et.Properties, model.Property{Name: fmt.Sprintf("P%d", k), EdmType: modellTypV2(ty), Nullable: true})
		}
		if p.assoc && i == 1 {
			inhalt += `<NavigationProperty Name="N" Relationship="N1.R" FromRole="F" ToRole="G"/>`
			et.NavigationProperties = []model.NavigationProperty{{Name: "N", Target: qn(p.typen), Collection: p.typen == 3}}
		}
		schemas[(i-1)%p.schemas].WriteString(`<EntityType Name="` + fmt.Sprintf("T%d", i) + `">` + inhalt + `</EntityType>`)
		imSchema[(i-1)%p.schemas] = append(imSchema[(i-1)%p.schemas], et)
	}
	for _, ts := range imSchema {
		soll.typen = append(soll.typen, ts...)
	}
	mult := map[bool]string{false: "0..1", true: "*"}[p.typen == 3]
	if p.assoc {
		schemas[0].WriteString(`<Association Name="R"><End Type="` + qn(1) + `" Role="F" Multiplicity="*"/><End Type="` + qn(p.typen) + `" Role="G" Multiplicity="` + mult + `"/></Association>`)
	}
	if p.sets && p.schemas > 0 {
		schemas[0].WriteString(d11Container(p, &soll, qn))
	}
	var doc strings.Builder
	doc.WriteString(`<edmx:Edmx xmlns:edmx="` + v2EdmxNs + `" Version="1.0"><edmx:DataServices>` + kopf.String())
	for i := range schemas {
		doc.WriteString(schema2(fmt.Sprintf("N%d", i+1), schemas[i].String()))
	}
	doc.WriteString(`</edmx:DataServices></edmx:Edmx>`)
	return []byte(doc.String()), soll
}

func d11Container(p d11V2, soll *d11Soll, qn func(int) string) string {
	var b strings.Builder
	b.WriteString(`<EntityContainer Name="C">`)
	for i := 1; i <= p.typen; i++ {
		fmt.Fprintf(&b, `<EntitySet Name="S%d" EntityType="%s"/>`, i, qn(i))
		soll.sets = append(soll.sets, model.EntitySet{Name: fmt.Sprintf("S%d", i), EntityType: qn(i)})
	}
	if p.assoc {
		fmt.Fprintf(&b, `<AssociationSet Name="AR" Association="N1.R"><End Role="F" EntitySet="S1"/><End Role="G" EntitySet="S%d"/></AssociationSet>`, p.typen)
		soll.bindings = append(soll.bindings, model.NavigationBinding{Set: "S1", Path: "N", Target: fmt.Sprintf("N1.C/S%d", p.typen)})
	}
	b.WriteString(`</EntityContainer>`)
	return b.String()
}

func faelleV2D11() []fall {
	var out []fall
	for _, p := range d11Alle() {
		doc, _ := d11Dokument(p)
		out = append(out, gut(fmt.Sprintf("Klasse %+v", p), doc))
	}
	return out
}

// zeileV2D11: jedes Dokument der Klasse ergibt das Modell und die Funde, die die Parameter der Erzeugung vorgeben
// (Orakel ohne den Leser), und jeder Wert jedes Parameters kommt vor.
func zeileV2D11(t *testing.T, neu CodecNeu) {
	alle := d11Alle()
	zaehler := map[string]int{}
	for _, p := range alle {
		for k, v := range map[string]any{"schemas": p.schemas, "typen": p.typen, "props": p.props, "assoc": p.assoc, "sets": p.sets, "fremd": p.fremd} {
			zaehler[fmt.Sprint(k, "=", v)]++
		}
		doc, soll := d11Dokument(p)
		m, b := pruefeFall(t, neu, gut(fmt.Sprintf("Klasse %+v", p), doc))
		if m.Version() != "2.0" {
			t.Errorf("%+v: Version %q", p, m.Version())
		}
		gleich(t, fmt.Sprintf("%+v Namespaces", p), m.Namespaces(), soll.namespaces)
		gleich(t, fmt.Sprintf("%+v EntityTypes", p), m.EntityTypes(), soll.typen)
		gleich(t, fmt.Sprintf("%+v EntitySets", p), m.EntitySets(), soll.sets)
		gleich(t, fmt.Sprintf("%+v Bindings", p), m.NavigationBindings(), soll.bindings)
		if !slices.Equal(b.Funde, soll.funde) && (len(b.Funde) > 0 || len(soll.funde) > 0) {
			t.Errorf("%+v: Funde %v, erwartet %v", p, b.Funde, soll.funde)
		}
	}
	for _, k := range []string{"schemas=0", "schemas=1", "schemas=2", "typen=0", "typen=1", "typen=2", "typen=3", "props=0", "props=1", "props=7",
		"assoc=true", "assoc=false", "sets=true", "sets=false", "fremd=true", "fremd=false"} {
		if zaehler[k] == 0 {
			t.Errorf("Wert %s kommt in keinem Dokument vor", k)
		}
	}
	if len(alle) < 150 {
		t.Errorf("%d Dokumente der Klasse, erwartet mindestens 150", len(alle))
	}
}

// navsV2 sind n Navigation-Properties auf die Association `Ns.R`.
func navsV2(n int) string {
	var b strings.Builder
	for i := range n {
		b.WriteString(nav2(fmt.Sprintf("n%d", i)))
	}
	return b.String()
}

func faelleV2E3() []fall {
	z := model.UrsacheZuVieleMembers
	return []fall{
		gross(gut("65 536 Properties", eins2(typ("T", props(65535))))),
		gross(schlecht("65 537 Properties", eins2(typ("T", props(65536))), z, "Ns.T")),
		gross(gut("32 768 Properties und 32 768 Navigation-Properties", eins2(typ("T", props(32767)+navsV2(32768))+assoz2))),
		gross(schlecht("eine Navigation-Property mehr", eins2(typ("T", props(32767)+navsV2(32769))+assoz2), z, "Ns.T")),
		gross(gut("Basis mit 256 Members und 255 Ableitungen", eins2(typ("B", props(255))+ableitungen(255)))),
		gross(schlecht("256 Ableitungen", eins2(typ("B", props(255))+ableitungen(256)), z, "Ns.D256")),
		gross(schlecht("66 Typen mit je 1000 Members", eins2(vieleTypen(66, 999)), z, "Ns.T66")),
	}
}

func zeileV2E3(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2E3()) }

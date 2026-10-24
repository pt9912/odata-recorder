package outboundtest

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func schemaAlias(ns, alias, inhalt string) string {
	return `<Schema xmlns="` + edmNs + `" Namespace="` + ns + `" Alias="` + alias + `">` + inhalt + `</Schema>`
}

func ohneKey(name string) string {
	return `<EntityType Name="` + name + `"><Key/>` + prop("Id", "Edm.Int32") + `</EntityType>`
}

func faelleD1() []fall {
	abstrakt := `<EntityType Name="A" Abstract="true"/>`
	out := []fall{
		schlecht("erster Typ", eins(typOhne("A", "")+typ("B", "")+typ("C", "")), model.UrsacheOhneSchluessel, "Ns.A"),
		schlecht("mittlerer Typ", eins(typ("A", "")+typOhne("B", "")+typ("C", "")), model.UrsacheOhneSchluessel, "Ns.B"),
		schlecht("letzter Typ", eins(typ("A", "")+typ("B", "")+typOhne("C", "")), model.UrsacheOhneSchluessel, "Ns.C"),
		schlecht("zweites Schema", []byte(edmx(schema("N1", typ("A", ""))+schema("N2", typOhne("B", "")))), model.UrsacheOhneSchluessel, "N2.B"),
		schlecht("zweites Schema desselben Namensraums", []byte(edmx(schema("Ns", typ("A", ""))+schema("Ns", typOhne("B", "")))), model.UrsacheOhneSchluessel, "Ns.B"),
		schlecht("leerer Key", eins(typ("A", "")+ohneKey("B")), model.UrsacheOhneSchluessel, "Ns.B"),
		schlecht("Key nur mit Kommentar", eins(`<EntityType Name="A"><Key><!-- c --></Key></EntityType>`), model.UrsacheOhneSchluessel, "Ns.A"),
		schlecht("abgeleiteter Typ, Basis ohne Schlüssel", eins(ableitung("D", "Ns.B", "")+typOhne("B", "")), model.UrsacheOhneSchluessel, "Ns.D"),
		gut("erbt eine Stufe", eins(typ("B", "")+ableitung("D", "Ns.B", ""))),
		gut("erbt zwei und drei Stufen", eins(typ("B", "")+ableitung("D1", "Ns.B", "")+ableitung("D2", "Ns.D1", "")+ableitung("D3", "Ns.D2", ""))),
		gut("abstrakter Typ ohne Schlüssel", eins(abstrakt)),
		gut("abstrakte Basis, Ableitung definiert den Schlüssel", eins(abstrakt+ableitung("D", "Ns.A", schluesselID))),
	}
	return out
}

func zeileD1(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleD1()) }

// d2doc ist ein Dokument mit Alias `Self`, einem ComplexType und einem Verweis auf ein externes Dokument.
func d2doc(inhalt string) []byte {
	return []byte(`<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0"><edmx:Reference Uri="u"><edmx:Include Namespace="Ext"/></edmx:Reference>` +
		`<edmx:DataServices>` + schemaAlias("Ns", "Self", `<ComplexType Name="Cx"/>`+inhalt) + `</edmx:DataServices></edmx:Edmx>`)
}

func faelleD2() []fall {
	set := func(ref string) string {
		return typ("T", "") + `<EntityContainer Name="C"><EntitySet Name="S" EntityType="` + ref + `"/></EntityContainer>`
	}
	binding := func(ref string) string {
		return typ("T", "") + `<EntityContainer Name="C"><EntitySet Name="S" EntityType="Ns.T"><NavigationPropertyBinding Path="P" Target="` + ref + `"/></EntitySet></EntityContainer>`
	}
	arten := []struct {
		name, ort string
		inhalt    func(string) string
	}{
		{"BaseType", "Ns.D@BaseType", func(r string) string { return typ("T", "") + ableitung("D", r, "") }},
		{"Typ einer Property", "Ns.T/P@Type", func(r string) string { return typ("T", prop("P", r)) }},
		{"Typ einer Navigation-Property", "Ns.T/N@Type", func(r string) string { return typ("T", `<NavigationProperty Name="N" Type="`+r+`"/>`) }},
		{"Collection einer Navigation-Property", "Ns.T/N@Type", func(r string) string { return typ("T", `<NavigationProperty Name="N" Type="Collection(`+r+`)"/>`) }},
		{"EntityType eines Entity-Sets", "Ns.C/S@EntityType", set},
		{"Target eines Bindings", "Ns.C/S@Target", binding},
	}
	var out []fall
	for _, a := range arten {
		for _, r := range []string{"Self.Nope", "Ns.Nope", "Zzz.T", "T", "Ext.Thing", "Self.Cx"} {
			if r == "Self.Cx" && (a.name == "Typ einer Property" || a.name == "Target eines Bindings") {
				continue
			}
			out = append(out, schlecht(a.name+" "+r, d2doc(a.inhalt(r)), model.UrsacheNichtAufloesbar, a.ort))
		}
	}
	for _, r := range []string{"Nope", "Zzz", "N"} {
		doc := `<EntityType Name="T"><Key><PropertyRef Name="` + r + `"/></Key>` + prop("Id", "Edm.Int32") + `<NavigationProperty Name="N" Type="Ns.T"/></EntityType>`
		out = append(out, schlecht("Name einer PropertyRef "+r, d2doc(doc), model.UrsacheNichtAufloesbar, "Ns.T/"+r+"@Name"))
	}
	return out
}

func zeileD2(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleD2()) }

func kettenDoc(n int, wurzelZuerst bool) []byte { return eins(kettenTypen(n, wurzelZuerst)) }

// kettenTypen sind die Typen T1 bis Tn, jeder die Ableitung des vorigen.
func kettenTypen(n int, wurzelZuerst bool) string {
	var teile []string
	for i := 1; i <= n; i++ {
		if i == 1 {
			teile = append(teile, typ("T1", ""))
		} else {
			teile = append(teile, ableitung(fmt.Sprintf("T%d", i), fmt.Sprintf("Ns.T%d", i-1), ""))
		}
	}
	if !wurzelZuerst {
		slices.Reverse(teile)
	}
	return strings.Join(teile, "")
}

func faelleD3() []fall {
	ab := ableitung
	nA := model.UrsacheNichtAufloesbar
	return []fall{
		schlecht("Selbstbezug", eins(ab("A", "Ns.A", "")), nA, "Ns.A@BaseType"),
		schlecht("Zyklus über zwei Typen", eins(ab("A", "Ns.B", "")+ab("B", "Ns.A", "")), nA, "Ns.A@BaseType"),
		schlecht("Zyklus über drei Typen", eins(ab("A", "Ns.B", "")+ab("B", "Ns.C", "")+ab("C", "Ns.A", "")), nA, "Ns.A@BaseType"),
		schlecht("Schwanz in einen Zyklus", eins(ab("A", "Ns.B", "")+ab("B", "Ns.C", "")+ab("C", "Ns.B", "")), nA, "Ns.A@BaseType"),
		schlecht("Zyklus über zwei Schemas", []byte(edmx(schema("N1", ab("A", "N2.B", ""))+schema("N2", ab("B", "N1.A", "")))), nA, "N1.A@BaseType"),
		gut("Kette von 64, Wurzel zuerst", kettenDoc(64, true)), gut("Kette von 64, Wurzel zuletzt", kettenDoc(64, false)),
		schlecht("Kette von 65, Wurzel zuerst", kettenDoc(65, true), model.UrsacheKette, "Ns.T65"),
		schlecht("Kette von 65, Wurzel zuletzt", kettenDoc(65, false), model.UrsacheKette, "Ns.T65"),
	}
}

func zeileD3(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleD3()) }

const (
	acht   = "Edm.String Edm.Boolean Edm.Int32 Edm.Int64 Edm.Decimal Edm.Guid Edm.Date Edm.DateTimeOffset"
	andere = "Edm.Binary Edm.Byte Edm.SByte Edm.Int16 Edm.Single Edm.Double Edm.Duration Edm.TimeOfDay Edm.Stream " +
		"Edm.Geography Edm.GeographyPoint Edm.GeographyLineString Edm.GeographyPolygon Edm.GeographyMultiPoint Edm.GeographyMultiLineString Edm.GeographyMultiPolygon Edm.GeographyCollection " +
		"Edm.Geometry Edm.GeometryPoint Edm.GeometryLineString Edm.GeometryPolygon Edm.GeometryMultiPoint Edm.GeometryMultiLineString Edm.GeometryMultiPolygon Edm.GeometryCollection Edm.Untyped"
	deklarationen = `<ComplexType Name="Cx"/><EnumType Name="En"/><TypeDefinition Name="Td" UnderlyingType="Edm.String"/>`
)

func faelleD4() []fall {
	var out []fall
	typen := append(strings.Fields(acht+" "+andere), "Collection(Edm.String)", "Ns.Cx", "Ns.En", "Ns.Td")
	for _, ty := range typen {
		schluessel := eins(deklarationen + `<EntityType Name="E"><Key><PropertyRef Name="K"/></Key>` + prop("K", ty) + `</EntityType>`)
		if slices.Contains(strings.Fields(acht), ty) {
			out = append(out, gut("Schlüssel "+ty, schluessel))
		} else {
			out = append(out, schlecht("Schlüssel "+ty, schluessel, model.UrsacheSchluesseltyp, "Ns.E/K"))
		}
		out = append(out, gut("Property "+ty, eins(deklarationen+typ("E", prop("P", ty)))))
	}
	pfad := `<EntityType Name="E"><Key><PropertyRef Name="Addr/City"/></Key>` + prop("Addr", "Ns.Cx") + `</EntityType>`
	return append(out, schlecht("Schlüssel über einen Pfad", eins(deklarationen+pfad), model.UrsacheSchluesseltyp, "Ns.E/Addr/City"))
}

// zeileD4: nur die acht Typen sind als Schlüssel gültig, jeder andere ist schluesseltyp (Nr. 9 f, 10).
func zeileD4(t *testing.T, neu CodecNeu) {
	if got := len(strings.Fields(acht)) + len(strings.Fields(andere)); got != 34 {
		t.Fatalf("%d bekannte Namen, erwartet 34", got)
	}
	if !slices.Equal(model.UnterstuetzteEDMTypen(), strings.Fields(acht)) {
		t.Errorf("UnterstuetzteEDMTypen: %v", model.UnterstuetzteEDMTypen())
	}
	for _, f := range faelleD4() {
		m, _ := pruefeFall(t, neu, f)
		if f.ursache != "" {
			continue
		}
		name, ty, _ := strings.Cut(f.name, " ")
		typen := m.EntityTypes()
		if len(typen) != 1 {
			t.Errorf("%s: %d Typen", f.name, len(typen))
			continue
		}
		if len(typen[0].Properties) == 0 {
			t.Errorf("%s: keine Properties", f.name)
			continue
		}
		p := typen[0].Properties[len(typen[0].Properties)-1]
		wantSupp := slices.Contains(strings.Fields(acht), ty)
		if p.EdmType != ty || p.Unterstuetzt() != wantSupp || (name == "Schlüssel" && !slices.Equal(typen[0].Keys, []string{"K"})) {
			t.Errorf("%s: Property %+v, Keys %v", f.name, p, typen[0].Keys)
		}
	}
}

func faelleD5() []fall {
	doc := func(ty string) []byte {
		return []byte(edmx(schemaAlias("Ns", "Self", deklarationen+`<EnumType Name="Color"/>`+typ("Other", "")+typ("T", prop("P", ty)))))
	}
	out := []fall{gut("Collection(Edm.String)", doc("Collection(Edm.String)")), gut("Collection(Ns.Cx)", doc("Collection(Ns.Cx)")),
		gut("EnumType über Alias", doc("Self.Color")), gut("Entity-Typ als Property-Typ", doc("Ns.Other"))}
	for _, ty := range []string{"Collection(Collection(Edm.String))", "Collection(Self.Nope)", "Collection()", "Collection(Edm.String", "Edm.Foo", "Edm.", "Edm", "edm.string", " Edm.String", "Edm.String "} {
		out = append(out, schlecht("Typ "+ty, doc(ty), model.UrsacheNichtAufloesbar, "Ns.T/P@Type"))
	}
	return out
}

func zeileD5(t *testing.T, neu CodecNeu) {
	faelle := faelleD5()
	pruefeFaelle(t, neu, faelle)
	m, b := pruefeFall(t, neu, faelle[2])
	typen := m.EntityTypes()
	want := model.Fund{Art: model.FundPropertyTyp, Ort: "Ns.T/P", Name: "Ns.Color"}
	if len(typen) < 2 || edmTypVon(typen[1].Properties, 1) != "Ns.Color" || !slices.Contains(b.Funde, want) {
		t.Errorf("EnumType: %+v %v", typen, b.Funde)
	}
}

func repeat(s string, n int) string { return strings.Repeat(s, n) }

// volldok ist ein Dokument mit jedem Pflichtattribut; o ersetzt den Wert eines Attributs, "-" lässt es weg.
func volldgut(o map[string]string) []byte {
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
	return []byte(`<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0"><edmx:DataServices><Schema xmlns="` + edmNs + `"` + a("Schema.Namespace", "Ns") + a("Schema.Alias", "-") + `>` +
		`<EntityType` + a("EntityType.Name", "T") + a("EntityType.Abstract", "-") + a("EntityType.OpenType", "-") + a("EntityType.HasStream", "-") + `>` +
		`<Key><PropertyRef` + a("PropertyRef.Name", "Id") + `/></Key><Property` + a("Property.Name", "Id") + a("Property.Type", "Edm.Int32") + a("Property.Nullable", "-") + `/>` +
		`<NavigationProperty` + a("NavigationProperty.Name", "N") + a("NavigationProperty.Type", "Ns.T") + a("NavigationProperty.ContainsTarget", "-") + `/></EntityType>` +
		`<EntityContainer` + a("EntityContainer.Name", "C") + `><EntitySet` + a("EntitySet.Name", "S") + a("EntitySet.EntityType", "Ns.T") + `>` +
		`<NavigationPropertyBinding` + a("NavigationPropertyBinding.Path", "N") + a("NavigationPropertyBinding.Target", "S") + `/></EntitySet></EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`)
}

func faelleD6() []fall {
	struk := model.UrsacheStruktur
	var out []fall
	for _, p := range []struct{ attr, ort string }{
		{"Schema.Namespace", "@Namespace"}, {"EntityType.Name", "Ns@Name"}, {"PropertyRef.Name", "Ns.T@Name"}, {"Property.Name", "Ns.T@Name"},
		{"Property.Type", "Ns.T/Id@Type"}, {"NavigationProperty.Name", "Ns.T@Name"}, {"NavigationProperty.Type", "Ns.T/N@Type"},
		{"EntityContainer.Name", "Ns@Name"}, {"EntitySet.Name", "Ns.C@Name"}, {"EntitySet.EntityType", "Ns.C/S@EntityType"},
		{"NavigationPropertyBinding.Path", "Ns.C/S@Path"}, {"NavigationPropertyBinding.Target", "Ns.C/S@Target"},
	} {
		out = append(out, schlecht("Pflicht "+p.attr+" fehlt", volldgut(map[string]string{p.attr: "-"}), struk, p.ort),
			schlecht("Pflicht "+p.attr+" leer", volldgut(map[string]string{p.attr: ""}), struk, p.ort))
	}
	for _, w := range []string{"True", "1", "yes", "", " true"} {
		for _, p := range []struct{ attr, ort string }{
			{"EntityType.Abstract", "Ns.T@Abstract"}, {"EntityType.OpenType", "Ns.T@OpenType"}, {"EntityType.HasStream", "Ns.T@HasStream"},
			{"Property.Nullable", "Ns.T/Id@Nullable"}, {"NavigationProperty.ContainsTarget", "Ns.T/N@ContainsTarget"},
		} {
			out = append(out, schlecht(fmt.Sprintf("%s %q", p.attr, w), volldgut(map[string]string{p.attr: w}), struk, p.ort))
		}
	}
	for _, p := range []string{"EntityType.Abstract", "EntityType.OpenType", "EntityType.HasStream", "Property.Nullable", "NavigationProperty.ContainsTarget"} {
		for _, w := range []string{"true", "false"} {
			out = append(out, gut(p+" "+w, volldgut(map[string]string{p: w})))
		}
	}
	return append(out, faelleD6Namen()...)
}

func faelleD6Namen() []fall {
	struk := model.UrsacheStruktur
	var out []fall
	for _, n := range []struct {
		name    string
		gueltig bool
	}{
		{"", false}, {"1abc", false}, {"a-b", false}, {"a b", false}, {"a.b", false}, {" a", false}, {"_x", true},
		{"Größe", true}, {repeat("ä", 128), true}, {repeat("ä", 129), false}, {repeat("a", 128), true}, {repeat("a", 129), false},
	} {
		docs := map[string][]byte{"Ns@Name": eins(typ(n.name, "")), "Ns.T@Name": eins(typ("T", prop(n.name, "Edm.String"))),
			"Ns@Alias": []byte(edmx(schemaAlias("Ns", n.name, "")))}
		for _, ort := range []string{"Ns@Name", "Ns.T@Name", "Ns@Alias"} {
			name := fmt.Sprintf("Name %.12q (%d Bytes) in %s", n.name, len(n.name), ort)
			if n.gueltig {
				out = append(out, gut(name, docs[ort]))
			} else {
				out = append(out, schlecht(name, docs[ort], struk, ort))
			}
		}
	}
	lang := repeat("a", 127) + "." + repeat("b", 128)
	for _, n := range []struct {
		ns      string
		gueltig bool
	}{{lang, true}, {repeat("a", 128) + "." + repeat("b", 128), false}, {"a.b", true}, {"a..b", false}, {".a", false}, {"a.", false}, {"A1.B2", true}} {
		doc := []byte(edmx(schema(n.ns, "")))
		if n.gueltig {
			out = append(out, gut(fmt.Sprintf("Namensraum %.12q (%d)", n.ns, len(n.ns)), doc))
		} else {
			out = append(out, schlecht(fmt.Sprintf("Namensraum %.12q (%d)", n.ns, len(n.ns)), doc, struk, "@Namespace"))
		}
	}
	return append(out, faelleD6Aliase()...)
}

func faelleD6Aliase() []fall {
	struk := model.UrsacheStruktur
	two := func(a, b string) []byte { return []byte(edmx(a + b)) }
	tt := func(ns string) string { return schema(ns, typ("T", "")) }
	return []fall{
		schlecht("Edm als Namespace", []byte(edmx(schema("Edm", ""))), struk, "Edm@Namespace"),
		schlecht("Edm als Alias", []byte(edmx(schemaAlias("Ns", "Edm", ""))), struk, "Ns@Alias"),
		schlecht("Alias doppelt für zwei Namensräume", two(schemaAlias("N1", "X", ""), schemaAlias("N2", "X", "")), struk, "N2@Alias"),
		schlecht("Alias gleich dem späteren Namensraum", two(schemaAlias("N1", "N2", ""), schema("N2", "")), struk, "N2@Namespace"),
		schlecht("Alias gleich dem früheren Namensraum", two(schema("N2", ""), schemaAlias("N1", "N2", "")), struk, "N1@Alias"),
		gut("zwei Aliase desselben Namensraums", two(schemaAlias("N1", "A", ""), schemaAlias("N1", "B", ""))),
		gut("Alias gleich dem eigenen Namensraum", []byte(edmx(schemaAlias("N1", "N1", "")))),
		schlecht("Entity-Typ doppelt", eins(typ("T", "")+typ("T", "")), struk, "Ns.T"),
		schlecht("Entity-Typ doppelt in zwei Schemas", two(tt("Ns"), tt("Ns")), struk, "Ns.T"),
		gut("gleicher Name in zwei Namensräumen", two(tt("N1"), tt("N2"))),
		schlecht("Property doppelt", eins(typ("T", prop("P", "Edm.String")+prop("P", "Edm.String"))), struk, "Ns.T/P"),
		schlecht("Property und NavigationProperty gleichen Namens", eins(typ("T", prop("P", "Edm.String")+`<NavigationProperty Name="P" Type="Ns.T"/>`)), struk, "Ns.T/P"),
		schlecht("EntitySet doppelt", eins(typ("T", "")+`<EntityContainer Name="C"><EntitySet Name="S" EntityType="Ns.T"/><EntitySet Name="S" EntityType="Ns.T"/></EntityContainer>`), struk, "Ns.C/S"),
		schlecht("PropertyRef doppelt", eins(`<EntityType Name="T"><Key><PropertyRef Name="Id"/><PropertyRef Name="Id"/></Key>`+prop("Id", "Edm.Int32")+`</EntityType>`), struk, "Ns.T"),
		schlecht("zwei Key", eins(`<EntityType Name="T"><Key><PropertyRef Name="Id"/></Key><Key><PropertyRef Name="Id"/></Key>`+prop("Id", "Edm.Int32")+`</EntityType>`), struk, "Ns.T"),
		schlecht("zweites EntityContainer im zweiten Schema", two(schema("N1", `<EntityContainer Name="C"/>`), schema("N2", `<EntityContainer Name="D"/>`)), struk, "N2.D"),
	}
}

func zeileD6(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleD6()) }

func faelleD7() []fall {
	struk := model.UrsacheStruktur
	keyAb := `<Key><PropertyRef Name="Id"/></Key>`
	return []fall{
		schlecht("Key in Basis und Ableitung", eins(typ("B", "")+ableitung("D", "Ns.B", keyAb)), struk, "Ns.D"),
		gut("Key nur im abgeleiteten Typ", eins(`<EntityType Name="B" Abstract="true">`+prop("Id", "Edm.Int32")+`</EntityType>`+ableitung("D", "Ns.B", keyAb))),
		schlecht("Key in Wurzel, zweiter und dritter Stufe", eins(typ("W", "")+ableitung("S", "Ns.W", keyAb)+ableitung("T", "Ns.S", keyAb)), struk, "Ns.S"),
		schlecht("Key in Wurzel und dritter Stufe", eins(typ("W", "")+ableitung("S", "Ns.W", "")+ableitung("T", "Ns.S", keyAb)), struk, "Ns.T"),
		schlecht("geerbter Name mit anderem Typ", eins(typ("B", prop("P", "Edm.String"))+ableitung("D", "Ns.B", prop("P", "Edm.Int32"))), struk, "Ns.D/P"),
		schlecht("geerbter Name mit gleichem Typ", eins(typ("B", prop("P", "Edm.String"))+ableitung("D", "Ns.B", prop("P", "Edm.String"))), struk, "Ns.D/P"),
		schlecht("geerbte Navigation-Property als Property", eins(typ("B", `<NavigationProperty Name="P" Type="Ns.B"/>`)+ableitung("D", "Ns.B", prop("P", "Edm.String"))), struk, "Ns.D/P"),
		gut("gleicher Name in zwei Geschwistern", eins(typ("B", "")+ableitung("D1", "Ns.B", prop("P", "Edm.String"))+ableitung("D2", "Ns.B", prop("P", "Edm.String")))),
	}
}

func zeileD7(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleD7()) }

func tiefDoc(n int) string { return verschachtelt(n, `<x:W xmlns:x="urn:x">`, `</x:W>`) }

func faelleD8(t *testing.T) []fall {
	t.Helper()
	ds := dsElement
	nA, struk := model.UrsacheNichtAufloesbar, model.UrsacheStruktur
	doubleKey := func(name string) string {
		return `<EntityType Name="` + name + `"><Key><PropertyRef Name="K"/></Key>` + prop("K", "Edm.Double") + `</EntityType>`
	}
	return []fall{
		schlecht("P1", wurzel(edmxNs, "Edmx", "3.0", "<"), model.UrsacheVersion, ""),
		schlecht("P2", wurzel(edmxV3, "Edmx", "4.0", ""), model.UrsacheVersion, ""),
		schlecht("P3", wurzel(edmxNs, "Edmx", "3.0", ""), model.UrsacheVersion, ""),
		schlecht("P4", append([]byte("<!DOCTYPE x>"), wurzel(edmxNs, "Edmx", "3.0", ds)...), model.UrsacheDoctype, ""),
		schlecht("P5", append([]byte(`<?xml version="1.0" encoding="ISO-8859-1"?><!DOCTYPE x>`), wurzel(edmxNs, "Edmx", "4.0", ds)...), model.UrsacheKodierung, ""),
		schlecht("P6", append(wurzel(edmxNs, "Edmx", "4.0", ""), 'x'), model.UrsacheWurzel, ""),
		schlecht("P7", eins(typ("A", "")+typOhne("1b", "")+"<"), struk, "Ns@Name"),
		schlecht("P8", eins(typ("A", "")+`<EntityType Name="B" x=>`+typOhne("1b", "")), model.UrsacheSyntax, ""),
		schlecht("P9 Tiefe, dann Syntax", eins(tiefDoc(62)+"<"), model.UrsacheTiefe, ""),
		schlecht("P9 Syntax, dann Tiefe", eins("<a b>"+tiefDoc(62)), model.UrsacheSyntax, ""),
		schlecht("P11", eins(typ("A", prop("P", "Ns.Nope"))+typOhne("1b", "")), struk, "Ns@Name"),
		schlecht("P12", eins(typOhne("T", prop("P", "Ns.Nope"))), nA, "Ns.T/P@Type"),
		schlecht("P13", eins(ableitung("A", "Ns.Nope", prop("P", "Ns.Nope2"))), nA, "Ns.A@BaseType"),
		schlecht("P14", eins(doubleKey("A")+typ("B", prop("P", "Ns.Nope"))), nA, "Ns.B/P@Type"),
		schlecht("P15", eins(typOhne("A", "")+doubleKey("B")), model.UrsacheSchluesseltyp, "Ns.B/K"),
		schlecht("P16", eins(typ("T", props(65535)+prop("bad", "Ns.Nope"))), model.UrsacheZuVieleMembers, "Ns.T"),
		schlecht("P17", eins(ableitung("A", "Ns.Nope", schluesselID+props(65535)+prop("x", "Edm.String"))), nA, "Ns.A@BaseType"),
		schlecht("P18", eins(ableitung("A", "Ns.N1", "")+ableitung("B", "Ns.N2", "")), nA, "Ns.A@BaseType"),
		schlecht("P19", eins(kettenTypen(65, true)+ableitung("X", "Ns.Y", "")+ableitung("Y", "Ns.X", "")), model.UrsacheKette, "Ns.T65"),
	}
}

func props(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, `<Property Name="p%d" Type="Edm.String"/>`, i)
	}
	return b.String()
}

// zeileD8: bei mehreren Verstößen meldet der Ablauf den, den er zuerst trifft (Zusatztabelle P1 bis P22).
func zeileD8(t *testing.T, neu CodecNeu) {
	pruefeFaelle(t, neu, faelleD8(t))
	r := teile(eins(typ("T", prop("P", "Ns.Nope"))), 60)
	ctx, abbruch := context.WithCancel(context.Background())
	r.nach = func(n int) {
		if n == 2 {
			abbruch()
		}
	}
	r.mitEnde = true
	if _, _, err := neu.aufruf(ctx, r, 1<<20); err != context.Canceled {
		t.Errorf("P20: %v", err)
	}
	q := &teilReader{teile: [][]byte{[]byte("<a b>")}, mitEnde: true, ende: errQuelle}
	_, _, err := neu.aufruf(context.Background(), q, 1<<20)
	erwarteFehler(t, err, model.UrsacheSyntax, "")
	_, _, err = neu.aufruf(context.Background(), &teilReader{teile: [][]byte{[]byte("<edmx:Edmx xmlns:edmx=")}, ende: errQuelle}, 1<<20)
	if err != errQuelle {
		t.Errorf("P22: %v", err)
	}
}

func zeileD9(t *testing.T, neu CodecNeu) {
	gesehen := map[model.MetadatenUrsache]bool{}
	for _, f := range alleFaelle(t) {
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

func zeileD10(t *testing.T, neu CodecNeu) {
	extra := []fall{
		gut("Facetten", eins(typ("T", `<Property Name="P" Type="Edm.String" MaxLength="10" Precision="3" Scale="1" Unicode="false" SRID="0" DefaultValue="x"/>`))),
		gut("Nullable am Schlüssel", eins(`<EntityType Name="T"><Key><PropertyRef Name="Id"/></Key><Property Name="Id" Type="Edm.Int32" Nullable="true"/></EntityType>`)),
		gut("Set eines abstrakten Typs ohne Schlüssel", eins(`<EntityType Name="T" Abstract="true"/><EntityContainer Name="C"><EntitySet Name="S" EntityType="Ns.T"/></EntityContainer>`)),
	}
	for _, f := range append(alleFaelle(t), extra...) {
		if f.ursache != "" {
			continue
		}
		m, b := pruefeFall(t, neu, f)
		if m.Version() != "4.0" || b.Funde == nil {
			t.Errorf("%s: Version %q, Funde %v", f.name, m.Version(), b.Funde)
		}
	}
}

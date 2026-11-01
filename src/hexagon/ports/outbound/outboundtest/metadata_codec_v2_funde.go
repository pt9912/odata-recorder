package outboundtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// stelleV2 ist eine der vierzehn ausgewerteten Stellen von SPEC-171 Nr. 4: key nennt sie in doc14, ort ist ihr Ort.
type stelleV2 struct{ name, key, ort string }

func stellenV2() []stelleV2 {
	return []stelleV2{
		{"Edmx", "Edmx", ""}, {"DataServices", "DataServices", ""}, {"Schema", "Schema", "Ns"}, {"EntityType", "EntityType", "Ns.T"},
		{"Key", "Key", "Ns.T"}, {"PropertyRef", "PropertyRef", "Ns.T"}, {"Property", "Property", "Ns.T/Id"},
		{"NavigationProperty", "NavigationProperty", "Ns.T/N"}, {"Association", "Association", "Ns.R"}, {"End der Association", "End", "Ns.R"},
		{"EntityContainer", "EntityContainer", "Ns.C"}, {"EntitySet", "EntitySet", "Ns.C/S"}, {"AssociationSet", "AssociationSet", "Ns.C/AS"},
		{"End der AssociationSet", "SetEnd", "Ns.C/AS"},
	}
}

// doc14 ist ein gültiges Dokument ohne Fund, das an jeder der vierzehn Stellen Kinder (kinder) und Attribute (attrs)
// aufnimmt; die Schlüssel nennt stellenV2.
func doc14(kinder, attrs map[string]string) []byte {
	k := func(s string) string { return kinder[s] }
	a := func(s string) string {
		if v := attrs[s]; v != "" {
			return " " + v
		}
		return ""
	}
	return []byte(`<edmx:Edmx xmlns:edmx="` + v2EdmxNs + `" Version="1.0"` + a("Edmx") + `>` + k("Edmx") +
		`<edmx:DataServices` + a("DataServices") + `>` + k("DataServices") +
		`<Schema xmlns="` + v2EdmNs + `" Namespace="Ns"` + a("Schema") + `>` + k("Schema") +
		`<EntityType Name="T"` + a("EntityType") + `>` + k("EntityType") + `<Key` + a("Key") + `>` + k("Key") +
		`<PropertyRef Name="Id"` + a("PropertyRef") + `>` + k("PropertyRef") + `</PropertyRef></Key>` +
		`<Property Name="Id" Type="Edm.Int32"` + a("Property") + `>` + k("Property") + `</Property>` +
		`<NavigationProperty Name="N" Relationship="Ns.R" FromRole="F" ToRole="G"` + a("NavigationProperty") + `>` + k("NavigationProperty") + `</NavigationProperty></EntityType>` +
		`<Association Name="R"` + a("Association") + `>` + k("Association") + `<End Type="Ns.T" Role="F" Multiplicity="*"` + a("End") + `>` + k("End") +
		`</End><End Type="Ns.T" Role="G" Multiplicity="0..1"/></Association>` +
		`<EntityContainer Name="C"` + a("EntityContainer") + `>` + k("EntityContainer") + `<EntitySet Name="S" EntityType="Ns.T"` + a("EntitySet") + `>` + k("EntitySet") + `</EntitySet>` +
		`<AssociationSet Name="AS" Association="Ns.R"` + a("AssociationSet") + `>` + k("AssociationSet") + `<End Role="F" EntitySet="S"` + a("SetEnd") + `>` + k("SetEnd") +
		`</End><End Role="G" EntitySet="S"/></AssociationSet></EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`)
}

// ortKindV2 ist der Ort eines Kindes mit Name `Nm` an einer Stelle: unter Schema und Container mit dem Namen, sonst der der Stelle.
func ortKindV2(s stelleV2) string {
	switch s.key {
	case "Schema":
		return s.ort + ".Nm"
	case "EntityContainer":
		return s.ort + "/Nm"
	}
	return s.ort
}

// konstrukteV2 sind die Konstrukte, die eine Stelle zusätzlich zu `Foo` meldet (SPEC-171 Nr. 4), mit ihrem Namensraum.
func konstrukteV2(key string) []string {
	switch key {
	case "Edmx", "DataServices":
		return []string{"Foo", "edmx:Reference"}
	case "Schema":
		return []string{"Foo", "ComplexType", "Using", "ValueTerm", "Documentation", "FunctionImport"}
	case "Association":
		return []string{"Foo", "ReferentialConstraint", "Documentation"}
	case "End":
		return []string{"Foo", "OnDelete", "Documentation"}
	case "EntityContainer":
		return []string{"Foo", "FunctionImport", "Documentation"}
	}
	return []string{"Foo", "Documentation"}
}

// elementV2 ist das Kindelement name mit Name `Nm` im Namensraum von EDM (oder, mit Präfix `edmx:`, von EDMX) und den
// Attributen attrs (leer oder mit führendem Leerzeichen).
func elementV2(name, attrs string) (xml, lokal string) {
	if l, ok := strings.CutPrefix(name, "edmx:"); ok {
		return `<` + l + ` xmlns="` + v2EdmxNs + `" Name="Nm"` + attrs + `/>`, l
	}
	return `<` + name + ` xmlns="` + v2EdmNs + `" Name="Nm"` + attrs + `/>`, name
}

// attributeKonstruktV2 sind die Attribute, die ein gemeldetes Konstrukt in G1 trägt: zwei mit Namensraum (einer deklariert,
// einer nicht) und zwei des Präfixes `xml`, die kein Fund sind.
const attributeKonstruktV2 = ` xmlns:f="urn:x" f:a="1" sap:label="l" xml:lang="de" xml:space="preserve"`

// g1FaelleV2 ist der Generator über Stelle (14) mal Art (konstrukt ohne und mit Attributen, erweiterung als Element,
// erweiterung als Attribut) mal Name; das Orakel kommt aus den Parametern der Erzeugung (SPEC-171 Nr. 4).
func g1FaelleV2() []g1Fall {
	var out []g1Fall
	for _, s := range stellenV2() {
		for _, k := range konstrukteV2(s.key) {
			xml, lokal := elementV2(k, "")
			out = append(out, g1Fall{fall: gut("Konstrukt "+k+" an "+s.name, doc14(map[string]string{s.key: xml}, nil)),
				funde: []model.Fund{{Art: model.FundKonstrukt, Ort: ortKindV2(s), Name: lokal}}})
			xml, _ = elementV2(k, attributeKonstruktV2)
			out = append(out, g1Fall{fall: gut("Konstrukt "+k+" mit Attributen an "+s.name, doc14(map[string]string{s.key: xml}, nil)),
				funde: []model.Fund{{Art: model.FundKonstrukt, Ort: ortKindV2(s), Name: lokal},
					{Art: model.FundErweiterung, Ort: ortKindV2(s), Name: "{urn:x}a"}, {Art: model.FundErweiterung, Ort: ortKindV2(s), Name: "{sap}label"}}})
		}
		out = append(out,
			g1Fall{fall: gut("Element fremder Namensraum an "+s.name, doc14(map[string]string{s.key: `<f:Foo xmlns:f="urn:x" Name="Nm" f:a="1" g:b="2" xml:lang="de"/>`}, nil)),
				funde: []model.Fund{{Art: model.FundErweiterung, Ort: ortKindV2(s), Name: "{urn:x}Foo"}}},
			g1Fall{fall: gut("Element ohne Namensraum an "+s.name, doc14(map[string]string{s.key: `<Foo xmlns="" Name="Nm"/>`}, nil)),
				funde: []model.Fund{{Art: model.FundErweiterung, Ort: ortKindV2(s), Name: "{}Foo"}}},
			g1Fall{fall: gut("Attribut mit Namensraum an "+s.name, doc14(nil, map[string]string{s.key: `xmlns:f="urn:x" f:a="1"`})),
				funde: []model.Fund{{Art: model.FundErweiterung, Ort: s.ort, Name: "{urn:x}a"}}},
			g1Fall{fall: gut("Attribut mit nicht deklariertem Präfix an "+s.name, doc14(nil, map[string]string{s.key: `sap:label="l"`})),
				funde: []model.Fund{{Art: model.FundErweiterung, Ort: s.ort, Name: "{sap}label"}}},
		)
	}
	return append(out, g1Eigene()...)
}

// g1Eigene sind die Funde V2-eigener Stellen mit fester Erwartung (G6), property_typ und Annotation-lose Inhalte.
func g1Eigene() []g1Fall {
	k := func(art model.FundArt, ort, name string) []model.Fund {
		return []model.Fund{{Art: art, Ort: ort, Name: name}}
	}
	const m = `xmlns:m="` + v2MetaNs + `" `
	mHas := "{" + v2MetaNs + "}HasStream"
	return []g1Fall{
		{gut("ReferentialConstraint an Association", doc14(map[string]string{"Association": `<ReferentialConstraint><Principal Role="F"/></ReferentialConstraint>`}, nil)), k(model.FundKonstrukt, "Ns.R", "ReferentialConstraint")},
		{gut("Documentation an Association", doc14(map[string]string{"Association": `<Documentation><Summary>s</Summary></Documentation>`}, nil)), k(model.FundKonstrukt, "Ns.R", "Documentation")},
		{gut("OnDelete an End", doc14(map[string]string{"End": `<OnDelete Action="Cascade"/>`}, nil)), k(model.FundKonstrukt, "Ns.R", "OnDelete")},
		{gut("FunctionImport am Container", doc14(map[string]string{"EntityContainer": `<FunctionImport Name="F" ReturnType="Edm.String"><Parameter Name="p" Type="Edm.Int32"/></FunctionImport>`}, nil)), k(model.FundKonstrukt, "Ns.C/F", "FunctionImport")},
		{gut("FunctionImport mit m:HttpMethod am Container", doc14(map[string]string{"EntityContainer": `<FunctionImport xmlns:m="` + v2MetaNs + `" Name="F" ReturnType="Edm.String" m:HttpMethod="GET"><Parameter Name="p" Type="Edm.Int32"/></FunctionImport>`}, nil)),
			append(k(model.FundKonstrukt, "Ns.C/F", "FunctionImport"), k(model.FundErweiterung, "Ns.C/F", "{"+v2MetaNs+"}HttpMethod")...)},
		{gut("m:HasStream am Typ", doc14(nil, map[string]string{"EntityType": m + `m:HasStream="true"`})), k(model.FundErweiterung, "Ns.T", mHas)},
		{gut("m:HasStream an der Property", doc14(nil, map[string]string{"Property": m + `m:HasStream="true"`})), k(model.FundErweiterung, "Ns.T/Id", mHas)},
		{gut("sap:label an der Association", doc14(nil, map[string]string{"Association": `xmlns:sap="urn:sap" sap:label="l"`})), k(model.FundErweiterung, "Ns.R", "{urn:sap}label")},
		{gut("sap:label am End", doc14(nil, map[string]string{"End": `xmlns:sap="urn:sap" sap:label="l"`})), k(model.FundErweiterung, "Ns.R", "{urn:sap}label")},
		{gut("sap:label an der AssociationSet", doc14(nil, map[string]string{"AssociationSet": `xmlns:sap="urn:sap" sap:creatable="false"`})), k(model.FundErweiterung, "Ns.C/AS", "{urn:sap}creatable")},
		{gut("sap:label am End der AssociationSet", doc14(nil, map[string]string{"SetEnd": `xmlns:sap="urn:sap" sap:x="1"`})), k(model.FundErweiterung, "Ns.C/AS", "{urn:sap}x")},
		{gut("Property mit Edm.Double", eins2(typ("T", prop("P", "Edm.Double")))), k(model.FundPropertyTyp, "Ns.T/P", "Edm.Double")},
		{gut("Property mit Edm.DateTimeOffset", eins2(typ("T", prop("P", "Edm.DateTimeOffset")))), k(model.FundPropertyTyp, "Ns.T/P", "Edm.DateTimeOffset")},
		{gut("Annotation im Inhalt eines Konstrukts ist Inhalt", doc14(map[string]string{"Schema": `<Function xmlns="` + v2EdmNs + `" Name="Fn"><X/></Function>`}, nil)), k(model.FundKonstrukt, "Ns.Fn", "Function")},
	}
}

func faelleV2G(t *testing.T) []fall {
	t.Helper()
	var out []fall
	for _, g := range g1FaelleV2() {
		out = append(out, g.fall)
	}
	return append(out, gut("ohne Wirkung", doc14(nil, map[string]string{"Property": `MaxLength="10" Precision="3" Scale="1"`})), gut("Reihenfolge", g4DocV2()))
}

// zeileV2G1: jedes Konstrukt der Mengen erscheint genau einmal mit Art, Ort und Name (SPEC-171 Nr. 4).
func zeileV2G1(t *testing.T, neu CodecNeu) {
	if n := len(stellenV2()); n != 14 {
		t.Errorf("%d Stellen, erwartet 14", n)
	}
	faelle := g1FaelleV2()
	if len(faelle) < 14*6 {
		t.Errorf("%d Fälle, erwartet mindestens %d", len(faelle), 14*6)
	}
	for _, g := range faelle {
		_, b := pruefeFall(t, neu, g.fall)
		if !slices.Equal(b.Funde, g.funde) {
			t.Errorf("%s: Funde %v, erwartet %v", g.name, b.Funde, g.funde)
		}
	}
}

// zeileV2G6: die Funde V2-eigener Stellen haben eine feste Erwartung in g1Eigene; sie laufen mit G1.
func zeileV2G6(t *testing.T, neu CodecNeu) {
	for _, g := range g1Eigene() {
		_, b := pruefeFall(t, neu, g.fall)
		if !slices.Equal(b.Funde, g.funde) {
			t.Errorf("%s: Funde %v, erwartet %v", g.name, b.Funde, g.funde)
		}
	}
}

// zeileV2G2: was ohne Wirkung bleibt, ist nie ein Fund (SPEC-171 Nr. 4): Schleife über die acht Attribute ohne Namensraum,
// die drei Attribute des Namensraums m, die zwei des Präfixes xml, Text, Kommentar und Verarbeitungsanweisung, mal vierzehn
// Stellen.
func zeileV2G2(t *testing.T, neu CodecNeu) {
	facetten := strings.Fields("MaxLength Precision Scale Unicode FixedLength ConcurrencyMode Collation StoreGeneratedPattern")
	meta := []string{"DataServiceVersion", "MaxDataServiceVersion", "IsDefaultEntityContainer"}
	zaehler := 0
	for _, s := range stellenV2() {
		var faelle []fall
		for _, f := range facetten {
			faelle = append(faelle, gut(f+" an "+s.name, doc14(nil, map[string]string{s.key: f + `="x"`})))
		}
		for _, mm := range meta {
			wert := "2.0"
			faelle = append(faelle, gut("m:"+mm+" an "+s.name, doc14(nil, map[string]string{s.key: `xmlns:m="` + v2MetaNs + `" m:` + mm + `="` + wert + `"`})))
		}
		for _, x := range []struct{ name, inhalt string }{{"Text", "text"}, {"Kommentar", "<!-- c -->"}, {"Verarbeitungsanweisung", "<?pi x?>"}} {
			faelle = append(faelle, gut(x.name+" an "+s.name, doc14(map[string]string{s.key: x.inhalt}, nil)))
		}
		for _, xa := range []string{`xml:lang="de"`, `xml:space="preserve"`} {
			faelle = append(faelle, gut(xa+" an "+s.name, doc14(nil, map[string]string{s.key: xa})))
		}
		for _, f := range faelle {
			if _, b := pruefeFall(t, neu, f); len(b.Funde) != 0 {
				t.Errorf("%s: Funde %v, erwartet keine", f.name, b.Funde)
			}
			zaehler++
		}
	}
	if zaehler != 14*16 {
		t.Errorf("%d Fälle, erwartet %d", zaehler, 14*16)
	}
}

// zeileV2G3: ein Bericht ohne Fund ist leer und nicht nil, bei jedem Fehler ist er nil (SPEC-125 Nr. 12).
func zeileV2G3(t *testing.T, neu CodecNeu) {
	_, b := pruefeFall(t, neu, gut("K", festesV2(t, "v2_composite.xml")))
	j, err := json.Marshal(b.Funde)
	if b.Funde == nil || len(b.Funde) != 0 || string(j) != "[]" || err != nil {
		t.Errorf("Funde %#v, JSON %s", b.Funde, j)
	}
	for _, f := range alleFaelleV2(t) {
		if f.ursache == "" || f.gross {
			continue
		}
		if _, b, err := lies(neu, f); err == nil || b.Funde != nil {
			t.Errorf("%s: Fehler %v, Funde %v, erwartet Fehler und nil", f.name, err, b.Funde)
		}
	}
}

func g4DocV2() []byte {
	return eins2(`<EntityType xmlns:x="urn:x" xmlns:m="` + v2MetaNs + `" Name="B" x:a="1" m:HasStream="true">` + schluesselID + `<Property Name="Wide" Type="Edm.Double" x:b="2"/>` +
		`<Documentation/><Documentation x:c="3"/></EntityType>` + ableitung("D1", "Ns.B", "") + ableitung("D2", "Ns.B", "") + `<ComplexType Name="Cx"/>` +
		typ("E", prop("P", "Edm.Int64")+`<Documentation/>`))
}

// zeileV2G4: Reihenfolge des Dokuments, keine Zusammenführung, ein Element (ein Typ, eine Property, ein gemeldetes
// Konstrukt) vor seinen Attributen, property_typ an der Stelle seiner Property, vor deren Attributen (SPEC-125 Nr. 12).
func zeileV2G4(t *testing.T, neu CodecNeu) {
	_, b := pruefeFall(t, neu, gut("G4", g4DocV2()))
	gleich(t, "Funde", b.Funde, []model.Fund{
		{Art: model.FundErweiterung, Ort: "Ns.B", Name: "{urn:x}a"}, {Art: model.FundErweiterung, Ort: "Ns.B", Name: "{" + v2MetaNs + "}HasStream"},
		{Art: model.FundPropertyTyp, Ort: "Ns.B/Wide", Name: "Edm.Double"}, {Art: model.FundErweiterung, Ort: "Ns.B/Wide", Name: "{urn:x}b"},
		{Art: model.FundKonstrukt, Ort: "Ns.B", Name: "Documentation"}, {Art: model.FundKonstrukt, Ort: "Ns.B", Name: "Documentation"},
		{Art: model.FundErweiterung, Ort: "Ns.B", Name: "{urn:x}c"}, {Art: model.FundKonstrukt, Ort: "Ns.Cx", Name: "ComplexType"},
		{Art: model.FundKonstrukt, Ort: "Ns.E", Name: "Documentation"},
	})
}

// zeileV2G5: nicht Unterstütztes erscheint nie als unterstützt (LH-RB-01): kein gemeldetes Konstrukt steht im Modell, und
// jede Property mit nicht unterstütztem Typ hat Unterstuetzt() falsch und einen Fund, über alle gültigen Fälle außer den großen.
func zeileV2G5(t *testing.T, neu CodecNeu) {
	zeileV2G5Konstrukte(t, neu)
	for _, f := range alleFaelleV2(t) {
		if f.ursache != "" || f.gross {
			continue
		}
		m, b, err := lies(neu, f)
		if err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		ohne, fund := ohneUnterstuetzung(t, f.name, m), map[string]bool{}
		for _, fu := range b.Funde {
			fund[fu.Name] = fu.Art == model.FundPropertyTyp || fund[fu.Name]
		}
		for ty, nicht := range ohne {
			if nicht != fund[ty] && ty != "Edm.DateTimeOffset(V2)" {
				t.Errorf("%s: Typ %s: nicht unterstützt %v, Fund property_typ %v", f.name, ty, nicht, fund[ty])
			}
		}
	}
}

// zeileV2G5Konstrukte: kein gemeldetes Konstrukt (sieben) steht im Modell, das gegen das Dokument ohne es unverändert ist.
func zeileV2G5Konstrukte(t *testing.T, neu CodecNeu) {
	basis, _, err := lies(neu, gut("Basis", doc14(nil, nil)))
	if err != nil {
		t.Fatal(err)
	}
	konstrukte := map[string]string{"ComplexType": "Schema", "FunctionImport": "EntityContainer", "ReferentialConstraint": "Association",
		"Documentation": "Association", "ValueTerm": "Schema", "Using": "Schema", "OnDelete": "End"}
	if len(konstrukte) != 7 {
		t.Fatalf("%d Konstrukte, erwartet 7", len(konstrukte))
	}
	for name, stelle := range konstrukte {
		m, b, err := lies(neu, gut(name, doc14(map[string]string{stelle: `<` + name + ` xmlns="` + v2EdmNs + `" Name="Nm"/>`}, nil)))
		if err != nil || len(b.Funde) != 1 {
			t.Errorf("%s: %v %v", name, b.Funde, err)
		}
		if fmt.Sprint(m.EntityTypes(), m.EntitySets(), m.NavigationBindings(), m.Namespaces()) != fmt.Sprint(basis.EntityTypes(), basis.EntitySets(), basis.NavigationBindings(), basis.Namespaces()) {
			t.Errorf("%s: das Modell weicht von dem ohne das Konstrukt ab", name)
		}
	}
}

// i1DokV2 hat sechs Namensräume, acht Typen in verwürfelter Reihenfolge, acht Sets und sechs Bindings (zwei AssociationSet
// mit je drei Navigation-Properties); die Erwartung kommt aus der Erzeugung.
func i1DokV2() (doc []byte, funde []model.Fund, typen, sets []string, bindings []model.NavigationBinding) {
	var schemas [6]strings.Builder
	var container strings.Builder
	var imSchema [6][]string
	for i := range 8 {
		n := (i * 3) % 8
		name := fmt.Sprintf("T%d", n)
		imSchema[n%6] = append(imSchema[n%6], name)
		navs := ""
		if i == 0 {
			navs = strings.ReplaceAll(nav2("n1")+nav2("n2")+nav2("n3"), "Ns.R", "N1.R")
		}
		fmt.Fprintf(&schemas[n%6], `<EntityType xmlns:x="urn:x" Name="%s" x:a="1">%s%s%s%s</EntityType>`, name, schluesselID, prop("c", "Edm.Double"), prop("a", "Edm.String")+prop("b", "Edm.Boolean"), navs)
	}
	for s, namen := range imSchema {
		for _, name := range namen {
			qn := fmt.Sprintf("N%d.%s", s+1, name)
			typen = append(typen, qn)
			funde = append(funde, model.Fund{Art: model.FundErweiterung, Ort: qn, Name: "{urn:x}a"}, model.Fund{Art: model.FundPropertyTyp, Ort: qn + "/c", Name: "Edm.Double"})
		}
	}
	first := typen[0]
	schemas[0].WriteString(`<Association Name="R"><End Type="` + first + `" Role="F" Multiplicity="*"/><End Type="` + first + `" Role="G" Multiplicity="*"/></Association>`)
	container.WriteString(`<EntityContainer Name="C">`)
	for _, t := range typen {
		s := "S" + t[strings.Index(t, ".")+1:]
		sets = append(sets, s)
		fmt.Fprintf(&container, `<EntitySet Name="%s" EntityType="%s"/>`, s, t)
	}
	container.WriteString(asetV2("A1", "N1.R", sets[3], sets[5]) + asetV2("A2", "N1.R", sets[6], sets[1]) + `</EntityContainer>`)
	schemas[0].WriteString(container.String())
	for _, a := range []struct{ von, nach string }{{sets[3], sets[5]}, {sets[6], sets[1]}} {
		for _, n := range []string{"n1", "n2", "n3"} {
			bindings = append(bindings, model.NavigationBinding{Set: a.von, Path: n, Target: "N1.C/" + a.nach})
		}
	}
	var b strings.Builder
	for i := range schemas {
		b.WriteString(schema2(fmt.Sprintf("N%d", i+1), schemas[i].String()))
	}
	return []byte(edmx2(b.String())), funde, typen, sets, bindings
}

// zeileV2I1: gleiche Bytes ergeben gleiche Werte in der Reihenfolge des Dokuments (SPEC-125 Nr. 14): 30 Läufe über ein
// Dokument mit acht Typen, acht Sets, sechs Bindings und sechs Namensräumen gegen die Erwartung der Erzeugung.
func zeileV2I1(t *testing.T, neu CodecNeu) {
	doc, funde, typen, sets, bindings := i1DokV2()
	for i := range 30 {
		m, b, err := lies(neu, fall{doc: doc})
		if err != nil {
			t.Fatal(err)
		}
		var namen, gesetzt []string
		for _, e := range m.EntityTypes() {
			namen = append(namen, e.Name)
		}
		for _, s := range m.EntitySets() {
			gesetzt = append(gesetzt, s.Name)
		}
		if !slices.Equal(namen, typen) || !slices.Equal(gesetzt, sets) || !slices.Equal(b.Funde, funde) ||
			!slices.Equal(m.NavigationBindings(), bindings) || len(m.Namespaces()) != 6 {
			t.Fatalf("Lauf %d: Reihenfolge der Typen, Sets, Bindings oder Funde weicht vom Dokument ab: %v", i, m.NavigationBindings())
		}
	}
}

// zeileV2I2: gleichzeitige Aufrufe über alle Fälle außer den großen von E3, dazu Abbruch des Kontexts und Fehler von src;
// der Detektor von make test-race meldet einen ungeordneten Zugriff, den ein Lauf ausführt (SPEC-125 Nr. 14).
func zeileV2I2(t *testing.T, neu CodecNeu) {
	var faelle []fall
	for _, f := range alleFaelleV2(t) {
		if !f.gross {
			faelle = append(faelle, f)
		}
	}
	einzeln := make([]string, len(faelle))
	for i, f := range faelle {
		m, b, err := lies(neu, f)
		einzeln[i] = ergebnis(m, b, err)
	}
	r := festesV2(t, "v2_referenz.xml")
	for _, e := range parallel(16, func(int) []string { return i2LaufV2(neu, faelle, einzeln, r) }) {
		t.Errorf("Ergebnis weicht vom einzelnen Aufruf ab: %s", e)
	}
}

// i2LaufV2 ist der Lauf einer Goroutine von I2: jeder Fall gegen sein Ergebnis aus dem einzelnen Aufruf, dazu der
// Abbruch des Kontexts im ersten Read, der Fehler von src und die Zweige von i2ZusatzV2.
func i2LaufV2(neu CodecNeu, faelle []fall, einzeln []string, r []byte) []string {
	var fehler []string
	for i, f := range faelle {
		if m, b, err := lies(neu, f); ergebnis(m, b, err) != einzeln[i] {
			fehler = append(fehler, f.name)
		}
	}
	ctx, abbruch := context.WithCancel(context.Background())
	rd := teile(r, 100)
	rd.nach = func(int) { abbruch() }
	if _, _, err := neu.aufruf(ctx, rd, 1<<20); err != context.Canceled {
		fehler = append(fehler, fmt.Sprint("Abbruch: ", err))
	}
	quelle := &teilReader{teile: [][]byte{r[:50]}, ende: errQuelle}
	if _, _, err := neu.aufruf(context.Background(), quelle, 1<<20); !errors.Is(err, errQuelle) {
		fehler = append(fehler, fmt.Sprint("Fehler von src: ", err))
	}
	return append(fehler, i2ZusatzV2(neu, r)...)
}

// i2ZusatzV2 führt in jeder Goroutine die Zweige aus, die kein Dokument der Menge erreicht: ungültige Argumente, einen
// Kontext, der vor dem Aufruf oder im letzten Read endet, BOM und UTF-16-Marke in einzelnen Reads, die Summe der Members
// über der Grenze und eine Navigation-Property.
func i2ZusatzV2(neu CodecNeu, r []byte) []string {
	var fehler []string
	hin := context.Background()
	pruefe := func(name string, err error, passt bool) {
		if !passt {
			fehler = append(fehler, fmt.Sprint(name, ": ", err))
		}
	}
	ursache := func(err error, u model.MetadatenUrsache) bool {
		var mf *model.MetadatenFehler
		return errors.As(err, &mf) && mf.Ursache == u
	}
	var keinCtx context.Context
	var keinSrc io.Reader
	_, _, err := neu.aufruf(keinCtx, teile(r), 1<<20)
	pruefe("ctx nil", err, errors.Is(err, model.ErrUngueltig))
	_, _, err = neu.aufruf(hin, keinSrc, 1<<20)
	pruefe("src nil", err, errors.Is(err, model.ErrUngueltig))
	_, _, err = neu.aufruf(hin, teile(r), 0)
	pruefe("grenze 0", err, errors.Is(err, model.ErrUngueltig))
	_, _, err = neu.aufruf(abgebrochen(), teile(r), 1<<20)
	pruefe("Kontext beendet", err, err == context.Canceled)
	ctxEnde, abbruch := context.WithCancel(hin)
	_, _, err = neu.aufruf(ctxEnde, &teilReader{teile: [][]byte{r}, mitEnde: true, nach: func(int) { abbruch() }}, 1<<20)
	pruefe("Kontext endet im letzten Read", err, err == context.Canceled)
	_, _, err = neu.aufruf(hin, teile(append([]byte(bom), r...), 1, 2), 1<<20)
	pruefe("BOM in einzelnen Reads", err, err == nil)
	_, _, err = neu.aufruf(hin, teile([]byte("\xFF\xFE<a/>"), 1), 1<<20)
	pruefe("UTF-16-Marke in einzelnen Reads", err, ursache(err, model.UrsacheKodierung))
	_, _, err = neu.aufruf(hin, teile(eins2(typ("B", props(255))+ableitungen(256))), 1<<20)
	pruefe("Summe der Members über der Grenze", err, ursache(err, model.UrsacheZuVieleMembers))
	_, _, err = neu.aufruf(hin, teile(eins2(typ("T", nav2("N"))+assoz2)), 1<<20)
	pruefe("Navigation-Property", err, err == nil)
	return fehler
}

// zeileV2I3: gleichzeitige Aufrufe mit verschiedenen Ausgängen (gültig, jede Ursache, Abbruch) vermischen sich nicht:
// Fehler, Modell und Bericht jedes Aufrufs gehören seinem Dokument.
func zeileV2I3(t *testing.T, neu CodecNeu) {
	docs := []fall{gut("R", festesV2(t, "v2_referenz.xml")), gut("V", festesV2(t, "v2_vererbung.xml")), gut("K", festesV2(t, "v2_composite.xml")),
		schlecht("ohne Schlüssel", eins2(typOhne("A", "")), model.UrsacheOhneSchluessel, "Ns.A"), schlecht("Syntax", []byte("<a b>"), model.UrsacheSyntax, ""),
		schlecht("Version", wurzelV2(v2EdmxNs, "Edmx", "3.0", dsV2), model.UrsacheVersion, ""), schlecht("nicht auflösbar", eins2(typ("T", nav2("N"))), model.UrsacheNichtAufloesbar, "Ns.T/N@Relationship"),
		gut("Funde", g4DocV2())}
	soll := make([]string, len(docs))
	for i, f := range docs {
		m, b, err := lies(neu, f)
		soll[i] = ergebnis(m, b, err)
	}
	for _, e := range parallel(16, func(int) []string {
		var fehler []string
		for range 50 {
			for i, f := range docs {
				if m, b, err := lies(neu, f); ergebnis(m, b, err) != soll[i] {
					fehler = append(fehler, f.name)
				}
			}
			ctx, abbruch := context.WithCancel(context.Background())
			rd := teile(docs[0].doc, 100)
			rd.nach = func(int) { abbruch() }
			if _, _, err := neu.aufruf(ctx, rd, 1<<20); err != context.Canceled {
				fehler = append(fehler, fmt.Sprint("Abbruch: ", err))
			}
		}
		return fehler
	}) {
		t.Errorf("Ergebnis von %s weicht vom einzelnen Aufruf ab", e)
	}
}

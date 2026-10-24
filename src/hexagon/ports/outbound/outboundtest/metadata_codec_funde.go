package outboundtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// position ist eine Stelle im Dokument, an der ein Kindelement stehen kann; ort ist der Ort des Elternelements.
type position struct {
	name string
	ort  string
	doc  func(kind string) []byte
}

func positionen() []position {
	const set = `<EntitySet Name="S" EntityType="Ns.T">`
	ent := func(inhalt string) []byte { return eins(typ("T", "") + inhalt) }
	return []position{
		{"Schema", "Ns", func(c string) []byte { return eins(c) }},
		{"EntityContainer", "Ns.C", func(c string) []byte { return eins(`<EntityContainer Name="C">` + c + `</EntityContainer>`) }},
		{"NavigationProperty", "Ns.T/N", func(c string) []byte {
			return eins(typ("T", `<NavigationProperty Name="N" Type="Ns.T">`+c+`</NavigationProperty>`))
		}},
		{"EntityType", "Ns.T", func(c string) []byte { return eins(typ("T", c)) }},
		{"Key", "Ns.T", func(c string) []byte {
			return eins(`<EntityType Name="T"><Key><PropertyRef Name="Id"/>` + c + `</Key>` + prop("Id", "Edm.Int32") + `</EntityType>`)
		}},
		{"Property", "Ns.T/P", func(c string) []byte { return eins(typ("T", `<Property Name="P" Type="Edm.String">`+c+`</Property>`)) }},
		{"EntitySet", "Ns.C/S", func(c string) []byte {
			return ent(`<EntityContainer Name="C">` + set + c + `</EntitySet></EntityContainer>`)
		}},
		{"Binding", "Ns.C/S", func(c string) []byte {
			return ent(`<EntityContainer Name="C">` + set + `<NavigationPropertyBinding Path="P" Target="S">` + c + `</NavigationPropertyBinding></EntitySet></EntityContainer>`)
		}},
		{"Edmx", "", func(c string) []byte {
			return []byte(`<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0">` + c + `<edmx:DataServices/></edmx:Edmx>`)
		}},
		{"DataServices", "", func(c string) []byte { return []byte(edmx(c)) }},
		{"PropertyRef", "Ns.T", func(c string) []byte {
			return eins(`<EntityType Name="T"><Key><PropertyRef Name="Id">` + c + `</PropertyRef></Key>` + prop("Id", "Edm.Int32") + `</EntityType>`)
		}},
		{"Annotations", "Ns.T", func(c string) []byte { return eins(`<Annotations Target="Ns.T">` + c + `</Annotations>`) }},
	}
}

// attributOrt ist eine der vierzehn Stellen, an denen ein Attribut mit Namensraum gemeldet wird (Nr. 7): doc setzt attr
// in das Start-Tag des Elements, eigene sind die Funde des Elements selbst, ort ist sein Ort. standard sagt, ob das
// Start-Tag eine Deklaration des Standard-Namensraums zusätzlich tragen darf.
type attributOrt struct {
	name     string
	ort      string
	eigene   []model.Fund
	standard bool
	doc      func(attr string) []byte
}

func attributOrte() []attributOrt {
	kons := func(ort string) []model.Fund {
		return []model.Fund{{Art: model.FundKonstrukt, Ort: ort, Name: "Function"}}
	}
	return []attributOrt{
		{"Edmx", "", nil, true, func(a string) []byte {
			return []byte(`<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0" ` + a + `><edmx:DataServices/></edmx:Edmx>`)
		}},
		{"DataServices", "", nil, true, func(a string) []byte {
			return []byte(`<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0"><edmx:DataServices ` + a + `/></edmx:Edmx>`)
		}},
		{"Schema", "Ns", nil, false, func(a string) []byte {
			return []byte(edmx(`<Schema xmlns="` + edmNs + `" Namespace="Ns" ` + a + `/>`))
		}},
		{"Annotations", "Ns.T", nil, true, func(a string) []byte { return eins(`<Annotations Target="Ns.T" ` + a + `/>`) }},
		{"EntityType", "Ns.T", nil, true, func(a string) []byte { return eins(`<EntityType Name="T" ` + a + `>` + schluesselID + `</EntityType>`) }},
		{"Key", "Ns.T", nil, true, func(a string) []byte {
			return eins(`<EntityType Name="T"><Key ` + a + `><PropertyRef Name="Id"/></Key>` + prop("Id", "Edm.Int32") + `</EntityType>`)
		}},
		{"PropertyRef", "Ns.T", nil, true, func(a string) []byte {
			return eins(`<EntityType Name="T"><Key><PropertyRef Name="Id" ` + a + `/></Key>` + prop("Id", "Edm.Int32") + `</EntityType>`)
		}},
		{"Property", "Ns.T/P", nil, true, func(a string) []byte { return eins(typ("T", `<Property Name="P" Type="Edm.String" `+a+`/>`)) }},
		{"NavigationProperty", "Ns.T/N", nil, true, func(a string) []byte {
			return eins(typ("T", `<NavigationProperty Name="N" Type="Ns.T" `+a+`/>`))
		}},
		{"EntityContainer", "Ns.C", nil, true, func(a string) []byte { return eins(`<EntityContainer Name="C" ` + a + `/>`) }},
		{"EntitySet", "Ns.C/S", nil, true, func(a string) []byte {
			return eins(typ("T", "") + `<EntityContainer Name="C"><EntitySet Name="S" EntityType="Ns.T" ` + a + `/></EntityContainer>`)
		}},
		{"Binding", "Ns.C/S", nil, true, func(a string) []byte {
			return eins(typ("T", "") + `<EntityContainer Name="C"><EntitySet Name="S" EntityType="Ns.T"><NavigationPropertyBinding Path="P" Target="S" ` + a + `/></EntitySet></EntityContainer>`)
		}},
		{"Annotation", "Ns.T", []model.Fund{{Art: model.FundAnnotation, Ort: "Ns.T", Name: "t"}}, true, func(a string) []byte {
			return eins(typ("T", `<Annotation Term="t" `+a+`/>`))
		}},
		{"konstrukt-Element", "Ns.Fn", kons("Ns.Fn"), true, func(a string) []byte { return eins(`<Function Name="Fn" ` + a + `/>`) }},
	}
}

// g1Attribute führt an jeder der vierzehn Stellen das Attribut mit Namensraum (ein Fund am Ort des Elements, nach den
// eigenen Funden des Elements) und die Attribute, die nie ein Fund sind (Deklarationen, Präfix xml).
func g1Attribute() []g1Fall {
	var out []g1Fall
	for _, o := range attributOrte() {
		fund := append(slices.Clone(o.eigene), model.Fund{Art: model.FundErweiterung, Ort: o.ort, Name: "{urn:x}a"})
		out = append(out, g1Fall{fall: gut("Attribut mit Namensraum an "+o.name, o.doc(`xmlns:f="urn:x" f:a="1"`)), funde: fund})
		for _, n := range []string{`xmlns:p="urn:p"`, `xmlns:p="urn:p" xmlns:q="urn:q"`, `xml:lang="de" xml:space="preserve"`, `xmlns:f="urn:x"`} {
			out = append(out, g1Fall{fall: gut("Nicht-Fund "+n+" an "+o.name, o.doc(n)), funde: o.eigene})
		}
		if o.standard {
			out = append(out, g1Fall{fall: gut("Nicht-Fund xmlns an "+o.name, o.doc(`xmlns="`+edmNs+`"`)), funde: o.eigene})
		}
	}
	return out
}

// g1Fremde sind fremde Elemente mit Attribut Name an jeder Position (Ort nach ortKind) und mit Attributen, die kein
// Fund sind, solange das Element selbst der Fund ist, dazu Annotationen im Inhalt eines konstrukt, eines fremden
// Elements und einer Annotation.
func g1Fremde() []g1Fall {
	var out []g1Fall
	for _, p := range positionen() {
		ort := p.ort
		switch p.name {
		case "Schema":
			ort += ".Nm"
		case "EntityContainer":
			ort += "/Nm"
		}
		fremd := model.Fund{Art: model.FundErweiterung, Ort: ort, Name: "{urn:x}Foo"}
		out = append(out, g1Fall{fall: gut("Element fremder Namensraum mit Name in "+p.name, p.doc(`<f:Foo xmlns:f="urn:x" Name="Nm"/>`)), funde: []model.Fund{fremd}},
			g1Fall{fall: gut("Element fremder Namensraum mit Attributen in "+p.name, p.doc(`<f:Foo xmlns:f="urn:x" f:a="1" g:b="2" xml:lang="de"/>`)), funde: []model.Fund{{Art: model.FundErweiterung, Ort: p.ort, Name: "{urn:x}Foo"}}})
	}
	im := func(name, inhalt string, funde ...model.Fund) g1Fall {
		return g1Fall{fall: gut(name, eins(inhalt)), funde: funde}
	}
	ann := `<Annotation Term="inner"/>`
	return append(out,
		im("Annotation im Inhalt eines konstrukt", `<Function Name="Fn">`+ann+`</Function>`, model.Fund{Art: model.FundKonstrukt, Ort: "Ns.Fn", Name: "Function"}),
		im("Annotation im Inhalt eines fremden Elements", `<f:Foo xmlns:f="urn:x">`+ann+`</f:Foo>`, model.Fund{Art: model.FundErweiterung, Ort: "Ns", Name: "{urn:x}Foo"}),
		im("fremdes Element im Inhalt einer Annotation", `<Annotation Term="t"><f:Foo xmlns:f="urn:x"/></Annotation>`, model.Fund{Art: model.FundAnnotation, Ort: "Ns", Name: "t"}),
		im("Annotation mit Term und Target", `<Annotation Term="t" Target="x"/>`, model.Fund{Art: model.FundAnnotation, Ort: "Ns", Name: "t"}),
		im("Annotation mit Name", `<Annotation Name="Nm" Term="t"/>`, model.Fund{Art: model.FundAnnotation, Ort: "Ns", Name: "t"}),
		im("Annotation mit Namen im Container", `<EntityContainer Name="C"><Annotation Name="Nm" Term="t"/></EntityContainer>`, model.Fund{Art: model.FundAnnotation, Ort: "Ns.C", Name: "t"}))
}

// g1Fall ist ein Dokument mit den Funden, die es genau meldet.
type g1Fall struct {
	fall
	funde []model.Fund
}

func konstruktElemente() []string {
	return strings.Fields("ComplexType EnumType TypeDefinition Function Action Term Singleton FunctionImport ActionImport ReferentialConstraint OnDelete Reference Zork")
}

func g1Faelle() []g1Fall {
	var out []g1Fall
	add := func(name string, doc []byte, funde ...model.Fund) {
		out = append(out, g1Fall{fall: gut(name, doc), funde: funde})
	}
	for _, p := range positionen() {
		for _, e := range konstruktElemente() {
			ns := edmNs
			if e == "Reference" {
				ns = edmxNs
			}
			ort, mitName := p.ort, ` Name="Nm"`
			if p.name == "Schema" || p.name == "EntityContainer" {
				ort += map[string]string{"Schema": ".Nm", "EntityContainer": "/Nm"}[p.name]
			}
			add(e+" in "+p.name, p.doc(`<`+e+` xmlns="`+ns+`"`+mitName+`/>`), model.Fund{Art: model.FundKonstrukt, Ort: ort, Name: e})
		}
		add("Annotation in "+p.name, p.doc(`<Annotation xmlns="`+edmNs+`" Term="t"><Annotation Term="inner"/></Annotation>`),
			model.Fund{Art: model.FundAnnotation, Ort: p.ort, Name: "t"})
		add("Element fremder Namensraum in "+p.name, p.doc(`<f:Foo xmlns:f="urn:x"/>`), model.Fund{Art: model.FundErweiterung, Ort: p.ort, Name: "{urn:x}Foo"})
	}
	add("Annotation ohne Term", eins(`<Annotation/>`), model.Fund{Art: model.FundAnnotation, Ort: "Ns", Name: ""})
	add("Annotation in Annotations", eins(`<Annotations Target="Ns.T"><Annotation Term="t"/></Annotations>`), model.Fund{Art: model.FundAnnotation, Ort: "Ns.T", Name: "t"})
	add("Annotations ohne Target", eins(`<Annotations><Annotation Term="t"/></Annotations>`), model.Fund{Art: model.FundAnnotation, Ort: "", Name: "t"})
	add("fremdes Element in Annotations", eins(`<Annotations Target="X"><f:Foo xmlns:f="urn:x"/></Annotations>`), model.Fund{Art: model.FundErweiterung, Ort: "X", Name: "{urn:x}Foo"})
	add("Property mit Edm.Double", eins(typ("T", prop("P", "Edm.Double"))), model.Fund{Art: model.FundPropertyTyp, Ort: "Ns.T/P", Name: "Edm.Double"})
	for _, a := range []struct{ el, attr, ort string }{
		{"EntityType", "OpenType", "Ns.T"}, {"EntityType", "HasStream", "Ns.T"}, {"NavigationProperty", "ContainsTarget", "Ns.T/N"},
	} {
		for _, w := range []string{"true", "false"} {
			doc := volldgut(map[string]string{a.el + "." + a.attr: w})
			var funde []model.Fund
			if w == "true" {
				funde = []model.Fund{{Art: model.FundKonstrukt, Ort: a.ort, Name: a.attr}}
			}
			add(a.attr+" "+w, doc, funde...)
		}
	}
	return append(append(out, g1Attribute()...), g1Fremde()...)
}

func faelleG(t *testing.T) []fall {
	t.Helper()
	var out []fall
	for _, g := range g1Faelle() {
		out = append(out, g.fall)
	}
	return append(out, gut("ohne Wirkung", g2Doc("")), gut("Reihenfolge", g4Doc()))
}

// zeileG1: jedes Konstrukt der Mengen erscheint genau einmal mit Art, Ort und Name (Nr. 7).
func zeileG1(t *testing.T, neu CodecNeu) {
	if n, o := len(positionen()), len(attributOrte()); n != 12 || o != 14 {
		t.Errorf("%d Positionen für Annotationen und %d Orte für Attribute, erwartet 12 und 14", n, o)
	}
	for _, g := range g1Faelle() {
		_, b := pruefeFall(t, neu, g.fall)
		want := g.funde
		if want == nil {
			want = []model.Fund{}
		}
		if !slices.Equal(b.Funde, want) {
			t.Errorf("%s: Funde %v, erwartet %v", g.name, b.Funde, want)
		}
	}
}

func g2Doc(extra string) []byte {
	return eins(`<EntityType Name="T" x="y">` + schluesselID +
		`<Property Name="P" Type="Edm.String" MaxLength="10" Precision="3" Scale="1" Unicode="false" SRID="0" DefaultValue="d" foo="bar"` + extra + `/>` +
		`<NavigationProperty Name="N" Type="Ns.T" Partner="N" Nullable="false"/><!-- c --><?pi x?>text</EntityType>` +
		`<EntityType Name="U"><Key><PropertyRef Name="Id" Alias="I"/></Key>` + prop("Id", "Edm.Int32") + `</EntityType>`)
}

// zeileG2: was ohne Wirkung bleibt, ist nie ein Fund (Nr. 7).
func zeileG2(t *testing.T, neu CodecNeu) {
	_, b := pruefeFall(t, neu, gut("Facetten", g2Doc("")))
	if len(b.Funde) != 0 {
		t.Errorf("Funde %v, erwartet keine", b.Funde)
	}
	_, b = pruefeFall(t, neu, gut("mit sap:", g2Doc(` xmlns:sap="urn:s" sap:label="l"`)))
	gleich(t, "Gegenprobe", b.Funde, []model.Fund{{Art: model.FundErweiterung, Ort: "Ns.T/P", Name: "{urn:s}label"}})
}

// zeileG3: ein Bericht ohne Fund ist leer und nicht nil (Nr. 12).
func zeileG3(t *testing.T, neu CodecNeu) {
	_, b := pruefeFall(t, neu, gut("K", festes(t, "composite.xml")))
	j, err := json.Marshal(b.Funde)
	if b.Funde == nil || len(b.Funde) != 0 || string(j) != "[]" || err != nil {
		t.Errorf("Funde %#v, JSON %s", b.Funde, j)
	}
	_, b, err = lies(neu, fall{doc: []byte("x")})
	if err == nil || b.Funde != nil {
		t.Errorf("Fehlerfall: %v %v", err, b.Funde)
	}
}

func g4Doc() []byte {
	return eins(`<EntityType xmlns:x="urn:x" Name="B" OpenType="true" x:a="1" HasStream="true">` + schluesselID + prop("Wide", "Edm.Double") +
		`<Annotation Term="one"/><Annotation Term="one"/></EntityType>` + ableitung("D1", "Ns.B", "") + ableitung("D2", "Ns.B", "") + ableitung("D3", "Ns.B", "") +
		typ("E", prop("P", "Edm.Int64")+`<Annotation Term="afterP"/>`))
}

// zeileG4: Reihenfolge des Dokuments, keine Zusammenführung, ein Fund je Deklaration (Nr. 12, 7).
func zeileG4(t *testing.T, neu CodecNeu) {
	_, b := pruefeFall(t, neu, gut("G4", g4Doc()))
	gleich(t, "Funde", b.Funde, []model.Fund{
		{Art: model.FundKonstrukt, Ort: "Ns.B", Name: "OpenType"}, {Art: model.FundErweiterung, Ort: "Ns.B", Name: "{urn:x}a"},
		{Art: model.FundKonstrukt, Ort: "Ns.B", Name: "HasStream"}, {Art: model.FundPropertyTyp, Ort: "Ns.B/Wide", Name: "Edm.Double"},
		{Art: model.FundAnnotation, Ort: "Ns.B", Name: "one"}, {Art: model.FundAnnotation, Ort: "Ns.B", Name: "one"},
		{Art: model.FundAnnotation, Ort: "Ns.E", Name: "afterP"},
	})
}

// zeileG5: nicht Unterstütztes erscheint nie als unterstützt, über alle gültigen Fälle außer den großen von E3 (Nr. 10, 12).
func zeileG5(t *testing.T, neu CodecNeu) {
	for _, f := range alleFaelle(t) {
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
			if nicht != fund[ty] {
				t.Errorf("%s: Typ %s: nicht unterstützt %v, Fund property_typ %v", f.name, ty, nicht, fund[ty])
			}
		}
		for ty, ist := range fund {
			if ist && !ohne[ty] {
				t.Errorf("%s: Fund property_typ für %s, den keine nicht unterstützte Property trägt", f.name, ty)
			}
		}
	}
}

// ohneUnterstuetzung liefert je EdmType der Properties, ob sie nicht unterstützt sind, und hält Unterstuetzt() gegen die Liste.
func ohneUnterstuetzung(t *testing.T, name string, m model.ServiceModel) map[string]bool {
	t.Helper()
	ohne := map[string]bool{}
	for _, e := range m.EntityTypes() {
		for _, p := range e.Properties {
			if p.Unterstuetzt() != slices.Contains(model.UnterstuetzteEDMTypen(), p.EdmType) {
				t.Errorf("%s: Unterstuetzt() von %+v", name, p)
			}
			ohne[p.EdmType] = !p.Unterstuetzt() || ohne[p.EdmType]
		}
	}
	return ohne
}

func i1Doc() (doc []byte, funde []model.Fund, typen []string) {
	var b strings.Builder
	for i := range 64 {
		n := (i * 37) % 64
		name := fmt.Sprintf("T%02d", n)
		typen = append(typen, "Ns."+name)
		fmt.Fprintf(&b, `<EntityType xmlns:x="urn:x" Name="%s" x:a="1">%s%s%s<Annotation Term="a%d"/><Function Name="f"/></EntityType>`,
			name, schluesselID, prop("c", "Edm.Double"), prop("a", "Edm.String")+prop("b", "Edm.Boolean"), n)
		funde = append(funde, model.Fund{Art: model.FundErweiterung, Ort: "Ns." + name, Name: "{urn:x}a"},
			model.Fund{Art: model.FundPropertyTyp, Ort: "Ns." + name + "/c", Name: "Edm.Double"},
			model.Fund{Art: model.FundAnnotation, Ort: "Ns." + name, Name: fmt.Sprintf("a%d", n)},
			model.Fund{Art: model.FundKonstrukt, Ort: "Ns." + name, Name: "Function"})
	}
	b.WriteString(`<EntityContainer Name="C">`)
	for _, t := range typen {
		fmt.Fprintf(&b, `<EntitySet Name="S%s" EntityType="%s"/>`, t[3:], t)
	}
	b.WriteString(`</EntityContainer>`)
	return eins(b.String()), funde, typen
}

// zeileI1: gleiche Bytes ergeben gleiche Werte in der Reihenfolge des Dokuments (Nr. 14).
func zeileI1(t *testing.T, neu CodecNeu) {
	doc, funde, typen := i1Doc()
	for i := range 20 {
		m, b, err := lies(neu, fall{doc: doc})
		if err != nil {
			t.Fatal(err)
		}
		var namen, sets []string
		for _, e := range m.EntityTypes() {
			namen = append(namen, e.Name)
			if got := fmt.Sprint(e.Properties); got != "[{Id Edm.Int32 true} {c Edm.Double true} {a Edm.String true} {b Edm.Boolean true}]" {
				t.Fatalf("Lauf %d: Properties %s", i, got)
			}
		}
		for _, s := range m.EntitySets() {
			sets = append(sets, s.Name)
		}
		if !slices.Equal(namen, typen) || len(sets) != 64 || sets[0] != "S"+typen[0][3:] || sets[63] != "S"+typen[63][3:] || !slices.Equal(b.Funde, funde) {
			t.Fatalf("Lauf %d: Reihenfolge der Typen, Sets oder Funde weicht vom Dokument ab", i)
		}
	}
}

// parallel startet n Goroutinen hinter einer Barriere; fehler sammelt, was sie melden.
func parallel(n int, arbeit func(i int) []string) []string {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var fehler []string
	start := make(chan struct{})
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			f := arbeit(i)
			mu.Lock()
			fehler = append(fehler, f...)
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()
	return fehler
}

// zeileI2: gleichzeitige Aufrufe über alle Fälle außer den großen von E3, dazu Abbruch des Kontexts und Fehler von
// src; der Detektor von make test-race meldet einen ungeordneten Zugriff, den ein Lauf ausführt.
func zeileI2(t *testing.T, neu CodecNeu) {
	var faelle []fall
	var soll []string
	for _, f := range alleFaelle(t) {
		if !f.gross {
			faelle = append(faelle, f)
			m, b, err := lies(neu, f)
			soll = append(soll, ergebnis(m, b, err))
		}
	}
	r := festes(t, "referenz.xml")
	for _, e := range parallel(16, func(int) []string {
		var fehler []string
		for i, f := range faelle {
			if m, b, err := lies(neu, f); ergebnis(m, b, err) != soll[i] {
				fehler = append(fehler, f.name)
			}
		}
		ctx, abbruch := context.WithCancel(context.Background())
		rd := teile(r, 100)
		rd.nach = func(int) { abbruch() }
		if _, _, err := neu.aufruf(ctx, rd, 1<<20); err != context.Canceled {
			fehler = append(fehler, fmt.Sprint("Abbruch: ", err))
		}
		if _, _, err := neu.aufruf(context.Background(), &teilReader{teile: [][]byte{r[:50]}, ende: errQuelle}, 1<<20); !errors.Is(err, errQuelle) {
			fehler = append(fehler, fmt.Sprint("Fehler von src: ", err))
		}
		return append(fehler, i2Zusatz(neu, r)...)
	}) {
		t.Errorf("Ergebnis weicht vom einzelnen Aufruf ab: %s", e)
	}
}

// i2Zusatz führt in jeder Goroutine die Zweige aus, die kein Dokument der Menge erreicht: ungültige Argumente, einen
// Kontext, der vor dem Aufruf oder im letzten Read endet, und BOM und UTF-16-Marke in einzelnen Reads.
func i2Zusatz(neu CodecNeu, r []byte) []string {
	var fehler []string
	var keinSrc io.Reader
	kein, hin := func() context.Context { return nil }, context.Background
	ctxEnde, abbruch := context.WithCancel(context.Background())
	ende := func() context.Context { return ctxEnde }
	letzter := &teilReader{teile: [][]byte{r}, mitEnde: true, nach: func(int) { abbruch() }}
	ursache := func(u model.MetadatenUrsache) func(error) bool {
		return func(err error) bool {
			var mf *model.MetadatenFehler
			return errors.As(err, &mf) && mf.Ursache == u
		}
	}
	for _, f := range []struct {
		name  string
		ctx   func() context.Context
		src   io.Reader
		max   int64
		passt func(error) bool
	}{
		{"ctx nil", kein, teile(r), 1 << 20, func(err error) bool { return errors.Is(err, model.ErrUngueltig) }},
		{"src nil", hin, keinSrc, 1 << 20, func(err error) bool { return errors.Is(err, model.ErrUngueltig) }},
		{"max 0", hin, teile(r), 0, func(err error) bool { return errors.Is(err, model.ErrUngueltig) }},
		{"Kontext beendet", abgebrochen, teile(r), 1 << 20, func(err error) bool { return err == context.Canceled }},
		{"Kontext endet im letzten Read", ende, letzter, 1 << 20, func(err error) bool { return err == context.Canceled }},
		{"BOM in einzelnen Reads", hin, teile(append([]byte(bom), r...), 1, 2), 1 << 20, func(err error) bool { return err == nil }},
		{"UTF-16-Marke in einzelnen Reads", hin, teile([]byte("\xFF\xFE<a/>"), 1), 1 << 20, ursache(model.UrsacheKodierung)},
	} {
		if _, _, err := neu.aufruf(f.ctx(), f.src, f.max); !f.passt(err) {
			fehler = append(fehler, fmt.Sprint(f.name, ": ", err))
		}
	}
	return fehler
}

// zeileI3: gleichzeitige Aufrufe mit verschiedenen Dokumenten beeinflussen sich nicht: jedes Ergebnis gleicht dem einzelnen.
func zeileI3(t *testing.T, neu CodecNeu) {
	docs := []fall{gut("R", festes(t, "referenz.xml")), gut("V", festes(t, "vererbung.xml")), gut("K", festes(t, "composite.xml")),
		schlecht("Fehler", eins(typOhne("A", "")), model.UrsacheOhneSchluessel, "Ns.A"), gut("Annotationen", g4Doc())}
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
		}
		return fehler
	}) {
		t.Errorf("Ergebnis von %s weicht vom einzelnen Aufruf ab", e)
	}
}

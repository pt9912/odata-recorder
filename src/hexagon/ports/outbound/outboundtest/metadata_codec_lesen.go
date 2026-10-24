package outboundtest

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	bom      = "\xEF\xBB\xBF"
	erklaert = `<?xml version="1.0" encoding="utf-8"?>`
)

func faelleB1(t *testing.T) []fall {
	t.Helper()
	var out []fall
	for _, f := range []struct{ name, doc string }{
		{"leer", ""}, {"Leerzeichen", " "}, {"Tab", "\t"}, {"LF", "\n"}, {"CR", "\r"}, {"nur BOM", bom},
		{"BOM und Leerraum", bom + " \n"}, {"kein XML", "hallo"}, {"JSON", `{"a":1}`}, {"nur <", "<"},
		{"nur Deklaration", `<?xml version="1.0"?>`}, {"nur Kommentar", "<!-- x -->"},
	} {
		out = append(out, schlecht(f.name, []byte(f.doc), model.UrsacheSyntax, ""))
	}
	return out
}

// zeileB1: leere und unvollständige Eingabe ist syntax (Nr. 3, 4), jedes echte Präfix von R.
func zeileB1(t *testing.T, neu CodecNeu) {
	pruefeFaelle(t, neu, faelleB1(t))
	r := festes(t, "referenz.xml")
	for i := 1; i < len(r); i++ {
		_, _, err := lies(neu, fall{doc: r[:i]})
		erwarteFehler(t, err, model.UrsacheSyntax, "")
	}
}

func koerper() string {
	return edmx(schema("Ns", typ("Größe", prop("Maß", "Edm.String"))))
}

func mit(body, vor, wo, was string) string {
	return vor + strings.Replace(body, wo, was, 1)
}

func faelleB2(t *testing.T) []fall {
	t.Helper()
	body := koerper()
	var out []fall
	for _, f := range []struct{ name, kopf string }{
		{"ohne Deklaration", ""}, {"utf-8", `<?xml version="1.0" encoding="utf-8"?>`}, {"UTF-8", `<?xml version="1.0" encoding="UTF-8"?>`},
		{"Utf-8", `<?xml version="1.0" encoding="Utf-8"?>`}, {"ohne version", `<?xml encoding="utf-8"?>`},
		{"BOM ohne Deklaration", bom}, {"BOM mit Deklaration", bom + erklaert}, {"Zeilenumbruch nach der Deklaration", erklaert + "\n"},
	} {
		out = append(out, gut(f.name, []byte(f.kopf+body)))
	}
	for _, e := range []string{"UTF8", "utf-16", "ISO-8859-1", "windows-1252", "us-ascii", "x-foo"} {
		out = append(out, schlecht("encoding "+e, []byte(`<?xml version="1.0" encoding="`+e+`"?>`+body), model.UrsacheKodierung, ""))
	}
	out = append(out,
		schlecht("UTF-16LE mit BOM", []byte("\xFF\xFE<\x00a\x00/\x00>\x00"), model.UrsacheKodierung, ""),
		schlecht("UTF-16BE mit BOM", []byte("\xFE\xFF\x00<\x00a\x00/\x00>"), model.UrsacheKodierung, ""),
		schlecht("UTF-16LE ohne BOM", []byte("<\x00a\x00/\x00>\x00"), model.UrsacheSyntax, ""),
	)
	for _, f := range []struct{ name, bytes string }{
		{"Byte 0xE4 ohne Deklaration", "\xE4"}, {"Byte 0x01", "\x01"}, {"überlanger UTF-8-Code", "\xC0\xAF"},
		{"Surrogat", "\xED\xA0\x80"}, {"U+FFFE", "\xEF\xBF\xBE"}, {"NUL", "\x00"},
	} {
		out = append(out, schlecht(f.name, []byte(mit(body, "", `Version="4.0"`, `Version="4.0" x="`+f.bytes+`"`)), model.UrsacheSyntax, ""))
	}
	for _, f := range []struct{ name, kopf string }{
		{"version 1.1", `<?xml version="1.1"?>`}, {"Leerraum vor der Deklaration", " " + erklaert},
		{"Deklaration nach Kommentar", "<!-- c -->" + erklaert}, {"zweite Deklaration", erklaert + erklaert},
	} {
		out = append(out, schlecht(f.name, []byte(f.kopf+body), model.UrsacheSyntax, ""))
	}
	return out
}

// zeileB2: UTF-8 mit und ohne Deklaration und BOM, jede andere Kodierung ist kodierung (Nr. 3).
func zeileB2(t *testing.T, neu CodecNeu) {
	faelle := faelleB2(t)
	pruefeFaelle(t, neu, faelle)
	soll, b, err := lies(neu, gut("Basis", []byte(koerper())))
	if err != nil || len(soll.EntityTypes()) != 1 || soll.EntityTypes()[0].Name != "Ns.Größe" {
		t.Fatalf("Basis: %v %v", err, soll.EntityTypes())
	}
	for _, f := range faelle {
		if f.ursache != "" {
			continue
		}
		m, b2, err2 := lies(neu, f)
		if ergebnis(m, b2, err2) != ergebnis(soll, b, err) {
			t.Errorf("%s: Ergebnis weicht von dem ohne Deklaration ab", f.name)
		}
	}
}

func entitaetsAngriff() string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE l [<!ENTITY e0 "x">`)
	for i := 1; i <= 10; i++ {
		fmt.Fprintf(&b, `<!ENTITY e%d "%s">`, i, strings.Repeat(fmt.Sprintf("&e%d;", i-1), 10))
	}
	return b.String() + "]>"
}

func faelleB3(t *testing.T) []fall {
	t.Helper()
	body := koerper()
	var out []fall
	for _, f := range []struct{ name, kopf, verweis string }{
		{"leer", `<!DOCTYPE x>`, ""}, {"interne Entität benutzt", `<!DOCTYPE x [<!ENTITY a "b">]>`, `x="&a;"`},
		{"SYSTEM", `<!DOCTYPE x SYSTEM "file:///etc/hostname">`, ""}, {"PUBLIC", `<!DOCTYPE x PUBLIC "-//X//EN" "http://example.invalid/x.dtd">`, ""},
		{"Parameter-Entität", `<!DOCTYPE x [<!ENTITY % p "q">%p;]>`, ""}, {"Entitätsangriff", entitaetsAngriff(), `x="&e10;"`},
		{"hinter einem Kommentar", `<!-- c --><!DOCTYPE x>`, ""},
	} {
		out = append(out, schlecht("DOCTYPE "+f.name, []byte(mit(body, f.kopf, `Version="4.0"`, `Version="4.0" `+f.verweis)), model.UrsacheDoctype, ""))
	}
	for _, d := range []string{`<!ELEMENT x>`, `<!doctype x>`} {
		out = append(out, schlecht("Direktive "+d, []byte(d+body), model.UrsacheSyntax, ""))
	}
	return out
}

func zeileB3(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleB3(t)) }

const refTerm = `<Annotation Term="%s"/>`

func faelleB4() []fall {
	mitTerm := func(term string) []byte { return eins(typ("T", fmt.Sprintf(refTerm, term))) }
	out := []fall{
		gut("fünf Entitäten und Zeichenreferenzen", mitTerm("a&amp;b&lt;c&gt;d&quot;e&apos;f&#65;g&#x41;h")),
		gut("Grenze unterhalb der Ersatzzeichen", mitTerm("&#xD7FF;")), gut("Grenze oberhalb der Ersatzzeichen", mitTerm("&#xE000;")),
	}
	for _, r := range []string{"&foo;", "&amp", "&#;", "&#0;", "&#1;", "&#xD800;", "&#xDFFF;", "&#55296;", "&#x110000;", "&#X41;"} {
		out = append(out, schlecht("Referenz "+r, mitTerm("x"+r+"y"), model.UrsacheSyntax, ""))
	}
	return append(out, schlecht("Ersatzzeichen im Text", eins(typOhne("T", schluesselID+"&#xD800;")), model.UrsacheSyntax, ""))
}

// zeileB4: die fünf Entitäten und Zeichenreferenzen im Bereich von XML 1.0, jede andere Referenz ist syntax (Nr. 4 b).
func zeileB4(t *testing.T, neu CodecNeu) {
	faelle := faelleB4()
	pruefeFaelle(t, neu, faelle)
	_, b, _ := lies(neu, faelle[0])
	if soll := []model.Fund{{Art: model.FundAnnotation, Ort: "Ns.T", Name: "a&b<c>d\"e'fAgAh"}}; fmt.Sprint(b.Funde) != fmt.Sprint(soll) {
		t.Errorf("Funde %v, erwartet %v", b.Funde, soll)
	}
}

// ref ändert R an einer Stelle.
func aender(r []byte, wo, was string) []byte { return []byte(strings.Replace(string(r), wo, was, 1)) }

const ersteWurzel = "<edmx:Edmx"

func faelleB5(t *testing.T) []fall {
	t.Helper()
	r := festes(t, "referenz.xml")
	ende := "</edmx:Edmx>"
	var out []fall
	for _, f := range []struct {
		name string
		doc  []byte
	}{
		{"Kommentar vor der Wurzel", aender(r, ersteWurzel, "<!-- c -->"+ersteWurzel)},
		{"Kommentar in der Wurzel", aender(r, "<edmx:DataServices>", "<!-- c --><edmx:DataServices>")},
		{"Kommentar nach der Wurzel", append(bytes.Clone(r), "<!-- c -->"...)},
		{"PI vor der Wurzel", aender(r, ersteWurzel, "<?foo bar?>"+ersteWurzel)}, {"PI nach der Wurzel", append(bytes.Clone(r), "<?foo?>"...)},
		{"PI in einem Element", aender(r, "<Key>", "<Key><?foo?>")}, {"xml-stylesheet", aender(r, ersteWurzel, `<?xml-stylesheet href="a"?>`+ersteWurzel)},
		{"CDATA", aender(r, "<Key>", "<Key><![CDATA[a<b&#xD800;]]>")}, {"Text zwischen Elementen", aender(r, "<Key>", "<Key>text")},
		{"Leerraum um die Wurzel", append(aender(r, ersteWurzel, "\n\t"+ersteWurzel), " \n\t"...)},
	} {
		out = append(out, gut(f.name, f.doc))
	}
	zweite := `<edmx:Edmx xmlns:edmx="` + edmxNs + `" Version="4.0"/>`
	for _, f := range []struct {
		name string
		doc  []byte
	}{
		{"Text nach der Wurzel", append(bytes.Clone(r), 'x')}, {"zweites Wurzelelement", append(bytes.Clone(r), zweite...)},
		{"Element nach der Wurzel", append(bytes.Clone(r), "<a/>"...)},
		{"doppeltes Attribut", aender(r, `<Property Name="ID" Type="Edm.Int32"`, `<Property Name="ID" Type="Edm.Int32" Name="x"`)},
		{"ungleiche End-Tags", aender(r, "</Key>", "</Kex>")}, {"Attribut ohne Anführungszeichen", aender(r, `Type="Edm.Int32"`, "Type=Edm.Int32")},
		{"Direktive <!FOO>", aender(r, "<Key>", "<!FOO><Key>")}, {"End-Tag der Wurzel fehlt", bytes.TrimSuffix(r, []byte(ende))},
	} {
		out = append(out, schlecht(f.name, f.doc, model.UrsacheSyntax, ""))
	}
	return out
}

// zeileB5: Kommentare, Verarbeitungsanweisungen, CDATA und Text sind ohne Wirkung; Nachlauf und doppelte Attribute sind syntax.
func zeileB5(t *testing.T, neu CodecNeu) {
	faelle := faelleB5(t)
	pruefeFaelle(t, neu, faelle)
	m, b, err := lies(neu, gut("R", festes(t, "referenz.xml")))
	for _, f := range faelle {
		if f.ursache != "" {
			continue
		}
		m2, b2, err2 := lies(neu, f)
		if ergebnis(m2, b2, err2) != ergebnis(m, b, err) {
			t.Errorf("%s: Ergebnis weicht von R ab", f.name)
		}
	}
}

func faelleB6() []fall {
	return []fall{gut("Standard-Namensraum", eins(typ("T", prop("P", "Edm.String")))), gut("Präfix edm", []byte(edmx(
		`<edm:Schema xmlns:edm="`+edmNs+`" Namespace="Ns"><edm:EntityType Name="T"><edm:Key><edm:PropertyRef Name="Id"/></edm:Key>`+
			`<edm:Property Name="Id" Type="Edm.Int32"/><edm:Property Name="P" Type="Edm.String"/></edm:EntityType></edm:Schema>`)))}
}

func zeileB6(t *testing.T, neu CodecNeu) {
	faelle := faelleB6()
	m0, b0, err0 := lies(neu, faelle[0])
	m1, b1, err1 := lies(neu, faelle[1])
	if err0 != nil || len(m0.EntityTypes()) != 1 || ergebnis(m0, b0, err0) != ergebnis(m1, b1, err1) {
		t.Errorf("Präfix und Standard-Namensraum ergeben verschiedene Ergebnisse: %v / %v", ergebnis(m0, b0, err0), ergebnis(m1, b1, err1))
	}
	for _, f := range []struct{ name, inhalt, ort, fund string }{
		{"Element ohne Präfix-Deklaration", `<EntityType Name="T">` + schluesselID + `<x:Property/></EntityType>`, "Ns.T", "{x}Property"},
		{"Attribut ohne Deklaration", typ("T", `<Property Name="P" Type="Edm.String" sap:label="a"/>`), "Ns.T/P", "{sap}label"},
		{"Attribut mit Deklaration", typ("T", `<Property xmlns:sap="urn:s" Name="P" Type="Edm.String" sap:label="a"/>`), "Ns.T/P", "{urn:s}label"},
	} {
		_, b, err := lies(neu, gut(f.name, eins(f.inhalt)))
		if err != nil || len(b.Funde) < 1 || b.Funde[len(b.Funde)-1] != (model.Fund{Art: model.FundErweiterung, Ort: f.ort, Name: f.fund}) {
			t.Errorf("%s: %v %v", f.name, b.Funde, err)
		}
	}
	v3 := `<Schema xmlns="http://schemas.microsoft.com/ado/2008/09/edm" Namespace="Ns"/>`
	m, b, err := lies(neu, gut("Schema von V3", []byte(edmx(v3))))
	if err != nil || len(m.EntityTypes()) != 0 || len(m.Namespaces()) != 0 || len(b.Funde) != 1 || b.Funde[0].Name != "{http://schemas.microsoft.com/ado/2008/09/edm}Schema" {
		t.Errorf("Schema von V3: %v %v %v", m.Namespaces(), b.Funde, err)
	}
}

// dsElement ist ein DataServices im richtigen Namensraum, gleich welchen Namensraum die Wurzel trägt.
const dsElement = `<d:DataServices xmlns:d="` + edmxNs + `"/>`

func wurzel(ns, name, version string, kinder string) []byte {
	v := ""
	if version != "-" {
		v = ` Version="` + version + `"`
	}
	return []byte(`<edmx:` + name + ` xmlns:edmx="` + ns + `"` + v + `>` + kinder + `</edmx:` + name + `>`)
}

func faelleC1() []fall {
	ds := dsElement
	out := []fall{gut("V4", wurzel(edmxNs, "Edmx", "4.0", ds))}
	for _, v := range []string{"4.0", "3.0", "1.0", "-"} {
		out = append(out, schlecht("V3-Namensraum, Version "+v, wurzel(edmxV3, "Edmx", v, ds), model.UrsacheVersion, ""))
	}
	fremd := `<x:DataServices xmlns:x="` + edmxNs + `x"/>`
	for _, f := range []struct {
		name string
		doc  []byte
	}{
		{"Edmx ohne Namensraum", []byte(`<Edmx Version="4.0">` + ds + `</Edmx>`)}, {"Foo im V4-Namensraum", wurzel(edmxNs, "Foo", "4.0", ds)},
		{"edmx klein", wurzel(edmxNs, "edmx", "4.0", ds)}, {"Edmx in example.com", wurzel("http://example.com/", "Edmx", "4.0", ds)},
		{"ohne DataServices", wurzel(edmxNs, "Edmx", "4.0", "")}, {"zwei DataServices", wurzel(edmxNs, "Edmx", "4.0", ds+ds)},
		{"DataServices im fremden Namensraum", wurzel(edmxNs, "Edmx", "4.0", fremd)},
	} {
		out = append(out, schlecht(f.name, f.doc, model.UrsacheWurzel, ""))
	}
	return append(out, gut("Reference vor DataServices", wurzel(edmxNs, "Edmx", "4.0", `<edmx:Reference Uri="x"><edmx:Include Namespace="Other"/></edmx:Reference>`+ds)),
		gut("ohne Schema", wurzel(edmxNs, "Edmx", "4.0", ds)))
}

func zeileC1(t *testing.T, neu CodecNeu) {
	faelle := faelleC1()
	pruefeFaelle(t, neu, faelle)
	m, b := pruefeFall(t, neu, faelle[len(faelle)-1])
	if m.Version() != "4.0" || len(m.EntityTypes()) != 0 || len(b.Funde) != 0 {
		t.Errorf("ohne Schema: %q %v %v", m.Version(), m.EntityTypes(), b.Funde)
	}
	_, b = pruefeFall(t, neu, faelle[len(faelle)-2])
	if len(b.Funde) != 1 || b.Funde[0] != (model.Fund{Art: model.FundKonstrukt, Name: "Reference"}) {
		t.Errorf("Reference: %v", b.Funde)
	}
}

func faelleC2() []fall {
	ds := dsElement
	out := []fall{gut("4.0", wurzel(edmxNs, "Edmx", "4.0", ds))}
	for _, v := range []string{"-", "", "4.01", "4.00", "4", "1.0", "3.0", "4.0.0", " 4.0", "4.0 ", "04.0", "4,0"} {
		out = append(out, schlecht(fmt.Sprintf("Version %q", v), wurzel(edmxNs, "Edmx", v, ds), model.UrsacheVersion, ""))
	}
	return out
}

func zeileC2(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleC2()) }

// auffuellen hängt Leerraum an, bis das Dokument die Länge n hat.
func auffuellen(doc []byte, n int) []byte {
	return append(bytes.Clone(doc), bytes.Repeat([]byte(" "), n-len(doc))...)
}

func faelleE1(t *testing.T) []fall {
	t.Helper()
	r := festes(t, "referenz.xml")
	l := int64(len(r) + 40)
	voll := auffuellen(r, int(l))
	mitBom := append([]byte(bom), voll...)
	return []fall{
		{name: "max gleich L-1", doc: voll, max: l - 1, ursache: model.UrsacheZuGross},
		{name: "max gleich L", doc: voll, max: l}, {name: "max gleich L+1", doc: voll, max: l + 1},
		{name: "max 1, Dokument von 2 Bytes", doc: []byte("<a"), max: 1, ursache: model.UrsacheZuGross},
		{name: "max MaxInt64", doc: r, max: math.MaxInt64},
		{name: "Syntaxfehler in Byte 5 vor max+1", doc: auffuellen([]byte("<a b>"), 15+10), max: 15, ursache: model.UrsacheSyntax},
		{name: "gültiges Präfix von max Bytes, danach Müll", doc: append(bytes.Clone(r), 'x'), max: int64(len(r)), ursache: model.UrsacheZuGross},
		{name: "das BOM zählt mit", doc: mitBom, max: l + 2, ursache: model.UrsacheZuGross},
		{name: "mit BOM an der Grenze", doc: mitBom, max: l + 3},
	}
}

func zeileE1(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleE1(t)) }

// verschachtelt legt n Ebenen von öffnend/schliessend um inhalt.
func verschachtelt(n int, auf, zu string) string {
	return strings.Repeat(auf, n) + strings.Repeat(zu, n)
}

func faelleE2() []fall {
	var out []fall
	for _, w := range []struct{ name, auf, zu string }{
		{"fremde Elemente", `<x:W xmlns:x="urn:x">`, `</x:W>`}, {"Annotation mit Record-Kindern", `<R>`, `</R>`},
		{"unbekannte Edm-Elemente", `<Function>`, `</Function>`},
	} {
		wrap := func(n int) []byte {
			inhalt := verschachtelt(n, w.auf, w.zu)
			if w.name == "Annotation mit Record-Kindern" {
				inhalt = `<Annotation Term="t">` + inhalt + `</Annotation>`
			}
			return []byte(edmx(`<Schema xmlns="` + edmNs + `" Namespace="Ns">` + inhalt + `</Schema>`))
		}
		anz := 61
		if w.name == "Annotation mit Record-Kindern" {
			anz = 60
		}
		out = append(out, gut(w.name+" bei Tiefe 64", wrap(anz)), schlecht(w.name+" bei Tiefe 65", wrap(anz+1), model.UrsacheTiefe, ""))
	}
	return append(out, gut("70 flache Geschwister", eins(strings.Repeat(`<Function/>`, 70))))
}

func zeileE2(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleE2()) }

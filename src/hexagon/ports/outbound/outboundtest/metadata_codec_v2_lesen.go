package outboundtest

import (
	"bytes"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func koerperV2() string {
	return edmx2(schema2("Ns", typ("Größe", prop("Maß", "Edm.String"))))
}

func faelleV2B1(t *testing.T) []fall {
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

// zeileV2B1: leere und unvollständige Eingabe ist syntax (SPEC-125 Nr. 3, 4): ein Fall je Element der Aufzählung und
// jedes echte Präfix des Referenzdokuments.
func zeileV2B1(t *testing.T, neu CodecNeu) {
	pruefeFaelle(t, neu, faelleV2B1(t))
	r := festesV2(t, "v2_referenz.xml")
	for i := 1; i < len(r); i++ {
		_, _, err := lies(neu, fall{doc: r[:i]})
		erwarteFehler(t, err, model.UrsacheSyntax, "")
	}
}

func faelleV2B2(t *testing.T) []fall {
	t.Helper()
	body := koerperV2()
	var out []fall
	for _, f := range []struct{ name, kopf string }{
		{"ohne Deklaration", ""}, {"utf-8", `<?xml version="1.0" encoding="utf-8"?>`}, {"UTF-8", `<?xml version="1.0" encoding="UTF-8"?>`},
		{"Utf-8", `<?xml version="1.0" encoding="Utf-8"?>`}, {"uTF-8", `<?xml version="1.0" encoding="uTF-8"?>`},
		{"einfache Anführungszeichen", `<?xml version='1.0' encoding='utf-8'?>`}, {"ohne version", `<?xml encoding="utf-8"?>`},
		{"BOM ohne Deklaration", bom}, {"BOM mit Deklaration", bom + erklaert}, {"Zeilenumbruch nach der Deklaration", erklaert + "\n"},
	} {
		out = append(out, gut(f.name, []byte(f.kopf+body)))
	}
	for _, e := range []string{"UTF8", "utf-16", "ISO-8859-1", "windows-1252", "us-ascii", "x-foo", "utf-8x", " utf-8"} {
		out = append(out, schlecht("encoding "+e, []byte(`<?xml version="1.0" encoding="`+e+`"?>`+body), model.UrsacheKodierung, ""))
	}
	out = append(out,
		schlecht("UTF-16LE mit BOM", []byte("\xFF\xFE<\x00a\x00/\x00>\x00"), model.UrsacheKodierung, ""),
		schlecht("UTF-16BE mit BOM", []byte("\xFE\xFF\x00<\x00a\x00/\x00>"), model.UrsacheKodierung, ""),
		schlecht("UTF-16LE ohne BOM", []byte("<\x00a\x00/\x00>\x00"), model.UrsacheSyntax, ""),
		schlecht("0xFF am Anfang", []byte("\xFF\xFF"+body), model.UrsacheSyntax, ""),
		schlecht("BOM doppelt", []byte(bom+bom+body), model.UrsacheSyntax, ""),
		schlecht("UTF-8-Zeichen am Ende abgeschnitten", []byte(body+"\xC3"), model.UrsacheSyntax, ""),
	)
	for _, z := range zeichen() {
		for _, o := range []struct{ name, wo, was string }{
			{"im Attribut", `Version="1.0"`, `Version="1.0" x="` + z.bytes + `"`},
			{"im Kommentar", "<edmx:DataServices>", "<!--" + z.bytes + "--><edmx:DataServices>"},
			{"im Text", "<edmx:DataServices>", "<edmx:DataServices>" + z.bytes},
		} {
			doc := []byte(mit(body, "", o.wo, o.was))
			if z.gueltig {
				out = append(out, gut("Zeichen "+z.name+" "+o.name, doc))
			} else {
				out = append(out, schlecht("Zeichen "+z.name+" "+o.name, doc, model.UrsacheSyntax, ""))
			}
		}
	}
	for _, f := range []struct{ name, kopf string }{
		{"version 1.1", `<?xml version="1.1"?>`}, {"Leerraum vor der Deklaration", " " + erklaert},
		{"Deklaration nach Kommentar", "<!-- c -->" + erklaert}, {"zweite Deklaration", erklaert + erklaert},
	} {
		out = append(out, schlecht(f.name, []byte(f.kopf+body), model.UrsacheSyntax, ""))
	}
	return out
}

// zeileV2B2: UTF-8 mit und ohne Deklaration und BOM, jede andere Kodierung ist kodierung (SPEC-125 Nr. 3).
func zeileV2B2(t *testing.T, neu CodecNeu) {
	faelle := faelleV2B2(t)
	pruefeFaelle(t, neu, faelle)
	soll, b, err := lies(neu, gut("Basis", []byte(koerperV2())))
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

func faelleV2B3(t *testing.T) []fall {
	t.Helper()
	body := koerperV2()
	var out []fall
	for _, f := range []struct{ name, kopf, verweis string }{
		{"leer", `<!DOCTYPE x>`, ""}, {"interne Entität benutzt", `<!DOCTYPE x [<!ENTITY a "b">]>`, `x="&a;"`},
		{"SYSTEM", `<!DOCTYPE x SYSTEM "file:///etc/hostname">`, ""}, {"PUBLIC", `<!DOCTYPE x PUBLIC "-//X//EN" "http://example.invalid/x.dtd">`, ""},
		{"Parameter-Entität", `<!DOCTYPE x [<!ENTITY % p "q">%p;]>`, ""}, {"Entitätsangriff", entitaetsAngriff(), `x="&e10;"`},
		{"hinter einem Kommentar", `<!-- c --><!DOCTYPE x>`, ""}, {"nach der Deklaration", erklaert + `<!DOCTYPE x>`, ""},
	} {
		out = append(out, schlecht("DOCTYPE "+f.name, []byte(mit(body, f.kopf, `Version="1.0"`, `Version="1.0" `+f.verweis)), model.UrsacheDoctype, ""))
	}
	for _, d := range []string{`<!ELEMENT x>`, `<!doctype x>`, `<!DOCTYP x>`} {
		out = append(out, schlecht("Direktive "+d, []byte(d+body), model.UrsacheSyntax, ""))
	}
	return out
}

func zeileV2B3(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2B3(t)) }

// refFaelleV2 sind die Werte von B4: gültig sagt, ob die Referenz zulässig ist.
func refFaelleV2() []struct {
	name, wert string
	gueltig    bool
} {
	out := []struct {
		name, wert string
		gueltig    bool
	}{
		{"U+FFFD als Zeichen", "�", true}, {"U+FFFD und Referenz mit führenden Nullen", "�&#0000000065;", true},
		{"fünf Entitäten und Zeichenreferenzen", "a&amp;b&lt;c&gt;d&quot;e&apos;f&#65;g&#x41;h", true},
		{"Grenze unterhalb der Ersatzzeichen", "&#xD7FF;", true}, {"Grenze oberhalb der Ersatzzeichen", "&#xE000;", true},
		{"Grenze FFFD", "&#xFFFD;", true}, {"Ersatzzeichen mit führenden Nullen", "&#0000055296;", false},
		{"Ersatzzeichen neben U+FFFD", "�&#xD800;", false}, {"Ersatzzeichen nach gültiger Referenz", "&#65;&#xD800;", false},
		{"Ersatzzeichen vor gültiger Referenz", "&#xD800;&#65;", false}, {"obere Grenze der Ersatzzeichen", "&#xDFFF;", false},
	}
	for _, r := range []string{"&foo;", "&amp", "&#;", "&#0;", "&#1;", "&#xD800;", "&#55296;", "&#x110000;", "&#X41;"} {
		out = append(out, struct {
			name, wert string
			gueltig    bool
		}{"Referenz " + r, "x" + r + "y", false})
	}
	return out
}

func faelleV2B4() []fall {
	var out []fall
	add := func(name string, doc []byte, gueltig bool) {
		if gueltig {
			out = append(out, gut(name, doc))
		} else {
			out = append(out, schlecht(name, doc, model.UrsacheSyntax, ""))
		}
	}
	for _, s := range stellenV2() {
		for _, r := range refFaelleV2() {
			add("Referenz in "+s.name+": "+r.name, doc14(nil, map[string]string{s.key: `x="` + r.wert + `"`}), r.gueltig)
		}
	}
	for _, r := range refFaelleV2() {
		add("Referenz im Text: "+r.name, eins2(typOhne("T", schluesselID+r.wert)), r.gueltig)
	}
	for _, x := range []struct {
		name, inhalt string
		gueltig      bool
	}{
		{"CDATA mit Ersatzreferenz", "<![CDATA[&#xD800;�]]>&#65;", true}, {"Ersatzreferenz nach CDATA", "<![CDATA[x]]>&#xD800;", false},
		{"Ersatzreferenz vor CDATA", "&#xD800;<![CDATA[x]]>", false}, {"Ersatzreferenz nach Kommentar", "&#65;<!-- c -->&#xD800;", false},
		{"U+FFFD und Referenzen neben CDATA", "�&#xFFFD;<![CDATA[&#xD800;�]]>&#65;", true},
	} {
		add("Text "+x.name, eins2(typOhne("T", schluesselID+x.inhalt)), x.gueltig)
	}
	return out
}

// zeileV2B4: die fünf Entitäten und Zeichenreferenzen im Bereich von XML 1.0, jede andere Referenz ist syntax, in jedem
// Attributwert jedes ausgewerteten Elements (vierzehn Stellen) und im Text (SPEC-125 Nr. 4 b).
func zeileV2B4(t *testing.T, neu CodecNeu) {
	faelle := faelleV2B4()
	pruefeFaelle(t, neu, faelle)
	if n := len(stellenV2()); n != 14 || len(faelle) < n*len(refFaelleV2()) {
		t.Errorf("%d Stellen und %d Fälle", n, len(faelle))
	}
}

func faelleV2B5(t *testing.T) []fall {
	t.Helper()
	r := festesV2(t, "v2_referenz.xml")
	ende := "</edmx:Edmx>"
	var out []fall
	for _, f := range []struct {
		name string
		doc  []byte
	}{
		{"Kommentar vor der Wurzel", aender(r, ersteWurzel, "<!-- c -->"+ersteWurzel)},
		{"Kommentar in der Wurzel", aender(r, "<edmx:DataServices", "<!-- c --><edmx:DataServices")},
		{"Kommentar nach der Wurzel", append(bytes.Clone(r), "<!-- c -->"...)},
		{"PI vor der Wurzel", aender(r, ersteWurzel, "<?foo bar?>"+ersteWurzel)}, {"PI nach der Wurzel", append(bytes.Clone(r), "<?foo?>"...)},
		{"PI in einem Element", aender(r, "<Key>", "<Key><?foo?>")}, {"xml-stylesheet", aender(r, ersteWurzel, `<?xml-stylesheet href="a"?>`+ersteWurzel)},
		{"CDATA", aender(r, "<Key>", "<Key><![CDATA[a<b�&#xD800;]]>")}, {"Text zwischen Elementen", aender(r, "<Key>", "<Key>text")},
		{"Leerraum um die Wurzel", append(aender(r, ersteWurzel, "\n\t"+ersteWurzel), " \n\t"...)},
		{"CR um die Wurzel", append(aender(r, ersteWurzel, "\r"+ersteWurzel), "\r"...)}, {"CR LF um die Wurzel", append(aender(r, ersteWurzel, "\r\n"+ersteWurzel), "\r\n"...)},
	} {
		out = append(out, gut(f.name, f.doc))
	}
	for _, w := range []struct{ name, zeichen string }{{"Leerzeichen", " "}, {"Tabulator", "\t"}, {"CR", "\r"}, {"LF", "\n"}} {
		out = append(out, gut(w.name+" nach der Wurzel", append(bytes.Clone(r), w.zeichen...)), gut(w.name+" vor der Wurzel", aender(r, ersteWurzel, w.zeichen+ersteWurzel)))
	}
	zweite := `<edmx:Edmx xmlns:edmx="` + v2EdmxNs + `" Version="1.0"/>`
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
	return append(out, faelleV2B5Nachlauf(r)...)
}

// faelleV2B5Nachlauf sind CDATA, Referenzen und Nicht-Leerraum vor und nach der Wurzel (SPEC-125 Nr. 4).
func faelleV2B5Nachlauf(r []byte) []fall {
	var out []fall
	for _, c := range []struct{ name, inhalt string }{{"leer", "<![CDATA[]]>"}, {"nur Leerraum", "<![CDATA[ ]]>"}, {"mit Text", "<![CDATA[x]]>"}} {
		out = append(out, schlecht("CDATA "+c.name+" nach der Wurzel", append(bytes.Clone(r), c.inhalt...), model.UrsacheSyntax, ""),
			schlecht("CDATA "+c.name+" vor der Wurzel hinter der Deklaration", aender(r, ersteWurzel, c.inhalt+ersteWurzel), model.UrsacheSyntax, ""),
			schlecht("CDATA "+c.name+" hinter einem Kommentar vor der Wurzel", aender(r, ersteWurzel, "<!-- c -->"+c.inhalt+ersteWurzel), model.UrsacheSyntax, ""))
	}
	for _, ref := range []string{"&#32;", "&#9;", "&#xA;", "&#xD;", "&#xA0;", "&amp;", "&lt;"} {
		out = append(out, schlecht("Referenz "+ref+" nach der Wurzel", append(bytes.Clone(r), ref...), model.UrsacheSyntax, ""),
			schlecht("Referenz "+ref+" vor der Wurzel hinter der Deklaration", aender(r, ersteWurzel, ref+ersteWurzel), model.UrsacheSyntax, ""))
	}
	for _, w := range []struct{ name, zeichen string }{{"U+00A0", "\xC2\xA0"}, {"U+0085", "\xC2\x85"}, {"U+2003", "\xE2\x80\x83"}, {"U+000C", "\x0C"}} {
		out = append(out, schlecht(w.name+" nach der Wurzel", append(bytes.Clone(r), w.zeichen...), model.UrsacheSyntax, ""),
			schlecht(w.name+" vor der Wurzel", aender(r, ersteWurzel, w.zeichen+ersteWurzel), model.UrsacheSyntax, ""),
			schlecht(w.name+" nach Leerraum nach der Wurzel", append(bytes.Clone(r), " \n"+w.zeichen...), model.UrsacheSyntax, ""))
	}
	return out
}

// zeileV2B5: Kommentare, Verarbeitungsanweisungen, CDATA und Text sind ohne Wirkung; Nachlauf und doppelte Attribute sind syntax.
func zeileV2B5(t *testing.T, neu CodecNeu) {
	faelle := faelleV2B5(t)
	pruefeFaelle(t, neu, faelle)
	m, b, err := lies(neu, gut("R", festesV2(t, "v2_referenz.xml")))
	for _, f := range faelle {
		if f.ursache != "" {
			continue
		}
		m2, b2, err2 := lies(neu, f)
		if ergebnis(m2, b2, err2) != ergebnis(m, b, err) {
			t.Errorf("%s: Ergebnis weicht von R ab", f.name)
		}
	}
	for _, g := range faelleV2B5Attribute() {
		_, b := pruefeFall(t, neu, g.fall)
		if g.ursache == "" && fmt.Sprint(b.Funde) != fmt.Sprint(g.funde) {
			t.Errorf("%s: Funde %v, erwartet %v", g.name, b.Funde, g.funde)
		}
	}
}

// faelleV2B5Attribute sind doppelte Attribute (gleicher Namensraum und gleicher lokaler Name, auch über zwei Präfixe
// derselben URI) und ihre gültigen Nachbarn (gleicher lokaler Name in zwei Namensräumen, mit und ohne Namensraum).
func faelleV2B5Attribute() []g1Fall {
	prop := func(attrs string) []byte { return eins2(typ("T", "<Property "+attrs+"/>")) }
	erw := func(name string) model.Fund { return model.Fund{Art: model.FundErweiterung, Ort: "Ns.T/P", Name: name} }
	const kopf = `Name="P" Type="Edm.String"`
	var out []g1Fall
	for _, d := range []struct{ name, attrs string }{
		{"zwei Präfixe derselben URI", `xmlns:a="urn:x" xmlns:b="urn:x" ` + kopf + ` a:k="1" b:k="2"`},
		{"dasselbe Präfix zweimal", `xmlns:a="urn:x" ` + kopf + ` a:k="1" a:k="2"`},
		{"zwei Präfixe derselben URI, getrennt durch ein anderes Attribut", `xmlns:a="urn:x" xmlns:b="urn:x" a:k="1" ` + kopf + ` b:k="2"`},
		{"dasselbe nicht deklarierte Präfix zweimal", kopf + ` p:k="1" p:k="2"`},
		{"xmlns-Deklaration zweimal", `xmlns:a="urn:x" xmlns:a="urn:y" ` + kopf},
		{"xml:lang zweimal", kopf + ` xml:lang="de" xml:lang="en"`},
		{"Name zweimal", kopf + ` Name="Q"`},
	} {
		out = append(out, g1Fall{fall: schlecht("doppeltes Attribut mit Namensraum: "+d.name, prop(d.attrs), model.UrsacheSyntax, "")})
	}
	for _, d := range []struct {
		name, attrs string
		funde       []model.Fund
	}{
		{"gleicher lokaler Name in zwei Namensräumen", `xmlns:a="urn:x" xmlns:b="urn:y" ` + kopf + ` a:k="1" b:k="2"`, []model.Fund{erw("{urn:x}k"), erw("{urn:y}k")}},
		{"gleicher lokaler Name mit und ohne Namensraum", `xmlns:a="urn:x" ` + kopf + ` a:Name="1"`, []model.Fund{erw("{urn:x}Name")}},
		{"zwei nicht deklarierte Präfixe", kopf + ` p:k="1" q:k="2"`, []model.Fund{erw("{p}k"), erw("{q}k")}},
		{"zwei Präfixe derselben URI mit verschiedenen lokalen Namen", `xmlns:a="urn:x" xmlns:b="urn:x" ` + kopf + ` a:k="1" b:l="2"`, []model.Fund{erw("{urn:x}k"), erw("{urn:x}l")}},
	} {
		out = append(out, g1Fall{fall: gut("Attribute mit Namensraum: "+d.name, prop(d.attrs)), funde: d.funde})
	}
	return out
}

func faelleV2B6() []fall {
	return []fall{gut("Standard-Namensraum", eins2(typ("T", prop("P", "Edm.String")))), gut("Präfix edm", []byte(edmx2(
		`<edm:Schema xmlns:edm="`+v2EdmNs+`" Namespace="Ns"><edm:EntityType Name="T"><edm:Key><edm:PropertyRef Name="Id"/></edm:Key>`+
			`<edm:Property Name="Id" Type="Edm.Int32"/><edm:Property Name="P" Type="Edm.String"/></edm:EntityType></edm:Schema>`)))}
}

func zeileV2B6(t *testing.T, neu CodecNeu) {
	faelle := faelleV2B6()
	m0, b0, err0 := lies(neu, faelle[0])
	m1, b1, err1 := lies(neu, faelle[1])
	if err0 != nil || len(m0.EntityTypes()) != 1 || ergebnis(m0, b0, err0) != ergebnis(m1, b1, err1) {
		t.Errorf("Präfix und Standard-Namensraum ergeben verschiedene Ergebnisse: %v / %v", ergebnis(m0, b0, err0), ergebnis(m1, b1, err1))
	}
	for _, f := range []struct{ name, inhalt, ort, fund string }{
		{"Element ohne Präfix-Deklaration", `<EntityType Name="T">` + schluesselID + `<x:Property/></EntityType>`, "Ns.T", "{x}Property"},
		{"edm-Element ohne Deklaration des Präfixes", `<EntityType Name="T">` + schluesselID + `<edm:Property Name="P" Type="Edm.String"/></EntityType>`, "Ns.T", "{edm}Property"},
		{"Attribut ohne Deklaration", typ("T", `<Property Name="P" Type="Edm.String" sap:label="a"/>`), "Ns.T/P", "{sap}label"},
		{"Attribut mit Deklaration", typ("T", `<Property xmlns:sap="urn:s" Name="P" Type="Edm.String" sap:label="a"/>`), "Ns.T/P", "{urn:s}label"},
	} {
		_, b, err := lies(neu, gut(f.name, eins2(f.inhalt)))
		if err != nil || len(b.Funde) < 1 || b.Funde[len(b.Funde)-1] != (model.Fund{Art: model.FundErweiterung, Ort: f.ort, Name: f.fund}) {
			t.Errorf("%s: %v %v", f.name, b.Funde, err)
		}
	}
}

// dsV2 ist ein DataServices im Namensraum des EDMX von V2, gleich welchen Namensraum die Wurzel trägt.
const dsV2 = `<d:DataServices xmlns:d="` + v2EdmxNs + `"/>`

// wurzelV2 ist ein Dokument mit Wurzel name im Namensraum ns (leer: keiner); version "-" lässt das Attribut weg.
func wurzelV2(ns, name, version, kinder string) []byte {
	v := ""
	if version != "-" {
		v = ` Version="` + version + `"`
	}
	if ns == "" {
		return []byte(`<` + name + v + `>` + kinder + `</` + name + `>`)
	}
	return []byte(`<edmx:` + name + ` xmlns:edmx="` + ns + `"` + v + `>` + kinder + `</edmx:` + name + `>`)
}

// wurzelUrsache ist das Orakel von C1 und C2 (SPEC-171 Nr. 2 a, b): die Ursache des Dokuments aus Name, Namensraum
// und Version der Wurzel, leer für ein gültiges.
func wurzelUrsache(ns, name, version string) model.MetadatenUrsache {
	switch {
	case name == "Edmx" && ns == edmxNs:
		return model.UrsacheVersion
	case name != "Edmx" || ns != v2EdmxNs:
		return model.UrsacheWurzel
	case version != "1.0":
		return model.UrsacheVersion
	}
	return ""
}

func faelleV2C1() []fall {
	var out []fall
	for _, name := range []string{"Edmx", "Foo", "edmx"} {
		for _, ns := range []string{v2EdmxNs, edmxNs, "http://example.com/", ""} {
			for _, v := range []string{"1.0", "3.0", "-"} {
				doc := wurzelV2(ns, name, v, dsV2)
				label := fmt.Sprintf("%s in %q Version %s", name, ns, v)
				if u := wurzelUrsache(ns, name, v); u != "" {
					out = append(out, schlecht(label, doc, u, ""))
				} else {
					out = append(out, gut(label, doc))
				}
			}
		}
	}
	fremd := `<x:DataServices xmlns:x="` + v2EdmxNs + `x"/>`
	for _, f := range []struct {
		name string
		doc  []byte
	}{
		{"ohne DataServices", wurzelV2(v2EdmxNs, "Edmx", "1.0", "")}, {"zwei DataServices", wurzelV2(v2EdmxNs, "Edmx", "1.0", dsV2+dsV2)},
		{"DataServices im fremden Namensraum", wurzelV2(v2EdmxNs, "Edmx", "1.0", fremd)},
		{"DataServices im Namensraum von V4", wurzelV2(v2EdmxNs, "Edmx", "1.0", `<x:DataServices xmlns:x="`+edmxNs+`"/>`)},
	} {
		out = append(out, schlecht(f.name, f.doc, model.UrsacheStruktur, ""))
	}
	undekl := []byte(`<edmx:Edmx Version="1.0">` + dsV2 + `</edmx:Edmx>`)
	return append(out, schlecht("Präfix der Wurzel ohne Deklaration", undekl, model.UrsacheWurzel, ""),
		gut("Reference vor DataServices", wurzelV2(v2EdmxNs, "Edmx", "1.0", `<edmx:Reference Uri="x"><edmx:Include Namespace="Other"/></edmx:Reference>`+dsV2)),
		gut("ohne Schema", wurzelV2(v2EdmxNs, "Edmx", "1.0", dsV2)))
}

func zeileV2C1(t *testing.T, neu CodecNeu) {
	faelle := faelleV2C1()
	pruefeFaelle(t, neu, faelle)
	m, b := pruefeFall(t, neu, faelle[len(faelle)-1])
	if m.Version() != "2.0" || len(m.EntityTypes()) != 0 || len(b.Funde) != 0 {
		t.Errorf("ohne Schema: %q %v %v", m.Version(), m.EntityTypes(), b.Funde)
	}
	_, b = pruefeFall(t, neu, faelle[len(faelle)-2])
	if len(b.Funde) != 1 || b.Funde[0] != (model.Fund{Art: model.FundKonstrukt, Name: "Reference"}) {
		t.Errorf("Reference: %v", b.Funde)
	}
}

func faelleV2C2() []fall {
	var out []fall
	for _, v := range []string{"1.0", "-", "", "1", "1.00", "1.0 ", " 1.0", "2.0", "3.0", "4.0", "1.0.0", "01.0", "1,0"} {
		doc := wurzelV2(v2EdmxNs, "Edmx", v, dsV2)
		if v == "1.0" {
			out = append(out, gut("1.0", doc))
			continue
		}
		out = append(out, schlecht(fmt.Sprintf("Version %q", v), doc, model.UrsacheVersion, ""))
	}
	return out
}

func zeileV2C2(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2C2()) }

// dsAttr ist ein Dokument, dessen DataServices die Attribute attrs trägt (der Präfix m bindet den Namensraum der Metadaten).
func dsAttr(attrs string) []byte {
	return wurzelV2(v2EdmxNs, "Edmx", "1.0", `<edmx:DataServices xmlns:m="`+v2MetaNs+`" `+attrs+`/>`)
}

func faelleV2C3() []fall {
	var out []fall
	for _, v := range []string{"1.0", "2.0"} {
		out = append(out, gut("DataServiceVersion "+v, dsAttr(`m:DataServiceVersion="`+v+`"`)))
	}
	for _, v := range []string{"3.0", "2.0.1", "", "1.0 ", "2", "1", "4.0", " 2.0", "2.00"} {
		out = append(out, schlecht(fmt.Sprintf("DataServiceVersion %q", v), dsAttr(`m:DataServiceVersion="`+v+`"`), model.UrsacheVersion, ""))
	}
	out = append(out,
		gut("DataServiceVersion fehlend", dsAttr("")), gut("DataServiceVersion ohne Namensraum", dsAttr(`DataServiceVersion="3.0"`)),
		gut("MaxDataServiceVersion 3.0", dsAttr(`m:MaxDataServiceVersion="3.0"`)),
		gut("DataServiceVersion mit nicht deklariertem Präfix", dsAttr(`x:DataServiceVersion="3.0"`)),
		schlecht("DataServiceVersion mit anderem Präfix derselben URI", dsAttr(`xmlns:z="`+v2MetaNs+`" z:DataServiceVersion="3.0"`), model.UrsacheVersion, ""),
		schlecht("zweites DataServices", wurzelV2(v2EdmxNs, "Edmx", "1.0", dsV2+dsV2), model.UrsacheStruktur, ""),
		schlecht("zweites DataServices mit falscher Version", wurzelV2(v2EdmxNs, "Edmx", "1.0", dsV2+`<d:DataServices xmlns:d="`+v2EdmxNs+`" xmlns:m="`+v2MetaNs+`" m:DataServiceVersion="3.0"/>`), model.UrsacheStruktur, ""),
		schlecht("kein DataServices", wurzelV2(v2EdmxNs, "Edmx", "1.0", ""), model.UrsacheStruktur, ""),
	)
	return out
}

// zeileV2C3: DataServiceVersion fehlend gilt als 1.0, nur 1.0 und 2.0 sind gültig; genau ein DataServices (SPEC-171 Nr. 2 c).
func zeileV2C3(t *testing.T, neu CodecNeu) {
	faelle := faelleV2C3()
	pruefeFaelle(t, neu, faelle)
	for _, f := range faelle {
		if f.ursache != "" {
			continue
		}
		if _, b := pruefeFall(t, neu, f); strings.Contains(f.name, "Max") && len(b.Funde) != 0 {
			t.Errorf("%s: Funde %v, erwartet keine", f.name, b.Funde)
		}
	}
}

func faelleV2C4() []fall {
	typen := typ("T", prop("P", "Edm.String"))
	var out []fall
	for _, ns := range []string{"http://schemas.microsoft.com/ado/2006/04/edm", "http://schemas.microsoft.com/ado/2007/05/edm", v2EdmNs} {
		out = append(out, gut("Schema in "+ns, []byte(edmx2(`<Schema xmlns="`+ns+`" Namespace="Ns">`+typen+`</Schema>`))))
	}
	for _, ns := range []string{"http://schemas.microsoft.com/ado/2009/11/edm", edmNs} {
		out = append(out, schlecht("Schema in "+ns, []byte(edmx2(`<Schema xmlns="`+ns+`" Namespace="Ns">`+typen+`</Schema>`)), model.UrsacheVersion, ""))
	}
	return append(out, gut("Schema im fremden Namensraum", []byte(edmx2(`<Schema xmlns="urn:x" Namespace="Ns"/>`))),
		gut("Schema ohne Namensraum", []byte(edmx2(`<Schema xmlns="" Namespace="Ns"/>`))))
}

// zeileV2C4: Schema ist in den drei Namensräumen von CSDL 1.0 bis 2.0 gültig und ergibt dasselbe Modell, in dem von CSDL 3.0
// und von V4 ist es version, in einem anderen ein Fund (SPEC-171 Nr. 3).
func zeileV2C4(t *testing.T, neu CodecNeu) {
	faelle := faelleV2C4()
	pruefeFaelle(t, neu, faelle)
	soll := ""
	for _, f := range faelle[:3] {
		m, b, err := lies(neu, f)
		if len(m.EntityTypes()) != 1 || err != nil {
			t.Errorf("%s: %v %v", f.name, m.EntityTypes(), err)
		}
		if soll == "" {
			soll = ergebnis(m, b, err)
		} else if ergebnis(m, b, err) != soll {
			t.Errorf("%s: Modell weicht vom ersten ab", f.name)
		}
	}
	for i, name := range []string{"{urn:x}Schema", "{}Schema"} {
		m, b, err := lies(neu, faelle[5+i])
		if err != nil || len(m.Namespaces()) != 0 || !slices.Equal(b.Funde, []model.Fund{{Art: model.FundErweiterung, Name: name}}) {
			t.Errorf("%s: %v %v %v", faelle[5+i].name, m.Namespaces(), b.Funde, err)
		}
	}
}

func faelleV2C5() []fall {
	ds := dsV2
	return []fall{
		schlecht("Müll davor, falsche Version", append([]byte("x"), wurzelV2(v2EdmxNs, "Edmx", "3.0", ds)...), model.UrsacheSyntax, ""),
		schlecht("Müll davor, V4-Wurzel", append([]byte("x"), wurzelV2(edmxNs, "Edmx", "1.0", ds)...), model.UrsacheSyntax, ""),
		schlecht("falsche Version, ungültiges XML dahinter", wurzelV2(v2EdmxNs, "Edmx", "3.0", "<"), model.UrsacheVersion, ""),
		schlecht("V4-Wurzel, ungültiges XML dahinter", wurzelV2(edmxNs, "Edmx", "1.0", "<"), model.UrsacheVersion, ""),
		schlecht("falsche Wurzel, falsche Version", wurzelV2(v2EdmxNs, "Foo", "3.0", ds), model.UrsacheWurzel, ""),
		schlecht("V4-Wurzel, falsche Version", wurzelV2(edmxNs, "Edmx", "3.0", ds), model.UrsacheVersion, ""),
		schlecht("falsche Wurzel, ungültiges XML dahinter", wurzelV2(v2EdmxNs, "Foo", "1.0", "<"), model.UrsacheWurzel, ""),
		schlecht("falsche DataServiceVersion, Verstoß im Schema dahinter", wurzelV2(v2EdmxNs, "Edmx", "1.0",
			`<edmx:DataServices xmlns:m="`+v2MetaNs+`" m:DataServiceVersion="3.0">`+schema2("1x", "")+`</edmx:DataServices>`), model.UrsacheVersion, ""),
		schlecht("Verstoß im Schema vor der falschen DataServiceVersion des zweiten DataServices", wurzelV2(v2EdmxNs, "Edmx", "1.0",
			`<edmx:DataServices>`+schema2("1x", "")+`</edmx:DataServices>`+dsV2), model.UrsacheStruktur, "@Namespace"),
		schlecht("falsche Version, doppeltes Attribut an der Wurzel", []byte(`<edmx:Edmx xmlns:edmx="`+v2EdmxNs+`" Version="3.0" Version="1.0">`+ds+`</edmx:Edmx>`), model.UrsacheSyntax, ""),
		schlecht("Schema von CSDL 3.0 vor dem Verstoß im nächsten", wurzelV2(v2EdmxNs, "Edmx", "1.0",
			`<edmx:DataServices><Schema xmlns="http://schemas.microsoft.com/ado/2009/11/edm" Namespace="N1"/>`+schema2("1x", "")+`</edmx:DataServices>`), model.UrsacheVersion, ""),
	}
}

func zeileV2C5(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2C5()) }

func faelleV2E1(t *testing.T) []fall {
	t.Helper()
	r := festesV2(t, "v2_referenz.xml")
	l := int64(len(r) + 40)
	var out []fall
	for _, g := range []int64{l - 1, l, l + 1} {
		for _, b := range []bool{false, true} {
			vor, name := "", fmt.Sprintf("grenze %d", g)
			if b {
				vor, name = bom, name+" mit BOM"
			}
			// pad ist das Dokument aus vor, r mit Leerraum aufgefüllt und ende, insgesamt n Bytes.
			pad := func(n int64, ende string) []byte {
				return []byte(vor + string(auffuellen(r, int(n)-len(vor)-len(ende))) + ende)
			}
			out = append(out,
				fall{name: name + ": gültig bis zur Grenze", doc: pad(g, ""), max: g},
				fall{name: name + ": ein Byte darüber", doc: pad(g+1, ""), max: g, ursache: model.UrsacheZuGross},
				fall{name: name + ": Verstoß auf dem letzten Byte", doc: pad(g, "x"), max: g, ursache: model.UrsacheSyntax},
				fall{name: name + ": Verstoß auf dem Byte grenze+1", doc: pad(g+1, "x"), max: g, ursache: model.UrsacheZuGross},
				fall{name: name + ": Verstoß davor", doc: []byte(vor + "<a b>" + strings.Repeat(" ", int(g)+10)), max: g, ursache: model.UrsacheSyntax},
			)
		}
	}
	return append(out,
		fall{name: "grenze 1, Dokument von 2 Bytes", doc: []byte("<a"), max: 1, ursache: model.UrsacheZuGross},
		fall{name: "grenze MaxInt64", doc: r, max: math.MaxInt64},
		fall{name: "grenze MaxInt64, Verstoß", doc: append(bytes.Clone(r), 'x'), max: math.MaxInt64, ursache: model.UrsacheSyntax},
	)
}

// zeileV2E1: höchstens grenze Bytes, das BOM zählt mit; das Byte grenze+1 ist zu_gross, ein Verstoß davor entscheidet,
// einer dahinter wird nie gesehen (SPEC-125 Nr. 5): Schleife grenze-1, grenze, grenze+1 mal fünf Lagen mal BOM.
func zeileV2E1(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2E1(t)) }

func faelleV2E2() []fall {
	var out []fall
	for _, w := range []struct {
		name, auf, zu string
		anz           int
		ein           func(string) string
	}{
		{"fremdes Element", `<x:W xmlns:x="urn:x">`, `</x:W>`, 61, func(i string) string { return i }},
		{"gemeldetes Konstrukt", `<Function>`, `</Function>`, 61, func(i string) string { return i }},
		{"Inhalt einer Property", `<x:W xmlns:x="urn:x">`, `</x:W>`, 59, func(i string) string {
			return typ("T", `<Property Name="P" Type="Edm.String">`+i+`</Property>`)
		}},
		{"Inhalt eines End", `<x:W xmlns:x="urn:x">`, `</x:W>`, 59, func(i string) string {
			return typ("T", "") + `<Association Name="R"><End Type="Ns.T" Role="F" Multiplicity="*">` + i + `</End><End Type="Ns.T" Role="G" Multiplicity="*"/></Association>`
		}},
		{"gemeldetes Konstrukt in einem Konstrukt", `<ReferentialConstraint>`, `</ReferentialConstraint>`, 60, func(i string) string {
			return typ("T", "") + `<Association Name="R"><End Type="Ns.T" Role="F" Multiplicity="*"/><End Type="Ns.T" Role="G" Multiplicity="*"/>` + i + `</Association>`
		}},
	} {
		wrap := func(n int) []byte { return eins2(w.ein(verschachtelt(n, w.auf, w.zu))) }
		out = append(out, gut(w.name+" bei Tiefe 64", wrap(w.anz)), schlecht(w.name+" bei Tiefe 65", wrap(w.anz+1), model.UrsacheTiefe, ""))
	}
	return append(out, gut("70 flache Geschwister", eins2(strings.Repeat(`<Function/>`, 70))))
}

func zeileV2E2(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2E2()) }

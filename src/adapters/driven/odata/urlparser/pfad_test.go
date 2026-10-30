package urlparser_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	segm = model.URLSegment
	kodi = model.URLKodierung
	pnu  = model.URLPfadNichtUnterstuetzt
	pfrm = model.URLPfadForm
)

// TestPfadFormen (Zeile 20): die Entscheidung je Form des Pfads, 23 Formen.
func TestPfadFormen(t *testing.T) {
	menge := &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: "Set"}}
	faelle := []fall{
		{pfad: "/", u: model.URLServiceDokument, ort: "pfad"},
		{pfad: "/$metadata", want: &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMetadaten}}},
		{pfad: "/$metadata/x", u: pfrm, ort: "pfad/2"},
		{pfad: "/$metadata/", u: segm, ort: "pfad/2"},
		{pfad: "/Set", want: menge},
		{pfad: "/Set(1)", want: entity("Set", teil("", zahlLit("1")))},
		{pfad: "/Set/$count", want: &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceAnzahl, EntitySet: "Set"}}},
		{pfad: "/Set(1)/$count", u: pfrm, ort: "pfad/2"},
		{pfad: "/Set/$count/x", u: pfrm, ort: "pfad/3"},
		{pfad: "/Set/Rest", u: pnu, ort: "pfad/2"},
		{pfad: "/Set(1)/Rest", u: pnu, ort: "pfad/2"},
		{pfad: "/Set(1)/Ns.Typ", u: pnu, ort: "pfad/2"},
		{pfad: "/Set(1)/$value", u: pnu, ort: "pfad/2"},
		{pfad: "/Set(1)/$ref", u: pnu, ort: "pfad/2"},
		{pfad: "/Set()", u: pnu, ort: "pfad/1"},
		{pfad: "/$batch", u: pnu, ort: "pfad/1"},
		{pfad: "/$entity", u: pnu, ort: "pfad/1"},
		{pfad: "/$all", u: pnu, ort: "pfad/1"},
		{pfad: "/$crossjoin", u: pnu, ort: "pfad/1"},
		{pfad: "/$value", u: segm, ort: "pfad/1"},
		{pfad: "/$Metadata", u: segm, ort: "pfad/1"},
		{pfad: "/$count", u: segm, ort: "pfad/1"},
		{pfad: "/Set//x", u: segm, ort: "pfad/2"},
	}
	if len(faelle) != 23 {
		t.Fatalf("%d Formen, erwartet 23", len(faelle))
	}
	pruefeFaelle(t, faelle)
}

// TestSegmentForm (Zeile 27): ein Segment außerhalb der Grammatik ist `segment`, 14 Formen.
func TestSegmentForm(t *testing.T) {
	formen := []struct{ pfad, ort string }{
		{"/Set//x", "pfad/2"}, {"/Set/", "pfad/2"}, {"/.", "pfad/1"}, {"/..", "pfad/1"}, {"/%2E%2E", "pfad/1"}, {"/Set;x", "pfad/1"},
		{"/%20", "pfad/1"}, {"/Set.x", "pfad/1"}, {"/Set)", "pfad/1"}, {"/Set-1", "pfad/1"}, {"/1Set", "pfad/1"},
		{"/-Set", "pfad/1"}, {"/.Set", "pfad/1"}, {"/Set%201", "pfad/1"},
	}
	var faelle []fall
	for _, f := range formen {
		faelle = append(faelle, fall{pfad: f.pfad, u: segm, ort: f.ort})
	}
	if len(faelle) != 14 {
		t.Fatalf("%d Formen, erwartet 14", len(faelle))
	}
	pruefeFaelle(t, faelle)
}

// TestNameWieGeschrieben (Zeile 28): die Schreibweise des Namens zählt, Unicode wird nicht normalisiert.
func TestNameWieGeschrieben(t *testing.T) {
	zusammen, zerlegt := "\xc3\xa9t", "e\xcc\x81t"
	for _, n := range []string{"products", "Products", "PRODUCTS", zusammen, "Gr\xc3\xb6\xc3\x9fe"} {
		if got := gelesen(t, "/"+n, "").Ressource.EntitySet; got != n {
			t.Errorf("Name %q gelesen als %q", n, got)
		}
	}
	pruefeFaelle(t, []fall{{pfad: "/" + zerlegt, u: segm, ort: "pfad/1"}})
	a, b := gelesen(t, "/S('"+zusammen+"')", ""), gelesen(t, "/S('"+zerlegt+"')", "")
	if a.Ressource.Schluessel[0].Literal.Text == b.Ressource.Schluessel[0].Literal.Text {
		t.Error("zusammengesetzt und zerlegt ergeben denselben Text")
	}
}

// stellen sind die sechs Stellen, an denen Text der URL dekodiert wird; %s steht für den geprüften Text.
func stellen() []struct {
	name, pfad, query string
	ort               string
} {
	return []struct {
		name, pfad, query string
		ort               string
	}{
		{"Name des Sets", "/S%s", "", "pfad/1"},
		{"Name eines Schlüsselteils", "/S(Id%s=1)", "", "pfad/1"},
		{"Schlüsselliteral", "/S('a%s')", "", "pfad/1"},
		{"Rest-Segment", "/S/R%s", "", "pfad/2"},
		{"Name einer Option", "/S", "x%s=1", "query/1"},
		{"Wert einer Option", "/S", "x=a%s", "query/1"},
	}
}

func eingesetzt(vorlage, s string) string { return strings.Replace(vorlage, "%s", s, 1) }

// TestEinmalDekodiert (Zeile 22): `%255F` wird einmal dekodiert und ergibt den Text `%5F`; fünf Stellen.
func TestEinmalDekodiert(t *testing.T) {
	// Name des Sets, Name eines Schlüsselteils, Name einer Option: der Text `%5F` ist kein Name und wird als solcher abgewiesen.
	pruefeFaelle(t, []fall{
		{pfad: "/S%255F", u: segm, ort: "pfad/1"},
		{pfad: "/S(Id%255F=1)", u: model.URLKey, ort: "pfad/1"},
		{pfad: "/S", query: "x%255F=1", u: model.URLOptionName, ort: "query/1"},
	})
	// Inhalt eines String-Literals und Wert einer Option: der Text ist `%27`, die Folge bleibt erhalten.
	if got := gelesen(t, "/S('%2527')", "").Ressource.Schluessel[0].Literal.Text; got != "'%27'" {
		t.Errorf("Literal: %q", got)
	}
	if got := gelesen(t, "/S", "x=%2527").Optionen.Sonstige[0].Wert; got != "%27" {
		t.Errorf("Wert: %q", got)
	}
}

// TestSchraegstrich (Zeile 23): `%2F` trennt nicht, ein rohes `/` trennt auch im Literal.
func TestSchraegstrich(t *testing.T) {
	if got := gelesen(t, "/S('a%2Fb')", "").Ressource.Schluessel[0].Literal.Text; got != "'a/b'" {
		t.Errorf("Literal: %q", got)
	}
	pruefeFaelle(t, []fall{
		{pfad: "/Set%2Fx", u: segm, ort: "pfad/1"},
		{pfad: "/Set('a/b')", u: model.URLKey, ort: "pfad/1"},
	})
}

// TestKodierungUngueltig (Zeile 24): fünf Formen an sechs Stellen sind `kodierung`, `%41` am Ende bleibt gültig; 36 Fälle.
func TestKodierungUngueltig(t *testing.T) {
	n := 0
	for _, st := range stellen() {
		for _, form := range []string{"%", "%1", "%G1", "%1G", "a%"} {
			pruefeFall(t, parserNeu(), fall{pfad: eingesetzt(st.pfad, form), query: eingesetzt(st.query, form), u: kodi, ort: st.ort})
			n++
		}
		f := fall{pfad: eingesetzt(st.pfad, "%41"), query: eingesetzt(st.query, "%41")}
		if st.name == "Rest-Segment" {
			f.u, f.ort = pnu, st.ort
		}
		pruefeFall(t, parserNeu(), f)
		n++
	}
	if n != 36 {
		t.Fatalf("%d Fälle, erwartet 36", n)
	}
}

// TestUTF8 (Zeile 25): ungültiges UTF-8 ist `kodierung`; das Escape und das rohe Zeichen ergeben denselben AST; 48 Fälle.
func TestUTF8(t *testing.T) {
	ungueltig := []string{"%FF", "%C3", "%C0%80", "%ED%A0%80", "%C3%28", "\xff", "\xc3"}
	n := 0
	for _, st := range stellen() {
		for _, form := range ungueltig {
			pruefeFall(t, parserNeu(), fall{pfad: eingesetzt(st.pfad, form), query: eingesetzt(st.query, form), u: kodi, ort: st.ort})
			n++
		}
		a, ea := parse(eingesetzt(st.pfad, "%C3%A4"), eingesetzt(st.query, "%C3%A4"))
		b, eb := parse(eingesetzt(st.pfad, "\xc3\xa4"), eingesetzt(st.query, "\xc3\xa4"))
		if (ea == nil) != (eb == nil) || (ea != nil && ea.Error() != eb.Error()) || !gleich(a, b) {
			t.Errorf("%s: Escape %v %v, roh %v %v", st.name, a, ea, b, eb)
		}
		n++
	}
	if n != 48 {
		t.Fatalf("%d Fälle, erwartet 48", n)
	}
}

// TestNul (Zeile 26): U+0000 ist `kodierung` an sechs Stellen; ein anderes Steuerzeichen im Literal bleibt Text.
func TestNul(t *testing.T) {
	n := 0
	for _, st := range stellen() {
		pruefeFall(t, parserNeu(), fall{pfad: eingesetzt(st.pfad, "%00"), query: eingesetzt(st.query, "%00"), u: kodi, ort: st.ort})
		n++
	}
	for _, c := range []string{"%0A", "%09"} {
		got := gelesen(t, "/S('x"+c+"y')", "").Ressource.Schluessel[0].Literal.Text
		if want := map[string]string{"%0A": "'x\ny'", "%09": "'x\ty'"}[c]; got != want {
			t.Errorf("%s: %q", c, got)
		}
		n++
	}
	if n != 8 {
		t.Fatalf("%d Fälle, erwartet 8", n)
	}
}

// TestRestSegmente (Zeile 41): zehn Formen hinter einer Menge und hinter einem Schlüssel; 20 Fälle.
func TestRestSegmente(t *testing.T) {
	formen := []struct {
		text string
		u    model.URLUrsache
	}{
		{"Name", pnu}, {"Ns.Typ", pnu}, {"Items(2)", pnu}, {"$value", pnu}, {"$ref", pnu}, {"$count", pnu},
		{"Items(", segm}, {"Items)", segm}, {"", segm}, {"a%20b", segm},
	}
	n := 0
	for _, vor := range []string{"/S", "/S(1)"} {
		for _, f := range formen {
			soll := fall{pfad: vor + "/" + f.text, u: f.u, ort: "pfad/2"}
			switch {
			case f.text == "$count" && vor == "/S":
				soll = fall{pfad: soll.pfad}
			case f.text == "$count":
				soll.u = pfrm
			}
			pruefeFall(t, parserNeu(), soll)
			n++
			// An dritter Stelle gilt dieselbe Form; `$count` ist dort ein Rest-Segment wie jedes andere.
			dritte := fall{pfad: vor + "/Mid/" + f.text, u: f.u, ort: "pfad/2"}
			if f.u == segm {
				dritte.ort = "pfad/3"
			}
			pruefeFall(t, parserNeu(), dritte)
		}
	}
	if n != 20 {
		t.Fatalf("%d Fälle, erwartet 20", n)
	}
}

// TestVerstossHinter501 (Zeile 42): Kodierung und Form eines späteren Segments werden auch hinter einem nicht unterstützten
// geprüft; 6 Fälle.
func TestVerstossHinter501(t *testing.T) {
	n := 0
	for _, vorn := range []struct{ pfad, ort string }{{"/$batch", "pfad/2"}, {"/Set(1)/Name", "pfad/3"}, {"/Set()", "pfad/2"}} {
		for _, hinten := range []struct {
			text string
			u    model.URLUrsache
		}{{"/%G1", kodi}, {"/a b", segm}} {
			pruefeFall(t, parserNeu(), fall{pfad: vorn.pfad + hinten.text, u: hinten.u, ort: vorn.ort})
			n++
		}
	}
	if n != 6 {
		t.Fatalf("%d Fälle, erwartet 6", n)
	}
}

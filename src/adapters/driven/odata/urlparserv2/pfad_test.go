package urlparserv2_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

type pfadFall struct {
	pfad, q string
	want    erw
}

func laufe(t *testing.T, p *urlparserv2.Parser, faelle []pfadFall) {
	t.Helper()
	for _, f := range faelle {
		pruefe(t, p, f.pfad, f.q, f.want)
	}
}

// TestZerlegung (Zeile 23): zerlegt wird vor der Dekodierung an rohem `/`; `%2F` trennt nichts. Menge 18.
func TestZerlegung(t *testing.T) {
	faelle := []pfadFall{
		{"//", "", ung(model.URLSegment, "pfad/1")}, {"/S//", "", ung(model.URLSegment, "pfad/2")},
		{"/S/", "", ung(model.URLSegment, "pfad/2")}, {"/S(1)/", "", ung(model.URLSegment, "pfad/2")},
		{"/.", "", ung(model.URLSegment, "pfad/1")}, {"/..", "", ung(model.URLSegment, "pfad/1")},
		{"/%2E", "", ung(model.URLSegment, "pfad/1")}, {"/%2e%2e", "", ung(model.URLSegment, "pfad/1")},
		{"/S/./x", "", ung(model.URLSegment, "pfad/2")}, {"/S;", "", ung(model.URLSegment, "pfad/1")},
		{"/S x", "", ung(model.URLSegment, "pfad/1")},
		{"/a%2Fb", "", ung(model.URLSegment, "pfad/1")}, {"/S(1)/a%2Fb", "", ung(model.URLSegment, "pfad/2")},
		{"/S('a%2Fb')", "", okR(entityR("S", teil("", model.LiteralZeichenkette, "'a/b'")))},
		{"/", "", kmp(model.KompatPfad, "pfad")},
		{"/S(1)", "", okR(entityR("S", zahlT("1")))}, {"/S/$count", "", okR(anzahlR("S"))}, {"/%73", "", okR(mengeR("s"))},
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 18 {
		t.Errorf("%d Fälle, erwartet 18", len(faelle))
	}
}

// TestPfadDekodierung (Zeile 24): jedes Segment wird genau einmal dekodiert; ungültige Folgen sind `kodierung` und gehen der
// Grammatik vor. Menge 28.
func TestPfadDekodierung(t *testing.T) {
	ae := zeichen(0xE4)
	faelle := []pfadFall{
		{"/S('%2527')", "", okR(entityR("S", teil("", model.LiteralZeichenkette, "'%27'")))},
		{"/%C3%A4", "", okR(mengeR(ae))}, {"/%c3%a4", "", okR(mengeR(ae))}, {"/" + ae, "", okR(mengeR(ae))},
		{"/S('a+b')", "", okR(entityR("S", teil("", model.LiteralZeichenkette, "'a+b'")))},
		{"/S('a%5Cb')", "", okR(entityR("S", teil("", model.LiteralZeichenkette, "'a\\b'")))},
		{"/S('a%01b')", "", okR(entityR("S", teil("", model.LiteralZeichenkette, "'a"+string([]byte{1})+"b'")))},
		{"/S(%FF", "", ung(model.URLKodierung, "pfad/1")},
	}
	ungueltig := []string{"%", "%1", "%G1", "%1G", "S%", "%FF", "%C3", "%C0%80", "%ED%A0%80", "%00"}
	for _, u := range ungueltig {
		faelle = append(faelle, pfadFall{"/" + u, "", ung(model.URLKodierung, "pfad/1")},
			pfadFall{"/S/" + u, "", ung(model.URLKodierung, "pfad/2")})
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 8+20 {
		t.Errorf("%d Fälle, erwartet 28", len(faelle))
	}
}

// TestErstesSegment (Zeile 25): `$metadata`; vier Wörter sind Kompatibilität `pfad`, jedes andere Segment mit `$` ist `segment`.
// Menge 16.
func TestErstesSegment(t *testing.T) {
	faelle := []pfadFall{
		{"/$metadata", "", okR(model.Ressource{Art: model.RessourceMetadaten})},
		{"/$metadata/x", "", ung(model.URLPfadForm, "pfad/2")},
		{"/%24metadata", "", okR(model.Ressource{Art: model.RessourceMetadaten})},
		{"/S", "", okR(mengeR("S"))}, {"/S(1)", "", okR(entityR("S", zahlT("1")))}, {"/S/$count", "", okR(anzahlR("S"))},
	}
	for _, w := range []string{"$batch", "$entity", "$all", "$crossjoin"} {
		faelle = append(faelle, pfadFall{"/" + w, "", kmp(model.KompatPfad, "pfad/1")})
	}
	for _, w := range []string{"$value", "$links", "$ref", "$count", "$Metadata", "$"} {
		faelle = append(faelle, pfadFall{"/" + w, "", ung(model.URLSegment, "pfad/1")})
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 16 {
		t.Errorf("%d Fälle, erwartet 16", len(faelle))
	}
}

// TestNamen (Zeile 26): ein Name besteht aus Zeichen der Kategorien `L` und `Nd` (am Anfang `L` oder `_`), 1 bis 128 Zeichen; der
// Rest hinter dem Namen ist leer oder eine Klammer. Menge 70. Das Orakel für die Einordnung eines Zeichens ist `unicode.Is`
// (akzeptiertes Negativ, Zeile 26 des Plans).
func TestNamen(t *testing.T) {
	zeichenL := []string{"a", "Z", zeichen(0xE4), zeichen(0x4E2D), zeichen(0x416), zeichen(0x3A9)}
	zeichenNd := []string{"7", zeichen(0x663), zeichen(0xFF11), zeichen(0x967)}
	zeichenNl := []string{zeichen(0x2160), zeichen(0x3021), zeichen(0x16EE)}
	zeichenNo := []string{zeichen(0xB2), zeichen(0xB3), zeichen(0xBD), zeichen(0x2460)}
	zeichenM := []string{zeichen(0x301), zeichen(0x300), zeichen(0x903)}
	sonst := []string{"-", ".", " ", "@", "*", "!", "~", ",", "'"}
	alle := map[*unicode.RangeTable][]string{unicode.L: zeichenL, unicode.Nd: zeichenNd, unicode.Nl: zeichenNl, unicode.No: zeichenNo,
		unicode.M: zeichenM}
	var faelle []pfadFall
	n := 0
	for _, gruppe := range [][]string{zeichenL, zeichenNd, zeichenNl, zeichenNo, zeichenM, sonst} {
		for _, c := range gruppe {
			r := []rune(c)[0]
			anfangOk := r == '_' || unicode.Is(unicode.L, r)
			innenOk := anfangOk || unicode.Is(unicode.Nd, r)
			erwA, erwI := ung(model.URLSegment, "pfad/1"), ung(model.URLSegment, "pfad/1")
			if anfangOk {
				erwA = okR(mengeR(c))
			}
			if innenOk {
				erwI = okR(mengeR("a" + c))
			}
			faelle = append(faelle, pfadFall{"/" + c, "", erwA}, pfadFall{"/a" + c, "", erwI})
			n += 2
		}
	}
	if n != 58 || len(alle) != 5 {
		t.Fatalf("%d Zeichenfälle, erwartet 58", n)
	}
	a := func(k int) string { return strings.Repeat("a", k) }
	faelle = append(faelle,
		pfadFall{"/a", "", okR(mengeR("a"))}, pfadFall{"/" + a(128), "", okR(mengeR(a(128)))}, pfadFall{"/" + a(129), "", ung(model.URLSegment, "pfad/1")},
		pfadFall{"/e" + zeichen(0x301), "", ung(model.URLSegment, "pfad/1")}, pfadFall{"/" + zeichen(0xE9), "", okR(mengeR(zeichen(0xE9)))},
		pfadFall{"/S(", "", ung(model.URLKey, "pfad/1")}, pfadFall{"/S(1", "", ung(model.URLKey, "pfad/1")},
		pfadFall{"/S)", "", ung(model.URLSegment, "pfad/1")}, pfadFall{"/S-1", "", ung(model.URLSegment, "pfad/1")},
		pfadFall{"/S.x", "", ung(model.URLSegment, "pfad/1")}, pfadFall{"/S 1", "", ung(model.URLSegment, "pfad/1")},
		pfadFall{"/S()", "", kmp(model.KompatPfad, "pfad/1")})
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 70 {
		t.Errorf("%d Fälle, erwartet 70", len(faelle))
	}
}

// TestKoerper (Zeile 27): Teile, benannt oder unbenannt, nur eine der beiden Formen, Leerraum, nacktes Wort. Menge 39.
func TestKoerper(t *testing.T) {
	var t16, t17 []string
	var w16 []model.SchluesselTeil
	for i := range 17 {
		t17 = append(t17, fmt.Sprintf("K%d=1", i))
		if i < 16 {
			t16 = append(t16, t17[i])
			w16 = append(w16, teil(fmt.Sprintf("K%d", i), model.LiteralZahl, "1"))
		}
	}
	key, lit := ung(model.URLKey, "pfad/1"), ung(model.URLLiteral, "pfad/1")
	zk := func(s string) model.Literal { return model.Literal{Art: model.LiteralZeichenkette, Text: s} }
	faelle := []pfadFall{
		{"/S(1)", "", okR(entityR("S", zahlT("1")))}, {"/S(A=1,B=2)", "", okR(entityR("S", teil("A", model.LiteralZahl, "1"), teil("B", model.LiteralZahl, "2")))},
		{"/S(" + strings.Join(t16, ",") + ")", "", okR(entityR("S", w16...))}, {"/S(" + strings.Join(t17, ",") + ")", "", ung(model.URLZuViele, "pfad/1")},
		{"/S(Id=1)", "", okR(entityR("S", teil("Id", model.LiteralZahl, "1")))},
		{"/S(Id=1,K='x')", "", okR(entityR("S", teil("Id", model.LiteralZahl, "1"), model.SchluesselTeil{Name: "K", Literal: zk("'x'")}))},
		{"/S(K='x',Id=1)", "", okR(entityR("S", model.SchluesselTeil{Name: "K", Literal: zk("'x'")}, teil("Id", model.LiteralZahl, "1")))},
		{"/S(1,2)", "", key}, {"/S(true,false)", "", key}, {"/S(1,Id=2)", "", key}, {"/S(Id=1,2)", "", key},
		{"/S(1,,2)", "", key}, {"/S(,1)", "", key}, {"/S(1,)", "", key}, {"/S(=1)", "", key}, {"/S(Id)", "", key}, {"/S(Id=)", "", key},
		{"/S(Id=1,Id=2)", "", key}, {"/S(1=2)", "", key},
		{"/S(I d=1)", "", key}, {"/S(Id%09=1)", "", key},
		{"/S(yes)", "", key}, {"/S(t)", "", key}, {"/S(true)", "", okR(entityR("S", teil("", model.LiteralWahrheitswert, "true")))},
		{"/S(false)", "", okR(entityR("S", teil("", model.LiteralWahrheitswert, "false")))},
		{"/S(TRUE)", "", okR(entityR("S", teil("", model.LiteralWahrheitswert, "true")))},
		{"/S(null)", "", lit}, {"/S(INF)", "", lit}, {"/S(NaN)", "", lit}, {"/S(inf)", "", key},
		{"/S(A=1, B=2)", "", key}, {"/S(A =1)", "", key}, {"/S(A= 1)", "", lit}, {"/S(A=1 ,B=2)", "", lit}, {"/S(A=1%09,B=2)", "", lit},
		{"/S(A=%201)", "", lit},
		{"/S('a,b')", "", okR(entityR("S", model.SchluesselTeil{Literal: zk("'a,b'")}))},
		{"/S(A='x,y',B='z')", "", okR(entityR("S", model.SchluesselTeil{Name: "A", Literal: zk("'x,y'")}, model.SchluesselTeil{Name: "B", Literal: zk("'z'")}))},
		{"/S(A='1,2,3')", "", okR(entityR("S", model.SchluesselTeil{Name: "A", Literal: zk("'1,2,3'")}))},
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 39 {
		t.Errorf("%d Fälle, erwartet 39", len(faelle))
	}
}

// TestRestSegmente (Zeile 34): `$count` und weitere Segmente; Rest-Segmente sind Kompatibilität `pfad`, ein Rest-Segment, das die
// Form verfehlt, ist `segment`; Kodierung und Form werden auch hinter einem Kompatibilitätsfehler geprüft. Menge 25.
func TestRestSegmente(t *testing.T) {
	var faelle []pfadFall
	for _, e := range ausserhalbPfade {
		faelle = append(faelle, pfadFall{e.pfad, "", kmp(model.KompatPfad, e.ort)})
	}
	faelle = append(faelle,
		pfadFall{"/S/$count", "", okR(anzahlR("S"))}, pfadFall{"/S/$count/x", "", ung(model.URLPfadForm, "pfad/3")},
		pfadFall{"/S(1)/$count", "", ung(model.URLPfadForm, "pfad/2")}, pfadFall{"/$metadata/x", "", ung(model.URLPfadForm, "pfad/2")},
		pfadFall{"/$metadata/", "", ung(model.URLSegment, "pfad/2")}, pfadFall{"/S/$count/", "", ung(model.URLSegment, "pfad/3")},
		pfadFall{"/S(1)/", "", ung(model.URLSegment, "pfad/2")}, pfadFall{"/S(1)/N!", "", ung(model.URLSegment, "pfad/2")},
		pfadFall{"/S(1)/N(", "", ung(model.URLSegment, "pfad/2")}, pfadFall{"/S(1)/$links/%FF", "", ung(model.URLKodierung, "pfad/3")})
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 25 {
		t.Errorf("%d Fälle, erwartet 25", len(faelle))
	}
}

// TestFunktionsimport (Zeile 35): ein Funktionsimport ist dem Parser ein Name wie ein Set (SPEC-174 Nr. 7). Menge 3.
func TestFunktionsimport(t *testing.T) {
	laufe(t, urlparserv2.New(), []pfadFall{
		{"/GetProducts", "", okR(mengeR("GetProducts"))},
		{"/GetProducts", "x=1", ok(mengeR("GetProducts"), model.Optionen{Sonstige: []model.SonstigeOption{{Name: "x", Wert: "1"}}})},
		{"/GetProducts(1)", "", okR(entityR("GetProducts", zahlT("1")))},
	})
}

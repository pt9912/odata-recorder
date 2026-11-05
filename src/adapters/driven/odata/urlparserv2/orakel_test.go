package urlparserv2_test

import (
	"fmt"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Das Orakel der Zeile 48 steht in dieser Datei und importiert nichts aus dem Produktcode außer den Typen des Modells: es kennt die
// Verdikte der Pfadformen und Optionsformen aus der Spezifikation (SPEC-160 Nr. 5, SPEC-173) und fügt sie nach der Rangfolge
// zusammen, ohne den Leser zu rufen.

// pForm ist eine Pfadform mit dem Ergebnis, das sie ohne Verstoß ergibt, und ihren Verstößen in der Lesefolge; stufe0 ist ein
// Verstoß der Gesamtform, der jedem anderen vorgeht.
type pForm struct {
	text   string
	ress   model.Ressource
	vds    []vd
	stufe0 *vd
}

// qForm ist eine Optionsform; behandelt nennt den Namen, wenn die Option zu den sieben behandelten gehört (ein zweites Vorkommen
// ist `option_doppelt`), v den einen Verstoß (ohne Ort, der Ort ist die Stelle der Option), anwenden ihren Beitrag zum Ergebnis.
type qForm struct {
	text      string
	behandelt string
	v         *vd
	anwenden  func(*model.Optionen)
}

func qOk(text, behandelt string, f func(*model.Optionen)) qForm {
	return qForm{text: text, behandelt: behandelt, anwenden: f}
}
func qU(text, behandelt string, u model.URLUrsache) qForm {
	v := vU(u, "")
	return qForm{text: text, behandelt: behandelt, v: &v}
}
func qK(text, behandelt string, k model.KompatUrsache) qForm {
	v := vK(k, "")
	return qForm{text: text, behandelt: behandelt, v: &v}
}

func zeit(text string) model.SchluesselTeil { return teil("", model.LiteralZeitpunkt, text) }

func pfadFormen() (gut, schlecht []pForm) {
	zk := func(s string) model.SchluesselTeil { return teil("", model.LiteralZeichenkette, s) }
	g := func(text string, r model.Ressource) { gut = append(gut, pForm{text: text, ress: r}) }
	g("/$metadata", model.Ressource{Art: model.RessourceMetadaten})
	g("/S", mengeR("S"))
	g("/GetProducts", mengeR("GetProducts"))
	g("/S(1)", entityR("S", zahlT("1")))
	g("/S('a')", entityR("S", zk("'a'")))
	g("/S('O''Hara')", entityR("S", zk("'O''Hara'")))
	g("/S(Id=1,K='x')", entityR("S", teil("Id", model.LiteralZahl, "1"), teil("K", model.LiteralZeichenkette, "'x'")))
	g("/S(true)", entityR("S", teil("", model.LiteralWahrheitswert, "true")))
	g("/S(42L)", entityR("S", zahlT("42")))
	g("/S(1.50M)", entityR("S", zahlT("1.5")))
	g("/S(guid'AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE')", entityR("S", teil("", model.LiteralGUID, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")))
	g("/S(datetime'2020-01-02T03:04')", entityR("S", zeit("2020-01-02T03:04:00Z")))
	g("/S/$count", anzahlR("S"))
	g("/S(A=1,B=2,C=3)", entityR("S", teil("A", model.LiteralZahl, "1"), teil("B", model.LiteralZahl, "2"), teil("C", model.LiteralZahl, "3")))
	b := func(text string, v ...vd) { schlecht = append(schlecht, pForm{text: text, vds: v}) }
	b("//", vU(model.URLSegment, "pfad/1"))
	b("/S/", vU(model.URLSegment, "pfad/2"))
	b("/.", vU(model.URLSegment, "pfad/1"))
	b("/S%", vU(model.URLKodierung, "pfad/1"))
	b("/S(1", vU(model.URLKey, "pfad/1"))
	b("/S(1,,2)", vU(model.URLKey, "pfad/1"))
	b("/S(null)", vU(model.URLLiteral, "pfad/1"))
	b("/S(Id=1,Id=2)", vU(model.URLKey, "pfad/1"))
	b("/S(1)/$count", vU(model.URLPfadForm, "pfad/2"))
	b("/$metadata/x", vU(model.URLPfadForm, "pfad/2"))
	b("/$value", vU(model.URLSegment, "pfad/1"))
	b("/S(K=guid'zz')", vU(model.URLLiteral, "pfad/1"))
	b("/S('"+strings.Repeat("a", 600)+"')", vU(model.URLZuLang, "pfad/1"))
	schlecht = append(schlecht, pForm{text: "/a" + strings.Repeat("/a", 17), stufe0: &vd{klasseU: true, u: model.URLZuViele, ort: "pfad"}})
	b("/", vK(model.KompatPfad, "pfad"))
	b("/$batch", vK(model.KompatPfad, "pfad/1"))
	b("/S()", vK(model.KompatPfad, "pfad/1"))
	b("/S(1)/N", vK(model.KompatPfad, "pfad/2"))
	b("/S(1)/$links/N", vK(model.KompatPfad, "pfad/2"))
	b("/S(1d)", vK(model.KompatLiteralTyp, "pfad/1"))
	b("/S(1e3)", vK(model.KompatLiteralTyp, "pfad/1"))
	b("/S(binary'AA')", vK(model.KompatLiteralTyp, "pfad/1"))
	b("/S(1)/$value", vK(model.KompatPfad, "pfad/2"))
	b("/S(1)/N/$value", vK(model.KompatPfad, "pfad/2"))
	b("/S/N", vK(model.KompatPfad, "pfad/2"))
	b("/S(A=1f,B=2)", vK(model.KompatLiteralTyp, "pfad/1"))
	return gut, schlecht
}

func optionsFormen() (gut, schlecht []qForm) {
	sonst := func(text, name, wert string) qForm {
		return qOk(text, "", func(o *model.Optionen) { o.Sonstige = append(o.Sonstige, model.SonstigeOption{Name: name, Wert: wert}) })
	}
	gut = []qForm{
		qOk("$top=5", "$top", func(o *model.Optionen) { o.Top = model.ZahlOption{Gesetzt: true, Wert: 5} }),
		qOk("$skip=2", "$skip", func(o *model.Optionen) { o.Skip = model.ZahlOption{Gesetzt: true, Wert: 2} }),
		qOk("$select=Id,Name", "$select", func(o *model.Optionen) { o.Select = &model.SelectOption{Namen: []string{"Id", "Name"}} }),
		qOk("$select=*", "$select", func(o *model.Optionen) { o.Select = &model.SelectOption{Alle: true} }),
		qOk("$orderby=Name%20desc", "$orderby", func(o *model.Optionen) { o.OrderBy = []model.Sortierung{{Name: "Name", Absteigend: true}} }),
		qOk("$inlinecount=allpages", "$inlinecount", func(o *model.Optionen) { o.Count = model.WahrOption{Gesetzt: true, Wert: true} }),
		qOk("$inlinecount=none", "$inlinecount", func(o *model.Optionen) { o.Count = model.WahrOption{Gesetzt: true} }),
		qOk("$format=json", "$format", func(*model.Optionen) {}),
		qOk("$format=application/json;odata=verbose", "$format", func(*model.Optionen) {}),
		sonst("a=b", "a", "b"), sonst("@x=1", "@x", "1"), sonst("sap-client=100", "sap-client", "100"),
		qOk("$top=0", "$top", func(o *model.Optionen) { o.Top = model.ZahlOption{Gesetzt: true} }),
	}
	n257 := strings.Join(namenListe(257), ",")
	schlecht = []qForm{
		qU("$top=abc", "$top", model.URLWert), qU("$top=%FF", "$top", model.URLKodierung), qU("%FF=1", "", model.URLKodierung),
		qU("$foo=1", "", model.URLOptionName), qU("$Top=1", "", model.URLOptionName), qU("$select=", "$select", model.URLWert),
		qU("$inlinecount=ALLPAGES", "$inlinecount", model.URLWert), qU("$format=", "$format", model.URLWert),
		qU("$orderby=A%20up", "$orderby", model.URLWert), qU("$select="+n257, "$select", model.URLZuViele),
		qK("$expand=X", "", model.KompatOption), qK("$count=true", "", model.KompatOption), qK("$format=xml", "$format", model.KompatFormat),
		qK("$filter=Id%20eq%201", "$filter", model.KompatFilter), qK("$orderby=tolower(N)", "$orderby", model.KompatOption),
		qK("$select=A/B", "$select", model.KompatOption), qK("$skiptoken=x", "", model.KompatOption),
	}
	return gut, schlecht
}

// orakel nennt die Erwartung an pfad und Optionen in der Rangfolge der Spezifikation: Gesamtform, dann der erste Verstoß der Klasse
// `ungueltig` in der Lesefolge (Pfad vor Query, ein zweites Vorkommen einer behandelten Option ist `option_doppelt`), sonst der erste
// Kompatibilitätsfehler, sonst das Ergebnis aus den Beiträgen der Formen. Die zweite Rückgabe ist das Etikett der Klasse.
func orakel(pf pForm, qs []qForm) (erw, string) {
	if pf.stufe0 != nil {
		return ung(pf.stufe0.u, pf.stufe0.ort), "u:" + string(pf.stufe0.u)
	}
	var kompat *vd
	for i := range pf.vds {
		v := pf.vds[i]
		if v.klasseU {
			return ung(v.u, v.ort), "u:" + string(v.u)
		}
		if kompat == nil {
			kompat = &v
		}
	}
	gesehen := map[string]bool{}
	var o model.Optionen
	for i, q := range qs {
		ort := fmt.Sprintf("query/%d", i+1)
		if q.behandelt != "" && gesehen[q.behandelt] {
			return ung(model.URLOptionDoppelt, ort), "u:" + string(model.URLOptionDoppelt)
		}
		if q.behandelt != "" {
			gesehen[q.behandelt] = true
		}
		switch {
		case q.v == nil:
			q.anwenden(&o)
		case q.v.klasseU:
			return ung(q.v.u, ort), "u:" + string(q.v.u)
		case kompat == nil:
			kompat = &vd{k: q.v.k, ort: ort}
		}
	}
	if kompat != nil {
		return kmp(kompat.k, kompat.ort), "k:" + string(kompat.k)
	}
	return ok(pf.ress, o), "ok"
}

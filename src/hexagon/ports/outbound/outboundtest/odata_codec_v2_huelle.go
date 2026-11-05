package outboundtest

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeilen H1 bis H10 prüfen die Hülle `d` und ihre Mitglieder (SPEC-172 Nr. 2) und die Rangfolge im Deuten (Zusatztabelle Z3).

// v2Nichtobjekte sind die JSON-Werte, die kein Objekt sind, in der Reihenfolge Array, String, Zahl, true, false, null.
func v2Nichtobjekte() []string { return []string{`[]`, `"a"`, `1`, `true`, `false`, `null`} }

func v2FaelleH1() []jfall {
	var out []jfall
	for _, w := range v2Nichtobjekte() {
		out = append(out, v2Roh("H1 Wurzel "+w, w, model.DatenStruktur, "")...)
	}
	return out
}

func v2H1(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH1()) }

// v2WeitereMitglieder sind die Namen neben `d`: ein Mitglied von V4, das Fehlerobjekt, ein fremder und der leere Name.
func v2WeitereMitglieder() []string {
	return []string{"error", "@odata.context", "value", "d2", "", "D", "__metadata"}
}

func v2FaelleH2() []jfall {
	var out []jfall
	ohneD := []string{`{}`, string(v2Datei("v4_fremd_olingo-04.json")), string(v2Datei("v4_fremd_olingo-08.json")), string(v2Datei("v4_fremd_net-21.json")),
		string(v2Datei("v2_echt_24-fehler.json")), `{"error":1}`, `{"@odata.context":"x"}`, `{"value":[]}`, `{"d2":{}}`, `{"":1}`, `{"D":{"results":[]}}`}
	for i, b := range ohneD {
		out = append(out, v2Roh(fmt.Sprintf("H2 ohne d %d", i), b, model.DatenStruktur, "/d")...)
	}
	for _, n := range v2WeitereMitglieder() {
		ort := "/" + n
		for _, d := range []string{`{"results":[]}`, `5`, `{"results":[{"ID":"a"}]}`} {
			out = append(out, v2Roh("H2 "+strconv.Quote(n)+" vor d="+d, `{`+v2NameJSON(n)+`:1,"d":`+d+`}`, model.DatenStruktur, ort)...)
			out = append(out, v2Roh("H2 "+strconv.Quote(n)+" hinter d="+d, `{"d":`+d+`,`+v2NameJSON(n)+`:1}`, model.DatenStruktur, ort)...)
		}
	}
	return out
}

func v2H2(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH2()) }

func v2FaelleH3() []jfall {
	var out []jfall
	elems := func(idx int, bad string) []string {
		out := []string{`{"ID":1}`, `{"ID":2}`, `{"ID":3}`}
		out[idx] = bad
		return out
	}
	for _, bad := range v2Nichtobjekte() {
		for idx := range 3 {
			out = append(out, jfall{name: fmt.Sprintf("H3 Element %d ist %s", idx, bad), m: "R", set: "Products", body: v2KolA(elems(idx, bad)...),
				ursache: model.DatenStruktur, ort: "/d/" + strconv.Itoa(idx)})
		}
	}
	for _, d := range []string{`"x"`, `5`, `true`, `false`, `null`} {
		out = append(out, v2Roh("H3 d="+d, `{"d":`+d+`}`, model.DatenStruktur, "/d")...)
	}
	out = append(out, jfall{name: "H3 gültiges Array", m: "R", set: "Products", body: v2KolA(`{"ID":1}`, `{"ID":2}`),
		nach: func(t *testing.T, r jergebnis) {
			if len(r.seite.Entities) != 2 || r.seite.Anzahl != (model.Anzahl{}) || r.seite.FolgeLink != (model.FolgeLink{}) {
				t.Errorf("H3 gültiges Array: %+v, erwartet zwei Entities, Anzahl und FolgeLink als Nullwerte", r.seite)
			}
		}})
	return out
}

func v2H3(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH3()) }

func v2FaelleH4() []jfall {
	var out []jfall
	for _, v := range []string{`{}`, `"a"`, `1`, `true`, `null`} {
		out = append(out, jfall{name: "H4 results=" + v, m: "R", set: "Products", body: []byte(`{"d":{"results":` + v + `}}`), ursache: model.DatenStruktur, ort: "/d/results"})
	}
	for _, b := range []string{`{"d":{}}`, `{"d":{"__count":"1"}}`, `{"d":{"__next":"x"}}`, `{"d":{"__count":"1","__next":"x"}}`} {
		out = append(out, jfall{name: "H4 results fehlt " + b, m: "R", set: "Products", body: []byte(b), ursache: model.DatenStruktur, ort: "/d/results"})
	}
	for _, bad := range v2Nichtobjekte() {
		for idx := range 3 {
			el := []string{`{"ID":1}`, `{"ID":2}`, `{"ID":3}`}
			el[idx] = bad
			out = append(out, jfall{name: fmt.Sprintf("H4 Element %d ist %s", idx, bad), m: "R", set: "Products", body: v2Kol(el...),
				ursache: model.DatenStruktur, ort: "/d/results/" + strconv.Itoa(idx)})
		}
	}
	for _, n := range []string{"__metadata", "__deferred", "value", "d", "results2", "@odata.count", "Results"} {
		out = append(out,
			jfall{name: "H4 fremdes Mitglied " + n + " vor results", m: "R", set: "Products", body: []byte(`{"d":{` + v2NameJSON(n) + `:{},"results":[]}}`),
				ursache: model.DatenStruktur, ort: "/d/" + n},
			jfall{name: "H4 fremdes Mitglied " + n + " hinter results", m: "R", set: "Products", body: []byte(`{"d":{"results":[],` + v2NameJSON(n) + `:{}}}`),
				ursache: model.DatenStruktur, ort: "/d/" + n})
	}
	return out
}

func v2H4(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH4()) }

// v2FaelleH5: das leere Ergebnis in den drei Formen (LH-FA-17, Boundary): Entities leer und nicht nil.
func v2FaelleH5() []jfall {
	leer := func(t *testing.T, r jergebnis) {
		if r.seite.Entities == nil || len(r.seite.Entities) != 0 {
			t.Errorf("H5: Entities %#v, erwartet leer und nicht nil", r.seite.Entities)
		}
	}
	return []jfall{
		{name: "H5 Objektform", m: "R", set: "Products", body: []byte(`{"d":{"results":[]}}`), nach: leer},
		{name: "H5 Array-Form", m: "R", set: "Products", body: []byte(`{"d":[]}`), nach: leer},
		{name: "H5 mit Anzahl", m: "R", set: "Products", body: []byte(`{"d":{"__count":"0","results":[]}}`), nach: func(t *testing.T, r jergebnis) {
			leer(t, r)
			if r.seite.Anzahl != (model.Anzahl{Vorhanden: true}) {
				t.Errorf("H5 mit Anzahl: %+v", r.seite.Anzahl)
			}
		}},
		{name: "H5 Echtdatei 22", m: "R", set: "Employees", body: v2Datei("v2_echt_22-leer.json"), nach: leer},
	}
}

func v2H5(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH5()) }

// v2CountFall ist eine Schreibweise von `__count`: gültig mit dem Wert, sonst mit der Ursache.
type v2CountFall struct {
	json    string
	wert    int64
	ursache model.DatenUrsache
}

// v2CountFaelle sind mehr als 30 Schreibweisen von `__count` (SPEC-172 Nr. 2 a).
func v2CountFaelle() []v2CountFall {
	return []v2CountFall{
		{`"6"`, 6, ""}, {`6`, 6, ""}, {`"0"`, 0, ""}, {`0`, 0, ""}, {`"9223372036854775807"`, 9223372036854775807, ""},
		{`9223372036854775807`, 9223372036854775807, ""}, {`9223372036854775808`, 0, model.DatenWert}, {`"9223372036854775808"`, 0, model.DatenWert},
		{`"-1"`, 0, model.DatenWert}, {`-1`, 0, model.DatenWert}, {`"1.5"`, 0, model.DatenWert}, {`1.5`, 0, model.DatenWert},
		{`"1e2"`, 100, ""}, {`1e2`, 100, ""}, {`"007"`, 0, model.DatenWert}, {`""`, 0, model.DatenWert}, {`"abc"`, 0, model.DatenWert},
		{`" 6"`, 0, model.DatenWert}, {`"+6"`, 0, model.DatenWert}, {`"6 "`, 0, model.DatenWert}, {`true`, 0, model.DatenTyp}, {`false`, 0, model.DatenTyp},
		{`null`, 0, model.DatenTyp}, {`{}`, 0, model.DatenTyp}, {`[]`, 0, model.DatenTyp}, {`"[REDACTED]"`, 0, model.DatenWert},
		{`6.0`, 6, ""}, {`"6.0"`, 6, ""}, {`-0`, 0, ""}, {`"-0"`, 0, ""}, {`0.5e1`, 5, ""}, {`1e-1`, 0, model.DatenWert}, {`9e18`, 9000000000000000000, ""},
		{`1e19`, 0, model.DatenWert}, {`"1e19"`, 0, model.DatenWert}, {`1e999999999`, 0, model.DatenWert}, {`"0x10"`, 0, model.DatenWert},
	}
}

func v2FaelleH6() []jfall {
	var out []jfall
	for _, c := range v2CountFaelle() {
		for _, vor := range []bool{true, false} {
			body := `{"d":{"results":[],"__count":` + c.json + `}}`
			if vor {
				body = `{"d":{"__count":` + c.json + `,"results":[]}}`
			}
			f := jfall{name: fmt.Sprintf("H6 __count=%s vor results %v", c.json, vor), m: "R", set: "Products", body: []byte(body), ursache: c.ursache}
			if c.ursache != "" {
				f.ort = "/d/__count"
			} else {
				c := c
				f.nach = func(t *testing.T, r jergebnis) {
					if r.seite.Anzahl != (model.Anzahl{Vorhanden: true, Wert: c.wert}) {
						t.Errorf("H6 __count=%s: Anzahl %+v, erwartet %d", c.json, r.seite.Anzahl, c.wert)
					}
				}
			}
			out = append(out, f)
		}
	}
	return append(out, jfall{name: "H6 fehlt", m: "R", set: "Products", body: v2Kol(), nach: func(t *testing.T, r jergebnis) {
		if r.seite.Anzahl != (model.Anzahl{}) {
			t.Errorf("H6 fehlt: Anzahl %+v, erwartet den Nullwert", r.seite.Anzahl)
		}
	}})
}

func v2H6(t *testing.T, s *jsuite) {
	if n := len(v2CountFaelle()); n < 30 {
		t.Errorf("H6: %d Schreibweisen, erwartet mindestens 30", n)
	}
	s.pruefeAlle(t, v2FaelleH6())
}

func v2FaelleH7() []jfall {
	var out []jfall
	for _, c := range []struct {
		json    string
		link    model.FolgeLink
		ursache model.DatenUrsache
	}{
		{`"Products?$skip=3"`, model.FolgeLink{Vorhanden: true, Wert: "Products?$skip=3"}, ""},
		{`"http://h/svc/Products?$skiptoken=1"`, model.FolgeLink{Vorhanden: true, Wert: "http://h/svc/Products?$skiptoken=1"}, ""},
		{`"a\n\"b"`, model.FolgeLink{Vorhanden: true, Wert: "a\n\"b"}, ""},
		{`"x"`, model.FolgeLink{Vorhanden: true, Wert: "x"}, ""},
		{`"[REDACTED]"`, model.FolgeLink{Vorhanden: true, Redigiert: true}, ""},
		{`"` + esc("005B") + `REDACTED]"`, model.FolgeLink{Vorhanden: true, Wert: "[REDACTED]"}, ""},
		{`"[redacted]"`, model.FolgeLink{Vorhanden: true, Wert: "[redacted]"}, ""},
		{`"[REDACTED] "`, model.FolgeLink{Vorhanden: true, Wert: "[REDACTED] "}, ""},
		{`""`, model.FolgeLink{}, model.DatenWert}, {`null`, model.FolgeLink{}, model.DatenTyp}, {`5`, model.FolgeLink{}, model.DatenTyp},
		{`true`, model.FolgeLink{}, model.DatenTyp}, {`{}`, model.FolgeLink{}, model.DatenTyp}, {`[]`, model.FolgeLink{}, model.DatenTyp},
	} {
		for _, vor := range []bool{true, false} {
			body := `{"d":{"results":[],"__next":` + c.json + `}}`
			if vor {
				body = `{"d":{"__next":` + c.json + `,"results":[]}}`
			}
			f := jfall{name: fmt.Sprintf("H7 __next=%s vor results %v", c.json, vor), m: "R", set: "Products", body: []byte(body), ursache: c.ursache}
			if c.ursache != "" {
				f.ort = "/d/__next"
			} else {
				c := c
				f.nach = func(t *testing.T, r jergebnis) {
					if r.seite.FolgeLink != c.link {
						t.Errorf("H7 __next=%s: FolgeLink %+v, erwartet %+v", c.json, r.seite.FolgeLink, c.link)
					}
				}
			}
			out = append(out, f)
		}
	}
	return out
}

func v2H7(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH7()) }

// v2FaelleH8: bei ParseEntity sind `results`, `__count` und `__next` gewöhnliche Namen; `d` ist ein Objekt.
func v2FaelleH8() []jfall {
	var out []jfall
	out = append(out, jfall{name: "H8 Properties dieser Namen", m: "W", set: "Wn", einzel: true, body: v2Ein(`"Id":1,"results":"x","__count":5,"__next":"y"`),
		props: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("results", model.ArtString, "x"), pw("__count", model.ArtInt32, "5"), pw("__next", model.ArtString, "y")}})
	out = append(out, jfall{name: "H8 Properties dieser Namen in einer Collection", m: "W", set: "Wn", body: v2Kol(`{"Id":1,"__count":5}`),
		props: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("__count", model.ArtInt32, "5")}})
	for _, n := range []string{"results", "__count", "__next"} {
		out = append(out, jfall{name: "H8 " + n + " ohne Property", m: "R", set: "Products", einzel: true, body: v2Ein(`"ID":1,"` + n + `":"1"`),
			ursache: model.DatenUnbekannteProperty, ort: "/d/" + n})
	}
	for _, d := range v2Nichtobjekte() {
		out = append(out, jfall{name: "H8 d=" + d, m: "R", set: "Products", einzel: true, body: []byte(`{"d":` + d + `}`), ursache: model.DatenStruktur, ort: "/d"})
	}
	return out
}

func v2H8(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH8()) }

// v2FaelleH9: die Methode entscheidet, nicht der Body (SPEC-172 Nr. 2 b, E3): ein Collection-Body an ParseEntity ist
// `unbekannte_property` an `results`, ein Entity-Body an ParseCollection `struktur` am ersten Mitglied von `d`.
func v2FaelleH9() []jfall {
	return []jfall{
		{name: "H9 Echtdatei 14 an ParseEntity", m: "R", set: "Employees", einzel: true, body: v2Datei("v2_echt_14-top.json"),
			ursache: model.DatenUnbekannteProperty, ort: "/d/results"},
		{name: "H9 Echtdatei 08 an ParseCollection", m: "R", set: "Employees", body: v2Datei("v2_echt_08-entity.json"),
			ursache: model.DatenStruktur, ort: "/d/__metadata"},
		{name: "H9 Echtdatei 20 an ParseEntity", m: "R", set: "Employees", einzel: true, body: v2Datei("v2_echt_20-inlinecount.json"),
			ursache: model.DatenUnbekannteProperty, ort: "/d/__count"},
		{name: "H9 Entity ohne __metadata an ParseCollection", m: "R", set: "Products", body: v2Ein(`"ID":1`), ursache: model.DatenStruktur, ort: "/d/ID"},
	}
}

func v2H9(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH9()) }

// v2FaelleH10 sind die Paare der Zusatztabelle Z3 (Rangfolge im Deuten, SPEC-172 Nr. 2 c): ein Fall je Paar, wo die Reihenfolge
// im Dokument zählt, in beiden Reihenfolgen.
func v2FaelleH10() []jfall {
	var out []jfall
	k := func(name, body string, u model.DatenUrsache, ort string) {
		out = append(out, jfall{name: "H10 " + name, m: "R", set: "Products", body: []byte(body), ursache: u, ort: ort})
	}
	e := func(m, set, name, body string, u model.DatenUrsache, ort string) {
		out = append(out, jfall{name: "H10 " + name, m: m, set: set, einzel: true, body: []byte(body), ursache: u, ort: ort})
	}
	st := model.DatenStruktur
	k("d fehlt und weiteres Mitglied", `{"@odata.context":"x","value":[{"ID":"a"}]}`, st, "/d")
	k("Mitglied vor d gegen Fehler in d", `{"x":1,"d":{"results":[{"ID":"a"}]}}`, st, "/x")
	k("Mitglied hinter d gegen Fehler in d", `{"d":{"results":[{"ID":"a"}]},"x":1}`, st, "/x")
	k("results falsch vor fremdem Mitglied", `{"d":{"results":5,"zzz":1}}`, st, "/d/results")
	k("fremdes Mitglied vor results falsch", `{"d":{"zzz":1,"results":5}}`, st, "/d/zzz")
	k("__count falsch vor results gegen Fehler im Element", `{"d":{"__count":"x","results":[{"ID":"a"}]}}`, model.DatenWert, "/d/__count")
	k("__count falsch hinter results gegen Fehler im Element", `{"d":{"results":[{"ID":"a"}],"__count":"x"}}`, model.DatenTyp, "/d/results/0/ID")
	k("__next falsch vor results gegen Fehler im Element", `{"d":{"__next":5,"results":[{"ID":"a"}]}}`, model.DatenTyp, "/d/__next")
	k("__next falsch hinter results gegen Fehler im Element", `{"d":{"results":[{"ID":"a"}],"__next":5}}`, model.DatenTyp, "/d/results/0/ID")
	k("Fehler in Element 0 und Element 1", `{"d":{"results":[{"ID":"a"},{"ID":"b"}]}}`, model.DatenTyp, "/d/results/0/ID")
	k("Fehler in Element 1 und fehlendes results", `{"d":{"__count":"1"}}`, st, "/d/results")
	k("Array: Fehler in Element 0 und Element 1", `{"d":[{"ID":"a"},{"ID":"b"}]}`, model.DatenTyp, "/d/0/ID")
	k("Element kein Objekt vor Fehler im Element", `{"d":{"results":[1,{"ID":"b"}]}}`, st, "/d/results/0")
	k("odata_typ gegen frühere unbekannte Property", `{"d":{"results":[{"zzz":1,"__metadata":{"type":"Nope"},"ID":1}]}}`, model.DatenODataTyp, "/d/results/0/__metadata/type")
	k("__metadata kein Objekt vor Fehler einer Property", `{"d":{"results":[{"__metadata":5,"ID":"x"}]}}`, st, "/d/results/0/__metadata")
	k("__metadata kein Objekt hinter Fehler einer Property", `{"d":{"results":[{"ID":"x","__metadata":5}]}}`, st, "/d/results/0/__metadata")
	k("Fehler einer Property gegen fehlenden Schlüssel", `{"d":{"results":[{"Name":5}]}}`, model.DatenTyp, "/d/results/0/Name")
	e("T", "K2", "zwei fehlende Schlüssel", `{"d":{"Note":"x"}}`, model.DatenSchluessel, "/d/A")
	e("T", "KS", "Marker am ersten Schlüssel und fehlender zweiter", `{"d":{"A":"[REDACTED]"}}`, model.DatenSchluessel, "/d/A")
	e("T", "KS", "Marker am zweiten Schlüssel und fehlender erster", `{"d":{"B":"[REDACTED]"}}`, model.DatenSchluessel, "/d/B")
	e("R", "Employees", "Navigation typ vor unbekannter Property", `{"d":{"EmployeeId":"1","ne_Team":5,"zzz":1}}`, model.DatenTyp, "/d/ne_Team")
	e("R", "Employees", "unbekannte Property vor Navigation typ", `{"d":{"EmployeeId":"1","zzz":1,"ne_Team":5}}`, model.DatenUnbekannteProperty, "/d/zzz")
	return out
}

func v2H10(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleH10()) }

func v2FaelleH() []jfall {
	return verbinde(v2FaelleH1(), v2FaelleH2(), v2FaelleH3(), v2FaelleH4(), v2FaelleH5(), v2FaelleH6(), v2FaelleH7(), v2FaelleH8(), v2FaelleH9(), v2FaelleH10())
}

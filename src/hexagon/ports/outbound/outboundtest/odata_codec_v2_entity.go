package outboundtest

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeilen N1 bis N6 prüfen die Entity (SPEC-172 Nr. 3, E4): `__metadata`, die Properties, die übersprungenen Namen und die
// unbekannten Namen.

// v2Typfall ist ein Wert von `__metadata.type` mit dem Ausgang: der Typ der Entity oder eine Ursache.
type v2Typfall struct {
	json    string
	typ     string
	ursache model.DatenUrsache
}

func v2Typfaelle() []v2Typfall {
	f := model.DatenODataTyp
	return []v2Typfall{
		{`"Z.A"`, "Z.A", ""}, {`"Z.A1"`, "Z.A1", ""}, {`"Z.A2"`, "Z.A2", ""},
		{`5`, "", f}, {`true`, "", f}, {`false`, "", f}, {`null`, "", f}, {`{}`, "", f}, {`[]`, "", f},
		{`""`, "", f}, {`"Nope"`, "", f}, {`"#Z.A1"`, "", f}, {`"Z.A1 "`, "", f}, {`" Z.A1"`, "", f}, {`"z.a1"`, "", f}, {`"Z.a1"`, "", f},
		{`"Z.Other"`, "", f}, {`"Z.AbsD"`, "", f}, {`"Z.B"`, "", f}, {`"Z.B1"`, "", f}, {`"Z.A1x"`, "", f}, {`"A1"`, "", f}, {`"Alias.A1"`, "", f},
		{`"` + esc("005A") + `.A1"`, "Z.A1", ""},
	}
}

func v2FaelleN1() []jfall {
	var out []jfall
	for _, w := range v2Nichtobjekte() {
		out = append(out, v2Beide("N1 __metadata="+w, "Z", "As", `"__metadata":`+w+`,"Id":1`, model.DatenStruktur, "/__metadata")...)
		out = append(out, v2Beide("N1 __metadata="+w+" hinter den Properties", "Z", "As", `"Id":1,"__metadata":`+w, model.DatenStruktur, "/__metadata")...)
	}
	out = append(out, v2BeideGut("N1 __metadata ohne type", "Z", "As", `"__metadata":{},"Id":1`, pw("Id", model.ArtInt32, "1"))...)
	for _, c := range v2Typfaelle() {
		for _, davor := range []bool{true, false} {
			eigene, props := `"Id":1`, []model.PropertyWert{pw("Id", model.ArtInt32, "1")}
			if c.typ != "Z.A" {
				eigene += `,"E1":5`
				props = append(props, pw("E1", model.ArtInt32, "5"))
			}
			members := `"__metadata":{"type":` + c.json + `},` + eigene
			if !davor {
				members = eigene + `,"__metadata":{"type":` + c.json + `}`
			}
			name := fmt.Sprintf("N1 type=%s davor %v", c.json, davor)
			if c.ursache != "" {
				out = append(out, v2Beide(name, "Z", "As", members, c.ursache, "/__metadata/type")...)
				continue
			}
			for _, f := range v2Beide(name, "Z", "As", members, "", "") {
				f.typ, f.schluessel, f.props = c.typ, "1", props
				out = append(out, f)
			}
		}
	}
	// Die Property eines abgeleiteten Typs vor __metadata ist eine Property, nicht unbekannt (Reihenfolge der Mitglieder frei).
	out = append(out, v2BeideGut("N1 Property des abgeleiteten Typs vor __metadata", "Z", "As", `"E2":"x","E1":5,"Id":1,"__metadata":{"type":"Z.A2"}`,
		pw("E2", model.ArtString, "x"), pw("E1", model.ArtInt32, "5"), pw("Id", model.ArtInt32, "1"))...)
	out = append(out, v2BeideGut("N1 Typ des Sets ohne __metadata", "Z", "As", `"Id":1,"Name":"n"`, pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "n"))...)
	out = append(out, v2Beide("N1 abgeleitete Property ohne __metadata", "Z", "As", `"Id":1,"E1":5`, model.DatenUnbekannteProperty, "/E1")...)
	out = append(out, v2BeideGut("N1 abstrakter Typ des Sets, abgeleiteter in __metadata", "Z", "Bs", `"__metadata":{"type":"Z.B1"},"Id":1,"Name":"n"`,
		pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "n"))...)
	out = append(out, v2Beide("N1 abstrakter Typ des Sets ohne __metadata hat keinen Schlüssel", "Z", "Bs", `"Name":"n"`, model.DatenSchluessel, "")...)
	return out
}

func v2N1(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleN1())
	v2N1Kreis(t, s)
}

// v2N1Kreis: ein Typ in `__metadata.type`, dessen Kette von `BaseType` sich schließt und den Typ des Sets nicht erreicht, ist
// kein abgeleiteter Typ des Sets (SPEC-172 Nr. 3, E4) und endet in `odata_typ`; der Leser kehrt zurück, statt der Kette ohne Ende
// zu folgen. Schleife über die Formen des Kreises (Selbstbezug, Ring über zwei und drei Typen, Anhang an einem Ring) und die
// drei Formen des Bodys.
func v2N1Kreis(t *testing.T, s *jsuite) {
	typen := []string{"Y.Self", "Y.R1", "Y.R2", "Y.Q1", "Y.Q2", "Y.Q3", "Y.T"}
	n := 0
	for _, typ := range typen {
		for _, f := range v2Beide("N1 Kreis "+typ, "Y", "Ss", `"__metadata":{"type":"`+typ+`"},"Id":1`, model.DatenODataTyp, "/__metadata/type") {
			n++
			fertig := make(chan jergebnis, 1)
			go func() { fertig <- s.lauf(f) }()
			select {
			case r := <-fertig:
				datenFehler(t, f.name, r.err, f.ursache, f.ort)
			case <-time.After(10 * time.Second):
				t.Fatalf("%s: der Leser kehrt nicht zurück", f.name)
			}
		}
	}
	if n != 3*len(typen) {
		t.Errorf("N1 Kreis: %d Läufe, erwartet %d", n, 3*len(typen))
	}
}

// v2Metawerte sind sechs Wertformen eines Mitglieds von `__metadata`.
func v2Metawerte() []string { return []string{`"x"`, `1`, `null`, `{"a":[1]}`, `[1]`, `true`} }

func v2FaelleN2() []jfall {
	var out []jfall
	for _, n := range []string{"uri", "etag", "id", "content_type", "media_src", "edit_media", "foo", "", "Type", "@odata.type"} {
		for _, w := range v2Metawerte() {
			out = append(out, v2BeideGut(fmt.Sprintf("N2 %q=%s", n, w), "R", "Products", `"__metadata":{`+v2NameJSON(n)+`:`+w+`},"ID":1,"Name":"a"`,
				pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, "a"))...)
		}
	}
	for i := range out {
		out[i].ueber = []model.Uebersprungen{}
		out[i].typ = "Demo.Product"
	}
	return out
}

func v2N2(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleN2()) }

// v2Muster ist ein gültiger Wert je Typ der sieben mit der erwarteten Art und dem Text.
func v2Muster(edm string) (json string, art model.WertArt, text string) {
	switch edm {
	case "Edm.String":
		return `"a"`, model.ArtString, "a"
	case "Edm.Boolean":
		return `true`, model.ArtBoolean, "true"
	case "Edm.Int32":
		return `1`, model.ArtInt32, "1"
	case "Edm.Int64":
		return `"5"`, model.ArtInt64, "5"
	case "Edm.Decimal":
		return `1.50`, model.ArtDecimal, "1.5"
	case "Edm.Guid":
		return `"0F8FAD5B-D9CB-469F-A165-70867728950E"`, model.ArtGUID, guidKlein
	}
	return v2Dt("0"), model.ArtDateTimeOffset, "1970-01-01T00:00:00Z"
}

func v2FaelleN3() []jfall {
	var out []jfall
	id := pw("Id", model.ArtInt32, "1")
	for _, edm := range v2Typen() {
		n, r := "N"+kurz(edm), "R"+kurz(edm)
		j, art, text := v2Muster(edm)
		out = append(out, v2BeideGut("N3 "+n+" fehlt", "W", "Ws", `"Id":1`, id)...)
		out = append(out, v2BeideGut("N3 "+n+" null", "W", "Ws", `"Id":1,"`+n+`":null`, id, pw(n, model.ArtNull, ""))...)
		out = append(out, v2BeideGut("N3 "+n+" Wert", "W", "Ws", `"Id":1,"`+n+`":`+j, id, pw(n, art, text))...)
		out = append(out, v2Beide("N3 "+r+" null", "W", "Ws", `"Id":1,"`+r+`":null`, model.DatenNull, "/"+r)...)
		out = append(out, v2BeideGut("N3 "+r+" fehlt", "W", "Ws", `"Id":1`, id)...)
		// Die Reihenfolge der Properties ist die des Dokuments, nicht die des Modells.
		out = append(out, v2BeideGut("N3 Reihenfolge "+r+" "+n, "W", "Ws", `"`+r+`":`+j+`,"`+n+`":`+j+`,"Id":1`, pw(r, art, text), pw(n, art, text), id)...)
	}
	return out
}

func v2N3(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleN3()) }

// v2UebersprungenWerte sind sieben JSON-Werte für eine Property, die der Leser nicht deutet, auch ein tiefes Objekt und ein
// gültiger Datumswert.
func v2UebersprungenWerte() []string {
	return []string{`"x"`, `1`, `true`, `null`, `{}`, `[1]`, `{"a":{"b":{"c":[1,{"d":null}]}}}`, v2Dt("915148800000"), `"2020-01-01"`}
}

func v2FaelleN4() []jfall {
	var out []jfall
	for i, edm := range v2Uebersprungen() {
		x := "X" + strconv.Itoa(i)
		for _, w := range v2UebersprungenWerte() {
			fs := v2BeideGut(fmt.Sprintf("N4 %s (%s)=%.30s", x, edm, w), "W", "Ws", `"Id":1,"`+x+`":`+w, pw("Id", model.ArtInt32, "1"))
			for k := range fs {
				fs[k].ueber = []model.Uebersprungen{{Name: x, Grund: model.UebersprungTyp}}
			}
			out = append(out, fs...)
		}
	}
	return out
}

func v2N4(t *testing.T, s *jsuite) {
	if n := len(v2Uebersprungen()); n != 8 {
		t.Errorf("N4: %d Typen, erwartet 8", n)
	}
	s.pruefeAlle(t, v2FaelleN4())
}

func v2FaelleN5() []jfall {
	var out []jfall
	nav := []string{`{"__deferred":{"uri":"x"}}`, `{"__metadata":{"type":"T.W"},"Id":2}`, `{"results":[{"Id":2}]}`, `[]`, `[{"Id":2}]`, `null`, `{}`,
		`{"a":{"b":[1,{"c":null}]}}`}
	for _, n := range []string{"Nav", "Navs"} {
		for _, w := range nav {
			fs := v2BeideGut(fmt.Sprintf("N5 %s=%.30s", n, w), "W", "Ws", `"Id":1,"`+n+`":`+w, pw("Id", model.ArtInt32, "1"))
			for k := range fs {
				fs[k].ueber = []model.Uebersprungen{{Name: n, Grund: model.UebersprungNavigation}}
			}
			out = append(out, fs...)
		}
		for _, w := range []string{`"x"`, `1`, `true`, `false`, `1.5`, `""`} {
			out = append(out, v2Beide(fmt.Sprintf("N5 %s=%s", n, w), "W", "Ws", `"Id":1,"`+n+`":`+w, model.DatenTyp, "/"+n)...)
		}
	}
	return out
}

func v2N5(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleN5()) }

// v2Namen6 sind mehr als zwölf Namen, die in keinem V2-Modell eine Property sind: Annotationen von V4, Schreibweisen von Namen
// des Modells, `results`, `__count` und `__next`, der leere Name, Namen mit anderer Zusammensetzung von Zeichen.
func v2Namen6() []string {
	return []string{"@odata.type", "Price@odata.type", "@", "@odata.id", "a@b@c", "Name@", "id", "ID", "iD", "Id ", " Id", "results", "__count", "__next", "__metadata2",
		"value", "", "Id\x00", "\xc3\xa9", "e\xcc\x81", "X", "Xa"}
}

func v2FaelleN6() []jfall {
	var out []jfall
	for _, set := range []string{"Ws", "Wo"} {
		for _, n := range v2Namen6() {
			for _, w := range wertformen() {
				out = append(out, v2Beide(fmt.Sprintf("N6 %s %q=%s", set, n, w), "W", set, `"Id":1,`+v2NameJSON(n)+`:`+w, model.DatenUnbekannteProperty, "/"+v2Zeiger(n))...)
			}
		}
	}
	return out
}

func v2N6(t *testing.T, s *jsuite) {
	if n := len(v2Namen6()); n < 12 {
		t.Errorf("N6: %d Namen, erwartet mindestens 12", n)
	}
	s.pruefeAlle(t, v2FaelleN6())
}

func v2FaelleN() []jfall {
	return verbinde(v2FaelleN1(), v2FaelleN2(), v2FaelleN3(), v2FaelleN4(), v2FaelleN5(), v2FaelleN6())
}

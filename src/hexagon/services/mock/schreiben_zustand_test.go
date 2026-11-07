package mock_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// folgeNamen ist die Folge der Bytes der neun Namen von Product, von Hand geschrieben (nicht aus dem Modell).
func folgeNamen() []string {
	return []string{"Big", "Count", "Day", "Flag", "Id", "Name", "Price", "Ref", "Ts"}
}

// erwPayload schreibt den Payload der Kanonform: jedes Mitglied der Folge, `null`, wo werte nichts nennt.
func erwPayload(werte map[string]string) string {
	var teile []string
	for _, n := range folgeNamen() {
		roh, da := werte[n]
		if !da {
			roh = "null"
		}
		teile = append(teile, `"`+n+`":`+roh)
	}
	return "{" + strings.Join(teile, ",") + "}"
}

// stand ist der Eintrag mit Version, Payload und Zustand; er wird mit dem Soll verglichen.
func gleichEintrag(a, b model.EntityState) bool {
	return a.WorkspaceID == b.WorkspaceID && a.EntitySet == b.EntitySet && a.EntityKey == b.EntityKey && a.Payload == b.Payload &&
		slices.Equal(a.KnownProperties, b.KnownProperties) && a.Completeness == b.Completeness && a.Version == b.Version && a.Deleted == b.Deleted
}

// teilEintrag ist ein partieller Eintrag: nur Id und Name bekannt.
func teilEintrag(id string, version int64) model.EntityState {
	return model.EntityState{
		WorkspaceID: wsID, EntitySet: setProducts, EntityKey: id, Completeness: model.EntityPartial, Version: version,
		Payload: `{"Id":` + id + `,"Name":"t` + id + `"}`, KnownProperties: []string{"Id", "Name"},
	}
}

// fuenfEintraege sind die Einträge 1 bis 5 von Products (Schlüssel in der Folge der Bytes), Namen n1 bis n5.
func fuenfEintraege() []model.EntityState {
	var r []model.EntityState
	for _, k := range []string{"1", "2", "3", "4", "5"} {
		e := voll(int(k[0]-'0'), "n"+k)
		r = append(r, e)
	}
	return r
}

// zustandsFall bereitet den Eintrag p vor: 0 complete, 1 partial, 2 Tombstone (Version 3), 3 bis 5 fehlend bei Coverage
// complete, partial, unknown.
func zustandsFall(u *umg, s int, p string) {
	es := fuenfEintraege()
	i := int(p[0] - '1')
	switch s {
	case 1:
		es[i] = teilEintrag(p, 2)
	case 2:
		es[i].Deleted, es[i].Version = true, 3
	case 3, 4, 5:
		es = slices.Delete(es, i, i+1)
		u.co.m = map[string]model.Coverage{}
		if s != 5 {
			u.co.setze(baselineID, setProducts, map[int]model.CoverageStatus{3: model.CoverageComplete, 4: model.CoveragePartial}[s])
		}
	}
	u.en.setze(setProducts, es...)
}

// TestZustandsmatrix (W20 bis W26, W40, M10, M25): Methode 4 × Eintrag 6 × Position im Set 3 × STRICT_COVERAGE 2 = 144; das
// Orakel ist die Tabelle von SPEC-193 Nr. 2 von Hand, Statuscodes als Literale.
func TestZustandsmatrix(t *testing.T) {
	zustaende := []string{"complete", "partial", "Tombstone", "fehlend complete", "fehlend partial", "fehlend unknown"}
	code := map[string][6]model.Code{
		"POST":   {e6009, e6009, 0, 0, 0, 0},
		"PATCH":  {0, e6003, e6008, e6008, e6003, e6003},
		"PUT":    {0, 0, e6008, e6008, e6003, e6003},
		"DELETE": {0, 0, e6008, e6008, e6003, e6003},
	}
	coverageGet := [6]int{0, 0, 0, 1, 1, 1}
	status := map[model.Code]int{e6009: 409, e6008: 404, e6003: 422}
	n := 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		for s, zname := range zustaende {
			for _, p := range []string{"1", "3", "5"} {
				for _, strict := range []bool{false, true} {
					n++
					u := schreibUmg(t)
					u.k.StrictCoverage = strict
					zustandsFall(u, s, p)
					vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
					a, err := u.mutiere(m, p, eingabeIDName(p, "neu-"+m))
					kontext := fmt.Sprintf("%s %s Position %s strict=%v", m, zname, p, strict)
					wantCode := code[m][s]
					wantCov := coverageGet[s]
					if m == "POST" {
						wantCov = 0
					}
					if got := u.p.anzahl("Coverage.Get"); got != wantCov {
						t.Errorf("%s: Coverage.Get %d, erwartet %d", kontext, got, wantCov)
					}
					if wantCov == 1 {
						if ev := u.p.von("Coverage.Get")[0]; ev.args[0] != baselineID || ev.args[1] != setProducts {
							t.Errorf("%s: Coverage.Get(%v, %v)", kontext, ev.args[0], ev.args[1])
						}
					}
					if wantCode != 0 {
						f := erwarteFehler(t, kontext, err, wantCode)
						if f != nil && f.Status != status[wantCode] {
							t.Errorf("%s: Status %d, erwartet %d", kontext, f.Status, status[wantCode])
						}
						u.ohneSchreiben(kontext)
						u.unveraendert(vorher, rev, kontext)
						continue
					}
					erwarteStatus(t, kontext, a, err, sollStatus(m))
					if u.p.anzahl("Entities.Put") != 1 || u.p.anzahl("Workspaces.Advance") != 1 || u.ws.revision(wsID) != rev+1 {
						t.Errorf("%s: Put %d Advance %d Revision %d", kontext, u.p.anzahl("Entities.Put"), u.p.anzahl("Workspaces.Advance"), u.ws.revision(wsID))
					}
					soll := erwartetNach(m, s, p)
					got := u.zustand()[p]
					if !gleichEintrag(got, soll) {
						t.Errorf("%s: Eintrag %+v, erwartet %+v", kontext, got, soll)
					}
					// Die übrigen Einträge bleiben unverändert, die Folge der Schlüssel bleibt aufsteigend.
					nachher := u.en.inhalt()[setProducts]
					for _, x := range nachher {
						if x.EntityKey == p {
							continue
						}
						i := slices.IndexFunc(vorher[setProducts], func(y model.EntityState) bool { return y.EntityKey == x.EntityKey })
						if i < 0 || !gleichEintrag(x, vorher[setProducts][i]) {
							t.Errorf("%s: Eintrag %s verändert", kontext, x.EntityKey)
						}
					}
					if !slices.IsSortedFunc(nachher, func(a, b model.EntityState) int { return strings.Compare(a.EntityKey, b.EntityKey) }) {
						t.Errorf("%s: Folge der Schlüssel nicht aufsteigend", kontext)
					}
				}
			}
		}
	}
	if n != 144 {
		t.Errorf("Menge %d, erwartet 144", n)
	}
}

// erwartetNach ist der Eintrag nach einer erfolgreichen Mutation, von Hand aus SPEC-193 Nr. 3 und 4.
func erwartetNach(m string, s int, p string) model.EntityState {
	name := "neu-" + m
	neu := model.EntityState{
		WorkspaceID: wsID, EntitySet: setProducts, EntityKey: p, Completeness: model.EntityComplete, KnownProperties: alleProperties(),
		Payload: erwPayload(map[string]string{"Id": p, "Name": `"` + name + `"`}),
	}
	alt := voll(int(p[0]-'0'), "n"+p)
	switch {
	case m == "POST" && s == 2:
		neu.Version = 4
	case m == "POST":
		neu.Version = 1
	case m == "PUT" && s == 0:
		neu.Version = 2
	case m == "PUT" && s == 1:
		neu.Version = 3
	case m == "PATCH":
		neu.Payload = strings.Replace(alt.Payload, `"Name":"n`+p+`"`, `"Name":"`+name+`"`, 1)
		neu.Version = 2
	case m == "DELETE" && s == 0:
		alt.Deleted, alt.Version = true, 2
		return alt
	default: // DELETE bei partial
		t := teilEintrag(p, 2)
		t.Deleted, t.Version = true, 3
		return t
	}
	return neu
}

// TestPostUeberTombstone (W22): POST über einen Tombstone legt neu an: Deleted falsch, Version alt plus 1, Payload und
// KnownProperties ersetzt, nicht gemischt. Version 3 (0, 1, 7) × Known alt 3 (Teilmenge, gleich, mehr als neu) = 9.
func TestPostUeberTombstone(t *testing.T) {
	bekannt := [][]string{{"Id"}, alleProperties(), append(alleProperties(), "Zusatz")}
	n := 0
	for _, v := range []int64{0, 1, 7} {
		for _, k := range bekannt {
			n++
			u := schreibUmg(t)
			tomb := voll(7, "alt")
			tomb.Deleted, tomb.Version, tomb.KnownProperties = true, v, k
			u.en.setze(setProducts, tomb)
			a, err := u.mutiere("POST", "7", eingabeIDName("7", "neu-POST"))
			erwarteStatus(t, "POST über Tombstone", a, err, 201)
			soll := erwartetNach("POST", 2, "7")
			soll.Version = v + 1
			if got := u.zustand()["7"]; !gleichEintrag(got, soll) {
				t.Errorf("Version %d Known %v: %+v, erwartet %+v", v, k, got, soll)
			}
		}
	}
	if n != 9 {
		t.Errorf("Menge %d, erwartet 9", n)
	}
}

// wertFall ist ein Wert eines Typs mit seinem Rohtext im Payload.
type wertFall struct {
	edm  string
	art  model.WertArt
	text string
	roh  string
}

// TestPayloadKanonform (W29, M12): Typ 8 × Wert 4 (typisch, kleinster, größter, null) = 32 und Edm.String mit sechs
// Sonderformen = 38; der Orakel-Rohtext ist von Hand geschrieben, der Dienst liest den Payload zu denselben Werten zurück.
func TestPayloadKanonform(t *testing.T) {
	q := func(s string) string { return `"` + s + `"` }
	f := []wertFall{
		{"Edm.String", model.ArtString, "abc", q("abc")}, {"Edm.String", model.ArtString, "", q("")}, {"Edm.String", model.ArtString, strings.Repeat("z", 300), q(strings.Repeat("z", 300))},
		{"Edm.Boolean", model.ArtBoolean, "true", "true"}, {"Edm.Boolean", model.ArtBoolean, "false", "false"}, {"Edm.Boolean", model.ArtBoolean, "true", "true"},
		{"Edm.Int32", model.ArtInt32, "42", "42"}, {"Edm.Int32", model.ArtInt32, "-2147483648", "-2147483648"}, {"Edm.Int32", model.ArtInt32, "2147483647", "2147483647"},
		{"Edm.Int64", model.ArtInt64, "42", q("42")}, {"Edm.Int64", model.ArtInt64, "-9223372036854775808", q("-9223372036854775808")}, {"Edm.Int64", model.ArtInt64, "9223372036854775807", q("9223372036854775807")},
		{"Edm.Decimal", model.ArtDecimal, "1.5", q("1.5")}, {"Edm.Decimal", model.ArtDecimal, "-79228162514264337593543950335", q("-79228162514264337593543950335")}, {"Edm.Decimal", model.ArtDecimal, "79228162514264337593543950335", q("79228162514264337593543950335")},
		{"Edm.Guid", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e", q("0f8fad5b-d9cb-469f-a165-70867728950e")}, {"Edm.Guid", model.ArtGUID, "00000000-0000-0000-0000-000000000000", q("00000000-0000-0000-0000-000000000000")}, {"Edm.Guid", model.ArtGUID, "ffffffff-ffff-ffff-ffff-ffffffffffff", q("ffffffff-ffff-ffff-ffff-ffffffffffff")},
		{"Edm.Date", model.ArtDate, "2020-01-01", q("2020-01-01")}, {"Edm.Date", model.ArtDate, "0001-01-01", q("0001-01-01")}, {"Edm.Date", model.ArtDate, "9999-12-31", q("9999-12-31")},
		{"Edm.DateTimeOffset", model.ArtDateTimeOffset, "2020-01-01T00:00:00Z", q("2020-01-01T00:00:00Z")}, {"Edm.DateTimeOffset", model.ArtDateTimeOffset, "0001-01-01T00:00:00Z", q("0001-01-01T00:00:00Z")}, {"Edm.DateTimeOffset", model.ArtDateTimeOffset, "9999-12-31T23:59:59.9999999Z", q("9999-12-31T23:59:59.9999999Z")},
	}
	for _, edm := range model.UnterstuetzteEDMTypen() {
		f = append(f, wertFall{edm, model.ArtNull, "", "null"})
	}
	// Sonderformen von Edm.String, aus Bytewerten (AGENTS.md §3.18): kein Backslash-u steht in dieser Datei.
	bs := "\\"
	cyr := string([]byte{0xD0, 0xB5})
	sonder := []wertFall{
		{"Edm.String", model.ArtString, "a\"b", `"a` + bs + `"b"`}, {"Edm.String", model.ArtString, "a" + bs + "b", `"a` + bs + bs + `b"`},
		{"Edm.String", model.ArtString, "a\x00b", `"a` + bs + "u0000" + `b"`}, {"Edm.String", model.ArtString, "a\x1fb", `"a` + bs + "u001f" + `b"`},
		{"Edm.String", model.ArtString, "a\x7fb", "\"a\x7fb\""}, {"Edm.String", model.ArtString, "a" + cyr + "b", `"a` + cyr + `b"`},
	}
	f = append(f, sonder...)
	if len(f) != 38 {
		t.Fatalf("Menge %d, erwartet 38", len(f))
	}
	for i, w := range f {
		prop := propJeTyp(w.edm)
		for _, m := range []string{"POST", "PUT"} {
			u := schreibUmg(t)
			id := schluesselFuer(m)
			e := eingabeIDName(id, "neu")
			e.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, id), pw(prop, w.art, w.text)}
			a, err := u.mutiere(m, id, e)
			erwarteStatus(t, fmt.Sprintf("%s %d", m, i), a, err, sollStatus(m))
			want := erwPayload(map[string]string{"Id": id, prop: w.roh})
			if got := u.zustand()[id].Payload; got != want {
				t.Errorf("%s Fall %d (%s): Payload %q, erwartet %q", m, i, prop, got, want)
			}
			// Rückleseprobe über den Lesepfad: dieselben Werte.
			u.entity(model.Optionen{}, id)
			if _, err := u.leseGET(); err != nil {
				t.Errorf("%s Fall %d: GET %v", m, i, err)
				continue
			}
			ev := u.p.von("Schreiber.SchreibeEntity")
			var gelesen model.Wert
			for _, pv := range ev[len(ev)-1].args[2].(model.EntityAntwort).Entity.Properties {
				if pv.Name == prop {
					gelesen = pv.Wert
				}
			}
			if gelesen.Art != w.art || gelesen.Text != w.text {
				t.Errorf("%s Fall %d (%s): zurückgelesen %+v, erwartet %v %q", m, i, prop, gelesen, w.art, w.text)
			}
		}
	}
}

// gefuellt ist ein Eintrag von Products, dessen jede Property gesetzt ist (Vorher-Werte, verschieden von allen Werten der Mutationen).
func gefuellt(id string) model.EntityState {
	e := voll(int(id[0]-'0'), "vorher")
	e.Payload = erwPayload(map[string]string{"Big": `"11"`, "Count": "12", "Day": `"2001-02-03"`, "Flag": "true", "Id": id, "Name": `"vorher"`,
		"Price": `"13.5"`, "Ref": `"1b2c3d4e-0000-0000-0000-000000000001"`, "Ts": `"2001-02-03T04:05:06Z"`})
	return e
}

// neueWerte sind die Werte der Mutation je Property (Name, Flag, Count, Big, Price, Ref, Day, Ts: die Folge des Modells).
func neueWerte() []model.PropertyWert {
	return []model.PropertyWert{
		pw("Name", model.ArtString, "neu"), pw("Flag", model.ArtBoolean, "false"), pw("Count", model.ArtInt32, "99"), pw("Big", model.ArtInt64, "98"),
		pw("Price", model.ArtDecimal, "9.25"), pw("Ref", model.ArtGUID, "aaaaaaaa-0000-0000-0000-00000000000f"), pw("Day", model.ArtDate, "2020-05-06"),
		pw("Ts", model.ArtDateTimeOffset, "2020-05-06T07:08:09Z"),
	}
}

// neueRoh sind die Rohtexte der Werte aus neueWerte.
func neueRoh() map[string]string {
	return map[string]string{"Name": `"neu"`, "Flag": "false", "Count": "99", "Big": `"98"`, "Price": `"9.25"`, "Ref": `"aaaaaaaa-0000-0000-0000-00000000000f"`,
		"Day": `"2020-05-06"`, "Ts": `"2020-05-06T07:08:09Z"`}
}

// TestNichtGenannteProperties (W31, W36, M14): Position der nicht genannten Properties {erste, mittlere, letzte, alle nullbaren,
// keine} × Methode {POST, PUT} = 10; nicht genannte nullbare Properties werden null, auch wenn sie vorher gesetzt waren, alle
// neun sind bekannt.
func TestNichtGenannteProperties(t *testing.T) {
	alle := neueWerte()
	faelle := map[string][]string{"erste": {"Name"}, "mittlere": {"Price"}, "letzte": {"Ts"}, "alle nullbaren": {"Name", "Flag", "Count", "Big", "Price", "Ref", "Day", "Ts"}, "keine": nil}
	n := 0
	for fname, weg := range faelle {
		for _, m := range []string{"POST", "PUT"} {
			n++
			u := schreibUmg(t)
			u.en.setze(setProducts, gefuellt("7")) // PUT ersetzt ihn; POST legt 8 an
			id := "7"
			if m == "POST" {
				id = "8"
			}
			e := model.EntityEingabe{Typ: typProduct, Schluessel: id, Properties: []model.PropertyWert{pw("Id", model.ArtInt32, id)}}
			werte := map[string]string{"Id": id}
			for _, p := range alle {
				if !slices.Contains(weg, p.Name) {
					e.Properties = append(e.Properties, p)
					werte[p.Name] = neueRoh()[p.Name]
				}
			}
			a, err := u.mutiere(m, id, e)
			erwarteStatus(t, fname+" "+m, a, err, sollStatus(m))
			got := u.zustand()[id]
			if got.Payload != erwPayload(werte) || !slices.Equal(got.KnownProperties, folgeNamen()) || got.Completeness != model.EntityComplete || got.Deleted {
				t.Errorf("%s %s: %+v", fname, m, got)
			}
			// GET: 200, die nicht genannte Property ist null (W32).
			u.entity(model.Optionen{}, id)
			if _, err := u.leseGET(); err != nil {
				t.Errorf("%s %s: GET nach dem Schreiben: %v", fname, m, err)
			}
		}
	}
	if n != 10 {
		t.Errorf("Menge %d, erwartet 10", n)
	}
}

// esc schreibt die Escape-Folge für ein Zeichen unter U+0020 (AGENTS.md §3.18: kein Backslash-u steht in dieser Datei).
func esc(hex4 string) string { return "\\" + "u" + hex4 }

// TestPatchNurGenannte (W33, M15): genannte {1, 2, alle} × Position {erste, mittlere, letzte} bei fünf Properties = 7 Fälle;
// die nicht genannten behalten ihren Rohtext byte-gleich, auch nicht kanonisch geschrieben.
func TestPatchNurGenannte(t *testing.T) {
	fuenf := []string{"Name", "Flag", "Count", "Price", "Ts"}
	roh := map[string]string{"Name": `"a` + esc("0041") + `"`, "Flag": "true", "Count": "100", "Price": `"1.50"`, "Ts": `"2001-02-03T04:05:06Z"`}
	faelle := [][]string{{"Name"}, {"Count"}, {"Ts"}, {"Name", "Count"}, {"Name", "Ts"}, {"Count", "Ts"}, fuenf}
	for _, genannt := range faelle {
		u := schreibUmg(t)
		alt := voll(2, "vorher")
		alt.Payload = erwPayload(map[string]string{"Id": "2", "Name": roh["Name"], "Flag": roh["Flag"], "Count": roh["Count"], "Price": roh["Price"], "Ts": roh["Ts"]})
		u.en.setze(setProducts, alt)
		e := model.EntityEingabe{Typ: typProduct}
		soll := map[string]string{"Id": "2"}
		for k, v := range roh {
			soll[k] = v
		}
		for _, p := range neueWerte() {
			if slices.Contains(genannt, p.Name) {
				e.Properties = append(e.Properties, p)
				soll[p.Name] = neueRoh()[p.Name]
			}
		}
		a, err := u.mutiere("PATCH", "2", e)
		erwarteStatus(t, fmt.Sprint(genannt), a, err, 204)
		got := u.zustand()["2"]
		if got.Payload != erwPayload(soll) || got.Version != 2 || got.Completeness != model.EntityComplete || !slices.Equal(got.KnownProperties, folgeNamen()) {
			t.Errorf("%v: %+v", genannt, got)
		}
	}
	if len(faelle) != 7 {
		t.Errorf("Menge %d, erwartet 7", len(faelle))
	}
}

// TestPatchVierZustaende (W34, M16): Zustand {Wert, null, leerer Text, nicht genannt} × Position {S1, S2, S3} = 12; die vier
// Zustände bleiben unterscheidbar, der Vorher-Wert ist von allen Nachher-Werten verschieden.
func TestPatchVierZustaende(t *testing.T) {
	n := 0
	for _, zustand := range []string{"Wert", "null", "leerer Text", "nicht genannt"} {
		for _, pos := range []string{"S1", "S2", "S3"} {
			n++
			u := schreibUmg(t)
			alt := model.EntityState{
				WorkspaceID: wsID, EntitySet: "Drei", EntityKey: "1", Completeness: model.EntityComplete, Version: 1,
				Payload: `{"Id":1,"S1":"v1","S2":"v2","S3":"v3"}`, KnownProperties: []string{"Id", "S1", "S2", "S3"},
			}
			u.en.setze("Drei", alt)
			u.ressource(model.RessourceEntity, "Drei", model.Optionen{}, zahl("", "1"))
			e := model.EntityEingabe{Typ: "N.Drei"}
			soll := map[string]string{"S1": `"v1"`, "S2": `"v2"`, "S3": `"v3"`}
			switch zustand {
			case "Wert":
				e.Properties, soll[pos] = []model.PropertyWert{pw(pos, model.ArtString, "w")}, `"w"`
			case "null":
				e.Properties, soll[pos] = []model.PropertyWert{pw(pos, model.ArtNull, "")}, "null"
			case "leerer Text":
				e.Properties, soll[pos] = []model.PropertyWert{pw(pos, model.ArtString, "")}, `""`
			}
			u.ei.ergebnis = e
			a, err := u.rufe(anfrage("PATCH", dienstRoot+"/x"))
			erwarteStatus(t, zustand+" "+pos, a, err, 204)
			want := fmt.Sprintf(`{"Id":1,"S1":%s,"S2":%s,"S3":%s}`, soll["S1"], soll["S2"], soll["S3"])
			var got model.EntityState
			for _, x := range u.en.inhalt()["Drei"] {
				got = x
			}
			if got.Payload != want || got.Version != 2 {
				t.Errorf("%s %s: Payload %q, erwartet %q", zustand, pos, got.Payload, want)
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge %d, erwartet 12", n)
	}
	// Die Vereinigung der bekannten Properties wächst um eine genannte, die der Eintrag nicht kannte.
	u := schreibUmg(t)
	alt := voll(2, "b")
	alt.Payload, alt.KnownProperties = `{"Id":2,"Name":"b"}`, []string{"Id", "Name"}
	u.en.setze(setProducts, alt)
	a, err := u.mutiere("PATCH", "2", model.EntityEingabe{Typ: typProduct, Properties: []model.PropertyWert{pw("Ts", model.ArtDateTimeOffset, "2020-05-06T07:08:09Z")}})
	erwarteStatus(t, "PATCH neue Property", a, err, 204)
	if got := u.zustand()["2"]; !slices.Equal(got.KnownProperties, []string{"Id", "Name", "Ts"}) || got.Payload != `{"Id":2,"Name":"b","Ts":"2020-05-06T07:08:09Z"}` {
		t.Errorf("neue Property: %+v", got)
	}
}

// TestPatchLeer (W35): PATCH mit `{}` ist gültig: 204, Payload byte-gleich, Version plus 1, Revision plus 1.
func TestPatchLeer(t *testing.T) {
	u := schreibUmg(t)
	alt := voll(2, "b")
	alt.Payload = ` { "Id" : 2 , "Name" : "b" } ` // nicht kanonisch: bleibt byte-gleich
	u.en.setze(setProducts, alt)
	a, err := u.mutiere("PATCH", "2", model.EntityEingabe{Typ: typProduct, Properties: []model.PropertyWert{}})
	erwarteStatus(t, "PATCH {}", a, err, 204)
	got := u.zustand()["2"]
	if got.Payload != alt.Payload || got.Version != 2 || u.p.anzahl("Entities.Put") != 1 || u.p.anzahl("Workspaces.Advance") != 1 || u.ws.revision(wsID) != 1 {
		t.Errorf("PATCH {}: %+v Put %d Advance %d", got, u.p.anzahl("Entities.Put"), u.p.anzahl("Workspaces.Advance"))
	}
}

// TestPutPartial (W37): PUT auf partial ersetzt, danach complete mit allen Properties bekannt; Known alt 3.
func TestPutPartial(t *testing.T) {
	for _, known := range [][]string{{"Id"}, {"Id", "Name", "Flag", "Count"}, {"Big", "Count", "Day", "Flag", "Id", "Name", "Price", "Ref"}} {
		u := schreibUmg(t)
		alt := teilEintrag("2", 2)
		alt.KnownProperties = known
		u.en.setze(setProducts, alt)
		a, err := u.mutiere("PUT", "2", eingabeIDName("2", "neu"))
		erwarteStatus(t, "PUT partial", a, err, 204)
		soll := erwartetNach("PUT", 1, "2")
		soll.Payload = erwPayload(map[string]string{"Id": "2", "Name": `"neu"`})
		if got := u.zustand()["2"]; !gleichEintrag(got, soll) {
			t.Errorf("Known %v: %+v, erwartet %+v", known, got, soll)
		}
	}
}

// TestFolgen (W43, M17): sechs Folgen von Mutationen, Zustand, Version und Revision je Schritt von Hand.
func TestFolgen(t *testing.T) {
	type schritt struct {
		m         string
		code      model.Code
		version   int64 // 0: der Eintrag fehlt oder ist Tombstone ohne Prüfung
		geloescht bool
	}
	folgen := [][]schritt{
		{{"POST", 0, 1, false}, {"DELETE", 0, 2, true}, {"POST", 0, 3, false}},
		{{"POST", 0, 1, false}, {"PATCH", 0, 2, false}, {"PATCH", 0, 3, false}},
		{{"POST", 0, 1, false}, {"PUT", 0, 2, false}, {"PUT", 0, 3, false}},
		{{"POST", 0, 1, false}, {"DELETE", 0, 2, true}, {"DELETE", e6008, 2, true}},
		{{"PUT", 0, 3, false}, {"DELETE", 0, 4, true}},
		{{"PATCH", 0, 2, false}, {"DELETE", 0, 3, true}, {"POST", 0, 4, false}},
	}
	for fi, folge := range folgen {
		u := schreibUmg(t)
		switch fi {
		case 4:
			alt := teilEintrag("7", 2)
			u.en.setze(setProducts, alt)
		case 5:
			u.en.setze(setProducts, voll(7, "alt"))
		}
		rev := int64(0)
		for si, s := range folge {
			a, err := u.mutiere(s.m, "7", eingabeIDName("7", fmt.Sprintf("f%d%d", fi, si)))
			kontext := fmt.Sprintf("Folge %d Schritt %d %s", fi, si, s.m)
			if s.code != 0 {
				pruefeFehler(t, kontext, err, s.code)
			} else {
				erwarteStatus(t, kontext, a, err, sollStatus(s.m))
				rev++
			}
			got := u.zustand()["7"]
			if got.Version != s.version || got.Deleted != s.geloescht || u.ws.revision(wsID) != rev {
				t.Errorf("%s: Version %d (erwartet %d) Deleted %v (erwartet %v) Revision %d (erwartet %d)", kontext, got.Version, s.version, got.Deleted, s.geloescht, u.ws.revision(wsID), rev)
			}
		}
	}
}

// TestVersionUndRevision (W44, W45, M18): Methode 4 × Version alt 3 (0, 1, 7) = 12 und Methode 4 × Revision 3 (0, 5, 2 hoch 40) = 12;
// die Fixture-Versionen sind verschieden vom Ergebnis, Advance erhält die in der Transaktion gelesene Revision.
func TestVersionUndRevision(t *testing.T) {
	n := 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		for _, v := range []int64{0, 1, 7} {
			n++
			u := schreibUmg(t)
			alt := voll(2, "b")
			alt.Version = v
			id := "2"
			if m == "POST" {
				alt.Deleted, id = true, "2" // POST über einen Tombstone: Version alt plus 1
			}
			u.en.setze(setProducts, alt)
			a, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
			erwarteStatus(t, m, a, err, sollStatus(m))
			if got := u.zustand()["2"].Version; got != v+1 {
				t.Errorf("%s Version alt %d: %d, erwartet %d", m, v, got, v+1)
			}
		}
		for _, r := range []int64{0, 5, 1 << 40} {
			n++
			u := schreibUmg(t)
			u.ws.setze(model.Workspace{ID: wsID, ServiceID: dienstID, BaselineSnapshotID: baselineID, Revision: r})
			id := "2"
			if m == "POST" {
				id = "7"
			}
			a, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
			erwarteStatus(t, m, a, err, sollStatus(m))
			adv := u.p.von("Workspaces.Advance")
			if len(adv) != 1 || adv[0].args[0] != wsID || adv[0].args[1] != r || u.ws.revision(wsID) != r+1 {
				t.Errorf("%s Revision %d: Advance %v, Revision danach %d", m, r, adv, u.ws.revision(wsID))
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}

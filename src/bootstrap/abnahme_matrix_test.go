package bootstrap_test

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// abFehlerIst ist das Ergebnis eines Schritts, der mit Fehler endet: Status und Code, Antwort der Nullwert.
func abFehlerIst(status int, code model.Code) abIst {
	return abIst{F: &inbound.Fehler{Code: code, Status: status}}
}

// Golden-Fixtures der Muss-Zeilen der Abnahmematrix (SPEC-049 Nr. 4): 36 Adressen, 92 Schritte, gegen den Mock mit echten Adaptern.

// abSoll ist die erwartete Antwort eines Schritts: Status, Header (Namen kanonisch), Body; bei einem Fehler Status und Code.
type abSoll struct {
	Status int
	Kopf   map[string][]string
	Body   string
	Code   model.Code
}

type abSchritt struct {
	Name   string
	Strikt bool
	Anf    abAnfrage
	Soll   abSoll
}

type abFixture struct {
	ID       string
	Schritte []abSchritt
	Revision int64
}

const abCTMinimal = "application/json;odata.metadata=minimal;charset=utf-8"

func abKopf(ct, coverage, loc string) map[string][]string {
	h := map[string][]string{"Odata-Version": {"4.0"}}
	if ct != "" {
		h["Content-Type"] = []string{ct}
	}
	if coverage != "" {
		h["X-Recorder-Coverage"] = []string{coverage}
	}
	if loc != "" {
		h["Location"] = []string{abU(loc)}
	}
	return h
}

func abKK(body string) abSoll {
	return abSoll{Status: 200, Kopf: abKopf(abCTMinimal, "complete", ""), Body: abU(body)}
}
func abKE(body string) abSoll {
	return abSoll{Status: 200, Kopf: abKopf(abCTMinimal, "", ""), Body: abU(body)}
}
func abKZ(body string) abSoll {
	return abSoll{Status: 200, Kopf: abKopf("text/plain;charset=utf-8", "complete", ""), Body: body}
}
func abKL(loc, body string) abSoll {
	return abSoll{Status: 201, Kopf: abKopf(abCTMinimal, "", loc), Body: abU(body)}
}
func abK0() abSoll                           { return abSoll{Status: 204, Kopf: abKopf("", "", "")} }
func abF(status int, code model.Code) abSoll { return abSoll{Status: status, Code: code} }

func abG(pfad, query string) abAnfrage { return abAnfrage{Methode: "GET", Pfad: pfad, Query: query} }
func abGH(pfad, query, accept string) abAnfrage {
	return abAnfrage{Methode: "GET", Pfad: pfad, Query: query, Header: map[string][]string{"Accept": {accept}}}
}
func abM(methode, pfad, body string) abAnfrage {
	return abAnfrage{Methode: methode, Pfad: pfad, Body: body}
}

func abS(name string, a abAnfrage, s abSoll) abSchritt { return abSchritt{Name: name, Anf: a, Soll: s} }
func abStrikt(name string, a abAnfrage, s abSoll) abSchritt {
	return abSchritt{Name: name, Strikt: true, Anf: a, Soll: s}
}

func abKZeile(cov string, body string) abSoll {
	s := abKK(body)
	s.Kopf = abKopf(abCTMinimal, cov, "")
	return s
}

// abFixtures ist die Tabelle von SPEC-049 Nr. 4: je Adresse die Schritte und die Revision danach.
func abFixtures() []abFixture {
	P := func(ids ...int) []string {
		var l []string
		for _, n := range ids {
			l = append(l, abEL(n))
		}
		return l
	}
	csdl := abSoll{Status: 200, Kopf: abKopf("application/xml", "", ""), Body: abCSDL}
	n409 := abF(409, model.CodeE6009)
	return []abFixture{
		{ID: "ODV4-GET-01", Schritte: []abSchritt{abS("a", abG("/Products", ""), abKK(abK(P(1, 2, 3, 4, 5)...)))}},
		{ID: "ODV4-GET-02", Schritte: []abSchritt{
			abS("a", abG("/Products(1)", ""), abKE(abEN(1))), abS("b", abG("/Products(4)", ""), abKE(abEN(4))), abS("c", abG("/Products(99)", ""), abF(404, model.CodeE6008))}},
		{ID: "ODV4-GET-03", Schritte: []abSchritt{
			abS("a", abG("/Labels", ""), abKK(`{"@odata.context":"{W}/$metadata#Labels","value":[]}`)), abS("b", abG("/Labels/$count", ""), abKZ("0")),
			abS("c", abG("/Labels(1)", ""), abF(404, model.CodeE6008))}},
		{ID: "ODV4-Q-01", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Name eq 'Anvil'&$select=Id,Name"), abKK(abAuswahl("Id,Name", 1)))}},
		{ID: "ODV4-Q-02", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Name ne 'Anvil'&$select=Id"), abKK(abAuswahl("Id", 2, 3, 4, 5)))}},
		{ID: "ODV4-Q-03", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Price gt 12.5&$select=Id"), abKK(abAuswahl("Id", 3)))}},
		{ID: "ODV4-Q-04", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Price ge 12.5&$select=Id"), abKK(abAuswahl("Id", 1, 3)))}},
		{ID: "ODV4-Q-05", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Price lt 0.25&$select=Id"), abKK(abAuswahl("Id", 4)))}},
		{ID: "ODV4-Q-06", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Price le 0.25&$select=Id"), abKK(abAuswahl("Id", 2, 4)))}},
		{ID: "ODV4-Q-07", Schritte: []abSchritt{
			abS("a", abG("/Products", "$filter=Active eq true and Stock gt 0&$select=Id"), abKK(abAuswahl("Id", 1, 3))),
			abS("b", abG("/Products", "$filter=Id eq 2 or Id eq 5&$select=Id"), abKK(abAuswahl("Id", 2, 5)))}},
		{ID: "ODV4-Q-08", Schritte: []abSchritt{abS("", abG("/Products", "$filter=not (Active eq true)&$select=Id"), abKK(abAuswahl("Id", 2, 5)))}},
		{ID: "ODV4-Q-09", Schritte: []abSchritt{
			abS("a", abG("/Products(1)", "$select=Id"), abKE(`{"@odata.context":"{W}/$metadata#Products(Id)/$entity","@odata.id":"{W}/Products(1)","Id":1}`)),
			abS("b", abG("/Products", "$filter=Id eq 3&$select=Name,Id"), abKK(abAuswahl("Name,Id", 3)))}},
		{ID: "ODV4-Q-10", Schritte: []abSchritt{abS("a", abG("/Products", "$select=Id,Name&$orderby=Name desc,Id"), abKK(abAuswahl("Id,Name", 5, 3, 2, 1, 4))),
			abS("b", abG("/Products", "$select=Id&$orderby=Active,Id desc"), abKK(abAuswahl("Id", 5, 2, 4, 3, 1)))}},
		{ID: "ODV4-Q-11", Schritte: []abSchritt{abS("", abG("/Products", "$orderby=Id desc&$top=2&$skip=1&$select=Id"), abKK(abAuswahl("Id", 4, 3)))}},
		{ID: "ODV4-Q-12", Schritte: []abSchritt{
			abS("a", abG("/Products", "$count=true&$top=0"), abKK(`{"@odata.context":"{W}/$metadata#Products","@odata.count":5,"value":[]}`)),
			abS("b", abG("/Products/$count", "$filter=Active eq true"), abKZ("3"))}},
		{ID: "ODV4-T-01", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Name eq 'Chain'&$select=Id,Name"), abKK(abAuswahl("Id,Name", 3)))}},
		{ID: "ODV4-T-02", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Active eq false&$select=Id,Active"), abKK(abAuswahl("Id,Active", 2)))}},
		{ID: "ODV4-T-03", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Id ge 4&$select=Id"), abKK(abAuswahl("Id", 4, 5)))}},
		{ID: "ODV4-T-04", Schritte: []abSchritt{
			abS("a", abG("/Products", "$filter=Stock eq 9007199254740993&$select=Id,Stock"), abKK(abAuswahl("Id,Stock", 1))),
			func() abSchritt {
				s := abS("b", abGH("/Products", "$filter=Stock eq 9007199254740993&$select=Id,Stock", "application/json;IEEE754Compatible=true"),
					abKK(strings.Replace(abAuswahl("Id,Stock", 1), `"Stock":9007199254740993`, `"Stock":"9007199254740993"`, 1)))
				s.Soll.Kopf["Content-Type"] = []string{"application/json;odata.metadata=minimal;IEEE754Compatible=true;charset=utf-8"}
				return s
			}()}},
		{ID: "ODV4-T-05", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Price eq -3.75&$select=Id,Price"), abKK(abAuswahl("Id,Price", 4)))}},
		{ID: "ODV4-T-06", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Ref eq 0000000c-0000-0000-0000-00000000000c&$select=Id,Ref"), abKK(abAuswahl("Id,Ref", 3)))}},
		{ID: "ODV4-T-07", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Released eq 2024-02-29&$select=Id,Released"), abKK(abAuswahl("Id,Released", 4)))}},
		{ID: "ODV4-T-08", Schritte: []abSchritt{abS("", abG("/Products", "$filter=Updated eq 2025-12-31T23:59:59.5Z&$select=Id,Updated"), abKK(abAuswahl("Id,Updated", 2)))}},
		{ID: "ODV4-H-01", Schritte: []abSchritt{
			func() abSchritt {
				s := abS("a", abGH("/Products(1)", "$select=Id", "application/json;odata.metadata=none"), abKE(`{"Id":1}`))
				s.Soll.Kopf["Content-Type"] = []string{"application/json;odata.metadata=none;charset=utf-8"}
				return s
			}(),
			abS("b", abGH("/Products(1)", "$select=Id", "text/html"), abKE(`{"@odata.context":"{W}/$metadata#Products(Id)/$entity","@odata.id":"{W}/Products(1)","Id":1}`)),
			abS("c", abGH("/Products(1)", "$select=Id", "application/json;odata.metadata=full"), abKE(`{"@odata.context":"{W}/$metadata#Products(Id)/$entity","@odata.id":"{W}/Products(1)","Id":1}`))}},
		{ID: "ODV4-H-02", Revision: 3, Schritte: []abSchritt{
			abS("a", abG("/Products", "$top=1"), abKK(abK(P(1)...))), abS("b", abG("/Products(1)", ""), abKE(abEN(1))), abS("c", abG("/Products/$count", ""), abKZ("5")),
			abS("d", abG("/$metadata", ""), csdl),
			abS("e", abM("POST", "/Products", `{"Id":11}`), abKL("{W}/Products(11)", `{"@odata.context":"{W}/$metadata#Products/$entity","@odata.id":"{W}/Products(11)","Id":11,`+
				`"Name":null,"Active":null,"Stock":null,"Price":null,"Ref":null,"Released":null,"Updated":null}`)),
			abS("f", abM("PATCH", "/Products(11)", `{"Name":"p"}`), abK0()), abS("g", abM("DELETE", "/Products(11)", ""), abK0())}},
		{ID: "ODV4-H-03", Schritte: []abSchritt{
			abS("a", abG("/$metadata", ""), csdl), abS("b", abG("/Products/$count", ""), abKZ("5")), abS("c", abG("/Products", ""), abKK(abK(P(1, 2, 3, 4, 5)...)))}},
		{ID: "ODV4-E-01", Schritte: []abSchritt{
			abS("a", abG("/Products", "$filter=Nope eq 1"), abF(400, model.CodeE6005)), abS("b", abG("/Products", "$top=-1"), abF(400, model.CodeE6005))}},
		{ID: "ODV4-E-02", Schritte: []abSchritt{abS("a", abG("/Unbekannt", ""), abF(404, model.CodeE6008)), abS("b", abG("/Products(99)", ""), abF(404, model.CodeE6008))}},
		{ID: "ODV4-E-03", Schritte: []abSchritt{
			abS("a", abG("/Products", "$search=x"), abF(501, model.CodeE6001)), abS("b", abG("/Products", "$apply=groupby((Name))"), abF(501, model.CodeE6001)),
			abS("c", abG("/Products", "$expand=Foo"), abF(501, model.CodeE6001)), abS("d", abG("", ""), abF(501, model.CodeE6001)),
			abS("e", abAnfrage{Methode: "OPTIONS", Pfad: "/Products"}, abF(501, model.CodeE6002))}},
		{ID: "ODV4-E-04", Schritte: []abSchritt{
			abStrikt("a", abG("/Orders", ""), abF(422, model.CodeE6003)),
			abS("b", abG("/Orders", ""), abKZeile("partial", `{"@odata.context":"{W}/$metadata#Orders","value":[{"@odata.id":"{W}/Orders(1)","Id":1}]}`)),
			abStrikt("c", abG("/Orders/$count", ""), abF(422, model.CodeE6003)),
			abStrikt("d", abG("/Products", "$top=1&$select=Id"), abKK(abAuswahl("Id", 1))),
			abS("e", abG("/Orders(9)", ""), abF(422, model.CodeE6003)), abStrikt("f", abG("/Orders(9)", ""), abF(422, model.CodeE6003)),
			abS("g", abG("/Orders(1)", ""), abF(422, model.CodeE6003)),
			abStrikt("h", abM("PATCH", "/Orders(9)", `{"Note":"x"}`), abF(422, model.CodeE6003)),
			abStrikt("i", abM("PATCH", "/Products(99)", `{"Name":"x"}`), abF(404, model.CodeE6008))}},
		{ID: "ODV4-CRUD-01", Revision: 1, Schritte: []abSchritt{
			abS("a", abM("POST", "/Products", `{"Id":10,"Name":"Eyelet","Active":true,"Stock":"12","Price":"1.5","Ref":"0000000a-0000-0000-0000-00000000000a","Released":"2026-07-01","Updated":"2026-07-01T09:00:00Z"}`),
				abKL("{W}/Products(10)", abCrud10())),
			abS("b", abG("/Products(10)", ""), abKE(abCrud10())), abS("c", abG("/Products/$count", ""), abKZ("6")),
			abS("d", abG("/Products", "$select=Id&$orderby=Id desc&$top=2"), abKK(abAuswahl("Id", 10, 5)))}},
		{ID: "ODV4-CRUD-02", Revision: 1, Schritte: []abSchritt{
			abS("a", abM("PATCH", "/Products(1)", `{"Name":"Anvil2","Stock":"10"}`), abK0()),
			abS("b", abG("/Products(1)", ""), abKE(abENmit(1, map[string]string{"Name": `"Anvil2"`, "Stock": "10"})))}},
		{ID: "ODV4-CRUD-03", Revision: 2, Schritte: []abSchritt{
			abS("a", abM("PUT", "/Products(2)", `{"Id":2,"Name":"Bolt2","Active":true}`), abK0()),
			abS("b", abG("/Products(2)", ""), abKE(abU(`{"@odata.context":"{W}/$metadata#Products/$entity","@odata.id":"{W}/Products(2)","Id":2,"Name":"Bolt2","Active":true,`+
				`"Stock":null,"Price":null,"Ref":null,"Released":null,"Updated":null}`))),
			abS("c", abM("PUT", "/Orders(1)", `{"Id":1,"Note":"n"}`), abK0()),
			abS("d", abG("/Orders(1)", ""), abKE(`{"@odata.context":"{W}/$metadata#Orders/$entity","@odata.id":"{W}/Orders(1)","Id":1,"Note":"n"}`))}},
		{ID: "ODV4-CRUD-04", Revision: 2, Schritte: []abSchritt{
			abS("a", abM("DELETE", "/Products(3)", ""), abK0()), abS("b", abG("/Products(3)", ""), abF(404, model.CodeE6008)), abS("c", abG("/Products/$count", ""), abKZ("4")),
			abS("d", abG("/Products", "$select=Id"), abKK(abAuswahl("Id", 1, 2, 4, 5))), abS("e", abM("DELETE", "/Products(3)", ""), abF(404, model.CodeE6008)),
			abS("f", abM("POST", "/Products", `{"Id":3,"Name":"Chain2"}`), abKL("{W}/Products(3)", `{"@odata.context":"{W}/$metadata#Products/$entity","@odata.id":"{W}/Products(3)",`+
				`"Id":3,"Name":"Chain2","Active":null,"Stock":null,"Price":null,"Ref":null,"Released":null,"Updated":null}`))}},
		{ID: "ODV4-CRUD-05", Revision: 3, Schritte: []abSchritt{
			abS("a", abM("PATCH", "/Products(1)", `{"Name":null}`), abK0()),
			abS("b", abG("/Products(1)", "$select=Id,Name,Active"), abKE(`{"@odata.context":"{W}/$metadata#Products(Id,Name,Active)/$entity","@odata.id":"{W}/Products(1)","Id":1,"Name":null,"Active":true}`)),
			abS("c", abM("PATCH", "/Products(2)", `{"Name":""}`), abK0()),
			abS("d", abG("/Products(2)", "$select=Id,Name,Active"), abKE(`{"@odata.context":"{W}/$metadata#Products(Id,Name,Active)/$entity","@odata.id":"{W}/Products(2)","Id":2,"Name":"","Active":false}`)),
			abS("e", abM("PATCH", "/Products(3)", `{}`), abK0()),
			abS("f", abG("/Products(3)", "$select=Id,Name,Active"), abKE(`{"@odata.context":"{W}/$metadata#Products(Id,Name,Active)/$entity","@odata.id":"{W}/Products(3)","Id":3,"Name":"Chain","Active":true}`))}},
		{ID: "ODV4-CRUD-06", Schritte: []abSchritt{
			abS("a", abM("POST", "/Products", `{"Id":1,"Name":"Dup"}`), n409), abS("b", abM("POST", "/Products", `{"Name":"NoKey"}`), abF(400, model.CodeE6007)),
			abS("c", abM("PATCH", "/Products(1)", `{"Id":2}`), abF(400, model.CodeE6007)), abS("d", abM("PATCH", "/Products(99)", `{"Name":"x"}`), abF(404, model.CodeE6008)),
			abS("e", abM("PATCH", "/Orders(9)", `{"Note":"x"}`), abF(422, model.CodeE6003)), abS("f", abM("PATCH", "/Orders(1)", `{"Note":"x"}`), abF(422, model.CodeE6003)),
			abS("g", abG("/Products/$count", ""), abKZ("5"))}},
	}
}

func abCrud10() string {
	return `{"@odata.context":"{W}/$metadata#Products/$entity","@odata.id":"{W}/Products(10)","Id":10,"Name":"Eyelet","Active":true,"Stock":12,"Price":1.5,` +
		`"Ref":"0000000a-0000-0000-0000-00000000000a","Released":"2026-07-01","Updated":"2026-07-01T09:00:00Z"}`
}

// abErwarteteIDs sind die 36 Adressen aus der Tabelle von SPEC-049, erzeugt aus den Bereichen und ihren Zahlen (nicht aus abFixtures).
func abErwarteteIDs() []string {
	var ids []string
	for _, b := range []struct {
		name string
		n    int
	}{{"GET", 3}, {"CRUD", 6}, {"Q", 12}, {"T", 8}, {"H", 3}, {"E", 4}} {
		for i := 1; i <= b.n; i++ {
			ids = append(ids, fmt.Sprintf("ODV4-%s-%02d", b.name, i))
		}
	}
	return ids
}

// abSchrittAus liefert den Schritt name der Fixture id (D-Folge der Determinismus-Messung, SPEC-049 Nr. 5 b).
func abSchrittAus(t *testing.T, id, name string) abSchritt {
	t.Helper()
	for _, f := range abFixtures() {
		if f.ID != id {
			continue
		}
		for _, s := range f.Schritte {
			if s.Name == name || (name == "" && len(f.Schritte) == 1) {
				return s
			}
		}
	}
	t.Fatalf("Schritt %s %s nicht in der Tabelle", id, name)
	return abSchritt{}
}

// abVergleiche nennt die Unterschiede von ist gegen soll (SPEC-049 Nr. 3 c): Status, die Menge der Header-Namen und jeder Wert genau, der Body
// byte-gleich; bei einem Fehler Status und Code, Antwort der Nullwert, Text von Error() allein Code und Meldung.
func abVergleiche(ist abIst, soll abSoll) []string {
	var d []string
	if soll.Code != 0 {
		if ist.F == nil {
			return append(d, fmt.Sprintf("kein Fehler mit Code (Status %d, Fehler %v), erwartet %s", ist.Status, ist.Err, soll.Code))
		}
		if ist.F.Status != soll.Status {
			d = append(d, fmt.Sprintf("Status %d, erwartet %d", ist.F.Status, soll.Status))
		}
		if ist.F.Code != soll.Code {
			d = append(d, fmt.Sprintf("Code %s, erwartet %s", ist.F.Code, soll.Code))
		}
		if ist.Antwort.Status != 0 || ist.Antwort.Header != nil || ist.Antwort.Body != nil {
			d = append(d, "die Antwort bei einem Fehler ist nicht der Nullwert")
		}
		if want := ist.F.Code.String() + ": " + ist.F.Code.Message(); ist.F.Error() != want {
			d = append(d, fmt.Sprintf("Error() %q, erwartet %q", ist.F.Error(), want))
		}
		return d
	}
	if ist.F != nil || ist.Err != nil {
		return append(d, fmt.Sprintf("Fehler %v, erwartet Status %d", ist.Err, soll.Status))
	}
	if ist.Status != soll.Status {
		d = append(d, fmt.Sprintf("Status %d, erwartet %d", ist.Status, soll.Status))
	}
	if !abKopfGleich(ist.Header, soll.Kopf) {
		d = append(d, fmt.Sprintf("Header %v, erwartet %v", ist.Header, soll.Kopf))
	}
	if ist.Body != soll.Body {
		d = append(d, fmt.Sprintf("Body %q, erwartet %q", ist.Body, soll.Body))
	}
	return d
}

// abKopfGleich vergleicht zwei Header-Mengen: gleiche Namen (kanonisch), je Name gleiche Werte in gleicher Folge.
func abKopfGleich(a, b map[string][]string) bool {
	na, nb := map[string][]string{}, map[string][]string{}
	for k, v := range a {
		na[http.CanonicalHeaderKey(k)] = v
	}
	for k, v := range b {
		nb[http.CanonicalHeaderKey(k)] = v
	}
	return reflect.DeepEqual(na, nb)
}

// abPruefeZaehlung nennt die Verstöße gegen SPEC-049 Nr. 3 e: genau die erwarteten IDs, jede einmal, jede mit den Schritten, die ihre Tabelle
// nennt, und jeder Schritt gelaufen.
func abPruefeZaehlung(fixtures []abFixture, gelaufen map[string]int) []string {
	var d []string
	var ids []string
	schritte := 0
	gesehen := map[string]int{}
	for _, f := range fixtures {
		gesehen[f.ID]++
		ids = append(ids, f.ID)
		schritte += len(f.Schritte)
		if gelaufen != nil && gelaufen[f.ID] != len(f.Schritte) {
			d = append(d, fmt.Sprintf("%s: %d Schritte gelaufen, Tabelle %d", f.ID, gelaufen[f.ID], len(f.Schritte)))
		}
	}
	for id, n := range gesehen {
		if n != 1 {
			d = append(d, fmt.Sprintf("%s steht %d mal in der Tabelle", id, n))
		}
	}
	sort.Strings(ids)
	soll := abErwarteteIDs()
	sort.Strings(soll)
	if len(fixtures) != 36 || !reflect.DeepEqual(uniq(ids), soll) {
		d = append(d, fmt.Sprintf("%d IDs in der Tabelle, erwartet genau die 36 aus SPEC-049", len(fixtures)))
	}
	if schritte != 92 {
		d = append(d, fmt.Sprintf("%d Schritte in der Tabelle, erwartet 92", schritte))
	}
	return d
}

func uniq(l []string) []string {
	var r []string
	for i, s := range l {
		if i == 0 || l[i-1] != s {
			r = append(r, s)
		}
	}
	return r
}

// abFixtureLaufen führt die Schritte einer Fixture in einer frischen Umgebung aus und prüft Revision und Ausgangszustand.
func abFixtureLaufen(t *testing.T, f abFixture) int {
	t.Helper()
	u := abFrisch(t)
	abPruefeAusgang(t, u)
	n := 0
	for _, s := range f.Schritte {
		var ist abIst
		if s.Strikt {
			ist = u.sendeStrikt(s.Anf)
		} else {
			ist = u.sende(s.Anf)
		}
		n++
		for _, d := range abVergleiche(ist, s.Soll) {
			t.Errorf("%s %s %s %s: %s", f.ID, s.Name, s.Anf.Methode, s.Anf.Pfad, d)
		}
	}
	if w, err := u.db.Workspaces().Get(context.Background(), "ws-1"); err != nil || w.Revision != f.Revision {
		t.Errorf("%s: Revision %d (%v), erwartet %d", f.ID, w.Revision, err, f.Revision)
	}
	if f.Revision == 0 {
		for _, d := range abStandGleich(u.stand("ws-1"), abAusgangsstand("ws-1", 0)) {
			t.Errorf("%s: Zustand nach der ID, erwartet der Ausgangszustand: %s", f.ID, d)
		}
	}
	return n
}

// abPruefeAusgang prüft, dass u im Ausgangszustand von SPEC-049 Nr. 2 steht (SPEC-049 Nr. 3 b).
func abPruefeAusgang(t *testing.T, u *abUmgebung) {
	t.Helper()
	for _, d := range abStandGleich(u.stand("ws-1"), abAusgangsstand("ws-1", 0)) {
		t.Errorf("Ausgangszustand: %s", d)
	}
}

// TestAbnahmeMatrix (T01 bis T36, T42): je Adresse ein Teilfall, jeder Schritt gegen die erwartete Antwort.
func TestAbnahmeMatrix(t *testing.T) {
	gelaufen := map[string]int{}
	fixtures := abFixtures()
	for _, f := range fixtures {
		t.Run(f.ID, func(t *testing.T) { gelaufen[f.ID] = abFixtureLaufen(t, f) })
	}
	for _, d := range abPruefeZaehlung(fixtures, gelaufen) {
		t.Error(d)
	}
}

// TestAbnahmeMatrixZaehlung (T37): die Tabelle führt genau die 36 IDs mit 92 Schritten; eine fehlende ID, ein fehlender Schritt und eine doppelte ID
// sind Verstöße, ein nicht gelaufener Schritt ebenso. Die 92 sind die Summe der Bereiche (GET 7, CRUD 29, Q 16, T 9, H 13, E 18).
func TestAbnahmeMatrixZaehlung(t *testing.T) {
	fixtures := abFixtures()
	gelaufen := map[string]int{}
	je := map[string]int{}
	for _, f := range fixtures {
		gelaufen[f.ID] = len(f.Schritte)
		je[strings.Split(f.ID, "-")[1]] += len(f.Schritte)
	}
	if d := abPruefeZaehlung(fixtures, gelaufen); len(d) != 0 {
		t.Errorf("Tabelle: %v", d)
	}
	for bereich, soll := range map[string]int{"GET": 7, "CRUD": 29, "Q": 16, "T": 9, "H": 13, "E": 18} {
		if je[bereich] != soll {
			t.Errorf("Bereich %s: %d Schritte, erwartet %d", bereich, je[bereich], soll)
		}
	}
	// M3: die acht Typen der Zeile EDM, je einer in ODV4-T-01 bis 08 (die Eigenschaft, die der Schritt wählt, hat den Typ aus dem CSDL).
	modell, _, err := csdl.New().ParseMetadata(context.Background(), strings.NewReader(abCSDL), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	produkt, _ := modell.EntityType("N.Product")
	typen := map[string]int{}
	var tFixtures int
	for _, f := range fixtures {
		if !strings.HasPrefix(f.ID, "ODV4-T-") {
			continue
		}
		tFixtures++
		var ersteQuery string
		for _, s := range f.Schritte[:1] {
			ersteQuery = s.Anf.Query
		}
		_, sel, _ := strings.Cut(ersteQuery, "$select=")
		namen := strings.Split(sel, ",")
		name := namen[len(namen)-1]
		for _, p := range produkt.Properties {
			if p.Name == name {
				typen[p.EdmType]++
			}
		}
	}
	soll := model.UnterstuetzteEDMTypen()
	if len(soll) != 8 || tFixtures != len(soll) || len(typen) != len(soll) {
		t.Errorf("Typen: %d unterstützt, %d Fixtures ODV4-T, %d verschiedene Typen gewählt; erwartet 8, 8, 8", len(soll), tFixtures, len(typen))
	}
	for _, typ := range soll {
		if typen[typ] != 1 {
			t.Errorf("Typ %s wird von %d Fixtures ODV4-T gewählt, erwartet genau eine", typ, typen[typ])
		}
	}
	// Gegenproben des Zählers: eine ID weniger, ein Schritt weniger, eine ID doppelt, ein Schritt nicht gelaufen.
	ohneID := fixtures[1:]
	ohneSchritt := append([]abFixture(nil), fixtures...)
	ohneSchritt[0].Schritte = nil
	doppelt := append(append([]abFixture(nil), fixtures...), fixtures[0])
	wenigerGelaufen := map[string]int{}
	for k, v := range gelaufen {
		wenigerGelaufen[k] = v
	}
	wenigerGelaufen["ODV4-CRUD-05"]--
	for name, fall := range map[string]struct {
		f []abFixture
		g map[string]int
	}{"ID fehlt": {ohneID, gelaufen}, "Schritt fehlt": {ohneSchritt, gelaufen}, "ID doppelt": {doppelt, gelaufen}, "Schritt nicht gelaufen": {fixtures, wenigerGelaufen}} {
		if len(abPruefeZaehlung(fall.f, fall.g)) == 0 {
			t.Errorf("%s: kein Verstoß gemeldet", name)
		}
	}
}

// TestAbnahmeMatrixVergleich (T38 bis T40): der Vergleich nennt die Menge der Header genau, den Body byte-gleich und bei einem Fehler Status und Code.
func TestAbnahmeMatrixVergleich(t *testing.T) {
	soll := abKK(`{"a":1}`)
	gleich := abIst{Status: 200, Header: map[string][]string{"Content-Type": soll.Kopf["Content-Type"], "Odata-Version": {"4.0"}, "X-Recorder-Coverage": {"complete"}}, Body: soll.Body}
	if d := abVergleiche(gleich, soll); len(d) != 0 {
		t.Fatalf("gleiche Antwort: %v", d)
	}
	t.Run("Header", func(t *testing.T) {
		mehr := gleich
		mehr.Header = map[string][]string{"Content-Type": soll.Kopf["Content-Type"], "Odata-Version": {"4.0"}, "X-Recorder-Coverage": {"complete"}, "Date": {"x"}}
		weniger := gleich
		weniger.Header = map[string][]string{"Content-Type": soll.Kopf["Content-Type"], "Odata-Version": {"4.0"}}
		anders := gleich
		anders.Header = map[string][]string{"Content-Type": soll.Kopf["Content-Type"], "Odata-Version": {"4.0"}, "X-Recorder-Coverage": {"partial"}}
		doppelt := gleich
		doppelt.Header = map[string][]string{"Content-Type": soll.Kopf["Content-Type"], "Odata-Version": {"4.0", "4.0"}, "X-Recorder-Coverage": {"complete"}}
		for name, ist := range map[string]abIst{"weiterer Header": mehr, "Header fehlt": weniger, "anderer Wert": anders, "Wert doppelt": doppelt} {
			if len(abVergleiche(ist, soll)) == 0 {
				t.Errorf("%s: kein Unterschied gemeldet", name)
			}
		}
	})
	t.Run("Body", func(t *testing.T) {
		for name, b := range map[string]string{"Zeilenende": soll.Body + "\n", "Leerzeichen": soll.Body + " ", "gekürzt": soll.Body[:len(soll.Body)-1], "anders": `{"a":2}`} {
			ist := gleich
			ist.Body = b
			if len(abVergleiche(ist, soll)) == 0 {
				t.Errorf("%s: kein Unterschied gemeldet", name)
			}
		}
	})
	t.Run("Fehler", func(t *testing.T) {
		fehler := abF(404, model.CodeE6008)
		ist := abFehlerIst(404, model.CodeE6008)
		if d := abVergleiche(ist, fehler); len(d) != 0 {
			t.Fatalf("gleicher Fehler: %v", d)
		}
		for name, i := range map[string]abIst{"anderer Status": abFehlerIst(400, model.CodeE6008), "anderer Code": abFehlerIst(404, model.CodeE6003)} {
			if len(abVergleiche(i, fehler)) == 0 {
				t.Errorf("%s: kein Unterschied gemeldet", name)
			}
		}
		if len(abVergleiche(gleich, fehler)) == 0 {
			t.Error("eine Antwort statt eines Fehlers: kein Unterschied gemeldet")
		}
		if len(abVergleiche(ist, soll)) == 0 {
			t.Error("ein Fehler statt einer Antwort: kein Unterschied gemeldet")
		}
	})
}

// TestAbnahmeMatrixAusgangszustand (T41): jede ID beginnt im Ausgangszustand; nach einer mutierenden ID steht die nächste Umgebung wieder darin und
// teilt weder Datei noch Ablage mit der vorigen.
func TestAbnahmeMatrixAusgangszustand(t *testing.T) {
	var vorige *abUmgebung
	for _, id := range []string{"ODV4-CRUD-04", "ODV4-CRUD-05", "ODV4-GET-01"} {
		for _, f := range abFixtures() {
			if f.ID != id {
				continue
			}
			u := abFrisch(t)
			abPruefeAusgang(t, u)
			if vorige != nil && (u.dbPfad == vorige.dbPfad || u.blobPfad == vorige.blobPfad || u.db == vorige.db) {
				t.Errorf("%s teilt Datei oder Ablage mit der vorigen ID", id)
			}
			for _, s := range f.Schritte {
				u.sende(s.Anf)
			}
			vorige = u
		}
	}
}

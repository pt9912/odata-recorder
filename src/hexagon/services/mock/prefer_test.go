package mock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Das Orakel der Antworten mit Wahl von `Prefer` (SPEC-286 Nr. 1, SPEC-288 Nr. 3) steht hier als Tabelle aus dem Vertrag; es teilt keinen Code mit dem Dienst.

// preferWahl ist eine der drei Wahlen mit einer Zeile, die sie ergibt, und dem Wert von `Preference-Applied`.
type preferWahl struct {
	name    string
	zeilen  []string
	applied string
}

func preferWahlen() []preferWahl {
	return []preferWahl{
		{"Keine", nil, ""},
		{"Minimal", []string{"return=minimal"}, "return=minimal"},
		{"Entity", []string{"return=representation"}, "return=representation"},
	}
}

// preferWahlNamens ist die Wahl mit dem Namen name.
func preferWahlNamens(name string) preferWahl {
	return preferWahlen()[slices.IndexFunc(preferWahlen(), func(w preferWahl) bool { return w.name == name })]
}

// preferZelle ist eine Zeile der Tabelle von SPEC-286 Nr. 1.
type preferZelle struct {
	status                                int
	body, contentType, location, entityID bool
}

func preferTabelle(methode, wahl string) preferZelle {
	switch {
	case methode == "DELETE":
		return preferZelle{status: 204}
	case methode == "POST" && wahl == "Minimal":
		return preferZelle{status: 204, location: true, entityID: true}
	case methode == "POST":
		return preferZelle{status: 201, body: true, contentType: true, location: true}
	case wahl == "Minimal":
		return preferZelle{status: 204, entityID: true}
	case wahl == "Entity":
		return preferZelle{status: 200, body: true, contentType: true}
	}
	return preferZelle{status: 204}
}

// preferAufrufe sind die Aufrufe von S24 SchreibeEntity, S25 Adresse und S26 SchreibeAnzahl je Antwort (SPEC-288 Nr. 3).
func preferAufrufe(methode, wahl string) [3]int {
	switch {
	case methode == "DELETE":
		return [3]int{0, 0, 1}
	case methode == "POST" && wahl == "Minimal", methode != "POST" && wahl == "Minimal":
		return [3]int{0, 1, 1}
	case methode == "POST":
		return [3]int{1, 1, 0}
	case wahl == "Entity":
		return [3]int{1, 0, 0}
	}
	return [3]int{0, 0, 1}
}

// preferOps nennt je Aufrufstelle des Schreibers den Namen der Operation im Protokoll.
func preferOps() [3]string {
	return [3]string{"Schreiber.SchreibeEntity", "Schreiber.Adresse", "Schreiber.SchreibeAnzahl"}
}

func preferUmg(t *testing.T, etag bool) *umg { return etagUmg(t, etag, 1, 1, 1) }

// preferErgebnis ist, was ein Aufruf ergab: Antwort, Fehler und der gelesene Body.
type preferErgebnis struct {
	ant  inbound.Antwort
	err  error
	body []byte
}

// preferSende ruft die Methode auf den Eintrag id; kopf trägt weitere Header der Anfrage, bau ändert den Aufbau nach der Ressource.
// POST, PATCH, PUT und DELETE tragen einen Body und eine Ressource der Mutation, GET und HEAD die Entity. Eine Antwort ohne Fehler mit `Body` nil
// ist ein Fehler des Aufrufs, für jede Methode und jede Wahl (SPEC-286 Nr. 1, SPEC-194 Nr. 3).
func (u *umg) preferSende(methode, id string, kopf map[string][]string, bau func(u *umg, m string)) preferErgebnis {
	u.t.Helper()
	var a inbound.Anfrage
	if methode == "GET" || methode == "HEAD" {
		u.entity(model.Optionen{}, id)
		a = anfrage(methode, dienstRoot+"/x")
	} else {
		u.setzeMutation(methode, id, eingabeIDName(id, "neu"))
		a = anfrage(methode, dienstRoot+"/x")
		a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
		a.Header["Content-Type"] = []string{"application/json"}
	}
	for k, v := range kopf {
		a.Header[k] = v
	}
	if bau != nil {
		bau(u, methode)
	}
	ant, err := u.rufe(a)
	r := preferErgebnis{ant: ant, err: err}
	if err == nil && ant.Body == nil {
		u.t.Errorf("%s %v: Body nil, der leere Strom ist der Body der Antwort ohne Entity (SPEC-286 Nr. 1)", methode, kopf)
	}
	if err == nil && ant.Body != nil {
		r.body = lies(u.t, ant)
	}
	return r
}

func preferKopf(zeilen []string) map[string][]string {
	if zeilen == nil {
		return nil
	}
	return map[string][]string{"Prefer": zeilen}
}

// preferElement ist ein Element eines Headers `Prefer` und was es für die Wahl bedeutet (SPEC-285 Nr. 3): entscheidet es, mit welchem Wert von
// `Preference-Applied` (leer: Keine).
type preferElement struct {
	text        string
	entscheidet bool
	applied     string
}

func preferElemente() []preferElement {
	return []preferElement{
		{"return=minimal", true, "return=minimal"}, {"return=representation", true, "return=representation"}, {"return=foo", true, ""},
		{"return", true, ""}, {"return=", true, ""}, {"respond-async", false, ""}, {"odata.x=1", false, ""}, {"return@x", false, ""},
		{"", false, ""}, {"RETURN=MINIMAL", true, "return=minimal"}, {"Return = Representation; x=y", true, "return=representation"},
	}
}

// preferForm ist ein Wert des Headers als Zeilen und die Wahl, die ihn liest: der erste entscheidende Wert.
type preferForm struct {
	zeilen  []string
	applied string
}

// preferFormen liefert jede Folge der Länge 0 bis n aus den Elementen, in jeder Verteilung auf Kopfzeilen (je Lücke dieselbe Zeile mit `,` oder die
// nächste Zeile): bei n = 2 sind es 1 + 11 + 121 · 2 = 254 Formen.
func preferFormen(n int) []preferForm {
	preferElemente := preferElemente()
	var r []preferForm
	var rek func(f []int)
	rek = func(f []int) {
		applied := ""
		for _, i := range f {
			if preferElemente[i].entscheidet {
				applied = preferElemente[i].applied
				break
			}
		}
		for v := 0; v < 1<<max(len(f)-1, 0); v++ {
			var zeilen []string
			for k, i := range f {
				if k > 0 && v&(1<<(k-1)) == 0 {
					zeilen[len(zeilen)-1] += "," + preferElemente[i].text
				} else {
					zeilen = append(zeilen, preferElemente[i].text)
				}
			}
			r = append(r, preferForm{zeilen, applied})
		}
		if len(f) < n {
			for i := range preferElemente {
				rek(append(slices.Clone(f), i))
			}
		}
	}
	rek(nil)
	return r
}

// preferAbweichend sind Formen, die einen Wert tragen, den kein Fehlertext nennen darf (SPEC-287 Nr. 2).
func preferAbweichend() []preferForm {
	return []preferForm{{[]string{"return=minimal; k=" + kanarienwert}, "return=minimal"}, {[]string{kanarienwert + "=1"}, ""},
		{[]string{"return=" + kanarienwert}, ""}, {[]string{"return=representation", kanarienwert}, "return=representation"}}
}

// TestPreferAntwortTabelle (Z10, SPEC-286 Nr. 1): die Antwort je Methode und Wahl trägt Status, Body und genau die Header der Tabelle, nie `Vary`:
// Methode 3 × Wahl 3 × Etag 2 × Accept 2 = 36; die Menge der Header-Namen ist gleich der erwarteten, nicht darin enthalten.
func TestPreferAntwortTabelle(t *testing.T) {
	accepts := [][]string{{"application/json;odata.metadata=none"}, {"application/json;odata.metadata=minimal"}}
	n := 0
	for _, m := range dreiMethoden() {
		for _, w := range preferWahlen() {
			for _, etag := range []bool{false, true} {
				for _, acc := range accepts {
					n++
					u := preferUmg(t, etag)
					id := schluesselFuer(m)
					kopf := map[string][]string{"Accept": acc}
					if w.zeilen != nil {
						kopf["Prefer"] = w.zeilen
					}
					r := u.preferSende(m, id, kopf, nil)
					z := preferTabelle(m, w.name)
					kontext := fmt.Sprintf("%s %s Etag %v %v", m, w.name, etag, acc)
					if r.err != nil || r.ant.Status != z.status {
						t.Errorf("%s: Status %d Fehler %v, erwartet %d", kontext, r.ant.Status, r.err, z.status)
						continue
					}
					soll := map[string][]string{"X-Fake-Kennzeichen": {"kennzeichen-1"}}
					if z.contentType {
						soll["Content-Type"] = []string{"text/plain"}
					}
					if z.location {
						soll["Location"] = []string{"http://adresse.test/fake"}
					}
					if z.entityID {
						soll["OData-EntityId"] = []string{"http://adresse.test/fake"}
					}
					if w.applied != "" {
						soll["Preference-Applied"] = []string{w.applied}
					}
					if etag {
						v := int64(2)
						if m == "POST" {
							v = 1
						}
						soll["ETag"] = []string{etagLit(v)}
					}
					if !reflect.DeepEqual(r.ant.Header, soll) {
						t.Errorf("%s: Header %v, erwartet genau %v", kontext, r.ant.Header, soll)
					}
					if z.body != strings.HasPrefix(string(r.body), "Schreiber.SchreibeEntity:PROBE-") || !z.body && len(r.body) != 0 {
						t.Errorf("%s: Body %q, erwartet mit Entity: %v", kontext, r.body, z.body)
					}
				}
			}
		}
	}
	if n != 36 {
		t.Errorf("Menge %d, erwartet 36", n)
	}
}

// TestPreferAdresseNieAusHost (Z11, SPEC-286 Nr. 1): `Location` und `OData-EntityId` tragen dieselbe Zeichenfolge aus Adresse, nie aus einem Header der
// Anfrage; Adresse erhält Wurzel, Set und den Schlüssel des geschriebenen Eintrags. Methode 3 × Wahl 3 = 9, mit fremden Host-Headern.
func TestPreferAdresseNieAusHost(t *testing.T) {
	fremd := map[string][]string{"Host": {"fremd.test"}, "X-Forwarded-Host": {"fremd2.test"}, "Forwarded": {"host=fremd3.test"}, "X-Forwarded-Proto": {"gopher"}}
	n := 0
	for _, m := range dreiMethoden() {
		for _, w := range preferWahlen() {
			n++
			u := preferUmg(t, false)
			u.c.sch.ort = "http://ort.test/Products(7)?x=1"
			kopf := map[string][]string{"Prefer": w.zeilen}
			maps2(kopf, fremd)
			r := u.preferSende(m, schluesselFuer(m), kopf, nil)
			kontext := m + " " + w.name
			if r.err != nil {
				t.Errorf("%s: %v", kontext, r.err)
				continue
			}
			z := preferTabelle(m, w.name)
			for name, da := range map[string]bool{"Location": z.location, "OData-EntityId": z.entityID} {
				if got := r.ant.Header[name]; da && !slices.Equal(got, []string{"http://ort.test/Products(7)?x=1"}) || !da && got != nil {
					t.Errorf("%s: %s %v, erwartet %v", kontext, name, got, da)
				}
			}
			ev := u.p.von("Schreiber.Adresse")
			if want := preferAufrufe(m, w.name)[1]; len(ev) != want {
				t.Errorf("%s: %d Aufrufe von Adresse, erwartet %d", kontext, len(ev), want)
				continue
			}
			for _, e := range ev {
				if e.args[1] != "http://recorder.test/svc" || e.args[2] != setProducts || e.args[3] != schluesselFuer(m) {
					t.Errorf("%s: Adresse %v", kontext, e.args)
				}
			}
			for name, werte := range r.ant.Header {
				if strings.Contains(strings.Join(werte, ","), "fremd") {
					t.Errorf("%s: %s trägt einen Host der Anfrage: %v", kontext, name, werte)
				}
			}
		}
	}
	if n != 9 {
		t.Errorf("Menge %d, erwartet 9", n)
	}
}

func maps2(ziel, quelle map[string][]string) {
	for k, v := range quelle {
		ziel[k] = v
	}
}

// preferAppliedWerte sind alle Werte der Kopfzeilen `Preference-Applied` einer Antwort, jede Schreibweise des Namens.
func preferAppliedWerte(a inbound.Antwort) []string {
	var r []string
	for n, w := range a.Header {
		if strings.EqualFold(n, "preference-applied") {
			r = append(r, w...)
		}
	}
	return r
}

// TestPreferApplied (Z12, SPEC-286 Nr. 1, 3): `Preference-Applied` ist genau `return=minimal` oder `return=representation`, in einer Kopfzeile, ohne Echo des
// Client-Werts und ohne weitere Präferenz, für jede Form (254) × Methode 3.
func TestPreferApplied(t *testing.T) {
	formen := preferFormen(2)
	if len(formen) != 254 {
		t.Fatalf("%d Formen, erwartet 254", len(formen))
	}
	n := 0
	for _, m := range dreiMethoden() {
		for _, f := range formen {
			n++
			u := preferUmg(t, false)
			r := u.preferSende(m, schluesselFuer(m), preferKopf(f.zeilen), nil)
			if r.err != nil {
				t.Errorf("%s %q: %v", m, f.zeilen, r.err)
				continue
			}
			want := []string(nil)
			if f.applied != "" {
				want = []string{f.applied}
			}
			if got := preferAppliedWerte(r.ant); !slices.Equal(got, want) {
				t.Errorf("%s %q: Preference-Applied %v, erwartet %v", m, f.zeilen, got, want)
			}
			if _, da := r.ant.Header["Preference-Applied"]; da != (want != nil) {
				t.Errorf("%s %q: Name der Kopfzeile in anderer Schreibweise: %v", m, f.zeilen, r.ant.Header)
			}
		}
	}
	if n != 3*254 {
		t.Errorf("Menge %d", n)
	}
}

// preferSchreibweisen sind die 64 Schreibweisen des Namens `Prefer`: jeder Buchstabe klein oder groß.
func preferSchreibweisen() []string {
	wort := "prefer"
	var r []string
	for m := 0; m < 1<<len(wort); m++ {
		b := []byte(wort)
		for i := range b {
			if m&(1<<i) != 0 {
				b[i] -= 'a' - 'A'
			}
		}
		r = append(r, string(b))
	}
	return r
}

// TestPreferNamenUndReihenfolge (Z5, SPEC-285 Nr. 1): alle Kopfzeilen `Prefer` jeder Schreibweise des Namens werden gelesen, in der Bytefolge der Namen,
// dann in ihrer Reihenfolge: jedes geordnete Paar der 64 Schreibweisen mit verschiedenen Werten (4032) und je Teilmenge der drei Namen `PREFER`, `Prefer`,
// `prefer` jede Wahl der Zeilen aus fünf (124 mit mindestens einem Namen).
func TestPreferNamenUndReihenfolge(t *testing.T) {
	namen := preferSchreibweisen()
	n := 0
	for _, a := range namen {
		for _, b := range namen {
			if a == b {
				continue
			}
			n++
			u := preferUmg(t, false)
			r := u.preferSende("PATCH", "2", map[string][]string{a: {"return=minimal"}, b: {"return=representation"}}, nil)
			want := "return=representation"
			if a < b {
				want = "return=minimal"
			}
			if got := preferAppliedWerte(r.ant); r.err != nil || !slices.Equal(got, []string{want}) {
				t.Errorf("%q und %q: %v, Fehler %v, erwartet %s (die kleinere Bytefolge zuerst)", a, b, got, r.err, want)
			}
		}
	}
	if n != 64*63 {
		t.Errorf("Menge der Paare %d", n)
	}
	optionen := []struct {
		zeilen  []string
		applied string
	}{{nil, "-"}, {[]string{"return=minimal"}, "return=minimal"}, {[]string{"return=representation"}, "return=representation"},
		{[]string{"return=foo"}, ""}, {[]string{"respond-async", "return=representation"}, "return=representation"}}
	drei := []string{"PREFER", "Prefer", "prefer"}
	m := 0
	for k := 0; k < 125; k++ {
		wahl := [3]int{k % 5, k / 5 % 5, k / 25}
		kopf := map[string][]string{}
		applied := "-"
		for i, name := range drei {
			o := optionen[wahl[i]]
			if o.zeilen == nil {
				continue
			}
			kopf[name] = o.zeilen
			if applied == "-" {
				applied = o.applied
			}
		}
		if applied == "-" {
			continue
		}
		m++
		u := preferUmg(t, false)
		r := u.preferSende("PUT", "2", kopf, nil)
		want := []string(nil)
		if applied != "" {
			want = []string{applied}
		}
		if got := preferAppliedWerte(r.ant); r.err != nil || !slices.Equal(got, want) {
			t.Errorf("%v: %v, Fehler %v, erwartet %v", kopf, got, r.err, want)
		}
	}
	if m != 124 {
		t.Errorf("Menge der Teilmengen %d, erwartet 124", m)
	}
}

// preferFehlerFall richtet einen Fehler einer Mutation ein, der nicht an der Wahl hängt.
type preferFehlerFall struct {
	name     string
	code     model.Code
	methoden []string
	id       string
	kopf     map[string][]string
	bau      func(u *umg, m string)
}

func preferFehlerFaelle() []preferFehlerFall {
	alle := []string{"POST", "PATCH", "PUT", "DELETE"}
	ohnePost := []string{"PATCH", "PUT", "DELETE"}
	return []preferFehlerFall{
		{"E6002 Filter", e6002, alle, "", nil, func(u *umg, m string) {
			if m == "POST" {
				u.ressource(model.RessourceMenge, setProducts, model.Optionen{Filter: idGroesser("0")})
				return
			}
			u.ressource(model.RessourceEntity, setProducts, model.Optionen{Filter: idGroesser("0")}, zahl("", "2"))
		}},
		{"E6003 partial", e6003, ohnePost, "99", nil, func(u *umg, _ string) { u.co.setze(baselineID, setProducts, model.CoveragePartial) }},
		{"E6005 Schlüssel", e6005, ohnePost, "", nil, func(u *umg, _ string) {
			u.ressource(model.RessourceEntity, setProducts, model.Optionen{})
		}},
		{"E6007 Body", e6007, []string{"POST", "PATCH", "PUT"}, "", nil, func(u *umg, _ string) {
			u.ei.f = func(context.Context, io.Reader, []string, string, model.EingabeArt) (model.EntityEingabe, error) {
				return model.EntityEingabe{}, model.ErrEingabeUngueltig
			}
		}},
		{"E6008 fehlt", e6008, ohnePost, "99", nil, nil},
		{"E6009 vorhanden", e6009, []string{"POST"}, "2", nil, nil},
		{"E6010 Vorbedingung", model.CodeE6010, ohnePost, "", map[string][]string{"If-Match": {`W/"999"`}}, nil},
		{"E9001 Schreiber", e9001, alle, "", nil, func(u *umg, _ string) {
			for _, op := range preferOps() {
				u.p.setzeFehler(op, model.ErrUngueltig)
			}
		}},
		{"E9001 Advance", e9001, alle, "", nil, func(u *umg, _ string) { u.p.setzeFehler("Workspaces.Advance", outbound.ErrRevisionVeraltet) }},
	}
}

func (f preferFehlerFall) idFuer(m string) string {
	if f.id != "" {
		return f.id
	}
	return schluesselFuer(m)
}

// TestPreferAppliedNurBeiErfolg (Z13, SPEC-286 Nr. 3): `Preference-Applied` steht nur an der Antwort einer gelungenen Mutation mit Wahl, an keiner Antwort
// mit Fehler, an keinem GET und DELETE. Methode 5 × Wahl 3 × {Erfolg, jeder Fehlerfall der Methode}.
func TestPreferAppliedNurBeiErfolg(t *testing.T) {
	n, soll := 0, 0
	for _, f := range preferFehlerFaelle() {
		soll += len(f.methoden) * 3
	}
	soll += 5*3 + 3 // Erfolg je Methode, GET mit unbekannter Entity
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE", "GET"} {
		for _, w := range preferWahlen() {
			n++
			u := preferUmg(t, true)
			r := u.preferSende(m, schluesselFuer(m), preferKopf(w.zeilen), nil)
			da := m != "DELETE" && m != "GET" && w.applied != ""
			if r.err != nil || len(preferAppliedWerte(r.ant)) > 0 != da {
				t.Errorf("%s %s Erfolg: Fehler %v, Preference-Applied %v, erwartet %v", m, w.name, r.err, preferAppliedWerte(r.ant), da)
			}
			if m == "GET" {
				n++
				u := preferUmg(t, true)
				r := u.preferSende(m, "99", preferKopf(w.zeilen), nil)
				pruefeFehler(t, "GET unbekannt "+w.name, r.err, e6008)
				if len(preferAppliedWerte(r.ant)) != 0 {
					t.Errorf("GET unbekannt %s: Preference-Applied %v", w.name, preferAppliedWerte(r.ant))
				}
			}
		}
	}
	for _, f := range preferFehlerFaelle() {
		for _, m := range f.methoden {
			for _, w := range preferWahlen() {
				n++
				u := preferUmg(t, true)
				r := u.preferSende(m, f.idFuer(m), maps3(preferKopf(w.zeilen), f.kopf), f.bau)
				pruefeFehler(t, fmt.Sprintf("%s %s %s", f.name, m, w.name), r.err, f.code)
				if len(preferAppliedWerte(r.ant)) != 0 || r.ant.Header != nil || r.ant.Body != nil {
					t.Errorf("%s %s %s: Antwort %+v neben dem Fehler", f.name, m, w.name, r.ant)
				}
			}
		}
	}
	if n != soll {
		t.Errorf("Menge %d, erwartet %d", n, soll)
	}
}

func maps3(a, b map[string][]string) map[string][]string {
	r := map[string][]string{}
	maps2(r, a)
	maps2(r, b)
	return r
}

// TestPreferFehlerGleich (Z14, Z17, SPEC-287 Nr. 1, 2): eine gescheiterte Mutation antwortet mit jeder Form von `Prefer` (254 und 4 mit einem Wert, den kein Text
// nennen darf) wie ohne: derselbe Fehler, dieselbe Folge der Aufrufe, keine Antwort, derselbe Zustand und dieselbe Revision, der Wert in keinem Text.
// Das trägt zugleich Z17: ein Header, der vor Schritt (l) oder bei einem Fehler gelesen und abgewiesen würde, änderte einen dieser Fälle.
func TestPreferFehlerGleich(t *testing.T) {
	formen := append(preferFormen(2), preferAbweichend()...)
	n, soll := 0, 0
	for _, f := range preferFehlerFaelle() {
		for _, m := range f.methoden {
			if m == "DELETE" {
				continue
			}
			soll += len(formen)
			u0 := preferUmg(t, true)
			r0 := u0.preferSende(m, f.idFuer(m), f.kopf, f.bau)
			pruefeFehler(t, f.name+" "+m+" ohne Prefer", r0.err, f.code)
			for _, form := range formen {
				n++
				u := preferUmg(t, true)
				vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
				r := u.preferSende(m, f.idFuer(m), maps3(preferKopf(form.zeilen), f.kopf), f.bau)
				kontext := fmt.Sprintf("%s %s %q", f.name, m, form.zeilen)
				if !reflect.DeepEqual(r.err, r0.err) {
					t.Errorf("%s: Fehler %v, ohne Prefer %v", kontext, r.err, r0.err)
					continue
				}
				if r.ant.Status != 0 || r.ant.Header != nil || r.ant.Body != nil {
					t.Errorf("%s: Antwort %+v neben dem Fehler", kontext, r.ant)
				}
				if !slices.Equal(preferOhneSchreiber(u.p.ops()), preferOhneSchreiber(u0.p.ops())) {
					t.Errorf("%s: Folge der Aufrufe %v, ohne Prefer %v", kontext, u.p.ops(), u0.p.ops())
				}
				if text := fmt.Sprintf("%v %+v", r.err, r.err); strings.Contains(text, kanarienwert) {
					t.Errorf("%s: der Wert von Prefer steht im Fehler %q", kontext, text)
				}
				u.unveraendert(vorher, rev, kontext)
			}
		}
	}
	if n != soll {
		t.Errorf("Menge %d, erwartet %d", n, soll)
	}
}

// preferOhneSchreiber sind die Operationen eines Protokolls ohne die des Schreibers: sie hängen an der Wahl.
func preferOhneSchreiber(ops []string) []string {
	return slices.DeleteFunc(slices.Clone(ops), func(o string) bool { return strings.HasPrefix(o, "Schreiber.") })
}

// TestPreferZustand (Z15, SPEC-287 Nr. 4): `Prefer` ändert nie den Zustand: mit jeder Form (254 und 4) und Methode 3 × Etag 2 schreibt `Put` denselben Eintrag, die
// Revision und der Zustand sind die ohne Header, und die Ports außer dem Schreiber sehen dieselbe Folge (Transaktion 1, Put 1, Advance 1).
func TestPreferZustand(t *testing.T) {
	formen := append(preferFormen(2), preferAbweichend()...)
	n := 0
	for _, m := range dreiMethoden() {
		for _, etag := range []bool{false, true} {
			u0 := preferUmg(t, etag)
			r0 := u0.preferSende(m, schluesselFuer(m), nil, nil)
			if r0.err != nil {
				t.Fatalf("%s ohne Prefer: %v", m, r0.err)
			}
			for op, want := range map[string]int{"Transaktion.WithWorkspaceTransaction": 1, "Entities.Put": 1, "Workspaces.Advance": 1, "Entities.Get": 1, "Workspaces.Get": 2} {
				if u0.p.anzahl(op) != want {
					t.Fatalf("%s ohne Prefer: %s %d Mal, erwartet %d", m, op, u0.p.anzahl(op), want)
				}
			}
			put0 := u0.p.von("Entities.Put")[0].args[0]
			for _, f := range formen {
				n++
				u := preferUmg(t, etag)
				r := u.preferSende(m, schluesselFuer(m), preferKopf(f.zeilen), nil)
				kontext := fmt.Sprintf("%s Etag %v %q", m, etag, f.zeilen)
				if r.err != nil {
					t.Errorf("%s: %v", kontext, r.err)
					continue
				}
				if put := u.p.von("Entities.Put"); len(put) != 1 || !reflect.DeepEqual(put[0].args[0], put0) {
					t.Errorf("%s: Put %v, ohne Prefer %v", kontext, put, put0)
				}
				if !reflect.DeepEqual(u.zustand(), u0.zustand()) || u.ws.revision(wsID) != u0.ws.revision(wsID) {
					t.Errorf("%s: Zustand oder Revision weichen ab", kontext)
				}
				if !slices.Equal(preferOhneSchreiber(u.p.ops()), preferOhneSchreiber(u0.p.ops())) {
					t.Errorf("%s: Folge der Aufrufe %v, ohne Prefer %v", kontext, u.p.ops(), u0.p.ops())
				}
			}
		}
	}
	if n != 3*2*(254+4) {
		t.Errorf("Menge %d", n)
	}
}

// TestPreferLesendOhneWirkung (Z16, SPEC-287 Nr. 3): GET, HEAD und DELETE ignorieren `Prefer` ohne Wirkung und ohne Fehler: Status, Header, Body und Folge der
// Aufrufe wie ohne Header, für jede Form (254 und 4) × Methode 3; DELETE antwortet bei `return=representation` `204` ohne Body.
func TestPreferLesendOhneWirkung(t *testing.T) {
	formen := append(preferFormen(2), preferAbweichend()...)
	n := 0
	for _, m := range []string{"GET", "HEAD", "DELETE"} {
		u0 := preferUmg(t, true)
		r0 := u0.preferSende(m, "2", nil, nil)
		if r0.err != nil {
			t.Fatalf("%s ohne Prefer: %v", m, r0.err)
		}
		for _, f := range formen {
			n++
			u := preferUmg(t, true)
			r := u.preferSende(m, "2", preferKopf(f.zeilen), nil)
			kontext := fmt.Sprintf("%s %q", m, f.zeilen)
			if r.err != nil || r.ant.Status != r0.ant.Status || !reflect.DeepEqual(r.ant.Header, r0.ant.Header) || !slices.Equal(r.body, r0.body) {
				t.Errorf("%s: %d %v %q Fehler %v, ohne Prefer %d %v %q", kontext, r.ant.Status, r.ant.Header, r.body, r.err, r0.ant.Status, r0.ant.Header, r0.body)
			}
			if !slices.Equal(u.p.ops(), u0.p.ops()) || !reflect.DeepEqual(u.zustand(), u0.zustand()) {
				t.Errorf("%s: Folge der Aufrufe oder Zustand weichen ab", kontext)
			}
		}
	}
	u := preferUmg(t, true)
	if r := u.preferSende("DELETE", "2", preferKopf([]string{"return=representation"}), nil); r.err != nil || r.ant.Status != 204 || len(r.body) != 0 {
		t.Errorf("DELETE mit return=representation: %d %q %v", r.ant.Status, r.body, r.err)
	}
	u = preferUmg(t, true)
	if r := u.preferSende("GET", "2", preferKopf([]string{"return=minimal"}), nil); r.err != nil || r.ant.Status != 200 || len(r.body) == 0 {
		t.Errorf("GET mit return=minimal: %d %q %v", r.ant.Status, r.body, r.err)
	}
	if n != 3*(254+4) {
		t.Errorf("Menge %d", n)
	}
}

// TestPreferSchreiberAufrufe (Z19, SPEC-288 Nr. 3): die Aufrufe des Schreibers je Antwort sind die Zahlen der Tabelle (SchreibeEntity, Adresse,
// SchreibeAnzahl), SchreibeCollection und SchreibeFehler nie; Methode 4 × Wahl 3 (DELETE nur ohne Wirkung) × Etag 2.
func TestPreferSchreiberAufrufe(t *testing.T) {
	n := 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		for _, w := range preferWahlen() {
			for _, etag := range []bool{false, true} {
				n++
				u := preferUmg(t, etag)
				if r := u.preferSende(m, schluesselFuer(m), preferKopf(w.zeilen), nil); r.err != nil {
					t.Fatalf("%s %s: %v", m, w.name, r.err)
				}
				want := preferAufrufe(m, w.name)
				for i, op := range preferOps() {
					if got := u.p.anzahl(op); got != want[i] {
						t.Errorf("%s %s Etag %v: %s %d Mal, erwartet %d", m, w.name, etag, op, got, want[i])
					}
				}
				for _, op := range []string{"Schreiber.SchreibeCollection", "Schreiber.SchreibeFehler"} {
					if u.p.anzahl(op) != 0 {
						t.Errorf("%s %s: %s gerufen", m, w.name, op)
					}
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}

// TestPreferSchreiberArgumente (Z20, SPEC-288 Nr. 2, 3): jeder Aufruf des Schreibers erhält `ctx` unverändert, `Accept` und die Wurzel; `Auswahl` ist `nil`, auch mit `$select`.
// Aufrufstelle S24, S25, S26 in den Antworten, die sie rufen (Methode 3 × Wahl 3) × mit und ohne `$select` × mit und ohne `Accept`.
func TestPreferSchreiberArgumente(t *testing.T) {
	n, gerufen := 0, 0
	for _, m := range dreiMethoden() {
		for _, w := range preferWahlen() {
			for _, sel := range []bool{false, true} {
				for _, acc := range [][]string{nil, {"application/json;odata.metadata=minimal", "text/html"}} {
					n++
					u := preferUmg(t, true)
					kopf := preferKopf(w.zeilen)
					if acc != nil {
						kopf = maps3(kopf, map[string][]string{"Accept": acc})
					}
					bau := func(u *umg, m string) {}
					if sel {
						bau = func(u *umg, m string) {
							opt := model.Optionen{Select: &model.SelectOption{Namen: []string{"Name"}}}
							if m == "POST" {
								u.ressource(model.RessourceMenge, setProducts, opt)
							} else {
								u.ressource(model.RessourceEntity, setProducts, opt, zahl("", "2"))
							}
						}
					}
					if r := u.preferSende(m, schluesselFuer(m), kopf, bau); r.err != nil {
						t.Fatalf("%s %s: %v", m, w.name, r.err)
					}
					kontext := fmt.Sprintf("%s %s $select %v Accept %v", m, w.name, sel, acc)
					for _, e := range u.p.von("Schreiber.SchreibeEntity") {
						gerufen++
						ea := e.args[2].(model.EntityAntwort)
						if e.ctx != u.ctx || ea.Auswahl != nil || !slices.Equal(ea.Accept, acc) || ea.Wurzel != "http://recorder.test/svc" ||
							ea.Set != setProducts || ea.Entity.Schluessel != schluesselFuer(m) || e.args[0] == nil {
							t.Errorf("%s: SchreibeEntity ctx gleich %v, %+v", kontext, e.ctx == u.ctx, ea)
						}
					}
					for _, e := range u.p.von("Schreiber.SchreibeAnzahl") {
						gerufen++
						if e.ctx != u.ctx || e.args[0] == nil || e.args[1].(int64) != 0 {
							t.Errorf("%s: SchreibeAnzahl ctx gleich %v, %v", kontext, e.ctx == u.ctx, e.args)
						}
					}
					for _, e := range u.p.von("Schreiber.Adresse") {
						gerufen++
						if e.args[1] != "http://recorder.test/svc" || e.args[2] != setProducts || e.args[3] != schluesselFuer(m) {
							t.Errorf("%s: Adresse %v", kontext, e.args)
						}
					}
				}
			}
		}
	}
	// POST: Keine 2 Aufrufe, Minimal 2, Entity 2; PATCH und PUT: Keine 1, Minimal 2, Entity 1; je Lauf 4 Läufe
	if n != 36 || gerufen != 4*(2+2+2+2*(1+2+1)) {
		t.Errorf("Menge %d, Aufrufe %d", n, gerufen)
	}
}

// preferStellen sind die Aufrufstellen des Schreibers je Antwort mit der Methode und der Wahl, die sie ruft.
func preferStellen() (r []struct{ m, wahl, op string }) {
	for _, m := range dreiMethoden() {
		for _, w := range preferWahlen() {
			for i, k := range preferAufrufe(m, w.name) {
				if k > 0 {
					r = append(r, struct{ m, wahl, op string }{m, w.name, preferOps()[i]})
				}
			}
		}
	}
	return r
}

// TestPreferSchreiberFehler (Z21, SPEC-288 Nr. 3, SPEC-287 Nr. 5): ein Fehler eines Aufrufs des Schreibers ist `ODR-E9001` mit Rollback, ein Kontextfehler
// (nackt, verpackt, in einem Join) bleibt ohne Code; danach läuft kein weiterer Port, der Zustand bleibt. Der Gegenstand der Abbildung ist je Stelle
// `Sonstiges` (SPEC-195 Nr. 11): `ErrNichtGefunden` und `ErrBereitsVorhanden` sind nur dort `ODR-E9001`, bei jedem der sieben anderen Gegenstände ein anderer Code.
// Stelle 14 × Fehler 7.
func TestPreferSchreiberFehler(t *testing.T) {
	stellen := preferStellen()
	if len(stellen) != 14 {
		t.Fatalf("%d Stellen, erwartet 14", len(stellen))
	}
	fehler := []struct {
		name    string
		err     error
		kontext bool
	}{
		{"ErrUngueltig", model.ErrUngueltig, false}, {"roh", errors.New("roh " + kanarienwert), false}, {"Kontext", context.Canceled, true},
		{"verpackt", fmt.Errorf("Port: %w", context.DeadlineExceeded), true}, {"Join", errors.Join(outbound.ErrNichtGefunden, context.Canceled), true},
		{"ErrNichtGefunden", outbound.ErrNichtGefunden, false}, {"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden, false},
	}
	// Jeder Gegenstand außer `Sonstiges` wird von mindestens einem Fehler der Liste von `Sonstiges` getrennt; die Schleife prüft es je Gegenstand.
	for g := outbound.GegenstandSonstiges + 1; g <= outbound.GegenstandDienst; g++ {
		getrennt := false
		for _, f := range fehler {
			if c, _ := outbound.Abbilden(f.err, g); !f.kontext && c != model.CodeE9001 {
				getrennt = true
			}
		}
		if !getrennt {
			t.Errorf("Gegenstand %d wird von keinem Fehler der Liste von Sonstiges getrennt", g)
		}
	}
	n := 0
	for _, s := range stellen {
		for _, f := range fehler {
			n++
			u := preferUmg(t, true)
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			u.p.setzeFehler(s.op, f.err)
			w := preferWahlNamens(s.wahl)
			r := u.preferSende(s.m, schluesselFuer(s.m), preferKopf(w.zeilen), nil)
			kontext := fmt.Sprintf("%s %s %s %s", s.m, s.wahl, s.op, f.name)
			if f.kontext {
				if r.err != f.err || fehlercode(r.err) != 0 {
					t.Errorf("%s: Fehler %v, erwartet derselbe Kontextfehler ohne Code", kontext, r.err)
				}
			} else {
				pruefeFehler(t, kontext, r.err, e9001)
			}
			if r.ant.Header != nil || r.ant.Body != nil {
				t.Errorf("%s: Antwort neben dem Fehler", kontext)
			}
			if ops := u.p.ops(); ops[len(ops)-1] != s.op {
				t.Errorf("%s: nach dem Fehler lief %s", kontext, ops[len(ops)-1])
			}
			for _, e := range u.p.ev {
				if e.ctx != nil && e.ctx != u.ctx {
					t.Errorf("%s: %s ohne das ctx des Aufrufs", kontext, e.op)
				}
			}
			u.unveraendert(vorher, rev, kontext)
		}
	}
	if n != 14*7 {
		t.Errorf("Menge %d", n)
	}
}

// TestPreferDarstellungLeer (Z22, SPEC-288 Nr. 2): ein leeres Kennzeichen (Name oder Wert) oder ein leerer ContentType der Darstellung von SchreibeEntity und
// von SchreibeAnzahl ist `ODR-E9001` ohne Antwort, mit Rollback. Antwort mit einem der Aufrufe (Methode 3 × Wahl 3) × Verstoß 4 × Aufruf.
func TestPreferDarstellungLeer(t *testing.T) {
	leer := []model.Darstellung{
		{ContentType: "text/plain", Kennzeichen: model.Kopfzeile{Name: "", Wert: "1"}},
		{ContentType: "text/plain", Kennzeichen: model.Kopfzeile{Name: "X-A", Wert: ""}},
		{ContentType: "", Kennzeichen: model.Kopfzeile{Name: "X-A", Wert: "1"}},
		{},
	}
	n := 0
	for _, s := range preferStellen() {
		if s.op == "Schreiber.Adresse" {
			continue
		}
		for _, d := range leer {
			n++
			u := preferUmg(t, true)
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			u.c.sch.kopfe = map[string]model.Darstellung{s.op: d}
			w := preferWahlNamens(s.wahl)
			r := u.preferSende(s.m, schluesselFuer(s.m), preferKopf(w.zeilen), nil)
			kontext := fmt.Sprintf("%s %s %s %+v", s.m, s.wahl, s.op, d)
			f := erwarteFehler(t, kontext, r.err, e9001)
			if f != nil && f.Ursache != nil || r.ant.Header != nil || r.ant.Body != nil {
				t.Errorf("%s: Antwort oder Ursache neben dem Fehler", kontext)
			}
			u.unveraendert(vorher, rev, kontext)
		}
	}
	if n != 4*(14-5) {
		t.Errorf("Menge %d", n)
	}
}

// TestPreferGleichzeitig (Z23, SPEC-289 Nr. 2): von n gleichzeitigen PATCH mit `return=representation` auf dieselbe Entity trägt jede Antwort `200` und die Entity
// ihres Eintrags: genau die ETags der Versionen v+1 bis v+n, jeder einmal, in den Bodys und in den Kopfzeilen. n = 2, 8, 32, je 20 frische Aufbauten.
func TestPreferGleichzeitig(t *testing.T) {
	for _, n := range []int{2, 8, 32} {
		for wdh := 0; wdh < 20; wdh++ {
			const v = 5
			u := etagUmg(t, true, v, v, v)
			u.c.sch.zeigeEtag = true
			u.etagDynamisch()
			m := u.mocker()
			var mu sync.Mutex
			var etags, kopfEtags []string
			aufEinmal(n, func(i int) {
				a := dynAnfrage("PATCH", "1", "n"+strconv.Itoa(i), wsID)
				a.Header["Prefer"] = []string{"return=representation"}
				ant, err := m.Mock(u.ctx, a)
				if err != nil || ant.Status != 200 {
					t.Errorf("n=%d: Status %d, Fehler %v", n, ant.Status, err)
					return
				}
				b, _ := io.ReadAll(ant.Body)
				_ = ant.Body.Close()
				text := string(b)
				kopfEtag, _ := etagWerte(ant)
				mu.Lock()
				etags = append(etags, text[strings.LastIndex(text, "|")+1:])
				kopfEtags = append(kopfEtags, kopfEtag...)
				mu.Unlock()
			})
			slices.Sort(etags)
			slices.Sort(kopfEtags)
			var want []string
			for k := 1; k <= n; k++ {
				want = append(want, etagLit(int64(v+k)))
			}
			slices.Sort(want)
			if !slices.Equal(etags, want) || !slices.Equal(kopfEtags, want) {
				t.Fatalf("n=%d Wiederholung %d: ETags der Bodys %v, der Kopfzeilen %v, erwartet %v", n, wdh, etags, kopfEtags, want)
			}
		}
	}
}

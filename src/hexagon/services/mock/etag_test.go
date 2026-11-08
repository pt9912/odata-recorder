package mock_test

import (
	"context"
	"errors"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// etagLit ist der ETag zu einer Version als Literal des Vertrags (SPEC-280 Nr. 1), unabhängig von model.EtagFuer gebaut.
func etagLit(v int64) string { return "W/" + strconv.Quote(strconv.FormatInt(v, 10)) }

// etagUmg ist der Aufbau der Mutationen mit den Einträgen 1 bis 3 von Products in den Versionen vs und dem Schalter Etag.
func etagUmg(t *testing.T, an bool, vs ...int64) *umg {
	t.Helper()
	u := schreibUmg(t)
	es := []model.EntityState{voll(1, "a"), voll(2, "b"), voll(3, "c")}
	for i, v := range vs {
		es[i].Version = v
	}
	u.en.setze(setProducts, es...)
	u.k.Dienste[0].Etag = an
	return u
}

// etagWerte sind alle Werte der Kopfzeilen mit dem Namen ETag (ohne Beachtung der Schreibweise des Namens); namen sind alle Namen
// der Antwort, sortiert.
func etagWerte(a inbound.Antwort) (werte, namen []string) {
	for n, w := range a.Header {
		namen = append(namen, n)
		if strings.EqualFold(n, "etag") {
			werte = append(werte, w...)
		}
	}
	sort.Strings(namen)
	return werte, namen
}

// ohneEtag sind die Namen ohne die Kopfzeile ETag.
func ohneEtag(namen []string) []string {
	return slices.DeleteFunc(slices.Clone(namen), func(n string) bool { return strings.EqualFold(n, "etag") })
}

// etagsImSchreiber sind die Etag-Werte der Entities, die der Dienst an SchreibeEntity (je Aufruf eines) und SchreibeCollection
// (je Element) übergab, in der Folge der Aufrufe.
func etagsImSchreiber(u *umg) []string {
	var r []string
	for _, e := range u.p.von("Schreiber.SchreibeEntity") {
		r = append(r, e.args[2].(model.EntityAntwort).Entity.Etag)
	}
	for _, e := range u.p.von("Schreiber.SchreibeCollection") {
		for _, x := range e.args[2].(model.CollectionAntwort).Entities {
			r = append(r, x.Etag)
		}
	}
	return r
}

// etagArt ist eine der zehn Arten von Antworten (SPEC-281 Nr. 1): wie sie entsteht, ihr Status, die Namen ihrer Header ohne
// ETag und der ETag bei an ("" heißt keiner). Die Einträge 1 bis 3 haben die Versionen 4, 6 und 9.
type etagArt struct {
	name   string
	rufe   func(u *umg) (inbound.Antwort, error)
	status int
	namen  []string
	etag   string
	// elemente sind die Etags im Schreiber bei an: je Entity einer Antwort mit Entity, je Element einer Collection.
	elemente []string
}

func etagArten() []etagArt {
	lesen := func(methode, id string) func(u *umg) (inbound.Antwort, error) {
		return func(u *umg) (inbound.Antwort, error) {
			u.entity(model.Optionen{}, id)
			return u.rufe(anfrage(methode, dienstRoot+"/x"))
		}
	}
	mutation := func(methode, id string) func(u *umg) (inbound.Antwort, error) {
		return func(u *umg) (inbound.Antwort, error) { return u.mutiere(methode, id, eingabeIDName(id, "x")) }
	}
	kennz := "X-Fake-Kennzeichen"
	return []etagArt{
		{"GET Entity", lesen("GET", "2"), 200, []string{"Content-Type", kennz}, etagLit(6), []string{etagLit(6)}},
		{"HEAD Entity", lesen("HEAD", "2"), 200, []string{"Content-Type", kennz}, etagLit(6), []string{etagLit(6)}},
		{"GET Collection", func(u *umg) (inbound.Antwort, error) { u.menge(model.Optionen{}); return u.leseGET() }, 200,
			[]string{"Content-Type", kennz, "X-Recorder-Coverage"}, "", []string{etagLit(4), etagLit(6), etagLit(9)}},
		{"GET $count", func(u *umg) (inbound.Antwort, error) { u.anzahl(model.Optionen{}); return u.leseGET() }, 200,
			[]string{"Content-Type", kennz, "X-Recorder-Coverage"}, "", nil},
		{"GET $metadata", func(u *umg) (inbound.Antwort, error) { return u.rufe(metadatenAnfrage()) }, 200, []string{"Content-Type", kennz}, "", nil},
		{"POST", mutation("POST", "9"), 201, []string{"Content-Type", "Location", kennz}, etagLit(1), []string{etagLit(1)}},
		{"PATCH", mutation("PATCH", "2"), 204, []string{kennz}, etagLit(7), nil},
		{"PUT", mutation("PUT", "2"), 204, []string{kennz}, etagLit(7), nil},
		{"DELETE", mutation("DELETE", "2"), 204, []string{kennz}, "", nil},
		{"Fehler", lesen("GET", "99"), 0, nil, "", nil},
	}
}

// etagDienste sind die zwei Formen eines Dienstes: ohne Satz (der Satz der Konfig) und mit eigenem Satz.
func etagDienste() map[string]func(u *umg) {
	return map[string]func(u *umg){
		"Satz der Konfig": func(*umg) {},
		"eigener Satz":    func(u *umg) { u.k.Dienste[0].Codecs = u.c.satz(false, "4.0") },
	}
}

// TestEtagAusUndAn (E5, E7, E14, E15): zehn Arten × zwei Formen des Dienstes × aus und an. Aus trägt keine Antwort ETag oder Etag
// im Schreiber und die Header sind die Literale der Art; an trägt die Kopfzeile und das Mitglied nach SPEC-281 Nr. 1 und sonst nichts
// Zusätzliches (Status, Namen und Body gleich dem Stand aus).
func TestEtagAusUndAn(t *testing.T) {
	n := 0
	for dname, form := range etagDienste() {
		for _, art := range etagArten() {
			var ausBody string
			for _, an := range []bool{false, true} {
				u := etagUmg(t, an, 4, 6, 9)
				form(u)
				was := dname + " " + art.name + " an=" + strconv.FormatBool(an)
				a, err := art.rufe(u)
				n++
				if art.status == 0 {
					pruefeFehler(t, was, err, e6008)
					if a.Header != nil || a.Body != nil || a.Status != 0 {
						t.Errorf("%s: Antwort %+v neben dem Fehler", was, a)
					}
					continue
				}
				if err != nil || a.Status != art.status {
					t.Fatalf("%s: Status %d, Fehler %v, erwartet %d", was, a.Status, err, art.status)
				}
				werte, namen := etagWerte(a)
				body := string(lies(t, a))
				schreibers := etagsImSchreiber(u)
				if !an {
					ausBody = body
					if len(werte) != 0 || !slices.Equal(namen, art.namen) {
						t.Errorf("%s: ETag %v, Namen %v, erwartet keinen und %v", was, werte, namen, art.namen)
					}
					for _, e := range schreibers {
						if e != "" {
							t.Errorf("%s: Etag %q im Schreiber, erwartet keiner", was, e)
						}
					}
					continue
				}
				wantNamen := art.namen
				if art.etag != "" {
					wantNamen = append(slices.Clone(art.namen), "ETag")
					sort.Strings(wantNamen)
				}
				switch {
				case art.etag == "" && len(werte) != 0, art.etag != "" && !slices.Equal(werte, []string{art.etag}):
					t.Errorf("%s: ETag %v, erwartet %q", was, werte, art.etag)
				case !slices.Equal(namen, wantNamen):
					t.Errorf("%s: Namen %v, erwartet %v", was, namen, wantNamen)
				case !slices.Equal(ohneEtag(namen), art.namen):
					t.Errorf("%s: Namen ohne ETag %v, erwartet %v", was, ohneEtag(namen), art.namen)
				}
				if !slices.Equal(schreibers, art.elemente) {
					t.Errorf("%s: Etag im Schreiber %v, erwartet %v", was, schreibers, art.elemente)
				}
				// Der Dienst ändert den Body nicht: der Schreiber schreibt ihn, und der Fake schreibt für beide Fälle dieselben Bytes.
				if body != ausBody {
					t.Errorf("%s: Body %q, aus %q", was, body, ausBody)
				}
			}
		}
	}
	if n != 2*10*2 {
		t.Errorf("%d Fälle, erwartet 40", n)
	}
}

// TestEtagWiederholung (E7): zehn Lesungen derselben Entity in den Versionen 0, 1 und 5, GET und HEAD: derselbe ETag, der der Version;
// jede Lesung ruft Entities.Get genau einmal.
func TestEtagWiederholung(t *testing.T) {
	n := 0
	for _, v := range []int64{0, 1, 5} {
		for _, methode := range []string{"GET", "HEAD"} {
			u := etagUmg(t, true, v, v+10, v+20)
			u.entity(model.Optionen{}, "1")
			for i := 0; i < 10; i++ {
				a, err := u.rufe(anfrage(methode, dienstRoot+"/x"))
				if err != nil {
					t.Fatal(err)
				}
				werte, _ := etagWerte(a)
				lies(t, a)
				n++
				if !slices.Equal(werte, []string{etagLit(v)}) {
					t.Errorf("v=%d %s Lesung %d: ETag %v, erwartet %q", v, methode, i, werte, etagLit(v))
				}
			}
			if got := etagsImSchreiber(u); len(got) != 10 || slices.ContainsFunc(got, func(e string) bool { return e != etagLit(v) }) {
				t.Errorf("v=%d %s: Etag im Schreiber %v", v, methode, got)
			}
			if g := u.p.anzahl("Entities.Get"); g != 10 {
				t.Errorf("v=%d %s: %d Aufrufe von Entities.Get, erwartet 10", v, methode, g)
			}
		}
	}
	if n != 60 {
		t.Errorf("%d Lesungen, erwartet 60", n)
	}
}

// TestEtagEinLesen (E8): Header und Mitglied entstehen aus einem Lesen; der Fake liefert bei jedem Get eine höhere Version, also ist
// ein zweites Lesen am Header sichtbar. Je Art mit Entity: GET, HEAD und die drei Mutationen mit Antwort.
func TestEtagEinLesen(t *testing.T) {
	for _, art := range etagArten() {
		if art.etag == "" {
			continue
		}
		u := etagUmg(t, true, 4, 6, 9)
		u.en.wachsend = true
		u.c.sch.zeigeEtag = true
		a, err := art.rufe(u)
		if err != nil {
			t.Fatalf("%s: %v", art.name, err)
		}
		werte, _ := etagWerte(a)
		lies(t, a)
		gets := u.p.anzahl("Entities.Get")
		if gets != 1 || len(werte) != 1 {
			t.Errorf("%s: %d Aufrufe von Entities.Get, ETag %v; erwartet 1 und einen", art.name, gets, werte)
			continue
		}
		// Version des Eintrags 2 (6) und die Zahl der Gets (1) steht im Header; bei POST gibt es keinen Eintrag davor (Version 0 + 1).
		base := int64(6)
		if art.name == "POST" {
			base = 0
		}
		want := etagLit(base + 1)
		if art.name == "PATCH" || art.name == "PUT" {
			want = etagLit(base + 1 + 1)
		}
		if werte[0] != want {
			t.Errorf("%s: ETag %q, erwartet %q", art.name, werte[0], want)
		}
		if es := etagsImSchreiber(u); len(es) > 0 && es[0] != werte[0] {
			t.Errorf("%s: Etag im Schreiber %q, im Header %q", art.name, es[0], werte[0])
		}
		if lauf := u.p.anzahl("Entities.Put"); (art.name == "POST" || art.name == "PATCH" || art.name == "PUT") && lauf != 1 {
			t.Errorf("%s: %d Aufrufe von Entities.Put, erwartet 1", art.name, lauf)
		}
	}
}

// TestEtagHeaderKeinEinfluss (E6): bei aus ändern If-Match und If-None-Match nichts: Methode × Header × Wert, die Antwort (Status, Namen, ETag)
// gleich der ohne Header, der Zustand gleich dem nach der Anfrage ohne Header; bei an gilt SPEC-283 (if_match_test.go).
func TestEtagHeaderKeinEinfluss(t *testing.T) {
	werte := []string{"*", `W/"99"`, etagLit(6), etagLit(7), "", `"6"`}
	n := 0
	for _, an := range []bool{false} {
		for _, methode := range []string{"PATCH", "PUT", "DELETE"} {
			ohne := etagUmg(t, an, 4, 6, 9)
			aOhne, err := ohne.mutiere(methode, "2", eingabeIDName("2", "x"))
			if err != nil {
				t.Fatal(err)
			}
			wOhne, nOhne := etagWerte(aOhne)
			for _, kopf := range []string{"If-Match", "If-None-Match"} {
				for _, w := range werte {
					u := etagUmg(t, an, 4, 6, 9)
					u.setzeMutation(methode, "2", eingabeIDName("2", "x"))
					a := anfrage(methode, dienstRoot+"/x")
					a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
					a.Header["Content-Type"] = []string{"application/json"}
					a.Header[kopf] = []string{w}
					got, err := u.rufe(a)
					was := strings.Join([]string{strconv.FormatBool(an), methode, kopf, w}, " ")
					n++
					if err != nil || got.Status != aOhne.Status {
						t.Errorf("%s: Status %d, Fehler %v, ohne Header %d", was, got.Status, err, aOhne.Status)
						continue
					}
					if wg, ng := etagWerte(got); !slices.Equal(wg, wOhne) || !slices.Equal(ng, nOhne) {
						t.Errorf("%s: ETag %v Namen %v, ohne Header %v und %v", was, wg, ng, wOhne, nOhne)
					}
					if !slices.EqualFunc(u.zustandsliste(), ohne.zustandsliste(), func(a, b model.EntityState) bool {
						return a.EntityKey == b.EntityKey && a.Version == b.Version && a.Deleted == b.Deleted && a.Payload == b.Payload
					}) {
						t.Errorf("%s: Zustand verschieden von dem ohne Header", was)
					}
				}
			}
		}
	}
	if n != 3*2*len(werte) {
		t.Errorf("%d Fälle, erwartet %d", n, 3*2*len(werte))
	}
}

// zustandsliste sind die Einträge von Products nach Schlüssel.
func (u *umg) zustandsliste() []model.EntityState {
	l := u.en.inhalt()[setProducts]
	sort.Slice(l, func(i, j int) bool { return l[i].EntityKey < l[j].EntityKey })
	return l
}

// TestEtagNurAusVersion (E2): zwei Sets × zwei Schlüssel × drei Workspaces (zwei Services) bei gleicher Version tragen denselben ETag; eine andere
// Version einen anderen.
func TestEtagNurAusVersion(t *testing.T) {
	u := schreibUmg(t)
	u.k.Dienste[0].Etag = true
	u.k.Dienste = append(u.k.Dienste, mock.Dienst{ID: "svc2", Root: "/svc2", Etag: true})
	u.md.setze(model.Service{ID: "svc2", ServiceRoot: "http://up.test", ODataVersion: "4.0", MetadataBlob: metaRef})
	u.ws.setze(model.Workspace{ID: "ws2", ServiceID: dienstID, BaselineSnapshotID: "b1"})
	u.ws.setze(model.Workspace{ID: "ws3", ServiceID: "svc2", BaselineSnapshotID: "b1"})
	u.en.jeWs = true
	zellen := []struct{ ws, root string }{{"ws1", "/svc"}, {"ws2", "/svc"}, {"ws3", "/svc2"}}
	verschieden := map[string]bool{}
	n := 0
	for _, v := range []int64{0, 3, 12} {
		for _, z := range zellen {
			for _, set := range []string{setProducts, "Kunden"} {
				var l []model.EntityState
				for _, k := range []string{"1", "2"} {
					x := voll(1, "a")
					x.WorkspaceID, x.EntitySet, x.EntityKey, x.Version = z.ws, set, k, v
					l = append(l, x)
				}
				u.en.m[u.en.schl(z.ws, set)] = l
			}
		}
		for _, z := range zellen {
			for _, set := range []string{setProducts, "Kunden"} {
				for _, k := range []string{"1", "2"} {
					u.ressource(model.RessourceEntity, set, model.Optionen{}, zahl("", k))
					a := anfrage("GET", z.root+"/x")
					a.Header["X-Recorder-Workspace"] = []string{z.ws}
					got, err := u.rufe(a)
					if err != nil {
						t.Fatalf("%s %s %s %s: %v", z.ws, z.root, set, k, err)
					}
					w, _ := etagWerte(got)
					lies(t, got)
					n++
					if !slices.Equal(w, []string{etagLit(v)}) {
						t.Errorf("v=%d %s %s %s %s: ETag %v, erwartet %q", v, z.ws, z.root, set, k, w, etagLit(v))
					}
					verschieden[w[0]] = true
				}
			}
		}
	}
	if n != 3*3*2*2 || len(verschieden) != 3 {
		t.Errorf("%d Fälle und %d verschiedene ETags, erwartet 36 und 3", n, len(verschieden))
	}
}

// TestEtagCollection (E9, E11): jedes Element einer Collection trägt den ETag, den GET derselben Entity liefert; kein Header ETag; der
// Tombstone erscheint nicht; sieben Abfragen × acht Einträge mit gemischten Versionen und einem Tombstone.
func TestEtagCollection(t *testing.T) {
	versionen := []int64{0, 3, 1, 7, 2, 9, 4, 5}
	var es []model.EntityState
	for i, v := range versionen {
		x := voll(i+1, "n"+strconv.Itoa(i+1))
		x.Version = v
		es = append(es, x)
	}
	es[4].Deleted = true
	abfragen := map[string]model.Optionen{
		"ohne Option": {}, "top": {Top: model.ZahlOption{Gesetzt: true, Wert: 3}}, "skip": {Skip: model.ZahlOption{Gesetzt: true, Wert: 2}},
		"orderby": {OrderBy: []model.Sortierung{{Name: "Id", Absteigend: true}}}, "filter": {Filter: idGroesser("2")},
		"select": {Select: &model.SelectOption{Namen: []string{"Name"}}}, "count": {Count: model.WahrOption{Gesetzt: true, Wert: true}},
	}
	n, elemente := 0, 0
	for name, opt := range abfragen {
		u := etagUmg(t, true)
		u.en.setze(setProducts, es...)
		u.menge(opt)
		u.c.sch.zeigeEtag = true
		a, err := u.leseGET()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		werte, namen := etagWerte(a)
		lies(t, a)
		if len(werte) != 0 || !slices.Equal(namen, []string{"Content-Type", "X-Fake-Kennzeichen", "X-Recorder-Coverage"}) {
			t.Errorf("%s: ETag %v, Namen %v", name, werte, namen)
		}
		col := u.letzteCollection()
		if len(col.Entities) == 0 {
			t.Errorf("%s: keine Elemente", name)
		}
		for _, e := range col.Entities {
			elemente++
			i, _ := strconv.Atoi(e.Schluessel)
			if e.Schluessel == "5" {
				t.Errorf("%s: der Tombstone 5 erscheint", name)
			}
			if e.Etag != etagLit(versionen[i-1]) {
				t.Errorf("%s: Element %s trägt %q, erwartet %q", name, e.Schluessel, e.Etag, etagLit(versionen[i-1]))
			}
			// derselbe Text wie der GET derselben Entity
			u.entity(model.Optionen{}, e.Schluessel)
			g, err := u.leseGET()
			if err != nil {
				t.Fatalf("%s: GET %s: %v", name, e.Schluessel, err)
			}
			if w, _ := etagWerte(g); !slices.Equal(w, []string{e.Etag}) {
				t.Errorf("%s: GET %s trägt %v, das Element %q", name, e.Schluessel, w, e.Etag)
			}
			lies(t, g)
			u.p.leeren()
		}
		n++
	}
	if n != 7 || elemente < 7 {
		t.Errorf("%d Abfragen, %d Elemente", n, elemente)
	}
}

// TestEtagAccept (E11): Accept ändert Header und Mitglied am Dienst nie; Accept × Entity und Collection.
func TestEtagAccept(t *testing.T) {
	accepts := [][]string{nil, {"application/json;odata.metadata=none"}, {"application/json;odata.metadata=minimal"}, {"application/json;odata.metadata=full"},
		{"application/json;IEEE754Compatible=true"}, {"text/html"}, {"application/json", "text/plain"}}
	n := 0
	for _, acc := range accepts {
		for _, art := range []string{"Entity", "Collection"} {
			u := etagUmg(t, true, 4, 6, 9)
			if art == "Entity" {
				u.entity(model.Optionen{}, "2")
			} else {
				u.menge(model.Optionen{})
			}
			a := anfrage("GET", dienstRoot+"/x")
			a.Header["Accept"] = acc
			got, err := u.rufe(a)
			if err != nil {
				t.Fatal(err)
			}
			werte, _ := etagWerte(got)
			lies(t, got)
			es := etagsImSchreiber(u)
			want := []string{etagLit(6)}
			if art == "Collection" {
				es = nil
				if len(werte) != 0 {
					t.Errorf("Collection %v: ETag %v", acc, werte)
				}
				for _, e := range u.letzteCollection().Entities {
					es = append(es, e.Etag)
				}
				want = []string{etagLit(4), etagLit(6), etagLit(9)}
			} else if !slices.Equal(werte, want) {
				t.Errorf("Entity %v: ETag %v, erwartet %v", acc, werte, want)
			}
			if !slices.Equal(es, want) {
				t.Errorf("%s %v: Etag im Schreiber %v, erwartet %v", art, acc, es, want)
			}
			n++
		}
	}
	if n != 14 {
		t.Errorf("%d Fälle, erwartet 14", n)
	}
}

// TestEtagKeinETagBeiFehler (E12): jede Antwort mit Fehler ist ein *inbound.Fehler ohne Antwort, und $count und $metadata tragen keinen ETag.
func TestEtagKeinETagBeiFehler(t *testing.T) {
	tomb := voll(3, "c")
	tomb.Deleted = true
	faelle := []struct {
		name string
		f    func(u *umg) (inbound.Antwort, error)
		code model.Code
	}{
		{"GET fehlend", func(u *umg) (inbound.Antwort, error) { u.entity(model.Optionen{}, "99"); return u.leseGET() }, e6008},
		{"GET Tombstone", func(u *umg) (inbound.Antwort, error) {
			u.en.setze(setProducts, tomb)
			u.entity(model.Optionen{}, "3")
			return u.leseGET()
		}, e6008},
		{"GET fehlend partial", func(u *umg) (inbound.Antwort, error) {
			u.co.setze(baselineID, setProducts, model.CoveragePartial)
			u.entity(model.Optionen{}, "99")
			return u.leseGET()
		}, e6003},
		{"HEAD Tombstone", func(u *umg) (inbound.Antwort, error) {
			u.en.setze(setProducts, tomb)
			u.entity(model.Optionen{}, "3")
			return u.rufe(anfrage("HEAD", dienstRoot+"/x"))
		}, e6008},
		{"POST doppelt", func(u *umg) (inbound.Antwort, error) { return u.mutiere("POST", "", eingabeIDName("1", "n")) }, e6009},
		{"PATCH fehlend", func(u *umg) (inbound.Antwort, error) { return u.mutiere("PATCH", "99", eingabeIDName("99", "n")) }, e6008},
		{"PUT Tombstone", func(u *umg) (inbound.Antwort, error) {
			u.en.setze(setProducts, tomb)
			return u.mutiere("PUT", "3", eingabeIDName("3", "n"))
		}, e6008},
		{"DELETE fehlend", func(u *umg) (inbound.Antwort, error) { return u.mutiere("DELETE", "99", model.EntityEingabe{}) }, e6008},
		{"PATCH widerspricht", func(u *umg) (inbound.Antwort, error) { return u.mutiere("PATCH", "1", eingabeIDName("2", "n")) }, e6007},
	}
	for _, f := range faelle {
		u := etagUmg(t, true, 4, 6, 9)
		a, err := f.f(u)
		pruefeFehler(t, f.name, err, f.code)
		if a.Header != nil || a.Body != nil || a.Status != 0 {
			t.Errorf("%s: Antwort %+v neben dem Fehler", f.name, a)
		}
	}
	for _, art := range etagArten() {
		if art.name != "GET $count" && art.name != "GET $metadata" {
			continue
		}
		u := etagUmg(t, true, 4, 6, 9)
		a, err := art.rufe(u)
		if err != nil {
			t.Fatal(err)
		}
		if w, _ := etagWerte(a); len(w) != 0 {
			t.Errorf("%s: ETag %v", art.name, w)
		}
		lies(t, a)
	}
}

// etagVon ruft GET auf die Entity id und liefert ihren ETag (leer, wenn der Aufruf einen Fehler ergibt).
func (u *umg) etagVon(id string) string {
	u.t.Helper()
	u.entity(model.Optionen{}, id)
	a, err := u.leseGET()
	if err != nil {
		return ""
	}
	w, _ := etagWerte(a)
	lies(u.t, a)
	if len(w) != 1 {
		u.t.Fatalf("GET %s: ETag %v", id, w)
	}
	return w[0]
}

// TestEtagPost (E14): POST auf einen neuen Schlüssel im Set mit Einträgen und im Set ohne Eintrag: 201 mit ETag der Version 1, Location und die
// übrigen Header unverändert.
func TestEtagPost(t *testing.T) {
	for _, an := range []bool{false, true} {
		for _, set := range []string{setProducts, "Kunden"} {
			u := etagUmg(t, an, 4, 6, 9)
			u.c.sch.zeigeEtag = true
			e := eingabeIDName("9", "n")
			u.ressource(model.RessourceMenge, set, model.Optionen{})
			u.ei.ergebnis = e
			a := anfrage("POST", dienstRoot+"/x")
			a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
			a.Header["Content-Type"] = []string{"application/json"}
			got, err := u.rufe(a)
			if err != nil || got.Status != 201 {
				t.Fatalf("%s an=%v: Status %d, Fehler %v", set, an, got.Status, err)
			}
			w, namen := etagWerte(got)
			lies(t, got)
			if !slices.Equal(ohneEtag(namen), []string{"Content-Type", "Location", "X-Fake-Kennzeichen"}) || got.Header["Location"][0] != "http://adresse.test/fake" {
				t.Errorf("%s an=%v: Namen %v, Location %v", set, an, namen, got.Header["Location"])
			}
			if an && !slices.Equal(w, []string{etagLit(1)}) || !an && len(w) != 0 {
				t.Errorf("%s an=%v: ETag %v", set, an, w)
			}
		}
	}
}

// TestEtagMutationen (E15): PATCH, PUT und DELETE × Version vorher 0, 1 und 4: PATCH und PUT tragen den ETag der Version danach, gleich dem des
// folgenden GET und verschieden vom vorigen; DELETE trägt keinen. PATCH mit `{}` und PUT desselben Objekts erhöhen die Version.
func TestEtagMutationen(t *testing.T) {
	n := 0
	for _, v := range []int64{0, 1, 4} {
		for _, methode := range []string{"PATCH", "PUT", "DELETE"} {
			for _, leer := range []bool{false, true} {
				if leer && methode == "DELETE" {
					continue
				}
				u := etagUmg(t, true, v, v, v)
				vorher := u.etagVon("1")
				e := eingabeIDName("1", "b")
				if leer && methode == "PATCH" {
					e = model.EntityEingabe{Typ: typProduct, Properties: nil}
				}
				if leer && methode == "PUT" {
					// gleicher Inhalt wie der Eintrag: Id und Name b
					e = eingabeIDName("1", "a")
				}
				a, err := u.mutiere(methode, "1", e)
				if err != nil || a.Status != 204 {
					t.Fatalf("%s v=%d: Status %d, Fehler %v", methode, v, a.Status, err)
				}
				w, _ := etagWerte(a)
				body := lies(t, a)
				n++
				if len(body) != 0 {
					t.Errorf("%s v=%d: Body %q", methode, v, body)
				}
				if methode == "DELETE" {
					if len(w) != 0 {
						t.Errorf("DELETE v=%d: ETag %v", v, w)
					}
					continue
				}
				if !slices.Equal(w, []string{etagLit(v + 1)}) || w[0] == vorher {
					t.Errorf("%s v=%d leer=%v: ETag %v, vorher %q, erwartet %q", methode, v, leer, w, vorher, etagLit(v+1))
				}
				if nachher := u.etagVon("1"); nachher != w[0] {
					t.Errorf("%s v=%d: folgender GET %q, Antwort %q", methode, v, nachher, w[0])
				}
			}
		}
	}
	if n != 3*(2+2+1) {
		t.Errorf("%d Fälle, erwartet 15", n)
	}
}

// TestEtagFehlschlag (E16): eine fehlgeschlagene Mutation lässt Zähler und ETag unverändert und trägt keinen ETag: die Fehlerarten × PATCH, PUT, POST,
// DELETE, wo sie möglich sind; dazu Fehler nach Put (Advance, Schreiber), Rollback und beendeter Kontext vor und in der Transaktion.
func TestEtagFehlschlag(t *testing.T) {
	boese := errors.New("Port-Fehler")
	faelle := []struct {
		name    string
		methode string
		id      string
		e       model.EntityEingabe
		praep   func(u *umg)
	}{
		{"E6009 POST doppelt", "POST", "", eingabeIDName("1", "n"), nil},
		{"E6008 PATCH fehlend", "PATCH", "99", eingabeIDName("99", "n"), nil},
		{"E6008 PUT fehlend", "PUT", "99", eingabeIDName("99", "n"), nil},
		{"E6008 DELETE fehlend", "DELETE", "99", model.EntityEingabe{}, nil},
		{"Zustand E6003 PATCH fehlend, partial", "PATCH", "99", eingabeIDName("99", "n"), func(u *umg) { u.co.setze(baselineID, setProducts, model.CoveragePartial) }},
		{"Zustand E6003 PATCH partieller Eintrag", "PATCH", "1", eingabeIDName("1", "n"), func(u *umg) {
			x := voll(1, "a")
			x.Version, x.Completeness = 4, model.EntityPartial
			u.en.setze(setProducts, x, voll(2, "b"), voll(3, "c"))
		}},
		{"E6007 widerspricht", "PATCH", "1", eingabeIDName("2", "n"), nil},
		{"E6007 reserviert", "PATCH", "1", eingabeIDName("1", model.Marker), nil},
		{"E6007 Typ", "PUT", "1", model.EntityEingabe{Typ: typSpecial, Schluessel: "1", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1")}}, nil},
		{"E7001 Advance", "PATCH", "1", eingabeIDName("1", "n"), func(u *umg) { u.p.setzeFehler("Workspaces.Advance", outbound.ErrRevisionVeraltet) }},
		{"Fehler Advance PUT", "PUT", "1", eingabeIDName("1", "n"), func(u *umg) { u.p.setzeFehler("Workspaces.Advance", boese) }},
		{"Fehler Put POST", "POST", "", eingabeIDName("9", "n"), func(u *umg) { u.p.setzeFehler("Entities.Put", boese) }},
		{"Fehler Schreiber nach Put POST", "POST", "", eingabeIDName("9", "n"), func(u *umg) { u.p.setzeFehler("Schreiber.SchreibeEntity", boese) }},
		{"Fehler Schreiber nach Put PATCH", "PATCH", "1", eingabeIDName("1", "n"), func(u *umg) { u.p.setzeFehler("Schreiber.SchreibeAnzahl", boese) }},
		{"Fehler Adresse POST", "POST", "", eingabeIDName("9", "n"), func(u *umg) { u.p.setzeFehler("Schreiber.Adresse", boese) }},
		{"Kontext in der Transaktion Put", "DELETE", "1", model.EntityEingabe{}, func(u *umg) { u.p.setzeFehler("Entities.Put", context.Canceled) }},
		{"Kontext in der Transaktion Advance", "PUT", "1", eingabeIDName("1", "n"), func(u *umg) { u.p.setzeFehler("Workspaces.Advance", context.DeadlineExceeded) }},
		{"Fehler bei der Bestätigung", "PATCH", "1", eingabeIDName("1", "n"), func(u *umg) { u.tx.nachF = boese }},
	}
	for _, f := range faelle {
		u := etagUmg(t, true, 4, 4, 4)
		// Ein Fall "Zustand …" ändert den Zustand vor den Lesungen, jeder andere speist erst danach einen Fehler ein.
		zustand := strings.HasPrefix(f.name, "Zustand")
		if zustand {
			f.praep(u)
		}
		vorher := map[string]string{}
		vorherStand := u.zustandsliste()
		revision := u.ws.revision(wsID)
		for _, id := range []string{"1", "2", "3"} {
			vorher[id] = u.etagVon(id)
		}
		if !zustand && f.praep != nil {
			f.praep(u)
		}
		u.p.leeren()
		a, err := u.mutiere(f.methode, f.id, f.e)
		if err == nil {
			t.Errorf("%s: kein Fehler, Status %d", f.name, a.Status)
			continue
		}
		if a.Header != nil || a.Body != nil || a.Status != 0 {
			t.Errorf("%s: Antwort %+v neben dem Fehler", f.name, a)
		}
		// ohne Fehler der Fakes lesen: die Einspeisung gilt für jede Operation
		for _, op := range []string{"Workspaces.Advance", "Entities.Put", "Schreiber.SchreibeEntity", "Schreiber.SchreibeAnzahl", "Schreiber.Adresse"} {
			u.p.setzeFehler(op, nil)
		}
		u.tx.nachF = nil
		for _, id := range []string{"1", "2", "3"} {
			if got := u.etagVon(id); got != vorher[id] {
				t.Errorf("%s: ETag von %s nach dem Fehler %q, vorher %q", f.name, id, got, vorher[id])
			}
		}
		if !slices.EqualFunc(u.zustandsliste(), vorherStand, func(a, b model.EntityState) bool {
			return a.EntityKey == b.EntityKey && a.Version == b.Version && a.Deleted == b.Deleted && a.Payload == b.Payload
		}) {
			t.Errorf("%s: Zustand verändert", f.name)
		}
		if got := u.ws.revision(wsID); got != revision {
			t.Errorf("%s: Revision %d, vorher %d", f.name, got, revision)
		}
	}
}

// TestEtagAndereBleiben (E17): drei Entities in zwei Workspaces (gleiche Schlüssel); eine Mutation ändert allein den ETag der einen Entity.
func TestEtagAndereBleiben(t *testing.T) {
	for _, methode := range []string{"PATCH", "PUT", "DELETE"} {
		u := schreibUmg(t)
		u.k.Dienste[0].Etag = true
		u.ws.setze(model.Workspace{ID: "ws2", ServiceID: dienstID, BaselineSnapshotID: "b1"})
		u.en.jeWs = true
		stand := map[string]string{}
		for _, ws := range []string{wsID, "ws2"} {
			var l []model.EntityState
			for i, id := range []int{1, 2, 3} {
				x := voll(id, "n")
				x.WorkspaceID, x.Version = ws, int64(2+i)
				l = append(l, x)
			}
			u.en.m[u.en.schl(ws, setProducts)] = l
		}
		lesen := func(ws, id string) string {
			u.entity(model.Optionen{}, id)
			a := anfrage("GET", dienstRoot+"/x")
			a.Header["X-Recorder-Workspace"] = []string{ws}
			got, err := u.rufe(a)
			if err != nil {
				return ""
			}
			w, _ := etagWerte(got)
			lies(t, got)
			return w[0]
		}
		for _, ws := range []string{wsID, "ws2"} {
			for _, id := range []string{"1", "2", "3"} {
				stand[ws+id] = lesen(ws, id)
			}
		}
		u.setzeMutation(methode, "2", eingabeIDName("2", "x"))
		a := anfrage(methode, dienstRoot+"/x")
		a.Body, a.Laenge = io.NopCloser(strings.NewReader("{}")), 2
		a.Header["Content-Type"] = []string{"application/json"}
		if _, err := u.rufe(a); err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, ws := range []string{wsID, "ws2"} {
			for _, id := range []string{"1", "2", "3"} {
				n++
				got := lesen(ws, id)
				if ws == wsID && id == "2" {
					if methode != "DELETE" && got == stand[ws+id] || methode == "DELETE" && got != "" {
						t.Errorf("%s: Ziel %s/%s trägt %q (vorher %q)", methode, ws, id, got, stand[ws+id])
					}
					continue
				}
				if got != stand[ws+id] {
					t.Errorf("%s: %s/%s trägt %q, vorher %q", methode, ws, id, got, stand[ws+id])
				}
			}
		}
		if n != 6 {
			t.Errorf("%d Einträge, erwartet 6", n)
		}
	}
}

// TestEtagLoeschenNeuanlage (E20): Löschen und Neuanlage desselben Schlüssels: alle ETags, die der Schlüssel in einer Geschichte trug, sind
// paarweise verschieden und streng steigend. Vorgeschichte (Version 0; 1 bis 3 PATCH; schon einmal gelöscht und neu angelegt) × 1 bis 3
// Wiederholungen von DELETE und POST.
func TestEtagLoeschenNeuanlage(t *testing.T) {
	ver := func(etag string) int64 {
		v, err := strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(etag, `W/"`), `"`), 10, 64)
		if err != nil {
			t.Fatalf("ETag %q: %v", etag, err)
		}
		return v
	}
	n := 0
	for vorgeschichte := 0; vorgeschichte <= 4; vorgeschichte++ {
		for wiederholungen := 1; wiederholungen <= 3; wiederholungen++ {
			u := etagUmg(t, true, 0, 0, 0)
			gesehen := []string{u.etagVon("1")}
			switch {
			case vorgeschichte >= 1 && vorgeschichte <= 3:
				for i := 0; i < vorgeschichte; i++ {
					a, err := u.mutiere("PATCH", "1", eingabeIDName("1", "p"+strconv.Itoa(i)))
					if err != nil {
						t.Fatal(err)
					}
					w, _ := etagWerte(a)
					gesehen = append(gesehen, w[0])
				}
			}
			// Vorgeschichte 4: der Schlüssel wurde schon einmal gelöscht und neu angelegt.
			davor := 0
			if vorgeschichte == 4 {
				davor = 1
			}
			for i := 0; i < davor+wiederholungen; i++ {
				a, err := u.mutiere("DELETE", "1", model.EntityEingabe{})
				if err != nil {
					t.Fatal(err)
				}
				lies(t, a)
				if got := u.etagVon("1"); got != "" {
					t.Errorf("Tombstone trägt den ETag %q", got)
				}
				a, err = u.mutiere("POST", "", eingabeIDName("1", "neu"))
				if err != nil {
					t.Fatal(err)
				}
				w, _ := etagWerte(a)
				lies(t, a)
				gesehen = append(gesehen, w[0])
			}
			n++
			for i := 1; i < len(gesehen); i++ {
				if ver(gesehen[i]) <= ver(gesehen[i-1]) {
					t.Errorf("Vorgeschichte %d, %d Wiederholungen: ETags %v nicht streng steigend", vorgeschichte, wiederholungen, gesehen)
					break
				}
			}
			for i := range gesehen {
				for j := i + 1; j < len(gesehen); j++ {
					if gesehen[i] == gesehen[j] {
						t.Errorf("Vorgeschichte %d: ETag %q zweimal in %v", vorgeschichte, gesehen[i], gesehen)
					}
				}
			}
			if last := gesehen[len(gesehen)-1]; u.etagVon("1") != last {
				t.Errorf("Vorgeschichte %d: GET %q, Antwort des POST %q", vorgeschichte, u.etagVon("1"), last)
			}
		}
	}
	if n != 15 {
		t.Errorf("%d Geschichten, erwartet 15", n)
	}
}

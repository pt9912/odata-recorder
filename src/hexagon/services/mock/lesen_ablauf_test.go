package mock_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// namenVon nennt die Namen der Properties einer Entity, kommagetrennt.
func namenVon(e model.Entity) string {
	var r []string
	for _, p := range e.Properties {
		r = append(r, p.Name)
	}
	return strings.Join(r, ",")
}

// beobachte fasst die Wirkung eines Aufrufs in einen Text: den Code des Fehlers oder die Ausgabe des Schreibers (Entities,
// Anzahl, Auswahl). Der Body wird gelesen und geschlossen.
func beobachte(u *umg, art model.RessourcenArt, a inbound.Antwort, err error) string {
	u.t.Helper()
	if err != nil {
		return "ERR " + fehlercode(err).String()
	}
	lies(u.t, a)
	sel := func(s []string) string {
		if s == nil {
			return "-"
		}
		return strings.Join(s, ",")
	}
	switch art {
	case model.RessourceMenge:
		c := u.letzteCollection()
		n := "-"
		if c.Anzahl.Vorhanden {
			n = strconv.FormatInt(c.Anzahl.Wert, 10)
		}
		return fmt.Sprintf("C %s n=%s sel=%s", ids(c.Entities), n, sel(c.Auswahl))
	case model.RessourceEntity:
		e := u.letzteEntity()
		return "E " + namenVon(e.Entity) + " sel=" + sel(e.Auswahl)
	case model.RessourceAnzahl:
		return fmt.Sprintf("A %d", u.letzteAnzahl())
	}
	return "M"
}

// TestOptionenJeArt hält L01 (Sendung (b), SPEC-196 Nr. 1): Art 4 × Option 7 = 28 Zellen, Orakel die Tabelle von Hand; `$filter`
// an der Entity ist ODR-E6001, `/$count` ignoriert `$top`, `$skip` und `$orderby`, `$select` und `$count` wirken nur, wo die
// Tabelle es sagt.
func TestOptionenJeArt(t *testing.T) {
	sel := &model.SelectOption{Namen: []string{"Name"}}
	// Bei `/$count` hat `$top`, `$skip` und `$orderby` keine sichtbare Wirkung der Zahl; die Zeile nimmt dort Werte, die der
	// Interpreter abweist (ungültige Zahl, unbekannte Property), damit ein Dienst, der die Option weiterreicht, auffällt.
	negativ := model.ZahlOption{Gesetzt: true, Wert: -1}
	unbekannt := &model.SelectOption{Namen: []string{"Unbekannt"}}
	opts := []struct {
		name string
		o    model.Optionen
		anz  model.Optionen
	}{
		{"filter", model.Optionen{Filter: idGroesser("1")}, model.Optionen{Filter: idGroesser("1")}},
		{"orderby", model.Optionen{OrderBy: []model.Sortierung{{Name: "Id", Absteigend: true}}}, model.Optionen{OrderBy: []model.Sortierung{{Name: "Unbekannt"}}}},
		{"top", model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1}}, model.Optionen{Top: negativ}},
		{"skip", model.Optionen{Skip: model.ZahlOption{Gesetzt: true, Wert: 1}}, model.Optionen{Skip: negativ}},
		{"count", model.Optionen{Count: model.WahrOption{Gesetzt: true, Wert: true}}, model.Optionen{Count: model.WahrOption{Gesetzt: true, Wert: true}}},
		{"select", model.Optionen{Select: sel}, model.Optionen{Select: unbekannt}},
		{"sonstige", model.Optionen{Sonstige: []model.SonstigeOption{{Name: "x", Wert: "y"}}}, model.Optionen{Sonstige: []model.SonstigeOption{{Name: "x", Wert: "y"}}}},
	}
	basis := map[model.RessourcenArt]string{
		model.RessourceMetadaten: "M", model.RessourceMenge: "C 1,2,3 n=- sel=-", model.RessourceEntity: "E Id,Name sel=-",
		model.RessourceAnzahl: "A 3",
	}
	abweichung := map[string]string{
		"2/filter": "C 2,3 n=- sel=-", "2/orderby": "C 3,2,1 n=- sel=-", "2/top": "C 1 n=- sel=-", "2/skip": "C 2,3 n=- sel=-",
		"2/count": "C 1,2,3 n=3 sel=-", "2/select": "C 1,2,3 n=- sel=Name",
		"3/filter": "ERR ODR-E6001", "3/select": "E Name sel=Name",
		"4/filter": "A 2",
	}
	n := 0
	for _, art := range model.AlleRessourcenArten() {
		for _, o := range opts {
			n++
			u := leseUmg(t)
			set := setProducts
			if art == model.RessourceMetadaten {
				set = ""
			}
			opt := o.o
			if art == model.RessourceAnzahl {
				opt = o.anz
			}
			u.ressource(art, set, opt, zahl("", "1"))
			a, err := u.leseGET()
			soll, da := abweichung[fmt.Sprint(int(art))+"/"+o.name]
			if !da {
				soll = basis[art]
			}
			if got := beobachte(u, art, a, err); got != soll {
				t.Errorf("Art %d, Option %s: %q, erwartet %q", art, o.name, got, soll)
			}
		}
	}
	if n != 4*7 {
		t.Errorf("%d Zellen, erwartet 28", n)
	}
}

// TestSetUnbekannt hält L02 (SPEC-196 Nr. 2): das Entity-Set muss bytegleich im Modell stehen, sonst ODR-E6008 vor der Coverage
// und vor jedem Lesen der Entities. Art 3 × Name 8; das Kelvin-Zeichen U+212A (E2 84 AA) entsteht aus Bytewerten und ist keine
// Schreibweise von K im Sinn dieses Vergleichs.
func TestSetUnbekannt(t *testing.T) {
	namen := []string{"products", "PRODUCTS", "Products ", "Product", "Prod ucts", "", "\xe2\x84\xaaunden", strings.Repeat("P", 1<<20)}
	n := 0
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl} {
		for _, name := range namen {
			n++
			u := leseUmg(t)
			u.ressource(art, name, model.Optionen{}, zahl("", "1"))
			_, err := u.leseGET()
			_ = erwarteFehler(t, fmt.Sprintf("Art %d, Name von %d Bytes", art, len(name)), err, model.CodeE6008)
			for _, op := range []string{"Coverage.Get", "Entities.Get", "Entities.ListSet", "Schreiber.SchreibeCollection", "Schreiber.SchreibeEntity"} {
				if u.p.anzahl(op) != 0 {
					t.Errorf("Art %d, Name von %d Bytes: %s gerufen", art, len(name), op)
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Läufe, erwartet 24", n)
	}
	// Kontrolle: der Name des Modells geht weiter.
	u := leseUmg(t)
	u.ressource(model.RessourceMenge, "Kunden", model.Optionen{})
	if _, err := u.leseGET(); err != nil {
		t.Errorf("Kunden: %v", err)
	}
}

// TestModellLesen hält L03 (SPEC-196 Nr. 2): Art 3 × Ausgang 4 (Erfolg, ErrBlobFehlt, ErrBlobBeschaedigt, ErrMetadatenUngueltig):
// Open und ParseMetadata je genau einmal mit ctx, Strom und MaxBytes unverändert, der Strom genau einmal geschlossen; zwei
// Aufrufe sind zwei Parse (kein Cache).
func TestModellLesen(t *testing.T) {
	faelle := []struct {
		name string
		op   string
		err  error
		code model.Code
	}{
		{"Erfolg", "", nil, 0},
		{"Blob fehlt", "Blobs.Open", outbound.ErrBlobFehlt, model.CodeE3001},
		{"Blob beschädigt", "Blobs.Open", outbound.ErrBlobBeschaedigt, model.CodeE3001},
		{"Metadaten ungültig", "MetadatenCodec.ParseMetadata", model.ErrMetadatenUngueltig, model.CodeE6004},
	}
	n := 0
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl} {
		for _, f := range faelle {
			n++
			u := leseUmg(t)
			u.ressource(art, setProducts, model.Optionen{}, zahl("", "1"))
			if f.op != "" {
				u.p.setzeFehler(f.op, f.err)
			}
			_, err := u.leseGET()
			if f.code == 0 && err != nil || f.code != 0 && fehlercode(err) != f.code {
				t.Errorf("Art %d, %s: %v", art, f.name, err)
			}
			opens := u.p.von("Blobs.Open")
			if len(opens) != 1 || opens[0].ctx != u.ctx || !reflect.DeepEqual(opens[0].args, []any{metaRef}) {
				t.Errorf("Art %d, %s: Open %v", art, f.name, opens)
			}
			stroeme := u.bl.strome()
			parses := u.p.von("MetadatenCodec.ParseMetadata")
			if f.op == "Blobs.Open" {
				if len(stroeme) != 0 || len(parses) != 0 {
					t.Errorf("Art %d, %s: %d Ströme, %d Parse", art, f.name, len(stroeme), len(parses))
				}
				continue
			}
			if len(stroeme) != 1 || stroeme[0].geschlossen() != 1 || len(parses) != 1 {
				t.Errorf("Art %d, %s: %d Ströme, %d Parse", art, f.name, len(stroeme), len(parses))
				continue
			}
			if p := parses[0]; p.ctx != u.ctx || p.args[0] != any(stroeme[0]) || p.args[1] != u.k.MaxBytes {
				t.Errorf("Art %d, %s: Parse mit %v", art, f.name, p.args)
			}
		}
	}
	if n != 12 {
		t.Errorf("%d Läufe, erwartet 12", n)
	}
	u := leseUmg(t)
	u.menge(model.Optionen{})
	m := u.mocker()
	for i := 0; i < 2; i++ {
		if _, err := m.Mock(u.ctx, anfrage("GET", dienstRoot+"/x")); err != nil {
			t.Fatal(err)
		}
	}
	if u.p.anzahl("Blobs.Open") != 2 || u.p.anzahl("MetadatenCodec.ParseMetadata") != 2 {
		t.Errorf("zwei Aufrufe: %d Open, %d Parse", u.p.anzahl("Blobs.Open"), u.p.anzahl("MetadatenCodec.ParseMetadata"))
	}
}

// TestCollectionAblauf hält L14 (SPEC-196 Nr. 5): Schleife über `$count` (wahr, falsch, nicht gesetzt): ListSet genau einmal mit
// (ctx, Workspace, Set), ohne Tombstones, FuehreAus mit den Optionen des Parsers unverändert; Anzahl.Vorhanden gleich Gesetzt und
// Wert, Anzahl.Wert gleich die Zahl nach dem Filter; ErrNichtGefunden von ListSet ist ODR-E7001.
func TestCollectionAblauf(t *testing.T) {
	zaehlt := []model.WahrOption{{}, {Gesetzt: true, Wert: true}, {Gesetzt: true, Wert: false}}
	for _, c := range zaehlt {
		u := leseUmg(t)
		tot := produkt(4, "d")
		tot.Deleted = true
		u.en.setze(setProducts, produkt(1, "a"), produkt(2, "b"), tot, produkt(3, "c"))
		u.menge(model.Optionen{
			Filter: idGroesser("0"), OrderBy: []model.Sortierung{{Name: "Id", Absteigend: true}},
			Top: model.ZahlOption{Gesetzt: true, Wert: 2}, Count: c,
		})
		a, err := u.leseGET()
		if err != nil {
			t.Fatal(err)
		}
		lies(t, a)
		ev := u.p.von("Entities.ListSet")
		if len(ev) != 1 || ev[0].ctx != u.ctx || !reflect.DeepEqual(ev[0].args, []any{wsID, setProducts}) {
			t.Errorf("Count %+v: ListSet %v", c, ev)
		}
		ant := u.letzteCollection()
		if ids(ant.Entities) != "3,2" || ant.Anzahl.Vorhanden != (c.Gesetzt && c.Wert) || ant.Anzahl.Wert != 3 || ant.Set != setProducts ||
			ant.Wurzel != "http://recorder.test/svc" {
			t.Errorf("Count %+v: Antwort %+v", c, ant)
		}
	}
	for _, r := range []model.RessourcenArt{model.RessourceMenge, model.RessourceAnzahl} {
		u := leseUmg(t)
		u.ressource(r, setProducts, model.Optionen{})
		u.p.setzeFehler("Entities.ListSet", outbound.ErrNichtGefunden)
		_, err := u.leseGET()
		_ = erwarteFehler(t, "ListSet ohne Workspace", err, model.CodeE7001)
	}
}

// TestDarstellungKopf hält L15 (SPEC-198 Nr. 2, Sendung von slice-odata-json-schreiben): der Fake liefert je Ausgabe fremde
// Namen und Werte; Content-Type und Kennzeichen stehen unverändert in der Antwort, kein weiterer Header außer dem der
// Coverage; Schleife über 4 Ausgaben ($metadata, Collection, Entity, Anzahl). Der Quelltext nennt keinen Namen (TestQuelltextOhneVersion).
func TestDarstellungKopf(t *testing.T) {
	ausgaben := []struct {
		art    model.RessourcenArt
		op     string
		ct     string
		kz     model.Kopfzeile
		erwCT  string
		header int
	}{
		{model.RessourceMetadaten, "Schreiber.SchreibeAnzahl", "text/x-probe", model.Kopfzeile{Name: "X-Fremd-M", Wert: "m 1"}, "application/xml", 2},
		{model.RessourceMenge, "Schreiber.SchreibeCollection", "application/x-menge;a=b", model.Kopfzeile{Name: "X-Fremd-C", Wert: "c;2"}, "application/x-menge;a=b", 3},
		{model.RessourceEntity, "Schreiber.SchreibeEntity", "application/x-entity", model.Kopfzeile{Name: "X-Fremd-E", Wert: "e3"}, "application/x-entity", 2},
		{model.RessourceAnzahl, "Schreiber.SchreibeAnzahl", "text/x-zahl", model.Kopfzeile{Name: "X-Fremd-A", Wert: "a4"}, "text/x-zahl", 3},
	}
	for _, o := range ausgaben {
		u := leseUmg(t)
		u.c.sch.kopfe = map[string]model.Darstellung{o.op: {ContentType: o.ct, Kennzeichen: o.kz}}
		set := setProducts
		if o.art == model.RessourceMetadaten {
			set = ""
		}
		u.ressource(o.art, set, model.Optionen{}, zahl("", "1"))
		a, err := u.leseGET()
		if err != nil {
			t.Fatal(err)
		}
		lies(t, a)
		if got := a.Header["Content-Type"]; len(got) != 1 || got[0] != o.erwCT {
			t.Errorf("%s: Content-Type %v", o.op, got)
		}
		if got := a.Header[o.kz.Name]; len(got) != 1 || got[0] != o.kz.Wert || len(a.Header) != o.header {
			t.Errorf("%s: Kennzeichen %v, %d Header: %v", o.op, got, len(a.Header), a.Header)
		}
	}
}

// TestWurzel hält L16 (SPEC-198 Nr. 6): Wurzel ist PublicBaseURL und der Root ohne abschließenden `/` (Root `/`: allein
// PublicBaseURL), nie ein Header; Root 3 × PublicBaseURL 2 × Header 3 (keiner, Host, X-Forwarded-Host).
func TestWurzel(t *testing.T) {
	n := 0
	for _, root := range []string{"/", "/svc", "/a/b"} {
		for _, basis := range []string{"http://recorder.test", "https://anders.test:8443"} {
			for _, h := range []string{"", "Host", "X-Forwarded-Host"} {
				n++
				u := leseUmg(t)
				u.k.Dienste = []mock.Dienst{{ID: dienstID, Root: root}}
				u.k.PublicBaseURL = basis
				u.menge(model.Optionen{})
				a := anfrage("GET", strings.TrimSuffix(root, "/")+"/x")
				if h != "" {
					a.Header[h] = []string{"evil.example"}
				}
				if _, err := u.rufe(a); err != nil {
					t.Fatal(err)
				}
				soll := basis + strings.TrimSuffix(root, "/")
				if got := u.letzteCollection().Wurzel; got != soll {
					t.Errorf("Root %q, Basis %q, Header %q: Wurzel %q, erwartet %q", root, basis, h, got, soll)
				}
			}
		}
	}
	if n != 18 {
		t.Errorf("%d Läufe, erwartet 18", n)
	}
}

// TestAccept hält L17 (M16, SPEC-198 Nr. 7): die Werte aller Schreibweisen des Namens, die Schreibweisen in der Bytefolge der
// Namen, dann in ihrer Reihenfolge, unverändert (leer und ungültig bleiben), ohne Header nil; der Name mit dem kyrillischen
// е (Bytes D0 B5) ist ein anderer Name. 10 Formen × Collection und Entity.
func TestAccept(t *testing.T) {
	kyr := "Acc\xd0\xb5pt"
	formen := []struct {
		name   string
		header map[string][]string
		soll   []string
	}{
		{"kein Header", nil, nil},
		{"leere Liste", map[string][]string{"Accept": {}}, nil},
		{"ein Wert", map[string][]string{"Accept": {"application/json"}}, []string{"application/json"}},
		{"zwei Werte", map[string][]string{"Accept": {"b/b", "a/a"}}, []string{"b/b", "a/a"}},
		{"drei Schreibweisen", map[string][]string{"accept": {"x"}, "ACCEPT": {"y", "y2"}, "Accept": {"z"}}, []string{"y", "y2", "z", "x"}},
		{"leerer Wert", map[string][]string{"Accept": {""}}, []string{""}},
		{"Wert mit q", map[string][]string{"Accept": {"application/json;q=0.5"}}, []string{"application/json;q=0.5"}},
		{"ungültiger Wert", map[string][]string{"Accept": {"@@@ ;;"}}, []string{"@@@ ;;"}},
		{"odata.metadata=none", map[string][]string{"Accept": {"application/json;odata.metadata=none"}}, []string{"application/json;odata.metadata=none"}},
		{"kyrillisches e", map[string][]string{"Accept": {"a"}, kyr: {"x"}}, []string{"a"}},
	}
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity} {
		for _, f := range formen {
			u := leseUmg(t)
			u.ressource(art, setProducts, model.Optionen{}, zahl("", "1"))
			a := anfrage("GET", dienstRoot+"/x")
			for k, v := range f.header {
				a.Header[k] = v
			}
			if _, err := u.rufe(a); err != nil {
				t.Fatal(err)
			}
			var got []string
			if art == model.RessourceMenge {
				got = u.letzteCollection().Accept
			} else {
				got = u.letzteEntity().Accept
			}
			if !slices.Equal(got, f.soll) || (got == nil) != (f.soll == nil) {
				t.Errorf("Art %d, %s: Accept %#v, erwartet %#v", art, f.name, got, f.soll)
			}
		}
	}
}

// TestSchreiberFehler hält L42 (SPEC-198 Nr. 2): ErrUngueltig des Schreibers ist ODR-E9001, ein Kontextfehler bleibt ohne
// Code; es gibt keine Antwort. Ausgabe 3 × Fehler 2.
func TestSchreiberFehler(t *testing.T) {
	ops := map[model.RessourcenArt]string{
		model.RessourceMenge: "Schreiber.SchreibeCollection", model.RessourceEntity: "Schreiber.SchreibeEntity", model.RessourceAnzahl: "Schreiber.SchreibeAnzahl",
	}
	for art, op := range ops {
		for _, f := range []error{model.ErrUngueltig, context.DeadlineExceeded} {
			u := leseUmg(t)
			u.ressource(art, setProducts, model.Optionen{}, zahl("", "1"))
			u.p.setzeFehler(op, f)
			a, err := u.leseGET()
			if errors.Is(f, model.ErrUngueltig) {
				_ = erwarteFehler(t, op, err, model.CodeE9001)
			} else if fehlercode(err) != 0 || !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("%s: Kontextfehler wurde %v", op, err)
			}
			if a.Body != nil {
				t.Errorf("%s: Antwort trotz Fehler", op)
			}
		}
	}
}

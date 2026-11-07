package mock_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// erwarteStatus prüft Status (201 mit Location, 204 ohne) einer Antwort ohne Fehler.
func erwarteStatus(t *testing.T, kontext string, a inbound.Antwort, err error, status int) {
	t.Helper()
	if err != nil || a.Status != status {
		t.Errorf("%s: Status %d Fehler %v, erwartet %d", kontext, a.Status, err, status)
	}
}

// TestMethodeMutation (W01, M1): Methode 9 × Dienst 2 = 18. Die vier Methoden laufen (201, 204, 204, 204); jede andere,
// auch in anderer Schreibweise, ist ODR-E6002, bei einem Dienst NurLesend ist jede der neun ODR-E6006; bei jedem Fehler ist
// vor allem Zugriff auf den Workspace nichts gerufen.
func TestMethodeMutation(t *testing.T) {
	methoden := []string{"POST", "PATCH", "PUT", "DELETE", "OPTIONS", "post", "Patch", "MERGE", "TRACE"}
	n := 0
	for _, nurLesend := range []bool{false, true} {
		for _, m := range methoden {
			n++
			u := schreibUmg(t)
			if nurLesend {
				u.k.Dienste[0].Codecs = u.c.satz(true, "4.0")
			}
			id := "2"
			if m == "POST" {
				id = "7"
			}
			a, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
			kontext := fmt.Sprintf("%s nurLesend=%v", m, nurLesend)
			switch {
			case nurLesend:
				pruefeFehler(t, kontext, err, e6006)
			case m == "POST" || m == "PATCH" || m == "PUT" || m == "DELETE":
				erwarteStatus(t, kontext, a, err, sollStatus(m))
				continue
			default:
				pruefeFehler(t, kontext, err, e6002)
			}
			if u.p.anzahl("Workspaces.Get") != 0 || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 || u.p.anzahl("Eingabe.LeseEntity") != 0 {
				t.Errorf("%s: ein Port wurde gerufen: %v", kontext, u.p.ops())
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}

// TestPaarung (W02, M2): Methode 4 × Art 4 × Dienst 2 = 32. Vier Paare laufen, zwölf sind ODR-E6002, bei NurLesend alle 16
// ODR-E6006; kein Port wird gerufen.
func TestPaarung(t *testing.T) {
	arten := model.AlleRessourcenArten()
	if len(arten) != 4 {
		t.Fatalf("%d Arten, erwartet 4", len(arten))
	}
	n, laufen := 0, 0
	for _, nurLesend := range []bool{false, true} {
		for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
			for _, art := range arten {
				n++
				u := schreibUmg(t)
				if nurLesend {
					u.k.Dienste[0].Codecs = u.c.satz(true, "4.0")
				}
				var teile []model.SchluesselTeil
				id := "7"
				if art == model.RessourceEntity {
					teile, id = []model.SchluesselTeil{zahl("", "2")}, "2"
				}
				u.ressource(art, setProducts, model.Optionen{}, teile...)
				u.ei.ergebnis = eingabeIDName(id, "neu")
				a, err := u.rufe(anfrage(m, dienstRoot+"/x"))
				kontext := fmt.Sprintf("%s Art %d nurLesend=%v", m, art, nurLesend)
				laeuft := m == "POST" && art == model.RessourceMenge || m != "POST" && art == model.RessourceEntity
				switch {
				case nurLesend:
					pruefeFehler(t, kontext, err, e6006)
				case laeuft:
					laufen++
					erwarteStatus(t, kontext, a, err, sollStatus(m))
					continue
				default:
					pruefeFehler(t, kontext, err, e6002)
				}
				if u.p.anzahl("Workspaces.Get") != 0 || u.p.anzahl("Eingabe.LeseEntity") != 0 || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
					t.Errorf("%s: ein Port wurde gerufen", kontext)
				}
			}
		}
	}
	if n != 32 || laufen != 4 {
		t.Errorf("Menge %d (erwartet 32), laufen %d (erwartet 4)", n, laufen)
	}
}

// sollStatus ist der Status des Erfolgs nach SPEC-194: 201 bei POST, sonst 204.
func sollStatus(methode string) int {
	if methode == "POST" {
		return 201
	}
	return 204
}

// TestOptionenMutation (W03, M3): Option 7 × Methode 4 = 28. Ein $filter ist ODR-E6002 (vier Fehler), jede andere Option hat
// keine Wirkung: dieselbe Antwort wie ohne Option (24 Läufe).
func TestOptionenMutation(t *testing.T) {
	filter := filterOpt(vergleich(model.VergleichEq, "Id", model.Literal{Art: model.LiteralZahl, Text: "1"}))
	optionen := []struct {
		name string
		opt  model.Optionen
	}{
		{"$filter", model.Optionen{Filter: filter}},
		{"$select", model.Optionen{Select: &model.SelectOption{Namen: []string{"Name"}}}},
		{"$orderby", model.Optionen{OrderBy: []model.Sortierung{{Name: "Name", Absteigend: true}}}},
		{"$top", model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 0}}},
		{"$skip", model.Optionen{Skip: model.ZahlOption{Gesetzt: true, Wert: 5}}},
		{"$count", model.Optionen{Count: model.WahrOption{Gesetzt: true, Wert: true}}},
		{"Sonstige", model.Optionen{Sonstige: []model.SonstigeOption{{Name: "x", Wert: "y"}}}},
	}
	n, gleich, fehler := 0, 0, 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		ohne := schreibUmg(t)
		id := "2"
		if m == "POST" {
			id = "7"
		}
		a0, err0 := ohne.mutiere(m, id, eingabeIDName(id, "neu"))
		if err0 != nil {
			t.Fatalf("%s ohne Option: %v", m, err0)
		}
		for _, o := range optionen {
			n++
			u := schreibUmg(t)
			u.setzeMutation(m, id, eingabeIDName(id, "neu"))
			url := model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: setProducts, Schluessel: []model.SchluesselTeil{zahl("", id)}}, Optionen: o.opt}
			if m == "POST" {
				url.Ressource = model.Ressource{Art: model.RessourceMenge, EntitySet: setProducts}
			}
			u.c.url.f = func(string, string) (model.ODataURL, error) { return url, nil }
			a, err := u.rufe(anfrage(m, dienstRoot+"/x"))
			if o.name == "$filter" {
				fehler++
				pruefeFehler(t, m+" $filter", err, e6002)
				u.ohneSchreiben(m + " $filter")
				continue
			}
			if err != nil || a.Status != a0.Status || fmt.Sprint(a.Header) != fmt.Sprint(a0.Header) || u.p.anzahl("Entities.Put") != 1 {
				t.Errorf("%s %s: Status %d Fehler %v, ohne Option %d", m, o.name, a.Status, err, a0.Status)
				continue
			}
			gleich++
		}
	}
	if n != 28 || fehler != 4 || gleich != 24 {
		t.Errorf("Menge %d Fehler %d gleich %d, erwartet 28, 4, 24", n, fehler, gleich)
	}
}

// TestOhneEingabe (W04, M4): Quelle 3 × Methode 4 = 12. Ein Dienst ohne wirksames Eingabe beantwortet POST, PATCH und PUT mit
// ODR-E6002, DELETE läuft; ohne Satz gilt Konfig.Eingabe, ein Satz fällt nie auf sie zurück.
func TestOhneEingabe(t *testing.T) {
	n := 0
	for q, quelle := range []string{"ohne Satz, Konfig.Eingabe", "Satz mit Eingabe", "Satz ohne Eingabe bei gesetzter Konfig.Eingabe"} {
		for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
			n++
			u := schreibUmg(t)
			switch q {
			case 1:
				u.k.Dienste[0].Codecs = u.c.satz(false, "4.0")
			case 2:
				s := u.c.satz(false, "4.0")
				s.Eingabe = nil
				u.k.Dienste[0].Codecs = s
			}
			id := "2"
			if m == "POST" {
				id = "7"
			}
			a, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
			ohneLeser := q == 2 && m != "DELETE"
			if ohneLeser {
				pruefeFehler(t, quelle+" "+m, err, e6002)
				if u.p.anzahl("Eingabe.LeseEntity") != 0 || u.p.anzahl("Workspaces.Get") != 0 {
					t.Errorf("%s %s: ein Port wurde gerufen", quelle, m)
				}
				continue
			}
			erwarteStatus(t, quelle+" "+m, a, err, sollStatus(m))
		}
	}
	if n != 12 {
		t.Errorf("Menge %d, erwartet 12", n)
	}
}

// TestKonfigMutation (W05): nil Transaktion ist model.ErrUngueltig, nil Eingabe ist erlaubt, New ruft keinen Port.
func TestKonfigMutation(t *testing.T) {
	u := neueUmg(t)
	u.k.Transaktion = nil
	if _, err := mockNew(u); !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("nil Transaktion: %v, erwartet ErrUngueltig", err)
	}
	u = neueUmg(t)
	u.k.Eingabe = nil
	if m, err := mockNew(u); err != nil || m == nil {
		t.Errorf("nil Eingabe: %v", err)
	}
	if ops := u.p.ops(); len(ops) != 0 {
		t.Errorf("New hat Ports gerufen: %v", ops)
	}
}

// TestDeleteBody (W15): DELETE liest und schließt den Body nie und ruft den Leser nicht, auch mit Body.
func TestDeleteBody(t *testing.T) {
	for _, body := range []string{"", "{}", string(make([]byte, 1024))} {
		u := schreibUmg(t)
		u.entity(model.Optionen{}, "2")
		z := &zaehlStrom{inhalt: body}
		a := anfrage("DELETE", dienstRoot+"/x")
		a.Body, a.Laenge = z, int64(len(body))
		if body == "" {
			a.Body, a.Laenge = nil, -1
		}
		ant, err := u.rufe(a)
		erwarteStatus(t, fmt.Sprintf("DELETE Body %d", len(body)), ant, err, 204)
		if z.gelesen != 0 || z.geschlossen != 0 || u.p.anzahl("Eingabe.LeseEntity") != 0 {
			t.Errorf("Body %d: gelesen %d geschlossen %d Leser %d, erwartet 0", len(body), z.gelesen, z.geschlossen, u.p.anzahl("Eingabe.LeseEntity"))
		}
	}
}

// TestPostSchluesselLeer (W17): ein leerer Schluessel des Lesers bei POST ist ODR-E9001 ohne Ursache, vor der Transaktion.
func TestPostSchluesselLeer(t *testing.T) {
	for _, props := range [][]model.PropertyWert{nil, {pw("Name", model.ArtString, "x")}, {pw("Id", model.ArtInt32, "7")}} {
		u := schreibUmg(t)
		_, err := u.mutiere("POST", "7", model.EntityEingabe{Typ: typProduct, Schluessel: "", Properties: props})
		f := erwarteFehler(t, "leerer Schluessel", err, e9001)
		if f != nil && f.Ursache != nil || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
			t.Errorf("Ursache %v oder Transaktion gerufen", f)
		}
	}
}

// TestSchluesselUrlMutation (W18): ein ungültiger Schlüssel in der URL ist ODR-E6005 vor dem Lesen des Bodys, bei PATCH, PUT und
// DELETE; Leser 0, Transaktion 0. Verstöße: Name falsch, Teil fehlt, Teil doppelt, Literal vom falschen Typ, Zahl außerhalb des
// Bereichs von Edm.Int32.
func TestSchluesselUrlMutation(t *testing.T) {
	zeichenkette := model.SchluesselTeil{Literal: model.Literal{Art: model.LiteralZeichenkette, Text: "'x'"}}
	verstoesse := [][]model.SchluesselTeil{
		{zahl("Falsch", "1")}, nil, {zahl("Id", "1"), zahl("Id", "2")}, {zeichenkette}, {zahl("", "2147483648")}, {zahl("", "-2147483649")},
		{zahl("", "1.5")}, {zahl("Id", "1"), zahl("Name", "2")},
	}
	n := 0
	for _, m := range []string{"PATCH", "PUT", "DELETE"} {
		for _, teile := range verstoesse {
			n++
			u := schreibUmg(t)
			u.ressource(model.RessourceEntity, setProducts, model.Optionen{}, teile...)
			u.ei.ergebnis = eingabeIDName("2", "neu")
			_, err := u.rufe(anfrage(m, dienstRoot+"/x"))
			pruefeFehler(t, fmt.Sprintf("%s %v", m, teile), err, e6005)
			if u.p.anzahl("Eingabe.LeseEntity") != 0 || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
				t.Errorf("%s %v: Leser oder Transaktion gerufen", m, teile)
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
}

// rangCfg sammelt die Verstöße eines Laufs der Rangfolge (W06, M5).
type rangCfg struct {
	laenge                    int64
	pfadFremd, ohneHeader     bool
	urlFehler, filter         bool
	wsUnbekannt, wsFremd      bool
	ohneMeta, setUnbekannt    bool
	schluesselBad, bodyFehler bool
	fehlendeEntity            bool
}

// rangStufen sind die zwölf Stufen der Rangfolge (SPEC-192 Nr. 2) mit dem Code der Stufe (0: Fehler ohne Code) und dem Port, den
// erst eine spätere Stufe ruft.
func rangStufen() []struct {
	name string
	code model.Code
	set  func(*rangCfg)
	port string
} {
	return []struct {
		name string
		code model.Code
		set  func(*rangCfg)
		port string
	}{
		{"Argumente", 0, func(c *rangCfg) { c.laenge = -5 }, ""},
		{"kein Dienst", model.CodeE5005, func(c *rangCfg) { c.pfadFremd = true }, ""},
		{"Header fehlt", model.CodeE5003, func(c *rangCfg) { c.ohneHeader = true }, ""},
		{"URL ungültig", e6005, func(c *rangCfg) { c.urlFehler = true }, "URL.ParseURL"},
		{"Methode und Ressource", e6002, func(c *rangCfg) { c.filter = true }, ""},
		{"Workspace unbekannt", e7001, func(c *rangCfg) { c.wsUnbekannt = true }, "Workspaces.Get"},
		{"Workspace fremd", model.CodeE5004, func(c *rangCfg) { c.wsFremd = true }, ""},
		{"Service ohne Metadaten", model.CodeE6004, func(c *rangCfg) { c.ohneMeta = true }, "Metadaten.Get"},
		{"Set unbekannt", e6008, func(c *rangCfg) { c.setUnbekannt = true }, "Blobs.Open"},
		{"Schlüssel der URL", e6005, func(c *rangCfg) { c.schluesselBad = true }, ""},
		{"Body", e6007, func(c *rangCfg) { c.bodyFehler = true }, "Eingabe.LeseEntity"},
		{"Zustand", e6008, func(c *rangCfg) { c.fehlendeEntity = true }, "Transaktion.WithWorkspaceTransaction"},
	}
}

// TestRangfolgeMutation (W06, M5): Stufe 12 → 66 Paare; je Paar sind beide Verstöße gesetzt, und der Code ist der der früheren
// Stufe; kein Port der späteren Stufe wird gerufen. Die nicht baubaren Paare stehen mit Grund in einer Liste.
func TestRangfolgeMutation(t *testing.T) {
	stufen := rangStufen()
	nichtBaubar := map[[2]int]string{
		{2, 5}: "ohne Header gibt es keinen Workspace, der unbekannt wäre",
		{2, 6}: "ohne Header gibt es keinen Workspace, der fremd wäre",
		{5, 6}: "ein Workspace ist nicht zugleich unbekannt und fremd",
	}
	baubar, nicht := 0, 0
	for i := range stufen {
		for j := i + 1; j < len(stufen); j++ {
			if _, ok := nichtBaubar[[2]int{i, j}]; ok {
				nicht++
				continue
			}
			baubar++
			u := schreibUmg(t)
			var c rangCfg
			c.laenge = -1
			stufen[i].set(&c)
			stufen[j].set(&c)
			a := anfrage("PATCH", dienstRoot+"/x")
			a.Laenge = c.laenge
			if c.pfadFremd {
				a.Pfad = "/anderer/x"
			}
			if c.ohneHeader {
				a.Header = map[string][]string{}
			}
			if c.wsUnbekannt {
				a.Header["X-Recorder-Workspace"] = []string{"unbekannt"}
			}
			if c.wsFremd {
				u.ws.setze(model.Workspace{ID: wsID, ServiceID: "anderer", BaselineSnapshotID: baselineID})
			}
			if c.ohneMeta {
				u.md.setze(model.Service{ID: dienstID})
			}
			teile := []model.SchluesselTeil{zahl("", "2")}
			if c.fehlendeEntity {
				teile = []model.SchluesselTeil{zahl("", "99")}
			}
			if c.schluesselBad {
				teile = []model.SchluesselTeil{zahl("Falsch", "2")}
			}
			set := setProducts
			if c.setUnbekannt {
				set = "Unbekannt"
			}
			var opt model.Optionen
			if c.filter {
				opt.Filter = idGroesser("1")
			}
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				if c.urlFehler {
					return model.ODataURL{}, model.ErrURLUngueltig
				}
				return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: set, Schluessel: teile}, Optionen: opt}, nil
			}
			u.ei.ergebnis = eingabeIDName(teile[0].Literal.Text, "neu")
			if c.bodyFehler {
				u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: model.EingabeSyntax})
			}
			_, err := u.rufe(a)
			kontext := stufen[i].name + " vor " + stufen[j].name
			if stufen[i].code == 0 {
				if !errors.Is(err, model.ErrUngueltig) || fehlercode(err) != 0 {
					t.Errorf("%s: %v, erwartet ErrUngueltig ohne Code", kontext, err)
				}
			} else {
				pruefeFehler(t, kontext, err, stufen[i].code)
			}
			for k := i + 1; k < len(stufen); k++ {
				if p := stufen[k].port; p != "" && u.p.anzahl(p) != 0 {
					t.Errorf("%s: %s wurde gerufen", kontext, p)
				}
			}
			u.ohneSchreiben(kontext)
		}
	}
	if baubar+nicht != 66 || nicht != len(nichtBaubar) || baubar != 63 {
		t.Errorf("baubar %d nicht baubar %d, erwartet 63 und 3", baubar, nicht)
	}
}

// zaehlStrom ist ein Body, der gelesene Bytes und Close zählt.
type zaehlStrom struct {
	inhalt      string
	gelesen     int
	geschlossen int
}

func (z *zaehlStrom) Read(p []byte) (int, error) {
	n := copy(p, z.inhalt[z.gelesen:])
	z.gelesen += n
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (z *zaehlStrom) Close() error {
	z.geschlossen++
	return nil
}

// mockNew baut den Dienst aus der Konfig des Aufbaus.
func mockNew(u *umg) (*mock.Mocker, error) { return mock.New(u.k) }

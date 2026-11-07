package mock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// folgeVorTx sind die Operationen vor der Transaktion; folgeInTx die in ihr (SPEC-192 Nr. 6), von Hand geschrieben.
func folgeVorTx(m string) []string {
	r := []string{"URL.ParseURL", "Workspaces.Get", "Metadaten.Get", "Blobs.Open", "MetadatenCodec.ParseMetadata"}
	if m != "DELETE" {
		r = append(r, "Eingabe.LeseEntity")
	}
	return append(r, "Transaktion.WithWorkspaceTransaction")
}

func folgeInTx(m string) []string {
	r := []string{"Workspaces.Get", "Entities.Get", "Entities.Put", "Workspaces.Advance"}
	if m == "POST" {
		return append(r, "Schreiber.SchreibeEntity", "Schreiber.Adresse")
	}
	return append(r, "Schreiber.SchreibeAnzahl")
}

// TestAufrufzahlenUndFolge (W46, W47): die Folge der Aufrufe jeder Mutation mit Erfolg, jede Operation jedes Fakes gezählt
// (eine gemeinsame Liste); bei einem Fehler vor der Transaktion läuft nur ein Anfang der Folge, bei einem Fehler in ihr endet
// sie vor Put (oder nach Put und Advance beim Fehler der Antwort). Methode 4 × Ausgang 3 = 12.
func TestAufrufzahlenUndFolge(t *testing.T) {
	n := 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		id := schluesselFuer(m)
		// Erfolg: genau die Folge, keine andere Operation.
		u := schreibUmg(t)
		a, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
		erwarteStatus(t, m+" Erfolg", a, err, sollStatus(m))
		if want := append(folgeVorTx(m), folgeInTx(m)...); !slices.Equal(u.p.ops(), want) {
			t.Errorf("%s Erfolg: Folge %v, erwartet %v", m, u.p.ops(), want)
		}
		// Das ctx des Aufrufs geht an jeden Port als dieselbe Instanz (nicht nur mit demselben Wert), S2 bis S5 und S17 bis S26.
		for _, e := range u.p.ev {
			if e.ctx != nil && e.ctx != u.ctx {
				t.Errorf("%s Erfolg: %s erhielt ein anderes ctx", m, e.op)
			}
		}
		n++
		// Fehler vor der Transaktion: unbekanntes Set.
		u = schreibUmg(t)
		u.ressource(model.RessourceEntity, "Unbekannt", model.Optionen{}, zahl("", id))
		if m == "POST" {
			u.ressource(model.RessourceMenge, "Unbekannt", model.Optionen{})
		}
		_, err = u.rufe(anfrage(m, dienstRoot+"/x"))
		pruefeFehler(t, m+" Set unbekannt", err, e6008)
		if want := []string{"URL.ParseURL", "Workspaces.Get", "Metadaten.Get", "Blobs.Open", "MetadatenCodec.ParseMetadata"}; !slices.Equal(u.p.ops(), want) {
			t.Errorf("%s Set unbekannt: Folge %v", m, u.p.ops())
		}
		n++
		// Fehler in der Transaktion: der Zustand des Eintrags passt nicht (POST: vorhanden, sonst fehlend bei Coverage complete).
		u = schreibUmg(t)
		key := "2"
		if m != "POST" {
			key = "99"
		}
		_, err = u.mutiere(m, key, eingabeIDName(key, "neu"))
		pruefeFehler(t, m+" Zustand", err, map[bool]model.Code{true: e6009, false: e6008}[m == "POST"])
		wantTx := []string{"Workspaces.Get", "Entities.Get"}
		if m != "POST" {
			wantTx = append(wantTx, "Coverage.Get")
		}
		if want := append(folgeVorTx(m), wantTx...); !slices.Equal(u.p.ops(), want) {
			t.Errorf("%s Zustand: Folge %v, erwartet %v", m, u.p.ops(), want)
		}
		n++
	}
	if n != 12 {
		t.Errorf("Menge %d, erwartet 12", n)
	}
}

// TestKeineVerbotenenOperationen (W46): die Operationen, die der Dienst nie rufen darf, bleiben bei jeder Mutation 0; die Liste
// der Operationen kommt aus den Methoden der Interfaces, nicht aus einer festen Zahl.
func TestKeineVerbotenenOperationen(t *testing.T) {
	verboten := []string{"Cursors.Put", "Cursors.Get", "Cursors.List", "Cursors.ResetAll", "Snapshots.Get", "Snapshots.List", "Snapshots.CreateWorkspace",
		"Snapshots.GetWorkspace", "Snapshots.ListWorkspace", "Entities.List", "Entities.ListSet", "Entities.ReplaceAll", "Workspaces.Create", "Workspaces.List",
		"Workspaces.Delete", "Coverage.List", "Schreiber.SchreibeCollection", "Schreiber.SchreibeFehler", "Blobs.Begin", "Blobs.Verify", "Blobs.List", "Blobs.Remove",
		"Metadaten.Create", "Metadaten.List", "Metadaten.SetMetadata"}
	// Zählprüfung gegen die Methoden der Interfaces: jede Methode ist erlaubt oder verboten genannt.
	erlaubt := []string{"Workspaces.Get", "Workspaces.Advance", "Entities.Get", "Entities.Put", "Coverage.Get", "Schreiber.SchreibeEntity",
		"Schreiber.SchreibeAnzahl", "Schreiber.Adresse", "Blobs.Open", "Metadaten.Get"}
	gesamt := anz[outbound.WorkspaceRepository]() + anz[outbound.EntityRepository]() + anz[outbound.CursorRepository]() +
		anz[outbound.SnapshotRepository]() + anz[outbound.CoverageRepository]() + anz[outbound.ODataSchreiber]() +
		anz[outbound.BlobRepository]() + anz[outbound.MetadataRepository]()
	if gesamt != len(erlaubt)+len(verboten) {
		t.Fatalf("%d Methoden der Interfaces, %d erlaubt und %d verboten genannt", gesamt, len(erlaubt), len(verboten))
	}
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		u := schreibUmg(t)
		id := schluesselFuer(m)
		if _, err := u.mutiere(m, id, eingabeIDName(id, "neu")); err != nil {
			t.Fatal(err)
		}
		for _, op := range verboten {
			if u.p.anzahl(op) != 0 {
				t.Errorf("%s: %s %d Mal gerufen, erwartet 0", m, op, u.p.anzahl(op))
			}
		}
	}
}

// anz zählt die Methoden eines Interfaces.
func anz[T any]() int { return reflect.TypeFor[T]().NumMethod() }

// TestPostAntwort (W51): 201 mit dem Body von SchreibeEntity, Content-Type, Kennzeichen und Location als genau drei Header;
// Wurzel aus PublicBaseURL und Root, Auswahl nil, Accept unverändert.
func TestPostAntwort(t *testing.T) {
	accepts := [][]string{nil, {"application/json"}, {"a/1", "b/2"}}
	for _, acc := range accepts {
		u := schreibUmg(t)
		u.c.sch.kopf = model.Darstellung{ContentType: "fremd/typ", Kennzeichen: model.Kopfzeile{Name: "X-Frei", Wert: "wert-9"}}
		u.c.sch.ort = "http://ort.test/anders"
		u.setzeMutation("POST", "7", eingabeIDName("7", "neu"))
		a := anfrage("POST", dienstRoot+"/x")
		if acc != nil {
			a.Header["Accept"] = acc
		}
		ant, err := u.rufe(a)
		erwarteStatus(t, "POST", ant, err, 201)
		want := map[string][]string{"Content-Type": {"fremd/typ"}, "X-Frei": {"wert-9"}, "Location": {"http://ort.test/anders"}}
		if fmt.Sprint(ant.Header) != fmt.Sprint(want) {
			t.Errorf("Header %v, erwartet %v", ant.Header, want)
		}
		if got := string(lies(t, ant)); !strings.HasPrefix(got, "Schreiber.SchreibeEntity:PROBE-") {
			t.Errorf("Body %q", got)
		}
		ea := u.letzteEntity()
		if ea.Wurzel != "http://recorder.test/svc" || ea.Set != setProducts || ea.Entity.Schluessel != "7" || ea.Auswahl != nil || !slices.Equal(ea.Accept, acc) {
			t.Errorf("EntityAntwort %+v", ea)
		}
		ad := u.p.von("Schreiber.Adresse")
		if len(ad) != 1 || ad[0].args[1] != "http://recorder.test/svc" || ad[0].args[2] != setProducts || ad[0].args[3] != "7" {
			t.Errorf("Adresse %v", ad)
		}
	}
}

// TestAntwort204 (W53): PATCH, PUT, DELETE → 204 mit leerem Strom (nie nil), Header nur das Kennzeichen; ContentType und Body der
// Probe gehen nie in die Antwort. Methode 3 × Darstellung 2 = 6.
func TestAntwort204(t *testing.T) {
	for _, m := range []string{"PATCH", "PUT", "DELETE"} {
		for _, dar := range []model.Darstellung{
			{ContentType: "text/plain", Kennzeichen: model.Kopfzeile{Name: "X-A", Wert: "1"}},
			{ContentType: "anderer/typ", Kennzeichen: model.Kopfzeile{Name: "X-B", Wert: "2"}},
		} {
			u := schreibUmg(t)
			u.c.sch.kopf = dar
			ant, err := u.mutiere(m, "2", eingabeIDName("2", "neu"))
			erwarteStatus(t, m, ant, err, 204)
			want := map[string][]string{dar.Kennzeichen.Name: {dar.Kennzeichen.Wert}}
			if ant.Body == nil || fmt.Sprint(ant.Header) != fmt.Sprint(want) {
				t.Errorf("%s: Body %v Header %v, erwartet %v", m, ant.Body, ant.Header, want)
			}
			if b, _ := io.ReadAll(ant.Body); len(b) != 0 {
				t.Errorf("%s: Body %q", m, b)
			}
		}
	}
}

// TestDarstellungLeerMutation (W54, M21 Teil): ein leeres Kennzeichen (Name oder Wert), ein leerer ContentType oder alles leer ist
// ODR-E9001 ohne Antwort; die Transaktion rollt zurück. Verstoß 4 × Antwort 2 (POST mit SchreibeEntity, PATCH mit SchreibeAnzahl).
func TestDarstellungLeerMutation(t *testing.T) {
	leer := []model.Darstellung{
		{ContentType: "text/plain", Kennzeichen: model.Kopfzeile{Name: "", Wert: "1"}},
		{ContentType: "text/plain", Kennzeichen: model.Kopfzeile{Name: "X-A", Wert: ""}},
		{ContentType: "", Kennzeichen: model.Kopfzeile{Name: "X-A", Wert: "1"}},
		{},
	}
	n := 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		for _, dar := range leer {
			n++
			u := schreibUmg(t)
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			u.c.sch.kopf = dar
			id := schluesselFuer(m)
			ant, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
			f := erwarteFehler(t, fmt.Sprintf("%s %+v", m, dar), err, e9001)
			if f != nil && (f.Ursache != nil || ant.Body != nil) {
				t.Errorf("%s: Ursache %v Antwort %+v", m, f.Ursache, ant)
			}
			u.unveraendert(vorher, rev, m)
		}
	}
	if n != 16 {
		t.Errorf("Menge %d, erwartet 16", n)
	}
}

// TestPreferIgnoriert (W55, M27): Prefer und If-Match werden nie gelesen: gleiche Antwort wie ohne Header. Wert 5 × Methode 3 = 15.
func TestPreferIgnoriert(t *testing.T) {
	werte := []map[string][]string{
		{"Prefer": {"return=minimal"}}, {"Prefer": {"return=representation"}}, {"Prefer": {"gibt es nicht"}},
		{"Prefer": {"return=minimal", "return=representation"}}, {"If-Match": {"*"}},
	}
	n := 0
	for _, m := range dreiMethoden() {
		ohne := schreibUmg(t)
		id := schluesselFuer(m)
		a0, err := ohne.mutiere(m, id, eingabeIDName(id, "neu"))
		if err != nil {
			t.Fatal(err)
		}
		b0 := lies(t, a0)
		for _, h := range werte {
			n++
			u := schreibUmg(t)
			u.setzeMutation(m, id, eingabeIDName(id, "neu"))
			a := anfrage(m, dienstRoot+"/x")
			for k, v := range h {
				a.Header[k] = v
			}
			ant, err := u.rufe(a)
			if err != nil || ant.Status != a0.Status || fmt.Sprint(ant.Header) != fmt.Sprint(a0.Header) || !slices.Equal(lies(t, ant), b0) {
				t.Errorf("%s %v: Status %d Fehler %v", m, h, ant.Status, err)
			}
		}
	}
	if n != 15 {
		t.Errorf("Menge %d, erwartet 15", n)
	}
}

// TestWiederholungMutation (W56, LH-QA-02): dieselbe Folge auf zwei gleich gestarteten Aufbauten ergibt dieselben Bytes, Header und
// denselben Endzustand; ein Payload mit neun Properties streute bei einer Iteration über eine Map. 50 Läufe.
func TestWiederholungMutation(t *testing.T) {
	folge := func() (string, map[string]model.EntityState) {
		u := schreibUmg(t)
		var protokoll []string
		for _, s := range []struct{ m, id string }{{"POST", "7"}, {"PATCH", "7"}, {"PUT", "2"}, {"DELETE", "3"}} {
			e := model.EntityEingabe{Typ: typProduct, Schluessel: s.id, Properties: append([]model.PropertyWert{pw("Id", model.ArtInt32, s.id)}, neueWerte()...)}
			ant, err := u.mutiere(s.m, s.id, e)
			protokoll = append(protokoll, fmt.Sprint(ant.Status, err, ant.Header))
			if ant.Body != nil {
				b, _ := io.ReadAll(ant.Body)
				protokoll = append(protokoll, string(b))
			}
		}
		return strings.Join(protokoll, "|"), u.zustand()
	}
	p0, z0 := folge()
	for i := 0; i < 50; i++ {
		p, z := folge()
		if p != p0 || len(z) != len(z0) {
			t.Fatalf("Lauf %d: Antworten weichen ab", i)
		}
		for k, x := range z0 {
			if !gleichEintrag(x, z[k]) {
				t.Fatalf("Lauf %d: Eintrag %s weicht ab", i, k)
			}
		}
	}
}

// bekannteFehler sind die sechs Fehler der Menge M19, Orakel von Hand aus SPEC-109 Nr. 5.
func bekannteFehler() []struct {
	err       error
	workspace model.Code
	sonst     model.Code
} {
	return []struct {
		err       error
		workspace model.Code
		sonst     model.Code
	}{
		{outbound.ErrNichtGefunden, model.CodeE7001, e9001},
		{outbound.ErrBereitsVorhanden, model.CodeE7007, e9001},
		{outbound.ErrBeschaedigt, model.CodeE3001, model.CodeE3001},
		{outbound.ErrBlobFehlt, model.CodeE3001, model.CodeE3001},
		{model.ErrUngueltig, e9001, e9001},
		{errors.New("roh " + kanarienwert), e9001, e9001},
	}
}

// stelle ist eine Aufrufstelle der Mutation (SPEC-195 Nr. 11): die Operation im Protokoll, ihr Aufruf, die Methode, die sie
// erreicht, und der Gegenstand (Workspace oder Sonstiges).
type stelle struct {
	name      string
	op        string
	nr        int
	methode   string
	workspace bool
	fehlend   bool
}

func stellen() []stelle {
	return []stelle{
		{"S17", "Eingabe.LeseEntity", 1, "PATCH", false, false},
		{"S18", "Transaktion.WithWorkspaceTransaction", 1, "PATCH", false, false},
		{"S19", "Workspaces.Get", 2, "PATCH", true, false},
		{"S20", "Entities.Get", 1, "PATCH", true, false},
		{"S21", "Coverage.Get", 1, "PATCH", false, true},
		{"S22", "Entities.Put", 1, "PATCH", true, false},
		{"S23", "Workspaces.Advance", 1, "PATCH", true, false},
		{"S24", "Schreiber.SchreibeEntity", 1, "POST", false, false},
		{"S25", "Schreiber.Adresse", 1, "POST", false, false},
		{"S26", "Schreiber.SchreibeAnzahl", 1, "PATCH", false, false},
	}
}

// laufStelle bereitet den Lauf einer Stelle vor und liefert Aufbau und Aufruf.
func laufStelle(t *testing.T, s stelle) (*umg, string) {
	u := schreibUmg(t)
	id := "2"
	switch {
	case s.fehlend:
		id = "99"
	case s.methode == "POST":
		id = "7"
	}
	return u, id
}

// TestPortFehlerMutation (W57, M19): Stelle 10 × Fehler 6 × Form 2 = 120, dazu S17 mit den beiden Klassen der Eingabe × 2 und S23 mit
// ErrRevisionVeraltet × 2 = 126; Gegenstand je Stelle nach SPEC-195 Nr. 11; ErrNichtGefunden an S20 und S21 ist kein Fehler.
func TestPortFehlerMutation(t *testing.T) {
	n := 0
	for _, s := range stellen() {
		for _, f := range bekannteFehler() {
			for _, verpackt := range []bool{false, true} {
				n++
				u, id := laufStelle(t, s)
				vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
				err := f.err
				if verpackt {
					err = fmt.Errorf("Port: %w", f.err)
				}
				u.p.setzeFehlerBeiNr(s.op, s.nr, err)
				_, got := u.mutiere(s.methode, id, eingabeIDName(id, "neu"))
				kontext := fmt.Sprintf("%s %v verpackt=%v", s.name, f.err, verpackt)
				want := f.sonst
				if s.workspace {
					want = f.workspace
				}
				switch {
				case errors.Is(f.err, outbound.ErrNichtGefunden) && s.name == "S20":
					want = e6008 // fehlend bei Coverage complete: der Zustand, kein Fehler
				case errors.Is(f.err, outbound.ErrNichtGefunden) && s.name == "S21":
					want = e6003 // Coverage unknown, kein Fehler
				}
				fe := erwarteFehler(t, kontext, got, want)
				zustandStattFehler := errors.Is(f.err, outbound.ErrNichtGefunden) && (s.name == "S20" || s.name == "S21")
				if fe != nil && !zustandStattFehler && !errors.Is(fe, f.err) {
					t.Errorf("%s: Ursache nicht erreichbar", kontext)
				}
				u.unveraendert(vorher, rev, kontext)
			}
		}
	}
	// S17 mit den Klassen der Eingabe, S23 mit ErrRevisionVeraltet.
	for _, verpackt := range []bool{false, true} {
		for _, c := range []struct {
			ur   model.EingabeUrsache
			code model.Code
		}{{model.EingabeSyntax, e6007}, {model.EingabeNavigation, e6002}} {
			n++
			u := schreibUmg(t)
			err := error(&model.EingabeFehler{Ursache: c.ur})
			if verpackt {
				err = fmt.Errorf("Port: %w", err)
			}
			u.p.setzeFehlerBeiNr("Eingabe.LeseEntity", 1, err)
			_, got := u.mutiere("PATCH", "2", eingabeIDName("2", "neu"))
			pruefeFehler(t, "S17 "+string(c.ur), got, c.code)
		}
		n++
		u := schreibUmg(t)
		err := outbound.ErrRevisionVeraltet
		if verpackt {
			err = fmt.Errorf("Port: %w", err)
		}
		u.p.setzeFehlerBeiNr("Workspaces.Advance", 1, err)
		_, got := u.mutiere("PATCH", "2", eingabeIDName("2", "neu"))
		pruefeFehler(t, "S23 ErrRevisionVeraltet", got, e9001)
	}
	if n != 126 {
		t.Errorf("Menge %d, erwartet 126", n)
	}
}

// kontextStelle ist ein Paar (Methode, Stelle) der Menge M20.
type kontextStelle struct {
	m, op string
	nr    int
}

// kontextStellen sind die 46 Paare der Menge M20, von Hand: S2 bis S5 (4 Methoden), S17 (3), S18 bis S20 (4 Methoden), S21 (3), S22
// und S23 (4), S24 (1), S26 (3).
func kontextStellen() []kontextStelle {
	var r []kontextStelle
	vier := []string{"POST", "PATCH", "PUT", "DELETE"}
	for _, m := range vier {
		for _, o := range []struct {
			op string
			nr int
		}{{"Workspaces.Get", 1}, {"Metadaten.Get", 1}, {"Blobs.Open", 1}, {"MetadatenCodec.ParseMetadata", 1}} {
			r = append(r, kontextStelle{m, o.op, o.nr})
		}
	}
	for _, m := range dreiMethoden() {
		r = append(r, kontextStelle{m, "Eingabe.LeseEntity", 1})
	}
	for _, m := range vier {
		r = append(r, kontextStelle{m, "Transaktion.WithWorkspaceTransaction", 1}, kontextStelle{m, "Workspaces.Get", 2}, kontextStelle{m, "Entities.Get", 1})
	}
	for _, m := range []string{"PATCH", "PUT", "DELETE"} {
		r = append(r, kontextStelle{m, "Coverage.Get", 1})
	}
	for _, m := range vier {
		r = append(r, kontextStelle{m, "Entities.Put", 1}, kontextStelle{m, "Workspaces.Advance", 1})
	}
	r = append(r, kontextStelle{"POST", "Schreiber.SchreibeEntity", 1})
	for _, m := range []string{"PATCH", "PUT", "DELETE"} {
		r = append(r, kontextStelle{m, "Schreiber.SchreibeAnzahl", 1})
	}
	return r
}

// TestKontextMutation (W58, M20): an jeder Aufrufstelle erhält der Port das ctx des Aufrufs unverändert, und ein Kontextfehler
// kehrt ohne Code zurück, nackt, verpackt und in einem errors.Join neben einem Fehler der Tabelle (vorn und hinten); danach läuft
// kein weiterer Port, der Zustand bleibt unverändert. Form 4 × Paare 46 = 184.
func TestKontextMutation(t *testing.T) {
	paare := kontextStellen()
	if len(paare) != 46 {
		t.Fatalf("%d Paare, erwartet 46", len(paare))
	}
	n := 0
	for _, p := range paare {
		for form := 0; form < 4; form++ {
			n++
			u := schreibUmg(t)
			id := schluesselFuer(p.m)
			if p.op == "Coverage.Get" {
				id = "99"
			}
			ke := context.Canceled
			err := error(ke)
			switch form {
			case 1:
				err = fmt.Errorf("Port: %w", ke)
			case 2:
				err = errors.Join(ke, outbound.ErrNichtGefunden)
			case 3:
				err = errors.Join(outbound.ErrNichtGefunden, ke)
			}
			u.p.setzeFehlerBeiNr(p.op, p.nr, err)
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			_, got := u.mutiere(p.m, id, eingabeIDName(id, "neu"))
			kontext := fmt.Sprintf("%s %s#%d Form %d", p.m, p.op, p.nr, form)
			if got != err || fehlercode(got) != 0 {
				t.Errorf("%s: Fehler %v mit Code %v, erwartet derselbe Fehler ohne Code", kontext, got, fehlercode(got))
			}
			ev := u.p.von(p.op)
			if len(ev) != p.nr {
				t.Errorf("%s: %d Aufrufe", kontext, len(ev))
			}
			for _, e := range u.p.ev {
				if e.ctx != nil && e.ctx != u.ctx {
					t.Errorf("%s: %s ohne das ctx des Aufrufs", kontext, e.op)
				}
			}
			// Nach dem Fehler läuft kein weiterer Port der Folge.
			ops := u.p.ops()
			if ops[len(ops)-1] != p.op {
				t.Errorf("%s: nach dem Fehler lief noch %v", kontext, ops[len(ops)-1])
			}
			u.unveraendert(vorher, rev, kontext)
		}
	}
	if n != 184 {
		t.Errorf("Menge %d, erwartet 184", n)
	}
}

// TestTransaktionFehler (W59): ein *inbound.Fehler und ein Kontextfehler aus der Transaktion kehren unverändert zurück, jeder
// andere Fehler (Beginn, Bestätigung) wird Sonstiges; eine Panik in f wird weitergereicht und lässt den Zustand unverändert.
func TestTransaktionFehler(t *testing.T) {
	eigener := &inbound.Fehler{Code: e7001, Status: 404}
	roh := errors.New("roh")
	faelle := []struct {
		name   string
		nachF  error
		beginn error
		want   error
		code   model.Code
	}{
		{"Beginn", nil, roh, nil, e9001},
		{"Bestätigung", roh, nil, nil, e9001},
		{"*Fehler", eigener, nil, eigener, e7001},
		{"Kontextfehler", context.DeadlineExceeded, nil, context.DeadlineExceeded, 0},
		{"Join mit Kontext", errors.Join(roh, context.Canceled), nil, nil, 0},
		{"ErrUngueltig", model.ErrUngueltig, nil, nil, e9001},
	}
	for _, f := range faelle {
		u := schreibUmg(t)
		vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
		u.tx.nachF = f.nachF
		if f.beginn != nil {
			u.p.setzeFehler("Transaktion.WithWorkspaceTransaction", f.beginn)
		}
		_, err := u.mutiere("PATCH", "2", eingabeIDName("2", "neu"))
		switch {
		case f.code == 0:
			if fehlercode(err) != 0 || !errors.Is(err, f.nachF) {
				t.Errorf("%s: %v", f.name, err)
			}
		case f.want != nil:
			if err != f.want {
				t.Errorf("%s: %v nicht unverändert", f.name, err)
			}
		default:
			pruefeFehler(t, f.name, err, f.code)
		}
		u.unveraendert(vorher, rev, f.name)
	}
	// Panik in f: wird weitergereicht, der Zustand bleibt unverändert.
	u := schreibUmg(t)
	vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
	u.p.setzePanik("Workspaces.Advance")
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("keine Panik weitergereicht")
			}
		}()
		_, _ = u.mutiere("PATCH", "2", eingabeIDName("2", "neu"))
	}()
	u.unveraendert(vorher, rev, "Panik")
}

// TestRollback (W60, M21): scheitert ein Aufruf nach Put (Advance, SchreibeEntity, Adresse, SchreibeAnzahl), rollt die Transaktion
// Put und Advance zurück; (Fehlerstelle, Methode) 9 × Fehler 2 (roh, Kontextfehler) = 18.
func TestRollback(t *testing.T) {
	paare := []struct{ op, m string }{
		{"Workspaces.Advance", "POST"}, {"Workspaces.Advance", "PATCH"}, {"Workspaces.Advance", "PUT"}, {"Workspaces.Advance", "DELETE"},
		{"Schreiber.SchreibeEntity", "POST"}, {"Schreiber.Adresse", "POST"},
		{"Schreiber.SchreibeAnzahl", "PATCH"}, {"Schreiber.SchreibeAnzahl", "PUT"}, {"Schreiber.SchreibeAnzahl", "DELETE"},
	}
	n := 0
	for _, p := range paare {
		for _, ke := range []bool{false, true} {
			n++
			u := schreibUmg(t)
			vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
			err := errors.New("roh " + kanarienwert)
			if ke {
				err = context.Canceled
			}
			u.p.setzeFehler(p.op, err)
			id := schluesselFuer(p.m)
			ant, got := u.mutiere(p.m, id, eingabeIDName(id, "neu"))
			if ke {
				if got != err {
					t.Errorf("%s %s: %v", p.m, p.op, got)
				}
			} else {
				pruefeFehler(t, p.m+" "+p.op, got, e9001)
			}
			if ant.Body != nil || u.p.anzahl("Entities.Put") != 1 {
				t.Errorf("%s %s: Antwort oder Put: %+v Put %d", p.m, p.op, ant, u.p.anzahl("Entities.Put"))
			}
			u.unveraendert(vorher, rev, p.m+" "+p.op)
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}

// TestFehlerWertMutation (W62, geprüft ist weniger als die Zeile sagt): ein Auslöser, ODR-E6007 aus einem Fehler des Lesers, mit dem
// Canary im Ort der Ursache, im Wert der Property Name und im Content-Type; sein Error() nennt den Canary nicht. Dazu der Text von
// Error() für fünf Codes (ODR-E6009, ODR-E6007, ODR-E6008, ODR-E6003, ODR-E6002) gegen Code und message. Die übrigen erreichbaren
// Codes und Canary im Schlüssel stehen in TestFehlerWert des Lesens (ein Fall je Fehlerzeile) und sind hier nicht geprüft.
func TestFehlerWertMutation(t *testing.T) {
	canary := kanarienwert
	u := schreibUmg(t)
	u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: model.EingabeUnbekannteProperty, Ort: "/" + canary})
	u.setzeMutation("PATCH", "2", eingabeIDName("2", canary))
	a := anfrage("PATCH", dienstRoot+"/x")
	a.Header["Content-Type"] = []string{"x/" + canary}
	_, err := u.rufe(a)
	f := erwarteFehler(t, "Canary", err, e6007)
	if f != nil && strings.Contains(f.Error(), canary) {
		t.Errorf("Error() nennt den Canary: %q", f.Error())
	}
	for _, c := range []model.Code{e6009, e6007, e6008, e6003, e6002} {
		if got, want := mitCodeText(c), c.String()+": "+c.Message(); got != want {
			t.Errorf("%v: %q, erwartet %q", c, got, want)
		}
	}
}

func mitCodeText(c model.Code) string {
	return (&inbound.Fehler{Code: c, Status: c.Stati()[0]}).Error()
}

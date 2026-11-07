package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// msSpeicher ist der Speicher der Ende-zu-Ende-Fälle der Mutationen: Entity-State und Revision eines Workspace `ws1` im Hauptspeicher,
// Transaktionen laufen hintereinander und rollen bei einem Fehler zurück. Nur der Zusammenhang von URL-Parser, JSON-Leser,
// CSDL-Leser, Schreiber und Dienst ist der Gegenstand (SPEC-196 *Ungeprüft*, Verdrahtung).
type msSpeicher struct {
	tx  sync.Mutex
	mu  sync.Mutex
	es  map[string]model.EntityState
	rev int64
}

func msSchluessel(set, key string) string { return set + "\x00" + key }

type msWorkspaces struct {
	outbound.WorkspaceRepository
	s *msSpeicher
}

func (w msWorkspaces) Get(_ context.Context, id string) (model.Workspace, error) {
	if id != "ws1" {
		return model.Workspace{}, outbound.ErrNichtGefunden
	}
	w.s.mu.Lock()
	defer w.s.mu.Unlock()
	return model.Workspace{ID: "ws1", ServiceID: "svc", BaselineSnapshotID: "b1", Revision: w.s.rev}, nil
}

func (w msWorkspaces) Advance(_ context.Context, _ string, erwartet int64) (int64, error) {
	w.s.mu.Lock()
	defer w.s.mu.Unlock()
	if w.s.rev != erwartet {
		return 0, outbound.ErrRevisionVeraltet
	}
	w.s.rev++
	return w.s.rev, nil
}

type msEntities struct {
	outbound.EntityRepository
	s *msSpeicher
}

func (e msEntities) Put(_ context.Context, x model.EntityState) error {
	e.s.mu.Lock()
	defer e.s.mu.Unlock()
	e.s.es[msSchluessel(x.EntitySet, x.EntityKey)] = x
	return nil
}

func (e msEntities) Get(_ context.Context, _, set, key string) (model.EntityState, error) {
	e.s.mu.Lock()
	defer e.s.mu.Unlock()
	if x, da := e.s.es[msSchluessel(set, key)]; da {
		return x, nil
	}
	return model.EntityState{}, outbound.ErrNichtGefunden
}

func (e msEntities) ListSet(_ context.Context, _, set string) ([]model.EntityState, error) {
	e.s.mu.Lock()
	defer e.s.mu.Unlock()
	var r []model.EntityState
	for _, x := range e.s.es {
		if x.EntitySet == set {
			r = append(r, x)
		}
	}
	return r, nil
}

type msTransaktion struct{ s *msSpeicher }

func (t msTransaktion) WithWorkspaceTransaction(_ context.Context, f func(outbound.WorkspaceStores) error) error {
	t.s.tx.Lock()
	defer t.s.tx.Unlock()
	t.s.mu.Lock()
	es, rev := maps.Clone(t.s.es), t.s.rev
	t.s.mu.Unlock()
	err := f(outbound.WorkspaceStores{Workspaces: msWorkspaces{s: t.s}, Entities: msEntities{s: t.s}})
	if err != nil {
		t.s.mu.Lock()
		t.s.es, t.s.rev = es, rev
		t.s.mu.Unlock()
	}
	return err
}

// stand ist eine Kopie von Entity-State und Revision.
func (s *msSpeicher) stand() (map[string]model.EntityState, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return maps.Clone(s.es), s.rev
}

// msDienst baut den Dienst mit den echten Adaptern (Parser der URL, Leser des CSDL, Leser der Bodies, Schreiber) und dem Speicher.
func msDienst(t *testing.T, xml, root string, status model.CoverageStatus, sch outbound.ODataSchreiber) (*mock.Mocker, *msSpeicher) {
	t.Helper()
	s := &msSpeicher{es: map[string]model.EntityState{}}
	if sch == nil {
		sch = jsoncodec.NewSchreiber()
	}
	m, err := mock.New(mock.Konfig{
		Dienste: []mock.Dienst{{ID: "svc", Root: root}}, Metadaten: mlMetadaten{}, Workspaces: msWorkspaces{s: s}, Entities: msEntities{s: s},
		Coverage: mlCoverage{status: status}, Blobs: mlBlobs{xml: xml}, MetadatenCodec: csdl.New(),
		URLCodec: urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())), Schreiber: sch, Eingabe: jsoncodec.NewEingabe(),
		Transaktion: msTransaktion{s: s}, PublicBaseURL: "http://recorder.test", MaxBytes: 8 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	return m, s
}

// msRuf ruft die Methode mit dem Body und liefert Status, Header, Body-Text und Fehler.
func msRuf(t *testing.T, m *mock.Mocker, methode, pfad, body string, header map[string][]string) (inbound.Antwort, string, error) {
	t.Helper()
	h := map[string][]string{"X-Recorder-Workspace": {"ws1"}}
	if body != "" {
		h["Content-Type"] = []string{"application/json"}
	}
	for k, v := range header {
		h[k] = v
	}
	a := inbound.Anfrage{Methode: methode, Pfad: pfad, Header: h, Laenge: -1}
	if body != "" {
		a.Body, a.Laenge = io.NopCloser(strings.NewReader(body)), int64(len(body))
	}
	ant, err := m.Mock(context.Background(), a)
	if err != nil {
		return ant, "", err
	}
	b, rerr := io.ReadAll(ant.Body)
	if rerr != nil {
		t.Fatal(rerr)
	}
	_ = ant.Body.Close()
	return ant, string(b), nil
}

// msPflichtXML hat einen Typ P mit dem Schlüssel Id, drei erforderlichen Properties R1 bis R3 und einer nullbaren O1.
func msPflichtXML() string {
	return `<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"><edmx:DataServices>` +
		`<Schema Namespace="N" xmlns="http://docs.oasis-open.org/odata/ns/edm"><EntityType Name="P"><Key><PropertyRef Name="Id"/></Key>` +
		`<Property Name="Id" Type="Edm.Int32" Nullable="false"/><Property Name="R1" Type="Edm.String" Nullable="false"/>` +
		`<Property Name="R2" Type="Edm.String" Nullable="false"/><Property Name="R3" Type="Edm.String" Nullable="false"/>` +
		`<Property Name="O1" Type="Edm.String"/></EntityType><EntityContainer Name="C"><EntitySet Name="Ps" EntityType="N.P"/></EntityContainer>` +
		`</Schema></edmx:DataServices></edmx:Edmx>`
}

// msVorhanden legt den Eintrag Ps(1) mit dem Zustand vor der Mutation an.
func msVorhanden(s *msSpeicher) {
	s.es[msSchluessel("Ps", "1")] = model.EntityState{
		WorkspaceID: "ws1", EntitySet: "Ps", EntityKey: "1", Completeness: model.EntityComplete, Version: 3,
		Payload: `{"Id":1,"O1":null,"R1":"x1","R2":"x2","R3":"x3"}`, KnownProperties: []string{"Id", "O1", "R1", "R2", "R3"},
	}
}

// gleicherStand prüft, dass Entity-State und Revision die vorherigen sind.
func gleicherStand(t *testing.T, s *msSpeicher, es map[string]model.EntityState, rev int64, kontext string) {
	t.Helper()
	nachher, r := s.stand()
	if r != rev || len(nachher) != len(es) {
		t.Errorf("%s: Revision %d (vorher %d), Einträge %d (vorher %d)", kontext, r, rev, len(nachher), len(es))
	}
	for k, x := range es {
		if y := nachher[k]; y.Payload != x.Payload || y.Version != x.Version || y.Deleted != x.Deleted {
			t.Errorf("%s: Eintrag %q verändert", kontext, k)
		}
	}
}

// TestPflichtEndeZuEnde (W38): PUT und POST ohne erforderliche Property sind 400 mit ODR-E6007, der Speicher bleibt unverändert; die
// erste fehlende in der Folge des Modells entscheidet; Position der fehlenden {erste, mittlere, letzte erforderliche} × Methode 2 = 6,
// dazu das Paar „unbekannte Property und fehlende Pflicht“ (der Verstoß des Mitglieds gewinnt).
func TestPflichtEndeZuEnde(t *testing.T) {
	voll := map[string]string{"R1": `"a"`, "R2": `"b"`, "R3": `"c"`}
	n := 0
	for _, fehlt := range []string{"R1", "R2", "R3"} {
		for _, meth := range []string{"POST", "PUT"} {
			n++
			m, s := msDienst(t, msPflichtXML(), "/svc", model.CoverageComplete, nil)
			msVorhanden(s)
			es, rev := s.stand()
			var teile []string
			teile = append(teile, `"Id":`+map[string]string{"POST": "7", "PUT": "1"}[meth])
			for _, r := range []string{"R1", "R2", "R3"} {
				if r != fehlt {
					teile = append(teile, `"`+r+`":`+voll[r])
				}
			}
			pfad := "/svc/Ps"
			if meth == "PUT" {
				pfad = "/svc/Ps(1)"
			}
			_, _, err := msRuf(t, m, meth, pfad, "{"+strings.Join(teile, ",")+"}", nil)
			var f *inbound.Fehler
			var ef *model.EingabeFehler
			if !errors.As(err, &f) || f.Code != model.CodeE6007 || f.Status != 400 || !errors.As(err, &ef) || ef.Ursache != model.EingabeFehlend || ef.Ort != "/"+fehlt {
				t.Errorf("%s ohne %s: %v (Ursache %v)", meth, fehlt, err, ef)
			}
			gleicherStand(t, s, es, rev, meth+" ohne "+fehlt)
		}
	}
	if n != 6 {
		t.Errorf("Menge %d, erwartet 6", n)
	}
	m, s := msDienst(t, msPflichtXML(), "/svc", model.CoverageComplete, nil)
	es, rev := s.stand()
	_, _, err := msRuf(t, m, "POST", "/svc/Ps", `{"Id":7,"R1":"a","Unbekannt":1}`, nil)
	var ef *model.EingabeFehler
	if !errors.As(err, &ef) || ef.Ursache != model.EingabeUnbekannteProperty {
		t.Errorf("Paar: %v, erwartet unbekannte_property", err)
	}
	gleicherStand(t, s, es, rev, "Paar")
}

// TestUnbekanntEndeZuEnde (W39): eine unbekannte Property ist 400 mit ODR-E6007 bei POST, PATCH und PUT, der Speicher bleibt
// unverändert; Position des Verstoßes {erstes, mittleres, letztes Mitglied} × Methode 3 = 9.
func TestUnbekanntEndeZuEnde(t *testing.T) {
	n := 0
	for pos := 0; pos < 3; pos++ {
		for _, meth := range []string{"POST", "PATCH", "PUT"} {
			n++
			m, s := msDienst(t, msPflichtXML(), "/svc", model.CoverageComplete, nil)
			msVorhanden(s)
			es, rev := s.stand()
			id := map[string]string{"POST": "7", "PATCH": "1", "PUT": "1"}[meth]
			teile := []string{`"Id":` + id, `"R1":"a"`, `"R2":"b"`, `"R3":"c"`}
			teile = append(teile[:pos+1], append([]string{`"Unbekannt":1`}, teile[pos+1:]...)...)
			pfad := "/svc/Ps"
			if meth != "POST" {
				pfad = "/svc/Ps(1)"
			}
			_, _, err := msRuf(t, m, meth, pfad, "{"+strings.Join(teile, ",")+"}", nil)
			var f *inbound.Fehler
			var ef *model.EingabeFehler
			if !errors.As(err, &f) || f.Code != model.CodeE6007 || f.Status != 400 || !errors.As(err, &ef) || ef.Ursache != model.EingabeUnbekannteProperty || ef.Ort != "/Unbekannt" {
				t.Errorf("%s Position %d: %v", meth, pos, err)
			}
			gleicherStand(t, s, es, rev, meth)
		}
	}
	if n != 9 {
		t.Errorf("Menge %d, erwartet 9", n)
	}
}

// TestLocationEndeZuEnde (W52, M26): Location gleich @odata.id der geschriebenen Entity, nie aus dem Host-Header, Wurzel aus
// PUBLIC_BASE_URL und Root. Host 3 × Root 3 × Schlüsseltyp 8 = 72; das Orakel ist die handgeschriebene Adresse.
func TestLocationEndeZuEnde(t *testing.T) {
	g := "01234567-89ab-cdef-0123-456789abcdef"
	typen := []struct{ set, json, literal string }{
		{"S0", `"abc"`, `'abc'`}, {"S1", `true`, `true`}, {"S2", `7`, `7`}, {"S3", `"9007199254740993"`, `9007199254740993`},
		{"S4", `"1.5"`, `1.5`}, {"S5", `"` + g + `"`, g}, {"S6", `"2026-11-07"`, `2026-11-07`}, {"S7", `"2026-11-07T12:00:00Z"`, `2026-11-07T12%3A00%3A00Z`},
	}
	hosts := []map[string][]string{nil, {"Host": {"evil.example"}}, {"Host": {"localhost:1"}}}
	n := 0
	for _, h := range hosts {
		for _, wurzel := range []struct{ root, praefix string }{{"/", ""}, {"/svc", "/svc"}, {"/a/b", "/a/b"}} {
			for _, ty := range typen {
				n++
				m, _ := msDienst(t, mlSchluesselXML(), wurzel.root, model.CoverageComplete, nil)
				pfad := wurzel.praefix + "/" + ty.set
				ant, body, err := msRuf(t, m, "POST", pfad, `{"K":`+ty.json+`}`, h)
				want := "http://recorder.test" + wurzel.praefix + "/" + ty.set + "(" + ty.literal + ")"
				if err != nil || ant.Status != 201 || len(ant.Header["Location"]) != 1 || ant.Header["Location"][0] != want {
					t.Errorf("%s %s: Status %d Location %v Fehler %v, erwartet %s", wurzel.root, ty.set, ant.Status, ant.Header["Location"], err, want)
					continue
				}
				var o map[string]any
				if json.Unmarshal([]byte(body), &o) != nil || o["@odata.id"] != want {
					t.Errorf("%s %s: @odata.id %v, erwartet %s (Body %s)", wurzel.root, ty.set, o["@odata.id"], want, body)
				}
			}
		}
	}
	if n != 72 {
		t.Errorf("Menge %d, erwartet 72", n)
	}
}

// TestNameLang (W65, F-15): ein Name von grenze Bytes in einer unbekannten Property ergibt in Error() des Dienstfehlers, in Error()
// seiner Ursache und im Fehlerobjekt keinen Namen über 256 Bytes. Weg 3 × Name 3 (256, 257 Bytes und 1 MiB; das Zeichen `#` steht
// als letztes).
func TestNameLang(t *testing.T) {
	n := 0
	for _, laenge := range []int{256, 257, 1 << 20} {
		name := strings.Repeat("n", laenge-1) + "#"
		m, s := msDienst(t, msPflichtXML(), "/svc", model.CoverageComplete, nil)
		es, rev := s.stand()
		_, _, err := msRuf(t, m, "POST", "/svc/Ps", `{"Id":7,"`+name+`":1}`, nil)
		var f *inbound.Fehler
		if !errors.As(err, &f) || f.Code != model.CodeE6007 {
			t.Fatalf("Länge %d: %v", laenge, err)
		}
		obj, oerr := model.NeuesFehlerobjekt(f.Code)
		if oerr != nil {
			t.Fatal(oerr)
		}
		roh, _ := json.Marshal(obj)
		wege := []string{f.Error(), f.Ursache.Error(), string(roh)}
		for i, w := range wege {
			n++
			// Der Ort ist der JSON-Pointer `/<Name>` (mindestens 257 Bytes): das letzte Zeichen des Namens steht in keinem Weg,
			// der Name nur im Weg 1 (Ursache) und dort höchstens in den ersten 256 Bytes des Ortes.
			if len(w) > 1280 || strings.Contains(w, "#") || (i != 1 && strings.Contains(w, "nnnn")) {
				t.Errorf("Länge %d Weg %d: %d Bytes, Name oder letztes Zeichen enthalten", laenge, i, len(w))
			}
		}
		if !strings.HasSuffix(wege[1], `"...`) || strings.Count(wege[1], "n") > 256+len("unbekannte_property")+len("ODR-E6007: Request-Body ungültig:") {
			t.Errorf("Länge %d: Ursache %d Bytes ohne Kürzung am Ende oder mit zu langem Namen", laenge, len(wege[1]))
		}
		gleicherStand(t, s, es, rev, fmt.Sprint("Länge ", laenge))
	}
	if n != 9 {
		t.Errorf("Menge %d, erwartet 9", n)
	}
}

// TestHappyPathEndeZuEnde (W78, LH-FA-07): POST, GET, PATCH, GET, DELETE, GET mit handgeschriebenen Bodies; die Übergänge sind in
// den folgenden GET sichtbar, das letzte GET ist 404. Schlüsseltyp 3 (Edm.Int32, Edm.String, zusammengesetzt aus Edm.Int64 und Edm.Guid).
func TestHappyPathEndeZuEnde(t *testing.T) {
	g := "01234567-89ab-cdef-0123-456789abcdef"
	faelle := []struct{ name, set, bodyKey, urlKey string }{
		{"Edm.Int32", "S2", `"K":7`, `7`}, {"Edm.String", "S0", `"K":"abc"`, `'abc'`},
		{"zusammengesetzt", "S9", `"L":"5","G":"` + g + `"`, `L=5,G=` + g},
	}
	for _, f := range faelle {
		m, _ := msDienst(t, mlSchluesselXML(), "/svc", model.CoverageComplete, nil)
		ant, _, err := msRuf(t, m, "POST", "/svc/"+f.set, `{`+f.bodyKey+`,"V":"eins"}`, nil)
		if err != nil || ant.Status != 201 {
			t.Fatalf("%s POST: %d %v", f.name, ant.Status, err)
		}
		pfad := "/svc/" + f.set + "(" + f.urlKey + ")"
		schritte := []struct {
			meth, body string
			status     int
			enthaelt   string
		}{
			{"GET", "", 200, `"V":"eins"`}, {"PATCH", `{"V":"zwei"}`, 204, ""}, {"GET", "", 200, `"V":"zwei"`}, {"DELETE", "", 204, ""},
		}
		for _, s := range schritte {
			ant, body, err := msRuf(t, m, s.meth, pfad, s.body, nil)
			if err != nil || ant.Status != s.status || !strings.Contains(body, s.enthaelt) {
				t.Errorf("%s %s: %d %v %q, erwartet %d mit %q", f.name, s.meth, ant.Status, err, body, s.status, s.enthaelt)
			}
		}
		_, _, err = msRuf(t, m, "GET", pfad, "", nil)
		var fe *inbound.Fehler
		if !errors.As(err, &fe) || fe.Code != model.CodeE6008 || fe.Status != 404 {
			t.Errorf("%s: GET nach DELETE: %v, erwartet ODR-E6008 mit 404", f.name, err)
		}
	}
}

// msFehlerSchreiber ist ein Schreiber, dessen SchreibeEntity scheitert.
type msFehlerSchreiber struct{ outbound.ODataSchreiber }

func (msFehlerSchreiber) SchreibeEntity(context.Context, io.Writer, model.ServiceModel, model.EntityAntwort) (model.Darstellung, error) {
	return model.Darstellung{}, errors.New("Schreiber scheitert")
}

// msSqliteDienst baut den Dienst über der echten SQLite-Datei: Import einer Baseline, Workspace `ws-a`, Metadaten des Service.
func msSqliteDienst(t *testing.T, sch outbound.ODataSchreiber) (*mock.Mocker, *wsDatei) {
	t.Helper()
	ctx := context.Background()
	d := wsNeueDatei(t)
	d.baseline("bsl-1", []wsEntity{{"Products", "1", `{"Id":1,"Name":"Milk"}`}})
	ref := d.blob([]byte(msProductsXML()))
	if err := d.db.Metadata().SetMetadata(ctx, "svc", ref, ref); err != nil {
		t.Fatal(err)
	}
	if _, err := d.dienst.Create(ctx, "bsl-1", "ws-a"); err != nil {
		t.Fatal(err)
	}
	if sch == nil {
		sch = jsoncodec.NewSchreiber()
	}
	m, err := mock.New(mock.Konfig{
		Dienste: []mock.Dienst{{ID: "svc", Root: "/svc"}}, Metadaten: d.db.Metadata(), Workspaces: d.db.Workspaces(), Entities: d.db.Entities(),
		Coverage: d.db.Coverage(), Blobs: d.store, MetadatenCodec: csdl.New(), URLCodec: urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())),
		Schreiber: sch, Eingabe: jsoncodec.NewEingabe(), Transaktion: d.db, PublicBaseURL: "http://recorder.test", MaxBytes: 1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	return m, d
}

// msProductsXML hat den Typ Product mit dem Schlüssel Id und der nullbaren Name im Set Products.
func msProductsXML() string {
	return `<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"><edmx:DataServices>` +
		`<Schema Namespace="N" xmlns="http://docs.oasis-open.org/odata/ns/edm"><EntityType Name="Product"><Key><PropertyRef Name="Id"/></Key>` +
		`<Property Name="Id" Type="Edm.Int32" Nullable="false"/><Property Name="Name" Type="Edm.String"/></EntityType>` +
		`<EntityContainer Name="C"><EntitySet Name="Products" EntityType="N.Product"/></EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`
}

// msRufWs ruft den Dienst für den Workspace ws-a.
func msRufWs(t *testing.T, m *mock.Mocker, methode, pfad, body string) (int, string, error) {
	t.Helper()
	ant, b, err := msRuf(t, m, methode, pfad, body, map[string][]string{"X-Recorder-Workspace": {"ws-a"}})
	return ant.Status, b, err
}

// TestMutationSqlite (W76): Ende-zu-Ende gegen den echten SQLite-Adapter (Transaktion, Repositories): die Abnahmeschleife POST, GET,
// PATCH, GET, DELETE, GET mit den Zustandsübergängen; n gleichzeitige POST desselben Schlüssels: genau eines 201; ein Fehler des
// Schreibers rollt zurück. 16 Goroutinen × 4 Runden.
func TestMutationSqlite(t *testing.T) {
	ctx := context.Background()
	m, d := msSqliteDienst(t, nil)
	schritte := []struct {
		meth, pfad, body string
		status           int
		enthaelt         string
	}{
		{"POST", "/svc/Products", `{"Id":7,"Name":"eins"}`, 201, `"Name":"eins"`},
		{"GET", "/svc/Products(7)", "", 200, `"Name":"eins"`},
		{"PATCH", "/svc/Products(7)", `{"Name":"zwei"}`, 204, ""},
		{"GET", "/svc/Products(7)", "", 200, `"Name":"zwei"`},
		{"DELETE", "/svc/Products(7)", "", 204, ""},
	}
	for _, s := range schritte {
		st, b, err := msRufWs(t, m, s.meth, s.pfad, s.body)
		if err != nil || st != s.status || !strings.Contains(b, s.enthaelt) {
			t.Fatalf("%s %s: %d %v %q", s.meth, s.pfad, st, err, b)
		}
	}
	_, _, err := msRufWs(t, m, "GET", "/svc/Products(7)", "")
	var fe *inbound.Fehler
	if !errors.As(err, &fe) || fe.Code != model.CodeE6008 {
		t.Errorf("GET nach DELETE: %v, erwartet ODR-E6008", err)
	}
	if w, err := d.dienst.Get(ctx, "ws-a"); err != nil || w.Revision != 3 {
		t.Errorf("Revision %d (%v), erwartet 3", w.Revision, err)
	}
	for runde := 0; runde < 4; runde++ {
		key := 100 + runde
		var wg sync.WaitGroup
		var mu sync.Mutex
		stati := map[int]int{}
		codes := map[model.Code]int{}
		start := make(chan struct{})
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				st, _, err := msRufWs(t, m, "POST", "/svc/Products", fmt.Sprintf(`{"Id":%d,"Name":"g%d"}`, key, i))
				mu.Lock()
				defer mu.Unlock()
				var f *inbound.Fehler
				if errors.As(err, &f) {
					codes[f.Code]++
				} else {
					stati[st]++
				}
			}()
		}
		close(start)
		wg.Wait()
		if stati[201] != 1 || codes[model.CodeE6009] != 15 || len(stati) != 1 || len(codes) != 1 {
			t.Errorf("Runde %d: Status %v Codes %v, erwartet ein 201 und 15 × ODR-E6009", runde, stati, codes)
		}
	}
	// Ein Fehler des Schreibers rollt zurück: Zustand und Revision wie vorher.
	m2, d2 := msSqliteDienst(t, msFehlerSchreiber{jsoncodec.NewSchreiber()})
	vorher := d2.state("ws-a")
	_, _, err = msRufWs(t, m2, "POST", "/svc/Products", `{"Id":9,"Name":"x"}`)
	if !errors.As(err, &fe) || fe.Code != model.CodeE9001 {
		t.Errorf("Fehler des Schreibers: %v, erwartet ODR-E9001", err)
	}
	if got := d2.state("ws-a"); len(got) != len(vorher) {
		t.Errorf("Zustand nach dem Rollback: %d Einträge, vorher %d", len(got), len(vorher))
	}
	if w, _ := d2.dienst.Get(ctx, "ws-a"); w.Revision != 0 {
		t.Errorf("Revision nach dem Rollback %d, erwartet 0", w.Revision)
	}
}

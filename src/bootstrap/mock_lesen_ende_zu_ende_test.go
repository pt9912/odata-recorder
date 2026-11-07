package bootstrap_test

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// Die Ports des Mock-Dienstes für die Ende-zu-Ende-Fälle: der Speicher ist in diesen Tests nicht der Gegenstand, nur das Zusammenspiel
// von URL-Parser, JSON-Leser, CSDL-Leser, Schreiber und Dienst (SPEC-196 *Ungeprüft*, Verdrahtung); jeder Port ruft allein, was der Dienst ruft.
type mlWorkspaces struct{ outbound.WorkspaceRepository }

func (mlWorkspaces) Get(_ context.Context, id string) (model.Workspace, error) {
	if id != "ws1" {
		return model.Workspace{}, outbound.ErrNichtGefunden
	}
	return model.Workspace{ID: "ws1", ServiceID: "svc", BaselineSnapshotID: "b1"}, nil
}

type mlMetadaten struct{ outbound.MetadataRepository }

func (mlMetadaten) Get(context.Context, string) (model.Service, error) {
	return model.Service{ID: "svc", MetadataBlob: "sha256:" + strings.Repeat("0", 63) + "1"}, nil
}

type mlBlobs struct {
	outbound.BlobRepository
	xml string
}

func (b mlBlobs) Open(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(b.xml)), nil
}

type mlEntities struct {
	outbound.EntityRepository
	es []model.EntityState
}

func (e mlEntities) Get(_ context.Context, _, set, key string) (model.EntityState, error) {
	for _, x := range e.es {
		if x.EntitySet == set && x.EntityKey == key {
			return x, nil
		}
	}
	return model.EntityState{}, outbound.ErrNichtGefunden
}

func (e mlEntities) ListSet(_ context.Context, _, set string) ([]model.EntityState, error) {
	var r []model.EntityState
	for _, x := range e.es {
		if x.EntitySet == set {
			r = append(r, x)
		}
	}
	return r, nil
}

type mlCoverage struct {
	outbound.CoverageRepository
	status model.CoverageStatus
}

func (c mlCoverage) Get(_ context.Context, _, set string) (model.Coverage, error) {
	return model.Coverage{EntitySet: set, Status: c.status, EvidenceJSON: "{}"}, nil
}

// mlDienst baut den Dienst mit den echten Adaptern: Parser der URL (mit Leser von `$filter`), Leser des CSDL, Schreiber der
// Antworten der Version 4.
func mlDienst(t *testing.T, xml string, status model.CoverageStatus, es ...model.EntityState) *mock.Mocker {
	t.Helper()
	m, err := mock.New(mock.Konfig{
		Dienste: []mock.Dienst{{ID: "svc", Root: "/svc"}}, Metadaten: mlMetadaten{}, Workspaces: mlWorkspaces{}, Entities: mlEntities{es: es},
		Coverage: mlCoverage{status: status}, Blobs: mlBlobs{xml: xml}, MetadatenCodec: csdl.New(),
		URLCodec: urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())), Schreiber: jsoncodec.NewSchreiber(),
		PublicBaseURL: "http://recorder.test", MaxBytes: 1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// mlRuf ruft GET (oder HEAD) mit Pfad und Query und liefert Status, Body und Fehler.
func mlRuf(t *testing.T, m *mock.Mocker, pfad, query string, header map[string][]string) (int, string, error) {
	t.Helper()
	h := map[string][]string{"X-Recorder-Workspace": {"ws1"}}
	for k, v := range header {
		h[k] = v
	}
	a, err := m.Mock(context.Background(), inbound.Anfrage{Methode: "GET", Pfad: pfad, RawQuery: query, Header: h, Laenge: -1})
	if err != nil {
		return 0, "", err
	}
	b, rerr := io.ReadAll(a.Body)
	if rerr != nil {
		t.Fatal(rerr)
	}
	_ = a.Body.Close()
	return a.Status, string(b), nil
}

func mlCode(err error) string {
	if f, ok := err.(*inbound.Fehler); ok {
		return f.Code.String()
	}
	return ""
}

// mlSchluesselXML ist das CSDL der Schlüsselfälle: T0 bis T7 mit einem Schlüssel K der acht EDM-Typen, T8 bis T10 zusammengesetzt.
func mlSchluesselXML() string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"><edmx:DataServices>` +
		`<Schema Namespace="N" xmlns="http://docs.oasis-open.org/odata/ns/edm">`)
	typ := func(name string, keys ...[2]string) {
		sb.WriteString(`<EntityType Name="` + name + `"><Key>`)
		for _, k := range keys {
			sb.WriteString(`<PropertyRef Name="` + k[0] + `"/>`)
		}
		sb.WriteString(`</Key>`)
		for _, k := range keys {
			sb.WriteString(`<Property Name="` + k[0] + `" Type="` + k[1] + `" Nullable="false"/>`)
		}
		sb.WriteString(`<Property Name="V" Type="Edm.String"/></EntityType>`)
	}
	for i, edm := range model.UnterstuetzteEDMTypen() {
		typ("T"+strconv.Itoa(i), [2]string{"K", edm})
	}
	typ("T8", [2]string{"S", "Edm.String"}, [2]string{"I", "Edm.Int32"})
	typ("T9", [2]string{"L", "Edm.Int64"}, [2]string{"G", "Edm.Guid"})
	typ("T10", [2]string{"A", "Edm.String"}, [2]string{"B", "Edm.Int32"}, [2]string{"C", "Edm.Guid"})
	sb.WriteString(`<EntityContainer Name="C">`)
	for i := 0; i <= 10; i++ {
		fmt.Fprintf(&sb, `<EntitySet Name="S%d" EntityType="N.T%d"/>`, i, i)
	}
	sb.WriteString(`</EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`)
	return sb.String()
}

// mlPayload ist der Test-Encoder der Kanonform des Payloads (SPEC-127 Nr. 11 a, SPEC-196 Nr. 3): Strings und Werte, die als
// String stehen, in Anführungszeichen, Boolean und Int32 als Literal, `null` als `null`.
func mlPayload(e model.Entity) string {
	var teile []string
	for _, p := range e.Properties {
		var v string
		switch p.Wert.Art {
		case model.ArtNull:
			v = "null"
		case model.ArtBoolean, model.ArtInt32:
			v = p.Wert.Text
		default:
			v = strconv.Quote(p.Wert.Text)
		}
		teile = append(teile, strconv.Quote(p.Name)+":"+v)
	}
	return "{" + strings.Join(teile, ",") + "}"
}

// TestSchluesselEndeZuEnde hält L33 (M7, Sendung (d) von slice-odata-url-pfad, SPEC-196 Nr. 7): der Schlüssel der URL findet die
// Entity, je Typ der acht, in mehreren Schreibweisen. Die Entity kommt aus dem JSON-Leser (Body einer Antwort, dann der Test-Encoder
// der Kanonform), die URL ist handgeschrieben, der echte Parser, der echte Leser des CSDL, der echte Schreiber und der Dienst
// laufen dazwischen. Typ 8 × Schreibweise 2 (kanonisch, abweichend) + zusammengesetzt 3 + leerer String-Schlüssel 1 = 20 Elemente.
func TestSchluesselEndeZuEnde(t *testing.T) {
	xml := mlSchluesselXML()
	modell, _, err := csdl.New().ParseMetadata(context.Background(), strings.NewReader(xml), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	g := "01234567-89ab-cdef-0123-456789abcdef"
	faelle := []struct {
		name, set, json, url string
	}{
		{"String kanonisch", "S0", `"K":"abc"`, `'abc'`}, {"String mit Apostroph", "S0", `"K":"o'k"`, `'o''k'`},
		{"Boolean kanonisch", "S1", `"K":true`, `true`}, {"Boolean TRUE", "S1", `"K":true`, `TRUE`},
		{"Int32 kanonisch", "S2", `"K":7`, `7`}, {"Int32 007", "S2", `"K":7`, `007`},
		{"Int64 kanonisch", "S3", `"K":"9007199254740993"`, `9007199254740993`}, {"Int64 -0", "S3", `"K":"0"`, `-0`},
		{"Decimal kanonisch", "S4", `"K":"1.5"`, `1.5`}, {"Decimal 1.50", "S4", `"K":"1.5"`, `1.50`},
		{"Guid kanonisch", "S5", `"K":"` + g + `"`, g}, {"Guid Großbuchstaben", "S5", `"K":"` + g + `"`, strings.ToUpper(g)},
		{"Date kanonisch", "S6", `"K":"2026-11-07"`, `2026-11-07`}, {"Date ohne Wirkung", "S6", `"K":"2026-11-07"`, `2026-11-07`},
		{"DateTimeOffset kanonisch", "S7", `"K":"2026-11-07T12:00:00Z"`, `2026-11-07T12:00:00Z`},
		{"DateTimeOffset mit Versatz", "S7", `"K":"2026-11-07T12:00:00Z"`, `2026-11-07T14:00:00+02:00`},
		{"String und Int32", "S8", `"S":"x","I":3`, `S='x',I=3`},
		{"Int64 und Guid", "S9", `"L":"5","G":"` + g + `"`, `L=5,G=` + g},
		{"drei Teile abweichende Folge", "S10", `"A":"z","B":1,"C":"` + g + `"`, `C=` + g + `,A='z',B=1`},
	}
	for _, f := range faelle {
		body := `{"@odata.context":"http://x/$metadata#` + f.set + `/$entity",` + f.json + `,"V":"v"}`
		e, err := jsoncodec.New().ParseEntity(context.Background(), strings.NewReader(body), modell, f.set, 1<<20)
		if err != nil {
			t.Errorf("%s: der Leser: %v", f.name, err)
			continue
		}
		m := mlDienst(t, xml, model.CoverageComplete, model.EntityState{WorkspaceID: "ws1", EntitySet: f.set, EntityKey: e.Schluessel,
			Payload: mlPayload(e), Completeness: model.EntityComplete, KnownProperties: []string{"K", "S", "I", "L", "G", "A", "B", "C", "V"}})
		status, antwort, err := mlRuf(t, m, "/svc/"+f.set+"("+f.url+")", "", nil)
		if err != nil || status != 200 || !strings.Contains(antwort, `"V":"v"`) || !strings.HasPrefix(antwort, `{"@odata.context":"http://recorder.test/svc/$metadata#`+f.set+`/$entity"`) {
			t.Errorf("%s: URL %s, Schlüssel des Lesers %q: Status %d, Fehler %v, Body %q", f.name, f.url, e.Schluessel, status, err, antwort)
		}
	}
	if len(faelle) != 19 {
		t.Errorf("%d Fälle, erwartet 19 (und der leere Schlüssel)", len(faelle))
	}
	// Element 20: der leere String-Schlüssel: kein Treffer, `complete` ist ODR-E6008, `partial` ODR-E6003.
	for status, soll := range map[model.CoverageStatus]string{model.CoverageComplete: "ODR-E6008", model.CoveragePartial: "ODR-E6003"} {
		_, _, err := mlRuf(t, mlDienst(t, xml, status), "/svc/S0('')", "", nil)
		if mlCode(err) != soll {
			t.Errorf("S0('') bei %s: %v, erwartet %s", status, err, soll)
		}
	}
}

// mlProdukte ist das CSDL der Abfragen: Products mit Id (Edm.Int32), Name und Count.
const mlProdukteXML = `<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"><edmx:DataServices>` +
	`<Schema Namespace="N" xmlns="http://docs.oasis-open.org/odata/ns/edm"><EntityType Name="P"><Key><PropertyRef Name="Id"/></Key>` +
	`<Property Name="Id" Type="Edm.Int32" Nullable="false"/><Property Name="Name" Type="Edm.String"/><Property Name="Count" Type="Edm.Int32"/></EntityType>` +
	`<EntityContainer Name="C"><EntitySet Name="Products" EntityType="N.P"/></EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`

func mlProdukte(t *testing.T) *mock.Mocker {
	t.Helper()
	var es []model.EntityState
	for i, n := range []string{"a", "b", "c"} {
		es = append(es, model.EntityState{WorkspaceID: "ws1", EntitySet: "Products", EntityKey: strconv.Itoa(i + 1), Completeness: model.EntityComplete,
			Payload: fmt.Sprintf(`{"Id":%d,"Name":%q,"Count":%d}`, i+1, n, 10*(i+1)), KnownProperties: []string{"Id", "Name", "Count"}})
	}
	return mlDienst(t, mlProdukteXML, model.CoverageComplete, es...)
}

// mlEl ist ein Element der Collection (ohne Kontext, SPEC-169 Nr. 5): Adresse und Properties.
func mlEl(id int, rest string) string {
	return fmt.Sprintf(`{"@odata.id":"http://recorder.test/svc/Products(%d)",%s}`, id, rest)
}

// TestAbfrageEndeZuEnde hält L34 (M17, Verdrahtung von Parser, Interpreter und Dienst): `$filter`, `$orderby`, `$top`, `$skip`,
// `$count`, `$select` und `/$count` mit Filter ergeben den erwarteten Body (handgeschrieben); die Fehler ODR-E6005 und ODR-E6001 des
// echten Parsers und des Interpreters. 8 Abfragen und 3 Fehler.
func TestAbfrageEndeZuEnde(t *testing.T) {
	ctx := `{"@odata.context":"http://recorder.test/svc/$metadata#Products"`
	a, b, c := mlEl(1, `"Id":1,"Name":"a","Count":10`), mlEl(2, `"Id":2,"Name":"b","Count":20`), mlEl(3, `"Id":3,"Name":"c","Count":30`)
	faelle := []struct{ name, pfad, query, soll string }{
		{"$filter", "/svc/Products", "$filter=Id gt 1", ctx + `,"value":[` + b + `,` + c + `]}`},
		{"$orderby gemischt", "/svc/Products", "$orderby=Count desc,Id", ctx + `,"value":[` + c + `,` + b + `,` + a + `]}`},
		{"$top", "/svc/Products", "$top=1", ctx + `,"value":[` + a + `]}`},
		{"$skip", "/svc/Products", "$skip=2", ctx + `,"value":[` + c + `]}`},
		{"$count", "/svc/Products", "$count=true", ctx + `,"@odata.count":3,"value":[` + a + `,` + b + `,` + c + `]}`},
		{"$select", "/svc/Products", "$select=Name", `{"@odata.context":"http://recorder.test/svc/$metadata#Products(Name)","value":[` +
			mlEl(1, `"Name":"a"`) + `,` + mlEl(2, `"Name":"b"`) + `,` + mlEl(3, `"Name":"c"`) + `]}`},
		{"alle zusammen", "/svc/Products", "$filter=Id gt 0&$orderby=Id desc&$skip=1&$top=1&$count=true&$select=Name,Id",
			`{"@odata.context":"http://recorder.test/svc/$metadata#Products(Name,Id)","@odata.count":3,"value":[` + mlEl(2, `"Id":2,"Name":"b"`) + `]}`},
		{"/$count mit Filter", "/svc/Products/$count", "$filter=Id gt 1", `2`},
	}
	m := mlProdukte(t)
	for _, f := range faelle {
		status, body, err := mlRuf(t, m, f.pfad, f.query, nil)
		if err != nil || status != 200 || body != f.soll {
			t.Errorf("%s: Status %d, Fehler %v\n Body %s\n soll %s", f.name, status, err, body, f.soll)
		}
	}
	fehler := []struct{ name, query, code string }{
		{"unbekannte Property in $orderby", "$orderby=Unbekannt", "ODR-E6005"}, {"$search", "$search=x", "ODR-E6001"}, {"$filter ungültig", "$filter=Id gt", "ODR-E6005"},
	}
	for _, f := range fehler {
		if _, _, err := mlRuf(t, m, "/svc/Products", f.query, nil); mlCode(err) != f.code {
			t.Errorf("%s: %v, erwartet %s", f.name, err, f.code)
		}
	}
}

// TestHostEndeZuEnde hält L35 (SPEC-198 Nr. 6, SPEC-168 Nr. 10): ein anderer Host ändert @odata.context und @odata.id nicht (der echte
// Schreiber), Body byte-gleich. Host 3 (fehlt, evil.example, localhost:1) × Art 2 (Collection, Entity).
func TestHostEndeZuEnde(t *testing.T) {
	m := mlProdukte(t)
	for _, pfad := range []string{"/svc/Products", "/svc/Products(2)"} {
		var erster string
		for i, host := range []string{"", "evil.example", "localhost:1"} {
			h := map[string][]string{}
			if host != "" {
				h["Host"] = []string{host}
				h["X-Forwarded-Host"] = []string{host}
			}
			status, body, err := mlRuf(t, m, pfad, "", h)
			if err != nil || status != 200 || strings.Contains(body, "evil") || strings.Contains(body, "localhost") || !strings.Contains(body, "http://recorder.test/svc/") {
				t.Fatalf("%s Host %q: %d %v %q", pfad, host, status, err, body)
			}
			if i == 0 {
				erster = body
			} else if body != erster {
				t.Errorf("%s Host %q: Body %q, ohne Host %q", pfad, host, body, erster)
			}
		}
	}
}

package bootstrap_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
)

// fakeMock ist der MockUseCase der Haken: er hält jede Anfrage mit ihrem ctx fest.
type fakeMock struct {
	anfragen []inbound.Anfrage
	ctxs     []context.Context
}

func (f *fakeMock) Mock(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	f.anfragen = append(f.anfragen, a)
	f.ctxs = append(f.ctxs, ctx)
	return inbound.Antwort{Status: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
}

// m3Haken setzt auf den Haken der Modi record und replay auf: er legt die Konfiguration von Mock und Snapshot-Dienst ab und
// zählt die Konstruktoren, die der Modus mock nie ruft.
type m3Haken struct {
	*haken
	mock     []mock.Konfig
	snapshot []snapshot.Konfig
	waechter []*maintenance.BlobWaechter
	oeffnen  int
	// verboten zählt je Name die Aufrufe der Konstruktoren eines Upstreams, einer Session oder des Imports.
	verboten map[string]int
	fake     *fakeMock
	srv      *fakeServer
}

func neuM3Haken() *m3Haken { return &m3Haken{haken: neuHaken(), verboten: map[string]int{}} }

// verbotene sind die sieben Konstruktoren, die der Modus mock nie ruft (SPEC-139 Nr. 11).
func verbotene() []string {
	return []string{"pipeline", "upstream", "session", "recorder", "import", "telemetrie", "replay"}
}

func (m *m3Haken) bauteile() bootstrap.Bauteile {
	b := m.haken.bauteile()
	echt := bootstrap.StandardBauteile()
	oeffne, waechter, pipeline, upstream, sess := b.Oeffne, b.NeuerWaechter, b.NeuePipeline, b.NeuerUpstream, b.NeueSession
	recorder, imp, tel, rep, srv := b.NeuerRecorder, b.NeuerImport, b.NeueTelemetrie, b.NeuerReplay, b.NeuerServer
	b.Oeffne = func(ctx context.Context, pfad string, opt sqlite.Options) (*sqlite.DB, error) {
		m.oeffnen++
		return oeffne(ctx, pfad, opt)
	}
	b.NeuerWaechter = func(bl outbound.BlobRepository, uhr support.Clock, karenz time.Duration) (*maintenance.BlobWaechter, error) {
		w, err := waechter(bl, uhr, karenz)
		m.waechter = append(m.waechter, w)
		return w, err
	}
	b.NeuePipeline = func(k capture.Konfig) (*capture.Pipeline, error) { m.verboten["pipeline"]++; return pipeline(k) }
	b.NeuerUpstream = func(k httpupstream.Konfig) (outbound.UpstreamClient, error) {
		m.verboten["upstream"]++
		return upstream(k)
	}
	b.NeueSession = func(k session.Konfig) (inbound.SessionUseCase, error) { m.verboten["session"]++; return sess(k) }
	b.NeuerRecorder = func(k record.Konfig) (inbound.RecordUseCase, error) { m.verboten["recorder"]++; return recorder(k) }
	b.NeuerImport = func(k baselineimport.Konfig) (inbound.ImportUseCase, error) { m.verboten["import"]++; return imp(k) }
	b.NeueTelemetrie = func(k telemetry.Config) (*telemetry.Adapter, error) { m.verboten["telemetrie"]++; return tel(k) }
	b.NeuerReplay = func(k replay.Konfig) (inbound.ReplayUseCase, error) { m.verboten["replay"]++; return rep(k) }
	b.NeuerSnapshot = func(k snapshot.Konfig) (inbound.SnapshotUseCase, error) {
		m.snapshot = append(m.snapshot, k)
		return echt.NeuerSnapshot(k)
	}
	b.NeuerMock = func(k mock.Konfig) (inbound.MockUseCase, error) {
		m.mock = append(m.mock, k)
		if m.fake != nil {
			return m.fake, nil
		}
		return echt.NeuerMock(k)
	}
	b.NeuerServer = func(k httpodata.Konfig) (bootstrap.Server, error) {
		s, err := srv(k)
		m.srv, _ = s.(*fakeServer)
		return s, err
	}
	return b
}

func mockKonfig(dir string) bootstrap.Konfiguration {
	k := basisKonfig(dir, "mock", bootstrap.Dienst{ID: "a", Root: "/a"}, bootstrap.Dienst{ID: "b", Root: "/b/c"})
	k.PublicBaseURL, k.StrictCoverage, k.MaxCaptureBody = "http://127.0.0.1:8080/api", true, 12345
	return k
}

// TestM3WahlJeModus fängt A07: Modus × Zugang des Handle mit `== nil` auf der Schnittstelle, im Modus mock kein Aufruf eines
// Konstruktors für Upstream, Pipeline, Recorder, Session, Import, Replay und Telemetrie, und `Verarbeiten` des Listeners ist der Mock
// (SPEC-139 Nr. 1, 5, 11).
func TestM3WahlJeModus(t *testing.T) {
	ctx := context.Background()
	type zugaenge struct{ record, sessions, importer bool }
	for modus, soll := range map[string]zugaenge{"record": {true, true, true}, "replay": {}, "mock": {}} {
		m := neuM3Haken()
		k := mockKonfig(t.TempDir())
		k.Modus = modus
		if modus != "mock" {
			k.Dienste = []bootstrap.Dienst{dienstA("http://127.0.0.1:9/a")}
		}
		if modus != "record" { // nur der Modus record liest UPSTREAM_CA_FILE (SPEC-139 Nr. 2)
			k.UpstreamCAFile = filepath.Join(t.TempDir(), "keine-ca.pem")
		}
		hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
		if err != nil {
			t.Fatal(modus, err)
		}
		got := zugaenge{hd.Record() != nil, hd.Sessions() != nil, hd.Import() != nil}
		if got != soll || hd.Workspaces() == nil || hd.Snapshots() == nil || hd.Wartung() == nil {
			t.Errorf("%s: Zugänge %+v, erwartet %+v, Workspaces/Snapshots/Wartung gesetzt", modus, got, soll)
		}
		if len(m.snapshot) != 1 {
			t.Errorf("%s: NeuerSnapshot %d-mal", modus, len(m.snapshot))
		}
		if modus != "mock" {
			continue
		}
		if len(m.mock) != 1 || m.oeffnen != 1 {
			t.Errorf("mock: NeuerMock %d-mal, Ablage %d-mal geöffnet", len(m.mock), m.oeffnen)
		}
		if len(m.verboten) != 0 || len(verbotene()) != 7 {
			t.Errorf("mock: verbotene Konstruktoren gerufen: %v", m.verboten)
		}
		_ = hd.Herunterfahren(ctx)
	}
	// Der Listener ruft Mock mit dem ctx und der Anfrage des Aufrufs, und zwar nur Mock.
	m := neuM3Haken()
	m.fake = &fakeMock{}
	hd, err := bootstrap.BuildMit(ctx, mockKonfig(t.TempDir()), nil, m.bauteile())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = hd.Herunterfahren(ctx) }()
	if err := hd.Hochfahren(ctx); err != nil {
		t.Fatal(err)
	}
	type marke struct{}
	aufrufCtx := context.WithValue(ctx, marke{}, "ja")
	anfrage := inbound.Anfrage{Methode: "PATCH", Pfad: "/a/Products(1)", RawQuery: "x=1", KorrelationsID: "k-1", Laenge: 7}
	antwort, err := m.server.Verarbeiten(aufrufCtx, anfrage)
	if err != nil || antwort.Status != 204 || len(m.fake.anfragen) != 1 {
		t.Fatalf("Verarbeiten: %+v, %v, %d Aufrufe des Mock", antwort, err, len(m.fake.anfragen))
	}
	if !reflect.DeepEqual(m.fake.anfragen[0], anfrage) || m.fake.ctxs[0].Value(marke{}) != "ja" {
		t.Errorf("Mock erhielt %+v, ctx-Marke %v", m.fake.anfragen[0], m.fake.ctxs[0].Value(marke{}))
	}
}

// TestM3KonfigDesMock fängt A08, A09 und A11: jeder Wert der Konfiguration von Mock, Snapshot-Dienst und Workspace-Dienst erreicht den
// Verbraucher unverändert; der Satz der Version 4 ist von jedem Dienst getragen; ein Wächter, eine Ablage (SPEC-139 Nr. 3, 11).
func TestM3KonfigDesMock(t *testing.T) {
	ctx := context.Background()
	for _, strikt := range []bool{true, false} {
		m := neuM3Haken()
		k := mockKonfig(t.TempDir())
		k.StrictCoverage = strikt
		hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
		if err != nil {
			t.Fatal(err)
		}
		mk, sk, wk := m.mock[0], m.snapshot[0], *m.workspace
		// A08: Werte
		if mk.StrictCoverage != strikt || mk.PublicBaseURL != k.PublicBaseURL || mk.MaxBytes != k.MaxCaptureBody {
			t.Errorf("Werte: %v %q %d", mk.StrictCoverage, mk.PublicBaseURL, mk.MaxBytes)
		}
		if len(mk.Dienste) != 2 {
			t.Fatalf("Dienste %+v", mk.Dienste)
		}
		for i, d := range k.Dienste {
			if mk.Dienste[i].ID != d.ID || mk.Dienste[i].Root != d.Root {
				t.Errorf("Dienst %d: %+v", i, mk.Dienste[i])
			}
			// A09: der Satz, sechs Felder je Dienst
			s := mk.Dienste[i].Codecs
			if s == nil {
				t.Fatalf("Dienst %d ohne Satz", i)
			}
			typen := map[string][2]reflect.Type{
				"Metadaten": {reflect.TypeOf(s.Metadaten), reflect.TypeOf(csdl.New())},
				"Daten":     {reflect.TypeOf(s.Daten), reflect.TypeOf(jsoncodec.New())},
				"Schreiber": {reflect.TypeOf(s.Schreiber), reflect.TypeOf(jsoncodec.NewSchreiber())},
				"Eingabe":   {reflect.TypeOf(s.Eingabe), reflect.TypeOf(jsoncodec.NewEingabe())},
				"URL":       {reflect.TypeOf(s.URL), reflect.TypeOf(urlparser.New())},
			}
			// A10: der Leser des Filters hängt im Parser des Satzes
			if _, err := s.URL.ParseURL("/Products", "$filter=Id gt 1"); err != nil {
				t.Errorf("Dienst %d: $filter im Satz: %v", i, err)
			}
			for feld, tt := range typen {
				if tt[0] == nil || tt[0] != tt[1] {
					t.Errorf("Dienst %d, Feld %s: Typ %v, erwartet %v", i, feld, tt[0], tt[1])
				}
			}
			if len(typen) != 5 || s.Version != "4.0" || s.NurLesend {
				t.Errorf("Dienst %d: Satz %+v", i, *s)
			}
			if mk.MetadatenCodec != s.Metadaten || mk.URLCodec != s.URL || mk.Schreiber != s.Schreiber {
				t.Errorf("Dienst %d: Codecs der Konfig sind nicht die des Satzes", i)
			}
		}
		// A11: ein Wächter, eine Ablage
		if len(m.waechter) != 1 {
			t.Fatalf("Wächter %d-mal gebaut", len(m.waechter))
		}
		w := m.waechter[0]
		blobs := map[string]any{"Mock": mk.Blobs, "Snapshot": sk.Blobs, "Workspace": wk.Blobs, "Wartung": m.haken.wartung.Blobs}
		for name, b := range blobs {
			if b != any(w) {
				t.Errorf("Blobs von %s ist nicht der Wächter: %T", name, b)
			}
		}
		if len(blobs) != 4 {
			t.Errorf("%d Verbraucher der Blobs", len(blobs))
		}
		repos := map[string][2]any{
			"Workspaces mock/snapshot": {mk.Workspaces, sk.Workspaces}, "Entities mock/snapshot": {mk.Entities, sk.Entities},
			"Workspaces snapshot/workspace": {sk.Workspaces, wk.Workspaces}, "Snapshots snapshot/workspace": {sk.Snapshots, wk.Snapshots},
			"Transaktion mock/workspace": {mk.Transaktion, wk.Transaktion}, "Transaktion mock/Ablage": {mk.Transaktion, m.haken.db},
		}
		for name, paar := range repos {
			if paar[0] == nil || paar[0] != paar[1] {
				t.Errorf("%s: %T und %T", name, paar[0], paar[1])
			}
		}
		if _, ok := mk.Coverage.(*sqlite.CoverageRepository); !ok {
			t.Errorf("Coverage %T", mk.Coverage)
		}
		if sk.Uhr == nil || sk.IDs == nil || sk.Uhr != wk.Uhr || sk.IDs != wk.IDs {
			t.Errorf("Uhr und Generator des Snapshot-Dienstes sind nicht die des Workspace-Dienstes")
		}
		// A08: Metadaten ist das Repository der Anlage (der Haken protokolliert ein Create über diesen Wrapper)
		_ = mk.Metadaten.Create(ctx, model.Service{ID: "zz", ServiceRoot: "/zz", ODataVersion: "4.0"})
		if !slices.Contains(m.p.liste(), "dienst:zz") {
			t.Errorf("Metadaten von Mock ist nicht das Repository der Anlage: %v", m.p.liste())
		}
		_ = hd.Herunterfahren(ctx)
	}
}

// TestM3Lebenszyklus fängt A12: im Modus mock entsteht beim Hochfahren keine Session, Wartung, Anlage mit 4.0 und Listener laufen
// in dieser Reihenfolge, das Herunterfahren schließt mit SHUTDOWN_TIMEOUT und danach die Ablage (SPEC-139 Nr. 6, 7, 11 d).
func TestM3Lebenszyklus(t *testing.T) {
	ctx := context.Background()
	m := neuM3Haken()
	k := mockKonfig(t.TempDir())
	hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
	if err != nil {
		t.Fatal(err)
	}
	if err := hd.Hochfahren(ctx); err != nil {
		t.Fatal(err)
	}
	dienste, err := m.haken.db.Metadata().List(ctx)
	if err != nil || len(dienste) != 2 {
		t.Fatalf("Services %+v, %v", dienste, err)
	}
	for _, d := range dienste {
		if d.ODataVersion != "4.0" {
			t.Errorf("Service %s mit Version %q", d.ID, d.ODataVersion)
		}
	}
	if err := hd.Herunterfahren(ctx); err != nil {
		t.Fatal(err)
	}
	soll := []string{pAblage, pWartung, "dienst:a", "dienst:b", pListen, "anhalten", pSchliessen, pAblageZu}
	// Serve läuft in einer eigenen Goroutine; seine Ereignisse ordnet nur der Test von Herunterfahren (TestHerunterfahrenReihenfolge).
	if got := slices.DeleteFunc(m.p.liste(), func(e string) bool { return e == pServe || e == "serve-ende" }); !slices.Equal(got, soll) {
		t.Errorf("Ereignisse %v, erwartet %v", got, soll)
	}
	if fr := m.srv.frist(); !slices.Equal(fr, []time.Duration{k.ShutdownTimeout}) {
		t.Errorf("Frist von Schliessen %v, erwartet [%v]", fr, k.ShutdownTimeout)
	}
	if len(m.verboten) != 0 {
		t.Errorf("verbotene Konstruktoren gerufen: %v", m.verboten)
	}
}

func istSelektor(e ast.Expr, paket, name string) bool {
	s, ok := e.(*ast.SelectorExpr)
	if !ok || s.Sel.Name != name {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	return ok && id.Name == paket
}

// TestM3FilterLeserInJedemAufruf fängt A10: jeder Aufruf von urlparser.New im Quelltext des Root trägt WithFilter(NewFilterLeser()), und
// urlparser.New steht nirgends als Wert (SPEC-160 Nr. 8, SPEC-139 Nr. 11 b).
func TestM3FilterLeserInJedemAufruf(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) == 0 {
		t.Fatalf("Quelldateien: %v, %v", dateien, err)
	}
	aufrufe, nennungen := 0, 0
	for _, name := range dateien {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.SelectorExpr:
				if istSelektor(n, "urlparser", "New") {
					nennungen++
				}
			case *ast.CallExpr:
				if !istSelektor(n.Fun, "urlparser", "New") {
					return true
				}
				aufrufe++
				mit := false
				for _, a := range n.Args {
					w, ok := a.(*ast.CallExpr)
					if !ok || !istSelektor(w.Fun, "urlparser", "WithFilter") || len(w.Args) != 1 {
						continue
					}
					l, ok := w.Args[0].(*ast.CallExpr)
					mit = mit || ok && istSelektor(l.Fun, "urlparser", "NewFilterLeser") && len(l.Args) == 0
				}
				if !mit {
					t.Errorf("%s: urlparser.New ohne WithFilter(NewFilterLeser())", name)
				}
			}
			return true
		})
	}
	if aufrufe < 1 || aufrufe != nennungen {
		t.Errorf("%d Aufrufe, %d Nennungen von urlparser.New", aufrufe, nennungen)
	}
}

package bootstrap_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// fehlerMock ist der MockUseCase der Fälle des Modus hybrid: er antwortet jede Anfrage mit demselben Fehler und zählt seine Aufrufe.
type fehlerMock struct {
	mu      sync.Mutex
	err     error
	aufrufe int
}

func (f *fehlerMock) Mock(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.aufrufe++
	return inbound.Antwort{}, f.err
}

// mockFehler ist ein Fehler des Mock mit dem Code und dem Status, den der Mock dazu liefert.
func mockFehler(code model.Code, status int) error {
	return &inbound.Fehler{Code: code, Status: status}
}

// m4Haken setzt auf m3Haken auf: er legt die Konfiguration des Hybrid-Dienstes ab, und mit mockFehler ersetzt er den Mock durch einen,
// der jede Anfrage mit demselben Fehler beantwortet.
type m4Haken struct {
	*m3Haken
	hybrid     []hybrid.Konfig
	mockFehler *fehlerMock
}

func neuM4Haken() *m4Haken { return &m4Haken{m3Haken: neuM3Haken()} }

func (m *m4Haken) bauteile() bootstrap.Bauteile {
	b := m.m3Haken.bauteile()
	mockNeu, hybridNeu := b.NeuerMock, b.NeuerHybrid
	b.NeuerMock = func(k mock.Konfig) (inbound.MockUseCase, error) {
		u, err := mockNeu(k)
		if err == nil && m.mockFehler != nil {
			return m.mockFehler, nil
		}
		return u, err
	}
	b.NeuerHybrid = func(k hybrid.Konfig) (inbound.HybridUseCase, error) {
		m.hybrid = append(m.hybrid, k)
		return hybridNeu(k)
	}
	return b
}

// upstreamAufruf ist ein Request, den der Referenzdienst des Tests sah.
type upstreamAufruf struct {
	Methode, Pfad, Query string
	Body                 []byte
}

// zaehlUpstream ist der Upstream der Fälle: er zählt Verbindungen (StateNew) und hält jeden Request fest.
type zaehlUpstream struct {
	srv          *httptest.Server
	mu           sync.Mutex
	verbindungen int
	aufrufe      []upstreamAufruf
}

func neuerZaehlUpstream(t *testing.T) *zaehlUpstream {
	t.Helper()
	u := &zaehlUpstream{}
	u.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		u.mu.Lock()
		u.aufrufe = append(u.aufrufe, upstreamAufruf{Methode: r.Method, Pfad: r.URL.Path, Query: r.URL.RawQuery, Body: b})
		u.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"u":1}`))
	}))
	u.srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			u.mu.Lock()
			u.verbindungen++
			u.mu.Unlock()
		}
	}
	u.srv.Start()
	t.Cleanup(u.srv.Close)
	return u
}

// host ist der Eintrag für FALLBACK_ALLOW_HOSTS, ziel die UpstreamURL eines Dienstes.
func (u *zaehlUpstream) host() string { return u.srv.Listener.Addr().String() }
func (u *zaehlUpstream) ziel() string { return u.srv.URL + "/odata" }

func (u *zaehlUpstream) stand() (int, []upstreamAufruf) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.verbindungen, slices.Clone(u.aufrufe)
}

// hybridKonfig ist die Konfiguration des Modus hybrid für Build mit Literal.
func hybridKonfig(dir string, werte policy.Werte, dienste ...bootstrap.Dienst) bootstrap.Konfiguration {
	k := basisKonfig(dir, "hybrid", dienste...)
	k.PublicBaseURL, k.StrictCoverage, k.MaxCaptureBody, k.Fallback = "http://127.0.0.1:8080/api", true, 12345, werte
	return k
}

// m4Ergebnis ist, was der Client eines Aufrufs von Verarbeiten sieht.
type m4Ergebnis struct {
	status int
	header map[string][]string
	body   string
	err    error
}

func (e m4Ergebnis) code() string {
	if f, ok := e.err.(*inbound.Fehler); ok {
		return f.Code.String()
	}
	return ""
}

func (e m4Ergebnis) kennzeichnung() string {
	if v := e.header["X-Recorder-Fallback"]; len(v) == 1 {
		return v[0]
	}
	return ""
}

// m4Sende ruft Verarbeiten des Listeners, wie der Adapter es tut.
func m4Sende(ctx context.Context, t *testing.T, srv func(context.Context, inbound.Anfrage) (inbound.Antwort, error), methode, pfad string, body []byte, laenge int64) m4Ergebnis {
	t.Helper()
	a := inbound.Anfrage{Methode: methode, Pfad: pfad, Laenge: laenge, Header: map[string][]string{"Accept": {"application/json"}}}
	if body != nil {
		a.Body = bytes.NewReader(body)
	}
	antwort, err := srv(ctx, a)
	if err != nil {
		return m4Ergebnis{err: err}
	}
	var b []byte
	if antwort.Body != nil {
		b, _ = io.ReadAll(antwort.Body)
		_ = antwort.Body.Close()
	}
	return m4Ergebnis{status: antwort.Status, header: antwort.Header, body: string(b)}
}

// TestM4WahlJeModus fängt M01: Modus × Anfrage mit Auslöser (der Mock antwortet mit ODR-E6001): im Modus hybrid trägt die Antwort die
// Kennzeichnung des Upstream-Fallbacks und der Hybrid-Dienst ist der Verarbeiter, in den Modi mock und replay ist die Antwort nie eine
// Fallback-Antwort und der Hybrid-Dienst wird nie gebaut; im Modus hybrid sind Sessions(), Record() und Import() der Nullwert der
// Schnittstelle, und Mock, Replay, Upstream-Client und Telemetrie werden je einmal gebaut, Recorder, Pipeline, Session und Import nie
// (SPEC-139 Nr. 1, 5, 12 a).
func TestM4WahlJeModus(t *testing.T) {
	ctx := context.Background()
	up := neuerZaehlUpstream(t)
	type soll struct {
		kennzeichnung string
		verbindungen  int
		hybridBauten  int
	}
	for modus, s := range map[string]soll{"hybrid": {"upstream", 1, 1}, "mock": {"", 0, 0}, "replay": {"", 0, 0}} {
		t.Run(modus, func(t *testing.T) {
			vorher, _ := up.stand()
			m := neuM4Haken()
			m.mockFehler = &fehlerMock{err: mockFehler(model.CodeE6001, 501)}
			k := hybridKonfig(t.TempDir(), policy.Werte{Upstream: true, Hosts: []string{up.host()}}, bootstrap.Dienst{ID: "a", Root: "/a", UpstreamURL: up.ziel()})
			k.Modus = modus
			var hd *bootstrap.Handle
			err := ohnePanik(func() (e error) { hd, e = bootstrap.BuildMit(ctx, k, nil, m.bauteile()); return e })
			if err != nil {
				t.Fatalf("Build im Modus %s: %v", modus, err)
			}
			if err := ohnePanik(func() error { return hd.Hochfahren(ctx) }); err != nil {
				t.Fatalf("Hochfahren im Modus %s: %v", modus, err)
			}
			defer func() { _ = hd.Herunterfahren(ctx) }()
			e := m4Sende(ctx, t, m.server.Verarbeiten, "GET", "/a/Products", nil, 0)
			if modus == "hybrid" {
				if e.err != nil || e.status != 200 || e.kennzeichnung() != s.kennzeichnung || e.body != `{"u":1}` {
					t.Errorf("Antwort %+v", e)
				}
			} else if e.kennzeichnung() != "" || e.err == nil {
				t.Errorf("Modus %s: Antwort %+v trägt eine Kennzeichnung oder ist keine Fehlerantwort", modus, e)
			}
			nach, _ := up.stand()
			if nach-vorher != s.verbindungen {
				t.Errorf("Modus %s: %d neue Verbindungen, erwartet %d", modus, nach-vorher, s.verbindungen)
			}
			if len(m.hybrid) != s.hybridBauten {
				t.Errorf("Modus %s: NeuerHybrid %d-mal, erwartet %d", modus, len(m.hybrid), s.hybridBauten)
			}
			if modus != "hybrid" {
				return
			}
			if hd.Sessions() != nil || hd.Record() != nil || hd.Import() != nil || hd.Workspaces() == nil || hd.Snapshots() == nil || hd.Wartung() == nil {
				t.Errorf("Zugänge: Sessions %v, Record %v, Import %v", hd.Sessions(), hd.Record(), hd.Import())
			}
			if want := map[string]int{"upstream": 1, "telemetrie": 1, "replay": 1}; !maps.Equal(m.verboten, want) || len(m.mock) != 1 {
				t.Errorf("Konstruktoren %v, Mock %d-mal; erwartet %v und einmal", m.verboten, len(m.mock), want)
			}
			for _, n := range []string{"pipeline", "session", "recorder", "import"} {
				if m.verboten[n] != 0 {
					t.Errorf("Modus hybrid ruft den Konstruktor %s", n)
				}
			}
			hk := m.hybrid[0]
			if hk.Mock != inbound.MockUseCase(m.mockFehler) || hk.Replay == nil || hk.Upstream == nil || hk.Policy == nil {
				t.Errorf("hybrid.Konfig %+v", hk)
			}
		})
	}
}

// TestM4Lebenszyklus fängt M01 für Hochfahren und Herunterfahren: im Modus hybrid entsteht keine Session (weder im Hochfahren noch in der
// Ablage), die Reihenfolge der Schritte ist die des Modus mock, und Schliessen erhält SHUTDOWN_TIMEOUT (SPEC-139 Nr. 6, 7, 12 b, c).
func TestM4Lebenszyklus(t *testing.T) {
	ctx := context.Background()
	m := neuM4Haken()
	k := hybridKonfig(t.TempDir(), policy.Werte{}, bootstrap.Dienst{ID: "a", Root: "/a"}, bootstrap.Dienst{ID: "b", Root: "/b/c"})
	hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
	if err != nil {
		t.Fatalf("Build im Modus hybrid: %v", err)
	}
	if err := ohnePanik(func() error { return hd.Hochfahren(ctx) }); err != nil {
		t.Fatalf("Hochfahren im Modus hybrid (kein Session-Dienst, keine Session): %v", err)
	}
	dienste, err := m.db.Metadata().List(ctx)
	if err != nil || len(dienste) != 2 {
		t.Fatalf("Services %+v, %v", dienste, err)
	}
	for _, d := range dienste {
		if sessions, err := m.db.Sessions().List(ctx, d.ID); err != nil || len(sessions) != 0 || d.ODataVersion != "4.0" {
			t.Errorf("Service %s: %d Sessions, Version %q, %v; erwartet 0 und 4.0", d.ID, len(sessions), d.ODataVersion, err)
		}
	}
	if err := hd.Herunterfahren(ctx); err != nil {
		t.Fatal(err)
	}
	soll := []string{pAblage, pWartung, "dienst:a", "dienst:b", pListen, "anhalten", pSchliessen, pAblageZu}
	if got := slices.DeleteFunc(m.p.liste(), func(e string) bool { return e == pServe || e == "serve-ende" }); !slices.Equal(got, soll) {
		t.Errorf("Ereignisse %v, erwartet %v", got, soll)
	}
	if fr := m.srv.frist(); !slices.Equal(fr, []time.Duration{k.ShutdownTimeout}) {
		t.Errorf("Frist von Schliessen %v, erwartet [%v]", fr, k.ShutdownTimeout)
	}
}

// m4Fall ist eine Anfrage mit dem Fehler des Mock und dem, was der Referenzdienst und der Client danach sehen.
type m4Fall struct {
	werte    func(u *zaehlUpstream) policy.Werte
	methode  string
	pfad     string
	body     []byte
	code     model.Code
	status   int
	wantVerb int
	// wantCode ist der Code der Fehlerantwort, leer bei einer Antwort mit Kennzeichnung upstream.
	wantCode string
}

// TestM4WerteDerPolicy fängt M02 (SPEC-261 Nr. 10 b): jeder der vier Schlüssel × {gesetzt, ungesetzt} erreicht policy.New und wirkt am
// Referenzdienst; das Orakel ist SPEC-261 Nr. 4 (Hauptschalter, Hosts, Routenfreigabe, Replay-Kategorie vor Upstream), nicht der Code der
// Policy. Die Zählprüfung hält die acht Paare gegen die Menge der Schlüssel; die Fälle ohne Paar sind Zusatz.
func TestM4WerteDerPolicy(t *testing.T) {
	ctx := context.Background()
	const lesen, schreiben = "/a/Products", "/a/Orders"
	basis := func(u *zaehlUpstream) policy.Werte { return policy.Werte{Upstream: true, Hosts: []string{u.host()}} }
	mit := func(f func(w *policy.Werte)) func(u *zaehlUpstream) policy.Werte {
		return func(u *zaehlUpstream) policy.Werte { w := basis(u); f(&w); return w }
	}
	faelle := []struct {
		schluessel, zustand string
		m4Fall
	}{
		{"HYBRID_FALLBACK", "gesetzt", m4Fall{werte: basis, methode: "GET", pfad: lesen, code: model.CodeE6003, status: 422, wantVerb: 1}},
		{"HYBRID_FALLBACK", "ungesetzt", m4Fall{werte: mit(func(w *policy.Werte) { w.Upstream = false }), methode: "GET", pfad: lesen,
			code: model.CodeE6003, status: 422, wantCode: "ODR-E6003"}},
		{"FALLBACK_ALLOW_HOSTS", "gesetzt", m4Fall{werte: basis, methode: "GET", pfad: lesen, code: model.CodeE6001, status: 501, wantVerb: 1}},
		{"FALLBACK_ALLOW_HOSTS", "ungesetzt", m4Fall{werte: mit(func(w *policy.Werte) { w.Upstream, w.Hosts = false, nil }), methode: "GET", pfad: lesen,
			code: model.CodeE6001, status: 501, wantCode: "ODR-E6001"}},
		{"FALLBACK_WRITE_ALLOW", "gesetzt", m4Fall{werte: mit(func(w *policy.Werte) { w.Schreibfreigaben = []string{"POST " + schreiben} }),
			methode: "POST", pfad: schreiben, body: []byte("x"), code: model.CodeE6002, status: 501, wantVerb: 1}},
		{"FALLBACK_WRITE_ALLOW", "ungesetzt", m4Fall{werte: basis, methode: "POST", pfad: schreiben, body: []byte("x"), code: model.CodeE6002,
			status: 501, wantCode: "ODR-E6002"}},
		{"HYBRID_REPLAY_FALLBACK", "gesetzt", m4Fall{werte: mit(func(w *policy.Werte) { w.ReplayKategorien = []string{"coverage_insufficient"} }),
			methode: "GET", pfad: lesen, code: model.CodeE6003, status: 422, wantCode: "ODR-E5003"}},
		{"HYBRID_REPLAY_FALLBACK", "ungesetzt", m4Fall{werte: basis, methode: "GET", pfad: lesen, code: model.CodeE6003, status: 422, wantVerb: 1}},
		// Zusatz: ein Host, der nicht der Zielhost ist, gibt nichts frei.
		{"FALLBACK_ALLOW_HOSTS", "zusatz: fremder Host", m4Fall{werte: mit(func(w *policy.Werte) { w.Hosts = []string{"fremd.test:1"} }),
			methode: "GET", pfad: lesen, code: model.CodeE6001, status: 501, wantCode: "ODR-E6001"}},
		// Zusatz: eine Freigabe für eine andere Methode gibt nichts frei.
		{"FALLBACK_WRITE_ALLOW", "zusatz: andere Methode", m4Fall{werte: mit(func(w *policy.Werte) { w.Schreibfreigaben = []string{"PUT " + schreiben} }),
			methode: "POST", pfad: schreiben, body: []byte("x"), code: model.CodeE6002, status: 501, wantCode: "ODR-E6002"}},
	}
	paare := map[string]bool{}
	for _, f := range faelle {
		if !strings.HasPrefix(f.zustand, "zusatz") {
			paare[f.schluessel+"/"+f.zustand] = true
		}
	}
	schluessel := []string{"HYBRID_FALLBACK", "FALLBACK_ALLOW_HOSTS", "FALLBACK_WRITE_ALLOW", "HYBRID_REPLAY_FALLBACK"}
	if len(paare) != 2*len(schluessel) {
		t.Fatalf("%d Paare aus Schlüssel und Zustand, erwartet %d", len(paare), 2*len(schluessel))
	}
	for _, s := range schluessel {
		for _, z := range []string{"gesetzt", "ungesetzt"} {
			if !paare[s+"/"+z] {
				t.Errorf("kein Fall für %s %s", s, z)
			}
		}
	}
	for _, f := range faelle {
		t.Run(f.schluessel+"/"+f.zustand, func(t *testing.T) {
			up := neuerZaehlUpstream(t)
			m := neuM4Haken()
			m.mockFehler = &fehlerMock{err: mockFehler(f.code, f.status)}
			k := hybridKonfig(t.TempDir(), f.werte(up), bootstrap.Dienst{ID: "a", Root: "/a", UpstreamURL: up.ziel()})
			hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			if err := hd.Hochfahren(ctx); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = hd.Herunterfahren(ctx) }()
			laenge := int64(len(f.body))
			e := m4Sende(ctx, t, m.server.Verarbeiten, f.methode, f.pfad, f.body, laenge)
			verb, aufrufe := up.stand()
			if verb != f.wantVerb || len(aufrufe) != f.wantVerb {
				t.Errorf("%d Verbindungen und %d Requests am Referenzdienst, erwartet %d", verb, len(aufrufe), f.wantVerb)
			}
			if f.wantCode != "" {
				if e.code() != f.wantCode {
					t.Errorf("Antwort %+v, erwartet Fehler %s", e, f.wantCode)
				}
				return
			}
			if e.err != nil || e.kennzeichnung() != "upstream" || e.body != `{"u":1}` {
				t.Errorf("Antwort %+v, erwartet die des Upstream mit Kennzeichnung", e)
			}
			if len(aufrufe) == 1 && (aufrufe[0].Methode != f.methode || aufrufe[0].Pfad != "/odata"+strings.TrimPrefix(f.pfad, "/a") || !bytes.Equal(aufrufe[0].Body, f.body)) {
				t.Errorf("Request am Referenzdienst %+v", aufrufe[0])
			}
		})
	}
}

// TestM4KonfigDesHybrid fängt M03 (SPEC-284 Nr. 1, 5 c, SPEC-139 Nr. 12 a): MaxBytes ist MAX_CAPTURE_BODY (der Default 8 MiB und ein gesetzter
// Wert), Body mit Länge unter, auf und über der Grenze bei bekannter Länge und bei Länge -1: bis zur Grenze Fallback, darüber nie;
// UpstreamURL je Dienst (ein Dienst ohne URL erreicht keinen Referenzdienst, ein anderer einen anderen); Upstream ist auch bei disabled gesetzt.
func TestM4KonfigDesHybrid(t *testing.T) {
	ctx := context.Background()
	for _, grenze := range []int64{12345, 8 << 20} {
		for _, art := range []string{"bekannte Länge", "Länge -1"} {
			for _, delta := range []int64{-1, 0, 1} {
				t.Run(fmt.Sprintf("Grenze %d/%s/%+d", grenze, art, delta), func(t *testing.T) {
					up := neuerZaehlUpstream(t)
					m := neuM4Haken()
					m.mockFehler = &fehlerMock{err: mockFehler(model.CodeE6002, 501)}
					k := hybridKonfig(t.TempDir(), policy.Werte{Upstream: true, Hosts: []string{up.host()}, Schreibfreigaben: []string{"POST /a/Orders"}},
						bootstrap.Dienst{ID: "a", Root: "/a", UpstreamURL: up.ziel()})
					k.MaxCaptureBody = grenze
					hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
					if err != nil {
						t.Fatal(err)
					}
					if err := hd.Hochfahren(ctx); err != nil {
						t.Fatal(err)
					}
					defer func() { _ = hd.Herunterfahren(ctx) }()
					if got := m.hybrid[0].MaxBytes; got != grenze {
						t.Errorf("MaxBytes %d, erwartet %d", got, grenze)
					}
					body := bytes.Repeat([]byte("b"), int(grenze+delta))
					laenge := int64(len(body))
					if art == "Länge -1" {
						laenge = -1
					}
					e := m4Sende(ctx, t, m.server.Verarbeiten, "POST", "/a/Orders", body, laenge)
					verb, aufrufe := up.stand()
					if delta <= 0 {
						if verb != 1 || e.err != nil || e.kennzeichnung() != "upstream" || len(aufrufe) != 1 || !bytes.Equal(aufrufe[0].Body, body) {
							t.Errorf("bis zur Grenze: %d Verbindungen, Antwort %+v", verb, e)
						}
						return
					}
					if verb != 0 || e.code() != "ODR-E6002" {
						t.Errorf("über der Grenze: %d Verbindungen, Antwort %+v, erwartet ODR-E6002 ohne Verbindung", verb, e)
					}
				})
			}
		}
	}
	t.Run("UpstreamURL je Dienst", func(t *testing.T) {
		u1, u2 := neuerZaehlUpstream(t), neuerZaehlUpstream(t)
		m := neuM4Haken()
		m.mockFehler = &fehlerMock{err: mockFehler(model.CodeE6003, 422)}
		k := hybridKonfig(t.TempDir(), policy.Werte{Upstream: true, Hosts: []string{u1.host(), u2.host()}},
			bootstrap.Dienst{ID: "a", Root: "/a", UpstreamURL: u1.ziel()}, bootstrap.Dienst{ID: "b", Root: "/b"},
			bootstrap.Dienst{ID: "c", Root: "/c", UpstreamURL: u2.ziel()})
		hd, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile())
		if err != nil {
			t.Fatal(err)
		}
		if err := hd.Hochfahren(ctx); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = hd.Herunterfahren(ctx) }()
		for _, f := range []struct {
			root    string
			wantErr bool
			u       *zaehlUpstream
		}{{"/a", false, u1}, {"/b", true, nil}, {"/c", false, u2}} {
			e := m4Sende(ctx, t, m.server.Verarbeiten, "GET", f.root+"/Products", nil, 0)
			if f.wantErr != (e.err != nil) || (f.wantErr && e.code() != "ODR-E6003") || (!f.wantErr && e.kennzeichnung() != "upstream") {
				t.Errorf("Dienst %s: Antwort %+v", f.root, e)
			}
		}
		if v1, a1 := u1.stand(); v1 != 1 || len(a1) != 1 || a1[0].Pfad != "/odata/Products" {
			t.Errorf("Referenzdienst von a: %d Verbindungen, %+v", v1, a1)
		}
		if v2, a2 := u2.stand(); v2 != 1 || len(a2) != 1 {
			t.Errorf("Referenzdienst von c: %d Verbindungen, %+v", v2, a2)
		}
		for i, d := range m.hybrid[0].Dienste {
			if want := []string{u1.ziel(), "", u2.ziel()}[i]; d.UpstreamURL != want || d.Root != k.Dienste[i].Root {
				t.Errorf("hybrid.Dienst %d: %+v", i, d)
			}
		}
	})
	t.Run("Upstream bei disabled", func(t *testing.T) {
		m := neuM4Haken()
		k := hybridKonfig(t.TempDir(), policy.Werte{}, bootstrap.Dienst{ID: "a", Root: "/a"})
		if _, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile()); err != nil {
			t.Fatalf("Build bei disabled: %v", err)
		}
		if len(m.hybrid) != 1 || m.hybrid[0].Upstream == nil || m.hybrid[0].Policy == nil || m.verboten["upstream"] != 1 {
			t.Errorf("hybrid.Konfig bei disabled: %+v, Upstream-Clients %d", m.hybrid, m.verboten["upstream"])
		}
	})
}

// TestM4BuildFehler fängt die Randformen von Build im Modus hybrid (SPEC-139 Nr. 2, 12 a): ohne PublicBaseURL ErrUngueltig, eine nicht lesbare
// UPSTREAM_CA_FILE ein Verstoß am Schlüssel, ein Verstoß der Policy ODR-E2001; die Ablage bleibt nicht offen.
func TestM4BuildFehler(t *testing.T) {
	ctx := context.Background()
	faelle := []struct {
		name    string
		aendere func(k *bootstrap.Konfiguration)
		pruefe  func(error) bool
	}{
		{"ohne PublicBaseURL", func(k *bootstrap.Konfiguration) { k.PublicBaseURL = "" }, istUngueltig},
		{"CA-Datei nicht lesbar", func(k *bootstrap.Konfiguration) { k.UpstreamCAFile = filepath.Join(t.TempDir(), "keine-ca.pem") }, func(e error) bool {
			return bootstrap.Meldung(e) == "ODR-E2001 UPSTREAM_CA_FILE: Datei nicht lesbar"
		}},
		{"enabled ohne Hosts", func(k *bootstrap.Konfiguration) { k.Fallback = policy.Werte{Upstream: true} }, func(e error) bool {
			return strings.HasPrefix(bootstrap.Meldung(e), "ODR-E2001 HYBRID_FALLBACK: enabled verlangt")
		}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			h := neuHaken()
			k := hybridKonfig(t.TempDir(), policy.Werte{}, bootstrap.Dienst{ID: "a", Root: "/a"})
			f.aendere(&k)
			if _, err := bootstrap.BuildMit(ctx, k, nil, h.bauteile()); err == nil || !f.pruefe(err) {
				t.Errorf("Fehler %v", err)
			}
			if h.p.zaehle(pAblage) != h.p.zaehle(pAblageZu) {
				t.Errorf("Ablage %d-mal geöffnet, %d-mal geschlossen", h.p.zaehle(pAblage), h.p.zaehle(pAblageZu))
			}
		})
	}
}

// TestM4UpstreamClientUndWarnung fängt M22 (SPEC-139 Nr. 4, 9, 12 a): Modus {record, replay, mock, hybrid} × HYBRID_FALLBACK {disabled,
// enabled} × UPSTREAM_TLS_INSECURE_SKIP_VERIFY {aus, true}: die Einträge tls_pruefung_aus zählt der Test; das Orakel ist die Spezifikation
// (record bei Prüfung aus, hybrid bei enabled und Prüfung aus: genau einer, sonst keiner). Der Upstream-Client der Modi record und hybrid
// trägt die drei Werte (Frist, Prüfen, CA) aus der Umgebung, in replay und mock gibt es keinen.
func TestM4UpstreamClientUndWarnung(t *testing.T) {
	pem := selbstsigniertesPEM(t)
	ca := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(ca, pem, 0o600); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, modus := range []string{"record", "replay", "mock", "hybrid"} {
		for _, fb := range []string{"disabled", "enabled"} {
			for _, skip := range []bool{false, true} {
				n++
				t.Run(fmt.Sprintf("%s/%s/skip=%v", modus, fb, skip), func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					h := neuHaken()
					extra := []string{"HYBRID_FALLBACK=" + fb, "FALLBACK_ALLOW_HOSTS=127.0.0.1:9", "UPSTREAM_TIMEOUT=7s", "UPSTREAM_CA_FILE=" + ca}
					if skip {
						extra = append(extra, "UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true")
					}
					code, stderr := fahre(ctx, t, cancel, umgebungFuer(t, modus, 1, extra...), h)
					if code != 0 {
						t.Fatalf("Rückgabe %d, stderr:\n%s", code, stderr)
					}
					zahl := 0
					for _, z := range logZeilen(t, stderr) {
						if z["ereignis"] == "tls_pruefung_aus" {
							zahl++
							if z["level"] != "WARN" || z["msg"] != "start: Zertifikatsprüfung zum Upstream abgeschaltet" {
								t.Errorf("Eintrag %v", z)
							}
						}
					}
					soll := 0
					if skip && (modus == "record" || modus == "hybrid" && fb == "enabled") {
						soll = 1
					}
					if zahl != soll {
						t.Errorf("%d Einträge tls_pruefung_aus, erwartet %d:\n%s", zahl, soll, stderr)
					}
					hatClient := modus == "record" || modus == "hybrid"
					if hatClient != (h.upstream != nil) {
						t.Fatalf("Upstream-Client %v im Modus %s", h.upstream, modus)
					}
					if hatClient && (h.upstream.Frist != 7*time.Second || h.upstream.Pruefen == skip || !bytes.Equal(h.upstream.CAPEM, pem)) {
						t.Errorf("Werte des Upstream-Clients %+v, erwartet Frist 7s, Prüfen %v, CA aus der Datei", *h.upstream, !skip)
					}
				})
			}
		}
	}
	if n != 16 {
		t.Errorf("%d Fälle, erwartet 4 × 2 × 2", n)
	}
}

// TestM4SatzImModusHybrid fängt M03 für den Mock im Modus hybrid: PUBLIC_BASE_URL, STRICT_COVERAGE, MAX_CAPTURE_BODY und ETAG_ENABLED erreichen ihn
// wie im Modus mock, und jeder Dienst trägt den Satz der Version 4 (SPEC-139 Nr. 11 a, b, 12 d).
func TestM4SatzImModusHybrid(t *testing.T) {
	ctx := context.Background()
	m := neuM4Haken()
	k := hybridKonfig(t.TempDir(), policy.Werte{}, bootstrap.Dienst{ID: "a", Root: "/a"}, bootstrap.Dienst{ID: "b", Root: "/b"})
	k.EtagEnabled = true
	if _, err := bootstrap.BuildMit(ctx, k, nil, m.bauteile()); err != nil {
		t.Fatal(err)
	}
	mk := m.mock[0]
	if mk.PublicBaseURL != k.PublicBaseURL || !mk.StrictCoverage || mk.MaxBytes != k.MaxCaptureBody || len(mk.Dienste) != 2 {
		t.Errorf("Konfig des Mock %+v", mk)
	}
	for i, d := range mk.Dienste {
		if !d.Etag || d.Codecs == nil || d.Codecs.Version != "4.0" || fmt.Sprintf("%T", d.Codecs.Metadaten) != fmt.Sprintf("%T", csdl.New()) {
			t.Errorf("Dienst %d des Mock: %+v", i, d)
		}
	}
}

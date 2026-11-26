package admin_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Der Netzwerkzugriffstest des Admin-Adapters (SPEC-041 Nr. 7, LH-QA-04): ein echter Loopback-Listener, admin.New mit den vier Routen-Konstruktoren und
// den Fakes des Pakets, ein http.Client gegen die Adresse. Alle Namen dieser Datei tragen das Präfix `sicherheit`.

// Die Orakel als Literale des Vertrags (SPEC-238 Nr. 3 c, 5; SPEC-109 Nr. 1), nie aus dem Code des Adapters gebildet, und die Zahlen der Mengen.
const (
	sicherheitUnbekannt = "/__admin/unbekannt"
	sicherheitObjekt401 = `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`
	sicherheitHealthz   = `{"status":"ok"}`
	sicherheitFrist     = 90 * time.Second
	sicherheitRouten    = 17  // 4 + 3 + 2 + 8
	sicherheitOps       = 20  // 7 + 3 + 2 + 8
	sicherheitZellen    = 504 // 18 Pfade, 7 Methoden, 4 Anlässe
)

func sicherheitMethoden() []string {
	return []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
}

func sicherheitAnlaesse() []string {
	return []string{"kein Header", "leerer Wert", "falsches Token", "richtiges Token unter Basic"}
}

// sicherheitRoutenJeKons sind die Zahlen der Routen je Konstruktor, sicherheitOpsJePort die der Operationen je Port (Workspaces mit Snapshots).
func sicherheitRoutenJeKons() [4]int { return [4]int{4, 3, 2, 8} }
func sicherheitOpsJePort() [4]int    { return [4]int{7, 3, 2, 8} }

// sicherheitZeile ist eine geschützte Zeile von SPEC-005: Methode, Muster, ein konkreter Pfad, die Abfrage und der Body einer gültigen Anfrage, der
// Konstruktor, der sie liefert, der Index ihrer Operation in der Summe der Zähler der Fakes und ob sie schreibt.
type sicherheitZeile struct {
	methode, muster, pfad, abfrage, body string
	konstruktor, zaehler                 int
	schreibend                           bool
}

// sicherheitZeilen sind die 17 Zeilen von SPEC-005 ohne `GET /healthz`, abgeschrieben aus der Tabelle der Spezifikation; die Anfänge der vier Ports in der
// Summe der Zähler sind 0, 7, 10 und 12.
func sicherheitZeilen() []sicherheitZeile {
	const (
		s  = 0
		i  = 7
		m  = 10
		w  = 12
		ws = "/__admin/workspaces"
	)
	return []sicherheitZeile{
		{"GET", "/__admin/sessions", "/__admin/sessions", "", "", 0, s + int(opList), false},
		{"POST", "/__admin/sessions", "/__admin/sessions", "", `{"service_id":"svc"}`, 0, s + int(opCreate), true},
		{"POST", "/__admin/sessions/{id}/close", "/__admin/sessions/ses-1/close", "", "", 0, s + int(opBeenden), true},
		{"DELETE", "/__admin/sessions/{id}", "/__admin/sessions/ses-1", "", "", 0, s + int(opDelete), true},
		{"GET", "/__admin/sessions/{id}/exchanges", "/__admin/sessions/ses-1/exchanges", "", "", 1, i + int(inspOpExchanges), false},
		{"GET", "/__admin/coverage", "/__admin/coverage", "baseline_id=bl-1", "", 1, i + int(inspOpCoverage), false},
		{"GET", "/__admin/diagnostics", "/__admin/diagnostics", "", "", 1, i + int(inspOpDiagnose), false},
		{"POST", "/__admin/imports", "/__admin/imports", "", gutBody, 2, m + int(impImport), true},
		{"GET", "/__admin/imports/{id}", "/__admin/imports/imp-1", "", "", 2, m + int(impGet), false},
		{"GET", ws, ws, "", "", 3, w + int(wsList), false},
		{"POST", ws, ws, "", `{"baseline_snapshot_id":"base-1","id":"ws-neu"}`, 3, w + int(wsCreate), true},
		{"GET", ws + "/{id}", ws + "/wsp-1", "", "", 3, w + int(wsGet), false},
		{"DELETE", ws + "/{id}", ws + "/wsp-1", "", "", 3, w + int(wsDelete), true},
		{"POST", ws + "/{id}/reset", ws + "/wsp-1/reset", "", "", 3, w + int(wsReset), true},
		{"GET", ws + "/{id}/snapshots", ws + "/wsp-1/snapshots", "", "", 3, w + int(snList), false},
		{"POST", ws + "/{id}/snapshots", ws + "/wsp-1/snapshots", "", `{"name":"nm-1"}`, 3, w + int(snCreate), true},
		{"POST", ws + "/{id}/restore", ws + "/wsp-1/restore", "", `{"snapshot_id":"snp-1"}`, 3, w + int(wsRestore), true},
	}
}

// sicherheitMelder zählt die roten Zellen einer Menge und meldet die ersten acht einzeln, den Rest als Summe.
type sicherheitMelder struct {
	t *testing.T
	n int
}

func (m *sicherheitMelder) rot(format string, args ...any) {
	m.t.Helper()
	m.n++
	if m.n <= 8 {
		m.t.Errorf(format, args...)
	}
}

func (m *sicherheitMelder) schluss(name string, zellen int) {
	m.t.Helper()
	if m.n > 0 {
		m.t.Errorf("%s: %d von %d Zellen rot", name, m.n, zellen)
	}
}

// sicherheitAntwort hält, was der Client sah.
type sicherheitAntwort struct {
	status int
	kopf   http.Header
	body   []byte
}

// sicherheitServer ist der Server unter Test: echter Loopback-Listener, vier Routen-Konstruktoren, die Fakes des Pakets.
type sicherheitServer struct {
	t           *testing.T
	adresse     string
	client      *http.Client
	audit       *supporttest.FakeAudit
	tele        *supporttest.FakeTelemetry
	sess        *sessionsFake
	insp        *inspektionFake
	imps        *importsFake
	wf          *workspaceFake
	sf          *snapshotFake
	p           *wsProtokoll
	token       string
	falsch      string
	ende        time.Time
	konstruktor [4][]admin.Route
}

// sicherheitTokenWert bildet ein Token von 64 Zeichen aus Bytewerten; saat verschiebt die Folge.
func sicherheitTokenWert(saat int) string {
	b := make([]byte, 0, 64)
	for i := range 32 {
		b = append(b, fmt.Sprintf("%02x", (i*37+saat*11+5)%256)...)
	}
	return string(b)
}

// sicherheitFakes baut die Fakes der vier Ports und die Routen der vier Konstruktoren über dem Audit.
func sicherheitFakes(t *testing.T, audit *supporttest.FakeAudit) *sicherheitServer {
	t.Helper()
	p := &wsProtokoll{}
	s := &sicherheitServer{
		t: t, audit: audit, tele: supporttest.NewFakeTelemetry(),
		sess: &sessionsFake{session: sessionA(), sessionen: []model.Session{sessionA()}},
		insp: &inspektionFake{},
		imps: &importsFake{ergebnis: inbound.ImportErgebnis{Import: importA(), Neu: true}, imp: importA()},
		wf:   &workspaceFake{p: p, ws: wsA(), bericht: inbound.Loeschbericht{Entfernt: 3, Zurueckgestellt: true}},
		sf:   &snapshotFake{p: p, snap: snapA()}, p: p,
		token: sicherheitTokenWert(1), falsch: sicherheitTokenWert(2),
	}
	var err [4]error
	s.konstruktor[0], err[0] = admin.SessionRouten(s.sess, audit)
	s.konstruktor[1], err[1] = admin.InspektionRouten(s.insp)
	s.konstruktor[2], err[2] = admin.ImportRouten(s.imps, audit)
	s.konstruktor[3], err[3] = admin.WorkspaceRouten(s.wf, s.sf, audit)
	for i, e := range err {
		if e != nil {
			t.Fatalf("Konstruktor %d: %v", i, e)
		}
	}
	return s
}

// sicherheitStart startet den Server auf 127.0.0.1:0 mit den Routen in der Folge von SPEC-300 Nr. 5 c (Sessions, Inspektion, Imports, Workspaces).
func sicherheitStart(t *testing.T) *sicherheitServer {
	t.Helper()
	s := sicherheitFakes(t, supporttest.NewFakeAudit())
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if ta, ok := l.Addr().(*net.TCPAddr); !ok || !ta.IP.IsLoopback() {
		t.Fatalf("der Listener liegt nicht am Loopback: %v", l.Addr())
	}
	srv, err := admin.New(admin.Konfig{
		Listener: l, Token: []byte(s.token), IDs: neueIDs(), Audit: s.audit, Telemetrie: s.tele, Routen: slices.Concat(s.konstruktor[:]...),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	go func() { _ = srv.Serve() }()
	s.adresse = l.Addr().String()
	s.client = &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{MaxIdleConnsPerHost: 4, DisableCompression: true}}
	s.ende = time.Now().Add(sicherheitFrist)
	t.Cleanup(func() {
		s.client.CloseIdleConnections()
		_ = srv.Schliessen(5 * time.Second)
	})
	return s
}

// anlass liefert den Wert des Headers Authorization des i-ten Anlasses (nil: ohne Header).
func (s *sicherheitServer) anlass(i int) *string {
	var v string
	switch i {
	case 0:
		return nil
	case 2:
		v = "Bearer " + s.falsch
	case 3:
		v = "Basic " + s.token
	}
	return &v
}

// sende sendet einen Request an den Server; die Frist des Tests und die des Clients begrenzen jeden Aufruf.
func (s *sicherheitServer) sende(methode, ziel string, auth *string, body string) sicherheitAntwort {
	s.t.Helper()
	if time.Now().After(s.ende) {
		s.t.Fatalf("die Frist von %v ist überschritten", sicherheitFrist)
	}
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(methode, "http://"+s.adresse+ziel, r)
	if err != nil {
		s.t.Fatal(err)
	}
	if auth != nil {
		req.Header["Authorization"] = []string{*auth}
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.t.Fatalf("%s %s: %v", methode, ziel, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		s.t.Fatalf("%s %s: Body: %v", methode, ziel, err)
	}
	return sicherheitAntwort{status: resp.StatusCode, kopf: resp.Header, body: b}
}

// zaehler liefert die Zähler aller Operationen der vier Ports in der Folge Sessions, Inspektion, Imports, Workspaces (mit Snapshots).
func (s *sicherheitServer) zaehler() []int {
	a, b, c, d := s.sess.zahlen(), s.insp.zahlen(), s.imps.zahlen(), s.p.zahlen()
	return slices.Concat(a[:], b[:], c[:], d[:])
}

// sicherheitPfade sind die 18 Pfade der Menge M-Abweisung: der unbekannte und der konkrete Pfad jeder der 17 Zeilen.
func sicherheitPfade() []string {
	pfade := []string{sicherheitUnbekannt}
	for _, z := range sicherheitZeilen() {
		pfade = append(pfade, z.pfad)
	}
	return pfade
}

// abweisungen sendet die Menge M-Abweisung einmal in der Folge Pfad, Methode, Anlass.
func (s *sicherheitServer) abweisungen() (ants []sicherheitAntwort, zellen int) {
	s.t.Helper()
	for _, pfad := range sicherheitPfade() {
		for _, m := range sicherheitMethoden() {
			for a := range sicherheitAnlaesse() {
				zellen++
				ants = append(ants, s.sende(m, pfad, s.anlass(a), ""))
			}
		}
	}
	return ants, zellen
}

// sicherheitPruefe401 hält eine Antwort gegen das 401 des Vertrags (SPEC-238 Nr. 5): Status, Body (bei HEAD leer), Content-Type, Content-Length,
// WWW-Authenticate, X-Correlation-Id, kein Allow. Das Ergebnis ist der Grund der Abweichung, sonst leer.
func sicherheitPruefe401(a sicherheitAntwort, methode string) string {
	soll := sicherheitObjekt401
	if methode == "HEAD" {
		soll = ""
	}
	switch {
	case a.status != 401:
		return fmt.Sprintf("Status %d, erwartet 401", a.status)
	case string(a.body) != soll:
		return fmt.Sprintf("Body %q, erwartet %q", a.body, soll)
	case !slices.Equal(a.kopf.Values("WWW-Authenticate"), []string{"Bearer"}):
		return fmt.Sprintf("WWW-Authenticate %q", a.kopf.Values("WWW-Authenticate"))
	case !slices.Equal(a.kopf.Values("Content-Type"), []string{"application/json"}):
		return fmt.Sprintf("Content-Type %q", a.kopf.Values("Content-Type"))
	case !slices.Equal(a.kopf.Values("Content-Length"), []string{strconv.Itoa(len(sicherheitObjekt401))}):
		return fmt.Sprintf("Content-Length %q", a.kopf.Values("Content-Length"))
	case len(a.kopf.Values("X-Correlation-Id")) != 1 || a.kopf.Get("X-Correlation-Id") == "":
		return fmt.Sprintf("X-Correlation-Id %q", a.kopf.Values("X-Correlation-Id"))
	case len(a.kopf.Values("Allow")) != 0:
		return fmt.Sprintf("Allow %q", a.kopf.Values("Allow"))
	}
	return ""
}

// sicherheitZelle benennt die i-te Zelle einer Menge M-Abweisung (Folge Pfad, Methode, Anlass).
func sicherheitZelle(i int) (pfad, methode, anlass string) {
	n, m := len(sicherheitAnlaesse()), len(sicherheitMethoden())
	return sicherheitPfade()[i/(m*n)], sicherheitMethoden()[i/n%m], sicherheitAnlaesse()[i%n]
}

// TestSicherheitRouten (TS01, M-Route): die Routen der vier Konstruktoren sind genau die 17 geschützten Zeilen von SPEC-005; jede steht im Literal, jede
// Zeile des Literals ist registriert, und New meldet keine Überlappung.
func TestSicherheitRouten(t *testing.T) {
	s := sicherheitFakes(t, supporttest.NewFakeAudit())
	zeilen := sicherheitZeilen()
	soll := map[string]int{}
	var jeKons [4]int
	for _, z := range zeilen {
		soll[z.methode+" "+z.muster] = z.konstruktor
		jeKons[z.konstruktor]++
	}
	if len(zeilen) != sicherheitRouten || len(soll) != sicherheitRouten || jeKons != sicherheitRoutenJeKons() {
		t.Fatalf("das Literal hat %d Zeilen, %d verschiedene, je Konstruktor %v; erwartet %d und %v", len(zeilen), len(soll), jeKons, sicherheitRouten, sicherheitRoutenJeKons())
	}
	registriert := map[string]bool{}
	for k, routen := range s.konstruktor {
		if len(routen) != sicherheitRoutenJeKons()[k] {
			t.Errorf("Konstruktor %d liefert %d Routen, erwartet %d", k, len(routen), sicherheitRoutenJeKons()[k])
		}
		for _, r := range routen {
			schluessel := r.Methode + " " + r.Muster
			if kons, da := soll[schluessel]; !da || kons != k {
				t.Errorf("Konstruktor %d registriert %q, das nicht in seinem Teil des Literals steht", k, schluessel)
			}
			registriert[schluessel] = true
		}
	}
	for schluessel := range soll {
		if !registriert[schluessel] {
			t.Errorf("die Zeile %q von SPEC-005 ist nicht registriert", schluessel)
		}
	}
	if alle := slices.Concat(s.konstruktor[:]...); len(alle) != sicherheitRouten {
		t.Errorf("die vier Konstruktoren liefern %d Routen, erwartet %d", len(alle), sicherheitRouten)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	k := admin.Konfig{Listener: l, Token: []byte(s.token), IDs: neueIDs(), Audit: s.audit, Telemetrie: s.tele, Routen: slices.Concat(s.konstruktor[:]...)}
	if _, err := admin.New(k); err != nil {
		t.Errorf("New meldet mit den Routen der vier Konstruktoren: %v", err)
	}
}

// TestSicherheitAbweisung (TS02, M-Abweisung): jede Zelle aus 18 Pfaden, 7 Methoden und 4 Anlässen ist 401 ODR-E2101 mit WWW-Authenticate: Bearer und
// X-Correlation-Id, nie 404 oder 405; die Menge läuft zweimal im selben Server (1008 Anfragen).
func TestSicherheitAbweisung(t *testing.T) {
	s := sicherheitStart(t)
	m := &sicherheitMelder{t: t}
	gesamt := 0
	for durchlauf := range 2 {
		ants, zellen := s.abweisungen()
		if zellen != sicherheitZellen || len(ants) != sicherheitZellen {
			t.Fatalf("Durchlauf %d: %d Zellen, erwartet %d", durchlauf+1, zellen, sicherheitZellen)
		}
		for i, a := range ants {
			gesamt++
			pfad, methode, anlass := sicherheitZelle(i)
			if grund := sicherheitPruefe401(a, methode); grund != "" {
				m.rot("Durchlauf %d, %s %s (%s): %s", durchlauf+1, methode, pfad, anlass, grund)
			}
		}
	}
	m.schluss("Abweisung", gesamt)
	if gesamt != 2*sicherheitZellen {
		t.Errorf("%d Anfragen, erwartet %d", gesamt, 2*sicherheitZellen)
	}
}

// TestSicherheitKeineAenderung (TS03, M-Port): nach der Menge der Abweisung, zweimal, hat kein Port einen Aufruf, jede Abweisung genau einen Eintrag
// admin_abgewiesen mit allen übrigen Feldern leer und genau ein CountError(ODR-E2101) und sonst kein Aufruf der Telemetrie. Die Zahl der Operationen je
// Fake gleicht der Zahl der Methoden seines Ports.
func TestSicherheitKeineAenderung(t *testing.T) {
	s := sicherheitStart(t)
	ports := [5]int{
		reflect.TypeFor[inbound.SessionUseCase]().NumMethod(), reflect.TypeFor[inbound.InspectionUseCase]().NumMethod(),
		reflect.TypeFor[inbound.ImportUseCase]().NumMethod(), reflect.TypeFor[inbound.WorkspaceUseCase]().NumMethod(), reflect.TypeFor[inbound.SnapshotUseCase]().NumMethod(),
	}
	fakes := [4]int{len(s.sess.zahlen()), len(s.insp.zahlen()), len(s.imps.zahlen()), len(s.p.zahlen())}
	if ports != [5]int{7, 3, 2, 6, 2} || fakes != sicherheitOpsJePort() || len(s.zaehler()) != sicherheitOps {
		t.Fatalf("Methoden der Ports %v, Zähler der Fakes %v (zusammen %d); erwartet [7 3 2 6 2], %v und %d", ports, fakes, len(s.zaehler()), sicherheitOpsJePort(), sicherheitOps)
	}
	var kennungen []string
	for range 2 {
		ants, _ := s.abweisungen()
		for _, a := range ants {
			kennungen = append(kennungen, a.kopf.Get("X-Correlation-Id"))
		}
	}
	for op, n := range s.zaehler() {
		if n != 0 {
			t.Errorf("Operation %d eines Ports wurde %d-mal gerufen, erwartet 0", op, n)
		}
	}
	if got := s.audit.Aufrufe(); got != len(kennungen) {
		t.Errorf("das Audit sah %d Aufrufe, erwartet %d (ein Append je Abweisung, kein anderer Aufruf)", got, len(kennungen))
	}
	m := &sicherheitMelder{t: t}
	eintraege := s.audit.Eintraege()
	if len(eintraege) != len(kennungen) {
		t.Fatalf("das Audit hält %d Einträge, erwartet %d", len(eintraege), len(kennungen))
	}
	for i, e := range eintraege {
		soll := support.Eintrag{Art: "admin_abgewiesen", ErrorCode: model.CodeE2101, KorrelationsID: kennungen[i]}
		if e != soll || e.ErrorCode.String() != "ODR-E2101" || kennungen[i] == "" {
			m.rot("Eintrag %d: %+v, erwartet %+v", i, e, soll)
		}
	}
	aufrufe := s.tele.Aufrufe()
	if len(aufrufe) != len(kennungen) {
		t.Fatalf("die Telemetrie sah %d Aufrufe, erwartet %d", len(aufrufe), len(kennungen))
	}
	for i, a := range aufrufe {
		if a.Op != supporttest.OpCountError || a.Code.String() != "ODR-E2101" {
			m.rot("Aufruf %d der Telemetrie: Operation %d, Code %v, erwartet CountError(ODR-E2101)", i, a.Op, a.Code)
		}
	}
	m.schluss("Keine Änderung", len(kennungen))
}

// sicherheitKontrolle ist die Antwort auf die gültige Anfrage einer Zeile mit dem richtigen Token und die Zähler der Ports davor und danach.
type sicherheitKontrolle struct {
	zeile         sicherheitZeile
	antwort       sicherheitAntwort
	davor, danach []int
}

// kontrollLauf sendet je Zeile von SPEC-005 eine gültige Anfrage mit dem richtigen Token unter Bearer.
func (s *sicherheitServer) kontrollLauf() []sicherheitKontrolle {
	s.t.Helper()
	var aus []sicherheitKontrolle
	bearer := "Bearer " + s.token
	for _, z := range sicherheitZeilen() {
		k := sicherheitKontrolle{zeile: z, davor: s.zaehler()}
		ziel := z.pfad
		if z.abfrage != "" {
			ziel += "?" + z.abfrage
		}
		k.antwort = s.sende(z.methode, ziel, &bearer, z.body)
		k.danach = s.zaehler()
		aus = append(aus, k)
	}
	return aus
}

// TestSicherheitKontrolle (TS04, M-Kontrolle-Admin): mit dem richtigen Token erreicht jede der 17 Routen ihren Handler (nie 401, 404 oder 405), und genau
// die Operation ihres Ports zählt einen Aufruf; die 9 schreibenden Routen stehen im Literal.
func TestSicherheitKontrolle(t *testing.T) {
	s := sicherheitStart(t)
	m := &sicherheitMelder{t: t}
	ks := s.kontrollLauf()
	schreibend := 0
	for _, k := range ks {
		name := k.zeile.methode + " " + k.zeile.muster
		if k.zeile.schreibend {
			schreibend++
		}
		if st := k.antwort.status; st == 401 || st == 404 || st == 405 {
			m.rot("%s: Status %d mit dem richtigen Token", name, st)
		}
		for op := range k.davor {
			soll := 0
			if op == k.zeile.zaehler {
				soll = 1
			}
			if k.danach[op]-k.davor[op] != soll {
				m.rot("%s: Operation %d zählt %d Aufrufe, erwartet %d", name, op, k.danach[op]-k.davor[op], soll)
			}
		}
	}
	m.schluss("Kontrolle", len(ks))
	if len(ks) != sicherheitRouten || schreibend != 9 {
		t.Errorf("%d Zeilen mit %d schreibenden, erwartet %d und 9", len(ks), schreibend, sicherheitRouten)
	}
}

// TestSicherheitHealthz (TS05, M-Health): nur GET /healthz ist ohne Token 200 mit dem Body des Vertrags, ohne Audit und ohne Zählung; jede andere Zelle aus
// 4 Pfaden, 3 Methoden und 2 Anlässen ist 401 wie in der Abweisung, mit genau einem Eintrag und einer Zählung.
func TestSicherheitHealthz(t *testing.T) {
	s := sicherheitStart(t)
	m := &sicherheitMelder{t: t}
	zellen, ok200 := 0, 0
	for _, pfad := range []string{"/healthz", "/healthz/", "/HEALTHZ", "/healthz/x"} {
		for _, methode := range []string{"GET", "POST", "HEAD"} {
			for _, a := range []int{0, 2} {
				zellen++
				audit, tele := len(s.audit.Eintraege()), len(s.tele.Aufrufe())
				ant := s.sende(methode, pfad, s.anlass(a), "")
				audit, tele = len(s.audit.Eintraege())-audit, len(s.tele.Aufrufe())-tele
				name := fmt.Sprintf("%s %s (%s)", methode, pfad, sicherheitAnlaesse()[a])
				if pfad == "/healthz" && methode == "GET" {
					ok200++
					if ant.status != 200 || string(ant.body) != sicherheitHealthz || ant.kopf.Get("X-Correlation-Id") == "" || audit != 0 || tele != 0 {
						m.rot("%s: Status %d, Body %q, X-Correlation-Id %q, Einträge %d, Zählungen %d", name, ant.status, ant.body, ant.kopf.Get("X-Correlation-Id"), audit, tele)
					}
					continue
				}
				if grund := sicherheitPruefe401(ant, methode); grund != "" || audit != 1 || tele != 1 {
					m.rot("%s: %s (Einträge %d, Zählungen %d)", name, grund, audit, tele)
				}
			}
		}
	}
	m.schluss("Healthz", zellen)
	if zellen != 24 || ok200 != 2 {
		t.Errorf("%d Zellen, davon %d mit 200; erwartet 24 und 2", zellen, ok200)
	}
}

// sicherheitWegTexte rendert die vier Ausgabewege des Adapters als Text: Bodies, Header, jeden Eintrag des Audits mit allen Feldern, jeden Aufruf der
// Telemetrie mit allen Argumenten.
func sicherheitWegTexte(ants []sicherheitAntwort, es []support.Eintrag, as []supporttest.Aufruf) [4]string {
	var body, kopf, audit, tele strings.Builder
	for _, a := range ants {
		body.Write(a.body)
		body.WriteByte('\n')
		_ = a.kopf.Write(&kopf)
	}
	for _, e := range es {
		fmt.Fprintf(&audit, "%+v\n", e)
	}
	for _, a := range as {
		fmt.Fprintf(&tele, "%+v\n", a)
	}
	return [4]string{body.String(), kopf.String(), audit.String(), tele.String()}
}

// sicherheitTokenSuche sucht je Weg, Token und Form (ganz, erste Hälfte, zweite Hälfte) in den Texten und liefert die Zahl der Suchzellen und die Funde.
func sicherheitTokenSuche(wege [4]string, token ...string) (zellen int, funde []string) {
	namen := [4]string{"Body", "Header", "Audit", "Telemetrie"}
	for w, text := range wege {
		for _, tok := range token {
			h := len(tok) / 2
			for f, teil := range [3]string{tok, tok[:h], tok[h:]} {
				zellen++
				if strings.Contains(text, teil) {
					funde = append(funde, fmt.Sprintf("%s, Form %d, Token %.6s", namen[w], f, tok))
				}
			}
		}
	}
	return zellen, funde
}

// TestSicherheitToken (TS06, M-Tokenweg): das Token des Servers und jedes falsche Token stehen in keinem Byte von Body, Header, Audit und Telemetrie, auch nicht
// als Hälfte, über alle Anfragen der Abweisung und der Kontrolle; die Kontrolle setzt das Token selbst in je einen der vier Wege ein, und die Suche findet es.
func TestSicherheitToken(t *testing.T) {
	s := sicherheitStart(t)
	ants, _ := s.abweisungen()
	for _, k := range s.kontrollLauf() {
		ants = append(ants, k.antwort)
	}
	if len(ants) != sicherheitZellen+sicherheitRouten {
		t.Errorf("%d Antworten, erwartet %d", len(ants), sicherheitZellen+sicherheitRouten)
	}
	zellen, funde := sicherheitTokenSuche(sicherheitWegTexte(ants, s.audit.Eintraege(), s.tele.Aufrufe()), s.token, s.falsch)
	if zellen != 24 {
		t.Errorf("%d Suchzellen, erwartet 24", zellen)
	}
	for _, f := range funde {
		t.Errorf("das Token steht im Ausgabeweg: %s", f)
	}
	for w := range 4 {
		var in [4]string
		switch w {
		case 0:
			in = sicherheitWegTexte([]sicherheitAntwort{{body: []byte("x" + s.token + "y")}}, nil, nil)
		case 1:
			in = sicherheitWegTexte([]sicherheitAntwort{{kopf: http.Header{"X-Kontrolle": {s.token}}}}, nil, nil)
		case 2:
			in = sicherheitWegTexte(nil, []support.Eintrag{{ServiceID: s.token}}, nil)
		default:
			f := supporttest.NewFakeTelemetry()
			f.Log(context.Background(), slog.LevelInfo, "kontrolle", support.Str("k", s.token))
			in = sicherheitWegTexte(nil, nil, f.Aufrufe())
		}
		_, gefunden := sicherheitTokenSuche(in, s.token)
		if len(gefunden) != 3 || !strings.HasPrefix(gefunden[0], []string{"Body", "Header", "Audit", "Telemetrie"}[w]) {
			t.Errorf("Kontrolle des Weges %d: die Suche findet %v, erwartet drei Funde in diesem Weg", w, gefunden)
		}
	}
}

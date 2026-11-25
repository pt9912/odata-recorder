package admin_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// wsZiel ist der Pfad der Routen für Workspaces (SPEC-298 Nr. 1), wsPfadID der Pfadwert, den die Tests der Routen mit `{id}`
// verwenden, wenn der Wert nicht selbst Gegenstand ist.
const (
	wsZiel   = "/__admin/workspaces"
	wsPfadID = "wsp-1"
)

// wsOp nennt die acht Operationen von inbound.WorkspaceUseCase (sechs) und inbound.SnapshotUseCase (zwei).
type wsOp int

const (
	wsList wsOp = iota
	wsCreate
	wsGet
	wsDelete
	wsReset
	wsRestore
	snList
	snCreate
	wsOpAnzahl
)

// wsAufruf ist ein Aufruf einer der acht Operationen mit den zwei Zeichenketten (nicht benutzte leer); der ctx des Aufrufs
// steht in ctxListe an derselben Stelle.
type wsAufruf struct {
	op   wsOp
	a, b string
}

// wsProtokoll hält die Aufrufe beider Fakes in der Folge ihres Eintreffens.
type wsProtokoll struct {
	mu       sync.Mutex
	aufrufe  []wsAufruf
	ctxs     []context.Context
	zaehlung [wsOpAnzahl]int
}

func (p *wsProtokoll) rufe(ctx context.Context, op wsOp, a, b string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.aufrufe = append(p.aufrufe, wsAufruf{op, a, b})
	p.ctxs = append(p.ctxs, ctx)
	p.zaehlung[op]++
}

// ctxListe liefert den ctx jedes Aufrufs als Kopie.
func (p *wsProtokoll) ctxListe() []context.Context {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.ctxs)
}

// zahlen liefert die Zahl der Aufrufe je Operation, gesamt die aller Operationen, liste die Aufrufe als Kopie.
func (p *wsProtokoll) zahlen() [wsOpAnzahl]int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.zaehlung
}

func (p *wsProtokoll) gesamt() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.aufrufe)
}

func (p *wsProtokoll) liste() []wsAufruf {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.aufrufe)
}

// zuruecksetzen leert das Protokoll (für Mengen, die einen Server für viele Requests benutzen).
func (p *wsProtokoll) zuruecksetzen() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.aufrufe, p.ctxs, p.zaehlung = nil, nil, [wsOpAnzahl]int{}
}

// nurDiagonale meldet jeden Zähler ungleich 1 bei op und ungleich 0 bei jeder anderen Operation.
func (p *wsProtokoll) nurDiagonale(t *testing.T, name string, op wsOp) {
	t.Helper()
	for o, n := range p.zahlen() {
		soll := 0
		if wsOp(o) == op {
			soll = 1
		}
		if n != soll {
			t.Errorf("%s: Operation %d wurde %d-mal gerufen, erwartet %d", name, o, n, soll)
		}
	}
}

// wsErgebnis ist, was eine Operation von workspaceFake liefert: je Operation zählt der Teil, den ihr Port zurückgibt.
type wsErgebnis struct {
	ws      model.Workspace
	liste   []model.Workspace
	bericht inbound.Loeschbericht
	err     error
}

// workspaceFake erfüllt inbound.WorkspaceUseCase: jede der sechs Operationen zählt ihren Aufruf und hält ctx und beide
// Argumente fest; Ergebnis und Fehler sind einstellbar (ws, liste, bericht, err) und kommen zugleich zurück, ergebnis ersetzt sie
// je Aufruf (der Zustand mit eigener Sperre liegt im Test). vor und nach sind je Operation Haltepunkte ohne Sperre des Fakes: vor
// läuft nach dem Zählen und vor dem Ergebnis, nach dahinter und vor der Rückkehr. Der Fake gerät nie in Panik.
type workspaceFake struct {
	p         *wsProtokoll
	ws        model.Workspace
	liste     []model.Workspace
	bericht   inbound.Loeschbericht
	err       error
	ergebnis  func(ctx context.Context, op wsOp, a, b string) wsErgebnis
	vor, nach [wsOpAnzahl]func(ctx context.Context)
}

func (f *workspaceFake) lauf(ctx context.Context, op wsOp, a, b string) wsErgebnis {
	f.p.rufe(ctx, op, a, b)
	if h := f.vor[op]; h != nil {
		h(ctx)
	}
	r := wsErgebnis{f.ws, f.liste, f.bericht, f.err}
	if f.ergebnis != nil {
		r = f.ergebnis(ctx, op, a, b)
	}
	if h := f.nach[op]; h != nil {
		h(ctx)
	}
	return r
}

func (f *workspaceFake) Create(ctx context.Context, baselineID, id string) (model.Workspace, error) {
	r := f.lauf(ctx, wsCreate, baselineID, id)
	return r.ws, r.err
}

func (f *workspaceFake) Get(ctx context.Context, id string) (model.Workspace, error) {
	r := f.lauf(ctx, wsGet, id, "")
	return r.ws, r.err
}

func (f *workspaceFake) List(ctx context.Context) ([]model.Workspace, error) {
	r := f.lauf(ctx, wsList, "", "")
	return r.liste, r.err
}

func (f *workspaceFake) Delete(ctx context.Context, id string) (inbound.Loeschbericht, error) {
	r := f.lauf(ctx, wsDelete, id, "")
	return r.bericht, r.err
}

func (f *workspaceFake) Reset(ctx context.Context, id string) (model.Workspace, error) {
	r := f.lauf(ctx, wsReset, id, "")
	return r.ws, r.err
}

func (f *workspaceFake) Restore(ctx context.Context, workspaceID, snapshotID string) (model.Workspace, error) {
	r := f.lauf(ctx, wsRestore, workspaceID, snapshotID)
	return r.ws, r.err
}

// snErgebnis ist, was eine Operation von snapshotFake liefert.
type snErgebnis struct {
	snap  model.WorkspaceSnapshot
	liste []model.WorkspaceSnapshot
	err   error
}

// snapshotFake erfüllt inbound.SnapshotUseCase wie workspaceFake für die zwei Operationen des Ports (ergebnis, vor und nach
// wie dort, indiziert mit snList und snCreate).
type snapshotFake struct {
	p         *wsProtokoll
	snap      model.WorkspaceSnapshot
	liste     []model.WorkspaceSnapshot
	err       error
	ergebnis  func(ctx context.Context, op wsOp, a, b string) snErgebnis
	vor, nach [wsOpAnzahl]func(ctx context.Context)
}

func (f *snapshotFake) lauf(ctx context.Context, op wsOp, a, b string) snErgebnis {
	f.p.rufe(ctx, op, a, b)
	if h := f.vor[op]; h != nil {
		h(ctx)
	}
	r := snErgebnis{f.snap, f.liste, f.err}
	if f.ergebnis != nil {
		r = f.ergebnis(ctx, op, a, b)
	}
	if h := f.nach[op]; h != nil {
		h(ctx)
	}
	return r
}

func (f *snapshotFake) Create(ctx context.Context, workspaceID, name string) (model.WorkspaceSnapshot, error) {
	r := f.lauf(ctx, snCreate, workspaceID, name)
	return r.snap, r.err
}

func (f *snapshotFake) List(ctx context.Context, workspaceID string) ([]model.WorkspaceSnapshot, error) {
	r := f.lauf(ctx, snList, workspaceID, "")
	return r.liste, r.err
}

// wsAuditNach ist der Audit der Routen: er reicht Append an den FakeAudit der Umgebung und ruft danach den Haltepunkt nach mit
// dem Eintrag (je Art ein Treffpunkt; er läuft in der Hülle, bevor Append zum Handler zurückkehrt, und umschließt nichts dahinter).
type wsAuditNach struct {
	support.AuditLog
	nach func(support.Eintrag)
}

func (a *wsAuditNach) Append(e support.Eintrag) error {
	err := a.AuditLog.Append(e)
	if a.nach != nil {
		a.nach(e)
	}
	return err
}

// wsEnv ist ein Server mit den Routen über beiden Fakes und dem FakeAudit der Umgebung (in der Hülle huelle); der Dienst liefert
// einen Workspace, einen Snapshot und einen Bericht (3 entfernt, zurückgestellt).
type wsEnv struct {
	*umgebung
	p      *wsProtokoll
	wf     *workspaceFake
	sf     *snapshotFake
	huelle *wsAuditNach
	routen []admin.Route
}

// wsNeu baut die Umgebung; ein Fehler von WorkspaceRouten oder New bricht den Test ab.
func wsNeu(t *testing.T) *wsEnv {
	t.Helper()
	k, u := konfigMit()
	p := &wsProtokoll{}
	e := &wsEnv{
		umgebung: u, p: p,
		wf: &workspaceFake{p: p, ws: wsA(), bericht: inbound.Loeschbericht{Entfernt: 3, Zurueckgestellt: true}},
		sf: &snapshotFake{p: p, snap: snapA()},
	}
	e.huelle = &wsAuditNach{AuditLog: u.audit}
	routen, err := admin.WorkspaceRouten(e.wf, e.sf, e.huelle)
	if err != nil || len(routen) != 8 {
		t.Fatalf("WorkspaceRouten: %d Routen, %v", len(routen), err)
	}
	k.Routen, e.routen = routen, routen
	neuMitKonfig(t, k, u)
	return e
}

// wsA ist ein Workspace des Dienstes, wsAJSON seine Antwort als Literal des Vertrags (SPEC-298 Nr. 7 a); snapA ist ein
// Snapshot mit Werten in den vier Feldern, die keine Antwort trägt, snapAJSON seine Antwort (Nr. 7 b).
func wsA() model.Workspace {
	return model.Workspace{ID: "wsp-1", ServiceID: "svc", BaselineSnapshotID: "base-1", Revision: 3, CreatedAt: zeitA()}
}

const wsAJSON = `{"id":"wsp-1","service_id":"svc","baseline_snapshot_id":"base-1","revision":3,"created_at":"2026-11-06T10:00:00.123456789Z"}`

func snapA() model.WorkspaceSnapshot {
	return model.WorkspaceSnapshot{
		ID: "snp-1", ServiceID: "svc-verborgen", WorkspaceID: "wsp-1", Name: "nm-1", CreatedAt: zeitA(),
		StateBlob: "blob-verborgen", StateHash: "hash-verborgen", MetadataHash: "meta-verborgen",
	}
}

const snapAJSON = `{"id":"snp-1","workspace_id":"wsp-1","name":"nm-1","created_at":"2026-11-06T10:00:00.123456789Z"}`

// Die Fehlerobjekte als Literale (SPEC-109 Nr. 2, SPEC-240) für die Codes, die ein Dienst der Workspaces liefert.
const (
	wsObjE7001 = `{"error":{"code":"ODR-E7001","message":"Workspace unknown","target":"workspace"}}`
	wsObjE7002 = `{"error":{"code":"ODR-E7002","message":"Snapshot unknown","target":"workspace"}}`
	wsObjE7006 = `{"error":{"code":"ODR-E7006","message":"Snapshot name already taken","target":"workspace"}}`
	wsObjE7007 = `{"error":{"code":"ODR-E7007","message":"Workspace id already taken","target":"workspace"}}`
	wsObjE7008 = `{"error":{"code":"ODR-E7008","message":"Snapshot kind does not fit the operation","target":"workspace"}}`
	wsObjE5004 = `{"error":{"code":"ODR-E5004","message":"Assignment belongs to another service","target":"replay"}}`
)

// wsRoute ist eine der acht Routen in der Folge von SPEC-298 Nr. 1: Methode, Muster mit `{id}`, Operation, Status des
// Erfolgs, `Allow` des Musters, der gültige Body (leer: die Route liest keinen) und die Argumente, die die Operation dazu
// bei Pfadwert wsPfadID bekommt, als Literale.
type wsRoute struct {
	name, methode, muster, allow string
	op                           wsOp
	status                       int
	body                         string
	args                         [2]string
}

// Indizes der acht Routen.
const (
	rList = iota
	rCreate
	rGet
	rDelete
	rReset
	rSnList
	rSnCreate
	rRestore
)

func wsRouten() []wsRoute {
	return []wsRoute{
		{"GET Liste", "GET", wsZiel, "GET, POST", wsList, 200, "", [2]string{"", ""}},
		{"POST anlegen", "POST", wsZiel, "GET, POST", wsCreate, 201, `{"baseline_snapshot_id":"base-1","id":"ws-neu"}`, [2]string{"base-1", "ws-neu"}},
		{"GET lesen", "GET", wsZiel + "/{id}", "DELETE, GET", wsGet, 200, "", [2]string{wsPfadID, ""}},
		{"DELETE", "DELETE", wsZiel + "/{id}", "DELETE, GET", wsDelete, 200, "", [2]string{wsPfadID, ""}},
		{"reset", "POST", wsZiel + "/{id}/reset", "POST", wsReset, 200, "", [2]string{wsPfadID, ""}},
		{"GET snapshots", "GET", wsZiel + "/{id}/snapshots", "GET, POST", snList, 200, "", [2]string{wsPfadID, ""}},
		{"POST snapshots", "POST", wsZiel + "/{id}/snapshots", "GET, POST", snCreate, 201, `{"name":"nm-1"}`, [2]string{wsPfadID, "nm-1"}},
		{"restore", "POST", wsZiel + "/{id}/restore", "POST", wsRestore, 200, `{"snapshot_id":"snp-1"}`, [2]string{wsPfadID, "snp-1"}},
	}
}

// pfad setzt id (unverändert, also kodiert) für `{id}` ein.
func (r wsRoute) pfad(id string) string { return strings.Replace(r.muster, "{id}", id, 1) }

// request bildet den Request der Route: id steht unverändert im Pfad, query hinter dem `?` (leer: ohne), kopf setzt Header.
func (r wsRoute) request(id, query string, kopf map[string][]string, body io.Reader) *http.Request {
	ziel := r.pfad(id)
	if query != "" {
		ziel += "?" + query
	}
	return anfrageMit(r.methode, ziel, kopf, body)
}

// gut bildet den Request mit dem gültigen Body der Route und dem Pfadwert wsPfadID.
func (r wsRoute) gut() *http.Request {
	var b io.Reader
	if r.body != "" {
		b = strings.NewReader(r.body)
	}
	return r.request(wsPfadID, "", nil, b)
}

// wsLeser liest aus Bytes, zählt jeden Aufruf von Read und liefert hinter den Bytes den Fehler rest (nil: das Ende); ende ist
// ein Treffpunkt, der einmal läuft, bevor der Leser nach den Bytes das Ende oder rest meldet.
type wsLeser struct {
	r       *bytes.Reader
	rest    error
	n       atomic.Int64
	ende    func()
	geendet bool
}

func wsNeuerLeser(b []byte, rest error) *wsLeser { return &wsLeser{r: bytes.NewReader(b), rest: rest} }

func (l *wsLeser) Read(p []byte) (int, error) {
	l.n.Add(1)
	n, err := l.r.Read(p)
	if err == io.EOF {
		if l.ende != nil && !l.geendet {
			l.geendet = true
			l.ende()
		}
		if l.rest != nil {
			err = l.rest
		}
	}
	return n, err
}

// wsBR ist eine der drei Routen mit Body (SPEC-298 Nr. 5): Index der Route, Pflichtschlüssel, alle erlaubten Schlüssel und
// ein Schlüssel einer anderen Route.
type wsBR struct {
	route      int
	pflicht    string
	schluessel []string
	fremd      string
}

func wsBodyRouten() []wsBR {
	return []wsBR{
		{rCreate, "baseline_snapshot_id", []string{"baseline_snapshot_id", "id"}, "name"},
		{rSnCreate, "name", []string{"name"}, "snapshot_id"},
		{rRestore, "snapshot_id", []string{"snapshot_id"}, "name"},
	}
}

// wsBForm ist eine Form des Bodys einer der drei Routen (Index in wsBodyRouten): die Bytes, die Header, der Fehler hinter den
// Bytes (nil: Ende), die angekündigte Länge (0: unverändert), bei gültiger Form die Argumente der Operation als Literale und
// ob die Form gültig ist.
type wsBForm struct {
	achse, name string
	br          int
	body        []byte
	kopf        map[string][]string
	rest        error
	laenge      int64
	gueltig     bool
	args        [2]string
}

func (f wsBForm) leser() io.Reader {
	if f.rest == nil {
		return bytes.NewReader(f.body)
	}
	return io.MultiReader(bytes.NewReader(f.body), errLeser{f.rest})
}

func (f wsBForm) route() wsRoute { return wsRouten()[wsBodyRouten()[f.br].route] }

func (f wsBForm) request() *http.Request {
	r := f.route().request(wsPfadID, "", f.kopf, f.leser())
	if f.laenge != 0 {
		r.ContentLength = f.laenge
	}
	return r
}

// wsGrosserKey liefert den Schlüssel mit großem ersten Buchstaben, wsEscKey den Schlüssel mit dem zweiten Zeichen als
// Escape (dekodiert gleich dem Schlüssel; aus Bytewerten).
func wsGrosserKey(k string) string { return strings.ToUpper(k[:1]) + k[1:] }

func wsEscKey(k string) string { return k[:1] + esc(fmt.Sprintf("%04x", k[1])) + k[2:] }

// wsUngueltigeBodys sind die 145 ungültigen Formen von M-Body ohne den Rand (SPEC-298 Nr. 5 a bis d, g): je Route Wertform 7,
// Struktur 7, Zeichen 3, Schlüssel 7 + 6 + 6, dazu Pflicht 4, Werte 20, Lesen 6 und Paare 45; s steht als Wert in jeder
// Form, die einen nennt.
func wsUngueltigeBodys(s string) []wsBForm {
	var out []wsBForm
	for i, br := range wsBodyRouten() {
		add := func(achse, name, body string) {
			out = append(out, wsBForm{achse: achse, name: name, br: i, body: []byte(body)})
		}
		p := `"` + br.pflicht + `":"` + s + `"`
		for _, kv := range [][2]string{{"leer", ""}, {"Leerraum", " \t\r\n"}, {"null", "null"}, {"Array", "[]"}, {"Zeichenkette", `"x"`}, {"Zahl", "1"}, {"Wahrheitswert", "true"}} {
			add("Wertform", kv[0], kv[1])
		}
		for _, kv := range [][2]string{
			{"offen", "{"}, {"nicht geschlossen", "{" + p}, {"Komma vor }", "{" + p + ",}"}, {"einfache Anführungszeichen", `{'` + br.pflicht + `':'` + s + `'}`},
			{"Text dahinter", "{" + p + "} x"}, {"Objekt dahinter", "{" + p + "}{}"}, {"Doppelpunkt fehlt", `{"` + br.pflicht + `" "` + s + `"}`},
		} {
			add("Struktur", kv[0], kv[1])
		}
		for _, kv := range [][2]string{
			{"BOM", string([]byte{0xEF, 0xBB, 0xBF}) + "{" + p + "}"},
			{"ungültiges UTF-8", `{"` + br.pflicht + `":"` + s + string([]byte{0xFF}) + `"}`},
			{"rohes NUL", `{"` + br.pflicht + `":"` + s + string([]byte{0x00}) + `"}`},
		} {
			add("Zeichen", kv[0], kv[1])
		}
		add("Schlüssel", "extra", "{"+p+`,"extra":"x"}`)
		add("Schlüssel", "Schreibweise", `{"`+wsGrosserKey(br.pflicht)+`":"`+s+`"}`)
		add("Schlüssel", "Schlüssel einer anderen Route", "{"+p+`,"`+br.fremd+`":"x"}`)
		for _, k := range br.schluessel {
			if k == br.pflicht {
				add("Schlüssel", "doppelt gleich "+k, "{"+p+","+p+"}")
				continue
			}
			add("Schlüssel", "doppelt gleich "+k, "{"+p+`,"`+k+`":"a","`+k+`":"a"}`)
		}
		add("Schlüssel", "Pflicht doppelt anders", "{"+p+`,"`+br.pflicht+`":"b"}`)
		add("Schlüssel", "Pflicht doppelt über Escape", "{"+p+`,"`+wsEscKey(br.pflicht)+`":"`+s+`"}`)
	}
	out = slices.Concat(out, wsUngueltigeBodys2(s))
	return out
}

// wsUngueltigeBodys2 sind die Achsen Pflicht, Werte, Lesen und Paare (75 Formen) von M-Body ungültig.
func wsUngueltigeBodys2(s string) []wsBForm {
	var out []wsBForm
	for i, br := range wsBodyRouten() {
		add := func(achse, name string, body []byte) {
			out = append(out, wsBForm{achse: achse, name: name, br: i, body: body})
		}
		add("Pflicht", "leeres Objekt", []byte("{}"))
		if len(br.schluessel) == 2 {
			add("Pflicht", "nur id", []byte(`{"id":"a"}`))
		}
		for _, k := range br.schluessel {
			for _, v := range []string{"null", "1", "true", "[]", "{}"} {
				glieder := []string{`"` + k + `":` + v}
				if k != br.pflicht {
					glieder = append([]string{`"` + br.pflicht + `":"` + s + `"`}, glieder...)
				}
				add("Werte", k+" "+v, []byte("{"+strings.Join(glieder, ",")+"}"))
			}
		}
		out = append(out,
			wsBForm{achse: "Lesen", name: "1048577 Bytes", br: i, body: auffuellen(`{"`+br.pflicht+`":"`+s+`"}`, bodyGrenze+1)},
			wsBForm{achse: "Lesen", name: "kürzer als Content-Length", br: i, body: []byte(`{"` + br.pflicht), rest: io.ErrUnexpectedEOF, laenge: 100})
		arten := []string{"Lesen", "Form", "Schlüssel", "Pflicht", "Typ", "Zeichen"}
		for a := range arten {
			for b := a + 1; b < len(arten); b++ {
				out = append(out, wsPaar(i, br, s, arten[a], arten[b]))
			}
		}
	}
	return out
}

// wsPaar bildet den Body mit genau den zwei Verstoßarten: Lesen (über 1 MiB), Form (Text hinter dem Objekt), Schlüssel (ein
// unbekannter), Pflicht (der Pflichtschlüssel fehlt), Typ (ein Wert ist keine Zeichenkette), Zeichen (BOM am Anfang). Bei den
// zwei Routen mit einem Schlüssel trägt Pflicht zusammen mit Typ denselben Body wie Typ: der Schlüssel steht dann mit dem
// falschen Typ (die Paare sind dort nicht trennbar).
func wsPaar(i int, br wsBR, s string, arten ...string) wsBForm {
	hat := func(a string) bool { return slices.Contains(arten, a) }
	gut := `"` + br.pflicht + `":"` + s + `"`
	var glieder []string
	switch {
	case hat("Typ") && len(br.schluessel) == 1:
		glieder = []string{`"` + br.pflicht + `":1`}
	case hat("Typ") && hat("Pflicht"):
		glieder = []string{`"id":1`}
	case hat("Typ"):
		glieder = []string{gut, `"id":1`}
	case !hat("Pflicht"):
		glieder = []string{gut}
	}
	if hat("Schlüssel") {
		glieder = append(glieder, `"extra":"x"`)
	}
	body := "{" + strings.Join(glieder, ",") + "}"
	if hat("Form") {
		body += " x"
	}
	if hat("Zeichen") {
		body = string([]byte{0xEF, 0xBB, 0xBF}) + body
	}
	b := []byte(body)
	if hat("Lesen") {
		b = auffuellen(body, bodyGrenze+1)
	}
	return wsBForm{achse: "Paar", name: strings.Join(arten, "+"), br: i, body: b}
}

// wsRand sind die 1536 Formen der Achse Rand: je Route jeder Bytewert 0 bis 255 vor und hinter einem gültigen Objekt; gültig
// sind genau die vier Zeichen des Leerraums (8 je Route).
func wsRand() []wsBForm {
	var out []wsBForm
	for i, br := range wsBodyRouten() {
		obj := []byte(`{"` + br.pflicht + `":"x"}`)
		for b := range 256 {
			ws := b == 0x20 || b == 0x09 || b == 0x0a || b == 0x0d
			for _, davor := range []bool{true, false} {
				body := append(slices.Clone(obj), byte(b))
				if davor {
					body = append([]byte{byte(b)}, obj...)
				}
				args := [2]string{wsPfadID, "x"}
				if i == 0 {
					args = [2]string{"x", ""}
				}
				out = append(out, wsBForm{
					achse: "Rand", name: "Byte " + strconv.Itoa(b) + " davor " + strconv.FormatBool(davor), br: i, body: body, gueltig: ws, args: args,
				})
			}
		}
	}
	return out
}

// wsFall ist eine Zelle von M-Fehler (SPEC-298 Nr. 8): der Fehler, ob der Dienst zugleich einen Wert liefert, und die
// Erwartung als Literale: abbruch heißt, die Verbindung wird abgebrochen (kein Zählen, keine Antwort), sonst Status, Code und
// Body; lesen heißt, der Fehler kommt vom Body des Requests, nicht vom Dienst.
type wsFall struct {
	route        int
	name         string
	err          error
	mitWert      bool
	abbruch      bool
	lesen        bool
	status       int
	code, objekt string
}

func wsE7001() *inbound.Fehler { return &inbound.Fehler{Code: model.CodeE7001, Status: 404} }

// wsAllgemein sind die 11 allgemeinen Ausgänge der Route r.
func wsAllgemein(r int) []wsFall {
	e := wsE7001()
	var nilZeiger *inbound.Fehler
	f := func(name string, err error, status int, code, objekt string) wsFall {
		return wsFall{route: r, name: name, err: err, status: status, code: code, objekt: objekt}
	}
	a := func(name string, err error) wsFall { return wsFall{route: r, name: name, err: err, abbruch: true} }
	out := []wsFall{
		f("ErrUngueltig", model.ErrUngueltig, 400, "ODR-E2102", objekt400),
		f("Fehler ohne Code", errors.New("x"), 500, "ODR-E9001", objekt500),
		f("Fehler der Ablage ohne Code", errors.New("nicht gefunden"), 500, "ODR-E9001", objekt500),
		a("Canceled", context.Canceled), a("DeadlineExceeded", context.DeadlineExceeded),
		a("Canceled neben E7001", errors.Join(context.Canceled, e)),
		f("verpackter E7001", fmt.Errorf("x: %w", e), 404, "ODR-E7001", wsObjE7001),
		f("Wert und E7001", e, 404, "ODR-E7001", wsObjE7001),
		f("nil-Zeiger", error(nilZeiger), 500, "ODR-E9001", objekt500),
		f("Join aus nil-Zeiger und E7001", errors.Join(error(nilZeiger), e), 404, "ODR-E7001", wsObjE7001),
		f("Join aus ErrUngueltig und E7001", errors.Join(model.ErrUngueltig, e), 404, "ODR-E7001", wsObjE7001),
	}
	out[7].mitWert = true
	return out
}

// wsCodes sind die Codes der Tabelle von SPEC-298 Nr. 8 je Route (28 Zellen).
func wsCodes(r int) []wsFall {
	c := func(code model.Code, status int, text, objekt string) wsFall {
		return wsFall{route: r, name: text, err: &inbound.Fehler{Code: code, Status: status}, status: status, code: text, objekt: objekt}
	}
	e3001, e9001 := c(model.CodeE3001, 500, "ODR-E3001", objektE3001), c(model.CodeE9001, 500, "ODR-E9001", objekt500)
	e7001 := c(model.CodeE7001, 404, "ODR-E7001", wsObjE7001)
	switch r {
	case rList:
		return []wsFall{e3001, e9001}
	case rCreate:
		return []wsFall{c(model.CodeE7002, 404, "ODR-E7002", wsObjE7002), c(model.CodeE7007, 409, "ODR-E7007", wsObjE7007), e3001, e9001}
	case rSnCreate:
		return []wsFall{e7001, c(model.CodeE7006, 409, "ODR-E7006", wsObjE7006), e3001, e9001}
	case rRestore:
		return []wsFall{
			e7001, c(model.CodeE7002, 404, "ODR-E7002", wsObjE7002), c(model.CodeE7008, 422, "ODR-E7008", wsObjE7008),
			c(model.CodeE5004, 409, "ODR-E5004", wsObjE5004), e3001, e9001,
		}
	}
	return []wsFall{e7001, e3001, e9001}
}

// wsLesen sind die 9 Zellen der Achse Lesen: drei Kontextfehler als Fehler des Bodys der drei Routen mit Body.
func wsLesen() []wsFall {
	var out []wsFall
	for _, br := range wsBodyRouten() {
		for _, e := range []struct {
			name string
			err  error
		}{
			{"Canceled", context.Canceled}, {"DeadlineExceeded", context.DeadlineExceeded},
			{"Canceled neben EOF", errors.Join(context.Canceled, io.ErrUnexpectedEOF)},
		} {
			out = append(out, wsFall{route: br.route, name: "Lesen " + e.name, err: e.err, abbruch: true, lesen: true})
		}
	}
	return out
}

// wsFehlerZellen sind die 125 Zellen von M-Fehler: 8 × 11 allgemeine Ausgänge, 28 Codes und 9 Zellen der Achse Lesen.
func wsFehlerZellen() []wsFall {
	var out []wsFall
	for r := range 8 {
		out = slices.Concat(out, wsAllgemein(r), wsCodes(r))
	}
	return slices.Concat(out, wsLesen())
}

// wsLaengen liefert die Längen von M-Länge zwischen von und bis: jede Länge, die eine Zweier- oder Zehnerpotenz um höchstens 1
// verfehlt (davor, auf ihr, dahinter), dazu die Längen von laengenDerWerte und bis, aufsteigend und ohne Doppelte; die Menge
// bildet der Test, nicht der Code der Route.
func wsLaengen(von, bis int) []int {
	menge := map[int]bool{bis: true}
	for _, n := range laengenDerWerte() {
		menge[n] = true
	}
	for _, basis := range []int{2, 10} {
		for p := 1; p <= 2*bis; p *= basis {
			menge[p-1], menge[p], menge[p+1] = true, true, true
		}
	}
	var out []int
	for n := range menge {
		if n >= von && n <= bis {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

// wsHoechstwert ist die größte Länge eines Wertes im Feld, die der Body noch fasst: bodyGrenze minus die Bytes des Bodys mit
// dem leeren Wert.
func wsHoechstwert(feld string) int {
	_, leer, _ := wsMitFeld(feld, "", "")
	return bodyGrenze - len(leer)
}

// wsLesefehler ist eine Art von M-Lesen: der Fehler hinter den Bytes des Bodys und was SPEC-298 Nr. 5 a aus ihm macht: `ende`
// (das Ende des Bodys, kein Fehler), `lese` (ein Lesefehler, der kein Kontextfehler ist: 400) oder `kontext` (ein
// Kontextfehler, auch verpackt: die Verbindung bricht ab).
type wsLesefehler struct {
	name string
	err  error
	art  string
}

// wsIsKontext ist ein Fehler, dessen Methode Is context.Canceled meldet, ohne ihn zu entpacken.
type wsIsKontext struct{}

func (wsIsKontext) Error() string      { return "Fehler mit Is" }
func (wsIsKontext) Is(ziel error) bool { return ziel == context.Canceled }

// wsLesenLaengen sind die vier Längen des Objekts von M-Lesen: der gültige Body der Route, 1 MiB - 1, 1 MiB und 1 MiB + 1.
func wsLesenLaengen(r wsRoute) []int {
	return []int{len(r.body), bodyGrenze - 1, bodyGrenze, bodyGrenze + 1}
}

// wsLesefehlerArten sind die 15 Arten von M-Lesen: das Ende, 8 Lesefehler und 6 Kontextfehler (auch in `errors.Join`, mit `%w`
// verpackt und als Fehler mit der Methode Is).
func wsLesefehlerArten() []wsLesefehler {
	lese := func(name string, e error) wsLesefehler { return wsLesefehler{name, e, "lese"} }
	kontext := func(name string, e error) wsLesefehler { return wsLesefehler{name, e, "kontext"} }
	return []wsLesefehler{
		{"EOF", io.EOF, "ende"},
		lese("ErrUnexpectedEOF", io.ErrUnexpectedEOF), lese("ErrClosedPipe", io.ErrClosedPipe), lese("ErrShortBuffer", io.ErrShortBuffer),
		lese("beliebiger Fehler", errors.New("boom")), lese("Frist des Netzes", os.ErrDeadlineExceeded),
		lese("Fehler des Netzes", &net.OpError{Op: "read", Net: "tcp", Err: errors.New("connection reset by peer")}),
		lese("MaxBytesError", &http.MaxBytesError{Limit: bodyGrenze}), lese("verpackter Lesefehler", fmt.Errorf("x: %w", io.ErrUnexpectedEOF)),
		kontext("Canceled", context.Canceled), kontext("DeadlineExceeded", context.DeadlineExceeded),
		kontext("Canceled und Lesefehler", errors.Join(context.Canceled, io.ErrUnexpectedEOF)),
		kontext("Lesefehler und DeadlineExceeded", errors.Join(io.ErrUnexpectedEOF, context.DeadlineExceeded)),
		kontext("verpacktes Canceled", fmt.Errorf("x: %w", context.Canceled)), kontext("Fehler mit Is", wsIsKontext{}),
	}
}

// wsLesenDaten liefert die Bytes vor dem Fehler an einer der vier Stellen von M-Lesen (0 keine, 1 das Objekt ohne sein letztes
// Byte, 2 das vollständige Objekt, 3 das Objekt und ein Leerzeichen); das Objekt ist der gültige Body der Route mit Leerraum vor
// der schließenden Klammer auf n Bytes.
func wsLesenDaten(r wsRoute, stelle, n int) []byte {
	obj := append(auffuellen(r.body[:len(r.body)-1], n-1), '}')
	return [][]byte{nil, obj[:n-1], obj, append(slices.Clone(obj), ' ')}[stelle]
}

// wsFremdeWerte sind die sieben Werte des fremden Schlüssels von M-Fremd (Literale des JSON): eine Zeichenkette, die leere
// Zeichenkette, `null`, eine Zahl, ein Wahrheitswert, ein Array und ein Objekt.
func wsFremdeWerte() []string { return []string{`"x"`, `""`, `null`, `1`, `true`, `[]`, `{}`} }

// wsFremde sind die Formen von M-Fremd (SPEC-298 Nr. 5 c): je Route mit Body und je Schlüssel einer anderen Route mit Body,
// der keiner dieser Route ist (Paar aus Route und fremdem Schlüssel), ein Body je Stelle des fremden Schlüssels (allein, vor
// allen erlaubten Schlüsseln der Route, zwischen zwei von ihnen und hinter allen: jede Stelle der Liste der erlaubten
// Schlüssel plus allein) und je Wert von wsFremdeWerte. Die Mengen entstehen aus wsBodyRouten, nicht aus einer Liste der Paare.
func wsFremde() []wsBForm {
	var out []wsBForm
	for i, br := range wsBodyRouten() {
		eigene := make([]string, len(br.schluessel))
		for k, s := range br.schluessel {
			eigene[k] = `"` + s + `":"x"`
		}
		for _, fremd := range wsFelder() {
			if slices.Contains(br.schluessel, fremd) {
				continue
			}
			for _, wert := range wsFremdeWerte() {
				f := `"` + fremd + `":` + wert
				out = append(out, wsBForm{achse: "Fremd", name: fremd + " allein " + wert, br: i, body: []byte("{" + f + "}")})
				for stelle := range len(eigene) + 1 {
					glieder := slices.Insert(slices.Clone(eigene), stelle, f)
					out = append(out, wsBForm{
						achse: "Fremd", name: fremd + " an Stelle " + strconv.Itoa(stelle) + " " + wert, br: i, body: []byte("{" + strings.Join(glieder, ",") + "}"),
					})
				}
			}
		}
	}
	return out
}

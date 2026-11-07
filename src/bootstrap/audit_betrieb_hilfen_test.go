package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/blobs"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// Aufgezeichnet ist ein Aufruf von Append samt der Konfig, deren Audit ihn empfing.
type Aufgezeichnet struct {
	Quelle  string
	Eintrag support.Eintrag
}

// AuditAufzeichnung ist der Decorator über der Instanz des Roots (SPEC-234 Nr. 8, SPEC-139 Nr. 3): er hält jeden
// Eintrag unverändert und ungeprüft fest, auch einen, den der Adapter abwiese, zählt jeden Aufruf und reicht ihn an
// die Instanz weiter, die die Konfig der Quelle trug; die Datei bleibt die Wirkung.
type AuditAufzeichnung struct {
	mu        sync.Mutex
	eintraege []Aufgezeichnet
}

type auditDurchreiche struct {
	rec    *AuditAufzeichnung
	quelle string
	ziel   support.AuditLog
}

// Append zeichnet auf und liefert das Ergebnis der Instanz unverändert.
func (d *auditDurchreiche) Append(e support.Eintrag) error {
	d.rec.mu.Lock()
	d.rec.eintraege = append(d.rec.eintraege, Aufgezeichnet{Quelle: d.quelle, Eintrag: e})
	d.rec.mu.Unlock()
	return d.ziel.Append(e)
}

// Umhuelle liefert das Audit der Konfig quelle, das aufzeichnet und an ziel weiterreicht.
func (a *AuditAufzeichnung) Umhuelle(quelle string, ziel support.AuditLog) support.AuditLog {
	return &auditDurchreiche{rec: a, quelle: quelle, ziel: ziel}
}

// Eintraege liefert eine Kopie der Aufrufe in der Folge ihres Eintretens.
func (a *AuditAufzeichnung) Eintraege() []Aufgezeichnet {
	a.mu.Lock()
	defer a.mu.Unlock()
	return slices.Clone(a.eintraege)
}

// SessionsMitFehlertext ist der Decorator über dem SessionRepository des Imports: er hält jedes Argument jeder
// Operation fest, zählt sie und liefert von Get den Fehler Fehler, wenn er gesetzt ist; sonst reicht er weiter.
type SessionsMitFehlertext struct {
	Fehler error
	mu     sync.Mutex
	ziel   outbound.SessionRepository
	ruf    map[string][]string
}

func (s *SessionsMitFehlertext) merke(op, arg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ruf == nil {
		s.ruf = map[string][]string{}
	}
	s.ruf[op] = append(s.ruf[op], arg)
}

// Alle liefert die Argumente der Aufrufe jeder Operation, die gerufen wurde.
func (s *SessionsMitFehlertext) Alle() map[string][]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	aus := map[string][]string{}
	for op, args := range s.ruf {
		aus[op] = slices.Clone(args)
	}
	return aus
}

// Aufrufe liefert die Argumente der Aufrufe der Operation op.
func (s *SessionsMitFehlertext) Aufrufe(op string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.ruf[op])
}

// Create reicht weiter.
func (s *SessionsMitFehlertext) Create(ctx context.Context, x model.Session) error {
	s.merke("Create", x.ID)
	return s.ziel.Create(ctx, x)
}

// Get liefert Fehler, wenn er gesetzt ist, sonst das Ergebnis des Ziels.
func (s *SessionsMitFehlertext) Get(ctx context.Context, id string) (model.Session, error) {
	s.merke("Get", id)
	if s.Fehler != nil {
		return model.Session{}, s.Fehler
	}
	return s.ziel.Get(ctx, id)
}

// SetStatus reicht weiter.
func (s *SessionsMitFehlertext) SetStatus(ctx context.Context, id string, st model.SessionStatus) error {
	s.merke("SetStatus", id)
	return s.ziel.SetStatus(ctx, id, st)
}

// List reicht weiter.
func (s *SessionsMitFehlertext) List(ctx context.Context, serviceID string) ([]model.Session, error) {
	s.merke("List", serviceID)
	return s.ziel.List(ctx, serviceID)
}

// FindOpen reicht weiter.
func (s *SessionsMitFehlertext) FindOpen(ctx context.Context) ([]model.Session, error) {
	s.merke("FindOpen", "")
	return s.ziel.FindOpen(ctx)
}

// Delete reicht weiter.
func (s *SessionsMitFehlertext) Delete(ctx context.Context, id string) error {
	s.merke("Delete", id)
	return s.ziel.Delete(ctx, id)
}

// geheimnis ist ein Wert, der in keinem Byte von audit.jsonl stehen darf, mit den Formen, in denen er dort stehen
// könnte.
type geheimnis struct {
	name   string
	formen [][]byte
}

// zugang sind Benutzer, Passwort und Query-Wert einer UpstreamURL (SPEC-139 Nr. 4); die Bytes entstehen aus
// Bytewerten (AGENTS.md §3.18).
type zugang struct {
	klasse                      string
	benutzer, passwort, anfrage []byte
}

// zugaenge liefert die zwei Klassen: formgültig (Bytes aus A-Z a-z 0-9 . _ -, die Kennungsform von SPEC-234 Nr. 3,
// sie kämen an der Formprüfung des Adapters vorbei) und formfremd (@ : / % ? und Bytes ab 0x80, in der URL
// prozent-kodiert).
func zugaenge() []zugang {
	return []zugang{
		{"gueltig", []byte("cnry-benutzer.G1"), []byte("cnry_passwort-G2"), []byte("cnry.query_G3-wert")},
		{"fremd", []byte{'b', 0xc3, 0xbc, '@', ':', '/', '%', '?', 'F', '1'}, []byte{'p', 0xc3, 0xa4, '@', '%', '2', '5', '/', '#', 'F', '2'}, []byte{'q', 0xe2, 0x82, 0xac, '&', '=', '%', '?', ' ', 'F', '3'}},
	}
}

// fehlertext ist der Text, den SessionsMitFehlertext im Fehler von Get trägt.
func fehlertext() []byte {
	return []byte{'F', 'e', 'h', 'l', 'e', 'r', 't', 'e', 'x', 't', '-', 0xc3, 0xa9, '-', 'Z', '9', ' ', 'g', 'e', 'h', 'e', 'i', 'm'}
}

func (z zugang) url(host string) string {
	u := url.URL{Scheme: "http", User: url.UserPassword(string(z.benutzer), string(z.passwort)), Host: host, Path: "/o",
		RawQuery: url.Values{"tok": {string(z.anfrage)}}.Encode()}
	return u.String()
}

// formenVon liefert roh, prozent-kodiert (Query, Pfadsegment, Userinfo), Hex (klein und groß), Base64 in den vier
// Alphabeten und Füllungen von encoding/base64 und, wenn paar gesetzt ist, dieselben Base64-Formen von
// benutzer:passwort. Weitere Kodierungen (Hash, ROT13, Teilstücke) sind die Prüfgrenze im Plan.
func formenVon(roh []byte, paar []byte) [][]byte {
	userinfo := strings.TrimSuffix(url.UserPassword(string(roh), "").String(), ":")
	fs := [][]byte{roh, []byte(url.QueryEscape(string(roh))), []byte(url.PathEscape(string(roh))), []byte(userinfo),
		[]byte(hex.EncodeToString(roh)), []byte(strings.ToUpper(hex.EncodeToString(roh)))}
	for _, quelle := range [][]byte{roh, paar} {
		if quelle == nil {
			continue
		}
		for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding} {
			fs = append(fs, []byte(enc.EncodeToString(quelle)))
		}
	}
	var aus [][]byte
	for _, f := range fs {
		if !slices.ContainsFunc(aus, func(g []byte) bool { return bytes.Equal(f, g) }) {
			aus = append(aus, f)
		}
	}
	return aus
}

// werte sind die drei Geheimnisse einer UpstreamURL: Benutzer, Passwort, Query-Wert.
func (z zugang) werte() []geheimnis {
	paar := append(append(slices.Clone(z.benutzer), ':'), z.passwort...)
	return []geheimnis{
		{z.klasse + "-benutzer", formenVon(z.benutzer, paar)},
		{z.klasse + "-passwort", formenVon(z.passwort, paar)},
		{z.klasse + "-query", formenVon(z.anfrage, nil)},
	}
}

// canaryWerte sind die sechs Werte der zwei Zugänge.
func canaryWerte() []geheimnis {
	var w []geheimnis
	for _, z := range zugaenge() {
		w = append(w, z.werte()...)
	}
	return w
}

// findet nennt die Form des Geheimnisses, die in daten steht, sonst den leeren Text.
func (g geheimnis) findet(daten []byte) string {
	for _, f := range g.formen {
		if bytes.Contains(daten, f) {
			return string(f)
		}
	}
	return ""
}

// canaryKonfig ist die Konfiguration mit zwei Diensten, deren UpstreamURL die zwei Zugänge tragen.
func canaryKonfig(dir, modus string, maxAge time.Duration) bootstrap.Konfiguration {
	z := zugaenge()
	k := basisKonfig(dir, modus,
		bootstrap.Dienst{ID: "x1", Root: "/r1", UpstreamURL: z[0].url("u1.test")},
		bootstrap.Dienst{ID: "x2", Root: "/r2", UpstreamURL: z[1].url("u2.test")})
	k.RetentionMaxAge = maxAge
	return k
}

// betrieb ist ein Prozess der Wurzel über einem Datenverzeichnis: echte Ablage, echte Dienste, echter Adapter; die
// Konfigs der vier Dienste tragen den Decorator.
type betrieb struct {
	t      *testing.T
	dir    string
	modus  string
	maxAge time.Duration
	rec    *AuditAufzeichnung
	sess   *SessionsMitFehlertext
	log    *puffer
	h      *bootstrap.Handle
}

func neuerBetrieb(t *testing.T, dir, modus string, maxAge time.Duration) *betrieb {
	t.Helper()
	return &betrieb{t: t, dir: dir, modus: modus, maxAge: maxAge, rec: &AuditAufzeichnung{}, sess: &SessionsMitFehlertext{}, log: &puffer{}}
}

func (b *betrieb) bauteile() bootstrap.Bauteile {
	bt := bootstrap.StandardBauteile()
	wartung, sessions, workspaces, importer := bt.NeueWartung, bt.NeueSession, bt.NeueWorkspaces, bt.NeuerImport
	bt.NeueWartung = func(k maintenance.Konfig) (inbound.MaintenanceUseCase, error) {
		k.Audit = b.rec.Umhuelle("wartung", k.Audit)
		return wartung(k)
	}
	bt.NeueSession = func(k session.Konfig) (inbound.SessionUseCase, error) {
		k.Audit = b.rec.Umhuelle("session", k.Audit)
		return sessions(k)
	}
	bt.NeueWorkspaces = func(k workspace.Konfig) (inbound.WorkspaceUseCase, error) {
		k.Audit = b.rec.Umhuelle("workspace", k.Audit)
		return workspaces(k)
	}
	bt.NeuerImport = func(k baselineimport.Konfig) (inbound.ImportUseCase, error) {
		k.Audit = b.rec.Umhuelle("import", k.Audit)
		b.sess.ziel = k.Sessions
		k.Sessions = b.sess
		return importer(k)
	}
	return bt
}

// baue ruft Build mit dem Logger des Betriebs; der Handle fährt am Ende des Tests herunter.
func (b *betrieb) baue() error {
	b.t.Helper()
	h, err := bootstrap.BuildMit(context.Background(), canaryKonfig(b.dir, b.modus, b.maxAge), slog.New(slog.NewTextHandler(b.log, nil)), b.bauteile())
	if err != nil {
		return err
	}
	b.t.Cleanup(func() { _ = h.Herunterfahren(context.Background()) })
	b.h = h
	return nil
}

// bauen ruft baue und bricht bei einem Fehler ab.
func (b *betrieb) bauen() {
	b.t.Helper()
	if err := b.baue(); err != nil {
		b.t.Fatalf("Build: %v", err)
	}
}

// datei ist der Stand von audit.jsonl zu einem Zeitpunkt.
type datei struct {
	b  []byte
	da bool
}

func liesAudit(t *testing.T, dir string) datei {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "audit.jsonl"))
	switch {
	case err == nil:
		return datei{b, true}
	case errors.Is(err, os.ErrNotExist):
		return datei{}
	}
	t.Fatalf("audit.jsonl nicht lesbar: %v", err)
	return datei{}
}

// saeen öffnet die Ablage des Datenverzeichnisses wie der Root, ruft f und schließt sie wieder; die Ausgangslage
// kommt so aus der Ablage, nie aus dem Dienst.
func saeen(t *testing.T, dir string, f func(d *wsDatei)) {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(dir, dateiDB), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	blobDir := filepath.Join(dir, verzeichnisB)
	if err := os.MkdirAll(blobDir, 0o700); err != nil {
		t.Fatal(err)
	}
	store, err := blobs.New(blobDir)
	if err != nil {
		t.Fatal(err)
	}
	f(&wsDatei{t: t, db: db, store: store})
}

func saeeDienst(d *wsDatei, id, root string) {
	d.t.Helper()
	if err := d.db.Metadata().Create(context.Background(), model.Service{ID: id, ServiceRoot: root, ODataVersion: "4.0"}); err != nil {
		d.t.Fatal(err)
	}
}

// saeeSession legt eine Session an, die alter alt ist; offen lässt sie offen, sonst ist sie complete.
func saeeSession(d *wsDatei, id, dienst string, alter time.Duration, offen bool) {
	d.t.Helper()
	ctx := context.Background()
	s := model.Session{ID: id, ServiceID: dienst, CreatedAt: time.Now().UTC().Add(-alter), Status: model.SessionOpen}
	if err := d.db.Sessions().Create(ctx, s); err != nil {
		d.t.Fatal(err)
	}
	if !offen {
		if err := d.db.Sessions().SetStatus(ctx, id, model.SessionComplete); err != nil {
			d.t.Fatal(err)
		}
	}
}

// saeeWorkspaces legt den Service svc, seine Session, die Baseline bsl-1 und je einen Workspace je Kennung an.
func saeeWorkspaces(d *wsDatei, ids ...string) {
	d.t.Helper()
	saeeDienst(d, "svc", "/")
	saeeSession(d, "ses-1", "svc", time.Hour, false)
	d.baseline("bsl-1", wsStandard())
	for _, id := range ids {
		w := model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-1", CreatedAt: wsT0()}
		if err := d.db.Workspaces().Create(context.Background(), w); err != nil {
			d.t.Fatal(err)
		}
	}
}

// besteht sagt, ob die Session oder der Workspace (Kennung mit ws-) in der Ablage steht.
func besteht(ctx context.Context, db *sqlite.DB, id string) bool {
	var err error
	if strings.HasPrefix(id, "ws-") {
		_, err = db.Workspaces().Get(ctx, id)
	} else {
		_, err = db.Sessions().Get(ctx, id)
	}
	return err == nil
}

// bestandVon liefert die Kennungen von ids, die nach dem Lauf in der Ablage stehen, sortiert.
func bestandVon(t *testing.T, dir string, ids []string) []string {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(dir, dateiDB), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var da []string
	for _, id := range ids {
		if besteht(ctx, db, id) {
			da = append(da, id)
		}
	}
	slices.Sort(da)
	return da
}

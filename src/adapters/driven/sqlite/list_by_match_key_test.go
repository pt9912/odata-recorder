package sqlite_test

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// stufe ist ein Kontext, der jeden Aufruf von Done und Err zählt; beim bei-ten läuft vorher, bevor der Aufruf antwortet.
// database/sql und der Treiber fragen einen offenen Kontext allein über Done ab, nie über Err; mit einem nil-Kanal starten sie
// keine Goroutine, jeder Aufruf liegt also in der Goroutine des Tests, zwischen zwei Schritten der Operation. Ist abbruch
// gesetzt, ist der Kontext ab dem bei-ten Aufruf beendet.
type stufe struct {
	mu      sync.Mutex
	bei, n  int
	vorher  func()
	abbruch bool
	ende    chan struct{}
}

func neueStufe(bei int, vorher func()) *stufe {
	return &stufe{bei: bei, vorher: vorher}
}

func (s *stufe) Deadline() (time.Time, bool) { return time.Time{}, false }
func (s *stufe) Value(any) any               { return nil }

func (s *stufe) schritt() {
	s.mu.Lock()
	s.n++
	jetzt := s.n == s.bei
	if jetzt && s.abbruch {
		s.ende = make(chan struct{})
		close(s.ende)
	}
	s.mu.Unlock()
	if jetzt && s.vorher != nil {
		s.vorher()
	}
}

func (s *stufe) Done() <-chan struct{} {
	s.schritt()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ende
}

func (s *stufe) Err() error {
	s.schritt()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ende != nil {
		return context.Canceled
	}
	return nil
}

func (s *stufe) aufrufe() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.n
}

// zweiAblagen ist eine Datei mit zwei Öffnern: der Leser und der Schreiber haben getrennte Pools.
func zweiAblagen(t *testing.T) (*sqlite.DB, *sqlite.DB) {
	t.Helper()
	pfad := filepath.Join(t.TempDir(), "lm.db")
	leser := open(t, pfad)
	schreiber := open(t, pfad)
	if _, err := leser.SQL().Exec(`INSERT INTO services(id, service_root, odata_version) VALUES ('d', '/', '4.0')`); err != nil {
		t.Fatal(err)
	}
	return leser, schreiber
}

const lmKey = "sha256:abc"

func lmExchange(id string, seq int, key string) model.Exchange {
	jetzt := time.Date(2026, 11, 27, 6, 0, 0, 0, time.UTC)
	return model.Exchange{
		ID: id, SessionID: "s", SequenceNo: seq, Method: "GET", Path: "/Products", MatchKey: key, StartedAt: jetzt,
		RequestHeaders: map[string][]string{"Accept": {"application/json"}}, RequestBody: model.BodyAbsent, ResponseStatus: 200,
		ResponseHeaders: map[string][]string{"Content-Type": {"application/json"}}, ResponseBody: model.BodyAbsent,
		Capture: model.CaptureComplete, RequestRedact: model.RedactionNone, ResponseRedact: model.RedactionNone,
		Integrity: model.IntegrityVerified, Equivalence: model.EquivalenceByte,
	}
}

// lmSession legt die offene Session s mit den Exchanges der Keys an, e<n> mit der Nummer n.
func lmSession(t *testing.T, d *sqlite.DB, keys ...string) {
	t.Helper()
	ctx := context.Background()
	s := model.Session{ID: "s", ServiceID: "d", CreatedAt: time.Date(2026, 11, 27, 6, 0, 0, 0, time.UTC), Status: model.SessionOpen}
	if err := d.Sessions().Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	for i, key := range keys {
		if err := d.Recording().Append(ctx, lmExchange(fmt.Sprintf("e%d", i+1), i+1, key)); err != nil {
			t.Fatal(err)
		}
	}
}

// lmStand ist das Ergebnis eines Aufrufs: die Kennungen der Treffer, die größte Nummer und ob die Session fehlte.
type lmStand struct {
	ids      string
	hoechste int
	fehlt    bool
}

func lmLies(ctx context.Context, t *testing.T, d *sqlite.DB) lmStand {
	t.Helper()
	l, n, err := d.Recording().ListByMatchKey(ctx, "s", lmKey)
	if errors.Is(err, outbound.ErrNichtGefunden) && len(l) == 0 && n == 0 {
		return lmStand{fehlt: true}
	}
	if err != nil {
		t.Fatalf("ListByMatchKey: %v", err)
	}
	var ids []string
	for _, e := range l {
		ids = append(ids, e.ID)
	}
	return lmStand{ids: strings.Join(ids, ","), hoechste: n}
}

// gleichzeitig hält an jedem Aufruf des Kontexts einer gelungenen Operation an, lässt dort schreibe über die zweite
// Verbindung laufen und prüft, dass das Ergebnis der Stand davor oder danach ist. Beide Stände kommen vor: die Haltepunkte
// liegen vor und nach der Wahl des Standes.
func gleichzeitig(t *testing.T, schreibe func(context.Context, *sqlite.DB) error, davor, danach lmStand) {
	t.Helper()
	leser, _ := zweiAblagen(t)
	lmSession(t, leser, lmKey, "sha256:fremd")
	probe := neueStufe(0, nil)
	if got := lmLies(probe, t, leser); got != davor {
		t.Fatalf("Probelauf %+v, erwartet %+v", got, davor)
	}
	gesehen := map[lmStand]int{}
	for k := 1; k <= probe.aufrufe(); k++ {
		leser, schreiber := zweiAblagen(t)
		lmSession(t, leser, lmKey, "sha256:fremd")
		var fehler error
		gelaufen := false
		got := lmLies(neueStufe(k, func() { gelaufen, fehler = true, schreibe(context.Background(), schreiber) }), t, leser)
		if !gelaufen || fehler != nil {
			t.Fatalf("Aufruf %d: Schreiber gelaufen %v, Fehler %v", k, gelaufen, fehler)
		}
		if got != davor && got != danach {
			t.Errorf("Aufruf %d von %d: gemischter Stand %+v, erwartet %+v oder %+v", k, probe.aufrufe(), got, davor, danach)
		}
		gesehen[got]++
	}
	if davor != danach && (gesehen[davor] == 0 || gesehen[danach] == 0) {
		t.Errorf("die Haltepunkte liegen nicht beidseits der Wahl des Standes: %v", gesehen)
	}
}

// TestListByMatchKeyGegenAppend hält SPEC-105 Nr. 14: ein gleichzeitiges Append ist in der Liste und in der größten Nummer
// ganz oder gar nicht enthalten, mit gleichem und mit fremdem Key.
func TestListByMatchKeyGegenAppend(t *testing.T) {
	haenge := func(key string) func(context.Context, *sqlite.DB) error {
		return func(ctx context.Context, d *sqlite.DB) error {
			return d.Recording().Append(ctx, lmExchange("e3", 3, key))
		}
	}
	t.Run("gleicher Key", func(t *testing.T) {
		gleichzeitig(t, haenge(lmKey), lmStand{ids: "e1", hoechste: 2}, lmStand{ids: "e1,e3", hoechste: 3})
	})
	t.Run("fremder Key", func(t *testing.T) {
		gleichzeitig(t, haenge("sha256:anders"), lmStand{ids: "e1", hoechste: 2}, lmStand{ids: "e1", hoechste: 3})
	})
}

// TestListByMatchKeyGegenDelete hält SPEC-105 Nr. 14: ein gleichzeitiges Delete der Session ist ganz oder gar nicht
// enthalten: der volle Stand oder ErrNichtGefunden, nie Treffer mit 0 und nie eine leere Liste ohne Fehler.
func TestListByMatchKeyGegenDelete(t *testing.T) {
	gleichzeitig(t, func(ctx context.Context, d *sqlite.DB) error { return d.Sessions().Delete(ctx, "s") },
		lmStand{ids: "e1", hoechste: 2}, lmStand{fehlt: true})
}

// abbruchStellen ist die Zahl der Aufrufe des Kontexts, die vor der Abfrage der Treffer liegen oder sie selbst sind:
// Verbindung, BEGIN, Abfrage der größten Nummer, deren Zeilen, Abfrage der Treffer. Die Zahl gehört zu database/sql und dem
// Treiber; der Probelauf bestätigt, dass mindestens so viele Aufrufe bestehen.
const abbruchStellen = 5

// TestListByMatchKeyAbbruch hält SPEC-105 Nr. 13 an jedem Schritt: endet der Kontext bei einem Aufruf, ist das Ergebnis der
// Fehler des Kontexts ohne Werte oder der volle Stand, nie eine verkürzte oder leere Liste ohne Fehler; an den ersten
// abbruchStellen Aufrufen ist es der Fehler: sie liegen vor jeder Abfrage.
func TestListByMatchKeyAbbruch(t *testing.T) {
	d, _ := zweiAblagen(t)
	lmSession(t, d, lmKey, "sha256:fremd", lmKey)
	probe := neueStufe(0, nil)
	voll := lmLies(probe, t, d)
	if voll != (lmStand{ids: "e1,e3", hoechste: 3}) || probe.aufrufe() < abbruchStellen {
		t.Fatalf("Probelauf %+v mit %d Aufrufen", voll, probe.aufrufe())
	}
	for k := 1; k <= probe.aufrufe(); k++ {
		ctx := neueStufe(k, nil)
		ctx.abbruch = true
		l, n, err := d.Recording().ListByMatchKey(ctx, "s", lmKey)
		switch {
		case err != nil && (!errors.Is(err, context.Canceled) || len(l) != 0 || n != 0):
			t.Errorf("Abbruch bei %d: %v %d %v", k, l, n, err)
		case err == nil && k <= abbruchStellen:
			t.Errorf("Abbruch bei %d vor der letzten Abfrage: gelungen mit %d Treffern und %d", k, len(l), n)
		case err == nil && (len(l) != 2 || n != 3):
			t.Errorf("Abbruch bei %d: verkürztes Ergebnis %d Treffer, %d", k, len(l), n)
		}
	}
}

// TestListByMatchKeyDekodiertAlleinDieTreffer hält SPEC-105 Nr. 14: ein unlesbarer Exchange mit anderem Key derselben
// Session stört nicht; ListBySession scheitert an derselben Zeile.
func TestListByMatchKeyDekodiertAlleinDieTreffer(t *testing.T) {
	d, _ := zweiAblagen(t)
	lmSession(t, d, lmKey, "sha256:fremd", lmKey)
	if _, err := d.SQL().Exec(`UPDATE exchanges SET request_headers_json = '[1]', started_at = 'kaputt' WHERE id = 'e2'`); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Recording().ListBySession(context.Background(), "s"); err == nil {
		t.Fatal("die Zeile ist lesbar: ListBySession gelang")
	}
	if got := lmLies(context.Background(), t, d); got != (lmStand{ids: "e1,e3", hoechste: 3}) {
		t.Errorf("Stand %+v", got)
	}
}

// TestListByMatchKeyGespeicherterLeererKey hält SPEC-105 Nr. 14: auch ein leer gespeicherter match_key, den das Schema
// zulässt, trifft den leeren Key nicht.
func TestListByMatchKeyGespeicherterLeererKey(t *testing.T) {
	d, _ := zweiAblagen(t)
	lmSession(t, d, lmKey, lmKey)
	if _, err := d.SQL().Exec(`UPDATE exchanges SET match_key = '' WHERE id = 'e2'`); err != nil {
		t.Fatal(err)
	}
	l, n, err := d.Recording().ListByMatchKey(context.Background(), "s", "")
	if err != nil || len(l) != 0 || n != 2 {
		t.Errorf("leerer Key: %d Treffer, %d, %v", len(l), n, err)
	}
}

// TestListByMatchKeyOrdnetNachDerNummer hält SPEC-105 Nr. 14: in der Datei weicht die Folge der Zeilen von der Folge der
// Nummern ab; die Abfrage der Treffer liefert sie dort in anderer Folge, die Operation nach sequence_no.
func TestListByMatchKeyOrdnetNachDerNummer(t *testing.T) {
	d, _ := zweiAblagen(t)
	lmSession(t, d, lmKey, "sha256:fremd", lmKey, lmKey)
	if _, err := d.SQL().Exec(`UPDATE exchanges SET rowid = 1000 - sequence_no`); err != nil {
		t.Fatal(err)
	}
	rows, err := d.SQL().Query(sqlite.AbfragenListByMatchKey()[1], "s", lmKey)
	if err != nil {
		t.Fatal(err)
	}
	spalten, err := rows.Columns()
	if err != nil || !rows.Next() {
		t.Fatalf("keine Zeile: %v", errors.Join(err, rows.Err()))
	}
	var erste string
	ziele := make([]any, len(spalten))
	ziele[0] = &erste
	for i := 1; i < len(ziele); i++ {
		ziele[i] = new(any)
	}
	if err := errors.Join(rows.Scan(ziele...), rows.Close()); err != nil || erste != "e4" {
		t.Fatalf("die Abfrage liefert zuerst %q (%v): die Folge der Datei weicht nicht ab", erste, err)
	}
	if got := lmLies(context.Background(), t, d); got != (lmStand{ids: "e1,e3,e4", hoechste: 4}) {
		t.Errorf("Stand %+v, erwartet e1,e3,e4 und 4", got)
	}
}

// TestListByMatchKeyNutztDenIndex hält SPEC-105 Nr. 14 und SPEC-004: der Plan jeder Abfrage der Operation steht als Literal.
// Die Treffer sucht idx_exchange_match nach Session und Key, die größte Nummer ein Index nach der Session, und keine legt eine
// temporäre Struktur an (kein USE TEMP, kein AUTOMATIC). Ob eine Abfrage den Index nur an einem Ende liest, sagt der Plan
// nicht; das hält TestListByMatchKeyDurchlaeuftNichtDieSession.
func TestListByMatchKeyNutztDenIndex(t *testing.T) {
	d, _ := zweiAblagen(t)
	lmSession(t, d, lmKey)
	abfragen := sqlite.AbfragenListByMatchKey()
	args := [][]any{{"s"}, {"s", lmKey}}
	soll := []string{
		"SEARCH s USING COVERING INDEX sqlite_autoindex_sessions_2 (id=?)\nCORRELATED SCALAR SUBQUERY 1\n" +
			"SEARCH exchanges USING COVERING INDEX sqlite_autoindex_exchanges_1 (session_id=?)",
		"SEARCH exchanges USING INDEX idx_exchange_match (session_id=? AND match_key=?)",
	}
	if len(abfragen) != 2 {
		t.Fatalf("%d Abfragen, erwartet 2", len(abfragen))
	}
	for i, q := range abfragen {
		rows, err := d.SQL().Query("EXPLAIN QUERY PLAN "+q, args[i]...)
		if err != nil {
			t.Fatal(err)
		}
		var plan []string
		for rows.Next() {
			var id, eltern, unbenutzt int
			var text string
			if err := rows.Scan(&id, &eltern, &unbenutzt, &text); err != nil {
				t.Fatal(err)
			}
			plan = append(plan, text)
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			t.Fatal(err)
		}
		ganz := strings.Join(plan, "\n")
		if strings.Contains(ganz, "USE TEMP") || strings.Contains(ganz, "AUTOMATIC") {
			t.Errorf("Abfrage %d legt eine temporäre Struktur an:\n%s", i, ganz)
		}
		if ganz != soll[i] {
			t.Errorf("Plan der Abfrage %d:\n%s\nerwartet:\n%s", i, ganz, soll[i])
		}
	}
}

// fuelle legt in der Session s die Exchanges e<von> bis e<bis> als Kopien des Exchange vorlage an, je mit seiner Nummer.
func fuelle(t *testing.T, d *sqlite.DB, vorlage string, von, bis int) {
	t.Helper()
	const spalten = `session_id, method, path, raw_query, match_key, started_at, finished_at, key_redacted, request_headers_json,
		request_body_state, request_blob, response_status, response_headers_json, response_body_state, response_blob,
		capture_status, request_redaction_status, response_redaction_status, body_integrity, replay_equivalence`
	if _, err := d.SQL().Exec(`WITH RECURSIVE n(i) AS (SELECT ? UNION ALL SELECT i + 1 FROM n WHERE i < ?)
		INSERT INTO exchanges(id, sequence_no, `+spalten+`) SELECT 'e' || i, i, `+spalten+` FROM n, exchanges WHERE id = ?`,
		von, bis, vorlage); err != nil {
		t.Fatal(err)
	}
}

// seitenDatei ist eine geschlossene Datei, in der die drei B-Bäume von exchanges (Tabelle, Index über Session und Nummer,
// Index über Session und Key) mehr als eine Seite haben: e1 bis e3000 mit dem Key sha256:aaa, danach e3001 und e3002 mit
// sha256:zzz. Die Treffer von sha256:zzz und die größte Nummer liegen damit in jedem Baum am Ende, die Session beginnt am
// Anfang. Die Abbildung nennt je Baum die Seite seiner Wurzel.
func seitenDatei(t *testing.T) (string, map[string]int) {
	t.Helper()
	pfad := filepath.Join(t.TempDir(), "seiten.db")
	d := open(t, pfad)
	if _, err := d.SQL().Exec(`INSERT INTO services(id, service_root, odata_version) VALUES ('d', '/', '4.0')`); err != nil {
		t.Fatal(err)
	}
	lmSession(t, d, "sha256:aaa")
	fuelle(t, d, "e1", 2, 3000)
	for _, n := range []int{3001, 3002} {
		if err := d.Recording().Append(context.Background(), lmExchange(fmt.Sprintf("e%d", n), n, "sha256:zzz")); err != nil {
			t.Fatal(err)
		}
	}
	wurzeln := map[string]int{}
	for _, name := range []string{"exchanges", "sqlite_autoindex_exchanges_1", "idx_exchange_match"} {
		var seite int
		if err := d.SQL().QueryRow("SELECT rootpage FROM sqlite_master WHERE name = ?", name).Scan(&seite); err != nil {
			t.Fatal(err)
		}
		wurzeln[name] = seite
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(pfad + "-wal"); err == nil && fi.Size() != 0 {
		t.Fatalf("WAL mit %d Bytes nach dem Schließen", fi.Size())
	}
	return pfad, wurzeln
}

// nulle überschreibt eine Seite des B-Baums mit der Wurzel wurzel mit Nullen: die Wurzel selbst oder, mit blatt, das erste
// Blatt (Dateiformat von SQLite: Seitengröße ab Byte 16, innere Seiten 0x02 und 0x05, Blätter 0x0a und 0x0d).
func nulle(t *testing.T, pfad string, wurzel int, blatt bool) {
	t.Helper()
	datei, err := os.ReadFile(pfad)
	if err != nil {
		t.Fatal(err)
	}
	groesse := int(binary.BigEndian.Uint16(datei[16:]))
	seite := wurzel
	for blatt {
		kopf := (seite - 1) * groesse
		if typ := datei[kopf]; typ == 0x0a || typ == 0x0d {
			break
		} else if typ != 0x02 && typ != 0x05 {
			t.Fatalf("Seite %d: Art %#x", seite, typ)
		}
		zelle := int(binary.BigEndian.Uint16(datei[kopf+12:]))
		seite = int(binary.BigEndian.Uint32(datei[kopf+zelle:]))
	}
	if blatt && seite == wurzel {
		t.Fatalf("der Baum mit der Wurzel %d hat nur eine Seite", wurzel)
	}
	clear(datei[(seite-1)*groesse : seite*groesse])
	if err := os.WriteFile(pfad, datei, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestListByMatchKeyDurchlaeuftNichtDieSession hält SPEC-135 Nr. 8 (der Aufwand wächst nicht mit der Session): in jedem der
// drei B-Bäume von exchanges ist das erste Blatt genullt, ein Durchlauf der Session scheitert an jedem. Die Operation gelingt:
// sie liest keinen der Bäume von seinem Anfang. Nicht geführt: eine Abfrage, die einen Teil der Session liest, in dem das erste
// Blatt nicht liegt.
func TestListByMatchKeyDurchlaeuftNichtDieSession(t *testing.T) {
	pfad, wurzeln := seitenDatei(t)
	for _, wurzel := range wurzeln {
		nulle(t, pfad, wurzel, true)
	}
	d := open(t, pfad)
	durchlaeufe := []string{
		"SELECT count(raw_query) FROM exchanges NOT INDEXED",
		"SELECT count(*) FROM exchanges INDEXED BY sqlite_autoindex_exchanges_1 WHERE session_id = 's'",
		"SELECT count(*) FROM exchanges INDEXED BY idx_exchange_match WHERE session_id = 's'",
	}
	if len(wurzeln) != 3 || len(durchlaeufe) != 3 {
		t.Fatalf("%d Bäume, %d Durchläufe", len(wurzeln), len(durchlaeufe))
	}
	for _, q := range durchlaeufe {
		var n int
		if err := d.SQL().QueryRow(q).Scan(&n); err == nil {
			t.Fatalf("%q gelang mit %d: er liest die genullte Seite nicht", q, n)
		}
	}
	l, n, err := d.Recording().ListByMatchKey(context.Background(), "s", "sha256:zzz")
	if err != nil || len(l) != 2 || l[0].ID != "e3001" || l[1].ID != "e3002" || n != 3002 {
		t.Errorf("ListByMatchKey: %d Treffer, größte %d, %v", len(l), n, err)
	}
}

// TestListByMatchKeyBeschaedigteDatei hält, dass beide Abfragen eine beschädigte Datei als ErrBeschaedigt melden (SPEC-103
// Nr. 6): die Wurzel je eines Index ist genullt, des Index der größten Nummer und des Index der Treffer.
func TestListByMatchKeyBeschaedigteDatei(t *testing.T) {
	for _, index := range []string{"sqlite_autoindex_exchanges_1", "idx_exchange_match"} {
		t.Run(index, func(t *testing.T) {
			pfad, wurzeln := seitenDatei(t)
			nulle(t, pfad, wurzeln[index], false)
			l, n, err := open(t, pfad).Recording().ListByMatchKey(context.Background(), "s", "sha256:zzz")
			if !errors.Is(err, outbound.ErrBeschaedigt) || len(l) != 0 || n != 0 {
				t.Errorf("ListByMatchKey: %d Treffer, %d, %v; erwartet ErrBeschaedigt", len(l), n, err)
			}
		})
	}
}

// TestListByMatchKeySpeicher hält SPEC-105 Nr. 14: bei drei Treffern allokiert die Operation für eine Session mit 1000
// Exchanges nicht mehr als für eine mit 10, bis auf die Abweichung 50. Gezählt werden die Allokationen des Go-Heaps, nicht
// der Speicher von SQLite.
func TestListByMatchKeySpeicher(t *testing.T) {
	messe := func(groesse int) float64 {
		d, _ := zweiAblagen(t)
		lmSession(t, d, lmKey, lmKey, lmKey, "sha256:fremd")
		fuelle(t, d, "e4", 5, groesse)
		if got := lmLies(context.Background(), t, d); !reflect.DeepEqual(got, lmStand{ids: "e1,e2,e3", hoechste: groesse}) {
			t.Fatalf("Session mit %d: %+v", groesse, got)
		}
		return testing.AllocsPerRun(20, func() {
			if _, _, err := d.Recording().ListByMatchKey(context.Background(), "s", lmKey); err != nil {
				t.Fatal(err)
			}
		})
	}
	klein, gross := messe(10), messe(1000)
	if gross-klein > 50 || klein-gross > 50 {
		t.Errorf("Allokationen je Aufruf: %.0f bei 10 Exchanges, %.0f bei 1000", klein, gross)
	}
}

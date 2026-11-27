package sqlite_test

import (
	"context"
	"errors"
	"fmt"
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

// TestListByMatchKeyNutztDenIndex hält SPEC-105 Nr. 14 und SPEC-004: die Abfrage der Treffer sucht über idx_exchange_match
// nach Session und Key, die der größten Nummer über einen Index nach der Session; keine durchläuft die Tabelle exchanges. Ein
// Index, der ohne die Bedingung durchlaufen wird, heißt im Plan auch SEARCH; deshalb steht je Abfrage die Bedingung als Literal.
func TestListByMatchKeyNutztDenIndex(t *testing.T) {
	d, _ := zweiAblagen(t)
	lmSession(t, d, lmKey)
	abfragen := sqlite.AbfragenListByMatchKey()
	args := [][]any{{"s"}, {"s", lmKey}}
	soll := []string{"(session_id=?)", "idx_exchange_match (session_id=? AND match_key=?)"}
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
		if strings.Contains(ganz, "SCAN exchanges") || !strings.Contains(ganz, "SEARCH exchanges USING") || !strings.Contains(ganz, soll[i]) {
			t.Errorf("Abfrage %d sucht nicht über %q:\n%s", i, soll[i], ganz)
		}
	}
}

// TestListByMatchKeySpeicher hält SPEC-105 Nr. 14: bei drei Treffern allokiert die Operation für eine Session mit 1000
// Exchanges nicht mehr als für eine mit 10, bis auf die Abweichung 50. Gezählt werden die Allokationen des Go-Heaps, nicht
// der Speicher von SQLite.
func TestListByMatchKeySpeicher(t *testing.T) {
	messe := func(groesse int) float64 {
		d, _ := zweiAblagen(t)
		lmSession(t, d, lmKey, lmKey, lmKey, "sha256:fremd")
		spalten := `session_id, method, path, raw_query, match_key, started_at, finished_at, key_redacted, request_headers_json,
			request_body_state, request_blob, response_status, response_headers_json, response_body_state, response_blob,
			capture_status, request_redaction_status, response_redaction_status, body_integrity, replay_equivalence`
		if _, err := d.SQL().Exec(`WITH RECURSIVE n(i) AS (SELECT 5 UNION ALL SELECT i + 1 FROM n WHERE i < ?)
			INSERT INTO exchanges(id, sequence_no, `+spalten+`) SELECT 'e' || i, i, `+spalten+` FROM n, exchanges WHERE id = 'e4'`,
			groesse); err != nil {
			t.Fatal(err)
		}
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

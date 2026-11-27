package integration_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// m6Handle ist ein hochgefahrener Prozess mit beiden Listenern: der Client des OData-Listeners und der des Admin-Listeners mit
// dem Token aus der Token-Datei (SPEC-300 Nr. 14).
type m6Handle struct {
	t     *testing.T
	h     *bootstrap.Handle
	odata *client
	admin *client
	// tokenDatei sind die Bytes der Token-Datei nach dem Hochfahren, token ist ihr Inhalt ohne Zeilenende.
	tokenDatei []byte
	token      string
}

// m6Start fährt einen Handle im Modus modus auf dir hoch; beide Listener binden 127.0.0.1:0.
func m6Start(t *testing.T, modus, dir, upstream string) *m6Handle {
	t.Helper()
	k := konfig(t, modus, dir, upstream)
	k.AdminListen = "127.0.0.1:0"
	h := hochgefahren(context.Background(), t, k)
	datei, err := os.ReadFile(k.AdminTokenFile)
	if err != nil {
		t.Fatalf("%s: Token-Datei: %v", modus, err)
	}
	return &m6Handle{t: t, h: h, odata: neuerClient(t, h.Adresse()), admin: neuerClient(t, h.AdminAdresse()),
		tokenDatei: datei, token: strings.TrimSpace(string(datei))}
}

func (m *m6Handle) beende() {
	m.t.Helper()
	if err := m.h.Herunterfahren(context.Background()); err != nil {
		m.t.Fatalf("Herunterfahren: %v", err)
	}
}

// m6Sende stellt eine Anfrage an den Admin-Listener mit genau dem Header Authorization auth (leer: keiner).
func (m *m6Handle) m6Sende(methode, pfad, auth, body string) antwort {
	m.t.Helper()
	req, err := http.NewRequest(methode, m.admin.basis+pfad, strings.NewReader(body))
	if err != nil {
		m.t.Fatal(err)
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := m.admin.c.Do(req)
	if err != nil {
		m.t.Fatalf("%s %s: %v", methode, pfad, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		m.t.Fatalf("%s %s: Body: %v", methode, pfad, err)
	}
	return antwort{status: resp.StatusCode, header: resp.Header, body: b}
}

// rufe ruft den Admin-Listener mit dem Token und verlangt den Status.
func (m *m6Handle) rufe(status int, methode, pfad, body string) antwort {
	m.t.Helper()
	a := m.m6Sende(methode, pfad, "Bearer "+m.token, body)
	if a.status != status {
		m.t.Fatalf("%s %s: %d %s, erwartet %d", methode, pfad, a.status, a.body, status)
	}
	return a
}

// m6Feld liest die Zeichenkette des Mitglieds name aus dem JSON-Objekt der Antwort.
func m6Feld(t *testing.T, a antwort, name string) string {
	t.Helper()
	s, ok := jsonVon(t, a.body)[name].(string)
	if !ok || s == "" {
		t.Fatalf("Antwort ohne %s: %s", name, a.body)
	}
	return s
}

// m6AuditZeile ist eine Zeile von audit.jsonl.
type m6AuditZeile = map[string]string

// m6Audit liest und parst DATA_DIR/audit.jsonl; eine fehlende oder leere Datei hat keine Zeile.
func m6Audit(t *testing.T, dir string) (roh []byte, zeilen []m6AuditZeile) {
	t.Helper()
	roh, err := os.ReadFile(filepath.Join(dir, "audit.jsonl"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(roh) == 0 {
		return nil, nil
	}
	for _, z := range strings.Split(strings.TrimSuffix(string(roh), "\n"), "\n") {
		m := m6AuditZeile{}
		if err := json.Unmarshal([]byte(z), &m); err != nil {
			t.Fatalf("Zeile %q: %v", z, err)
		}
		zeilen = append(zeilen, m)
	}
	return roh, zeilen
}

// m6Felder ist die Zeile ohne die zwei Felder, die der Test nicht als Literal kennt, in sortierter Form.
func m6Felder(z m6AuditZeile) string {
	var f []string
	for k, v := range z {
		if k != "timestamp" && k != "correlation_id" {
			f = append(f, k+"="+v)
		}
	}
	slices.Sort(f)
	return strings.Join(f, " ")
}

// m6Ablage ist der Zustand der Ablage: jede Zeile jeder Tabelle und jede Datei des Blob-Verzeichnisses mit ihrem Inhalt.
type m6Ablage struct {
	tabellen map[string][]string
	blobs    map[string]string
}

// m6Zustand liest die Ablage unter dir nur lesend, neben dem laufenden Handle.
func m6Zustand(t *testing.T, dir string) m6Ablage {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(dir, "odata-recorder.db")+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	z := m6Ablage{tabellen: map[string][]string{}, blobs: map[string]string{}}
	for _, name := range m6Spalte(t, db, `SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`) {
		z.tabellen[name] = m6Zeilen(t, db, `SELECT * FROM "`+name+`"`)
	}
	err = filepath.WalkDir(filepath.Join(dir, "blobs"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		z.blobs[p] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return z
}

// m6Zeilen liefert die Zeilen der Abfrage als Text, sortiert.
func m6Zeilen(t *testing.T, db *sql.DB, abfrage string) []string {
	t.Helper()
	rows, err := db.Query(abfrage)
	if err != nil {
		t.Fatalf("%s: %v", abfrage, err)
	}
	defer func() { _ = rows.Close() }()
	spalten, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	var aus []string
	for rows.Next() {
		werte := make([]any, len(spalten))
		ziele := make([]any, len(spalten))
		for i := range werte {
			ziele[i] = &werte[i]
		}
		if err := rows.Scan(ziele...); err != nil {
			t.Fatal(err)
		}
		aus = append(aus, fmt.Sprintf("%q", fmt.Sprint(werte...)))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	slices.Sort(aus)
	return aus
}

// m6Spalte liefert die erste Spalte der Abfrage.
func m6Spalte(t *testing.T, db *sql.DB, abfrage string) []string {
	t.Helper()
	rows, err := db.Query(abfrage)
	if err != nil {
		t.Fatalf("%s: %v", abfrage, err)
	}
	defer func() { _ = rows.Close() }()
	var aus []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		aus = append(aus, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return aus
}

// unterschiede nennt die Tabellen und Blob-Dateien, in denen sich die zwei Zustände unterscheiden.
func (a m6Ablage) unterschiede(b m6Ablage) []string {
	var aus []string
	namen := map[string]bool{}
	for n := range a.tabellen {
		namen[n] = true
	}
	for n := range b.tabellen {
		namen[n] = true
	}
	for n := range namen {
		if !slices.Equal(a.tabellen[n], b.tabellen[n]) {
			aus = append(aus, "Tabelle "+n)
		}
	}
	for p, v := range a.blobs {
		if w, da := b.blobs[p]; !da || w != v {
			aus = append(aus, "Blob "+filepath.Base(p))
		}
	}
	for p := range b.blobs {
		if _, da := a.blobs[p]; !da {
			aus = append(aus, "Blob "+filepath.Base(p))
		}
	}
	slices.Sort(aus)
	return aus
}

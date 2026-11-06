package audit_test

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/audit"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

const logMsg = "audit: Eintrag nicht geschrieben"

// scheiternde Ablagen: je Art ein Aufbau, in dem Append an einer anderen Stelle scheitert.
func scheiternde(t *testing.T) map[string]*aufbau {
	t.Helper()
	jahr0 := func(int) time.Time { return time.Date(0, 12, 31, 0, 0, 0, 0, time.UTC) }
	return map[string]*aufbau{
		"Öffnen":    neuAn(t, filepath.Join(t.TempDir(), "CANARY-pfad", audit.DateiName), festeZeit),
		"Schreiben": neuAn(t, "/dev/full", festeZeit),
		"Zeit":      neu(t, jahr0),
	}
}

// A16: je Stelle des Scheiterns × Art genau ein Log-Eintrag mit festem Text und den Schlüsseln der
// Spezifikation (Soll-Objekt exakt), ohne Fehlertext, Pfad und die Felder import_id, error_code
// und KorrelationsID des Eintrags.
func TestLogBeiFehler(t *testing.T) {
	arten := []support.AuditArt{support.ArtSessionGeloescht, support.ArtSessionAufbewahrung, support.ArtWorkspaceGeloescht, support.ArtImportFehlgeschlagen}
	paare := 0
	for stelle := range scheiternde(t) {
		for _, art := range arten {
			a := scheiternde(t)[stelle]
			e := gueltig(art)
			e.KorrelationsID = "korr-canary"
			err := a.d.Append(e)
			if err == nil {
				t.Errorf("%s/%s: Append gelang", stelle, art)
				continue
			}
			paare++
			soll := map[string]any{
				"level": "ERROR", "msg": logMsg, "ereignis": "audit_fehlgeschlagen",
				"error_code": "ODR-E9001", "grund": string(art),
			}
			for _, kv := range [][2]string{{"service_id", e.ServiceID}, {"session_id", e.SessionID}, {"workspace_id", e.WorkspaceID}} {
				if kv[1] != "" {
					soll[kv[0]] = kv[1]
				}
			}
			zeilen := a.logZeilen(t)
			if len(zeilen) != 1 || !reflect.DeepEqual(zeilen[0], soll) {
				t.Errorf("%s/%s: Log %v, erwartet genau %v", stelle, art, zeilen, soll)
			}
			log := a.log.String()
			for _, verboten := range []string{"CANARY", "korr-canary", err.Error(), "no such file", "no space", "w-import_id"} {
				if strings.Contains(log, verboten) {
					t.Errorf("%s/%s: das Log nennt %q: %s", stelle, art, verboten, log)
				}
			}
		}
	}
	if paare != 3*4 {
		t.Errorf("%d Paare, erwartet 12", paare)
	}
}

// A16: ein gelungenes Append schreibt nichts in das Log.
func TestLogBeiErfolgLeer(t *testing.T) {
	a := neu(t, festeZeit)
	if err := a.d.Append(gueltig(support.ArtSessionGeloescht)); err != nil {
		t.Fatal(err)
	}
	if a.log.Len() != 0 {
		t.Errorf("Log nach Erfolg: %q", a.log.String())
	}
}

// A17: ein Eintrag, der die Prüfung nicht besteht, trägt im Log nur ereignis und error_code —
// kein grund, keine Kennung, keinen Wert, auch wenn die Art gültig ist.
func TestLogBeiUngueltig(t *testing.T) {
	mit := func(f func(*support.Eintrag)) support.Eintrag {
		e := gueltig(support.ArtSessionGeloescht)
		e.ServiceID, e.SessionID = "canary-svc", "canary-ses"
		f(&e)
		return e
	}
	faelle := map[string]support.Eintrag{
		"Art unbekannt":     mit(func(e *support.Eintrag) { e.Art = "canary-art" }),
		"Pflichtfeld fehlt": mit(func(e *support.Eintrag) { e.SessionID = "" }),
		"Form":              mit(func(e *support.Eintrag) { e.ServiceID = "canary svc" }),
		"verbotenes Feld":   mit(func(e *support.Eintrag) { e.WorkspaceID = "canary-wsp" }),
		"verbotener Code":   mit(func(e *support.Eintrag) { e.ErrorCode = model.CodeE6101 }),
		"Korrelations-ID":   mit(func(e *support.Eintrag) { e.KorrelationsID = "canary korr" }),
	}
	for name, e := range faelle {
		a := neu(t, festeZeit)
		if err := a.d.Append(e); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v, erwartet ErrUngueltig", name, err)
		}
		soll := map[string]any{"level": "ERROR", "msg": logMsg, "ereignis": "audit_fehlgeschlagen", "error_code": "ODR-E9001"}
		if zeilen := a.logZeilen(t); len(zeilen) != 1 || !reflect.DeepEqual(zeilen[0], soll) {
			t.Errorf("%s: Log %v, erwartet genau %v", name, zeilen, soll)
		}
		if strings.Contains(a.log.String(), "canary") {
			t.Errorf("%s: das Log nennt einen Wert des Eintrags: %s", name, a.log.String())
		}
	}
	if len(faelle) != 6 {
		t.Fatalf("%d Fälle, erwartet 6", len(faelle))
	}
}

// blockUhr blockiert im ersten Aufruf von Now bis zur Freigabe und meldet jeden zweiten Aufruf.
type blockUhr struct {
	mu          sync.Mutex
	n           int
	ersterDrin  chan struct{}
	freigabe    chan struct{}
	zweiterDrin chan struct{}
}

func (u *blockUhr) Now() time.Time {
	u.mu.Lock()
	u.n++
	k := u.n
	u.mu.Unlock()
	switch k {
	case 1:
		close(u.ersterDrin)
		<-u.freigabe
	case 2:
		close(u.zweiterDrin)
	}
	return stand()
}

func (u *blockUhr) Aufrufe() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.n
}

// A18: solange der erste Append in Uhr.Now() blockiert, ruft der zweite die Uhr nicht; die
// Sperre beginnt vor Uhr.Now(). Beide Zeilen stehen danach in der Datei.
func TestSperreUmfasstUhr(t *testing.T) {
	a := neu(t, festeZeit)
	u := &blockUhr{ersterDrin: make(chan struct{}), freigabe: make(chan struct{}), zweiterDrin: make(chan struct{})}
	d, err := audit.New(audit.Config{Pfad: a.pfad, Uhr: u, IDs: a.ids, Log: slogNull()})
	if err != nil {
		t.Fatal(err)
	}
	fertig := make(chan error, 2)
	go func() { fertig <- d.Append(gueltig(support.ArtSessionGeloescht)) }()
	select {
	case <-u.ersterDrin:
	case <-time.After(frist):
		t.Fatal("der erste Append erreicht Uhr.Now() nicht")
	}
	go func() { fertig <- d.Append(gueltig(support.ArtWorkspaceGeloescht)) }()
	select {
	case <-u.zweiterDrin:
		t.Error("der zweite Append rief Uhr.Now(), während der erste sie hielt")
	case <-time.After(300 * time.Millisecond):
	}
	if n := u.Aufrufe(); n != 1 && !t.Failed() {
		t.Errorf("Uhr %d-mal gerufen, erwartet 1", n)
	}
	close(u.freigabe)
	for range 2 {
		select {
		case err := <-fertig:
			if err != nil {
				t.Errorf("Append: %v", err)
			}
		case <-time.After(frist):
			t.Fatal("ein Append kehrt nach der Freigabe nicht zurück")
		}
	}
	if n, zeilen := u.Aufrufe(), strings.Count(a.inhalt(t), "\n"); n != 2 || zeilen != 2 {
		t.Errorf("Uhr %d-mal gerufen, %d Zeilen, erwartet 2 und 2", n, zeilen)
	}
}

// A19: 64 gleichzeitige Append mit einer Uhr, die zwischen dem Ablesen und der Rückgabe die
// Prozessorzeit abgibt: 64 vollständige Zeilen, jede gültiges JSON, Zeitpunkte nicht abnehmend,
// 64 verschiedene IDs.
func TestGleichzeitigeAppend(t *testing.T) {
	const n = 64
	var mu sync.Mutex
	zaehler := 0
	uhr := func(int) time.Time {
		mu.Lock()
		zaehler++
		z := stand().Add(time.Duration(zaehler) * time.Nanosecond)
		mu.Unlock()
		runtime.Gosched()
		return z
	}
	a := neu(t, uhr)
	arten := []support.AuditArt{support.ArtSessionGeloescht, support.ArtSessionAufbewahrung, support.ArtWorkspaceGeloescht, support.ArtImportFehlgeschlagen}
	start := make(chan struct{})
	var wg sync.WaitGroup
	fehler := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			fehler <- a.d.Append(gueltig(arten[i%4]))
		}()
	}
	close(start)
	wg.Wait()
	close(fehler)
	for err := range fehler {
		if err != nil {
			t.Errorf("Append: %v", err)
		}
	}
	zeilen := strings.Split(strings.TrimSuffix(a.inhalt(t), "\n"), "\n")
	if len(zeilen) != n {
		t.Fatalf("%d Zeilen, erwartet %d", len(zeilen), n)
	}
	ids, vorher := map[string]bool{}, ""
	for i, z := range zeilen {
		var m map[string]string
		if err := json.Unmarshal([]byte(z), &m); err != nil {
			t.Fatalf("Zeile %d %q ist kein JSON: %v", i, z, err)
		}
		if m["timestamp"] < vorher {
			t.Errorf("Zeile %d: Zeitpunkt %s nimmt ab gegen %s", i, m["timestamp"], vorher)
		}
		vorher = m["timestamp"]
		ids[m["correlation_id"]] = true
	}
	if len(ids) != n {
		t.Errorf("%d verschiedene IDs, erwartet %d", len(ids), n)
	}
}

// A20: nach jedem Rückkehrpfad mit Fehler ist die Sperre frei: ein zweites Append kehrt innerhalb
// der Frist zurück und gelingt, wo die Ursache behoben ist. Pfade: Form (vor der Sperre), Zeit,
// Öffnen, Schreiben (hier nur die Rückkehr: /dev/full lässt kein Schreiben gelingen).
func TestSperreFreiNachFehler(t *testing.T) {
	gut := func(a *aufbau) error {
		return mitFrist(t, "zweites Append", func() error { return a.d.Append(gueltig(support.ArtSessionGeloescht)) })
	}
	schlecht := gueltig(support.ArtSessionGeloescht)
	schlecht.Art = "unbekannt"

	form := neu(t, festeZeit)
	if err := form.d.Append(schlecht); !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Form: %v", err)
	}
	if err := gut(form); err != nil {
		t.Errorf("nach der Form: %v", err)
	}

	zeit := neu(t, func(n int) time.Time {
		if n == 1 {
			return time.Date(0, 12, 31, 0, 0, 0, 0, time.UTC)
		}
		return stand()
	})
	if err := zeit.d.Append(gueltig(support.ArtSessionGeloescht)); !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Zeit: %v", err)
	}
	if err := gut(zeit); err != nil {
		t.Errorf("nach der Zeit: %v", err)
	}

	wurzel := t.TempDir()
	oeffnen := neuAn(t, filepath.Join(wurzel, "fehlt", audit.DateiName), festeZeit)
	if err := oeffnen.d.Append(gueltig(support.ArtSessionGeloescht)); err == nil {
		t.Error("Öffnen: Append gelang")
	}
	if err := mkdirFuer(wurzel); err != nil {
		t.Fatal(err)
	}
	if err := gut(oeffnen); err != nil {
		t.Errorf("nach dem Öffnen: %v", err)
	}

	schreiben := neuAn(t, "/dev/full", festeZeit)
	for i := range 2 {
		if err := gut(schreiben); err == nil {
			t.Errorf("Schreiben %d: Append gelang auf /dev/full", i)
		}
	}
}

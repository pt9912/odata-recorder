package e2e_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/tests/e2e/fixture"
)

// Der Canary-Test am gebauten Binary (SPEC-041 Nr. 4 bis 6, LH-QA-04.a): je Stelle der Default-Listen und je ein konfigurierter Eintrag geht ein Canary-Wert als
// Anfrage und Antwort durch das Binary im Modus record; danach sucht der Test in jedem Artefakt des Prozesses (sicherheit_suche_test.go).

// Die Spurwerte: je ein Name, den keine Liste nennt (Header, Query, JSON-Mitglied `titel` der Anfrage); die Redaction lässt ihn stehen, ihr Fund belegt, dass die
// Aufnahme beobachtet ist und die Suche liest, was die Redaction stehen lässt.
const (
	sicherheitSpurHeader = "SPUR-hdr-3c91"
	sicherheitSpurQuery  = "SPUR-qry-5e07"
	sicherheitSpurJSON   = "SPUR-json-8a42"
	sicherheitSync       = 10 * time.Second
)

func sicherheitSpuren() []string {
	return []string{sicherheitSpurHeader, sicherheitSpurQuery, sicherheitSpurJSON}
}

// sicherheitMarkerDateien sind die Kontrolldateien unter HOME und unter dem Arbeitsverzeichnis des Prozesses (SPEC-041 Nr. 6 d): je Ort eine Datei mit einem
// eigenen Marker, relativ zur Arbeitswurzel.
func sicherheitMarkerDateien() map[string]string {
	return map[string]string{
		filepath.Join("home", "marker-home.txt"): "KONTROLLE-HOME-7c3d",
		filepath.Join("cwd", "marker-cwd.txt"):   "KONTROLLE-CWD-2f9a",
	}
}

// sicherheitPflanzeMarker legt die Marker-Dateien unter der Arbeitswurzel ab.
func sicherheitPflanzeMarker(t *testing.T, root string) {
	t.Helper()
	for rel, marker := range sicherheitMarkerDateien() {
		if err := os.WriteFile(filepath.Join(root, rel), []byte(marker), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// sicherheitMarkerWerte sind die Marker der Dateien unter HOME und dem Arbeitsverzeichnis.
func sicherheitMarkerWerte() []string {
	var w []string
	for _, m := range sicherheitMarkerDateien() {
		w = append(w, m)
	}
	return w
}

// sicherheitGlieder bildet die Mitglieder `"name":"wert"` der Stellen einer Art in einer Tiefe, die Werte als JSON-Zeichenketten.
func sicherheitGlieder(stellen []sicherheitStelle, art string, tiefe int) string {
	var teile []string
	for _, s := range stellen {
		if s.art == art && s.tiefe == tiefe {
			teile = append(teile, fmt.Sprintf("%q:%s", s.name, sicherheitJSONWert(s.wert())))
		}
	}
	return strings.Join(teile, ",")
}

// sicherheitJSONWert ist der Wert als JSON-Zeichenkette mit den Anführungszeichen.
func sicherheitJSONWert(wert string) string {
	j, _ := json.Marshal(wert)
	return string(j)
}

// sicherheitBody bildet den JSON-Body einer Seite: Mitglieder der Tiefe 0, das Objekt `n1` mit denen der Tiefe 1 und das Array `arr` mit einem Objekt der Tiefe 2;
// mit nichtstring steht vor allen ein Mitglied unter einem treffenden Namen, dessen Wert ein Objekt ist (SPEC-118 Nr. 5): jede Stelle hinter ihm ist die, die
// ein Redaktor ungeschützt lässt, der den Fehler übergeht.
func sicherheitBody(stellen []sicherheitStelle, art, kopf string, nichtstring bool) string {
	vor := ""
	if nichtstring {
		vor = `"secret":{"verschachtelt":true},`
	}
	return "{" + vor + kopf + sicherheitGlieder(stellen, art, 0) + `,"n1":{` + sicherheitGlieder(stellen, art, 1) + `},"arr":[{` + sicherheitGlieder(stellen, art, 2) + `}]}`
}

// sicherheitQuery bildet die Query einer Art: Name und Query-kodierter Wert je Stelle.
func sicherheitQuery(stellen []sicherheitStelle, art string) string {
	var teile []string
	for _, s := range stellen {
		if s.art == art {
			teile = append(teile, s.name+"="+url.QueryEscape(s.wert()))
		}
	}
	return strings.Join(teile, "&")
}

// sicherheitAnfrageVon bildet die Anfrage `POST /odata/Products?<Query>` mit allen Stellen der Anfrage unter stellen: Header, Query und Body, der Pfad trägt keinen
// Canary (SPEC-041 Nr. 5).
func sicherheitAnfrageVon(stellen []sicherheitStelle, nichtstring bool) anfrage {
	header := map[string][]string{"Content-Type": {"application/json"}, "X-Test-Marker": {sicherheitSpurHeader}}
	for _, s := range stellen {
		if s.art == "hq" {
			header[s.name] = []string{s.wert()}
		}
	}
	kopf := `"titel":"` + sicherheitSpurJSON + `",`
	return anfrage{
		methode: http.MethodPost, pfad: "/odata/Products?marker=" + sicherheitSpurQuery + "&" + sicherheitQuery(stellen, "qq"),
		header: header, body: sicherheitBody(stellen, "jq", kopf, nichtstring),
	}
}

// sicherheitAufbau prüft, dass jede Stelle genau einen Wert trägt: ein Aufbau, der eine Stelle auslässt oder verdoppelt, ist ein Fehler des Tests, kein Fund.
func sicherheitAufbau(t *testing.T, name string, stellen []sicherheitStelle, text string) {
	t.Helper()
	for _, s := range stellen {
		form := s.formen()[map[string]int{"hq": 0, "hr": 0, "qq": 1, "ql": 1, "qc": 1, "jq": 2, "jr": 2}[s.art]]
		if n := strings.Count(text, form); n != 1 {
			t.Fatalf("Aufbau %s: die Stelle %s trägt %d Werte, erwartet 1", name, s.kennung, n)
		}
	}
}

// sicherheitUpstream ist der Test-Server im Prozess des Tests: er antwortet auf `POST /odata/Products` mit 201, den Headern der Antwort, `Location` und
// `Content-Location` mit je den Query-Namen und einem Body mit den Stellen der Antwort, auf `GET /odata/$metadata` mit einem gültigen EDMX (SPEC-041 Nr. 5 a).
type sicherheitUpstream struct {
	srv   *httptest.Server
	posts atomic.Int64
}

func sicherheitNeuerUpstream(t *testing.T, stellen []sicherheitStelle) *sicherheitUpstream {
	t.Helper()
	antwort := sicherheitAnfrageSeite(stellen, false)
	u := &sicherheitUpstream{}
	var text strings.Builder
	u.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/odata/$metadata":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(fixture.CSDL))
		case r.Method == http.MethodPost && r.URL.Path == "/odata/Products":
			u.posts.Add(1)
			h := w.Header()
			h.Set("Content-Type", "application/json")
			for _, s := range antwort {
				if s.art == "hr" {
					h.Set(s.name, s.wert())
				}
			}
			h.Set("Location", u.srv.URL+"/odata/Products(1)?"+sicherheitQuery(antwort, "ql"))
			h.Set("Content-Location", u.srv.URL+"/odata/Products(1)?"+sicherheitQuery(antwort, "qc"))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(sicherheitBody(antwort, "jr", `"Id":1,`, false)))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	u.srv.Start()
	t.Cleanup(u.srv.Close)
	for _, s := range sicherheitArt(antwort, "hr") {
		text.WriteString(s.wert() + "\n")
	}
	text.WriteString(u.srv.URL + "/odata/Products(1)?" + sicherheitQuery(antwort, "ql") + "\n" + u.srv.URL + "/odata/Products(1)?" + sicherheitQuery(antwort, "qc") + "\n")
	text.WriteString(sicherheitBody(antwort, "jr", `"Id":1,`, false))
	for _, art := range []string{"hr", "ql", "qc", "jr"} {
		sicherheitAufbau(t, "Antwort "+art, sicherheitArt(antwort, art), text.String())
	}
	return u
}

// sicherheitArt filtert die Stellen einer Art.
func sicherheitArt(stellen []sicherheitStelle, art string) []sicherheitStelle {
	var aus []sicherheitStelle
	for _, s := range stellen {
		if s.art == art {
			aus = append(aus, s)
		}
	}
	return aus
}

// sicherheitFall ist ein Fall des Canary-Tests (SPEC-041 Nr. 5 b): Name, Zahl der erwarteten Suchzellen, Status der Antwort an den Client, die Teile der Zeile auf
// stderr, an der die Aufnahme beobachtet wird, und wie der Fall vom Regellauf abweicht.
type sicherheitFall struct {
	name                           string
	zellen                         int
	status                         int
	zeile                          []string
	nichtstring, upstreamWeg, voll bool
}

func sicherheitFaelle() []sicherheitFall {
	return []sicherheitFall{
		{name: "F1 Regellauf", zellen: 752, status: 201},
		{name: "F2 nicht redigierbar", zellen: 752, status: 201, nichtstring: true,
			zeile: []string{`"ereignis":"capture"`, `"richtung":"request"`, `"error_code":"ODR-E2401"`}},
		{name: "F3 Upstream nicht erreichbar", zellen: 336, status: 502, upstreamWeg: true,
			zeile: []string{`"ereignis":"upstream"`, `"grund":"nicht_erreichbar"`}},
		{name: "F4 Speicher voll", zellen: 752, status: 201, voll: true,
			zeile: []string{`"ereignis":"ablage"`, `"error_code":"ODR-`}},
	}
}

// sicherheitZaehlung zählt die Suchzellen und die Kontrollzellen über alle Fälle.
type sicherheitZaehlung struct{ suche, kontrolle int }

// sicherheitMomentbild liefert Namen und Größen aller Dateien unter den Verzeichnissen als Text; vor und während der Aufnahme verschwindende Dateien fehlen darin.
func sicherheitMomentbild(dirs []string) string {
	var sb strings.Builder
	for _, d := range dirs {
		_ = filepath.WalkDir(d, func(p string, e os.DirEntry, err error) error {
			if err != nil || e.IsDir() {
				return nil
			}
			if fi, err := e.Info(); err == nil {
				fmt.Fprintf(&sb, "%s:%d\n", p, fi.Size())
			}
			return nil
		})
	}
	return sb.String()
}

// sicherheitWarte ruft bedingung alle 25 ms bis sie wahr ist; nach sicherheitSync ist der Test rot (die Aufnahme ist nicht beobachtet), und die Suche läuft
// dennoch, damit ein Fund neben der Meldung steht.
func sicherheitWarte(t *testing.T, p *prozess, was string, bedingung func() bool) {
	t.Helper()
	ende := time.Now().Add(sicherheitSync)
	for !bedingung() {
		if p.beendet() || time.Now().After(ende) {
			t.Errorf("Aufnahme nicht beobachtet binnen %v: %s (Prozess beendet: %v)\nstderr: %s", sicherheitSync, was, p.beendet(), p.fehl)
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// sicherheitZeileDa sagt, ob stderr eine Zeile mit allen Teilen trägt.
func sicherheitZeileDa(p *prozess, teile []string) bool {
	for _, z := range strings.Split(p.fehl.String(), "\n") {
		if alle(z, teile) {
			return true
		}
	}
	return false
}

// sicherheitSpurenDa sagt, ob jeder Spurwert in einer Datei unter den Verzeichnissen steht; eine Datei, die währenddessen verschwindet, zählt nicht.
func sicherheitSpurenDa(dirs []string) bool {
	gefunden := map[string]bool{}
	for _, d := range dirs {
		_ = filepath.WalkDir(d, func(p string, e os.DirEntry, err error) error {
			if err != nil || e.IsDir() {
				return nil
			}
			if b, err := os.ReadFile(p); err == nil {
				for _, w := range inBytes(p, b, sicherheitSpuren()) {
					gefunden[w.wert] = true
				}
			}
			return nil
		})
	}
	return len(gefunden) == len(sicherheitSpuren())
}

// sicherheitWarteStill wartet, bis die Dateien unter den Verzeichnissen sich dreimal im Abstand von 50 ms nicht ändern: die Suche liest keine Datei, die der
// Prozess eben schreibt oder umbenennt.
func sicherheitWarteStill(t *testing.T, p *prozess, dirs []string) {
	t.Helper()
	davor, gleich := "", 0
	sicherheitWarte(t, p, "die Dateien werden nicht still", func() bool {
		jetzt := sicherheitMomentbild(dirs)
		if jetzt == davor {
			gleich++
		} else {
			davor, gleich = jetzt, 0
		}
		time.Sleep(25 * time.Millisecond)
		return gleich >= 3
	})
}

// sicherheitKontrollen prüft die Kontrollen der Suche (SPEC-041 Nr. 6 d) zu einem Zeitpunkt: Verstöße sind rot mit eigener Meldung, nie ein Fund. Gezählt wird
// beim Zeitpunkt vor dem Herunterfahren; die Kontrolldateien, die Marker und den leeren stdout prüft auch der zweite.
func sicherheitKontrollen(t *testing.T, f sicherheitFall, vor bool, e sicherheitErgebnis, strom map[string][]byte, z *sicherheitZaehlung) {
	t.Helper()
	zaehle := func(ok bool, was string) {
		t.Helper()
		if vor {
			z.kontrolle++
		}
		if !ok {
			t.Errorf("Kontrolle %s (%s, vor dem Herunterfahren: %v): ohne sie ist der Leerstand der Suche nichts wert", was, f.name, vor)
		}
	}
	for rel, marker := range kontrollDateien() {
		zaehle(hat(e.kontroll, rel, marker), "Kontrolldatei "+rel)
	}
	for rel, marker := range sicherheitMarkerDateien() {
		zaehle(hat(e.kontroll, rel, marker), "Marker-Datei "+rel)
	}
	if f.zeile != nil {
		zaehle(hat(e.kontroll, "stderr", kontrollDienst), "Marker auf stderr")
	}
	zaehle(len(strom["stdout"]) == 0, "leerer stdout")
	if !vor {
		return
	}
	db, wal, blob := false, false, false
	for _, d := range e.dateien {
		db = db || filepath.Base(d) == "odata-recorder.db"
		wal = wal || filepath.Base(d) == "odata-recorder.db-wal"
		blob = blob || strings.Contains(d, "blobs"+string(filepath.Separator))
	}
	zaehle(db, "Datenbank odata-recorder.db")
	zaehle(len(strom["stderr"]) > 0, "nichtleerer stderr")
	if f.zeile == nil {
		zaehle(wal, "odata-recorder.db-wal")
		zaehle(blob, "Blob-Datei")
		for _, s := range sicherheitSpuren() {
			zaehle(hat(e.kontroll, "", s), "Spurwert "+s)
		}
	}
}

// TestSicherheitCanary (TS10 bis TS14): vier Fälle mit je einem frischen Prozess im Modus record und je einer Anfrage mit allen Stellen der Anfrage; die Suche über
// die Arbeitswurzel, stdout und stderr findet keinen Canary-Wert, nach der Synchronisation und nach dem Herunterfahren, und die Kontrollen der Suche halten.
func TestSicherheitCanary(t *testing.T) {
	binaryPfad(t)
	z := &sicherheitZaehlung{}
	stellen := sicherheitStellen()
	if len(stellen) != sicherheitStellenZahl {
		t.Fatalf("%d Stellen, erwartet %d", len(stellen), sicherheitStellenZahl)
	}
	soll := 0
	for _, f := range sicherheitFaelle() {
		soll += f.zellen
		t.Run(f.name, func(t *testing.T) { sicherheitLaufe(t, f, stellen, z) })
	}
	if !t.Failed() && (z.suche != soll || soll != 2592 || z.kontrolle != 48) {
		t.Errorf("%d Suchzellen und %d Kontrollzellen gezählt; erwartet 2592 (%d) und 48", z.suche, z.kontrolle, soll)
	}
}

// sicherheitLaufe fährt einen Fall: Prozess starten, Anfrage senden, synchronisieren, suchen, herunterfahren, suchen.
func sicherheitLaufe(t *testing.T, f sicherheitFall, alle []sicherheitStelle, z *sicherheitZaehlung) {
	t.Helper()
	stellen := alle
	if f.upstreamWeg {
		stellen = sicherheitAnfrageSeite(alle, true)
	}
	root := t.TempDir()
	for _, d := range []string{"tmp", "home", "cwd", "data"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	data := filepath.Join(root, "data")
	wurzeln, dirs := []sicherheitWurzel{{"root", root}}, []string{root}
	if f.voll {
		var err error
		if data, err = os.MkdirTemp("/voll", "data"); err != nil {
			t.Fatalf("DATA_DIR unter /voll: %v", err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(data) })
		wurzeln, dirs = append(wurzeln, sicherheitWurzel{"voll", "/voll"}), append(dirs, "/voll")
	}
	up := sicherheitNeuerUpstream(t, alle)
	a := sicherheitAnfrageVon(sicherheitAnfrageSeite(alle, true), f.nichtstring)
	sicherheitAufbau(t, "Anfrage Header", sicherheitArt(stellen, "hq"), headerText(a))
	sicherheitAufbau(t, "Anfrage Query", sicherheitArt(stellen, "qq"), a.pfad)
	sicherheitAufbau(t, "Anfrage Body", sicherheitArt(stellen, "jq"), a.body)
	t.Chdir(filepath.Join(root, "cwd"))
	k := konfig{modus: "record", dir: data, upstream: up.srv.URL, dienst: kontrollDienst, extra: []string{
		"REDACT_HEADERS=" + sicherheitEigenHeader, "REDACT_QUERY_PARAMS=" + sicherheitEigenQuery, "REDACT_JSON_PATHS=" + sicherheitEigenJSON,
		"TMPDIR=" + filepath.Join(root, "tmp"), "HOME=" + filepath.Join(root, "home"),
	}}
	adresse := freierPort(t)
	p := startRoh(t, adresse, k.env(adresse))
	p.modus = "record"
	p.bereit()
	pflanzeKontrolle(t, filepath.Join(root, "tmp"))
	sicherheitPflanzeMarker(t, root)
	var ballast string
	switch {
	case f.upstreamWeg:
		up.srv.Close()
	case f.voll:
		ballast = fuelle(t)
		t.Cleanup(func() { _ = os.Remove(ballast) })
	}
	w := holen(t, adresse, a)
	if w.status != f.status {
		t.Errorf("Status der Antwort %d, erwartet %d\n%s", w.status, f.status, w.body)
	}
	if f.upstreamWeg && !strings.Contains(string(w.body), "ODR-E4001") {
		t.Errorf("Body der Antwort %q, erwartet ODR-E4001", w.body)
	}
	posts := int64(1)
	if f.upstreamWeg {
		posts = 0
	}
	if n := up.posts.Load(); n != posts {
		t.Errorf("der Test-Server sah %d POST, der Aufbau erwartet %d", n, posts)
	}
	if f.zeile != nil {
		sicherheitWarte(t, p, "Zeile auf stderr mit "+strings.Join(f.zeile, " "), func() bool { return sicherheitZeileDa(p, f.zeile) })
	} else {
		sicherheitWarte(t, p, "die drei Spurwerte stehen in Dateien", func() bool { return sicherheitSpurenDa(dirs) })
	}
	sicherheitWarteStill(t, p, dirs)
	zellen := 0
	for _, vor := range []bool{true, false} {
		if !vor {
			if ballast != "" {
				if err := os.Remove(ballast); err != nil {
					t.Fatal(err)
				}
			}
			if code := p.beende(); code != 0 {
				t.Fatalf("SIGTERM: Exit %d\nstderr: %s", code, p.fehl)
			}
		}
		strom := map[string][]byte{"stdout": p.aus.bytes(), "stderr": p.fehl.bytes()}
		e := sicherheitSuche(t, wurzeln, strom, stellen, slices.Concat(werteKontrolle(), sicherheitSpuren(), sicherheitMarkerWerte()))
		zellen += e.zellen
		for _, gefunden := range e.funde {
			t.Errorf("%s, vor dem Herunterfahren: %v: %v", f.name, vor, gefunden)
		}
		sicherheitKontrollen(t, f, vor, e, strom, z)
	}
	if zellen != f.zellen {
		t.Errorf("%s: %d Suchzellen gezählt, erwartet %d", f.name, zellen, f.zellen)
	}
	z.suche += zellen
}

// TestSicherheitUmgebung (SPEC-041 Nr. 1, *Umgebung*, und Nr. 9 f): der Container, in dem die Tests laufen, ist netzlos und hat ein schreibgeschütztes
// Wurzeldateisystem: jede Schnittstelle außer `lo` fehlt oder ist nicht aufwärts, das Anlegen einer Datei in `/` scheitert mit EROFS, und das Anlegen unter `/tmp`
// gelingt (ohne diese Gegenprobe zählte jeder andere Fehler als Schreibschutz). Der Test liest nur den Zustand des Containers und läuft im dritten Binary.
func TestSicherheitUmgebung(t *testing.T) {
	binaryPfad(t)
	schnittstellen, err := net.Interfaces()
	if err != nil {
		t.Fatalf("Schnittstellen: %v", err)
	}
	for _, s := range schnittstellen {
		if s.Name != "lo" && s.Flags&net.FlagUp != 0 {
			t.Errorf("die Schnittstelle %s ist oben: der Container ist nicht netzlos", s.Name)
		}
	}
	ziel := "/sicherheit-umgebung"
	if err := os.WriteFile(ziel, []byte("x"), 0o600); err == nil {
		_ = os.Remove(ziel)
		t.Errorf("das Anlegen einer Datei in / gelingt: das Wurzeldateisystem ist beschreibbar")
	} else if !errors.Is(err, syscall.EROFS) {
		t.Errorf("das Anlegen einer Datei in / scheitert mit %v, erwartet EROFS", err)
	}
	f, err := os.CreateTemp("/tmp", "sicherheit-umgebung")
	if err != nil {
		t.Errorf("das Anlegen einer Datei unter /tmp scheitert: %v", err)
		return
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
}

// headerText ist der Text der Header einer Anfrage, in dem die Werte der Header-Stellen einzeln gezählt werden.
func headerText(a anfrage) string {
	var sb strings.Builder
	for n, vs := range a.header {
		for _, v := range vs {
			sb.WriteString(n + ": " + v + "\n")
		}
	}
	return sb.String()
}

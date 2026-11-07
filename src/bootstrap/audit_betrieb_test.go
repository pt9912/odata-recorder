package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Die Fänger dieser Datei prüfen die Datei <DATA_DIR>/audit.jsonl im Betrieb über das Handle (SPEC-139 Nr. 2 bis 4, 6,
// 7, 10; SPEC-234 Nr. 3 und 4; SPEC-235 Nr. 2, 6, 7): echte Ablage, echte Dienste, echter Adapter; das Orakel ist die
// gelesene und geparste Datei und die Kennungen der Ausgangslage, nie ein Dienst.

// sollZeile ist eine erwartete Zeile der Datei: Art und die gesetzten Felder.
type sollZeile struct{ art, service, session, workspace, importID, code string }

// fall ist ein Fall F1 bis F7 des Plans: Modus, Art, Auslöser, Ausgangslage und erwartete Zeilen.
type fall struct {
	nr, modus  string
	art        support.AuditArt
	quelle     string
	maxAge     time.Duration
	saee       func(d *wsDatei)
	ausloeser  func(ctx context.Context, b *betrieb) (string, error) // nil: Hochfahren ist der Auslöser
	fehlerOK   func(err error) bool                                  // nil: der Auslöser liefert nil
	soll       []sollZeile
	weg        []string // gesäte Kennungen, die der Auslöser entfernt
	bleibt     []string // gesäte Kennungen, die bestehen bleiben
	fehlertext bool
}

func codeIst(c model.Code) func(error) bool {
	return func(err error) bool {
		var f *inbound.Fehler
		return errors.As(err, &f) && f.Code == c
	}
}

func auditFaelle() []fall {
	const alt, jung = 48 * time.Hour, time.Hour
	basis := func(d *wsDatei) {
		saeeDienst(d, "x1", "/r1")
		saeeDienst(d, "x2", "/r2")
	}
	aufbewahrung := func(d *wsDatei) {
		basis(d)
		for _, s := range []string{"x1", "x2"} {
			saeeSession(d, "ses-alt-"+s, s, alt, false)
			saeeSession(d, "ses-jung-"+s, s, jung, false)
		}
	}
	aufbewahrungSoll := []sollZeile{
		{art: "session_aufbewahrung", service: "x1", session: "ses-alt-x1"},
		{art: "session_aufbewahrung", service: "x2", session: "ses-alt-x2"},
	}
	weg, bleibt := []string{"ses-alt-x1", "ses-alt-x2"}, []string{"ses-jung-x1", "ses-jung-x2"}
	// portFehler ist der Fehler von Sessions.Get in F7: ohne Code, mit dem Canary im Text.
	portFehler := errors.New("Sessions.Get: " + string(fehlertext()))
	workspaceLoeschen := func(id string) func(ctx context.Context, b *betrieb) (string, error) {
		return func(ctx context.Context, b *betrieb) (string, error) {
			bericht, err := b.h.Workspaces().Delete(ctx, id)
			return fmt.Sprintf("%+v", bericht), err
		}
	}
	importieren := func(id, quelle string, fehler error) func(ctx context.Context, b *betrieb) (string, error) {
		return func(ctx context.Context, b *betrieb) (string, error) {
			b.sess.Fehler = fehler
			e, err := b.h.Import().Import(ctx, inbound.ImportAnfrage{ImportID: id, ServiceID: "x1", SourceSessionIDs: []string{quelle}})
			return fmt.Sprintf("%+v", e), err
		}
	}
	return []fall{
		{nr: "F1", modus: modusRecord, art: support.ArtSessionGeloescht, quelle: "session",
			saee: func(d *wsDatei) {
				basis(d)
				saeeSession(d, "ses-f1", "x1", jung, false)
				saeeSession(d, "ses-f1-bleibt", "x1", jung, false)
			},
			ausloeser: func(ctx context.Context, b *betrieb) (string, error) {
				bericht, err := b.h.Sessions().Delete(ctx, "ses-f1")
				return fmt.Sprintf("%+v", bericht), err
			},
			soll: []sollZeile{{art: "session_geloescht", service: "x1", session: "ses-f1"}},
			weg:  []string{"ses-f1"}, bleibt: []string{"ses-f1-bleibt"}},
		{nr: "F2", modus: modusRecord, art: support.ArtSessionAufbewahrung, quelle: "wartung", maxAge: 24 * time.Hour,
			saee: aufbewahrung, soll: aufbewahrungSoll, weg: weg, bleibt: bleibt},
		{nr: "F3", modus: modusReplay, art: support.ArtSessionAufbewahrung, quelle: "wartung", maxAge: 24 * time.Hour,
			saee: aufbewahrung, soll: aufbewahrungSoll, weg: weg, bleibt: bleibt},
		{nr: "F4", modus: modusRecord, art: support.ArtWorkspaceGeloescht, quelle: "workspace",
			saee:      func(d *wsDatei) { saeeWorkspaces(d, "ws-f4", "ws-f4-bleibt") },
			ausloeser: workspaceLoeschen("ws-f4"), soll: []sollZeile{{art: "workspace_geloescht", workspace: "ws-f4"}},
			weg: []string{"ws-f4"}, bleibt: []string{"ws-f4-bleibt"}},
		{nr: "F5", modus: modusReplay, art: support.ArtWorkspaceGeloescht, quelle: "workspace",
			saee:      func(d *wsDatei) { saeeWorkspaces(d, "ws-f5", "ws-f5-bleibt") },
			ausloeser: workspaceLoeschen("ws-f5"), soll: []sollZeile{{art: "workspace_geloescht", workspace: "ws-f5"}},
			weg: []string{"ws-f5"}, bleibt: []string{"ws-f5-bleibt"}},
		{nr: "F6", modus: modusRecord, art: support.ArtImportFehlgeschlagen, quelle: "import",
			saee: basis, ausloeser: importieren("imp-f6", "ses-unbekannt", nil), fehlerOK: codeIst(model.CodeE7003),
			soll: []sollZeile{{art: "import_fehlgeschlagen", service: "x1", importID: "imp-f6", code: "ODR-E7003"}}},
		{nr: "F7", modus: modusRecord, art: support.ArtImportFehlgeschlagen, quelle: "import", fehlertext: true,
			saee: func(d *wsDatei) {
				basis(d)
				saeeSession(d, "ses-f7", "x1", jung, false)
			},
			ausloeser: importieren("imp-f7", "ses-f7", portFehler),
			fehlerOK:  func(err error) bool { return codeIst(model.CodeE9001)(err) && errors.Is(err, portFehler) },
			soll:      []sollZeile{{art: "import_fehlgeschlagen", service: "x1", importID: "imp-f7", code: "ODR-E9001"}}},
	}
}

// ablauf sind die Beobachtungen eines Falls: die Datei zu den Zeitpunkten Z1 bis Z4, die Aufzeichnung der Einträge,
// das Ergebnis von Hochfahren und Auslöser und der Bestand der gesäten Kennungen.
type ablauf struct {
	dir        string
	z          [4]datei
	zHoch      datei
	rec        []Aufgezeichnet
	ergebnis   string
	err        error
	hochfahren error
	bestand    []string
	log        string
	ops        map[string][]string
	von, bis   time.Time
}

// lauf fährt den Fall: säen, Build, Hochfahren, Auslöser (Z1), Herunterfahren (Z2); mit neustart ein zweites Build
// auf demselben Datenverzeichnis ohne Aufbewahrung (Z3), Hochfahren und Herunterfahren (Z4). Mit stoerung ist
// audit.jsonl ein Verzeichnis.
func (f fall) lauf(t *testing.T, stoerung, neustart bool) ablauf {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	saeen(t, dir, f.saee)
	if stoerung {
		if err := os.Mkdir(filepath.Join(dir, "audit.jsonl"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	lies := func() datei {
		if stoerung {
			return datei{}
		}
		return liesAudit(t, dir)
	}
	b := neuerBetrieb(t, dir, f.modus, f.maxAge)
	b.bauen()
	a := ablauf{dir: dir}
	a.von = time.Now()
	a.hochfahren = b.h.Hochfahren(ctx)
	var text string
	if f.ausloeser != nil {
		text, a.err = f.ausloeser(ctx, b)
	}
	a.bis = time.Now()
	a.ergebnis = fmt.Sprintf("hochfahren=%v ausloeser=%q fehler=%v", a.hochfahren, text, a.err)
	a.z[0] = lies()
	if err := b.h.Herunterfahren(ctx); err != nil {
		t.Fatalf("%s: Herunterfahren: %v", f.nr, err)
	}
	a.z[1] = lies()
	a.rec, a.log, a.ops = b.rec.Eintraege(), b.log.String(), b.sess.Alle()
	a.bestand = bestandVon(t, dir, slices.Concat(f.weg, f.bleibt))
	if neustart {
		b2 := neuerBetrieb(t, dir, f.modus, 0)
		b2.bauen()
		a.z[2] = lies()
		if err := b2.h.Hochfahren(ctx); err != nil {
			t.Fatalf("%s: zweites Hochfahren: %v", f.nr, err)
		}
		a.zHoch = lies()
		if err := b2.h.Herunterfahren(ctx); err != nil {
			t.Fatalf("%s: zweites Herunterfahren: %v", f.nr, err)
		}
		a.z[3] = lies()
	}
	return a
}

func feldFolge() []string {
	return []string{"timestamp", "correlation_id", "art", "service_id", "session_id", "workspace_id", "import_id", "error_code"}
}

// zeilenVon parst die Datei zeilenweise (SPEC-235 Nr. 3): beendet durch Zeilenende, ohne Leerraum, ein Objekt aus
// Zeichenketten, die Felder in der Folge von feldFolge.
func zeilenVon(t *testing.T, name string, d datei) []map[string]string {
	t.Helper()
	if len(d.b) == 0 {
		return nil
	}
	if d.b[len(d.b)-1] != '\n' {
		t.Errorf("%s: die Datei endet nicht mit einem Zeilenende", name)
	}
	var aus []map[string]string
	for _, z := range bytes.Split(bytes.TrimSuffix(d.b, []byte("\n")), []byte("\n")) {
		m := map[string]string{}
		if err := json.Unmarshal(z, &m); err != nil {
			t.Errorf("%s: Zeile %q ist kein Objekt aus Zeichenketten: %v", name, z, err)
			continue
		}
		if bytes.ContainsAny(z, " \t\r") {
			t.Errorf("%s: Zeile %q trägt Leerraum", name, z)
		}
		letzte := -1
		for _, k := range feldFolge() {
			if _, da := m[k]; !da {
				continue
			}
			i := bytes.Index(z, []byte(`"`+k+`":`))
			if i < letzte {
				t.Errorf("%s: Feld %s steht vor einem früheren in %q", name, k, z)
			}
			letzte = i
		}
		aus = append(aus, m)
	}
	return aus
}

// feldFormen ist die Form jedes Feldes einer Zeile nach seiner Quelle: timestamp aus der Uhr des Roots (SPEC-235
// Nr. 3), correlation_id aus dem Generator des Roots (Präfix cor, 16 Zufallsbyte als Hex, SPEC-114), die Kennungen aus
// der Anfrage und der Ablage (SPEC-234 Nr. 3), error_code der Text eines Codes (SPEC-109); art prüft die Tabelle.
func feldFormen() map[string]*regexp.Regexp {
	kennung := regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)
	return map[string]*regexp.Regexp{
		"timestamp":      regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z$`),
		"correlation_id": regexp.MustCompile(`^cor-[0-9a-f]{32}$`),
		"service_id":     kennung, "session_id": kennung, "workspace_id": kennung, "import_id": kennung,
		"error_code": regexp.MustCompile(`^ODR-E\d{4}$`),
	}
}

// pruefeStruktur ist das Struktur-Orakel von SPEC-234 Nr. 2 als Literal: nur Felder der geschlossenen Menge, je Art
// genau die Pflichtfelder.
func pruefeStruktur(t *testing.T, name string, zeilen []map[string]string) {
	t.Helper()
	pflicht := map[string][]string{
		"session_geloescht":     {"service_id", "session_id"},
		"session_aufbewahrung":  {"service_id", "session_id"},
		"workspace_geloescht":   {"workspace_id"},
		"import_fehlgeschlagen": {"error_code", "import_id", "service_id"},
	}
	for _, m := range zeilen {
		soll, bekannt := pflicht[m["art"]]
		if !bekannt {
			t.Errorf("%s: Art %q ist keine der vier", name, m["art"])
			continue
		}
		var felder []string
		for k, v := range m {
			if !slices.Contains(feldFolge(), k) {
				t.Errorf("%s: Feld %q gehört nicht zur geschlossenen Menge", name, k)
			}
			if re, ok := feldFormen()[k]; ok && !re.MatchString(v) {
				t.Errorf("%s: Feld %s trägt %q, erwartet die Form %s", name, k, v, re)
			}
			if k != "timestamp" && k != "correlation_id" && k != "art" {
				felder = append(felder, k)
			}
			if v == "" {
				t.Errorf("%s: Feld %q ist leer", name, k)
			}
		}
		slices.Sort(felder)
		if !slices.Equal(felder, soll) {
			t.Errorf("%s: Art %s trägt die Felder %v, erwartet %v", name, m["art"], felder, soll)
		}
	}
}

func sortiert(zs []sollZeile) []string {
	var aus []string
	for _, z := range zs {
		aus = append(aus, fmt.Sprintf("%+v", z))
	}
	slices.Sort(aus)
	return aus
}

// pruefeInhalt ist das Orakel von M1: die Zeilen der Datei gegen die Kennungen der Ausgangslage, Zeit in UTC mit
// neun Stellen im Fenster des Aufrufs, Korrelations-ID in der Form einer Kennung und je Zeile verschieden.
func pruefeInhalt(t *testing.T, name string, zeilen []map[string]string, soll []sollZeile, von, bis time.Time) {
	t.Helper()
	zeit := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z$`)
	kennung := regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)
	var ist []sollZeile
	ids := map[string]bool{}
	for _, m := range zeilen {
		ist = append(ist, sollZeile{m["art"], m["service_id"], m["session_id"], m["workspace_id"], m["import_id"], m["error_code"]})
		ts, err := time.Parse(time.RFC3339Nano, m["timestamp"])
		if !zeit.MatchString(m["timestamp"]) || err != nil || ts.Before(von.Add(-time.Second)) || ts.After(bis.Add(time.Second)) {
			t.Errorf("%s: timestamp %q (%v) nicht in UTC mit neun Stellen im Fenster %v bis %v", name, m["timestamp"], err, von, bis)
		}
		if !kennung.MatchString(m["correlation_id"]) || ids[m["correlation_id"]] {
			t.Errorf("%s: correlation_id %q ist keine Kennung oder doppelt", name, m["correlation_id"])
		}
		ids[m["correlation_id"]] = true
	}
	if !slices.Equal(sortiert(ist), sortiert(soll)) {
		t.Errorf("%s: Zeilen %+v, erwartet %+v", name, ist, soll)
	}
}

// pruefeAufzeichnung: die Dienste übergeben genau die erwarteten Einträge, aus der erwarteten Konfig, ohne
// Korrelations-ID (SPEC-234 Nr. 7).
func pruefeAufzeichnung(t *testing.T, f fall, rec []Aufgezeichnet) {
	t.Helper()
	var ist []sollZeile
	for _, r := range rec {
		e := r.Eintrag
		ist = append(ist, sollZeile{string(e.Art), e.ServiceID, e.SessionID, e.WorkspaceID, e.ImportID, e.ErrorCode.String()})
		if r.Quelle != f.quelle || e.KorrelationsID != "" {
			t.Errorf("%s: Eintrag %+v aus %q mit Korrelations-ID %q, erwartet aus %q ohne", f.nr, e, r.Quelle, e.KorrelationsID, f.quelle)
		}
	}
	if !slices.Equal(sortiert(ist), sortiert(f.soll)) {
		t.Errorf("%s: Aufgezeichnet %+v, erwartet %+v", f.nr, ist, f.soll)
	}
}

// TestAuditEintraegeJeFall fängt V01a bis V01d (M1): jeder Auslöser der sieben Fälle schreibt die erwarteten Zeilen in
// die Datei, in der Struktur von SPEC-234 Nr. 2, und lässt Gesäetes bestehen, das er nicht trifft.
func TestAuditEintraegeJeFall(t *testing.T) {
	faelle := auditFaelle()
	if len(faelle) != 7 {
		t.Fatalf("%d Fälle, erwartet 7", len(faelle))
	}
	for _, f := range faelle {
		t.Run(f.nr, func(t *testing.T) {
			a := f.lauf(t, false, false)
			if a.hochfahren != nil {
				t.Fatalf("Hochfahren: %v", a.hochfahren)
			}
			if (f.fehlerOK == nil) != (a.err == nil) || (f.fehlerOK != nil && !f.fehlerOK(a.err)) {
				t.Errorf("Fehler des Auslösers: %v", a.err)
			}
			zeilen := zeilenVon(t, f.nr, a.z[0])
			pruefeStruktur(t, f.nr, zeilen)
			pruefeInhalt(t, f.nr, zeilen, f.soll, a.von, a.bis)
			pruefeAufzeichnung(t, f, a.rec)
			if !slices.Equal(a.bestand, sortierteKopie(f.bleibt)) {
				t.Errorf("Bestand %v, erwartet %v (entfernt: %v)", a.bestand, f.bleibt, f.weg)
			}
			// Der Import ruft nach dem Fehler von Get keine weitere Operation des Repositorys (SPEC-127 Nr. 4).
			if f.quelle == "import" {
				ids := map[string]string{"F6": "ses-unbekannt", "F7": "ses-f7"}
				if want := map[string][]string{"Get": {ids[f.nr]}}; !reflect.DeepEqual(a.ops, want) {
					t.Errorf("Aufrufe von Sessions im Import: %v, erwartet %v", a.ops, want)
				}
			}
			if strings.Contains(a.log, "audit: Eintrag nicht geschrieben") {
				t.Errorf("gesunde Ablage, aber das Log meldet einen Ausfall: %s", a.log)
			}
		})
	}
}

func sortierteKopie(s []string) []string {
	s = slices.Clone(s)
	slices.Sort(s)
	return s
}

// TestAuditNeustart fängt V02a: Herunterfahren schreibt nichts, und ein zweites Build auf demselben Datenverzeichnis
// ändert die Datei weder nach Build noch nach Hochfahren noch nach Herunterfahren (Z2 gleich Z1, Z3, Hochfahren, Z4).
func TestAuditNeustart(t *testing.T) {
	for _, f := range auditFaelle() {
		t.Run(f.nr, func(t *testing.T) {
			a := f.lauf(t, false, true)
			if !a.z[1].da || len(a.z[1].b) == 0 {
				t.Fatalf("Z2 ohne Zeilen: %+v", a.z[1])
			}
			stand := map[string]datei{"Z1": a.z[0], "Z3": a.z[2], "Hochfahren": a.zHoch, "Z4": a.z[3]}
			for name, d := range stand {
				if d.da != a.z[1].da || !bytes.Equal(d.b, a.z[1].b) {
					t.Errorf("%s ist nicht gleich Z2: %q gegen %q", name, d.b, a.z[1].b)
				}
			}
		})
	}
}

// TestAuditNeustartHaengtAn fängt V02b: löscht der zweite Lauf über Hochfahren eine weitere gesäte Session, stehen die
// Bytes des ersten Laufs als Anfang der Datei, danach genau eine neue Zeile nach dem Orakel.
func TestAuditNeustartHaengtAn(t *testing.T) {
	ctx := context.Background()
	f := auditFaelle()[1]
	dir := t.TempDir()
	saeen(t, dir, f.saee)
	b1 := neuerBetrieb(t, dir, f.modus, f.maxAge)
	b1.bauen()
	if err := b1.h.Hochfahren(ctx); err != nil {
		t.Fatal(err)
	}
	vorStartup := liesAudit(t, dir)
	// Ein zweites Startup im selben Lauf findet nichts mehr und schreibt nichts (SPEC-234 Nr. 10).
	if _, err := b1.h.Wartung().Startup(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b1.h.Herunterfahren(ctx); err != nil {
		t.Fatal(err)
	}
	eins := liesAudit(t, dir)
	if len(zeilenVon(t, "Lauf 1", eins)) != 2 || !bytes.Equal(eins.b, vorStartup.b) {
		t.Fatalf("Lauf 1: %q, vor dem zweiten Startup %q", eins.b, vorStartup.b)
	}
	saeen(t, dir, func(d *wsDatei) { saeeSession(d, "ses-alt2-x1", "x1", 72*time.Hour, false) })
	b2 := neuerBetrieb(t, dir, f.modus, f.maxAge)
	b2.bauen()
	if d := liesAudit(t, dir); !bytes.Equal(d.b, eins.b) {
		t.Errorf("Build des zweiten Laufs änderte die Datei: %q", d.b)
	}
	von := time.Now()
	if err := b2.h.Hochfahren(ctx); err != nil {
		t.Fatal(err)
	}
	bis := time.Now()
	if err := b2.h.Herunterfahren(ctx); err != nil {
		t.Fatal(err)
	}
	zwei := liesAudit(t, dir)
	if !bytes.HasPrefix(zwei.b, eins.b) {
		t.Fatalf("die Bytes des ersten Laufs stehen nicht am Anfang: %q gegen %q", zwei.b, eins.b)
	}
	neu := datei{b: zwei.b[len(eins.b):], da: true}
	zeilen := zeilenVon(t, "Lauf 2", neu)
	pruefeStruktur(t, "Lauf 2", zeilen)
	pruefeInhalt(t, "Lauf 2", zeilen, []sollZeile{{art: "session_aufbewahrung", service: "x1", session: "ses-alt2-x1"}}, von, bis)
	if n := len(zeilenVon(t, "gesamt", zwei)); n != 3 {
		t.Errorf("%d Zeilen, erwartet 3", n)
	}
	if got := bestandVon(t, dir, []string{"ses-alt2-x1", "ses-jung-x1"}); !slices.Equal(got, []string{"ses-jung-x1"}) {
		t.Errorf("Bestand %v", got)
	}
}

// TestAuditCanary fängt V03 (M2): kein Wert der Zugangsdaten und kein Fehlertext steht in einem Byte der Datei, roh,
// prozent-kodiert oder als Base64 von benutzer:passwort, und kein Dienst übergibt einen Eintrag, der ihn trägt; 172
// Prüfungen aus Wert × Fall × Zeitpunkt.
func TestAuditCanary(t *testing.T) {
	n := 0
	ids := map[string]string{}
	ft := geheimnis{"fehlertext", [][]byte{fehlertext()}}
	for _, f := range auditFaelle() {
		a := f.lauf(t, false, true)
		werte := canaryWerte()
		if f.fehlertext {
			werte = append(werte, ft)
		}
		for z, d := range a.z {
			pruefeStruktur(t, fmt.Sprintf("%s Z%d", f.nr, z+1), zeilenVon(t, f.nr, d))
			for _, w := range werte {
				n++
				if form := w.findet(d.b); form != "" {
					t.Errorf("%s Z%d: %s steht als %q in der Datei", f.nr, z+1, w.name, form)
				}
			}
		}
		// Die Generator-ID ist zufällig: über alle Fälle und Läufe kehrt keine wieder (ein abgeleiteter, fester Wert wiederholt sich).
		for _, m := range zeilenVon(t, f.nr, a.z[1]) {
			if vorher, doppelt := ids[m["correlation_id"]]; doppelt {
				t.Errorf("%s: correlation_id %q kehrt aus %s wieder", f.nr, m["correlation_id"], vorher)
			}
			ids[m["correlation_id"]] = f.nr
		}
		for _, r := range a.rec {
			e := r.Eintrag
			for _, feld := range []string{string(e.Art), e.ServiceID, e.SessionID, e.WorkspaceID, e.ImportID, e.ErrorCode.String(), e.KorrelationsID} {
				for _, w := range werte {
					if form := w.findet([]byte(feld)); form != "" {
						t.Errorf("%s: ein Dienst übergab %s als %q in %+v", f.nr, w.name, form, e)
					}
				}
			}
		}
	}
	if n != 172 {
		t.Errorf("%d Prüfungen, erwartet (6 · 7 + 1) · 4 = 172", n)
	}
	if len(ids) != 9 {
		t.Errorf("%d verschiedene correlation_id, erwartet 9 (eine je Zeile der sieben Fälle)", len(ids))
	}
}

// TestAuditCanaryOrakel fängt das Orakel selbst: die zwei Klassen haben die Form, die der Plan nennt, und jede Form
// jedes der sechs Werte wird in einer Zeile gefunden, in einer sauberen nicht.
func TestAuditCanaryOrakel(t *testing.T) {
	kennung := regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)
	for _, z := range zugaenge() {
		for _, b := range [][]byte{z.benutzer, z.passwort, z.anfrage} {
			if kennung.Match(b) != (z.klasse == "gueltig") {
				t.Errorf("Klasse %s: %q hat die Kennungsform nicht wie verlangt", z.klasse, b)
			}
			if z.klasse == "fremd" && !bytes.ContainsFunc(b, func(r rune) bool { return r >= 0x80 }) {
				t.Errorf("%q trägt kein Byte ab 0x80", b)
			}
		}
	}
	werte := canaryWerte()
	if len(werte) != 6 {
		t.Fatalf("%d Werte, erwartet 6", len(werte))
	}
	for _, w := range werte {
		if mindest := map[bool]int{true: 3, false: 1}[strings.HasPrefix(w.name, "fremd")]; len(w.formen) < mindest {
			t.Errorf("%s: %d Formen, erwartet mindestens %d", w.name, len(w.formen), mindest)
		}
		for _, f := range w.formen {
			zeile := append(append([]byte(`{"art":"x","correlation_id":"`), f...), []byte(`"}`)...)
			if w.findet(zeile) == "" {
				t.Errorf("%s: die Form %q wird nicht gefunden", w.name, f)
			}
		}
		if w.findet([]byte(`{"art":"session_geloescht","correlation_id":"cor-1"}`)) != "" {
			t.Errorf("%s: gefunden in einer sauberen Zeile", w.name)
		}
	}
}

// TestAuditBuildNimmtZugangsdaten fängt die Voraussetzung von V03 (SPEC-139 Nr. 4): Build und Hochfahren nehmen eine
// UpstreamURL mit Userinfo und Query an, formgültig und formfremd, in beiden Modi.
func TestAuditBuildNimmtZugangsdaten(t *testing.T) {
	for _, modus := range modi() {
		k := canaryKonfig(t.TempDir(), modus, 0)
		for i, d := range k.Dienste {
			u, err := url.Parse(d.UpstreamURL)
			z := zugaenge()[i]
			pw, _ := u.User.Password()
			if err != nil || u.User.Username() != string(z.benutzer) || pw != string(z.passwort) || u.Query().Get("tok") != string(z.anfrage) {
				t.Fatalf("%s: die URL von %s trägt die Zugangsdaten nicht: %q (%v)", modus, d.ID, d.UpstreamURL, err)
			}
		}
		b := neuerBetrieb(t, k.DataDir, modus, 0)
		b.bauen()
		if err := b.h.Hochfahren(context.Background()); err != nil {
			t.Errorf("%s: Hochfahren: %v", modus, err)
		}
	}
}

// TestAuditOhneLoeschungKeineDatei fängt V05a: Hochfahren und Herunterfahren ohne Löschung legen die Datei nicht an
// und rufen Append nie, auch nicht bei einer Aufbewahrung ohne Treffer (die junge Session bleibt) und nicht beim
// zweiten Herunterfahren, nicht beim zweiten Hochfahren und zweiten Startup, nicht bei Herunterfahren ohne
// Hochfahren und nicht bei Hochfahren mit beendetem ctx; in beiden Modi.
func TestAuditOhneLoeschungKeineDatei(t *testing.T) {
	ctx := context.Background()
	for _, modus := range modi() {
		for _, mitJunger := range []bool{false, true} {
			name := fmt.Sprintf("%s/junge=%v", modus, mitJunger)
			dir := t.TempDir()
			maxAge := time.Duration(0)
			if mitJunger {
				maxAge = 24 * time.Hour
			}
			saeen(t, dir, func(d *wsDatei) {
				saeeDienst(d, "x1", "/r1")
				if mitJunger {
					saeeSession(d, "ses-jung", "x1", time.Hour, false)
				}
			})
			b := neuerBetrieb(t, dir, modus, maxAge)
			keine := func(wann string) { keineDatei(t, name, wann, b) }
			b.bauen()
			keine("Build")
			if err := b.h.Hochfahren(ctx); err != nil {
				t.Fatalf("%s: Hochfahren: %v", name, err)
			}
			keine("Hochfahren")
			if err := b.h.Hochfahren(ctx); !errors.Is(err, model.ErrUngueltig) {
				t.Errorf("%s: zweites Hochfahren: %v", name, err)
			}
			if _, err := b.h.Wartung().Startup(ctx); err != nil {
				t.Errorf("%s: zweites Startup: %v", name, err)
			}
			keine("zweitem Hochfahren und Startup")
			for i := 1; i <= 2; i++ {
				if err := b.h.Herunterfahren(ctx); err != nil {
					t.Fatalf("%s: Herunterfahren: %v", name, err)
				}
				keine(fmt.Sprintf("Herunterfahren %d", i))
			}
			if mitJunger && len(bestandVon(t, dir, []string{"ses-jung"})) != 1 {
				t.Errorf("%s: die junge Session wurde gelöscht", name)
			}
		}
		// Herunterfahren ohne Hochfahren schließt nur die Ablage; ein beendeter ctx vor Schritt 1 tut nichts.
		o := neuerBetrieb(t, t.TempDir(), modus, 0)
		o.bauen()
		if err := o.h.Herunterfahren(ctx); err != nil {
			t.Errorf("%s: Herunterfahren ohne Hochfahren: %v", modus, err)
		}
		keineDatei(t, modus, "Herunterfahren ohne Hochfahren", o)
		a := neuerBetrieb(t, t.TempDir(), modus, 0)
		a.bauen()
		abgebrochen, abbruch := context.WithCancel(ctx)
		abbruch()
		if err := a.h.Hochfahren(abgebrochen); !errors.Is(err, context.Canceled) {
			t.Errorf("%s: Hochfahren mit beendetem ctx: %v", modus, err)
		}
		keineDatei(t, modus, "Hochfahren mit beendetem ctx", a)
	}
}

// keineDatei meldet, wenn im Datenverzeichnis des Betriebs audit.jsonl steht oder Append gerufen wurde.
func keineDatei(t *testing.T, name, wann string, b *betrieb) {
	t.Helper()
	if d := liesAudit(t, b.dir); d.da || len(b.rec.Eintraege()) != 0 {
		t.Errorf("%s: nach %s steht audit.jsonl (%q) oder wurde Append gerufen: %+v", name, wann, d.b, b.rec.Eintraege())
	}
}

// TestAuditAblehnungenKeinEintrag fängt V05b (M3): eine Ablehnung vor der Tatsache hinterlässt keinen Eintrag und
// keine Datei, eine Wiederholung keinen zweiten Eintrag.
func TestAuditAblehnungenKeinEintrag(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	saeen(t, dir, func(d *wsDatei) {
		saeeDienst(d, "x1", "/r1")
		saeeSession(d, "ses-fertig", "x1", time.Hour, false)
		saeeSession(d, "ses-offen", "x1", time.Hour, true)
		saeeWorkspaces(d) // Service svc, Session ses-1 als Quelle der Baseline bsl-1
	})
	b := neuerBetrieb(t, dir, modusRecord, 0)
	b.bauen()
	loesche := func(id string) func() error {
		return func() error { _, err := b.h.Sessions().Delete(ctx, id); return err }
	}
	faelle := []struct {
		name  string
		code  model.Code
		rufe  func() error
		zeile int
	}{
		{"Session unbekannt", model.CodeE7003, loesche("ses-gibt-es-nicht"), 0},
		{"Session leere Kennung", model.CodeE7003, loesche(""), 0},
		{"Session offen", model.CodeE7009, loesche("ses-offen"), 0},
		{"Session von einer Baseline referenziert", model.CodeE7004, loesche("ses-1"), 0},
		{"Workspace unbekannt", model.CodeE7001, func() error { _, err := b.h.Workspaces().Delete(ctx, "ws-gibt-es-nicht"); return err }, 0},
		{"Import, Dienst unbekannt", model.CodeE5005, func() error {
			_, err := b.h.Import().Import(ctx, inbound.ImportAnfrage{ImportID: "imp-x", ServiceID: "unbekannt", SourceSessionIDs: []string{"ses-fertig"}})
			return err
		}, 0},
		{"Wiederholung", model.CodeE7003, func() error {
			if err := loesche("ses-fertig")(); err != nil {
				t.Fatalf("erstes Delete: %v", err)
			}
			return loesche("ses-fertig")()
		}, 1},
	}
	if len(faelle) != 7 {
		t.Fatalf("%d Ablehnungen, erwartet 7", len(faelle))
	}
	for _, f := range faelle {
		if !codeIst(f.code)(f.rufe()) {
			t.Errorf("%s: Fehler nicht %v", f.name, f.code)
		}
		d := liesAudit(t, dir)
		if n := len(zeilenVon(t, f.name, d)); n != f.zeile || len(b.rec.Eintraege()) != f.zeile || d.da != (f.zeile > 0) {
			t.Errorf("%s: %d Zeilen (Datei da: %v) und %d Aufrufe von Append, erwartet %d", f.name, n, d.da, len(b.rec.Eintraege()), f.zeile)
		}
	}
}

// TestAuditBuildFehlerOhneDatei fängt V05c (M4): ein Fehler der Ablage beim Build kommt zurück, audit.jsonl entsteht
// nicht, Append wird nicht gerufen; in beiden Modi.
func TestAuditBuildFehlerOhneDatei(t *testing.T) {
	bilder := []struct {
		name        string
		vorbereiten func(t *testing.T) string
	}{
		{"DATA_DIR fehlt", func(t *testing.T) string { return filepath.Join(t.TempDir(), "gibt-es-nicht") }},
		{"DATA_DIR ist eine Datei", func(t *testing.T) string {
			p := filepath.Join(t.TempDir(), "datei")
			if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
			return p
		}},
		{"Datenbank ist ein Verzeichnis", func(t *testing.T) string {
			d := t.TempDir()
			if err := os.Mkdir(filepath.Join(d, dateiDB), 0o700); err != nil {
				t.Fatal(err)
			}
			return d
		}},
	}
	if len(bilder) != 3 {
		t.Fatalf("%d Fehlerbilder, erwartet 3", len(bilder))
	}
	for _, modus := range modi() {
		for _, bild := range bilder {
			dir := bild.vorbereiten(t)
			b := neuerBetrieb(t, dir, modus, 0)
			if err := b.baue(); err == nil {
				t.Errorf("%s/%s: Build gelang", modus, bild.name)
			}
			if _, err := os.Stat(filepath.Join(dir, "audit.jsonl")); err == nil {
				t.Errorf("%s/%s: audit.jsonl entstand", modus, bild.name)
			}
			if ents, err := os.ReadDir(dir); err == nil {
				for _, e := range ents {
					if e.Name() != dateiDB {
						t.Errorf("%s/%s: Build legte %s an", modus, bild.name, e.Name())
					}
				}
			} else if !errors.Is(err, fs.ErrNotExist) && !strings.Contains(err.Error(), "not a directory") {
				t.Errorf("%s/%s: %v", modus, bild.name, err)
			}
			if len(b.rec.Eintraege()) != 0 {
				t.Errorf("%s/%s: Append wurde gerufen", modus, bild.name)
			}
		}
	}
}

// TestAuditSchreibfehlerAendertNichts fängt V06: ist audit.jsonl ein Verzeichnis, gelingt Build, und in jedem Fall F1
// bis F7 sind Hochfahren, Ergebnis und Fehler des Auslösers und der Bestand der Ablage dieselben wie bei gesunder
// Ablage; das Log trägt je Eintrag genau eine Zeile des Ausfalls, ohne Pfad und ohne Canary.
func TestAuditSchreibfehlerAendertNichts(t *testing.T) {
	for _, f := range auditFaelle() {
		t.Run(f.nr, func(t *testing.T) {
			gesund := f.lauf(t, false, false)
			a := f.lauf(t, true, false)
			if a.ergebnis != gesund.ergebnis || !slices.Equal(a.bestand, gesund.bestand) {
				t.Errorf("gestörte Ablage: %s, Bestand %v; gesund: %s, Bestand %v", a.ergebnis, a.bestand, gesund.ergebnis, gesund.bestand)
			}
			pruefeAufzeichnung(t, f, a.rec)
			pfad := filepath.Join(a.dir, "audit.jsonl")
			if fi, err := os.Stat(pfad); err != nil || !fi.IsDir() {
				t.Errorf("audit.jsonl ist kein Verzeichnis mehr: %v, %v", fi, err)
			}
			if ents, err := os.ReadDir(pfad); err != nil || len(ents) != 0 {
				t.Errorf("audit.jsonl trägt Einträge: %v, %v", ents, err)
			}
			n := 0
			for _, zeile := range strings.Split(a.log, "\n") {
				if !strings.Contains(zeile, `msg="audit: Eintrag nicht geschrieben"`) {
					continue
				}
				n++
				for _, teil := range []string{"level=ERROR", "ereignis=audit_fehlgeschlagen", "error_code=ODR-E9001"} {
					if !strings.Contains(zeile, teil) {
						t.Errorf("Zeile %q ohne %s", zeile, teil)
					}
				}
				if strings.Contains(zeile, "audit.jsonl") || strings.Contains(zeile, os.TempDir()) {
					t.Errorf("Zeile %q nennt den Pfad", zeile)
				}
				for _, w := range append(canaryWerte(), geheimnis{"fehlertext", [][]byte{fehlertext()}}) {
					if form := w.findet([]byte(zeile)); form != "" {
						t.Errorf("Zeile %q trägt %s", zeile, w.name)
					}
				}
			}
			if n != len(f.soll) {
				t.Errorf("%d Zeilen des Ausfalls im Log, erwartet %d", n, len(f.soll))
			}
		})
	}
}

// TestAuditAufrufstellen fängt V07: die Aufrufstellen von Audit.Append im Code der Dienste sind die vier Zeilen
// V01a bis V01d (per Syntaxbaum gegen das Literal W), die Arten sind vier, und die Zellen (Modus, Art) der Fälle sind
// Modus × Arten ohne (replay, session_geloescht) und (replay, import_fehlgeschlagen).
func TestAuditAufrufstellen(t *testing.T) {
	soll := []string{
		"baselineimport/importer.go ArtImportFehlgeschlagen",
		"maintenance/startup.go ArtSessionAufbewahrung",
		"session/loeschen.go ArtSessionGeloescht",
		"workspace/loeschen.go ArtWorkspaceGeloescht",
	}
	wurzel := filepath.Join("..", "hexagon", "services")
	var ist, sonst []string
	err := filepath.WalkDir(wurzel, func(pfad string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(token.NewFileSet(), pfad, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(wurzel, pfad)
		erlaubt := map[*ast.SelectorExpr]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Append" {
					if feld, ok := sel.X.(*ast.SelectorExpr); ok && feld.Sel.Name == "Audit" {
						erlaubt[feld] = true
						ist = append(ist, filepath.ToSlash(rel)+" "+artDesAufrufs(x))
					}
				}
			case *ast.BinaryExpr:
				// k.Audit == nil in New ist keine Aufrufstelle.
				for _, seite := range []ast.Expr{x.X, x.Y} {
					if feld, ok := seite.(*ast.SelectorExpr); ok && feld.Sel.Name == "Audit" && (x.Op == token.EQL || x.Op == token.NEQ) {
						erlaubt[feld] = true
					}
				}
			case *ast.SelectorExpr:
				// Jede andere Verwendung des Feldes Audit (Alias, Methodenwert, Übergabe) ist ein Fund außerhalb von W.
				if x.Sel.Name == "Audit" && !erlaubt[x] {
					sonst = append(sonst, filepath.ToSlash(rel))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(ist)
	if !slices.Equal(ist, soll) {
		t.Errorf("Aufrufstellen von Audit.Append: %v, erwartet %v", ist, soll)
	}
	if len(sonst) != 0 {
		t.Errorf("Verwendungen des Feldes Audit, die kein Aufruf von Append und kein Vergleich mit nil sind: %v", sonst)
	}
	if n := len(support.Arten()); n != 4 {
		t.Errorf("%d Arten, erwartet 4", n)
	}
	var zellen, erwartet []string
	for _, f := range auditFaelle() {
		zellen = append(zellen, f.modus+" "+string(f.art))
	}
	for _, modus := range modi() {
		for _, s := range support.Arten() {
			if modus == modusReplay && (s.Art == support.ArtSessionGeloescht || s.Art == support.ArtImportFehlgeschlagen) {
				continue
			}
			erwartet = append(erwartet, modus+" "+string(s.Art))
		}
	}
	slices.Sort(zellen)
	slices.Sort(erwartet)
	if got := slices.Compact(zellen); !slices.Equal(got, erwartet) || len(erwartet) != 6 {
		t.Errorf("Zellen der Fälle %v, erwartet Modus × Arten ohne zwei: %v", got, erwartet)
	}
}

// artDesAufrufs liefert den Namen der Konstante im Feld Art des Eintrags, den der Aufruf übergibt.
func artDesAufrufs(c *ast.CallExpr) string {
	if len(c.Args) == 1 {
		if lit, ok := c.Args[0].(*ast.CompositeLit); ok {
			for _, el := range lit.Elts {
				kv, ok := el.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if k, ok := kv.Key.(*ast.Ident); ok && k.Name == "Art" {
					if v, ok := kv.Value.(*ast.SelectorExpr); ok {
						return v.Sel.Name
					}
				}
			}
		}
	}
	return "?"
}

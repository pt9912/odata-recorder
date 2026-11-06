package audit_test

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/audit"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// A02: je Pflichtfeld von Config ein Fall; der Wert steht nicht im Text, die Datei bleibt unberührt.
func TestNewWeistAb(t *testing.T) {
	basis := neu(t, festeZeit)
	gut := func(pfad string) audit.Config {
		return audit.Config{Pfad: pfad, Uhr: basis.uhr, IDs: basis.ids, Log: slog.New(slog.NewJSONHandler(io.Discard, nil))}
	}
	canary := "/CANARY-new/audit.jsonl"
	faelle := map[string]audit.Config{}
	c := gut(canary)
	c.Uhr = nil
	faelle["Uhr"] = c
	c = gut(canary)
	c.IDs = nil
	faelle["IDs"] = c
	c = gut(canary)
	c.Log = nil
	faelle["Log"] = c
	faelle["Pfad"] = gut("")
	if len(faelle) != 4 {
		t.Fatalf("%d Fälle, erwartet 4", len(faelle))
	}
	for name, cfg := range faelle {
		d, err := audit.New(cfg)
		if !errors.Is(err, model.ErrUngueltig) || d != nil {
			t.Errorf("%s fehlt: New = (%v, %v), erwartet (nil, ErrUngueltig)", name, d, err)
			continue
		}
		if strings.Contains(err.Error(), "CANARY") {
			t.Errorf("%s fehlt: der Text nennt den Pfad: %q", name, err)
		}
	}
}

// A02: New gelingt für einen Pfad in einem fehlenden Verzeichnis und berührt nichts.
func TestNewBeruehrtDateiNicht(t *testing.T) {
	wurzel := t.TempDir()
	fehlend := filepath.Join(wurzel, "fehlt", audit.DateiName)
	neuAn(t, fehlend, festeZeit)
	if _, err := os.Stat(filepath.Join(wurzel, "fehlt")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("New hat das Verzeichnis angelegt: %v", err)
	}
	vorhanden := filepath.Join(wurzel, "da.jsonl")
	if err := os.WriteFile(vorhanden, []byte("alt\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	neuAn(t, vorhanden, festeZeit)
	noch := filepath.Join(wurzel, "noch.jsonl")
	neuAn(t, noch, festeZeit)
	if _, err := os.Stat(noch); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("New hat die Datei angelegt: %v", err)
	}
	if b, _ := os.ReadFile(vorhanden); string(b) != "alt\n" {
		t.Errorf("New hat die vorhandene Datei geändert: %q", b)
	}
}

// A03: die Zeile je Art als Bytewerte der Spezifikation (SPEC-235 Nr. 3); die Datei ist
// auf Rohbytes gegengelesen (LC_ALL=C grep -P).
func TestZeileJeArt(t *testing.T) {
	a := neu(t, festeZeit)
	zeit := `{"timestamp":"2026-11-06T10:00:00.000000000Z","correlation_id":"`
	soll := []string{
		zeit + corID(1) + `","art":"session_geloescht","service_id":"w-service_id","session_id":"w-session_id"}` + "\n",
		zeit + corID(2) + `","art":"session_aufbewahrung","service_id":"w-service_id","session_id":"w-session_id"}` + "\n",
		zeit + corID(3) + `","art":"workspace_geloescht","workspace_id":"w-workspace_id"}` + "\n",
		zeit + corID(4) + `","art":"import_fehlgeschlagen","service_id":"w-service_id","import_id":"w-import_id","error_code":"ODR-E6101"}` + "\n",
	}
	arten := []support.AuditArt{support.ArtSessionGeloescht, support.ArtSessionAufbewahrung, support.ArtWorkspaceGeloescht, support.ArtImportFehlgeschlagen}
	if len(arten) != len(soll) || len(arten) != 4 {
		t.Fatalf("%d Arten und %d Zeilen, erwartet je 4", len(arten), len(soll))
	}
	for _, art := range arten {
		if err := a.d.Append(gueltig(art)); err != nil {
			t.Fatalf("Append %s: %v", art, err)
		}
	}
	if got, want := a.inhalt(t), strings.Join(soll, ""); got != want {
		t.Errorf("Datei:\n%q\nerwartet:\n%q", got, want)
	}
	if got := a.inhalt(t); strings.ContainsAny(got, " \t\r\\") || strings.Count(got, "\n") != 4 {
		t.Errorf("Leerraum, Maskierung oder Zeilenzahl falsch: %q", got)
	}
}

// A04: neun Nachkommastellen in UTC an den Rändern; die Zeit eines anderen Ortes wird umgerechnet.
func TestZeitpunkt(t *testing.T) {
	plus2 := time.FixedZone("plus2", 2*60*60)
	faelle := []struct {
		name string
		zeit time.Time
		soll string
	}{
		{"0 ns", time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC), "2026-11-06T10:00:00.000000000Z"},
		{"1 ns", time.Date(2026, 11, 6, 10, 0, 0, 1, time.UTC), "2026-11-06T10:00:00.000000001Z"},
		{"10 ns", time.Date(2026, 11, 6, 10, 0, 0, 10, time.UTC), "2026-11-06T10:00:00.000000010Z"},
		{"100 ms", time.Date(2026, 11, 6, 10, 0, 0, 100000000, time.UTC), "2026-11-06T10:00:00.100000000Z"},
		{"999999999 ns", time.Date(2026, 11, 6, 10, 0, 0, 999999999, time.UTC), "2026-11-06T10:00:00.999999999Z"},
		{"Jahr 1", time.Date(1, 1, 1, 0, 0, 0, 1, time.UTC), "0001-01-01T00:00:00.000000001Z"},
		{"Jahr 9999", time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC), "9999-12-31T23:59:59.999999999Z"},
		{"anderer Ort", time.Date(2026, 11, 6, 12, 0, 0, 5, plus2), "2026-11-06T10:00:00.000000005Z"},
	}
	for _, f := range faelle {
		a := neu(t, func(int) time.Time { return f.zeit })
		if err := a.d.Append(gueltig(support.ArtWorkspaceGeloescht)); err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		if want := `{"timestamp":"` + f.soll + `","correlation_id":`; !strings.HasPrefix(a.inhalt(t), want) {
			t.Errorf("%s: Zeile %q, erwartet Anfang %q", f.name, a.inhalt(t), want)
		}
	}
	if len(faelle) != 8 {
		t.Fatalf("%d Fälle, erwartet 8", len(faelle))
	}
}

// A05: eine leere KorrelationsID ergibt genau eine Kennung des Generators; eine gesetzte gilt
// unverändert, der Generator wird nicht gerufen (kein Aufruf irgendeiner seiner Operationen).
func TestKorrelationsID(t *testing.T) {
	a := neu(t, festeZeit)
	e := gueltig(support.ArtSessionGeloescht)
	if err := a.d.Append(e); err != nil {
		t.Fatal(err)
	}
	if a.ids.Korrelation() != 1 || a.ids.Gesamt() != 1 {
		t.Errorf("leere ID: %d Aufrufe von KorrelationsID, %d insgesamt, erwartet 1 und 1", a.ids.Korrelation(), a.ids.Gesamt())
	}
	if err := a.d.Append(e); err != nil {
		t.Fatal(err)
	}
	e.KorrelationsID = "gesetzt-1"
	if err := a.d.Append(e); err != nil {
		t.Fatal(err)
	}
	if a.ids.Gesamt() != 2 {
		t.Errorf("gesetzte ID: %d Aufrufe insgesamt, erwartet 2", a.ids.Gesamt())
	}
	zeilen := strings.Split(strings.TrimSuffix(a.inhalt(t), "\n"), "\n")
	if len(zeilen) != 3 {
		t.Fatalf("%d Zeilen in der Datei, erwartet 3", len(zeilen))
	}
	for i, id := range []string{corID(1), corID(2), "gesetzt-1"} {
		if !strings.Contains(zeilen[i], `"correlation_id":"`+id+`"`) {
			t.Errorf("Zeile %d: %q, erwartet correlation_id %q", i, zeilen[i], id)
		}
	}
}

// A06: jede Verformung jeder Art ist ErrUngueltig; nichts geschrieben, weder Uhr noch Generator
// gerufen. Je Art 15 + 2 × Länge Formen: leer, Groß- und Titelschreibung, Leerraum davor und
// dahinter (je 5), jedes echte Präfix und jeder echte Rest hinter dem Anfang, um ein Zeichen
// verlängert (3), mit einem Zeichen davor. Jede Form läuft als Eintrag mit den Feldern der Art
// und als Eintrag ohne Felder.
func verformungen(art string) []string {
	var out []string
	out = append(out, "", strings.ToUpper(art), strings.ToUpper(art[:1])+art[1:])
	for _, w := range []string{" ", "\t", "\n", "\r", "\v"} {
		out = append(out, w+art, art+w)
	}
	for k := 1; k < len(art); k++ {
		out = append(out, art[:k], art[k:])
	}
	return append(out, art+"x", art+"_", art+"0", "x"+art)
}

func TestArtAusserhalbDerMenge(t *testing.T) {
	gesamt := 0
	for art := range orakelPflicht() {
		formen := verformungen(string(art))
		einmalig := map[string]bool{}
		for _, f := range formen {
			einmalig[f] = true
		}
		if len(formen) != 15+2*len(art) || len(einmalig) != len(formen) {
			t.Fatalf("Art %s: %d Formen (%d verschieden), erwartet %d", art, len(formen), len(einmalig), 15+2*len(art))
		}
		for _, f := range formen {
			if _, bekannt := orakelPflicht()[support.AuditArt(f)]; bekannt {
				t.Fatalf("Verformung %q gleicht einer Art", f)
			}
			mitFeldern := gueltig(art)
			mitFeldern.Art = support.AuditArt(f)
			ohneFelder := support.Eintrag{Art: support.AuditArt(f)}
			for name, e := range map[string]support.Eintrag{"mit den Feldern der Art": mitFeldern, "ohne Felder": ohneFelder} {
				a := neu(t, festeZeit)
				if err := a.d.Append(e); !errors.Is(err, model.ErrUngueltig) {
					t.Errorf("Art %q (von %s, %s): %v, erwartet ErrUngueltig", f, art, name, err)
				}
				a.nichtsGerufen(t, fmt.Sprintf("Art %q (%s)", f, name))
				gesamt++
			}
		}
	}
	if want := 2 * (4*15 + 2*(17+20+19+21)); gesamt != want {
		t.Errorf("%d Aufrufe geprüft, erwartet %d", gesamt, want)
	}
}

// A07: 4 Arten × 5 Felder = 20 Zellen, je Zelle ein Fall je Richtung: ein fehlendes Pflichtfeld
// und ein gesetztes verbotenes Feld sind ErrUngueltig, die erlaubte Belegung schreibt.
func TestFeldBelegung(t *testing.T) {
	zellen := 0
	for art, soll := range orakelPflicht() {
		for i, f := range orakelFelder() {
			zellen++
			a := neu(t, festeZeit)
			erlaubt := gueltig(art)
			if err := a.d.Append(erlaubt); err != nil {
				t.Errorf("%s: erlaubte Belegung abgewiesen: %v", art, err)
			}
			verstoss := gueltig(art)
			setze(&verstoss, f, !soll[i])
			zuvor := a.inhalt(t)
			if err := a.d.Append(verstoss); !errors.Is(err, model.ErrUngueltig) {
				was := "gesetztes verbotenes Feld"
				if soll[i] {
					was = "fehlendes Pflichtfeld"
				}
				t.Errorf("%s %s (%s): %v, erwartet ErrUngueltig", art, f, was, err)
			}
			if a.inhalt(t) != zuvor {
				t.Errorf("%s %s: der Verstoß hat geschrieben", art, f)
			}
		}
	}
	if zellen != 20 {
		t.Errorf("%d Zellen, erwartet 20", zellen)
	}
}

// kennungsFelder sind die fünf Felder der Kennungsform mit der Art, die sie trägt.
func kennungsFelder() []struct {
	name string
	art  support.AuditArt
	key  string
	set  func(*support.Eintrag, string)
} {
	return []struct {
		name string
		art  support.AuditArt
		key  string
		set  func(*support.Eintrag, string)
	}{
		{"service_id", support.ArtSessionGeloescht, "service_id", func(e *support.Eintrag, v string) { e.ServiceID = v }},
		{"session_id", support.ArtSessionGeloescht, "session_id", func(e *support.Eintrag, v string) { e.SessionID = v }},
		{"workspace_id", support.ArtWorkspaceGeloescht, "workspace_id", func(e *support.Eintrag, v string) { e.WorkspaceID = v }},
		{"import_id", support.ArtImportFehlgeschlagen, "import_id", func(e *support.Eintrag, v string) { e.ImportID = v }},
		{"KorrelationsID", support.ArtSessionGeloescht, "correlation_id", func(e *support.Eintrag, v string) { e.KorrelationsID = v }},
	}
}

func inKlasse(b byte) bool {
	return strings.IndexByte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-", b) >= 0
}

// A08: 5 Felder × 256 Bytewerte an 3 Stellen (65 gültig je Stelle), dazu die Längen 0, 1, 64, 65
// und Mehrbyte-UTF-8; die Datei trägt genau die gültigen Zeilen.
func TestKennungsform(t *testing.T) {
	for _, f := range kennungsFelder() {
		a := neu(t, festeZeit)
		gueltigeZeilen := 0
		pruefe := func(wert string, soll bool, was string) {
			e := gueltig(f.art)
			f.set(&e, wert)
			err := a.d.Append(e)
			if soll && err != nil || !soll && !errors.Is(err, model.ErrUngueltig) {
				t.Errorf("%s %s = %q: %v, erwartet gültig=%v", f.name, was, wert, err, soll)
			}
			if soll && err == nil {
				gueltigeZeilen++
			}
		}
		for stelle, form := range []string{"%sxx", "x%sx", "xx%s"} {
			treffer := 0
			for b := range 256 {
				wert := fmt.Sprintf(form, string([]byte{byte(b)}))
				if inKlasse(byte(b)) {
					treffer++
				}
				pruefe(wert, inKlasse(byte(b)), fmt.Sprintf("Byte %#02x an Stelle %d", b, stelle))
			}
			if treffer != 65 {
				t.Fatalf("Klasse hat %d Bytes, erwartet 65", treffer)
			}
		}
		pruefe(strings.Repeat("a", 1), true, "Länge 1")
		pruefe(strings.Repeat("a", 64), true, "Länge 64")
		pruefe(strings.Repeat("a", 65), false, "Länge 65")
		pruefe(strings.Repeat("a", 64)+"\n", false, "Länge 64 und LF")
		pruefe("\x00", false, "NUL")
		pruefe("a\"b", false, "Anführungszeichen")
		pruefe("a\\b", false, "Backslash")
		pruefe(string([]byte{'a', 0xc3, 0xa9}), false, "Mehrbyte-UTF-8")
		pruefe(string([]byte{0xe6, 0x97, 0xa5, 0xe6, 0x9c, 0xac}), false, "Mehrbyte-UTF-8 lang")
		if f.name != "KorrelationsID" {
			pruefe("", false, "Länge 0")
		}
		if got := strings.Count(a.inhalt(t), "\n"); got != gueltigeZeilen || gueltigeZeilen != 3*65+2 {
			t.Errorf("%s: %d Zeilen in der Datei, %d gültige Aufrufe, erwartet %d", f.name, got, gueltigeZeilen, 3*65+2)
		}
	}
}

// codeTexte ist die Spalte „Code“ der Tabelle in SPEC-109 Nr. 2 aus dem Wortlaut (36 Codes).
func codeTexte() []string {
	return strings.Fields(`ODR-E2001 ODR-E2101 ODR-E2401 ODR-E3001 ODR-E4001 ODR-E5001 ODR-E5002 ODR-E5003
ODR-E5004 ODR-E5005 ODR-E5006 ODR-E5007 ODR-E5008 ODR-E5009 ODR-E5010 ODR-E6001 ODR-E6002 ODR-E6003
ODR-E6004 ODR-E6005 ODR-E6006 ODR-E6101 ODR-E6102 ODR-E7001 ODR-E7002 ODR-E7003 ODR-E7004 ODR-E7005
ODR-E7006 ODR-E7007 ODR-E7008 ODR-E7009 ODR-E7010 ODR-E7011 ODR-E7012 ODR-E9001`)
}

// A09: jeder der 36 Codes wird als sein Text geschrieben; jeder ganzzahlige Wert von -5 bis 60,
// der keiner ist (0, negativ, nach dem letzten), ist ErrUngueltig.
func TestErrorCode(t *testing.T) {
	texte := codeTexte()
	alle := model.AlleCodes()
	if len(texte) != 36 || len(alle) != 36 {
		t.Fatalf("%d Texte und %d Codes, erwartet je 36", len(texte), len(alle))
	}
	a := neu(t, festeZeit)
	gueltigeWerte := map[model.Code]string{}
	for _, text := range texte {
		c, ok := model.ParseCode(text)
		if !ok {
			t.Fatalf("Orakel nennt einen Code, den das Modell nicht kennt: %q", text)
		}
		gueltigeWerte[c] = text
	}
	if len(gueltigeWerte) != 36 {
		t.Fatalf("%d verschiedene Codes im Orakel, erwartet 36", len(gueltigeWerte))
	}
	schreiber, verworfen := 0, 0
	for v := -5; v <= 60; v++ {
		e := gueltig(support.ArtImportFehlgeschlagen)
		e.ErrorCode = model.Code(v)
		text, soll := gueltigeWerte[model.Code(v)]
		err := a.d.Append(e)
		switch {
		case soll && err != nil:
			t.Errorf("Code %d (%s) abgewiesen: %v", v, text, err)
		case soll:
			schreiber++
		case !errors.Is(err, model.ErrUngueltig):
			t.Errorf("Wert %d: %v, erwartet ErrUngueltig", v, err)
		default:
			verworfen++
		}
	}
	if schreiber != 36 || verworfen != 66-36 {
		t.Errorf("%d gültige und %d verworfene Werte, erwartet 36 und 30", schreiber, verworfen)
	}
	zeilen := strings.Split(strings.TrimSuffix(a.inhalt(t), "\n"), "\n")
	if len(zeilen) != 36 {
		t.Fatalf("%d Zeilen, erwartet 36", len(zeilen))
	}
	for _, text := range texte {
		if !strings.Contains(a.inhalt(t), `"error_code":"`+text+`"}`+"\n") {
			t.Errorf("Datei trägt %s nicht", text)
		}
	}
}

// A10: die Nullzeit und jede Zeit außerhalb der Jahre 1 bis 9999 (auch erst nach der Umrechnung
// nach UTC) sind ErrUngueltig; die Uhr wurde einmal gerufen, der Generator nie, nichts geschrieben.
func TestZeitAusserhalb(t *testing.T) {
	plus2 := time.FixedZone("plus2", 2*60*60)
	minus2 := time.FixedZone("minus2", -2*60*60)
	faelle := map[string]time.Time{
		"Nullzeit":                        {},
		"Jahr 0":                          time.Date(0, 12, 31, 23, 59, 59, 999999999, time.UTC),
		"Jahr 10000":                      time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
		"Jahr 1 vor UTC, 0 danach":        time.Date(1, 1, 1, 1, 0, 0, 0, plus2),
		"Jahr 9999 vor UTC, 10000 danach": time.Date(9999, 12, 31, 23, 30, 0, 0, minus2),
	}
	for name, zeit := range faelle {
		a := neu(t, func(int) time.Time { return zeit })
		if err := a.d.Append(gueltig(support.ArtSessionGeloescht)); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v, erwartet ErrUngueltig", name, err)
		}
		if n := a.uhr.Aufrufe(); n != 1 {
			t.Errorf("%s: Uhr %d-mal gerufen, erwartet 1", name, n)
		}
		if n := a.ids.Gesamt(); n != 0 {
			t.Errorf("%s: Generator %d-mal gerufen, erwartet 0", name, n)
		}
		if !a.dateiFehlt() {
			t.Errorf("%s: Datei angelegt", name)
		}
	}
	if len(faelle) != 5 {
		t.Fatalf("%d Fälle, erwartet 5", len(faelle))
	}
}

// A11: die Datei entsteht mit 0600 (bei Umask 0 nicht weiter); eine vorhandene behält ihre Rechte.
func TestRechte(t *testing.T) {
	alt := syscall.Umask(0)
	t.Cleanup(func() { syscall.Umask(alt) })
	a := neu(t, festeZeit)
	if err := a.d.Append(gueltig(support.ArtSessionGeloescht)); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(a.pfad); err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("neue Datei: %v, %v, erwartet 0600", info, err)
	}
	b := neu(t, festeZeit)
	if err := os.WriteFile(b.pfad, []byte("alt\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(b.pfad, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := b.d.Append(gueltig(support.ArtSessionGeloescht)); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(b.pfad); err != nil || info.Mode().Perm() != 0o640 {
		t.Errorf("vorhandene Datei: %v, %v, erwartet 0640", info, err)
	}
}

// A12: der zweite Eintrag steht hinter dem ersten, ein neuer Adapter auf derselben Datei hängt
// an, Vorhandenes bleibt byte-gleich.
func TestAnhaengen(t *testing.T) {
	a := neu(t, festeZeit)
	if err := os.WriteFile(a.pfad, []byte("vorher\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	staende := []string{a.inhalt(t)}
	for i, art := range []support.AuditArt{support.ArtSessionGeloescht, support.ArtWorkspaceGeloescht} {
		if err := a.d.Append(gueltig(art)); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
		staende = append(staende, a.inhalt(t))
	}
	b := neuAn(t, a.pfad, festeZeit)
	if err := b.d.Append(gueltig(support.ArtImportFehlgeschlagen)); err != nil {
		t.Fatal(err)
	}
	staende = append(staende, a.inhalt(t))
	for i := 1; i < len(staende); i++ {
		if !strings.HasPrefix(staende[i], staende[i-1]) || len(staende[i]) <= len(staende[i-1]) {
			t.Errorf("Stand %d trägt Stand %d nicht als Anfang:\n%q\n%q", i, i-1, staende[i], staende[i-1])
		}
	}
	if len(staende) != 4 || !strings.HasPrefix(staende[3], "vorher\n") {
		t.Errorf("%d Stände, Anfang %q", len(staende), staende[3])
	}
}

// A13: eine abgebrochene letzte Zeile bekommt genau ein LF; leere Datei, Datei mit LF am Ende
// und fehlende Datei bekommen keines.
func TestZeilenende(t *testing.T) {
	faelle := []struct {
		name      string
		vorhanden bool
		vorher    string
		trenner   string
	}{
		{"keine Datei", false, "", ""},
		{"leere Datei", true, "", ""},
		{"endet mit LF", true, "alt\n", ""},
		{"nur LF", true, "\n", ""},
		{"endet ohne LF", true, "halb", "\n"},
		{"endet mit CR", true, "x\r", "\n"},
		{"zwei Zeilen, halbe dritte", true, "a\nb", "\n"},
	}
	for _, f := range faelle {
		a := neu(t, festeZeit)
		if f.vorhanden {
			if err := os.WriteFile(a.pfad, []byte(f.vorher), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		if err := a.d.Append(gueltig(support.ArtWorkspaceGeloescht)); err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		zeile := `{"timestamp":"2026-11-06T10:00:00.000000000Z","correlation_id":"` + corID(1) + `","art":"workspace_geloescht","workspace_id":"w-workspace_id"}` + "\n"
		if got, want := a.inhalt(t), f.vorher+f.trenner+zeile; got != want {
			t.Errorf("%s: %q, erwartet %q", f.name, got, want)
		}
	}
}

// A14: scheitert das Öffnen, kommt der Fehler des Betriebssystems zurück und nichts ist
// geschrieben; nach dem Anlegen des Verzeichnisses gelingt das nächste Append mit genau einer Zeile.
func TestOeffnenScheitert(t *testing.T) {
	verz := t.TempDir()
	v := neuAn(t, verz, festeZeit)
	err := v.d.Append(gueltig(support.ArtSessionGeloescht))
	if !errors.Is(err, syscall.EISDIR) || errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Pfad ist ein Verzeichnis: %v, erwartet EISDIR", err)
	}
	if es, _ := os.ReadDir(verz); len(es) != 0 {
		t.Errorf("im Verzeichnis liegt etwas: %v", es)
	}
	wurzel := t.TempDir()
	a := neuAn(t, filepath.Join(wurzel, "fehlt", audit.DateiName), festeZeit)
	err = a.d.Append(gueltig(support.ArtSessionGeloescht))
	if !errors.Is(err, fs.ErrNotExist) || errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Elternverzeichnis fehlt: %v, erwartet ErrNotExist", err)
	}
	if _, err := os.Stat(filepath.Join(wurzel, "fehlt")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("das Verzeichnis wurde angelegt: %v", err)
	}
	if err := os.Mkdir(filepath.Join(wurzel, "fehlt"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := a.d.Append(gueltig(support.ArtSessionGeloescht)); err != nil {
		t.Fatalf("Append nach dem Anlegen: %v", err)
	}
	if n := strings.Count(a.inhalt(t), "\n"); n != 1 {
		t.Errorf("%d Zeilen, erwartet 1", n)
	}
}

// A15: scheitert das Schreiben (/dev/full), kommt ENOSPC zurück, und Append kehrt zurück.
func TestSchreibenScheitert(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Fatalf("/dev/full fehlt: %v", err)
	}
	a := neuAn(t, "/dev/full", festeZeit)
	err := mitFrist(t, "Append auf /dev/full", func() error { return a.d.Append(gueltig(support.ArtSessionGeloescht)) })
	if !errors.Is(err, syscall.ENOSPC) || errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Append auf /dev/full: %v, erwartet ENOSPC", err)
	}
}

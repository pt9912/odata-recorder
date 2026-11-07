//go:build unix

package admin_test

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die sieben Gründe und das Muster des Textes als Literale des Vertrags (SPEC-239 Nr. 5, 6), nicht aus dem Code.
const (
	gNichtLesbar   = "Datei nicht lesbar"
	gVerzeichnis   = "ist ein Verzeichnis"
	gKeineDatei    = "keine reguläre Datei"
	gRechte        = "Gruppen- oder Weltrechte"
	gMehrZeilen    = "mehr als eine Zeile"
	gZuKurz        = "weniger als 32 Zeichen"
	gNichtAnlegbar = "Datei nicht anlegbar"
	textKopf       = "ODR-E2001 ADMIN_TOKEN_FILE: "
)

func wdh(s string, n int) string { return strings.Repeat(s, n) }

func hex64() string { return wdh("0123456789abcdef", 4) }

func schreibe(t *testing.T, pfad, inhalt string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(pfad, mode); err != nil {
		t.Fatal(err)
	}
}

func erwarteGrund(t *testing.T, name string, err error, grund string) {
	t.Helper()
	var tf *admin.TokenFehler
	if !errors.As(err, &tf) {
		t.Errorf("%s: Fehler = %v, erwartet *TokenFehler %q", name, err, grund)
		return
	}
	if tf.Grund != grund {
		t.Errorf("%s: Grund = %q, erwartet %q", name, tf.Grund, grund)
	}
	if want := textKopf + grund; err.Error() != want {
		t.Errorf("%s: Text = %q, erwartet %q", name, err.Error(), want)
	}
}

type stand struct {
	modus  fs.FileMode
	inhalt string
	zeit   int64
}

func schnappschuss(t *testing.T, pfad string) stand {
	t.Helper()
	fi, err := os.Lstat(pfad)
	if err != nil {
		t.Fatal(err)
	}
	s := stand{modus: fi.Mode(), zeit: fi.ModTime().UnixNano()}
	if fi.Mode().IsRegular() {
		b, err := os.ReadFile(pfad)
		if err != nil {
			t.Fatal(err)
		}
		s.inhalt = string(b)
	}
	return s
}

func namen(t *testing.T, dir string) []string {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var n []string
	for _, e := range es {
		n = append(n, e.Name())
	}
	sort.Strings(n)
	return n
}

func gleich(a, b []string) bool { return strings.Join(a, "\x00") == strings.Join(b, "\x00") }

// dateiFall ist eine Datei mit Inhalt und Modus; grund leer heißt gültig mit dem Ergebnis erwartet.
type dateiFall struct {
	name, inhalt string
	modus        os.FileMode
	grund        string
	erwartet     string
}

// pruefeFall legt die Datei an, ruft LadeToken und prüft Ergebnis und Unverändertheit der Datei.
func pruefeFall(t *testing.T, f dateiFall) {
	t.Helper()
	pfad := filepath.Join(t.TempDir(), "admin.token")
	schreibe(t, pfad, f.inhalt, f.modus)
	vorher := schnappschuss(t, pfad)
	got, err := admin.LadeToken(pfad)
	if f.grund == "" {
		if err != nil {
			t.Errorf("%s: Fehler %v bei gültiger Datei", f.name, err)
		} else if string(got) != f.erwartet {
			t.Errorf("%s: Token = %q, erwartet %q", f.name, got, f.erwartet)
		}
	} else {
		erwarteGrund(t, f.name, err, f.grund)
		if got != nil {
			t.Errorf("%s: Token %q trotz Verstoß", f.name, got)
		}
	}
	if nachher := schnappschuss(t, pfad); nachher != vorher {
		t.Errorf("%s: Datei verändert: %+v -> %+v", f.name, vorher, nachher)
	}
}

// T1 (M-Token-Inhalt): Kern 5 x Rand links 8 x Rand rechts 8 gleich 320; der Kern ist das Literal je Fall.
func TestInhaltKernUndRand(t *testing.T) {
	nbsp := "\xc2\xa0"
	kerne := []string{
		"abcdefghijklmnopqrstuvwxyz012345",
		hex64(),
		"0123456789abcdef 0123456789abcdef",
		nbsp + wdh("\xc3\xa4", 14) + nbsp,
		wdh("ab", 50),
	}
	raender := []string{"", "\n", "\r\n", " ", "\t", "\v", "\f", "\r"}
	n := 0
	for ki, kern := range kerne {
		for li, l := range raender {
			for ri, r := range raender {
				pruefeFall(t, dateiFall{name: "kern" + strconv.Itoa(ki) + "/links" + strconv.Itoa(li) + "/rechts" + strconv.Itoa(ri),
					inhalt: l + kern + r, modus: 0o600, erwartet: kern})
				n++
			}
		}
	}
	if len(kerne) != 5 || len(raender) != 8 || n != 320 {
		t.Fatalf("Menge: %d Kerne, %d Ränder, %d Elemente, erwartet 5, 8, 320", len(kerne), len(raender), n)
	}
}

func rechteFaelle() []dateiFall {
	var fs []dateiFall
	for _, m := range []os.FileMode{0o400, 0o500, 0o600, 0o700} {
		fs = append(fs, dateiFall{name: "gueltig" + strconv.FormatInt(int64(m), 8), inhalt: hex64() + "\n", modus: m, erwartet: hex64()})
	}
	for _, bit := range []os.FileMode{0o040, 0o020, 0o010, 0o004, 0o002, 0o001} {
		fs = append(fs, dateiFall{name: "bit" + strconv.FormatInt(int64(bit), 8), inhalt: hex64() + "\n", modus: 0o600 | bit, grund: gRechte})
	}
	for _, m := range []os.FileMode{0o644, 0o660, 0o777} {
		fs = append(fs, dateiFall{name: "kombi" + strconv.FormatInt(int64(m), 8), inhalt: hex64() + "\n", modus: m, grund: gRechte})
	}
	return fs
}

// T2 (M-Rechte): 4 gültige, 6 einzelne Bits, 3 Kombinationen gleich 13.
func TestRechteMenge(t *testing.T) {
	fs := rechteFaelle()
	if len(fs) != 13 {
		t.Fatalf("Menge M-Rechte = %d Elemente, erwartet 13", len(fs))
	}
	for _, f := range fs {
		pruefeFall(t, f)
	}
}

// T4: es zählen die Rechte der aufgelösten Datei, nicht die des Symlinks.
func TestSymlinkRechteDerAufgeloestenDatei(t *testing.T) {
	dir := t.TempDir()
	offen := filepath.Join(dir, "offen")
	zu := filepath.Join(dir, "zu")
	schreibe(t, offen, hex64(), 0o644)
	schreibe(t, zu, hex64(), 0o400)
	for _, c := range []struct {
		name, ziel, grund string
	}{{"offen", offen, gRechte}, {"zu", zu, ""}} {
		l1 := filepath.Join(dir, c.name+"-l1")
		l2 := filepath.Join(dir, c.name+"-l2")
		if err := os.Symlink(c.ziel, l1); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(l1, l2); err != nil {
			t.Fatal(err)
		}
		for _, p := range []string{l1, l2} {
			got, err := admin.LadeToken(p)
			if c.grund == "" {
				if err != nil || string(got) != hex64() {
					t.Errorf("%s: %q, %v; erwartet gültig", p, got, err)
				}
				continue
			}
			erwarteGrund(t, p, err, c.grund)
		}
	}
}

func kurzFaelle() []dateiFall {
	a := func(n int) string { return wdh("a", n) }
	return []dateiFall{
		{name: "31", inhalt: a(31), grund: gZuKurz},
		{name: "32", inhalt: a(32), erwartet: a(32)},
		{name: "33", inhalt: a(33), erwartet: a(33)},
		{name: "31+nl", inhalt: a(31) + "\n", grund: gZuKurz},
		{name: "30+Rand", inhalt: " " + a(30) + " ", grund: gZuKurz},
		{name: "31+innen", inhalt: a(15) + " " + a(16), erwartet: a(15) + " " + a(16)},
		{name: "leer", inhalt: "", grund: gZuKurz},
		{name: "nl", inhalt: "\n", grund: gZuKurz},
		{name: "31 Leerzeichen", inhalt: wdh(" ", 31), grund: gZuKurz},
		{name: "40 Leerzeichen", inhalt: wdh(" ", 40), grund: gZuKurz},
		{name: "40 Tabs", inhalt: wdh("\t", 40), grund: gZuKurz},
		{name: "crlf crlf", inhalt: "\r\n\r\n", grund: gZuKurz},
	}
}

// T5: Grenze bei 32 Bytes nach dem Entfernen des Leerraums; leere Datei und nur Leerraum nie gültig.
func TestLaengeNachEntfernen(t *testing.T) {
	fs := kurzFaelle()
	if len(fs) != 12 {
		t.Fatalf("Menge = %d Elemente, erwartet 12", len(fs))
	}
	for _, f := range fs {
		f.modus = 0o600
		pruefeFall(t, f)
	}
}

func zeilenFaelle() []dateiFall {
	a, b := wdh("a", 40), wdh("b", 40)
	return []dateiFall{
		{name: "lf", inhalt: a + "\n" + b, grund: gMehrZeilen},
		{name: "crlf", inhalt: a + "\r\n" + b, grund: gMehrZeilen},
		{name: "cr", inhalt: a + "\r" + b, grund: gMehrZeilen},
		{name: "lflf", inhalt: a + "\n\n" + b, grund: gMehrZeilen},
		{name: "nachlauf", inhalt: a + "\n\n", erwartet: a},
		{name: "vorlauf", inhalt: "\n" + a, erwartet: a},
		{name: "crlf-ende", inhalt: a + "\r\n", erwartet: a},
		{name: "rand", inhalt: " " + a + "\n", erwartet: a},
	}
}

// T7: ein \n oder \r im Rest ist ein Verstoß; abschließender Umbruch und umgebender Leerraum nicht.
func TestMehrereZeilen(t *testing.T) {
	for _, f := range zeilenFaelle() {
		f.modus = 0o600
		pruefeFall(t, f)
	}
}

// T8: ein Verzeichnis, auch als Ziel eines Symlinks.
func TestVerzeichnis(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(sub, link); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{sub, link} {
		_, err := admin.LadeToken(p)
		erwarteGrund(t, p, err, gVerzeichnis)
	}
}

type art struct{ name, pfad string }

// nichtRegulaer legt die fünf Formen von M-Nicht-regulär an; der Socket lebt bis zum Ende des Tests.
func nichtRegulaer(t *testing.T) []art {
	t.Helper()
	dir := t.TempDir()
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(fifo, 0o666); err != nil {
		t.Fatal(err)
	}
	null := filepath.Join(dir, "null-link")
	if err := os.Symlink("/dev/null", null); err != nil {
		t.Fatal(err)
	}
	ohneZiel := filepath.Join(dir, "ohne-ziel")
	if err := os.Symlink(filepath.Join(dir, "gibt-es-nicht"), ohneZiel); err != nil {
		t.Fatal(err)
	}
	sock := filepath.Join(dir, "sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return []art{{"fifo", fifo}, {"/dev/null", "/dev/null"}, {"Symlink auf /dev/null", null}, {"Symlink ohne Ziel", ohneZiel}, {"Socket", sock}}
}

// mitFrist ruft f in einer Goroutine; ein Hängen über die Frist ist rot.
func mitFrist(t *testing.T, name string, f func() ([]byte, error)) ([]byte, error) {
	t.Helper()
	type erg struct {
		b   []byte
		err error
	}
	c := make(chan erg, 1)
	go func() {
		b, err := f()
		c <- erg{b, err}
	}()
	select {
	case r := <-c:
		return r.b, r.err
	case <-time.After(10 * time.Second):
		t.Fatalf("%s: LadeToken blockiert", name)
		return nil, nil
	}
}

// T9 (M-Nicht-regulär): fünf Formen, ohne zu blockieren; das Orakel ist das Literal.
func TestNichtRegulaer(t *testing.T) {
	fs := nichtRegulaer(t)
	if len(fs) != 5 {
		t.Fatalf("Menge = %d Elemente, erwartet 5", len(fs))
	}
	for _, f := range fs {
		vorher := schnappschuss(t, f.pfad)
		_, err := mitFrist(t, f.name, func() ([]byte, error) { return admin.LadeToken(f.pfad) })
		erwarteGrund(t, f.name, err, gKeineDatei)
		if nachher := schnappschuss(t, f.pfad); nachher != vorher {
			t.Errorf("%s: verändert", f.name)
		}
	}
	// Der Symlink ohne Ziel bleibt, es entstand nichts daneben.
	dir := filepath.Dir(fs[3].pfad)
	if got, want := namen(t, dir), []string{"fifo", "null-link", "ohne-ziel", "sock"}; !gleich(got, want) {
		t.Errorf("Verzeichnis = %v, erwartet %v", got, want)
	}
}

// T10: ein Fehler von Stat außer „fehlt“ ist Datei nicht lesbar und erzeugt nichts.
func TestStatFehlerErzeugtNichts(t *testing.T) {
	dir := t.TempDir()
	schreibe(t, filepath.Join(dir, "datei"), hex64(), 0o600)
	if err := os.Symlink(filepath.Join(dir, "b"), filepath.Join(dir, "a")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "a"), filepath.Join(dir, "b")); err != nil {
		t.Fatal(err)
	}
	vorher := namen(t, dir)
	for _, c := range []struct{ name, pfad string }{
		{"Pfadelement ist Datei", filepath.Join(dir, "datei", "x")},
		{"Symlinkschleife", filepath.Join(dir, "a")},
		{"NUL im Pfad", filepath.Join(dir, "x") + "\x00y"},
	} {
		_, err := admin.LadeToken(c.pfad)
		erwarteGrund(t, c.name, err, gNichtLesbar)
		if nachher := namen(t, dir); !gleich(nachher, vorher) {
			t.Errorf("%s: Verzeichnis %v -> %v", c.name, vorher, nachher)
		}
	}
}

// T11 (Zusatztabelle): bei mehreren Verstößen gilt der kleinste Rang.
func TestRangfolge(t *testing.T) {
	a, b := wdh("a", 40), wdh("b", 40)
	dir := t.TempDir()
	sub := filepath.Join(dir, "verz")
	if err := os.Mkdir(sub, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(sub, 0o777); err != nil {
		t.Fatal(err)
	}
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(fifo, 0o666); err != nil {
		t.Fatal(err)
	}
	for _, f := range []dateiFall{
		{name: "Rechte+zwei Zeilen", inhalt: a + "\n" + b, modus: 0o644, grund: gRechte},
		{name: "Rechte+zu kurz", inhalt: "ab", modus: 0o644, grund: gRechte},
		{name: "zwei Zeilen zusammen zu kurz", inhalt: "ab\ncd", modus: 0o600, grund: gMehrZeilen},
		{name: "Rechte+zwei Zeilen+zu kurz", inhalt: "ab\ncd", modus: 0o644, grund: gRechte},
	} {
		pruefeFall(t, f)
	}
	_, err := admin.LadeToken(sub)
	erwarteGrund(t, "Verzeichnis 0777", err, gVerzeichnis)
	_, err = mitFrist(t, "FIFO 0666", func() ([]byte, error) { return admin.LadeToken(fifo) })
	erwarteGrund(t, "FIFO 0666", err, gKeineDatei)
}

// T12: ein leerer Pfad ist model.ErrUngueltig und kein TokenFehler; ein relativer gilt zum Arbeitsverzeichnis.
func TestPfadLeerUndRelativ(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	_, err := admin.LadeToken("")
	var tf *admin.TokenFehler
	if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &tf) {
		t.Errorf("leerer Pfad: %v, erwartet model.ErrUngueltig ohne TokenFehler", err)
	}
	if n := namen(t, dir); len(n) != 0 {
		t.Errorf("leerer Pfad erzeugte %v", n)
	}
	if err := os.Mkdir("rel", 0o700); err != nil {
		t.Fatal(err)
	}
	schreibe(t, "rel/x", hex64()+"\n", 0o600)
	if got, err := admin.LadeToken("rel/x"); err != nil || string(got) != hex64() {
		t.Errorf("relativ vorhanden: %q, %v", got, err)
	}
	got, err := admin.LadeToken("rel/neu")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rel", "neu"))
	if err != nil || string(b) != string(got)+"\n" {
		t.Errorf("relativ erzeugt: Datei %q (%v), Rückgabe %q", b, err, got)
	}
}

// T13: Symlink und Ketten von Symlinks auf eine reguläre Datei sind gültig; die Rückgabe ist der Inhalt des Ziels.
func TestSymlinkKetteGueltig(t *testing.T) {
	dir := t.TempDir()
	ziel := filepath.Join(dir, "ziel")
	schreibe(t, ziel, hex64()+"\n", 0o400)
	vorher := ziel
	for i := 1; i <= 3; i++ {
		l := filepath.Join(dir, "l"+strconv.Itoa(i))
		if err := os.Symlink(vorher, l); err != nil {
			t.Fatal(err)
		}
		vorher = l
		if got, err := admin.LadeToken(l); err != nil || string(got) != hex64() {
			t.Errorf("Kette %d: %q, %v", i, got, err)
		}
	}
}

// T14: erzeugtes Token 64 Zeichen 0-9a-f, Datei Token plus Zeilenende, 100 verschiedene Werte; das Ergebnis ist eine Kopie.
func TestErzeugtesToken(t *testing.T) {
	muster := regexp.MustCompile(`^[0-9a-f]{64}$`)
	gesehen := map[string]bool{}
	for i := 0; i < 100; i++ {
		pfad := filepath.Join(t.TempDir(), "admin.token")
		got, err := admin.LadeToken(pfad)
		if err != nil {
			t.Fatal(err)
		}
		if !muster.Match(got) {
			t.Fatalf("Token %q passt nicht auf %s", got, muster)
		}
		b, err := os.ReadFile(pfad)
		if err != nil || string(b) != string(got)+"\n" {
			t.Fatalf("Datei %q (%v), erwartet Token plus Zeilenende", b, err)
		}
		if cap(got) >= 512 {
			t.Fatalf("Ergebnis ist kein eigener Puffer: cap = %d", cap(got))
		}
		gesehen[string(got)] = true
	}
	if len(gesehen) != 100 {
		t.Errorf("%d verschiedene Token in 100 Läufen", len(gesehen))
	}
	// Eine vorhandene Datei liefert ebenfalls eine Kopie des Rests, nicht den Lesepuffer.
	pfad := filepath.Join(t.TempDir(), "admin.token")
	schreibe(t, pfad, hex64()+"\n", 0o600)
	got, err := admin.LadeToken(pfad)
	if err != nil || cap(got) >= 512 {
		t.Errorf("vorhandene Datei: cap = %d, %v", cap(got), err)
	}
}

// T15: die Zufallsquelle ist crypto/rand.
func TestZufallsquelle(t *testing.T) {
	es, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	importe := map[string]bool{}
	dateien := 0
	for _, e := range es {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		dateien++
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			p, _ := strconv.Unquote(im.Path.Value)
			importe[p] = true
		}
	}
	if dateien == 0 {
		t.Fatal("keine Quelldatei gefunden")
	}
	if !importe["crypto/rand"] {
		t.Error("crypto/rand fehlt")
	}
	for _, v := range []string{"math/rand", "math/rand/v2"} {
		if importe[v] {
			t.Errorf("%s importiert", v)
		}
	}
}

// T16 (M-Umask): bei jeder Umask trägt die Hilfsdatei von Anfang an keine Rechte von Gruppe oder Welt, die Datei
// am Ende genau 0600. Die Umask gilt für den Prozess: kein Test des Pakets läuft parallel.
func TestRechteBeiJederUmask(t *testing.T) {
	umasks := []int{0o000, 0o022, 0o077, 0o277}
	if len(umasks) != 4 {
		t.Fatal("Menge M-Umask")
	}
	for _, um := range umasks {
		dir := t.TempDir()
		pfad := filepath.Join(dir, "admin.token")
		var haltRechte os.FileMode
		var haltGroesse int64 = -1
		func() {
			alt := syscall.Umask(um)
			defer syscall.Umask(alt)
			_, err := admin.LadeMitHaltepunkten(pfad, func(h string) {
				if fi, err := os.Stat(h); err == nil {
					haltRechte, haltGroesse = fi.Mode().Perm(), fi.Size()
				}
			}, nil)
			if err != nil {
				t.Errorf("umask %04o: %v", um, err)
			}
		}()
		if haltGroesse != 0 || haltRechte&0o077 != 0 {
			t.Errorf("umask %04o: Hilfsdatei nach dem Anlegen: Rechte %04o, Größe %d", um, haltRechte, haltGroesse)
		}
		fi, err := os.Stat(pfad)
		if err != nil || fi.Mode().Perm() != 0o600 {
			t.Errorf("umask %04o: Datei %v, %v; erwartet 0600", um, fi, err)
		}
	}
}

// T17: die Datei erscheint vollständig oder gar nicht.
func TestVollstaendigOderGarNicht(t *testing.T) {
	dir := t.TempDir()
	pfad := filepath.Join(dir, "admin.token")
	hilfsGroesse := int64(-1)
	_, lerr := os.Lstat(pfad)
	fehlte := false
	got, err := admin.LadeMitHaltepunkten(pfad, nil, func(h string) {
		_, lerr = os.Lstat(pfad)
		fehlte = errors.Is(lerr, fs.ErrNotExist)
		if fi, err := os.Stat(h); err == nil {
			hilfsGroesse = fi.Size()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !fehlte {
		t.Error("pfad existierte vor dem Link")
	}
	if hilfsGroesse != 65 {
		t.Errorf("Hilfsdatei vor dem Link: %d Bytes, erwartet 65", hilfsGroesse)
	}
	if fi, err := os.Stat(pfad); err != nil || fi.Size() != 65 {
		t.Errorf("pfad nach der Rückkehr: %v, %v; erwartet 65 Bytes", fi, err)
	}
	if len(got) != 64 {
		t.Errorf("Token %q", got)
	}
}

// T18: ein Aufruf, dem ein anderer zwischen „fehlt“ und Link zuvorkommt, liefert dessen Token.
func TestWettlaufDeterministisch(t *testing.T) {
	dir := t.TempDir()
	pfad := filepath.Join(dir, "admin.token")
	var innen []byte
	var innenErr error
	aussen, err := admin.LadeMitHaltepunkten(pfad, nil, func(string) { innen, innenErr = admin.LadeToken(pfad) })
	if err != nil || innenErr != nil {
		t.Fatalf("außen %v, innen %v", err, innenErr)
	}
	if len(innen) != 64 || !bytes.Equal(aussen, innen) {
		t.Errorf("außen %q, innen %q; erwartet gleich", aussen, innen)
	}
	if b, err := os.ReadFile(pfad); err != nil || string(b) != string(innen)+"\n" {
		t.Errorf("Datei %q (%v), erwartet das Token des Siegers", b, err)
	}
	if n := namen(t, dir); !gleich(n, []string{"admin.token"}) {
		t.Errorf("Verzeichnis = %v, erwartet allein admin.token", n)
	}
}

// T19: 16 Goroutinen in 20 Läufen erzeugen genau eine Datei, alle liefern deren Inhalt.
func TestWettlaufNebenlaeufig(t *testing.T) {
	for lauf := 0; lauf < 20; lauf++ {
		dir := t.TempDir()
		pfad := filepath.Join(dir, "admin.token")
		start := make(chan struct{})
		var wg sync.WaitGroup
		res := make([][]byte, 16)
		errs := make([]error, 16)
		for i := range res {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				res[i], errs[i] = admin.LadeToken(pfad)
			}()
		}
		close(start)
		wg.Wait()
		b, err := os.ReadFile(pfad)
		if err != nil {
			t.Fatal(err)
		}
		for i := range res {
			if errs[i] != nil || len(res[i]) != 64 || string(b) != string(res[i])+"\n" {
				t.Fatalf("Lauf %d, Goroutine %d: %q, %v; Datei %q", lauf, i, res[i], errs[i], b)
			}
		}
		if n := namen(t, dir); !gleich(n, []string{"admin.token"}) {
			t.Fatalf("Lauf %d: Verzeichnis = %v", lauf, n)
		}
		if fi, err := os.Stat(pfad); err != nil || fi.Mode().Perm() != 0o600 {
			t.Fatalf("Lauf %d: %v, %v; erwartet 0600", lauf, fi, err)
		}
	}
}

// T20: die Hilfsdatei ist auf jedem Rückkehrpfad entfernt.
func TestHilfsdateiEntfernt(t *testing.T) {
	// nach dem Erzeugen
	dir := t.TempDir()
	if _, err := admin.LadeToken(filepath.Join(dir, "admin.token")); err != nil {
		t.Fatal(err)
	}
	if n := namen(t, dir); !gleich(n, []string{"admin.token"}) {
		t.Errorf("nach Erzeugen: %v", n)
	}
	// nach verlorenem Link
	dir = t.TempDir()
	pfad := filepath.Join(dir, "admin.token")
	if _, err := admin.LadeMitHaltepunkten(pfad, nil, func(string) { _, _ = admin.LadeToken(pfad) }); err != nil {
		t.Fatal(err)
	}
	if n := namen(t, dir); !gleich(n, []string{"admin.token"}) {
		t.Errorf("nach verlorenem Link: %v", n)
	}
	// nach verlorenem Link, wenn an der Stelle ein Symlink ohne Ziel liegt: der Verstoß der Vorhandenen gilt
	dir = t.TempDir()
	pfad = filepath.Join(dir, "admin.token")
	_, err := admin.LadeMitHaltepunkten(pfad, nil, func(string) {
		if err := os.Symlink(filepath.Join(dir, "gibt-es-nicht"), pfad); err != nil {
			t.Error(err)
		}
	})
	erwarteGrund(t, "Symlink ohne Ziel nach Wettlauf", err, gKeineDatei)
	if n := namen(t, dir); !gleich(n, []string{"admin.token"}) {
		t.Errorf("nach Wettlauf mit Symlink: %v", n)
	}
	// nach Datei nicht anlegbar
	dir = t.TempDir()
	_, err = admin.LadeToken(filepath.Join(dir, "fehlt", "admin.token"))
	erwarteGrund(t, "Eltern fehlt", err, gNichtAnlegbar)
	if n := namen(t, dir); len(n) != 0 {
		t.Errorf("nach nicht anlegbar: %v", n)
	}
}

// T21: fehlt das Verzeichnis, ist es Datei nicht anlegbar; es entsteht weder Datei noch Verzeichnis.
func TestVerzeichnisFehlt(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink(filepath.Join(dir, "gibt-es-nicht"), filepath.Join(dir, "ohne-ziel")); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		filepath.Join(dir, "a", "b", "admin.token"),
		filepath.Join(dir, "ohne-ziel", "admin.token"),
	} {
		_, err := admin.LadeToken(p)
		erwarteGrund(t, p, err, gNichtAnlegbar)
		if _, err := os.Lstat(p); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s: entstand: %v", p, err)
		}
	}
	if n := namen(t, dir); !gleich(n, []string{"ohne-ziel"}) {
		t.Errorf("Verzeichnis = %v, erwartet allein ohne-ziel", n)
	}
}

// T22: ein zweiter Aufruf liefert dasselbe Token und lässt die Datei unverändert (auch die Änderungszeit).
func TestZweiterAufruf(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "admin.token")
	erst, err := admin.LadeToken(pfad)
	if err != nil {
		t.Fatal(err)
	}
	alt := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(pfad, alt, alt); err != nil {
		t.Fatal(err)
	}
	fi1, _ := os.Stat(pfad)
	vorher := schnappschuss(t, pfad)
	zweit, err := admin.LadeToken(pfad)
	if err != nil || !bytes.Equal(erst, zweit) {
		t.Fatalf("zweiter Aufruf: %q, %v; erwartet %q", zweit, err, erst)
	}
	fi2, _ := os.Stat(pfad)
	if nachher := schnappschuss(t, pfad); nachher != vorher || !os.SameFile(fi1, fi2) {
		t.Errorf("Datei verändert: %+v -> %+v", vorher, nachher)
	}
}

// verstoss ist ein Pfad mit dem Grund, den LadeToken meldet.
type verstoss struct{ name, pfad, grund string }

// verstoesse legt alle Verstöße von T2, T5, T7, T8 und T9 an (Rechte 9, zu kurz 9, Zeilen 4, Verzeichnis 2, nicht regulär 5).
func verstoesse(t *testing.T) []verstoss {
	t.Helper()
	var vs []verstoss
	datei := func(f dateiFall) {
		pfad := filepath.Join(t.TempDir(), "admin.token")
		schreibe(t, pfad, f.inhalt, f.modus)
		vs = append(vs, verstoss{f.name, pfad, f.grund})
	}
	for _, f := range rechteFaelle() {
		if f.grund != "" {
			datei(f)
		}
	}
	for _, f := range kurzFaelle() {
		f.modus = 0o600
		if f.grund != "" {
			datei(f)
		}
	}
	for _, f := range zeilenFaelle() {
		f.modus = 0o600
		if f.grund != "" {
			datei(f)
		}
	}
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	link := filepath.Join(dir, "link")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(sub, link); err != nil {
		t.Fatal(err)
	}
	vs = append(vs, verstoss{"Verzeichnis", sub, gVerzeichnis}, verstoss{"Symlink auf Verzeichnis", link, gVerzeichnis})
	for _, a := range nichtRegulaer(t) {
		vs = append(vs, verstoss{a.name, a.pfad, gKeineDatei})
	}
	return vs
}

// T23: ein Aufruf ändert, kürzt, löscht und korrigiert eine vorhandene Datei nie, auch bei einem Verstoß nicht.
func TestVorhandeneDateiBleibtUnveraendert(t *testing.T) {
	vs := verstoesse(t)
	if len(vs) != 9+9+4+2+5 {
		t.Fatalf("Menge = %d Verstöße, erwartet 29", len(vs))
	}
	for _, v := range vs {
		vorher := schnappschuss(t, v.pfad)
		_, err := mitFrist(t, v.name, func() ([]byte, error) { return admin.LadeToken(v.pfad) })
		erwarteGrund(t, v.name, err, v.grund)
		if nachher := schnappschuss(t, v.pfad); nachher != vorher {
			t.Errorf("%s: %+v -> %+v", v.name, vorher, nachher)
		}
		if n := namen(t, filepath.Dir(v.pfad)); slicesContains(n, ".tmp-") {
			t.Errorf("%s: Hilfsdatei im Verzeichnis: %v", v.name, n)
		}
	}
}

func slicesContains(namen []string, teil string) bool {
	for _, n := range namen {
		if strings.Contains(n, teil) {
			return true
		}
	}
	return false
}

// T24: der Text ist ODR-E2001 ADMIN_TOKEN_FILE: <Grund> mit einem der sieben Gründe, ohne Inhalt und Pfad.
func TestFehlertext(t *testing.T) {
	sieben := []string{gNichtLesbar, gVerzeichnis, gKeineDatei, gRechte, gMehrZeilen, gZuKurz, gNichtAnlegbar}
	kanarie := string([]byte{0x4b, 0x41, 0x4e, 0x41, 0x52, 0x49, 0x45, 0x2d, 0x58, 0x59, 0x5a})
	dir := filepath.Join(t.TempDir(), kanarie+"-verz")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	datei := filepath.Join(dir, kanarie+".datei")
	schreibe(t, datei, kanarie, 0o600)
	unterverz := filepath.Join(dir, kanarie+"-sub")
	if err := os.Mkdir(unterverz, 0o700); err != nil {
		t.Fatal(err)
	}
	ohneZiel := filepath.Join(dir, kanarie+"-link")
	if err := os.Symlink(filepath.Join(dir, "gibt-es-nicht"), ohneZiel); err != nil {
		t.Fatal(err)
	}
	rechte := filepath.Join(dir, kanarie+"-r")
	schreibe(t, rechte, wdh(kanarie, 3), 0o644)
	zeilen := filepath.Join(dir, kanarie+"-z")
	schreibe(t, zeilen, wdh(kanarie, 3)+"\n"+kanarie, 0o600)
	kurz := filepath.Join(dir, kanarie+"-k")
	schreibe(t, kurz, kanarie, 0o600)
	faelle := []verstoss{
		{"nicht lesbar", filepath.Join(datei, "x"), gNichtLesbar},
		{"Verzeichnis", unterverz, gVerzeichnis},
		{"keine Datei", ohneZiel, gKeineDatei},
		{"Rechte", rechte, gRechte},
		{"Zeilen", zeilen, gMehrZeilen},
		{"zu kurz", kurz, gZuKurz},
		{"nicht anlegbar", filepath.Join(dir, kanarie+"-fehlt", "x"), gNichtAnlegbar},
	}
	gesehen := map[string]bool{}
	for _, f := range faelle {
		_, err := admin.LadeToken(f.pfad)
		erwarteGrund(t, f.name, err, f.grund)
		gesehen[f.grund] = true
		if err == nil {
			continue
		}
		for _, geheim := range []string{kanarie, dir, f.pfad, filepath.Base(f.pfad)} {
			if strings.Contains(err.Error(), geheim) {
				t.Errorf("%s: Text %q nennt %q", f.name, err.Error(), geheim)
			}
		}
	}
	if len(faelle) != 7 || len(gesehen) != 7 {
		t.Errorf("%d Fälle, %d verschiedene Gründe, erwartet 7 und 7", len(faelle), len(gesehen))
	}
	gruendeImQuelltext(t, sieben)
}

// gruendeImQuelltext zählt die Konstanten grund* in tokendatei.go (go/parser) und vergleicht sie mit dem Orakel.
func gruendeImQuelltext(t *testing.T, erwartet []string) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "tokendatei.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var im []string
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, s := range gd.Specs {
			vs := s.(*ast.ValueSpec)
			for i, n := range vs.Names {
				if !strings.HasPrefix(n.Name, "grund") {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					t.Fatalf("%s ist kein Literal", n.Name)
				}
				v, _ := strconv.Unquote(lit.Value)
				im = append(im, v)
			}
		}
	}
	sort.Strings(im)
	want := append([]string(nil), erwartet...)
	sort.Strings(want)
	if len(im) != 7 || !gleich(im, want) {
		t.Errorf("Gründe im Quelltext = %q, erwartet %q (7)", im, want)
	}
}

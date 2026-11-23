//go:build unix

package admin_test

import (
	"bytes"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// Das Fenster zwischen Stat und Lstat (SPEC-239 Nr. 1 a, Nr. 4): TW1 bis TW3. Orakel sind Literale des Vertrags
// (Gründe, Token, Namen) und die Bytes der Datei auf der Platte; einzig TW1 „legt an“ vergleicht mit der Rückgabe
// des inneren LadeToken und prüft sie gegen die Datei. Jeder Aufruf von LadeMitStatHalt läuft unter mitFrist.

var token64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// eintrag ist eine Zeile der Menge M-Eintrag: wie er an pfad entsteht, das Orakel (token oder grund) und die Namen,
// die danach im Verzeichnis liegen (sortiert).
type eintrag struct {
	name   string
	legeAn func(pfad string) error
	token  string
	grund  string
	namen  []string
}

func datei(pfad, inhalt string, modus os.FileMode) error {
	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		return err
	}
	return os.Chmod(pfad, modus)
}

// verzeichnis legt ein Verzeichnis mit den Rechten 0777 an (die Art geht dem Rechteverstoß vor, SPEC-239 Nr. 6).
func verzeichnis(pfad string) error {
	if err := os.Mkdir(pfad, 0o700); err != nil {
		return err
	}
	return os.Chmod(pfad, 0o777)
}

// eintraege sind die elf Einträge von M-Eintrag; der Inhalt entsteht aus ASCII-Literalen.
func eintraege() []eintrag {
	zweiZeilen := hex64() + "\n" + wdh("b", 64) + "\n"
	nur := []string{"admin.token"}
	return []eintrag{
		{"e1 Datei 0600", func(p string) error { return datei(p, hex64()+"\n", 0o600) }, hex64(), "", nur},
		{"e2 Verzeichnis 0777", verzeichnis, "", gVerzeichnis, nur},
		{"e3 FIFO", func(p string) error {
			if err := syscall.Mkfifo(p, 0o666); err != nil {
				return err
			}
			return os.Chmod(p, 0o666)
		}, "", gKeineDatei, nur},
		{"e4 Datei 0644", func(p string) error { return datei(p, hex64()+"\n", 0o644) }, "", gRechte, nur},
		{"e5 zwei Zeilen 0600", func(p string) error { return datei(p, zweiZeilen, 0o600) }, "", gMehrZeilen, nur},
		{"e6 leer 0600", func(p string) error { return datei(p, "", 0o600) }, "", gZuKurz, nur},
		{"e7 zwei Zeilen 0644", func(p string) error { return datei(p, zweiZeilen, 0o644) }, "", gRechte, nur},
		{"e8 Symlink auf Datei", func(p string) error {
			ziel := filepath.Join(filepath.Dir(p), "ziel")
			if err := datei(ziel, hex64()+"\n", 0o600); err != nil {
				return err
			}
			return os.Symlink(ziel, p)
		}, hex64(), "", []string{"admin.token", "ziel"}},
		{"e9 Symlink ohne Ziel", func(p string) error {
			return os.Symlink(filepath.Join(filepath.Dir(p), "gibt-es-nicht"), p)
		}, "", gKeineDatei, nur},
		{"e10 Symlink auf Verzeichnis", func(p string) error {
			ziel := filepath.Join(filepath.Dir(p), "verz")
			if err := verzeichnis(ziel); err != nil {
				return err
			}
			return os.Symlink(ziel, p)
		}, "", gVerzeichnis, []string{"admin.token", "verz"}},
		{"e11 Schleife von Symlinks", func(p string) error {
			schleife := filepath.Join(filepath.Dir(p), "schleife")
			if err := os.Symlink(p, schleife); err != nil {
				return err
			}
			return os.Symlink(schleife, p)
		}, "", gNichtLesbar, []string{"admin.token", "schleife"}},
	}
}

// zugang ist der Zeitpunkt, zu dem der Eintrag erscheint (die Achse Zeitpunkt von M-Eintrag-Zugang).
type zugang int

const (
	vorStat zugang = iota
	zwischenStatLstat
	beimLink
)

func (z zugang) String() string {
	return [...]string{"vor dem Stat", "zwischen Stat und Lstat", "beim Link"}[z]
}

// festeZeit ist die Änderungszeit, auf die jede Fixture vor dem Aufruf gesetzt wird: weit vor der Zeit eines
// Schreibens im Aufruf, damit jedes Schreiben, auch mit denselben Bytes, die Zeit ändert.
func festeZeit() time.Time { return time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC) }

// eStand ist der Stand eines Eintrags, jede Zustandsgröße, an der ein Schreiben, Kürzen, Löschen oder Korrigieren
// (SPEC-239 Nr. 4 Satz 3) sichtbar wird: Art und Rechte (modus), Größe, Bytes einer Datei, Ziel eines Symlinks,
// Änderungszeit, Identität (dev, ino), Zahl der Verweise (nlink) und Besitzer (uid, gid).
type eStand struct {
	fehlt           bool
	modus           fs.FileMode
	groesse         int64
	inhalt          string
	ziel            string
	zeit            int64
	dev, ino, nlink uint64
	uid, gid        uint32
}

func standVon(pfad string) eStand {
	fi, err := os.Lstat(pfad)
	if err != nil {
		return eStand{fehlt: true}
	}
	s := eStand{modus: fi.Mode(), groesse: fi.Size(), zeit: fi.ModTime().UnixNano()}
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		s.dev, s.ino, s.nlink, s.uid, s.gid = uint64(st.Dev), uint64(st.Ino), uint64(st.Nlink), st.Uid, st.Gid
	}
	switch {
	case fi.Mode().IsRegular():
		b, _ := os.ReadFile(pfad)
		s.inhalt = string(b)
	case fi.Mode()&fs.ModeSymlink != 0:
		s.ziel, _ = os.Readlink(pfad)
	}
	return s
}

// legeEintragAn legt den Eintrag an pfad an, setzt die Änderungszeit jeder Datei und jedes Verzeichnisses des Eintrags
// auf festeZeit (ein Symlink selbst bleibt, sein Ziel steht unter seinem Namen in e.namen) und liefert den Stand.
func legeEintragAn(e eintrag, pfad string) (map[string]eStand, error) {
	if err := e.legeAn(pfad); err != nil {
		return nil, err
	}
	dir := filepath.Dir(pfad)
	for _, n := range e.namen {
		p := filepath.Join(dir, n)
		if fi, err := os.Lstat(p); err == nil && fi.Mode()&fs.ModeSymlink == 0 {
			if err := os.Chtimes(p, festeZeit(), festeZeit()); err != nil {
				return nil, err
			}
		}
	}
	return staende(dir, e.namen), nil
}

func staende(dir string, namen []string) map[string]eStand {
	m := map[string]eStand{}
	for _, n := range namen {
		m[n] = standVon(filepath.Join(dir, n))
	}
	return m
}

// zelle ist das Ergebnis eines Aufrufs in einer Zelle: Rückgabe, Verzeichnis danach, Zähler der Haltepunkte und der
// Stand der Einträge im Moment, in dem sie angelegt waren, und nach dem Aufruf.
type zelle struct {
	token                      []byte
	err                        error
	namen                      []string
	nachStat, anlegen, vorLink int
	vorher, nachher            map[string]eStand
}

// laufeZelle legt den Eintrag zum Zeitpunkt z an und ruft LadeMitStatHalt; die Haltepunkte nachAnlegen und vorLink
// des Aufrufs zählen, was der Aufruf an Hilfsdatei und Link tut.
func laufeZelle(t *testing.T, e eintrag, z zugang) zelle {
	t.Helper()
	dir := t.TempDir()
	pfad := filepath.Join(dir, "admin.token")
	var r zelle
	lege := func() {
		var err error
		if r.vorher, err = legeEintragAn(e, pfad); err != nil {
			t.Errorf("%s: Anlegen: %v", e.name, err)
		}
	}
	nachStat := func() { r.nachStat++ }
	vorLink := func(string) { r.vorLink++ }
	switch z {
	case vorStat:
		var err error
		if r.vorher, err = legeEintragAn(e, pfad); err != nil {
			t.Fatalf("%s: Anlegen: %v", e.name, err)
		}
	case zwischenStatLstat:
		nachStat = func() { r.nachStat++; lege() }
	case beimLink:
		vorLink = func(string) { r.vorLink++; lege() }
	}
	r.token, r.err = mitFrist(t, e.name+"/"+z.String(), func() ([]byte, error) {
		return admin.LadeMitStatHalt(pfad, nachStat, func(string) { r.anlegen++ }, vorLink)
	})
	r.namen = namen(t, dir)
	r.nachher = staende(dir, e.namen)
	return r
}

// TW2 (M-Eintrag-Zugang): 11 Einträge x 3 Zugänge gleich 33 Zellen; der Ausgang ist je Eintrag der des Literals,
// in jedem Zugang derselbe. Der Aufruf erzeugt nichts (Hilfsdatei und Link nur beim Zugang „beim Link“, wo der
// Aufruf erst erzeugen will) und hinterlässt allein den Eintrag und seine Ziele.
func TestFensterMengeEintragZugang(t *testing.T) {
	es := eintraege()
	zs := []zugang{vorStat, zwischenStatLstat, beimLink}
	ausgaenge := map[string]int{}
	n := 0
	for _, e := range es {
		for _, z := range zs {
			name := e.name + "/" + z.String()
			r := laufeZelle(t, e, z)
			n++
			if e.token != "" {
				ausgaenge["Token"]++
				if r.err != nil || string(r.token) != e.token {
					t.Errorf("%s: %q, %v; erwartet das Token %q", name, r.token, r.err, e.token)
				}
			} else {
				ausgaenge[e.grund]++
				erwarteGrund(t, name, r.err, e.grund)
				if r.token != nil {
					t.Errorf("%s: Token %q trotz Verstoß", name, r.token)
				}
			}
			if !gleich(r.namen, e.namen) {
				t.Errorf("%s: Verzeichnis = %v, erwartet %v", name, r.namen, e.namen)
			}
			wantStat, wantErzeugen := 1, 0
			if z == vorStat {
				wantStat = 0
				if e.name == "e9 Symlink ohne Ziel" {
					wantStat = 1
				}
			}
			if z == beimLink {
				wantErzeugen = 1
			}
			if r.nachStat != wantStat || r.anlegen != wantErzeugen || r.vorLink != wantErzeugen {
				t.Errorf("%s: nachStat %d, Hilfsdatei %d, Link %d; erwartet %d, %d, %d", name, r.nachStat, r.anlegen, r.vorLink,
					wantStat, wantErzeugen, wantErzeugen)
			}
		}
	}
	if len(es) != 11 || len(zs) != 3 || n != 33 {
		t.Fatalf("Menge: %d Einträge, %d Zugänge, %d Zellen; erwartet 11, 3, 33", len(es), len(zs), n)
	}
	for _, a := range []string{"Token", gVerzeichnis, gKeineDatei, gRechte, gMehrZeilen, gZuKurz, gNichtLesbar} {
		if ausgaenge[a] < 2 {
			t.Errorf("Ausgang %q kam %d-mal vor, erwartet mindestens zweimal", a, ausgaenge[a])
		}
	}
}

// pruefeUnveraendert prüft eine Zelle gegen SPEC-239 Nr. 4 Satz 3: jede Zustandsgröße von eStand jedes Eintrags
// (Art und Rechte, Größe, Bytes, Ziel, Änderungszeit, Identität, Verweise, Besitzer) ist nach dem Aufruf die, die sie
// beim Anlegen war, und im Verzeichnis liegen allein die Namen des Eintrags. Die Änderungszeit der Fixture ist vor
// dem Aufruf auf festeZeit gesetzt, ein Schreiben im Aufruf ändert sie in jedem Lauf.
func pruefeUnveraendert(t *testing.T, name string, e eintrag, r zelle) {
	t.Helper()
	if len(r.vorher) != len(e.namen) {
		t.Errorf("%s: %d Stände beim Anlegen, erwartet %d", name, len(r.vorher), len(e.namen))
	}
	for _, nm := range e.namen {
		if r.vorher[nm] != r.nachher[nm] {
			t.Errorf("%s: %s verändert: %+v -> %+v", name, nm, r.vorher[nm], r.nachher[nm])
		}
	}
	if !gleich(r.namen, e.namen) {
		t.Errorf("%s: Verzeichnis = %v, erwartet %v", name, r.namen, e.namen)
	}
}

// TW3 (M-Eintrag x Zugang zwischen Stat und Lstat, beim Link): 11 x 2 gleich 22 Zellen, je Zelle die Zustandsgrößen
// von eStand (Existenz, Art und Rechte, Größe, Bytes, Ziel, Zeit, Identität, Verweise, Besitzer: 9) und das
// Verzeichnis (pruefeUnveraendert).
func TestFensterAendertNie(t *testing.T) {
	n := 0
	for _, e := range eintraege() {
		for _, z := range []zugang{zwischenStatLstat, beimLink} {
			pruefeUnveraendert(t, e.name+"/"+z.String(), e, laufeZelle(t, e, z))
			n++
		}
	}
	if n != 22 {
		t.Fatalf("Menge: %d Zellen, erwartet 22", n)
	}
}

// zusatzEintraege sind die Arten von SPEC-239 Nr. 1 c, die M-Eintrag nicht führt und die ohne Recht und ohne Blockieren
// entstehen: ein Socket und ein Symlink auf ein Gerät (/dev/null); eine Gerätedatei an pfad selbst braucht das Recht
// mknod und fehlt.
func zusatzEintraege(t *testing.T) []eintrag {
	nur := []string{"admin.token"}
	return []eintrag{
		{"z1 Socket", func(p string) error {
			l, err := net.Listen("unix", p)
			if err != nil {
				return err
			}
			t.Cleanup(func() { _ = l.Close() })
			return nil
		}, "", gKeineDatei, nur},
		{"z2 Symlink auf Gerät", func(p string) error { return os.Symlink("/dev/null", p) }, "", gKeineDatei, nur},
	}
}

// Zusatzmenge: 2 Arten x 3 Zugänge gleich 6 Zellen; Ausgang `keine reguläre Datei` in jedem Zugang, nichts erzeugt,
// nichts verändert.
func TestFensterNichtRegulaerZusatz(t *testing.T) {
	n := 0
	zs := []zugang{vorStat, zwischenStatLstat, beimLink}
	for _, e := range zusatzEintraege(t) {
		for _, z := range zs {
			name := e.name + "/" + z.String()
			r := laufeZelle(t, e, z)
			n++
			erwarteGrund(t, name, r.err, e.grund)
			if r.token != nil {
				t.Errorf("%s: Token %q trotz Verstoß", name, r.token)
			}
			pruefeUnveraendert(t, name, e, r)
			wantStat, wantErzeugen := 1, 0
			if z == vorStat {
				wantStat = 0
			}
			if z == beimLink {
				wantErzeugen = 1
			}
			if r.nachStat != wantStat || r.anlegen != wantErzeugen || r.vorLink != wantErzeugen {
				t.Errorf("%s: nachStat %d, Hilfsdatei %d, Link %d; erwartet %d, %d, %d", name, r.nachStat, r.anlegen, r.vorLink,
					wantStat, wantErzeugen, wantErzeugen)
			}
		}
	}
	if n != 2*len(zs) {
		t.Fatalf("Menge: %d Zellen, erwartet 6", n)
	}
}

// erzeugtOk prüft, dass pfad allein im Verzeichnis liegt, 0600 trägt und das Token plus Zeilenende enthält.
func erzeugtOk(t *testing.T, name, pfad string, token []byte) {
	t.Helper()
	if !token64.Match(token) {
		t.Errorf("%s: Token %q, erwartet 64 Zeichen 0-9a-f", name, token)
	}
	b, err := os.ReadFile(pfad)
	if err != nil || string(b) != string(token)+"\n" {
		t.Errorf("%s: Datei %q (%v), erwartet das Token mit Zeilenende", name, b, err)
	}
	if fi, err := os.Stat(pfad); err != nil || fi.Mode().Perm() != 0o600 {
		t.Errorf("%s: %v, %v; erwartet 0600", name, fi, err)
	}
	if n := namen(t, filepath.Dir(pfad)); !gleich(n, []string{"admin.token"}) {
		t.Errorf("%s: Verzeichnis = %v, erwartet allein admin.token", name, n)
	}
}

// TW1 (M-Aktion): was der Haltepunkt zwischen Stat und Lstat tut, ist legt an, löscht oder nichts; der Haltepunkt
// läuft in jeder Zelle genau einmal.
func TestFensterAktion(t *testing.T) {
	// legt an: ein innerer Aufruf läuft zu Ende; der äußere erzeugt nichts und liefert dessen Token.
	t.Run("legt an", func(t *testing.T) {
		pfad := filepath.Join(t.TempDir(), "admin.token")
		var innen []byte
		var innenErr error
		halte := 0
		aussen, err := mitFrist(t, "legt an", func() ([]byte, error) {
			return admin.LadeMitStatHalt(pfad,
				func() { halte++; innen, innenErr = admin.LadeToken(pfad) },
				func(string) { t.Error("der äußere Aufruf legt eine Hilfsdatei an") },
				func(string) { t.Error("der äußere Aufruf versucht einen Link") })
		})
		if err != nil || innenErr != nil {
			t.Fatalf("außen %v, innen %v", err, innenErr)
		}
		if halte != 1 {
			t.Errorf("Haltepunkt %d-mal gerufen, erwartet 1", halte)
		}
		if !bytes.Equal(aussen, innen) {
			t.Errorf("außen %q, innen %q; erwartet gleich", aussen, innen)
		}
		erzeugtOk(t, "legt an", pfad, aussen)
	})
	// löscht: der Symlink ohne Ziel, der beim Stat lag, ist weg; Lstat meldet „fehlt“, der Aufruf erzeugt.
	t.Run("löscht", func(t *testing.T) {
		pfad := filepath.Join(t.TempDir(), "admin.token")
		if err := os.Symlink(filepath.Join(filepath.Dir(pfad), "gibt-es-nicht"), pfad); err != nil {
			t.Fatal(err)
		}
		halte, anlegen, link := 0, 0, 0
		got, err := mitFrist(t, "löscht", func() ([]byte, error) {
			return admin.LadeMitStatHalt(pfad,
				func() {
					halte++
					if err := os.Remove(pfad); err != nil {
						t.Error(err)
					}
				},
				func(string) { anlegen++ }, func(string) { link++ })
		})
		if err != nil {
			t.Fatalf("Fehler %v", err)
		}
		if halte != 1 || anlegen != 1 || link != 1 {
			t.Errorf("Haltepunkt %d, Hilfsdatei %d, Link %d; erwartet je 1", halte, anlegen, link)
		}
		erzeugtOk(t, "löscht", pfad, got)
	})
	// nichts: nichts liegt an pfad, der Haltepunkt tut nichts, der Aufruf erzeugt wie ohne Haltepunkt.
	t.Run("nichts", func(t *testing.T) {
		pfad := filepath.Join(t.TempDir(), "admin.token")
		halte, anlegen, link := 0, 0, 0
		got, err := mitFrist(t, "nichts", func() ([]byte, error) {
			return admin.LadeMitStatHalt(pfad, func() { halte++ },
				func(string) { anlegen++ }, func(string) { link++ })
		})
		if err != nil {
			t.Fatalf("Fehler %v", err)
		}
		if halte != 1 || anlegen != 1 || link != 1 {
			t.Errorf("Haltepunkt %d, Hilfsdatei %d, Link %d; erwartet je 1", halte, anlegen, link)
		}
		erzeugtOk(t, "nichts", pfad, got)
	})
}

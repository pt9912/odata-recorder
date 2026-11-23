package bootstrap_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// Die festen Texte von SPEC-299 Nr. 6 und 9 als Orakel, ohne den Code von admintls.go.
const (
	adminListenKey    = "ADMIN_LISTEN"
	adminZusammenhang = "außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY"
	adminWarnung      = "WARN|start: Admin-Listener ohne TLS erreichbar|ereignis=admin_klartext"
)

// adminEin bildet die Eingabe von pruefeAdminTransport mit gültiger Form (SPEC-299 Nr. 6).
func adminEin(listen, cert, key string, klartext bool) bootstrap.AdminEingabe {
	return bootstrap.AdminEingabe{Listen: listen, CertPfad: cert, KeyPfad: key, Klartext: klartext, FormGueltig: true}
}

// adminErwartet vergleicht Weg, `TLS` und Verstöße eines Aufrufs mit dem Orakel; `TLS` ist genau bei wegTLS gesetzt.
func adminErwartet(t *testing.T, was string, tr bootstrap.AdminTransport, vs []bootstrap.TLSVerstoss, weg bootstrap.AdminWeg, soll []bootstrap.TLSVerstoss) {
	t.Helper()
	if tr.Weg != weg || (tr.TLS != nil) != (weg == bootstrap.WegTLS) || !slices.Equal(vs, soll) {
		t.Errorf("%s: Weg %d, TLS %v, Verstöße %v; erwartet Weg %d, TLS %v, Verstöße %v", was, tr.Weg, tr.TLS != nil, vs, weg, weg == bootstrap.WegTLS, soll)
	}
}

// adminErfasser hält die Einträge des Loggers als `Stufe|Text|Attribute` fest.
type adminErfasser struct {
	mu        sync.Mutex
	eintraege []string
}

func (e *adminErfasser) Enabled(context.Context, slog.Level) bool { return true }

func (e *adminErfasser) Handle(_ context.Context, r slog.Record) error {
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a.String())
		return true
	})
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eintraege = append(e.eintraege, r.Level.String()+"|"+r.Message+"|"+strings.Join(attrs, ","))
	return nil
}

func (e *adminErfasser) WithAttrs([]slog.Attr) slog.Handler { return e }

func (e *adminErfasser) WithGroup(string) slog.Handler { return e }

func (e *adminErfasser) liste() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Clone(e.eintraege)
}

// mengeAdminPaar ist M-Paar (SPEC-299 Nr. 4): die 9 einzelnen Verstöße (Rang 1 bis 3 an beiden Schlüsseln, Rang 4: abgelaufen, noch nicht gültig,
// Schlüssel passt nicht), beide Pfade leer und die 2 Paare von Verstößen (Rang 2 und Rang 3 an beiden Schlüsseln), unter den Namen `ADMIN_TLS_*`.
func (w *werkstatt) mengeAdminPaar() []tlsFall {
	n := tlsNamen{adminCert, adminKey}
	spaeter := w.wurzel.Blatt(w.t, tlstest.Optionen{Von: w.jetzt.Add(time.Hour), Bis: w.jetzt.Add(2 * time.Hour)})
	return []tlsFall{
		w.paar(n, zLeer, zGueltig), w.paar(n, zGueltig, zLeer),
		w.paar(n, zNichtLesbar, zGueltig), w.paar(n, zGueltig, zNichtLesbar),
		w.paar(n, zKeinPEM, zGueltig), w.paar(n, zGueltig, zKeinPEM),
		w.paar(n, zBesonders, zGueltig),
		w.fall(n, "Zertifikat noch nicht gültig", w.datei(spaeter.ZertifikatPEM()), w.datei(w.schluessel(spaeter)), tv(adminCert, gAusserhalb)),
		w.paar(n, zGueltig, zBesonders),
		w.paar(n, zLeer, zLeer),
		w.paar(n, zNichtLesbar, zNichtLesbar), w.paar(n, zKeinPEM, zKeinPEM),
	}
}

// adminLoopbackHosts sind die drei Schreibweisen, die SPEC-299 Nr. 2 als Loopback nennt.
func adminLoopbackHosts() []string { return []string{"127.0.0.1", "[::1]", "localhost"} }

// adminLoopbackSoll ist das Orakel von SPEC-299 Nr. 2 ohne den Code von admintls.go: wahr genau dann, wenn s aus einer der drei Schreibweisen, einem `:`
// und einem Rest ohne `:` besteht (der Host ist der Text vor dem letzten `:`, bytegleich, und der Port ändert nichts).
func adminLoopbackSoll(s string) bool {
	for _, h := range adminLoopbackHosts() {
		if rest, ok := strings.CutPrefix(s, h+":"); ok && !strings.Contains(rest, ":") {
			return true
		}
	}
	return false
}

// adminNachbarn liefert jede Zeichenkette im Abstand 1 von s, ohne s: jeder der 256 Bytewerte an jeder Stelle eingefügt oder an die Stelle eines Bytes
// gesetzt, dazu jedes Byte gelöscht. Die Bytes entstehen aus ihrem Wert, nie aus einem Zeichen des Quelltexts (kein Alphabet, das ein Wert verfehlt).
func adminNachbarn(s string) map[string]bool {
	aus := map[string]bool{}
	for i := 0; i <= len(s); i++ {
		for c := range 256 {
			b := string([]byte{byte(c)})
			aus[s[:i]+b+s[i:]] = true
			if i < len(s) {
				aus[s[:i]+b+s[i+1:]] = true
			}
		}
		if i < len(s) {
			aus[s[:i]+s[i+1:]] = true
		}
	}
	delete(aus, s)
	return aus
}

// adminZaehlformel ist die Zahl der Nachbarn nach einem zweiten Verfahren, ohne die Menge zu bilden: Löschen je Lauf gleicher Bytes, Ersetzen je Byte mit
// den 255 anderen, Einfügen an n+1 Stellen mit 256 Werten abzüglich der n Doppelten (dasselbe Byte vor und hinter seinem Gleichen).
func adminZaehlformel(s string) int {
	laeufe := 0
	for i := range s {
		if i == 0 || s[i] != s[i-1] {
			laeufe++
		}
	}
	n := len(s)
	return laeufe + n*255 + (n+1)*256 - n
}

// adminSchreibweisen liefert jede Groß- und Kleinschreibung der Buchstaben von s (s kleingeschrieben) ohne s selbst: 2^len(s) - 1.
func adminSchreibweisen(s string) []string {
	var aus []string
	for maske := 1; maske < 1<<len(s); maske++ {
		b := []byte(s)
		for i := range b {
			if maske>>i&1 == 1 {
				b[i] -= 'a' - 'A'
			}
		}
		aus = append(aus, string(b))
	}
	return aus
}

// adminUnicode sind 31 Zeichen, die ein Trimmen oder Falten des Hosts treffen könnte: die 25 Leerzeichen von `unicode.IsSpace` (sechs davon in ASCII, die
// übrigen mit zwei oder drei Bytes) und sechs Formatzeichen (Weicher Trennstrich, Nullbreiten-Zeichen, Wortverbinder, BOM); sie entstehen aus ihrem
// Wert (`AGENTS.md` §3.18).
func adminUnicode() []rune {
	r := []rune{0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x20, 0x85, 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xAD, 0x200B, 0x200C, 0x200D, 0x2060, 0xFEFF}
	for c := rune(0x2000); c <= 0x200A; c++ {
		r = append(r, c)
	}
	return r
}

// TestAdminLoopback (AT01, M-Host): adminLoopback ist wahr genau für `127.0.0.1`, `[::1]` und `localhost`, bytegleich, bei jedem Port; jeder andere
// Host, auch ohne Host und ohne `:`, ist falsch (SPEC-299 Nr. 2). Menge: 3 + 17 Hosts × 3 Ports = 60, dazu 3 Eingaben ohne `:` = 63.
func TestAdminLoopback(t *testing.T) {
	zellen := adminMengeHost()
	for _, z := range zellen {
		if got := bootstrap.AdminLoopback(z.listen); got != z.loopback {
			t.Errorf("AdminLoopback(%q) = %v, erwartet %v", z.listen, got, z.loopback)
		}
	}
	if len(zellen) != 63 {
		t.Fatalf("Menge M-Host hat %d Zellen, erwartet 63 (3 + 17 Hosts × 3 Ports und 3 Eingaben ohne `:`)", len(zellen))
	}
}

// adminZelle ist eine Eingabe `listen` mit der Antwort, die SPEC-299 Nr. 2 für sie verlangt (Literal, nie aus dem Code).
type adminZelle struct {
	listen   string
	loopback bool
}

// adminMengeHost ist M-Host (SPEC-299 Nr. 2): 3 Hosts loopback und 17 nicht-loopback bei drei Ports, dazu drei Eingaben ohne `:`.
func adminMengeHost() []adminZelle {
	loopback := adminLoopbackHosts()
	fremd := []string{
		"0.0.0.0", "[::]", "", "127.0.0.2", "127.255.255.255", "LOCALHOST", "Localhost", "localhost.", "xlocalhost", "localhostx",
		"[0:0:0:0:0:0:0:1]", "[::ffff:127.0.0.1]", "[::2]", "::1", "10.0.0.1", "recorder.internal", "127.0.0.1.",
	}
	var aus []adminZelle
	for _, port := range []string{"1", "8081", "65535"} {
		for _, h := range loopback {
			aus = append(aus, adminZelle{h + ":" + port, true})
		}
		for _, h := range fremd {
			aus = append(aus, adminZelle{h + ":" + port, false})
		}
	}
	for _, ohneDoppelpunkt := range []string{"127.0.0.1", "localhost", ""} {
		aus = append(aus, adminZelle{ohneDoppelpunkt, false})
	}
	return aus
}

// adminMengeStruktur ist M-Struktur (SPEC-299 Nr. 2): Port-Teil und Klammern der drei Hosts, Schreibweisen derselben Adressen und Namen mit und ohne Port
// und 31 Unicode-Zeichen vor dem Host, hinter dem Host und im Port: 3 × 12 + 16 + 17 × 2 + 3 × 31 × 3 = 36 + 16 + 34 + 279 = 365 Zellen.
func adminMengeStruktur() []adminZelle {
	var aus []adminZelle
	formen := []struct {
		muster string
		wahr   bool
	}{
		{"%s", false}, {"%s:", true}, {"%s:8081", true}, {"%s:x", true}, {"%s:8081 ", true}, {"%s:8081]", true},
		{"%s:8081:8081", false}, {"%s::", false}, {"%s::8081", false}, {":%s:8081", false}, {" %s:8081", false}, {"%s :8081", false},
	}
	for _, h := range adminLoopbackHosts() {
		for _, f := range formen {
			aus = append(aus, adminZelle{fmt.Sprintf(f.muster, h), f.wahr})
		}
	}
	aus = append(aus, []adminZelle{
		{"[::1]:8081", true}, {"[::1]", false}, {"::1", false}, {"::1:8081", false}, {"::1]:8081", false}, {"[::1:8081", false}, {"[::1:8081]", false},
		{"[[::1]]:8081", false}, {"[[::1]:8081", false}, {"[::1]]:8081", false}, {"[::1]:[::1]:8081", false}, {"[ ::1]:8081", false}, {"[::1 ]:8081", false},
		{"[::1]%lo:8081", false}, {"[::1%lo]:8081", false}, {":[::1]:8081", false},
	}...)
	for _, h := range []string{
		"127.1", "127.0.1", "0177.0.0.1", "0x7f.0.0.1", "0x7f000001", "2130706433", "127.000.000.001", "[::0001]", "[0::1]", "[::ffff:7f00:1]",
		"[::1%lo]", "localhost6", "ip6-localhost", "localhost.localdomain", "LocalHost", "127.0.0.1%lo", "localho" + string(rune(0x17F)) + "t",
	} {
		aus = append(aus, adminZelle{h + ":8081", false}, adminZelle{h, false})
	}
	for _, h := range adminLoopbackHosts() {
		for _, c := range adminUnicode() {
			aus = append(aus, adminZelle{string(c) + h + ":8081", false}, adminZelle{h + string(c) + ":8081", false}, adminZelle{h + ":80" + string(c) + "81", true})
		}
	}
	return aus
}

// TestAdminLoopbackNachbarn (AT01, M-Nachbar, Zeichenmenge): jede Zeichenkette im Abstand 1 von einer der sechs Eingaben (drei Hosts, je mit `:8081` und
// ohne Port), jedes Byte von 0 bis 255 an jeder Stelle eingefügt, ersetzt oder gelöscht, und jede andere Schreibweise von `localhost` (mit und ohne Port)
// hat die Antwort des Orakels von SPEC-299 Nr. 2. Die Zahl der Nachbarn und die der wahren je Eingabe sind Literale und stimmen mit einem zweiten Verfahren
// überein: Nachbarn nach `adminZaehlformel`; wahre: der Port ändert nichts, also bleibt wahr, wer nur im Port ohne `:` ändert (Einfügen an 5 Stellen mit
// 255 Bytewerten ohne `:` abzüglich 4 Doppelter = 1271, Ersetzen an 4 Stellen mit 254 = 1016, Löschen 4), und ohne Port allein das angehängte `:`.
func TestAdminLoopbackNachbarn(t *testing.T) {
	const port = ":8081"
	faelle := []struct {
		basis      string
		zahl, wahr int
	}{
		{"127.0.0.1", 4855, 1}, {"127.0.0.1" + port, 7410, 2291},
		{"[::1]", 2810, 1}, {"[::1]" + port, 5365, 2291},
		{"localhost", 4855, 1}, {"localhost" + port, 7410, 2291},
	}
	summe, wahre := 0, 0
	pruefe := func(was string, menge map[string]bool) (wahr int) {
		var falsch []string
		for n := range menge {
			soll := adminLoopbackSoll(n)
			if soll {
				wahr++
			}
			if got := bootstrap.AdminLoopback(n); got != soll {
				falsch = append(falsch, n)
			}
		}
		slices.Sort(falsch)
		if len(falsch) > 0 {
			t.Errorf("%s: %d von %d Eingaben mit falscher Antwort, die ersten: %q", was, len(falsch), len(menge), falsch[:min(len(falsch), 3)])
		}
		return wahr
	}
	for _, f := range faelle {
		menge := adminNachbarn(f.basis)
		if len(menge) != f.zahl || adminZaehlformel(f.basis) != f.zahl {
			t.Fatalf("%q: Generator %d, Formel %d, Literal %d", f.basis, len(menge), adminZaehlformel(f.basis), f.zahl)
		}
		if got := pruefe("Nachbarn von "+f.basis, menge); got != f.wahr {
			t.Fatalf("%q: %d wahre Nachbarn nach dem Orakel, erwartet %d", f.basis, got, f.wahr)
		}
		summe += len(menge)
		wahre += f.wahr
	}
	schreibweisen := adminSchreibweisen("localhost")
	if len(schreibweisen) != 511 {
		t.Fatalf("%d Schreibweisen von localhost, erwartet 511", len(schreibweisen))
	}
	for _, s := range schreibweisen {
		if bootstrap.AdminLoopback(s+port) || bootstrap.AdminLoopback(s) {
			t.Errorf("AdminLoopback(%q) ist wahr, die Schreibweise zählt", s)
		}
	}
	if summe != 32705 || wahre != 6876 {
		t.Fatalf("%d Nachbarn, davon %d wahre; erwartet 32705 und 6876", summe, wahre)
	}
}

// TestAdminLoopbackStruktur (AT01, M-Nachbar, Strukturmenge): die Form der Eingabe, getrennt vom Zeichen: Port-Teil und Klammern der drei Hosts als Tabelle
// mit Antwort als Literal, die Schreibweisen derselben Adressen und Namen mit und ohne Port (alle falsch) und 31 Unicode-Zeichen (Leerzeichen und Formatzeichen) vor dem Host,
// hinter dem Host und im Port. Menge: 3 × 12 Formen + 16 Klammern + 17 × 2 Schreibweisen + 3 × 31 × 3 Stellen = 36 + 16 + 34 + 279 = 365.
func TestAdminLoopbackStruktur(t *testing.T) {
	zellen := adminMengeStruktur()
	for _, z := range zellen {
		if got := bootstrap.AdminLoopback(z.listen); got != z.loopback {
			t.Errorf("AdminLoopback(%q) = %v, erwartet %v", z.listen, got, z.loopback)
		}
	}
	if len(adminUnicode()) != 31 || len(zellen) != 365 {
		t.Fatalf("%d Zellen und %d Zeichen; erwartet 365 und 31", len(zellen), len(adminUnicode()))
	}
}

// TestAdminLoopbackPorts (AT01, Achse Port von M-Host): der Port ändert nichts (SPEC-299 Nr. 2): jede Zahl von 0 bis 65535 und sieben Texte, die keine
// Portnummer sind, hinter den drei Hosts (wahr) und hinter drei anderen (falsch); 6 × (65536 + 7) Eingaben, die der Test zählt.
func TestAdminLoopbackPorts(t *testing.T) {
	ports := []string{"", "x", "-1", "65536", "99999999999", " ", "8081 "}
	for p := range 65536 {
		ports = append(ports, strconv.Itoa(p))
	}
	zahl := 0
	for _, h := range []struct {
		host string
		soll bool
	}{{"127.0.0.1", true}, {"[::1]", true}, {"localhost", true}, {"127.0.0.2", false}, {"localhost.", false}, {"0.0.0.0", false}} {
		for _, p := range ports {
			zahl++
			if got := bootstrap.AdminLoopback(h.host + ":" + p); got != h.soll {
				t.Fatalf("AdminLoopback(%q) = %v, erwartet %v", h.host+":"+p, got, h.soll)
			}
		}
	}
	if zahl != 6*(65536+7) {
		t.Fatalf("%d Eingaben, erwartet %d", zahl, 6*(65536+7))
	}
}

// TestAdminWeg (AT02, M-Weg): der Weg folgt der Tabelle von SPEC-299 Nr. 3; TLS gewinnt über Loopback und Flag, das Flag wirkt allein im Weg
// Klartext. Menge: 8 Hosts (3 loopback, 5 nicht-loopback) × TLS (gültiges Paar, kein Paar) × Flag = 32 Zellen.
func TestAdminWeg(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	hosts := []struct {
		listen   string
		loopback bool
	}{
		{"127.0.0.1:8081", true}, {"[::1]:8081", true}, {"localhost:8081", true},
		{"0.0.0.0:8081", false}, {"[::]:8081", false}, {":8081", false}, {"recorder.internal:8081", false}, {"10.0.0.1:8081", false},
	}
	zeilen := []struct {
		paar, flag     bool
		loopback, rest bootstrap.AdminWeg // der Weg bei Loopback und bei einem anderen Host
	}{
		{false, false, bootstrap.WegLoopback, bootstrap.WegKeiner},
		{false, true, bootstrap.WegLoopback, bootstrap.WegKlartext},
		{true, false, bootstrap.WegTLS, bootstrap.WegTLS},
		{true, true, bootstrap.WegTLS, bootstrap.WegTLS},
	}
	zahl := 0
	for _, h := range hosts {
		for _, z := range zeilen {
			zahl++
			c, k := "", ""
			if z.paar {
				c, k = cert, key
			}
			weg, soll := z.rest, []bootstrap.TLSVerstoss(nil)
			if h.loopback {
				weg = z.loopback
			} else if weg == bootstrap.WegKeiner {
				soll = []bootstrap.TLSVerstoss{{Schluessel: adminListenKey, Grund: adminZusammenhang}}
			}
			tr, vs := bootstrap.PruefeAdminTransport(adminEin(h.listen, c, k, z.flag), w.jetzt)
			adminErwartet(t, fmt.Sprintf("%s, Paar %v, Flag %v", h.listen, z.paar, z.flag), tr, vs, weg, soll)
		}
	}
	if zahl != 32 {
		t.Fatalf("Menge M-Weg hat %d Zellen, erwartet 32", zahl)
	}
}

// adminMengeNachbarn ist M-Nachbar (SPEC-299 Nr. 2) als Zellen: die Nachbarn der sechs Eingaben und die 2 × 511 anderen Schreibweisen von `localhost`; die
// Antwort ist das Orakel `adminLoopbackSoll` (33727 Zellen).
func adminMengeNachbarn() []adminZelle {
	var aus []adminZelle
	for _, h := range adminLoopbackHosts() {
		for _, basis := range []string{h, h + ":8081"} {
			for n := range adminNachbarn(basis) {
				aus = append(aus, adminZelle{n, adminLoopbackSoll(n)})
			}
		}
	}
	for _, s := range adminSchreibweisen("localhost") {
		aus = append(aus, adminZelle{s + ":8081", false}, adminZelle{s, false})
	}
	return aus
}

// TestAdminWegHosts (AT02, Aufrufstelle von adminLoopback in pruefeAdminTransport): die Hostmengen von AT01 gehen als `Listen` durch pruefeAdminTransport, und
// der Weg folgt dem Orakel von SPEC-299 Nr. 2 und 3: ohne Paar und ohne Flag ist ein Loopback-Host `wegLoopback`, jeder andere der Verstoß an `ADMIN_LISTEN`;
// mit Flag ist jeder andere Host `wegKlartext`; mit Paar ist es `wegTLS`. Menge: M-Host (63) und M-Struktur (365) in drei Armen (ohne Paar und Flag, mit Flag,
// mit Paar) = 1284 Zellen, M-Nachbar (33727) in den zwei Armen ohne Paar = 67454; 68738 Zellen.
func TestAdminWegHosts(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	var falsch []string
	zahl := 0
	pruefe := func(z adminZelle, c, k string, flag bool) {
		zahl++
		weg, soll := bootstrap.WegKeiner, []bootstrap.TLSVerstoss{tv(adminListenKey, adminZusammenhang)}
		switch {
		case c != "":
			weg, soll = bootstrap.WegTLS, nil
		case z.loopback:
			weg, soll = bootstrap.WegLoopback, nil
		case flag:
			weg, soll = bootstrap.WegKlartext, nil
		}
		tr, vs := bootstrap.PruefeAdminTransport(adminEin(z.listen, c, k, flag), w.jetzt)
		if tr.Weg != weg || (tr.TLS != nil) != (weg == bootstrap.WegTLS) || !slices.Equal(vs, soll) {
			falsch = append(falsch, fmt.Sprintf("%q, Paar %v, Flag %v: Weg %d, Verstöße %v; erwartet Weg %d, Verstöße %v", z.listen, c != "", flag, tr.Weg, vs, weg, soll))
		}
	}
	for _, z := range slices.Concat(adminMengeHost(), adminMengeStruktur()) {
		pruefe(z, "", "", false)
		pruefe(z, "", "", true)
		pruefe(z, cert, key, false)
	}
	for _, z := range adminMengeNachbarn() {
		pruefe(z, "", "", false)
		pruefe(z, "", "", true)
	}
	slices.Sort(falsch)
	if len(falsch) > 0 {
		t.Errorf("%d von %d Zellen mit falschem Weg, die ersten: %q", len(falsch), zahl, falsch[:min(len(falsch), 3)])
	}
	if zahl != 68738 {
		t.Fatalf("%d Zellen, erwartet 68738", zahl)
	}
}

// TestAdminJetzt (AT04, Argument `jetzt` von pruefeAdminTransport): die Gültigkeit des Zertifikats gilt zur Zeit `jetzt`, nicht zur Uhr des Hosts (SPEC-299
// Nr. 4 und 13). Menge: Fenster des Blatts (Vergangenheit, Gegenwart, Zukunft) × `jetzt` (Mitte der Vergangenheit, Gegenwart, Mitte der Zukunft) = 9
// Zellen; gültig ist genau die Diagonale (`wegTLS`), jede andere Zelle der Verstoß „außerhalb seiner Gültigkeit“ am Zertifikat; zwei der drei `jetzt` liegen
// 25 und 75 Jahre von der Uhr des Hosts.
func TestAdminJetzt(t *testing.T) {
	w := neueWerkstatt(t)
	tag := func(jahr int, monat time.Month) time.Time { return time.Date(jahr, monat, 15, 12, 0, 0, 0, time.UTC) }
	fenster := []struct{ von, bis time.Time }{
		{tag(2000, 1), tag(2001, 1)}, {w.jetzt.Add(-time.Hour), w.jetzt.Add(time.Hour)}, {tag(2100, 1), tag(2101, 1)},
	}
	jetzte := []time.Time{tag(2000, 6), w.jetzt, tag(2100, 6)}
	zahl := 0
	for i, f := range fenster {
		blatt := w.wurzel.Blatt(t, tlstest.Optionen{Von: f.von, Bis: f.bis})
		cert, key := w.datei(blatt.ZertifikatPEM()), w.datei(w.schluessel(blatt))
		for j, jetzt := range jetzte {
			zahl++
			weg, soll := bootstrap.WegKeiner, []bootstrap.TLSVerstoss{tv(adminCert, gAusserhalb)}
			if i == j {
				weg, soll = bootstrap.WegTLS, nil
			}
			tr, vs := bootstrap.PruefeAdminTransport(adminEin("0.0.0.0:8081", cert, key, false), jetzt)
			adminErwartet(t, fmt.Sprintf("Fenster %d, jetzt %d", i, j), tr, vs, weg, soll)
		}
	}
	if zahl != 9 {
		t.Fatalf("%d Zellen, erwartet 9", zahl)
	}
}

// TestAdminKeinRueckfall (AT03, M-Rueckfall): ein Paar mit Verstoß liefert keinen Weg, und der Start fällt nie auf Klartext zurück, auch nicht mit
// Flag und nicht bei Loopback (SPEC-299 Nr. 3). Menge: die 9 einzelnen Verstöße von M-Paar × Host (nicht-loopback, loopback) × Flag wahr = 18.
func TestAdminKeinRueckfall(t *testing.T) {
	w := neueWerkstatt(t)
	einzeln := w.mengeAdminPaar()[:9]
	zahl := 0
	for _, f := range einzeln {
		for _, host := range []string{"0.0.0.0:8081", "127.0.0.1:8081"} {
			zahl++
			tr, vs := bootstrap.PruefeAdminTransport(adminEin(host, f.cert, f.key, true), f.jetzt)
			adminErwartet(t, f.name+", "+host, tr, vs, bootstrap.WegKeiner, f.erwartet)
		}
	}
	if zahl != 18 {
		t.Fatalf("Menge M-Rueckfall hat %d Zellen, erwartet 18", zahl)
	}
}

// TestAdminPaar (AT04, M-Paar): das Paar geht unverändert durch ladeTLS unter den Namen `ADMIN_TLS_*`: jeder Verstoß steht unter dem Namen seines
// Schlüssels, der Grund des Rangs 1 nennt den anderen Namen, zwei Verstöße stehen nach Schlüssel sortiert, beide Pfade leer ist kein TLS
// (SPEC-299 Nr. 4). Menge: 12 Zellen (9 einzelne Verstöße, beide Pfade leer, 2 Paare); das Orakel ist die Liste der Verstöße als Literal.
func TestAdminPaar(t *testing.T) {
	w := neueWerkstatt(t)
	arbeit := t.TempDir()
	tlstest.Datei(t, arbeit, "cert.pem", w.gut.ZertifikatPEM(), 0o600)
	tlstest.Datei(t, arbeit, "key.pem", w.schluessel(w.gut), 0o600)
	t.Chdir(arbeit) // das Arbeitsverzeichnis trägt gültige Dateien `cert.pem` und `key.pem`, die bei leeren Pfaden nichts liest
	faelle := w.mengeAdminPaar()
	if len(faelle) != 12 {
		t.Fatalf("Menge M-Paar hat %d Zellen, erwartet 12", len(faelle))
	}
	for _, f := range faelle {
		tr, vs := bootstrap.PruefeAdminTransport(adminEin("127.0.0.1:8081", f.cert, f.key, false), f.jetzt)
		weg := bootstrap.WegKeiner
		if f.kein {
			weg = bootstrap.WegLoopback
		}
		adminErwartet(t, f.name, tr, vs, weg, f.erwartet)
	}
	// Die Zeile einer Zelle mit zwei Verstößen steht in der Reihenfolge der Schlüssel, `ADMIN_TLS_CERT` vor `ADMIN_TLS_KEY`.
	for _, i := range []int{10, 11} {
		if got := faelle[i].erwartet; len(got) != 2 || got[0].Schluessel != adminCert || got[1].Schluessel != adminKey {
			t.Fatalf("Orakel der Zelle %q: %v", faelle[i].name, got)
		}
	}
}

// TestAdminEinPfad (AT05, M-EinPfad): genau ein gesetzter Pfad ist der Verstoß des Rangs 1 am leeren Schlüssel, der gesetzte wird weiter geprüft, und
// der Zusammenhang entfällt (SPEC-299 Nr. 4, 6). Menge: Seite (nur Zertifikat, nur Schlüssel) × Host × Flag × gesetzte Datei (gültig, Verzeichnis) = 16.
func TestAdminEinPfad(t *testing.T) {
	w := neueWerkstatt(t)
	gueltig := map[bool]string{true: w.datei(w.gut.ZertifikatPEM()), false: w.datei(w.schluessel(w.gut))}
	zahl := 0
	for _, nurZertifikat := range []bool{true, false} {
		for _, host := range []string{"127.0.0.1:8081", "0.0.0.0:8081"} {
			for _, flag := range []bool{false, true} {
				for _, verzeichnis := range []bool{false, true} {
					zahl++
					pfad := gueltig[nurZertifikat]
					if verzeichnis {
						pfad = w.dir
					}
					e, soll := adminEin(host, "", pfad, flag), []bootstrap.TLSVerstoss{tv(adminCert, gErforderlich(adminKey))}
					if nurZertifikat {
						e.CertPfad, e.KeyPfad = pfad, ""
						soll = []bootstrap.TLSVerstoss{tv(adminKey, gErforderlich(adminCert))}
					}
					if verzeichnis && nurZertifikat {
						soll = []bootstrap.TLSVerstoss{tv(adminCert, gNichtLesbar), soll[0]}
					} else if verzeichnis {
						soll = []bootstrap.TLSVerstoss{soll[0], tv(adminKey, gNichtLesbar)}
					}
					tr, vs := bootstrap.PruefeAdminTransport(e, w.jetzt)
					adminErwartet(t, host, tr, vs, bootstrap.WegKeiner, soll)
				}
			}
		}
	}
	if zahl != 16 {
		t.Fatalf("Menge M-EinPfad hat %d Zellen, erwartet 16", zahl)
	}
}

// adminNichtLoopback ist M-Nichtloopback (SPEC-299 Nr. 2, 6): Wildcard IPv4, Wildcard IPv6, leerer Host, Name, IP-Adresse.
func adminNichtLoopback() []string {
	return []string{"0.0.0.0:8081", "[::]:8081", ":8081", "recorder.internal:8081", "10.0.0.5:8081"}
}

// TestAdminZusammenhang (AT06, M-Nichtloopback): der Zusammenhang ist genau eine Zeile am Schlüssel `ADMIN_LISTEN` mit festem Grund, ohne den Wert von
// `ADMIN_LISTEN`, ohne Pfad und ohne den Namen des Flags (SPEC-299 Nr. 6). Menge: 5 Hosts, kein Pfad, Flag falsch, Form gültig.
func TestAdminZusammenhang(t *testing.T) {
	w := neueWerkstatt(t)
	hosts := adminNichtLoopback()
	for _, listen := range hosts {
		tr, vs := bootstrap.PruefeAdminTransport(adminEin(listen, "", "", false), w.jetzt)
		adminErwartet(t, listen, tr, vs, bootstrap.WegKeiner, []bootstrap.TLSVerstoss{tv(adminListenKey, adminZusammenhang)})
		host, port, err := net.SplitHostPort(listen)
		if err != nil {
			t.Fatal(err)
		}
		for _, teil := range []string{host, port, listen, "PLAINTEXT"} { // der Wert von ADMIN_LISTEN in drei Zerlegungen und der Name des Flags
			if len(vs) == 1 && teil != "" && strings.Contains(vs[0].Grund+vs[0].Schluessel, teil) {
				t.Errorf("%s: die Zeile %v nennt %q", listen, vs[0], teil)
			}
		}
	}
	if len(hosts) != 5 {
		t.Fatalf("Menge M-Nichtloopback hat %d Elemente, erwartet 5", len(hosts))
	}
}

// TestAdminFormUngueltig (AT07, M-Form): bei `FormGueltig` falsch entfällt der Zusammenhang (kein Weg, kein Verstoß an `ADMIN_LISTEN`); die Dateien
// eines gesetzten Paares werden dennoch geprüft (SPEC-299 Nr. 6). Menge: 5 Hosts × Paar (kein Pfad, gültiges Paar, fehlende Datei) × Flag = 30.
func TestAdminFormUngueltig(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	paare := []struct {
		name, cert, key string
		soll            []bootstrap.TLSVerstoss
	}{
		{"kein Pfad", "", "", nil},
		{"gültiges Paar", cert, key, nil},
		{"fehlende Datei", w.fehlt(), key, []bootstrap.TLSVerstoss{tv(adminCert, gNichtLesbar)}},
	}
	zahl := 0
	for _, listen := range adminNichtLoopback() {
		for _, p := range paare {
			for _, flag := range []bool{false, true} {
				zahl++
				e := adminEin(listen, p.cert, p.key, flag)
				e.FormGueltig = false
				tr, vs := bootstrap.PruefeAdminTransport(e, w.jetzt)
				adminErwartet(t, listen+", "+p.name, tr, vs, bootstrap.WegKeiner, p.soll)
			}
		}
	}
	if zahl != 30 {
		t.Fatalf("Menge M-Form hat %d Zellen, erwartet 30", zahl)
	}
}

// TestAdminWarnung (AT08, M-Warnung): warne schreibt genau dann einen Eintrag, wenn der Weg wegKlartext ist: Stufe WARN, fester Text, ein Attribut
// `ereignis`; `log` nil verwirft (SPEC-299 Nr. 9). Menge: Weg (4) × Log (nil, Handler) = 8.
func TestAdminWarnung(t *testing.T) {
	zahl := 0
	for _, weg := range []bootstrap.AdminWeg{bootstrap.WegKeiner, bootstrap.WegTLS, bootstrap.WegLoopback, bootstrap.WegKlartext} {
		tr := bootstrap.AdminTransport{Weg: weg}
		zahl++
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Weg %d, log nil: Panik %v", weg, r)
				}
			}()
			bootstrap.AdminWarne(tr, nil)
		}()
		zahl++
		e := &adminErfasser{}
		bootstrap.AdminWarne(tr, slog.New(e))
		var soll []string
		if weg == bootstrap.WegKlartext {
			soll = []string{adminWarnung}
		}
		if got := e.liste(); !slices.Equal(got, soll) {
			t.Errorf("Weg %d: Einträge %q, erwartet %q", weg, got, soll)
		}
	}
	if zahl != 8 {
		t.Fatalf("Menge M-Warnung hat %d Zellen, erwartet 8", zahl)
	}
}

// ausgabeErfasst führt f aus und liefert, was in dieser Zeit nach `os.Stdout`, `os.Stderr` und `log.Default()` geschrieben wurde.
func ausgabeErfasst(t *testing.T, f func()) string {
	t.Helper()
	var puffer syncPuffer
	var wg sync.WaitGroup
	altOut, altErr, altLog := os.Stdout, os.Stderr, log.Writer()
	var enden []*os.File
	for _, ziel := range []**os.File{&os.Stdout, &os.Stderr} {
		lesen, schreiben, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		*ziel = schreiben
		enden = append(enden, schreiben)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = io.Copy(&puffer, lesen)
			_ = lesen.Close()
		}()
	}
	log.SetOutput(&puffer)
	func() {
		defer func() {
			os.Stdout, os.Stderr = altOut, altErr
			log.SetOutput(altLog)
			for _, e := range enden {
				_ = e.Close()
			}
		}()
		f()
	}()
	wg.Wait()
	return puffer.String()
}

// TestAdminKeinGeheimnis (AT09, M-Geheimnis): weder Pfad noch Inhalt noch Host noch eine Eigenschaft des Zertifikats stehen in einer Zeile oder im Eintrag;
// die Funktionen schreiben nichts nach stdout, stderr und `log.Default()` und ändern, löschen und verschieben keine Datei (SPEC-299 Nr. 12, 13). Menge:
// Ausgabeweg (Zeilen der Verstöße, Eintrag der Warnung, übrige Ausgabe) × Canary (Pfad, Inhalt, Host) = 9, dazu der Schnappschuss je Zelle von M-Paar (12);
// „die Funktionen“ ist die Menge von `adminFunktionen` (5), jede auf allen vier Wegen unter der Erfassung von `stdout`, `stderr` und `log.Default()`.
func TestAdminKeinGeheimnis(t *testing.T) {
	w := neueWerkstatt(t) // das Verzeichnis trägt `geheim-pfad-4711`, das Subject `geheim-cn-4711`
	const text, host = "geheim-text-4711", "geheim-host-4711.example"
	n := tlsNamen{adminCert, adminKey}
	vor := []byte(text + "\n")
	faelle := append(w.mengeAdminPaar(),
		w.fall(n, "Text vor kein PEM", w.datei(slices.Concat(vor, []byte("kein PEM\n"))), w.datei(w.schluessel(w.gut)), tv(adminCert, gKeinZert)),
		w.fall(n, "Text vor abgelaufenem Zertifikat", w.datei(slices.Concat(vor, w.abgelaufen.ZertifikatPEM())), w.datei(slices.Concat(vor, w.schluessel(w.abgelaufen))), tv(adminCert, gAusserhalb)),
		w.fall(n, "Text vor Schlüssel, der nicht passt", w.datei(slices.Concat(vor, w.gut.ZertifikatPEM())), w.datei(slices.Concat(vor, w.schluessel(w.fremd))), tv(adminKey, gPasstNicht)))
	var zeilen []string
	var geaendert []string
	erfasser := &adminErfasser{}
	soll := 0
	funktionen := w.adminFunktionen(w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut)))
	uebrig := ausgabeErfasst(t, func() {
		for _, f := range funktionen { // jede der fünf Funktionen auf jedem Weg: was sie nach stdout, stderr und log.Default() schreiben, steht in `uebrig`
			f.ruf()
		}
		for _, f := range faelle {
			soll += len(f.erwartet)
			vorher := w.momentaufnahme(f.cert, f.key)
			_, vs := bootstrap.PruefeAdminTransport(adminEin("127.0.0.1:8081", f.cert, f.key, false), f.jetzt)
			if nachher := w.momentaufnahme(f.cert, f.key); nachher != vorher {
				geaendert = append(geaendert, f.name)
			}
			for _, v := range vs {
				zeilen = append(zeilen, zeile(v.Schluessel, v.Grund))
			}
		}
		_, vs := bootstrap.PruefeAdminTransport(adminEin(host+":8081", "", "", false), w.jetzt)
		for _, v := range vs {
			zeilen = append(zeilen, zeile(v.Schluessel, v.Grund))
		}
		tr, _ := bootstrap.PruefeAdminTransport(adminEin(host+":8081", "", "", true), w.jetzt)
		bootstrap.AdminWarne(tr, slog.New(erfasser))
	})
	if len(geaendert) != 0 {
		t.Errorf("Dateien oder Verzeichnis geändert bei: %q", geaendert)
	}
	wege := map[string]string{"Zeilen der Verstöße": strings.Join(zeilen, "\n"), "Eintrag der Warnung": strings.Join(erfasser.liste(), "\n"), "übrige Ausgabe": uebrig}
	canaries := map[string][]string{"Pfad": {"geheim-pfad-4711"}, "Inhalt": {text, "geheim-cn-4711"}, "Host": {host}}
	zahl := 0
	for weg, ausgabe := range wege {
		for art, werte := range canaries {
			zahl++
			for _, c := range werte {
				if strings.Contains(ausgabe, c) {
					t.Errorf("%s nennt den Canary %s %q:\n%s", weg, art, c, ausgabe)
				}
			}
		}
	}
	if zahl != 9 || len(zeilen) != soll+1 || len(erfasser.liste()) != 1 || uebrig != "" || len(faelle) != 15 || len(funktionen) != 5 {
		t.Errorf("Zellen %d, Zeilen %d (erwartet %d), Einträge %d, übrige Ausgabe %q", zahl, len(zeilen), soll+1, len(erfasser.liste()), uebrig)
	}
}

// adminAttrappe ist ein net.Listener, der jede seiner Operationen zählt; umhuelle ruft keine.
type adminAttrappe struct{ n atomic.Int32 }

func (a *adminAttrappe) Accept() (net.Conn, error) { a.n.Add(1); return nil, net.ErrClosed }
func (a *adminAttrappe) Close() error              { a.n.Add(1); return nil }
func (a *adminAttrappe) Addr() net.Addr            { a.n.Add(1); return &net.TCPAddr{} }

// adminKonfig nennt die Felder eines `*tls.Config`, die der Betrieb liest und die keine Funktion von SPEC-299 nach dem Laden ändert.
func adminKonfig(cfg *tls.Config) string {
	if cfg == nil {
		return "nil"
	}
	return fmt.Sprintf("%x|%d|%d|%d|%q", cfg.Certificates[0].Certificate, cfg.MinVersion, cfg.MaxVersion, cfg.ClientAuth, cfg.NextProtos)
}

// adminFunktion ist eine der Funktionen von SPEC-299 mit festem Eingang auf jedem Weg; ruf liefert eine Beschreibung des Ergebnisses, die jeder Aufruf
// gleich hat und die soll aus den Zusagen des Vertrags nennt (nie aus einem Aufruf: ein Aufruf trüge eine Wirkung der Funktion in die Erwartung).
type adminFunktion struct {
	name, soll string
	ruf        func() string
}

// adminFunktionen führt die Menge, über die die Allaussagen von SPEC-299 Nr. 12 und 13 gehen: `adminLoopback`, `pruefeAdminTransport`, `umhuelle`, `warne`
// und der Wert `adminTransport`, den `umhuelle` und `warne` gemeinsam lesen. Jede Funktion läuft auf allen vier Wegen; der geteilte Wert des Weges `wegTLS`
// ist derselbe für alle Aufrufe, `adminTransport` kommt zuletzt und prüft ihn nach den Aufrufen der anderen.
func (w *werkstatt) adminFunktionen(cert, key string) []adminFunktion {
	wege := []bootstrap.AdminWeg{bootstrap.WegKeiner, bootstrap.WegTLS, bootstrap.WegLoopback, bootstrap.WegKlartext}
	geteilt := w.ladeAdmin(cert, key, w.jetzt)
	// Die Konfiguration, die SPEC-299 Nr. 7 und SPEC-291 Nr. 5 und 6 verlangen: das Blatt der Datei, TLS 1.2 ohne Obergrenze, kein Client-Zertifikat, nur HTTP/1.1.
	vorher := adminKonfig(&tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{w.gut.Zertifikat.Raw}}}, MinVersion: tls.VersionTLS12, NextProtos: []string{"http/1.1"}})
	transport := func(weg bootstrap.AdminWeg) bootstrap.AdminTransport {
		if weg == bootstrap.WegTLS {
			return geteilt
		}
		return bootstrap.AdminTransport{Weg: weg}
	}
	eingaben := []bootstrap.AdminEingabe{
		adminEin("127.0.0.1:8081", "", "", false), adminEin("0.0.0.0:8081", "", "", false), adminEin("0.0.0.0:8081", "", "", true), adminEin("0.0.0.0:8081", cert, key, false),
	}
	var ll, pr, um, wa string // die Beschreibungen, die das Orakel verlangt: aus den Zusagen von SPEC-299, nicht aus einem Aufruf
	for _, l := range []string{"127.0.0.1:8081", "[::1]:1", "localhost:65535", "0.0.0.0:8081", "LOCALHOST:8081", ""} {
		ll += fmt.Sprintf("%v;", l == "127.0.0.1:8081" || l == "[::1]:1" || l == "localhost:65535")
	}
	for _, p := range []struct {
		weg bootstrap.AdminWeg
		vs  []bootstrap.TLSVerstoss
		cfg string
	}{
		{bootstrap.WegLoopback, nil, "nil"}, {bootstrap.WegKeiner, []bootstrap.TLSVerstoss{tv(adminListenKey, adminZusammenhang)}, "nil"},
		{bootstrap.WegKlartext, nil, "nil"}, {bootstrap.WegTLS, nil, vorher},
	} {
		pr += fmt.Sprintf("%d|%v|%s;", p.weg, p.vs, p.cfg)
	}
	for _, weg := range wege {
		um += fmt.Sprintf("%d:%v:0;", weg, weg != bootstrap.WegTLS) // der Listener selbst bis auf wegTLS, keine Operation von ihm gerufen
		eintraege := []string(nil)
		if weg == bootstrap.WegKlartext {
			eintraege = []string{adminWarnung}
		}
		wa += fmt.Sprintf("%d:%q;", weg, eintraege)
	}
	return []adminFunktion{
		{name: "adminLoopback", soll: ll, ruf: func() string {
			var sb strings.Builder
			for _, l := range []string{"127.0.0.1:8081", "[::1]:1", "localhost:65535", "0.0.0.0:8081", "LOCALHOST:8081", ""} {
				fmt.Fprintf(&sb, "%v;", bootstrap.AdminLoopback(l))
			}
			return sb.String()
		}},
		{name: "pruefeAdminTransport", soll: pr, ruf: func() string {
			var sb strings.Builder
			for _, e := range eingaben {
				tr, vs := bootstrap.PruefeAdminTransport(e, w.jetzt)
				fmt.Fprintf(&sb, "%d|%v|%s;", tr.Weg, vs, adminKonfig(tr.TLS))
			}
			return sb.String()
		}},
		{name: "umhuelle", soll: um, ruf: func() string {
			var sb strings.Builder
			for _, weg := range wege {
				l := &adminAttrappe{}
				fmt.Fprintf(&sb, "%d:%v:%d;", weg, bootstrap.AdminUmhuelle(transport(weg), l) == l, l.n.Load())
			}
			return sb.String()
		}},
		{name: "warne", soll: wa, ruf: func() string {
			var sb strings.Builder
			for _, weg := range wege {
				e := &adminErfasser{}
				bootstrap.AdminWarne(transport(weg), slog.New(e))
				bootstrap.AdminWarne(transport(weg), nil)
				fmt.Fprintf(&sb, "%d:%q;", weg, e.liste())
			}
			return sb.String()
		}},
		{name: "adminTransport", soll: fmt.Sprint(bootstrap.WegTLS, vorher), ruf: func() string { return fmt.Sprint(geteilt.Weg, adminKonfig(geteilt.TLS)) }},
	}
}

// adminQuelleOhneZustand liest `admintls.go` und meldet jede Deklaration `var` und jede Funktion `init` auf der obersten Ebene, jeden Import von `os`,
// `io/ioutil`, `path/filepath`, `syscall` und `log`, jeden Aufruf von `print` und `println` und jede Funktion außerhalb der vier, die `adminFunktionen`
// führt (SPEC-299 Nr. 12, 13): Zustand im Paket braucht eine solche Deklaration in der Datei, ein Dateizugriff und jede Ausgabe, die `ausgabeErfasst` nicht
// sieht, einen dieser Importe oder die eingebauten Funktionen, und die Menge der Funktionen ist die der Quelle. Eine Deklaration `var` in einer anderen
// Datei des Pakets fängt dieser Test nicht.
func adminQuelleOhneZustand(t *testing.T) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "admintls.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var funktionen []string
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			if d.Tok == token.VAR {
				t.Errorf("admintls.go: Deklaration `var` auf der obersten Ebene (Zustand im Paket)")
			}
		case *ast.FuncDecl:
			if d.Name.Name == "init" && d.Recv == nil {
				t.Errorf("admintls.go: Funktion init")
			}
			funktionen = append(funktionen, d.Name.Name)
		}
	}
	for _, i := range f.Imports {
		if slices.Contains([]string{`"os"`, `"io/ioutil"`, `"path/filepath"`, `"syscall"`, `"log"`}, i.Path.Value) {
			t.Errorf("admintls.go: Import %s (Ausgabe oder Dateizugriff)", i.Path.Value)
		}
	}
	ast.Inspect(f, func(n ast.Node) bool { // die eingebauten print und println schreiben auf fd 2 am Austausch von os.Stderr vorbei
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && (id.Name == "print" || id.Name == "println") {
				t.Errorf("admintls.go: Aufruf von %s (Ausgabe auf fd 2)", id.Name)
			}
		}
		return true
	})
	slices.Sort(funktionen)
	if !slices.Equal(funktionen, []string{"adminLoopback", "pruefeAdminTransport", "umhuelle", "warne"}) {
		t.Errorf("admintls.go: Funktionen %q, erwartet adminLoopback, pruefeAdminTransport, umhuelle und warne (jede Funktion braucht ihren Fänger in adminFunktionen)", funktionen)
	}
}

// TestAdminGleichzeitig (AT18, M-Gleichzeitig): die Funktionen halten keinen Zustand, gleichzeitige Aufrufe teilen nichts, jeder Aufruf liefert ein eigenes
// `*tls.Config`, und zwei Paare auf denselben Dateien sind unabhängig (SPEC-299 Nr. 13). 16 Goroutinen × 20 Läufe = 320 Aufrufe je Art (dasselbe Paar,
// ein zweites Paar auf denselben Dateien, warne); den Zustand im Paket fängt der Detektor von `make test-race` je Prozess, einen Cache zusätzlich die Zeiger.
// Dieselben 16 × 20 Läufe gelten je Funktion der Menge von `adminFunktionen` (5 × 320): jeder Lauf liefert die Beschreibung, die das Orakel verlangt,
// `umhuelle` ruft keine Operation des Listeners, der geteilte Wert bleibt, was er war, und die Quelle trägt keine Deklaration, die Zustand im Paket hielte.
func TestAdminGleichzeitig(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	const goroutinen, laeufe = 16, 20
	adminQuelleOhneZustand(t)
	funktionen := w.adminFunktionen(cert, key)
	for _, f := range funktionen {
		soll := f.soll
		ergebnisse := make([]string, goroutinen*laeufe)
		var gruppe sync.WaitGroup
		start := make(chan struct{})
		for g := range goroutinen {
			gruppe.Add(1)
			go func() {
				defer gruppe.Done()
				<-start
				for l := range laeufe {
					ergebnisse[g*laeufe+l] = f.ruf()
				}
			}()
		}
		close(start)
		gruppe.Wait()
		for i, r := range ergebnisse {
			if r != soll {
				t.Errorf("%s, Lauf %d von %d: %s, erwartet %s", f.name, i, len(ergebnisse), r, soll)
				break
			}
		}
	}
	if len(funktionen) != 5 {
		t.Fatalf("Menge der Funktionen hat %d Elemente, erwartet 5", len(funktionen))
	}
	erfasser := &adminErfasser{}
	var mu sync.Mutex
	var konfigs []*tls.Config
	var beschreibungen []string
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range laeufe {
				for range 2 { // dasselbe Paar, dann ein zweites Paar auf denselben Dateien
					tr, vs := bootstrap.PruefeAdminTransport(adminEin("127.0.0.1:8081", cert, key, false), w.jetzt)
					if tr.Weg != bootstrap.WegTLS || len(vs) != 0 {
						t.Errorf("Weg %d, Verstöße %v; erwartet wegTLS ohne Verstoß", tr.Weg, vs)
						continue
					}
					mu.Lock()
					konfigs = append(konfigs, tr.TLS)
					beschreibungen = append(beschreibungen, beschreibe(tr.TLS))
					mu.Unlock()
				}
				tr, _ := bootstrap.PruefeAdminTransport(adminEin("0.0.0.0:8081", "", "", true), w.jetzt)
				bootstrap.AdminWarne(tr, slog.New(erfasser))
			}
		}()
	}
	close(start)
	wg.Wait()
	verschieden := map[*tls.Config]bool{}
	for _, k := range konfigs {
		verschieden[k] = true
	}
	if len(konfigs) != 2*goroutinen*laeufe || len(verschieden) != len(konfigs) || len(erfasser.liste()) != goroutinen*laeufe {
		t.Fatalf("%d Konfigurationen, davon %d verschiedene, %d Einträge; erwartet je 640 und 320", len(konfigs), len(verschieden), len(erfasser.liste()))
	}
	for i, b := range beschreibungen {
		if b != beschreibungen[0] {
			t.Fatalf("Aufruf %d: %s, erwartet %s", i, b, beschreibungen[0])
		}
	}
	konfigs[0].NextProtos[0] = "veraendert"
	for i, k := range konfigs[1:] {
		if k.NextProtos[0] != "http/1.1" {
			t.Fatalf("Konfiguration %d teilt die Liste der Protokolle mit der ersten", i+1)
		}
	}
}

// ladeAdmin lädt das Paar wie der Start (nicht-loopback, ohne Flag, `FormGueltig` wahr) und verlangt wegTLS ohne Verstoß.
func (w *werkstatt) ladeAdmin(cert, key string, jetzt time.Time) bootstrap.AdminTransport {
	w.t.Helper()
	tr, vs := bootstrap.PruefeAdminTransport(adminEin("0.0.0.0:8081", cert, key, false), jetzt)
	if tr.Weg != bootstrap.WegTLS || len(vs) != 0 {
		w.t.Fatalf("ladeAdmin: Weg %d, Verstöße %v; erwartet wegTLS", tr.Weg, vs)
	}
	return tr
}

// TestAdminUmhuelle (AT10, M-Umhuellung): umhuelle umhüllt nur im Weg wegTLS, sonst liefert sie den übergebenen Listener selbst (SPEC-299 Nr. 8 c).
// Menge: Weg (4); bei wegTLS spricht ein Client mit der CA TLS und ein Klartext-Client erhält keine Antwort, bei den anderen drei gelingt ein Klartext-Client.
func TestAdminUmhuelle(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	mitTLS := w.ladeAdmin(cert, key, w.jetzt)
	zahl := 0
	for _, weg := range []bootstrap.AdminWeg{bootstrap.WegKeiner, bootstrap.WegTLS, bootstrap.WegLoopback, bootstrap.WegKlartext} {
		zahl++
		tr := bootstrap.AdminTransport{Weg: weg}
		if weg == bootstrap.WegTLS {
			tr = mitTLS
		}
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		ul := bootstrap.AdminUmhuelle(tr, l)
		if (ul == l) == (weg == bootstrap.WegTLS) {
			t.Errorf("Weg %d: der Listener ist %s, erwartet %s", weg, map[bool]string{true: "derselbe", false: "ein anderer"}[ul == l], map[bool]string{true: "ein anderer", false: "derselbe"}[weg == bootstrap.WegTLS])
		}
		antwort := func(dial func(addr string) (net.Conn, error)) string {
			fertig := make(chan struct{})
			go func() {
				defer close(fertig)
				c, err := ul.Accept()
				if err != nil {
					return
				}
				defer func() { _ = c.Close() }()
				_ = c.SetDeadline(time.Now().Add(5 * time.Second))
				if n, err := c.Read(make([]byte, 16)); err == nil && n > 0 {
					_, _ = c.Write([]byte("ok"))
				}
			}()
			defer func() { <-fertig }()
			c, err := dial(l.Addr().String())
			if err != nil {
				_ = ul.Close() // der Accept kehrt zurück
				return "Fehler " + err.Error()
			}
			defer func() { _ = c.Close() }()
			_ = c.SetDeadline(time.Now().Add(5 * time.Second))
			_, _ = c.Write([]byte("GET / HTTP/1.1\r\n\r\n"))
			b, err := io.ReadAll(c)
			if err != nil && len(b) == 0 {
				return "Fehler " + err.Error()
			}
			return string(b)
		}
		klartext := antwort(func(a string) (net.Conn, error) { return net.DialTimeout("tcp", a, 5*time.Second) })
		if got, soll := klartext == "ok", weg != bootstrap.WegTLS; got != soll {
			t.Errorf("Weg %d: Klartext-Client erhält %q, erwartet %s", weg, klartext, map[bool]string{true: "ok", false: "keine Antwort"}[soll])
		}
		if weg == bootstrap.WegTLS {
			mitCA := antwort(func(a string) (net.Conn, error) {
				return tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", a, w.wurzel.Klient("localhost"))
			})
			if mitCA != "ok" {
				t.Errorf("Client mit der CA erhält %q, erwartet ok", mitCA)
			}
		}
		_ = ul.Close()
		_ = l.Close()
	}
	if zahl != 4 {
		t.Fatalf("Menge M-Umhuellung hat %d Elemente, erwartet 4", zahl)
	}
}

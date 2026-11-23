package bootstrap_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
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

// adminAlphabet sind die 68 Zeichen der Nachbarn von M-Nachbar (SPEC-299 Nr. 2).
const adminAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.:[]-_"

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

// adminNachbarn liefert jede Zeichenkette im Abstand 1 von s (ein Zeichen des Alphabets eingefügt, gelöscht oder ersetzt), ohne s.
func adminNachbarn(s string) map[string]bool {
	aus := map[string]bool{}
	for i := 0; i <= len(s); i++ {
		for _, c := range adminAlphabet {
			aus[s[:i]+string(c)+s[i:]] = true
			if i < len(s) {
				aus[s[:i]+string(c)+s[i+1:]] = true
			}
		}
		if i < len(s) {
			aus[s[:i]+s[i+1:]] = true
		}
	}
	delete(aus, s)
	return aus
}

// adminZaehlformel ist die Zahl der Nachbarn nach einem zweiten Verfahren: Löschen je Lauf gleicher Zeichen, Ersetzen je Zeichen mit den 67 anderen,
// Einfügen an n+1 Stellen mit 68 Zeichen abzüglich der n Doppelten (dasselbe Zeichen vor und hinter seinem Gleichen).
func adminZaehlformel(s string) int {
	laeufe := 0
	for i := range s {
		if i == 0 || s[i] != s[i-1] {
			laeufe++
		}
	}
	n, a := len(s), len(adminAlphabet)
	return laeufe + n*(a-1) + (n+1)*a - n
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

// TestAdminLoopback (AT01, M-Host): adminLoopback ist wahr genau für `127.0.0.1`, `[::1]` und `localhost`, bytegleich, bei jedem Port; jeder andere
// Host, auch ohne Host und ohne `:`, ist falsch (SPEC-299 Nr. 2). Menge: 3 + 17 Hosts × 3 Ports = 60, dazu 3 Eingaben ohne `:` = 63.
func TestAdminLoopback(t *testing.T) {
	loopback := []string{"127.0.0.1", "[::1]", "localhost"}
	fremd := []string{
		"0.0.0.0", "[::]", "", "127.0.0.2", "127.255.255.255", "LOCALHOST", "Localhost", "localhost.", "xlocalhost", "localhostx",
		"[0:0:0:0:0:0:0:1]", "[::ffff:127.0.0.1]", "[::2]", "::1", "10.0.0.1", "recorder.internal", "127.0.0.1.",
	}
	zahl := 0
	pruefe := func(listen string, soll bool) {
		t.Helper()
		zahl++
		if got := bootstrap.AdminLoopback(listen); got != soll {
			t.Errorf("AdminLoopback(%q) = %v, erwartet %v", listen, got, soll)
		}
	}
	for _, port := range []string{"1", "8081", "65535"} {
		for _, h := range loopback {
			pruefe(h+":"+port, true)
		}
		for _, h := range fremd {
			pruefe(h+":"+port, false)
		}
	}
	for _, ohneDoppelpunkt := range []string{"127.0.0.1", "localhost", ""} {
		pruefe(ohneDoppelpunkt, false)
	}
	if len(loopback) != 3 || len(fremd) != 17 || zahl != 63 {
		t.Fatalf("Menge M-Host hat %d Zellen (%d + %d Hosts), erwartet 63 (3 + 17)", zahl, len(loopback), len(fremd))
	}
}

// TestAdminLoopbackNachbarn (AT01, M-Nachbar): jede Zeichenkette im Abstand 1 von einem der drei Hosts und jede andere Schreibweise von `localhost`
// ist nicht-loopback. Die Zahl je Host ist ein Literal und stimmt mit der Formel überein (Generator und Formel sind zwei Verfahren).
func TestAdminLoopbackNachbarn(t *testing.T) {
	for _, h := range []struct {
		host string
		zahl int
	}{{"127.0.0.1", 1283}, {"[::1]", 742}, {"localhost", 1785}} {
		menge := adminNachbarn(h.host)
		soll := adminZaehlformel(h.host)
		if h.host == "localhost" {
			schreibweisen := adminSchreibweisen(h.host)
			if len(schreibweisen) != 511 {
				t.Fatalf("%d Schreibweisen von localhost, erwartet 511", len(schreibweisen))
			}
			for _, s := range schreibweisen {
				menge[s] = true
			}
			soll += len(schreibweisen) - len(h.host) // die neun Schreibweisen mit einem Großbuchstaben sind Nachbarn zugleich
		}
		if len(menge) != h.zahl || soll != h.zahl {
			t.Fatalf("%s: Generator %d, Formel %d, Literal %d", h.host, len(menge), soll, h.zahl)
		}
		for n := range menge {
			if bootstrap.AdminLoopback(n + ":8081") {
				t.Errorf("AdminLoopback(%q) ist wahr, ein Nachbar von %q ist nicht-loopback", n+":8081", h.host)
			}
		}
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
// Ausgabeweg (Zeilen der Verstöße, Eintrag der Warnung, übrige Ausgabe) × Canary (Pfad, Inhalt, Host) = 9, dazu der Schnappschuss je Zelle von M-Paar (12).
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
	uebrig := ausgabeErfasst(t, func() {
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
	if zahl != 9 || len(zeilen) != soll+1 || len(erfasser.liste()) != 1 || uebrig != "" || len(faelle) != 15 {
		t.Errorf("Zellen %d, Zeilen %d (erwartet %d), Einträge %d, übrige Ausgabe %q", zahl, len(zeilen), soll+1, len(erfasser.liste()), uebrig)
	}
}

// TestAdminGleichzeitig (AT18, M-Gleichzeitig): die Funktionen halten keinen Zustand, gleichzeitige Aufrufe teilen nichts, jeder Aufruf liefert ein eigenes
// `*tls.Config`, und zwei Paare auf denselben Dateien sind unabhängig (SPEC-299 Nr. 13). 16 Goroutinen × 20 Läufe = 320 Aufrufe je Art (dasselbe Paar,
// ein zweites Paar auf denselben Dateien, warne); den Zustand im Paket fängt der Detektor von `make test-race` je Prozess, einen Cache zusätzlich die Zeiger.
func TestAdminGleichzeitig(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	const goroutinen, laeufe = 16, 20
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

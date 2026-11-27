package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// Die Zeilen der Paare als Literale (SPEC-299 Nr. 4, SPEC-291 Nr. 2) und die Zeile der Warnung ohne Zeitstempel (SPEC-299 Nr. 9).
const (
	zeileAdminCertFehlt  = "ODR-E2001 ADMIN_TLS_CERT: Datei nicht lesbar"
	zeileAdminKeyFehlt   = "ODR-E2001 ADMIN_TLS_KEY: Datei nicht lesbar"
	zeileListenCertFehlt = "ODR-E2001 LISTEN_TLS_CERT: Datei nicht lesbar"
	zeileListenKeyFehlt  = "ODR-E2001 LISTEN_TLS_KEY: Datei nicht lesbar"
	zeileAbgelaufen      = "ODR-E2001 ADMIN_TLS_CERT: Zertifikat außerhalb seiner Gültigkeit"
	zeileWarnung         = `level=WARN msg="start: Admin-Listener ohne TLS erreichbar" ereignis=admin_klartext`
)

// warnungen liefert jede Zeile von stderr, die die Warnung nennt, ohne den Zeitstempel.
func warnungen(stderr string) []string {
	var aus []string
	for _, z := range strings.Split(stderr, "\n") {
		if strings.Contains(z, "ohne TLS erreichbar") || strings.Contains(z, "admin_klartext") {
			_, rest, _ := strings.Cut(z, " ")
			aus = append(aus, rest)
		}
	}
	return aus
}

// TestAdminLaden (C03): 6 Verstöße des Transports × {ohne, mit Verstoß von LISTEN_TLS_*} × 4 Modi = 48 Zellen über Start, dazu je
// Modus einer mit beendetem ctx = 52. Orakel: Rückgabe 1, auf stderr allein die Zeilen als Literal nach Schlüssel sortiert, DATA_DIR
// nicht angelegt, keine Token-Datei, kein Bind, die Ablage nie geöffnet.
func TestAdminLaden(t *testing.T) {
	w := neueWerkstatt(t)
	fehlt := w.fehlt()
	paar := []string{"ADMIN_TLS_CERT=" + fehlt, "ADMIN_TLS_KEY=" + fehlt}
	faelle := []struct{ env, zeilen []string }{
		{[]string{"ADMIN_LISTEN=0.0.0.0:8081"}, []string{zeileZusammenhang}},
		{paar, []string{zeileAdminCertFehlt, zeileAdminKeyFehlt}},
		{[]string{"ADMIN_LISTEN=0.0.0.0:8081", "ADMIN_ALLOW_PLAINTEXT=yes"}, []string{zeileFlag}},
		{append([]string{"ADMIN_ALLOW_PLAINTEXT=yes"}, paar...), []string{zeileFlag, zeileAdminCertFehlt, zeileAdminKeyFehlt}},
		{append([]string{"ADMIN_LISTEN=keinport"}, paar...), []string{zeileForm, zeileAdminCertFehlt, zeileAdminKeyFehlt}},
		{[]string{"ADMIN_TLS_CERT=" + w.datei(w.abgelaufen.ZertifikatPEM()), "ADMIN_TLS_KEY=" + w.datei(w.schluessel(w.abgelaufen))},
			[]string{zeileAbgelaufen}},
	}
	zellen := 0
	for _, modus := range vierModi() {
		for i, f := range faelle {
			for _, mitListen := range []bool{false, true} {
				zellen++
				env, soll := f.env, f.zeilen
				if mitListen {
					env = append(slices.Clone(env), "LISTEN_TLS_CERT="+fehlt, "LISTEN_TLS_KEY="+fehlt)
					soll = append(slices.Clone(soll), zeileListenCertFehlt, zeileListenKeyFehlt)
				}
				pruefeKeinStart(context.Background(), t, fmt.Sprintf("%s %d, LISTEN_TLS %v", modus, i, mitListen), modus, env, soll)
			}
		}
		zellen++
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		pruefeKeinStart(ctx, t, modus+" Paar fehlt, ctx beendet", modus, paar, []string{zeileAdminCertFehlt, zeileAdminKeyFehlt})
	}
	if zellen != 52 {
		t.Errorf("%d Zellen, erwartet 52", zellen)
	}
}

// pruefeKeinStart fährt Start mit env und erwartet allein die Zeilen soll ohne jede Wirkung auf Ablage, Token-Datei und Listener.
func pruefeKeinStart(ctx context.Context, t *testing.T, name, modus string, env, soll []string) {
	t.Helper()
	wurzel := t.TempDir()
	daten, tokenPfad := filepath.Join(wurzel, "daten"), filepath.Join(wurzel, "eigen.token")
	binds, oeffnen := 0, 0
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	stderr := &syncPuffer{}
	env = adminEnv(modus, daten, append([]string{"ADMIN_TOKEN_FILE=" + tokenPfad}, env...)...)
	code := bootstrap.StartMit(ctx, env, stderr, zaehlendeBauteile(&binds, &oeffnen))
	if want := strings.Join(soll, "\n") + "\n"; code != 1 || stderr.String() != want {
		t.Errorf("%s: Rückgabe %d, stderr %q, erwartet 1 und %q", name, code, stderr.String(), want)
	}
	if fileExists(daten) || fileExists(tokenPfad) || binds != 0 || oeffnen != 0 {
		t.Errorf("%s: DATA_DIR %v, Token-Datei %v, %d Binds, Ablage %d-mal", name, fileExists(daten), fileExists(tokenPfad), binds, oeffnen)
	}
}

// TestAdminBuild (C04): Literale der Konfiguration. 6 Verstöße {Admin-Paar, LISTEN_TLS_*, beide, Zusammenhang, abgelaufenes
// Admin-Paar, Host ohne Form} × 4 Modi = 24 Zellen: allein die Zeilen als Literal, weder Datenbank noch Blob-Verzeichnis, kein Bind.
func TestAdminBuild(t *testing.T) {
	w := neueWerkstatt(t)
	fehlt := w.fehlt()
	altCert, altKey := w.datei(w.abgelaufen.ZertifikatPEM()), w.datei(w.schluessel(w.abgelaufen))
	adminPaar := func(k *bootstrap.Konfiguration) { k.AdminTLSCert, k.AdminTLSKey = fehlt, fehlt }
	listenPaar := func(k *bootstrap.Konfiguration) { k.ListenTLSCert, k.ListenTLSKey = fehlt, fehlt }
	faelle := []struct {
		setze func(*bootstrap.Konfiguration)
		soll  []string
	}{
		{adminPaar, []string{zeileAdminCertFehlt, zeileAdminKeyFehlt}},
		{listenPaar, []string{zeileListenCertFehlt, zeileListenKeyFehlt}},
		{func(k *bootstrap.Konfiguration) { adminPaar(k); listenPaar(k) },
			[]string{zeileAdminCertFehlt, zeileAdminKeyFehlt, zeileListenCertFehlt, zeileListenKeyFehlt}},
		{func(k *bootstrap.Konfiguration) { k.AdminListen = "0.0.0.0:0" }, []string{zeileZusammenhang}},
		{func(k *bootstrap.Konfiguration) { k.AdminTLSCert, k.AdminTLSKey = altCert, altKey }, []string{zeileAbgelaufen}},
		{func(k *bootstrap.Konfiguration) { k.AdminListen = "keinport" }, []string{zeileZusammenhang}},
	}
	zellen := 0
	for _, modus := range vierModi() {
		for i, f := range faelle {
			zellen++
			dir := t.TempDir()
			k := listenerKonfig(dir, modus)
			f.setze(&k)
			binds, oeffnen := 0, 0
			h, err := bootstrap.BuildMit(context.Background(), k, nil, zaehlendeBauteile(&binds, &oeffnen))
			if got := konfigZeilen(t, err); h != nil || !slices.Equal(got, f.soll) {
				t.Errorf("%s %d: %q, erwartet %q ohne Handle", modus, i, got, f.soll)
			}
			if fileExists(filepath.Join(dir, dateiDB)) || fileExists(filepath.Join(dir, verzeichnisB)) || oeffnen != 0 || binds != 0 {
				t.Errorf("%s %d: die Ablage oder ein Listener wurde geöffnet (%d, %d)", modus, i, oeffnen, binds)
			}
		}
	}
	if zellen != 24 {
		t.Errorf("%d Zellen, erwartet 24", zellen)
	}
}

// TestAdminBuildOhneListener (C04, SPEC-300 Nr. 11): ein leeres AdminListen mit einem Paar, dessen Dateien fehlen, × {ohne, mit
// Verstoß von LISTEN_TLS_*} × 4 Modi = 8 Zellen: Build ruft die Prüfung nicht, hält keinen Transport und meldet allein die Zeilen
// von LISTEN_TLS_*.
func TestAdminBuildOhneListener(t *testing.T) {
	fehlt := filepath.Join(t.TempDir(), "fehlt.pem")
	zellen := 0
	for _, modus := range vierModi() {
		for _, mitListen := range []bool{false, true} {
			zellen++
			k := listenerKonfig(t.TempDir(), modus)
			k.AdminListen, k.AdminTLSCert, k.AdminTLSKey = "", fehlt, fehlt
			var soll []string
			if mitListen {
				k.ListenTLSCert, k.ListenTLSKey = fehlt, fehlt
				soll = []string{zeileListenCertFehlt, zeileListenKeyFehlt}
			}
			p := &pruefProtokoll{echt: true}
			tr, zeilen := buildMit(t, p, k)
			if !slices.Equal(zeilen, soll) || len(p.eingaben) != 0 || tr.Weg != bootstrap.WegKeiner || tr.TLS != nil {
				t.Errorf("%s %v: Zeilen %q, %d Aufrufe, Weg %d; erwartet %q ohne Aufruf", modus, mitListen, zeilen, len(p.eingaben), tr.Weg, soll)
			}
		}
	}
	if zellen != 8 {
		t.Errorf("%d Zellen, erwartet 8", zellen)
	}
}

// startLauf ist ein Start mit Haken am Bind: jede Adresse geht auf 127.0.0.1:0, die Adresse des zweiten Binds (ADMIN_LISTEN)
// landet in admin.
type startLauf struct {
	stderr   *syncPuffer
	admin    chan string
	ergebnis chan int
	abbruch  context.CancelFunc
}

func starte(env []string) startLauf {
	ctx, abbruch := context.WithTimeout(context.Background(), 20*time.Second)
	l := startLauf{stderr: &syncPuffer{}, admin: make(chan string, 1), ergebnis: make(chan int, 1), abbruch: abbruch}
	b := bootstrap.StandardBauteile()
	echt := b.Listen
	binds := 0
	b.Listen = func(ctx context.Context, network, _ string) (net.Listener, error) {
		lis, err := echt(ctx, network, "127.0.0.1:0")
		if binds++; err == nil && binds == 2 {
			l.admin <- lis.Addr().String()
		}
		return lis, err
	}
	go func() { l.ergebnis <- bootstrap.StartMit(ctx, env, l.stderr, b) }()
	return l
}

// warte wartet auf `start: bereit` oder eine Zeile eines Verstoßes.
func (l startLauf) warte() {
	wartetBis(func() bool {
		return strings.Contains(l.stderr.String(), "start: bereit") || strings.Contains(l.stderr.String(), "ODR-")
	})
}

// beende wartet, bricht den Start ab und liefert seine Rückgabe.
func (l startLauf) beende() int {
	l.warte()
	l.abbruch()
	return <-l.ergebnis
}

// klartext sind die Einträge des Wegs Klartext über die Umgebung.
func klartext() []string { return []string{"ADMIN_LISTEN=0.0.0.0:8081", "ADMIN_ALLOW_PLAINTEXT=true"} }

// TestAdminWarnungStart (C06, C05 am Start nach SPEC-300 Nr. 13 d): 3 Wege × 4 Modi = 12 Zellen über Start; das Flag steht in jedem
// Weg auf true. Im Weg Klartext steht genau die eine Zeile der Warnung vor `start: bereit`, sonst keine; der Admin-Listener spricht
// im Weg TLS das Paar der Umgebung und sonst Klartext.
func TestAdminWarnungStart(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	wege := map[bootstrap.AdminWeg][]string{
		bootstrap.WegTLS:      append(klartext(), "ADMIN_TLS_CERT="+cert, "ADMIN_TLS_KEY="+key),
		bootstrap.WegLoopback: {"ADMIN_ALLOW_PLAINTEXT=true"},
		bootstrap.WegKlartext: klartext(),
	}
	zellen := 0
	for _, modus := range vierModi() {
		for weg, env := range wege {
			zellen++
			l := starte(adminEnv(modus, filepath.Join(t.TempDir(), "d"), env...))
			l.warte()
			select {
			case adresse := <-l.admin:
				pruefeTransportAm(t, fmt.Sprintf("%s Weg %d", modus, weg), adresse, weg == bootstrap.WegTLS, w.wurzel)
			default:
				t.Errorf("%s Weg %d: kein Admin-Listener: %s", modus, weg, l.stderr.String())
			}
			code := l.beende()
			var soll []string
			if weg == bootstrap.WegKlartext {
				soll = []string{zeileWarnung}
			}
			aus := l.stderr.String()
			spaet := soll != nil && strings.Index(aus, "ohne TLS") > strings.Index(aus, "start: bereit")
			if code != 0 || !slices.Equal(warnungen(aus), soll) || spaet {
				t.Errorf("%s Weg %d: Rückgabe %d, erwartet %q vor `start: bereit`: %s", modus, weg, code, soll, aus)
			}
		}
	}
	if zellen != 12 {
		t.Errorf("%d Zellen, erwartet 12", zellen)
	}
}

// TestAdminWarnungFolge (C06): im Weg Klartext × 4 Modi je {Hochfahren scheitert, Build scheitert, ctx nach dem Laden beendet, zwei
// Starts nacheinander} = 16 Zellen. Die Warnung steht je Start einmal nach einem gelungenen Build, sonst nicht.
func TestAdminWarnungFolge(t *testing.T) {
	zellen := 0
	for _, modus := range vierModi() {
		umgebung := func(extra ...string) []string {
			return adminEnv(modus, filepath.Join(t.TempDir(), "d"), append(klartext(), extra...)...)
		}
		erwarte := func(fall string, code, soll int, stderr string, zeilen ...string) {
			zellen++
			if code != soll || !slices.Equal(warnungen(stderr), zeilen) {
				t.Errorf("%s %s: Rückgabe %d, erwartet %d und %q: %s", modus, fall, code, soll, zeilen, stderr)
			}
		}
		l := starte(umgebung("ADMIN_TOKEN_FILE=" + t.TempDir()))
		code := <-l.ergebnis
		l.abbruch()
		erwarte("Hochfahren scheitert", code, 1, l.stderr.String(), zeileWarnung)

		h := neuHaken()
		h.fehlerIn = map[string]error{pAblage: errPhase}
		stderr := &syncPuffer{}
		code = bootstrap.StartMit(context.Background(), umgebung(), stderr, h.bauteile())
		erwarte("Build scheitert", code, 1, stderr.String())

		beendet, cancel := context.WithCancel(context.Background())
		cancel()
		binds, oeffnen := 0, 0
		stderr = &syncPuffer{}
		code = bootstrap.StartMit(beendet, umgebung(), stderr, zaehlendeBauteile(&binds, &oeffnen))
		erwarte("ctx beendet", code, 0, stderr.String())
		if binds != 0 || oeffnen != 0 || stderr.String() != "" {
			t.Errorf("%s ctx beendet: Ablage %d, Binds %d: %s", modus, oeffnen, binds, stderr.String())
		}

		env, summe, alle := umgebung(), 0, ""
		for range 2 {
			l := starte(env)
			summe += l.beende()
			alle += l.stderr.String()
		}
		erwarte("zwei Starts", summe, 0, alle, zeileWarnung, zeileWarnung)
	}
	if zellen != 16 {
		t.Errorf("%d Zellen, erwartet 16", zellen)
	}
}

// pemZeile ist die erste Zeile zwischen den Rändern eines PEM-Blocks, ein Teilstück des Inhalts.
func pemZeile(inhalt []byte) string {
	block, _ := pem.Decode(inhalt)
	return strings.Split(string(pem.EncodeToMemory(block)), "\n")[1]
}

// TestAdminTransportKeinGeheimnis (SPEC-299 Nr. 12, SPEC-300 Nr. 12): 5 Anlässe {Weg TLS, Weg Klartext, Zusammenhang, Paar fehlt,
// Schlüssel passt nicht} × 4 Modi = 20 Starts. Weder Pfad, Inhalt, Eigenschaft des Zertifikats, Wert von ADMIN_LISTEN noch Token
// stehen auf stderr des Starts, auf os.Stdout, os.Stderr, in log oder slog; je Quelle die Kontrolle, dass die Erfassung sie findet.
// Ein Schreiben auf die Deskriptoren 1 und 2 an os.Stdout und os.Stderr vorbei sieht er nicht.
func TestAdminTransportKeinGeheimnis(t *testing.T) {
	w := neueWerkstatt(t) // Verzeichnis `geheim-pfad-4711`, Subject `geheim-cn-4711`
	const host = "geheim-host-4711.example"
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	anlaesse := [][]string{
		{"ADMIN_LISTEN=" + host + ":8081", "ADMIN_TLS_CERT=" + cert, "ADMIN_TLS_KEY=" + key},
		{"ADMIN_LISTEN=" + host + ":8081", "ADMIN_ALLOW_PLAINTEXT=true"},
		{"ADMIN_LISTEN=" + host + ":8081"},
		{"ADMIN_TLS_CERT=" + w.fehlt(), "ADMIN_TLS_KEY=" + w.fehlt()},
		{"ADMIN_TLS_CERT=" + cert, "ADMIN_TLS_KEY=" + w.datei(w.schluessel(w.fremd))},
	}
	verboten := []string{
		"geheim-pfad-4711", "geheim-cn-4711", host, w.gut.Zertifikat.SerialNumber.String(),
		pemZeile(w.schluessel(w.gut)), pemZeile(w.schluessel(w.fremd)), pemZeile(w.gut.ZertifikatPEM()),
	}
	slogVorher, logVorher, flags := slog.Default(), log.Writer(), log.Flags()
	t.Cleanup(func() {
		slog.SetDefault(slogVorher)
		log.SetOutput(logVorher)
		log.SetFlags(flags)
	})
	var standard syncPuffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&standard, nil)))
	zellen, tokens := 0, 0
	for _, modus := range vierModi() {
		for i, env := range anlaesse {
			zellen++
			daten := filepath.Join(t.TempDir(), "d")
			var aus string
			prozess := ausgabeErfasst(t, func() {
				l := starte(adminEnv(modus, daten, env...))
				l.beende()
				aus = l.stderr.String()
			})
			gesucht := slices.Clone(verboten)
			if token, err := os.ReadFile(filepath.Join(daten, adminTokenStandard)); err == nil {
				tokens++
				text := strings.TrimSpace(string(token))
				gesucht = append(gesucht, text[:len(text)/2], text[len(text)/2:])
			}
			alle := aus + prozess + standard.String()
			for _, g := range gesucht {
				if aus == "" || strings.Contains(alle, g) {
					t.Errorf("%s Anlass %d: %q steht in einer Ausgabe, oder stderr ist leer: %s", modus, i, g, alle)
				}
			}
		}
	}
	prozess := ausgabeErfasst(t, func() {
		_, _ = fmt.Fprint(os.Stdout, "quelle-stdout ")
		_, _ = fmt.Fprint(os.Stderr, "quelle-os.Stderr ")
		log.Print("quelle-log")
		slog.Info("quelle-slog")
	})
	for _, quelle := range []string{"quelle-stdout", "quelle-os.Stderr", "quelle-log", "quelle-slog"} {
		if !strings.Contains(prozess+standard.String(), quelle) {
			t.Errorf("Kontrolle %s: die Erfassung findet die Quelle nicht", quelle)
		}
	}
	if zellen != 20 || tokens != 8 {
		t.Errorf("%d Zellen und %d Starts mit Token-Datei, erwartet 20 und 8", zellen, tokens)
	}
}

// pruefeTransportAm fragt GET /healthz und GET /__admin/sessions ohne Token am Admin-Listener: über TLS ist der Klient mit der Wurzel
// allein erfolgreich (TLS 1.3), im Klartext allein der Klartext-Klient.
func pruefeTransportAm(t *testing.T, name, adresse string, tlsWeg bool, wurzel *tlstest.Aussteller) {
	t.Helper()
	klar := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
	mitTLS := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{DisableKeepAlives: true, TLSClientConfig: wurzel.Klient("localhost")}}
	gut, schema, schlecht, anderes := klar, "http://", mitTLS, "https://"
	if tlsWeg {
		gut, schema, schlecht, anderes = mitTLS, "https://", klar, "http://"
	}
	resp, err := gut.Get(schema + adresse + "/healthz")
	if err != nil {
		t.Errorf("%s: %v", name, err)
		return
	}
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || body.String() != `{"status":"ok"}` || tlsWeg && (resp.TLS == nil || resp.TLS.Version != 0x0304) {
		t.Errorf("%s: healthz %d %q", name, resp.StatusCode, body.String())
	}
	if status, _, _ := sende(gut, "GET", schema+adresse+"/__admin/sessions", "", ""); status != 401 {
		t.Errorf("%s: ohne Token %d, erwartet 401", name, status)
	}
	if status, _, _ := sende(schlecht, "GET", anderes+adresse+"/healthz", "", ""); status == 200 {
		t.Errorf("%s: der Klient des anderen Transports erhält 200", name)
	}
}

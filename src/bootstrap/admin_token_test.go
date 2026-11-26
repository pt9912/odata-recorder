package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func schreibeDatei(t *testing.T, pfad, inhalt string, rechte os.FileMode) {
	t.Helper()
	if err := os.WriteFile(pfad, []byte(inhalt), rechte); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(pfad, rechte); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, pfad string) string {
	t.Helper()
	if err := os.MkdirAll(pfad, 0o700); err != nil {
		t.Fatal(err)
	}
	return pfad
}

// merkeToken setzt NeuerAdmin so, dass token den Wert der Konfig hält.
func merkeToken(token *[]byte) func(*bootstrap.Bauteile) {
	return func(b *bootstrap.Bauteile) {
		neu := b.NeuerAdmin
		b.NeuerAdmin = func(k admin.Konfig) (bootstrap.Server, error) {
			*token = slices.Clone(k.Token)
			return neu(k)
		}
	}
}

// TestAdminToken (B04): 4 Modi × {Datei fehlt, 32 Zeichen, Symlink} = 12, der zweite Start je Modus = 4, der Start über die
// Umgebung je Modus = 4.
func TestAdminToken(t *testing.T) {
	erzeugt := regexp.MustCompile(`^[0-9a-f]{64}\n$`)
	zellen := 0
	hoch := func(k bootstrap.Konfiguration) []byte {
		var token []byte
		h, _ := adminBau(t, k, nil, merkeToken(&token))
		if err := h.Hochfahren(context.Background()); err != nil {
			t.Fatalf("%s: Hochfahren: %v", k.Modus, err)
		}
		if err := h.Herunterfahren(context.Background()); err != nil {
			t.Fatalf("%s: Herunterfahren: %v", k.Modus, err)
		}
		return token
	}
	for _, modus := range vierModi() {
		for _, art := range []string{"fehlt", "32 Zeichen", "Symlink"} {
			zellen++
			k := listenerKonfig(t.TempDir(), modus)
			switch art {
			case "32 Zeichen":
				schreibeDatei(t, k.AdminTokenFile, strings.Repeat("k", 32), 0o600)
			case "Symlink":
				ziel := filepath.Join(t.TempDir(), "geheimnis")
				schreibeDatei(t, ziel, "  "+strings.Repeat("s", 40)+"\n", 0o400)
				if err := os.Symlink(ziel, k.AdminTokenFile); err != nil {
					t.Fatal(err)
				}
			}
			token := hoch(k)
			inhalt, err := os.ReadFile(k.AdminTokenFile)
			if err != nil || string(token) != strings.TrimSpace(string(inhalt)) || len(token) < 32 {
				t.Errorf("%s %s: Token %q, Datei %q (%v)", modus, art, token, inhalt, err)
			}
			if art != "fehlt" {
				continue
			}
			if fi, err := os.Stat(k.AdminTokenFile); err != nil || fi.Mode().Perm() != 0o600 || !erzeugt.Match(inhalt) {
				t.Errorf("%s: erzeugte Datei %q (%v)", modus, inhalt, err)
			}
			zellen++
			zweites := hoch(k)
			danach, _ := os.ReadFile(k.AdminTokenFile)
			if string(zweites) != string(token) || string(danach) != string(inhalt) {
				t.Errorf("%s: zweiter Start: Token %q, Datei %q", modus, zweites, danach)
			}
		}
		zellen++
		startUeberUmgebung(t, modus)
	}
	if zellen != 20 {
		t.Errorf("%d Zellen, erwartet 20", zellen)
	}
}

// startUeberUmgebung fährt Start mit ADMIN_LISTEN und ADMIN_TOKEN_FILE ungleich den Defaults: der Haken Listen sieht die
// Adresse, der Haken NeuerAdmin das Token der genannten Datei.
func startUeberUmgebung(t *testing.T, modus string) {
	t.Helper()
	dir := t.TempDir()
	datei := filepath.Join(dir, "mein.token")
	schreibeDatei(t, datei, strings.Repeat("u", 48), 0o600)
	var adressen []string
	var token []byte
	b := bootstrap.StandardBauteile()
	merkeToken(&token)(&b)
	echt := b.Listen
	b.Listen = func(ctx context.Context, network, addr string) (net.Listener, error) {
		adressen = append(adressen, addr)
		return echt(ctx, network, "127.0.0.1:0")
	}
	env := adminEnv(modus, filepath.Join(dir, "d"), "LISTEN=127.0.0.1:8080", "PUBLIC_BASE_URL=", "ADMIN_LISTEN=localhost:7001", "ADMIN_TOKEN_FILE="+datei)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	puffer := &syncPuffer{}
	ergebnis := make(chan int, 1)
	go func() { ergebnis <- bootstrap.StartMit(ctx, env, puffer, b) }()
	ende := time.Now().Add(20 * time.Second)
	for !strings.Contains(puffer.String(), "start: bereit") && time.Now().Before(ende) {
		time.Sleep(2 * time.Millisecond)
	}
	cancel()
	if code := <-ergebnis; code != 0 {
		t.Errorf("%s: Start %d, stderr %s", modus, code, puffer.String())
	}
	if !slices.Equal(adressen, []string{"127.0.0.1:8080", "localhost:7001"}) || string(token) != strings.Repeat("u", 48) {
		t.Errorf("%s: Listen %q, Token %q", modus, adressen, token)
	}
}

// TestAdminTokenVerstoss (B05): (a) 7 Gründe, (b) Dateien des Dienstes, (c) Gegenseite, (d) Vorrang, (e) leerer Pfad.
func TestAdminTokenVerstoss(t *testing.T) {
	gruende := []struct {
		grund   string
		bereite func(dir string) string
	}{
		{"Datei nicht lesbar", func(dir string) string { schreibeDatei(t, dir+"/x", "x", 0o600); return dir + "/x/tok" }},
		{"ist ein Verzeichnis", func(dir string) string { return mkdir(t, dir+"/tok") }},
		{"keine reguläre Datei", func(dir string) string {
			if err := syscall.Mkfifo(dir+"/tok", 0o600); err != nil {
				t.Fatal(err)
			}
			return dir + "/tok"
		}},
		{"Gruppen- oder Weltrechte", func(dir string) string {
			schreibeDatei(t, dir+"/tok", strings.Repeat("a", 64), 0o644)
			return dir + "/tok"
		}},
		{"mehr als eine Zeile", func(dir string) string {
			schreibeDatei(t, dir+"/tok", strings.Repeat("a", 40)+"\nb\n", 0o600)
			return dir + "/tok"
		}},
		{"weniger als 32 Zeichen", func(dir string) string {
			schreibeDatei(t, dir+"/tok", strings.Repeat("a", 31), 0o600)
			return dir + "/tok"
		}},
		{"Datei nicht anlegbar", func(dir string) string { return dir + "/fehlt/tok" }},
	}
	dienstNamen := []string{"audit.jsonl", "odata-recorder.db", "odata-recorder.db-wal", "odata-recorder.db-journal", "odata-recorder.dbx"}
	formen := []func(daten string) string{
		func(daten string) string { return daten },
		func(daten string) string {
			return filepath.Dir(daten) + "/./" + filepath.Base(daten) + "/../" + filepath.Base(daten)
		},
		func(daten string) string {
			verweis := filepath.Join(t.TempDir(), "verweis")
			if err := os.Symlink(daten, verweis); err != nil {
				t.Fatal(err)
			}
			return verweis
		},
	}
	gegenseite := []func(daten string) string{
		func(string) string { return t.TempDir() + "/audit.jsonl" },
		func(daten string) string { return mkdir(t, daten+"/unter") + "/audit.jsonl" },
		func(daten string) string { return daten + "/audit.jsonlx" },
		func(daten string) string { return daten + "/xodata-recorder.db" },
		func(daten string) string { return daten + "/odata-recorder.d" },
	}
	zellen := 0
	for _, modus := range vierModi() {
		konfiguriere := func(pfad func(daten string) string) bootstrap.Konfiguration {
			k := listenerKonfig(mkdir(t, filepath.Join(t.TempDir(), "CANARY-pfad")), modus)
			k.AdminTokenFile = pfad(k.DataDir)
			return k
		}
		zelle := func(name string, pfad func(daten string) string, pruefe func(t *testing.T, k bootstrap.Konfiguration)) {
			zellen++
			t.Run(modus+"/"+name, func(t *testing.T) { pruefe(t, konfiguriere(pfad)) })
		}
		fehler := func(zeile string) func(t *testing.T, k bootstrap.Konfiguration) {
			return func(t *testing.T, k bootstrap.Konfiguration) { erwarteTokenFehler(t, k, zeile) }
		}
		const dienst = "ODR-E2001 ADMIN_TOKEN_FILE: Datei, die der Dienst selbst schreibt"
		for _, g := range gruende {
			zelle("a "+g.grund, func(string) string { return g.bereite(mkdir(t, filepath.Join(t.TempDir(), "CANARY-tok"))) },
				fehler("ODR-E2001 ADMIN_TOKEN_FILE: "+g.grund))
		}
		for _, name := range dienstNamen {
			for i, form := range formen {
				zelle(fmt.Sprintf("b %s Form %d", name, i), func(daten string) string { return form(daten) + "/" + name }, fehler(dienst))
			}
		}
		for i, pfad := range gegenseite {
			zelle(fmt.Sprintf("c %d", i), pfad, erwarteErzeugt)
		}
		zelle("d", func(daten string) string {
			schreibeDatei(t, daten+"/audit.jsonl", "a\nb\n", 0o644)
			return daten + "/audit.jsonl"
		}, fehler(dienst))
		zelle("e", func(string) string { return "" }, fehler(""))
	}
	if zellen != 116 {
		t.Errorf("%d Zellen, erwartet 116", zellen)
	}
}

// erwarteErzeugt: Hochfahren gelingt und legt die Token-Datei an.
func erwarteErzeugt(t *testing.T, k bootstrap.Konfiguration) {
	t.Helper()
	hd, err := bootstrap.BuildMit(context.Background(), k, nil, neuHaken().bauteile())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	t.Cleanup(func() { _ = hd.Herunterfahren(context.Background()) })
	err = hd.Hochfahren(context.Background())
	if inhalt, lerr := os.ReadFile(k.AdminTokenFile); err != nil || lerr != nil || len(inhalt) != 65 {
		t.Errorf("%s: Hochfahren %v, Token-Datei %q (%v)", k.AdminTokenFile, err, inhalt, lerr)
	}
}

// erwarteTokenFehler: Hochfahren endet am ersten Schritt mit genau der Zeile (leer: model.ErrUngueltig ohne KonfigFehler), ohne
// Listen, Startwartung, Service, Session und neues Audit-Log, und die benannte Datei bleibt wie davor.
func erwarteTokenFehler(t *testing.T, k bootstrap.Konfiguration, zeile string) {
	t.Helper()
	h := neuHaken()
	hd, err := bootstrap.BuildMit(context.Background(), k, nil, h.bauteile())
	if err != nil {
		t.Fatalf("%s: Build: %v", k.Modus, err)
	}
	t.Cleanup(func() { _ = hd.Herunterfahren(context.Background()) })
	davor, davorErr := leseRegulaer(k.AdminTokenFile)
	audit := filepath.Join(k.DataDir, "audit.jsonl")
	_, ohneAudit := os.Stat(audit)
	err = hd.Hochfahren(context.Background())
	var kf *bootstrap.KonfigFehler
	switch {
	case zeile == "" && (!errors.Is(err, model.ErrUngueltig) || errors.As(err, &kf)):
		t.Errorf("%s: %v, erwartet model.ErrUngueltig", k.Modus, err)
	case zeile != "" && (!errors.As(err, &kf) || !slices.Equal(kf.Zeilen, []string{zeile}) || strings.Contains(err.Error(), "CANARY")):
		t.Errorf("%s %s: %v, erwartet %q", k.Modus, k.AdminTokenFile, err, zeile)
	}
	if n := h.p.zaehle(pListen) + h.p.zaehle(pWartung) + h.dienstZaehler("Create") + h.p.zaehle("session:x1"); n != 0 {
		t.Errorf("%s %s: %d Aufrufe nach dem Schritt der Token-Datei", k.Modus, k.AdminTokenFile, n)
	}
	danach, danachErr := leseRegulaer(k.AdminTokenFile)
	if _, err := os.Stat(audit); errors.Is(ohneAudit, os.ErrNotExist) && !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%s %s: audit.jsonl ist entstanden", k.Modus, k.AdminTokenFile)
	}
	if davor != danach || (davorErr == nil) != (danachErr == nil) {
		t.Errorf("%s %s: die Datei ist angelegt oder geändert", k.Modus, k.AdminTokenFile)
	}
}

// leseRegulaer liest eine reguläre Datei; ein FIFO wird nie geöffnet.
func leseRegulaer(pfad string) (string, error) {
	fi, err := os.Stat(pfad)
	if err != nil || !fi.Mode().IsRegular() {
		return "", err
	}
	b, err := os.ReadFile(pfad)
	return string(b), err
}

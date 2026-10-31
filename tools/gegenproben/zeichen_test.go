package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// Die 25 Fälle 131 bis 155 tragen die Namen aus SPEC-102, jeder genau einmal und in dieser Folge.
func TestZeichenFaelleHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []string{"sauber", "menge-z", "bom-anfang-und-mitte", "ausgabeform", "cr-und-zeilenende", "zeilen-trenner-brechen-nicht",
		"ungueltiges-utf8", "utf16-bom", "binaer-grenze", "baseline-ausgenommen", "symlinks-nicht-gefolgt", "leere-datei",
		"leere-menge", "wurzel-fehlt", "aufruf", "dateiarten", "reihenfolge-und-wiederholung", "grosse-zeile", "paare",
		"pfad-sonderzeichen", "treiber-menge", "treiber-aufrufe-von-docker", "treiber-exit", "treiber-aufraeumen",
		"treiber-ohne-repo-und-ohne-dateien"}
	fs := zeichenFaelle()
	if len(want) != 25 || len(fs) != 25 {
		t.Fatalf("%d erwartete Namen und %d Fälle, erwartet 25", len(want), len(fs))
	}
	for i, n := range want {
		if fs[i].Name != "zeichen/"+n {
			t.Errorf("Fall %d heißt %s statt zeichen/%s", 131+i, fs[i].Name, n)
		}
		if i < 20 && fs[i].Gate != gateZeichen || i >= 20 && fs[i].Gate != gateZeichenTreiber {
			t.Errorf("Fall %d: Gate %d", 131+i, fs[i].Gate)
		}
		if !fs[i].Getrennt {
			t.Errorf("Fall %d liest die Ströme nicht getrennt (Randform 13)", 131+i)
		}
	}
}

// Das Orakel der Menge Z in den Gegenproben hat 253 Elemente in den Teilmengen 63, 170, 1, 1, 16 und 2.
func TestZeichenOrakelHat253Elemente(t *testing.T) {
	zahl := map[string]int{}
	for _, b := range zcZ() {
		zahl[b.kat] += int(b.bis-b.von) + 1
	}
	for kat, n := range map[string]int{"Cc": 63, "Cf": 170, "Zl": 1, "Zp": 1, "Zs": 16, "Cn": 2} {
		if zahl[kat] != n {
			t.Errorf("Kategorie %s: %d Zeichen statt %d", kat, zahl[kat], n)
		}
	}
	if len(zcKurznamen()) != 29 {
		t.Errorf("%d Kurznamen statt 29", len(zcKurznamen()))
	}
}

// Die Programmliste nennt zeichen und tar neben den übrigen (SPEC-102 Randform 2), und die Fälle
// rufen das Programm und den Treiber im Verzeichnis des Falls.
func TestStandardNenntZeichenUndTar(t *testing.T) {
	u := standard()
	have := map[string]bool{}
	for _, p := range u.Programme {
		have[p] = true
	}
	for _, n := range []string{"zeichen", "git", "tar"} {
		if !have[n] {
			t.Errorf("Programm %s fehlt in standard().Programme", n)
		}
	}
	if !filepath.IsAbs(u.ZeichenTreiber) || !strings.HasSuffix(u.ZeichenTreiber, "tools/harness/zeichen.sh") {
		t.Errorf("ZeichenTreiber: %q", u.ZeichenTreiber)
	}
	name, args, wd := u.Befehl("/k", Fall{Gate: gateZeichen, Argumente: []string{"pruefen", "{dir}"}})
	if name != "zeichen" || strings.Join(args, "|") != "pruefen|/k" || wd != "/k" {
		t.Errorf("Programm: %q %v %q", name, args, wd)
	}
	name, args, wd = u.Befehl("/k", Fall{Gate: gateZeichenTreiber})
	if name != "bash" || len(args) != 1 || args[0] != u.ZeichenTreiber || wd != "/k" {
		t.Errorf("Treiber: %q %v %q", name, args, wd)
	}
}

// inZ ist wahr, wenn r nach dem Orakel zu Z gehört.
func inZ(r rune) bool {
	for _, b := range zcZ() {
		if r >= b.von && r <= b.bis {
			return true
		}
	}
	return false
}

// Die Fälle zu zeichen führen Zeichen von Z nur als Bytewerte (SPEC-233 Nr. 14): kein Rohzeichen
// von Z und kein ungültiges UTF-8 in den Quelltexten der Gruppe, keine Backslash-u-Folge als Text
// in einem Fixture.
func TestZeichenFaelleTragenNurBytewerte(t *testing.T) {
	for _, f := range zeichenFaelle() {
		for datei, inhalt := range f.Neu {
			for _, s := range []string{datei, inhalt} {
				if strings.Contains(s, "\\u") {
					t.Errorf("%s: %s trägt eine Escape-Folge als Text", f.Name, datei)
				}
			}
		}
	}
	for _, d := range []string{"zeichen.go", "zeichen_treiber.go", "zeichen_test.go"} {
		b, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		if !utf8.Valid(b) {
			t.Errorf("%s: kein gültiges UTF-8", d)
		}
		for i, r := range string(b) {
			if inZ(r) {
				t.Errorf("%s: Zeichen U+%04X an Stelle %d", d, r, i)
			}
		}
	}
}

// ersatzDocker ruft das Ersatz-Skript für docker mit dem Protokollverzeichnis log.
func ersatzDocker(t *testing.T, log string, env []string, args ...string) (int, string, string) {
	t.Helper()
	skript := filepath.Join(t.TempDir(), "docker")
	if err := os.WriteFile(skript, []byte(zcErsatzDocker), 0o755); err != nil {
		t.Fatal(err)
	}
	code, out, errOut, _ := starteStroeme(10*time.Second, log, append([]string{"ZC_LOG=" + log}, env...), true, "sh", append([]string{skript}, args...)...)
	return code, out, errOut
}

// Das Ersatz-Skript hält jeden Aufruf mit allen Argumenten fest, auch den unbekannten, und antwortet
// nach den Werten der Umgebung (SPEC-102 Randform 12).
func TestErsatzDockerProtokolliertJedenAufruf(t *testing.T) {
	log := t.TempDir()
	if code, out, _ := ersatzDocker(t, log, nil, "create", "--network", "none", "i:t", "pruefen", "/work"); code != 0 || out != "zc-kennung\n" {
		t.Errorf("create: Exit %d, stdout %q", code, out)
	}
	if code, out, _ := ersatzDocker(t, log, []string{"ZC_CREATE_RC=7"}, "create"); code != 7 || out != "" {
		t.Errorf("create mit Fehler: Exit %d, stdout %q", code, out)
	}
	if code, out, errOut := ersatzDocker(t, log, []string{"ZC_START_RC=3"}, "start", "-a", "zc-kennung"); code != 3 || out != "start-stdout\n" || errOut != "start-stderr\n" {
		t.Errorf("start: Exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	if code, _, _ := ersatzDocker(t, log, nil, "rm", "-f", "zc-kennung"); code != 0 {
		t.Errorf("rm: Exit %d", code)
	}
	if code, _, _ := ersatzDocker(t, log, nil, "run", "-v", "x:y"); code != 99 {
		t.Errorf("run: Exit %d statt 99", code)
	}
	b, err := os.ReadFile(filepath.Join(log, "docker.log"))
	if err != nil {
		t.Fatal(err)
	}
	soll := "create --network none i:t pruefen /work\ncreate\nstart -a zc-kennung\nrm -f zc-kennung\nrun -v x:y\n"
	if string(b) != soll {
		t.Errorf("Protokoll %q statt %q", b, soll)
	}
}

// Ein Treiber-Fall ohne Protokoll des Ersatzes ist rot: fehlt der Tar-Strom, meldet zcStrom den Grund.
func TestStromFehltIstEinGrund(t *testing.T) {
	if _, err := zcStrom(t.TempDir()); err == nil {
		t.Error("Tar-Strom fehlt, aber kein Fehler")
	}
	if a := zcAufrufe(t.TempDir()); a != nil {
		t.Errorf("Aufrufe ohne Protokoll: %q", a)
	}
}

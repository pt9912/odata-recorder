package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Die Pfade stehen hier unabhängig vom Programm.
const (
	dFehler    = "src/hexagon/model/fehler.go"
	dSpez      = "spec/spezifikation.md"
	dAusnahmen = "harness/katalog-ausnahmen.txt"
	dHandbuch  = "docs/user/benutzerhandbuch.md"
)

// codeListe sind die 40 Codes der Fabrik: ODR-E<t><h>01 mit t aus 2 bis 9 und h aus 0 bis 4, aufsteigend.
func codeListe() []string {
	var out []string
	for t := 2; t <= 9; t++ {
		for h := 0; h <= 4; h++ {
			out = append(out, fmt.Sprintf("ODR-E%d%d01", t, h))
		}
	}
	return out
}

func mengeVon(l []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range l {
		m[s] = true
	}
	return m
}

func ohne(l []string, weg ...string) []string {
	w := mengeVon(weg)
	var out []string
	for _, s := range l {
		if !w[s] {
			out = append(out, s)
		}
	}
	return out
}

// fehlerGo schreibt je Code ein Literal in Anführungszeichen, ab Zeile 4.
func fehlerGo(codes []string) string {
	var b strings.Builder
	b.WriteString("package model\n\nvar tabelle = []zeile{\n")
	for _, c := range codes {
		fmt.Fprintf(&b, "\t{%q, 1},\n", c)
	}
	b.WriteString("}\n")
	return b.String()
}

const kopfK = "| ID | Code | Bedingung | Aktion |\n|---|---|---|---|\n"

func zeilenK(codes []string) string {
	var b strings.Builder
	for i, c := range codes {
		fmt.Fprintf(&b, "| `SPEC-%03d` | `%s` | Bedingung | Aktion |\n", i+1, c)
	}
	return b.String()
}

func spezText(codes []string) string {
	return "# Spezifikation\n\n## 3. Davor\n\nText.\n\n## 4. Fehler-Codes und Logging-Felder\n\n" +
		kopfK + zeilenK(codes) + "\nText danach.\n\n## 5. Metriken\n\nText.\n"
}

func ausnahmenText(codes []string) string {
	var b strings.Builder
	for _, c := range codes {
		fmt.Fprintf(&b, "%s | Grund\n", c)
	}
	return b.String()
}

const kopfH = "| Code | Bedeutung |\n|---|---|\n"

func zeilenH(codes []string) string {
	var b strings.Builder
	for _, c := range codes {
		fmt.Fprintf(&b, "| `%s` | Die Bedeutung. |\n", c)
	}
	return b.String()
}

func handbuchText(codes []string) string {
	return "# Benutzerhandbuch\n\nSoftware-Version: dev\n\n## Fehlercodes\n\nText.\n\n" + kopfH + zeilenH(codes) + "\n## Ende\n\nText.\n"
}

// grund ist ein grüner Baum: alle Codes in fehler.go und in §4, die Ausnahmen mit Grund, jeder Code aus H mit Zeile.
func grund(codes, ausnahmen []string) map[string]string {
	return map[string]string{
		dFehler:    fehlerGo(codes),
		dSpez:      spezText(codes),
		dAusnahmen: ausnahmenText(ausnahmen),
		dHandbuch:  handbuchText(ohne(codes, ausnahmen...)),
	}
}

func schreibe(t *testing.T, m map[string]string) string {
	t.Helper()
	w := t.TempDir()
	for p, inhalt := range m {
		f := filepath.Join(w, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f, []byte(inhalt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return w
}

func fahre(args ...string) (int, string, string) {
	var o, e bytes.Buffer
	c := run(args, &o, &e)
	return c, o.String(), e.String()
}

// zeileVon ist die Zeile (ab 1), in der text zum ersten Mal die Nadel trägt.
func zeileVon(t *testing.T, text, nadel string) int {
	t.Helper()
	for i, z := range strings.Split(text, "\n") {
		if strings.Contains(z, nadel) {
			return i + 1
		}
	}
	t.Fatalf("%q nicht im Text", nadel)
	return 0
}

func befunde(texte ...string) string {
	var b strings.Builder
	for _, x := range texte {
		b.WriteString("katalog: Fehlercodes: " + x + "\n")
	}
	return b.String()
}

func eingaben(texte ...string) string {
	var b strings.Builder
	for _, x := range texte {
		b.WriteString("katalog: Eingabe: " + x + "\n")
	}
	return b.String()
}

// erwarte fährt das Programm auf dem Baum und vergleicht Ausgangswert, stdout und stderr.
func erwarte(t *testing.T, name string, m map[string]string, code int, out, err string) {
	t.Helper()
	w := schreibe(t, m)
	c, o, e := fahre(w)
	if c != code || o != out || e != err {
		t.Errorf("%s: Exit %d stdout %q stderr %q; erwartet Exit %d stdout %q stderr %q", name, c, o, e, code, out, err)
	}
}

func erwarteOk(t *testing.T, name string, m map[string]string) {
	t.Helper()
	erwarte(t, name, m, 0, "katalog: ok\n", "")
}

func erwarteBefunde(t *testing.T, name string, m map[string]string, texte ...string) {
	t.Helper()
	erwarte(t, name, m, 1, "", befunde(texte...))
}

func erwarteEingabe(t *testing.T, name string, m map[string]string, texte ...string) {
	t.Helper()
	erwarte(t, name, m, 2, "", eingaben(texte...))
}

func zaehle(t *testing.T, was string, ist, soll int) {
	t.Helper()
	if ist != soll {
		t.Fatalf("%s: %d Elemente statt %d", was, ist, soll)
	}
}

func quote(x string) string { return fmt.Sprintf("%q", x) }

// verzeichnisBaum ist der Inhalt eines Baums: Pfad, Größe und Inhalt jeder Datei, geordnet.
func verzeichnisBaum(t *testing.T, w string) string {
	t.Helper()
	var zeilen []string
	err := filepath.Walk(w, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(w, p)
		if i.IsDir() {
			zeilen = append(zeilen, "d "+rel)
			return nil
		}
		roh, err := os.ReadFile(p)
		zeilen = append(zeilen, fmt.Sprintf("f %s %d %x", rel, i.Size(), roh))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(zeilen)
	return strings.Join(zeilen, "\n")
}

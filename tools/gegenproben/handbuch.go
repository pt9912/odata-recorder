package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Die Fälle H1 bis H10 von `pruefung handbuch` (SPEC-200, Gegenprobe) legen je ein Wegwerfverzeichnis
// mit Handbuch, Beispiel-Datei, Makefile, Dockerfile und einem Ersatz-Binary an. Das Ersatz-Binary
// ist ein Skript, das jeden Start in log/binary.log zählt, auf `version` mit der Version dev und sonst
// mit seiner Ausgabe antwortet.

const (
	hbHandbuch = "docs/user/benutzerhandbuch.md"
	hbBeispiel = "docs/user/beispiele/aufruf.txt"
	hbBinary   = "bin/odata-recorder"
	hbLog      = "log/binary.log"
	hbAusgabe  = "ersatz"
)

// hbBeispielDatei ist die Beispiel-Datei für den Aufruf ohne Argument mit der Standardausgabe stdout.
func hbBeispielDatei(stdout string) string {
	return "$ docker run --rm odata-recorder:build /out/odata-recorder\n--- Standardausgabe\n" + stdout +
		"\n--- Fehlerausgabe\n--- Ausgangswert\n0\n"
}

// hbHandbuchText ist das Handbuch mit der Software-Version version, einem Absatz prosa und dem Block
// block hinter der Bindungszeile.
func hbHandbuchText(version, prosa, block string) string {
	return "# Handbuch\n\nSoftware-Version: " + version + "\nG\xc3\xbcltigkeitsbereich: Ersatz\nStand: 2026-10-30\n\n## Aufruf\n\n" +
		prosa + "\n\nDatei: [beispiele/aufruf.txt](beispiele/aufruf.txt)\n```\n" + block + "```\n"
}

// hbErsatz ist das Ersatz-Binary: es zählt jeden Start und antwortet auf `version` und ohne Argument.
func hbErsatz(stdout string) string {
	return "#!/bin/sh\nd=${0%/*}/..\necho AUFRUF >> \"$d/" + hbLog + "\"\ncase \"$*\" in\n" +
		"version) echo 'odata-recorder dev';;\n'') echo '" + stdout + "';;\nesac\n"
}

// hbStarts prüft, wie oft der Fall das Ersatz-Binary gestartet hat.
func hbStarts(want int) func(Lauf) string {
	return func(l Lauf) string {
		b, err := os.ReadFile(filepath.Join(l.Dir, hbLog))
		if err != nil {
			return "Protokoll des Ersatz-Binary fehlt"
		}
		if n := strings.Count(string(b), "AUFRUF"); n != want {
			return fmt.Sprintf("Ersatz-Binary %d-mal gestartet, erwartet %d", n, want)
		}
		return ""
	}
}

// hbOpt ändert einen Fall nach dem Muster.
type hbOpt func(*hbFall)

type hbFall struct {
	version, prosa, block string
	datei, ersatz         string
	binary                string
}

func hbVersion(v string) hbOpt { return func(f *hbFall) { f.version = v } }

func hbProsa(p string) hbOpt { return func(f *hbFall) { f.prosa = p } }

func hbBlock(b string) hbOpt { return func(f *hbFall) { f.block = b } }

func hbDatei(d string) hbOpt { return func(f *hbFall) { f.datei = d } }

func hbAntwort(a string) hbOpt { return func(f *hbFall) { f.ersatz = a } }

func hbBinaryPfad(p string) hbOpt { return func(f *hbFall) { f.binary = p } }

// hbBauen baut einen Fall: Standard ist das saubere Handbuch H1; die Optionen ändern es. starts ist die Zahl
// der Starts des Ersatz-Binary, die der Fall prüft (negativ: keine Prüfung).
func hbBauen(name string, exit, starts int, opts ...hbOpt) Fall {
	h := hbFall{version: "dev", prosa: "Der Aufruf steht im Beispiel.", datei: hbBeispielDatei(hbAusgabe), ersatz: hbAusgabe, binary: hbBinary}
	for _, o := range opts {
		o(&h)
	}
	if h.block == "" {
		h.block = h.datei
	}
	f := Fall{
		Name: "handbuch/" + name, Gate: gatePruefung, Wegwerfmodul: true, Getrennt: true, WantExit: exit,
		Argumente: []string{"handbuch", "{dir}/" + h.binary, "{dir}"},
		Neu: map[string]string{
			hbHandbuch:   hbHandbuchText(h.version, h.prosa, h.block),
			hbBeispiel:   h.datei,
			"Makefile":   "build: ## bauen\n\tdocker build --target build -t odata-recorder:build .\n",
			"Dockerfile": "RUN go build -o /out/odata-recorder ./x\n",
			hbBinary:     hbErsatz(h.ersatz),
			hbLog:        "",
		},
		Ausfuehrbar: []string{hbBinary},
	}
	if starts >= 0 {
		f.Nachher = hbStarts(starts)
	}
	return f
}

// hbMit setzt die Erwartungen an die Ströme: ein Text mit `1>` davor steht auf stdout, sonst auf stderr.
func hbMit(f Fall, want ...string) Fall {
	for _, w := range want {
		if v, ok := strings.CutPrefix(w, "!"); ok {
			f.Verboten = append(f.Verboten, v)
		} else if v, ok := strings.CutPrefix(w, "1>"); ok {
			f.WantOut = append(f.WantOut, v)
		} else {
			f.WantErr = append(f.WantErr, w)
		}
	}
	return f
}

// handbuchFaelle sind die Fälle H1 bis H10 (SPEC-102, Fälle 121 bis 130).
func handbuchFaelle() []Fall {
	return []Fall{
		hbMit(hbBauen("h1-sauber", 0, 2), "1>pruefung handbuch: ok", "!Planung"),
		hbMit(hbBauen("h2-planung", 1, 2, hbProsa("Das ist geplant.")), "Planung A1: geplant"),
		hbMit(hbBauen("h3-chronik", 1, 2, hbProsa("Das war bisher so.")), "Chronik B1: bisher"),
		hbMit(hbBauen("h4-kennung", 1, 2, hbProsa("Siehe spec-110.")), "Kennung C6: spec-110"),
		hbMit(hbBauen("h5-version-weicht-ab", 1, 2, hbVersion("1.0")),
			"Software-Version 1.0 weicht von der Ausgabe von version ab (dev)"),
		hbMit(hbBauen("h6-binary-weicht-ab", 1, 2, hbAntwort("anders")), "aufruf.txt: Standardausgabe weicht ab", "!Beispiel weicht von"),
		hbMit(hbBauen("h7-block-weicht-ab", 1, 2, hbBlock(hbBeispielDatei(hbAusgabe)+"x\n")),
			"Beispiel weicht von beispiele/aufruf.txt ab"),
		hbMit(hbBauen("h8-make-ziel", 1, 2, hbProsa("Siehe `make nix`.")), "make-Ziel nix unbekannt"),
		hbMit(hbBauen("h9-binary-fehlt", 2, -1, hbBinaryPfad("bin/gibt-es-nicht")), "Binary fehlt", "!Planung"),
		hbMit(hbBauen("h10-zitat", 0, 2, hbDatei(hbBeispielDatei("geplant")), hbAntwort("geplant")),
			"1>pruefung handbuch: ok", "!Planung"),
	}
}

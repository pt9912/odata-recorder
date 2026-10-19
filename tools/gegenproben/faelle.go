package main

import "strings"

// Gate benennt das Programm, das ein Fall in der Kopie aufruft.
type Gate int

const (
	gateACheck Gate = iota
	gateLint
)

// Fall ist eine Gegenprobe: ein Eingriff in die Kopie der Eingaben und die
// Erwartung an das Gate.
type Fall struct {
	Name     string
	Gate     Gate
	Neu      map[string]string // Datei -> Inhalt, angelegt in der Kopie
	Anhang   map[string]string // Datei -> Text, an die Datei der Kopie gehängt
	Ersetze  []Ersetzung
	Entferne []string // Dateien, aus der Kopie gelöscht
	WantExit int
	WantText []string // jeder Text muss in der Ausgabe stehen
}

// Ersetzung tauscht in einer Datei der Kopie genau eine Stelle.
type Ersetzung struct {
	Datei, Alt, Neu string
}

const modul = "github.com/pt9912/odata-recorder/"

func goDatei(paket string, importe ...string) string {
	var b strings.Builder
	b.WriteString("// Package " + paket + " gehört zu einer Gegenprobe.\npackage " + paket + "\n")
	if len(importe) > 0 {
		b.WriteString("\nimport (\n")
		for _, i := range importe {
			b.WriteString("\t_ \"" + modul + i + "\"\n")
		}
		b.WriteString(")\n")
	}
	return b.String()
}

const regelnKopf = "\n    rules:\n"

func faelle() []Fall {
	return []Fall{
		{Name: "a-check/basis", Gate: gateACheck},
		{
			Name: "a-check/kern-auf-driven-adapter", Gate: gateACheck,
			Neu: map[string]string{
				"src/adapters/driven/ziel/ziel.go": goDatei("ziel"),
				"src/hexagon/model/kern/kern.go":   goDatei("kern", "src/adapters/driven/ziel"),
			},
			WantExit: 1, WantText: []string{"core-impurity"},
		},
		{
			Name: "a-check/adapter-auf-adapter", Gate: gateACheck,
			Neu: map[string]string{
				"src/adapters/driven/ziel/ziel.go": goDatei("ziel"),
				"src/adapters/driving/quelle/q.go": goDatei("quelle", "src/adapters/driven/ziel"),
			},
			WantExit: 1, WantText: []string{"lateral-adapter"},
		},
		{
			Name: "a-check/erlaubter-weg-ueber-port", Gate: gateACheck,
			Neu: map[string]string{
				"src/hexagon/model/kern/kern.go":          goDatei("kern"),
				"src/hexagon/ports/outbound/port/port.go": goDatei("port", "src/hexagon/model/kern"),
				"src/adapters/driven/ziel/ziel.go":        goDatei("ziel", "src/hexagon/ports/outbound/port"),
				"src/hexagon/services/dienst/dienst.go":   goDatei("dienst", "src/hexagon/ports/outbound/port", "src/hexagon/model/kern"),
			},
		},
		{Name: "lint/basis", Gate: gateLint, WantText: []string{"0 issues"}},
		{
			Name: "lint/direktive-nolint", Gate: gateLint,
			Anhang:   map[string]string{"src/bootstrap/bootstrap.go": "\n//nolint:unused // Probe\n"},
			WantExit: 1, WantText: []string{"Direktive", "nolintlint"},
		},
		{
			Name: "lint/ausnahme-ohne-why", Gate: gateLint,
			Ersetze:  []Ersetzung{{".golangci.yml", regelnKopf, regelnKopf + "      - linters:\n          - forbidigo\n        path: ^probe/\n"}},
			WantExit: 1, WantText: []string{"# Why:"},
		},
		{
			Name: "lint/ausnahme-mit-why", Gate: gateLint,
			Ersetze:  []Ersetzung{{".golangci.yml", regelnKopf, regelnKopf + "      # Why: Probe einer begründeten Ausnahme.\n      - linters:\n          - forbidigo\n        path: ^probe/\n"}},
			WantText: []string{"0 issues"},
		},
		{
			Name: "lint/direktive-mit-leerzeichen", Gate: gateLint,
			Anhang:   map[string]string{"src/bootstrap/bootstrap.go": "\n//  nolint:unused // Probe\n"},
			WantExit: 1, WantText: []string{"Direktive", "nolintlint"},
		},
		{
			Name: "lint/profil-fehlt", Gate: gateLint,
			Entferne: []string{".golangci.yml"},
			WantExit: 2, WantText: []string{"fehlt oder ist nicht lesbar"},
		},
		{
			Name: "lint/why-falsch-eingerueckt", Gate: gateLint,
			Ersetze:  []Ersetzung{{".golangci.yml", regelnKopf, regelnKopf + "        # Why: Probe, falsch eingerückt.\n      - linters:\n          - forbidigo\n        path: ^probe/\n"}},
			WantExit: 1, WantText: []string{"# Why:"},
		},
	}
}

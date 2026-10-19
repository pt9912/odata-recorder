package main

import "strings"

// Gate benennt das Programm, das ein Fall in der Kopie aufruft.
type Gate int

const (
	gateACheck Gate = iota
	gateLint
	gateMigration
)

// Fall ist eine Gegenprobe: ein Eingriff in die Kopie der Eingaben und die
// Erwartung an das Gate.
type Fall struct {
	Name               string
	Gate               Gate
	Neu                map[string]string // Datei -> Inhalt, angelegt in der Kopie
	Anhang             map[string]string // Datei -> Text, an die Datei der Kopie gehängt
	Ersetze            []Ersetzung
	Entferne           []string          // Dateien, aus der Kopie gelöscht
	Kopiere            map[string]string // Ziel -> Quelle, beide in der Kopie
	Argumente          []string          // ersetzt die Argumente des Aufrufs von migration-check.sh
	OhneGoDateienInSrc bool              // löscht jede Datei *.go unter src/ der Kopie
	WantExit           int
	WantText           []string // jeder Text muss in der Ausgabe stehen
	Verboten           []string // keiner dieser Texte darf in der Ausgabe stehen
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

// zweitesModul liefert die Dateien eines Moduls tools/probe mit eigener go.mod.
func zweitesModul(zusatz string) map[string]string {
	return map[string]string{
		"tools/probe/go.mod":   "module probe\n\ngo 1.27\n",
		"tools/probe/probe.go": goDatei("probe") + zusatz,
	}
}

const (
	erzeugtesDDL = "erzeugt.sql"
	migrationen  = "src/adapters/driven/sqlite/migrations"
)

// migrationKopie legt die Migration als „erzeugtes DDL" neben die Kopie.
func migrationKopie() map[string]string {
	return map[string]string{erzeugtesDDL: migrationen + "/0001_init.sql"}
}

const regelnKopf = "\n    rules:\n"

func aCheckFaelle() []Fall {
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
		{
			Name: "a-check/bibliothek-in-fremdem-adapter", Gate: gateACheck,
			Neu:      map[string]string{"src/adapters/driven/andere/andere.go": "// Package andere importiert die SQLite-Bibliothek.\npackage andere\n\nimport _ \"modernc.org/sqlite\"\n"},
			WantExit: 1, WantText: []string{"tech-leak"},
		},
	}
}

func lintFaelle() []Fall {
	return []Fall{
		{Name: "lint/basis", Gate: gateLint, WantText: []string{"0 issues"}},
		{
			Name: "lint/direktive-nolint", Gate: gateLint,
			Anhang:   map[string]string{"src/bootstrap/bootstrap.go": "\n//" + "nolint:unused // Probe\n"},
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
			Anhang:   map[string]string{"src/bootstrap/bootstrap.go": "\n//  " + "nolint:unused // Probe\n"},
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
		{
			Name: "lint/zweites-modul-sauber", Gate: gateLint,
			Neu:      zweitesModul(""),
			WantText: []string{"0 issues"},
		},
		{
			Name: "lint/befund-im-zweiten-modul", Gate: gateLint,
			Neu:      zweitesModul("\nfunc unbenutzt() {}\n"),
			WantExit: 1, WantText: []string{"unused"},
		},
		{
			Name: "lint/modul-ohne-go-datei", Gate: gateLint,
			Neu:      map[string]string{"tools/leer/go.mod": "module leer\n\ngo 1.27\n"},
			WantExit: 1, WantText: []string{"tools/leer: keine Datei *.go"},
		},
		{
			Name: "lint/go-nur-unter-tools", Gate: gateLint,
			Neu:                zweitesModul(""),
			OhneGoDateienInSrc: true,
			WantExit:           1, WantText: []string{"./: keine Datei *.go"},
		},
		{
			Name: "lint/ohne-go-datei", Gate: gateLint,
			OhneGoDateienInSrc: true,
			WantExit:           1, WantText: []string{"keine Datei *.go"},
		},
	}
}

func migrationFaelle() []Fall {
	return []Fall{
		{Name: "migration/gleich", Gate: gateMigration, Kopiere: migrationKopie(), WantText: []string{"migration-check: ok"}},
		{
			Name: "migration/abweichung", Gate: gateMigration, Kopiere: migrationKopie(),
			Anhang:   map[string]string{erzeugtesDDL: "-- zusätzliche Zeile\n"},
			WantExit: 1, WantText: []string{"weicht vom erzeugten DDL ab", "< -- zusätzliche Zeile"}, Verboten: []string{"migration-check: ok"},
		},
		{
			Name: "migration/weitere-datei", Gate: gateMigration, Kopiere: migrationKopie(),
			Neu:      map[string]string{migrationen + "/0002_x.sql": "-- x\n"},
			WantExit: 1, WantText: []string{"ungeprüft"},
		},
		{
			Name: "migration/migration-fehlt", Gate: gateMigration, Kopiere: migrationKopie(),
			Entferne: []string{migrationen + "/0001_init.sql"},
			WantExit: 2, WantText: []string{"Migration", "fehlt"}, Verboten: []string{"erzeugtes DDL"},
		},
		{
			Name: "migration/erzeugtes-ddl-fehlt", Gate: gateMigration,
			WantExit: 2, WantText: []string{"erzeugtes DDL", "fehlt"}, Verboten: []string{"Migration src"},
		},
		{
			Name: "migration/falsche-aufrufform", Gate: gateMigration, Kopiere: migrationKopie(),
			Argumente: []string{"nur-eins"},
			WantExit:  2, WantText: []string{"Aufruf"},
		},
	}
}

func faelle() []Fall {
	alle := aCheckFaelle()
	alle = append(alle, lintFaelle()...)
	return append(alle, migrationFaelle()...)
}

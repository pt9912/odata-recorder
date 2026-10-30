package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Gate benennt das Programm, das ein Fall im Verzeichnis des Falls aufruft.
type Gate int

const (
	gateACheck Gate = iota
	gateLint
	gateMigration
	gatePruefung
	gateGroessenmarke
	gateTreiber
)

// Fall ist eine Gegenprobe: ein Eingriff in das Verzeichnis des Falls (die Kopie der Eingaben, bei
// einem Wegwerfmodul ein leeres Verzeichnis, das der Fall selbst füllt) und die Erwartung an das Gate.
type Fall struct {
	Name     string
	Gate     Gate
	Neu      map[string]string // Datei -> Inhalt, angelegt im Verzeichnis des Falls
	Anhang   map[string]string // Datei -> Text, an die Datei der Kopie gehängt
	Ersetze  []Ersetzung
	Entferne []string          // Dateien, aus der Kopie gelöscht
	Kopiere  map[string]string // Ziel -> Quelle, beide in der Kopie
	// Argumente ersetzen die Argumente von migration-check.sh (nil: die Vorgabe) und sind bei
	// gatePruefung die Argumente von pruefung; dort steht {dir} für das Verzeichnis des Falls und
	// {binary} für das Binary des Baus. Bei gateGroessenmarke sind es die Argumente von groessenmarke,
	// bei gateTreiber die von groessenmarke.sh (Programm und Planverzeichnis), beide mit {dir}.
	Argumente          []string
	Bau                *Bau // erzeugt vor dem Aufruf das Binary in der Kopie
	OhneGoDateienInSrc bool // löscht jede Datei *.go unter src/ der Kopie
	// Wegwerfmodul legt statt der Kopie der Eingaben ein leeres Verzeichnis an; der Fall schreibt
	// go.mod und die Pakete selbst (Neu).
	Wegwerfmodul bool
	WantExit     int
	WantText     []string // jeder Text muss in der Ausgabe stehen
	Verboten     []string // keiner dieser Texte darf in der Ausgabe stehen
	// Ausfuehrbar nennt Dateien aus Neu, die ausführbar werden (Ersatz-Skripte).
	Ausfuehrbar []string
	// Env sind zusätzliche Einträge der Umgebung des Aufrufs; {dir} steht für das Verzeichnis des
	// Falls und {PATH} für den PATH des Prozesses.
	Env []string
	// Vorbereitung läuft nach den Eingriffen und vor dem Aufruf, etwa um ein Git-Repo anzulegen.
	Vorbereitung func(dir string) error
	// Nachher prüft nach dem Aufruf das Protokoll der Ersatz-Skripte; sein Ergebnis ist der Grund,
	// aus dem der Fall rot ist, oder "". Es wird nur gefragt, wenn der Aufruf selbst die
	// Erwartung erfüllt.
	Nachher func(l Lauf) string
}

// Lauf ist, was Nachher vom Aufruf eines Falls sieht: das Verzeichnis und das TMPDIR des Falls,
// die Umgebung des Aufrufs, die Umgebung des Programms und die Ausgabe.
type Lauf struct {
	U   Umgebung
	Dir string
	Tmp string
	Env []string
	Out string
}

// treiber ruft den Treiber groessenmarke.sh im Verzeichnis des Falls mit denselben Umgebungswerten
// und liefert Exit-Code und Ausgabe.
func (l Lauf) treiber(args ...string) (int, string) {
	code, out, _ := starte(fallFrist, l.Dir, l.Env, "sh", append([]string{l.U.Treiber}, args...)...)
	return code, out
}

// Bau beschreibt das Binary, das ein Fall vor dem Aufruf von pruefung build in der Kopie erzeugt.
type Bau struct {
	Version      string
	OhneTrimpath bool
}

// umgebung liefert die zusätzliche Umgebung des Aufrufs: das eigene TMPDIR des Falls und die
// Einträge von Env.
func (f Fall) umgebung(dir, tmp string) []string {
	r := strings.NewReplacer("{dir}", dir, "{PATH}", os.Getenv("PATH"))
	env := []string{"TMPDIR=" + tmp}
	for _, e := range f.Env {
		env = append(env, r.Replace(e))
	}
	return env
}

// ersetzeOrte setzt dir und das Binary in die Argumente des Falls ein.
func (f Fall) ersetzeOrte(dir string) []string {
	r := strings.NewReplacer("{dir}", dir, "{binary}", filepath.Join(dir, binaerName))
	args := make([]string, len(f.Argumente))
	for i, a := range f.Argumente {
		args[i] = r.Replace(a)
	}
	return args
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
	binaerName   = "odata-recorder"
	probeVersion = "gegenprobe"
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
		{
			Name: "a-check/xml-im-kern", Gate: gateACheck,
			Neu:      map[string]string{"src/hexagon/model/kern/kern.go": "// Package kern importiert die XML-Bibliothek der Standardbibliothek.\npackage kern\n\nimport _ \"encoding/xml\"\n"},
			WantExit: 1, WantText: []string{"core-impurity"},
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
		{
			Name: "migration/zu-viele-argumente", Gate: gateMigration, Kopiere: migrationKopie(),
			Argumente: []string{"eins", "zwei", "drei"},
			WantExit:  2, WantText: []string{"Aufruf"},
		},
	}
}

// basisPaket ist das Paket basis/ ohne Befund mit einem bestandenen Test.
func basisPaket() map[string]string {
	return map[string]string{
		"basis/basis.go": leeresPaket("basis"),
		"basis/basis_test.go": "package basis\n\nimport \"testing\"\n\n" +
			"func TestBasis(t *testing.T) {\n\tif 1+1 != 2 {\n\t\tt.Fatal(\"rechnen\")\n\t}\n}\n",
	}
}

// testFaelle sind die Fälle 25 bis 28 und 33 (SPEC-102), die Fälle von `pruefung test`. Die Fälle 25 bis 28
// legen ein Wegwerfmodul an (Feld Wegwerfmodul, TestTestFaelleLegenEinWegwerfmodulAn) und fahren kein
// Paket von src/; der Fall test/aufruf ruft ohne Wurzel.
func testFaelle() []Fall {
	aufruf := []string{"test", "{dir}"}
	return []Fall{
		{
			Name: "test/basis", Gate: gatePruefung, Argumente: aufruf, Wegwerfmodul: true,
			Neu:      wegwerfModul(basisPaket()),
			WantText: []string{"pruefung test: ok"},
		},
		{
			Name: "test/nicht-formatiert", Gate: gatePruefung, Argumente: aufruf, Wegwerfmodul: true,
			Neu: wegwerfModul(basisPaket(),
				map[string]string{"basis/unformatiert.go": "package basis\n\nvar   unformatiert   =   1\n"}),
			WantExit: 1, WantText: []string{"nicht gofmt-formatiert", "unformatiert.go"},
		},
		{
			Name: "test/vet-befund", Gate: gatePruefung, Argumente: aufruf, Wegwerfmodul: true,
			Neu: wegwerfModul(basisPaket(), map[string]string{"basis/vet.go": "package basis\n\nimport \"fmt\"\n\n" +
				"func vetBefund() string { return fmt.Sprintf(\"%d\", \"text\") }\n"}),
			WantExit: 1, WantText: []string{"go vet:"},
		},
		{
			Name: "test/roter-test", Gate: gatePruefung, Argumente: aufruf, Wegwerfmodul: true,
			Neu: wegwerfModul(basisPaket(), map[string]string{"basis/rot_test.go": "package basis\n\nimport \"testing\"\n\n" +
				"func TestRot(t *testing.T) { t.Fatal(\"rot\") }\n"}),
			WantExit: 1, WantText: []string{"go test:"},
		},
		{
			Name: "test/aufruf", Gate: gatePruefung,
			WantExit: 2, WantText: []string{"Aufruf:"},
		},
	}
}

// buildFaelle sind die Fälle von `pruefung build`; das Binary entsteht in der Kopie.
func buildFaelle() []Fall {
	aufruf := []string{"build", "{binary}", probeVersion}
	return []Fall{
		{
			Name: "build/basis", Gate: gatePruefung, Argumente: aufruf, Bau: &Bau{Version: probeVersion},
			WantText: []string{"pruefung build: ok"},
		},
		{
			Name: "build/ohne-trimpath", Gate: gatePruefung, Argumente: aufruf,
			Bau:      &Bau{Version: probeVersion, OhneTrimpath: true},
			WantExit: 1, WantText: []string{"nicht mit -trimpath gebaut"},
		},
		{
			Name: "build/falsche-version", Gate: gatePruefung, Argumente: aufruf,
			Bau:      &Bau{Version: "andere"},
			WantExit: 1, WantText: []string{"Version des Binary", "andere", probeVersion},
		},
		{
			Name: "build/binary-fehlt", Gate: gatePruefung,
			Argumente: []string{"build", "{dir}/gibt-es-nicht", probeVersion},
			WantExit:  2, WantText: []string{"Binary fehlt"},
		},
	}
}

// wegwerfName ist der Name des Moduls, das die Fälle von `pruefung race` schreiben; Importpfade
// der Pakete beginnen mit ihm.
const wegwerfName = "wegwerf"

// leeresPaket ist eine Go-Datei, die nur ihr Paket nennt.
func leeresPaket(name string) string {
	return "package " + name + "\n"
}

// wegwerfModul liefert go.mod eines Moduls ohne Abhängigkeiten samt den Dateien der Pakete.
func wegwerfModul(pakete ...map[string]string) map[string]string {
	m := map[string]string{"go.mod": "module " + wegwerfName + "\n\ngo 1.27\n"}
	for _, p := range pakete {
		for datei, inhalt := range p {
			m[datei] = inhalt
		}
	}
	return m
}

// zaehleOhneKante ändert denselben Zähler aus zwei Goroutinen, zwischen denen keine Kante liegt;
// erst danach wartet der Aufrufer auf beide. Der Detektor meldet es aus der Ordnung der Zugriffe
// (SPEC-122 Nr. 8).
const zaehleOhneKante = `	var n int
	var wg sync.WaitGroup
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			n++
		}()
	}
	wg.Wait()
`

// saubererZaehler ist ein Paket, dessen Test zwei Goroutinen einen Zähler unter einer Sperre
// erhöhen lässt und auf beide wartet.
func saubererZaehler(dir string) map[string]string {
	return map[string]string{
		dir + "/sauber.go": leeresPaket("sauber"),
		dir + "/sauber_test.go": "package sauber\n\nimport (\n\t\"sync\"\n\t\"testing\"\n)\n\n" +
			"func TestZaehlerUnterSperre(t *testing.T) {\n\tvar mu sync.Mutex\n\tvar wg sync.WaitGroup\n\tn := 0\n" +
			"\twg.Add(2)\n\tfor range 2 {\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tmu.Lock()\n\t\t\tn++\n\t\t\tmu.Unlock()\n\t\t}()\n\t}\n" +
			"\twg.Wait()\n\tif n != 2 {\n\t\tt.Fatalf(\"Zaehler %d\", n)\n\t}\n}\n",
	}
}

// rennenImTest ist ein Paket, dessen Rennen allein im Code seines Tests liegt.
func rennenImTest(dir string) map[string]string {
	return map[string]string{
		dir + "/rennen.go": leeresPaket("rennen"),
		dir + "/rennen_test.go": "package rennen\n\nimport (\n\t\"sync\"\n\t\"testing\"\n)\n\n" +
			"func TestRennen(t *testing.T) {\n" + zaehleOhneKante + "\tt.Log(n)\n}\n",
	}
}

// rennenImCode ist ein Paket, dessen Rennen im Code des Pakets liegt und das sein Test aufruft.
func rennenImCode(dir string) map[string]string {
	return map[string]string{
		dir + "/zaehle.go": "package rennen\n\nimport \"sync\"\n\n" +
			"// Zaehle liefert den Zaehler nach dem Rennen.\nfunc Zaehle() int {\n" + zaehleOhneKante + "\treturn n\n}\n",
		dir + "/zaehle_test.go": "package rennen\n\nimport \"testing\"\n\nfunc TestZaehle(t *testing.T) {\n\tZaehle()\n}\n",
	}
}

// roterTest ist ein Paket, dessen Test ohne Rennen rot ist.
func roterTest(dir string) map[string]string {
	return map[string]string{
		dir + "/rot.go":      leeresPaket("rot"),
		dir + "/rot_test.go": "package rot\n\nimport \"testing\"\n\nfunc TestRot(t *testing.T) {\n\tt.Fatal(\"rot\")\n}\n",
	}
}

// rennenUndBlockade ist ein Paket, dessen Test ein Rennen erzeugt, auf beide Goroutinen wartet und
// danach blockiert.
func rennenUndBlockade(dir string) map[string]string {
	return map[string]string{
		dir + "/zeit.go": leeresPaket("zeit"),
		dir + "/zeit_test.go": "package zeit\n\nimport (\n\t\"sync\"\n\t\"testing\"\n)\n\n" +
			"func TestRennenUndBlockade(t *testing.T) {\n" + zaehleOhneKante + "\tt.Log(n)\n\tselect {}\n}\n",
	}
}

// raceFaelle sind die Fälle 34 bis 36 (SPEC-102), die Fälle von `pruefung race`. Jeder legt ein
// Wegwerfmodul an (Feld Wegwerfmodul, TestRaceFaelleLegenEinWegwerfmodulAn) und schreibt dessen
// Pakete selbst. Sie setzen CGO_ENABLED nicht selbst und erben 0 von der Stufe gegenproben; dass die
// Stufe es setzt, fängt kein Test (SPEC-102, Aufbau).
func raceFaelle() []Fall {
	wurzel := []string{"race", "{dir}"}
	return []Fall{
		{
			Name: "race/ohne-rennen", Gate: gatePruefung, Argumente: wurzel,
			Wegwerfmodul: true,
			Neu:          wegwerfModul(saubererZaehler("src/hexagon/sauber")),
			WantText:     []string{"pruefung race: ok"},
		},
		{
			Name: "race/rennen-in-jedem-teilbaum", Gate: gatePruefung, Argumente: wurzel,
			Wegwerfmodul: true,
			Neu: wegwerfModul(rennenImTest("src/hexagon/rennen"), rennenImCode("src/adapters/rennen"),
				rennenImCode("src/bootstrap/rennen"), roterTest("src/hexagon/rot")),
			WantExit: 1,
			WantText: []string{
				"WARNING: DATA RACE",
				"FAIL\t" + wegwerfName + "/src/hexagon/rennen", "FAIL\t" + wegwerfName + "/src/adapters/rennen",
				"FAIL\t" + wegwerfName + "/src/bootstrap/rennen", "FAIL\t" + wegwerfName + "/src/hexagon/rot",
			},
		},
		{
			Name: "race/rennen-und-zeitgrenze", Gate: gatePruefung,
			Argumente:    []string{"race", "{dir}", "-timeout=5s", "./..."},
			Wegwerfmodul: true,
			Neu:          wegwerfModul(rennenUndBlockade("src/hexagon/zeit")),
			WantExit:     1, WantText: []string{"WARNING: DATA RACE", "test timed out"},
		},
	}
}

func faelle() []Fall {
	alle := aCheckFaelle()
	alle = append(alle, lintFaelle()...)
	alle = append(alle, migrationFaelle()...)
	alle = append(alle, testFaelle()...)
	alle = append(alle, buildFaelle()...)
	alle = append(alle, raceFaelle()...)
	return append(alle, groessenmarkeFaelle()...)
}

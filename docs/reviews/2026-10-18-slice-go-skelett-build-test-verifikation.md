# Verifikation: slice-go-skelett-build-test

**Rolle:** Verifier (Modul 11) — Frage „Bauen wir es richtig?" gegen Plan und DoD.
**Gegenstand:** `docs/plan/planning/done/slice-go-skelett-build-test.md`, Commits 703f174, 5906398, 3386932 (Review-Report 425e373).
**Bezug:** `LH-QA-06`, `LH-RB-03`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md)
**Datum:** 2026-10-18

## Verdikt

**DoD-Liefer-Punkte 1 bis 4 erfüllt, belegt durch eigene Sensor-Läufe. Keine HIGH- und keine MEDIUM-Befunde.** Freigabe zur Closure; die offenen Closure-Pflichten bleiben unten als offen gemeldet.

## Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis |
|---|---|
| `make gates` (volle Kette inkl. a-check, d-check, build, test, schema-check, Gate-Nachweis) | Exit 0; Log zeigt `docker build --target build` und `--target test` |
| `make test` | Exit 0 |
| `make build` | Exit 0; `odata-recorder:build /out/odata-recorder version` gibt `odata-recorder dev`, `foo` gibt `Unbekanntes Kommando: foo` mit Exit 2 |
| `make build VERSION=1.2.3` | Exit 0 (die Versionsprüfung im Ziel trägt den Wert durch) |
| `git status` nach den Läufen | sauber; im Repo wurde nichts geändert außer diesem Report |

## Mutationen (Wegwerfkopie unter dem Scratchpad, nicht im Repo)

Basislauf der Kopie grün (`test` 0, `build` 0). Jede Mutation machte das Gate rot (Exit 2):

| Zusage | Mutation | Ergebnis |
|---|---|---|
| Exit-Code 2 bei Aufrufmuster-Fehler | `exitAufruf = 1` | `make test` rot |
| gofmt-Prüfung | nicht formatierte Datei `x.go` | `make test` rot |
| `go vet` | `Printf("%d", "x")` | `make test` rot |
| Version per `-X main.version` | `-X` entfernt, `VERSION=9.9` | `make build` rot |
| gleiche Zusage | Variablenname `main.versio` | `make build` rot |
| `-trimpath` | Flag entfernt | `make build` rot |
| nil-Writer verwirft Ausgabe | beide nil-Schutzblöcke deaktiviert | `make test` rot |
| `version` mit Zusatzargument endet mit 2 | Bedingung auf `args[0] == "version"` gelockert | `make test` rot (siehe F2) |

## DoD-Prüfung (§2)

1. **Go-Modul und Composition Root mit Test: erfüllt.** `go.mod` (`github.com/pt9912/odata-recorder`, keine Fremdabhängigkeit), `src/bootstrap/bootstrap.go` (`Run`), `src/bootstrap/main/main.go` (einzige Stelle mit `os.Exit`, Version per `-ldflags`), `src/bootstrap/bootstrap_test.go` im Paket `bootstrap_test` mit Table-Test (Happy, unbekannt, Zusatzargument, leeres Argument, `--help`) und `TestRunNilWriter`.
2. **Dockerfile: erfüllt.** Stufen `deps`, `source`, `test` (`gofmt -l`, `go vet`, `go test`), `build` (`-trimpath -buildvcs=false`). Das einzige `FROM` trägt `@sha256:`-Digest. Netz nur in `deps` (`go mod download`); `test` und `build` laufen mit `RUN --network=none`. `.dockerignore` ist eine Allowlist (`go.mod`, `go.sum`, `src/`).
3. **Gates: erfüllt.** `harness/mk/go-gates.mk` setzt `GATE_CHECKS += build test`; `harness/README.md` §Sensors führt beide Zeilen (Z. 66 und 67) mit Bindung [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
4. **`make gates` grün: erfüllt** (siehe Sensor-Belege).

## Plan-vs-Code (§3 gegen Diff 0accd47..HEAD)

Alle elf Pfade aus §3 sind im Diff (`go.mod`, `src/bootstrap/{bootstrap.go,main/main.go,bootstrap_test.go}`, `Dockerfile`, `.dockerignore`, `harness/mk/go-gates.mk`, `roadmap.md`, `harness/vertraege/SPEC-036-test.md` und `README.md`, `harness/README.md`). Es gibt keine Datei im Code-Diff, die §3 nicht nennt, und keine genannte Datei ohne Änderung. Die Review-Korrekturen (5906398) sind im Plan nachgezogen (3386932; `AGENTS.md` §3.9). Umfang: 181 Zeilen hinzugefügt, 6 entfernt über die Pfade aus §3, unter der Rückführungs-Marke von 400. Die Beschreibung von `Run` im Plan (`Run(args, version, stdout, stderr) int`, nil-Writer) entspricht dem Code. Die Review-Befunde F1 bis F7 sind im Code, in `SPEC-036-test.md` und im Plan sichtbar umgesetzt (Randformen entschieden und getestet, `make build` prüft Version und `-trimpath`, nil-Schutz, Roadmap-Marker in §3).

## Konformität

- **[ADR-0001](../plan/adr/0001-implementierungssprache-go.md):** Go, Wurzel-`go.mod`, Quellbaum unter `src/`, Build und Test ausschließlich über `make` in Containern (`AGENTS.md` §3.1). Folgepflicht „Dockerfile und Build-Gates im ersten Code-Slice" ist erfüllt. Konform.
- **[ADR-0002](../plan/adr/0002-architektur-gate-a-check.md):** `make a-check` läuft in `make gates` grün; `src/bootstrap/main` liegt im Glob `src/bootstrap/**` (`composition_root` in `.a-check.yml`) und importiert nur `bootstrap` und die Standardbibliothek. Konform (siehe F4 zur Aussagekraft).
- **[ADR-0003](../plan/adr/0003-lint-profil-golangci.md):** Lint ist laut Plan §1 ausdrücklich nicht im Slice (slice-lint-stufe); `make lint-profil` ist grün. Ein Lint-Lauf gegen den Code existiert noch nicht. Nicht verletzt, nicht belegbar (siehe F3).
- **Architektur-Sicht `spec/architecture.md` §2:** `ARC-012` (Composition Root unter `src/bootstrap`) und der Baum bleiben gewahrt; das Unterverzeichnis `main` liegt innerhalb des Globs (Risiko in §6 zu Recht „entfallen"). Keine Welle, kein Slice in der Sicht (`AGENTS.md` §3.4).
- **Hard Rules:** §3.10 (Mutation je Zusage): oben nachgefahren. §3.11 (Kommentar sagt nur zu, was ein Test prüft): die Kommentare in `go-gates.mk`, `Dockerfile` und `bootstrap.go` nennen Zusagen, die durch Test oder Gate-Mutation gedeckt sind.

## Befunde

### F1 — INFO: Der Basislauf bindet sich an das Docker-Build-Cache-Verhalten
`make test` und `make build` laufen bei warmem Cache ohne Netz. Das Kriterium „aus einem frischen Klon ohne Host-Toolchain" (§5) braucht beim ersten Lauf Netz für `docker pull` und `go mod download` (leere `go.mod`, also faktisch nur das Image). Ein Lauf aus kaltem Cache im frischen Klon wurde nicht gefahren; die Zusage „Netz nur in `deps`" ist durch das Dockerfile belegt, nicht durch einen Lauf mit gesperrtem Netz.

### F2 — LOW: Mutation „Zusatzargument" wird teilweise durch Panic rot
Die gelockerte Bedingung `args[0] == "version"` greift bei leeren Argumenten auf `args[0]` zu und panikt dort, bevor der Test „version mit Zusatzargument" ansprechen kann. Rot wird die Mutation also; ob der benannte Grund (Zusatzargument) den Test allein tragen würde, ist mit dieser Mutation nicht getrennt. Gesonderte Gegenprobe (`len(args) >= 1 && args[0] == "version"`) wurde nicht gefahren. Der Test „version mit Zusatzargument" erwartet Exit 2 und Text `Unbekanntes Kommando: version` und fängt diese Variante nach Lesen des Codes.

### F3 — INFO: Lint-Konformität nicht maschinell belegt
`make lint` existiert noch nicht (slice-lint-stufe). Die Aussage des Reviews (F8 dort) bleibt Lesebefund.

### F4 — INFO: a-check prüft einen sehr dünnen Baum
Außerhalb von `src/bootstrap/**` gibt es keinen Code; die Importrichtungs-Regeln sind noch ohne Gegenprobe (`harness/README.md` §Sensors nennt das ausdrücklich: „prüft heute einen leeren Baum"). Kein Befund gegen diesen Slice.

### F5 — INFO: `VERSION` mit Sonderzeichen
`make build` reicht `VERSION` unquotiert in `--build-arg` und in die Prüfung durch. Für das Skelett ohne Wirkung, bei einer Release-Version mit Leerzeichen oder Shell-Zeichen wäre es zu prüfen. Nicht Teil der DoD.

## Offen (keine Befunde, Closure folgt nach dieser Verifikation)

- Review durchgeführt, Report liegt vor: Report liegt (`2026-10-18-slice-go-skelett-build-test-review.md`); Häkchen setzt die Closure.
- Doku-Update bei berührtem öffentlichem Vertrag
- Closure-Notiz mit Steering-Loop-Lerneintrag (§7: „Wird bei der Closure geschrieben.")
- Beobachtungs-Register fortgeschrieben
- Jedes Risiko aus §6 trägt einen Ausgang: zwei stehen auf „entfallen", eines (Go-Toolchain-Image altert) auf „weiter offen"; die Bestätigung gehört in die Closure
- Die drei Paarungen (Anker · Folge-Slice · Register)

## Negativbefunde (geprüft, ohne Befund)

Digest-Pin und `--network=none`-Verteilung im Dockerfile; Exit-Codes und Texte in `bootstrap.go` gegen `SPEC-036-test.md`; Eintrag beider Targets in `GATE_CHECKS`; Slice-Größe gegen die Rückführungs-Marke; keine Änderungen an `spec/` oder ADRs im Slice-Diff.

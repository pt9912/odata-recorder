# Review-Report: slice-gegenproben — 2026-10-19

**Review-Art:** Code (gegen Plan + Konventionen)

**Gegenstand:** `slice-gegenproben`, Commits `869eeb0` (SPEC-102 vor dem Code) und `a4dd41f` (Umsetzung), Basis `3117b17`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Eingangs-Kontext:**

- Slice-Plan `slice-gegenproben` (§1, §2, §3, §4, §6)
- [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Status Proposed; die ausstehende Owner-Entscheidung ist kein Befund)
- `harness/vertraege/SPEC-102-gegenproben.md`
- [LH-QA-06](../../spec/lastenheft.md)
- `AGENTS.md` (Hard Rules §3.2, §3.7, §3.9 bis §3.14)

**Mutationen (selbst gefahren, über `make gegenproben`, danach `git checkout`; der Stempel ist unverändert):**

| Mutation | Erwartung | Ergebnis |
|---|---|---|
| `# Why:`-Prüfung in `lint.sh` aus | Fall `lint/ausnahme-ohne-why` rot | rot ("Verstoß nicht gemeldet"), Stempel unverändert |
| Direktiv-Prüfung in `lint.sh` aus | Fall `lint/direktive-nolint` rot | rot |
| `lint.sh` endet immer mit Exit 0 | beide Verstoß-Fälle rot | rot |
| Rolle `domain` des Modells auf `app` | `a-check/kern-auf-driven-adapter` rot | rot |
| Kante Driven Adapter zu Outbound Port entfernt | `a-check/erlaubter-weg-ueber-port` rot | rot |
| Glob der Driven Adapter verschoben | drei a-check-Fälle rot | rot |
| Prüfung auf eingerückte Regel (F-6) in `lint.sh` aus | roter Fall | **grün** (F-1) |
| Profil-fehlt-Prüfung (F-5) in `lint.sh` aus | roter Fall | **grün** (F-1) |
| Direktive mit mehreren Leerzeichen (F-4) in `lint.sh` aus | roter Fall | **grün** (F-1) |
| `sleep 250` am Anfang von `lint.sh` (Randform 6) | Fälle nach 120 s rot | alle vier Lint-Fälle rot erst nach 250 s (F-2) |
| `.gitignore`-/Löschfall der Prüfsumme | siehe F-5 | Skript endet mit Exit 1 |

Beobachtet: Stempel und Prüfsumme stimmen auf HEAD überein; `make gegenproben` bei unverändertem Stand 0,042 s, kalt 13,4 s (Budget 1 s / 30 s gehalten); `--no-cache-filter` führt die Stufe bei geänderter Eingabe tatsächlich aus (Fall-Ausgabe erscheint); ein roter Lauf lässt den Stempel stehen; `make docs-check` ohne Befund.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Plan §1 und der Folge-Slice-Eintrag von `slice-lint-stufe` weisen dieser Gegenprobe die Fälle F-4 (Direktive mit zwei Leerzeichen), F-5 (Profil fehlt) und F-6 (Einrückung des `# Why:`-Blocks) zu; SPEC-102 führt sie nicht unter den acht Fällen, und weder Plan noch SPEC-102 nehmen sie zurück. Entfernt man die drei Prüfungen aus `lint.sh`, bleibt `make gegenproben` grün; das `entfallen` in §6 von `slice-lint-stufe` ("die Gegenprobe hält sie") trägt damit nicht. | `AGENTS.md` §3.13 · §3.9 · [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) | `docs/plan/planning/done/slice-gegenproben.md` · "ein fehlendes Profil, eine Direktive mit zwei Leerzeichen nach `//` und die Einrückung des `# Why:`-Blocks" | ja (die drei Mutationen oben, `make gegenproben`) | Zugesagter Fall nicht geliefert, Plan und Vertrag uneins |
| F-2 | MEDIUM | Die Frist von 120 s (Randform 6) wird nicht durchgesetzt: `CommandContext` beendet nur `sh`, das Kind `golangci-lint` hält die Ausgabe-Pipe offen, `cmd.Run()` kehrt erst mit ihm zurück (gemessen: 250 s statt 120 s bei `sleep 250`; bei einem hängenden Linter hängt der Docker-Build). Der Test deckt nur `bewerte(..., abgelaufen=true)`, nicht `rufe`. | `AGENTS.md` §3.10 · §3.12 · [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) | `tools/gegenproben/main.go` · `ctx, stop := context.WithTimeout(context.Background(), fallFrist)` | ja (Mutation `sleep` in `lint.sh`) | Zeitgrenze erreicht Kindprozesse nicht |
| F-3 | MEDIUM | Zu den Randformen 1, 2, 7, 8 und 9 gibt es keinen Test und keine Mutation; `main_test.go` prüft nur `bewerte` und die Eindeutigkeit der Namen. Die Prüfung der unveränderten Eingabe (Randform 9) vergleicht Hashes von `.`, während die Fälle in Kopien laufen; nichts wird rot, wenn man sie entfernt. Der Kopf von SPEC-102 sagt sie als Zusagen zu. | `AGENTS.md` §3.10 · §3.11 | `tools/gegenproben/main_test.go` · `func TestBewerte` | nein (kein Gate erfasst fehlende Tests; Lesebefund) | Zusage des Vertrags ohne fangenden Test |
| F-4 | LOW | Die Zeile von [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) steht im ADR-Index doppelt. | `AGENTS.md` §5 Regel 3 | `docs/plan/adr/README.md` · Tabellenzeile zur Kennung 0005 (zwei identische Zeilen) | ja (`grep -c`) | Index-Zeile doppelt |
| F-5 | LOW | `gegenproben-eingabe.sh` endet mit Exit 1, wenn die zuletzt sortierte Datei fehlt (getrackt, im Arbeitsbaum gelöscht; gemessen mit `tools/harness/lint.sh`) — `make gegenproben` bricht dann ohne Meldung ab. Auch ohne `.git` scheitert das Skript. SPEC-102 entscheidet diese Randform des Skripts nicht. | `AGENTS.md` §3.12 | `tools/harness/gegenproben-eingabe.sh` · `[ -f "$f" ] && sha256sum -- "$f"` | ja (Datei verschieben, Skript aufrufen) | Randform eines Werkzeugs nicht entschieden |
| F-6 | LOW | [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) sagt, "die Pflicht zum Fahren steht in der Spezifikation" (Alternative B); SPEC-102 enthält sie nicht (nur "Stand: vorgesehen für …"), der Plan §6 sagt, sie stehe in der ADR, und die Entscheidung der ADR nennt sie selbst samt Fällen. Die Verortung ist dreifach uneins. | `AGENTS.md` §3.8 · §3.11 | `docs/plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md` · "die Pflicht zum Fahren steht in der Spezifikation" | nein | Verortung einer Regel widerspricht sich |
| F-7 | LOW | Der Kopf von `.dockerignore` sagt, der Kontext enthalte "nur was Build, Test und Lint brauchen"; `tools/gegenproben` und `.a-check.yml` brauchen weder Build noch Test noch Lint des Produkts. | `AGENTS.md` §3.7 | `.dockerignore` · "nur was Build, Test und Lint brauchen" | nein | Kommentar beschreibt nicht, was da ist |
| F-8 | INFO | `tools/gegenproben` liegt außerhalb von `make lint` (SPEC-102 nennt das). Das Profil meldet auf dem Programm vier Befunde: `errcheck` (`defer os.RemoveAll`), `forbidigo` zweimal (`os.Stdout`, `os.Stderr` in `main`), `gochecknoglobals` (`eingaben`). Zudem enthält `faelle.go` den Text `//nolint:unused` in einem String, den `lint.sh` als Direktive läse, sobald das Modul im Lint-Pfad läge. | [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) | `tools/gegenproben/main.go` · `var eingaben = []string{` | ja (golangci-lint mit `.golangci.yml` auf dem Modul) | Neuer Code außerhalb des Lint-Pfads |
| F-9 | INFO | `go.sum` steht in der Allowlist von `.dockerignore`, wird aber weder in die Stufe `gegenproben` kopiert noch in die Prüfsumme einbezogen oder von `hashEingaben` gelesen. Mit der ersten Abhängigkeit fehlt der Stufe `go.sum` und das Modul-Cache (`-mod=readonly`, kein Netz). Die Liste der Eingaben steht außerdem viermal (Programm, Skript, `Dockerfile`, `.dockerignore`). | `AGENTS.md` §3.11 | `Dockerfile` · "COPY go.mod .a-check.yml .golangci.yml ./" | nein | Eingabenliste mehrfach geführt |
| F-10 | INFO | Die Prüfsumme folgt `.gitignore` (`git ls-files`), der Docker-Kontext nicht: eine ignorierte Datei unter `src/` liegt im Kontext, aber nicht in der Prüfsumme. | Maintainability | `tools/harness/gegenproben-eingabe.sh` · `--exclude-standard` | ja (ignorierte Datei anlegen) | Hash-Menge und Kontext-Menge verschieden |
| F-11 | INFO | SPEC-102 trägt die Zeile "Stand: vorgesehen für `make a-check` und `make lint`"; sie nennt keinen Zustand und keinen Beleg. Die Rückführungs-Marke (600 Zeilen) liegt in der Metrik des Plans bei 576. | `AGENTS.md` §3.7 · §3.14 | `harness/vertraege/SPEC-102-gegenproben.md` · "**Stand:** vorgesehen für" | nein | Zustandsfeld ohne Zustand |

## Negativbefunde

- `tools/gegenproben/` (Fälle 1 bis 8 gegen den Code; Randformen 3, 4, 5 und 7 gegen `bewerte`, `fahre`): geprüft, ohne weiteren Befund; Randformen 1, 2, 6, 8, 9 siehe F-2 und F-3.
- `Dockerfile` (Stufen `a-check-bin`, `gegenproben`): geprüft, ohne Befund; netzlos, Pins als Build-Argumente, Programm wird vor dem Lauf mit gofmt, vet und Test geprüft.
- `harness/mk/gegenproben.mk` (Prüfsumme, Stempel nur nach grünem Lauf, `--no-cache-filter`, nicht in `GATE_CHECKS`): geprüft, ohne Befund.
- `tools/harness/gegenproben-eingabe.sh`: alle Eingaben der Stufe außer `go.sum` sind in der Prüfsumme (siehe F-5, F-9, F-10); Pins fließen ein; Stempel stimmt auf HEAD.
- `harness/vertraege/SPEC-102-gegenproben.md`, `harness/README.md`, `.dockerignore`: geprüft, Befunde F-7, F-11 (`.dockerignore`, SPEC); `harness/README.md` ohne Befund.
- `docs/plan/` ([ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), Index, Plan, `slice-lint-stufe`, Roadmap): Befunde F-1, F-4, F-6; Roadmap ohne Befund.
- Suppression (`AGENTS.md` §3.2): keine Direktive im Code, keine neue Ausnahme in `.golangci.yml`; geprüft, ohne Befund.

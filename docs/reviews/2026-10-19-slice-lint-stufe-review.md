# Review-Report: slice-lint-stufe — 2026-10-19

**Review-Art:** Code (gegen Plan, ADR und Vertrag; nicht gegen die DoD)

**Gegenstand:** `git diff 7f425d3..HEAD` — Commit `fa96a11` (Vertrag SPEC-101 vor dem Code) und `8d93411` (Umsetzung); davor nur Lifecycle-Moves. Geprüft: `tools/harness/lint.sh`, `Dockerfile` (Stufe `lint`), `harness/mk/lint-profil.mk`, `.dockerignore`, `.golangci.yml` (Kopfkommentar), `harness/README.md` §Sensors, `harness/vertraege/SPEC-101-lint-profil.md`, Plan- und Roadmap-Nachzüge.

**Skill:** `.harness/skills/reviewer.md` (Arbeitsbaum)
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Eingangs-Kontext:**

- Plan `slice-lint-stufe` (§1, §3, §6)
- [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `LH-QA-06`
- `harness/vertraege/SPEC-101-lint-profil.md`
- `AGENTS.md` §3.2, §3.7, §3.9–§3.14

**Methode.** Randformen aus SPEC-101 einzeln mutiert. Jede Mutation lief in einer Wegwerfkopie im Scratchpad, ausgeführt in einem Container aus dem gepinnten golangci-lint-Image ohne Netz (`lint.sh` per `docker cp`); das Repo blieb unverändert (`git status` sauber). `make lint` selbst lief grün, aber aus dem Build-Cache.

Mutationen und Ergebnis (alle wie im Vertrag, außer F-1):

- `//nolint` in `*_test.go`, in einer Zeichenkette, `// lint:ignore` und `//lint:ignore`, `//<Tab>nolint`: Befund `lint.sh: <Datei>:<Zeile>`, Exit 1.
- Keine `*.go`: Befund der Prüfung 4, Exit 1.
- Profil fehlt: Meldung und Exit 2.
- Leerzeile zwischen `# Why:`-Block und Regel: Befund.
- `# Why:` ohne Text, `# Why:` nicht als erste Zeile des Blocks, `#Why:`: Befund.
- `warn-unused` fehlt, `false` oder auskommentiert: Befund.
- `exclusions` fehlt: Befund `warn-unused fehlt`.
- Regeln mit vier Leerzeichen: Befund je Regel.
- `exclusions` ohne `rules:`: kein Befund (wie Randform 3).
- `//  nolint` mit zwei Leerzeichen: von `lint.sh` nicht erkannt, aber `nolintlint` meldet es und das Gate wird rot.
- `/* nolint */` und `//lint:file-ignore`: unwirksam, die Befunde bleiben sichtbar.

---

## Findings

HIGH: keines. Die HIGH-Regeln des Skills sind geprüft: keine ADR-Verletzung (kein Linter gelockert, Profil nur im Kopfkommentar berührt), keine Suppression, keine Geheimnis-Senke, Kommentare ohne verworfene Alternativen oder abwesenden Text.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | SPEC-101 Randform 3 sagt: `exclusions.rules` leer ist kein Befund der Prüfung 2. Mit `    rules: []` (gültiges YAML, leere Liste) meldet `lint.sh` `rules nicht als Block-Liste` und das Gate wird rot. Der Plan führt die Randform in §6 als „entfallen: entschieden", der Code folgt der Entscheidung nicht. | AGENTS.md §3.12, §3.10; SPEC-101 Randform 3 | `tools/harness/lint.sh` · `inx && /^    rules:[ ]*[^ #]/ { befund(NR, "rules nicht als Block-Liste"); next }` | ja — `.golangci.yml` mit `rules: []` in `make lint` | Randform des Vertrags im Code abweichend umgesetzt |
| F-2 | LOW | Der Kopfkommentar sagt, der Kontext sei eine Allowlist „nur was der Go-Build braucht"; mit `.golangci.yml` und `tools/harness/lint.sh` steht dort jetzt mehr als der Go-Build. Der Kommentar nennt eine Zusage, die der Zustand nicht trägt. | AGENTS.md §3.7, §3.11 | `.dockerignore` · `nur was der Go-Build braucht` | nein | Kommentar sagt mehr zu als der Zustand |
| F-3 | INFO | `ARG LINT_IMAGE=scratch` ist der Vorgabewert. Ein direkter `docker build --target lint` ohne das Argument scheitert an einem Image ohne `sh`, nicht mit einer Meldung zum fehlenden Argument. Über `make lint` tritt das nicht auf. | Maintainability | `Dockerfile` · `ARG LINT_IMAGE=scratch` | ja — `docker build --target lint .` | Vorgabewert eines Pflicht-Arguments |
| F-4 | INFO | Die Prüfung 1 erkennt `//  nolint` (zwei Leerzeichen) nicht, weil der Text nur ein optionales Leerzeichen zulässt. SPEC-101 nennt nur die vier Formen mit höchstens einem Leerzeichen, der Code folgt dem. Fängt aktuell `nolintlint` ab (rot bestätigt); fällt aber `nolintlint` aus dem Profil, bleibt die Form durch. Für die Gegenprobe zu `make lint`. | AGENTS.md §3.11 | `tools/harness/lint.sh` · `//[[:space:]]?(nolint\|lint:ignore)` | ja — Datei mit `//  nolint:x` | Direktiv-Schreibweise nur durch zweiten Linter gedeckt |
| F-5 | INFO | Randform 1 (Exit-Code 2, Meldung `Profil .golangci.yml fehlt`) ist über `make lint` nicht erreichbar: `COPY .golangci.yml` bricht den Build vorher mit anderem Fehler ab. Der Pfad ist nur bei direktem Aufruf von `lint.sh` zu sehen. Zuständig: slice-gegenproben. | AGENTS.md §3.10 | `Dockerfile` · `COPY .golangci.yml ./` | ja — Profil umbenennen, `make lint` | Zusage nur am Skript, nicht am Gate beobachtbar |
| F-6 | INFO | Ein Kommentar im Körper der vorigen Regel, der mit `# Why:` beginnt und direkt über der nächsten Regel steht, zählt als Why-Block dieser Regel (Mutation: Einrückung acht Leerzeichen, Gate grün). SPEC-101 nennt nur „Kommentarblock unmittelbar unter"; die Einrückung ist nicht entschieden. Außerdem laufen die Warnungen von `warn-unused` bei einem Cache-Treffer nicht erneut in die Ausgabe. | AGENTS.md §3.12 | `tools/harness/lint.sh` · `inr && /^ *#/ { if (!prev) first = $0; prev = 1; next }` | ja | Einrückung des Why-Blocks nicht entschieden |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `harness/mk/lint-profil.mk` (Target `lint`, `GATE_CHECKS`, Build-Argument aus `GOLANGCI_LINT_IMAGE`, Pin an einer Stelle) | geprüft, ohne Befund |
| `Dockerfile` Stufen `deps`, `source`, `lint` (netzlos, `--network=none`, Module aus `deps`, `COPY src`) | geprüft, ohne Befund |
| `harness/README.md` §Sensors (Zeile `make lint`, „Nicht behauptet") | geprüft, ohne Befund |
| `harness/vertraege/SPEC-101-lint-profil.md` (Randformen 1–7 vor dem Code, Exit-Codes) | geprüft, ohne Befund außer F-1 |
| `.golangci.yml` (Kopfkommentar nennt die gelesene Form, keine Schwelle verändert) | geprüft, ohne Befund |
| Plan-Nachzug §3.9 (`slice-lint-stufe` §3, Verweise in `slice-gegenproben`, `slice-go-skelett-build-test`, Roadmap) | geprüft, ohne Befund |
| `AGENTS.md` §3.2 (kein `//nolint` im Diff, Ausnahmen nur mit `# Why:`) | geprüft, ohne Befund |

Nicht geprüft: Vollständigkeit der DoD und der Bericht nach §3.10 je Zusage (Verifier), Bedarf gegen die Realität (Validator).

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 1 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Randform des Vertrags im Code abweichend umgesetzt · Kommentar sagt mehr zu als der Zustand · Vorgabewert eines Pflicht-Arguments · Direktiv-Schreibweise nur durch zweiten Linter gedeckt · Zusage nur am Skript, nicht am Gate beobachtbar · Einrückung des Why-Blocks nicht entschieden

## Verdikt

**Merge-blockierend:** ja — wegen F-1 (MEDIUM): eine im Vertrag entschiedene Randform verhält sich im Code anders; der Plan führt sie als „entfallen".

**Übergabe:** Findings gehen an den Implementer; F-1 berührt eine Randform, die SPEC-101 vor dem Code entschieden hat — hält der Implementer die Spec für falsch, geht der Punkt an den Architect (AGENTS.md §3.12), er wird nicht im Code umentschieden. Die Finding-Klassen gehen in die Slice-Closure §7.

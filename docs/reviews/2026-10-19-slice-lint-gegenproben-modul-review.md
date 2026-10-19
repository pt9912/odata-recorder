# Review slice-lint-gegenproben-modul

* Rolle: Reviewer (`.harness/skills/reviewer.md`), frischer Kontext, kein Self-Review.
* Plan: [slice-lint-gegenproben-modul](../plan/planning/in-progress/slice-lint-gegenproben-modul.md)
* Diff: `66de972..HEAD` (Vertrag `70a2001`, Umsetzung `fdbd23b`); Summe aus `git diff --numstat` über die Pfade aus §3 des Plans: 174 Zeilen, unter der Marke von 300.
* Bezug: `LH-QA-06`, [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md); Verträge `SPEC-101`, `SPEC-102`
* Geprüft gegen Plan, ADRs, Verträge und `AGENTS.md` §3.2, §3.6, §3.7, §3.9–§3.14. Nicht gegen die DoD (Verifier).

## Ergebnis

0 HIGH · 2 MEDIUM · 2 LOW · 2 INFO. Mit `make lint` und `make gegenproben` im Ausgangszustand grün; Stempel zu HEAD passend.

## Findings

| # | Kategorie | Quelle | Pfad · Anker | Befund | verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | [AGENTS.md §3.10](../../AGENTS.md), `SPEC-101` Randformen 4 und 7 | `tools/harness/lint.sh` · `# Jedes Verzeichnis tools/<Name> mit eigener go.mod ist ein weiteres Modul.` | Die Schleife über `tools/*/go.mod` ist die neue Zusage des Vertrags (Befund im zweiten Modul färbt das Ziel; Modul ohne Go-Datei meldet Prüfung 4). Keine Gegenprobe und kein Test fängt sie. Mutation: die Schleife ersatzlos gelöscht, `make gegenproben` bleibt `11 Fälle, 0 rot`, Exit 0; die Kopien der Fälle enthalten `tools/gegenproben` nicht (`Eingaben`). Der Plan schließt weitere Gegenproben in §1 aus, §3.10 verlangt je Zusage eine fangende Mutation; beide Texte stehen gegeneinander. Ein Befund im zweiten Modul hält heute nur die von Hand gefahrene Probe fest. | ja (Mutation oben) | Zusage eines neuen Vertrags ohne fangenden Test |
| F-2 | MEDIUM | [AGENTS.md §3.12](../../AGENTS.md), [AGENTS.md §3.10](../../AGENTS.md), `SPEC-102` Randform 8 | `tools/gegenproben/main.go` · `defer func() { err = errors.Join(err, os.RemoveAll(basis)) }()` | Ein Fehler beim Entfernen der Wegwerfverzeichnisse macht jetzt aus einem sonst grünen oder roten Lauf einen Abbruch mit Exit 2, und die Ergebnisse gehen verloren (`run` wertet `err` vor `erg`). Diese Randform steht in `SPEC-102` nicht (Randform 8 nennt nur das Entfernen), sie ist im Code entschieden, nicht davor. Kein Test fängt sie: Mutation `defer func() { _ = os.RemoveAll(basis) }()` lässt `make gegenproben` und die Tests grün. | ja (Mutation oben) | Randform im Code entschieden, ohne Test |
| F-3 | LOW | `SPEC-101` Prüfung 4 („Jedes Modul trägt mindestens eine Datei `*.go`") | `tools/harness/lint.sh` · `find . -name '*.go' -print -quit` | Die Prüfung des Wurzelmoduls sucht ab `.` und findet die Dateien von `tools/gegenproben`. Im Image ohne Go-Dateien unter `src` meldet `lint.sh` den Befund `./: keine Datei *.go im Modul` nicht; das Ziel bleibt nur rot, weil golangci-lint selbst mit `no go files to analyze` endet (Exit 1). Der im Vertrag zugesagte Befund der eigenen Prüfung entfällt für das Wurzelmodul. | ja (`docker run` auf dem Lint-Image, `rm -rf src`) | Eigene Prüfung vom Nachbarmodul maskiert |
| F-4 | LOW | [AGENTS.md §3.9](../../AGENTS.md) | `docs/plan/planning/in-progress/slice-lint-gegenproben-modul.md` · `` `Dockerfile`, `.dockerignore` `` und `` `harness/README.md` `` in der Tabelle von §3 | §3 nennt `.dockerignore` (Allowlist) und `harness/README.md` als geänderte Dateien; im Diff ist keine von beiden berührt (`!tools/gegenproben` stand schon in `.dockerignore`; die Zeile von `make lint` in `harness/README.md` behauptet nichts Überholtes). Der Plan folgt der Umsetzung nicht. | ja (`git diff --stat 66de972..HEAD`) | Slice-Plan zieht der Umsetzung nicht nach |
| F-5 | INFO | `SPEC-101` Randform 7 | `tools/harness/lint.sh` · `for m in tools/*/go.mod; do` | Ein Modul tiefer (`tools/a/b/go.mod`) wird von keinem Lauf geprüft: golangci-lint an der Wurzel überspringt verschachtelte Module, die Schleife sieht nur eine Ebene. Das entspricht dem Vertrag („`tools/<Name>`"); der Vertrag nennt die Grenze nicht. Zuständig: Architect, falls die Form gewollt ist. | ja (`docker run`, Probe `tools/a/b`) | Grenze der Modul-Suche nicht benannt |
| F-6 | INFO | [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) | `.golangci.yml` · `path: ^tools/gegenproben/main\.go$` | Kein Befund: Die Ausnahme hat einen `# Why:`-Block und wirkt nur auf `main.go` (Probe: `os.Stdout` in `tools/gegenproben/probe.go` bleibt rot, mit verändertem Pfad werden `main.go:63` beide Treffer rot). Im Wurzellauf meldet golangci-lint sie als „Skipped 0 issues", im Modullauf trägt sie; `warn-unused` führt nicht zum Fehlschlag (`SPEC-101` Prüfung 3). | ja | — |

## Mutationen und Proben dieses Reviews

| Probe | Ergebnis |
|---|---|
| Global `var Probe = 1` in `tools/gegenproben/faelle.go` | `make lint` rot, `gochecknoglobals` im Modul (Randform 7) |
| `os.Stdout` in neuer Datei des Moduls | rot, `forbidigo`; Ausnahme wirkt nur auf `main.go` |
| Ausnahme-Pfad auf `xmain` verändert | rot, zwei `forbidigo`-Treffer in `main.go` |
| `//nolint` in `main_test.go` des Moduls | rot, Befund von `lint.sh` (Randform 5) |
| Modul ohne Go-Datei (`rm tools/gegenproben/*.go`) | Befund `lint.sh: tools/gegenproben: keine Datei *.go im Modul`, Exit 1 (Randform 4) |
| `tools/gegenproben` fehlt | Exit 0, nur Wurzelmodul (Randform 7) |
| `.golangci.yml` fehlt | Meldung, Exit 2 (Randform 1) |
| Schleife in `lint.sh` gelöscht | `make gegenproben` grün — F-1 |
| `RemoveAll`-Fehler verworfen | `make gegenproben` grün — F-2 |
| Wurzelmodul ohne Go-Dateien | kein Befund der eigenen Prüfung 4 — F-3 |

Arbeitsbaum nach den Proben: `git checkout` der Quellen und von `harness/gegenproben.stempel`; `git status --short` leer.

## Geprüft, ohne Befund

* `tools/gegenproben/main.go`, `faelle.go`, `main_test.go`: Verhalten des Umbaus unverändert (Reihenfolge `legeAn`, `haengeAn`, `entferne`, `ersetze` wie zuvor; `ohneFehlendesGoSum` ersetzt die Bedingung gleichwertig; Tests führen `Umgebung` statt Globals, `HashEingabenGoSumIstOptional` fängt das fehlende `go.mod`). Die aus zwei Teilen gebildete Zeichenkette `//nolint` trifft den Suchtext von `lint.sh` nicht mehr.
* `Dockerfile` Stufe `lint`: `COPY tools/gegenproben tools/gegenproben` genügt, `.dockerignore` lässt das Verzeichnis durch; Relativpfad `../../$profil` trägt für `tools/<Name>`.
* Eingaben der Gegenproben: `tools/gegenproben`, `Dockerfile`, `.dockerignore`, `.golangci.yml`, `tools/harness/lint.sh` stehen in `tools/harness/gegenproben-eingabe.sh`; der Stempel wurde neu geschrieben und passt zu HEAD. Die Kopien der Fälle (`Eingaben` im Programm) enthalten das Modul nicht; das ist folgerichtig, begründet aber F-1.
* `SPEC-101`, `SPEC-102`: Randformen 1–8 stehen vor dem ersten Code-Commit (`70a2001` vor `fdbd23b`); Kommentare in den Änderungen tragen Zusage, Kopplung oder Abgrenzung (`AGENTS.md` §3.7), keine verworfene Alternative. Keine Lockerung eines Gates (§3.6), keine Direktive (§3.2).
* `docs/plan/planning/done/slice-gegenproben.md`, `roadmap.md`: nur Verweisnachzug und Marker.

## Hinweise an andere Rollen

* Verifier: DoD 1–3 (Befundfreiheit, beide Module unter `make lint`, Spec-Nachzug) und `make gates` sind seine Frage; dieser Review hat sie nicht geprüft.
* Konflikt-Pfad: F-1 stellt Plan §1 („Weitere Gegenproben" ausgeschlossen) gegen `AGENTS.md` §3.10; keine Rollen-Auseinandersetzung bisher, die Auflösung gehört dem Architect/Planner, falls der Implementer widerspricht.

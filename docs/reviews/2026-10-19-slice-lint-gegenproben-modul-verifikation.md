# Verifikation slice-lint-gegenproben-modul

* Rolle: Verifier (Modul 11), frischer Kontext; berichtet, repariert nichts.
* Plan: [slice-lint-gegenproben-modul](../plan/planning/in-progress/slice-lint-gegenproben-modul.md)
* Diff: `66de972` (Move nach `in-progress/`) bis `HEAD` = `6b702f9`.
* Bezug: [LH-QA-06](../../spec/lastenheft.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md); Verträge [SPEC-101](../../harness/vertraege/SPEC-101-lint-profil.md), [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md)
* Vorgänger: [Review](2026-10-19-slice-lint-gegenproben-modul-review.md); F-1 bis F-5 daraus im Stand `6b702f9` behoben, F-6 INFO.

## Urteil

**Liefer-Punkte DoD 1 bis 4 und die technischen Closure-Trigger aus §5 sind erfüllt. Der Slice ist noch nicht schließbar:** zwei MEDIUM-Befunde (V-1, V-2) und ein LOW (V-3) betreffen Zusagen und Plan-Stand; die Closure-Pflichten (Review-Haken, Closure-Notiz, Register, Risiko-Ausgänge) stehen erwartungsgemäß noch aus. 0 HIGH.

## Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis |
|---|---|
| `make gates` | Exit 0 (Arbeitsbaum vorher und nachher sauber) |
| `make lint` ungecacht (`--no-cache-filter lint`) | zwei Läufe von golangci-lint, beide `0 issues`: Wurzel und `tools/gegenproben`; die Ausnahme für `main.go` erscheint im Modullauf in der Regelliste |
| `make gegenproben` ohne Stempel, 4 Läufe nacheinander | je Exit 0, `14 Fälle, 0 rot`, 11,4 bis 11,7 s (Budget 30 s); keine Flakiness; der Stempel wird bytegleich neu geschrieben (`git status` leer) |
| `make gegenproben` mit Stempel | `Eingaben unverändert … nichts läuft neu`: Stempel passt zu HEAD |
| Probe aus §5: `func unbenutzt() {}` an `tools/gegenproben/faelle.go` | `make lint` rot (`faelle.go:133:6: func unbenutzt is unused (unused)`, Exit 1 der Stufe); danach `git checkout` |

## DoD-Konformität

| DoD | Urteil | Beleg |
|---|---|---|
| 1 Programm ohne Befund; Globals zu Umgebung, `RemoveAll` geprüft, `greife` zerlegt, `go.sum`-Bedingung ausgelagert, `//nolint` aus zwei Teilen | erfüllt | `Umgebung`/`standard()` in `main.go`; `errors.Join(err, u.Aufraeumen(basis))`; `greife` ruft `legeAn`, `haengeAn`, `entferne`, `ersetze`; `ohneFehlendesGoSum`; `"\n//" + "nolint…"` in `faelle.go`; `make lint` 0 issues |
| 2 beide Module unter `make lint`; Ausnahme für `main.go` mit `# Why:` | erfüllt | `Dockerfile` Stufe `lint` kopiert `tools/gegenproben`; Schleife in `lint.sh`; Ausnahme in `.golangci.yml` mit `# Why:`-Block, ohne den Block meldet `lint.sh` `.golangci.yml:203` |
| 3 SPEC-101/-102 ziehen nach; `make gegenproben` grün | erfüllt | Grenze „außerhalb von `make lint`" entfällt in beiden; Prüfung 4, Randformen 4 und 7 in SPEC-101; Fälle 12 bis 14 und Randform 8 in SPEC-102; `harness/README.md` ohne „elf"; Suche nach „außerhalb von make lint" findet nur Plan und Closure-Text des erledigten Slice |
| 4 `make gates` grün | erfüllt | siehe oben |

Closure-Trigger §5: `make gates` grün mit beiden Modulen, Probe-Befund macht `make lint` rot, `make gegenproben` grün: erfüllt. Offen bis zur Closure: Closure-Notiz mit Lerneintrag (§7 trägt „Wird bei der Closure geschrieben."), Review-Haken, Register-Fortschreibung, Risiko-Ausgänge, Doku-Haken; alle DoD-Häkchen stehen auf `[ ]`.

## Mutationen

Alle mit `rm -f harness/gegenproben.stempel` und jeweils zurückgenommen (`git status` am Ende sauber).

| Zusage | Mutation | Ergebnis |
|---|---|---|
| Modul-Schleife (SPEC-101 Teil 2, Randform 7) | Schleife ersatzlos wirkungslos (`/nonexistent/*/go.mod`) | rot: `lint/befund-im-zweiten-modul: Verstoß nicht gemeldet` |
| Befund im zweiten Modul färbt das Ziel | `golangci-lint` im Modul `\|\| true` | rot, derselbe Fall |
| Prüfung 4, Wurzelmodul | Prüfung gelöscht | rot: `lint/ohne-go-datei: Ausgabe nennt "keine Datei *.go" nicht` |
| Prüfung 4, Modul unter `tools/` | Prüfung im Modul-Zweig durch `:` ersetzt | **grün (14 Fälle, 0 rot)** — V-1 |
| Prüfung 4, Wurzelmodul außerhalb `tools/` (F-3) | `-path ./tools -prune` entfernt | **grün** — V-1 |
| Ausnahme `main.go` | Pfad auf `xmain` | `make lint` rot, zwei `forbidigo`-Treffer |
| Ausnahme `main.go` | Regel gelöscht | `make lint` rot, zwei `forbidigo`-Treffer |
| Ausnahme `main.go` | `# Why:`-Block entfernt | `make lint` rot, Befund von `lint.sh` |
| Ausnahme nur für `main.go` | Pfad auf `^tools/gegenproben/` verbreitert | `make lint` grün — kein Test hält die Enge; im Vertrag nicht zugesagt, INFO V-4 |
| Randform 8 (Aufräum-Fehler) | Fehler verworfen (`_ = u.Aufraeumen`) | rot: `TestRunMeldetEntfernenFehlerNachDenErgebnissen` |
| Aufräumen | Aufruf entfernt | rot: `TestRunRaeumtWegwerfverzeichnisseAuf` und der Test zu Randform 8 |
| Randform 8 (Exit 2 nach den Ergebnissen) | Zweig `fehler != nil` entfernt | rot: Test zu Randform 8 |
| Umgebung `Programme` | Prüfung übersprungen | rot: `TestRunProgrammFehlt` |
| Umgebung `Befehl` | `standard().Befehl` statt `u.Befehl` | rot: vier Tests |
| Umgebung `Eingaben` im Standard | `tools/harness/lint.sh` entfernt | `make gegenproben` rot, alle Lint-Fälle Exit 2 |
| `go.sum` optional | Bedingung auf jedes `ErrNotExist` erweitert | rot: `TestRunEingabeFehlt`, `TestHashEingabenGoSumIstOptional` |
| Umgebung `Eingaben` in `run` | `standard().Eingaben` statt `u.Eingaben` | grün; äquivalent, weil Tests dieselbe Liste führen — kein Befund |

## Findings

| # | Stufe | Quelle | Anker | Befund |
|---|---|---|---|---|
| V-1 | MEDIUM | [AGENTS.md §3.10](../../AGENTS.md), SPEC-101 Prüfung 4 | `tools/harness/lint.sh` · `find "$d" -name '*.go'` und `find . -path ./tools -prune` | Prüfung 4 ist eine Zusage dieses Slice; zwei ihrer Teile hat keine Mutation rot. (a) Ein Modul unter `tools/` ohne Go-Datei: Fall 14 löscht die Dateien von `src/`, also trifft er nur das Wurzelmodul. Ersetzt man die Modul-Prüfung durch `:`, bleibt `make gegenproben` bei 14 Fällen, 0 rot. (b) Der Ausschluss `tools` in der Wurzelprüfung (die Korrektur zu Review F-3): ohne ihn bleibt alles grün, weil die Kopien der Fälle `tools/` nicht enthalten; im Lint-Image mit dem Modul unter `tools/` fände `find .` dessen Dateien. Der Ausschluss ist damit nur im Handlauf belegt. Ein dritter Fall (Modul `tools/probe` mit `go.mod` ohne Go-Datei, Exit 1, Text `keine Datei *.go`) und einer für Wurzel ohne `src`-Dateien bei vorhandenem zweiten Modul fingen beide. |
| V-2 | LOW | [AGENTS.md §3.9](../../AGENTS.md) | Plan §6, zweites Risiko: „die elf Fälle und die Tests des Programms bleiben unverändert grün" | Der Stand ist 14 Fälle (drei neue in `faelle.go`), zwei Fälle ändern die Direktiv-Zeichenkette, die Tests sind auf `Umgebung` umgestellt und um einen Test erweitert. §6 beschreibt den Umbau weiter als verhaltens- und fallgleich; §3 nennt zwar `faelle.go`, §1 die neuen Fälle, §6 zieht nicht nach. Der Ausgang muss bei der Closure nach dem tatsächlichen Stand lauten (Verhalten unverändert, Fälle 12 bis 14 neu, begründet in §1). |
| V-3 | LOW | [AGENTS.md §3.5](../../AGENTS.md) | [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) Status `Accepted`, Tabellenzeile `tools/gegenproben` | Der Commit `fdbd23b`/`6b702f9` ändert den Text einer angenommenen ADR („elf Fälle für `a-check` und `lint.sh`" zu „die Fälle für …"). Inhaltlich eine Zahlenbereinigung ohne Wirkung auf Entscheidung oder Gründe, formal eine Änderung am immutablen Dokument ohne `Supersedes`; der Plan §3 führt die ADR nicht unter den berührten Dateien. Zuständig: Planner entscheidet, ob die Tatsachenberichtigung zulässig ist und im Plan zu nennen ist. |
| V-4 | INFO | [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) | `.golangci.yml` · `path: ^tools/gegenproben/main\.go$` | Die Enge der Ausnahme (nur `main.go`) hält kein Test; verbreitert auf das Verzeichnis bleibt `make lint` grün. Die Enge ist weder in SPEC-101 noch SPEC-102 zugesagt, nur im Review von Hand geprüft (dort F-6). Keine Forderung. |
| V-5 | INFO | Zeitbudget SPEC-102 | `time make gegenproben` | 11,4 bis 11,7 s bei erzwungenem Lauf, unter 30 s; der Vertragstext „gemessen am Stand der Lieferung: 13,5 s" nennt den alten Stand mit elf Fällen. Kein Verstoß. |

Die Review-Befunde F-1 bis F-5 sind im Stand `6b702f9` wirksam behoben, soweit eine Mutation es zeigt: F-1 (Schleife) durch Fälle 12/13 gefangen, F-2 (Randform 8) durch Test und SPEC-102, F-4 im Plan nachgezogen; F-3 ist im Code behoben, aber nur teilweise gefangen (V-1).

## Plan-vs-Code-Diff (`66de972..HEAD`)

| Plan §3 | Code | Abweichung |
|---|---|---|
| `main.go`, `faelle.go`, `main_test.go` Refactor | umgesetzt | zusätzlich drei neue Fälle und ein neuer Test (laut §1 gedeckt, §6 nicht nachgezogen: V-2) |
| `tools/harness/lint.sh` je Modul | umgesetzt | Prüfung 4 und Modul-Schleife über die Planzeile hinaus; im Vertrag SPEC-101 vor dem Code (`70a2001`) entschieden |
| `Dockerfile` | umgesetzt (`COPY tools/gegenproben`) | keine; `.dockerignore` ist im Plan gestrichen, im Diff unberührt |
| `.golangci.yml` | umgesetzt | keine |
| SPEC-101, SPEC-102, `harness/README.md` | umgesetzt | keine |
| nicht im Plan | [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Zeile geändert), `slice-gegenproben.md` (Verweis auf `in-progress/`), `roadmap.md` („Nichts in Arbeit." entfernt), `harness/gegenproben.stempel` | ADR: V-3; die übrigen sind erwartete Nachzüge; Review-Report und dieser Bericht |

Randformen vor dem Code: SPEC-101 (`70a2001`) liegt vor dem Umsetzungs-Commit (`fdbd23b`); Randform 8 von SPEC-102 kam nach Review im Fix-Commit, ist aber mit Test belegt.

## Hinweise an den Planner

* Vor der Closure: V-1 durch zwei Fälle (oder Streichen der Zusage aus SPEC-101 mit Begründung), V-2 im Plan §6, V-3 entscheiden.
* Closure-Pflichten sind noch alle offen: Closure-Notiz mit Lerneintrag, Register (`BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` trifft wiederum zu: V-1 ist eine Zusage ohne fangenden Test), Risiko-Ausgänge, Haken der DoD.

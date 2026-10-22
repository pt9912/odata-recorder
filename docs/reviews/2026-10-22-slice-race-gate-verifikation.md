# Verifikation: slice-race-gate

* Datum: 2026-10-22
* Rolle: verifier (frischer Kontext; kein Code, kein Vertrag, kein Plan, keine `AGENTS.md` geändert; `.claude/agents/architect.md` ist eine fremde, uncommittete Änderung und blieb ungelesen und unberührt)
* Gegenstand: Spitze `a8889a2`, Stempel `.harness/state/gates-passed.head` gleich HEAD. Basis der Größenmarke `73e123e`; Implementer `4acb295` bis `92668e6`, Architect `cb49d16`, `790122c`, Reviewer `96b6558`, Architect `fa42e09`, Implementer `e13107d`, `0ec39a1`, `a8889a2`.
* Plan: `docs/plan/planning/in-progress/slice-race-gate.md`
* Maßstab: DoD (§2), Tabelle *Zusage · Fänger* (§3, Z1 bis Z11, N1 bis N3) und Raster (§6) gegen `SPEC-122` Nr. 1 bis 10, `SPEC-102` (Fälle 34 bis 36, Aufbau, Randform 6, Zeitbudget), `SPEC-036`; [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md); [LH-QA-03](../../spec/lastenheft.md#lh-qa-03--isolation-und-konsistenz), [LH-QA-06](../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit); `AGENTS.md` §3.1 bis §3.16. [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md) (Accepted) und das Zeitbudget 120 s sind Entscheidungen des Nutzers und nicht Gegenstand.
* Hinweis zum Skill: `.harness/skills/closure-note-reviewer.md` gilt für Slices in `done/`; dieser Slice liegt in `in-progress/`, der Skill passt nicht auf diese Rolle und wurde nicht angewandt.
* Belege selbst gefahren (Abschnitt 4): `make test-race` erzwungen, `make gegenproben` mit Scratch-Stempel, `make docs-check`, Mutationen in Wegwerfkopien von `git archive HEAD`, Raten je Prozess im Image von `RACE_IMAGE`. Der Repo-Baum und `harness/gegenproben.stempel` blieben unberührt (`git status --short` zeigt nur `.claude/agents/architect.md`; `sha256sum` des Stempels vor und nach dem Lauf `c1f1d04c…`).

## Urteil

**Freigegeben: kein HIGH, kein MEDIUM, kein LOW, drei INFO.** Jede DoD-Zeile, die der Code erfüllen kann, ist belegt; die Zeilen der Closure sind offen und im Plan als Aufgabe des Planners führbar (Abschnitt 1). F-1 bis F-6 und I-2 des Reviews sind im Stand geschlossen (Abschnitt 3). Jede von mir gefahrene Mutation ist aus dem benannten Grund rot oder ein vom Vertrag benanntes Negativ (Abschnitt 5).

* INFO: V1 (Planner) die Zahlen ohne Detektor in `client_test.go` sind Einzelmessungen einer lastabhängigen Rate; V2 (Planner) `make gegenproben` brauchte bei mir 94,5 s, über dem im Vertrag genannten Höchstwert 93,8 s, unter dem Budget 120 s; V3 (Implementer, Planner) die Fehlermeldung `Kopie für %s` in `fahre` nennt den Fall ohne Kopie nicht (der Planner führt sie schon als offen).

## 1. DoD Punkt für Punkt (Plan §2)

| # | DoD | Stand | Beleg |
|---|---|---|---|
| 1 | Ziel mit Race-Detektor: ADR Accepted und im Index, `SPEC-122` vor dem Code, Stufe `race` auf gepinntem Image, Umgebung wie `deps` und `lint`, `RUN --network=none`, `make test-race` an `GATE_CHECKS`, Zeile in `harness/README.md` §Sensors, übrige Stufen ohne cgo | erfüllt | `docs/plan/adr/README.md:14` ([ADR-0007](../plan/adr/0007-race-detektor-im-gate.md), Accepted); `SPEC-122` im Index `harness/vertraege/README.md:31`; `Dockerfile:55-65` (Stufe `race`, `ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly`, zwei `RUN --network=none`); `harness/mk/go-gates.mk:10,20,23` (Pin mit Digest, Ziel, `GATE_CHECKS += build test test-race`); `harness/README.md:69`; `make test-race` erzwungen: `pruefung race: ok` nach 66,0 s, rc 0, Wanduhr 78,7 s; Z10 gemutet: `ENV CGO_ENABLED=1` in `deps`, `source`, `test`, `build` rot mit `cgo: C compiler "gcc" not found` |
| 2 | Fänger des Ziels: Fälle 34 bis 36; jede Zusage hat ihren Fänger (Tabelle); heutige Pakete grün oder Befund entschieden | erfüllt | Abschnitt 2 und 5; `make gegenproben`: `gegenproben: 36 Fälle, 0 rot` mit `race/ohne-rennen`, `race/rennen-in-jedem-teilbaum`, `race/rennen-und-zeitgrenze` ok; `src/adapters/driven/httpupstream/` läuft unter `-race` grün (Teil von `./...`, `pruefung race: ok`) |
| 3 | Regel und Nachzug (`AGENTS.md` §3.10 im Wortlaut des Entwurfs; `SPEC-114` bis `SPEC-121` *Ungeprüft*; Pläne; Kommentare in Tests) | Nachzug erfüllt, `AGENTS.md` offen für die Closure | Suche nach `-race`, `Race-Detektor`, `Datenrennen`, `CGO_ENABLED` in `spec/`, `src/`, `harness/`, `tools/`, `docs/plan/` außerhalb von `done/`, `.claude/commands/`, `.harness/skills/`, `README.md`, `AGENTS.md`: kein Satz „das Gate fährt kein `-race`“ mehr; `spec/spezifikation.md:1426` (Historie) und die *Ungeprüft*-Sätze nennen `make test-race`; Kommentare der drei Tests im Stand (Abschnitt 3). `AGENTS.md` §3.10 ändert nur der Nutzer; der Wortlaut steht als Entwurf in Plan §3, die Voraussetzung (Ziel `make test-race` existiert, `AGENTS.md` §4) ist erfüllt |
| 4 | `make gates` grün, `make gegenproben` grün | erfüllt | `make gegenproben GEGENPROBEN_STEMPEL=<Scratch>`: rc 0, 36 Fälle, 0 rot, Wanduhr 94,5 s; die Prüfsumme (`93ea7aeb…`) gleicht dem versionierten Stempel, also gehört er zu den Eingaben von HEAD; `make gates`: siehe Abschnitt 4 (nach dem Commit des Berichts) |
| 5 | Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-10-22-slice-race-gate-review.md` (`96b6558`): 0 HIGH, 4 MEDIUM, 2 LOW, 4 INFO; nachverifiziert in Abschnitt 3 |
| 6 | Doku-Update (Vertrag, `harness/README.md` §Sensors und *Nicht behauptet*, ADR-Index) | erfüllt | `harness/README.md:69,86`; ADR-Index; `make docs-check`: 193 Dateien, 0 Befunde |
| 7 | Closure-Notiz mit Steering-Loop-Lerneintrag; Ausgang von `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` auf `verkörpert` | offen für die Closure | §7 trägt „Wird bei der Closure gefüllt“; Voraussetzungen erfüllt (Ziel liegt, Finding-Klassen des Reviews und dieses Berichts liegen vor) |
| 8 | Beobachtungs-Register fortgeschrieben | offen für die Closure | Belege: Finding-Klassen im Review („Allaussage über eine Menge ohne Fänger je Element“, „Fänger prüft den Mechanismus, nicht seine Verwendung“, „Kommentar nennt die Rate einer anderen Mutation“) |
| 9 | Jedes Risiko aus §6 trägt einen Ausgang | offen für die Closure (acht Ausgänge) | *offen* tragen: Rennen im heutigen Code, Laufzeit, Kosten, Nachzug; *weiter offen*: Wortlaut der Regel; *entfallen* mit Beleg: Image, Plattform, Gegenprobe, Zeit von `make gegenproben` (*eingetreten*), Pakete, Klasse der Zeilen, Marke, Randformen. Messwerte für Rennen im heutigen Code, Laufzeit und Kosten liegen vor (Plan §6; mein `make test-race`: Stufe 66,0 s) |
| 10 | Drei Paarungen getragen | offen für die Closure | von der nächsten Welle-Closure geprüft |

Ein Häkchen setze ich nicht.

## 2. Tabelle *Zusage · Fänger* und Vertrag

Jede der Zusagen `SPEC-122` Nr. 1 bis 10 und `SPEC-102` Fälle 34 bis 36 (Aufbau, Zeitbudget, Randform 6) hat eine Zeile (Z1 bis Z11, N1 bis N3, Zusatzzeilen); jede Zeile mit Fänger nennt einen Fänger, der im Code existiert:

* Z1 bis Z4: `TestRaceOhneBefund`, `TestRaceMuster`, `TestRaceSetztCgo`, `TestRaceBefund`, `TestRaceBefundDurchSignal`, `TestStroeme`, `TestAufrufFehlerStartenNichts`, `TestWurzelFehlt`, `TestProgrammNichtAufrufbar`, `TestSchreibtNichts` in `tools/pruefung/main_test.go`.
* Z5 bis Z8: Fälle 34 bis 36 in `tools/gegenproben/faelle.go` (`raceFaelle`); Namen, Gate, Argumente, Erwartung gleichen `SPEC-102`.
* Z9: benannte Ausgaben der Stufe (Abschnitt 5, alle fünf Meldungen selbst gesehen).
* Z10: benannte Ausgabe von `make test` und `make build` für `deps`, `source`, `test`, `build`; die Stufe `pruefung` ist als akzeptiertes Negativ mit Grund und Folge geführt (`SPEC-122` Nr. 9, N3).
* Z11: `TestFaelleDerPruefungHabenDieNamenAusDemVertrag`, `TestWegwerfmodulKopiertDieEingabenNicht`, `TestRaceFaelleLegenEinWegwerfmodulAn`, `TestRaceFaelleFuehrenNurASCII`.

Vertrag gegen Code: `pruefeRace` (`tools/pruefung/main.go:167-187`) führt `go test -race -count=1 -trimpath -buildvcs=false <Muster>`, setzt `CGO_ENABLED=1` nur für den Aufruf, schreibt den Befund `go test -race:` samt beiden Strömen nach stderr (Exit 1), `pruefung race: ok` nach stdout; kein `-timeout`, kein `-failfast`, kein `gofmt`/`go vet`; `Wurzel fehlt` und `go` nicht aufrufbar sind Exit 2. Die Imports von `tools/pruefung` (`bytes context errors fmt io os os/exec path/filepath strings testing`) gleichen der Liste in `SPEC-122` Nr. 9. Das Raster in §6 hat keine leere Zelle, keine Zelle ist im Code entschieden.

## 3. Nachverifikation der Befunde des Reviews

| Befund | Stand | Beleg | Ergebnis |
|---|---|---|---|
| F-1 (Z10, Stufe `pruefung`) | anders gelöst: akzeptiertes Negativ mit Grund und Folge | `SPEC-122` Nr. 4 begrenzt die Zusage auf `deps`, `source`, `test`, `build`; Nr. 9 und N3 führen `pruefung` mit Grund (nur Standardbibliothek ohne cgo, Imports verglichen) und Folge; selbst gemutet: `ENV CGO_ENABLED=1` in `pruefung` lässt `make test` (`pruefung test: ok`) und `make build` (`pruefung build: ok`) grün, in den vier anderen Stufen rot | geschlossen |
| F-2 (i) `Wegwerfmodul` je Fall | `TestRaceFaelleLegenEinWegwerfmodulAn` (`main_test.go:308-315`) | `Wegwerfmodul: false` in Fall 34, 35, 36 je rot mit dem Namen des Falls; `if false && f.Wegwerfmodul` in `legeVerzeichnisAn` rot an `TestWegwerfmodulKopiertDieEingabenNicht` | geschlossen |
| F-2 (ii) Feld `Env` | gestrichen | `grep Env` in `tools/`, `harness/`, `spec/`, `Dockerfile`: nur `cmd.Env` von `os/exec` (`tools/gegenproben/main.go:258`, `tools/pruefung/main.go:61`); Plan nennt `Env` nur als entfallen (§3 Zeile `tools/gegenproben/`, Auftrag 3); `SPEC-102` Aufbau sagt, die Fälle erben `0` und die Stufe setzt es, was kein Fall fängt | geschlossen |
| F-3 Ende durch Signal | `TestRaceBefundDurchSignal` (`main_test.go:~311`) | `l.code > 0` rot, ebenso `starte`, das ein Signal als Code 0 meldet; Zeilen Z3 und Zusatztabelle und `SPEC-122` Nr. 1, 3 nennen den Fall | geschlossen; `pruefeTest` hat dieselbe Form ohne Test (offen für die Closure) |
| F-4 Kommentar A15 | `client_test.go:42-48` nennt die Rate je Mutation, mit und ohne Detektor | gemessen (Abschnitt 5): Mutation 1 mit Detektor 200/200 rot, ohne 170 bis 192/200; Mutation 2 mit 200/200, ohne 9 bis 17/200 gegen die genannten 200, 197, 14 | geschlossen; V1 |
| F-5 Kommentar B19 | `anhaengen_test.go:79-82`, „hoch“ ersetzt durch 100 von 100 | gemessen 100/100 mit Detektor (100 mit `DATA RACE`) und ohne; unmutiert 0/100 | geschlossen |
| F-6 „Kopie“-Kommentare | `faelle.go` (`Gate`, `Fall`, `Neu`, `Argumente`), `main.go` (`befehl`, `fahre` samt Rumpf, `fuehreFall`) im Wortlaut des Auftrags | `git diff fa42e09..HEAD -- tools/gegenproben` | geschlossen; Rest V3 |
| I-1 (Bericht des Implementers lag dem Review nicht vor) | die Raten stehen im Plan §3 *Rate* mit Herkunft | meine Messungen decken sich in der Größenordnung (Fixtures 200/200 nach Plan; Abschnitt 5) | geschlossen |
| I-2 (B8, Seite von `Drain`) | `wahl_test.go:410-423` sagt nicht mehr zu, der Test fange die Seite von `Drain` | Mutation a: 45/60 rot mit, 13/60 ohne Detektor, nie `DATA RACE`; Mutation b: 0/60 mit und ohne; unmutiert 0/30; der *Ungeprüft*-Satz in `SPEC-121` zur Drain-Seite stimmt nicht (offen für die Closure) | Kommentar geschlossen |
| I-3 Größe | siehe Abschnitt 6 | | Messung liegt vor |
| I-4 Fallfrist bei zwei Läufen zugleich | `SPEC-102` Randform 6 nennt es („ein Lauf zugleich je Host“) | | geschlossen |

Weitere Prüfungen: der Diff `git diff a7def59..HEAD -- src` besteht aus drei Dateien und nur aus Kommentarzeilen (19 Einfügungen, 8 Löschungen); der Diff seit `fa42e09` unter `src/` ebenso. Die Zahl der Fälle ist überall 36 (`SPEC-102` Zeitbudget, Plan §6, `faelle()` mit 33 + 3, Ausgabe `36 Fälle, 0 rot`); keine veraltete Zahl gefunden. Die Kommentare in `tools/gegenproben/` nennen das Verzeichnis des Falls; `raceFaelle` sagt zu, was `TestRaceFaelleLegenEinWegwerfmodulAn` fängt, und nennt als ungefangen, dass die Stufe `CGO_ENABLED` setzt.

## 4. Gates und Läufe

| Lauf | Ergebnis |
|---|---|
| `make test-race DOCKER_BUILD='docker build --progress=plain --no-cache-filter race'` | `#21 65.93 pruefung race: ok`, rc 0, Wanduhr 1 m 18,7 s |
| `make gegenproben GEGENPROBEN_STEMPEL=<Scratch>` | `gegenproben: 36 Fälle, 0 rot`, rc 0, Wanduhr 1 m 34,5 s; versionierter Stempel unverändert, Scratch-Stempel gleich ihm (`93ea7aeb…`) |
| `make docs-check` | `d-check: 193 Datei(en) geprüft, 0 Befund(e)`, rc 0 |
| `make gates` | nach dem Commit dieses Berichts (Rückgabecode und Stempel stehen in der Antwort an den Aufrufer) |

Auf dem Host lief während meiner Läufe ein fremder Container (`u-boot:test-docker-tools`); er blieb unberührt und erklärt die Lastabhängigkeit der Raten ohne Detektor (V1, V2).

## 5. Eigene Mutationen

Alle in Wegwerfkopien von `git archive HEAD` unter `scratchpad/verify-race-gate/work/`, je als Patch mit `git apply`; Unit-Ebene im Image der Stufe `deps` (`go test`), Stufen mit `docker build --target …`, Raten im Image von `RACE_IMAGE` (`go test -c`, danach je Lauf ein Prozess mit `-test.count=1`, `-test.list` bestätigt, dass die Regex einen Test trifft).

| Zeile | Mutation | Ergebnis |
|---|---|---|
| Z3 (neu) | `l.code > 0` statt `!= 0` | rot `TestRaceBefundDurchSignal` |
| Z3 (neu) | `starte` setzt den Code nur bei `> 0` (Signal gilt als 0) | rot `TestRaceBefundDurchSignal` |
| Z3 | Ok-Zeile nach stderr | rot `TestRaceOhneBefund`, `TestStroeme/race_ok` |
| Z1 | `-race` entfernt; `-count=1` entfernt; `-failfast` ergänzt | je rot `TestRaceOhneBefund`, `TestRaceMuster` |
| Z2 | Umgebung unverändert (nil) | rot `TestRaceSetztCgo/vorher_'0'` und `vorher_''` |
| Z4 | Wurzel als Datei angenommen; `go` nicht aufrufbar als Exit 1 | rot `TestWurzelFehlt`; rot `TestProgrammNichtAufrufbar/race_ohne_go` |
| Z9 | Build-Argument weggelassen; `RACE_IMAGE=` leer; Image von `deps`; `golang:1.27.3`; `debian:bookworm-slim` | rot mit `stat /bin/sh: no such file or directory`; `base name (${RACE_IMAGE}) should not be blank`; `cgo: C compiler "gcc" not found`; `Go-Version von race (go1.27.3) weicht von deps (go1.27.1) ab`; `head: cannot open '/usr/local/go/VERSION'` |
| Z10 (neu: vier Stufen) | `ENV CGO_ENABLED=1` in `deps`, `source`, `test` (Ziel `test`), `build` (Ziel `build`) | je rot, `cgo: C compiler "gcc" not found` |
| Z11 (neu) | `Wegwerfmodul: false` in Fall 34, 35, 36 | je rot `TestRaceFaelleLegenEinWegwerfmodulAn` mit dem Namen des Falls |
| Z11 | Mechanismus ignoriert (`if false && f.Wegwerfmodul`) | rot `TestWegwerfmodulKopiertDieEingabenNicht` |
| N3 | `ENV CGO_ENABLED=1` in `pruefung`, Ziele `test` und `build` | grün (`pruefung test: ok`, `pruefung build: ok`), wie der Vertrag sagt |
| N3 | `ENV CGO_ENABLED=1` in `lint` / in `gegenproben` | grün (`0 issues`; `36 Fälle, 0 rot`), wie der Vertrag sagt |
| Kommentar A15 | Mutation 1 (Header-Map am Client als Header benutzt) | 200/200 rot mit Detektor (200 mit `DATA RACE`); ohne 170/200 unter Last, 192/200 danach; unmutiert 0/200 mit und ohne |
| Kommentar A15 | Mutation 2 (Map nur beschrieben) | 200/200 mit Detektor (200 `DATA RACE`); ohne 9/200 und 17/200 |
| Kommentar B19 | Sperre in `haengeAn` entfernt | 100/100 mit (100 `DATA RACE`) und 100/100 ohne; unmutiert 0/100 |
| Kommentar B8 | a: `Unlock`/`Lock` in `nimmAn` vor `r.zustand` | 45/60 mit, 13/60 ohne Detektor, nie `DATA RACE`; unmutiert 0/30 |
| Kommentar B8 | b: `Unlock`/`Lock` in `sperre` nach `sz.drain = true` | 0/60 mit und 0/60 ohne, wie der Kommentar sagt |

Grün geblieben sind allein N3 (vom Vertrag als Negativ geführt) und B8 b (der Kommentar sagt, der Test fange es nicht). Nicht neu gefahren: Z5 bis Z8 auf Fallebene; der Code von `pruefeRace` und `starte` ist seit dem Review unverändert (`git diff fa42e09..HEAD -- tools/pruefung` berührt nur `main_test.go`), und die Reviewer-Mutationen z5 bis z8 gelten damit weiter.

## 6. Größe (§3.14)

`git diff --numstat 73e123e..HEAD -- Dockerfile harness/mk tools/pruefung tools/gegenproben src`: 413 Einfügungen, 62 Löschungen, **475** (Reviewer beim Stand `790122c`: 422). Marke 800, Schätzung 620. Die Pfade der Marke sind vollständig (jede Datei des Diffs unter diesen Pfaden steht in Plan §3). Die Messung nach dem Review trägt der Planner in §4 ein; 475 liegt unter der Schätzung.

## 7. Sendungen (§3.13)

`slice-bootstrap-m2` liegt in `open/` (nicht in `done/`); die Sendung (`tests/` in die Stufe `source` und den Build-Kontext) steht in dessen §1 *Übernimmt (Quellstand)*, DoD und §3 (`docs/plan/planning/open/slice-bootstrap-m2.md:43,71,95`) mit Verweis auf `slice-race-gate`; sein *Ausdrücklich NICHT* trifft sie nicht (Plan §6, *Pakete*). `slice-http-adapter` trägt die Sendung in §1 (Zeile 40) und DoD (Zeile 69). Keine Zuweisung an einen Slice in `done/`.

## Befunde

| ID | Kategorie | Befund | Quelle | Pfad | Zuständig |
|---|---|---|---|---|---|
| V1 | INFO | Der Kommentar von `TestGleichzeitig` nennt Einzelmessungen (200 und 197, 200 und 14 von 200). Die Rate ohne Detektor schwankt mit der Last: Mutation 1 170 und 192, Mutation 2 9 und 17 von 200 in meinen Läufen. Der Satz sagt „Gemessen“ und bleibt wahr; wer ihn als feste Rate liest, irrt um bis zu 15 Punkte. Kein Fänger hängt an der Zahl. | `AGENTS.md` §3.10, §3.11 | `src/adapters/driven/httpupstream/client_test.go:42-48` | Planner (ob die Closure eine Spanne statt eines Wertes verlangt) |
| V2 | INFO | `SPEC-102` Zeitbudget nennt 93,8 s als höchsten gemessenen Wert („rund 28 % darunter“); mein Lauf brauchte 94,5 s (Last durch einen fremden Container). Das Budget 120 s hält. | `SPEC-102` Zeitbudget | `harness/vertraege/SPEC-102-gegenproben.md:84-89` | Planner (nur Kenntnisnahme) |
| V3 | INFO | `fahre` meldet einen Fehler beim Anlegen als `Kopie für %s`, auch für ein Wegwerfmodul; zwei weitere Kommentare mit „Kopie“ nennen den Fall ohne Kopie nicht. Eine Ausfallfolge gibt es nicht (die Meldung erscheint nur bei einem Fehler des Dateisystems). | `AGENTS.md` §3.7 | `tools/gegenproben/main.go` · `fmt.Errorf("Kopie für %s: %w", …)` | Implementer (folgt dem Planner, schon als offen geführt) |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `Dockerfile` (Stufe `race`, Kommentare), `harness/mk/go-gates.mk`, `harness/mk/gegenproben.mk` | geprüft, ohne Befund; Pin und Digest gleichen `SPEC-122` Nr. 4; fünf Z9-Varianten und vier Z10-Stufen rot mit den benannten Meldungen |
| `tools/pruefung/main.go`, `main_test.go` | geprüft, ohne Befund; 11 Mutationen rot aus dem genannten Grund, darunter beide Formen des Endes durch ein Signal |
| `tools/gegenproben/faelle.go`, `main.go`, `main_test.go` | geprüft, ohne Befund außer V3; `Env` restlos entfernt; vier Mutationen zu Z11 rot; 36 Fälle ok |
| `src/` (drei Test-Kommentare) | geprüft, ohne Befund außer V1; nur Kommentarzeilen im Diff, Bestandscode unverändert |
| `SPEC-122`, `SPEC-102`, `SPEC-036`, Plan §1, §3, §4, §6 | Code und Vertrag gleich; Plan folgt dem Code (Kopf, Tabelle, Marke, Ausgänge); keine leere Zelle im Raster |
| Hard Rules `AGENTS.md` §3.1, §3.2, §3.13 | Docker-only (kein Interpreter auf dem Host, Mutationen und Raten im Container), kein `//nolint`, keine Zuweisung an einen Slice in `done/` |
| Repo-Baum, Stempel | unverändert; Scratch-Kopien, Tags und Volume entfernt |

## Offen für die Closure (Aufgabe des Planners, kein Befund des Codes)

* Closure-Notiz §7 mit Lerneintrag; Register: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` auf `verkörpert` (Ziel und `AGENTS.md` §3.10), Finding-Klassen des Reviews; Move nach `done/` (erst Inhalt, dann reiner `git mv`).
* Wortlaut von `AGENTS.md` §3.10: der Nutzer sieht den Diff vor dem Commit; der Entwurf steht in Plan §3, das Ziel existiert.
* Ausgänge der acht Risiken in §6 (Rennen im heutigen Code, Laufzeit, Kosten, Image, Plattform, Gegenprobe, Pakete, Nachzug) samt den vorliegenden Messwerten; Größenmessung nach dem Review in §4 (475).
* `pruefeTest` (`SPEC-036`, nicht im Diff) hat dieselbe Form `l.code != 0` ohne Test für ein Ende durch ein Signal.
* Der *Ungeprüft*-Satz in `SPEC-121` zur Drain-Seite stimmt nicht (Mutation b äquivalent, I-2).
* `SPEC-102` fehlt im Index `harness/vertraege/README.md`.
* Die Fehlermeldung `Kopie für %s` und zwei Kommentare mit „Kopie“ (V3).

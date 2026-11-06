# Verifikationsbericht: slice-mutationslauf-werkzeug - 2026-11-06

**Rolle:** Verifier (Modul 11) - **Frage:** Bauen wir es richtig? - gegen Plan, DoD und Vertrag `SPEC-210`; nicht Reviewer (Diff), nicht Validator.

**Gegenstand:** Zweig `slice-mutationslauf-werkzeug`, HEAD `b55d7b5`, Diff `9c9b406..b55d7b5` (Implementer `ee83825`, `7f4bc8c`, `61d1792`; Review `d3db5ec`; Fix-Runde `61d1792`/`5f9326b`; Architect `f8526b3`/`b55d7b5`). Bezug: `LH-QA-06`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md). Review: `docs/reviews/2026-11-06-slice-mutationslauf-werkzeug-review.md`.

**Maßstab:** `.harness/skills/reviewer.md` und `.claude/commands/implement-slice.md` in der Fassung auf `main` (`90c30e7`), `AGENTS.md` §3.10, §3.15, §3.16, §3.17, §3.18, `SPEC-128` Nr. 5 d.

**Methode.** Einmal gefahren, keine Wiederholungsschleife. Belege des Implementers nicht übernommen. Alle Läufe im Image `odata-recorder:race` mit `docker create --network none --memory=2g --memory-swap=2g --cpus=4`, Quelle per `docker cp` (kein Mount), unter `flock /tmp/odata-recorder-heavy.lock`, ein schwerer Lauf zugleich; Wegwerfkopie der Quelle im Scratchpad `ver-ml`, alle Container danach entfernt (`docker ps -a` mit Label `odata-recorder.mutationslauf=1` und `ver-ml=1`: 0). Host: 20 CPUs, Last (Minutenwert) im Verlauf 5 bis 11, ein fremder Container nebenher. Messender: der Verifier.

- **Grundlauf** (Stand `b55d7b5`, `docker cp` des Archivs): `gofmt -l` leer, `go vet ./...` ohne Befund, `go test -count=1 ./...` im Modul `tools/mutationslauf`: `ok` in 214,6 s (Last beim Start 6,7 bis 12,9). Der Plan nennt 212,9 s.
- **Probe-Mutationen** (§3.17, Mengen der Fix-Runde): 32 Mutationen an Elementen, die weder Review noch Implementer nannten, je einmal gefahren; 16 rot, 16 überlebt (eine davon äquivalent). Zwei übersetzten zuerst nicht (P23, P25) und wurden in anderer Form wiederholt; P14 wurde nach einem Fehlschluss im ersten Wortlaut (der brach auch das Signal und war darum rot) neu gefasst. Tabelle unten.
- **Raten** der nebenläufigen Zeilen: vier selbst gemessen (Z6, Z12, Signal in der Nullprobe, Race-Binary der Nullprobe).
- **Ende-zu-Ende** des Treibers gegen Docker: schlechte Liste, fehlendes Image, ungültige Umgebung, SIGTERM, Gesamtgrenze.
- **Messung des Gewinns:** `messsatz.json` und `nachbau.sh` aus dem Repo, je einmal, kalt gegen warm einmal.
- **`make gates`** im Worktree: siehe *DoD*.

**Nicht aufgelöst (ehrlich).** Z8 (Rate) nicht nachgemessen (`TestHaengerNachGnade` und `TestTreiberFrist` laufen im Grundlauf, ihre Rate nicht); Speicherspitze (`memory.peak` 1213 MiB) nicht nachgemessen; `golangci-lint` auf dem Modul nicht gefahren (kein Gate, Plan §6 *weiter offen*); der Lauf mit 200 Mutationen und ein Paket mit teurem Bau (`modernc.org/sqlite`) nicht gefahren; `tools/harness/mutationslauf.sh` nur unter `sh` (dash) gefahren, wie der Implementer. Werkzeug des Verifiers: das Werkzeug `Edit` stand mir nicht zur Verfügung; Berichtigungen an diesem Bericht per `sed` in der C-Locale (bytegenau) und ein einmaliger `perl`-Aufruf ohne `-C` auf diese Datei; die Rohbytes danach gegengelesen (`LC_ALL=C grep`: keine doppelt kodierte Folge, keine Escape-Folge im Bericht).

## Ergebnis in Kurzform

HIGH: keine. MEDIUM: 8 (V-1 bis V-7 und V-13). LOW: 5 (V-8 bis V-12). INFO: 7. **Empfehlung: Closure nein, bis (a) V-6 (DoD-Punkt 3, `harness/README.md`) erledigt, (b) die Größenentscheidung beim Planner gefallen und (c) V-8 (doppelt kodierte Zeile in `SPEC-210`) berichtigt ist; V-1 bis V-5 und V-13 sind Lücken der Fänger, nicht des Codes (jeder der betroffenen Fälle verhielt sich im Ende-zu-Ende-Lauf richtig, soweit gefahren) und dürfen mit Grund und Folge als akzeptierte Negative stehen, oder ein Fall je Zelle schließt sie.** Die Messung des Gewinns bestätigt den Vertrag (kein Laufzeitgewinn zugesagt), widerlegt aber die Zahl „rund ein Viertel langsamer“ des Plans in einem Einzellauf (V-12 / I-3).

## DoD und Entscheidungen

| Punkt | Befund |
|---|---|
| DoD 1: Liste in einem Container-Lauf, Wegwerfkopie je Mutation, Ergebnis gleich Einzellauf | erfüllt: `TestGleichwertigWieEinzellauf` (acht Arten mal zwei Klassen gegen `go test -count=1 [-race]`), `TestKopieJeMutation`, `TestBaumUnveraendert` im Grundlauf grün; Ende-zu-Ende auf Satz A: vier `rot`, ein `gruen`, 50/50 und 50/50, wie der Nachbau |
| DoD 2: Ansatzpunkte (a) bis (d) umgesetzt oder je Punkt mit Messung als nicht tragend benannt | erfüllt: Plan §3 *Messung des Gewinns* benennt je Punkt; eigene Messung bestätigt, siehe *Messung des Gewinns*. Die Liste des Satzes liegt im Repo (`tools/harness/messung-mutationslauf/`), das schließt Review M-8 (wiederholbar: ja, ich habe sie ohne Scratchpad des Implementers gefahren) |
| DoD 3: Vertrag unter `harness/vertraege/` samt Tabelle und Raster; `harness/README.md` nennt das Werkzeug und seine Grenzen | **teilweise**: Vertrag, Tabelle, Raster liegen; `harness/README.md` ist im Diff unberührt (kein Eintrag in *Nicht behauptet*). Der Plan stellt die Zeile zurück bis `slice-test-determinism` in `done/` liegt; das ist auf `main` jetzt der Fall, der Zweig hat `main` nicht gemergt. V-6 |
| DoD 4: `make gates` grün; `make gegenproben` nur bei geänderter Eingabe | `make gates`: siehe Zeile unten. Eingabe der Gegenproben unberührt: `tools/harness/gegenproben-eingabe.sh` führt `src`, `go.mod`, `go.sum`, `.a-check.yml`, `.d-check.yml`, `.golangci.yml`, `lint.sh`, `migration-check.sh`, `groessenmarke.sh`, `zeichen.sh`, `tools/gegenproben`, `tools/pruefung`, `tools/groessenmarke`, `tools/zeichen`, `Dockerfile`, `.dockerignore`; keine davon im Diff, `make gegenproben` nicht nötig |
| DoD 5 bis 9 (Review, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen) | Review liegt vor; Closure-Notiz, Register, Ausgänge und Paarungen offen, wie vor der Closure zu erwarten. §6 trägt noch vier Zeilen *offen*/*weiter offen* mit Adresse (Formelkonflikt, Rollen rufen das Werkzeug nicht, Gate für das Modul samt Ausnahme in `.golangci.yml`, Größenüberschreitung) |
| Gate-Kette unberührt | bestätigt: `git diff --name-only 9c9b406..b55d7b5` nennt nur `tools/mutationslauf/*`, `tools/harness/mutationslauf.sh`, `tools/harness/messung-mutationslauf/*`, `SPEC-210`, Plan, Review. Kein Ziel in `harness/mk/`, keine Stufe im `Dockerfile`, kein `GATE_CHECKS`, keine Sensors-Tabelle, `.golangci.yml` und `harness/README.md` unberührt |
| `AGENTS.md` §3.1 | bestätigt am Treiber: `docker create --network none --memory=<M> --memory-swap=<M> --cpus=4 --pids-limit=1024 --label …`, Eingaben per `docker cp`, kein `-v`, `--volume`, `--mount`, `--tmpfs`; `TestTreiberOptionen` fängt die Mounts, `TestTreiberContainerOptionen` pinnt den ganzen Befehl; der Treiber schreibt nichts auf den Host (`TestTreiberSchreibtNichts`) |
| Entscheidungen 1 bis 8 des Architect (Plan §3) | im Code wiederzufinden: Stufe 2 (`stufe2`), Binary je Mutation und Art, Nullprobe, Cache im Container mit Grenze 1,5 GiB und `MUTATIONSLAUF_KALT`, Rate 50/50 und höchstens vier zugleich, Kopie des Wurzelmoduls (`go.mod`, `go.sum`, `src`, `tests`), Last (`--memory`, `GOMEMLIMIT`, `-p 2`, Fristen als Konstanten mit `TestFristen`), keine Änderung der Gate-Kette |
| Das Werkzeug als „von den Rollen gerufen“ oder „Gate“ dargestellt? | nein, in keinem Text: Kopf und *Stand* von `SPEC-210` („kein Gate“, „keine Rollen-Anweisung ruft es“), Kopf des Treibers, Plan §6. `README.md`, `harness/README.md`, `docs/user/` nennen es nicht (siehe V-6 für die geschuldete Zeile) |
| Rohbytes (`LC_ALL=C grep`, `AGENTS.md` §3.18) | 25 geänderte Dateien: kein BOM, kein NBSP, keine Nullbreiten- oder weichen Trennzeichen, keine Folge Backslash-`u`-vier-Hex in Quelle, Test und Doku. **Eine Zeile doppelt kodiert:** `SPEC-210` Zeile 97 (V-8) |
| Roadmap, Welle, Register | nicht angefasst |

`make gates` im Worktree auf `99ad900` (Bericht in der Fassung dieses Commits), Exit 0, 3 min 36 s: `d-check: 643 Datei(en) geprüft, 0 Befund(e)`, `groessenmarke slice-mutationslauf-werkzeug.md: ok, Messung 3463 von Marke 3000`, `zeichen: ok, 1312 Datei(en) geprüft`, `baseline-verify` und `migration-check` ok. Die Kette fährt die Tests des Moduls `tools/mutationslauf` nicht (`SPEC-210` Nr. 12 d); sie liefen separat (Grundlauf oben). Nach dem Eintrag dieser Zeile liefen `make docs-check`, `make zeichen-check`, `make groessenmarke` erneut: siehe letzte Zeile des Berichts.

## Größenüberschreitung (bewertet, nicht entschieden)

- **Messung stimmt:** `git diff --numstat 9c9b406..b55d7b5` über `tools/mutationslauf` und `tools/harness/mutationslauf.sh` gibt 3463 Einfügungen und 0 Löschungen, davon 2211 in `*_test.go` (Handoff 3129, nach `7f4bc8c`; 3101 nach `ee83825`). Schätzung 2155, Marke 3000.
- **`SPEC-128` Nr. 5 d:** „Die Messung über der Marke ist eine Überschreitung … gedeckt, wenn ein Eintrag `ueberschreitung` mit N >= Messung steht … N ist die Obergrenze, bis zu der der Eintrag gilt; wächst die Messung darüber, braucht der Stand einen neuen Eintrag mit eigenem Grund.“ Zwei Einträge (3200, 3500) mit je eigenem Grund entsprechen dem Wortlaut: der zweite deckt 3463, der erste ist durch ihn überholt, aber nicht falsch. `marke` bleibt 3000 und gleich der Formel (Nr. 5 b); `SPEC-128` Nr. 1 nennt die Datenlage („die Folge ist ein Eintrag mit Grund … keine Rückführung“) ausdrücklich.
- **Verschiebt der Eintrag die Marke faktisch?** Formal nein (Nr. 5 b prüft die Marke gegen die Formel, nicht das N). Materiell trägt jeder Eintrag die Messung aufgerundet auf Hundert; ein dritter Eintrag nach der Closure-Runde wäre derselbe Vorgang. Die Messung liegt 15 % über der Marke und 61 % über der Schätzung; der Handoff lag bei +45 % über der Schätzung, die Fix-Runde nach dem Review fügte 334 Zeilen hinzu (+11 %). Der Handoff lag 74 bis 109 % über der eigenen Erwartung des Architect (1500 bis 1800 Zeilen); die Formel von `SPEC-128` unterschätzte diesen Slice, wie ihre Datenlage es als möglich nennt.
- **Widerspruch im Plan, nicht im Vertrag:** Plan §4 *Rückführungen* nennt „die Messung nach dem Code liegt über der Marke“ als Anlass für `in-progress` nach `next`; `SPEC-128` Nr. 5 d und `AGENTS.md` §3.14 (Schnitt vor dem Code oder am vorab benannten Teil, sonst Messung mit Grund) lassen den Eintrag zu. Beides steht im Plan; §6 hält die Frage mit Ausgang *offen*, Adresse Planner. Das Gate kann es nicht entscheiden. Ein vorab benannter Teil (Cache, Ansatzpunkt (c)) besteht nicht mehr (Plan §4: klein, kein Schnitt). Der Schnitt, den der Architect nicht sieht (Programm gegen Treiber), teilt Kopie und Lauf.

## Tabelle gegen Vertrag

Die Tabelle Z1 bis Z13 deckt Nr. 1 bis 12 von `SPEC-210`, die Zeilen ohne Tabellenzeile (Nr. 1 a, 3 a, 4 a) trägt der Absatz *Fänger der Fix-Runde* mit Test je Nr. (Review L-2 erledigt). Nr. 10 d (Werkzeug erzeugt und entfernt kein Image) hat Fänger in `TestEinContainerJeAufruf` (`build`, `rmi`, `image rm`, `volume` werden nicht gerufen). Die Zusatztabelle von Nr. 11 trägt `TestListeZusatztabelle`; die Zeile `alt` leer und `neu` leer (Review M-7) ist im Vertrag entschieden und im Ende-zu-Ende-Lauf bestätigt (`Mutation 2 (a): alt leer`, kein `neu ändert nichts`; Reihenfolge Name, Datei, alt, test, klasse wie Nr. 2 d). Ausnahmen:

- Nr. 5 c, der Anker `^` des Musters: Fänger fehlt (V-1).
- Nr. 7 b, Abbruch im Arm ohne Detektor: Fänger fehlt (V-2).
- Nr. 1 c, Diagnose, zwei Zellen (V-3); Nr. 7 c, Ergebnis `rot` bei entfallenem Arm (V-13).
- Nr. 10 a, Gesamtgrenze in der Nullprobe, und Signal vor dem Start des Containers (V-4, V-5).
- Nr. 3 b, Fehler einzelner `docker cp` (V-9), Nr. 2 b Dateiart von `datei` und `paket` (V-10).

## Probe-Mutationen (Klasse, nicht Befund)

Elemente, die weder der Review-Bericht noch der Absatz *Mengen der Fix-Runde* nennen. „rot“ = von einem Test gefangen (der erste genannte Test; Grund erkennbar), „überlebt“ = alle Tests des Moduls außer den dreien mit Wartezeit (`TestHaengerNachGnade`, `TestAbbruchRaeumtAuf`, `TestTreiberFrist`) und die sechs mit echten Prozessen grün. P-Nummern sind meine; Grundlauf mit demselben Aufruf vorher grün.

| Nr. | Menge / Achse (Plan, Vertrag) | Mutation | Ergebnis |
|---|---|---|---|
| P01 | Umfang der Kopie, Elemente des Wurzelmoduls | `src` nicht kopiert (Review nannte `tests`, `go.sum`; Implementer `go.mod`) | rot (`TestKopieUmfang`) |
| P02 | Umfang der Kopie, Nicht-Elemente | `harness`, `spec` mitkopiert | **überlebt** (Fixture führt nur `docs`, `tools`, ein verschachteltes Modul) - V-11 |
| P03 | Umfang, Modus | Berechtigung der Kopie `&^ 0o007` | rot (`TestKopieUmfang`) |
| P04 | Umfang, anderes Modul, Testdaten | Verzeichnis `testdata` übersprungen | rot (`TestKopieUmfang`) |
| P05 | Cache-Adapter, tiefes Verzeichnis | Unterverzeichnis `tief` übersprungen | rot (`TestCacheAdapter`) |
| P06 | Cache-Adapter, Dateiart | `!IsDir` statt `IsRegular` (Verweis zählt) | rot (`TestCacheAdapter`) |
| P09 | Optionen der Test-Prozesse | `-test.count=1` entfällt | überlebt, **äquivalent** (Vorgabe von `-test.count` ist 1) - I-1 |
| P10 | Anker des Musters (Nr. 5 c) | `^` vor dem Testnamen im `-test.run` fehlt | **überlebt** - V-1 |
| P11 | Anker des Musters (Nr. 5 a) | `^` im `-test.list` fehlt | **überlebt** - V-1 |
| P12 | Diagnose, Grenze 40 | 41 Zeilen statt 40 | rot (`TestDiagnoseAlle`, `TestDiagnose`) |
| P13 | Diagnose, Ergebnis mal Klasse | Text des Baus mit Detektor entfällt (`uebersetzungsfehler`, Klasse `nebenläufig`) | **überlebt** - V-3 |
| P14 | Phase des Abbruchs | `ctx.Err() != context.Canceled` statt `== nil` (Gesamtgrenze in der Nullprobe wird Exit 2) | **überlebt** (zweiter Lauf; der erste Wortlaut `!= DeadlineExceeded` brach das Signal und war rot - verworfen) - V-4 |
| P15 | Abbruch mitten in den Prozessen | Prüfung `ctx.Err()` vor jedem Start im Arm entfällt | überlebt; nur die Zahl `r/N` im Abbruch ändert sich - I-2 |
| P16 | Arm mal Abbruch | `ohne.abbruch` aus der Bedingung | **überlebt** - V-2 |
| P17 | Stufe 2 der Rate | läuft mit dem Binary mit Detektor | rot (`TestRateNurFuerNebenlaeufig/7_und_3`) |
| P20 | `k` der Ergebniszeile | `runtime.NumCPU()` statt `GOMAXPROCS(0)` | überlebt (Einheitentest läuft ohne Grenze des Containers); im Ende-zu-Ende-Lauf `last=…/4` bei `--cpus=4` - I-4 |
| P22 | Dateiart von `datei` | Verweis als `datei` zulässig (`fi.IsDir()` statt `!IsRegular()`) | **überlebt** - V-10 |
| P23 | Dateiart von `paket` | eine Datei als `paket` zulässig | **überlebt** - V-10 |
| P25 | Image-Abgleich, eine Seite fehlt | `go.sum` nur auf einer Seite: kein Abweichen | rot (`TestImageAbgleich`) |
| P26 | Umgebung, Schreibweise der Zahl | `strconv.Itoa(n) != v` entfällt (`007`, `+5`) | rot (`TestUmgebung`) |
| P37 | Argumente des Programms | höchstens 5 statt 4 | rot (`TestRunEingabe`) |
| D01 | Reste, Status | `status=dead` aus `ps` | rot (`TestTreiberRaeumtAuf`) |
| D02 | Baum ohne `.git` | `--exclude=./.git` entfällt | rot (`TestTreiberAufrufUndImage`) |
| D03 | `GOMEMLIMIT` | `* 767` statt `* 768` | rot (`TestTreiberUmgebung`) |
| D04 | `MUTATIONSLAUF_GRENZE` | `0*` (führende Null) entfällt | rot (`TestTreiberUmgebung`) |
| D05 | Signal vor dem Start | `trap 'exit 2' HUP INT TERM` ohne HUP | **überlebt**; mit dash bestätigt: SIGHUP ohne Trap beendet die Shell, die EXIT-Trap läuft nicht - V-5 |
| D07 | Aufrufstelle `docker cp` (Liste) | `|| exit 2` entfällt | **überlebt** - V-9 |
| D08 | Aufrufstelle `docker cp` (Programm) | `|| exit 2` entfällt | **überlebt** - V-9 |
| P40 | Aufräumen je Mutation | `.test.race` der Mutation nicht entfernt | rot (`TestBinariesWerdenEntfernt`) |
| P41 | Aufräumen der Nullprobe | `.test` der Nullprobe nicht entfernt | rot (`TestBinariesWerdenEntfernt`) |
| P42 | Ergebnis `rot` der Rate | `rot` nur, wenn der Arm mit Detektor rot ist | **überlebt** (`rot` bei entfallenem Arm wird `flatternd`) - V-13 |
| P50 | Zeichenmenge von `name` (M-6) | `~` zusätzlich zugelassen | **überlebt** (`TestListeZeichen` führt elf Nicht-Zeichen als Auswahl) - V-11 |

Zu den acht MEDIUM des Reviews (Menge, Achsen, je Achse eine ungenannte Mutation; „gefangen“ nur wenn die Klasse fällt):

| Review | Menge (Plan) | Ungenannte Mutation je Achse | Ergebnis |
|---|---|---|---|
| M-1 Umfang der Kopie | 4 Elemente des Wurzelmoduls, 3 Nicht-Elemente, 3 Modi, 2 Dateiarten, 5 des anderen Moduls | `src` (P01), Modus (P03), Testdaten (P04); Nicht-Elemente `harness`, `spec` (P02) | Achsen gefangen bis auf die Nicht-Elemente außerhalb der Fixture (V-11) |
| M-2 echter Cache-Adapter | Größe: 5 Formen; Leeren | tiefes Verzeichnis (P05), Dateiart (P06) | gefangen; `go clean -cache` mit festem Argumentvektor (`TestCacheAdapter`) |
| M-3 Container-/Bau-Optionen | 9 + 4 + 5 + 4 Elemente | Befehlszeile des Treibers vollständig gepinnt (`TestTreiberContainerOptionen`); `GOMEMLIMIT`-Faktor (D03), Zeichenform der Grenze (D04, P26); Option `-test.count` äquivalent (P09); Anker `^` des Musters (P10, P11) | Container und Bau gefangen; Optionen der Test-Prozesse lückenhaft (V-1) |
| M-4 Diagnose | 4 Ergebnisse, 2 Klassen, 2 Arme | Zeilengrenze 41 (P12) rot; Zelle `uebersetzungsfehler` mal `nebenläufig` (P13) | Menge nicht ganz iteriert (V-3) |
| M-5 Signal in der Nullprobe | 3 Phasen mal 3 Folgen | Gesamtgrenze statt Signal in der Nullprobe (P14) | andere Ursache in derselben Phase überlebt (V-4) |
| M-6 Zeichenmengen | 65 + 63 Zeichen, 11 außerhalb, 7 Präfixe | Zeichen außerhalb der Menge: `~` (P50); die Ränder sind Schleifen über alle 65 und 63 Zeichen | Menge selbst gefangen; die Nicht-Zeichen sind eine Auswahl ohne benannte Prüfgrenze (V-11) |
| M-7 leeres Paar | `alt` leer, `neu` leer | Ende-zu-Ende: ein Befund `alt leer` | entschieden im Vertrag, im Lauf bestätigt |
| M-8 Messung wiederholbar | Satz A/B, Nachbau, Skript | Messung ohne Dateien des Implementers gefahren | wiederholbar |

### Raten der nebenläufigen Zeilen (Verifier)

Messender: der Verifier, Test-Binary aus `go test -c` im Image `odata-recorder:race` (`--memory=2g --cpus=4`), Last des Hosts 5 bis 7; je Mutation 20 Prozesse mit `-test.count=1` und ein Prozess mit `-test.count=20` (Wirkungs-Fänger, kein Datenrennen, kein Detektor). Die Mutationen sind meine, bis auf die vierte (Race-Binary der Nullprobe), die auch der Implementer nannte:

| Zeile | Mutation | Test | mit Mutation | ohne Mutation |
|---|---|---|---|---|
| Z6 Deckel | Deckel `gleichzeitig+2` (6) | `TestRateNurFuerNebenlaeufig` | **20 von 20** Prozessen rot, `-test.count=20`-Prozess rot | 0 von 20, 0 |
| Z12 Gleichzeitigkeit der Baue | Nullprobe mit Detektor läuft in einer Goroutine neben der ohne | `TestSpeicherUndGleichzeitigkeit` | **20 von 20**, rot | 0 von 20, 0 |
| Z8 / M-5 Signal in der Nullprobe | `ctx.Err() != nil` aus der Schleife der Mutationen: nach dem Signal startet die erste Mutation | `TestSignalInNullprobe` | **20 von 20**, rot | 0 von 20, 0 |
| Z8 Race-Binary der Nullprobe | `.test.race` aus der Liste der zu entfernenden Pfade | `TestBinariesWerdenEntfernt` | **20 von 20**, rot | 0 von 20, 0 |

Z8 (Hänger, Prozessgruppe, Gesamtgrenze) nicht nachgemessen. Die Wirkungs-Fänger sind damit in jedem Lauf rot, wie `AGENTS.md` §3.10 für Nebenläufigkeit verlangt.

## Messung des Gewinns (selbst gefahren)

Aufruf wie in Plan §5: `sh tools/harness/mutationslauf.sh tools/harness/messung-mutationslauf/messsatz.json` und `sh tools/harness/messung-mutationslauf/nachbau.sh`; ein Lauf je Seite, unter `flock`, Host 20 CPUs, Last 5,2 bis 7,9, ein fremder Container.

| Seite | Wanduhr (ganzer Aufruf) | im Container | Prozesse | Ergebnis je Mutation |
|---|---|---|---|---|
| Werkzeug, warm | 19,7 s | `zusammenfassung dauer` 13,8 s, Summe der Mutationen 4,2 s | 105 | m1 `rot` 50/50 und 50/50, m2 bis m4 `rot`, m5 `gruen` |
| bisheriges Vorgehen (Nachbau) | 21,5 s | 17,4 s | 105 | gleich |
| Werkzeug, `MUTATIONSLAUF_KALT=1` | 31,6 s | 26,3 s | 105 | gleich außer `dauer` und `last` |

- **Ergebnis je Mutation gleich** in beiden Vorgehen und kalt gleich warm (Z7, Nr. 8 b bestätigt am Satz A).
- **Cache (c):** kalt kostet rund 12 s mehr als warm, passend zu den „rund 10 s“ des Plans.
- **Laufzeit:** in diesem Einzellauf liegt das Werkzeug 8 % unter dem Nachbau (Wanduhr), nicht 25 % darüber, wie Plan §6 und Review (30,2 bis 30,5 s gegen 23,7 bis 24,4 s) es für Satz A berichten. Die Streuung der Läufe des Implementers (20,8 bis 35,1 s) ist größer als der Unterschied; ein Lauf je Seite beweist weder das eine noch das andere. Der Vertrag sagt keinen Gewinn zu und behauptet „nicht schneller“ nur als Messung (Kopf); ich lese den Satz als vorsichtig, nicht als falsch. I-3.
- **Rest der Texte:** Kopf und *Stand* von `SPEC-210` ohne Laufzeitzusage; Plan §6 *Kopf von SPEC-210* nennt den Ausgang. Zwei Stellen sagen noch „billiger“: Plan §1 *Ausdrücklich NICHT*, Punkt „Eine Lockerung von `AGENTS.md` §3.10“ („das Werkzeug macht sie billiger, nicht kleiner“), und der Titel des Slice. V-12.

## Ende-zu-Ende des Treibers

| Fall | Ergebnis |
|---|---|
| Liste mit Befunden (`alt trifft nicht`, doppelter Name, fehlende Datei, `alt` leer, ungültiger Test, unbekannte Klasse) | Exit 2, sechs Zeilen `mutationslauf: Eingabe: Mutation <n> (<name>): …` in der Reihenfolge von Nr. 2 d; kein Container |
| Image fehlt | `Image fehlt: gibt-es-nicht:x (make test-race bzw. make test)`, Exit 2 |
| `MUTATIONSLAUF_SPEICHER=5g` | `Umgebung ungültig: MUTATIONSLAUF_SPEICHER`, Exit 2 |
| SIGTERM an den Treiber nach 9 s (Nullprobe von Satz A) | Exit 3, fünf Zeilen `abbruch` und `zusammenfassung`, Container danach 0 |
| `MUTATIONSLAUF_GRENZE=12` | Exit 3 nach 12,0 s, fünf Zeilen `abbruch`, Container danach 0 |

Die Gesamtgrenze in der Nullprobe verhält sich im Lauf richtig (Exit 3); nur der Fänger fehlt (V-4).

---

## Befunde

### V-1 (MEDIUM): Der Anker `^` des Musters für `-test.run` und `-test.list` ist nicht gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-210` Nr. 5 a, 5 c (`'^<Test>$'`)
- `pfad`: `tools/mutationslauf/lauf.go` · `[]string{"-test.run", "^" + test + "$"}` und `tools/mutationslauf/nullprobe.go` · `start("-test.list", "^"+m.Test+"$")`
- `befund`: Z4 und die Mutationen des Implementers führen den Anker `$` (Präfix `TestA` gegen `TestAB`); der Anker `^` fehlt in beiden Aufrufen, ohne dass ein Test rot wird (P10, P11, alle Tests des Moduls einschließlich der sechs mit echten Prozessen grün). Ein Muster `TestAdd$` trifft auch `TestMyTestAdd`; der Lauf zählte dessen Rot als Rot des zuständigen Tests.
- `verifizierbar`: ja (Mutationen P10, P11)
- `klasse`: Menge der Anker, ein Element gefangen

### V-2 (MEDIUM): Abbruch eines Prozesses im Arm ohne Detektor ist nicht gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.16 (Abbruch mal Operation); `SPEC-210` Nr. 6 b, 7 b, 9 c
- `pfad`: `tools/mutationslauf/lauf.go` · `case mit.abbruch || ohne.abbruch || ctx.Err() != nil:`
- `befund`: Die Menge „Abbruch im Arm“ hat zwei Elemente (Arm mit, Arm ohne Detektor). `ohne.abbruch` aus der Bedingung zu streichen (P16) lässt alle Tests grün. Folge im Betrieb: beendet das Werkzeug einen hängenden Prozess im Arm ohne Detektor (nach `-test.timeout` und 30 Sekunden), zählt der Prozess nicht als rot, und die Mutation endet als `gruen` oder `flatternd` statt als `abbruch`; ein hängender Test würde als gefangen oder als nicht gefangen gelesen statt als nicht ausgeführt.
- `verifizierbar`: ja (Mutation P16)
- `klasse`: Menge der Arme, ein Element gefangen

### V-3 (MEDIUM): Die Diagnose führt zwei Zellen der Menge Ergebnis mal Klasse nicht

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-210` Nr. 1 c
- `pfad`: `tools/mutationslauf/lauf.go` · `return bauErgebnis(a), a.Text` (zweiter Bau, mit Detektor) und `TestDiagnoseAlle` in `tools/mutationslauf/huelle_test.go`
- `befund`: Der Plan nennt als Menge 4 Ergebnisse mal 2 Klassen mal 2 Arme; `TestDiagnoseAlle` führt `uebersetzungsfehler`/`andere`, `rot-paket`/`andere`, `flatternd`/`nebenläufig` und `abbruch` in beiden Klassen, also fünf der acht Zellen, als Liste und nicht als Schleife. Ohne den Text des Baus mit Detektor (P13, `uebersetzungsfehler`, Klasse `nebenläufig`) bleibt alles grün; der Compilerfehler der Detektor-Art würde nie ausgegeben.
- `verifizierbar`: ja (Mutation P13)
- `klasse`: Klasse generierbar benannt, im Handoff als Liste

### V-4 (MEDIUM): Die Gesamtgrenze in der Nullprobe hat keinen Fänger

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.16 (Abbruch mal Operation); `SPEC-210` Nr. 10 a, 10 b
- `pfad`: `tools/mutationslauf/lauf.go` · `len(befunde) > 0 && ctx.Err() == nil`
- `befund`: Nr. 10 b: die Gesamtgrenze „handelt … wie bei einem Signal“. `TestSignalInNullprobe` bricht den Kontext mit `cancel()` ab (Ursache Signal). Die Bedingung auf `ctx.Err() != context.Canceled` zu ändern (Gesamtgrenze in der Nullprobe wird Exit 2 und `Ausgangsstand rot`) lässt alle Tests grün (P14). Der Code ist heute richtig (Ende-zu-Ende `MUTATIONSLAUF_GRENZE=12`: Exit 3), nur der Fänger fehlt: die Menge „Ursache (Signal, Gesamtgrenze) mal Phase (Nullprobe, Bau, Stufe 1, Arme, Stufe 2)“ ist in der Nullprobe auf eine Ursache geschnitten.
- `verifizierbar`: ja (Mutation P14)
- `klasse`: Randform-Zelle ohne Fänger (Ursache des Abbruchs)

### V-5 (MEDIUM): Ein Signal an den Treiber vor `docker start` ist nicht gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.16 (Abbruch); `SPEC-210` Nr. 10 a („in jedem Fall … entfernt er den Container“), Nr. 10 c
- `pfad`: `tools/harness/mutationslauf.sh` · `trap 'exit 2' HUP INT TERM`
- `befund`: `TestTreiberRaeumtAuf` sendet HUP, INT und TERM erst, wenn `docker start` im Ersatz-Skript erschien. Der Abschnitt zwischen `docker create` und `start` (drei Aufrufe von `docker cp`, bei 19 MB einige Sekunden) hängt an der Zeile `trap 'exit 2' HUP INT TERM`; ohne `HUP` in ihr (P-D05) bleibt alles grün, und die EXIT-Trap läuft dann nicht (mit dash gezeigt: eine Shell, die an SIGHUP ohne Trap endet, ruft ihre EXIT-Trap nicht). Folge: der Container mit Label bleibt bis zum nächsten Aufruf, der ihn räumt (Nr. 10 c).
- `verifizierbar`: ja (Mutation D05)
- `klasse`: Randform-Zelle ohne Fänger (Abbruch vor dem Start)

### V-6 (MEDIUM): DoD-Punkt 3 ist nicht erfüllt - `harness/README.md` nennt das Werkzeug nicht

- `kategorie`: MEDIUM
- `quelle`: Plan §2 (DoD 3); `AGENTS.md` §3.9; Plan §4 (Bedingung `slice-test-determinism`)
- `pfad`: `harness/README.md` · *Nicht behauptet*
- `befund`: Der Plan schreibt die Zeile erst nach `slice-test-determinism` in `done/`. Das Slice liegt jetzt auf `main` in `done/`, der Zweig `slice-mutationslauf-werkzeug` ist nicht mit `main` gemergt; die Datei ist im Diff `9c9b406..b55d7b5` unberührt. Die Grenzen des Werkzeugs (Nr. 5 e, 6 d, 12) stehen nur in `SPEC-210`; wer `harness/README.md` liest, erfährt nicht, dass es das Werkzeug gibt und kein Gate ist. Ob die Zeile der Implementer nachträgt oder der Planner beim Merge ordnet, steht im Plan (§4); beides ist offen.
- `verifizierbar`: nein (kein Gate)
- `klasse`: DoD-Punkt offen, Voraussetzung erfüllt

### V-7 (MEDIUM): Die Größenentscheidung steht offen beim Planner, und der Plan widerspricht sich

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.14; `SPEC-128` Nr. 5 d; Plan §4, §6
- `pfad`: Plan · `ueberschreitung: 3500 |` und `Rückführungen`
- `befund`: Bewertung siehe Abschnitt *Größenüberschreitung*. Beide Einträge entsprechen Nr. 5 d, die Marke ist nicht verschoben; der Plan nennt in §4 dieselbe Überschreitung als Rückführungsanlass und lässt in §6 den Ausgang offen. Kein Gate sieht den Widerspruch; die Closure setzt die Entscheidung voraus.
- `verifizierbar`: ja (`make groessenmarke`, siehe unten)
- `klasse`: Marke überschritten, Entscheidung offen

### V-13 (MEDIUM): `rot` bei entfallenem Arm ist nicht gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.16 (Zusatztabelle, je Modus eine Zeile mit Fänger); `SPEC-210` Nr. 7 b, 7 c, Zusatztabelle *Klasse `nebenläufig`*
- `pfad`: `tools/mutationslauf/lauf.go` · `case rot == gesamt:`
- `befund`: Nr. 7 c: `rot` heißt, alle Prozesse aller Arme mit Prozessen sind rot; ein Arm mit der Zahl 0 entfällt. Z6 fängt für den entfallenen Arm nur, dass kein Prozess startet (`Arme 0/3: race 1`), nicht das Ergebnis. `rot` nur zu melden, wenn der Arm mit Detektor Prozesse rot hatte (`rot == gesamt && mit.rot > 0`, P42), lässt alle Tests grün; eine Mutation mit `n_detektor` gleich 0 und lauter roten Prozessen ohne Detektor würde dann `flatternd` gemeldet. Der Modus „ein Arm“ ist ein Modus der Zusatztabelle ohne Fänger für das Ergebnis.
- `verifizierbar`: ja (Mutation P42)
- `klasse`: Modus der Zusatztabelle, Ergebnis ohne Fänger

### V-8 (LOW): Eine Zeile von `SPEC-210` ist doppelt als UTF-8 kodiert

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.18; `SPEC-210` Nr. 11 (Zusatztabelle)
- `pfad`: `harness/vertraege/SPEC-210-mutationslauf.md` · `| `alt` leer **und** `neu` leer |`
- `befund`: Die Zeile (Zeile 97, seit `f8526b3`) trägt die Bytefolge `C3 83 C2 A4` statt `C3 A4` im Befundtext `neu ändert nichts`, `entfällt` und `über` (zweimal kodiert, sichtbar als großes A mit Tilde und Währungszeichen vor dem Rest). `LC_ALL=C grep -c` auf `C3 83` über alle 25 geänderten Dateien: genau diese eine Zeile; vor `f8526b3` kein Treffer. `make zeichen-check` meldet sie nicht (gültiges UTF-8, außerhalb der 253 Zeichen); das Gate nennt diese Klasse nicht (`harness/README.md` *Nicht behauptet*: Zeichen außerhalb der Menge). Der Vertrag nennt den Wortlaut des Befunds `neu ändert nichts` in Nr. 2 f richtig; die Zusatztabelle weicht davon ab.
- `verifizierbar`: ja (`LC_ALL=C grep -c $'\xc3\x83'`)
- `klasse`: doppelt kodiertes UTF-8 durch Schreibwerkzeug

### V-9 (LOW): Zwei der drei `docker cp` des Treibers haben keinen Fänger für sich

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15 (Aufrufstellen sind eine Menge)
- `pfad`: `tools/harness/mutationslauf.sh` · `"$docker" cp "$liste" "$c:/tmp/liste.json" || exit 2` und `-cf - mutationslauf | "$docker" cp - "$c:/tmp" || exit 2`
- `befund`: Der Fall `cp scheitert` setzt `STUB_CP=nein` für alle Aufrufe; das erste `cp` endet dann den Treiber, die anderen beiden bleiben ungeprüft. Ohne ihr `|| exit 2` (D07, D08) bleiben alle Tests grün. Folge: das Programm im Container findet die Liste oder sein Quellverzeichnis nicht und endet selbst mit Exit 2; die Meldung ist eine andere.
- `verifizierbar`: ja (Mutationen D07, D08)
- `klasse`: Fänger führt eine von drei Aufrufstellen

### V-10 (LOW): Dateiart von `datei` und `paket` nur zum Teil gefangen

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15; `SPEC-210` Nr. 2 b (`datei` reguläre Datei, `paket` Verzeichnis)
- `pfad`: `tools/mutationslauf/liste.go` · `if err != nil || !fi.Mode().IsRegular() {` und `err != nil || !fi.IsDir()`
- `befund`: Ein Verweis als `datei` (P22) und eine Datei als `paket` (P23) sind zulässig, ohne dass ein Test rot wird. Der Code prüft heute beides richtig (`Lstat`, `IsRegular`, `IsDir`). Folge einer Regression: ein Verweis als `datei` schriebe in der Kopie durch den Verweis (Nr. 4 c, der Baum des Aufrufers ändert sich nie, gilt im Container für die Kopie des Baums).
- `verifizierbar`: ja (Mutationen P22, P23)
- `klasse`: Dateiart, ein Teil gefangen

### V-11 (LOW): Die Nicht-Elemente zweier Mengen sind eine Auswahl, die Prüfgrenze fehlt

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.17 (Menge, die die Zusage meint)
- `pfad`: `tools/mutationslauf/kopie.go` · `[]string{"src", "tests"}`
- `befund`: Nr. 4 a zählt vier Elemente auf; alles andere gehört nicht in die Kopie. `TestKopieUmfang` führt davon `docs`, `tools` und ein verschachteltes Modul; `harness` und `spec` mitzukopieren (P02) bleibt grün. Folge: eine größere Kopie, kein anderes Ergebnis. Dasselbe gilt für die Zeichenmenge von `name` (Nr. 2 b): `TestListeZeichen` führt elf Zeichen außerhalb der 65 als Auswahl; `~` zusätzlich zuzulassen (P50) bleibt grün. Beide Mengen sind unbeschränkt (jede andere Datei, jedes andere Zeichen); weder Plan noch Vertrag nennen die Prüfgrenze mit Grund und Folge als akzeptiertes Negativ.
- `verifizierbar`: ja (Mutationen P02, P50)
- `klasse`: Menge unbeschränkt, Prüfgrenze nicht benannt

### V-12 (LOW): Zwei Texte sagen noch „billiger“

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.7; Plan §6 *Kopf von SPEC-210*
- `pfad`: Plan · `das Werkzeug macht sie billiger, nicht kleiner` und der Titel `Mutationsläufe der Rollen in einem Container-Lauf mit gemeinsamem Cache`
- `befund`: Der Architect hat den Satz im Kopf von `SPEC-210` entfernt (Ausgang *entschieden*); in Plan §1 (*Ausdrücklich NICHT*, Punkt zu `AGENTS.md` §3.10) steht er weiter, und die Roadmap (`docs/plan/planning/in-progress/roadmap.md` Zeile 19: „Architect-Aufträge offen“) beschreibt einen früheren Stand. Beide sind Texte des Planners.
- `verifizierbar`: nein
- `klasse`: Zusage im Text, die die Messung nicht trägt

## INFO

- **I-1** P09 (`-test.count=1` entfällt) ist äquivalent: ein Test-Binary läuft ohne das Flag mit `-test.count=1`. Die Zusage in Nr. 5 c bleibt im Wortlaut; kein Fänger nötig.
- **I-2** P15: ohne die Prüfung vor jedem Start im Arm starten nach einem Abbruch weitere Prozesse, die sofort enden; sichtbar nur in `r/N` der Zeile.
- **I-3** Die Zahl „rund ein Viertel langsamer“ (Plan §6) trägt mein Einzellauf nicht (siehe *Messung des Gewinns*); beide Seiten streuen mit der Last stärker als der Unterschied.
- **I-4** P20: `k` kommt aus `runtime.GOMAXPROCS(0)`; mit Go 1.27.1 und `--cpus=4` ist es 4 (Ende-zu-Ende `last=5.53/4`); der Einheitentest kann es nicht fangen (keine Grenze des Containers im Test).
- **I-5** Eine Mutation mit `abbruch` ohne Text gibt eine Zeile `<name>: ` auf stderr; Nr. 1 c sagt die Ausgabe nicht zu.
- **I-6** `TestAbbruchRaeumtAuf`, `TestHaengerNachGnade`, `TestTreiberFrist` laufen im Grundlauf grün, ihre eigenen Mutationen habe ich nicht erneut gefahren (Review und Implementer haben sie; Z8 Rate nicht nachgemessen).
- **I-7** Der Zweig hat `main` nicht gemergt; `harness/README.md` und die Sensors-Tabelle auf `main` sind seit `slice-test-determinism` weitergegangen. Der Merge liegt beim Planner.

## Geprüft, ohne Befund

- `tools/mutationslauf/` (Programm, 7 Dateien): Entscheidungen 1 bis 8 und `SPEC-210` Nr. 1 b, 1 d, 2 a bis f, 3 c, 4 b, 5 a, 5 b, 6, 7, 8, 9, 10 e im Code wiedergefunden; Grundlauf grün; 16 Probe-Mutationen rot.
- `tools/harness/mutationslauf.sh` (Treiber): `AGENTS.md` §3.1 erfüllt; Ende-zu-Ende fünf Fälle wie zugesagt.
- `tools/harness/messung-mutationslauf/`: Liste und Nachbau im Repo, wiederholbar (Aufruf in Plan §5); Ergebnis je Mutation stimmt.
- `harness/vertraege/SPEC-210-mutationslauf.md` gegen den Code: Wortlaut der Befunde (Nr. 2 f) stimmt mit `liste.go` und dem Ende-zu-Ende-Lauf überein.
- Gate-Kette, `.golangci.yml`, `harness/README.md`, Sensors-Tabelle, `Dockerfile`, `harness/mk/`: unberührt.

**Letzter Lauf nach dem letzten Eintrag in diesem Bericht:** `make docs-check` (643 Dateien, 0 Befunde), `make zeichen-check` (1312 Dateien, ok), `make groessenmarke` (Messung 3463 von Marke 3000, ok) grün; `make gates` grün auf `99ad900`.

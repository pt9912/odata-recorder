# Review-Report: slice-gates-ohne-worktrees — 2026-11-18

**Review-Art:** Code (gegen Plan, Vertrag `SPEC-212` und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-gates-ohne-worktrees`, Zweig `slice-gates-ohne-worktrees`, `fe099d91..9849fd6b`: Start-Move, Architect (`SPEC-212`, Index), Roadmap-Nachzug, Planner (Groessenmarke), Implementer (`.a-check.yml`, `.d-check.yml`, `tools/gegenproben/` mit `arbeitsbaeume.go`, `arbeitsbaeume_test.go`, `faelle.go`, `guide.go`, `strom_test.go`, `SPEC-102`, `SPEC-039`, `SPEC-040`, `SPEC-212` Stand, `harness/README.md`, Plan §3 und §6).

**Skill:** `.harness/skills/reviewer.md` @ `9849fd6b` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext** (die Vertraege, gegen die geprueft wurde):

- Slice-Plan `slice-gates-ohne-worktrees` (DoD §2, §3 Tabelle Zusage und Faenger Nr. 1 bis 12 mit Nachweis, §4 Block, §6 Raster und Risiken)
- [`SPEC-212`](../../harness/vertraege/SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md) Nr. 1 bis 6, [`SPEC-039`](../../harness/vertraege/SPEC-039-architektur-gate.md), [`SPEC-040`](../../harness/vertraege/SPEC-040-docs-check.md), [`SPEC-102`](../../harness/vertraege/SPEC-102-gegenproben.md)
- [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
- `LH-QA-06`
- `AGENTS.md` §3.6 bis §3.18; Uebergangsregeln `docs/plan/planning/README.md` Nr. 2 und 7

**Vorgehen und Belege.** Gelesen: der gesamte Diff. Der Zweig wurde nicht veraendert (`git status --short` leer vor und nach dem Lauf; `git ls-files .wt` leer). Mutationen und Pin-Proben liefen in Kopien: `git archive HEAD` unter dem Scratchpad, ohne Mount (`docker create`, `docker cp`, `docker start`), weil der Daemon das Scratchpad nicht sieht. Schwere Laeufe (`docker build`) unter `flock` auf `.wt/heavy.lock`. Fuer die Mutationen baute ich aus dem Stand `9849fd6b` ein Image mit demselben `tools/gegenproben` und denselben Pins von `a-check` und `d-check`, mit zwei Abweichungen: `faelle()` liefert nur die letzten neun Faelle (205 bis 213), und `pruefung`, `groessenmarke`, `zeichen`, `katalog` sind Ersatz-Skripte mit Exit 0 (die neun Faelle rufen sie nicht). Die Konfigurationen ersetzte ich je Lauf per `docker cp` in `/src`. Ein voller Bau der Stufe `gegenproben` (213 Faelle, Last 237 bis 60 durch fremde Laeufe) endete mit 9 roten Faellen `lint/*` wegen "Zeit ueberschritten" (`SPEC-102` Randform 6, bekannt); die Faelle 205 bis 213 waren in diesem Lauf alle gruen. Die Stufe `lint` des `Dockerfile` war fuer diesen Kontext im Cache (Lauf gruen, 1,9 s, kein neuer Lauf).

Gegengelesen auf Rohbytes (`LC_ALL=C grep -P` auf BOM, NBSP, Nullbreiten- und Steuerzeichen, `\\u[0-9A-Fa-f]{4}`): alle 16 geaenderten oder neuen Textdateien des Diffs (`.a-check.yml`, `.d-check.yml`, `harness/README.md`, `harness/vertraege/README.md`, `SPEC-039`, `SPEC-040`, `SPEC-102`, `SPEC-212`, die fuenf Go-Dateien, Plan, Roadmap, `slice-gate-integration-determinism`): kein Treffer. Dieser Report wurde nach dem Schreiben ebenso gegengelesen.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `SPEC-212` Nr. 5 Zeile 9 sagt zu, `file` und `structure` lesen `.wt/` nicht; Fall 213 fuehrt die 13 Pfade der Menge (Regeln von `file:` und die zweite Regel von `structure`), nicht die erste Regel von `structure` (`harness/README.md`, Tabellenzellen). Eigene Mutation: `files: "harness/README.md"` zu `"**/harness/README.md"` im Block `structure` bleibt gruen, 0 von 9 rot, Last 11,7. Ein Szenario: jemand weitet den Glob der ersten Regel auf `**/`; die Kopie von `harness/README.md` in jedem Arbeitsbaum wird mitgemessen, kein Fall wird rot. | `AGENTS.md` §3.15 (ein Fall je Element einer Menge), §3.17 (Klasse der Zusage) | `tools/gegenproben/arbeitsbaeume.go` · `for _, r := range gdMenge()`; `harness/vertraege/SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md` · "`file` und `structure` lesen `.wt/` nicht" | ja — Mutation `.d-check.yml` Zeile 231, Lauf der Faelle 205 bis 213 | Fänger schmaler als Zusage (Menge der Regeln von `structure` nicht erschoepft) |
| F-2 | MEDIUM | Das akzeptierte Negativ `SPEC-212` Nr. 6 a begruendet "nie getrackt" mit der Datei `.git` des Arbeitsbaums; das trifft nur einen `git worktree`. Gemessen (Wegwerf-Repo): `git add .wt/w/x.go` in einem Arbeitsbaum legt nichts ab, `git add .wt/plain/y.go` in einem Verzeichnis ohne `.git` legt die Datei ab (`git ls-files .wt` nennt sie). Nr. 6 b sagt "Kein Arbeitsablauf legt sie an", Nr. 1 nennt aber `.wt/heavy.lock`, eine einfache Datei unter `.wt/`, und das Repo fuehrt weder `.gitignore` noch einen Eintrag fuer `.wt/`. Eine so eingecheckte Datei bleibt fuer `make a-check` und `make docs-check` unsichtbar; die Bedingung der Entscheidung (Nr. 2: "nur Pfade, die nie Teil eines Commits sind") hat keinen Fänger. Grund und Folge des Negativs tragen die gemessene Form nicht. | `AGENTS.md` §3.15 (akzeptiertes Negativ ohne tragenden Grund), §3.6 | `harness/vertraege/SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md` · "Ein Arbeitsbaum trägt eine Datei `.git`, `git add` nimmt ihn als eingebettetes Repository" | ja — `git add` in Wegwerf-Repo, `git ls-files .wt` | Negativ-Grund deckt die Form nicht |
| F-3 | MEDIUM | `.d-check.yml` fuehrt zehn Module; `SPEC-212` Nr. 3 b nennt fuenf als Leser von `scan.ignore` (`links`, `anchors`, `ids`, `codepaths`, `matrix`) und Nr. 1 vier als nicht lesend (`file`, `structure`, `targets`, `planning`); `spans` steht in keiner der beiden Aussagen. Gemessen (`.wt/**` aus `scan.ignore` entfernt, Defekte unter `.wt/w/`): `links` (2), `anchors` (2), `ids` (5) und `codepaths` (2) melden dann, mit dem Eintrag 0 Befunde; fuer `matrix` und `spans` liess sich kein Befund erzeugen, die Aussage ist ungemessen. Nr. 3 b misst am Pin nur `links`. Fuer `anchors`, `ids`, `matrix`, `codepaths` und `spans` nennt kein Fall und kein Negativ in `SPEC-212` Nr. 6 Grund und Folge; der Zeiger in der Grenze von `SPEC-102` ("`SPEC-212` Nr. 6 e, f") trifft den Docker-Kontext und die nicht gefahrenen Gates. | `AGENTS.md` §3.15 (akzeptiertes Negativ ohne Grund und Folge), §3.17 (Klasse: jedes Modul, das `.wt/` lesen kann) | `harness/vertraege/SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md` · "Das Muster gilt für die Module, die `scan.ignore` lesen"; `harness/vertraege/SPEC-102-gegenproben.md` · "haben keinen eigenen Fall" | ja — Probe mit und ohne den Eintrag (`d-check`, echte `.d-check.yml`) | Negativ ohne Grund und Folge, Modul-Klasse unvollstaendig |
| F-4 | LOW | Die Grenze von `SPEC-102` sagt, die Module von `d-check` "außer `links`, `file` und `structure`" lesen `scan.ignore` wie `links`. Das trifft `targets` und `planning` nicht (`.d-check.yml` Zeilen 142 und 248: `makefiles` folgt `scan.ignore` nicht, `file` liest es nicht; `SPEC-212` Nr. 1: Pfade an der Wurzel verankert). Der Kommentar in `arbeitsbaeume.go` nennt vier Module richtig. | `AGENTS.md` §3.11 | `harness/vertraege/SPEC-102-gegenproben.md` · "lesen `scan.ignore` wie `links` und haben keinen eigenen Fall" | nein — Prosa | Aussage nennt mehr Elemente als gemessen |
| F-5 | LOW | Die Roadmap fuehrt `slice-gates-ohne-worktrees` weiter als "angelegt in `open/`" und "die Reihenfolge ... legt der User beim Start fest", obwohl der Slice in `in-progress/` liegt; dieselbe Angabe "(Linie C, `open/`; ...)" steht in `done/slice-gate-integration-determinism.md` §6 neben dem nachgezogenen Link. Beide Zustandsangaben sind nur im Pfad des Links richtig. Gegen die HIGH-Liste: keine Erzaehlung der Entstehung, nur ein veralteter Ortsvermerk, daher LOW. | `AGENTS.md` §3.7 (Zustandsfelder), §3.9 | `docs/plan/planning/in-progress/roadmap.md` · "angelegt in `open/`, Linie C, Kettenslice; Worktrees unter `.wt/` aus den Gates des Hauptbaums ausnehmen" | nein — kein Gate liest die Prosa | Zustandsfeld nach Lifecycle-Move nicht nachgezogen |
| F-6 | INFO | Pin-Semantik am 2026-11-18 selbst gemessen, deckt sich mit `SPEC-212` Nr. 3 und 4: `a-check` `exclude` auf oberster Ebene wirkt; `.wt/**` nimmt `.wt/w/` und `.wt/a/.wt/b/` aus und laesst `src/hexagon/model/.wt/x.go` geprueft; `.wt/*`, `.wt/`, `.wt`, `/.wt/**`, `./.wt/**` nehmen nichts aus; `**/.wt/**` nimmt das Verzeichnis im Quellbaum mit aus; ein Tippfehler im Schluessel (`excludes:`) endet mit Exit 2, nicht still. `d-check`: `.wt/**` nimmt `.wt/` samt Verschachtelung aus und laesst `x/.wt/b.md` geprueft; `.wt/*` verhaelt sich wie `.wt/**`; `.wt` (ohne Suffix) ebenso, `.wt/` und `/.wt/**` nehmen nichts aus (Nr. 3 b nennt `.wt` ohne Suffix nicht). Ein Arbeitsbaum-Abbild des Repos unter `.wt/w/` und `.wt/a/.wt/b/` aendert die Laeufe von `a-check` und `d-check` im Hauptbaum nicht (Exit 0, 0 Befunde). | `SPEC-212` Nr. 3, Nr. 4 | `.a-check.yml` · `exclude: [".wt/**"]`; `.d-check.yml` · `".wt/**"` | ja — Proben des Reviewers | — |
| F-7 | INFO | `harness/gegenproben.stempel` ist gegenueber `main` unveraendert (Uebergangsregel Nr. 2); die Pruefsumme der Eingaben dieses Stands (`607af839...`) weicht vom Stempel (`88f9e74c...`) ab, der Integrationsschritt faehrt `make gegenproben` neu. Zeitbudget: ein Lauf der neun Faelle 205 bis 213 im Container (`--cpus=4`, Last 7,2) dauerte 0,84 s einschliesslich Start des Containers; die Faelle brauchen kein Produkt, nur `a-check`, `d-check` und Dateien im Wegwerfverzeichnis. `time make gegenproben` ist weder vom Implementer noch von mir auf ruhigem Host gemessen (Plan §6, Adresse Verifier). | Uebergangsregel Nr. 2; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` | `harness/gegenproben.stempel` | ja — `git diff fe099d91..HEAD -- harness/gegenproben.stempel` leer | — |
| F-8 | INFO | Nicht selbst gefahren: `make gates` im Hauptbaum mit einem Arbeitsbaum unter `.wt/` (Plan §6, Adresse Verifier); die Groessenmessung wurde nachgerechnet: `git diff --numstat 0c354d5a..HEAD` ueber die Pfade des Blocks gibt 331 Einfuegungen und 11 Loeschungen, Summe 342, unter der Marke 1600. | `SPEC-128` Nr. 4 | Plan §6 · "Messung am Handoff nach `SPEC-128` Nr. 4: 342" | ja — `make groessenmarke` | — |

**Entscheidung zu `AGENTS.md` §3.6 (eigene, ohne den Befund des Architect zu uebernehmen).** Die Ausnahme ist keine Schwellen-Senkung. Gruende, am Diff und an Messungen: (1) `.a-check.yml` gewinnt eine Zeile `exclude`, Schichten, Kanten und `tech`-Listen bleiben, `.d-check.yml` gewinnt einen Eintrag in `scan.ignore`, Module, Klassen, Regeln und Grenzen bleiben (Diff gelesen: 5 und 3 Einfuegungen, keine Zeile geaendert ausser der `ignore`-Zeile). (2) Die Prüfmenge eines Standes aendert sich nicht: `git ls-files .wt` ist leer (0 Eintraege, gemessen), ein Commit verliert keine geprueften Dateien; ein Arbeitsbaum wird an seinem eigenen Ort von denselben Gates gelesen. (3) Der Absatz zum Carveout in §3.6 betrifft eine *befristete* Ausnahme fuer einen Teil des geprueften Standes; hier ist kein Teil des Standes ausgenommen. Vorbehalt: Die Bedingung (2) haengt an keinem Gate (F-2); solange sie unbewacht ist, traegt die Entscheidung den Satz "nie getrackt" nur als Gewohnheit. Das ist ein Mangel des Negativs, kein Grund fuer eine ADR; wuerde eine Datei unter `.wt/` getrackt, waere die Ausnahme eine Luecke und der Fall der Folge-Entscheidung in `SPEC-212` Nr. 2.

---

## Mutationstabelle (Zeile des Plans §3, eigene Mutation, eigenes Ergebnis)

Messender: der Reviewer. Last des Hosts je Lauf (Load Average 1 Minute, `/proc/loadavg`): 11,7 bis 13,4 (alle Mutationslaeufe), 7,2 (Zeitmessung). Jede Mutation: ein Lauf des Programms der Gegenproben im Container, `--memory=4g --memory-swap=4g --cpus=4`, ohne Netz, mutierte Konfiguration per `docker cp`; die Faelle sind deterministisch (kein Zufall, keine Nebenlaeufigkeit), eine Mutation ist rot in 1 von 1 Lauf. Basislauf ohne Mutation: 9 Faelle, 0 rot. Ein Fall zaehlt nur, wenn er aus dem benannten Grund rot wird: die Meldung nennt "erlaubter Fall abgelehnt (Exit 1)" bzw. "Verstoss nicht gemeldet", also den Gate-Befund, nicht den Aufbau (der Basislauf baut dieselben Verzeichnisse).

| Plan §3 | Zusage | Eigene Mutation | Rot (Faelle 205 bis 213) | Urteil |
|---|---|---|---|---|
| 1 | `a-check`: Verletzung unter `.wt/w/` ungemeldet | `exclude` entfernt | 205, 208 | gefangen |
| 1 | dasselbe | Muster `.wt/*` | 205, 208 | gefangen |
| 2 | `a-check`: Verletzung im Hauptbaum gemeldet | Muster `**` | 206, 207 | gefangen |
| 3 | `a-check`: `.wt` im Quellbaum geprueft | Muster `**/.wt/**` | 207 | gefangen |
| 4 | `a-check`: Arbeitsbaum im Arbeitsbaum ungemeldet | Muster `.wt/*/x/**` | 208 | gefangen |
| 5 | `docs-check`: Verweis unter `.wt/w/` ungemeldet | `".wt/**"` aus `scan.ignore` entfernt | 209, 212 | gefangen |
| 6 | `docs-check`: Verweis in `x/a.md` gemeldet | Muster `**` | 210, 211 | gefangen |
| 7 | `docs-check`: `x/.wt/b.md` geprueft | Muster `**/.wt/**` | 211 | gefangen |
| 8 | `docs-check`: `.wt/a/.wt/b/c.md` ungemeldet | Muster `.wt/*/*.md` | 212 | gefangen |
| 9 | `file` und `structure` lesen `.wt/` nicht | `file:` `AGENTS.md` zu `**/AGENTS.md` | 213 | gefangen |
| 9 | dasselbe | `structure:` `AGENTS.md` zu `**/AGENTS.md` | 213 | gefangen |
| 9 | dasselbe, Elemente ausser `AGENTS.md` (je ein Lauf) | `file:` `.claude/agents/*.md`, `.harness/skills/*.md`, `harness/conventions.md`, `.claude/commands/*.md` je zu `**/…` | 213 (4 von 4) | gefangen |
| 9 | dasselbe, Element, das Plan und Vertrag nicht nennen | `structure:` `harness/README.md` zu `**/harness/README.md` | keiner (0 von 9) | **ungefangen (F-1)** |
| 10 | Zeile und Eintrag stehen in den Konfigurationen | Zeile weg | wie Zeile 1 und 5 | gefangen |
| (eigen) | Wurzelverankerung des Eintrags | `".wt/**"` zu `".wt/"` | 209, 212 | gefangen |
| (eigen) | Wurzelverankerung des Eintrags | `".wt/**"` zu `"/.wt/**"` | 209, 212 | gefangen |
| 11, 12 | Negative (Gates ohne `.wt/`, getrackter Pfad, Kopie ohne `.git`, `mutationslauf.sh`) | kein Fänger, akzeptierte Negative; F-2 zum getrackten Pfad | — | siehe F-2, F-3 |

Klasse der Zusage (§3.17), aufgeloest gegen die Menge "jedes Gate und jedes Modul, das `.wt/` lesen kann": `make a-check` (1), `make docs-check` mit den Modulen `links` (gemessen, Faelle), `anchors`, `ids`, `codepaths` (gemessen am Pin, kein Fall), `matrix`, `spans` (nicht gemessen), `file`, `structure`, `targets`, `planning` (nicht lesend; `file` und `structure` je mit Fall bis auf die erste Regel von `structure`, `targets` selbst gemessen: eine `.mk`-Datei unter `.wt/w/` und `.wt/w/harness/mk/` bleibt ohne Befund, im Gegensatz zu derselben Datei an der Wurzel). Die uebrigen Gates der Kette folgen `SPEC-212` Nr. 1; eine eigene Gegenprobe zu `zeichen-check`, `lint`, `build` und weiteren habe ich nicht gefahren (Mechanismus gelesen: `.dockerignore` fuehrt `*` zuerst, `working-tree-hash.sh` und `baseline-verify.sh` lesen `git ls-files` bzw. ein benanntes Verzeichnis).

## Plan nachgezogen (§3.9) und Rohbytes (§3.18)

Plan §3 (Tabelle, Nachweis), §6 (Raster, Adressen mit Ausgang *offen*) und der Kopf tragen die Aenderungen des Implementers; `SPEC-102` fuehrt die Faelle 205 bis 213 in Festlegung, Aufbau, Folge, Randform 13, Stand und Grenze, `strom_test.go` zaehlt 44 und 169; Treffer von "174 bis 204" und "171 bis 173" ausserhalb dieser Stellen sind unveraendert richtig. Die Zeile *Nicht behauptet* in `harness/README.md` nennt Nr. 6 a, b, d, e und den Mutationslauf (Nr. 6 c), nicht die Module aus F-3. Rohbytes: kein Treffer (siehe oben).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `.a-check.yml` | geprueft, ohne Befund (Schluessel auf oberster Ebene, Kommentar der Klasse Abgrenzung, ein Tippfehler im Schluessel endet mit Exit 2) |
| `.d-check.yml` | geprueft, ohne Befund (Eintrag und Kommentar, Anker auf `SPEC-212` Nr. 3 b) |
| `tools/gegenproben/arbeitsbaeume.go`, `faelle.go` (Faelle 205 bis 208) | geprueft, ohne Befund ausser F-1 (Fixture gleicht nie dem Wert der Mutation; 205 `.wt/w/x/x.go` gegen Muster `.wt/*`, 208 `.wt/a/.wt/b/x/x.go` gegen `.wt/*/x/**`, beide rot) |
| `tools/gegenproben/arbeitsbaeume_test.go`, `strom_test.go`, `guide.go` | geprueft, ohne Befund (Namen, Reihenfolge, Hüllen 44 und 169; `gdBloecke` unveraendert in der Zeilenfolge; `gofmt`, `vet`, Tests und Lint der Stufen gruen im gebauten Stand) |
| `harness/vertraege/SPEC-039`, `SPEC-040`, `SPEC-102` | geprueft, Befund F-4 |
| `harness/vertraege/SPEC-212`, Index `harness/vertraege/README.md` | geprueft, Befunde F-2 und F-3 |
| `harness/README.md` | geprueft, ohne Befund (nennt nur, was `SPEC-212` Nr. 6 traegt) |
| Plan §3 und §6, Marke | geprueft, ohne Befund (342 nachgerechnet, 15 Laeufe der Nachweis-Tabelle nachgefahren) |
| Roadmap, `done/slice-gate-integration-determinism.md` | geprueft, Befund F-5 (Links nachgezogen, Start-Move rein, `R100`) |
| `harness/gegenproben.stempel`, `Dockerfile`, `.dockerignore`, `.golangci.yml`, `harness/mk/`, `tools/harness/` | geprueft, ohne Befund (unveraendert) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 2 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Fänger schmaler als Zusage (Menge der Regeln nicht erschoepft) · Negativ-Grund deckt die gemessene Form nicht · Negativ ohne Grund und Folge, Modul-Klasse unvollstaendig · Aussage nennt mehr Elemente als gemessen · Zustandsfeld nach Lifecycle-Move nicht nachgezogen

## Verdikt

**Merge-blockierend:** ja — drei MEDIUM (F-1, F-2, F-3); kein HIGH. Die neun Faelle und alle 15 Mutationen des Plans sind wie berichtet gefangen; die Befunde betreffen die Klasse der Zusage ausserhalb der genannten Mutationen.

**Übergabe:** Findings gehen an den Implementer; F-2 beruehrt eine Entscheidung des Architect (`SPEC-212` Nr. 2 und 6 a). Die Finding-Klassen gehen zusaetzlich in die Slice-Closure §7. Nicht geprueft und beim Verifier: `make gates` im Hauptbaum mit einem Arbeitsbaum unter `.wt/`, `time make gegenproben` auf ruhigem Host, die Gates `make lint`, `make build`, `make test`, `make test-race`, `make handbuch-check`, `make katalog-check`, `make migration-check`, `make schema-check` gegen `.wt/` (Mechanismus gelesen, kein Lauf mit Arbeitsbaum), `matrix` und `spans` gegen `.wt/` (kein Befund erzeugbar).

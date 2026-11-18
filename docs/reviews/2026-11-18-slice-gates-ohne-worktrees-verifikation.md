# Verifikations-Report: slice-gates-ohne-worktrees — 2026-11-18

**Rolle:** Verifier (Modul 11, `.harness/skills/closure-note-reviewer.md`); Frage: Bauen wir es richtig, gegen Plan, DoD und Vertrag. Nicht die Frage des Reviewers (Diff) und nicht die des Validators.

**Gegenstand:** `slice-gates-ohne-worktrees`, Zweig `slice-gates-ohne-worktrees`, HEAD `969e80e9` (enthält origin/main `7c888289`); die Commits des Slice laut `git log --no-merges fe099d91..HEAD`, soweit sie Slice sind (`04b27b41` bis `0c37f71e`, ohne die Commits von `slice-prefer-return`). Plan: [slice-gates-ohne-worktrees](../plan/planning/done/slice-gates-ohne-worktrees.md). Vertrag: [SPEC-212](../../harness/vertraege/SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md) Nr. 1 bis 6; Nachzüge [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md), [SPEC-039](../../harness/vertraege/SPEC-039-architektur-gate.md), [SPEC-040](../../harness/vertraege/SPEC-040-docs-check.md), `harness/README.md`. Bezug: `LH-QA-06`, [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md). Vorgänger: [Review 1](2026-11-18-slice-gates-ohne-worktrees-review.md), [Review 2](2026-11-18-slice-gates-ohne-worktrees-review-2.md).

**Vorgehen.** Ich habe die Behauptungen der Implementer- und Review-Berichte nicht übernommen, sondern die Sensoren selbst gefahren (unten je mit Last). Der Zweig blieb unverändert: `git status --short` vor und nach den Läufen leer (die Arbeitsbäume unter `.wt/` und Hilfsdateien liegen ausserhalb von git; `git check-ignore -v .wt/probe` trifft `/.wt/` in `.git/info/exclude`). `harness/gegenproben.stempel` ist gegenüber `fe099d91` unverändert (`git diff --stat` leer). `Dockerfile`, `.dockerignore`, `.golangci.yml`, `harness/mk/`, `tools/harness/` sind gegenüber `fe099d91` unverändert (leerer Diff), wie der Plan §3 zusagt.

## 1. Eigene Läufe (Belege)

| Lauf | Ergebnis | Last (1 min) |
|---|---|---|
| Probe-Arbeitsbaum im Worktree des Zweigs: Kopie von `src/` (773 Dateien) und eine Verletzung `x/x.go` (importiert das Modell, liegt in keiner Schicht) unter dem Pfad .wt/probe; `make a-check` | `gesamt: 0 Befund(e)`, Exit 0 | 12,1 |
| Dieselbe Probe, `.a-check.yml` ohne die Zeile `exclude` (Kopie unter .wt/mut, gleicher Pin `a-check@sha256:9738…`, `--memory=4g --cpus=4`) | Exit 1, `gesamt: 1341 Befund(e)`, alle `wrong-direction` unter .wt/probe | 12 |
| Dieselbe Probe, Kopie mit der Konfiguration des Zweigs | `gesamt: 0 Befund(e)` | 12 |
| Hauptbaum `/home/builder/work/odata-recorder`, `make a-check` (dort fehlt die Ausnahme; der Zweig trägt sie) | `gesamt: 5387 Befund(e)` (Arbeitsbaum des Zweigs samt Probe wird mitgelesen) | 12 |
| Hauptbaum, `make docs-check` | Exit 1, Befunde unter .wt/gates-ohne-wt (`codepath-missing`) | 9 |
| Zweig, `make docs-check` mit Probe und einer defekten Datei b.md unter .wt/probe | `904 Datei(en) geprüft, 0 Befund(e)`; mit zusätzlicher defekter Datei `x/a.md`: `905 Datei(en)`, genau 1 Befund `x/a.md:3 target-missing`; nach dem Löschen wieder 0 | 9 |
| `make groessenmarke` | `ok, Messung 1057 von Marke 2300`; Messpunkt `zeilen=16 produktcode=375 schaetzung=1655 marke=2300 messung=1057` | 12 |
| `make zeichen-check` | `zeichen: ok, 1893 Datei(en) geprüft` | 12 |
| `make gates` im Zweig mit dem Probe-Arbeitsbaum unter .wt/ (unter `flock` auf `.wt/heavy.lock`, Beginn bei Last 11,4; Ende bei 20,5, Spitze 25 durch fremde Läufe) | **Exit 0**; darin `make lint` mit `0 issues`, `zeichen: ok`, `make a-check`, `make docs-check`, Docker-Gates (Test, Race, Build, Handbuch, Katalog, Schema, Migration, Integration, Determinismus, Kontrakt), Gate-Nachweis | 11 bis 25 |
| Gegenproben-Fälle 205 bis 216 (`fahre(standard(), faelle()[Ende-12:], ".")` im Container aus dem Image `odata-recorder:gp-wt`, Stand `HEAD` von `tools/gegenproben` per `docker cp`, echte Konfigurationen in `/src`, ohne Netz, `--memory=4g --memory-swap=4g --cpus=4`) | ohne Mutation: 12 Fälle, 0 rot (je 0,4 s) | 6 bis 12 |
| `make gegenproben` als Ganzes / Stufe `gegenproben` des `Dockerfile` | **nicht gefahren** (siehe Abschnitt 6) | — |

Der Probe-Arbeitsbaum ist ein Stand-in: eine Kopie ohne `.git` im Worktree des Zweigs. Er belegt, dass beide Gates ein Verzeichnis unter .wt/ mit Verletzungen nicht mehr lesen, und dass `make gates` mit einem solchen Verzeichnis grün ist. Er ersetzt nicht den Lauf im Hauptbaum (Abschnitt 4, DoD 2).

Nach der Messung gelöscht (Probe, Kopien, Ausgabedatei); `git status --short` leer.

## 2. Stichprobe: eigene Mutationen an Zeilen der Tabelle (nicht die der Reviewer)

Messender: der Verifier; Konfiguration je Lauf als Kopie, die Fälle 205 bis 216 gegen sie gefahren (Aufbau wie oben). Alle Fälle sind deterministisch (ein `a-check`- oder `d-check`-Prozess je Fall): 1 von 1 Lauf rot je Mutation. Last 9,3 bis 12,4.

| Plan §3 Zeile | Mutation (nicht in Review 1 oder 2 gefahren) | Roter Fall, Grund | Urteil |
|---|---|---|---|
| 2 | `.a-check.yml`: `exclude: [".wt/**", "x/**"]` (verdeckt den Hauptbaum) | Fall 206 `arbeitsbaum-verletzung-im-hauptbaum`: „Verstoß nicht gemeldet“ | gefangen |
| 6, 7, 11 | `.d-check.yml`: `scan.ignore` um `x/**` ergänzt | Fälle 210 `arbeitsbaum-verweis-im-hauptbaum`, 211 `wt-im-quellbaum-bleibt-geprueft`, 215 `module-im-hauptbaum-gemeldet`: „Verstoß nicht gemeldet“ | gefangen (drei von drei erwarteten) |
| 4 | `.a-check.yml`: `exclude: [".wt/w/**"]` (nur ein Name) | Fall 208 `arbeitsbaum-im-arbeitsbaum`: „erlaubter Fall abgelehnt (Exit 1)“, `wrong-direction` unter `.wt/a/.wt/b/x/x.go` | gefangen; 205 bleibt grün (richtig: `.wt/w/` ist ausgenommen) |
| 8 | `.d-check.yml`: `".wt/**"` zu `".wt/w/**"` | Fall 212 `arbeitsbaum-im-arbeitsbaum`: `target-missing` unter `.wt/a/.wt/b/c.md`; 209 grün | gefangen |
| 13 | `.d-check.yml`: `targets` aus `modules:` genommen | Aufbau von Fall 213: `Eingriff … modules: nennt "targets" nicht, die Listen von SPEC-212 Nr. 3 b führen es`, 0 Fälle gefahren (fail-closed) | gefangen, in der Stufe, nicht in `go test` (so im Plan §3 und §6 gesagt) |

Klassen-Auflösung (`AGENTS.md` §3.17) für die Fix-Befunde beider Runden: F-1 (Menge „Regeln von `structure` und `file`“, 7 Regeln, 15 Paare, 14 Dateien) und F-3 (Menge „Module von d-check“, 10, davon 6 lesend) hat Review 2 an allen Elementen gefahren, darunter vier, die weder Befund noch Vertrag nannten; ich habe die Mengen am Code und an der echten Konfiguration nachgezählt (`.d-check.yml` Zeile `modules:` führt zehn Module; Fall 213 als Schleife, Fälle 214 und 215 als Schleife über sechs, der Aufbau bricht bei einem Modul ausserhalb der Listen) und die obigen fünf Mutationen an Elementen gefahren, die dort nicht standen. Die Fänger sind Schleifen über die Mengen, keine Aufzählungen. Kein Element der Klasse blieb ungefangen. F-2 (getrackter Pfad) ist ein akzeptiertes Negativ mit Grund und Folge in `SPEC-212` Nr. 6 a.

## 3. Tabelle gegen den Vertrag (`AGENTS.md` §3.15)

**`SPEC-212` Nr. 5, jede Zeile:** Zeilen 1 bis 13 des Vertrags entsprechen den Zeilen 1 bis 13 der Tabelle von Plan §3 (Fall 205 bis 216 plus Zeile 13), Zeile 14 des Plans trägt Nr. 6 g (Verdrahtung: Fall 1 und 5), Zeilen 15 und 16 tragen Nr. 6 als akzeptierte Negative. Jede Zeile hat Fall, Mutation und rot gesehene Fälle; beide Seiten jeder Grenze führt ein Paar (1/2, 5/6, 10/11; 9 gegen Fall 167 und Zeile 12).

**`SPEC-212` Nr. 4, jede Randform:**

| Nr. 4 | Fänger oder Negativ im Vertrag selbst |
|---|---|
| 1, 2, 3, 8, 10, 13 | Fänger (Nr. 5 Zeilen 1 bis 12) |
| 4, 5, 6, 14 | akzeptiertes Negativ Nr. 6 d (Grund: Eigenschaft der Werkzeuge am Pin; Folge: Pin-Wechsel kann sie ändern, Prüfsumme der Gegenproben ändert sich) |
| 7 | „trivial, kein Fall“ (kein Verhalten) |
| 11 | Nr. 6 a (Grund, Folge) |
| 12 | Nr. 6 b (Grund, Folge) |
| 9 | **kein Fänger, kein Negativ**: die Zelle sagt „nicht gelesen: nicht im Mount“, Ort „der Mount ist die Wurzel des Hauptbaums“; weder Nr. 5 noch Nr. 6 noch Plan §3 führt eine Zeile dazu (V-6) |

**`SPEC-212` Nr. 6:** a, b, c, d, e, f tragen Grund und Folge; g trägt den Fänger (Zeilen 1 und 5). Nr. 6 f (nicht gefahrene Gates) ist mit diesem Report für `make gates` eingelöst, siehe Abschnitt 1.

**Ein Negativ ohne Grund und Folge im Vertrag selbst:** die feste Menge der Pfade (Stand-Absatz von `SPEC-212`: „Die Menge der Pfade … ist fest: eine neue Datei … fängt kein Fall“). Der Plan trägt die Folge (§3, letzter Absatz; §1), der Vertrag nennt weder Grund noch Folge und führt das Negativ nicht in Nr. 6 (V-1, gleich G-3 aus Review 2, nicht behoben).

## 4. DoD, Punkt für Punkt (Plan §2)

| # | Punkt | Urteil | Beleg |
|---|---|---|---|
| 1 | Architect: welche Gates lesen `.wt/`, ohne ADR | **erfüllt** | `SPEC-212` Nr. 1 (Messung 1351 und 39 Befunde), Nr. 2; ich habe die Zahl der Befunde im Hauptbaum und im Zweig reproduziert (Abschnitt 1: mit Zeile 0, ohne 1341 an der Probe); keine ADR-Nummer 0030 bis 0039 vergeben (Plan §6) |
| 2 | Jedes betroffene Gate nimmt `.wt/` aus, Fänger je Zeile von Nr. 5; `make gates` grün, auch im Hauptbaum mit einem Arbeitsbaum | **Konfiguration und Fänger: erfüllt. `make gates` im Zweig mit Stand-in-Arbeitsbaum: erfüllt (Exit 0). `make gates` im Hauptbaum: erst nach dem Integrationsschritt** | Zeilen in `.a-check.yml` (`exclude: [".wt/**"]`) und `.d-check.yml` (`scan.ignore`); Abschnitt 1 und 2. Der Hauptbaum trägt die Ausnahme nicht (`make a-check` dort 5387 Befunde, `make docs-check` rot), bis der Zweig integriert ist; die Checkbox bleibt offen bis dahin, der Planner haken bei Closure. Eine zweite Bedingung der Gegenproben: die Stufe `gegenproben` wurde nicht gefahren (Abschnitt 6) |
| 3 | Review durchgeführt, Report unter `docs/reviews/` | **erfüllt** | zwei Reports; Runde 2: 0 HIGH, 0 MEDIUM, 5 LOW, 4 INFO |
| 4 | Closure-Notiz mit Steering-Loop-Lerneintrag | **nicht erfüllt, Aufgabe des Planners** | §7 steht auf „noch nicht begonnen“ |
| 5 | Beobachtungs-Register fortgeschrieben (`BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt`) | **offen, Aufgabe des Planners bei Closure** | nicht Gegenstand dieser Runde |
| 6 | Jedes Risiko aus §6 trägt einen Ausgang | **erfüllt in der Form**: jede Zeile trägt einen Ausgang; fünf stehen auf *weiter offen* oder *offen* (siehe 5) und werden bei Closure aufgelöst | Plan §6 |
| 7 | Drei Paarungen (Anker, Folge-Slice, Register) | **offen, Closure** | — |

## 5. Plan §6: Adressen und Ausgänge (`AGENTS.md` §3.13)

| Zeile §6 (Zeile in der Datei) | Ausgang | Adresse | Nehmer führt die Sendung? | Urteil |
|---|---|---|---|---|
| Größe, Marke 2300 (204) | weiter offen | Planner (Closure) | n. a. (Planner) | in Ordnung; Rechnung stimmt (*S* = 80·16 + 375 = 1655; 135·1655 = 223425; aufgerundet 23, 2300; unter 300000); Messung 1057 von 2300 selbst gefahren |
| Zeitbudget von `make gegenproben` ungemessen (205) | weiter offen | „Verifier (`time make gegenproben` auf ruhigem Host)“ | **nein**: der Nehmer `slice-gegenproben-zeitbudget` (liegt in `open/`) trägt die Sendung weder in §1 noch in der DoD (Suche nach `gates-ohne-worktrees`, `205` und dem Lauf unter Last ergibt keinen Treffer); eine Rolle ist keine Adresse | V-3 |
| Kollision mit Kettenslice (208) | weiter offen | Start laut §4 | n. a. | in Ordnung |
| nicht gefahrene Gates (209) | weiter offen | Verifier | n. a. (Rolle, die jetzt liefert) | mit Abschnitt 1 für `make gates` im Zweig **erledigt**; im Hauptbaum erst nach Integration |
| Pfad der Arbeitsbäume, `.gitignore` (210) | weiter offen | Planner | Entscheidung beim User, Planner ist die Adresse | in Ordnung |
| Roadmap „angelegt in `open/`“ (212) | offen | Planner | Roadmap Zeile 26 und Drift-Zeile 175 unverändert | V-4 |
| Zeile 13 gleich „Test“ gegen „Aufbau“ (215) | weiter offen | **Reviewer** | Reviewer hat Runde 2 gefahren und die Zeile an den Architect zurückgegeben; Adresse nicht mehr gültig | V-2 |

## 6. Was ich nicht prüfen konnte

- **`make gegenproben` / `time make gegenproben`**: nicht gefahren. Der Stempel `harness/gegenproben.stempel` ist nach Übergangsregel Nr. 2 beim Integrationsschritt nachzuziehen und darf von dieser Rolle nicht verändert werden; ein grüner Lauf von `make gegenproben` im Zweig schriebe ihn neu, weil die Eingaben sich geändert haben. Damit sind das Zeitbudget (SPEC-102) und die Frist der neun Fälle `lint/*` unter Last, der Fall `build/ohne-trimpath` (G-9) und der Aufbau der Stufe (`go vet`, `go test` und das Image mit den Binaries) von mir **nicht** gemessen. Gefahren habe ich die zwölf Fälle 205 bis 216 gegen die echten Konfigurationen im Image `odata-recorder:gp-wt` (gleiche Stufe, Programme `a-check`, `d-check` des Pins) und `make gates` (das die Gegenproben nicht enthält, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)).
- **`make gates` im Hauptbaum mit einem echten Arbeitsbaum**: der Hauptbaum hat die Ausnahme erst nach der Integration; ich fuhr stattdessen `make gates` im Zweig mit dem Probe-Arbeitsbaum als Stand-in. Eine Abweichung, die nur der Hauptbaum zeigt (zum Beispiel ein Gate, das `git ls-files` mit anderem Ausschluss liest), ist damit nicht ausgeschlossen. Die Gates, die `.wt/` nicht lesen (Nr. 1 des Vertrags), blieben im Lauf mit Probe grün; der Befund „kein weiteres Gate liest `.wt/`“ ist damit für den Zweig bestätigt.
- Symlinks, ein fehlendes `.wt/`, `planning` und `targets` gegen `.wt/` (kein Fall, Nr. 6 d), sowie die Randformen des Musters für fünf der sechs lesenden Module (Nr. 6 d).
- Die Last schwankte zwischen 6 und 25 (fremde Läufe); die Fälle 205 bis 216 sind deterministisch, die Wiederholung einer Mutation ändert nichts an der Erwartung, sodass die Last die Urteile nicht berührt.

## 7. Befunde

Rohbytes: alle 16 geänderten Dateien des Slice (`.a-check.yml`, `.d-check.yml`, `tools/gegenproben/*`, `harness/vertraege/*`, `harness/README.md`, der Plan, beide Reports) ohne Treffer auf BOM, NBSP, Nullbreiten- und Zeilentrenner, Steuerzeichen, CR und Folge Backslash-u-vier-Hex (`LC_ALL=C grep -P`); dieser Report ist am Ende ebenso gegengelesen.

| ID | Schwere | Befund | Fundstelle | Zuständig |
|---|---|---|---|---|
| V-1 | MEDIUM (nach dem Wortlaut von `AGENTS.md` §3.15: ein Negativ ohne Grund und Folge ist eine Zusage ohne Fänger; Wirkung eng, nicht merge-blockierend, weil Plan §3 die Folge nennt und der Aufbau eine Regel ohne Pfad fängt) | Das akzeptierte Negativ „die Menge der 14 Pfade ist fest, eine neue Datei unter einem Glob fängt kein Fall“ steht im Vertrag nur im Stand-Absatz, ohne Grund und ohne Folge, und fehlt in Nr. 6. Der Plan nennt die Folge (§3, Absatz zu Zeile 13), nicht den Grund (die Menge ist `gdMenge()` aus `SPEC-040` Nr. 2 plus `harness/README.md`). Gleich G-3 aus Review 2, durch die Architect-Runde nicht behoben | `harness/vertraege/SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md:94` (Stand), Nr. 6 (Zeilen 85 bis 92) ohne Eintrag | Architect |
| V-2 | LOW | `SPEC-212` Nr. 5 Zeile 13 nennt als Fänger „Test im Programm der Gegenproben: jedes Modul … der echten `.d-check.yml`“; der Träger ist der Aufbau der Fälle 213 bis 215 in der Stufe (Stand-Absatz, Plan §3 und §6 sagen es). Der Vertrag trägt zwei Aussagen über denselben Fänger. §6 hält die Zeile mit Adresse *Reviewer*, der Runde 2 durchlief; ein Bericht allein ist keine Adresse (§3.13). Gleich G-2 | `harness/vertraege/SPEC-212-…:82` gegen `:94`; `docs/plan/planning/done/slice-gates-ohne-worktrees.md:215` | Architect (Wortlaut), Planner (Adresse im Plan) |
| V-3 | LOW | Die Zeile des Plans zum Zeitbudget (Fälle 205 bis 216 ungemessen, neun Fälle `lint/*` unter Last rot) nennt als Adresse „Verifier“. Eine Rolle ist keine Adresse, und der Verifier kann sie hier nicht einlösen (Stempel, Abschnitt 6). Der Nehmer `slice-gegenproben-zeitbudget` (`open/`) trägt die Sendung nicht in §1 oder DoD; die Beobachtung `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` steht dort ebenfalls nicht. §3.13: die Adresse ist ungültig, der Punkt geht an den Planner | `docs/plan/planning/done/slice-gates-ohne-worktrees.md:205`; `docs/plan/planning/open/slice-gegenproben-zeitbudget.md` §1 und §2 | Planner: entweder die Sendung (Fälle 205 bis 216 in der Messung, Last als Ursache roter `lint/*`) in §1 oder DoD des Nehmers eintragen, mit der Kennung dieses Slice, oder einen neuen Slice anlegen; bis dahin ist das Zeitbudget unbelegt |
| V-4 | LOW | Zustandsfelder der Roadmap nicht nachgezogen: „angelegt in `open/`“ und „Reihenfolge legt der User beim Start fest“, obwohl der Slice in `in-progress/` liegt. Offen und an den Planner adressiert (§6); gleich F-5 und G-5 | `docs/plan/planning/in-progress/roadmap.md:26`, `:175` | Planner (Closure) |
| V-5 | LOW | Plan §3, erster Block *Nachweis* beschreibt den Stand vor der Fix-Runde („alle neun Fälle“, „meldet 213 Fälle, 0 rot“, 15 Läufe, Zeile 10 „Zeile bzw. Eintrag weg“), nennt ihn „den Stand dieses Plans“ und steht neben dem zweiten Block mit 216 Fällen. Gleich G-4, nicht nachgezogen | `docs/plan/planning/done/slice-gates-ohne-worktrees.md:118`, `:136` | Implementer/Planner |
| V-6 | INFO | `SPEC-212` Nr. 4 Zeile 9 (Arbeitsbaum ausserhalb des Repos, „nicht im Mount“) hat weder Fänger noch ein Negativ in Nr. 6; die Zelle trägt nur den Ort „Mount ist die Wurzel“ (`a-check.mk`, `d-check.mk`). Verhalten, keine Zusage dieses Repos; ein Satz in Nr. 6 d würde es schliessen | `harness/vertraege/SPEC-212-…:60` | Architect |
| V-7 | INFO | Marke 2300 nach dem ersten Code-Commit auf 1600 ersetzt (G-1): im Plan §4 und §6 ehrlich benannt (`ersetzt: 1600 \| …` mit Grund; „keine Antwort auf eine Messung“); nachgerechnet, stimmt; die Messung (1057) liegt unter beiden Marken, die Ersetzung war nicht nötig. `AGENTS.md` §3.14 lässt das Ersetzen nur vor dem ersten Code-Commit zu; keine Folge für das Urteil | Plan §4 Block, §6 | Planner (zur Kenntnis) |
| V-8 | INFO | Kommentar von `abModuleDateien` beschreibt die Schleife ungenau (G-7); Umfang 975 von 1057 Zeilen im Parser der Blöcke (G-8); `build/ohne-trimpath` rot im Lauf des Implementers unter Last 175 bis 222, von Review 2 bei Last 10 zweimal grün gesehen (G-9), von mir nicht wiederholt (Stufe nicht gefahren) | `tools/gegenproben/arbeitsbaeume.go` | Implementer (Pflege) |

Von den neun Befunden der zweiten Runde habe ich G-1 bis G-5 als V-7, V-2/V-1, V-5, V-4 selbst am Text bestätigt (alle noch vorhanden, keiner blockierend); G-6 (Wortlaut „`git add` legt seinen Inhalt nicht ab“ gegen den Eintrag `160000`, kein Einfluss auf die Gates) und G-7 bis G-9 habe ich nicht neu bewertet über V-8 hinaus.

## 8. Urteil

**Der Slice ist bestanden, mit Auflagen vor der Closure; kein HIGH, kein blockierender Befund.**

- **Zweck belegt:** Mit der Ausnahme im Zweig meldet `make a-check` an einem Arbeitsbaum-Stand-in 0 Befunde, ohne sie 1341 `wrong-direction`; `make docs-check` meldet eine defekte Datei unter dem Arbeitsbaum nicht, dieselbe Datei im Hauptbaum genau einmal. Die Fänger der Zeilen 1 bis 13 sind in meiner Stichprobe an fünf eigenen Mutationen aus dem benannten Grund rot (Abschnitt 2); die Klassen (7 Regeln, 10 Module, 4 Arbeitsbaum-Formen) sind als Schleifen gefangen. `make gates` ist im Zweig mit Probe-Arbeitsbaum Exit 0 (nicht behauptet für den Hauptbaum).
- **Erst nach dem Integrationsschritt erfüllbar:** `make gates` und die Gates `make a-check`, `make docs-check` im Hauptbaum mit einem echten Arbeitsbaum (DoD 2, §5 Closure-Trigger); der Stempel der Gegenproben und `make gegenproben` samt Zeitbudget (Übergangsregel Nr. 2).
- **Offen für den Planner und den Architect vor der Closure:** V-1 und V-2 (Wortlaut von `SPEC-212`), V-3 (Adresse des Zeitbudgets, Nehmer trägt die Sendung nicht), V-4 und V-5 (Roadmap, Plan §3); DoD 4 bis 7 sind Aufgaben der Closure.
- **Nicht behauptet:** ein grünes `make gegenproben`, ein grünes `make gates` im Hauptbaum, eine Messung des Zeitbudgets.

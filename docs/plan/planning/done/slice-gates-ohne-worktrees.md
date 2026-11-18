# Slice slice-gates-ohne-worktrees: Worktrees unter .wt/ aus den Gates des Hauptbaums ausnehmen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): berührt voraussichtlich `.a-check.yml`, `.d-check.yml`, `.golangci.yml` oder `harness/mk/`, also Nr. 7 (ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an den Zuschnitt der Gates, der Bezug ist der nächste, den es gibt), [ADR-0002](../../adr/0002-architektur-gate-a-check.md) (Architektur-Gate `a-check`), [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Doku-Gate), [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Größenmarke). Geber: `slice-gate-integration-determinism` (Befund aus dem Parallelbetrieb, Risiko in dessen §6 mit Ausgang *weiter offen*).

**Berührte Spec-Stellen:** [`SPEC-212`](../../../../harness/vertraege/SPEC-212-arbeitsbaeume-ausserhalb-der-pruefmenge.md) (neu, Architect: Bestand, Bindung ohne ADR, Ausnahme je Gate, Randformen, Fänger) · `SPEC-039` (Gate `make a-check`) · `SPEC-040` (Gate `make docs-check`) · `SPEC-102` (Gegenproben: neue Fälle nach der Folge des Vertrags).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner der Closure von `slice-gate-integration-determinism`). **Datum:** 2026-11-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Ein Gate im Hauptbaum liest die Worktrees unter `.wt/` nicht mit. Ein Lauf von `make a-check` im Hauptbaum meldete im Parallelbetrieb 1414 Befunde `wrong-direction` aus den Arbeitsbäumen der Linien (Meldung der Linie von `w3`); jeder Arbeitsbaum ist eine volle Kopie des Repos, und ein Gate, das `.` scannt, liest sie als Teil des Hauptbaums.

**Hintergrund (Geber `slice-gate-integration-determinism`):** Der Architect hat gemessen (`SPEC-212` Nr. 1): Nur `make a-check` (1351 Befunde) und `make docs-check` (39 Befunde) lesen `.wt/` im Hauptbaum; alle übrigen Gates tun es nicht (Kontext als Allowlist, `git ls-files`, benannte Pfade). Die Abhilfe ist je Gate ein Muster, `.wt/**` an der Wurzel; sie braucht keine ADR, weil keine Schwelle und keine Regel sinkt und kein Arbeitsablauf `.wt/` committet (`SPEC-212` Nr. 2, Nr. 6 a). Von den zehn Modulen von `d-check` lesen sechs `scan.ignore`, vier nicht (`SPEC-212` Nr. 3 b).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung der Übergangsregeln für Worktrees (Pfad, Zahl, Sperre) — Sache des Planners und des Users; der Slice ändert nur, was die Gates lesen.
- Eine Lockerung eines Schwellenwerts oder einer Architekturregel — `AGENTS.md` §3.6; ausgenommen wird ein Pfad, keine Regel.
- Die Gate-Kette selbst (`make gates`, `GATE_CHECKS`) — `slice-gate-integration-determinism` hat sie festgelegt (`SPEC-211`).
- `tools/harness/mutationslauf.sh` (kopiert `.wt/` in den Container, kein Gate, ändert kein Urteil) — akzeptiertes Negativ, `SPEC-212` Nr. 6 c.
- Eine neue Datei unter einem Glob von `structure` oder `file`, die in der festen Menge der 14 Pfade des Aufbaus von Fall 213 fehlt — akzeptiertes Negativ (`SPEC-212` Nr. 6 h): die Menge ist der Stand der Konfiguration, kein Fall liest das Dateisystem des Repos.
- Eine versionierte `.gitignore` mit `/.wt/` — sie ändert, was `git ls-files --exclude-standard` den Gates liefert, und ist keine Gate-Konfiguration; offene Entscheidung für Planner und User (§6, `SPEC-212` Nr. 6 a).

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [x] Der Architect hat festgestellt, welche Gates `.wt/` im Hauptbaum lesen (`SPEC-212` Nr. 1: `make a-check` und `make docs-check`; `make zeichen-check` und die übrigen nicht), und entschieden, dass die Abhilfe keine ADR braucht (`SPEC-212` Nr. 2).
- [ ] Jedes betroffene Gate nimmt `.wt/` aus (`.a-check.yml`: `exclude: [".wt/**"]`; `.d-check.yml`: `".wt/**"` in `scan.ignore`, `SPEC-212` Nr. 3), mit einem Fänger je Zeile von `SPEC-212` Nr. 5 (ein Worktree mit einer Verletzung bleibt ungemeldet, dieselbe Verletzung im Hauptbaum wird gemeldet); `make gates` grün, auch im Hauptbaum mit einem Arbeitsbaum unter `.wt/`.
      **Offen, mit Grund:** Konfiguration, Fänger (Fälle 205 bis 216) und `make gates` im Zweig mit einem Probe-Arbeitsbaum unter `.wt/` (Exit 0) sind belegt; `make gates` im Hauptbaum mit einem echten Arbeitsbaum ist erst nach dem Integrationsschritt erfüllbar, weil der Hauptbaum die Ausnahme erst nach dem Merge trägt. Beleg des Verifiers für den Zweck: `make a-check` meldet am Probe-Arbeitsbaum 0 Befunde mit der Ausnahme und 1341 `wrong-direction` ohne sie; `make docs-check` meldet einen Defekt unter `.wt/` nicht und denselben Defekt im Hauptbaum genau einmal. Adresse: Integrationsschritt (Übergangsregeln Nr. 5: `make gates` im Hauptbaum unter der Sperre).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben (Auslöser `BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt`); keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect hat entschieden (`SPEC-212`), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.a-check.yml` | update | Schlüssel `exclude: [".wt/**"]` (`SPEC-212` Nr. 3 a), mit Kommentar der Klasse Abgrenzung |
| `.d-check.yml` | update | Eintrag `".wt/**"` in `scan.ignore` samt Zeile im Kommentar davor (`SPEC-212` Nr. 3 b) |
| `tools/gegenproben/faelle.go` | update | `aCheckArbeitsbaeumeFaelle`: Fälle 205 bis 208 (`SPEC-212` Nr. 5 Zeilen 1 bis 4) neben `aCheckTestsFaelle`, in `faelle()` hinter den Fällen 174 bis 204 |
| `tools/gegenproben/arbeitsbaeume.go`, `arbeitsbaeume_test.go` | new | Fälle 209 bis 216 (Zeilen 5 bis 12): die vier Fälle 209 bis 212 mit dem Block `scan` der echten `.d-check.yml` und `modules: [links]`; Fall 213 mit den echten Blöcken `scan`, `structure`, `file` als Schleife über jede Regel beider Blöcke und jede Datei ihres Globs; die Fälle 214 und 215 (ein Defekt je lesendem Modul unter `.wt/w/` und in `x/`) und 216 (Zellgrenze im Hauptbaum); Tests der Fall-Folge, der Konfiguration, des Laufs gegen d-check und der Mengen (Zeile 13: zehn Module, sieben Regeln) |
| `tools/gegenproben/guide.go` | update | `gdBloecke` aus `gdEchtKonfig` herausgezogen (dieselbe Zeilenfolge, der Kopf ist Parameter) |
| `tools/gegenproben/strom_test.go` | update | Zählung der Hüllen: 44 gemeinsam, 172 getrennt |
| `harness/vertraege/SPEC-102-gegenproben.md` | update (Implementer) | die Einträge der Fälle 214 bis 216 in der Folge des Vertrags, ihre Zeile in der Festlegung, Randform 13, Stand; Fall 213 und die Grenze nach `SPEC-212` Nr. 3 b stehen im Vertrag |
| `harness/vertraege/SPEC-039-…`, `SPEC-040-…`, `SPEC-212-…` | update (Implementer) | je ein Zeiger auf `SPEC-212` in der Festlegung (`SPEC-039` Nr. 3), die Fälle in Gegenprobe bzw. Nr. 11 und Stand (`SPEC-040`: Fälle 214 bis 216); `SPEC-212` *Stand* auf vorhanden |
| `harness/README.md` | update (Implementer) | Absatz *Nicht behauptet*: die Ausnahme `.wt/**` und was von ihr nicht behauptet ist (`SPEC-212` Nr. 6 a bis e; Nr. 6 d führt auch `targets` und `planning`); Verdrahtung der Zeilen (Nr. 6 g) fängt Fall 205 und 209 |
| `.golangci.yml`, `harness/mk/`, `.dockerignore`, `tools/harness/*.sh` | keine Änderung | die Messung zeigt kein weiteres Gate (`SPEC-212` Nr. 1) |

**Erwartete Produktcode-Größe (für die Marke des Planners, `SPEC-128` Nr. 1):** Summe aus Einfügungen und Löschungen über `.a-check.yml`, `.d-check.yml`, `tools/gegenproben/` (Tests eingeschlossen) und `harness/README.md`, geschätzt 150 bis 220 Zeilen: Konfigurationen rund 15 (Zeile und Kommentar je Datei), a-check-Fälle rund 50, d-check-Fälle samt Konfigurationsbau rund 90, Tests der Fall-Folge rund 20, `harness/README.md` rund 2. Dazu kommen die Fänger der Zeilen 9 bis 13 (Schleife über die Regeln in Fall 213, Fälle 214 bis 216, Test der Mengen, Tests dazu), geschätzt weitere 150 bis 230 Zeilen. `harness/vertraege/` zählt die Messung nicht mit (`SPEC-128`). Gesamt 300 bis 450 Zeilen, Mitte 375 (*p* in §4). Kleiner Slice, kein Schnitt erwartet.

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (`SPEC-212`, dieser Abschnitt, Raster in §6; erledigt) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock` auf die Sperre der Linien; Zeichen jenseits von ASCII an Bytewerten gebildet und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Quelle `SPEC-212` Nr. 5, die Mutation sieht der Implementer selbst rot und trägt Rate und Messenden hier ein):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | `a-check`: Verletzung unter `.wt/w/…` bleibt ungemeldet | Fall der Gegenproben: Datei außerhalb jeder Schicht, die ein Paket des Modells importiert, unter `.wt/w/x/x.go`: Exit 0 | `exclude` entfernt; Muster `.wt/*` | andere |
| 2 | `a-check`: dieselbe Datei im Hauptbaum wird gemeldet | `x/x.go`: Exit 1, `wrong-direction` | Muster `**` | andere |
| 3 | `a-check`: ein Verzeichnis `.wt` im Quellbaum bleibt geprüft | `src/hexagon/model/.wt/x.go` importiert einen Dienst: Exit 1, `core-impurity` | Muster `**/.wt/**` | andere |
| 4 | `a-check`: Arbeitsbaum in einem Arbeitsbaum bleibt ungemeldet | `.wt/a/.wt/b/x/x.go`: Exit 0 | Muster, das nur die erste Ebene trifft | andere |
| 5 | `docs-check`: fehlendes Linkziel unter `.wt/w/` bleibt ungemeldet | Exit 0, `0 Befund(e)` | `".wt/**"` aus `scan.ignore` entfernt | andere |
| 6 | `docs-check`: derselbe Verweis in `x/a.md` wird gemeldet | Exit 1, `target-missing` | Muster `**` | andere |
| 7 | `docs-check`: `x/.wt/b.md` bleibt geprüft | Exit 1, `target-missing` | Muster `**/.wt/**` | andere |
| 8 | `docs-check`: `.wt/a/.wt/b/c.md` bleibt ungemeldet | Exit 0 | Muster, das nur die erste Ebene trifft | andere |
| 9 | `docs-check`: `file` und `structure` lesen `.wt/` nicht, jede Regel beider Blöcke | Fall 213: Schleife über jede Regel beider Blöcke der echten Konfiguration (7) und jede Datei ihres Globs (15 Paare, 14 Dateien), unter `.wt/w/<Pfad>` die Eingabe, die die Regel verletzt: Exit 0, `0 Befund(e)` | Glob je Regel zu `**/<Glob>` (7 Läufe), an einer Regel je Block zu `*/*/<Glob>` (2 Läufe) | andere |
| 10 | `docs-check`: ein Defekt je lesendem Modul (`links`, `anchors`, `ids`, `matrix`, `codepaths`, `spans`) unter `.wt/w/` bleibt ungemeldet | Fall 214: Exit 0, `0 Befund(e)` | `".wt/**"` aus `scan.ignore` entfernt (sechs Befunde, je Modul einer) | andere |
| 11 | `docs-check`: derselbe Defekt je Modul in `x/` wird gemeldet | Fall 215: Exit 1, `6 Befund(e)`, je Modul sein Code | Muster `**`; je Modul die Defekt-Datei aus dem Aufbau entfernt (6 Läufe) | andere |
| 12 | `docs-check`: die Zellgrenze der ersten Regel von `structure` gilt im Hauptbaum | Fall 216: Exit 1, `2 Befund(e)`, `section-cell-oversized` je Spalte | Block `structure` aus der Konfiguration des Falls weggelassen | andere |
| 13 | die Mengen der Klasse sind vollständig (zehn Module, sieben Regeln) | Aufbau der Fälle 213 bis 215 im Lauf der Stufe `gegenproben` gegen die echte `.d-check.yml` (die Stufe kopiert sie erst nach `go test`) und `go test` des Aufbaus gegen eine Beispiel-Konfiguration: jedes Modul der Zeile `modules:` steht in genau einer der Listen *lesend* (6) und *nicht lesend* (4); die Regeln beider Blöcke sind 7, jede mit mindestens einem Pfad der Prüfmenge; ein Modul außerhalb der Listen, eine Regel ohne Pfad oder in unbekannter Form ist ein Fehler des Aufbaus, kein grüner Fall | ein elftes Modul in `modules:` (etwa `diagrams`) der echten Konfiguration (Fehler des Aufbaus in der Stufe) und der Beispiel-Konfiguration (`go test` rot); eine achte Regel in `file:` ohne Pfad in der Menge der Dateien | andere |
| 14 | Zeile und Eintrag stehen in den Konfigurationen | Fall 1 und 5 (Mutation: Zeile weg) | wie Zeilen 1 und 5 | andere |
| 15 | die übrigen Gates lesen `.wt/` nicht | kein Fänger: akzeptiertes Negativ, `SPEC-212` Nr. 6 e und f (Grund und Folge dort); Beleg des Verifiers: `make gates` im Hauptbaum mit einem Arbeitsbaum unter `.wt/` | — | andere |
| 16 | getrackter Pfad unter `.wt/`, Kopie ohne `.git`, `mutationslauf.sh`, `targets` und `planning`, Randformen des Musters je Modul, feste Menge der Pfade von Fall 213, Arbeitsbaum außerhalb des Repos | kein Fänger: akzeptierte Negative, `SPEC-212` Nr. 6 a, b, c, d, h, i (Grund und Folge dort) | — | andere |

Sechzehn Datenzeilen, keine nebenläufig (*a* = 0, *b* = 16; ein akzeptiertes Negativ zählt als `andere`, `SPEC-128` Nr. 3 e).

Beide Seiten jeder Grenze führt ein Paar: Zeilen 1 und 2, 5 und 6, 10 und 11; die Gegenseite der Zeile 9 führen Fall 167 von `SPEC-102` (`file`, Abschnitt) und Zeile 12 (Zelle); die Grenze „Wurzel gegen tieferer Pfad“ führen Zeilen 3 und 7, am Modul `links` (für die fünf anderen lesenden Module `SPEC-212` Nr. 6 d).

**Fälle der Gegenproben** (Kennungen in der Folge von `SPEC-102`, Namen vergibt der Implementer): Zeile 1 Fall 205 (`a-check/arbeitsbaum-ungemeldet`), 2 Fall 206, 3 Fall 207, 4 Fall 208; Zeile 5 Fall 209 (`docs-check/arbeitsbaum-ungemeldet`), 6 Fall 210, 7 Fall 211, 8 Fall 212, 9 Fall 213, 10 Fall 214 (`docs-check/arbeitsbaum-module-ungemeldet`), 11 Fall 215 (`docs-check/module-im-hauptbaum-gemeldet`), 12 Fall 216 (`docs-check/zellgrenze-im-hauptbaum`); Zeile 13 der Aufbau der Fälle 213 bis 215 und Tests. Fall 213 baut als Schleife über die Regeln der echten Konfiguration, nicht über `gdMenge()` allein; eine Regel ohne Pfad ist ein Fehler des Aufbaus.

**Nachweis der Zeilen 1 bis 8 und 14** (`AGENTS.md` §3.10; Messender: der Implementer, je Mutation ein Lauf von `gegenproben` im Image der Stufe, Container ohne Netz mit `--memory=4g --cpus=4`, die mutierte Konfiguration per `docker cp` in `/src`, unter `flock`; Load Average des Hosts 22 bis 83 in den Stichproben während der Läufe, fremde Läufe zugleich). Die Fälle 205 bis 212 sind deterministisch (keine Nebenläufigkeit, keine Wahrscheinlichkeit): eine Mutation ist rot in 1 von 1 Lauf, eine Wiederholung ändert nichts an der Erwartung. Die Spalte *rot* nennt die Fälle 205 bis 212, die der Lauf rot meldet.

| Zeile | Mutation | rot (nur Fälle 205 bis 212) |
|---|---|---|
| 1 | `exclude` aus `.a-check.yml` entfernt | 205, 208 |
| 1 | Muster `.wt/*` | 205, 208 |
| 2 | Muster `**` | 206, 207 |
| 3 | Muster `**/.wt/**` | 207 |
| 4 | Muster `.wt/*/x/**` (nur die erste Ebene; gemessen: trifft `.wt/w/x/x.go`, nicht `.wt/a/.wt/b/x/x.go`) | 208 |
| 5 | `".wt/**"` aus `scan.ignore` entfernt | 209, 212 |
| 6 | Muster `**` in `scan.ignore` | 210, 211 |
| 7 | Muster `**/.wt/**` in `scan.ignore` | 211 |
| 8 | Muster `.wt/*/*.md` (nur die erste Ebene; `.wt/*` verhielte sich wie `.wt/**`, `SPEC-212` Nr. 3 b) | 212 |
| 14 | Zeile bzw. Eintrag weg | wie Zeile 1 und 5 |

Bei Mutation 2 färben sich zusätzlich die übrigen Fälle von `a-check` (Exit 0 statt Exit 1). In allen Läufen sind genau die Fälle der Tabelle rot und keiner von 205 bis 212 sonst; der Grund steht in der Ausgabe des Falls (bei `a-check` die Meldung `wrong-direction` einer Datei unter `.wt/` bzw. kein Exit 1 daheim, bei `docs-check` der Befund `target-missing` unter `.wt/` bzw. fehlend im Hauptbaum).

**Nachweis der Zeilen 9 bis 13** (`AGENTS.md` §3.10, §3.17; Messender: der Implementer; Container `gp-fix` aus dem Image der Stufe `gegenproben` mit dem Stand der Gegenproben per `docker cp` in `/gp`, die mutierte `.d-check.yml` per `docker cp` in `/src`, `--network none --memory=4g --memory-swap=4g --cpus=4`, unter `flock`; Load Average (1 Minute) 7,2 bis 11,0 zu Beginn der Läufe). Die Fälle sind deterministisch (ein `d-check`-Prozess je Fall, keine Nebenläufigkeit): jede Mutation ist rot in 1 von 1 Lauf. Die Läufe der Konfigurations-Mutationen fahren die acht Fälle 209 bis 216 über `fahre(standard(), arbeitsbaumFaellePfad("/src/.d-check.yml"), ".")` in einem Test, der nicht im Repo liegt (Prüfung des Falls wie im Lauf der Stufe); der Lauf ohne Mutation meldet 8 Fälle, 0 rot. Die Menge der Zeile 9 sind die Regeln von `structure` und `file` der echten Konfiguration, 7 Regeln, 15 Paare Regel/Datei, 14 Dateien (`harness/README.md`, `AGENTS.md`, 3 Dateien `.claude/commands/`, 6 `.claude/agents/`, 2 `.harness/skills/`, `harness/conventions.md`); die Menge der Zeile 10 und 11 sind die sechs lesenden Module.

| Zeile | Mutation | roter Fall, Grund |
|---|---|---|
| 9 | `files: "harness/README.md"` (`structure`) zu `**/harness/README.md` | 213: 2 Befunde `section-cell-oversized` unter `.wt/w/harness/README.md` |
| 9 | `files: "AGENTS.md"` (`structure`) zu `**/AGENTS.md` | 213: `section-lines-exceeded` unter `.wt/w/AGENTS.md` |
| 9 | `files: "AGENTS.md"` (`file`) zu `**/AGENTS.md` | 213: `file-bytes-exceeded` und `file-lines-exceeded` unter `.wt/w/AGENTS.md` |
| 9 | `.claude/commands/*.md` zu `**/.claude/commands/*.md` (3 Dateien) | 213: 6 Befunde (je Datei beide Codes) unter `.wt/w/.claude/commands/` |
| 9 | `.claude/agents/*.md` zu `**/…` (6 Dateien) | 213: 12 Befunde unter `.wt/w/.claude/agents/` |
| 9 | `.harness/skills/*.md` zu `**/…` (2 Dateien) | 213: 4 Befunde unter `.wt/w/.harness/skills/` |
| 9 | `harness/conventions.md` zu `**/harness/conventions.md` (ein Element der Menge) | 213: 2 Befunde unter `.wt/w/harness/conventions.md` |
| 9 | `*/*/<Glob>` an `harness/README.md`, an `AGENTS.md` (beide Blöcke) und an `.claude/commands/*.md` (4 Läufe) | 213: Befunde unter `.wt/w/…` (wie oben je Regel); 216 zusätzlich rot, weil die Regel die Datei an der Wurzel nicht mehr trifft |
| 10 | `".wt/**"` aus `scan.ignore` entfernt | 214: 6 Befunde unter `.wt/w/`, je Modul einer (`target-missing`, `anchor-missing`, `id-unlinked`, `matrix-inactive`, `codepath-missing`, `span-unclosed`); dazu 209 und 212; 213 bleibt grün (`file` und `structure` lesen `scan.ignore` nicht) |
| 11 | Muster `**` in `scan.ignore` | 215: 0 Befunde statt 6 (0 Dateien geprüft); dazu 210 und 211 |
| 11 | je Modul (`links`, `anchors`, `ids`, `matrix`, `codepaths`, `spans`) die Defekt-Datei aus dem Aufbau entfernt (6 Läufe) | 215 in 6 von 6 Läufen: 5 statt 6 Befunde |
| 12 | `structure` aus den Blöcken der Konfiguration von Fall 216 weggelassen (`gdEchtKonfig`) | 216: Exit 0 statt 1, kein Befund `section-cell-oversized` |
| 13 | ein elftes Modul (`diagrams`) in `modules:` der echten Konfiguration | 213, 214 und 215: Fehler des Aufbaus „steht in 0 der zwei Listen“ |
| 13 | eine achte Regel in `file:` ohne Pfad der Menge (`.claude/hooks/*.md`) | 213: Fehler des Aufbaus „trifft keinen Pfad der Menge“ |
| 13 | ein elftes Modul in der Beispiel-Konfiguration der Tests (`go test`) | 5 Tests rot (`TestArbeitsbaumFaelleBauenDieDateienAusDerEchtenKonfiguration`, `TestFall213LegtJedesPaarUnterWtAn`, `TestModuleFaelleLegenJedenDefektAn`, `TestKlassifikationDerModule`, `TestArbeitsbaumFaelleLaufenGegenDCheck`) |

Lauf der Stufe `gegenproben` am Stand der Zeilen 9 bis 13 (`docker build` mit den Build-Args von `make gegenproben`, ohne `make`, damit der Stempel unberührt bleibt; zwei Läufe): 216 Fälle, alle zwölf Fälle 205 bis 216 und die übrigen `a-check/*`, `docs-check/*` und `guide/*` ok; rot sind die neun Fälle `lint/*` mit „Zeit überschritten“ (Randform 6 von `SPEC-102`, Frist von 120 s bei Last, auch bei Load 40 bis 55 gesehen, nicht Gegenstand des Slice) und im zweiten Lauf zusätzlich `build/ohne-trimpath` („Bau des Binary scheitert“, Ursache nicht untersucht). Die Last stieg in beiden Läufen von 16 und 20 zu Beginn auf 175 bis 222 am Ende (der Lauf selbst und fremde Läufe); ein Lauf ohne Spitze war auf dem Host nicht zu bekommen.

Zeile 13 trägt der Aufbau der Fälle 213 bis 215 im Lauf der Stufe `gegenproben` gegen die echte Konfiguration und `go test` des Aufbaus gegen eine Beispiel-Konfiguration (`TestKlassifikationDerModule`, `TestAbPaareFaengtDieMengenUndFormen`); die Stufe kopiert die echte `.d-check.yml` erst nach `go test`, ein Test liest sie dort nicht. Eine Regel, deren Glob mit `**/` oder `*/` beginnt, wählt die Pfade der Menge nach ihrem Rest (`TestAbPaareLoestDenGlobVonDerWurzel`), damit die Mutation der Zeile 9 ein Befund des Falls ist und kein Fehler des Aufbaus. Akzeptiertes Negativ: die feste Menge der 14 Pfade (Grund und Folge: `SPEC-212` Nr. 6 h).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln: ein Besitzer zugleich). Die Reihenfolge zu den Kettenslices `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu` und `slice-gegenproben-zeitbudget` legt der User beim Start fest; der Slice ist klein und kann vor ihnen laufen, weil er den Parallelbetrieb selbst entlastet.

**Größenmarke** (Summe aus Einfügungen und Löschungen von `git diff --numstat <basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nie, `SPEC-128` Nr. 4 a): vom Planner nach `SPEC-128` Nr. 1 neu gesetzt, **nach** dem ersten Code-Commit (`0ba46599`): die Marke 1600 (*b* = 12, *p* = 185) hat der Planner vor dem Code gesetzt; die Fix-Runde des Architect hat danach die Tabelle von 12 auf 16 Zeilen und die Erwartung an den Produktcode um die Fänger der Zeilen 9 bis 13 erweitert, und die Marke folgt dem nach `AGENTS.md` §3.14 mit Eintrag `ersetzt` und Grund. Die Fänger der Zeilen 9 bis 13 (Fall 213 als Schleife, Fälle 214 bis 216, Test der Mengen) sind beim Setzen nicht gebaut; die Marke ist keine Rückführung und keine Messung des Codes, sie rechnet die gewachsene Tabelle. *a* = 0, *b* = 16 (sechzehn Datenzeilen, alle `andere`), *f* = 0, *p* = 375 (Mitte der neuen Erwartung des Architect: 150 bis 220 für den Stand vor der Fix-Runde plus 150 bis 230 für die Fänger der Zeilen 9 bis 13, zusammen 300 bis 450), also *S* = 80·16 + 375 = 1655 und *M* = 2300; 135·1655 = 223425 liegt unter 300000, keine Schnittfrage. Die Messung am Handoff (342 von 1600) steht in §6. Die Pfade nennen zusätzlich `Dockerfile`, `.dockerignore`, `.golangci.yml`, `harness/mk/` und `tools/harness/`, die der Plan ausschließt (§3, keine Änderung): eine Änderung dort zählt mit und ist eine Abweichung vom Plan.

```groessenmarke
pfad: .a-check.yml
pfad: .d-check.yml
pfad: tools/gegenproben/
pfad: harness/README.md
pfad: Dockerfile
pfad: .dockerignore
pfad: .golangci.yml
pfad: harness/mk/
pfad: tools/harness/
produktcode: 375
fakes: 0
basis: 9520072414786e3746593a1d70600968c237e63a
schaetzung: 1655
marke: 2300
ersetzt: 1600 | Tabelle von 12 auf 16 Zeilen gewachsen (b = 16) und die Erwartung an den Produktcode um die Fänger der Zeilen 9 bis 13 (p = 375 statt 185); die ältere Marke rechnete die kleinere Tabelle, gesetzt nach dem ersten Code-Commit
```

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): ein Slice der Gate-Kette hält Nr. 7, oder der User hält die Ausnahme doch für eine Senkung nach `AGENTS.md` §3.6 und verlangt eine ADR (`SPEC-212` Nr. 2, Bedingung); dann gilt der Weg des Konflikts, eine Folge-Entscheidung, und der Plan folgt ihr.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum mit einem Worktree unter `.wt/`, Closure-Notiz in §7 mit Lerneintrag (Regel mit Anker oder benannte Spec-Lücke).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die Messung zeigt, dass nur `make a-check` den Pfad liest, und die Abhilfe ist eine Zeile — **Ausgang:** eingetreten in der Form, dass zwei Gates ihn lesen (`make a-check`, `make docs-check`) und je eine Zeile genügt; die Marke folgt der Tabelle in §3.
- Größe: Marke 2300 aus der Tabelle (sechzehn Zeilen, alle `andere`) und *p* = 375 nach `SPEC-128` Nr. 1, keine Schnittfrage; die Marke ist nach dem ersten Code-Commit neu gesetzt (ersetzt 1600, Grund und Hergang in §4) — **Ausgang:** entfallen: die Messung am Ende liegt unter der Marke (1057 von 2300, §7); Messung am Handoff nach `SPEC-128` Nr. 4, vor den Fängern der Zeilen 9 bis 13: 342 (331 Einfügungen, 11 Löschungen von `git diff --numstat 0c354d5a..<letzter Commit>` über die Pfade des Blocks, Tests eingeschlossen; davon 260 in den neuen Dateien `arbeitsbaeume.go` und `arbeitsbaeume_test.go`), unter der Marke 2300; nach den Fängern der Zeilen 9 bis 13: 1057 (1046 Einfügungen, 11 Löschungen von `git diff --numstat 9520072414786e3746593a1d70600968c237e63a..<letzter Commit>` über die Pfade des Blocks, Tests eingeschlossen; davon 975 in `arbeitsbaeume.go` und `arbeitsbaeume_test.go`), unter der Marke 2300; `Dockerfile`, `.dockerignore`, `.golangci.yml`, `harness/mk/` und `tools/harness/` unverändert; Messung am Ende (`SPEC-128` Nr. 9) in §7.
- Das Zeitbudget von `make gegenproben` (`SPEC-102`) ist mit den Fällen 205 bis 216 nicht gemessen (Last des Hosts 22 bis 83 im Lauf; die neun Fälle brauchen allein unter 1 s), und der Lauf ohne Mutation färbte unter Last neun Fälle `lint/*` mit „Zeit überschritten“ (`SPEC-102` Randform 6); dazu `build/ohne-trimpath` im zweiten Lauf der Stufe unter Last 175 bis 222 rot (Ursache nicht untersucht; bei Last 10 in Runde 2 des Reviews zweimal grün) — **Ausgang:** weiter offen; Adresse: [`slice-gegenproben-zeitbudget`](../open/slice-gegenproben-zeitbudget.md) §1 (Beleg der neun Fälle `lint/*`) und §6 (Folgepflicht Eichung, Adresse Planner), eingetragen mit der Kennung `slice-gates-ohne-worktrees` im Commit dieser Closure. Der Nehmer schließt die Eichung auf ruhigem Host in §1 aus (Folgepflicht, kein Liefer-Punkt); die Messung `time make gegenproben` mit den Fällen 205 bis 216 und die Ursache von `build/ohne-trimpath` gehören dazu und stehen deshalb als Folgepflicht in seinem §6, nicht in seiner DoD; die neun Fälle `lint/*` unter Last deckt sein Ziel (ein Lauf über der Lastgrenze ist ungültig) als Beleg, in seinem §1 genannt. Die Messung darf kein Lauf im Slice-Zweig fahren (der Lauf schriebe den Stempel, Übergangsregel Nr. 2).
- Ein Gate liest `.wt/` und die Ausnahme verdeckt einen echten Befund an einem Pfad gleichen Namens im Hauptbaum — **Ausgang:** entfallen durch die Entscheidung: das Muster ist an der Wurzel verankert (`.wt/**`), nicht `**/.wt/**`; ein Verzeichnis `.wt` im Quellbaum bleibt geprüft (`SPEC-212` Nr. 3, Nr. 4 Zeile 3; Fänger Zeilen 3 und 7). Offen bleibt ein getrackter Pfad unter `.wt/` (akzeptiertes Negativ mit Grund und Folge, `SPEC-212` Nr. 6 a: die Bedingung hängt am lokalen Ausschluss und am Verzicht auf `git add -f`).
- Raster der Randformen — **Ausgang:** entfallen: unten gefüllt, jede Zelle nennt eine Zeile von `SPEC-212` Nr. 4.
- Der Slice berührt Konfiguration der Gate-Kette und kollidiert mit einem anderen Kettenslice — **Ausgang:** entfallen: Nr. 7 hielt, zum Start und bis zur Closure lag kein anderer Slice der Kette in `in-progress/`.
- Die Gates `make lint`, `make build`, `make test`, `make test-race`, `make handbuch-check`, `make katalog-check`, `make migration-check`, `make schema-check` und `make gegenproben` sind im Hauptbaum nicht gefahren (Last des Hosts über 20 bis 170); ihr Mechanismus ist gelesen und für den Docker-Kontext gemessen — **Ausgang:** weiter offen; der Verifier fuhr `make gates` im Zweig mit einem Probe-Arbeitsbaum unter `.wt/` (Exit 0, Last 11 bis 25); offen bleibt der Hauptbaum mit einem echten Arbeitsbaum (`SPEC-212` Nr. 6 f), erst nach dem Merge erfüllbar; Adresse: Integrationsschritt (Übergangsregeln Nr. 5, `make gates` im Hauptbaum unter der Sperre; DoD 2 trägt den Punkt).
- Die Übergangsregeln Nr. 3 nennen als Pfad der Arbeitsbäume `/Development/odata-recorder-wt-<kurzname>`, der Betrieb legt sie unter `.wt/` an, und der Ausschluss `/.wt/` steht nur in `.git/info/exclude` des Hauptbaums, nicht im Repo; eine versionierte `.gitignore` mit `/.wt/` trüge ihn in jeden Klon und schlösse dort `git add` ohne `-f`, änderte aber, was `git ls-files --exclude-standard` den Gates liefert — **Ausgang:** weiter offen; Adresse: User (zwei Entscheidungen: ob `/.wt/` als versionierte `.gitignore` ins Repo kommt, und ob die Übergangsregeln Nr. 3 den Pfad `.wt/` statt `/Development/odata-recorder-wt-<kurzname>` nennen; der Planner ändert Nr. 3 nicht vor der Entscheidung; `SPEC-212` Nr. 6 a und b tragen die Folge; die Roadmap führt den Punkt als beim User offen).
- Verdacht, dass die ADR doch nötig ist — **Ausgang:** entfallen: Entscheidung ohne ADR (`SPEC-212` Nr. 2); es wurde keine Nummer aus 0030 bis 0039 vergeben und kein Index-Eintrag geschrieben.
- Die Roadmap führt diesen Slice weiter als „angelegt in `open/`“ und „die Reihenfolge legt der User beim Start fest“, und der Slice-Plan `done/slice-gate-integration-determinism.md` nennt in §6 „(Linie C, `open/`)“ — **Ausgang:** entfallen: die Roadmap trägt in der Closure den Zustand nach dem Move; die Datei in `done/` bleibt unberührt, ihre Angabe nennt den Ort zur Zeit ihrer Closure.
- Die Tabelle von §3 hat sechzehn statt zwölf Zeilen (alle `andere`, *a* = 0, *b* = 16), und die Erwartung an den Produktcode wuchs um die Fänger der Zeilen 9 bis 13; der Block in §4 trug *b* = 12 und *p* = 185 — **Ausgang:** eingetreten und nachgezogen: der Block trägt *b* = 16, *p* = 375, *S* = 1655, *M* = 2300 und `ersetzt: 1600`. Der Planner setzte die Marke nach dem ersten Code-Commit (`0ba46599`), weil die Fix-Runde des Architect erst danach kam; `SPEC-128` Nr. 5 c und Nr. 10 e lassen den Stand zu (das Programm sieht den Verlauf nicht, der Diff des Plans zeigt ihn dem Reviewer), `AGENTS.md` §3.14 verlangt die Marke vor dem Code und lässt eine ältere mit Grund ersetzen. Die neue Marke ist keine Antwort auf eine Messung (342 am Handoff), sondern die Formel auf die gewachsene Tabelle; der Reviewer prüft, dass sie nicht nach der Messung gewählt ist.
- Die Ersetzung der Marke 1600 durch 2300 nach dem ersten Code-Commit stand neben `AGENTS.md` §3.14 (Marke vor dem Code); die Messung 1057 liegt unter beiden Marken, die Ersetzung war für den Lauf nicht nötig — **Ausgang:** eingetreten: der Eintrag `ersetzt` bleibt (Einschätzung in §7); `BEO-REPO/marke-nach-dem-ersten-code-commit-ersetzt` führt die Beobachtung.
- Pflege an `tools/gegenproben/arbeitsbaeume.go`: Der Kommentar von `abModuleDateien` sagt, ein Modul ohne Defekt mache den Aufbau ungültig; die Schleife überspringt es still, und erst Fall 215 meldet es (Review 2 G-7, Verifikation V-8). Dazu trägt der Parser der Blöcke `structure` und `file` den Hauptteil der 975 von 1057 gemessenen Zeilen in `arbeitsbaeume.go` und `arbeitsbaeume_test.go` (G-8) — **Ausgang:** weiter offen; Adresse: User (Entscheidung, ob der Kommentar in einem eigenen Slice oder bei der nächsten Änderung an `tools/gegenproben/` berichtigt wird und ob der Parser bleibt); kein Slice in `open/` nimmt es an, `slice-gegenproben-zeitbudget` schließt Änderungen am Format der Messreihen aus und führt den Punkt nicht. Der Planner ändert den Code der Closure nicht.
- Die Fänger der Zeilen 9 bis 13 (Schleife über sieben Regeln in Fall 213, Fälle 214 bis 216, Prüfung der Mengen) — **Ausgang:** entfallen: gebaut; Fall 213 als Schleife über 7 Regeln und 15 Paare, die Fälle 214 bis 216, der Aufbau mit Prüfung der Mengen; Belege je Zeile in §3, Nachweis der Zeilen 9 bis 13. Reviewer und Verifier lösen sie gegen die Klasse auf (`AGENTS.md` §3.17), nicht gegen die genannten Mutationen.
- Der Fänger der Zeile 13 ist der Aufbau der Fälle 213 bis 215 im Lauf der Stufe `gegenproben` und `go test` des Aufbaus gegen eine Beispiel-Konfiguration, kein Test, der die echte `.d-check.yml` liest (die Stufe kopiert die echte Datei erst nach `go test`, `Dockerfile` ist nicht Gegenstand des Slice) — **Ausgang:** entfallen: der Reviewer hat die Zeile in [Runde 2](../../../reviews/2026-11-18-slice-gates-ohne-worktrees-review-2.md) mit der Mutation „elftes Modul“ aufgelöst und den Wortlaut an den Architect zurückgegeben; `SPEC-212` Nr. 5 Zeile 13 und *Stand* nennen jetzt den Träger, §3 Zeile 13 gleichlautend.

**Raster der Randformen** (`AGENTS.md` §3.16; Operation × Pfadform; jede Zelle nennt die Zeile von `SPEC-212` Nr. 4, die Fänger stehen in der Tabelle von §3):

| Operation | Arbeitsbaum, tiefer verschachtelt | `.wt` im Quellbaum | Symlink | `.wt/` fehlt | Verletzung im Hauptbaum neben `.wt/` | Arbeitsbaum außerhalb | getrackt, Kopie ohne `.git` |
|---|---|---|---|---|---|---|---|
| `make a-check` | Nr. 4 Z. 2 | Z. 3 | Z. 4, 5 | Z. 6 | Z. 8 | Z. 9 (Nr. 6 i) | Z. 11, 12 |
| `make docs-check` | Z. 2 | Z. 3 | Z. 4, 5 | Z. 6 | Z. 8 | Z. 9 (Nr. 6 i) | Z. 11, 12; Z. 10 für `file` und `structure` |
| `make zeichen-check`, Gate-Nachweis, Docker-Gates, `make gegenproben` | gilt nicht: lesen `.wt/` nicht (`SPEC-212` Nr. 1) | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | Nr. 6 b (`zeichen-check` liest die Kopie ohne `.git`) |

Die Zeile `make docs-check` gilt für jedes der zehn Module von `d-check`: Nr. 4 Zeile 13 für die sechs lesenden (Fänger §3 Zeilen 10 und 11), Zeile 10 für `file` und `structure` (Fänger §3 Zeilen 9 und 12), Zeile 14 für `targets` und `planning` (kein Fall, `SPEC-212` Nr. 6 d). Die Pfadformen tiefer verschachtelt und `.wt` im Quellbaum führt allein das Modul `links` (Fälle 211 und 212); für die anderen fünf lesenden Module gilt `SPEC-212` Nr. 6 d.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

- **Was hat funktioniert:** `make a-check` und `make docs-check` lesen `.wt/` nicht mehr (`exclude: [".wt/**"]` in `.a-check.yml`, `".wt/**"` in `scan.ignore` von `.d-check.yml`, Vertrag `SPEC-212`, keine ADR). Belege des Verifiers: am Probe-Arbeitsbaum 0 Befunde von `make a-check` mit der Ausnahme, 1341 `wrong-direction` ohne sie; `make docs-check` meldet einen Defekt unter `.wt/` nicht und denselben Defekt in `x/` genau einmal; `make gates` im Zweig mit dem Probe-Arbeitsbaum Exit 0. Die Fänger (Fälle 205 bis 216 der Gegenproben) laufen als Schleife über die Klassen der Zusage: Fall 213 über die 7 Regeln und 15 Paare der Blöcke `structure` und `file`, die Fälle 214 und 215 über die sechs lesenden Module; die Menge der zehn Module und sieben Regeln prüft der Aufbau selbst und bricht bei einem elften Modul oder einer Regel ohne Pfad ab (Zeile 13). Die Ausnahme ist keine Senkung nach `AGENTS.md` §3.6, weil die Prüfmenge eines Standes gleich bleibt (`git ls-files .wt` ist leer, `SPEC-212` Nr. 2), und der Reviewer hat die Entscheidung gegen den Diff selbst gestellt, ohne dem Architect zu folgen.
- **Messung (`SPEC-128` Nr. 4 und 9):** 1057 gegen Marke 2300 (1046 Einfügungen, 11 Löschungen von `git diff --numstat 9520072414786e3746593a1d70600968c237e63a..HEAD` über die Pfade des Blocks, Tests eingeschlossen, am Closure-Stand gerechnet; 975 davon in `arbeitsbaeume.go` und `arbeitsbaeume_test.go`), Auslastung 0,46, unter der Marke. Der Handoff maß vor den Fängern der Zeilen 9 bis 13 342; der Zuwachs auf das 3,1-Fache folgt dem Wachstum der Tabelle von 12 auf 16 Zeilen, nicht dem Code der ersten Fassung.
- **Marke, Einschätzung:** Die erste Marke 1600 (*b* = 12, *p* = 185) stand vor dem ersten Code-Commit. Die Fix-Runde des Architect hat nach Review 1 die Tabelle auf 16 Zeilen und die Erwartung auf *p* = 375 gehoben, und der Planner hat die Marke danach nach dem Code auf 2300 ersetzt. Die Rechnung stimmt (Verifikation V-7 nachgerechnet), und die Messung lag unter beiden Marken, die Ersetzung war für den Lauf nicht nötig. Sie steht neben `AGENTS.md` §3.14 (Marke vor dem Code): Wer nur das Ergebnis liest, kann sie nicht von einer nach der Messung gewählten unterscheiden, und das Programm sieht den Verlauf nicht (`SPEC-128` Nr. 5 c, Nr. 10 e). Der Eintrag `ersetzt` bleibt: Die Tabelle hat 16 Zeilen, die Marke 1600 rechnete 12, und ein Zurück auf 1600 wäre die falsche Rechnung bei grünem Ergebnis. Eine Rückführung gibt es nicht. Die Lücke ist die des Ablaufs: Eine Architect-Runde nach dem Start, die den Plan vergrößert, hat keinen Schritt, der die Marke vor dem nächsten Code-Commit neu setzt; die Beobachtung steht im Register.
- **Was ging anders als geplant:** Zwei Review-Runden und eine Verifikation statt einer Runde. Review 1 (0 HIGH, 3 MEDIUM, 2 LOW, 3 INFO) fand, dass der erste Fänger zu Zeile 9 eine Auswahl führte (13 Pfade, die erste Regel von `structure` fehlte, F-1) und dass von zehn Modulen von `d-check` nur eines gemessen war (F-3). Beide Fixes hat Review 2 gegen die Klasse aufgelöst (je Element der Menge eine Mutation, auch an Elementen, die der Befund nicht nannte): 0 HIGH, 0 MEDIUM, 5 LOW, 4 INFO. Die Verifikation (0 HIGH, 1 MEDIUM, 4 LOW, 3 INFO; bestanden mit Auflagen) fand ein akzeptiertes Negativ ohne Grund (V-1) und zwei Aussagen über denselben Fänger in Vertrag und Plan (V-2); der Architect hat sie in `e62885fe` nachgezogen (`SPEC-212` Nr. 6 h und i, Fänger der Zeile 13 am Träger, Plan §3 Nachweis).
- **Befund-Bilanz, Ausgang je Befund:**
  - Review 1 (`2026-11-18-slice-gates-ohne-worktrees-review`): F-1 MEDIUM behoben (Fall 213 als Schleife über die Regeln, Review 2 gegen die Klasse aufgelöst); F-2 MEDIUM behoben (getrackter Pfad unter `.wt/` ist ein akzeptiertes Negativ mit Grund und Folge, `SPEC-212` Nr. 6 a und b); F-3 MEDIUM behoben (sechs lesende und vier nicht lesende Module, Fälle 214 bis 216, `SPEC-212` Nr. 3 b und 6 d); F-4 LOW behoben (Grenze in `SPEC-102`); F-5 LOW behoben in dieser Closure (Zustandsfelder der Roadmap); F-6 INFO Pin-Semantik bestätigt, kein Eingriff; F-7 INFO Stempel gehört dem Integrationsschritt (Übergangsregel Nr. 2), Zeitbudget beim Nehmer (§6); F-8 INFO Messung nachgerechnet.
  - Review 2 (`…-review-2`): G-1 LOW Marke nach dem Code ersetzt, eingetreten (siehe Einschätzung); G-2 LOW behoben (Zeile 13 nennt ihren Träger in `SPEC-212`, Plan §3 und §6); G-3 LOW behoben (`SPEC-212` Nr. 6 h mit Grund und Folge); G-4 LOW behoben (Plan §3 Nachweis auf dem Stand der Fix-Runde); G-5 LOW behoben in dieser Closure (wie F-5); G-6 INFO Wortlaut von Nr. 6 a gilt für den Inhalt eines Arbeitsbaums, kein Eingriff; G-7 INFO und G-8 INFO weiter offen mit Adresse User (§6, Pflege an `arbeitsbaeume.go`); G-9 INFO `build/ohne-trimpath` an den Nehmer `slice-gegenproben-zeitbudget` (§6 dort).
  - Verifikation (`…-verifikation`, 0 HIGH, 1 MEDIUM, 4 LOW, 3 INFO): V-1 MEDIUM behoben (Nr. 6 h); V-2 LOW behoben (Zeile 13); V-3 LOW behoben (Nehmer trägt die Sendung, §6 oben); V-4 LOW behoben in dieser Closure (Roadmap); V-5 LOW behoben (Plan §3 Nachweis); V-6 INFO behoben (Nr. 6 i); V-7 INFO wie G-1; V-8 INFO wie G-7 und G-8, weiter offen mit Adresse User.
- **Steering-Loop-Eintrag:** Fünf Lehren, gezählt, nicht verkörpert (kein Zielort, `liegt in` entfällt): (1) Eine Gate-Ausnahme für Fremdkopien ist keine Senkung nach `AGENTS.md` §3.6, solange die Prüfmenge eines Commits gleich bleibt; die Bedingung (`git ls-files .wt` leer) hängt an einem lokalen Ausschluss, den kein Gate prüft, und gehört als akzeptiertes Negativ mit Grund und Folge in den Vertrag (`SPEC-212` Nr. 6 a), nicht als Satz „nie getrackt“. (2) Der Fänger einer Zusage über eine Konfiguration läuft über die Menge der Konfiguration (jedes Modul, jede Regel), nicht über die Auswahl, die dem Autor einfiel; erst Review 1 zählte die Module (zehn, nicht fünf), und der Aufbau prüft die Menge jetzt selbst (Zeile 13). Das ist dieselbe Klasse wie `AGENTS.md` §3.15 und §3.17, hier vor dem ersten Fix; Auslöser `BEO-REPO/auswahl-statt-menge-der-konfiguration` (1×, neu). (3) Eine Marke, die nach dem Start der Architect-Arbeit neu zu setzen ist, hat keinen Schritt vor dem nächsten Code-Commit; `BEO-REPO/marke-nach-dem-ersten-code-commit-ersetzt` (1×, neu). (4) Eine Zeile der Tabelle, deren Fänger „ein Test“ nennt, während der Träger der Aufbau der Fälle in der Stufe ist, sagt dem Leser einen Lauf zu, den `go test` nicht fährt (G-2, V-2); der Träger gehört in die Zeile (gezählt unter `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`). (5) Nach einem Rot unter Last (neun Fälle `lint/*`, `build/ohne-trimpath` bei Last 175 bis 222, grün bei Last 10) gilt: auf ruhigem Host wiederholen, bevor ein Befund entsteht (`BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last`, 2×). Ein Diff-Vorschlag an den User für `AGENTS.md` und `.harness/skills/reviewer.md` liegt im Bericht der Closure, nicht im Repo. Auslöser: `BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt` (`slice-gate-integration-determinism`, 1×; mit diesem Slice behoben).
- **Beobachtungs-Register (`../observations/`):** `BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt/`: Stand auf verkörpert (Ausnahme in `.a-check.yml` und `.d-check.yml`, `SPEC-212`). `evidence/slice-gates-ohne-worktrees.md` ergänzt in `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft/` (Zähler 61×; Stand dort verkörpert), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle/` (28×; verkörpert), `BEO-REPO/akzeptiertes-negativ-ohne-folge/` (16×; verkörpert), `BEO-REPO/ein-negativ-benennt-die-ungefangene-klasse-zu-eng/` (2×, offen), `BEO-REPO/roadmap-links-brechen-im-zweig-nach-dem-start-move/` (2×, offen), `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last/` (2×, offen). Neu angelegt, je 1× und offen: `BEO-REPO/auswahl-statt-menge-der-konfiguration/`, `BEO-REPO/marke-nach-dem-ersten-code-commit-ersetzt/`. Nicht gezählt: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (die Fälle 205 bis 216 fahren kein Produkt und brauchen 0,4 bis 0,84 s; die roten Fälle `lint/*` sind Last, kein Zuschnitt, und stehen als Beleg beim Nehmer des Zeitbudgets) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (Review 2 fand an den Fixes keine ungefangene Achse; F-1 und F-3 betrafen den ersten Fänger, nicht einen Fix). Der Schwellen-Übertritt 3× ist für keine neue Beobachtung erreicht.
- **Folge-Slices:** keiner. Die Sendung zum Zeitbudget steht in `slice-gegenproben-zeitbudget` (§1, §6), die Pflege an `arbeitsbaeume.go` und die Entscheidungen zu `.gitignore` und Pfad der Arbeitsbäume beim User (§6).
- **Risiken aus §6:** Jede Zeile von §6 trägt genau einen Ausgang: eingetreten (Zahl der Gates, Tabelle, Ersetzung der Marke), entfallen (Messung, Verdeckung, Raster, ADR, Kollision, Roadmap, Fänger der Zeilen 9 bis 13 und 13) oder weiter offen. Weiter offen: Zeitbudget und `build/ohne-trimpath` (Nehmer `slice-gegenproben-zeitbudget`), `make gates` im Hauptbaum (Integrationsschritt), Pfad der Arbeitsbäume und `.gitignore` (User), Pflege an `arbeitsbaeume.go` (User).
- **Validator:** n/a. Das Werkzeug ist eine Gate-Konfiguration ohne externen Bedarfsbeleg; der Verifier hat den Zweck an einem Probe-Arbeitsbaum belegt.
- **Drei Paarungen:** Anker: kein `liegt in` in diesem Eintrag, nichts zu paaren. Folge-Slice: kein Folge-Slice angelegt; die Sendung führt der Nehmer `slice-gegenproben-zeitbudget` in §1 und §6 mit der Kennung `slice-gates-ohne-worktrees`. Register: acht `evidence/slice-gates-ohne-worktrees.md` liegen vor (sechs ergänzt, zwei neu), siehe Register-Zeile.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt` (1×, Ausgang *offen*); `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` betrifft die Fälle 205 bis 216 nicht (kein Produkt, §7).

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

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an den Zuschnitt der Gates, der Bezug ist der nächste, den es gibt), [ADR-0002](../../adr/0002-architektur-gate-a-check.md) (Architektur-Gate `a-check`), [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Doku-Gate). Geber: `slice-gate-integration-determinism` (Befund aus dem Parallelbetrieb, Risiko in dessen §6 mit Ausgang *weiter offen*).

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

**Hintergrund (Geber `slice-gate-integration-determinism`):** Der Architect hat gemessen (`SPEC-212` Nr. 1): Nur `make a-check` (1351 Befunde) und `make docs-check` (39 Befunde) lesen `.wt/` im Hauptbaum; alle übrigen Gates tun es nicht (Kontext als Allowlist, `git ls-files`, benannte Pfade). Die Abhilfe ist je Gate ein Muster, `.wt/**` an der Wurzel; sie braucht keine ADR, weil keine Schwelle und keine Regel sinkt und `.wt/` nie Teil eines Commits ist (`SPEC-212` Nr. 2).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung der Übergangsregeln für Worktrees (Pfad, Zahl, Sperre) — Sache des Planners und des Users; der Slice ändert nur, was die Gates lesen.
- Eine Lockerung eines Schwellenwerts oder einer Architekturregel — `AGENTS.md` §3.6; ausgenommen wird ein Pfad, keine Regel.
- Die Gate-Kette selbst (`make gates`, `GATE_CHECKS`) — `slice-gate-integration-determinism` hat sie festgelegt (`SPEC-211`).
- `tools/harness/mutationslauf.sh` (kopiert `.wt/` in den Container, kein Gate, ändert kein Urteil) — akzeptiertes Negativ, `SPEC-212` Nr. 6 c.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [x] Der Architect hat festgestellt, welche Gates `.wt/` im Hauptbaum lesen (`SPEC-212` Nr. 1: `make a-check` und `make docs-check`; `make zeichen-check` und die übrigen nicht), und entschieden, dass die Abhilfe keine ADR braucht (`SPEC-212` Nr. 2).
- [ ] Jedes betroffene Gate nimmt `.wt/` aus (`.a-check.yml`: `exclude: [".wt/**"]`; `.d-check.yml`: `".wt/**"` in `scan.ignore`, `SPEC-212` Nr. 3), mit einem Fänger je Zeile von `SPEC-212` Nr. 5 (ein Worktree mit einer Verletzung bleibt ungemeldet, dieselbe Verletzung im Hauptbaum wird gemeldet); `make gates` grün, auch im Hauptbaum mit einem Arbeitsbaum unter `.wt/`.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben (Auslöser `BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt`); keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect hat entschieden (`SPEC-212`), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.a-check.yml` | update | Schlüssel `exclude: [".wt/**"]` (`SPEC-212` Nr. 3 a), mit Kommentar der Klasse Abgrenzung |
| `.d-check.yml` | update | Eintrag `".wt/**"` in `scan.ignore` samt Zeile im Kommentar davor (`SPEC-212` Nr. 3 b) |
| `tools/gegenproben/` | update | Fälle der Fänger (`SPEC-212` Nr. 5: vier Fälle `a-check`, vier `d-check`, einer mit der echten `file`-Konfiguration); a-check-Fälle in `faelle.go` neben `aCheckTestsFaelle`, d-check-Fälle in einer neuen Datei neben `guide.go` mit eigener Konfiguration (`scan` der echten `.d-check.yml`, `modules: [links]`); Tests der Fall-Folge |
| `harness/vertraege/SPEC-102-gegenproben.md` | update (Implementer) | die neuen Fälle nach der Folge des Vertrags (nächste Nummer nach 204), Zeitbudget, Zeile *Stand* |
| `harness/vertraege/SPEC-039-…`, `SPEC-040-…` | update (Implementer) | je ein Zeiger auf `SPEC-212` in der Festlegung |
| `harness/README.md` | update (Implementer) | Absatz *Nicht behauptet*: Verdrahtung der Zeilen (`SPEC-212` Nr. 6 g), ungelesene Gates (Nr. 6 f) |
| `.golangci.yml`, `harness/mk/`, `.dockerignore`, `tools/harness/*.sh` | keine Änderung | die Messung zeigt kein weiteres Gate (`SPEC-212` Nr. 1) |

**Erwartete Produktcode-Größe (für die Marke des Planners, `SPEC-128` Nr. 1):** Summe aus Einfügungen und Löschungen über `.a-check.yml`, `.d-check.yml`, `tools/gegenproben/` (Tests eingeschlossen) und `harness/README.md`, geschätzt 150 bis 220 Zeilen: Konfigurationen rund 15 (Zeile und Kommentar je Datei), a-check-Fälle rund 50, d-check-Fälle samt Konfigurationsbau rund 90, Tests der Fall-Folge rund 20, `harness/README.md` rund 2. `harness/vertraege/` zählt die Messung nicht mit (`SPEC-128`). Kleiner Slice, kein Schnitt erwartet.

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (`SPEC-212`, dieser Abschnitt, Raster in §6; erledigt) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock` auf die Sperre der Linien; Zeichen jenseits von ASCII an Bytewerten gebildet und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Quelle `SPEC-212` Nr. 5, die Mutation sieht der Implementer selbst rot und trägt Rate und Messenden hier ein):

| Nr. | Zusage | Fänger (Fall der Gegenproben) | Mutation, die ihn rot macht |
|---|---|---|---|
| 1 | `a-check`: Verletzung unter `.wt/w/…` bleibt ungemeldet | Datei außerhalb jeder Schicht, die ein Paket des Modells importiert, unter `.wt/w/x/x.go`: Exit 0 | `exclude` entfernt; Muster `.wt/*` |
| 2 | `a-check`: dieselbe Datei im Hauptbaum wird gemeldet | `x/x.go`: Exit 1, `wrong-direction` | Muster `**` |
| 3 | `a-check`: ein Verzeichnis `.wt` im Quellbaum bleibt geprüft | `src/hexagon/model/.wt/x.go` importiert einen Dienst: Exit 1, `core-impurity` | Muster `**/.wt/**` |
| 4 | `a-check`: Arbeitsbaum in einem Arbeitsbaum bleibt ungemeldet | `.wt/a/.wt/b/x/x.go`: Exit 0 | Muster, das nur die erste Ebene trifft |
| 5 | `docs-check`: fehlendes Linkziel unter `.wt/w/` bleibt ungemeldet | Exit 0, `0 Befund(e)` | `".wt/**"` aus `scan.ignore` entfernt |
| 6 | `docs-check`: derselbe Verweis in `x/a.md` wird gemeldet | Exit 1, `target-missing` | Muster `**` |
| 7 | `docs-check`: `x/.wt/b.md` bleibt geprüft | Exit 1, `target-missing` | Muster `**/.wt/**` |
| 8 | `docs-check`: `.wt/a/.wt/b/c.md` bleibt ungemeldet | Exit 0 | Muster, das nur die erste Ebene trifft |
| 9 | `docs-check`: `file` und `structure` lesen `.wt/` nicht | echte Blöcke `scan`, `structure`, `file`, `.wt/w/AGENTS.md` mit 50 000 Zeilen: Exit 0 | Glob `AGENTS.md` zu `**/AGENTS.md` |
| 10 | Zeile und Eintrag stehen in den Konfigurationen | Fall 1 und 5 (Mutation: Zeile weg) | wie Zeilen 1 und 5 |
| 11 | die übrigen Gates lesen `.wt/` nicht | kein Fänger: akzeptiertes Negativ, `SPEC-212` Nr. 6 e und f (Grund und Folge dort); Beleg des Verifiers: `make gates` im Hauptbaum mit einem Arbeitsbaum unter `.wt/` | — |
| 12 | getrackter Pfad unter `.wt/`, Kopie ohne `.git`, `mutationslauf.sh` | kein Fänger: akzeptierte Negative, `SPEC-212` Nr. 6 a, b, c | — |

Beide Seiten jeder Grenze führt ein Paar: Zeilen 1 und 2, 5 und 6; die Grenze „Wurzel gegen tieferer Pfad“ führen Zeilen 3 und 7.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln: ein Besitzer zugleich). Die Reihenfolge zu den Kettenslices `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu` und `slice-gegenproben-zeitbudget` legt der User beim Start fest; der Slice ist klein und kann vor ihnen laufen, weil er den Parallelbetrieb selbst entlastet.

**Größenmarke:** Prosa, kein Block, vorläufig. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung des Architect steht in §3 (150 bis 220 Zeilen). Erwartet ist ein kleiner Slice.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): ein Slice der Gate-Kette hält Nr. 7, oder der User hält die Ausnahme doch für eine Senkung nach `AGENTS.md` §3.6 und verlangt eine ADR (`SPEC-212` Nr. 2, Bedingung); dann gilt der Weg des Konflikts, eine Folge-Entscheidung, und der Plan folgt ihr.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum mit einem Worktree unter `.wt/`, Closure-Notiz in §7 mit Lerneintrag (Regel mit Anker oder benannte Spec-Lücke).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die Messung zeigt, dass nur `make a-check` den Pfad liest, und die Abhilfe ist eine Zeile — **Ausgang:** eingetreten in der Form, dass zwei Gates ihn lesen (`make a-check`, `make docs-check`) und je eine Zeile genügt; die Marke folgt der Tabelle in §3.
- Ein Gate liest `.wt/` und die Ausnahme verdeckt einen echten Befund an einem Pfad gleichen Namens im Hauptbaum — **Ausgang:** entfallen durch die Entscheidung: das Muster ist an der Wurzel verankert (`.wt/**`), nicht `**/.wt/**`; ein Verzeichnis `.wt` im Quellbaum bleibt geprüft (`SPEC-212` Nr. 3, Nr. 4 Zeile 3; Fänger Zeilen 3 und 7). Offen bleibt ein getrackter Pfad unter `.wt/` (Negativ, `SPEC-212` Nr. 6 a).
- Raster der Randformen — **Ausgang:** entfallen: unten gefüllt, jede Zelle nennt eine Zeile von `SPEC-212` Nr. 4.
- Der Slice berührt Konfiguration der Gate-Kette und kollidiert mit einem anderen Kettenslice — **Ausgang:** weiter offen: Nr. 7, einer zugleich, Start laut §4.
- Die Gates `make lint`, `make build`, `make test`, `make test-race`, `make handbuch-check`, `make katalog-check`, `make migration-check`, `make schema-check` und `make gegenproben` sind im Hauptbaum nicht gefahren (Last des Hosts über 20 bis 170); ihr Mechanismus ist gelesen und für den Docker-Kontext gemessen — **Ausgang:** weiter offen; Adresse: Verifier (`make gates` im Hauptbaum mit einem Arbeitsbaum unter `.wt/`, `SPEC-212` Nr. 6 f).
- Die Übergangsregeln Nr. 3 nennen als Pfad der Arbeitsbäume `/Development/odata-recorder-wt-<kurzname>`, der Betrieb legt sie unter `.wt/` an, und der Ausschluss `/.wt/` steht nur in `.git/info/exclude` des Hauptbaums, nicht im Repo — **Ausgang:** weiter offen; Adresse: Planner (Übergangsregeln und Pfad sind Sache des Planners, §1; `SPEC-212` Nr. 6 b trägt die Folge).
- Verdacht, dass die ADR doch nötig ist — **Ausgang:** entfallen: Entscheidung ohne ADR (`SPEC-212` Nr. 2); es wurde keine Nummer aus 0030 bis 0039 vergeben und kein Index-Eintrag geschrieben.

**Raster der Randformen** (`AGENTS.md` §3.16; Operation × Pfadform; jede Zelle nennt die Zeile von `SPEC-212` Nr. 4, die Fänger stehen in der Tabelle von §3):

| Operation | Arbeitsbaum, tiefer verschachtelt | `.wt` im Quellbaum | Symlink | `.wt/` fehlt | Verletzung im Hauptbaum neben `.wt/` | Arbeitsbaum außerhalb | getrackt, Kopie ohne `.git` |
|---|---|---|---|---|---|---|---|
| `make a-check` | Nr. 4 Z. 2 | Z. 3 | Z. 4, 5 | Z. 6 | Z. 8 | Z. 9 | Z. 11, 12 |
| `make docs-check` | Z. 2 | Z. 3 | Z. 4, 5 | Z. 6 | Z. 8 | Z. 9 | Z. 11, 12; Z. 10 für `file` |
| `make zeichen-check`, Gate-Nachweis, Docker-Gates, `make gegenproben` | gilt nicht: lesen `.wt/` nicht (`SPEC-212` Nr. 1) | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | Nr. 6 b (`zeichen-check` liest die Kopie ohne `.git`) |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `slice-gate-integration-determinism` (Befund aus dem Parallelbetrieb, 2026-11-18); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt` (1×, Ausgang *offen*).

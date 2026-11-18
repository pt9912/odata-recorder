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

**Berührte Spec-Stellen:** noch nicht bestimmt; der Architect nennt sie (Verträge der Gates unter `harness/vertraege/`, `SPEC-040` für das Doku-Gate, der Vertrag von `make a-check`).
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

**Hintergrund (Geber `slice-gate-integration-determinism`):** Welche Gates `.wt/` mitlesen, ist nur für `make a-check` gemeldet. Ungeprüft sind `make docs-check` (`scan.roots` ist `.`, `scan.ignore` in `.d-check.yml` nennt `.wt/` nicht), `make lint`, `make zeichen-check` und die Gegenproben. Der Architect stellt zuerst fest, welche Gates es tun; die Abhilfe (`.wt/` bei den Gates ausnehmen) ist eine Senkung der Reichweite eines Gates nur dem Anschein nach, weil kein Arbeitsbaum Teil des geprüften Stands ist; ob dafür eine ADR nötig ist (`AGENTS.md` §3.6), entscheidet er mit Grund.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung der Übergangsregeln für Worktrees (Pfad, Zahl, Sperre) — Sache des Planners und des Users; der Slice ändert nur, was die Gates lesen.
- Eine Lockerung eines Schwellenwerts oder einer Architekturregel — `AGENTS.md` §3.6; ausgenommen wird ein Pfad, keine Regel.
- Die Gate-Kette selbst (`make gates`, `GATE_CHECKS`) — `slice-gate-integration-determinism` hat sie festgelegt (`SPEC-211`).

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] Der Architect hat festgestellt, welche Gates `.wt/` im Hauptbaum lesen (`make a-check`, `make docs-check`, `make lint`, `make zeichen-check`, die Gegenproben; Messung je Gate mit einem Worktree unter `.wt/`), und entschieden, ob die Abhilfe eine ADR braucht (Sendung von `slice-gate-integration-determinism`, dessen §6).
- [ ] Jedes betroffene Gate nimmt `.wt/` aus, mit einem Fänger je Gate (ein Worktree mit einer Verletzung bleibt ungemeldet, dieselbe Verletzung im Hauptbaum wird gemeldet); `make gates` grün.
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

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.a-check.yml` | update (nach Entscheidung des Architect) | `.wt/` aus den Schichten oder dem Scan ausnehmen |
| `.d-check.yml` | update (nach Entscheidung des Architect) | `scan.ignore` nennt `.wt/**` |
| `.golangci.yml`, `harness/mk/` | update (nur, wenn die Messung ein Gate trifft) | Ausnahme je Gate |
| `harness/vertraege/` | update (Architect, vor dem Code) | Festlegung je Gate, Tabelle *Zusage · Fänger* |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Aufträge unten, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock` auf die Sperre der Linien; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Aufträge an den Architect:**

1. **Bestand messen.** Je Gate, ob es `.wt/` im Hauptbaum liest; die Zahl der Befunde je Gate mit einem Worktree als Kopie.
2. **Bindung.** Ob die Ausnahme eine ADR braucht (Linie C: ADR-Nummern 0030 bis 0039) oder die Verträge genügen.
3. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): Worktree an anderem Pfad als `.wt/`, ein Worktree, der selbst `.wt/` trägt, ein Arbeitsbaum außerhalb des Repos (kein Befund), ein Gate, das den Pfad nur als Muster kennt.
4. **Tabelle *Zusage · Fänger*** (§3) und **erwartete Produktcode-Größe** für die Marke.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln: ein Besitzer zugleich). Die Reihenfolge zu den Kettenslices `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu` und `slice-gegenproben-zeitbudget` legt der User beim Start fest; der Slice ist klein und kann vor ihnen laufen, weil er den Parallelbetrieb selbst entlastet.

**Größenmarke:** Prosa, kein Block, vorläufig. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect. Erwartet ist ein kleiner Slice (wenige Zeilen Konfiguration je Gate, Fänger in den Gegenproben oder als Handlauf).

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Architect entscheidet, dass die Ausnahme ein Gate lockert und eine ADR braucht, die der User noch nicht beschlossen hat, oder ein Slice der Gate-Kette hält Nr. 7.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum mit einem Worktree unter `.wt/`, Closure-Notiz in §7 mit Lerneintrag (Regel mit Anker oder benannte Spec-Lücke).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die Messung zeigt, dass nur `make a-check` den Pfad liest, und die Abhilfe ist eine Zeile — **Ausgang:** weiter offen: der Slice bleibt klein, die Marke folgt der Tabelle des Architect.
- Ein Gate liest `.wt/` und die Ausnahme verdeckt einen echten Befund an einem Pfad gleichen Namens im Hauptbaum — **Ausgang:** weiter offen; Adresse: Architect (Auftrag 3, Randformen).
- Raster der Randformen (Auftrag 3) ist noch leer — **Ausgang:** weiter offen: der Architect füllt es vor dem ersten Code-Commit (`AGENTS.md` §3.16).
- Der Slice berührt Konfiguration der Gate-Kette und kollidiert mit einem anderen Kettenslice — **Ausgang:** weiter offen: Nr. 7, einer zugleich, Start laut §4.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `slice-gate-integration-determinism` (Befund aus dem Parallelbetrieb, 2026-11-18); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/worktrees-im-arbeitsverzeichnis-vom-gate-mitgescannt` (1×, Ausgang *offen*).

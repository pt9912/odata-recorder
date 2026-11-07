# Slice slice-mutationslauf-modulwurzel: Der Mutationslauf fährt ein Paket an der Wurzel eines Moduls unter tools/

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice. Er ändert `tools/mutationslauf` und den Vertrag `SPEC-210`, **nicht** die Gate-Kette (`Makefile`-Fragmente, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`); die Bedingung Gate-Kette (Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)) greift nur, wenn der Architect doch eine Datei der Kette berührt, dann hält der Slice an und gibt an den Planner zurück.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Werkzeuge der Rollen, der Bezug ist der nächste, den es gibt), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (das Werkzeug liegt außerhalb der Gate-Kette; dieser Slice ändert das nicht).

**Berührte Spec-Stellen:** [SPEC-210](../../../../harness/vertraege/SPEC-210-mutationslauf.md) (Mutationslauf: Pfade eines Pakets im Modul, Nullprobe; der Architect entscheidet die Lesart und die Randformen).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Closure von `slice-handbuch-meldungskatalog`). **Datum:** 2026-11-07.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Übernimmt:** `slice-handbuch-meldungskatalog` (Rückgabe des Reviewers an den Planner, Review 2026-11-07, bestätigt in der Verifikation; Geber trägt den Punkt in §6 als Zeile mit Ausgang *eingetreten* und dieser Kennung).

**Ziel:** `tools/mutationslauf` fährt eine Mutationsliste für ein Paket, das **an der Wurzel seines Moduls** liegt (`tools/katalog`, `tools/zeichen`: das Paket ist das Modul). Heute schneidet `relInModul` bei `Paket == Modulwurzel` nichts ab (`TrimPrefix(p, Dir+"/")` auf `p == Dir`), `nullprobe.go` ruft sie statt `paketArg`, und die Nullprobe läuft in einem Verzeichnis, das nicht existiert (`fork/exec … no such file`); jede solche Liste endet mit `Ausgangsstand rot`. Die Mutationen der Pläne von `slice-handbuch-meldungskatalog` fuhr deshalb ein Wegwerfprogramm außerhalb des Repos mit demselben Listenformat.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Aufnahme des Werkzeugs in die Gate-Kette, ein `make`-Ziel oder ein Aufruf durch die Rollen — Entscheidung des Users ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), Roadmap *Offene Wellen*, Eintrag `slice-mutationslauf-werkzeug`).
- Ein Cache über Läufe, ein Gate für das Modul `tools/mutationslauf`, die Klasse `nebenläufig` — andere Punkte der Offenen von `SPEC-210`, beim User.
- Änderungen an `tools/katalog` und `tools/zeichen` — der Slice liefert den Fänger am Werkzeug, die Pakete bleiben unverändert.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Eine Liste für ein Paket an der Modulwurzel (mindestens `tools/katalog` und `tools/zeichen` als Elemente der Menge) läuft: Nullprobe im Verzeichnis des Moduls, Mutationen mit dem Ergebnis rot oder grün; jede Zusage des geänderten Vertrags hat eine Zeile in der Tabelle *Zusage · Fänger* (§3) mit einer vom Implementer rot gesehenen Mutation, als Schleife über die Menge der Pfadformen (Wurzel, Unterpaket, Datei an der Wurzel, Datei im Unterpaket) (`AGENTS.md` §3.15, §3.17).
- [ ] `SPEC-210` nennt die Lesart und die Randformen (Paket gleich Modulwurzel); Entscheidung des Architect vor dem ersten Code-Commit.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/mutationslauf/lauf.go` | update | `relInModul`: Paket gleich Modulwurzel ergibt `.` (der Pfad im Modul), Aufrufer `paketArg`, Nullprobe und Mutation |
| `tools/mutationslauf/nullprobe.go` | update | die Nullprobe läuft im Verzeichnis des Moduls (heute nicht vorhanden bei Paket gleich Wurzel) |
| `tools/mutationslauf/*_test.go` | update | Fänger je Pfadform, Schleife über die Menge |
| `harness/vertraege/SPEC-210-mutationslauf.md` | update (Architect) | Lesart, Randformen, *Ungeprüft* |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Lesart, Randformen, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-handbuch-meldungskatalog` liegt in `done/` (das Programm `tools/katalog` ist das Paket an der Modulwurzel, an dem der Fehler auftrat). Kein weiterer Trigger; der Slice hat keinen Code-Bezug zu einem Produkt-Slice.

**Größenmarke:** Prosa, kein Block. Die Marke setzt der Planner vor dem ersten Code-Commit nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect. Erwartet ist ein kleiner Slice (zwei Funktionen, Fänger, Vertragsnummer).

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Architect entscheidet, dass die Lesart eine Änderung der Gate-Kette verlangt.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (geschärfte Regel, neuer Sensor oder benannte Spec-Lücke in `SPEC-210`).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Eine Änderung von `relInModul` verschiebt die Pfade der Pakete in Unterverzeichnissen (`tools/zeichen/…`, `tools/mutationslauf/…`), die heute laufen — **Ausgang:** weiter offen: der Fänger führt alle vier Pfadformen als Schleife (DoD).
- Randformen von Paket gleich Modulwurzel (Datei an der Wurzel, Paket mit Nachsilbe `/`, Modul ohne Unterpaket) sind in `SPEC-210` nicht entschieden — **Ausgang:** weiter offen: der Architect füllt das Raster *Operation × Argumentart* vor dem ersten Code-Commit (`AGENTS.md` §3.16); der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `slice-handbuch-meldungskatalog`; noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` und `BEO-REPO/raten-messung-mit-zusatzlast-im-selben-container` berühren die Läufe des Werkzeugs, nicht diese Änderung; kein Eintrag erreicht mit diesem Slice 3×.

# Slice slice-test-flackern: Drei Tests, die unter Last rot werden, hängen nicht mehr an Zeit oder Reihenfolge

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung, die von der DoD verschieden wäre.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe), [ADR-0007](../../adr/0007-race-detektor-im-gate.md), Register `BEO-REPO/test-fremden-pakets-flackert-unter-last`.

**Berührte Spec-Stellen:** —.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Lese-Schritt der Closure von `welle-betrieb`). **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** `TestGleichzeitigeSpans` (`src/hexagon/services/hybrid`), `TestUeberlangeEingabe` (Paket `urlparser`) und `TestAnfrageBodyBytes` (`src/adapters/driving/httpodata`) werden unter Last nicht mehr aus einem Grund rot, den das Produkt nicht hat.

**Übernimmt** (Lese-Schritt, `AGENTS.md` §3.13): die offenen Fälle aus `BEO-REPO/test-fremden-pakets-flackert-unter-last`. Die Belege liegen unter `evidence/` des Eintrags; die Ursache ist allein für den ersten Test gelesen (Beleg `slice-admin-tls-betrieb`), nicht nachgestellt.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `TestLeserHaeltNieAn` — behoben mit [`slice-lasttest-leser-zeitgrenze`](../done/slice-lasttest-leser-zeitgrenze.md).
- `TestWettlaufNebenlaeufig` — behoben mit [`slice-token-datei-wettlauf`](../done/slice-token-datei-wettlauf.md).
- Eine Änderung am Produktcode — ein Fehler des Produkts geht an den Planner.
- Eine höhere Zeitgrenze als Behebung — sie verschiebt den Bruch nur.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] `TestGleichzeitigeSpans` wartet an einem Haltepunkt statt auf die Reihenfolge der Goroutinen; die Ursache ist vorher nachgestellt.
- [ ] Für `TestUeberlangeEingabe` und `TestAnfrageBodyBytes` ist die Ursache untersucht; je Test eine Behebung oder ein akzeptiertes Negativ mit Grund, Folge und Maß.
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
| `src/hexagon/services/hybrid/telemetrie_gleichzeitig_test.go` | update | Haltepunkt |
| `grenzen_test.go` des Pakets `urlparser`, `src/adapters/driving/httpodata/anfrage_test.go` | update | nach Befund |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Trigger außer der Priorisierung; nicht Teil der Erstabnahme.

**Größenmarke:** Den Block nach `SPEC-128` Nr. 1 setzt der Planner beim Start.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß): über der Schnittgrenze nach `SPEC-128` (nicht zu erwarten).
- `in-progress` → `open` (blockiert): eine Ursache liegt im Produkt.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Eine Ursache lässt sich nicht nachstellen — **Ausgang:** weiter offen; dann trägt der zweite Liefer-Punkt ein Negativ mit der gemessenen Zahl der Läufe.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten** (Anlage, 2026-11-28): `BEO-REPO/test-fremden-pakets-flackert-unter-last` (Gegenstand), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (jede Rate mit Last und Messendem). Der Planner sichtet beim Start erneut.

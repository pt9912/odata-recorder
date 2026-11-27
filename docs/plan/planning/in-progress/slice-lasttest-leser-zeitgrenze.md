# Slice slice-lasttest-leser-zeitgrenze: Der Fänger des Lasttest-Lesers hängt nicht an einer Echtzeitgrenze, die unter Last reißt

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung über die DoD hinaus. Er ändert einen Test unter `tests/lasttest/`, nicht die Gate-Kette (Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe), `LH-QA-01` (der Test gehört zum Lasttest).

**Berührte Spec-Stellen:** [SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md) Nr. 10 a (der Leser entleert stderr ohne Unterbrechung); die Zusage bleibt.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Closure von `slice-replay-speicher-und-suche`). **Datum:** 2026-11-27.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice.

**Übernimmt:** aus `slice-replay-speicher-und-suche` (Plan §6, Verifikation V-6): `make test-race` war in einem Lauf von `make gates` rot an `TestLeserHaeltNieAn` (`tests/lasttest/leser_test.go`, „nach 1s kein Wert“) bei hoher Last des Hosts.

**Ziel:** `TestLeserHaeltNieAn` fängt weiter, dass der Leser den Schreiber nie anhält, und wird auf einem belasteten Host nicht allein deshalb rot, weil 20 MiB unter dem Detektor nicht binnen 1 s Echtzeit durch die Pipe gehen (`pipeLauf`, `warteAuf(…, time.Second)`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung am Leser oder an der Zusage von `SPEC-216` Nr. 10 a — der Befund betrifft den Fänger, nicht das Werkzeug; ein Lauf des Lasttests ist nicht nötig.
- Die weiteren lastabhängigen Tests in `urlparser`, `admin` und `httpodata` (`BEO-REPO/test-fremden-pakets-flackert-unter-last`) — andere Pakete, Ursache nicht untersucht; Adresse Planner beim Lese-Schritt des Registers.
- Eine Lockerung eines Gates (`make test-race` bleibt, wie es ist; `AGENTS.md` §3.6).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Ursache belegt: die Dauer von `pipeLauf` mit 20 MiB unter dem Detektor steht mit Last des Hosts und Messendem im Plan.
- [ ] Der Fänger hängt nicht mehr an der Grenze von 1 s, oder die Grenze ist aus der Messung begründet; die Mutationen der Zeile A1 des Plans von `slice-lasttest-auswertung` (Scanner mit 64 KiB, Pause beim Auswerten, Grenze entfällt, Sammler ohne Sperre) bleiben je rot, mit Rate (`AGENTS.md` §3.10, §3.15).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/lasttest/leser_test.go` | update | `pipeLauf` und `TestLeserHaeltNieAn` |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-replay-speicher-und-suche` liegt in `done/` und ist nach `main` gemergt. Wird ein Lauf von `make gates` vorher an diesem Test rot, hat der Slice Vorrang.

**Größenmarke:** Prosa, kein Block; die Marke setzt der Planner vor dem ersten Code-Commit nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14. Erwartet ist ein kleiner Slice (eine Testdatei).

**Rückführungen — vorab benannt:**

- `in-progress` → `open` (blockiert): die Zusage lässt sich ohne Echtzeitgrenze nicht fangen; dann entscheidet der Architect über ein benanntes Negativ.
- `in-progress` → `next` (zu groß): die Marke liegt nach `SPEC-128` über der Schnittgrenze.

## 5. Closure-Trigger

DoD vollständig, Review bestanden, `make gates` grün, Closure-Notiz mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Ohne Zeitgrenze fängt der Test die Mutation „Pause beim Auswerten“ nicht mehr — **Ausgang:** weiter offen: die Mutationen der Zeile A1 sind die DoD.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register.

Beim Anlegen entstanden aus der Closure von `slice-replay-speicher-und-suche`; noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/test-fremden-pakets-flackert-unter-last` (3×, Schwelle erreicht; dieser Slice ist die Adresse für `tests/lasttest`) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (verkörpert, `AGENTS.md` §3.10).

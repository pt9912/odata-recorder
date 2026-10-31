# Slice slice-test-determinism: `make test-determinism` — 100 Wiederholungen des Replay mit gleichen normalisierten Antworten und Cursor-Ständen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-02`, `LH-FA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-038` (Vertrag `make test-determinism`, `harness/vertraege/SPEC-038-test-determinism.md`: *100 Wiederholungen identischer Inputs; identische normalisierte Outputs und Cursor-Stände*; normalisiert werden nur die Header aus `VOLATILE_HEADERS`) · `LH-QA-02.a` (Determinismus-Randbedingungen) · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make test-determinism` (`SPEC-038`) fährt eine feste Replay-Folge 100-mal nach Reset und vergleicht die Läufe: gleiche Statuscodes, fachliche Bodies und relevante Header (normalisiert nur um `VOLATILE_HEADERS`) und gleiche Cursor-Stände. Das ist die Messmethode von `LH-QA-02`: „100 identische Testläufe nach Reset; Vergleich normalisierter Antworten“. Der Slice liefert den Teil für `replay`; der Teil für `mock` hat `slice-mock-abnahme` (100 identische Läufe nach Reset).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Läufe des Mock — `slice-mock-abnahme` führt sie als eigene DoD; ob dessen Test am Ziel `make test-determinism` hängt oder daneben steht, ist nicht entschieden (§6, Frage an den Architect); dieser Slice ändert den Plan dort nicht (`AGENTS.md` §3.13: keine Sendung ohne Annahme).
- Parallele Workspaces und Neustart (`LH-QA-03`) — `slice-mock-abnahme`.
- Zeit- und Hostabhängigkeit der Header jenseits `VOLATILE_HEADERS` — jede andere Abweichung ist nach `SPEC-038` ein Befund, kein Gegenstand.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Das Ziel `make test-determinism` und der Vertrag `SPEC-038` (der Stand „entsteht mit dem Slice, der den Code liefert“ wird zum gelieferten Umfang); Einbindung in die Gate-Kette oder daneben nach der Entscheidung des Users zu `slice-test-integration` §6.
- [ ] 100 Läufe derselben Replay-Folge nach Reset haben gleiche normalisierte Antworten und gleiche Cursor-Stände; die Normalisierung nimmt allein `VOLATILE_HEADERS` aus.
- [ ] Je Zusage eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10): ein Reset, der den Cursor stehen lässt, und eine Abweichung eines Headers außerhalb von `VOLATILE_HEADERS` färben den Test rot; wird ein Fänger nur mit Wahrscheinlichkeit rot, nennt der Plan die Rate mit Zahl.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-038`, Sensors-Tabelle in `harness/README.md`).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `Makefile`, `harness/mk/` (neues Fragment) | update / neu | Ziel `make test-determinism`; Bindung in der Sensors-Tabelle |
| `tests/` | neu | die Replay-Folge, 100 Läufe, Vergleich (Pfad und Aufbau: Architect; Binary oder Handle des Composition Root, §6) |
| `harness/vertraege/SPEC-038-test-determinism.md` | update | Stand und Umfang |
| `harness/README.md` | update | Sensors-Tabelle: Ziel, Vertrag, Bindung |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag unter `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-test-integration` liegen in `done/` (Replay läuft, Aufbau und Referenzdienst stehen). Der Vertrag `SPEC-038` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile` und `harness/mk/` und startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): Die Marke setzt der Planner beim Start nach `AGENTS.md` §3.14 (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Zahl steht erst mit der Tabelle *Zusage · Fänger* des Architect. Der Planner schätzt den Slice als klein.
- `in-progress` → `open` (blockiert): der Reset zwischen den Läufen ist ohne Admin-API nicht erreichbar (siehe §6).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Reset zwischen den Läufen ist vor `slice-bootstrap-m6` nicht über HTTP erreichbar; der Aufbau (laufendes Binary oder Handle des Composition Root, wie im M2-Test) und der Weg des Reset sind eine Entscheidung zwischen Architektur und Test — **Ausgang:** offen; Adresse Architect, entscheidet vor dem Code.
- Doppelte Messung mit `slice-mock-abnahme` (100 identische Läufe des Mock): ob ein Ziel beide fährt oder zwei Tests nebeneinander stehen, ist in keinem Plan entschieden — **Ausgang:** offen; Adresse Architect.
- Laufzeit von 100 Läufen und Einbindung in die Gate-Kette oder daneben — **Ausgang:** offen; folgt der Entscheidung des Users zu `slice-test-integration` §6.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in `SPEC-038` oder „gilt nicht“ mit Grund — **Ausgang:** offen.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (12×, verkörpert in `AGENTS.md` §3.16; „100 identische Läufe“ und „jede Abweichung ist ein Befund“ sind Allaussagen mit Fänger über alle Fälle); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

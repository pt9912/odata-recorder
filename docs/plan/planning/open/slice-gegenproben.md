# Slice slice-gegenproben: Gegenproben für a-check und Lint, außerhalb der Gate-Kette

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-fundament](../welle-fundament.md).

**Bezug:** `LH-QA-06`, [ADR-0002](../../adr/0002-architektur-gate-a-check.md), [ADR-0003](../../adr/0003-lint-profil-golangci.md)

**Berührte Spec-Stellen:** `SPEC-039` (`make a-check`) · `SPEC-101` (Lint-Profil).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make gegenproben` weist in Kopien des Arbeitsbaums nach, dass `a-check` und `make lint` einen Verstoß rot melden und einen erlaubten Fall nicht.
Zu den Fällen für `make lint` gehören die Randformen aus `SPEC-101`, die über das Gate nicht erreichbar sind oder nur ein zweiter Linter deckt: ein fehlendes Profil, eine Direktive mit zwei Leerzeichen nach `//` und die Einrückung des `# Why:`-Blocks (aus [slice-lint-stufe](../done/slice-lint-stufe.md), Review-Befunde F-4 bis F-6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Gegenproben für `docs-check`, `schema-check` und `baseline-verify` — sie haben ihre Gegenprobe zum Teil in den Gates selbst, und kein neuer Verstoß ist offen; sie folgen bei Bedarf als eigener Slice.
- Aufnahme der Gegenproben in `make gates` — sie laufen länger und prüfen die Gates, nicht den Stand; die Entscheidung hält die ADR in diesem Slice fest.
- Neue Gate-Regeln — die Gegenprobe prüft die vorhandenen Regeln und fügt keine hinzu.
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

- [ ] Gegenprobe `a-check-negativ`: In einer Kopie des Arbeitsbaums lehnt `a-check` einen Import des Domain-Kerns auf einen Driven Adapter und einen Import zwischen zwei Adaptern ab und lässt den erlaubten Weg über einen Port zu.
- [ ] Gegenprobe `lint-gegenprobe`: In einer Kopie lehnt `make lint` eine `//nolint`-Direktive und eine Ausnahme ohne `# Why:` ab und lässt eine begründete Ausnahme zu.
- [ ] `make gegenproben` fährt beide außerhalb von `GATE_CHECKS`; die Entscheidung steht als ADR (Gegenproben verlassen die Gate-Kette), Zeilen in `harness/README.md` §Sensors.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review.
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; kein Zähler wird gesetzt. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/harness/a-check-negativ.sh` | neu | Kopie des Baums, Verstoß einfügen, Exit prüfen — nach dem Muster von `pgwire-recorder` |
| `tools/harness/lint-gegenprobe.sh` | neu | Kopie des Baums, `//nolint` und `# Why:`-Lücke einfügen, Exit prüfen |
| harness/vertraege/SPEC-102-gegenproben.md | neu | Vertrag der Gegenproben: Namensregel, Liste, Kopie, Verstoßformen, Exit-Codes |
| harness/mk/gegenproben.mk | neu | `a-check-negativ`, `lint-gegenprobe`, `gegenproben`; nicht in `GATE_CHECKS` |
| docs/plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md (neu), `docs/plan/adr/README.md` | neu / update | Entscheidung mit drei Alternativen; Index mitziehen |
| `harness/README.md` | update | Sensors-Zeilen für die Gegenproben |

- Die Gegenproben arbeiten in einer Kopie des Arbeitsbaums (Wegwerfverzeichnis, `docker cp`), nie im Arbeitsbaum selbst.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-go-skelett-build-test](../done/slice-go-skelett-build-test.md) und [slice-lint-stufe](../in-progress/slice-lint-stufe.md) liegen in `done/`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 600 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen) (Gegenprobe-Code zählt etwa doppelt).
- `in-progress` → `open` (blockiert): das Skelett bietet keinen Ort, an dem ein Verstoß ohne Fremdpaket einzufügen ist.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, `make gegenproben` ist grün und die ADR ist `Accepted` samt Index-Eintrag; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen der Gegenproben (`AGENTS.md` §3.12): die Kopie des Baums scheitert, Docker fehlt, die einzufügende Datei fehlt, ein Verstoß wird nicht gefunden — **Ausgang:** entfallen: entschieden in SPEC-102 vor dem ersten Code-Commit.
- Eine Gegenprobe prüft nur die eingefügten Verstoßformen — **Ausgang:** weiter offen: weitere Formen kommen mit den Bibliotheks-Grenzen (`tech`) in welle-persistenz.
- Die Gegenproben laufen nicht in `make gates` und altern unbemerkt — **Ausgang:** entfallen: die ADR legt den Lauf am Ende jeder Welle und vor dem Release fest, und welle-dev-container fährt ihn automatisch in einem eigenen Container.

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

Wird bei der Closure geschrieben.

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

**Vorgelagert — offene Beobachtungen sichten:** Das [Beobachtungs-Register](../observations/README.md) trägt nur seine `README.md`; keine Treffer.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

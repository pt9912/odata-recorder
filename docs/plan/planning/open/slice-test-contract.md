# Slice slice-test-contract: `make test-contract` — Referenzanfragen und Negativfälle am HTTP-Eingang

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-042` (Vertrag `make test-contract`, `harness/vertraege/SPEC-042-test-contract.md`: *OData-V4-JSON, Metadata, Fehlerfälle und HTTP-Semantik*; Referenzanfragen und Negativfälle) · `SPEC-049` (Abnahmematrix, Fixtures `ODV4-*`) · `LH-RB-01.a` (Nachweis: Referenzanfragen und Negativfälle) · `LH-QA-05.a` (Fehlerklassen) · `SPEC-109` (Tabelle der Codes).
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

**Ziel:** `make test-contract` (`SPEC-042`) fährt Referenzanfragen und Negativfälle gegen den HTTP-Eingang: die Fixtures der Muss-Zeilen aus `SPEC-049` (`ODV4-*`, geliefert von `slice-mock-abnahme`) und je Fehlerklasse von `LH-QA-05` (fehlender Replay-Treffer, ungültige Metadaten, nicht unterstützte Query-Option, unvollständige Datenabdeckung) einen Negativfall, der Fehlerklasse, Status und diagnostischen Kontext prüft, ohne ein Geheimnis offenzulegen (Messmethode von `LH-QA-05`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neue Golden-Fixtures — `slice-mock-abnahme` schreibt sie; dieser Slice ruft sie auf und verschiebt sie nicht (ob die Fixtures dort bereits unter `make test` laufen und hier ein zweites Mal, ist §6).
- Zeilen der Spalte *Soll* und *Nicht in MVP* von `SPEC-049` (ETag, `Prefer`, `$expand`, V2) — die Folgewellen [welle-etag](../welle-etag.md), [welle-expand-batch](../welle-expand-batch.md) und [welle-odata-v2](../welle-odata-v2.md); sie hängen ihre Fälle an dieses Ziel, wenn sie es nutzen wollen (Entscheidung dort).
- Konformität gegen jede OData-Implementierung — Grenze von `SPEC-042`: geprüft wird gegen Referenzanfragen und Negativfälle.

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

- [ ] Das Ziel `make test-contract` und der Vertrag `SPEC-042` (der Stand „entsteht mit dem Slice, der den Code liefert“ wird zum gelieferten Umfang); die Muss-Fixtures `ODV4-*` aus `slice-mock-abnahme` laufen über das Ziel (Aufruf, keine Kopie).
- [ ] Je Fehlerklasse von `LH-QA-05` ein Negativfall am HTTP-Eingang, der Klasse, Status und diagnostischen Kontext prüft und keine Geheimnisse offenlegt.
- [ ] Je Zusage eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10): eine Fehlerklasse, die auf eine andere abgebildet wird, färbt den Negativfall rot.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-042`, Sensors-Tabelle in `harness/README.md`).
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
| `Makefile`, `harness/mk/` (neues Fragment) | update / neu | Ziel `make test-contract`; Bindung in der Sensors-Tabelle |
| `tests/` | neu | Negativfälle je Fehlerklasse von `LH-QA-05`; Aufruf der Fixtures `ODV4-*` (Pfad: Architect) |
| `harness/vertraege/SPEC-042-test-contract.md` | update | Stand und Umfang |
| `harness/README.md` | update | Sensors-Tabelle: Ziel, Vertrag, Bindung |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag unter `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-abnahme` und `slice-bootstrap-m3` liegen in `done/` (die Fixtures `ODV4-*` und der Mock-Betrieb stehen; vier Fehlerklassen von `LH-QA-05` erreichen den Eingang erst dort). Der Vertrag `SPEC-042` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile` und `harness/mk/` und startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): Die Marke setzt der Planner beim Start nach `AGENTS.md` §3.14 (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Zahl steht erst mit der Tabelle *Zusage · Fänger* des Architect. Der Planner schätzt den Slice als klein.
- `in-progress` → `open` (blockiert): eine Fehlerklasse von `LH-QA-05` erreicht den HTTP-Eingang nicht (siehe §6).

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

- Die Fixtures `ODV4-*` laufen nach dem Plan von `slice-mock-abnahme` unter `make test`; läuft derselbe Satz hier ein zweites Mal, ist das eine Entscheidung über Doppelung gegen ein gemeinsames Ziel — **Ausgang:** offen; Adresse Architect (beide Pläne vor dem Code lesen, `AGENTS.md` §3.13).
- Die Fehlerklasse „ungültige Metadaten“ ist beim Import sichtbar, nicht an einer Abfrage am Eingang; der Negativfall läuft dort, wo die Klasse den Eingang erreicht (Admin-Weg erst ab `slice-bootstrap-m6`, sonst über den Handle des Composition Root) — **Ausgang:** offen; Adresse Architect.
- Die Spalten *Soll* und *Nicht in MVP* von `SPEC-049` haben kein Ziel, an das ihre Welle ihre Fälle hängen müsste — **Ausgang:** offen; Entscheidung bei Start der jeweiligen Welle.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in `SPEC-042` oder „gilt nicht“ mit Grund — **Ausgang:** offen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (12×, verkörpert in `AGENTS.md` §3.16; „je Fehlerklasse“ ist eine Allaussage mit Fänger über alle Klassen); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

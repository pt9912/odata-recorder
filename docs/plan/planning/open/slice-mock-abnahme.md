# Slice slice-mock-abnahme: Abnahmetests des Mock: Matrix, Determinismus, Isolation, Neustart

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-QA-02`, `LH-QA-03`, `LH-RB-01`, `LH-FA-08`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-QA-02.a` · `LH-QA-03.a` · `LH-RB-01.a` · `SPEC-049` (Abnahmematrix) · `SPEC-048` (Fault Injection, nur Neustart-Teil) · `architecture.md` §6 (Qualitätssicherung).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Tests in `tests/` belegen den Mock gegen die Zusagen des Lastenhefts, die kein Slice allein trägt: die Muss-Zeilen der Abnahmematrix (`SPEC-049`, `ODV4-*`) als Golden-Fixtures, 100 identische Läufe nach Reset mit gleichen normalisierten Antworten, 20 parallele Workspaces ohne Beobachtung fremden Zustands und Cursors, und ein regulärer Neustart, nach dem abgeschlossene Mutationen und Snapshots vorhanden sind.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Soll-Zeilen der Matrix (`Prefer`, `If-Match`, `412`, zusätzliche Funktionen) — `welle-etag` und `welle-expand-batch`.
- Die Fault-Injection-Matrix der Persistenz (`SPEC-048`) — Bestand von `welle-persistenz`; hier prüft nur der Neustart-Test die Folge für Mutationen und Snapshots.
- Der Lasttest zu `LH-QA-01` — `welle-betrieb`; er braucht die Spans.
- Neue Produktfunktion — der Slice ändert `src/` nur, wo ein Test einen Fehler findet, und meldet ihn dann an den Slice, dem die Zusage gehört.

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

- [ ] Abnahmematrix: die Muss-Zeilen von `SPEC-049` (`ODV4-GET`, `-CRUD`, `-Q`, `-T`, `-H`, `-E`) als Golden-Fixtures, je Fixture-ID ein Test (`LH-RB-01`).
- [ ] Determinismus: 100 identische Läufe nach Reset ergeben gleiche normalisierte Antworten (`LH-QA-02`), Vergleich ohne `VOLATILE_HEADERS`.
- [ ] Isolation und Neustart: mindestens 20 parallele Workspaces beobachten keinen fremden Zustand und keine fremden Cursor; nach regulärem Neustart sind abgeschlossene Mutationen und Snapshots vorhanden (`LH-QA-03`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/` | neu | Golden-Fixtures der Matrix, Läufe über 100 Wiederholungen, Paralleltest, Neustarttest; Pakete im Quellstand der Stufe `source`, damit `make test-race` sie fährt (Muster von `slice-bootstrap-m2`) |
| `spec/spezifikation.md` | update | Fixture-IDs von `SPEC-049` als Testadressen mit Inhalt (Architect) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster entstehen mit der Festlegung der Fixtures in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Fixtures sind „geplante Testadressen“ (`SPEC-049`) und werden vom Architect festgelegt, nicht vom Implementer.

- Die Läufe sind Teil von `make test`: wie lange die 100 Läufe dauern dürfen, ist eine Randform des Vertrags (§6).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-snapshot-use-case` liegt in `done/`. Die Spezifikation trägt die Fixtures der Matrix `SPEC-049` mit Inhalt, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt; die Gruppe der Matrix lässt sich vorab von den drei übrigen Gruppen trennen (kein geteilter Mechanismus), wenn die Schätzung × 1,25 über 3000 liegt.
- `in-progress` → `open` (blockiert — Carveout?): die Fixtures der Matrix sind vom Architect nicht festgelegt (siehe §6).

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

- Die Fixture-IDs `ODV4-*` sind „geplante Testadressen“ ohne Inhalt (`SPEC-049`) — **Ausgang:** offen; der Architect legt Inhalt und Zahl der Fixtures vor dem ersten Code-Commit fest, sonst entscheidet der Implementer, was „abgenommen“ heißt.
- Laufzeit der 100 Läufe und des Paralleltests in `make test` und die Frist der Gegenproben, die Pakete des Produkts mitfahren (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`) — **Ausgang:** offen; Zeitbudget als Randform im Vertrag.
- Raten der nebenläufigen Fänger (20 Workspaces) — **Ausgang:** offen; Rate je Lauf mit Zahl der Prozesse und Messendem (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`).
- Der Neustarttest: im Test-Binary oder als Prozess (`SIGTERM`, `LH-FA-02.c`) — **Ausgang:** offen; Architect.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Offen und für diesen Slice einschlägig: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für den Paralleltest, `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) für die Laufzeit der Läufe in der Gate-Kette, `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (7×, `AGENTS.md` §3.10). Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

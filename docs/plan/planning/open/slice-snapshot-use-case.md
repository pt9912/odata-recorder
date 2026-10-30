# Slice slice-snapshot-use-case: SnapshotUseCase und Restore des Workspace

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-10`, `LH-FA-09`, `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-10.a` · `LH-FA-09.b` (Abgrenzung zum Reset) · `SPEC-108` Nr. 12 bis 16 (Workspace-Snapshot) · `SPEC-025`, `SPEC-065`, `SPEC-094`, `SPEC-096` · `ARC-002`, `ARC-003` · `architecture.md` §4.5.
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

**Ziel:** `SnapshotUseCase` erzeugt benannte Snapshots, die ausschließlich den Entity-State eines Workspace festhalten, und listet sie; `WorkspaceUseCase.Restore` ersetzt den Entity-State atomar durch den eines Snapshots, erhöht die Revision und lässt alle Replay-Cursor unverändert; unbekannte oder fremde Snapshot-IDs, ein Snapshot eines anderen Service, ein vergebener Name und die Art `baseline` ergeben die Codes der Spezifikation, ohne dass ein Teilzustand übernommen wird.

**Übernimmt:** den Punkt „Snapshots erzeugen und wiederherstellen“ aus `slice-workspace-use-case` (dessen §1 *Ausdrücklich NICHT* nennt `SnapshotUseCase` und `Restore` als Sendung an den Mock).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte `GET` und `POST …/snapshots` und `POST …/restore` — `welle-admin-api`; der Use Case ist über den Composition Root erreichbar.
- Löschen einzelner Snapshots — `SPEC-108` kennt es nicht (ein Snapshot geht nur mit seinem Workspace).
- Versionskontrolle mit Merge von Snapshot-Zweigen — Out-of-Scope von `LH-FA-10`.

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

- [ ] Erzeugen und Auflisten: Name im Workspace eindeutig (`ODR-E7006`), nur der Entity-State im Manifest, Cursor nie Bestandteil; ein leerer Workspace ergibt einen leeren Snapshot (`LH-FA-10`, Boundary).
- [ ] Restore: atomar, neue Revision, Cursor unverändert; unbekannte oder fremde ID `ODR-E7002`, anderer Service `ODR-E5004`, Art `baseline` `ODR-E7008`, jeweils ohne Teilzustand.
- [ ] Restore und Mutation desselben Workspace sind serialisiert; die Zusage hat einen Wirkungs-Fänger, nicht nur den Detektor (`AGENTS.md` §3.10).
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
| `src/hexagon/ports/inbound/` | update | `SnapshotUseCase`, `Restore` am `WorkspaceUseCase` |
| `src/hexagon/services/snapshot/` | neu | Erzeugen und Auflisten, Manifest über `BlobRepository` |
| `src/hexagon/services/workspace/` | update | `Restore` ohne Zugriff auf `CursorRepository` |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-10` (Happy, leerer Workspace mit fortgeschrittenem Cursor, ungültige ID); der Fake von `CursorRepository` zählt jede Operation, damit „Restore berührt keinen Cursor“ fangbar ist |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Fake eines Ports hält jedes Argument fest und zählt jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-schreiben` liegt in `done/` (erst dort entstehen die Änderungen, die ein Snapshot festhält). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): Form des Manifests und des `state_hash` sind nicht entschieden (siehe §6).

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

- Abbruch zwischen Schreiben des Manifest-Blobs und dem Commit des Snapshots (Blob zuerst, dann Referenz; ein verwaister Blob ist zulässig, `SPEC-048`) — **Ausgang:** offen; Fänger durch Fehler-Injektion an den Haltepunkten des Ports, nicht am `ctx`-Argument allein.
- Bildung von `state_hash` und Manifest (deterministisch, `LH-QA-02`; Reihenfolge der Entities, Darstellung der Maske) — **Ausgang:** offen; Architect.
- Restore gleichzeitig mit Mutation und mit Reset desselben Workspace — **Ausgang:** offen; Serialisierung je Workspace aus `slice-mock-schreiben`, Wirkungs-Fänger mit Rate je Lauf (`AGENTS.md` §3.10).
- Restore einer Maske: ein Snapshot hält Entities mit Property-Maske; Wiederherstellung darf Unbekanntes nicht zu `null` machen (`SPEC-044`) — **Ausgang:** offen; Raster-Zelle.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Offen und für diesen Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes von `CursorRepository`, `SnapshotRepository` und `BlobRepository`; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für Restore unter Mutation. Verkörpert: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15). Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

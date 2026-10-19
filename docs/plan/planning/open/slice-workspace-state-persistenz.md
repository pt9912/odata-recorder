# Slice slice-workspace-state-persistenz: Workspace-Zustand: Workspace, Entity, Cursor, Wartung, Rundlauf

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-persistenz](../welle-persistenz.md).

**Bezug:** `LH-FA-04`, `LH-FA-07`, `LH-FA-09`, `LH-FA-10`, [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)

**Berührte Spec-Stellen:** `SPEC-003` (Workspace und Entity) · `SPEC-004` · `ARC-004` · `ARC-009`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-19.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `WorkspaceRepository`, `EntityRepository`, `CursorRepository`, `SnapshotRepository` (Art `workspace`) und `StorageMaintenance` sind Ports und im SQLite-Adapter erfüllt, und ein Rundlauftest belegt alle Repositories zusammen auf einer Datenbank.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die CRUD-Logik, das Replay-Matching und der Cursor-Fortschritt als Fachlogik — Fachlogik der Welle welle-record-replay; hier liegt nur die Ablage.
- Der Abgleich der Blobs gegen die Referenzmenge als Use Case — `MaintenanceUseCase` kommt mit den Services; hier liefert `StorageMaintenance` die Menge und die Integritätsprüfung.
- Aufbewahrung nach Alter (`RETENTION_MAX_AGE`) — Fachlogik der Folgewelle.

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

- [ ] Domain-Modell und Ports für Workspace (Revision, Baseline-Bezug), Entity (Version, Vollständigkeit, gelöscht), Cursor und Workspace-Snapshots; die Adapter-Implementierung auf den Tabellen aus `SPEC-004`, mit `WithWorkspaceTransaction`.
- [ ] Reset und Löschen eines Workspace laufen in einer Transaktion über Workspace, Entity, Cursor und Snapshot; `StorageMaintenance` liefert `foreign_key_check`, `integrity_check` und die Menge der referenzierten Blob-Hashes.
- [ ] Rundlauftest: alle Repositories schreiben in eine Datei, die Datenbank wird geschlossen und wieder geöffnet, alles liest sich zurück, `foreign_key_check` und `integrity_check` sind leer.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/hexagon/model/` | update | Workspace, Entity, Cursor, Snapshot (workspace) |
| `src/hexagon/ports/outbound/` | update | fünf Ports und `WithWorkspaceTransaction` |
| `src/adapters/driven/sqlite/` | update | Implementierungen auf `workspaces`, `entities`, `replay_cursors`, `snapshots` |
| `spec/spezifikation.md` | update | Randformen: Revision veraltet (Konflikt), Löschen mit offenen Cursors, Restore auf unbekannten Snapshot, leerer Workspace |

- Der Rundlauftest ist das *Mehr* der Welle gegenüber den Slice-DoDs: er fährt alle Repositories auf derselben Datei, über ein Schließen und Wiederöffnen hinweg.

**Zusage · Fänger** (`AGENTS.md` §3.15):

| Zusage | Fänger |
|---|---|
| Reset ist atomar | Test: Fehler beim Cursor-Löschen lässt Entity-State unverändert; Mutation: Transaktion entfernt |
| Veraltete Revision wird abgelehnt | Test: zweiter Schreiber mit alter Revision; Mutation: Prüfung entfernt |
| Löschen eines Workspace nimmt Entities und Cursor mit | Test: danach keine Zeilen mehr; Mutation: `CASCADE` im Test umgangen |
| `StorageMaintenance` meldet Verletzungen | Test: absichtlich verletzte Datenbank (Fremdschlüssel aus) liefert Befund |
| Der Rundlauf bleibt über ein Wiederöffnen gültig | Rundlauftest |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-baseline-persistenz](slice-baseline-persistenz.md) und [slice-blob-ablage](slice-blob-ablage.md) liegen in `done/`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 900 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): die Transaktionsgrenze über fünf Ports lässt sich nicht fachlich ausdrücken, ohne SQL im Port.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, der Rundlauftest ist grün und eine Mutation je Zusage wurde rot gesehen; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen der Repositories (`AGENTS.md` §3.12): leerer, fehlender oder unbekannter Wert, doppelter Schlüssel, Abbruch mitten in der Transaktion, gleichzeitiger Zugriff — **Ausgang:** entfallen: entschieden in `spezifikation.md` (Produkt) vor dem ersten Code-Commit.
- Der Slice ist der größte der Welle und kann die Marke überschreiten — **Ausgang:** weiter offen: wird bei der Rückführung in §4 benannt und dann als Beobachtung geführt.
- Der Cursor-Zustand (`consumed_json`) ist frei geformt und wächst mit dem Replay — **Ausgang:** weiter offen: wird mit welle-record-replay sichtbar.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (2×), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (1×) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15); sie fließen als Randformen in §6 und als Tabelle *Zusage · Fänger* in §3 ein.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

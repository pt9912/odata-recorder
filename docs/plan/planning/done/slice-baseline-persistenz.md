# Slice slice-baseline-persistenz: Baseline-Persistenz: Service, Snapshot, Import, Coverage

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-persistenz](../welle-persistenz.md).

**Bezug:** `LH-FA-10`, `LH-FA-11`, `LH-FA-18`, [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)

**Berührte Spec-Stellen:** `SPEC-107` (Vertrag der Repositories) · `SPEC-043` (Import- und Provenienz-Schema) · `SPEC-044` (Baseline-Coverage) · `SPEC-004` · `ARC-004` · `ARC-009`.
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

**Ziel:** `MetadataRepository`, `SnapshotRepository` (Art `baseline`), `ImportRepository` und `CoverageRepository` sind Ports und im SQLite-Adapter erfüllt, und die Import-Publikation läuft in einer fachlichen Transaktion.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Import-Logik (OData-Interpretation, Coverage-Berechnung) — Fachlogik der Welle welle-record-replay; hier liegt nur die Ablage.
- Workspaces, Entities und Snapshots der Art `workspace` — [slice-workspace-state-persistenz](../next/slice-workspace-state-persistenz.md).
- Admin-API und Löschen von Sessions als Use Case — Services und Adapter der Folgewellen.

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

- [x] Domain-Modell und Ports für Service-Metadaten, Baseline-Snapshots, Import-Einträge, Baseline-Quellen und Coverage; die Adapter-Implementierung auf den Tabellen aus `SPEC-004`.
- [x] Import-Publikation in einer Transaktion: Snapshot, Import-Eintrag, Baseline-Quellen und Coverage werden ganz oder gar nicht geschrieben; ein zweiter Import mit demselben Anfrage-Hash ist idempotent.
- [x] Vertragstests am Port gegen SQLite einschließlich der Randformen aus §6.
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/model/` | update | Service, Snapshot (baseline), Import, Publication, Coverage mit `Validate`; Modelltests |
| `src/hexagon/ports/outbound/` | update | vier Ports (`MetadataRepository`, `SnapshotRepository`, `ImportRepository`, `CoverageRepository`); `Publish` ist die Transaktionsgrenze der Import-Publikation |
| `src/hexagon/ports/outbound/outboundtest/` | update | Vertragstests `RunServiceVertrag`, `RunBaselineVertrag` und zwei Hooks in `Verbindung` |
| `src/adapters/driven/sqlite/` | update | Implementierungen auf den Tabellen `services`, `snapshots`, `imports`, `baseline_sources`, `coverage`; Adapter-Tests |
| `spec/spezifikation.md` | update | `SPEC-107`: Randformen leerer Import, doppelter Anfrage-Hash mit anderem Inhalt, Quelle nicht vorhanden oder eines anderen Service, Abbruch mitten in der Publikation |

- Der Vertrag ist `SPEC-107`; Löschen von Services und Baselines liefert dieser Slice nicht (kein Konsument in der Welle).
- Die Fremdschlüssel des Schemas (`baseline_kind`, `uq_import_baseline`, `fk_baseline_sources_kind`) sind der Rückhalt; die Tests belegen sie über den Port, nicht über rohes SQL.

**Zusage · Fänger** (`AGENTS.md` §3.15). Die Tests liegen als Vertrag in `outboundtest` (`TestServiceVertrag`, `TestBaselineVertrag`, gegen SQLite in `repositories_test.go`), die Hooks `FuegeWorkspaceSnapshotHinzu` und `BrichNachImportAb` setzen der Adapter die Zustände, die kein Port erzeugt:

| Zusage | Fänger (Test) · Mutation |
|---|---|
| Publikation ist atomar, auch beim Abbruch mitten im Schreiben (SPEC-107 Nr. 5) | `Publish ist atomar` (Quelle nach Snapshot und Import unbekannt oder fremd, Kontextabbruch nach dem Import-Eintrag über den Hook) · `ROLLBACK` in `inTx` durch `COMMIT` ersetzt; Fehler von `schreibePublikation` verschluckt; `ROLLBACK` mit dem abgebrochenen Kontext |
| Gleicher Anfrage-Hash, gleicher Import, auch bei abweichendem Inhalt; anderer Service ist eigene Anfrage; gleichzeitig (Nr. 7) | `gleicher Anfrage-Hash ist idempotent`, `gleichzeitige Publish mit gleichem Hash ordnen sich` · Hash-Abfrage unwirksam; Hash über Services; Treffer als neu gemeldet; `BEGIN` statt `BEGIN IMMEDIATE` |
| Gültigkeit der Publikation, leerer Import, leere Coverage zulässig (Nr. 6) | `ungültige Publikationen werden abgelehnt und schreiben nichts` (28 Verstöße, darunter Evidenz und `report_json` ohne gültiges UTF-8), `TestPublicationValidate` und `TestImportValidate` · je `case` in `baseline.go` durch `false` ersetzt, je Prüfung der Liste in `Publication.Validate` entfernt, Duplikat-Prüfungen entfernt, `Validate` in `Publish` entfernt |
| Prüfreihenfolge (Nr. 8), eine leere Quell-Session ist unbekannt | `Prüfreihenfolge von Publish` (ein Fall je Paar), `Publish ist atomar` (Quellen `s1` und leere Kennung) · Leere Quelle in `schreibeQuelle` übersprungen; Kennungen vor Hash; Service-Prüfung entfernt; Quelle vor Kennungen; Quellen rückwärts; Kontext vor Gültigkeit |
| Baseline unveränderlich, auch gegen einen `workspace`-Snapshot gleicher Kennung (Kopf) | `Baselines sind unveränderlich` · Kennungsprüfung entfernt; zusätzlich `INSERT OR REPLACE` |
| Quellen verhindern das Löschen der Session, `BaselinesOf`, `Sources`, `Get` des Imports (Nr. 9, 10) | `Baseline-Quellen: BaselinesOf, Sources, Löschen der Session` · Quelle nicht geschrieben; Sortierung absteigend; Prüfung der Session und der Baseline entfernt; Fehler des Imports vertauscht |
| Quelle eines anderen Service ist `ErrUngueltig` (Nr. 8) | `Publish ist atomar`, `Prüfreihenfolge von Publish` · Prüfung des Service der Quelle entfernt |
| Nur Snapshots der Art `baseline` sichtbar (Nr. 4) | `nur Baselines sind sichtbar, geordnet nach Zeit`, `Coverage lesen`, `Baseline-Quellen` (Hook legt einen `workspace`-Snapshot an) · Filter in `Get`, `List` und `baselineBekannt` einzeln entfernt |
| Ordnung nach Zeit, nicht nach Text; Zeiten als UTC (Nr. 4, SPEC-105) | `nur Baselines sind sichtbar, geordnet nach Zeit` (Zone `UTC-5`, gleiche Zeit), `TestPublishSchreibtArtZeitformatUndHaeltDieSchluessel` (Import und Snapshot in `UTC+5`, der Text ist UTC) · Kennung vor Zeit; Zeit ohne Umrechnung in UTC bei Snapshot und Import |
| Service und Metadaten (Nr. 1–3) | `TestServiceVertrag` (fünf Fälle) · `Validate` und Prüfung in `SetMetadata` entfernt; doppeltes `Create`; `List` absteigend; Fehler von `Get` und `SetMetadata`; Blob-Format in `BlobReferenzGueltig` (Präfix, Länge, Kleinbuchstaben); `SetMetadata` prüft die Kennung vor den Argumenten; leere Kennung als `ErrUngueltig` statt `ErrNichtGefunden` bei `Metadata.Get` und `SetMetadata` (Nr. 3, 13) |
| Coverage lesen, Evidenz byte-gleich, leere Argumente, unbekannte Baseline (Nr. 11) | `Coverage lesen, Evidenz byte-gleich` (Leerzeichen, Zeilenumbruch, Umlaut, Schlüsselfolge) · Evidenz mit `json()` in `Get` und in `List` minifiziert; Sortierung absteigend; Prüfung leerer Argumente und der Baseline entfernt |
| Kontext (Nr. 12), Argument vor Kontext | `abgebrochener Kontext beendet jede Operation ohne Änderung` (Service und Baseline) · Kontext in `Publish`, in den Lesern und in `Metadata` ignoriert; `ctx.Err()` vor der Argumentprüfung in `SetMetadata` und `Coverage.List` |
| Leere Kennung: `ErrUngueltig` bei `Coverage.Get`, `Coverage.List`, `Snapshots.List`, sonst `ErrNichtGefunden` (Nr. 13) | `nur Baselines sind sichtbar…`, `Baseline-Quellen…`, `Get und List`, `SetMetadata ersetzt…`, `abgebrochener Kontext…` · je Leser (`Snapshots.Get`, `Imports.Get`, `Sources`, `BaselinesOf`, `Metadata.Get`, `SetMetadata`) mit `ErrUngueltig` bei leerer Kennung |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-session-recording-persistenz](../done/slice-session-recording-persistenz.md) liegt in `done/`.

**Marke überschritten:** Die Summe aus `git diff --numstat` über die Pfade aus §3 liegt bei 1567 Zeilen über alle Pfade aus §3 einschließlich `spec/spezifikation.md` (11 Zeilen), 1556 über `src/`. Sie tragen die Vertragstests (`outboundtest`, 666 Zeilen), die Modell- und Adapter-Tests (205 Zeilen), der Adapter (430 Zeilen) sowie Modell und Ports (255 Zeilen). Der Slice bleibt ungeteilt, weil Modell, Ports und Adapter einzeln nichts liefern und die Publikation nur mit ihren Lesern prüfbar ist.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 700 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): das Modell der Coverage-Evidenz verlangt eine Entscheidung der Spezifikation, die fehlt.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und eine Mutation je Zusage wurde rot gesehen; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen der Repositories (`AGENTS.md` §3.12): leerer, fehlender oder unbekannter Wert, doppelter Schlüssel, Abbruch mitten in der Transaktion, gleichzeitiger Zugriff — **Ausgang:** entfallen: entschieden in `spezifikation.md` (Produkt) vor dem ersten Code-Commit.
- Randformen, die der Review nach dem ersten Code nannte (Reihenfolge in `SetMetadata`, leere Kennungen, ungültiges UTF-8 in JSON, Coverage unter `Snapshot.ID`, leere Quell-Session, nicht geprüfte Kennungen mit NUL oder nur Leerzeichen): jetzt in `SPEC-107` entschieden bzw. unter *Ungeprüft* genannt — **Ausgang:** entfallen, die Spezifikation trägt sie.
- `Sources`, `BaselinesOf` und `Coverage.List` lesen Kennung und Inhalt in zwei Abfragen ohne Transaktion; ohne Löschen von Baselines hat das keine Wirkung (`SPEC-107` *Ungeprüft*) — **Ausgang:** weiter offen, wird mit dem Löschen von Baselines zu einem Wettlauf; es gibt heute keinen Slice dafür (keine Adresse, `AGENTS.md` §3.13), der Planner führt ihn mit dem Löschen.
- Eine `complete`-Coverage ohne Beweis wird in der Ablage nicht abgelehnt (`SPEC-107` *Ungeprüft*, `LH-FA-11` verlangt nachweislich vollständig) — **Ausgang:** weiter offen; der Planner entscheidet, wo der Satz durchgesetzt wird (Fachlogik des Imports oder Ablage); keine Adresse, weil kein Slice den Punkt annimmt.
- Die Evidenz der Coverage (`evidence_json`) ist frei geformt und kann mit der Fachlogik wachsen — **Ausgang:** weiter offen: wird mit welle-record-replay sichtbar und dann als Beobachtung geführt.

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

- **Was hat funktioniert:** Modell, vier Ports und SQLite-Adapter erfüllen `SPEC-107`; die Publikation ist atomar, idempotent über den Anfrage-Hash und unter Gleichzeitigkeit geordnet (`BEGIN IMMEDIATE`). Die Fänger der Tabelle *Zusage · Fänger* wurden mit rund 190 Mutationen rot gesehen (Implementer, Reviewer 105, Verifier 53).
- **Was ging anders als geplant:** Das Review fand 7 MEDIUM (vier Zusagen ohne rot werdende Mutation, drei Randformen im Code statt vor dem Code entschieden), die Verifikation bestand mit Auflagen (leere Quell-Session ohne Fänger, Köpfe der Vertragsläufe, Marke ohne Spec). `SPEC-107` wuchs um die Reihenfolge in `SetMetadata`, die leere Kennung je Leser, gültiges UTF-8 und die Quelle ohne Kennung; NUL-Kennungen, `complete` ohne Beweis und der Lese-Wettlauf stehen unter *Ungeprüft*. Die Marke ist überschritten und in §4 benannt (1567 Zeilen).
- **Steering-Loop-Eintrag:** Nichts verkörpert; gezählt, nicht verkörpert: `BEO-REPO/randform-erst-im-review-entschieden` steht jetzt bei 6× (der Eintrag geht in die Closure von welle-persistenz) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` bei 8× (verkörpert in `AGENTS.md` §3.15).
- **Beobachtungs-Register (`../observations/`):** je eine weitere Datei `evidence/slice-baseline-persistenz.md` in den beiden Einträgen; `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` blieb ohne neues Auftreten.
- **Folge-Slices:** keine. Drei Punkte bleiben ohne Adresse beim Planner (`AGENTS.md` §3.13, kein Nehmer in der Welle): wo `LH-FA-11` (`complete` nur mit Beweis) durchgesetzt wird, das Löschen von Baselines samt Transaktion der Leser, und die Form der Evidenz.
- **Risiken aus §6:** Randformen *entfallen*, die drei übrigen *weiter offen*, siehe §6.
- **Drei Paarungen:** von der Closure der Welle welle-persistenz geprüft.

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

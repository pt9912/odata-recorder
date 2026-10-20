# Welle welle-persistenz: Persistenz hinter den Outbound Ports

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. **Geplante Wellen bekommen noch keine
Datei:** Sie stehen in der Roadmap unter *Nächste Wellen* und nirgends sonst —
zwei Positionen, nicht drei.

**Zielmeilenstein:** kein Meilenstein-Bezug; sie ist Vorbedingung von M2 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-19.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Die Daten des Recorders liegen in einer SQLite-Datei und im Dateisystem hinter den Outbound Ports: Sessions, Exchanges, Blobs, Service-Metadaten, Baselines mit Import und Coverage, Workspaces mit Entity-State, Cursor und Snapshots, jeweils mit den Transaktionsgrenzen der Architektur.
Eine Welle, weil der Closure-Trigger mehr beobachtet als die DoDs der Slices: das Zusammenspiel aller Repositories auf einer Datenbank, über ein Schließen und Wiederöffnen hinweg.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-fundament` liegt in `done/` (Gates für Build, Test, Lint und `a-check`, dazu die Gegenproben).
- [ADR-0006](../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) ist `Accepted` (SQLite-Bibliothek).

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün und führt dabei `migration-check` aus; `make gegenproben` ist grün.
- Der Rundlauftest aus `slice-workspace-state-persistenz` ist grün: alle Repositories schreiben in eine Datei, die Datenbank wird geschlossen und wieder geöffnet, alles liest sich zurück, `foreign_key_check` und `integrity_check` sind leer.
- Die Closure-Notiz liegt als `welle-persistenz-results.md` im Ruheort `done/`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-sqlite-adapter-migration | SQLite-Adapter mit Migration aus dem Schema | `LH-FA-03` |
| slice-session-recording-persistenz | Session und Recording: Modell, Ports und SQLite-Repositories | `LH-FA-02`, `LH-FA-03` |
| slice-blob-ablage | Blob-Ablage: content-adressiert, mit Capture-Abschluss | `LH-FA-02`, `LH-FA-14` |
| slice-baseline-persistenz | Baseline-Persistenz: Service, Snapshot, Import, Coverage | `LH-FA-10`, `LH-FA-11`, `LH-FA-18` |
| slice-workspace-state-persistenz | Workspace-Zustand: Workspace, Entity, Cursor, Wartung, Rundlauf | `LH-FA-04`, `LH-FA-07`, `LH-FA-09`, `LH-FA-10` |

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: Welle welle-record-replay (braucht die Repositories für Capture, Matching und Cursor).
- Wird blockiert von: —.
- Reihenfolge der Slices: `slice-sqlite-adapter-migration` zuerst; danach `slice-session-recording-persistenz` und `slice-blob-ablage`; dann `slice-baseline-persistenz`; zuletzt `slice-workspace-state-persistenz`.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Fachlogik (Record, Replay, Matching, Import-Interpretation, CRUD, Query-Interpreter) — Welle welle-record-replay; die Welle liefert Ablage und Ports, keine Use Cases.
- Services, HTTP- und Admin-Adapter — Folgewellen; die Repositories werden hier nur über ihre Ports und die Vertragstests geprüft.
- Migrationskette über mehrere Schemaversionen — braucht eine zweite Schemaversion.
- Coverage-Schwelle und Lasttest (`LH-QA-01`) — kommen mit der Fachlogik, nicht mit der Ablage.
- Aufbewahrung nach Alter und Startwartung als Use Case — kommen mit den Services; hier liefern die Ports die Operationen.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-persistenz-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).

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

**Berührte Spec-Stellen:** `SPEC-108` (Vertrag der Repositories) · `SPEC-003` (Workspace und Entity) · `SPEC-004` · `ARC-004` · `ARC-009`.
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

**Ziel:** `WorkspaceRepository`, `EntityRepository`, `CursorRepository`, die Workspace-Operationen des `SnapshotRepository` (Art `workspace`), `StorageMaintenance` und `WithWorkspaceTransaction` sind Ports und im SQLite-Adapter erfüllt; die Repositories laufen eigenständig und innerhalb der Transaktion, und ein Rundlauftest belegt alle Repositories zusammen auf einer Datenbank über ein Wiederöffnen.

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
| `src/hexagon/model/` | update | `Workspace`, `EntityState`, `ReplayCursor`, `WorkspaceSnapshot` mit `Validate` |
| `src/hexagon/ports/outbound/` | update | `WorkspaceRepository`, `EntityRepository`, `CursorRepository`, `StorageMaintenance`, `WorkspaceTransactor` mit `WorkspaceStores`; `SnapshotRepository` um `CreateWorkspace`, `GetWorkspace`, `ListWorkspace` erweitert; `ErrRevisionVeraltet`, `ErrFremderService`, `ErrFalscheArt` |
| `src/adapters/driven/sqlite/` | update | Implementierungen auf `workspaces`, `entities`, `replay_cursors`, `snapshots`; das Interface `querier` (von `*sql.DB` und `*sql.Conn` erfüllt) und `zugriff` tragen dieselben Repositories eigenständig und auf der Verbindung der Transaktion; `WithWorkspaceTransaction` ruft `inTx` |
| `src/hexagon/ports/outbound/outboundtest/` | update | Verträge `Workspace`, `Entity`, `Cursor`, `WorkspaceSnapshot`, `Wartung`, `Transaktion` (mit den Zusammensetzungen Reset und Restore); `InTransaktion` fährt die ersten vier ein zweites Mal auf den Repositories der Transaktion; Hooks `Zeilen`, `VerletzeFremdschluessel`, `VerletzeCheck`, `BrichNachLoeschenAb` |
| `spec/spezifikation.md` | update | `SPEC-108`: Vertrag der Repositories, entschieden vor dem ersten Code-Commit |
| `spec/architecture.md` | update | `StorageMaintenance` liefert auch die Blob-Referenzen der Service-Metadaten |

- Der Rundlauftest (`TestRundlauf`) ist das *Mehr* der Welle gegenüber den Slice-DoDs: er fährt alle Repositories auf derselben Datei, über ein Schließen und Wiederöffnen hinweg.
- Die Befunde von `Check` sind Texte ohne festes Format; die Tests prüfen, dass der Befund die Quelle (`foreign_key_check`, `integrity_check`) und bei einer Fremdschlüssel-Verletzung die Tabelle nennt.

**Zusage · Fänger** (`AGENTS.md` §3.15). Teiltests heißen wie in den Vertragsdateien; jeder Vertrag läuft eigenständig (`Test<Name>Vertrag`) und auf den Repositories der Transaktion (`Test<Name>VertragInTransaktion`, außer `Wartung` und `Transaktion`). Mutationen sind am geprüften Code gefahren und rot gesehen, Gründe der roten Tests stehen in der letzten Spalte der Zeile in Kurzform; Halbbedingungen der `Validate`-Funktionen fängt je `TestWorkspaceValidate`, `TestEntityStateValidate`, `TestReplayCursorValidate`, `TestWorkspaceSnapshotValidate` und `TestKnownPropertiesJSON`/`TestConsumedJSON`.

| Zusage (`SPEC-108`) | Fänger und Mutationen |
|---|---|
| Kopf: Argument vor Kontext, jede Operation ohne Änderung bei abgebrochenem Kontext (Nr. 17) | je Vertrag `abgebrochener Kontext beendet jede Operation …` (`Workspace`, `Entity`, `Cursor`, `WorkspaceSnapshot`, `Wartung`) und `ein abgebrochener Kontext rollt zurück` (`Transaktion`); Mutationen: `ctx = context.Background()` am Anfang jeder der 25 Operationen (alle rot); Gültigkeit erst nach dem Kontext geprüft bei `Create`, `List`, `Advance`, `Put`, `ListSet`, `ReplaceAll`, `Cursor.Put`, `CreateWorkspace` (alle rot) |
| Kopf: leere Kennung ist `ErrNichtGefunden`, `ErrUngueltig` nur an den genannten Stellen | `leere Kennung` (`Workspace`), `Get, List und ListSet bei Unbekanntem …` (`Entity`), `Get und List bei Unbekanntem …` und `ResetAll setzt zurück` (`Cursor`), `GetWorkspace bei Unbekanntem …`, `ListWorkspace ordnet …`; Mutationen: leere Kennung als `ErrUngueltig` bei `Workspaces.Get`, `Entities.Get`, `Entities.List`, `Cursors.List`, `Cursors.ResetAll`, `ListWorkspace` (alle rot) |
| Kopf: Listen nie `nil` | die Leerfälle in `workspaceFall4`, `entityFall4`/`entityFall6`, `cursorFall1`/`cursorFall5`, `wsSnapshotFall6`, `wartungFall1`/`wartungFall5`; Mutation: `sammle` startet mit `var out []T` (rot) |
| Kopf: Zeiten UTC, Jahr 1 bis 9999 | `Create und Get tragen alle Felder` (Zeit mit Versatz geschrieben, `Location() == UTC` gelesen), `CreateWorkspace und GetWorkspace tragen alle Felder`, `List ordnet nach Zeit, nicht nach Text` (Versatz), `Create lehnt Ungültiges ab …` (Jahr 0 und 10000), `TestWorkspaceZustandSchreibtArtZeitformatUndHaeltDieSchluessel` (Text mit neun Nachkommastellen); Mutationen: Schreiben ohne `UTC()` bei Workspace und Snapshot, Lesen mit `Local()` bei Workspace und Snapshot (rot) |
| Nr. 1: `nil` bestätigt, die Repositories arbeiten auf der Transaktion | `nil bestätigt alles` (Schreiben über alle vier, Lesen in `f` und danach); Mutation: Stores mit `zugriff` ohne Verbindung (rot) |
| Nr. 1: ein Fehler von `f` rollt zurück und kommt unverändert | `ein Fehler rollt zurück und wird unverändert geliefert` (Zeilenzählung in vier Tabellen, Identität des Fehlers, Fehler einer Operation in `f`); Mutationen: Fehler verschluckt, Fehler ersetzt, Rollback entfernt (rot) |
| Nr. 1: eine Panik rollt zurück und wird weitergereicht | `eine Panik rollt zurück und wird weitergereicht`; Mutation: `recover` in `WithWorkspaceTransaction` (rot) |
| Nr. 1: abgebrochener Kontext rollt zurück, auch wenn `f` `nil` liefert | `ein abgebrochener Kontext rollt zurück` (vor dem Beginn: `f` läuft nicht; in `f`: Fehler des Kontexts, Zeilenzählung); Mutationen: `COMMIT` mit `WithoutCancel`, Kontext in `WithWorkspaceTransaction` ignoriert (rot) |
| Nr. 1: Transaktionen laufen nacheinander | `Transaktionen laufen nacheinander` (acht Lesen-und-Erhöhen in je einer Transaktion), `gleichzeitige Advance …`; Mutation: `BEGIN` statt `BEGIN IMMEDIATE` (rot) |
| Nr. 2: `Create` mit Revision 0, ohne Entity-State; Gültigkeit je Feld | `Create und Get tragen alle Felder`, `Create lehnt Ungültiges ab …` (Kennung, Service, Baseline, Zeit, Jahr 0 und 10000, Revision 1 und -1); Mutationen: `Validate` entfernt, jede Halbbedingung (Kennung, Service, Baseline, Zeit, Revision, Revision nur `>0`) einzeln, Revision 1 geschrieben, Entity angelegt (rot) |
| Nr. 2: Prüfreihenfolge Gültigkeit, Kennung, Service, Baseline (unbekannt, Art, Service) | `Prüfreihenfolge von Create` (13 Fälle, jedes Paar benachbarter Verstöße); Mutationen: Service vor Kennung, Baseline vor Service, Service-Fremd vor Art, jede der fünf Prüfungen entfernt (rot) |
| Nr. 3: `Get`, `List` nach Zeit und Kennung, leerer Service ungültig | `Get und List`, `List ordnet nach Zeit, nicht nach Text` (Gleichstand der Zeit, Versatz); Mutationen: absteigend, Kennung zuerst, nur Zeit, Kennung absteigend, ohne Service-Filter, ohne Leerprüfung, `NichtGefunden` entfernt (rot) |
| Nr. 4: `Advance` zählt, lehnt Veraltetes ab, negatives `erwartet` ungültig vor allem | `Advance zählt und lehnt Veraltetes ab` (zu klein, zu groß, unbekannte und leere Kennung, negativ auch bei abgebrochenem Kontext); Mutationen: Prüfung entfernt, `+2`, ohne Schreiben, alle Workspaces, `NichtGefunden` entfernt, Negativprüfung entfernt (rot) |
| Nr. 4: gleichzeitige `Advance`: genau eines gelingt | `gleichzeitige Advance mit derselben Revision` (acht Goroutinen); Mutationen: Prüfung entfernt (alle gelingen), `BEGIN` ohne `IMMEDIATE` (rot) |
| Nr. 5: `Delete` nimmt Entity-State, Cursors und Snapshots mit, andere bleiben | `Delete nimmt Entities, Cursors und Snapshots mit` (Zeilenzählung, zweiter Workspace, Baseline, Sessions); Mutationen: Snapshot-Löschen entfernt, alle Workspace-Snapshots, alle Snapshots, Schema ohne `CASCADE` bei `entities`, Schema ohne beide `CASCADE` bei `replay_cursors` (rot) |
| Nr. 5: `Delete` ist atomar, unbekannt ist `ErrNichtGefunden` | `Delete ist atomar` (Abbruch nach dem Löschen der Snapshots), `leere Kennung`; Mutationen: `COMMIT` und `BEGIN` zwischen den Löschungen, Existenzprüfung entfernt (rot) |
| Nr. 6: `EntityState` gültig | `TestEntityStateValidate`, `Put lehnt Ungültiges ab`; Mutationen: jede der drei Kennungen, Payload (leer, kein UTF-8), Vollständigkeit (leer, ohne `partial`), Version (`<0`, `<=0`), leere, doppelte, nicht-UTF-8-Property, `Validate` in `Put` entfernt (rot) |
| Nr. 7: `Put` ersetzt, trennt Workspace, Menge, Schlüssel; unbekannter Workspace | `Put ersetzt und trennt …`, `Put lehnt Ungültiges ab …` (Gültigkeit vor Workspace); Mutationen: je Spalte beim Ersetzen nicht übernommen (fünf), Workspace-Prüfung entfernt (rot) |
| Nr. 7: `Get` auch Tombstone, Payload byte-gleich in `Get`, `List`, `ListSet`, Properties sortiert | `Put und Get tragen alle Felder, Payload byte-gleich` (Leerraum, Zeilenumbruch, Umlaut, `null`); Mutationen: Payload beim Schreiben, in `Get`, in `List` mit `json()` minifiziert, Tombstones in `Get` und `List` ausgeblendet, Properties unsortiert, in der Eingabe sortiert, `NichtGefunden` entfernt (rot) |
| Nr. 8: `List`, `ListSet` nach Menge und Schlüssel, leerer State, unbekannter Workspace, leeres `entity_set` | `List und ListSet ordnen nach Menge und Schlüssel` (Byte-Ordnung, Großbuchstaben vor Kleinbuchstaben), `Get, List und ListSet bei Unbekanntem, Leerem und Tombstones`; Mutationen: Menge absteigend, Schlüssel absteigend, Schlüssel vor Menge, ohne Workspace-Filter, ohne Mengenfilter, Leerprüfung und Workspace-Prüfung entfernt (rot); `ORDER BY entity_set` allein ist äquivalent, weil der Primärschlüssel-Index die Ordnung liefert |
| Nr. 9: `ReplaceAll` ersetzt genau, ganz oder gar nicht | `ReplaceAll ersetzt den State` (anderer Workspace unberührt, leere Liste), `ReplaceAll ist ganz oder gar nicht` (ungültig, fremder Workspace, doppelt, jeweils nichts geändert), `ReplaceAll ist atomar` (Abbruch nach dem Löschen); Mutationen: ohne Löschen, Löschen aller Workspaces, `Validate`, Workspace-Vergleich und Doppelprüfung entfernt, Doppelprüfung nur über den Schlüssel, `COMMIT` und `BEGIN` zwischen Löschen und Schreiben (rot) |
| Nr. 9: unbekannter Workspace auch bei leerer Liste, Reihenfolge Gültigkeit, Kontext, Workspace | `ReplaceAll prüft in der Reihenfolge der Spezifikation`; Mutationen: Workspace-Prüfung entfernt, Gültigkeit nach dem Kontext (rot); Workspace-Prüfung vor dem Kontext ist äquivalent, weil sie den Kontext selbst nutzt |
| Nr. 10: `ReplayCursor` gültig | `TestReplayCursorValidate`, `Put lehnt Ungültiges ab`; Mutationen: drei Kennungen und Service einzeln, Strategie (ganz, je eine der vier entfernt), Position (ganz, nur zu klein, nur zu groß), `consumed` (`0`, negativ, doppelt), `consumed` unsortiert und in der Eingabe sortiert (rot) |
| Nr. 11: `Put` schreibt oder ersetzt; Prüfreihenfolge Workspace, Session, Service | `Put und Get tragen alle Felder` (vier Strategien, unsortierte Eingabe), `Put ersetzt und trennt die Schlüssel`, `Prüfreihenfolge von Put` (sieben Fälle); Mutationen: Service nur gegen Session, nur gegen Workspace, ohne Prüfung, Service vor Session, Session vor Workspace, `NichtGefunden` der Session entfernt, je Spalte beim Ersetzen nicht übernommen (drei) (rot) |
| Nr. 11: `Get`, `List` nach Replay-, dann Recording-Session | `Get und List bei Unbekanntem und Leerem`, `List ordnet nach Replay- und Recording-Session`; Mutationen: absteigend, Recording zuerst, Recording absteigend, `NichtGefunden` entfernt, ohne Recording-Filter beim `Get` (rot); `ORDER BY replay_session_id` allein ist äquivalent (Primärschlüssel-Index) |
| Nr. 12: `ResetAll` setzt `position` und `consumed` zurück, Rest bleibt, andere Workspaces unberührt | `ResetAll setzt zurück`; Mutationen: Position bleibt, `consumed` bleibt, Strategie geändert, alle Workspaces, Workspace-Prüfung entfernt (rot) |
| Nr. 12: das Löschen einer Recording-Session löscht ihre Cursors | `das Löschen einer Recording-Session löscht ihre Cursors` (Zeilenzählung); Mutation: Schema ohne beide `CASCADE` auf `sessions` (rot) |
| Nr. 13: `CreateWorkspace` gültig, Reihenfolge Workspace, Service, Kennung und Name | `CreateWorkspace lehnt Ungültiges ab` (auch Blob ohne Präfix und in Großbuchstaben), `Prüfreihenfolge von CreateWorkspace` (acht Fälle, darunter Kennung einer Baseline); Mutationen: jede Halbbedingung (Kennung, Service, Workspace, Name, Zeit, Blob, beide Hashes), `Validate` entfernt, Service-Prüfung entfernt, Workspace-Prüfung entfernt, Kennungsprüfung und Namensprüfung entfernt, Kennung vor Service (rot); Name vor Kennung ist äquivalent (beide `ErrBereitsVorhanden`) |
| Nr. 13: Name je Workspace eindeutig | `der Name ist je Workspace eindeutig` (gleicher Name in anderem Workspace, anderer Schreibweise); Mutation: Name global geprüft (rot) |
| Nr. 14: `GetWorkspace` und `ListWorkspace`, Art vor Workspace | `GetWorkspace bei Unbekanntem, Fremdem und Baselines` (auch mit unbekanntem und leerem Workspace), `ListWorkspace ordnet nach Zeit und Kennung`; Mutationen: Baseline als `NichtGefunden`, Baseline nur bei bekanntem Workspace, fremder Workspace nicht geprüft, absteigend, Kennung zuerst, nur Zeit, Kennung absteigend, alle Workspaces (rot) |
| Nr. 14: Baseline-Operationen liefern nie einen Workspace-Snapshot | `Baselines und Workspace-Snapshots bleiben getrennt`; Mutationen: `Get` und `List` der Baselines ohne Art-Filter (rot) |
| Nr. 15: `Check` meldet Fremdschlüssel- und Integritätsbefunde, sonst leer | `Check einer sauberen Datenbank ist leer`, `Check meldet eine Fremdschlüssel-Verletzung` (absichtlich verletzte Datenbank, Tabelle im Text), `Check meldet eine Verletzung eines CHECK-Constraints` (Befund aus `integrity_check`); Mutationen: ohne `foreign_key_check`, ohne `integrity_check`, `ok` als Befund, Befunde verworfen, Tabelle fehlt im Text (rot) |
| Nr. 16: `ReferencedBlobs` aus Request, Response, Snapshots jeder Art, Service-Metadaten; ohne Doppelte, aufsteigend | `ReferencedBlobs liefert jede Quelle einmal, aufsteigend` (je Quelle ein eigener Wert, ein Wert in mehreren Quellen), `ReferencedBlobs einer leeren Datenbank`, `TestRundlauf`; Mutationen: je eine der vier Quellen entfernt, `UNION ALL` bei der letzten Vereinigung, `DESC`, Präfix entfernt, Null nicht ausgefiltert (rot) |
| Nr. 17: siehe Kopf | siehe Kopf |
| Der Rundlauf bleibt über ein Wiederöffnen gültig | `TestRundlauf` (alle Repositories schreiben, Schließen, Wiederöffnen, jede Art liest sich zurück, `Check` leer, Referenzen); Mutationen: `consumed` nicht gespeichert, Tombstones nicht gespeichert, Revision nicht gespeichert, Version nicht gespeichert, `Check` meldet immer (rot im Rundlauf allein) |
| Reset und Restore setzen die Operationen in einer Transaktion zusammen | `Reset: ReplaceAll, ResetAll und Advance in einer Transaktion` (Fehler am Ende lässt State, Cursors und Revision unverändert), `Restore: Entity-State ersetzen, Cursors bleiben` (Cursors unverändert, Fehler lässt alles unverändert); Mutationen: Stores nicht auf der Transaktion, kein Rollback, Fehler von `f` verschluckt (rot) |

**Ungeprüft** (kein Fänger, in `SPEC-108` als solches geführt oder hier benannt): dass `entity_key` kanonisch ist, dass `state_hash` zum Blob passt, das Wachsen von `position` (Fachlogik des Replay); das Format der Befunde von `Check` über Quelle und Tabelle hinaus; ein `Advance` an der Grenze von `int64` (über die Ports nicht erreichbar); `WithWorkspaceTransaction` mit `f == nil` und die Nutzung der Stores nach dem Ende von `f`.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-baseline-persistenz](../done/slice-baseline-persistenz.md) und [slice-blob-ablage](../done/slice-blob-ablage.md) liegen in `done/`.

**Marke überschritten:** Die Summe aus `git diff --numstat 640a52f..72e6ad3` über die Pfade aus §3 liegt bei 3116 Zeilen (3102 Einfügungen, 14 Löschungen), davon rund 2100 Tests (`outboundtest` 1720, Adapter-Tests 260, Modell-Tests 150) und rund 1000 Produktcode (Modell 183, Ports 90, Adapter 710). Der Slice bleibt ungeteilt, weil Ports, Adapter und Vertragstests einzeln nichts liefern und die Zusammensetzungstests (Reset, Restore, Rundlauf) alle Teile brauchen. Möglicher Schnitt, falls er zurückgeführt würde: (1) Workspace und Workspace-Snapshots mit `zugriff`, (2) Entity und Cursor, (3) Wartung, Transaktion, Zusammensetzungen und Rundlauf; jeder Teil brauchte den vorigen.

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

- Randformen der Repositories (`AGENTS.md` §3.12): leerer, fehlender oder unbekannter Wert, doppelter Schlüssel, Abbruch mitten in der Transaktion, gleichzeitiger Zugriff — **Ausgang:** entfallen: entschieden in `spezifikation.md` (Produkt) vor dem ersten Code-Commit; zwei Lücken (leere Kennung bei Lesern und Mutatoren, `ReplaceAll` auf unbekanntem Workspace) fand der Implementer beim Lesen von `SPEC-108` und gab sie vor dem Code an den Architect zurück.
- Der Slice ist der größte der Welle und kann die Marke überschreiten — **Ausgang:** weiter offen: wird bei der Rückführung in §4 benannt und dann als Beobachtung geführt.
- Die Ordnung von `List` und `ListSet` der Entities und der Cursors ruht für den hinteren Schlüssel auf dem Primärschlüssel-Index; `ORDER BY` nennt ihn, ein Test kann das Weglassen nicht von der Index-Ordnung unterscheiden — **Ausgang:** weiter offen: wird als Beobachtung geführt, solange kein zweiter Adapter die Ordnung selbst erzeugen muss.
- Die Befunde von `Check` sind Texte ohne festes Format; die Wartung der Folgewelle liest sie vermutlich nur als Meldung — **Ausgang:** weiter offen: wird mit dem `MaintenanceUseCase` sichtbar.
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

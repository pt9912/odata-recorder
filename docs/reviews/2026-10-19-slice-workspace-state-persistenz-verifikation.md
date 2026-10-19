# Verifikations-Report: slice-workspace-state-persistenz — 2026-10-19

**Art:** Verifikation (Modul 11) — gegen DoD, Closure-Trigger, Spezifikation und Entscheidungen. Nicht Review, nicht Validierung.

**Gegenstand:** [slice-workspace-state-persistenz](../plan/planning/in-progress/slice-workspace-state-persistenz.md), Diff `640a52f..dc22ee0` (Vertrag `SPEC-108` davor; Umsetzung `72e6ad3`, Plan-Nachzug `51f9c90`, Behebung der Review-Befunde `dc22ee0`).
Bezug: [LH-FA-04](../../spec/lastenheft.md), [LH-FA-07](../../spec/lastenheft.md), [LH-FA-09](../../spec/lastenheft.md), [LH-FA-10](../../spec/lastenheft.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); Vertrag `SPEC-108`; Sicht `ARC-004`, `ARC-009`.
Vorgänger: [Review-Report](2026-10-19-slice-workspace-state-persistenz-review.md) (0 HIGH, 3 MEDIUM, 5 LOW, 5 INFO).

**Rolle:** Verifier · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

## Urteil: BESTANDEN MIT AUFLAGEN

Der Code erfüllt `SPEC-108` Kopf und Nr. 1 bis 17, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) (Treiber nur im Adapter; `make a-check` und `make lint` grün) und die Transaktionsgrenze der Sicht (`ARC-004`: fachliche Grenze `WithWorkspaceTransaction`, keine offene Verbindung im Port). `make gates` ist grün.
Alle drei MEDIUM des Reviews (F-1, F-2, F-3) sind behoben und mit neuen Fängern versehen; ihre Mutationen wurden selbst nachgefahren und werden aus dem benannten Grund rot. F-4 bis F-11 sind behoben oder ehrlich geführt; F-12 und F-13 liegen weiter beim Planner.
Kein HIGH, kein MEDIUM. Drei LOW und zwei INFO, keiner ändert Produktcode zwingend; der DoD-Punkt „Reset und Löschen in einer Transaktion über alle vier Repositories“ ist funktional erfüllt, aber für das **Löschen** ohne zusammengesetzten Test belegt (V-1).
Offen und erwartet: DoD-Häkchen, Closure-Notiz, Register, drei Paarungen; die Entscheidung über die überschrittene Marke und die drei Adressen ohne Nehmer (Planner).

**Anzahl:** HIGH 0 · MEDIUM 0 · LOW 3 · INFO 2.

## Läufe (Verifier selbst)

| Lauf | Ergebnis |
|---|---|
| `make gates` (HEAD `dc22ee0`, Arbeitsbaum sauber) | Exit 0: a-check 0 Befunde, d-check 93 Dateien 0 Befunde, baseline-verify, build, test (`gofmt`, `go vet`, Tests), lint, schema-check (alle Fälle `ok`, darunter *Workspace loeschen nimmt Entities und Cursor mit*, *Session loeschen nimmt ihre Cursor mit*), migration-check `ok`; `git status` danach sauber |
| Mutationen | rund 330 Läufe in Wegwerf-Kopien (`git archive HEAD` unter `/tmp/claude-1000/-Development-odata-recorder/4e17703b-9e29-4182-bc79-0efc3533bb9c/scratchpad/vw`, je Lauf eine eigene Kopie unter `…/scratchpad/pmw/`), Container `golang:1.27-alpine@sha256:8a5910f3…`, `CGO_ENABLED=0`, mit vorab geladenem Modul-Cache; je Lauf `gofmt -l` und `go vet` vor `go test -failfast`. Fehlbauten (ungenutzte Variable oder Import, nicht gefundene Stelle) verworfen und neu gefasst. Skripte: `…/scratchpad/pm.py`, `m_a.py` bis `m_e.py`; Arbeitsbaum des Repos unberührt |
| Explorativ (nicht committet) | `…/scratchpad/ex2/src/adapters/driven/sqlite/explore_v_test.go`: Löschen in `f` mit späterem Fehler (alle vier Tabellen unverändert), Löschen mit `nil` (alle leer, `Check` leer), Löschen, Neuanlage und `ReplaceAll` in einer Transaktion samt Wiederöffnen, Savepoint-Wiederverwendung (drei gescheiterte `Create`, danach `Create`), verschachtelte Transaktion und Schreiben außerhalb in `f` (je `database is locked (5)` nach 5 s, äußere Transaktion bestätigt): alle wie `SPEC-108` Nr. 1 |
| Stabilität | `TestTransaktionVertrag`, `TestWorkspaceVertrag`, `TestRundlauf` mit `-count=5`: grün |
| `-race` | nicht verfügbar (`CGO_ENABLED=0`, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)); Gleichzeitigkeit stützt sich auf `gleichzeitige Advance …`, `Transaktionen laufen nacheinander` und die Mutation `BEGIN IMMEDIATE` zu `BEGIN` (rot) |

## DoD und Closure-Trigger

| Punkt | Befund |
|---|---|
| 1 Modell und Ports für Workspace, Entity, Cursor, Workspace-Snapshots; Adapter auf `SPEC-004`; `WithWorkspaceTransaction` | erfüllt: `src/hexagon/model/workspace.go`, `ports/outbound/workspace_repository.go`, `baseline_repository.go`, `errors.go`; `adapters/driven/sqlite/{workspaces,entities,cursors,workspace_snapshots,maintenance,workspace_tx,zugriff}.go` |
| 2 Reset und Löschen in einer Transaktion über Workspace, Entity, Cursor, Snapshot; `StorageMaintenance` | erfüllt mit einer Lücke im Beleg (V-1): Reset ist als Ablauf getestet (`transaktionResetFall`: `ReplaceAll`, `ResetAll`, `Advance`; Fehler am Ende lässt State, Cursors, Revision unverändert), Restore ebenso (`transaktionRestoreFall`); `Delete` nimmt alle vier Tabellen atomar mit (`Delete nimmt Entities, Cursors und Snapshots mit`, `Delete ist atomar`, beide auch auf den Repositories der Transaktion); `Check` und `ReferencedBlobs` sind erfüllt |
| 3 Rundlauftest: schreiben, schließen, wieder öffnen, zurücklesen, `foreign_key_check`/`integrity_check` leer | erfüllt: `TestRundlauf` öffnet denselben Pfad neu, liest zurück (Zeilen siehe unten), `Check` ist nach dem Wiederöffnen leer und nicht `nil`; sieben Persistenz-Mutationen rot, nur gegen `^TestRundlauf$` gefahren |
| 4 `make gates` grün | erfüllt |
| 5 bis 10 Review-Report, Doku-Update, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | Review-Report liegt vor, Haken steht offen; Closure-Notiz (§7 „Wird bei der Closure geschrieben“), Register und Paarungen stehen erwartungsgemäß offen; Risiko-Ausgänge siehe unten |
| Closure-Trigger (§5): Gates, Rundlauf grün, eine Mutation je Zusage rot gesehen | Gates ja, Rundlauf ja, Mutation je Zusage ja (siehe Zusage · Fänger) |

## Plan-vs-Code-Diff (`640a52f..dc22ee0`)

| Plan §3 | Code | Abweichung |
|---|---|---|
| `src/hexagon/model/` | `workspace.go`, `model_test.go` | keine |
| `src/hexagon/ports/outbound/` | `workspace_repository.go` (Workspace, Entity, Cursor, `StorageMaintenance`, `WorkspaceStores`, `WorkspaceTransactor`), `baseline_repository.go` (`CreateWorkspace`, `GetWorkspace`, `ListWorkspace`), `errors.go` (`ErrRevisionVeraltet`, `ErrFremderService`, `ErrFalscheArt`) | keine |
| `src/adapters/driven/sqlite/` | die sieben Dateien oben, `repositories_test.go`, `workspaces_test.go`, `baselines.go` (Anpassung) | keine; `zugriff.schreibe` trägt seit `dc22ee0` einen Savepoint je Operation, im Plan §3 (Zeile Adapter) nur als „`zugriff` tragen dieselben Repositories … auf der Verbindung der Transaktion“ geführt, die Fänger-Tabelle nennt ihn (V-4) |
| `outboundtest/` | sechs neue Verträge, `in_transaktion.go`, `vertrag.go` (Hooks) | keine |
| `spec/spezifikation.md` | `SPEC-108` vor dem Code (`6485e99`, `640a52f`), Nr. 1 nach dem Review geschärft (`dc22ee0`) | keine |
| `spec/architecture.md` | `StorageMaintenance` nennt die Blob-Referenzen der Service-Metadaten (Zeile `StorageMaintenance` der Port-Tabelle) | keine; die Sicht nennt weder Slice noch ADR noch Welle (`AGENTS.md` §3.4) |
| nicht im Plan | Review-Report | harmlos |

§1-Ausschlüsse gewahrt: keine CRUD-Fachlogik, kein Replay-Matching, kein Cursor-Fortschritt, kein `MaintenanceUseCase`, keine Aufbewahrung. `.golangci.yml`, `.a-check.yml`, `Dockerfile`, `go.mod`, `Makefile` unberührt; kein `//nolint` (`AGENTS.md` §3.2, §3.6 gewahrt). Die Schichten halten: die neuen Adapterdateien importieren nur `model`, `ports/outbound` und Bibliotheken.

## Sicht gegen Code: Transaktionsgrenze und StorageMaintenance

- `ARC-004` (Port-Design): „Die Transaktionsgrenze ist fachlich (`WithWorkspaceTransaction` …), nicht als offene SQL-Connection exponiert.“ Erfüllt: `WorkspaceTransactor.WithWorkspaceTransaction(ctx, func(WorkspaceStores) error)`; `WorkspaceStores` trägt nur Port-Interfaces; die Verbindung (`*sql.Conn`) bleibt im Adapter (`zugriff`).
- „Reset (Entity-State und Cursor) und Löschen eines Workspace laufen in einer solchen Transaktion über `WorkspaceRepository`, `EntityRepository`, `CursorRepository` und `SnapshotRepository`.“ Die vier Repositories liegen alle in `WorkspaceStores`; Reset und Restore sind als Zusammensetzung getestet, das Löschen als Einzeloperation über vier Tabellen und in der Transaktion (V-1 zum fehlenden Zusammensetzungstest).
- `StorageMaintenance` liefert `foreign_key_check`, `integrity_check` und die Referenzmenge aus Exchanges (Request, Response), Snapshots jeder Art und Service-Metadaten: erfüllt, vier Quellen je einzeln mutiert und rot.
- Die Wartung selbst (Abgleich der Blobs) bleibt dem `MaintenanceUseCase`, ausdrücklich außerhalb des Slice.

## Review-Befunde: Nachweis der Behebung

| Review | Zustand |
|---|---|
| F-1 (`ReplaceAll`, Reihenfolge je Art von ungültig) | behoben: Fall `ReplaceAll prüft jede Art von ungültig vor Kontext und Workspace` (`entity_vertrag.go`, `entityFall9`). Selbst nachgefahren, sechs Mutationen (Kontext vor jeder der drei Arten; jede der drei Arten erst nach der Workspace-Prüfung) und „alle Gültigkeit nach dem Kontext“: alle rot, die ersten vier aus dem neuen Fall, die Arten „ungültiger Eintrag“ aus `ReplaceAll prüft in der Reihenfolge der Spezifikation` |
| F-2 (`Put` in Gegenrichtung) | behoben: `Put ersetzt auch einen Tombstone und senkt die Version` (`entityFall10`). Sechs Mutationen des `ON CONFLICT … DO UPDATE` (Tombstone bleibt, Version `max`, Version nur beim Wiederbeleben `max`, leere Properties überschreiben nicht, Vollständigkeit `partial` bleibt, Payload bleibt bei Tombstone): alle rot |
| F-3 (Atomarität einer Operation in `f`) | behoben und in `SPEC-108` Nr. 1 entschieden: Savepoint je schreibender Operation in `zugriff.schreibe`; Fänger `eine gescheiterte Operation nimmt ihre Teilwirkung zurück, auch wenn f den Fehler verschluckt` (`transaktionTeilwirkung`: `Delete` nach dem Löschen der Snapshots, `ReplaceAll` nach dem Löschen; eine gelungene Operation davor bleibt). Selbst gefahren: `ROLLBACK TO` entfernt, `ROLLBACK TO` durch `RELEASE` ersetzt, `ROLLBACK TO` mit dem abgebrochenen Kontext, `ROLLBACK TO` durch `ROLLBACK` (gelungene Operation davor ginge verloren), Fehler der Operation verschluckt: alle rot. Der Kommentar von `schreibe` nennt den Savepoint zutreffend |
| F-4 (Äquivalenz „Session vor Workspace“) | behoben: die Plan-Zeile führt „Session vor Workspace“ jetzt als äquivalent mit Begründung; selbst gefahren, grün wie behauptet |
| F-5 (Spalte, Zahl der Operationen) | behoben: der Satz zur „letzten Spalte“ ist entfallen, die Zahl lautet 20; alle 20 Operationen mit `ctx = context.Background()` selbst gefahren, alle rot |
| F-6 (Zählerstände im Register) | behoben: §8 nennt 6, 2 und 8 Evidence-Dateien, wie `ls …/evidence` (6, 2, 8). Offen bleibt die Verkörperung von `BEO-REPO/randform-erst-im-review-entschieden` (Stand offen bei 6 Dateien, über der Schwelle; `AGENTS.md` §3.12 trägt keinen Herkunfts-Anker): Sache der Welle-Closure, nicht dieses Slice |
| F-7 (Kommentar `TestRundlauf`) | behoben: der Kommentar nennt jetzt, was geprüft wird; `state_hash` und `metadata_hash` des Workspace-Snapshots werden verglichen. Selbst gefahren, beide Konstant-Mutationen im Rundlauf allein rot |
| F-8 (`f == nil`) | behoben und in `SPEC-108` Nr. 1 entschieden (`ErrUngueltig`, vor dem Kontext); Fänger `f ohne Wert ist ungültig`. Prüfung entfernt, Prüfung nach dem Kontext, `nil` statt Fehler: alle rot |
| F-9 (Wartezeit auf die Sperre) | in `SPEC-108` Nr. 1 entschieden („Nicht unterstützt“, Fehler der Datenbank nach `busy_timeout`, kein Hängen); im Plan unter *Ungeprüft* geführt. Verhalten selbst bestätigt (10 s für zwei Zugriffe, äußere Transaktion bestätigt) |
| F-10 (NUL, ungültiges UTF-8) | in `SPEC-108` unter *Ungeprüft* aufgenommen und im Plan geführt: ehrlich |
| F-11 (Ordnung ohne Fänger) | im Plan §6 auf `Entities`, `Cursors` und `ReferencedBlobs` erweitert, weiter offen; Äquivalenz selbst bestätigt (siehe unten) |
| F-12 (Marke) | im Plan benannt, Entscheidung beim Planner (V-2) |
| F-13 (Adressen ohne Nehmer) | im Plan §1 ehrlich als „keine Adresse, geht an den Planner“ geführt; weiter offen beim Planner |

## Zusage · Fänger: Prüfung der Tabelle (Plan §3) gegen `SPEC-108`

Jede Nummer und Randform von `SPEC-108` hat eine Zeile; alle benannten Teiltests und Konstruktionen lösen auf (154 Kennzeichnungen im Tabellenblock gegen die Quelltexte gelesen, nur drei Code-Fragmente ohne Teiltest-Bezug). „rot“ nennt den Teiltest, aus dem die Mutation rot wurde (Meldung geprüft); eigene Mutationen, nicht die des Reviews wiederholt, wo die Bauart es zuließ.

| Nummer / Randform | Eigene Mutationen | Ergebnis |
|---|---|---|
| Kopf: Argument vor Kontext, Nr. 17 | `ctx = context.Background()` am Anfang jeder der 20 Operationen | alle rot (`abgebrochener Kontext …` je Vertrag; `Delete` zusätzlich in `Delete ist atomar`, `ReplaceAll` in `… Reihenfolge der Spezifikation`, `WithWorkspaceTransaction` in `ein abgebrochener Kontext rollt zurück`) |
| Kopf: leere Kennung | `ErrUngueltig` bei `Workspaces.Get`, `Advance`, `Delete`, `Entities.Get`, `List`, `ReplaceAll`, `Cursors.Get`, `List`, `ResetAll`, `GetWorkspace`, `ListWorkspace` (11) | alle rot |
| Kopf: Listen nie `nil` | `var out []T` in `sammle` | rot (`Create und Get tragen alle Felder`) |
| Kopf: Zeiten UTC, Jahr 1 bis 9999 | Schreiben ohne `UTC()`, Lesen mit `Local()`, Obergrenze 10000, Untergrenze 0 | alle rot (die Jahresgrenzen zuerst in den gemeinsamen Session-Tests; die Workspace-Fälle Jahr 0 und 10000 stehen zusätzlich in `Create lehnt Ungültiges ab …`) |
| Nr. 1: `nil` bestätigt, Stores auf der Transaktion | je Store (`Workspaces`, `Entities`, `Cursors`, `Snapshots`) ohne Verbindung | alle vier rot (`nil bestätigt alles`) |
| Nr. 1: Fehler, Panik, Kontext, nacheinander | Rollback entfernt; Fehler verschluckt; `COMMIT` mit `WithoutCancel`; `BEGIN IMMEDIATE` zu `BEGIN` | alle rot (`Transaktion` und gemeinsame Tests; Rollback entfernt baut nur mit `&& false`, dann rot) |
| Nr. 1: Atomarität je Operation (neu) | `ROLLBACK TO` entfernt / durch `RELEASE` / mit Kontext / durch `ROLLBACK`; Fehler verschluckt | alle rot |
| Nr. 1: `RELEASE` mit dem Kontext der Operation statt `WithoutCancel` | `RELEASE operation` auf `ctx` | **grün** (V-3) |
| Nr. 1: `f == nil` (neu) | Prüfung entfernt; Prüfung nur bei lebendem Kontext; Rückgabe `nil` | alle rot (`f ohne Wert ist ungültig`) |
| Nr. 2: `Create` | Revision `<0`, Kennung, Service, Baseline, Zeit je einzeln aus `Validate`; Revision 1 geschrieben; Service vor Kennung; Baseline vor Service; Fremd vor Art; Art, Fremd, Service-Prüfung, Kennungsprüfung entfernt | alle rot (`TestWorkspaceValidate`, `Prüfreihenfolge von Create`, `Create und Get …`) |
| Nr. 3: `Get`, `List` | `ORDER BY` (Kennung zuerst, Zeit absteigend, Kennung absteigend, ganz entfernt), ohne Service-Filter, ohne Leerprüfung | alle rot (`List ordnet nach Zeit, nicht nach Text`, `Get und List`) |
| Nr. 4: `Advance` | `>` statt `!=`, `<` statt `!=`, `+2`, Negativgrenze `-1`, Update für alle Workspaces | alle rot (`Advance zählt und lehnt Veraltetes ab`); Gleichzeitigkeit über `BEGIN`-Mutation rot |
| Nr. 5: `Delete` | Snapshot-Löschen ohne Wirkung / für alle Workspace-Snapshots; Existenzprüfung entfernt; `CASCADE` bei `entities` entfernt; beide `CASCADE` bei `replay_cursors.workspace_id` entfernt | alle rot (`Delete nimmt Entities, Cursors und Snapshots mit`); nur eines der beiden `CASCADE` von `replay_cursors.recording_session_id` entfernt: grün, äquivalent (zwei Pfade), beide entfernt: rot |
| Nr. 6: `EntityState` | drei Kennungen, Payload, Vollständigkeit (ganz, nur `complete`), Version (`<=0`, aus), Property (leer, UTF-8, doppelt) einzeln; `Validate` in `Put` entfernt | alle rot (`TestEntityStateValidate`, `Put lehnt Ungültiges ab`) |
| Nr. 7: `Put`, `Get` | je Spalte beim Ersetzen nicht übernommen (fünf); Workspace-Prüfung entfernt; Tombstone in `Get` und `List` ausgeblendet; Properties unsortiert | alle rot |
| Nr. 7 Gegenrichtung (neu) | siehe F-2 oben | alle rot |
| Nr. 8: `List`, `ListSet` | Reihenfolge Schlüssel/Menge vertauscht, Menge absteigend, Schlüssel absteigend; ohne Workspace-Filter; ohne Mengenfilter; Leerprüfung; Prüfung des unbekannten Workspace | alle rot; `ORDER BY` ganz entfernt: grün, äquivalent (Primärschlüssel-Index) wie im Plan |
| Nr. 9: `ReplaceAll` | ohne Löschen; Löschen aller Workspaces; `Validate`, Workspace-Vergleich, Doppelprüfung, Workspace-Prüfung entfernt; Doppelprüfung nur über den Schlüssel | alle rot |
| Nr. 9: Reihenfolge (neu) | siehe F-1 oben; Workspace-Prüfung vor dem Kontext | rot bzw. grün-äquivalent (Abfrage nutzt den Kontext), wie im Plan |
| Nr. 10: `ReplayCursor` | drei Kennungen, Service, Strategie ganz und je Wert, Position (`>`, `<`), `consumed` (`<0`, doppelt), unsortiert gespeichert | alle rot (`TestReplayCursorValidate`, `TestConsumedJSON`) |
| Nr. 11: `Put` | Service nur Workspace / nur Session / ohne; `NichtGefunden` der Session entfernt; Workspace-Prüfung entfernt; je Spalte beim Ersetzen nicht übernommen (drei); **Session vor Workspace**; `Get` ohne Recording-Filter | alle rot außer „Session vor Workspace“: grün, äquivalent wie im Plan (beide `ErrNichtGefunden`) |
| Nr. 11: `List` | `DESC`, Recording zuerst, Recording absteigend, Prüfung des Workspace; `ORDER BY` ganz entfernt | rot; letzteres grün, äquivalent |
| Nr. 12: `ResetAll` | Position bleibt; `consumed` bleibt; Strategie geändert; alle Workspaces; Workspace-Prüfung entfernt; beide `CASCADE` auf `sessions` entfernt | alle rot |
| Nr. 13: `CreateWorkspace` | vier Kennungen, Zeit, Blob, beide Hashes einzeln; Workspace-Prüfung, Service-Prüfung, Kennungsprüfung, Namensprüfung entfernt; Name global geprüft; **Name vor Kennung** | alle rot außer „Name vor Kennung“: grün, äquivalent wie im Plan (beide `ErrBereitsVorhanden`) |
| Nr. 14: `GetWorkspace`, `ListWorkspace` | Baseline als `NichtGefunden`; Workspace-Vergleich entfernt; Workspace vor Art; `DESC`, Kennung zuerst, ohne `ORDER BY`, alle Workspaces, Prüfung des unbekannten Workspace entfernt; Baseline-Leser ohne Art-Filter | alle rot |
| Nr. 15: `Check` | ohne `foreign_key_check`; ohne `integrity_check`; `ok` als Befund; Befunde verworfen; Tabelle fehlt; immer ein Befund | alle rot |
| Nr. 16: `ReferencedBlobs` | je Quelle entfernt (vier); nur Baselines; `UNION ALL` bei der letzten Vereinigung; `DESC`; `NULL` bei den Metadaten nicht ausgefiltert | alle rot; `UNION ALL` bei der ersten Vereinigung und `ORDER BY 1` entfernt: grün, äquivalent (`UNION` dedupliziert und sortiert) |
| Rundlauf nach Wiederöffnen | Hashes des Workspace-Snapshots konstant (beide); `consumed` konstant `[]`; Tombstones nicht gespeichert; Version 0; Revision `* 0`; `Check` meldet immer | alle sieben rot, gefahren gegen `^TestRundlauf$` allein |
| Reset und Restore in einer Transaktion | Stores ohne Verbindung; Rollback entfernt; Fehler verschluckt | rot (siehe Nr. 1) |

Die als äquivalent geführten Mutanten (`ORDER BY` bei `Entities`, `Cursors`, `ReferencedBlobs`; Workspace-Prüfung vor dem Kontext; Session vor Workspace; Name vor Kennung; ein `CASCADE` von zweien; `UNION ALL` vor der letzten `UNION`) sind äquivalent: selbst gefahren, alle grün, und die Begründung trägt jeweils (Index-Ordnung, doppelte Pfade, `UNION`, beide Fehler gleich).

## Rundlauf: was nach dem Wiederöffnen gelesen wird

`TestRundlauf` (`workspaces_test.go`) schreibt Service, Session, Exchange (Header, Blob-Referenzen), Baseline mit Import und Coverage, Workspace (Revision 1 nach `Advance`), zwei Entities (eines Tombstone), Cursor (`consumed` `1,4`), Workspace-Snapshot; schließt `DB`, öffnet denselben Pfad; liest Service, Session, Exchange (Header, Response-Blob), Baseline, Import, Coverage, Workspace (Revision, Zeit, Baseline), `Entities.List` und `Get` des Tombstone, Cursor (ganze Struktur), Workspace-Snapshot samt beiden Hashes, `Check` (leer, nicht `nil`) und `ReferencedBlobs` (fünf Referenzen). Der Kommentar nennt genau das; er sagt nicht „alles“ (§3.11 erfüllt). Nicht zurückgelesen und im Plan so benannt: `Workspaces.List`, `Cursors.List`, `ListWorkspace`, `Sessions.List`. Nicht Teil des Rundlaufs: Schreiben über `WithWorkspaceTransaction` (die Transaktion ist durch den Vertrag belegt; mein explorativer Test hat Schreiben in der Transaktion samt Wiederöffnen bestätigt).

## Kommentare gegen Tests (`AGENTS.md` §3.7, §3.11)

Geprüft: Köpfe der sechs Vertragsläufe (`RunWorkspaceVertrag` „Nr. 2 bis 5 und Nr. 17“, `RunEntityVertrag` „Nr. 6 bis 9 und Nr. 17“, `RunCursorVertrag` „Nr. 10 bis 12 und Nr. 17“, `RunWorkspaceSnapshotVertrag` „Nr. 13, 14 und 17“, `RunWartungVertrag` „Nr. 15 bis 17“, `RunTransaktionVertrag` „Nr. 1 und die Abläufe Reset und Restore“), `schreibe`, `WithWorkspaceTransaction`, `TestRundlauf`, Kommentare der Ports. Keine Zusage ohne Fänger, kein Konjunktiv über Verworfenes, keine Chronik. Die Köpfe nennen weniger, als die Läufe prüfen (die Querschnitts-Zusagen „leere Kennung“, „Listen nie `nil`“, „Zeiten UTC“ aus dem Kopf von `SPEC-108` fehlen in der Aufzählung): keine Überdehnung, siehe V-5.

## Risiken und Adressen (Plan §6, `AGENTS.md` §3.13)

| Risiko / Adresse | Ausgang | Urteil |
|---|---|---|
| Randformen der Repositories | entfallen: entschieden in `SPEC-108` vor dem Code; zwei Lücken gab der Implementer vor dem Code zurück, drei fand das Review und `SPEC-108` Nr. 1 führt sie | ehrlich; Nr. 1 enthält jetzt Atomarität, `f == nil`, Sperre |
| Größter Slice der Welle, Marke | weiter offen; Entscheidung beim Planner (§4) | siehe V-2 |
| Ordnung ruht auf Index / `UNION` | weiter offen, Beobachtung | Äquivalenz bestätigt |
| `Check`-Befunde ohne festes Format | weiter offen, mit dem `MaintenanceUseCase` | angenommen als offen |
| Cursor-Zustand wächst mit dem Replay | weiter offen, mit welle-record-replay | angenommen als offen |
| §1: CRUD-Logik, Matching, Cursor-Fortschritt, `MaintenanceUseCase`, Aufbewahrung | „keine Adresse, geht an den Planner“ | ehrlich geführt. Die Roadmap-Zeile von welle-record-replay nennt Matching und Cursor, aber weder CRUD (`LH-FA-07`) noch Wartung noch Aufbewahrung; ohne Entscheidung des Planners kann der Slice nicht sauber nach `done/` (V-6) |
| Zusagen der Vorgänger (Workspace-Snapshots, `StorageMaintenance`) | in §1/DoD angenommen | erfüllt |

Ein Slice mit „weiter offen“ geht nur nach `done/`, wenn der Planner die Ausgänge annimmt (Modul 5); das ist keine Verifier-Entscheidung.

## Findings

| ID | Stufe | Befund | Quelle | Pfad | Verifizierbar |
|---|---|---|---|---|---|
| V-1 | LOW | Der DoD-Punkt 2 und `ARC-004` nennen „Reset **und Löschen** eines Workspace in einer Transaktion über Workspace, Entity, Cursor und Snapshot“. Für Reset und Restore gibt es zusammengesetzte Tests (`transaktionResetFall`, `transaktionRestoreFall`), für das Löschen keinen: kein Test ruft `Workspaces.Delete` in `f` und lässt `f` danach scheitern oder prüft, dass die Löschung über alle vier Tabellen mit dem Fehler zurückrollt. Das Verhalten ist richtig (selbst gefahren: Löschen mit späterem Fehler lässt Workspace, Entities, Cursors, Snapshot unverändert; Löschen mit `nil` leert alle vier, `Check` leer); die Einzelzusagen (`Delete` atomar, `Delete` in `f` bestätigt) sind gefangen. Fehlt der Fänger der Zusammensetzung, würde eine Mutation, die das Löschen aus der äußeren Transaktion herauslöst (zweite Verbindung), nur über die 5 s `busy_timeout` rot, nicht über einen benannten Test. | `AGENTS.md` §3.15 · DoD Punkt 2 · `ARC-004` | `src/hexagon/ports/outbound/outboundtest/transaktion_vertrag.go` · `RunTransaktionVertrag` (neben `transaktionResetFall`, Zeile 222) | ja, explorativer Test (selbst gefahren) |
| V-2 | LOW | Die Marke in §4 nennt 3116 Zeilen aus `git diff --numstat 640a52f..72e6ad3`; `AGENTS.md` §3.14 verlangt die Metrik bis zum letzten Commit des Slice. Bis `dc22ee0` gemessen: 3218 (3204 Einfügungen, 14 Löschungen) über `src/`, 3225 mit den Spec-Pfaden aus §3. Die Marke ist so oder so überschritten (900); der Plan nennt Basis und Endpunkt ausdrücklich, die Zahl ist also eine Zwischenmessung und kein falscher Wert. Die Entscheidung über das Weiterführen (Zeile 149) liegt beim Planner. | `AGENTS.md` §3.14 | `docs/plan/planning/in-progress/slice-workspace-state-persistenz.md` · Zeile 147 | ja, `git diff --numstat` |
| V-3 | LOW | `zugriff.schreibe` löst den Savepoint mit `context.WithoutCancel(ctx)` (`ohneAbbruch`) auf; die Mutation `RELEASE` auf `ctx` bleibt grün (selbst gefahren, gegen die ganze Menge der `Transaktion`-Tests und alle Pakete). Wirkung der Mutation: bricht der Kontext nach der letzten Anweisung einer gelungenen Operation ab, meldet sie einen Fehler und behält ihre Wirkung, wenn `f` ihn verschluckt, also eine Teilwirkung bei Fehler, was `SPEC-108` Nr. 1 ausschließt. Der Code ist richtig; die Entscheidung, dass der Abschluss einer gelungenen Operation den Abbruch nicht sieht, hat keinen Fänger und der Plan führt sie nicht in der Tabelle (die Zeile „jede Operation bleibt in `f` atomar“ nennt nur den Rollback-Zweig). Verwandt, äquivalent: der `SAVEPOINT` mit `context.Background()` (jede folgende Anweisung nutzt `ctx`). | `AGENTS.md` §3.15 | `src/adapters/driven/sqlite/zugriff.go` · Zeile 73 (`_, err := z.conn.ExecContext(ohneAbbruch, "RELEASE operation")`) | ja, Mutation grün (selbst gefahren) |
| V-4 | INFO | §3 des Plans (Zeile „Adapter“) und Plan-Kopf führen den Savepoint nicht als Bestandteil der Adapter-Änderung; er steht nur in der Fänger-Tabelle (Zeile „Nr. 1: jede Operation bleibt in `f` atomar“) und im Kommentar von `schreibe`. `AGENTS.md` §3.9 verlangt, dass §3 der Korrektur folgt: die Pfade stimmen, die Beschreibung nicht vollständig. | `AGENTS.md` §3.9 | `docs/plan/planning/in-progress/slice-workspace-state-persistenz.md` · §3, Zeile `src/adapters/driven/sqlite/` | ja |
| V-5 | INFO | Die Köpfe der Vertragsläufe nennen nur ihre Nummern, prüfen aber auch die Kopf-Zusagen (leere Kennung, Listen nie `nil`, Zeiten); dieselbe Klasse wie V-2 im Report zu `slice-baseline-persistenz`. Der Leser eines zweiten Adapters liest weniger, als der Lauf fordert. | `AGENTS.md` §3.7, §3.11 | `src/hexagon/ports/outbound/outboundtest/workspace_vertrag.go` · Zeile 15 (und `entity_vertrag.go` 13, `cursor_vertrag.go` 13) | ja |
| V-6 | INFO | Adressen ohne Nehmer bleiben (`AGENTS.md` §3.13): CRUD-Fachlogik (`LH-FA-07`), `MaintenanceUseCase`, Aufbewahrung; ebenso die Entscheidung über die Marke. Der Plan führt sie als Übergabe an den Planner, nicht als Zuweisung. Ehrlich, aber vor `done/` zu entscheiden; ohne Nehmer bliebe `LH-FA-07` für die Fachlogik ungeschlossen (die Ablage dafür liefert dieser Slice). | `AGENTS.md` §3.13 · `LH-FA-07` | `docs/plan/planning/in-progress/slice-workspace-state-persistenz.md` · §1 (Zeilen 36 bis 42) | nein |

## Offen und erwartet (kein Befund)

Häkchen der DoD, Closure-Notiz mit Steering-Loop-Lerneintrag (Klassen dieses Slice: „Prüfreihenfolge je Paar ohne Fänger“, „Zusage nur in einer Richtung gefangen“, „Randform im Code entschieden statt vor dem Code“ treten erneut auf; die Beobachtung `BEO-REPO/randform-erst-im-review-entschieden` steht bei 6 Dateien auf `Stand: offen`, die Verkörperung samt Herkunfts-Anker folgt der Welle-Closure), Beobachtungs-Register, drei Paarungen, Häkchen für den Review-Report.

## Übergabe

Der Implementer ergänzt optional V-1 (ein Fall in `RunTransaktionVertrag`: `Workspaces.Delete` in `f`, danach Fehler, vier Tabellen unverändert; danach `nil`, vier Tabellen leer) und V-3 (ein Fall, in dem der Kontext nach der letzten Anweisung einer gelungenen Operation abbricht und `f` den Fehler verschluckt, oder der Satz im Plan, dass der Abschluss den Abbruch bewusst nicht sieht); V-4 und V-5 sind Nachzug in Plan und Köpfen. V-2 und V-6 gehen an den Planner (Entscheidung über das Weiterführen, Adressen der drei Ausschlüsse, Annahme der Ausgänge „weiter offen“). Danach genügt eine erneute Verifikation der Zeilen „Nr. 1: Atomarität“ und „Reset und Restore“.

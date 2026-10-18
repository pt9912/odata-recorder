# Spezifikation — OData Recorder

**Bezug zum Lastenheft:** `spec/lastenheft.md` (LH-FA-01 bis LH-FA-18, LH-QA-01 bis LH-QA-07, LH-RB-01 bis LH-RB-04). Bei Konflikten gilt das Lastenheft. Diese technische Spezifikation konkretisiert, erweitert aber nicht den vereinbarten Umfang.

**Rolle:** Technisch verbindliches, fortschreibbares Spec-Stratum nach `ai-harness-course`. Priorität **Muss** ist implementierungsbindend; **Soll/Kann** sind technisch vorbereitet, aber keine Abnahmevoraussetzung der ersten Version. Technische Defaults und Schwellenwerte sind bis zur Freigabe Vorschläge.

---

## 1. Algorithmen und Datenflüsse

### LH-FA-01.a — Transparenter HTTP-Proxy

**Eingabe:** HTTP-Request an konfigurierten Service-Root; **Ausgabe:** HTTP-Response des Upstreams.

1. Gateway ermittelt Service und Betriebsart, validiert Zielkonfiguration und erstellt Korrelations-ID.
2. Proxy leitet Methode, Pfad, Query, Body und End-to-End-Header weiter; Hop-by-Hop-Header werden nach HTTP-Regeln behandelt.
3. Upstream-Response wird an den Client gestreamt. Weiterleitungen, absolute OData-Links und Cookies werden nur gemäß expliziter Rewrite-Policy angepasst.
4. HTTP-Transport-Encoding und komprimierte Bodies werden im transparenten Pfad nicht unnötig transformiert; semantische Analyse erfolgt auf einer getrennten, begrenzten Capture-Darstellung.

**Komplexität:** O(n) über transportierte Bytes; Streaming-Speicher O(buffer). **Fehlermodi:** Timeout, TLS-Fehler, Upstream nicht erreichbar, Client-Abbruch → definierter Gateway-Fehler, soweit noch keine Response gesendet wurde.

### LH-FA-02.a — Capture und atomarer Abschluss

1. Session mit monotoner `sequence_no` erzeugen; Beginn und Ende jedes Austauschs erfassen.
2. Request- und Response-Metadaten sammeln; Body-Streams mit begrenzten Puffern in Blob-Storage schreiben, während die Weiterleitung fortgesetzt wird.
3. Sensitive Header und konfigurierte JSON-Pfade **vor dauerhafter Ablage** redigieren; Redaction-Status dokumentieren.
4. Nach vollständig abgeschlossener Antwort Metadaten und Blob-Referenzen atomar committen; unvollständige Exchanges gesondert markieren.
5. Speicherfehler dürfen nicht als erfolgreiches Recording ausgegeben werden; eine bereits laufende Proxy-Response kann nicht rückwirkend verändert werden.

**Fehlermodi:** Speicher voll, Abbruch, nicht redigierbarer Body, ungültige Content-Encoding. Default: sicherheitsrelevanter Capture-Fehler verwirft den sensiblen Body und meldet `ODR-E2401`.

### LH-FA-03.a — Recording laden und prüfen

- Versionskennung, Session- und Exchange-IDs, referenzierte Blobs, Checksummen und Reihenfolge prüfen.
- Fehlende oder korrupte Blob-Datei ergibt `ODR-E3001`; kein stiller Ersatz durch leeren Body.
- Schreibvorgänge: Blob temporär schreiben → fsync/rename → SQLite-Transaktion referenzieren. Nicht referenzierte Blobs dürfen später bereinigt werden.
- Recordings sind nach Abschluss logisch unveränderlich; neue Sessions werden append-only ergänzt.

### LH-FA-04.a — Deterministischer Replay-Cursor

**Eingabe:** Workspace, ausgewählte Session, Request, Replay-Policy. **Ausgabe:** gespeicherte Response oder Mismatch.

1. Match-Key nach LH-FA-05.a erzeugen.
2. Im Modus `sequential` die nächste noch nicht verbrauchte passende Exchange **innerhalb des dem Testlauf zugeordneten Cursors** suchen.
3. Response inklusive Status und ausgewählten End-to-End-Headern ausliefern; zeit- oder hostabhängige Header nur nach expliziter Policy transformieren.
4. Cursor erst nach erfolgreicher Auswahl atomar fortschreiben; bei gleichzeitigem Zugriff pro Cursor serialisieren.
5. `strict` vergleicht ausschließlich die nächste erwartete Exchange; `first` und `last` verändern keinen Cursor.

**Fehlermodi:** kein Treffer → `ODR-E5001`; Recording erschöpft → `ODR-E5002`. Replay hat **keinen** Upstream-Zugriff.

### LH-FA-04.b — Explizite Zuordnung paralleler Replay-Clients

- Ein Replay-Request trägt einen vom Test-Harness gesetzten Workspace-Identifier und eine Recording-Session-Zuordnung. Die Transportabbildung erfolgt über `X-Recorder-Workspace`, `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session`; diese Steuerheader werden niemals an einen Upstream weitergereicht.
- Standard ist `assignment=explicit`. Fehlt eine Zuordnung, ist sie unbekannt oder mehrdeutig, wird `ODR-E5003` geliefert. Ein implizites `ordered`-Assignment ist ausschließlich als opt-in für seriell ausgeführte Tests zulässig.
- Cursor-Identität: `(workspace_id, replay_session_id, recording_session_id)`. Zwei parallele Clients teilen nur dann absichtlich einen Cursor, wenn sie dieselbe explizite Replay-Session-ID verwenden; Cursor-Zugriffe sind atomar.
- Replay-Session-IDs werden beim Start eines Testlaufs vergeben. Das Löschen eines Workspaces entfernt dessen Cursor. Ein Workspace-Reset setzt sämtliche Cursor auf Position 0.

### LH-FA-05.a — Kanonisches HTTP-Matching

Match-Key = SHA-256 über eine eindeutig kodierte Tupelfolge:
`method`, `service_id`, `normalized_path`, sortierte Multimenge dekodierter Query-Paare, ausgewählte Vary-Header, Body-Hash.

- Query-Paare werden stabil nach Schlüssel und Wert sortiert; wiederholte Parameter bleiben erhalten.
- Prozentkodierung wird nur soweit normalisiert, wie dekodierte Zeichenfolgen identisch sind; `+`-Semantik ist Query-kontextabhängig.
- Pfadsegmentgrenzen, Groß-/Kleinschreibung und OData-Literale bleiben erhalten.
- Bei JSON-Body kann eine opt-in JSON-Kanonisierung Schlüsselreihenfolge normalisieren; Default ist Byte-Hash.
- Authentifizierungstokens und volatile Header werden nicht als Match-Key benutzt; fachlich relevante Header (`Accept`, `OData-Version`, `If-Match`, `Prefer`) werden nach Policy berücksichtigt.
- Eine semantische Gleichsetzung von `$filter`-ASTs ist **nicht** Bestandteil des Muss-Umfangs.

### LH-FA-06.a — OData-V4-Metadatenmodell

1. `$metadata` im Record-Modus als HTTP-Austausch erfassen.
2. EDMX/CSDL-XML (V4) parsen und in internes Modell überführen: Schemas, EntityTypes, Property-Typen, Nullable, Keys, EntitySets, NavigationProperties.
3. Referenzen und Vererbung auflösen; nicht unterstützte Typen/Annotations explizit markieren.
4. Mock liefert die gespeicherten Metadaten konsistent mit der ausgewählten Service-Version aus.
5. Ohne gültiges Modell keine zustandsbehaftete Verarbeitung dieses EntitySets; reines Replay bleibt möglich.

### LH-FA-18.a — Expliziter Import in eine Mock-Baseline

**Eingabe:** Service-Modell, ausgewählte vollständige Recordings/Exchanges, Import-Policy. **Ausgabe:** unveränderliche Snapshot-ID, Coverage-Evidenz, Import-Report.

1. Session und Metadatenversion fixieren; nur erfolgreich abgeschlossene, redaktionsverträgliche OData-V4-JSON-Responses für einzelne Entities und Collections berücksichtigen.
2. `@odata.context`, EntitySet, Key-Typen und Property-Typen anhand des Modells validieren. `@odata.nextLink` markiert unvollständige Pagination, solange die komplette Kette nicht belegbar erfasst ist. `$top`, `$skip` und Filter ohne Vollständigkeitsbeweis erzeugen keine vollständige Coverage.
3. Jedes Entity in typisierten kanonischen Key überführen. Einzel-Entity- und Collection-Responses werden als beobachtete Fakten behandelt; `null`-Properties bleiben von nicht selektierten Properties unterscheidbar.
4. Pro Key deterministisch deduplizieren: identische kanonische Payloads zusammenführen; widersprüchliche vollständige Werte oder inkompatible Typen sind Konflikte und stoppen den Import. Teilprojektionen dürfen nur widerspruchsfrei ergänzt werden; fehlende Felder sind kein `null`.
5. Provenienz pro Entity speichern (`session_id`, `exchange_id`, EntitySet, Metadatenhash, Importzeit) und Coverage-Evidenz pro EntitySet aggregieren.
6. Vollständiges Snapshot-Manifest in temporärem Blob ablegen und nach Integritätsprüfung in einer SQLite-Transaktion als neue Baseline publizieren. Bei Fehler bleibt die bisherige Baseline unverändert. Import ist explizit, kein Record-Nebeneffekt.

**Fehlermodi:** ungültiger Key, widersprüchliche Werte, beschädigte Antwort, fehlendes Metadatenmodell, redigierte notwendige Key-Felder → `ODR-E6101`. **Komplexität:** O(n) Entitäten plus Indexkosten; Import speichert Zwischenstände begrenzt bzw. in Staging-Tabellen.

### LH-FA-07.a — Stateful CRUD

- `GET` liest aus Workspace-Snapshot plus lokalen Änderungen.
- `POST` validiert EntitySet, Key und Property-Typen, erzeugt einen Datensatz und liefert die vereinbarte OData-V4-Antwort.
- `PATCH` verändert ausschließlich übergebene Properties; `PUT` ersetzt nach der unterstützten OData-V4-Semantik.
- `DELETE` entfernt die Entity aus der Sicht des Workspaces.
- Alle Schreibvorgänge werden innerhalb einer Workspace-Transaktion serialisiert; ein fehlgeschlagener Vorgang ändert den State nicht.
- Navigation- und Referential-Integrity-Regeln werden nur umgesetzt, soweit der unterstützte Modellumfang sie eindeutig definiert. Sonst expliziter `ODR-E6002`.
- ETags/If-Match sind als **Soll** in LH-FA-15 beschrieben, nicht stillschweigend als Muss vorausgesetzt.

### LH-FA-07.b — OData-V4-CRUD-Vertragsmatrix

| Operation | Erfolg | Rand-/Fehlerfall |
|---|---|---|
| `GET /Products(42)` | `200` + Entity + `@odata.context` nach JSON-Metadatenlevel | nicht vorhanden bei vollständiger Kenntnis: `404`; unbekannte Coverage: Coverage-Fehler |
| `GET /Products(Id=42,Region='EU')` | Composite-Key-Auflösung mit typisierten Literalen | unvollständiger/ungültiger Key: `400` |
| `POST /Products` | `201`, Entity oder `204` gemäß gültiger `Prefer`-Semantik; `Location` bei Erstellung | Duplicate Key: `409`; Typfehler: `400` |
| `PATCH /Products(42)` | `204` oder `200` gemäß `Prefer: return=representation` | unbekannte Property: `400`; fehlende Entity: `404` bei vollständiger Kenntnis |
| `PUT /Products(42)` | Ersetzt die vollständig übermittelten beschreibbaren Properties; fehlende nullable Properties werden `null`, fehlende erforderliche Properties: `400` | Upsert per PUT ist nicht Teil der ersten Version; unbekannte Entity: `404` |
| `DELETE /Products(42)` | `204` | nicht vorhandene Entity: `404` bei vollständiger Kenntnis |

`Prefer: return=minimal` und `Prefer: return=representation` werden für mutierende Requests explizit ausgewertet. `Location` und Entity-ID werden aus Service-Root und typisiertem Key erzeugt; keine unvalidierte Client-Host-Injektion. `@odata.context`, `@odata.count` und `value` folgen dem ausgehandelten JSON-Metadatenlevel. ETags bleiben Soll-Umfang.

### LH-FA-08.b — EDM-Typen und Ausdruckssemantik

| Typ | Lesen/Schreiben | Vergleich/Sortierung | Grenze |
|---|---|---|---|
| `Edm.String`, `Edm.Boolean` | Muss | Muss | Stringvergleich ordinal, keine Locale-Collation |
| `Edm.Int32`, `Edm.Int64` | Muss | Muss | `Int64` ohne JS-Float-Rundung, OData-JSON-konforme Darstellung |
| `Edm.Decimal` | Muss | Muss | Dezimale Arithmetik ohne binären Float; Scale/Precision validieren |
| `Edm.Date`, `Edm.DateTimeOffset` | Muss | Muss | ISO-Formate; Offset auf Instant normieren, Date ohne Zeitzone |
| `Edm.Guid` | Muss | Gleichheit | kanonische UUID |
| `Edm.Binary`, `Edm.Duration`, `Edm.TimeOfDay`, `Edm.Geography*` | nicht im Muss-Umfang | — | explizit `501`, sofern keine reine Replay-Antwort |

**Dreiwertige Logik:** Vergleich mit `null` folgt der dokumentierten OData-V4-Semantik; `eq null`/`ne null` sind unterstützt, logische Ausdrücke werden mit `true`/`false`/`null` ausgewertet, nur `true` lässt einen Datensatz durch `$filter`. Nicht unterstützte Typ-/Operator-Kombinationen liefern `501`, syntaktisch oder typologisch ungültige Ausdrücke `400`. URL-Key-Syntax unterstützt einfache und benannte zusammengesetzte Keys; Literale werden typisiert und String-Escaping geprüft.

### LH-FA-08.a — Query-Interpreter

**Muss-Untermenge:** `$filter` mit `eq`, `ne`, `gt`, `ge`, `lt`, `le`, `and`, `or`, `not` und Klammern für primitive unterstützte EDM-Typen; `$select`, `$orderby`, `$top`, `$skip`, `$count`.

**Pipeline:** Parse → Typprüfung anhand Metadaten → Filter → stabile Sortierung (bei Gleichstand Entity-Key) → Count vor Pagination → Skip/Top → Projektion → OData-V4-JSON.

- Ungültiger Ausdruck: `400` mit OData-Fehlerobjekt.
- Nicht unterstützter, aber gültiger Ausdruck: expliziter `501`-Fehler `ODR-E6001`.
- `$expand`, `$batch`, Actions und Functions gehören zu LH-FA-16 (Soll).
- Standardisierte Query-Fehler und Formatdetails sind durch OData-V4-Contract-Tests zu prüfen.

### LH-FA-09.a — Workspace-Isolation und Reset

Jeder Workspace besitzt `workspace_id`, `baseline_snapshot_id`, isolierten Entity-State und isolierte Replay-Cursor. Kein Write darf andere Workspaces beeinflussen. Reset ist eine atomare Wiederherstellung des Baseline-Zustands inklusive Cursor-Reset. Während Reset werden konkurrierende Schreibvorgänge serialisiert oder kontrolliert abgewiesen.

### LH-FA-09.b — Reset-Semantik

`reset(workspace)` ersetzt den Entity-State atomar durch die Baseline, setzt **alle** Replay-Cursor des Workspaces auf Position 0 und erhöht die Workspace-Revision. Der Vorgang ist auch ohne bisherige Mutationen definiert und idempotent bezüglich des sichtbaren State/Cursor-Inhalts.

### LH-FA-10.a — Snapshots

Snapshots sind unveränderliche, versionierte Zustandsreferenzen. Erstellung erfasst konsistent Entity-State und optional Cursor-Zustand (Cursor sind niemals Bestandteil des Entity-Snapshots). Restore erstellt eine neue Workspace-Revision und lässt sämtliche Replay-Cursor unverändert; Referenzen auf nicht vorhandene Snapshots führen zu `ODR-E7002`.

### LH-FA-11.a — Coverage und Teilmengen

Coverage je EntitySet: `complete`, `partial`, `unknown`. `complete` darf nur aus explizit verifizierter Vollerfassung einschließlich aller Continuation-Seiten entstehen; ein einzelnes `$top`-Resultat ist `partial`. Vor Query-Ausführung wird Coverage geprüft. Ein EntitySet ohne beobachtete Entities ist nicht automatisch `complete`. Bei `strict_coverage=true` wird eine Collection-Abfrage über einen nicht vollständigen Bestand mit `ODR-E6003` abgewiesen. Einzel-Entity-Lookups dürfen nur dann eine definitive `404` liefern, wenn das Nichtvorhandensein abgesichert ist; sonst Coverage-Fehler.

### LH-FA-12.a — Moduswahl und Fallback

`record`, `replay`, `mock`, `hybrid`. Der Modus ist eine Startkonfiguration, nicht per Admin-API live umschaltbar. Hybrid-Policy ist explizit konfiguriert. Default: `mock` → Fehler; optional `mock` → `replay` → Fehler nur für definierte Kategorien (`unsupported`, `coverage_insufficient`), niemals für fachliche `404`, `409` oder `412`; **kein** Upstream-Fallback. Upstream-Fallback nur per Allowlist für lesende Methoden und autorisierte Zielhosts. Für mutierende Methoden ist Upstream-Fallback standardmäßig verboten. Unbekannte Funktionen werden nicht stillschweigend als erfolgreich behandelt.

### LH-FA-12.b — Hybrid-Entscheidungsmatrix

| Ergebnis Mock | Default | Optionaler Replay-Fallback | Upstream-Fallback |
|---|---|---|---|
| Erfolg | Antwort | Nein | Nein |
| Fachliche `404` / `409` / `412` | Antwort/Fehler beibehalten | Nein | Nein |
| Coverage unvollständig | `ODR-E6003` | Nur explizite Policy | Nur GET/HEAD + Allowlist |
| Nicht unterstützte Operation | `ODR-E6001` | Nur explizite Policy | Nur GET/HEAD + Allowlist |
| Ungültige Anfrage | HTTP 400 | Nein | Nein |
| Replay-Miss nach opt-in | `ODR-E5001` | — | Nur gesonderte GET/HEAD-Allowlist |

### LH-FA-13.a — Management und Diagnose

Separater Admin-Listener mit Endpunkten für Sessions, Exchanges, Coverage, Workspaces, Reset, Snapshot und Health. Admin-Endpunkte benötigen Authentisierung bei nicht-loopback Bindings; Default bindet an `127.0.0.1`. Änderungen sind auditierbar. HTTP-Response enthält strukturierte Fehlerkennung und Korrelations-ID.

### LH-FA-02.b — Streaming- und Capture-Lifecycle

- Der HTTP-Adapter besitzt und schließt eingehende Request- und ausgehende Response-Streams; er gibt sie nur als geliehene, kontextgebundene Streams an Transport/Capture weiter. Der Capture-Writer hat `Write`, `Commit` und `Abort`; genau einer der beiden Abschlüsse ist zulässig.
- Transport und Capture werden mit begrenztem Puffer entkoppelt. Backpressure des Capture-Writers darf nicht unbeschränkt Speicher belegen; bei Größenlimit/Fehler wird Capture abgebrochen und der Proxy-Stream gemäß Transportvertrag fortgesetzt.
- Vor dem ersten gesendeten Response-Byte kann ein Transportfehler als Gateway-Status beantwortet werden. Nach dem ersten gesendeten Byte wird der Client-Status **nicht** rückwirkend geändert; der Capture-Status ist `incomplete`, Diagnose wird geschrieben.
- Redaction geschieht vor dauerhafter Ablage; nicht redigierbare Bodies werden verworfen. Ein redigierter Body ist als `replay_equivalence=semantic_or_limited` gekennzeichnet und nicht als bytegenau garantiert.
- Der Capture-Limit-Test sendet einen Body oberhalb von 64 MiB, verifiziert begrenzten Speicherverbrauch, korrektes Live-Streaming und einen nicht replayfähigen unvollständigen Capture.

### LH-FA-14.a — Redaction und Datensicherheit

Default-Redaction: `Authorization`, `Cookie`, `Set-Cookie`, Proxy-Authorization und konfigurierbare Query-/JSON-Felder. Unbekannte sensible Payloads lassen sich nicht automatisch sicher erkennen: Capture muss deaktivierbar sein, und ein expliziter Datenklassifizierungsprozess ist notwendig. Secrets ausschließlich aus sicherer Konfiguration, nicht in Logs oder Exporten. Replay kann redigierte Auth-Daten nicht originalgetreu wiederherstellen.

### LH-FA-15.a — Optimistische Concurrency (Soll)

Entity-Versionen als ETag; `If-Match`/`If-None-Match` entsprechend dem unterstützten OData-Vertrag auswerten. Konflikt ergibt `412`; Versionserhöhung erfolgt atomar. Ohne aktivierte ETag-Unterstützung wird kein scheinbar gültiger Concurrency-Schutz behauptet.

### LH-FA-16.a — Erweiterte OData-Funktionen (Soll)

Erweiterungspunkte für `$expand`, `$batch` mit atomaren Changesets, Actions und Functions. Jede Funktion benötigt eigenen Parser-, Validator-, Executor- und Serializer-Contract-Test. Unbekannte Operationen im Mock-Modus liefern `501`; Replay kann sie unabhängig vom Parser bedienen.

### LH-FA-17.a — OData V2 (Kann)

Versionierter Adapter für V2-Metadaten, Payloads und Query-Semantik; keine automatische V2/V4-Vermischung in einem Service-Modell. Nicht Bestandteil der ersten Abnahme.

---

## 2. Datenstrukturen und Schemas

### SPEC-001 — Fachliche Recording-Struktur

```json
{
  "format": "odata-recorder",
  "version": 1,
  "session_id": "session-01",
  "service_id": "catalog",
  "exchanges": [
    {
      "sequence": 1,
      "method": "GET",
      "path": "/Products",
      "raw_query": "$top=10",
      "match_key": "sha256:...",
      "request": {"headers": {}, "body_blob": null},
      "response": {"status": 200, "headers": {}, "body_blob": "sha256:..."},
      "capture_status": "complete"
    }
  ]
}
```

Persistierte `headers` sind redigiert; rohe Credentials gehören nicht in das Schema. Blob-Hash bezieht sich auf die **gespeicherte redigierte** Darstellung.

### SPEC-002 — OData-Modell

`ServiceModel{version, namespaces, entity_types, entity_sets, navigation_bindings}`; `EntityType{qualified_name, keys[], properties[], navigation_properties[]}`; `Property{name, edm_type, nullable}`. Typ- und Namensvalidierung vor State-Zugriff.

### SPEC-003 — Workspace und Entity

`Workspace{id, service_id, baseline_snapshot_id, revision}`; `EntityState{workspace_id, entity_set, canonical_key, payload_json, known_properties, entity_completeness, entity_version, deleted}`; `ReplayCursor{workspace_id, replay_session_id, recording_session_id, strategy, position}`. Zusammengesetzte Schlüssel werden als kanonisch typisierte Tupel serialisiert, nicht durch naive String-Konkatenation.

### SPEC-004 — Normatives Schema für d-migrate 1.7.1

Das neutrale d-migrate-Schemamodell (YAML-Quelle des Repos) ist die einzige Quelle der Tabellen, Spalten, Schlüssel, Indizes, UNIQUE-, CHECK- und **zusammengesetzten FOREIGN-KEY-Constraints**. d-migrate **1.7.1** versteht `tables.<table>.constraints` mit `type: foreign_key`, `columns: [...]` und `references: {table, columns: [...], on_delete: ...}` sowie `type: unique` und `type: check` mit `expression`. Das Schemaformat selbst bleibt `schema_format: "1.0"`; die Anwendungsschema-Version ist `0.5.0`.

Normative DB-Invarianten:

1. `snapshots(id,service_id)` ist UNIQUE; `workspaces(baseline_snapshot_id,service_id)` referenziert dieses Paar. Ein Workspace darf nie die Baseline eines anderen Service verwenden.
2. `imports(baseline_snapshot_id,service_id)` referenziert dasselbe Paar; `baseline_snapshot_id` ist für erfolgreiche Imports eindeutig (mehrere NULL bei noch nicht abgeschlossenen Imports sind erlaubt). `snapshot_kind=baseline` wird zusätzlich **innerhalb derselben Transaktion** geprüft; ein gewöhnlicher `workspace`-Snapshot darf nicht stillschweigend promotet werden.
3. `sessions(id,service_id)` und `workspaces(id,service_id)` sind UNIQUE; `replay_cursors(recording_session_id,service_id)` und `(workspace_id,service_id)` erzwingen dieselbe Service-Identität. Die Replay-Session ist ein clientdefinierter, workspacegebundener Identifikator, nicht die Recording-Session.
4. `exchanges(session_id,sequence_no)` ist UNIQUE; Index `(session_id,match_key)` unterstützt Matching. Enumerationen für Capture/Redaction/Body/Replay/Coverage und numerische Werte sind als `check`-Constraints in der YAML-Datei enthalten.
5. Jede SQLite-Verbindung setzt `PRAGMA foreign_keys=ON`. Alle mutierenden UseCases prüfen fachliche Invarianten transaktional und dürfen Datenbank-Constraint-Fehler nicht ignorieren. `PRAGMA foreign_key_check` und `integrity_check` sind verpflichtende Gates.

**Migrationsregel:** Das DDL wird ausschließlich mit gepinntem `d-migrate:1.7.1` generiert; es existiert keine parallele manuell gepflegte Tabellen-DDL. CI muss `schema validate`, `schema generate --target sqlite`, DDL-Ausführung, Negativtests und die vollständige Migrationskette auf einer Kopie bestehender Datenbanken durchführen. Die vorliegenden Dokumente behaupten noch keinen erfolgreichen d-migrate-Lauf.

**Versioniertes Snapshot-Manifest:** `manifest_version=1`, `snapshot_kind`, `service_id`, `metadata_hash`, `coverage[]`, `entities[]`, `provenance[]`, `source_session_ids[]`; pro importierter Property werden Quelle (`session_id`, `exchange_id`), Projektion und Hash festgehalten. Canonical JSON und SHA-256 sichern die Integrität. Unveränderliche Baselines dürfen nicht überschrieben werden. Snapshot-Promotion ist ein eigener validierter Vorgang; Coverage wird nie implizit auf `complete` gesetzt.

### SPEC-043 — Import- und Provenienz-Schema

`ImportManifest{import_id, service_id, metadata_hash, source_session_ids[], created_at, entity_count, baseline_snapshot_id, coverage[]}`; `ImportedEntity{entity_set, typed_key, payload, selected_properties[], provenance[]}`. Snapshot-Blobs enthalten eine versionierte kanonische JSON-Darstellung und ihren SHA-256-Hash. `baseline_snapshot_id` wird vor Workspace-Erzeugung auf Existenz geprüft; SQLite-Fremdschlüssel und Migrationstests sichern referenzielle Integrität. Garbage Collection löscht nur Blobs, die weder von Exchange noch Snapshot referenziert werden; Crash nach Blob-Rename vor DB-Commit darf höchstens verwaiste Blobs erzeugen.

### SPEC-044 — Baseline-Coverage und Entity-Vollständigkeit

Jeder Mock-Read löst zunächst `workspace.baseline_snapshot_id` auf und liest Coverage ausschließlich für `(baseline_snapshot_id, entity_set)`. Lokale Writes ändern diese Baseline-Evidenz nicht. Ein neuer Snapshot kann eine neue Baseline erzeugen, aber `complete` wird nicht allein aus lokalem Schreiben abgeleitet. Bei `strict_coverage=true` darf eine nicht nachweislich vollständige Collection keine scheinbar vollständige Ergebnismenge liefern.

Für Entity-Reads gilt: Projektionen ausschließlich bekannter Properties sind zulässig, wenn Identität und Key verifiziert sind; vollständiges `GET` auf `partial` ergibt `ODR-E6003`. `PATCH` auf einer partiellen Entity ist im MVP **abzuweisen**; `PUT` darf nur als explizite vollständige Ersetzung mit validierten Pflichtfeldern und atomarem State-Update ausgeführt werden, ansonsten `ODR-E6003`. Keine unbekannte Property wird durch `null` ersetzt. Tests: `$select`, Null-vs-missing, Pflichtfeld, PATCH/PUT, zwei Baselines eines Services.

### SPEC-045 — Capture-/Replay-Fähigkeitsmatrix

| Capture | Redaction | Body-Integrität | Default Replay |
|---|---|---|---|
| `complete` | `none` | `verified` oder `absent` | Byte-äquivalent zulässig |
| `complete` | `applied` | `verified` oder `absent` | Nur explizit aktiviertes `semantic_or_limited` |
| `complete` | `unsafe` | beliebig | Verboten |
| `incomplete`/`failed` | beliebig | beliebig | Verboten |
| beliebig | beliebig | `invalid` | Verboten |

Default ist **strict byte replay**; redigierte Exchanges sind nicht stillschweigend replayfähig. Bei aktivierter eingeschränkter Policy müssen Antwortfelder mit Redaction sichtbar als eingeschränkt dokumentiert sein; Auth-Credentials können nicht rekonstruiert werden.

### SPEC-046 — Sicheres Streaming-Capture

1. Proxy-Transport und Capture-Tee sind entkoppelt. Rohbytes dürfen nur in flüchtigem RAM bis zur vorgegebenen Grenze liegen, niemals in Klartext in temporären Dateien, Swap-bewussten Spools oder Blob-Dateien. Für nicht sicher streambar redigierbare JSON-Payloads wird die Capture-Seite vor Persistenz verworfen; der Proxy läuft weiter.
2. Komprimierte Payloads werden **nur** mit begrenzter Dekompression und vollständig validierter Encoding-Kette analysiert; sonst `capture_status=failed`, `replay_equivalence=none`. Keine ungeprüften gzip-Rohbytes persistieren, wenn Body-Redaction konfiguriert ist.
3. Unbekannte MIME-Typen, fragmentierte JSON-Tokens, unvollständige JSON-Dokumente und Bodies über dem konfigurierten Grenzwert sind **fail-closed für Capture**. Header-Redaction erfolgt vor jedem dauerhaften Metadaten-Write.
4. Die Streaming-Redaction muss Token über Chunk-Grenzen korrekt erkennen. Bei Absturz dürfen nur bereits redigierte Blobs existieren. Fehler werden mit `ODR-E2401` protokolliert, ohne Payload im Log.
5. Ohne sichere Speicher-/Spool-Garantie ist für große oder unbekannte Bodies die **Body-Aufzeichnung deaktiviert**, nicht bloß auf ein unsicheres Tempfile ausgelagert.

### SPEC-052 — Replay-HTTP-Identität und Lebenszyklus

Pflichtheader bei sequenziellem Replay: `X-Recorder-Workspace` (bestehende Workspace-ID), `X-Recorder-Recording-Session` (aufgezeichnete Session-ID), `X-Recorder-Replay-Session` (vom Test-Harness vergebene, innerhalb des Workspace stabile Client-Session-ID). Kein Header ist ein Upstream-Header; sie werden vor Weiterleitung entfernt. Der Client generiert eine kollisionsresistente Replay-Session-ID (UUID empfohlen) und behält sie über die Request-Folge. Zwei Clients dürfen dieselbe Recording-Session mit unterschiedlichen Replay-Session-IDs unabhängig abspielen. Fehlender oder leerer Header: `ODR-E5003`, HTTP 400, ohne Cursoränderung. Servicefremde IDs: HTTP 409, ohne Cursoränderung. Eine neue Replay-Session startet an Position 0; ein Reset des Workspace löscht alle Cursor atomar; ein Snapshot-Restore setzt die Cursor entsprechend dem explizit gewählten Restore-Vertrag zurück. Keine implizite Ableitung aus Cookies oder Client-IP.

### SPEC-053 — Capture-Bedrohungsmodell und Testvertrag

Geltungsbereich: vom Recorder verwaltete RAM-Puffer, Capture-Queues, Blob-Verzeichnis, temporäre Dateien und Logs. Unredigierte Daten dürfen **nicht durch Recorder-Code** persistent geschrieben werden. Eine Aussage über OS-Swap, Core-Dumps, Container-Runtime oder externe HTTP-Libraries gilt nur unter expliziten Betriebsbedingungen: Core-Dumps deaktiviert, Swap deaktiviert oder verschlüsselt und vom Betreiber kontrolliert, keine unverschlüsselten Diagnose-Dumps, kein Body-Logging in Middleware. Das Betriebsprofil muss diese Voraussetzungen vor Aktivierung von Body-Capture prüfen und dokumentieren.

Bounded memory: konfigurierbare maximale Capture-Bodygröße (Default 8 MiB), maximale entpackte Größe (8 MiB), maximale Queue-Größe (16 MiB), Dekompressions-Verhältnis höchstens 20:1 und Deadline 30 s. Überlauf, unbekanntes Encoding, fehlerhaftes JSON, Abbruch und Timeout setzen `capture_status=failed`, `replay_equivalence=none`; unredigierte Capture-Buffer werden verworfen, während Proxy-Streaming soweit möglich fortgesetzt wird. Nur vollständig redigierte und validierte Bodies dürfen Blob-Dateien erreichen. Backpressure darf nicht zu unbeschränkter Pufferung führen. Sicherheits-Fixtures: JSON-Tokens über Chunk-Grenzen, gzip-Bombe, Content-Encoding-Kette, MIME-Mismatch, Body-Abbruch, langsamer Consumer, Queue-Überlauf, Crash zwischen Blob-Rename und DB-Commit, Redaction-Fehler und Middleware-Logging.

### SPEC-047 — Import-API-Vertrag

`POST /__admin/imports` benötigt Admin-Autorisierung und JSON:

```json
{"import_id":"stable-client-id","service_id":"catalog","source_session_ids":["session-01"],"metadata_hash":"sha256:...","conflict_policy":"reject","target":"new_baseline"}
```

`import_id` ist idempotent: gleiche ID + identischer kanonischer Request-Hash liefert denselben Report; gleiche ID + anderer Inhalt liefert HTTP 409. Nur `target=new_baseline` und `conflict_policy=reject` sind im MVP zulässig. Erfolgreich: HTTP 201 mit `import_id`, `baseline_snapshot_id`, `coverage[]`, `entity_count`, `report`; Wiederholung: HTTP 200. Ungültige Session/Metadaten: 400/422; nicht autorisiert: 401/403; fachlicher Key-Konflikt: 409; Storagefehler: 500 mit Rollback. Import und Baseline-Sichtbarkeit werden atomar committed.

### SPEC-048 — Crash-Recovery und Fault Injection

Fault-Injection-Punkte: vor redigiertem Blob-Write, nach fsync, nach Rename/vor SQLite-Commit, während SQLite-Commit, nach Commit/vor HTTP-Antwort. Erwartung: kein unredigierter persistenter Body; keine referenzierte fehlende Datei nach erfolgreichem Commit; ein verwaister **redigierter** Blob nach Rename ist zulässig und per GC entfernbar. Startup prüft DB-Integrität, FK-Check und referenzierte Blob-Hashes. GC löscht nur nach erneuter Referenzprüfung; Snapshot- und Exchange-Blobs werden berücksichtigt.

### SPEC-049 — MVP-OData-V4-Abnahmematrix

| Bereich | Muss (Golden-Fixture) | Soll | Nicht in MVP |
|---|---|---|---|
| GET | Collection/Key/leer (`ODV4-GET-01..03`) | ETag | Navigation-Expand |
| POST/PATCH/PUT/DELETE | primitive Properties, Key, `null`, Missing (`ODV4-CRUD-01..06`) | Concurrency via If-Match | Deep Insert |
| Query | `eq/ne/gt/ge/lt/le`, `and/or/not`, `$select`, `$orderby`, `$top`, `$skip`, `$count` (`ODV4-Q-01..12`) | zusätzliche Funktionen | `$search`, `$apply` |
| EDM | `Edm.String`, `Boolean`, `Int32`, `Int64`, `Decimal`, `Guid`, `Date`, `DateTimeOffset` (`ODV4-T-01..08`) | weitere Primitive | komplexe Navigationstypen |
| Header | `Accept`, `OData-Version`, `Content-Type` (`ODV4-H-01..03`) | `Prefer`, `If-Match` | herstellerspezifische Header |
| Fehler | 400, 404, 501, Coverage (`ODV4-E-01..04`) | 412 | proprietäre Fehlercodes |

Die Fixture-IDs sind **geplante Testadressen**, keine bereits ausgeführten Tests. HTTP-/OData-Verhalten ist an Referenz-Fixtures und dem Lastenheft zu messen; die Matrix erweitert den Muss-Umfang nicht.

### SPEC-005 — Admin-API (vorläufiger Contract)

| Methode | Pfad | Bedeutung |
|---|---|---|
| GET | `/healthz` | Prozessbereitschaft |
| GET | `/__admin/sessions` | Aufzeichnungen auflisten |
| GET | `/__admin/sessions/{id}/exchanges` | Exchanges lesen |
| GET | `/__admin/coverage` | Coverage lesen |
| POST | `/__admin/imports` | Recording-Session explizit in neue Baseline importieren |
| GET | `/__admin/imports/{id}` | Import-Report und Provenienz lesen |
| POST | `/__admin/workspaces` | Workspace erzeugen |
| POST | `/__admin/workspaces/{id}/reset` | Baseline wiederherstellen |
| POST | `/__admin/workspaces/{id}/snapshots` | Snapshot erzeugen |
| POST | `/__admin/workspaces/{id}/restore` | Snapshot wiederherstellen |

Admin-API wird über separaten Listener ausgeliefert; konkrete OpenAPI-Datei ist vor Implementierung zu erstellen.

### SPEC-006 — Fehlerobjekt

```json
{
  "error": {
    "code": "ODR-E5001",
    "message": "No matching recorded exchange",
    "target": "replay"
  }
}
```

Für HTTP-Antworten des Mock-OData-Dienstes gilt das OData-V4-Fehlerformat; Admin-Fehler können zusätzliche Korrelationsfelder tragen.

---

**Importvertrag:** siehe `SPEC-047`; alle Schreibendpunkte erzwingen Admin-Autorisierung und Audit-Logging.

## 3. Defaults und Konstanten

| ID | Name | Vorschlagswert | Begründung |
|---|---|---|---|
| `SPEC-007` | `LISTEN` | `127.0.0.1:8080` | Sichere lokale Voreinstellung |
| `SPEC-008` | `ADMIN_LISTEN` | `127.0.0.1:8081` | Getrennte Steuerung |
| `SPEC-009` | `REPLAY_STRATEGY` | `sequential` | Wiederholte Requests deterministisch |
| `SPEC-010` | `HYBRID_FALLBACK` | `disabled` | Keine unerwarteten Backend-Aufrufe |
| `SPEC-011` | `STRICT_COVERAGE` | `true` | Keine irreführenden Collection-Antworten |
| `SPEC-012` | `UPSTREAM_TIMEOUT` | `30s` | Begrenzte Wartezeit |
| `SPEC-013` | `MAX_CAPTURE_BODY` | `64 MiB` | Schutz vor unbeschränkter Ablage |
| `SPEC-014` | `SQLITE_BUSY_TIMEOUT` | `5000ms` | Geordnete Konkurrenz |
| `SPEC-015` | `RECORDING_FORMAT_VERSION` | `1` | Versionierbare Persistenz |

---

## 4. Fehler-Codes und Logging-Felder

| ID | Code | Bedingung | Aktion |
|---|---|---|---|
| `SPEC-016` | `ODR-E2001` | Ungültige Konfiguration | Start verhindern |
| `SPEC-017` | `ODR-E2401` | Capture/Redaction fehlgeschlagen | Body nicht persistieren, Diagnose |
| `SPEC-018` | `ODR-E3001` | Recording/Blob beschädigt | Laden verweigern |
| `SPEC-019` | `ODR-E4001` | Upstream nicht erreichbar | HTTP 502/504 je Ursache |
| `SPEC-020` | `ODR-E5001` | Replay-Mismatch | Definierte Fehlerantwort |
| `SPEC-021` | `ODR-E5002` | Replay erschöpft | Definierte Fehlerantwort |
| `SPEC-022` | `ODR-E6001` | Nicht unterstützte OData-Query | HTTP 501 |
| `SPEC-023` | `ODR-E6002` | Nicht unterstützte Mutation | HTTP 501 |
| `SPEC-024` | `ODR-E6003` | Unvollständige Coverage | Explizite Mock-Fehlerantwort |
| `SPEC-050` | `ODR-E5003` | Replay-Zuordnung fehlt oder mehrdeutig | HTTP 400, kein Cursor-Fortschritt |
| `SPEC-051` | `ODR-E6101` | Baseline-Import ungültig/inkonsistent | Import atomar abbrechen |
| `SPEC-025` | `ODR-E7002` | Snapshot unbekannt | HTTP 404 (Admin) |

Logs enthalten `timestamp`, `level`, `mode`, `service_id`, `session_id`, `exchange_id`, `workspace_id`, `correlation_id`, `error_code`, `duration_ms`; keine Credentials oder unredigierten Bodies.

---

## 5. Metriken und Tracing-Felder

| ID | Span | Pflicht-Attribute | Quelle |
|---|---|---|---|
| `SPEC-026` | `odata.gateway.request` | `http.request.method`, `http.response.status_code`, `recorder.mode` | Gateway |
| `SPEC-027` | `odata.record.capture` | `recorder.session_id`, `recorder.capture_status` | Capture |
| `SPEC-028` | `odata.replay.match` | `recorder.match_strategy`, `recorder.match_result` | Replay |
| `SPEC-029` | `odata.mock.query` | `recorder.entity_set`, `recorder.coverage_status` | Mock |
| `SPEC-030` | `odata.workspace.mutation` | `recorder.workspace_id`, `recorder.mutation_type` | State |

Keine sensiblen Query-Literale oder Request-Bodies als Trace-Attribute.

---

## 6. Externe Verträge

| ID | System | Version | Vertrag |
|---|---|---|---|
| `SPEC-031` | OData Protocol | V4.0 | OASIS OData Part 1/2 und JSON Format |
| `SPEC-032` | OData CSDL | V4.0 XML | EDMX/CSDL-Metadaten |
| `SPEC-033` | HTTP | HTTP/1.1; HTTP/2 nach Transportfähigkeit | RFC-konforme Request-/Response-Semantik |
| `SPEC-034` | SQLite | Version mit UPSERT, WAL, FK-Unterstützung | Migrationen + DB-Contract-Tests |
| `SPEC-035` | Docker/OCI | OCI-kompatibles Runtime-Image | Dockerfile, Compose-Integration |

Protokollversionen werden beim Start und pro Service validiert. Herstellerspezifische OData-Erweiterungen sind nicht automatisch Bestandteil des Mock-Vertrags.

---

## 7. Festlegungen der Harness-Werkzeuge

| ID | Werkzeug | Festlegung |
|---|---|---|
| `SPEC-036` | `make test` | Unit-Tests für Domain, Matching, Query-Interpreter, CRUD |
| `SPEC-037` | `make test-integration` | Record → Upstream stoppen → Replay; Mock CRUD → Reset |
| `SPEC-038` | `make test-determinism` | 100 Wiederholungen identischer Inputs; identische normalisierte Outputs und Cursor-Stände |
| `SPEC-039` | `make arch-check` | Hexagonale Importregeln; keine Adapter-Imports im Core |
| `SPEC-040` | `make docs-check` | Kennungen, Links und Dokumentkonventionen prüfen |
| `SPEC-041` | `make test-security` | Redaction, Admin-Bindings, Fallback-Verbote, Credential-Leak-Checks |
| `SPEC-042` | `make test-contract` | OData-V4-JSON, Metadata, Fehlerfälle und HTTP-Semantik |

Abnahmeparameter für Performance und Concurrency sind den Qualitätsanforderungen `LH-QA-*` zu entnehmen; Messaufbau muss reproduzierbar im Test-Harness hinterlegt sein.

---

## 8. Historie

| Datum | Änderung |
|---|---|
| 2026-10-17 | d-migrate 1.7.1, Integritätsconstraints, Replay-/Capture-Verträge, SPEC-ID-Korrektur |
| 2026-10-17 | Initiale technische Spezifikation auf Basis des Lastenhefts und der Kursvorlage |

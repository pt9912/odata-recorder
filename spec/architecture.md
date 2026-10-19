# Architektur — OData Recorder

**Status:** Draft · **Letzte Änderung:** 2026-10-18

**Rolle:** Derivative Architektur-Sicht gemäß `ai-harness-course`: keine neuen Anforderungen, keine Meilensteine, keine Historie, keine ADR-Bezüge.
**Bezug:** `spec/lastenheft.md`, `spec/spezifikation.md`.

**Architekturstil:** Hexagonale Architektur (Ports & Adapters), modularer Monolith.
Die Struktur ergänzt Aufzeichnung und Wiedergabe um OData-Metadaten, Query-Auswertung und zustandsbehaftete Workspaces.
Die hier dargestellten Ports sind konzeptionelle Verträge, keine bereits implementierten APIs.

---

## 1. Komponenten-Übersicht

**Abhängigkeitsrichtung:**

```text
Driving Actor → Driving Adapter → Inbound Port ← Application Service
Application Service → Domain Model
Application Service → Outbound Port ← Driven Adapter → Driven Actor
Composition Root verdrahtet konkrete Implementierungen.
```

Der Application Core kennt weder HTTP-Server- oder -Client-Bibliotheken, SQLite-Treiber, OData-XML-/JSON-Bibliotheken, Dateisystem noch CLI-Frameworks.
Die Transport- und Persistenztechnologie bleibt an den Adaptern.

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
flowchart TB
  Client[OData Client] --> HTTP[HTTP/OData Driving Adapter]
  Operator[Operator / CI] --> CLI[CLI Driving Adapter]
  Operator --> Admin[Admin API Driving Adapter]
  HTTP --> In[Inbound Ports]
  CLI --> In
  Admin --> In
  In --> Svc[Application Services]
  Svc --> Domain[Domain Model]
  Svc --> Out[Outbound Ports]
  Out --> Upstream[HTTP Upstream Adapter]
  Out --> Store[SQLite / Blob Adapter]
  Out --> Codec[OData Codec Adapter]
  Svc --> Sup[Support Ports: Audit, Telemetry, Clock, ID]
  HTTP --> Sup
  Admin --> Sup
  Sup --> Obs[Audit / Telemetry Adapter]
  Sup --> Clock[Clock / ID Adapter]
  Upstream --> Remote[Real OData Service]
  Store --> Disk[Persistent Storage]
  Boot[Composition Root] -. wires .-> HTTP
  Boot -. wires .-> CLI
  Boot -. wires .-> Admin
  Boot -. wires .-> Svc
  Boot -. wires .-> Upstream
  Boot -. wires .-> Store
  Boot -. wires .-> Codec
  Boot -. wires .-> Clock
  Boot -. wires .-> Obs
  classDef box color:#1f2328
  class Client,HTTP,Operator,CLI,Admin,In,Svc,Domain,Out,Upstream,Store,Clock,Codec,Sup,Obs,Remote,Disk,Boot box
```

| ID | Komponente | Rolle |
|---|---|---|
| `ARC-001` | Domain Model (`src/hexagon/model`) | Session, Exchange, Request, Response, OData Service Model, Query AST, Entity, Workspace, Snapshot, Coverage (Auswertung der Coverage-Evidenz), fachliche Invarianten |
| `ARC-002` | Application Services (`src/hexagon/services`) | Record (mit Capture-Pipeline und Redaction), Replay, Mock, Hybrid (Fallback-Policy), **Baseline-Import**, Workspace, Snapshot, Session, Inspection, Maintenance; Use-Case-Orchestrierung; implementiert die Inbound Ports |
| `ARC-003` | Inbound Ports (`src/hexagon/ports/inbound`) | Anwendungsfälle für Request-Verarbeitung, Recording, Replay, Mock und Administration |
| `ARC-004` | Outbound Ports (`src/hexagon/ports/outbound`) | UpstreamClient, RecordingRepository, SessionRepository, BlobRepository, CursorRepository, WorkspaceRepository, EntityRepository, SnapshotRepository, MetadataRepository, CoverageRepository, ImportRepository, ODataCodec, StorageMaintenance |
| `ARC-005` | HTTP/OData Driving Adapter (`src/adapters/driving/httpodata`) | HTTP-Eingang, Header-/Body-Streaming, Mapping von Request und Response auf Domain-DTOs; OData-Parsing und -Serialisierung laufen über den Codec-Port im Core |
| `ARC-006` | CLI Driving Adapter (`src/adapters/driving/cli`) | Konfiguration, Kommandoauswahl, Start-/Exit-Verhalten |
| `ARC-007` | Admin API Driving Adapter (`src/adapters/driving/admin`) | Workspaces, Snapshots, Sessions, Coverage, Diagnose |
| `ARC-008` | HTTP Upstream Driven Adapter (`src/adapters/driven/httpupstream`) | HTTP/TLS-Verbindung zum echten OData-Service |
| `ARC-009` | Recording/State Driven Adapter (`src/adapters/driven/sqlite`) | Transaktionale Speicherung und Abfrage der Sessions, Records und Workspace-Daten; SQLite-SQL und Migrationen |
| `ARC-010` | Blob Driven Adapter (`src/adapters/driven/blobs`) | Bytegetreue, content-adressierte Body-Ablage und Integritätsprüfung |
| `ARC-011` | OData Codec Driven Adapter (`src/adapters/driven/odata`) | Implementiert den Port `ODataCodec`: CSDL-Parsing, OData-URL-Parsing, JSON-Serialisierung, externe Protokollrepräsentation |
| `ARC-012` | Bootstrap / Composition Root (`src/bootstrap`) | Verdrahtung aller Ports, Services und Adapter; Lifecycle; stößt beim Start vor dem Öffnen der Listener `MaintenanceUseCase` an, legt danach im Modus `record` für jeden konfigurierten Service die erste Recording-Session an (`SessionUseCase.OpenInitialSession`), beendet beim geordneten Shutdown alle offenen Sessions (`SessionUseCase.CloseAll`) und wählt je Betriebsmodus den Use Case, den der OData-Listener bedient |
| `ARC-019` | Clock/ID Driven Adapter (`src/adapters/driven/clock`) | Systemzeit und ID-Vergabe hinter `Clock` und `IDGenerator`; in Tests ersetzbar |
| `ARC-021` | Querschnitts-Ports (`src/hexagon/ports/support`) | `AuditLog`, `Telemetry`, `Clock`, `IDGenerator`; jede Schicht darf sie nutzen, damit Adapter ohne Use Case Audit-Einträge, Korrelations-IDs und Spans erzeugen können |
| `ARC-020` | Audit/Telemetry Driven Adapter (`src/adapters/driven/audit`, `src/adapters/driven/telemetry`) | Audit-Log, strukturierte Logs, Metriken und Spans hinter `AuditLog` und `Telemetry` |

**Baseline-Import:** `ImportUseCase` in `ARC-002` liest Exchanges über `RecordingRepository`, interpretiert OData-V4-JSON über den Port `ODataCodec` (`ARC-011`), validiert Schlüssel und Typen gegen `MetadataRepository`, schreibt ein unveränderliches Snapshot-Manifest über `SnapshotRepository`/`BlobRepository`, führt Import-Eintrag und Baseline-Quellen über `ImportRepository` und veröffentlicht die Baseline atomar.
Kein HTTP-Adapter und kein SQLite-Adapter entscheidet selbst über Deduplizierung oder Coverage.

**Abgrenzung:** Die OData-Query-Semantik (Filter, Sortierung, Paging, Typregeln) gehört zum Core; das Parsen konkreter HTTP-/CSDL-/JSON-Darstellungen liegt im Codec-Adapter.
Der Core verwendet einen technologieunabhängigen AST.

**Codec-Zugriff:** Kein Driving Adapter ruft den Codec auf und importiert `adapters/driven/odata`.
Der HTTP-Adapter reicht Methode, Pfad, `raw_query`, Header und Body-Strom unverändert als Domain-DTO an einen Inbound Port;
Parsen und Serialisieren geschieht im Application Service über den Outbound Port `ODataCodec`.
Die konkrete Implementierung verdrahtet allein `src/bootstrap` (Composition Root).
Damit bleibt der Importgraph auf die erlaubten Kanten beschränkt, und der Codec lässt sich im Test durch einen Fake ersetzen.

---

## 2. Schichten und Constraints

| Komponente(n) | Schicht | Verantwortlichkeit | Darf importieren | Darf NICHT importieren |
|---|---|---|---|---|
| `ARC-001` | Domain | Fachliche Werte, Invarianten, AST | Standardbibliothek ohne I/O | Services, Ports, Adapter, SQL/HTTP-Bibliotheken |
| `ARC-003`, `ARC-004`, `ARC-021` | Ports | Fachliche Schnittstellen | Domain | Adapter, CLI, SQLite, HTTP-Server |
| `ARC-002` | Application | Use Cases, State- und Matching-Logik | Domain, Inbound Ports (implementiert sie), Outbound Ports, Querschnitts-Ports | Driving Adapter, Driven Adapter, andere Services, SQLite, konkrete HTTP-Clients |
| `ARC-005`–`ARC-007` | Driving Adapters | Transport- und Bedienoberfläche | Inbound Ports, Querschnitts-Ports, Domain | Driven Adapter, andere Driving Adapter, Application Services |
| `ARC-008`–`ARC-011`, `ARC-019`, `ARC-020` | Driven Adapters (einschließlich Codec, Clock, Audit, Telemetry) | Upstream, Persistenz, Serialisierung | Domain, passende Outbound Ports, Querschnitts-Ports | Driving Adapter, Application Services |
| `ARC-012` | Composition Root | Dependency Injection, Konfiguration, Lifecycle | alle | — |

**Zusätzliche Hard Rules:**

- Der Core verwendet in öffentlichen fachlichen Port-Signaturen keinen Typ einer Datenbank- oder HTTP-Bibliothek.
- HTTP-Header und Bodies werden am Adapter-Rand in kanonische Domain-DTOs abgebildet; Rohbytes dürfen als Byte-Folgen oder Streams über explizite Ports weitergegeben werden.
- SQLite-SQL und Migrationen liegen ausschließlich in `ARC-009` (`adapters/driven/sqlite/migrations`); Dateisystemzugriffe für Blobs in `ARC-010`.
- Die OData-Parserbibliothek ist nur in `ARC-011` erlaubt.
  Ein AST-Typ im Domain Model darf keine Parserbibliothek referenzieren.
- Adapter importieren einander nicht, Services ebenso wenig: Ein Use Case ruft einen anderen nur über dessen Inbound Port auf (Hybrid → Mock, Replay).
  Gemeinsame Funktionen gehören in Domain oder Ports.
  Die Capture-Pipeline ist ein internes Paket des Record-Service, kein eigener Service.
- Kein inneres Package beendet den Prozess; Signalbehandlung liegt in CLI/Bootstrap.
- Die Admin API ist von der OData-Service-API getrennt und darf keine fachlichen Mock-Abkürzungen direkt in die Datenbank schreiben.

### 2.1 Package-Struktur

```text
src/
├── hexagon/
│   ├── model/
│   │   ├── recording/
│   │   ├── odata/
│   │   ├── workspace/
│   │   ├── coverage/
│   │   ├── redaction/
│   │   └── errors/
│   ├── ports/
│   │   ├── inbound/
│   │   ├── outbound/
│   │   └── support/
│   └── services/
│       ├── record/
│       │   └── capture/
│       ├── replay/
│       ├── mock/
│       ├── hybrid/
│       ├── inspection/
│       ├── maintenance/
│       ├── baselineimport/
│       ├── session/
│       ├── snapshot/
│       └── workspace/
├── adapters/
│   ├── driving/
│   │   ├── httpodata/
│   │   ├── admin/
│   │   └── cli/
│   └── driven/
│       ├── httpupstream/
│       ├── sqlite/
│       │   └── migrations/
│       ├── blobs/
│       ├── audit/
│       ├── telemetry/
│       ├── clock/
│       └── odata/
│           ├── csdl/
│           ├── urlparser/
│           └── jsoncodec/
└── bootstrap/
tests/
├── contract/
└── integration/
```

### 2.2 Inbound Ports

| Port | Aufgabe | Rückgabe |
|---|---|---|
| `ImportUseCase` | Aufgezeichnete Entities explizit in eine neue Mock-Baseline überführen; Import-Report und Provenienz lesen; Wiederholung mit gleicher `import_id` idempotent behandeln | Snapshot-ID, Import-Report, Coverage-Evidenz |
| `RecordUseCase` | HTTP-Austausch an Upstream weiterleiten und erfassen; die Aufzeichnung einer Session anhalten (`Drain`: neue Exchanges abweisen, laufende bis zur Frist abwarten, überfällige als `incomplete` schreiben) | Response, Capture-Status bzw. Zahl der abgebrochenen Exchanges |
| `ReplayUseCase` | Recording zu explizitem Workspace, Recording-Session und Replay-Session matchen | aufgezeichnete Response oder Mismatch |
| `MockUseCase` | OData-Request auf Workspace-State ausführen | OData-Response |
| `HybridUseCase` | Request im Modus `hybrid` an `MockUseCase` geben und nach Fallback-Policy an `ReplayUseCase` oder den Upstream weiterreichen | OData-Response oder Fallback-Fehler |
| `WorkspaceUseCase` | Workspace erstellen, auflisten, lesen, löschen, resetten, mutieren; Snapshot wiederherstellen | Workspace-Revision bzw. Verwaltungsdaten |
| `SnapshotUseCase` | Snapshot erzeugen und auflisten | Snapshot-ID bzw. Verwaltungsdaten |
| `SessionUseCase` | Recording-Sessions anlegen (auch die erste Session je Service beim Start), beenden (auch alle beim geordneten Shutdown), auflisten, lesen, löschen | Verwaltungsdaten bzw. Löschbericht |
| `MaintenanceUseCase` | Startwartung vor dem Öffnen der Listener: offene Sessions auf `incomplete` setzen, Aufbewahrung anwenden (über `SessionRepository` und `ImportRepository`, nicht über `SessionUseCase`), Integritätsprüfung, Blob-Bereinigung; löschende Use Cases (`SessionUseCase`, `WorkspaceUseCase`) rufen die Blob-Bereinigung nach dem Commit hier auf; `MaintenanceUseCase` ruft selbst keinen anderen Use Case auf, der Aufrufweg ist damit zyklenfrei | Wartungsbericht |
| `InspectionUseCase` | Exchanges, Coverage und Diagnose lesen | paginierte Verwaltungsdaten |

Der HTTP-Adapter entscheidet **nicht** über Replay-Cursor, Match-Key, Coverage oder CRUD-Semantik.
Der Core entscheidet **nicht** über TLS, Socket-Lifecycle oder HTTP-Chunking.

### 2.3 Outbound Ports

| Port | Verantwortung |
|---|---|
| `UpstreamClient` | HTTP-Request an erlaubten Zielservice senden; Response streamen |
| `RecordingRepository` | Exchanges append-only persistieren und lesen |
| `SessionRepository` | Sessions anlegen, Status setzen (`open` → `complete`/`incomplete`), auflisten, offene Sessions finden, samt Exchanges und Cursor löschen |
| `BlobRepository` | Bodies sicher ablegen/lesen, Hash prüfen; `CaptureWriter` (`Write`, genau ein `Commit` oder `Abort`) liefert er für das Capture; Blobs auflisten, entfernen und ihre Integrität prüfen; der Abgleich gegen die Referenzmenge (`StorageMaintenance`) gehört zu `MaintenanceUseCase` |
| `MetadataRepository` | Service-Metadaten und Versionen verwalten |
| `WorkspaceRepository` | Workspaces mit Revision und Baseline-Bezug persistieren, auflisten und löschen |
| `EntityRepository` | Transaktionale Entity-Reads und -Writes pro Workspace |
| `SnapshotRepository` | Unveränderliche Snapshots, versionierte Manifeste und Restore |
| `StorageMaintenance` | Datenbank-Integritätsprüfung (`foreign_key_check`, `integrity_check`) ausführen und die Menge aller von Exchanges, Snapshots und Service-Metadaten referenzierten Blob-Hashes liefern |
| `CoverageRepository` | Coverage-Status mit Evidenz je Baseline speichern |
| `ImportRepository` | Import-Einträge (`import_id`, Anfrage-Hash, Report) und die Baseline-Quellen einer Baseline in der Transaktion der Baseline schreiben und lesen; zu einer Session die Baselines abfragen, deren Quelle sie ist |
| `CursorRepository` | Veränderlichen Replay-Cursor-Zustand je `(workspace, replay_session, recording_session)` atomar lesen und fortschreiben; getrennt vom append-only `RecordingRepository` |
| `ODataCodec` | CSDL-, OData-URL- und OData-JSON-Darstellung in Domain-Modell und AST überführen und zurück |
| `AuditLog` (Querschnitts-Port) | Audit-Einträge der Verwaltungs- und Löschvorgänge ohne Geheimnisse anhängen |
| `Telemetry` (Querschnitts-Port) | Strukturierte Logs, Metriken und Spans (`SPEC-026`–`SPEC-030`) ausgeben |
| `Clock` / `IDGenerator` (Querschnitts-Ports) | Deterministische Zeit und IDs in Tests |

**Streaming-Port-Lifecycle:** Der HTTP-Adapter besitzt Request-/Response-Streams und schließt sie.
Der Record-Service erhält kontextgebundene Stream-Handles; `CaptureWriter` erlaubt `Write` und genau einen Abschluss (`Commit` oder `Abort`).
Ein Capture-Fehler nach Beginn der Response-Ausgabe darf den HTTP-Status nicht rückwirkend verändern.
Backpressure und Body-Limit führen zum Abbruch der Aufzeichnung, nicht zu unbeschränktem Puffern.
Redigierte Bodies sind als eingeschränkt replayfähig gekennzeichnet.

**Replay-Zuordnung:** Die Zuordnung wird an der HTTP-Grenze entgegengenommen und im Replay-Service gegen die Spezifikation validiert (`LH-FA-04.b`, `SPEC-052`).
Der Cursor gehört zum Tupel `(workspace, replay_session, recording_session)` und liegt hinter `CursorRepository`.

**Reset vs. Restore:** Die Wirkung auf Entity-State und Cursor legen `LH-FA-09` und `LH-FA-10` fest; Reset greift auf `EntityRepository` und `CursorRepository` zu, Restore nur auf `EntityRepository`.

**Port-Design:** Die Transaktionsgrenze ist fachlich (`WithWorkspaceTransaction` oder äquivalentes Unit-of-Work-Konzept), nicht als offene SQL-Connection exponiert.
Reset (Entity-State und Cursor) und Löschen eines Workspace laufen in einer solchen Transaktion über `WorkspaceRepository`, `EntityRepository`, `CursorRepository` und `SnapshotRepository`;
ebenso die Import-Publikation (Snapshot, Import-Eintrag, Baseline-Quellen, Coverage über `SnapshotRepository`, `ImportRepository`, `CoverageRepository`) und das Löschen einer Session (Prüfung der Baseline-Quellen über `ImportRepository`, Exchanges und Cursor über `SessionRepository`);
die Blob-Bereinigung folgt dem Commit über `MaintenanceUseCase`.
Für große Bodies existiert ein Streaming-Port mit begrenztem Puffer und explizitem Abschluss.

### 2.4 Fachliche Zustände

**Exchange-Capture:** Ein laufender Austausch ist noch keine Zeile.
Er wird mit `capture_status` `complete`, `incomplete` oder `failed` abgeschlossen und erst dann angehängt.
Nur `complete`-Exchanges sind regulär replayfähig.

**Session:** `open → complete` oder `open → incomplete` (Werte `sessions.status`);
`SessionUseCase` setzt den Status beim Beenden der Aufzeichnung, `MaintenanceUseCase` beim Start (offene Sessions werden `incomplete`);
Replay und Import weisen `open`-Sessions ab.

**Replay:** `UNASSIGNED → READY → MATCHING → RESPONDING → READY`.
Ein Mismatch, ein nicht replayfähiger Treffer und ein Strict-Mismatch beenden die Anfrage mit einem Fehler und lassen den Cursor unverändert in `READY`.
`EXHAUSTED` gilt je Match-Key (im Modus `strict` am Ende der Folge) und sperrt weitere Anfragen mit anderem Key nicht.
Der Cursor gehört zur Application Session / zum Workspace, nicht zum HTTP-Adapter.

**Mock:** `READY → VALIDATING → READING|MUTATING → SERIALIZING → READY`.
Mutationen werden erst nach erfolgreicher Validierung committed.

**Coverage:** `unknown`, `partial`, `complete`; vor Collection-Reads geprüft; null beobachtete Entities sind kein Leerheitsbeweis; `complete` ist nur nach belegter Vollerfassung zulässig.

### 2.5 Modus-Routing

| Modus | Eingangsweg | Upstream erlaubt? |
|---|---|---|
| `record` | Record Service → Upstream + Recording | Ja, konfiguriert |
| `replay` | Replay Service → Recording | Nein |
| `mock` | Mock Service → Entity State | Nein |
| `hybrid` | Mock; bei explizit erlaubtem `unsupported`/`coverage_insufficient` optional Replay; sonst Fehler | Nur explizit erlaubte Requests (`LH-FA-12.a`) |

Fachliche `404`, `409`, `412` und ungültige Requests lösen keinen Fallback aus.
Der Betriebsmodus wird beim Prozessstart festgelegt und nicht über die Admin API gewechselt;
der Composition Root bindet je Modus genau einen Use Case (`RecordUseCase`, `ReplayUseCase`, `MockUseCase` oder `HybridUseCase`) an den OData-Listener, der HTTP-Adapter kennt den Modus nicht.
Der Use Case des Modus ermittelt Service und Recording-Session aus Pfad und Headern des Rohrequests (`ODR-E5005`, `ODR-E5003`, `ODR-E5009`); der HTTP-Adapter reicht Methode, Pfad, Query, Header und Body unverändert weiter und vergibt nur die Korrelations-ID und den Span `odata.gateway.request`.
Unbekannte Operationen im Mock dürfen nicht in einen unkontrollierten Upstream-Fallback münden.
Die Fallback-Policy wird vor dem Senden an den Upstream geprüft.
`HybridUseCase` ruft `MockUseCase` und `ReplayUseCase` über deren Inbound Ports auf, nie über deren Pakete.

---

## 3. Externe Abhängigkeiten

| ID | System | Rolle | Substituierbarkeit |
|---|---|---|---|
| `ARC-013` | OData-V4-Server | Externer Driven Actor für Record / erlaubten Hybrid-Fallback | Über `UpstreamClient` fakebar |
| `ARC-014` | SQLite | Externer Driven Actor: Persistenz für Sessions, Exchanges, Entity State, Cursor und Coverage | Über Repository-Ports austauschbar |
| `ARC-015` | Dateisystem / OCI Volume | Externer Driven Actor: Persistente Blob-Ablage und Datenverzeichnis | Über Blob-Port austauschbar |
| `ARC-016` | HTTP/OData-Client | Driving Actor, der Record/Replay/Mock aufruft | Test-HTTP-Client |
| `ARC-017` | OData V4 / CSDL / HTTP | Externe Protokollverträge für die Adapter | Contract-Test-Suite |
| `ARC-018` | Docker/OCI Runtime | Prozessausführung, Konfiguration, Persistenz-Volumes | Binary auch nativ startbar |

Die Bibliothekswahl ist keine Eigenschaft des Domain Models.
Ein Austausch des HTTP-Stacks, SQLite-Treibers oder OData-Parsers darf die fachlichen Use Cases nicht verändern.

---

## 4. Sequenz-Diagramme

### 4.1 LH-FA-01 / LH-FA-02 — Record

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant C as OData Client
    participant H as HTTP Adapter
    participant R as Record Service
    participant U as Upstream Port
    participant B as Blob/Recording Ports
    participant Q as Session Repository
    participant S as OData Server
    C->>H: GET /Products?$top=10
    H->>R: Record(raw request)
    R->>R: Resolve service from path, correlation id
    R->>Q: Select open session of service (ODR-E5009 if none)
    R->>U: Forward(request)
    U->>S: HTTP Request
    S-->>U: HTTP Response
    U-->>R: Status, Header, Body Stream
    R->>B: Persist redacted body + exchange
    R-->>H: Response + capture status
    H-->>C: HTTP Response
```

Die technische Streaming-Implementierung darf Weitergabe und Persistenz überlappen; das Diagramm beschreibt logische Verantwortlichkeiten, keine vollständige Byte-für-Byte-Zeitordnung.

### 4.2 LH-FA-18 — Recording-Import

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant O as Operator (Driving Actor)
    participant A as Admin Adapter
    participant I as Import Service
    participant R as Recording Repository
    participant C as OData Codec Port
    participant S as Snapshot Repository
    O->>A: POST /__admin/imports
    A->>I: Import(session IDs, service)
    I->>R: Read completed exchanges
    R-->>I: Recorded bodies + provenance
    I->>C: Parse and validate against metadata
    C-->>I: Typed entities and coverage evidence
    I->>I: Merge/deduplicate or reject conflicts
    I->>S: Write manifest and publish atomically
    S-->>I: Immutable baseline ID
    I-->>A: Import report
    A-->>O: Baseline ID + coverage
```

### 4.3 LH-FA-04 / LH-FA-05 — Replay

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant C as Client
    participant H as HTTP Adapter
    participant R as Replay Service
    participant P as Recording Repository
    participant K as Cursor Repository
    C->>H: GET /Products
    H->>R: Replay(request, workspace_id, replay_session_id, recording_session_id)
    R->>R: Normalize + Match
    R->>P: Read matching exchanges
    R->>K: Read isolated cursor
    alt Match
      P-->>R: Stored response
      R->>K: Commit cursor (workspace, replay_session, recording_session) atomically
      R-->>H: Response
      H-->>C: Recorded response
    else Miss
      R-->>H: ODR-E5001
      H-->>C: Error response
    end
```

Der Replay-Service besitzt keine Upstream-Abhängigkeit; eine Netzwerkverbindung zum Backend ist in diesem Modus architektonisch ausgeschlossen.

### 4.4 LH-FA-07 / LH-FA-08 — Stateful Mock

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant C as Client
    participant H as HTTP/OData Adapter
    participant M as Mock Service
    participant E as Entity Repository
    participant V as Coverage Repository
    C->>H: GET /Products?$filter=Price gt 100
    H->>M: Execute(raw request, workspace)
    M->>M: Parse OData into AST via codec port
    M->>E: Read entity state
    E-->>M: Typed entities
    M->>V: Read coverage (baseline, entity_set)
    V-->>M: Coverage status
    M->>M: Validate and execute AST (domain model)
    M->>M: Serialize result via codec port
    M-->>H: OData response + coverage
    H-->>C: HTTP response with OData V4 JSON
```

### 4.5 LH-FA-09 / LH-FA-10 — Reset und Snapshot

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant T as Test Runner (Driving Actor)
    participant A as Admin Adapter
    participant W as Workspace Service
    participant S as Snapshot Repository
    participant E as Entity Repository
    participant K as Cursor Repository
    T->>A: POST /__admin/workspaces/{id}/reset
    A->>W: Reset(workspace)
    W->>S: Read baseline snapshot
    S-->>W: Immutable state
    Note over W,K: one unit of work
    W->>E: Replace state
    E-->>W: New revision
    W->>K: Reset all cursors of workspace
    W-->>A: Reset complete
    A-->>T: 200 + revision
```

Restore verwendet denselben Weg ohne den Schritt zum `CursorRepository`:

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant T as Test Runner
    participant A as Admin Adapter
    participant W as Workspace Service
    participant S as Snapshot Repository
    participant E as Entity Repository
    T->>A: POST /__admin/workspaces/{id}/restore
    A->>W: Restore(workspace, snapshot_id)
    W->>S: Read snapshot
    S-->>W: Immutable state or unknown snapshot
    W->>E: Replace state atomically (new revision)
    W-->>A: Restore complete
    A-->>T: 200 + revision
```

### 4.6 LH-FA-11 / LH-FA-12 — Coverage und Hybrid-Fallback

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant C as Client
    participant H as HTTP Adapter
    participant G as Hybrid Service
    participant M as Mock Service
    participant R as Replay Service
    participant U as Upstream Port
    C->>H: OData request
    H->>G: Handle(request)
    G->>M: Try mock (MockUseCase)
    alt Business error (400, 404, 409, 412)
      M-->>G: Error
      G-->>H: Error without fallback
    else Supported and sufficiently covered
      M-->>G: Mock response
      G-->>H: Response
    else Explicitly eligible for replay
      G->>R: Try replay (ReplayUseCase)
      alt Recorded exchange found
        R-->>G: Recorded response
        G-->>H: Response
      else Replay miss ODR-E5001 and upstream fallback explicitly allowed
        G->>U: Forward allowed request
        U-->>G: Upstream response
        G-->>H: Response
      else Any other replay error (exhausted, mismatch, header)
        R-->>G: Error
        G-->>H: Error without fallback
      end
    else Upstream fallback explicitly allowed
      G->>U: Forward allowed request
      U-->>G: Upstream response
      G-->>H: Response
    else
      G-->>H: Explicit error
    end
    H-->>C: HTTP Response
```

Fallback ist eine eigene Policy-Entscheidung und kein automatisch erzwungener Teil der Sequenz.

---

### 4.7 LH-FA-11 / LH-FA-18 — Baseline-Provenienz und Property-Masken

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant C as Mock Client
    participant H as OData Adapter
    participant M as Mock Service
    participant W as Workspace Repository
    participant E as Entity Repository
    participant V as Coverage Repository
    C->>H: GET /Products(1)
    H->>M: Execute(raw request, workspace)
    M->>M: Parse URL into AST via codec port (entity, projection)
    M->>W: Resolve baseline_snapshot_id
    W-->>M: Immutable baseline ID
    M->>E: Read entity + known_properties
    E-->>M: Payload + property mask
    M->>V: Read coverage (baseline, entity_set)
    V-->>M: Coverage status
    M->>M: Check baseline coverage and projection
    alt Requested properties known
      M->>M: Serialize result via codec port
      M-->>H: OData response
      H-->>C: 200 OData JSON
    else Unknown properties
      M-->>H: ODR-E6003
      H-->>C: Coverage error
    end
```

Die Coverage ist an die **Baseline-ID** gebunden, nicht an den Service.
Die Property-Maske gehört zu jedem gespeicherten Entity-State und zu jedem Snapshot.
Lokale Writes ändern die Herkunftsevidenz nicht.

### 4.8 LH-FA-02 / LH-FA-14 — Capture ohne Klartext-Spool

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant U as Upstream Port
    participant R as Record Service
    participant B as Blob Repository Port
    participant X as Recording Repository Port
    U-->>R: Response stream
    R-->>R: Stream to client via HTTP adapter
    R->>R: Bounded tee, decode + redact in RAM (capture pipeline)
    alt Safe and complete
      R->>B: CaptureWriter.Write redacted bytes only, then Commit
      B-->>R: Verified blob hash
      R->>X: Append complete exchange
    else Unsafe/unknown/over limit
      R->>B: CaptureWriter.Abort
      R->>X: Append capture failure without body
    end
```

Redaction und Capture-Pipeline gehören zum Record-Service (`src/hexagon/services/record/capture`, Regeln in `src/hexagon/model/redaction`); der Blob-Adapter sieht nur redigierte Bytes.
Rohbytes werden **nie** als unredigierte Temp- oder Blob-Dateien geschrieben.
Unsichere Capture-Pfade scheitern geschlossen, ohne den laufenden Proxy unnötig zu blockieren.
Replay nutzt die Matrix aus `SPEC-045`.

### 4.9 LH-FA-02 / LH-FA-03 — Session beenden und Wiederherstellung

```mermaid
%%{init: {'themeVariables': {'lineColor': '#8b949e', 'signalColor': '#8b949e', 'signalTextColor': '#8b949e', 'actorLineColor': '#8b949e', 'loopTextColor': '#8b949e', 'textColor': '#c9d1d9', 'primaryTextColor': '#1f2328', 'nodeTextColor': '#1f2328', 'actorTextColor': '#1f2328', 'noteTextColor': '#1f2328', 'edgeLabelBackground': '#30363d'}}}%%
sequenceDiagram
    participant O as Operator (Driving Actor)
    participant A as Admin Adapter
    participant S as Session Service
    participant P as Record Service
    participant X as Recording Repository
    participant R as Session Repository
    participant B as Bootstrap
    participant M as Maintenance Service
    O->>A: POST /__admin/sessions/{id}/close
    A->>S: Close(session)
    S->>P: Drain(session, SHUTDOWN_TIMEOUT)
    P-->>S: In-flight exchanges finished or aborted as incomplete
    S->>X: Read exchanges' capture status
    X-->>S: All complete? / any incomplete or failed
    S->>R: Set status complete or incomplete
    S-->>A: Session status
    A-->>O: 200 + status
    Note over B,R: Start nach Absturz, vor den Listenern
    B->>M: Startup()
    M->>R: Find open sessions, set incomplete
    B->>S: OpenInitialSession per service (mode record)
```

## 5. Fehlermodelle und Resilienz

| Fehlerquelle | Behandlungsschicht | Verhalten / Logging |
|---|---|---|
| Upstream-TLS/Timeout | `ARC-008` → `ARC-002` | Klassifizierter Transportfehler, HTTP 502/504 soweit möglich |
| Replay-Mismatch | `ARC-002` | `ODR-E5001`, Session/Exchange/Match-Strategie |
| Defektes Recording/Blob | `ARC-009`/`ARC-010` | `ODR-E3001`, keine Ersatzdaten |
| Unbekannte OData-Funktion | `ARC-011` → `ARC-002` | Expliziter 501-Pfad |
| Ungenügende Coverage | `ARC-002` | `ODR-E6003`, keine irreführende leere Collection |
| Concurrency-Konflikt | `ARC-002`/`ARC-009` | Atomarer Rollback, 409 oder 412 je Vertragsfall |
| Redaction-Fehler | `ARC-002` (Record-Service, Capture-Pipeline) | Keine unredigierte Ablage; Capture-Fehler |
| Ungültige Admin-Anfrage | `ARC-007` | 400/401/404/409/422, strukturierter Audit-Eintrag |
| Prozess-Signal | `ARC-006`/`ARC-012` | Graceful Shutdown nach `LH-FA-02.c` |
| Offene Session nach Absturz | `ARC-012` → `ARC-002` | Beim Start `incomplete`, bevor Requests angenommen werden |

**Concurrency:** Gleichzeitige HTTP-Verbindungen sind zulässig.
Replay-Cursor und Workspace-Schreibtransaktionen werden jeweils separat synchronisiert; Tests beeinflussen andere Workspaces nicht.
SQLite wird mit kurzer Transaktionsdauer und Busy-Timeout betrieben.

**Import-Konsistenz:** Importiert wird in einen temporären Zustand.
Ein Key-Konflikt, eine Metadatenabweichung oder ein defekter Blob verhindert die Publikation der gesamten Baseline.
Provenienz und Coverage-Evidenz sind Teil des Manifests.

**Migration und Recovery:** Die initiale SQLite-Migration definiert die FK-Beziehungen einschließlich `workspaces.baseline_snapshot_id`;
Recovery prüft DB/FK/Blob-Integrität und unterscheidet verwaiste redigierte Blobs von fehlenden referenzierten Blobs.
Die Fault-Injection-Matrix steht in `SPEC-048`.

**Crash-Konsistenz:** SQLite-Commit und Blob-Rename sind nicht eine einzige globale Transaktion.
Der Blob wird zuerst sicher geschrieben; erst danach referenziert ihn eine DB-Transaktion.
Verwaiste Blobs sind tolerierbar und werden später bereinigt; fehlende referenzierte Blobs sind Integritätsfehler.

**Graceful Shutdown:** Ablauf und Frist legt `LH-FA-02.c` fest; Bootstrap und CLI orchestrieren, die Services finalisieren offene Captures.

**Sicherheit:** Redaction vor Persistenz;
Admin-Zugriff nur mit Token (`LH-FA-13.a`), Listener standardmäßig loopback;
kein produktiver Auth-Proxy oder generischer Internet-Proxy;
TLS-Zertifikate werden geprüft.
Im Hybrid-Modus verhindern Allowlists ungewollte Mutationen.

---

**Mock-Integrität:** Coverage wird je unveränderlicher Baseline gespeichert; der Workspace verweist per Fremdschlüssel auf seine Baseline.
Partielle Entity-Projektionen tragen eine Property-Maske.
Unbekannte Werte sind nicht gleich `null`; partielle Entity-Updates werden entsprechend `SPEC-044` beschränkt.

## 6. Qualitätssicherung der Architektur

- **Import-Tests:** Paginierte Teilmengen, Duplikate, widersprüchliche Keys, Projektionen, Provenienz und atomarer Rollback.
- **Domain-Tests:** Query AST, Schlüsselkanonisierung, Entity-Validierung, Coverage-Invarianten, Snapshot-Revisionsmodell; ohne Netzwerk und Datenbank.
- **Application-Tests:** Fake-Ports für Record/Replay/Mock, deterministische Uhr und IDs; Sequenz-Matching und Reset ohne Adapter.
- **Adapter-Contract-Tests:** HTTP Header/Body, CSDL, OData-V4-JSON, SQLite-Migrationen, Blob-Roundtrip, Fehlerzuordnung.
- **Contract-Matrix:** Decimal, Date, DateTimeOffset, Int64, null, Composite Keys, `@odata.context`, `@odata.count`, `Location`, CRUD-Statuscodes; `Prefer` als Soll.
- **End-to-End:** Test-OData-Server starten → Requests aufzeichnen → Server stoppen → Replay; danach Mock CRUD, Filter, Snapshot und Reset.
- **Security-Tests:** Header-/Body-Redaction, Admin-Token und Admin-Transport, SSRF-/Upstream-Allowlist, kein mutierender Hybrid-Fallback ohne ausdrückliche Routenfreigabe.
- **Concurrency-Tests:** parallele Workspaces, Replay-Cursor-Isolation, atomare Mutationen und Resets.
- **Determinismus-Tests:** gleicher Ausgangszustand und gleiche Requests ergeben gleiche normalisierte Antworten und State-Hashes.

**Zusätzliche Contract-Gates:** Zwei Replay-Sessions in einem Workspace;
zwei Baselines eines Services mit unterschiedlicher Coverage;
`$select`-Teilentity bei GET/PATCH/PUT;
komprimierte und fragmentierte Redaction;
Crash zwischen Blob-Rename und DB-Commit;
idempotenter Import.

**Architektur-Gate:** Importgraph für `src/hexagon/{model,ports,services}`, `src/adapters/{driving,driven}` und `src/bootstrap` prüfen.
Erlaubte Kanten:

```text
model            → (keine inneren Pakete)
ports/inbound    → model
ports/outbound   → model
ports/support    → model
services         → model, ports/inbound, ports/outbound, ports/support
adapters/driving → model, ports/inbound, ports/support
adapters/driven  → model, ports/outbound, ports/support
bootstrap        → alle
```

Pakete innerhalb eines Service (etwa `services/record/capture`) gehören zur Gruppe `services`; die Regel „Services importieren einander nicht" gilt zwischen Services, nicht innerhalb eines Service.

Die konkreten Imports werden im Gate geprüft, nicht nur durch Konvention behauptet.
Integrationstest-Hilfen dürfen Adapter verwenden, sind aber nicht Bestandteil des ausgelieferten Programms.

---

## 7. Erweiterungspunkte

- **OData V2:** eigener Codec-/Version-Adapter ohne Verunreinigung des V4-Domain-Modells.
- **`$expand`, `$batch`, Actions, Functions:** separate AST-Knoten, Validatoren und Executors; nicht unterstützte Operationen bleiben im Replay nutzbar.
- **Weitere Storage-Backends:** Repository-Ports statt direkter SQLite-Kopplung.
- **Semantisches Matching:** optionaler OData-AST-Matcher neben konservativem HTTP-Matching.
- **Weitere Driving Adapter:** Testframework-Integration oder eingebetteter Server über Inbound Ports.
- **Coverage-Import:** explizite, paginierte Vollerfassung von EntitySets; getrennt von passivem Recording.

---

## 8. Hauptrisiken

| Risiko | Gegenmaßnahme |
|---|---|
| HTTP-/OData-Typen gelangen in den Core | Adapter-Mapping, Import-Gate, Contract-Tests |
| Importierte Teilmengen werden als vollständige Baseline interpretiert | Expliziter Import, Provenienz und Coverage-Evidenz |
| Replay und Mock vermischen Semantik | Getrennte Services und getrennte State-Modelle |
| Teilaufzeichnungen werden als vollständige Datenbasis interpretiert | Coverage-Status und strikter Default |
| Versehentliche Upstream-Schreiboperationen | Fallback-Allowlist, Schreib-Fallback nur mit ausdrücklicher Routenfreigabe, Security-Tests |
| Unredigierte Secrets in Recordings | Redaction vor Persistenz, Leak-Tests, Capture-Fehlerpolitik |
| Große Bodies führen zu Speicherüberlauf | Streaming, Body-Limits, content-adressierte Blob-Ablage |
| Nebenläufige Tests beeinflussen sich | Workspace-Isolation, transaktionale State-Änderungen |
| Snapshot und Blob-Storage driften auseinander | Hash-Prüfung, commit-after-blob, Recovery-Prüfung |
| Zu große OData-Vollständigkeitszusage | Expliziter unterstützter Query-Subset, 501 für nicht unterstützte Semantik |

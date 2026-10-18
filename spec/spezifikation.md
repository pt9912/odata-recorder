# Spezifikation — OData Recorder

**Bezug zum Lastenheft:** `spec/lastenheft.md` (LH-FA-01 bis LH-FA-18, LH-QA-01 bis LH-QA-07, LH-RB-01 bis LH-RB-04). Bei Konflikten gilt das Lastenheft. Diese technische Spezifikation konkretisiert, erweitert aber nicht den vereinbarten Umfang.

**Rolle:** Technisch verbindliches, fortschreibbares Spec-Stratum nach `ai-harness-course`. Priorität **Muss** ist implementierungsbindend; **Soll/Kann** sind technisch vorbereitet, aber keine Abnahmevoraussetzung der ersten Version. Technische Defaults und Schwellenwerte sind bis zur Freigabe Vorschläge.

---

## 1. Algorithmen und Datenflüsse

### LH-FA-01.a — Transparenter HTTP-Proxy

**Eingabe:** HTTP-Request an konfigurierten Service-Root; **Ausgabe:** HTTP-Response des Upstreams.

1. Gateway ermittelt den Service aus dem Service-Root-Pfad (`SPEC-054`), die Betriebsart aus der Startkonfiguration, validiert die Zielkonfiguration und erstellt eine Korrelations-ID. Ein Pfad ohne konfigurierten Service-Root ergibt `ODR-E5005` (HTTP 404).
2. Proxy leitet Methode, Pfad, Query, Body und End-to-End-Header weiter; Hop-by-Hop-Header werden nach HTTP-Regeln behandelt.
3. Upstream-Response wird an den Client gestreamt. Weiterleitungen (`Location`) und absolute OData-Links im Body werden in der ersten Version nicht umgeschrieben (`LINK_REWRITE=off`, `SPEC-092`): Record liefert und speichert sie unverändert, Replay liefert sie byteidentisch aus. Ein Client, der einem aufgezeichneten absoluten Link des Upstreams folgt, erreicht den Upstream am Recorder vorbei; Tests, die das ausschließen müssen, verwenden Mock-Links (`PUBLIC_BASE_URL`, `LH-FA-04.c`). Cookies (`Set-Cookie`) werden als Header redigiert (`LH-FA-14.a`).
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
- Recordings sind nach Abschluss logisch unveränderlich; neue Sessions werden append-only ergänzt. Eine ganze Session kann nur über die Retention-Regeln aus `LH-RB-04.a` entfernt werden.

### LH-FA-04.a — Deterministischer Replay-Cursor

**Eingabe:** Workspace, ausgewählte Session, Request, Replay-Policy. **Ausgabe:** gespeicherte Response oder Mismatch.

1. Match-Key nach LH-FA-05.a erzeugen.
2. Im Modus `sequential` unter den Exchanges der Session mit diesem Match-Key die mit der kleinsten `sequence_no` wählen, die nicht in `consumed` des dem Testlauf zugeordneten Cursors steht.
3. Response inklusive Status und ausgewählten End-to-End-Headern ausliefern; zeit- oder hostabhängige Header nur nach expliziter Policy transformieren.
4. Cursor erst nach erfolgreicher Auswahl atomar fortschreiben (`sequence_no` zu `consumed` hinzufügen, `position` = Anzahl der Elemente); bei gleichzeitigem Zugriff pro Cursor serialisieren. Beispiel: aufgezeichnet A(1), B(2), A(3); fragt der Client B, A, A, erhält er A(1), dann A(3).
5. `strict` vergleicht ausschließlich die Exchange mit `sequence_no = position + 1`; passt deren Match-Key nicht, ist das ein Fehler (`ODR-E5007`) und der Cursor bleibt stehen. `first` und `last` wählen die erste bzw. letzte passende Exchange und verändern den Cursor nicht. Die Strategie ist am Cursor gespeichert.

**Fehlermodi:** `ODR-E5001`, wenn die Session keinen Exchange mit diesem Match-Key enthält; `ODR-E5002`, wenn Exchanges mit diesem Key existieren, im Modus `sequential` aber alle verbraucht sind oder im Modus `strict` die Session erschöpft ist; `ODR-E5007`, wenn im Modus `strict` die nächste erwartete Exchange einen anderen Match-Key hat. Replay hat **keinen** Upstream-Zugriff.

### LH-FA-04.b — Explizite Zuordnung paralleler Replay-Clients

- Ein Replay-Request trägt einen vom Test-Harness gesetzten Workspace-Identifier und eine Recording-Session-Zuordnung. Die Transportabbildung erfolgt über `X-Recorder-Workspace`, `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session`; diese Steuerheader werden niemals an einen Upstream weitergereicht.
- Standard ist `assignment=explicit`. Fehlt eine Zuordnung oder ist sie mehrdeutig, wird `ODR-E5003` (HTTP 400) geliefert; eine unbekannte Workspace-ID ergibt `ODR-E7001`, eine unbekannte Recording-Session `ODR-E7003` (je HTTP 404), eine Zuordnung zu einem anderen Service `ODR-E5004` (HTTP 409). Ein implizites `ordered`-Assignment ist ausschließlich als opt-in für seriell ausgeführte Tests zulässig.
- Cursor-Identität: `(workspace_id, replay_session_id, recording_session_id)`. Zwei parallele Clients teilen nur dann absichtlich einen Cursor, wenn sie dieselbe explizite Replay-Session-ID verwenden; Cursor-Zugriffe sind atomar.
- Replay-Session-IDs werden beim Start eines Testlaufs vergeben. Das Löschen eines Workspaces entfernt dessen Cursor. Ein Workspace-Reset setzt sämtliche Cursor auf Position 0 und leert `consumed`.

### LH-FA-04.c — Adressierung von Service und Workspace

- **Service:** Jeder Service hat einen Service-Root-Pfad (`SERVICE_ROOT_PATH`, `SPEC-054`), der in `services.service_root` festgehalten wird. Ist er nicht gesetzt und genau ein Service konfiguriert, ist der Root `/`. Bei mehreren Services muss jeder Root ausdrücklich gesetzt sein, sonst verhindert der Start die Ausführung (`ODR-E2001`); damit ändert das Hinzufügen eines Services nie still den Root eines vorhandenen. Der Pfad nach dem Root ist der OData-Pfad.
- **Workspace (Mock, Replay):** über den Header `X-Recorder-Workspace`. Im Modus `mock` und `hybrid` ist er Pflicht; fehlt er, ergibt das `ODR-E5003`. `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` sind im Replay Pflicht und im Modus `hybrid`, sobald ein Replay-Fallback konfiguriert ist (`HYBRID_REPLAY_FALLBACK`, `SPEC-079`); fehlen sie dort, ergibt das `ODR-E5003`. Alle `X-Recorder-*`-Header werden vor jeder Weiterleitung entfernt.
- **Öffentliche Basis-URL:** `Location`, `@odata.id`, `@odata.context` und `@odata.nextLink` werden aus `PUBLIC_BASE_URL` (`SPEC-055`) plus Service-Root und typisiertem Key erzeugt, nie aus dem `Host`-Header des Clients. Ist `PUBLIC_BASE_URL` nicht gesetzt, leitet der Prozess sie aus `LISTEN` ab (`0.0.0.0` wird `127.0.0.1`); bei einem nicht-loopback `LISTEN` ohne gesetzte `PUBLIC_BASE_URL` verhindert der Start die Ausführung (`ODR-E2001`).

### LH-FA-05.a — Kanonisches HTTP-Matching

Match-Key = SHA-256 über eine eindeutig kodierte Tupelfolge:
`method`, `service_id`, `normalized_path`, sortierte Multimenge dekodierter Query-Paare, ausgewählte Vary-Header, Hash des redigierten Bodys (`LH-FA-14.a`).

- Query-Paare werden stabil nach Schlüssel und Wert sortiert; wiederholte Parameter bleiben erhalten.
- Ein Query-Paar wird als (dekodierter Schlüssel, dekodierter Wert) verglichen. `%XX` wird zu Bytes dekodiert (Hex-Groß-/Kleinschreibung ist gleich), die Bytes müssen gültiges UTF-8 sein. `+` bleibt das Zeichen `+` und wird nicht als Leerzeichen gelesen; `a+b` und `a%20b` sind damit verschieden (konservativ, nie falsche Gleichheit).
- Syntaktisch ungültige Kodierung (`%` ohne zwei Hex-Ziffern, kein gültiges UTF-8) ist kein Fehler: das Paar geht unverändert als Roh-Zeichenfolge in den Key und ist nur mit einem byteidentischen Paar gleich (`MATCH_INVALID_ENCODING=literal`, Default; `reject` liefert HTTP 400 mit `ODR-E5006`).
- Pfadsegmentgrenzen, Groß-/Kleinschreibung und OData-Literale bleiben erhalten.
- Bei JSON-Body normalisiert `MATCH_BODY_MODE=json-canonical` (opt-in) die Schlüsselreihenfolge; Default ist `bytes` (Byte-Hash des redigierten Bodys).
- Authentifizierungstokens und volatile Header (`LH-QA-02.a`) gehen nicht in den Key ein; als Vary-Header gelten die Header aus `MATCH_VARY_HEADERS` (`SPEC-061`, Default `Accept`, `OData-Version`, `Content-Type`, `If-Match`, `Prefer`).
- Eine semantische Gleichsetzung von `$filter`-ASTs ist **nicht** Bestandteil des Muss-Umfangs.

### LH-FA-06.a — OData-V4-Metadatenmodell

1. `$metadata` im Record-Modus als HTTP-Austausch erfassen.
2. EDMX/CSDL-XML (V4) parsen und in internes Modell überführen: Schemas, EntityTypes, Property-Typen, Nullable, Keys, EntitySets, NavigationProperties.
3. Referenzen und Vererbung auflösen; nicht unterstützte Typen/Annotations explizit markieren. Granularität: Das Modell wird mit `ODR-E6004` (Start verhindert bzw. Import abgewiesen; HTTP 422 in der Admin-API) abgelehnt bei nicht parsebarem CSDL, EDMX-Version ungleich 4.0, EntityType ohne Schlüssel, nicht auflösbarer Referenz oder einem Schlüssel-Property mit nicht unterstütztem EDM-Typ. Eine Property mit nicht unterstütztem EDM-Typ oder eine herstellerspezifische Annotation lässt das Modell gültig; sie wird im Validierungsbericht ausgewiesen (`unsupported_features`), Lesezugriffe darauf ergeben `501` mit `ODR-E6001`, schreibende `501` mit `ODR-E6002`.
4. Mock liefert die gespeicherten Metadaten konsistent mit der ausgewählten Service-Version aus.
5. Ohne gültiges Modell keine zustandsbehaftete Verarbeitung dieses EntitySets; reines Replay bleibt möglich.

### LH-FA-18.a — Expliziter Import in eine Mock-Baseline

**Eingabe:** Service-Modell, ausgewählte vollständige Recordings/Exchanges, Import-Policy. **Ausgabe:** unveränderliche Snapshot-ID, Coverage-Evidenz, Import-Report.

1. Session und Metadatenversion fixieren; nur erfolgreich abgeschlossene, redaktionsverträgliche OData-V4-JSON-Responses für einzelne Entities und Collections berücksichtigen.
2. `@odata.context`, EntitySet, Key-Typen und Property-Typen anhand des Modells validieren. `@odata.nextLink` markiert unvollständige Pagination, solange die komplette Kette nicht belegbar erfasst ist. `$top`, `$skip` und Filter ohne Vollständigkeitsbeweis erzeugen keine vollständige Coverage. **Vollständigkeitsbeweis (einziger Weg zu `complete`):** Die Session enthält zu einem EntitySet eine ungefilterte Collection-Abfrage (ohne `$filter`, `$top`, `$skip`, `$search`, `$apply`, `$expand`), alle Antworten der `@odata.nextLink`-Kette sind erfasst (`capture_status=complete`, Status 200, dieselbe Metadatenversion), und die letzte Seite trägt keinen `@odata.nextLink`. Eine solche Kette mit leerem `value` belegt ein leeres, vollständiges EntitySet. Sonst lautet die Coverage `partial` (Kette begonnen oder Teilabfrage beobachtet) oder `unknown` (nichts beobachtet).
3. Jedes Entity in typisierten kanonischen Key überführen. Einzel-Entity- und Collection-Responses werden als beobachtete Fakten behandelt; `null`-Properties bleiben von nicht selektierten Properties unterscheidbar.
4. Pro Key deterministisch deduplizieren: identische kanonische Payloads zusammenführen; widersprüchliche vollständige Werte oder inkompatible Typen sind Konflikte und stoppen den Import. Teilprojektionen dürfen nur widerspruchsfrei ergänzt werden; fehlende Felder sind kein `null`.
5. Provenienz pro Entity speichern (`session_id`, `exchange_id`, EntitySet, Metadatenhash, Importzeit) und Coverage-Evidenz pro EntitySet aggregieren.
6. Vollständiges Snapshot-Manifest in temporärem Blob ablegen und nach Integritätsprüfung in einer SQLite-Transaktion als neue Baseline publizieren. Bei Fehler bleibt die bisherige Baseline unverändert. Import ist explizit, kein Record-Nebeneffekt.

**Fehlermodi:** ungültiger Key, widersprüchliche Werte, beschädigte Antwort, fehlendes oder ungültiges Metadatenmodell, redigierte notwendige Key-Felder → `ODR-E6101` bzw. `ODR-E6004` (Modell). Redigierte Nicht-Key-Properties werden nicht übernommen und als unbekannt maskiert (`LH-FA-14.a`). **Komplexität:** O(n) Entitäten plus Indexkosten; Import speichert Zwischenstände begrenzt bzw. in Staging-Tabellen.

### LH-FA-07.a — Stateful CRUD

- `GET` liest aus Workspace-Snapshot plus lokalen Änderungen.
- `POST` validiert EntitySet, Key und Property-Typen, erzeugt einen Datensatz und liefert die vereinbarte OData-V4-Antwort. Der Key muss im Request-Body stehen; der Mock vergibt keine Schlüssel und erzeugt keine Zeitstempel in Entity-Daten (Determinismus, `LH-QA-02.a`). Fehlt ein Schlüssel, ergibt das `400`.
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
| `POST /Products` | `201` + Entity + `Location` (`204` nur bei unterstütztem `Prefer: return=minimal`) | Duplicate Key: `409`; Typfehler: `400` |
| `PATCH /Products(42)` | `204` oder `200` gemäß `Prefer: return=representation` | unbekannte Property: `400`; fehlende Entity: `404` bei vollständiger Kenntnis |
| `PUT /Products(42)` | Ersetzt die vollständig übermittelten beschreibbaren Properties; fehlende nullable Properties werden `null`, fehlende erforderliche Properties: `400` | Upsert per PUT ist nicht Teil der ersten Version; unbekannte Entity: `404` bei vollständiger Kenntnis, sonst `ODR-E6003` |
| `DELETE /Products(42)` | `204` | nicht vorhandene Entity: `404` bei vollständiger Kenntnis |

Ohne `Prefer`-Header gilt (Muss): `POST` → `201` mit Entity, `PATCH`/`PUT` → `204`. `Prefer: return=minimal` und `Prefer: return=representation` werden für mutierende Requests ausgewertet (Soll, `SPEC-049`); wird `Prefer` nicht unterstützt, gilt das Muss-Verhalten ohne Fehler. `Location` und Entity-ID werden aus Service-Root und typisiertem Key erzeugt; keine unvalidierte Client-Host-Injektion. `@odata.context`, `@odata.count` und `value` folgen dem ausgehandelten JSON-Metadatenlevel. ETags bleiben Soll-Umfang.

### LH-FA-08.b — EDM-Typen und Ausdruckssemantik

| Typ | Lesen/Schreiben | Vergleich/Sortierung | Grenze |
|---|---|---|---|
| `Edm.String`, `Edm.Boolean` | Muss | Muss | Stringvergleich ordinal, keine Locale-Collation |
| `Edm.Int32`, `Edm.Int64` | Muss | Muss | `Int64` ohne JS-Float-Rundung, OData-JSON-konforme Darstellung |
| `Edm.Decimal` | Muss | Muss | Dezimale Arithmetik ohne binären Float; Scale/Precision validieren |
| `Edm.Date`, `Edm.DateTimeOffset` | Muss | Muss | ISO-Formate; Offset auf Instant normieren, Date ohne Zeitzone |
| `Edm.Guid` | Muss | Gleichheit | kanonische UUID |
| `Edm.Binary`, `Edm.Duration`, `Edm.TimeOfDay`, `Edm.Geography*` | nicht im Muss-Umfang | — | explizit `501`, sofern keine reine Replay-Antwort |

**Null-Semantik** (OData V4.0 Part 2 §5.1.1.1; `null` ist ein Wert, kein Fehler):

- `eq`: `null` ist gleich sich selbst und nur sich selbst (`null eq null` ist `true`, `null eq 5` ist `false`); `ne` ist die Negation davon.
- `gt`, `lt`: sobald ein Operand `null` ist, `false`.
- `ge`, `le`: ist nur ein Operand `null`, `false`; sind beide `null`, `true`.
- `and`, `or`, `not` werten `null` als unbekannt: `null and false` = `false`, `null or true` = `true`, `not null` = `null`; alle anderen Kombinationen mit `null` ergeben `null`. Ein `null`-Ergebnis lässt den Datensatz nicht durch `$filter`; nur `true` tut das.
- Eine nicht beobachtete Property (Property-Maske, `LH-FA-18.a`) ist nicht `null`; ihre Verwendung im Filter ergibt `ODR-E6003`.

Nicht unterstützte Typ-/Operator-Kombinationen liefern `501`, syntaktisch oder typologisch ungültige Ausdrücke `400`. URL-Key-Syntax unterstützt einfache und benannte zusammengesetzte Keys; Literale werden typisiert und String-Escaping geprüft.

### LH-FA-08.a — Query-Interpreter

**Muss-Untermenge:** `$filter` mit `eq`, `ne`, `gt`, `ge`, `lt`, `le`, `and`, `or`, `not` und Klammern für primitive unterstützte EDM-Typen; `$select`, `$orderby`, `$top`, `$skip`, `$count`.

**Pipeline:** Parse → Typprüfung anhand Metadaten → Filter → stabile Sortierung (bei Gleichstand Entity-Key) → Count vor Pagination → Skip/Top → Projektion → OData-V4-JSON.

- Ungültiger Ausdruck: `400` mit OData-Fehlerobjekt. Dazu zählen `$top`/`$skip` mit negativem oder nicht ganzzahligem Wert; `$top=0` ist gültig und liefert eine leere `value`-Liste.
- `$count=true` liefert `@odata.count` (gefilterte Menge vor Paging); das Pfadsegment `/$count` liefert die Anzahl als `text/plain`; beide verwenden dieselbe Zählung.
- Nicht unterstützter, aber gültiger Ausdruck: expliziter `501`-Fehler `ODR-E6001`.
- `$expand` und `$batch` gehören zu LH-FA-16 (Soll); Actions und Functions gehören nicht zum ersten Abnahmeumfang (Lastenheft §5).
- Standardisierte Query-Fehler und Formatdetails sind durch OData-V4-Contract-Tests zu prüfen.

### LH-FA-09.a — Workspace-Isolation und Reset

Jeder Workspace besitzt `workspace_id`, `baseline_snapshot_id`, isolierten Entity-State und isolierte Replay-Cursor. Kein Write darf andere Workspaces beeinflussen. Reset ist eine atomare Wiederherstellung des Baseline-Zustands inklusive Cursor-Reset. Während Reset werden konkurrierende Schreibvorgänge serialisiert oder kontrolliert abgewiesen.

### LH-FA-09.b — Reset-Semantik

`reset(workspace)` ersetzt in **einer** Transaktion den Entity-State durch die Baseline, setzt **alle** Replay-Cursor des Workspaces auf Position 0 (`consumed` leer) und erhöht die Workspace-Revision; ein Abbruch lässt weder den neuen State mit altem Cursor noch das Umgekehrte zurück. Der Vorgang ist auch ohne bisherige Mutationen definiert und idempotent bezüglich des sichtbaren State/Cursor-Inhalts.

### LH-FA-10.a — Snapshots

Snapshots sind unveränderliche, benannte, versionierte Zustandsreferenzen. Erstellung erfasst konsistent ausschließlich den Entity-State; Replay-Cursor sind niemals Bestandteil eines Snapshots. Restore ersetzt den Entity-State atomar, erstellt eine neue Workspace-Revision und lässt sämtliche Replay-Cursor unverändert. Ein Restore mit unbekannter Snapshot-ID führt zu `ODR-E7002`, ohne dass ein Teilzustand übernommen wird; ein Snapshot eines anderen Service ergibt `ODR-E5004`, ein Snapshot, der zu einem anderen Workspace gehört, `ODR-E7002`. Ein Workspace-Snapshot gehört zu genau einem Workspace (`snapshots.workspace_id`) und trägt einen im Workspace eindeutigen Namen.

### LH-FA-11.a — Coverage und Teilmengen

Coverage je EntitySet: `complete`, `partial`, `unknown`. `complete` darf nur aus explizit verifizierter Vollerfassung einschließlich aller Continuation-Seiten entstehen; ein einzelnes `$top`-Resultat ist `partial`. Vor Query-Ausführung wird Coverage geprüft. Ein EntitySet ohne beobachtete Entities ist nicht automatisch `complete`. Bei `STRICT_COVERAGE=true` wird eine Collection-Abfrage über einen nicht vollständigen Bestand mit `ODR-E6003` abgewiesen. Einzel-Entity-Lookups dürfen nur dann eine definitive `404` liefern, wenn das Nichtvorhandensein abgesichert ist; sonst Coverage-Fehler.

**Coverage-Kennzeichnung:** Jede Mock-Antwort auf eine Collection-Abfrage trägt den Header `X-Recorder-Coverage: complete|partial|unknown`. Bei `STRICT_COVERAGE=false` wird eine Collection-Abfrage über einen nicht vollständigen Bestand aus den bekannten Entities beantwortet, und der Header nennt `partial` bzw. `unknown`; der Antwort-Body bleibt unverändert OData-V4-konform. Der Header wird nie entfernt (Lastenheft §1: Unvollständigkeit bleibt erkennbar).

### LH-FA-12.a — Moduswahl und Fallback

`record`, `replay`, `mock`, `hybrid`. Der Modus ist eine Startkonfiguration, nicht per Admin-API live umschaltbar. Hybrid-Policy ist explizit konfiguriert. Default: `mock` → Fehler; optional `mock` → `replay` → Fehler nur für definierte Kategorien (`unsupported`, `coverage_insufficient`), niemals für fachliche `404`, `409` oder `412`; **kein** Upstream-Fallback. Upstream-Fallback nur per Allowlist (`FALLBACK_ALLOW_HOSTS`, `SPEC-062`) für lesende Methoden und autorisierte Zielhosts. Für mutierende Methoden ist Upstream-Fallback verboten, es sei denn, eine ausdrückliche Routenfreigabe (`FALLBACK_WRITE_ALLOW`, Liste aus Methode und Pfadmuster, Default leer) deckt den Request; ohne Eintrag wird nichts an den Upstream gesendet. Unbekannte Funktionen werden nicht stillschweigend als erfolgreich behandelt.

### LH-FA-12.b — Hybrid-Entscheidungsmatrix

| Ergebnis Mock | Default | Optionaler Replay-Fallback | Upstream-Fallback |
|---|---|---|---|
| Erfolg | Antwort | Nein | Nein |
| Fachliche `404` / `409` / `412` | Antwort/Fehler beibehalten | Nein | Nein |
| Coverage unvollständig | `ODR-E6003` | Nur explizite Policy | GET/HEAD + Allowlist; mutierend nur mit `FALLBACK_WRITE_ALLOW` |
| Nicht unterstützte Operation | `ODR-E6001` | Nur explizite Policy | GET/HEAD + Allowlist; mutierend nur mit `FALLBACK_WRITE_ALLOW` |
| Ungültige Anfrage | HTTP 400 | Nein | Nein |
| Replay-Miss nach opt-in | `ODR-E5001` | — | Nur gesonderte GET/HEAD-Allowlist (`FALLBACK_ALLOW_HOSTS`) |

### LH-FA-13.a — Management und Diagnose

Separater Admin-Listener mit Endpunkten für Sessions, Exchanges, Coverage, Workspaces, Reset, Snapshot, Import, Diagnose und Health (`SPEC-005`). Änderungen sind auditierbar. Jede HTTP-Response enthält strukturierte Fehlerkennung und Korrelations-ID.

**Admin-Authentisierung (unabhängig vom Binding):** Jeder Admin-Endpunkt außer `GET /healthz` verlangt `Authorization: Bearer <token>`. Das Token stammt aus `ADMIN_TOKEN_FILE` (`SPEC-056`). Existiert die Datei beim Start nicht, erzeugt der Prozess ein zufälliges Token (mindestens 256 Bit), legt es dort mit Rechten `0600` ab und schreibt es nie in ein Log. Der Vergleich erfolgt zeitkonstant. Fehlendes oder falsches Token ergibt `ODR-E2101` (HTTP 401) ohne Zustandsänderung; ein Audit-Eintrag ohne Tokenwert wird geschrieben. Der Admin-Listener bindet standardmäßig an `127.0.0.1`; in einem Container ist `ADMIN_LISTEN=0.0.0.0:8081` mit TLS zulässig. Eine Token-Datei, die vorhanden, aber nicht lesbar ist, nach Entfernung von Zeilenumbruch und Leerraum weniger als 32 Zeichen enthält oder auf POSIX-Systemen Gruppen- oder Weltrechte trägt, verhindert den Start (`ODR-E2001`); ein leeres Token wird nie akzeptiert. Fehlt die Datei und ist `DATA_DIR` nicht beschreibbar, verhindert das den Start ebenso (`ODR-E2001`). **Transportschutz:** Ist `ADMIN_LISTEN` nicht loopback, verlangt der Start `ADMIN_TLS_CERT`/`ADMIN_TLS_KEY` (`SPEC-089`) und verhindert sonst die Ausführung (`ODR-E2001`); nur `ADMIN_ALLOW_PLAINTEXT=true` (`SPEC-090`) erlaubt unverschlüsselten Zugriff, mit einer Warnung im Log. Die Token-Datei liegt im Datenvolumen und ist Teil von dessen Sicherungen; das Betriebsprofil nennt sie als schützenswert. **Status:** Eine Session hat den Status `open`, `complete` oder `incomplete`. Eine Session ist `open`, solange ihre Aufzeichnung läuft. Sie wird `complete`, wenn die Aufzeichnung regulär endet (Beenden der Aufzeichnung, geordneter Shutdown nach `LH-FA-02.c`) und alle Exchanges `capture_status=complete` tragen, sonst `incomplete`. Beim Start setzt die Wiederherstellung jede noch `open` stehende Session vor der Annahme von Requests auf `incomplete`. Replay und Import gegen eine `open`-Session sowie das Löschen einer `open`-Session weist `ODR-E7009` (HTTP 409) ab; eine `incomplete`-Session ist zulässig, nur ihre `complete`-Exchanges sind replayfähig (`SPEC-045`). Zustandswechsel stehen nur in dieser Richtung (`open` → `complete`|`incomplete`); ein Workspace wird durch `revision`, `baseline_snapshot_id` und `created_at` beschrieben.

### LH-FA-02.b — Streaming- und Capture-Lifecycle

- Der HTTP-Adapter besitzt und schließt eingehende Request- und ausgehende Response-Streams; er gibt sie nur als geliehene, kontextgebundene Streams an Transport/Capture weiter. Der Capture-Writer hat `Write`, `Commit` und `Abort`; genau einer der beiden Abschlüsse ist zulässig.
- Transport und Capture werden mit begrenztem Puffer entkoppelt. Backpressure des Capture-Writers darf nicht unbeschränkt Speicher belegen; bei Größenlimit/Fehler wird Capture abgebrochen und der Proxy-Stream gemäß Transportvertrag fortgesetzt.
- Vor dem ersten gesendeten Response-Byte kann ein Transportfehler als Gateway-Status beantwortet werden. Nach dem ersten gesendeten Byte wird der Client-Status **nicht** rückwirkend geändert; der Capture-Status ist `incomplete`, Diagnose wird geschrieben.
- Redaction geschieht vor dauerhafter Ablage; nicht redigierbare Bodies werden verworfen. Ein redigierter Response-Body ist als `replay_equivalence=semantic_or_limited` gekennzeichnet und nicht als bytegenau garantiert.
- Der Capture-Limit-Test sendet einen Body oberhalb von `MAX_CAPTURE_BODY` (`SPEC-013`), verifiziert begrenzten Speicherverbrauch, korrektes Live-Streaming und einen nicht replayfähigen unvollständigen Capture.

### LH-FA-02.c — Beenden des Prozesses

Bei SIGTERM/SIGINT schließt der Prozess die Listener, weist neue Requests mit HTTP 503 ab, lässt laufende Exchanges bis `SHUTDOWN_TIMEOUT` (`SPEC-059`) abschließen, finalisiert offene Capture-Vorgänge und schließt die Repositories. Nach Ablauf der Frist werden noch offene Exchanges mit `capture_status=incomplete` abgeschlossen; ein erzwungener Abbruch erzeugt nie einen als `complete` markierten Teil-Exchange (Crash-Fälle: `SPEC-048`).

### LH-FA-14.a — Redaction und Datensicherheit

Default-Redaction der Header: `Authorization`, `Proxy-Authorization`, `Cookie`, `Set-Cookie`, `X-API-Key`, `X-Auth-Token`, `X-CSRF-Token`; `REDACT_HEADERS` (`SPEC-060`) ergänzt die Liste, verkürzt sie nicht. Zusätzlich werden Query-Parameter und JSON-Felder redigiert (`REDACT_QUERY_PARAMS`, `REDACT_JSON_PATHS`, `SPEC-084`/`SPEC-085`, beide mit sicheren Defaults; die Listen ergänzen sich, ein Eintrag lässt sich nur durch Änderung der Konfiguration entfernen). Auch `raw_query` wird redigiert gespeichert. Redigiert wird immer durch den festen Marker `[REDACTED]` (Header- und JSON-String-Wert); der Marker ist das einzige Platzhalterformat.

Header-Redaction beschränkt die Replayfähigkeit nicht (`SPEC-045`): redigierte Header werden im Replay nicht ausgeliefert, nie mit Marker. Reguläre Logs enthalten weder Query-Strings noch Bodies noch Header-Werte. **Geltungsbereich der Redaction:** Sie erfasst Header, `raw_query` und die konfigurierten JSON-Pfade des Bodys. Query-Werte in `Location` und `Content-Location` werden wie `raw_query` behandelt (`REDACT_QUERY_PARAMS`); Link-Annotationen im Body (`@odata.nextLink`, `@odata.id`) nur über `REDACT_JSON_PATHS`. Der Pfad (`exchanges.path`) wird nicht redigiert: Geheimnisse im Pfad sind nicht unterstützt und gehören nicht in aufgezeichnete URLs.

**Redaction und Match-Key:** Der Match-Key (LH-FA-05.a) wird ausschließlich über die redigierte Darstellung gebildet. Dieselbe Redaction-Funktion läuft beim Record und beim Replay über den eingehenden Request, bevor der Key berechnet wird; ein Hash eines unredigierten Geheimnisses wird nie persistiert, und ein Replay-Request mit demselben Geheimnis trifft denselben Key. Grenze: Requests, die sich nur im redigierten Teil unterscheiden, haben denselben Key; das ist gewollt, weil Geheimnisse nicht unterscheidbar sein dürfen. Solche Exchanges tragen `key_redacted=true`; trifft ein Replay einen davon, trägt die Antwort den Header `X-Recorder-Match-Redacted: true` und das Log eine Warnung, sodass das Zusammenfallen nicht unbemerkt bleibt.

**Redaction und Import:** Ein Property-Wert mit dem Marker oder unter einem redigierten JSON-Pfad gilt im Import als nicht beobachtet (Property-Maske `unknown`, nicht `null`, nicht als Nutzdaten übernommen). Ein redigiertes Key-Feld ergibt `ODR-E6101`. Unbekannte sensible Payloads lassen sich nicht automatisch sicher erkennen: Capture muss deaktivierbar sein, und ein expliziter Datenklassifizierungsprozess ist notwendig. Secrets ausschließlich aus sicherer Konfiguration, nicht in Logs oder Exporten. Replay kann redigierte Auth-Daten nicht originalgetreu wiederherstellen.

### LH-FA-15.a — Optimistische Concurrency (Soll)

Entity-Versionen als ETag; `If-Match`/`If-None-Match` entsprechend dem unterstützten OData-Vertrag auswerten. Konflikt ergibt `412`; Versionserhöhung erfolgt atomar. Ohne aktivierte ETag-Unterstützung wird kein scheinbar gültiger Concurrency-Schutz behauptet.

### LH-FA-16.a — Erweiterte OData-Funktionen (Soll)

Erweiterungspunkte für `$expand` und `$batch` mit atomaren Changesets. Actions und Functions sind nicht Teil des ersten Abnahmeumfangs (Lastenheft §5) und bleiben Erweiterungspunkt der Architektur. Jede Funktion benötigt eigenen Parser-, Validator-, Executor- und Serializer-Contract-Test. Unbekannte Operationen im Mock-Modus liefern `501`; Replay kann sie unabhängig vom Parser bedienen.

### LH-FA-17.a — OData V2 (Kann)

Versionierter Adapter für V2-Metadaten, Payloads und Query-Semantik; keine automatische V2/V4-Vermischung in einem Service-Modell. Nicht Bestandteil der ersten Abnahme.

### LH-QA-01.a — Replay-Performance

Messaufbau gemäß `LH-QA-01` für Antworten allein aus lokalem Speicher (Replay und Mockserver; Record und Pass-through sind ausgenommen): lokaler Speicher, 2 vCPU, 4 GiB RAM, 10.000 Exchanges, Bodies bis 64 KiB, 20 Requests/s, 10 Minuten nach Warm-up. Gemessen wird die Zeit von Request-Eingang bis zum ersten Response-Byte im Prozess (Replay: Span `odata.replay.match`, `SPEC-028`; Mockserver: Span `odata.mock.query`, `SPEC-029`). Für den Mockserver gelten dieselben Ressourcen und Raten mit einem Workspace aus 10.000 Entities über mehrere EntitySets (Payloads bis 64 KiB) und einer Mischung aus Einzel-Lookups und Collection-Abfragen mit `$filter`/`$top`. Der Match-Key-Index `(session_id, match_key)` (`SPEC-004`) muss ohne Tabellenscan nutzbar sein. Der Zielwert p95 ≤ 100 ms ist vorläufig; die Bestätigung durch den fachlichen Owner steht aus (`LH-QA-01`).

### LH-QA-02.a — Determinismus-Randbedingungen

- **Volatile Header** (`VOLATILE_HEADERS`, `SPEC-063`) werden im Vergleich der normalisierten Antworten ignoriert und gehen nicht in den Match-Key ein; alle übrigen Header sind relevant.
- Der Mock vergibt keine Schlüssel und erzeugt keine Zeitstempel in Entity-Daten (`LH-FA-07.a`); ETags (Soll) leiten sich aus dem Versionszähler der Entity ab.
- Stabile Sortierung bei Gleichstand nach Entity-Key (`LH-FA-08.a`); Collections ohne `$orderby` werden nach Entity-Key geliefert.
- Zeit und serverseitig erzeugte IDs (Korrelations-ID, Workspace-ID, Snapshot-ID) kommen aus `Clock`/`IDGenerator` und sind in Tests ersetzbar; Replay-Session-IDs und `import_id` vergibt der Client.

### LH-QA-03.a — Isolation und Konsistenz

Messaufbau: mindestens 20 parallele Workspaces, jeder mit eigener Mutationsfolge; kein Workspace beobachtet Zustand oder Cursor eines anderen. Nach regulärem Neustart sind abgeschlossene Mutationen und Snapshots vorhanden; Fehler-Injektion nach `SPEC-048`. Pro Workspace werden Schreibvorgänge serialisiert (`LH-FA-07.a`), pro Cursor Zugriffe atomar (`LH-FA-04.b`).

### LH-QA-04.a — Datensicherheit

Der Nachweis besteht aus zwei Tests (`make test-security`, `SPEC-041`): Canary-Secrets in Headern, Query-Parametern und Body-Feldern werden aufgezeichnet und anschließend in allen Speicher- und Log-Artefakten gesucht, ohne Treffer. Ein Netzwerkzugriffstest prüft, dass die Admin-Endpunkte außer `GET /healthz` ohne Token mit `ODR-E2101` abgewiesen werden und dass der Admin-Listener nur über Loopback oder mit TLS erreichbar ist (`LH-FA-13.a`). Die Maßnahmen stehen in `LH-FA-13.a`, `LH-FA-14.a` und `SPEC-046`/`SPEC-053`.

### LH-QA-05.a — Fehlerklassen

Unterscheidbare Klassen und ihre Codes: fehlender Replay-Treffer `ODR-E5001`/`ODR-E5002`, ungültige Metadaten `ODR-E6004`, nicht unterstützte Query-Option `ODR-E6001`, unvollständige Coverage `ODR-E6003`, Zuordnungsfehler `ODR-E5003`/`ODR-E5004`/`ODR-E7001`/`ODR-E7003`. Jede Antwort trägt Code, HTTP-Status und Korrelations-ID, aber keine Credentials oder unredigierten Bodies.

### LH-QA-06.a — Container und Neustart

Der Dienst startet ohne externe Infrastruktur aus einem Container-Image mit einem Datenvolumen (`DATA_DIR`); ein frisches, leeres Verzeichnis ist ein gültiger Startzustand. Der Neustarttest startet den Container neu und prüft Recordings, Workspaces und Snapshots. Gleiche Konfiguration und Eingabedaten ergeben gleiches Verhalten (`LH-QA-02.a`).

### LH-QA-07.a — Protokolltreue

Der Record-Pfad leitet Status, Body und relevante End-to-End-Header unverändert weiter (`LH-FA-01.a`). Der Vergleichstest ruft einen Referenzdienst direkt und über den Proxy auf und vergleicht Status, Body und Header ohne `VOLATILE_HEADERS`. Unbekannte OData-Erweiterungen verhindern die HTTP-Aufzeichnung nicht; sie werden als `unsupported_features` markiert (`LH-FA-06.a`).

### LH-RB-01.a — OData-Kompatibilität

Der unterstützte Teilumfang ist `SPEC-049` zusammen mit `LH-FA-08.a`/`LH-FA-08.b`. Alles außerhalb liefert `501` (`ODR-E6001`/`ODR-E6002`) und erscheint nie als unterstützt. Nachweis: Referenzanfragen und Negativfälle (`make test-contract`, `SPEC-042`).

### LH-RB-02.a — HTTP/TLS-Konfiguration

Upstream: `UPSTREAM_TLS_VERIFY=true` (`SPEC-057`) prüft Zertifikatskette und Hostnamen; `UPSTREAM_CA_FILE` ergänzt die Vertrauensanker. Die Prüfung wird nur durch `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` (`SPEC-087`) abgeschaltet; dann schreibt der Start eine Warnung ins Log. Listener: `LISTEN_TLS_CERT`/`LISTEN_TLS_KEY` (`SPEC-088`) und `ADMIN_TLS_CERT`/`ADMIN_TLS_KEY` (`SPEC-089`); ohne sie sprechen die Listener HTTP, was auf Loopback der Default ist und für den Admin-Listener sonst `LH-FA-13.a` unterliegt. Ein Zertifikatsfehler zum Upstream ergibt `ODR-E4001` (HTTP 502). Nachweis: Integrationstests für HTTP, HTTPS und Zertifikatsfehler.

### LH-RB-03.a — Linux und CI

Build und Tests laufen in Containern (Docker-only, `AGENTS.md` §3.1); das Binary ist auch nativ unter Linux startbar. Nachweis: reproduzierbarer Container-Build und End-to-End-Test (`SPEC-037`).

### LH-RB-04.a — Zugriff, Redaction und Aufbewahrung

Zugriffskontrolle: `LH-FA-13.a`. Redaction: `LH-FA-14.a`. Aufbewahrung: Sessions werden nie automatisch gelöscht, solange `RETENTION_MAX_AGE=0` (`SPEC-058`) gilt. Ist ein Alter gesetzt, löscht der Start Sessions, die älter sind; eine Session, die als Quelle einer Baseline geführt wird (`baseline_sources`), überspringt der Lauf mit einer Warnung im Log, ohne den Start abzubrechen. Außerdem löscht `DELETE /__admin/sessions/{id}` eine Session samt Exchanges und der zugehörigen Replay-Cursor in einer Transaktion; ist die Session Quelle einer Baseline, wird sie mit `ODR-E7004` (HTTP 409) abgewiesen, bevor etwas gelöscht wird. Blobs, die danach weder von einem Exchange noch von einem Snapshot referenziert werden, entfernt die Bereinigung unmittelbar nach dem Commit (best effort) und in jedem Fall beim nächsten Start, bevor der Dienst Requests annimmt; spätestens dann sind die Daten vom Datenträger entfernt (Löschzeitpunkt für `LH-RB-04`). Löschungen werden im Audit-Log festgehalten. Das Löschen ändert die logische Unveränderlichkeit abgeschlossener Sessions (`LH-FA-03.a`) nicht, weil es ganze Sessions entfernt, nie Inhalte ändert.

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
      "sequence_no": 1,
      "started_at": "2026-10-18T10:00:00Z",
      "finished_at": "2026-10-18T10:00:00Z",
      "method": "GET",
      "path": "/Products",
      "raw_query": "$top=10",
      "match_key": "sha256:...",
      "key_redacted": false,
      "request": {"headers": {}, "body": {"state": "absent", "blob": null}},
      "response": {"status": 200, "headers": {}, "body": {"state": "present", "blob": "sha256:..."}},
      "capture_status": "complete",
      "redaction": {"request": "none", "response": "none"},
      "body_integrity": "verified",
      "replay_equivalence": "byte"
    }
  ]
}
```

Persistierte `headers` und `raw_query` sind redigiert (`LH-FA-14.a`); rohe Credentials gehören nicht in das Schema. Blob-Hash bezieht sich auf die **gespeicherte redigierte** Darstellung.

`body.state` gilt je Richtung und unterscheidet `absent` (der Austausch hatte keinen Body), `present` (Blob vorhanden) und `discarded` (Body bewusst nicht persistiert, `blob` ist `null`; `capture_status` ist dann `failed` oder `incomplete`). Die Felder entsprechen den Spalten des Schemas (`SPEC-004`): `sequence_no`, `started_at`, `finished_at`, `request_body_state`/`response_body_state`, `request_redaction_status`/`response_redaction_status` (Feld `redaction`), `key_redacted`, `capture_status`, `body_integrity`, `replay_equivalence`. `redaction.*=unsafe` bedeutet, dass eine erforderliche Body-Redaction nicht abgeschlossen werden konnte; der Body wird verworfen. Ist `redaction.request=unsafe`, lässt sich der Match-Key nicht bilden, und der Exchange ist `capture_status=failed`. `key_redacted=true` hält fest, dass die Redaction Eingaben des Match-Keys verändert hat (`LH-FA-14.a`).

### SPEC-002 — OData-Modell

`ServiceModel{version, namespaces, entity_types, entity_sets, navigation_bindings}`; `EntityType{qualified_name, keys[], properties[], navigation_properties[]}`; `Property{name, edm_type, nullable}`. Typ- und Namensvalidierung vor State-Zugriff.

### SPEC-003 — Workspace und Entity

`Workspace{id, service_id, baseline_snapshot_id, baseline_kind, revision, created_at}` (`baseline_kind` ist immer `baseline` und stützt den Fremdschlüssel auf den Snapshot); der Status eines Workspace in der Verwaltung ist das Tripel aus `revision`, `baseline_snapshot_id` und `created_at`; `EntityState{workspace_id, entity_set, entity_key, payload_json, known_properties, entity_completeness, entity_version, deleted}`; `ReplayCursor{workspace_id, replay_session_id, recording_session_id, service_id, strategy, position, consumed}` (`service_id` stützt die Fremdschlüssel auf Workspace und Session); `consumed` ist die Menge der verbrauchten `sequence_no`, `position` ihre Anzahl. Zusammengesetzte Schlüssel werden als kanonisch typisierte Tupel serialisiert, nicht durch naive String-Konkatenation. **Zustandsrepräsentation:** `entities` hält den vollständigen Entity-State eines Workspace, kein Delta über der Baseline. Die Anlage eines Workspace materialisiert den Baseline-Snapshot in `entities`; `deleted=1` ist ein Tombstone, der ein definitives `404` belegt. Ein Workspace-Snapshot (`state_blob`) enthält den vollständigen Entity-State einschließlich Tombstones; Reset ersetzt den Entity-State durch den der Baseline, Restore durch den des Snapshots.

### SPEC-004 — Normatives Schema für d-migrate 1.7.1

Das neutrale d-migrate-Schemamodell (YAML-Quelle des Repos) ist die einzige Quelle der Tabellen, Spalten, Schlüssel, Indizes, UNIQUE-, CHECK- und **zusammengesetzten FOREIGN-KEY-Constraints**. d-migrate **1.7.1** versteht `tables.<table>.constraints` mit `type: foreign_key`, `columns: [...]` und `references: {table, columns: [...], on_delete: ...}` sowie `type: unique` und `type: check` mit `expression`. Das Schemaformat selbst bleibt `schema_format: "1.0"`; die Anwendungsschema-Version (`version` in `tools/schema/schema.yaml`) wird unabhängig von der Dokumentversion des Lastenhefts geführt; ihre Änderung folgt jeder Änderung des Schemas.

Normative DB-Invarianten:

1. `snapshots(id,service_id)` und `snapshots(id,service_id,snapshot_kind)` sind UNIQUE; `workspaces(baseline_snapshot_id,service_id,baseline_kind)` referenziert das Tripel, `baseline_kind` ist per Check immer `baseline`. Ein Workspace kann damit weder die Baseline eines anderen Service noch einen `workspace`-Snapshot als Baseline verwenden. Ein `workspace`-Snapshot trägt `workspace_id` und `name` (Check `ck_snapshot_scope`), der Name ist je Workspace eindeutig; `workspace_id` ist kein Fremdschlüssel (Zirkel mit `workspaces`), sondern wird transaktional geprüft, und das Löschen eines Workspace löscht seine Snapshots in derselben Transaktion.
2. `imports(baseline_snapshot_id,service_id)` referenziert dasselbe Paar; `baseline_snapshot_id` ist Pflicht und eindeutig: eine Zeile in `imports` entsteht in derselben Transaktion wie die Baseline (nur erfolgreiche Imports; ein fehlgeschlagener Import hinterlässt keine Zeile und wird im Audit-Log festgehalten, daher darf seine `import_id` erneut verwendet werden). `baseline_sources(baseline_snapshot_id,session_id)` hält die Quell-Sessions abfragbar fest; der Fremdschlüssel auf `sessions` ist `RESTRICT`. `snapshot_kind=baseline` wird zusätzlich **innerhalb derselben Transaktion** geprüft; ein gewöhnlicher `workspace`-Snapshot darf nicht stillschweigend promotet werden.
3. `sessions(id,service_id)` und `workspaces(id,service_id)` sind UNIQUE; `replay_cursors(recording_session_id,service_id)` und `(workspace_id,service_id)` erzwingen dieselbe Service-Identität; ein Cursor wird mit seiner Recording-Session gelöscht (`CASCADE`). Die Replay-Session ist ein clientdefinierter, workspacegebundener Identifikator, nicht die Recording-Session.
4. `exchanges(session_id,sequence_no)` ist UNIQUE; Index `(session_id,match_key)` unterstützt Matching. Enumerationen für Capture/Redaction/Body/Replay/Coverage und numerische Werte sind als `check`-Constraints in der YAML-Datei enthalten, ebenso die Zustandsinvarianten eines Exchange: `sequence_no >= 1`; ein verworfener Body setzt `capture_status` ungleich `complete` voraus; `redaction=unsafe` verlangt einen verworfenen Response-Body und `replay_equivalence=none`, auf der Request-Seite `capture_status=failed`. `baseline_sources` und `coverage` führen `baseline_kind` und referenzieren damit nur Snapshots der Art `baseline`. Nur im Anwendungstest (`make test-integration`) belegt werden Invarianten, die ein neutrales Check nicht ausdrückt: `position` entspricht der Anzahl in `consumed_json`, die JSON-Spalten enthalten gültiges JSON, und das Löschen eines Workspace löscht seine Snapshots in derselben Transaktion (kein Fremdschlüssel).
5. Jede SQLite-Verbindung setzt `PRAGMA foreign_keys=ON`. Alle mutierenden UseCases prüfen fachliche Invarianten transaktional und dürfen Datenbank-Constraint-Fehler nicht ignorieren. `PRAGMA foreign_key_check` und `integrity_check` sind verpflichtende Gates.

**Migrationsregel:** Das DDL wird ausschließlich mit gepinntem `d-migrate:1.7.1` generiert; es existiert keine parallele manuell gepflegte Tabellen-DDL. CI muss `schema validate`, `schema generate --target sqlite`, DDL-Ausführung, Negativtests und die vollständige Migrationskette auf einer Kopie bestehender Datenbanken durchführen. Als Gate verdrahtet ist `make schema-check`: `schema validate`, `schema generate --target sqlite`, DDL-Ausführung in SQLite, die Negativtests der Constraints und `foreign_key_check`/`integrity_check`; `make schema-validate` und `make schema-generate` sind Werkzeuge ohne Gate-Eigenschaft. Nicht belegt ist die Migrationskette auf bestehenden Datenbanken; sie entsteht mit der zweiten Schemaversion.

**Versioniertes Snapshot-Manifest:** `manifest_version=1`, `snapshot_kind`, `service_id`, `metadata_hash`, `coverage[]`, `entities[]`, `provenance[]`, `source_session_ids[]`; pro importierter Property werden Quelle (`session_id`, `exchange_id`), Projektion und Hash festgehalten. Canonical JSON und SHA-256 sichern die Integrität. Unveränderliche Baselines dürfen nicht überschrieben werden. Snapshot-Promotion ist ein eigener validierter Vorgang; Coverage wird nie implizit auf `complete` gesetzt.

### SPEC-043 — Import- und Provenienz-Schema

`ImportManifest{import_id, service_id, metadata_hash, source_session_ids[], created_at, entity_count, baseline_snapshot_id, coverage[]}`; `ImportedEntity{entity_set, typed_key, payload, selected_properties[], provenance[]}`. Snapshot-Blobs enthalten eine versionierte kanonische JSON-Darstellung und ihren SHA-256-Hash. `baseline_snapshot_id` wird vor Workspace-Erzeugung auf Existenz und `snapshot_kind=baseline` geprüft; SQLite-Fremdschlüssel und Migrationstests sichern referenzielle Integrität. Garbage Collection löscht nur Blobs, die weder von Exchange noch Snapshot referenziert werden; Crash nach Blob-Rename vor DB-Commit darf höchstens verwaiste Blobs erzeugen.

### SPEC-044 — Baseline-Coverage und Entity-Vollständigkeit

Jeder Mock-Read löst zunächst `workspace.baseline_snapshot_id` auf und liest Coverage ausschließlich für `(baseline_snapshot_id, entity_set)`. Lokale Writes ändern diese Baseline-Evidenz nicht. Ein neuer Snapshot kann eine neue Baseline erzeugen, aber `complete` wird nicht allein aus lokalem Schreiben abgeleitet. Bei `STRICT_COVERAGE=true` darf eine nicht nachweislich vollständige Collection keine scheinbar vollständige Ergebnismenge liefern.

Für Entity-Reads gilt: Projektionen ausschließlich bekannter Properties sind zulässig, wenn Identität und Key verifiziert sind; vollständiges `GET` auf `partial` ergibt `ODR-E6003`. `PATCH` auf einer partiellen Entity ist im MVP **abzuweisen** (`ODR-E6003`, HTTP 422); `PUT` darf nur als explizite vollständige Ersetzung mit validierten Pflichtfeldern und atomarem State-Update ausgeführt werden, ansonsten `ODR-E6003`. Keine unbekannte Property wird durch `null` ersetzt. Tests: `$select`, Null-vs-missing, Pflichtfeld, PATCH/PUT, zwei Baselines eines Services.

### SPEC-045 — Capture-/Replay-Fähigkeitsmatrix

| Capture | Response-Body-Redaction | Body-Integrität | Default Replay |
|---|---|---|---|
| `complete` | `none` | `verified` oder `absent` | Byte-äquivalent zulässig |
| `complete` | `applied` | `verified` oder `absent` | Nur explizit aktiviertes `semantic_or_limited` (`REPLAY_REDACTED_BODIES=true`) |
| `complete` | `unsafe` | beliebig | Verboten |
| `incomplete`/`failed` | beliebig | beliebig | Verboten |
| beliebig | beliebig | `invalid` | Verboten |

Die Spalte betrifft ausschließlich die **Redaction des Response-Bodys**, weil nur dieser ausgeliefert wird. **Verbotener Treffer:** Ein Exchange, den diese Matrix mit „Verboten" belegt, ist kein Kandidat: Er wird weder ausgeliefert noch im Cursor verbraucht. Findet das Matching nur solche Exchanges, antwortet der Dienst mit `ODR-E5008` (HTTP 422, kein Cursor-Fortschritt); findet es keinen Exchange mit dem Key, gilt `ODR-E5001`. Im Modus `sequential` wird der nächste replayfähige Kandidat gewählt; im Modus `strict` ist ein nicht replayfähiger nächster Exchange (`sequence_no = position + 1`) ein `ODR-E5008`. Eine Redaction des Request-Bodys (`redaction.request=applied`), von Query-Parametern oder von Headern schränkt die Replayfähigkeit nicht ein; sie wirkt über den Match-Key (`LH-FA-14.a`). Die Redaction von Headern (`Authorization`, `Cookie`, `Set-Cookie` usw., `LH-FA-14.a`) schränkt die Replayfähigkeit nicht ein: Status und Body werden byteidentisch ausgeliefert, redigierte Header werden weggelassen und nie mit dem Marker `[REDACTED]` ausgeliefert. Default ist **strict byte replay** für Exchanges ohne Body-Redaction; Exchanges mit Body-Redaction sind nicht stillschweigend replayfähig. Bei aktivierter eingeschränkter Policy sind Antwortfelder mit Redaction als eingeschränkt gekennzeichnet; Auth-Credentials können nicht rekonstruiert werden.

### SPEC-046 — Sicheres Streaming-Capture

1. Proxy-Transport und Capture-Tee sind entkoppelt. Rohbytes dürfen nur in flüchtigem RAM bis zur vorgegebenen Grenze liegen, niemals in Klartext in temporären Dateien, Swap-bewussten Spools oder Blob-Dateien. Für nicht sicher streambar redigierbare JSON-Payloads wird die Capture-Seite vor Persistenz verworfen; der Proxy läuft weiter.
2. Komprimierte Payloads werden **nur** mit begrenzter Dekompression und vollständig validierter Encoding-Kette analysiert; sonst `capture_status=failed`, `replay_equivalence=none`. Keine ungeprüften gzip-Rohbytes persistieren, wenn Body-Redaction konfiguriert ist.
3. **Bekannte Typen:** `application/json` (auch `+json`, mit OData-Parametern) wird nach `LH-FA-14.a` redigiert; `application/xml` ist nur für die Antwort auf `/$metadata` bekannt (EDMX/CSDL enthält keine Geheimnisse, `redaction=none`, kein JSON-Pfad-Rewrite). Jeder andere Typ, auch `multipart/mixed` (`$batch`) und `text/*`, ist unbekannt. Unbekannte MIME-Typen, fragmentierte JSON-Tokens, unvollständige JSON-Dokumente und Bodies über dem konfigurierten Grenzwert sind **fail-closed für Capture**. Header-Redaction erfolgt vor jedem dauerhaften Metadaten-Write.
4. Die Streaming-Redaction muss Token über Chunk-Grenzen korrekt erkennen. Bei Absturz dürfen nur bereits redigierte Blobs existieren. Fehler werden mit `ODR-E2401` protokolliert, ohne Payload im Log.
5. Ohne sichere Speicher-/Spool-Garantie ist für große oder unbekannte Bodies die **Body-Aufzeichnung deaktiviert**, nicht bloß auf ein unsicheres Tempfile ausgelagert.

### SPEC-052 — Replay-HTTP-Identität und Lebenszyklus

Pflichtheader bei sequenziellem Replay: `X-Recorder-Workspace` (bestehende Workspace-ID), `X-Recorder-Recording-Session` (aufgezeichnete Session-ID), `X-Recorder-Replay-Session` (vom Test-Harness vergebene, innerhalb des Workspace stabile Client-Session-ID). Kein Header ist ein Upstream-Header; sie werden vor Weiterleitung entfernt. Der Client generiert eine kollisionsresistente Replay-Session-ID (UUID empfohlen) und behält sie über die Request-Folge. Zwei Clients dürfen dieselbe Recording-Session mit unterschiedlichen Replay-Session-IDs unabhängig abspielen. Fehlender, leerer oder mehrdeutiger Header: `ODR-E5003`, HTTP 400, ohne Cursoränderung. Unbekannte Workspace-ID: `ODR-E7001`, unbekannte Recording-Session: `ODR-E7003` (je HTTP 404). Servicefremde IDs: `ODR-E5004`, HTTP 409, ohne Cursoränderung. Eine neue Replay-Session startet an Position 0; ein Reset des Workspace setzt alle Cursor atomar auf Position 0 zurück; ein Snapshot-Restore lässt die Cursor unverändert (`LH-FA-10.a`). Keine implizite Ableitung aus Cookies oder Client-IP.

### SPEC-053 — Capture-Bedrohungsmodell und Testvertrag

Geltungsbereich: vom Recorder verwaltete RAM-Puffer, Capture-Queues, Blob-Verzeichnis, temporäre Dateien und Logs. Unredigierte Daten dürfen **nicht durch Recorder-Code** persistent geschrieben werden. Eine Aussage über OS-Swap, Core-Dumps, Container-Runtime oder externe HTTP-Libraries gilt nur unter expliziten Betriebsbedingungen: Core-Dumps deaktiviert, Swap deaktiviert oder verschlüsselt und vom Betreiber kontrolliert, keine unverschlüsselten Diagnose-Dumps, kein Body-Logging in Middleware. Das Betriebsprofil muss diese Voraussetzungen vor Aktivierung von Body-Capture prüfen und dokumentieren.

Bounded memory: maximale Capture-Bodygröße `MAX_CAPTURE_BODY` (`SPEC-013`), maximale entpackte Größe (gleich `MAX_CAPTURE_BODY`), maximale Queue-Größe (16 MiB), Dekompressions-Verhältnis höchstens 20:1 und Deadline 30 s. Überlauf, unbekanntes Encoding, fehlerhaftes JSON, Abbruch und Timeout setzen `capture_status=failed`, `replay_equivalence=none`; unredigierte Capture-Buffer werden verworfen, während Proxy-Streaming soweit möglich fortgesetzt wird. Nur vollständig redigierte und validierte Bodies dürfen Blob-Dateien erreichen. Backpressure darf nicht zu unbeschränkter Pufferung führen. Sicherheits-Fixtures: JSON-Tokens über Chunk-Grenzen, gzip-Bombe, Content-Encoding-Kette, MIME-Mismatch, Body-Abbruch, langsamer Consumer, Queue-Überlauf, Crash zwischen Blob-Rename und DB-Commit, Redaction-Fehler und Middleware-Logging.

### SPEC-047 — Import-API-Vertrag

`POST /__admin/imports` benötigt Admin-Authentisierung (`LH-FA-13.a`) und JSON:

```json
{"import_id":"stable-client-id","service_id":"catalog","source_session_ids":["session-01"],"metadata_hash":"sha256:...","conflict_policy":"reject","target":"new_baseline"}
```

`import_id` ist idempotent: gleiche ID + identischer kanonischer Request-Hash liefert denselben Report; gleiche ID + anderer Inhalt liefert HTTP 409. Nur `target=new_baseline` und `conflict_policy=reject` sind im MVP zulässig. Erfolgreich: HTTP 201 mit `import_id`, `baseline_snapshot_id`, `coverage[]`, `entity_count`, `report`; Wiederholung: HTTP 200. Statuscodes: fehlerhaftes JSON oder fehlende Pflichtfelder `400`; unbekannte Session `404` (`ODR-E7003`); ungültige Metadaten `422` (`ODR-E6004`); ungültiger oder widersprüchlicher Dateninhalt, auch ein Key-Konflikt zwischen Payloads, `422` (`ODR-E6101`); gleiche `import_id` mit anderem Inhalt `409` (`ODR-E6102`); fehlende oder ungültige Authentisierung `401` (`ODR-E2101`); Storagefehler `500` mit Rollback. Idempotent sind nur erfolgreiche Imports (nur sie hinterlassen eine Zeile in `imports`); nach einem fehlgeschlagenen Import darf dieselbe `import_id` erneut verwendet werden. Starten zwei Requests gleichzeitig mit derselben `import_id`, gewinnt einer über den Primärschlüssel; der andere erhält bei gleichem kanonischem Request-Hash denselben Report, sonst `ODR-E6102`. Import und Baseline-Sichtbarkeit werden atomar committed.

### SPEC-048 — Crash-Recovery und Fault Injection

Fault-Injection-Punkte: vor redigiertem Blob-Write, nach fsync, nach Rename/vor SQLite-Commit, während SQLite-Commit, nach Commit/vor HTTP-Antwort. Erwartung: kein unredigierter persistenter Body; keine referenzierte fehlende Datei nach erfolgreichem Commit; ein verwaister **redigierter** Blob nach Rename ist zulässig und per GC entfernbar. Startup prüft DB-Integrität, FK-Check und referenzierte Blob-Hashes. GC löscht nur nach erneuter Referenzprüfung; Snapshot- und Exchange-Blobs werden berücksichtigt.

### SPEC-049 — MVP-OData-V4-Abnahmematrix

| Bereich | Muss (Golden-Fixture) | Soll | Nicht in MVP |
|---|---|---|---|
| GET | Collection/Key/leer (`ODV4-GET-01..03`) | ETag | Navigation-Expand |
| POST/PATCH/PUT/DELETE | primitive Properties, Key, `null`, Missing (`ODV4-CRUD-01..06`) | Concurrency via If-Match | Deep Insert |
| Query | `eq/ne/gt/ge/lt/le`, `and/or/not`, `$select`, `$orderby`, `$top`, `$skip`, `$count` (`ODV4-Q-01..12`) | zusätzliche Funktionen | `$search`, `$apply` |
| EDM | `Edm.String`, `Boolean`, `Int32`, `Int64`, `Decimal`, `Guid`, `Date`, `DateTimeOffset` (`ODV4-T-01..08`) | weitere Primitive | komplexe Navigationstypen |
| Header | `Accept`, `OData-Version`, `Content-Type` (`ODV4-H-01..03`) | `Prefer: return=…`, `If-Match` | herstellerspezifische Header |
| Fehler | 400, 404, 501, Coverage (`ODV4-E-01..04`) | 412 | proprietäre Fehlercodes |

Die Fixture-IDs sind **geplante Testadressen**, keine bereits ausgeführten Tests. HTTP-/OData-Verhalten ist an Referenz-Fixtures und dem Lastenheft zu messen; die Matrix erweitert den Muss-Umfang nicht.

### SPEC-005 — Admin-API (vorläufiger Contract)

| Methode | Pfad | Bedeutung |
|---|---|---|
| GET | `/healthz` | Prozessbereitschaft (ohne Authentisierung) |
| GET | `/__admin/sessions` | Aufzeichnungen mit Kennung und Status auflisten (leere Liste, wenn keine existieren) |
| GET | `/__admin/sessions/{id}/exchanges` | Exchanges lesen |
| DELETE | `/__admin/sessions/{id}` | Session löschen (Retention, `LH-RB-04.a`) |
| GET | `/__admin/coverage` | Coverage lesen |
| POST | `/__admin/imports` | Recording-Session explizit in neue Baseline importieren |
| GET | `/__admin/imports/{id}` | Import-Report und Provenienz lesen; unbekannt: `ODR-E7005` |
| GET | `/__admin/workspaces` | Workspaces mit Kennung, Service, Revision, Baseline und Erstellzeit auflisten |
| POST | `/__admin/workspaces` | Workspace aus einer Baseline erzeugen (`{"baseline_snapshot_id": "..."}`; `id` optional, sonst serverseitig vergeben) |
| GET | `/__admin/workspaces/{id}` | Workspace lesen; unbekannt: `ODR-E7001` |
| DELETE | `/__admin/workspaces/{id}` | Workspace samt Cursor entfernen |
| POST | `/__admin/workspaces/{id}/reset` | Baseline und alle Cursor zurücksetzen; unbekannte ID: `ODR-E7001` |
| GET | `/__admin/workspaces/{id}/snapshots` | Snapshots auflisten |
| POST | `/__admin/workspaces/{id}/snapshots` | Benannten Snapshot erzeugen (`{"name": "..."}`) |
| POST | `/__admin/workspaces/{id}/restore` | Snapshot wiederherstellen (`{"snapshot_id": "..."}`); Cursor unverändert |
| GET | `/__admin/diagnostics` | Konfigurationszusammenfassung ohne Geheimnisse, Modus, Zähler je Fehlercode |

Admin-API wird über separaten Listener ausgeliefert und verlangt die Authentisierung aus `LH-FA-13.a`; konkrete OpenAPI-Datei ist vor Implementierung zu erstellen.

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

**Importvertrag:** siehe `SPEC-047`; alle Admin-Endpunkte außer `GET /healthz` erzwingen die Admin-Authentisierung aus `LH-FA-13.a`, alle Schreibendpunkte zusätzlich Audit-Logging.

## 3. Defaults und Konstanten

| ID | Name | Vorschlagswert | Begründung |
|---|---|---|---|
| `SPEC-007` | `LISTEN` | `127.0.0.1:8080` | Sichere lokale Voreinstellung |
| `SPEC-008` | `ADMIN_LISTEN` | `127.0.0.1:8081` | Getrennte Steuerung |
| `SPEC-009` | `REPLAY_STRATEGY` | `sequential` | `sequential`, `strict`, `first` oder `last`; wiederholte Requests deterministisch (`LH-FA-04.a`) |
| `SPEC-010` | `HYBRID_FALLBACK` | `disabled` | `disabled` oder `enabled`; Hauptschalter des Upstream-Fallbacks im Modus `hybrid`. Erst `enabled` macht `HYBRID_REPLAY_FALLBACK` (Kategorien), `FALLBACK_ALLOW_HOSTS` (Ziel) und `FALLBACK_WRITE_ALLOW` (mutierend) wirksam; keine unerwarteten Backend-Aufrufe |
| `SPEC-011` | `STRICT_COVERAGE` | `true` | Keine irreführenden Collection-Antworten |
| `SPEC-012` | `UPSTREAM_TIMEOUT` | `30s` | Begrenzte Wartezeit |
| `SPEC-013` | `MAX_CAPTURE_BODY` | `8 MiB` | Schutz vor unbeschränkter Ablage; einzige Quelle der Grenze (`SPEC-053`, `LH-FA-02.b`) |
| `SPEC-014` | `SQLITE_BUSY_TIMEOUT` | `5000ms` | Geordnete Konkurrenz |
| `SPEC-015` | `RECORDING_FORMAT_VERSION` | `1` | Versionierbare Persistenz |
| `SPEC-054` | `SERVICE_ROOT_PATH` | `/` bei genau einem Service, sonst Pflichtangabe je Service | Adressierung (`LH-FA-04.c`) |
| `SPEC-055` | `PUBLIC_BASE_URL` | abgeleitet aus `LISTEN`; bei nicht-loopback `LISTEN` Pflichtangabe | Basis für `Location`, `@odata.id`, `@odata.nextLink` |
| `SPEC-056` | `ADMIN_TOKEN_FILE` | `<DATA_DIR>/admin.token` | Admin-Authentisierung (`LH-FA-13.a`) |
| `SPEC-057` | `UPSTREAM_TLS_VERIFY` | `true` | Zertifikatsprüfung; abgeschaltet nur durch `UPSTREAM_TLS_INSECURE_SKIP_VERIFY` (`LH-RB-02.a`) |
| `SPEC-058` | `RETENTION_MAX_AGE` | `0` (aus) | Löschfrist für Sessions als Dauer mit Einheit, z. B. `30d` oder `720h`; `0` schaltet aus (`LH-RB-04.a`) |
| `SPEC-059` | `SHUTDOWN_TIMEOUT` | `10s` | Frist für laufende Exchanges beim Beenden (`LH-FA-02.c`) |
| `SPEC-060` | `REDACT_HEADERS` | leer (ergänzt die Default-Liste) | `LH-FA-14.a` |
| `SPEC-061` | `MATCH_VARY_HEADERS` | `Accept, OData-Version, Content-Type, If-Match, Prefer` | `LH-FA-05.a` |
| `SPEC-062` | `FALLBACK_ALLOW_HOSTS` | leer | Upstream-Fallback nur zum Host der konfigurierten `UPSTREAM_URL`, und nur wenn dieser Host hier aufgeführt ist; gilt für lesenden wie mutierenden Fallback (`LH-FA-12.a`) |
| `SPEC-063` | `VOLATILE_HEADERS` | `Date, Server, Via, Age, X-Request-Id, X-Correlation-Id, Traceparent` | `LH-QA-02.a` |
| `SPEC-075` | `MODE` | `mock` | Betriebsart `record`, `replay`, `mock`, `hybrid`; Startkonfiguration (`LH-FA-12.a`) |
| `SPEC-076` | `UPSTREAM_URL` | keiner (Pflicht in `record` und bei Upstream-Fallback) | Ziel je Service; ohne Wert kein Upstream-Zugriff |
| `SPEC-077` | `DATA_DIR` | `data` (relativ zum Arbeitsverzeichnis) | Verzeichnis für SQLite, Blobs und `ADMIN_TOKEN_FILE` |
| `SPEC-078` | `REPLAY_ASSIGNMENT` | `explicit` | `ordered` ist Opt-in für seriell ausgeführte Tests (`LH-FA-04.b`) |
| `SPEC-079` | `HYBRID_REPLAY_FALLBACK` | leer | Kategorien (`unsupported`, `coverage_insufficient`), für die der Replay-Fallback gilt |
| `SPEC-080` | `FALLBACK_WRITE_ALLOW` | leer | Routenfreigaben (Methode und Pfadmuster) für mutierenden Upstream-Fallback |
| `SPEC-081` | `MATCH_INVALID_ENCODING` | `literal` | `literal` oder `reject` (`LH-FA-05.a`) |
| `SPEC-082` | `MATCH_BODY_MODE` | `bytes` | `bytes` oder `json-canonical` |
| `SPEC-083` | `REPLAY_REDACTED_BODIES` | `false` | Replay von Exchanges mit Response-Body-Redaction (`SPEC-045`) |
| `SPEC-084` | `REDACT_QUERY_PARAMS` | `access_token, token, api_key, apikey, key, sig, signature, code, client_secret` | `LH-FA-14.a` |
| `SPEC-085` | `REDACT_JSON_PATHS` | `$..password, $..passwd, $..secret, $..client_secret, $..access_token, $..refresh_token, $..api_key` | `LH-FA-14.a` |
| `SPEC-086` | `UPSTREAM_CA_FILE` | keine | zusätzliche Vertrauensanker (`LH-RB-02.a`) |
| `SPEC-087` | `UPSTREAM_TLS_INSECURE_SKIP_VERIFY` | `false` | schaltet die Zertifikatsprüfung ab, nur ausdrücklich (`LH-RB-02.a`) |
| `SPEC-088` | `LISTEN_TLS_CERT`, `LISTEN_TLS_KEY` | keine | TLS des OData-Listeners |
| `SPEC-089` | `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` | keine | TLS des Admin-Listeners; bei nicht-loopback `ADMIN_LISTEN` Pflicht (`LH-FA-13.a`) |
| `SPEC-090` | `ADMIN_ALLOW_PLAINTEXT` | `false` | erlaubt unverschlüsselten Admin-Zugriff bei nicht-loopback `ADMIN_LISTEN` |
| `SPEC-091` | `SERVICES_FILE` | keine | Konfigurationsdatei für mehrere Services: eine Liste mit `id`, `service_root` und `upstream_url` je Service; ohne Datei gilt genau ein Service aus `SERVICE_ROOT_PATH` und `UPSTREAM_URL` |
| `SPEC-092` | `LINK_REWRITE` | `off` | Umschreiben absoluter Upstream-Links und Weiterleitungen; nur `off` ist in der ersten Version zulässig, jeder andere Wert verhindert den Start (`ODR-E2001`) |

---

## 4. Fehler-Codes und Logging-Felder

| ID | Code | Bedingung | Aktion |
|---|---|---|---|
| `SPEC-016` | `ODR-E2001` | Ungültige Konfiguration | Start verhindern |
| `SPEC-064` | `ODR-E2101` | Admin-Authentisierung fehlt oder ungültig | HTTP 401, kein Zustandswechsel |
| `SPEC-017` | `ODR-E2401` | Capture/Redaction fehlgeschlagen | Body nicht persistieren, Diagnose |
| `SPEC-018` | `ODR-E3001` | Recording/Blob beschädigt | Laden verweigern |
| `SPEC-019` | `ODR-E4001` | Upstream nicht erreichbar | HTTP 502/504 je Ursache |
| `SPEC-020` | `ODR-E5001` | Replay-Mismatch (kein Exchange mit diesem Match-Key) | HTTP 404 |
| `SPEC-021` | `ODR-E5002` | Replay erschöpft | HTTP 410 |
| `SPEC-074` | `ODR-E5007` | Strict-Replay: nächste erwartete Exchange hat anderen Match-Key | HTTP 409, kein Cursor-Fortschritt |
| `SPEC-022` | `ODR-E6001` | Nicht unterstützte OData-Query oder lesender Zugriff auf nicht unterstützte Property/Typ | HTTP 501 |
| `SPEC-023` | `ODR-E6002` | Nicht unterstützte Mutation oder schreibender Zugriff auf nicht unterstützte Property/Typ | HTTP 501 |
| `SPEC-024` | `ODR-E6003` | Unvollständige Coverage (Collection, partielle Entity, unbeobachtete Property) | HTTP 422 |
| `SPEC-050` | `ODR-E5003` | Workspace-/Replay-Zuordnung fehlt, leer oder mehrdeutig | HTTP 400, kein Cursor-Fortschritt |
| `SPEC-065` | `ODR-E5004` | Zuordnung gehört zu einem anderen Service | HTTP 409, kein Cursor-Fortschritt |
| `SPEC-066` | `ODR-E5005` | Kein Service für den Pfad konfiguriert | HTTP 404 |
| `SPEC-067` | `ODR-E5006` | Ungültige Kodierung bei `MATCH_INVALID_ENCODING=reject` | HTTP 400 |
| `SPEC-068` | `ODR-E6004` | Metadaten ungültig oder nicht unterstützt | Start verhindern bzw. Import abweisen (422) |
| `SPEC-051` | `ODR-E6101` | Baseline-Import ungültig/inkonsistent | Import atomar abbrechen, HTTP 422 |
| `SPEC-069` | `ODR-E6102` | `import_id` mit abweichendem Inhalt wiederverwendet | HTTP 409 |
| `SPEC-070` | `ODR-E7001` | Workspace unbekannt | HTTP 404 (Admin und Mock/Replay) |
| `SPEC-025` | `ODR-E7002` | Snapshot unbekannt | HTTP 404 (Admin) |
| `SPEC-071` | `ODR-E7003` | Recording-Session unbekannt | HTTP 404 |
| `SPEC-072` | `ODR-E7004` | Session von einer Baseline referenziert, Löschen abgewiesen | HTTP 409 |
| `SPEC-073` | `ODR-E7005` | Import unbekannt | HTTP 404 (Admin) |
| `SPEC-093` | `ODR-E5008` | Treffer nicht replayfähig (`SPEC-045`) | HTTP 422, kein Cursor-Fortschritt |
| `SPEC-094` | `ODR-E7006` | Snapshot-Name im Workspace bereits vergeben | HTTP 409 (Admin) |
| `SPEC-095` | `ODR-E7007` | Workspace-ID bereits vergeben | HTTP 409 (Admin) |
| `SPEC-096` | `ODR-E7008` | Snapshot-Art passt nicht zur Operation (Workspace-Snapshot als Baseline, Restore einer Baseline) | HTTP 422 (Admin), nichts wird geändert |
| `SPEC-097` | `ODR-E7009` | Session ist `open`: Replay, Import und Löschen abgewiesen | HTTP 409 |

Fallback-Entscheidungen (`LH-FA-12.b`) hängen am `ODR-`-Code, nie am HTTP-Status. Logs enthalten `timestamp`, `level`, `mode`, `service_id`, `session_id`, `exchange_id`, `workspace_id`, `correlation_id`, `error_code`, `duration_ms`; keine Credentials oder unredigierten Bodies.

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

Die Targets sind Festlegungen der Spezifikation. Vorhanden ist heute nur `make docs-check` (`SPEC-040`); die übrigen entstehen mit dem Slice, der sie liefert, und gehören dann in `make help` und `harness/README.md` §Sensors.

| ID | Werkzeug | Festlegung |
|---|---|---|
| `SPEC-036` | `make test` | Unit-Tests für Domain, Matching, Query-Interpreter, CRUD |
| `SPEC-037` | `make test-integration` | Record → Upstream stoppen → Replay; Mock CRUD → Reset |
| `SPEC-038` | `make test-determinism` | 100 Wiederholungen identischer Inputs; identische normalisierte Outputs und Cursor-Stände |
| `SPEC-039` | `make arch-check` | Hexagonale Importregeln; keine Adapter-Imports im Core |
| `SPEC-040` | `make docs-check` | Kennungen, Links und Dokumentkonventionen prüfen |
| `SPEC-041` | `make test-security` | Redaction, Admin-Bindings, Fallback-Verbote, Credential-Leak-Checks |
| `SPEC-042` | `make test-contract` | OData-V4-JSON, Metadata, Fehlerfälle und HTTP-Semantik |

Abnahmeparameter für Performance und Concurrency stehen in `LH-QA-01.a` und `LH-QA-03.a`; der Messaufbau muss reproduzierbar im Test-Harness hinterlegt sein.

---

## 8. Historie

| Datum | Änderung |
|---|---|
| 2026-10-18 | Review-Findings eingearbeitet: Admin-Authentisierung, Snapshot-Restore ohne Cursor-Änderung, Weg zu `complete`, Header-Redaction ohne Replay-Einschränkung, Adressierung, Fehlercodes, Konkretisierung von `LH-QA-*`/`LH-RB-*` |
| 2026-10-17 | d-migrate 1.7.1, Integritätsconstraints, Replay-/Capture-Verträge, SPEC-ID-Korrektur |
| 2026-10-17 | Initiale technische Spezifikation auf Basis des Lastenhefts und der Kursvorlage |

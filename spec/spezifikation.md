# Spezifikation — OData Recorder

**Bezug zum Lastenheft:** `spec/lastenheft.md` (LH-FA-01 bis LH-FA-18, LH-QA-01 bis LH-QA-07, LH-RB-01 bis LH-RB-04).
Bei Konflikten gilt das Lastenheft.
Diese technische Spezifikation konkretisiert, erweitert aber nicht den vereinbarten Umfang.

**Rolle:** Technisch verbindliches, fortschreibbares Spec-Stratum nach `ai-harness-course`.
Priorität **Muss** ist implementierungsbindend; **Soll/Kann** sind technisch vorbereitet, aber keine Abnahmevoraussetzung der ersten Version.
Technische Defaults und Schwellenwerte sind bis zur Freigabe Vorschläge.

---

## 1. Algorithmen und Datenflüsse

### LH-FA-01.a — Transparenter HTTP-Proxy

**Eingabe:** HTTP-Request an konfigurierten Service-Root; **Ausgabe:** HTTP-Response des Upstreams.

1. Der Use Case des Betriebsmodus (`ARC-002`) ermittelt den Service aus dem Service-Root-Pfad (`SPEC-054`) und validiert die Zielkonfiguration; die Betriebsart steht in der Startkonfiguration, die Korrelations-ID vergibt der HTTP-Adapter.
   Ein Pfad ohne konfigurierten Service-Root ergibt `ODR-E5005` (HTTP 404).
   Im Modus `record` wählt der Use Case außerdem die offene Recording-Session des Service; gibt es keine, ergibt das `ODR-E5009` (HTTP 409), und der Request wird nicht weitergeleitet.
2. Proxy leitet Methode, Pfad, Query, Body und End-to-End-Header weiter; Hop-by-Hop-Header werden nach HTTP-Regeln behandelt.
3. Upstream-Response wird an den Client gestreamt.
   Weiterleitungen (`Location`) und absolute OData-Links im Body werden in der ersten Version nicht umgeschrieben (`LINK_REWRITE=off`, `SPEC-092`): Record liefert und speichert sie unverändert, Replay liefert sie byteidentisch aus.
   Ein Client, der einem aufgezeichneten absoluten Link des Upstreams folgt, erreicht den Upstream am Recorder vorbei;
   Tests, die das ausschließen müssen, verwenden Mock-Links (`PUBLIC_BASE_URL`, `LH-FA-04.c`).
   Cookies (`Set-Cookie`) werden als Header redigiert (`LH-FA-14.a`).
4. HTTP-Transport-Encoding und komprimierte Bodies werden im transparenten Pfad nicht unnötig transformiert; semantische Analyse erfolgt auf einer getrennten, begrenzten Capture-Darstellung.

**Komplexität:** O(n) über transportierte Bytes; Streaming-Speicher O(buffer).
**Fehlermodi:** Timeout, TLS-Fehler, Upstream nicht erreichbar, Client-Abbruch → definierter Gateway-Fehler, soweit noch keine Response gesendet wurde.

### LH-FA-02.a — Capture und atomarer Abschluss

1. Session mit monotoner `sequence_no` erzeugen; Beginn und Ende jedes Austauschs erfassen.
2. Request- und Response-Metadaten sammeln; Body-Streams mit begrenzten Puffern in Blob-Storage schreiben, während die Weiterleitung fortgesetzt wird.
3. Sensitive Header und konfigurierte JSON-Pfade **vor dauerhafter Ablage** redigieren; Redaction-Status dokumentieren.
4. Nach vollständig abgeschlossener Antwort Metadaten und Blob-Referenzen atomar committen; unvollständige Exchanges gesondert markieren.
5. Speicherfehler dürfen nicht als erfolgreiches Recording ausgegeben werden; eine bereits laufende Proxy-Response kann nicht rückwirkend verändert werden.

**Fehlermodi:** Speicher voll, Abbruch, nicht redigierbarer Body, ungültige Content-Encoding.
Default: sicherheitsrelevanter Capture-Fehler verwirft den sensiblen Body und meldet `ODR-E2401`.

### LH-FA-03.a — Recording laden und prüfen

- Versionskennung, Session- und Exchange-IDs, referenzierte Blobs, Checksummen und Reihenfolge prüfen.
- Fehlende oder korrupte Blob-Datei ergibt `ODR-E3001`; kein stiller Ersatz durch leeren Body.
- Schreibvorgänge: Blob temporär schreiben → fsync/rename → SQLite-Transaktion referenzieren.
  Nicht referenzierte Blobs dürfen später bereinigt werden.
- Recordings sind nach Abschluss logisch unveränderlich; neue Sessions werden append-only ergänzt.
  Eine ganze Session kann nur über die Retention-Regeln aus `LH-RB-04.a` entfernt werden.

### LH-FA-04.a — Deterministischer Replay-Cursor

**Eingabe:** Workspace, ausgewählte Session, Request, Replay-Policy.
**Ausgabe:** gespeicherte Response oder Mismatch.

1. Match-Key nach LH-FA-05.a erzeugen.
2. Im Modus `sequential` unter den Exchanges der Session mit diesem Match-Key die mit der kleinsten `sequence_no` wählen, die nicht in `consumed` des dem Testlauf zugeordneten Cursors steht.
3. Response inklusive Status und ausgewählten End-to-End-Headern ausliefern; zeit- oder hostabhängige Header nur nach expliziter Policy transformieren.
4. Cursor erst nach erfolgreicher Auswahl atomar fortschreiben (`sequence_no` zu `consumed` hinzufügen, `position` = Anzahl der Elemente); bei gleichzeitigem Zugriff pro Cursor serialisieren.
   Beispiel: aufgezeichnet A(1), B(2), A(3); fragt der Client B, A, A, erhält er A(1), dann A(3).
5. `strict` vergleicht ausschließlich die Exchange mit `sequence_no = position + 1`; passt deren Match-Key nicht, ist das ein Fehler (`ODR-E5007`) und der Cursor bleibt stehen.
   `first` und `last` wählen die erste bzw. letzte passende Exchange und verändern den Cursor nicht.
   Die Strategie ist am Cursor gespeichert.

**Fehlermodi:** `ODR-E5001`, wenn die Session keinen Exchange mit diesem Match-Key enthält;
`ODR-E5002`, wenn im Modus `sequential` alle replayfähigen Exchanges mit diesem Key verbraucht sind und kein unverbrauchter nicht replayfähiger übrig ist, oder im Modus `strict` die Session erschöpft ist;
`ODR-E5007`, wenn im Modus `strict` die nächste erwartete Exchange einen anderen Match-Key hat;
`ODR-E5008`, wenn zu dem Key nur nicht replayfähige Exchanges (`SPEC-045`) wählbar wären.
In den Modi `first` und `last` werden nicht replayfähige Exchanges übersprungen; bleibt keiner, gilt `ODR-E5008`.
Im Modus `strict` ergibt ein nicht replayfähiger Exchange an `position + 1` `ODR-E5008`, sofern sein Match-Key passt (ein abweichender Key hat Vorrang: `ODR-E5007`), und der Cursor bleibt stehen;
der Test ist damit nicht fortsetzbar, der Weg heraus ist ein Workspace-Reset (gewollt: `strict` bildet die aufgezeichnete Folge lückenlos ab).
Replay hat **keinen** Upstream-Zugriff.

### LH-FA-04.b — Explizite Zuordnung paralleler Replay-Clients

- Ein Replay-Request trägt einen vom Test-Harness gesetzten Workspace-Identifier und eine Recording-Session-Zuordnung.
  Die Transportabbildung erfolgt über `X-Recorder-Workspace`, `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session`; diese Steuerheader werden niemals an einen Upstream weitergereicht.
- Standard ist `assignment=explicit`.
  Fehlt eine Zuordnung oder ist sie mehrdeutig, wird `ODR-E5003` (HTTP 400) geliefert;
  eine unbekannte Workspace-ID ergibt `ODR-E7001`, eine unbekannte Recording-Session `ODR-E7003` (je HTTP 404), eine Zuordnung zu einem anderen Service `ODR-E5004` (HTTP 409).
  Ein implizites `ordered`-Assignment ist ausschließlich als opt-in für seriell ausgeführte Tests zulässig.
- Cursor-Identität: `(workspace_id, replay_session_id, recording_session_id)`.
  Zwei parallele Clients teilen nur dann absichtlich einen Cursor, wenn sie dieselbe explizite Replay-Session-ID verwenden; Cursor-Zugriffe sind atomar.
- Replay-Session-IDs werden beim Start eines Testlaufs vergeben.
  Das Löschen eines Workspaces entfernt dessen Cursor.
  Ein Workspace-Reset setzt sämtliche Cursor auf Position 0 und leert `consumed`.

### LH-FA-04.c — Adressierung von Service und Workspace

- **Service:** Jeder Service hat einen Service-Root-Pfad (`SERVICE_ROOT_PATH`, `SPEC-054`), der in `services.service_root` festgehalten wird.
  Ist er nicht gesetzt und genau ein Service konfiguriert, ist der Root `/`.
  Mehrere Services werden über `SERVICES_FILE` (`SPEC-091`) konfiguriert und müssen je einen ausdrücklich gesetzten Root tragen, sonst verhindert der Start die Ausführung (`ODR-E2001`);
  dasselbe gilt für gleichzeitig gesetzte `SERVICES_FILE` und Einzelschlüssel sowie für überlappende Roots (ein Root, der Präfix eines anderen ist, auch `/` neben `/a`).
  Die Service-Kennung ist `SERVICE_ID` (`SPEC-098`) bzw. `id` je Eintrag der Datei; `services.odata_version` ist in der ersten Version immer `4.0` (V2 ist nicht Teil der Abnahme).
  Die ausdrücklichen Roots stellen sicher, dass das Hinzufügen eines Services nie still den Root eines vorhandenen ändert.
  Der Pfad nach dem Root ist der OData-Pfad.
- **Workspace (Mock, Replay):** über den Header `X-Recorder-Workspace`.
  Im Modus `mock` und `hybrid` ist er Pflicht; fehlt er, ergibt das `ODR-E5003`.
  `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` sind im Replay Pflicht und im Modus `hybrid`, sobald ein Replay-Fallback konfiguriert ist (`HYBRID_REPLAY_FALLBACK`, `SPEC-079`);
  fehlen sie dort, ergibt das `ODR-E5003`.
  Alle `X-Recorder-*`-Header werden vor jeder Weiterleitung entfernt.
- **Öffentliche Basis-URL:** `Location`, `@odata.id`, `@odata.context` und `@odata.nextLink` werden aus `PUBLIC_BASE_URL` (`SPEC-055`) plus Service-Root und typisiertem Key erzeugt, nie aus dem `Host`-Header des Clients.
  Ist `PUBLIC_BASE_URL` nicht gesetzt, leitet der Prozess sie nur dann aus `LISTEN` ab, wenn dessen Host Loopback ist (`127.0.0.1`, `::1`, `localhost`);
  in den Modi `mock` und `hybrid`, die Links erzeugen, ist bei jedem anderen Host, auch beim Wildcard `0.0.0.0` oder `::` (etwa im Container), `PUBLIC_BASE_URL` Pflicht, und ohne sie verhindert der Start die Ausführung (`ODR-E2001`);
  in `record` und `replay` ist sie nicht erforderlich.

### LH-FA-05.a — Kanonisches HTTP-Matching

Match-Key = SHA-256 über eine eindeutig kodierte Tupelfolge:
`method`, `service_id`, `normalized_path`, sortierte Multimenge dekodierter Query-Paare, ausgewählte Vary-Header, Hash des redigierten Bodys (`LH-FA-14.a`).

- Query-Paare werden stabil nach Schlüssel und Wert sortiert; wiederholte Parameter bleiben erhalten.
- Ein Query-Paar wird als (dekodierter Schlüssel, dekodierter Wert) verglichen.
  `%XX` wird zu Bytes dekodiert (Hex-Groß-/Kleinschreibung ist gleich), die Bytes müssen gültiges UTF-8 sein.
  `+` bleibt das Zeichen `+` und wird nicht als Leerzeichen gelesen; `a+b` und `a%20b` sind damit verschieden (konservativ, nie falsche Gleichheit).
- Syntaktisch ungültige Kodierung (`%` ohne zwei Hex-Ziffern, kein gültiges UTF-8) ist kein Fehler: das Paar geht unverändert als Roh-Zeichenfolge in den Key und ist nur mit einem byteidentischen Paar gleich (`MATCH_INVALID_ENCODING=literal`, Default;
  `reject` liefert HTTP 400 mit `ODR-E5006`).
- Pfadsegmentgrenzen, Groß-/Kleinschreibung und OData-Literale bleiben erhalten.
- Bei JSON-Body normalisiert `MATCH_BODY_MODE=json-canonical` (opt-in) die Schlüsselreihenfolge; Default ist `bytes` (Byte-Hash des redigierten Bodys).
- Authentifizierungstokens und volatile Header (`LH-QA-02.a`) gehen nicht in den Key ein;
  als Vary-Header gelten die Header aus `MATCH_VARY_HEADERS` (`SPEC-061`, Default `Accept`, `OData-Version`, `Content-Type`, `If-Match`, `Prefer`).
- Eine semantische Gleichsetzung von `$filter`-ASTs ist **nicht** Bestandteil des Muss-Umfangs.

Kodierung der Tupelfolge, `normalized_path`, Randformen von Query, Vary-Header und Body: `SPEC-116`.

### LH-FA-06.a — OData-V4-Metadatenmodell

1. `$metadata` im Record-Modus als HTTP-Austausch erfassen.
2. EDMX/CSDL-XML (V4) parsen und in internes Modell überführen: Schemas, EntityTypes, Property-Typen, Nullable, Keys, EntitySets, NavigationProperties.
3. Referenzen und Vererbung auflösen; nicht unterstützte Typen/Annotations explizit markieren.
   Granularität: Das Modell wird mit `ODR-E6004` (Start verhindert bzw. Import abgewiesen;
   HTTP 422 in der Admin-API) abgelehnt bei nicht parsebarem CSDL, EDMX-Version ungleich 4.0, EntityType ohne Schlüssel, nicht auflösbarer Referenz oder einem Schlüssel-Property mit nicht unterstütztem EDM-Typ.
   Eine Property mit nicht unterstütztem EDM-Typ oder eine herstellerspezifische Annotation lässt das Modell gültig;
   sie wird im Validierungsbericht ausgewiesen (`unsupported_features`), Lesezugriffe darauf ergeben `501` mit `ODR-E6001`, schreibende `501` mit `ODR-E6002`.
4. Mock liefert die gespeicherten Metadaten konsistent mit der ausgewählten Service-Version aus.
5. Ohne gültiges Modell keine zustandsbehaftete Verarbeitung dieses EntitySets; reines Replay bleibt möglich.

### LH-FA-18.a — Expliziter Import in eine Mock-Baseline

**Eingabe:** Service-Modell, ausgewählte vollständige Recordings/Exchanges, Import-Policy.
**Ausgabe:** unveränderliche Snapshot-ID, Coverage-Evidenz, Import-Report.

1. Session und Metadatenversion fixieren; nur erfolgreich abgeschlossene, redaktionsverträgliche OData-V4-JSON-Responses für einzelne Entities und Collections berücksichtigen.
2. `@odata.context`, EntitySet, Key-Typen und Property-Typen anhand des Modells validieren.
   `@odata.nextLink` markiert unvollständige Pagination, solange die komplette Kette nicht belegbar erfasst ist.
   `$top`, `$skip` und Filter ohne Vollständigkeitsbeweis erzeugen keine vollständige Coverage.
   **Vollständigkeitsbeweis (einziger Weg zu `complete`):** Die Session enthält zu einem EntitySet eine ungefilterte Collection-Abfrage.
   Ungefiltert ist ein `GET` auf die EntitySet-Collection, dessen Query leer ist oder ausschließlich aus `$format`, `$count`, `$select` und `$orderby` besteht (Positivliste);
   jede andere Option schließt die Abfrage aus, auch `$filter`, `$top`, `$skip`, `$skiptoken`, `$deltatoken`, `$search`, `$apply`, `$expand`, `$compute` und herstellerspezifische Optionen.
   Die Kette beginnt an dieser Abfrage, eine Session, die nur eine Folgeseite enthält, belegt nichts.
   Außerdem gilt: alle Antworten der `@odata.nextLink`-Kette sind erfasst (`capture_status=complete`, Status 200, dieselbe Metadatenversion), und die letzte Seite trägt keinen `@odata.nextLink`.
   **Kettenauflösung:** Der Link einer Seite wird auf Pfad und Query reduziert (ein absoluter Link muss auf den Host der `UPSTREAM_URL` des Services zeigen, sonst bricht die Kette ab);
   ein relativer Link wird gegen die Request-URL der Seite aufgelöst;
   vom Pfad wird der Service-Pfad der `UPSTREAM_URL` als Präfix entfernt, der Rest ist der OData-Pfad nach dem Service-Root wie in `exchanges.path` (beginnt der Pfad nicht mit diesem Präfix, bricht die Kette ab), daraus wird mit Methode `GET` und den Vary-Headern der Vorgängerseite der Match-Key nach `LH-FA-05.a` gebildet;
   die nächste Seite ist der Exchange derselben Session mit diesem Match-Key und der kleinsten `sequence_no` größer als die der Vorgängerseite.
   Der Match-Key umfasst die gesamte normalisierte Query, nicht nur das Token; Seiten verschiedener Abfragen lassen sich so nicht verbinden.
   Trägt eine Seite der Kette `key_redacted=1`, ist die Kette nicht belegbar, und die Coverage bleibt `partial`.
   Eine solche Kette mit leerem `value` belegt ein leeres, vollständiges EntitySet.
   Sonst lautet die Coverage `partial` (Kette begonnen oder Teilabfrage beobachtet) oder `unknown` (nichts beobachtet).
3. Jedes Entity in typisierten kanonischen Key überführen.
   Einzel-Entity- und Collection-Responses werden als beobachtete Fakten behandelt; `null`-Properties bleiben von nicht selektierten Properties unterscheidbar.
4. Pro Key deterministisch deduplizieren: identische kanonische Payloads zusammenführen; widersprüchliche vollständige Werte oder inkompatible Typen sind Konflikte und stoppen den Import.
   Teilprojektionen dürfen nur widerspruchsfrei ergänzt werden; fehlende Felder sind kein `null`.
5. Provenienz pro Entity speichern (`session_id`, `exchange_id`, EntitySet, Metadatenhash, Importzeit) und Coverage-Evidenz pro EntitySet aggregieren.
6. Vollständiges Snapshot-Manifest in temporärem Blob ablegen und nach Integritätsprüfung in einer SQLite-Transaktion als neue Baseline publizieren.
   Bei Fehler bleibt die bisherige Baseline unverändert.
   Import ist explizit, kein Record-Nebeneffekt.

**Fehlermodi:** ungültiger Key, widersprüchliche Werte, beschädigte Antwort, fehlendes oder ungültiges Metadatenmodell, redigierte notwendige Key-Felder → `ODR-E6101` bzw. `ODR-E6004` (Modell).
Redigierte Nicht-Key-Properties werden nicht übernommen und als unbekannt maskiert (`LH-FA-14.a`).
**Komplexität:** O(n) Entitäten plus Indexkosten; Import speichert Zwischenstände begrenzt bzw. in Staging-Tabellen.

### LH-FA-07.a — Stateful CRUD

- `GET` liest aus Workspace-Snapshot plus lokalen Änderungen.
- `POST` validiert EntitySet, Key und Property-Typen, erzeugt einen Datensatz und liefert die vereinbarte OData-V4-Antwort.
  Der Key muss im Request-Body stehen; der Mock vergibt keine Schlüssel und erzeugt keine Zeitstempel in Entity-Daten (Determinismus, `LH-QA-02.a`).
  Fehlt ein Schlüssel, ergibt das `400`.
- `PATCH` verändert ausschließlich übergebene Properties; `PUT` ersetzt nach der unterstützten OData-V4-Semantik.
- `DELETE` entfernt die Entity aus der Sicht des Workspaces.
- Alle Schreibvorgänge werden innerhalb einer Workspace-Transaktion serialisiert; ein fehlgeschlagener Vorgang ändert den State nicht.
- Navigation- und Referential-Integrity-Regeln werden nur umgesetzt, soweit der unterstützte Modellumfang sie eindeutig definiert.
  Sonst expliziter `ODR-E6002`.
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

Ohne `Prefer`-Header gilt (Muss): `POST` → `201` mit Entity, `PATCH`/`PUT` → `204`.
`Prefer: return=minimal` und `Prefer: return=representation` werden für mutierende Requests ausgewertet (Soll, `SPEC-049`); wird `Prefer` nicht unterstützt, gilt das Muss-Verhalten ohne Fehler.
`Location` und Entity-ID werden aus Service-Root und typisiertem Key erzeugt; keine unvalidierte Client-Host-Injektion.
`@odata.context`, `@odata.count` und `value` folgen dem ausgehandelten JSON-Metadatenlevel.
ETags bleiben Soll-Umfang.

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
- `and`, `or`, `not` werten `null` als unbekannt: `null and false` = `false`, `null or true` = `true`, `not null` = `null`; alle anderen Kombinationen mit `null` ergeben `null`.
  Ein `null`-Ergebnis lässt den Datensatz nicht durch `$filter`; nur `true` tut das.
- Eine nicht beobachtete Property (Property-Maske, `LH-FA-18.a`) ist nicht `null`; ihre Verwendung im Filter ergibt `ODR-E6003`.

Nicht unterstützte Typ-/Operator-Kombinationen liefern `501`, syntaktisch oder typologisch ungültige Ausdrücke `400`.
URL-Key-Syntax unterstützt einfache und benannte zusammengesetzte Keys; Literale werden typisiert und String-Escaping geprüft.

### LH-FA-08.a — Query-Interpreter

**Muss-Untermenge:** `$filter` mit `eq`, `ne`, `gt`, `ge`, `lt`, `le`, `and`, `or`, `not` und Klammern für primitive unterstützte EDM-Typen; `$select`, `$orderby`, `$top`, `$skip`, `$count`.

**Pipeline:** Parse → Typprüfung anhand Metadaten → Filter → stabile Sortierung (bei Gleichstand Entity-Key) → Count vor Pagination → Skip/Top → Projektion → OData-V4-JSON.

- Ungültiger Ausdruck: `400` mit OData-Fehlerobjekt.
  Dazu zählen `$top`/`$skip` mit negativem oder nicht ganzzahligem Wert; `$top=0` ist gültig und liefert eine leere `value`-Liste.
- `$count=true` liefert `@odata.count` (gefilterte Menge vor Paging); das Pfadsegment `/$count` liefert die Anzahl als `text/plain`; beide verwenden dieselbe Zählung.
- Nicht unterstützter, aber gültiger Ausdruck: expliziter `501`-Fehler `ODR-E6001`.
- `$expand` und `$batch` gehören zu LH-FA-16 (Soll); Actions und Functions gehören nicht zum ersten Abnahmeumfang (Lastenheft §5).
- Standardisierte Query-Fehler und Formatdetails sind durch OData-V4-Contract-Tests zu prüfen.

### LH-FA-09.a — Workspace-Isolation und Reset

Jeder Workspace besitzt `workspace_id`, `baseline_snapshot_id`, isolierten Entity-State und isolierte Replay-Cursor.
Kein Write darf andere Workspaces beeinflussen.
Reset ist eine atomare Wiederherstellung des Baseline-Zustands inklusive Cursor-Reset.
Während Reset werden konkurrierende Schreibvorgänge serialisiert oder kontrolliert abgewiesen.

### LH-FA-09.b — Reset-Semantik

`reset(workspace)` ersetzt in **einer** Transaktion den Entity-State durch die Baseline, setzt **alle** Replay-Cursor des Workspaces auf Position 0 (`consumed` leer) und erhöht die Workspace-Revision;
ein Abbruch lässt weder den neuen State mit altem Cursor noch das Umgekehrte zurück.
Der Vorgang ist auch ohne bisherige Mutationen definiert und idempotent bezüglich des sichtbaren State/Cursor-Inhalts.

### LH-FA-10.a — Snapshots

Snapshots sind unveränderliche, benannte, versionierte Zustandsreferenzen.
Erstellung erfasst konsistent ausschließlich den Entity-State; Replay-Cursor sind niemals Bestandteil eines Snapshots.
Restore ersetzt den Entity-State atomar, erstellt eine neue Workspace-Revision und lässt sämtliche Replay-Cursor unverändert.
Ein Restore mit unbekannter Snapshot-ID führt zu `ODR-E7002`, ohne dass ein Teilzustand übernommen wird;
ein Snapshot eines anderen Service ergibt `ODR-E5004`, ein Snapshot, der zu einem anderen Workspace gehört, `ODR-E7002`.
Ein Workspace-Snapshot gehört zu genau einem Workspace (`snapshots.workspace_id`) und trägt einen im Workspace eindeutigen Namen (Wiederholung: `ODR-E7006`);
ein Restore mit einem Snapshot der Art `baseline` ergibt `ODR-E7008`.

### LH-FA-11.a — Coverage und Teilmengen

Coverage je EntitySet: `complete`, `partial`, `unknown`.
`complete` darf nur aus explizit verifizierter Vollerfassung einschließlich aller Continuation-Seiten entstehen; ein einzelnes `$top`-Resultat ist `partial`.
Vor Query-Ausführung wird Coverage geprüft.
Ein EntitySet ohne beobachtete Entities ist nicht automatisch `complete`.
Bei `STRICT_COVERAGE=true` wird eine Collection-Abfrage über einen nicht vollständigen Bestand mit `ODR-E6003` abgewiesen.
Einzel-Entity-Lookups dürfen nur dann eine definitive `404` liefern, wenn das Nichtvorhandensein abgesichert ist; sonst Coverage-Fehler.

**Coverage-Kennzeichnung:** Jede Mock-Antwort auf eine Collection-Abfrage trägt den Header `X-Recorder-Coverage: complete|partial|unknown`.
Bei `STRICT_COVERAGE=false` wird eine Collection-Abfrage über einen nicht vollständigen Bestand aus den bekannten Entities beantwortet, und der Header nennt `partial` bzw. `unknown`;
der Antwort-Body bleibt unverändert OData-V4-konform.
Der Header wird nie entfernt (Lastenheft §1: Unvollständigkeit bleibt erkennbar).

### LH-FA-12.a — Moduswahl und Fallback

`record`, `replay`, `mock`, `hybrid`.
Der Modus ist eine Startkonfiguration, nicht per Admin-API live umschaltbar.
Hybrid-Policy ist explizit konfiguriert.
Default: `mock` → Fehler;
optional `mock` → `replay` → Fehler nur für definierte Kategorien (`unsupported`, `coverage_insufficient`), niemals für fachliche `404`, `409` oder `412`;
**kein** Upstream-Fallback.
Upstream-Fallback nur per Allowlist (`FALLBACK_ALLOW_HOSTS`, `SPEC-062`) für lesende Methoden und autorisierte Zielhosts.
Für mutierende Methoden ist Upstream-Fallback verboten, es sei denn, eine ausdrückliche Routenfreigabe (`FALLBACK_WRITE_ALLOW`, Liste aus Methode und Pfadmuster, Default leer) deckt den Request;
ohne Eintrag wird nichts an den Upstream gesendet.
Unbekannte Funktionen werden nicht stillschweigend als erfolgreich behandelt.

### LH-FA-12.b — Hybrid-Entscheidungsmatrix

| Ergebnis Mock | Default | Optionaler Replay-Fallback | Upstream-Fallback |
|---|---|---|---|
| Erfolg | Antwort | Nein | Nein |
| Fachliche `404` / `409` / `412` | Antwort/Fehler beibehalten | Nein | Nein |
| Coverage unvollständig | `ODR-E6003` | Nur explizite Policy | GET/HEAD + Allowlist; mutierend nur mit `FALLBACK_WRITE_ALLOW` |
| Nicht unterstützte Operation | `ODR-E6001` | Nur explizite Policy | GET/HEAD + Allowlist; mutierend nur mit `FALLBACK_WRITE_ALLOW` |
| Ungültige Anfrage | HTTP 400 | Nein | Nein |
| Replay-Miss nach opt-in | `ODR-E5001` | — | Nur GET/HEAD und nur für Hosts aus `FALLBACK_ALLOW_HOSTS` |

Unter den Replay-Fehlern darf nur `ODR-E5001` (kein Exchange mit diesem Match-Key) einen Fallback auslösen; die übrigen Auslöser der Matrix (Mock) bleiben davon unberührt.
Alle anderen Replay-Fehler (`ODR-E3001`, `ODR-E5002`, `ODR-E5003`, `ODR-E5004`, `ODR-E5006`, `ODR-E5007`, `ODR-E5008`, `ODR-E7001`, `ODR-E7003`, `ODR-E7009`) werden unverändert ausgeliefert;
ein erschöpfter Key fällt nie still auf den Live-Upstream durch.

### LH-FA-13.a — Management und Diagnose

Separater Admin-Listener mit Endpunkten für Sessions, Exchanges, Coverage, Workspaces, Reset, Snapshot, Import, Diagnose und Health (`SPEC-005`).
Änderungen sind auditierbar.
Jede HTTP-Response enthält strukturierte Fehlerkennung und Korrelations-ID.

**Admin-Authentisierung (unabhängig vom Binding):** Jeder Admin-Endpunkt außer `GET /healthz` verlangt `Authorization: Bearer <token>`.
Das Token stammt aus `ADMIN_TOKEN_FILE` (`SPEC-056`).
Existiert die Datei beim Start nicht, erzeugt der Prozess ein zufälliges Token (mindestens 256 Bit), legt es dort mit Rechten `0600` ab und schreibt es nie in ein Log.
Der Vergleich erfolgt zeitkonstant.
Fehlendes oder falsches Token ergibt `ODR-E2101` (HTTP 401) ohne Zustandsänderung; ein Audit-Eintrag ohne Tokenwert wird geschrieben.
Der Admin-Listener bindet standardmäßig an `127.0.0.1`; in einem Container ist `ADMIN_LISTEN=0.0.0.0:8081` mit TLS zulässig.
Eine Token-Datei, die vorhanden, aber nicht lesbar ist, nach Entfernung von Zeilenumbruch und Leerraum weniger als 32 Zeichen enthält oder auf POSIX-Systemen Gruppen- oder Weltrechte trägt, verhindert den Start (`ODR-E2001`);
ein leeres Token wird nie akzeptiert.
Fehlt die Datei und ist das Verzeichnis von `ADMIN_TOKEN_FILE` nicht beschreibbar, verhindert das den Start ebenso (`ODR-E2001`);
ebenso eine Token-Datei, die ein Verzeichnis ist, nach Auflösung eines Symlinks keine reguläre Datei ergibt oder mehr als eine Zeile enthält (nur ein abschließender Zeilenumbruch und umgebender Leerraum werden entfernt).
Im Container-Betrieb (`LH-QA-06`) stellt der Betreiber das Token als Datei mit Rechten `0400` oder `0600` bereit (etwa über das Mount-Setting des Secrets) oder lässt es vom Prozess im Datenvolumen erzeugen;
Weltleserechte sind kein Weg.
In Kubernetes ist ein gemountetes Secret ein Symlink auf eine reguläre Datei; es wird mit `defaultMode: 0400` bereitgestellt und gilt dann als gültig.
**Transportschutz:** Ist `ADMIN_LISTEN` nicht loopback, verlangt der Start `ADMIN_TLS_CERT`/`ADMIN_TLS_KEY` (`SPEC-089`) und verhindert sonst die Ausführung (`ODR-E2001`);
nur `ADMIN_ALLOW_PLAINTEXT=true` (`SPEC-090`) erlaubt unverschlüsselten Zugriff, mit einer Warnung im Log.
Die Token-Datei liegt im Datenvolumen und ist Teil von dessen Sicherungen; das Betriebsprofil nennt sie als schützenswert.
**Status:** Eine Session hat den Status `open`, `complete` oder `incomplete`.
Eine Recording-Session entsteht im Modus `record` beim Start des Prozesses für jeden konfigurierten Service (ID serverseitig vergeben) oder durch `POST /__admin/sessions` (`id` optional;
eine noch offene Session desselben Prozesses bleibt offen).
Record-Requests schreiben in die offene Session ihres Service; gibt es mehrere offene, wählt der Header `X-Recorder-Recording-Session`, fehlt er dort, ergibt das `ODR-E5003`.
Das Beenden der Aufzeichnung ist `POST /__admin/sessions/{id}/close` (unbekannte ID `ODR-E7003`; auf einer bereits beendeten Session wirkungslos) oder der geordnete Shutdown (`LH-FA-02.c`).
Läuft beim Beenden noch ein Exchange der Session, wartet `close` auf dessen Abschluss, höchstens `SHUTDOWN_TIMEOUT` (`SPEC-059`);
danach noch laufende Exchanges werden als `incomplete` geschrieben und die Session wird `incomplete`.
Mit dem Beginn von `close` nimmt die Session keine neuen Exchanges an; ein neuer Record-Request ergibt dann `ODR-E5009`.
Ein Record-Request, für den der Service keine offene Session hat, wird nicht weitergeleitet und ergibt `ODR-E5009` (HTTP 409).
Eine doppelte Session-`id` bei `POST /__admin/sessions` ergibt `ODR-E7010` (HTTP 409); der Endpunkt steht in allen Modi zur Verfügung.
Eine Session ist `open`, solange ihre Aufzeichnung läuft.
Sie wird `complete`, wenn die Aufzeichnung regulär endet (Beenden der Aufzeichnung, geordneter Shutdown nach `LH-FA-02.c`) und alle Exchanges `capture_status=complete` tragen, sonst `incomplete`.
Beim Start setzt die Wiederherstellung jede noch `open` stehende Session vor der Annahme von Requests auf `incomplete`.
Replay und Import gegen eine `open`-Session sowie das Löschen einer `open`-Session weist `ODR-E7009` (HTTP 409) ab;
eine `incomplete`-Session ist zulässig, nur ihre `complete`-Exchanges sind replayfähig (`SPEC-045`).
Zustandswechsel stehen nur in dieser Richtung (`open` → `complete`|`incomplete`); ein Workspace wird durch `revision`, `baseline_snapshot_id` und `created_at` beschrieben.

### LH-FA-02.b — Streaming- und Capture-Lifecycle

- Der HTTP-Adapter besitzt und schließt eingehende Request- und ausgehende Response-Streams; er gibt sie nur als geliehene, kontextgebundene Streams an Transport/Capture weiter.
  Der Capture-Writer hat `Write`, `Commit` und `Abort`; genau einer der beiden Abschlüsse ist zulässig.
- Transport und Capture werden mit begrenztem Puffer entkoppelt.
  Backpressure des Capture-Writers darf nicht unbeschränkt Speicher belegen; bei Größenlimit/Fehler wird Capture abgebrochen und der Proxy-Stream gemäß Transportvertrag fortgesetzt.
- Vor dem ersten gesendeten Response-Byte kann ein Transportfehler als Gateway-Status beantwortet werden.
  Nach dem ersten gesendeten Byte wird der Client-Status **nicht** rückwirkend geändert; der Capture-Status ist `incomplete`, Diagnose wird geschrieben.
- Redaction geschieht vor dauerhafter Ablage; nicht redigierbare Bodies werden verworfen.
  Ein redigierter Response-Body ist als `replay_equivalence=semantic_or_limited` gekennzeichnet und nicht als bytegenau garantiert.
- Der Capture-Limit-Test sendet einen Body oberhalb von `MAX_CAPTURE_BODY` (`SPEC-013`), verifiziert begrenzten Speicherverbrauch, korrektes Live-Streaming und einen nicht replayfähigen, nicht vollständig erfassten Capture (`capture_status=failed`, `SPEC-119` Nr. 10).

### LH-FA-02.c — Beenden des Prozesses

Bei SIGTERM/SIGINT schließt der Prozess die Listener, weist neue Requests mit HTTP 503 ab, lässt laufende Exchanges bis `SHUTDOWN_TIMEOUT` (`SPEC-059`) abschließen, finalisiert offene Capture-Vorgänge und schließt die Repositories.
Nach Ablauf der Frist werden noch offene Exchanges mit `capture_status=incomplete` abgeschlossen;
ein erzwungener Abbruch erzeugt nie einen als `complete` markierten Teil-Exchange (Crash-Fälle: `SPEC-048`).

### LH-FA-14.a — Redaction und Datensicherheit

Default-Redaction der Header: `Authorization`, `Proxy-Authorization`, `Cookie`, `Set-Cookie`, `X-API-Key`, `X-Auth-Token`, `X-CSRF-Token`;
`REDACT_HEADERS` (`SPEC-060`) ergänzt die Liste, verkürzt sie nicht.
Zusätzlich werden Query-Parameter und JSON-Felder redigiert (`REDACT_QUERY_PARAMS`, `REDACT_JSON_PATHS`, `SPEC-084`/`SPEC-085`, beide mit sicheren Defaults;
die Listen ergänzen sich, ein Eintrag lässt sich nur durch Änderung der Konfiguration entfernen).
Auch `raw_query` wird redigiert gespeichert.
Redigiert wird immer durch den festen Marker `[REDACTED]` (Header- und JSON-String-Wert); der Marker ist das einzige Platzhalterformat.

Header-Redaction beschränkt die Replayfähigkeit nicht (`SPEC-045`): redigierte Header werden im Replay nicht ausgeliefert, nie mit Marker.
Reguläre Logs enthalten weder Query-Strings noch Bodies noch Header-Werte.
**Geltungsbereich der Redaction:** Sie erfasst Header, `raw_query` und die konfigurierten JSON-Pfade des Bodys.
Query-Werte in `Location` und `Content-Location` werden wie `raw_query` behandelt (`REDACT_QUERY_PARAMS`); Link-Annotationen im Body (`@odata.nextLink`, `@odata.id`) nur über `REDACT_JSON_PATHS`.
Der Pfad (`exchanges.path`) wird nicht redigiert: Geheimnisse im Pfad sind nicht unterstützt und gehören nicht in aufgezeichnete URLs.
Randformen der Redaction von Headern, Query und `Location` sowie die Syntax der Pfadmuster: `SPEC-115`; die Replayfähigkeit als Funktion: `SPEC-117`; den Redaktor der Bodies: `SPEC-118`.

**Redaction und Match-Key:** Der Match-Key (LH-FA-05.a) wird ausschließlich über die redigierte Darstellung gebildet.
Dieselbe Redaction-Funktion läuft beim Record und beim Replay über den eingehenden Request, bevor der Key berechnet wird;
ein Hash eines unredigierten Geheimnisses wird nie persistiert, und ein Replay-Request mit demselben Geheimnis trifft denselben Key.
Grenze: Requests, die sich nur im redigierten Teil unterscheiden, haben denselben Key; das ist gewollt, weil Geheimnisse nicht unterscheidbar sein dürfen.
Solche Exchanges tragen `key_redacted=true`;
trifft ein Replay einen davon, trägt die Antwort den Header `X-Recorder-Match-Redacted: true` und das Log eine Warnung, sodass das Zusammenfallen nicht unbemerkt bleibt.

**Redaction und Import:** Ein Property-Wert mit dem Marker oder unter einem redigierten JSON-Pfad gilt im Import als nicht beobachtet (Property-Maske `unknown`, nicht `null`, nicht als Nutzdaten übernommen).
Ein redigiertes Key-Feld ergibt `ODR-E6101`.
Unbekannte sensible Payloads lassen sich nicht automatisch sicher erkennen: Capture muss deaktivierbar sein, und ein expliziter Datenklassifizierungsprozess ist notwendig.
Secrets ausschließlich aus sicherer Konfiguration, nicht in Logs oder Exporten.
Replay kann redigierte Auth-Daten nicht originalgetreu wiederherstellen.

### LH-FA-15.a — Optimistische Concurrency (Soll)

Entity-Versionen als ETag; `If-Match`/`If-None-Match` entsprechend dem unterstützten OData-Vertrag auswerten.
Konflikt ergibt `412`; Versionserhöhung erfolgt atomar.
Ohne aktivierte ETag-Unterstützung wird kein scheinbar gültiger Concurrency-Schutz behauptet.

### LH-FA-16.a — Erweiterte OData-Funktionen (Soll)

Erweiterungspunkte für `$expand` und `$batch` mit atomaren Changesets.
Actions und Functions sind nicht Teil des ersten Abnahmeumfangs (Lastenheft §5) und bleiben Erweiterungspunkt der Architektur.
Jede Funktion benötigt eigenen Parser-, Validator-, Executor- und Serializer-Contract-Test.
Unbekannte Operationen im Mock-Modus liefern `501`; Replay kann sie unabhängig vom Parser bedienen.

### LH-FA-17.a — OData V2 (Kann)

Versionierter Adapter für V2-Metadaten, Payloads und Query-Semantik; keine automatische V2/V4-Vermischung in einem Service-Modell.
Nicht Bestandteil der ersten Abnahme.

### LH-QA-01.a — Replay-Performance

Messaufbau gemäß `LH-QA-01` für Antworten allein aus lokalem Speicher (Replay und Mockserver;
Record und Pass-through sind ausgenommen): lokaler Speicher, 2 vCPU, 4 GiB RAM, 10.000 Exchanges, Bodies bis 64 KiB, 20 Requests/s, 10 Minuten nach Warm-up.
Gemessen wird die Zeit von Request-Eingang bis zum ersten Response-Byte im Prozess (Messpunkt ist der Gesamtspan `odata.gateway.request`, `SPEC-026`, vom Request-Eingang bis zum ersten Response-Byte;
`odata.replay.match`, `SPEC-028`, und `odata.mock.query`, `SPEC-029`, zerlegen ihn und weisen den Anteil der Auswahl bzw. Query-Ausführung aus).
Für den Mockserver gelten dieselben Ressourcen und Raten mit einem Workspace aus 10.000 Entities über mehrere EntitySets (Payloads bis 64 KiB) und einer Mischung aus Einzel-Lookups und Collection-Abfragen mit `$filter`/`$top`.
Der Match-Key-Index `(session_id, match_key)` (`SPEC-004`) muss ohne Tabellenscan nutzbar sein.
Der Zielwert p95 ≤ 100 ms ist vorläufig; die Bestätigung durch den fachlichen Owner steht aus (`LH-QA-01`).

### LH-QA-02.a — Determinismus-Randbedingungen

- **Volatile Header** (`VOLATILE_HEADERS`, `SPEC-063`) werden im Vergleich der normalisierten Antworten ignoriert und gehen nicht in den Match-Key ein; alle übrigen Header sind relevant.
- Der Mock vergibt keine Schlüssel und erzeugt keine Zeitstempel in Entity-Daten (`LH-FA-07.a`); ETags (Soll) leiten sich aus dem Versionszähler der Entity ab.
- Stabile Sortierung bei Gleichstand nach Entity-Key (`LH-FA-08.a`); Collections ohne `$orderby` werden nach Entity-Key geliefert.
- Zeit und serverseitig erzeugte IDs (Korrelations-ID, Session-ID, Exchange-ID, Workspace-ID, Snapshot-ID) kommen aus `Clock`/`IDGenerator` (Vertrag `SPEC-114`) und sind in Tests ersetzbar;
  Replay-Session-IDs und `import_id` vergibt der Client.

### LH-QA-03.a — Isolation und Konsistenz

Messaufbau: mindestens 20 parallele Workspaces, jeder mit eigener Mutationsfolge; kein Workspace beobachtet Zustand oder Cursor eines anderen.
Nach regulärem Neustart sind abgeschlossene Mutationen und Snapshots vorhanden; Fehler-Injektion nach `SPEC-048`.
Pro Workspace werden Schreibvorgänge serialisiert (`LH-FA-07.a`), pro Cursor Zugriffe atomar (`LH-FA-04.b`).

### LH-QA-04.a — Datensicherheit

Der Nachweis besteht aus zwei Tests (`make test-security`, `SPEC-041`): Canary-Secrets in Headern, Query-Parametern und Body-Feldern werden aufgezeichnet und anschließend in allen Speicher- und Log-Artefakten gesucht, ohne Treffer.
Ein Netzwerkzugriffstest prüft, dass die Admin-Endpunkte außer `GET /healthz` ohne Token mit `ODR-E2101` abgewiesen werden und dass der Admin-Listener nur über Loopback oder mit TLS erreichbar ist (`LH-FA-13.a`).
Die Maßnahmen stehen in `LH-FA-13.a`, `LH-FA-14.a` und `SPEC-046`/`SPEC-053`.

### LH-QA-05.a — Fehlerklassen

Unterscheidbare Klassen und ihre Codes: fehlender oder nicht ausgelieferter Replay-Treffer `ODR-E5001`/`ODR-E5002`/`ODR-E5007`/`ODR-E5008`, Session-/Workspace-Zustand `ODR-E5009`/`ODR-E7004`/`ODR-E7006`/`ODR-E7007`/`ODR-E7008`/`ODR-E7009`/`ODR-E7010`, ungültige Metadaten `ODR-E6004`, nicht unterstützte Query-Option `ODR-E6001`, unvollständige Coverage `ODR-E6003`, Zuordnungsfehler `ODR-E5003`/`ODR-E5004`/`ODR-E7001`/`ODR-E7003`.
Jede Antwort trägt Code, HTTP-Status und Korrelations-ID, aber keine Credentials oder unredigierten Bodies.

### LH-QA-06.a — Container und Neustart

Der Dienst startet ohne externe Infrastruktur aus einem Container-Image mit einem Datenvolumen (`DATA_DIR`); ein frisches, leeres Verzeichnis ist ein gültiger Startzustand.
Der Neustarttest startet den Container neu und prüft Recordings, Workspaces und Snapshots.
Gleiche Konfiguration und Eingabedaten ergeben gleiches Verhalten (`LH-QA-02.a`).

### LH-QA-07.a — Protokolltreue

Der Record-Pfad leitet Status, Body und relevante End-to-End-Header unverändert weiter (`LH-FA-01.a`).
Der Vergleichstest ruft einen Referenzdienst direkt und über den Proxy auf und vergleicht Status, Body und Header ohne `VOLATILE_HEADERS`.
Unbekannte OData-Erweiterungen verhindern die HTTP-Aufzeichnung nicht; sie werden als `unsupported_features` markiert (`LH-FA-06.a`).

### LH-RB-01.a — OData-Kompatibilität

Der unterstützte Teilumfang ist `SPEC-049` zusammen mit `LH-FA-08.a`/`LH-FA-08.b`.
Alles außerhalb liefert `501` (`ODR-E6001`/`ODR-E6002`) und erscheint nie als unterstützt.
Nachweis: Referenzanfragen und Negativfälle (`make test-contract`, `SPEC-042`).

### LH-RB-02.a — HTTP/TLS-Konfiguration

Upstream: `UPSTREAM_TLS_VERIFY=true` (`SPEC-057`) prüft Zertifikatskette und Hostnamen; `UPSTREAM_CA_FILE` ergänzt die Vertrauensanker.
Die Prüfung wird nur durch `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` (`SPEC-087`) abgeschaltet; dann schreibt der Start eine Warnung ins Log.
Listener: `LISTEN_TLS_CERT`/`LISTEN_TLS_KEY` (`SPEC-088`) und `ADMIN_TLS_CERT`/`ADMIN_TLS_KEY` (`SPEC-089`);
ohne sie sprechen die Listener HTTP, was auf Loopback der Default ist und für den Admin-Listener sonst `LH-FA-13.a` unterliegt.
Ein Zertifikatsfehler zum Upstream ergibt `ODR-E4001` (HTTP 502).
Nachweis: Integrationstests für HTTP, HTTPS und Zertifikatsfehler.

### LH-RB-03.a — Linux und CI

Build und Tests laufen in Containern (Docker-only, `AGENTS.md` §3.1); das Binary ist auch nativ unter Linux startbar.
Nachweis: reproduzierbarer Container-Build und End-to-End-Test (`SPEC-037`).

### LH-RB-04.a — Zugriff, Redaction und Aufbewahrung

Zugriffskontrolle: `LH-FA-13.a`.
Redaction: `LH-FA-14.a`.
Aufbewahrung: Sessions werden nie automatisch gelöscht, solange `RETENTION_MAX_AGE=0` (`SPEC-058`) gilt.
Ist ein Alter gesetzt, löscht der Start Sessions, die älter sind;
eine Session, die als Quelle einer Baseline geführt wird (`baseline_sources`), überspringt der Lauf mit einer Warnung im Log, ohne den Start abzubrechen.
Außerdem löscht `DELETE /__admin/sessions/{id}` eine Session samt Exchanges und der zugehörigen Replay-Cursor in einer Transaktion;
ist die Session Quelle einer Baseline, wird sie mit `ODR-E7004` (HTTP 409) abgewiesen, bevor etwas gelöscht wird.
Blobs, die danach weder von einem Exchange noch von einem Snapshot referenziert werden, entfernt die Bereinigung unmittelbar nach dem Commit (best effort) und in jedem Fall beim nächsten Start, bevor der Dienst Requests annimmt;
spätestens dann sind die Daten vom Datenträger entfernt (Löschzeitpunkt für `LH-RB-04`).
Löschungen werden im Audit-Log festgehalten.
Das Löschen ändert die logische Unveränderlichkeit abgeschlossener Sessions (`LH-FA-03.a`) nicht, weil es ganze Sessions entfernt, nie Inhalte ändert.

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

Persistierte `headers` und `raw_query` sind redigiert (`LH-FA-14.a`); rohe Credentials gehören nicht in das Schema.
Blob-Hash bezieht sich auf die **gespeicherte redigierte** Darstellung.

`body.state` gilt je Richtung und unterscheidet `absent` (der Austausch hatte keinen Body), `present` (Blob vorhanden) und `discarded` (Body bewusst nicht persistiert, `blob` ist `null`;
`capture_status` ist dann `failed` oder `incomplete`).
Die Felder entsprechen den Spalten des Schemas (`SPEC-004`): `sequence_no`, `started_at`, `finished_at`, `request_body_state`/`response_body_state`, `request_redaction_status`/`response_redaction_status` (Feld `redaction`), `key_redacted`, `capture_status`, `body_integrity`, `replay_equivalence`.
`redaction.*=unsafe` bedeutet, dass eine erforderliche Body-Redaction nicht abgeschlossen werden konnte; der Body wird verworfen.
Ist `redaction.request=unsafe`, lässt sich der Match-Key nicht bilden;
er ist dann `null`, und der Exchange ist `capture_status=failed` (ein `null`-Key ist nur bei `failed` zulässig und passt nie auf einen Request).
`key_redacted=true` hält fest, dass die Redaction Eingaben des Match-Keys verändert hat (`LH-FA-14.a`).

### SPEC-002 — OData-Modell

`ServiceModel{version, namespaces, entity_types, entity_sets, navigation_bindings}`; `EntityType{qualified_name, keys[], properties[], navigation_properties[]}`; `Property{name, edm_type, nullable}`.
Typ- und Namensvalidierung vor State-Zugriff.

### SPEC-003 — Workspace und Entity

`Workspace{id, service_id, baseline_snapshot_id, baseline_kind, revision, created_at}` (`baseline_kind` ist immer `baseline` und stützt den Fremdschlüssel auf den Snapshot);
der Status eines Workspace in der Verwaltung ist das Tripel aus `revision`, `baseline_snapshot_id` und `created_at`;
`EntityState{workspace_id, entity_set, entity_key, payload_json, known_properties, entity_completeness, entity_version, deleted}`;
`ReplayCursor{workspace_id, replay_session_id, recording_session_id, service_id, strategy, position, consumed}` (`service_id` stützt die Fremdschlüssel auf Workspace und Session);
`consumed` ist die Menge der verbrauchten `sequence_no`, `position` ihre Anzahl.
Zusammengesetzte Schlüssel werden als kanonisch typisierte Tupel serialisiert, nicht durch naive String-Konkatenation.
**Zustandsrepräsentation:** `entities` hält den vollständigen Entity-State eines Workspace, kein Delta über der Baseline.
Die Anlage eines Workspace materialisiert den Baseline-Snapshot in `entities`; `deleted=1` ist ein Tombstone, der ein definitives `404` belegt.
Ein Workspace-Snapshot (`state_blob`) enthält den vollständigen Entity-State einschließlich Tombstones; Reset ersetzt den Entity-State durch den der Baseline, Restore durch den des Snapshots.

### SPEC-004 — Normatives Schema für d-migrate 1.7.1

Das neutrale d-migrate-Schemamodell (YAML-Quelle des Repos) ist die einzige Quelle der Tabellen, Spalten, Schlüssel, Indizes, UNIQUE-, CHECK- und **zusammengesetzten FOREIGN-KEY-Constraints**. d-migrate **1.7.1** versteht `tables.<table>.constraints` mit `type: foreign_key`, `columns: [...]` und `references: {table, columns: [...], on_delete: ...}` sowie `type: unique` und `type: check` mit `expression`.
Das Schemaformat selbst bleibt `schema_format: "1.0"`;
die Anwendungsschema-Version (`version` in `tools/schema/schema.yaml`) wird unabhängig von der Dokumentversion des Lastenhefts geführt;
ihre Änderung folgt jeder Änderung des Schemas.

Normative DB-Invarianten:

1. `snapshots(id,service_id)` und `snapshots(id,service_id,snapshot_kind)` sind UNIQUE;
   `workspaces(baseline_snapshot_id,service_id,baseline_kind)` referenziert das Tripel, `baseline_kind` ist per Check immer `baseline`.
   Ein Workspace kann damit weder die Baseline eines anderen Service noch einen `workspace`-Snapshot als Baseline verwenden.
   Ein `workspace`-Snapshot trägt `workspace_id` und `name` (Check `ck_snapshot_scope`), der Name ist je Workspace eindeutig;
   `workspace_id` ist kein Fremdschlüssel (Zirkel mit `workspaces`), sondern wird transaktional geprüft, und das Löschen eines Workspace löscht seine Snapshots in derselben Transaktion.
2. `imports(baseline_snapshot_id,service_id)` referenziert dasselbe Paar;
   `baseline_snapshot_id` ist Pflicht und eindeutig: eine Zeile in `imports` entsteht in derselben Transaktion wie die Baseline (nur erfolgreiche Imports;
   ein fehlgeschlagener Import hinterlässt keine Zeile und wird im Audit-Log festgehalten, daher darf seine `import_id` erneut verwendet werden).
   `baseline_sources(baseline_snapshot_id,session_id)` hält die Quell-Sessions abfragbar fest; der Fremdschlüssel auf `sessions` ist `RESTRICT`.
   `snapshot_kind=baseline` wird zusätzlich **innerhalb derselben Transaktion** geprüft; ein gewöhnlicher `workspace`-Snapshot darf nicht stillschweigend promotet werden.
3. `sessions(id,service_id)` und `workspaces(id,service_id)` sind UNIQUE;
   `replay_cursors(recording_session_id,service_id)` und `(workspace_id,service_id)` erzwingen dieselbe Service-Identität;
   ein Cursor wird mit seiner Recording-Session gelöscht (`CASCADE`).
   Die Replay-Session ist ein clientdefinierter, workspacegebundener Identifikator, nicht die Recording-Session.
4. `exchanges(session_id,sequence_no)` ist UNIQUE; Index `(session_id,match_key)` unterstützt Matching.
   Enumerationen für Capture/Redaction/Body/Replay/Coverage und numerische Werte sind als `check`-Constraints in der YAML-Datei enthalten, ebenso die Zustandsinvarianten eines Exchange: `sequence_no >= 1`;
   ein verworfener Body setzt `capture_status` ungleich `complete` voraus;
   `redaction=unsafe` verlangt einen verworfenen Response-Body und `replay_equivalence=none`, auf der Request-Seite `capture_status=failed`.
   `baseline_sources` und `coverage` führen `baseline_kind` und referenzieren damit nur Snapshots der Art `baseline`.
   Nur im Anwendungstest (`make test-integration`) belegt werden Invarianten, die ein neutrales Check nicht ausdrückt: `position` entspricht der Anzahl in `consumed_json`, die JSON-Spalten enthalten gültiges JSON, und das Löschen eines Workspace löscht seine Snapshots in derselben Transaktion (kein Fremdschlüssel).
5. Jede SQLite-Verbindung setzt `PRAGMA foreign_keys=ON`.
   Alle mutierenden UseCases prüfen fachliche Invarianten transaktional und dürfen Datenbank-Constraint-Fehler nicht ignorieren.
   `PRAGMA foreign_key_check` und `integrity_check` sind verpflichtende Gates.

**Migrationsregel:** Das DDL wird ausschließlich mit gepinntem `d-migrate:1.7.1` generiert; es existiert keine parallele manuell gepflegte Tabellen-DDL.
CI muss `schema validate`, `schema generate --target sqlite`, DDL-Ausführung, Negativtests und die vollständige Migrationskette auf einer Kopie bestehender Datenbanken durchführen.
Als Gate verdrahtet ist `make schema-check`: `schema validate`, `schema generate --target sqlite`, DDL-Ausführung in SQLite, die Negativtests der Constraints und `foreign_key_check`/`integrity_check`;
`make schema-validate` und `make schema-generate` sind Werkzeuge ohne Gate-Eigenschaft.
Nicht belegt ist die Migrationskette auf bestehenden Datenbanken; sie entsteht mit der zweiten Schemaversion.

**Versioniertes Snapshot-Manifest:** `manifest_version=1`, `snapshot_kind`, `service_id`, `metadata_hash`, `coverage[]`, `entities[]`, `provenance[]`, `source_session_ids[]`;
pro importierter Property werden Quelle (`session_id`, `exchange_id`), Projektion und Hash festgehalten.
Canonical JSON und SHA-256 sichern die Integrität.
Unveränderliche Baselines dürfen nicht überschrieben werden.
Snapshot-Promotion ist ein eigener validierter Vorgang; Coverage wird nie implizit auf `complete` gesetzt.

### SPEC-103 — Öffnen und Migrieren der SQLite-Datei

Der Driven Adapter für die Persistenz öffnet genau eine Datei und bringt sie auf die Schemaversion des Programms.
Die Schemaversion der Datei steht in `PRAGMA user_version`; das Programm kennt heute die Version `1`, ihr DDL ist das aus `tools/schema/schema.yaml` erzeugte (`SPEC-004`).

1. Jede Verbindung der Datei setzt `foreign_keys=ON` und `busy_timeout` aus `SQLITE_BUSY_TIMEOUT` (`SPEC-014`, ohne Angabe 5000 ms).
   Jedes erfolgreiche Öffnen stellt zuletzt `journal_mode=WAL` sicher; der Modus bleibt in der Datei, jede Verbindung sieht ihn, und eine Datei im Modus `delete` wird dabei umgestellt.
   Das Umstellen wiederholt sich bei `SQLITE_BUSY` (Code 5) bis zum Ablauf des `busy_timeout`, bei jedem anderen Fehler nicht; danach scheitert Öffnen mit dem Fehler der Datenbank, die Migration bleibt bestehen, und ein späteres Öffnen stellt den Modus erneut sicher.
   Bricht der Kontext während der Wiederholung ab, endet Öffnen mit dem Fehler des Kontexts.
   Eine abgelehnte Datei bleibt byte-gleich: vor der Prüfung der Version schreibt Öffnen nichts in sie.
2. `user_version` `0` und keine Tabelle: die Datei ist leer oder neu; das DDL der Version `1` wird in einer Transaktion angewendet und `user_version` auf `1` gesetzt.
3. `user_version` `1`: nichts geschieht; ein zweites Öffnen verändert die Datei nicht.
4. `user_version` größer als die bekannte Version: Öffnen scheitert mit einem Fehler, der die gefundene und die bekannte Version nennt; die Datei bleibt unverändert.
5. `user_version` `0` mit vorhandenen Tabellen (eine fremde Datenbank): Öffnen scheitert, die Datei bleibt unverändert.
6. Eine Datei, die keine SQLite-Datenbank ist oder deren Aufbau SQLite als beschädigt meldet (Fehlercodes `SQLITE_NOTADB` und `SQLITE_CORRUPT`): Öffnen scheitert mit dem Fehler `ErrBeschaedigt`, den die meldende Schicht als `ODR-E3001` (`SPEC-018`) ausgibt; die Datei bleibt unverändert.
   Andere Fehler der Datei (etwa `SQLITE_CANTOPEN`, `SQLITE_IOERR`) gibt Öffnen unverändert weiter; eine `integrity_check`-Prüfung beim Öffnen gehört nicht dazu (`StorageMaintenance`).
7. Ein leerer Pfad, ein Pfad auf ein Verzeichnis oder ein Verzeichnis, das nicht existiert: Öffnen scheitert; der Adapter legt kein Verzeichnis an (`DATA_DIR` legt der Composition Root an, `SPEC-077`).
8. Zwei gleichzeitige Öffnungen einer neuen Datei: genau eine wendet die Migration an, beide enden mit Erfolg.
9. Ein Schließen ist idempotent: ein zweites `Close` ist kein Fehler.
10. Der Pfad `:memory:` ist nicht zugelassen; Tests nutzen eine Datei in einem temporären Verzeichnis.

### SPEC-105 — Sessions und Exchanges: Vertrag der Repositories

`SessionRepository` und `RecordingRepository` sind die Outbound Ports für Sessions und ihre Exchanges (`ARC-004`).
Kennungen vergibt der Aufrufer; die Zeitstempel sind UTC, liegen in den Jahren 1 bis 9999 und werden als RFC 3339 mit genau neun Nachkommastellen gespeichert, damit die Textordnung die Zeitordnung ist, und gelesen.
Jede Operation ist atomar: sie geschieht ganz oder gar nicht.
Die Nullzeit `0001-01-01T00:00:00Z` gilt als nicht gesetzt.
Die Prüfung der Gültigkeit eines Arguments geht der Prüfung des Kontexts voraus: ein ungültiges Argument ist `ErrUngueltig`, auch bei abgebrochenem Kontext.

**Session** (`open` → `complete` oder `incomplete`; ein Zustand außer `open` ist endgültig):
1. `Create` legt eine Session im Zustand `open` an. Ein anderer Zustand, eine leere Kennung, ein leerer `service_id` oder eine Zeit außerhalb der Jahre 1 bis 9999 ist `ErrUngueltig`; danach folgt die Prüfung der Kennung (`ErrBereitsVorhanden`) und die des Service (`ErrDienstUnbekannt`).
2. `Get` einer unbekannten Kennung ist `ErrNichtGefunden`.
3. `SetStatus` erlaubt nur `open` → `complete` und `open` → `incomplete`. Jede andere Folge (auch auf `open`, auch aus einem geschlossenen Zustand, auch derselbe Zustand noch einmal) ist `ErrUngueltigerUebergang` und ändert nichts; eine unbekannte Kennung ist `ErrNichtGefunden`.
4. `List` liefert die Sessions eines Service nach der Zeit `created_at` (nicht nach ihrem Text, Zeitzonen eingeschlossen), dann nach Kennung aufsteigend; ein Service ohne Sessions oder ein unbekannter Service liefert eine leere Liste, ein leerer `service_id` ist `ErrUngueltig`.
5. `FindOpen` liefert alle Sessions im Zustand `open` in derselben Reihenfolge, über alle Services.
6. `Delete` löscht die Session mit ihren Exchanges und Cursors in einer Transaktion. Ist die Session Quelle einer Baseline (`baseline_sources`), ist es `ErrReferenziert` und nichts wird gelöscht; eine unbekannte Kennung ist `ErrNichtGefunden`.

**Exchange** (append-only):
7. `Append` hängt einen Exchange an eine Session im Zustand `open`. Eine unbekannte Session ist `ErrNichtGefunden`, eine geschlossene `ErrSessionGeschlossen`.
8. `sequence_no` ist lückenlos: der erste Exchange einer Session trägt `1`, jeder weitere den Vorgänger plus eins; jede andere Zahl ist `ErrUngueltigeSequenz` und ändert nichts. Gleichzeitige Anhänge an dieselbe Session ordnen sich; genau einer erhält jede Nummer, der andere `ErrUngueltigeSequenz`.
   Die Prüfung von `Append` folgt dieser Reihenfolge: Gültigkeit des Exchange, Session (Nr. 7), Kennung (Nr. 9), Sequenz (Nr. 8); der erste Verstoß bestimmt den Fehler.
9. Eine bekannte Exchange-Kennung ist `ErrBereitsVorhanden`; der gespeicherte Exchange bleibt unverändert. Es gibt keine Operation, die einen Exchange ändert oder einzeln löscht.
10. Ein Exchange, der die Invarianten von `SPEC-001` und `SPEC-004` verletzt, ist `ErrUngueltig` (Prüfung `Validate` im Modell): Pflichtfelder gesetzt, Zeiten in den Jahren 1 bis 9999, `finished_at` nicht vor `started_at`, Response-Status `0` oder `100` bis `599`, Header als gültiges UTF-8, Aufzählungswerte bekannt, `body_state=present` genau mit Blob, `match_key` nur bei `capture_status=failed` leer, ein verworfener Body nie bei `capture_status=complete`, `redaction.request=unsafe` nur bei `failed` und verworfenem Request-Body, `redaction.response=unsafe` nur bei verworfenem Response-Body und `replay_equivalence=none`.
11. `Get` einer unbekannten Kennung ist `ErrNichtGefunden`; `ListBySession` liefert die Exchanges nach `sequence_no` aufsteigend, eine unbekannte Session ist `ErrNichtGefunden`, eine Session ohne Exchanges eine leere Liste.
12. Request-Header werden nie als `nil` gelesen (leer ist eine leere Abbildung); Response-Header sind `nil`, wenn keine gespeichert sind, und eine leere Abbildung, wenn eine leere gespeichert wurde. Ein Response-Status `0` heißt: keiner.
13. Ein abgebrochener Kontext beendet jede Operation beider Repositories mit dem Fehler des Kontexts, ohne eine Änderung zu hinterlassen.

Ungeprüft bleibt: Seitenweises Lesen großer Sessions (`InspectionUseCase`) und die Blob-Ablage (Blobs sind hier nur Verweise).

### SPEC-106 — Blob-Ablage: Vertrag des BlobRepository

`BlobRepository` ist der Outbound Port für Body-Bytes (`ARC-004`), der Adapter legt sie als Dateien in einem Verzeichnis ab (`ARC-010`).
Eine Blob-Referenz ist `sha256:` und 64 Kleinbuchstaben-Hexziffern des SHA-256 der Bytes; so steht sie in `body.blob` (`SPEC-001`).
Jede andere Zeichenfolge ist als Referenz `ErrUngueltig`, auch eine mit Großbuchstaben, Pfadteilen oder falscher Länge.
Das Verzeichnis legt der Aufrufer an (`DATA_DIR`, `SPEC-077`); der Adapter legt darin nur seine Unterverzeichnisse an.
Ein leerer Pfad, ein fehlender, kein Verzeichnis oder ein Verzeichnis, in dem ein Unterverzeichnis nicht anzulegen ist, lässt das Erzeugen des Adapters scheitern, ohne eine Datei anzulegen; ein davor angelegtes leeres Unterverzeichnis bleibt.
Eine Datei ist nie sichtbar, bevor ihr Inhalt vollständig ist: geschrieben wird in eine temporäre Datei im selben Dateisystem, `fsync`, dann atomares Umbenennen auf den Namen der Referenz, dann `fsync` des Verzeichnisses.
Schlägt das Sichern des Verzeichnisses nach dem Umbenennen fehl, ist `Commit` ein Fehler und der Blob kann bestehen bleiben; er ist dann nicht referenziert und die Bereinigung entfernt ihn (`SPEC-043`).
Dateien haben die Rechte `0600`, Verzeichnisse `0700`, unabhängig von der `umask` des Prozesses.
Die Prüfung der Gültigkeit eines Arguments geht der Prüfung des Kontexts voraus; ein abgebrochener Kontext beendet jede Operation mit dem Fehler des Kontexts und ohne Änderung.

1. `Begin(ctx, grenze)` liefert einen `CaptureWriter` für höchstens `grenze` Bytes (`MAX_CAPTURE_BODY`, `SPEC-013`); eine `grenze` kleiner als `1` ist `ErrUngueltig`.
   Der Writer ist an den Kontext von `Begin` gebunden: nach dessen Abbruch enden `Write` und `Commit` mit dem Fehler des Kontexts, und der Writer gilt als abgebrochen; ohne einen weiteren Aufruf bleibt seine temporäre Datei bis zu `Abort` oder bis zum nächsten Erzeugen des Adapters bestehen.
2. `Write` hängt Bytes an. Überschreitet die Summe die `grenze`, ist es `ErrZuGross`; der Writer gilt als abgebrochen und seine Datei ist gelöscht. Genau `grenze` Bytes sind zulässig.
3. `Commit` schließt den Writer und liefert die Referenz und die Länge. Ein leerer Body ist zulässig (Referenz des leeren Inhalts).
   Gleiche Bytes haben dieselbe Referenz und dieselbe Datei: ein weiterer `Commit` mit gleichen Bytes legt keine zweite Datei an, auch gleichzeitig.
   Ist die Datei zur Referenz vorhanden, aber beschädigt, ersetzt der `Commit` sie durch die geprüften Bytes; liegt dort eine Nicht-Datei (ein Verzeichnis, ein Verweis), entfernt der `Commit` sie zuvor, einen Verweis selbst und nie sein Ziel.
4. Ein Writer hat genau einen Abschluss. `Write` und `Commit` nach `Commit` oder `Abort` sind `ErrAbgeschlossen`; ein zweiter `Commit` ist `ErrAbgeschlossen` und ändert nichts.
   `Abort` löscht die temporäre Datei und hinterlässt nichts; `Abort` nach `Abort` und `Abort` nach `Commit` sind kein Fehler, das zweite lässt den abgeschlossenen Blob bestehen.
5. `Open` liest die Datei einer Referenz vollständig gegen ihren Hash, bevor es den Leser liefert; der Leser liest die geprüfte Datei. Eine fehlende Datei ist `ErrBlobFehlt`, eine Datei mit anderem Hash oder eine Nicht-Datei (Verzeichnis, Verweis) ist `ErrBlobBeschaedigt`; beide meldet die Schicht darüber als `ODR-E3001` (`SPEC-018`), ein Ersatz durch leere Bytes findet nie statt.
6. `Verify` prüft wie `Open` ohne Leser und liefert dieselben Fehler.
7. `List` liefert die Referenzen aller vorhandenen Blob-Dateien aufsteigend sortiert; temporäre Dateien und fremde Dateien im Verzeichnis zählen nicht.
8. `Remove` löscht die Datei einer Referenz; liegt dort eine Nicht-Datei, löscht es sie (einen Verweis selbst, ein Verzeichnis samt Inhalt); eine unbekannte Referenz ist kein Fehler (die Bereinigung ist wiederholbar).
9. Beim Erzeugen des Adapters werden übrig gebliebene temporäre Dateien früherer Läufe gelöscht (Absturz vor `Commit`).
10. Der Adapter prüft die Bytes nicht auf Redaction (`SPEC-053`): er speichert, was er bekommt.

Ungeprüft bleibt: dass der Leser von `Open` dieselbe geöffnete Datei trägt, die geprüft wurde (der Adapter öffnet sie nicht neu; ein Test könnte den Wechsel zwischen Prüfung und Rückgabe nicht erzwingen), dass der Adapter `fsync` aufruft und die Haltbarkeit auf dem Datenträger (das Betriebssystem sichert sie zu), und Blob-Dateien auf Dateisystemen ohne `0600` (kein Teil von Linux-Dateisystemen des Containers).
Nicht unterstützt: zwei Adapter auf einem Verzeichnis zugleich; das zweite Erzeugen löscht die temporären Dateien offener Writer des ersten, deren `Commit` dann mit einem Fehler endet.
Unentschieden bleibt ein `Remove` gegen einen `Commit` gleicher Bytes zugleich: der Blob besteht danach oder nicht; die Bereinigung prüft die Referenzen vor dem Löschen (`SPEC-043`).

### SPEC-107 — Service, Baseline und Import: Vertrag der Repositories

`MetadataRepository`, `SnapshotRepository`, `ImportRepository` und `CoverageRepository` sind die Outbound Ports für Services, Baselines (Snapshots der Art `baseline`), Imports und Coverage (`ARC-004`); die Tabellen sind die aus `SPEC-004`.
Kennungen vergibt der Aufrufer; Zeitstempel, Jahresbereich, Nullzeit, Atomarität und die Reihenfolge der Prüfungen (Argument vor Kontext) gelten wie in `SPEC-105`.
Eine Blob-Referenz ist eine nach `SPEC-106` gültige; die Repositories prüfen ihr Format, nicht das Vorhandensein der Datei.
Eine Baseline ist unveränderlich: kein Port ändert oder löscht sie, ein zweites Schreiben unter ihrer Kennung ist `ErrBereitsVorhanden` und lässt die gespeicherte unverändert (`SPEC-004`).

**Service** (`MetadataRepository`):
1. `Create` legt einen Service an. Eine leere Kennung, ein leerer `service_root` oder eine leere `odata_version` ist `ErrUngueltig`; `metadata_blob` und `metadata_hash` sind beide gesetzt oder beide leer, sonst `ErrUngueltig`; eine bekannte Kennung ist `ErrBereitsVorhanden`.
2. `Get` einer unbekannten Kennung ist `ErrNichtGefunden`; `List` liefert alle Services nach Kennung aufsteigend, ohne Services eine leere Liste.
3. `SetMetadata(id, blob, hash)` ersetzt beide Felder; beide leer löscht sie, nur eines gesetzt oder eine ungültige Referenz ist `ErrUngueltig`, eine unbekannte Kennung `ErrNichtGefunden`. Die Argumente werden vor der Kennung geprüft: ein ungültiges Paar ist auch bei unbekannter Kennung `ErrUngueltig`.

**Baseline lesen** (`SnapshotRepository`, nur Art `baseline`):
4. `Get` einer unbekannten Kennung oder der Kennung eines Snapshots anderer Art ist `ErrNichtGefunden`. `List(service_id)` liefert die Baselines eines Service nach `created_at` (als Zeit), dann Kennung; ein leerer `service_id` ist `ErrUngueltig`, ein unbekannter Service oder einer ohne Baselines liefert eine leere Liste. Snapshots anderer Art erscheinen nie.

**Import veröffentlichen** (`ImportRepository.Publish`):
5. `Publish(Publication{Import, Snapshot, SourceSessionIDs, Coverage})` schreibt den Snapshot, den Import-Eintrag, die Baseline-Quellen und die Coverage in einer Transaktion: ganz oder gar nicht; jeder Fehler, auch ein Abbruch des Kontexts mitten im Schreiben, hinterlässt keine Zeile in einer der vier Tabellen. Es liefert den Import und ob er neu ist.
6. `Publication` ist gültig, wenn: Kennungen von Import und Snapshot gesetzt; `Import.ServiceID` gleich `Snapshot.ServiceID` und `Import.BaselineSnapshotID` gleich `Snapshot.ID`; `RequestHash`, `StateHash` und `MetadataHash` nicht leer; `StateBlob` eine gültige Referenz; `ReportJSON` gültiges JSON und gültiges UTF-8; mindestens eine Quell-Session (ein leerer Import ist ungültig), keine doppelt; Coverage-Einträge mit nichtleerem `entity_set`, Status `complete`, `partial` oder `unknown`, gültigem JSON und gültigem UTF-8 als Evidenz, ohne doppelten `entity_set`; Zeiten gültig. Sonst ist es `ErrUngueltig`. Eine leere Coverage ist zulässig.
7. Gleicher `RequestHash` im selben Service ist idempotent: `Publish` liefert den gespeicherten Import mit „nicht neu“ und schreibt nichts, auch wenn Kennungen, Quellen oder Coverage der neuen Anfrage abweichen. Derselbe Hash in einem anderen Service ist eine eigene Anfrage. Zwei gleichzeitige `Publish` mit demselben Hash ordnen sich: genau einer schreibt, der andere liefert dessen Import. Die Coverage-Einträge der `Publication` tragen keine Baseline-Kennung; `Publish` schreibt sie unter der Kennung des Snapshots.
8. Die Prüfung von `Publish` folgt dieser Reihenfolge: Gültigkeit (Nr. 6), Kontext, Hash (Nr. 7), Service (`ErrDienstUnbekannt`), Kennung des Snapshots oder des Imports (`ErrBereitsVorhanden`), Quell-Sessions: eine unbekannte (auch eine leere Kennung) ist `ErrNichtGefunden`, eine eines anderen Service `ErrUngueltig`; der erste Verstoß bestimmt den Fehler.
9. Eine Session in `baseline_sources` ist nicht löschbar (`ErrReferenziert`, `SPEC-105`). `BaselinesOf(session_id)` liefert die Kennungen der Baselines, die aus ihr hervorgingen, aufsteigend; eine unbekannte Session ist `ErrNichtGefunden`, eine ohne Baselines liefert eine leere Liste. `Sources(baseline_id)` liefert ihre Quell-Sessions aufsteigend; eine unbekannte Baseline ist `ErrNichtGefunden`.
10. `Get(import_id)` liest den Import-Eintrag, unbekannt ist `ErrNichtGefunden`.

**Coverage lesen** (`CoverageRepository`, schreibt nur über `Publish`):
11. `Get(baseline_id, entity_set)` liefert den Eintrag, ein unbekannter ist `ErrNichtGefunden`. `List(baseline_id)` liefert alle Einträge nach `entity_set` aufsteigend, eine Baseline ohne Einträge eine leere Liste, eine unbekannte `ErrNichtGefunden`. Die Evidenz wird byte-gleich gelesen, wie sie geschrieben wurde. Ein leeres `entity_set` oder eine leere Kennung ist `ErrUngueltig`.
12. Ein abgebrochener Kontext beendet jede Operation der vier Repositories mit dem Fehler des Kontexts, ohne eine Änderung.
13. Eine leere Kennung ist bei `Coverage.Get`, `Coverage.List` und `Snapshots.List` `ErrUngueltig` (Nr. 4, 11), bei allen anderen Lesern (`Snapshots.Get`, `Imports.Get`, `Sources`, `BaselinesOf`, `Metadata.Get`) und als Kennung von `SetMetadata` `ErrNichtGefunden`.

Ungeprüft bleibt: dass `MetadataHash` der Baseline zum Hash des Service passt und dass `StateHash` zum Inhalt des Blobs passt (Fachlogik des Imports), das Löschen von Services und Baselines, die Form der Evidenz, dass eine `complete`-Coverage einen Beweis trägt, Kennungen mit NUL-Byte oder nur aus Leerzeichen, und ein Wettlauf zwischen einem Leser, der Kennung und Inhalt in zwei Abfragen liest (`Sources`, `BaselinesOf`, `Coverage.List`), und dem Löschen einer Baseline (das keine Operation bietet).

### SPEC-108 — Workspace, Entity, Cursor und Workspace-Snapshot: Vertrag der Repositories

`WorkspaceRepository`, `EntityRepository`, `CursorRepository`, die Workspace-Operationen des `SnapshotRepository`, `StorageMaintenance` und `WithWorkspaceTransaction` sind die Outbound Ports für den veränderlichen Zustand (`ARC-004`); die Tabellen sind die aus `SPEC-004`, Modell und Begriffe die aus `SPEC-003`.
Sie legen nur ab: Anlage-, Reset- und Restore-Ablauf sind Fachlogik, die diese Operationen in einer Transaktion zusammensetzt.
Kennungen vergibt der Aufrufer; Zeit, Atomarität und die Reihenfolge der Prüfungen (Argument vor Kontext) gelten wie in `SPEC-105`.
Ein Workspace-Snapshot ist unveränderlich; kein Port ändert oder löscht ihn einzeln, nur das Löschen seines Workspace entfernt ihn.
Eine leere Kennung oder ein leerer Schlüssel ist bei jeder Operation, die einen Eintrag liest, ändert oder löscht, `ErrNichtGefunden` (wie in `SPEC-107`); `ErrUngueltig` ist sie nur dort, wo eine Nummer es nennt (`List(service_id)`, `ListSet(entity_set)`) und bei den Einträgen, die `Put`, `ReplaceAll`, `Create` und `CreateWorkspace` schreiben.
Gelesene Listen (`known_properties`, `consumed`, Ergebnislisten) sind bei leerem Inhalt eine leere Liste, nie `nil`.

**Transaktion:**
1. `WithWorkspaceTransaction(ctx, f)` ruft `f` mit Repositories für Workspace, Entity, Cursor und Snapshot, die auf einer Transaktion arbeiten. Liefert `f` `nil`, wird bestätigt, sonst zurückgerollt und der Fehler von `f` unverändert geliefert; dasselbe gilt für eine Panik in `f` (sie wird weitergereicht) und für einen abgebrochenen Kontext (dessen Fehler). Transaktionen laufen nacheinander; der Schreiber hält sie für die Dauer von `f`. Ein `f` ohne Wert (`nil`) ist `ErrUngueltig`; das Argument geht dem Kontext voraus.
   Jede Operation der Repositories bleibt auch in `f` atomar: scheitert sie, nimmt sie ihre eigene Teilwirkung zurück, auch wenn `f` den Fehler verschluckt und `nil` liefert; die Wirkung der übrigen, gelungenen Operationen bestätigt die Transaktion dann.
   Nicht unterstützt: die Repositories nach dem Ende von `f` zu benutzen, eine Transaktion in `f` zu öffnen und in `f` über Repositories außerhalb der Transaktion zu schreiben. Die beiden letzten Fälle warten auf die Sperre des Schreibers und enden nach `busy_timeout` mit dem Fehler der Datenbank (kein Hängen); ein Abbruch des Kontexts wirkt beim Warten erst danach.

**Workspace** (`revision` zählt ab `0`):
2. `Create` legt einen Workspace mit `revision=0` an. Eine leere Kennung, ein leerer `service_id` oder `baseline_snapshot_id`, eine Zeit außerhalb der Jahre 1 bis 9999 oder eine Revision außer `0` ist `ErrUngueltig`. Danach folgen: bekannte Kennung `ErrBereitsVorhanden`, unbekannter Service `ErrDienstUnbekannt`, unbekannte Baseline `ErrNichtGefunden`, ein Snapshot der Art `workspace` als Baseline `ErrFalscheArt`, eine Baseline eines anderen Service `ErrFremderService`; der erste Verstoß bestimmt den Fehler. Der Entity-State entsteht nicht mit.
3. `Get` einer unbekannten Kennung ist `ErrNichtGefunden`; `List(service_id)` liefert die Workspaces eines Service nach `created_at` (als Zeit), dann Kennung, ein unbekannter Service eine leere Liste, ein leerer `service_id` ist `ErrUngueltig`.
4. `Advance(id, erwartet)` erhöht die Revision von `erwartet` auf `erwartet+1` und liefert die neue; ist die gespeicherte Revision eine andere, ist es `ErrRevisionVeraltet` und nichts ändert sich; eine unbekannte Kennung ist `ErrNichtGefunden`, ein negatives `erwartet` `ErrUngueltig`, auch bei leerer oder unbekannter Kennung. Zwei gleichzeitige `Advance` mit derselben Revision: genau eines gelingt.
5. `Delete` löscht den Workspace mit seinem Entity-State, seinen Cursors und seinen Workspace-Snapshots in einer Transaktion; Cursors sind dabei kein Hindernis. Eine unbekannte Kennung ist `ErrNichtGefunden`, andere Workspaces, Baselines und Recording-Sessions bleiben unberührt.

**Entity** (vollständiger Zustand eines Workspace, mit Tombstones):
6. `EntityState` ist gültig, wenn `workspace_id`, `entity_set` und `entity_key` nicht leer sind, `payload` und `known_properties` gültiges JSON in gültigem UTF-8 (`known_properties` ein Array nichtleerer, verschiedener Zeichenketten), `completeness` `complete` oder `partial` und `version` nicht negativ ist; sonst `ErrUngueltig`. Der Schlüssel ist ein vom Aufrufer kanonisch gebildeter Text, den die Ablage nicht deutet.
7. `Put` schreibt einen Eintrag, ersetzt einen vorhandenen mit gleichem `(workspace, entity_set, entity_key)`; ein unbekannter Workspace ist `ErrNichtGefunden`. `Get` liefert den Eintrag, auch einen Tombstone (`deleted=true`); ein unbekannter ist `ErrNichtGefunden`. Payload wird byte-gleich gelesen, `known_properties` aufsteigend sortiert.
8. `List(workspace)` liefert alle Einträge einschließlich Tombstones nach `entity_set`, dann `entity_key` aufsteigend, `ListSet(workspace, entity_set)` die einer Menge; ein unbekannter Workspace ist `ErrNichtGefunden`, ein leerer State eine leere Liste, ein leeres `entity_set` `ErrUngueltig`.
9. `ReplaceAll(workspace, einträge)` ersetzt den gesamten State: danach enthält er genau diese Einträge, ganz oder gar nicht. Ein Eintrag mit anderem Workspace, ein ungültiger oder ein doppelter `(entity_set, entity_key)` ist `ErrUngueltig` und ändert nichts; eine leere Liste leert den State. Ein unbekannter Workspace ist `ErrNichtGefunden`, auch bei leerer Liste; die Prüfung folgt der Gültigkeit der Einträge und dem Kontext.

**Cursor** (`(workspace, replay_session, recording_session)`):
10. `ReplayCursor` ist gültig, wenn die drei Kennungen und `service_id` nicht leer sind, `strategy` eine von `sequential`, `strict`, `first`, `last` ist, `consumed` aus Zahlen größer 0 ohne Doppelte besteht und `position` ihre Anzahl ist; sonst `ErrUngueltig`. `consumed` wird aufsteigend sortiert gelesen.
11. `Put` schreibt oder ersetzt einen Cursor. Prüfung nach der Gültigkeit: unbekannter Workspace `ErrNichtGefunden`, unbekannte Recording-Session `ErrNichtGefunden`, ein `service_id`, der nicht der des Workspace und der Session ist, `ErrFremderService`. `Get` eines unbekannten Cursors ist `ErrNichtGefunden`; `List(workspace)` liefert die Cursors nach `replay_session_id`, dann `recording_session_id`, ein unbekannter Workspace ist `ErrNichtGefunden`.
12. `ResetAll(workspace)` setzt alle Cursors des Workspace auf `position=0` und leeres `consumed`, `strategy` und Kennungen bleiben; ohne Cursors ist es kein Fehler, ein unbekannter Workspace `ErrNichtGefunden`. Das Löschen einer Recording-Session löscht ihre Cursors (`SPEC-105`).

**Workspace-Snapshot** (Operationen des `SnapshotRepository`, nur Art `workspace`):
13. `CreateWorkspace` legt einen Snapshot mit `workspace_id` und `name` an. Gültig sind: nichtleere Kennung, `service_id`, `workspace_id`, `name`, `state_hash` und `metadata_hash`, eine gültige Blob-Referenz (`SPEC-106`) als `state_blob`, eine gültige Zeit; sonst `ErrUngueltig`. Danach: unbekannter Workspace `ErrNichtGefunden`, ein `service_id` außer dem des Workspace `ErrFremderService`, eine bekannte Kennung oder ein Name, den der Workspace schon trägt, `ErrBereitsVorhanden`.
14. `GetWorkspace(workspace, id)`: ein unbekannter Snapshot oder einer eines anderen Workspace ist `ErrNichtGefunden`, ein Snapshot der Art `baseline` `ErrFalscheArt`, auch bei unbekanntem Workspace. `ListWorkspace(workspace)` liefert die Snapshots des Workspace nach `created_at`, dann Kennung; ein unbekannter Workspace ist `ErrNichtGefunden`. Die Operationen der Baselines (`SPEC-107`) liefern nie einen Workspace-Snapshot.

**StorageMaintenance:**
15. `Check` führt `PRAGMA foreign_key_check` und `PRAGMA integrity_check` aus und liefert die Befunde als Liste von Texten; eine Datenbank ohne Verletzung liefert eine leere Liste.
16. `ReferencedBlobs` liefert die Menge der Blob-Referenzen, die ein Exchange (Request oder Response), ein Snapshot jeder Art oder die Metadaten eines Service tragen, ohne Doppelte aufsteigend sortiert.
17. Ein abgebrochener Kontext beendet jede Operation mit dem Fehler des Kontexts, ohne eine Änderung.

Ungeprüft bleibt: Kennungen, Namen und Schlüssel mit NUL-Byte oder ungültigem UTF-8 (die Ablage speichert und liest sie byte-gleich), dass `entity_key` kanonisch ist, dass `state_hash` zum Blob passt und dass `position` bei einem Fortschreiben des Cursors nur wächst (Fachlogik des Replay); dass eine Baseline eines anderen Service beim Restore `ErrFremderService` statt `ErrNichtGefunden` liefert (sie ist über den Workspace gebunden); dass ein Aufrufer außerhalb von `WithWorkspaceTransaction` Reset und Restore ohne Transaktion zusammensetzt.

### SPEC-109 — Fehlermodell: Codes, Fehlerobjekt und Abbildung der Port-Fehler

Das Modell trägt zu jedem Code aus §4 einen HTTP-Status und einen Zielbereich; einen Code außerhalb von §4 gibt es nicht.
Die Abbildung ordnet den Fehlern der Outbound Ports (`SPEC-105` bis `SPEC-108`, `SPEC-103`) einen Code zu; sie ist eine reine Funktion ohne Seiteneffekt.

1. **Fehlerobjekt** (`SPEC-006`): `{"error":{"code","message","target"}}` mit genau diesen drei Feldern, alle Zeichenketten. Der Status ist der der HTTP-Antwort und steht nicht im Body. Korrelationsfelder (`LH-QA-05.a`) trägt der HTTP-Adapter, nicht das Modell.
2. **Status und Zielbereich je Code.** `Target` ist einer von `config`, `admin`, `record`, `storage`, `upstream`, `replay`, `service`, `mock`, `import`, `workspace`, `session`, `internal`. `–` heißt: der Code wird nie als HTTP-Antwort ausgeliefert (Status `0` im Modell).

   | Code | Status | Target | Code | Status | Target |
   |---|---|---|---|---|---|
   | `ODR-E2001` | – | `config` | `ODR-E6004` | 422 | `import` |
   | `ODR-E2101` | 401 | `admin` | `ODR-E6101` | 422 | `import` |
   | `ODR-E2401` | – | `record` | `ODR-E6102` | 409 | `import` |
   | `ODR-E3001` | 500 | `storage` | `ODR-E7001` | 404 | `workspace` |
   | `ODR-E4001` | 502 oder 504 | `upstream` | `ODR-E7002` | 404 | `workspace` |
   | `ODR-E5001` | 404 | `replay` | `ODR-E7003` | 404 | `session` |
   | `ODR-E5002` | 410 | `replay` | `ODR-E7004` | 409 | `session` |
   | `ODR-E5003` | 400 | `replay` | `ODR-E7005` | 404 | `import` |
   | `ODR-E5004` | 409 | `replay` | `ODR-E7006` | 409 | `workspace` |
   | `ODR-E5005` | 404 | `service` | `ODR-E7007` | 409 | `workspace` |
   | `ODR-E5006` | 400 | `replay` | `ODR-E7008` | 422 | `workspace` |
   | `ODR-E5007` | 409 | `replay` | `ODR-E7009` | 409 | `session` |
   | `ODR-E5008` | 422 | `replay` | `ODR-E7010` | 409 | `session` |
   | `ODR-E5009` | 409 | `record` | `ODR-E7011` | 409 | `session` |
   | `ODR-E6001` | 501 | `mock` | `ODR-E7012` | 404 | `session` |
   | `ODR-E6002` | 501 | `mock` | `ODR-E9001` | 500 | `internal` |
   | `ODR-E6003` | 422 | `mock` | | | |

   `ODR-E4001` trägt 502 (nicht erreichbar, Zertifikatsfehler) oder 504 (Zeitüberschreitung); den Status wählt der Aufrufer, jeder andere ist ungültig.
   `ODR-E6004` trägt 422 nur als Antwort auf einen Import; beim Start ist er ohne Status.
3. **`message`** ist ein kurzer fester Text je Code, nie ein Wert aus Request, Daten oder Konfiguration; sein Wortlaut ist kein Vertrag, Clients entscheiden am Code. Die Ursache steht im Log (`error_code`, §4), nie im Body.
4. **Gegenstand.** Die Abbildung erhält den Fehler und den Gegenstand, auf den sich die Operation bezog: `Session`, `Workspace`, `Snapshot` (auch eine Baseline), `Snapshotname`, `Import`, `Exchange`, `Dienst` oder keinen (`Sonstiges`). Den Gegenstand kennt der Aufrufer, nicht das Repository.
5. **Abbildung.** Der erste zutreffende Eintrag der Tabelle bestimmt den Code (Prüfung mit `errors.Is`, auch durch Verpacken und `errors.Join`); ein Gegenstand, den die Zeile nicht nennt, fällt auf `ODR-E9001`.

   | Fehler | Gegenstand → Code |
   |---|---|
   | `ErrBlobFehlt`, `ErrBlobBeschaedigt`, `ErrBeschaedigt` | jeder → `ODR-E3001` |
   | `ErrPfad` | jeder → `ODR-E2001` |
   | `ErrZuGross` | jeder → `ODR-E2401` |
   | `ErrFremderService` | jeder → `ODR-E5004` |
   | `ErrDienstUnbekannt` | jeder → `ODR-E5005` |
   | `ErrFalscheArt` | jeder → `ODR-E7008` |
   | `ErrReferenziert` | jeder → `ODR-E7004` |
   | `ErrSessionGeschlossen` | jeder → `ODR-E5009` |
   | `ErrUngueltigerUebergang` | jeder → `ODR-E7011` |
   | `ErrNichtGefunden` | `Session` → `ODR-E7003`; `Workspace` → `ODR-E7001`; `Snapshot` → `ODR-E7002`; `Import` → `ODR-E7005`; `Exchange` → `ODR-E7012`; `Dienst` → `ODR-E5005` |
   | `ErrBereitsVorhanden` | `Session` → `ODR-E7010`; `Workspace` → `ODR-E7007`; `Snapshotname` → `ODR-E7006`; `Import` → `ODR-E6102` |
   | `ErrUngueltig` (Modell) | `Import` → `ODR-E6101` |
   | `ErrUngueltigeSequenz`, `ErrAbgeschlossen`, `ErrRevisionVeraltet` | jeder → `ODR-E9001` |
   | jeder andere Fehler | jeder → `ODR-E9001` |

   `ErrNichtGefunden` für `Sonstiges` (Cursor, Coverage, Entity) und `ErrUngueltig` außerhalb des Imports sind Zustände, die der Service fachlich deutet (fehlender Cursor, `ODR-E6003`, Antwort des Mock) oder Programmfehler; die Abbildung macht daraus `ODR-E9001`, bis ein Vertrag es anders bestimmt.
   `ErrUngueltigeSequenz` ist ein Programmfehler des Aufrufers, der die Folge zu ordnen hat (`SPEC-105` Nr. 8); `ErrRevisionVeraltet` bekommt einen Code mit dem Vertrag, der es auswertet.
6. **Kein Code:** `nil` ist kein Fehler; ein abgebrochener oder abgelaufener Kontext (`context.Canceled`, `context.DeadlineExceeded`) ist der Abbruch des Aufrufers und bekommt keinen `ODR-`-Code; die Abbildung meldet in beiden Fällen „kein Code“ und gibt keinen Fehlerwert zurück.
   Der Abbruch des Aufrufers gilt vor jeder Zeile von Nr. 5: enthält ein Fehler (verpackt oder in einem `errors.Join`) einen Kontextfehler, ist das Ergebnis „kein Code“, auch wenn zugleich ein Fehler der Tabelle darin steht.
7. **Adapter.** Ein Fehler eines Adapters bildet nur ab, wenn er einen Fehler des Port-Pakets verpackt; das Paket führt dafür `ErrBeschaedigt` und `ErrPfad`. Die Fehler des SQLite-Adapters verpacken sie: `ErrBeschaedigt`, `ErrSchemaNeuer` und `ErrFremdeDatenbank` den Fehler `ErrBeschaedigt` (`SPEC-103` Nr. 4 bis 6, Datei nicht ladbar), `ErrPfad` den Fehler `ErrPfad` (`SPEC-103` Nr. 7). Der Name des Adapters bleibt, `errors.Is` trifft beide.
8. **Vollständigkeit.** Jede Variable `Err…` des Port-Pakets und des SQLite-Adapters hat einen Eintrag in Nr. 5, oder die Tests schlagen fehl; der Test findet die Variablen im Quelltext und braucht keine Liste von Hand.

Ungeprüft bleibt: der Code der Fehler eines Ports, den es noch nicht gibt (Upstream-Transport): er bekommt seine Zeile in Nr. 5 mit dem Port.

### SPEC-110 — Startkonfiguration: Syntax der Werte und Prüfung

Der Composition Root liest die Startkonfiguration aus der Umgebung des Prozesses (Variablen mit dem Namen des Schlüssels aus §3) und aus `SERVICES_FILE`; es gibt keine Kommandozeilen-Optionen und keine `.env`-Datei.
Die Services erhalten typisierte Werte, nie die Umgebung.
Jede Verletzung ist `ODR-E2001` (`SPEC-016`) und verhindert den Start.

1. **Leer.** Leerraum am Rand eines Wertes wird entfernt. Ein danach leerer Wert gilt als ungesetzt (`FOO=` und fehlendes `FOO` sind gleich): es gilt der Default aus §3, und fehlt ein Pflichtwert, ist das `ODR-E2001`. Eine leere Liste ist daher der Default; bei `REDACT_HEADERS`, `REDACT_QUERY_PARAMS` und `REDACT_JSON_PATHS` ergänzt eine gesetzte Liste den Default und ersetzt ihn nie (`LH-FA-14.a`). Eine Menge ohne Einträge drücken nur `MATCH_VARY_HEADERS=none` und `VOLATILE_HEADERS=none` aus; `none` gilt dort nur als einziger Eintrag und genau in dieser Schreibweise. Jeder andere Eintrag, der ohne Beachtung der Schreibweise `none` lautet (`None`, `none,Accept`, `REDACT_HEADERS=none`), ist in jeder Liste ungültig: eine Liste, die `none` als Header- oder Parameter-Namen läse, ließe den Betreiber glauben, er habe den Default ersetzt.
2. **Schlüssel dieser Version.** Geprüft werden `MODE`, `LISTEN`, `DATA_DIR`, `SERVICE_ID`, `SERVICE_ROOT_PATH`, `UPSTREAM_URL`, `SERVICES_FILE`, `REPLAY_STRATEGY`, `REPLAY_ASSIGNMENT`, `REPLAY_REDACTED_BODIES`, `MATCH_VARY_HEADERS`, `MATCH_INVALID_ENCODING`, `MATCH_BODY_MODE`, `REDACT_HEADERS`, `REDACT_QUERY_PARAMS`, `REDACT_JSON_PATHS`, `MAX_CAPTURE_BODY`, `UPSTREAM_TIMEOUT`, `UPSTREAM_CA_FILE`, `UPSTREAM_TLS_VERIFY`, `UPSTREAM_TLS_INSECURE_SKIP_VERIFY`, `SHUTDOWN_TIMEOUT`, `RETENTION_MAX_AGE`, `SQLITE_BUSY_TIMEOUT`, `VOLATILE_HEADERS`, `LINK_REWRITE`, und zwar unabhängig vom Modus.
   Gesetzt (nichtleer) und nicht geliefert sind `ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, `STRICT_COVERAGE` und `PUBLIC_BASE_URL`: das ist `ODR-E2001` („nicht unterstützt“), damit keine Sicherheitseinstellung still wirkungslos bleibt. Jede andere Variable der Umgebung wird ignoriert.
3. **Aufzählung und Wahrheitswert.** Genau die Schreibweise aus §3, Groß- und Kleinschreibung zählt (`Record` ist ungültig). Ein Wahrheitswert ist `true` oder `false`, sonst nichts (`1`, `yes` sind ungültig). `LINK_REWRITE` kennt nur `off` (`SPEC-092`).
4. **Dauer** (`UPSTREAM_TIMEOUT`, `SHUTDOWN_TIMEOUT`, `RETENTION_MAX_AGE`, `SQLITE_BUSY_TIMEOUT`): eine Ganzzahl aus Dezimalziffern, ohne Leerraum direkt gefolgt von `ms`, `s`, `m`, `h` oder `d` (`d` sind 24 Stunden); `30s`, `5000ms`, `30d` und `720h` sind gültig. Ohne Einheit, mit Vorzeichen, Bruch (`1.5h`), Verkettung (`1h30m`) oder anderer Einheit ist ungültig. Der Wert liegt bei `UPSTREAM_TIMEOUT`, `SHUTDOWN_TIMEOUT` und `SQLITE_BUSY_TIMEOUT` bei mindestens 1 ms, und das Ergebnis passt in 63 Bit Nanosekunden; `SQLITE_BUSY_TIMEOUT` ist höchstens 2147483647 ms. `RETENTION_MAX_AGE` lässt `0` zu (aus, `SPEC-058`), mit oder ohne Einheit.
5. **Größe** (`MAX_CAPTURE_BODY`): eine Ganzzahl aus Dezimalziffern, optional nach höchstens einem Leerzeichen `B`, `KiB`, `MiB` oder `GiB` (Faktor 1024); `8 MiB`, `8MiB` und `8388608` (Bytes) sind gleich gültig. Der Wert liegt zwischen 1 Byte und 1 GiB; `0`, Vorzeichen, Bruch und Einheiten wie `MB` sind ungültig.
6. **Liste** (`MATCH_VARY_HEADERS`, `REDACT_*`, `VOLATILE_HEADERS`): Einträge durch Komma getrennt, Leerraum um einen Eintrag wird entfernt; ein leerer Eintrag (`a,,b`, Komma am Ende) ist ungültig. Es sind Mengen: Reihenfolge und Doppelte (bei Headern ohne Beachtung der Schreibweise) haben keine Wirkung. Header-Namen sind HTTP-Token (RFC 9110); ein Query-Parameter-Name enthält weder Leerraum noch `=`, `&`, `#` oder `,`; ein JSON-Pfad hat die Form `$..name` mit `name` aus Buchstaben, Ziffern, `_` und `-` (jede andere Form ist in dieser Version ungültig).
   Die Header-Namen einer Menge liefert die Konfiguration in kanonischer Schreibweise: erster Buchstabe und jeder Buchstabe nach einem `-` groß, alle übrigen klein (`Odata-Version` für `OData-Version`, `X-Api-Key` für `X-API-Key`, `Etag`). Die Schreibweise in §3 und in den Default-Listen nennt dieselbe Menge; die Verwender vergleichen Header-Namen ohne Beachtung der Schreibweise.
7. **Weitere Schlüssel.**
   `LISTEN` ist `host:port` mit nichtleerem Host und Port 1 bis 65535 (nur Dezimalziffern); der Host ist ein Name aus Buchstaben, Ziffern, `.`, `-` und `_` (darunter eine IPv4-Adresse) oder eine IP-Adresse in eckigen Klammern (`[::1]`, ohne Zone). Der Name wird nicht aufgelöst; den Aufbau seiner Labels (Länge, `-` am Rand, leere Labels) prüft die Konfiguration nicht.
   `DATA_DIR` ist ein nichtleerer Pfad ohne NUL-Zeichen, absolut oder relativ zum Arbeitsverzeichnis.
   `SERVICE_ID` und `id` sind 1 bis 64 Zeichen aus Buchstaben, Ziffern, `.`, `_` und `-`.
   `SERVICE_ROOT_PATH` und `service_root` sind `/` oder Segmente `/a/b` ohne abschließenden Schrägstrich; ein Segment ist nichtleer, besteht aus Buchstaben, Ziffern, `.`, `_`, `~` und `-` und ist weder `.` noch `..`.
   `UPSTREAM_URL` und `upstream_url` sind absolute `http`- oder `https`-URLs mit nichtleerem Host, ohne Zugangsdaten (`user:pass@`), Query und Fragment; ein abschließender `/` des Pfads wird entfernt. Ein genannter Port ist 1 bis 65535 (`host:` ohne Ziffern ist ungültig); weitere Zeichenregeln des Hosts sind die der URL-Syntax (RFC 3986), die Konfiguration fügt keine hinzu.
   `SERVICES_FILE`, `UPSTREAM_CA_FILE`: nichtleerer Pfad.
8. **Modus** (`SPEC-075`). `record` und `replay` sind lieferbar. `mock` und `hybrid` sind gültige Werte, aber diese Version liefert den Betrieb nicht: der Start scheitert mit `ODR-E2001` und der Begründung „Modus nicht unterstützt“. Das gilt auch für den Default `mock` ohne gesetzten `MODE`; die Meldung nennt `MODE` und die lieferbaren Werte. Die Syntax aller Schlüssel wird davor geprüft und mitgemeldet.
9. **Services** (`LH-FA-04.c`). Ohne `SERVICES_FILE` gilt genau ein Service aus `SERVICE_ID`, `SERVICE_ROOT_PATH` (Default `/`) und `UPSTREAM_URL`. Ist `SERVICES_FILE` gesetzt, ist jeder nichtleere der drei Einzelschlüssel Doppelkonfiguration und `ODR-E2001`.
   Die Datei ist ein JSON-Array von Objekten mit den Feldern `id`, `service_root` und `upstream_url` (höchstens 1 MiB, UTF-8, nach dem Array nichts als Leerraum). `id` und `service_root` sind Pflicht, auch bei einem einzigen Eintrag; `upstream_url` ist Pflicht im Modus `record`; ein unbekanntes Feld, ein Feld des falschen Typs, ein Element, das kein Objekt ist (auch `null`), ein leeres Array, ein doppeltes `id` oder eine nicht lesbare Datei ist `ODR-E2001`.
   Für die Werte der Felder gilt Nr. 1: Leerraum am Rand wird entfernt, und ein danach leerer Text oder `null` gilt als fehlend (Pflichtfeld fehlt, bei `upstream_url` nur im Modus `record`).
   Ein Feld des falschen Typs (Zahl, Wahrheitswert, Array, Objekt) steht in der Datei und fehlt nicht: es meldet allein „muss eine Zeichenkette sein“ und wird danach nicht weiter geprüft, auch nicht auf „Pflichtfeld fehlt“ oder auf die Form des Wertes; ein Schlüssel hat dann eine Zeile. Das gilt für `id`, `service_root` und `upstream_url` (bei `upstream_url` in jedem Modus, weil der Typ vor der Pflicht des Modus geprüft wird).
   Steht ein Feld in einem Objekt mehrfach, gilt der letzte Wert; das ist kein Verstoß.
   Zwei Roots überlappen, wenn einer dem anderen an einer Segmentgrenze vorausgeht oder beide gleich sind (`/` neben `/a` ja, `/a` neben `/ab` nein); jede Überlappung ist `ODR-E2001`.
   Im Modus `record` braucht jeder Service eine `UPSTREAM_URL`; im Modus `replay` ist sie nicht erforderlich, und eine gesetzte bleibt ohne Wirkung, wird aber auf ihre Form geprüft (Nr. 7), in jedem Modus. Die Pflicht hängt am gültigen Modus: ist `MODE` ungültig, entfällt die Pflichtprüfung von `UPSTREAM_URL`, bis `MODE` behoben ist; ihr Verstoß folgt dann im nächsten Lauf.
10. **Zusammenhang.** `UPSTREAM_TLS_VERIFY=false` ohne `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` und `UPSTREAM_TLS_VERIFY=true` zugleich mit `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` sind Widersprüche (`SPEC-057`, `SPEC-087`) und `ODR-E2001`; ohne `UPSTREAM_TLS_VERIFY` schaltet `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` die Prüfung ab.
    Der Widerspruch wird nur geprüft, wenn beide Schlüssel formal gültig sind (Nr. 3): ein ungültiger Wert (`UPSTREAM_TLS_VERIFY=yes`, `=1`, `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=TRUE`) meldet nur seine Form und gilt für den Zusammenhang weder als `false` noch als `true`. Die Meldung rät dann nie, `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` zu setzen, denn das schaltete die Zertifikatsprüfung wegen eines Tippfehlers ab. Ein Schlüssel meldet damit höchstens eine Zeile.
11. **Dateien und Reihenfolge.** Zuerst die reinen Prüfungen (Nr. 1 bis 10), dann die Dateien: `SERVICES_FILE` wird gelesen (Nr. 9), `UPSTREAM_CA_FILE` muss lesbar sein und mindestens ein PEM-Zertifikat enthalten. `DATA_DIR` legt der Composition Root erst an, wenn alle anderen Prüfungen bestanden sind (Rechte `0700`, mit Elternverzeichnissen, idempotent); ein vorhandener Nicht-Ordner oder ein nicht anlegbarer Pfad ist `ODR-E2001`. Bei einem Verstoß wird nichts angelegt.
12. **Meldung.** Alle Verstöße werden gemeldet, nicht nur der erste, nach Schlüsselname sortiert (Bytefolge des Schlüssels, bei gleichem Schlüssel nach Grund in Bytefolge; diese Folge macht die Ausgabe unabhängig von der Reihenfolge der Prüfungen, in dieser Version meldet jeder Schlüssel höchstens einen Grund (Nr. 9 für ein Feld falschen Typs, Nr. 10 für einen ungültigen Wahrheitswert), die Folge nach Grund ist eine Totalordnung ohne beobachtbaren Fall; Einträge der Datei als `SERVICES_FILE[<index>].<feld>`, der Index zählt als Text: `[10]` steht vor `[2]`; eine Zeile mit gleichem Schlüssel und Grund erscheint einmal), einer je Zeile `ODR-E2001 <Schlüssel>: <Grund>` auf stderr; der Prozess endet mit Exit-Code 1. Die Meldung nennt Schlüssel und Grund, nie den Wert.
13. **Wiederholung.** Dieselbe Umgebung und Datei ergeben dasselbe Ergebnis; das Laden ist synchron ohne Kontext, einen Abbruch durch ein Signal behandelt der Prozess (`LH-FA-02.c`), `DATA_DIR` kann dann angelegt sein.

Ungeprüft bleibt: ob `LISTEN` frei ist, ob ein Hostname aufgelöst wird und ob `UPSTREAM_URL` erreichbar ist; das trifft der Start der Listener bzw. die erste Anfrage. Ebenso die Größe von `UPSTREAM_CA_FILE`: die Datei liegt in der Hand des Betreibers, wird einmal beim Start gelesen, und ein zu großer oder endloser Inhalt ist ein Betreiberfehler, der den Start nicht übersteht.

### SPEC-114 — Zeit und Kennungen: Vertrag von `Clock` und `IDGenerator`

`Clock` und `IDGenerator` (`ARC-019`, `ARC-021`) liefern jede Zeit und jede vom Server vergebene Kennung (`LH-QA-02.a`); kein Modell, Service oder Adapter liest die Systemuhr oder eine Zufallsquelle selbst.
Die Ports haben weder Argumente noch Kontext noch Fehlerrückgabe und sind für gleichzeitigen Aufruf sicher; ihre Test-Implementierungen `FestUhr` und `FestIDs` (Paket `supporttest`) erfüllen denselben Vertrag, soweit unten nichts anderes steht.

**Uhr** (`Now() time.Time`):
1. `Now()` liefert die aktuelle Zeit in UTC (Ort `time.UTC`), ohne monotone Uhrablesung und ohne Rundung: die Auflösung ist die der Plattform, bis zu Nanosekunden. Die Speicherung behält sie (`SPEC-105`: neun Nachkommastellen).
2. Eine Uhr liefert nie weniger als ein Aufruf derselben Uhr, der vorher zurückkehrte. Springt die Systemzeit zurück, bleibt `Now()` auf dem bisherigen Höchstwert stehen, bis die Systemzeit ihn wieder erreicht; so liegt `finished_at` nie vor `started_at` (`SPEC-105` Nr. 10). Zwei Aufrufe dürfen denselben Wert liefern: die Zeit ordnet keine Exchanges, das tut `sequence_no` (`SPEC-105` Nr. 8).
3. Die Systemzeit wird nicht auf Bereich geprüft. Die Nullzeit und eine Zeit außerhalb der Jahre 1 bis 9999 kann nur eine verstellte Systemuhr liefern; `Now()` reicht sie (in UTC) unverändert weiter, und der aufnehmende Port weist sie als `ErrUngueltig` ab (`SPEC-105` Nr. 1), statt dass ein Ersatzwert sie versteckt. Nr. 2 gilt auch dann; ein Neustart des Prozesses beendet das Stehenbleiben auf einem solchen Höchstwert.
4. `Clock` kennt weder Warten noch Timer; Fristen laufen über den Kontext.
5. `FestUhr`: `NewFestUhr(start)` rechnet `start` nach UTC um. Die Nullzeit und ein Jahr außerhalb 1 bis 9999 **nach** der Umrechnung sind ein Programmfehler des Tests und lösen eine Panik aus (zuerst die Nullzeit, dann der Bereich). `Now()` liefert den Stand und ändert ihn nicht; `Advance(d)` rückt ihn um `d` vor. `d == 0` ist erlaubt; `d < 0` oder ein Ergebnis nach dem Jahr 9999 ist eine Panik und ändert den Stand nicht. Gleichzeitige `Advance` summieren sich.

**Kennungen** (`IDGenerator`):
6. Fünf Operationen ohne Argument, jede liefert eine Zeichenkette: `SessionID` (`ses`), `ExchangeID` (`exc`), `KorrelationsID` (`cor`), `WorkspaceID` (`wsp`), `SnapshotID` (`snp`); `SnapshotID` dient Baseline- und Workspace-Snapshots. Die Form ist `<Präfix>-<32 Zeichen aus 0-9a-f>`, genau 36 Zeichen aus `[a-z0-9-]`, damit sie ohne Maskierung in URL-Pfad, Header, Dateiname und Logzeile steht. Eine Kennung ordnet nichts: geordnet wird nach Zeit bzw. `sequence_no`, dann nach Kennung (`SPEC-105` Nr. 4, 11).
7. Die Zeichen sind 128 Bit aus `crypto/rand`, ohne Seed und ohne Konfiguration. Fällt die Quelle aus, beendet die Laufzeit den Prozess (seit Go 1.24 gibt es keinen Fehlerwert); deshalb hat der Port keine Fehlerrückgabe.
8. Zwei Kennungen derselben Operation stimmen mit vernachlässigbarer Wahrscheinlichkeit überein (bei 128 Zufallsbits wird eine Kollision erst bei rund 2^64 Kennungen wahrscheinlich); Kennungen verschiedener Operationen unterscheiden sich durch das Präfix. Der Generator prüft nicht gegen den Speicher, und die Verbraucher wiederholen nicht; trifft ein Port dennoch `ErrBereitsVorhanden` auf eine erzeugte Kennung, gilt `SPEC-109` wie für jede andere.
9. Kennungen, die der Client vergibt (Replay-Session-ID, `import_id`, eine mitgegebene Workspace- oder Session-`id`), erzeugt, prüft und ändert der Generator nicht; ihre Form (Länge, Alphabet) legt der Vertrag des Use Case fest, der sie annimmt.
10. `FestIDs`: dieselbe Form (Nr. 6), die 32 Zeichen sind ein Zähler je Operation, der je Instanz bei 1 beginnt und um 1 steigt (`ses-00000000000000000000000000000001`, `ses-…02`). Der Zähler steht **dezimal** mit führenden Nullen auf 32 Stellen (`%032d`): auf `…09` folgt `…10`, nicht `…0a`; so bleibt die Kennung lesbar und die Folge im Test vorhersagbar, und die Form von Nr. 6 bleibt erfüllt. Zwei Instanzen zählen unabhängig; bei gleichzeitigem Aufruf erhält jeder Aufruf genau eine Nummer, ohne Lücke und ohne Doppelung.

Ungeprüft bleibt: die Genauigkeit der Plattformuhr und der Abgleich ihrer Zeit (NTP), die kryptografische Güte der Zufallsquelle (kein Test kann sie beobachten), das Verhalten bei Ausfall der Zufallsquelle (Nr. 7; es liegt bei der Go-Laufzeit, ein Test brauchte einen Unterprozess) sowie das Ausbleiben von Datenrennen über das Verhalten hinaus, das die Tests unter Last zeigen; das Gate fährt nicht mit dem Race-Detektor.

### SPEC-115 — Redaction von Headern, Query und `Location`, Pfadmuster: Vertrag der Funktionen

Die Funktionen sind rein: ohne Ein-/Ausgabe, ohne Kontext, ohne Uhr und Zufall, und sie lassen ihre Eingaben unverändert (das Ergebnis ist eine Kopie).
Die Mengen der Namen sind die wirksamen (Default plus Konfiguration, `SPEC-110` Nr. 1 und 6) und kommen als Argument.
Kein Fehler dieser Funktionen enthält einen Wert aus der Eingabe, nur den Namen des Feldes.

1. **Marker.** `[REDACTED]` (ASCII, zehn Zeichen, unkodiert) ist der einzige Platzhalter eines Header-, Query- und JSON-String-Wertes (`LH-FA-14.a`).
2. **Header.** `RedigiereHeader(Header, Namen, Parameter)` liefert eine neue Map, nie `nil`; eine leere oder `nil`-Eingabe ergibt eine leere Map.
   Ein Name zählt als redigiert, wenn er ohne Beachtung der ASCII-Schreibweise einem Namen aus `Namen` gleicht (`REDACT_HEADERS` samt Default, `SPEC-060`).
   Die Namen des Ergebnisses stehen in kanonischer Schreibweise (`SPEC-110` Nr. 6), sofern der Name ein HTTP-Token ist (der leere Name ist keines); ein anderer Name bleibt unverändert, wird nur nach seiner ASCII-Schreibweise mit `Namen` verglichen und nie mit einem anderen Namen zusammengeführt: `bad name` und `BAD NAME` bleiben zwei Einträge, jeder für sich redigiert oder nicht. Ein Nicht-Token-Name steht in keiner Liste der Konfiguration (`SPEC-110` Nr. 6) und geht nie in den Match-Key ein; ein kanonischer Name, unter dem man ihn zusammenführen könnte, existiert nicht.
   Token-Namen, die sich nur in der Schreibweise unterscheiden (`cookie`, `Cookie`), werden zu einem Eintrag: die Werte folgen einander in der Bytefolge der Eingabenamen, danach in ihrer Reihenfolge.
   Ein redigierter Header trägt im Ergebnis genau einen Wert, den Marker, gleich ob die Eingabe keinen, einen leeren, einen oder mehrere Werte hatte.
   Ein nicht redigierter Header behält Werte und Reihenfolge, auch eine leere Liste und leere Werte.
3. **`Location` und `Content-Location`.** Für diese beiden Header (Name ohne Beachtung der ASCII-Schreibweise, in Request und Response gleich), die nicht nach Nr. 2 ganz redigiert sind, gilt je Wert: der Teil nach dem ersten `?` bis zum ersten folgenden `#` (oder zum Ende) ist ein `raw_query` und wird nach Nr. 4 mit `Parameter` (`REDACT_QUERY_PARAMS`, `SPEC-084`) redigiert, der Rest bleibt unverändert; ohne `?` bleibt der Wert unverändert. Das erste `?` zählt auch dann, wenn davor ein `#` steht (`/p#f?token=1` ergibt `/p#f?token=[REDACTED]`): im Zweifel wird mehr redigiert, nie weniger.
   Ein so veränderter Header ist **nicht** „redigiert“ im Sinn von Nr. 5: er wird im Replay mit dem Marker im Query-Wert ausgeliefert, weil der Header selbst ausgeliefert werden muss (eine Weiterleitung oder `201 Created` ohne `Location` ist keine Wiederholung) und das Geheimnis nicht wiederhergestellt wird.
   Andere Header mit URL-Inhalt (`Referer`, `Link`, `Origin`) werden nicht untersucht (`LH-FA-14.a` Geltungsbereich).
4. **Query.** `RedigiereQuery(RawQuery, Parameter)` liefert den redigierten `raw_query` und ob sich dabei ein Wert änderte.
   - `raw_query` trägt kein führendes `?`; ein führendes `?` ist Teil des ersten Schlüssels. Zerlegt wird nur an `&` (nicht an `;`) in Segmente, auch leere; die Segmente bleiben in ihrer Reihenfolge und werden mit `&` verbunden, ein leerer Eingabetext ergibt `""`.
   - Ein Segment ohne `=` hat keinen Wert und bleibt unverändert, auch wenn sein Text ein Name aus `Parameter` ist. Sonst ist der Text bis zum ersten `=` der Schlüssel, der Rest der Wert.
   - Der Schlüssel wird `%XX`-dekodiert (`+` bleibt `+`, Hex-Schreibweise gleich) und ohne Beachtung der ASCII-Schreibweise mit `Parameter` verglichen; ist die Kodierung ungültig (`SPEC-116` Nr. 4), wird der Schlüsseltext unverändert verglichen.
   - Bei einem Treffer lautet das Segment `<Schlüsseltext unverändert>=[REDACTED]`, auch bei leerem Wert; jedes Vorkommen eines Parameters wird ersetzt. Die Änderung ist „ein Wert änderte sich“, wenn der Wert vorher nicht schon der Marker war.
   - Die Redaction hängt nicht von `MATCH_INVALID_ENCODING` ab und scheitert nie.
   - Wiederholung: `RedigiereQuery` und `RedigiereHeader` auf ihr eigenes Ergebnis ergeben dasselbe Ergebnis, und die zweite Anwendung meldet keine Änderung.
5. **Ausgelieferte Header.** `OhneRedigierte(Header)` liefert die Map ohne jeden Header, der mindestens einen Wert hat und dessen Werte alle der Marker sind (`SPEC-045`: redigierte Header werden weggelassen, nie mit Marker ausgeliefert). Ein Header mit leerer Wertliste, mit einem leeren Wert oder mit dem Marker als Teil eines längeren Wertes bleibt.
6. **Pfadmuster.** `ParseJSONPfad(Text)` nimmt die Form `$..name` an, mit `name` aus einem oder mehr ASCII-Buchstaben, Ziffern, `_` und `-` (`SPEC-110` Nr. 6); jede andere Form, auch mit Leerraum am Rand, ohne Namen, mit weiterem Segment (`$..a.b`), mit Index oder Platzhalter, ist `ErrUngueltig`. Dieselbe Funktion prüft `REDACT_JSON_PATHS` beim Start.
   Bedeutung: ein Mitglied eines Objekts, dessen Name dem `name` ohne Beachtung der ASCII-Schreibweise gleicht, in beliebiger Tiefe; ein String-Wert wird durch den Marker ersetzt, die Behandlung anderer Wertarten und des Stroms regelt `SPEC-118`.
7. **`key_redacted`** (`SPEC-001`) ist wahr, wenn die Query-Redaction einen Wert änderte oder `redaction.request=applied` ist (`SPEC-116` Nr. 7). Header und `Location` verändern keinen Eingang des Match-Keys.

Ungeprüft bleibt: Geheimnisse an Stellen, die keine Liste nennt (Pfad, `$filter`-Werte, nicht gelistete Parameter, andere Header); Unicode-Gleichheit über die ASCII-Schreibweise hinaus (`Ä` und `ä` sind verschiedene Namen).

### SPEC-116 — Match-Key: Kodierung und Randformen

`BildeSchluessel(Eingabe, Regeln)` ist die einzige Stelle, die einen Match-Key bildet (Record, Replay und Import rufen sie; `LH-FA-14.a`).
Sie nimmt den Request **unredigiert** und redigiert selbst nach `SPEC-115`; so ist die Darstellung, aus der der Key entsteht, die, die persistiert wird, und ein Aufrufer kann den Key für Header und Query nicht aus einer unredigierten Form bilden.
Die Grenze ist der Body: die Funktion redigiert ihn nicht und kann nicht sehen, ob er redigiert ist; die Zusage, dass kein Geheimnis aus dem Body in den Key eingeht, trägt der Aufrufer, der die Bytes durch den Redaktor von `SPEC-046` leitet (Record und Replay, `LH-FA-14.a`). Ein Test der Funktion kann sie nicht fangen; ein Body mit Geheimnis ergäbe einen Key, der vom Geheimnis abhängt (als SHA-256, nie im Klartext), und einen Replay, der bei anderem Geheimnis nicht findet.
Eingabe: `Method`, `ServiceID`, `Path` (der OData-Pfad nach dem Service-Root wie in `exchanges.path`, `LH-FA-04.c`), `RawQuery`, `Header` des Requests, `Body` (die Bytes des **bereits redigierten** Request-Bodys) und `RequestRedaction` (`none` oder `applied`).
Regeln: `MATCH_VARY_HEADERS`, `VOLATILE_HEADERS`, `REDACT_HEADERS`, `REDACT_QUERY_PARAMS`, `MATCH_INVALID_ENCODING`, `MATCH_BODY_MODE` (`SPEC-061`, `SPEC-063`, `SPEC-060`, `SPEC-084`, `SPEC-081`, `SPEC-082`); ein anderer Wert der beiden Aufzählungen, auch der leere, ist `ErrUngueltig`.
Ergebnis: `MatchKey`, `KeyRedacted`, der redigierte `RawQuery` und die redigierten `Header`.
Die Kodierung ist Teil des Aufzeichnungsformats (`SPEC-015`): eine Änderung ist eine neue Format-Version.

1. **Kodierung.** Der Key ist `sha256:` und die 64 Kleinbuchstaben-Hex-Ziffern des SHA-256 über die Bytefolge: die Marke `odr-match-v1`, `Method`, `ServiceID`, der normalisierte Pfad (Nr. 3), die Paare (Nr. 4), die Vary-Header (Nr. 5), der Body (Nr. 6).
   Eine Zeichenfolge steht als Länge (4 Byte, Big Endian) und ihre Bytes, eine Anzahl als 4 Byte Big Endian; so ist die Folge eindeutig lesbar und zwei verschiedene Tupel ergeben nie dieselbe Folge. Als Zeichenfolge stehen auch die Marke und der Text eines Paares der Art 2 (Nr. 4); allein die Art-Bytes (1 Byte) und die Digests (32 Byte) haben ein festes Maß ohne Länge. Alle Hex-Ziffern, die der Key selbst erzeugt, sind klein (`sha256:…` und `\u00XX` der Kanonform, Nr. 6); allein die Escapes des normalisierten Pfades sind groß (Nr. 3).
2. **Methode und Service.** `Method` und `ServiceID` gehen unverändert ein, die Schreibweise zählt (`get` ≠ `GET`); leer ist `ErrUngueltig`. Weitere Formprüfungen gibt es hier nicht: die Konfiguration prüft die Kennung, der Adapter die Methode (`SPEC-110` Nr. 7).
3. **Pfad.** `Path` ist nichtleer, beginnt mit `/` und enthält weder `?` noch `#` noch ein Steuerzeichen (U+0000 bis U+001F, U+007F); sonst `ErrUngueltig`. Der Service-Root ist nicht Teil des Pfades; der Aufruf des Roots selbst hat den Pfad `/`.
   Normalisiert werden allein `%XX`-Folgen: die Hex-Ziffern werden groß geschrieben, und ein Escape eines ASCII-Zeichens aus `A-Z a-z 0-9 - . _ ~` wird durch das Zeichen ersetzt; jeder andere Escape bleibt (`%2F` bleibt ein Escape und wird nie zur Segmentgrenze, `%27` bleibt ungleich `'`, `%20` ungleich `+`).
   Alles Übrige bleibt byteweise: Groß-/Kleinschreibung, doppelte und abschließende Schrägstriche, die Segmente `.` und `..` (es wird nichts aufgelöst), Nicht-ASCII-Bytes und OData-Literale.
   Der Pfad wird nicht dekodiert, nur die Form der Escapes normalisiert; ein Escape, dessen Byte kein UTF-8 ergibt (`%FF`), ist gültig und bleibt (`%ff` wird `%FF`).
   Ein `%`, dem keine zwei Hex-Ziffern folgen, ist eine ungültige Kodierung (Nr. 4): bei `literal` bleibt das Zeichen `%` stehen, bei `reject` ist das Ergebnis `ErrKodierung`.
4. **Query-Paare** aus dem **redigierten** `raw_query` (Nr. 8). Leere Segmente (`a=1&&b`, Anfang, Ende, ein `raw_query` aus `""`) ergeben kein Paar. Ein Segment ohne `=` ist ein Paar der Art 0 (Schlüssel); mit `=` eines der Art 1 (Schlüssel und Wert, geteilt am ersten `=`, so ist `a=b=c` der Wert `b=c`, und `=v` der leere Schlüssel); `a` und `a=` sind verschieden, `;` trennt nichts.
   Schlüssel und Wert werden getrennt dekodiert: `%XX` wird zu Bytes (Hex-Schreibweise gleich), `+` bleibt `+`; ungültig ist die Kodierung, wenn ein `%` nicht von zwei Hex-Ziffern gefolgt wird oder das Ergebnis (oder ein Roh-Byte) kein gültiges UTF-8 ist.
   Ein ungültiges Paar ist bei `literal` ein Paar der Art 2 mit dem Segmenttext unverändert, bei `reject` ist das Ergebnis `ErrKodierung`; die Prüfung gilt für die redigierte Darstellung, ein ungültiger Wert unter einem redigierten Parameter ist damit weg (`token=%zz` ist `token=[REDACTED]`).
   Die Paare werden aufsteigend nach (Art, Schlüsselbytes, Wertbytes), die Art 2 nach dem Text, sortiert; gleiche Paare bleiben beide (Multimenge). Kodiert: Anzahl, je Paar ein Byte Art (0, 1, 2), dann Schlüssel und bei Art 1 Wert als Zeichenfolge bzw. bei Art 2 der Text als Zeichenfolge. Der Text der Art 2 ist der Segmenttext des redigierten `raw_query` byteweise, auch in der Schreibweise seiner Hex-Ziffern (`a=%2f%zz` und `a=%2F%zz` sind verschiedene Paare der Art 2).
5. **Vary-Header.** Die Menge ist `MATCH_VARY_HEADERS` ohne die Namen aus `VOLATILE_HEADERS` und aus `REDACT_HEADERS`: Authentisierung und flüchtige Header gehen nie in den Key ein (`LH-FA-05.a`, `LH-QA-02.a`); ein Name in beiden Mengen ist kein Fehler, die Ausnahme gewinnt.
   Die Werte stammen aus der **redigierten** Darstellung (Nr. 8): ein Vary-Header mit URL-Inhalt (`Location`, `Content-Location`) geht mit dem Marker im Query-Wert ein (`SPEC-115` Nr. 3), nie mit dem Geheimnis; zwei Requests, die sich nur im Geheimnis dort unterscheiden, ergeben denselben Key.
   Kodiert wird die Anzahl der Namen der Menge, die der Request mit mindestens einem Wert trägt (nicht die der konfigurierten), dann je Name in kanonischer Schreibweise (Bytefolge aufsteigend, unabhängig von der Reihenfolge der Konfiguration): Name als Zeichenfolge, Anzahl der Werte, jeder Wert als Zeichenfolge in der Reihenfolge der Eingabe, Leerzeichen und Tabulatoren an den Rändern eines Wertes entfernt. Schreibweisen desselben Namens werden wie in `SPEC-115` Nr. 2 zusammengeführt.
   Ein fehlender Header (kein Eintrag oder leere Wertliste) steht nicht im Tupel; ein leerer Wert steht darin, ist also vom fehlenden verschieden. Zwei Wertzeilen (`a`, `b`) sind vom Wert `a, b` verschieden. Alle anderen Header gehen nicht ein. Eine leere Menge (`none`) und eine Menge, deren Namen der Request alle nicht trägt, ergeben die Anzahl 0.
6. **Body.** Ein leerer oder `nil`-Body hat die Art 0 und keine Digest-Bytes. Sonst steht ein Byte Art und der SHA-256-Wert (32 Byte):
   - `bytes`: Art 1, SHA-256 der Bytes.
   - `json-canonical`: ist der Body genau ein gültiger JSON-Wert (RFC 8259, Leerraum außen erlaubt, kein BOM, nichts danach, Verschachtelung höchstens 512, Strings gültiges UTF-8 und ohne einzelne Surrogat-Escapes), Art 3 und SHA-256 der Kanonform; sonst Art 2 und SHA-256 der Bytes (kein Fehler: konservativer Rückfall, `LH-FA-05`).
   - Kanonform: ohne Leerraum; Objektmitglieder aufsteigend nach den dekodierten Namensbytes, bei gleichem Namen in Eingabereihenfolge (doppelte Namen bleiben beide); Arrays in Reihenfolge; Strings als dekodierter Wert, mit `\"`, `\\` und `\u00XX` mit kleinen Hex-Ziffern für jedes Zeichen unter U+0020 (auch `\n` und `\t` ohne Kurzform: `\u000a`, `\u0009`; die Eingabeschreibweise `\u001F` oder `\n` ergibt dasselbe), sonst als UTF-8 (U+007F und `/` unverändert); Zahlen als Literal unverändert (`1`, `1.0`, `1e0` und `-0` sind verschieden); `true`, `false`, `null`.
   Der Modus gehört wie `ServiceID` zur Aufzeichnung: ein anderer Modus beim Replay ergibt andere Keys.
7. **`KeyRedacted`** und `RequestRedaction`: `applied` oder eine geänderte Query ergibt `true` (`SPEC-115` Nr. 7); `unsafe` ergibt `ErrUngueltig` ohne Key (`SPEC-001`: der Exchange ist `failed` mit `null`-Key), ebenso jeder andere Wert, auch der leere.
8. **Redigierte Darstellung.** Header und `raw_query` des Ergebnisses sind die nach `SPEC-115` redigierten; Pfad und Body sind nicht redigiert (`LH-FA-14.a`). Alle Eingänge des Keys, die ein Geheimnis tragen können (Query, Vary-Header), stammen aus dieser Darstellung; für den Body gilt die Grenze der Einleitung. Der Key eines bereits redigierten Requests ist der Key des unredigierten (Idempotenz): Record und Replay finden denselben Key, und ein anderes Geheimnis ergibt denselben Key.
9. **Fehler.** `ErrUngueltig` (Nr. 2, 3, 7 und die Aufzählungen der Regeln) ist ein Programmfehler des Aufrufers (`ODR-E9001`, `SPEC-109` Nr. 5). `ErrKodierung` (neu im Modell) tritt nur bei `reject` auf; der Dienst, der den Key bildet, antwortet `ODR-E5006` (HTTP 400, `SPEC-067`), im Record **vor** jeder Weiterleitung, ohne Exchange anzuhängen. Beide Fehler tragen keinen Wert des Requests. Die Prüfungen laufen in dieser Reihenfolge, der erste Verstoß entscheidet: Aufzählungen der Regeln, `RequestRedaction`, `Method`, `ServiceID`, Form des Pfades (Nr. 3), dann die Kodierung des Pfades, dann die der Query (beide `ErrKodierung`).
10. **Gleichwertigkeit.** Die Reihenfolge der Query-Paare, die Schreibweise der Hex-Ziffern (bei gültig kodierten Paaren und im Pfad, nicht im Text der Art 2) und der Header-Namen (der Token-Namen), die Reihenfolge der Einträge der Header-Map und die Hex-Schreibweise im Pfad ändern den Key nicht; Methode, Service, Schreibweise des Pfades, Segmentgrenzen, `+` gegen `%20`, die Werte und Wiederholungen eines Parameters, ein Vary-Header und der Body tun es.

Ungeprüft bleibt: Unicode-Normalisierung (`é` zusammengesetzt und zerlegt sind verschieden, konservativ), semantisch gleiche `$filter`-Ausdrücke (`LH-FA-05.a`), die Größe des Bodys (der Aufrufer begrenzt ihn: `MAX_CAPTURE_BODY` im Record; die Grenze im Replay legt der Dienst fest, der den Request liest) und die Laufzeit der Kanonform (zwei Durchläufe über höchstens diese Größe).

### SPEC-117 — Replayfähigkeit: Funktion über die Fähigkeitsmatrix

`Replayfaehig(Exchange, RedigierteBodies)` setzt die Matrix aus `SPEC-045` um; `RedigierteBodies` ist `REPLAY_REDACTED_BODIES` (`SPEC-083`). Die Matrix bleibt die Quelle; diese Nummer legt fest, wie die Funktion mit ihren Rändern umgeht.

1. Das Ergebnis ist wahr genau dann, wenn `capture_status=complete`, die Integrität `verified` oder `absent` ist und die Response-Body-Redaction `none` ist, oder `applied` bei `RedigierteBodies=true`. Alles andere ist falsch: `incomplete`, `failed`, `invalid`, `unsafe`, `applied` bei `false`, und jeder unbekannte oder leere Wert einer der drei Spalten (die Funktion schließt im Zweifel aus).
2. Nicht befragt werden: die Redaction des Request-Bodys, `key_redacted`, die Header, die Body-Zustände, `replay_equivalence`, `match_key` und die Zeiten; keines dieser Felder ändert das Ergebnis (`SPEC-045`: Request-, Query- und Header-Redaction schränken die Replayfähigkeit nicht ein).
3. Die Funktion prüft den Exchange nicht (`Exchange.Validate` ist eine eigene Frage), liefert nie einen Fehler und ist auch für den Nullwert definiert (falsch). Sie verändert nichts und liefert für dieselbe Eingabe dasselbe.
4. Wahl und Fehlercodes (`ODR-E5001`, `ODR-E5007`, `ODR-E5008`) und Cursor sind nicht Teil der Funktion: sie ist das Prädikat, das der Dienst über die Kandidaten legt (`SPEC-045`, `LH-FA-04.a`).

### SPEC-118 — Streaming-Redaktor für Bodies: Vertrag der Funktion

`RedigiereBody(dst, src, BodyEingang, BodyRegeln)` liest einen Body als Strom aus `src` (`io.Reader`), schreibt den redigierten Body nach `dst` (`io.Writer`) und liefert `(RedactionStatus, error)`. Sie setzt `SPEC-046` Nr. 3 und 4 und `LH-FA-14.a` um; die Pfadmuster sind die von `SPEC-115` Nr. 6.
`BodyEingang{ContentType []string, Metadata bool}`: `ContentType` sind alle Werte des Headers `Content-Type` (leer oder `nil`, wenn er fehlt), `Metadata` ist wahr genau für die Antwort auf die Anfrage `<service_root>/$metadata` (der Pfad nach dem Service-Root ist `/$metadata`, die Query zählt nicht); der Aufrufer setzt es, nie für einen Request-Body.
`BodyRegeln{Pfade []JSONPfad, MaxBytes int64}`: `Pfade` sind die wirksamen Muster (Default plus Konfiguration, `SPEC-085`, `SPEC-110` Nr. 6), `MaxBytes` ist `MAX_CAPTURE_BODY` (`SPEC-013`); die Funktion nimmt jeden Wert ab `1` bis `math.MaxInt64` an (bei `MaxInt64` kommt `zu_gross` nie vor, und die Rechnung läuft nicht über), die Obergrenze von 1 GiB prüft die Konfiguration (`SPEC-013`), nicht die Funktion.
Die Funktion hat keinen Zustand zwischen Aufrufen, keine Uhr, keinen Kontext und startet keine Goroutine; ein Abbruch ist der Fehler von `src`. Ein Aufruf gehört einer Goroutine, `src` und `dst` gehören während des Aufrufs dem Aufrufer.

1. **Argumente.** `MaxBytes` kleiner als `1`, ein `nil` für `dst` oder `src` und ein Nullwert von `JSONPfad` in `Pfade` (er würde nie treffen und die Redaction still abschalten) sind `ErrUngueltig`; der Status ist dann leer, `src` wird nicht gelesen, in `dst` steht nichts. Eine leere oder `nil`-Menge `Pfade` ist gültig: der Body wird geprüft und nichts ersetzt. Diese Prüfung kommt vor jeder anderen.
2. **Leere Eingabe.** Liefert `src` kein Byte, ist das Ergebnis `none` ohne Fehler und ohne Ausgabe, gleich welcher `ContentType` und `Metadata` (kein Body ist nichts zu schützen; ob der Body `absent` ist, entscheidet der Aufrufer). Dies kommt vor der Prüfung des Typs. Ein Body aus nur Leerraum ist nicht leer (Nr. 4: `unvollstaendig`).
3. **Typ.** Der Media-Type ist der Wert bis zum ersten `;` ohne Leerzeichen und Tabulatoren an den Rändern, ohne Beachtung der ASCII-Schreibweise; Parameter (`charset`, `odata.metadata`) werden nicht ausgewertet. Es gilt genau ein Wert in `ContentType`; kein Wert, ein leerer, mehrere Werte oder ein unbekannter Typ sind `unsafe` mit Grund `typ`, ohne Ausgabe.
   - **JSON:** `application/json` und `application/<x>+json` mit nichtleerem `<x>`. Der Typ lässt den Body nur in die strenge Prüfung (Nr. 4), eine falsche Auszeichnung fällt dort durch (`syntax`).
   - **EDMX:** `application/xml` und `Metadata=true`. Der Body geht byte-gleich nach `dst`, Status `none`, ohne Umschreiben (`SPEC-046` Nr. 3). Das erste Byte, das weder Leerraum (Leerzeichen, Tabulator, LF, CR) noch ein einmaliges UTF-8-BOM am Anfang ist, muss `<` sein, sonst ist der Strom `unsafe` mit Grund `mismatch` (auch wenn er davor endet), und dieses Byte wird nie geschrieben; so geht ein als XML ausgezeichneter JSON-Body nicht unredigiert durch.
   - **Unbekannt:** alles andere, darunter `application/xml` bei `Metadata=false`, `text/*` (auch `text/xml`, `text/json`), `multipart/mixed`, `application/x-www-form-urlencoded`, `application/octet-stream`, `application/json-seq`, `application/jsonx`. `Metadata=true` ändert den Typ JSON nicht.
4. **JSON: Prüfung.** Der Body ist genau ein JSON-Wert nach RFC 8259 in dieser strengen Lesart, jeder Verstoß ist `unsafe`:
   - Leerraum außerhalb von Strings ist Leerzeichen, Tabulator, LF, CR, auch vor dem Wert und danach; jedes andere Byte außerhalb von Strings, das die Grammatik nicht erlaubt, ist `syntax` (darunter BOM, NUL, `NaN`, einfache Anführungszeichen, ein Komma vor `}` oder `]`, `True`, `01`, `1.`, `.5`, `+1`, ein zweiter Wert nach dem ersten, Nachlauf jeder Art).
   - In einem String ist ein rohes Byte unter `0x20` `syntax`; `0x7F` ist erlaubt. Ein Escape ist `\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t` oder `\u` mit genau vier Hex-Ziffern in beliebiger Schreibweise, jeder andere `syntax`. Ein einzelnes Surrogat als `\uD800` ist gültig (anders als in `SPEC-116` Nr. 6: hier prüft nur die Grammatik, und ein Surrogat gleicht nie einem ASCII-Namen).
   - Der Inhalt eines Strings, Namen und Werte, ist gültiges UTF-8 (nach den Regeln von Go: kein überlanger Code, kein Surrogat, höchstens U+10FFFF), sonst `utf8`; das gilt auch für den Inhalt eines Strings, der ersetzt wird. Eine Mehrbyte-Folge darf an jeder Stelle zwischen zwei Reads geteilt sein.
   - Die Verschachtelung ist höchstens 512; die 513. offene Klammer ist `tiefe`.
   - Endet der Strom, bevor der Wert vollständig ist (offener String, offene Klammer, unvollständige Zahl wie `-` oder `1e`, unvollständiges Literal, fehlender Wert nach `:` oder `,`), ist das `unvollstaendig`; eine Zahl, die am Ende des Stroms gültig steht (`12` als ganzer Body), ist vollständig. Ein Strom aus nur Leerraum ist `unvollstaendig`.
   - Zahlen und Literale gehen unverändert durch; es gibt keine Grenze ihrer Länge außer `MaxBytes`.
   - Unter einem treffenden Namen (Nr. 5) entscheidet das erste Byte des Wertes vor der Zahlen- und Literalgrammatik: beginnt er mit Ziffer, `-`, `{` oder `[`, ist es `nichtstring`, auch wenn der Rest ungültig wäre (`01`, `1.`, `1e`, `-` sind dort `nichtstring`, unter einem nicht treffenden Namen `syntax` bzw. `unvollstaendig`); jeder andere Anfang, der kein String ist (`x`, `.5`, `+1`, `NaN`), folgt der Grammatik dieser Nummer.
5. **JSON: Redaction.** Ein Mitglied eines Objekts trifft, wenn sein Name, **nach** der Dekodierung aller Escapes (`password` ist `password`), über `JSONPfad.Passt` (ASCII-Schreibweise, `SPEC-115` Nr. 6) einem Muster aus `Pfade` gleicht, in beliebiger Tiefe, auch im Wert eines nicht treffenden Mitglieds und in einem Objekt in einem Array. Ein String, der kein Mitgliedsname ist (Element eines Arrays, Wert, Wurzel), trifft nie. Mehrere Mitglieder gleichen Namens treffen je für sich. Der Wert eines Treffers:
   - **String:** er wird zu `"[REDACTED]"` (Marker, `SPEC-115` Nr. 1); sein Inhalt wird geprüft (Nr. 4), aber nie geschrieben und nie gepuffert. Ist der Rohtext des Inhalts, ohne Escapes, schon genau `[REDACTED]`, ist das keine Änderung, der Status bleibt `none` (`SPEC-115` Nr. 4, Idempotenz Nr. 8); jeder andere Inhalt, darunter `\u005bREDACTED]` (mit Escape), `[REDACTED] ` und der leere String, ist eine Änderung (`applied`).
   - **`null`, `true`, `false`:** unverändert; sie tragen kein Geheimnis, und der Marker ersetzt nur Strings (`LH-FA-14.a`).
   - **Zahl, Objekt, Array:** `unsafe` mit Grund `nichtstring`, entschieden am ersten Byte des Wertes: mit dem einen erlaubten Platzhalter ließe sich ihr Inhalt nicht verbergen, ohne den Typ zu ändern, und unverändert wäre das Geheimnis offen (fail-closed, `LH-FA-14`). Ein Treffer innerhalb eines Treffers (`{"secret":{"password":"x"}}`) ist daher `nichtstring` am äußeren.
   - Alle anderen Bytes gehen **byte-gleich** nach `dst`: Leerraum, Reihenfolge, Namen samt ihrer Escapes, Zahlenliterale, Strings ohne Treffer. Der Status ist `applied`, wenn mindestens ein String geändert wurde, sonst `none`.
6. **Grenzen.** Die Funktion liest höchstens `MaxBytes + 1` Bytes aus `src`; steht das Byte `MaxBytes + 1` im Strom, ist das `unsafe` mit Grund `zu_gross`, ein Verstoß hinter diesem Byte wird nie gesehen; ein Body von genau `MaxBytes` Bytes ist zulässig. Gezählt wird die Eingabe (die entpackte, `SPEC-053`), nicht die Ausgabe, die länger sein kann (`""` wird `"[REDACTED]"`); die Grenze der Ausgabe setzt der Schreiber (`SPEC-106` Nr. 1) und sein Fehler ist ein Fehler von `dst` (Nr. 7).
   Der Speicher hängt nicht von der Länge des Bodys, eines Strings, eines Namens oder einer Zahl ab: er wächst nur mit der Verschachtelung (höchstens 512), mit dem längsten Namen in `Pfade` (ein längerer Mitgliedsname kann nicht treffen und wird nicht weiter gehalten; ein Name von genau dieser Länge trifft, ein um ein Zeichen längerer nicht) und mit festen Puffern je Richtung, deren Größe nicht vom Strom abhängt (Konstanten im Code, heute 32 KiB Eingabe und 32 KiB Ausgabe samt kleiner Reserve); die Zusage ist die Unabhängigkeit von der Länge, nicht die Zahl, und der Test hält den Zuwachs unter 1 MiB bei 9 MiB Strom.
7. **Fehler.** Jeder Verstoß ist Status `unsafe` und ein Fehler `*UnsicherFehler` mit `Grund`; `errors.Is(err, ErrUnsicher)` ist wahr. Die Gründe sind `typ`, `mismatch`, `syntax`, `utf8`, `tiefe`, `nichtstring`, `zu_gross`, `unvollstaendig`, `eingabe` und `ausgabe`. Der erste Verstoß im Strom entscheidet (Reihenfolge der Bytes); nach ihm wird nichts mehr gelesen und kein weiteres Eingabebyte verarbeitet. Einzig der schon redigierte, noch gepufferte Teil der Ausgabe geht ein letztes Mal an `dst` (Nr. 9), außer der Verstoß ist `ausgabe`: dann bekommt `dst` nichts mehr, auch kein zweiter Versuch. Der Code ist `ODR-E2401` (`SPEC-017`), `Code()` des Fehlers liefert ihn. Der Text nennt Grund und nie ein Byte der Eingabe.
   - `eingabe`: `src` liefert einen anderen Fehler als `io.EOF`, auch nachdem der Body vollständig angekommen ist (Abbruch des Clients, Zeitüberschreitung, Herunterfahren; ein Body ohne Ende gilt nicht als vollständig). Daten, die ein `Read` zusammen mit dem Fehler liefert, werden vorher verarbeitet; `n > 0` zusammen mit `io.EOF` ist Daten und Ende. Die Ursache steht im Fehler (`errors.Is(err, context.Canceled)` bleibt wahr).
   - `ausgabe`: `dst` liefert einen Fehler oder schreibt weniger Bytes ohne Fehler (`io.ErrShortWrite`); die Ursache steht im Fehler (`ErrZuGross` bleibt wahr).
   - **Rangfolge der Ursachen:** entschieden ist, was der Ablauf zuerst trifft, und der Ablauf ist: ein `Read`, die Verarbeitung seiner Daten Byte für Byte (ein Verstoß, der Fehler eines Schreibens bei vollem Puffer), am Ende der Daten das Leeren der Ausgabe nach `dst`, erst danach die Auswertung eines mit den Daten gelieferten Lesefehlers. Daraus: ein Verstoß in den Daten schlägt einen Fehler von `src` aus demselben `Read` und einen Fehler von `dst` beim Leeren am Ende derselben Daten (der Verstoß ist dann der Grund, ein Fehler des anschließenden Leerens wird verworfen); ein Fehler von `dst` bei vollem Puffer vor dem Verstoßbyte schlägt den Verstoß; ein Fehler von `dst` beim Leeren schlägt einen Fehler von `src` aus demselben `Read` (`ausgabe`, `src` wird nicht weiter ausgewertet); `zu_gross` kommt nach der Verarbeitung der `MaxBytes` Bytes und vor dem Leeren. Weil der Zeitpunkt der Schreibvorgänge zum Ablauf gehört, gilt Nr. 8 für den Grund nur bei einem `dst` ohne Fehler.
   - Leere Reads (`0, nil`) sind erlaubt und ändern nichts; ein `src`, der nie Daten liefert, endet durch seinen Fehler (die Frist trägt der Aufrufer, `SPEC-053`). Ein `src`, der endlos `0, nil` liefert, ohne je zu lesen, bricht den Vertrag von `io.Reader`; die Funktion zählt solche Reads nicht und bricht sie nicht ab.
8. **Strom und Wiederholung.** Ausgabe, Status und Grund hängen nicht davon ab, wie `src` den Strom in Reads zerlegt (jede Zweiteilung, einzelne Bytes, leere Reads dazwischen); gleicher Strom, gleiches Ergebnis. Die Funktion auf ihre eigene Ausgabe angewendet liefert dieselben Bytes und `none`.
9. **Ausgabe bei Fehlern.** Ins `dst` gelangt nie ein Byte eines String-Wertes unter einem treffenden Namen, auch nicht bei `unsafe`: `dst` trägt bei jedem Fehler außer `ausgabe` genau die Ausgabe für die Eingabebytes vor dem Verstoßbyte (bei `eingabe` für alle gelieferten, bei `zu_gross` für die ersten `MaxBytes`), das Verstoßbyte und alles danach nie, bei `typ` und bei einem Fehler von `src` vor dem ersten Byte nichts (bei `mismatch` stehen davor liegender Leerraum und BOM). Das ist ein redigiertes, unvollständiges Präfix und kein Body: der Marker eines ersetzten Strings erscheint erst mit dessen schließendem Anführungszeichen (ein offener ersetzter String trägt nichts), ein Literal oder eine Mehrbyte-Folge vor dem Verstoß kann halb stehen. Bei `ausgabe` steht in `dst`, was `dst` bis zum fehlgeschlagenen Write angenommen hat, mehr nie. Der Aufrufer behandelt jeden Fehler als Verwerfen von `dst` (`Abort` des Schreibers): ein Blob entsteht nur aus einem Aufruf ohne Fehler (`SPEC-046` Nr. 4, `SPEC-048`). Bei `unsafe` ist der Body der Richtung `discarded`, die Redaction der Richtung `unsafe` und `replay_equivalence=none` (`SPEC-001`); `capture_status` bestimmt der Aufrufer nach Art des Fehlers (Abbruch gegen Verstoß, `SPEC-053`).
10. **Aufrufer.** Der Strom, den der Redaktor liest, ist der entpackte: `Content-Encoding` löst der Aufrufer auf und prüft es vorher (`SPEC-046` Nr. 2), der Redaktor sieht es nicht. Jeder Body, der in einen Blob oder in einen Match-Key (`SPEC-116` Einleitung) eingeht, Request wie Response und im Replay der eingehende Request, läuft durch diese Funktion; der Key wird aus der Ausgabe gebildet, nie aus den Eingabebytes. Im Replay ist `MaxBytes` ebenfalls `MAX_CAPTURE_BODY`, und ein Request-Body mit Fehler lässt sich nicht zu einem Key redigieren: der Dienst antwortet `ODR-E5001` ohne Cursor-Fortschritt (wie ein `null`-Key, der nie passt, `SPEC-001`). Im Record führt die Capture-Pipeline den Redaktor und stellt dem Dienst die redigierten Request-Bytes bereit (`SPEC-119` Nr. 13), im Replay ruft der Dienst die Funktion selbst.

Ungeprüft bleibt: ein Typed-nil als `dst` oder `src` (`(*bytes.Buffer)(nil)` besteht die `nil`-Prüfung von Nr. 1 und löst bei der ersten Nutzung einen Panic aus; wie bei `io.Copy` ein Programmierfehler des Aufrufers); ein `src`, der endlos `0, nil` liefert (Nr. 7); Geheimnisse unter Namen, die keine Liste nennt, und in Strings eines Arrays unter einem nicht treffenden Namen; Unicode-Gleichheit von Namen über die ASCII-Schreibweise hinaus (`api_Key` mit dem Kelvin-Zeichen trifft `$..api_key` nicht, `SPEC-115`); Geheimnisse im Inhalt von EDMX (er enthält nach `SPEC-046` Nr. 3 keine); der Wahrheitswert und `null` unter einem treffenden Namen; das Ausbleiben von Datenrennen zwischen gleichzeitigen Aufrufen (die Funktion hält keinen gemeinsamen Zustand; das Gate fährt nicht mit dem Race-Detektor); die Laufzeit (ein Durchlauf, O(n) über höchstens `MaxBytes`).

### SPEC-119 — Capture-Pipeline: Vertrag des Pakets `capture`

Das Paket `capture` des Record-Service (`ARC-002`) entkoppelt den Proxy-Strom von der Aufzeichnung, entpackt begrenzt, ruft den Redaktor (`SPEC-118`) und den `CaptureWriter` (`SPEC-106`) und meldet je Body ein `Ergebnis` und je Austausch den Zustand (`Abschluss`). Es setzt `SPEC-046`, `SPEC-053` und `LH-FA-02.b` um und importiert nur Domain und Outbound Ports. Es kennt weder Repository noch Upstream noch Session: Zeit, Kennungen, Header, Match-Key und das Anhängen des Exchange gehören dem Dienst.
Das Lastenheft nennt einen nicht vollständig erfassten Austausch „unvollständig“ (`LH-FA-02.b`, `LH-FA-14`); als Wert bezeichnet `incomplete` hier den Abbruch des Austauschs, `failed` die Verwerfung durch die Capture-Seite (Nr. 10). Beide sind nicht vollständig erfasst und nach `SPEC-045` nie replayfähig; der Unterschied ist nur, ob die Ursache der Austausch (Client, Upstream, Herunterfahren) oder die Aufzeichnung war.

```text
Konfig{Blobs outbound.BlobRepository; Pfade []model.JSONPfad; MaxBytes, QueueBytes int64; Frist time.Duration; MaxVerhaeltnis int64}
Standard(blobs, pfade, maxBytes) Konfig             New(Konfig) (*Pipeline, error)
(*Pipeline).Tee(ctx, src io.Reader, Eingang) (*Aufnahme, error)
Eingang{Richtung (Request|Response); ContentType, ContentEncoding []string; Metadata bool}
(*Aufnahme).Read(p) (int, error)   .Abbrechen()   .Ergebnis() Ergebnis
Abschluss(request, response Ergebnis) (Zustand, error)
```

Der Proxy liest den Body aus der `Aufnahme` statt aus `src` und leitet ihn wie bisher weiter; `Metadata` setzt der Aufrufer (`SPEC-118`). `Read` ruft nur eine Goroutine auf; `Abbrechen` und `Ergebnis` darf jede Goroutine aufrufen, auch während `Read` in `src` wartet.

1. **Konfiguration.** `Standard` setzt `QueueBytes` auf 16 MiB, `Frist` auf 30 s und `MaxVerhaeltnis` auf 20 (`SPEC-053`); `MaxBytes` ist `MAX_CAPTURE_BODY` (`SPEC-013`) und gilt für Request und Response. `New` liefert `model.ErrUngueltig` und keine Pipeline bei: `Blobs` nil; `MaxBytes`, `QueueBytes` oder `MaxVerhaeltnis` kleiner als `1`; `Frist` nicht größer als `0`; einem Nullwert von `JSONPfad` in `Pfade`. Eine leere oder `nil`-Menge `Pfade` ist gültig (wie `SPEC-118` Nr. 1). Eine Obergrenze (1 GiB für `MaxBytes`) prüft die Konfiguration (`SPEC-013`), nicht `New`; `QueueBytes` steht in keinem Verhältnis zu `MaxBytes`.
2. **Eröffnen.** `Tee` liefert `model.ErrUngueltig`, ohne ein Byte aus `src` zu lesen und ohne Goroutine, bei: `ctx` nil, `src` nil, `Richtung` weder `Request` noch `Response` (auch der Nullwert), `Metadata` bei `Request`. Sonst startet genau eine Goroutine der Aufnahme, und die Frist (Nr. 9) beginnt; ist `ctx` schon beendet, wird die Aufnahme dennoch eröffnet und ist sofort `abbruch` (Nr. 9). `ContentType` und `ContentEncoding` sind alle Werte des jeweiligen Headers; leer oder `nil` heißt: fehlt.
3. **Durchreichen.** `Read(p)` ruft `src.Read(p)` genau einmal und gibt dessen `(n, err)` zurück: dieselben Bytes in `p`, derselbe Fehlerwert (nicht verpackt; auch `io.EOF`, auch `n > 0` zugleich mit einem Fehler). Das gilt in jedem Zustand der Aufnahme (laufend, beendet, verworfen, übergelaufen, abgebrochen), und `Read` wartet nie auf die Aufnahme: ein langsamer oder blockierter Schreiber verzögert den Proxy nicht. `Read` behält `p` nicht über den Aufruf hinaus, es kopiert. `src` schließt die Pipeline nie: den Strom besitzt und schließt der HTTP-Adapter (`LH-FA-02.b`).
4. **Queue.** Nach jedem `Read` mit `n > 0` legt der Tee, solange die Aufnahme läuft, eine Kopie der `n` Bytes in eine Queue, aus der die Goroutine der Aufnahme liest; das Einlegen blockiert nie. `ausstehend` ist die Zahl der eingelegten und noch nicht entnommenen Rohbytes. Wäre `ausstehend + n` größer als `QueueBytes`, ist das ein Überlauf (Nr. 9): genau `QueueBytes` ist zulässig, ein Byte mehr nicht, und ein einzelner `Read` über `QueueBytes` ist ein Überlauf. Die Aufnahme hält nie mehr als `QueueBytes` Rohbytes in der Queue, dazu das eine Stück, das die Goroutine gerade bearbeitet, und die festen Puffer von Dekompressor und Redaktor; wie viel Speicher darüber hinaus der Body braucht, hängt nicht von seiner Länge ab. Das erste `io.EOF` und der erste Fehler von `src` beenden die Queue; spätere Reads reichen nur durch. Ein Read mit `n = 0` legt nichts ein. Ist das Ergebnis der Aufnahme bestimmt, ist die Queue leer und `Read` kopiert nichts mehr. Ausgang, Grund, Blob und Redaction hängen nicht davon ab, wie `src` den Strom in Reads zerlegt (jede Zweiteilung, einzelne Bytes, leere Reads; `SPEC-118` Nr. 8); allein der Überlauf hängt an der Größe der Reads und am Tempo der Goroutine. Rohbytes liegen nur in der Queue und in diesen Puffern, nie in einer Datei (`SPEC-046` Nr. 1).
5. **Rohgröße.** Der Tee zählt die Rohbytes aller Reads. Sobald mehr als `MaxBytes` geliefert werden, legt er nur noch bis `MaxBytes` ein (genau `MaxBytes` Rohbytes sind zulässig) und beendet die Queue mit der Ursache Rohgröße (Grund `zu_gross`, Nr. 10), nach den Bytes davor. Das begrenzt auch die Arbeit des Dekompressors, wenn der entpackte Strom klein bleibt (viele leere gzip-Glieder).
6. **Content-Encoding.** Die Werte werden an Kommas geteilt, Leerzeichen und Tabulatoren am Rand jedes Teils entfernt; die Kette ist die Folge aller Teile über alle Werte in der Reihenfolge der Header-Zeilen, der Name gilt ohne Beachtung der ASCII-Schreibweise. Bekannt sind `gzip`, `deflate` (das zlib-Format, RFC 1950; ein Strom ohne zlib-Kopf ist `dekodierung`) und `identity`. Jede andere Kette ist `encoding`, schon in `Tee` entschieden und ohne ein Byte zu lesen: ein unbekannter Name (darunter `br`, `zstd`, `compress`, `x-gzip`), ein leerer Teil (`gzip,`, `,gzip`, ein leerer Wert, nur Leerraum) und mehr als zwei Glieder (`identity` zählt). Ein fehlender Header ist die leere Kette: nichts wird entpackt. Auch ein Body ohne Rohbyte trägt bei unbekannter Kette den Grund `encoding`; Antworten, die nach HTTP keinen Body haben (Antwort auf `HEAD`, 1xx, 204, 304), gibt der Dienst nicht in die Pipeline. Entpackt wird als Strom in umgekehrter Reihenfolge der Kette, ohne den entpackten Body zu sammeln. Mehrere gzip-Glieder hintereinander in einer Stufe sind ein Strom. `br` und `zstd` hat die Standardbibliothek nicht, ein Drittpaket bräuchte eine ADR: ein solcher Body ist `failed` und nicht replayfähig (`SPEC-046` Nr. 2); ob der Dienst `Accept-Encoding` einschränkt, entscheidet der Record-Service, nicht die Pipeline.
7. **Nach dem Lesen.** Meldet der Redaktor keinen Fehler, prüft die Goroutine in dieser Reihenfolge: (a) die Queue ist erschöpft; ein Byte nach dem Ende des entpackten Stroms (Nachlauf, auch hinter einem zlib-Strom) ist `dekodierung`; (b) es gilt `entpackt ≤ MaxVerhaeltnis × roh`, sonst `bombe` (ohne Entpacken ist das Verhältnis 1 und die Prüfung trivial); `entpackt` sind die Bytes am Eingang des Redaktors, `roh` alle Rohbytes; genau `MaxVerhaeltnis : 1` ist zulässig, und die Rechnung läuft nicht über. Das Verhältnis wird erst am Ende geprüft, damit der Grund nicht davon abhängt, wie die Rohbytes eintreffen; davor begrenzen `MaxBytes` (entpackt, `SPEC-053`) und die Frist die Arbeit. Hat der Redaktor einen Fehler gemeldet, entscheidet dieser (Nr. 9).
8. **Redaktor und Writer.** Die Goroutine der Aufnahme ist die einzige, die Dekompressor, Redaktor und `CaptureWriter` benutzt (ein Writer gehört einer Goroutine). Sie ruft `RedigiereBody(dst, entpackt, BodyEingang{ContentType, Metadata}, BodyRegeln{Pfade, MaxBytes})`; `dst` eröffnet den Writer mit `Begin(ctx der Aufnahme, MaxBytes)` erst beim ersten Byte, so dass ein Body ohne Ausgabe (leer, `typ`) nie eine temporäre Datei anlegt. Bei jedem Fehler des Redaktors, des Writers oder einer Prüfung der Nr. 7 und bei jeder Ursache der Nr. 9 vor dem `Commit` ruft sie `Abort` genau einmal, falls ein Writer offen ist, und nie `Commit`; offen ist ein Writer, den `Begin` lieferte und dessen `Commit` nicht gelang (scheitert `Commit`, folgt `Abort`, damit die temporäre Datei geht, `SPEC-106` Nr. 1); nach gelungenem `Commit` ruft sie nichts mehr auf. Ein Fehler von `Abort` ändert das Ergebnis nicht. Sonst ruft sie `Commit`; die Länge, die er liefert, muss der Zahl der an den Writer geschriebenen Bytes gleichen, sonst Grund `ablage` (der Blob bleibt dann als verwaister, redigierter Blob, `SPEC-048`; es folgt kein `Abort`). Ein entpackter Strom ohne ein Byte (leere Rohbytes, ein leeres gzip-Glied) ist Body `absent`, ohne Writer. Ein gelungener `Commit` wird nie zurückgenommen: eine Ursache der Nr. 9 danach ändert das Ergebnis nicht, und hängt der Dienst den Exchange nicht an, bleibt ein verwaister redigierter Blob (`SPEC-048`, zulässig, die Bereinigung entfernt ihn). Der Writer bekommt nur Bytes des Redaktors, bei einem Fehler das redigierte Präfix (`SPEC-118` Nr. 9), nie ein Rohbyte.
9. **Ursachen und ihre Rangfolge.** Zwei Arten. *Im Strom*: ein Verstoß in den Bytes, Rohgröße, Fehler der Dekompression, Fehler von `src`; sie liegen in der Reihenfolge der Bytes und werden in ihr gesehen: Daten vor einem Fehler von `src` werden verarbeitet, und ein Verstoß in ihnen geht dem Fehler vor (`SPEC-118` Nr. 7). *Außerhalb des Stroms*: Überlauf (Nr. 4), Ablauf der Frist, `Abbrechen`, Abbruch von `ctx`; sie wirken sofort, ohne Rücksicht auf die Bytes, die noch in der Queue stehen (sie werden verworfen), und die erste in der Zeit gewinnt, spätere ändern nichts. Eine solche Ursache vor `Commit` verhindert ihn. Die Frist läuft von `Tee` bis zum Ergebnis auf der Systemuhr (nicht `Clock`, `SPEC-114`) und umfasst `Begin`, `Write` und `Commit`; der Writer ist an den Kontext der Aufnahme gebunden (`SPEC-106` Nr. 1), den Frist, `Abbrechen` und Abbruch von `ctx` beenden; ein Fehler des Writers, der einer des Kontexts ist (`context.Canceled`, `context.DeadlineExceeded`), zählt als die Ursache, die den Kontext beendete. Eine Deadline, die `ctx` des Aufrufers trägt, ist ein Abbruch von `ctx` (`abbruch`), nicht die Frist der Pipeline (`frist`). Daraus bei gleichzeitigen Ursachen:
   - `encoding` (Nr. 6) geht jeder anderen vor, auch einem Abbruch von `ctx` vor `Tee`.
   - Verstoß in den Bytes vor einem späteren Fehler von `src`: der Verstoß. Fehler von `src` mitten in einem gzip-Strom: `abbruch`, nicht `dekodierung` (der Fehler von `src` ist die Ursache, der Dekompressor folgt ihm); endet `src` mit `io.EOF` mitten im Strom, ist es `dekodierung`.
   - Überlauf gegen einen Verstoß in den Bytes, die noch in der Queue stehen: der Überlauf.
   - Frist, `Abbrechen` und Abbruch von `ctx` untereinander: die zeitlich erste.
   - Ein Fehler des Redaktors, des Writers oder der Nr. 7, der aus den Bytes oder dem Writer selbst folgt, bleibt, was er ist, auch wenn danach eine Ursache außerhalb eintritt.
10. **Grund und Ausgang.** Jede Ursache hat genau einen Grund und einen Ausgang:

    | Ursache | Grund | Ausgang |
    |---|---|---|
    | Redaktor `unsafe`: `typ`, `mismatch`, `syntax`, `utf8`, `tiefe`, `nichtstring`, `unvollstaendig` | derselbe | `Verstoss` |
    | Redaktor `zu_gross` (entpackt) und Rohgröße (Nr. 5) | `zu_gross` | `Verstoss` |
    | Kette (Nr. 6) | `encoding` | `Verstoss` |
    | Fehler des Dekompressors (Kopf, Prüfsumme, ungültiger Strom, `io.EOF` mitten im Strom), Nachlauf (Nr. 7a) | `dekodierung` | `Verstoss` |
    | Verhältnis (Nr. 7b) | `bombe` | `Verstoss` |
    | Überlauf (Nr. 4) | `ueberlauf` | `Verstoss` |
    | Frist | `frist` | `Verstoss` |
    | Redaktor `ausgabe` mit `outbound.ErrZuGross` (Ausgabe über der Grenze des Schreibers, `SPEC-118` Nr. 6, 7) | `ausgabe_zu_gross` | `Verstoss` |
    | Fehler von `Begin`, `Write`, `Commit` sonst, falsche Länge (Nr. 8) | `ablage` | `Verstoss` |
    | Fehler von `src` ungleich `io.EOF`, `Abbrechen`, Abbruch von `ctx` | `abbruch` | `Abbruch` |

    `Verstoss` ist eine Verwerfung durch die Capture-Seite (`SPEC-053`: Überlauf, unbekanntes Encoding, fehlerhaftes JSON, Zeitüberschreitung und Speicherfehler sind `failed`, `LH-FA-02.a` Nr. 5), `Abbruch` das vorzeitige Ende des Austauschs (`LH-FA-02.b`, `LH-FA-02.c`).
11. **Ergebnis.** `Ergebnis{Richtung, Ausgang, Grund, Body, Blob, Redaction, Integrity, Fehler, Redigiert}`; `Ergebnis()` wartet auf die Goroutine und liefert bei jedem Aufruf denselben Wert.
    - `Vollstaendig` (Grund leer, `Fehler` nil): Body `absent` (Blob leer, Redaction `none`, Integrity `absent`) oder `present` (Blob die Referenz des `Commit`, Redaction `none` oder `applied` wie vom Redaktor, Integrity `verified`).
    - `Abbruch` und `Verstoss`: Body `discarded`, Blob leer, Redaction `unsafe`, Integrity `absent` (`SPEC-118` Nr. 9, `SPEC-001`), auch wenn kein Byte angekommen war.
    - `Fehler` ist ein `*Fehler` mit Grund, `Code()` liefert `ODR-E2401`, der Text nennt den Grund und nie ein Byte des Bodys, eines Headers oder eines Pfades; `errors.Is` erreicht die Ursache (`model.ErrUnsicher`, `outbound.ErrZuGross`, den Fehler von `src`, `context.Canceled`). Die Pipeline protokolliert nichts: der Dienst protokolliert den Fehler ohne Payload (`SPEC-046` Nr. 4).
    - `Integrity=invalid` liefert die Pipeline nie.
12. **Zustand des Austauschs.** `Abschluss(request, response)` ist eine reine Funktion; `model.ErrUngueltig`, wenn `request.Richtung` nicht `Request` oder `response.Richtung` nicht `Response` ist (auch der Nullwert), sonst:
    - Body, Blob und Redaction je Richtung aus dem Ergebnis.
    - `Capture`: `failed`, wenn der Ausgang des Requests nicht `Vollstaendig` ist (auch bei `Abbruch`: ohne vollständigen redigierten Request-Body gibt es keinen Key, und `redaction.request=unsafe` verlangt `failed`, `SPEC-001`) oder die Response `Verstoss` ist; `incomplete`, wenn der Request `Vollstaendig` und die Response `Abbruch` ist; sonst `complete`.
    - `KeyMoeglich`: der Request ist `Vollstaendig`.
    - `Integrity`: `verified`, wenn mindestens ein Body `present` ist, sonst `absent`.
    - `Equivalence`: `none`, wenn ein Body `discarded` ist; sonst `semantic_or_limited`, wenn die Redaction der Response `applied` ist; sonst `byte`. Die Redaction des Request-Bodys ändert sie nicht (`SPEC-045`).
    Ein Ergebnis fremder Herkunft prüft `Abschluss` nur auf die Richtung; ob der Exchange gültig ist, entscheidet `Exchange.Validate`. Jeder Zustand aus Ergebnissen der Pipeline, in einen Exchange mit gültigem Kopf eingesetzt (MatchKey genau dann nichtleer, wenn `KeyMoeglich`), besteht `Exchange.Validate`.
13. **Request-Body und Match-Key.** Bei Richtung `Request` und Body `present` trägt `Ergebnis.Redigiert` die Bytes, die der Redaktor erzeugte und die in den Blob gingen; sie, nie die Rohbytes, sind das Argument `Body` von `BildeSchluessel` (`SPEC-116`, `SPEC-118` Nr. 10). Der Puffer nimmt nur Bytes auf, die der Writer angenommen hat, ist also höchstens `MaxBytes` groß, und liegt nur im Speicher, bis der Dienst das Ergebnis freigibt; in allen anderen Fällen (`absent`, `discarded`, Richtung `Response`) ist `Redigiert` `nil`. Ist `KeyMoeglich` falsch, bildet der Dienst keinen Key (`SPEC-001`).
14. **Nebenläufigkeit und Ende.** Jede Aufnahme hat genau eine Goroutine, und sie endet in jedem Fall: mit dem Ende der Queue, einem Fehler oder einer Ursache der Nr. 9; die Frist beendet auch ein `src`, das nie liefert und nie endet (`SPEC-118` Nr. 7). Wenn `Ergebnis()` zurückkehrt, ist kein Writer offen und die Goroutine rührt Writer, Redaktor und Dekompressor nicht mehr an. `Abbrechen` ist nach dem Ergebnis wirkungslos, darf mehrfach und gleichzeitig mit `Read` und `Ergebnis` aufgerufen werden, wartet weder auf `src` noch auf den Writer und lässt `Read` unberührt. Ein Austausch, dessen Request-Body der Upstream nicht zu Ende liest, wird vom Dienst mit `Abbrechen` beendet und ist dann `Abbruch` (Nr. 12: `failed`). Aufnahmen teilen keinen Zustand.
15. **Fault Injection, Capture-Seite** (`SPEC-048`). Die Punkte sind die Aufrufe des Writers: `Begin` (vor dem Blob-Write), `Write` mitten im Body, `Commit`. Scheitert einer, gibt es kein `Commit` (nach `Begin` und `Write`) bzw. kein `complete`, `Abort` läuft, der Grund ist `ablage` (oder `ausgabe_zu_gross`, Nr. 10), und im Writer steht nie ein Rohbyte. Der Punkt zwischen gelungenem `Commit` und dem Anhängen liegt beim Dienst (Nr. 8).

Ungeprüft bleibt: Datenrennen zwischen `Read`, `Abbrechen` und der Goroutine (das Gate fährt kein `-race`; die Tests bestimmen die Reihenfolge mit Kanälen, was ein Rennen nicht ausschließt); ein Panic eines Ports oder einer Funktion der Pipeline (Programmfehler, wie bei `SPEC-118`: nicht abgefangen, damit er nicht verborgen wird); ein Typed-nil als `Blobs` oder `src`; der Speicher über alle gleichzeitigen Aufnahmen (die Queue gilt je Aufnahme; eine Grenze für die Zahl gleichzeitiger Aufnahmen gibt diese Version nicht vor); ein Aufruf des Writers, der nie zurückkehrt (die Pipeline kann ihn nicht unterbrechen, `Ergebnis()` wartet dann mit); die Werte 16 MiB, 20 und 30 s außer in `Standard` (die Tests fahren mit kleinen Werten).

### SPEC-043 — Import- und Provenienz-Schema

`ImportManifest{import_id, service_id, metadata_hash, source_session_ids[], created_at, entity_count, baseline_snapshot_id, coverage[]}`;
`ImportedEntity{entity_set, typed_key, payload, selected_properties[], provenance[]}`.
Snapshot-Blobs enthalten eine versionierte kanonische JSON-Darstellung und ihren SHA-256-Hash.
`baseline_snapshot_id` wird vor Workspace-Erzeugung auf Existenz (`ODR-E7002`) und `snapshot_kind=baseline` (`ODR-E7008`) geprüft;
SQLite-Fremdschlüssel und Migrationstests sichern referenzielle Integrität.
Garbage Collection löscht nur Blobs, die weder von Exchange noch Snapshot referenziert werden; Crash nach Blob-Rename vor DB-Commit darf höchstens verwaiste Blobs erzeugen.

### SPEC-044 — Baseline-Coverage und Entity-Vollständigkeit

Jeder Mock-Read löst zunächst `workspace.baseline_snapshot_id` auf und liest Coverage ausschließlich für `(baseline_snapshot_id, entity_set)`.
Lokale Writes ändern diese Baseline-Evidenz nicht.
Ein neuer Snapshot kann eine neue Baseline erzeugen, aber `complete` wird nicht allein aus lokalem Schreiben abgeleitet.
Bei `STRICT_COVERAGE=true` darf eine nicht nachweislich vollständige Collection keine scheinbar vollständige Ergebnismenge liefern.

Für Entity-Reads gilt: Projektionen ausschließlich bekannter Properties sind zulässig, wenn Identität und Key verifiziert sind; vollständiges `GET` auf `partial` ergibt `ODR-E6003`.
`PATCH` auf einer partiellen Entity ist im MVP **abzuweisen** (`ODR-E6003`, HTTP 422);
`PUT` darf nur als explizite vollständige Ersetzung mit validierten Pflichtfeldern und atomarem State-Update ausgeführt werden, ansonsten `ODR-E6003`.
Keine unbekannte Property wird durch `null` ersetzt.
Tests: `$select`, Null-vs-missing, Pflichtfeld, PATCH/PUT, zwei Baselines eines Services.

### SPEC-045 — Capture-/Replay-Fähigkeitsmatrix

| Capture | Response-Body-Redaction | Body-Integrität | Default Replay |
|---|---|---|---|
| `complete` | `none` | `verified` oder `absent` | Byte-äquivalent zulässig |
| `complete` | `applied` | `verified` oder `absent` | Nur explizit aktiviertes `semantic_or_limited` (`REPLAY_REDACTED_BODIES=true`) |
| `complete` | `unsafe` | beliebig | Verboten |
| `incomplete`/`failed` | beliebig | beliebig | Verboten |
| beliebig | beliebig | `invalid` | Verboten |

Die Spalte betrifft ausschließlich die **Redaction des Response-Bodys**, weil nur dieser ausgeliefert wird.
**Verbotener Treffer:** Ein Exchange, den diese Matrix mit „Verboten" belegt, ist kein Kandidat: Er wird weder ausgeliefert noch im Cursor verbraucht.
Findet das Matching nur solche Exchanges, antwortet der Dienst mit `ODR-E5008` (HTTP 422, kein Cursor-Fortschritt); findet es keinen Exchange mit dem Key, gilt `ODR-E5001`.
Im Modus `sequential` wird der nächste replayfähige Kandidat gewählt; im Modus `strict` ist ein nicht replayfähiger nächster Exchange (`sequence_no = position + 1`) ein `ODR-E5008`.
Eine Redaction des Request-Bodys (`redaction.request=applied`), von Query-Parametern oder von Headern schränkt die Replayfähigkeit nicht ein; sie wirkt über den Match-Key (`LH-FA-14.a`).
Die Redaction von Headern (`Authorization`, `Cookie`, `Set-Cookie` usw., `LH-FA-14.a`) schränkt die Replayfähigkeit nicht ein: Status und Body werden byteidentisch ausgeliefert, redigierte Header werden weggelassen und nie mit dem Marker `[REDACTED]` ausgeliefert.
Default ist **strict byte replay** für Exchanges ohne Body-Redaction; Exchanges mit Body-Redaction sind nicht stillschweigend replayfähig.
Bei aktivierter eingeschränkter Policy sind Antwortfelder mit Redaction als eingeschränkt gekennzeichnet; Auth-Credentials können nicht rekonstruiert werden.

### SPEC-046 — Sicheres Streaming-Capture

1. Proxy-Transport und Capture-Tee sind entkoppelt.
   Rohbytes dürfen nur in flüchtigem RAM bis zur vorgegebenen Grenze liegen, niemals in Klartext in temporären Dateien, Swap-bewussten Spools oder Blob-Dateien.
   Für nicht sicher streambar redigierbare JSON-Payloads wird die Capture-Seite vor Persistenz verworfen; der Proxy läuft weiter.
2. Komprimierte Payloads werden **nur** mit begrenzter Dekompression und vollständig validierter Encoding-Kette analysiert; sonst `capture_status=failed`, `replay_equivalence=none`.
   Keine ungeprüften gzip-Rohbytes persistieren, wenn Body-Redaction konfiguriert ist.
3. **Bekannte Typen:** `application/json` (auch `+json`, mit OData-Parametern) wird nach `LH-FA-14.a` redigiert;
   `application/xml` ist nur für die Antwort auf `/$metadata` bekannt (EDMX/CSDL enthält keine Geheimnisse, `redaction=none`, kein JSON-Pfad-Rewrite).
   Jeder andere Typ, auch `multipart/mixed` (`$batch`) und `text/*`, ist unbekannt.
   Unbekannte MIME-Typen, fragmentierte JSON-Tokens, unvollständige JSON-Dokumente und Bodies über dem konfigurierten Grenzwert sind **fail-closed für Capture**.
   Header-Redaction erfolgt vor jedem dauerhaften Metadaten-Write.
4. Die Streaming-Redaction muss Token über Chunk-Grenzen korrekt erkennen.
   Bei Absturz dürfen nur bereits redigierte Blobs existieren.
   Fehler werden mit `ODR-E2401` protokolliert, ohne Payload im Log.
5. Ohne sichere Speicher-/Spool-Garantie ist für große oder unbekannte Bodies die **Body-Aufzeichnung deaktiviert**, nicht bloß auf ein unsicheres Tempfile ausgelagert.
6. Der Vertrag des Redaktors (Typ, Grenzen, Wertarten, Fehlergründe, Verhalten bei Chunk-Grenzen und Abbruch) steht in `SPEC-118`.
7. Der Vertrag der Capture-Pipeline (Tee, Queue, Dekompression, Zustände nach Fehlerart, Abbruch und Herunterfahren) steht in `SPEC-119`.

### SPEC-052 — Replay-HTTP-Identität und Lebenszyklus

Pflichtheader bei sequenziellem Replay: `X-Recorder-Workspace` (bestehende Workspace-ID), `X-Recorder-Recording-Session` (aufgezeichnete Session-ID), `X-Recorder-Replay-Session` (vom Test-Harness vergebene, innerhalb des Workspace stabile Client-Session-ID).
Kein Header ist ein Upstream-Header; sie werden vor Weiterleitung entfernt.
Der Client generiert eine kollisionsresistente Replay-Session-ID (UUID empfohlen) und behält sie über die Request-Folge.
Zwei Clients dürfen dieselbe Recording-Session mit unterschiedlichen Replay-Session-IDs unabhängig abspielen.
Fehlender, leerer oder mehrdeutiger Header: `ODR-E5003`, HTTP 400, ohne Cursoränderung.
Unbekannte Workspace-ID: `ODR-E7001`, unbekannte Recording-Session: `ODR-E7003` (je HTTP 404); Recording-Session im Status `open`: `ODR-E7009` (HTTP 409).
Servicefremde IDs: `ODR-E5004`, HTTP 409, ohne Cursoränderung.
Eine neue Replay-Session startet an Position 0; ein Reset des Workspace setzt alle Cursor atomar auf Position 0 zurück; ein Snapshot-Restore lässt die Cursor unverändert (`LH-FA-10.a`).
Keine implizite Ableitung aus Cookies oder Client-IP.

### SPEC-053 — Capture-Bedrohungsmodell und Testvertrag

Geltungsbereich: vom Recorder verwaltete RAM-Puffer, Capture-Queues, Blob-Verzeichnis, temporäre Dateien und Logs.
Unredigierte Daten dürfen **nicht durch Recorder-Code** persistent geschrieben werden.
Eine Aussage über OS-Swap, Core-Dumps, Container-Runtime oder externe HTTP-Libraries gilt nur unter expliziten Betriebsbedingungen: Core-Dumps deaktiviert, Swap deaktiviert oder verschlüsselt und vom Betreiber kontrolliert, keine unverschlüsselten Diagnose-Dumps, kein Body-Logging in Middleware.
Das Betriebsprofil muss diese Voraussetzungen vor Aktivierung von Body-Capture prüfen und dokumentieren.

Bounded memory: maximale Capture-Bodygröße `MAX_CAPTURE_BODY` (`SPEC-013`), maximale entpackte Größe (gleich `MAX_CAPTURE_BODY`), maximale Queue-Größe (16 MiB), Dekompressions-Verhältnis höchstens 20:1 und Deadline 30 s.
Überlauf, unbekanntes Encoding, fehlerhaftes JSON, ein Fehler der Ablage und die Überschreitung der Frist der Capture-Seite setzen `capture_status=failed`, `replay_equivalence=none`; der Abbruch des Austauschs (Client oder Upstream beenden den Strom, Herunterfahren) setzt `incomplete` (`SPEC-119` Nr. 10);
unredigierte Capture-Buffer werden verworfen, während Proxy-Streaming soweit möglich fortgesetzt wird.
Nur vollständig redigierte und validierte Bodies dürfen Blob-Dateien erreichen.
Backpressure darf nicht zu unbeschränkter Pufferung führen.
Sicherheits-Fixtures: JSON-Tokens über Chunk-Grenzen, gzip-Bombe, Content-Encoding-Kette, MIME-Mismatch, Body-Abbruch, langsamer Consumer, Queue-Überlauf, Crash zwischen Blob-Rename und DB-Commit, Redaction-Fehler und Middleware-Logging.

### SPEC-047 — Import-API-Vertrag

`POST /__admin/imports` benötigt Admin-Authentisierung (`LH-FA-13.a`) und JSON:

```json
{"import_id":"stable-client-id","service_id":"catalog","source_session_ids":["session-01"],"metadata_hash":"sha256:...","conflict_policy":"reject","target":"new_baseline"}
```

`import_id` ist idempotent: gleiche ID + identischer kanonischer Request-Hash liefert denselben Report; gleiche ID + anderer Inhalt liefert HTTP 409.
Nur `target=new_baseline` und `conflict_policy=reject` sind im MVP zulässig.
Erfolgreich: HTTP 201 mit `import_id`, `baseline_snapshot_id`, `coverage[]`, `entity_count`, `report`; Wiederholung: HTTP 200.
Statuscodes: fehlerhaftes JSON oder fehlende Pflichtfelder `400`;
unbekannte Session `404` (`ODR-E7003`);
Session im Status `open` `409` (`ODR-E7009`);
ungültige Metadaten `422` (`ODR-E6004`);
ungültiger oder widersprüchlicher Dateninhalt, auch ein Key-Konflikt zwischen Payloads, `422` (`ODR-E6101`);
gleiche `import_id` mit anderem Inhalt `409` (`ODR-E6102`);
fehlende oder ungültige Authentisierung `401` (`ODR-E2101`);
Storagefehler `500` mit Rollback.
Idempotent sind nur erfolgreiche Imports (nur sie hinterlassen eine Zeile in `imports`); nach einem fehlgeschlagenen Import darf dieselbe `import_id` erneut verwendet werden.
Starten zwei Requests gleichzeitig mit derselben `import_id`, gewinnt einer über den Primärschlüssel; der andere erhält bei gleichem kanonischem Request-Hash denselben Report, sonst `ODR-E6102`.
Import und Baseline-Sichtbarkeit werden atomar committed.

### SPEC-048 — Crash-Recovery und Fault Injection

Fault-Injection-Punkte: vor redigiertem Blob-Write, nach fsync, nach Rename/vor SQLite-Commit, während SQLite-Commit, nach Commit/vor HTTP-Antwort.
Erwartung: kein unredigierter persistenter Body; keine referenzierte fehlende Datei nach erfolgreichem Commit; ein verwaister **redigierter** Blob nach Rename ist zulässig und per GC entfernbar.
Startup prüft DB-Integrität, FK-Check und referenzierte Blob-Hashes.
GC löscht nur nach erneuter Referenzprüfung; Snapshot- und Exchange-Blobs werden berücksichtigt.

### SPEC-049 — MVP-OData-V4-Abnahmematrix

| Bereich | Muss (Golden-Fixture) | Soll | Nicht in MVP |
|---|---|---|---|
| GET | Collection/Key/leer (`ODV4-GET-01..03`) | ETag | Navigation-Expand |
| POST/PATCH/PUT/DELETE | primitive Properties, Key, `null`, Missing (`ODV4-CRUD-01..06`) | Concurrency via If-Match | Deep Insert |
| Query | `eq/ne/gt/ge/lt/le`, `and/or/not`, `$select`, `$orderby`, `$top`, `$skip`, `$count` (`ODV4-Q-01..12`) | zusätzliche Funktionen | `$search`, `$apply` |
| EDM | `Edm.String`, `Boolean`, `Int32`, `Int64`, `Decimal`, `Guid`, `Date`, `DateTimeOffset` (`ODV4-T-01..08`) | weitere Primitive | komplexe Navigationstypen |
| Header | `Accept`, `OData-Version`, `Content-Type` (`ODV4-H-01..03`) | `Prefer: return=…`, `If-Match` | herstellerspezifische Header |
| Fehler | 400, 404, 501, Coverage (`ODV4-E-01..04`) | 412 | proprietäre Fehlercodes |

Die Fixture-IDs sind **geplante Testadressen**, keine bereits ausgeführten Tests.
HTTP-/OData-Verhalten ist an Referenz-Fixtures und dem Lastenheft zu messen; die Matrix erweitert den Muss-Umfang nicht.

### SPEC-005 — Admin-API (vorläufiger Contract)

| Methode | Pfad | Bedeutung |
|---|---|---|
| GET | `/healthz` | Prozessbereitschaft (ohne Authentisierung) |
| GET | `/__admin/sessions` | Aufzeichnungen mit Kennung und Status auflisten (leere Liste, wenn keine existieren) |
| POST | `/__admin/sessions` | Neue Recording-Session anlegen (`{"service_id": "...", "id": "..."}`, `id` optional); Status `open`; vergebene `id` `ODR-E7010` |
| POST | `/__admin/sessions/{id}/close` | Aufzeichnung beenden: Status `complete` oder `incomplete` (Regel unter „Status“ in `LH-FA-13.a`); unbekannt: `ODR-E7003` |
| GET | `/__admin/sessions/{id}/exchanges` | Exchanges lesen |
| DELETE | `/__admin/sessions/{id}` | Session löschen (Retention, `LH-RB-04.a`); Quelle einer Baseline `ODR-E7004`, Status `open` `ODR-E7009` |
| GET | `/__admin/coverage` | Coverage lesen |
| POST | `/__admin/imports` | Recording-Session explizit in neue Baseline importieren |
| GET | `/__admin/imports/{id}` | Import-Report und Provenienz lesen; unbekannt: `ODR-E7005` |
| GET | `/__admin/workspaces` | Workspaces mit Kennung, Service, Revision, Baseline und Erstellzeit auflisten |
| POST | `/__admin/workspaces` | Workspace aus einer Baseline erzeugen (`{"baseline_snapshot_id": "..."}`; `id` optional, sonst serverseitig vergeben; vergebene `id` `ODR-E7007`, Snapshot ohne Art `baseline` `ODR-E7008`) |
| GET | `/__admin/workspaces/{id}` | Workspace lesen; unbekannt: `ODR-E7001` |
| DELETE | `/__admin/workspaces/{id}` | Workspace samt Cursor entfernen |
| POST | `/__admin/workspaces/{id}/reset` | Baseline und alle Cursor zurücksetzen; unbekannte ID: `ODR-E7001` |
| GET | `/__admin/workspaces/{id}/snapshots` | Snapshots auflisten |
| POST | `/__admin/workspaces/{id}/snapshots` | Benannten Snapshot erzeugen (`{"name": "..."}`); Name im Workspace vergeben: `ODR-E7006` |
| POST | `/__admin/workspaces/{id}/restore` | Snapshot wiederherstellen (`{"snapshot_id": "..."}`); Cursor unverändert; Baseline-Snapshot `ODR-E7008` |
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
| `SPEC-010` | `HYBRID_FALLBACK` | `disabled` | `disabled` oder `enabled`; Hauptschalter des Upstream-Fallbacks im Modus `hybrid`. Erst `enabled` macht `FALLBACK_ALLOW_HOSTS` (Ziel) und `FALLBACK_WRITE_ALLOW` (mutierend) wirksam; `HYBRID_REPLAY_FALLBACK` (Kategorien des Replay-Fallbacks) ist davon unabhängig, weil der Replay-Fallback keinen Upstream kontaktiert (`LH-FA-12.a`); keine unerwarteten Backend-Aufrufe |
| `SPEC-011` | `STRICT_COVERAGE` | `true` | Keine irreführenden Collection-Antworten |
| `SPEC-012` | `UPSTREAM_TIMEOUT` | `30s` | Begrenzte Wartezeit |
| `SPEC-013` | `MAX_CAPTURE_BODY` | `8 MiB` | Schutz vor unbeschränkter Ablage; einzige Quelle der Grenze (`SPEC-053`, `LH-FA-02.b`) |
| `SPEC-014` | `SQLITE_BUSY_TIMEOUT` | `5000ms` | Geordnete Konkurrenz |
| `SPEC-015` | `RECORDING_FORMAT_VERSION` | `1` | Versionierbare Persistenz |
| `SPEC-054` | `SERVICE_ROOT_PATH` | `/` bei genau einem Service, sonst Pflichtangabe je Service über `SERVICES_FILE` | Adressierung (`LH-FA-04.c`) |
| `SPEC-055` | `PUBLIC_BASE_URL` | abgeleitet aus `LISTEN` nur bei Loopback-Host; in `mock` und `hybrid` sonst (auch Wildcard-Host) Pflichtangabe | Basis für `Location`, `@odata.id`, `@odata.nextLink` |
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
| `SPEC-091` | `SERVICES_FILE` | keine | Konfigurationsdatei für mehrere Services: eine Liste mit `id`, `service_root` und `upstream_url` je Service; ohne Datei gilt genau ein Service aus `SERVICE_ID`, `SERVICE_ROOT_PATH` und `UPSTREAM_URL`; beides zugleich verhindert den Start (`ODR-E2001`) |
| `SPEC-098` | `SERVICE_ID` | `default` | Kennung des einzigen Services; Teil des Match-Keys, bei Änderung nach einer Aufzeichnung passen vorhandene Recordings nicht mehr |
| `SPEC-092` | `LINK_REWRITE` | `off` | Umschreiben absoluter Upstream-Links und Weiterleitungen; nur `off` ist in der ersten Version zulässig, jeder andere Wert verhindert den Start (`ODR-E2001`) |

Syntax der Werte (Dauern, Größen, Listen, leere Werte), Prüfung und Verhalten bei `mock` und `hybrid`: `SPEC-110`.

---

## 4. Fehler-Codes und Logging-Felder

| ID | Code | Bedingung | Aktion |
|---|---|---|---|
| `SPEC-016` | `ODR-E2001` | Ungültige Konfiguration | Start verhindern |
| `SPEC-064` | `ODR-E2101` | Admin-Authentisierung fehlt oder ungültig | HTTP 401, kein Zustandswechsel |
| `SPEC-017` | `ODR-E2401` | Capture/Redaction fehlgeschlagen | Body nicht persistieren, Diagnose |
| `SPEC-018` | `ODR-E3001` | Recording/Blob beschädigt oder nicht ladbar (Datei fremd oder von neuerer Schemaversion) | Laden verweigern |
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
| `SPEC-099` | `ODR-E5009` | Record ohne annehmende Session des Services (keine offene, oder sie wird gerade beendet) | HTTP 409, Request wird nicht weitergeleitet |
| `SPEC-100` | `ODR-E7010` | Session-ID bereits vergeben | HTTP 409 (Admin) |
| `SPEC-097` | `ODR-E7009` | Session ist `open`: Replay, Import und Löschen abgewiesen | HTTP 409 |
| `SPEC-111` | `ODR-E7011` | Session ist nicht `open`: Beenden abgewiesen (Übergang aus einem geschlossenen Zustand) | HTTP 409 (Admin), nichts wird geändert |
| `SPEC-112` | `ODR-E7012` | Exchange unbekannt | HTTP 404 (Admin) |
| `SPEC-113` | `ODR-E9001` | Interner Fehler (ein Fehler, den die Abbildung keinem Code zuordnet, `SPEC-109`) | HTTP 500, Ursache nur im Log |

Fallback-Entscheidungen (`LH-FA-12.b`) hängen am `ODR-`-Code, nie am HTTP-Status.
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

Protokollversionen werden beim Start und pro Service validiert.
Herstellerspezifische OData-Erweiterungen sind nicht automatisch Bestandteil des Mock-Vertrags.

Abnahmeparameter für Performance und Concurrency stehen in `LH-QA-01.a` und `LH-QA-03.a`; der Messaufbau muss reproduzierbar im Test-Harness hinterlegt sein.

---

## 7. Historie

| Datum | Änderung |
|---|---|
| 2026-10-21 | Streaming-Redaktor für Bodies (`SPEC-118`): Typ, Prüfung, Wertarten unter treffenden Namen, Grenzen, Fehlergründe, Aufruferpflichten; Zeiger in `SPEC-046` und `SPEC-115` |
| 2026-10-20 | `SPEC-116` Nr. 5 und 8: Vary-Werte aus der redigierten Darstellung, Anzahl der vorhandenen Namen am Anfang der Vary-Folge; Einleitung: Grenze des Bodys benannt |
| 2026-10-20 | Randformen von `SPEC-115` und `SPEC-116` entschieden: Hex-Schreibweise von `\u00XX` in der Kanonform, Marke und Text der Art 2 als Zeichenfolge, Nicht-Token-Header-Namen, `#` vor dem ersten `?` in `Location`, Escapes ohne UTF-8 im Pfad, Hex-Schreibweise im Text der Art 2 |
| 2026-10-20 | Redaction von Headern, Query und `Location`, Pfadmuster (`SPEC-115`), Kodierung des Match-Keys samt `normalized_path`, Query-Paaren, Vary-Headern und Body-Hash (`SPEC-116`), Replayfähigkeit als Funktion (`SPEC-117`) |
| 2026-10-20 | `SPEC-114` *Ungeprüft*: der Ausfall der Zufallsquelle (Nr. 7) ist als nicht getestet benannt |
| 2026-10-20 | `SPEC-114` Nr. 10: der Zähler von `FestIDs` wird dezimal geschrieben (auf `…09` folgt `…10`) |
| 2026-10-20 | Vertrag von `Clock` und `IDGenerator` (`SPEC-114`): UTC ohne Rundung, nicht fallende Uhr, Bereich der Systemzeit, Form und Herkunft der Kennungen, Kollision, Client-Kennungen, Test-Implementierungen |
| 2026-10-20 | Randformen von `SPEC-109` und `SPEC-110` entschieden: `none` in Listen, kanonische Schreibweise der Header-Mengen, Werte der Dienste-Datei (doppelt, `null`, leer), Form von `UPSTREAM_URL` in jedem Modus, Hostname und Port, Sortierung der Meldung, Kontextfehler im `errors.Join` |
| 2026-10-20 | `SPEC-110` Nr. 9 und 12: ein Feld des falschen Typs in `SERVICES_FILE` meldet nur den Typ (eine Zeile je Schlüssel); die Sortierung nach Grund bleibt als Totalordnung ohne beobachtbaren Fall |
| 2026-10-20 | `SPEC-110` Nr. 10 und 12: ein formal ungültiger Wahrheitswert von `UPSTREAM_TLS_VERIFY` oder `UPSTREAM_TLS_INSECURE_SKIP_VERIFY` meldet nur die Form, der Widerspruch wird nur bei zwei gültigen Werten geprüft; je Schlüssel höchstens ein Grund |
| 2026-10-20 | Fehlermodell und Abbildung der Port-Fehler (`SPEC-109`), Startkonfiguration (`SPEC-110`), Codes `ODR-E7011`, `ODR-E7012`, `ODR-E9001` |
| 2026-10-18 | Verträge der Werkzeuge liegen in `harness/vertraege/`; der Abschnitt entfällt |
| 2026-10-18 | Review-Findings eingearbeitet: Admin-Authentisierung, Snapshot-Restore ohne Cursor-Änderung, Weg zu `complete`, Header-Redaction ohne Replay-Einschränkung, Adressierung, Fehlercodes, Konkretisierung von `LH-QA-*`/`LH-RB-*` |
| 2026-10-17 | d-migrate 1.7.1, Integritätsconstraints, Replay-/Capture-Verträge, SPEC-ID-Korrektur |
| 2026-10-17 | Initiale technische Spezifikation auf Basis des Lastenhefts und der Kursvorlage |

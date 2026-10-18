# Review-Report: spec-straten-nachreview — 2026-10-18

**Review-Art:** Design (Nachreview der drei Spec-Straten, des Persistenzschemas und der Schema-Werkzeuge; kein Code, kein Slice-Plan, keine ADR)

**Gegenstand (Arbeitsbaum, nicht committet):** `spec/lastenheft.md` (v0.6.0 Draft), `spec/spezifikation.md`, `spec/architecture.md`, `tools/schema/schema.yaml` (unverändert seit `81f8228`, gegen die Spezifikation geprüft), `harness/mk/schema.mk`, zwei neue Zeilen in `harness/README.md` §Sensors, `.harness/skills/reviewer.md`

**Vorgänger:** `docs/reviews/2026-10-18-spec-straten-initial.md` (35 Findings, F-1..F-35). Dieser Report nummeriert neu (N-1..).

**Skill:** `.harness/skills/reviewer.md` (Arbeitsbaum, mit Repo-HIGH-Regeln) · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-18

**Eingangs-Kontext:** die drei Straten, `tools/schema/schema.yaml`, `harness/mk/schema.mk`, `harness/README.md`, `AGENTS.md` §3.4/§3.7, `.d-check.yml`, Vorgänger-Report.

**Sensor-Läufe dieses Reviews:**

- `make docs-check` → 23 Dateien, 0 Befunde.
- `make schema-validate` → Validation passed (9 Tabellen, 60 Spalten, 1 Index, 20 Constraints).
- `make schema-generate` → 11 Statements, 0 skipped objects, 0 Warnungen. Das DDL lief in SQLite 3.45.1 mit `PRAGMA foreign_keys=ON`.
- Skriptete Gegenproben auf diesem DDL: Löschen einer Session mit Cursor → `FOREIGN KEY constraint failed`; zweiter `imports`-Datensatz mit gleicher `id` → `UNIQUE constraint failed` (siehe N-3, N-14).
- Skript-Lauf über die Spezifikation: `SPEC-001`..`SPEC-073` lückenlos, je einmal definiert, jede referenzierte SPEC-ID aufgelöst.

Nicht geprüft: DoD-/Abnahme-Konformität (Verifier), Bedarf gegen die Realität (Validator). Die OData-Semantik-Aussagen (N-24) stammen aus Fachwissen, nicht aus der Primärquelle.

---

## Teil A — Status der Vorgänger-Findings (selbst am Text nachgeprüft)

| Vorgänger | Status | Beleg / Rest |
|---|---|---|
| F-1 Restore/Cursor | behoben | `LH-FA-10.a`, `SPEC-052`, `LH-FA-09.b`, Arch 2.3 und Restore-Diagramm sagen übereinstimmend „Cursor unverändert". Rest: N-18 (Reset „löscht" vs. „setzt auf 0"). |
| F-2 Admin-Auth | behoben | `LH-FA-13.a` (Bearer, `ADMIN_TOKEN_FILE`, `ODR-E2101`/`ODR-E2001`), `SPEC-064`, Arch §5. Neue Aspekte: N-12. |
| F-3 Weg zu `complete` | behoben | `LH-FA-18.a` Schritt 2 „Vollständigkeitsbeweis (einziger Weg zu `complete`)". |
| F-4 Default-Replay | teilweise | Header-Redaction ist ausgenommen (`SPEC-045`). Rest: Request-Body-Redaction sperrt weiterhin das Replay der Antwort (N-5). |
| F-5 `MAX_CAPTURE_BODY` | behoben | nur `SPEC-013` = 8 MiB; kein „64 MiB" mehr in den Straten (grep). |
| F-6 QA ohne Konkretisierung | teilweise | `LH-QA-01.a`, `-02.a`, `-03.a`, `-05.a`, `-06.a`, `-07.a` vorhanden. `LH-QA-04` hat keinen Abschnitt (N-20). |
| F-7 RB ohne Konkretisierung | behoben | `LH-RB-01.a`..`-04.a`. Neue Aspekte: N-3, N-19. |
| F-8 volatile Header, Key-Vergabe | behoben | `SPEC-063`, `LH-QA-02.a`, `LH-FA-07.a`. Rest: N-18 (ID-Quelle). |
| F-9 Adressierung | behoben | `LH-FA-04.c`, `SPEC-054`/`-055`. Rest: N-9. |
| F-10 Match-Key/Redaction | teilweise | Regel steht (`LH-FA-14.a`). Folge: N-6. |
| F-11 Import/Platzhalter | behoben | Marker `[REDACTED]`, Property-Maske `unknown`. |
| F-12 Admin-API-Lücken | behoben im Text | `SPEC-005` erweitert, `ODR-E7001`/`E7002`/`E7003`. Das Schema trägt Teile nicht (N-1, N-2, N-13). |
| F-13 Fehlerklassen | teilweise | `ODR-E6004` neu, `E5003` einheitlich 400. Offen: HTTP-Status zu `E5001`/`E5002`/`E6003` (N-7). |
| F-14 Ports/Importkanten | teilweise | Codec- und Cursor-Port, `baselineimport` ergänzt. Rest: N-10. |
| F-15 Metadaten-Granularität | behoben | `LH-FA-06.a` Schritt 3 (`ODR-E6004` vs. `unsupported_features`). |
| F-16 Matching-Negativfall | behoben | `MATCH_INVALID_ENCODING`, `+`-Regel. Schlüssel nicht im Defaults-Katalog (N-8). |
| F-17 Schema vs. Capture-Matrix | teilweise | Felder in `SPEC-001` und Schema, aber `body.state` ist im Schema nicht abbildbar (N-2). |
| F-18 Sprachbezug Architektur | behoben | kein `os.Exit`/`database/sql`/`net/http`/Fremd-Repo mehr; verbleibend nur Code-Modul-Pfade (`cmd/`, `internal/`, `main.go`), die §3.4 erlaubt. |
| F-19 Non-strict | behoben | Header `X-Recorder-Coverage`. |
| F-20 normativer Inhalt nur in Arch | behoben | `LH-FA-02.c` in der Spezifikation; Arch verweist. |
| F-21 `Prefer` Muss/Soll | behoben | `LH-FA-07.b`: Muss-Default ohne `Prefer`, Auswertung Soll. |
| F-22 Write-Fallback-Freigabe | behoben im Text | `FALLBACK_WRITE_ALLOW`. Rest: N-8, N-17. |
| F-23 Redaction-Default | teilweise | Header-Liste erweitert. Query/JSON-Default leer (N-11). |
| F-24 Import-Statuscodes | behoben | `SPEC-047`, `SPEC-051`, `SPEC-069`. Rest: N-14. |
| F-25 Stratum-Inhalt | behoben | Lastenheft-Historie bereinigt (v0.6.0 als Tatsachenberichtigung gekennzeichnet). |
| F-26 Diagramme | behoben | Pfad `/__admin/workspaces/{id}/reset`, Restore-Diagramm, Replay-Miss → Upstream in 4.6. Neu: N-10 (Teilnehmer „Workspace Repository"). |
| F-27 Grenzfälle | teilweise | `$top`/`$skip`, `/$count` geklärt. `strict`-Mismatch ohne Code (N-4). |
| F-28 Null-Semantik | behoben im Text | ausformuliert; Konformität siehe N-24. |
| F-29 Dopplung | gemildert | Arch verweist nun auf `LH-FA-09`/`-10`/`-12.a`/`-02.c`; Restdopplung in Arch 2.4/2.5 inhaltlich konsistent. |
| F-30 Phantom-Targets | behoben | Spec §7 sagt „Vorhanden ist heute nur `make docs-check`"; `schema-*` stehen in `make help` und §Sensors. |
| F-31 Versionsbezug | behoben | `SPEC-004`: Schemaversion unabhängig vom Lastenheft. |
| F-32 QA-01-Freigabe | offen (Owner) | unverändert (N-26). |
| F-33 Skill | behoben | zwei und mehr Repo-HIGH-Regeln im Skill. |
| F-34 d-migrate unverifiziert | teilweise bestätigt | validate, generate und DDL-Lauf bestätigt (siehe Sensor-Läufe); Negativtests und Migrationskette weiterhin ungeprüft, wie `SPEC-004` selbst sagt. |
| F-35 LH-ID-Klasse | offen | `.d-check.yml` führt die `LH-*`-Regel weiterhin auskommentiert (N-27). |

---

## Teil B — Findings

HIGH: keine. Begründung: Die vier früheren HIGH-Befunde sind behoben; neue Widersprüche zwischen Spezifikation/Architektur und Lastenheft wurden nicht gefunden (siehe Negativbefunde). Der Replay-Cursor-Befund N-4 liegt an der Grenze, ist aber eine Mehrdeutigkeit ohne Lastenheft-Verletzung.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| N-1 | MEDIUM | Das Schema kann benannte, workspacebezogene Snapshots nicht abbilden. `snapshots` hat weder `name` noch `workspace_id`. Folge: `GET /__admin/workspaces/{id}/snapshots` kann nicht nach Workspace filtern, `{"name": ...}` hat keine Spalte, und ein Restore könnte einen Snapshot eines anderen Workspace desselben Service laden (`ODR-E5004` prüft nur den Service). | LH-FA-10, LH-FA-13, SPEC-004 | `spec/spezifikation.md` · „Benannten Snapshot erzeugen"; `tools/schema/schema.yaml` · „uq_snapshots_id_service" | ja – DDL-Gegenprobe / Admin-Contract-Test | Schema trägt Spec-Vertrag nicht |
| N-2 | MEDIUM | `body.state` (`absent`/`present`/`discarded`) für Request und Response hat im Schema keine Entsprechung. Es gibt nur ein `body_integrity` und ein `redaction_status` je Exchange und nullable `request_blob`/`response_blob`; „leerer Body" und „verworfener Body" sind für jede Richtung nicht unterscheidbar, obwohl `SPEC-001` behauptet, die Felder entsprächen den Spalten. | LH-FA-02 (Boundary), SPEC-001 | `spec/spezifikation.md` · „`body.state` unterscheidet `absent`"; `tools/schema/schema.yaml` · „request_blob:" | ja – Schema-Validierung plus Roundtrip-Test | Schema trägt Spec-Vertrag nicht |
| N-3 | MEDIUM | Session-Löschen (`DELETE /__admin/sessions/{id}`, Retention beim Start) ist gegen das Schema nicht ausführbar spezifiziert. Gegenprobe: Eine Session mit `replay_cursors`-Zeile scheitert an `fk_cursor_recording_service` (`ON DELETE RESTRICT`), ein Fehlercode dafür fehlt (`ODR-E7004` meint nur Baseline-Provenienz). Provenienz liegt nur im Snapshot-Manifest-Blob bzw. `report_json`, es gibt keine abfragbare Beziehung zu `sessions`. Offen bleibt, was der Start-Löschlauf mit einer von einer Baseline referenzierten Session tut (Start abbrechen oder überspringen). | LH-RB-04, LH-FA-03 | `spec/spezifikation.md` · „Eine Session, auf die eine Baseline-Provenienz verweist, wird abgewiesen"; `tools/schema/schema.yaml` · „fk_cursor_recording_service" | ja – DDL-Gegenprobe, Integrationstest | Löschpfad nicht schemakonform spezifiziert |
| N-4 | MEDIUM | Die Cursor-Semantik ist nicht eindeutig. `LH-FA-04.a` verlangt „die nächste noch nicht verbrauchte passende Exchange", der Cursor ist aber ein einzelner `position`-Integer je `(workspace, replay_session, recording_session)` (`SPEC-003`, Schema); `strategy` aus `SPEC-003` hat keine Spalte. Aufgezeichnet A(x), B, A(y); ein Client fragt B, dann A, dann A: ob A(x) vor A(y) kommt, hängt davon ab, ob `position` ein High-Water-Mark oder ein Verbrauchsmaß ist. Zusätzlich hat der `strict`-Mismatch (Key existiert, ist aber nicht die nächste erwartete Exchange) weder `ODR-E5001` noch `ODR-E5002` als zutreffenden Code. | LH-FA-04 | `spec/spezifikation.md` · „die nächste noch nicht verbrauchte passende Exchange"; · „`ODR-E5002`, wenn Exchanges mit diesem Key existieren"; `tools/schema/schema.yaml` · „ck_replay_position" | ja – `make test-determinism` mit permutierter Request-Reihenfolge | Replay-Cursor-Zustand unterbestimmt |
| N-5 | MEDIUM | Die Matrix `SPEC-045` kennt eine einzige `redaction`-Angabe je Exchange. Wird nur der Request-Body redigiert (z. B. `REDACT_JSON_PATHS` auf ein Passwortfeld), gilt der Exchange als `applied` und ist im Default nicht replayfähig, obwohl die ausgelieferte Response unverändert ist und der Match-Key ohnehin über die redigierte Darstellung läuft. LH-FA-04 Happy Path scheitert im Default für jeden Austausch mit redigiertem Request-Body. | LH-FA-04, LH-FA-14 | `spec/spezifikation.md` · „Nur explizit aktiviertes `semantic_or_limited` (`REPLAY_REDACTED_BODIES=true`)" | ja – Integrationstest Record mit redigiertem Request-Body → Replay im Default | Default-Policy sperrt Muss-Happy-Path (Rest) |
| N-6 | MEDIUM | Der Match-Key läuft ausschließlich über die redigierte Darstellung. Requests, die sich nur im redigierten Teil unterscheiden (Query-Parameter aus `REDACT_QUERY_PARAMS`, redigierter JSON-Pfad), erhalten denselben Key; bei `first`/`last` wird die falsche Response geliefert, ohne dass der Unterschied gemeldet wird. Das Lastenheft verlangt, dass relevante Unterschiede „nicht unbemerkt zusammenfallen". | LH-FA-05, LH-FA-14 | `spec/spezifikation.md` · „Der Match-Key (LH-FA-05.a) wird ausschließlich über die redigierte Darstellung gebildet." | ja – Match-Key-Unit-Test mit zwei Requests, die sich nur im redigierten Feld unterscheiden | Matching/Redaction-Wechselwirkung (Rest) |
| N-7 | MEDIUM | Für `ODR-E5001`, `ODR-E5002` und `ODR-E6003` fehlt der HTTP-Status („Definierte Fehlerantwort", „Explizite Mock-Fehlerantwort"). `LH-QA-05.a` behauptet, jede Antwort trage Code und HTTP-Status; die Negativtests aus `LH-QA-05` können den Status nicht prüfen. | LH-QA-05, LH-FA-04, LH-FA-11 | `spec/spezifikation.md` · „Explizite Mock-Fehlerantwort" | ja – Negativtests `make test-contract` | Fehlerklasse ohne Statusabbildung |
| N-8 | MEDIUM | Viele im Text benannte Konfigurationsschlüssel fehlen im Defaults-Katalog §3 (nur `grep`-Treffer im Text, keine `SPEC-`-Zeile): `FALLBACK_WRITE_ALLOW`, `MATCH_INVALID_ENCODING`, `MATCH_BODY_MODE`, `REPLAY_REDACTED_BODIES`, `REDACT_QUERY_PARAMS`, `REDACT_JSON_PATHS`, `UPSTREAM_CA_FILE`, `UPSTREAM_TLS_INSECURE_SKIP_VERIFY`, `DATA_DIR`, `LISTEN_TLS_*`/`ADMIN_TLS_*`, `assignment`. Der Upstream-Zielort, der Betriebsmodus und das opt-in für `ordered`-Assignment haben gar keinen Schlüssel. Sicherheitsrelevante Defaults (Write-Fallback leer, TLS-Prüfung an) sind damit nicht an einer Stelle prüfbar. | LH-FA-12, LH-RB-02, LH-FA-05 | `spec/spezifikation.md` · „`FALLBACK_WRITE_ALLOW`, Liste aus Methode und Pfadmuster, Default leer" | ja – `make test-security` gegen Defaults | Konfigurationsschlüssel ohne Katalogeintrag |
| N-9 | MEDIUM | Im Hybrid-Modus ist der Replay-Fallback nicht adressierbar: `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` sind „nur im Replay Pflicht", für `hybrid` ist nur der Workspace Pflicht. Woher der Replay-Fallback Recording-Session und Cursor-Identität nimmt, steht nirgends. | LH-FA-12, LH-FA-04 | `spec/spezifikation.md` · „`X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` sind nur im Replay Pflicht." | ja – Integrationstest Hybrid mit Replay-Fallback | Adressierung im Hybrid-Pfad fehlt |
| N-10 | MEDIUM | Der Port-Katalog der Architektur deckt die neuen Spec-Verträge nicht. `RecordingRepository` ist „append-only", `SPEC-005`/`LH-RB-04.a` verlangen Session-Löschen und Retention (kein Port, kein Inbound-Use-Case); `WorkspaceUseCase` kennt kein Auflisten/Löschen, `SnapshotUseCase` kein Auflisten, für `/__admin/diagnostics` gibt es keinen Use Case. Das Diagramm 4.7 nutzt einen Teilnehmer „Workspace Repository", der in 2.3 nicht existiert (wer persistiert `workspaces`?). Das Architektur-Gate würde den gezeichneten Fluss ablehnen, solange der Port fehlt. | Maintainability, AGENTS.md §3.4 | `spec/architecture.md` · „`RecordingRepository` | Sessions und Exchanges append-only persistieren und lesen"; · „participant W as Workspace Repository" | ja – `make arch-check` (Target existiert noch nicht) | Port-Katalog und Importkanten inkonsistent (zweites Auftreten) |
| N-11 | MEDIUM | Redaction-Defaults für Query-Werte und Body-Felder sind leer (`REDACT_QUERY_PARAMS`, `REDACT_JSON_PATHS` ohne Default und ohne Katalogeintrag). `raw_query` wird unredigiert im Recording persistiert; ein Token im Query-String (`?access_token=…`) landet im Recording, ohne dass irgendeine Zusage verletzt wird. Der Skill führt „Query-Werte" als Geheimnis-Senke. | LH-FA-14, LH-QA-04 | `spec/spezifikation.md` · „Zusätzlich sind konfigurierbare Query- und JSON-Felder redigierbar" | ja – Canary-Test mit Secret im Query-String | Redaction-Default unvollständig (Rest) |
| N-12 | MEDIUM | Der Admin-Token wird bei `ADMIN_LISTEN=0.0.0.0:8081` über unverschlüsseltes HTTP gesendet, weil Admin-TLS optional ist und „ohne sie die Listener HTTP sprechen". Die Spezifikation erklärt den Bind auf alle Interfaces für zulässig, „weil das Token den Zugriff schützt"; auf dem Transportweg ist das Token abgreifbar, danach steht die gesamte Admin-API offen (Recordings, Import, Löschen). | LH-QA-04, LH-RB-02, LH-FA-13 | `spec/spezifikation.md` · „in einem Container ist `ADMIN_LISTEN=0.0.0.0:8081` zulässig, weil das Token den Zugriff schützt" | ja – `make test-security` (Netzwerkzugriffstest) | Admin-Transportschutz nicht gefordert |
| N-13 | MEDIUM | Statuswerte fehlen. `LH-FA-13` Happy Path verlangt „Kennungen und Status" von Sessions und Workspaces; `SPEC-005` listet „Status", aber `workspaces` hat keine Statusspalte, `sessions.status` und `imports.status` sind ohne Check-Constraint und ohne Wertemenge in der Spezifikation. | LH-FA-13 | `spec/spezifikation.md` · „Workspaces mit Kennung, Service, Revision und Status auflisten"; `tools/schema/schema.yaml` · „status:" | ja – Admin-Contract-Test | Statusmodell nicht definiert |
| N-14 | MEDIUM | Die Wiederverwendung der `import_id` nach fehlgeschlagenem Import kollidiert mit dem Schema. `imports.id` ist Primärschlüssel, und `SPEC-004` erlaubt Zeilen mit `baseline_snapshot_id=NULL` für „noch nicht abgeschlossene Imports". Gegenprobe: Ein zweiter Insert mit gleicher `id` scheitert mit `UNIQUE constraint failed`. Nach der Spezifikation ist die Wiederholung erlaubt, nach dem Schema nicht; ob die Zeile zurückgesetzt oder `ODR-E6102` geliefert wird, ist offen. | LH-FA-18, SPEC-047 | `spec/spezifikation.md` · „nach einem fehlgeschlagenen Import darf dieselbe `import_id` erneut verwendet werden" | ja – Import-Contract-Test (Wiederholung) | Idempotenz-Regel nicht schemakonform |
| N-15 | MEDIUM | Reset ist als atomar inklusive Cursor-Reset spezifiziert; Diagramm 4.5 und die Ports zeigen zwei getrennte Schritte über `EntityRepository` und `CursorRepository`. Ob beide in derselben fachlichen Transaktion laufen, steht nicht da. Ein Absturz dazwischen lässt zurückgesetzten Entity-State mit unverändertem Cursor zurück, den `LH-FA-09` Boundary ausschließt. | LH-FA-09 | `spec/architecture.md` · „W->>K: Reset all cursors of workspace"; `spec/spezifikation.md` · „Reset ist eine atomare Wiederherstellung des Baseline-Zustands inklusive Cursor-Reset" | ja – Fault-Injection-Test (`SPEC-048`) | Atomarität über zwei Ports nicht festgelegt |
| N-16 | LOW | Der Workspace verweist über `fk_workspace_baseline_service` auf irgendeinen Snapshot desselben Service; dass er `snapshot_kind='baseline'` sein muss, wird nur für Imports „innerhalb derselben Transaktion" geprüft, nicht für die Workspace-Erzeugung (`SPEC-043` prüft nur Existenz). Ein `workspace`-Snapshot als Baseline hätte keine `coverage`-Zeilen. | SPEC-004 | `spec/spezifikation.md` · „`baseline_snapshot_id` wird vor Workspace-Erzeugung auf Existenz geprüft" | ja – Workspace-Erzeugung mit Workspace-Snapshot | Schema-Invariante nur teilweise erzwungen |
| N-17 | LOW | Die Architektur nennt als Security-Test „Verbot mutierender Hybrid-Fallbacks" und als Risiko-Gegenmaßnahme „read-only Default"; die Spezifikation erlaubt mutierenden Fallback mit `FALLBACK_WRITE_ALLOW`. Der Testauftrag ist enger als die Zusage. | Maintainability | `spec/architecture.md` · „Verbot mutierender Hybrid-Fallbacks" | nein | Normative Dopplung zwischen Straten (Rest) |
| N-18 | LOW | Zwei kleine Widersprüche zu Cursor und IDs: `SPEC-052` und Arch 4.5 „löschen"/„reset" die Cursor, `LH-FA-04.b`/`LH-FA-09.b` „setzen auf Position 0". Replay-Session-IDs sind laut `SPEC-052`/`SPEC-004` clientdefiniert, `LH-QA-02.a` führt sie unter IDs aus `IDGenerator`, `LH-FA-04.b` sagt „beim Start eines Testlaufs vergeben". | LH-FA-04, LH-QA-02 | `spec/spezifikation.md` · „Zeit und IDs (Korrelations-ID, Workspace-/Replay-Session-IDs) kommen aus `Clock`/`IDGenerator`" | ja – Determinismus-Test | Begriffs-/Quellenabweichung innerhalb der Spezifikation |
| N-19 | LOW | Nach Session-Löschung räumt „die Bereinigung" die Blobs ab, ohne Auslöser oder Frist (`SPEC-048` beschreibt nur die Referenzprüfung). Für `LH-RB-04` (Aufbewahrungsvorgaben, personenbezogene Daten) bleibt unbestimmt, wann die Daten tatsächlich vom Datenträger verschwinden. | LH-RB-04 | `spec/spezifikation.md` · „nicht mehr referenzierte Blobs räumt die Bereinigung ab" | nein | Löschzeitpunkt unbestimmt |
| N-20 | LOW | `LH-QA-04` (Datensicherheit, Canary-Messmethode, Netzwerkzugriffstest) hat keinen eigenen Abschnitt `LH-QA-04.a`; der Bezug läuft nur implizit über `LH-FA-13.a`/`-14.a`/`SPEC-041`. Alle übrigen QA- und RB-Einträge haben einen. | LH-QA-04, Source Precedence | `spec/spezifikation.md` · „### LH-QA-05.a — Fehlerklassen" (Nachbar; `LH-QA-04.a` fehlt) | ja – `make doc-complete` nach Aktivierung der LH-ID-Klasse | ID-Abdeckung Spec fehlt (QA, Rest) |
| N-21 | LOW | Kleine Abweichungen Spec ↔ Schema: `SPEC-003` `canonical_key` vs. Spalte `entity_key`; `SPEC-001` `sequence` vs. `sequence_no`; `LH-FA-02.a` Schritt 1 „Beginn und Ende jedes Austauschs erfassen" ohne Spalte; `services.service_root` ist gespeichert, `SPEC-054` leitet den Root aus der Service-Anzahl ab. | SPEC-004 | `tools/schema/schema.yaml` · „service_root:" | ja – Schema-Validierung | Namens-/Feldabweichung Spec ↔ Schema |
| N-22 | LOW | Fehlerabbildung an Rändern: Zugriff auf eine Property mit nicht unterstütztem EDM-Typ ergibt „501 mit `ODR-E6001`", obwohl `ODR-E6001` laut `SPEC-022` „Nicht unterstützte OData-Query" ist; die „abzuweisende" PATCH-Aktion auf partieller Entity (`SPEC-044`) nennt keinen Code und keinen Status. | LH-QA-05 | `spec/spezifikation.md` · „Zugriffe darauf ergeben `501` mit `ODR-E6001`" | ja – Contract-Test | Fehlercode-Zuordnung am Rand |
| N-23 | LOW | Der Service-Root wechselt von `/` zu `/{service_id}/`, sobald ein zweiter Service konfiguriert wird; die URLs bestehender Clients ändern sich dadurch still. `PUBLIC_BASE_URL` (Default `http://127.0.0.1:8080`) ist nicht an `LISTEN` gekoppelt; eine abweichende `LISTEN` oder ein Container-Port-Mapping erzeugt `@odata.nextLink`/`Location` auf dem falschen Ziel. | LH-FA-01, LH-FA-07 | `spec/spezifikation.md` · „Ist genau ein Service konfiguriert, ist der Root `/`" | ja – Contract-Test mit zwei Services | Adressierung instabil bei Konfigurationswechsel |
| N-24 | INFO | `LH-FA-08.b` setzt `eq`/`ne` zwischen zwei Properties auf `null`, sobald eine `null` ist. Nach meiner Kenntnis von OASIS OData V4.0 Part 2 gilt `null eq null` als wahr; nicht gegen die Primärquelle geprüft. Zuständig: Architect. | LH-RB-01 | `spec/spezifikation.md` · „`eq` und `ne` zwischen zwei Properties ergeben `null`, sobald eine davon `null` ist." | ja – Referenzdienst-Vergleich `ODV4-Q-*` | Referenzierte Semantik nicht verifiziert |
| N-25 | INFO | `SPEC-004` ist durch diesen Lauf teilweise bestätigt (validate, generate ohne übersprungene Objekte, DDL läuft in SQLite). Die Schema-Targets sind „kein Gate", während `SPEC-004` „CI muss … durchführen" und `foreign_key_check`/`integrity_check` als „verpflichtende Gates" nennt; ein Gate dafür existiert nicht. Mit den Findings N-1..N-3, N-13, N-14 ändert sich das Schema, `version` (`0.5.0`) muss dabei der Regel aus `SPEC-004` folgen. Zuständig: Planner/Implementer. | AGENTS.md §4 | `spec/spezifikation.md` · „Nicht belegt sind die Negativtests der Constraints und die Migrationskette" | ja – `make schema-validate` | Schema-Gate noch nicht verdrahtet |
| N-26 | INFO | `LH-QA-01` ist weiterhin „Vorläufiger, nicht freigegebener Abnahmezielwert"; Owner-Entscheidung steht aus (Vorgänger F-32). | LH-QA-01 | `spec/lastenheft.md` · „ohne Entscheidung keine fachliche Freigabe" | nein | Offene Owner-Entscheidung |
| N-27 | INFO | `.d-check.yml` führt die `LH-*`-ID-Klasse weiterhin auskommentiert; `make docs-check` ist grün, ohne ID-Abdeckung der Straten zu prüfen (Vorgänger F-35). N-20 wäre dort mechanisch aufgefallen. Zuständig: Planner. | Maintainability | `.d-check.yml` · „{regex: '<PREFIX>-[A-Z]{2}-\d{2}', target: spec/lastenheft.md, link-policy: always}" | ja – `make doc-complete` | Abdeckungs-Gate nicht aktiviert |
| N-28 | INFO | Der generierte Admin-Token liegt im Datenvolumen (`<DATA_DIR>/admin.token`) und wird nie geloggt; wer den Container frisch startet, muss die Datei aus dem Volume lesen, und Volume-Sicherungen enthalten das Token. Kein Spec-Fehler, aber im Betriebsprofil zu benennen. Zuständig: Architect. | LH-QA-04 | `spec/spezifikation.md` · „legt es dort mit Rechten `0600` ab und schreibt es nie in ein Log" | nein | Betriebshinweis Token-Ablage |

## Negativbefunde (geprüft, ohne Befund)

| Bereich | Ergebnis |
|---|---|
| `spec/lastenheft.md` → andere Straten (`direction: no-downward`) | geprüft, ohne Befund (kein `SPEC-`, `ARC-`, Spezifikations- oder Architektur-Verweis; `make docs-check` 0 Befunde) |
| `spec/spezifikation.md` → Architektur | geprüft, ohne Befund (kein `ARC-` oder `architecture`-Verweis) |
| `spec/architecture.md`: Wellen, Slices, ADR, Meilensteine, Historie, Sprachbezug (AGENTS.md §3.4) | geprüft, ohne Befund (Suchlauf; verbleibend nur Code-Modul-Pfade `cmd/`, `internal/`, `main.go`) |
| Lastenheft v0.6.0 gegen v0.5.0 (`git diff`) | geprüft, ohne Befund (nur Version, Datum, Historienzeile; Anforderungstext unverändert, Änderung ausdrücklich als Tatsachenberichtigung vor `Accepted` gekennzeichnet) |
| ID-Integrität Spezifikation | geprüft, ohne Befund (`SPEC-001`..`SPEC-073` je einmal, alle Verweise aufgelöst; alle im Text genutzten `ODR-E*` stehen in §4) |
| ID-Abdeckung `LH-FA-01..18`, `LH-QA-01..03,05..07`, `LH-RB-01..04` | geprüft, ohne Befund (je ein `.a`-Abschnitt; Ausnahme `LH-QA-04`, siehe N-20) |
| Lastenheft-Widersprüche der neuen Abschnitte (Restore, Reset, Fallback, Admin-Auth, Coverage) | geprüft, ohne Befund (kein Gegensatz zu `LH-FA-09`, `-10`, `-11`, `-12`, `-13`, `-18`) |
| Hard Rules §3.5/§3.6 (ADR-Immutabilität, Gate-Lockerung) | nicht anwendbar – keine ADRs vorhanden, keine Gate-Schwelle gesenkt |
| `harness/mk/schema.mk` (Kommentare nach AGENTS.md §3.7, Docker-only, netzlos, Container-Cleanup) | geprüft, ohne Befund (Kommentare tragen Zusage/Abgrenzung/Grenze; Digest-Pin; beide Targets in `make help`; Läufe oben) |
| `harness/README.md` §Sensors (zwei neue Zeilen) | geprüft, ohne Befund (Targets existieren, Bindung „kein Gate" deckt sich mit `schema.mk`; ersetzte Platzhalterzeilen) |
| `.harness/skills/reviewer.md` | geprüft, ohne Befund (mindestens zwei repo-spezifische HIGH-Regeln, Output-Schema vorhanden) |
| `docs/plan/adr/`, `docs/plan/planning/` | nicht anwendbar – keine ADRs, keine Slice-Pläne |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 15 |
| LOW | 8 |
| INFO | 5 |

Die 35 Vorgänger-Findings: 23 behoben, 8 teilweise behoben (mit Rest in N-Findings), 1 gemildert (F-29), 1 teilweise bestätigt (F-34), 2 offen bei Owner/Planner (F-32, F-35); keine unbehobenen HIGH.

**Finding-Klassen dieses Laufs:** Schema trägt Spec-Vertrag nicht (N-1, N-2, N-13, N-14, N-21) · Löschpfad nicht schemakonform spezifiziert · Replay-Cursor-Zustand unterbestimmt · Default-Policy sperrt Muss-Happy-Path (Rest) · Matching/Redaction-Wechselwirkung (Rest) · Fehlerklasse ohne Statusabbildung · Konfigurationsschlüssel ohne Katalogeintrag · Adressierung im Hybrid-Pfad fehlt · Port-Katalog und Importkanten inkonsistent (zweites Auftreten) · Redaction-Default unvollständig (Rest) · Atomarität über zwei Ports nicht festgelegt · Admin-Transportschutz nicht gefordert

## Verdikt

**Merge-blockierend:** nein für die Ablage des Arbeitsbaums als `Draft`; ja für den Übergang der Straten nach `Accepted` bzw. für den ersten Slice-Plan, solange die Schema-Findings N-1..N-3, N-13, N-14 und die Cursor-Mehrdeutigkeit N-4 offen sind, weil `SPEC-004` das Schema zur einzigen Quelle der Tabellen erklärt und sie aus dem Schema abgeleitet würden.

**Konflikt-Pfad:** Kein Implementer-Widerspruch bisher; kein Finding wurde herabgestuft. Die Kategorie von N-4 (MEDIUM statt HIGH) ist begründet: Mehrdeutigkeit ohne Verletzung eines Lastenheft-Kriteriums; sie würde HIGH, sobald eine Lesart den `LH-FA-04`-Boundary-Fall bricht.

**Übergabe:** Findings gehen an den Architect (Spec-Autor); N-25 und N-27 zusätzlich an Planner. Finding-Klassen gehen in die Slice-Closure §7 und den Zähler (zweites Auftreten: „Port-Katalog und Importkanten inkonsistent"). Der Report ersetzt keine Verifikation.

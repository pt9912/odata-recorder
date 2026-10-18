# Review-Report: spec-straten-review-6 — 2026-10-18

**Review-Art:** Design (sechster Lauf über die drei Spec-Straten, das Persistenzschema, die Schema-Werkzeuge und die Harness-Dokumente; kein Code, kein Slice-Plan, keine ADR)

**Gegenstand:** Stand `beed493` auf `main` (Arbeitsbaum sauber): `spec/lastenheft.md` (0.7.0), `spec/spezifikation.md`, `spec/architecture.md`, `tools/schema/schema.yaml` (0.8.0), `tools/schema/check.py`, `harness/mk/schema.mk`, `AGENTS.md`, `harness/README.md`, `harness/conventions.md`. Schwerpunkt: die Änderungen des letzten Commits `beed493` (Behebungen aus Review 5).

**Vorgänger:** `docs/reviews/2026-10-18-spec-straten-initial.md` (F-1..F-35), `...-nachreview.md` (N-1..N-28), `...-review-3.md` (R-1..R-19), `...-review-4.md` (S-1..S-20), `...-review-5.md` (T-1..T-17). Dieser Report nummeriert neu (U-1..). Bekannt und bewusst offen, nicht erneut gemeldet: Owner-Entscheidung zu `LH-QA-01` (Zielwert), `LH-*`-ID-Klasse in `.d-check.yml` nicht aktiviert, Migrationskette des Schemas, Toolchain/Sprache und Dockerfile.

**Skill:** `.harness/skills/reviewer.md` (Arbeitsbaum) · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-18

**Eingangs-Kontext:** die drei Straten, Schema und Schema-Werkzeuge, `AGENTS.md` (§3.1, §3.4, §3.7), `harness/README.md`, `harness/conventions.md`, die fünf Vorgänger-Reports, `git show beed493`.

**Sensor-Läufe dieses Reviews:**

- `make docs-check` → 27 Dateien, 0 Befunde.
- `make schema-check` → grün (alle Fälle `ok`, darunter „match_key null bei complete" abgewiesen, „match_key null bei failed" angenommen; `foreign_key_check`, `integrity_check`).

Nicht geprüft: DoD-/Abnahme-Konformität (Verifier), Bedarf gegen die Realität (Validator). Aussagen zur OData-Semantik nicht gegen die Primärquelle geprüft.

**Bewertung der Behebungen aus Review 5 (Stichprobe):** behoben und in sich stimmig: T-1 (Positivliste `$format`/`$count`/`$select`/`$orderby`; die Gegenprobe „Session nur mit Folgeseite" scheitert an `$skiptoken`, jede unbekannte oder herstellerspezifische Option schließt aus; der Weg zu `complete` ist nach eigener Prüfung geschlossen), T-2 (nur `ODR-E5001` löst Fallback aus; Sequenz 4.6 und Matrix stimmen überein), T-4 (`match_key` nullable nur bei `failed`: Spezifikation `SPEC-001`, Schema 0.8.0 und `check.py` stimmen überein), T-5 (Sequenzen 4.4/4.7: Serialisierung über den Codec-Port im Service), T-6 (a)–(d), T-10 (`src/` statt Sprachpfaden; keine Sprachbezüge mehr in den Straten), T-11, T-12, T-13, T-15, T-17. Teilweise behoben, Reste als neue Befunde: T-3 → U-2, T-7 → U-1, T-8 → U-3, T-9 → U-6, T-14 → U-5. Neu durch die Behebungen entstanden: U-1, U-4.

---

## Findings

HIGH: keines. Die HIGH-Regeln des Skills sind geprüft: kein Widerspruch zwischen den Straten in Richtung Lastenheft, keine neue Geheimnis-Senke, Reset/Restore/Cursor unverändert gemäß `LH-FA-09`/`LH-FA-10`, `complete`-Coverage nur über den Vollständigkeitsbeweis (siehe oben).

| ID | Kategorie | Quelle | Pfad | Befund | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| U-1 | MEDIUM | LH-FA-13, LH-QA-01, Maintainability (Importgraph) | `spec/architecture.md` · „adapters/driving → model, ports/inbound"; · „Ungültige Admin-Anfrage \| `ARC-007` \| 400/401/404/409/422, strukturierter Audit-Eintrag"; `spec/spezifikation.md` · „Quelle" `Gateway` in `SPEC-026` | Mit `AuditLog`, `Telemetry` und `IDGenerator` als Outbound Ports (Behebung zu T-7) bleibt offen, wie die Driving Adapter sie erreichen: Der Graph erlaubt ihnen nur `model` und `ports/inbound`. Der Admin-Adapter schreibt den Audit-Eintrag der abgewiesenen Anfrage (`ODR-E2101`, kein Use Case läuft), der HTTP-Adapter erzeugt Korrelations-ID (`LH-FA-01.a`) und den Gesamtspan `odata.gateway.request` (Messpunkt von `LH-QA-01.a`, endet beim ersten Response-Byte, das nur der Adapter kennt). Szenario: Der Implementer importiert die Outbound Ports im Adapter und bricht den Gate-Graphen, oder der Audit-Eintrag bei 401 und der Gateway-Span entfallen. | ja – später `make arch-check`; heute nein | Querschnittsbelang ohne erlaubten Zugriffsweg für Driving Adapter |
| U-2 | MEDIUM | LH-FA-02, LH-FA-13, Maintainability | `spec/spezifikation.md` · „wartet `close` auf dessen Abschluss"; `spec/architecture.md` · „S->>X: Read exchanges' capture status" | Das Beenden einer Session verlangt, dass Exchanges während `close` zu Ende laufen, neue abgewiesen und überfällige als `incomplete` geschrieben werden. Laufende Exchanges gehören dem Record-Service (kein Eintrag, solange sie laufen), `close` dem Session-Service; Services rufen einander nur über Inbound Ports, und keiner der Ports (`RecordUseCase`, `SessionUseCase`) trägt diese Koordination, der Zustand „beendet sich" hat keinen Ort (`sessions.status` kennt nur `open`/`complete`/`incomplete`). Sequenz 4.9 zeigt das Warten nicht. Szenario: Ein Exchange wird zwischen Statusprüfung und Statuswechsel angehängt; die Session steht `complete`, obwohl ein später angehängter Exchange `incomplete` ist. Ebenso unbelegt: Wer im Modus `record` beim Prozessstart die erste Session anlegt (`ARC-012` nennt nur `MaintenanceUseCase`). | ja – Integrationstest `close` unter laufendem Exchange | Koordination zwischen zwei Services ohne Port und Zustand |
| U-3 | LOW | LH-RB-04, Maintainability | `spec/architecture.md` · „nach Löschvorgängen die Blob-Bereinigung"; · „die Blob-Bereinigung folgt dem Commit über `BlobRepository`"; · „Integritätsprüfung" | Zuständigkeit der Bereinigung nach dem Löschen steht zweimal verschieden: `MaintenanceUseCase` bzw. der löschende Use Case über `BlobRepository`. Außerdem fehlt der Zugriff für die Aufgaben der Startwartung: Die DB-/FK-Prüfung (`SPEC-048`) hat keinen Outbound Port, und die Referenzmenge für den Blob-Abgleich (Blobs der Exchanges und Snapshots) ist keinem Port als Abfrage zugewiesen. | ja – Integrationstest Session löschen, Blob-Existenz und Start prüfen | Zusage ohne Träger in der Sicht (Rest T-8) |
| U-4 | LOW | Maintainability | `spec/architecture.md` · „Out --> Clock[Clock / ID Adapter]"; · „M->>M: Check baseline coverage and projection" | (a) `ARC-020` (Audit/Telemetry) steht in der Tabelle und der Schichtzeile, aber nicht im Komponentendiagramm (kein Knoten, keine `Boot`-Kante). (b) Sequenzen 4.4 und 4.7 prüfen Coverage vor der Query (`SPEC-044`), zeigen aber keinen Zugriff auf `CoverageRepository`. (c) `services/coverage` steht in Baum und `ARC-002`, hat aber keinen Inbound Port; `InspectionUseCase` liest Coverage, und Services importieren einander nicht. | nein (Sichtprüfung) | Diagramm/Baum decken die ARC-Tabelle nicht ab (2. Auftreten: S-15, U-4) |
| U-5 | LOW | LH-QA-05, Maintainability | `spec/spezifikation.md` · „\| POST \| `/__admin/sessions` \| Neue Recording-Session anlegen"; · „Ein Record-Request, für den der Service keine offene Session hat" | `ODR-E7010` steht in der Admin-Tabelle `SPEC-005` nicht an `POST /__admin/sessions`, `ODR-E5009` in keinem Algorithmus des Record-Pfads (`LH-FA-01.a`, `LH-FA-02.a`), nur in `LH-FA-13.a`, `SPEC-099` und `LH-QA-05.a`. Die Fehlerklasse ist damit am Auslöser nicht auffindbar. | ja – Contract-Test `POST` mit doppelter ID, Record nach `close` | Katalogcode ohne Nennung am Auslöser (3. Auftreten: S-13, T-14, U-5) |
| U-6 | LOW | LH-FA-11, LH-FA-18 | `spec/spezifikation.md` · „der Rest ist der OData-Pfad wie in `exchanges.path`" | Rest von T-9: Die Kettenauflösung bildet den Link-Pfad auf `exchanges.path` ab, ob `exchanges.path` der Pfad nach dem Service-Root (`LH-FA-04.c`: „Der Pfad nach dem Root ist der OData-Pfad") oder der Client-Pfad samt Root ist, definiert aber weder `SPEC-001` noch `LH-FA-02.a`. Relative `@odata.nextLink`-Werte (zulässig in OData V4) werden nicht behandelt; sie führen über „beginnt der Pfad nicht mit diesem Präfix" zum Abbruch. Folge ist sicher (Coverage bleibt `partial`), aber bei Root ungleich `/` kann der Weg zu `complete` unbenutzbar werden. | ja – Import-Test mit Root `/a` und Upstream-Basispfad | Vollständigkeitsbeweis ohne eindeutige Pfadbasis (Rest T-9) |
| U-7 | LOW | LH-QA-02, Maintainability | `spec/spezifikation.md` · „Zeit und serverseitig erzeugte IDs (Korrelations-ID, Workspace-ID, Snapshot-ID)" | Mit der Behebung zu T-3 vergibt der Prozess Session-IDs serverseitig beim Start (`LH-FA-13.a`); die Aufzählung in `LH-QA-02.a` nennt Session-ID und Exchange-ID nicht. Ob sie aus `IDGenerator` stammen, ist offen, obwohl `LH-QA-02` gleiche Recording-Session voraussetzt. | ja – Determinismus-Test (`make test-determinism`, noch nicht vorhanden) | Quellenliste serverseitiger IDs nach Ergänzung unvollständig |
| U-8 | INFO | LH-FA-12, Maintainability | `spec/spezifikation.md` · „Nur `ODR-E5001` (kein Exchange mit diesem Match-Key) darf einen Fallback auslösen"; · „Nur gesonderte GET/HEAD-Allowlist (`FALLBACK_ALLOW_HOSTS`)" | Nur Formulierung, kein Verhaltensbefund: (a) Der Satz gilt für Replay-Fehler, steht aber wörtlich neben Mock-Zeilen (`ODR-E6001`/`ODR-E6003`), die Fallback auslösen. (b) Die „anderen Replay-Fehler" nennen `ODR-E3001` nicht. (c) „gesondert" nennt eine eigene Liste, `SPEC-062` führt nur eine; ob `FALLBACK_WRITE_ALLOW` den Weg nach einem Replay-Miss öffnet, sagt die Matrix mit „nur GET/HEAD" nein, Sequenz 4.6 unterscheidet es nicht. Zuständig: Architect, beim nächsten Spec-Schnitt. | nein | Formulierung des Fallback-Auslösers |
| U-9 | INFO | Reviewer-Skill (§Pflege, Steering-Loop) | `docs/reviews/*.md` | Schwelle erreicht: „Port-Katalog und Importkanten inkonsistent" ist mit U-1/U-3/U-4 das sechste Auftreten (F-14, N-10, R-1, S-1/S-2, T-6, U-1), „Katalogcode ohne Nennung am Auslöser" das dritte (S-13, T-14, U-5). Ein `make arch-check` (`SPEC-039`) bzw. ein Katalog-Abgleich Codes ↔ Auslöser würde beide fangen, ist aber noch kein Gate. Zuständig: Architect/Planner für die Folge-Entscheidung; keine Aktion des Implementers. | nein | Steering-Loop-Schwelle |

---

## Negativbefunde (geprüft, ohne Befund)

| Bereich | Ergebnis |
|---|---|
| `spec/lastenheft.md` 0.7.0 → andere Straten (`no-downward`) | geprüft, ohne Befund (`make docs-check` 0 Befunde) |
| Lastenheft ↔ Spezifikation: Reset/Restore/Cursor, Betriebsarten, Fallback-Defaults, Redaction, Admin-Authentisierung, Sessions (`LH-FA-13`) | geprüft, ohne Befund bei Richtung und Umfang |
| Vollständigkeitsbeweis (Positivliste, Kettenauflösung, `key_redacted`, leere Kette) gegen `LH-FA-11`/`LH-FA-18` und die Skill-HIGH-Regel | geprüft, ohne Befund bei der Zusage; Pfadbasis U-6 |
| Hybrid-Entscheidungsmatrix gegen Sequenz 4.6 und Modus-Routing 2.5 | geprüft, ohne Befund bei Auslösern und Verboten; Formulierung U-8 |
| `match_key`-Nullbarkeit: `SPEC-001`, `SPEC-004`, `schema.yaml` (`ck_match_key_null_failed`), `check.py` | geprüft, ohne Befund (per `make schema-check` bestätigt) |
| Spezifikation ↔ `tools/schema/schema.yaml` (Tabellen, Enumerationen, Zustandsinvarianten, FKs, `ON DELETE`, Schemaversion 0.8.0) | geprüft, ohne Befund |
| `harness/mk/schema.mk` | geprüft, ohne Befund (unverändert) |
| `spec/architecture.md`: Wellen, Slices, ADR-Bezüge, Sprachbezüge (AGENTS.md §3.4) | geprüft, ohne Befund |
| `spec/architecture.md`: ARC-Tabelle ↔ Port-Katalog (`ARC-004` ↔ 2.3), Service-Liste ↔ Baum, Restore im `WorkspaceUseCase` (4.5), `MaintenanceUseCase` (4.9), Session-Zustand 2.4 | geprüft, ohne Befund bei Namen und Zuordnung; Reste U-1..U-4 |
| `AGENTS.md`, `harness/README.md`, `harness/conventions.md` | geprüft, ohne Befund (keine Platzhalter mehr, keine Aussage über nicht vorhandene Targets) |
| Hard Rules §3.2, §3.3, §3.5, §3.6 | nicht anwendbar bzw. ohne Befund – keine ADRs, keine Gate-Schwelle gesenkt |
| `docs/plan/adr/`, `docs/plan/planning/` | nicht anwendbar – keine ADRs, keine Slice-Pläne |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 5 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Querschnittsbelang ohne erlaubten Zugriffsweg für Driving Adapter · Koordination zwischen zwei Services ohne Port und Zustand · Zusage ohne Träger in der Sicht (Rest T-8) · Diagramm/Baum decken die ARC-Tabelle nicht ab · Katalogcode ohne Nennung am Auslöser (3. Auftreten) · Vollständigkeitsbeweis ohne eindeutige Pfadbasis (Rest T-9) · Quellenliste serverseitiger IDs unvollständig · Formulierung des Fallback-Auslösers · Steering-Loop-Schwelle

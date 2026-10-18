# Review-Report: spec-straten-review-7 — 2026-10-18

**Review-Art:** Design (siebter Lauf über die drei Spec-Straten, das Persistenzschema, die Schema-Werkzeuge und die Harness-Dokumente; kein Code, kein Slice-Plan, keine ADR)

**Gegenstand:** Stand `019e86e` auf `main` (Arbeitsbaum sauber): `spec/lastenheft.md` (0.7.0), `spec/spezifikation.md`, `spec/architecture.md`, `tools/schema/schema.yaml` (0.8.0), `tools/schema/check.py`, `harness/mk/schema.mk`, `AGENTS.md`, `harness/README.md`, `harness/conventions.md`. Schwerpunkt: die Änderungen des Fix-Commits `8510b9a` (Behebungen aus Review 6); der Commit `019e86e` (eine Zeile pro Satz) wurde nur auf beschädigte Markdown-Struktur geprüft.

**Vorgänger:** `docs/reviews/2026-10-18-spec-straten-initial.md` (F-1..F-35), `...-nachreview.md` (N-1..N-28), `...-review-3.md` (R-1..R-19), `...-review-4.md` (S-1..S-20), `...-review-5.md` (T-1..T-17), `...-review-6.md` (U-1..U-9). Dieser Report nummeriert neu (V-1..). Bekannt und bewusst offen, nicht erneut gemeldet: Owner-Entscheidung zu `LH-QA-01` (Zielwert), `LH-*`-ID-Klasse in `.d-check.yml` nicht aktiviert, Migrationskette des Schemas, Toolchain/Sprache und Dockerfile, fehlendes `make arch-check`.

**Skill:** `.harness/skills/reviewer.md` (Arbeitsbaum) · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-18

**Eingangs-Kontext:** die drei Straten, Schema und Schema-Werkzeuge, `AGENTS.md` (§3.4, §3.7), die sechs Vorgänger-Reports, `git show 8510b9a`.

**Sensor-Läufe dieses Reviews:**

- `make docs-check` → 28 Dateien, 0 Befunde.
- `make schema-check` → grün (`foreign_key_check`, `integrity_check` ok).
- Struktur-Skript über die Umformatierung (Tabellen, Codeblöcke, Listen, Backtick-/Fettdruck-Paarung je Zeile): Codeblöcke balanciert, keine unterbrochene Tabelle, keine verwaiste Listeneinrückung; ein Treffer (V-6).

Nicht geprüft: DoD-/Abnahme-Konformität (Verifier), Bedarf gegen die Realität (Validator). Aussagen zur OData-Semantik nicht gegen die Primärquelle geprüft.

**Bewertung der Behebungen aus Review 6 (Stichprobe):** behoben und in sich stimmig: U-1 (`ports/support` mit `AuditLog`, `Telemetry`, `Clock`, `IDGenerator`; `ARC-021` in Tabelle, Schichtzeilen und Importgraph; Driving und Driven Adapter dürfen sie importieren), U-2 im Kern (`RecordUseCase.Drain`, Sequenz 4.9, Session-Zustand 2.4), U-3 (`StorageMaintenance` in `ARC-004` und 2.3, Zuständigkeit der Bereinigung einheitlich bei `MaintenanceUseCase`), U-4 (a)–(c) (`ARC-020` im Diagramm, Coverage-Zugriff in 4.4/4.7, `services/coverage` entfernt, Baum und Service-Liste stimmen überein), U-5 (`ODR-E7010` an `POST /__admin/sessions`, `ODR-E5009` in `LH-FA-01.a` und `LH-FA-02.a`), U-6 (relative Links, Pfad „nach dem Service-Root"), U-7 (Session-/Exchange-ID aus `IDGenerator`), U-8 (Fallback-Wortlaut, `ODR-E3001` ergänzt). Schema und Spezifikation unverändert konsistent (`make schema-check`). Neu durch die Behebungen entstanden: V-1 bis V-5.

---

## Findings

HIGH: keines. Die HIGH-Regeln des Skills sind geprüft: kein Widerspruch zu `spec/lastenheft.md`, keine neue Geheimnis-Senke (die neuen Querschnitts-Ports tragen nach Beschreibung nur Audit-Einträge ohne Geheimnisse), Reset/Restore/Cursor unverändert, `complete`-Coverage nur über den Vollständigkeitsbeweis, keine Wellen/Slices/ADR-/Sprachbezüge in der Architektur (AGENTS.md §3.4).

| ID | Kategorie | Quelle | Pfad | Befund | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| V-1 | MEDIUM | Maintainability (Wiederholung: S-15, U-4) | `spec/architecture.md` · „Out --> Clock[Clock / ID Adapter]" | Das Komponentendiagramm hängt `Clock` an `Out` (Outbound Ports), obwohl `ARC-021`, Abschnitt 2.3 und der Importgraph `Clock`/`IDGenerator` als Querschnitts-Ports (`ports/support`) führen. Außerdem zeigt es keine Kante der Driving Adapter zu `Sup`, die Schichtzeilen und Graph erlauben. Szenario: Der Implementer legt `Clock` nach dem Diagramm unter `ports/outbound`; der HTTP-Adapter (Korrelations-ID, Span-Zeitstempel) darf es dort nicht importieren, und der Graph-Check rot oder die Kante wird nachträglich gelockert. Dritte Wiederholung des Musters „Diagramm deckt ARC-Tabelle nicht ab". | nein (Sichtprüfung); später `make arch-check` | Diagramm/Baum decken die ARC-Tabelle nicht ab (3. Auftreten) |
| V-2 | LOW | Maintainability (`AGENTS.md` §3.7 nicht einschlägig; Sicht-Regel „Services importieren einander nicht") | `spec/architecture.md` · „im Modus `record` die erste Recording-Session über `SessionUseCase` anlegen"; · „löschende Use Cases rufen die Blob-Bereinigung nach dem Commit hier auf" | `MaintenanceUseCase` ruft `SessionUseCase` (Session anlegen, Aufbewahrung), und `SessionUseCase` ruft beim Löschen `MaintenanceUseCase` (Blob-Bereinigung): eine zyklische Abhängigkeit zweier Services über Inbound Ports. Importgraph bleibt grün, aber die Verdrahtung im Composition Root braucht verzögerte Bindung oder einen Zyklus-Bruch, den die Sicht nicht benennt. | nein | Zyklus zwischen zwei Use Cases über Inbound Ports |
| V-3 | LOW | LH-FA-01, Maintainability | `spec/spezifikation.md` · „Im Modus `record` wählt das Gateway außerdem die offene Recording-Session des Service"; `spec/architecture.md` · „H->>R: Record(request)" | „Gateway" ist als Komponente in der Sicht nicht definiert; Service-Auflösung (`ODR-E5005`) und Session-Wahl (`ODR-E5003`, `ODR-E5009`) sind keiner Komponente zugewiesen. Die Sicht sagt, der HTTP-Adapter kenne den Modus nicht (2.5) und entscheide nicht fachlich (2.2), Sequenz 4.1 zeigt keinen Zugriff auf `SessionRepository`. Damit ist offen, ob Adapter oder `RecordUseCase` die Session wählt. | ja – Contract-Test Record ohne offene Session | Fachliche Entscheidung ohne zugewiesene Komponente |
| V-4 | LOW | LH-FA-02, Maintainability | `spec/spezifikation.md` · „Mit dem Beginn von `close` nimmt die Session keine neuen Exchanges an"; `spec/architecture.md` · „neue Exchanges abweisen" | (a) Der Abweis neuer Exchanges während `close`/`Drain` hat keinen Fehlercode: `ODR-E5009` heißt „ohne offene Session", die Session steht in dieser Zeit aber noch `open`; der Shutdown-Pfad nennt HTTP 503 (`LH-FA-02.c`). (b) Der geordnete Shutdown (Sessions → `complete`) hat keinen benannten Port-Aufruf: Sequenz 4.9 zeigt nur `close` über die Admin API, Abschnitt 5 nur „Bootstrap und CLI orchestrieren". | ja – Integrationstest Request während `close` | Zustandsübergang ohne Fehlercode und Port-Aufruf |
| V-5 | LOW | LH-FA-02, Maintainability | `spec/architecture.md` · „im Modus `record` die erste Recording-Session"; `spec/spezifikation.md` · „entsteht im Modus `record` beim Start des Prozesses" | „erste Session" (Sicht) und „beim Start eine Session" (Spezifikation) lassen offen, ob je Start und je konfiguriertem Service eine Session angelegt wird; Record-Requests brauchen eine offene Session je Service (`ODR-E5009`), die Service-Datei erlaubt mehrere Services. Szenario: Zwei Services, eine Startsession: Requests des zweiten Services ergeben `ODR-E5009`. | ja – Start-Test mit zwei Services | Startverhalten bei mehreren Services nicht festgelegt |
| V-6 | INFO | Maintainability (Umformatierung `019e86e`) | `spec/architecture.md` · „**Reset vs." | Die Satz-pro-Zeile-Umformatierung hat „vs." als Satzende gewertet und den Fettdruck `**Reset vs. Restore:**` über zwei Zeilen geteilt; gerendert bleibt er intakt, die Quelle ist an dieser Stelle irritierend. Sonst keine Beschädigung (Tabellen, Codeblöcke, Listen intakt). Zuständig: Implementer bei der nächsten Spec-Berührung. | nein | Umformatierung trennt an Abkürzung |

---

## Negativbefunde (geprüft, ohne Befund)

| Bereich | Ergebnis |
|---|---|
| `spec/lastenheft.md` 0.7.0 → andere Straten (`no-downward`) | geprüft, ohne Befund (`make docs-check` 0 Befunde) |
| Lastenheft ↔ Spezifikation: Reset/Restore/Cursor, Fallback-Defaults, Redaction, Admin-Authentisierung, Sessions | geprüft, ohne Befund bei Richtung und Umfang; `ODR-E5009`/`ODR-E7010`/`SHUTDOWN_TIMEOUT` stehen nur in der Spezifikation und widersprechen dem Lastenheft nicht |
| Vollständigkeitsbeweis inkl. relativer `nextLink` gegen `LH-FA-11`/`LH-FA-18` | geprüft, ohne Befund |
| Hybrid-Matrix und Fallback-Wortlaut gegen Sequenz 4.6, Modus-Routing 2.5 | geprüft, ohne Befund |
| Fehlercode-Katalog (`SPEC-099`, `SPEC-100`, `LH-QA-05.a`, `SPEC-005`) | geprüft, ohne Befund; Codes stehen am Auslöser, keine Dubletten bei SPEC-IDs |
| Spezifikation ↔ `tools/schema/schema.yaml` / `check.py` / `harness/mk/schema.mk` | geprüft, ohne Befund (`make schema-check` grün; Fix-Commit änderte das Schema nicht) |
| `spec/architecture.md`: ARC-Tabelle ↔ Port-Katalog (`ARC-004` ↔ 2.3), Service-Liste ↔ Baum ↔ Inbound-Ports, Importgraph ↔ Schichttabelle | geprüft, ohne Befund bei Namen und Kanten; V-1 betrifft nur das Diagramm |
| `spec/architecture.md`: Sequenzen 4.4, 4.7, 4.9 gegen Ports und Services | geprüft, ohne Befund bei Teilnehmern und Aufrufen; Lücken V-3, V-4 |
| `spec/architecture.md`: Wellen, Slices, ADR-Bezüge, Sprachbezüge (AGENTS.md §3.4) | geprüft, ohne Befund |
| Umformatierung `019e86e`: Tabellen, Codeblöcke/Mermaid, Listen, Fettdruck/Backticks | geprüft, ohne Befund außer V-6 |
| `AGENTS.md`, `harness/README.md`, `harness/conventions.md` | geprüft, ohne Befund (unverändert; bekannte offene Platzhalter-Sektionen sind nicht Gegenstand) |
| `docs/plan/adr/`, `docs/plan/planning/` | nicht anwendbar – keine ADRs, keine Slice-Pläne |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 4 |
| INFO | 1 |

**Finding-Klassen dieses Laufs:** Diagramm/Baum decken die ARC-Tabelle nicht ab (3. Auftreten, daher MEDIUM; Steering-Loop-Schwelle erreicht, Folge-Entscheidung beim Architect: ein Diagramm-Abgleich oder `make arch-check`) · Zyklus zwischen zwei Use Cases über Inbound Ports · Fachliche Entscheidung ohne zugewiesene Komponente · Zustandsübergang ohne Fehlercode und Port-Aufruf · Startverhalten bei mehreren Services nicht festgelegt · Umformatierung trennt an Abkürzung

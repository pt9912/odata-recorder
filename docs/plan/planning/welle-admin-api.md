# Welle welle-admin-api: Admin-API, Authentisierung und CLI-Kommandos

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); bis ihr Trigger eintritt steht sie in der
Roadmap unter *Nächste Wellen*. Ihre Slice-Dateien legt `plan-welle` (Schritt 6)
beim Start an, weil ein Slice-Plan vor seinem Start veraltet.

**Zielmeilenstein:** M6 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Ein Test steuert Aufzeichnung, Import, Workspace und Snapshot ausschließlich über den Admin-Listener: Session anlegen, über den OData-Listener aufzeichnen, beenden, importieren, Workspace anlegen, im Modus `mock` ändern, Snapshot erzeugen, zurücksetzen und wiederherstellen; Anfragen ohne oder mit falschem Token weist der Listener mit `ODR-E2101` ab, ohne Zustand zu ändern, und er ist standardmäßig nur über Loopback oder mit TLS erreichbar (`LH-FA-13`, `LH-QA-04`, `LH-FA-18` Import-Autorisierung, `LH-FA-09`, `LH-FA-10`).
Die Welle liefert die Endpunkte von `SPEC-005`, die Authentisierung und den Transportschutz von `LH-FA-13.a`, `InspectionUseCase` und die CLI-Kommandos über Start und Version hinaus.
Eine Welle, weil der Closure-Trigger den Rundlauf über HTTP und den Sicherheitsnachweis aus Canary-Secrets und Netzwerkzugriff beobachtet, den keine DoD eines einzelnen Endpunkt-Slice trägt.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-betrieb` liegt in `done/`: `AuditLog` (jeder Schreibendpunkt wird auditiert, `SPEC-005`), `Telemetry` (Zähler je Fehlercode für `/__admin/diagnostics`) und der TLS-Baustein des Listeners sind die Grundlage des Admin-Listeners.
- M5 ist erreicht: der Lasttest nach `LH-QA-01.a` läuft reproduzierbar und weist p95 aus dem Gesamtspan aus.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Der M6-Test aus `slice-bootstrap-m6` ist grün und Teil von `make test`: der Rundlauf aus §1 ausschließlich über Admin-Listener und OData-Listener, jeder Schreibendpunkt hinterlässt einen Audit-Eintrag, und `GET /healthz` antwortet ohne Token.
- Der Sicherheitsnachweis aus `slice-test-security` ist grün und Teil der Gate-Kette: Canary-Secrets in Headern, Query-Parametern und Body-Feldern erscheinen in keinem Speicher- und Log-Artefakt, jeder Admin-Endpunkt außer `GET /healthz` weist ohne Token mit `ODR-E2101` ab, und der Admin-Listener ist nur über Loopback oder mit TLS erreichbar (`LH-QA-04.a`).
- Ein nicht autorisierter Client, der einen Import anfordert, wird ohne Änderung abgewiesen (`LH-FA-18`).
- Die Closure-Notiz liegt als `welle-admin-api-results.md` im Ruheort `done/`; ihre *Steering-Loop-Einträge* tragen je verkörperter Regel mit dem Anker `seit welle-admin-api` das Feld `liegt in`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-inspection-use-case | `InspectionUseCase`: Exchanges, Coverage und Diagnose lesen, paginiert, ohne Geheimnisse | `LH-FA-13`, `LH-FA-11`; `SPEC-005` |
| slice-admin-listener-token | Admin-Listener (Loopback als Default) und Token-Authentisierung: `ADMIN_TOKEN_FILE`, zeitkonstanter Vergleich, `ODR-E2101`, `GET /healthz` ohne Token | `LH-FA-13`, `LH-QA-04`; `LH-FA-13.a`, `SPEC-056`, `SPEC-064` |
| slice-admin-tls | Transportschutz des Admin-Listeners: `ADMIN_TLS_*`, `ADMIN_ALLOW_PLAINTEXT`, Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` | `LH-FA-13`, `LH-RB-02`; `SPEC-089`, `SPEC-090` |
| slice-admin-sessions | Endpunkte für Sessions, Exchanges, Coverage und Diagnose | `LH-FA-13`, `LH-FA-02`; `SPEC-005`, `SPEC-124` |
| slice-admin-imports | Endpunkte für Imports (idempotent, Statuscodes), Import-Autorisierung | `LH-FA-18`, `LH-FA-13`; `SPEC-047` |
| slice-admin-workspaces | Endpunkte für Workspaces, Reset, Snapshots und Restore | `LH-FA-09`, `LH-FA-10`, `LH-FA-13`; `SPEC-005` |
| slice-cli-kommandos | CLI-Kommandos über Start und Version hinaus (Umfang und Form legt der Architect fest) | `LH-FA-13`; `ARC-006` |
| slice-test-security | Sicherheitsnachweis: Canary-Secrets in allen Artefakten, Netzwerkzugriff auf Admin-Endpunkte | `LH-QA-04`, `LH-RB-04`; `LH-QA-04.a` |
| slice-bootstrap-m6 | Composition Root für den Admin-Listener und M6-Test (Rundlauf über HTTP) | `LH-FA-13`, `LH-FA-18`; `SPEC-110` |

Die Slice-Dateien legt `plan-welle` (Schritt 6) beim Start der Welle an: Ein Slice-Plan altert vor seinem Start, und die Pläne dieser Welle hängen an den Verträgen von `welle-betrieb` (Audit, Telemetry, TLS).

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: keine Welle; sie schließt die Muss-Anforderungen für M7 (Erstabnahme).
- Wird blockiert von: `welle-betrieb` (`AuditLog`, `Telemetry`, TLS-Baustein), mittelbar `welle-mock` und `welle-record-replay`.
- Unabhängig: `welle-hybrid`, `welle-etag`, `welle-expand-batch`, `welle-odata-v2`.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-admin-listener-token` → `slice-admin-tls`; `slice-inspection-use-case` unabhängig davon; `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` brauchen `slice-admin-listener-token` und `slice-inspection-use-case` (Sessions), je ihren Use Case aus `welle-record-replay` und `welle-mock`; `slice-cli-kommandos` nach den Endpunkt-Slices (Form nach Entscheidung des Architect); `slice-test-security` braucht Listener, Token und TLS; `slice-bootstrap-m6` zuletzt.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes der Use Cases hinter dem Admin-Adapter; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für die Zweige der Authentisierung und der Startprüfung; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Ein Risiko des Vertrags: „der Vergleich erfolgt zeitkonstant“ (`LH-FA-13.a`) hat keinen Fänger, der aus dem benannten Grund rot wird; der Plan führt es als akzeptiertes Negativ mit Grund und Folge oder lässt den Architect einen Fänger nennen (`AGENTS.md` §3.15). Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Grafische Oberfläche — ausgeschlossen im Lastenheft (`LH-FA-13`, §5).
- Umschalten des Betriebsmodus zur Laufzeit — nicht Teil des ersten Abnahmeumfangs (`LH-FA-13`).
- Authentisierung über das Bearer-Token hinaus (Benutzer, Rollen, OIDC) — die Spezifikation kennt ein Token (`LH-FA-13.a`).
- Die OpenAPI-Datei zu `SPEC-005` — der Architect erstellt sie vor der Implementierung; sie ist Voraussetzung, nicht Ergebnis der Welle.
- Neue Fachlogik in den Endpunkten — der Admin-Adapter ruft Inbound Ports und schreibt nie direkt in die Datenbank (Architektur §2).
- Export und Abruf der Telemetrie — `welle-betrieb`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-admin-api-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).

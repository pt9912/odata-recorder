# Welle welle-admin-api: Admin-API und Authentisierung

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); bis ihr Trigger eintritt steht sie in der
Roadmap unter *Nächste Wellen*. Ihre Slice-Dateien liegen seit
2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten.

**Zielmeilenstein:** M6 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Ein Test steuert Aufzeichnung, Import, Workspace und Snapshot ausschließlich über den Admin-Listener: Session anlegen, über den OData-Listener aufzeichnen, beenden, importieren, Workspace anlegen, im Modus `mock` ändern, Snapshot erzeugen, zurücksetzen und wiederherstellen; Anfragen ohne oder mit falschem Token weist der Listener mit `ODR-E2101` ab, ohne Zustand zu ändern, und er ist standardmäßig nur über Loopback oder mit TLS erreichbar (`LH-FA-13`, `LH-QA-04`, `LH-FA-18` Import-Autorisierung, `LH-FA-09`, `LH-FA-10`).
Die Welle liefert die Endpunkte von `SPEC-005`, die Authentisierung und den Transportschutz von `LH-FA-13.a` und `InspectionUseCase`.
Eine Welle, weil der Closure-Trigger den Rundlauf über HTTP und den Sicherheitsnachweis aus Canary-Secrets und Netzwerkzugriff beobachtet, den keine DoD eines einzelnen Endpunkt-Slice trägt.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-betrieb` liegt in `done/`: `AuditLog` (jeder Schreibendpunkt wird auditiert, `SPEC-005`), `Telemetry` (Zähler je Fehlercode für `/__admin/diagnostics`) und der TLS-Baustein des Listeners sind die Grundlage des Admin-Listeners.
- M5 ist erreicht: der Lasttest nach `LH-QA-01.a` läuft reproduzierbar und weist p95 aus dem Gesamtspan aus.

**Der Trigger ist Priorisierung, keine Code-Abhängigkeit** (Entscheidung des Users zu Betrieb vor Admin, [Abgleichsbericht](../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §10 Nr. 1): `slice-lasttest` und M5 liefern dem Code der Admin-API nichts. Code-Abhängigkeiten von `welle-betrieb` haben nur vier Slices: `slice-audit-log`, `slice-telemetry-port`, `slice-telemetry-einbau-gateway` und `slice-listener-tls`; die Pläne tragen sie als Start-Trigger (§5).

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Der M6-Test aus `slice-bootstrap-m6` ist grün und Teil von `make test`: der Rundlauf aus §1 ausschließlich über Admin-Listener und OData-Listener, jeder Schreibendpunkt hinterlässt einen Audit-Eintrag, und `GET /healthz` antwortet ohne Token.
- Der Sicherheitsnachweis aus `slice-test-security` ist grün und Teil der Gate-Kette: Canary-Secrets in Headern, Query-Parametern und Body-Feldern erscheinen in keinem Speicher- und Log-Artefakt, jeder Admin-Endpunkt außer `GET /healthz` weist ohne Token mit `ODR-E2101` ab, und der Admin-Listener ist nur über Loopback oder mit TLS erreichbar (`LH-QA-04.a`).
- Ein nicht autorisierter Client, der einen Import anfordert, wird ohne Änderung abgewiesen (`LH-FA-18`).
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (Admin-Listener, Token, Schlüssel `ADMIN_*`, Endpunkte, Fehlercodes); das Kataloge-Gate aus `slice-handbuch-meldungskatalog` ist grün.
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
| slice-test-security | Sicherheitsnachweis: Canary-Secrets in allen Artefakten, Netzwerkzugriff auf Admin-Endpunkte | `LH-QA-04`, `LH-RB-04`; `LH-QA-04.a` |
| slice-bootstrap-m6 | Composition Root für den Admin-Listener und M6-Test (Rundlauf über HTTP) | `LH-FA-13`, `LH-FA-18`; `SPEC-110` |

Die Slice-Dateien liegen seit 2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten, und `plan-welle` gleicht sie beim Start gegen den Ist-Stand ab. Marken und Betriebsregeln setzt der Planner beim Start (`SPEC-128`).

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. Das Binary ändert sich in dieser Welle erst mit `slice-bootstrap-m6` (der Admin-Listener erreicht es vorher nicht): verbindlich trägt den Punkt dieser Slice (Endpunkte, `ADMIN_*`, Token, `ODR-E2101`), die übrigen Slices tragen ihn bedingt, nämlich nur, wenn sie einen Fehlercode, Schlüssel oder ein Kommando in den Katalog bringen. Kommandos über Start und Version hinaus liefert die Welle nicht (§6), also auch keinen Eintrag im Kommando-Katalog.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: keine Welle; sie schließt die Muss-Anforderungen für M7 (Erstabnahme).
- Wird blockiert von: `welle-betrieb`, konkret von `slice-audit-log` (`AuditLog`), `slice-telemetry-port` (`Telemetry`, Zähler der Diagnose), `slice-telemetry-einbau-gateway` (Voraussetzung von `slice-test-security`) und `slice-listener-tls` (TLS-Baustein, nur `slice-admin-tls`); der Trigger „M5 erreicht“ (§2) ist Priorisierung. Mittelbar von `welle-mock` (`slice-snapshot-use-case`, `slice-bootstrap-m3`) und `welle-record-replay` (`slice-workspace-use-case`, `slice-bootstrap-m2`).
- Unabhängig: `welle-hybrid`, `welle-etag`, `welle-expand-batch`, `welle-odata-v2`.
- Reihenfolge der Slices (Pfeil heißt: braucht; `slice-audit-log` aus `welle-betrieb`):
  - `slice-admin-listener-token` braucht `slice-audit-log`; `slice-admin-tls` braucht `slice-admin-listener-token` und `slice-listener-tls`.
  - `slice-inspection-use-case` braucht `slice-telemetry-port` und sonst keinen Slice dieser Welle.
  - Die drei Endpunkt-Slices brauchen alle `slice-admin-listener-token` und `slice-audit-log`; `slice-admin-sessions` zusätzlich `slice-inspection-use-case`; `slice-admin-imports` hat keine weitere Kante (Rahmen und `AuditLog` genügen, `slice-inspection-use-case` nicht); `slice-admin-workspaces` zusätzlich `slice-workspace-use-case` und `slice-snapshot-use-case`.
  - `slice-test-security` braucht die drei Endpunkt-Slices, `slice-admin-listener-token`, `slice-admin-tls`, `slice-bootstrap-m2`, `slice-audit-log` und `slice-telemetry-einbau-gateway`.
  - `slice-bootstrap-m6` braucht `slice-bootstrap-m2`, `slice-bootstrap-m3`, `slice-telemetry-port`, `slice-audit-log`, `slice-handbuch-meldungskatalog` und alle übrigen Slices der Welle; er ändert `src/bootstrap/` und `SPEC-110` Nr. 2 wie `slice-bootstrap-m4` von `welle-hybrid` und läuft nicht zugleich mit ihm.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes der Use Cases hinter dem Admin-Adapter; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für die Zweige der Authentisierung und der Startprüfung; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Ein Risiko des Vertrags: „der Vergleich erfolgt zeitkonstant“ (`LH-FA-13.a`) hat keinen Fänger, der aus dem benannten Grund rot wird; der Plan führt es als akzeptiertes Negativ mit Grund und Folge oder lässt den Architect einen Fänger nennen (`AGENTS.md` §3.15). Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Grafische Oberfläche — ausgeschlossen im Lastenheft (`LH-FA-13`, §5).
- CLI-Kommandos über Start und Version hinaus — entfallen (Entscheidung des Users, 2026-10-30): `LH-FA-13` verlangt eine Verwaltungsschnittstelle, keine Kommandos; die Spezifikation hat keinen Vertrag dazu, `ARC-006` beschreibt nur Konfiguration, Kommandoauswahl und Start- und Exit-Verhalten. Verlangt der Owner sie später, ist das ein Change Request am Lastenheft vor `Accepted`, und der Planner legt den Slice dann neu an.
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

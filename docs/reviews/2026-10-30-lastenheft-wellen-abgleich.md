# Lastenheft gegen Wellen: Abgleich und Schnitt der Produktwellen

**Datum:** 2026-10-30 · **Rolle:** Planner · **Gegenstand:** `spec/lastenheft.md` Version 0.7.0, Roadmap, Welle-Pläne.

## 1. Zweck

Der User fragte, ob alle Wellen angelegt sind und ob aus Lastenheft und Pflichtenheft nichts mehr offen ist. Dieser Bericht gleicht jede Anforderung des Lastenhefts gegen die Wellen ab, benennt die Anforderungen ohne Welle, begründet den Schnitt der neuen Produktwellen und hält fest, was der Planner nicht entscheiden kann.

## 2. Methode

- Lastenheft §3 bis §5 vollständig gelesen; Spezifikation (`LH-*.a`-Abschnitte, `SPEC-005`, `SPEC-047`, `SPEC-049`, `SPEC-110`, Tabellen §3 bis §5), Architektur (Komponenten, Ports, Modus-Routing, Sequenzen 4.4 bis 4.6, Erweiterungspunkte), die ADRs, Roadmap und die Pläne der Wellen `welle-fundament`, `welle-persistenz`, `welle-record-replay` samt Slices in `done/` und `open/`.
- Die Out-of-Scope-Listen der `done/`- und `open/`-Slices wurden gelesen, weil jede dort genannte „Folgewelle“ eine Sendung ist, die eine Adresse braucht (`AGENTS.md` §3.13).
- Das Beobachtungs-Register wurde gesichtet (siehe §9).
- Nicht gelesen und nicht berührt: die Worktrees der Wellen `welle-dev-container`, `welle-benutzerhandbuch`, `welle-sensoren`. Ihre Inhalte sind nach der Beauftragung eingesetzt (`LH-RB-03` gehört zu `welle-dev-container`, die Coverage-Schwelle zu `welle-sensoren`).
- Bezeichner: „geliefert“ heißt, die Slices dazu liegen in `done/`; „in Arbeit“ heißt, die Welle `welle-record-replay` führt den Slice, er liegt aber noch in `open/` oder `in-progress/`.

## 3. Tabelle: Anforderung zu Welle

| Anforderung | Welle | Stand |
|---|---|---|
| `LH-FA-01` Weiterleitung | `welle-record-replay` (`slice-record-service`, `slice-upstream-adapter` geliefert; `slice-http-adapter` offen); Diagnose als strukturiertes Log: `welle-betrieb` | in Arbeit, Rest geplant |
| `LH-FA-02` Aufzeichnung | `welle-record-replay` (Capture, Record-Dienst, Sessions geliefert) | geliefert bis auf HTTP-Eingang |
| `LH-FA-03` Persistenz | `welle-persistenz`, `welle-record-replay` (`slice-session-wartung`) | geliefert |
| `LH-FA-04` Replay | `welle-record-replay` (`slice-replay-service`, `slice-workspace-use-case` offen); Nutzung im Fallback: `welle-hybrid` | in Arbeit |
| `LH-FA-05` Matching | `welle-record-replay` (`slice-redaction-match-key` geliefert, `slice-replay-service` offen) | in Arbeit |
| `LH-FA-06` Metadaten | Lesen und Erfassen: `welle-record-replay` (`slice-odata-codec-csdl`); **Auslieferung `$metadata` im Mock: `welle-mock`** | Teil geliefert, Rest geplant |
| `LH-FA-07` CRUD | `welle-mock` (`slice-mock-schreiben`, `slice-odata-json-eingabe`, `slice-odata-json-schreiben`) | geplant |
| `LH-FA-08` Abfrageoptionen | `welle-mock` (`slice-odata-url-pfad`, `slice-odata-url-filter`, `slice-edm-ausdruecke`, `slice-query-interpreter`) | geplant |
| `LH-FA-09` Workspaces, Reset | Anlegen, Lesen, Reset: `welle-record-replay` (`slice-workspace-use-case`); Isolation unter Mutation: `welle-mock`; HTTP: `welle-admin-api` | in Arbeit, Rest geplant |
| `LH-FA-10` Snapshots | `welle-mock` (`slice-snapshot-use-case`); Port und Tabellen: `welle-persistenz` geliefert; HTTP: `welle-admin-api` | geplant |
| `LH-FA-11` Coverage | Beweis beim Import: `welle-record-replay` (`slice-baseline-import` geliefert); Prüfung vor der Query: `welle-mock` (`slice-mock-lesen`); Endpunkt: `welle-admin-api` | Teil geliefert, Rest geplant |
| `LH-FA-12` Betriebsarten, Fallback | `record`, `replay`: `welle-record-replay`; `mock`: `welle-mock` (`slice-bootstrap-m3`); `hybrid`, Fallback: `welle-hybrid` | in Arbeit, Rest geplant |
| `LH-FA-13` Management | Use-Case-Ebene: `welle-record-replay` (`slice-session-*`); Listener, Token, Endpunkte, Diagnose: `welle-admin-api` (CLI-Kommandos entfallen, §10 Nr. 3); Audit: `welle-betrieb` | Teil geliefert, Rest geplant |
| `LH-FA-14` Schutz sensibler Daten | `welle-record-replay` (Redaction, Streaming-Redaction, Capture geliefert) | geliefert |
| `LH-FA-15` ETags (Soll) | `welle-etag` | geplant |
| `LH-FA-16` `$expand`, `$batch` (Soll) | `welle-expand-batch` | geplant |
| `LH-FA-17` OData V2 (Kann) | `welle-odata-v2` | geplant |
| `LH-FA-18` Import | Use Case, Codec: `welle-record-replay` (geliefert); **Import-Autorisierung (Abnahmeregel) über HTTP: `welle-admin-api`** (`slice-admin-imports`) | Teil geliefert, Rest geplant |
| `LH-QA-01` Replay-Performance | `welle-betrieb` (`slice-lasttest`); Zielwert nicht freigegeben | geplant, ungeprüft |
| `LH-QA-02` Determinismus | `welle-record-replay` (M2-Test), `welle-mock` (`slice-mock-abnahme`, 100 Läufe) | in Arbeit, Rest geplant |
| `LH-QA-03` Isolation, Konsistenz | `welle-persistenz`, `slice-race-gate` (`welle-record-replay`); 20 Workspaces und Neustart: `welle-mock` (`slice-mock-abnahme`) | Teil geliefert, Rest geplant |
| `LH-QA-04` Datensicherheit | Redaction: `welle-record-replay`; Canary-Test über alle Artefakte und Netzwerktest der Admin-Endpunkte: `welle-admin-api` (`slice-test-security`) | Teil geliefert, Rest geplant |
| `LH-QA-05` Fehlerdiagnose | `welle-record-replay` (`slice-fehlermodell-konfiguration`), `welle-mock` (`ODR-E6001`, `ODR-E6003`) | Teil geliefert, Rest geplant |
| `LH-QA-06` Portabilität | Build, Test: `welle-fundament`; Persistenz: `welle-persistenz`; **Container-Image mit Datenvolumen und Neustarttest: `welle-betrieb` (`slice-container-image`, nach `welle-dev-container`, §10 Nr. 5), vorher ohne Welle** | Teil geliefert, Rest geplant |
| `LH-QA-07` Protokolltreue | `welle-record-replay` (`slice-protokolltreue`) | in Arbeit |
| `LH-RB-01` OData-Kompatibilität | Lesen: `welle-record-replay`; Matrix `SPEC-049`: `welle-mock` (`slice-mock-abnahme`) | Teil geliefert, Rest geplant |
| `LH-RB-02` HTTP, TLS | Upstream: `welle-record-replay` (`slice-upstream-tls` geliefert); Listener: `welle-betrieb` (`slice-listener-tls`); Admin-Listener: `welle-admin-api` (`slice-admin-tls`) | Teil geliefert, Rest geplant |
| `LH-RB-03` Entwicklungsumgebung | `welle-dev-container` (Worktree `wt-dev-container`), `welle-fundament` | nicht in diesem Abgleich angelegt |
| `LH-RB-04` Schutz von Daten | Aufbewahrung, Löschen: `welle-record-replay`; Audit der Löschungen: `welle-betrieb` (`slice-audit-log`); Zugriffskontrolle: `welle-admin-api` | Teil geliefert, Rest geplant |

**Ohne Welle vor diesem Abgleich (jetzt zugewiesen):** `LH-FA-07`, `LH-FA-08`, `LH-FA-10`, `LH-FA-12` (Modi `mock` und `hybrid`), `LH-FA-15`, `LH-FA-16`, `LH-FA-17`, die HTTP-Seite von `LH-FA-13`, `LH-QA-01`, die Container-Hälfte von `LH-QA-06`, `LH-RB-02` (Listener), der Audit-Teil von `LH-RB-04`.

**Weiterhin ohne Welle:** der Konfigurationsschlüssel `LINK_REWRITE` (`SPEC-092`). Er hat keinen Eintrag im Lastenheft, kein Akzeptanzkriterium und damit keinen beobachtbaren Closure-Trigger; eine Welle dafür wäre eine Phantom-Welle. Siehe §10, Frage 4.

## 4. Schnitt der Wellen und Gründe

| Welle | Aufwand | Meilenstein | Entscheidung |
|---|---|---|---|
| `welle-mock` | L (11 Slices) | M3 | Snapshots (`LH-FA-10`) gehören hinein: `Restore` und Mutation teilen die Workspace-Transaktion, ein Snapshot ohne Mutation zeigt nichts, und der Closure-Trigger (Snapshot-Restore im M3-Test, Neustart-Konsistenz aus `LH-QA-03`) beobachtet beides zusammen. Sie stehen als letzter Slice vor den Abnahmetests; `slice-snapshot-use-case` ist ein Teil, den die Welle bei Bedarf als vorab benannten Schnitt abgeben kann, weil er keinen Mechanismus mit dem Query-Ast teilt. |
| `welle-hybrid` | M (3 Slices) | M4 | Eigene Welle: Hybrid ist kein Rechenkern, sondern die Zusage „schreibende Requests nie implizit weitergereicht“ über alle Zweige; das *Mehr* ist der Verbindungszähler am Referenzdienst. |
| `welle-betrieb` | L (7 Slices) | M5 | Nicht mit der Admin-API zusammengelegt: Die Admin-API **braucht** `AuditLog` (jeder Schreibendpunkt, `SPEC-005`), `Telemetry` (Zähler je Fehlercode in `/__admin/diagnostics`) und den TLS-Baustein; die umgekehrte Richtung (Betrieb braucht Admin-API) gibt es nicht, auch der Lasttest läuft ohne HTTP-Admin. Das Container-Image wurde hierher gelegt, weil es keine Welle hatte und der Neustarttest den ganzen Dienst fährt. |
| `welle-admin-api` | L (8 Slices) | M6 | Eigene Welle wegen des Sicherheitsnachweises (`LH-QA-04.a`) und des Rundlaufs über HTTP. Die CLI-Kommandos waren aufgenommen, obwohl das Lastenheft sie nicht fordert (§8, Punkt 5); sie entfallen (§10 Nr. 3). |
| `welle-etag` | S (3 Slices) | — | Unabhängig von den beiden anderen Soll- und Kann-Wellen. `Prefer` ist aufgenommen: Es ist der Soll-Teil der CRUD-Matrix (`LH-FA-07.b`), hat aber keine eigene Anforderung. |
| `welle-expand-batch` | M (4 Slices) | — | `$expand` und `$batch` stehen in `LH-FA-16` gemeinsam; sie teilen Interpreter und Workspace-Transaktion, die Äste sind sonst unabhängig. Der Start hängt an einer Entscheidung des Architect zur Datenbasis der Navigationen (§8, Punkt 6). |
| `welle-odata-v2` | L (4 Slices) | — | Kann-Umfang mit eigener ADR vor dem Start; eigene Welle, weil V2 sonst die Closure der Soll-Wellen an eine Kann-Anforderung bände. |

**Warum nicht eine `welle-erweiterungen`:** Die drei Punkte sind voneinander unabhängig startbar (kein gemeinsamer Mechanismus außer dem Mock-Kern), haben verschiedene Priorität (Soll, Soll, Kann) und verschiedene Closure-Belege; eine gemeinsame Welle bände den Abschluss der Soll-Punkte an V2.

**Coverage-Schwelle, `LH-QA-07`, `LH-RB-03`:** nicht neu angelegt (`welle-sensoren`, `slice-protokolltreue`, `welle-dev-container`), nur in §3 genannt.

## 5. Kanten des Graphen und ihre Prüfung

| Kante | Prüfung gegen Spezifikation und Architektur |
|---|---|
| `welle-record-replay` → `welle-mock` | `MockUseCase` liest die Baseline aus dem Import, den Workspace, den Codec zum Lesen und läuft im Composition Root und HTTP-Adapter aus `slice-bootstrap-m2`. Die Slices der reinen Funktionen (URL-Parser, Ausdrücke, Interpreter) brauchen nur das Modell und könnten früher starten (Frage 7). |
| `welle-mock` → `welle-hybrid` | `HybridUseCase` ruft `MockUseCase` über den Inbound Port (Architektur §2, §2.5); ohne `MockUseCase` kein Hybrid. `ReplayUseCase` kommt über `welle-record-replay` transitiv. |
| `welle-mock` → `welle-betrieb` | Spans `SPEC-029` und `SPEC-030` und der Lasttest des Mockservers (`LH-QA-01.a`) brauchen Mock und Mutation. Der Rest der Welle (Port, Audit, TLS, Image) braucht den Mock nicht, die Welle schließt aber nur mit ihm. |
| `welle-betrieb` → `welle-admin-api` | `AuditLog` (`SPEC-005` Schlusszeile), `Telemetry` (Diagnose), TLS-Baustein (`SPEC-089`). Snapshot- und Restore-Endpunkte brauchen zusätzlich `welle-mock`, das über `welle-betrieb` mittelbar vorliegt; die Kante `welle-mock` → `welle-admin-api` ist deshalb nicht eingezeichnet. |
| `welle-mock` → `welle-etag`, `welle-expand-batch`, `welle-odata-v2` | Mutation und `entity_version` (ETag), Interpreter und Workspace-Transaktion (`$expand`, `$batch`), Ports des Mock-Kerns (V2). Keine Kante untereinander; zwei Berührungen ohne Kante stehen in den Welle-Plänen. |
| `welle-record-replay` → `welle-dev-container` | Reihenfolge nach M2 (Entscheidung des Users, 2026-10-30): Die Gate-Kette ist dann stabil, wenn der Container sie fährt. Inhaltlich unabhängig; die Welle ändert weder `src/` noch Produktverträge. Option: Gerüst und Lauf vorziehen (Welle-Plan §2). |
| `welle-dev-container` → `slice-container-image` | Der Neustarttest im Container fährt Docker im Testlauf und braucht den Daemon des Arbeitscontainers; die Kante gilt für diesen Slice, nicht für die Welle `welle-betrieb` (deren Trigger bleibt `welle-mock`). |
| `welle-record-replay` → `welle-sensoren` | Zwei Trigger: Teil 1 (`slice-unsichtbare-zeichen`, `slice-guide-groessen-gate`) nach `slice-groessenmarke-werkzeug`, vor `slice-workspace-use-case` (Reihenfolge, kein Blocker); Teil 2 (`slice-abdeckung-deklaration`, `slice-coverage-gate`) nach `welle-record-replay` (Backfill über 618 Testfunktionen in 105 Dateien berührt die Dateien von M2). |

Nicht eingezeichnet, weil nicht vorhanden: `welle-hybrid` → `welle-admin-api` (die Diagnose nennt den Modus, braucht den Hybrid-Betrieb nicht), `welle-admin-api` → `welle-betrieb` (siehe oben), Kanten von den Soll- und Kann-Wellen zu den Muss-Wellen (Priorisierung, keine Abhängigkeit).

## 6. Meilensteine

M3 bis M6 haben je einen beobachtbaren Test als Trigger (Mock ohne Upstream; Verbindungszähler im Hybrid; Lasttest mit p95; Rundlauf über den Admin-Listener). M7 ist die Erstabnahme des Lastenhefts („Muss-Anforderungen sind abnahmebindend“) mit einer externen Bestätigung (Owner bestätigt `LH-QA-01`, Lastenheft `Accepted`). Die Soll- und Kann-Wellen tragen keinen Meilenstein (Lastenheft: nicht Teil der Erstabnahme).

## 7. Welle-Dateien und Slice-Dateien: Entscheidung

- **Welle-Dateien für alle sieben neuen Wellen** (Entscheidung des Users, per `cp` aus `welle.template.md`, `diff -q` ohne Unterschied, in-place gefüllt). Sie führen Ziel, Trigger, Closure-Kriterien, Slice-Liste (Name, Ziel, Bezug), Abhängigkeiten, Out-of-Scope und Register-Sichtung. Der Hinweis der Vorlage „geplante Wellen bekommen noch keine Datei“ ist in der Lifecycle-Zeile durch den Verweis auf diese Entscheidung ersetzt.
- **Slice-Dateien in `open/` nur für `welle-mock`** (11 Pläne, per `cp` aus `slice.template.md`): Die Welle startet als nächste und braucht Pläne, deren Ziel und Abgrenzung der Architect vor dem Vertrag lesen kann. Sie bleiben, wie sie sind, bis der Slice vor dem Code steht (Marke und Sichtung setzt der Planner dann, `AGENTS.md` §3.14).
- **Keine Slice-Dateien für die übrigen sechs Wellen:** Ein Slice-Plan altert vor seinem Start; die Pläne von `welle-hybrid` bis `welle-admin-api` hängen an Verträgen und Schnitten von `welle-mock`, die Soll- und Kann-Wellen an Entscheidungen des Architect (Navigationsdaten, 412-Code, V2-ADR). `plan-welle` Schritt 6 legt sie beim Start an; jede Welle-Datei sagt es in §4.
- **Geschlossene Sendungen (`AGENTS.md` §3.13):** `slice-snapshot-use-case` trägt den Punkt aus `slice-workspace-use-case` in §1 (*Übernimmt*), `slice-bootstrap-m3` den Nachzug von `SPEC-110`, `slice-odata-url-filter` den Punkt `$filter` aus `slice-odata-url-pfad`. Die Audit-Sendung der Slices `slice-session-use-case` und `slice-session-wartung` (beide in `done/`, nicht änderbar) steht in der Tabelle von `welle-betrieb` und wird in `slice-audit-log` beim Start als Übernahme geführt.
- **Handbuch (Entscheidung des User, `welle-benutzerhandbuch`):** Das Handbuch beschreibt nur den Ist-Zustand des Binary; ein Gate hält Fehlercodes, Schlüssel und Kommandos gegen das Handbuch. Jede Produktwelle trägt deshalb ein Closure-Kriterium („Das Handbuch deckt alles ab, was das Binary dieser Welle kann; das Kataloge-Gate ist grün“) und in §4 die Pflicht, den DoD-Punkt „Handbuch-Kapitel im selben Slice nachgezogen“ in jede Slice-Datei zu übernehmen, die sichtbares Verhalten liefert. In `welle-mock` trägt ihn `slice-bootstrap-m3` verbindlich (erst dort erreicht der Modus `mock` das Binary), `slice-mock-lesen`, `-mock-schreiben` und `slice-snapshot-use-case` bedingt; Parser, Ausdrücke, Interpreter, Codecs und `slice-mock-abnahme` nicht. **Trigger berührt:** `slice-handbuch-meldungskatalog` startet erst nach `slice-bootstrap-m2`, liegt also nicht automatisch vor dem Trigger von `welle-mock`; ich habe ihn als zweite Bedingung in den Trigger aufgenommen (Roadmap-Zeile und Welle-Datei), weil sonst kein Slice seine DoD erfüllen kann. Die Folgewellen erben die Bedingung über `welle-mock`. Der Aufwand bleibt (L, M, L, L, S, M, L), jeder Slice mit sichtbarem Verhalten trägt ein Kapitel mehr.
- **Roadmap:** Die sieben Wellen stehen in *Nächste Wellen* (mit Verweis auf die Datei), nicht unter *Offene Wellen*: Der Abschnitt ist nach dem Regelwerk derivativ (ein Zeiger je **eröffnete** Welle), keine dieser Wellen ist gestartet. Die Regel (Frage 6): *Offene Wellen* trägt eine Welle, deren Trigger eingetreten ist.

## 8. Lücken und Widersprüche (Lastenheft nicht geändert)

1. **`LH-FA-18` ist nicht ohne HTTP abnehmbar:** Die Abnahmeregel *Import-Autorisierung* („nicht autorisierter Admin-Client“) setzt den Admin-Listener voraus. `welle-record-replay` liefert den Use Case; der Beleg dieser Regel liegt erst in `welle-admin-api`. `LH-FA-18` ist bis dahin nur teilweise belegt.
2. **`LH-FA-12`, Verweis auf `LH-FA-04`:** Das Lastenheft nennt Record, Replay und Mock und einen „kombinierten Betrieb“; der Replay-Fallback des Mock (`HYBRID_REPLAY_FALLBACK`, `SPEC-079`) hat im Lastenheft weder Anforderung noch Akzeptanzkriterium, nur die Spezifikation (`LH-FA-12.a/b`) und die Architektur kennen ihn. Der Closure-Test von `welle-hybrid` leitet ihn aus der Spezifikation ab.
3. **`LH-QA-01`:** Zielwert „vorläufig, nicht freigegeben“; ohne Entscheidung des Owners gibt es keine fachliche Freigabe. Die Welle `welle-betrieb` misst und berichtet; ein Durchfallen ist ein Carveout, kein stilles Rot. M7 hängt an der Bestätigung.
4. **Spezifikation nennt Tests und Gates, die es nicht gibt:** `SPEC-037` (End-to-End-Test), `SPEC-041` (`make test-security`), `SPEC-042` (`make test-contract`) werden in `LH-QA-04.a`, `LH-RB-01.a`, `LH-RB-03.a` genannt, haben aber keinen Abschnitt, und die Make-Ziele stehen nicht in `harness/README.md`. Sie sind den Slices `slice-test-security` und `slice-mock-abnahme` zugeordnet; der Architect legt Vertrag und Zielnamen fest.
5. **CLI:** Das Lastenheft fordert keine CLI-Kommandos (nur „bedienbare Schnittstelle“ in §1 und die Verwaltung in `LH-FA-13`); die Spezifikation hat keinen Abschnitt zu Kommandos, nur `ARC-006`. `slice-cli-kommandos` hatte damit keine Anforderung und keinen Vertrag und entfällt (Frage 3).
6. **Navigationsdaten für `$expand`:** Der Import liest eingebettete Navigationen nicht (`SPEC-126` Nr. 9); ein Mock-`$expand` hat damit keine Datenbasis. Die Welle `welle-expand-batch` beginnt erst nach der Entscheidung des Architect.
7. **Kein Fehlercode für `412`:** Die Fehlertabelle (`SPEC-016` bis `SPEC-113`) führt keinen `ODR`-Code für einen Precondition-Fehler, obwohl die Entscheidungsmatrix `412` nennt (`LH-FA-12.b`). Vor `welle-etag` zu ergänzen.
8. **Service-Dokument** (`GET` auf den Service-Root) und **Aushandlung des JSON-Metadatenlevels**: kein Vertrag, beide Randformen des Mock (Service-Dokument: Frage 2).
9. **Metriken:** Die Spezifikation nennt „Metriken“ (`SPEC-026` bis `SPEC-030` sind Spans) und `/__admin/diagnostics` mit Zählern, aber keinen Abrufweg für Metriken und keinen Export von Spans.
10. **Startkonfiguration:** `SPEC-110` Nr. 8 lehnt den Modus `mock`, auch den Default, ab, und Nr. 2 lehnt `ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL` ab. Jede Welle hebt ihren Teil auf; das ist je Welle ein Nachzug durch den Architect.
11. **Der Lastenheft-Absatz *Erstabnahme* und §5** passen zur Aufteilung: Out-of-Scope-Punkte von `LH-FA-08` (`$expand`) und §5 (V2, Actions, Functions) liegen bei den Soll- und Kann-Wellen oder bleiben `501`. Ein Widerspruch zur Aufteilung wurde nicht gefunden.

## 9. Register-Sichtung

Das Register trägt 19 Einträge. Offen sind sieben: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×), `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×), `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×), `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (1×), `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×). Keiner erreicht beim Anlegen 3×; der Planner schreibt dabei keine Evidenz. Die einschlägigen Einträge stehen als Risiko in den Welle-Plänen (§5 dort) und in §8 der Mock-Slices; `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` berührt keine der Wellen. Der Eintrag `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` ist `geplant` in `slice-groessenmarke-werkzeug`; die Marken der Mock-Slices setzt der Planner beim Start mit dem Werkzeug.

## 10. Offene Fragen an den User

1. **Reihenfolge Betrieb und Admin-API:** Der Schnitt stellt `welle-betrieb` vor `welle-admin-api`, weil Audit, Telemetry und TLS-Baustein dort entstehen. Wer die Admin-API früher will (Teststeuerung über HTTP), kann `AuditLog` und die Zähler in `welle-admin-api` ziehen und `welle-betrieb` danach schneiden; das ändert die Kante.
   **Ausgang (Entscheidung des Users, 2026-10-30):** Betrieb vor Admin-API bleibt; die Admin-API braucht `AuditLog`, `Telemetry` und den TLS-Baustein, die Gegenrichtung gibt es nicht. Keine Änderung der Kante.
2. **Service-Dokument:** Gehört `GET <Service-Root>` zum Muss-Umfang des Mock? Auswirkung auf das Handbuch: Was das Binary nicht kennt, darf das Handbuch nicht nennen (Kataloge-Gate); gehört es nicht dazu, steht es weder im Binary noch im Handbuch, und `slice-mock-lesen` liefert `501`.
   **Ausgang (Entscheidung des Users, 2026-10-30):** Nicht im Muss-Umfang von `welle-mock`. Der Architect von `slice-mock-lesen` prüft vor dem ersten Code-Commit am Referenz-Client des M3-Tests, ob eine Mindestfassung mit eigenem Vertrag nötig ist (Auftrag in `slice-mock-lesen` §1 und §6, Ausgang dort *offen*, Adresse Architect); bis dahin und sonst `501`, kein Eintrag im Handbuch.
3. **CLI-Kommandos:** Soll das Lastenheft sie fordern (Change Request vor `Accepted`), oder entfällt `slice-cli-kommandos`? Auswirkung auf das Handbuch: Ohne Anforderung und Vertrag entsteht kein Kommando und damit kein Eintrag im Kommando-Katalog.
   **Ausgang (Entscheidung des Users, 2026-10-30):** `slice-cli-kommandos` entfällt; `welle-admin-api` führt 8 Slices. Es gibt keinen Lastenheft-Anker (`LH-FA-13` verlangt eine Verwaltungsschnittstelle, keine Kommandos). Der Kommando-Katalog des Handbuchs führt weiter die Kommandos, die `Run` kennt (Start, Version). Verlangt der Owner Kommandos später, ist es ein Change Request vor `Accepted`.
4. **`LINK_REWRITE`:** Gibt es eine Anforderung dafür? Ohne sie bleibt der Schlüssel `off` und es entsteht keine Welle.
   **Ausgang (Entscheidung des Users, 2026-10-30):** Keine Welle. `off` ist zugesagt (`SPEC-092`: nur `off` ist zulässig, jeder andere Wert verhindert den Start mit `ODR-E2001`); ohne Eintrag im Lastenheft gibt es keinen beobachtbaren Closure-Trigger.
5. **Container-Image und Neustarttest** (`LH-QA-06`): in `welle-betrieb` (so geplant) oder als eigener Slice früher, weil er nur `record` und `replay` braucht? Der Neustarttest im Container setzt einen Docker-Daemon im Testlauf voraus; wie das mit `welle-dev-container` zusammenhängt, ist offen.
   **Ausgang (Entscheidung des Users, 2026-10-30):** `slice-container-image` bleibt in `welle-betrieb`. Kante `welle-dev-container` → `slice-container-image` (der Graph der Roadmap führt sie als Kante von `welle-dev-container` zu `welle-betrieb`, beschriftet mit dem Slice). Der Trigger von `welle-betrieb` (`welle-mock` in `done/`) bleibt; die Bedingung hängt am Slice (`welle-betrieb` §4), die Welle schließt erst mit ihm.
6. **Roadmap-Lesart:** Die Wellen mit Datei stehen unter *Nächste Wellen*. Nach dem Wortlaut des Regelwerks gehört jede flache Welle-Datei unter *Offene Wellen*; die Lesart hier hält „offen“ für gestartet. Eine Zeile Änderung, falls gewünscht.
   **Ausgang (Entscheidung des Users, 2026-10-30):** Der Ist-Zustand bleibt. Die Regel steht in der Roadmap unter *Offene Wellen*: Eine Welle steht dort, wenn ihr Trigger eingetreten ist; sonst steht sie unter *Nächste Wellen* mit Link. Folge der Umplanung: `welle-dev-container` (Trigger jetzt `welle-record-replay` in `done/`) steht unter *Nächste Wellen*.
7. **Früherer Start einzelner Slices:** Die reinen Funktionen (`slice-odata-url-pfad`, `slice-odata-url-filter`, `slice-edm-ausdruecke`, `slice-query-interpreter`) bräuchten `welle-record-replay` nicht. Soll die Reihenfolge das nutzen (Eintrag im Drift-Log), oder bleibt der Trigger der Welle?
   **Ausgang (Entscheidung des Users, 2026-10-30):** Nein, der Trigger der Welle bleibt. Frühestens nach `welle-dev-container` (parallele Arbeit in getrennten Arbeitsbäumen trägt dann); eine Änderung ist dann eine Umplanung mit Eintrag im Drift-Log der Roadmap.
8. **Fachlicher Owner von `LH-QA-01`:** Wer bestätigt den Zielwert, und bis wann braucht M7 ihn?
   **Ausgang:** Offen, Adresse: User (als Owner bestätigt oder nennt er den Wert). Frist: vor dem Start von `slice-lasttest`, sicher vor M7. Festgehalten in `welle-betrieb` §3 und im Status von M7 der Roadmap.
9. **Regel „nur `Edit` und `Write`“:** Das Anlegen der 18 Dateien aus Vorlagen lief mit `cp` und anschließend ausschließlich mit `Edit`; ein Skript wurde nicht benutzt (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).

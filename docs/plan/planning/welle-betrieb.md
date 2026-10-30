# Welle welle-betrieb: Telemetrie, Audit-Log, Listener-TLS, Container-Image und Lasttest

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); bis ihr Trigger eintritt steht sie in der
Roadmap unter *Nächste Wellen*. Ihre Slice-Dateien legt `plan-welle` (Schritt 6)
beim Start an, weil ein Slice-Plan vor seinem Start veraltet.

**Zielmeilenstein:** M5 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Der Recorder meldet, was er tut, und weist seine Zusagen zum Betrieb nach: strukturierte Logs, Metriken und Spans (`SPEC-026` bis `SPEC-030`), Audit-Einträge für Löschungen und Verwaltungsvorgänge (`LH-RB-04`, `LH-FA-13`), HTTPS am OData-Listener (`LH-RB-02`), ein Container-Image mit Datenvolumen, das einen Neustart übersteht (`LH-QA-06`), und ein reproduzierbarer Lasttest, der die zusätzliche Verarbeitungszeit von Replay und Mockserver am Gesamtspan misst (`LH-QA-01`).
Vier Punkte hat `welle-record-replay` ausdrücklich hierher gegeben: keinen Port `Telemetry` (der Record-Dienst loggt über `slog`), keine Audit-Einträge der Löschvorgänge (`slice-session-use-case`, `slice-session-wartung`), kein TLS des Listeners und den Lasttest, der ohne Spans nicht messbar ist.
Eine Welle, weil der Closure-Trigger mehr beobachtet als die DoDs: die Messung am Gesamtspan setzt den Einbau in alle Pfade voraus und läuft erst mit Mock und Telemetry zusammen, und der Neustarttest im Container fährt den ganzen Dienst.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-mock` liegt in `done/`: der Einbau der Spans `SPEC-029` und `SPEC-030` und der Lasttest des Mockservers brauchen `MockUseCase` und `WorkspaceUseCase` mit Mutation.
- M3 ist erreicht: der M3-Test aus `slice-bootstrap-m3` ist grün und Teil von `make test`.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Ein Test durchläuft Record, Replay und Mock und findet die Spans `SPEC-026` bis `SPEC-030` mit ihren Pflicht-Attributen, ohne Query-Literale und Bodies als Attribut; die Zähler je Fehlercode steigen für jeden Code, den der Test auslöst.
- Jede Löschung (Session, Aufbewahrung, Workspace) schreibt einen Audit-Eintrag ohne Geheimnisse (`LH-RB-04.a`).
- Der HTTPS-Test am OData-Listener ist grün: gültiges Zertifikat wird bedient, ein Client mit unbekannter CA lehnt ab, ein Start mit nur einem der beiden Schlüssel `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` scheitert mit `ODR-E2001`.
- Der Neustarttest im Container ist grün: leeres Datenvolumen, Aufzeichnung, Neustart des Containers, Recordings lesbar (`LH-QA-06.a`).
- Der Lasttest nach `LH-QA-01.a` läuft reproduzierbar mit 2 vCPU und 4 GiB und gibt p95 des Gesamtspans für Replay und Mock aus; die Zahlen stehen in der Closure-Notiz. Die Bestätigung des Zielwerts durch den fachlichen Owner (`LH-QA-01`) ist kein Closure-Kriterium dieser Welle; sie ist Teil von M7. **Offene Entscheidung, Adresse: User** (Owner von `LH-QA-01`): der Zielwert; Frist vor dem Start von `slice-lasttest`, sicher vor M7 ([Bericht](../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §10 Nr. 8). Bis dahin misst und berichtet der Lasttest gegen den vorläufigen Wert des Lastenhefts.
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (TLS-Schlüssel des Listeners, Container-Image, Fehlercodes); das Kataloge-Gate aus `slice-handbuch-meldungskatalog` ist grün.
- Die Closure-Notiz liegt als `welle-betrieb-results.md` im Ruheort `done/`; ihre *Steering-Loop-Einträge* tragen je verkörperter Regel mit dem Anker `seit welle-betrieb` das Feld `liegt in`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-telemetry-port | Port `Telemetry` und Adapter: strukturierte Logs mit den Feldern der Spezifikation, Metriken, Spans, Zähler je Fehlercode; keine Credentials und Bodies | `LH-QA-05`, `LH-QA-04`, `LH-FA-01`; `SPEC-026`, `SPEC-027`, `SPEC-028`, `SPEC-029`, `SPEC-030` |
| slice-telemetry-einbau-gateway | Einbau in HTTP-Adapter (Gesamtspan), Record-Dienst (Capture-Span) und Replay-Dienst (Match-Span); ersetzt den Logger des Record-Dienstes | `LH-QA-01`, `LH-FA-01`, `LH-QA-05`; `SPEC-026`, `SPEC-027`, `SPEC-028` |
| slice-telemetry-einbau-mock | Einbau in Mock-Dienst (Query-Span) und Workspace-Mutation (Mutations-Span) | `LH-QA-01`, `LH-FA-07`; `SPEC-029`, `SPEC-030` |
| slice-audit-log | Port `AuditLog` und Adapter; Audit-Einträge der Löschvorgänge in `SessionUseCase`, `MaintenanceUseCase` und `WorkspaceUseCase` (Sendung der Slices `slice-session-use-case` und `slice-session-wartung`) | `LH-RB-04`, `LH-FA-13`; `LH-RB-04.a`, `LH-FA-13.a` |
| slice-listener-tls | TLS des OData-Listeners (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`), wiederverwendbarer Baustein für den Admin-Listener; HTTPS- und Zertifikatsfehler-Test | `LH-RB-02`, `LH-QA-04`; `LH-RB-02.a`, `SPEC-088` |
| slice-container-image | Laufzeit-Image mit Datenvolumen, Start aus leerem Verzeichnis, Neustarttest | `LH-QA-06`, `LH-RB-03`; `LH-QA-06.a`, `SPEC-077` |
| slice-lasttest | Lasttest nach `LH-QA-01.a` mit Messpunkt am Gesamtspan für Replay und Mock | `LH-QA-01`; `LH-QA-01.a`, `SPEC-026` |

Die Slice-Dateien legt `plan-welle` (Schritt 6) beim Start der Welle an: Ein Slice-Plan altert vor seinem Start, und die Pläne dieser Welle hängen an den Schnitten von `welle-mock` (Einbau der Spans).

**Bedingung des Slice `slice-container-image`** (Entscheidung des Users, 2026-10-30; `plan-welle` Schritt 6 übernimmt sie in dessen Start): Er startet erst, wenn `welle-dev-container` in `done/` liegt. Der Neustarttest im Container (`LH-QA-06.a`) fährt Docker im Testlauf und braucht dafür den Daemon des Arbeitscontainers; ohne ihn wäre der Test an den Daemon des Hosts gebunden, den die Gate-Läufe teilen. Der Trigger der Welle (§2) bleibt: Die übrigen Slices starten ohne `welle-dev-container`; die Welle schließt (§3) erst, wenn auch dieser Slice in `done/` liegt. Liegt `welle-dev-container` bei Eintritt des Triggers noch nicht in `done/`, läuft `slice-container-image` zuletzt oder die Welle wartet an dieser Stelle.

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. `plan-welle` (Schritt 6) übernimmt ihn in jede Slice-Datei; in dieser Welle betrifft er `slice-listener-tls` (Schlüssel `LISTEN_TLS_*`), `slice-container-image` (Start und Datenvolumen) und `slice-telemetry-port`, soweit Logfelder oder Codes sichtbar werden; `slice-audit-log` und `slice-lasttest` nur, wenn sie ein sichtbares Verhalten des Binary einführen.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: `welle-admin-api` (`AuditLog` für jeden Schreibendpunkt, `Telemetry` für die Zähler der Diagnose, TLS-Baustein des Listeners).
- Wird blockiert von: `welle-mock` (Einbau der Spans von Mock und Workspace, Lasttest des Mockservers), mittelbar `welle-record-replay`; für `slice-container-image` allein zusätzlich von `welle-dev-container` (Bedingung des Slice in §4, Kante im Abhängigkeitsgraphen der Roadmap).
- Unabhängig: `welle-hybrid`, `welle-etag`, `welle-expand-batch`, `welle-odata-v2`.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-telemetry-port` → `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` → `slice-lasttest`; `slice-audit-log` braucht keinen Slice der Telemetrie und liegt davor oder danach; `slice-listener-tls` und `slice-container-image` brauchen keinen Slice der Welle (das Image fährt den Dienst in den Modi, die dann vorliegen).
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für `slice-lasttest` (jede Zahl p95 hängt an Last und Messendem und trägt beides); `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) für `slice-container-image` und `slice-lasttest`, die lange Läufe in die Gate-Kette bringen könnten (Entscheidung des Architect: Gate oder Werkzeug, Muster von [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für den Einbau in nebenläufige Pfade; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für Testhilfen, die `ctx` und Ströme umhüllen; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Abruf der Metriken und Export der Spans an einen Collector — die Spezifikation nennt keinen Abrufweg (Frage an den Architect vor dem Start von `slice-telemetry-port`); ein verpflichteter externer Dienst scheidet nach Lastenheft §5 aus.
- Admin-Endpunkte, `/__admin/diagnostics`, TLS des Admin-Listeners — `welle-admin-api`.
- Bestätigung des Zielwerts von `LH-QA-01` — Entscheidung des fachlichen Owners, nicht des Repos.
- Tuning bei verfehltem Zielwert (Indizes, Puffer) — eigener Slice nach der Messung; fällt der Wert durch, ist es ein Carveout mit Folge-Slice, kein stilles Rot.
- Orchestrierung des Betreibers (Compose-Dateien, Kubernetes) — das Image und sein Start sind Gegenstand, nicht seine Umgebung.
- Der Arbeitscontainer für Entwickler (`LH-RB-03`) — `welle-dev-container`.
- Coverage-Schwelle als Gate — gehört zu `welle-sensoren`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-betrieb-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).

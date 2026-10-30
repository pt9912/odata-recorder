# Welle welle-record-replay: Aufzeichnen und Wiederabspielen

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. **Geplante Wellen bekommen noch keine
Datei:** Sie stehen in der Roadmap unter *Nächste Wellen* und nirgends sonst —
zwei Positionen, nicht drei.

**Zielmeilenstein:** M2 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-20.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Ein Test zeichnet über den Recorder gegen einen Referenzdienst auf und spielt dieselben Antworten danach ohne Upstream byte-gleich wieder ab (`LH-FA-01`, `LH-FA-02`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`).
Der Weg dahin ist der vollständige Pfad der Spezifikation: Record mit Capture und Redaction, Session beenden, Import der Aufzeichnung in eine Baseline (`LH-FA-18`, `LH-FA-11`), Workspace aus der Baseline (`LH-FA-09`), Replay mit Zuordnung, Match-Key und Cursor.
Ein Replay ohne Workspace gibt es nicht (`SPEC-052`), ein Workspace ohne Baseline auch nicht (`SPEC-003`, `SPEC-108`), und eine Baseline entsteht nur durch den Import; deshalb gehören OData-Codec und Import in diese Welle.
Eine Welle, weil der Closure-Trigger mehr beobachtet als die DoDs der Slices: den Rundlauf durch alle Schichten, vom HTTP-Eingang über Services und Ports bis in die SQLite-Datei und zurück.
Drei Slices kommen aus dem Steering Loop und der Messmethode dazu: `slice-race-gate` bringt den Race-Detektor ins Gate, bevor der HTTP-Adapter entsteht (`BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`), `slice-groessenmarke-werkzeug` macht die Größenmarke zu einem Werkzeug, bevor die nächste Marke gesetzt wird (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`), und `slice-protokolltreue` den Vergleichstest zu `LH-QA-07` am Referenzdienst des M2-Tests.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-persistenz` liegt in `done/` (Repositories, Blob-Ablage und Rundlauftest hinter den Outbound Ports).
- M1 ist erreicht: `make gates` fährt Build, Test, Lint und `a-check` auf Go-Code, jedes durch eine Gegenprobe belegt.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün und fährt das Ziel mit `go test -race` aus `slice-race-gate`; `make gegenproben` ist grün.
- Der M2-Test aus `slice-bootstrap-m2` ist grün und Teil von `make test`: Er startet den Recorder im Modus `record` gegen einen Referenzdienst, zeichnet `$metadata` und Collection- und Entity-Abfragen auf, beendet die Session, importiert sie in eine Baseline, legt einen Workspace an und spielt die Abfragen im Modus `replay` ab, nachdem der Referenzdienst beendet ist; Status, Header (soweit nach `SPEC-045` auszuliefern) und Body sind byte-gleich, und der Referenzdienst hat im Replay keine Verbindung gesehen.
- Der Vergleichstest aus `slice-protokolltreue` ist grün und Teil von `make test`: Direktzugriff und Zugriff über den Recorder im Modus `record` ergeben am Referenzdienst dieselbe entpackte Darstellung nach `LH-QA-07.a` der Spezifikation.
- Die Closure-Notiz liegt als `welle-record-replay-results.md` im Ruheort `done/`; ihre *Steering-Loop-Einträge* tragen je Regel mit dem Anker `seit welle-record-replay` das Feld `liegt in` (`AGENTS.md` §3.10 mit der Rate je Mutation und je Lauf von `go test`, `AGENTS.md` §3.14 und §3.15, beide samt der Schärfung vom 2026-10-21 — in §3.14 Schnitt beim Handoff, Faktor 1,25 auf die ganze Schätzung, Marke vom Planner gesetzt mit Obergrenze 3000, Klassen nebenläufig und jede andere Zeile, in §3.15 gleiche verglichene Werte und, seit der Schärfung vom 2026-10-23, die Aufrufstellen der Ports als Menge und die Allaussage über Stellen und Arten als Schleife —, `AGENTS.md` §3.16, `.harness/skills/reviewer.md`, die Zielorte von `slice-race-gate` einschließlich des Rests von §3.10, das Ziel von `slice-groessenmarke-werkzeug` für `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` `.claude/commands/implement-slice.md` (Schritt 14), `.claude/agents/implementer.md` und die MEDIUM-Liste von `.harness/skills/reviewer.md` für `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`, die Schärfung von `AGENTS.md` §3.17 in `.claude/commands/implement-slice.md` (Schritt 19) und `.claude/agents/verifier.md` für `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`) und den Eintrag zu `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (`AGENTS.md` §3.16).

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-fehlermodell-konfiguration | Fehlermodell mit `ODR`-Codes und Startkonfiguration | `LH-QA-05`, `LH-FA-12`, `LH-FA-04` |
| slice-clock-idgenerator | Zeit und Kennungen hinter Querschnitts-Ports | `LH-QA-02` |
| slice-redaction-match-key | Redaction von Headern und Query, Match-Key, Replayfähigkeit | `LH-FA-05`, `LH-FA-14`, `LH-QA-02` |
| slice-streaming-redaction | Streaming-Redaction von JSON-Bodies, fail-closed | `LH-FA-14`, `LH-FA-02`, `LH-QA-04` |
| slice-capture-pipeline | Capture-Pipeline: Tee, Grenzen, Dekompression, Abschluss | `LH-FA-02`, `LH-FA-14`, `LH-QA-04` |
| slice-record-service | `RecordUseCase`, Port `UpstreamClient` und `Drain` | `LH-FA-01`, `LH-FA-02`, `LH-FA-13`, `LH-QA-07` |
| slice-upstream-adapter | Upstream-Adapter `httpupstream` | `LH-FA-01`, `LH-RB-02`, `LH-QA-07` |
| slice-upstream-tls | TLS-Konfiguration des Upstream-Adapters `httpupstream` | `LH-RB-02`, `LH-FA-01` |
| slice-session-wartung | `MaintenanceUseCase`: Startwartung, Blob-Abgleich und Aufbewahrung | `LH-FA-02`, `LH-FA-03`, `LH-RB-04` |
| slice-session-use-case | `SessionUseCase`, Teil Verwaltung: Anlegen, Lesen, Auflisten, Löschen | `LH-FA-02`, `LH-FA-03` |
| slice-session-beenden | `SessionUseCase`, Teil Beenden: `Drain`, Statusregel, `CloseAll` | `LH-FA-02`, `LH-FA-03`, `LH-FA-13` |
| slice-odata-codec-csdl | OData-Codec, Teil Metadaten (CSDL) | `LH-FA-06`, `LH-RB-01` |
| slice-odata-codec-json | OData-Codec, Teil Entities und Collections (JSON) | `LH-FA-18`, `LH-RB-01` |
| slice-baseline-import | `ImportUseCase` mit Coverage-Beweis | `LH-FA-18`, `LH-FA-11`, `LH-FA-14` |
| slice-groessenmarke-werkzeug | Größenmarke als Werkzeug statt als Prosa | `LH-QA-06` |
| slice-workspace-use-case | `WorkspaceUseCase`: Anlegen, Lesen, Löschen, Reset | `LH-FA-09`, `LH-FA-04` |
| slice-replay-service | `ReplayUseCase`: Zuordnung, Strategien, Cursor | `LH-FA-04`, `LH-FA-05`, `LH-FA-14` |
| slice-race-gate | Race-Detektor im Gate | `LH-QA-03`, `LH-QA-06` |
| slice-http-adapter | HTTP-Driving-Adapter für Record und Replay | `LH-FA-01`, `LH-FA-04`, `LH-FA-02` |
| slice-bootstrap-m2 | Composition Root, Referenzdienst und M2-Test | `LH-FA-12`, `LH-FA-04`, `LH-QA-02` |
| slice-protokolltreue | Vergleichstest der Protokolltreue (Direktzugriff gegen Proxy) | `LH-QA-07`, `LH-FA-01` |

**Auftrag beim Start eines Slice aus `open/`** (Entscheidung des Users vom 2026-10-21): Die Pläne in `open/` bleiben, wie sie sind, bis der jeweilige Slice vor dem Code steht. Vor seinem ersten Code-Commit setzt der Planner die Marke nach `AGENTS.md` §3.14, statt sie nur nachzuziehen, und zieht die Sichtung nach:

1. §4 nach der Form von `AGENTS.md` §3.14: Marke mit Metrik; Schätzung aus der Tabelle *Zusage · Fänger* für den Stand beim Handoff, × 85, wenn der Code oder sein Fänger nebenläufig ist, sonst × 40 (jede andere Zeile; der erste Slice mit Zeilen, die keine reinen Funktionen sind, nennt den gemessenen Wert in §4), plus 200 je Port-Fake oder Test-Server; liefert der Slice einen Leser fremder Bytes, zählt jede andere Zeile × 65 statt × 40, und die Produktcode-Größe, die der Architect erwartet, kommt als fester Anteil hinzu; die Marke setzt der Planner auf die ganze Schätzung, feste Anteile eingeschlossen, × 1,25, aufgerundet auf volle Hundert, und ersetzt eine ältere Marke dabei mit Grund in §4; geschnitten wird vor dem ersten Code-Commit nur, wenn die Schätzung × 1,25 über 3000 liegt, und lässt sich kein Teil ohne geteilten Mechanismus abtrennen, steht der Grund in §4; ein Schnitt beim Handoff nur an einem Teil, den §4 vor dem Code benennt. Eine Marke mit freier Schätzung tragen `slice-odata-codec-csdl`, `slice-odata-codec-json`, `slice-baseline-import`, `slice-workspace-use-case`, `slice-replay-service`, `slice-http-adapter` und `slice-bootstrap-m2`. Liegt `slice-groessenmarke-werkzeug` in `done/`, setzt und prüft der Planner die Marke mit dem Werkzeug, und die Faktoren stehen in dessen Vertrag; bis dahin gilt die Form dieses Absatzes.
2. Die Sichtung in §8 mit den Ausgängen aus `d143752`: `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` geplant in `slice-groessenmarke-werkzeug` (bis dahin gilt `AGENTS.md` §3.14), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` verkörpert in `AGENTS.md` §3.16, `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` geplant in `slice-race-gate` (die Rate je Mutation und je Lauf ist verkörpert in `AGENTS.md` §3.10), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` verkörpert in `AGENTS.md` §3.15 und `BEO-REPO/randform-erst-im-review-entschieden` verkörpert in `AGENTS.md` §3.16, beide geschärft; der Stand jedes Eintrags steht in seiner `state.md` im Register.

**Zusatz für den Start von `slice-workspace-use-case`** (Entscheidung des Architect in `slice-groessenmarke-werkzeug` §3, *Ausgang des Architect* Nr. 13; Ausnahme von der Entscheidung des Users oben nicht zulässig, daher hier und nicht im Plan in `open/` geführt): Beim Start trägt der Planner in den Start-Trigger des Plans ein, dass `slice-groessenmarke-werkzeug` in `done/` liegt, setzt und prüft die Marke mit dem Werkzeug (Block und Tabelle nach `SPEC-128` Nr. 3) und führt die Sendung nach `AGENTS.md` §3.13 (§1 *Ausdrücklich NICHT* und DoD des Plans vorher lesen). Gestartet ist er vor der Closure des Werkzeug-Slice (Linie A, Zweig `slice-workspace-use-case`, Worktree `odata-recorder-wt-workspace`): Der Zweck der Bedingung, die Marke mit dem Werkzeug zu setzen, ist erfüllt, weil das Werkzeug in `main` lauffähig ist; die Marke gilt mit dem Vertrag `SPEC-128` am Tag der Marke, den Stand nennt der Plan in §4.

**Zusatz zur Reihenfolge vor `slice-workspace-use-case`** (Umplanung vom 2026-10-30, Drift-Log der Roadmap): Die beiden Slices von Teil 1 der Welle [welle-sensoren](welle-sensoren.md) (`slice-unsichtbare-zeichen`, `slice-guide-groessen-gate`) liegen vor `slice-workspace-use-case`. Das ist eine Reihenfolge, kein Blocker: Es gibt keine Abhängigkeit im Code (die Sensoren ändern `Makefile`, `harness/`, `tools/` und `Dockerfile`, nicht `src/`), und der Start-Trigger des Plans bleibt `slice-baseline-import` und `slice-groessenmarke-werkzeug` in `done/`. Beim Start von `slice-workspace-use-case` liegt Teil 1 nicht in `done/` und ist nicht begonnen; der Start ändert dessen Dateien nicht (`Makefile`, `harness/`, `tools/`, `Dockerfile`), und der Plan nennt den Grund in §4. Der Plan in `open/` blieb dafür unverändert.

**Zusatz für den Parallelbetrieb** (Umplanung vom 2026-10-30, Drift-Log der Roadmap): Die Slices ab `slice-workspace-use-case` (Linie A) laufen in einem eigenen Worktree und Zweig, parallel zu Teil 1 von [welle-mock](welle-mock.md) (Linie B); die Übergangsregeln (Sperre für schwere Läufe, Stempel, Zweig im Kopf `Verantwortlich:`, Kennungsbereich, was nur der Integrationsschritt schreibt) stehen in [`README.md`](README.md#parallelbetrieb-in-worktrees-übergangsregeln) und gelten für jeden Start eines Slice dieser Welle. Die Pläne in `open/` bleiben dafür unverändert; die Start-Trigger ändern sich nicht.

**Zusatz für den Start von `slice-bootstrap-m2`** (Umplanung vom 2026-10-30): Der Plan in `open/` bleibt bis dahin unverändert. Der Planner zieht beim Start §1 *Ausdrücklich NICHT* nach: Die Kommandos über Start und Version hinaus entfallen (Entscheidung des Users, siehe §6), sie sind keine Folgewelle mehr; das Verhalten des Slice ändert sich dadurch nicht.

`slice-session-wartung` ist nach dieser Form begonnen und trägt `MaintenanceUseCase`: §4 dort nennt die Metrik, die Schätzung aus der Tabelle des Architect (2150) und die Marke 2700, vom Planner nach dem Vertrag gesetzt; seine Sichtung in §8 trägt die Zahlen und Ausgänge des Registers vom Start. `SessionUseCase` ist `slice-session-use-case`, vor dem ersten Code-Commit geschnitten (die gemeinsame Erwartung lag bei Schätzung × 1,25 rund 3250 bis 3750, `AGENTS.md` §3.14); der Plan dort ist nach dieser Form begonnen: §4 dort nennt die Metrik und die Zählung des Planner beim Start (30 Zeilen, 9 nebenläufig, 5 Port-Fakes, Schätzung × 1,25 rund 3300, Spanne 2700 bis 3700; die ältere Erwartung „rund 2000 bis 2100“ ist ersetzt). Beim Start ist er ein zweites Mal vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14): `slice-session-use-case` trägt die Verwaltung (Anlegen, `OpenInitialSession`, Lesen, Auflisten, Löschen mit der Bereinigung; nach der Tabelle des Architect 20 Zeilen, 5 nebenläufig, 4 Fakes, Schätzung 1825, × 1,25 = 2300, Marke 2300 gesetzt), `slice-session-beenden` das Beenden (`Drain`, Statusregel, `CloseAll`; beim Schnitt Schätzung 1315, × 1,25 = 1700; nach der Tabelle des Architect 30 Zeilen, 17 nebenläufig, 2 Fakes, Schätzung 2365, × 1,25 = 2956, Marke 3000 vom Planner nach dem Vertrag gesetzt, Metrik in §4 dort; gemessen 2998 nach zwei Durchgängen der Verifikation, die Marke ist nicht überschritten). Die Sichtung in §8 beider Pläne trägt die Zahlen des Registers vom Start.

`slice-race-gate` und `slice-protokolltreue` sind nach dieser Form schon nachgezogen (§4 dort, die Marke nach `AGENTS.md` §3.14 mit Grund der Ersetzung); ihre Sichtung in §8 trägt die Ausgänge aus `d143752`. `slice-upstream-tls` trägt die Marke nach dieser Form ohne Ausnahme (§4 dort): die Schätzung nimmt für den geschriebenen Code des TLS-Commits den gemessenen Wert, die Formel der Tabelle gibt dieselbe Marke; die Zahlen seiner Sichtung in §8 stammen aus den Dateien unter `evidence/`.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: Teil 2 von `welle-mock` (Teil 1, die reinen Funktionen, braucht sie nicht und läuft parallel, Zusatz in §4) und über `welle-mock` die Folgewellen (wegen der Services, Ports und des Codecs); `welle-benutzerhandbuch` mit den Slices nach `slice-bootstrap-m2`; Teil 2 von [welle-sensoren](welle-sensoren.md) (Backfill und Schwelle gegen den Code von M2); [welle-dev-container](welle-dev-container.md) (Reihenfolge nach M2: die Gate-Kette ist dann stabil).
- Wird blockiert von: — (`welle-persistenz` ist geschlossen). In der Reihenfolge davor: Teil 1 von [welle-sensoren](welle-sensoren.md) liegt vor `slice-workspace-use-case` (Zusatz in §4, kein Blocker).
- Unabhängig: inhaltlich Welle welle-dev-container (sie ändert weder `src/` noch Produktverträge); die Reihenfolge ist nicht beliebig, siehe Blockiert.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-fehlermodell-konfiguration` und `slice-clock-idgenerator` zuerst; `slice-redaction-match-key` → `slice-streaming-redaction` → `slice-capture-pipeline` → `slice-record-service`; `slice-session-wartung` (`MaintenanceUseCase`, braucht keinen Slice des Record-Asts) → `slice-session-use-case` (Verwaltung; vor dem ersten Code-Commit aus `slice-session-wartung` geschnitten) → `slice-session-beenden` (braucht zusätzlich `slice-record-service`: `Drain`; beim Start aus `slice-session-use-case` geschnitten); `slice-record-service` → `slice-upstream-adapter` (der Adapter braucht den Port `UpstreamClient` im Code; der Dienst ist gegen einen Fake des Ports geprüft und braucht den Adapter nicht) → `slice-upstream-tls` (TLS-Konfiguration des Adapters, beim Handoff von `slice-upstream-adapter` geschnitten; braucht dessen Code in `done/`); `slice-odata-codec-csdl` → `slice-odata-codec-json` → `slice-baseline-import` (braucht zusätzlich `slice-redaction-match-key`, `slice-session-wartung` und `slice-session-use-case`) → `slice-groessenmarke-werkzeug` (Reihenfolge, keine Abhängigkeit im Code: der Slice berührt `tools/` und `harness/`, nicht `src/`; vor `slice-workspace-use-case`, damit dessen Marke nach dem Werkzeug gesetzt wird; braucht `slice-baseline-import` in `done/` wegen der drei Messpunkte) → `slice-workspace-use-case` → `slice-replay-service` (braucht zusätzlich `slice-redaction-match-key` und `slice-streaming-redaction`); `slice-race-gate` braucht keinen Slice der Welle (er ändert `Dockerfile`, Gate-Kette und Gegenproben, nicht den Code der Pakete) und liegt vor `slice-http-adapter`; `slice-http-adapter` braucht `slice-record-service`, `slice-replay-service` und `slice-race-gate`; `slice-bootstrap-m2` nach den übrigen (braucht zusätzlich `slice-upstream-adapter`, `slice-upstream-tls`, `slice-session-wartung`, `slice-session-use-case` und `slice-session-beenden`); `slice-protokolltreue` zuletzt (braucht `slice-bootstrap-m2`: Referenzdienst und Composition Root). Der Ast Record und der Ast Codec und Import hängen erst im Replay und im Bootstrap zusammen und lassen sich in dieser Reihenfolge einzeln liefern.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Mock-Modus (zustandsbehaftetes CRUD, Query-Interpreter, `$metadata`-Auslieferung, Serialisierung zurück ins OData-JSON) und Hybrid-Modus mit Fallback — Folgewellen; der Codec liefert hier nur das Lesen.
- CLI-Kommandos über Start und Version hinaus — entfallen (Entscheidung des Users, 2026-10-30): Das Lastenheft fordert keine (`LH-FA-13` verlangt eine Verwaltungsschnittstelle, keine Kommandos); keine Folgewelle trägt sie.
- Admin-API (Sessions, Imports, Workspaces, Snapshots, Diagnose, Authentisierung) — Folgewelle; Session beenden, Import und Workspace erreicht der M2-Test über den Composition Root, nicht über HTTP.
- Snapshot erzeugen und wiederherstellen (`LH-FA-10`) — gehört zum Mock, der Zustand ändert; `WorkspaceUseCase` liefert Anlegen, Lesen, Löschen und Reset.
- Telemetrie, Audit-Log und Metriken (`AuditLog`, `Telemetry`, Spans) — Folgewelle; ohne sie ist der Lasttest zu `LH-QA-01` nicht messbar, er bleibt ausdrücklich ungeprüft.
- TLS des Listeners (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`) und des Admin-Listeners — Folgewelle; der Upstream-Client trägt die TLS-Optionen des Upstreams (`LH-RB-02`).
- Umschreiben von Links und Weiterleitungen (`LINK_REWRITE`) — nach Spezifikation in der ersten Version nur `off`.
- Optimistische Concurrency und ETags (`LH-FA-15`), erweiterte OData-Funktionen (`LH-FA-16`), OData V2 (`LH-FA-17`) — Soll und Kann, nicht Teil des Muss-Umfangs.
- Coverage-Schwelle als Gate — kommt mit mehr Code; hier entstehen die Tests der Verträge, nicht eine Schwelle.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-record-replay-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).

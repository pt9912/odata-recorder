# Slice slice-telemetry-port: Ports `Telemetry` und `TelemetryReader`, Regeln der Attribute, Zähler je Fehlercode und `FakeTelemetry` (Teil `P` des Schnitts)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-FA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-230` bis `SPEC-232` (spezifikation.md §5, der Vertrag dieses Slice) · `SPEC-026` bis `SPEC-030` (spezifikation.md §5) · die Logfelder am Ende von spezifikation.md §4 · `SPEC-121` Nr. 12 (Protokoll des Record-Dienstes, heute über `*slog.Logger`), `SPEC-123` Nr. 11, `SPEC-124` Nr. 9 · `SPEC-005` (Diagnose: Zähler je Fehlercode) · `SPEC-109`, `SPEC-113` (Fehlerabbildung) · `ARC-020`, `ARC-021` · `architecture.md` §Support-Ports.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-telemetry-port)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Querschnitts-Ports `Telemetry` und `TelemetryReader` (Paket `src/hexagon/ports/support`, `ARC-021`) und der Kern ihres Driven Adapters (`src/adapters/driven/telemetry`, `ARC-020`, nur Standardbibliothek) stehen, mit allem, was die drei Teile des Vertrags (`SPEC-230` bis `SPEC-232`) gemeinsam brauchen: die Typen `Attr`, `Snapshot`, `Verteilung`, `SpanRecord` mit `support.Spans()` und `support.Eimer()`; die Regeln der Attribute (geschlossene Schlüssel, neun Wertformen, genau ein Verwurf je Attribut in der Rangfolge Schlüssel vor Wert, sieben Gründe); `CountError` mit dem Zähler je Fehlercode (33 Codes, nebenläufig ohne Verlust); `Snapshot` als vollständige Kopie (jeder Schlüssel, auch mit 0); `New` mit `Config` und der Abweisung ungültiger Werte; ein `Log`, das die Regeln anwendet und Verwürfe zählt, aber noch keine Zeile schreibt, und ein `StartSpan` ohne Wirkung; ein Fake im Paket `supporttest`, der jedes Argument festhält und jede Operation zählt. Der Slice liefert Ports und Kern, nicht Zeile, Spans und Einbau: kein Dienst und kein Adapter des Produkts ruft den Port, und kein Composition Root verdrahtet den Adapter.

**Teil des Schnitts vom 2026-10-30 (Planner, vor dem Code):** Dieser Slice ist Teil `P` der Kette `P` → `L` → `S`, die der Architect vorgeschlagen hat (§4). Der Vertrag `SPEC-230` bis `SPEC-232` gehört zu allen drei Slices; dieser liefert `SPEC-230` Nr. 1 bis 4, Nr. 8 und 9 ohne Verteilung und Ring (die Form und die Zähler `Errors` und `Dropped`), Nr. 10 (Fake), `SPEC-232` Nr. 1 (`New`, `Config`) und Nr. 5 a (`CountError` verliert nichts). `slice-telemetry-log` liefert Zeile, Handler, Senke und den Canary über `Log` (`SPEC-231` Nr. 1 bis 5, `SPEC-232` Nr. 2, 3, 5 c, d und f für die Seite von `Log` und `Handle`); `slice-telemetry-spans` liefert `StartSpan`, `Set`, `End`, Verteilung, Ring und Span-Zeile (`SPEC-230` Nr. 5 bis 7 und den Rest von Nr. 8 und 9, `SPEC-231` Nr. 6, `SPEC-232` Nr. 5 b und e, dazu 5 f für die Seite von `StartSpan` und `End`). Die Entscheidungen des Architect in §3 gelten für alle drei.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Zeile, Stufen, `msg`, `Handler`, Abbildung der zehn Ereignisse der Dienste, Ausfall der Senke, Sperre der Zeile, Canary über `Log` und `Handler` — `slice-telemetry-log`; die Zeile hängt am Schreiber und an der Senke, die der Kern dieses Slice nur annimmt (nil-Prüfung in `New`), nicht benutzt.
- `StartSpan` mit Wirkung, `Set`, `End`, Eltern-Kind, Verteilung, Ring, Span-Zeile, Canary über die Wege der Spans — `slice-telemetry-spans`; das Ende eines Spans füllt Verteilung und Zeile zugleich und braucht den Schreiber von `slice-telemetry-log`.

- Einbau in den HTTP-Adapter (Gesamtspan), den Record-Dienst (Capture-Span, Ersatz des `*slog.Logger`) und den Replay-Dienst (Match-Span) — `slice-telemetry-einbau-gateway`; der Port ist vom Einbau trennbar, und der Einbau braucht Code, den dieser Slice nicht kennt (HTTP-Adapter, Replay-Dienst).
- Einbau in Mock-Dienst und Workspace-Mutation (`SPEC-029`, `SPEC-030`) — `slice-telemetry-einbau-mock`; der Port kennt die fünf Span-Namen als Daten und braucht dafür weder `MockUseCase` noch `WorkspaceUseCase`.
- Verdrahtung des Adapters im Composition Root und Umstellung der Dienste `session` und `maintenance` auf den Handler — Empfehlung: `slice-telemetry-einbau-gateway`; die Annahme steht aus (§6); dieser Slice ändert `src/bootstrap` nicht.
- Endpunkt, Netz-Port und Export an einen Collector für Metriken und Spans — das Lastenheft führt keine Anforderung dazu, ein verpflichteter externer Dienst scheidet nach Lastenheft §5 aus; abgerufen wird im Prozess (`TelemetryReader`) und über die Span-Zeile im Log (Entscheidung 1).
- Der Admin-Endpunkt `/__admin/diagnostics`, der die Zähler je Fehlercode ausgibt — `welle-admin-api` (`SPEC-005`, `slice-inspection-use-case`); dieser Slice liefert den Zähler und den Lese-Teil `TelemetryReader`, nicht den Endpunkt. Die Sendung ist dort noch nicht auf den Namen des Lese-Teils gezogen (§6).
- Spans und Zähler des Hybrid-Fallbacks — kein Nehmer in `welle-betrieb`; der Port schließt sie nicht aus (Entscheidung 2, §6).
- Audit-Einträge — `slice-audit-log`; ein Audit-Eintrag ist kein Log-Eintrag (anderer Port, anderer Zweck: Nachweis einer Löschung, nicht Diagnose).

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Ports `Telemetry` und `TelemetryReader` und Fake: die Operationen und Typen von `SPEC-230`, die Mengen der Schlüssel (13 für `Log`, 11 Pflicht-Attribute der fünf Spans), die neun Wertformen mit Grenzen und Zeichen, die Rangfolge der Verwürfe, `support.Spans()` und `support.Eimer()`; `FakeTelemetry` hält jedes Argument fest und zählt jede Operation (`SPEC-230` Nr. 1 bis 4 und 10, `LH-QA-05`, `LH-FA-01`).
- [ ] Kern des Adapters: `New` mit `Config` weist ungültige Werte ab, `CountError` zählt je Code der Tabelle (33) ohne Verlust und einen fremden Code als `fehlercode`, `Snapshot` ist vollständig (auch mit 0) und eine Kopie, `Log` wendet die Regeln der Attribute an und zählt Verwürfe, schreibt aber keine Zeile, `StartSpan` ist ohne Wirkung; nebenläufig sicher und unter `make test-race` ohne Befund; ein Ausfall ändert das Ergebnis des Aufrufers nicht (`SPEC-230` Nr. 8 und 9, `SPEC-232` Nr. 1 und 5 a, `LH-QA-04`: die Formen schließen die Zeichen der Träger aus).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary, oder macht er Logfelder am Binary sichtbar: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Port wird erst vom Einbau gerufen). Der Architect hat entschieden: kein Schlüssel, Code oder Kommando (Entscheidung 7); die Logfelder werden am Binary sichtbar, sobald der Einbau den Adapter verdrahtet.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/support/` | update | Ports `Telemetry` und `TelemetryReader` mit `Attr`, `Snapshot`, `support.Spans()`, `support.Eimer()` neben `Clock` und `IDGenerator` |
| `src/hexagon/ports/support/supporttest/` | update | `FakeTelemetry` (hält jedes Argument fest, zählt jede Operation) |
| `src/adapters/driven/telemetry/` | neu | Kern des Adapters: `New`, `Config`, Regeln der Attribute, Verwurf-Zähler, `CountError`, `Snapshot`; `Log` prüft und zählt, schreibt keine Zeile; `StartSpan` ohne Wirkung; Zeile, `Handler()` und Senke folgen in `slice-telemetry-log`, Spans, Verteilung und Ring in `slice-telemetry-spans` |
| `spec/spezifikation.md` | update | `SPEC-230` bis `SPEC-232` in §5, je ein Verweis in §4 (Logfelder) und bei `SPEC-005`; Wortlaut von `SPEC-121` Nr. 12 unverändert |
| `spec/architecture.md` | update | eine Zeile `TelemetryReader` in der Tabelle der Querschnitts-Ports; `ARC-020`, `ARC-021` bleiben (die Aufzählung `Telemetry` deckt den Lese-Teil nicht, die Zeile der Tabelle nennt ihn); frei von Slices, Wellen und ADR-Bezügen (`AGENTS.md` §3.4) |
| `.a-check.yml` | unverändert | der Adapter liegt in `src/adapters/driven/` und hat die Kante zu `ports-support`; kein Eintrag nötig |
| `docs/plan/adr/`, `go.mod`, `go.sum` | unverändert | Standardbibliothek genügt, keine ADR (Entscheidung 1); `go.mod` und `go.sum` bleiben, Fänger P14 |
| Handbuch | unverändert | Binary unverändert (Entscheidung 7) |
| Tests | neu | je Zusage ein Fall (Tabelle unten); nebenläufige Tests laufen unter `make test-race` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation **vor dem ersten Code-Commit** (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Der Slice liegt seit dem Start in `in-progress/`; der Architect arbeitete im selben Zweig (Übergangsregel 3 in [`README.md`](../README.md) §Parallelbetrieb), der Planner hat danach geschnitten und die Marke gesetzt (§4). Die Zeilen mit dem Präfix `L` und `S` und die ihnen zugewiesenen Teile des Rasters stehen seit dem Schnitt in `slice-telemetry-log` und `slice-telemetry-spans`; die Entscheidungen des Architect gelten für alle drei.

**Auslegung des Planners beim Schnitt (§6, Ausgang *entschieden*; bestätigt vom Architect):** Der Kern in diesem Slice enthält ein `Log`, das Schlüssel und Werte prüft und Verwürfe zählt (der Fänger P17 braucht es), aber keine Zeile schreibt; `New` nimmt `Writer` und `Clock` an und weist nil ab (P10), benutzt sie aber erst in den Folge-Slices. Die Zeile 5 der Zusatztabelle in §6 („Pflicht-Attribut mit ungültigem Wert, dann `End`“) braucht `StartSpan` und `End`; sie ist beim Schnitt von P17 nach S02 in `slice-telemetry-spans` gewandert, P17 fängt die übrigen fünf Zeilen. Bestätigt vom Architect: `Log` wendet `MinLevel` nicht an; `MinLevel` ist in `New` jeder Wert gültig.

**Kennungsbereiche der Linie D** (Setzung des Auftrags vom 2026-10-30; der Integrationsschritt nimmt sie in die Übergangsregeln, `README.md` §Parallelbetrieb Nr. 6, auf): `SPEC-230` bis `SPEC-259`, `ARC-040` bis `ARC-049`, ADR `0040` bis `0049`, `CO-060` bis `CO-079`. Dieser Slice vergibt `SPEC-230`, `SPEC-231`, `SPEC-232`; `ARC-*`, ADR und `CO-*` nicht. Die bestehenden Kennungen `SPEC-026` bis `SPEC-030`, `SPEC-005` und `SPEC-121` stehen unverändert; ihr Zusammenhang mit dem Vertrag steht in `SPEC-230` bis `SPEC-232` und in je einem Verweis in §4 und bei `SPEC-005`. `MR-*` vergibt dieser Zweig nicht, `LH-*` ändert er nicht.

**Entscheidungen des Architect** (Alternative, Grund, Folge im Betrieb; die Regeln im Einzelnen stehen in der Spezifikation, `AGENTS.md` §3.8). Die Nummern sind die der Aufträge des Planners:

| Auftrag | Entscheidung (Ort) | Alternative | Grund | Folge im Betrieb |
|---|---|---|---|---|
| 1 Abrufweg, Senke | Spans und Zähler sind im Prozess über den **zweiten** Port `TelemetryReader` lesbar (ein Stand zu einem Zeitpunkt), und jedes `End` schreibt eine `INFO`-Zeile mit `duration_ms`, `span_id`, `parent_id` ins Log des Prozesses; nur Standardbibliothek, keine ADR (`SPEC-230` Nr. 9, `SPEC-231` Nr. 6) | (a) OpenTelemetry-SDK mit Export an einen Collector; (b) Endpunkt `/__admin/metrics`; (c) nur Lesen im Prozess; (d) Lese-Operation im Port `Telemetry` | (a) neue Abhängigkeit samt ADR, und ein Collector ist ein verpflichteter externer Dienst (Lastenheft §5); (b) neue Schnittstelle des Produkts ohne Anforderung des Owners; (c) erreicht den Lasttest nicht, der das gebaute Binary als eigenen Prozess misst; (d) ein schreibender Dienst könnte lesen, `InspectionUseCase` brauchte Schreibrecht und Dienste sähen den Adapter | Je beendetem Span eine Zeile (20 Requests/s ergeben 1200 Zeilen je Minute im Container-Log); für Werkzeuge außerhalb des Prozesses sind Spans nur über diese Zeilen lesbar, ein Endpunkt oder Collector braucht eine Anforderung des Owners (Frage in §6); der Lasttest liest `duration_ms` mit der Auflösung einer Mikrosekunde aus den Zeilen oder läuft im Prozess |
| 2 Form des Ports | Drei Operationen (`Log`, `StartSpan`, `CountError`), der Span mit `Set` und `End`; der Span-Kontext reist im `ctx`, den `StartSpan` zurückgibt; die Menge der Span-Namen ist **geschlossen** und steht als Daten in `support.Spans()`; der Port prüft die Form Token der Werte, keine Wertemenge (`SPEC-230` Nr. 1, 5, 6) | offene Namen; Wertemengen von `recorder.mutation_type` und `recorder.match_strategy` hier festlegen; Eltern als Argument | Ein Tippfehler im Namen färbt einen Test, statt still eine Reihe zu erzeugen; ein Hybrid-Span ist eine Zeile in §5 und im Register, keine neue Operation; die Werte gehören den Einbau-Slices; der `ctx` wird schon durchgereicht | Ein Wert der Einbau-Slices ohne die Form Token wird verworfen und zählt `wert`; Eltern-Kind hängt daran, dass die Stelle den zurückgegebenen `ctx` weiterreicht |
| 3 Metriken | `odr_errors_total` (33 Codes), `odr_span_duration_ms` (5 Namen, 12 Eimer bis 5000 ms, Zahl, Summe), `odr_telemetry_dropped_total` (7 Gründe); feste Zahl der Reihen (`SPEC-230` Nr. 4, 8) | Quantile (t-digest) statt Eimern; Dauer als Liste aller Messwerte; Zähler je Pfad | Feste Eimer haben konstanten Speicher und eine Grenze bei 100 ms, dem vorläufigen Zielwert von `LH-QA-01`; kein Schlüssel stammt aus einem Wert eines Aufrufers | p95 aus den Eimern ist nur auf den Eimer genau; exakte Dauern stehen in den Span-Zeilen und im Ring |
| 4 Logfelder | JSON-Zeile mit fester Reihenfolge, `msg`, fehlende Felder fehlen; Stufen `DEBUG` bis `ERROR`, Default `INFO`; `Handler()` für die drei Dienste ohne Änderung ihres Codes; Wortlaut von `SPEC-121` Nr. 12 bleibt (`SPEC-231`) | Dienste auf `Telemetry.Log` umstellen; `slog.JSONHandler` mit Ersetzung der Felder | Der Handler lässt Text, Stufe und Attribute der Dienste unverändert und wendet dieselben Regeln der Werte an; `slog.JSONHandler` kennt weder geschlossene Schlüssel noch Wertformen | Bis zur Verdrahtung loggen die Dienste im Format, das ihr Logger trägt; der Satz „einen Port `Telemetry` gibt es noch nicht“ in `SPEC-121` Nr. 12 ist dann veraltet (§6) |
| 5 Sensible Inhalte | Fünf Stellen: Typ, geschlossene Schlüssel, Wertform, feste Schlüssel der Metriken, nie ein Fehlertext (`SPEC-232` Nr. 2) | Liste verbotener Muster (`Bearer`, `password`) oder Prüfung am Inhalt | Muster erkennen Geheimnisse nicht verlässlich; die Formen schließen die Zeichen der Träger aus (Leerraum, Anführungszeichen, `=`, `$`, `'`, `@`, `/`) | Ein wohlgeformter Wert in einem zulässigen Schlüssel und der Inhalt von `msg` gehen durch (akzeptiertes Negativ L17); jede Stelle des Einbaus übergibt nur eigene Felder und trägt einen Fänger |
| 6 Fehler, Nebenläufigkeit | Die Senke wird synchron unter einer Sperre beschrieben, ein `Write` je Zeile; ein Fehler verliert die Zeile und zählt `senke`; die Aufzeichnung hängt nicht an der Senke (`SPEC-232` Nr. 3, 5) | asynchroner Puffer mit eigener Goroutine, Verlust bei vollem Puffer | Ein Puffer braucht einen Lebenszyklus (Leeren beim Beenden) und macht Tests und den Lasttest ungeordnet; der Standard-Fehlerstrom blockiert erst bei vollem Puffer des Log-Treibers | Ein hängender Log-Treiber hält den Request (`SPEC-232` Ungeprüft); Zeilen gleichzeitiger Aufrufe stehen in unbestimmter Reihenfolge |
| 7 Katalog, Handbuch | Das Binary bleibt unverändert: kein Schlüssel, Code oder Kommando; Stufe, Format und Senke sind fest (`SPEC-231` Nr. 7, `SPEC-232` Nr. 6) | Schlüssel für Stufe und Format | Jeder Schlüssel ist ein Nachzug von `SPEC-110` und ein Handbuch-Kapitel; ein Bedarf ist nicht belegt | Der Betreiber kann die Stufe nicht ändern, bis ein Slice den Schlüssel liefert; die Handbuch-Pflicht dieses Slice entfällt, die der Logfelder liegt beim Einbau, der den Adapter verdrahtet (Auftrag an den Planner, §6) |
| 8 Tabelle, Block | 53 Zeilen (46 `andere`, 7 `nebenläufig`), 2 Fakes, erwarteter Produktcode 1310 Zeilen; die Nr. tragen das Präfix des Teils (`P`, `L`, `S`); Schätzung 5985, Schnittfrage, Schnitt in drei Slices vorgeschlagen (§4) | ein Slice mit Marke 8100; Zeilen bündeln | `SPEC-128` Nr. 1: ab Schätzung 2223 wird vor dem Code geschnitten; Bündeln senkte die Schätzung, ohne Arbeit zu sparen (`SPEC-128` Nr. 10 b) | Der Planner setzt die Marke je Teil (§4) |
| 9 Mengen | Absatz *Mengen für die Fänger* unten | — | — | — |

**Mengen für die Fänger** (`AGENTS.md` §3.15, §3.17; die Zahl liefert jeweils die Tabelle oder der Wortlaut, eine Zählprüfung im Test vergleicht sie):

- Felder des Logkatalogs: zehn; `timestamp` und `level` setzt der Adapter, acht setzt der Aufrufer, dazu fünf weitere (`ereignis`, `richtung`, `grund`, `blob`, `befunde`): 13 Schlüssel der Menge `Log` (`SPEC-230` Nr. 3).
- Spans: 5 Namen mit 11 Pflicht-Attributen (gezählt aus der Tabelle in §5), 55 Paare Span × Schlüssel (die Menge der Schlüssel prüft P02; die Wirkung am Span `slice-telemetry-spans`).
- Fehlercodes: 33 (`model.AlleCodes()`, gegen die Tabelle in `SPEC-109` Nr. 2 und §4).
- Wertformen: 9 (Token, Kennung, Name, Methode, Status, Zahl, Fehlercode, Dauer, Blob-Referenz); davon 6 mit Text, je 12 Zeichen außerhalb (Leerzeichen, `'`, `"`, `{`, `=`, `$`, `@`, `/`, Zeilenumbruch, NUL, Byte 0x80, `:`).
- Gründe des Verwurfs: 7; Eimer-Grenzen: 11 (12 Eimer, geprüft als Form des Stands in P08; die Eimer-Regel `slice-telemetry-spans`).
- Operationen der Ports: 6 (`Log`, `StartSpan`, `CountError`, `Set`, `End`, `Snapshot`).
- Beide Seiten jeder Grenze: Token 32 und 33, Kennung 64 und 65, Name 128 und 129, Methode 16 und 17, Status 99, 100, 599, 600, Zahl −1 und 0, Dauer −1 ns, 0, NaN, ±unendlich, Blob-Referenz 70, 71, 72 Zeichen, `MaxSpans` −1, 0, 1, 65536, 65537.
- Die Mengen zu Stufen, `msg`, Arten von `slog.Attr`, Ereignissen der Dienste und dem Canary über `Log` stehen in `slice-telemetry-log`, die zu Ring, Eimer-Regel und dem Canary über die Spans in `slice-telemetry-spans`.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Dieser Slice führt die Zeilen `P01` bis `P17` (Ports, Regeln der Attribute, Zähler, Lesen, Fake); die Zeilen `L01` bis `L19` (Zeile, Handler, Canary) stehen in `slice-telemetry-log`, die Zeilen `S01` bis `S17` (Spans) in `slice-telemetry-spans`. Die Namen der Tests sind Vorschläge, der Fänger ist die beschriebene Schleife mit Orakel. Ein akzeptiertes Negativ zählt `andere`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| P01 | Die vier Konstruktoren `Str`, `Int`, `Millis`, `ErrorCode` setzen Schlüssel, Art und Wert (`SPEC-230` Nr. 1); `ErrorCode` mit einem Code außerhalb der Tabelle ergibt den leeren Text | `TestAttrKonstruktoren`: Schleife über die 4 Konstruktoren (Zählprüfung 4), Orakel Art und Wert; dazu `ErrorCode(0)` | `ErrorCode` setzt die Art `Int` | andere |
| P02 | Die Mengen der Schlüssel sind geschlossen: `Log` 13, Span 11 (`SPEC-230` Nr. 2, 3, 4) | `TestSchluesselmengen`: Schleife über 24 Schlüssel × 2 Mengen (48 Paare, Zählprüfung 13 und 11), Orakel die Tabelle der Spezifikation; dazu 6 fremde Schlüssel (`timestamp`, `level`, `msg`, `Mode`, leer, `mode` mit Leerzeichen) | `timestamp` in die Menge `Log` aufgenommen | andere |
| P03 | Die neun Wertformen nehmen an und weisen ab an beiden Seiten jeder Grenze, bei leerem Wert und bei falscher Art (`SPEC-230` Nr. 3) | `TestWertformGrenzen`: Schleife über 9 Formen (Zählprüfung 9), je gültig, kürzester, längster, eins darüber, leer, falsche Art; Status 99, 100, 599, 600; Zahl −1, 0; Dauer −1 ns, 0, NaN, ±unendlich; Orakel ein Prüfer aus der Beschreibung, nicht aus dem Code des Adapters | Länge Token `<=` zu `<` (32 abgewiesen) | andere |
| P04 | Die sechs Formen mit Text weisen jedes Zeichen außerhalb ab (`SPEC-230` Nr. 3) | `TestWertformZeichen`: Schleife über 6 Formen × 12 Zeichen (72 Paare), Orakel ein unabhängiger regulärer Ausdruck je Form; Zeichen aus Bytewerten gebaut, Datei auf Rohbytes gegengelesen (`AGENTS.md` §3.18) | Zeichenklasse Kennung enthält `:` | andere |
| P05 | `error_code` nimmt genau die 33 Texte der Tabelle an (`SPEC-230` Nr. 3) | `TestFehlercodeForm`: Schleife über `model.AlleCodes()` (Zählprüfung 33 gegen die Tabelle) und 5 fremde Texte (`ODR-E0000`, `odr-e2001`, `ODR-E2001` mit Leerzeichen, leer, `ODR-E20011`) | Prüfung nur auf das Präfix `ODR-E` | andere |
| P06 | `CountError(c)` erhöht genau den Zähler von `c` um 1 (`SPEC-230` Nr. 8) | `TestCountErrorJeCode`: Schleife über 33 Codes; danach `Errors[c]` gleich 1 und die Summe aller gleich 1; zweimal derselbe Code gibt 2 | Index um 1 verschoben (zählt den Nachbarn) | andere |
| P07 | Ein Code außerhalb der Tabelle erhöht keinen Zähler und zählt `fehlercode` (`SPEC-230` Nr. 8) | `TestCountErrorUnbekannt`: 4 Fälle (`0`, `-1`, letzter Code plus 1, `math.MaxInt`); Summe von `Errors` 0, `Dropped[fehlercode]` 1 je Fall | unbekannter Code wird auf `CodeE9001` gezählt | andere |
| P08 | `Snapshot` trägt jeden Schlüssel, auch mit 0 (`SPEC-230` Nr. 9) | `TestSnapshotVollstaendig`: frischer Adapter; Zählprüfung 33 `Errors`, 5 `Durations` mit je 12 `Counts`, 7 `Dropped` | nur Einträge ungleich 0 | andere |
| P09 | `Snapshot` ist eine Kopie in beide Richtungen (`SPEC-230` Nr. 9) | `TestSnapshotKopie`: Änderung an `Errors`, `Durations`, `Counts`, `Dropped` des Stands wirkt nicht zurück; ein Aufruf nach dem Lesen ändert den früheren Stand nicht (4 Teile); `Spans` und `Attrs` fängt S12 | gibt die interne Map zurück | andere |
| P10 | `New` weist ab: nil-`Writer`, nil-`Clock`, `MaxSpans` unter 0 und über 65536; 0 gilt als 1024; der Fehler nennt keinen Wert (`SPEC-232` Nr. 1) | `TestNewConfig`: 9 Fälle (nil-`Writer`, nil-`Clock`, −1, 0, 1, 65536, 65537, gültig, `MinLevel` über `slog.LevelError`); `errors.Is(model.ErrUngueltig)`, Fehlertext gleich dem von `model.ErrUngueltig` | keine Prüfung nach oben | andere |
| P11 | `FakeTelemetry` hält jedes Argument ungeprüft fest und zählt jede Operation (`SPEC-230` Nr. 10) | `TestFakeAufzeichnung`: Schleife über die 6 Operationen (Zählprüfung 6); Orakel die erwartete Folge mit `ctx`-Identität; ein Schlüssel außerhalb der Menge wird durchgereicht; zweites `End` gezählt; Änderung des übergebenen Slices danach ändert die Aufzeichnung nicht | `Snapshot` wird nicht gezählt | andere |
| P12 | Der Fake hält den Eltern-Span fest und gibt einen abgeleiteten `ctx` zurück (`SPEC-230` Nr. 10) | `TestFakeEltern`: 3 Fälle (Kind im `ctx` des Elternspans, ohne Eltern, Wert und Abbruch des Eingabe-`ctx` bleiben) | Rückgabe-`ctx` ist `Background` | andere |
| P13 | `Telemetry` hat 3 Methoden, `Span` 2, `TelemetryReader` 1; kein Lesen in `Telemetry`; Adapter und Fake erfüllen beide Ports (`SPEC-230`, `ARC-021`) | `TestPortSignaturen`: Reflexion mit Zählprüfung 3, 2, 1 und den Formen der Methoden; Zuweisung von Adapter und Fake an beide Schnittstellen | `Snapshot` in `Telemetry` aufgenommen | andere |
| P14 | Der Adapter importiert nur `model`, `ports/support` und die Standardbibliothek; `go.mod` und `go.sum` bleiben (`SPEC-230` Einleitung) | `make a-check` (Kante `adapters-driven` zu `ports-support`), bei Import eines Dienst-Pakets rot; `git diff <Basis> -- go.mod go.sum` leer in der Closure-Notiz | Import von `services/record` im Adapter | andere |
| P15 | `CountError` verliert keinen Zuwachs (`SPEC-232` Nr. 5 a) | `TestCountErrorGleichzeitig`: 64 Goroutinen × 500 Aufrufe, Codes reihum über 33; Orakel die Summe je Code; Wirkungs-Fänger und Detektor (`make test-race`, je Prozess `-test.count=1`); Rate je Prozess vom Implementer | Zuwachs ohne Sperre | nebenläufig |
| P16 | `FakeTelemetry` ist für gleichzeitigen Aufruf sicher und zählt genau (`SPEC-230` Nr. 10) | `TestFakeGleichzeitig`: 64 × 500 Aufrufe; die Aufzeichnung hat 32000 Einträge; Detektor | Zähler ohne Sperre | nebenläufig |
| P17 | Ein Attribut zählt höchstens einen Verwurf, in der Rangfolge Schlüssel vor Wert; der erste angenommene Wert eines Schlüssels gilt (`SPEC-230` Nr. 4) | `TestVerwurfRangfolge`: die 5 Zeilen 1 bis 4 und 6 der Zusatztabelle in §6 über `Log` (die Zeile 5 braucht `StartSpan` und `End` und fängt S02 in `slice-telemetry-spans`); Orakel der Stand von `Dropped` und der gültige Wert | Wert wird vor dem Schlüssel geprüft (unbekannter Schlüssel mit ungültigem Wert zählt `wert`) | andere |

**Betriebsregeln für alle Rollen:**

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts (Entscheidung des Users vom 2026-10-21).
- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` (höchstens zwei Container je Linie zugleich) unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Die Image-Falle: das Image `odata-recorder:race` trägt `/src/src` eingebacken; ein Lauf, der Quellen hineinkopiert, leert den Pfad zuvor oder nutzt die Stufe `deps`. Ein Rot unter Last wird auf ruhigem Host wiederholt, bevor es als Befund gilt.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python (Register-Eintrag `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` steht bei 1× in der Basis dieses Zweigs, laut Auftrag des Users im Hauptbaum bei 3×: eine weitere Verletzung ist ein Befund des Reviewers); keine Hilfsdateien im Wurzelverzeichnis. Commit-Nachrichten gehen als Datei (`git commit -F`), nicht als Heredoc.
- Eine Hülle um ein Argument (ein Fake, der `ctx`, einen Strom oder einen Writer umhüllt, ein Test-Writer der Senke) darf keine Eigenschaft des Originals verdecken (Abbruch, Fehler, Blockieren, Zahl der Aufrufe; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`): der Fake hält jedes Argument fest, zählt jede Operation, und eine Panik des Fakes färbt den benannten Test, nicht einen anderen.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Die Klassen der Tabelle sind generierbar (hier: jedes Feld des Katalogs, jeder Span-Name und jedes Pflicht-Attribut von `SPEC-026` bis `SPEC-030`, jeder Fehlercode): schon der Handoff führt den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt, nicht im Slice-Commit.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13); der Bericht allein ist keine Adresse.
- Der Implementer hält an einer Randform ohne Ort in der Spezifikation an und gibt sie an den Architect zurück (`AGENTS.md` §3.12); er entscheidet sie nicht im Code.
- Wert mit Escape-Folge oder Rohbytes: der Plan und die Berichte schreiben eine solche Folge nicht als Zeichen, sondern beschrieben (Werkzeug-Falle Backslash-u, `AGENTS.md` §3.18).

**Aufträge je Rolle** (Planner, beim Schnitt vom 2026-10-30; sie gelten für `slice-telemetry-log` und `slice-telemetry-spans` sinngemäß dort):

- *Implementer:* P15 und P16 sind nebenläufig; ein Datenrennen zählt der Detektor von `make test-race` je Prozess (`-test.count=1`), die Wirkung (ein verlorener Zuwachs, eine zu kurze Aufzeichnung) eine Wiederholung (`-count=50`). Der Bericht nennt je Mutation die Rate als *n* von *n* bei 50 Prozessen mit `-test.count=1` unter `make test-race` und mit `-count=50`, die Zahl der Prozesse und das Messende, und der Implementer trägt sie mit der Tabelle in einen Absatz von §3 ein (`AGENTS.md` §3.10). Die Schleifen P02 bis P05 laufen als Schleifen über die Mengen mit Orakel und Zählprüfung (24 Schlüssel, 9 Formen, 72 Paare, 33 Codes), ein Orakel ist nicht der Code des Adapters. `FakeTelemetry` ist eine Hülle um nichts, aber ein Fake mit Argumenten: er hält jedes Argument fest, zählt jede Operation (auch `Snapshot` und das zweite `End`), gibt den `ctx` mit derselben Identität weiter und verdeckt keine Eigenschaft des Originals; eine Panik des Fakes färbt den benannten Test. Die Zusatzzeit der Stufen `test` und `race` (vorher und nachher, Zahl der Läufe, Zustand des Hosts) steht in §7; die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) liegt beim User, nicht im Slice.
- *Reviewer:* löst jede Zeile der Tabelle und jede Zelle des Rasters mit einer Mutation auf, fährt P15 und P16 selbst mit Rate, prüft die Schleifen gegen die Klasse (jedes Zeichen, jede Grenze, jeder Code) und den Fake auf verdeckte Eigenschaften; eine im Code entschiedene Randform ist ein MEDIUM (`AGENTS.md` §3.16). Er prüft zusätzlich, dass der Kern nichts enthält, was `slice-telemetry-log` oder `slice-telemetry-spans` gehört (kein Schreiben der Zeile, keine Wirkung von `StartSpan`).
- *Verifier:* prüft die Tabelle gegen `SPEC-230` Nr. 1 bis 4, 8 bis 10 und `SPEC-232` Nr. 1 und 5 a, dass `go.mod` und `go.sum` gegenüber der Basis leer im Diff sind, dass das Binary unverändert blieb, und die Messpunkt-Zeile (Handoff und Ende, `SPEC-128` Nr. 9).
- *Planner (Closure):* setzt die Messpunkt-Zeile, die Auslastung, meldet dem Architect, liegt sie über 1,00 oder unter 0,60, und legt dem User die Diff-Vorlage zu `AGENTS.md` §3.14 vor (§6).

**Rollenfolge** (`.claude/agents/`, kein Self-Review): Architect (Vertrag, Tabelle, Raster, Block `groessenmarke`; keine ADR, keine Bibliothek) → Planner (Schnitt, Marke, §4) → Implementer (Code, Mutationen selbst rot, Absatz zu Zusagen und Raten in §3, ein Handoff-Commit) → Reviewer (Report unter `docs/reviews/`) → Fix-Runden (Menge je Befund, `AGENTS.md` §3.17) → Verifier → Planner (Closure mit Lerneintrag; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorlage an den User) → Integrationsschritt (Merge, Register, Roadmap, Stempel, `make gates` im Hauptbaum).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`, 2026-10-30, Linie D): kein anderer Slice muss in `done/` liegen. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Entscheidung zum Trigger der Welle (Planner, Auftrag des Users vom 2026-10-30, produktive Slices laufen parallel):** Der Start-Trigger von [welle-betrieb](../welle-betrieb.md) §2 (`welle-mock` in `done/`, M3 erreicht) wird für **diesen** Slice gelockert, die Welle bleibt unverändert. Grund: Code-Unabhängigkeit — der Slice braucht nur `src/hexagon/ports/support` (`slice-clock-idgenerator`) und die Fehlercode-Tabelle (`slice-fehlermodell-konfiguration`), beide in `done/`, und ruft weder `MockUseCase` noch `WorkspaceUseCase`; die Mock-Spans sind Namen als Daten (siehe unten). Der Trigger der Welle bindet den Einbau in Mock und Workspace und den Lasttest, nicht den Port. Die Lockerung gilt nur, solange der Slice **ohne eine Wirkung am Binary** bleibt: Die Handbuch-Pflicht der Welle ([welle-betrieb](../welle-betrieb.md) §4) greift genau dann, wenn der Slice einen neuen Code, Schlüssel oder ein Kommando in den Katalog des Binary bringt oder Logfelder am Binary sichtbar macht; dann wartet er auf `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` in `done/` (Rückführung `in-progress` → `open`, unten). Der Architect hat entschieden, dass keiner der drei Fälle eintritt (Entscheidung 7); der Plan ändert `src/bootstrap` nicht und berührt damit `slice-bootstrap-m2` nicht. Die Closure-Trigger der Welle (§3 dort) bleiben; die Roadmap und `welle-betrieb.md` schreibt der Integrationsschritt nach (§6).

**Abhängigkeiten — Grund je Eintrag:**

- Code, den der Slice braucht und der schon in `done/` liegt: `src/hexagon/ports/support` und `supporttest` (`slice-clock-idgenerator`), die Fehlercode-Tabelle in `src/hexagon/model` (`slice-fehlermodell-konfiguration`), der Katalog und die Tabelle der Spans in der Spezifikation. Mehr braucht ein Port mit Adapter nicht.
- Handbuch-Punkt der DoD (nur wenn er greift, siehe DoD): `slice-handbuch-geruest` (Ort der Kapitel) und `slice-handbuch-meldungskatalog` (das Kataloge-Gate) aus `welle-benutzerhandbuch` liegen dann in `done/`; ohne sichtbare Änderung am Binary entfällt die Bedingung.

**Unabhängig von `welle-mock` / M3, von M2 und von Linie A, weil** der Port die Namen der Spans als Daten kennt und nichts ruft: `MockUseCase`, `WorkspaceUseCase`-Mutation, HTTP-Adapter und Replay-Dienst sind Orte des Einbaus (`slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`), nicht des Ports. Der Start-Trigger der Welle (`welle-mock` in `done/`, [welle-betrieb](../welle-betrieb.md) §2) gilt für den Abschluss mit Mock-Spans und Lasttest; dieser Slice hängt an keinem Code daraus. Reihenfolge-Vorsicht, kein Blocker: der Slice ändert `src/bootstrap` nicht (die Verdrahtung ist Einbau), so dass er die Dateien der Slices von M2 (`slice-bootstrap-m2`) nicht berührt.

**Größenmarke dieses Teils:** Der Architect hat die Teile des Blocks geschrieben, die ihm gehören (`pfad`, `produktcode`, `fakes`, ungeschnitten für alle drei Teile: 1310 Zeilen Produktcode, zwei Fakes); der Planner hat sie beim Schnitt auf den Teil `P` verteilt und `basis`, `schaetzung` und `marke` in eigenem Commit vor dem ersten Code-Commit gesetzt (`SPEC-128`: Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Die Klasse `nebenläufig` hat der Architect je Zeile gesetzt (`SPEC-128` Nr. 3 e): in diesem Teil P15 und P16.

Rechnung dieses Teils: *a* = 2 (P15, P16), *b* = 15 (P01 bis P14, P17), *f* = 1 (`FakeTelemetry`; `FakeSenke` liefert `slice-telemetry-log`), *p* = 580, also *S* = 85·2 + 80·15 + 200·1 + 580 = 2150 und *M* = 100·⌈135·2150/10000⌉ = 100·⌈29,03⌉ = 3000. Die Schnittfrage (*S* ≥ 2223) stellt sich für diesen Teil nicht.

```groessenmarke
pfad: src/hexagon/ports/support
pfad: src/adapters/driven/telemetry
produktcode: 580
fakes: 1
basis: a0396b0cb0769476edaeb5bfaa23dcc5a76c041c
schaetzung: 2150
marke: 3000
```

`basis` ist der Commit des Schnitts (der Vertrag, die Tabelle und das Raster stehen); gegen `main` mit dem Stand `afd8cf2` und gegen den Architect-Stand `a5fbbc9` ergibt `git diff --numstat` über die beiden Pfade 0 Datensätze, die Messung vor dem Code ist 0, und `make groessenmarke` meldet für den Plan die Marke 3000 ohne Befund. `fakes` zählt `FakeTelemetry`; `FakeSenke` liefert `slice-telemetry-log`.

**Schnitt vor dem Code — Entscheidung des Planners am 2026-10-30 (a: geschnitten).** Der Architect hat die Schnittfrage gestellt: die ungeschnittene Tabelle hat 53 Zeilen, 46 `andere` (*b*) und 7 `nebenläufig` (*a*), dazu *f* = 2 und *p* = 1310, also *S* = 85·7 + 80·46 + 200·2 + 1310 = 5985 und *M* = 8100. Der Planner hat nachgezählt (Zeilen je Präfix: `P` 17, `L` 19, `S` 17; nebenläufig P15, P16, L18, L19, S15, S16, S17; Fakes `FakeTelemetry` in `P`, `FakeSenke` in `L`, keiner in `S`; Produktcode 580 + 380 + 350 = 1310) und das Ergebnis bestätigt: 53 Zeilen, 7 und 46, *S* = 5985. Das liegt über der Schwelle (*S* ≥ 2223, `SPEC-128` Nr. 1), und kein Zuschnitt der Tabelle ohne Verlust von Zusagen senkt *S* unter sie (Bündeln von Zeilen senkte die Schätzung, nicht die Arbeit, `SPEC-128` Nr. 10 b). Geschnitten wird vor dem Code in drei Slices entlang der Abhängigkeitskette; jeder Teil liegt unter der Schwelle, baut auf dem vorigen auf und teilt mit dem nächsten keinen Mechanismus, den er nicht schon geliefert hat. Die Nr. der Tabelle tragen das Präfix des Teils, so dass die Zeilen ohne Umbau auf die Pläne verteilt wurden:

| Teil | Inhalt | Zeilen *b* / *a* | *f* | *p* | *S* | *M* |
|---|---|---|---|---|---|---|
| `P` (dieser Slice, `slice-telemetry-port`) | Ports `Telemetry` und `TelemetryReader` mit allen Typen, Regeln der Attribute (Schlüssel, Wertformen, Verwurf), `CountError`, `Snapshot`, `New`, `FakeTelemetry` | 15 / 2 | 1 | 580 | 2150 | 3000 |
| `L` (`slice-telemetry-log`, nach `P`) | Zeile, Stufen, `msg`, `Handler`, Abbildung der zehn Ereignisse der Dienste, Ausfall der Senke, Canary über `Log` und `Handler`, Sperre der Zeile | 17 / 3 | 1 | 380 | 2195 | 3000 |
| `S` (`slice-telemetry-spans`, nach `L`) | `StartSpan`, `Set`, `End`, Eltern-Kind, Verteilung, Ring, Span-Zeile, Canary über Spans, Sperre des Stands | 14 / 4 | 0 | 350 | 1810 | 2500 |

Die Zeilen *b* / *a* von `L` und `S` und ihre *S* und *M* zählen den Stand nach der Entscheidung des Architect zu `SPEC-232` Nr. 5 f (L20 und S18 kamen hinzu: insgesamt 55 Zeilen, 9 `nebenläufig`); die Zählung im Absatz davor ist die der Entscheidung vom Schnitt. `L` liegt mit *S* = 2195 knapp unter der Schwelle von 2223.

Die Naht „Logs und Spans einerseits, Metriken und Zähler andererseits“ trägt nicht: das Ende eines Spans füllt Verteilung und Zeile zugleich, und die Regeln der Attribute gelten für Zeilen und Spans; die Kette `P` → `L` → `S` schneidet an den Stellen, an denen der Code des vorigen Teils schon steht (`L` braucht `New`, die Regeln und `Dropped` aus `P`, `S` braucht den Zeilenschreiber aus `L`). Der Port (`Telemetry` mit `StartSpan`) steht in `P` vollständig, der Adapter liefert in `P` einen wirkungslosen `StartSpan` bis `S`. Die Marken der Teile (`L` 3000 und `S` 2500 setzt der Planner beim Start des jeweiligen Slice neu, `basis` dann der Stand nach dem Vorgänger) summieren sich zu 8500 und liegen über der des ungeteilten Slice (8100), weil jeder Teil seinen festen Anteil trägt; kein Teil liegt über 3000. Folgen für andere Pläne, die der Planner im Integrationsschritt nachzieht (in `open/` schreibt dieser Zweig nichts): `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` starten nach `slice-telemetry-spans`; `slice-lasttest` liest die Span-Zeile aus `slice-telemetry-spans`; `slice-inspection-use-case` braucht nur diesen Teil `P`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke dieses Teils (3000, Messung nach `SPEC-128`).
- `in-progress` → `open` (blockiert): die Wahl einer Bibliothek (nicht gewählt, Entscheidung 1) oder der Slice bringt einen Code, Schlüssel oder ein Kommando in den Katalog des Binary und `slice-handbuch-geruest` oder `slice-handbuch-meldungskatalog` liegen nicht in `done/` (Lockerung des Triggers, siehe oben; nicht erwartet, Entscheidung 7).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Abrufweg der Metriken und Spans war nicht spezifiziert; ohne Entscheidung wäre der Lasttest (`slice-lasttest`, DoD: das gebaute Binary als eigener Prozess) nicht lesbar und `InspectionUseCase` erreichte den Zähler nicht — **Ausgang:** entfallen; entschieden in `SPEC-230` Nr. 9 (Lese-Teil `TelemetryReader`) und `SPEC-231` Nr. 6 (Span-Zeile im Log), Entscheidung 1. Der Nehmer `slice-lasttest` führt den Weg noch nicht (siehe Auftrag unten).
- Eine Bibliothek für Spans und Metriken wäre eine neue Abhängigkeit — **Ausgang:** entfallen; Standardbibliothek genügt (Entscheidung 1), `go.mod` und `go.sum` bleiben (Fänger P14), keine ADR.
- Spezifikationslücke: §5 nannte Spans, keine Metriken — **Ausgang:** entfallen; `SPEC-230` Nr. 8 trägt Namen, Art, Einheit, Schlüssel und Eimer. Das Lastenheft führt weiter keine Anforderung an Metriken oder Spans außer der Messmethode von `LH-QA-01`; die Frage dazu steht unten.
- Frage an den User (Anforderung des Owners, nicht blockierend): Das Lastenheft nennt „reguläre Logs“ (`LH-QA-04`), aber keine Beobachtbarkeit von Spans und Metriken. Die Span-Zeile (`SPEC-231` Nr. 6) macht Spans außerhalb des Prozesses lesbar, ohne Endpunkt, Netz-Port oder Bibliothek; sie ist die kleinste Lösung, die `slice-lasttest` erreicht. Braucht das Produkt einen Abruf für Anwender (Endpunkt, Export an einen Collector), ist das eine Anforderung im Lastenheft — **Ausgang:** weiter offen; Adresse: User über den Planner.
- Sendung an `slice-inspection-use-case` (`welle-admin-api`, Datei besteht seit dem Merge von `main`): der Zähler je Fehlercode ist der Gegenstand von `/__admin/diagnostics`. Der Nehmer führt die Zähler in seiner DoD und Auftrag 5, nennt aber „den Port `Telemetry`“ und eine Operation, die Zähler liest, als Anforderung an diesen Slice (sein §6, Ausgang *offen*, Adresse Planner von `welle-betrieb`). Die Antwort des Architect: das Lesen ist der **eigene** Port `TelemetryReader` mit `Snapshot().Errors`, ein Eintrag je Code der Tabelle (33), auch mit 0 (`SPEC-230` Nr. 9); `Telemetry` hat keine Leseoperation, und der Fake der Diagnose ist `FakeTelemetry` (beide Ports). Wer zählt, ist der Einbau (`SPEC-230` Nr. 8). Mit dem Schnitt (§4) liegt der Lese-Teil in diesem Teil `P` (`Snapshot().Errors` mit 33 Einträgen und `CountError` sind hier geliefert), so dass der Start von `slice-inspection-use-case` nur `slice-telemetry-port` in `done/` braucht, nicht `slice-telemetry-log` und nicht `slice-telemetry-spans` — **Ausgang:** weiter offen; Adresse: Planner (§1 oder DoD des Nehmers auf `TelemetryReader` und `Snapshot().Errors` ziehen, `AGENTS.md` §3.13; in `open/` schreibt dieser Zweig nichts).
- Sendung ohne Nehmer: Umstellung der Dienste `session` und `maintenance` (und `record`) auf den Handler — **Empfehlung des Architect:** Nehmer `slice-telemetry-einbau-gateway`, weil er als erster den Adapter im Composition Root erzeugt (nach `slice-bootstrap-m2`, das den `*slog.Logger` anlegt) und `slog.New(Handler())` allen drei Diensten gibt, ohne dass ihr Code sich ändert (`SPEC-231` Nr. 5, 7); die Dienste `session` und `maintenance` sind damit ohne eigenen Slice umgestellt. **Ausgang:** weiter offen; Adresse: Planner (Datei von `slice-telemetry-einbau-gateway` um die Verdrahtung ergänzen, bevor dieser Slice sie nennt; in `open/` schreibt dieser Zweig nichts).
- Sendung ohne Nehmer, vom Geber `welle-hybrid` ([welle-hybrid](../welle-hybrid.md) §6): Spans und Zähler des Hybrid-Fallbacks. Der Port schließt sie nicht aus: ein Span ist eine Zeile in §5 und in `support.Spans()` (`SPEC-230` Nr. 5), ein Zähler eine Folge-Entscheidung (eine Operation oder eine Reihe in `SPEC-230` Nr. 8). `slice-hybrid-service` schließt „Spans, Audit und Metriken des Fallbacks“ ausdrücklich aus (sein §1, Adresse `welle-betrieb`) und ist damit keine Adresse (`AGENTS.md` §3.13). **Empfehlung des Architect:** ein neuer Slice `slice-telemetry-einbau-hybrid` in `welle-betrieb`, nach `slice-hybrid-service` und nach `slice-telemetry-spans` (Teil `S` des Schnitts, §4). Der Planner legt ihn nicht an, bevor der User entschieden hat (Meldung mit Empfehlung im Bericht des Schnitts). **Ausgang:** weiter offen; Adresse: User über den Planner (derselbe Punkt steht in §6 von `slice-telemetry-einbau-gateway` und `slice-telemetry-spans`).
- Auftrag an den Planner für `slice-lasttest` (Auftrag 2 dort): der Abruf des Gesamtspans ist die Span-Zeile im Log des Prozesses (`SPEC-231` Nr. 6: `msg` gleich `odata.gateway.request`, `duration_ms` mit drei Nachkommastellen, `span_id`, `parent_id` für den Anteil von `odata.replay.match` und `odata.mock.query`) oder, wenn der Lasttest im Prozess läuft, `TelemetryReader` — **Ausgang:** weiter offen; Adresse: Planner (Risiko-Ausgang im Plan des Nehmers nachziehen).
- Auftrag an den Planner für `slice-telemetry-einbau-gateway` (`SPEC-121` Nr. 12): der Satz „einen Port `Telemetry` gibt es noch nicht“ in `SPEC-121` Nr. 12 wird mit dem Einbau veraltet; der Wortlaut von Nr. 12 gehört dem Einbau-Slice (diese Entscheidung ändert ihn nicht) — **Ausgang:** weiter offen; Adresse: Planner.
- Auftrag an den Planner für das Handbuch: der Einbau, der den Adapter verdrahtet, macht die Logfelder (und die Span-Zeile) am Binary sichtbar; die Handbuch-Pflicht der Welle greift dort (Entscheidung 7), nicht in diesem Slice — **Ausgang:** weiter offen; Adresse: Planner (`slice-telemetry-einbau-gateway`, Abhängigkeiten `slice-handbuch-geruest`, `slice-handbuch-meldungskatalog`).
- Schnittfrage: *S* = 5985, Marke 8100 ohne Schnitt (§4) — **Ausgang:** entfallen; der Planner hat vor dem ersten Code-Commit in drei Slices geschnitten (`slice-telemetry-port` mit Marke 3000, `slice-telemetry-log` 3000, `slice-telemetry-spans` 2500, die beiden letzten Dateien in `open/`).
- Auslegung des Planners beim Schnitt (§3): der Kern dieses Slice enthält ein `Log`, das die Regeln der Attribute anwendet und `Dropped` zählt, aber keine Zeile schreibt, und ein `StartSpan` ohne Wirkung; die Zeile 5 der Zusatztabelle ist von P17 nach S02 gewandert — **Ausgang:** entschieden; Ort: `SPEC-232` Nr. 1 und §3 dieses Plans (bestätigt vom Architect; `Log` wendet `MinLevel` nicht an, `MinLevel` gehört zu L04).
- Spezifikationslücke, vom Planner beim Verteilen der Tabelle bemerkt: `SPEC-232` Nr. 5 f („Die Sperre wird nicht über `Clock.Now()` gehalten“) hatte in der Tabelle keine Zeile (`AGENTS.md` §3.15); der Wortlaut ist jetzt „Keine Sperre des Adapters wird über `Clock.Now()` gehalten“ — **Ausgang:** entschieden; Fänger L20 in `slice-telemetry-log` und S18 in `slice-telemetry-spans`; die Seite von P hat keinen Aufruf von `Clock.Now()` (`CountError` ruft keine `Clock`, `Log` schreibt hier keine Zeile, `StartSpan` ist ohne Wirkung).
- Raster `New` (Zelle „unbekannter oder fremder Wert“, `MinLevel`) — **Ausgang:** entschieden; Ort: `SPEC-232` Nr. 1 (jeder `MinLevel` ist gültig); Fänger P10.
- Die Einbau-Slices (`slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`) starten nach `slice-telemetry-spans`, nicht nach diesem Slice, weil sie Spans rufen; `slice-lasttest` liest die Span-Zeile aus `slice-telemetry-spans` — **Ausgang:** weiter offen; Adresse: Planner (Abhängigkeiten der drei Dateien nachziehen, `AGENTS.md` §3.13; in `open/` schreibt dieser Zweig nichts).
- Spans und Zähler des Hybrid-Fallbacks und die Verdrahtung samt Umstellung von `session` und `maintenance` setzen auf `slice-telemetry-log` und `slice-telemetry-spans` auf; die Aufträge oben bleiben mit ihren Adressen.
- Der Stand von `AGENTS.md` §3.14 (Schätzung × 1,25, Faktoren 85 und 40) weicht vom Vertrag `SPEC-128` (Faktoren 85, 80, 200, Zuschlag 135 vom Hundert) ab; die Marke dieses Slice folgt `SPEC-128` — **Ausgang:** weiter offen; Adresse: User (Diff-Vorlage für `AGENTS.md` im Closure des Planners; Änderungen an `AGENTS.md` schreibt dieser Zweig nicht).
- Nachzüge für den Integrationsschritt (Roadmap, `welle-betrieb.md`, Übergangsregeln, Register, Historie-Zeile in spezifikation.md §7): Linie D mit Kennungsbereichen (belegt: `SPEC-230` bis `SPEC-232`), Start dieses Slice vor dem Welle-Trigger, die Slices des Schnitts, Register-Zähler — **Ausgang:** weiter offen; Adresse: Planner im Integrationsschritt. Die Historie-Zeile trägt dieser Zweig nicht ein, weil alle Zweige dieselbe Stelle der Tabelle schreiben.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten ist gefüllt, jede Zelle trägt ihren Ort — **Ausgang:** entfallen; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect und steht dann hier.

**Prüfung der Nehmer durch den Planner** (`AGENTS.md` §3.13, beim Start gelesen: §1 *Ausdrücklich NICHT* und DoD): (1) `slice-telemetry-einbau-gateway` nimmt den Einbau in `record` und `replay` mit der Kennung dieses Slice an (*Übernimmt*, DoD: Gesamtspan, Capture- und Match-Span, Protokolleintrag über den Port mit unverändertem Wortlaut); kein *Ausdrücklich NICHT* trifft ihn; die Verdrahtung und die Umstellung von `session` und `maintenance` nimmt er noch nicht an (oben). (2) `slice-telemetry-einbau-mock` nimmt den Einbau in `mock` und Workspace-Mutation an; kein Ausschluss trifft ihn, die Werte von `recorder.mutation_type` liegen bei ihm (Entscheidung 2, Form Token). (3) `slice-audit-log` schließt das Diagnose-Log und die Zähler aus und nennt diesen Slice — Gegenstück stimmt. (4) `slice-lasttest` liest den Gesamtspan und nennt den Abrufweg als offenes Risiko mit Adresse Architect — dieser Slice liefert die Entscheidung, die Datei des Nehmers bleibt unverändert (oben). (5) `slice-inspection-use-case` besteht (nach dem Merge von `main`) und führt die Zähler, nennt aber den Port `Telemetry` statt `TelemetryReader` (oben); die Slices `slice-admin-*` führen den Punkt nicht. (6) Der Umstieg der Dienste `session` und `maintenance` und die Spans und Zähler des Hybrid-Fallbacks haben keinen Nehmer (oben, mit Empfehlung); `slice-hybrid-service` schließt die Spans aus.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). `SPEC-230` bis `SPEC-232` steht für die Nummern in spezifikation.md §5. Dieser Teil trägt die Zeilen `CountError`, `Snapshot` und `New`; die Zeilen `Log` und `Handler` stehen in `slice-telemetry-log`, die Zeilen `StartSpan`, `Set` und `End` in `slice-telemetry-spans`. Die Regeln der Schlüssel und Werte, die `Log` in diesem Teil anwendet, fängt P03 bis P05 und P17; die Zelle der Zeile `Log`, die in diesem Teil gilt, ist „leerer Wert“ und „unbekannter oder fremder Wert“ (`SPEC-230` Nr. 3 und 4).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | sensibler Wert | Abbruch (`ctx`) | Wiederholung |
|---|---|---|---|---|---|---|
| Erzeugen (`New`) | `SPEC-232` Nr. 1 (nil-`Writer`, nil-`Clock`: `model.ErrUngueltig`) | `SPEC-232` Nr. 1 (`MaxSpans` 0 gilt als 1024) | `SPEC-232` Nr. 1 (`MaxSpans` unter 0 oder über 65536; jeder `MinLevel` ist gültig) | `SPEC-232` Nr. 1 (der Fehlertext nennt keinen Wert) | gilt nicht: kein `ctx` (`SPEC-230` Nr. 1) | gilt nicht: jeder Aufruf erzeugt einen eigenen Adapter ohne gemeinsamen Zustand (`SPEC-232` Nr. 1) |
| Zähler erhöhen (`CountError`) | gilt nicht: die Operation hat nur das Argument `code`, ein fehlender Wert ist `0` (`SPEC-230` Nr. 8) | gilt nicht: `0` ist der Nullwert, siehe unbekannter Wert | `SPEC-230` Nr. 8 (Code außerhalb der Tabelle: `fehlercode`) | gilt nicht: das Argument ist ein `model.Code`, keine Zeichenkette; die Schlüssel der Reihe sind die 33 Codes (`SPEC-232` Nr. 2) | `SPEC-230` Nr. 1 (kein `ctx`; wirkt unabhängig davon) | `SPEC-230` Nr. 8 (jeder Aufruf um 1); `SPEC-232` Nr. 5 a (kein Verlust) |
| Lesen (`Snapshot`) | gilt nicht: kein Argument | `SPEC-230` Nr. 9 (frischer Adapter: alle Schlüssel mit 0, `Spans` leer) | gilt nicht: kein Argument | gilt nicht: der Stand enthält nur, was die Formen zuließen (`SPEC-232` Nr. 2) | gilt nicht: kein `ctx` (`SPEC-230` Nr. 1) | `SPEC-230` Nr. 9 (ändert und zählt nichts; Kopie je Aufruf) |

**Zusatztabelle: Verstöße am selben Schlüssel** (`AGENTS.md` §3.16; `SPEC-230` Nr. 4, Rangfolge Schlüssel vor Wert, genau ein Verwurf je Attribut):

| Paar | Eingabe | Stand danach | Fänger |
|---|---|---|---|
| unbekannter Schlüssel mit ungültigem Wert | `Log` mit `timestamp` und einem Wert der Form Leerraum | `schluessel` 1, `wert` 0 | `TestVerwurfRangfolge` Zeile 1, P17 |
| doppelter Schlüssel, zweiter Wert ungültig | `session_id` zweimal, zweiter mit Leerraum | erster gilt; `schluessel` 1, `wert` 0 | Zeile 2, P17 |
| doppelter Schlüssel, erster ungültig, zweiter gültig | `session_id` zweimal, erster mit Leerraum | zweiter gilt; `wert` 1, `schluessel` 0 | Zeile 3, P17 |
| doppelter Schlüssel, beide gültig und verschieden | `session_id` zweimal | erster gilt; `schluessel` 1 | Zeile 4, P17; L07 in `slice-telemetry-log` für alle 13 Schlüssel |
| fremder Schlüssel, im selben Aufruf doppelt | `Log` mit `x` zweimal | `schluessel` 2 (beide Attribute zählen einen Verwurf) | Zeile 6, P17 |

Die Zeile „Pflicht-Attribut mit ungültigem Wert, dann `End`“ (`wert` 1, `pflicht` 1) braucht `StartSpan` und `End` und steht mit ihrem Fänger S02 in `slice-telemetry-spans`.

**Zusatztabelle: Modus** (`AGENTS.md` §3.16; die Zusage hängt am Modus): in diesem Teil hängt keine Zusage an einem Modus. Die Modi `MinLevel` und Senke defekt und der Weg `Log` oder `Handler` stehen mit ihren Fängern L04, L10, L14 in `slice-telemetry-log` und S13 in `slice-telemetry-spans`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register beim Start gesichtet (2026-10-30, Zähler aus den Dateien unter `evidence/` der Basis `f4d9ac3`; neunzehn Einträge in `BEO-REPO`). Der Adapter ist nebenläufig und der Fake umhüllt `ctx`: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) gilt für Pfade, die kein Lauf ausführt (Ausfall der Senke, Abbruch), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für den Fake des Ports, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für jede Wirkung, die der Detektor nicht meldet (Zähler unter Last: Rate mit Zahl der Läufe, `AGENTS.md` §3.10); offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1× in der Basis dieses Zweigs; laut Auftrag des Users im Hauptbaum und den anderen Zweigen 3×, dann ist die Schwelle erreicht und die Verkörperung Sache des nächsten Lese-Schritts des Integrationsschritts — dieser Zweig setzt keinen Zähler). Zusätzlich für die Fänger: `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `implement-slice` Schritt 14), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (7×, `AGENTS.md` §3.10) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16) für die Allaussagen „kein Credential in keinem Log“ und „der Zähler verliert nichts“. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16) und `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×, `AGENTS.md` §3.14). Von den offenen erreicht keiner mit diesem Slice 3×, außer dem Eintrag zu den Betriebsregeln nach Stand des Hauptbaums.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

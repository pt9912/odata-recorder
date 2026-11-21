# Slice slice-telemetry-einbau-mock: Einbau von `Telemetry` in Mock-Dienst (Query-Span) und Workspace-Mutation (Mutations-Span)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, `LH-FA-07`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-290` (neu: Telemetrie des Mock-Dienstes) · `SPEC-029`, `SPEC-030` (spezifikation.md §5) · `SPEC-139` Nr. 3, 11, 13 (Adapter des Prozesses, `Log` des Dienstes `workspace`) · `SPEC-231` Nr. 4, 7 (Einträge des Dienstes `workspace`) · `SPEC-129` Nr. 13 · `SPEC-195` Nr. 1, 11 · `SPEC-192` Nr. 9 · `SPEC-230` Nr. 8 · `LH-QA-01.a` (Mockserver: Zerlegung des Gesamtspans durch `odata.mock.query`) · `LH-FA-07.b` (Vertragsmatrix der Mutationen) · `LH-FA-11.a` (Coverage-Status) · `LH-QA-05.a` (Fehlercodes) · `SPEC-044` · `ARC-002`, `ARC-003`, `ARC-021` · `architecture.md` §4.4, §4.5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `MockUseCase` setzt den Span `odata.mock.query` (`SPEC-029`, `SPEC-290`, Attribute `recorder.entity_set`, `recorder.coverage_status`) um die Query-Ausführung einer Abfrage (Collection, Einzel-Lookup, `/$count`; ab dem Lesen des Service bis zur Rückkehr von `Mock`, nicht `$metadata`) und den Span `odata.workspace.mutation` (`SPEC-030`, Attribute `recorder.workspace_id`, `recorder.mutation_type`) um jede der vier Mutationen `POST`, `PATCH`, `PUT` und `DELETE` des Mock-Dienstes, samt Transaktion; der Dienst zählt keinen Fehler, den Zähler je Fehlercode führt der Adapter, der das Fehlerobjekt schreibt (`SPEC-290` Nr. 5); der Composition Root verdrahtet den Adapter des Prozesses für den Mock (Modi `mock` und `hybrid`) und als Logger des Dienstes `workspace` (`SPEC-139` Nr. 13). Kein Query-Literal (`$filter`-Werte), kein Body und kein Key-Wert wird Attribut.

**Übernimmt (Sendung von `slice-telemetry-port`):** der Einbau in die Pfade `mock` und Workspace-Mutation.

**Übernimmt (Sendung von `slice-telemetry-log`, akzeptiertes Negativ L15; Geber: `slice-telemetry-log`; Eintrag des Planners vom 2026-11-07):** der Fänger je Stelle: ein wohlgeformter Wert in einem zulässigen Schlüssel geht durch die Form des Adapters (`SPEC-232` Nr. 2, *Ungeprüft*), der Schutz ist ein Test je Stelle des Mock-Pfads, dass kein Wert aus dem Request (Entity-Set, Workspace-ID, Key-Wert, Query-Literal) in einem Attribut oder in `msg` einer Zeile landet, außer nach Prüfung gegen das Modell.

**Übernimmt (Sendung von `slice-telemetry-handler`, Rückgabe 2 des Architect; Geber: `slice-telemetry-handler`; Eintrag des Planners vom 2026-11-07):** die Einträge des Dienstes `workspace` (`SPEC-129` Nr. 13: `bereinigung_fehlgeschlagen`, `bereinigung_zurueckgestellt`): der Dienst schreibt im Composition Root über denselben `*slog.Logger` wie `record`, `session` und `maintenance`, der Handler behandelt sie wie die zehn Einträge der Menge M13; `SPEC-231` Nr. 4 und 7 nennen sie seit dem Vertrag dieses Slice (Architect, 2026-11-21, Auftrag 7); die Verdrahtung im Composition Root (`SPEC-139` Nr. 13 b) und der Fänger je Stelle (Request-Wert in `msg` oder Schlüssel) folgen hier (Zeilen Q17 und Q18). Der Ausschluss unten zu `audit_fehlgeschlagen` berührt sie nicht.

**Übernimmt (Sendung von `slice-mock-schreiben`, Rückgabe R2 / W66, Review F-15; Geber: `slice-mock-schreiben`; Eintrag des Planners vom 2026-11-07, §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung):** der Fänger an den Wegen Log und Span für die Länge des Ortes im `Error()` des Eingabe-Lesers: ein Canary „Name einer unbekannten Property von `grenze` Bytes“ (`SPEC-180` Nr. 3, Obergrenze 256 Bytes des Ortes) geht durch den Mutations-Pfad des Mock-Dienstes (`POST`, `PATCH`, `PUT`), und weder eine Log-Zeile (`msg`, Attribute) noch ein Span trägt mehr als den Code und die gekürzte Meldung. Der Dienst schreibt selbst nichts in Log, Span oder Audit (W66, akzeptiertes Negativ dort).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port und Adapter `Telemetry` — `slice-telemetry-port`.
- Zeile, Stufen, Senke, `Handler`, Spans, Verteilung, Ring und Span-Zeile im Adapter — `slice-telemetry-log`, `slice-telemetry-handler` und `slice-telemetry-spans`; alle drei liegen beim Start dieses Slice in `done/`, dieser Slice ändert sie nicht.
- Der Eintrag `audit_fehlgeschlagen` und die Abbildung des Adapters `audit` (`SPEC-231` Nr. 4 und 7, Sendung von `slice-audit-log`) — `slice-telemetry-einbau-gateway`; die Löschung eines Workspace ist kein Mutations-Span.
- Gesamtspan, Capture-Span und Match-Span (`SPEC-026` bis `SPEC-028`) — `slice-telemetry-einbau-gateway`; der Gesamtspan liegt im HTTP-Adapter und ist modusunabhängig.
- Spans und Zähler des Hybrid-Fallbacks — kein Slice von `welle-betrieb` nimmt sie an (Sendung ohne Nehmer, siehe §6 von `slice-telemetry-einbau-gateway`, Adresse: Planner).
- Mutations-Spans der Admin-Operationen (`Reset`, `Restore`, `Create`, `Delete` des Workspace, Snapshot-`Create`): keine Operation von `WorkspaceUseCase` oder `SnapshotUseCase` trägt einen Span (Entscheidung des Architect, `SPEC-290` Nr. 2); ein Vertrag, der `Reset` oder `Restore` einbezieht, ergänzt die Werte von `recorder.mutation_type`, das gehört `welle-admin-api`, falls sie es will.
- Ein Span für `$metadata` — Entscheidung des Architect (`SPEC-290` Nr. 2): keine Query-Ausführung, die Zeit steht im Gesamtspan.
- Der Zähler je Fehlercode für die Fehler des Mock-Pfads — der HTTP-Adapter, der das Fehlerobjekt schreibt (`slice-telemetry-einbau-gateway`; sein Ziel trägt „jeden Fehler, den der Adapter als Fehlerobjekt schreibt“ und seine DoD den Zähler je Fehlercode, geprüft am 2026-11-21); der Dienst `mock` ruft `CountError` nie (`SPEC-290` Nr. 5, `SPEC-230` Nr. 8: die ausliefernde Stelle zählt; sonst zählte jeder Fehler doppelt und im Modus `hybrid` auch der, den der Hybrid-Dienst verwirft).
- Die Messung des Mockservers (`LH-QA-01.a`) — `slice-lasttest`; dieser Slice liefert den Span, der die Query-Ausführung ausweist.

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

- [ ] Query-Span: jede Abfrage des Mock-Dienstes (Collection, Key-Lookup, `/$count`, mit `GET` und `HEAD`; `$metadata` nicht, `SPEC-290` Nr. 2), die die Zuordnung bestand, öffnet und beendet `odata.mock.query` mit Entity-Set (Name aus dem Modell oder der Platzhalter) und Coverage-Status (vier Werte), auch im Fehlerfall (`ODR-E6003`, `ODR-E6001`, Abbruch); der Span umfasst die Query-Ausführung bis zur Rückkehr von `Mock`, nicht die Serialisierung außerhalb (`LH-QA-01.a`).
- [ ] Mutations-Span: jede Mutation des Mock-Dienstes (`POST`, `PATCH`, `PUT`, `DELETE`, `LH-FA-07.b`), die die Zuordnung bestand, setzt `odata.workspace.mutation` mit der Kennung des gelesenen Workspace und der Art (`post`, `patch`, `put`, `delete`) um die Transaktion; eine abgewiesene Mutation (Constraint, Coverage) trägt den Span mit ihrem Ergebnis; der Dienst ruft `CountError` nie, der Zähler je Fehlercode gehört dem Adapter, der das Fehlerobjekt schreibt (`SPEC-290` Nr. 5, Zeile Q2).
- [ ] Verdrahtung im Composition Root in den Modi `mock` und `hybrid` (`SPEC-139` Nr. 13) und Test durch Mock-Lesen und Mutation: die Spans `SPEC-029` und `SPEC-030` mit ihren Pflicht-Attributen, ohne Query-Literal und Body (Canary), mit einem Fänger je Stelle des Mock-Pfads, dass kein Request-Wert ungeprüft in einem Attribut oder in `msg` landet (Sendung L15 von `slice-telemetry-log`); die Einträge des Dienstes `workspace` laufen über denselben Logger des Roots, mit demselben Fänger je Stelle (Sendung von `slice-telemetry-handler`); der Canary „Name einer unbekannten Property von 256 und mehr Bytes“ durch `POST`, `PATCH` und `PUT` zeigt in Log und Span höchstens die gekürzte Meldung (Sendung R2 von `slice-mock-schreiben`, W66).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Macht der Slice einen Fehlercode, Schlüssel oder neue Logfelder am Binary sichtbar: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb.
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
| `src/hexagon/services/mock/` | update | `Konfig.Telemetrie` mit nil-Prüfung, Query-Span (Beginn nach Schritt (f), `Set` von Set und Coverage, `End` auf jedem Rückweg) und Mutations-Span um die Transaktion, Abbildung der Methode auf die vier Arten (`SPEC-290`) |
| `src/hexagon/services/workspace/` | keine Änderung | Entscheidung des Architect (Auftrag 1, `SPEC-290` Nr. 2): keine Operation von `WorkspaceUseCase` trägt einen Span; der Dienst schreibt seine Einträge über das `*slog.Logger`-Feld unverändert (`SPEC-129` Nr. 13) |
| `src/bootstrap/` | update | Adapter des Prozesses (Senke: `stderr` von `Start`, `Build` verwirft), `Telemetrie` des Mock in den Modi `mock` und `hybrid`, `slog.New(Handler())` als `Log` des Dienstes `workspace` in jedem Modus (`SPEC-139` Nr. 13) |
| `spec/spezifikation.md` | geliefert (Architect, in diesem Zweig) | `SPEC-290`, Nachzug `SPEC-139`, `SPEC-231`, `SPEC-195`, `SPEC-192`, §5, Historie; dieser Teil schreibt sie nicht |
| Tests | update, neu | Fakes der Ports und ein Journal-Fake der Telemetrie; der Test durch Lesen und Mutation, der Test durch den Composition Root mit dem echten Adapter |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-21 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-290`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind entschieden, der Implementer entscheidet keine im Code.

**Entscheidungen des Architect** (2026-11-21; die Aufträge 1 bis 7, je mit Ort in `spec/spezifikation.md`):

1. **Mutation und ihre Art** (`SPEC-290` Nr. 2 b, Nr. 4 d). Den Span `odata.workspace.mutation` tragen `POST`, `PATCH`, `PUT` und `DELETE` des Mock-Dienstes (`LH-FA-07.b`); `Create`, `Get`, `List`, `Delete`, `Reset`, `Restore` und die Operationen von `SnapshotUseCase` tragen keinen, die Dienste `workspace` und `snapshot` erhalten kein `Telemetry` (Alternative, Grund und Folge in `SPEC-290` Nr. 2). `recorder.mutation_type` hat vier Werte: `post`, `patch`, `put`, `delete`; die Menge wächst nur durch einen Vertrag, der eine Operation einbezieht. Die Frage des Plans, ob Anlegen, Reset und Restore dazugehören, ist damit entschieden: nein, und ihre Eintragung in `welle-admin-api` bleibt dort.
2. **Grenzen der Spans** (`SPEC-290` Nr. 3). Beide Spans beginnen nach der Rückkehr von `Workspaces.Get` (Schritt (f)) und vor `Metadaten.Get` und enden auf jedem Rückweg mit der Rückkehr von `Mock`; der Span der Query schließt Coverage, Interpreter und den Schreiber in den Speicher ein, nicht das Schreiben auf die Verbindung; der Span der Mutation liegt im Dienst um `WithWorkspaceTransaction`, nie in `f`, so dass die Rückführung „Typ des Adapters im Dienst“ (§4) entfällt. `StartSpan` erhält das `ctx` des Aufrufs, das gelieferte `ctx` wird verworfen (kein Port erhält es); Eltern ist der Gesamtspan über das `ctx` des Aufrufs, ohne Zutun des Dienstes. `$metadata` trägt keinen Span (`SPEC-290` Nr. 2).
3. **Attribute** (`SPEC-290` Nr. 4). `recorder.coverage_status`: `complete`, `partial`, `unknown`, `nicht_gelesen`; `recorder.entity_set`: der Name aus dem Modell nach der Prüfung, sonst der Platzhalter `_.` (nie ein Name eines Sets, weil ein einfacher Name kein `.` enthält); `recorder.workspace_id`: die Kennung des gelesenen Workspace, bei `ODR-E7001` und `ODR-E5004` gibt es keinen Span, so dass der Wert des Headers nie ein Attribut wird. Beide Attribute stehen schon bei `StartSpan`, damit jedes Ende vollständig ist (`pflicht` zählt nicht); `Set` ersetzt nur `entity_set` und `coverage_status`.
4. **Fehlerzähler** (`SPEC-290` Nr. 5). Der Dienst `mock` ruft `CountError` nie, für keinen der 16 Codes; der HTTP-Adapter zählt jedes Fehlerobjekt, das er schreibt (`SPEC-230` Nr. 8). Verdikt: die Entscheidung gilt, der Plan hat falsch behauptet („der Zähler steigt für die Fehler des Mock-Pfads“ im Ziel und in der DoD, berichtigt); ein Zähler im Dienst zählte jeden Fehler doppelt und im Modus `hybrid` auch die, die der Hybrid-Dienst verwirft.
5. **Aufrufstellen als Menge** (`SPEC-290` Nr. 6, Nr. 7). Sechs Stellen von `Telemetrie` (T1 bis T6), zehn Operationen mit Span (6 Lesezugriffe, 4 Mutationen), die Rückwege (Query 12 Stellen, Mutation 13) und die Mengen unten.
6. **Tabelle, Raster, Gegenproben, Größe:** unten, §4 und §6. Gegenproben: keine; der Slice ändert kein Gate und kein Werkzeug (`tools/gegenproben` bleibt unberührt).
7. **Einträge des Dienstes `workspace`** (`SPEC-231` Nr. 4 und 7, `SPEC-139` Nr. 13 b). Beide Ereignisse (`bereinigung_fehlgeschlagen`, `bereinigung_zurueckgestellt`) sind nachgetragen, die Schlüssel `ereignis`, `workspace_id`, `error_code` liegen in der Menge `Log`; der Dienst erhält `slog.New(Handler())` des Adapters des Prozesses in jedem Modus, sein Code und sein Wortlaut bleiben (`SPEC-129` Nr. 13).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-21; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). `Nr.` ist eine Kennung: `Q` gehört zum Teil 1 (Grundlage, Query, Verdrahtung, Einträge `workspace`), `M` zum Teil 2 (Mutation), siehe den Schnitt in §4; die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Q1 | (`SPEC-290` Nr. 1) `New` liefert `model.ErrUngueltig` und keinen Dienst bei nil `Telemetrie`, im Text ohne Wert, ohne Aufruf von `Telemetrie` und ohne Port | Fall in `konfig_test.go`: nil `Telemetrie` bei sonst gültiger Konfig; Orakel Literal (Fehler, kein Dienst, 0 Aufrufe am Fake); die Fälle der übrigen Felder bleiben | nil-Prüfung entfällt; Prüfung nur bei leeren `Dienste` | andere |
| Q2 | (Nr. 1, 5) Der Dienst ruft von `Telemetrie` nur `StartSpan`, `Set` und `End`, nie `Log` und nie `CountError`: für keinen der 16 Codes und keinen Kontextfehler | Schleife über M-Codes (22 Szenarien, auch die der Mutation) und M-Kontext (ein Kontextfehler je Art, 10): `FakeTelemetry.Anzahl` für `Log` und `CountError` gleich 0; dazu der Quelltext des Pakets (`go/parser`): kein Selektor `Log` oder `CountError` auf `support.Telemetry`; Zählprüfung: die Literalliste des Tests führt 16 Codes (von Hand gegen `SPEC-195` Nr. 10, Negativ 4) | `CountError(code)` in `abbilden` oder `mitCode`; `Log` im Zweig `ODR-E9001`; `CountError` beim Kontextfehler | andere |
| Q3 | (Nr. 2 a, 6 T1 und T4) Jede der sechs Operationen (`GET`, `HEAD` je an Collection, Entity, Anzahl) öffnet `odata.mock.query` genau einmal und beendet ihn genau einmal; der Span trägt allein die zwei Pflicht-Schlüssel | Schleife 2 Methoden × 3 Arten gleich 6: `FakeTelemetry` zählt `StartSpan` 1 und `End` 1, Name `SpanMockQuery`, Schlüssel genau die aus `support.Spans()` (nicht aus dem Code); Orakel Literal | `HEAD` ohne Span; `/$count` ohne Span; Name `SpanWorkspaceMutation`; ein drittes Attribut (`http.response.status_code`); zweites `End` | andere |
| Q4 | (Nr. 2 c, 7 a) Keinen Span tragen `$metadata` und jede Anfrage, die an den Schritten (a) bis (f) endet: kein `StartSpan`, kein Aufruf von `Telemetrie` | Menge M-OhneSpan (21 Fälle, Orakel Literal): `$metadata` mit `GET` und `HEAD`; je Schritt: `ctx` nil, leere Methode, Pfad ohne Dienst, Header fehlt, leer, zwei Werte, URL `ODR-E6005`, `ODR-E6001`, `ODR-E6006`, Art außerhalb, `OPTIONS`, `NurLesend` mit `POST`, `POST` an Entity, `$filter` an Mutation, Dienst ohne Leser, Workspace unbekannt, Workspace eines anderen Service, `Workspaces.Get` mit Fehler, mit Kontextfehler; `FakeTelemetry.Aufrufe()` leer | `StartSpan` am Anfang von `Mock`; `StartSpan` vor `holeWorkspace`; `StartSpan` auch für `$metadata` | andere |
| Q5 | (Nr. 3 a und b, 6) Beginn nach `Workspaces.Get` und vor `Metadaten.Get`; Folge `StartSpan`, `Set(entity_set)`, `Set(coverage_status)`, `End`; `End` nach dem letzten Aufruf des Schreibers | Journal-Fake (`telemetrieFake` schreibt in dasselbe Protokoll `prot` wie die Port-Fakes): Folge je Fall als Literal (Collection, Entity mit Treffer, Entity ohne Treffer, Anzahl, je `GET`): `Workspaces.Get`, `StartSpan`, `Metadaten.Get`, `Blobs.Open`, …, `Schreibe…`, `End` | `StartSpan` nach `holeService`; `StartSpan` vor `holeWorkspace`; `End` vor dem Schreiber; `Set(coverage_status)` vor `Coverage.Get` | andere |
| Q6 | (Nr. 3 b) Genau ein `End` auf jedem Rückweg der Query, mit Fehler mit Code und mit Kontextfehler | M-Rückwege-Query: 12 Stellen (S3 bis S11, S14 bis S16) mit Fehler mit Code und 9 mit Kontextfehler (S3 bis S11) gleich 21 Fälle; je Fall `StartSpan` 1, `End` 1, Ergebnis unverändert (Code bzw. Kontextfehler ohne Code); Zählprüfung 12 und 9 als Literal des Tests (von Hand gegen `SPEC-195` Nr. 11, Negativ 4) | `End` nur im Erfolgsweg; `End` erst nach dem Zweig von `fuehreAus`; `defer` entfällt im Zweig von `Coverage.Get` | andere |
| Q7 | (Nr. 3 d) `StartSpan` und jeder Port der Query erhalten das `ctx` des Aufrufs unverändert, nie das des Spans | Schleife über die 11 Stellen S2 bis S12 und `StartSpan`: der Fake hält jedes `ctx` fest, `FakeTelemetry.StartSpan` liefert ein abgeleitetes (eigener Wert); Identität; Zählprüfung 11 als Literal des Tests (Negativ 4) | `ctx, span := Telemetrie.StartSpan(ctx, …)` | andere |
| Q8 | (Nr. 3 e) Der Span der Query ist Kind des Gesamtspans, wenn das `ctx` ihn trägt, sonst ohne Eltern | Test durch den Composition Root mit dem echten Adapter (`NeueTelemetrie`, feste Uhr): Elternspan `SpanGatewayRequest` im `ctx`, dann `Mock` mit `GET`: `ParentID` im `Snapshot` gleich dessen `ID`, `parent_id` der Zeile gleich dessen `span_id`; ohne Elternspan leer | `StartSpan(context.Background(), …)` | andere |
| Q9 | (Nr. 4 a) `recorder.entity_set` ist der Name aus dem Modell nach der Prüfung, sonst `_.`; nie der Text des Requests | Schleife 3 Arten × {bekanntes Set, unbekanntes Set (Name mit `'`, Leerzeichen, `$`, `/`, einem Byte ab 0x80, 300 Bytes; aus Bytewerten), Modell nicht lesbar (S4)}: Attribut des Fakes gleich dem Orakel, der Canary in keinem Attribut; mit dem echten Adapter ein Set mit Namen außerhalb der Form (Byte ab 0x80 im Modell): `Dropped[wert]` 1, `Dropped[pflicht]` 0, Wert `_.` | Wert aus `u.Ressource.EntitySet` bei `StartSpan`; Platzhalter leer; Platzhalter `unbekannt`; kein `Set` nach der Prüfung | andere |
| Q10 | (Nr. 4 b) `recorder.coverage_status` bei Collection und Anzahl ist der gelesene Status, auch bei `ODR-E6003` des strikten Modus | Schleife 2 Arten × 3 Status (`model.AlleCoverageStatus()`, Zählprüfung 3) × 2 Werte von `STRICT_COVERAGE` gleich 12; Orakel Literal (Status, Ausgang 200 oder `ODR-E6003`); dazu strikt, `partial`, `$filter` mit unbekannter Property: `ODR-E6003` und `partial` | `Set` nach der Entscheidung des strikten Modus; Wert immer `unknown`; `partial` und `unknown` vertauscht | andere |
| Q11 | (Nr. 4 b) Beim Lookup ohne Treffer ist der Wert der gelesene Status, wo die Coverage nicht gelesen wird `nicht_gelesen` | M-Coverage-Lookup: ohne Treffer × 3 Status × 2 Modi gleich 6 (`ODR-E6008` bei `complete`, sonst `ODR-E6003`); nicht gelesen: Treffer, Tombstone, `ErrNichtGefunden` der Coverage (`unknown`, gelesen), Lesefehler, Status `bogus` (`ODR-E9001`), Kontextfehler, unbekanntes Set gleich 7 (davon einer `unknown`); `Select` mit unbekannter Property bei Treffer (`ODR-E6003`, `nicht_gelesen`); Orakel Literal | `Set` auch beim Treffer; `unknown` bei Lesefehler; `Set` mit dem Rohwert `bogus` | andere |
| Q12 | (Nr. 6) Zahl der Aufrufe je Anfrage: `StartSpan` 1, `End` 1, `Set` je T2 und T3 mit einem Attribut; der Quelltext trägt sechs Stellen | in jedem Fall von Q3, Q9, Q10 und Q11 zählt der Fake jede Operation (Orakel: die Tabelle von `SPEC-290` Nr. 6 als Literal); `go/parser`: `StartSpan`, `Set` und `End` je 2 Aufrufe im Paket ohne `_test.go`, keine andere Operation von `Telemetry` oder `Span` | `Set` mit beiden Schlüsseln in einem Aufruf; `Set(entity_set)` auch bei einem Weg ohne Set; `End` doppelt; dritter `StartSpan` | andere |
| Q13 | (Nr. 9 a) *n* gleichzeitige Anfragen ergeben je Span-Name *n* `StartSpan` und *n* `End`; jeder Span endet genau einmal und trägt die Attribute seiner Anfrage | 16 Goroutinen × 6 Operationen in 20 Läufen unter dem Detektor (`make test-race`), je Goroutine eigenes Set und eigener Workspace; Wirkungs-Fänger: der Fake führt Buch je Span (`End` je Span 1, Attribute der Anfrage); Rate je Prozess im Plan (`-test.count=1`) | Span in einem Feld des `Mocker` gehalten; `End` nur je zweiten Aufruf; ein geteilter Span je Set | nebenläufig |
| Q14 | (Nr. 8; Sendung L15) Kein Wert aus dem Request erreicht ein Attribut oder die Senke | Menge M-Stellen-Query, 8 Stellen (Set, Schlüssel, Literal in `$filter`, Name in `$filter`, Name in `$orderby`, Name in `$select`, Header `X-Recorder-Workspace`, `Accept`), je in jeder Art, die sie liest (Orakel Literal der Paare); Canary aus Bytewerten (`AGENTS.md` §3.18); geprüft gegen jedes Attribut des Fakes und die Bytes der Senke des echten Adapters; Zählprüfung 8 | `Str("recorder.entity_set", <Wert der Stelle>)` je Stelle (8 Mutationen); `err.Error()` als Attribut | andere |
| Q15 | (`SPEC-139` Nr. 13 a, d) In den Modi `mock` und `hybrid` erhält der Mock den Adapter des Prozesses; eine Anfrage schreibt die Zeile mit `ereignis` `span` und beiden Attributen in die Senke | Test durch den Composition Root (`Build` mit `NeueTelemetrie`, Senke ein Puffer, feste Uhr), `GET` auf eine Collection in beiden Modi; die Zeile als Literal aus `SPEC-231` Nr. 6 (Felder und Reihenfolge, Orakel aus der Spezifikation), `Dropped` alle 0; `Build` mit nil-Adapter scheitert nicht still | `Telemetrie` ein zweiter, eigener Adapter; Mock ohne Telemetrie (Build scheitert); Modus `hybrid` ohne den Adapter | andere |
| Q16 | (`SPEC-139` Nr. 13) `Start` übergibt seinen `stderr` als Senke, `Build` ohne `Start` verwirft | `start` mit einem Puffer als `stderr` und den Bauteilen der Tests: eine Anfrage im Modus `mock`, danach steht die Zeile des Spans im Puffer und `stdout` bleibt leer; `Build` allein: kein Byte irgendwo | Senke `io.Discard` in `start`; Senke `os.Stderr` statt des Arguments | andere |
| Q17 | (`SPEC-231` Nr. 4 und 7, `SPEC-139` Nr. 13 b) Die Einträge `bereinigung_fehlgeschlagen` und `bereinigung_zurueckgestellt` des Dienstes `workspace` laufen in jedem Modus über `slog.New(Handler())` des Adapters: Zeile mit Text, Stufe und Feldern der Stelle, ohne Verwurf | Test durch den Composition Root, 4 Modi × 3 Fälle (`CleanupBlobs` liefert `ODR-E3001`, `ODR-E9001`, `Zurueckgestellt`; Fake über `NeueWartung`) gleich 12; Zeile als Literal aus `SPEC-129` Nr. 13 (Text, Stufe `ERROR` bzw. `INFO`, Felder und Reihenfolge nach `SPEC-231` Nr. 1), `Dropped` alle 0; ein Kontextfehler der Bereinigung schreibt nichts (1 Fall je Modus) | `Log: h.log` im Dienst (Textzeile); anderer Text; `error_code` fehlt; Stufe `WARN` | andere |
| Q18 | (`SPEC-231` Nr. 4; Sendung von `slice-telemetry-handler`) Der Eintrag des Dienstes `workspace` trägt nie einen Wert eines abgewiesenen Aufrufs | Fall: `Delete` mit einer Kennung außerhalb der Form (70 Bytes, `/`, Leerzeichen, Byte ab 0x80; aus Bytewerten) gibt `ODR-E7001`, die Senke bleibt leer, `CleanupBlobs` wird nicht gerufen; Fall: gelungenes `Delete` einer Kennung der Form: `workspace_id` ist sie unverändert, kein anderes Feld trägt sie | Eintrag auch nach `ErrNichtGefunden`; Kennung in `msg` | andere |
| Q19 | (Nr. 2 Entscheidung) Die Dienste `workspace` und `snapshot` rufen `StartSpan` nie; ihre `Konfig` trägt kein `Telemetrie` | `go/parser` über beide Pakete ohne `_test.go`: kein Selektor `StartSpan` oder `CountError`, kein Feld vom Typ `support.Telemetry` (Zählprüfung 2 Pakete); durch den Root: die 8 Operationen (`Create`, `Get`, `List`, `Delete`, `Reset`, `Restore`, Snapshot `Create`, `List`) lassen `Durations` der beiden Span-Namen bei Count 0 | `StartSpan` in `Reset`; Feld `Telemetrie` in `workspace.Konfig` | andere |
| Q20 | (Nr. 5, Folge) Der Zähler je Fehlercode des Mock-Pfads steigt durch den Adapter, nicht durch den Dienst | **Akzeptiertes Negativ.** Grund: der Fänger braucht den HTTP-Adapter mit `Telemetrie`, den dieser Slice nicht baut (`SPEC-136` Nr. 16, `slice-telemetry-einbau-gateway`); der Dienst zählt nie (Q2). Folge: bis der Adapter zählt, steigt `odr_errors_total` für keinen Fehler des Mock; Ziel und DoD des Nehmers tragen „jeden Fehler, den der Adapter schreibt“ | – | andere |
| Q21 | (`SPEC-232` Nr. 3, Nr. 9 *Ungeprüft*) Ein Ausfall der Senke ändert Antwort und Fehler des Dienstes nicht | **Akzeptiertes Negativ.** Grund: keine Operation von `Telemetry` liefert einen Wert oder Fehler (`SPEC-230` Nr. 1), der Dienst kann sein Ergebnis nicht aus ihr bilden; der Fänger gegen den Ausfall liegt am Adapter (`SPEC-232` Nr. 3). Folge: ein Dienst, der eine Panik oder einen Wert der Telemetrie auswertete, bräche die Zusage; der Typ lässt es nicht zu (0 von 3 Operationen liefern einen Wert) | – | andere |
| Q22 | (Nr. 3 b und c; *Ungeprüft* Dauer) Die Grenzen sind durch die Folge der Aufrufe belegt, nicht durch gemessene Zeit | **Akzeptiertes Negativ.** Grund: eine Zeitmessung im Test schwankt mit der Last (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`); die Dauer misst `slice-lasttest`. Folge: Arbeit, die ein Eingriff ohne Änderung der Folge der Ports vor den Beginn oder hinter das Ende legt, bleibt ungefangen (Q5 und M2 binden 8 Folgen, nicht die Zeit) | – | andere |
| M1 | (Nr. 2 b, 6 T5 und T6) Jede der vier Mutationen öffnet `odata.workspace.mutation` genau einmal und beendet ihn genau einmal; kein `Set`, genau die zwei Pflicht-Schlüssel | Schleife 4 Methoden: `StartSpan` 1, `End` 1, `Set` 0, Name `SpanWorkspaceMutation`, Schlüssel genau die aus `support.Spans()`; Orakel Literal | `PATCH` und `PUT` ohne Span; Name `SpanMockQuery`; `Set(mutation_type)` nach der Transaktion | andere |
| M2 | (Nr. 3 a bis c) Der Span umschließt die Transaktion: Beginn nach `Workspaces.Get` und vor `Metadaten.Get`, Ende nach der Rückkehr von `WithWorkspaceTransaction`; in `f` kein Aufruf von `Telemetrie` | Journal-Fake: Folge je Methode als Literal (`Workspaces.Get`, `StartSpan`, `Metadaten.Get`, …, Beginn der Transaktion, …, Ende der Transaktion, `End`); zwischen Beginn und Ende der Transaktion steht keine Operation des Fakes; 4 Folgen | `StartSpan` in `f`; `End` vor der Transaktion; `End` in `f` | andere |
| M3 | (Nr. 3 b) Genau ein `End` auf jedem Rückweg der Mutation, mit Fehler mit Code und mit Kontextfehler | M-Rückwege-Mutation: 13 Stellen (S3 bis S5, S17 bis S26) mit Fehler mit Code und 12 mit Kontextfehler (alle außer S25) gleich 25 Fälle; je Fall `End` 1, Rollback (der Fake der Transaktion), Ergebnis unverändert; Zählprüfung 13 und 12 als Literal des Tests (Negativ 4) | `End` nur im Erfolgsweg; `End` vor dem Fehler-Zweig von `f` | andere |
| M4 | (Nr. 4 d) `recorder.mutation_type` ist `post`, `patch`, `put` oder `delete`, nie der Text des Requests | Schleife 4 Methoden, Orakel Literal je Methode, Zählprüfung 4; dazu `post` in Kleinschreibung als Methode: kein Span (`ODR-E6002`, `SPEC-192` Nr. 3) | rohe Methode (Form `Token` verletzt, `wert`); `patch` und `put` vertauscht; Wert `update` | andere |
| M5 | (Nr. 4 c) `recorder.workspace_id` ist die Kennung des gelesenen Workspace, nicht der Header; bei unbekanntem oder fremdem Workspace kein Span | Fall: der Fake liefert zu Header `ws-kopf` einen Workspace `ws-gelesen` (Fixture ungleich dem Wert der Mutation): Attribut `ws-gelesen`; unbekannt (`ODR-E7001`) und fremd (`ODR-E5004`) mit Canary im Header: kein Span, Canary in keinem Attribut; Kennung mit Leerzeichen mit dem echten Adapter: `Dropped[wert]` 1, `Dropped[pflicht]` 1, Antwort unverändert; Workspace nach Schritt (f) gelöscht: Span trägt die Kennung, `ODR-E7001` | Header-Wert als Attribut; `StartSpan` vor `holeWorkspace` mit dem Header | andere |
| M6 | (Nr. 3 f, 9 b) Ein Abbruch mitten in der Transaktion ändert das Ergebnis nicht (Kontextfehler ohne Code, Rollback), und der Span endet genau einmal | deterministisch: der `Put`-Fake beendet den `ctx`, 4 Methoden; Wirkungs-Fänger: 16 Goroutinen × 4 Methoden, die Hälfte mit Abbruch in S22, in 20 Läufen unter dem Detektor: `End` je Aufruf 1, Revision und Entity-State der abgebrochenen unverändert, der übrigen plus 1 (Rate je Mutation im Plan) | `End` nur im Erfolgsweg; `End` unter `if ctx.Err() == nil`; Ergebnis des Aufrufs aus dem Span | nebenläufig |
| M7 | (Nr. 3 d und e) `StartSpan` und jeder Port der Mutation erhalten das `ctx` des Aufrufs unverändert; der Span ist Kind des Gesamtspans | Schleife über die 9 Stellen S17 bis S24 und S26 und `StartSpan` (Zählprüfung: 11 + 9 gleich 20 mit Q7); durch den Root mit dem echten Adapter: `ParentID` gleich der `ID` des Elternspans | `ctx, span := Telemetrie.StartSpan(ctx, …)`; `StartSpan(context.Background(), …)` | andere |
| M8 | (Nr. 8; Sendung L15) Kein Wert aus dem Request erreicht ein Attribut oder die Senke, auch nicht über die Mutation | Menge M-Stellen-Mutation, 6 Stellen (Name und Wert einer Property im Body, `Content-Type`, `Prefer`, `If-Match`, `If-None-Match`) und die vier aus Q14, die eine Mutation liest (Set, Schlüssel, Header `X-Recorder-Workspace`, `Accept`), je durch jede Methode, die sie liest; Canary und Prüfung wie Q14; Zählprüfung 6 + 4 | `Str("recorder.workspace_id", <Wert der Stelle>)` je Stelle (10 Mutationen) | andere |
| M9 | (Nr. 8; Sendung R2 von `slice-mock-schreiben`) Der Name einer unbekannten Property von `grenze` Bytes durch `POST`, `PATCH` und `PUT` erreicht weder eine Zeile noch einen Span | Test durch den Root mit dem echten Adapter: 3 Methoden × 4 Längen (255, 256, 257 und 4096 Bytes aus Bytewerten) gleich 12; die Senke (alle Zeilen), die Attribute des Fakes und `msg` tragen nicht die ersten 8 Bytes des Namens; Antwort `ODR-E6007` | Attribut aus `err.Error()` der Abbildung; `Log` im Zweig `liesBody` mit der Ursache | andere |
| M10 | (Nr. 9 a) *n* gleichzeitige Mutationen ergeben *n* Spans mit Art und Kennung ihrer Anfrage | 16 Goroutinen × 4 Methoden in verschiedenen Workspaces, 20 Läufe unter dem Detektor: der Fake führt Buch je Span (`mutation_type` und `workspace_id` gleich der Anfrage, `End` je Span 1) | Span im `Mocker` gehalten; Attribute aus der letzten Anfrage | nebenläufig |

Zeilen gesamt 32: *a* = 3 (Q13, M6, M10), *b* = 29 (die akzeptierten Negative Q20 bis Q22 zählen als `andere`, `SPEC-128` Nr. 3 e).

**Mengen mit Zählprüfung** (Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15):

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| M-Op (Q3, Q4, M1) | `GET` und `HEAD` × Collection, Entity, Anzahl gleich 6; `$metadata` × 2 gleich 2; 4 Mutationen; gleich 12 | Span ja oder nein, Name | 12; die Zahl der Methoden aus `SPEC-192` Nr. 3 |
| M-OhneSpan (Q4) | Schritt (a) 2, (b) 1, (c) 3, (d) 4, (e) 5, (f) 4, `$metadata` 2 gleich 21 | kein Aufruf von `Telemetrie` | 21 |
| M-Codes (Q2) | die 16 Codes von `SPEC-195` Nr. 10, je ein Szenario, dazu ein zweites für die sechs Codes, die vor und nach dem Beginn entstehen (`ODR-E5005`, `ODR-E6001`, `ODR-E6002`, `ODR-E6005`, `ODR-E7001`, `ODR-E9001`) gleich 22 | Code, `StartSpan` 0 oder 1, `CountError` 0 | 16 und 22 als Literal des Tests (Negativ 4) |
| M-Kontext (Q2) | ein Kontextfehler je Art: 6 Lesezugriffe + 4 Mutationen gleich 10 | Kontextfehler ohne Code, `CountError` 0 | 10 |
| M-Rückwege-Query (Q6) | 12 Stellen mit Code, 9 mit Kontextfehler gleich 21 | `End` 1 | 12 und 9 als Literal (Negativ 4) |
| M-Rückwege-Mutation (M3) | 13 Stellen mit Code, 12 mit Kontextfehler gleich 25 | `End` 1, Rollback | 13 und 12 |
| M-Coverage (Q10, Q11) | Collection und Anzahl: 2 × 3 Status × 2 Modi gleich 12; Lookup ohne Treffer: 3 × 2 gleich 6; nicht gelesen 7; gleich 25 | Wert, Code | 3 Status gleich `len(model.AlleCoverageStatus())`; 25 |
| M-ctx (Q7, M7) | S2 bis S12 gleich 11, S17 bis S24 und S26 gleich 9 gleich 20, dazu `StartSpan` je Art | Identität des `ctx` | 20 als Literal (`SPEC-195` Nr. 11: „20 ein `ctx`“, Negativ 4) |
| M-Art (M4) | 4 Methoden | `post`, `patch`, `put`, `delete` | 4 |
| M-Stellen-Query (Q14) | 8 Stellen, je Art, die sie liest | kein Canary in Attribut und Senke | 8 |
| M-Stellen-Mutation (M8) | 6 Stellen + 4 Stellen der Query, die eine Mutation liest gleich 10 | wie oben | 14 Stellen insgesamt gleich `SPEC-290` Nr. 8 (Q14 acht, M8 sechs) |
| M-Canary (M9) | 3 Methoden × 4 Längen gleich 12 | kein Fund des Namens | 12 |
| M-Einträge-`workspace` (Q17) | 4 Modi × 3 Fälle gleich 12, dazu 4 Fälle ohne Eintrag | Zeile als Literal | 12 |
| M-Gleichzeitig (Q13, M10) | 16 Goroutinen × (6 Operationen bzw. 4 Methoden), 20 Läufe | Buch je Span | 16 × 6, 16 × 4 |

**Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** der Name eines Sets im Request (Prüfgrenze 300 Bytes, die Länge, die Q9 führt; Grund: der Dienst setzt nie den Namen des Requests, die Länge ändert den Weg nicht; Folge: ein Name von 1 MiB wird nicht eigens geprüft, `SPEC-160` begrenzt die URL) und der Body (Prüfgrenze 4096 Bytes des Namens in M9; die Grenze des Bodys ist `MaxBytes`).

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; die Rangfolge steht in `SPEC-195` Nr. 3, `SPEC-192` Nr. 2, `SPEC-196` Nr. 2 und `SPEC-197` Nr. 2):

| Paar oder Modus | Erwartung | Fänger |
|---|---|---|
| unbekanntes Set und Coverage `partial` im strikten Modus | `ODR-E6008` (das Set geht der Coverage voraus): `_.` und `nicht_gelesen` | Q9, Q11 |
| bekanntes Set, strikt, `partial`, `$filter` mit unbekannter Property | `ODR-E6003` vor `ODR-E6005`: Name des Sets und `partial` | Q10 |
| Lookup auf eine `partial`-Entity mit `$select` einer unbekannten Property | `ODR-E6003`, Coverage nicht gelesen: Name und `nicht_gelesen` | Q11 |
| Lookup auf einen Tombstone | `ODR-E6008`, `nicht_gelesen` | Q11 |
| `PATCH` auf fehlenden Eintrag, Coverage `unknown` | `ODR-E6003`, der Span trägt kein Coverage-Attribut | M3 |
| Fehler mit Code und Kontextfehler zugleich (`SPEC-109` Nr. 6) | Kontextfehler ohne Code, `End` 1 | Q6, M3 |
| Modus `mock` und `hybrid` | beide tragen die Zeile des Spans | Q15 |
| Modus `record`, `replay`, `mock`, `hybrid` beim Dienst `workspace` | die Zeile des Eintrags | Q17 |
| `STRICT_COVERAGE` an und aus × drei Status | Wert gleich, Ausgang verschieden | Q10 |

**Akzeptierte Negative dieses Vertrags** (Grund und Folge; die Zeilen Q20 bis Q22 oben tragen ihre eigenen): (1) der Platzhalter `_.` für den Namen eines Sets außerhalb der Form (`SPEC-290` Nr. 4 e): Q9 führt den Fall mit dem echten Adapter, kein weiterer Name wird geführt; Folge: ein Set mit Nicht-ASCII im Namen erscheint als `_.`, 0 von 4 Werten von `coverage_status` betroffen. (2) Ein Typed-nil in `Telemetrie` (`SPEC-290` *Ungeprüft*): `New` prüft nil; Folge: eine Panik beim ersten Span, im Betrieb unerreichbar (der Root übergibt den Adapter). (3) Der Span des Hybrid-Fallbacks (`SPEC-290` Nr. 7 g, *Ungeprüft*): kein Nehmer, siehe §6. (4) Die Zahlen der Mengen (16 Codes, 12 und 9 und 13 und 12 Rückwege, 20 Stellen mit `ctx`, 14 Stellen des Requests) gegen die Tabellen der Spezifikation prüft kein Test, weil die Stufe `source` des Images `spec/` nicht trägt; Folge: ändert jemand die Tabelle von `SPEC-195` Nr. 11 oder die Liste von `SPEC-195` Nr. 10 ohne den Test, sieht es erst ein Review (einmal von Hand gezählt am 2026-11-21: 26 Stellen, davon 12 und 13 hinter dem Beginn, 20 mit `ctx`, 16 Codes).

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Mutationsart, jeder Coverage-Status, jeder Fehlercode des Mock-Pfads), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans`, `slice-mock-lesen`, `slice-mock-schreiben`, `slice-snapshot-use-case` und `slice-bootstrap-m3` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-port`: der Port und sein Fake, die dieser Slice ruft.
- `slice-telemetry-log`, `slice-telemetry-handler` und `slice-telemetry-spans` (Schnitt des Planners vom 2026-11-07): *Code-Abhängigkeit*. Ohne Zeile, Handler und Spans im Adapter schreibt das Binary keine Span-Zeile; der Einbau prüft Wirkung und Span-Zeile.
- `slice-mock-lesen` und `slice-mock-schreiben` (`welle-mock`): die Orte der beiden Spans (`src/hexagon/services/mock`); ohne Lesen und Mutation gibt es weder Query-Ausführung noch Zustandsänderung zu umschließen.
- `slice-snapshot-use-case` (`welle-mock`): Auftrag 1 hat entschieden, dass `Reset`, `Restore` und Snapshot keinen Span tragen (`SPEC-290` Nr. 2); die Kante bleibt, weil Q19 die acht Operationen von `WorkspaceUseCase` und `SnapshotUseCase` durch den Composition Root ruft und der Dienst dafür bestehen muss.
- `slice-bootstrap-m3` (`welle-mock`): der Modus `mock` im Composition Root, über den der Test die Spans erreicht (kein HTTP-Adapter für Admin, kein CLI).
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` aus `welle-benutzerhandbuch` liegen dann in `done/`; beide sind Bedingung von `welle-mock` und damit hier erfüllt.

**Abhängig von `welle-mock` / M3, weil** die Orte beider Spans in `welle-mock` entstehen; das ist der Grund des Start-Triggers der Welle ([welle-betrieb](../welle-betrieb.md) §2). **Nicht abhängig** von `slice-telemetry-einbau-gateway`, `slice-container-image` und `slice-lasttest`. Reihenfolge-Vorsicht, kein Blocker: beide Einbau-Slices ändern die Verdrahtung des Adapters in `src/bootstrap/`; läuft dieser Slice nach `slice-telemetry-einbau-gateway`, übernimmt er dessen Verdrahtung, läuft er davor, legt er sie an und der Gateway-Slice zieht sie nach (der Planner nennt die Reihenfolge beim Start des zweiten). Der Vertrag sagt, was am Ende gilt: ein Adapter des Prozesses (`SPEC-139` Nr. 13), den der zweite Slice verwendet statt einen weiteren zu erzeugen; der Adapter des Replay-Dienstes (`SPEC-139` Nr. 3) und das `Log` von `record`, `session` und `maintenance` gehören dem Gateway-Slice, und `SPEC-139` Nr. 3 und 13 c sind dort nachzuziehen, sobald er den Replay-Dienst und die drei Logger umstellt (`SPEC-231` Nr. 7 nennt sie schon).

**Größenmarke:** noch nicht gesetzt, **Schnittfrage ja** (Architect, 2026-11-21, `SPEC-128` Nr. 1). Die Tabelle hat 32 Zeilen, *a* = 3 (Q13, M6, M10), *b* = 29, *f* = 1 (ein Journal-Fake der Telemetrie, der in das Protokoll `prot` der Port-Fakes schreibt; `FakeTelemetry` besteht), *p* = 95 (Produktcode: `Konfig` und nil-Prüfung rund 8, Query-Span mit zwei `Set` rund 30, Mutations-Span mit Abbildung der Methode rund 25, Composition Root rund 32); *S* = 85·3 + 80·29 + 200·1 + 95 = 2870, *M* = 3900, über der Schnittgrenze (*S* ≥ 2223). Der Schnitt steht **vor dem Code** (`AGENTS.md` §3.14), die Naht ist die des Plans, Query-Span gegen Mutations-Span, an den Präfixen der Zeilen:

| Teil | Inhalt | Zeilen der Tabelle | *a* | *b* | *f* | *p* | *S* | *M* |
|---|---|---|---|---|---|---|---|---|
| 1 | Konfiguration, Query-Span, Attribute, Verdrahtung im Composition Root, Einträge des Dienstes `workspace` | Q1 bis Q22 (22) | 1 | 21 | 1 | 70 | 2035 | 2800 |
| 2 | Mutations-Span, Canary der Sendung R2 | M1 bis M10 (10) | 2 | 8 | 0 | 25 | 835 | 1200 |

Der Teil 2 setzt den Teil 1 voraus (das Feld `Telemetrie`, die Verdrahtung); die Teile folgen nacheinander (Reihenfolge, die Naht teilt keinen Mechanismus außer dem Feld, das Teil 1 liefert). Der Journal-Fake gehört zu Teil 1 und steht Teil 2 zur Verfügung. Die Marke setzt der Planner nach dem Schnitt für jeden Teil mit `make groessenmarke` (`SPEC-128`); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/hexagon/services/mock/` und `src/bootstrap/`, Tests eingeschlossen (der Block mit `pfad`, `produktcode` und `fakes` des Architect entsteht mit dem Schnitt, weil erst dann feststeht, welche Zeilen ein Plan trägt). Der Schnitt ist der Vorschlag des Architect, der Planner prüft ihn und legt den zweiten Plan an; dass die Formel bisher überschätzte (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`), ändert die Schnittfrage nicht, sie folgt der Formel (`AGENTS.md` §3.14).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach dem Schnitt setzt (`AGENTS.md` §3.14); ein weiterer Schnitt ist nicht vorgesehen, weil beide Teile nach dem Schnitt unter der Schnittgrenze liegen.
- `in-progress` → `open` (blockiert): der Adapter des Prozesses lässt sich nicht für den Mock und den Dienst `workspace` zugleich verdrahten, ohne die Signatur von `Build` zu ändern (`SPEC-139` Nr. 13: sie bleibt); die frühere Bedingung, der Mutations-Span brauche einen Typ des Adapters im Dienst, entfällt, weil der Span um die Transaktion liegt und nicht in ihr (`SPEC-290` Nr. 3 c).

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

- Ein Span, der den Zustandswechsel des Workspace umschließt, kann den Abbruch (`ctx`) mitten in der Transaktion sehen; der Span muss trotzdem enden und darf das Ergebnis der Mutation nicht ändern — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-290` Nr. 3 b, c, f): der Span liegt um die Transaktion, `End` wirkt bei beendetem `ctx` und endet genau einmal, `f` ruft `Telemetrie` nie; Fänger M2, M3 und der Wirkungs-Fänger M6 (Klasse `nebenläufig`, Rate je Mutation im Plan des Implementers) mit `make test-race`.
- `recorder.entity_set` und `recorder.workspace_id` stammen aus dem Request; ein Wert aus dem Request darf nur als Attribut stehen, wenn er geprüft ist — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-290` Nr. 4 a, c): der Name kommt aus dem Modell nach der Prüfung, sonst steht der Platzhalter `_.`; die Kennung kommt aus dem gelesenen Workspace, ohne gefundenen Workspace gibt es keinen Span; Fänger Q9, Q14, M5, M8 mit Canary aus Bytewerten.
- Der Span `odata.mock.query` zerlegt den Gesamtspan nur, wenn beide verbunden sind (`LH-QA-01.a`); wird die Verbindung über `ctx` nicht gehalten, weist der Lasttest den Anteil nicht aus — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-290` Nr. 3 d, e): `StartSpan` erhält das `ctx` des Aufrufs, so dass der Gesamtspan Eltern ist; kein Port erhält das gelieferte `ctx`; Fänger Q7, Q8, M7 (der Test durch den Composition Root mit dem echten Adapter).
- Pfad ohne Lauf unter dem Detektor (Abbruch, Ausfall der Senke) — `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) — **Ausgang:** entschieden (Architect, 2026-11-21): der Abbruch hat einen Wirkungs-Fänger (M6, die Hälfte der 16 Goroutinen bricht in S22 ab, `End` je Aufruf 1, Zustand unverändert); der Ausfall der Senke ist ein akzeptiertes Negativ mit Grund und Folge (Zeile Q21: der Dienst kann sein Ergebnis nicht aus `Telemetry` bilden, der Fänger liegt am Adapter, `SPEC-232` Nr. 3).
- Einträge des Dienstes `workspace` (`SPEC-129` Nr. 13) fehlen in `SPEC-231` Nr. 4 und 7 (Sendung von `slice-telemetry-handler`, §1) — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-231` Nr. 4 und 7, `SPEC-139` Nr. 13 b); Fänger Q17 und Q18.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-21): das Raster unten trägt in jeder Zelle einen Ort oder `gilt nicht` mit Grund.
- Der Zähler je Fehlercode des Mock-Pfads liegt beim Adapter, nicht beim Dienst; die DoD und das Ziel dieses Plans sagten das Gegenteil (Auftrag 4, Verdikt: die Entscheidung `SPEC-230` Nr. 8 gilt, der Plan hat falsch behauptet) — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-290` Nr. 5); Ziel und DoD sind in diesem Commit berichtigt; Nehmer `slice-telemetry-einbau-gateway` trägt den Zähler in Ziel und DoD (gelesen am 2026-11-21, §1 *Ausdrücklich NICHT* trifft die Sendung nicht, `AGENTS.md` §3.13), Fänger dort; hier das akzeptierte Negativ Q20.
- Schnittfrage: *S* = 2870 für die ungeschnittene Tabelle, über der Schnittgrenze (§4) — **Ausgang:** weiter offen; Adresse: Planner (legt vor dem Start den zweiten Plan für die Zeilen M1 bis M10 an, setzt beide Marken und zieht die Sendung des Canary R2 von `slice-mock-schreiben` (§1) in den Plan des Teils 2, `AGENTS.md` §3.13).
- Handbuch: Mit der Verdrahtung (`SPEC-139` Nr. 13 d) schreibt das Binary in den Modi `mock` und `hybrid` je Anfrage mit Span eine JSON-Zeile nach `stderr`, und die Einträge des Dienstes `workspace` erscheinen als JSON-Zeile statt als Textzeile; der Handbuch-Punkt der DoD greift (Logfelder am Binary sichtbar) — **Ausgang:** weiter offen; Adresse: Planner (ob Teil 1 das Kapitel schreibt oder `slice-handbuch-mock`; der Planner liest dessen §1 *Ausdrücklich NICHT* und DoD vor der Zuweisung, `AGENTS.md` §3.13).
- Span des Hybrid-Fallbacks (nicht der Versuch im Mock): kein Slice von `welle-betrieb` nimmt ihn an (§1, Sendung ohne Nehmer, geführt in `slice-telemetry-einbau-gateway` §6) — **Ausgang:** weiter offen; Adresse: Planner (Frage an den User, dort gestellt).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-21; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund; „Nr.“ ohne Nennung heißt `SPEC-290`):

| Operation | fehlender Wert (Workspace, Entity-Set) | leerer Wert | unbekannter oder fremder Wert | Abbruch (`ctx`) | Wiederholung | Coverage `partial`/`unknown` |
|---|---|---|---|---|---|---|
| Query-Span: Collection | Nr. 7 a: Header fehlt `ODR-E5003`, Pfad ohne Set `ODR-E6001`, beide ohne Span | Nr. 7 a: leerer Header `ODR-E5003`, ein leerer Name im Pfad ist ein Fehler des Lesens der URL (Schritt (d), `SPEC-160`), beide ohne Span | Nr. 4 a und 7 d: unbekanntes Set trägt den Span mit `_.`; Nr. 4 c und 7 a: unbekannter oder fremder Workspace ohne Span | Nr. 3 f und 7 b: nach dem Beginn `End` genau einmal ohne Code; davor Nr. 7 a (Schritt (f)) ohne Span | Nr. 7 e | Nr. 4 b und 7 f: der gelesene Status, auch bei `ODR-E6003` des strikten Modus |
| Query-Span: Key-Lookup | wie Collection | Nr. 7 a (Header, Name) und Nr. 7 c: leerer Schlüsseltext ist `ODR-E6008` oder `ODR-E6003`, mit Span | wie Collection; ein Schlüssel ohne Treffer: Nr. 4 b (`SPEC-197` Nr. 3) | wie Collection | Nr. 7 e | Nr. 4 b: ohne Treffer der gelesene Status; mit Treffer und mit Tombstone `nicht_gelesen` (`SPEC-196` Nr. 7 c, e) |
| Query-Span: `/$count` | wie Collection | wie Collection | wie Collection | wie Collection | Nr. 7 e | wie Collection (`/$count` ist eine Collection-Abfrage, `SPEC-197` Nr. 2) |
| Query: `$metadata` | gilt nicht: kein Span, Nr. 2 c (keine Query-Ausführung, kein Set, keine Coverage); der Fehler des Headers ist Nr. 7 a | gilt nicht: kein Span, Nr. 2 c | gilt nicht: kein Span, Nr. 2 c | gilt nicht: kein Span, Nr. 2 c | gilt nicht: kein Span, Nr. 2 c | gilt nicht: kein Span, Nr. 2 c |
| Mutations-Span: `POST`, `PATCH`, `PUT`, `DELETE` | Nr. 7 a: Header fehlt `ODR-E5003`, ohne Span; ein Set fehlt im Pfad: die Paarung ist `ODR-E6002` (Schritt (e)), ohne Span | Nr. 7 a: leerer Header `ODR-E5003`; ein leerer Body ist `ODR-E6007` nach dem Beginn, Nr. 7 b, mit Span | Nr. 4 c und 7 d: unbekannter oder fremder Workspace ohne Span; ein unbekanntes Set `ODR-E6008` mit Span (Nr. 7 b, kein `entity_set` an diesem Span) | Nr. 3 f: Abbruch auch in `f`, `End` genau einmal, Rollback (`SPEC-192` Nr. 7) | Nr. 7 e | gilt nicht: der Span trägt kein Coverage-Attribut (Nr. 4); die Coverage entscheidet den Code (`SPEC-193` Nr. 2), nicht den Span |
| `Reset`, `Restore`, `Create`, `Delete` des Workspace, Snapshot `Create` und `List` | gilt nicht: kein Span, Nr. 2 (Entscheidung: kein `Telemetry` in `workspace.Konfig` und `snapshot.Konfig`); die Fehler dieser Operationen stehen in `SPEC-129` Nr. 12 | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: kein Span, Nr. 2 |
| Zähler je Fehlercode des Mock-Pfads | gilt nicht: der Dienst ruft `CountError` nie, Nr. 5 | gilt nicht: Nr. 5 | gilt nicht: Nr. 5 (die 16 Codes zählt der Adapter, `SPEC-230` Nr. 8) | gilt nicht: Nr. 5 (ein Kontextfehler ohne Code zählt auch beim Adapter nicht) | gilt nicht: Nr. 5 | gilt nicht: Nr. 5 (`ODR-E6003` zählt der Adapter, im Modus `hybrid` nur, wenn er ihn ausliefert) |
| Einträge `bereinigung_*` des Dienstes `workspace` (Log) | `SPEC-129` Nr. 9 a und b: ein `Delete` ohne Kennung ist `ODR-E7001`, kein Eintrag | wie links | `SPEC-231` Nr. 4: ein unbekannter Workspace schreibt nichts (`SPEC-129` Nr. 9 b); Q18 | `SPEC-129` Nr. 13: ein Kontextfehler der Bereinigung schreibt nichts | `SPEC-129` Nr. 14 g: ein zweites `Delete` ist `ODR-E7001`, kein zweiter Eintrag | gilt nicht: `Delete` liest keine Coverage |
| Verdrahtung des Adapters (`Build`, `Start`) | `SPEC-139` Nr. 13: `Build` ohne `Start` übergibt `io.Discard` | `SPEC-290` Nr. 1: nil `Telemetrie` ist `model.ErrUngueltig` | gilt nicht: der Root bekommt keinen Wert eines Aufrufers | `SPEC-139` Nr. 7: der Adapter startet keine Goroutine und hält keine Datei, das `Write` läuft im Aufruf (`SPEC-232` Nr. 5 d); `Herunterfahren` hat nichts zurückzubringen | gilt nicht: `Build` läuft einmal je `Handle` (`SPEC-139` Nr. 10) | gilt nicht: keine Coverage im Root |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Einbau trifft nebenläufige Pfade und Fakes, die `ctx` umhüllen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) sind offen; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) gilt für jede Wirkung, die der Detektor nicht meldet; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

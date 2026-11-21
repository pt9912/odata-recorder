# Slice slice-telemetry-einbau-mock-mutation: Mutations-Span `odata.workspace.mutation` um die Transaktion des Mock-Dienstes und Canary der Länge des Ortes (Teil 2 des Schnitts von `slice-telemetry-einbau-mock`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-FA-07`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-290` (Telemetrie des Mock-Dienstes; Nr. 2 b, 3, 4 c und d, 5, 6 T5 und T6, 8, 9) · `SPEC-030` (spezifikation.md §5) · `SPEC-139` Nr. 13 (Verdrahtung aus Teil 1, unverändert) · `SPEC-192` Nr. 3, 7, 9 · `SPEC-195` Nr. 1, 11 · `SPEC-180` Nr. 3 (Obergrenze 256 Bytes des Ortes) · `SPEC-230` Nr. 8 · `LH-FA-07.b` (Vertragsmatrix der Mutationen) · `LH-QA-05.a` (Fehlercodes) · `SPEC-044` · `ARC-002`, `ARC-003`, `ARC-021` · `architecture.md` §4.4, §4.5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md), Vorschlag des Architect vom 2026-11-21). **Datum:** 2026-11-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `MockUseCase` setzt den Span `odata.workspace.mutation` (`SPEC-030`, `SPEC-290`, Attribute `recorder.workspace_id`, `recorder.mutation_type`) um jede der vier Mutationen `POST`, `PATCH`, `PUT` und `DELETE` des Mock-Dienstes, samt Transaktion (`WithWorkspaceTransaction`; der Span liegt im Dienst um die Transaktion, nie in `f`, `SPEC-290` Nr. 3); eine abgewiesene Mutation (Constraint, Coverage, Abbruch) trägt den Span mit ihrem Ergebnis; der Dienst zählt keinen Fehler, den Zähler je Fehlercode führt der Adapter, der das Fehlerobjekt schreibt (`SPEC-290` Nr. 5). Kein Body, kein Name oder Wert einer Property und kein Key-Wert wird Attribut. Der Slice setzt den Teil 1 voraus: das Feld `Telemetrie` der `Konfig` des Mock-Dienstes, die nil-Prüfung, den Journal-Fake der Telemetrie und die Verdrahtung des Adapters des Prozesses in den Modi `mock` und `hybrid`; er ändert die Verdrahtung nicht.

**Übernimmt (Sendung von `slice-telemetry-port`, Teil Mutation; Geber: `slice-telemetry-port`; Schnitt des Planners vom 2026-11-21):** der Einbau in die Pfade Workspace-Mutation. Der Einbau in den Pfad `mock` (Query) liegt in [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md).

**Übernimmt (Sendung von `slice-telemetry-log`, akzeptiertes Negativ L15, Teil Mutation; Geber: `slice-telemetry-log`; Eintrag des Planners vom 2026-11-07, geteilt am 2026-11-21):** der Fänger je Stelle der Mutation: ein wohlgeformter Wert in einem zulässigen Schlüssel geht durch die Form des Adapters (`SPEC-232` Nr. 2, *Ungeprüft*), der Schutz ist ein Test je Stelle des Mutations-Pfads (`POST`, `PATCH`, `PUT`, `DELETE`), dass kein Wert aus dem Request (Name und Wert einer Property im Body, `Content-Type`, `Prefer`, `If-Match`, `If-None-Match`, dazu Set, Schlüssel, Header `X-Recorder-Workspace`, `Accept`) in einem Attribut oder in `msg` einer Zeile landet, außer nach Prüfung gegen das Modell. Der Fänger der Stellen der Query liegt in Teil 1.

**Übernimmt (Sendung von `slice-mock-schreiben`, Rückgabe R2 / W66, Review F-15; Geber: `slice-mock-schreiben`; Eintrag des Planners vom 2026-11-07, §1 *Ausdrücklich NICHT* und DoD des Nehmers vorher gelesen, kein Ausschluss trifft die Sendung; beim Schnitt am 2026-11-21 an diesen Teil gelegt):** der Fänger an den Wegen Log und Span für die Länge des Ortes im `Error()` des Eingabe-Lesers: ein Canary „Name einer unbekannten Property von `grenze` Bytes“ (`SPEC-180` Nr. 3, Obergrenze 256 Bytes des Ortes) geht durch den Mutations-Pfad des Mock-Dienstes (`POST`, `PATCH`, `PUT`), und weder eine Log-Zeile (`msg`, Attribute) noch ein Span trägt mehr als den Code und die gekürzte Meldung. Der Dienst schreibt selbst nichts in Log, Span oder Audit (W66, akzeptiertes Negativ dort).

**Übernimmt (Sendung von `slice-telemetry-einbau-mock`, Orakel auf dem Stand ohne Span der Mutation; Geber: `slice-telemetry-einbau-mock`; Eintrag des Planners vom 2026-11-21 bei der Closure von Teil 1, §1 *Ausdrücklich NICHT* und DoD dieses Teils vorher gelesen, kein Ausschluss trifft die Sendung):** die Orakel, die Teil 1 auf den Stand ohne Span der Mutation setzt, heben sich hier auf den Endzustand von `SPEC-290` Nr. 6: in `TestTelemetrieImQuelltext` (`telemetrie_gleichzeitig_test.go`, Zeile Q12 des Plans von Teil 1) zählt der Quelltext des Pakets ohne `_test.go` `StartSpan`, `Set` und `End` je 2 statt 1, 2 und 1; die Spalte `span` von `codeSzenarien` (`telemetrie_ablauf_test.go`) ist für die vier Mutationen 1 statt 0; der Parameter `span` in `TestKeinZaehlerBeiKontextfehler` ist für die vier Mutationen 1 statt 0. Die Hebung ändert bestehende Tests von Teil 1, nicht seinen Vertrag; ein fehlender Span der Mutation (Mutation `PATCH` und `PUT` ohne Span der Zeile M1) färbt `TestTelemetrieImQuelltext` und die Zeilen von `codeSzenarien` rot, das sieht der Implementer von Teil 2 vor der Übergabe.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Feld `Telemetrie` der `Konfig`, nil-Prüfung, Query-Span `odata.mock.query`, Verdrahtung im Composition Root und Einträge des Dienstes `workspace` — [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) (Teil 1, liegt bei diesem Start in `done/`); dieser Teil baut darauf auf und ändert sie nicht.
- Port und Adapter `Telemetry` — `slice-telemetry-port`.
- Zeile, Stufen, Senke, `Handler`, Spans, Verteilung, Ring und Span-Zeile im Adapter — `slice-telemetry-log`, `slice-telemetry-handler` und `slice-telemetry-spans`; alle drei liegen beim Start in `done/`, dieser Slice ändert sie nicht.
- Der Eintrag `audit_fehlgeschlagen` und die Abbildung des Adapters `audit` (`SPEC-231` Nr. 4 und 7, Sendung von `slice-audit-log`) — `slice-telemetry-einbau-gateway`; die Löschung eines Workspace ist kein Mutations-Span.
- Gesamtspan, Capture-Span und Match-Span (`SPEC-026` bis `SPEC-028`) — `slice-telemetry-einbau-gateway`; der Gesamtspan liegt im HTTP-Adapter und ist modusunabhängig.
- Spans und Zähler des Hybrid-Fallbacks — kein Slice von `welle-betrieb` nimmt sie an (Sendung ohne Nehmer, siehe §6 von `slice-telemetry-einbau-gateway`, Adresse: Planner).
- Mutations-Spans der Admin-Operationen (`Reset`, `Restore`, `Create`, `Delete` des Workspace, Snapshot-`Create`): keine Operation von `WorkspaceUseCase` oder `SnapshotUseCase` trägt einen Span (Entscheidung des Architect, `SPEC-290` Nr. 2); ein Vertrag, der `Reset` oder `Restore` einbezieht, ergänzt die Werte von `recorder.mutation_type`, das gehört `welle-admin-api`, falls sie es will.
- Der Zähler je Fehlercode für die Fehler des Mock-Pfads — der HTTP-Adapter, der das Fehlerobjekt schreibt (`slice-telemetry-einbau-gateway`); der Dienst `mock` ruft `CountError` nie (`SPEC-290` Nr. 5, `SPEC-230` Nr. 8).
- Ein Handbuch-Kapitel zu den Logzeilen — `slice-handbuch-telemetrie` (Entscheidung des Users vom 2026-11-21: eigener Slice; Entscheidung des Planners vom 2026-11-21 (§6): keine Zeile dieses Slice). Die Zeile des Spans je Mutation ist dort nur mit gefundenem Workspace sichtbar und steht im Kapitel erst, wenn ein Beispiel sie trägt (§6 dort).

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

- [ ] Mutations-Span: jede Mutation des Mock-Dienstes (`POST`, `PATCH`, `PUT`, `DELETE`, `LH-FA-07.b`), die die Zuordnung bestand, setzt `odata.workspace.mutation` mit der Kennung des gelesenen Workspace und der Art (`post`, `patch`, `put`, `delete`) um die Transaktion; eine abgewiesene Mutation (Constraint, Coverage, Abbruch) trägt den Span mit ihrem Ergebnis, und der Span endet auf jedem Rückweg genau einmal; der Dienst ruft `CountError` nie, der Zähler je Fehlercode gehört dem Adapter, der das Fehlerobjekt schreibt (`SPEC-290` Nr. 5, Zeile Q2 von Teil 1); der Span ist Kind des Gesamtspans über das `ctx` des Aufrufs, gleichzeitige Mutationen teilen nichts (M1 bis M7, M10); die Orakel von Teil 1 sind auf den Endzustand gehoben (Sendung von `slice-telemetry-einbau-mock`: `TestTelemetrieImQuelltext` zählt `StartSpan`, `Set` und `End` je 2, die Spalte `span` von `codeSzenarien` und der Parameter `span` in `TestKeinZaehlerBeiKontextfehler` sind für die Mutationen 1).
- [ ] Fänger über die Mutation: kein Request-Wert ungeprüft in einem Attribut oder in `msg` (Sendung L15 von `slice-telemetry-log`, M8, Canary aus Bytewerten), und der Canary „Name einer unbekannten Property von 256 und mehr Bytes“ durch `POST`, `PATCH` und `PUT` zeigt in Log und Span höchstens die gekürzte Meldung (Sendung R2 von `slice-mock-schreiben`, W66, M9).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: der Slice führt keinen Fehlercode, keinen Schlüssel und kein Kommando ein; die Zeile des Spans je Mutation beschreibt kein Kapitel dieses Slice (Entscheidung des Planners vom 2026-11-21, §6; das Kapitel zu den Logzeilen liefert `slice-handbuch-telemetrie`); das Kataloge-Gate ist grün, und §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando des Binary dazukam.
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
| `src/hexagon/services/mock/` | update | Mutations-Span um die Transaktion (Beginn nach Schritt (f), `End` auf jedem Rückweg, nie in `f`), Abbildung der Methode auf die vier Arten (`SPEC-290` Nr. 2 b, 3, 4 d); das Feld `Telemetrie` und die nil-Prüfung liefert Teil 1 |
| `src/hexagon/services/workspace/` | keine Änderung | Entscheidung des Architect (`SPEC-290` Nr. 2): keine Operation von `WorkspaceUseCase` trägt einen Span |
| `src/bootstrap/` | nur Tests | die Verdrahtung des Adapters steht aus Teil 1; die Tests durch den Composition Root mit dem echten Adapter (M5, M7, M8, M9) |
| `spec/spezifikation.md` | geliefert (Architect) | `SPEC-290` und Nachzüge; dieser Teil schreibt sie nicht |
| Tests | update, neu | Fakes der Ports und der Journal-Fake der Telemetrie aus Teil 1; der Test durch die Mutation, der Test durch den Composition Root mit dem echten Adapter |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-21 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-290`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind entschieden, der Implementer entscheidet keine im Code. Die Entscheidungen des Architect (Aufträge 1 bis 7, je mit Ort in `spec/spezifikation.md`) stehen in §3 von [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md); für diesen Teil gelten Nr. 1 (Mutation und ihre Art), Nr. 2 (Grenzen der Spans), Nr. 3 (Attribute `workspace_id`, `mutation_type`), Nr. 4 (Fehlerzähler) und Nr. 5 (Aufrufstellen T5 und T6, Rückwege der Mutation).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-21, Zeilen M1 bis M10 des ungeschnittenen Plans, vom Planner beim Schnitt unverändert übernommen; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). `Nr.` ist die Kennung des ungeschnittenen Plans; die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| M1 | (Nr. 2 b, 6 T5 und T6) Jede der vier Mutationen öffnet `odata.workspace.mutation` genau einmal und beendet ihn genau einmal; kein `Set`, genau die zwei Pflicht-Schlüssel | Schleife 4 Methoden: `StartSpan` 1, `End` 1, `Set` 0, Name `SpanWorkspaceMutation`, Schlüssel genau die aus `support.Spans()`; Orakel Literal | `PATCH` und `PUT` ohne Span; Name `SpanMockQuery`; `Set(mutation_type)` nach der Transaktion | andere |
| M2 | (Nr. 3 a bis c) Der Span umschließt die Transaktion: Beginn nach `Workspaces.Get` und vor `Metadaten.Get`, Ende nach der Rückkehr von `WithWorkspaceTransaction`; in `f` kein Aufruf von `Telemetrie` | Journal-Fake: Folge je Methode als Literal (`Workspaces.Get`, `StartSpan`, `Metadaten.Get`, …, Beginn der Transaktion, …, Ende der Transaktion, `End`); zwischen Beginn und Ende der Transaktion steht keine Operation des Fakes; 4 Folgen | `StartSpan` in `f`; `End` vor der Transaktion; `End` in `f` | andere |
| M3 | (Nr. 3 b) Genau ein `End` auf jedem Rückweg der Mutation, mit Fehler mit Code und mit Kontextfehler | M-Rückwege-Mutation: 13 Stellen (S3 bis S5, S17 bis S26) mit Fehler mit Code und 12 mit Kontextfehler (alle außer S25) gleich 25 Fälle; je Fall `End` 1, Rollback (der Fake der Transaktion), Ergebnis unverändert; Zählprüfung 13 und 12 als Literal des Tests (Negativ 4) | `End` nur im Erfolgsweg; `End` vor dem Fehler-Zweig von `f` | andere |
| M4 | (Nr. 4 d) `recorder.mutation_type` ist `post`, `patch`, `put` oder `delete`, nie der Text des Requests | Schleife 4 Methoden, Orakel Literal je Methode, Zählprüfung 4; dazu `post` in Kleinschreibung als Methode: kein Span (`ODR-E6002`, `SPEC-192` Nr. 3) | rohe Methode (Form `Token` verletzt, `wert`); `patch` und `put` vertauscht; Wert `update` | andere |
| M5 | (Nr. 4 c) `recorder.workspace_id` ist die Kennung des gelesenen Workspace, nicht der Header; bei unbekanntem oder fremdem Workspace kein Span | Fall: der Fake liefert zu Header `ws-kopf` einen Workspace `ws-gelesen` (Fixture ungleich dem Wert der Mutation): Attribut `ws-gelesen`; unbekannt (`ODR-E7001`) und fremd (`ODR-E5004`) mit Canary im Header: kein Span, Canary in keinem Attribut; Kennung mit Leerzeichen mit dem echten Adapter: `Dropped[wert]` 1, `Dropped[pflicht]` 1, Antwort unverändert; Workspace nach Schritt (f) gelöscht: Span trägt die Kennung, `ODR-E7001` | Header-Wert als Attribut; `StartSpan` vor `holeWorkspace` mit dem Header | andere |
| M6 | (Nr. 3 f, 9 b) Ein Abbruch mitten in der Transaktion ändert das Ergebnis nicht (Kontextfehler ohne Code, Rollback), und der Span endet genau einmal | deterministisch: der `Put`-Fake beendet den `ctx`, 4 Methoden; Wirkungs-Fänger: 16 Goroutinen × 4 Methoden, die Hälfte mit Abbruch in S22, in 20 Läufen unter dem Detektor: `End` je Aufruf 1, Revision und Entity-State der abgebrochenen unverändert, der übrigen plus 1 (Rate je Mutation im Plan) | `End` nur im Erfolgsweg; `End` unter `if ctx.Err() == nil`; Ergebnis des Aufrufs aus dem Span | nebenläufig |
| M7 | (Nr. 3 d und e) `StartSpan` und jeder Port der Mutation erhalten das `ctx` des Aufrufs unverändert; der Span ist Kind des Gesamtspans | Schleife über die 9 Stellen S17 bis S24 und S26 und `StartSpan` (Zählprüfung: 11 + 9 gleich 20 mit Q7 aus Teil 1); durch den Root mit dem echten Adapter: `ParentID` gleich der `ID` des Elternspans | `ctx, span := Telemetrie.StartSpan(ctx, …)`; `StartSpan(context.Background(), …)` | andere |
| M8 | (Nr. 8; Sendung L15) Kein Wert aus dem Request erreicht ein Attribut oder die Senke, auch nicht über die Mutation | Menge M-Stellen-Mutation, 6 Stellen (Name und Wert einer Property im Body, `Content-Type`, `Prefer`, `If-Match`, `If-None-Match`) und die vier aus Q14 (Teil 1), die eine Mutation liest (Set, Schlüssel, Header `X-Recorder-Workspace`, `Accept`), je durch jede Methode, die sie liest; Canary und Prüfung wie Q14; Zählprüfung 6 + 4 | `Str("recorder.workspace_id", <Wert der Stelle>)` je Stelle (10 Mutationen) | andere |
| M9 | (Nr. 8; Sendung R2 von `slice-mock-schreiben`) Der Name einer unbekannten Property von `grenze` Bytes durch `POST`, `PATCH` und `PUT` erreicht weder eine Zeile noch einen Span | Test durch den Root mit dem echten Adapter: 3 Methoden × 4 Längen (255, 256, 257 und 4096 Bytes aus Bytewerten) gleich 12; die Senke (alle Zeilen), die Attribute des Fakes und `msg` tragen nicht die ersten 8 Bytes des Namens; Antwort `ODR-E6007` | Attribut aus `err.Error()` der Abbildung; `Log` im Zweig `liesBody` mit der Ursache | andere |
| M10 | (Nr. 9 a) *n* gleichzeitige Mutationen ergeben *n* Spans mit Art und Kennung ihrer Anfrage | 16 Goroutinen × 4 Methoden in verschiedenen Workspaces, 20 Läufe unter dem Detektor: der Fake führt Buch je Span (`mutation_type` und `workspace_id` gleich der Anfrage, `End` je Span 1) | Span im `Mocker` gehalten; Attribute aus der letzten Anfrage | nebenläufig |

Zeilen gesamt 10: *a* = 2 (M6, M10), *b* = 8.

**Verteilung beim Schnitt** (Planner, 2026-11-21): Die Zeile Q12 von Teil 1 zählt im Quelltext des Pakets `StartSpan` 1, `Set` 2 und `End` 1 (T1 bis T4); dieser Teil ergänzt T5 und T6 und hebt die Literale des Tests von Q12 auf je 2 (`StartSpan`, `Set`, `End`: 2, 2, 2), wie `SPEC-290` Nr. 6 sie für den Endzustand zählt. Der Test von Q12 gehört Teil 1, die Änderung der Literale diesem Teil; der Plan von Teil 1 nennt das ebenfalls (§3 dort).

**Mengen mit Zählprüfung** (Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15):

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| M-Op (M1; mit Q3 und Q4 aus Teil 1) | 4 Mutationen; mit den 6 Lesezugriffen und `$metadata` × 2 aus Teil 1 gleich 12 | Span ja, Name | 4; zusammen 12 mit Teil 1, die Zahl der Methoden aus `SPEC-192` Nr. 3 |
| M-Rückwege-Mutation (M3) | 13 Stellen mit Code, 12 mit Kontextfehler gleich 25 | `End` 1, Rollback | 13 und 12 |
| M-ctx (M7; mit Q7 aus Teil 1) | S17 bis S24 und S26 gleich 9; mit S2 bis S12 (11) aus Teil 1 gleich 20, dazu `StartSpan` je Art | Identität des `ctx` | 9 als Literal; 20 zusammen (`SPEC-195` Nr. 11: „20 ein `ctx`“, Negativ 4) |
| M-Art (M4) | 4 Methoden | `post`, `patch`, `put`, `delete` | 4 |
| M-Stellen-Mutation (M8; mit Q14 aus Teil 1) | 6 Stellen + 4 Stellen der Query, die eine Mutation liest gleich 10 | kein Canary in Attribut und Senke | 14 Stellen insgesamt gleich `SPEC-290` Nr. 8 (Q14 acht, M8 sechs) |
| M-Canary (M9) | 3 Methoden × 4 Längen gleich 12 | kein Fund des Namens | 12 |
| M-Gleichzeitig (M10; mit Q13 aus Teil 1) | 16 Goroutinen × 4 Methoden, 20 Läufe | Buch je Span | 16 × 4 |

**Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** der Body (Prüfgrenze 4096 Bytes des Namens in M9; Grund: die Länge des Namens ändert den Weg nicht, die Grenze des Bodys ist `MaxBytes`; Folge: ein Name von 1 MiB wird nicht eigens geprüft). Der Name eines Sets im Request ist Gegenstand von Teil 1 (Q9).

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; die Rangfolge steht in `SPEC-195` Nr. 3, `SPEC-192` Nr. 2, `SPEC-196` Nr. 2 und `SPEC-197` Nr. 2):

| Paar oder Modus | Erwartung | Fänger |
|---|---|---|
| `PATCH` auf fehlenden Eintrag, Coverage `unknown` | `ODR-E6003`, der Span trägt kein Coverage-Attribut | M3 |
| Fehler mit Code und Kontextfehler zugleich (`SPEC-109` Nr. 6) | Kontextfehler ohne Code, `End` 1 | M3 (und Q6 aus Teil 1 für die Query) |
| Modus `mock` und `hybrid` | beide tragen die Zeile des Spans der Mutation | M5, M8 und M9 durch den Root (der Modus ist Teil des Aufbaus; die Zeile des Spans der Query prüft Q15 in Teil 1) |

**Akzeptierte Negative dieses Vertrags** (Grund und Folge): (1) Ein Typed-nil in `Telemetrie` (`SPEC-290` *Ungeprüft*): Teil 1 prüft nil in `New`; Folge: eine Panik beim ersten Span, im Betrieb unerreichbar (der Root übergibt den Adapter). (2) Der Span des Hybrid-Fallbacks (`SPEC-290` Nr. 7 g, *Ungeprüft*): kein Nehmer, siehe §6. (3) Die Zahlen der Mengen (13 und 12 Rückwege, 9 Stellen mit `ctx`, 14 Stellen des Requests) gegen die Tabellen der Spezifikation prüft kein Test, weil die Stufe `source` des Images `spec/` nicht trägt; Folge: ändert jemand die Tabelle von `SPEC-195` Nr. 11 ohne den Test, sieht es erst ein Review (einmal von Hand gezählt am 2026-11-21: 26 Stellen, davon 13 hinter dem Beginn der Mutation, 20 mit `ctx` zusammen mit der Query). (4) Dauer und Ausfall der Senke wie Q21 und Q22 in Teil 1 (die Grenzen des Spans sind durch die Folge der Aufrufe belegt, nicht durch gemessene Zeit; kein Dienst bildet sein Ergebnis aus `Telemetry`).

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Mutationsart, jeder Fehlercode des Mutations-Pfads), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) (Teil 1) liegt in `done/`; mit ihm liegen die Bedingungen von Teil 1 in `done/` (`slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans`, `slice-mock-lesen`, `slice-mock-schreiben`, `slice-snapshot-use-case`, `slice-bootstrap-m3`). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-290`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-einbau-mock` (Teil 1): *Code-Abhängigkeit*. Das Feld `Telemetrie` und die nil-Prüfung der `Konfig`, der Journal-Fake der Telemetrie (`f` des Schnitts liegt dort) und die Verdrahtung des Adapters in den Modi `mock` und `hybrid`; ohne sie gibt es weder einen Ort für den Aufruf noch einen Test durch den Composition Root.
- `slice-mock-schreiben` (`welle-mock`): der Ort des Spans (`src/hexagon/services/mock`), die Mutation und ihre Transaktion.
- `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans` und `slice-telemetry-port`: Port, Fake und Adapter, die dieser Teil ruft; über Teil 1 transitiv erfüllt.

**Reihenfolge und Konflikte** (Planner, 2026-11-21): Dieser Teil liegt nach Teil 1 im selben Worktree (Übergangsregeln Nr. 3). Er ändert in `src/bootstrap/` nur Tests und in `src/hexagon/services/mock/` Code; `slice-listener-tls-verdrahtung` ändert in `src/bootstrap/` `konfig*.go`, `wurzel.go` und `lebenszyklus.go`. Die Dateien sind verschieden, die beiden Teile dürfen zugleich laufen; ändert einer die Datei des anderen, laufen sie nacheinander (`welle-betrieb` §5). Wer `main` nachzieht, misst die fremden Zeilen unter `src/bootstrap/` mit (`SPEC-128` Nr. 10 g): er misst vor dem Merge, nennt den Messpunkt und trägt danach, wenn nötig, `ueberschreitung` mit Grund ein (Nr. 5 d), keine Rückführung.

**Größenmarke:** noch nicht gesetzt; der Teil des Architect steht im Block unten (Zeilen M1 bis M10, `SPEC-128` Nr. 3 k), `basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke`. Vorgerechnet beim Schnitt (Planner, 2026-11-21, `SPEC-128` Nr. 1): 10 Zeilen, *a* = 2, *b* = 8, *f* = 0, *p* = 25 (Mutations-Span mit Abbildung der Methode rund 25): *S* = 85·2 + 80·8 + 200·0 + 25 = 835, *M* = 100 · ⌈135·835 / 10000⌉ = 100 · ⌈11,27⌉ = 1200; keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/hexagon/services/mock/` und `src/bootstrap/`, Tests eingeschlossen. Den Block setzt der Planner beim Start (Prosa-Marke und Block ersetzen einander nicht, der Block steht allein).

```groessenmarke
pfad: src/hexagon/services/mock/
pfad: src/bootstrap/
produktcode: 25
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (`AGENTS.md` §3.14); ein weiterer Schnitt ist nicht vorgesehen, die Schätzung liegt weit unter der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Span lässt sich nicht um die Transaktion legen, ohne dass `f` einen Typ des Adapters kennt (`SPEC-290` Nr. 3 c entscheidet das Gegenteil: der Span liegt um die Transaktion, nicht in ihr).

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
- `recorder.workspace_id` stammt aus dem Request; ein Wert aus dem Request darf nur als Attribut stehen, wenn er geprüft ist — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-290` Nr. 4 c): die Kennung kommt aus dem gelesenen Workspace, ohne gefundenen Workspace gibt es keinen Span; Fänger M5, M8 mit Canary aus Bytewerten.
- Der Span der Mutation ist nur dann Kind des Gesamtspans, wenn das `ctx` des Aufrufs ihn trägt — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-290` Nr. 3 d, e): `StartSpan` erhält das `ctx` des Aufrufs, kein Port erhält das gelieferte `ctx`; Fänger M7 (der Test durch den Composition Root mit dem echten Adapter).
- Pfad ohne Lauf unter dem Detektor (Abbruch, Ausfall der Senke) — `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) — **Ausgang:** entschieden (Architect, 2026-11-21): der Abbruch hat einen Wirkungs-Fänger (M6, die Hälfte der 16 Goroutinen bricht in S22 ab, `End` je Aufruf 1, Zustand unverändert); der Ausfall der Senke ist ein akzeptiertes Negativ mit Grund und Folge (Zeile Q21 in Teil 1).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-21): das Raster unten trägt in jeder Zelle einen Ort oder `gilt nicht` mit Grund.
- Der Zähler je Fehlercode des Mock-Pfads liegt beim Adapter, nicht beim Dienst — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-290` Nr. 5); Nehmer `slice-telemetry-einbau-gateway` trägt den Zähler in Ziel und DoD (gelesen am 2026-11-21, `AGENTS.md` §3.13); der Fänger des Zählers im Mock-Pfad ist das akzeptierte Negativ Q20 in Teil 1.
- Schnitt und Größenmarke: der Plan ist am 2026-11-21 vor dem Code aus `slice-telemetry-einbau-mock` geschnitten (ungeschnitten *S* = 2870 über der Schnittgrenze; dieser Teil *S* = 835, Teil 1 *S* = 2035) — **Ausgang:** weiter offen; Adresse: Planner (setzt beim Start den Block mit `basis`, `schaetzung` und `marke`, `AGENTS.md` §3.14).
- Handbuch: die Zeile des Spans je Mutation ist eine weitere Logzeile des Binary; ein Kapitel dazu ist nur als gebundenes Beispiel zulässig (`SPEC-200` Nr. 1), und das Beispiel-Format kennt keine Anfrage an den Listener (`SPEC-200` Nr. 6 d, 7 b) — **Ausgang:** eingetreten → Folge-Slice `slice-handbuch-telemetrie` (Entscheidung des Users vom 2026-11-21, `welle-betrieb` §5, Punkt c): dieser Slice schreibt kein Kapitel, §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando dazukam; der Nehmer trägt das Kapitel in §1 (*Übernimmt*, Geber dieser Slice, gelesen am 2026-11-21, kein *Ausdrücklich NICHT* dort trifft die Sendung, `AGENTS.md` §3.13); dass die Zeile des Spans je Mutation ohne Workspace nicht als Beispiel erreichbar ist, steht dort in §1 und §6.
- Span des Hybrid-Fallbacks (nicht der Versuch im Mock): kein Slice von `welle-betrieb` nimmt ihn an (§1, Sendung ohne Nehmer, geführt in `slice-telemetry-einbau-gateway` §6) — **Ausgang:** weiter offen; Adresse: Planner (Frage an den User, dort gestellt).
- Gleichzeitiger Lauf mit `slice-telemetry-einbau-mock` (Teil 1) ist ausgeschlossen (Start-Trigger), mit `slice-listener-tls-verdrahtung` zulässig, solange keiner die Dateien des anderen ändert (§4) — **Ausgang:** weiter offen; Adresse: Planner (beim Start des zweiten Teils die Dateien beider vergleichen und die Reihenfolge nennen, `welle-betrieb` §5).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-21, Zeilen der Mutation des ungeschnittenen Plans; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund; „Nr.“ ohne Nennung heißt `SPEC-290`):

| Operation | fehlender Wert (Workspace, Entity-Set) | leerer Wert | unbekannter oder fremder Wert | Abbruch (`ctx`) | Wiederholung | Coverage `partial`/`unknown` |
|---|---|---|---|---|---|---|
| Mutations-Span: `POST`, `PATCH`, `PUT`, `DELETE` | Nr. 7 a: Header fehlt `ODR-E5003`, ohne Span; ein Set fehlt im Pfad: die Paarung ist `ODR-E6002` (Schritt (e)), ohne Span | Nr. 7 a: leerer Header `ODR-E5003`; ein leerer Body ist `ODR-E6007` nach dem Beginn, Nr. 7 b, mit Span | Nr. 4 c und 7 d: unbekannter oder fremder Workspace ohne Span; ein unbekanntes Set `ODR-E6008` mit Span (Nr. 7 b, kein `entity_set` an diesem Span) | Nr. 3 f: Abbruch auch in `f`, `End` genau einmal, Rollback (`SPEC-192` Nr. 7) | Nr. 7 e | gilt nicht: der Span trägt kein Coverage-Attribut (Nr. 4); die Coverage entscheidet den Code (`SPEC-193` Nr. 2), nicht den Span |
| Zähler je Fehlercode des Mock-Pfads (Mutation) | gilt nicht: der Dienst ruft `CountError` nie, Nr. 5 | gilt nicht: Nr. 5 | gilt nicht: Nr. 5 (die 16 Codes zählt der Adapter, `SPEC-230` Nr. 8) | gilt nicht: Nr. 5 (ein Kontextfehler ohne Code zählt auch beim Adapter nicht) | gilt nicht: Nr. 5 | gilt nicht: Nr. 5 |

Die Zeilen zur Query, zu `$metadata`, zu den Operationen `Reset`, `Restore`, `Create`, `Delete` des Workspace und Snapshot, zu den Einträgen `bereinigung_*` und zur Verdrahtung des Adapters stehen in §6 von [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md); die Zeilen der Admin-Operationen (`gilt nicht`, Nr. 2) gelten für beide Teile.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-21, Zähler wie im Plan von [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) §8; der Planner zieht Treffer und Zähler beim Start nach). Der Einbau trifft nebenläufige Pfade und Fakes, die `ctx` umhüllen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) sind offen; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) gilt für jede Wirkung, die der Detektor nicht meldet; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

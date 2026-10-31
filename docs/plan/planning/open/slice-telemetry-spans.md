# Slice slice-telemetry-spans: Spans, Verteilung, Ring und Span-Zeile des Adapters von `Telemetry`

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-QA-01`, `LH-FA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-230` Nr. 5 bis 7 (Spans, Kontext und Eltern-Kind, Ende), Nr. 8 (Verteilung), Nr. 9 (`Spans` des Stands) · `SPEC-231` Nr. 6 (Span-Zeile) · `SPEC-232` Nr. 1 (Ring), Nr. 4 (Grenzen), Nr. 5 b, e und f (Stand, `ID`s und keine Sperre über `Clock.Now()` unter Nebenläufigkeit), Nr. 2 und 3 für die Wege der Spans · `SPEC-026` bis `SPEC-030` (spezifikation.md §5: die fünf Spans und ihre Pflicht-Attribute) · `LH-QA-01.a` (Gesamtspan als Messpunkt, über die Span-Zeile lesbar) · `ARC-020`.
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

**Ziel:** Der Adapter von `Telemetry` (`src/adapters/driven/telemetry`, `ARC-020`) macht aus dem wirkungslosen `StartSpan` die Spans von `SPEC-026` bis `SPEC-030`: `StartSpan`, `Set` und `End` mit geschlossener Menge der fünf Namen und ihrer Pflicht-Attribute, der `ctx` des Aufrufs leitet ab und trägt den Span, Eltern-Kind über den `ctx`, Dauer nach der `Clock`, Verteilung je Name mit zwölf Eimern, Ring der zuletzt beendeten `MaxSpans` Spans; jedes `End` schreibt eine `INFO`-Zeile mit `span_id`, `parent_id` und `duration_ms`, aus der ein anderer Prozess (der Lasttest von `LH-QA-01.a`) den Gesamtspan und den Anteil von `odata.replay.match` und `odata.mock.query` liest; `Snapshot` zeigt Verteilung und Ring als einen Stand, auch bei gleichzeitigen Enden; ein Canary erreicht über keinen Weg der Spans eine Zeile, ein Attribut, eine Metrik oder einen Fehlerwert (`LH-QA-04`). Der Slice liefert den Adapter, nicht den Einbau: kein Dienst und kein Adapter des Produkts ruft `StartSpan`, bevor die Einbau-Slices laufen.

**Teil des Schnitts vom 2026-10-30 (Planner, vor dem Code):** Teil `S` der Kette `P` → `L` → `S`, den der Architect in `slice-telemetry-port` §4 vorgeschlagen hat. Der Vertrag `SPEC-230` bis `SPEC-232` gehört zu allen drei Slices; dieser liefert `SPEC-230` Nr. 5 bis 7, den Teil von Nr. 8 und 9, der an `End` hängt (Verteilung, Ring), `SPEC-231` Nr. 6, `SPEC-232` Nr. 5 b und e sowie 5 f für die Seite von `StartSpan` und `End`. `P` liefert Ports, Regeln der Attribute, `CountError`, `Snapshot` als Form, `New` und `FakeTelemetry`, `L` Zeile, Handler und Senke. Die Entscheidungen des Architect (1 bis 9, Alternativen, Gründe, Folgen) stehen in `slice-telemetry-port` §3 und gelten hier unverändert.

**Übernimmt (Sendung von `slice-telemetry-port`, Teil `S`):** `StartSpan`, `Set`, `End`, Eltern-Kind, Verteilung, Ring, Span-Zeile, Canary über die Wege der Spans und Sperre des Stands.

**Übernimmt (Sendung von `slice-telemetry-port`, Klasse *Stand zu einem Zeitpunkt*, `SPEC-230` Nr. 9):** Der Geber misst `TestSnapshotZeitpunkt` (P15) über die 38 Zähler, die sein Teil ändert (35 Codes von `Errors` seit `ODR-E5010` aus `slice-http-lebenszyklus`, die Gründe `fehlercode`, `schluessel`, `wert` von `Dropped`; die Zahlen folgen `model.AlleCodes()`, 33 und 36 vor `ODR-E6005` aus `slice-odata-url-pfad`); die Fenster der Sperre um `Durations` und `Spans` und um die Gründe `senke`, `pflicht`, `spanname` und `zustand` von `Dropped` sind in seinem Teil unbeobachtbar (Negativ N-1 dort), weil erst dieser Slice sie ändert. Dieser Slice führt den Fänger für *alle* Lücken zwischen `Errors`, `Durations`, `Dropped` und `Spans`, nicht nur für die zwischen Verteilung und Ring (S16): das Orakel von S16 reicht über die Kette der vier Sammlungen, einschließlich der vier Gründe und der Zählung von `Count`, `Counts`, `Sum` und `Spans` (Geber `slice-telemetry-port`, dort §3 N-1 und §6; `AGENTS.md` §3.13).

**Übernimmt (Sendung von `slice-telemetry-port`, Review F-10):** Ersatz des Teils „`StartSpan`, `Set`, `End` ohne Wirkung“ von `TestKernSchreibtNichts`: sobald die Spans Wirkung haben, hält dieser Teil des Tests im Geber nichts mehr fest, und dieser Slice ersetzt ihn durch die Fänger der Spans.

**Hinweis (Sendung von `slice-telemetry-port`, Verifikation V-8):** `filtern(op, ...)` im Adapter mit `op == ""` ist die Menge `Log` (`opLog = ""`); ein Span mit leerem Namen darf nie bis dorthin kommen.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ports, Attribut, Regeln der Schlüssel und Wertformen, `CountError`, `New`, `FakeTelemetry` — `slice-telemetry-port`; die Menge der Namen und die Pflicht-Attribute stehen dort als Daten in `support.Spans()`, dieser Slice benutzt sie.
- Zeile, Stufen, `Handler`, Senke und deren Ausfall, Sperre der Zeile — `slice-telemetry-log`; dieser Slice schreibt die Span-Zeile über deren Schreiber und ändert die Form der Zeile nicht.
- Einbau der Spans in HTTP-Adapter, Record- und Replay-Dienst (`slice-telemetry-einbau-gateway`), Mock-Dienst und Workspace-Mutation (`slice-telemetry-einbau-mock`) — der Port kennt die Namen als Daten und braucht dafür weder `MockUseCase` noch `WorkspaceUseCase`; beide Einbau-Slices starten nach diesem Slice.
- Spans und Zähler des Hybrid-Fallbacks — ein Span ist eine Zeile in der Tabelle von spezifikation.md §5 und ein Eintrag in `support.Spans()`; kein Nehmer in `welle-betrieb` (§6, Empfehlung des Architect: neuer Slice `slice-telemetry-einbau-hybrid` nach `slice-hybrid-service` und diesem Slice).
- Der Lasttest und seine Auswertung — `slice-lasttest` liest die Span-Zeile `odata.gateway.request`; dieser Slice liefert sie.
- Endpunkt, Netz-Port und Export an einen Collector — kein Lastenheft-Bedarf (`slice-telemetry-port` §1, Frage an den User dort).

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

- [ ] Spans: `StartSpan`, `Set`, `End` mit geschlossener Menge der fünf Namen und ihrer Pflicht-Attribute nach `SPEC-026` bis `SPEC-030` (fremder Name, fehlendes Pflicht-Attribut, fremder Schlüssel, ungültiger Wert), Eltern-Kind über den `ctx`, zweites `End` und `Set` nach `End` ohne Wirkung, Aufzeichnung auch bei beendetem `ctx`, `ID` je Span eindeutig (`SPEC-230` Nr. 5 bis 7, `LH-QA-05`).
- [ ] Verteilung, Ring und Span-Zeile: Dauer nach der `Clock`, zwölf Eimer mit der Regel „erster Eimer mit Dauer ≤ Grenze“, `Count` und `Sum` je Name, Ring der letzten `MaxSpans` Spans, eine `INFO`-Zeile je `End` mit `span_id`, `parent_id`, `duration_ms` und den Attributen; `Snapshot` ist ein Stand (Verteilung und Ring unter einer Sperre, und die Lücken zwischen `Errors`, `Durations`, `Dropped` und `Spans`, einschließlich der Gründe `senke`, `pflicht`, `spanname`, `zustand`; Geber `slice-telemetry-port`), unter `make test-race` ohne Befund (`SPEC-230` Nr. 8 und 9, `SPEC-231` Nr. 6, `SPEC-232` Nr. 5 b und e, `LH-QA-01.a`).
- [ ] Keine Geheimnisse und kein Ausfall: ein Canary erreicht über 4 Einbringungswege, 24 Schlüssel und 6 Ausgabewege weder Zeile noch Span-Attribut noch Metrik noch Fehlerwert; eine defekte Senke ändert Aufzeichnung und Ergebnis des Aufrufers nicht (`SPEC-232` Nr. 2 und 3, `LH-QA-04`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary, oder macht er Logfelder am Binary sichtbar: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Adapter wird erst vom Einbau verdrahtet; Logfelder und Span-Zeile werden dort sichtbar).
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
| `src/adapters/driven/telemetry/` | update | `StartSpan`, `Set`, `End`, Eltern-Kind, Verteilung, Ring, Span-Zeile über den Zeilenschreiber aus `slice-telemetry-log`, Stand unter einer Sperre |
| `spec/spezifikation.md` | unverändert | `SPEC-230` Nr. 5 bis 9, `SPEC-231` Nr. 6 und `SPEC-232` stehen; eine Berichtigung zieht der Implementer im selben Commit nach (`AGENTS.md` §3.9) |
| `spec/architecture.md`, `.a-check.yml`, `go.mod`, `go.sum`, `docs/plan/adr/` | unverändert | wie in `slice-telemetry-port` §3 |
| `src/hexagon/ports/support/`, `src/bootstrap/`, die Dienste | unverändert | Port und Daten (`support.Spans()`, `support.Eimer()`) stehen; der Einbau ist `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` |
| Handbuch | unverändert | Binary unverändert (Entscheidung 7 des Architect) |
| Tests | neu | je Zusage ein Fall (Tabelle unten); Canary-Test über die Wege der Spans; nebenläufige Tests laufen unter `make test-race` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation und wurden vom Architect geschrieben (Quelle: `slice-telemetry-port` bis zum Schnitt, Zeilen mit dem Präfix `S`); die Randformen entscheidet der Architect, nicht der Implementer.

**Mengen für die Fänger** (`AGENTS.md` §3.15, §3.17; die Zahl liefert jeweils die Tabelle oder der Wortlaut, eine Zählprüfung im Test vergleicht sie):

- Spans: 5 Namen mit 11 Pflicht-Attributen (gezählt aus der Tabelle in spezifikation.md §5), 55 Paare Span × Schlüssel, davon 11 zulässig und 44 fremd.
- Eimer-Grenzen: 11 (12 Eimer); Ring bei drei Plätzen mit 2, 3, 4 Enden; `MaxSpans` 1 und 65536.
- Operationen, die der Slice anfasst: `StartSpan`, `Set`, `End`, `Snapshot` (Teil `Durations` und `Spans`).
- Einbringungswege des Canary: 4 (Header, Body, Query-Literal, Konfigurationswert); Eingabestellen: 24 Schlüssel (13 von `Log`, 11 der Spans); Ausgabewege: 6 (Zeile über `Log`, Zeile über den `Handler`, Span-Zeile, `Spans` und `Attrs` des Stands, `Errors`, `Durations` und `Dropped` des Stands, Fehlerwert von `New`); 96 Paare Weg × Stelle, jedes über alle 6 Ausgabewege geprüft.
- Modi der Span-Zeile: 3 (`MinLevel` bis `INFO`, darüber, Senke defekt).
- Beide Seiten jeder Grenze: Ring bei 2, 3, 4 Enden, jede Eimer-Grenze und 1 ns darüber, die Dauer 0 (zwei gleiche Zeiten), `MinLevel` bei Gleichheit mit `INFO`.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Namen der Tests sind Vorschläge, der Fänger ist die beschriebene Schleife mit Orakel. Ein akzeptiertes Negativ zählt `andere`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| S01 | Die Menge der Span-Namen ist geschlossen; ein fremder Name liefert den `ctx`, einen wirkungslosen Span und zählt `spanname` (`SPEC-230` Nr. 5) | `TestSpanNamen`: Schleife über `support.Spans()` (Zählprüfung 5 und 11 Pflicht-Attribute gegen die Tabelle in §5) und 6 fremde Namen (leer, Großschreibung, mit Leerzeichen, `odata.gateway`, `odata.gateway.requests`, `odata.mock`); `Set` und `End` des wirkungslosen Spans zählen nichts | Namensprüfung mit `HasPrefix("odata.")` | andere |
| S02 | Fehlt ein Pflicht-Attribut, wird der Span aufgezeichnet und `pflicht` zählt je fehlendes (`SPEC-230` Nr. 5); ein Pflicht-Attribut mit ungültigem Wert zählt `wert` und beim Ende `pflicht` (Zeile 5 der Zusatztabelle in §6, `SPEC-230` Nr. 4) | `TestPflichtAttribute`: Schleife über die 11 Pflicht-Attribute (je einmal weggelassen) und der Fall „alle gesetzt“; Orakel Eintrag im Ring und `pflicht`; ein ungültiger Wert zählt `wert` beim Setzen und `pflicht` beim Ende | Span ohne Pflicht-Attribut wird verworfen | andere |
| S03 | An einem Span sind allein seine Pflicht-Attribute zulässig (`SPEC-230` Nr. 5) | `TestSpanSchluessel`: Schleife über 5 Spans × 11 Schlüssel (55 Paare); Orakel die Tabelle, `schluessel` 1 für jedes fremde Paar (44) | alle elf Schlüssel an jedem Span zulässig | andere |
| S04 | `Set` ersetzt einen gültigen Wert; ein ungültiger lässt den früheren stehen (`SPEC-230` Nr. 5) | `TestSetErsetzt`: Schleife über 11 Schlüssel × 2 Fälle (gültig nach gültig, ungültig nach gültig) | der erste Wert gilt | andere |
| S05 | `ParentID` ist die `ID` des Spans im `ctx`, sonst leer; ein beendeter Eltern-Span bleibt Eltern; ein fremder Wert im `ctx` macht den Span zur Wurzel (`SPEC-230` Nr. 6) | `TestElternKind`: 5 Fälle (Wurzel, Kind, Enkel, Eltern beendet, Span des Fakes im `ctx`) und `odata.replay.match` und `odata.mock.query` als Kinder des Gesamtspans (2 Fälle) | `ParentID` immer leer | andere |
| S06 | Der Rückgabe-`ctx` leitet den übergebenen ab; ein nil-`ctx` gilt als `Background` (`SPEC-230` Nr. 1, 6) | `TestSpanCtx`: 5 Fälle (Abbruch wirkt, Frist, Wert, nil, zwei Geschwister unabhängig) | `Background()` als Wurzel des Rückgabe-`ctx` | andere |
| S07 | Die Dauer ist Ende minus Anfang nach der `Clock`; die Eimer-Regel (`SPEC-230` Nr. 7, 8) | `TestEimer`: `FestUhr` und `Advance(d)` für 0, jede der 11 Grenzen und jede um 1 ns darüber (23 Fälle); Zählprüfung 11 Grenzen und 12 `Counts`; Orakel „erster Eimer mit Dauer ≤ Grenze“ | `<` statt `<=` bei der Wahl des Eimers | andere |
| S08 | `Count` ist die Summe von `Counts`, `Sum` die Summe der Dauern, je Name getrennt (`SPEC-230` Nr. 8, 9) | `TestVerteilung`: 5 Namen × 3 Enden mit festen Dauern; Orakel Zahl und Summe; nur der Eintrag des eigenen Namens steigt | `Sum` wird nicht addiert | andere |
| S09 | Ein zweites `End` und `Set` nach dem `End` sind ohne Wirkung und zählen `zustand` (`SPEC-230` Nr. 7) | `TestEndDoppelt`: Schleife über 5 Spans × 2 Fälle (10); Orakel Ring, Verteilung und Zahl der Zeilen unverändert | das zweite `End` zählt die Verteilung doppelt | andere |
| S10 | `End` zeichnet auch bei beendetem `ctx` und nach dem Ende des Eltern-Span auf (`SPEC-230` Nr. 7) | `TestEndNachAbbruch`: `ctx` abgebrochen, Frist abgelaufen, Eltern vor dem Kind beendet (3 Fälle); Orakel ein Eintrag im Ring | `End` verwirft bei `ctx.Err()` | andere |
| S11 | Ein Span ohne `End` wird nicht aufgezeichnet (`SPEC-230` Nr. 7) | `TestSpanOhneEnde`: 1000 Starts ohne `End`; Stand und Zeilen unverändert; dass der Adapter sie nicht hält, prüft das Review am Code | `StartSpan` trägt in den Ring ein | andere |
| S12 | Der Ring hält die zuletzt beendeten `MaxSpans`, den ältesten voran; `Spans` und `Attrs` des Stands sind Kopien (`SPEC-230` Nr. 9, `SPEC-232` Nr. 1) | `TestRing`: `MaxSpans` 3 mit 2, 3, 4 Enden und `MaxSpans` 1; Orakel Reihenfolge und Inhalt; Änderung an `Spans[i].Attrs` wirkt nicht zurück | behält die ersten statt der letzten | andere |
| S13 | Jedes `End` schreibt eine `INFO`-Zeile mit den Feldern von `SPEC-231` Nr. 6; unter `MinLevel` über `INFO` entfällt die Zeile, die Aufzeichnung bleibt; bei defekter Senke zählt `senke`, die Aufzeichnung bleibt | `TestSpanZeile`: Schleife über 5 Spans × 3 Modi (`MinLevel` bis `INFO`, darüber, Senke defekt; 15 Fälle); Orakel die Felder in der Reihenfolge der Tabelle in §5 und der Stand; `TestCanarySpans`: 4 Einbringungswege × 24 Schlüssel (96 Paare, Zählprüfung 4 und 24) über 6 Ausgabewege | Zeile für `odata.workspace.mutation` fehlt; Zeile erscheint unter `WARN`; ein ungültiger Wert eines Span-Schlüssels erscheint in der Span-Zeile | andere |
| S14 | Keine Span-Operation gerät in Panik (`SPEC-232` Nr. 3) | `TestSpanKeinePanik`: Schleife über `StartSpan`, `Set`, `End` × 5 Argumentarten | `ctx.Value` ohne nil-Prüfung | andere |
| S15 | Die Aufzeichnung eines Endes (Ring und Verteilung) liegt in einem Abschnitt unter der Sperre (`SPEC-232` Nr. 5 b, Seite von `End`) | `TestEndSnapshotGleichzeitig`: 16 Goroutinen beenden je 500 Spans, 8 lesen `Snapshot` in einer Schleife; Orakel je Stand `Count` gleich Summe von `Counts` und Zahl der Ring-Einträge gleich `Count` (`MaxSpans` 65536); Wirkungs-Fänger mit Rate, kein Haltepunkt (`SPEC-232` Ungeprüft) | `End` gibt die Sperre zwischen Ring und Verteilung frei | nebenläufig |
| S16 | Die Kopie im `Snapshot` liegt in einem Abschnitt unter der Sperre (`SPEC-232` Nr. 5 b, Seite von `Snapshot`) | `TestSnapshotEinStand`: derselbe Aufbau wie S15; Orakel dieselbe Invariante und die Kette der Zähler über `Errors`, `Durations`, `Dropped` und `Spans` (Sendung von `slice-telemetry-port`, §1); Wirkungs-Fänger mit Rate | `Snapshot` kopiert Verteilung und Ring in zwei Abschnitten | nebenläufig |
| S17 | Die `ID`s der Spans sind paarweise verschieden (`SPEC-230` Nr. 6, `SPEC-232` Nr. 5 e) | `TestSpanIDs`: 32 × 500 Spans gleichzeitig; Orakel paarweise verschieden und die Form `spn-` mit 16 Ziffern; Detektor | Zähler ohne atomare Erhöhung | nebenläufig |
| S18 | Keine Sperre des Adapters liegt über `Clock.Now()` (`SPEC-232` Nr. 5 f) | `TestUhrRuftAdapter`: Schleife über 2 Wege (`StartSpan`, `End`) × 3 Rückrufe in `Now()` (`CountError`, `Snapshot`, `Log`; 6 Fälle, Zählprüfung 2 und 3); Aufruf in einer Goroutine mit Frist von 2 s, Orakel „kehrt zurück“; Deadlock in jedem Lauf | `End` ruft `Now()` unter der Sperre des Stands | nebenläufig |

Zählung dieses Teils: 18 Zeilen, davon 4 `nebenläufig` (*a*, S15, S16, S17, S18) und 14 `andere` (*b*). Zeile 5 der Zusatztabelle „Pflicht-Attribut mit ungültigem Wert, dann `End`“ (im ungeschnittenen Plan bei P17) fängt S02, weil sie `StartSpan` und `End` braucht.

**Betriebsregeln für alle Rollen:**

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts (Entscheidung des Users vom 2026-10-21).
- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` (höchstens zwei Container je Linie zugleich) unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Die Image-Falle: das Image `odata-recorder:race` trägt `/src/src` eingebacken; ein Lauf, der Quellen hineinkopiert, leert den Pfad zuvor oder nutzt die Stufe `deps`. Ein Rot unter Last wird auf ruhigem Host wiederholt, bevor es als Befund gilt.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`: eine weitere Verletzung ist ein Befund des Reviewers); keine Hilfsdateien im Wurzelverzeichnis. Commit-Nachrichten gehen als Datei (`git commit -F`), nicht als Heredoc.
- Eine Hülle um ein Argument (ein Fake, der `ctx`, einen Strom oder einen Writer umhüllt, ein Test-Writer der Senke) darf keine Eigenschaft des Originals verdecken (Abbruch, Fehler, Blockieren, Zahl der Aufrufe; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`): die Tests dieses Slice nutzen `FakeSenke` aus `slice-telemetry-log` und `FestUhr` aus `supporttest` und fügen keine eigene Hülle hinzu, die Abbruch, Fehler oder die Zahl der Aufrufe verdeckt; eine Panik eines Fakes färbt den benannten Test.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Die Klassen der Tabelle sind generierbar (hier: jeder Span-Name und jedes Pflicht-Attribut von `SPEC-026` bis `SPEC-030`, jede Eimer-Grenze, jedes Paar Weg × Stelle des Canary): schon der Handoff führt den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben wird nur im Integrationsschritt gesetzt, nicht im Slice-Commit.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13); der Bericht allein ist keine Adresse.
- Der Implementer hält an einer Randform ohne Ort in der Spezifikation an und gibt sie an den Architect zurück (`AGENTS.md` §3.12); er entscheidet sie nicht im Code.
- Wert mit Escape-Folge oder Rohbytes: der Plan und die Berichte schreiben eine solche Folge nicht als Zeichen, sondern beschrieben (Werkzeug-Falle Backslash-u, `AGENTS.md` §3.18).

**Aufträge je Rolle** (Planner, beim Schnitt vom 2026-10-30):

- *Implementer:* S15 und S16 sind Wirkungs-Fänger für die Zusage „unter einer Sperre, ohne sie dazwischen freizugeben“; der Detektor von `make test-race` meldet ein Freigeben dazwischen nicht. Beide gelten nur als gefangen, wenn ihre Mutation in jedem Lauf rot wird: der Bericht nennt die Rate je Mutation als *n* von *n* bei 50 Prozessen mit `-test.count=1` unter `make test-race` und mit `-count=50`, die Zahl der Prozesse und das Messende, und der Implementer trägt sie mit der Tabelle in einen Absatz von §3 ein (`AGENTS.md` §3.10); erreicht ein Fänger nicht *n* von *n*, bekommt die Zusage einen schärferen Aufbau (mehr Goroutinen, kürzere Abschnitte) oder steht als akzeptiertes Negativ mit Grund und Folge im Plan, und ein Haltepunkt an einem Argument, das der Test übergibt (`ctx`, `Clock`, Senke), wird vorher gesucht. S17 und der Zähler der `ID` melden Detektor und Wirkung je Prozess. S18 ist nebenläufig und färbt sich durch Deadlock in jedem Lauf (Frist von 2 s); der Bericht nennt seine Rate mit Zahl der Prozesse. Die Canary-Schleife S13 läuft als Schleife über 4 Einbringungswege × 24 Schlüssel über alle 6 Ausgabewege (576 Prüfungen), nicht als Liste einzelner Fälle. Die Tests nutzen `FakeSenke` und `FestUhr` ohne Hülle, die Eigenschaften verdeckt. Die Zusatzzeit der Stufen `test` und `race` (vorher und nachher, Zahl der Läufe, Zustand des Hosts) steht in §7; die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) liegt beim User, nicht im Slice.
- *Reviewer:* löst jede Zeile der Tabelle und jede Zelle des Rasters mit einer Mutation auf, fährt S15 und S16 selbst mit Rate (*n* von *n* bei 50 Prozessen, beide Läufe), prüft den Canary gegen die Klasse (alle 96 Paare × 6 Ausgabewege) und die Fenster der Sperre (je Fenster eine Zeile: End zwischen Ring und Verteilung, Snapshot zwischen Verteilung und Ring); eine im Code entschiedene Randform ist ein MEDIUM (`AGENTS.md` §3.16).
- *Verifier:* prüft die Tabelle gegen `SPEC-230` Nr. 5 bis 9, `SPEC-231` Nr. 6 und `SPEC-232`, dass dieser Slice nichts aus `slice-telemetry-port` oder `slice-telemetry-log` ändert, dass das Binary unverändert blieb, und die Messpunkt-Zeile (Handoff und Ende, `SPEC-128` Nr. 9).
- *Planner (Closure):* setzt die Messpunkt-Zeile, die Auslastung und meldet dem Architect, liegt sie über 1,00 oder unter 0,60.

**Rollenfolge** (`.claude/agents/`, kein Self-Review): Architect (Vertrag, Tabelle, Raster, Block `groessenmarke`) → Planner (Schnitt, Marke beim Start, §4) → Implementer (Code, Mutationen selbst rot, Absatz zu Zusagen und Raten in §3, ein Handoff-Commit) → Reviewer (Report unter `docs/reviews/`) → Fix-Runden (Menge je Befund, `AGENTS.md` §3.17) → Verifier → Planner (Closure mit Lerneintrag; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorlage an den User) → Integrationsschritt (Merge, Register, Roadmap, Stempel, `make gates` im Hauptbaum).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-telemetry-port` und `slice-telemetry-log` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-port`: *Code-Abhängigkeit*. Typen, `support.Spans()`, `support.Eimer()`, die Regeln der Attribute, `Snapshot` mit `Durations` und `Spans`, `Dropped` und der wirkungslose `StartSpan`, den dieser Slice ersetzt.
- `slice-telemetry-log`: *Code-Abhängigkeit*. Die Span-Zeile (`SPEC-231` Nr. 6) nutzt den Zeilenschreiber, die Stufen mit `MinLevel`, die Sperre der Zeile und den Ausfall der Senke aus diesem Slice, und die Canary-Tests bauen auf `FakeSenke` auf. Ohne ihn gäbe es nichts, in das `End` schreibt; die Kette `P` → `L` → `S` ist die Reihenfolge des Codes, nicht eine Wahl.
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` aus `welle-benutzerhandbuch` liegen dann in `done/`; ohne sichtbare Änderung am Binary entfällt die Bedingung.

**Trigger der Welle:** Der Start-Trigger von [welle-betrieb](../welle-betrieb.md) §2 (`welle-mock` in `done/`, M3 erreicht) ist für diesen Slice wie für `slice-telemetry-port` gelockert (Entscheidung des Planners nach Auftrag des Users vom 2026-10-30, dort §4): der Slice braucht keinen Code von Mock oder Workspace und ändert `src/bootstrap` nicht. Die Lockerung gilt nur, solange der Slice ohne Wirkung am Binary bleibt; sonst wartet er auf die beiden Handbuch-Slices (Rückführung `in-progress` → `open`, unten).

**Größenmarke:** `basis`, `schaetzung` und `marke` setzt der Planner beim Start, vor dem ersten Code-Commit, in eigenem Commit (`SPEC-128`: Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Erwartet nach der Tabelle dieses Teils: *a* = 4 (S15, S16, S17, S18), *b* = 14, *f* = 0 (die Tests nutzen `FakeSenke` aus `slice-telemetry-log` und `FestUhr`), *p* = 350 (Teil des Architect, 1310 für alle drei Teile), also *S* = 85·4 + 80·14 + 200·0 + 350 = 1810 und *M* = 2500; die Schnittfrage (*S* ≥ 2223) stellt sich nicht. Der Planner rechnet beim Start neu, falls die Tabelle sich bis dahin änderte. Pfad: allein der Adapter; der Slice ändert `src/hexagon/ports/support` nicht, und eine Änderung dort wäre eine Abweichung vom Plan (§6), die den Pfad in den Block nimmt.

```groessenmarke
pfad: src/adapters/driven/telemetry
produktcode: 350
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (Messung nach `SPEC-128`).
- `in-progress` → `open` (blockiert): der Slice bringt einen Code, Schlüssel oder ein Kommando in den Katalog des Binary oder macht Logfelder oder die Span-Zeile am Binary sichtbar, und `slice-handbuch-geruest` oder `slice-handbuch-meldungskatalog` liegen nicht in `done/` (nicht erwartet); oder der Zeilenschreiber aus `slice-telemetry-log` nimmt die Span-Zeile nicht auf (§6, Adresse Architect).

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

- Die Wirkungs-Fänger S15 und S16 werden nur mit Wahrscheinlichkeit rot, weil unter der Sperre kein Aufruf eines übergebenen Arguments liegt, an dem ein Test anhalten könnte (`SPEC-232` Ungeprüft); der Detektor meldet ein Freigeben dazwischen nicht (`SPEC-122` Nr. 8 b) — **Ausgang:** weiter offen; Adresse: Implementer (Rate *n* von *n* bei 50 Prozessen mit Zahl und Messendem im Absatz von §3, sonst schärferer Aufbau oder akzeptiertes Negativ mit Grund und Folge: die Diagnose weicht um einen Span ab, das Ergebnis des Aufrufers bleibt unberührt).
- Die Span-Zeile nutzt den Zeilenschreiber aus `slice-telemetry-log`; nimmt er nur die Attribute von `Log` entgegen, aber nicht die Felder `span_id`, `parent_id` und `duration_ms`, wäre das eine Abweichung — **Ausgang:** weiter offen; Adresse: Architect (der Implementer hält an und gibt zurück, `AGENTS.md` §3.12).
- Spans und Zähler des Hybrid-Fallbacks haben keinen Nehmer in `welle-betrieb` (`slice-hybrid-service` schließt sie aus): ein Span ist eine Zeile in der Tabelle von spezifikation.md §5 und ein Eintrag in `support.Spans()`, ein Zähler eine Folge-Entscheidung — **Ausgang:** weiter offen; Adresse: Planner (Empfehlung des Architect: neuer Slice `slice-telemetry-einbau-hybrid` nach `slice-hybrid-service` und diesem Slice; der Planner legt ihn nicht an, bevor der User entschieden hat).
- Auftrag für `slice-lasttest`: der Abruf des Gesamtspans ist die Span-Zeile dieses Slice (`msg` gleich `odata.gateway.request`, `duration_ms` mit drei Nachkommastellen, `span_id`, `parent_id` für den Anteil von `odata.replay.match` und `odata.mock.query`) oder, im Prozess, `TelemetryReader` — **Ausgang:** weiter offen; Adresse: Planner (Risiko-Ausgang und Auftrag 2 im Plan des Nehmers nachziehen, `AGENTS.md` §3.13; in `open/` schreibt dieser Zweig nichts).
- Einbau-Slices (`slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`) starten nach diesem Slice, nicht nach `slice-telemetry-port`; ihr Start-Trigger nennt heute nur den ersten — **Ausgang:** weiter offen; Adresse: Planner (Abhängigkeiten der beiden Dateien ergänzen).
- Sperre über `Clock.Now()` (`SPEC-232` Nr. 5 f), Seite von `StartSpan` und `End` — **Ausgang:** entschieden; Fänger S18.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten ist gefüllt, jede Zelle trägt ihren Ort — **Ausgang:** entfallen; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect und steht dann hier.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Dieser Teil trägt die Zeilen der Spans; `Log`, `Handler`, `CountError` und `Snapshot` stehen in `slice-telemetry-log` und `slice-telemetry-port`. Die Nummern stehen in spezifikation.md §5.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | sensibler Wert | Abbruch (`ctx`) | Wiederholung |
|---|---|---|---|---|---|---|
| Span beginnen (`StartSpan`) | `SPEC-230` Nr. 5 (Pflicht-Attribut fehlt erst am Ende: `pflicht`) | `SPEC-230` Nr. 3 und 5 (leerer Name: `spanname`; leerer Wert: `wert`) | `SPEC-230` Nr. 5 (fremder Name: wirkungsloser Span, `spanname`; fremder Schlüssel: `schluessel`) | `SPEC-232` Nr. 2 | `SPEC-230` Nr. 1 und 6 (der Rückgabe-`ctx` leitet ab) | `SPEC-230` Nr. 6 (jeder Aufruf ein neuer Span mit neuer `ID`) |
| Span setzen und beenden (`Set`, `End`) | `SPEC-230` Nr. 5 (`pflicht`) | `SPEC-230` Nr. 3 (`wert`) | `SPEC-230` Nr. 4 und 5 (`schluessel`, `wert`; ungültiger Wert lässt den früheren stehen) | `SPEC-232` Nr. 2 | `SPEC-230` Nr. 7 (`End` zeichnet auch bei beendetem `ctx` auf) | `SPEC-230` Nr. 7 (zweites `End`, `Set` nach `End`: `zustand`) |
| Lesen der Spans und Verteilungen (`Snapshot`, Teil `Durations` und `Spans`) | gilt nicht: kein Argument | `SPEC-230` Nr. 9 (frischer Adapter: `Spans` leer, alle Verteilungen mit 0) | gilt nicht: kein Argument | gilt nicht: der Stand enthält nur, was die Formen zuließen (`SPEC-232` Nr. 2) | gilt nicht: kein `ctx` (`SPEC-230` Nr. 1) | `SPEC-230` Nr. 9 (ändert und zählt nichts; Kopie je Aufruf), `SPEC-232` Nr. 5 b (ein Stand) |

**Zusatztabelle: Verstöße am selben Schlüssel** (`AGENTS.md` §3.16; `SPEC-230` Nr. 4, Rangfolge Schlüssel vor Wert, genau ein Verwurf je Attribut): die Zeilen 1 bis 4 und 6 fängt `TestVerwurfRangfolge` (P17) in `slice-telemetry-port`; die Zeile 5 fängt dieser Slice:

| Paar | Eingabe | Stand danach | Fänger |
|---|---|---|---|
| Pflicht-Attribut mit ungültigem Wert, dann `End` | `StartSpan` mit `recorder.session_id` und Leerraum, dann `End` | Span aufgezeichnet ohne das Attribut; `wert` 1, `pflicht` 1 | S02 |

**Zusatztabelle: Modus** (`AGENTS.md` §3.16; die Zusage hängt am Modus):

| Modus | Zusage | Fänger |
|---|---|---|
| `MinLevel` bis `INFO` | Aufzeichnung und Span-Zeile (`SPEC-231` Nr. 6) | S13 (Modus 1 von 3) |
| `MinLevel` über `INFO` | Aufzeichnung, keine Span-Zeile, kein Verwurf | S13 (Modus 2 von 3), L04 in `slice-telemetry-log` |
| Senke defekt | Aufzeichnung, keine Zeile, `senke` 1 (`SPEC-232` Nr. 3) | S13 (Modus 3 von 3), L14 in `slice-telemetry-log` |

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die Messpunkt-Zeile von `make groessenmarke` (Handoff und Ende, von Hand gemessener Produktcode, Auslastung), die Raten der Wirkungs-Fänger S15 und S16 (*n* von *n* bei 50 Prozessen mit `-test.count=1` unter `make test-race` und mit `-count=50`, Messendes), die Zusatzzeit der Stufen `test` und `race`, und dass das Binary unverändert blieb.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, Zähler aus den Dateien unter `evidence/`; der Planner zieht Treffer und Zähler beim Start nach). Einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für Pfade, die kein Lauf ausführt, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für S15 und S16 (Rate mit Zahl der Läufe, `AGENTS.md` §3.10), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (7×, `AGENTS.md` §3.10) für die Sperre des Stands, `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16) für „kein Credential in keinem Span“, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `implement-slice` Schritt 14) für 5 Namen, 11 Pflicht-Attribute und 12 Eimer, `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für die Hüllen der Tests, dazu `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (offen; dieser Zweig setzt keinen Zähler).

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF.

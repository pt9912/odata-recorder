# Slice slice-telemetry-einbau-replay: Match-Span `odata.replay.match` im Replay-Dienst und ein Adapter der Telemetrie je Prozess in `replay` und `hybrid` (Teil 2 des Schnitts von `slice-telemetry-einbau-gateway`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-296` (Match-Span; Nr. 1 bis 9), neu vom Architect vor dem Code · `SPEC-028` (spezifikation.md §5) · `LH-QA-01.a` (der Match-Span weist den Anteil der Auswahl am Gesamtspan aus) · `SPEC-135` Nr. 1, 15, 16 und 18 (Konfig, Fehlercodes, Aufrufstellen des Replay-Dienstes) · `SPEC-139` Nr. 3, 11 und 13 c (ein Adapter des Prozesses in jedem Modus) · `SPEC-230` Nr. 8 · `SPEC-109` · `ARC-002`, `ARC-005`, `ARC-012`, `ARC-021` (der Architect ordnet sie beim Schnitt keinem Teil zu).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md), Vorschlag des Architect vom 2026-11-21). **Datum:** 2026-11-22.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Replay-Dienst setzt `odata.replay.match` (`SPEC-028`, `SPEC-296`) mit den Pflicht-Attributen `recorder.match_strategy` und `recorder.match_result`: der Span beginnt nach der Zuordnung und vor `Sessions.Get`, ein `Set` mit beiden Attributen folgt der Rückkehr der Transaktion, `End` steht auf jedem Rückweg nach dem Beginn genau einmal, auch bei einem Kontextfehler; das Ergebnis folgt dem Code der Auswahl, wo die Auswahl nicht lief, stehen die Platzhalter `nicht_gelesen` und `nicht_gewaehlt`. Der Dienst ruft `CountError` nie (`SPEC-296` Nr. 5; den Zähler je Fehlercode führt der HTTP-Adapter). Der Adapter des Prozesses ersetzt den zweiten Adapter des Replay-Dienstes: in `replay` und `hybrid` gibt es einen Adapter je Prozess (`SPEC-139` Nr. 3 und 13 c). Kein Query-Literal und kein Body wird Attribut. Der Teil setzt [slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md) (Teil 1) voraus: den Gesamtspan im HTTP-Adapter, dessen `ctx` den Match-Span zum Kind macht, und die Übergabe des Adapters des Prozesses an `httpodata.New`; er ändert beides nicht.

**Übernimmt (Sendung von `slice-telemetry-port`, Teil Replay; Geber: `slice-telemetry-port`; Schnitt des Planners vom 2026-11-22):** der Einbau in den Pfad `replay`; der Port selbst und der Fake sind dort geliefert. Der Einbau in den HTTP-Adapter liegt in Teil 1, der in den Pfad `record` in [slice-telemetry-einbau-record](slice-telemetry-einbau-record.md) (Teil 3).

**Übernimmt (Sendung von [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) (Teil 1), Verdrahtung des Adapters, Replay-Dienst; Geber: `slice-telemetry-einbau-mock`; Eintrag des Planners vom 2026-11-21 bei der Closure von Teil 1 jenes Schnitts, beim Schnitt vom 2026-11-22 geteilt; §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung):** der Adapter des Prozesses (`SPEC-139` Nr. 13: ein `telemetry.New` je Prozess in `verdrahte` am Haken `NeueTelemetrie`, `Writer` der Fehlerstrom von `Start`, `Clock` die eine Uhr, `MinLevel` `INFO`) ersetzt den Adapter des Replay-Dienstes (`SPEC-139` Nr. 3, `telemetry.New` mit `io.Discard`, in `replay` und `hybrid` der zweite Adapter); `SPEC-139` Nr. 3 und Nr. 13 c sind mit dem Vertrag nachgezogen (Architect, 2026-11-21). Die Tests, die die Zahl der Adapter je Modus festhalten (`m3_wurzel_test.go`, `m4_wurzel_test.go`, `wurzel_test.go` über `alleTel`), ändern sich dabei von zwei auf einen Adapter in `replay` und `hybrid` (Zeile W10). Der Teil des Adapters für die Dienste `record`, `session` und `maintenance` liegt in Teil 3.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Gesamtspan, der Zähler je Fehlercode, `Konfig.Telemetrie` und `Konfig.Modus` des HTTP-Adapters und die Übergabe im Root — [slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md) (Teil 1, liegt bei diesem Start in `done/`); der Match-Span liest den Gesamtspan nur über das `ctx` (`SPEC-296` Nr. 3 e).
- Der Capture-Span, die Einträge der Dienste `record`, `session` und `maintenance`, der Wortlaut von `SPEC-121` Nr. 12 und der Eintrag `audit_fehlgeschlagen` — [slice-telemetry-einbau-record](slice-telemetry-einbau-record.md) (Teil 3).
- Port und Adapter `Telemetry`; Zeile, Stufen, Senke, `Handler`, Spans, Verteilung und Ring — `slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans`; alle liegen in `done/`, dieser Slice ändert sie nicht.
- Spans von Mock-Dienst und Workspace-Mutation — `slice-telemetry-einbau-mock` und `slice-telemetry-einbau-mock-mutation`.
- Spans und Zähler des Hybrid-Fallbacks: der Versuch des Hybrid-Dienstes ruft `Replay` und trägt den Match-Span (Zeile W10, `kein_treffer`); der Span des Fallbacks und der Zähler der verworfenen Versuche haben keinen Nehmer (`SPEC-296` Nr. 7 e, `SPEC-290` Nr. 7 g; Sendung ohne Nehmer, §6 von Teil 1, Adresse: Planner).
- Der Lasttest und seine Messung — `slice-lasttest`; dieser Slice liefert den Match-Span, den der Lasttest liest.
- Das Handbuch-Kapitel über die Logzeilen des Prozesses — `slice-handbuch-telemetrie`; es braucht Teil 1 und Teil 3, nicht diesen Teil (die Zeile des Match-Spans ist ohne importierte Austausche kein Beispiel); der Slice schreibt kein Kapitel und führt keinen Fehlercode, Schlüssel oder Kommando ein (siehe DoD).

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

- [ ] Match-Span im Replay-Dienst: `odata.replay.match` beginnt nach der Zuordnung und vor `Sessions.Get`, trägt Strategie und Ergebnis in einem `Set` nach der Rückkehr der Transaktion (Platzhalter `nicht_gelesen` und `nicht_gewaehlt`, wo die Auswahl nicht lief) und endet auf jedem Rückweg genau einmal, auch bei einem Kontextfehler; kein Aufruf von `Telemetrie` in `f`, der Dienst ruft `CountError` nie, das `ctx` des Aufrufs geht unverändert an `StartSpan` und an jeden Port (`SPEC-028`, `SPEC-296`, W1 bis W9, W11).
- [ ] Ein Adapter je Prozess und Test durch den Composition Root: der Replay-Dienst erhält in `replay` und `hybrid` den Adapter des Prozesses, der zweite Adapter mit `io.Discard` entfällt, die Zahl der Adapter ist in jedem Modus 1; die Zeile `odata.replay.match` steht in der Senke als Kind des Gesamtspans, im Modus `hybrid` trägt der Versuch `kein_treffer` (`SPEC-139` Nr. 3 und 13 c nachgezogen, Sendung von `slice-telemetry-einbau-mock`; W10).
- [ ] Fänger: kein Wert aus dem Request erreicht ein Attribut oder die Senke (Canary aus Bytewerten, W12), und gleichzeitige Aufrufe teilen weder Span noch Ergebnis (W13).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: der Slice führt keinen Fehlercode, keinen Schlüssel und kein Kommando ein; die Logzeilen des laufenden Prozesses beschreibt das Kapitel von `slice-handbuch-telemetrie` (Entscheidung des Users vom 2026-11-21); das Kataloge-Gate ist grün, und §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando des Binary dazukam.
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
| `src/hexagon/services/replay/` | update | Match-Span (`StartSpan` nach der Zuordnung, ein `Set` mit beiden Attributen nach der Transaktion, `End` auf jedem Rückweg), Zuordnung des Ergebnisses aus dem Code der Auswahl (`SPEC-296`); `Konfig` bleibt |
| `src/bootstrap/` | update | der Adapter des Prozesses statt eines zweiten für den Replay-Dienst (`baueReplay`, `SPEC-139` Nr. 13 c); die Übergabe an `httpodata.New` liefert Teil 1 |
| `spec/spezifikation.md` | geliefert (Architect, auf `main`) | `SPEC-296` und Nachzug `SPEC-135`, `SPEC-139`, `SPEC-230`; dieser Teil schreibt sie nicht |
| Tests | update, neu | Fakes der Ports; Journal-Fake der Telemetrie im Paket `replay`; Tests durch den Composition Root mit dem echten Adapter; angepasst: die Zahl der Adapter je Modus von zwei auf einen in `replay` und `hybrid` (`m3_wurzel_test.go`, `m4_wurzel_test.go`, `wurzel_test.go` über `alleTel`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-21 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-296`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidungen des Architect** (2026-11-21; die Aufträge 1 bis 10, je mit Ort in `spec/spezifikation.md`) stehen in §3 von [slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md); für diesen Teil gelten Nr. 2 (Attribute `recorder.match_strategy` mit fünf Werten und dem Platzhalter `nicht_gelesen`, `recorder.match_result` mit sechs Werten; 15 von 20 Paaren sind erreichbar), Nr. 3 (Eltern-Kind: der Match-Span beginnt nach dem Gesamtspan und endet vor ihm; die Dienste geben ihr `ctx` unverändert an `StartSpan` und verwerfen das gelieferte), Nr. 5 (kein Dienst zählt), Nr. 6 (Aufrufstellen W1 bis W4: `StartSpan`, `Set`, `End`, `Log`; Ports mit `ctx` hinter dem Beginn: Replay 8, S1, S3 bis S8, S10) und Nr. 9 (ein Adapter je Prozess, Zeile W10).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-21, Zeilen W1 bis W13 des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). `Nr.` ist die Kennung des ungeschnittenen Plans: **W** gehört zu diesem Teil, **G** zu Teil 1, **C** und **E** zu Teil 3. *Journal-Fake*: `FakeTelemetry`, dessen Aufrufe (und die der Fakes der Ports) in ein gemeinsames Protokoll schreiben, mit Buch je Span (Attribute, Zahl der `End`). Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| W1 | (`SPEC-296` Nr. 2) Kein Span und kein Aufruf von `Telemetrie`, was an den Schritten (a) bis (c) endet | Menge M-OhneSpan-Replay: Schritt (a) 4 (`ctx` nil, leere Methode, `Laenge` −2, `Body` nil bei `Laenge` 1), (b) 3 (kein Dienst, leerer Pfad, Pfad ohne `/`), (c) 12 (3 Header × fehlt, leere Wertliste, zwei Werte gleich 9, dazu die Form der Replay-Session-ID: 129 Bytes, Leerzeichen, Byte ab 0x80 gleich 3) gleich 19; `FakeTelemetry.Aufrufe()` leer; Zählprüfung 19 | `StartSpan` am Anfang von `Replay`; `StartSpan` vor der Zuordnung; Span auch für `ODR-E5005` | andere |
| W2 | (Nr. 3 a, 6) Beginn nach der Zuordnung und vor `Sessions.Get`; Folge `StartSpan`, Ports, `Set`, `Log`, `End` | Journal-Fake (Ports und `Telemetrie` im selben Protokoll): Folge je Fall als Literal: Treffer mit `sequential`; Treffer mit `KeyRedacted` (mit `Log`); `first` ohne `Put`; Fehler in `f` (`Set` nach der Rückkehr der Transaktion) gleich 4 | `StartSpan` nach `Sessions.Get`; `StartSpan` vor der Zuordnung; `Set` vor der Transaktion; `End` vor `Log` | andere |
| W3 | (Nr. 3 b, 6) Genau ein `End` auf jedem Rückweg nach dem Beginn, mit Code, mit Kontextfehler und mit Antwort; `Set` genau dann, wenn S4 gerufen wurde | Menge M-Rückwege-Replay: S1 bis S8 mit Fehler mit Code 8, mit Kontextfehler 8 (S2 über den Body), die elf Entscheidungen des Dienstes ohne Port (`SPEC-296` Nr. 6) 11, die Rückkehr mit Antwort 1, dazu das Paar Kontextfehler und Fehler der Tabelle (`errors.Join`) an den 8 Stellen gleich 36 Fälle; je Fall `StartSpan` 1, `End` 1, Ergebnis unverändert (Code oder Kontextfehler ohne Code); Zählprüfung 8, 8, 11 als Literal (Handkopie von `SPEC-135` Nr. 18 und `SPEC-296` Nr. 6, akzeptiertes Negativ 4) | `End` nur im Erfolgsweg; kein `End` bei Kontextfehler; kein `End` bei `ODR-E7009`; `Set` auch ohne S4 | andere |
| W4 | (Nr. 4 a bis c) Strategie und Ergebnis am Ende gleich dem Orakel für jedes erreichbare Paar | Menge M-Paare: 15 Szenarien (`sequential` 4, `strict` 5, `first` 3, `last` 3), Orakel Literal (Strategie, Ergebnis, Code oder Antwort); die 5 Paare, die es nicht gibt, sind kein Szenario (Zählprüfung 15 von 20) | `kein_treffer` und `erschoepft` vertauscht; `anderer_key` immer `kein_treffer`; Ergebnis nach der Prüfung des Zustands des Bodys gesetzt | andere |
| W5 | (Nr. 4 a, b, d) Platzhalter `nicht_gelesen` und `nicht_gewaehlt`, wo die Auswahl nicht lief; `Set` nur, wo S4 gerufen wurde | Menge M-Platzhalter: S1 4 Codes, Body und Key 3, S3 2, S4 vor `f` 1, S5 3, S6 1, Kontextfehler an S1 bis S6 6 gleich 20; Orakel Literal (beide Platzhalter; bei S4 bis S6 ein `Set` mit den Platzhaltern, vor S4 keins) | Platzhalter leer; Platzhalter `unbekannt`; Strategie der Konfig vor S6 gesetzt; `Set` auch vor S4 | andere |
| W6 | (Nr. 4 a, 7 d) Die Strategie des bestehenden Cursors gilt im Span, sonst die der Konfig | Menge M-Cursor: 4 Strategien am Cursor × eine abweichende der Konfig (4); fehlender Cursor × 4 Strategien der Konfig (4) gleich 8; Orakel Literal | Strategie der Konfig immer; Strategie des Cursors nur bei `sequential` | andere |
| W7 | (Nr. 4 b) Das Ergebnis bleibt `treffer`, wenn nach der Auswahl der Aufruf scheitert | Menge M-NachAuswahl: Zustand des Bodys `discarded` und ein unbekannter 2; `ErrBlobFehlt` und `ErrBlobBeschaedigt` an S7 2; S8 mit `ErrNichtGefunden`, `ErrFremderService` und anderem Fehler 3; Bestätigung der Transaktion 1; Kontextfehler an S7 und S8 2 gleich 10; Orakel Literal (`treffer`, Strategie, Ergebnis des Aufrufs unverändert) | Ergebnis nach einem Fehler zurück auf `nicht_gewaehlt`; Ergebnis erst nach S8 gesetzt | andere |
| W8 | (Nr. 3 c) `Telemetrie` wird in `f` nie gerufen; `Set` folgt der Rückkehr der Transaktion | Journal-Fake mit Fake der Transaktion: zwischen Beginn und Rückkehr von `WithWorkspaceTransaction` kein Aufruf von `Telemetrie`; 4 Strategien × {Erfolg, Fehler in `f`} gleich 8 Fälle | `Set` nach `waehle` in `f`; `StartSpan` in `f` | andere |
| W9 | (Nr. 3 d, 6) `StartSpan` und jede Stelle mit `ctx` erhalten das `ctx` des Aufrufs unverändert, nie das des Spans | Schleife über die 8 Stellen S1, S3 bis S8, S10 und `StartSpan`: der Fake hält jedes `ctx`, `StartSpan` liefert ein abgeleitetes mit eigenem Wert; Identität; Zählprüfung 8 als Literal (`SPEC-135` Nr. 18: „8 ein `ctx`“, akzeptiertes Negativ 4) | `ctx, span := StartSpan(…)`; das `ctx` des Spans nur an `Log`; nur an `Blobs.Open` | andere |
| W10 | (`SPEC-139` Nr. 13 c, `SPEC-296` Nr. 3 e, 7 e) Der Replay-Dienst erhält den einen Adapter des Prozesses; `odata.replay.match` steht in dessen Senke, als Kind des Gesamtspans und vor ihm; im Modus `hybrid` trägt der Versuch `kein_treffer` | Test durch den Composition Root (`BuildMitSenke`, feste Uhr): 2 Modi (`replay`, `hybrid`) × `GET` mit den drei Headern; Zeile `odata.replay.match` als Literal, `parent_id` gleich `span_id` der Zeile des Gesamtspans, die nach ihr steht, `Dropped` alle 0; `hybrid` mit unbekanntem Key und Upstream-Fake: Zeile `kein_treffer`, dann Fallback; die Zahl der Adapter (`alleTel`) ist 1 in allen 4 Modi (angepasst: bisher 2 in `replay` und `hybrid`) | zweiter Adapter mit `io.Discard` für den Replay-Dienst (Zahl 2, Zeile fehlt); eine andere `Clock` | andere |
| W11 | (Nr. 1, 5, 6) Der Dienst ruft `CountError` nie und `Log` nur wie bisher; der Quelltext nennt `StartSpan`, `Set`, `End` und `Log` je einmal | Schleife über die 13 Codes von `SPEC-135` Nr. 15, je ein Szenario, und 8 Kontextfehler: `CountError` 0; `go/parser` über das Paket ohne `_test.go`: `StartSpan` 1, `Set` 1, `End` 1, `Log` 1, kein `CountError`; Zählprüfung 13 | `CountError(code)` in `abbilden` oder `mitCode`; zweiter `Set`; `Log` im Weg des Fehlers | andere |
| W12 | (Nr. 8) Kein Wert aus dem Request erreicht ein Attribut oder die Senke | Menge M-Stellen-Replay: 8 Stellen (Pfad, Name und Wert der Query, Name und Wert eines Headers, Body, `X-Recorder-Replay-Session`, Text der Ursache eines Fehlers eines Ports) × 2 Arten (Treffer, Fehler) gleich 16; Canary aus Bytewerten in der Form Token, geprüft an jedem Attribut des Fakes und an der Senke des echten Adapters; Zählprüfung 8 | je Stelle der Wert in `recorder.match_result` (8 Mutationen); `err.Error()` als Attribut | andere |
| W13 | (Nr. 9 a) *n* gleichzeitige `Replay` ergeben *n* `StartSpan` und *n* `End`, jeder Span mit Strategie und Ergebnis seines Aufrufs | Wirkungs-Fänger: 16 Goroutinen × 4 Strategien × 2 Ergebnisse (Treffer, `kein_treffer`) in 20 Läufen; der Fake führt Buch je Span; eine Schranke an `Blobs.Open` hält alle 16 Aufrufe offen; Rate je Prozess | Ergebnis in einem Feld des `Replayer` geteilt; `End` nur jeden zweiten Aufruf; die Strategie in einer geteilten Variable | nebenläufig |

Zeilen: W1 bis W13 gleich 13 (*a* = 1 (W13), *b* = 12); die akzeptierten Negative stehen unter der Tabelle und zählen nicht als Zeile (`SPEC-128` Nr. 3 e gilt für Zeilen).

**Verteilung beim Schnitt** (Planner, 2026-11-22): Die Zeilen sind im Wortlaut des Architect übernommen, keine geändert. Die Zeile W10 prüft den Gesamtspan als Elternspan durch den Composition Root und setzt Teil 1 voraus. Die Zählprüfung der Adapter (`alleTel`: 1 in allen vier Modi) ist Gegenstand von W10; Teil 3 hängt den Adapter für die Dienste `record`, `session` und `maintenance` an denselben einen Adapter und ändert die Zahl nicht.

**Mengen mit Zählprüfung** (Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15):

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| M-OhneSpan-Replay (W1) | (a) 4, (b) 3, (c) 12 gleich 19 | kein Aufruf | 19 |
| M-Rückwege-Replay (W3) | S1 bis S8 mit Code 8, mit Kontextfehler 8, Entscheidungen 11, Antwort 1, Paar 8 gleich 36 | `End` 1, Ergebnis | 8, 8, 11 = `SPEC-135` Nr. 18, `SPEC-296` Nr. 6 |
| M-Paare (W4) | `sequential` 4, `strict` 5, `first` 3, `last` 3 gleich 15 von 20 | Strategie, Ergebnis | 15 und 20 = `SPEC-296` Nr. 4 c |
| M-Platzhalter (W5) | S1 4, Key 3, S3 2, S4 1, S5 3, S6 1, Kontext 6 gleich 20 | Platzhalter, `Set` ja oder nein | 20 |
| M-Cursor (W6) | 4 + 4 gleich 8 | Strategie | 8 |
| M-NachAuswahl (W7) | 2 + 2 + 3 + 1 + 2 gleich 10 | `treffer` | 10 |
| M-ctx-Replay (W9) | S1, S3 bis S8, S10 gleich 8, dazu `StartSpan` | Identität des `ctx` | 8 = `SPEC-135` Nr. 18 („8 ein `ctx`“) |
| M-Stellen-Replay (W12) | 8 Stellen × 2 Arten gleich 16 | kein Canary | 8 = `SPEC-296` Nr. 8 |
| M-Gleichzeitig (W13) | 16 Goroutinen × 8 (Strategie und Ergebnis), 20 Läufe | Buch je Span | 16 |

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptiertes Negativ):** Pfad, Query und Body des Requests (Prüfgrenze: ein kurzer Canary als Wert; Grund: der Dienst setzt nie einen Wert des Requests, die Länge ändert den Weg nicht; Folge: ein Wert von 1 MiB wird nicht eigens geprüft, `SPEC-119` begrenzt den Body).

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; im Wortlaut des Architect, die Zeilen der Spans von Teil 1 und Teil 3 stehen dort):

| Paar oder Modus | Erwartung | Fänger |
|---|---|---|
| Gesamtspan und Match-Span | `parent_id` gleich `span_id`; die Zeile des Match-Spans steht vor der des Gesamtspans | W10 |
| Match `treffer` und Fehler nach der Auswahl (`ODR-E3001`, `ODR-E9001`, Fehler von `Put`) | Span `treffer`; Status und Zähler im Gesamtspan und beim Adapter | W7, G9 |
| Cursor mit anderer Strategie als die Konfig | Strategie des Cursors im Span | W6 |
| Modus `hybrid`: Replay-Versuch mit `ODR-E5001`, dann Fallback | Match-Span `kein_treffer`; der Fehler zählt nicht, die Antwort des Fallbacks bestimmt Status und Zähler | W10 |

G9 ist die Zeile des Zählers in Teil 1 ([slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md)).

**Akzeptierte Negative dieses Vertrags** (Grund, Folge und Maß): (1) die Dauer des Match-Spans (`SPEC-296` *Ungeprüft*): eine Zeitmessung im Test schwankt mit der Last (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`); die Dauer misst `slice-lasttest`. Folge: Arbeit, die ein Eingriff vor den Beginn oder hinter das Ende legt, ohne die Folge der Aufrufe zu ändern, bleibt ungefangen; Maß: W2 bindet 1 Folge, nicht die Zeit. (2) Ein Ausfall der Senke ändert Antwort und Ergebnis nicht: wie das Negativ (2) in §3 von Teil 1 (keine Operation von `Telemetry` liefert einen Fehler, 0 von 5 Operationen); der Fänger liegt am Adapter der Telemetrie (`SPEC-232` Nr. 3). (3) Die Zahlen der Mengen (13 Codes von `SPEC-135` Nr. 15, 8 Stellen mit `ctx` und 8 Fehlerrückgaben von `SPEC-135` Nr. 18, die elf Entscheidungen, 15 von 20 Paaren) gegen die Tabellen der Spezifikation prüft kein Test, weil die Stufe `source` des Images `spec/` nicht trägt; Folge: ändert jemand die Tabelle von `SPEC-135` Nr. 18 oder die Menge von `SPEC-296` Nr. 6 ohne den Test, sieht es erst ein Review (der Architect hat am 2026-11-21 von Hand gezählt: 10 Stellen von `SPEC-135` Nr. 18 mit 8 Fehlerrückgaben und 8 `ctx`, 13 Codes). (4) Ein Typed-nil in `Telemetrie` (`SPEC-296` *Ungeprüft*): `New` des Replay-Dienstes prüft nil; Folge: eine Panik beim ersten Span, im Betrieb unerreichbar (der Root übergibt den Adapter). (5) Der Span des Hybrid-Fallbacks und der Zähler der verworfenen Versuche (`SPEC-296` Nr. 7 e, `SPEC-290` Nr. 7 g): kein Nehmer, siehe §6.

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Strategie, jedes Ergebnis, jede Aufrufstelle des Ports, jeder Fehlercode des Replay-Dienstes), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md) (Teil 1) liegt in `done/`; mit ihm liegen die Bedingungen von Teil 1 in `done/` (`slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans`, `slice-http-adapter`, `slice-http-lebenszyklus`, `slice-bootstrap-m2`, `slice-telemetry-einbau-mock`); dazu `slice-replay-service` in `done/`. Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-296`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-einbau-gateway` (Teil 1): *Code-Abhängigkeit*. Der Gesamtspan, dessen `ctx` den Match-Span zum Kind macht (W10 prüft es durch den Root), und die Übergabe des Adapters des Prozesses an `httpodata.New`; beide Teile ändern `src/bootstrap/wurzel.go` und laufen nacheinander (`welle-betrieb` §5).
- `slice-replay-service` (`welle-record-replay`): der Ort des Match-Spans (`src/hexagon/services/replay`); der Dienst liegt in `done/`.
- `slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler` und `slice-telemetry-spans`: Port, Fake und Adapter, die dieser Teil ruft; über Teil 1 transitiv erfüllt.
- `slice-bootstrap-m2` (`welle-record-replay`): der Composition Root, der Adapter und Dienste je Modus verdrahtet; ohne ihn gäbe es keinen Test durch den Root (W10).

**Reihenfolge und Konflikte** (Planner, 2026-11-22): Die drei Teile des Schnitts ändern `src/bootstrap/wurzel.go` und laufen nacheinander: Teil 1, dieser Teil, Teil 3 ([slice-telemetry-einbau-record](slice-telemetry-einbau-record.md)); `slice-telemetry-einbau-record` startet erst, wenn dieser Teil in `done/` liegt. Wer `main` nachzieht, misst die fremden Zeilen unter `src/bootstrap/` mit (`SPEC-128` Nr. 10 g): er misst vor dem Merge, nennt den Messpunkt und trägt danach, wenn nötig, `ueberschreitung` mit Grund ein (Nr. 5 d), keine Rückführung.

**Größenmarke:** noch nicht gesetzt; der Teil des Architect steht im Block unten (Zeilen W1 bis W13, `SPEC-128` Nr. 3 k), `basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke`. Vorgerechnet beim Schnitt (Planner, 2026-11-22, `SPEC-128` Nr. 1, nach der Tabelle nachgezählt): 13 Zeilen, *a* = 1 (W13), *b* = 12, *f* = 1 (Journal-Fake der Telemetrie im Paket `replay`), *p* = 45 (Beginn, Variablen für Strategie und Ergebnis in `f`, die Abbildung des Codes auf das Ergebnis, `Set`, `End`, der Adapter im Root statt des zweiten): *S* = 85·1 + 80·12 + 200·1 + 45 = 1290, 135·1290 = 174150, *M* = 100 · ⌈17,415⌉ = 1800; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223). Der Teil liegt 933 unter der Schnittgrenze: jede weitere Zeile der Tabelle kostet 80, zwölf weitere lösen die Schnittfrage neu aus (*S* = 2250; elf geben 2170). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/hexagon/services/replay/` und `src/bootstrap/`, Tests eingeschlossen. Den Block setzt der Planner beim Start (Prosa-Marke und Block ersetzen einander nicht, der Block steht allein).

```groessenmarke
pfad: src/hexagon/services/replay/
pfad: src/bootstrap/
produktcode: 45
fakes: 1
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (`AGENTS.md` §3.14); ein weiterer Schnitt ist nicht vorgesehen, die Schätzung liegt weit unter der Schnittgrenze.
- `in-progress` → `open` (blockiert): das Ergebnis des Match-Spans lässt sich nicht aus dem Code der Auswahl ableiten, ohne dass `f` einen Typ des Adapters kennt (`SPEC-296` Nr. 3 c und 4 entscheiden das Gegenteil: `f` ruft `Telemetrie` nie, das `Set` folgt der Transaktion).

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

- Der Match-Span darf das Ergebnis des Aufrufs nicht ändern und muss bei Abbruch (`ctx`) an jedem Port genau einmal enden; ein Zweig ohne Lauf unter dem Detektor (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`) — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-296` Nr. 3 b, c, f): `End` wirkt bei beendetem `ctx` und endet genau einmal, `f` ruft `Telemetrie` nie; der Abbruch trägt Fänger an jedem Port (W3, 36 Fälle) und den Wirkungs-Fänger W13 unter `make test-race`; der Ausfall der Senke ist akzeptiertes Negativ (2) mit Grund, Folge und Maß.
- Die Zahl der Adapter der Telemetrie je Modus ändert sich in `replay` und `hybrid` von zwei auf einen; die Tests, die sie festhalten (`m3_wurzel_test.go`, `m4_wurzel_test.go`, `wurzel_test.go` über `alleTel`), ziehen um — **Ausgang:** weiter offen; Adresse: Teil 2 (Zeile W10).
- Der Match-Span ist nur dann Kind des Gesamtspans, wenn das `ctx` des Aufrufs ihn trägt, und nur dann Teil derselben Senke, wenn der Replay-Dienst den Adapter des Prozesses erhält — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-296` Nr. 3 d und e, `SPEC-139` Nr. 13 c): `StartSpan` erhält das `ctx` des Aufrufs, kein Port erhält das gelieferte `ctx` (W9); der Test durch den Root mit dem echten Adapter ist W10; er setzt Teil 1 voraus (§4).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-21): das Raster unten trägt in jeder Zelle einen Ort oder `gilt nicht` mit Grund; eine Randform, die der Implementer im Code findet, geht an den Architect (`AGENTS.md` §3.12).
- Spans und Zähler des Hybrid-Fallbacks (Sendung ohne Nehmer, `SPEC-296` Nr. 7 e): der Versuch des Hybrid-Dienstes trägt den Match-Span (W10), der Fallback trägt keinen — **Ausgang:** weiter offen; Adresse: Planner (Frage an den User, geführt in §6 von [slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md)).
- Schnitt und Größenmarke: der Plan ist am 2026-11-22 vor dem Code aus `slice-telemetry-einbau-gateway` geschnitten (ungeschnitten *S* = 4605 über der Schnittgrenze; dieser Teil *S* = 1290, Teil 1 *S* = 1875, Teil 3 *S* = 1440) — **Ausgang:** weiter offen; Adresse: Planner (setzt beim Start den Block mit `basis`, `schaetzung` und `marke`, `AGENTS.md` §3.14).
- Reihenfolge der drei Teile: alle ändern `src/bootstrap/wurzel.go` — **Ausgang:** entschieden (Planner, 2026-11-22): sie laufen nacheinander (1, 2, 3); der Start-Trigger dieses Teils nennt Teil 1, der von Teil 3 nennt diesen Teil (§4, `welle-betrieb` §5).
- Handbuch: der Match-Span ist eine weitere Logzeile des Binary, im Beispiel-Format ohne importierte Austausche nicht erreichbar; ein Kapitel zu Logzeilen ist nur als gebundenes Beispiel zulässig (`SPEC-200` Nr. 1) — **Ausgang:** entschieden (Planner, 2026-11-22): dieser Slice schreibt kein Kapitel, `slice-handbuch-telemetrie` startet nach Teil 1 und Teil 3 (Entscheidung des Users vom 2026-11-21), §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando dazukam.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-21, Zeile des Match-Spans des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund; „Nr.“ ohne Nennung heißt `SPEC-296`):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch (Client, `ctx`, `Drain`) | Wiederholung | Herunterfahren |
|---|---|---|---|---|---|---|
| Match-Span (`SPEC-296`; Strategie, Ergebnis) | Nr. 7 a (ein fehlender Header der Zuordnung ist `ODR-E5003` vor dem Beginn, Nr. 2); Nr. 4 a (ein fehlender Cursor: Strategie der Konfig) | Nr. 7 a (leerer Header: `ODR-E5003`; leerer Body ändert den Span nicht) | Nr. 7 b (unbekannte Session, unbekannter oder fremder Workspace: Fehler nach dem Beginn mit `nicht_gewaehlt`; unbekannter Key: `kein_treffer`); Nr. 4 b (Ergebnis je Code der Auswahl) | Nr. 3 f (Kontextfehler an jedem Port, im Body des Requests: ein `End`, Ergebnis unverändert) | Nr. 7 c (`sequential` und `strict`: zwei Spans mit verschiedenem Ergebnis; `first` und `last`: gleiche Attribute, verschiedene `ID`) | Nr. 7 f (der Dienst kennt kein `Drain`: ein laufender Aufruf endet mit seinem `ctx`, der Span dort) |

Die Zeilen des Gesamtspans und des Zählers stehen in §6 von [slice-telemetry-einbau-gateway](../next/slice-telemetry-einbau-gateway.md), die des Capture-Spans und der Einträge der Dienste in §6 von [slice-telemetry-einbau-record](slice-telemetry-einbau-record.md).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-22; der Planner zieht Treffer und Zähler beim Start nach, der Zähler ist die Zahl der Dateien unter `evidence/`). Der Einbau trifft nebenläufige Pfade und Fakes, die `ctx` umhüllen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (6×) für die Fakes der Ports; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×) gilt für jede Wirkung, die der Detektor nicht meldet (W13); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, verkörpert). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (67×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (62×, `AGENTS.md` §3.17). Keine Beobachtung steht mit diesem Slice offen unter 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

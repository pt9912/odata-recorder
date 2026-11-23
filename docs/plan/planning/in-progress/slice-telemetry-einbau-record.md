# Slice slice-telemetry-einbau-record: Capture-Span `odata.record.capture` im Record-Dienst und Einträge der Dienste `record`, `session`, `maintenance` und des Audit-Adapters über den Adapter des Prozesses (Teil 3 des Schnitts von `slice-telemetry-einbau-gateway`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, `LH-FA-01`, `LH-QA-05`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-295` (Capture-Span; Nr. 1 bis 9), neu vom Architect vor dem Code · `SPEC-027` (spezifikation.md §5) · `SPEC-121` Nr. 1, 12 und 13 (Konfig, Protokoll und Herunterfahren des Record-Dienstes) · `SPEC-139` Nr. 3, 13 d und f (der Adapter des Prozesses als Telemetrie und Logger der Dienste) · `SPEC-231` Nr. 4, 7 und 8 (Verdrahtung der Dienste, Fänger je Stelle) · `SPEC-232` Nr. 2 und 3 · `SPEC-235` Nr. 5 (Eintrag `audit_fehlgeschlagen`) · `SPEC-230` Nr. 8 · `SPEC-109` · `ARC-002`, `ARC-005`, `ARC-012`, `ARC-021` (der Architect ordnet sie beim Schnitt keinem Teil zu).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-telemetry-einbau-record`, Worktree `/Development/odata-recorder-wt-telemetry-record`; der Zweig entsteht von `main` nach dem Start-Commit, der Implementer nimmt `main` vor dem ersten Code-Commit und vor dem Handoff per `git merge main` auf)

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md), Vorschlag des Architect vom 2026-11-21). **Datum:** 2026-11-22.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Record-Dienst setzt `odata.record.capture` (`SPEC-027`, `SPEC-295`) je Exchange mit den Pflicht-Attributen `recorder.session_id` und `recorder.capture_status` (der Status steht bei `StartSpan` auf `unbestimmt`, am Ende auf `complete`, `incomplete` oder `failed`, auch bei einem Exchange, der nicht angehängt wird); sein `Konfig` wächst additiv um `Telemetrie` (nil ergibt `model.ErrUngueltig`, `Log` bleibt Pflicht); der Dienst ruft `CountError` nie (den Zähler je Fehlercode führt der HTTP-Adapter, `SPEC-295` Nr. 6). Die Einträge der Dienste `record`, `session` und `maintenance` und des Audit-Adapters (`SPEC-121` Nr. 12 und die übrigen) laufen als Zeile über `slog.New(Handler())` des Adapters des Prozesses, ohne dass ein Dienst sein `*slog.Logger`-Feld, seinen Code oder den Wortlaut seiner Einträge ändert (`SPEC-231` Nr. 7; die Spezifikation gewinnt, siehe §3 Entscheidung 4 von [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md)); der Logger, den `Build` erhält, trägt danach allein die Einträge des Roots. Kein Query-Literal und kein Body wird Attribut. Der Teil setzt Teil 1 (Gesamtspan, Übergabe des Adapters) und Teil 2 ([slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md); der eine Adapter je Prozess in jedem Modus) voraus und ändert beide nicht.

**Übernimmt (Sendung von `slice-telemetry-port`, Teil Record; Geber: `slice-telemetry-port`; Schnitt des Planners vom 2026-11-22):** der Einbau in den Pfad `record`; der Port selbst und der Fake sind dort geliefert. Der Einbau in den HTTP-Adapter liegt in Teil 1, der in den Pfad `replay` in Teil 2.

**Übernimmt (Sendung von `slice-audit-log`, Stand beim Anlegen von `slice-audit-verdrahtung`; Geber: `slice-audit-log`):** `SPEC-231` Nr. 4 und 7 nennen den Adapter `audit` (Verdrahtung des Audit-Adapters mit dem Logger der Telemetrie) und den Protokolleintrag `audit_fehlgeschlagen` (`SPEC-235` Nr. 5, Schlüssel nur aus der Menge `Log`); der Architect hat sie mit dem Vertrag nachgetragen, die Verdrahtung im Composition Root folgt hier (Zeilen E2 und E4). Gilt für die Dienste `record`, `session` und `maintenance`, soweit sie im Root verdrahtet sind; die Dienste der Pfade `mock` und Workspace liegen bei `slice-telemetry-einbau-mock`.

**Übernimmt (Sendung von `slice-telemetry-log` §1, Schnitt in `slice-telemetry-log` und `slice-telemetry-handler`; Geber: `slice-telemetry-log` und `slice-telemetry-handler`; Eintrag des Planners vom 2026-11-07, beim Schnitt vom 2026-11-22 an diesen Teil gelegt):** (a) die Umstellung der Dienste `record`, `session` und `maintenance` auf `slog.New(Handler())` und die Verdrahtung des Adapters mit dem Standard-Fehlerstrom als Senke, der Systemuhr und `MinLevel` gleich `slog.LevelInfo` im Composition Root (`SPEC-231` Nr. 7), soweit die Dienste dort verdrahtet sind; die Dienste behalten ihr `*slog.Logger`-Feld und ändern ihren Code nicht; (b) der Wortlaut von `SPEC-121` Nr. 12 zum Port `Telemetry` (der Satz „einen Port `Telemetry` gibt es noch nicht“ veraltet mit diesem Teil; der Architect hat ihn mit dem Vertrag berichtigt); (c) der Fänger je Stelle zum akzeptierten Negativ L15 von `slice-telemetry-log`: eine Stelle, die einen Request-Wert in `session_id` oder `msg` ablegt, gibt ihn aus, der Schutz ist ein Test je Stelle der Einträge der Dienste (`SPEC-232` Nr. 2, *Ungeprüft*; Zeile E3). Das Ziel („statt über einen `*slog.Logger`“) und `SPEC-231` Nr. 7 (die Dienste behalten das Feld) sagten verschieden, was der Record-Dienst am Ende hält; die Spezifikation gewinnt (Entscheidung des Architect, §3 Nr. 4 von Teil 1; Ziel und DoD sind berichtigt).

**Übernimmt (Sendung von `slice-telemetry-handler` §6, Spec-Lücke; Geber: `slice-telemetry-handler`; Eintrag des Planners vom 2026-11-07, beim Schnitt vom 2026-11-22 an diesen Teil gelegt):** (d) der Wortlaut der Texte der Ereignisse `upstream` und `ablage` des Record-Dienstes. `SPEC-121` Nr. 12 sagt „festem Text“, nannte ihn aber für diese beiden nicht; die Tests des Handlers (Menge M13, `TestDienstEintraege`) nehmen die Texte des Codes (`record: Upstream ohne Antwort`, `record: Exchange nicht angehängt`, Quelle `src/hexagon/services/record/protokoll.go`), also ein Orakel aus dem Code. Der Architect hat den Wortlaut mit dem Vertrag festgelegt (der Text des Codes gilt, für alle drei Einträge, `SPEC-121` Nr. 12); der Fänger je Stelle aus (c) und die Fixture von `TestDienstEintraege` nehmen den Text aus der Spezifikation (Zeile E1). Der Handler ändert den Code der Dienste nicht (`slice-telemetry-handler` §1).

**Übernimmt (Sendung von [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) (Teil 1), Verdrahtung des Adapters; Geber: `slice-telemetry-einbau-mock`; Eintrag des Planners vom 2026-11-21 bei der Closure von Teil 1 jenes Schnitts, beim Schnitt vom 2026-11-22 geteilt; §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung):** (e) der Adapter des Prozesses (`SPEC-139` Nr. 13: ein `telemetry.New` je Prozess in `verdrahte` am Haken `NeueTelemetrie`, `Writer` der Fehlerstrom von `Start`, `Clock` die eine Uhr, `MinLevel` `INFO`) ist die `Telemetrie` des Record-Dienstes, und `slog.New(Handler())` dieses Adapters ist das `Log` der Dienste `record`, `session` und `maintenance` (Punkt (a) oben, Adapter und Senke sind dann der eine des Prozesses); der Teil des Replay-Dienstes (`SPEC-139` Nr. 3 und 13 c) liegt in Teil 2.

**Übernimmt (Sendung von [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) (Teil 2), Kopf eines Tests; Geber: `slice-telemetry-einbau-replay`; Eintrag des Planners vom 2026-11-23 bei der Closure von Teil 2, §6 dort Punkt (d); §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung):** (f) der Kopf von `TestEinAdapterJeProzessMitDerSenke` in `src/bootstrap/wurzel_test.go` sagt zu, im Modus `replay` erhalte der Replay-Dienst den Adapter des Prozesses; der Test zählt Adapter und prüft den `Writer`, die Übergabe an den Dienst fängt `TestWurzelReplayMatchLiegtImGesamtspan` (Zeile W10 von Teil 2). Der Kopf nennt nur, was der Test prüft (ein Adapter, `Writer` gleich der Senke von `Build`), und verweist für die Übergabe auf den Test von W10 (`AGENTS.md` §3.11). Der Teil berichtigt den Kopf (ein Kommentar in `wurzel_test.go`, ohne Zeile der Tabelle; der Block führt `src/bootstrap/` als Verzeichnis, die Zeilen zählen mit).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Gesamtspan, der Zähler je Fehlercode, `Konfig.Telemetrie` und `Konfig.Modus` des HTTP-Adapters und die Übergabe im Root — [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md) (Teil 1, liegt bei diesem Start in `done/`); der Capture-Span liest den Gesamtspan nur über das `ctx` (`SPEC-295` Nr. 3 e).
- Der Match-Span und der eine Adapter je Prozess in `replay` und `hybrid` — [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) (Teil 2, liegt bei diesem Start in `done/`).
- Port und Adapter `Telemetry`; Zeile, Stufen, Senke, `Handler`, Spans, Verteilung und Ring — `slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans`; alle liegen in `done/`, dieser Slice ändert sie nicht.
- Spans von Mock-Dienst und Workspace-Mutation — `slice-telemetry-einbau-mock` und `slice-telemetry-einbau-mock-mutation`.
- Spans und Zähler des Hybrid-Fallbacks — [slice-telemetry-einbau-hybrid](../open/slice-telemetry-einbau-hybrid.md) (Entscheidung des Users vom 2026-11-23, `welle-betrieb` §5, Punkt b); der Dienst `record` gehört nicht zum Modus `hybrid` (`SPEC-295` Nr. 7 f).
- Der Zähler je Fehlercode für `ODR-E2401` (`capture`) und die Codes von `ablage`: Variante A (Entscheidung des Users vom 2026-11-23, §6): allein der HTTP-Adapter zählt, was ein Client als Fehlerobjekt erhielt, und kein Dienst zählt zusätzlich (`SPEC-294` Nr. 5).
- Der Lasttest und seine Messung — `slice-lasttest` (`LH-QA-01.a` nimmt `record` aus).
- Das Handbuch-Kapitel über die Logzeilen des Prozesses (Gesamtspan, Einträge der Dienste als JSON-Zeile) — `slice-handbuch-telemetrie` (Entscheidung des Users vom 2026-11-21: eigener Slice, Start nach Teil 1 und diesem Teil, §4 dort); der Slice schreibt kein Kapitel und führt keinen Fehlercode, Schlüssel oder Kommando ein (siehe DoD).

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

- [ ] Capture-Span im Record-Dienst: `odata.record.capture` beginnt nach der Annahme, nach `ExchangeID` und `Clock.Now()` und vor `Senden`, trägt Kennung der Session und Status der Aufnahme (ein `Set`) und endet je Exchange genau einmal, nach dem Anhängen und dem Eintrag `ablage`, vor dem Ende des Laufs, das `Drain` abwartet; `Konfig.Telemetrie` mit nil-Prüfung, der Dienst ruft `CountError` und `Log` des Ports nie, das `ctx` des Aufrufs geht unverändert an `StartSpan` und an jeden der fünf Ports hinter dem Beginn (`SPEC-027`, `SPEC-295`, C1 bis C6, C8, C10, C11).
- [ ] Einträge der Dienste: `Log` von `record`, `session`, `maintenance` und `audit` ist `slog.New(Handler())` des Adapters des Prozesses, Attribute und Text der Einträge unverändert, der Text nach `SPEC-121` Nr. 12 (die Fixture von `TestDienstEintraege` und die Tests des Pakets `record` nehmen den Wortlaut der Spezifikation statt des Codes); der Eintrag `audit_fehlgeschlagen` steht als Zeile in der Senke; mit einem Fänger je Stelle der zehn Einträge, dass kein Request-Wert in `session_id` oder `msg` landet (Sendungen (a) bis (d) oben, Sendung L15 von `slice-telemetry-log`; E1 bis E4).
- [ ] Verdrahtung im Composition Root und Test durch Record: der Adapter des Prozesses ist die `Telemetrie` des Record-Dienstes, die Zeile `odata.record.capture` steht in der Senke als Kind des Gesamtspans, ohne Query-Literal und Body (Canary, C9); der Logger von `Build` trägt allein die Einträge des Roots (Sendung (e) von `slice-telemetry-einbau-mock`; C7, E2); der Kopf von `TestEinAdapterJeProzessMitDerSenke` nennt nur, was der Test prüft (Sendung (f) von `slice-telemetry-einbau-replay`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: der Slice führt keinen Fehlercode, keinen Schlüssel und kein Kommando ein; die Logzeilen des laufenden Prozesses (Gesamtspan, Einträge der Dienste als JSON-Zeile) beschreibt das Kapitel von `slice-handbuch-telemetrie` (Entscheidung des Users vom 2026-11-21, §6); das Kataloge-Gate ist grün, und §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando des Binary dazukam.
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
| `src/hexagon/services/record/` | update | `Konfig.Telemetrie` mit nil-Prüfung, Capture-Span (`SPEC-295`); `Log` bleibt ein `*slog.Logger` (`SPEC-231` Nr. 7), Text und Attribute der Einträge bleiben (`SPEC-121` Nr. 12) |
| `src/hexagon/services/session/`, `src/hexagon/services/maintenance/` | nur Tests | Fänger je Stelle der sechs Einträge (`SPEC-231` Nr. 8); kein Produktcode |
| `src/adapters/driven/telemetry/` | nur Tests | `TestDienstEintraege` (Menge M13) nimmt den Text der drei Einträge des Record-Dienstes aus `SPEC-121` Nr. 12 statt aus dem Code; kein Produktcode |
| `src/bootstrap/` | update | `Telemetrie` des Record-Dienstes, `slog.New(Handler())` als `Log` von `record`, `session`, `maintenance` und `audit.Config` (`SPEC-139` Nr. 13 d); der Logger von `Build` trägt allein die Einträge des Roots |
| `spec/spezifikation.md` | geliefert (Architect, auf `main`) | `SPEC-295` und Nachzug `SPEC-121`, `SPEC-139`, `SPEC-230`, `SPEC-231`, §5; dieser Teil schreibt sie nicht |
| Tests | update, neu | Journal-Fake der Telemetrie im Paket `record`; Tests durch den Composition Root mit dem echten Adapter; angepasst: `Telemetrie` in jedem `record.Konfig`, die Tests, die Einträge der Dienste über den übergebenen Logger von `Build` lesen (`audit_logger_test.go` und verwandte), der Kopf von `TestEinAdapterJeProzessMitDerSenke` in `wurzel_test.go` (Sendung (f) von `slice-telemetry-einbau-replay`, ein Kommentar) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-21 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-295`, `SPEC-231`, `SPEC-121`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidungen des Architect** (2026-11-21; die Aufträge 1 bis 10, je mit Ort in `spec/spezifikation.md`) stehen in §3 von [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md); für diesen Teil gelten Nr. 2 (Attribute `recorder.session_id` und `recorder.capture_status`), Nr. 3 (Eltern-Kind: der Capture-Span endet nach dem Gesamtspan, aber vor der Rückkehr des Handlers; `Drain` kehrt erst zurück, wenn jeder Capture-Span der Session beendet ist), Nr. 4 (Ersatz des Loggers: das Verdikt, die Entscheidung gilt), Nr. 5 (kein Dienst zählt; Variante A, vom User am 2026-11-23 entschieden, §6), Nr. 6 (Aufrufstellen R1 bis R3 und fünf Ports mit `ctx` hinter dem Beginn), Nr. 8 (Wortlaut der Einträge des Record-Dienstes) und Nr. 9 (ein Adapter je Prozess).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-21, Zeilen C1 bis C11 und E1 bis E4 des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). `Nr.` ist die Kennung des ungeschnittenen Plans: **C** und **E** gehören zu diesem Teil, **G** zu Teil 1, **W** zu Teil 2. *Journal-Fake*: `FakeTelemetry`, dessen Aufrufe (und die der Fakes der Ports) in ein gemeinsames Protokoll schreiben, mit Buch je Span (Attribute, Zahl der `End`). Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| C1 | (`SPEC-295` Nr. 1) `record.New` liefert `model.ErrUngueltig` und keinen Dienst bei nil `Telemetrie`, ohne Aufruf von `Telemetrie`; `Log` bleibt Pflicht | Fälle im Test der Konfiguration: nil `Telemetrie` und die 7 bisherigen nil-Felder (`Sessions`, `Exchanges`, `Upstream`, `Pipeline`, `Uhr`, `IDs`, `Log`) gleich 8; Orakel Literal; 0 Aufrufe | nil-Prüfung für `Telemetrie` entfällt; `Log` aus der Pflicht genommen | andere |
| C2 | (Nr. 2) Kein Span und kein Aufruf von `Telemetrie`, was vor der Annahme endet | Menge M-VorAnnahme: Argumente 4 (`ctx` nil, leere Methode, `Laenge` −2, `Body` nil bei `Laenge` 1), `ODR-E5005` 1, `ODR-E5006` und `ODR-E9001` (roher `#`) 2, `ODR-E5003` 3 (leerer Wert, zwei Werte, mehrere offene Sessions), `ODR-E7003` 1, `ODR-E5004` 1, `ODR-E5009` (Session nicht `open`) 1, Fehler der Ablage 2, Kontextfehler 2, `ODR-E5009` im `Drain` 1 gleich 18; `FakeTelemetry.Aufrufe()` leer; Zählprüfung 18 | `StartSpan` am Anfang von `Record`; `StartSpan` vor `nimmAn`; Span auch für `ODR-E5009` des `Drain` | andere |
| C3 | (Nr. 3 a, 6 R1) Beginn nach der Annahme, nach `ExchangeID` und `Clock.Now()`, vor `Senden` | Journal-Fake: Folge `Sessions.Get`, `ExchangeID`, `Now`, `StartSpan`, `Senden` (Wahl über den Header) und `Sessions.List`, `ExchangeID`, `Now`, `StartSpan`, `Senden` (die eine offene Session) gleich 2; Orakel Literal | `StartSpan` vor `nimmAn`; `StartSpan` nach `Senden`; `StartSpan` vor `ExchangeID` | andere |
| C4 | (Nr. 3 b, 7 a, 6 R3) `End` genau einmal je Exchange, nach dem Anhängen und dem Eintrag `ablage`, vor dem Ende des Laufs, das `Drain` abwartet | Menge M-Ausgänge: Wege (`Close` nach EOF beider Seiten, `Close` ohne EOF, ohne Antwort) 3 × Anhängen (gelingt, `Append` scheitert, `ListBySession` scheitert) 3 gleich 9, dazu ein laufender Exchange, den `Drain` beendet, gleich 10; je Fall `StartSpan` 1, `End` 1; Journal-Folge `Append`, `Log(ablage)` (wo nicht angehängt), `End`, Rückkehr von `Drain`; Zählprüfung 10 | `End` nur bei gelungenem Anhängen; `End` vor `Append`; `End` nach dem Ende des Laufs (`Drain` kehrt vor `End` zurück); zweites `End` | andere |
| C5 | (Nr. 4 a, b, 7 a) `recorder.session_id` ist `Exchange.SessionID`; `recorder.capture_status` steht bei `StartSpan` auf `unbestimmt` und am Ende auf dem Status der Aufnahme, ein `Set` | Menge M-Status: `complete` (`Close` nach EOF) 1; `incomplete` (Antwort durch `Close` ohne EOF abgebrochen, `Drain` nach vollständigem Request, ohne Antwort mit vollständigem Request) 3; `failed` (`multipart/mixed`, `Drain` bei laufender Request-Aufnahme, ohne Antwort bei nicht vollständigem Request, Request-Aufnahme mit `abbruch`) 4 gleich 8; Orakel Literal, `Set` 1; Session über den Header und über die eine offene (2 Fälle) | Wert vor `Abschluss` gesetzt; `Set` nur bei `complete`; Platzhalter leer; `session_id` aus dem Header statt aus `Exchange.SessionID`; Status des Anhängens statt der Aufnahme | andere |
| C6 | (Nr. 3 d, 6) `StartSpan` erhält das `ctx` des Aufrufs unverändert; keiner der fünf Ports hinter dem Beginn erhält das `ctx` von `StartSpan` | Schleife über 5 Stellen (`Senden`, `Tee` des Requests, `Tee` der Antwort, `ListBySession`, `Append`): `FakeTelemetry.StartSpan` liefert ein `ctx` mit eigenem Wert, kein Port sieht ihn, `StartSpan` sieht die Identität des `ctx` von `Record`; Zählprüfung 5 als Literal | `ctx, span := StartSpan(…)`; das `ctx` des Spans nur an `Append`; nur an `Tee` | andere |
| C7 | (Nr. 3 e, `SPEC-139` Nr. 13 d und f) Der Record-Dienst erhält den Adapter des Prozesses; der Capture-Span ist Kind des Gesamtspans und endet nach ihm | Test durch den Composition Root (Modus `record`, Fake des Upstreams, `BuildMitSenke`, `Hochfahren`): `GET` mit Antwort: Zeile `odata.gateway.request`, dann Zeile `odata.record.capture` mit `parent_id` gleich `span_id`; Zeile als Literal (beide Attribute, `complete`), `Dropped` alle 0; ein Fall ohne Elternspan (`Record` über den Handle): `parent_id` fehlt | `StartSpan(context.Background(), …)`; ein zweiter Adapter für den Record-Dienst; `End` vor der Rückkehr der Antwort | andere |
| C8 | (Nr. 1, 5, 6) Der Dienst ruft `CountError` und `Log` des Ports nie; der Quelltext nennt `StartSpan`, `Set` und `End` je einmal | Schleife über 9 Szenarien mit Code (`ODR-E5003`, `ODR-E5004`, `ODR-E5005`, `ODR-E5006`, `ODR-E5009`, `ODR-E7003`, `ODR-E4001` mit 502 und 504, `ODR-E9001`) und die 3 Einträge (`capture`, `upstream`, `ablage`): `CountError` 0, `Log` des Ports 0; `go/parser`: `StartSpan` 1, `Set` 1, `End` 1, kein `Log`, kein `CountError`; Zählprüfung 9 und 3 | `CountError(c)` in `protokolliere`; `Log` am Port statt `slog`; zweiter `Set` | andere |
| C9 | (Nr. 8) Kein Wert aus dem Request erreicht ein Attribut oder die Senke | Menge M-Stellen-Record: 9 Stellen (Pfad, Name und Wert der Query, Name und Wert eines Headers, Body des Requests, Header und Body der Antwort, Text der Ursache eines Fehlers von `Senden`) × 2 Arten (Exchange angehängt, ohne Antwort) gleich 18; Canary aus Bytewerten in den Formen Kennung und Token, geprüft an jedem Attribut des Fakes und an der Senke des echten Adapters; Zählprüfung 9 | je Stelle der Wert in `recorder.session_id` und in `recorder.capture_status` (18 Mutationen) | andere |
| C10 | (Nr. 9 a) *n* gleichzeitige Exchanges ergeben *n* `StartSpan` und *n* `End`, jeder Span mit Kennung der Session und Status seines Exchange, kein Span zweimal | Wirkungs-Fänger: 16 Goroutinen × 2 Sessions × 3 Status in 20 Läufen unter `make test-race`; der Fake führt Buch je Span; eine Schranke an `Senden` hält alle 16 Exchanges offen; Rate je Prozess | Span im `sessionZustand` geteilt; `End` nur jeden zweiten Exchange; der Status aus einer geteilten Variable | nebenläufig |
| C11 | (Nr. 3 b, 9 b) Nach der Rückkehr von `Drain` ist jeder Span der Session beendet | Wirkungs-Fänger: 16 laufende Exchanges an einer Schranke, `Drain(0)` gegen gleichzeitig beginnende `Record`: bei der Rückkehr `End` gleich der Zahl der Annahmen (Orakel aus dem Bericht und der Zahl der Annahmen), kein offener Span; Rate je Prozess | `End` in einer Goroutine nach dem Ende des Laufs; `End` nach `endeLauf` | nebenläufig |
| E1 | (`SPEC-121` Nr. 12) Die drei Einträge tragen den Text der Spezifikation; der Test der Zeile nimmt ihn aus ihr | Schleife über 3 Texte (Literal aus `SPEC-121` Nr. 12) gegen den `slog.Record` des Dienstes (3 Szenarien: `capture`, `upstream`, `ablage`) und gegen die Fixture von `TestDienstEintraege` (gleich dem Literal der Spezifikation, nicht dem Wert des Codes); Zählprüfung 3 | Text je Eintrag geändert (3 Mutationen); `angehängt` ohne `ä` | andere |
| E2 | (`SPEC-139` Nr. 13 d, `SPEC-231` Nr. 7) `Log` von `record`, `session`, `maintenance` und `audit` ist `slog.New(Handler())` des Adapters des Prozesses; der Logger von `Build` sieht keinen Eintrag eines Dienstes | Test durch den Root (`BuildMitSenke`, Haken der Konstruktoren halten die Konfig): Menge M-Verdrahtung: 4 Stellen, je ein Eintrag mit erlaubtem Schlüssel (`ereignis`) über das gehaltene `Log`: die Zeile steht als JSON in der Senke, der übergebene Logger bleibt leer; `maintenance` und `audit` in 4 Modi, `record` und `session` im Modus `record`: 10 Fälle; Root-Einträge (`start: bereit`) bleiben Textzeilen | `Log: h.log` je Stelle (4 Mutationen) | andere |
| E3 | (`SPEC-231` Nr. 8; Sendung L15) Kein Eintrag der Dienste `record`, `session` und `maintenance` legt einen Wert des Requests oder einen Fehlertext in `msg`, `session_id` oder ein anderes Feld | Menge M-Quellen-Einträge: je Stelle ein Fänger mit Canary an jeder Quelle, die sie lesen kann: `record` 3 Stellen × 9 Quellen (Nr. 8 von `SPEC-295`) gleich 27, `session` 3 Stellen × Text des Fehlers des Ports gleich 3, `maintenance` 4 Stellen × Text des Fehlers gleich 4, zusammen 34 Paare; geprüft am `slog.Record` des Dienstes (Handler-Fake) auf Canary in `msg` und jedem Attribut; Zählprüfung 10 Stellen (3, 3, 4) und 34 | je Stelle `err.Error()` in `msg` oder `grund`; der Pfad in `msg`; ein Header-Wert in `session_id` (10 Mutationen) | andere |
| E4 | (`SPEC-231` Nr. 4, `SPEC-235` Nr. 5) Der Eintrag `audit_fehlgeschlagen` steht als Zeile in der Senke, mit Text, Stufe und Feldern der Stelle, ohne Verwurf | Test durch den Root: `audit.jsonl` ist ein Verzeichnis, ein Eintrag wird ausgelöst (Abbruch der Wartung), 4 Modi: Zeile als Literal (`audit: Eintrag nicht geschrieben`, `ERROR`, `ereignis`, `error_code`, `grund`), `Dropped` alle 0, kein Fehlertext, kein Pfad | `Log: h.log` im Audit-Adapter (Textzeile); Fehlertext in der Zeile; `error_code` fehlt | andere |

Zeilen: C1 bis C11 und E1 bis E4 gleich 15 (*a* = 2 (C10, C11), *b* = 13); die akzeptierten Negative stehen unter der Tabelle und zählen nicht als Zeile (`SPEC-128` Nr. 3 e gilt für Zeilen).

**Verteilung beim Schnitt** (Planner, 2026-11-22): Die Zeilen sind im Wortlaut des Architect übernommen, keine geändert. Die Zeile C7 prüft den Gesamtspan als Elternspan durch den Composition Root und setzt Teil 1 voraus; der eine Adapter je Prozess in allen Modi (Teil 2, W10) ist die Voraussetzung von E2 und E4 in den Modi `replay` und `hybrid`. Die Zählprüfung der Zeile E3 (10 Stellen) und die der Zeile C9 (9 Stellen) gelten für Variante A, die der User am 2026-11-23 entschieden hat.

**Mengen mit Zählprüfung** (Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15):

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| M-VorAnnahme (C2) | 4 + 1 + 2 + 3 + 1 + 1 + 1 + 2 + 2 + 1 gleich 18 | kein Aufruf | 18 |
| M-Ausgänge (C4) | 3 Wege × 3 Anhänge + 1 gleich 10 | `End` 1, Folge | 10 |
| M-Status-Record (C5) | `complete` 1, `incomplete` 3, `failed` 4 gleich 8 | Status | 3 Werte = `model.CaptureStatus`, 8 |
| M-ctx-Record (C6) | `Senden`, `Tee` × 2, `ListBySession`, `Append` gleich 5 | Identität des `ctx` | 5 = `SPEC-295` Nr. 6 |
| M-Stellen-Record (C9) | 9 Stellen × 2 Arten gleich 18 | kein Canary | 9 = `SPEC-295` Nr. 8 |
| M-Verdrahtung (E2) | `record`, `session`, `maintenance`, `audit` gleich 4 Stellen; 10 Fälle | JSON in der Senke, Logger von `Build` leer | 4 = `SPEC-139` Nr. 13 d |
| M-Quellen-Einträge (E3) | 27 + 3 + 4 gleich 34 Paare aus 10 Stellen | kein Canary in `msg` und Attribut | 10 Stellen (3, 3, 4) = `SPEC-231` Nr. 8 |
| M-Gleichzeitig (C10, C11) | 16 Goroutinen × (6 Sessions und Status, laufende Exchanges), 20 Läufe | Buch je Span | 16 |

**Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** Pfad, Query und Body des Requests und der Antwort (Prüfgrenze: ein kurzer Canary als Wert; Grund: der Dienst setzt nie einen Wert des Requests, die Länge ändert den Weg nicht; Folge: ein Wert von 1 MiB wird nicht eigens geprüft, `SPEC-119` begrenzt den Body).

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; im Wortlaut des Architect, die Zeilen der Spans von Teil 1 und Teil 2 stehen dort):

| Paar oder Modus | Erwartung | Fänger |
|---|---|---|
| Gesamtspan und Capture-Span | `parent_id` gleich `span_id`; die Zeile des Gesamtspans steht vor der des Capture-Spans | C7 |
| Capture `complete` und nicht angehängt (`Append` scheitert) | Span `complete`, Eintrag `ablage`, `End` einmal; der Fehler zählt nicht | C4, C5 |

**Akzeptierte Negative dieses Vertrags** (Grund, Folge und Maß): (1) die Dauer des Capture-Spans (`SPEC-295` *Ungeprüft*): eine Zeitmessung im Test schwankt mit der Last (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`); die Dauer misst `slice-lasttest`. Folge: Arbeit, die ein Eingriff vor den Beginn oder hinter das Ende legt, ohne die Folge der Aufrufe zu ändern, bleibt ungefangen; Maß: C3 und C4 binden 2 Folgen, nicht die Zeit. (2) Ein Ausfall der Senke ändert Antwort und Ergebnis nicht: wie das Negativ (2) in §3 von Teil 1 (keine Operation von `Telemetry` liefert einen Fehler, 0 von 5 Operationen); der Fänger liegt am Adapter der Telemetrie (`SPEC-232` Nr. 3). (3) Die Zahlen der Mengen (5 Ports mit `ctx` hinter dem Beginn, die zehn Einträge der Dienste, die 9 Quellen von `SPEC-295` Nr. 8) gegen die Tabellen der Spezifikation prüft kein Test, weil die Stufe `source` des Images `spec/` nicht trägt; Folge: ändert jemand die Aufrufstellen von `SPEC-295` Nr. 6 oder die Stellen von `SPEC-231` Nr. 8 ohne den Test, sieht es erst ein Review (der Architect hat am 2026-11-21 von Hand gezählt: 5 Ports mit `ctx` hinter dem Beginn im Record). (4) Ein Typed-nil in `Telemetrie` (`SPEC-295` *Ungeprüft*): `New` prüft nil; Folge: eine Panik beim ersten Span, im Betrieb unerreichbar (der Root übergibt den Adapter). (5) Der Weg des Fehlers von `Tee` im Record-Dienst (`SPEC-295` Nr. 3 b, 4 b): kein Lauf erreicht ihn (`SPEC-121` *Ungeprüft*); Folge: der Span endete dort mit `unbestimmt`; Maß: 1 von 3 Wegen des Exchange ist unerreichbar, C4 führt die beiden anderen. (6) Der Zähler zählt `ODR-E2401` und die Codes von `ablage` nicht (Entscheidung 5 von Teil 1, Empfehlung A): kein Fänger, weil es die Zusage ist; Folge: ein Betreiber liest sie nur im Log; die Entscheidung des Users vom 2026-11-23 (Variante A, §6) hält dieses Negativ.

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jeder Wert von `recorder.capture_status`, jede Aufrufstelle des Ports, jede Stelle der Einträge der Dienste), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md) (Teil 1) und [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) (Teil 2) liegen in `done/`; mit ihnen liegen die Bedingungen der Teile in `done/` (`slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans`, `slice-http-adapter`, `slice-http-lebenszyklus`, `slice-replay-service`, `slice-bootstrap-m2`, `slice-telemetry-einbau-mock`); dazu `slice-record-service` in `done/`. Die Frage A/B zum Zähler je Fehlercode (§6) hat der User am 2026-11-23 entschieden (Variante A). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-295`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-einbau-gateway` (Teil 1): *Code-Abhängigkeit*. Der Gesamtspan, dessen `ctx` den Capture-Span zum Kind macht (C7 prüft es durch den Root), und die Übergabe des Adapters des Prozesses an `httpodata.New`.
- `slice-telemetry-einbau-replay` (Teil 2): *Reihenfolge*. Beide Teile ändern `src/bootstrap/wurzel.go` und laufen nacheinander (`welle-betrieb` §5); Teil 2 bringt den einen Adapter je Prozess in `replay` und `hybrid`, auf dem E2 und E4 in allen vier Modi aufbauen.
- `slice-record-service` (`welle-record-replay`): der Ort des Capture-Spans (`src/hexagon/services/record`); der Dienst liegt in `done/`.
- `slice-telemetry-log`, `slice-telemetry-handler`: *Code-Abhängigkeit*. Das Binary schreibt erst mit Zeile, Senke und `Handler()` eine Logzeile; die Verdrahtung nach `SPEC-231` Nr. 7 braucht beide, sonst gäbe es keinen `slog.Handler` für die Dienste.
- `slice-telemetry-spans` und `slice-telemetry-port`: Port, Fake und Adapter mit Span-Zeile, die dieser Teil ruft.
- `slice-audit-log` und `slice-audit-verdrahtung`: der Adapter `audit` und die Verdrahtung des Audit-Adapters, deren Logger dieser Teil auf `slog.New(Handler())` stellt (E2, E4).
- `slice-bootstrap-m2` (`welle-record-replay`): der Composition Root und der Handle, über den der Test Record durchläuft (kein CLI, keine Admin-API).

**Reihenfolge und Konflikte** (Planner, 2026-11-22): Die drei Teile laufen nacheinander: Teil 1, Teil 2, dieser Teil; Teil 2 und dieser Teil ändern `src/bootstrap/wurzel.go`, Teil 1 änderte in `src/bootstrap/` `lebenszyklus.go` (`starteListener`) und Tests (`telemetrie_gateway_test.go`, `telemetrie_wurzel_test.go`, `vertrag_http_test.go`, `audit_betrieb_test.go`). Dieser Teil ändert in `src/bootstrap/` auch Tests, die Einträge der Dienste über den Logger von `Build` lesen (`audit_logger_test.go` und verwandte); `audit_betrieb_test.go` (Zählstelle der Arten, 6 → 7) hat `slice-admin-sessions-beenden` geändert, der Zweig von Teil 1 trägt es seit dem Merge von `main`. `slice-v2-bootstrap` (Teil A; änderte `wurzel.go`, `lebenszyklus.go` und `konfig*.go`) liegt beim Start in `done/` und ist in `main`; die Basis dieses Slice trägt es. Wer `main` nachzieht, misst die fremden Zeilen an den Pfaden des Blocks mit (`SPEC-128` Nr. 10 g): er misst vor dem Merge, nennt den Messpunkt und trägt danach, wenn nötig, `ueberschreitung` mit Grund ein (Nr. 5 d), keine Rückführung.

**Stand der Bedingungen beim Start** (Planner, 2026-11-23, `main` = `208f633c` nach den Moves): Teil 1 [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md) und Teil 2 [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) liegen in `done/` (an den Dateien geprüft), mit ihnen `slice-telemetry-port`, `slice-telemetry-log`, `slice-telemetry-handler`, `slice-telemetry-spans`, `slice-http-adapter`, `slice-http-lebenszyklus`, `slice-replay-service`, `slice-bootstrap-m2`, `slice-telemetry-einbau-mock`, `slice-record-service`, `slice-audit-log` und `slice-audit-verdrahtung`. Der Vertrag steht auf `main`: `SPEC-295` samt Randformen, der Nachzug in `SPEC-121` Nr. 12, `SPEC-139` Nr. 13 d und f, `SPEC-231` Nr. 7 und 8 und `SPEC-235` Nr. 5; Tabelle, Raster und Mengen stehen in §3 und §6. Die Frage A/B hat der User am 2026-11-23 entschieden (Variante A, §6). Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`. Der Trigger der Welle ist eingetreten (Roadmap); die Ausnahme je Slice nach `welle-betrieb` §2 wird nicht gebraucht. **Kollisionen** (`git diff --name-only main...<Zweig>`): [slice-admin-tls](slice-admin-tls.md) (Zweig `slice-admin-tls`, in Arbeit) ändert dreizehn Dateien: `.a-check.yml`, `harness/vertraege/SPEC-039-architektur-gate.md`, `spec/spezifikation.md`, `src/bootstrap/admintls.go`, `src/bootstrap/admintls_test.go`, `src/bootstrap/export_test.go`, seinen Plan, drei Pläne in `open/` (`slice-admin-tls-betrieb`, `slice-bootstrap-m6`, `slice-test-security`), `welle-admin-api.md`, die Roadmap und einen Review-Report; **keine Datei des Blocks** (weder `wurzel.go` noch eine der Test-Dateien des Blocks), gemeinsam im Verzeichnis `src/bootstrap/` ist allein `export_test.go`, die der Block nicht führt (siehe *Größenmarke*), und gemeinsam in der Roadmap und in `welle-admin-api.md` (die schreibt nur der Integrationsschritt, Übergangsregeln Nr. 4). Der Zweig `spec-m6-vertrag` (Architect, Vertrag und Plan von [slice-bootstrap-m6](../open/slice-bootstrap-m6.md)) trägt keinen Commit vor `main` (`git diff --name-only main...spec-m6-vertrag` ist leer); [slice-bootstrap-m6](../open/slice-bootstrap-m6.md) ändert später `src/bootstrap/` und startet nach allen Slices der Welle, auch nach diesem. Weitere Slices in `in-progress/` gibt es nicht. Die Reihenfolge der Merges entscheidet der Integrationsschritt; die Dateien sind disjunkt.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-21, `basis`, `schaetzung` und `marke` vom Planner beim Start am 2026-11-23 vor dem ersten Code-Commit nach `AGENTS.md` §3.14 gesetzt): Der Planner hat die Tabelle am Start neu gezählt und mit `make groessenmarke` geprüft: 15 Zeilen (C1 bis C11, E1 bis E4), *a* = 2 (C10, C11 `nebenläufig`), *b* = 13 (alle übrigen `andere`), *f* = 1 (Journal-Fake der Telemetrie im Paket `record`), *p* = 30 (Feld, nil-Prüfung, Beginn, `Set` in `setzeZustand`, `End` in `schliesseAb`, vier `Log` im Root): *S* = 85·2 + 80·13 + 200·1 + 30 = 1440, 135·1440 = 194400, *M* = 100 · ⌈19,44⌉ = 2000; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223). Der Teil liegt 783 unter der Schnittgrenze: jede weitere Zeile der Tabelle kostet 80, zehn weitere lösen die Schnittfrage neu aus (*S* = 2240; neun geben 2160). Der User hat für die Frage A/B Variante A gewählt (2026-11-23): die Tabelle bleibt bei 15 Zeilen, und die zwei Zeilen der Variante B (*b* = 15, *S* = 1600, *M* = 2200) entfallen. **Messbasis** `208f633c` (der Commit des Verweis-Nachzugs nach dem Move nach `in-progress/` auf `main`, der letzte Commit von `main` vor dem Code-Start; die Messung ist 0). Eine ältere Marke gibt es nicht (kein Eintrag `ersetzt`). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. **Der Block nennt unter `src/bootstrap/` die Dateien einzeln** (`SPEC-128` Nr. 10 g), nicht das Verzeichnis: `slice-admin-tls` ändert dort `admintls.go`, `admintls_test.go` und `export_test.go` (zusammen 1080 Zeilen), und das Verzeichnis im Block ließe sie nach dem Merge von `main` mitzählen (Marke 2000). Legt der Implementer eine weitere Datei an oder ändert er eine weitere Datei, nennt er sie im selben Commit im Block (`src/bootstrap/export_test.go` nur, wenn er dort etwas ändern muss; dann zählen die 37 fremden Zeilen von `slice-admin-tls` mit, falls dessen Zweig vorher mergt).

```groessenmarke
pfad: src/hexagon/services/record/
pfad: src/hexagon/services/session/
pfad: src/hexagon/services/maintenance/
pfad: src/adapters/driven/telemetry/
pfad: src/bootstrap/wurzel.go
pfad: src/bootstrap/wurzel_test.go
pfad: src/bootstrap/audit_logger_test.go
pfad: src/bootstrap/audit_betrieb_hilfen_test.go
pfad: src/bootstrap/audit_betrieb_test.go
pfad: src/bootstrap/telemetrie_wurzel_test.go
pfad: src/bootstrap/telemetrie_record_test.go
produktcode: 30
fakes: 1
basis: 208f633c
schaetzung: 1440
marke: 2000
```

**Messpunkt beim Start** (Planner, 2026-11-23, `make groessenmarke` im Hauptbaum unter der Sperre `/tmp/odata-recorder-heavy.lock`, Last des Hosts 11,9 vor und 11,3 nach dem Lauf, `main` = `208f633c` ohne den Start-Commit; Exit 0, keine Befunde, kein Eintrag nötig): `groessenmarke slice-telemetry-einbau-record.md: ok, Messung 0 von Marke 2000` und `Messpunkt slice-telemetry-einbau-record.md: zeilen=15 nebenlaeufig=2 fakes=1 produktcode=30 schaetzung=1440 marke=2000 messung=0` (Schätzung des Blocks gleich Tabelle, Marke gleich Formel). Beim Handoff und bei der Closure trägt der Planner die Zeile des Programms mit Stand und von Hand gemessenem Produktcode nach (`SPEC-128` Nr. 9).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2000 des Blocks, vor dem Code nach der Formel gesetzt (`AGENTS.md` §3.14); ein weiterer Schnitt ist nicht vorgesehen, die Schätzung liegt weit unter der Schnittgrenze.
- `in-progress` → `open` (blockiert): der Wortlaut von `SPEC-121` Nr. 12 lässt sich über `slog.New(Handler())` des Adapters nicht erhalten (siehe §6), oder der Capture-Span lässt sich nicht beenden, ohne dass `Drain` vor `End` zurückkehrt (`SPEC-295` Nr. 3 b entscheidet das Gegenteil).

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

- Frage A/B zum Zähler je Fehlercode: der Zähler zählt, was ein Client als Fehlerobjekt erhielt; `ODR-E2401` (`capture`) und die Codes von `ablage` stehen nicht in `odr_errors_total` (Entscheidung 5 von Teil 1, `SPEC-294` Nr. 5). Variante A: der Vertrag gilt, der Adapter zählt, kein Dienst zählt. Variante B: der Record-Dienst zählt `capture` und `ablage` zusätzlich (zwei Stellen mehr in `SPEC-295` Nr. 6, zwei Zeilen der Tabelle). Grund und Alternative stehen in `SPEC-294` Nr. 5 — **Ausgang:** entschieden (User, 2026-11-23): Variante A; der Vertrag `SPEC-295` Nr. 6 und die Tabelle (C8, 15 Zeilen) gelten, wie sie stehen; das akzeptierte Negativ (6) in §3 bleibt.
- Der Einbau ändert den Konstruktor des Record-Dienstes (Feld `Log`); Fänger und Tests der Vorgänger, die das Feld setzen, brechen — **Ausgang:** entfallen für `Log`: das Feld bleibt (Entscheidung 4 von Teil 1, `SPEC-231` Nr. 7, `SPEC-295` Nr. 1), die Fänger der bisherigen Einträge bleiben unverändert grün; eingetreten ist die additive Pflicht `Telemetrie`: jedes `record.Konfig` der Tests wächst um das Feld (Zeile C1).
- Pfad ohne Lauf unter dem Detektor (Abbruch von Client und `Drain`, Ausfall der Senke) — **Ausgang:** entfallen als Frage des Plans: der Abbruch trägt Wirkungs-Fänger (C4, C11), der Ausfall der Senke ist akzeptiertes Negativ (2) mit Grund, Folge und Maß; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×): die Raten je Prozess misst der Implementer (`AGENTS.md` §3.10).
- Der Capture-Span endet nach dem Gesamtspan, aber `Drain` und das Herunterfahren dürfen nicht vor `End` zurückkehren — **Ausgang:** entschieden (Architect, 2026-11-21, `SPEC-295` Nr. 3 b, e): der Handler wartet auf `Close`, `Drain` kehrt erst zurück, wenn jeder Capture-Span der Session beendet ist; Fänger C4 und C11 (`SPEC-121` Nr. 13).
- Handbuch: Mit der Verdrahtung schreibt das Binary die Einträge der Dienste `record`, `session` und `maintenance` als JSON-Zeile (`SPEC-231` Nr. 7), sichtbare Logzeilen des Prozesses; ein Kapitel dazu ist nur als gebundenes Beispiel zulässig (`SPEC-200` Nr. 1), und das Beispiel-Format kennt keine Anfrage an den Listener (`SPEC-200` Nr. 6 d, 7 b) — **Ausgang:** eingetreten → Folge-Slice `slice-handbuch-telemetrie` (Entscheidung des Users vom 2026-11-21, `welle-betrieb` §5, Punkt c): dieser Slice schreibt kein Kapitel; der Nehmer trägt das Kapitel in §1 (*Übernimmt*, Geber der Reihe nach die Teile von `slice-telemetry-einbau-gateway`, `AGENTS.md` §3.13) und startet nach Teil 1 und diesem Teil (§4 dort).
- Das Ziel (Protokolleintrag „über den Port `Telemetry` statt über einen `*slog.Logger`“) widerspricht `SPEC-231` Nr. 7 (die Dienste behalten das `*slog.Logger`-Feld und erhalten `slog.New(Handler())`); die Spezifikation gewinnt, der Wortlaut des Ziels und der DoD war zu berichtigen — **Ausgang:** eingetreten: Verdikt des Architect (Entscheidung 4 von Teil 1): die Entscheidung gilt, der Plan hat falsch behauptet; Ziel und DoD sind berichtigt (2026-11-21).
- Fänger je Stelle zu L15 (`slice-telemetry-log`) für die Einträge der Dienste — **Ausgang:** weiter offen; Adresse: dieser Teil (Zeile E3, `SPEC-231` Nr. 8; DoD, Eintrag des Planners vom 2026-11-07).
- Texte der Ereignisse `upstream` und `ablage` standen nicht im Wortlaut von `SPEC-121` Nr. 12 („festem Text“); die Tests von `slice-telemetry-handler` nehmen die Texte des Codes (Orakel aus dem Code), der Wortlaut war Spec-Lücke (Geber `slice-telemetry-handler`, Befund F3 des Reviews, V3 der Verifikation) — **Ausgang:** eingetreten: der Wortlaut steht in `SPEC-121` Nr. 12 (der Text des Codes gilt, für alle drei Einträge); der Implementer tauscht das Orakel in `TestDienstEintraege` und im Paket `record` (Zeile E1).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-21): das Raster unten trägt in jeder Zelle einen Ort oder `gilt nicht` mit Grund; eine Randform, die der Implementer im Code findet, geht an den Architect (`AGENTS.md` §3.12).
- Schnitt und Größenmarke: der Plan ist am 2026-11-22 vor dem Code aus `slice-telemetry-einbau-gateway` geschnitten (ungeschnitten *S* = 4605 über der Schnittgrenze; dieser Teil *S* = 1440, Teil 1 *S* = 1875, Teil 2 *S* = 1290) — **Ausgang:** entfallen (Planner, 2026-11-23, beim Start): *S* = 1440 und Marke 2000 sind aus der Tabelle nachgerechnet und stehen im Block in §4 (`basis` `208f633c`, `schaetzung` 1440, `marke` 2000, vor dem ersten Code-Commit, `AGENTS.md` §3.14); keine Schnittfrage, 783 unter der Schnittgrenze (§4, *Größenmarke*).
- Die Marke steht beim Start (`BEO-REPO/marke-beim-start-nicht-gesetzt`, 1×) — **Ausgang:** entfallen (Planner, 2026-11-23): der Block trägt `basis`, `schaetzung` und `marke`, bevor der Zweig entsteht.
- Gemeinsame Dateien in `src/bootstrap/` mit einem zugleich laufenden Slice: dieser Teil ändert Tests, die Einträge über den Logger von `Build` lesen; `slice-admin-sessions-beenden` hat `audit_betrieb_test.go` geändert (gemergt, im Zweig von Teil 1 seit dessen Merge von `main`), `slice-v2-bootstrap` (Teil A) hat `wurzel.go`, `lebenszyklus.go` und `konfig*.go` geändert (liegt in `done/`) — **Ausgang:** entfallen (Planner, 2026-11-23, beim Start): `git diff --name-only main...slice-admin-tls` (13 Dateien) nennt keine Datei des Blocks; gemeinsam im Verzeichnis ist allein `src/bootstrap/export_test.go` (37 Zeilen von `slice-admin-tls`), die der Block nicht führt; der Zweig `spec-m6-vertrag` trägt keinen Commit (§4, *Stand der Bedingungen beim Start*).
- Zeilen fremder Slices an den Pfaden des Blocks: ein anderer Slice, der auf `main` landet, bevor dieser endet, könnte eine Datei des Blocks ändern, und die Messung zählte sie mit (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`, 6×; dieser Slice wäre mit fremden Zeilen ein siebter Vorgang) — **Ausgang:** weiter offen; beim Start geprüft (Planner, 2026-11-23): die Basis trägt keine fremde Zeile, und der einzige zugleich laufende Slice (`slice-admin-tls`) ändert keine Datei des Blocks; Adresse: Implementer (misst beim Handoff und vor jedem Merge von `main` `git diff --numstat <basis>..<Elter des Merge> -- <Pfade des Blocks>`, nennt die fremden Zeilen in der Messpunkt-Zeile und setzt `ueberschreitung` nur bei Überschreitung).
- Reihenfolge der drei Teile: Teil 2 und dieser Teil ändern `src/bootstrap/wurzel.go`, Teil 1 `lebenszyklus.go` — **Ausgang:** entschieden (Planner, 2026-11-22): sie laufen nacheinander (1, 2, 3); dieser Teil startet, wenn Teil 2 in `done/` liegt (§4, `welle-betrieb` §5).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-21, Zeilen des Capture-Spans und der Einträge der Dienste des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund; „Nr.“ ohne Nennung heißt die SPEC der Zeile: `SPEC-295`, bei den Einträgen der Dienste die genannte):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch (Client, `ctx`, `Drain`) | Wiederholung | Herunterfahren |
|---|---|---|---|---|---|---|
| Capture-Span (`SPEC-295`; Exchange anhängen, Capture-Status) | Nr. 1 (nil `Telemetrie`); Nr. 7 b (Request ohne Body ändert den Span nicht; eine fehlende Session-Kennung ist `ODR-E5003` vor dem Span, Nr. 2) | Nr. 7 b (leerer Body; der leere Header ist `ODR-E5003`) | Nr. 7 c (unbekannte oder fremde Session: `ODR-E7003`, `ODR-E5004` vor der Annahme, kein Span); Nr. 4 c (Kennung außerhalb der Form: `wert`, `pflicht`) | Nr. 3 f und 7 a (Client und `Drain` beenden den Kontext, der Span endet einmal, mit dem Status der Aufnahme) | Nr. 7 d | Nr. 7 e und 3 b (nach der Rückkehr von `Drain` ist jeder Span der Session beendet; ein abgewiesener Exchange trägt keinen) |
| Einträge der Dienste über `slog.New(Handler())` (`SPEC-231` Nr. 4, 7 und 8, `SPEC-121` Nr. 12) | `Log` nil ist `model.ErrUngueltig` (`SPEC-121` Nr. 1, `SPEC-295` Nr. 1); ein nicht gesetztes Attribut entfällt, kein Feld ohne Wert (`SPEC-230` Nr. 4, `SPEC-231` Nr. 1) | `SPEC-230` Nr. 3 und 4 (ein leerer Wert ist ein Verstoß der Form, `wert`, das Attribut entfällt, der Eintrag wird geschrieben); `SPEC-231` Nr. 3 (ein leerer `msg` ist zulässig) | `SPEC-231` Nr. 8 (ein Wert des Requests in `msg` oder `session_id`: Fänger je Stelle); `SPEC-230` Nr. 4 a (ein Schlüssel außerhalb der Menge: `schluessel`) | `SPEC-230` Nr. 1 und `SPEC-231` Nr. 5 (`Log` wirkt auch bei beendetem `ctx`); `SPEC-121` Nr. 12 (`upstream` bei beendetem `ctx` von `Record`: kein Eintrag) | `SPEC-231` Nr. 1 (jeder Aufruf eine Zeile, gleiche Einträge nicht zusammengefasst) | `SPEC-121` Nr. 13 und `SPEC-295` Nr. 3 b (der Eintrag entsteht vor dem Ende des Laufs, das `Drain` abwartet); `SPEC-232` Nr. 3 (Ausfall der Senke); gilt nicht für den Handler: er schreibt synchron im Aufruf (`SPEC-232` Nr. 5 d) |

Die Zeilen des Gesamtspans und des Zählers stehen in §6 von [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md), die des Match-Spans in §6 von [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-22) und beim Start erneut (2026-11-23; der Zähler ist die Zahl der Dateien unter `evidence/`, Treffer und Zähler beim Start nachgezogen). Der Einbau trifft nebenläufige Pfade und Fakes, die `ctx` umhüllen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (8×) für die Fakes der Ports; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (10×) gilt für jede Wirkung, die der Detektor nicht meldet (C10, C11); `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×) und `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×) für die Größenmarke (§4); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×, verkörpert). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (77×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (72×, `AGENTS.md` §3.17). Offen unter 3× steht mit diesem Slice allein `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×, §6: beim Start gesetzt).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

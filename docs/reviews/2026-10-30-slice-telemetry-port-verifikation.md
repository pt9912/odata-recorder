# Verifikations-Report: slice-telemetry-port (Teil P der Kette P, L, S) - 2026-10-30

**Rolle:** Verifier (Modul 11) - "Bauen wir es richtig?" gegen DoD und Spezifikation, nicht gegen den Bericht des Implementers und nicht gegen den Diff.

**Gegenstand:** `slice-telemetry-port`, Zweig `slice-telemetry-port`, HEAD `ed3142b` (Basis des Slice `a0396b0`, Marke gesetzt in `8c0aac5`, Marke 3000). Fix-Commits: `a557519` (Entscheidungen des Architect in der Spezifikation), `84a8d7e`, `ed3142b`. Vorgänger-Berichte nur als Eingang: [Review](2026-10-30-slice-telemetry-port-review.md) (F-1 bis F-11, Commit `c728c37`) und [Re-Review](2026-10-30-slice-telemetry-port-review-2.md) (R-1 bis R-7, Commit `876d7a0`).

**Gegen was geprüft:** Plan `docs/plan/planning/in-progress/slice-telemetry-port.md` (DoD, §3 Tabelle P01 bis P17 samt Nachträgen, §4, §6), `SPEC-230` (Nr. 1 bis 4, 8 bis 10), `SPEC-232` (Nr. 1, 5 a; 5 f soweit P) in `spec/spezifikation.md`, `AGENTS.md` §3.7 bis §3.18.

**Modell:** claude-sonnet-5-5 - **Datum:** 2026-10-30 - **Arbeitsweise:** Sensoren selbst gefahren (Container `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, jeder Lauf unter flock auf die gemeinsame Sperrdatei der schweren Läufe); Mutationen und Probe-Pakete auf Kopien in einem Wegwerfverzeichnis des Scratchpads, im Repo wurde nichts außer diesem Report geändert.

---

## Verdikt

**NICHT BESTANDEN** - wegen eines MEDIUM (V-1), das ein reines Fänger-Defizit ist: Der Code und der Gegenstand der DoD sind konform, aber eine Zusage von `SPEC-230` Nr. 9 ("Stand zu einem Zeitpunkt") hat in einem Fenster der Sperre weder einen Fänger noch ein benanntes Negativ (`AGENTS.md` §3.15: eine Zusage ohne Fänger ist ein MEDIUM). Die Mutation bleibt in der ganzen Suite grün (Detektor 0 von 50, ohne Detektor 0 von 50, `-test.count=50` grün); ein Wirkungs-Fänger aus einem Absatz Test fängt sie in 30 von 30 Prozessen. Fix: ein Test (Vorlage unter §5) oder ein benanntes Negativ des Architect. Danach genügt ein Blick auf den einen Fänger, kein neuer Gesamtlauf.

Alle anderen DoD-Punkte, soweit sie dem Verifier gehören, sind belegt (§2). Die Offenen aus dem Re-Review: R-5, R-6, R-7 geschlossen; R-2, R-3, R-4 geschlossen; R-1 bis auf eine Restlücke der Klasse (V-2, LOW).

Befunde: 0 HIGH, 1 MEDIUM, 3 LOW, 4 INFO.

## 1. Sensor-Belege (selbst gefahren, nicht übernommen)

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` | Exit 0 (echter Exit-Code des Befehls `flock ... make gates`, nicht über `tail` maskiert) | Ausgabe: a-check `gesamt: 0 Befund(e)`; baseline-verify OK 54 Dateien; d-check 368 Dateien, 0 Befunde; Stufen `build`, `test`, `race`, `lint`, `groessenmarke` durchgelaufen; schema-check und migration-check ok |
| Cache der Stufen | Die Docker-Schichten von `pruefung test`, `pruefung race` und `lint.sh` standen als `CACHED` in der Ausgabe: Quellstand identisch mit einem früheren Lauf. Das Grün ist damit kein neuer Lauf dieser Schichten; ich habe die Pakete darum selbst gefahren (nächste Zeilen). | Ausgabe des Gate-Laufs, Zeilen mit `CACHED` |
| `go test -race -count=1` der drei Pakete (selbst) | ok: `adapters/driven/telemetry` 25,7 s, `ports/support` 1,0 s, `ports/support/supporttest` 2,5 s (Host unter Last) | Container `odata-recorder:race` |
| Eigenes Probe-Paket, plain und unter `-race` | ok in allen drei Paketen | §4 |
| `make groessenmarke` | `slice-telemetry-port.md: ok, Messung 2222 von Marke 3000`; Messpunkt `zeilen=17 nebenlaeufig=2 fakes=1 produktcode=580 schaetzung=2150 marke=3000 messung=2222`. Nachgerechnet: `git diff --numstat a0396b0..HEAD` über die beiden Pfade des Blocks, 2221 Einfügungen + 1 Löschung = 2222. Auslastung 2222/3000 = 0,74 (1,03 der Schätzung). | Gate-Ausgabe und eigene Summe |
| `go.mod`, `go.sum` | gegenüber `a0396b0` leer im Diff (0 Zeilen) | `git diff a0396b0 -- go.mod go.sum` |
| Binary unverändert | Byte-gleich: `sha256` des Binary aus `src/bootstrap/main` (`-trimpath -buildvcs=false -ldflags="-s -w -X main.version=dev"`) an `a0396b0` und an `ed3142b` ist in beiden Ständen `a48cc10672798458795e888837996ab141fe9ce2dfe95bae667bdccb47b32ee5` (3666080 Byte). Kein Aufrufer des Ports außerhalb der Pakete `ports/support` und `adapters/driven/telemetry`. | eigener Bau beider Stände im Container |
| Dateien außerhalb der beiden Pfade des Blocks, des Plans und der Reviews | nur `spec/spezifikation.md` (Architect-Commit `a557519`) | `git diff --name-only a0396b0..HEAD` |
| Rohbytes (`AGENTS.md` §3.18) | in allen geänderten Dateien unter `src/` und im Plan keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP (`LC_ALL=C grep -P`); dieser Report ist ebenso gegengelesen (§9) | |
| Suppression (`AGENTS.md` §3.2) | keine `//nolint` im Diff | `git diff a0396b0..HEAD -- src` |

## 2. DoD Punkt für Punkt

| DoD-Punkt (Plan §2) | Ergebnis | Beleg |
|---|---|---|
| 1 Ports `Telemetry` (3 Methoden), `Span` (2), `TelemetryReader` (1); Typen `Attr`, `Art`, `Snapshot`, `Verteilung`, `SpanRecord`, `Verwurf`; Konstruktoren `Str`, `Int`, `Millis`, `ErrorCode`; `Spans()` (5 Spans, 11 Pflicht-Attribute), `Eimer()` (11 Grenzen), `Verwuerfe()` (7), `Snapshot.Kopie()`; `FakeTelemetry`/`FakeSpan` | erfüllt | Quelle gelesen; Probe `TestProbeKopieNullwert` (Konstruktoren, Eimer), `TestProbeSchluesselMengenMitOrakel` (Spans gegen eigene Tabelle: 5 und 11), `TestProbeCountErrorGrenzen` (`Eimer()`, `Verwuerfe()`, `Spans()` liefern frische Slices); Mutationen K6, K7, K8 rot |
| 2 `New` mit `Config` weist ab (nil-Writer, nil-Clock, `MaxSpans` unter 0 und über 65536; 0 gilt als 1024; jeder `MinLevel` gültig; Fehler ohne Wert) | erfüllt | `TestProbeNewWiederholung` (6 Werte außerhalb, 7 innerhalb, 7 `MinLevel` bis MinInt/MaxInt, Fehlertext gleich `model.ErrUngueltig`); T2, T7, T8 rot |
| 2 `CountError` zählt je Code der Tabelle (33), fremder Code als `fehlercode` | erfüllt | `TestProbeCountErrorGrenzen` (Nachbarn +-1 und +-1000 jedes Codes, 0, MinInt, MaxInt, MaxInt32, MinInt32), `TestProbeNewWiederholung` (Modell je Adapter, 20000 Schritte); T4, T5 rot |
| 2 `Snapshot` vollständig (auch mit 0) und tiefe Kopie | erfüllt, mit Restlücken V-1 (Fenster) und V-2 (Fixture) | `TestProbeNewWiederholung`, `TestProbeKopieKlasse` (486 Stände) |
| 2 `Log` wendet die Regeln an und zählt Verwürfe, schreibt keine Zeile; `StartSpan` ohne Wirkung; kein Handler, keine Senke, `MinLevel` nicht angewendet | erfüllt (Quelle: `Log` ruft nur `filtern` und `verwerfen`; `Writer`, `Clock`, `MinLevel` werden gehalten, nicht gelesen; `StartSpan` gibt `ctx` und `keinSpan{}` zurück; kein `Handler`, kein Ring, keine Füllung der Verteilung) | `TestKernSchreibtNichts`, `TestLogOhneMinLevel`; T6 rot. F-10 (Ersatz in L und S) ist offen mit Adresse Planner und hier nicht doppelt |
| 2 nebenläufig sicher, `make test-race` ohne Befund | erfüllt für die Zugriffe, die die Tests ausführen; Lücken V-1, V-3 | §5; `make gates` und eigener Lauf unter `-race` |
| 2 Formen schließen die Zeichen der Träger aus | erfüllt | Probe 460 543 + 93 691 Werte gegen eigene Ausdrücke (§4) |
| 3 `make gates` grün | erfüllt, Exit 0 | §1 |
| 4 Review durchgeführt, Report liegt vor | erfüllt (zwei Reports) | `docs/reviews/` |
| 5 Doku-Update | erfüllt: SPEC-230 bis 232 in §5; Plan folgt dem Code (zwei Verweise siehe V-4) | `git diff a0396b0..HEAD -- spec` |
| 6 Handbuch: Binary unverändert | erfüllt (byte-gleich, §1); §7 nennt es | Plan §7 |
| 7 Closure-Notiz mit Lerneintrag | offen (Planner, §10) | Plan §7 "Wird bei der Closure gefüllt." |
| 8 Beobachtungs-Register fortgeschrieben | offen (Integrationsschritt, §10) | |
| 9 Jedes Risiko aus §6 mit Ausgang | offen: Risiken mit Ausgang *weiter offen* und Adresse (Planner/User), zwei Adressen aus diesem Report (§10) | Plan §6 |
| 10 Drei Paarungen | Closure | |

## 3. Plan gegen Code: Zeilen P01 bis P17 gegen den Vertrag

Mutations-Kennungen aus §7. *best.* = vorhandene Tests des Repos; *Probe* = mein Probe-Paket.

| Zeile | Vertrag | Fänger im Repo | Urteil |
|---|---|---|---|
| P01 | SPEC-230 Nr. 1 | `TestAttrKonstruktoren` | ok; K8 (`Millis` auf ganze Millisekunden) rot (best.) |
| P02 | Nr. 2, 3, 4: 13 und 11 Schlüssel | `TestSchluesselmengen`, `TestSpansDaten` | ok; R14, R15 rot (best.); Probe: 6 Operationen x 35 Schlüssel = 210 Paare gegen eigene Tabelle |
| P03 | Nr. 3: neun Wertformen, Grenzen | `TestWertformGrenzen`, `TestWertformGrenzenVonHand` | ok; R2, R3, R7, R8, R10, R12 rot (best.); R6 grün (best.), V-5 |
| P04 | Nr. 3, `SPEC-232` Nr. 2: jedes Byte außerhalb | `TestWertformZeichen` (28672), `TestWertformStellen` (61440) | ok; R1, R3, R4, R5 rot (best.) |
| P05 | Nr. 3: 33 Codes | `TestFehlercodeForm` | ok; Probe 93 691 Werte |
| P06 | Nr. 8 | `TestCountErrorJeCode` | ok; T4, T5 rot (best.) |
| P07 | Nr. 8 | `TestCountErrorUnbekannt` | ok (5 Fälle statt 4 im Plan: auch `MinInt`) |
| P08 | Nr. 9 | `TestSnapshotVollstaendig`, `TestFakeSnapshotStand`, `TestFakeNullwert` | ok; T3, T4 rot (best.) |
| P09 | Nr. 9 | `TestSnapshotKopie` (Adapter und `ports/support`), `TestSnapshotKopieBeispiel` | ok mit Restlücken V-1 (Fenster), V-2 (Fixture); K2, K4, K5, K9, K12, K13, K14 rot (best.) |
| P10 | `SPEC-232` Nr. 1 | `TestNewConfig`, `TestNewUnabhaengig` | ok; T2, T7, T8 rot (best.); Probe: 6 Adapter aus derselben Config gegen ein Modell je Adapter |
| P11 | Nr. 10 | `TestFakeAufzeichnung`, `TestFakeSpanJeOperation`, `TestFakeSpanID`, `TestFakeNullwert` | ok mit Lücke V-3; X1, X2, X3, X6, X7, X9, X11 rot (best.) |
| P12 | Nr. 10 | `TestFakeEltern` | ok; X8 (bereits abgebrochener `ctx` am Start) grün, V-5 |
| P13 | Nr. 1, ARC-021 | `TestTelemetrySignaturen`, `TestPortSignaturen`, `TestAdapterErfuelltPorts` | ok; der Plan nennt im Fänger `TestPortSignaturen`, in der Mutationsliste `TestTelemetrySignaturen` (beide existieren; Namen sind Vorschläge) |
| P14 | Einleitung SPEC-230 | `make a-check`, `TestAdapterImporte`, `git diff go.mod go.sum` | ok; I1 bis I4 alle rot; R-5 geschlossen (§6) |
| P15 | `SPEC-232` Nr. 5 a | `TestCountErrorGleichzeitig`, `TestAdapterGleichzeitig` | ok für die gemessenen Fenster (§5); Lücke V-1 |
| P16 | Nr. 10 | `TestFakeGleichzeitig`, `TestFakeSnapshotGleichzeitig` | ok für Aufzeichnung und Stand; Lücke V-3 (Lesen) |
| P17 | Nr. 4 | `TestVerwurfRangfolge`, `TestLogOhneMinLevel` | ok; R11 rot (best.); T1 grün (best.), V-5 |

Raster (§6) *Erzeugen / Zähler erhöhen / Lesen*: jede Zelle nennt ihren Ort, die Orte tragen den Wortlaut (`SPEC-232` Nr. 1, `SPEC-230` Nr. 8 und 9); keine im Code entschiedene Zelle gefunden. Zusatztabelle *Verstöße am selben Schlüssel*: fünf Zeilen in `TestVerwurfRangfolge`; mein Orakel auf 30 000 Zufallsfolgen (Schlüssel doppelt, fremd, Art falsch, Wert gültig/ungültig) stimmt mit dem Code überein.

## 4. Probe-Paket (Generatoren und eigenes Orakel, Saat ab 200000)

Orakel: reguläre Ausdrücke und Tabellen ausschließlich aus dem Wortlaut von `SPEC-230` Nr. 3, nicht aus `regeln.go`. Alle Probes bestanden am Stand `ed3142b` (plain und unter `-race`).

| Klasse | Generator | Umfang | Saat |
|---|---|---|---|
| Textformen (Token, Kennung, Name, Methode, Blob) | jede Stelle x 256 Bytewerte (Längen 1, 2, 3, Mitte, Grenze-1, Grenze), Einfügen jedes Bytes an jeder Stelle, Kürzungen und Verlängerungen bis 2 über die Grenze, 20 000 Zufallswerte je Form | 460 543 Werte; schließt die Namensstellen 71 bis 127 ein (Name der Länge 128, alle 128 Stellen x 256) | 200001 |
| Fehlercode | 33 Codes: jede Stelle x 256, jede Kürzung, Voran-/Anhängen jedes Bytes, Kleinschreibung, Leerraum, Verdopplung; Codes außerhalb über `ErrorCode` | 93 691 Werte | - |
| Zahl, Status | Grenzen, Zufall; Status 100+2^32 (Verengung der Breite), MinInt64, MaxInt64 | je Wert gegen Orakel | 200002 |
| Dauer | 24 handgesetzte Werte (0, +0.0, -0.0 via `math.Copysign`, kleinster positiver, negativ kleinster, NaN mit und ohne Vorzeichenbit, +-Inf, MaxFloat64, subnormal, Nachbarn von 0 mit `Nextafter`, `Float64frombits` von Bit 63 allein und mit Bit 0) und 20 000 Zufallsbitmuster; `Millis` mit negativen, MinInt64- und MaxInt64-Dauern | Orakel: nicht NaN, nicht Inf, Vorzeichenbit nicht gesetzt | 200002 |
| Schlüssel x Operation, Art | 6 Operationen x 35 Schlüssel (24 des Katalogs, 11 fremde Schreibweisen: `Mode`, `mode ` mit Leerzeichen, leer, `timestamp`, `level`, `msg`, ...); 24 Schlüssel x 7 Arten (-1, 0, 1, 2, 3, 4, 100); 30 000 Zufallsfolgen mit Rangfolge-Orakel (Schlüssel vor Wert, erster angenommener Wert gilt) | 210 Paare, 168 Art-Fälle, 30 000 Folgen | 200003, 200004 |
| `New` x Wiederholung | `MaxSpans` MinInt bis MaxInt (13 Werte), 7 `MinLevel`, nil-Writer/-Clock; 6 Adapter aus derselben Config, 20 000 Schritte (`CountError` bekannt/unbekannt, `Log` mit drei Verwürfen, `Snapshot` mit Änderung an allen Sammlungen), Modell je Adapter, Stand alle 97 Schritte | | 200005 |
| `Snapshot.Kopie()` | 6 Saaten x 81 Zustände (`Errors`, `Durations`, `Dropped`, `Spans` je nil/leer/gefüllt), je Verteilung `Counts` und je `SpanRecord` `Attrs` in zufälligem Zustand und zufälliger Länge; Gleichheit mit `reflect.DeepEqual` (unterscheidet nil und leer), Adressen aller Backing-Arrays disjunkt, jedes Element in beide Richtungen geändert | 486 Stände | 200100 bis 200105 |
| Fake | 40 Saaten x 300 Schritte über die 6 Operationen gegen ein Zählmodell (`ctx`-Identität, Argumente, Eltern, `Anzahl` auch für Operationen außerhalb: -3 bis 11, MinInt, MaxInt); Eltern-Kette; `SetzeStand`/`Snapshot` als Kopie in beide Richtungen; 64 Goroutinen x 300 Runden mit 6 Operationen | | 200200 bis 200239, 200300 |
| Nebenläufig (Wirkung) | 48 Goroutinen x 400 Runden `CountError` bekannt/unbekannt und `Log` mit zwei Verwürfen, dazwischen `Snapshot`; genaue Summen | | - |

## 5. Nebenläufige Zusagen: Raten selbst gemessen

Messender: der Verifier. Image `odata-recorder:race`; Test-Binary je Mutation einmal gebaut (`go test -race -c` und `go test -c`), je Zelle 50 Prozesse mit `-test.count=1` und `-test.run` auf die Tests des Repos. *Detektor*: Binary mit `-race` (die Zelle von `make test-race`); *Wirkung*: Binary ohne Detektor; *Wiederholung*: ein Prozess ohne Detektor mit `-test.count=50`. DR = Prozesse mit `DATA RACE`; Absturz = Prozesse mit `fatal error` oder `panic:`. **Last:** Load Average (1 Minute) laut der Kernel-Datei loadavg zu Beginn jeder Zeile im Container, zwischen 7 und 18; keine künstliche Last. **Abweichung vom Auftrag:** eine Einzelmessung ist hier eine Zeile (drei Zellen) in einem eigenen flock-Aufruf, nicht eine Zelle, weil die Schlange der Sperre lang war.

Adapter, Tests `TestCountErrorGleichzeitig` und `TestAdapterGleichzeitig`; eigene Formulierungen, jedes Fenster einzeln:

| Id | Mutation | Last | Detektor, 50 Prozesse | Wirkung, 50 Prozesse | Wiederholung `-count=50` |
|---|---|---|---|---|---|
| - | keine (Gegenprobe) | 9,2 | 0 von 50 rot | 0 von 50 | grün |
| W1 | `CountError` bekannter Zweig: Wert lesen, Sperre freigeben, Sperre nehmen, Wert+1 schreiben | 16,7 | 50 von 50 rot, DR 0 | 50 von 50 | rot |
| W2 | `CountError` unbekannter Zweig: Fenster ebenso | 7,1 | 50 von 50 rot, DR 0 | 50 von 50 | rot |
| W3 | `verwerfen` (von `Log`): Fenster je Grund | 14,2 | 50 von 50 rot, DR 0 | 50 von 50 | rot |
| W4 | `Snapshot`: Sperre vor `Kopie()` freigegeben | 9,7 | 50 von 50 rot, DR 0, Absturz 50 (`concurrent map`) | 50 von 50, Absturz 50 | Absturz |
| W5 | `Snapshot`: `Errors` unter einer Sperre geklont, Sperre freigegeben, der Rest unter einer zweiten (Fenster zwischen den Sammlungen) | 18,4 | **0 von 50 rot** | **0 von 50** | **grün** |
| W6 | `verwerfen` ohne Sperre | 10,3 | 50 von 50 rot, DR 50, Absturz 50 | 50 von 50, Absturz 50 | Absturz |
| W7 | `CountError`: Sperre erst nach der Prüfung des bekannten Zweigs (Lesen der Map ohne Sperre) | 13,0 | 50 von 50 rot, DR 50, Absturz 50 | 50 von 50, Absturz 50 | Absturz |

Die Fenster W1 bis W3 meldet der Detektor nicht (DR 0); der Wirkungs-Fänger des Repos fängt sie im Detektor-Lauf in 50 von 50 Prozessen, ohne Detektor ebenso. Das Fehlen der Sperre (W6, W7) fängt der Detektor in 50 von 50, W4 fängt die Laufzeit mit einem Absturz (wie im Plan, F-3). **W5 bleibt in jeder Spalte grün: V-1.**

Wirkungs-Fänger des Verifiers für das Fenster W5 (Vorlage für V-1): 16 Schreiber x 3000 Runden (`CountError(E2001)`, dann `CountError(0)`), 8 Leser x 1500 Stände; Orakel in jedem Stand: `Dropped[fehlercode] <= Summe(Errors) <= Dropped[fehlercode] + 16` (gilt in einem zusammenhängenden Stand, weil jeder Schreiber erst einen bekannten, dann einen unbekannten Code zählt).

| Binary | Last | 30 Prozesse `-test.count=1`, ohne Detektor | 1 Prozess `-test.count=30` |
|---|---|---|---|
| Stand `ed3142b` (Gegenprobe) | 10,5 | 0 von 30 rot | grün |
| W5 | 10,5 | **30 von 30 rot** | rot |
| W4 (Absturz der Laufzeit) | 10,6 | 30 von 30 rot | Absturz |

Fake, Tests `TestFakeGleichzeitig` und `TestFakeSnapshotGleichzeitig`:

| Id | Mutation | Last | Detektor, 50 Prozesse | Wirkung, 50 Prozesse | Wiederholung `-count=50` |
|---|---|---|---|---|---|
| - | keine (Gegenprobe, zweimal gemessen) | 9,7 und 7,7 | 0 von 50 rot (beide) | 0 von 50 (beide) | grün (beide) |
| V1 | `aufzeichnen`: Slice gelesen, Sperre freigegeben, danach angehängt (Fenster) | 8,9 | 50 von 50 rot, DR 0 | 50 von 50 | rot |
| V2 | `Anzahl` ohne Sperre | 14,6 | **0 von 50** | **0 von 50** | **grün** |
| V3 | `Aufrufe()` ohne Sperre | 10,6 | **0 von 50** | **0 von 50** | **grün** |
| V5 | `SetzeStand` ohne Sperre | 10,0 | 50 von 50 rot, DR 50 | 36 von 50 | rot |

V2 und V3 bleiben grün: `Anzahl` und `Aufrufe()` werden in den Tests des Repos nur nach `wg.Wait()` gerufen, nie gleichzeitig zu Schreibern; V-3. Die Raten des Plans sind Messwerte des Implementers unter höherer Last (Load Average 20 bis 98); meine Zahlen liegen bei 7 bis 18 und sind für die Detektor-Spalte gleich (50 von 50), für die Wirkung ohne Detektor gleich oder höher (R-6, §6).

## 6. Die Offenen des Re-Reviews

| Punkt | Urteil | Beleg |
|---|---|---|
| R-5 "jede Datei des Pakets" | geschlossen: `TestAdapterImporte` liest `*.go` des Verzeichnisses einschließlich `_test.go` (`filepath.Glob("*.go")`); Kopf und Zeile P14 sagen "jede Datei des Verzeichnisses, auch der Dateien `_test.go`". Die Begründung "a-check scannt sie ebenso, `.a-check.yml` hat kein `exclude`" stimmt (`.a-check.yml` ohne `exclude`; die Gate-Ausgabe listet `_test.go`-Dateien unter den ungeschichteten, also scannt a-check sie). Eigene Mutationen I1 bis I4 (ein Import außerhalb der Menge in einer externen Testdatei; in einer Produktdatei ein Import, der nur in Tests erlaubt ist; in einer anderen Testdatei ein anderer Driven Adapter; in `export_test.go` `ports/outbound`) alle rot. | §3 P14 |
| R-6 Raten des Plans | geschlossen: der Plan sagt jetzt, die Spalte *Wirkung* sei ein Messwert unter Last und kein Teil der Zusage, die Zusage trage der Detektor; die Zahlen des Re-Reviews (29 von 30) stehen daneben. Meine Messung (§5): Detektor-Spalte in jeder Zelle mit Mutation 50 von 50 (Adapter und Fake, ausgenommen die grünen Mutationen W5, V2, V3); Wirkung ohne Detektor 50 von 50, nur `SetzeStand` ohne Sperre 36 von 50 (Plan: 27 von 50). | Plan §3, Absatz nach der Tabelle *Raten der Fix-Runde* |
| R-7 doppelte Fake-IDs | geschlossen: akzeptiertes Negativ mit Grund (Eindeutigkeit der IDs ist nur für den Adapter zugesagt, `SPEC-232` Nr. 5 e, Fänger in S) und Folge (ein Einbau-Test, der gleichzeitig startet und die ID als Schlüssel nutzt, wird nicht vom Fake gefangen, er nimmt den Adapter). Gelesen; die Rate der Mutation selbst (ID-Fenster) ist ungelaufen (§12). | Plan §3 "Akzeptiertes Negativ (R-7)" |
| R-1 `Kopie()` auf jeder Ebene | bis auf eine Restlücke geschlossen: die Menge "jede Sammlung jeder Ebene x nil/leer/gefüllt" ist gedeckt (K4, K5 und die Mutationen des Re-Reviews rot); die Menge "jede Position x jeder Zustand" nicht: V-2 | §7 |
| R-2 Fake `Snapshot` gegen `SetzeStand` | geschlossen: V5 (`SetzeStand` ohne Sperre) 50 von 50 unter dem Detektor; `Snapshot` ohne Sperre habe ich nicht selbst gemessen (§12), der Plan nennt 50 von 50 | §5 |
| R-3 Stellen 71 bis 127 | geschlossen: `TestWertformStellen` (61440 Werte) und mein Probe (Name der Länge 128, jede Stelle x 256) stimmen überein | §4 |
| R-4 Dauer-Grenze | geschlossen: Fixtures `0`, `SmallestNonzeroFloat64`, `1e-300`, `1e-6`, `-0.0`, `-SmallestNonzeroFloat64` im Repo; mein Probe fügt `1e-320`, `Nextafter`, 20 000 Zufallsbitmuster hinzu, ohne Abweichung | §4 |

## 7. Eigene Mutationen (Elemente, die weder Reviewer noch Implementer nannten)

57 Mutationen auf Kopien des Quellstands (R 18, T 8, K 14, X 13, I 4), ein Lauf je Mutation und Paketgruppe; dazu 7 W- und 4 V-Mutationen mit Raten (§5). *best.* = vorhandene Tests des Repos (ohne Probe), *Probe* = mein Probe-Paket. Eine Mutation, die nur das vet-Gate des Aufbaus brach, wurde neu formuliert (R12). R9 (aus Versehen äquivalent: das Vorzeichenbit fängt dieselben Werte) habe ich nicht gefahren.

| Id | Mutation | best. | Probe |
|---|---|---|---|
| R1 | Kennung nimmt `~` | rot | rot |
| R2 | Token darf mit `_` beginnen | rot | rot |
| R3 | Name darf mit `.` beginnen | rot | rot |
| R4 | Blob-Präfix: erstes Zeichen nicht geprüft | rot | rot |
| R5 | Großbuchstabe bis `[` | rot | rot |
| R6 | Status auf 32 Bit verengt (`int32(Ganz)`) | **grün** | rot |
| R7 | Zahl auf 32 Bit verengt | rot | rot |
| R8 | Dauer: `MaxFloat64` abgewiesen | rot | rot |
| R10 | `Int` für eine `Float`-Art angenommen | rot | rot |
| R11 | Rangfolge: Doppelschlüssel nach dem Wert geprüft | rot | rot |
| R12 | Kennung `-` abgewiesen | rot | rot |
| R13 | Blob ab 71 Zeichen (Längenprüfung `<` statt `!=`) | grün | grün (äquivalent: `nurZeichen` begrenzt auf 64) |
| R14 | `Mode` in der Menge `Log` | rot | rot |
| R15 | `recorder.entity_set` als Kennung | rot | rot |
| R16 | Status 404 abgewiesen | grün | rot (Einzelwert einer Menge von 500) |
| R17, R18, R19 | Einzelwerte abgewiesen (Zahl 2^40, Blob `sha256:ff...`, Methode `PATCH`) | grün | grün (unbeschränkte Mengen, kein Befund) |
| T1 | `Log` zählt höchstens 2 Verwürfe je Aufruf | **grün** | rot |
| T2 | `MaxSpans` 65536 abgewiesen | rot | rot |
| T3 | `Dropped` ohne den siebten Grund | rot | rot |
| T4 | `Errors` ohne den ersten Code | rot | rot |
| T5 | `E2001` als unbekannter Code | rot | rot |
| T6 | `StartSpan` liefert abgeleiteten `ctx` | rot | grün (nicht Gegenstand der Probe) |
| T7 | `MaxSpans` -1 gültig | rot | rot |
| T8 | `MaxSpans` im Fehlertext | rot | rot |
| K1 | `Counts` mit genau einem Element geteilt | **grün** | rot |
| K2 | `Attrs` mit genau einem Element geteilt | rot | rot |
| K3 | `Kopie` überspringt `SpanRecord` 1 | **grün** | rot |
| K4 | leere `Durations` zu nil | rot | rot |
| K5 | `Dropped` mit einem Eintrag geteilt | rot | grün (Probe hat nie genau einen) |
| K6 | Eimer 5001 ms | rot | rot |
| K7 | Pflicht-Attribute von `odata.replay.match` vertauscht | rot | grün (nicht Gegenstand der Probe) |
| K8 | `Millis` auf ganze Millisekunden | rot | rot |
| K9 | Schleife über `SpanRecord` ab Position 1 | rot | rot |
| K10 | überspringt Position 3 | **grün** | rot |
| K11 | überspringt Position 4 | **grün** | rot |
| K12 | überspringt Position 0 | rot | rot |
| K13 | überspringt jede zweite | rot | rot |
| K14 | überspringt Position 2 | rot | rot |
| X1 | `StartSpan` des Fakes kopiert `attrs` nicht | rot | grün (Probe ändert `attrs` nicht nach dem Aufruf) |
| X2 | `SetzeStand` kopiert nicht | rot | rot |
| X3 | `Snapshot` des Fakes kopiert nicht | rot | rot |
| X4 | `Anzahl` untere Grenze `< 1` | grün | grün (äquivalent) |
| X5 | `CountError` des Fakes maskiert hohe Bits des Codes | **grün** | rot |
| X6 | `Log` des Fakes hält `Background` statt des `ctx` | rot | rot |
| X7 | `Aufrufe()` teilt `Attrs` | rot | grün (Probe ändert die Rückgabe nicht) |
| X8 | `StartSpan` des Fakes: bei bereits abgebrochenem `ctx` ohne Abbruch | **grün** | grün |
| X9 | `CountError` des Fakes ignoriert Codes außerhalb der Tabelle | rot | rot |
| X10 | `Log` des Fakes ignoriert leeren `msg` | **grün** | grün |
| X11 | `Set` des Fakes ignoriert leere Attrs | rot | rot |
| X12 | `StartSpan` des Fakes ersetzt leeren Namen | **grün** | grün |
| X13 | `Log` des Fakes hebt Stufe unter Debug an | **grün** | grün |
| I1 bis I4 | Importe außerhalb der Menge (§6, R-5) | rot (alle 4) | - |

Aus den grünen Zeilen folgen V-2 (K3, K10, K11, K1), V-3 (X10, X12, X13, X5 und die Raten V2, V3) und V-5 (R6, T1, X8).

## 8. Befunde

| ID | Schwere | Befund | Zuständig |
|---|---|---|---|
| V-1 | MEDIUM | **Fenster der Sperre in `Snapshot` ohne Fänger und ohne Negativ.** `SPEC-230` Nr. 9 sagt "einen Stand zu einem Zeitpunkt", P15 sagt "`Log` und `Snapshot` sind unter der Sperre". Die Mutation W5 (`Errors` unter einer Sperre kopiert, freigegeben, `Durations` und `Dropped` unter einer zweiten) bleibt in jeder Spalte der Suite grün: Detektor 0 von 50, ohne Detektor 0 von 50, `-count=50` grün (der Detektor sieht nichts, weil jeder Zugriff unter einer Sperre liegt; `TestAdapterGleichzeitig` prüft nur, dass kein Zähler fällt). Der Wirkungs-Fänger aus §5 fängt sie in 30 von 30 Prozessen (Gegenprobe 0 von 30), ist also leicht möglich. Das Fenster steht weder in der Tabelle noch als Negativ im Plan (`AGENTS.md` §3.10: "je Fenster der Sperre eine Zeile mit Fänger oder benanntem Negativ"; `SPEC-232` Ungeprüft nennt nur die Sperre von Nr. 5 b für Ring und Verteilung, nicht Errors gegen Dropped). Klasse: Fenster der Sperre x Operation (`CountError` 2 Zweige, `verwerfen`, `Snapshot` vor und nach `Kopie()`, `Snapshot` zwischen den Sammlungen). | Implementer (Fänger, Vorlage §5) oder Architect (benanntes Negativ in `SPEC-232` Ungeprüft) |
| V-2 | LOW | **Fix R-1 deckt die Klasse "Position x Zustand" nur diagonal.** `kopieStand` setzt für Verteilung `i` und `SpanRecord` `i` den Zustand `i mod 3`; gefüllt ist nur Position 2 und 5. Grün in der Suite: K3 (Position 1 übersprungen), K10 (Position 3), K11 (Position 4), K1 (`Counts` mit einem Element). Rot dagegen: jede natürliche Schleifen-Abweichung (K9 ab Position 1, K12 Position 0, K13 jede zweite, K14 Position 2) und alle Mutationen der Reviewer. Der Kommentar von `kopieStand` sagt "so dass jede Position beider Zustände mehrfach vorkommt" (`AGENTS.md` §3.11: weiter als der Test; jeder Zustand kommt zweimal vor, nicht jede Position). Mein Probe (486 Stände, Zustände unabhängig je Position) besteht und fängt alle vier. Fix: zweiter Durchlauf mit verschobenem Zustand `(i+1) mod 3` oder 18 Einträge; der Kommentar sagt, was er prüft. | Implementer |
| V-3 | LOW | **"Sicher für gleichzeitigen Aufruf" und "jedes Argument ungeprüft" (`SPEC-230` Nr. 10) haben Fänger nur für einen Teil der Menge.** (a) Lesen: `Anzahl` ohne Sperre (V2) und `Aufrufe()` ohne Sperre (V3) bleiben in jeder Spalte grün (0 von 50 Detektor, 0 von 50 ohne, `-count=50` grün), weil die Tests beide erst nach `wg.Wait()` rufen. (b) Argumente: Fänger für `attrs` (außerhalb der Menge, Art 99, leerer Schlüssel) und `Code` (0); grün bleiben X10 (`Log` ignoriert leeren `msg`), X12 (`StartSpan` ersetzt leeren Namen), X13 (`Log` hebt Stufe unter Debug an), X5 (hohe Bits des Codes). Der Code des Fakes ist konform (Probe 40 Saaten grün); es ist die Lücke im Fänger. | Implementer |
| V-4 | LOW | **Zwei Rang-Zeiger lösen nicht auf.** Kommentare nennen `P18` (`TestFakeSpanJeOperation`) und `P19` (`TestLogOhneMinLevel`); der Plan führt P01 bis P17 und hat beide Fänger in P11 und P17 gefaltet (§6, letzter Punkt). `AGENTS.md` §3.7: ein Rang-Zeiger muss auf eine vorhandene Zeile zeigen. | Implementer |
| V-5 | INFO | Drei grüne Mutationen ohne Zusageverletzung im Plan: R6 (Status auf 32 Bit verengt; die Fixtures `MaxInt64` und `MinInt64` fallen durch die Verengung aus der Form, `100+2^32` wäre der Vertreter), T1 (`Log` mit drei oder mehr Verwürfen je Aufruf; P17 hat höchstens zwei) und X8 (Abbruch des `ctx` nur "nach dem Start" getestet, nicht "schon abgebrochen am Start"). Natürliche Defekte der Nachbarschaft werden gefangen (R7 rot). | Implementer (optional) |
| V-6 | INFO | Der Commit `ed3142b` ändert den Re-Review-Report einer anderen Rolle (der Re-Review-Report oben, ein Link auf [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) wegen des Doku-Gates). Inhalt unverändert, aber ein Artefakt des Reviewers. | Planner (Kenntnis) |
| V-7 | INFO | Gelesen, nicht gefahren: `new(telemetry.Adapter)` oder `var a telemetry.Adapter` hat nil-Maps; `CountError` mit einem Code außerhalb schreibt in `a.dropped` und gerät in Panik, obwohl `SPEC-230` einleitend "keine Operation gerät in Panik" sagt. Erreichbar nur an `New` vorbei; der Konstruktor ist der einzige Weg im Produkt; ob der Nullwert unterstützt wird, steht nirgends. | Architect (Kenntnis) |
| V-8 | INFO | Hinweis für `slice-telemetry-spans`: `filtern(op, ...)` mit `op == ""` ist die Menge `Log` (`opLog = ""`); ein Span mit leerem Namen darf nie bis dorthin kommen. | Planner (Adresse `slice-telemetry-spans`, `AGENTS.md` §3.13) |

Nicht doppelt: F-10 (`TestKernSchreibtNichts`, Adresse Planner, im Plan §6 mit Ausgang *offen*), Zeitbudget der Gegenproben ([ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md); Zusatzzeit: das Paket `telemetry` unter `-race` brauchte 25,7 s bei Load Average um 10, der Plan nennt 10,7 s ohne Detektor), Stempel.

## 9. Gegengelesen

Rohbytes: dieser Report enthält keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP (`LC_ALL=C grep -P` auf die Datei dieses Reports vor dem Commit). Die Probe-Dateien lagen außerhalb des Repos und sind nicht committet.

## 10. Closure-Vorbedingungen (nur genannt)

- DoD-Haken im Plan sind alle offen (Planner, nach den Fixes).
- Plan §7: Closure-Notiz mit Steering-Loop-Lerneintrag und Messpunkt-Zeile *Ende* (`SPEC-128` Nr. 9) fehlen ("Wird bei der Closure gefüllt."). Das Gate meldet `messung=2222`; Auslastung 0,74 der Marke, 1,03 der Schätzung: innerhalb 0,60 bis 1,00 der Marke, keine Meldung an den Architect nötig.
- §6: Risiken mit Ausgang *weiter offen* und Adresse (Planner/User): Frage an den User zum Abruf, Nehmer `slice-inspection-use-case`, Verdrahtung `session`/`maintenance`, Hybrid-Fallback, Auftrag `slice-lasttest`, Auftrag `SPEC-121` Nr. 12, Auftrag Handbuch, Abhängigkeiten der Einbau-Slices, `AGENTS.md` §3.14 gegen `SPEC-128`, Nachzüge des Integrationsschritts, F-10. Dazu aus diesem Report: V-1 (je nach Entscheidung eine Zeile in §6), V-8 (Adresse `slice-telemetry-spans`).
- Diff-Vorlage zu `AGENTS.md` §3.14 an den User (Planner, nicht dieser Zweig).
- Drei Paarungen (Anker, Folge-Slice, Register): Closure.

## 11. Register-Kandidaten (genannt, nicht geschrieben)

- `BEO-REPO/fixture-diagonale-statt-produkt`: eine Fixture belegt Position und Zustand reihum und deckt damit Zustände, nicht Position x Zustand (V-2). Zähler 1x.
- `BEO-REPO/sperrfenster-zwischen-sammlungen-ohne-faenger`: ein Stand aus mehreren Sammlungen unter einer Sperre hat ein Fenster zwischen den Sammlungen, das weder Detektor noch Suite sieht (V-1). Verwandt: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`; ein weiterer Beleg dort statt eines neuen Eintrags ist denkbar.
- `BEO-REPO/universalaussage-im-vertrag-ungeprueft`: ein weiterer Beleg "sicher für gleichzeitigen Aufruf" und "jedes Argument ungeprüft" mit Fängern nur für einen Teil der Operationen und Argumente (V-3).
- `BEO-REPO/rang-zeiger-auf-gefaltete-zeile`: ein Kommentar nennt eine Zeile, die der Plan nach einer Fix-Runde gefaltet hat (V-4, `AGENTS.md` §3.7). Zähler 1x.

## 12. Ungelaufen

- Rate der Fake-Mutationen V4 (Fenster in der ID-Zählung, R-7) und V6 (Fenster beim Lesen von `gesetzt`, äquivalent erwartet): in der Schlange, ohne Ergebnis bei Abgabe.
- `Snapshot` des Fakes ohne Sperre (R-2) selbst nicht gemessen; der Fake-Block misst nur `SetzeStand` ohne Sperre (V5).
- Zellen einzeln: je Mutation eine Zeile mit drei Zellen in einem Aufruf (§5).
- Mutationen an `Log`/`CountError` in Kombination mit `Clock`-Aufrufen (`SPEC-232` Nr. 5 f) gehören zu L und S; die Seite von P hat keinen `Clock.Now()`-Aufruf (gelesen).
- Das vollständige Gate wurde mit Docker-Schichten aus dem Cache grün gemeldet; ein `--no-cache`-Lauf von `make gates` ist nicht gefahren (eigene Läufe der Pakete und der Probes stehen in §1).
- `make docs-check` auf dieser Datei: siehe Übergabe (lief in der Schlange, Ergebnis nicht Teil dieses Textes).

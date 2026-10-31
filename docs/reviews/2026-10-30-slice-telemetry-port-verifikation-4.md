# Verifikations-Report 4: slice-telemetry-port (Teil P der Kette P, L, S), gezielte Nachverifikation des Deltas nach Fix-Runde 5 - 2026-10-31

**Rolle:** Verifier (Modul 11), vierte Runde, frischer Kontext, gezielt auf das Delta `3dac9de..4668d31`. Nicht Gegenstand: V-14 (gelöst, Planner), V-22 (INFO), die Closure.

**Gegenstand:** `slice-telemetry-port`, Zweig `slice-telemetry-port`, HEAD `4668d31` (Fix-Runde 5). Plan: [slice-telemetry-port](../plan/planning/in-progress/slice-telemetry-port.md); Vorbericht: [Verifikation 3](2026-10-30-slice-telemetry-port-verifikation-3.md).

**Gegen was geprüft:** `AGENTS.md` §3.10, §3.11, §3.14, §3.15, §3.17; `SPEC-230` Nr. 1 bis 4, 8 bis 10 und `SPEC-232` Nr. 4 in `spec/spezifikation.md`; der Vertrag [SPEC-128](../../harness/vertraege/SPEC-128-groessenmarke.md) (Nr. 3 n, Nr. 5 d, e); der Diff `git diff 3dac9de..4668d31` (5 Dateien: 4 Testdateien, der Plan).

**Modell:** claude-sonnet-5-5. **Arbeitsweise:** Gate und Stufen selbst gefahren; Mutationen und Probe-Tests nur auf Kopien im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-telemetrie-4/`), Image `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, jede Gruppe unter `flock /tmp/odata-recorder-heavy.lock`. Im Repo geändert ist nur dieser Report. Kein Container, keine wartende Sperre blieb zurück.

---

## Verdikt

**NICHT BESTANDEN** - wegen eines MEDIUM (V-23): V-15 a ist nicht gegen die Klasse geschlossen. Die Kommentar-Zusage "Abbruch, Frist und Werte bleiben" hat drei Elemente, die der Fix aus dem Zustand "Eingabe-ctx ist beim Aufruf schon abgebrochen oder abgelaufen" nicht fängt; drei Mutationen an Stellen, die weder Report 3 noch der Implementer nannten, bleiben grün (§3).

Im Übrigen ist das Delta tragfähig: V-16, V-17, V-18, V-19 und V-5 sind gegen ihre Klassen geschlossen (§2: 36 eigene Mutationen, 22 rot, 4 grün an den Grenzen des Plans, 10 grün gegen Klassen, die zu V-23 bis V-25 gehören); V-15 b habe ich am Test gelesen, nicht neu mutiert. Der Zeit-Fänger ist bei Last stabil (0 von 400). Die Messung 3020 ist nachgerechnet, der Eintrag `ueberschreitung: 3100` formgerecht, die Marke 3000 unberührt. Produktcode unverändert.

Befunde: 0 HIGH, 1 MEDIUM, 2 LOW, 2 INFO.

## 1. Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` im Worktree | **Exit 0** (Exit-Code in Datei geschrieben, aus der Datei gelesen, nicht über `tail`); Load Average vor dem Lauf 7,70, danach 5,53 | `gates.exit`, `gates.log` im Scratchpad: a-check `gesamt: 0 Befund(e)`, d-check `371 Datei(en) geprüft, 0 Befund(e)`, `migration-check: ok` |
| `make groessenmarke` (im Gate) | `groessenmarke slice-telemetry-port.md: ok, Messung 3020 von Marke 3000`; `Messpunkt ... zeilen=17 nebenlaeufig=2 fakes=1 produktcode=580 schaetzung=2150 marke=3000 messung=3020` | Gate-Ausgabe |
| Cache im Gate | Die Schichten von `test`, `race`, `lint` standen als `CACHED`; darum einzeln ohne Cache (`--no-cache-filter`, Argumente wie im Makefile) | `stage-*.log` |
| Stufe `test` ohne Cache | Exit 0, `pruefung test: ok` (79,5 s) | `stage-test.log` |
| Stufe `race` ohne Cache | Exit 0, `pruefung race: ok` (110,5 s) | `stage-race.log` |
| Stufe `lint` ohne Cache | Exit 0, `0 issues.` | `stage-lint.log` |
| Last der drei Stufen | Load Average 6,99 vor, 8,94 nach dem Lauf | `stages.load` |
| Produktcode unverändert | `git diff --name-only 3dac9de..4668d31` nennt 4 Dateien `*_test.go` und den Plan; keine Datei unter `src/` ohne `_test.go`, nichts unter `spec/`, `harness/`, `go.mod` | `git diff` |
| `make docs-check` auf diesem Report | Exit 0, `d-check: 372 Datei(en) geprüft, 0 Befund(e)` | `docs.log` |

Nicht gefahren, wie beauftragt: `make gegenproben`, Stempel.

## 2. Klasse gegen Klasse: eigene Mutationen

Mengen aus dem Wortlaut der Kommentare und von `SPEC-230`/`SPEC-232` gebildet, Mutationen an Elementen, die weder Report 3 noch Fix-Runde 5 nannten. Läufe: `go test -count=1` über `ports/support`, `supporttest`, `adapters/driven/telemetry`, ohne Detektor (die Fänger sind deterministisch), Load Average bei den Läufen 3 bis 10. Jede grüne Mutation habe ich mit einem Probe-Test gegengeprüft: die Probe ist auf dem unveränderten Stand grün und unter der Mutation rot, die Mutation ist also nicht äquivalent.

| Id | Klasse (Zusage) | Mutation | bestehende Suite | Probe |
|---|---|---|---|---|
| F04, F05, F06 | Fake: Slices kopiert (`SPEC-230` Nr. 10), je Aufrufstelle | `Log`, `StartSpan` bzw. `Set` ohne `slices.Clone` | rot (`TestFakeAufzeichnung`) | - |
| F07 | `Aufrufe()` "als Kopie" | innerer `Attrs`-Clone entfernt | rot (`TestFakeAufzeichnung`) | - |
| F09 | `SetzeStand`/`Snapshot` "nil oder leer, wie gesetzt" | `gesetzt` nur, wenn eine Sammlung nicht nil | rot (`TestFakeSnapshotStand`) | - |
| F12 | Aufzeichnung vollständig | `Aufrufe()` liefert nur die letzten 1000 | rot (`TestFakeLesenGleichzeitig`, `TestFakeGleichzeitig`) | - |
| C1, C4 | ctx "dieselbe Identität", `StartSpan` | aufgezeichnet wird der abgeleitete bzw. der auf `Background` ersetzte ctx | rot (`TestFakeAufzeichnung`/`TestFakeArgumenteUngeprueft`/`TestFakeEltern`) | - |
| C2, C3 | ctx "dieselbe Identität", `Log` | aufgezeichnet wird ein Abgeleiteter bzw. nil als `Background` | rot (`TestFakeArgumenteUngeprueft`) | - |
| L37, L64, S63, T50 | Attributlisten jeder Länge (`TestFakeLaengen`), je Operation | `Log` bei Länge 37 bzw. 64, `Set` bei 63, `StartSpan` bei 50 ohne Attribute | rot (`TestFakeLaengen`) | - |
| L65 | dasselbe, Länge 65 | `Log` bei Länge 65 | **grün** (N-2, Grenze wie im Plan) | - |
| E12, E16 | Eltern-Kette Tiefe 1 bis 16 | Eltern nil bei der 12. bzw. 16. ID | rot (`TestFakeEltern`) | - |
| E17 | dasselbe, Tiefe 17 | Eltern nil bei der 17. ID | **grün** (N-2, Grenze wie im Plan) | - |
| P05, P06 | `Kopie`: nil bleibt nil, nested | `Attrs` bzw. `Counts` nil zu leer | rot (`TestSnapshotKopie`) | - |
| P07, P08 | `Kopie`: leer bleibt leer, nested | `Counts` bzw. `Attrs` leer zu nil | rot (`TestSnapshotKopie`) | - |
| KC9, KA11 | `Kopie` tief, Längen 1 bis 14 | `Counts` der Länge 9 bzw. `Attrs` der Länge 11 geteilt | rot (`TestSnapshotKopie`) | - |
| V40 | "jedes verletzende Attribut zählt" | `Log` zählt höchstens 40 Verwürfe | **grün** (der Fänger endet bei 40; ein Deckel bei 40 ist dort unsichtbar, ein Deckel bei 39 rot laut Plan, bei 41 grün: A19e) | - |
| V41 | dasselbe | höchstens 41 | **grün** (N-2) | - |
| **F01** | `StartSpan`-ctx: Frist bleibt, Eingabe-ctx **schon abgelaufen** | ein abgelaufener Eingabe-ctx wird durch `Background` ersetzt | **grün** | rot (`TestProbeAbgelaufenVorher`) |
| **F02** | Abbruch bleibt, Eingabe-ctx **schon abgebrochen mit Ursache** | `Err` bleibt (`Canceled`), Ursache und Frist gehen verloren | **grün** | rot (`TestProbeUrsacheVorher`, `TestProbeAbgelaufenVorher`) |
| **F03** | Werte bleiben, Eingabe-ctx **schon abgebrochen** | Werte gehen beim abgebrochenen ctx verloren, `Err` und Ursache bleiben | **grün** | rot (`TestProbeWertNachAbbruch`) |
| **F08** | `Snapshot` "als Kopie", Fake **ohne** gesetzten Stand | der leere Stand ist eine Paketvariable, geteilt zwischen Aufrufen und Fakes | **grün** | rot (`TestProbeLeerGeteilt`) |
| **P04** | `Kopie` "tief", jede Position von `Spans` | `Attrs` nur der ersten 6 Einträge kopiert | **grün** | rot (`TestProbePositionen`, 7, 8, 20 und 1024 Einträge) |
| **F10, F11, F13** | "jedes Argument unverändert" (`SPEC-230` Nr. 10), Länge der Zeichenketten | `Log` kürzt `msg` auf 128 Bytes, `Text` auf 40 Bytes; `StartSpan` den Namen auf 40 Bytes | **grün** | rot (`TestProbeStringLaengen`, `TestProbeLogMsgText`) |
| **A01** | "kein Wert eines Aufrufers bildet einen Schlüssel" (`SPEC-230` Nr. 8), Art `Float` | `Log` legt für jedes `Float`-Attribut einen Eintrag in `Dropped` an | **grün** | rot (`TestProbeReihenFloat`) |
| **A02** | dasselbe, Codes zwischen 101 und 99998 | `CountError` legt für einen solchen unbekannten Code einen Eintrag in `Errors` an | **grün** | rot (`TestProbeReihenCodes`) |

Zusammen 36 Mutationen: 22 rot (F04 bis F07, F09, F12, C1 bis C4, L37, L64, S63, T50, E12, E16, P05 bis P08, KC9, KA11), 4 grün an den Grenzen, die der Plan als N-2 benennt (L65, E17, V40, V41), 10 grün gegen Klassen (F01, F02, F03, F08, P04, F10, F11, F13, A01, A02; V-23 bis V-25).

**Fix-Runde 5 gegen ihre Klassen:** V-16 (Sammlungen x Richtung, Fake: F04 bis F07, F09), V-17 (Längen 1 bis 14 von `Counts` und `Attrs`, nested nil und leer: KC9, KA11, P05 bis P08), V-18 (Länge je Operation, Tiefe, Identität des ctx: L37, L64, S63, T50, E12, E16, C1 bis C4) und V-19 (Reihen; Ausnahme A01, A02 in V-25) halten gegen 22 rote Mutationen, davon achtzehn an Elementen, die weder Report 3 noch der Bericht des Implementers nannten (F04 bis F07, C1 bis C4, L37, S63, T50, E12, KC9, KA11, P05 bis P08). Die Grenzen des Plans stimmen auf die Id genau: Länge 64 und Tiefe 16 rot, 65 und 17 grün; ein Deckel bei 40 Verwürfen ist im Fänger bis 40 nicht sichtbar (V40 grün), V-5 deckt Deckel bis 39 (Plan: A19d rot).

## 3. V-15 a: die Klasse "Abbruch, Frist und Werte bleiben" ist nicht geschlossen (V-23)

**Menge (aus dem Wortlaut des Kommentars von `StartSpan` und `SPEC-230` Nr. 6, "Abbruch, Frist und Werte bleiben"):** Eigenschaft des Eingabe-ctx (Werte, Abbruch mit Fehler, Abbruch mit Ursache, Frist als Zeitpunkt) x Zustand des Eingabe-ctx beim Aufruf (aktiv; schon abgebrochen; schon abgelaufen) x Aufrufstelle (ohne und mit Span im ctx). `TestFakeCtx` führt 6 Eigenschaften x 2 Stellen; die Eigenschaft "Abbruch vor dem Aufruf" prüft allein `Err() != nil`. Ungeprüft bleiben aus der Menge: Werte, Ursache und Frist eines beim Aufruf schon abgebrochenen oder schon abgelaufenen ctx. Drei Mutationen (F01 bis F03) bleiben grün, jede mit Probe rot gesehen. Das ist dieselbe Form wie der Befund von Report 3 (K6: Frist geht verloren, Rest bleibt) und das `AGENTS.md` §3.17 "Falsch"-Beispiel (der Fix macht die genannte Mutation rot, benachbarte Elemente bleiben).

## 4. N-2 (Obergrenzen K4e, K3e, A19e)

- **Grund stimmt:** `filtern` (Schleife über `attrs`), `Kopie` (`maps.Clone`, `slices.Clone`) und der Fake haben keine Fallunterscheidung nach der Länge, gelesen im Code; die Menge ist unbeschränkt.
- **Folge im Betrieb stimmt:** kein Dienst übergibt mehr als 13 Schlüssel an `Log` (`SPEC-230` Nr. 3), ein Span trägt höchstens 3 Pflicht-Attribute, `Counts` hat 12, der Ring höchstens 3 Attribute je Span; die Tiefe der Spans im Produkt ist drei (Gesamtspan, `match`, `query`). Die Folge "ein Dienst-Test mit mehr als 64 Attributen oder einer Kette über 16 Spans würde nicht gefangen" ist die einzige reale Wirkung.
- **Grenzen:** 64, 16, 40 und 14 sind nicht aus dem Vertrag hergeleitet, sondern Vielfache der Betriebsmaxima (64 etwa das Fünffache von 13, 16 das Fünffache von 3, 14 gleich 12 plus 2); der Plan nennt das nicht, der Verifier hält es für tragfähig, weil die Grenze in jeder Id auf die Stelle genau gemessen ist (§2). Eine Herleitung im Plan wäre besser, ist aber kein Befund.
- **Lücke:** N-2 nennt nur Längen, Tiefe und Zahl der Verwürfe. Weitere unbeschränkte Dimensionen der geprüften Mengen stehen nicht darin: V-25.

## 5. Messung nach [SPEC-128](../../harness/vertraege/SPEC-128-groessenmarke.md)

- **Metrik:** `git diff --numstat a0396b0cb0769476edaeb5bfaa23dcc5a76c041c HEAD -- src/hexagon/ports/support src/adapters/driven/telemetry`: Einfügungen 3019, Löschungen 1, Summe **3020** (gleich dem Gate, gleich dem Plan). Anteil des Deltas `3dac9de..4668d31` an diesen Pfaden: 429 Zeilen (Summe aus Einfügungen und Löschungen).
- **Form des Eintrags:** `ueberschreitung: 3100 | <Grund>` im Block, getrennt am ersten `|` (Nr. 3 n); N gleich 3100 = Messung 3020 aufgerundet auf volle Hundert (Nr. 5 d), N größer oder gleich Messung, Grund nicht leer (Nr. 5 e). Das Gate meldet `ok`, keinen Befund.
- **Grund trägt die Überschreitung:** Er nennt Ursache (sechs Befunde der Verifikation 3 verlangen Fänger über die ganze Menge als Schleifen), die betroffenen Tests und dass der Produktcode unverändert blieb. Das passt zum Diff: 429 Zeilen im Delta, davon 401 in vier Testdateien und 28 Löschungen im Delta; kein Produktcode. Die Überschreitung ist 20 Zeilen über der Marke und 1,40 der Schätzung; der Grund erklärt die Überschreitung, nicht den Abstand zur Schätzung (die Formel schätzt den Stand beim Handoff, `AGENTS.md` §3.14).
- **Marke:** unverändert 3000 (`git diff 3dac9de..4668d31` hat keine Änderungszeile `marke:` im Block); bestätigt. Marke gleich Formel (Nr. 5 b): 100 x ceil(135 x 2150 / 10000) = 3000.

## 6. Zeit-Fänger `TestFakeCtx` (Frist 20 ms, Wartegrenze 2 s)

Test-Binary einmal gebaut (`go test -c`), je Prozess `-test.run '^TestFakeCtx$' -test.count=1`, im Container `odata-recorder:race`, keine Mutation. Messender: der Verifier. Load Average (1 Minute) am Ende der Gruppe aus `/proc/loadavg`, Host-Last anderer Linien lief zugleich.

| Gruppe | Prozesse | rot | Load Average am Ende |
|---|---|---|---|
| ohne Zusatzlast (`--cpus=4`) | 100 | **0** | 3,81 |
| 8 Rechenschleifen im selben Container (`--cpus=4`) | 100 | **0** | 3,39 |
| 24 Rechenschleifen im selben Container (`--cpus=4`) | 100 | **0** | 9,85 |
| 24 Rechenschleifen (`--cpus=1`) | 100 | **0** | 9,74 |

Rate ohne Mutation 0 von 400. Die Rate unter Mutation habe ich für die Zeitzeile nicht gemessen: der Test wird nur durch das Verhalten rot (Ablauf der Frist erreicht den Rückgabe-ctx), die Wartegrenze 2 s ist 100 Mal die Frist. Über Load Average 10 und über 24 Schleifen habe ich nicht gemessen. Kein Befund.

## 7. Plan-Zahlen gegen Gate und Tests

| Aussage des Plans | Prüfung | Ergebnis |
|---|---|---|
| Tabelle P01 bis P17 hat 17 Zeilen, `zeilen=17` | gezählt (Zeilen `| P01 |` bis `| P17 |` der Zusage-Tabelle), Gate | stimmt |
| Messung 3020, Auslastung 3020/3000 = 1,01 und 3020/2150 = 1,40 | nachgerechnet | stimmt (1,0067; 1,4047) |
| 52 rote Mutationen und 3 grüne Negative in den Raten | die Tabelle der Fix-Runde 5 gezählt: 6 + 6 + 12 + 8 + 3 + 3 + 3 + 5 + 6 = 52; Negative K4e, K3e, A19e | stimmt |
| Tests: 12 Fälle `TestFakeCtx`, 12 `TestFakeStandKopie`, 195 `TestFakeLaengen`, 115 `TestReihenFest`, 123 `TestVerwurfJeAttribut`, 96 Stände `TestSnapshotKopie` | gelesen und gerechnet (6 x 2; 6 x 2; 3 x 65; 7 + 25 x 4 + 8; 41 x 3; 16 Zustände x 6 Positionen) | stimmen, die Zählprüfungen im Test tragen dieselben Zahlen |
| §6: V-5 mit Ausgang *entfallen*, Fänger R6, T1, X8 | gelesen; die Fänger (`TestWertformGrenzen`, `TestVerwurfJeAttribut`, `TestFakeCtx`) laufen grün | stimmt; die Zusage von X8 (Abbruch vor dem Aufruf) trägt aber nur `Err`, siehe V-23 |
| V-21 bereinigt | der Satz "Load Average 4,7 bis 9,2" ist durch den Hinweis auf das ersetzte Orakel ersetzt (`git diff`) | stimmt |
| Zeilen der Chronik der Fix-Runde 3 | Zeile V-2 der Tabelle der Fix-Runde 3 nennt noch "je 5 Zustände ... 30 Stände" neben der geltenden Zeile P09 (16 Zustände, 96 Stände) | INFO V-26 |

## 8. Befunde

Zuständigkeit: Implementer = Test oder Code im Slice; Planner = Plan und Adresse.

| Nr. | Schwere | Zuständig | Befund |
|---|---|---|---|
| **V-23** | **MEDIUM** | Implementer | **V-15 a nicht gegen die Klasse geschlossen (`AGENTS.md` §3.11, §3.17).** Der Kommentar von `FakeTelemetry.StartSpan` ("Abbruch, Frist und Werte bleiben") und `SPEC-230` Nr. 6 nennen drei Eigenschaften des Eingabe-ctx; die Menge Eigenschaft x Zustand beim Aufruf (aktiv, schon abgebrochen mit Ursache, schon abgelaufen) hat keinen Fänger außerhalb von "aktiv" und "schon abgebrochen: `Err`". Grün bleiben F01 (abgelaufener ctx wird ersetzt), F02 (Ursache und Frist eines abgebrochenen ctx gehen verloren), F03 (Werte eines abgebrochenen ctx gehen verloren); Proben rot. Fix: `TestFakeCtx` als Schleife Zustand (3) x Eigenschaft (Werte, Fehler, Ursache, Frist) x Stelle (2) mit Orakel aus dem Eingabe-ctx; oder der Kommentar sagt weniger. |
| V-24 | LOW | Implementer | **Fake ohne gesetzten Stand teilt den leeren Stand (F08).** `SPEC-230` Nr. 10 sagt "den leeren Stand: vier leere, nicht-nil Sammlungen" und Nr. 9 "Der Stand ist eine Kopie". `TestFakeStandKopie` ändert nur Stände, die `SetzeStand` gesetzt hat; der Zweig `!gesetzt` hat keinen Fall für "Änderung am gelesenen leeren Stand wirkt auf den nächsten Snapshot oder einen anderen Fake". Fix: eine Zeile je Richtung im selben Test, oder N-2-Eintrag. |
| V-25 | LOW | Planner / Implementer | **N-2 benennt nicht alle unbeschränkten Dimensionen der Fänger (`AGENTS.md` §3.17).** Grün bleiben: P04 (nur die ersten 6 Positionen von `Spans` werden tief kopiert; der Ring hält bis 65536 Einträge, die Fänger führen höchstens 6), F10, F11, F13 (Zeichenketten über 40 bzw. 128 Bytes im Fake: `Msg`, `Text`, `Name`), A01 (Art `Float` in `TestReihenFest`: nur `String` und `Int`), A02 (unbekannte Codes zwischen 101 und 99998). Folge im Betrieb: keine, wenn der Code die Länge, Position und Art nicht verzweigt (gelesen: keine Verzweigung); ein Fehler, der erst ab diesen Größen oder bei der Art `Float` greift, würde nicht gefangen. Entweder Fall je Dimension (ein Eintrag mit 7 und mit 1024 Positionen, eine lange Zeichenkette je Feld, ein `Float`-Attribut in P07) oder N-2 nennt sie mit Grund und Folge. |
| V-26 | INFO | Planner | **Zeile V-2 der Chronik der Fix-Runde 3 nennt "30 Stände" und "5 Zustände" neben der geltenden Zeile P09 (96 Stände, 16 Zustände).** Die Chronik gehört zum ersetzten Orakel; der Planner kann sie in der Closure entfernen oder als ersetzt kennzeichnen (gleiche Form wie V-21). Der Satz zu V-14 in §6 ("bis 5,4") ist Auftrag des Planners und nicht Gegenstand dieser Runde. |
| V-27 | INFO | Planner | **Die Grenzen 64, 16, 40 und 14 haben keine Herleitung im Plan** (§4 dieses Reports): sie sind Vielfache der Betriebsmaxima, jede auf die Stelle genau gemessen. Eine Zeile mit der Herleitung (13 Schlüssel, 3 Attribute je Span, Tiefe 3, `Counts` 12) macht N-2 prüfbar. Kein Fix nötig. |

Gelöst gegenüber Report 3 (kein Befund mehr): V-15 b, V-16, V-17, V-18 (Längen, Tiefe, `Anzahl`), V-19, V-5; V-21 bereinigt.

## 9. Konformität gegen die DoD des Plans

| DoD-Punkt | Stand |
|---|---|
| 1 Ports, Typen, Fake | erfüllt bis auf V-23 (Fänger, nicht Verhalten) |
| 2 Kern des Adapters | erfüllt; `make test-race` ohne Befund (Stufe `race` ohne Cache) |
| 3 `make gates` grün | erfüllt (Exit 0; `test`, `race`, `lint` zusätzlich ohne Cache) |
| 4 Review, 5 Doku-Update, 6 Handbuch | wie in Report 3; das Delta ändert weder Produktcode noch Spezifikation |
| 7 bis 10 Closure-Notiz, Register, Ausgänge, Paarungen | Aufgabe des Planners nach dieser Verifikation |

Messpunkt-Zeile: Handoff-Wert 3020 nachgerechnet und gleich dem Gate; die Ende-Zeile setzt der Planner.

## 10. Ungeprüft und Gegengelesen

**Ungeprüft:** `make gegenproben` und `harness/gegenproben.stempel` (beauftragt nicht gefahren); Last über Load Average 10 und mehr als 24 Rechenschleifen beim Zeit-Fänger; die Raten der Mutationen F01 bis F13 mit mehreren Prozessen (die Fänger sind deterministisch, je Mutation ein Lauf); Mutationen an `support.Spans()` (V-15 b nur am Test gelesen: Schleife über alle Namen und Pflicht-Elemente, zweiter und dritter Aufruf verglichen) und an den Fenster-Fängern (`TestSnapshotZeitpunkt`), die Report 3 beschreibt; die Wertformen über die Stichproben der Vorrunden hinaus (R6, R6z nur aus dem Plan übernommen); Binary-Hash (Produktcode unverändert, nicht neu verglichen).

**Gegengelesen:** dieser Report auf Rohbytes (`LC_ALL=C grep -P`: keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP) und mit `make docs-check` im Worktree; `make docs-check` Exit 0, `d-check: 372 Datei(en) geprüft, 0 Befund(e)`; Rohbytes-Prüfung: 0 Treffer für Escape-Folge, BOM, NBSP.

## 11. Register-Kandidaten (genannt, nicht geschrieben)

- Klasse "Zustand des Arguments beim Aufruf" als Dimension der Menge (aktiv, schon abgebrochen, schon abgelaufen) neben der Eigenschaft: Fix deckt die Eigenschaften, nicht ihre Zustände (V-23); zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` gehörig.
- Klasse "unbeschränkte Dimension der Menge, die N-2 nicht nennt" (Position, Länge einer Zeichenkette, Art): jede Runde nennt eine weitere Dimension (V-17, V-18, V-25).

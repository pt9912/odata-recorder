# Verifikation slice-admin-abweisen-begrenzen (höchstens 20 Einträge `admin_abgewiesen` je Fenster von 60 s) — 2026-11-28

Verifier, frischer Kontext, genau eine Verifikation. **Modell:** Sonnet 5.5.
Stand HEAD `f755e65b`, Zweig und Worktree des Slice, vor der Prüfung `git status` leer; Basis der Marke `fd960822`.
Gegen: DoD des Plans (§2), Tabelle *Zusage · Fänger* (W01 bis W18), Fenster der Sperre, Mutationen, Raten, §6; `SPEC-301`, `SPEC-300` Nr. 5 und 12, `SPEC-041` Nr. 7, 9, 10; `LH-FA-13` samt `LH-FA-13.a`, `LH-QA-04`; `2026-11-28-slice-admin-abweisen-begrenzen-review`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

## Urteil

**DoD-Liefer-Punkte 1 bis 3, Gates und Review: bestätigt.** Kein HIGH, kein MEDIUM, zwei LOW, drei INFO. Zwei Auflagen für die Closure, keine blockiert den Merge.

Die Fix-Runde deckt die Klassen der Befunde F-1 bis F-3 des Reviews (eigene Mutationen an Elementen, die der Befund nicht nannte, rot gesehen); bei F-2 bleibt eine Prüfgrenze ungenannt (V-1).

## Auflagen für die Closure

1. **V-1:** die Prüfgrenze der Zusage „`Uhr.Now()` genau einmal“ in §6 als akzeptiertes Negativ nennen oder um eine Zelle (spät im Fenster gedrosselt) schließen.
2. **V-2:** die zwei fehlenden Zeilen in der Tabelle *Zusage · Fänger* nachtragen.

Dazu die üblichen Closure-Punkte, siehe V-4.

## Sensoren und Belege

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` | **nicht neu gefahren** (Auftrag): der Nachweis passt | `.harness/state/gates-passed.head` gleich HEAD `f755e65b`; `gates-passed.diffsha` gleich der Ausgabe von `tools/harness/working-tree-hash.sh` vor diesem Report (`6c39772f…`), geschrieben 02:20:39. Der Nachweis sagt nicht, welche Stufen aus dem Schicht-Cache kamen; deshalb die Zeilen darunter. `handbuch-check`, `katalog-check`, `test-security`, `lint`, `a-check`, `groessenmarke` sind allein durch den Nachweis belegt |
| Paket `admin`, `go test -count=1` ohne und mit `-race` | grün (27,8 s und 103,7 s) | eigener Lauf |
| `bootstrap` `TestAdminAbweisenBegrenzt`, `TestAdminAbweisung`, `TestAdminKonfig` mit `-race`; `tests/integration` `TestM6*` | grün | eigener Lauf |
| Handlauf am Binary der Stufe `build` aus `git archive HEAD` | siehe unten | eigener Lauf |
| Marke: `git diff --numstat fd960822..HEAD` über die Pfade des Blocks | 1151 von 2400 | nachgerechnet |
| `make gegenproben` | nicht gefahren (Auftrag) | — |

Läufe: je ein Container ohne Netz und ohne Mount, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`; Last des Hosts nach dem Warten 3,61 und 6,10, 14 GiB verfügbar.

## Handlauf am gebauten Binary

| Prüfung | Ergebnis |
|---|---|
| 25 Abweisungen (ohne Header, falsches Token, `Basic`, unbekannter Pfad, leerer Wert), Modus `record` | alle 401; 20 Zeilen in `audit.jsonl`, ihre IDs sind die der ersten 20 Antworten; Zähler `ODR-E2101` 25 |
| Antworten 1 bis 25 | ein Header-Satz (ohne `X-Correlation-Id` und `Date`: `Content-Length`, `Content-Type`, `Www-Authenticate`), ein Body, kein `Retry-After`, 25 verschiedene IDs |
| Fenster in Echtzeit, erste Abweisung 20 s nach dem Start | +41 s: 20 Zeilen; +59 s: 20; +60,3 s: 21; nach 25 weiteren 40; Zähler 53. Der Beginn ist die erste Abweisung, nicht der Start. Damit ist die Seite, die §6 Negativ 1 offen lässt, einmal von Hand belegt |
| gültiges Token und `GET /healthz` bei vollem Fenster | 200 |
| Neustart auf demselben `DATA_DIR` | 25 Abweisungen ergeben 20 weitere Zeilen (60) |
| 600 Abweisungen, 64 zugleich | 600-mal 401, genau 20 weitere Zeilen (80), jede Zeile ein Objekt, Zähler 600, nichts davon auf `stderr` |
| Modi `replay`, `mock`, `hybrid` | je 25-mal 401, 20 Zeilen, Zähler 25 |

**Handbuch wahr:** höchstens 20 Zeilen je 60 s ab der ersten dieser Zeilen, dieselbe Antwort ohne Zeile, der Zähler zählt jede Anfrage (am Binary gesehen); das erzeugte Token hat 64 Zeichen, aus 32 Bytes `crypto/rand` (`tokendatei.go`). Der Satz, je Anfrage ohne Token entstehe eine Zeile, steht nicht mehr da; eine zweite Stelle mit dieser Aussage gibt es nicht. Die Folge eines Rücksprungs der Systemzeit (Review F-5) steht offen in §6 beim Architect; die Obergrenze „höchstens“ bleibt auch dann wahr.

## Zeilen W01 bis W18: je eine eigene Mutation

21 Mutationen, alle rot am Fänger, den der Plan nennt; der Stand ohne Mutation danach grün.

| Zeile | Mutation | rot |
|---|---|---|
| W01 | nil-`Uhr` nicht geprüft | `TestNewUhr`, `TestNewNilFelder` |
| W02, W03 | Grenze 21; `>` statt `>=` | `TestFensterZahl`; `TestFensterGrenze` |
| W04, W05 | Beginn bei `New`; `IsZero` als „kein Fenster“ | `TestFensterBeginn`; `TestFensterUhrWerte` |
| W06 bis W08 | `Retry-After`; `CountError` bei gedrosselt entfällt; leerer `Append` | `TestGedrosseltAntwort`; `TestGedrosseltZaehlt`; `TestGedrosseltOhnePort` |
| W09 | 404 holt eine Zulassung | `TestFensterNurAbweisung` |
| W10 bis W12 | Sperre entfällt (ohne und mit Detektor); `Now()` vor der Sperre; Sperre über `Append` | `TestFensterGleichzeitig`; `TestFensterZulassungAtomar`; `TestFensterOhneSperreUeberAppend` |
| W13, W14 | Zählen erst nach erfolgreichem `Append`; Fenster als Paketvariable | `TestFensterAppendFehler`; `TestFensterJeServer` |
| W15 | angehaltener Server holt Zulassung; Timer im Fenster | `TestFensterAngehalten`; `TestFensterOhneZeitgeber` |
| W16 | achtes Feld in `Eintrag` | `TestFensterKeineNeuenArten` |
| W17 | Root ohne `Uhr`; zweite Instanz der Uhr | `TestAdminAbweisenBegrenzt`; `TestAdminKonfig` |
| W18 | Netzwerktest ohne Vorrücken; keine Grenze | `TestSicherheitKeineAenderung` |

## Fix-Runde gegen die Klassen (Elemente, die Review und Plan nicht nannten)

Ganzes Paket `admin` je Mutation.

| Klasse | Mutation | Ergebnis |
|---|---|---|
| F-1 Beginn bleibt | allein die 7. Zulassung setzt den Beginn | rot: `TestFensterBeginnBleibt`, `TestFensterReihenfolge` |
| F-1 | jede Zulassung schiebt um 1 ns; jede gedrosselte schiebt um 1 ns | rot, je sechs Tests |
| F-1 | ab dem zweiten Fenster setzt die 13. den Beginn | rot: `TestFensterBeginnBleibt`, `TestFensterReihenfolge` |
| F-1 | eine Zulassung später als 45 s setzt den Beginn | rot: `TestFensterReihenfolge` |
| F-2 eine Lesung | zweite Lesung allein bei der 5. Zulassung | rot: `TestFensterReihenfolge` |
| F-2 | zweite Lesung allein bei gedrosselten später als 30 s | **grün** (V-1) |
| F-2 | `Serve` ruft `Now()`; `Schliessen` ruft `Now()` | **grün** (V-3) |
| Rechnung, Sperre | Differenz über `UnixNano`; `TryLock` drosselt | rot |
| Sperre nie über fremde Aufrufe (`SPEC-301` Nr. 2) | Fenstersperre über `KorrelationsID`, `CountError`, `Vergleich`, `Handler` | rot, vier von vier (`TestFensterZulassungAtomar`; `TestBlockierterAufrufHaeltDieUebrigenNicht`) |
| Log bei Drosselung | `Telemetrie.Log` an der gedrosselten | rot |

**F-3 Fenster der Sperre**, rote Prozesse je Reihe (`-test.count=1`; Messender: Verifier; `--cpus=4`; ohne Detektor 20 Prozesse / mit Detektor 10). `TestFensterZulassungAtomar` ist in höchstens 1 Prozess je Reihe rot und steht nicht in der Tabelle.

| Teilung der Sperre | `TestFensterReihenfolge` | `TestFensterGleichzeitig` |
|---|---|---|
| Lesen – Vergleich (Review X4) | 20/20 / 10/10 | 0/20 / 0/10 |
| Vergleich – Öffnen | 20/20 / 10/10 | 19/20 / 10/10 |
| Öffnen – Prüfen | 20/20 / 10/10 | 0/20 / 0/10 |
| Prüfen – Erhöhen | 20/20 / 9/10 | 20/20 / 10/10 |
| im Öffnen: erst `offen` und Beginn, dann Zahl (neu) | 20/20 / 10/10 | 0/20 / 3/10 |
| im Öffnen: erst Zahl, dann `offen` und Beginn (neu) | 20/20 / 9/10 | 0/20 / 0/10 |

Die Raten der vier Fenster des Plans sind damit bestätigt. Die zwei feineren Teilungen fängt `TestFensterReihenfolge` ebenfalls. Die Grenze bei einem Prozessor steht in §6 Negativ 6; ich habe sie nicht nachgemessen.

F-4 ist behoben (§6 nennt die Zeile *Nachzug von `TestJedeAbweisungEinEintrag`*), F-5 steht offen beim Architect, F-6 steht als §6 Negativ 6.

## DoD gegen Beleg

| DoD-Punkt (§2) | Stand | Beleg |
|---|---|---|
| 1. Fenster und Verdrahtung (W01 bis W17) | erfüllt | Code: `abweisen_fenster.go`, eine Stelle in `schreibeFehler`, `Uhr: h.uhr` in `lebenszyklus.go`; Zeilenmutationen oben; Handlauf |
| 2. Nachzug der Zählung (W18; `SPEC-041` Nr. 7 c) | erfüllt | `TestSicherheitKeineAenderung` führt 1008 Einträge und die Kontrolle mit 20 Einträgen und 1008 `CountError`; zwei Mutationen rot; Paket `admin` grün; die angepassten Bestandstests halten ihre Erwartungen je zugelassener Abweisung |
| 3. Handbuch, `make handbuch-check` | erfüllt | Sätze gelesen und am Binary geprüft; der Check durch den Gate-Nachweis |
| `make gates` grün | erfüllt | Nachweis gegen HEAD |
| Review durchgeführt, Report liegt vor | erfüllt, mit Einschränkung | das Review deckt `fd960822..d5a86f0a`; die Fix-Runde `f755e65b` hat kein Reviewer gesehen und ist hier an Mutationen geprüft |
| Closure-Notiz, Register, Ausgang je Risiko, Paarungen | offen | Sache der Closure |

**Plan gegen Code:** die 15 geänderten Dateien unter `src/` stehen im Block (§4); `abweisen_gleichzeitig_test.go` und `fakes_test.go` stehen dort und sind unverändert, wie §3 sagt. Jeder in W01 bis W18 genannte Test existiert genau einmal. Keine Abweichung des Codes von §3 gefunden.

## Befunde

| ID | Kategorie | Befund | Quelle |
|---|---|---|---|
| V-1 | LOW | Eine zweite Lesung der Uhr allein an gedrosselten Abweisungen später als 30 s nach dem Beginn bleibt im ganzen Paket grün. `TestFensterEineLesung` führt gedrosselt bei Abstand 0 und vor dem Beginn; die Menge der Abstände ist unbeschränkt, und §6 nennt keine Prüfgrenze. LOW, weil die gemessene Form den Wert verwirft und am Verhalten nichts ändert. | `SPEC-301` Nr. 2 a; `AGENTS.md` §3.17 |
| V-2 | LOW | Die Tabelle führt für „Sperre nie über `CountError`, `IDs`, `Vergleich`, `Handler`“ (`SPEC-301` Nr. 2; nur `Append` steht als W12) und für die Geheimnisse (Nr. 7) keine Zeile. Fänger gibt es für den ersten Satz (vier Mutationen rot, siehe oben); zu Nr. 7 habe ich keine Mutation gefahren, die Tests der Geheimnisse sind im Paket grün. | `AGENTS.md` §3.15 |
| V-3 | INFO | `Serve` und `Schliessen` dürfen `Now()` rufen, ohne dass etwas rot wird. Der Wortlaut deckt das nicht ab: `SPEC-238` Nr. 12 spricht vom Handler, `SPEC-301` von `New` und den Antworten. Keine Verletzung, keine Auflage. | `SPEC-238` Nr. 12, `SPEC-301` Nr. 5 g |
| V-4 | INFO | Für die Closure: die DoD-Häkchen sind nicht gesetzt; in §6 tragen zwei Zeilen den Ausgang „offen“ (Textstelle in `SPEC-041` Nr. 7; Satz im Handbuch zum Rücksprung der Zeit), beide beim Architect, und brauchen einen der drei Ausgänge; die Zeile zu den Testmengen steht auf „weiter offen“. | `AGENTS.md` §3.13 |
| V-5 | INFO | Vier der sechs gemessenen Teilungen der Sperre hängen ohne Detektor an `TestFensterReihenfolge` allein (Lesen – Vergleich, Öffnen – Prüfen, die zwei im Öffnen). Wer diesen Test ändert, prüft die Raten neu. | Messung oben |

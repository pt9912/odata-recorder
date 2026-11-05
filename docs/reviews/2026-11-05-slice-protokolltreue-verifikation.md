# Verifikation: slice-protokolltreue

* Datum: 2026-11-05
* Rolle: verifier (frischer Kontext; kein Produktcode, kein Test, kein Vertrag, kein Plan, keine Roadmap, keine Welle und kein Register geändert; einmaliger Lauf, keine Wiederholungsschleife)
* Gegenstand: Zweig `slice-protokolltreue`, Spitze `e2620cc`, Diff `195afaa..e2620cc`; Basis der Größenmarke und der Randformen `7c4e1df`; Implementer `d3c5823` bis `c588bb6`, Review `ce97f35`, Fix-Runde `fe76239` bis `adb7674`, Architect `e2620cc`.
* Plan: `docs/plan/planning/done/slice-protokolltreue.md`
* Maßstab: DoD (§2), Tabelle *Zusage · Fänger* (§3, Z1 bis Z22 mit Handoff-Absatz und Raten), Messpunkt (§4), Raster und Zusatztabelle (§6) gegen `LH-QA-07.a` Nr. 1 bis 7, [LH-QA-07](../../spec/lastenheft.md#lh-qa-07--protokolltreue), `LH-FA-01.a`, `SPEC-119` Nr. 6, `SPEC-121` Nr. 4; `AGENTS.md` §3.9 bis §3.18; [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) (Prüfpakete unter `tests/`), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md) (Race-Detektor), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Größenmarke).
* Hinweis zum Skill: `.harness/skills/closure-note-reviewer.md` gilt für Slices in `done/`; dieser Slice liegt in `in-progress/` und seine Closure-Notiz (§7) ist leer, der Skill passt nicht auf diese Rolle und wurde nicht angewandt.
* Belege selbst gefahren: Basislauf der Vergleichstests, Probe-Mutationen auf Wegwerfkopien (Scratchpad `ver-pt/`, Image `odata-recorder:race`, `--network none --memory 4g`, unter `flock /tmp/odata-recorder-heavy.lock`), Z21 je 30 Prozesse, `make gates` im Worktree. `df` vor den Läufen: 31 G frei, nichts gelöscht.

## Urteil

**Closure: ja, mit einer Auflage an den Planner.** Kein HIGH. Ein MEDIUM (V1: Kettenklasse, geordnete Paare), drei LOW, vier INFO. Jede DoD-Zeile, die der Code erfüllen kann, ist belegt; die offenen Zeilen sind Closure-Pflichten des Planners (Abschnitt 1). Der Fix der Runde (Zeilenliste, Kettenklasse) deckt die Klasse der Zeilenliste vollständig nach meinen Proben; die Kettenklasse deckt er in Schreibweise, leerem Rohbody und mehreren gzip-Gliedern, aber nicht in den geordneten Paaren (V1). Der Produktcode ist unberührt (`git diff 195afaa..HEAD -- src` leer), die Zusage `LH-QA-07` bleibt für die 16 Antwortarten erfüllt: der Vergleich über den echten Recorder findet keine Abweichung (`TestProtokolltreue` grün).

## Befunde nach Schwere

| Nr. | Schwere | Befund | Adresse |
|---|---|---|---|
| V1 | MEDIUM | Kettenklasse nicht ganz gedeckt (`AGENTS.md` §3.15, §3.17): `LH-QA-07.a` Nr. 3 und `SPEC-119` Nr. 6 meinen jede Kette aus höchstens zwei Gliedern über `gzip`, `deflate`, `identity` (3 Einzelketten, 9 geordnete Paare, je als eine Zeile oder als zwei Zeilen von `Content-Encoding`), in umgekehrter Reihenfolge entpackt. Der Fänger führt als Paare nur `deflate, gzip` (`TestKetten`, `TestKettenSchreibweise`, in einer und zwei Zeilen). Probe-Mutation M2: in `entpacke` wird das Paar `gzip, deflate` vor dem Entpacken vertauscht, alle Tests grün. M3: `slices.Compact(glieder)` nach der Längenprüfung (`gzip, gzip` gilt als ein Glied), alle Tests grün; die drei Glieder `gzip, gzip, gzip` fängt `TestNichtEntpackbar` nur, weil die Längenprüfung vor der Verdichtung liegt. Die Vergleichsfunktion selbst ist richtig: eine Probe über alle 12 Ketten (3 Einzelketten, 9 Paare) und die zwei-Zeilen-Form der 9 Paare entpackt zum selben Inhalt und erkennt einen geänderten Body (21 Fälle, grün). Die Lücke liegt im Fänger, nicht im Code; die Folge liegt im Test (ein Entpackfehler einer nicht geführten Kette würde als Abweichung oder als falsches Grün erscheinen), nicht im Produkt. | Planner: Fänger als Schleife über alle 12 Ketten und die zwei-Zeilen-Form der Paare ergänzen (die Probe `ver-pt/probe_chain_test.go` im Scratchpad ist die Form, etwa 45 Zeilen) oder die Prüfgrenze als akzeptiertes Negativ mit Grund und Folge in §3 nennen |
| V2 | LOW | Statuscode: kein Fall hat einen Status von 5xx oder 1xx (die Fälle tragen 200, 204, 302, 304, 404; `andererStatus` liefert 201, 500, 303, 200). Probe-Mutation M4: der Vergleich übergeht den Status, wenn die direkte Seite 500 oder höher trägt, alle Tests grün. Die Vergleichsfunktion ist richtig (Probe über 12 Statuscodes von 100 bis 599, gegen 200 und gegen sich selbst plus eins, grün). Folge im Betrieb: ein Proxy, der einen 5xx des Upstream verändert, fällt dem Vergleich nicht auf, wenn die Vergleichsfunktion so mutiert wäre; der Referenzdienst sendet keinen 5xx. | Planner: Fall oder akzeptiertes Negativ (Menge der Statusklassen) |
| V3 | LOW | Anzahl der Zeilen eines Namens: `TestZeilenliste` führt Formen bis drei Zeilen. Probe-Mutation M5: `relevante` kürzt jede Liste auf drei Zeilen, alle Tests grün. Probe über 1 bis 6 gleiche Zeilen mit geändertem letztem Wert: richtig (grün). Die Menge ist unbeschränkt; die Prüfgrenze (drei) steht nicht als akzeptiertes Negativ im Plan. | Planner: Prüfgrenze mit Grund und Folge nennen |
| V4 | LOW | Schreibweise des Namens nur für ASCII: Probe-Mutation M1: `asciiKlein` durch `strings.ToLower` ersetzt (kleingeschrieben auch Nicht-ASCII, etwa das Kelvin-Zeichen), alle Tests grün. Auf dem Draht nicht erreichbar (ein Header-Name ist ein Token aus ASCII-Zeichen, `net/http` weist anderes ab), und `LH-QA-07.a` Nr. 1 sagt „ohne Beachtung der ASCII-Schreibweise“. Fast äquivalent; als LOW geführt, weil eine Allaussage der Spezifikation („nur ASCII“) keinen Fänger hat. | Planner: Negativ mit Grund und Folge (keine im Betrieb) oder ein Fall mit einem Nicht-ASCII-Byte im Namen |
| V5 | INFO | Größenmarke: Messung 1517 (1141 + 251 + 125, Produktcode 0), bestätigt mit `git diff --numstat 7c4e1df..HEAD -- tests/`. Unter der Marke 2700 des Gates (`SPEC-128`). Nach dem Wortlaut von `AGENTS.md` §3.14 (Marke 1500, Formelkonflikt, §6 *offen*, Adresse User) läge die Messung 17 Zeilen darüber; die Entscheidung bleibt beim User, und die Marke des Blocks ist die des Gates. Der Satz in §4 „Schätzung, die der Messung gleicht (1404)“ nennt den Stand beim Handoff, nicht den Stand nach der Fix-Runde (1517); er ist als solcher benannt. | User (Formel), Planner (Closure) |
| V6 | INFO | Zeilen ohne Befund im Vergleich, die der Vergleichstest nicht fängt, aber die Einheiten des Produkts: ein Body über `kopierPuffer` (32 KiB; Probe-Mutation M-pd: ein Byte weniger bei vollem Stück) und ein Body ohne `Content-Type` (Probe-Mutation M-pe: der Schutz vor dem Erraten des Typs entfernt) bleiben im Vergleich grün, weil kein Fall des Referenzdienstes sie trägt (kein Fall, der Plan zählt sie nicht als Antwortart). Beide fängt der Test des Adapters: `TestAntwortBytes`, `TestLaufenderRequestEndet`, `TestSchliessenFrueh` rot bei M-pd, `TestAntwortContentType` rot bei M-pe. | — |
| V7 | INFO | Z21 nachgemessen (Abschnitt 4): 30 von 30, ohne Mutation 0 von 30. Z21 fängt ein Rennen des Test-Servers; ein Rennen des Recorders unter dem Vergleich meldet nur der Detektor, soweit ein Lauf es auslöst (`SPEC-122` Nr. 8). | — |
| V8 | INFO | Probe-Mutation M7 (der Vergleich prüft nur die Zeilenlisten, nicht, ob ein Name vorhanden ist) baute nicht (`pda` unbenutzt) und ist nicht gefahren; äquivalent auf dem Draht (ein Name mit null Zeilen kommt nicht vor), kein Befund. | — |

## 1. DoD Punkt für Punkt (Plan §2)

| # | DoD | Stand | Beleg |
|---|---|---|---|
| 1 | Vergleichstest nach `LH-QA-07.a`: derselbe Request direkt und über den Recorder im Modus `record` an den Referenzdienst; Status, Body nach dem Entpacken (`SPEC-119` Nr. 6), relevante Header ohne Hop-by-Hop, `Content-Encoding`, `Content-Length`, `VOLATILE_HEADERS`; `Accept-Encoding` nur `gzip`, `deflate`, `identity`; je Antwortart ein Fall | erfüllt | `TestProtokolltreue` über 16 Fälle in `faelle()`: `$metadata`, Sammlung, Entity, `$count`, `$value` (beide `text/plain`), Fehler 404, `gzip`, `deflate`, Kette, `identity`, `HEAD`, 204, 304, 302, Header mit mehreren Zeilen und leerem Wert, nicht angefragtes `br`; `TestAcceptEncodingNurBekannte` (Z19) |
| 2 | Mutationen des Vergleichs: Byte, Header, Status, übergangener Header färben rot; ein anderes Coding über den Proxy färbt nicht | erfüllt | `TestBodyAbweichung`, `TestHeaderWertAbweichung`, `TestStatusAbweichung`, `TestHeaderNurEineSeite` (Veränderer zwischen Client und Recorder), `TestCodingUndLaenge` (44 Fälle von Codings, grün); echte Mutationen am Proxy (Plan §3, 4 von 4) und von mir (Abschnitt 3) |
| 3 | Test läuft in der Stufe `test` (netzlos, Loopback); Randformen vor dem Code in der Spezifikation | erfüllt | `tests/integration` gehört zur Stufe `source` und damit zu `make test` ([ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)); Randformen in `LH-QA-07.a` Nr. 1 bis 7 seit `7c4e1df`, vor dem ersten Code-Commit `d3c5823` (`git log`: `7c4e1df` ist Vorfahr von `d3c5823`); Nr. 7 d nach dem Review durch den Architect (`e2620cc`) |
| 4 | `make gates` grün | erfüllt | Abschnitt 5 |
| 5 | Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-11-05-slice-protokolltreue-review.md` (`ce97f35`): 2 MEDIUM, 2 LOW, 3 INFO; Nachverifikation Abschnitt 2 |
| 6 | Doku-Update (`LH-QA-07.a`, Nachzug durch den Architect) | erfüllt | `spec/spezifikation.md` Nr. 7 d und Zeile der Historie (`e2620cc`); im Diff sonst keine Änderung der Spezifikation seit `7c4e1df` |
| 7 | Closure-Notiz mit Steering-Loop-Lerneintrag | offen für die Closure | §7 trägt „Wird bei der Closure gefüllt“ |
| 8 | Beobachtungs-Register fortgeschrieben | offen für die Closure | Kandidaten: Fänger führt nur eine Form der Klasse (Zeilenliste, Kettenklasse; Review-Findings 1 und 2, hier V1 mit den geordneten Paaren als drittes Beispiel derselben Klasse) |
| 9 | Jedes Risiko aus §6 trägt einen Ausgang | offen für die Closure | Stände: nicht angefragtes Coding (entfallen, Z15 mit Mutation rot); Unterschied (nicht eingetreten, Vergleich ohne Abweichung in 16 Antwortarten); gemeinsamer Referenzdienst (Z22: `git diff 195afaa..HEAD` über `m2_test.go`, `referenzdienst_test.go`, `text_test.go` leer, von mir gelesen); Nebenläufigkeit (Z21 30 von 30); Randformen (keine im Code entschieden, Reihenfolge der Schreibweisen als Festlegung der Vergleichsfunktion vom Architect geurteilt); Formelkonflikt der Größenmarke bleibt *offen* (User) |
| 10 | Drei Paarungen getragen | offen für die Closure | von der nächsten Welle-Closure geprüft |

Ein Häkchen setze ich nicht.

## 2. Nachverifikation der Befunde des Reviews

| Befund | Stand | Beleg | Ergebnis |
|---|---|---|---|
| 1 MEDIUM Zeilenliste (leer neben gefüllt, gleiche Zeilen, Leerraum am Rand) | `TestZeilenliste`: 17 Formen, 17 × 17 = 289 Paare, gleich genau bei derselben Form | gelesen; Mutanten M5 (Kürzen auf drei Zeilen) grün, alle anderen Elemente der Klasse tot: Probe über 1 bis 6 Zeilen (Abschnitt 3); Mutanten a, b, c des Reviews sind Teil der Paarmenge (leer neben gefüllt in jeder Stellung, `a, a`, `a `, ` a`, Tabulator) | geschlossen; Rest V3 |
| 2 MEDIUM Kettenklasse (Schreibweise, leerer Rohbody bei unbekannter Kette, mehrere gzip-Glieder) | `TestKettenSchreibweise` (16), `TestNichtEntpackbar` (28 Fälle mit leerem Rohbody), `TestMehrereGzipGlieder` (4) | gelesen; Elemente der Klasse, die der Review nicht nannte, in Abschnitt 3 | geschlossen für die drei genannten Teilklassen; die geordneten Paare bleiben offen (V1) |
| 3 LOW Reihenfolge der Schreibweisen in einer Map | `TestSchreibweisenInEinerMap`, Urteil des Architect (§6): Festlegung der Vergleichsfunktion, nicht des Produkts | Mutation `out[k] = h[n]` hier rot (4 Formen) | geschlossen |
| 4 LOW Z13 Wortlaut und Negativ HEAD | Z13 im Plan führt die handgebaute Umsetzung; Negativ in `LH-QA-07.a` Nr. 7 d mit Grund und Folge | gelesen: Zeile Z13, Handoff-Absatz (2), §6 *Negativ HEAD*, Nr. 7 d stimmen überein | geschlossen |

## 3. Tabelle *Zusage · Fänger* gegen den Vertrag, Fix gegen die Klasse (§3.17)

Z1 bis Z19 haben je einen Fänger, der im Code liegt, und eine Zählprüfung am Ende der Schleife (gelesen: 16, 68, 36, 63, 27, 15, 44, 30, 9, 38, 28, 289, 16, 2; die Zahl 33 von `TestBodyAbweichung` nach Plan, nicht einzeln nachgerechnet); der Basislauf ist grün. Z20 ist ein akzeptiertes Negativ mit Grund und Folge (Abschnitt 6), Z21 nebenläufig (Abschnitt 4), Z22 durch den leeren Diff.

Probe-Mutationen des Verifiers auf Wegwerfkopien, je eine Ersetzung, Paket `tests/integration` (Vergleichstests, ohne Z21), Elemente, die weder der Reviewer noch der Implementer nannten:

| Klasse (Zusage) | Element | Mutation | Ergebnis |
|---|---|---|---|
| Hop-by-Hop-Liste (Z9, Nr. 2) | `Keep-Alive` | Name in der Liste als `Keepalive` | rot: `TestHopByHopUndConnection` |
| `VOLATILE_HEADERS` (Z8, Nr. 2) | `X-Correlation-Id` | aus dem Default in `src/bootstrap/konfig_werte.go` entfernt | rot: `TestVolatileHeaders`, `TestFaelleDirektWieErwartet` |
| Header des Proxys (Z3, Nr. 1) | `Location` (Weiterleitung) | `setzeHeader` im Proxy lässt `Location` aus | rot: `TestProtokolltreue`, `TestSelbstvergleich`, `TestFaelleDirektWieErwartet` |
| Header des Proxys (Z3) | `ETag` | Wert in Kleinschreibung | rot: `TestProtokolltreue` |
| Zeilenliste (Z6, Nr. 1) | letzte Zeile gewinnt | `ziel.Add` zu `ziel.Set` | rot: `TestProtokolltreue` |
| Status (Z1, Nr. 1) | 302 | Proxy sendet 307 | rot: `TestProtokolltreue`, `TestSelbstvergleich`, `TestFaelleDirektWieErwartet` |
| Status (Z1) | 404 | Proxy sendet 410 | rot: `TestProtokolltreue`, `TestFaelleDirektWieErwartet` |
| Kettenklasse (Z11, Nr. 3) | Paar `gzip, deflate` | M2: in `entpacke` vertauscht | **grün (V1)** |
| Kettenklasse (Z11) | Paar `gzip, gzip` | M3: `slices.Compact` nach der Längenprüfung | **grün (V1)** |
| Statusklasse (Z1) | 5xx | M4: Status bei 500 oder höher übergangen | **grün (V2)** |
| Zeilenanzahl (Z5, Z6) | vierte Zeile und mehr | M5: Liste auf drei Zeilen gekürzt | **grün (V3)** |
| Schreibweise (Z7, Nr. 1) | Nicht-ASCII-Byte im Namen | M1: `strings.ToLower` statt `asciiKlein` | **grün (V4)** |
| Body (Z2, Nr. 3) | Body über 32 KiB | M-pd: ein Byte weniger bei vollem Stück | Vergleich grün, Test des Adapters rot (V6) |
| Body (Z2) | Body ohne `Content-Type` | M-pe: Schutz vor dem Erraten des Typs entfernt | Vergleich grün, Test des Adapters rot (V6) |

Probe-Pakete über die Klassen (`ver-pt/probe_chain_test.go`, nur im Scratchpad): alle 12 Ketten und die zwei-Zeilen-Form der 9 Paare (21 Fälle), 1 bis 6 Zeilen, 12 Statuscodes: gegen den Stand `e2620cc` grün, die Vergleichsfunktion verhält sich in der Klasse richtig; die Lücken liegen in den Fängern, nicht im Code.

Ein Fix gegen die Klasse (§3.17) liegt für die Zeilenliste vor (alle Elemente der Menge der 17 Formen, Paar für Paar) und für drei von vier Teilklassen der Kette; die Mengen sind im Bericht des Implementers (Plan §3, Handoff) mit ihrer Zahl genannt. Ohne Zahl bleibt die Prüfgrenze der Zeilenanzahl (V3).

## 4. Z21 nachgemessen

Messender: der Verifier. Test-Binary einmal mit `-race` gebaut (Image `odata-recorder:race`, `CGO_ENABLED=1`, `--network none --memory 4g`, unter `flock`), 30 Prozesse nacheinander mit `-test.run 'TestParallel$' -test.count=1`.

| Lauf | Prozesse | rot | davon mit `DATA RACE` |
|---|---|---|---|
| Mutation: Paketvariable `zaehlerMut++` in `bediene` des Referenzdienstes ohne Sperre | 30 | 30 | 30 |
| ohne Mutation | 30 | 0 | 0 |

Die Rate der Tabelle (30 von 30, ohne 0 von 30) ist bestätigt; der Reviewer maß dasselbe. Sie gilt für ein Rennen im Test-Server; ein Rennen des Recorders unter dem Vergleich meldet nur der Detektor bei einem Lauf, der es auslöst (`SPEC-122` Nr. 8, [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)).

## 5. Gates

`make gates` im Worktree auf dem Stand `338d0e7` (Spitze mit diesem Bericht), unter `flock /tmp/odata-recorder-heavy.lock`, Rückgabe 0; `groessenmarke slice-protokolltreue.md: ok, Messung 1517 von Marke 2700`; `zeichen: ok, 1259 Datei(en) geprüft`. Ein früherer Lauf vor dem Commit des Berichts war ebenfalls grün (Rückgabe 0). Die Stufen `test` und `race` laufen über `./...` und damit über `tests/integration`.

## 6. Akzeptierte Negative (Z20, `LH-QA-07.a` Nr. 7 a bis d)

| Negativ | Grund | Folge im Betrieb | Urteil |
|---|---|---|---|
| a) Abbruch im Body | Der Vergleich liest nur vollständige Antworten; ein Abbruch ist keine Antwort des Referenzdienstes | ein Proxy, der einen Abbruch anders weiterreicht, fällt in diesem Vergleich nicht auf; der Ort (`LH-FA-02.b`, `SPEC-119` Nr. 12) ist genannt, ein Test des Verhaltens liegt im Adapter (`src/adapters/driving/httpodata/`, Tests zum Abbruch und zum Schließen gefunden) | Grund und Folge tragen; akzeptiert |
| b) Upstream nicht erreichbar | Direkt gibt es keine Antwort, über den Proxy `ODR-E4001` (`SPEC-120` Nr. 7) | keine; die Fehlerabbildung ist Gegenstand von `SPEC-120` (Test in `src/adapters/driving/httpodata/fehler_test.go` gefunden) | akzeptiert |
| c) Coding des Clients außerhalb der drei | Der Test kann `br` und `zstd` nicht entpacken; ein Client, der jedes genannte Coding ausschließt, bekommt direkt 406 | ein Unterschied im Weg, den ein anderes Coding nimmt, bleibt dem Test von `SPEC-121` Nr. 4 überlassen | akzeptiert |
| d) Body bei `HEAD`, 204 und 304 über den Draht | Der HTTP-Server des Recorders liefert dort keinen Body aus; ein Mutator kann ihn nicht senden | ein Proxy, der bei `HEAD` einen Body hielte, fiele diesem Vergleich nicht auf; für 204 und 304 liegt der Fänger an handgebauten Antworten (`TestOhneBody`, beide Seiten, rot bei Mutation „Body bei 304 nicht verglichen“ laut Plan), bei `HEAD` prüft dieselbe, methodenunabhängige Funktion den leeren Body | Grund und Folge tragen; die Zeile im Raster (§6 *Body*, *fehlender Wert*) nennt Nr. 3, und Nr. 7 d trägt den Wortlaut; akzeptiert |

Alle vier tragen Grund und Folge im Vertrag, nicht nur im Plan; keine Zusage ohne Fänger bleibt als Negativ ohne Folge stehen.

## 7. Raster (§6) gegen den Vertrag

Fünfzehn Zellen und vier Paare gelesen: jede Zelle nennt Nr. 1 bis 7 von `LH-QA-07.a` oder „gilt nicht“ mit Grund; die genannten Nummern tragen den Wortlaut der Entscheidung. Keine Zelle ist im Code entschieden außer der Reihenfolge der Zeilen zweier Schreibweisen in einer handgebauten Map, die der Architect als Festlegung der Vergleichsfunktion urteilte (§6, `TestSchreibweisenInEinerMap`). Die Allaussage „kein Unterschied in Status, Body und relevanten Headern“ hat ihren Fänger über alle 16 Fälle (§3.15).

## 8. Einhaltung der Hard Rules (Stichproben)

* §3.2 und §3.7: keine `//nolint`-Direktive in den drei Testdateien; Kommentare im Indikativ, der Kopf von `TestFaelleDirektWieErwartet` im Indikativ (`c588bb6`).
* §3.18: `LC_ALL=C grep -nP` auf Folgen aus Backslash, `u` und vier Hex-Ziffern, BOM und NBSP in `tests/integration/vergleich_test.go`, `vergleichsdienst_test.go`, `protokolltreue_test.go`: leer; `brotliBytes()` ist aus Bytewerten gebaut. Gegengelesen wurden diese drei Dateien; dieser Bericht trägt keine solche Folge (siehe Abschnitt 5 für den Lauf von `make zeichen-check`).
* §3.9: der Plan trägt die Fix-Runde in §3 (Handoff-Absatz, Zuordnung (6)), §4 (Messpunkt 1517) und §6 (Negativ HEAD, Urteil zur Reihenfolge) nach.
* §3.14: Messpunkt 1517 von 2700 bestätigt (V5).

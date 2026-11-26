# Verifikations-Report: slice-lasttest-auswertung (Teil 2 des Lasttests, Zeilen A1 bis A13) — 2026-11-26

**Verifikations-Art:** Code gegen DoD (`§2`), Spezifikation und Entscheidungen des Architect: DoD- und Entscheidungs-Konformität, Plan-vs-Code-Diff, Fix-Runde gegen die Klasse (`AGENTS.md` §3.17). Kein Review (der Review ist `2026-11-26-slice-lasttest-auswertung-review`) und keine Validierung.

**Bezug:** `LH-QA-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md); `harness/vertraege/SPEC-216-lasttest.md` Nr. 7 bis 13 und 16, `SPEC-231` Nr. 6.

**Gegenstand:** Zweig `slice-lasttest-auswertung`, `HEAD` `9b840a49` (Handoff `5cc21efc`, Review-Report `de091453`, Fix-Commit `5679f23c`, danach Merges von `spec-lasttest-randformen` und `main`); Messbasis `ea3547be`. Gegen `main` ändert der Zweig den Plan, den Review-Report und zehn Dateien unter `tests/lasttest/` (`git diff --name-only main...HEAD`); `src/`, `Dockerfile`, `harness/`, `.golangci.yml`, `.a-check.yml` und `tools/` sind unverändert.

**Rolle:** Verifier (Modul 11) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-26

**Urteil: freigegeben unter Auflagen.** Kein HIGH, sechs MEDIUM (V-1 bis V-6), vier LOW (V-7 bis V-10), drei INFO (V-11 bis V-13). Der Produktcode tut an jeder geprüften Stelle, was `SPEC-216` Nr. 10 bis 13 verlangt, einschließlich der zwei geänderten Stellen; die Ränder ohne Ort sind V-7 und V-8. Die Fix-Runde färbt die Mutationen, die der Review nannte, deckt aber die Klasse bei F-1, F-2, F-3, F-4, F-6 und F-7 nicht: an Elementen, die weder Review noch Implementer nannten, bleiben Mutationen grün, eines davon (null fehlende Spans) hatte der Review selbst genannt. Die sechs MEDIUM sind vor dem Merge zu schließen; sie verlangen Fälle in den Tests und keine Änderung am Produktcode.

**Eingang und Beleg der Behauptungen.** `make gates` als Ganzes habe ich nicht gefahren. Der Nachweis `.harness/state/gates-passed.head` nennt `9b840a49`, `gates-passed.diffsha` stimmt mit `tools/harness/working-tree-hash.sh` am Baum vor diesem Report überein (Stempel 11:17): der Gate-Lauf deckt genau diesen Stand. `pruefung test`, `pruefung race`, `a-check`, `lint` und die Messungen habe ich selbst gefahren.

## DoD (§2) Punkt für Punkt

| Punkt | Stand | Beleg |
|---|---|---|
| Leser und Auswertung (A1 bis A6) | erfüllt im Verhalten, Fänger mit Auflagen | Code gegen Nr. 10 a bis g gelesen; Tests frisch grün; V-2 bis V-6 |
| Gültigkeit, Last, Urteil, Ausgabe (A7 bis A13) | erfüllt im Verhalten, Fänger mit Auflagen | Folge der Schlüssel je Zeilenart und die 13 Tokens gegen Nr. 13 gelesen; V-1, V-6, V-8 |
| `make gates` grün | erfüllt | Stempel wie oben; `pruefung test` und `race` selbst frisch grün |
| Review durchgeführt, Report liegt vor | erfüllt | `2026-11-26-slice-lasttest-auswertung-review` (nicht freigegeben, danach Fix-Runde) |
| Doku-Update, falls öffentlicher Vertrag berührt | erfüllt | kein Vertrag und keine Spezifikation durch den Teil geändert |
| Handbuch und Katalog | erfüllt bis auf den Satz in §7 | `src/` und `docs/user/` unverändert; `katalog-check` liegt in der gestempelten Kette; §7 ist Closure |
| Closure-Notiz, Register, Ausgänge der Risiken, drei Paarungen | offen (Closure) | §7 leer; §6 trägt je Zeile einen Ausgang, vier davon *offen* mit Adresse (Höhe der Marke, Zeit des Pakets, Sendung an Teil 3, Zahl der Proben) |

**Die zwei geänderten Stellen.** `urteilBetriebsart` ohne Lauf liefert `ungueltig` mit `zaehlung`, `laeufe` 0 und alle Dauern 0 (Nr. 12 d); `TestStreuung` führt leere Liste und `nil` gegen die ganze Struktur. `warteAufSpans` liefert wahr genau bei erreichter oder überschrittener Zahl, auch im Zweig der Frist (`return s.anzahl() >= erwartet`). Im Produktcode dieses Teils ruft niemand die Funktion; alle Aufrufe stehen in Tests, den Aufrufer baut Teil 3 (V-13). Die Rückgabe ist nur in Teilen gefangen (V-2).

**Entscheidungen des Architect.** Die Randformen aus Nr. 10 b bis 10 g, 11, 12 d und 13 (`ereignis` Byte für Byte, leere Texte, `0.000`, Summe 0, erstes Feld der Last-Datei, Urteil ohne Lauf, Modus fehlend oder anderer Art, Folge in `Hinweis`, abgeschnittene Dauern, `kind`) stehen so im Code; keine Abweichung gefunden. `a-check` meldet 0 Befunde (Schicht `tests`), kein `//nolint`, kein Gate gelockert.

**Plan-vs-Code.** Die zehn Dateien und ihre Zeilenzahlen stimmen mit §3 *Messung* überein; alle im Plan genannten Testnamen bestehen; keine Datei außerhalb der Tabelle §3. Abweichungen im Text: V-7, V-9, V-10.

## Findings

| ID | Kategorie | Befund | Ort | Mutation (grün: zuständiger Test und alle Tests von Teil 2) | Klasse |
|---|---|---|---|---|---|
| V-1 | MEDIUM | Zeile `Last` und Gültigkeit der Proben: eine Probe über der Schwelle ist im Lauf allein in der Mitte (`spitze`, 12. von 24) und am ersten Tick (`TestLastNebenDerLast`) geführt, nie **vor dem Warm-up** und nie **nach dem Nachlauf**, die Nr. 11 beide nennt. Die kleinste Probe und die Schwelle haben keinen Fänger: die erste Probe jedes Falls ist zugleich die kleinste (1,00), und `Schwelle` im Ergebnis wird nur bei 20 Kernen geprüft (Fixture gleich Wert der Mutation). | `tests/lasttest/last.go` · `for _, pr := range p.proben {` und `Schwelle: schwelle(p.kerne)` | `for _, pr := range p.proben[:len(p.proben)-1]`; `... p.proben[1:]`; `e.Min` nicht mehr gebildet; `Schwelle: 1000` | F-6 und A8: Stelle der Probe im Lauf, Felder der Zeile `Last` (`AGENTS.md` §3.15) |
| V-2 | MEDIUM | `warteAufSpans`: geführt sind 1, 2, 5 und 100 fehlende Spans. **Null fehlende** (der Review nannte „null, eins, mehrere“; zugleich der Normalfall im Betrieb), **mehr als erwartet** und **ein Teil der Spans kommt, dann läuft die Frist ab** fehlen; die Abweichungs-Tabelle des Plans nennt dieselben vier Zahlen. | `tests/lasttest/auswertung_test.go` · `func TestNachlaufWarten`; `auswertung.go` · `for s.anzahl() < erwartet {` | wartet bei erreichter Zahl die ganze Frist ab; `for s.anzahl() != erwartet`; Rückgabe wahr bei jedem Fortschritt; `ctx.Err()`-Prüfung vor der Schleife. Dazu `return false` im Zweig der Frist (nur im Wettlauf unterscheidbar: akzeptiertes Negativ, falls so beschlossen) | F-2 und Rückgabe der geänderten Stelle |
| V-3 | MEDIUM | Der Bezugspunkt der Nachlauf-Frist hat keinen Fänger: eine Frist, die mit jedem neu gelesenen Span neu beginnt (Leerlauf-Frist), wartet über „höchstens 10 s“ (Nr. 10 e) hinaus und bleibt grün. Kein Fall stellt die Uhr zwischen zwei Spans vor und verlangt danach das Ende bei 10 s ab Aufruf. Eine naive Fassung derselben Mutation wird rot, aber aus einem Nebengrund (zwei Wartende an der Fake-Uhr, „Wartender der Frist: nicht eingetreten“); sie zählt nicht. | `tests/lasttest/auswertung.go` · `ctx, cancel := mitFrist(ctx, uhr, nachlaufFrist)` | Frist neu bei jedem Anstieg von `s.anzahl()` | F-1: Frist am Aufruf, Achse Bezugspunkt |
| V-4 | MEDIUM | `recorder.mode` an einem Gesamtspan **mit Kind** hat keinen Fänger. `TestPhasen` und `TestModusJeBetriebsart` bauen nur Gesamtspans ohne Kind, `TestZuordnung` verwirft die Verstöße, `TestSammlerGleichzeitig` führt nur den richtigen Wert. In `replay` und `mock` trägt der Gesamtspan im Regelfall ein Kind: die Mutation hebt die Prüfung dort auf. Der Kommentar von `nimm` sagt „jeder Gesamtspan“. | `tests/lasttest/auswertung.go` · `if z.Modus != s.betriebsart {` | `if !g.HatKind && z.Modus != s.betriebsart` | F-3: Zeile des Gesamtspans, Achse Kind |
| V-5 | MEDIUM | Eine Zeile **ohne `ereignis`** (Pflichtfeld, Nr. 10 b) ist nicht geführt: unter den 46 fremden Formen fehlen `msg`, `span_id` und `duration_ms` einzeln, `ereignis` nicht (allein in `{}` und `null`, denen auch alles andere fehlt). Eine JSON-Logzeile mit `msg`, `span_id` und `duration_ms` ohne `ereignis` zählte unbemerkt als Span. | `tests/lasttest/leser_test.go` · `plan := []string{` | fehlendes `ereignis` gilt als `span` | F-7 und A2: jedes Pflichtfeld fehlend |
| V-6 | MEDIUM | Zwei Zusagen von F-4 ohne Fänger: (a) `ohne_kind` zählt „Warm-up, Messphase und jeder weitere“ (Nr. 10 c); ein Gesamtspan ohne Kind hinter *W* + *N* ist nicht geführt. (b) `toleranz_ms` ist ohne zwei gültige Läufe `0.000` (Nr. 13 b); der Fall mit einem ungültigen Lauf prüft `Streuung`, nicht `Toleranz`. Nr. 13 b nennt als Fänger „`TestStreuung` (je Fall Streuung, Toleranz, …)“: das trifft nicht zu. | `tests/lasttest/auswertung.go` · `OhneKind: s.ohneKind`; `tests/lasttest/urteil_test.go` · `ein ungültiger Lauf` | `ohne_kind` nur über `s.gesamt[:bis]`; `u.Toleranz` auch bei einem ungültigen von zwei Läufen gesetzt | F-4: Phase des Spans; Felder des Urteils |
| V-7 | LOW | Das akzeptierte Negativ (9) ist falsch: die Kurzprüfung in `leseSpan` ist **nicht** im Ergebnis gleichwertig. Eine Zeile mit führendem CR vor `{` (Leerraum von JSON) zählt als fremd; ohne die Kurzprüfung wäre sie eine Span-Zeile (Probe: Leerzeichen und TAB gelesen, CR fremd). Das ist eine im Code entschiedene Randform ohne Ort in `SPEC-216`; dass der Adapter eine solche Zeile nicht schreibt, ist Annahme, nicht gemessen. | `tests/lasttest/leser.go` · `bytes.TrimLeft(zeile, " \t")`; Plan §3 Negativ (9) | Probe; dazu `json.Unmarshal(zeile, …)` ohne Kurzprüfung: Tests grün | Negativ mit falscher Aussage (`AGENTS.md` §3.15) |
| V-8 | LOW | `exitCode(ausgang{})` ist 0: der Wert ohne Eingabe ist Erfolg. `TestExit` schließt das Paar „beide nicht gefahren“ aus, Nr. 12 e entscheidet es nicht; für das Urteil ohne Lauf hat der Architect das Gegenteil entschieden (Nr. 12 d). Ob ein Weg des Programms es erreicht, entscheidet Teil 3. | `tests/lasttest/urteil.go` · `return exitBestanden`; `urteil_test.go` · `ohne das Paar „beide nicht gefahren“` | Probe: Ergebnis 0 | im Code entschiedene Randform (`AGENTS.md` §3.16) |
| V-9 | LOW | Nachzug unvollständig (`AGENTS.md` §3.9, §3.11): die Zeile A8 der Tabelle nennt weiter „Wirkung: eine Probe, die die Last anhält, verschiebt den Takt“, den kein Test mehr führt (F-9 ist nur in Testkopf und *Treffpunkt* berichtigt); der Kommentar von `starteProben` sagt „halten sie nicht an“, der Plan sagt, das prüfe erst Teil 3; *Berührte Spec-Stellen* im Kopf nennt Nr. 16 g, h, j, nicht 16 o und p, die die Negative (10) und (11) tragen. | Plan §3 Zeile A8 und Kopf; `tests/lasttest/last.go` · `und halten sie nicht an` | gelesen | berichtigte Aussage nicht an allen Stellen |
| V-10 | LOW | Zählprüfungen: der Plan sagt „Zählprüfung je Menge im Test“. Die von A11 rechnet `len(faelle)*2` plus Literale und vergleicht mit Literalen: die 10 übrigen der 48 Fälle zählt sie nicht gegen die Quelle. Ohne Zählung sind außerdem die 13 Golden-Zeilen (A13), die 7 Zahlenformen, die 15 fremden Werte von `grund`, die 9 Grenzen (A8) und die 6 Dauern (A2). | `tests/lasttest/urteil_test.go` · `if n := len(faelle)*2 + 3 + 2 + 2 + 2 + 1` | die zwei Fälle „ohne Lauf“ aus dem Test entfernt: `TestStreuung` grün | Zählprüfung ohne Quelle |
| V-11 | INFO | `bleibtOffen` (2000 Wechsel der Goroutinen) trägt in allen gemessenen Armen (unten), kein Fehlalarm am richtigen Code. Das Maß im Plan („eine spätere nicht“) ist ehrlich. Für die Frist ist er allerdings der einzige Fänger, obwohl ein deterministischer am Argument vorliegt (`AGENTS.md` §3.10): die Fake-Uhr hält `bis` des Wartenden. Probe: `bis` gleich Beginn + 10 s grün am richtigen Code, rot bei halbierter Frist (5 s) und bei 10 s − 1 ns (9,999999999 s). | `tests/lasttest/auswertung_test.go` · `func bleibtOffen` | Probe | Vorschlag |
| V-12 | INFO | Zwei kleine ungeführte Elemente: der Leerraum der Last-Datei ist an 5 von 25 Zeichen mit `White_Space` geprüft (Nr. 11 nennt „die übrigen“); `Hinweis` führt `binary_exit` mit dem einen Wert 1. | `last_test.go`, `ausgabe_test.go` | `FieldsFunc` über die 5 Zeichen; `"binary_exit", "1"`: beide grün | Menge ohne Maß |
| V-13 | INFO | Die Sendung an Teil 3 steht allein in diesem Plan; `slice-lasttest` führt sie weder in §1 noch in der DoD. Sie muss zusätzlich den Gebrauch der Rückgabe nennen: `auswerte` zählt beim Aufruf; treffen Spans nach abgelaufener Frist und vor dem Ende des Binary ein, trägt nur noch die Rückgabe `false` von `warteAufSpans` den Verstoß `zaehlung` (Nr. 10 e „zu wenige nach 10 s“). | Plan §6 *Sendung an Teil 3*; `docs/plan/planning/next/slice-lasttest.md` | gelesen | Adresse nimmt noch nicht an (`AGENTS.md` §3.13) |

## Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

| Befund | Menge aus dem Wortlaut | Eigene Mutationen | Deckt die Klasse |
|---|---|---|---|
| F-1 Frist am Aufruf | Wert (kürzer, länger), Bezugspunkt | 10 s + 1 ns rot; Leerlauf-Frist grün | Wert ja, Bezugspunkt nein (V-3) |
| F-2 Warten und Rückgabe | fehlend 0, 1, mehrere; zu viele; Teilankunft; Ende des Kontexts | fünf grün | nein (V-2) |
| F-3 `recorder.mode` | Betriebsart × Wert × Zeile (Stelle, mit und ohne Kind) | „die letzte Zeile gewinnt“ rot; „nur ohne Kind“ grün | Betriebsart und Wert ja, Kind nein (V-4) |
| F-4 Zählung und Urteil | Phase × Zähler; Felder des Urteils je Art der Läufe | beide grün | nein (V-6) |
| F-5 Ausgabe | Zeilenart × Feld, beide Werte der Booleschen, Rest unter 1 ms | allein `binary_exit` als Literal grün | ja (Rest V-12) |
| F-6 `beende` | wartet; fasst jede Probe zusammen; Felder | vier grün | Warten ja, Zusammenfassen nein (V-1) |
| F-7 Randformen | jedes Pflichtfeld fehlend, leer, falscher Art; Form der Last-Datei | `ereignis` fehlt grün; Leerraum grün | nein (V-5, V-12) |

**Zahl der gültigen Formen in A2:** nachgezählt **10** (3 des Plans, 1 mit Leerraum, 3 mit `recorder.mode` fehlend, `null`, Zahl, 3 mit leerem Text); die Prüfung `len(gueltig) != 10` stimmt. Fremde: 13 + 33 = 46.

**Orakel.** Die Erwartungen von A7, A9, A10, A11, A12, A13 sind Literale, die von A2, A5, A8 und den Zahlenformen eigene Rechnung ohne Aufruf des Codes. Die Zählungen siehe V-10.

**Randformen-Raster und Kommentare.** §6 führt zwei Zeilen (p95; Vergleich und Ausgabe), jede Zelle mit Ort; für Proben, Nachlauf und `beende` steht keine Zeile, ihre Orte liegen in Nr. 10 e und 11. Zusagen ohne Fänger in Kommentaren: `nimm` (V-4), `beende` „fasst alle zusammen“ (V-1), `starteProben` (V-9).

**Akzeptierte Negative.** (4), (5) a bis d, (6), (8), (10), (11) nennen Grund, Folge und Maß, gelesen ohne Befund; (7) nicht nachgemessen; (9) ist falsch (V-7).

## Raten der nebenläufigen Fänger

Messender: der Verifier. Host mit 20 Kernen, ein Container ohne Netz (`--memory=4g --memory-swap=4g --cpus=4`), je Zelle 20 Prozesse nacheinander mit `-test.count=1`. 1-Minuten-Last des Hosts am Ende jeder Zeile 2,1 bis 3,0 (vor dem Lauf 6,4).

| Mutation | Test | rot von 20 (ohne 1, ohne 4, mit 1, mit 4) | Nullprobe |
|---|---|---|---|
| `beende` ohne `<-p.fertig` (F-6) | `TestLastNebenDerLast` | 20, 20, 20, 20 | 0, 0, 0, 0 |
| `if` statt `for` (F-2) | `TestNachlaufWarten` | 20, 20, 20, 20 | 0, 0, 0, 0 |
| `erwartet-1` (F-2) | `TestNachlaufWarten` | 20, 20, 20, 20 | 0, 0, 0, 0 |
| Frist halbiert (F-1) | `TestNachlaufWarten` | 20, 20, 20, 20 | 0, 0, 0, 0 |
| Frist − 1 ns (F-1) | `TestNachlaufWarten` | 20, 20, 20, 20 | 0, 0, 0, 0 |
| Probe synchron (A8) | `TestLastNebenDerLast` | 20, 20, 20, 20 | 0, 0, 0, 0 |

„ohne“ und „mit“ meint den Detektor, 1 und 4 `GOMAXPROCS`. Zusätzlich je Mutation der ersten fünf und je Arm 200 Prozesse, 12 zugleich: 200 von 200 rot in allen 20 Armen, Nullproben 0 von 200. Diese Parallelläufe erzeugten kaum Last (1,6 bis 2,5): das Verhalten unter einer Host-Last von 10 und mehr habe ich nicht gemessen. Die Angaben des Plans bestätigen sich.

## Gates (frisch, vom Verifier)

| Lauf | Ergebnis |
|---|---|
| `pruefung test` (aus dem kopierten Baum gebaut, leerer Go-Cache, `--cpus=4`) | ok, 552 s (Last 1,7 vor, 5,2 nach; dazwischen bis 11) |
| `pruefung race` (ebenso) | ok, 629 s (Last 5,2 vor, 3,7 nach) |
| `make a-check` | 0 Befunde (Hinweis auf 97 ungeprüfte Dateien unter `tools/gegenproben`, Bestand) |
| `make lint` | Exit 0 aus dem Zwischenspeicher; zusätzlich `lint.sh` mit leeren Caches im Container: `0 issues`, Exit 0 |
| `make groessenmarke` | ok, Messung 2118 von Marke 2700 |
| `make docs-check`, `make zeichen-check` | nach dem Schreiben dieses Reports gefahren; ein Commit des Reports setzt beide grün voraus |

Die bekannten Flakes `TestHandbuchZeitgrenzeUndWarten` und `TestAnfrageBodyBytes` traten in je einem Lauf nicht auf (kein Beleg, dass sie nicht auftreten).

**Messung.** `git diff --numstat ea3547be..HEAD` über `tests/lasttest/`: 2118 Einfügungen, 0 Löschungen; Produktcode 760, Tests 1358; Auslastung 0,78 der Marke; wie berichtet. **Abdeckung** des Pakets unter dem Detektor: 100,0 %, 0 von 364 Blöcken ohne Zähler. **Zeit des Pakets allein** (`go test -count=1`, warmer Bau, `--cpus=4`, je drei Läufe, abwechselnd; Last vor und nach jedem Lauf 3,7 bis 5,5): ohne Detektor 39,6 s, 38,8 s, 39,8 s; mit Detektor 81,4 s, 81,7 s, 81,1 s. Das bestätigt den Implementer (39,5 bis 40,7 s; 80,2 bis 81,6 s); der Trigger von 120 s ist nicht eingetreten.

## Mutationen (Beleg zu den Findings)

21 eigene Mutationen an Elementen, die Review und Plan nicht nennen, je gegen den zuständigen Test und bei Grün gegen alle Tests von Teil 2 und `TestPaketOhneZeit` (Nullprobe grün): 3 rot (10 s + 1 ns; Modus „letzte Zeile gewinnt“; naive Leerlauf-Frist, Nebengrund, V-3), 18 grün (fünf zu V-2, eine zu V-3, eine zu V-4, eine zu V-5, zwei zu V-6, vier zu V-1, eine zu V-7, eine zu V-10, zwei zu V-12). Dazu sechs Mutationen für die Raten (alle rot) und vier Proben in Wegwerfkopien. Treiber und Listen im Scratchpad unter `lasttest-auswertung-verifier/`, nicht im Repo; der Anwender der Ersetzungen ist eine Kopie des Programms des Implementers (gelesen).

## Auflagen

1. **Implementer, vor dem Merge:** V-1 bis V-6 je als Menge aus dem Wortlaut schließen und die genannte Mutation rot sehen: Probe über der Schwelle an vier Stellen (vor dem Warm-up, erster Tick, Mitte, nach dem Nachlauf), eine Zeile `Last` mit kleinster Probe ungleich der ersten und anderer Kernzahl; Nachlauf mit 0 fehlenden, zu vielen und Teilankunft vor Ablauf; Frist ab Aufruf bei verteilt eintreffenden Spans; Modus an Gesamtspans mit Kind; Zeile ohne `ereignis`; Gesamtspan ohne Kind hinter *W* + *N*; `Toleranz` bei ungültigem Lauf.
2. **Implementer, im selben Commit:** Negativ (9) berichtigen oder die Kurzprüfung angleichen (V-7; bei einer Verhaltensänderung vorher der Architect, `AGENTS.md` §3.12); V-9 und V-10 nachziehen.
3. **Architect:** Exit ohne gefahrene Betriebsart (V-8); die offene Frage nach dem Tick bei 660 s (Review I-1) bleibt.
4. **Planner:** Sendung an `slice-lasttest` eintragen, mit dem Gebrauch der Rückgabe von `warteAufSpans` und dem Fehlerzähler über alle Phasen (V-13); Ausgang des Risikos *Zeit des Pakets* (81 s) setzen; Closure-Punkte der DoD.

Eine zweite Verifikationsrunde verlangt dieser Report nicht: die Auflagen 1 und 2 sind am Diff und an den genannten Mutationen prüfbar.

## Nicht nachgemessen

`make gates` als Ganzes und die Gates außerhalb des Pakets (Beleg: Stempel); `make gegenproben` (Auftrag); die 215 Mutationen und die Raten des Implementers zu A1 und zum Sammler; die Raten unter hoher Host-Last; die Zeitangabe zu Negativ (9); die Mutation zu Negativ (7); die Prüfgrenzen von Negativ (5); ob die Zeile des gebauten Binary vom Leser gelesen wird (Teil 3); die Zeit von Teil 1 allein.

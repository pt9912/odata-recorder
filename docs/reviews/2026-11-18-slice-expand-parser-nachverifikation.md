# Nachverifikationsbericht: slice-expand-parser — 2026-11-18

**Art:** Nachverifikation der Fix-Runde nach der Verifikation (Bauen wir es richtig? — gegen die Klasse der Zusage, `AGENTS.md` §3.17), Modul 11. Nicht das Review, nicht die Validierung. Der Bericht der ersten Verifikation (`docs/reviews/2026-11-18-slice-expand-parser-verifikation.md`) bleibt unverändert; dieser ergänzt ihn nicht, er löst nur dessen Fix auf.

**Gegenstand:** Zweig `slice-expand-parser`, `HEAD` `8eb61157`. Fix in `ef038aa2` (neue Datei `expand_teile_test.go`, `expand_test.go`, Plan §3, §6) und `8eb61157` (Messpunkt, Plan §4). Ausgangslage: Verifikationsbericht `b667048e` (nicht bestanden, V-1 MEDIUM, V-2 LOW, V-3 LOW, V-4 und V-5 INFO).

**Bezug:** `LH-FA-16` (Unterpunkt a), `LH-RB-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md) (Schichten, nur Abgrenzung).

**Verträge, gegen die geprüft wurde:** `SPEC-262` (Nr. 3 d, 4 a bis e, 6, 8), `SPEC-160` Nr. 5 und 6, `SPEC-162` Nr. 2 und 3, `SPEC-122` Nr. 5 (Zeitgrenze), `AGENTS.md` §3.10, §3.15, §3.17, Plan `slice-expand-parser` §3 (*Fix-Runde nach der Verifikation*, Zeilen 11, 14, 16, 24), §4, §6.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

## Urteil

**Nicht bestanden.** 0 HIGH, 4 MEDIUM (N-1 bis N-4), 0 LOW, 3 INFO. Der Fix von V-1 und V-2 hält: die Klasse „Zahl der Teile, Zahl der Segmente, Länge des Segments“ ist gegen sieben eigene Mutationen an Elementen, die kein früherer Bericht nannte, rot, und die Prüfgrenze über 300 ist begründet. Der Code weicht an keiner Stelle von `SPEC-262` ab (eine Probe über 182 197 Eingaben, Orakel `exOrakel`, bestätigt es). Nicht gedeckt sind vier weitere Achsen derselben Klasse, die der Fix nicht kreuzt: die Stelle eines Elements im Wert jenseits von 5, die Form und Länge der Dubletten, die Folge verschiedener Segmentarten und die Größe der Elemente und des Werts an der Stelle in der Liste. Elf von zwanzig Mutationen dieser Achsen bleiben im ganzen Paket grün; mit den Probe-Generatoren (Abschnitt 4) sind alle rot.

## 1. Der Fix gegen die Klasse (V-1, V-2)

Die Menge folgt aus dem Wortlaut von `SPEC-262` Nr. 4 c und Nr. 6 (Zahl der Teile eines qualifizierten Namens, Zahl der Segmente, Länge, Zeichen, Struktur, Stellen). Die Tabelle des Plans (Zeile V-1: 107 440 + 22 940 + 1644 + 936 Eingaben) ist gegen die Zählung der Tests nachgerechnet (`exZaehlt` hält jede Menge; 820 × 10 × 3 × 4 + 300 × 3 × 2 × 4 + 340 × 4 + 40 × 3 × 4 = 107 440, Stufen 300 + 39 + 1 = 340).

Eigene Mutationen an Elementen, die weder Review noch Implementer noch die erste Verifikation nannten, gefahren mit `sh tools/harness/mutationslauf.sh` im Image `odata-recorder:race`, ohne Netz und Mount, unter `heavy.lock`, Klasse `andere` (deterministisch, keine Nebenläufigkeit):

| Menge | Mutation | Ergebnis | roter Test |
|---|---|---|---|
| Zahl der Teile | `nv21` (`SplitN` mit 5), `nv22` (`SplitN` mit 200) | rot, rot | `TestExpandSegmentTeile` |
| Zahl der Segmente | `nv23` (`SplitN` mit 2500), `nv24` (mehr als 4050 Segmente ungültig) | rot, rot | `TestExpandSegmentZahl` |
| Teil an einer Stelle | `nv25` (Teil mit Index 3001 ungeprüft) | **grün** | — (benannte Prüfgrenze: eine Stelle zwischen erster, mittlerer, letzter bei mehr als 40 Teilen; siehe I-2) |
| Zustand des Parsers | `nv26` (13 Elemente schreiben `expand`), `nv27` (30 Elemente) | rot, rot | `TestExpandParserUnveraendert` (V-2 gedeckt: die 32 Listenlängen) |

V-1 und V-2 sind gegen die Klasse gedeckt, soweit der Plan es als Prüfgrenze nennt.

## 2. Weitere Achsen der Klasse (Kreuzformen)

Achsen, die der Fix nicht kreuzt, aus dem Wortlaut gebildet: Stelle des Elements im Wert (`SPEC-262` Nr. 3 d: „gleich wo er steht“, Nr. 6: bis 32 Elemente) × Art; Dublette (Nr. 4 e: „bytegleich“) × Form × Länge; Folge der Segmentarten (Nr. 4 c: jedes Segment ein Name, ein qualifizierter Name oder `$ref`, `$count`, `$value`); Größe (Teile, Segmente, Bytes) × Stelle in der Liste × Zahl der Elemente; Länge des Werts (Nr. 6: „allein durch `rawQuery` begrenzt“) vor `(` und vor der Zählung. Dekodierung (`%2E`, `%2F`, `%2C`, `%24`) in anderen Stellungen ist gedeckt (Abschnitt 3).

Zwanzig Mutationen (`nv01` bis `nv20`), erster Lauf **ohne** Probe-Dateien im Baum:

| ID | Mutation | Achse | Ergebnis |
|---|---|---|---|
| nv01 | `*` ab Stelle 12 wird übergangen | Stelle × Art | **grün** |
| nv02 | Pfad ab Stelle 9 ist `wert` | Stelle × Art | rot-paket |
| nv03 | Name ab Stelle 20 fehlt im Ergebnis | Stelle × Art | rot-paket |
| nv04 | Dublette nach Länge und letztem Byte | Dublette × Form | rot-paket |
| nv05 | Dublette nach den ersten 16 Bytes | Dublette × Länge | **grün** |
| nv06 | Dublette nur bis zwei `/` im Element | Dublette × Form | **grün** |
| nv07 | `$ref` nach einem qualifizierten Segment ist `wert` | Folge der Segmentarten | **grün** |
| nv08 | `$value` nach `$count` ist `wert` | Folge der Segmentarten | **grün** |
| nv09 | Pfad mit mindestens 40 `/` ab Listenstelle 3 ist `wert` | Größe × Stelle | **grün** |
| nv10 | Element mit mindestens 40 `.` ab Listenstelle 1 ist `wert` | Größe × Stelle | **grün** |
| nv11 | Element über 500 Bytes bei mindestens 20 Elementen ist `wert` | Größe × Zahl | **grün** |
| nv12 | `(` nur in den ersten 2048 Bytes des Werts gesucht | Länge des Werts | **grün** |
| nv13 | Zahl der Elemente nur in den ersten 512 Bytes gezählt | Länge des Werts × Zahl | **grün** |
| nv14 | Dublette nach Schreibweise ab 65 Zeichen | Dublette × Länge | rot-paket |
| nv15 | Hex-Ziffer `f` (klein) ungültig | Dekodierung | rot-paket |
| nv16 | Segment nach `$ref` ist `wert` | Folge der Segmentarten | rot (`TestExpandSegmentZahl`) |
| nv17 | `ungueltig` an Stelle 8 bis 23 ist `501` | Stelle × Art | **grün** |
| nv18 | doppeltes `*` ab Stelle 11 nicht erkannt | Stelle × Dublette | rot (`TestExpandDoppeltesElementAbstand`) |
| nv19 | doppelter Pfad ab Stelle 11 nicht erkannt | Stelle × Dublette | rot (`TestExpandDoppeltesElementAbstand`) |
| nv20 | Element mit drei `.` bei sechs Segmenten ist `wert` | Teile × Segmente | rot (`TestExpandSegmentKreuz`) |

Summe: 4 rot, 5 rot-paket, 11 grün. Grün ist ein Verstoß gegen `SPEC-262`, den kein Test im Paket meldet; jede der elf Mutationen widerspricht dem Wortlaut (etwa nv06: `a/b/c,a/b/c` liefert `501` statt `wert`; nv07: `Ns.T/$ref` liefert `wert` statt `501`).

## 3. Dekodierung in anderen Stellungen, Rang mit solchen Namen

Sechs Mutationen am Decoder (`nd1` bis `nd6`): `%2F` am Ende des Werts, `%2E` am Anfang, `%2E` nach `%2E`, `%24` nach `/`, `%2e` ab Byte 100, `%2C` nach `/` je als Literal belassen. Ergebnis ohne Probe: 1 rot (`nd5`, `TestExpandSegmentTeile`), 5 rot-paket (andere `TestExpand*` und Dekodierungstests); keine grün. Die Probe P6 (420 Eingaben: 10 Escapes × 21 Stellungen × kurz und nach 120 Segmenten) hat das Ergebnis des Orakels. Rang mit solchen Namen (lange Namen, lange Pfade, Escapes) ist im Code ein Zweig je Element und von der Größe unabhängig; die Stellen jenseits von 5 sind N-1.

## 4. Probe-Paket aus Generatoren (zulässiges Mittel, §3.17)

Sieben Generatoren über die Klassen, Orakel `exOrakel` (kein Code des Lesers), als Testdateien im Baum des Verifiers gefahren und danach entfernt (`git status` sauber). Alle sind auf dem Stand `8eb61157` grün, der Code weicht also nirgends ab:

| Probe | Menge | Eingaben |
|---|---|---|
| P1 | Art (6) × jede Stelle 0 bis k-1 für k in {2, 3, 7, 12, 25, 32}; Paare {`*`, `1a`}, {Pfad, `1a`}, {`1a`, `*`}, {`*`, Pfad} an jedem Stellenpaar | 7582 |
| P2 | Dublette: 7 Längen und 15 Pfadformen × Abstand {0, 1, 5, 20}; verschiedene Elemente mit gleichem Präfix, gleichem Ende, gleicher Länge (Länge 2 bis 128, Pfade mit 17 bis 100 Segmenten) | 137 |
| P3 | alle Paare und Tripel der 11 Segmentarten von `exSegmentArten` | 1452 |
| P4 | Element (Pfad, Teile, Name von 128 Zeichen, Gemisch) mit 45, 100, 200 Gliedern an jeder Stelle von Listen mit 2 bis 32 Elementen, mit und ohne Verstoß am Nachbarn | 1968 |
| P5 | 31 bis 34 Elemente mit 1 bis 128 Zeichen, mit `(` am Ende; Pfade mit 100 bis 4000 Segmenten vor `(` | 88 |
| P6 | Escapes in 21 Stellungen (Abschnitt 3) | 420 |
| P7 | jede Zahl von Teilen und Segmenten von 301 bis 4090 an erster, mittlerer, letzter Stelle (die Lücken zwischen den Stufen) | 170 550 |

Mit P1 bis P6 im Baum sind alle elf grünen Mutationen aus Abschnitt 2 rot-paket (zweiter Lauf: 4 rot, 16 rot-paket, 0 grün). Die Dateien liegen nicht im Repo; sie sind eine Vorlage für den Fix von N-1 bis N-4.

## 5. Prüfgrenze über 300 und Laufzeit

**Grund und Folge der Prüfgrenze.** Der Plan nennt als Grund, der Aufwand je Eingabe wachse mit der Zahl. Gemessen mit P7 (alle Zahlen 301 bis 4090, drei Stellen, 170 550 Eingaben): **30,6 s ohne Detektor, 62,5 s mit Detektor** (ein Lauf je Variante, Container `--cpus=4` ohne Netz, Host-Last 12 bis 27 laut `/proc/loadavg`, `nproc` 3, Messender dieser Verifier). Das stützt den Grund: ein dichter Lauf kostet mehr als das Zwanzigfache der jetzigen Tests, ohne Detektor etwa so viel wie das ganze Paket heute mit dem Detektor. Die Folge im Betrieb (eine Regel, die nur eine Zahl zwischen zwei Stufen trifft und keine größere, bleibt ungefangen; ein Schwellenwert trifft 4090) ist richtig benannt; `nv25` ist ein Beispiel (grün) und `nv23`, `nv24` sind die Gegenbeispiele für Schwellenwerte (rot). Die Prüfgrenze ist nach dem Wortlaut von §3.17 ein akzeptiertes Negativ mit Grund und Folge; kein Befund.

**Laufzeit von `TestExpand*`** (Test-Binary einmal gebaut, einmal gefahren, `-test.count=1`; Container `--cpus=4`, ohne Netz; Host-Last 6,2 bis 6,6 bei `nproc` 3; Messender dieser Verifier, eine Messung, Größenordnung):

| Lauf | ohne Detektor | mit Detektor |
|---|---|---|
| `TestExpand*` | 1,3 s | 8,7 s |
| ganzes Paket `urlparser` | 4,3 s | 26,3 s |

Die Zeitgrenze von `make test-race` ist die Standardgrenze von `go test`, 10 Minuten je Paket (`SPEC-122` Nr. 5). Das Paket braucht mit dem Detektor 26,3 s, also rund ein Dreiundzwanzigstel der Grenze; die neuen Tests (`expand_teile_test.go`) tragen davon einen Teil. Die Zeitgrenze ist nicht gefährdet. Auch ein Fix von N-1 bis N-4 in der Größe der Proben P1 bis P6 (zusammen rund 0,1 s ohne Detektor) bleibt weit darunter; nur die dichte Variante P7 (62,5 s mit Detektor) würde das Paket mehr als verdreifachen und gehört deshalb nicht in die Schleife.

## 6. Plan und Kommentare

- Plan §3 *Fix-Runde nach der Verifikation*: Mengen, Tabelle, Prüfgrenze und die 18 Mutationen sind stimmig mit dem Code (Köpfe der vier Tests nennen ihre Mengen, die Zählungen stimmen). Zeile 16 (`TestExpandRangImWert`, Stellen 0 bis 5) und die Prüfgrenze „höchstens vier leere Elemente und drei Namen neben `*`“ nennen die Stellen jenseits von 5 und die Achsen von N-1 nicht als akzeptiertes Negativ mit Grund und Folge; der Fix-Runde-Bericht F-4 hat die Menge „jede Stelle 0 bis 5“ als Klasse behandelt, `SPEC-262` Nr. 3 d sagt „gleich wo er steht“ bei bis zu 32 Stellen.
- §4: `make groessenmarke` meldet Messung 2004 von Marke 3500, Auslastung 0,57, wie der Plan (Handoff am Stand `ef038aa2`); die Zeile `8eb61157` trägt nur den Messpunkt.
- Rohbytes (§3.18): dieser Bericht ist nach dem Schreiben mit `LC_ALL=C grep -nP` auf BOM, NBSP und die Folge Backslash, `u`, vier Hex-Ziffern gegengelesen (keine Fundstelle).

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Klasse |
|---|---|---|---|---|---|
| N-1 | MEDIUM | Die Stelle eines Elements im Wert ist nur an 0 bis 5 (Paare) und an erster und letzter Stelle der 32er Liste geführt; ein Verstoß der Klasse `ungueltig` an Stelle 8 bis 23, der als `501` gelesen wird (`nv17`), und ein `*` ab Stelle 12, das übergangen wird (`nv01`), bleiben im ganzen Paket grün, obwohl `SPEC-262` Nr. 3 d „gleich wo er steht“ sagt und Nr. 6 bis 32 Elemente zulässt. Der Plan nennt dafür keine Prüfgrenze mit Grund und Folge. Folge im Betrieb: ein Umbau, der Stellen jenseits von 5 anders behandelt, liefert `501` statt `400` oder umgekehrt. | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 3 d, 6 | `expand_test.go` · `TestExpandRangImWert`, `expand_grenzen_test.go` · `TestExpandStern`; Plan §3 Zeile 16, F-4 | Menge unbeschränkt bis 32, nicht als Schleife über jede Stelle geführt; Probe P1 (7582) deckt sie |
| N-2 | MEDIUM | Dubletten sind nur für kurze Namen, `*` und Pfade bis zu zwei `/` geführt: `a/b/c,a/b/c` (Pfad mit drei und mehr Segmenten, `nv06`) wird ohne Fänger `501` statt `wert`; zwei verschiedene Elemente mit gemeinsamen 16 Bytes Präfix (`nv05`) gelten ohne Fänger als Dublette. `SPEC-262` Nr. 4 e sagt „bytegleich“ und meint jede Form und Länge, auch die Gegenseite (verschieden). | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 4 e | `expand_grenzen_test.go` · `TestExpandDoppeltesElement`, `…Abstand`; Plan §3 Zeile 14, F-2 | Menge Form × Länge × Abstand, beide Seiten der Gleichheit; Probe P2 (137) deckt sie |
| N-3 | MEDIUM | Die Folge verschiedener Segmentarten ist nicht geführt: die Tests setzen eine Art an eine Stelle zwischen `N` oder alle Segmente von einer Art; `Ns.T/$ref` (`nv07`) und `$count/$value` (`nv08`) als `wert` bleiben grün, obwohl `SPEC-262` Nr. 4 c jedes Segment einzeln zulässt. Gefangen sind nur Folgen nach `$ref` (`nv16`). | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 4 c | `expand_segment_test.go`, `expand_teile_test.go` · `TestExpandSegmentZahl`; Plan §3 Zeile 11 | Menge Art × Art (× Art), 11 Arten, 1452 Folgen bis Länge 3; Probe P3 deckt sie |
| N-4 | MEDIUM | Die Größe eines Elements (Zahl der Teile, Zahl der Segmente, Bytes) ist nur für ein einzelnes Segment oder an den Listenstellen 0 bis 2 mit kurzen Nachbarn geführt, die Länge des Werts nur mit kurzen Namen; Mutationen wie ein Pfad mit 40 oder mehr Segmenten ab Listenstelle 3 (`nv09`), ein Element mit 40 oder mehr Teilen ab Stelle 1 (`nv10`), ein Element über 500 Bytes bei 20 Elementen (`nv11`), `(` nur in den ersten 2048 Bytes (`nv12`) und die Zählung nur in den ersten 512 Bytes (`nv13`: 33 lange Namen sind ein Ergebnis statt `zu_viele`) bleiben grün. `SPEC-262` Nr. 6 sagt „die Länge des Werts ist allein durch `rawQuery` begrenzt“. Die Kreuzform Teile × Segmente × Elementzahl × Listenstelle ist nicht geprüft und nicht als Prüfgrenze benannt. | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 3 b, 3 c, 6 | `expand_teile_test.go` · `TestExpandSegmentZahl` (Listen mit 3 Elementen), `TestExpandSegmentKreuz`; `expand_grenzen_test.go` · `TestExpandZahl` (kurze Namen); Plan §3 V-1 | Menge Größe × Stelle × Elementzahl × Länge des Werts; Probe P4 (1968) und P5 (88) decken sie |
| I-1 | INFO | Die Prüfgrenze über 300 (Stufen 397 bis 4083 in 97er Schritten und 4090) ist begründet: eine dichte Probe aller Zahlen 301 bis 4090 braucht 30,6 s ohne und 62,5 s mit Detektor (Last 12 bis 27). Kein Befund; der Plan nennt „16 Sekunden“ für jede Stelle, eine andere Menge. | `AGENTS.md` §3.17 | Plan §3 Prüfgrenze | akzeptiertes Negativ mit Grund und Folge |
| I-2 | INFO | `nv25` (Teil mit Index 3001 ungeprüft) bleibt grün: eine Stelle zwischen erster, mittlerer und letzter bei mehr als 40 Teilen, im Plan als Prüfgrenze benannt. | `AGENTS.md` §3.17 | Plan §3 Prüfgrenze | akzeptiertes Negativ |
| I-3 | INFO | Zeitaufwand von `TestExpand*`: 1,3 s ohne, 8,7 s mit Detektor (Paket 4,3 s und 26,3 s; Last 6,2 bis 6,6, `nproc` 3, eine Messung); die Zeitgrenze von 10 Minuten (`SPEC-122` Nr. 5) ist nicht gefährdet. | `SPEC-122` Nr. 5 | `make test-race` | — |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 |
| LOW | 0 |
| INFO | 3 |

## Übergabe an den Planner

Merge-blockierend sind N-1 bis N-4: je Befund entweder eine Schleife über die Menge (die Proben P1 bis P6 mit ihren Zahlen sind die Vorlage; zusammen rund 0,1 s ohne Detektor, die Zeitgrenze trägt es) oder ein akzeptiertes Negativ mit Grund und Folge im Plan. Der Bericht des Fixes nennt dann je Befund die Menge, die Zahl ihrer Elemente und eine Mutation an einem Element, das dieser Bericht nicht nannte (etwa eine Stelle zwischen 6 und 30 für N-1, eine Folge `$count/$ref` für N-3). Die Finding-Klasse dieses Laufs: *Menge als Achse geführt, Kreuzung der Achsen nicht (Stelle × Art, Form × Länge, Art × Art, Größe × Stelle)*. Der Verifikationsbericht `b667048e` bleibt als Ausgangslage unverändert.

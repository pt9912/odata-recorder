# Verifikations-Report 3: slice-odata-codec-json — 2026-10-24

**Rolle:** Verifier, dritte Verifikation, frischer Kontext (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-126`, `AGENTS.md` §3.9 bis §3.17). Kein Review, keine Reparatur.

**Gegenstand:** `slice-odata-codec-json` (Welle `welle-record-replay`), HEAD `f47a9e3`; Basis der Marke `bc05c62`. Geprüft: zweite Verifikation `2d40873` (W-1, G-1, K-1, E-1, P-1, D-1, N-1, N-2), Entscheidungen des Architect `e04c240`, dritte Fix-Runde `f3ea895` (Suite, `tools/gegenproben`, Plan), Stempel `f47a9e3`. Der Produktcode ist seit `0acc49a` unverändert (`git diff --stat 0acc49a..HEAD -- src/adapters src/hexagon/model` leer); die dritte Runde ändert nur die Suite (`outboundtest/`), `tools/gegenproben`, den Plan und den Stempel.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-24

**Methode.** Belege statt Behauptungen. Alle Mutationen in Wegwerfkopien außerhalb des Repos (Quelle `:ro`, `--network none`, `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container, Gates nie parallel zu Mutationen; Limit-lose Mutationen gab es nicht). Eine neue Basis aus `git archive HEAD` (Scratchpad `…/scratchpad/vf4/`, flüchtig: `base/` der Abzug, `runner/*.txt` die Mutationen, `probe/` das Probe-Paket, `out/` die Ausgaben). Die 84 Mutationen der Verifier 1 und 2 laufen unverändert gegen den neuen Stand; dazu 64 eigene Mutationen an Elementen, die weder Verifier 1 und 2 noch der Implementer nannten (`w1`, `w2`, `w3`), und ein erweitertes Probe-Paket. Das Repo ist bis auf diesen Report unverändert (`.claude/agents/architect.md` nicht angefasst). Fremde Container (`pgr-it-*`, `u-boot-devcontainer-*`, ein `d-check`-Lauf) liefen zeitweise auf dem Host und wurden nicht berührt.

---

## 1. Ergebnis in einer Tabelle: die Befunde der zweiten Verifikation gegen die Klasse

| Befund | Klasse (Wortlaut `SPEC-126`) und im Lauf gezählte Menge | Eigene Mutationen an Elementen, die der Befund nicht nannte | Urteil |
|---|---|---|---|
| W-1 (MEDIUM) `Close` durch die Hülle verdeckt | Nr. 2 „von `src` allein `Read`“, die acht Methoden von Menge O (`Close`, `WriteTo`, `ReadByte`, `UnreadByte`, `ReadRune`, `UnreadRune`, `Seek`, `ReadAt`): A5 zählt `oZahl` = 8 und hält bei einer anderen Zahl an; ich habe die Zahl im Lauf gelesen (`Menge O: 8 Methoden`, 22 203 Fälle der Suite + 2 mit Deckel) | je Methode eine Typzusicherung in `lies` (**8 von 8 rot**, in A5 allein und in der ganzen Suite in je **48 von 50 Zeilen**; A4, I2, J2 am Zähler `jsuite.optional`; K3 und K4 rufen den Leser nicht); dazu `io.Copy(io.Discard, src)` (Weg über `WriterTo`, rot), `Close` an vier Stellen (vor der Argumentprüfung, nach `io.EOF`, bei `zu_gross`, bei einem Lesefehler: **4 von 4 rot**, die beiden letzten in den 7 bzw. 3 Zeilen, die den Pfad erreichen) | **gelöst** |
| W-1, Hülle selbst | `pruefeHuelle`: jeder der 8 Aufrufe zählt genau seine Methode, `Close` erreicht `inner` | Hülle zählt `Close` nicht (SH1), reicht `Close` nicht durch (SH2), `ReadAt` zählt `Seek` (SH3), `Seek` zählt nichts (SH4), Hülle ohne Zählung und ohne Durchreichen zusammen mit der Mutation `Close` im Leser (SH5): **5 von 5 rot** aus dem benannten Grund (`Hülle: Aufruf von … zählt …`); Zähler je Zeile entfernt zusammen mit `Seek` im Leser (SH7): rot in 45 Zeilen; **SH6 grün** (Hülle durchreicht nicht, zählt nicht *und* `pruefeHuelle` entfernt, zusammen mit der Mutation im Leser: zwei Mutationen an der Suite, `pruefeHuelle` ist der einzige Wächter der Hülle; der Implementer nennt es) | **gelöst** |
| G-1 (MEDIUM) Gegenproben an der Zeitgrenze | die 4 Fälle 25 bis 28 laufen im Wegwerfmodul; ich habe drei Läufe von `make gegenproben` selbst gefahren (siehe 4) | `TestTestFaelleLegenEinWegwerfmodulAn`: der Fall `test/basis`, `test/nicht-formatiert`, `test/vet-befund`, `test/roter-test` je ohne `Wegwerfmodul: true`: **4 von 4 rot** (`Wegwerfmodul nicht gesetzt, der Fall kopiert die Eingaben`; der Befund nannte zwei, der Fänger fängt alle vier) | **gelöst** |
| K-1 (LOW) K1 ab U+0080 nur in `ParseEntity` | Nr. 14 „kein anderes Zeichen als `~` und `/` wird ersetzt“, beide Methoden: im Lauf gezählt **2 223 872** (2 × 1 111 936); die 384 ASCII-Namen je Methode in `jfK1`; K1 allein 9,5 s | `~2` für U+0100, U+00E9, U+0080, U+D7FF, U+FFFD, U+1FFFF, U+10FFFF (je in beiden Pfaden) und für den Bereich U+2000 bis U+206F: **8 von 8 rot** aus dem benannten Grund; der Verifier-Fall NK1c (U+10FFFE nur im Pfad der Collection) jetzt rot (vorher grün); ein Bereich von 42 000 Zeichen (KRange4) endet an der Zeitgrenze, weil der Replacer je Aufruf neu gebaut wird: Aufbau der Mutation, kein Befund | **gelöst** |
| E-1 (LOW) „jeder andere Fehler“ unbeschränkt | Nr. 2: Menge M2, 12 Fehlerarten × (279 + 517) = **9552** im Lauf gezählt; Prüfgrenze mit Grund und Folge im Plan (*Akzeptierte Negative*, „Die Fehlerarten von `src`“) | „bei diesem Fehler noch einmal lesen“ für 11 Arten und für jede Art von Fehler zugleich (`io.ErrUnexpectedEOF`, `ErrNoProgress`, `ErrClosedPipe`, `ErrShortBuffer`, `ErrShortWrite`, `context.Canceled`, `DeadlineExceeded`, `os.ErrDeadlineExceeded`, `os.ErrClosed`, umhüllter `io.EOF`, Typ mit `Timeout()`, Typ mit `Temporary()`; dazu alle): **13 von 13 rot** („`Fehler <nil>, erwartet den Fehler von src`“, „N Reads nach dem Fehler“); vier Mutationen „Fehler gilt als Ende“ (`errors.Is` auf `ErrShortWrite`, `os.ErrClosed`, umhüllter `io.EOF`, Typ mit `Timeout()`): **4 von 4 rot**. Die Mutationen auf `Timeout()`/`Temporary()` werden schon an `context.DeadlineExceeded` rot (der Typ hat dieselben Methoden), der Zeigertyp allein ist damit nicht isoliert; die Klasse „Typ“ deckt beides | **gelöst** (Prüfgrenze benannt) |
| P-1 (LOW) P5 nur bis `grenze` 393 217 | Nr. 5 „für jeden Wert von `grenze`“: Menge P6, 6 × 3 × 2 × 2 = **72** (unter dem Detektor 24: `make test-race` grün) | Grenze ab 3, 5, 7, 9, 12, 15 MiB gelesenen Bytes nicht mehr durchgesetzt: **6 von 6 rot** (`B5 P6 … Bytes geliefert, höchstens grenze+1`); Anforderung ab 3, 12, 16 MiB ein Byte zu groß: **3 von 3 rot**; ab 16 MiB + 1 (PT16M1: **unerreichbar**, der letzte Read liegt vor der Schwelle, kein Mutant), ab 17 MiB und ab 32 MiB grün: das ist das benannte Negativ (`grenze` über 16 MiB + 1) | **gelöst** (Prüfgrenze benannt, Wortlaut siehe V3-2) |
| D-1 (LOW) Abweichungen nur im Bericht | Plan §3 *Abweichungen und offene Punkte* und §6 (Zeile „Abweichungen der dritten Fix-Runde“ mit Ausgang und Adresse); Zeile A3 nennt, was `fehlerStrom` tut; Negativ „Proben unter dem Detektor“ mit Grund und Folge | gelesen | **gelöst** |
| N-1 (INFO) | Negativ in zwei Richtungen getrennt | NU1, NU2 (Annahme von `C5` als inneres Folgebyte) bleiben grün, wie das Negativ es sagt; VM5 rot | **gelöst** |
| N-2 (INFO) | Hülle verdeckt `Close` und jede optionale Schnittstelle | siehe W-1 | **gelöst** |

Von den 84 Mutationen der Verifier 1 und 2 (`v1` bis `v9`, `vold`) sind **79 rot aus dem benannten Grund**, 2 grün als benanntes Negativ (NU1, NU2), 3 **äquivalent** (NK1b, NK2b, NK3b: `len(weitere) > 1` trifft allein den Aufruf `fehler(model.DatenStruktur, "value", index)`, dessen Segmente nie das ersetzte Zeichen tragen). NE2 und NE3 waren in meinem ersten Lauf an der Zeitgrenze von 120 s unter Last (zwei Arbeiter, Last 9 bis 12) und sind mit 400 s einzeln **rot** (A3). Die sieben Mutationen, die in der zweiten Verifikation nur das Probe-Paket fing (PE2, PE3, PP1, PP1b, PK1c, PC1, PZ1), sind in der Suite jetzt **7 von 7 rot** (A3, B5, K1, A5, B1).

## 2. DoD und Plan

| DoD-Punkt (Plan §2) | Befund |
|---|---|
| 1 bis 3 Port, Schlüssel, Nachweise | erfüllt: Produktcode unverändert; in rund 1,62 Mio. Probe-Läufen dieser Runde (siehe 3) und den 22 Mio. des Verifiers 1 ohne Abweichung |
| 4 `make gates` grün | erfüllt, siehe 4 |
| 5 Review, Report liegt vor | erfüllt (`64b7cc9`); die Verifikationen sind `2b50f3d`, `2d40873` und dieser Report |
| 6 Doku-Update | erfüllt (`SPEC-126`, `SPEC-102`; die Zahlen 8, 9552, 72/24, 2 223 872 und 1 112 320 stimmen in Plan, Zeilen A3, A5, B5, K1, Raster und Lauf überein) |
| 7 bis 10 Closure, Register, Risiken, Paarungen | noch offen (§7 „Wird bei der Closure gefüllt“, §6 alle Risiken *offen bis zur Closure*); Sache des Planners |

**Plan gegen Code (`AGENTS.md` §3.9).** Dateitabelle (`outboundtest/` mit `fortschrittWaechter`, Menge O, `pruefeHuelle`, `jsuite.optional`, `endlichStrom`, `drossel`; `tools/gegenproben` und `SPEC-102` mit dem Test; Stempel) und die Zeilen A3, A5, B5, K1 stimmen mit dem Code überein. **Tabelle gegen Vertrag:** Nr. 2 (Menge O, Fehlerarten), Nr. 5 (Grenze), Nr. 14 (Zeichen) haben je eine Zeile mit Menge, Zahl und Fänger als Schleife; die Zahlen zählt der Test selbst. Frist von 120 s je Fall und Parallelität in `tools/gegenproben/main.go` unverändert (`f3ea895` ändert nur `faelle.go` und `main_test.go`), kein Gate gelockert (`AGENTS.md` §3.6; [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md): die Gegenproben liegen außerhalb der Gate-Kette).

**Die Abweichungen (1) bis (4) des Implementers (Plan §3).**
1. *Deckel `b6Deckel` in P6 bei Länge `grenze + 1`:* **zulässige Konkretisierung.** Der Wortlaut des Architect (Länge `grenze` und `grenze + 1`, Zähler höchstens `grenze + 1`, Mutation „Anforderung ab 512 KiB“) lässt die Mutation „Anforderung nicht auf `rest + 1` begrenzt“ ohne den Deckel unentdeckt (der Body hat dann genau `grenze + 1` Bytes, mehr liefert `src` nicht); der Deckel ändert weder die 72 Läufe noch die erwarteten Ergebnisse (ein richtiger Leser liest genau `grenze + 1`), folgt dem Muster `a5Deckel`, und die Zeile B5 nennt ihn mit Grund. Die Mutationen PW3M, PW12M, PW16M (ein Byte zu viel angefordert) werden an ihm rot. Der Architect muss nichts nachziehen; er bestätigt den Punkt in einer Zeile (§6 nennt ihn als Adresse).
2. *Zähler `jsuite.optional` je Zeile:* nötig und im Geist der Entscheidung (A4, I2, J2 sehen den Fehler des Aufrufs nicht an; SH7 und die 8 Mutationen belegen, dass alle drei am Zähler rot werden).
3. *`pruefeHuelle`:* nötig, SH1 bis SH5 belegen es; SH6 zeigt, dass er der einzige Wächter der Hülle ist.
4. *Keine Zeit je Fall der Gegenproben:* das Werkzeug meldet nur `ok` je Fall; die Zeit des Schritts aller Fälle ist die obere Grenze für den längsten (57,2 bis 58,5 s, siehe 4). Das erfüllt den Wortlaut „Zeit des längsten Falls“ des Architect nicht streng, ersetzt ihn aber durch eine Zahl, die ihn umschließt; eine Zeit je Fall wäre eine Werkzeug-Änderung.

**Größe (`AGENTS.md` §3.14).** `git diff --numstat bc05c62..HEAD` über die Pfade der Marke (`src/hexagon/model/`, `src/hexagon/ports/outbound/`, `src/adapters/driven/odata/jsoncodec/`, `src/bootstrap/`, `src/hexagon/services/session/beenden_test.go`, `.a-check.yml`): **7236** Zeilen (7234 Einfügungen, 2 Löschungen), davon Produktcode **1640** (ohne `_test.go`, `outboundtest/`, `testdata/`), Tests und Suite **5596**; gleich dem Bericht des Implementers. Marke 4500: Überschreitung 2736 Zeilen (60,8 %), in §6 *Größe* ehrlich mit Zahl, Aufteilung (358 Zeilen der Runde: Hülle 146, P6 100, K1 59, M2 und A5 53) und Grund ausgewiesen; die Marke setzt der Planner (§3.14). Getrennt, außerhalb der Metrik: `tools/gegenproben` **57** (45 Einfügungen, 12 Löschungen; `faelle.go` 39, `main_test.go` 18); `harness/` **15** gegen `bc05c62` (davon der Stempel 2, der Rest die Änderung des Architect an `SPEC-102`; der Implementer hat nur den Stempel geändert). §6 nennt beides getrennt.

## 3. Probe-Paket aus Generatoren (Orakel unabhängig vom Leser)

Das Paket liegt im Scratchpad `…/scratchpad/vf4/probe/` (`v3probe_test.go`, `v4probe_test.go`, `probe_util_test.go`, `par_test.go`; gefahren als Paket `src/revtest` in einer Kopie). Der Leser nutzt `encoding/json` nicht; `json.Valid`, `utf8` stehen nur im Probe-Paket. Alle Läufe gegen HEAD **ohne Abweichung am Code**:

| Test | Klasse | Läufe |
|---|---|---|
| `TestV3Stromende` | jedes echte Präfix von 907 Dokumenten × 6 Endungen × 2 Methoden, Orakel `json.Valid` | 189 212 |
| `TestV3Grenzlagen` | Grenzlagen B5 mit anderen Puffern als 2^k (24 Vielfache von 32 768, 6 andere Basen, ±3, Stücke, `0, nil`, `io.EOF` im letzten Read) | 22 896 |
| `TestV3Endlos` | Strom aus Leerraum ohne Ende, `grenze` von 1000 bis 64 MiB | 92 |
| `TestV3FehlerExotisch` | 16 Fehlerarten (darunter `syscall.ECONNRESET`, `*net.OpError`, `*url.Error`) × jeder Index × Körper | 4 480 |
| `TestV3PointerCollection` | jedes Zeichen U+0080 bis U+10FFFF als Name im Pfad einer Collection | 1 111 936 |
| `TestV3Close`, `TestV3ZaehleE` | `Close` ohne Hülle; Nachzählung der Menge E (104 Präfixe, 192 Körper je Methode) | 12 + 1 |
| **`TestV4VollerStrom` (neu)** | **Differenztest über jede Methode eines Stroms:** ein Strom mit Read, den 8 Methoden von Menge O, der Schreibseite von `io` (`Write`, `WriteString`, `WriteByte`, `ReadFrom`, `WriteAt`) und `Len`, `Size`, `Peek`, `Buffered`, `Discard`, `Name`, `Stat`, `Fd`, `SetReadDeadline`, `Truncate`, `Sync`, `Reset`, `Bytes`, `String` (30 Methoden, jede zählt ihren Aufruf) gegen einen Strom mit Read allein (Orakel: der Strom selbst); 9 964 Dokumente (die 907 Dokumente, jedes dritte Präfix, Randfälle) × 2 Methoden × 3 Zerlegungen (ganz, 1, 7) × 5 `grenze` (1, L−1, L, L+1, 1 GiB); gleiche Ergebnisse, kein Aufruf außer `Read` | 293 454 |

Zusammen **1 622 083** Läufe. **Gegen die Mutationen:** `TestV4VollerStrom` ist rot gegen 18 von 19 Mutationen aus `w1` (alle acht Methoden, `io.Copy`, die Schreibseite `io.Writer`, `io.StringWriter`, `io.ByteWriter`, `Len`, `Peek`, `Size`, `Close` an drei Stellen), grün gegen `Close` im Lesefehler-Pfad (der Strom liefert dort nie einen Lesefehler; die Suite fängt diese Mutation in 3 Zeilen). Die Schreibseite von `io` und `Len`, `Size`, `Peek` sind genau die Mutationen, die die Suite (Hülle) nicht fängt (siehe V3-1); das Probe-Paket fängt sie.

## 4. Gates, Laufzeiten, Last

Alle Läufe am Stand `f47a9e3`, Repo unverändert (vor dem Schreiben dieses Reports).

| Sensor | Ergebnis | Laufzeit |
|---|---|---|
| `make gates` | Rückgabecode 0 (Docker-Cache der Stufen, darum kurz) | 11 s |
| `make lint`, `make test-race`, `make test` | Rückgabecode 0 (Cache) | je 1 s |
| Stufe `test` frisch (`docker build --no-cache-filter test --target test`) | grün | 88 s |
| Stufe `race` frisch (`--no-cache-filter race`, Image `golang:1.27.1-trixie`) | grün | 124 s |
| Stufe `lint` frisch (`--no-cache-filter lint`) | grün | 31 s |
| Suite A3, A5, B5, K1 (`go test -v`, `--cpus=4`) | grün; A3 0,13 s, A5 0,20 s, B5 1,55 s, K1 9,45 s | 11,3 s |
| `make gegenproben` (je Lauf erzwungen über `GEGENPROBEN_STEMPEL=<Pfad im Scratchpad>`, der Stempel des Repos blieb unberührt) | **Lauf 1: 37 Fälle, 0 rot; Lauf 2: 37 Fälle, 0 rot; Lauf 3: 37 Fälle, 0 rot** | Wanduhr 88 s, 89 s, 87 s; **Zeit des Schritts aller Fälle 57,9 s, 58,5 s, 57,2 s** gegen die Frist von 120 s je Fall |
| `make gegenproben` am Repo | „Eingaben unverändert seit dem letzten grünen Lauf (0b2a987d…)“; der Hash der drei erzwungenen Läufe ist derselbe wie der Stempel `f47a9e3` | 0,4 s |

Last (`/proc/loadavg`, 20 Kerne): vor den drei Läufen 5,76 / 5,74 / 5,79 (1 Minute), im Lauf auf 51 bis 62 steigend (die Fälle laufen mit `NumCPU()` zugleich), danach abklingend. In der zweiten Verifikation lagen dieselben Schritte bei 117,8 bis 120,6 s, Wanduhr 130 bis 151 s, zwei von drei Läufen rot. Der Abstand zur Frist beträgt jetzt rund 62 s. Den längsten Fall einzeln misst das Werkzeug nicht (Abweichung 4).

**Test des Fängers `TestTestFaelleLegenEinWegwerfmodulAn`** (Wegwerfkopie von `tools/gegenproben`, Go-Image, ohne Netz): Kontrolle `ok`; ohne `Wegwerfmodul: true` bei `test/basis`, `test/nicht-formatiert`, `test/vet-befund`, `test/roter-test` je **rot** (`<Fall>: Wegwerfmodul nicht gesetzt, der Fall kopiert die Eingaben`). Eine Mutation von `tools/pruefung` (Schritt `go vet` entfernt) endet schon am eigenen Unit-Test von `pruefung` in der Stufe `pruefung`, vor den Gegenproben; die Empfindlichkeit der Fälle 25 bis 28 gegen den Eingriff, den sie meinen, habe ich deshalb nicht unabhängig von den Unit-Tests belegt (der Inhalt der Fälle ist gegenüber der Fassung vor der Änderung gleich, allein der Ort ändert sich).

**Raten (Stichprobe, Messender: Verifier; jeder Prozess `-test.count=1`; Mutationen aus `runner/rate.txt` des Verifiers 2):**

| Zeile | Mutation | Prozesse rot |
|---|---|---|
| J1 | Zähler des Pakets in `lies`, Detektor (Image `golang:1.27.1-trixie`) | 10 von 10 (`WARNING: DATA RACE`) |
| J2 | Puffer des Pakets unter Sperre, ohne Detektor | 10 von 10; 3 Prozesse mit `-test.count=20`: 60 von 60 Läufen |
| I1 | `Properties` über Map, ohne Detektor | 10 von 10; 3 Prozesse mit `-test.count=20`: 60 von 60 Läufen |

Die Zahlen des Implementers sind nicht widerlegt. Für die Fänger der dritten Runde (Mengen O, M2, P6, K1, Fall 25 bis 28) gibt es keine Rate: Schleifen über feste Mengen ohne Zufall und ohne Nebenläufigkeit im geprüften Code; die Hülle hält ihren Zustand je Aufruf (J1 fährt sie aus 16 Goroutinen unter dem Detektor).

## 5. Befunde

| ID | Schwere | Befund | Zuständig |
|---|---|---|---|
| V3-1 | LOW | **Das Negativ „Methoden außerhalb von `io`“ (Zeile A5, W-1) schneidet an der falschen Stelle.** Menge O sind die sieben Leser-Schnittstellen aus der Liste von `SPEC-126` Nr. 2 plus `Close`; die Schreibseite des Pakets `io` (`io.Writer`, `io.StringWriter`, `io.ByteWriter`, `io.ReaderFrom`, `io.WriterAt`) liegt *im* Paket `io`, steht nicht in Menge O und nicht im Negativ. Mutationen `src.(io.Writer).Write`, `src.(io.StringWriter).WriteString`, `src.(io.ByteWriter).WriteByte` im Leser bleiben in der **ganzen Suite grün** (die Hülle trägt sie nicht); `Len()`, `Size()`, `Peek` ebenso, wie das Negativ es sagt. Die Spezifikation schließt die Liste (sieben Schnittstellen und `Close`), der Satz „allein `Read`“ ist weiter; ein Leser, der in `src` schriebe, wäre ein Verstoß gegen den Satz und kein Verstoß gegen die Liste. Folge im Betrieb: gering (ein Strom, der zugleich `io.Writer` ist, etwa eine Verbindung); das Probe-Paket (`TestV4VollerStrom`, 293 454 Läufe) fängt alle diese Formen. Fix: das Negativ nennt die Schreibseite von `io` ausdrücklich (oder Menge O wächst um sie); keine Änderung am Code. | Architect (Wortlaut des Negativs), Planner |
| V3-2 | INFO | Der Wortlaut des Negativs zu P-1 („eine Mutation, die die Grenze erst ab 32 MiB nicht mehr durchsetzt, bliebe unbemerkt“) untertreibt: die Menge P6 endet bei `grenze` 16 MiB + 1, grün bleiben auch Schwellen ab 16 MiB + 2 (PT17M ab 17 MiB gemessen, PT32M). Der Satz davor („`grenze` von 16 MiB + 2 bis 1 GiB … nicht gefahren“) ist richtig; die Folge sollte dieselbe Schwelle nennen. | Planner / Architect (Wortlaut) |
| V3-3 | INFO | SPEC-102 nennt „Die `go`-Aufrufe der Fälle 25 bis 28 kosten bei kaltem `GOCACHE` bis zu 27 s je Fall“; das war die Messung mit den Paketen unter `src/bootstrap`. Die Fälle fahren jetzt ein Modul aus einem Paket; die Aussage ist als obere Grenze nicht falsch, aber nicht neu gemessen, und das Zeitbudget (≤ 120 s im Ganzen, 87 bis 89 s gemessen) hält. | Architect (Messung nachziehen oder Satz streichen) |
| V3-4 | INFO | Abweichung 1 (Deckel in P6) und Abweichung 4 (keine Zeit je Fall): wie in 2 bewertet, zulässig; §6 trägt sie als *offen bis zur Closure* mit Adresse Architect, der Architect bestätigt sie in einer Zeile, ohne Änderung. | Architect (bestätigen), Planner (Ausgang) |
| V3-5 | INFO | Offen als benanntes Negativ, ohne Änderung: NU1 und NU2 (Annahme eines ungültigen inneren Fortsetzungsbytes), Mutationen auf Fehlerwerte und Schwellen außerhalb der Mengen (`syscall.Errno`, `*net.OpError`, `grenze` über 16 MiB + 1, Methoden außerhalb von `io`). Äquivalent (kein Befund): NK1b, NK2b, NK3b, PT16M1. | Closure (Negativ-Liste) |

Gültige Seite: **keine Verletzung der Spezifikation, kein HIGH, kein MEDIUM.** Der Produktcode ist in allen gefahrenen Klassen richtig.

## 6. Register (nichts geschrieben): Evidence-Kandidaten für den Lese-Schritt der Closure

Unter `docs/plan/planning/observations/BEO-REPO/` gibt es zu diesem Slice keine Evidence-Datei. Vorzumerken (Planner entscheidet):

1. `fix-deckt-die-genannte-mutation-nicht-die-klasse` (verkörpert in `AGENTS.md` §3.17, drei Evidence-Dateien): vierte Runde dieses Slice; diesmal **keine Lücke in einer Klasse** (K-1, E-1, P-1 mit Prüfgrenze und 34 von 34 Mutationen an nicht genannten Elementen rot), aber ein erster Fall, in dem die Klasse der Hilfe fehlte (Nr. 2). Als Gegenbeleg („der Fix hielt“) aufnehmen.
2. **Neu: `hilfe-ersetzt-argument-verdeckt-eigenschaften`** (aus W-1, 1×): eine Testhilfe, die ein Argument des Falls umhüllt (`fortschrittWaechter` um `src`), verdeckt, was das Original trägt (`Close`, `WriterTo`, …); Gegenmittel (Hülle trägt die Schnittstellen, zählt je Aufruf im eigenen Zustand, `pruefeHuelle` prüft die Hülle selbst, Zähler je Zeile). Beleg: SH6 (ohne `pruefeHuelle` und ohne Durchreichen grün).
3. **Neu: `gegenprobe-haengt-an-der-laufzeit-des-produkts`** (aus G-1, 1×): ein Fall des Werkzeugs fuhr das Produkt mit (`src/bootstrap/...`) und wuchs mit ihm; Ursache war der Zuschnitt, nicht die Last von außen — die Verifikation 2 hatte „Last“ vermutet, der Architect hat die Ursache berichtigt. Der Merkpunkt `rate-ohne-detektor-schwankt-mit-der-last` (2×) bekommt aus dieser Runde keinen Beleg (die Ursache war eine andere).
4. `akzeptiertes-negativ-ohne-folge` (verkörpert): V3-1 und V3-2 (die Grenze des Negativs liegt an der falschen Stelle bzw. nennt eine andere Schwelle als die Menge); Klasse: **Negativ, dessen Abgrenzung die Lücke nicht trifft**.
5. `suite-ohne-generator-fuer-klassen-der-spezifikation` (2×, offen): diese Runde schließt die Lücken der Klassen mit Generatoren in der Suite (O, M2, P6, K1); der Differenztest `voll` gegen `rein` (30 Methoden) ist der Generator für „allein `Read`“ über alle Methoden, den eine Suite mit Hülle nicht bietet; Probe-Paket wieder ohne Abweichung am Produktcode.
6. `universalaussage-im-vertrag-ungeprueft`: „allein `Read`“ (Nr. 2) ist unbeschränkt; die Suite deckt die geschlossene Liste, der Differenztest die Klasse.
7. `groessenmarke-ueberschritten-ohne-rueckfuehrung`: vierter Messpunkt dieses Slice (5612, 6500, 6938, 7236 bei Marke 4500).
8. `befund-an-rolle-ohne-artefakt` (3×): D-1 ist diesmal im Artefakt gelandet (Plan §3 *Abweichungen*, §6 mit Adresse); als Gegenbeleg aufnehmen.

---

## Verdikt

**Bestanden.** Die Befunde der zweiten Verifikation sind **gegen die Klasse gelöst**: W-1 (Menge O mit 8 Methoden, 8 von 8 Mutationen in A5 allein und in 48 von 50 Zeilen rot, dazu `io.Copy` und `Close` an vier Stellen; die Hülle hat einen Fänger, 5 von 5 Mutationen an ihr rot), G-1 (drei erzwungene Läufe von `make gegenproben`: 37 Fälle, 0 rot, Schritt 57,2 bis 58,5 s statt 117,8 bis 120,6 s; der Fänger der Wegwerfmodule 4 von 4 rot), K-1 (2 223 872 Läufe, 8 von 8 Mutationen an nicht genannten Zeichen rot), E-1 (9552 Läufe, 17 von 17 Mutationen an nicht genannten Fehlerarten rot, Prüfgrenze benannt), P-1 (72 Läufe, 9 von 9 Mutationen rot, Prüfgrenze benannt), D-1 und N-1 (Plan trägt Zeile, Negativ, Adresse). Von 84 Mutationen der Verifier 1 und 2 sind 79 rot, 2 grün als benanntes Negativ, 3 äquivalent; die sieben, die in der zweiten Runde nur das Probe-Paket fing, sind in der Suite rot. Das Probe-Paket (1 622 083 Läufe, davon 293 454 neue des Differenztests) findet keine Abweichung am Produktcode. Die Gates sind frisch grün (`make gates` 0; Stufen `test` 88 s, `race` 124 s, `lint` 31 s), Marke und Plan stimmen mit dem Code überein (7236 Zeilen, Produktcode 1640). **Offen ist kein MEDIUM**; V3-1 ist ein LOW (Wortlaut des Negativs), V3-2 bis V3-5 sind INFO und gehen als Befund oder akzeptiertes Negativ in die Closure.

**Übergabe.** An den **Planner**: die Closure (DoD 7 bis 10, §6 Ausgänge, Register nach 6; die Überschreitung der Marke 4500 um 2736 Zeilen mit Grund; den Stempel von `f47a9e3` als Beleg zitieren, er ist mit drei erzwungenen Läufen belegt) und V3-2. An den **Architect**: V3-1 (Negativ um die Schreibseite von `io` ergänzen), V3-3 (Messung in `SPEC-102` nachziehen oder streichen), V3-4 (Abweichungen 1 und 4 in einer Zeile bestätigen). Der **Implementer** hat nichts zu tun.

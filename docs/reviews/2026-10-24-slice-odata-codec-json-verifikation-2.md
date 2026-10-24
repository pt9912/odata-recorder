# Verifikations-Report 2: slice-odata-codec-json — 2026-10-24

**Rolle:** Verifier, zweite Verifikation, frischer Kontext (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-126`, `AGENTS.md` §3.9 bis §3.17). Kein Review, keine Reparatur.

**Gegenstand:** `slice-odata-codec-json` (Welle `welle-record-replay`), HEAD `09760a2`; Basis der Marke `bc05c62`. Geprüft: erste Verifikation `2b50f3d` (V-1 bis V-9), Entscheidungen des Architect `dd25a9c`, Fix-Runde `9f11493` (Suite und Plan), Stempel `09760a2`. Der Produktcode ist seit `0acc49a` unverändert (`git diff 0acc49a..HEAD -- src/adapters src/hexagon/model` leer); die zweite Runde ändert nur Suite (`outboundtest/`), Plan und Stempel.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-24

**Methode.** Belege statt Behauptungen. Alle Mutationen in Wegwerfkopien außerhalb des Repos (Quelle `:ro`, `--network none`, `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container, Gates nie parallel zu Mutationen). Eine neue Basis aus `git archive HEAD` (Scratchpad `…/scratchpad/vf3/`, flüchtig; `runner/v1.txt` bis `v9.txt` die Mutationen, `probe/v3probe_test.go` das neue Probe-Paket, `out/` die Ausgaben). Die Mutationen des Verifiers 1 (`v1` bis `v3`) laufen unverändert gegen den neuen Stand; dazu rund 30 eigene Mutationsläufe an Elementen, die weder Verifier 1 noch der Implementer nannten (`v4` bis `v9`), und ein neues Probe-Paket aus Generatoren. Das Repo ist bis auf diesen Report unverändert (`.claude/agents/architect.md` nicht angefasst). Fremde Container (`pgr-it-*`) liefen zweimal zeitweise auf dem Host und wurden nicht berührt; ich habe auf ihr Ende gewartet.

---

## 1. Ergebnis in einer Tabelle: V-1 bis V-9 gegen die Klasse

| Befund aus Verifikation 1 | Klasse (Wortlaut `SPEC-126`) und im Lauf gezählte Menge | Eigene Mutationen an Elementen, die weder Verifier 1 noch Implementer nannten | Urteil |
|---|---|---|---|
| V-1 (MEDIUM) Ende des Stroms in einer Zahl | Nr. 4 „endet vor dem Ende des Wertes, in jedem Zustand, auch als ganzer Body“: Menge E. Ich habe sie **unabhängig nachgezählt** (Probe `TestV3ZaehleE`, nach dem Wortlaut des Plans gebaut, Orakel `json.Valid`): 53 + 10 + 16 + 17 + 8 = **104** Präfixe, 88 auch mit Leerzeichen, **192** Körper je Methode; die Suite zählt dieselbe Zahl und hält bei einer anderen an | NZ1 (Ende nimmt `tru`/`fals`/`nul` als vollständig), NZ2 (`zExpVz` nur nach `E-`), NZ3 (`zExpVz` nur bei negativer Zahl und `+`), NZ4 (`zExpE` nur bei negativer Zahl und `E`), NZ5 (`zPunkt` nur bei Länge 4, `-12.`): **5 von 5 rot** aus dem benannten Grund (`struktur` statt `syntax` an einem Körper aus E) | **gelöst** |
| V-2 (MEDIUM) Lesen nach einem Fehler | Nr. 2 „jeder andere Fehler … gleich zu welchem Zeitpunkt“: Menge M2, im Lauf gezählt **3980** (5 × 279 + 5 × 517) | NE1 (noch einmal lesen nur bei Fehler allein vor dem ersten Byte) rot; NE4 (nur bei Fehler mit Bytes im selben Read, ungewrapped) rot, beide mit „N Reads nach dem Fehler, erwartet 0“; VR1 und VM10 des Verifiers 1 rot, VM10 **im Fall** (`errQuelle` erwartet), nicht an der Zeitgrenze. **Aber:** NE2, NE3 und NEall2 (einmal noch lesen nur bei `io.ErrNoProgress`, `io.ErrClosedPipe`, `io.ErrShortBuffer`) bleiben **grün** | **gelöst für die Art der Mutation, Rest siehe E-1** |
| V-3 (MEDIUM) Puffergrenze | Nr. 5 „jeder Wert von `grenze` und jede Zerlegung“: Menge P5, im Lauf gezählt **1080** (10 × 3 × 3 × 3 × 2 × 2) | NP2 (`gelesen` um `want` statt um die gelieferten Bytes erhöht) rot (Panik/`zu_gross` bei gültigem Body in B5); VM11 und VM12 rot. **Aber:** NP1 und NPallP1 (Grenze wird nach 512 KiB gelesenen Bytes nicht mehr durchgesetzt) bleiben **grün**, P5 reicht bis `grenze` 393 217 | **gelöst für den Puffer, Rest siehe P-1** |
| V-4 (LOW) K1 ab U+0080 | Nr. 14 „kein anderes Zeichen als `~` und `/` wird ersetzt“: im Lauf gezählt **1 111 936** (`K1 Zeichen ab U+0080`) + 384 = **1 112 320** | VM8, VM9 (U+3042, U+1F600) rot; NK1c (U+10FFFE wird `~2`, **nur im Pfad einer Collection**) **grün** in der ganzen Suite | **gelöst für `ParseEntity`, Rest siehe K-1** |
| V-5 (LOW) A2 mit dem Byte dahinter | Nr. 3 „gleich welche Bytes danach folgen“: im Lauf gezählt **262 147** (4 + 1020 + 3 + 261 120) | NA1 (`FF FF` bleibt im BOM-Zustand), NA2 (`FE FE`), NA3 (`EF BB BB`): **3 von 3 rot** (`B3 A2 hinten … erwartet syntax`); VM18, VM19 rot | **gelöst** |
| V-6 a bis d (LOW) | Plan und Bericht | (a) Zeile F-7 berichtigt (liest „an den Reads mit leerem Puffer“); (b) Eigenschaft statt Zusage, akzeptiertes Negativ im Plan; (c) `HeapAlloc`, Methodenwechsel und Zeitgrenze unter Last als Entscheidung und Negativ im Absatz *Speicherwächter*; (d) `drossel` in Dateitabelle und Absatz | **gelöst** (gelesen; eine neue Lücke in (c), siehe G-1) |
| V-7 (INFO) | Negativ „Fortsetzungsbyte zwischen den Grenzen“ | VM5 (Reject von `A5` als inneres Fortsetzungsbyte) ist jetzt **rot** (durch `K1 oben`, `U+0940`), vorher grün; NU1 und NU2 (Annahme von `C5` an einer inneren Stelle) bleiben grün, NU3 (`7F`) rot | Entscheidung gilt; Text des Negativs ungenau, siehe N-1 |
| V-8 (INFO) | Standardbibliothek ohne Gate | unverändert | kein Befund |
| V-9 (INFO) | Stempel in §6 getrennt gemeldet | steht in §6 *Größe* | **gelöst** |

Die 21 Mutationen des Verifiers 1 (`v1`), die 5 des Gesamtlaufs (`v2`) und die 22 Hex-Ziffern (`v3`): **48 von 48 rot**, jede aus dem benannten Grund (`v1` einschließlich VM5, VM6, VM7 bis VM9, VM10 und VM12 im Fall statt an der Zeitgrenze; VA5a und VA5b an den Zählern von A5).

## 2. DoD und Plan

| DoD-Punkt (Plan §2) | Befund |
|---|---|
| 1 bis 3 Port, Schlüssel, Nachweise | erfüllt: Produktcode unverändert, in rund 1,33 Mio. neuen Läufen (siehe 3) und den 22 Mio. des Verifiers 1 ohne Abweichung |
| 4 `make gates` grün | erfüllt, siehe 4 |
| 5 Review, Report liegt vor | erfüllt (`64b7cc9`); die Verifikationen sind `2b50f3d` und dieser Report |
| 6 Doku-Update | erfüllt (`SPEC-126`, Architect `dd25a9c`; Zahlen 192, 3980, 1080, 1 112 320, 262 147 stimmen in Plan, Raster und Tabelle; `1027` und `390` stehen nur noch als Vorher-Werte in der Entscheidungstabelle) |
| 7 bis 10 Closure, Register, Risiken, Paarungen | noch offen (§7 „Wird bei der Closure gefüllt“, §6 alle Risiken *offen bis zur Closure*); Sache des Planners nach diesem Report |

**Größe (Punkt 3 des Auftrags).** `git diff --numstat bc05c62..HEAD` über die Pfade der Marke: **6938** Zeilen (6936 Einfügungen, 2 Löschungen), davon Produktcode **1640** (ohne `_test.go`, `outboundtest/`, `testdata/`), Tests und Suite **5298**; gleich dem Bericht. Der Zuwachs der zweiten Runde ist 438 (6500 → 6938); die Summe aus Einfügungen und Löschungen gegen `56ad5ec` über die Suite-Dateien ist 476 (88 + 65 + 194 + 109 + 20), wie im Plan. Marke 4500: Überschreitung 2438 Zeilen (54,2 %), in §6 *Größe* ehrlich mit Zahl, Aufteilung und Grund ausgewiesen („Marke nicht geändert“); `harness/gegenproben.stempel` (1 Zeile, 2 im Diff) ist getrennt gemeldet. Keine Änderung der Marke (Sache des Planners).

## 3. Probe-Paket aus Generatoren (Orakel unabhängig vom Leser)

Der Leser nutzt `encoding/json` nicht; `json.Valid`, `utf8` und `math/big` stehen nur im Probe-Paket. Alle Läufe gegen HEAD **ohne Abweichung am Code**:

| Test | Klasse (neu, so vorher nicht gefahren) | Läufe |
|---|---|---|
| `TestV3Stromende` | **Stromende-Präfixe jedes gültigen Dokuments:** 907 Dokumente (224 Zahlen-Schreibweisen als Wurzel, im Array, im Objekt, verschachtelt; Strings mit jedem Escape, Surrogatpaar, `é`, `😀`; Literale; Objekt/Array; eine Entity und eine Collection mit allen acht Typen) × jedes echte Präfix × {ohne, Leerzeichen, Tabulator, LF, CR, Leerzeichen+LF} × beide Methoden; Erwartung `syntax` genau, wenn `json.Valid` das Präfix ablehnt (UTF-8 am Ende angefangen: ohne Leerraum), sonst nicht `syntax` | 189 212 (179 324 `syntax` erwartet, 9 888 gültige JSON-Werte) |
| `TestV3Grenzlagen` | **Grenzlagen B5 mit anderen Puffern als 2^k:** `grenze` an 24 Vielfachen von 32 768 (1 bis 257, bis 8,4 MiB) und 6 anderen Basen (3000, 12 345, 40 000, 70 000, 98 304, 100 003), je ±3; Länge g−1, g, g+1; Stücke {ganz, 7, 4096, 32 767, 32 768, 32 769, 65 536}; Varianten {glatt, jeder zweite Read `0, nil`, `io.EOF` im letzten Read}; Zähler der gelieferten Bytes ≤ g+1 | 22 896 |
| `TestV3Endlos` | Strom aus Leerraum ohne Ende (Deckel g+100 000), `grenze` von 1000 bis 64 MiB: genau g+1 Bytes geliefert, `zu_gross` | 92 (46 Grenzen × 2 Methoden) |
| `TestV3FehlerExotisch` | **Fehler von `src` mit exotischen Typen an jedem Read-Index:** 16 Fehler (`errors.New`, `io.ErrNoProgress`, `io.ErrClosedPipe`, `io.ErrShortBuffer`, `io.ErrShortWrite`, `os.ErrDeadlineExceeded`, `os.ErrClosed`, `ECONNRESET`, `EAGAIN`, `EINTR`, `*net.OpError`, `*os.PathError{Err: io.EOF}`, `errors.Join(io.EOF, …)`, `*url.Error{Err: io.ErrUnexpectedEOF}`, ein Typ mit `Timeout()`/`Temporary()`, doppelt umhüllter `io.EOF`) × `E1`- und `C1`-artige Körper × jeder Index k = 0 bis L × Fehler allein oder im Read des Bytes k; Erwartung `errors.Is`, kein `DatenFehler`, 0 Reads nach dem Fehler (der Fake wiederholt den Fehler) | 4 480 |
| `TestV3PointerCollection` | jedes Zeichen U+0080 bis U+10FFFF (ohne Surrogate) als Name in einem Element von `value` (Ort `/value/1/…`) | 1 111 936 |
| `TestV3Close` | Strom mit `Close` ohne Hülle, 6 Bodys × 2 Methoden: 0 `Close` | 12 |
| `TestV3ZaehleE` | Nachzählung der Menge E (siehe 1) | 1 |

Zusammen **1 328 629** Läufe (zusätzlich Verifier 1: rund 22 Mio.). **Gegen die Mutationen:** das Probe-Paket ist **rot gegen PE2, PE3, PP1b, PK1c, PC1, PZ1** (`io.ErrNoProgress`/`io.ErrClosedPipe` einmal wiederholt; Grenze nach 512 KiB nicht durchgesetzt; Collection-Pfad mit U+10FFFE; `Close` über Typzusicherung; `zExpVz`), also an genau den Stellen, an denen die Suite grün bleibt.

## 4. Gates, Laufzeiten, Last

Alle Läufe am Stand `09760a2`, Repo unverändert (vor dem Schreiben dieses Reports).

| Sensor | Ergebnis | Laufzeit |
|---|---|---|
| `make gates` | Rückgabecode 0 (Docker-Cache der Stufen, darum kurz) | 12 s |
| `make test` frisch (`docker build --no-cache-filter test --target test`) | grün | 76 s |
| `make test-race` frisch (`--no-cache-filter race`, Image `golang:1.27.1-trixie`) | grün | 118 s |
| `make lint` frisch (`--no-cache-filter lint`) | grün | 33 s |
| Vertragssuite allein (`go test ./src/revtest -v`, `--cpus=4`) | grün | 41 s (A4 24 s; K1 4,8 s) |
| **Gegenproben** (`docker build --no-cache-filter gegenproben`, die Zeile des Targets; `make gegenproben` meldet „Eingaben unverändert“) | **Lauf 1: 37 Fälle, 2 rot** (`test/basis`, `test/roter-test`: „Zeit überschritten“, 120,6 s); **Lauf 2: 37 Fälle, 2 rot** (dieselben); **Lauf 3: 37 Fälle, 0 rot** (117,8 s von 120 s) | 130 s, 130 s, 151 s |

Last (`/proc/loadavg`, 20 Kerne): vor den Läufen 7,0 (Lauf 1, danach 66), 7,0 (Lauf 2), 5,3 (Lauf 3, danach 38). Beim Gate-Lauf `test`/`race`/`lint` 3,7 bis 7,6. Zweimal lief ein fremder Container (`pgr-it-*`, `pgwire-recorder:integration`) kurz zuvor; die Läufe 2 und 3 begannen erst nach seinem Ende und bei einer Last unter 8.

**Raten (Stichprobe, Messender: Verifier; jeder Prozess `-test.count=1`; Mutationen aus `runner/rate.txt` des Verifiers 1):**

| Zeile | Mutation | Prozesse rot |
|---|---|---|
| J1 | Zähler des Pakets in `lies`, Detektor (Image `golang:1.27.1-trixie`) | 10 von 10 |
| J2 | Puffer des Pakets unter Sperre, ohne Detektor | 10 von 10; 3 Prozesse mit `-test.count=20`: 60 von 60 Läufen |
| I1 | `Properties` über Map, ohne Detektor | 10 von 10; 3 Prozesse mit `-test.count=20`: 60 von 60 Läufen |

Die Zahlen des Implementers (30 von 30, 20 von 20, 100 von 100) sind nicht widerlegt.

## 5. Befunde

| ID | Schwere | Befund | Zuständig |
|---|---|---|---|
| W-1 | **MEDIUM** | **Die Zusage „der Leser ruft `Close` nie“ (`SPEC-126` Nr. 2, Zeile A5) hat seit `9f11493` keinen Fänger mehr.** Der neue Wächter `fortschrittWaechter` umhüllt in `jsuite.aufruf` jedes `src` mit einem Typ, der nur `Read` hat. Eine Mutation, die `src` per Typzusicherung schließt (`if c, ok := src.(io.Closer); ok { defer c.Close() }` in `lies`) sieht den `Close` des Fakes (`a5Strom`, `deckelStrom`, `teilReader`) nie; die Zähler `rd.closes` und `e.closes` in A5 bleiben 0. Beleg: dieselbe Mutation ist am Stand `56ad5ec` (Suite davor) **rot** in A5 („Close 1“), am Stand `09760a2` **grün in der ganzen Suite** (NC1, NC1all) und im Probe-Paket (`TestV3Close`) rot. Die Zeile A5 der Tabelle und die Zeile A5 des Berichts nennen „`Close` gerufen“ als rote Mutation; das stimmt nicht mehr (`AGENTS.md` §3.9, §3.15). Eine Testhilfe hat eine Mutation verdeckt. Fix: die Hülle reicht `Close` an `inner` weiter, wenn es ein `io.Closer` ist (oder A5 reicht den Wächter selbst durch), und der Fänger zählt es. | Implementer |
| G-1 | **MEDIUM** | **Die Zusage „37 Fälle ohne rot“ (Stempel `09760a2`, Plan-Bericht, „2 min 20“) ist nicht verlässlich reproduzierbar.** In drei Läufen auf einem ruhigen Host (Last 5 bis 7 vor dem Lauf) endeten zwei rot (`test/basis` und `test/roter-test`, 120,6 s: Zeitgrenze von 120 s je Fall), der dritte grün mit 117,8 s, also 2 Sekunden Abstand zur Grenze. Die Last wächst im Lauf auf 38 bis 66: `tools/gegenproben/main.go` fährt die Fälle mit `runtime.NumCPU()` zugleich, jeder dieser Fälle fährt die ganze Suite; die Last, die das negative Akzeptieren im Plan (*Speicherwächter*: „jenseits etwa des Dreifachen der Kernzahl … wer es sieht, wiederholt bei Ruhe“) als Ursache nennt, erzeugt der Lauf **selbst**. Mit dem Wachstum der Suite (41 s statt 34 s allein, Stufe test 76 s statt 71 s) ist der Abstand zur Frist aufgebraucht. Der Stempel gilt nur nach einem grünen Lauf; er bezeugt hier ein knappes Glück. Gates sind nicht betroffen (die Gegenproben liegen außerhalb der Gate-Kette, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)), aber die Closure darf den Stempel nicht als Beleg zitieren. Kein Befund am Produktcode. | Planner (Beleg und Negativ berichtigen), Architect (Laufzeit-Budget der Suite gegen die Frist, keine Lockerung ohne ADR, `AGENTS.md` §3.6) |
| K-1 | LOW | **K1 ab U+0080 fährt nur `ParseEntity` (Bericht Abweichung 2), und das steht nur im Bericht.** NK1c (U+10FFFE wird `~2`, nur wenn der Pfad `/value/<i>` vorangeht) bleibt in der **ganzen Suite grün**; die Collection kennt unter den Zeichen ab U+0080 nur 6 Vertreter (U+0080, U+07FF, U+0800, U+FFFF, U+10000, U+10FFFF) und `é` (E5). Die Folge ist im Bericht genannt, aber weder Zeile K1 noch *Akzeptierte Negative* tragen sie, und der Bericht sagt „an Architect zurückgeben“; die Entscheidung fehlt. `TestV3PointerCollection` (1 111 936 Läufe) fährt die Collection ohne Abweichung und fängt NK1c. Die Menge kostet 4,8 s; eine zweite Methode etwa 5 s mehr. | Architect (Menge oder Negativ mit Grund und Folge), Implementer |
| E-1 | LOW | **V-2 deckt die Art der Mutation, nicht die Menge „jeder andere Fehler“.** `SPEC-126` Nr. 2 sagt „jeder andere Fehler“ (unbeschränkt); M2 führt 5 Arten. Mutationen, die einen einzelnen Fehlerwert besonders behandeln (einmal noch lesen bei `io.ErrNoProgress`, `io.ErrClosedPipe`, `io.ErrShortBuffer`) bleiben grün; `AGENTS.md` §3.17: bei unbeschränkter Menge nennt der Bericht die Prüfgrenze mit Grund und Folge; das fehlt. `TestV3FehlerExotisch` (16 Arten × jeder Index) fängt alle drei. Der Code behandelt jeden Fehler gleich (`case err != nil`), die Folge im Betrieb ist klein. | Architect (Negativ oder Menge), Implementer |
| P-1 | LOW | **V-3 deckt den Puffer, aber P5 reicht nur bis `grenze` 393 217.** Eine Mutation, die die Grenze nach 512 KiB gelesenen Bytes nicht mehr durchsetzt (NP1, NPallP1: `rest < want && gelesen < 1<<19`), bleibt grün; die Standardgrenze ist 8 MiB. `TestV3Endlos` (Strom ohne Ende, `grenze` bis 64 MiB, genau g+1 Bytes) ist rot. Die Prüfgrenze von P5 steht nicht im Plan (`SPEC-126` Nr. 5: „für jeden Wert von `grenze`“). | Architect (Negativ mit Grund und Folge), Implementer |
| D-1 | LOW | **Die Abweichungen (1) bis (3) der zweiten Runde stehen nur im Bericht** (dieselbe Klasse wie V-6 c, `befund-an-rolle-ohne-artefakt`): (1) `fehlerStrom` liefert nach dem Fehler Bytes und `io.EOF`, die Zeile A3 sagt „weitere Bytes oder Fehler“ (Zeile nicht nachgezogen, `AGENTS.md` §3.9); (3) die Probe unter dem Detektor (A2: 1027 statt 262 147, K1: Probe) hat weder Grund noch Folge unter *Akzeptierte Negative* (die Zeilen B3 und K1 nennen die Probe, nicht den Grund; Vorbild A4: „eine Zweiteilung hat keinen Zustand, den der Detektor sähe“); keiner der drei hat in §6 eine Zeile *offen* mit Adresse. | Planner (§6), Architect (Entscheidung) |
| N-1 | INFO | Das Negativ „Fortsetzungsbyte zwischen den Grenzen“ (V-7) ist seit `K1 oben` in einer Richtung geschlossen (Ablehnung eines gültigen Zeichens, VM5 jetzt rot), in der anderen offen (Annahme eines ungültigen inneren Bytes, NU1 `C5` als zweites Byte einer Drei-Byte-Folge, NU2 als zweites Byte einer Vier-Byte-Folge: grün; `7F` rot, weil Grenzbyte). Der Plantext sagt es nur als Ganzes. | Architect |
| N-2 | INFO | `endlichStrom` und `fehlerStrom` sind korrekt und verdecken nach meinen Mutationen nichts: NE1, NE4, VR1, VM10 werden im Fall rot; `fehlerStrom` zählt jeden Read nach dem Fehler und beendet den Strom nach 10 000 Reads. `fortschrittWaechter` ist als Zählung korrekt (VM12, VA5b im Fall rot), verdeckt aber `Close` (W-1) und jede andere optionale Schnittstelle von `src` (`io.WriterTo`, `io.ByteReader` …): ein Leser, der `src` über eine solche Schnittstelle liest, liest die Hülle, nicht den Fake. | Implementer |

Gültige Seite: **keine Verletzung der Spezifikation, kein HIGH.** Der Produktcode ist in allen gefahrenen Klassen richtig.

## 6. Plan gegen Code (`AGENTS.md` §3.9) und Tabelle gegen Vertrag

- Tabellenzeilen A3 (M2 3980), B1 (E 192), B3 (A2 262 147), B5 (P5 1080), K1 (1 112 320) und die Entscheidungstabelle V-1 bis V-5 stimmen mit `SPEC-126` Nr. 2 bis 5, 14 und mit dem Lauf überein; die Dateitabelle nennt `fortschrittWaechter`, `endlichStrom`, `drossel`; die Raster-Zellen („abgeschnitten … `1e+`“, „Lesefehler … jeder Fehler“, `grenze` P5, BOM 262 147) tragen die neuen Zahlen. Zeile A5 ist falsch geworden (W-1), Zeile A3 sagt über den Fake mehr, als er tut (D-1).
- Der Bericht der zweiten Runde nennt je Befund Menge, Zahl und eine Mutation an einem nicht genannten Element; die Zahlen stimmen mit meinen überein, die Mutationen laufen rot (NA/NZ/NE1/NE4 bestätigen die Klasse mit anderen Elementen). Nicht genannt: die Menge der Fehlerarten ist unbeschränkt (E-1), P5 reicht nicht über 393 217 (P-1), K1 nicht über die Collection (K-1).
- Die Rate-Zeile des Berichts („keine; Schleifen ohne Zufall“) stimmt; die Goroutinen von `gFahre` ändern nur die Reihenfolge der Meldungen.

## 7. Register (nichts geschrieben): Evidence-Kandidaten für den Lese-Schritt der Closure

Unter `docs/plan/planning/observations/BEO-REPO/` gibt es zu diesem Slice keine Evidence-Datei. Vorzumerken (Planner entscheidet):

1. `fix-deckt-die-genannte-mutation-nicht-die-klasse` (verkörpert in `AGENTS.md` §3.17): dritte Runde in diesem Slice; wieder Lücken in unbeschränkten Mengen ohne Prüfgrenze (E-1, P-1, K-1); als weitere Evidence-Datei.
2. `suite-ohne-generator-fuer-klassen-der-spezifikation` (1×, offen): dieser Slice (Review: 14 Überlebende; Verifikation 1: V-1, V-3 bis V-5); damit 2×. Die Generatoren der Suite haben die Klassen von V-1 bis V-5 geschlossen; das Probe-Paket fand in dieser Runde noch keine Abweichung am Code, aber die Lücken P-1, E-1, K-1 nur mit Generatoren.
3. `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen): G-1 (Gegenprobe an der Zeitgrenze bei selbst erzeugter Last, 2 von 3 rot) und Verifikation 1 (Last von außen); damit 3×, löst beim Lese-Schritt die Verkörperung aus.
4. `befund-an-rolle-ohne-artefakt` (2×, offen): D-1 (Abweichungen nur im Bericht), davor V-6 c; damit 3×.
5. `groessenmarke-ueberschritten-ohne-rueckfuehrung`: dritter Messpunkt dieses Slice (5612, 6500, 6938 bei Marke 4500).
6. `vertrag-sagt-mehr-zu-als-ein-gate-prueft` / neuer Eintrag: W-1 (eine Testhilfe nimmt der Zusage „ruft `Close` nie“ den Fänger; die Hülle um ein Argument verdeckt, was das Argument sonst trägt; Klasse: **Hilfe, die ein Argument des Falls ersetzt, verdeckt die Eigenschaften des Originals**).
7. `akzeptiertes-negativ-ohne-folge` (verkörpert): D-1 (Probe unter dem Detektor ohne Grund und Folge) und N-1.
8. `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×, offen): in dieser Runde kein neuer Beleg.

---

## Verdikt

**Nicht bestanden (eng, kein Neubau).** V-1 bis V-5 der ersten Verifikation sind **gegen die Klasse gelöst**: die Mengen E (192), M2 (3980), P5 (1080), K1 (1 112 320) und A2 (262 147) habe ich im Lauf selbst nachgezählt, von 20 verwertbaren eigenen Mutationen an nicht genannten Elementen sind 12 rot aus dem benannten Grund (die 8 grünen sind K-1, E-1, P-1, N-1 und W-1), die 48 Mutationen des Verifiers 1 sind alle rot, und der Produktcode besteht rund 1,33 Mio. neue Läufe aus fünf neuen Klassen ohne Abweichung. Die Gates sind frisch grün (test 76 s, race 118 s, lint 33 s, `make gates` 0). Nicht bestanden, weil die zweite Runde **eine Zusage ihren Fänger gekostet hat** (**W-1 MEDIUM**: `Close` verdeckt durch `fortschrittWaechter`, `AGENTS.md` §3.15) und weil der Stempel der Gegenproben (**G-1 MEDIUM**) in zwei von drei Läufen nicht reproduzierbar ist; dazu **vier LOW** (K-1, E-1, P-1, D-1: unbeschränkte Klassen ohne Prüfgrenze, Abweichungen nur im Bericht) und zwei INFO.

**Übergabe.** An den **Implementer**: W-1 (die Hülle muss `Close` durchreichen, der Fänger A5 zählt es wieder; Mutation `src.(io.Closer)` rot sehen), K-1/E-1/P-1 nach der Entscheidung des Architect. An den **Architect**: K-1, E-1, P-1, N-1 (Menge oder Negativ mit Grund und Folge, `AGENTS.md` §3.15, §3.17) und G-1 (Verhältnis der Laufzeit der Suite zur Frist von 120 s; eine Lockerung der Frist ist ein ADR). An den **Planner**: G-1 (Stempel nicht als Beleg zitieren, Negativ im Absatz *Speicherwächter* berichtigen: die Last ist selbst erzeugt), D-1 (§6-Zeilen), die Register-Merkliste (7) und die Closure.

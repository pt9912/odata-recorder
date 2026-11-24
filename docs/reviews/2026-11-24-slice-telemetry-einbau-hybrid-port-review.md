# Review-Report: slice-telemetry-einbau-hybrid-port (Teil 1) — 2026-11-24

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-telemetry-einbau-hybrid-port`, HEAD `1ff88517` (Handoff-Commit, Code und Plan zusammen), Messbasis `ccd4d9fd`; Teil 1 des Schnitts von `slice-telemetry-einbau-hybrid`, Zeilen P1 bis P9: Port `support.FallbackTelemetry`, `SpanHybridFallback` und sechste Zeile in `support.Spans()`, `VerworfeneCodes()`, `Snapshot.Verworfen` samt `Kopie()`, `FakeTelemetry.CountDiscarded`, im Adapter `CountDiscarded`, die Reihe der verworfenen Versuche und zwei Schlüssel in `regeln.go` (`src/hexagon/ports/support/`, `supporttest/`, `src/adapters/driven/telemetry/`). Teil 2 (Dienst und Verdrahtung) ist nicht Gegenstand.

**Skill:** `.harness/skills/reviewer.md` @ `c3741274` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-24

**Eingangs-Kontext:**

- Plan `slice-telemetry-einbau-hybrid-port`: §1 (Ziel, Sendung, Ausschlüsse), §2 (DoD, nur zur Abgrenzung), §3 (Tabelle P1 bis P9, Mengen mit Zählprüfung, Mutationen und Raten des Implementers, Nullblöcke, akzeptierte Negative (1) bis (12)), §4 (Block, Marke 1100, Schätzung 790, Basis `ccd4d9fd`), §6 (Raster, offene Rückgabe zur Zeile P8, Adresse Planner), §8.
- `SPEC-320`, `SPEC-321` (Nr. 1 bis 9, *Ungeprüft*), `SPEC-230` Nr. 1, 3, 5, 8 bis 11, `SPEC-232` Nr. 5 a, b und g und *Ungeprüft*, `SPEC-231` Nr. 6, `SPEC-284` (Port der Konfig), `SPEC-294` Nr. 5 (Variante A), `SPEC-296`; die Port-Tabelle und `ARC-021` in `spec/architecture.md`.
- `AGENTS.md` §3.6, §3.7, §3.9 bis §3.18; `harness/README.md` §Sensors.
- `LH-FA-12`, `LH-QA-04`, `LH-QA-05`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Bezug des Plans, nur genannt).
- Vorherige Reviews am Modul: `slice-telemetry-einbau-record`, `-replay`, `-gateway`, `-mock` (Klassen *Fänger führt nur eine Form der Eingabe*, *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*, *Zweig ohne Lauf unter dem Detektor*) und `slice-telemetry-port` (Absturz `concurrent map` als Fänger der Kopie ohne Sperre).

**Unverändert (geprüft mit `git diff --name-only ccd4d9fd..HEAD`):** außerhalb der vier Pfade des Blocks ändert der Zweig nur Plan-Dokumente (`slice-telemetry-einbau-hybrid-port.md`, `roadmap.md`, `welle-betrieb.md`, die zwei letzten im Start-Commit des Planners); `Makefile`, `harness/`, `Dockerfile`, `.golangci.yml`, `.a-check.yml`, `spec/`, `tools/`, `src/bootstrap/`, `src/hexagon/services/`, `src/adapters/driving/` und `tests/` sind unverändert. Kein `//nolint`, kein `t.Skip` im neuen Code (`git diff ccd4d9fd..HEAD -- src` durchsucht), kein Backslash-`u`, kein BOM und kein NBSP in den Zeilen des Diffs; das Byte `0x80` in `verworfen_test.go` entsteht aus `string([]byte{0x80})` (gegengelesen mit `LC_ALL=C grep -P` und `-F`).

**Größenmarke:** `git diff --numstat ccd4d9fd..HEAD` über die vier Pfade des Blocks: 839 Einfügungen, 175 Löschungen, Summe 1014 gegen die Marke 1100, wie berichtet; Produktcode 96 (`support/telemetry.go` 31, `supporttest/telemetrie.go` 26, `telemetry/telemetry.go` 34, `regeln.go` 5), Tests 918; fremde Zeilen 0 (`git diff --numstat ccd4d9fd..d7bd305b` über dieselben Pfade ist leer). `make groessenmarke`: `Messung 1014 von Marke 1100`, Exit 0.

**Sensoren dieses Laufs:** Der Reviewer fuhr nicht `make gates`, sondern einzeln: `make test` mit `--no-cache-filter test` (Stufe frisch, `pruefung test: ok`, 400 s im Container, 421 s gesamt bei einer Last des Hosts von 15 bis 22), `make test-race` mit `--no-cache-filter race` (Stufe frisch, `pruefung race: ok`, 469 s im Container, 493 s gesamt, Last 10 bis 30; keiner der bekannten Flakes `TestHandbuchZeitgrenzeUndWarten` und `TestAnfrageBodyBytes` trat auf), `make a-check lint` (Exit 0 nach 5 s: alle Stufen aus dem Build-Cache, also der Lauf des Implementers, nicht frisch), `make groessenmarke` (siehe oben). Abdeckung: `-race -coverprofile` je Paket, `support` 100,0 %, `supporttest` 94,1 %, `telemetry` 99,3 %; Blöcke mit Zähler 0: in `supporttest` allein `vertrag.go` (unverändert), in `telemetry` allein `regeln.go:72` und `regeln.go:137` (unverändert, wie im Plan); kein Block mit Zähler 0 im neuen Code.

**Nicht nachgemessen:** `make gates` als Kette, `make test-integration`, `make test-determinism`, `make test-contract`, `make handbuch-check`, `make katalog-check`, `make gegenproben` (nicht gefahren, Auftrag); `make lint` frisch (nur aus dem Cache bestätigt); `make docs-check` und `make zeichen-check` laufen nach dem Schreiben dieses Reports (Ergebnis im Bericht an den Aufrufer, nicht hier); die Raten des Implementers mit 30 Prozessen je Zelle (hier 20 je Zelle); die Mutationen P8 `*` Zählung unter `idMu` im gültigen Zweig sowie P9 `*` je Zweig (nur die Gesamtmutation P9 und die Entsprechungen unten); die Bestandstests gegen die Kopie von `Errors` und `Dropped` ohne Sperre (siehe unten).

---

## Eigene Mutationen

Ein eigener Treiber im Scratchpad (`hybrid-port-reviewer/`, außerhalb des Repos, nach dem Vorbild des Treibers des Implementers): Container `odata-recorder:race` ohne Netz, Mount `:ro`, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, nacheinander, nie neben `make gates`; je Mutation eine Wegwerfkopie von `go.mod`, `go.sum` und `src/` und ein Prozess `go test -count=1` über **alle** Tests der drei Pakete (`support`, `supporttest`, `telemetry`, ohne `-run`, damit auch fremde Tests fangen), `CGO_ENABLED=0`; nie im Worktree. Die Nullprobe (Baum ohne Änderung) ist grün (`ok` in allen drei Paketen).

**37 Mutationen: 34 rot aus dem benannten Grund, 1 rot durch eine Zusicherung zur Übersetzungszeit, 1 rot durch einen Aufbaufehler (nicht gezählt), 1 äquivalent (grün).** Jede liegt an einem Element, das der Plan oder der Handoff nicht nannte oder an einer anderen Stelle als dort (Stelle der Zeile in `Spans()`, Art des Attributs, Zweig, Sammlung, Operation des Fakes):

| Zusage (Zeile) | Mutation | roter Test |
|---|---|---|
| P1 (6) | sechste Zeile an erster Stelle; drittes Pflicht-Attribut `recorder.fallback_x`; `recorder.fallback_type` doppelt statt `recorder.fallback_result`; Konstante mit angehängtem Leerzeichen; `recorder.Fallback_result`; Zeile trägt `SpanMockQuery` statt `SpanHybridFallback` | `TestSpansDaten` (alle sechs; die letzte auch `TestSchluesselmengen`) |
| P2, P3 (4) | `recorder.fallback_type` mit Art `Int`; `recorder.fallback_type` nicht in `spanRegel`; `recorder.fallback_type` mit Form Name; `recorder.fallback_type` in der Menge Log | `TestSchluesselmengen`, `TestCanarySpans`, `TestEndGleichzeitig` (die ersten zwei); `TestHybridSchluesselForm` (Form Name); `TestWithAttrs`, `TestCanaryHandler`, `TestHandlerAbleitenMasse` und drei weitere (Log) |
| P4 (11) | ungültiger Code zählt `schluessel` statt `fehlercode`; ungültiger Code legt einen Schlüssel an; gültiger Code zählt zusätzlich `fehlercode`; ungültiger Code zählt zusätzlich in `Errors`; Prüfung gegen `a.errors` statt `a.verworfen`; nur der letzte Code zählt doppelt; Anlage von `verworfen` für drei Codes; Anlage ohne Einträge; `VerworfeneCodes()` teilt den Slice; fünfter Code `ODR-E6004` | `TestCountDiscardedJeCode` (vier), `TestReihenFest` (fünf), `TestVerworfeneCodes` (zwei); der letzte Code auch `TestReihenGetrennt`, `TestCountDiscardedGleichzeitig`, `TestCountDiscardedBlockierteSenke` |
| P5 (5) | `Snapshot` ohne `Verworfen`; `Snapshot` mit leerer Karte statt `a.verworfen`; `Kopie()` macht eine leere Karte zu nil; `Kopie()` macht nil zu einer leeren Karte; `Kopie()` teilt `Verworfen` nur bei mehr als einem Eintrag | `TestReihenFest` (zwei), `TestSnapshotKopieBeispiel` und `TestSnapshotKopie` (zwei), `TestSnapshotKopie` (eine) |
| P6 (8) | Fake zählt `CountDiscarded` doppelt; Fake verwirft Codes unter 1; Fake setzt ungültige Codes auf 0; Fake zählt unter `OpCountError`; `Aufrufe()` liefert das innere Slice; leerer Stand des Fakes ohne `Verworfen`; gesetzter Stand ohne `Verworfen` wird ergänzt; `SetzeStand` kopiert flach | `TestFakeCountDiscarded`, `TestFakeAufzeichnung`, `TestFakeAnzahl` und weitere (die ersten fünf), `TestFakeSnapshotStand` und `TestFakeStandKopie` (die letzten drei) |
| P7 (2) | `FallbackTelemetry` ohne Einbettung von `Telemetry`; `FallbackTelemetry` mit `TelemetryReader` | `TestTelemetrySignaturen` (die zweite); die erste rot durch die Zusicherung `var _ support.Telemetry = support.FallbackTelemetry(nil)` (`[build failed]`) |
| P6 (1) | `OpCountDiscarded` vor `OpSnapshot` in der Folge der Konstanten | **Aufbaufehler, nicht gezählt**: `panic: index out of range` in `aufzeichnen` aus `TestFakeAufzeichnung`; die Mutation bricht das Feld `zaehler`, keine Zusage |
| P4 (1) | `CountDiscarded` prüft zusätzlich `code > 0` | **grün, äquivalent**: jeder Schlüssel von `verworfen` ist positiv, die Prüfung ändert kein Ergebnis |

## Auflösung der Nebenläufigkeit (Raten)

Je Zelle 20 Prozesse `-test.count=1` (ein Test-Binary je Arm), Container `--cpus=4` auf einem Host mit 20 Prozessoren, `GOMAXPROCS` 1 und 4, Arme ohne Detektor (`CGO_ENABLED=0`) und mit `-race`; Messender: der Reviewer (Modell claude-sonnet-5-5), Treiber `rate3.sh`; Last des Hosts (1-Minuten-Mittel) 2,0 bis 4,5 vor und nach jeder Zelle der ersten Tabelle, 12 bis 22 bei der Zeile mit 16 Lesern (Größenordnung, `AGENTS.md` §3.10). Spalten: 1 Prozessor ohne Detektor · 1 mit · 4 ohne · 4 mit; in Klammern die Zahl der Prozesse mit `DATA RACE` (DR) und mit `fatal error` der Laufzeit (FE).

| Test | Mutation | 1 ohne | 1 mit | 4 ohne | 4 mit |
|---|---|---|---|---|---|
| `TestCountDiscardedGleichzeitig` | Nullprobe | 0/20 | 0/20 | 0/20 | 0/20 |
| dasselbe | `CountDiscarded` ohne Sperre | 0/20 | 20/20 (DR 20) | 20/20 (FE 20) | 20/20 (DR 20) |
| dasselbe | `Verworfen` im `Snapshot` ohne Sperre in einer Schleife | 0/20 | 20/20 (DR 20) | 20/20 (FE 20) | 20/20 (DR 20) |
| dasselbe | Prüfen und Handeln in getrennten Abschnitten unter `mu` | 0/20 | 0/20 | 20/20 | 20/20 |
| dasselbe | `Verworfen` über `Kopie()` ohne Sperre | 0/20 | 0/20 | 17/20 (FE 17) | 2/20 (FE 2, DR 0) |
| dasselbe mit 16 statt 2 Lesern (Wegwerfbaum), Nullprobe | keine | 0/20 | 0/20 | 0/20 | 0/20 |
| dasselbe mit 16 Lesern | `Verworfen` über `Kopie()` ohne Sperre | 0/20 | 0/20 | 20/20 (FE 20) | 7/20 (FE 7, DR 0) |
| dasselbe | Zweig des **ungültigen** Codes ohne Sperre | 0/20 | 0/20 | 0/20 | 0/20 |
| dasselbe | Zweig des **ungültigen** Codes unter `idMu` | 0/20 | 0/20 | 0/20 | 0/20 |
| `TestCountDiscardedBlockierteSenke` | Nullprobe | 0/20 | 0/20 | 0/20 | 0/20 |
| dasselbe | `CountDiscarded` unter der Sperre der Zeile | 20/20 | 20/20 | 20/20 | 20/20 |

Die Raten des Plans stimmen in der Größenordnung (Plan: `Kopie()` ohne Sperre 25/30 und 4/30 bei 4 Prozessoren, hier 17/20 und 2/20; die Zellen mit 0 oder allen Prozessen sind gleich). Der Detektor meldet das Lesen über `maps.Clone` in keinem der 80 Prozesse mit Detektor (DR 0); der Fänger ist die Laufzeit (`fatal error: concurrent map clone and map write`). Auch mit acht Mal so vielen Lesern gibt es unter dem Detektor keinen Fänger in jedem Prozess.

**Ganze Suite des Adapters, ein Prozess je Mutation, mit Detektor (`CGO_ENABLED=1`):** Zweig des ungültigen Codes von `CountDiscarded` ohne Sperre: **grün** (`ok`, 77 s); unter `idMu`: **grün** (`ok`, 95 s). Kontrolle am Bestand, derselbe Eingriff in `CountError`: **rot** (`WARNING: DATA RACE`, `TestCountErrorGleichzeitig`: `Summe 29156 plus 2754, erwartet 32000`, `fatal error: concurrent map writes`).

**Bestandstests gegen die Kopie von `Errors` und `Dropped` ohne Sperre:** vom Reviewer nicht nachgemessen (die Sperre `/tmp/odata-recorder-heavy.lock` war über eine Stunde von anderen Läufen belegt). Beleg des Vorgängers: die Verifikation von `slice-telemetry-port` (Mutation W4, „`Snapshot`: Sperre vor `Kopie()` freigegeben“, `TestCountErrorGleichzeitig` und `TestAdapterGleichzeitig`) nennt 50 von 50 Prozessen rot mit und ohne Detektor, DR 0, Absturz in 50 (`concurrent map`), bei 16 Schreibern und 16 Lesern; diese Tests schreiben `Verworfen` nicht.

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Kein Test ruft `CountDiscarded` gleichzeitig mit einem Code außerhalb der vier auf: `TestCountDiscardedGleichzeitig` zählt allein die vier Codes der Menge, der Zweig des ungültigen Codes (`fehlercode`) läuft nur nacheinander. Die Mutationen „Zweig des ungültigen Codes ohne Sperre“ und „unter `idMu`“ bleiben in der ganzen Suite des Adapters mit Detektor grün (0 von 20 Prozessen in allen vier Armen), während derselbe Eingriff in `CountError` rot wird (`TestCountErrorGleichzeitig` führt beide Zweige); `SPEC-232` Nr. 5 a sagt zu, dass `CountDiscarded` keinen Zuwachs verliert, `SPEC-230` Nr. 11 zählt `fehlercode` für jeden anderen Code. | `AGENTS.md` §3.15, §3.10; Skill: *Fänger führt nur eine Form der Eingabe* | `src/adapters/driven/telemetry/verworfen_test.go` · `func TestCountDiscardedGleichzeitig`; `src/adapters/driven/telemetry/telemetry.go` · `func (a *Adapter) CountDiscarded` | ja: Mutationen `nl-cd-ungueltig-ohne-sperre`, `nl-cd-ungueltig-unter-idmu` (siehe oben); `make test-race` bleibt grün | Fänger führt nur eine Form der Eingabe (Zweig des Aufrufs, nebenläufig) |
| F-2 | LOW | Der Typkommentar des Adapters nennt zwei Ports („Adapter erfüllt Telemetry und TelemetryReader“); der Adapter erfüllt auch `FallbackTelemetry` (die Zusicherung steht darüber, der Paketkommentar ist nachgezogen). | `AGENTS.md` §3.9, §3.7 | `src/adapters/driven/telemetry/telemetry.go` · `// Adapter erfüllt Telemetry und TelemetryReader` | nein | Berichtigung erreicht nicht jede Stelle |
| F-3 | LOW | Der Plan zählt „51 Mutationen, davon 26 an Elementen, die der Plan nicht nennt (mit `*`)“; die Tabelle trägt 27 Einträge mit `*` (P1 2, P2 1, P3 5, P4 7, P5 3, P6 3, P7 1, P8 3, P9 2). | `AGENTS.md` §3.9, §3.15 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid-port.md` · `davon 26 an Elementen, die der Plan nicht nennt` | nein | Zahl im Plan weicht von der Tabelle ab |
| F-4 | LOW | Der Satz zur Nullprobe nennt für P8 und P9 „0 von 30 Prozessen je Arm“; die Tabelle der Raten führt P9 mit 20 Prozessen je Zelle (`Nullprobe P9`, 0/20) und der Absatz davor nennt 20 bei P9. | `AGENTS.md` §3.9, §3.10 | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid-port.md` · `bei P8 und P9 in 0 von 30 Prozessen je Arm rot` | nein | Zahl im Plan weicht von der Tabelle ab |
| F-5 | INFO | Offene Rückgabe zur Zeile P8 (§6, Adresse Planner): Das Lesen von `Verworfen` über `Snapshot.Kopie()` ohne Sperre hat keinen Fänger in jedem Prozess, und der Vertrag verlangt keinen (Belege unten); das Negativ (12) stimmt und bleibt, die Zeile P8 nennt den Detektor als Fänger für das Lesen ohne Sperre allgemein, was nur für die Schleife gilt. | `AGENTS.md` §3.10, §3.15; `SPEC-232` *Ungeprüft* | `docs/plan/planning/in-progress/slice-telemetry-einbau-hybrid-port.md` · ``Lesen von `Verworfen` über `Snapshot.Kopie()` ohne Sperre`` | ja: Mutation `nl-snapshot-kopie-nur-verworfen-ohne-sperre`, Raten oben | Detektor sieht `maps.Clone` nicht (akzeptiertes Negativ bestätigt) |
| F-6 | INFO | `Errors` und `Dropped` liest dieselbe `Kopie()` über `maps.Clone`; `TestCountDiscardedGleichzeitig` schreibt beide nicht, die Mutationen „Kopie ohne Sperre“ je Sammlung sind dort in 0 von 20 Prozessen rot. Die Bestandstests der Sammlungen (`TestCountErrorGleichzeitig`, `TestAdapterGleichzeitig`) fangen sie nach dem Beleg des Vorgängers in 50 von 50 Prozessen über den Absturz der Laufzeit; hier nicht nachgemessen. | `AGENTS.md` §3.15 | `src/adapters/driven/telemetry/telemetry_test.go` · `func TestAdapterGleichzeitig` | ja: Mutationen `nl-snapshot-kopie-nur-errors-ohne-sperre`, `nl-snapshot-kopie-nur-dropped-ohne-sperre` | Detektor sieht `maps.Clone` nicht |

**Belege zu F-1.** Die Menge M-Zuwachs des Plans (16 × 4 × 1000) hat die Achsen Goroutine, Code und Runde; der Zweig des Aufrufs (Code der Menge oder nicht) steht nicht darunter, obwohl `CountDiscarded` sie verschieden behandelt (`a.verworfen[code]++` oder `a.dropped[support.VerwurfFehlercode]++`). Die Aufrufer von `CountDiscarded` in Tests (`grep`): `verworfen_test.go` 87, 100, 119, 134, 146, 174, 383 (allein 383 ist gleichzeitig, und dort nur mit der Menge), 459; `telemetry_test.go` 88, 101 (nacheinander). Zeile P9 führt beide Zweige (`0` neben den vier Codes), Zeile P8 nur einen; der Plan trägt dafür kein Negativ.

**Belege zu F-5.** `SPEC-232` *Ungeprüft* hält fest, dass die Sperre in Nr. 5 b nur durch einen Wirkungs-Fänger belegt ist, der mit Wahrscheinlichkeit rot wird, und sagt zu `Verworfen` nichts anderes; `SPEC-321` Nr. 5 und Nr. 9 nennen den Zähler, keinen Fänger für das Lesen. `AGENTS.md` §3.10 nennt für einen Zugriff, den der Detektor nicht meldet, das akzeptierte Negativ mit Grund; der Grund (kein Argument, an dem ein Test anhalten könnte, weder an `CountDiscarded` noch an `Snapshot`) trifft, ein Haltepunkt im Port ist nicht möglich. Gemessen: Detektor 0 von 80 Prozessen mit Meldung, Laufzeit 17 von 20 und 2 von 20, bei einem Prozessor 0 von 40; mit 16 statt 2 Lesern 20 von 20 und 7 von 20 bei 4 Prozessoren, 0 von 40 bei einem Prozessor, also auch mit mehr Last kein Fänger in jedem Prozess. Die Folge im Betrieb ist der Abbruch des Prozesses durch die Laufzeit, kein falscher Zahlenwert; Negativ (12) sagt „fällt nur durch die Laufzeit oder die Folge der Zähler im Stand auf“ und trägt das Maß. Der Architect muss nicht entscheiden: kein Wortlaut von `SPEC-232` Nr. 5 b verlangt den Fänger.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/support/telemetry.go` (Port `FallbackTelemetry` mit den drei Operationen von `Telemetry` und `CountDiscarded`, `SpanHybridFallback`, sechste Zeile in `Spans()` mit 13 Attributen, `VerworfeneCodes()` in der Reihenfolge der Spezifikation, frische Slices, `Snapshot.Verworfen`, `Kopie()` tief und nil-treu) | geprüft, ohne Befund (13 eigene Mutationen rot: Stelle und Inhalt der sechsten Zeile, Konstante, geteilter Slice, fünfter Code, nil und leer, Einbettung) |
| `src/hexagon/ports/support/supporttest/telemetrie.go` (`CountDiscarded` hält den Code unverändert fest und zählt; `Aufrufe()` liefert eine Kopie; leerer Stand mit fünf nicht-nil Sammlungen; `OpCountDiscarded` am Ende der Folge) | geprüft, ohne Befund (9 eigene Mutationen: 8 rot aus dem Grund, 1 Aufbaufehler, nicht gezählt) |
| `src/adapters/driven/telemetry/telemetry.go`, `regeln.go` (`CountDiscarded` unter `mu`, nicht unter `zeilenMu`, ohne Uhr, ohne `CountError`; `Verworfen` im Stand; zwei Schlüssel mit Form Token und Art String, allein am eigenen Span, nicht in der Menge Log) | geprüft, mit F-1 (Zweig des ungültigen Codes, nebenläufig), F-2 (Typkommentar), F-5 und F-6; 15 eigene Mutationen: 14 rot, 1 äquivalent |
| Form der Schlüssel (`recorder.fallback_type`, `recorder.fallback_result`: leer, Großbuchstabe, Byte ab `0x80`, 33 Zeichen, Ziffer, Unterstrich, Punkt, Bindestrich, Leerraum, 18 gültige Werte; Paar von Verstößen `wert` 2 und `pflicht` 2) | geprüft, ohne Befund; Zählung 12, 24 und 36 Läufe gelesen, `[]byte{0x80}` für das Byte; keine Fixture gleicht dem Wert einer Mutation |
| M-Codes und M-Spans gegen den Wortlaut (`SPEC-230` Nr. 11 und `SPEC-109`: 41 Codes gleich 4 + 37, Randwerte 0, -1 und 42 liegen auf beiden Seiten der Tabelle; 6 × 13 gleich 78 Paare, 13 zulässig; 6 Paare Art × Ergebnis × 2 Wege) | geprüft, ohne Befund; Zahlen gezählt und gleich denen der Tests |
| Variante A (`Telemetry` bei drei Operationen, `odr_errors_total` unverändert, `CountError` ändert `Verworfen` nie, `CountDiscarded` ändert `Errors` nie) | geprüft, ohne Befund (`TestTelemetrySignaturen`, `TestReihenGetrennt`; die Mutation „zählt zusätzlich in `Errors`“ ist rot) |
| Alte Tests der Zahlen der Menge der Spans (sieben Dateien im Adapter, `support/telemetry_test.go`, `supporttest/telemetrie_test.go`) | geprüft, ohne Befund: der Diff ändert Zahlen, Namen und Kommentare und fügt Fälle hinzu; keine Assertion entfällt (die Löschungen im `git diff` gelesen: nur ersetzte Zahlen, aus vier Sammlungen werden fünf); kein Fänger gelockert (`AGENTS.md` §3.6) |
| Tests außerhalb des Blocks, die `support.Spans()` durchlaufen (`services/mock`, `services/record`, `services/replay`, `adapters/driving/httpodata`, `src/bootstrap/`, `tests/`) | geprüft, ohne Befund: jeder sucht den Namen oder liest Stelle 0, keiner zählt Zeilen oder `Durations`; `go test ./src/...` im frischen `make test` grün |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft, mit F-2; sonst ohne Befund (kein Konjunktiv, kein abwesender Text; „allein der Hybrid-Dienst erhält ihn“ steht in `SPEC-230` Nr. 11 und gehört zu Teil 2) |
| Plan-Nachzug (§3.9) und Raster (§3.16) | geprüft, mit F-3 und F-4; Kopf, §1, §3, §4 (Messpunkt beim Handoff), §6 (Raster, Rückgabe zu P8 mit Ausgang *offen* und Adresse) stimmen mit dem Code überein |
| Katalog und Handbuch (`SPEC-201`, `SPEC-200`) | geprüft, ohne Befund: kein Fehlercode, Schlüssel oder Kommando erreichbar gemacht, das Binary ändert sich nicht |
| ADR und Hard Rules (Docker-only, Suppression, Gate nicht gelockert) | geprüft, ohne Befund: kein ADR berührt, kein Gate geändert |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 3 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (Zweig des Aufrufs, nebenläufig) · Berichtigung erreicht nicht jede Stelle · Zahl im Plan weicht von der Tabelle ab · Detektor sieht `maps.Clone` nicht (akzeptiertes Negativ)

## Verdikt

**Merge-blockierend:** ja, wegen F-1 (MEDIUM). F-2 bis F-4 sind Nachzüge im selben Commit wie der Fix; F-5 und F-6 bestätigen akzeptierte Negative und verlangen keine Aktion am Code.

**Zur offenen Rückgabe zu P8 (Adresse Planner, §6):** Das Negativ (12) wird bestätigt; `SPEC-232` Nr. 5 b und `SPEC-321` verlangen keinen Fänger in jedem Prozess für das Lesen von `Verworfen` über `Kopie()`; ein Konflikt, der den Architect braucht, liegt nicht vor. Die Zeile P8 der Tabelle und das Negativ (12) nennen dann dieselbe Grenze (F-5).

**Übergabe:** Findings gehen an den Implementer; die Finding-Klassen gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ersetzt keine Verifikation — DoD- und Spec-Konformität prüft der Verifier separat.

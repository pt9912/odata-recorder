# Verifikation slice-if-match (If-Match und If-None-Match, `ODR-E6010`)

**Rolle:** Verifier (einmal). **Datum:** 2026-11-09. **Stand:** Zweig `slice-if-match`, HEAD `e1a41ccc` (Fix-Runde nach dem [Review](2026-11-08-slice-if-match-review.md)). **Bezug:** [LH-FA-15](../../spec/lastenheft.md) (`LH-FA-15.a`), [LH-FA-07](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
**Plan:** [slice-if-match](../plan/planning/in-progress/slice-if-match.md); Block `basis` `89f4747d`, `marke` 3400.

## Urteil

Die Zusagen des Slice sind gehalten: Vertrag `SPEC-283` Nr. 1 bis 10, `SPEC-280`/`SPEC-281` (Nachzüge), `SPEC-109`, `SPEC-195` Nr. 10 und der Katalog stimmen mit dem Code überein; die Tabelle F1 bis F29 hat zu jeder Zeile einen vorhandenen Fänger; die Gates sind grün, `test` und `test-race` auch ohne Cache. Keine Abweichung, die den Übergang nach `done/` hindert.
Die Fix-Runde deckt die Klasse M1 (Version 0 bis 2^63-1 bei `If-None-Match: *`) im Modell vollständig und im Dienst für den Stern; sie lässt zwei Lücken, die die genannten Mutationen nicht zeigten (Auflage 2 und 3, beide nicht blockierend). Eine Auflage (1) geht an den Integrations-Planner: nach dem Merge mit `main` werden vier Tests des Pakets `policy` rot.

**Auflagen**

1. **Integration (Planner), blockierend für den Merge, nicht für `done/`:** Nach einem Merge mit `main` (`79c21077`) sind `TestNullwerte`, `TestAusloeser`, `TestMatrix` (alle `entscheide_test.go`, Zeilen 150, 239, 391) und `TestTotalitaet` (`total_test.go:193`) im Paket `src/hexagon/services/hybrid/policy` rot: `model.AlleCodes() liefert 41 Codes, erwartet 40`. Die siebente Zählstelle ist damit eingetreten (Plan §3 *bedingt*, §6 *Angenommen*); der Implementer trägt die Nachzüge nach dem Merge (40 auf 41, 43 auf 44, die Mengen 62 608, 774 000 und die Zählung von Hand 1176 / 110 neu berechnet). Wer als zweiter nach `main` geht, zieht sie nach.
2. **Fänger (Implementer, vor Closure):** Der Dienst gibt die Version unverändert an `model.VorbedingungErfuellt`; Tests führen sie dort mit einem Tag nur bis 100 (F1: 0, 1, 9, 10, 99, 100; F2: 0, 1, 7, 10), größere Versionen nur mit `*` (`TestIfMatchSternBeiJederVersion`). Die Mutationen d8 und d9 unten (Version als `int32` bzw. auf das niedrige Byte gekürzt) an der Aufrufstelle bleiben grün. Entweder Versionen 2^32 und 2^63-2 in `TestIfMatchPasst` und `TestIfMatchVeraltet` aufnehmen oder als akzeptiertes Negativ mit Grund und Folge in die Prüfgrenze (1) im Plan schreiben (die steht heute nur für das Modell).
3. **Fänger oder Negativ (Implementer, vor Closure):** `SPEC-283` Nr. 9 d sagt „die Auswertung prüft keinen Kontext“. Kein Test übergibt einen beendeten Kontext (`ctx`) zusammen mit einem nicht erfüllten Header; die Fakes prüfen `ctx` nie, der Abbruch kommt als Wert eines Ports (F29). Die Mutationen g1 (Auswertung übersprungen, wenn `ctx.Err() != nil`; schreibt trotz nicht erfüllter Bedingung) und g2 (`ctx.Err()` vor der Auswertung zurückgegeben) bleiben grün. Einen Fall mit beendetem `ctx` und gesunden Ports (Erwartung `ODR-E6010`, `Put` 0) ergänzen oder die Aussage als akzeptiertes Negativ mit Grund (der Dienst ruft `ctx.Err()` nirgends, jeder Abbruch der Tests kommt über einen Port) in den Plan schreiben.
4. **Plan (Implementer/Planner bei Closure):** §6 trägt die Ausgänge *offen bis zur Closure* (Risiken, Randformen, Zählstellen, Merge-Reihenfolge); die DoD-Häkchen, §7 (Closure-Notiz) und die Register-Zeile fehlen erwartungsgemäß. Die Rate SQLite (unten) gehört mit den Messwerten dieses Berichts in §3.

## Belege der Sensoren (selbst gefahren)

Alle Docker-Läufe unter `flock /tmp/odata-recorder-heavy.lock`, `--memory=4g`, `--network none`, ohne Mount (Eingaben per `docker cp`); Last des Hosts beim Lauf 3 bis 11 bei 4 CPUs des Containers. Die Läufe waren durch den `flock` verkettet.

| Sensor | Ergebnis |
|---|---|
| `make gates` (einmal, am HEAD) | Exit 0; `docs-check` 829 Dateien, 0 Befunde; `baseline-verify` OK; die Stufen `build`, `test`, `race`, `lint`, `katalog`, `handbuch`, `zeichen` kamen größtenteils aus dem Layer-Cache der Läufe des Implementers (gleiche Eingaben) |
| `make test` mit `--no-cache-filter test` (Stufe `pruefung test`: `gofmt`, `go vet`, alle Tests, ungecacht) | Exit 0, Stufe 225 s |
| `make test-race` mit `--no-cache-filter race` (ungecacht) | Exit 0, `pruefung race: ok`, 228 s |
| `make katalog-check`, `make handbuch-check` | je Exit 0 |
| F23, Gegenprobe in einer Kopie des Baums ohne die Zeile `ODR-E6010` des Handbuchs, `docker build --target katalog` | `katalog: Fehlercodes: ODR-E6010 fehlt im Handbuch`, Exit 1; mit der Zeile grün (`make katalog-check`) |
| Größenmarke (eigene Messung `git diff --numstat 89f4747d..HEAD` über die Pfade des Blocks, ohne `policy`) | 1574 (1554 + 20) von 3400, Auslastung 0,46; der Handoff meldete 1478, die Fix-Runde fügt 96 hinzu |
| `git status` des Worktrees | sauber; die Läufe schrieben nur in den Scratchpad |

Nicht gefahren: `make test-integration`, `make test-determinism` (Pflichtlauf vor der Welle-Closure, nicht Teil dieser Verifikation), `make gegenproben` (kein Werkzeug des Harness geändert; `git diff 89f4747d..HEAD --stat -- tools Makefile harness Dockerfile .golangci.yml .a-check.yml` ist leer).

## Plan-vs-Code-Diff

- Berührt sind: `src/hexagon/model/` (`vorbedingung.go` 105 Zeilen neu, `fehler.go` +2, Tests), `src/hexagon/services/mock/schreiben.go` (+19 -2, Produktcode), Testdateien (Zählstellen und Fänger), `src/adapters/driving/httpodata/if_match_test.go`, `src/bootstrap/if_match_ende_zu_ende_test.go`, Handbuch (Zeile `ODR-E6010`, Satz in `ETAG_ENABLED`), `spec/spezifikation.md` (Fix-Commit des Architect: Nr. 6), der Plan und der Review-Report. Kein Produktcode im Adapter (F27 nur als Test) und keine Änderung an `konfig.go`/`wurzel.go`, wie der Plan zusagt.
- Produktcode 126 Zeilen gegen Block 200: Der Block ist eine Erwartung (`SPEC-128`), keine Grenze; die Messung liegt unter der Marke.
- **F29** steht in der Tabelle des *Nachtrags* der Fix-Runde (Spalten `Zusage · Fänger · Mutation · roter Test`, ohne `Klasse`), nicht in der Tabelle *Zusage · Fänger*; `SPEC-128` Nr. 3 a liest nur die Tabelle mit den fünf Köpfen, die Zeilenzahl 28 und die Schätzung 2460 bleiben. Zählte F29 mit, wäre *S* = 2540 und *M* = 3500; die Messung 1574 ändert dadurch nichts (INFO).
- Alle in Tabelle und Bericht genannten Tests (37 Namen) existieren je einmal (`grep '^func <Name>('` über `src`).

## Zählstellen der Codes (40 auf 41)

Quelle `src/hexagon/model/fehler.go` (`CodeE6010` hinter `CodeE6009`, `codeZeilen`: 412, `mock`, „Precondition failed“). Zählstellen mit Test: `model/fehler_test.go`, `services/mock/lesen_fehler_test.go` (16 erreichbare), `services/mock/rang_test.go`, `services/mock/lesen_gleichzeitig_test.go`, `services/mock/schreiben_weitere_test.go`, `adapters/driven/audit/audit_test.go`, `adapters/driven/telemetry/regeln_test.go`, `services/snapshot/fehler_test.go`, das sind acht Dateien plus die Quelle. Gesucht nach weiteren Stellen: `grep` nach `AlleCodes()` über `src`, nach `CodeE6009`/`ODR-E6009` über `src`, `tests`, `tools`, `harness`, `spec`, `docs/user`: Die Listen in `src/bootstrap/vertrag_http_test.go:586` und `tests/integration/vertrag_fehlerklassen_test.go:379` sind Teilmengen (Codes, die der Eingang liefert), keine Zählung; `tools/katalog/*_test.go` führen „40 Codes“ einer eigenen Fabrik, unabhängig von `fehler.go` (`make katalog-check` grün). Die Spezifikation trägt 41 (`SPEC-230` Nr. 8, Zeile 3957), 16 erreichbare Codes (`SPEC-195` Nr. 10) und die Zeilen in `SPEC-109` und im Katalog. **Siebente Stelle:** Paket `policy`, siehe Auflage 1 (im Wegwerf-Merge mit `main` gemessen: vier Tests rot). Im Wegwerf-Merge mit `slice-hybrid-service` sind die `hybrid`-Tests grün (`TestCodesSchleife` prüft `>= 40`).

## Merge-Proben (`git merge-tree`, Wegwerf-Baum aus `git archive`)

- Gegen `main` (`79c21077`): ein Textkonflikt, `spec/spezifikation.md` am Ende in der Änderungstabelle (Zeilen 4111 bis 4117: drei Zeilen aus `main` gegen die Zeile des Architect); beide Seiten bleiben. Kein Konflikt in Go oder Plänen. `go test` des Wegwerf-Baums: `model`, `mock`, `snapshot`, `inspection`, `audit`, `telemetry`, `httpodata` grün; `policy` rot (Auflage 1).
- Gegen `slice-hybrid-service` (`e8345367`): ein Textkonflikt am selben Ort der Änderungstabelle, beide Seiten bleiben; `SPEC-283` und `SPEC-284` stehen getrennt (Zeilen 3341 und 3454), und `SPEC-284` nennt `412` des Mock als Zweig ohne Fallback (kein Widerspruch zu `SPEC-283` Nr. 7). `MockUseCase` ändert der Zweig nicht. `go build ./...`, `go vet ./src/...` und `go test` über `src/hexagon/...`, `src/adapters/driving/...`, `src/bootstrap` im Wegwerf-Baum: alle grün außer `policy` (Auflage 1) und `admin/TestWettlaufNebenlaeufig`, der im Gesamtlauf rot war und in drei Wiederholungen allein grün blieb (Last des Gesamtlaufs; `admin` liegt außerhalb dieses Slice, nicht weiter verfolgt).

## Vertrag `SPEC-283` gegen den Code

| Nr. | Ergebnis |
|---|---|
| 1 Wann und für wen (nur `PATCH`/`PUT`/`DELETE`, nur `Dienst.Etag`) | gehalten: `schreibe` setzt `vorbedingung` nur bei `d.Etag` und Methode ungleich `POST`; F24, F25, F26 |
| 2 Header lesen (alle Zeilen, Schreibweise, Reihenfolge der Namen, gesetzt = mindestens eine Zeile) | gehalten: `werteVon`; F9, F10 |
| 3 Grammatik (Leerraum nur Leerzeichen/Tab, leere Elemente, `"a,b"`, `W/`, `*` allein, Bytemenge) | gehalten: `liesEtagListe`, `liesTag`, `etagZeichen` gelesen; `TestVorbedingungGrammatik` 30 271 Werte |
| 4 schwacher Vergleich auf den Bytes | gehalten: Tag ab erstem `"`, `slices.Contains` |
| 5 a bis d | gehalten (`vorbedingung.go`); `*` nur bei vorhandenem Eintrag (Nr. 6 sichert das) |
| 6 Rangfolge und Ort | gehalten: Aufruf in `imWorkspace` nach `entscheide`, vor `Put`; die Spezifikation führt seit der Fix-Runde `Workspaces.Get` 2 und den Rang des Programmfehlers als akzeptiertes Negativ, der Test zählt 2 (`TestIfMatchVeraltet`) |
| 7 Antwort (412, `mock`, feste `message`, ohne Ursache, kein `ETag`) | gehalten: `TestCodeTabelle`, F21 |
| 8 a bis d Nebenläufigkeit | gehalten durch Wirkungs-Fänger (unten); keine Sperre im Dienst |
| 9 a bis d | gehalten: F19, F20, F29; 9 c ist akzeptiertes Negativ der Spezifikation; **9 d zweite Hälfte „prüft keinen Kontext“ ohne Fänger** (Auflage 3) |
| 10 `GET`/`HEAD` | gehalten: F26 |

## Tabelle Zusage · Fänger gegen den Vertrag

F1 bis F27 und F29 haben je einen Fänger, der laut Bericht des Implementers (58 Mutationen) rot war; ich habe die Tests ausgeführt (grün), die Mengen in den Zählprüfungen gelesen (`n != 120`, `3*10*2`, `18`, 5 × `etagKlasse` mit mindestens 4000 Versionen) und Mutationen unten selbst gefahren. F28 ist ein akzeptiertes Negativ mit Grund und Folge (`SPEC-283` *Ungeprüft*). F23 ist belegt durch die eigene Gegenprobe. Zu den Zusagen ohne Fänger gehört nur der Teil von Nr. 9 d (Auflage 3).

## Fix-Runde gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

Messender: `tools/harness/mutationslauf.sh` (`SPEC-210`, Image `odata-recorder:race`, `--memory=4g`, Klasse `andere`, ein Prozess je Stufe, 30 Mutationen in 69 s, Last 4,1 bis 4,6 bei 4 CPUs). Die Liste habe ich aus dem Wortlaut der Zusage gebildet, nicht aus dem Befund.

### M1: Version 0 bei `If-None-Match: *`

**Menge:** jede Version 0 bis 2^63-1 der Zusage „`*` ist bei vorhandenem Eintrag für `If-None-Match` nie, für `If-Match` immer erfüllt“. Elemente: im Modell 4196 Versionen × 5 Schreibweisen des Sterns × 2 Header (`etagKlasse`: 0 bis 300 vollständig, Zehner- und Zweierpotenzen mit Nachbarn, 2^63-1, 4000 Stichproben) und die Grenze 0/1/2; im Dienst 10 Versionen × 3 Methoden × 2 Header. Die Menge ist unbeschränkt (2^63); Prüfgrenze: die 4196 Elemente, Folge: ein Fehler, der nur bei einer nicht gezogenen Zahl auftritt, bliebe ungefangen (Plan, Prüfgrenze 1, mit Grund: der Vergleich arbeitet auf der Zeichenfolge).

| Mutation (nicht vom Befund oder vom Implementer genannt) | Ort | Ergebnis |
|---|---|---|
| v1 Stern zählt nur ab Version über 2 | `If-None-Match` | rot |
| v2 Stern zählt nicht bei 2^32 | `If-None-Match` | rot |
| v3 Stern zählt nur bei geraden Versionen | `If-None-Match` | rot |
| v4 Stern zählt nicht ab 2^63-2 | `If-None-Match` | rot |
| v5 Stern zählt nicht bei 7 | `If-None-Match` | rot |
| v6 Stern zählt nicht bei 1000 | `If-None-Match` | rot |
| v8 Stern zählt nur bis 2^62 | `If-None-Match` | rot |
| v9 `If-Match: *` nicht bei 0 | `If-Match` | rot |
| v10 `If-Match: *` nicht bei 2 | `If-Match` | rot |
| v11 `If-Match: *` nicht ab 100 | `If-Match` | rot |
| v12 `If-Match: *` nicht bei 2^63-1 | `If-Match` | rot |
| d1 bis d5 Dienst wertet nicht aus bei 2; 3; ab 2^32; ab 100; nur bei ungeraden | Aufrufstelle | rot (5 von 5) |
| d6 Dienst wertet bei 5 nicht aus | Aufrufstelle | `rot-paket` (`TestIfMatchListePosition` fängt es, nicht der benannte Test) |
| d7 Dienst reicht die Version mit gesetztem Bit 0 | Aufrufstelle | rot |
| v7 Stern zählt nicht bei **12345** | Modell | **grün**: die Zahl liegt weder in 0 bis 300 noch auf einer Potenz noch in den Stichproben; erwartetes Negativ der unbeschränkten Menge, im Plan benannt |
| d8 Dienst reicht die Version als `int32` | Aufrufstelle | **grün** |
| d9 Dienst reicht nur das niedrige Byte der Version | Aufrufstelle | **grün** |

d8 und d9 sind Lücken der Klasse „Version am Dienst“: F1 und F2 prüfen den Dienst nur bis Version 100, die großen Versionen nur mit `*` (Auflage 2). Das Modell ist durch `TestVorbedingungVersionen` (8392 Formen) gegen jede Verkürzung der Zahl gedeckt; die Lücke liegt allein im Durchreichen am Aufruf.

### M2: Abbruch (Fehler oder Kontextabbruch an `Entities.Get` zusammen mit nicht erfülltem Header)

**Menge:** Operation (3: `PATCH`, `PUT`, `DELETE`) × Header (2) × Fehlerart (3: Port-Fehler mit Code `ODR-E3001`, `context.Canceled`, `context.DeadlineExceeded`) = 18 Läufe in `TestIfMatchAbbruchEntitiesGet`; `Put`, `Advance`, `Coverage.Get`, Schreiber je 0. Ports der Transaktion, an denen ein Abbruch der Auswertung vorgehen kann: `Workspaces.Get` (läuft vor `Entities.Get`, die Auswertung kann ihm nicht vorgehen) und `Entities.Get`; `Put` und `Advance` laufen hinter der Auswertung und werden bei `ODR-E6010` nicht erreicht.

| Mutation (keine der `m2-…` des Implementers) | Ergebnis |
|---|---|
| e1 Fehler verschluckt, wenn `If-None-Match` gesetzt | rot |
| e2 Fehler verschluckt, wenn `If-Match` gesetzt | rot |
| e3 Fehler verschluckt bei `DELETE` mit Vorbedingung | rot |
| e4 Fehler verschluckt bei `PUT` mit Vorbedingung | rot |
| e5 Kontextfehler verschluckt | rot |
| e6 Port-Fehler wird bei nicht erfüllter Bedingung `ODR-E6010` | rot |
| e7 Kontextfehler wird bei nicht erfüllter Bedingung `ODR-E6010` | rot |
| g1 Auswertung entfällt, wenn `ctx.Err() != nil` (schreibt) | **grün** |
| g2 `ctx.Err()` vor der Auswertung zurückgegeben | **grün** |

e1 bis e7 sind rot; die Mengen der Fehlerarten und der Methoden sind am Port gedeckt. g1 und g2 liegen in der zweiten Hälfte von Nr. 9 d („prüft keinen Kontext“) und brauchen einen beendeten `ctx` im Aufruf; kein Test übergibt einen (Auflage 3). Das ist ein Satz des Wortlauts ohne Fänger, kein Fehler im Verhalten: der Dienst ruft `ctx.Err()` nirgends (`grep`).

## Nebenläufigkeit F15 bis F18: Raten selbst gemessen

Last des Hosts je Lauf 5,8 bis 7,6 bei 4 CPUs (erster Wert von `/proc/loadavg`), Messender: `docker exec` mit `go test -c` und einem Test-Binary, `-test.run '^<Test>$' -test.count=1`, je Prozess ein Lauf, vier zugleich (außer wo vermerkt). Ohne Detektor (`n_detektor` 0, wie im Plan begründet: der Test zählt Wirkungen am Haltepunkt; `make test-race` ist grün). Die Mutationen sind eigene Fassungen derselben Klassen (Wegwerfkopie, Patch):

| Zusage | Test | Mutation | Prozesse | rot |
|---|---|---|---|---|
| F15 | `TestIfMatchWettlaufPatch` | keine | 20 | 0 (20 grün) |
| F16 | `TestIfMatchWettlaufPut` | keine | 20 | 0 (20 grün) |
| F17 | `TestIfMatchWettlaufFalsche` | keine | 20 | 0 (20 grün) |
| F18 | `TestIfMatchWettlaufDelete` | keine | 20 | 0 (20 grün) |
| F15 | `TestIfMatchWettlaufPatch` | f15: Version in eigener Transaktion davor gelesen und dort verglichen | 20 | 20 von 20 |
| F16 | `TestIfMatchWettlaufPut` | f15 | 20 | 20 von 20 |
| F15 | `TestIfMatchWettlaufPatch` | f15b: Version in eigener Transaktion davor gelesen, in der zweiten Transaktion verglichen | 20 | 20 von 20 |
| F17 | `TestIfMatchWettlaufFalsche` | `Put` vor der Auswertung | 20 | 20 von 20 |
| F15 bis F18 | `TestIfMatchWettlaufSQLite` (echte SQLite, 4 Formen × 5 Aufbauten × 20 Anfragen je Prozess) | keine | 20 | 0 (20 grün) |
| F15 | `TestIfMatchWettlaufSQLite` | f15 (Vergleich vor der Haupttransaktion) | 40 | 40 von 40 |
| F15 | `TestIfMatchWettlaufSQLite` | f15b (veraltete Version in der Haupttransaktion) | 40 | **19 von 40** |
| F15 | `TestIfMatchWettlaufSQLite` | f15b | 20, einzeln nacheinander | **11 von 20** |

**Bewertung des Befunds des Reviewers (14 von 20):** bestätigt, gleiche Größenordnung (19 von 40, 11 von 20, Last 5,8 bis 7,6). Der SQLite-Test fängt die Klasse „Version in eigener Transaktion davor gelesen“ nur mit Wahrscheinlichkeit (rund die Hälfte der Prozesse) und zählt allein nicht als Fänger der Zusage (`AGENTS.md` §3.10). **Das Negativ trägt trotzdem,** weil der Fänger der Zusage der Test mit dem Haltepunkt am Fake ist: er wurde in jeder Fassung (f15, f15b, f17) in 20 von 20 Prozessen rot und ohne Mutation in 20 von 20 grün; der SQLite-Test ist die zweite, echte Transaktion dahinter und im Plan (Nachtrag) mit der Rate benannt. Der Plantext nennt für F15 am besten beide Zahlen (Haltepunkt 20 von 20, SQLite 11 bis 19 von 20 je Fassung) und führt den SQLite-Test nicht als Fänger der Zusage (Auflage 4). Einzelzahlen gelten als Größenordnung, die Last des Hosts schwankt.

## Verbleibende Risiken

- Die Raten der SQLite-Messung sind Größenordnungen unter fremder Last (Host mit mehreren Worktrees).
- Die Prüfgrenzen „Version am Dienst bis 100“ (Auflage 2) und „kein beendeter `ctx`“ (Auflage 3) sind kleine Lücken; kein Client erreicht sie, solange die Ports `ctx` beachten.
- `ODR-E6010` im Modus `hybrid` (F28) bleibt ungefangen bis zur später schließenden Welle; die Policy nennt den Code nicht (`SPEC-261` Nr. 3 c).
- Das Handbuch zeigt kein Beispiel zu `If-Match` (Plan §6, akzeptiert bis `slice-handbuch-mock`); `make handbuch-check` und das Kataloge-Gate sind grün.

## Gegenlesen (`AGENTS.md` §3.18)

Der Bericht enthält keine Folge aus Backslash, `u` und vier Hex-Ziffern; gegengelesen wurde diese Datei mit `LC_ALL=C grep -P '\\u[0-9A-Fa-f]{4}'`.

# Nachverifikation: slice-capture-pipeline

* Datum: 2026-10-21
* Rolle: verifier (frischer Kontext; kein Code, keine Spec, kein Plan, kein Lastenheft, keine `AGENTS.md` geändert)
* Gegenstand: `69e0f56` (Architect) und `10b281e` (Implementer) gegen die Befunde V1 bis V15 der Verifikation `docs/reviews/2026-10-21-slice-capture-pipeline-verifikation.md` (`11071b1`)
* Plan: `docs/plan/planning/done/slice-capture-pipeline.md`
* Maßstab: `SPEC-119` Nr. 4, 8, 9, 12, 14, 15 und *Ungeprüft*; `SPEC-053`, `SPEC-001`, `LH-FA-02.b`, `LH-FA-02.c`, `LH-FA-13.a` (Abschnitte der Spezifikation); `spec/architecture.md`; Lastenheft [LH-FA-02](../../spec/lastenheft.md) und [LH-FA-14](../../spec/lastenheft.md) (nur gelesen); `AGENTS.md` §3.4, §3.9, §3.11, §3.15, §3.16; [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
* Belege selbst gefahren: 15 Mutationen auf Kopien von `src/` (Abschnitt 2), zwei Wegwerf-Proben; sieben Zeitläufe (Abschnitt 3); `make gegenproben` zweimal, einmal mit dem Stempel im Scratchpad (Abschnitt 6); `git diff --numstat` (Abschnitt 7); `make gates` auf `10b281e` vor diesem Bericht (Abschnitt 8). Alle `go test` im Image der Stufe `deps` des Dockerfiles (Go-Image per Digest gepinnt), `--network none`, Quelle schreibgeschützt eingehängt. Der Arbeitsbaum des Repos blieb bis auf diesen Bericht unberührt.

## Urteil

**Freigegeben: kein HIGH, kein MEDIUM.** Die vier MEDIUM V1 bis V4 der Verifikation sind behoben. Die vier Mutanten M38, M10, M64 und M12b sind jetzt aus dem benannten Grund rot, ebenso 10 von 11 eigenen Mutationen. Die Sätze zum Abbruch-Status nennen den Abbruch nirgends mehr pauschal `incomplete`. Es bleiben drei LOW (V16 Architect, V17 Implementer, V18 Architect) und zwei INFO (V19, V20). Beim Planner offen sind V9, V10 (Größe 4609 gegen die Marke 3600), V12 und die Randform des Nehmers. DoD-Häkchen, §7 und die Entscheidung über die Größe trifft der Planner.

## 1. V1, V6, V7, V15: Spezifikation und Sicht (`69e0f56`)

**Abbruch-Status, alle Stellen in `spec/`.** Gesucht nach `incomplete`, `unvollständig`, `Abbruch`, `abgebrochen`, `abort` über `spec/spezifikation.md`, `spec/architecture.md` und `spec/lastenheft.md`:

| Stelle | Wortlaut nach `69e0f56` | Gegen `SPEC-119` Nr. 9, 12 |
|---|---|---|
| `SPEC-119` Einleitung (Zeile 986) | Wert `incomplete` oder `failed`, „welcher gilt, entscheidet Nr. 12“; `incomplete`: Request `Vollstaendig`, Abbruch während der Response, „ohne dass die Capture-Seite sie zuvor verworfen hatte“; `failed`: jede Verwerfung, auch mit späterem Abbruch, und jeder Abbruch im Request-Body | stimmt mit Nr. 12; „zuvor“ siehe V16 |
| `LH-FA-13.a` (Zeile 334) | abgebrochen, `incomplete`; `failed`, wenn der Request-Body nicht beendet war „oder die Capture-Seite einen Body schon verworfen hatte“ | ja (Ursache ist `Abbrechen`, die Zeit ordnet); siehe V20 |
| `LH-FA-02.b` (Zeile 352) | nach dem ersten gesendeten Byte `incomplete`, `failed` wie oben | Ergebnis ja; „schon“ siehe V16 |
| `LH-FA-02.c` (Zeile 360) | nach der Frist abgebrochen, `incomplete`, `failed` wie oben | ja (Herunterfahren ruft `Abbrechen`); siehe V20 |
| `SPEC-053` (Zeile 1128) | Abbruch `incomplete`, außer im Request-Body und „nach einer Verwerfung durch die Capture-Seite, die ihm vorausging“ (Nr. 9, 10, 12) | ja: der Verweis auf Nr. 9 trägt die Reihenfolge |
| `SPEC-001` (Zeile 528) | `discarded` nur bei `failed` oder `incomplete` | ja |
| `architecture.md` Zeile 175 (`RecordUseCase`) | „überfällige abbrechen und mit dem Capture-Status nach `SPEC-119` Nr. 12 schreiben, `incomplete` oder `failed`“ | ja; siehe V20 |
| `architecture.md` Zeile 552 (Sequenz §4.9) | „finished, or aborted as incomplete or failed“ | ja |
| `architecture.md` §2.4, §4.8 | `complete`, `incomplete` oder `failed`; „capture failure without body“ | ja, kein pauschaler Satz |
| Lastenheft [LH-FA-14](../../spec/lastenheft.md) | „als unvollständig markiert“, kein Wert | unverändert (V12, beim Planner) |

Kein Satz in `spec/` nennt den Abbruch mehr pauschal `incomplete`. Lastenheft und `AGENTS.md` sind seit `11071b1` unverändert (`git diff --stat` leer).

**Nr. 14 und 15 gegen Nr. 8, 9, 10.** Nr. 15 sagt jetzt: `Abort` nur bei offenem Writer, nach gescheitertem `Write` oder `Commit`, nicht nach gescheitertem `Begin`. Der Grund ist `ausgabe_zu_gross` bei `outbound.ErrZuGross`, bei einem Kontextfehler und beendetem Kontext der Grund der Ursache, sonst `ablage`. Das deckt sich mit Nr. 8 (Definition von „offen“), Nr. 9 (Kontextfehler) und Nr. 10 (Tabelle) (V6 erledigt). Die falsche Länge (Commit gelungen, kein `Abort`) ist kein „Scheitern“ im Sinn von Nr. 15, also kein Widerspruch. Nr. 14 nimmt den Writer-Aufruf aus, der nie zurückkehrt, und verweist auf *Ungeprüft*; *Ungeprüft* sagt dasselbe mit Rückverweis auf Nr. 14 (V7 erledigt). *Ungeprüft* nennt jetzt die Folge eines Panics in `src.Read`: Er endet in der Goroutine des Aufrufers, die Aufnahme endet spätestens mit der Frist, und der Prozess endet nur, wenn die Goroutine des Aufrufers den Panic nicht fängt. Das passt zu Nr. 14 (V15 erledigt).

**`spec/architecture.md` und `AGENTS.md` §3.4.** Die zwei geänderten Zeilen nennen `SPEC-119` Nr. 12 (aufwärts) und keine Welle, keinen Slice, keinen Hash und keine ADR. §2.2 und §4.9 existieren unter diesen Nummern, die der Plan-Kopf nennt.

**V13, Sendung (b).** `slice-record-service` §1 nennt `RequestRedact` des Zustands, „als `Body` und `RequestRedaction` von `SchluesselEingabe`“. Beide Felder heißen so in `src/hexagon/model/schluessel.go` (Zeilen 39, 40), `RequestRedact` so in `abschluss.go`.

## 2. Mutationen (V2 bis V5, V13)

**Verfahren** wie in der Verifikation: je Mutation eine Kopie von `HEAD` (`git archive`, Scratchpad), Ersetzung mit `perl`, geprüft, dass die Datei sich änderte (Diff je Mutant gelesen), `go test -count=1 -timeout 240s` über das Paket. Alle 15 bauten, keine hing; „an der Zeitgrenze“ heißt: die 10 s des Tests, nicht der Timeout von `go test`. Gezählt wird nur Rot aus dem benannten Grund. **Ergebnis: 14 rot, 1 grün.**

| # | Zeile in §3 | Mutation | Ergebnis |
|---|---|---|---|
| M38 | Frist, `Abbrechen`, `ctx` wirken sofort (Nr. 9) | `queueLeser` entnimmt vorhandene Bytes, bevor er den Kontext prüft (Prüfung nur beim Warten) | rot: `TestUrsacheVorVerstossInDerQueue`, alle drei Fälle `syntax` statt `frist`/`abbruch` |
| MV2a | ebenso | Kontext vor der Entnahme nur geprüft, solange noch kein Rohbyte entnommen ist | rot: dieselben drei Fälle, `syntax` |
| MV2F | ebenso, Fall Frist | Leser übergeht eine beendete Aufnahme nur, wenn die Ursache die Frist ist | rot: nur Fall `Frist` (`syntax`) |
| MV2A | ebenso, Fall `Abbrechen` | dasselbe nur für die Ursache `context.Canceled` | rot: nur Fall `Abbrechen` |
| MV2C | ebenso, Fall `ctx` | dasselbe nur für eine Ursache des Aufrufers | rot: nur Fall `Abbruch von ctx` |
| M10 | `ausstehend` bei Entnahme in Teilen (Nr. 4); schneller Konsument | Teil-Entnahme senkt `ausstehend` nicht | rot: `TestAusstehendBeiEntnahmeInTeilen` (alle Schritte ab „ein Byte“, Fall 50), `TestKeinUeberlaufBeiSchnellemKonsument` (Fall 40000 an der Zeitgrenze) |
| MV3e | ebenso | nur die Teil-Entnahme senkt `ausstehend`, die Entnahme eines ganzen Stücks nicht | rot: beide Fälle des schnellen Konsumenten, Schritte „Teil größer“ (82 statt 30) und „genau das Stück“ |
| MV3c | ebenso | bei einer Teil-Entnahme wird `ausstehend` auf den Rest des ersten Stücks gesetzt (die Stücke dahinter fallen aus der Zählung) | **grün** (V17); Probe rot |
| M64 | `src` schließt die Pipeline nie (Nr. 3) | die Goroutine schließt `src` nach der Verarbeitung | rot: `TestSrcNieGeschlossen`, alle 28 Zeilen von `captureFaelle` |
| MV4a | ebenso | `verwirf` schließt `src` | rot: die 22 Zeilen mit Verwerfung, nicht die Gutfälle und nicht `encoding` |
| MV4b | ebenso | `Tee` schließt `src` bei unbekannter Kette | rot: nur Zeile `encoding` |
| MV4c | ebenso | `Read` schließt `src` nach `io.EOF` (Goroutine des Aufrufers) | rot: 24 Zeilen; nicht rot: `frist`, `abbruch src`, `abbruch Abbrechen`, `abbruch ctx` |
| MV13b | `io.EOF` und Fehler von `src` beenden die Queue (Nr. 4) | nach einem Fehler von `src` legt die Queue weiter ein | rot: `TestNachEndeVorDemErgebnis`, Fall Fehler von `src` (`Ausstehend` 10 statt 7, `syntax` statt `abbruch`) |
| M12b | Verhältnis am Aufruf (Nr. 7b) | `entpackt` am Aufruf um eins zu klein | rot: `TestVerhaeltnisAmAufruf`, Fall 2001 |
| MV5a | ebenso | `roh` am Aufruf um eins zu groß | rot: Fälle 99 und 2001 |

Je Ursache der Nr. 9 trennt der neue Test also seinen eigenen Fall (MV2F, MV2A, MV2C). Das ist mehr, als die Zeile in §3 behauptet.

**Probe zu MV3c** (Wegwerf-Test über `QueueAllein`, nicht committet): 60 und 30 einlegen, 1 Byte entnehmen, `Ausstehend()` muss 89 sein, danach ergeben 12 Bytes einen Überlauf. `HEAD`: grün. MV3c: rot (`Ausstehend = 59`, kein Überlauf bei 89 + 12 > 100).

## 3. Zeitabhängigkeit des Falls Frist in `TestUrsacheVorVerstossInDerQueue`

Vorgebautes Testbinary aus `HEAD` (gebaut im Image der Stufe `deps`), netzlos ausgeführt.

| Lauf | Bedingung | Ergebnis |
|---|---|---|
| A | `-count=50`, nur `TestUrsacheVorVerstossInDerQueue` | grün, 13 s |
| B | ganzes Paket `-count=50` (gleichzeitig mit D, E, E2 und P) | grün, 135 s |
| C | wie A, `--cpus=0.05` | grün, 14 s |
| D | ganzes Paket `-count=1`, `--cpus=0.05` | grün, 21 s |
| E | `--cpus=0.05`, acht Busy-Schleifen im Container, `-count=20`, nur der neue Test | grün, 10 s |
| E2 | wie E bei `--cpus=0.01` | grün, 21 s |
| P | Probe: `time.Sleep(300 ms)` vor dem ersten `Read` des Tests (Kopie) | rot nur im Fall Frist, an der Zeitgrenze |

In keinem Lauf wurde der Test rot, ohne dass der Code falsch war. Lauf P bestätigt das Fenster, das §3 des Plans und der Kopf des Tests nennen: Erreicht die Goroutine den ersten `Write` nicht innerhalb der 250 ms, endet der Test nach 10 s rot. Das ist dieselbe Klasse wie V14 (V19).

## 4. V8: Raster

Die vier Zellen *Wiederholung* nennen jetzt einen Ort: `New` · `MaxBytes`… und `Tee` · `ContentType` bzw. `Metadata` „kein gemeinsamer Zustand (Nr. 14)“, `Tee` · `Richtung` „zwei Aufnahmen ohne gemeinsamen Zustand (Nr. 14); … prüft `Abschluss` nur an der Richtung (Nr. 12)“. Der Kopf erklärt (b) mit Nr. 14, „Aufnahmen teilen keinen Zustand“. Die Sätze in Nr. 14 und Nr. 12 („Ein Ergebnis fremder Herkunft prüft `Abschluss` nur auf die Richtung“) tragen die Entscheidung. Die Zusatzzeile *Frist, `Abbrechen`, Abbruch von `ctx` × Verstoß in der Queue* verweist auf Nr. 9, „sie wirken sofort, ohne Rücksicht auf die Bytes … (sie werden verworfen)“. Keine Zelle ist leer. V8 ist erledigt.

## 5. Plan gegen Code (`AGENTS.md` §3.9), Kommentare (§3.11)

**`69e0f56`** zieht Kopf (`architecture.md` §2.2 und §4.9), §1 (Status-Regel), §3 (Zeilen `spec/spezifikation.md`, `spec/architecture.md` und Fänger-Texte zu Nr. 14, 15) und §6 (Risiko „Normative Sätze …“, Raster-Zeile 260) nach. **Übersehen:** das Risiko in §6, Zeile 269, sagt weiter: „entschieden in `SPEC-119` Einleitung und Nr. 10: `incomplete` ist der Abbruch des Austauschs, `failed` die Verwerfung durch die Capture-Seite“. Die Einleitung sagt das seit `69e0f56` nicht mehr. Die Zeilen 271 und 274 nennen die Ausnahme (V18).

**`10b281e`** zieht §3 (Testdateien mit `QueueAllein`; fünf neue und zwei erweiterte Zeilen), §4 (54 Zeilen, gezählt: Zeile 108 bis 161) und §6 (Raster, Ausgang „Fänger-Lücken der Verifikation“, Größe) nach. Die Mutationsangaben der neuen Zeilen stimmen mit meinen Läufen überein: M64 macht 28 Zeilen rot, M10 die Schritte ab „ein Byte“, den Fall 50 und den Fall 40000, M12b den Fall 2001, M38 alle drei Fälle. Im Produktcode ist nur der Kommentar an `arbeite` geändert (Diff gelesen).

**Kommentare.**

* `arbeite` sagt „endet mit dem Ergebnis; ein Aufruf des Writers, der nie zurückkehrt, hält sie auf (SPEC-119 Nr. 14, Ungeprüft)“. Das entspricht Nr. 14 (V7).
* `queueLeser` sagt „kehrt sofort zurück, sobald der Kontext der Aufnahme beendet ist“. Fängt jetzt: M38, MV2a, MV2F, MV2A, MV2C.
* `TestKeinUeberlaufBeiSchnellemKonsument` benennt die Kopplung des Falls 40000 an den Redaktor ehrlich: „in Teilen, solange der Eingabepuffer des Redaktors kleiner ist (heute 32 KiB; SPEC-118 Nr. 6 sagt die Zahl nicht zu)“. Er verweist für die Zusage selbst auf den Test ohne diese Kopplung. Der Puffer ist `bodyEinPuffer = 32 << 10` in `src/hexagon/model/redaktor.go`, und M10 wird im Fall 40000 rot, also wird dort wirklich in Teilen entnommen. Die Zeile in §3 sagt dasselbe.
* `umg.huelle` („umhüllt jedes `src`, das `tee` und die Treiber von `captureFaelle` an `Tee` geben“) hält, was es sagt: `TestSrcNieGeschlossen` bricht ab, wenn nicht genau ein `src` umhüllt wurde, und M64 färbt alle 28 Zeilen mit „`Close` 1-mal“, keine mit „0 Ströme umhüllt“.
* Die Kommentare von `QueueAllein` in `export_test.go` stimmen mit dem Code (`maxRoh` ist `math.MaxInt64`, die Meldung des Überlaufs bleibt gesetzt).

## 6. Gegenproben-Stempel (V11)

`make gegenproben` auf `10b281e`: „Eingaben unverändert seit dem letzten grünen Lauf (`12810e79…`) — nichts läuft neu“, der Stempel bleibt. Der Stempel ist damit aktuell. Danach ein erzwungener Lauf mit dem Stempel im Scratchpad (`GEGENPROBEN_STEMPEL=<Scratchpad>`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), kein Gate): „33 Fälle, 0 rot“, 76 s, geschriebene Prüfsumme gleich `harness/gegenproben.stempel`. Der Stempel im Repo ist nicht berührt (`git status` leer). V11 ist erledigt.

## 7. Größe (nur Bericht)

`git diff --numstat e085e94..10b281e` über `src/hexagon/services/record/capture/`, `spec/spezifikation.md` und `spec/architecture.md`, Tests eingeschlossen: **4609**. Davon Produktcode 872, Tests 3702 (einschließlich `export_test.go`), Spezifikation 31 (16 + 15), Architektur 4. Das ist der Wert des Implementers. Seit `d7d7737` kamen 218 hinzu. Die Marke 3600 ist um 1009 überschritten (28 %), die Erwartung des Plans (4050) um 559. Die Entscheidung trifft der Planner (V10).

## 8. Gate-Lauf

`make gates` auf `10b281e` vor diesem Bericht: rc 0. Die Stufe `test` lief ungecacht (`pruefung test: ok`, 52 s), `lint` meldet 0 issues, und `migration-check: ok` ist die letzte Zeile. `make docs-check` und `make gates` nach dem Commit dieses Berichts: Ergebnis im Handback an den Planner.

## Stand der Befunde V1 bis V15

| ID | Stand | Beleg |
|---|---|---|
| V1 MEDIUM | erledigt | Abschnitt 1; Rest zur Wortwahl in V16 |
| V2 MEDIUM | erledigt | M38, MV2a, MV2F, MV2A, MV2C rot; Tabellenzeile und Rasterzeile vorhanden |
| V3 MEDIUM | erledigt | M10 und MV3e rot; eine Lücke bleibt (V17) |
| V4 MEDIUM | erledigt | M64, MV4a, MV4b, MV4c rot |
| V5 LOW | erledigt | M12b, MV5a rot |
| V6 LOW | erledigt | Nr. 15 (Abschnitt 1) |
| V7 LOW | erledigt | Nr. 14 und Kommentar an `arbeite` |
| V8 LOW | erledigt | Abschnitt 4 |
| V9 LOW | offen | Planner (Closure): §6, Risiko *Raster*, steht weiter auf „offen bis zum Review“ |
| V10 INFO | offen | Planner: 4609 gegen 3600 (Abschnitt 7) |
| V11 LOW | erledigt | Abschnitt 6 |
| V12 INFO | offen | Planner |
| V13 INFO | erledigt | Tabellenzeile 120 (MV13b rot); Sendung (b) mit `RequestRedact` |
| V14 INFO | unverändert | dokumentiert; dazu V19 |
| V15 INFO | erledigt | *Ungeprüft* (Abschnitt 1) |

## Befunde dieser Nachverifikation

| ID | Kategorie | Befund | Zuständig |
|---|---|---|---|
| V16 | LOW | Zwei der neu geschriebenen Sätze setzen das Kriterium für `failed` nach einem Abbruch in die Zeit: `LH-FA-02.b` (Zeile 352, „die Capture-Seite einen Body schon verworfen hatte“) und die Einleitung von `SPEC-119` (Zeile 986, Abbruch „(Client, Upstream, Herunterfahren), ohne dass die Capture-Seite sie zuvor verworfen hatte“). Ein Fehler von `src` ist aber nach Nr. 9 eine Ursache *im Strom* und wird in der Folge der Bytes gesehen, nicht in der Zeit. Beispiel: Die Response liefert `{"a":x`, dann bricht der Upstream die Verbindung ab, und die Goroutine hat `x` noch nicht gelesen. Nach Nr. 9 ist das `syntax`/`Verstoss` und nach Nr. 12 `failed`, unabhängig vom Tempo (Nr. 4; `TestVerstossVorFehler` fängt es im Code). Wörtlich gelesen sagen beide Sätze `incomplete`, weil beim Abbruch noch nichts verworfen war. Der Wert ist eindeutig, weil die Einleitung auf Nr. 12 verweist, aber die Erläuterung nennt ein Kriterium, das kein Abnahmetest beobachten kann. `SPEC-053` („die ihm vorausging“, mit Nr. 9) und `LH-FA-02.c`/`LH-FA-13.a` (Ursache `Abbrechen`, außerhalb des Stroms, dort ordnet die Zeit) sind richtig. Vorschlag: „oder vor dem Abbruch eine Verwerfung lag; bei einem Fehler von `src` in der Folge der Bytes (Nr. 9)“. | Architect |
| V17 | LOW | Zeile *`ausstehend` sinkt um jedes entnommene Byte, auch bei Entnahme in Teilen* (Nr. 4): `TestAusstehendBeiEntnahmeInTeilen` entnimmt nie in Teilen, während ein zweites Stück in der Queue steht. Darum bleibt MV3c grün (bei der Teil-Entnahme fallen die Stücke dahinter aus der Zählung). Gemutet unterschätzt `ausstehend` die Queue bei jeder Teil-Entnahme, und bei Reads über 32 KiB und langsamem Writer hält sie mehr als `QueueBytes`. Das ist die Grenze aus Nr. 4 und `SPEC-053`. M10, MV3e und die Mutation des Implementers (Länge des Puffers) sind rot, die Zusage hat also einen Fänger. Es fehlt ein Schritt: Teil-Entnahme bei zwei Stücken, etwa 60 und 30, dann 1 entnehmen, ergibt 89 (Probe in Abschnitt 2). | Implementer |
| V18 | LOW | Plan §6, Zeile 269 (Risiko *Widerspruch zwischen Lastenheft … und `SPEC-053`*): „entschieden in `SPEC-119` Einleitung und Nr. 10: `incomplete` ist der Abbruch des Austauschs, `failed` die Verwerfung …“. Die Einleitung sagt das seit `69e0f56` nicht mehr; der Satz gibt den alten, pauschalen Wortlaut wieder (`AGENTS.md` §3.9). | Architect (oder Planner bei der Closure) |
| V19 | INFO | Der Fall Frist in `TestUrsacheVorVerstossInDerQueue` hat das benannte Fenster von 250 ms (Lauf P). Rot wurde er in keinem Lauf, auch nicht bei `--cpus=0.01` mit Last. Fällt er, meldet `warteAufAnkunft` „src wurde nicht erreicht“, obwohl der Test auf den ersten `Write` des Writers wartet. Der Text stammt aus dem gemeinsamen Hilfsmittel (seit `3bcf469`). | — (dokumentiert) |
| V20 | INFO | Zur offenen Randform des Nehmers (Exchange mit beiden Seiten vollständig beim `Drain`): Die Sätze aus `69e0f56` in `LH-FA-13.a` und `LH-FA-02.c` („werden abgebrochen und als `incomplete` geschrieben“, Ausnahme `failed`) und in `architecture.md` Zeile 175 („nach `SPEC-119` Nr. 12 …, `incomplete` oder `failed`“) setzen deren Antwort schon voraus. Sind beide Ergebnisse `Vollstaendig`, ergibt Nr. 12 `complete`: der `Commit` bleibt (Nr. 8), und `Abbrechen` wirkt nach dem Ergebnis nicht mehr (Nr. 14). Was der Planner entscheidet, zieht eine der beiden Seiten nach. Verwandt ist der `Drain` vor dem ersten Byte der Response: es gibt kein Ergebnis der Response, und `Abschluss` braucht eines. Sendung (e) nennt den Austausch ohne Antwort-Strom nur für den Transportfehler. | Planner |

## Zuständigkeiten

* **Architect:** V16 (zwei Sätze, Wortwahl), V18 (eine Zeile im Plan, falls nicht der Planner sie bei der Closure nachzieht).
* **Implementer:** V17 (ein Schritt in `TestAusstehendBeiEntnahmeInTeilen`, MV3c selbst rot sehen, Zeile in §3 nachziehen). Im Produktcode ist nichts zu ändern.
* **Planner:** V9, V10, V12, V20 und die Randform des Nehmers; DoD-Häkchen, §7 mit Steering-Loop-Eintrag (Kandidat aus diesem Lauf: „ein Fänger trennt die Mutation nur in einer Form der Eingabe“, V17 nach V3 und V5), Register `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (dieser Lauf sah keinen roten Zeitlauf).

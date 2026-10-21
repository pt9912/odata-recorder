# Verifikation: slice-capture-pipeline

* Datum: 2026-10-21
* Rolle: verifier (frischer Kontext; kein Code, keine Spec, kein Plan, kein Lastenheft geändert)
* Gegenstand: `git diff e085e94..d7d7737` (Basis `e085e94`: Plan und `SPEC-119` vom Architect; danach `3bcf469` Implementierung, `7bb2784` Architect, `49063d4` Fänger, `e4bb5a2` Review, `5f69ae3` Architect zu F-1 bis F-3, `d7d7737` Fänger zu F-4 bis F-16, F-18, F-19)
* Plan: `docs/plan/planning/in-progress/slice-capture-pipeline.md` (Welle `welle-record-replay`)
* Maßstab: DoD (§2), `SPEC-119` unabhängig gelesen, dazu `SPEC-053`, `SPEC-046`, `SPEC-045`, `SPEC-001`, `SPEC-118`, `SPEC-106`, `LH-FA-02.b`, `LH-FA-02.c`, `LH-FA-13.a` (Abschnitte der Spezifikation), Lastenheft [LH-FA-02](../../spec/lastenheft.md) und [LH-FA-14](../../spec/lastenheft.md); Tabelle Zusage·Fänger (§3), Raster (§6), `AGENTS.md` §3.7, §3.9 bis §3.16; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
* Eingang: DoD-Bestätigung und Messungen des Implementers (Plan §6, Risiko *Größe*), Review `docs/reviews/2026-10-21-slice-capture-pipeline-review.md` (F-1 bis F-19, I-1 bis I-8)
* Belege selbst gefahren: `make gates` auf `d7d7737` (rc 0); `go test -count=1` des Pakets im gepinnten Go-Image (Stufe `deps` des Dockerfiles, netzlos, Quelle schreibgeschützt eingehängt; grün, 1,9 s); Zeitläufe A bis G2 (Abschnitt 5); 24 eigene Mutationen auf Kopien von `src/` (Abschnitt 4); drei Wegwerf-Proben auf `HEAD` und auf den grünen Mutanten; `make gegenproben` einmal (Stempel danach mit `git checkout` zurückgesetzt); `git diff --numstat`; `make docs-check` und `make gates` nach dem Commit dieses Berichts (Abschnitt 12).

## Urteil

**Nicht freigegeben — vier MEDIUM, sechs LOW, fünf INFO.** Der Code ist richtig, soweit dieser Lauf ihn prüfen konnte: 19 von 23 gezählten Mutationen wurden aus dem benannten Grund rot, und `HEAD` besteht die drei Proben, die drei der vier grünen Mutanten rot machen. Die MEDIUM sind drei Zusagen, bei denen ein Mutant grün überlebt (V2 bis V4, Implementer), und ein Rest von F-1: drei Sätze nennen den Abbruch weiter ohne die Ausnahme für den Request-Body (V1, Architect). Die Zeittests sind in allen Läufen grün geblieben, auch bei 0,01 CPU. Die Größe 4391 gegen die Marke 3600 entscheidet der Planner (V10). DoD-Häkchen und §7 setzt der Planner.

## 1. DoD (§2) gegen Code, Spec und Tests

| DoD-Punkt | Befund |
|---|---|
| Tee, Queue, Rohgröße, Frist, `Abbrechen` (Nr. 2 bis 5, 9, 14) | Im Code erfüllt: `pipeline.go` (`Tee`, Prüfung der Argumente vor jedem Lesen), `aufnahme.go` (`Read` ruft `src.Read` einmal und legt danach ein, `Abbrechen` ruft nur die `CancelCauseFunc`), `queue.go` (Einlegen unter `mu` ohne Warten, Grenze an `ausstehend + e`, Rohgröße gekürzt bis `MaxBytes`). Die Fänger sind rot bei M04, M09, M20, M26 und M29. **Lücken:** V2 (Ursache außerhalb gegen Verstoß in der Queue, M38 grün), V3 (Entnahme in Teilen, M10 grün), V4 (`src` nie schließen, M64 grün). |
| Begrenzte Dekompression (Nr. 6, 7) | Im Code erfüllt: `dekodierung.go` (Kette, Stufen mit `bufio`, zlib statt raw-deflate, `uebrig`, `verhaeltnisOK` ohne Multiplikation). Fänger: `TestKette`, `TestKetteReihenfolge`, `TestEntpackt`, `TestNachlauf`, `TestBombe`, `TestVerhaeltnisAmAufruf`, `TestEntpacktGrenze`. **Lücke:** V5 (am Aufruf ist die Grenze nur in Schritten von 20 entpackten Bytes belegt, M12b grün). |
| Redaktor, Writer, Abschluss (Nr. 8, 10 bis 13, 15) | Erfüllt: `schreiber.go` (`Begin` beim ersten Byte der Ausgabe, `Abort` höchstens einmal, `Redigiert` nur angenommene Bytes), `aufnahme.go` (`schliesse`: Länge geprüft, nach gelungenem `Commit` kein Aufruf mehr), `abschluss.go`. Fänger rot bei M06, M07, M13, M16 bis M19, M34, M43 und M53. Die Fixtures aus `SPEC-053` stehen in `TestFixturesSPEC053`; der Crash zwischen Blob-Rename und DB-Commit liegt beim Dienst (Sendung (f)). |
| `make gates` grün | Selbst gefahren: rc 0 auf `d7d7737` und nach dem Commit dieses Berichts (Abschnitt 12). |
| Review durchgeführt, Report liegt vor | Erfüllt für `49063d4`. Die Commits danach (`5f69ae3`, `d7d7737`) prüft dieser Bericht mit Mutationen; ob ein Re-Review nötig ist, entscheidet der Planner. |
| Doku-Update (öffentlicher Vertrag) | Erfüllt: `SPEC-119`; Zeiger in `SPEC-046` Nr. 7, `SPEC-053`, `SPEC-118` Nr. 10, `LH-FA-02.b`, `LH-FA-02.c`, `LH-FA-13.a`. Rest von F-1 in V1. |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | Offen, Sache des Planners (Abschnitt 9). |

Die Importe des Pakets gehen nur an Modell und Outbound Ports (`make a-check` im Gate-Lauf grün, `TestKeineDateiUndKeinLog` mit Liste der erlaubten Importe). Kein `//nolint` im Paket.

## 2. `SPEC-119` unabhängig gelesen; F-1 über alle Stellen

**`incomplete` gegen `failed` beim Abbruch.** Jede Stelle, die den Abbruch eines Austauschs nennt:

| Stelle | Wortlaut zum Abbruch | Stimmt mit Nr. 12 |
|---|---|---|
| `SPEC-119` Nr. 12, Nr. 14 | Request nicht `Vollstaendig` ergibt `failed`, auch bei `Abbruch`; nicht zu Ende gelesener Request-Body: `Abbruch (Nr. 12: failed)` | Maßstab |
| `SPEC-119` Einleitung, erster Satz | `incomplete` ist der Abbruch nach vollständig redigiertem Request, `failed` die Verwerfung und der Abbruch im Request-Body | ja |
| `SPEC-119` Einleitung, letzter Satz | „der Unterschied ist nur, ob die Ursache der Austausch (Client, Upstream, Herunterfahren) oder die Aufzeichnung war“ | **nein**: ein Client-Abbruch im Request-Body hat den Austausch als Ursache und ist `failed` (V1) |
| `SPEC-053` (Spezifikation) | Abbruch setzt `incomplete`, „außer im Request-Body: … `failed`“ | ja |
| `LH-FA-02.b` (Spezifikation) | nach dem ersten gesendeten Byte `incomplete` („`failed`, wenn der Request-Body noch nicht beendet war“) | ja |
| `LH-FA-02.c` (Spezifikation) | nach Ablauf der Frist `incomplete` (`failed` bei offenem Request-Body) | ja |
| `LH-FA-13.a` (Spezifikation, Status) | beim Beenden noch laufende Exchanges `incomplete` (im Request-Body `failed`) | ja |
| `SPEC-001` (Body-Zustände) | `discarded` nur bei `failed` oder `incomplete`; `redaction.request=unsafe` verlangt `failed` | ja |
| `spec/architecture.md` Zeile 175 (`RecordUseCase`) | „`Drain`: … überfällige als `incomplete` schreiben“ | **nein** (V1) |
| `spec/architecture.md` Zeile 552 (Sequenz `close`) | „In-flight exchanges finished or aborted as incomplete“ | **nein** (V1) |
| Lastenheft [LH-FA-14](../../spec/lastenheft.md) | „der Austausch wird als unvollständig markiert“ (Beschreibung und Negative für Redaktions- und Speicherfehler) | Kein Wert genannt. `SPEC-119` liest „unvollständig“ als „nicht vollständig erfasst“ (`failed` oder `incomplete`). Wer wörtlich `incomplete` abnimmt, trifft `ablage` und `frist` nicht (I-8 unverändert, V12) |
| Lastenheft [LH-FA-02](../../spec/lastenheft.md) | Speicherfehler: „kein vollständig gespeicherter Austausch vorgetäuscht“ | ja (`failed`) |

Das Lastenheft ist nicht geändert. Der Code folgt Nr. 12 (`captureStatus` in `abschluss.go`; `TestAbschluss`, Zeilen mit Request `Abbruch`).

**Weitere Sätze, die einander widersprechen** (sonst fand ich keinen Widerspruch zwischen Nr. 1 bis 15, Einleitung und *Ungeprüft*):

* Nr. 15 sagt „Scheitert einer, … `Abort` läuft, der Grund ist `ablage` (oder `ausgabe_zu_gross`)“. Nr. 8 sagt `Abort` nur, „falls ein Writer offen ist“: Ein gescheitertes `Begin` hat keinen offenen Writer, also kein `Abort`. Code und Test folgen Nr. 8 (`TestAusgabeZuGross`: „`Begin` ohne Writer: 0“). Nach Nr. 9 ist ein Kontextfehler des Writers bei beendetem Kontext `frist` oder `abbruch`, nicht `ablage` (V6).
* Nr. 14 sagt, die Goroutine „endet in jedem Fall“. *Ungeprüft* nimmt den Writer-Aufruf aus, der nie zurückkehrt. Der Kommentar an `arbeite` sagt ohne diese Ausnahme „endet in jedem Fall“ (V7).

## 3. Zusage·Fänger-Tabelle (§3) und Raster (§6)

**Tabelle.** 50 Zeilen, wie der Implementer meldet. Jeder genannte Fänger existiert unter seinem Namen: 58 Funktionen `Test…` im Paket, von denen die Tabelle 57 nennt. Der Plan nennt zusätzlich `TestBodySpeicher`; es steht in `src/hexagon/model/redaktor_test.go` und ist dort als Träger genannt. Im Paket steht ein Test, den die Tabelle nicht nennt: `TestNachEndeVorDemErgebnis`. Er ist ein zusätzlicher Fänger zu Nr. 4 (V13). Keine Zelle ist leer.

Gegen den Vertrag gelesen hat jede Nummer von `SPEC-119` mindestens eine Zeile. Drei Zusagen haben eine Zeile, deren Eingaben die Mutation nicht trennen, oder gar keine Zeile:

* Nr. 9, „sie wirken sofort, ohne Rücksicht auf die Bytes, die noch in der Queue stehen“. Die Tabelle führt den Satz nur für den Überlauf (`TestUeberlaufVorVerstoss`), nicht für Frist, `Abbrechen` und Abbruch von `ctx` (V2).
* Nr. 4: `ausstehend` sinkt um jedes entnommene Byte. `TestKeinUeberlaufBeiSchnellemKonsument` liest in Stücken von 1000 Bytes; diese entnimmt der Leser immer ganz (V3).
* Nr. 3, „`src` schließt die Pipeline nie“: keine Zeile (V4).

**Raster.** 14 Zeilen mal 6 Spalten, dazu 39 Zusatzzeilen. Keine Zelle ist leer. Vier Zellen der Spalte *Wiederholung* tragen nur „(b)“: `New` · `MaxBytes`…, `Tee` · `Richtung`, `Tee` · `ContentType`, `Tee` · `Metadata`. Die Erklärung von (b) im Kopf nennt keine Nummer; der Ort wäre Nr. 14 (V8). Es fehlt die Zusatzzeile *Frist, `Abbrechen`, Abbruch von `ctx` × Verstoß in der Queue* (zu V2). Der Ort steht in Nr. 9, also ist das keine Lücke der Spezifikation.

## 4. Mutationen (selbst gefahren)

**Verfahren.** Für jede Mutation entsteht eine eigene Kopie des Stands `HEAD` (`git archive HEAD src`, im Scratchpad außerhalb des Repos). `perl` ersetzt eine Stelle; das Skript prüft, dass sich die Datei geändert hat. Danach läuft `go test -count=1 -timeout 240s` über das Paket im Image der Stufe `deps` des Dockerfiles (Digest-Pin des Go-Images, `--network none`, Quelle schreibgeschützt eingehängt). Der Arbeitsbaum des Repos blieb unberührt (`git status` leer). Die Mutationen übernehmen keine des Implementers oder des Reviewers, sondern sind eigene Varianten. Alle 24 bauten, keine hing. Gezählt wird nur, was aus dem benannten Grund rot wird. M12 wurde aus einem anderen Grund rot und ist durch M12b ersetzt. Gezählt sind damit 23: **19 rot, 4 grün.**

| # | Zusage | Mutation | Ergebnis |
|---|---|---|---|
| M01 | Nr. 9 letzter Punkt: ein Fehler des Writers bleibt, was er ist, auch bei danach beendetem Kontext | `befundWriter`: beendeter Kontext vor jeder Zuordnung | rot: `TestFehlerBleibtNachUrsache` (Writer im `Write` und im `Commit`, je `Abbrechen` und `cancel`), `TestAusgabeZuGross` (`Begin` und `Commit` bei beendetem Kontext) |
| M02 | Nr. 9: ein Verstoß des Redaktors bleibt | `befundRedaktor`: beendeter Kontext vor jeder Zuordnung | rot: `TestFehlerBleibtNachUrsache` (Verstoß des Redaktors), `TestAusgabeZuGross` |
| M03 | Nr. 9 und 7a: ein Nachlauf bleibt | Prüfung `ctx.Err()` vor `pruefeNachLesen` gezogen | rot: `TestFehlerBleibtNachUrsache` (Nachlauf × `Abbrechen`, × `cancel`) |
| M04 | Nr. 5: der Überlauf geht der Rohgröße vor | Überlauf nur geprüft, wenn der Read die Rohgröße nicht überschreitet | rot: `TestUeberlaufGegenRohgroesse` (zwei Fälle) |
| M06 | Nr. 8: `Begin` beim ersten Byte der **Ausgabe** | `Begin` beim ersten Byte am Eingang des Redaktors | rot: `TestWriterErstBeimErstenByte` (`typ`), `TestAbortNieCommit` (`typ`, `mismatch`), `TestAusgabeZuGross` (`Begin`) |
| M07 | Nr. 8: scheitert `Commit`, folgt `Abort` | kein `Abort` nach gescheitertem `Commit` | rot: `TestAbortNieCommit` (`ablage` am `Commit`), `TestWriterFehler`, `TestAusgabeZuGross`, `TestFrist` (`Commit` wartet) |
| M09 | Nr. 4: `ausstehend + e` gegen `QueueBytes` | nur `e` gegen `QueueBytes` | rot: `TestQueueGrenze` (60 und 41, leere Reads und 101), `TestUeberlaufGegenRohgroesse`, `TestQueueHoechststand`, `TestReadWartetNichtAufSchreiber` |
| M10 | Nr. 4: `ausstehend` sinkt um jedes entnommene Byte | eine Teil-Entnahme senkt `ausstehend` nicht | **grün** (V3); Probe rot |
| M12 | Nr. 7b am Aufruf | `verhaeltnisOK(roh, entpackt-1, …)` | rot nur bei leeren Bodys (`entpackt` wird −1): nicht der benannte Grund, nicht gezählt |
| M12b | Nr. 7b am Aufruf, ein entpacktes Byte über 20:1 | `verhaeltnisOK(roh, max(entpackt-1, 0), …)` | **grün** (V5); Probe rot |
| M13 | Nr. 13 über mehrere `Write` | `Redigiert` hält nur den ersten `Write` | rot: `TestRedigiertUndKey` (Request über 32 KiB) |
| M16 | Nr. 12: Request `Vollstaendig` × Response `Abbruch` ist `incomplete` | `failed` | rot: `TestAbschluss` (`AX`, `NX`, `PX`) |
| M17 | Nr. 12: `verified`, wenn mindestens ein Body `present` ist | beide `present` verlangt | rot: `TestAbschluss` (acht Zeilen) |
| M18 | Nr. 12: `none` bei verworfenem Request-Body | nur der Response-Body zählt | rot: `TestAbschluss` (sechs Zeilen `X*`, `V*`) |
| M19 | Nr. 12: `KeyMoeglich` heißt Request `Vollstaendig`, auch `absent` | zusätzlich `present` verlangt | rot: `TestAbschluss` (`A*`), `TestAbschlussGueltig` |
| M20 | Nr. 9: `encoding` geht dem Abbruch von `ctx` vor `Tee` vor | Kontext vor dem in `Tee` entschiedenen Ergebnis geprüft | rot: `TestTeeCtxBeendet` |
| M26 | Nr. 4: nach dem Ergebnis kopiert `Read` nichts | `q.schliesse()` nach dem Ergebnis entfernt | rot: `TestNachEndeKeineKopie` (nach Verstoß) |
| M29 | Nr. 8, 9: `Begin` mit dem Kontext der Aufnahme | `Begin` mit dem `ctx` des Aufrufers | rot: `TestFrist` (c bis e, an der Zeitgrenze), `TestErsteUrsacheGewinnt` (d), `TestWriterErstBeimErstenByte`, `TestWriterFehler`, `TestGoroutineEndetImmer` |
| M34 | Nr. 8: Länge des `Commit` gleich den geschriebenen Bytes | nur `<` geprüft | rot: `TestCommitLaenge` (Länge +1) |
| M38 | Nr. 9: Frist, `Abbrechen` und `ctx` wirken sofort, Bytes in der Queue verfallen; Kommentar an `queueLeser` | `queueLeser` entnimmt vorhandene Bytes, bevor er den Kontext prüft | **grün** (V2); Probe rot |
| M43 | Nr. 12: Richtung beider Ergebnisse geprüft | nur die des Requests | rot: `TestAbschlussArgumente` (Nullwert als Response, beide `Request`) |
| M53 | Nr. 8: `Commit` erst nach Nr. 7 und ohne Ursache der Nr. 9 | `Commit` vor den Prüfungen | rot: `TestAbortNieCommit` (`bombe`, Nachlauf: `Commit 1-mal`), `TestCommitUndAbbruch`, `TestWriterEinerGoroutine`, `TestFehlerBleibtNachUrsache`, `TestAusgabeZuGross`, `TestWriterErstBeimErstenByte` |
| M63 | Nr. 9: `context.DeadlineExceeded` des Writers bei beendetem Kontext ist `frist` | nur `context.Canceled` erkannt | rot: `TestFrist` (c bis e), `TestErsteUrsacheGewinnt` (d), `TestWriterFehler` (Frist, drei Aufrufe) |
| M64 | Nr. 3: `src` schließt die Pipeline nie | die Goroutine schließt `src`, wenn es `io.Closer` ist | **grün** (V4) |

Schwerpunkte des Auftrags: **Rangfolge Nr. 9**: M01, M02, M03, M04, M20 und M63 rot, M38 grün. **Writer-Lebenszyklus**: M06, M07, M29, M34 und M53 rot. **Queue-Grenze, beide Seiten**: M04 und M09 rot, M10 grün. **Verhältnis am Aufruf**: M12b grün. **`Redigiert` über mehrere Writes**: M13 rot. **Abschluss**: M16 bis M19 und M43 rot.

**Proben** (Wegwerf-Test, nicht committet). Er lief in einer Kopie von `HEAD` und in den drei Mutanten M10, M12b und M38:

* *Teil-Entnahme.* Ein Body von 1 MiB (`identity`) kommt in Reads zu 40000 Bytes, `QueueBytes` ist 40000. Nach jedem Read wartet die Probe auf `Ausstehend() == 0`; der Redaktor liest in Stücken zu 32 KiB. `HEAD`: grün. M10: rot an der Zeitgrenze, weil `Ausstehend` nicht auf 0 fällt.
* *Abbrechen vor Verstoß in der Queue.* Die Goroutine ist am Start angehalten, `{"a":x` steht samt `io.EOF` in der Queue, dann `Abbrechen()` und Freigabe. `HEAD`: `abbruch`. M38: `syntax`/`Verstoss`.
* *Ein Byte über 20:1 am Aufruf.* 2001 entpackte Bytes bei 100 Rohbytes (`gzMitRoh`). `HEAD`: `bombe`. M12b: `Vollstaendig`.

## 5. Zeitabhängigkeit (F-13 und `TestFrist`)

Alle Läufe liefen im Image der Stufe `deps`, netzlos.

| Lauf | Bedingung | Ergebnis |
|---|---|---|
| A | `-count=50`, `-run 'TestErsteUrsacheGewinnt\|TestFrist$'` | grün, 21,2 s |
| B | ganzes Paket `-count=50` | grün, 107 s |
| C | wie A, `--cpus=0.1` | grün, 21,3 s |
| D | wie A, `--cpus=0.05` | grün, 21,3 s |
| E | wie A, `--cpus=0.05`, im selben Container acht Busy-Schleifen | grün, 38,9 s |
| F | ganzes Paket `-count=1`, `--cpus=0.05` | grün, 18,7 s |
| G | wie E bei `--cpus=0.01` | nach rund 13 Minuten ohne Ausgabe abgebrochen (das Bauen des Testbinarys kam unter dieser Last nicht zum Ende); nicht gezählt, ersetzt durch G2 |
| G2 | vorgebautes Testbinary, `--cpus=0.01`, acht Busy-Schleifen, `-test.count=20`, wie A | grün, 91 s |

Rot ohne Codefehler trat in keinem Lauf auf. Gelesen habe ich das so: `TestFrist` hat kein Zeitfenster. Kommt die Frist vor dem Writer, ist das Ergebnis ebenfalls `frist`; späte Pausen färben nur den Fang der Mutation „Writer-Fehler als `ablage`“ in dem einen Lauf. `TestErsteUrsacheGewinnt` (a) bis (c) hat kein Fenster. Fall (d) behält das Fenster, das der Plan in §3 benennt: Erreicht die Goroutine den Writer nicht innerhalb von 250 ms, meldet der Fake kein Ende des Kontexts, und der Test endet nach 10 s rot. Das ist dokumentiert und wurde in keinem Lauf beobachtet (V14).

## 6. Review-Findings F-1 bis F-19, I-1 bis I-8

| Finding | Behandlung | Beleg |
|---|---|---|
| F-1 MEDIUM `incomplete`/`failed` | `5f69ae3`: `SPEC-053`, `LH-FA-02.b`, `LH-FA-02.c`, `LH-FA-13.a` und der erste Satz der Einleitung nachgezogen | Abschnitt 2; Rest in V1 |
| F-2 MEDIUM zwei Zellen im Code entschieden | Nr. 9 und 10 entscheiden jetzt beide; Code unverändert | M01 (`ErrZuGross` aus `Begin`/`Commit` bei beendetem Kontext), M63 rot |
| F-3 MEDIUM Folge des Panics | *Ungeprüft* nennt die Folge „Ende des Prozesses“, den Grund und die Ports als panikfrei | Abschnitt 11; Zusatz in V15 |
| F-4 MEDIUM `Redigiert` über mehrere Writes | Fall „Request über 32 KiB hat mehrere Writes“ (mindestens drei `Write`) | M13 rot |
| F-5 MEDIUM Verhältnis am Aufruf | `TestVerhaeltnisAmAufruf` mit 100, 99 und 101 Rohbytes | behandelt; Rest in V5 (M12b) |
| F-6 MEDIUM drei Paare | (a) `TestVerstossVorVerhaeltnis`, (b) gleicher Read in `TestVerstossVorFehler`, (c) `TestRohgroesseUndNachlauf` und Rasterzeile | in Tabelle und Raster |
| F-7 MEDIUM Fehler bleibt nach Ursache | `TestFehlerBleibtNachUrsache` | M01, M02, M03 rot |
| F-8 MEDIUM Kontextfehler an `Begin`/`Commit` | `TestFrist` (d), (e); `TestWriterFehler` (je Aufruf) | M07, M63 rot |
| F-9 MEDIUM mehrere Werte `Content-Type` | `TestRedaktorAnbindung`: beide Reihenfolgen | Fall vorhanden (`writer_test.go` Zeile 53) |
| F-10 MEDIUM Wasserstand je Aufnahme | `TestAufnahmenUnabhaengig` (b) und (c), zwei Pipelines | Fall vorhanden |
| F-11 MEDIUM Datei und Log | `TestKeineDateiUndKeinLog`: Liste erlaubter Importe, Aufrufe `print`, `println`, `panic`, `fmt.Print*`, `fmt.Fprint*`, Gegenprobe über acht Quelltexte | Fall vorhanden |
| F-12 MEDIUM Goroutine endet | `TestGoroutineEndetImmer` über alle Zeilen von `captureFaelle` mit Kontext-Prüfung | M29 rot dort (Zeile `gut`) |
| F-13 MEDIUM Zeitfenster | (c) ohne Fenster, (d) über den Kanal des Fakes | Abschnitt 5 |
| F-14 LOW Kette | `gzip, gzip`, `identity, identity`, Rand mit `\n`, `\r`, `\v`, `\f` und U+00A0, U+0130, `gzip,,identity` | Fälle vorhanden |
| F-15 LOW Fehlertext | `verworfen()` vergleicht `Error()` mit dem ganzen Text | `helfer_test.go` Zeile 555 |
| F-16 LOW zwei Zellen | `TestReadLaengeNull`, `TestAbschlussArgumente` | M43 rot |
| F-17 LOW Plan-Zahlen | §3, §4 und §6 nennen 50 Zeilen; §4 nennt 44 nur als Stand beim Setzen der Marke | gelesen |
| F-18 LOW `var _ = io.EOF` | entfernt | `grep` leer |
| F-19 LOW Kommentar `befundRedaktor` | Kommentar sagt jetzt „wird als ablage gemeldet; kein Test fängt diesen Zweig“ | `aufnahme.go` Zeile 167 bis 169 |
| I-1 | keine Aktion | — |
| I-2 Größe | offen beim Planner | V10 |
| I-3 Test-Haken | `queue.hoechst`, `queue.zustand()`, `Aufnahme.entpackt` tragen Kommentare „nur die Tests …“, `Pipeline.start` „nil in jedem Lauf außer im Test“; im Plan §3 genannt | `d7d7737` ändert im Produktcode nur Kommentare (Diff gelesen) |
| I-4, I-5 | Plan-Sätze entfernt | `grep` leer |
| I-6 `-race` | §8 dieses Plans berichtigt; die sieben Pläne unter `open/` sind nicht Gegenstand | — |
| I-7 | keine Aktion | — |
| I-8 Lastenheft-Wort | unverändert, Risiko in §6 mit Ausgang *entfallen* | V12 |

## 7. Plan gegen Code (§3.9), Kommentare (§3.7, §3.11)

Die Dateien in §3 des Plans entsprechen dem Diff: sieben Produktdateien, neun Testdateien, `spec/spezifikation.md`. Typen und Funktionen tragen die Namen aus `SPEC-119` (`Konfig`, `Standard`, `New`, `Tee`, `Eingang`, `Aufnahme.Read`/`Abbrechen`/`Ergebnis`, `Abschluss`, `Zustand` mit den zehn Feldern aus Nr. 12). Der Kopf (`Bezug`, `Berührte Spec-Stellen`) nennt `LH-FA-13.a`, §1 die Regel „Abbruch im Request immer `failed`“, §6 die Ausgänge zu F-1 bis F-3. Den Nachzug zu den Befunden dieses Berichts trägt der Plan noch nicht (V2 Rasterzeile, V8).

Kommentare: Die Test-Haken sind ehrlich benannt (I-3). Zwei Kommentare sagen mehr zu, als ein Test fängt (§3.11):

* `queueLeser`: „kehrt sofort zurück, sobald der Kontext der Aufnahme beendet ist“. M38 bleibt grün (V2).
* `arbeite`: „endet in jedem Fall“. Nicht beim hängenden Writer (V7).

Sonst tragen die Kommentare Zusage, Kopplung oder Abgrenzung im Indikativ.

## 8. Übergaben (§3.13)

| Nehmer | Sendung | Befund |
|---|---|---|
| `slice-record-service` | (a) `Tee` je Body, nie für Antworten ohne Body, `Metadata` nur für `$metadata`; (b) `Ergebnis`, `Abschluss`, Key nur bei `KeyMoeglich` aus `Redigiert`; (c) `Abbrechen` für jeden nicht zu Ende gelesenen Strom; (d) Protokoll mit `ODR-E2401` ohne Payload; (e) `Accept-Encoding` und Austausch ohne Antwort-Strom; (f) Fault-Injection-Punkt zwischen `Commit` und Anhängen | §1 *Übernimmt (Capture-Pipeline)* trägt (a) bis (f) mit der Kennung `slice-capture-pipeline`. Kein Ausschluss trifft eine Sendung: HTTP-Server, Session-Use-Case, Link-Umschreiben und Fallback; `Drain` gehört dem Nehmer. Der Nehmer liegt in `open/`. In (b) heißt es „`RequestRedaction` des Zustands“; das Feld von `Zustand` heißt `RequestRedact` (`RequestRedaction` ist das Feld der Eingabe von `BildeSchluessel`) (V13) |
| `slice-replay-service` | nichts | richtig: der Replay ruft `RedigiereBody` selbst |

## 9. Größe (§4) und Closure-Pflichten

Nachgemessen mit `git diff --numstat e085e94..d7d7737` über `src/hexagon/services/record/capture/` und `spec/spezifikation.md`, Tests eingeschlossen: **4391** (Produktcode 871, Tests 3493 einschließlich `export_test.go`, Spezifikation 27 = 14 + 13). Das ist der Wert des Implementers.

Die Marke 3600 ist um **791** überschritten (22 %). Die Erwartung des Plans nach dem Review (3750 bei 50 Zeilen) ist um 641 überschritten. Beim Review standen 3771. Seitdem kamen 620 hinzu: +3 Produktcode (Kommentare), +601 Tests, +16 Spezifikation. Ob die Marke angehoben wird oder der Slice nach `next` zurückgeht, entscheidet der Planner; die Korrekturen zu V2 bis V5 kommen hinzu (V10).

Für die Closure bleibt dem Planner:

* DoD-Häkchen und §7 mit Steering-Loop-Eintrag. Kandidaten aus diesem Lauf sind die Klassen „Paar von Verstößen ohne Fänger“ (V2), „Fänger trennt die Mutation nicht“ (V3, V5), „Zusage ohne Fänger“ (V4) und „normative Sätze widersprechen sich“ (V1, V6), dazu die Klassen des Reviews.
* Ausgänge in §6: *Raster … offen bis zum Review* steht nach dem Review noch so (V9). *Größe* und *Übergaben* sind offen.
* Beobachtungs-Register: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` steht bei 2×; dieser Lauf sah keinen roten Zeitlauf.

## 10. Gegenproben-Stempel

`harness/gegenproben.stempel` ist **veraltet**. Er trägt `1e3793dc…` (Stand `49063d4`). `src` ist Eingabe der Prüfsumme (`tools/harness/gegenproben-eingabe.sh`), und `d7d7737` änderte acht Dateien unter `src/` (Tests und Kommentare). Die Prüfsumme auf `d7d7737` ist `41c73f3a89405b6014241fa1321eb7d5c07229b12b46a48fdac576ba52c26279`. `make gegenproben` ([ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), kein Gate) lief einmal neu: 33 Fälle, 0 rot, 78 s. Der Stempel ist danach mit `git checkout` zurückgesetzt und nicht committet (V11).

## 11. Akzeptierte Negative

| Negativ | Benannt mit Grund und Folge |
|---|---|
| Panic eines Ports oder des Redaktors | ja: *Ungeprüft* und §3 „Nicht gefangen“. Grund: nicht abgefangen, damit nichts verborgen wird, und der Writer ist in unbekanntem Zustand. Folge: Ende des Prozesses und aller Sessions, weil die Goroutine der Aufnahme kein Handler ist; der Neustart setzt Sessions auf `incomplete`, verwaiste Blobs sind zulässig. Für `src` trägt die Begründung nicht ganz (V15) |
| Typed-nil als `Blobs` oder `src` | ja, im selben Satz |
| Hängender Writer | ja: „die Pipeline kann ihn nicht unterbrechen, `Ergebnis()` wartet dann mit“; der Kommentar an `arbeite` sagt es nicht (V7) |
| Datenrennen | ja: das Gate fährt kein `-race` (`CGO_ENABLED=0`); die Tests ordnen mit Kanälen und laufen mit `-count=50` (hier Lauf B grün) |
| Crash zwischen Blob-Rename und DB-Commit | ja: beim Dienst, Sendung (f), im Nehmer eingetragen |
| Speicher über alle gleichzeitigen Aufnahmen | ja: die Queue gilt je Aufnahme, keine Grenze für die Zahl der Aufnahmen in dieser Version, kein Nehmer. Die Folge, Speicher bis zur Zahl der Aufnahmen mal `QueueBytes`, ergibt sich daraus, steht aber nicht als Zahl da |
| Fehler des Redaktors ohne `*model.UnsicherFehler` | ja: *Ungeprüft*, Kommentar an `befundRedaktor` |

## 12. Gate-Lauf

`make gates` auf `d7d7737` vor diesem Bericht: rc 0. Die Stufen `test`, `lint` und `build` kamen aus dem Docker-Cache, deren Schicht nur nach einem grünen `RUN` entsteht; das ungecachte `go test -count=1` dieses Laufs ist grün. `make docs-check` und `make gates` nach dem Commit dieses Berichts: Ergebnis im Handback an den Planner.

## Befunde

| ID | Kategorie | Befund | Zuständig |
|---|---|---|---|
| V1 | MEDIUM | F-1 ist nicht an allen Stellen nachgezogen. (a) `SPEC-119` Einleitung, letzter Satz: „der Unterschied ist nur, ob die Ursache der Austausch … oder die Aufzeichnung war“. Ein Client-Abbruch im Request-Body hat den Austausch als Ursache und ist nach Nr. 12 `failed`. (b) `spec/architecture.md` Zeile 175, `RecordUseCase`: „`Drain`: … überfällige als `incomplete` schreiben“. (c) Zeile 552, Sequenz `close`: „In-flight exchanges finished or aborted as incomplete“. (b) und (c) beschreiben genau den Ablauf, den `slice-record-service` baut (`Drain`, Sendung (c)); ein überfälliger Exchange mit offenem Request-Body ist nach `SPEC-119` Nr. 12 und `LH-FA-13.a` `failed`. | Architect |
| V2 | MEDIUM | Nr. 9: Frist, `Abbrechen` und Abbruch von `ctx` „wirken sofort, ohne Rücksicht auf die Bytes, die noch in der Queue stehen (sie werden verworfen)“; ebenso der Kommentar an `queueLeser`. Kein Fänger: M38 bleibt grün, die Probe trennt ihn (`HEAD` `abbruch`, M38 `syntax`). Gemutet wird ein Client-Abbruch während der Antwort zu `Verstoss`/`failed` statt `Abbruch`/`incomplete`, wenn in der Queue noch ein Verstoß steht. Die Tabelle führt den Satz nur für den Überlauf, dem Raster fehlt die Zusatzzeile *Frist, `Abbrechen`, `ctx` × Verstoß in der Queue* (`AGENTS.md` §3.15, §3.16: Paar von Verstößen ohne Fänger). | Implementer (Fänger, Tabellen- und Rasterzeile) |
| V3 | MEDIUM | Nr. 4: `ausstehend` ist die Zahl der eingelegten und noch nicht entnommenen Rohbytes, „ein schneller Konsument läuft nie über“. `TestKeinUeberlaufBeiSchnellemKonsument` liest in Stücken zu 1000 Bytes, die der Leser ganz entnimmt. M10 (Teil-Entnahme senkt `ausstehend` nicht) bleibt grün, die Probe mit Reads über 32 KiB trennt ihn. Gemutet laufen große Bodys mit großen Reads über; bei `gzip` liest die Stufe immer in 4096er-Stücken. Kein Geheimnis betroffen (Fänger trennt die Mutation nicht). | Implementer |
| V4 | MEDIUM | Nr. 3: „`src` schließt die Pipeline nie“. Die Tabelle hat keine Zeile, M64 (die Goroutine schließt `src`, wenn es `io.Closer` ist) bleibt grün; kein `src` der Tests hat `Close`. Gemutet schlösse ein früher Verstoß den Request-Body, den der Proxy noch liest. Im heutigen Code gibt es keinen solchen Weg (Zusage ohne Fänger, §3.15). | Implementer |
| V5 | LOW | Nr. 7b am Aufruf: „ein Byte mehr `bombe`“ ist an der reinen Funktion belegt, am Aufruf nur in Schritten von 20 entpackten Bytes (99, 100 und 101 Rohbytes bei 2000 entpackten). M12b (`entpackt` um eins zu klein gezählt) bleibt grün, die Probe mit 2001 entpackten Bytes bei 100 Rohbytes trennt ihn. | Implementer |
| V6 | LOW | `SPEC-119` Nr. 15 sagt „Scheitert einer, … `Abort` läuft, der Grund ist `ablage` (oder `ausgabe_zu_gross`)“. Nach Nr. 8 gibt es bei gescheitertem `Begin` kein `Abort` (kein Writer offen). Nach Nr. 9 ist ein Kontextfehler bei beendetem Kontext `frist` oder `abbruch`. Code und Tests folgen Nr. 8 und 9. | Architect |
| V7 | LOW | Nr. 14 „endet in jedem Fall“ und der Kommentar an `arbeite` („endet in jedem Fall“) nehmen den Writer-Aufruf nicht aus, der nie zurückkehrt; nur *Ungeprüft* tut es (`AGENTS.md` §3.11). | Implementer (Kommentar), Architect (Nr. 14, falls nachgezogen) |
| V8 | LOW | Vier Zellen *Wiederholung* des Rasters tragen nur „(b)“ ohne Ort: `New` · `MaxBytes`…, `Tee` · `Richtung`, `Tee` · `ContentType`, `Tee` · `Metadata`. Der Kopf erklärt (b) ohne Nummer; der Ort wäre Nr. 14 (`AGENTS.md` §3.16). | Implementer (Plan-Nachzug) |
| V9 | LOW | §6, Risiko *Raster*: Der Ausgang steht nach dem Review noch auf „offen bis zum Review“. | Planner (Closure) |
| V10 | INFO | Größe 4391 gegen die Marke 3600 (Abschnitt 9); keine Entscheidung des Verifiers. | Planner |
| V11 | LOW | `harness/gegenproben.stempel` ist nach `d7d7737` veraltet (Abschnitt 10); `make gegenproben` ist grün, der Stempel ist nicht nachgezogen. | Implementer: `make gegenproben`, Stempel committen |
| V12 | INFO | Lastenheft [LH-FA-14](../../spec/lastenheft.md) „als unvollständig markiert“ gegen `failed` für Redaktions-, Speicher- und Fristfehler (I-8): Das Lastenheft nennt keinen Wert, und `SPEC-119` legt das Wort aus. Wer bei der Abnahme wörtlich `incomplete` prüft, trifft `ablage` und `frist` nicht. Das Lastenheft ist nicht geändert. | Planner (Frage an den Auftraggeber, falls nötig) |
| V13 | INFO | Zwei Namen: `TestNachEndeVorDemErgebnis` fängt Nr. 4, steht aber in keiner Tabellenzeile. Die Sendung (b) im Nehmer sagt „`RequestRedaction` des Zustands“, das Feld heißt `RequestRedact`. | Implementer (Tabelle); Planner (Wortlaut der Sendung) |
| V14 | INFO | `TestErsteUrsacheGewinnt` (d) behält das benannte Fenster: Erreicht die Goroutine den Writer nicht in 250 ms, endet der Test nach 10 s rot. Nicht beobachtet, auch nicht bei `--cpus=0.01` mit acht Busy-Schleifen (Abschnitt 5, Lauf G2). | — (dokumentiert) |
| V15 | INFO | *Ungeprüft* begründet „Ende des Prozesses“ damit, dass Redaktor, Dekompressor und Writer in der Goroutine der Aufnahme laufen. `src.Read` läuft dagegen in der Goroutine des Aufrufers (`Aufnahme.Read`). Ein Typed-nil als `src` endet dort, wo der Aufrufer liest: beim Response-Body im Handler (`net/http` fängt ihn), beim Request-Body in der Goroutine des Upstream-Clients (nicht gefangen). Die Folge ist richtig als schlimmster Fall, die Begründung deckt `src` nicht. | Architect (optional) |

## Zuständigkeiten

* **Architect:** V1 (drei Sätze zu `incomplete`/`failed`), V6 (Nr. 15), V7 (Nr. 14, falls nachgezogen), V15 (optional).
* **Implementer:** V2, V3, V4 (Fänger samt Tabellen- und Rasterzeilen, Mutation selbst rot sehen), V5, V7 (Kommentar), V8, V11 (Stempel), V13 (Tabelle). Im Produktcode ist nach diesem Lauf nichts zu ändern; HEAD besteht alle drei Proben.
* **Planner:** V9, V10 (Anheben oder Zerlegen), V12, V13 (Sendung), Closure-Pflichten aus Abschnitt 9, Entscheidung über ein Re-Review der Commits nach `49063d4`.

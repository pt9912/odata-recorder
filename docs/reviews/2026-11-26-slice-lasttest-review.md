# Review-Report: slice-lasttest (Teil 3, Messlauf `make lasttest`) — 2026-11-26

**Review-Art:** Code — gegen Plan, Vertrag, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) und die Hard Rules. Keine Prüfung gegen die DoD.

**Gegenstand:** Zweig slice-lasttest, HEAD `75fffee6`, `git diff 95779768..HEAD` über die Pfade des Blocks der Größenmarke. Die Änderungen unter `src/` und `tests/integration/` stammen aus `main` und sind nicht Gegenstand.

**Skill:** `.harness/skills/reviewer.md` @ `75fffee6` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-26

**Eingangs-Kontext:** Plan slice-lasttest (§1, §3, §4, §6), `SPEC-216` vollständig, `SPEC-128` (Formel, Messung), `SPEC-102` (Fälle 217 bis 225), `LH-QA-01`, `AGENTS.md` §3, Report zu slice-lasttest-auswertung (wiederkehrende Klassen).

## Verdikt

**Nicht freigegeben: 0 HIGH, 6 MEDIUM, 1 LOW, 4 INFO.**
Am Verhalten des Codes habe ich bei Takt, Messfenster, Fehlerzählung, Gültigkeit, Urteil, Exit, Abnahme und Treiber keinen Fehler gefunden, außer der Abweichung in F-6.
F-1 bis F-5 sind fehlende Fänger an der Stelle, an der der Ablauf die Funktionen der Teile 1 und 2 ruft: jede dieser Verdrahtungen lässt sich brechen, und das Paket bleibt grün, mehrfach mit der Folge „ungültiger oder falscher Lauf endet als bestanden“.

## Findings

| ID | Kategorie | Befund und Versagen | Quelle | Pfad (Anker) | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Lage des Messfensters am Aufruf ist nicht gefangen: Warm-up und Messphase vertauscht bleibt grün, weil `kurzesProfil` beide auf 1 s setzt. Bei `abnahme` rechnete p95 dann über die letzten 1200 statt über 12000 Spans, und der Lauf bliebe `gueltig=ja`, weil die Zählprüfung nur die Summe sieht. | `AGENTS.md` §3.15, `SPEC-216` Nr. 8 a, 10 e, 10 f | `tests/lasttest/ablauf.go` · `s.auswerte(je*int(e.Profil.Warmup/time.Second)`; `tests/lasttest/ersatz_test.go` · `p.Warmup, p.Messphase = name, laeufe, time.Second, time.Second` | ja, `make test` bleibt grün | Fixture gleicht dem Wert der Mutation |
| F-2 | MEDIUM | Dass alle Läufe einer Betriebsart, und nur ihre, ins Urteil gehen, hat keinen Fänger: ungültige Läufe weggelassen, nur der letzte Lauf, nur der erste Lauf und Läufe über beide Betriebsarten gesammelt bleiben grün. Die Folge wäre `bestanden`, Exit 0 und `abnahme=ja` trotz eines ungültigen Laufs oder einer Streuung über der Toleranz; beim Sammeln entfiele für `mock` die Prüfung der Streuung (`len(laeufe) == 2`). Kein Test des Ablaufs hat zwei Läufe, die sich in Gültigkeit oder p95 unterscheiden. | `AGENTS.md` §3.15, `SPEC-216` Nr. 12 c | `tests/lasttest/ablauf.go` · `laeufe = append(laeufe, l.messung)` | ja, `make test` bleibt grün | Aggregation am Aufruf ohne Fänger |
| F-3 | MEDIUM | Der Zielwert ist in einer Richtung nicht gefangen: das Urteil gegen die Konstante `abnahmewertMs` statt gegen `LASTTEST_ZIEL_MS` und `ziel_ms` der Zeile `Abnahme` fest auf 100 bleiben grün. Bei `LASTTEST_ZIEL_MS=80` und p95 90 ms stünde `bestanden` mit Exit 0, und die Zeile `Abnahme` nennte einen anderen Zielwert als der Aufruf. | `AGENTS.md` §3.15, `SPEC-216` Nr. 12 a, 13 | `tests/lasttest/ablauf.go` · `urteilBetriebsart(string(b), laeufe, e.ZielMs)`, `abnahmeZeile{ZielMs: e.ZielMs,` | ja, `make test` bleibt grün | Fixture gleicht dem Wert der Mutation |
| F-4 | MEDIUM | Von den Verstößen, die der Ablauf aus `auswerte` übernimmt, hat keiner einen Fänger am Ablauf: `v.Zaehlung` ohne den Anteil von `auswerte` und `v.Modus` verworfen bleiben grün. Ein Lauf mit mehr Gesamtspans als Anfragen (Nr. 10 e „zu viele“) oder ein Binary im falschen Modus wäre gültig. Die Zeile L12 führt allein `Fehler`, `Last` und `Binary`. | `AGENTS.md` §3.15, §3.17, `SPEC-216` Nr. 10 d, 10 e, 10 h | `tests/lasttest/ablauf.go` · `v.Zaehlung = v.Zaehlung \|\| !st.nachlauf \|\| lesefehler` | ja, `make test` bleibt grün | Klasse nur an einem Teil ihrer Elemente geführt |
| F-5 | MEDIUM | Das Paar Abbruch und Gesamtfrist hat keinen Fall: die zwei Zweige in `unterbrochen` vertauscht bleiben grün. Kommt das Signal nach dem Ende der Gesamtfrist und vor der Prüfung (etwa während `beendeBinaer`, bis 15 s), endete der Aufruf mit 3 statt mit 2 gegen den Vorrang „2 vor 3“. | Reviewer-Skill (Paar von Ursachen ohne Fänger), `AGENTS.md` §3.16, `SPEC-216` Nr. 12 e | `tests/lasttest/ablauf.go` · `func (e ablaufEingabe) unterbrochen(` | ja, `make test` bleibt grün | Paar von Ursachen ohne Fänger |
| F-6 | MEDIUM | Mit beendetem Kontext (Abbruch oder Gesamtfrist) entfällt die Meldung zum Binary, auch für ein Binary, das zuvor selbst endete; `SPEC-216` Nr. 9 c sagt „Ein Binary, das selbst endet, meldet …“ ohne Ausnahme, und weder Nr. 4 c, 4 d noch das Raster in §6 nennen eine. Endet das Binary im Lauf (etwa Exit 137) und folgt ein Abbruch, steht allein `abgebrochen`, und der Grund ist verloren. Die Entscheidung steht allein im Kommentar von `lauf`; ihr Entfernen bleibt grün. | `AGENTS.md` §3.12, §3.16 | `tests/lasttest/ablauf.go` · `Die Meldungen zum Binary` und `case ctx.Err() != nil:` | ja, `make test` bleibt grün | Randform im Code entschieden |
| F-7 | LOW | Die Ausnahme `d-check:ignore` trägt die Begründung „geplante Datei, entsteht mit dem Code“, die Dateien bestehen mit diesem Slice; die Unterdrückung bleibt ohne geltenden Grund. | `AGENTS.md` §3.2, §3.9 | `harness/vertraege/SPEC-216-lasttest.md` · Kopfzeile `**Werkzeug:**`; Plan §3, Zeile `Dockerfile`, `harness/mk/lasttest.mk` | nein, `make docs-check` meldet eine veraltete Begründung nicht | Veraltete Begründung einer Unterdrückung |
| I-1 | INFO | `verspaetung_max_ms` geht in keine Gültigkeit ein, und nach einem Stillstand startet der Takt alle überfälligen Anfragen unmittelbar nacheinander. Ein Lauf, der die 20 Requests/s nicht gehalten hat, kann `gueltig=ja` und `abnahme=ja` tragen; der Vertrag verlangt nichts anderes (Nr. 10 h), Nr. 8 b „wird nicht aufgeholt“ liest sich neben Nr. 8 a zweideutig. Adresse: Architect (Nr. 16 g). | `LH-QA-01`, `SPEC-216` Nr. 8 a, 8 b | `tests/lasttest/takt.go` · `faellig := t0.Add(` | nein | Setzung ohne Grenze |
| I-2 | INFO | Dass die Verspätung beim Start der Goroutine und nicht in der Schleife des Takts gemessen wird, fängt kein Test; die gesetzte Uhr unterscheidet die zwei Stellen nicht. | `SPEC-216` Nr. 8 b | `tests/lasttest/takt.go` · `spaet := uhr.Jetzt().Sub(faellig)` | ja | Messpunkt ohne Fänger |
| I-3 | INFO | stderr des Binary wird allein ausgewertet; scheitert das Binary, sieht der Aufrufer den Exit, nicht dessen Textzeile. Vertragsgemäß (Nr. 10 b). | `SPEC-216` Nr. 9 c, 10 b | `tests/lasttest/prozess.go` · `func entleere(` | nein | Diagnose |
| I-4 | INFO | Zum akzeptierten Negativ (11) des Plans: mit dem echten `docker`-Client erreichte TERM am Treiber und INT an die Prozessgruppe das Programm, je ein Handlauf, siehe unten. Für den Verifier. | `SPEC-216` Nr. 3 e, 4 d | `tools/harness/lasttest.sh` · `kill -TERM "$pid"` | ja, Handlauf | Beleg |

## Eigene Läufe

Messender: Reviewer, Host mit 20 Kernen, jeder schwere Lauf unter `flock /tmp/odata-recorder-heavy.lock`, ein Container zur Zeit.

**Mutationen:** Wegwerfkopie aus `git archive HEAD`, schreibgeschützt eingehängt, im Container aus `odata-recorder:race` (`--network none --memory=4g --memory-swap=4g --cpus=4`, `CGO_ENABLED=0`, ohne Detektor), je Mutation ein Prozess `go test -count=1 ./tests/lasttest/` ohne die 22 Tests von Teil 1 (Referenzdienst, Aufbau, Prüfung, Determinismus). Ohne Mutation grün. Last der letzten Minute 4,3 bis 7,8.

| Ergebnis | Mutation (roter Test) |
|---|---|
| grün (Befund) | Warm-up und Messphase vertauscht (F-1) · ungültige Läufe nicht ins Urteil; nur der letzte; nur der erste; Läufe über beide Betriebsarten (F-2) · Urteil gegen `abnahmewertMs`; Zeile `Abnahme` mit 100 (F-3) · Zählung von `auswerte` überschrieben; Modus verworfen (F-4) · Frist vor Abbruch (F-5) · Meldung auch mit beendetem Kontext (F-6) · Verspätung in der Schleife (I-2) |
| grün (kein Befund) | stdout nach 16 MiB nicht mehr entleert (der Fänger schreibt 8 MiB) · CPU-Zeit erst nach der Liste · `beendeBinaer` ohne zweite Prüfung an der Frist · Startwert der Mock-Liste fest auf 7 (Wert der Fixture; mit 1 rot in `TestNurDieListe`) |
| rot | p95 mit Warm-up (`TestFelderDerMessung`) · Urteil gegen 1000000 (`TestAbnahmeImAblauf`) · Probe nach dem Ende des Binary (`TestAblauf`) · Binary-Ende vor der Bereitschaft ignoriert (`TestBereitschaft`, `TestProzess`) · Nachlauf 60 s und 1 s (`TestNachlaufWarten`) · Nachlauf auf eine Zeile weniger (`TestSpaeteSpans`) · Abbruch erreicht den Lauf nicht (`TestAbbruch`) · Prüfung nach dem Aufbau entfernt (`TestAbbruch`) · Proben erst nach der Last (`TestProbenDauer`) · `binary_exit` und `binary_kill` (`TestHinweisImAblauf`) · Entities der Zeile `Aufbau` (`TestAufbauZeile`) · Last-Verstoß verändert (`TestVerstoesseImAblauf`) |

**Handlauf `make lasttest LASTTEST_PROFIL=rauch`** (19:02 bis 19:04 Uhr, Last 3,8 vor und 2,6 nach dem Lauf): Exit 0, `Aufbau exchanges=300 entities=300 daten_mib=2 dauer_s=8.306`; je Betriebsart 260 Anfragen, 260 Spans, 0 Fehler, `ohne_kind=0`, gültig; p95 `replay` 14,278 ms, `mock` 1,903 ms; `abnahme=nein grund=profil`; Proben 2,56 bis 3,37 bei Schwelle 10,00. Keine Zahl nach `LH-QA-01.a`.

**Grenzen und Signale** (zwei weitere Läufe `rauch` über den Treiber, 19:14 Uhr): `docker inspect` des laufenden Containers: Netz `none`, 2 CPU, 4294967296 Bytes Speicher und Speicher mit Swap, `AutoRemove`, 0 Mounts, nicht privilegiert; zwei Prozesse (`lasttest`, `odata-recorder`). TERM an den Treiber und INT an die Prozessgruppe im ersten Lauf der Betriebsart `replay`: Exit 2 nach unter 0,4 s, `lasttest: abgebrochen` je einmal von Programm und Treiber, keine Zeile `Messung`, `Urteil` oder `Abnahme`, danach kein Container des Images.

**Größenmarke:** `git diff --numstat 95779768..HEAD` über die 37 Pfade des Blocks selbst summiert: 4038, gleich dem Eintrag `ueberschreitung: 4100`; jeder geänderte Pfad außerhalb von `docs/`, `spec/`, `harness/vertraege/`, `src/` und `tests/integration/` steht im Block.

**Nicht gefahren:** Profil `abnahme` (Verifier), `make gegenproben`, `make gates`, `make test-race`, `make lint`, `make a-check`.

## Geprüft, ohne Befund

| Bereich | Geprüft |
|---|---|
| `tests/lasttest/` (Code) | Takt aus *t0* und *k* ohne Warten auf Antworten; keine Obergrenze der Verbindungen; Frist je Anfrage, Body vollständig, jeder Status außer 200 ein Fehler, Fehler über alle Phasen; keine Wiederholung, keine Zusatzanfrage; Warm-up und Messphase nach der Folge der Zeilen; Rang des p95 ganzzahlig; Vorrang der Gültigkeit; `binary` allein; Streuung, Toleranz, Urteil; Exit und Schreibfehler; kein Urteil auf den vier Wegen; `abnahme=ja` allein bei Profil `abnahme`, Ziel 100, beiden Betriebsarten und zwei bestandenen Urteilen (die Zeile `Abnahme` steht nach `SPEC-216` Nr. 13 bei jedem Profil, bei `rauch` mit `nein`); Umgebung des Binary; Ströme, Bereitschaft, Ende; Kommentare ohne Konjunktiv und ohne abwesenden Text; keine Direktive `//nolint`; Ausgaben ohne Token, Header, Body oder Pfad des Hosts |
| `tools/harness/`, `harness/mk/` | Folge Argument, Variablen, `info`, `build`, `run`; Optionen von `docker run`; Exit je Ausgang; Signale; Aufräumen von `TMPDIR`; kein Eintrag in `GATE_CHECKS`; keine Vorgaben im Fragment |
| `tools/gegenproben/` | die Argumente von `docker run` werden als Liste gelesen, jede unbekannte Option ist ein Befund; Mengen der Variablen, Exits und Signale gegen den Vertrag |
| `Dockerfile`, `.dockerignore`, `.golangci.yml` | Stufe `lasttest` ohne Netz mit Binary der Stufe `build`; Allowlist; die Ausnahme zu `forbidigo` gilt der einen Datei und trägt `# Why:` |
| `harness/README.md`, `SPEC-102`, `SPEC-216` Stand | Zeile unter den Werkzeugen, *Nicht behauptet*, Zählung der Fälle |
| Plan §3, §4, §6 | Tabelle, akzeptierte Negative mit Grund und Folge, Block der Marke, offene Punkte mit Adresse; außer F-4, F-6 und F-7 |

**Finding-Klassen dieses Laufs:** Fixture gleicht dem Wert der Mutation (F-1, F-3) · Verdrahtung des Ablaufs zu den Funktionen früherer Teile ohne Fänger (F-1 bis F-4) · Paar von Ursachen ohne Fänger (F-5) · Randform im Code entschieden (F-6) · veraltete Begründung einer Unterdrückung (F-7).

**Übergabe:** F-1 bis F-5 und F-7 an den Implementer; F-6 an den Architect (Randform) und danach an den Implementer; I-1 an den Architect; I-4 an den Verifier.

# Verifikation: slice-session-wartung

* Datum: 2026-10-23
* Rolle: verifier (frischer Kontext; kein Code, kein Vertrag, kein Plan, keine `AGENTS.md` geändert; `.claude/agents/architect.md` ist eine fremde, uncommittete Änderung und blieb ungelesen und unberührt)
* Gegenstand: Spitze `704f007`, Stempel `.harness/state/gates-passed.head` gleich HEAD zu Beginn. Basis der Größenmarke `9a2e28c`; Architect `bad4760`, `fae4ebb`, `1fea654`, `ad5171e`, `7774a01`, `d1073f3`; Implementer `186a15f`, `da9c702`, `686d323`, `4b1ed40`, `8870d37`, `1ff42a8`, `704f007`; Reviewer `2866c8a`.
* Plan: `docs/plan/planning/done/slice-session-wartung.md`
* Maßstab: DoD (§2), Tabelle *Zusage · Fänger* (§3, W1 bis W30), Raster und Zusatztabelle (§6) gegen `SPEC-123` Nr. 1 bis 12 und *Ungeprüft*, die Nachbarn `SPEC-048`, `SPEC-058`, `SPEC-105`, `SPEC-106`, `SPEC-109` Nr. 5 und 6, `SPEC-114`, `SPEC-119`, `SPEC-121`, `SPEC-122` Nr. 8; [ADR-0008](../plan/adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md) (Accepted, Entscheidung des Nutzers, nicht Gegenstand), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md); [LH-RB-04](../../spec/lastenheft.md#lh-rb-04--schutz-von-daten-und-zugangsdaten), [LH-FA-13](../../spec/lastenheft.md#lh-fa-13--management-und-diagnose-muss), [LH-FA-03](../../spec/lastenheft.md#lh-fa-03--persistenz-und-wiederverwendung-muss), [LH-QA-03](../../spec/lastenheft.md#lh-qa-03--isolation-und-konsistenz); `AGENTS.md` §3.1 bis §3.16.
* Hinweis zum Skill: `.harness/skills/closure-note-reviewer.md` gilt für Slices in `done/`; dieser Slice liegt in `in-progress/`, der Skill wurde nicht angewandt.
* Belege selbst gefahren (Abschnitt 4): `make test` und `make test-race` mit `--no-cache-filter` erzwungen, `make docs-check`, 84 Mutationen als Patch per `git apply` auf Wegwerfkopien von `git archive HEAD` (ohne `.git`, unter dem Scratchpad), Testbinaries mit und ohne `-race` für die Raten. Der Repo-Baum blieb unberührt (`git status --short` zeigt vor und nach dem Lauf nur `.claude/agents/architect.md`).

## Urteil

**Freigegeben mit Auflagen: kein HIGH, ein MEDIUM, drei LOW, zwei INFO.** Jeder Punkt der DoD, den der Code erfüllen kann, ist belegt (Abschnitt 1); die Punkte der Closure sind offen und führbar (Abschnitt 7). Die Befunde F-1 bis F-5 und I-2 bis I-4 des Reviews sind mit dem Stand geschlossen, jede von mir gefahrene Mutation der Zeilen W3 (g), W18, W25, W28, W29 und W30 ist aus dem benannten Grund rot (Abschnitte 3 und 5). Von 84 gültigen Mutationen blieben zwei grün, beide vom Plan als gleichwertig oder als akzeptiertes Negativ benannt (Abschnitt 5).

Die Auflagen vor der Closure:

1. V-1 (MEDIUM, Architect): die Randform *Abbruch beim Warten auf die Sperre des Wächters* hat in Raster und Vertrag keinen Ort; der Code entscheidet sie mit `sync.RWMutex` (ein beendeter Kontext wird beim Warten nicht gesehen).
2. V-4 (LOW, Planner): der Plan folgt dem Code an drei Stellen nicht (erledigte *Aufträge an den Implementer* stehen offen, Fakes und Zeilenzahl in §4 und §6).
3. V-2 (LOW, Architect) und V-3 (LOW, Planner): ein Satz des Vertrags und die Rate von W27 stimmen mit meiner Messung nicht überein.

* MEDIUM: V-1 (Architect).
* LOW: V-2 (Architect, Implementer), V-3 (Planner, Implementer), V-4 (Planner).
* INFO: V-5 (Architect), V-6 (Planner).

## 1. DoD Punkt für Punkt (Plan §2)

| # | DoD | Stand | Beleg |
|---|---|---|---|
| 1 | `Startup`, Stufen 1, 2 und 4 (`SPEC-123` Nr. 3 bis 5, 7): `foreign_key_check`/`integrity_check`, offene Sessions auf `incomplete`, Abgleich mit erneuter Referenzprüfung, Meldung fehlender und beschädigter Blobs | erfüllt | `src/hexagon/services/maintenance/wartung.go:53-67` (vier Stufen in fester Reihenfolge), `startup.go:15-41` (Stufe 1 und 2), `abgleich.go:14-80` (Stufe 4: Prüfen, drei Lesungen bei `Startup`); Fänger `TestReihenfolge`, `TestIntegritaet`, `TestOffeneSessions`, `TestAbgleich`, `TestPruefenBlobs`, `TestRemoveFehler`; Mutationen o1 bis o3, s8, b5, b6, i2, t8 rot (Abschnitt 5). Der Aufruf vor der Annahme von Requests ist Sache des Composition Root (Sendung an `slice-bootstrap-m2`, Abschnitt 6) |
| 2 | Aufbewahrung als Stufe 3 (`SPEC-123` Nr. 6): älter löschen, Quelle einer Baseline überspringen mit Warnung | erfüllt | `startup.go:45-88` (`grenze` einmal aus `Uhr.Now()`, Vergleich mit `Before` über die Grenze, `ErrReferenziert` übersprungen, `ErrNichtGefunden` übergangen); `TestAlterGrenze`, `TestAlterAus`, `TestAlterSehrGross`, `TestAlterAlleServices`, `TestAlterDelete`; Mutationen a1, a2b, a3, a4, a5, r1 bis r4, p1, b9, t11 rot |
| 3 | `CleanupBlobs` samt `BlobWaechter`: Zurückstellen, Karenz, `Commit` gegen `Remove` über die Sperre; Fänger der Fenster | erfüllt | `wartung.go:72-88`, `waechter.go:42-99`; `TestZurueckstellen`, `TestKarenz`, `TestWaechterDurchgereicht`, `TestWaechterZeit`, `TestSperreAbgleich`, `TestCommitMerktUnterSperre`, `TestAbgleicheNacheinander`, `TestWaechterGleichzeitig`, `TestSperreFrei`, `TestSperreStartup`, `TestKarenzUhr`; Raten in Abschnitt 5 |
| 4 | `make gates` grün | erfüllt, Lauf nach dem Commit des Berichts | `make test` (`pruefung test: ok`, ohne Cache) und `make test-race` (`pruefung race: ok`, Stufe 55,5 s, ohne Cache) selbst gefahren, Rückgabecode 0; `make gates`: Abschnitt 4 |
| 5 | Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-10-23-slice-session-wartung-review.md` (`2866c8a`): 0 HIGH, 4 MEDIUM, 3 LOW, 7 INFO; nachverifiziert in Abschnitt 3 |
| 6 | Doku-Update, falls ein öffentlicher Vertrag berührt ist | erfüllt | `SPEC-123` samt Nachzügen; `spec/spezifikation.md:805` nennt die Ausnahme von `SPEC-109` Nr. 5 (I-4); [ADR-0008](../plan/adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md) `Accepted` im Index `docs/plan/adr/README.md:15`; `spec/architecture.md` ohne Slice, Welle, Hash oder ADR-Bezug (Suche in der ganzen Datei: kein Treffer, `AGENTS.md` §3.4) |
| 7 | Closure-Notiz mit Steering-Loop-Lerneintrag | offen für die Closure | §7 trägt „Wird bei der Closure gefüllt“; Voraussetzungen erfüllt (Review und dieser Bericht liegen vor, Finding-Klassen in beiden benannt) |
| 8 | Beobachtungs-Register fortgeschrieben | offen für die Closure | `docs/plan/planning/observations/BEO-REPO/` zählt wie vor dem Slice (keine neue Evidence-Datei); Belege in Abschnitt 7 |
| 9 | Jedes Risiko aus §6 trägt einen Ausgang | offen für die Closure | dreizehn Risiken: fünf *entfallen* mit Ort der Entscheidung, eines *eingetreten* (`slice-session-use-case`), sieben *weiter offen* oder *offen*; Vorschläge in Abschnitt 7 |
| 10 | Drei Paarungen getragen | offen für die Closure | von der nächsten Welle-Closure geprüft |

Ein Häkchen setze ich nicht.

## 2. Tabelle *Zusage · Fänger*, Raster und Vertrag

**Tabelle.** Zu W1 bis W30 existiert je Zeile der genannte Test (`grep -n "^func Test"`, 34 Tests): `TestNewArgumente`, `TestNewWaechterArgumente`, `TestCtxNil`, `TestKontext`, `TestReihenfolge`, `TestFehlerJeStufe`, `TestFehlerJeStufeBereinigung`, `TestFehlerCodes`, `TestIntegritaet`, `TestOffeneSessions`, `TestAlterGrenze`, `TestAlterAus`, `TestAlterSehrGross`, `TestAlterAlleServices`, `TestAlterDelete`, `TestAbgleich`, `TestPruefenBlobs`, `TestStartupOhneSchutz`, `TestZurueckstellen`, `TestKarenz`, `TestRemoveFehler`, `TestWaechterDurchgereicht`, `TestWaechterZeit`, `TestBericht`, `TestProtokoll`, `TestFehlerWert`, `TestImporte`, `TestSperreAbgleich`, `TestCommitMerktUnterSperre`, `TestAbgleicheNacheinander`, `TestWaechterGleichzeitig`, `TestSperreFrei`, `TestSperreStartup`, `TestKarenzUhr`. Jede Zusage von `SPEC-123` Nr. 1 bis 12 hat eine Zeile; die Allaussage der Nr. 9 *Jeder Rückkehrpfad gibt die Sperre frei* hat W28 (elf Fälle, der Ausgang je Fall geprüft), die Nr. 7 *von ihrem Beginn bis zu ihrem Ende* hat W29 (vier Haltepunkte), die Nr. 8b *einmal, nach dem Nehmen der Sperre und nach dem Zurückstellen* hat W30 (a) bis (c).

**Vertrag gegen Code.** Gelesen und mit dem Wortlaut der Nr. 1 bis 12 verglichen, ohne Abweichung: Konfiguration (`wartung.go:36-41`, `waechter.go:34-40`); Kontext vor Port-Fehler (`fehler.go:13-16`, `startup.go:73-88`, `abgleich.go:27-49`); `Check` mit jedem Element als Befund (`startup.go:15-26`); `FindOpen` vor jeder Lesung und `Now()` nach der Sperre und dem Zurückstellen (`wartung.go:76-85`); `Before` über die Grenze (`startup.go:49,60`); drei Lesungen bei `Startup`, zwei bei `CleanupBlobs` (`abgleich.go:14-80`); Karenz „echt vor“ (`waechter.go:94-99`); Bereinigung der Referenzen bei jedem `Commit` (`waechter.go:81-92`); Abbildung nach Nr. 10 als Ausnahme von `SPEC-109` Nr. 5 (`fehler.go:20-29`). Das Protokoll (`protokoll.go`) trägt je Ereignis genau die Attribute der Nr. 11 und keinen Fehlertext.

**Raster.** Jede Zelle nennt einen Ort. Eine Randform ohne Ort habe ich gefunden: *Abbruch beim Warten auf die Sperre* (V-1). Die übrigen Zellen sind im Code so entschieden, wie der Wortlaut der genannten Stelle sagt; ich habe keine weitere im Code entschiedene Zelle gefunden.

**Akzeptierte Negative und Gleichwertigkeiten** (Plan §3 *Nicht gefangen*, `SPEC-123` *Ungeprüft*), gegen den Code geprüft:

| Negativ | Befund |
|---|---|
| Filter der ersten Lesung (Schritt 3) | Mutation i1b (nur `referenziert[ref] \|\|` entfernt) bleibt grün, wie benannt. Die Gleichwertigkeit gilt nur, solange die Menge nicht schrumpft; schrumpft sie, verhalten sich Code und Mutant verschieden und der Wortlaut des Grundes beschreibt den Mutanten, nicht den Code (V-2) |
| Fenster zwischen zweiter Lesung und `Remove`; Schreiber, der den Wächter umgeht; Dauer der Sperre; Startzeit; Panic eines Ports | Grund und Folge stimmen mit dem Code überein: die Sperre hält Commits auf, keine Referenzen; der Träger der Verdrahtung ist `slice-bootstrap-m2` mit Fänger (Abschnitt 6); die Frist der Capture-Pipeline ist 30 s (`SPEC-119` Nr. 1) |
| Karenz fest gegen `SQLITE_BUSY_TIMEOUT` | die Zahl „das Hundertfache“ stimmt nicht (600 s zu 5 s sind das 120fache, V-5); die Folge ist unverändert |
| Gegenstand `Session` in `mitCode` (Plan W5, gleichwertig) | gs1 bleibt grün, wie benannt; gs2 (Gegenstand `Session` und `mitCode` ohne Normalisierung zugleich) ist rot an `TestFehlerCodes` |

## 3. Nachverifikation der Befunde des Reviews

| Befund | Stand | Beleg | Ergebnis |
|---|---|---|---|
| F-1 Freigabe der Sperre auf jedem Rückkehrpfad (W28) | `TestSperreFrei` mit elf Fällen (`bereinigung_test.go:275-347`), Ausgang je Fall, `sperreFrei` prüft `TryLock` und `TryRLock` | neun eigene Mutationen ws1 bis ws9 (Freigabe fehlt auf Zurückgestellt, `FindOpen`-Fehler, Fehler des Abgleichs, gelungen; `Startup` Prüfstufe, Remove-Fehler, gelungen; `Commit` Fehler, gelungen): jede rot im Fall ihres Pfades (die Ausgabe nennt den Fall); ws4, ws7 und ws9 hängen im Vollauf den Prozess (Zeitgrenze), im Lauf `-run TestSperreFrei` rot im Fall des Pfades | geschlossen |
| F-2 Sperre der Prüfstufe von `Startup` (W29) | `TestSperreStartup` mit vier Haltepunkten (`bereinigung_test.go:361-384`) | m29a (Sperre nach dem Prüfen): rot in (a) und (b); m29b (Sperre nur um das Prüfen): rot in (c) und (d) und in `TestAbgleicheNacheinander`; rot 16 von 16 Prozessen, 800 von 800 Wiederholungen, mit und ohne Detektor; m29a auch bei `--cpus=0.05` 8 von 8, 400 von 400 | geschlossen |
| F-3 Kontext an fünf Operationen des Wächters (W18) | `pruefeKontextDurchgereicht` (`waechter_test.go:122-155`): je Operation beendeter und lebender `ctx` mit Wert | `context.Background()` in `Begin`, `Open`, `Verify`, `List`, `Remove`: je rot in `TestWaechterDurchgereicht` (`Begin` zusätzlich `TestWaechterZeit`, `Remove` zusätzlich `TestKontext`) | geschlossen |
| F-4 `Now()` einmal in der Karenz (W30) | `TestKarenzUhr` (`waechter_test.go:259-293`) | n1 (`Now()` je Kandidat) und n4 (zusätzlich `Uhr.Now()` im Wächter je Kandidat): rot (a); n2 (`Now()` vor der Sperre): rot (b) und (c); n3 (nach der Sperre, vor `FindOpen`): rot (b) und (c) | geschlossen |
| F-5 Kontextfehler beendet jede Stufe (W3 (g)) | `TestFehlerJeStufe` und `TestFehlerJeStufeBereinigung` je Fall zweimal, `pruefeFehlerArt` prüft `==`, kein `*Fehler`, Zähler | g01 bis g13 (Kontextfehler an `Check`, `FindOpen` Stufe 2, `SetStatus`, `Dienste.List`, `Sessions.List`, `Delete`, `ReferencedBlobs` der Prüfstufe, `Verify`, `Blobs.List`, beide Lesungen, `Remove`, `FindOpen` der Bereinigung verschluckt): alle 13 rot an dem Fall der Stelle (`<nil>` statt `context canceled`) | geschlossen |
| F-6 Plan W5 und erledigter Auftrag | W5 nennt die gleichwertige Mutation als gleichwertig (gs1 grün, gs2 rot, bestätigt); die *Aufträge an den Implementer* (Plan Zeile 145 bis 153) stehen weiter als offen, obwohl `4b1ed40` bis `704f007` sie erledigen | | W5 geschlossen, Aufträge offen (V-4) |
| F-7 Fakes in §4 | Plan Zeile 177: „4 Port-Fakes … Sessions und Services in einem Typ“; der Code führt `fakeSessions`, `fakeDienste`, `fakeSpeicher`, `fakeBlobs` (mit `fakeWriter`) und `fang` | | offen, Sache des Planners (V-4) |
| I-1 Filter der ersten Lesung | der Architect hat es als akzeptiertes Negativ in *Ungeprüft* gefasst; der Plan führt es in §3 | i1b grün; der Wortlaut des Grundes stimmt im Schrumpffall nicht (V-2) | anders gelöst, Rest V-2 |
| I-2 Wirkung von W25 | `waechter_test.go:244-246`: „Zusatzprüfung zur Wirkung, kein Fänger … Fänger der Zusage ist `TryLock` im Halt“; Plan W25 nennt die Zahlen des Reviewers | r25 (Zeit nach der Freigabe): rot 16 von 16, 800 von 800 | geschlossen |
| I-3 `TestCtxNil` nennt die Panik des Fakes | `konfig_test.go:81-83`; Plan W2 | x1, x2 (Prüfung entfernt): rot durch Panik des Fakes in `TestCtxNil` | geschlossen |
| I-4 `SPEC-109` Nr. 5 nennt die Ausnahme | `spec/spezifikation.md:805` | gelesen | geschlossen |
| I-5 Größe | Messung in Abschnitt 6; der Planner trägt sie in §4 und §7 ein | | offen für die Closure |
| I-6 Abweichungen des Implementers (`python3`, `sed -i`, `docker run`) | kein Ergebnis betroffen; die Gates laufen über `make` | | kein Befund |
| I-7 Bericht des Implementers | liegt weiter nicht im Repo; auch die Nachbesserungsrunde (`4b1ed40` bis `704f007`) trägt nur die Kopfzeilen | | V-6 |

## 4. Gates und Belege

| Lauf | Ergebnis |
|---|---|
| `git status --short` vor dem Lauf | ` M .claude/agents/architect.md` (fremd) |
| `git diff 9a2e28c..HEAD -- src` außerhalb von `src/hexagon/ports/inbound/maintenance.go` und `src/hexagon/services/maintenance/` | leer; der Diff berührt außerhalb von `src` nur `spec/spezifikation.md`, den Plan und den Review-Report |
| `git diff da9c702..HEAD -- wartung.go startup.go abgleich.go waechter.go fehler.go protokoll.go inbound/maintenance.go` | leer: der Produktcode ist seit `da9c702` unverändert; die vier Commits `4b1ed40` bis `704f007` ändern nur Testdateien |
| `make test DOCKER_BUILD='docker build --progress=plain --no-cache-filter test'` | Rückgabecode 0, `pruefung test: ok` (45,1 s) |
| `make test-race DOCKER_BUILD='docker build --progress=plain --no-cache-filter race'` | Rückgabecode 0, `pruefung race: ok` (55,5 s); kein Datenrennen im Bestand |
| `make docs-check` vor dem Commit des Berichts | `d-check: … Datei(en) geprüft, 0 Befund(e)` (Zahl in der Nachricht an den Planner) |
| kein `//nolint` im Diff; `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/`, `tools/`, `Dockerfile` unverändert seit `9a2e28c` | `grep` ohne Treffer, `git diff --stat` leer |
| `make gates` nach dem Commit des Berichts | siehe die Nachricht an den Planner (Rückgabecode, Stempel gleich `git rev-parse HEAD`) |

## 5. Mutationen und Raten

**Methode.** Jede Mutation ist ein Patch (`git apply --unidiff-zero --recount --ignore-whitespace`) auf einer frischen Kopie von `go.mod`, `go.sum` und `src/` aus `git archive HEAD`; Lauf `go test ./src/hexagon/services/maintenance/` im Image `odata-recorder:test` (`--network none`, Kopie `:ro`, Build-Cache in einem Wegwerf-Volume). Rot zählt nur, wenn die Ausgabe den Fall nennt, den die Mutation treffen soll. Drei Patche übersetzten nicht (a2, c5, i1: „declared and not used“) und sind durch a2b, c5b, i1b ersetzt; bleiben 84 gültige Mutationen, 82 rot aus dem benannten Grund, 2 grün.

| Zeile | Mutation (Kennung) | Ergebnis |
|---|---|---|
| W28 | Freigabe fehlt auf einem Pfad (ws1 bis ws9, Pfade siehe Abschnitt 3) | rot `TestSperreFrei`, im Fall des Pfades |
| W29 | m29a, m29b | rot `TestSperreStartup` (a)/(b) und (c)/(d) |
| W30 | n1, n2, n3, n4 | rot `TestKarenzUhr` (a), (b), (c) wie benannt |
| W18 | `context.Background()` in `Begin`, `Open`, `Verify`, `List`, `Remove` (w18_begin_bg, w18_open_bg, w18_verify_bg, w18_list_bg, w18_remove_bg) | rot `TestWaechterDurchgereicht` |
| W3 (g) | g01 bis g13 | rot `TestFehlerJeStufe`, `TestFehlerJeStufeBereinigung` (g06, g08, g12 zusätzlich `TestKontext`) |
| W3 | k1 (nur `Canceled` ist Kontextfehler), k2 (nur `DeadlineExceeded`) | rot `TestKontext` (beide Formen des Kontextfehlers sind Fall) |
| W2 | x1, x2 (Prüfung auf `nil` entfernt) | rot `TestCtxNil` durch Panik des Fakes |
| W1 | b1 (`Log` nil gültig), b2 (`karenz` 0 gültig), b4 (`MaxAge` 0 ungültig) | rot `TestNewArgumente`, `TestNewWaechterArgumente`; b4 zusätzlich weitere Tests (Gutfall `MaxAge` 0) |
| W4 | o1, o2, o3 (benachbarte Stufen vertauscht) | rot `TestReihenfolge` |
| W5 | f1 (`mitCode` ohne Normalisierung), f2 (Status 404), f3 (alles `ODR-E9001`), gs2; gs1 (nur Gegenstand `Session`) | rot `TestFehlerCodes`; **gs1 grün (gleichwertig, im Plan benannt)** |
| W6 | s8 (Befund nur bei `nil`), b5 (nur der erste Befund), b11 (Zahl im Eintrag) | rot `TestIntegritaet`, `TestProtokoll` |
| W7 | b6 (Ziel `complete`) | rot `TestOffeneSessions`, `TestStartupOhneSchutz` |
| W8, W9 | a1 (`<=`), a2b (`Sub` und `>`), a3 (`MaxAge` 0 schaltet nicht aus), a4 (`+MaxAge`), a5 (Richtung) | rot `TestAlterGrenze`, `TestAlterAus`, `TestAlterSehrGross` (a2b nur dort: Jahr 1 bleibt wegen der Sättigung) |
| W11 | r1, r2 (`==` statt `errors.Is`), r3 (`ErrNichtGefunden` bricht ab), r4 (`ErrReferenziert` bricht ab), b9 (kein Eintrag), p1 (Stufe `ERROR`), t11 (Kennungen vertauscht) | rot `TestAlterDelete`, `TestProtokoll` |
| W12 | i2 (zweiter Filter weg) | rot `TestAbgleich`; **i1b (erster Filter weg) grün (akzeptiertes Negativ, V-2)** |
| W14, W15 | z1 (Zurückstellen erst ab zwei Sessions), z2 (Merker fehlt) | rot `TestZurueckstellen`, `TestSperreFrei`, `TestKarenzUhr` |
| W16, W19 | c1 (`<=`), c2 (doppelte Karenz), c3 (erster Commit gilt), c4, c5b (Bereinigung zu früh, nie), c6 (gescheiterter Commit merkt), b14 (Zeit ignoriert), b15 (Karenz ignoriert) | rot `TestKarenz`, `TestWaechterZeit` (b15 zusätzlich `TestCommitMerktUnterSperre`, `TestKarenzUhr`, `TestWaechterGleichzeitig`) |
| W17 | t8 (Bericht bei Fehler leer) | rot `TestRemoveFehler`, `TestFehlerJeStufe`, `TestSperreFrei` |
| W24 bis W27 | t4 (`Begin` ohne Wächter-Writer) und die vier Mutanten der Raten | rot, siehe Raten |

**Raten.** Messender: der Verifier, am 2026-10-23, Host mit 20 Kernen, Last des Hosts 5 bis 16 (zum Teil von den Läufen selbst). Testbinaries (`go test -c`) ohne Detektor im Image der Stufe `test`, mit Detektor (`-race`, cgo) im Image der Stufe `race`; ein Prozess je Container, `-test.count=50` (W27: 1; unmutiert 20), `-test.v`. „Prozesse“ sind die mit Rückgabecode ungleich 0, „Wdh.“ die Zeilen `--- FAIL` von `--- FAIL` und `--- PASS`.

| Fänger | Mutation | Last | ohne Detektor: Prozesse · Wdh. rot | mit Detektor: Prozesse · Wdh. rot |
|---|---|---|---|---|
| W24 (c) `TestSperreAbgleich` | keine Lesersperre im `Commit` (r24) | 16 Prozesse | 16 von 16 · 800 von 800 | 16 von 16 · 800 von 800 |
| dto. | r24 | `--cpus=0.05`, 8 Prozesse | 8 von 8 · 400 von 400 | 8 von 8 · 400 von 400 |
| dto. | r24 | `--cpus=0.02`, 12 Prozesse | 12 von 12 · 600 von 600 | 12 von 12 · 600 von 600 |
| W25 `TestCommitMerktUnterSperre` | Zeit nach der Freigabe (r25) | 16 Prozesse; `--cpus=0.05`, 8 | 16 von 16 · 800 von 800; 8 von 8 · 400 von 400 | dto. |
| W26 `TestAbgleicheNacheinander` | beide ohne Sperre (r26); nur `CleanupBlobs` (r26c); nur `Startup` (r26s) | 16 Prozesse | je 16 von 16 · 800 von 800 | je 16 von 16 · 800 von 800 |
| dto. | r26 | `--cpus=0.05`, 8; `--cpus=0.02`, 12 | 8 von 8 · 400 von 400; 12 von 12 · 600 von 600 | dto. |
| W27 `TestWaechterGleichzeitig` (Prozesse, `count` 1) | Zustand ohne eigene Sperre (r27) | 40 Prozesse (zwei Läufe) | 40 von 40; 40 von 40 | 40 von 40 |
| dto. | r27 | `--cpus=0.05`, 16 Prozesse | **14 von 16** | 16 von 16 |
| dto. | r27 | `--cpus=0.05`, 24 Prozesse (drei Läufe) | **22, 23 und 24 von 24** | – |
| W28 `TestSperreFrei` | ws1, ws8 | 16 Prozesse; ws1 bei `--cpus=0.05`, 8 | je 16 von 16 · 800 von 800; 8 von 8 · 400 von 400 | dto. |
| W29 `TestSperreStartup` | m29a, m29b | 16 Prozesse; m29a bei `--cpus=0.05`, 8 | je 16 von 16 · 800 von 800; 8 von 8 · 400 von 400 | dto. |
| W30 `TestKarenzUhr` | n1, n2, n3 | 16 Prozesse; n2 bei `--cpus=0.05`, 8 | je 16 von 16 · 800 von 800; 8 von 8 · 400 von 400 | dto. |
| die sechs Tests W24 bis W26, W28 bis W30 | unmutiert | 16 Prozesse, 50 mal; `--cpus=0.05`, 8 | 0 von 16 · 0 von 4800; 0 von 8 · 0 von 2400 | dto. |
| W24, W26 | unmutiert | `--cpus=0.02`, 12 | 0 von 12 · 0 von 1200 | dto. |
| W27 | unmutiert | 40 Prozesse, 20 mal; `--cpus=0.05`, 16 | 0 von 40 · 0 von 800; 0 von 16 · 0 von 320 | dto. |

Die Zahlen des Plans (W24 (c), W26, W27: 16 von 16, 800 von 800, 8 von 8, 400 von 400, 40 von 40, unmutiert 0 von 800) bestätige ich bis auf eine Zeile: W27 ohne Detektor bei `--cpus=0.05` (V-3). W28 bis W30 sind deterministisch (Haltepunkte und nacheinander laufende Aufrufe) und tragen keine Rate des Implementers; meine Messung zeigt dort keinen grünen Prozess. Mit Detektor ist W27 in allen Läufen 100 %.

**Eigene Proben** (nicht im Repo, in Wegwerfkopien): `TestSchrumpft` (Datei `ra` in der ersten Lesung referenziert, in der zweiten nicht): Code `Entfernt` 0, Datei besteht; Mutant i1b `Entfernt` 1, Datei weg (V-2). `TestProbeLockWarteAufCtx` (ein `Commit` hält im Fake, `CleanupBlobs` mit beendetem `ctx`): der Aufruf kehrt nach 500 ms nicht zurück und liefert `context canceled` erst nach der Freigabe des `Commit` (V-1).

## 6. Plan folgt dem Code, Größe, Sendungen

**Plan (`AGENTS.md` §3.9).** Kopf, §1, §2, §3 (Tabelle W1 bis W30, Raster, Zusatztabelle mit W28 bis W30), §8 stimmen mit dem Stand überein (34 Tests, elf Einträge im Register). Drei Stellen folgen dem Code nicht (V-4): die *Aufträge an den Implementer aus dem Review* (Zeile 145 bis 153) stehen als offen; §4 (Zeile 177) zählt 27 Zeilen und vier Fakes; §6 (Zeile 265) nennt „27 Zeilen in der Tabelle, Schätzung 2150“.

**Größe (`AGENTS.md` §3.14).** Messung: `git diff --numstat 9a2e28c..HEAD -- src/hexagon/ports/inbound src/hexagon/services/maintenance`, Einfügungen 2973, Löschungen 0, Summe **2973** (Marke 2700; Reviewer bei `686d323`: 2716). Die Teile: `maintenance.go` 32, Produktcode 437 (`abgleich.go` 99, `fehler.go` 29, `protokoll.go` 34, `startup.go` 88, `waechter.go` 99, `wartung.go` 88), Tests 2504 (`helfer_test.go` 750, `startup_test.go` 449, `bereinigung_test.go` 384, `waechter_test.go` 358, `konfig_test.go` 169, `alter_test.go` 143, `abgleich_test.go` 125, `protokoll_test.go` 98, `export_test.go` 28). Die Pfade der Marke sind vollständig: außerhalb ändert der Diff `src` nicht. Die Überschreitung um 273 Zeilen (10,1 %) besteht aus 16 Zeilen, die schon bei `686d323` (2716) über der Marke standen, und 257 Zeilen der Nachbesserungsrunde (`4b1ed40` bis `704f007`: `git diff --numstat 686d323..HEAD` 286 Einfügungen, 29 Löschungen, nur Tests in `bereinigung_test.go`, `waechter_test.go`, `startup_test.go`, `helfer_test.go`, `konfig_test.go`; Produktcode unverändert). Eine Rückführung („über 2700 Zeilen“, §4) ist bei der Closure zu begründen; ein vorab benannter Schnitt stand nicht in §4. Zur Schätzung nach der Formel von §3.14: 30 Zeilen, davon 9 nebenläufig (W16, W19, W24 bis W30), 21 andere, fünf Fakes: 9 × 85 + 21 × 40 + 5 × 200 = 2605, mal 1,25 = 3256, aufgerundet 3300. Ob eine Marke über 3000 hier einen Schnitt verlangt hätte, entscheidet der Planner; die Zeilen W28 bis W30 und zwei der Fakes entstanden erst mit dem Review.

**Sendungen und Adressen (`AGENTS.md` §3.13).** Geprüft in `docs/plan/planning/open/`:

| Nehmer | Sendung | §1 / DoD | *Ausdrücklich NICHT* trifft sie? |
|---|---|---|---|
| `slice-bootstrap-m2` | Startwartung vor der ersten Session und den Listenern; ein Wächter je Prozess (`NewBlobWaechter`, `StandardKarenz`) für jeden Schreiber und die Wartung, mit Fänger | §1 Zeile 38 und 40, DoD Zeile 74, Kopf Zeile 16 | nein (vier Ausschlüsse: Modus `mock`/`hybrid`, CLI und Admin-Listener, Lasttest, `slice-protokolltreue`) |
| `slice-session-use-case` | Aufruf von `CleanupBlobs` nach dem Löschen einer Session; Fehler und `Zurueckgestellt` ändern das Ergebnis nicht | §1 Zeile 44, DoD Zeile 74, Risiko Zeile 143 | nein (der Ausschluss `MaintenanceUseCase` nennt den Port als geliefert) |
| `slice-workspace-use-case` | Aufruf nach dem Löschen eines Workspace; dieselbe Pflicht | §1 Zeile 40, DoD Zeile 67 | nein (drei Ausschlüsse: Mutation des Entity-State, Snapshots/Restore, Admin-Endpunkt) |

Alle drei liegen in `open/`, keiner in `done/`. `slice-baseline-import` setzt `slice-session-wartung` in `done/` voraus (Start, Zeile 101), ohne Sendung. `spec/architecture.md` ist im Diff unverändert und enthält weder Slice, Welle, Hash noch ADR-Bezug (`AGENTS.md` §3.4).

## 7. Offen für die Closure (Planner)

Kein Befund des Codes; damit nichts verloren geht:

1. **Größe** (§4, §7): Messung 2973 gegen Marke 2700 in §4 eintragen, die Überschreitung begründen (Tests der Nachbesserungsrunde, W28 bis W30 und zwei Fakes mehr als die Schätzung), die Schätzung nach §3.14 berichtigen (30 Zeilen, fünf Fakes, Abschnitt 6); Risiko „Größe dieses Slice“ (§6) mit Ausgang; `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` ist ein weiterer Beleg.
2. **Plan nachziehen** (V-4): die *Aufträge an den Implementer aus dem Review* als erledigt führen oder entfernen (die sieben sind mit `4b1ed40` bis `704f007` erledigt), §4 und §6 auf 30 Zeilen und fünf Fakes.
3. **Register** (`../observations/`): `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo` (der dritte Beleg: Bericht des Implementers fehlt im Repo, Review I-7, hier V-6; der Ausgang ist Entscheidung des Nutzers); `BEO-REPO/befund-an-rolle-ohne-artefakt` und `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×; ein weiterer Beleg ist hier nicht entstanden, die Zählung folgt aus den Dateien); `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (V-3 und die Zeilen des Reviews zu `--cpus=0.02`); `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` und `BEO-REPO/randform-erst-im-review-entschieden` (F-1 bis F-5 des Reviews, V-1 hier); `BEO-REPO/akzeptiertes-negativ-ohne-folge` (V-2: Gleichwertigkeit mit einem Grund, der im Schrumpffall nicht gilt); `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (V-4).
4. **Risiko-Ausgänge** (§6): *entfallen* mit Ort der Entscheidung tragen fünf (Bezugszeitpunkt, `Commit` gegen `Remove`, Exchange zwischen `Commit` und `Append`, Fenster der übrigen Schreiber, Datenbank oder Blob beschädigt), *eingetreten* trägt eines (Größe beider Use Cases, `slice-session-use-case`). *Weiter offen* oder *offen* tragen sieben: Nebenläufigkeit der Bereinigung (die Raten sind gemessen, Abschnitt 5; der Ausgang folgt dem Entscheid zu V-1 und V-3), Randformen (die Suche durch Review und Verifikation ist erfolgt; V-1 ist die eine gefundene Zelle ohne Ort), offene Annahme (Träger `slice-session-use-case`), Verdrahtung des Wächters (Träger `slice-bootstrap-m2`), Größe dieses Slice, Rate ohne Detektor, Bericht des Implementers.
5. **Lerneintrag** (§7) mit den Finding-Klassen aus Review und diesem Bericht; ein Feld `liegt in` nur, wenn mit dem Slice etwas verkörpert wurde.
6. **Move nach `done/`** von Hand (zwei Commits nach `AGENTS.md` §3.3: erst Inhalt, dann reiner `git mv`; `.claude/agents/architect.md` bleibt beim Nutzer); Nachzüge: Verweise auf den Pfad in `roadmap.md` und den offenen Plänen (`slice-session-use-case`, `slice-workspace-use-case`, `slice-bootstrap-m2`, `slice-baseline-import`), die Zeile des Slice in der Roadmap, und der Vermerk „in `done/`“ in den Starts der Nehmer.

## Findings

| ID | Schwere | Ort | Aussage | Ausfallszenario | Quelle | Adressat |
|---|---|---|---|---|---|---|
| V-1 | MEDIUM | `src/hexagon/services/maintenance/wartung.go:76` · `w.k.Blobs.sperre.Lock()` und `abgleich.go:15`; `SPEC-123` Nr. 2 und Raster (Zeile `Startup`, `CleanupBlobs` · `ctx` · Abbruch) | Eine Randform ohne Ort: *Abbruch des Kontexts beim Warten auf die Sperre des Wächters*. Nr. 2 sagt, ein Kontextfehler beende die Operation an der Stelle; das Warten auf die Sperre sieht den Kontext nicht (`sync.RWMutex.Lock` kennt keinen), der Code entscheidet es still. *Ungeprüft* nennt nur die Dauer der Sperre für wartende `Commit` | Ein Admin-Aufruf löscht eine Session; der Abgleich eines anderen Aufrufs oder ein `Commit` hält die Sperre. Der Client trennt, der Kontext endet, der Aufruf von `CleanupBlobs` bleibt in `Lock()` hängen, bis die Sperre frei ist; erst dann kehrt er mit `context canceled` zurück (Probe `TestProbeLockWarteAufCtx`: nach 500 ms nicht zurück, nach der Freigabe `context canceled`). Bei einem langen Abgleich stauen sich solche Aufrufe über ihre Frist und über `SHUTDOWN_TIMEOUT` hinaus, und weil ein wartender Schreiber neue Leser aufhält, warten auch die `Commit` der Capture-Pipeline hinter ihm | `AGENTS.md` §3.16 (Zelle ohne Ort, im Code entschieden), §3.12 | Architect (Entscheidung: als *Ungeprüft* mit Folge führen, oder Vertrag ändern); Implementer erst danach |
| V-2 | LOW | `spec/spezifikation.md` `SPEC-123` *Ungeprüft*, Punkt „Filter der ersten Lesung“; Plan §3 *Nicht gefangen*; `abgleich.go:63-67` | Der Grund „gleichwertig, solange die Menge nicht schrumpft; schrumpft sie, ist ein Kandidat, der nun unreferenziert ist, zu Recht entfernt“ beschreibt den Mutanten ohne den Filter, nicht den Code: im Schrumpffall ist die Referenz der ersten Lesung kein Kandidat (Schritt 3) und die Datei bleibt, bis der nächste Abgleich sie sieht. Mutation i1b bleibt grün; Probe `TestSchrumpft`: Code `Entfernt` 0, Mutant 1 | Wer Schritt 3 als überflüssig streicht, ändert das Verhalten bei einem Löschen zwischen den Lesungen; kein Test wird rot, der Vertrag nennt das Gegenteil dessen, was der Code tut. Datenverlust entsteht nicht (die Datei war zwischen den Lesungen unreferenziert geworden) | `AGENTS.md` §3.15 (Grund der Gleichwertigkeit), §3.7 | Architect (Wortlaut); Implementer (optional ein Fänger: die zweite Lesung liefert eine Referenz weniger) |
| V-3 | LOW | Plan §3 Zeile 106 und 107 (Rate W27 ohne Detektor bei `--cpus=0.05`: „16 von 16 … in keinem war ein Prozess grün“) | Die Zahl ist eine Einzelmessung. Meine vier Läufe bei `--cpus=0.05` ohne Detektor: 14 von 16, 22, 23 und 24 von 24 (83 von 88, je Lauf 0 bis 12,5 % grüne Prozesse). Der Plan verlangt für Fänger mit Wahrscheinlichkeit „die Spanne mehrerer Läufe und den Messenden“ (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`) | Wer die Zeile liest, hält W27 unter `make test` für sicher rot; unter Last sind 5 von 88 Prozessen grün. Mit Detektor (`make test-race`, das Gate) ist die Rate 100 % | `AGENTS.md` §3.10 | Planner (Spanne und Messenden eintragen); Implementer |
| V-4 | LOW | Plan Zeile 145 bis 153, 177, 265 | Der Plan folgt dem Code nicht: die sieben *Aufträge an den Implementer* stehen als offen, sind aber mit `4b1ed40` bis `704f007` erledigt (F-6); §4 zählt vier Fakes mit einem Typ für Sessions und Services, der Code hat fünf (F-7); §4 und §6 nennen 27 Zeilen der Tabelle, §3 führt 30 | Die nächste Änderung liest erledigte Aufträge als offen und rechnet die Marke aus 27 Zeilen; die Schätzung nach der Formel steigt von 2688 auf 3256 (Abschnitt 6) | `AGENTS.md` §3.9, §3.14 | Planner |
| V-5 | INFO | `SPEC-123` Nr. 8b, Plan §3 (Folge der Karenz gegen `SQLITE_BUSY_TIMEOUT`) | „die Karenz von zehn Minuten ist das Hundertfache davon“ und „der Default liegt beim Hundertsten der Karenz“: 600 s zu 5 s sind das 120fache | keiner; die Folge im Betrieb gilt unverändert | `AGENTS.md` §3.7 | Architect |
| V-6 | INFO | Repo: kein Bericht des Implementers zu `4b1ed40` bis `704f007` | Die Nachbesserungsrunde trägt nur Kopfzeilen; *Zusage · Mutation · roter Test* und die Raten des Implementers liegen nicht im Repo (`AGENTS.md` §3.10 verlangt den Bericht an den Aufrufer, nicht ins Repo). Ich habe alle Zeilen neu gefahren und nichts übernommen | keiner für den Code; für das Register (Abschnitt 7, Nr. 3) | `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo` | Planner (Closure §7) |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/inbound/maintenance.go`, Produktcode (`wartung.go`, `startup.go`, `abgleich.go`, `waechter.go`, `fehler.go`, `protokoll.go`) | geprüft gegen `SPEC-123` Nr. 1 bis 12; bis auf V-1 und V-2 ohne Befund; seit `da9c702` unverändert |
| Fehlerabbildung Nr. 10 | f1, f2, f3, gs2 rot, gs1 gleichwertig wie benannt; `Remove` mit `ErrBlobFehlt` und `ErrBlobBeschaedigt` ist `ODR-E3001` |
| Kontextfehler | k1 und k2 rot: beide Formen (`Canceled`, `DeadlineExceeded`) sind Fall; g01 bis g13 rot |
| Altersvergleich | a1 bis a5 rot; Jahr 1 und Jahr 1800 bei größtem `MaxAge` (`TestAlterSehrGross`), `MaxAge` 0 (`TestAlterAus`) |
| Wächter | Sperren, Zeit des Commit, Karenz „echt vor“, Bereinigung: c1 bis c6, b14, b15, t4, r24 bis r27 rot; kein Datenrennen im Bestand (`make test-race`) |
| Tests (`helfer_test.go`, `export_test.go` und die sieben Testdateien je Themenkreis) | die Kommentare der nach dem Review geänderten Tests sagen nur zu, was der Test fängt (`TestCtxNil`, Zusatzprüfung in `TestCommitMerktUnterSperre`, `sperreFrei`); keine Konjunktive über verworfene Alternativen |
| Hard Rules `AGENTS.md` §3.1, §3.2, §3.4, §3.13 | kein `//nolint`, kein lokaler Toolchain-Aufruf; die Gates liefen über `make`, die Mutationsläufe (Auftrag) über `docker run` im Image der Stufe `test` bzw. `race` mit der Kopie `:ro`, einem Wegwerf-Volume für den Build-Cache und einem schreibbaren Scratch-Verzeichnis nur für die Testbinaries (gleiche Klasse wie I-6 des Reviews, kein Ergebnis betroffen); `spec/architecture.md` ohne Slice-, Wellen-, Hash- oder ADR-Bezug; die Sendungen tragen ihre Nehmer |

## Verdikt

**Freigegeben mit Auflagen.** Die DoD ist erfüllt, soweit der Code sie erfüllen kann, und Plan, Vertrag und Code stimmen überein, bis auf V-1 bis V-4. Vor der Closure: V-1 (der Architect entscheidet die Randform), V-4 und V-3 (der Planner zieht den Plan nach), V-2 (der Architect schärft den Wortlaut). Die Closure-Punkte stehen in Abschnitt 7; `make gates` läuft nach dem Commit dieses Berichts.

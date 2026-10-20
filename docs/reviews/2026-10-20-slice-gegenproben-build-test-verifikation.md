# Verifikation: slice-gegenproben-build-test

* Datum: 2026-10-20
* Rolle: verifier (frischer Kontext, kein Self-Review, kein Produktcode geändert)
* Gegenstand: `git diff f6a1431^..HEAD` (Stand `a1aafd8`); Basis der Verhaltensgleichheit `2091ba6`
* Plan: `docs/plan/planning/done/slice-gegenproben-build-test.md`
* Verträge: `harness/vertraege/SPEC-036-test.md`, `harness/vertraege/SPEC-102-gegenproben.md`
* Eingang: Review `docs/reviews/2026-10-20-slice-gegenproben-build-test-review.md`; ein Bericht des Implementers mit Sensor-Belegen lag nicht vor, deshalb sind alle Belege unten selbst gefahren.
* Maßstab: DoD (Plan §2), Zusage · Fänger (Plan §3, `AGENTS.md` §3.15), Verträge, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), Baseline-Regelwerk Modul 11 (Bewusstes Brechen).
* Methode: `git archive HEAD` in Wegwerfkopien im Scratchpad; Mutationen mit `perl` an der Kopie; Docker `golang:1.27-alpine` mit `--network none`. Die Unit-Tests von `tools/pruefung` und die neun Fälle 25 bis 33 liefen gegen das mutierte Binary (Treiber: eine Testdatei nur in der Kopie, die `run` mit den Fällen aus `testFaelle()` und `buildFaelle()` ruft). Dockerfile-Läufe als echter `docker build` auf Kopien. Der Arbeitsbaum blieb unverändert (`git status` leer, Stempel unverändert).
* Closure-Note-Reviewer (`.harness/skills/closure-note-reviewer.md`): nicht einschlägig; der Skill gilt für `done/`, und die Closure-Notiz des Slice steht noch aus (§7: „Wird bei der Closure geschrieben").

## Zusammenfassung

Keine HIGH. Fünf MEDIUM (alle „Zusage ohne Fänger" nach `AGENTS.md` §3.15, zwei davon vom Plan offengelegt), vier LOW, drei INFO.
`make gates` ist grün (Exit 0, selbst gefahren). Die Fälle 25 bis 33 laufen in der Stufe `gegenproben` grün (`33 Fälle, 0 rot`, ohne Build-Cache der Stufe).
Der Slice ist **nicht abschließbar**, solange die DoD-Zeile Zeitbudget unbelegt ist und die Closure-Arbeit (Notiz, Register, Risiko-Ausgänge, M1 *erreicht*) fehlt; beides ist im Plan so benannt und ehrlich dargestellt.
Die Review-Befunde F3, F4, F5, F6, F9 sind behoben; F1 ist nur offengelegt, nicht behoben (V1); F2 ist ehrlich offen.

## 1. DoD-Zeilen (Plan §2)

Alle Kästchen im Plan sind noch leer; das ist vor der Closure stimmig. Befund je Zeile gegen den Code:

| DoD-Zeile | Befund | Beleg |
|---|---|---|
| Prüfungen als Go-Programm `tools/pruefung`, `Dockerfile` ruft es, `go-gates.mk` ohne eigene Prüfung, Wirkung unverändert | erfüllt, mit einer Abweichung (V7) | `tools/pruefung/{go.mod,main.go,main_test.go}`; `Dockerfile` Stufen `pruefung`, `test` (`pruefung test /src`), `build` (`pruefung build /out/odata-recorder "$VERSION"`); `go-gates.mk` ohne die zwei Prüfzeilen; Verhaltensgleichheit unter 4 |
| Fälle 25 bis 33 in `tools/gegenproben` und `SPEC-102`; Grund in der Ausgabe; alles Go, kein neues Shell-Skript | erfüllt | `faelle.go` `testFaelle()` (Fälle 25 bis 28, 33) und `buildFaelle()` (29 bis 32); Namen stimmen mit `SPEC-102` und mit `TestFaelleDerPruefungHabenDieNamenAusDemVertrag` überein; `WantText` nennt je Verstoß-Fall den Grund (`nicht gofmt-formatiert`+Dateiname, `go vet:`, `go test:`, `nicht mit -trimpath gebaut`, `Version des Binary`+beide Versionen, `Binary fehlt`, `Aufruf:`); `git diff --name-status` zeigt keine neue `.sh`; `gegenproben-eingabe.sh` ist nur geändert |
| `make gegenproben` grün im Zeitbudget; M1 *erreicht* | **offen**: grün ja, Zeitbudget unbelegt, M1 steht noch auf *offen* | Stufe frisch gefahren (`--no-cache-filter gegenproben`, 33 Fälle, 0 rot): 49,1 s in der Stufe (parallel zu meinen Mutationsläufen), danach zweimal 34,6 s und 35,4 s in der Stufe, 47,5 s und 48,0 s Wanduhr bei Lastmittel 20 bis 44 auf 20 Kernen. Nie unter den 30 s aus `SPEC-102`; eine ruhige Maschine stand nicht zur Verfügung. Roadmap M1: „offen — Gegenproben zu `make build` und `make test` fehlen" |
| `make gates` grün | erfüllt | `make gates` selbst gefahren, Exit 0 (Build, Test, Lint, `a-check`, `docs-check`, `schema-check`, `migration-check` ohne Fehler) |
| Review mit Report unter `docs/reviews/` | erfüllt | Review-Report vom 2026-10-20 liegt vor (Rollenwechsel: anderer Kontext) |
| Doku-Update bei berührtem Vertrag | erfüllt | `SPEC-036`, `SPEC-102`, `harness/README.md` (Sensors, „Nicht behauptet"), `gegenproben.mk` nachgezogen |
| Closure-Notiz mit Lerneintrag | offen (Closure-Arbeit) | Plan §7: „Wird bei der Closure geschrieben." |
| Beobachtungs-Register fortgeschrieben | offen (Closure-Arbeit) | kein Eintrag im Diff; Anlass siehe V9 |
| Jedes Risiko aus §6 trägt einen Ausgang | offen | Risiken 1 und 3: „wird bei der Closure eingetragen"; Risiken 2 und 5: „weiter offen" (Zeit, Bau in der Kopie); Risiko 4: *entfallen* |
| Drei Paarungen getragen | offen (von der nächsten Welle-Closure geprüft) | nicht Gegenstand dieses Laufs |

**Zeitbudget, Ehrlichkeit.** Plan §6, `SPEC-102` §Zeitbudget und Review F10 sagen übereinstimmend, dass das Budget für 33 Fälle nicht belegt ist; kein Code lockert das Budget, kein Text behauptet die Einhaltung. Meine Messungen bestätigen die Größenordnung des Implementers (34,6 bis 49,1 s in der Stufe). „Weiter offen" ist vertretbar, weil die frühere Messung der 24 Fälle bei ruhiger Maschine 13,5 s ergab und die jetzige Last die Zeit mindestens verdreifacht; „eingetreten" wäre erst mit einer ruhigen Messung belegt. Bewertung: ehrlich. Die DoD-Zeile bleibt trotzdem unerfüllt, bis der Planner entscheidet: messen auf ruhiger Maschine, oder das Budget im Vertrag ändern (ein Zeitbudget ist keine Gate-Schwelle nach `AGENTS.md` §3.6, aber eine Vertragszusage ohne Fänger, siehe V4).

## 2. Zusage · Fänger gegen den Vertrag

Jede Zeile des Vertrags gegen die Tabelle in Plan §3; „Zeile" = steht in der Tabelle, „Fänger" = ein Test oder Fall wird bei Bruch rot.

| Vertragsstelle | Zeile | Fänger gesehen |
|---|---|---|
| SPEC-102 Fälle 25, 26, 27, 28, 29, 30, 31, 32, 33 | ja, je eine | ja (Mutationen unten M01 bis M06, M23) |
| SPEC-036 `pruefung test`: gofmt-, vet-, Test-Befund; Exit 0 ohne Befund | ja | ja |
| SPEC-036 `pruefung build`: `-trimpath`, Version, Exit 0 ohne Befund | ja | ja |
| SPEC-036 Randformen 1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 13, 14 | ja | ja (Unit, teils Fall) |
| SPEC-036 Randformen 9 und 12 | **nein** | nein, vom Vertrag so gesagt (9: „nicht von einer Gegenprobe gefangen"; 12: Nicht-Zusage) |
| SPEC-036 „hält beim ersten Befund an", `gofmt:`-Befund bei Exit ungleich 0, Vorgabe `./...` ohne Muster | **nein** | ja, im Unit-Test (`TestTestBefundHaeltAnDerStelle`, `TestTestOhneBefund`), nur ohne Tabellenzeile |
| SPEC-036 Stufe `pruefung` prüft sich mit `gofmt`, `go vet`, `go test` | **nein** | **nein** (V3) |
| SPEC-036/SPEC-102 Stufen `test` und `build` rufen `pruefung` | ja (letzte Zeile, „kein Fall") | nein, offengelegt (V2) |
| SPEC-102 Randform 2 (`pruefung`, `go`, `gofmt` aufrufbar) | **nein** | **nein** (V5) |
| SPEC-102 Randform 2a (Bau scheitert) | **nein** | ja, `TestFallMitBauMeldetScheiternDesBaus` (Mutation G2c rot), nur ohne Tabellenzeile |
| SPEC-102 Eingaben nennen `tools/pruefung` | ja (nur Existenz) | Existenz ja, Prüfsumme nein (V1) |
| SPEC-102 Zeitbudget | nein | nein (V4) |

### Eigene Mutationen (20 am Programm, 6 an den Gegenproben, 4 am Dockerfile, 2 am Eingabe-Skript)

Spalten: „Unit" = Tests in `tools/pruefung` (bei G-Zeilen: in `tools/gegenproben`), „Fälle" = die neun Gegenproben-Fälle gegen das mutierte Binary. „rot (Fall)" nennt den Fall, der rot wurde.

| Nr. | Zusage / Zeile | Mutation (gefahren) | Unit | Fälle |
|---|---|---|---|---|
| M01 | gofmt-Befund | `strings.TrimSpace(l.stdout) != ""` durch `false` | rot | rot (`test/nicht-formatiert`) |
| M02 | vet-Befund | `go vet`-Schritt entfernt | rot | rot (`test/vet-befund`, Text dann `go test:` statt `go vet:`; `go test` fängt Printf selbst, der Fall prüft den Wortlaut `go vet:` nicht gegen diesen Fall hinaus, siehe INFO I1) |
| M03 | roter Test | `go test`-Schritt entfernt | rot | rot (`test/roter-test`, „Verstoß nicht gemeldet") |
| M04 | `-trimpath`-Prüfung | durch `false` ersetzt | rot (`TestBuildOhneTrimpath`) | rot (`build/ohne-trimpath`) |
| M05 | Versionsprüfung | `ist != soll` durch `false` | rot (`TestBuildVersionBefund`, `TestStroeme`) | rot (`build/falsche-version`) |
| M06 | Binary fehlt, Exit 2 | `return exitAufruf` zu `exitBefund` im Zweig | rot (`TestBuildBinaryFehlt`) | rot (`build/binary-fehlt`) |
| M07 | falsche Argumentzahl `build` (Plan-Zeile) | `len(args) == 3` durch `true` | rot (`TestAufrufFehlerStartenNichts`) | **grün** (wie die Plan-Zeile sagt) |
| M08 | falsche Argumentzahl `test` | `len(args) >= 2` durch `true` | rot (dieselbe) | **grün** (wie gesagt) |
| M09 | Randform 6, Leerraum (F3) | `TrimRight(…, "\n")` zu `TrimSpace(…)` | rot (`TestBuildVersionBefund`) | grün |
| M24 | Randform 6, CRLF (F3) | `TrimRight(…, "\r\n")` | rot (`TestBuildVersionBefund`) | grün |
| M10 | Randform 11, `ok` (build) | `ok`-Zeile nach stderr | rot (`TestBuildOhneBefund`, `TestStroeme`) | grün (Fälle lesen beide Ströme) |
| M28 | Randform 11, `ok` (test) | `ok`-Zeile nach stderr | rot (`TestTestOhneBefund`, `TestStroeme`) | grün |
| M29 | Randform 11, Befund | Befundtext nach stdout | rot (`TestBuildOhneTrimpath`) | grün |
| M11 | `starte` ohne `cmd.Dir` | `cmd.Dir = dir` entfernt | **grün** | rot (3 Fälle) |
| M12 | Muster weitergeben | Muster verworfen | rot (zwei Tests) | rot (2 Fälle) |
| M25 | Randform 13, Strich | Argumente mit `-` verworfen | rot (`TestTestMusterMitStrichGehenUnveraendertAnGo`) | grün |
| M13 | Randform 4, Exit ungleich 0 bei `version` | Zweig entfernt | rot (`TestBuildBinaryStartetNicht`) | grün |
| M14 | Randform 5, kein Go-Binary | Exit-Zweig entfernt | rot (`TestBuildInformationNichtLesbar`, `TestBuildKeinGoBinaryMitEchtemGo`) | grün |
| M15 | Randform 7 (`test`) | Aufrufbarkeitsprüfung entfernt | rot (`TestProgrammNichtAufrufbar`) | grün |
| M17 | Randform 2, Wurzel | `Stat`-Prüfung entfernt | rot (`TestWurzelFehlt`) | grün |
| M18 | Randform 1, leeres Argument | `a == ""` durch `a == "" && false` | rot (`TestAufrufFehlerStartenNichts`) | grün |
| M19 | `go test`-Flags | `-trimpath` aus `go test` entfernt | rot (drei Tests) | grün |
| M21 | Randform 6, stderr | `alles()` statt `stdout` | rot (`TestBuildVersionBefund`) | grün |
| M22 | Randform 14 | `exitAufruf = 1` | **grün** | rot (`test/aufruf`, `build/binary-fehlt`) |
| M23 | Aufruf ohne Argumente (Fall 33) | Zweig `len(args) == 0` entfernt | rot (`TestAufrufFehlerStartenNichts`) | rot (`test/aufruf`) |
| M26 | Randform 10, kein Zustand | Datei in `Wurzel` geschrieben | rot (`TestSchreibtNichts`) | grün |
| M27 | Randform 8 | eigene Prüfung auf Go-Dateien hinzugefügt | rot (`TestTestOhneGoDateienEntscheidenDieProgramme` unter anderen) | grün |
| G2c | Randform 2a | Bau-Fehler nur bei Zeitüberschreitung gemeldet | rot (`TestFallMitBauMeldetScheiternDesBaus`) | — |
| G3 | Bau vor dem Gate | `f.Bau != nil` durch `false` | rot (zwei Tests) | — |
| G4 | `-trimpath` im Bau | immer gesetzt | rot (`TestBauenTrimpathUndVersion`) | — |
| G5 | Version im Bau | Version nicht übergeben | rot (`TestBauenTrimpathUndVersion`) | — |
| **G1** | **SPEC-102 Randform 2** | `"pruefung", "go", "gofmt"` aus `standard().Programme` gestrichen | **grün** | — |
| D1 | `make test` rot bei Verstoß | unformatierte Datei in `src/bootstrap/`, `docker build --target test` | Exit 1, Text `nicht gofmt-formatiert:` und Dateiname | — |
| D3 | `make build` rot bei Versionsabweichung | `-X main.version=fest` im Dockerfile, `VERSION=1.2` | Exit 1, Text `Version des Binary ist 'odata-recorder fest', erwartet 'odata-recorder 1.2'` | — |
| **D2** | **Stufe `test` ruft `pruefung`** | `RUN … pruefung test /src` durch `RUN true`, dazu unformatierte Datei | **Exit 0, nichts rot** (V2) | — |
| **D6** | **Stufe `pruefung` prüft sich selbst** | `go vet ./... && go test ./...` aus der Stufe entfernt, dazu ein roter Test in `tools/pruefung` | **Exit 0, nichts rot** (V3) | — |
| F1a | Eingaben nennen `tools/pruefung` (Existenz) | `tools/pruefung` in der Wegwerfkopie gelöscht | `gegenproben-eingabe.sh` Exit 1, `Eingabe tools/pruefung fehlt` | — |
| **F1b** | **Prüfsumme erfasst `tools/pruefung`** | Eintrag aus dem Array `EINGABEN` gestrichen | Prüfsumme bei Änderung an `main.go` **gleich**; das Skript endet auch bei fehlendem Verzeichnis mit Exit 0 (V1) | — |

Die Tabellenzeilen des Plans, die „Mutation: …" nennen, wurden bis auf die Dockerfile-Verdrahtung rot gesehen (M01 bis M07, M08 bis M10, M13 bis M15, M17, M23 bis M27). Die Plan-Aussagen „fängt nur den Aufruf ohne Argumente" (Fall `test/aufruf`) und „Rest fängt `TestAufrufFehlerStartenNichts`" sind genau (M07, M08, M23).

## 3. Plan-vs-Code (`AGENTS.md` §3.9)

* §3 Dateiliste und Änderungsarten stimmen mit dem Diff überein: `tools/pruefung/` neu; `Dockerfile` (Stufe `pruefung`, Aufrufe in `test`/`build`, Lint-Stufe kopiert `tools/pruefung`, Stufe `gegenproben` bekommt `pruefung`), `go-gates.mk` (zwei Prüfzeilen entfallen), `.dockerignore` (`!tools/pruefung`), `.golangci.yml` (Pfad `^tools/(gegenproben|pruefung)/main\.go$`, `# Why:` angepasst, keine Schwelle gesenkt), `gegenproben-eingabe.sh` (Array), `tools/gegenproben` (Fälle, `Bau`, `Programme`), Verträge, `harness/README.md`, `gegenproben.mk`.
* Kopf (`Bezug`, `Berührte Spec-Stellen`) stimmt. §1 stimmt bis auf V7. §6 trägt die Ausgänge ehrlich (zwei „weiter offen", einer *entfallen*, zwei „bei der Closure").
* §3 „Zusage · Fänger": Lücken siehe Abschnitt 2 und V6.
* Randformen vor dem Code (`AGENTS.md` §3.12): Nr. 4 (Exit ungleich 0), Nr. 6 (CRLF), Nr. 11 bis 13 und SPEC-102 Randform 2a wurden **nach** den Code-Commits (a7ad164 bis 0513356) in Commits bd98a06 und ff9c8c2 entschieden. Der Weg war regelkonform (Review fand sie, Planner entschied, Plan §6 hält das für Nr. 4 fest), aber es ist erneut „Randform erst im Review entschieden" (V9).
* Kommentare (§3.7, §3.11): keine verworfene Alternative, kein abwesender Text, keine Chronik in Kopf- und Hilfetexten gefunden. Kopf von `tools/gegenproben/main.go` sagt ausdrücklich, dass die Verdrahtung im `Dockerfile` nicht geprüft wird; `SPEC-102` §Grenze und `harness/README.md` §Nicht behauptet nennen sie ebenso. Die Zusage im Kommentar der Stufe `pruefung` („Es wird geprüft (gofmt, vet, Test) und gebaut", `SPEC-036`: „prüft es mit gofmt, go vet und go test") nennt eine Prüfung, die kein Test und kein Fall fängt (V3); der Text ist wahr, aber nicht abgesichert.

## 4. Verhaltensgleichheit von `make test` und `make build` gegenüber `2091ba6`

| Aspekt | Vorher | Jetzt | Urteil |
|---|---|---|---|
| `make test` | `gofmt -l .`, `go vet ./...`, `go test -trimpath -buildvcs=false ./...` in `/src` | `pruefung test /src`: dieselben drei Aufrufe, Reihenfolge, Flags; Text `nicht gofmt-formatiert:` erhalten; Exit 1 | gleich (D1 rot mit demselben Text) |
| `make build` Bau | `go build -trimpath -buildvcs=false -ldflags=… -o /out/odata-recorder` mit `TARGETOS/TARGETARCH` | unverändert | gleich |
| `make build` Prüfung | zwei Zeilen in `go-gates.mk` nach dem `docker build`, gegen `docker run` des Images | `pruefung build` in der Stufe `build`, unmittelbar nach `go build` | gleich oder strenger: ein Verstoß verhindert jetzt das Image (D3 rot); ein Cross-Build bleibt wie vorher nur prüfbar, wenn das Binary auf dem Bau-Host startet |
| `VERSION=` (leer) | `docker build` Exit 0, Binary druckt `odata-recorder ` , Vergleich besteht, `make build` grün (D5) | `pruefung build` meldet Aufruffehler, `make build` Exit 2 (D4) | **abweichend, strenger** (V7) |
| Meldungstexte | `build: Version des Binary ist …` | `Version des Binary ist …` | geändert, vom Plan §1 erlaubt |

Das Ergebnis: kein Gate ist geschwächt; eine Verschärfung (leere `VERSION`) steht im Vertrag (`SPEC-036` Randform 1), nicht im Plan.

## 5. Stand der Review-Befunde

| Befund | Erwartung des Auftrags | Verifiziert |
|---|---|---|
| F1 Prüfsumme/`tools/pruefung` | behoben | **nur offengelegt, nicht behoben**: Existenz wird geprüft (F1a), aber die Liste ist jetzt *ein* Array; wer `tools/pruefung` streicht, verliert Existenzprüfung *und* Prüfsumme zugleich, nichts wird rot (F1b). Plan-Zeile und `SPEC-102` §Grenze sagen das ehrlich („fängt nichts"). Bleibt MEDIUM nach §3.15 (V1) |
| F2 Verdrahtung `test`/`build` | ehrlich offen | bestätigt (D2); Plan §3 letzte Zeile, `SPEC-102` §Grenze, `harness/README.md` §Nicht behauptet nennen es |
| F3 Leerraum/CRLF | behoben | behoben: `SPEC-036` Randform 6 entscheidet, `TestBuildVersionBefund` hat drei neue Fälle, M09 und M24 rot |
| F4 Randformen im Vertrag | behoben | behoben im Vertrag (`SPEC-102` Randform 2 und 2a); die Tabelle trägt dazu keine Zeile und `standard().Programme` ist ungepinnt (V5, V6) |
| F5 Hilfetext und Kopf | behoben | behoben: `gegenproben.mk` (Kopf und `##`-Text) und Paketkommentar von `tools/gegenproben/main.go` nennen `pruefung` |
| F6 `test/aufruf` | behoben | behoben durch genaue Tabellenzeile (Fall = nur null Argumente; Rest `TestAufrufFehlerStartenNichts`); M07, M08 bestätigen die Aufteilung. Der Fall selbst deckt weiter nur „keine Argumente" |
| F7, F8 | INFO | als Randform 12 und 13 in `SPEC-036` entschieden; Randform 13 durch Test gefangen (M25) |
| F9 Streams | behoben | behoben: `SPEC-036` Randform 11, `TestStroeme` (M10, M28, M29 rot) |
| F10 Marke | INFO | Zahl stimmt nicht (V8) |

**Marken-Satz in Plan §4.** Gemessen: `git diff --numstat 2091ba6..HEAD` über die Pfade aus Plan §3 plus `.golangci.yml` ergibt 916 Einfügungen und 62 Löschungen, **Summe 978**; der Plan sagt 974. Die Differenz von 4 entspricht genau `harness/mk/gegenproben.mk` (2 + 2), das §3 nennt. Die Marke (über 900) ist überschritten, der Satz „Marke überschritten" stimmt; die Zahl nicht.

## Findings

### V1 MEDIUM — Prüfsumme erfasst `tools/pruefung` ohne Fänger (F1 nicht behoben)
* kategorie: MEDIUM
* quelle: `AGENTS.md` §3.15; `SPEC-102` §Eingaben und §Grenze
* pfad: `tools/harness/gegenproben-eingabe.sh` · `EINGABEN=(src go.mod go.sum … tools/gegenproben tools/pruefung Dockerfile .dockerignore)`
* befund: Die Zusage „`tools/pruefung/` gehört zu den Eingaben der Prüfsumme" hat keinen Fänger; Plan und Vertrag sagen das ausdrücklich. Seit dem Array ist die Lage schlechter als beim Review: ein Eintrag fehlt in der Existenzprüfung und in `git ls-files` zugleich.
* Fehlerfall (gefahren): `tools/pruefung` aus `EINGABEN` gestrichen. Eine Änderung an `tools/pruefung/main.go` lässt die Prüfsumme gleich (beide Läufe `25c74f28…`), `make gegenproben` meldet „unverändert" und fährt die Fälle 25 bis 33 nicht; ein gelöschtes Verzeichnis `tools/pruefung` lässt das Skript mit Exit 0 enden.
* Ehrlichkeit: offengelegt, kein Verstecken. Handlung beim Planner: Fänger bauen (etwa eine Gegenprobe des Skripts) oder die Lücke als bewusst offen im Register führen.
* verifizierbar: ja

### V2 MEDIUM — Verdrahtung der Stufen `test` und `build` ohne Fänger (F2, offengelegt)
* kategorie: MEDIUM
* quelle: `AGENTS.md` §3.15, §3.10; Plan §3 letzte Zeile
* pfad: `Dockerfile` · `RUN --network=none pruefung test /src`, `RUN --network=none pruefung build /out/odata-recorder "$VERSION"`
* Fehlerfall (gefahren, D2): erste Zeile durch `RUN true` ersetzt, unformatierte Datei in `src/bootstrap/`: `docker build --target test` endet mit Exit 0. Die Gegenrichtung ist gefangen (D1, D3).
* Ehrlichkeit: Plan §3, `SPEC-102` §Grenze und `harness/README.md` §Nicht behauptet sagen es.
* verifizierbar: ja

### V3 MEDIUM — Stufe `pruefung` prüft sich selbst, kein Fänger, nicht offengelegt
* kategorie: MEDIUM
* quelle: `AGENTS.md` §3.15, §3.11; `SPEC-036` („das `Dockerfile` baut es in einer eigenen Stufe, prüft es mit `gofmt`, `go vet` und `go test`")
* pfad: `Dockerfile` · `FROM deps AS pruefung … && go vet ./... && go test ./... && go build …`
* befund: Die Zusage steht im Vertrag und im Stufenkommentar; Plan §3 hat keine Zeile, `SPEC-102` §Grenze nennt sie nicht (sie nennt nur die Stufe `gegenproben`).
* Fehlerfall (gefahren, D6): `go vet ./... && go test ./...` aus der Stufe entfernt und ein roter Test in `tools/pruefung` hinzugefügt: `docker build --target test` endet mit Exit 0. Auch `make lint` (zweites Modul) fängt nur Lint, nicht den roten Test; die Unit-Tests von `pruefung` laufen dann nirgends.
* verifizierbar: ja

### V4 MEDIUM — Zeitbudget 30 s: Vertragszusage ohne Beleg und ohne Fänger, DoD-Zeile unerfüllt
* kategorie: MEDIUM
* quelle: Plan §2 Zeile 3, §6 Risiko 4; `SPEC-102` §Zeitbudget
* pfad: `harness/vertraege/SPEC-102-gegenproben.md` · „höchstens 30 s … die Einhaltung des Budgets ist für 33 Fälle nicht belegt"
* befund: Der Vertrag führt weiter die Zahl 30 s, widerlegt sie aber in derselben Zeile durch die Messung (38,5 s). Meine drei Messungen (34,6 bis 49,1 s in der Stufe, unter fremder Last) liegen alle darüber. Kein Test und kein Gate prüft das Budget.
* Ehrlichkeit: gut. Plan, Vertrag und Review benennen die Lücke gleichlautend; „weiter offen" ist ohne ruhige Messung vertretbar (siehe Abschnitt 1).
* Handlung beim Planner: ruhige Messung beschaffen oder das Budget im Vertrag ändern; der Slice kann nicht nach `done/`, solange die DoD-Zeile so steht.
* verifizierbar: ja (Messung)

### V5 MEDIUM — `SPEC-102` Randform 2 für `pruefung`, `go`, `gofmt` ohne Fänger und ohne Tabellenzeile
* kategorie: MEDIUM (geringe Wirkung: ein fehlendes Programm macht die Fälle trotzdem einzeln rot)
* quelle: `AGENTS.md` §3.15; `SPEC-102` Randform 2
* pfad: `tools/gegenproben/main.go` · `Programme: []string{"a-check", "golangci-lint", "sh", "pruefung", "go", "gofmt"}`
* Fehlerfall (gefahren, G1): `"pruefung", "go", "gofmt"` aus `standard().Programme` gestrichen: alle Unit-Tests von `tools/gegenproben` grün. `umgebungMit` setzt eigene `Programme`; die Liste von `standard()` pinnt kein Test.
* verifizierbar: ja

### V6 LOW — Tabelle „Zusage · Fänger" unvollständig, Fänger existieren aber
* kategorie: LOW
* quelle: `AGENTS.md` §3.15 („Jede Zusage, Randform und jeder Fall des Vertrags hat eine Zeile")
* pfad: Plan §3 Tabelle
* befund: Es fehlen Zeilen für `SPEC-102` Randform 2a (Fänger vorhanden, G2c rot), „hält beim ersten Befund an", `gofmt:`-Befund bei gofmt-Fehler und Vorgabe `./...` (Fänger im Unit-Test vorhanden) sowie für `SPEC-036` Randformen 9 und 12 (vom Vertrag als nicht gefangen bzw. Nicht-Zusage benannt; eine Zeile „kein Fänger, siehe Vertrag" wäre die vollständige Form). Die Tabelle ist enger als der Vertrag, der Code nicht.
* verifizierbar: ja (Textvergleich)

### V7 LOW — Leere `VERSION` wird jetzt abgelehnt; Plan §1 sagt „Wirkung unverändert"
* kategorie: LOW
* quelle: Plan §1 („ändert aber ihre Wirkung nicht"), Plan §2 Zeile 1; `SPEC-036` Randform 1
* Fehlerfall (gefahren, D4/D5): `--build-arg VERSION=`: vorher Exit 0 und grünes `make build` (Binary druckt `odata-recorder `), jetzt Exit 2 in der Stufe `build`. Strenger, nicht schwächer, im Vertrag entschieden; Plan §1 und DoD nennen die Abweichung nicht.
* verifizierbar: ja

### V8 LOW — Marken-Zahl in Plan §4 stimmt nicht
* kategorie: LOW
* quelle: `AGENTS.md` §3.14, §3.7 (Zustandsfeld)
* pfad: Plan §4 · „974 Zeilen (Summe aus `git diff --numstat 2091ba6..HEAD` über die Pfade aus §3 plus `.golangci.yml`)"
* befund: Nach der genannten Metrik 978 (916 + 62); 974 ergibt sich nur ohne `harness/mk/gegenproben.mk`. Die Aussage „überschritten" stimmt, die Zahl ist um 4 zu niedrig.
* verifizierbar: ja (`git diff --numstat 2091ba6..HEAD -- <Pfade aus §3> .golangci.yml`)

### V9 LOW — Randformen wieder erst nach dem Code entschieden (Register-Evidenz)
* kategorie: LOW
* quelle: `AGENTS.md` §3.12, §3.16; `BEO-REPO/randform-erst-im-review-entschieden`
* befund: Fünf Randformen (`SPEC-036` Nr. 4 Exit ungleich 0, Nr. 6 CRLF, Nr. 11 bis 13, `SPEC-102` 2a) kamen erst durch das Review; der Weg über Planner ist eingehalten. Der Slice sollte beim Fortschreiben des Registers diese Evidenz in der vorhandenen Beobachtung ablegen (DoD „Register fortgeschrieben").
* verifizierbar: nein (Prozess)

### I1 INFO — Fall `test/vet-befund` unterscheidet `go vet:` und `go test:` nicht
Bei entferntem `go vet`-Schritt (M02) bleibt der Fall rot, weil `go test` das Printf selbst als Build-Fehler meldet; der erwartete Text `go vet:` fehlt dann, deshalb ist der Fall *aus dem benannten Grund* rot. Kein Befund, nur Rot-Grund bestätigt.

### I2 INFO — Rückführungs-Marke selbst gewährt
Plan §4 stellt die Überschreitung fest und begründet „bleibt ungeteilt" im Plan selbst; die Entscheidung gehört dem Planner. Bestätigung des Planners fehlt im Artefakt.

### I3 INFO — Offene Closure-Arbeit
Closure-Notiz, Register, Risiko-Ausgänge 1 und 3, M1 *erreicht* und die Paarungen fehlen planmäßig. Belege für die Ausgänge, die ich liefern kann: Risiko 3 („`pruefung` ändert das Verhalten der Gates unbemerkt"): `make gates` grün, Mutationen je Zusage siehe Abschnitt 2, roter Verstoß je Ziel (D1 für `test`, D3 für `build`); einzige Verhaltensabweichung V7.

## Geprüft, ohne Befund

* `make gates`: Exit 0.
* `make gegenproben` (Stufe frisch gefahren): 33 Fälle, 0 rot; der Stempel `harness/gegenproben.stempel` stimmt mit der Prüfsumme des aktuellen Stands überein (`70c82a8f…`), nichts wurde geschrieben.
* `harness/README.md`, `harness/mk/go-gates.mk`, `harness/mk/gegenproben.mk`, `.dockerignore`, `.golangci.yml`: stimmen mit Plan §3 und Code überein.
* `tools/pruefung/main.go` gegen `SPEC-036`: Texte, Exit-Codes und Randformen 1 bis 8, 10, 11, 13, 14 stimmen (Mutationen oben); Randform 9 und 12 sind im Code, wie der Vertrag sagt (kein Zeitlimit, kein Zustand).
* Zeit: Mutationen an `pruefung` haben je mindestens einen Fänger, außer den in V1 bis V3 und V5 genannten; kein Fall der neun blieb bei einer Mutation grün, die der Plan als „Fall" benennt.

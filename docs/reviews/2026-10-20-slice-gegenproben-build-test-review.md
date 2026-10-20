# Review: slice-gegenproben-build-test

* Datum: 2026-10-20
* Rolle: reviewer (frischer Kontext, kein Self-Review)
* Gegenstand: `git diff 2091ba6^..HEAD` (Commits a7ad164 bis bd98a06)
* Plan: `docs/plan/planning/in-progress/slice-gegenproben-build-test.md`
* Verträge: `harness/vertraege/SPEC-036-test.md`, `harness/vertraege/SPEC-102-gegenproben.md`
* Maßstab: Plan, ADR-0001, ADR-0003, ADR-0005, AGENTS.md §3.1 bis §3.16, Reviewer-Skill. Nicht geprüft: DoD (Verifier).
* Methode: Wegwerfkopien von `git archive HEAD` im Scratchpad, Docker `golang:1.27-alpine` (`--network none`), nie im Arbeitsbaum. Je Mutation liefen die Unit-Tests von `tools/pruefung` und die zehn Fälle 25 bis 33 gegen das mutierte `pruefung`-Binary (Treiber: eine zusätzliche Testdatei nur in der Kopie). Dockerfile-Mutationen mit echtem `docker build --target test|build`.

## Zusammenfassung

Keine HIGH. Zwei MEDIUM (Zusage ohne Fänger), vier LOW, vier INFO.
`pruefung` ist gegen die Plan-Tabelle gut gefangen: von rund 35 Mutationen am Programm blieben nur die in F3 und F10 genannten (plus zwei äquivalente) in allen Spalten grün, die Fälle 25 bis 33 fangen jeweils die Mutation, die die Plan-Tabelle benennt.
Verhaltensgleichheit der Gates zum Stand vor dem Slice: nichts geschwächt (siehe unten).

## Findings

### F1 MEDIUM — Prüfsumme nennt `tools/pruefung` in keiner Weise gefangen
* kategorie: MEDIUM
* quelle: AGENTS.md §3.15 (Zusage ohne Fänger), SPEC-102 §Eingaben
* pfad: `tools/harness/gegenproben-eingabe.sh` · `git ls-files -z --cached --others --exclude-standard`
* befund: SPEC-102 sagt zu, dass `tools/pruefung/` zu den Eingaben der Prüfsumme gehört, und Plan §3 nennt es als eigene Zeile; die Zusage-Fänger-Tabelle führt dafür keine Zeile, und kein Test oder Fall fängt es.
* Fehlerfall (Mutation gefahren): `tools/pruefung` aus der zweiten Liste (der `git ls-files`-Liste) gestrichen. Eine Änderung an `tools/pruefung/main.go` lässt die Prüfsumme gleich (beide Läufe `c95a75ee…`-Klasse, im Scratch-Git verglichen), `make gegenproben` meldet „unverändert" und fährt die Fälle 25 bis 33 nicht, obwohl `pruefung` sich geändert hat. Die erste Liste (Existenzprüfung) bleibt grün.
* verifizierbar: ja (Skript zweimal mit geänderter Datei aufrufen)
* klasse: Zusage des Vertrags ohne Fänger in der Zusage-Fänger-Tabelle

### F2 MEDIUM — Verdrahtung der Stufen `test` und `build` ohne Fänger (offengelegt)
* kategorie: MEDIUM
* quelle: AGENTS.md §3.15, §3.10; Plan §3 letzte Zeile der Tabelle; SPEC-102 §Grenze
* pfad: `Dockerfile` · `RUN --network=none pruefung test /src` und `RUN --network=none pruefung build /out/odata-recorder "$VERSION"`
* befund: Die Zusage „die Stufen rufen `pruefung`" hat als Fänger nur einen Beleg im Bericht des Implementers, keinen Test, Fall oder benannte Gate-Ausgabe.
* Fehlerfall (Mutation gefahren): (A) `RUN … pruefung test /src` durch `RUN true` ersetzt, dazu eine unformatierte Datei in `src/bootstrap/`: `docker build --target test` endet mit Exit 0. (D) `RUN … pruefung build …` durch `RUN true` ersetzt: `--target build` endet mit Exit 0. Auch ein anderes Muster oder eine andere Wurzel in der Zeile bliebe unbemerkt, soweit die Wurzel existiert.
* Einordnung: `harness/README.md` §Nicht behauptet und SPEC-102 §Grenze benennen die Lücke ehrlich; das ändert die Klasse nicht (§3.15: „ein MEDIUM"). Die Gegenrichtung ist gefangen: roter Verstoß im Quellstand (unformatierte Datei) lässt `--target test` rot werden (Exit 1, Text `nicht gofmt-formatiert`); Mutationen B (`-trimpath` aus `go build` entfernt) und C (`main.version=dev` fest) lassen `--target build` mit dem erwarteten Text rot werden.
* verifizierbar: ja
* klasse: Zusage ohne Fänger (Verdrahtung nur im Dockerfile)

### F3 LOW — Versionsvergleich normalisiert mehr als der Vertrag sagt, kein Fänger
* kategorie: LOW (Konventions-Anker: AGENTS.md §3.15; wäre MEDIUM nach Tabelle, hier ohne Auswirkung auf die heutigen Binaries)
* quelle: SPEC-036 `pruefung build` („nach Entfernen der abschließenden Zeilenumbrüche … genau") und Randform 6
* pfad: `tools/pruefung/main.go` · `ist := strings.TrimRight(l.stdout, "\n")`
* befund: Der Vertrag sagt „genau `odata-recorder <Version>`", Tests prüfen weder führende noch abschließende Leerzeichen noch `\r\n`.
* Fehlerfall (Mutation gefahren): `TrimRight(…, "\n")` zu `strings.TrimSpace(…)` — Unit-Tests grün, alle neun Fälle grün; ein Binary, das `odata-recorder 1.2.3 ` oder ` odata-recorder 1.2.3` ausgibt, besteht. Ebenso `"\r\n"`: grün (Randform „CRLF" im Vertrag nicht entschieden).
* verifizierbar: ja
* klasse: Randform des Vertrags ohne Fänger

### F4 LOW — Randformen der Gegenproben für den neuen Fall-Mechanismus nicht im Vertrag
* kategorie: LOW (Anker: AGENTS.md §3.12/§3.16)
* quelle: SPEC-102 §Randformen 2 und 5
* pfad: `harness/vertraege/SPEC-102-gegenproben.md` · „Ein Gate-Programm (`a-check`, `golangci-lint`, `sh`) ist nicht aufrufbar" ; Code `tools/gegenproben/main.go` · `Programme:  []string{"a-check", "golangci-lint", "sh", "pruefung", "go", "gofmt"}` und `"Bau des Binary scheitert"`
* befund: Der Code prüft jetzt auch `pruefung`, `go`, `gofmt` auf Aufrufbarkeit und meldet ein gescheitertes `go build` eines Falls mit Bau als roten Fall („Bau des Binary scheitert"); beide Entscheidungen stehen nicht in den Randformen von SPEC-102 (Randform 2 nennt drei Programme).
* Fehlerfall: Fehlt `pruefung` im Image, entscheidet die Meldung `gegenproben: pruefung nicht aufrufbar` / Exit 2 allein der Code; ein Leser des Vertrags erwartet sie nicht. Das Verhalten ist durch `TestFallMitBauMeldetScheiternDesBaus` gefangen (Code-seitig sauber), nur die Spec zieht nicht nach.
* verifizierbar: nein (Spec-Text)
* klasse: Code entscheidet Randform still

### F5 LOW — Hilfetext und Kopf nennen die Gegenproben zu `pruefung` nicht
* kategorie: LOW (Anker: AGENTS.md §3.7, §3.9)
* quelle: Maintainability
* pfad: `harness/mk/gegenproben.mk` · `gegenproben: ## Gegenproben zu a-check, lint und migration-check, nur bei geänderter Eingabe`; ebenso der Kopfkommentar derselben Datei; `tools/gegenproben/main.go` · `// Command gegenproben weist an a-check und an tools/harness/lint.sh nach`
* befund: `make help` und die Paketbeschreibung nennen die Gate-Gruppe, die `harness/README.md` und SPEC-102 jetzt um `pruefung` erweitern; der Slice zieht die Plan-Zeilen nach, diese beiden Stellen nicht.
* Fehlerfall: Ein Leser von `make help` schließt, `make gegenproben` fahre nichts zu `make test`/`make build` (die Paketbeschreibung war schon vor dem Slice ohne `migration-check.sh`).
* verifizierbar: ja (Text)
* klasse: Beschreibung driftet hinter dem Vertrag

### F6 LOW — Fall `test/aufruf` deckt nur „keine Argumente", die Tabelle sagt „Argumentzahl"
* kategorie: LOW
* quelle: AGENTS.md §3.15 (Plan §3: „Mutation: Prüfung der Argumentzahl entfernt")
* pfad: `tools/gegenproben/faelle.go` · `Name: "test/aufruf", Gate: gatePruefung,`
* befund: Der Fall ruft `pruefung` ohne Argumente. Die Mutationen `len(args) >= 2` → `true` (test) und `len(args) == 3` → `>= 3` (build) bleiben für alle neun Fälle grün; nur die Unit-Tests fangen sie. Die Tabellenzeile benennt den Fall als Fänger.
* Fehlerfall: Mutation `aufrufGueltig` für `build`: Gegenproben melden `ok test/aufruf`, Unit-Test `TestAufrufFehlerStartenNichts` wird rot (dort gefangen, Tabelle benennt ihn nicht).
* verifizierbar: ja
* klasse: Fänger in der Tabelle enger als die Zusage

### F7 INFO — Kein Zeitlimit für das Binary bei `version`
* kategorie: INFO (Randform „Abbruch/unbekannt" von `pruefung build`)
* pfad: `tools/pruefung/main.go` · `l, err := starte("", binary, "version")` (`exec.CommandContext(context.Background()`)
* befund: Ein Binary, das bei `version` nicht endet, hält `pruefung` und damit den Docker-Build an; Vertrag und Raster entscheiden diese Randform nicht (Raster „Abbruch" nennt nur Nr. 10). Für den heutigen Quellstand ohne Folge. Gleiches gilt für `go vet`/`go test`.
* Hinweis an Planner/Architect: Randform vor Vertragserweiterung entscheiden (§3.12).

### F8 INFO — Muster mit führendem `-` werden an `go` durchgereicht
* kategorie: INFO
* pfad: `tools/pruefung/main.go` · `append([]string{"vet"}, muster...)`
* befund: `pruefung test /src -run=X` übergibt das Argument als Flag; SPEC-036 nennt „Muster", entscheidet das nicht. Aufrufer ist heute nur das Dockerfile.

### F9 INFO — Befunde gehen nach stderr, Erfolg nach stdout; im Vertrag nicht benannt
* kategorie: INFO
* pfad: `tools/pruefung/main.go` · `_, _ = fmt.Fprintln(stderr, "das Binary ist nicht mit -trimpath gebaut")`
* befund: SPEC-036 nennt Texte und Exit-Codes, nicht den Strom (außer „nur stdout zählt" für `version`). Die Trennung ist in den Unit-Tests gepinnt (Mutationen `ok`→stderr, Befund→stdout, Aufruf→stdout: rot); die Gegenproben fangen sie nicht, weil sie beide Ströme zusammen lesen (`starte` in `tools/gegenproben/main.go`). Der Vertrag lässt das offen.

### F10 INFO — Einzelheiten ohne Handlung
* Rückführungs-Marke (Plan §4, „über 900 Zeilen", Summe aus `git diff --numstat` über die Pfade aus §3): gemessen 898 (844 Einfügungen, 54 Löschungen, `2091ba6..HEAD`) gegen Schätzung 600; die Marke ist nicht erreicht, aber knapp unterschritten (§3.14).
* `VERSION=` leer: vorher bestand der Vergleich `"odata-recorder "` gegen `"odata-recorder "`, jetzt endet `make build` mit Exit 2 (Randform „leer" von `pruefung build`, in SPEC-036 Nr. 1 entschieden). Strenger, nicht schwächer.
* In `pruefeBuild` ist `nichtAufrufbar(stderr, u.Go)` vor dem Versionsaufruf äquivalent zur späteren Meldung bei `starte`-Fehler (Mutation „Prüfung entfernt": alle Tests grün, gleiche Meldung und Exit 2). Der Plan nennt für Nr. 7 beim Build eine Mutation; sie ist nicht unterscheidbar.
* Die Unit-Tests vergleichen Exit-Codes gegen die Konstanten (`exitAufruf` usw.); die Zahlen 1/2 pinnen allein die Gegenproben (Mutation `exitAufruf = 1`: Unit grün, Fälle `test/aufruf` und `build/binary-fehlt` rot).
* Zeitbudget (SPEC-102): Darstellung ehrlich — Text und Plan §6 nennen 38,5 s unter Last, keine Messung auf ruhiger Maschine, „nicht belegt". Keine Behauptung, die die Messung nicht trägt.

## Auflösung der Tabelle „Zusage · Fänger" (Plan §3) und des Randform-Rasters

Spalte „Unit" = Unit-Tests in `tools/pruefung`, „Fall" = Gegenprobe. „rot" = Mutation wurde in der genannten Spalte rot gesehen.

| Zeile / Zelle | Mutation (gefahren) | Fall | Unit |
|---|---|---|---|
| gofmt-Befund | gofmt-Schritt entfernt | rot `test/nicht-formatiert` | rot |
| vet-Befund | vet-Schritt entfernt | rot `test/vet-befund` (Text `go vet:`; `go test` fängt Printf selbst, der Text unterscheidet) | rot |
| roter Test | test-Schritt entfernt | rot `test/roter-test` | rot |
| `test` ohne Befund Exit 0 | `-trimpath` aus `go test` entfernt | grün (nicht Gegenstand der Fälle) | rot |
| fehlendes `-trimpath` | Prüfung entfernt | rot `build/ohne-trimpath` | rot |
| falsche Version | Prüfung entfernt | rot `build/falsche-version` | rot |
| Binary fehlt Exit 2 | Exit 1 statt 2 | rot `build/binary-fehlt` | rot |
| falscher Aufruf Exit 2 | `len(args)==0`-Zweig, Zahl-Prüfungen | `test/aufruf` rot nur für Null-Argumente (siehe F6) | rot |
| `build` ohne Befund Exit 0 | (Basisfall; Version/Trimpath-Mutationen oben) | `build/basis` grün bei allen Mutationen, die nicht seine Prüfung treffen | rot |
| Nr. 1 leeres Argument | Prüfung entfernt | grün | rot |
| Nr. 2 Wurzel fehlt | Prüfung entfernt | grün | rot |
| Nr. 3 Binary fehlt | siehe oben | rot | rot |
| Nr. 4 Binary startet nicht / Exit ≠ 0 | Exit-Zweig entfernt; Start-Fehler verschluckt | grün | rot |
| Nr. 5 kein Go-Binary | Exit-Zweig von `go version -m` entfernt | grün | rot (inkl. echtes `go`) |
| Nr. 6 Zusatztext / stderr | `HasPrefix`; `alles()` statt `stdout` | grün | rot |
| Nr. 6 Leerraum / CRLF | `TrimSpace`; `"\r\n"` | grün | **grün** (F3) |
| Nr. 7 Programm nicht aufrufbar | gofmt/go-Prüfung (test) entfernt | grün | rot |
| Nr. 7 beim Build | `go`-Prüfung entfernt | grün | grün (äquivalent, F10) |
| Nr. 8 ohne Go-Dateien | eigene Go-Dateien-Prüfung hinzugefügt | rot (`test/*` abgelehnt) | rot |
| Nr. 10 kein Zustand | Datei in `Wurzel` geschrieben | grün | rot |
| Nr. 11 Exit-Codes | `exitAufruf=1` | rot | grün (F10) |
| `starte` ohne `cmd.Dir` | Verzeichnis ignoriert | rot (`test/basis` u. a.) | grün |
| Wurzel-/Muster-Weitergabe | Muster nicht weitergereicht | grün | rot |
| Strom-Trennung | ok → stderr, Befund → stdout, Aufruf → stdout | grün | rot |
| Dockerfile ruft `pruefung` | Aufruf entfernt (A, D) | grün | — (F2) |
| Dockerfile-Flags/Version | `-trimpath` entfernt (B), `main.version` fest (C) | — | `docker build` rot, Text wie erwartet |
| Prüfsumme nennt `tools/pruefung` | aus Liste gestrichen | grün | — (F1) |

Die Mutationen aus dem Bericht des Implementers wurden nicht blind wiederholt; Überschneidungen (gofmt/vet/test, Trimpath, Version, Binary fehlt, Exit) sind oben mit eigener Kopie erneut gefahren.

## Prüfung der Prüfpunkte

* Kommentare §3.7/§3.11: keine verworfene Alternative, kein abwesender Text (Suche in den Diff-Kommentaren ohne Treffer). Zusagen in Kopfzeilen (`main.go`, `faelle.go`, SPEC-102 §Grenze) nennen nur, was gefangen wird; SPEC-102 sagt selbst, dass der Bau in der Kopie vom Image abweichen kann. Abweichung: F5 (Beschreibung zu eng, keine Überzusage).
* `.golangci.yml` §3.6: die Ausnahme `^tools/(gegenproben|pruefung)/main\.go$` für `os.Std(out|err)` trägt `# Why:` mit Text; sie ist eine Ausnahme je Pfad und senkt keine Schwelle; `lint.sh` nimmt `tools/pruefung` als zweites Modul (`tools/*/go.mod`) und das Dockerfile kopiert es in die Lint-Stufe. Kein Befund.
* Verhaltensgleichheit zu `2091ba6` (`git show 2091ba6:Dockerfile` und `go-gates.mk`): `gofmt -l .`, `go vet ./...`, `go test -trimpath -buildvcs=false ./...` laufen unverändert in `/src` (Test-Zweig der Mutationen N3/M4 pinnt Verzeichnis und Flags); der Bau behält `TARGETOS/TARGETARCH`, `-trimpath`, `-buildvcs=false`, `-ldflags` und `VERSION`; Version und `-trimpath` werden weiter gegen das gebaute Binary geprüft, jetzt im Build selbst (Fehlschlag verhindert das Image, strenger). Ein Cross-Build bleibt prüfbar nur wie vorher (Binary muss auf dem Build-Host starten). Nichts geschwächt.
* Plan §3/§6 gegen Code (§3.9): Dateiliste, Stufe `pruefung`, Aufrufe, Lint-Ausnahme, Eingabenliste und Fallnamen stimmen. Abweichung: der Plan nennt unter „Mutation" für Nr. 7 beim Build etwas, das nicht unterscheidbar ist (F10); §6 trägt den Ausgang „weiter offen" für den Bau in der Kopie und das Zeitbudget ehrlich.
* Vertrag vs. Code (§3.16): Randformen Nr. 1 bis 11 von `pruefung` stimmen mit dem Code überein; offen sind die in F3, F4, F7, F8, F9 genannten Formen.

## Geprüft, ohne Befund

* `tools/pruefung/` (Quelle und Tests): bis auf F3, F7, F8, F9 und die INFO-Punkte ohne Befund; alle nicht äquivalenten Mutationen außer F3 in mindestens einer Spalte rot.
* `harness/mk/go-gates.mk`, `.dockerignore`: ohne Befund.
* `.golangci.yml`: ohne Befund.
* `harness/README.md`: Sensors-Tabelle und „Nicht behauptet" stimmen mit Dockerfile und Vertrag überein.
* `harness/vertraege/SPEC-036-test.md`: stimmt mit dem Code überein (Randform 4 inkl. Exit ≠ 0 gefangen durch `TestBuildBinaryStartetNicht`).
* `tools/gegenproben/` (außer F4, F5, F6): Bau-Zweig, `ersetzeOrte`, `befehl` durch Unit-Tests gefangen; WantText der Fälle ist spezifisch genug (kein Fall blieb bei einer kaputten `pruefung`-Mutation unbemerkt, wenn die Mutation im Katalog liegt).

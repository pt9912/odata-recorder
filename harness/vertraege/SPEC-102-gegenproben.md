# SPEC-102 — Gegenproben (`make gegenproben`)

**Werkzeug:** `make gegenproben`; Quelle `tools/gegenproben/` (eigenes Go-Modul), Stufe `gegenproben` des `Dockerfile`.

**Festlegung:** Eine Gegenprobe weist an einem Gate nach, dass es einen Verstoß rot meldet und einen erlaubten Fall nicht.
`make gegenproben` fährt die Gegenproben von `make a-check`, `make lint`, `make migration-check`, `make test`, `make build` und `make test-race`; sie gehören nicht zu `GATE_CHECKS`.

**Aufbau:** Ein Go-Programm läuft in einem Docker-Build ohne Netz, auf dem golangci-lint-Image mit dem `a-check`-Binary aus dessen Image und dem Programm `pruefung` aus `SPEC-036`.
Je Fall legt es eine Kopie der Eingaben in einem Wegwerfverzeichnis an, schreibt oder ändert dort Dateien, ruft das Gate als Prozess und vergleicht Exit-Code und Ausgabe.
Alle Fälle laufen nebenläufig (höchstens so viele wie Prozessorkerne); ein Docker-Build, -Lauf oder `docker cp` je Fall gibt es nicht.
Der Arbeitsbaum wird nie verändert.
Die Fälle 34 bis 36 (`make test-race`, `SPEC-122`) legen statt der Kopie der Eingaben ein Wegwerfmodul ohne Abhängigkeiten an und schreiben die Pakete selbst: ein Lauf über `src/` unter `-race` wäre das Ziel und nicht seine Gegenprobe.
Der Inhalt der Dateien, die sie schreiben, führt kein Zeichen über ASCII, damit ein Fixture nicht davon abhängt, ob das Werkzeug des Schreibers eine Escape-Folge in das Zeichen umsetzt.
Sie laufen in dieser Stufe auf dem golangci-lint-Image, das Go 1.27.1 und gcc 14.2.0 trägt (gemessen am 2026-10-22), nicht auf dem Image der Stufe `race`; `pruefung race` setzt `CGO_ENABLED=1` für seinen Aufruf selbst, die Umgebung der Stufe (`CGO_ENABLED=0`) bleibt.
Die Fälle setzen `CGO_ENABLED` nicht selbst, sie erben `0` von der Stufe; dass die Stufe es setzt, fängt kein Fall. Die Mutation „`pruefung race` setzt cgo nicht“ fangen die Fälle 34 bis 36 nur, solange die Umgebung der Stufe cgo nicht einschaltet; der Test von `pruefung` fängt sie in jeder Umgebung (`SPEC-122` Nr. 1).

**Eingaben:** `src/`, `go.mod`, `go.sum`, `.a-check.yml`, `.golangci.yml`, `tools/harness/lint.sh`, `tools/harness/migration-check.sh`, `tools/gegenproben/`, `tools/pruefung/`, `Dockerfile` und `.dockerignore`; dazu die Pins beider Images.
`tools/harness/gegenproben-eingabe.sh` bildet daraus eine Prüfsumme (SHA-256 über die Dateien und die Pins).
Stimmt sie mit dem Stempel `harness/gegenproben.stempel` überein, meldet `make gegenproben` den unveränderten Stand und fährt nichts; sonst fährt es die Stufe, ohne den Build-Cache für sie, und schreibt den Stempel nur nach einem grünen Lauf.
Der Stempel ist versioniert und reist mit dem Commit in andere Klone und Container; die Prüfsumme bindet ihn an die Eingaben, ein veralteter Stempel stimmt nicht mehr überein und erzwingt den Lauf.
Ein roter Lauf lässt den Stempel unverändert, ein fehlender Stempel erzwingt den Lauf; Änderungen an Doku, Plan und Reports berühren die Prüfsumme nicht.

**Fälle** (Name · Gate · Eingriff in der Kopie · Erwartung):
1. `a-check/basis` · `a-check` · keiner · Exit 0.
2. `a-check/kern-auf-driven-adapter` · `a-check` · ein Paket in `src/hexagon/model` importiert ein Paket aus `src/adapters/driven` · Exit 1, Ausgabe nennt `core-impurity`.
3. `a-check/adapter-auf-adapter` · `a-check` · ein Paket in `src/adapters/driving` importiert ein Paket aus `src/adapters/driven` · Exit 1, Ausgabe nennt `lateral-adapter`.
4. `a-check/erlaubter-weg-ueber-port` · `a-check` · ein Driven Adapter importiert einen Outbound Port, ein Service importiert Port und Modell · Exit 0.
5. `lint/basis` · `lint.sh` · keiner · Exit 0, Ausgabe nennt `0 issues` (golangci-lint lief).
6. `lint/direktive-nolint` · `lint.sh` · `//nolint:unused // x` in einer Datei von `src/` · Exit 1, Ausgabe nennt `Direktive` (eigene Prüfung) und `nolintlint` (golangci-lint).
7. `lint/ausnahme-ohne-why` · `lint.sh` · eine Regel unter `exclusions.rules` ohne `# Why:` · Exit 1, Ausgabe nennt `# Why:`.
8. `lint/ausnahme-mit-why` · `lint.sh` · eine Regel unter `exclusions.rules` mit `# Why:` und Text · Exit 0, Ausgabe nennt `0 issues`.
9. `lint/direktive-mit-leerzeichen` · `lint.sh` · `//  nolint:unused // x` (zwei Leerzeichen) in einer Datei von `src/` · Exit 1, Ausgabe nennt `Direktive` und `nolintlint`.
10. `lint/profil-fehlt` · `lint.sh` · `.golangci.yml` ist aus der Kopie gelöscht · Exit 2, Ausgabe nennt `fehlt oder ist nicht lesbar`.
11. `lint/why-falsch-eingerueckt` · `lint.sh` · ein `# Why:`-Block über der Regel ist nicht mit sechs Leerzeichen eingerückt · Exit 1, Ausgabe nennt `# Why:`.
12. `lint/zweites-modul-sauber` · `lint.sh` · ein Modul `tools/probe` mit `go.mod` und einer Go-Datei ohne Befund · Exit 0, Ausgabe nennt `0 issues`.
13. `lint/befund-im-zweiten-modul` · `lint.sh` · dasselbe Modul mit einer ungenutzten Funktion · Exit 1, Ausgabe nennt `unused`.
14. `lint/ohne-go-datei` · `lint.sh` · alle Go-Dateien von `src/` sind aus der Kopie gelöscht · Exit 1, Ausgabe nennt `keine Datei *.go`.
15. `lint/modul-ohne-go-datei` · `lint.sh` · ein Modul `tools/leer` mit `go.mod` und ohne Go-Datei · Exit 1, Ausgabe nennt `tools/leer: keine Datei *.go`.
16. `lint/go-nur-unter-tools` · `lint.sh` · alle Go-Dateien von `src/` sind gelöscht, ein Modul `tools/probe` trägt eine · Exit 1, Ausgabe nennt `./: keine Datei *.go`.
17. `migration/gleich` · `migration-check.sh` · das erzeugte DDL ist eine Kopie der Migration · Exit 0, Ausgabe nennt `migration-check: ok`.
18. `migration/abweichung` · `migration-check.sh` · das erzeugte DDL trägt eine zusätzliche Zeile · Exit 1, Ausgabe nennt `weicht vom erzeugten DDL ab` und die Zeile des Unterschieds aus `diff`, aber nicht `migration-check: ok`.
19. `migration/weitere-datei` · `migration-check.sh` · neben `0001_init.sql` liegt eine Datei `0002_x.sql` · Exit 1, Ausgabe nennt `ungeprüft`.
20. `a-check/bibliothek-in-fremdem-adapter` · `a-check` · ein Paket unter `src/adapters/driven/andere` importiert `modernc.org/sqlite` · Exit 1, Ausgabe nennt `tech-leak`.
21. `migration/migration-fehlt` · `migration-check.sh` · `0001_init.sql` ist aus der Kopie gelöscht · Exit 2, Ausgabe nennt die Migration als fehlend und nicht das erzeugte DDL.
22. `migration/erzeugtes-ddl-fehlt` · `migration-check.sh` · die Datei mit dem erzeugten DDL fehlt · Exit 2, Ausgabe nennt das erzeugte DDL als fehlend und nicht die Migration.
23. `migration/falsche-aufrufform` · `migration-check.sh` · ein Argument statt zwei · Exit 2, Ausgabe nennt `Aufruf`.
24. `migration/zu-viele-argumente` · `migration-check.sh` · drei Argumente statt zwei · Exit 2, Ausgabe nennt `Aufruf`.
25. `test/basis` · `pruefung test` mit einem Go-Muster, das nur die Pakete unter `src/bootstrap` einschließt · keiner · Exit 0, Ausgabe nennt `pruefung test: ok`.
26. `test/nicht-formatiert` · `pruefung test` · eine Datei in `src/bootstrap/` mit ungleichmäßigen Leerzeichen · Exit 1, Ausgabe nennt `nicht gofmt-formatiert` und den Dateinamen.
27. `test/vet-befund` · `pruefung test` · eine Datei in `src/bootstrap/` mit einem `Printf`, dessen Verb nicht zum Argument passt · Exit 1, Ausgabe nennt `go vet:`.
28. `test/roter-test` · `pruefung test` · eine Testdatei in `src/bootstrap/` mit einem Test, der `t.Fatal` ruft · Exit 1, Ausgabe nennt `go test:`.
29. `build/basis` · `pruefung build` · das Binary entsteht in der Kopie mit `-trimpath` und der Version, die der Aufruf nennt · Exit 0, Ausgabe nennt `pruefung build: ok`.
30. `build/ohne-trimpath` · `pruefung build` · das Binary entsteht ohne `-trimpath` · Exit 1, Ausgabe nennt `nicht mit -trimpath gebaut`.
31. `build/falsche-version` · `pruefung build` · das Binary trägt eine andere Version als der Aufruf nennt · Exit 1, Ausgabe nennt `Version des Binary` und beide Versionen.
32. `build/binary-fehlt` · `pruefung build` · der Pfad des Binary zeigt auf keine Datei · Exit 2, Ausgabe nennt `Binary fehlt`.
33. `test/aufruf` · `pruefung` · ohne Argumente · Exit 2, Ausgabe nennt `Aufruf:`.
34. `race/ohne-rennen` · `pruefung race` · Wegwerfmodul mit einem Paket unter `src/hexagon/`, dessen Test zwei Goroutinen einen Zähler unter einer Sperre erhöhen lässt und auf beide wartet · Exit 0, Ausgabe nennt `pruefung race: ok`.
35. `race/rennen-in-jedem-teilbaum` · `pruefung race` ohne Muster · Wegwerfmodul mit je einem Paket unter `src/hexagon/` (Rennen nur im Testcode), `src/adapters/` und `src/bootstrap/` (Rennen im Code des Pakets, vom Test aufgerufen) und einem vierten Paket unter `src/hexagon/` mit einem Test, der `t.Fatal` ruft, ohne Rennen; in den drei Paketen mit Rennen starten zwei Goroutinen ohne Kante zwischen ihnen und ändern denselben Zähler, der Test wartet erst danach auf beide · Exit 1, Ausgabe nennt `WARNING: DATA RACE` und für alle vier Pakete eine Zeile `FAIL` mit dem Importpfad.
36. `race/rennen-und-zeitgrenze` · `pruefung race` mit den Argumenten `-timeout=5s ./...` · Wegwerfmodul mit einem Paket, dessen Test ein Rennen wie in Fall 35 erzeugt, auf beide Goroutinen wartet und danach blockiert (`select {}`) · Exit 1, Ausgabe nennt `WARNING: DATA RACE` und `test timed out`.
Die erste Spalte ist der Name im Bericht; ein Name kommt genau einmal vor.
Das Rennen der Fälle 35 und 36 ist nicht zeitabhängig (`SPEC-122` Nr. 8).
Gemessen am 2026-10-22 mit dem Testbinary der Fixtures unter `-race`, je Lauf ein eigener Prozess mit `-test.count=1`: das Rennen jedes der drei Teilbäume von Fall 35 wurde in 200 von 200 Läufen gemeldet, die Fixture von Fall 34 blieb in 200 von 200 Läufen ohne Meldung, Fall 36 mit beiden Meldungen (`WARNING: DATA RACE` und `test timed out`) in 100 von 100.

**Randformen:**
1. Eine Eingabe fehlt in der Eingabekopie (etwa `.a-check.yml`) oder die Kopie scheitert: Meldung `gegenproben: <Grund>`, Exit-Code 2, kein Fall läuft.
2. Ein Programm (`a-check`, `golangci-lint`, `sh`, `pruefung`, `go`, `gofmt`) ist nicht aufrufbar: Meldung `gegenproben: <Programm> nicht aufrufbar`, Exit-Code 2.
2a. Der Bau des Binary eines Falls (Fälle 29 bis 31) scheitert: der Fall ist rot, Text `Bau des Binary scheitert` und die letzten Zeilen der Ausgabe von `go build`; das Gate wird nicht gerufen.
3. Ein Verstoß-Fall endet mit Exit 0: der Fall ist rot, Text `Verstoß nicht gemeldet`, Exit-Code 1 des Programms.
4. Ein Verstoß-Fall endet mit einem anderen Code als der erwarteten 1 (etwa 2 bei einem Aufruffehler des Gates) oder die Ausgabe nennt das erwartete Wort nicht: der Fall ist rot mit erwartetem und tatsächlichem Wert, Exit-Code 1.
5. Ein Erlaubt-Fall endet nicht mit Exit 0: der Fall ist rot, die letzten Zeilen der Ausgabe stehen im Bericht, Exit-Code 1.
6. Ein Fall läuft länger als 120 s: die Prozessgruppe des Gates wird beendet, der Fall ist rot, Text `Zeit überschritten`. Die Frist ist auf einen Lauf zugleich je Host bemessen: jeder Lauf startet bis zu so viele Fälle wie Kerne, zwei Läufe teilen sie, und ein Fall kann die Frist ohne Fehler im Gate reißen; ein einzeln wiederholter Lauf zeigt, ob es an der Last lag.
7. Alle Fälle laufen auch bei einem roten Fall zu Ende; das Programm meldet erst danach, Exit 0 genau dann, wenn kein Fall rot ist.
8. Wegwerfverzeichnisse werden am Ende entfernt, auch bei roten Fällen; scheitert das Entfernen, stehen die Ergebnisse trotzdem in der Ausgabe, danach folgt die Meldung `gegenproben: <Grund>` und der Exit-Code 2; ein Abbruch des Builds lässt nichts zurück, weil der Container verworfen wird.
9. Die Eingabe ist unverändert: nach dem Lauf sind die Dateien der Eingaben bytegleich; ändert ein Fall sie, endet das Programm mit Meldung und Exit-Code 2.
10. Ohne Git-Arbeitsbaum oder bei einem Eingabepfad, der fehlt (`go.sum` ausgenommen), meldet `gegenproben-eingabe.sh` den Fehler und endet ungleich 0; `make gegenproben` bricht dann ab, ohne den Stempel zu ändern.
11. Die Fälle teilen den Cache von golangci-lint, nicht dessen Sperre: jeder Fall hat ein eigenes `TMPDIR`, weil golangci-lint dort eine Sperre gegen parallele Läufe ablegt; ohne das bricht ein nebenläufiger Lauf mit `parallel golangci-lint is running` ab.

Die Fälle 25 bis 28 rufen `go` im Wegwerfverzeichnis auf und teilen den Cache von `go` (`GOCACHE`) wie den von golangci-lint; die Fälle 29 bis 32 bauen das Binary mit `go build` in der Kopie, nicht im Image der Stufe `build`.
Weicht dieser Bau vom Bau des Image ab (Zielplattform, Flags), fängt die Gegenprobe die Abweichung nicht.
Die Fälle 34 bis 36 rufen `pruefung race`, das `go` mit cgo ruft; unter `-race` baut `go` die Standardbibliothek mit Instrumentierung neu, im kalten `GOCACHE` der Stufe und geteilt von den Fällen.
Weichen Compiler oder Go-Version des Images der Stufe `gegenproben` von denen des Images `RACE_IMAGE` ab, fängt die Gegenprobe die Abweichung nicht; trägt das Image keinen Compiler mehr (das golangci-lint-Image sagt ihn nicht zu), sind die Fälle rot, mit der Meldung von `go` im Bericht, und der Wechsel des Pins ändert die Prüfsumme und erzwingt den Lauf.
Dass das Ziel `src/` ganz und den Quellstand der Stufe `source` fährt, belegt `make gates`, nicht die Gegenprobe.

**Zeitbudget** (Wanduhr, `time make gegenproben`): bei geänderten Eingaben und kaltem Docker-Build-Cache höchstens 120 s, bei unveränderten Eingaben höchstens 1 s.
Gemessen am 2026-10-22 mit allen 36 Fällen auf einem Host mit 20 Kernen und schwankender Last, Images lokal vorhanden: 80,5 bis 93,8 s im Ganzen (`time make gegenproben`).
Davon laufen die Fälle (Schritt `RUN gegenproben` der Stufe) 54,3 bis 67 s, das Prüfen und Bauen des Programms rund 7 s und der Export des Images der Stufe 16,8 bis 18 s.
Bei unveränderten Eingaben dauert der Aufruf 0,3 s.
Die `go`-Aufrufe der Fälle 25 bis 28 kosten bei kaltem `GOCACHE` bis zu 27 s je Fall.
Das Budget liegt rund 28 % über dem höchsten gemessenen Wert, 93,8 s; kein Gate und kein Hook prüft es.
Liegt ein Lauf darüber, misst der Implementer neu und trägt den Wert mit Grund ein.

**Pflicht zum Fahren:** `make gegenproben` läuft bei jeder Änderung am geprüften Gate, am Profil, an der Architektur-Regel, an der Gegenprobe oder an diesem Vertrag, am Ende jeder Welle und vor dem Release.

**Stand:** vorhanden für `make a-check`, `make lint`, `make migration-check`, `make test`, `make build` und `make test-race`.

**Grenze:** Kein Gate und kein Hook liest den Stempel; er belegt nur, dass ein Lauf zu dieser Prüfsumme grün war, und wer ihn von Hand schreibt, belegt nichts.
Die Fälle prüfen die Importrichtungen `core-impurity` und Adapter auf Adapter sowie die Direktiv- und `# Why:`-Regeln; weitere Verstoßformen kommen mit den Bibliotheksgrenzen.
Das Programm selbst prüft die Stufe mit `gofmt`, `go vet` und `go test`; `make lint` prüft es als zweites Modul (`SPEC-101`).
Die Liste der Eingaben in `tools/harness/gegenproben-eingabe.sh` steht einmal im Skript; dass sie `tools/pruefung` nennt, fängt kein Fall — wer den Eintrag entfernt, lässt die Prüfsumme bei einer Änderung an `pruefung` gleich.
Die Fälle zu `make test` und `make build` prüfen `pruefung` mit den Verstoßformen aus der Liste; sie fangen nicht, dass ein Ziel `pruefung` nicht ruft (das Gate ist im `Dockerfile` verdrahtet).
Die Gegenproben von `docs-check`, `schema-check` und `baseline-verify` fehlen.

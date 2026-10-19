# SPEC-102 — Gegenproben (`make gegenproben`)

**Werkzeug:** `make gegenproben`; Quelle `tools/gegenproben/` (eigenes Go-Modul), Stufe `gegenproben` des `Dockerfile`.

**Festlegung:** Eine Gegenprobe weist an einem Gate nach, dass es einen Verstoß rot meldet und einen erlaubten Fall nicht.
`make gegenproben` fährt die Gegenproben von `make a-check` und `make lint`; sie gehören nicht zu `GATE_CHECKS`.

**Aufbau:** Ein Go-Programm läuft in einem Docker-Build ohne Netz, auf dem golangci-lint-Image mit dem `a-check`-Binary aus dessen Image.
Je Fall legt es eine Kopie der Eingaben in einem Wegwerfverzeichnis an, schreibt oder ändert dort Dateien, ruft das Gate als Prozess und vergleicht Exit-Code und Ausgabe.
Alle Fälle laufen nebenläufig (höchstens so viele wie Prozessorkerne); ein Docker-Build, -Lauf oder `docker cp` je Fall gibt es nicht.
Der Arbeitsbaum wird nie verändert.

**Eingaben:** `src/`, `go.mod`, `.a-check.yml`, `.golangci.yml`, `tools/harness/lint.sh` und `tools/gegenproben/`; dazu die Pins beider Images.
Nur eine Änderung daran lässt die Stufe neu laufen, ein roter Lauf wird nicht zwischengespeichert.
Bei unveränderten Eingaben meldet `make gegenproben`, dass der letzte Lauf grün war und nichts neu läuft.

**Fälle** (Name · Gate · Eingriff in der Kopie · Erwartung):
1. `a-check/basis` · `a-check` · keiner · Exit 0.
2. `a-check/kern-auf-driven-adapter` · `a-check` · ein Paket in `src/hexagon/model` importiert ein Paket aus `src/adapters/driven` · Exit 1, Ausgabe nennt `core-impurity`.
3. `a-check/adapter-auf-adapter` · `a-check` · ein Paket in `src/adapters/driving` importiert ein Paket aus `src/adapters/driven` · Exit 1.
4. `a-check/erlaubter-weg-ueber-port` · `a-check` · ein Driven Adapter importiert einen Outbound Port, ein Service importiert Port und Modell · Exit 0.
5. `lint/basis` · `lint.sh` · keiner · Exit 0.
6. `lint/direktive-nolint` · `lint.sh` · `//nolint:unused // x` in einer Datei von `src/` · Exit 1, Ausgabe nennt `Direktive`.
7. `lint/ausnahme-ohne-why` · `lint.sh` · eine Regel unter `exclusions.rules` ohne `# Why:` · Exit 1, Ausgabe nennt `# Why:`.
8. `lint/ausnahme-mit-why` · `lint.sh` · eine Regel unter `exclusions.rules` mit `# Why:` und Text · Exit 0.
Die erste Spalte ist der Name im Bericht; ein Name kommt genau einmal vor.

**Randformen:**
1. Eine Eingabe fehlt in der Eingabekopie (etwa `.a-check.yml`) oder die Kopie scheitert: Meldung `gegenproben: <Grund>`, Exit-Code 2, kein Fall läuft.
2. Ein Gate-Programm (`a-check`, `golangci-lint`, `sh`) ist nicht aufrufbar: Meldung `gegenproben: <Programm> nicht aufrufbar`, Exit-Code 2.
3. Ein Verstoß-Fall endet mit Exit 0: der Fall ist rot, Text `Verstoß nicht gemeldet`, Exit-Code 1 des Programms.
4. Ein Verstoß-Fall endet mit einem anderen Code als der erwarteten 1 (etwa 2 bei einem Aufruffehler des Gates) oder die Ausgabe nennt das erwartete Wort nicht: der Fall ist rot mit erwartetem und tatsächlichem Wert, Exit-Code 1.
5. Ein Erlaubt-Fall endet nicht mit Exit 0: der Fall ist rot, die letzten Zeilen der Ausgabe stehen im Bericht, Exit-Code 1.
6. Ein Fall läuft länger als 120 s: der Prozess wird beendet, der Fall ist rot, Text `Zeit überschritten`.
7. Alle Fälle laufen auch bei einem roten Fall zu Ende; das Programm meldet erst danach, Exit 0 genau dann, wenn kein Fall rot ist.
8. Wegwerfverzeichnisse werden am Ende entfernt, auch bei roten Fällen; ein Abbruch des Builds lässt nichts zurück, weil der Container verworfen wird.
9. Die Eingabe ist unverändert: nach dem Lauf sind die Dateien der Eingaben bytegleich.

**Zeitbudget** (Wanduhr, `time make gegenproben`): bei kaltem Docker-Build-Cache höchstens 60 s, bei unveränderten Eingaben höchstens 5 s.

**Stand:** vorgesehen für `make a-check` und `make lint`.

**Grenze:** Die Fälle prüfen die Importrichtungen `core-impurity` und Adapter auf Adapter sowie die Direktiv- und `# Why:`-Regeln; weitere Verstoßformen kommen mit den Bibliotheksgrenzen.
Die Gegenproben von `docs-check`, `schema-check` und `baseline-verify` fehlen.

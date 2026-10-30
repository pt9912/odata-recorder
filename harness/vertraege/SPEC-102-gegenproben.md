# SPEC-102 — Gegenproben (`make gegenproben`)

**Werkzeug:** `make gegenproben`; Quelle `tools/gegenproben/` (eigenes Go-Modul), Stufe `gegenproben` des `Dockerfile`.

**Festlegung:** Eine Gegenprobe weist an einem Gate nach, dass es einen Verstoß rot meldet und einen erlaubten Fall nicht.
`make gegenproben` fährt die Gegenproben von `make a-check`, `make lint`, `make migration-check`, `make test`, `make build`, `make test-race` und `make handbuch-check`; sie gehören nicht zu `GATE_CHECKS`.

**Aufbau:** Ein Go-Programm läuft in einem Docker-Build ohne Netz, auf dem golangci-lint-Image mit dem `a-check`-Binary aus dessen Image und dem Programm `pruefung` aus `SPEC-036`.
Je Fall legt es eine Kopie der Eingaben in einem Wegwerfverzeichnis an, schreibt oder ändert dort Dateien, ruft das Gate als Prozess und vergleicht Exit-Code und Ausgabe.
Alle Fälle laufen nebenläufig (höchstens so viele wie Prozessorkerne); ein Docker-Build, -Lauf oder `docker cp` je Fall gibt es nicht.
Der Arbeitsbaum wird nie verändert.
Die Fälle 25 bis 28 (`make test`) und 34 bis 36 (`make test-race`, `SPEC-122`) legen statt der Kopie der Eingaben ein Wegwerfmodul ohne Abhängigkeiten an und schreiben die Pakete selbst: ein Lauf über `src/` wäre das Ziel und nicht seine Gegenprobe.
Die Fälle 25 bis 28 fahren damit kein Paket von `src/`: die Tests des Produkts wachsen mit jedem Slice, und ein Fall, der sie mitfährt, braucht neben den übrigen Fällen (bis zu so viele wie Kerne, Randform 6) die Frist von 120 s auf, ohne dass das Gate fehlt (gemessen am 2026-10-24: `test/basis` und `test/roter-test` in zwei von drei Läufen bei 120,6 s rot, mit der Vertragssuite von `ODataCodec` in `src/bootstrap`).
Die Parallelität bleibt, wie sie ist; die Last, die einen Fall an die Frist bringt, erzeugen die Fälle selbst, und kein Fall hängt von der Laufzeit des Produkts ab.
Dass jeder dieser Fälle ein Wegwerfmodul anlegt, fängt der Test `TestTestFaelleLegenEinWegwerfmodulAn` in `tools/gegenproben/main_test.go` (Muster: `TestRaceFaelleLegenEinWegwerfmodulAn`).
Der Inhalt der Dateien, die sie schreiben, führt kein Zeichen über ASCII, damit ein Fixture nicht davon abhängt, ob das Werkzeug des Schreibers eine Escape-Folge in das Zeichen umsetzt.
Sie laufen in dieser Stufe auf dem golangci-lint-Image, das Go 1.27.1 und gcc 14.2.0 trägt (gemessen am 2026-10-22), nicht auf dem Image der Stufe `race`; `pruefung race` setzt `CGO_ENABLED=1` für seinen Aufruf selbst, die Umgebung der Stufe (`CGO_ENABLED=0`) bleibt.
Die Fälle setzen `CGO_ENABLED` nicht selbst, sie erben `0` von der Stufe; dass die Stufe es setzt, fängt kein Fall. Die Mutation „`pruefung race` setzt cgo nicht“ fangen die Fälle 34 bis 36 nur, solange die Umgebung der Stufe cgo nicht einschaltet; der Test von `pruefung` fängt sie in jeder Umgebung (`SPEC-122` Nr. 1).
Die Fälle 38 bis 102 (`make groessenmarke`, `SPEC-128`) legen ein Wegwerfverzeichnis an, schreiben den Plan und die Messung als Dateien hinein und rufen das Programm `groessenmarke`, das die Stufe `groessenmarke` des `Dockerfile` baut und die Stufe `gegenproben` übernimmt; sie rufen weder `go` noch ein Gate.
Die Fälle 103 bis 114 rufen `tools/harness/groessenmarke.sh` im Wegwerfverzeichnis mit zwei Ersatz-Skripten, die jeden Aufruf mit allen Argumenten in eine Datei schreiben und zählen: einem Ersatz für `git`, der am Anfang von `PATH` steht, und einem für das Programm.
Die Fälle 115 bis 120 rufen denselben Treiber mit dem echten `git` des Images (2.47.3, gemessen am 2026-10-30) auf einem Repo, das der Fall im Wegwerfverzeichnis anlegt (`git init`, zwei Commits; Konfiguration allein über `-c` und `GIT_CONFIG_GLOBAL=/dev/null`).
Diese Fälle tragen Zeichen jenseits von ASCII nur dort, wo `SPEC-128` sie verlangt (`nebenläufig`, Pfade mit Umlaut), und bilden sie aus Bytewerten.
Die Fälle 121 bis 130 (`make handbuch-check`, `SPEC-200`) rufen `pruefung handbuch` im Wegwerfverzeichnis mit einem Ersatz-Skript als Binary und rufen weder `go` noch `docker` noch `git`.
Die Fälle 38 bis 130 lesen stdout und stderr getrennt, die Fälle 1 bis 37 gemeinsam (Randform 13).

**Eingaben:** `src/`, `go.mod`, `go.sum`, `.a-check.yml`, `.golangci.yml`, `tools/harness/lint.sh`, `tools/harness/migration-check.sh`, `tools/harness/groessenmarke.sh`, `tools/gegenproben/`, `tools/pruefung/`, `tools/groessenmarke/`, `Dockerfile` und `.dockerignore`; dazu die Pins beider Images.
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
25. `test/basis` · `pruefung test` mit dem Wegwerfmodul als Wurzel und ohne Muster · Wegwerfmodul (`go.mod`, ein Paket ohne Befund mit einem bestandenen Test) · Exit 0, Ausgabe nennt `pruefung test: ok`.
26. `test/nicht-formatiert` · `pruefung test` · im Wegwerfmodul eine Datei mit ungleichmäßigen Leerzeichen · Exit 1, Ausgabe nennt `nicht gofmt-formatiert` und den Dateinamen.
27. `test/vet-befund` · `pruefung test` · im Wegwerfmodul eine Datei mit einem `Printf`, dessen Verb nicht zum Argument passt · Exit 1, Ausgabe nennt `go vet:`.
28. `test/roter-test` · `pruefung test` · im Wegwerfmodul ein Paket mit einem Test, der `t.Fatal` ruft · Exit 1, Ausgabe nennt `go test:`.
29. `build/basis` · `pruefung build` · das Binary entsteht in der Kopie mit `-trimpath` und der Version, die der Aufruf nennt · Exit 0, Ausgabe nennt `pruefung build: ok`.
30. `build/ohne-trimpath` · `pruefung build` · das Binary entsteht ohne `-trimpath` · Exit 1, Ausgabe nennt `nicht mit -trimpath gebaut`.
31. `build/falsche-version` · `pruefung build` · das Binary trägt eine andere Version als der Aufruf nennt · Exit 1, Ausgabe nennt `Version des Binary` und beide Versionen.
32. `build/binary-fehlt` · `pruefung build` · der Pfad des Binary zeigt auf keine Datei · Exit 2, Ausgabe nennt `Binary fehlt`.
33. `test/aufruf` · `pruefung` · ohne Argumente · Exit 2, Ausgabe nennt `Aufruf:`.
34. `race/ohne-rennen` · `pruefung race` · Wegwerfmodul mit einem Paket unter `src/hexagon/`, dessen Test zwei Goroutinen einen Zähler unter einer Sperre erhöhen lässt und auf beide wartet · Exit 0, Ausgabe nennt `pruefung race: ok`.
35. `race/rennen-in-jedem-teilbaum` · `pruefung race` ohne Muster · Wegwerfmodul mit je einem Paket unter `src/hexagon/` (Rennen nur im Testcode), `src/adapters/` und `src/bootstrap/` (Rennen im Code des Pakets, vom Test aufgerufen) und einem vierten Paket unter `src/hexagon/` mit einem Test, der `t.Fatal` ruft, ohne Rennen; in den drei Paketen mit Rennen starten zwei Goroutinen ohne Kante zwischen ihnen und ändern denselben Zähler, der Test wartet erst danach auf beide · Exit 1, Ausgabe nennt `WARNING: DATA RACE` und für alle vier Pakete eine Zeile `FAIL` mit dem Importpfad.
36. `race/rennen-und-zeitgrenze` · `pruefung race` mit den Argumenten `-timeout=5s ./...` · Wegwerfmodul mit einem Paket, dessen Test ein Rennen wie in Fall 35 erzeugt, auf beide Goroutinen wartet und danach blockiert (`select {}`) · Exit 1, Ausgabe nennt `WARNING: DATA RACE` und `test timed out`.
37. `a-check/xml-im-kern` · `a-check` · ein Paket in `src/hexagon/model` importiert `encoding/xml` · Exit 1, Ausgabe nennt `core-impurity`.

Die Fälle 38 bis 120 gehören zu `make groessenmarke` (`SPEC-128`). *Plan* heißt ein Plan mit einer Tabelle *Zusage · Fänger* von *a* nebenläufigen und *b* anderen Zeilen und einem Block (`SPEC-128` Nr. 3); *Fixture B* ist der Plan mit *a* = 0, *b* = 10, `fakes` 0, `produktcode` 0, also *S* = 800 und der Marke 1100, mit `schaetzung: 800`, `marke: 1100` und einer Messung ohne Datensatz, soweit ein Fall nichts anderes sagt. Aufruf ist `groessenmarke pruefen <Plan> <Messung>`, wo nichts anderes steht.
In diesen Fällen heißt „Ausgabe nennt“ der Strom, den `SPEC-128` der Meldung zuweist: Befunde, Eingabefehler und Meldungen des Treibers (Nr. 6 a, 7 d) stehen auf stderr, die Zeile `ok`, die Hinweise von `pruefen`, die Messpunkt-Zeile, `keine-messung` und `kein Slice in Arbeit` auf stdout; „stdout nennt“ und „stderr nennt“ lesen diesen Strom allein, „aber nicht“ gilt für beide Ströme zusammen (Randform 13).
38. bis 40. `groessenmarke/messpunkt-csdl`, `-json`, `-import` · Plan mit 44 (davon 2), 49 (2) und 70 (3) Zeilen, `fakes` 0, 0 und 8, `produktcode` 1506, 1500 und 2000, `schaetzung` 5036, 5430 und 9215, `marke` 6800, 7400 und 12500, Messung 4333, 7236 und 10027 · Exit 0, Ausgabe nennt `ok, Messung <Messung> von Marke <Marke>`.
41. bis 43. `groessenmarke/marke-der-platte-csdl`, `-json`, `-import` · derselbe Plan mit der Marke, die die Formel vor `SPEC-128` ergab (5600, 6000 und 8200), Messung 0 · Exit 1, Ausgabe nennt `weicht von der Formel ab` und die erwartete Marke 6800, 7400 bzw. 12500.
44. `groessenmarke/marke-eine-stufe-zu-niedrig` · Fixture B mit `marke: 1000` · Exit 1, Ausgabe nennt `weicht von der Formel ab` und `erwartet 1100`.
45. `groessenmarke/marke-eine-stufe-zu-hoch` · Fixture B mit `marke: 1200` · Exit 1, wie Fall 44.
46. `groessenmarke/schaetzung-um-eins-zu-niedrig` · Fixture B mit `schaetzung: 799` und der Marke 1100 (die Formel zu 799) · Exit 1, Ausgabe nennt `weicht von der Tabelle ab` und `Tabelle ergibt 800`, aber nicht `weicht von der Formel ab`.
47. `groessenmarke/schaetzung-um-eins-zu-hoch` · wie Fall 46 mit `schaetzung: 801` · Exit 1, wie Fall 46.
48. `groessenmarke/tabelle-gewachsen-im-code` · wie Fall 46, Messung 1 · Exit 0, Ausgabe nennt `Hinweis` und `Tabelle ergibt 800, Schätzung 799`.
49. `groessenmarke/messung-gleich-marke` · Fixture B, Messung 1100 · Exit 0, Ausgabe nennt `ok, Messung 1100 von Marke 1100`.
50. `groessenmarke/messung-ueber-marke-ohne-eintrag` · Fixture B, Messung 1101 · Exit 1, Ausgabe nennt `ohne Eintrag` und die Zeile `ueberschreitung: 1200 |`.
51. `groessenmarke/eintrag-deckt` · Fixture B, Messung 1101, Eintrag `ueberschreitung: 1101 | Grund` · Exit 0.
52. `groessenmarke/eintrag-reicht-nicht` · Fixture B, Messung 1101, Eintrag `ueberschreitung: 1100 | Grund` · Exit 1, Ausgabe nennt `Eintrag reicht nicht` und `größtes N 1100`.
53. `groessenmarke/eintrag-ohne-grund` · Fixture B, Messung 1101, Eintrag `ueberschreitung: 1200 |` · Exit 1, Ausgabe nennt `ohne Grund`, aber nicht `ohne Eintrag` (`SPEC-128` Nr. 5 d: ein leerer Grund deckt die Überschreitung und ist selbst der Befund).
54. `groessenmarke/eintrag-ohne-ueberschreitung` · Fixture B, Messung 1100, Eintrag `ueberschreitung: 1200 | Grund` · Exit 0, Ausgabe nennt `Eintrag ohne Überschreitung`.
55. `groessenmarke/paar-marke-unter-formel-und-messung-ueber-marke` · Fixture B mit `schaetzung: 800`, `marke: 1000`, Messung 1001 · Exit 1, Ausgabe nennt `weicht von der Formel ab` und `ohne Eintrag`.
56. `groessenmarke/paar-eintrag-und-marke-unter-formel` · wie Fall 55 mit dem Eintrag `ueberschreitung: 1100 | Grund` · Exit 1, Ausgabe nennt `weicht von der Formel ab`, aber nicht `ohne Eintrag`.
57. `groessenmarke/schnitt-ohne-grund` · Plan mit 27 Zeilen `andere`, `produktcode` 63 (*S* = 2223), `marke` 3100, ohne `schnitt-grund`, Messung 0 · Exit 1, Ausgabe nennt `Schnittfrage ohne Grund`.
58. `groessenmarke/schnitt-mit-grund` · wie Fall 57 mit `schnitt-grund: Grund` · Exit 0.
59. `groessenmarke/schnitt-genau-darunter` · wie Fall 57 mit `produktcode` 62 (*S* = 2222), `marke` 3000, ohne `schnitt-grund` · Exit 0.
60. `groessenmarke/kein-block` · Plan ohne Block und ohne Zeile `**Größenmarke**` · Aufruf `abfrage <Plan>`: Exit 0, stdout `keine-messung`, stderr nennt `kein Block`.
61. `groessenmarke/prosa-form` · Plan ohne Block mit der Zeile `**Größenmarke** (…): über 2400 Zeilen …` (die Form der älteren Pläne) · `abfrage`: Exit 0, stdout `keine-messung`, stderr nennt `Prosa-Form`.
62. `groessenmarke/marke-steht-aus` · Fixture B mit Block ohne `basis`, `schaetzung`, `marke` · `abfrage`: Exit 0, stdout `keine-messung`, stderr nennt `Marke steht aus: Schätzung 800, Marke 1100`.
63. `groessenmarke/paar-marke-und-basis-fehlen` · Fixture B mit `schaetzung`, ohne `marke` und `basis` · `abfrage`: Exit 2, stderr nennt `Block unvollständig` sowie `marke` und `basis`.
64. `groessenmarke/block-unbekannter-schluessel` · Fixture B mit der Zeile `marken: 1100` · Exit 2, stderr nennt `Schlüssel marken unbekannt`.
65. bis 70. `groessenmarke/zahl-tausender-leerzeichen`, `-tausender-punkt`, `-minus`, `-fuehrende-null`, `-leer`, `-zehn-stellen` · Fixture B mit `produktcode:` auf `1 500`, `1.500`, `-1`, `01`, leer und `1234567890` · Exit 2, stderr nennt `Zahl ungültig: produktcode`.
71. bis 78. `groessenmarke/pfad-leer`, `-absolut`, `-punktpunkt`, `-magie`, `-doppelt`, `-in-spec`, `-in-docs`, `-in-vertraegen` · Fixture B mit einem `pfad:` auf leer, `/src`, `a/../b`, `:(top)src`, zweimal `src/`, `spec/lastenheft.md`, `docs/plan` und `harness/vertraege/README.md` · Exit 2, stderr nennt `Pfad ungültig` (bei `-doppelt` `Pfad doppelt`) und den Pfad.
79. `groessenmarke/tabelle-fehlt` · Plan, dessen einzige Tabelle den Kopf `Zeile · Zusage · Mutation · roter Test` trägt · Exit 2, stderr nennt `Tabelle fehlt`.
80. `groessenmarke/tabelle-ohne-zeilen` · Kopf und Trennzeile, keine Datenzeile · Exit 2, stderr nennt `Tabelle ohne Zeilen`.
81. `groessenmarke/tabelle-mehrdeutig` · zwei Tabellen mit dem Kopf der Tabelle · Exit 2, stderr nennt `Tabelle mehrdeutig`.
82. `groessenmarke/tabelle-zellenzahl` · eine Zeile mit einem ungeschützten `|` in einem Code-Span (sechs Zellen) · Exit 2, stderr nennt `Zeile 1: 6 Zellen statt 5`.
83. `groessenmarke/tabelle-nr-doppelt` · zwei Zeilen mit derselben `Nr.` · Exit 2, stderr nennt `doppelt`.
84. bis 88. `groessenmarke/klasse-leer`, `-ascii-ae`, `-grossbuchstabe`, `-zaehlt-nicht`, `-fett` · eine Zeile mit der Klasse leer, `nebenlaeufig`, `Nebenläufig`, `zählt nicht` und `**andere**` · Exit 2, stderr nennt `Klasse`, `unbekannt` und die Zelle.
89. `groessenmarke/messung-binaer` · Fixture B, Messung mit den Datensätzen `-`/`-`/`x.bin` und `3`/`2`/`y.go` · Exit 0, Ausgabe nennt `Messung 5 von Marke 1100` und `1 Binärdatei(en) nicht gezählt`.
90. `groessenmarke/messung-umbenennung` · Messung mit einem Datensatz, dessen Pfadfeld leer ist (Form einer Umbenennung) · Exit 2, stderr nennt `Umbenennung in der Messung`.
91. `groessenmarke/messung-leer` · Fixture B, Messung nur mit Kopf · Exit 0, Ausgabe nennt `Messung 0 von Marke 1100`.
92. `groessenmarke/messung-fremde-basis` · Kopf mit einer anderen `basis` als der Block · Exit 2, stderr nennt `gehört nicht zu diesem Plan`.
93. bis 95. `groessenmarke/messung-pfad-fehlt`, `-pfad-zusaetzlich`, `-pfade-vertauscht` · Kopf mit einer `pfad`-Zeile weniger, mehr und in anderer Reihenfolge als der Block (zwei Pfade im Block) · Exit 2, wie Fall 92.
96. `groessenmarke/messung-kopf-fehlt` · Messung ohne Kopf, nur Datensätze · Exit 2, stderr nennt `Messung nicht lesbar`.
97. `groessenmarke/messung-ohne-numstat-zeile` · Kopf ohne Zeile `numstat` · Exit 2, wie Fall 96.
98. `groessenmarke/messung-datensatz-ohne-nul` · letzter Datensatz ohne abschließendes NUL · Exit 2, wie Fall 96.
99. `groessenmarke/messung-datensatz-ohne-tab` · Datensatz ohne Tab · Exit 2, wie Fall 96.
100. `groessenmarke/aufruf` · ohne Argumente · Exit 2, stderr nennt `Aufruf:`.
101. `groessenmarke/plan-fehlt` · `pruefen` mit einem Pfad, der nicht existiert · Exit 2, stderr nennt `Plan fehlt`.
102. `groessenmarke/messung-fehlt` · `pruefen` mit vorhandenem Plan und fehlender Messung · Exit 2, stderr nennt `Messung fehlt`.
103. `groessenmarke/treiber-aufrufe-von-git` · Treiber mit beiden Ersatz-Skripten; Plan mit vollständigem Block und den Pfaden `src/a` und `src/b` · die Datei des Ersatzes für `git` zeigt genau drei Aufrufe in dieser Reihenfolge: `rev-parse --verify HEAD`, `merge-base --is-ancestor <basis> HEAD`, `diff --numstat -z --no-renames <basis>..HEAD -- :(literal)src/a :(literal)src/b :(exclude,literal)harness/vertraege`; der Ersatz des Programms zeigt `abfrage` und `pruefen` je einmal und schreibt je Kommando einen Text auf stdout und einen auf stderr; der Treiber reicht stdout und stderr von `pruefen` und stderr von `abfrage` an denselben Strom weiter und gibt stdout von `abfrage` nicht aus (`SPEC-128` Nr. 7 c) · Exit 0.
104. `groessenmarke/treiber-pfad-argumente` · Pfade mit Leerzeichen, `*`, `"`, `[x]` und einem Umlaut im Block · der Ersatz für `git` zeigt je Pfad ein Argument `:(literal)<Pfad>`, byte-gleich.
105. `groessenmarke/treiber-basis-kein-vorfahr` · Ersatz für `git` beendet `merge-base` mit Exit 1 · Exit 2, Ausgabe nennt `ist kein Vorfahr von HEAD`; der Ersatz des Programms zeigt keinen Aufruf von `pruefen`, der Ersatz für `git` keinen von `diff`.
106. `groessenmarke/treiber-basis-unbekannt` · `merge-base` mit Exit 128 · Exit 2, Ausgabe nennt `unbekannt oder git scheitert`; sonst wie Fall 105.
107. `groessenmarke/treiber-head-nicht-lesbar` · `rev-parse` mit Exit 128 · Exit 2, Ausgabe nennt `HEAD nicht lesbar`; kein `merge-base`, kein `diff`, kein `pruefen`.
108. `groessenmarke/treiber-git-diff-scheitert` · `diff` mit Exit 128 und einer Meldung · Exit 2, Ausgabe nennt `git diff scheitert` und die Meldung; kein Aufruf von `pruefen`.
109. `groessenmarke/treiber-git-nicht-aufrufbar` · `PATH` ohne `git` · Exit 2, Ausgabe nennt `git nicht aufrufbar`; kein Aufruf von `pruefen`.
110. `groessenmarke/treiber-keine-plaene` · leeres Planverzeichnis · Exit 0, Ausgabe nennt `kein Slice in Arbeit`; der Ersatz des Programms und der für `git` zeigen keinen Aufruf. Dasselbe ein Verzeichnis mit `roadmap.md`, `README.md`, `slice-c.txt` und einem Unterverzeichnis mit `slice-d.md`, aber ohne Datei `slice-*.md`. Ein Planverzeichnis, das fehlt: Exit 2. In jedem Fall mit mindestens einem Plan (103 bis 109, 111 bis 120) nennt die Ausgabe `kein Slice in Arbeit` nicht.
111. `groessenmarke/treiber-plan-suche` · Planverzeichnis mit `slice-b.md`, `slice-a.md`, `roadmap.md`, `README.md`, `slice-c.txt` und einem Unterverzeichnis mit `slice-d.md` · der Ersatz des Programms zeigt `abfrage` genau für `slice-a.md`, dann `slice-b.md`.
112. `groessenmarke/treiber-exit-aggregation` · drei Pläne, deren Ersatz-Programm 0, 1 und 2 liefert, und die Paare 0/1 und 0/0 · Exit 2 bei einer 2, 1 bei einer 1 und keiner 2, sonst 0; der Ersatz zeigt für jede Kombination alle Pläne gefahren. Ein Plan, dessen Ersatz von `pruefen` jeden Exit von 0 bis 255 liefert (256 Läufe): 0 und 1 bleiben, 2 bis 255 ergeben Exit 2. Ein Exit von `abfrage` ungleich 0 (1, 2, 3, 125, 127, 255) ergibt Exit 2 ohne Aufruf von `git` und `pruefen`; `abfrage` mit Exit 0 ohne Ausgabe ergibt Exit 2 und `abfrage ohne Ergebnis`.
113. `groessenmarke/treiber-ohne-block-kein-git` · Pläne ohne Block, mit Prosa-Form und mit ausstehender Marke · Exit 0, der Ersatz für `git` zeigt keinen Aufruf, die Ausgabe trägt die Hinweise.
114. `groessenmarke/treiber-aufraeumen` · eigenes `TMPDIR`; ein Lauf mit Exit 0, einer mit Fehler von `git diff`, und je ein Lauf, in dem der Ersatz von `abfrage`, von `git diff` oder von `pruefen` dem Treiber das Signal HUP, INT oder TERM sendet (9 Läufe, Exit 2) · danach ist das Verzeichnis leer.
115. `groessenmarke/treiber-git-ausschluesse` · echtes `git`, Pfade des Blocks `tools/x/`, `tools/x` und `harness/`; der zweite Commit ändert je eine Datei mit 3 Einfügungen in `tools/x/`, `harness/mk/`, `harness/vertraege/`, `spec/`, `docs/` und `src/` · Ausgabe nennt `Messung 6` (`tools/x/` und `tools/x` zählen die Datei einmal).
116. `groessenmarke/treiber-git-sonderzeichen` · echtes `git`, Dateien `a b.txt`, `a*`, `ab`, `[x].txt`, `ä.txt` und `q"r.txt` mit je 2 Einfügungen, Pfade des Blocks `a*`, `a b.txt`, `[x].txt`, `ä.txt` und `q"r.txt` · Ausgabe nennt `Messung 10` (`ab` zählt nicht: `a*` ist ein Name, kein Muster).
117. `groessenmarke/treiber-git-umbenennung` · echtes `git`, `diff.renames=true` in der Konfiguration des Repos, eine Datei mit 10 Zeilen im Pfad wird umbenannt, ohne Änderung · Ausgabe nennt `Messung 20` (Löschung und Einfügung).
118. `groessenmarke/treiber-git-binaer` · echtes `git`, eine Binärdatei (Bytes mit NUL) und eine Textdatei mit 4 Einfügungen im Pfad · Ausgabe nennt `Messung 4` und `1 Binärdatei(en) nicht gezählt`.
119. `groessenmarke/treiber-git-basis-kein-vorfahr` · echtes `git`, `basis` ist ein Commit eines anderen Zweigs · Exit 2, Ausgabe nennt `ist kein Vorfahr von HEAD`.
120. `groessenmarke/treiber-git-basis-gleich-spitze` · echtes `git`, `basis` ist `HEAD` (als Kurzform des Hashs), Marke gleich Formel · Exit 0, Ausgabe nennt `Messung 0`.
121. bis 130. `handbuch/h1-sauber` bis `handbuch/h10-zitat` (die Fälle H1 bis H10 der Gegenprobe von `SPEC-200`; die Nummern sind vorläufig, der Integrationsschritt vergibt sie beim Zusammenführen) · `pruefung handbuch <Binary> <Wurzel>` im Wegwerfverzeichnis mit Handbuch, Beispiel-Datei, `Makefile`, `Dockerfile` und einem Ersatz-Skript als Binary, das jeden Start in einer Datei zählt, auf `version` mit `odata-recorder dev` und sonst mit seiner Ausgabe antwortet; „nennt“ liest den Strom, den `SPEC-200` Nr. 2 b der Meldung zuweist (stdout: `ok`, stderr: Befunde und Eingabefehler):
    - 121 `h1-sauber`: sauberes Handbuch, ein gebundenes Beispiel, das Ersatz-Skript liefert die Datei · Exit 0, stdout nennt `pruefung handbuch: ok`; das Skript wurde zweimal gestartet (`version` und die Datei).
    - 122 `h2-planung`, 123 `h3-chronik`, 124 `h4-kennung`: ein Satz mit `geplant`, `bisher` bzw. `spec-110` in der Prosa · Exit 1, stderr nennt `Planung A1: geplant`, `Chronik B1: bisher` bzw. `Kennung C6: spec-110`.
    - 125 `h5-version-weicht-ab`: `Software-Version: 1.0` im Kopf, das Skript antwortet `dev` · Exit 1, stderr nennt `Software-Version 1.0 weicht von der Ausgabe von version ab (dev)`.
    - 126 `h6-binary-weicht-ab`: das Skript antwortet nicht mit der Datei · Exit 1, stderr nennt `aufruf.txt: Standardausgabe weicht ab` und nicht `Beispiel weicht von`.
    - 127 `h7-block-weicht-ab`: der Block ist die Datei mit einer Zeile mehr · Exit 1, stderr nennt `Beispiel weicht von beispiele/aufruf.txt ab`.
    - 128 `h8-make-ziel`: `make nix` in Inline-Code, das `Makefile` führt nur `build` · Exit 1, stderr nennt `make-Ziel nix unbekannt`.
    - 129 `h9-binary-fehlt`: der Pfad des Binary zeigt auf keine Datei · Exit 2, stderr nennt `Binary fehlt` (der Fall zählt keine Starts; dass ein Eingabefehler das Binary nicht startet, fangen die Tests von `pruefung`, `SPEC-200` Nr. 2 c).
    - 130 `h10-zitat`: die Datei, der Block und die Ausgabe des Skripts tragen das Wort `geplant` · Exit 0, stdout nennt `pruefung handbuch: ok` (das Zitat gilt nicht als Prosa).
    Der Inhalt der Dateien führt kein Zeichen jenseits von ASCII außer dem Umlaut in `Gültigkeitsbereich`, den der Fall aus Bytewerten bildet. Die Fälle fangen die Verdrahtung der Stufe `handbuch` und des Ziels nicht (`SPEC-200` Nr. 14 h).
Die erste Spalte ist der Name im Bericht; ein Name kommt genau einmal vor.
Das Rennen der Fälle 35 und 36 ist nicht zeitabhängig (`SPEC-122` Nr. 8).
Gemessen am 2026-10-22 mit dem Testbinary der Fixtures unter `-race`, je Lauf ein eigener Prozess mit `-test.count=1`: das Rennen jedes der drei Teilbäume von Fall 35 wurde in 200 von 200 Läufen gemeldet, die Fixture von Fall 34 blieb in 200 von 200 Läufen ohne Meldung, Fall 36 mit beiden Meldungen (`WARNING: DATA RACE` und `test timed out`) in 100 von 100.

**Randformen:**
1. Eine Eingabe fehlt in der Eingabekopie (etwa `.a-check.yml`) oder die Kopie scheitert: Meldung `gegenproben: <Grund>`, Exit-Code 2, kein Fall läuft.
2. Ein Programm (`a-check`, `golangci-lint`, `sh`, `pruefung`, `go`, `gofmt`, `groessenmarke`, `git`) ist nicht aufrufbar: Meldung `gegenproben: <Programm> nicht aufrufbar`, Exit-Code 2. Für `git` gilt das nur, solange die Fälle 115 bis 120 es verlangen; Fall 109 entfernt `git` aus `PATH` selbst.
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
12. Die Fälle 103 bis 120 setzen `TMPDIR` auf ein eigenes Verzeichnis und lesen das Protokoll der Ersatz-Skripte nach dem Lauf; ein Fall, dessen Protokoll fehlt, ist rot. Die Ersatz-Skripte halten jedes Argument fest und zählen jeden Aufruf, auch die Operationen, die der Treiber nie rufen darf (`diff` nach einem Fehler von `merge-base`, `pruefen` nach einem Fehler von `git`).
13. Die Hülle, die einen Fall als Prozess startet, liefert den Fällen 1 bis 37 stdout und stderr gemeinsam und den Fällen 38 bis 130 getrennt. Eine Erwartung „nennt“ der Fälle 38 bis 130 prüft den Strom, den `SPEC-128` der Meldung zuweist (Absatz vor Fall 38); steht der Text nur im anderen Strom, ist der Fall rot. „Nennt nicht“ gilt für beide Ströme zusammen. Exit-Code und Zeitgrenze (Randform 6) sind in beiden Formen dieselben. Ein Test der Hülle (Muster: der Test der Ersatz-Skripte, Randform 12) startet einen Prozess, der auf stdout und stderr je einen Text schreibt und mit Exit 3 endet, und prüft die getrennte Lieferung beider Ströme, den Exit-Code und die gemeinsame Lieferung der Fälle 1 bis 37.

Die Fälle 25 bis 28 rufen `go` im Wegwerfverzeichnis auf und teilen den Cache von `go` (`GOCACHE`) wie den von golangci-lint; die Fälle 29 bis 32 bauen das Binary mit `go build` in der Kopie, nicht im Image der Stufe `build`.
Weicht dieser Bau vom Bau des Image ab (Zielplattform, Flags), fängt die Gegenprobe die Abweichung nicht.
Die Fälle 34 bis 36 rufen `pruefung race`, das `go` mit cgo ruft; unter `-race` baut `go` die Standardbibliothek mit Instrumentierung neu, im kalten `GOCACHE` der Stufe und geteilt von den Fällen.
Weichen Compiler oder Go-Version des Images der Stufe `gegenproben` von denen des Images `RACE_IMAGE` ab, fängt die Gegenprobe die Abweichung nicht; trägt das Image keinen Compiler mehr (das golangci-lint-Image sagt ihn nicht zu), sind die Fälle rot, mit der Meldung von `go` im Bericht, und der Wechsel des Pins ändert die Prüfsumme und erzwingt den Lauf.
Dass das Ziel `src/` ganz und den Quellstand der Stufe `source` fährt, belegt `make gates`, nicht die Gegenprobe.

**Zeitbudget** (Wanduhr, `time make gegenproben`): bei geänderten Eingaben und kaltem Docker-Build-Cache höchstens 120 s, bei unveränderten Eingaben höchstens 1 s.
Gemessen am 2026-10-22 mit den Fällen 1 bis 36 auf einem Host mit 20 Kernen und schwankender Last, Images lokal vorhanden: 80,5 bis 93,8 s im Ganzen (`time make gegenproben`).
Davon laufen die Fälle (Schritt `RUN gegenproben` der Stufe) 54,3 bis 67 s, das Prüfen und Bauen des Programms rund 7 s und der Export des Images der Stufe 16,8 bis 18 s.
Bei unveränderten Eingaben dauert der Aufruf 0,3 s.
Das Budget liegt rund 28 % über dem höchsten gemessenen Wert, 93,8 s; kein Gate und kein Hook prüft es.
Liegt ein Lauf darüber, misst der Implementer neu und trägt den Wert mit Grund ein.
Die Fälle 38 bis 120 starten je einen Prozess ohne Compiler (das Programm, `sh`, die Ersatz-Skripte, in den Fällen 115 bis 120 `git`); gemessen am 2026-10-30 auf einem Host mit 20 Kernen und fremder Last (Load Average 17 bis 69): die 83 Fälle allein brauchen 0,20 und 0,23 s, die 37 übrigen 45,1 und 44,5 s, alle 120 44,4 und 45,8 s, das Programm der Gegenproben dabei in einem Container mit `--cpus=4`; `time make gegenproben` bei geänderter Eingabe dauerte 143 s mit den 37 Fällen (Schritt `RUN gegenproben` 94,9 s) und 127 s mit den 120 (84,5 s). Die Zusatzzeit der Fälle 38 bis 120 liegt unter der Streuung. Beide Läufe von `make gegenproben` liegen über dem Budget, bei einer Last von mehr als dem Dreifachen der Kernzahl; ein Lauf auf ruhigem Host steht aus. Dazu kommt der Bau des Programms der Stufe `groessenmarke` (`gofmt`, `go vet`, `go test`, Bau): 10,6 bis 11,6 s im Schritt `RUN` bei kaltem Cache, parallel zu den übrigen Stufen. Die Schleife über 256 Exit-Codes in Fall 112 und die 9 Signal-Läufe in Fall 114 kommen hinzu (rund 265 Läufe des Treibers mit je einer Handvoll Prozessen); gemessen am 2026-10-30 auf demselben Host mit dem Programm der Gegenproben in einem Container mit `--cpus=4` bei einem Load Average von 4 bis 6: die 83 Fälle allein brauchen 2,70 bis 2,90 s (vorher 0,23 bis 0,26 s, je drei Läufe), die Schleifen also rund 2,5 s; alle 120 Fälle 46,4 und 46,9 s (vorher 45,9 und 46,0 s, je zwei Läufe, Load Average 4 bis 46), weil die Fälle 38 bis 120 neben den langen Fällen 1 bis 37 laufen; `time make gegenproben` bei geänderter Eingabe 114,9 s (Schritt `RUN gegenproben` 76,1 s, Export des Images 26,7 s, Load Average 67 am Ende des Laufs). Das Budget von 120 s hält in dieser einen Messung; die Zusatzzeit der Fälle 38 bis 120 liegt bei rund 2,5 s allein und unter der Streuung im Ganzen.

**Pflicht zum Fahren:** `make gegenproben` läuft bei jeder Änderung am geprüften Gate, am Profil, an der Architektur-Regel, an der Gegenprobe, am Programm `groessenmarke` oder seinem Treiber, an diesem Vertrag oder an `SPEC-128`, am Ende jeder Welle und vor dem Release.

**Stand:** vorhanden für `make a-check`, `make lint`, `make migration-check`, `make test`, `make build`, `make test-race`, `make groessenmarke` (die Fälle 38 bis 120) und `make handbuch-check` (die Fälle 121 bis 130).

**Grenze:** Kein Gate und kein Hook liest den Stempel; er belegt nur, dass ein Lauf zu dieser Prüfsumme grün war, und wer ihn von Hand schreibt, belegt nichts.
Die Fälle prüfen die Importrichtungen `core-impurity` und Adapter auf Adapter sowie die Direktiv- und `# Why:`-Regeln; die Bibliotheksgrenzen halten Fall 20 (`modernc.org/sqlite` in einem fremden Adapter) und Fall 37 (`encoding/xml` im Kern, die Regel der `tech`-Liste von `.a-check.yml`); die Seite „erlaubt“ von `encoding/xml` im Paket `src/adapters/driven/odata/csdl` hält Fall 1, der den Baum von `src/` fährt; weitere Verstoßformen kommen mit weiteren Bibliotheksgrenzen.
Das Programm selbst prüft die Stufe mit `gofmt`, `go vet` und `go test`; `make lint` prüft es als zweites Modul (`SPEC-101`).
Die Liste der Eingaben in `tools/harness/gegenproben-eingabe.sh` steht einmal im Skript; dass sie `tools/pruefung` nennt, fängt kein Fall — wer den Eintrag entfernt, lässt die Prüfsumme bei einer Änderung an `pruefung` gleich; dasselbe gilt für `tools/groessenmarke` und `tools/harness/groessenmarke.sh`.
Die Fälle zu `make test` und `make build` prüfen `pruefung` mit den Verstoßformen aus der Liste; sie fangen nicht, dass ein Ziel `pruefung` nicht ruft (das Gate ist im `Dockerfile` verdrahtet).
Die Fälle 38 bis 120 fangen weder den Aufruf von `docker` in `tools/harness/groessenmarke-docker.sh` noch die Verdrahtung von Stufe und Ziel (`SPEC-128` Nr. 10 h); die Argumente von `git` nach `SPEC-128` Nr. 4 b fängt der Ersatz in den Fällen 103 und 104, die Lesart dieser Argumente durch das echte `git` die Fälle 115 bis 120. Weicht das `git` des Hosts vom Image ab (2.47.3), fängt keine Gegenprobe die Abweichung.
Die Gegenproben von `docs-check`, `schema-check` und `baseline-verify` fehlen.

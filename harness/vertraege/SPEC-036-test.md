# SPEC-036 — Build und Unit-Tests (`make build`, `make test`)

**Werkzeug:** `make build` und `make test`, beide über das `Dockerfile`; Eingangs-Image per Digest gepinnt, alle Stufen außer `deps` ohne Netz.

**Festlegung `make build`:** Baut das Binary aus `src/bootstrap/main` mit `-trimpath` und ohne VCS-Stempel; es liegt im Image `odata-recorder:build` unter `/out/odata-recorder`.
Die Version kommt aus der Make-Variablen `VERSION` (Vorgabe `dev`).
Das Ziel schlägt fehl, wenn `pruefung build` (siehe unten) das Binary beanstandet: `odata-recorder version` gibt nicht `odata-recorder <VERSION>` aus, oder die Build-Information des Binary trägt nicht `-trimpath=true`.
Die Prüfung läuft in der Stufe `build` des `Dockerfile`, unmittelbar nach `go build`.

**Festlegung `make test`:** Schlägt fehl, wenn `pruefung test` (siehe unten) einen Befund meldet: `gofmt -l` nennt eine Datei, `go vet` meldet einen Befund oder ein Unit-Test ist rot.
Die Prüfung läuft in der Stufe `test` des `Dockerfile`.
Der Fachumfang (Domain, Matching, Query-Interpreter, CRUD) wächst mit dem Code.

**Prüfprogramm `pruefung`** (Quelle `tools/pruefung/`, eigenes Go-Modul ohne Abhängigkeiten; das `Dockerfile` baut es in einer eigenen Stufe, prüft es mit `gofmt`, `go vet` und `go test` und gibt es an die Stufen `test`, `build`, `race`, `handbuch` (über die Stufe `build`) und `gegenproben`).
Aufrufe:
- `pruefung test <Wurzel> [Muster...]`: arbeitet im Verzeichnis `Wurzel`; ohne Muster gilt `./...`.
  Es führt der Reihe nach `gofmt -l .`, `go vet <Muster>` und `go test -trimpath -buildvcs=false <Muster>` aus und hält beim ersten Befund an.
  Befunde (Exit 1): `gofmt` nennt mindestens eine Datei — Text `nicht gofmt-formatiert:` und die Dateien; `gofmt` endet ungleich 0 (etwa bei einem Syntaxfehler) — Text `gofmt:` und dessen Ausgabe; `go vet` endet ungleich 0 — Text `go vet:` und dessen Ausgabe; `go test` endet ungleich 0 — Text `go test:` und dessen Ausgabe.
  Ohne Befund: Text `pruefung test: ok`, Exit 0.
- `pruefung build <Binary> <Version>`: prüft das Binary an `Binary`.
  Es ruft `<Binary> version`; gibt das nach Entfernen der abschließenden Zeilenumbrüche nicht genau `odata-recorder <Version>` aus, ist das ein Befund (Exit 1), Text `Version des Binary ist '<Ausgabe>', erwartet 'odata-recorder <Version>'`.
  Es ruft `go version -m <Binary>`; trägt die Ausgabe keine Zeile mit `-trimpath=true`, ist das ein Befund (Exit 1), Text `das Binary ist nicht mit -trimpath gebaut`.
  Ohne Befund: Text `pruefung build: ok`, Exit 0.
- `pruefung race <Wurzel> [Muster...]`: fährt `go test -race` mit cgo; Vertrag und Randformen in `SPEC-122`.
- `pruefung handbuch <Binary> <Wurzel>`: prüft das Benutzerhandbuch gegen das Binary; Vertrag und Randformen in `SPEC-129`, dort stehen auch seine Abweichungen von den Randformen 3, 4, 10, 12 und 14.

**Randformen von `pruefung`:**
1. Keine Argumente, ein unbekanntes Kommando, ein leeres Argument oder die falsche Argumentzahl (`test` oder `race` ohne `Wurzel`, `build` mit anderem als genau `Binary` und `Version`, `handbuch` mit anderem als genau `Binary` und `Wurzel`): Text `Aufruf:` und alle Formen, Exit 2, nichts wird gestartet.
2. `Wurzel` fehlt oder ist kein Verzeichnis (`test`, `race`): Text `Wurzel fehlt: <Pfad>`, Exit 2.
3. `Binary` fehlt: Text `Binary fehlt: <Pfad>`, Exit 2.
4. `Binary` ist keine ausführbare Datei, startet nicht oder endet bei `version` mit Exit-Code ungleich 0 (auch bei richtiger Ausgabe): Text `Binary startet nicht:` und der Grund, Exit 1.
5. `Binary` ist kein Go-Binary (`go version -m` endet ungleich 0): Text `Build-Information des Binary nicht lesbar:` und die Ausgabe, Exit 1.
6. `Binary` gibt die Version mit Zusatztext oder auf stderr statt stdout aus: Befund der Versionsprüfung; nur stdout zählt. Entfernt werden nur abschließende `\n`; `odata-recorder <Version>\r\n` ist ein Befund.
7. Ein Programm (`gofmt`, `go`) ist nicht aufrufbar: Text `<Programm> nicht aufrufbar`, Exit 2.
8. Ein Verzeichnis ohne Go-Dateien: kein Befund von `pruefung`, solange `go vet` und `go test` Exit 0 melden.
9. Eine nicht lesbare Datei in `Wurzel` ist ein Befund von `gofmt` (Randform 1 der Befunde); sie wird nicht von einer Gegenprobe gefangen, weil der Prüflauf mit Administratorrechten läuft.
10. Abbruch und Herunterfahren: `pruefung` hält keinen Zustand und schreibt keine Datei; ein Abbruch des Builds lässt nichts zurück. `pruefung handbuch` legt ein Arbeitsverzeichnis an (`SPEC-129` Nr. 2 e).
11. Streams: Befunde und Fehler gehen nach stderr, die Zeile `pruefung <Kommando>: ok` nach stdout.
12. Zeit: `pruefung` setzt kein Zeitlimit; ein hängendes `go test`, `go vet` oder `<Binary> version` hält den Lauf, bis der Aufrufer ihn abbricht. `pruefung handbuch` setzt 10 s je Lauf des Binary (`SPEC-129` Nr. 7 c).
13. Argumente, die mit `-` beginnen, gelten als Muster und gehen unverändert an `go`; `pruefung` prüft sie nicht selbst.
14. Exit-Codes: 0 kein Befund, 1 mindestens ein Befund, 2 Aufruf, `Wurzel` oder `Binary` fehlt oder ein Programm ist nicht aufrufbar.

**Randformen von `bootstrap.Run`:**
1. Keine Argumente oder genau `version`: Ausgabe `odata-recorder <Version>` nach stdout, Exit-Code 0.
2. Jedes andere Aufrufmuster — unbekanntes Kommando, leeres Argument, Flag wie `--help`, `version` mit Zusatzargument: `Unbekanntes Kommando: <erstes Argument>` nach stderr, Exit-Code 2.
3. Ein nil-Writer verwirft die Ausgabe; der Exit-Code bleibt unverändert.

**Stand:** vorhanden für das Skelett aus Composition Root und Version; `pruefung` trägt die Prüfungen von `make test`, `make build` und `make test-race` (`SPEC-122`), ihre Gegenproben liefert `SPEC-102`.

**Grenze:** Dass die Stufen `test`, `build` und `race` des `Dockerfile` `pruefung` rufen und dass die Stufe `pruefung` sich selbst mit `gofmt`, `go vet` und `go test` prüft, fängt kein Fall und kein Test; die Verdrahtung ist im `Dockerfile` und wird nur beim Lesen geprüft.
Prüft kein Zusammenspiel mit Upstream oder Datenbank; das tun `make test-integration` und die Verträge `SPEC-037` und `SPEC-042`.

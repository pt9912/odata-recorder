# SPEC-036 — Build und Unit-Tests (`make build`, `make test`)

**Werkzeug:** `make build` und `make test`, beide über das `Dockerfile`; Eingangs-Image per Digest gepinnt, alle Stufen außer `deps` ohne Netz.

**Festlegung `make build`:** Baut das Binary aus `src/bootstrap/main` mit `-trimpath` und ohne VCS-Stempel; es liegt im Image `odata-recorder:build` unter `/out/odata-recorder`.
Die Version kommt aus der Make-Variablen `VERSION` (Vorgabe `dev`).
Das Ziel schlägt fehl, wenn `pruefung build` (siehe unten) das Binary beanstandet: `odata-recorder version` gibt nicht `odata-recorder <VERSION>` aus, oder die Build-Information des Binary trägt nicht `-trimpath=true`.
Die Prüfung läuft in der Stufe `build` des `Dockerfile`, unmittelbar nach `go build`.

**Festlegung `make test`:** Schlägt fehl, wenn `pruefung test` (siehe unten) einen Befund meldet: `gofmt -l` nennt eine Datei, `go vet` meldet einen Befund oder ein Unit-Test ist rot.
Die Prüfung läuft in der Stufe `test` des `Dockerfile`.
Der Fachumfang (Domain, Matching, Query-Interpreter, CRUD) wächst mit dem Code.

**Prüfprogramm `pruefung`** (Quelle `tools/pruefung/`, eigenes Go-Modul ohne Abhängigkeiten; das `Dockerfile` baut es in einer eigenen Stufe, prüft es mit `gofmt`, `go vet` und `go test` und gibt es an die Stufen `test`, `build` und `gegenproben`).
Aufrufe:
- `pruefung test <Wurzel> [Muster...]`: arbeitet im Verzeichnis `Wurzel`; ohne Muster gilt `./...`.
  Es führt der Reihe nach `gofmt -l .`, `go vet <Muster>` und `go test -trimpath -buildvcs=false <Muster>` aus und hält beim ersten Befund an.
  Befunde (Exit 1): `gofmt` nennt mindestens eine Datei — Text `nicht gofmt-formatiert:` und die Dateien; `gofmt` endet ungleich 0 (etwa bei einem Syntaxfehler) — Text `gofmt:` und dessen Ausgabe; `go vet` endet ungleich 0 — Text `go vet:` und dessen Ausgabe; `go test` endet ungleich 0 — Text `go test:` und dessen Ausgabe.
  Ohne Befund: Text `pruefung test: ok`, Exit 0.
- `pruefung build <Binary> <Version>`: prüft das Binary an `Binary`.
  Es ruft `<Binary> version`; gibt das nach Entfernen der abschließenden Zeilenumbrüche nicht genau `odata-recorder <Version>` aus, ist das ein Befund (Exit 1), Text `Version des Binary ist '<Ausgabe>', erwartet 'odata-recorder <Version>'`.
  Es ruft `go version -m <Binary>`; trägt die Ausgabe keine Zeile mit `-trimpath=true`, ist das ein Befund (Exit 1), Text `das Binary ist nicht mit -trimpath gebaut`.
  Ohne Befund: Text `pruefung build: ok`, Exit 0.

**Randformen von `pruefung`:**
1. Keine Argumente, ein unbekanntes Kommando, ein leeres Argument oder die falsche Argumentzahl (`test` ohne `Wurzel`, `build` mit anderem als genau `Binary` und `Version`): Text `Aufruf:` und die beiden Formen, Exit 2, nichts wird gestartet.
2. `Wurzel` fehlt oder ist kein Verzeichnis: Text `Wurzel fehlt: <Pfad>`, Exit 2.
3. `Binary` fehlt: Text `Binary fehlt: <Pfad>`, Exit 2.
4. `Binary` ist keine ausführbare Datei, startet nicht oder endet bei `version` mit Exit-Code ungleich 0 (auch bei richtiger Ausgabe): Text `Binary startet nicht:` und der Grund, Exit 1.
5. `Binary` ist kein Go-Binary (`go version -m` endet ungleich 0): Text `Build-Information des Binary nicht lesbar:` und die Ausgabe, Exit 1.
6. `Binary` gibt die Version mit Zusatztext oder auf stderr statt stdout aus: Befund der Versionsprüfung; nur stdout zählt.
7. Ein Programm (`gofmt`, `go`) ist nicht aufrufbar: Text `<Programm> nicht aufrufbar`, Exit 2.
8. Ein Verzeichnis ohne Go-Dateien: kein Befund von `pruefung`, solange `go vet` und `go test` Exit 0 melden.
9. Eine nicht lesbare Datei in `Wurzel` ist ein Befund von `gofmt` (Randform 1 der Befunde); sie wird nicht von einer Gegenprobe gefangen, weil der Prüflauf mit Administratorrechten läuft.
10. Abbruch und Herunterfahren: `pruefung` hält keinen Zustand und schreibt keine Datei; ein Abbruch des Builds lässt nichts zurück.
11. Exit-Codes: 0 kein Befund, 1 mindestens ein Befund, 2 Aufruf, `Wurzel` oder `Binary` fehlt oder ein Programm ist nicht aufrufbar.

**Randformen von `bootstrap.Run`:**
1. Keine Argumente oder genau `version`: Ausgabe `odata-recorder <Version>` nach stdout, Exit-Code 0.
2. Jedes andere Aufrufmuster — unbekanntes Kommando, leeres Argument, Flag wie `--help`, `version` mit Zusatzargument: `Unbekanntes Kommando: <erstes Argument>` nach stderr, Exit-Code 2.
3. Ein nil-Writer verwirft die Ausgabe; der Exit-Code bleibt unverändert.

**Stand:** vorhanden für das Skelett aus Composition Root und Version; `pruefung` trägt die Prüfungen beider Ziele, ihre Gegenproben liefert `SPEC-102`.

**Grenze:** Prüft kein Zusammenspiel mit Upstream oder Datenbank; das tun `make test-integration` und die Verträge `SPEC-037` und `SPEC-042`.

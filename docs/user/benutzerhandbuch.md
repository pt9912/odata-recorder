# Benutzerhandbuch OData Recorder

Software-Version: dev
Gültigkeitsbereich: das Binary, das `make build` erzeugt
Stand: 2026-10-30

## Einleitung

Dieses Handbuch richtet sich an Anwendungsentwickler und die Testautomatisierung (QA), die den OData Recorder aufrufen, und an das Plattform- und CI-Team, das ihn baut.

Voraussetzungen: ein Linux-Rechner mit Docker und GNU `make`.
## Bauen

Aufgabe: das Binary aus dem Quellstand bauen.

Voraussetzung: Docker und GNU `make` sind installiert, und das aktuelle Verzeichnis ist die Wurzel des Repos.

Vorgehen:

1. Rufe `make build` auf.
2. Setze die Software-Version mit der Variable `VERSION`, zum Beispiel mit `make build VERSION=1.2.3`. Ohne Angabe ist die Version `dev`.

Ergebnis: Das Image `odata-recorder:build` enthält das Binary unter `/out/odata-recorder`.

## Aufruf und Version

Aufgabe: das Binary aufrufen und die Software-Version lesen.

Das Binary läuft im Image. Der Aufruf lautet `docker run --rm odata-recorder:build /out/odata-recorder`, danach steht das Kommando.

Das erste Beispiel ruft das Binary ohne Argument auf, das zweite mit dem Kommando `version`. Beide zeigen die Standardausgabe, die Fehlerausgabe und den Ausgangswert.

Datei: [beispiele/aufruf-ohne-argument.txt](beispiele/aufruf-ohne-argument.txt)
```
$ docker run --rm odata-recorder:build /out/odata-recorder
--- Standardausgabe
odata-recorder dev
--- Fehlerausgabe
--- Ausgangswert
0
```

Datei: [beispiele/aufruf-version.txt](beispiele/aufruf-version.txt)
```
$ docker run --rm odata-recorder:build /out/odata-recorder version
--- Standardausgabe
odata-recorder dev
--- Fehlerausgabe
--- Ausgangswert
0
```

## Fehlerbehebung

### Unbekanntes Kommando

Meldung: `Unbekanntes Kommando: <Argument>`

Lösung: Rufe das Binary ohne Argument oder nur mit `version` auf.

Das erste Beispiel zeigt ein Argument, das kein Kommando ist, das zweite `version` mit einem weiteren Argument; `<Argument>` in der Meldung steht im ersten Beispiel für `foo`, im zweiten für `version`.

Datei: [beispiele/fehler-unbekanntes-kommando.txt](beispiele/fehler-unbekanntes-kommando.txt)
```
$ docker run --rm odata-recorder:build /out/odata-recorder foo
--- Standardausgabe
--- Fehlerausgabe
Unbekanntes Kommando: foo
--- Ausgangswert
2
```

Datei: [beispiele/fehler-version-mit-zusatz.txt](beispiele/fehler-version-mit-zusatz.txt)
```
$ docker run --rm odata-recorder:build /out/odata-recorder version x
--- Standardausgabe
--- Fehlerausgabe
Unbekanntes Kommando: version
--- Ausgangswert
2
```

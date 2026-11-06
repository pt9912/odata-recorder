# Benutzerhandbuch OData Recorder

Software-Version: dev
Gültigkeitsbereich: das Binary, das `make build` erzeugt
Stand: 2026-11-06

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

## Fehlercodes

Dieser Abschnitt führt die Fehlercodes, die das Binary dem Anwender zeigt. Ein Fehlercode beginnt mit `ODR-E` und trägt vier Ziffern.

| Code | Bedeutung |
|---|---|
| `ODR-E2001` | Die Startkonfiguration ist ungültig. |
| `ODR-E2401` | Die Aufnahme oder die Schwärzung eines Austauschs ist fehlgeschlagen. |
| `ODR-E3001` | Gespeicherte Daten sind beschädigt oder nicht ladbar, etwa weil die Datei fremd ist oder von einer neueren Schemaversion stammt. |
| `ODR-E4001` | Der Upstream ist nicht erreichbar. |
| `ODR-E5001` | Kein aufgezeichneter Austausch passt zur Anfrage. |
| `ODR-E5002` | Alle aufgezeichneten Austausche sind verbraucht. |
| `ODR-E5003` | Die Zuordnung zu Workspace und Replay fehlt, ist leer oder mehrdeutig. |
| `ODR-E5004` | Die Zuordnung gehört zu einem anderen Service. |
| `ODR-E5005` | Für den Pfad ist kein Service konfiguriert. |
| `ODR-E5006` | Die Kodierung der Anfrage ist ungültig. |
| `ODR-E5007` | Der nächste erwartete Austausch hat einen anderen Match-Key. |
| `ODR-E5008` | Der gefundene Austausch lässt sich nicht abspielen. |
| `ODR-E5009` | Keine Aufnahme-Session des Services nimmt die Anfrage an. |
| `ODR-E5010` | Der Server ist angehalten und verarbeitet die Anfrage nicht. |
| `ODR-E7001` | Der Workspace ist unbekannt. |
| `ODR-E7003` | Die Aufnahme-Session ist unbekannt. |
| `ODR-E7009` | Die Session ist noch offen. |
| `ODR-E9001` | Ein interner Fehler ist aufgetreten. |

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

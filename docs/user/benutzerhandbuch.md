# Benutzerhandbuch OData Recorder

Software-Version: dev
Gültigkeitsbereich: das Binary, das `make build` erzeugt
Stand: 2026-11-08

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

## Kommandos

Dieser Abschnitt führt die Kommandos des Binary.

| Kommando | Bedeutung |
|---|---|
| `start` | Startet den Dienst mit den Einstellungen aus den Umgebungsvariablen. |
| `version` | Gibt die Software-Version aus. |

## Einstellungen

Dieser Abschnitt führt die Einstellungen, die der Dienst aus den Umgebungsvariablen liest. Jeder Schlüssel ist der Name einer Variablen.

| Schlüssel | Bedeutung |
|---|---|
| `MODE` | Betriebsart des Dienstes: `record` nimmt Austausche mit dem Upstream auf, `replay` spielt aufgezeichnete Austausche ab. `mock` beantwortet Anfragen allein aus dem Bestand, ohne Upstream. Ohne Angabe gilt `mock`; `hybrid` ist ein gültiger Wert, aber der Start lehnt ihn ab. |
| `LISTEN` | Host und Port, an dem der Dienst Anfragen annimmt, in der Form `host:port`. Ohne Angabe gilt `127.0.0.1:8080`. |
| `DATA_DIR` | Verzeichnis, in dem der Dienst seine Daten ablegt, absolut oder relativ zum Arbeitsverzeichnis. Ohne Angabe gilt `data`. |
| `SERVICE_ID` | Kennung des einzigen Services, 1 bis 64 Zeichen aus Buchstaben, Ziffern, `.`, `_` und `-`. Ohne Angabe gilt `default`. |
| `SERVICE_ROOT_PATH` | Pfad, unter dem der einzige Service Anfragen annimmt: `/` oder Segmente wie `/a/b` ohne abschließenden Schrägstrich. Ohne Angabe gilt `/`. |
| `UPSTREAM_URL` | URL des Upstreams, absolut mit `http` oder `https`, ohne Zugangsdaten, Query und Fragment. Der Modus `record` verlangt sie. |
| `SERVICES_FILE` | Pfad einer JSON-Datei, die mehrere Services beschreibt. Zugleich gesetzte `SERVICE_ID`, `SERVICE_ROOT_PATH` oder `UPSTREAM_URL` verhindern den Start. |
| `REPLAY_STRATEGY` | Strategie, mit der das Abspielen wiederholte gleiche Anfragen beantwortet: `sequential`, `strict`, `first` oder `last`. Ohne Angabe gilt `sequential`. |
| `REPLAY_ASSIGNMENT` | Wie eine Anfrage einem Workspace und einem Replay zugeordnet wird: `explicit` oder `ordered`. Ohne Angabe gilt `explicit`. |
| `REPLAY_REDACTED_BODIES` | Wahrheitswert `true` oder `false`: ob das Abspielen auch Austausche liefert, deren Antwort-Body geschwärzt ist. Ohne Angabe gilt `false`. |
| `MATCH_VARY_HEADERS` | Header-Namen, die in den Match-Key eingehen, durch Komma getrennt; `none` steht für keinen Header. Ohne Angabe gilt `Accept, OData-Version, Content-Type, If-Match, Prefer`. |
| `MATCH_INVALID_ENCODING` | Umgang mit einer ungültigen Kodierung der Anfrage beim Zuordnen: `literal` oder `reject`. Ohne Angabe gilt `literal`. |
| `MATCH_BODY_MODE` | Wie der Body einer Anfrage in den Match-Key eingeht: `bytes` oder `json-canonical`. Ohne Angabe gilt `bytes`. |
| `REDACT_HEADERS` | Zusätzliche Header-Namen, deren Werte der Dienst bei der Aufnahme schwärzt, durch Komma getrennt. Die Liste ergänzt die eingebaute Liste. |
| `REDACT_QUERY_PARAMS` | Zusätzliche Namen von Query-Parametern, deren Werte der Dienst schwärzt, durch Komma getrennt. Die Liste ergänzt die eingebaute Liste. |
| `REDACT_JSON_PATHS` | Zusätzliche JSON-Pfade der Form `$..name`, deren Werte der Dienst schwärzt, durch Komma getrennt. Die Liste ergänzt die eingebaute Liste. |
| `MAX_CAPTURE_BODY` | Größte Länge eines Bodys, die der Dienst aufnimmt: eine Zahl, optional mit `B`, `KiB`, `MiB` oder `GiB`, von 1 Byte bis 1 GiB. Ohne Angabe gilt `8 MiB`. |
| `UPSTREAM_TIMEOUT` | Wartezeit bis zum Kopf der Antwort des Upstreams: eine Zahl mit der Einheit `ms`, `s`, `m`, `h` oder `d`. Ohne Angabe gilt `30s`. |
| `UPSTREAM_CA_FILE` | Pfad einer PEM-Datei mit Zertifikaten, denen der Dienst beim Upstream zusätzlich vertraut. |
| `UPSTREAM_TLS_VERIFY` | Wahrheitswert: ob der Dienst das Zertifikat des Upstreams prüft. Ohne Angabe gilt `true`; `false` verlangt `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true`. |
| `UPSTREAM_TLS_INSECURE_SKIP_VERIFY` | Wahrheitswert: `true` schaltet die Prüfung des Zertifikats des Upstreams ab. Ohne Angabe gilt `false`; `true` widerspricht `UPSTREAM_TLS_VERIFY=true`. |
| `SHUTDOWN_TIMEOUT` | Frist, die laufende Austausche beim Beenden bekommen: eine Dauer wie bei `UPSTREAM_TIMEOUT`. Ohne Angabe gilt `10s`. |
| `RETENTION_MAX_AGE` | Löschfrist für Sessions: eine Dauer wie bei `UPSTREAM_TIMEOUT`, `0` schaltet sie aus. Ohne Angabe gilt `0`. |
| `SQLITE_BUSY_TIMEOUT` | Wartezeit der Datenbank bei gleichzeitigem Zugriff: eine Dauer wie bei `UPSTREAM_TIMEOUT`, höchstens 2147483647 ms. Ohne Angabe gilt `5000ms`. |
| `VOLATILE_HEADERS` | Header-Namen, die der Vergleich der Antworten übergeht, durch Komma getrennt; `none` steht für keinen Header. Ohne Angabe gilt `Date, Server, Via, Age, X-Request-Id, X-Correlation-Id, Traceparent`. |
| `LINK_REWRITE` | Umschreiben absoluter Links; der Dienst nimmt nur `off` an, und `off` gilt ohne Angabe. |
| `STRICT_COVERAGE` | Wahrheitswert; wirkt im Modus `mock`: `true` weist eine Collection-Abfrage über einen nicht vollständigen Bestand ab, `false` beantwortet sie aus den bekannten Entities. Ohne Angabe gilt `true`. |
| `ETAG_ENABLED` | Wahrheitswert; wirkt im Modus `mock` für Services der OData-Version 4.0: `true` liefert zu jeder Entity einen ETag der Form `W/"<n>"`, dessen Zahl allein der Zähler der Entity bestimmt. Er steht in der Kopfzeile `ETag` bei der Abfrage einer Entity (`GET`, `HEAD`) und bei `POST`, `PATCH` und `PUT` sowie, wenn die Antwort `odata.metadata=minimal` ist, als Mitglied `@odata.etag` der Entity im Body, auch bei jedem Element einer Collection. Jede erfolgreiche Änderung einer Entity (`POST`, `PATCH`, `PUT`, auch mit leerem Inhalt) ergibt einen neuen ETag; ein fehlgeschlagener Vorgang, das Lesen und die Änderung einer anderen Entity ändern ihn nicht. `false` liefert keinen ETag. Die Header `If-Match` und `If-None-Match` ändern bei `true` wie bei `false` nichts an der Antwort und am Bestand. Ohne Angabe gilt `false`. |
| `PUBLIC_BASE_URL` | Öffentliche Basis der Links, die der Dienst im Modus `mock` erzeugt: `http://` oder `https://`, Host, optional Port und Segmente, ohne abschließenden Schrägstrich, höchstens 2048 Bytes. Ohne Angabe leitet der Dienst sie aus `LISTEN` ab, wenn dessen Host `127.0.0.1`, `[::1]` oder `localhost` ist; bei jedem anderen Host verlangt der Modus `mock` sie. |

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
| `ODR-E6001` | Der Mock unterstützt die Abfrage oder den lesenden Zugriff nicht. |
| `ODR-E6002` | Der Mock unterstützt die Änderung oder den schreibenden Zugriff nicht. |
| `ODR-E6003` | Die aufgezeichneten Daten decken das Entity-Set nicht vollständig ab, und der strikte Modus weist die Abfrage ab. |
| `ODR-E6004` | Die Metadaten des Services fehlen oder sind ungültig. |
| `ODR-E6005` | Die OData-URL der Anfrage ist ungültig. |
| `ODR-E6007` | Der Request-Body ist nicht lesbar oder verstößt gegen das Modell. |
| `ODR-E6008` | Die Entity oder das Entity-Set ist unbekannt. |
| `ODR-E6009` | Der Schlüssel der Entity ist bereits vergeben. |
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

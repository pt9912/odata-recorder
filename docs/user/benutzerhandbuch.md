# Benutzerhandbuch OData Recorder

Software-Version: dev
Gültigkeitsbereich: das Binary, das `make build` erzeugt
Stand: 2026-11-26

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
| `MODE` | Betriebsart des Dienstes: `record` nimmt Austausche mit dem Upstream auf, `replay` spielt aufgezeichnete Austausche ab. `mock` beantwortet Anfragen allein aus dem Bestand, ohne Upstream. `hybrid` beantwortet Anfragen zuerst aus dem Bestand wie `mock` und wendet sich nur dann an die aufgezeichneten Austausche oder an den Upstream, wenn der Mock mit einem freigegebenen Fehlercode scheitert; die vier Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` gehören zu ihm (Abschnitt `Modus hybrid`). Ohne Angabe gilt `mock`. |
| `LISTEN` | Host und Port, an dem der Dienst Anfragen annimmt, in der Form `host:port`. Ohne Angabe gilt `127.0.0.1:8080`. |
| `LISTEN_TLS_CERT` | Pfad einer PEM-Datei mit dem Zertifikat des Dienstes; zuerst steht das Zertifikat des Dienstes, danach stehen die Zwischenzertifikate. Mit `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` nimmt der Dienst Anfragen über HTTPS an; beide oder keiner (Abschnitt `TLS`). Ohne Angabe nimmt der Dienst Anfragen über HTTP an. |
| `LISTEN_TLS_KEY` | Pfad einer PEM-Datei mit dem privaten Schlüssel zum Zertifikat. Beide Pfade dürfen auf dieselbe Datei zeigen. Ohne Angabe nimmt der Dienst Anfragen über HTTP an. |
| `DATA_DIR` | Verzeichnis, in dem der Dienst seine Daten ablegt, absolut oder relativ zum Arbeitsverzeichnis. Ohne Angabe gilt `data`. |
| `SERVICE_ID` | Kennung des einzigen Services, 1 bis 64 Zeichen aus Buchstaben, Ziffern, `.`, `_` und `-`. Ohne Angabe gilt `default`. |
| `SERVICE_ROOT_PATH` | Pfad, unter dem der einzige Service Anfragen annimmt: `/` oder Segmente wie `/a/b` ohne abschließenden Schrägstrich. Ohne Angabe gilt `/`. |
| `UPSTREAM_URL` | URL des Upstreams, absolut mit `http` oder `https`, ohne Zugangsdaten, Query und Fragment. Der Modus `record` verlangt sie, ebenso der Modus `hybrid` bei `HYBRID_FALLBACK=enabled`; im Modus `hybrid` ist sie das Ziel des Upstream-Fallbacks. |
| `SERVICE_ODATA_VERSION` | OData-Version des einzigen Services: `4.0` oder `2.0`, genau in dieser Schreibweise. Ohne Angabe gilt `4.0`. Der Modus `hybrid` startet nicht mit einem Service der Version `2.0`. |
| `SERVICES_FILE` | Pfad einer JSON-Datei, die mehrere Services beschreibt, ein Objekt je Service mit den Feldern `id`, `service_root`, `upstream_url` und `odata_version`; `odata_version` ist wie `SERVICE_ODATA_VERSION` `4.0` oder `2.0`, ohne Angabe gilt `4.0`. Zugleich gesetzte `SERVICE_ID`, `SERVICE_ROOT_PATH`, `UPSTREAM_URL` oder `SERVICE_ODATA_VERSION` verhindern den Start. |
| `REPLAY_STRATEGY` | Strategie, mit der das Abspielen wiederholte gleiche Anfragen beantwortet: `sequential`, `strict`, `first` oder `last`. Ohne Angabe gilt `sequential`. |
| `REPLAY_ASSIGNMENT` | Wie eine Anfrage einem Workspace und einem Replay zugeordnet wird: `explicit` oder `ordered`. Ohne Angabe gilt `explicit`. |
| `REPLAY_REDACTED_BODIES` | Wahrheitswert `true` oder `false`: ob das Abspielen auch Austausche liefert, deren Antwort-Body geschwärzt ist. Ohne Angabe gilt `false`. |
| `MATCH_VARY_HEADERS` | Header-Namen, die in den Match-Key eingehen, durch Komma getrennt; `none` steht für keinen Header. Ohne Angabe gilt `Accept, OData-Version, Content-Type, If-Match, Prefer`. |
| `MATCH_INVALID_ENCODING` | Umgang mit einer ungültigen Kodierung der Anfrage beim Zuordnen: `literal` oder `reject`. Ohne Angabe gilt `literal`. |
| `MATCH_BODY_MODE` | Wie der Body einer Anfrage in den Match-Key eingeht: `bytes` oder `json-canonical`. Ohne Angabe gilt `bytes`. |
| `REDACT_HEADERS` | Zusätzliche Header-Namen, deren Werte der Dienst bei der Aufnahme schwärzt, durch Komma getrennt. Die Liste ergänzt die eingebaute Liste. |
| `REDACT_QUERY_PARAMS` | Zusätzliche Namen von Query-Parametern, deren Werte der Dienst schwärzt, durch Komma getrennt. Die Liste ergänzt die eingebaute Liste. Ein Name besteht aus den druckbaren ASCII-Zeichen ohne Leerzeichen und ohne `=`, `&`, `#`, `,` und `%`, also kein `%`; ein Name mit einem anderen Zeichen beendet den Start mit einem Fehler. |
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
| `STRICT_COVERAGE` | Wahrheitswert; wirkt in den Modi `mock` und `hybrid`: `true` weist eine Collection-Abfrage über einen nicht vollständigen Bestand ab, `false` beantwortet sie aus den bekannten Entities; im Modus `hybrid` gibt es mit `false` für diese Abfrage keinen Fallback, weil der Mock sie ohne Fehler beantwortet. Ohne Angabe gilt `true`. |
| `ETAG_ENABLED` | Wahrheitswert; wirkt in den Modi `mock` und `hybrid` für Services der OData-Version 4.0: `true` liefert zu jeder Entity einen ETag der Form `W/"<n>"`, dessen Zahl allein der Zähler der Entity bestimmt. Er steht in der Kopfzeile `ETag` und, in der Stufe `odata.metadata=minimal`, als Mitglied `@odata.etag` der Entity. `true` wertet außerdem bei `PATCH`, `PUT` und `DELETE` die Kopfzeilen `If-Match` und `If-None-Match` gegen den ETag der Entity aus (der Vergleich ist schwach, `W/` zählt nicht) und weist eine nicht erfüllte Bedingung mit `412` ab, ohne den Zustand zu ändern (der Fehlercode steht in der Tabelle der Fehlercodes). `false` liefert keinen ETag und liest beide Kopfzeilen nie. Ohne Angabe gilt `false`. |
| `PUBLIC_BASE_URL` | Öffentliche Basis der Links, die der Dienst in den Modi `mock` und `hybrid` erzeugt: `http://` oder `https://`, Host, optional Port und Segmente, ohne abschließenden Schrägstrich, höchstens 2048 Bytes. Ohne Angabe leitet der Dienst sie aus `LISTEN` ab, wenn dessen Host `127.0.0.1`, `[::1]` oder `localhost` ist, mit dem Schema `https`, wenn `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` gesetzt sind, sonst mit `http`; bei jedem anderen Host verlangen beide Modi sie. Eine gesetzte Basis gleicht der Dienst nicht mit TLS ab: Setze sie auf die Adresse, die der Client sieht. |
| `HYBRID_FALLBACK` | Hauptschalter des Upstream-Fallbacks im Modus `hybrid`: `disabled` oder `enabled`, in dieser Schreibweise. Mit `enabled` verlangt die Prüfung mindestens einen Eintrag in `FALLBACK_ALLOW_HOSTS` und je Service eine `UPSTREAM_URL`. Ohne Angabe gilt `disabled`: dann geht keine Anfrage an den Upstream, auch wenn die übrigen Schlüssel gesetzt sind. |
| `HYBRID_REPLAY_FALLBACK` | Kategorien, für die im Modus `hybrid` der Replay-Fallback gilt, durch Komma getrennt: `unsupported` und `coverage_insufficient`, in dieser Schreibweise. Ein anderer Eintrag, auch ein leerer, verhindert den Start. Ohne Angabe ist die Liste leer. Eine Anfrage, die der Mock nicht beantwortet, erhält für die Kategorie dann die Antwort aus den aufgezeichneten Austauschen. |
| `FALLBACK_ALLOW_HOSTS` | Hosts, zu denen im Modus `hybrid` der Upstream-Fallback gehen darf, durch Komma getrennt: ein Name oder eine IPv6-Adresse in eckigen Klammern, optional mit `:` und Port, ohne Schema und Pfad. Ein ungültiger Eintrag, auch ein leerer, verhindert den Start. Ohne Angabe ist die Liste leer. Ein Eintrag gilt nur für den Host der `UPSTREAM_URL` des Services der Anfrage, mit dem Port, den die URL nennt oder ihr Schema voraussetzt. |
| `FALLBACK_WRITE_ALLOW` | Routen, für die im Modus `hybrid` ein schreibender Upstream-Fallback gilt, durch Komma getrennt: eine Methode (`POST`, `PUT`, `PATCH`, `DELETE` oder `MERGE`), ein Leerzeichen und ein Pfadmuster, das mit `/` beginnt, je Segment höchstens ein `*` trägt und am Ende `**` tragen darf. Ein ungültiger Eintrag, auch ein leerer, verhindert den Start. Ohne Angabe ist die Liste leer. Ein Eintrag gilt nur für eine Anfrage, deren Methode und Pfad er deckt, und nur zu einem Host, den `FALLBACK_ALLOW_HOSTS` freigibt. |
| `ADMIN_LISTEN` | Host und Port, an dem der Dienst die Admin-Schnittstelle annimmt, in der Form `host:port` wie bei `LISTEN`. Loopback sind allein die Hosts `127.0.0.1`, `[::1]` und `localhost`, genau in dieser Schreibweise; jeder andere Host verlangt `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` oder `ADMIN_ALLOW_PLAINTEXT=true` (Kapitel `TLS`). Ohne Angabe gilt `127.0.0.1:8081`. |
| `ADMIN_TOKEN_FILE` | Pfad der Datei mit dem Token der Admin-Schnittstelle, absolut oder relativ zum Arbeitsverzeichnis (Abschnitt `Admin-Schnittstelle`). Ohne Angabe gilt die Datei `admin.token` im Verzeichnis von `DATA_DIR`. |
| `ADMIN_TLS_CERT` | Pfad einer PEM-Datei mit dem Zertifikat der Admin-Schnittstelle; zuerst steht ihr Zertifikat, danach stehen die Zwischenzertifikate. Mit `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` nimmt die Admin-Schnittstelle Anfragen über HTTPS an; setze beide oder keinen (Kapitel `TLS`). Ohne Angabe nimmt sie Anfragen über HTTP an. |
| `ADMIN_TLS_KEY` | Pfad einer PEM-Datei mit dem privaten Schlüssel zum Zertifikat von `ADMIN_TLS_CERT`. Beide Pfade dürfen auf dieselbe Datei zeigen. |
| `ADMIN_ALLOW_PLAINTEXT` | Wahrheitswert `true` oder `false`: ob die Admin-Schnittstelle auf einem Host außerhalb von Loopback ohne TLS laufen darf. Ohne Angabe gilt `false`. Bei Loopback und bei gesetztem Zertifikatspaar bleibt der Wert ohne Wirkung (Kapitel `TLS`). |

## Modus hybrid

Aufgabe: Anfragen aus dem Bestand beantworten und nur dort auf aufgezeichnete Austausche oder den Upstream zurückgreifen, wo der Mock sie nicht beantwortet.

Voraussetzung: `MODE=hybrid`. Der Start verlangt weder eine Session noch einen Workspace. Jede Anfrage nennt den Workspace im Header `X-Recorder-Workspace`; fehlt er, antwortet der Dienst mit `ODR-E5003`, ist er unbekannt, mit `ODR-E7001`.

Ablauf einer Anfrage:

1. Der Mock antwortet zuerst. Ist seine Antwort kein Fehler, geht sie unverändert an den Client, ohne Kennzeichnung.
2. Scheitert der Mock mit `ODR-E6001` (nicht unterstützte Abfrage), `ODR-E6002` (nicht unterstützte Änderung) oder `ODR-E6003` (Bestand nicht vollständig), entscheiden `HYBRID_REPLAY_FALLBACK`, `HYBRID_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW`, ob ein Fallback stattfindet. Entschieden wird am Fehlercode, nie am HTTP-Status.
3. Jeder andere Fehler kehrt unverändert zurück und löst nie einen Fallback aus: die fachlichen Fehler `ODR-E6008` (404), `ODR-E6009` (409) und `ODR-E6010` (412), die ungültigen Anfragen `ODR-E6005` und `ODR-E6007` sowie jeder Fehler des Replays außer `ODR-E5001`, darunter `ODR-E5002`.

Replay-Fallback: Steht die Kategorie des Fehlercodes in `HYBRID_REPLAY_FALLBACK` (`unsupported` für `ODR-E6001` und `ODR-E6002`, `coverage_insufficient` für `ODR-E6003`), antwortet der Dienst aus den aufgezeichneten Austauschen, für lesende und für ändernde Methoden, und sendet nichts an den Upstream. Der Client nennt dafür die Header `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session`; fehlt einer, ist die Antwort `ODR-E5003`. Diese Prüfung geschieht erst im Replay-Fallback: eine Anfrage, die der Mock beantwortet, braucht die beiden Header nie.

Upstream-Fallback: Nur mit `HYBRID_FALLBACK=enabled` und nur zum Host der `UPSTREAM_URL` des Services, den `FALLBACK_ALLOW_HOSTS` nennt.

- Lesende Anfragen (`GET`, `HEAD`) gehen bei `ODR-E6001` und `ODR-E6003` an den Upstream, ebenso nach einem Replay-Fallback ohne Treffer (`ODR-E5001`).
- Ändernde Anfragen (`POST`, `PUT`, `PATCH`, `DELETE`, `MERGE`) gehen bei `ODR-E6001`, `ODR-E6002` und `ODR-E6003` nur an den Upstream, wenn ein Eintrag von `FALLBACK_WRITE_ALLOW` ihre Methode und ihren Pfad deckt. Nach einem Replay-Fallback ohne Treffer erreicht eine ändernde Anfrage den Upstream nie.
- Jede andere Methode (etwa `OPTIONS`) erreicht weder den Replay noch den Upstream.
- Sind für eine Kategorie beide Wege offen, gilt der Replay-Fallback. Wer für eine Kategorie den Replay-Fallback setzt, schaltet damit den schreibenden Upstream-Fallback für deren Codes ab, weil die Zeile des Replay-Misses nur `GET` und `HEAD` kennt.
- Ein `UPSTREAM_URL` mit einem Byte ab `0x80` im Host oder mit der Zone einer IPv6-Adresse gibt nie einen Fallback frei. Der Betreiber setzt die ASCII-Form, den Punycode des Hosts.
- Die Antwort des Upstreams geht mit ihrem Status und ihren Headern an den Client, auch mit Status 3xx, 4xx oder 5xx. Ist der Upstream nicht erreichbar, ist die Antwort `ODR-E4001` mit 502, läuft `UPSTREAM_TIMEOUT` ab, mit 504.

Kennzeichnung: Jede Antwort aus einem Fallback trägt den Header `X-Recorder-Fallback` mit dem Wert `replay` oder `upstream`. Eine Antwort des Mock trägt ihn nie. Absolute Links in einer Fallback-Antwort bleiben, wie der Upstream oder die Aufzeichnung sie lieferte, denn `LINK_REWRITE` steht auf `off`.

Body: Eine Anfrage mit Body, für die ein Fallback möglich ist, liest der Dienst bis `MAX_CAPTURE_BODY` vollständig, bevor der Mock antwortet. Ein Body über dieser Grenze erhält nie einen Fallback; der Client sieht die Antwort des Mock.

Herunterfahren: Eine Anfrage, die auf den Upstream wartet, ist eine laufende Anfrage. Der Dienst wartet höchstens `SHUTDOWN_TIMEOUT` auf sie; danach endet sie ohne Antwort, und eine vom Upstream angenommene Änderung kann ausgeführt sein, ohne dass der Client die Antwort sieht. Setze `SHUTDOWN_TIMEOUT` über die Antwortzeit des Upstreams.

Beim Start im Modus `hybrid` mit `HYBRID_FALLBACK=enabled` und `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` schreibt der Dienst genau eine Warnung, dass die Prüfung des Zertifikats abgeschaltet ist.

## Admin-Schnittstelle

Aufgabe: Sessions, Imports, Workspaces und Snapshots verwalten und den Zustand des Dienstes lesen.

Voraussetzung: Zugriff auf die Token-Datei und auf die Adresse von `ADMIN_LISTEN`; auf Loopback heißt das Zugriff auf denselben Rechner (oder Container) wie der Dienst. Die Schnittstelle läuft in jedem Modus neben dem Dienst, an der Adresse von `ADMIN_LISTEN`: ohne `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` über HTTP, mit beiden über HTTPS. Auf einem Host außerhalb von Loopback, auch `0.0.0.0`, `[::]` oder `127.0.0.2`, verlangt sie dieses Zertifikatspaar oder `ADMIN_ALLOW_PLAINTEXT=true` (Kapitel `TLS`); sonst endet der Start mit `ODR-E2001` am Schlüssel `ADMIN_LISTEN`. `LISTEN` und `ADMIN_LISTEN` brauchen verschiedene Adressen. Derselbe Port an derselben Adresse beendet den Start mit `ODR-E2001 ADMIN_LISTEN: Adresse nicht bindbar`, ebenso derselbe Port, wenn `LISTEN` den Host `0.0.0.0` oder `[::]` trägt. `localhost` zählt dabei als die Adresse, zu der der Rechner den Namen auflöst. Derselbe Port an zwei verschiedenen einzelnen Adressen (etwa `127.0.0.2` für `LISTEN` und `127.0.0.1` für `ADMIN_LISTEN`) startet beide. Ein anderer Port für `ADMIN_LISTEN` vermeidet den Konflikt. Jede Variable, deren Name mit `ADMIN_` beginnt, die keiner der Schlüssel der Tabelle `Einstellungen` ist und die einen Wert trägt, beendet den Start mit `ODR-E2001`; eine leere oder nur aus Leerraum bestehende Variable bleibt ohne Wirkung.

Token: Der Dienst liest beim Start die Datei von `ADMIN_TOKEN_FILE`, bevor er eine Adresse öffnet. Fehlt sie, legt er sie mit einem zufälligen Token von 64 Zeichen und den Rechten `0600` an; ein Verzeichnis legt er dafür nicht an. Eine eigene Datei trägt eine Zeile mit mindestens 32 Zeichen, ohne Rechte für Gruppe und andere (`0400` oder `0600`). Wählst du das Token selbst, nimm einen zufälligen Wert: der Dienst begrenzt die Zahl der Versuche nicht. Ein Neustart auf derselben Datei nutzt dasselbe Token. Ein Verstoß beendet den Start mit `ODR-E2001 ADMIN_TOKEN_FILE:` und einem der folgenden Gründe, nie mit dem Pfad oder dem Inhalt:

| Grund | Bedeutung |
|---|---|
| `Datei nicht lesbar` | Der Pfad führt über eine Datei, hat eine Schleife von Verweisen oder ist nicht lesbar. |
| `ist ein Verzeichnis` | An der Stelle liegt ein Verzeichnis. |
| `keine reguläre Datei` | An der Stelle liegt etwas anderes als eine Datei oder ein Verweis ohne Ziel. |
| `Gruppen- oder Weltrechte` | Die Datei hat Rechte für Gruppe oder andere. |
| `mehr als eine Zeile` | Die Datei hat mehr als eine Zeile. |
| `weniger als 32 Zeichen` | Das Token ist kürzer als 32 Zeichen. |
| `Datei nicht anlegbar` | Die Datei fehlt, und der Dienst kann sie nicht anlegen. |
| `Datei, die der Dienst selbst schreibt` | Der Pfad nennt in `DATA_DIR` die Datei `audit.jsonl` oder eine Datei der Datenbank `odata-recorder.db`. |

Vorgehen:

1. Lies das Token aus der Datei.
2. Sende jede Anfrage mit dem Header `Authorization: Bearer <Token>`. Allein `GET /healthz` verlangt kein Token.

| Methode und Pfad | Aufgabe |
|---|---|
| `GET /healthz` | Bereitschaft des Dienstes, ohne Token: `{"status":"ok"}`. |
| `GET /__admin/sessions` | Sessions auflisten. |
| `POST /__admin/sessions` | Aufnahme-Session anlegen, Body `{"service_id":"…","id":"…"}`, `id` optional. |
| `POST /__admin/sessions/{id}/close` | Session beenden; der Dienst wartet auf laufende Austausche. |
| `GET /__admin/sessions/{id}/exchanges` | Austausche einer Session seitenweise lesen (`limit`, `after`). |
| `DELETE /__admin/sessions/{id}` | Session mit ihren Austauschen löschen. |
| `GET /__admin/coverage` | Abdeckung einer Baseline lesen (`baseline_id`). |
| `POST /__admin/imports` | Session in eine Baseline importieren; nur im Modus `record`. |
| `GET /__admin/imports/{id}` | Bericht eines Imports lesen; nur im Modus `record`. |
| `GET /__admin/workspaces` | Workspaces auflisten. |
| `POST /__admin/workspaces` | Workspace aus einer Baseline anlegen, Body `{"baseline_snapshot_id":"…","id":"…"}`, `id` optional. |
| `GET /__admin/workspaces/{id}` | Workspace lesen. |
| `DELETE /__admin/workspaces/{id}` | Workspace entfernen. |
| `POST /__admin/workspaces/{id}/reset` | Workspace auf die Baseline zurücksetzen. |
| `GET /__admin/workspaces/{id}/snapshots` | Snapshots auflisten. |
| `POST /__admin/workspaces/{id}/snapshots` | Snapshot anlegen, Body `{"name":"…"}`. |
| `POST /__admin/workspaces/{id}/restore` | Snapshot wiederherstellen, Body `{"snapshot_id":"…"}`. |
| `GET /__admin/diagnostics` | Modus, Einstellungen ohne Geheimnisse und Zähler je Fehlercode lesen. |

Ergebnis: Jede Antwort trägt den Header `X-Correlation-Id`, ein Fehler dazu ein Objekt mit `code`, `message` und `target`. Eine Anfrage ohne gültiges Token ist `401` mit `ODR-E2101`; der Dienst schreibt dazu eine Zeile in die Datei `audit.jsonl` in `DATA_DIR`. Ein unbekannter Pfad, eine falsche Methode oder ein ungültiger Wert ist `ODR-E2102`. In den Modi `replay`, `mock` und `hybrid` gibt es die beiden Pfade des Imports nicht, und eine dort angelegte Session nimmt keinen Austausch an; sie bleibt `open`, bis sie jemand beendet, und der nächste Start setzt sie `incomplete`. Beim Beenden nimmt die Schnittstelle keine Verbindung mehr an, und eine Anfrage auf einer offenen Verbindung ist `503` mit `ODR-E5010`.

## TLS

Aufgabe: den Dienst Anfragen über HTTPS annehmen lassen.

Voraussetzung: eine PEM-Datei mit dem Zertifikat des Dienstes (zuerst das Zertifikat des Dienstes, danach die Zwischenzertifikate) und eine PEM-Datei mit dem privaten Schlüssel, den der Dienst lesen darf. Beide Inhalte dürfen in derselben Datei stehen; dann zeigen beide Schlüssel auf diese Datei. Ein verschlüsselter Schlüssel ist ungültig, der Dienst kennt keine Passphrase.

Vorgehen:

1. Setze `LISTEN_TLS_CERT` auf den Pfad der Zertifikatsdatei und `LISTEN_TLS_KEY` auf den Pfad der Schlüsseldatei. Setze beide oder keinen; ein Pfad ist absolut oder relativ zum Arbeitsverzeichnis.
2. Starte den Dienst mit `start`.
3. Setze in den Modi `mock` und `hybrid` die Schlüssel `LISTEN` und `PUBLIC_BASE_URL` so, dass die Links die Adresse tragen, die der Client erreicht.

Ergebnis: Der Dienst nimmt unter `LISTEN` Anfragen über HTTPS an, in jedem Modus. Ohne beide Schlüssel nimmt er Anfragen über HTTP an. Er bedient TLS 1.2 und TLS 1.3 und spricht HTTP/1.1; ein Client, der allein HTTP/2 anbietet, kommt über den Handshake nicht hinaus. Die Dateien liest der Dienst einmal beim Start: ein neues Zertifikat oder ein neuer Schlüssel gilt ab dem nächsten Start. Läuft das Zertifikat im Betrieb ab, bedient der Dienst es weiter und meldet nichts; der Client lehnt es ab.

Prüfung beim Start: Der Dienst prüft, dass beide Dateien lesbar sind, dass die Zertifikatsdatei ein PEM-Zertifikat und die Schlüsseldatei einen PEM-Schlüssel enthält, dass der Schlüssel zum Zertifikat passt und dass das Zertifikat des Dienstes zur Zeit des Starts gilt. Bei einem Verstoß endet der Start mit der Meldung `ODR-E2001`, bevor der Dienst das Datenverzeichnis anlegt und bevor er eine Adresse öffnet. Der Dienst prüft nicht, ob die Zwischenzertifikate gelten, ob die Kette zu einer Wurzel führt, ob der Name im Zertifikat zum Host passt und ob das Zertifikat widerrufen ist; einen solchen Fehler zeigt der Client.

Die Meldung nennt den Schlüssel und einen der folgenden Gründe, nie den Pfad, den Inhalt der Dateien oder eine Eigenschaft des Zertifikats:

| Grund | Bedeutung |
|---|---|
| `erforderlich, wenn LISTEN_TLS_CERT gesetzt ist` | `LISTEN_TLS_KEY` fehlt, `LISTEN_TLS_CERT` ist gesetzt. |
| `erforderlich, wenn LISTEN_TLS_KEY gesetzt ist` | `LISTEN_TLS_CERT` fehlt, `LISTEN_TLS_KEY` ist gesetzt. |
| `Datei nicht lesbar` | Die Datei fehlt, ist ein Verzeichnis oder nicht lesbar. |
| `Datei enthält kein PEM-Zertifikat` | Die Datei enthält kein PEM-Zertifikat, das der Dienst lesen kann. |
| `Datei enthält keinen PEM-Schlüssel` | Die Datei enthält keinen PEM-Block `PRIVATE KEY` und keinen Block, dessen Typ auf ` PRIVATE KEY` endet. |
| `Zertifikat außerhalb seiner Gültigkeit` | Das Zertifikat ist abgelaufen, oder seine Gültigkeit beginnt nach dem Start. |
| `Schlüssel ungültig oder passt nicht zum Zertifikat` | Der Schlüssel ist ungültig oder gehört nicht zum ersten Zertifikat der Datei. |

Ein Client, dessen Handshake scheitert, etwa weil er das Zertifikat ablehnt, erreicht den Dienst nicht. Der Dienst schreibt dazu keine Zeile; den Fehler sieht der Betreiber am Client.

Rechte: Lege die Schlüsseldatei mit den Rechten `0400` oder `0600` an. Der Dienst prüft die Rechte nicht und startet auch mit einer Datei, die jeder lesen darf.

### TLS der Admin-Schnittstelle

Aufgabe: die Admin-Schnittstelle über HTTPS annehmen lassen oder, auf einem Host außerhalb von Loopback, bewusst ohne TLS betreiben.

Voraussetzung: wie oben eine PEM-Datei mit dem Zertifikat und eine PEM-Datei mit dem privaten Schlüssel, hier für die Admin-Schnittstelle. `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` dürfen auf dieselben Dateien zeigen wie `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY`; die Schlüssel des Dienstes und der Admin-Schnittstelle wirken unabhängig voneinander.

Vorgehen:

1. Setze `ADMIN_TLS_CERT` auf den Pfad der Zertifikatsdatei und `ADMIN_TLS_KEY` auf den Pfad der Schlüsseldatei, beide oder keinen.
2. Setze `ADMIN_LISTEN`, wenn die Schnittstelle auf einem Host außerhalb von Loopback erreichbar sein soll.
3. Nur wenn die Schnittstelle dort ohne TLS laufen soll: setze `ADMIN_ALLOW_PLAINTEXT=true` statt des Zertifikatspaars.
4. Starte den Dienst mit `start`.

Ergebnis: Mit dem Zertifikatspaar nimmt die Admin-Schnittstelle Anfragen allein über HTTPS an, in jedem Modus und auf jedem Host, mit TLS 1.2 und TLS 1.3 und HTTP/1.1; ein Fehler im Paar lässt den Start scheitern, nie auf HTTP zurückfallen. Ohne das Paar nimmt sie auf Loopback Anfragen über HTTP an. Außerhalb von Loopback verlangt sie ohne das Paar `ADMIN_ALLOW_PLAINTEXT=true`. Die Dateien liest der Dienst beim Start: ein neues Zertifikat oder ein neuer Schlüssel gilt ab dem nächsten Start. Läuft das Zertifikat im Betrieb ab, bedient die Schnittstelle es weiter und meldet nichts; der Client lehnt es ab.

Prüfung beim Start: Der Dienst prüft das Paar wie das des Dienstes. Bei einem Verstoß endet der Start mit der Meldung `ODR-E2001`, bevor der Dienst das Datenverzeichnis anlegt, bevor er die Token-Datei liest oder anlegt und bevor er eine Adresse öffnet; die Verstöße beider Zertifikatspaare stehen in derselben Meldung, nach Schlüssel sortiert. Ein anderer Wert von `ADMIN_ALLOW_PLAINTEXT` als `true` oder `false` ist der Verstoß `ODR-E2001 ADMIN_ALLOW_PLAINTEXT: Wahrheitswert muss true oder false sein`; die Prüfung des Hosts entfällt dann, und die Dateien des Paars prüft der Dienst trotzdem. Die Meldung nennt den Schlüssel und einen der folgenden Gründe, nie den Pfad, den Inhalt der Dateien, eine Eigenschaft des Zertifikats oder den Wert von `ADMIN_LISTEN`:

| Schlüssel und Grund | Bedeutung |
|---|---|
| `ADMIN_LISTEN: außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY` | Der Host liegt außerhalb von Loopback, beide Pfade fehlen, und `ADMIN_ALLOW_PLAINTEXT` ist nicht `true`. |
| `ADMIN_TLS_KEY: erforderlich, wenn ADMIN_TLS_CERT gesetzt ist` | `ADMIN_TLS_KEY` fehlt, `ADMIN_TLS_CERT` ist gesetzt. |
| `ADMIN_TLS_CERT: erforderlich, wenn ADMIN_TLS_KEY gesetzt ist` | `ADMIN_TLS_CERT` fehlt, `ADMIN_TLS_KEY` ist gesetzt. |
| `Datei nicht lesbar` | Die Datei fehlt, ist ein Verzeichnis oder nicht lesbar. |
| `Datei enthält kein PEM-Zertifikat` | Die Datei enthält kein PEM-Zertifikat, das der Dienst lesen kann. |
| `Datei enthält keinen PEM-Schlüssel` | Die Datei enthält keinen PEM-Block `PRIVATE KEY` und keinen Block, dessen Typ auf ` PRIVATE KEY` endet. |
| `Zertifikat außerhalb seiner Gültigkeit` | Das Zertifikat ist abgelaufen, oder seine Gültigkeit beginnt nach dem Start. |
| `Schlüssel ungültig oder passt nicht zum Zertifikat` | Der Schlüssel ist ungültig oder gehört nicht zum ersten Zertifikat der Datei. |

Warnung ohne TLS: Läuft die Schnittstelle außerhalb von Loopback ohne Zertifikatspaar, weil `ADMIN_ALLOW_PLAINTEXT=true` gesetzt ist, schreibt der Start in jedem Modus genau eine Zeile der Stufe `WARN` auf die Fehlerausgabe, mit dem Text `start: Admin-Listener ohne TLS erreichbar` und dem Attribut `ereignis=admin_klartext`. Bei Loopback und mit Zertifikatspaar schreibt er sie nicht.

Verantwortung für den Klartext: Ohne TLS kann jeder auf dem Weg das Token und die Antworten mitlesen, und mit dem Token die Schnittstelle bedienen. Wer `ADMIN_ALLOW_PLAINTEXT=true` setzt, trägt dafür die Verantwortung, etwa durch ein abgeschottetes Netz oder einen Proxy, der TLS beendet.

Ein Client, dessen Handshake scheitert, etwa weil er das Zertifikat ablehnt, erreicht die Schnittstelle nicht. Der Dienst schreibt dazu keine Zeile; den Fehler sieht der Betreiber am Client.

## Fehlercodes

Dieser Abschnitt führt die Fehlercodes, die das Binary dem Anwender zeigt. Ein Fehlercode beginnt mit `ODR-E` und trägt vier Ziffern.

| Code | Bedeutung |
|---|---|
| `ODR-E2001` | Die Startkonfiguration ist ungültig. |
| `ODR-E2101` | Die Anfrage an die Admin-Schnittstelle hat kein gültiges Token (HTTP 401). |
| `ODR-E2102` | Die Anfrage an die Admin-Schnittstelle ist ungültig (HTTP 400, 404 oder 405). |
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
| `ODR-E6006` | Der Service hat die Version `2.0`, und die Anfrage nutzt eine Eigenschaft von OData V2 außerhalb des Umfangs (HTTP 501). |
| `ODR-E6007` | Der Request-Body ist nicht lesbar oder verstößt gegen das Modell. |
| `ODR-E6008` | Die Entity oder das Entity-Set ist unbekannt. |
| `ODR-E6009` | Der Schlüssel der Entity ist bereits vergeben. |
| `ODR-E6010` | Die Bedingung von `If-Match` oder `If-None-Match` trifft den ETag der Entity nicht (HTTP 412); die Änderung unterbleibt. |
| `ODR-E6101` | Der Import einer Session in eine Baseline ist ungültig oder widersprüchlich (HTTP 422). |
| `ODR-E6102` | Eine Kennung des Imports ist mit anderem Inhalt erneut benutzt (HTTP 409). |
| `ODR-E7001` | Der Workspace ist unbekannt. |
| `ODR-E7002` | Der Snapshot ist unbekannt. |
| `ODR-E7003` | Die Aufnahme-Session ist unbekannt. |
| `ODR-E7004` | Eine Baseline verweist auf die Session; sie lässt sich nicht löschen (HTTP 409). |
| `ODR-E7005` | Der Import ist unbekannt. |
| `ODR-E7006` | Der Name des Snapshots ist im Workspace vergeben (HTTP 409). |
| `ODR-E7007` | Die Kennung des Workspace ist vergeben (HTTP 409). |
| `ODR-E7008` | Die Art des Snapshots passt nicht zur Aktion (HTTP 422). |
| `ODR-E7009` | Die Session ist noch offen. |
| `ODR-E7010` | Die Kennung der Session ist vergeben (HTTP 409). |
| `ODR-E7011` | Die Session ist nicht offen (HTTP 409). |
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

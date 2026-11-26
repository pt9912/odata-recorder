# Verifikationsbericht: slice-lasttest (Teil 3, Messlauf `make lasttest`) — 2026-11-26

**Rolle:** Verifier (Modul 11), Frage „Bauen wir es richtig?“ gegen Plan und DoD.

**Stand der Prüfung und der Messung:** Zweig slice-lasttest, Commit `e7bfebb4`; die Pfade des Binary (`src/`, `go.mod`, `go.sum`) sind gleich `main`. Bezug: `LH-QA-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md).

**Eingang:** Plan slice-lasttest (§1 bis §6), `SPEC-216`, `SPEC-102`, `SPEC-128`, `LH-QA-01`, Review-Report 2026-11-26-slice-lasttest-review.

**Gliederung:** Eine Verifikation in zwei Schritten. Schritt 1 ist alles außer dem Lauf mit Profil `abnahme`, Schritt 2 der eine Aufruf `abnahme` (`SPEC-216` Nr. 1).

## Schlussurteil

**Das Werkzeug verhält sich in den geprüften Punkten wie zugesagt. Die DoD ist nicht vollständig.** Der eine Aufruf `abnahme` ist gefahren (Exit 3): `mock` ist `bestanden` (p95 29,032 ms gegen 100 ms), `replay` ist `ungueltig` mit `grund=binary` und hat **keinen** p95-Wert; `abnahme=nein grund=ergebnis`. `LH-QA-01` ist damit nicht abgenommen. Vor der Closure braucht es eine Entscheidung des Planners (V-7, V-8).

Befunde zum Slice: 0 HIGH, 1 MEDIUM, 2 LOW, 4 INFO. Dazu 1 Produktbefund (HIGH), der nicht dem Werkzeug zugerechnet ist.

## DoD-Stand

| DoD-Punkt | Stand | Beleg |
|---|---|---|
| Lauf (Takt, Prozess, Ablauf, Ausgabe, Sendungen) | **teilweise**: Lauf gefahren, p95 für `mock`, keines für `replay` (V-7) | Code gelesen; Handlauf `rauch`; Mutationen; Aufruf `abnahme` |
| Werkzeug und Verdrahtung | **erfüllt** | `make -n gates` nennt `lasttest.sh` 0-mal; Dockerfile, Fragment, Zeile in der Tabelle der Werkzeuge gelesen; `docker inspect`; Fälle 217 bis 225 als Modultests grün |
| Bericht und Entscheidung (Zahlen, Carveout, Paketzeit) | **offen** (Closure) | Zahlen in Schritt 2; Paket unter dem Detektor 96,05 s (ein Lauf, `--cpus=4`, Last 3,6 bis 4,2), unter 120 s |
| `make gates` grün | **erfüllt** | eigener Lauf auf `e7bfebb4`, Exit 0 |
| Review mit Report | **erfüllt** | Report liegt vor; Fix-Runde gegen die Klassen geprüft |
| Doku-Update | **erfüllt** | `harness/README.md` Zeile und *Nicht behauptet* |
| Handbuch oder Satz in §7 | Binary unverändert belegt (kein Commit des Slice an `src/`); der Satz in §7 ist **offen** | `git log --no-merges --first-parent 95779768..HEAD -- src go.mod go.sum` leer |
| Closure-Notiz | **offen** | Plan §7 |

## Schritt 1

**Verdikt:** Die Messung durfte auf diesem Stand gefahren werden. Auf dem Messweg (Programm, Treiber, Stufe, Binary) gab es keinen Befund, der vorher eine Korrektur verlangt hätte.

### Sensoren (selbst gefahren)

Host mit 20 Kernen, 31 GiB; jeder schwere Lauf unter `flock /tmp/odata-recorder-heavy.lock`; Messender ist der Verifier.

| Sensor | Ergebnis |
|---|---|
| `make gates` auf `e7bfebb4` (21:19 bis 21:20, Last 3,5 bis 3,8) | Exit 0; `groessenmarke slice-lasttest.md: ok, Messung 4316`; die Stufen `test` und `race` kamen aus dem Schicht-Cache |
| `pruefung race /src` frisch (Container aus `odata-recorder:race`, ohne Netz, 4 GiB, 4 CPU) | `ok` |
| `pruefung test /src` frisch (`GOFLAGS` mit `-count=1`, Image `odata-recorder:test`) | `ok`; Paket `tests/lasttest` ohne Detektor 50,3 s |
| `make lasttest LASTTEST_PROFIL=rauch` (21:41 bis 21:42, Last 5,5 vor und 4,1 nach dem Lauf) | Exit 0, 39 s; alle Zeilenarten ohne `Hinweis` |
| Größenmarke, eigene Summe von `git diff --numstat 95779768..HEAD` über die Pfade des Blocks | 4316, gedeckt durch `ueberschreitung: 4400`; Pfade außerhalb des Blocks allein aus `main` (`src/`, `tests/integration/`) |
| Modultests `tools/gegenproben` zum Lasttest (5 Tests, darunter die 9 Fälle gegen den Treiber) | grün; `make gegenproben` nicht gefahren |

**Handlauf `rauch`:** `Aufbau exchanges=300 entities=300 daten_mib=2 dauer_s=8.515`; `replay` und `mock` je 260 Anfragen, 260 Spans, 0 Fehler, `ohne_kind=0`, `gueltig=ja`; p95 14,923 ms und 2,044 ms; `verspaetung_max_ms` 1,162 und 1,190; Proben 4,07 bis 5,56 bei Schwelle 10,00; `abnahme=nein grund=profil`. Keine Zahl nach `LH-QA-01.a`.

**Laufender Container:** Netz `none`, 2 CPU, 4294967296 Bytes Speicher und Speicher mit Swap, `AutoRemove`, 0 Mounts, nicht privilegiert, `/out/lasttest`, die fünf `LASTTEST_`-Werte; danach kein Container des Images.

**Vom Plan an den Verifier gegeben:** `net.DialTimeout` steht nicht im Paket (Negativ (5)); `sammler.nimm` hat im Programm eine Aufrufstelle, den Leser von stderr (Sendung (a) von Teil 2).

### Fix-Runde zum Review gegen die Klassen (`AGENTS.md` §3.17)

Wegwerfkopie aus `git archive HEAD`, im Container ohne Netz (4 GiB, 4 CPU), `CGO_ENABLED=0`, ohne Detektor; je Mutation ein Prozess `go test -count=1` über die 71 Tests des Pakets außerhalb der Dateien von Teil 1. Nullprobe grün. Je Mutation eine Stelle oder Form, die der Report nicht nannte.

| Klasse | rot | grün |
|---|---|---|
| F-1 (weitergereichte Werte) | Recording-Session gleich Baseline; Mock-Workspace gleich Session; Sammler fest `replay`; Umgebung fest `replay` | Dauer der Proben zweimal Messphase (obere Seite, akzeptiertes Negativ (7)) |
| F-2 (alle Läufe, nur ihre) | `Abnahme` allein aus dem ersten Urteil; Exit allein aus einem Feld | — |
| F-3 (Werte an der Abnahme) | Profil fest; Betriebsart fest | — |
| F-4 (Verstöße der Auswertung) | `zaehlung` allein bei zu wenig; allein ohne Messphase | — |
| F-5 (Abbruch und Frist) | nach dem Aufbau allein ohne Fehler; Frist als Abbruch | Prüfung nach dem Lauf allein bei `nr == 1` (V-2) |
| F-6 (Meldung zum Binary) | Exit auf 128 gekappt | zwei Formen des Fensters „Bereitschaft aufgegeben, Binary endet danach“ (V-3) |
| I-1, Nr. 8 f | Vergleich in Millisekunden | Probe über der Schwelle zählt allein bei Verspätung 0 (V-1) |
| weitere Zusagen (Nr. 4 c, 8 b, 9, 12 b, 13 e) | ohne Keep-Alive; globale Grenze ruhender Verbindungen; `SIGINT` statt `SIGTERM`; Port auf `0.0.0.0`; Abnahme ohne das erste Urteil; Schreibfehler ohne Exit 2 bei 3 | Gesamtfrist mindestens 15 min (gleichwertig: kein Profil liegt darunter) |
| Treiber (Fälle 217 bis 225) | 8 von 8: Exit 2 als 3; 124 als `docker run`; leerer Startwert als Vorgabe; `--init`; `--no-cache` am Bau; Abbruch nach dem Bau übergangen; Ersatzwert für leere Kerne; `SIGKILL` statt `TERM` | — |

F-7: kein Marker `d-check:ignore` mehr im Kopf von `SPEC-216` und in der Zeile `Dockerfile` des Plans; `make docs-check` im Lauf der Gates grün. Die Fix-Runde deckt die Klassen; es bleiben zwei Lücken in den Fängern (V-1, V-2), der Code ist an beiden Stellen richtig.

### Befunde aus Schritt 1

| ID | Kategorie | Befund und Folge | Quelle | Adresse |
|---|---|---|---|---|
| V-1 | LOW | Die Fänger zur Probe über der Schwelle (`TestProbenImAblauf`, `TestVerstoesseImAblauf`) laufen allein mit Verspätung 0; die Mutation „Probe zählt allein bei Verspätung 0“ bleibt grün. Im echten Lauf ist die Verspätung nie 0, unter dieser Mutation bliebe die Probe ohne Wirkung. Das Paar der zwei Ursachen von `last` hat keinen Fall. | `SPEC-216` Nr. 8 f, 11; `AGENTS.md` §3.15, §3.16 | Implementer |
| V-2 | LOW | Abbruch und Gesamtfrist im zweiten Lauf einer Betriebsart haben keinen Fall (Achse Lauf): die Prüfung allein nach `nr == 1` bleibt grün. Das Profil `abnahme` fährt zwei Läufe. | `SPEC-216` Nr. 4 c, 4 d; `AGENTS.md` §3.15 | Implementer |
| V-3 | INFO | Endet das Binary zwischen dem Aufgeben der Bereitschaft und der Auswertung, hält kein Test die Wahl der Meldung; kein Einfluss auf p95, Urteil oder Exit. | `SPEC-216` Nr. 9 c | Planner (Negativ) |
| V-4 | INFO | Plan §6 nennt `/` (51 GiB frei) als Ort der Schreibschicht; die Wurzel von Docker ist `/docker` (96 % belegt, 21 bis 23 GiB frei). | Plan §6 | Planner |
| V-5 | INFO | Plan Zeile L10 nennt 9 Exits des Containers; Fall 224 und `SPEC-102` führen 12. | Plan §3 | Implementer |
| V-6 | INFO | `harness/gegenproben.stempel` gehört nicht zum Stand der Eingaben (der Treiber ist neu in der Liste); `make gegenproben` ist im Integrationsschritt geschuldet. | `SPEC-102` *Pflicht zum Fahren*, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) | Planner |

### Plan gegen Code

| Gegenstand | Ergebnis |
|---|---|
| Plan §3, Dateien | entsprechen dem Diff; der Block der Größenmarke deckt jeden eigenen Pfad |
| Tabelle *Zusage · Fänger*, L1 bis L12 | jede Zeile hat Tests oder Fälle; offen allein die Kombinationen V-1 und V-2 |
| Randformen (§6) | je Entscheidung eine Stelle in `SPEC-216`; ich habe keine Randform gefunden, die allein der Code entscheidet |
| Raten und die 155 Mutationen des Implementers | nicht nachgemessen |

## Schritt 2: Aufruf `abnahme`

Ein Aufruf mit den Vorgaben auf `e7bfebb4`, am 2026-11-26 von 22:54:34 bis 23:43:34 (49 Minuten), unter `flock /tmp/odata-recorder-heavy.lock`; Messender ist der Verifier. Kein zweiter Aufruf (`SPEC-216` Nr. 1).

`env -u LASTTEST_BETRIEBSART -u LASTTEST_PROFIL -u LASTTEST_ZIEL_MS -u LASTTEST_STARTWERT make lasttest`

### Ausgabe, wörtlich

stdout (ohne die Zeile, mit der `make` den Aufruf des Treibers nennt):

```
lasttest: Start profil=abnahme betriebsart=beide ziel_ms=100 startwert=1 kerne=20
lasttest: Aufbau exchanges=10000 entities=10000 daten_mib=69 dauer_s=275.532
lasttest: Messung betriebsart=replay lauf=1 anfragen=13200 spans=102 messphase=0 fehler=13099 fremde_zeilen=1 p95_ms=0.000 max_ms=0.000 kind=0 kind_p95_ms=0.000 kind_anteil=0.0 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=185 verspaetung_max_ms=3.134 bytes=333306 treiber_cpu_s=4.828 gueltig=nein grund=binary
lasttest: Last betriebsart=replay lauf=1 proben=24 min=1.01 max=3.40 schwelle=10.00
lasttest: Messung betriebsart=replay lauf=2 anfragen=13200 spans=92 messphase=0 fehler=13109 fremde_zeilen=1 p95_ms=0.000 max_ms=0.000 kind=0 kind_p95_ms=0.000 kind_anteil=0.0 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=183 verspaetung_max_ms=20.057 bytes=259267 treiber_cpu_s=4.889 gueltig=nein grund=binary
lasttest: Last betriebsart=replay lauf=2 proben=24 min=1.20 max=1.78 schwelle=10.00
lasttest: Messung betriebsart=mock lauf=1 anfragen=13200 spans=13200 messphase=12000 fehler=0 fremde_zeilen=1 p95_ms=29.032 max_ms=50.419 kind=12000 kind_p95_ms=28.892 kind_anteil=98.1 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=2 verspaetung_max_ms=1.299 bytes=97900912 treiber_cpu_s=6.283 gueltig=ja grund=ok
lasttest: Last betriebsart=mock lauf=1 proben=23 min=1.20 max=2.45 schwelle=10.00
lasttest: Messung betriebsart=mock lauf=2 anfragen=13200 spans=13200 messphase=12000 fehler=0 fremde_zeilen=1 p95_ms=28.841 max_ms=51.371 kind=12000 kind_p95_ms=28.663 kind_anteil=98.0 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=2 verspaetung_max_ms=1.259 bytes=97900912 treiber_cpu_s=6.317 gueltig=ja grund=ok
lasttest: Last betriebsart=mock lauf=2 proben=23 min=1.03 max=2.45 schwelle=10.00
lasttest: Urteil betriebsart=replay laeufe=2 ziel_ms=100 p95_ms=0.000 streuung_ms=0.000 toleranz_ms=0.000 ergebnis=ungueltig grund=binary
lasttest: Urteil betriebsart=mock laeufe=2 ziel_ms=100 p95_ms=29.032 streuung_ms=0.191 toleranz_ms=2.884 ergebnis=bestanden grund=ok
lasttest: Abnahme ziel_ms=100 abnahmewert_ms=100 abnahme=nein grund=ergebnis
```

stderr (Meldungen des Programms und von `make`, ohne die Ausgabe des Baus):

```
lasttest: Binary endete (Exit 137)
lasttest: Binary endete (Exit 137)
make: *** [harness/mk/lasttest.mk:8: lasttest] Fehler 3
```

Es steht keine Zeile `Hinweis`. `make` endete mit 2, der Treiber mit 3.

### Host-Bedingungen

| Gegenstand | Wert |
|---|---|
| Host | 20 Kerne, 31 GiB |
| Vor dem Start (22:54) | Last 1,89 · 2,24 · 3,96; 15 GiB verfügbar; Swap 10 von 15 GiB belegt; `/docker` 23 GiB frei (96 % belegt); Sperre frei; keine `LASTTEST_`-Variable; kein Container des Repos |
| Nach dem Ende (23:43) | Last 1,53 · 1,38 · 1,57; 15 GiB verfügbar; Swap 10 von 15 GiB belegt; `/docker` 23 GiB frei; kein Container des Images |
| Proben des Programms | kleinste 1,01, größte 3,40, Schwelle 10,00; 24 · 24 · 23 · 23 Proben je Lauf |
| Fremde Container, liefen beim Start und beim Ende, unberührt | `d-migrate-mcp`, `d-migrate-mysql`, `d-migrate-postgres`, `d-migrate-oracle`, `d-migrate-mssql`, `k3d-viper-registry` |
| Andere Arbeit | Ich habe während des Laufs nichts gestartet. Fremde Prozesse habe ich nicht mitgeschrieben; zur Last belegen allein die Proben. |

### Urteil nach `LH-QA-01.a`

| Betriebsart | Urteil der Zeile | Was die Zeilen sagen |
|---|---|---|
| `replay` | `ungueltig`, `grund=binary` | Kein p95: `messphase=0` in beiden Läufen. `p95_ms=0.000` ist kein Messwert (`SPEC-216` Nr. 10 f, 13 b) und darf in keiner Closure-Notiz als Zahl stehen. Das Binary endete in beiden Läufen mit Exit 137; 13099 und 13109 von 13200 Anfragen sind Fehler. |
| `mock` | `bestanden` | p95 29,032 ms (Läufe 29,032 und 28,841 ms), Maximum 51,371 ms, Anteil des Query-Spans 98,1 % und 98,0 %; Streuung 0,191 ms bei Toleranz 2,884 ms; 0 Fehler; Verspätung höchstens 1,299 ms. |
| Abnahme | `abnahme=nein grund=ergebnis` | `LH-QA-01` ist auf diesem Stand nicht abgenommen: für `replay` ist die Zusage nicht nachgewiesen, für `mock` liegt die Zahl unter dem vorläufigen Zielwert. |

### Zur Hochrechnung `replay` aus dem Zwischenbericht

Mein Zwischenbericht an den Auftraggeber erwartete für `replay` ein Urteil `ungueltig` durch Zeitüberschreitungen oder ein Ende des Binary wegen Speicher.

- **Die Zeilen zeigen:** `ungueltig` mit `grund=binary`, Exit 137 in beiden Läufen, nur 102 und 92 Gesamtspans, bis zu 185 und 183 gleichzeitig offene Anfragen.
- **Falsch war:** die Aussage, die Zeile `Messung` trage trotzdem ein p95. Sie trägt keines.
- **Die Zeilen zeigen nicht:** wer das Signal sandte, den Zeitpunkt des Endes, die Verarbeitungszeit eines Replay bei 10.000 Exchanges, ob `ListBySession` (`SPEC-135` Nr. 8) die Ursache ist, und bei welcher Rate oder Größe der Dienst hielte. Die hochgerechneten 0,5 s je Request sind weder bestätigt noch widerlegt.

### Zusätzlicher Beleg außerhalb der Zeilen: Kernel-Protokoll

`journalctl -k` zu diesem Zeitraum nennt zwei Abbrüche durch den OOM-Killer der Speicher-Gruppe des Containers `d0f46c08e751` (derselbe, den `docker events` mit Image `odata-recorder:lasttest` und Exit 3 führt):

| Zeitpunkt | Prozess | Speicher |
|---|---|---|
| 22:59:26 | `odata-recorder` | `anon-rss` 4147816 kB; Gruppe 4194304 kB von 4194304 kB; `lasttest` rund 28 MiB |
| 23:10:37 | `odata-recorder` | `anon-rss` 4147740 kB; Gruppe 4194304 kB von 4194304 kB; `lasttest` rund 29 MiB |

Exit 137 ist damit in beiden Läufen die Grenze von 4 GiB, erreicht durch den Speicher des Binary, nicht durch den des Lasttest-Programms. Warum das Binary den Speicher belegt, belegt auch dieses Protokoll nicht.

### Vertrag am echten Lauf

Am echten Binary in voller Größe beobachtet, jeweils wie zugesagt: Aufbau mit 10000 Exchanges und 10000 Entities in 275,5 s (Frist 30 min), `daten_mib=69`; ein Binary, das selbst endet, meldet sich auf stderr, der Takt läuft zu Ende (`anfragen=13200`), der Lauf trägt `binary` allein und die folgenden Läufe fahren (`SPEC-216` Nr. 9 c); 23 oder 24 Proben (Nr. 11); Streuung, Toleranz und das größere p95 (Nr. 12 c); `Urteil` ohne gültigen Lauf mit `0.000` (Nr. 13 b); Folge der Zeilen (Nr. 13 d); Exit 3 (Nr. 12 e); Dauer 49 min unter der Gesamtfrist von 120 min.

### Befunde aus Schritt 2

| ID | Kategorie | Befund und Folge | Quelle | Adresse |
|---|---|---|---|---|
| V-7 | MEDIUM | DoD-Punkt „Lauf“ und §1 des Plans verlangen p95 für Replay und Mock; für `replay` gibt es keinen. Der Plan hat einen Weg für das Überschreiten des Zielwerts (Carveout, §2 und §6), aber keinen für `ungueltig` durch das Ende des Binary. Vor der Closure ist zu entscheiden, ob der Punkt mit diesem Befund als erfüllt gilt und welcher Carveout mit Frist, Trigger und Folge-Slice ihn trägt; kein stilles Rot. | Plan §1, §2, §6; `SPEC-216` Nr. 12 d | Planner |
| V-8 | HIGH (Produkt) | Unter der Replay-Last von `LH-QA-01.a` (10.000 Exchanges, 20 Requests/s, 2 vCPU, 4 GiB) erreicht das Binary die Speichergrenze und wird in beiden Läufen beendet; der Dienst antwortet danach nicht mehr. Der Mangel liegt im Produkt, nicht im Werkzeug dieses Slice. | `LH-QA-01`; Zeilen und Kernel-Protokoll oben | Planner (Folge-Slice; naheliegende, nicht belegte Spur: `SPEC-135` Nr. 8) |

Für offene Einträge des Plans §6 liegen jetzt Zahlen vor (Adresse: Planner): Dauer und Größe des Aufbaus (275,5 s, 69 MiB, Aufruf 49 min); Stand des Binary im Handlauf (`e7bfebb4`, `slice-bootstrap-m6-endpunkte` enthalten, `slice-bootstrap-m6-listener` nicht); Paketzeit unter dem Detektor (96,05 s). Der Trigger zur Prüfung der Setzungen (`SPEC-216` Nr. 16 g, drei gültige Läufe je Betriebsart) ist nicht erreicht: `mock` hat zwei, `replay` keinen.

## Nicht geprüft

`make gegenproben`; Mutationen unter dem Detektor und Raten; Signale am echten Container (Handlauf des Reviewers auf `75fffee6`, der Treiber ist seither unverändert); die Antwortgröße der Collections des Mock (`SPEC-216` Nr. 16 l; die Zeilen nennen allein die Summe `bytes`); die Ursache des Speicherverbrauchs beim Replay.

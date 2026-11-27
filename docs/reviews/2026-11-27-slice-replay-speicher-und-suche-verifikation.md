# Verifikationsbericht: slice-replay-speicher-und-suche — 2026-11-27

**Rolle:** Verifier (Baseline-Regelwerk Modul 11): „Bauen wir es richtig?“ gegen Plan und DoD.

**Gegenstand:** Zweig slice-replay-speicher-und-suche, Stand `9c3f5f10` (enthält `main` bei `a4b79211`); eigener Anteil `git diff 0ee80d43..HEAD`.

**Eingang:** Plan slice-replay-speicher-und-suche, Review-Report `2026-11-27-slice-replay-speicher-und-suche-review`, Fix-Runde `4d455997`.

**Berührt:** `LH-QA-01`, `LH-QA-01.a`, `LH-FA-04`, `LH-FA-05`; `SPEC-105` Nr. 13, 14; `SPEC-135` Nr. 6, 8, 9, 18 und *Ungeprüft*; `SPEC-216` Nr. 1, 6 g, 12; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); [CO-001](../plan/carveouts/done/CO-001-lasttest-replay-ohne-p95.md).

**Gliederung:** Eine Verifikation in zwei Schritten. Schritt 1 ist alles außer dem Lauf mit Profil `abnahme`, Schritt 2 der eine Aufruf `abnahme` (`SPEC-216` Nr. 1).

---

## Verdikt

**Die Behebung verhält sich in den geprüften Punkten wie zugesagt, und der eine Aufruf `abnahme` auf `9c3f5f10` schreibt `abnahme=ja grund=ok`** (Exit 0): `replay` `bestanden` mit p95 10,613 ms, `mock` `bestanden` mit p95 32,648 ms, gegen den Zielwert 100. Kein Befund am Verhalten des Produkts. Die Liefer-Punkte der DoD sind erbracht; offen sind die Punkte der Closure (Closure-Notiz mit diesen Zahlen, [CO-001](../plan/carveouts/done/CO-001-lasttest-replay-ohne-p95.md), Register, Ausgänge in §6, Paarungen) und der Befund V-2 (LOW).

---

## Schritt 1

### Zwischenverdikt

**Die Messung darf auf `9c3f5f10` gefahren werden**, unter den Vorbedingungen unten und nach der Entscheidung zu V-1.

### Läufe

Host mit 20 Kernen, 31 GiB, fremde Datenbank-Container (unberührt). Messender ist der Verifier. Jeder Container-Lauf unter `flock /tmp/odata-recorder-heavy.lock`. Ein Reviewer arbeitete zugleich in einem anderen Worktree; die Last schwankte zwischen 3 und 58.

| Lauf | Ergebnis |
|---|---|
| `make gates` (11:10 bis 11:12, Last 12,8 und 13,5) | Exit 0. Die Stufen `build`, `test`, `race`, `lint`, `handbuch`, `katalog` kamen aus dem Cache von Docker (gleiche Eingaben, kein neuer Lauf); `test-integration`, `test-determinism`, `test-contract`, `test-security` liefen. `groessenmarke`: Messung 1075 von Marke 2700, Hinweis „Tabelle ergibt 2070, Schätzung 1990“ |
| `go test -count=1 ./...` im Container ohne Netz (4 CPU, 4 GiB) | alle Pakete grün |
| `go test -race -count=1 -trimpath ./...`, ebenso (Last 4,1 bis 7,6) | alle Pakete grün, kein Datenrennen |
| `make lasttest LASTTEST_PROFIL=rauch` (12:17, Last 9,8) | Exit 0, `gueltig=ja` beide; `replay` p95 14,798 ms, `mock` 2,293 ms, 0 Fehler; `abnahme=nein grund=profil`. Proben des Programms 8,84 bis 9,97 gegen die Schwelle 10,00 |
| `make gegenproben` | nicht gefahren (Auftrag) |

Abweichung vom Auftrag: `make gates` und der Bau- und Mutationslauf starteten bei Last 13, über der vorgegebenen 10 (sie standen ohne Lastprüfung hinter der Sperre). Alle späteren Läufe prüften Last und Speicher unter der Sperre.

### DoD

| Punkt | Stand | Beleg |
|---|---|---|
| 1 Ursache gemessen (L, C, s mit Methode, Last, Messendem; derselbe Aufbau nach der Behebung) | **erfüllt**, Stichprobe bestätigt | Plan §3; eigene Messung unten. Die Reihenfolge „vor der Behebung“ ist aus `git` nicht prüfbar (ein Handoff-Commit); die Werte „vorher“ sind auf dem unveränderten Code nachgemessen |
| 2 Behebung (`SPEC-105` Nr. 14, `SPEC-135` Nr. 8, 9, 18; R01 bis R21 mit Fänger und Mutation) | **erfüllt** | Code gelesen; Verhalten am Binary unten; Tests grün; eigene Mutationen unten (V-2) |
| 3 Erneuter Abnahmelauf | **Lauf erbracht**, Closure offen | Schritt 2: p95, Last, Messender, Commit. Eintrag in die Closure-Notiz, Bewegung von [CO-001](../plan/carveouts/done/CO-001-lasttest-replay-ohne-p95.md) nach `done/` und Bindung in `harness/README.md` stehen aus |
| `make gates` grün | **erfüllt** | oben, mit dem Vorbehalt zum Cache; ungecachte Testläufe grün |
| Review mit Report | **erfüllt** | Report liegt vor; Ausgänge unten |
| Doku-Update (`SPEC-105`, `SPEC-135`) | **erfüllt** | auf `main` seit `a4b79211` |
| Handbuch bei sichtbarem Verhalten | **gilt nicht** | kein neuer Schlüssel, Code, Kommando; §7 muss das noch sagen |
| Closure-Notiz, Register, Ausgänge §6, Paarungen | **offen** | Closure. Die Häkchen der DoD sind im Plan nicht gesetzt |

### Verhalten am gebauten Binary

Binary aus `9c3f5f10`; Aufzeichnung im Modus `record` (`$metadata`, A, B, A, A), Import und Workspace über das `Handle`, Abspielen im Modus `replay`. Container ohne Netz aus einem Image `FROM scratch`.

| Strategie | Folge | Ergebnis |
|---|---|---|
| `sequential` | A, B, A, A, A, unbekannt, B, B | v1, B v1, v2, v3, 410 `ODR-E5002`, 404 `ODR-E5001`, 410, 410 |
| `strict` | `$metadata`, A, A, unbekannt, B, A, A, A | 200, v1, 409 `ODR-E5007`, 404 `ODR-E5001`, B v1, v2, v3, 410 `ODR-E5002` |
| `first` | A, A, B, A, unbekannt | v1, v1, B v1, v1, 404 |
| `last` | A, A, B, A, unbekannt | v3, v3, B v1, v3, 404 |

Dazu: ohne Zuordnung 400 `ODR-E5003`, unbekannte Session 404 `ODR-E7003`, unbekannter Workspace 404 `ODR-E7001`. Alles entspricht `SPEC-135` Nr. 5, 6 und 9.

**Ohne beschreibbares Temp-Verzeichnis (Review F-2, vom Implementer nicht nachgestellt).** `--read-only`, `tmpfs` allein für `DATA_DIR`, weder `/tmp` noch `/var/tmp` noch `/usr/tmp` vorhanden, Arbeitsverzeichnis `/`, das Binary ohne `TMPDIR`; 3102 Exchanges, darunter 100 eines Keys mit 40 000 Bytes Header und 3000 eines Keys.

| Binary | Key mit 100 großen Treffern | übrige Keys |
|---|---|---|
| `9c3f5f10` | 200 in `sequential`, `strict`, `first`, `last` | 200 |
| Gegenprobe: `recording.go` von `fece6d88` (`ORDER BY` mit `INDEXED BY`) | 500 `ODR-E9001` in allen vier Strategien | 200 (auch 3000 kleine Treffer) |

Der Befund ist damit nachgestellt und die Behebung am Binary bestätigt.

### Stichprobe der Messung (L, C, s)

Aufbau des Lasttests (Profil `abnahme`, Startwert 1, 10 000 Exchanges, Aufbau 4 min 41 s), ein Container ohne Netz mit 2 CPU und 4 GiB, Last des Hosts 3,4 bis 5,6. L direkt am Adapter; C und s am gebauten Binary über HTTP (der Implementer maß den Dienst direkt). „vorher“ ist das Binary aus `17086a9a` (`src/` gleich `0ee80d43`).

| Größe | Plan vorher / nachher | eigene Messung vorher / nachher |
|---|---|---|
| L gehalten | 15,60 MiB / unter Auflösung | `ListBySession` 15,62 MiB / `ListByMatchKey` 0,006 MiB |
| L `TotalAlloc`, Allokationen | 45,9 MiB, 708 000 / 0,055 bis 0,063 MiB, 564 bis 585 (je `Replay`) | 45,9 MiB, 708 417 / 0,012 MiB, 181 (je Aufruf des Ports) |
| C CPU je Anfrage | 106 bis 108 ms / 0,67 bis 0,78 ms | 106,4 ms (50 Anfragen) / 1,05 ms (3000, mit HTTP und Telemetrie) |
| `VmHWM` k = 1, 8, 32, 64 | 42, 220 bis 227, 848 bis 883, 1550 bis 1770 / 14, 17, 21 bis 24, 26 bis 32 MiB | 44, 212, 824, 1546 / 23, 26, 32, 41 MiB (Prozess mit Listener) |
| s je gleichzeitiger Anfrage | 24 bis 29 / 0,2 bis 0,5 MiB | 21 bis 25 / 0,20 bis 0,27 MiB |

Die Größenordnungen und die Folgerung (K1 erklärt Speicher und CPU) sind bestätigt. Im Lauf C „nachher“ endeten 50 von 3000 Anfragen nicht mit 200: die Zahl gleicht den 50 Anfragen, die der Lauf „vorher“ in derselben Replay-Session schon verbraucht hatte (Artefakt der Sonde, Code nicht mitgeschrieben); alle übrigen Läufe der Sonde mit frischer Replay-Session hatten nur 200. Unter k = 64 endete in diesem Aufbau keine Anfrage mit `ODR-E9001`, die Wandzeit der 64 war aber 4,1 s (vorher 5,6 s) bei `SQLITE_BUSY_TIMEOUT` 5 s.

### Fix-Runde gegen die Klassen (`AGENTS.md` §3.17)

| Befund | Ausgang | Prüfung |
|---|---|---|
| F-1 (Aufwand wächst mit der Session) | gedeckt, mit benanntem Negativ | Q4, Q5 rot |
| F-2 (temporäre Struktur) | Verhalten behoben; Fänger deckt die zwei Konstanten, nicht die Aufrufstelle | am Binary; Q1, Q2 grün (V-2) |
| F-3 (Ordnung) | gedeckt | Test gelesen, grün; Mutationen des Plans |
| F-4 (`SPEC-216` Nr. 6 g) | berichtigt | Wortlaut gleich `tools/schema/check.py` (`SELECT id`, ohne `ORDER BY`) |
| F-5 (Abbruch des Treibers) | Negativ in §3, §6 offen an den Architect | gelesen |
| I-1 (Warten auf den Schreiber) | `SPEC-135` *Ungeprüft* | gelesen |
| I-2 (Einheit), I-3 ([CO-001](../plan/carveouts/done/CO-001-lasttest-replay-ohne-p95.md) „Ursache“) | berichtigt bzw. §6 offen an den Planner | gelesen |
| I-4 (beschädigte Datei) | die zwei Abfragen gedeckt, das Lesen der Zeilen nicht | Q7 grün (V-3) |

### Eigene Mutationen

Wegwerfkopie von `9c3f5f10`, je Mutation `go test -count=1` über `src/adapters/driven/sqlite` und `src/hexagon/services/replay` in einem Container ohne Netz; Nullprobe grün.

| Mutation | Ergebnis |
|---|---|
| Q1: an der Aufrufstelle `abfrageTreffer+" ORDER BY +sequence_no"` (temporäre Sortierung) | **grün** |
| Q2: an der Aufrufstelle `INDEXED BY idx_exchange_match` und `ORDER BY sequence_no` (der Zustand aus F-2) | **grün** |
| Q3: `SELECT DISTINCT` in `abfrageTreffer` | grün, gleichwertig (Plan der Abfrage unverändert) |
| Q4: dritte Abfrage `count(*) … length(match_key) > 0` über die Session | rot (`TestListByMatchKeyDurchlaeuftNichtDieSession`) |
| Q5: größte Nummer über `ORDER BY sequence_no+0 DESC LIMIT 1` | rot (`TestListByMatchKeyNutztDenIndex`, `TestListByMatchKeyDurchlaeuftNichtDieSession`) |
| Q6: `NOT INDEXED` in `abfrageTreffer` | rot (drei Tests) |
| Q7: `sammle` ohne `klassifiziere` bei `rows.Err()` | **grün** |
| Q8: größte Nummer über den Pool statt über die Verbindung des Standes | rot (`TestListByMatchKeyGegenAppend`, `TestListByMatchKeyGegenDelete`) |
| Q9: `LIMIT 10` an den Treffern | **grün** |

### Plan gegen Code

- Die Dateien der Tabelle in §3 sind geändert, keine weitere; `repositories_test.go` unverändert wie angekündigt. Keine Migration, kein neuer Index, nichts unter `tests/lasttest/`, `Makefile`, `harness/mk/`, `Dockerfile`.
- Produktcode rund 100 Zeilen (Erwartung 80 bis 120); Messung 1075 Zeilen gegen Marke 2700.
- DoD 2 nennt „R01 bis R21“, die Tabelle führt auch R22 (aus der Fix-Runde, mit Fänger und Mutation).

### Befunde

| ID | Kategorie | Befund | Adresse |
|---|---|---|---|
| V-1 | MEDIUM, **entschieden** | DoD 3 verlangt den Lauf „auf einem Commit, der die Behebung trägt“; der Auflösungs-Trigger von [CO-001](../plan/carveouts/done/CO-001-lasttest-replay-ohne-p95.md) verlangt ihn „auf einem Commit von `main`“. `9c3f5f10` war beim Lauf kein Commit von `main`. Entscheidung des Auftraggebers vor Schritt 2: gemessen wird auf `9c3f5f10`; der Zweig wird vor slice-bootstrap-m6-transport nach `main` gemergt, ohne dass sich dazwischen `src/` oder `tests/` auf `main` ändern; damit ist der gemessene Commit ein Commit von `main` mit gleichem `src/` und `tests/`. [CO-001](../plan/carveouts/done/CO-001-lasttest-replay-ohne-p95.md) schließt der Planner nur bei `abnahme=ja` | Auftraggeber (Merge), Planner prüft bei der Closure `git diff 9c3f5f10 <Merge-Commit> -- src tests` (leer) |
| V-2 | LOW | `SPEC-105` Nr. 14 („keine Abfrage der Operation lässt die Datenbank sortieren“): der Fänger prüft die Texte der zwei Konstanten, nicht die ausgeführten Anweisungen. Ein `ORDER BY` an der Aufrufstelle, auch in genau der Form von F-2, bleibt grün (Q1, Q2). Heute führt die Operation die Konstanten unverändert aus; das Negativ steht nicht im Plan | Implementer oder Planner (Negativ) |
| V-3 | INFO | Beschädigte Datei beim Lesen der Zeilen (nicht schon bei der Abfrage): kein Test über `ListByMatchKey` (Q7); der Helfer `sammle` ist gemeinsam und vom Slice nicht geändert | Planner |
| V-4 | INFO | Dass alle Treffer geliefert werden, ist bis 9 Treffer geprüft (Q9). Die Sonden am Binary zählen die Treffer eines Keys nicht und belegen darüber hinaus nichts | Planner |
| V-5 | INFO | Außerhalb der DoD: spielt eine Replay-Session alle 10 000 Anfragen, wächst die CPU je Anfrage mit dem Cursor von 1,1 auf 3,1 ms und p95 am Client von 10,7 auf 14,4 ms. Der Lasttest wechselt die Replay-Session alle 100 Anfragen und sieht das nicht | Planner, Validator |
| V-6 | INFO | Roter Lauf `TestLeserHaeltNieAn`: außerhalb des Slice (`tests/lasttest/` unverändert). Der Test verlangt, dass 20 MiB unter dem Detektor binnen 1 s Echtzeit geschrieben und gelesen sind (`warteAuf(…, time.Second)`). Eigene Läufe: 15 von 15 grün bei Last um 7, der ganze Test 1,5 bis 1,8 s; der volle Race-Lauf grün. Ein Fänger an einer Zeitgrenze, unter hoher Last des Hosts rot; kein Befund am Produkt. Ein `make gates` nach dem Merge kann daran rot werden | Planner |

Aus dem Review offen und hier nicht neu bewertet: F-5 (Architect), die Frage an den User zur Zahl gleichzeitiger Requests.

### Vorbedingungen für Schritt 2 (aus Schritt 1)

Ein Aufruf im Hintergrund; für rund 60 Minuten keine weitere schwere Arbeit auf dem Host; Last (1 und 5 min) beim Start unter 5; mindestens 8 GiB verfügbar; keine `LASTTEST_`-Variable; kein Container des Repos. Zur Ursache bei Fehlern: das Werkzeug wertet von stderr des Binary allein die Zeilen der Spans aus und verwirft die übrigen; ein `ODR-E9001` mit `SQLITE_BUSY` wäre an der Ausgabe nur als Zahl `fehler` sichtbar.

---

## Schritt 2: Aufruf `abnahme`

Ein Aufruf mit den Vorgaben auf `9c3f5f10`, am 2026-11-27 von 12:33:03 bis 13:21:50 (49 Minuten), unter `flock /tmp/odata-recorder-heavy.lock`; Messender ist der Verifier. Kein zweiter Aufruf (`SPEC-216` Nr. 1). Exit 0.

`env -u LASTTEST_BETRIEBSART -u LASTTEST_PROFIL -u LASTTEST_ZIEL_MS -u LASTTEST_STARTWERT flock /tmp/odata-recorder-heavy.lock make -C /Development/odata-recorder-wt-replay lasttest`

### Ausgabe, wörtlich

stdout (ohne die Zeilen, mit denen `make` das Verzeichnis und den Aufruf des Treibers nennt):

```
lasttest: Start profil=abnahme betriebsart=beide ziel_ms=100 startwert=1 kerne=20
lasttest: Aufbau exchanges=10000 entities=10000 daten_mib=69 dauer_s=283.098
lasttest: Messung betriebsart=replay lauf=1 anfragen=13200 spans=13200 messphase=12000 fehler=0 fremde_zeilen=1 p95_ms=10.613 max_ms=24.354 kind=12000 kind_p95_ms=10.547 kind_anteil=99.6 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=1 verspaetung_max_ms=1.312 bytes=32835588 treiber_cpu_s=7.663 gueltig=ja grund=ok
lasttest: Last betriebsart=replay lauf=1 proben=23 min=1.33 max=5.37 schwelle=10.00
lasttest: Messung betriebsart=replay lauf=2 anfragen=13200 spans=13200 messphase=12000 fehler=0 fremde_zeilen=1 p95_ms=10.399 max_ms=30.864 kind=12000 kind_p95_ms=10.352 kind_anteil=99.6 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=1 verspaetung_max_ms=3.833 bytes=32835588 treiber_cpu_s=7.190 gueltig=ja grund=ok
lasttest: Last betriebsart=replay lauf=2 proben=23 min=1.79 max=3.05 schwelle=10.00
lasttest: Messung betriebsart=mock lauf=1 anfragen=13200 spans=13200 messphase=12000 fehler=0 fremde_zeilen=1 p95_ms=31.509 max_ms=58.986 kind=12000 kind_p95_ms=31.260 kind_anteil=97.8 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=2 verspaetung_max_ms=2.472 bytes=97900912 treiber_cpu_s=6.981 gueltig=ja grund=ok
lasttest: Last betriebsart=mock lauf=1 proben=23 min=1.59 max=2.99 schwelle=10.00
lasttest: Messung betriebsart=mock lauf=2 anfragen=13200 spans=13200 messphase=12000 fehler=0 fremde_zeilen=1 p95_ms=32.648 max_ms=120.327 kind=12000 kind_p95_ms=32.421 kind_anteil=97.7 ohne_kind=0 kind_ohne_eltern=0 gleichzeitig_max=3 verspaetung_max_ms=3.294 bytes=97900912 treiber_cpu_s=7.298 gueltig=ja grund=ok
lasttest: Last betriebsart=mock lauf=2 proben=24 min=1.67 max=4.14 schwelle=10.00
lasttest: Urteil betriebsart=replay laeufe=2 ziel_ms=100 p95_ms=10.613 streuung_ms=0.214 toleranz_ms=2.000 ergebnis=bestanden grund=ok
lasttest: Urteil betriebsart=mock laeufe=2 ziel_ms=100 p95_ms=32.648 streuung_ms=1.139 toleranz_ms=3.150 ergebnis=bestanden grund=ok
lasttest: Abnahme ziel_ms=100 abnahmewert_ms=100 abnahme=ja grund=ok
```

stderr trägt allein die Ausgabe des Baus: keine Zeile `lasttest:`, keine Meldung von `make`. Es steht keine Zeile `Hinweis`.

### Host-Bedingungen

| Gegenstand | Wert |
|---|---|
| Host | 20 Kerne, 31 GiB |
| Vor dem Start (12:33) | Last 3,39 · 4,76 · 7,82; 16 GiB verfügbar; Swap 12 von 15 GiB belegt; `/docker` 54 GiB frei (89 % belegt); Sperre frei; keine `LASTTEST_`-Variable; kein Container des Repos |
| Nach dem Ende (13:21) | Last 3,89 · 3,23 · 3,03; 15 GiB verfügbar; Swap 11 von 15 GiB belegt; `/docker` 54 GiB frei; kein Container des Images |
| Proben des Programms | kleinste 1,33, größte 5,37, Schwelle 10,00; 23 · 23 · 23 · 24 Proben je Lauf |
| Fremde Container, liefen beim Start und beim Ende, unberührt | `d-migrate-mcp`, `d-migrate-mysql`, `d-migrate-postgres`, `d-migrate-oracle`, `d-migrate-mssql`, `k3d-viper-registry` |
| Kernel-Protokoll seit dem Start (`journalctl -k`) | keine Zeile zum OOM-Killer |
| Andere Arbeit | Ich habe während des Laufs nichts gestartet; der Teil-C-Review war nach Angabe des Auftraggebers beendet. Fremde Prozesse habe ich nicht mitgeschrieben; zur Last belegen allein die Proben |
| Stand | HEAD `9c3f5f10` vorher und nachher, uncommittet allein dieser Bericht; `main` vorher und nachher `a4b79211` |

### Urteil nach `LH-QA-01.a`

| Betriebsart | Urteil der Zeile | Was die Zeilen sagen |
|---|---|---|
| `replay` | `bestanden`, `grund=ok` | p95 des Gesamtspans 10,613 ms (Läufe 10,613 und 10,399 ms), Ziel 100 ms; Streuung 0,214 ms bei Toleranz 2,000 ms; `kind_p95` 10,547 und 10,352 ms, `kind_anteil` 99,6 (Match-Span); je Lauf 13 200 Anfragen, 0 Fehler, 12 000 in der Messphase |
| `mock` | `bestanden`, `grund=ok` | p95 32,648 ms (Läufe 31,509 und 32,648 ms), Ziel 100 ms; Streuung 1,139 ms bei Toleranz 3,150 ms; `kind_anteil` 97,8 und 97,7; je Lauf 13 200 Anfragen, 0 Fehler |
| Abnahme | `abnahme=ja grund=ok` | Profil `abnahme`, `betriebsart=beide`, `ziel_ms=100`, Startwert 1 |

`abnahme=ja` heißt nicht, dass der Zielwert bestätigt ist (`SPEC-216` Nr. 12 b); die Bestätigung durch den Owner steht aus.

Für die Closure-Notiz: p95 `replay` 10,613 ms, p95 `mock` 32,648 ms, je zwei Läufe; Messender Verifier; Commit `9c3f5f10`; Last des Hosts 3,39 vor dem Start, Proben 1,33 bis 5,37; 2026-11-27.

Keine Anfrage endete als Fehler; der Hinweis zu `SQLITE_BUSY` und `ODR-E9001` hatte in diesem Aufruf keinen Anlass. Der Auflösungs-Trigger von [CO-001](../plan/carveouts/done/CO-001-lasttest-replay-ohne-p95.md) ist mit `abnahme=ja` erfüllt, sobald die Bedingung aus V-1 gehalten ist.

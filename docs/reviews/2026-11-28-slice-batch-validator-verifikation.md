# Verifikation slice-batch-validator (Prüfung der Struktur eines `$batch`) — 2026-11-28

**Rolle:** Verifier (genau ein Lauf). **Stand:** Zweig `slice-batch-validator`, HEAD `173851be`. **Eingang:** Plan `slice-batch-validator` (DoD, §3 mit Bericht, Prüfgrenzen und Fix-Runde, §6), [Review](2026-11-28-slice-batch-validator-review.md) (Gegenstand `f264a2b4`), Fix-Runde `ae5a805b`.
**Bezug:** `LH-FA-16`, `LH-FA-16.a`; `SPEC-184`, `SPEC-182` Nr. 4; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md). **Messender:** der Verifier. **Nicht Teil der Erstabnahme;** `LH-FA-16` ist mit diesem Slice nicht erfüllt.

## Urteil

**Bestätigt mit Auflagen.** DoD 1 bis 3, Review und Gate-Nachweis sind erfüllt, der Produktcode folgt `SPEC-184`. Die Fix-Runde deckt die Klasse F-1. Die Klasse F-2 deckt sie nur für die Fragen, die die Weite-Fänger stellen; die Prüfgrenze sagt mehr zu, als gefangen ist. Ein MEDIUM, ein LOW, zwei INFO. Keine Auflage verlangt eine Änderung am Produktcode.

## Auflagen (vor `done/`)

1. **V-1** (an den Implementer): Fänger ergänzen oder die Prüfgrenze samt den Zeilen 37 und 40 auf den Stand des Fangs bringen, mit ungefangener Klasse und Maß.
2. **V-2** (an den Implementer, zusammen mit 1): Fänger oder benanntes Negativ.
3. **Closure** (an den Planner): offene Häkchen der DoD, Closure-Notiz mit Lerneintrag und dem Satz zum Binary, Register, Ausgang je Risiko in §6 (darin V-3), drei Paarungen.
4. **Integration:** der Commit dieses Reports ändert den Baum, der Gate-Stempel gilt für `173851be`; `harness/gegenproben.stempel` gehört zum Integrationsschritt und ist hier nicht geprüft.

## Findings

| ID | Kategorie | Befund | Quelle |
|---|---|---|---|
| V-1 | MEDIUM | Die Prüfgrenze sagt „jede Stelle gefahren bis 1000. Ungefangen bleibt eine Regel, die erst jenseits greift“. Vier Regeln, die ab Stelle 14 bzw. ab Abstand 3 greifen, bleiben grün: (a) ein Verweis auf den unmittelbar folgenden Request eines Changeset wird ab Request 14 angenommen; (b) `methode_nicht_unterstuetzt` in einem Changeset entfällt ab Request 14; (c) ein Einzelrequest mit Verweis wird ab Teil 14 angenommen; (d) ein Verweis löst gegen ein Changeset auf, das drei oder mehr Teile zurückliegt. Geführt sind: (a) vorwärts nur bis Länge 12, also 66 der 499500 Paare; die genannten 2042 gelten für `content_id_doppelt` und den Verweis rückwärts, nicht für vorwärts. (b) Requests 1 bis 12. (c) Teil 1 und 2. (d) Abstand 1 und 2. Die Zeile 37 („Verstoß und `Ort` an jedem Teil und Request 1 bis 1000“) gilt in einem Changeset ab Request 13 für `GET`, nicht für die anderen Methoden. | `AGENTS.md` §3.15, §3.17; Zeilen 37 und 40; `SPEC-184` Nr. 1, 4 |
| V-2 | LOW | Außerhalb eines Changeset tragen nur `GET` und `PATCH` einen Verweis in einem Fall; „ein `DELETE` außerhalb darf verweisen“ bleibt grün, schon an Teil 1. Daneben bleibt „ein `DELETE` mit Fortsetzung (`$id/x`) wird nie `referenz`“ grün: Fortsetzungen führt allein `POST`. Beides ist weder geführt noch als ungeführt benannt. Der Code unterscheidet hier nicht nach der Methode. | `AGENTS.md` §3.15 (Achsen einer Menge); `SPEC-184` Nr. 4 |
| V-3 | INFO | F-3 des Reviews steht in §6 mit Ausgang *offen* und Adresse Planner. `slice-batch-changesets` §1 führt den Aufruf von Leser und Prüfung, den Fang des Paars (Negativ zu Zeile 43) nennt er weiter nicht. | `AGENTS.md` §3.13 |
| V-4 | INFO | Die Maße 3990 und 2042 stimmen. Die Zählprüfungen der Weite-Fänger decken 2994 bzw. 1997 davon. Die übrigen 996 Paare aus Teil und Request (Teil 1 mit Request 3 bis 998) folgen aus der Schleife von `TestPruefeBatchWeiteRequests` und sind dort nicht als solche gezählt; die übrigen 45 Paare zweier Stellen (bis Länge 12) tragen `TestPruefeBatchContentID` und `TestPruefeBatchReferenz`. Keine Zählprüfung hält die Summen 3990 und 2042 selbst. | Plan §3, Prüfgrenzen |

## DoD-Konformität

| DoD | Ergebnis | Beleg |
|---|---|---|
| 1 Prüfung (`SPEC-184`) | erfüllt | `batch_pruefen.go` gegen `SPEC-184` Kopf und Nr. 1 bis 5 und `SPEC-182` Nr. 4 gelesen: Methoden je Ebene bytegleich, `Content-ID` je Changeset, `Verweis` ruft `ContentIDGueltig`, nur rückwärts, Rang Methode vor Eindeutigkeit vor Auflösbarkeit, `Ort` `n` bzw. `n/m`, keine Schreibzugriffe auf `b`. Klasse, Code und Sentinel prüft jeder Fall |
| 2 Raster, Fänger je Zelle, Contract-Test, `make a-check` | erfüllt, mit V-1 und V-2 | Beide Zeilen des Rasters tragen je Zelle Ort oder „gilt nicht“ mit Grund; die drei Paare der Zusatztabelle laufen in `TestPruefeBatchRang`; alle 17 in §3 genannten Tests stehen in der Datei; `TestPruefeBatchContract` führt Beispiel und vier Zeilen aus `SPEC-183` Nr. 8; `model` importiert hier allein `strconv` und `strings`, `.a-check.yml` gibt der Schicht keine ausgehende Kante |
| 3 `make gates` grün | erfüllt | `gates-passed.head` gleich `173851be`, `gates-passed.diffsha` gleich dem Hash des Arbeitsbaums (`d887797e…`, geschrieben 17:00:02), Baum sauber. Nicht neu gefahren; die Ausgabe des Laufs hat der Verifier nicht gesehen |
| Review | erfüllt | Report liegt vor. Die Fix-Runde ist nicht erneut reviewt; ihre Deckung steht unten |
| Doku-Update | erfüllt | Der Vertrag steht in `SPEC-184`; der Slice ändert keinen öffentlichen Vertrag |
| Katalog und Handbuch | erfüllt, Satz in §7 offen | Der Diff bringt keinen Code, Schlüssel oder Befehl; `PruefeBatch` und `Verweis` haben außerhalb von `model` keinen Aufrufer |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen | Sache der Closure |

## Plan gegen Code

`git diff main...HEAD` trägt vier Dateien: die zwei aus §3, den Plan und den Review-Report. Kein Gate, keine Konfiguration, keine Datei der Kettenslices. Die Messung 890 = 84 + 806 gegen die Basis `e27f718e` ist nachgerechnet und liegt unter der Marke 900. Produktcode ist seit `f264a2b4` unverändert; die Fix-Runde ändert Tests und Plan, wie §3 sagt.

## Maße der Prüfgrenzen

Unabhängig nachgezählt (Mengen aus den Schleifen der Tests, ohne deren Zählcode):

- Paare aus Teil und Request mit Summe bis 1000: 499500 gesamt, **3990** geführt (Teil *n* mit Request 1 und 2, Teil 1 mit Request *n*, Summe 1000). Das Gitter 12 × 12 bringt 110 weitere, wie die Zeile es gesondert nennt.
- Paare zweier Stellen eines Changeset bis 1000: 499500 gesamt, 1997 aus (erste oder vorige, *n*), 66 bis Länge 12, vereinigt **2042**. Welche Frage an diesen Paaren gestellt wird, sagt V-1.

## Fix-Runde gegen die Klassen des Reviews und Belege des Implementers

Ein Lauf von `tools/harness/mutationslauf.sh` unter `flock` hinter der laufenden Zeitmessung (ein Container, ohne Netz, 4g ohne Swap, 4 CPU, Image `odata-recorder:test`; 17:08:21 bis 17:10:16, Last 5,37 am Anfang und am Ende, 14 GiB verfügbar): 19 Mutationen, Klasse `andere`, **12 rot, 7 grün**, kein Übersetzungsfehler, kein Abbruch. Bei `grün` lief auch das ganze Paket `model`; außerhalb davon ruft niemand die Funktionen. Bei `rot` gibt das Werkzeug keine Ausgabe; dass der Grund der benannte ist, folgt aus dem Zuschnitt der Mutation.

| Gegenstand | Mutation | Ergebnis |
|---|---|---|
| Beleg 37 | `GET` im Changeset erlaubt | rot |
| Beleg 38 | nur die vorige `Content-ID` gemerkt | rot |
| Beleg 39 | auch `(` trennt | rot |
| Beleg 39, als gleichwertig gemeldet | `ende > 0` statt `>= 0` | grün, gleichwertig bestätigt |
| Beleg 40 | spätere Requests lösen auf | rot |
| Beleg 41 | Eindeutigkeit vor Methode | rot |
| Beleg 42 | `b` wird umgestellt | rot |
| Beleg 43 | `referenz` statt `content_id_doppelt` | rot |
| F-1, Mutation des Reviews | `DELETE` merkt seine `Content-ID` nicht | rot |
| F-1, neues Element | `PATCH` mit Header ohne Body ist nie doppelt | rot |
| F-1, neues Element | `PUT` mit Header und Body wird nicht gemerkt | rot |
| F-2, Mutation des Reviews | Teilnummer modulo 20 | rot |
| F-2, neues Element | ab Request 600 jede zweite `Content-ID` vergessen | rot |
| F-2, neues Element | V-1 (a) | **grün** |
| F-2, neues Element | V-1 (b) | **grün** |
| F-2, neues Element | V-1 (c) | **grün** |
| F-2, neues Element | V-1 (d) | **grün** |
| Nachbarklasse | `DELETE` außerhalb eines Changeset darf verweisen (V-2) | **grün** |
| Nachbarklasse | `DELETE` mit Fortsetzung wird nie `referenz` (V-2) | **grün** |

- **F-1:** die Klasse ist gedeckt (Träger und zweites Vorkommen × 4 Methoden × 5 Gestalten × 3 Abstände); beide neuen Elemente rot.
- **F-2:** gedeckt für die Teilnummer im `Ort`, die Zahl der Teile, Requests und gemerkten `Content-ID`, `methode`, den Selbstverweis, `content_id_doppelt` und den Verweis rückwärts. Nicht gedeckt und nicht benannt: V-1.

## Ungeprüft

`make gates` (Stempel passt), `make gegenproben` (Auftrag), `make test`, `make test-race`, `make lint`, `make a-check`, `make groessenmarke` einzeln. Die 58 und 25 Mutationen des Implementers sind nicht wiederholt, allein die Stichprobe oben. Die Laufzeiten der neuen Tests (0,6 s und 4,0 s) sind nicht nachgemessen. Welcher Fänger V-1 und V-2 schließt, ist nicht erprobt. Neben dem Container des Mutationslaufs lief allein nacheinander der von `make docs-check` für diesen Report.

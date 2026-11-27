# Review-Report: slice-replay-speicher-und-suche (Replay sucht nach Session und Match-Key) — 2026-11-27

**Review-Art:** Code, gegen Plan, Entscheidungen und Hard Rules (nicht gegen die DoD).

**Gegenstand:** slice-replay-speicher-und-suche, Zweig-Stand `fece6d88`; eigener Anteil `git diff 0ee80d43..HEAD` über die Pfade des Blocks. Die mit `main` gemergte Änderung an der Spezifikation ist nicht Gegenstand.

**Skill:** `.harness/skills/reviewer.md` @ `fece6d88` ·
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-27

**Eingangs-Kontext:**

- Plan slice-replay-speicher-und-suche (§1, §3 Messung, Zusage und Fänger R01 bis R21, Mutationen, Negative; §4; §6)
- `SPEC-105` Nr. 12 bis 14, `SPEC-135` Nr. 6, 8, 9, 18 und *Ungeprüft*, `SPEC-216` Nr. 6 g, `SPEC-004`
- `LH-QA-01`, `LH-QA-01.a`, `LH-FA-04`, `LH-FA-05`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)
- `AGENTS.md` (Hard Rules)

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Abfrage der größten Nummer darf die ganze Session im Index durchlaufen, ohne dass etwas rot wird: `max(sequence_no + 0)` und `count(*)` statt `max(sequence_no)` bleiben grün (der Plan der Abfrage nennt in beiden Formen `(session_id=?)`); bei 10 000 Exchanges 784 µs und 461 µs statt 16,5 µs, wachsend mit der Session. `SPEC-135` Nr. 8 sagt, der Aufwand wachse nicht mit der Session, und *Ungeprüft* dort nimmt genau dieses Wachstum von der Lücke aus; das Negativ des Plans nennt allein eine Abfrage *neben* den zwei geprüften, und der Kommentar des Tests sagt „keine durchläuft die Tabelle exchanges“. | `AGENTS.md` §3.15, §3.11; `SPEC-135` Nr. 8 und *Ungeprüft* | `src/adapters/driven/sqlite/list_by_match_key_test.go` · „die der größten Nummer über einen Index nach der Session“ | ja — beide Mutationen bleiben im Testpaket grün | Plan der Abfrage unterscheidet Suche und Durchlauf nicht |
| F-2 | MEDIUM | `abfrageTreffer` braucht mit `INDEXED BY idx_exchange_match` und `ORDER BY sequence_no` eine Sortierung (`USE TEMP B-TREE FOR ORDER BY`) über ganze Zeilen; ab einer Größe lagert SQLite in eine temporäre Datei aus. Ohne beschreibbares Temp-Verzeichnis endet `ListByMatchKey` mit `disk I/O error (6410)` (10 000 Treffer eines Keys; auch 100 Treffer mit 40 000 Bytes Header), während `ListBySession` auf denselben Daten gelingt; im Dienst wäre das `ODR-E9001` für jeden Request dieses Keys. Die Abhängigkeit ist neu im Replay-Pfad und steht weder in der Spezifikation noch im Plan, kein Test führt sie. Nicht HIGH, weil beides zusammentreffen muss (kein beschreibbares Temp-Verzeichnis, Treffer eines Keys im Bereich von MiB) und mit beschreibbarem `/tmp` alle Fälle gelangen. | `SPEC-105` Nr. 14; `SPEC-135` Nr. 8; `AGENTS.md` §3.12, §3.16 | `src/adapters/driven/sqlite/recording.go` · `AND ?2 <> '' ORDER BY sequence_no` | nein — kein Gate; Probe im Container mit gesperrtem `/tmp` und `/var/tmp` | Neue Abhängigkeit von temporären Dateien ohne entschiedene Randform |
| F-3 | LOW | Die Abfrage ohne `ORDER BY` bleibt grün: im Vertrag der Ablage werden die Exchanges in der Reihenfolge ihrer Nummern angehängt, die Fixture hat damit die erwartete Ordnung auch ohne Sortierung. R01 führt als Mutation allein die Ordnung nach der Kennung. Über den Port ist keine abweichende Reihenfolge erreichbar, im Betrieb folgt nichts. | `AGENTS.md` §3.15 (Fixture gleicht dem Wert der Mutation); `SPEC-105` Nr. 14 | `src/hexagon/ports/outbound/outboundtest/vertrag.go` · „die Kennungen gegen die Folge der“ | ja — Mutation bleibt im Testpaket grün | Fixture gleicht dem Wert der Mutation |
| F-4 | LOW | `SPEC-216` Nr. 6 g sagt, `make schema-check` belege den Plan der Abfrage; geprüft wird dort `SELECT id` ohne `ORDER BY`, und die Abfrage des Produkts wählte mit `ORDER BY` ohne `INDEXED BY` den anderen Index. Der Befund des Implementers ist am Text bestätigt und steht in §6 des Plans offen an den Architect; den Plan der Operation hält R10. | `SPEC-216` Nr. 6 g; `AGENTS.md` §3.11 | `harness/vertraege/SPEC-216-lasttest.md` · „belegt den Plan der Abfrage“; `tools/schema/check.py` · „Match-Key-Index wird genutzt“ | nein — Review | Vertrag sagt mehr zu, als ein Gate prüft |
| F-5 | LOW | Bei echtem, asynchronem Abbruch liefert `ListByMatchKey` selten `interrupted (9)` statt des Kontextfehlers (2 von 6 400 und 1 von 24 000 Aufrufen); im Dienst wird daraus `ODR-E9001`. `ListBySession` tut dasselbe (1 von 24 000), es ist also keine Verschlechterung. Der Haltepunkt von R06 liefert `Done` gleich `nil` und sieht den Weg nicht; das Negativ zu R06 sagt „Fehler des Kontexts oder voller Stand“ zu. Eine hängende Transaktion blieb nicht zurück (200 Folgeaufrufe ohne Fehler, Checkpoint nicht belegt). | `SPEC-105` Nr. 13; `SPEC-135` Nr. 2; `AGENTS.md` §3.15 | `src/adapters/driven/sqlite/recording.go` · `treffer, err = sammle(rows, leseExchange)`; Plan §3 · „R06, Abbruch während des Lesens der Zeilen“ | nein — Wirkung nur mit Wahrscheinlichkeit | Abbruch des Treibers statt Kontextfehler |
| I-1 | INFO | An den Architect: `ODR-E9001` mit `SQLITE_BUSY` unter 64 zugleich freigegebenen Transaktionen ist nach dem Wortlaut kein Verstoß gegen `LH-QA-01` (p95 bei 20 Requests/s; diese Last erzeugt der Lasttest nicht von selbst) und nicht durch den Diff entstanden (Messung des Implementers vor und nach der Behebung; die Transaktion ist unverändert, Code gelesen). Die Abbildung folgt `SPEC-109` Nr. 5 und `SPEC-135` Nr. 15. Es ist aber ein Verhalten des Produkts ohne eigene Randform: staut sich die Last, antwortet es nach `SQLITE_BUSY_TIMEOUT` mit 500, und ein einziger Fehler macht einen Lauf des Lasttests ungültig (`SPEC-216` Nr. 10 h). Selbst nicht nachgestellt. | `LH-QA-01`; `SPEC-135` Nr. 11, 17 | Plan §6 · „Nebenbefund der Messung“ | nein | Randform belegte Datenbank ohne Ort |
| I-2 | INFO | An Planner und Verifier: K1 trägt (eigene Probe, 10 000 Exchanges: `ListBySession` 68 ms und 619 739 Allokationen, `ListByMatchKey` 56 µs und 154 Allokationen). Die Zahlen nach der Behebung gelten für einen Treffer je Key (Daten des Lasttests). „rund 4,15 GB“ liest die `kB` des Kernels dezimal; 4147816 KiB sind rund 4,25 GB, die Folgerung bleibt (s × 185 gleich 4,3 bis 5,3 GiB gegen rund 3,96 GiB). | `AGENTS.md` §3.10 | Plan §3 · „rund 4,15 GB“ | nein | Einheit in der Auswertung |
| I-3 | INFO | CO-001 führt weiter „Ursache: nicht belegt“, während §3 des Plans K1 belegt; Sache der Closure. | `AGENTS.md` §3.9 | Carveout CO-001 · „Ursache: nicht belegt“ | nein | Berichtigung erreicht nicht jede Stelle |
| I-4 | INFO | Ohne `klassifiziere` an der Abfrage der größten Nummer bleibt der Vertrag der Ablage grün: kein Test führt eine beschädigte Datei durch `ListByMatchKey` (`ODR-E3001`). `SPEC-105` Nr. 14 sagt dazu nichts. | Maintainability | `src/adapters/driven/sqlite/recording.go` · `err := c.QueryRowContext(ctx, abfrageHoechsteSequenz` | ja — Mutation bleibt im Testpaket grün | Fehlerklasse je Operation ohne Fall |

## Eigene Mutationen und Läufe

Wegwerfkopie des Stands `fece6d88`, je ein Container ohne Netz (`--memory=4g`, 4 CPU) unter der gemeinsamen Sperre, Last des Hosts zu Beginn der Läufe 5 bis 9. Mutationen mit `tools/harness/mutationslauf.sh` (`MUTATIONSLAUF_SPEICHER=4g`, Klasse `andere`, ohne Detektor).

| Mutation | Ergebnis |
|---|---|
| `max(sequence_no + 0)`; `count(*)` | grün (F-1) |
| Abfrage der Treffer ohne `ORDER BY` | grün (F-3) |
| Fehler der ersten Abfrage ohne `klassifiziere` | grün (I-4) |
| `ROLLBACK` mit dem Kontext des Aufrufs; ohne `ROLLBACK` | rot (`TestListByMatchKeyAbbruch`) |
| `BEGIN IMMEDIATE`; Treffer über den Pool statt über die Verbindung des Standes | rot (`TestListByMatchKeyGegenAppend`) |
| `waehleStrict` über die ungefilterte Liste | rot (`TestKandidatenAlleinDerKey`) |
| Dienst gibt die Länge der Liste statt der größten Nummer weiter | rot (`TestStrictGroessteNummer`) |
| R21: der Wrapper überschreibt `ListByMatchKey` nicht (`go test`, ohne Werkzeug) | rot aus dem genannten Grund (Status 200, Zähler 0); Stand grün |

Proben außerhalb des Repos: Pläne beider Abfragen (`EXPLAIN QUERY PLAN`); Zeit und Allokationen bei 10 000 Exchanges; echter Abbruch (F-5); 32 Leser gegen 500 `Append` (32 229 Lesungen, kein Fehler, nie gemischter Stand); viele Treffer mit und ohne beschreibbares Temp-Verzeichnis (F-2). `go test -race` über `src/adapters/driven/sqlite` und `src/hexagon/services/replay` einmal grün.

Nicht gefahren: `make gates`, `make test`, `make gegenproben`, `make lasttest` (kein Profil), die Messung des Implementers (L, C, s) mit ihrem Aufbau, die ganze Suite von `src/bootstrap`, die Ursache des roten Ausgangsstands von `mutationslauf.sh` für `src/bootstrap`, Raten zu R07 und R08.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Replay-Verhalten: `sequential` und Cursor, `first`, `last` (Wahl nach `SequenceNo`, unabhängig von der Ordnung), `strict` mit p + 1 (`ODR-E5002`, `ODR-E5007`, `ODR-E5008`; gleichwertig bei lückenloser `sequence_no`), `ODR-E5001` zuerst, leerer Key, fremder Service, Löschen zwischen S1 und S3 (`replay.go`, `auswahl.go`) | geprüft, ohne Befund. Gewollte Änderung nach `SPEC-105` Nr. 14: ein unlesbarer Exchange mit anderem Key stört nicht mehr |
| Tabelle R01 bis R21 (vom Planner) gegen `SPEC-105` Nr. 14 und `SPEC-135` Nr. 8, 9, 18 Satz für Satz | jede Zusage hat eine Zeile; offen ist allein der Aufwand aus der Folge von Nr. 8 (F-1) |
| Beide Werte aus demselben Stand: Lesetransaktion, `Append`, `Delete`, Rücknahme ohne Kontext (`recording.go`) | geprüft, ohne Befund |
| SQL: Parameterbindung, `INDEXED BY` und Index in der Migration seit ihrer ersten Fassung, `NULL`- und leerer Key | geprüft, ohne Befund außer F-2 |
| Port-Vertrag, Fakes und Wrapper der Dienste und des Composition Root (`ports/outbound`, `services/*`, `src/bootstrap`) | geprüft, ohne Befund |
| Bestandstests (§6, Auftrag an den Reviewer): Umbenennung auf die neue Operation, Zählliterale 34 auf 35, 27 auf 28, 140 auf 145, 30 auf 31; `TestListBySession` in `TestListByMatchKey` | geprüft, ohne Befund: keine Zelle und keine Prüfung entfällt, `ListBySession` ist im Replay-Dienst neu ein Verstoß |
| Gates, `Makefile`, `harness/`, Lint-Profil (`AGENTS.md` §3.6, §3.2) | im eigenen Anteil unverändert, keine Suppression |
| Kommentare (`AGENTS.md` §3.7); „Ohne INDEXED BY wählt SQLite …“ als Kopplung an die Stelle gelesen, nicht als verworfene Alternative | geprüft, ohne Befund außer F-1 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 3 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Plan der Abfrage unterscheidet Suche und Durchlauf nicht · Neue Abhängigkeit von temporären Dateien ohne entschiedene Randform · Fixture gleicht dem Wert der Mutation · Vertrag sagt mehr zu, als ein Gate prüft · Abbruch des Treibers statt Kontextfehler · Randform belegte Datenbank ohne Ort · Einheit in der Auswertung · Berichtigung erreicht nicht jede Stelle · Fehlerklasse je Operation ohne Fall

## Verdikt

**Merge-blockierend:** ja, bis F-1 und F-2 einen Ausgang haben (Behebung oder entschiedenes Negativ mit Grund und Folge; die Randform von F-2 gehört nach `AGENTS.md` §3.12 dem Architect). Am Verhalten des Replay und an der Behebung der Ursache fand der Lauf keinen Fehler; die Daten des Lasttests (ein Treffer je Key) berührt F-2 nicht.

**Übergabe:** Findings an den Implementer; F-2, F-4, F-5 und I-1 gehen über den Plan §6 an den Architect, I-2 und I-3 an Planner und Closure. Der Report ersetzt keine Verifikation.

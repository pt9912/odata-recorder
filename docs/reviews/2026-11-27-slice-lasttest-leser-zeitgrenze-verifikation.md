# Verifikations-Report: slice-lasttest-leser-zeitgrenze — 2026-11-27

**Frage:** Bauen wir es richtig? Gegen DoD, Plan (§3 mit LZ1 bis LZ6, §6) und `SPEC-216` Nr. 9 e, 10 a, 14 a, 16 s.

**Gegenstand:** Zweig `slice-lasttest-leser-zeitgrenze`, HEAD `2170d89a` (`main` `aef8a827` enthalten), Arbeitsbaum sauber.

**Modell:** claude-sonnet-5-5 (Verifier) · **Datum:** 2026-11-27

**Bezug:** `LH-QA-06`, `LH-QA-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)

## Urteil

Die drei Liefer-Punkte der DoD und das Review sind belegt. Kein Befund, der die Lieferung blockiert. Offen sind allein die Punkte der Closure (Planner).

## DoD

| Punkt | Ergebnis | Beleg |
|---|---|---|
| Ursache belegt (Dauer von `pipeLauf` mit 20 MiB unter dem Detektor, mit Last und Messendem im Plan) | erfüllt | Plan §3; eigene Stichprobe liegt in den Bereichen des Plans (unten) |
| Fänger hängt nicht an der Grenze von 1 s, Mutationen der Zeile A1 je rot, mit Rate | erfüllt | `warteAuf` mit 1 s ist aus `pipeLauf` entfernt; Mutationen unten |
| `make gates` grün | erfüllt | Stempel: `gates-passed.head` gleich `2170d89a`, `gates-passed.diffsha` gleich dem jetzt gerechneten Hash des Arbeitsbaums (`93a9ef53…`), geschrieben 21:29, nach dem Commit von 21:10; nicht neu gefahren |
| Review mit Report | erfüllt | `2026-11-27-slice-lasttest-leser-zeitgrenze-review`: F-1 durch den Architect (`SPEC-216` Nr. 16 s) und das Negativ in §3, F-2 im Kopf von `pipeLauf`, I-1 und I-2 in §3 und §6 |
| Closure-Notiz, Beobachtungs-Register, Ausgänge der Risiken, Paarungen | offen | Closure steht aus; die Häkchen der DoD sind noch leer |

## Eigene Läufe

Ein Container aus `odata-recorder:race` (go1.27.1), Quellstand aus `git archive` von `2170d89a`, `--network=none --memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, Host mit 20 Kernen. Test-Binary mit `-race`, je Lauf ein Prozess mit `-test.count=1`. „Schleifen“ sind Endlosschleifen im selben Container. Die 1-Minuten-Last des Hosts lag über die Reihen zwischen 1,9 und 14,0.

### Mutationen

| Mutation | 0 Schleifen | 8 Schleifen | Meldung |
|---|---|---|---|
| LZ1: 1 ms Pause am Anfang von `nimm` | 10 von 10 rot | 5 von 5 rot | „warteten 5.232s auf die Uhr“ |
| LZ2: `lies` über `bufio.Scanner` mit 64 KiB | 10 von 10 rot | 5 von 5 rot | Write „closed pipe“, „token too long“, Spans 0 |
| LZ4: Grenze von 1 MiB entfällt | 10 von 10 rot | 5 von 5 rot | Zeile `l2` gelesen, 383 MB alloziert |
| LZ5: `nimm` ohne Sperre (`TestSammlerGleichzeitig`) | 10 von 10 rot | 5 von 5 rot | `DATA RACE` |
| LZ6: letzte Zeile ohne Zeilenende verworfen | 5 von 5 rot | nicht gefahren | Zeilen `s1 l1 s2` |
| LZ6: Fehler des Stroms verschluckt | 5 von 5 rot | nicht gefahren | Fehler `<nil>` |
| eigene e1: 1 ns Pause im Zweig `bufio.ErrBufferFull` von `lies` | 5 von 5 rot | 5 von 5 rot | „warteten 48ns auf die Uhr“ |
| eigene e2: Warten von 1 µs auf `time.After` vor der Rückkehr am Ende des Stroms | 5 von 5 rot | nicht gefahren | „warteten 1µs auf die Uhr“ |
| eigene e3: 1 ns Pause je fremde Zeile in `lies` | 5 von 5 rot | nicht gefahren | „warteten 99.409µs auf die Uhr“ |
| Kontrolle: LZ1 mit `gewartet < 0` statt `gewartet != 0` | 0 von 3 rot | nicht gefahren | bestätigt Plan §3: die Zusicherung über die Uhr trägt LZ1 allein |

Die eigenen Mutationen e1 bis e3 liegen an Stellen von `lies`, die weder Plan noch Review führen. Gefahren ist allein der genannte Test, nicht das Paket; `TestPaketOhneZeit` zählt hier nicht als Fänger.

### Unveränderter Leser und Stand davor

| Stand, Test | 0 Schleifen | 8 Schleifen | 16 Schleifen |
|---|---|---|---|
| neu, `TestLeserHaeltNieAn` | 0 von 10 rot | 0 von 20 rot | 0 von 15 rot |
| neu, `TestSammlerGleichzeitig` | nicht gefahren | 0 von 10 rot | nicht gefahren |
| alt (`679059d0`, nur `leser_test.go`), `TestLeserHaeltNieAn` | 0 von 5 rot | 4 von 8 rot | 3 von 3 rot |

Jeder rote Lauf des alten Stands meldet „nach 1s kein Wert“. Prozessdauer neu: 2,4 s, 4,9 bis 6,5 s, 9,7 bis 11,3 s (samt der Sekunde, die der Detektor am Ende wartet), weit unter der Wache von 60 s.

Dauer von `pipeLauf` mit 20 MiB (neu, je 5 Prozesse): 0,28 bis 0,31 s ohne, 0,92 bis 1,11 s mit 8, 1,78 bis 2,21 s mit 16 Schleifen. Der Plan nennt 0,28 bis 0,31, 0,81 bis 1,22 und 1,60 bis 2,69 s.

### Negativ

`SPEC-216` Nr. 16 s und Plan §3 nennen „eine Sperre, deren Halter nicht auf die Uhr der Blase wartet“ als nicht gefahren. Gefahren: je Gesamtspan in `nimm` Warten auf eine Sperre, deren Halter 1 ms in `syscall.Nanosleep` steht. Ergebnis 0 von 3 rot bei 13,7 s statt 2,4 s Prozessdauer. Das Negativ trifft für diese Form zu. Ein- und Ausgabe ist weiter nicht gefahren.

## Kommentar und Zusagen

| Prüfung | Ergebnis |
|---|---|
| Kopf von `pipeLauf` (`AGENTS.md` §3.11) | trifft zu: Uhr (Kontrolle), Ende vor dem Strom (LZ2), Zahlen der Zeilen, Grenze nach Nr. 16 s (Negativ oben). Die Sätze zum Hänger sind nicht neu gefahren; sie stehen auf m4 und m6 des Reviews |
| Kopf von `TestLeserHaeltNieAn` und `schreibeStrom` | trifft zu: 1 + 19 gleich 20 Läufe; Fälle LZ4 und LZ6 oben rot |
| Tabelle *Zusage · Fänger* gegen Nr. 9 e und 10 a | jede vom Slice berührte Zusage hat eine Zeile; die ungefangene Verlangsamung steht als Negativ mit Grund und Maß |
| `SPEC-216` Nr. 14 a | `pipeLauf` liest die Uhr über `Uhr` mit `systemuhr`; `TestPaketOhneZeit` läuft in den gestempelten Gates |
| Nr. 16 s, Stand-Vermerk | entfallen, wie in Plan §3 angekündigt; sonst keine Änderung am Vertrag im Zweig |

## Entscheidungs-Konformität und Plan-vs-Code

- `AGENTS.md` §3.6 und §3.2: keine Änderung an `Makefile`, `harness/mk/`, `Dockerfile`, Lint-Profil, `.a-check.yml`, `.d-check.yml` oder `tools/`; kein Gate gelockert. Die Reichweite des Fängers ist gegenüber 1 s Echtzeit bei Verlangsamung ohne Warten auf die Uhr kleiner; das ist eine Entscheidung am Vertrag (Nr. 16 s), keine Schwelle eines Gates.
- Plan gegen Diff: geändert sind `tests/lasttest/leser_test.go`, `SPEC-216` (Stand-Vermerk), Plan und Review-Report; keine Datei außerhalb von §3, kein Produktcode.
- Größenmarke: `git diff --numstat 679059d0..HEAD` über den Pfad des Blocks gleich 76 Zeilen, Marke 700.
- Ausdrücklich NICHT (§1): eingehalten.

## An den Planner

| ID | Kategorie | Punkt |
|---|---|---|
| V-1 | INFO | Die ersten zwei Risiken in §6 stehen auf „weiter offen: die Mutationen der Zeile A1 sind die DoD“. Die Mutationen sind rot (oben); der Ausgang ist bei der Closure zu setzen. |
| V-2 | INFO | Das Risiko zum Hänger, der das Test-Binary nach 60 s abbricht (§6), wartet auf den Ausgang bei der Closure. |

Nicht gefahren: `make gates` (Stempel passt), `make gegenproben`, `make lasttest`. `make docs-check` lief nach dem Schreiben dieses Reports.

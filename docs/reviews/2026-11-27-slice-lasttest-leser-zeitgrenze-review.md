# Review-Report: slice-lasttest-leser-zeitgrenze (Leser-Fänger ohne Grenze von 1 s Echtzeit) — 2026-11-27

**Review-Art:** Code, gegen Plan, Vertrag und Hard Rules (nicht gegen die DoD).

**Gegenstand:** slice-lasttest-leser-zeitgrenze, Zweig-Stand `a2100d4d`; eigener Anteil `git diff 679059d0..HEAD -- tests/lasttest` (eine Datei, `leser_test.go`). Die übrigen Änderungen des Bereichs stammen aus dem Start-Commit und dem Merge von `main` und sind nicht geprüft.

**Skill:** `.harness/skills/reviewer.md` @ `a2100d4d` ·
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-27

**Eingangs-Kontext:**

- Plan slice-lasttest-leser-zeitgrenze (§1, §3 mit LZ1 bis LZ6, Messung, Mutationen, akzeptiertem Negativ; §4; §6)
- `SPEC-216` Nr. 9 e, 10 a, 14 a, 16
- `LH-QA-06`, `LH-QA-01`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `AGENTS.md` (§3.6, §3.10, §3.11, §3.15)

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Ein Leser, der je Gesamtspan 1 ms auf einen Kanal wartet, der außerhalb der Blase entstand (Ticker aus `systemuhr`), hält den Schreiber an und bleibt grün (0 von 5 rot, Test 12,4 s statt 2,4 s), auch `TestPaketOhneZeit` bleibt grün. Negativ, Grenze im Kopf von `pipeLauf` und Mutationsachsen nennen allein Rechenzeit, Systemaufruf, `time.Sleep` und `time.After`; ein Warten, das `synctest` nicht als blockiert zählt, steht nirgends als geführte oder ungeführte Form. | `AGENTS.md` §3.15 | Plan §3 · „(Rechenzeit, blockierender Systemaufruf). *Maß:* 2 von 2“; `tests/lasttest/leser_test.go` · „(Rechenzeit, Systemaufruf), sieht der Lauf nicht“ | nein — kein Gate führt die Form | Fänger schmaler als Zusage, Menge ohne Achsenliste |
| F-2 | LOW | Der Kopf von `pipeLauf` sagt, ein nie endendes Lesen ende im panic von `synctest`, wenn alle Goroutinen der Blase blockiert sind, sonst von `wache`. Eine Warteschleife mit `time.Sleep` bei blockiertem Schreiber endet erst nach 60 s durch `wache`; und ein an sich endliches Warten auf einen Mutex, den eine auf die Uhr wartende Goroutine hält, wird in der Blase zum Hänger mit demselben Ende. | `AGENTS.md` §3.11 | `tests/lasttest/leser_test.go` · „Ein Lesen, das nie endet, endet im panic: von synctest“ | nein — Mutationen m4 und m6 | Kommentar sagt mehr zu, als zutrifft |
| I-1 | INFO | Die Hänger-Fälle m4 und m6 enden als panic des ganzen Test-Binary („Frist des Tests … überschritten“) nach 60 s, nicht als `--- FAIL` mit dem Grund; die übrigen Tests des Pakets laufen dann nicht. | Reviewer-Skill, Hinweis zu `panic:` | `tests/lasttest/leser_test.go` · „wache(t, 60*time.Second)“ | nein | Rot nicht aus dem benannten Grund |
| I-2 | INFO | An den Architect und den Planner: die offene Entscheidung über das Negativ (Plan §6) betrifft neben Rechenzeit und Systemaufruf die Form aus F-1; §4 nennt für eine ohne Echtzeitgrenze nicht fangbare Zusage die Rückführung und die Entscheidung des Architect. Der Stand mit 1 s hätte die Form aus F-1 nach Rechnung gesehen (über 5000 Wartezeiten zu je 1 ms), am alten Stand nicht gefahren. | `AGENTS.md` §3.13 | Plan §6 · „ob das Negativ trägt“ | nein | Offene Entscheidung an den Architect |

## Eigene Mutationen und Läufe

Kopie des Stands `a2100d4d` in einem Container ohne Netz und Mount (`--memory=4g --memory-swap=4g --cpus=4`, Image der Stufe `race`, go1.27.1), unter der gemeinsamen Sperre; Last des Hosts beim Start 3,8, 12 GiB verfügbar. Test-Binary mit `-race`, je Lauf ein Prozess mit `-test.count=1`, Test `TestLeserHaeltNieAn`.

| Lauf | Ergebnis |
|---|---|
| unverändert, ohne Zusatzlast (Last um 3,8); ohne `-race` | 0 von 10 rot (erster Lauf 2,4 s); 0 von 5 rot |
| unverändert, 8 Endlosschleifen im Container (1-Minuten-Last des Hosts 2,1 vor und 2,0 nach der Reihe, gedrosselte Prozesse zählt sie nur teilweise) | 0 von 10 rot (erster Lauf 5,7 s) |
| ganzes Paket `tests/lasttest` unter `-race`, ein Lauf | grün |
| m1: Senden an `s.neu` ohne `default` (blockiert für immer) | 3 von 3 rot: `--- FAIL`, panic „deadlock“ von `synctest` |
| m2: Senden an `s.neu` mit `time.After(time.Millisecond)` statt `default` | 3 von 3 rot: „warteten 5.231s auf die Uhr“ |
| m3: 1 ms je Gesamtspan auf einen Ticker von außerhalb der Blase | **0 von 5 rot** (F-1); `TestPaketOhneZeit` grün |
| m4: Warten auf einen Mutex, den eine 1 ms schlafende Goroutine hält | 1 von 1 rot: panic aus `wache` nach 60 s (F-2, I-1) |
| m5: Auswertung in eigener Goroutine über ungepufferten Kanal, 1 µs je Zeile | 3 von 3 rot: „warteten 5.232ms auf die Uhr“ |
| m6: Warteschleife mit `time.Sleep`, die nie endet, mitten im Strom | 1 von 1 rot: panic aus `wache` nach 60 s (F-2, I-1) |
| m7: Kontrolle LZ1, 1 ms Pause am Anfang von `nimm` | 3 von 3 rot: „warteten 5.232s auf die Uhr“ |
| m8: Kontrolle, `lies` endet mit `nil` mitten im Strom | 3 von 3 rot: Write „closed pipe“, Zahlen falsch; Ende ohne panic, also ohne hängende Goroutine |

Nicht gefahren: `make gates`, `make test`, `make test-race`, `make lint`, `make gegenproben`, `make lasttest`; die Messung und die Raten des Implementers mit ihren Zahlen und Zusatzlasten; der alte Stand; die zwei Formen des Negativs (`syscall.Nanosleep`, Rechenschleife); die Mutationen zu LZ4, LZ5 und LZ6 (ihre Zusicherungen ändert der Diff nicht).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tests/lasttest/`: Warten auf die Uhr und auf Kanäle der Blase, Ende vor dem Strom, Zahl der Zeilen | geprüft, ohne Befund außer F-1 |
| Goroutinen und Pipe in der Blase (`schreibeStrom`, gepufferter Fehlerkanal, `pr.Close`) auf Leck und Hänger am Stand | geprüft, ohne Befund |
| Go-Version: `go.mod` 1.27, Images 1.27 und 1.27.1, `testing/synctest` im Image gebaut und gelaufen | geprüft, ohne Befund |
| Flackern unter Zusatzlast (Stichprobe, 10 Prozesse je Reihe) | geprüft, ohne Befund |
| Zeitzugriff nach `SPEC-216` Nr. 14 a (`systemuhr` aus `uhr.go`, `TestPaketOhneZeit`) | geprüft, ohne Befund |
| `AGENTS.md` §3.6 und §3.2: Gate-Konfiguration, `GATE_CHECKS`, Lint-Profil, `//nolint` | geprüft, ohne Befund: der eigene Anteil ändert nichts davon; die geringere Reichweite des Fängers gegenüber 1 s Echtzeit steht als Negativ beim Architect (F-1, I-2) |
| `harness/vertraege/`, `spec/`: Zusagen von `SPEC-216` Nr. 9 e und 10 a | geprüft, ohne Befund: der eigene Anteil ändert keinen Vertrag, und keine andere lebende Stelle nennt die Grenze von 1 s für diesen Test |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 1 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger schmaler als Zusage, Menge ohne Achsenliste · Kommentar sagt mehr zu, als zutrifft · Rot nicht aus dem benannten Grund · Offene Entscheidung an den Architect

## Verdikt

**Merge-blockierend:** ja — F-1 (MEDIUM): die Form gehört in die laufende Entscheidung über das Negativ.

**Übergabe:** F-1 und F-2 an den Implementer, F-1 und I-2 zugleich an den Architect. Der Report ersetzt keine Verifikation.

# Review-Report: slice-etag-versionen — 2026-11-08

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-etag-versionen`, HEAD `4204a3a0` (Code des Implementers `7783f4f5`, Messung `a3d204ba`, Entscheidungen des Architect `a33588f0` und `4204a3a0`), Basis `af1f1bac`; 30 Dateien, Produktcode 110 Zeilen (`model/etag.go`, `entity.go`, `abfrage.go`, `mock/deuten.go`, `lesen.go`, `mock.go`, `schreiben.go`, `jsoncodec/schreiber_etag.go`, `schreiber_bau.go`, `bootstrap/konfig.go`, `konfig_werte.go`, `wurzel.go`), Tests, Plan §3 (Tabelle E1 bis E24, Absatz Mutationen und Raten), §6 (Raster 6 × 5, Zusatztabelle je Modus), Spezifikation `SPEC-280` bis `SPEC-282` und die Nachzüge.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08

**Eingangs-Kontext:**

- Plan von `slice-etag-versionen` (§1, §3, §6), `SPEC-280`, `SPEC-281`, `SPEC-282`, `SPEC-110` Nr. 2 und 15, `SPEC-168` Nr. 3 (e), `SPEC-169` Nr. 5, `SPEC-193` Nr. 4, `SPEC-194` Nr. 2 und 3, `SPEC-198` Nr. 8, `SPEC-049`
- `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18
- `LH-FA-15`, `LH-QA-02`, `LH-FA-07`
- ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (nur genannt)
- vorheriger Report am Modul: `2026-11-08-slice-mock-abnahme-review.md` (Form, Raten-Messung)

**Nicht geprüft (Rolle des Verifiers oder Planners):** DoD, `make gates`, Größenauslastung (1954 von 3000), der Ausgang der §6-Punkte *offen*. Die drei Punkte, die der Architect mit `4204a3a0` entschieden hat (E24, E6, `SPEC-110` Nr. 15), sind bewertet, nicht als Befund doppelt.

## Eigene Mutationen

Messender ist der Reviewer-Agent. Wegwerfkopie von `src/` und `tests/` je Mutation, Quellen schreibgeschützt eingehängt, Container aus dem Image `odata-recorder:deps`, ohne Netz, `--memory=4g --cpus=4`, jeder Lauf unter `flock /tmp/odata-recorder-heavy.lock`, `go test -count=1` des Pakets. Treiber ist ein Skript im Scratchpad des Reviewers (`/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/rev-etag-k7m3/mut.sh`, Mutationen in `rmuts.sh`), nicht `tools/harness/mutationslauf.sh`: dessen Nullprobe-Grenze trägt `src/bootstrap` nicht. Last des Hosts bei den Läufen 4,7 bis 12,4 (drei Slices parallel). Gewählt sind die vier vom Implementer nicht gefahrenen Mutationen und Elemente der Klassen, die der Implementer nicht nannte (`AGENTS.md` §3.17).

| Nr. | Zusage | Mutation (Wegwerfkopie) | Ergebnis |
|---|---|---|---|
| R1 | E14 (nicht gefahren) | `201` ohne Header `Location` | rot `TestEtagAusUndAn`, `TestEtagPost` (erster Lauf übersetzte nicht, `ort` unbenutzt; wiederholt mit `_ = ort`) |
| R2 | E22 (nicht gefahren) | Restore schreibt `Version: e.version + 1` (zählt fort) | rot `TestEtagRestoreStelltHer` |
| R3 | E13 „Rundlauf“ (nicht gefahren) | Schreiber schreibt den Etag-Text roh ohne Anführungszeichen (ungültiges JSON) | rot `TestEtagMitglied`, `TestEtagGenerator`, `TestEtagRundlauf` |
| R4 | E7 „zufälliger Anteil“ (nicht gefahren) | `etagVon` addiert `rand.IntN(2)` zur Version | rot `TestEtagWiederholung`, `TestEtagEinLesen`, `TestEtagAusUndAn`, `TestEtagGleichzeitigeAenderungen` und weitere |
| R5 | E9, Element nicht genannt | Collection: Etag nur an jedem zweiten Element | rot `TestEtagCollection`, `TestEtagAccept`, `TestEtagAusUndAn` |
| R6 | E9, Element nicht genannt | Collection: Etag nicht am letzten Element | rot (sechs Tests, darunter `TestEtagAusUndAn`, `TestEtagWiederholung`); die Mutation trifft auch Listen aus einem Element |
| R7 | E10, Fall nicht genannt | Projektion behält `Etag` nur, wenn alle Properties bleiben | rot `TestFuehreAusBehaeltEtag` |
| R8 | E5, Stelle nicht genannt (`204`) | `kopf204` prüft `Dienst.Etag` nicht | rot `TestEtagAusUndAn` |
| R9 | E5, Stelle nicht genannt (`201`) | `antwort` setzt `Entity.Etag` ohne `Dienst.Etag` | rot `TestEtagAusUndAn`, `TestEtagPost` |
| R10 | E15, Methode nicht genannt | `ETag` des `204` nur bei `PATCH`, nicht bei `PUT` | rot `TestEtagAusUndAn`, `TestEtagEinLesen`, `TestEtagMutationen` |
| R11 | E7 | Name der Kopfzeile `etag` statt `ETag` | rot `TestEtagAusUndAn` |
| R12 | E14 | `201` ohne Kopfzeile `ETag` (Mitglied bleibt) | rot `TestEtagAusUndAn`, `TestEtagEinLesen`, `TestEtagPost` |
| R13 | E9/E19 | Etag nur bei Version bis 3 | rot (`TestEtagGleichzeitigeAenderungen`, `TestEtagLesenWaehrendAenderung`, `TestEtagAusUndAn`, weitere) |
| R14 | E11, alle `Accept` mit Wert | Kopfzeile der Entity entfällt bei jedem nichtleeren `Accept` | rot `TestEtagAccept` |
| R17 | E4 | `etagFuerSatz` liefert auch bei leerer Version wahr | rot `TestEtagFuerSatz` |
| R18 | E21 | Reset schreibt die Baseline mit Version + 1 | rot `TestEtagResetStelltHer` |
| R20 | E20 Negativ (DELETE erhöht die Version nicht) | `DELETE` lässt die Version unverändert | rot in den vorhandenen Tests `TestDeleteWettlauf`, `TestRevisionWettlauf`, `TestFensterZuordnung`, `TestIsolationWorkspaces`, `TestZustandsmatrix`, `TestFolgen` (Paket vollständig gefahren): der Plan-Satz „mindestens sechs vorhandene Tests“ stimmt |
| R21 | E20 | `POST` über Tombstone und ohne Vorgänger schreibt die Version ohne + 1 | rot `TestEtagAusUndAn`, `TestEtagEinLesen`, `TestEtagPost` und vorhandene Tests |
| R22 | E15 | `PUT` behält die Version | rot `TestEtagAusUndAn`, `TestEtagEinLesen`, `TestEtagMutationen`, `TestEtagAndereBleiben` |
| R24 | E1, Klasse *Version* | `EtagFuer` addiert 1 für Versionen von 5001 bis 1999999999 | **grün** (Paket `model` vollständig), siehe F1 |

**Raten der Wirkungs-Fänger (E18, E19)**, Mutationen des Implementers, Messender der Reviewer, ein `go test -c` je Mutation, dann *N* Prozesse des Test-Binary mit `-test.count=1` (ohne Detektor), Last vor und nach dem Lauf 8,9 bis 11,6:

| Mutation · Test | Prozesse rot / *N* |
|---|---|
| Version vor der Transaktion gelesen · `TestEtagGleichzeitigeAenderungen` (Fake) | 10 / 10 |
| zweites Lesen für den Header · `TestEtagLesenWaehrendAenderung` | 10 / 10 |
| Version vor der Transaktion gelesen · `TestEtagGleichzeitigSQLite` | 5 / 5 |
| ohne Mutation · beide Tests des Fakes | 0 / 10 |

Das deckt sich mit den Zahlen des Plans (20 / 20, 20 / 20, 10 / 10). „Antwort liest nach `Put` erneut“ mit 0 / 20 ist äquivalent, weil das erneute Lesen unter derselben Sperre den eben geschriebenen Eintrag liefert; die Wirkung ohne Gleichzeitigkeit (zusätzlicher Aufruf von `Entities.Get`, Fake `wachsend`) fängt `TestEtagEinLesen` deterministisch, und der Plan sagt das. Kein Datenrennen liegt vor, der Plan benennt den Wirkungs-Fänger und nicht den Detektor, wie `AGENTS.md` §3.10 verlangt.

## Findings

### F1 · MEDIUM

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.17 und Reviewer-Skill (Fänger schmaler als Zusage, Menge unbeschränkt, Prüfgrenze fehlt als akzeptiertes Negativ)
- `pfad`: `src/hexagon/model/etag_test.go` · `for n := int64(0); n <= 300; n++ {`; Plan §3 Absatz *Generatoren und Mengen* · „`EtagFuer`: Schleife 0 bis 300 (301 Elemente) und 9 Literale“
- `befund`: Die Zusage von `SPEC-280` Nr. 1 gilt für jede Version von 0 bis 2^63−1; der Fänger führt 301 Werte und neun Literale, der Plan nennt die Prüfgrenze weder als akzeptiertes Negativ noch mit Grund und Folge. Die Mutation R24 (Versionen von 5001 bis 1999999999 um 1 erhöht) bleibt grün, eine Version in diesem Band erhielte im Betrieb einen falschen ETag.
- `verifizierbar`: ja — R24 in einer Baumkopie, `go test ./src/hexagon/model/` grün
- `klasse`: Menge der Zusage unbeschränkt, Fänger führt Anfang und Randwerte, Prüfgrenze nicht benannt

### F2 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.7 (ein Kommentar beschreibt, was da ist) und §3.11 (ein Kommentar sagt nur zu, was ein Test prüft)
- `pfad`: `src/hexagon/services/mock/fakes_test.go` · „jedes zweite Lesen eines Eintrags liefert eine höhere Version.“
- `befund`: Der Fake addiert zur Version die Zahl der bisherigen Aufrufe von `Get`, den eigenen eingeschlossen; schon das erste Lesen liefert eine höhere Version als der Eintrag, und jedes weitere eine höhere als das vorige. Der zweite Satz des Kommentars nennt „jedes zweite“.
- `verifizierbar`: nein — kein Gate liest Kommentare
- `klasse`: Kommentar nennt eine engere Wirkung als der Code

### F3 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15 und §3.16 (Raster: jede Zelle hat einen Fänger; der Plan nennt nicht mehr Elemente als der Test)
- `pfad`: Plan §3 Zeile E16 · „Kontext beendet vor und in der Transaktion“; `src/hexagon/services/mock/etag_test.go` · `{"Kontext in der Transaktion Put", "DELETE", "1"`
- `befund`: Die Zeile nennt den Kontext „vor“ der Transaktion; die Fälle von `TestEtagFehlschlag` speisen `context.Canceled` und `context.DeadlineExceeded` nur als Fehler von `Entities.Put` und `Workspaces.Advance` ein, ein beendeter Kontext vor dem Aufruf kommt nicht vor. Die Zellen *Abbruch* der Raster-Zeilen *GET Key* und *GET Collection* (`SPEC-281` Nr. 4 c) haben bei an keinen Fall; die Diff-Zeilen führen keinen neuen Aufruf eines Ports und berühren keinen Kontext, die vorhandenen Tests des Pakets (`lesen_fehler_test.go`, `lesen_nachzug_test.go`) decken den Abbruch beim Lesen bei aus.
- `verifizierbar`: nein — eine Mutation, die den Abbruch beim Lesen nur bei an anders behandelt, ist im Code nicht angelegt
- `klasse`: Plan nennt mehr Fälle als der Test, Zelle ohne eigenen Fänger bei Nebenwirkung null

### F4 · INFO

- `kategorie`: INFO
- `quelle`: Maintainability (Hinweis an Architect und Planner)
- `pfad`: `src/bootstrap/wurzel.go` · `return an && satz.Version == odataVersion`
- `befund`: Der Modus `mock` verdrahtet jeden Dienst mit `satzVersion4()`, `satz.Version` ist dort immer `4.0`; die Prüfung ist im Betrieb nicht erreichbar, ihren Wert trägt allein `TestEtagFuerSatz` (R17 rot). Sie schützt eine künftige Verdrahtung von Diensten der Version 2.0 und folgt dem Wortlaut von `SPEC-280` Nr. 2.
- `verifizierbar`: ja — `TestEtagFuerSatz`
- `klasse`: Wächter vor einer Verdrahtung, die es noch nicht gibt

### F5 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.15 (akzeptiertes Negativ mit Grund und Folge), Bewertung der drei Punkte des Architect
- `pfad`: `spec/spezifikation.md` · „**Akzeptiertes Negativ** (Zugriff): geprüft wird die Wirkung“; Plan §6 Abweichung E24
- `befund`: E6: die Mutation „der Dienst liest den Header und verwirft ihn folgenlos“ ist im Verhalten äquivalent; das Negativ in `SPEC-280` Nr. 3 nennt Grund (Map ohne Haken) und Folge (keine Wirkung auf Antwort oder Zustand) und genügt; das Raster der 72 Fälle fängt jede Wirkung (R-Probe `m_e6_ifmatch` des Implementers rot). `SPEC-110` Nr. 15: die Lesart (Wert in jedem Modus gelesen und getragen, nur `mock` gibt ihn weiter) steht jetzt im Wortlaut, `TestKonfigEtagEnabled` (3 Modi × 10 Werte) prüft sie. E24: das Beispiel als Datei ist an `slice-handbuch-mock` gegeben; die Handbuch-Zeile steht, ihr Inhalt stimmt mit dem Code überein (Mitglied nur in der Stufe `minimal`, auch je Element der Collection, `If-Match` ohne Wirkung); ob `make handbuch-check` das Entfernen der Zeile fängt, ist offen und gehört dem Verifier.
- `verifizierbar`: ja — `TestEtagHeaderKeinEinfluss`, `TestKonfigEtagEnabled`
- `klasse`: Entscheidung des Architect bestätigt

### F6 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.15 (Raster), Bewertung
- `pfad`: Plan §6 · Raster *Operation × Argumentart*, Zusatztabelle je Modus
- `befund`: Alle 30 Zellen nennen einen Ort, dessen Wortlaut die Entscheidung trägt (stichprobenhaft gegen `SPEC-280` Nr. 2, 4, 5 und `SPEC-281` Nr. 1, 2, 4 gelesen), und die fünf Zeilen der Zusatztabelle nennen einen Fänger (E3, E4, E5, E6, E7 bis E15, E23). Aufgelöst mit je einer eigenen Mutation: Zeilen GET Key, GET Collection, POST, PATCH/PUT, DELETE, Reset/Restore je mindestens eine von R1 bis R22 (Spalten *vorhanden*, *fehlt oder Tombstone*, *nicht aktiviert*, *Wiederholung*); Spalte *Abbruch*: siehe F3.
- `verifizierbar`: ja — R1 bis R22
- `klasse`: Raster aufgelöst

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/` (`etag.go`, `entity.go`, `abfrage.go`) | geprüft, ohne Befund außer F1; `Entity.Etag` in allen Konstruktionsstellen von `Entity` des Produktcodes (`projiziere`) übernommen |
| `src/hexagon/services/mock/` (Produktcode) | geprüft, ohne Befund; Header und Mitglied aus einem Lesen (`erg.Entities[0].Etag`, `neu` aus `Put`), Tombstone und Fehler ohne ETag, `DELETE` ohne ETag, `$count`/`$metadata` ohne ETag |
| `src/hexagon/services/mock/` (Tests, Fakes) | F2, F3; Fake von `Entities` hält Argumente und zählt jede Operation, E8 zählende Version greift (`gets != 1` und Wert) |
| `src/adapters/driven/odata/jsoncodec/` | geprüft, ohne Befund; 1827 Formen von `Entity.Etag` (165 gültig, 1662 ungültig) gegen ein Orakel mit `big.Int`, R3 rot |
| `src/adapters/driven/odata/jsoncodecv2/` | geprüft, ohne Befund; V2 schreibt nie ein Etag (Mutation „V2 schreibt es“ des Implementers; Dienst der Version 2.0 ist `NurLesend`) |
| `src/adapters/driving/httpodata/` | geprüft, ohne Befund; `TestAntwortEtag` hält Wert und Zahl der Kopfzeile |
| `src/bootstrap/` (Konfiguration, Wurzel, Abnahme) | geprüft, ohne Befund außer F4 (INFO); `TestAbnahmeMatrixMitEtag` hält 36 IDs und 92 Schritte, entfernt Kopfzeile und Mitglied und vergleicht den Rest byte-gleich; die bestehende Abnahme läuft mit `Dienst.Etag` aus unverändert |
| Nebenläufigkeit (E18, E19) | geprüft, ohne Befund; Raten 10/10, 10/10, 5/5 (Reviewer) gegen 0/10 ohne Mutation |
| Klassen (§3.17) | `Entity.Etag` 1827 Formen, `FuehreAus` 1120 Fälle (R7 an einem nicht genannten Fall rot), Collection-Elemente (R5, R6 rot), Aufrufstellen des Etag (`204`, `201`, Lesen: R8, R9, R12 rot); `EtagFuer` siehe F1 |
| Spezifikation `SPEC-280` bis `SPEC-282` und Nachzüge | geprüft, ohne Befund; keine Aussage widerspricht dem Lastenheft, `SPEC-280` Nr. 4 d hält `LH-QA-02` (Reset/Restore R18, R2 rot); Architektur-Sicht nicht berührt |
| Handbuch `docs/user/benutzerhandbuch.md` | geprüft, ohne Befund; Zeile stimmt mit dem Code überein |
| Hard Rules §3.2, §3.6 | geprüft, ohne Befund; kein `//nolint`, kein Gate geändert (`GATE_CHECKS` unverändert) |
| Zeichen (§3.18) | geprüft, ohne Befund; Rohbytes der Diff-Zeilen gegengelesen (`LC_ALL=C grep -P` auf `\u`-Folgen, BOM, NBSP): keine Treffer in den geänderten Dateien; Muster in Tests entstehen aus Bytewerten (`"\x5c"`) |

## Summary

0 HIGH, 1 MEDIUM, 2 LOW, 3 INFO.

- MEDIUM F1: Prüfgrenze von `EtagFuer` (0 bis 300 und neun Literale gegen die Menge 0 bis 2^63−1) steht nicht als akzeptiertes Negativ im Plan; Mutation R24 bleibt grün.
- Die vier vom Implementer nicht gefahrenen Mutationen (E14 `Location` fehlt, E22 Restore zählt fort, E13 Rundlauf, E7 zufälliger Anteil) sind rot; 15 weitere eigene Mutationen (R5 bis R14, R17, R18, R20 bis R22) sind rot, R24 ist grün.
- Die Raten des Plans bestätigt (Reviewer: 10/10, 10/10, 5/5 rot, 0/10 grün ohne Mutation); „liest nach `Put` erneut“ 0/20 ist äquivalent und von `TestEtagEinLesen` deterministisch gefangen.
- E6, E24 und `SPEC-110` Nr. 15 sind durch `4204a3a0` entschieden und bewertet (F5); sie sind kein Befund.
- Nicht gefahren: Ratenläufe unter `-race`, `make gates`, `make handbuch-check` (Rolle des Verifiers).

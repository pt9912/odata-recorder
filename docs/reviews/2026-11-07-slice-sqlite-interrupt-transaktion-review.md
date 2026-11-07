# Review slice-sqlite-interrupt-transaktion

* Rolle: Reviewer (`.harness/skills/reviewer.md`), einmaliges Review, kein Wiederholungsloop
* Gegenstand: Zweig `slice-sqlite-interrupt-transaktion`, Commits `bc22d5c7` (reiner Move) und `5d76a7b1` (Code, Tests, Plan); geprüft gegen Plan (§3 Tabelle Zusage · Fänger, §6 Raster), `SPEC-108` Nr. 1 (*Kontext in der Transaktion*, Sätze a bis f), `AGENTS.md` §3 und [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)
* Nicht Gegenstand: DoD und Gate-Läufe (Verifier)
* Messung: Mutationen in einer Wegwerfkopie von `src/` (nicht im Worktree), Container `odata-recorder:race`, `--memory 3g`, ohne Netz, unter `flock /tmp/odata-recorder-heavy.lock`, ein Container je Stapel, Tests `TestTransaktion*`, `TestAufrufstellen*`, `TestSteuer*` ohne `-race` (je Lauf etwa 22 s, Unmutiert grün). Die Raten der Prozesslaufe des Implementers habe ich nicht wiederholt (Host ausgelastet); die hier genannten Mutationen sind deterministisch (Klasse `andere`), ein Lauf je Mutation.

## Ergebnis in einem Satz

Der Code trägt den Vertrag in der Sache (Prüfung vor jedem Statement, Statement ohne Abbruch, `COMMIT` nach Prüfung ohne Abbruch, Zustand *unbrauchbar* nur aus Fehlern der Datenbank); kein HIGH. Zwei MEDIUM-Befunde betreffen Fänger, die schmaler sind als ihre Zusage (Satz f).

## Findings

### MEDIUM

**M-1**
* kategorie: MEDIUM
* quelle: `AGENTS.md` §3.15, §3.17; Reviewer-Skill (*Fänger führt nur eine Form der Eingabe*)
* pfad: `src/adapters/driven/sqlite/zugriff.go` · `if frei := v.steuere(ctx, "RELEASE operation"); scheitert == nil {`
* befund: Satz f nennt zwei getrennte Fehlerquellen im Fehlerpfad (Zurücknehmen und Abschluss des Savepoints), der Code behandelt sie getrennt (`scheitert`, `frei`), Fänger führt nur den Fall, in dem beide scheitern (`hebeSavepointAuf` hebt den Savepoint auf, `ROLLBACK TO` und `RELEASE` melden dieselbe Meldung `no such savepoint`). Gemessen: (M1) der Fehler von `ROLLBACK TO` im Fehlerpfad wird verworfen, der von `RELEASE` bleibt: alle Tests grün; (M2b) der Fehler von `RELEASE` im Fehlerpfad wird nicht geliefert, der von `ROLLBACK TO` bleibt: alle Tests grün. Plan Zeile 15 nennt als Mutation „Fehler von `ROLLBACK TO` wieder verworfen“ und als Fänger `Unbrauchbar/operation`; die gefahrene Mutation 15 des Implementers verwirft beide Fehler zugleich, die in der Tabelle genannte allein bleibt grün. Zeile 18 deckt nur `RELEASE` auf dem Erfolgspfad.
* verifizierbar: ja (Mutation M1 bzw. M2b in `zugriff.schreibe`, Tests `TestTransaktionUnbrauchbar`)
* klasse: Fänger deckt nur den Fall „beide Teilschritte scheitern“, nicht jeden einzeln

**M-2**
* kategorie: MEDIUM
* quelle: `AGENTS.md` §3.12, §3.16 (Zelle ohne Ort, im Code entschieden); Reviewer-Skill (*Paar von Ursachen ohne Fänger*)
* pfad: `src/adapters/driven/sqlite/zugriff.go` · `if v.unbrauchbar != nil {` (in `schreibe`, vor `SAVEPOINT operation`)
* befund: Trifft eine schreibende Operation auf eine unbrauchbare Transaktion und ihr Kontext ist beendet, liefert der Code den Fehler der Transaktion; Satz a verlangt bei beendetem Kontext vor dem ersten Statement dessen Fehler („der erste Schritt, der scheitert, bestimmt den Fehler“), Satz f verlangt für „jede weitere schreibende Operation“ den Fehler der Transaktion. Das Raster in §6 hat für das Paar (beendeter Kontext × unbrauchbar) weder Zelle noch Zeile der Zusatztabelle, und `Unbrauchbar/weitere` ruft die Operationen nur mit `context.WithoutCancel`; eine Vertauschung der Reihenfolge fängt kein Test aus dem Grund (meine Mutation M10 wurde rot, aber nur, weil sie einen zusätzlichen `Err`-Aufruf einführt und `Haltepunkt/position` die Zahl der Prüfungen zählt; die Reihenfolge selbst ist damit nicht gefangen). Das ist eine Schwester der offenen Randform „unbrauchbar vor oder nach der Prüfung der Argumente“ (§6, Adresse Architect), nicht dieselbe; der Architect kann beide in einem Satz entscheiden.
* verifizierbar: nein (Vertragslücke; nach der Entscheidung ein Fall in `Unbrauchbar/weitere` mit beendetem Kontext)
* klasse: Randform ohne Ort, im Code entschieden (Paar Kontext × Zustand)

### LOW

**L-1**
* kategorie: LOW
* quelle: `AGENTS.md` §3.9 (Plan folgt der Korrektur), §3.11
* pfad: `src/hexagon/ports/outbound/outboundtest/transaktion_operationen.go` · `Pruefungen int` (Kommentar „ein Statement je Prüfung“); Plan §3 Tabelle, Zeile `transaktion_vertrag.go`: „adapterunabhängig, ohne Zählung von Prüfungen“
* befund: Die Struktur `Operation` im adapterunabhängigen Paket trägt die Zahl der Statements des SQLite-Adapters je Operation (`Pruefungen`); die Plan-Zeile sagt das Gegenteil für die Suite. Die Zahl wird nur von den Tests im Paket `sqlite` gelesen, steht aber in einer Datei, die für jeden Adapter gedacht ist.
* verifizierbar: nein
* klasse: Aussage im Plan weicht vom Ort der Zahl ab

### INFO

**I-1** Zeile 4 (akzeptiertes Negativ) trägt. Gemessen: die Prüfung in `ExecContext` nach das Statement verlegt (M8) bleibt grün, wie die Zeile sagt; Grund (der `SAVEPOINT` ist erstes Statement jeder schreibenden Operation, Fehler der Datenbank der Listen kommen aus einem `QueryRow`) und Folge sind genannt. Die Mutation „Prüfung in `QueryRowContext` entfernt“ macht `Vorrang` rot (Zeile 3/4 des Implementers; Schleife über die 5 Einzelnen).

**I-2** Abweichungen 2 bis 6 des Implementers gegenüber der Tabelle des Architect (Zeile 1 Fänger `position` statt Suite für die Prüfung in `ExecContext`; Zeile 4 Negativ für `Exec`/`Query`; Zeile 6 Mutation „ganze Transaktion zurückrollen“; Zeile 10 und 19 ergänzt; Zeile 14 gleichwertige Prüfungen) tragen: jede nennt Grund, die Plan-Zeilen sind im selben Commit nachgezogen (§3.9). Zeile 14 ist ein Paar gleichwertiger Wächter (`d.sql.Conn`, `BEGIN IMMEDIATE`); jeder allein ist entfernbar, ohne dass etwas rot wird, mit Grund genannt, daher kein Befund.

**I-3** „unbrauchbar“ ist korrekt gebildet: der Zustand lebt je Transaktion (`txVerbindung`), nur `steuere` setzt ihn (Fehler der Datenbank aus `ROLLBACK TO`/`RELEASE`, nie ein Kontext, denn `steuere` läuft ohne Abbruch), der erste Fehler bleibt (M7, Überschreiben, rot), lesende Operationen laufen weiter (`weitere`, 11 lesende), `WithWorkspaceTransaction` liefert ihn nur, wenn `f` `nil` liefert (M9, Fehler von `f` ersetzt, rot). Entfernen der Prüfung in `schreibe` (M5) rot.

**I-4** Klassen statt genannter Mutationen (§3.17), an Elementen, die der Plan nicht nannte, rot gesehen: *Art des Kontextfehlers*: `ExecContext` liefert die Konstante `context.Canceled` statt `ctx.Err()` (M3) rot in der Suite an allen 8 schreibenden Operationen und in `Vorrang`; dasselbe in `QueryContext` (M4) an allen 6 Listen; `inTx` liefert vor `COMMIT` `context.Canceled` statt `ctx.Err()` (M6) rot in `transaktionFall4` (Art `DeadlineExceeded`/Ursache).

**I-5** Offene Randform „unbrauchbar vor oder nach der Prüfung der Argumente“ (Adresse Architect), bewertet: der Code prüft nach den Argumenten (`schreibe` wird erst nach der Argumentprüfung erreicht), das stimmt mit der Reihenfolge „Argument vor Kontext“ in `SPEC-108` überein und ist die naheliegende Lesart; ohne Ort im Vertrag bleibt es eine im Code entschiedene Zelle. Kein eigener Befund (nicht doppelt), Bewertung: kein Korrektheitsrisiko, nur ein fehlender Satz; siehe M-2 zur Erweiterung.

**I-6** Raten (§3.10): Plan nennt Zahl der Prozesse (20 nacheinander, 8 gleichzeitige), Last (24 Busy-Loops), Messenden (Implementer), `-race`, `-count=1` je Prozess und ein Host-Load-Average von 8 bis 70; Einzelzahlen sind als Größenordnung gekennzeichnet; das Zeitfenster von 500 ms der Zeilen 5/6 ist als offen benannt und durch die Haltepunkt-Fänger ohne Fenster ergänzt. Ich habe die Raten nicht wiederholt.

**I-7** Hard Rules: §3.2 kein `//nolint` im Diff und kein Symbol oder Dateiname mit „diag“ in `src/`; §3.18 in den geänderten Dateien unter `src/` weder eine Folge aus Backslash, `u` und vier Hex-Ziffern noch BOM, NBSP, Nullbreitenzeichen oder Steuerzeichen (`LC_ALL=C grep -P` über die Dateien des Diffs; in den Docs des Diffs keine Steuerzeichen); §3.7/§3.11 die Kommentare von `txVerbindung`, `inTx`, `schreibe`, `haltepunkt` und `TestTransaktionAbbruchHaltepunkt` sagen zu, was `TestTransaktionAbbruchImStatement`, `Haltepunkt/*`, `Unbrauchbar/*` und `Aufrufstellen` prüfen; der Kopf von `TestTransaktionAbbruchHaltepunkt` nennt das Offene (Auslöser außer `Err`).

## Auflösung der Tabelle Zusage · Fänger (je Zeile eine Mutation)

Von mir gefahren (rot = roter Test aus dem benannten Grund, kein Aufbaufehler): M3, M4, M5, M6, M7, M9 rot; M1, M2b, M8 grün (siehe M-1, I-1). Die übrigen Zeilen löse ich über die Mutationen des Implementers auf, soweit ihr Wortlaut sie trägt; nicht von mir wiederholt: Zeilen 5 bis 9, 11 bis 13, 16, 18, 19.

| Zeile | Ergebnis |
|---|---|
| 1 bis 3 | M3/M4 an Arten und Operationen rot (I-4); Zeile 1 Eigenheit der Prüfung in `ExecContext` nur durch `position` fangbar, wie der Plan sagt |
| 4 | Negativ für `Exec`/`Query` bestätigt (M8 grün), I-1 |
| 5 bis 9, 11 bis 13 | nicht wiederholt; Aufbau der Fänger geprüft: Haltepunkt ohne Goroutine und Zeitfenster, Schleifen über die Mengen (8, 6, 5, 31 Positionen) |
| 10 | siehe I-2; `ROLLBACK TO`/`RELEASE` ohne Kontext durch `steuere` |
| 14 | gleichwertige Wächter, I-2 |
| 15 | **M-1** |
| 16, 17 | M5, M7, M9 rot |
| 18 | `RELEASE` Erfolgspfad nicht wiederholt; Fänger `release` führt ihn getrennt vom Fehlerpfad |
| 19 | Zählprüfungen gelesen (24 Aufrufstellen 9/5/10, 7 Steuer-Statements, 19 Operationen, `go/parser`); nicht mutiert |
| 20, 21 | akzeptierte Negative mit Grund und Folge, tragen |

## geprüft, ohne Befund

* `src/adapters/driven/sqlite/tx.go`, `workspace_tx.go`: Prüfung vor `COMMIT`, `COMMIT` ohne Abbruch, `ROLLBACK` ohne Abbruch, Zustand *unbrauchbar* als Fehler der Transaktion; kein Befund außer dem in M-2 genannten Rangfragen.
* `src/adapters/driven/sqlite/export_test.go`, `transaktion_abbruch_test.go` (Umformung der Diagnose, kein `DIAG_N`, keine Umgebungsvariable), `aufrufstellen_test.go`: ohne Befund.
* `src/hexagon/ports/outbound/outboundtest/transaktion_vertrag.go`: ohne Befund (`transaktionFall4` je Art, Folge-Transaktion nutzbar).
* `docs/plan/planning/done/slice-sqlite-interrupt-transaktion.md`: §3 und §6 im selben Commit nachgezogen; Befunde nur L-1 und M-2 (Raster).

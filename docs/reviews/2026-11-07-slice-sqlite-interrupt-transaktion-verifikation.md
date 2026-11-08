# Verifikation slice-sqlite-interrupt-transaktion (Kontext-Abbruch beendet die Transaktion des SQLite-Adapters nicht mehr)

**Rolle:** Verifier (`Bauen wir es richtig?`, gegen Plan, DoD und Spezifikation). **Datum:** 2026-11-07.
**Gegenstand:** Zweig `slice-sqlite-interrupt-transaktion`, HEAD `3ea3b9fa` (Fix-Runde nach dem Review `cd53a653` und dem Entscheid des Architect `dea715a4`), Basis des Blocks `e8237239`, Abzweig von `main` bei `4d441211`.
**Eingang:** DoD-Bestätigung und Bericht des Implementers in §3 des Plans (Tabelle Zeilen 1 bis 23, Absatz *Mutationen und Raten*, *Fix-Runde zum Review*, *Natürlicher Ablauf der Diagnose*, *Laufzeit*, *Mengen mit Zählprüfung*), Review `docs/reviews/2026-11-07-slice-sqlite-interrupt-transaktion-review.md` (M-1, M-2, L-1).
**Bezug:** `LH-QA-03`, `LH-QA-06`, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md).
**Spezifikation:** `SPEC-108` Nr. 1 (Absatz *Kontext in der Transaktion*, Sätze a bis f), Nr. 17; `SPEC-109` Nr. 5 und 6; `SPEC-128` Nr. 1.

## Urteil

**Erfüllt mit Vorbehalt.** Kein HIGH, kein MEDIUM, kein Widerspruch des Produktcodes zum Vertrag. `txVerbindung` (`zugriff.go`), `inTx` (`tx.go`) und `WithWorkspaceTransaction` (`workspace_tx.go`) setzen `SPEC-108` Nr. 1 a bis f Satz für Satz um. `make gates` ist auf HEAD grün (Exit 0, siehe V-3 zum Cache). Die Größenmarke hält (1796 von 2700). Die drei Befunde der Fix-Runde (M-1, M-2, L-1) sind gegen die Klasse gedeckt: sechs eigene Mutationen an Elementen, die weder der Befund noch der Plan nannten, wurden alle rot, im benannten Grund. Die Last-Flake ist in meiner Messung nicht mehr zu sehen (12 von 12 Prozessläufen unter Last grün, Vergleichswert der Diagnose 19 Prozent rot).

Der Vorbehalt ist ein Zustand des Plans, kein Fehler des Codes: DoD-Häkchen, Closure-Notiz (§7), die Ausgänge der Risiken in §6 (sechs tragen noch „offen bis zur Closure“) und die Nachführung des Registers gehören dem Planner und der Closure. Dazu zwei kleine Befunde am Plan (V-1, V-2) und einer zur Aussagekraft des Gates (V-3).

## Eigene Sensoren (nicht die Behauptung, die Belege)

| Sensor | Ergebnis |
|---|---|
| `make groessenmarke` im Zweig (unter `flock /tmp/odata-recorder-heavy.lock`) | `ok, Messung 1796 von Marke 2700`; Messpunkt `zeilen=23 nebenlaeufig=6 fakes=1 produktcode=90 schaetzung=2160 marke=2700 messung=1796`. Hinweis des Werkzeugs: `Tabelle ergibt 2160, Schätzung 2000` (V-1) |
| Größenmarke von Hand nach `SPEC-128` Nr. 1: `git diff --numstat e82372390145be87c2b4f260eeb3901ba4d63a33..HEAD` über die beiden Pfade des Blocks (`src/adapters/driven/sqlite/`, `src/hexagon/ports/outbound/outboundtest/`) | Einfügungen plus Löschungen **1796** (darin `diag_interrupt_test.go` mit 266 Löschungen und die 190 Zeilen des umgeformten `transaktion_abbruch_test.go`), gleich der Meldung des Implementers; Auslastung 0,67 der Marke |
| `make gates` auf HEAD (sauberer Baum, `git status` leer; eigenes Log im Scratchpad) | Exit 0: `a-check` 0 Befunde, `d-check` 850 Dateien 0 Befunde, `baseline-verify` v6.18.1 OK, `schema-check` und `migration-check` ok, `zeichen-check` 1714 Dateien ok, Stufen `build`, `test`, `race` gebaut. Alle Docker-Stufen kamen aus dem Cache (V-3) |
| Paket `sqlite` selbst gefahren (Image `odata-recorder:race`, Baumkopie per `docker cp`, `--network none --memory 3g`, unter `flock`, Test-Binary mit `-race`) | `TestTransaktion*`, `TestAufrufstellen*`, `TestSteuer*` mit `-test.count=1`: PASS in 42 s (Host-Last 3 bis 14) |
| Last-Flake: 2 Runden zu 6 gleichzeitigen Prozessen (`-race`, `-test.count=1` je Prozess; `TestTransaktionVertrag`, `TestTransaktionAbbruchImStatement`, `TestTransaktionAbbruchHaltepunkt`, `TestTransaktionUnbrauchbar`), daneben 24 Busy-Loops im selben Container (Messender: der Verifier) | **12 von 12 Prozessläufen grün, 0 rot.** Vergleichswert vor der Behebung (Diagnose): 457 von 2400 (19 Prozent) rot unter 8 Prozessen und 24 Busy-Loops. Grenzen: 12 Läufe sind keine 2400; die Busy-Loops liefen mit `timeout 420` und waren am Ende der zweiten Runde womöglich abgelaufen (Größenordnung der Last, kein fester Wert, `AGENTS.md` §3.10); die natürliche Messung `DIAG_N=150` habe ich nicht wiederholt (der Quelltext liegt nur im Diagnose-Commit `1ad35ba3`) |
| Pfade | gegen `main` (`4d441211`) berührt der Zweig `src/adapters/driven/sqlite/`, `outboundtest/`, den Plan, die Evidenz des Registers, `state.md`, den Review-Report und `spec/spezifikation.md` (7 neue Zeilen, SPEC-108 Nr. 1). `Makefile`, `harness/mk/`, `GATE_CHECKS`, `.golangci.yml`, `.a-check.yml`, `Dockerfile` und `tools/` unverändert (Nr. 7 der Übergangsregeln gehalten) |
| Namen | `grep -rniE "diag|DIAG_N" src`: kein Treffer in den Dateien des Slice (Treffer nur im fremden Paket `inspection`, dort Fachwort `Diagnose`) |
| Rohbytes (`AGENTS.md` §3.18) | der Review hat die Dateien des Diffs gegengelesen; dieser Report ist unten gegengelesen |

## Plan gegen Code

- Dateien wie in §3: `zugriff.go` (Typ `txVerbindung` mit Feld `conn querier` und `unbrauchbar error`, `steuere`, `schreibe` mit Zustandsprüfung vor `SAVEPOINT`), `tx.go` (`ctx.Err()` vor `COMMIT`, `COMMIT` mit `context.WithoutCancel`), `workspace_tx.go` (Zustand als Fehler der Transaktion, nur wenn `f` `nil` liefert), sechs neue Testdateien, `export_test.go` (`TransaktionAktiv`, `ScheitereBei`, `StatementAufVerbindung`), `outboundtest/transaktion_operationen.go` (neu), `transaktion_vertrag.go` (erweitert). Keine Datei darüber hinaus.
- Sätze gegen Code, ohne Befund: a (Prüfung in `ExecContext`, `QueryContext`; `QueryRowContext` reicht den beendeten Kontext an `database/sql`, das den Fehler aus `Row.Scan` liefert), b/c (Statement ohne Abbruch, `RELEASE` und `ROLLBACK TO` ohne Kontext), d (Prüfung, dann `COMMIT` ohne Abbruch; `ROLLBACK` ohne Kontext), e (`BEGIN IMMEDIATE` behält den Kontext, Negativ Zeile 20), f (erster Fehler bleibt, beide Steuerstatements getrennt, Reihenfolge Argumente, Zustand, Kontext, Datenbank). Der Weg außerhalb von `f` (`conn == nil`) ändert sich nur am `COMMIT`, wie §1 und die Bestätigung des Planners sagen.
- Fix-Runde gegen den Plan (§3.9): L-1 erledigt (`Pruefungen` steht nicht mehr in `outboundtest`, `pruefungsZahl` im Paket `sqlite`, Plan-Zeile nachgezogen); M-1 und M-2 im selben Commit in §3 (Tabelle, Mutationen, Mengen) nachgezogen; die zwei Randformen in §6 tragen Ausgang *entschieden* mit Ort `SPEC-108` Nr. 1 f und Fänger Zeile 22 und 23.
- Spezifikation: `SPEC-108` Nr. 1 trägt a bis f vor dem ersten Code-Commit (`e8237239` vor `5d76a7b1`); `dea715a4` ergänzt die Reihenfolge der Prüfungen in f; Nr. 17, `SPEC-109` Nr. 5 und 6 unverändert, kein neuer Fehlercode (`make katalog-check` im Lauf von `make gates` grün).

## Tabelle Zusage · Fänger gegen den Vertrag

Jeder Satz von a bis f hat mindestens eine Zeile (a: 1 bis 4; b/c: 5 bis 11; d: 12, 13; e: 14, 20; f: 15 bis 18, 22, 23; Mengen: 19) oder ein benanntes Negativ mit Grund und Folge (4, 20, 21, Teil von 22). Die Test-Symbole der Tabelle existieren (`grep` über `src`: `transaktionKontextVorOperation`, `transaktionKontextVorDatenbankfehler`, `transaktionTeilwirkung`, `transaktionVerschluckteFehler`, `transaktionAbschluss`, `transaktionFall4`, `transaktionOperationenMenge`, `TestTransaktionAbbruchImStatement`, `TestTransaktionAbbruchHaltepunkt`, `TestTransaktionUnbrauchbar`, `TestAufrufstellenZaehlung`, `TestSteuerStatementsZaehlung`).

### Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Eigene Mutationen auf einer Baumkopie im Container (ohne `-race`, Klasse *andere*: deterministisch, ein Lauf je Mutation genügt), Tests `TestTransaktionUnbrauchbar`. Alle sechs betreffen Elemente, die der Befund und die Mutationen des Plans nicht nannten, alle rot im benannten Grund:

| Menge (Zahl) | Meine Mutation | Rot |
|---|---|---|
| M-1, Zeile 15: 2 Steuerstatements × Erfolgs-/Fehlerpfad, 8 Operationen × 5 Fälle = 40 Blätter | der letzte Fehler statt des ersten bestimmt den Zustand (`if v.unbrauchbar == nil` entfernt) | `steuerung/<Op>/Fehlerpfad_beide` 8 von 8, dazu `weitere` und `argument` (27 FAIL-Zeilen); kein Blatt der anderen Fälle |
| M-1, Fehlerpfad | der Fehler von `RELEASE` hat Vorrang vor dem von `ROLLBACK TO` | `steuerung/<Op>/Fehlerpfad_beide` 8 von 8 (27 FAIL-Zeilen) |
| M-1, Erfolgspfad | `RELEASE` des Erfolgspfads liefert den Fehler, merkt den Zustand aber nicht | `steuerung/<Op>/Erfolgspfad_nur_RELEASE` 8 von 8 (10 FAIL-Zeilen: 8 Blätter, `steuerung`, Eltern) |
| M-2, Zeile 23: 8 Operationen × 3 Arten = 24 | die Zustandsprüfung entfällt allein bei `Canceled` mit Ursache (`context.Cause(ctx) != ctx.Err()`) | `kontext/<Op>` 8 von 8 (10 FAIL-Zeilen); die Art `DeadlineExceeded` aus dem Plan ist damit nicht die einzige gefangene Art |
| M-2, Zeile 22: 8 Operationen = 6 Aufrufe + 2 benannte Negative | `Entities.Put` prüft den Zustand vor den Argumenten (nur dieses Element; der Plan nannte `Cursors.Put`) | `argument/Entities.Put`, sonst nichts (3 FAIL-Zeilen) |
| M-2, Zählung 6 + 2 | `UngueltigeAufrufe()` ohne den Eintrag `Cursors.Put` | `argument` rot durch die Zählung (`Fatalf`, 2 FAIL-Zeilen), also fängt die Schleife das Fehlen einer Operation |

Die benannten Negative von Zeile 22 (`Workspaces.Delete`, `Cursors.ResetAll`) trägt der Quelltext: beide rufen `schreibe` und prüfen die Kennung erst in der Datenbank (`pruefeWorkspace`), `SPEC-108` Nr. 5 und 12 nennen für sie nur `ErrNichtGefunden`. Die Folge im Plan stimmt.

Nicht von mir wiederholt, weil die Mutationen des Implementers und die des Reviews sie tragen und ihr Wortlaut zu den Zeilen passt: Zeilen 1 bis 14, 16 bis 19 (der Review löste M3 bis M7 und M9 selbst auf; die Haltepunkt-Tests ohne Zeitfenster stehen in Raten 20 von 20 und 24 von 24 im Plan).

## Befunde

**V-1 (LOW) Marke und Schätzung liegen an der alten Tabelle.** `make groessenmarke` meldet `Tabelle ergibt 2160, Schätzung 2000`: der Block wurde vor dem Code bei 21 Zeilen und 6 nebenläufigen Zeilen gesetzt (`SPEC-128` Nr. 1), die Tabelle hat jetzt 23 Zeilen (Zeilen 22 und 23 aus dem Entscheid `dea715a4`). Das Werkzeug urteilt `ok` (Messung 1796, Marke 2700); eine Marke, die ein Gate nach dem Code nachzieht, wäre nach `AGENTS.md` §3.14 verboten. Folge: keine, die Messung liegt weit unter beiden Zahlen. Adresse Planner: bei der Closure vermerken, dass die Schätzung die zwei Zeilen nicht trägt, Marke und Block nicht ändern.

**V-2 (LOW) Die Rate unter Last und der Lauf `DIAG_N` stammen von vor der Fix-Runde.** Der Absatz *Rate je Prozesslauf* (0 von 24 unter Last) und *Natürlicher Ablauf der Diagnose* (0 von 2400) messen den Stand von `5d76a7b1`. Die Fix-Runde `3ea3b9fa` änderte `zugriff.go` (Feld `conn` vom Typ `querier`), `export_test.go` und Tests; die Fix-Runde sagt selbst „ohne `-race`, Load nicht gemessen“. Der Closure-Trigger (§5) verlangt `TestTransaktionVertrag` unter der Last der Diagnose in keinem gemessenen Prozesslauf rot. Meine 12 von 12 grünen Läufe auf HEAD (siehe oben) decken das für den Stand nach der Fix-Runde, in kleinerer Zahl. Folge: der Plan benennt die Messung weiter als die von `5d76a7b1`; die Closure-Notiz nennt meine Zahl daneben. Kein Code-Mangel.

**V-3 (INFO) Der grüne `make gates` ist ein Cache-Treffer.** Der Lauf dauerte 34 s; jede Docker-Stufe (`build`, `test`, `race`, ...) meldete `CACHED`. Ein Cache-Treffer belegt, dass dieselbe Eingabe schon einmal grün lief (die Schicht gibt es nur nach Erfolg), aber er fährt den Test nicht neu. Für `race` ist das für diesen Slice unerheblich, weil ich das Paket mit `-race` selbst gefahren habe (42 s PASS plus die 12 Lastläufe); er gilt als Beleg, dass lint, a-check, d-check, `test`, `race` und `build` auf HEAD grün sind. Ein Lauf ohne Cache ist nicht nötig.

**V-4 (INFO) Offene Punkte der Closure, nicht des Codes.** §6 trägt sechs Risiken mit „offen bis zur Closure“ (Fenster zwischen Prüfung und Statement, Lese-Statements, Variante B, Laufzeit, Last/`race`, Zeile 4 Negativ und die Randform-Zeilen); nach meiner Prüfung sind alle sachlich entschieden und gefangen (Fänger Zeilen 7 bis 9 für das Fenster, `weitere` für Variante B, Laufzeit 1,9 s hält die Grenze von 2,5 s laut Plan; die Laufzeit habe ich nicht einzeln neu gemessen). Der Ausgang wird bei der Closure gesetzt, nicht vom Verifier. Das Register `state.md` trägt „Adresse Planner (Closure)“, keinen Zähler (richtig, `AGENTS.md`-Vorgabe).

## Gate-Lauf

`make gates` auf HEAD ist grün und nötig nur einmal (Closure). Ein weiterer Lauf vor dem Merge ist erst nötig, wenn nach HEAD noch Code, Test oder Spezifikation in diesem Zweig geändert wird (Plan-Nachzug der Closure ändert keine Stufe außer `d-check`).

## Verdikt je DoD-Punkt (Häkchen setzt der Planner)

| DoD | Verdikt |
|---|---|
| Abbruch beendet die Transaktion nicht mehr, Regressionstest mit erzwungener Reihenfolge | erfüllt (Zeilen 5 und 6, `TestTransaktionAbbruchImStatement`, Haltepunkte 7 bis 9; Mutation „ohne `WithoutCancel`“ nach Plan 20 von 20 und 24 von 24, mein Lauf der unveränderten Fänger grün) |
| Jede Zusage hat Fänger, Mutation rot gesehen, Raten mit Zahl, Last, Messendem; `TestTransaktionVertrag` unter Last grün | erfüllt; Raten des Plans vor der Fix-Runde (V-2), meine 12 von 12 danach |
| Diagnose-Tests umgeformt, `DIAG_N` und Messung nicht im Gate | erfüllt (kein „diag“ in `src`, `transaktion_abbruch_test.go` ohne Umgebungsvariable) |
| `make gates` grün | erfüllt (V-3) |
| Review durchgeführt, Report liegt vor | erfüllt (`cd53a653`: 0 HIGH, 2 MEDIUM, 1 LOW; alle drei in `3ea3b9fa` behoben und von mir gegen die Klasse aufgelöst) |
| Doku-Update (Vertrag in `SPEC-108` Nr. 1) | erfüllt; Binary und Katalog unverändert |
| Closure-Notiz mit Lerneintrag | offen (§7 „Wird bei der Closure gefüllt“) |
| Beobachtungs-Register fortgeschrieben | erfüllt (`state.md`, `evidence/slice-sqlite-interrupt-transaktion.md`, kein Zähler); Zielort der Regel nach der Closure |
| Jedes Risiko aus §6 trägt einen Ausgang | offen bis zur Closure (V-4) |
| Drei Paarungen | zur Closure |

## Aufräumen

Der Container der Messungen (`ver-sqlite`) ist entfernt, es läuft kein Prozess des Verifiers; kein `pkill`, kein Push. Die Busy-Loops liefen im Container und endeten mit ihm.

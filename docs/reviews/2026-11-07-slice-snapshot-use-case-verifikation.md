# Verifikationsbericht: slice-snapshot-use-case — 2026-11-07

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan, DoD, Vertrag und Entscheidungen; nicht der Diff (Reviewer) und nicht der Bedarf (Validator).

**Gegenstand:** Slice `slice-snapshot-use-case`, Zweig `slice-snapshot-use-case`, HEAD `46248426` (Implementierung `833824cd`, Review `1134888e`, Fix-Runde `46248426`). Plan `docs/plan/planning/in-progress/slice-snapshot-use-case.md`, Vertrag `SPEC-129` Nr. 16 bis 26, Review-Report `docs/reviews/2026-11-07-slice-snapshot-use-case-review.md`. Modell: Sonnet 5.5.

**Methode.** Einmal gefahren, keine Wiederholungsschleife. Die Behauptungen des Implementers sind nicht übernommen, die Belege selbst gemessen. Docker nur unter `flock /tmp/odata-recorder-heavy.lock` mit `--network none --memory 6g`; Mutationen nur in Baumkopien (`git archive HEAD` in das Scratch-Verzeichnis), nie im Worktree. Der Worktree trägt von mir nur diese Datei.

---

## 1. Sensoren, selbst gefahren

| Sensor | Ergebnis |
|---|---|
| `make gates` (ein Lauf, HEAD `46248426`) | Exit 0 im ersten Lauf, keine Wiederholung nötig. `TestTransaktionVertrag` ist nicht rot geworden. Ausgabe: `docs-check` 0 Befunde (848 Dateien, mit dem Review-Report), `baseline-verify` OK, `groessenmarke slice-snapshot-use-case.md: ok, Messung 5478 von Marke 10600`, `zeichen` ok. |
| Größenmarke, eigene Messung | `git diff --numstat 6e1e52d3..HEAD` über die Pfade des Blocks `groessenmarke`: 5451 Einfügungen plus 27 Löschungen = 5478; Marke 10600, also 52 %. Der Gate-Wert gleicht der eigenen Zählung. `src/bootstrap/workspace_sqlite_test.go` liegt außerhalb der Pfade (Rückgabe R4 e, im Plan genannt). |
| Katalog | `model.AlleCodes()` zählt 39 (`TestCodeTabelle` führt 39 Codes, selbst gezählt in `fehler_test.go`). `src/hexagon/model/`, `harness/katalog-ausnahmen.txt` und `tests/` sind im Diff `6e1e52d3..HEAD` unberührt: kein neuer Code. `make katalog-check` lief in `make gates` grün. |
| Binary unverändert | Im Diff liegt unter `src/bootstrap/` nur `snapshot_sqlite_test.go` (neu) und `workspace_sqlite_test.go` (ein Feld, Test). Kein Produktcode des Composition Root. Die DoD-Zeile zum Handbuch („Binary unverändert“) ist damit zu füllen; §7 trägt sie noch nicht (Closure-Sache des Planners). |
| Symbole der Tabelle | Alle 81 in Backticks genannten `Test…`-Symbole der Spalte *Fänger* (T01 bis T81) existieren als `func <Symbol>(` unter `src/` (81 von 81, Skript über `grep -rn`); ebenso alle Symbole im Absatz *Mutationen und Raten*. |
| Zeilen der Tabelle | 81 Zeilen T01 bis T81 (`grep -c '^| T[0-9][0-9] '` auf dem Tabellenbereich), vier akzeptierte Negative (T78 bis T81) mit Grund und Folge. |

## 2. Tabelle gegen den Vertrag (jeder Satz eine Zeile)

Ich habe `SPEC-129` Nr. 16 bis 26 Satz für Satz gegen die Konkordanz und die Zeilen T01 bis T81 gelegt.

- Nr. 16 (a) bis (i): T29 (a), T34 bis T36 und T37 (b bis d), T38/T39 (e), T40 (f), T45 (g), T30/T48/T61 (h), T32 (i). Entscheidung „Lesen vor der Transaktion“: T60.
- Nr. 17: T31, T32, T33, T47; „Workspace-Snapshots bleiben, kein anderer Workspace ändert sich“: T53 (Isolation) und T25 (der Port hat kein Löschen). Revision „Restore gewinnt immer“: T60.
- Nr. 18: Rangfolge T38, Codes T20, T34 bis T37, T49; Entscheidung `ODR-E5004` als Negativ T78; `ErrRevisionVeraltet` ist `ODR-E9001`: `TestRevisionVeraltetIstProgrammfehler` im Paket `workspace` (kein eigener Eintrag in der Tabelle, aber ein Test).
- Nr. 19 (a) bis (h): T60 bis T66, T70.
- Nr. 20: Form T07 bis T14, Lesen T41 bis T45, `state_hash` T16; Rundlauf T46.
- Nr. 21 bis 26: T01 bis T05, T15 bis T28, T67 bis T71, T74.

Befund zur Vollständigkeit: kein Satz ohne Zeile gefunden. Zwei Nebensätze tragen ihre Zeile nur mittelbar, ohne dass es eine Lücke wäre: „Die Workspace-Snapshots bleiben“ (Nr. 17, T53/T25) und „`ErrRevisionVeraltet` unerreichbar“ (Nr. 17, nur über den benannten Test, nicht über eine Zeile).

## 3. Fix der Review-Befunde gegen die Klasse (`AGENTS.md` §3.17)

Die Fix-Runde `46248426` ändert nur Tests, die Spezifikation und den Plan; Produktcode ist unverändert. Ich habe jeden Fix gegen die Menge aufgelöst, die die Zusage meint, mit Mutationen an Elementen, die der Befund nicht nannte. Alle Mutanten liefen in Kopien des Baums; die Spalte *Ergebnis* ist die Antwort von `go test` über das Paket.

### F-1 (Wert im Text eines Verstoßes) — Menge: Glied × Stelle des Verstoßes im Leser

Menge nach Wortlaut Nr. 20 *Lesen* („die Ursache nennt den Ort und keinen Wert“): jede Stelle in `manifest.go` und `manifest_snapshot.go`, die ein Glied oder einen Wert des Manifests lesen kann (Kopf 6 Glieder, Eintrag 7 Glieder, Elemente der Properties, Syntax, Doppelte, `Validate()`).

| Mutation (Wert in den Text) | Ergebnis |
|---|---|
| `completeness` (K1), `service_id` (K2), Text von `Validate()` (K3), `deleted` (Q9) | rot (`TestManifestEintrag`, `TestManifestKopf`, `TestManifestVerstoss`) — Kontrollen, Fix wirkt |
| Wert eines Gliedes falschen Typs in `zeichenkette` (Q1), Element der Properties (Q2), Rest bei leerem Glied (Q4), Wert in `objekt` (Q5), `entities` (Q6) | rot, aber durch den bestehenden `TestVerstoesseNennenOrtOhneWert` in `ort_test.go`, nicht durch den neuen Canary |
| Schlüssel im Text von `Validate()` (Q7b), Payload dort (Q7c), Schlüssel bei „doppelt“ (Q3b) | rot |
| **`entity_set` bei „doppelt“ (Q3)** | **grün** |
| **`entity_set` im Text der Invarianten (Q7)** | **grün** |
| **Properties bei „doppelt“ (Q3c)** | **grün** |
| **`version` bei „doppelt“ (Q3d)** | **grün** |

Ursache: `kanarienEintraege()` setzt den Canary nur in `typed_key` und `payload`; `entity_set`, die Properties und `version` des Eintrags tragen nie den Wert, und die Fälle für „doppelt“ und „Invarianten“ gehen nur durch diese Eintrags-Fixture. Die Fälle mit Canary in `completeness`, `deleted`, `version`, den Kopfgliedern und der doppelten Property sind richtig; die Klasse „jedes Glied an jeder Stelle“ ist es nicht.

### F-2 (Zählprüfung Quelltext) — Menge: Rückgaben mit Fehlerwert in den Dateien des Restore

`TestZweigeRestore` zählt jetzt per `go/parser` 16 Rückgaben mit Fehlerwert in `wiederherstellen.go`. Die Datei `manifest_snapshot.go` (11 Aufrufe von `verstoss`, 23 `return` mit Fehlerwert) gehört nach Auftrag D zur selben Z-Liste (Z26) und ist nicht gezählt. Mutation Q8 (eine zusätzliche Rückgabe mit `verstoss("x")` in `leseSnapshotEintrag`): **grün**. Die Regeln-Zählung in `TestManifestVerstoss` (16) zählt Regeln der Fixtures, nicht Stellen des Quelltexts.

### F-3 (Satz zu `ODR-E5004`)

Behoben: `spec/spezifikation.md` Abschnitt `LH-FA-10.a` trägt jetzt „Baseline eines anderen Service `ODR-E5004`, … auch der Workspace-Snapshot eines Workspace eines anderen Service `ODR-E7002` (`SPEC-129` Nr. 18)“. Die Menge der Stellen, die den Satz tragen, hat vier: der Abschnitt, Nr. 18 (Entscheidung), die Anwendung von `SPEC-109` Nr. 5 (nennt `ODR-E5004` aus dem Vergleich) und T78. Ich habe in `spec/spezifikation.md` nach `Snapshot eines anderen Service` gesucht: keine fünfte Stelle mit dem alten Wortlaut. Die Frage, ob das Lastenheft für Workspace-Snapshots `ODR-E5004` verlangt, bleibt offen (R3, beim User).

### F-4 (übersprungene Zelle) — Menge: Zellen Stelle × Fehler

`TestFehlerAbbildungJedeStelle` zählt `uebersprungen` getrennt und prüft `faelle + uebersprungen == 28*28` und `uebersprungen == 1`. Die Menge der übersprungenen Zellen ist die eine Zelle S19 × `ErrFalscheArt`; sie wird in `TestRestorePortFehler` geführt (T49). Kein Befund.

### F-5 (Kopie der Eingabe) — Menge: Liste der Einträge und jede Property-Liste

`TestManifestKopieEingabe` hält die Listen fest, die `entFake.List` ausgegeben hat, und vergleicht sie mit einer tiefen Kopie nach `Create`; die Fixture ist weder nach Menge und Schlüssel noch nach Properties geordnet. Eigene Mutanten an Elementen, die der Befund nicht nannte, liefen für die Folge der Einträge und der Properties in den Reviewer-Mutationen (N12, Y01) und in der Fix-Runde (F5a, F5b, laut Plan rot). Ich habe sie nicht erneut gefahren. Kein Befund.

### F-6 (T61 berichtigt)

Der Plan-Absatz zu T61 nennt jetzt die zweite, nachfolgende Transaktion als rot ohne Zeitlimit und die verschachtelte Form als anderen Fall. Stimmt mit dem Review überein.

## 4. Eigene Mutationen außerhalb der Fix-Klassen (Produktcode)

Zur Kontrolle, dass die Tabelle auch an Stellen trägt, die weder Plan noch Review nannten (jeweils ein Lauf des Pakets):

| Mutation | Ergebnis |
|---|---|
| `Restore`: Ergebnis ohne `CreatedAt` (Feld nicht kopiert) | rot (`TestRestoreRevision`) |
| `Restore`: Entity-State in umgekehrter Folge an `ReplaceAll` | rot (`TestRestoreWirkung`, `TestRestoreMaske`, `TestRestoreFensterLesen`, `TestPaketOhneZustand` wegen des Imports) |
| `Restore`: Ergebnis trägt die Revision vor `Advance` | rot (`TestRestoreRevision`, `TestRestoreFensterLesen`) |
| `Create`: Fehler von `Write` verschluckt | rot (`TestBlobAblauf`, `TestKennungZeit`, `TestPortFehlerSnapshot`) |
| `Create`: `Uhr.Now()` zweimal | rot (`TestKennungZeit`, `TestKonfigSnapshot`, `TestLesenSnapshot`) |
| `Create`: Fehler von `Abort` wirkt auf das Ergebnis | rot (`TestBlobAblauf`, `TestPortFehlerSnapshot`, `TestZweigeSnapshot`) |
| `List`: Bedingung `len(l)==0` statt `l==nil` | grün — äquivalent (eine leere, nicht-nil Liste wird durch eine leere ersetzt) |

## 5. Befunde

### V-1 (MEDIUM): Fix zu F-1 deckt die Klasse „Glied × Stelle“ nicht vollständig

- `quelle`: `AGENTS.md` §3.17 (Fix deckt die Klasse); `SPEC-129` Nr. 20 *Lesen* („Ort … keinen Wert“); Plan T45.
- `befund`: Vier Mutationen an Elementen der Menge, die der Befund nicht nannte, bleiben grün: `entity_set` im Text von „doppelt“ und der Invarianten, die Properties und `version` im Text von „doppelt“ (Abschnitt 3). Ursache: `kanarienEintraege()` trägt den Canary nur in `typed_key` und `payload`. Der Produktcode erfüllt die Zusage heute; eine Regression an diesen Stellen bliebe grün.
- `abhilfe`: Canary in `entity_set` und in die Properties der Fixture; für „doppelt“ und „Invarianten“ je Glied des Eintrags (7) eine Zeile, als Schleife über die Glieder; Zählprüfung der Stellen gegen die `verstoss`-Aufrufe der Quelle.
- `adresse`: Implementer (eine weitere Prüfrunde nur auf Anweisung des Planners; nach der Regel „Durchstich vor Härtung“ kein Blocker für den Übergang, sondern zur Entscheidung des Planners).

### V-2 (MEDIUM): Zählprüfung der Z-Liste nur für eine der beiden Quelldateien des Restore

- `quelle`: `AGENTS.md` §3.15, §3.17; Plan Auftrag D und T72 („Z-Liste, Zählprüfung gegen den Quelltext per `go/parser`“).
- `befund`: Der Fix zu F-2 zählt `wiederherstellen.go` (16). `manifest_snapshot.go` trägt Zweig Z26 (16 Regeln, 11 Aufrufe von `verstoss`) ohne Zählprüfung am Quelltext; eine zusätzliche Fehlerrückgabe (Mutation Q8) lässt kein Gate rot werden. Der Dienst `snapshot` zählt `dienst.go` (13), so dass die Klasse „jede Datei der Dienste“ zwei von drei Dateien deckt.
- `adresse`: Implementer; gleiche Entscheidung des Planners wie V-1.

### V-3 (MEDIUM): Sendung der Verdrahtung an `slice-bootstrap-m3` ohne Annahme im Nehmer (`AGENTS.md` §3.13)

- `quelle`: `AGENTS.md` §3.13; Plan T81 und Auftrag K („die Verdrahtung des `SnapshotUseCase` ist `slice-bootstrap-m3`, dessen §1 nennt `slice-snapshot-use-case` als Geber“).
- `befund`: Das §1 von `slice-bootstrap-m3` hat als Ziel „genau `MockUseCase`“ im Modus `mock` und nennt `slice-snapshot-use-case` nur im Ausschluss „Neue Fachlogik — der Slice verdrahtet, was … liefern“. Weder §1 noch DoD von `slice-bootstrap-m3` führen den Wächter der Wartung als `Blobs` des Snapshot-Dienstes, die gemeinsame Datenbank für Workspaces, Entities, Snapshots und Transaktion oder die Bindung des `SnapshotUseCase`; kein Eintrag trägt „Übernimmt (Sendung von `slice-snapshot-use-case`)“. `slice-admin-workspaces` (der Verbraucher des Use Case) schreibt ausdrücklich „Nicht gebraucht: … `slice-bootstrap-m3`“ und führt die Verdrahtung ebenfalls nicht. T81 steht als akzeptiertes Negativ „liefert `slice-bootstrap-m3`“; damit hat die Zusage („ein Schreiber des Blobs ohne Wächter ist ungeschützt“, `SPEC-123` Nr. 9) keinen Nehmer.
- `abhilfe`: Der Planner liest §1 *Ausdrücklich NICHT* und DoD von `slice-bootstrap-m3` und `slice-admin-workspaces` und trägt die Sendung dort mit der Kennung des Gebers ein, oder weist sie einem Slice zu, der den Use Case wirklich bindet; bis dahin ist R4 e/T81 keine Adresse.
- `adresse`: Planner.

### V-4 (INFO): Offene Rückgaben (§3.13)

R1: erledigt (Planner). R2: weiter offen, Adresse Planner, Kennungsfrage beim User. R3: offen, Adresse Planner (Meldung an den User), in `spec/spezifikation.md` und im Plan als Lockerung geführt. R4: offen, Adresse Architect (a bis c, e) und Reviewer (d); das Review hat (a), (b), (d), (e) aufgelöst und (c) beim Architect gelassen. Jede Rückgabe trägt Ausgang und Adresse im Plan §6; R2 und R3 stehen beim User, wie vorgegeben, und blockieren diesen Slice nicht.

### V-5 (INFO): Raten nicht nachgemessen

Die Raten des Implementers (50 von 50, 10 von 10) habe ich nicht nachgemessen; `make gates` (`test-race` eingeschlossen) lief einmal grün durch. Von den Fenstern habe ich nur die Fänger gelesen und die Mutationen R2/R3 (Folge, Revision) gefahren, die `TestRestoreFensterLesen` rot machen. Der Flake `TestTransaktionVertrag` trat nicht auf.

### V-6 (INFO): DoD-Kästchen

Alle Kästchen der DoD stehen auf `[ ]`; §7 trägt noch die Platzhalter-Zeile. Das ist der Stand vor der Closure und liegt beim Planner; die Liefer-Punkte 1 bis 3 sind durch die Tabelle gedeckt (Abschnitt 2), `make gates` ist grün, der Review-Report liegt vor, und die Zeile zum Handbuch ist erfüllbar (Binary unverändert, Abschnitt 1).

## 6. DoD-Konformität im Einzelnen

| DoD-Punkt | Stand |
|---|---|
| Erzeugen und Auflisten (E7006, nur Entity-State, leerer Workspace) | belegt: T03, T07, T13, T14, T19, T26, Mutationen rot |
| Restore atomar, Revision, Cursor, E7002/E5004/E7008 ohne Teilzustand | belegt: T31 bis T37, T48, T54 bis T57; T78 als benanntes Negativ |
| Serialisierung mit Wirkungs-Fänger | belegt: T60 bis T66 (Haltepunkt am Fake, Wirkungs-Fänger gegen den echten Adapter); Raten nicht nachgemessen (V-5) |
| `make gates` grün | selbst gefahren, ein Lauf, grün |
| Review | Report liegt vor, 3 MEDIUM, 2 LOW, 4 INFO; die Fix-Runde setzt F-1 (teilweise), F-2 (teilweise), F-3 bis F-6 um |
| Doku-Update | Spezifikation und Plan nachgezogen (`LH-FA-10.a`, Nr. 16 bis 26) |
| Handbuch/Katalog | Binary unverändert, kein neuer Code, `make katalog-check` grün; §7 trägt die Zeile noch nicht (Closure) |
| Closure, Register, Risiken, Paarungen | noch nicht gefüllt (Closure des Planners) |

## 7. Urteil

**erfüllt, mit Vorbehalt.** Die Lieferung erfüllt die DoD-Punkte 1 bis 3 und `make gates` im ersten Lauf, bleibt mit 5478 Zeilen unter der Marke 10600 und führt keinen neuen Code ein; die Tabelle T01 bis T81 hat zu jedem Satz von `SPEC-129` Nr. 16 bis 26 eine Zeile, und jedes Symbol existiert. Die Fix-Runde löst F-3 bis F-6 vollständig, F-1 und F-2 aber nur für die genannten Elemente und nicht für die Klasse (V-1, V-2): vier Mutationen an Elementen der Menge „Glied × Stelle“ und eine zusätzliche Fehlerrückgabe in `manifest_snapshot.go` bleiben grün. Ein Produktfehler liegt nicht vor. V-3 (Sendung ohne Annahme) liegt beim Planner. Ob V-1 bis V-3 vor dem Übergang nach `done/` eine Runde brauchen, entscheidet der Planner; ich habe keine Wiederholungsschleife gefahren.

## 8. Ungeprüft

Raten T60 bis T71 (nicht nachgemessen); die Fänger T65, T66, T68, T70, T71 gegen den echten Adapter nicht mutiert; die Verdrahtung im Composition Root (T81, V-3).

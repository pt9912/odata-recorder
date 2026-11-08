# Review-Report: slice-mock-abnahme — 2026-11-08

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-mock-abnahme`, HEAD `0daee175`, Basis `7d6e18f9` für den Code (Commits `d63b4755`, `429d4da4`, `1b81a474`, `72b64c72` des Implementers, `1a595a32` des Architect, `0daee175` zur Rückgabe R5); die sechs Dateien `src/bootstrap/abnahme_*_test.go` (2325 Zeilen, kein Produktcode), Plan §3 (Tabelle T01 bis T83, Absatz *Mutationen und Raten*, Mengen M1 bis M8), §6 (Raster, Zusatztabelle Z1 bis Z12, Rückgaben R1 bis R10) und `SPEC-049` Nr. 1 bis 10.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08

**Eingangs-Kontext:**

- Plan von `slice-mock-abnahme` (§3 bis §6), `SPEC-049` Nr. 1 bis 10, daraus die Verweise `SPEC-193` Nr. 2 und 3, `SPEC-192` Nr. 6 bis 8, `SPEC-129` Nr. 10, 16 bis 23
- `AGENTS.md` §3.2, §3.7, §3.9 bis §3.18
- `LH-QA-02`, `LH-QA-03`, `LH-RB-01`
- ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (nur genannt)
- vorherige Reports am Modul: `2026-11-07-slice-telemetry-spans-review.md` (Form, Raten-Messung)

**Nicht geprüft (Rolle des Verifiers oder Planners):** DoD, `make gates`, Größenauslastung (2325 von 9700) und Laufzeit des Pakets `src/bootstrap` (Rückgabe R10, Planner). Die Laufzeit ist bewertet, nicht als Befund doppelt.

## Eigene Mutationen

Messender ist der Reviewer-Agent. Container aus dem Image `odata-recorder:race`, ohne Netz, `--memory 3g`, ein Container zur Zeit, jeder Lauf unter `flock /tmp/odata-recorder-heavy.lock`; der Baum liegt als Wegwerfkopie im Container (`/src`, Original in `/pristine`), je Mutation wird ein Text genau einmal ersetzt (`perl -0pi`) und danach die Kopie zurückgesetzt. Ratenläufe: das Test-Binary des Pakets einmal übersetzt, dann `-test.run <Test> -test.count=1` in 20 Prozessen nacheinander; Last ist der Eins-Minuten-Wert aus `/proc/loadavg` zu Beginn und am Ende der Reihe. Ausgangsstand: `go test -run TestAbnahme ./src/bootstrap/` grün, 25,6 s (Last 4).

Gewählt sind Elemente der Mengen, die der Implementer nicht nannte (M1 Fixture-IDs, M2 Schritte, M5 Workspaces, M7 Fenster, M8 Läufe), und die Klasse der Zusage statt der genannten Mutation (`AGENTS.md` §3.17).

| Nr. | Zusage / Menge | Mutation (Wegwerfkopie) | Ergebnis |
|---|---|---|---|
| A1 | T13, M2 (`ODV4-Q-10`, zweiter Schlüssel) | `ordne` wertet nur das erste Element von `$orderby` aus | **grün** (`TestAbnahmeMatrix`, 6,2 s) |
| A2 | T25, M2 (`OData-Version` an jeder Antwortart, nicht genannt: `201`) | der Mock setzt das Kennzeichen an `201` nicht | rot `ODV4-H-02`, `ODV4-CRUD-01` |
| A3 | T57, M5 | `Cursors.List` ohne Filter auf den Workspace | rot `TestAbnahmeIsolation` |
| A4 | T55, M5 | `Entities.Get` ohne Filter auf den Workspace | rot `TestAbnahmeIsolation` |
| A5 | T48 / T57 | `ResetAll` setzt `position` auf 0, lässt `consumed` stehen | rot `TestAbnahmeDeterminismus` (Stand nach Reset, 100 von 100 abweichend); **`TestAbnahmeIsolation` grün** (siehe F5) |
| B1 | `SPEC-193` Nr. 2, `PUT` an `partial` | `PUT` an einem `partial`-Eintrag ergibt `ODR-E6003` | **grün** (`TestAbnahme*`) |
| B2 | `SPEC-193` Nr. 2, `POST` über Tombstone | `POST` über einen Tombstone ergibt `ODR-E6009` | **grün** (`TestAbnahme*`) |
| B3 | T52 (Frist von 120 s) | Wächter von `abDetLaufen` lässt Fristen bis 150 s zu | **grün** (`TestAbnahmeDeterminismus*`) |
| T56 | T56, M5 | `Advance` ohne Kennung in der Bedingung (`WHERE ? <> ''`) | rot in **20 von 20** (ohne Detektor), Last 3,26 / 3,96 |
| T57 | T57, M5 | `ResetAll` setzt alle Cursors zurück | rot in **20 von 20** (ohne Detektor), Last 8,71 / 4,80 |
| T59x | T59, M5 (Variante) | `ReplaceAll` löscht ohne Filter auf den Workspace | rot in **20 von 20** (ohne Detektor), Last 3,72 / 3,60 |
| B4 | T73, M7 (Fenster; `PATCH` liest `Konfig.Entities` vor der Transaktion, nicht genannt) | `schreibe` ruft für `PATCH` vor `transaktion` `Entities.Get` | rot in **20 von 20** mit Detektor (Last 5,96 / 4,44) und **20 von 20** ohne (Last 4,22 / 3,69); rot über die Wirkung (Zähler am Haltepunkt der Entities) |
| T75 | T75, M7 (Nachmessung der Zeile des Implementers) | `Advance` mit `mu.w.Revision` statt `w.Revision` | rot in **20 von 20** ohne Detektor (Last 59,6 / 55,2) und **20 von 20** mit Detektor (Last 62,9 / 58,0), der Host war dabei stark belastet |
| T76a | T76 | Haltepunkt gibt bei *n*−1 frei | rot in **20 von 20** (ohne Detektor), Last 4,80 / 4,80 |
| T76c | T76 (nicht genannt: Haltepunkt ohne *n* hält nicht) | `halte` kehrt bei *N* gleich 0 sofort zurück | rot in **20 von 20** (ohne Detektor), Last 6,33 / 2,12; rot nach 10,1 s über `Warte` und dann `Fatal`, nicht über die 50 ms |

Ein erster Ratenlauf zu T75 war ungültig (die Mutation übersetzte nicht, das Binary der vorigen Mutation lief weiter) und ist verworfen; die Zahlen oben stammen vom wiederholten Lauf mit Übersetzungsprüfung.

**Abdeckungsprofil (T78):** `go test -race -coverprofile -coverpkg` über `services/mock`, `services/workspace`, `services/snapshot` mit `-run TestAbnahme`: 71,7 Prozent der Anweisungen; in `mock/schreiben.go` 116 Blöcke (26 ohne Treffer), `workspace/reset.go` 19 (8), `workspace/wiederherstellen.go` 37 (19), `snapshot/dienst.go` 41 (18): zusammen 213, 142 mit Treffer, 71 ohne — wie im Plan.

## Findings

### F1 · MEDIUM

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (eine Zusage ohne Fänger; ein Fänger führt jede Form der Eingabe, die der Code verschieden behandelt; eine Fixture gleicht nie dem Wert der Mutation)
- `pfad`: `src/bootstrap/abnahme_matrix_test.go` · `{ID: "ODV4-Q-10", Schritte: []abSchritt{abS("", abG("/Products", "$select=Id,Name&$orderby=Name desc,Id")`; Plan §3 Zeile `| T13 | `ODV4-Q-10`: `$orderby` mit `desc`, zweiter Schlüssel, `null` am Ende |`
- `befund`: Die Zusage nennt den zweiten Schlüssel, die Mutationen der Zeile T13 und des Absatzes (desc aufsteigend, `null` zuerst) betreffen ihn nicht. In den Ausgangsdaten sind die Namen der fünf Produkte paarweise verschieden (`null` kommt einmal vor), und der zweite Schlüssel `Id` ist zugleich der implizite Schlüssel der Ordnung; die Mutation A1 (`ordne` wertet von `$orderby` nur das erste Element aus) lässt `TestAbnahmeMatrix` und die übrigen `TestAbnahme*` grün. Eine Fixture mit Gleichstand im ersten Schlüssel und anderer Richtung im zweiten fehlt.
- `verifizierbar`: ja — Mutation A1 in einer Baumkopie, `go test -run 'TestAbnahmeMatrix$' ./src/bootstrap/` grün
- `klasse`: Zusage nennt ein Element, das die Fixture nicht trennt (Gleichstand fehlt)

### F2 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15 (beide Seiten jeder Grenze)
- `pfad`: `src/bootstrap/abnahme_determinismus_test.go` · `if dl, ok := ctx.Deadline(); !ok || time.Until(dl) > abDetFrist {`; Plan §3 Zeile `| T52 |`
- `befund`: `TestAbnahmeDeterminismusFrist` prüft eine Frist ohne, eine von 180 s, eine abgelaufene und eine im Lauf ablaufende; die Obergrenze von 120 s hat keinen Fall dicht darüber. Der Wächter lässt in der Mutation B3 bis 150 s zu, ohne dass ein Test rot wird. Die Konstante `abDetFrist` selbst ist durch den Vergleich mit 120 s gehalten, der Wächter nicht.
- `verifizierbar`: ja — Mutation B3
- `klasse`: Grenze nur von einer Seite geprüft

### F3 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.9 (der Plan folgt jeder Korrektur), §3.15
- `pfad`: Plan §3 Zeile `| T52 | Nr. 5 f: die Frist von 120 s ist gesetzt, ihr Ablauf ist `Fatal` |` gegen die Zeile `| T52 | (a) die Läufe prüfen die Frist nicht; (b) der Test setzt keine Frist; (c) der Ablauf der Frist ist keine Fatal |`
- `befund`: Die Spalte *Fänger* der Tabelle nennt allein `TestAbnahmeDeterminismusFrist`; die Mutation (b) *der Test setzt keine Frist* wird nach dem Absatz *Mutationen und Raten* von `TestAbnahmeDeterminismus` gefangen. Die Tabelle und der Absatz nennen für dieselbe Zeile verschiedene Fänger.
- `verifizierbar`: nein — Lesen des Plans
- `klasse`: Fänger-Spalte weicht von der Mutationstabelle ab

### F4 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.10 (Rate je Mutation mit Zahl der Prozesse, Last und Messendem), Skill *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*
- `pfad`: Plan §3 Zeilen `| T54 |` bis `| T60 |` (Klasse *nebenläufig*) und `| T77 | Nr. 8: Raten je Mutation von T71 bis T75 sind 20 von 20 |`
- `befund`: Die Zeilen T54 bis T60 tragen die Klasse *nebenläufig*, der Absatz *Raten der Fenster* nennt Raten allein für T62 und T71 bis T76; für T55 bis T60 steht keine Rate. Gemessen sind sie nicht flatternd: T56, T57 und die Variante von T59 sind in je 20 von 20 Prozessen ohne Detektor rot (Last 3 bis 9). Der Plan kennt die Zahl nur nicht.
- `verifizierbar`: ja — Ratenläufe T56, T57, T59x
- `klasse`: Rate einer nebenläufig klassierten Zeile fehlt im Plan

### F5 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.15 (Orakel), `SPEC-049` Nr. 6 c und e
- `pfad`: `src/bootstrap/abnahme_isolation_test.go` · `if a, o := ergebnis[k][i], orakel[k][i]; a != o {`; `src/bootstrap/abnahme_neustart_test.go` · `if l := abListe(t, u, "ws-1"); !reflect.DeepEqual(l, liste1) {`
- `befund`: Das Orakel der Isolation sind Einzelläufe desselben Codes, das des Neustarts die Werte vor dem Schließen; beide prüfen relativ. Ein Fehler, der im Einzellauf wie im Paralleltest gleich wirkt, färbt die Isolation nicht (A5: `ResetAll` lässt `consumed` stehen, Isolation grün, `TestAbnahmeDeterminismus` rot über T48). Das ist der Zuschnitt der Zusage (Isolation, nicht Korrektheit des Reset) und durch T47 und T48 gedeckt; die Nr. 6 e nennt „Position 0 mit leerem consumed“ als Beobachtung, der Test prüft sie nur im Vergleich mit dem Orakel. Kein Fänger fehlt; Zuständig: keine Aktion.
- `verifizierbar`: ja — Mutation A5
- `klasse`: Orakel relativ, absolute Erwartung durch andere Zeile gedeckt

### F6 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.15 (akzeptiertes Negativ mit Grund und Folge), `SPEC-049` Nr. 10 h
- `pfad`: `spec/spezifikation.md` · `(h) Kombinationen: jede Fixture deckt eine Facette der Zeile`
- `befund`: Die Matrix deckt von den 20 Zellen *Eintrag × Methode* in `SPEC-193` Nr. 2 neun. Die Mutationen B1 (`PUT` an `partial` ergibt `ODR-E6003`) und B2 (`POST` über einen Tombstone ergibt `ODR-E6009`) bleiben grün. Nr. 10 h nennt das Kreuzprodukt „jede Methode gegen jeden Typ“, nicht gegen den Zustand des Eintrags; der Grund und die Folge („fiele erst im Test ihres Vertrags auf“) tragen, die Zellen stehen in den Tests des Dienstes. Der Wortlaut nennt die Achse Zustand des Eintrags nicht.
- `verifizierbar`: ja — Mutationen B1 und B2
- `klasse`: akzeptiertes Negativ nennt eine Achse des Kreuzprodukts nicht

### F7 · INFO

- `kategorie`: INFO
- `quelle`: `SPEC-049` Nr. 3 c (Menge der Header-Namen genau gleich)
- `pfad`: `src/bootstrap/abnahme_aufbau_test.go` · `n := http.CanonicalHeaderKey(k)`; `src/bootstrap/abnahme_matrix_test.go` · `func abKopfGleich`
- `befund`: `abSende` und `abKopfGleich` kanonisieren die Namen auf beiden Seiten; die Schreibweise der Namen in der `Antwort` des Mock (etwa `Odata-Version` statt `OData-Version`) ist nicht Teil des Vergleichs. Der HTTP-Adapter kanonisiert ohnehin (Nr. 10 a nimmt ihn aus); die Folge ist klein, der Wortlaut „Namen genau“ trägt es nicht ausdrücklich.
- `verifizierbar`: nein
- `klasse`: Vergleich enger als der Wortlaut (Schreibweise der Namen)

### F8 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.10 (Zweige unter dem Detektor aus dem Profil), `SPEC-049` Nr. 8
- `pfad`: Plan §3 Absatz *Abdeckungsprofil*: `Die 71 sind Fehlerzweige und kein Zweig eines Fensters`
- `befund`: Die Zählung des Plans ist nachgemessen und stimmt (213, 142, 71). Die Aufzählung der Kategorien nennt den Zweig `typ != mu.eingabe.Typ` mit `ODR-E6007` (`SPEC-193` Nr. 2 a, `schreiben.go` im Block ohne Treffer) nicht; er gehört weder zu den Fehlern eines Ports noch zur Form des Payloads. Er ist kein Fenster, das Negativ trägt also, die Aufzählung ist nur unvollständig.
- `verifizierbar`: ja — Profil
- `klasse`: Aufzählung eines Negativs unvollständig

### F9 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.10, `SPEC-049` Nr. 8 (Rückgabe R9)
- `pfad`: Plan §3 Absatz `Ein Zugriff über `Konfig.Entities` **innerhalb** der Funktion der Transaktion`
- `befund`: Die Fenster T71 bis T75 sind über die Wirkung gefangen (Zähler am Haltepunkt der Entities, Zahl der `201`, `204`, `ODR-E9001`), nicht über die Frist. Bestätigt: B4 (nicht genannt) und T75 (nachgemessen) sind in 20 von 20 Prozessen mit und ohne Detektor rot, T75 auch bei Last 56 bis 63. Der Zugriff in der Funktion der Transaktion ist der eine Fall, der über die Frist fällt (Meldung „Haltepunkt nicht erreicht“, bis zu 10 s je Lauf); Plan und Spezifikation nennen ihn als zulässig mit der Dauer als Folge. Die Meldung nennt als Ursache den Haltepunkt, nicht die Sperre; wer sie liest, muss R9 kennen.
- `verifizierbar`: ja — Mutationen B4, T75
- `klasse`: Fänger über die Frist, Ursache in der Meldung nicht genannt

### F10 · INFO

- `kategorie`: INFO
- `quelle`: Maintainability (Rückgabe R10, Zuständigkeit Planner)
- `pfad`: `src/bootstrap/abnahme_determinismus_test.go` · `func TestAbnahmeDeterminismusFrist`
- `befund`: Bewertung der Laufzeit, kein Befund: die Abnahme-Tests zusammen brauchen 25,6 s ohne Detektor (Last 4), die 100 Läufe 3,5 s; das ist weit unter der Frist von 120 s. Der Fall *Frist im Lauf abgelaufen* setzt 400 ms voraus und erwartet, dass 100 Läufe länger brauchen; das gilt bei 35 ms je Lauf mit Abstand, ein Host, der einen Lauf in unter 4 ms schafft, machte den Test rot ohne Fehler im Mock.
- `verifizierbar`: nein
- `klasse`: Fänger mit Zeitannahme über die Geschwindigkeit des Hosts

## Antworten auf die Schwerpunkte

- **§3.10, Nebenläufigkeit:** Raten 20 von 20 mit Zahl der Prozesse (20 je Arm), Last (Paar je Reihe) und Messendem stehen im Plan für T62 und T71 bis T76. Nachgemessen T75 (beide Arme), eigene Mutation B4 (beide Arme), T76 (a) und (c): alle 20 von 20. Das Fenster ist über die Wirkung gefangen; allein der Zugriff in der Funktion der Transaktion fällt über die Frist (R9, zulässig, F9).
- **§3.11, Kommentare:** Stichprobe aller Kopf- und Funktionskommentare der sechs Dateien; jede Zusage hat einen Test. Auffälligkeit nur in F7 (Wortlaut „Namen genau“ im Vergleich enger).
- **§3.2:** keine Direktive `//nolint` in den Dateien (`grep`), keine Ausnahme in `.golangci.yml` berührt.
- **§3.18:** `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}'` über die sechs Dateien ohne Treffer; Zeichen jenseits von ASCII sind Umlaute und `…` im Kommentar, kein BOM, kein NBSP, keine Escape-Folge als Fixture.
- **Fixtures ungleich dem Wert der Mutation:** stichprobenweise gegen die Mutationen der Tabelle geprüft (Q-03 bis Q-06 mit Gleichheit an genau einem Produkt, Q-11 mit `$top` und `$skip` verschieden, T-06 nach R5 mit Hex-Buchstaben, CRUD-01 d mit Zahl gegen Text); eine Ausnahme ist der Gleichstand in F1.
- **Das Orakel (Einzelläufe) unterscheidet die Workspaces (T61):** ja, in den Schritten 2 und 5 für alle 190 Paare (R6); der Wortlaut ist berichtigt. Relative Prüfung: F5.
- **Akzeptierte Negative:** T79 bis T83 und die 71 Blöcke tragen Grund und Folge. F6 und F8 betreffen den Wortlaut, nicht die Tragfähigkeit.
- **Abweichung T52 (c) / T21:** T21 ist nach R5 an `ODV4-T-06` und `ODV4-CRUD-01` rot (Plan; die Guid von Produkt 3 trägt Hex-Buchstaben, Spezifikation, Aufbau und Filter stimmen überein). T52: F3.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/abnahme_aufbau_test.go` (Aufbau, Ausgangsdaten, `abParallel`) | geprüft, ohne Befund (die Goroutinen werden nacheinander am Tor gesammelt, das Signal fällt nach dem letzten; `FestUhr` und `FestIDs` sind gleichzeitig sicher) |
| `src/bootstrap/abnahme_matrix_test.go` (36 IDs, 88 Schritte, Zählprüfung) | geprüft, Befund F1 und F7; Zählung gegen eine aus Bereichen erzeugte Menge, Gegenproben des Zählers im Test |
| `src/bootstrap/abnahme_determinismus_test.go` (100 Läufe, Stand, Cursor, Vergleich) | geprüft, Befund F2 und F10; die vier Namen der Antwort und `VOLATILE_HEADERS` aus der Konfiguration, keine Kopie |
| `src/bootstrap/abnahme_isolation_test.go` (20 Workspaces, 228 Schritte, 380 Paare) | geprüft, Befund F5; Mutationen A3, A4, T56, T57, T59x rot |
| `src/bootstrap/abnahme_neustart_test.go` (Schritte 1 bis 9, Zusagen (1) bis (6), `TestAbnahmeKeinSkip`) | geprüft, Befund F5 (relatives Orakel der Snapshots) |
| `src/bootstrap/abnahme_nebenlaeufig_test.go` (Haltepunkt, fünf Fenster) | geprüft, ohne Befund am Code; Raten bestätigt, F9 |
| `spec/spezifikation.md` `SPEC-049` Nr. 1 bis 10 gegen Plan und Tests (R5 bis R9) | geprüft, ohne Befund (Aufbau, Haltepunkt und Cursors des Neustarts stimmen mit dem Code überein) |
| Plan §6 Raster und Zusatztabelle Z1 bis Z12 gegen die Fixtures `ODV4-E-04` und `ODV4-CRUD-06` | geprüft, ohne Befund (jede Zeile hat einen Fänger) |
| `.golangci.yml`, `.a-check.yml`, `Makefile`, `Dockerfile`, `harness/` | unverändert (Diff seit `7d6e18f9`: nur `spec/spezifikation.md`, der Plan und die sechs Dateien) |

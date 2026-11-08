# Verifikation: slice-mock-abnahme — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-049` Nr. 1 bis 10, die Tabelle *Zusage · Fänger* und die Hard Rules; nicht Diff-Review (siehe `2026-11-08-slice-mock-abnahme-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-mock-abnahme`, HEAD `73a1f341` (Fix-Runde `267bdc89`, Messung `028ba949`, Entscheidung R12 `73a1f341`), Worktree `/Development/odata-recorder-wt-mock-abnahme`. Plan: `docs/plan/planning/in-progress/slice-mock-abnahme.md`. Bezug: `LH-QA-02`, `LH-QA-03`, `LH-RB-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden, mit Auflagen an den Planner (Plan-Nachzug und Closure); keine Fix-Runde am Code.** `make gates` ist im eigenen Lauf grün, der Diff trägt genau die sechs Testdateien (2335 Zeilen, kein Produktcode) plus Spezifikation und Plan, die Fix-Runde deckt die Klassen von F1 und F6 (eigene Mutationen an nicht genannten Elementen rot), F4 ist gemessen bestätigt. Drei Befunde: ein Plan-Widerspruch zu R12 (V1, MEDIUM, Plan), eine Lücke der Obergrenze des Frist-Wächters zwischen 120 und 125 s (V2, LOW) und die offenen Closure-Punkte (V3, INFO).

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/gates.log`: `EXIT=0`; `d-check: 855 Datei(en) geprüft, 0 Befund(e)`; Stufen `pruefung test` und `pruefung race` laufen; `zeichen: ok, 1737 Datei(en)`. |
| `make groessenmarke` | im selben Lauf: `slice-mock-abnahme.md: ok, Messung 2335 von Marke 9700`; `git diff --stat main...HEAD -- src`: 6 Dateien, 2335 Einfügungen, 0 Löschungen. Auslastung 0,24 (erwartet 0,56): Meldung an den Architect in der Closure, wie vom Aufrufer bestimmt. |
| Plan-vs-Code-Diff | `src/` ändert nur `src/bootstrap/abnahme_{aufbau,matrix,determinismus,isolation,neustart,nebenlaeufig}_test.go`, die sechs Pfade des Blocks (§3 und `groessenmarke`-Block); `produktcode: 0` stimmt; `Makefile`, `Dockerfile`, `harness/mk/`, `.golangci.yml`, `.a-check.yml` unberührt (Diff nennt sie nicht). `spec/spezifikation.md` nur `SPEC-049` und Historie. |
| Merge-Nähe | der Zweig steht 32 Commits hinter `main`; `git merge-tree --write-tree HEAD main`: Rückgabe 0, kein Konflikt. |
| §3.2 / §3.18 | `nolint` je Datei 0; `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}|\xEF\xBB\xBF|\xC2\xA0'` über die sechs Dateien und den Review-Report: 0 Treffer. Dieser Bericht ist nach dem Schreiben mit demselben Muster gegengelesen (siehe Schluss). |
| Zählung `SPEC-049` Nr. 3 e / Nr. 10 h | 36 IDs, 92 Schritte, Summe der Bereiche 7, 29, 16, 9, 13, 18 steht in `abPruefeZaehlung` (Zeile 324), Kopf (Zeile 22) und Fänger `TestAbnahmeMatrixZaehlung` (läuft in `make test`). *Elf von 20 Zellen* eigenhändig nachgezählt (Tabelle `SPEC-193` Nr. 2, Methoden der `abM`-Schritte): `POST` 3 (vorhanden `complete` `CRUD-06` a, Tombstone `CRUD-04` f, fehlend `complete` `CRUD-01`), `PATCH` 4 (`CRUD-02`, `CRUD-06` f, d, e), `PUT` 2 (`CRUD-03` a, c), `DELETE` 2 (`CRUD-04` a, e); 3 + 4 + 2 + 2 = 11. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Abnahmematrix: Muss-Zeilen als Golden-Fixtures, je ID ein Test | **erfüllt, Haken fehlt** | `TestAbnahmeMatrix` mit 36 Teilfällen `ODV4-*`; `make gates` grün; Mutationen siehe §4. |
| Determinismus: 100 Läufe nach Reset | **erfüllt, Haken fehlt** | `TestAbnahmeDeterminismus` (Frist 120 s gesetzt); Fänger T44 bis T53; Mutation V2 siehe unten. |
| Isolation und Neustart | **erfüllt, Haken fehlt** | `TestAbnahmeIsolation` (20 Workspaces), `TestAbnahmeNeustart`; Raten in §5. |
| `make gates` grün | erfüllt | eigener Lauf, `EXIT=0`. Nach einem Merge von `main` einmal wiederholen (32 Commits Abstand). |
| Review durchgeführt, Report liegt vor | erfüllt (Haken fehlt) | `docs/reviews/2026-11-08-slice-mock-abnahme-review.md`; F1 bis F4 in der Fix-Runde abgearbeitet, F6 und F8 entschieden (R11), R12 entschieden. |
| Doku-Update | erfüllt | öffentlicher Vertrag `SPEC-049` Nr. 1 bis 10 samt Historie; kein Produkt-Vertrag berührt. |
| Closure-Notiz, Register, Paarungen | **offen, Sache des Planners** | §7 ist `Wird bei der Closure gefüllt`; siehe V3. |
| Ausgang je Risiko (§6) | **offen** | R2 (User), R10 (Planner), `Zahl der Codes` (Planner) sind *offen* oder *weiter offen*; die Raten-, Laufzeit- und Größen-Zeilen tragen *weiter offen* und gehören mit der Messung der Closure auf einen Ausgang. |

Liefer-Punkte: drei, am Limit. Ein Haken ist noch nirgends gesetzt; das setzt der Planner mit der Closure.

## 3. Tabelle *Zusage · Fänger* gegen den Vertrag

Geprüft Zeile für Zeile gegen `SPEC-049` Nr. 3 bis 10 und die Tabelle von Nr. 4. Jede der Zeilen T01 bis T78 nennt einen Fänger, dessen Symbol im Quelltext steht (`grep -n "func Test\(Abnahme\)"` über die sechs Dateien: `TestAbnahmeMatrix`, `…Zaehlung`, `…Vergleich`, `…Ausgangszustand`, `TestAbnahmeKeinSkip`, `TestAbnahmeDeterminismus`, `…Vergleich`, `…Auswertung`, `…Frist`, `TestAbnahmeIsolation`, `…OrakelUnterscheidet`, `…Start`, `TestAbnahmeNeustart`, `TestAbnahmeNebenlaeufig{Post,Delete,Patch,Loeschen,Reset}`, `TestAbnahmeHaltepunkt`). Befund:

### V1 · MEDIUM — die Tabelle führt nach R12 einen Fänger, den es nicht gibt

- `quelle`: `AGENTS.md` §3.15 (jede Zusage hat einen Fänger oder ein akzeptiertes Negativ mit Grund und Folge), §3.9 (der Plan folgt jeder Korrektur; wer eine Aussage berichtigt, zieht jede tragende Stelle nach), `SPEC-049` Nr. 10 i.
- `pfad`: Plan §3, Zeile `| T33 |` (Spalte *Mutation*: „der Eintrag bleibt nach dem `PUT` `partial`“ mit Fänger `TestAbnahmeMatrix/ODV4-CRUD-03`); Zeile `| T32 bis T35 |` des Absatzes *Mutationen und Raten* (endet auf „…, Rückgabe R12“ ohne den Ausgang); Zeile `| T83 |` (nennt Nr. 10 e bis h, nicht i); Zeile `| Z13 |` der Zusatztabelle (Ergebnis „danach `complete`“, Fänger `ODV4-CRUD-03` c und d).
- `befund`: R12 ist in §6 *entschieden* (akzeptiertes Negativ, `SPEC-049` Nr. 10 i), und die Spezifikation sagt zu, dass `CRUD-03` d „der Eintrag bleibt `partial`“ nicht fängt. Die Tabelle führt dieselbe Mutation weiter als von `CRUD-03` gefangen (T33), die Z13-Zeile lässt „danach `complete`“ von c und d tragen, T83 nennt Nr. 10 i nicht, und die Zeile im Absatz *Mutationen und Raten* verweist auf „Rückgabe R12“, ohne zu sagen, dass sie entschieden ist. Eigene Messung bestätigt die Spezifikation: Mutation m7 (nach `PUT` an `partial` bleibt der Eintrag `partial`, nur im Zweig `methodePut` von `entscheide`) lässt `go test -run TestAbnahme ./src/bootstrap/` **grün** (`ok`, 26,3 s). Der Plan behauptet also an drei Stellen einen Fänger, der nicht existiert, und führt die Entscheidung des Architect an der vierten nicht nach. Das Negativ selbst trägt (Grund und Folge stehen in Nr. 10 i, der Ersatz-Fänger `TestPutPartial` im Test des Dienstes).
- `Auflage an den Planner` (vor der Closure, ein Commit, `AGENTS.md` §3.9): in T33 die Mutation „der Eintrag bleibt nach dem `PUT` `partial`“ streichen und als eigene Zeile (oder in T83: „Nr. 10 e bis i“) als akzeptiertes Negativ mit Grund und Folge von Nr. 10 i führen; in der Zeile `T32 bis T35` des Absatzes „Rückgabe R12“ durch „akzeptiertes Negativ, `SPEC-049` Nr. 10 i“ ersetzen; Z13 auf „204, danach `GET` 200 (der Zustand `complete` ist nicht beobachtbar, Nr. 10 i)“ zurücknehmen. Der Zähler *a* = 16, *b* = 67 der Marke bleibt (der Planner entscheidet, ob eine Zeile T84 die Zahl 83 ändert; die Marke ist gesetzt und gemessen, `SPEC-128` Nr. 5).

Sonst: Zusage ohne Fänger und ohne Negativ habe ich nicht gefunden. Die Zeilen T79 bis T83 tragen Grund und Folge in der Spezifikation (Nr. 10 a bis h).

## 4. Fix-Runde gegen die Klasse aufgelöst (§3.17)

Messender: Verifier-Agent. Container aus dem Image `odata-recorder:race` (vom `make gates`-Lauf auf HEAD gebaut), ohne Netz, `--memory 3g`, ein Container zur Zeit, unter `flock /tmp/odata-recorder-heavy.lock`; je Mutation Wegwerfkopie von `/src`, `perl -0pi` genau einmal, geprüft, dass die Datei sich änderte und übersetzt. Skripte im Scratchpad: `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/mut/`. Last bei den Läufen 4 bis 8.

| Befund | Menge (Wortlaut der Zusage) | Elemente | Mutation, die der Befund nicht nannte | Ergebnis |
|---|---|---|---|---|
| F1 (`Q-10` b) | die Elemente der Ordnung: Zahl der ausgewerteten, Richtung des 1. und 2., Reihenfolge der Auswertung, Typ der Spalte | 5 | m10: die Elemente von `$orderby` werden in umgekehrter Reihenfolge ausgewertet; m3: `Edm.Boolean` ordnet `true` vor `false` (`wahrheitsRang`) | rot `TestAbnahmeMatrix/ODV4-Q-10` (beide) |
| F1 | die vom Befund genannten | 3 | m1: nur das erste Element; m2: alle Elemente in der Richtung des ersten | rot `ODV4-Q-10` (beide) |
| F6 (`CRUD-03` c/d, `CRUD-04` f) | Zellen *Eintrag × Methode* von `SPEC-193` Nr. 2 | 20, in der Matrix 11 | m4: `PUT` an `partial` ergibt `ODR-E6003`; m5: `POST` über Tombstone ergibt `ODR-E6009` | rot `ODV4-CRUD-03` bzw. `ODV4-CRUD-04` |
| F6 | Mitglieder des gelöschten Eintrags beim `POST` über Tombstone (nicht genannt) | Payload | m6: der neue Eintrag übernimmt den Payload des Tombstone | rot `ODV4-CRUD-04` |
| F6 | `PUT` an `partial` bleibt `partial` | 1 | m7 | **grün**, wie in Nr. 10 i zugesagt (V1) |
| F2 (Frist-Wächter) | Obergrenze der Frist `time.Until(dl) > abDetFrist` | ab 120 s | f3: Wächter bis 126 s; f4: bis 150 s; f5: Wächter 1 s unter 120 s | rot (`…Frist` bzw. `TestAbnahmeDeterminismus`) |
| F2 | dasselbe, Werte dicht über 120 s | | f1: Wächter bis 121 s; f2: Wächter bis 124 s | **grün** (V2) |

### V2 · LOW — der Frist-Wächter ist nur 5 s über der Grenze gefangen

- `quelle`: `AGENTS.md` §3.15 (beide Seiten jeder Grenze), §3.17.
- `pfad`: `src/bootstrap/abnahme_determinismus_test.go` · `knapp, c0 := context.WithTimeout(context.Background(), abDetFrist+5*time.Second)` in `TestAbnahmeDeterminismusFrist`.
- `befund`: Die Fix-Runde führt die Fälle 125 s und 180 s; die Mutation B3 (150 s) und jede Obergrenze ab 126 s sind rot. Eine Obergrenze zwischen 120 und 125 s (f1: 121 s, f2: 124 s) bleibt grün: die Grenze ist von oben nur mit 5 s Abstand geprüft, von unten durch `TestAbnahmeDeterminismus` (f5 rot). Die Folge im Betrieb ist klein: der Test setzt die Frist selbst auf 120 s, der Wächter schützt allein vor einer künftig längeren Frist, und diese würde ab 125 s auffallen.
- `Empfehlung`: ein Fall mit `abDetFrist+time.Second` (Rest des Abstands bleibt größer als der Zeitverbrauch bis `Deadline`); kein Hindernis für die Closure.

## 5. Raten der nebenläufigen Zeilen (Befund F4, `AGENTS.md` §3.10)

Der Plan führt T55 bis T60 mit Zahl der Prozesse, Last und Messendem (Absatz *Raten der Isolation*, je 20 Prozesse mit und ohne Detektor, 20 von 20). Nachgemessen an Mutationen, die der Plan nicht nennt: Test-Binary je einmal ohne und mit `-race` übersetzt, dann `-test.run 'TestAbnahmeIsolation$' -test.count=1` in 20 Prozessen nacheinander. Last: Eins-Minuten-Wert zu Beginn und am Ende der Reihe.

| Nr. | Mutation (Wegwerfkopie) | ohne Detektor | mit Detektor | Last |
|---|---|---|---|---|
| T55x | `EntityRepository.liste` (dann `ListSet`) liest ohne Filter auf den Workspace (`WHERE ? <> ''`) | rot 20 von 20 | rot 20 von 20 | 8,42 / 4,82; 4,82 / 4,39 |
| T58a | `CreateWorkspace` prüft den Namen eines Snapshots global (`WHERE ? <> '' AND name = ?`) | rot 20 von 20 | rot 20 von 20 | 5,64 / 4,01; 4,01 / 4,38 |
| T60x | `EntityRepository.Put` schreibt die Kennung `ws-00` | rot 20 von 20 | rot 20 von 20 | 4,19 / 4,81; 4,81 / 4,42 |

Die Zeilen T56, T57 und T59 hat der Reviewer in 20 von 20 gemessen; zusammen sind T55 bis T60 bestätigt. **T54** (Menge von 19 Workspaces, ein Workspace läuft nicht) ist ungemessen und bleibt wie bekannt ein offener Punkt; die Zeile ist keine Wirkung eines Fremdzustands, sondern eine Zählung (`TestAbnahmeIsolation` prüft 20 und 228 Schritte) und kein Datenrennen. Die Raten von T62 und T71 bis T76 habe ich nicht neu gemessen; das Review hat T75, B4, T76 (a) und (c) nachgemessen (20 von 20).

## 6. Offene Punkte für die Closure (Planner)

### V3 · INFO

- §7 (Closure-Notiz mit Lerneintrag), Beobachtungs-Register, drei Paarungen und die Ausgänge der Risiken (siehe DoD-Tabelle) sind offen; kein Slice geht mit einem Risiko ohne Ausgang nach `done/`.
- Auslastung 2335 von 9700 = 0,24 (erwartet 0,56): der Planner meldet es in der Closure an den Architect (Register `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`); keine Rückführung, die Messung liegt unter der Marke.
- R10 (Nullprobe von `mutationslauf` scheitert am Paket `src/bootstrap`, 150 s gegen `-test.timeout=2m`) bleibt beim Planner. Die Abnahme-Tests selbst brauchen 26 s (`TestAbnahme*`), `TestAbnahmeMatrix` 6 s.
- Der Zweig steht 32 Commits hinter `main`; vor dem Merge `make gates` wiederholen und die Zahl der Codes angleichen (Plan §6, `Zahl der Codes`, Adresse Planner); dieser Slice fügt keinen Code hinzu.
- Die Befunde F5 bis F10 des Reviews (INFO) habe ich nicht erneut bewertet; F6 und F8 sind durch R11 erledigt, F10 (Zeitannahme von 400 ms im Fall *Frist im Lauf abgelaufen*) bleibt Hinweis ohne Aktion.

## 7. Rohbyte-Gegenlesen (`AGENTS.md` §3.18)

Dieser Bericht enthält Umlaute, kein BOM, kein NBSP und keine Escape-Folge als Wert; gegengelesen mit `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}|\xEF\xBB\xBF|\xC2\xA0' docs/reviews/2026-11-08-slice-mock-abnahme-verifikation.md`. Ergebnis: 0 Treffer.

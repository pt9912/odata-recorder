# Verifikationsbericht: slice-bootstrap-m6 (Teil D: Tests über den Admin-Listener) — 2026-11-27

**Gegenstand:** HEAD `541770fa`, Zweig `slice-bootstrap-m6`, gegen DoD und Plan `slice-bootstrap-m6`. Genau eine Verifikation. · **Rolle:** Verifier · **Modell:** Sonnet 5.5

**Gelesen:** Plan (§1 bis §8); `SPEC-300` Nr. 12 bis 14 samt *Cursor-Test* und *Ungeprüft* und `SPEC-139` am Stand `main` `ffa91e04`; `SPEC-037` Nr. 7 und 10; `LH-FA-02`, `LH-FA-09`, `LH-FA-10`, `LH-FA-13`, `LH-FA-14`, `LH-FA-18`, `LH-QA-04` (Lastenheft 0.10.0 aus `main`); [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); Report `2026-11-27-slice-bootstrap-m6-review`.

## Verdikt

**Bestätigt mit Auflagen.** Die drei Liefer-Punkte der DoD sind erfüllt, die Tests sind grün, am gebauten Binary ist in diesem Lauf kein Produktfehler gefunden. V-1 und V-2 sind vor der Closure im Plan zu erledigen; beide ändern keinen Code.

| ID | Kategorie | Auflage |
|---|---|---|
| V-1 | MEDIUM | Der Fix zu Review F-1 führt drei Formen der fehlerhaften Anfrage (405, 404 des Pfads, unbekanntes Feld) und nennt keine Prüfgrenze (`AGENTS.md` §3.15, §3.17). Die Zusage in D02, „eine fehlerhafte Anfrage mit gültigem Token ist `ODR-E2102` ohne Zustandsänderung“, ist weiter als der Fänger und im Wortlaut zu weit: am Binary antworten fehlerhafte Anfragen auch mit `ODR-E5005`, `ODR-E6101`, `ODR-E7001`, `ODR-E7002`, `ODR-E7003`, `ODR-E7007`, `ODR-E7009`, `ODR-E7010`, `ODR-E7011`. Mutation f1c (eine doppelte Session-ID beendet vor dem 409 die vorhandene Session) bleibt in den M6-Tests und in Stufe 11 grün, Mutation f1b (doppelter Schlüssel wird angenommen) in den M6-Tests; beide werden allein in Tests des Pakets `admin` am Fake rot. Zu erledigen: D02 und die Zeile `LH-FA-13` Negative auf die drei Formen begrenzen und den Rest als benanntes Negativ mit Grund, Folge und Fänger am Fake führen (oder die Fälle ergänzen). |
| V-2 | LOW | §6 *405 am Listener* steht auf *offen* (Adresse Architect). `SPEC-300` *Ungeprüft* sagt auch auf `main` weiter, der 405-Nachweis laufe am Listener in 0 von 4 Modi; `TestM6AbweisungOhneAenderung` führt ihn in 4 Modi. Die Closure braucht den Ausgang. |

## DoD

| DoD-Punkt | Befund | Beleg |
|---|---|---|
| 1 M6-Test (D01, D02, D07) | erfüllt; die Hälfte *fehlerhaft* mit V-1 | `TestM6Rundlauf`, `TestM6AbweisungOhneAenderung` (66 Zellen, 54 fehlerhafte Anfragen), `TestM6Cursor` selbst gefahren, grün; Folge der Schritte, elf Zeilen der neun Arten, sechs Zellen der Lesewege am Quelltext gegen `SPEC-300` Nr. 14 gelesen |
| 2 Zellen Modus × Art (D03) | erfüllt | `TestAuditAdminZellen`, `TestAuditAufrufstellen` grün; 32 Zellen, 29 mit Eintrag, die Ausnahme ist aus dem Test entfallen |
| 3 Stufe 11 (D04, D05) | erfüllt | `TestE2EAdmin` am selbst gebauten Binary: fünf Teilfälle PASS, `-count=3` in einem Prozess grün; Schritte von `SPEC-037` Nr. 7 gelesen; das Token auf Deskriptor 2 macht die Stufe rot (f2a, f2b); der Satz aus `LH-FA-14` 0.10.0 (nicht angehängt: `incomplete` an der Session) ist vom Fall *Speicher voll* getragen |
| 4 `make gates` | erfüllt, nicht neu gefahren | `.harness/state/gates-passed.head` ist `541770fa`, `gates-passed.diffsha` gleicht dem selbst gerechneten Hash des sauberen Baums vor diesem Bericht; Stempel 20:51, Commit 20:32. Der Stempel sagt nicht, welche Stufen aus dem Cache kamen |
| 5 Review | erfüllt | Report liegt vor; die Fix-Runde ist unten gegen die Klassen geprüft, ein zweites Review gibt es nicht |
| 6 Doku-Update | erfüllt | `SPEC-037` *Stand* (elf Stufen) und Nr. 10 a, h, j; `harness/README.md` §Sensors; `e2e-abdeckung.md` zeigt auf Zeile 143 (`stufe 11`) |
| 7 Handbuch | erfüllt, soweit vor der Closure möglich | unter `src/` allein zwei `_test.go`; kein Code, Schlüssel oder Kommando neu; der Satz in §7 fehlt noch |
| 8 bis 11 Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Closure) | §7 leer; §6 mit einem Punkt *offen* (V-2) |

## Plan gegen Code

- Die geänderten Dateien des Slice gleichen §3 (dazu `roadmap.md` und `welle-admin-api.md` aus dem Start); kein Produktcode.
- Größe: `git diff --numstat 5dd673ca HEAD` über die Pfade des Blocks gibt 1498, unter `ueberschreitung: 1500`.
- `main` ist 8 Commits voraus (Lastenheft 0.10.0, Spezifikation); `git merge-tree` meldet keinen Konflikt. Nach dem Merge nennt §6 (*Frage an den User*) noch das unveränderte Lastenheft und den alten Namen des Satzes in `SPEC-124` Nr. 12 (INFO).
- Tabelle *Kriterien · Träger*: zwölf Zeilen der Rückverfolgungs-Matrix, jede in der Tabelle; die genannten Tests der fünf Kriterien ohne Szenario und der Entity-State-Hälfte von `LH-FA-09` Boundary bestehen und sind grün (selbst gefahren: `bootstrap` und `baselineimport`; `TestWorkspaceResetAufDerDatei` setzt tatsächlich zuerst einen unveränderten Workspace zurück). `TestE2EAufzeichnung` und `TestE2ESchutz` nicht neu gefahren.

## Fix-Runde gegen die Klassen (`AGENTS.md` §3.17)

Eigene Mutationen in einer Wegwerfkopie, je an einem Element, das der Befund und die Messung des Implementers nicht nennen; je ein Lauf, deterministisch.

| Mutation | Klasse | M6-Tests | Stufe 11 | Paket `admin` |
|---|---|---|---|---|
| f1a: `POST /__admin/imports` nimmt ein unbekanntes Feld an | F-1 | rot (`record fehlerhaft POST imports mit unbekanntem Feld`) | nicht gefahren | rot |
| f1b: doppelter Schlüssel im Body der Session angenommen | F-1, außerhalb der drei Formen | **grün** | nicht gefahren | rot (`TestS6S7UngueltigerBody`) |
| f1c: doppelte Session-ID beendet die vorhandene Session | F-1, semantisch fehlerhaft | **grün** | **grün** | rot (`TestS11S13Fehler`, `TestFolgeOhneZustand`) |
| f2a: Fast-Treffer *erste Hälfte* schreibt den Kandidaten auf Deskriptor 2 | F-2 | — | rot (beide Läufe) | — |
| f2b: `POST /__admin/sessions` mit gültigem Token schreibt das Token auf Deskriptor 2 | F-2 | — | rot (Lauf nach dem Neustart) | — |
| f3a: nicht redigierbarer Body auf `stderr` | F-3 | — | rot (Marker, vor und nach dem Herunterfahren) | — |

F-2 und F-3 decken ihre Klasse. F-1 deckt die drei genannten Formen, nicht die Klasse (V-1). F-4: der Satz ist aus `harness/README.md` entfernt, eine zweite Stelle im Index fand sich nicht. F-5 und F-6 sind nachgezogen. F-7: Stufe 11 ist wiederholbar; dasselbe Muster in `volleKonfig` (älterer Helfer, außerhalb des Blocks) steht unverändert (INFO).

## Stichprobe am gebauten Binary

Eigene Wegwerf-Probe, Binary wie Stufe `build` gebaut, Prozesse `record`, `mock`, `replay` auf einem `DATA_DIR`. Zustandsvergleich über die Lesewege (Sessions, Workspaces, Snapshots, Coverage, Exchanges) und `audit.jsonl`.

| Kriterium | Beobachtung |
|---|---|
| `LH-FA-13` Negative, mit Token | 26 Anfragen, 25 abgewiesen (400, 404, 409, 422), keine änderte einen Leseweg; die eine angenommene ist ein unbekannter Query-Schlüssel an `GET …/exchanges`, nach `SPEC-293` ohne Wirkung. Neue Zeilen in `audit.jsonl` nur `import_fehlgeschlagen` |
| `LH-FA-09` Happy Path | zwei Workspaces einer Baseline, `POST` und `DELETE` an A: B unverändert |
| `LH-FA-09` Boundary (Entity-State) | `reset` am unveränderten Workspace B: Collection gleich der Baseline |
| `LH-FA-09` Negative | `reset` unbekannt: 404 `ODR-E7001`, A und B unverändert |
| `LH-FA-10` Negative | unbekannte, leere und fremde Snapshot-ID: 404 `ODR-E7002`, kein Teilzustand |
| `LH-FA-18` Negative (ungültiger Key) | fehlender und typfremder Key: 422 `ODR-E6101`, die bisherige Baseline liefert unverändert |
| `LH-FA-02` Happy Path | `GET` und `POST`: `sequence_no` in der Folge des Sendens in den Exchanges, Wiedergabe beider im Modus `replay` |

Nicht am Binary geprobt: `LH-FA-18` mit widersprüchlichen Werten und mit wohlgeformtem, nicht passendem `metadata_hash` (die Probe traf nur die Formprüfung der Anfrage, 400 `ODR-E2102`) und `LH-FA-10` Boundary; für sie gelten allein die oben gefahrenen Tests der Dienste.

## Grenzen dieses Laufs

Ein Container aus `odata-recorder:test` mit Wegwerfkopie von `541770fa` (`--network none --memory=4g --memory-swap=4g --cpus=4`, `tmpfs` `/voll` von 1 MiB, kein Mount) unter der gemeinsamen Sperre, Last des Hosts 2,6 bis 4,8. Die Stufe lief nicht über den Treiber und ohne `--read-only`. Nicht gefahren: `make gates`, `make gegenproben`, der Detektor, Raten. Danach allein `make docs-check` für diesen Bericht.

# Verifikation: slice-bootstrap-m3-e2e-mock — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-037` Nr. 5 bis 7 und 10 i bis l, `SPEC-139` Nr. 11 f, `SPEC-195` Nr. 5, die Tabelle *Zusage · Fänger* E01 bis E09 und die Hard Rules; nicht Diff-Review (siehe `2026-11-08-slice-bootstrap-m3-e2e-mock-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-bootstrap-m3-e2e-mock`, HEAD `6c8a6fe8` (Fix-Runde `4b0e59ff` nach dem Review, Architect `56c413d1` und `d2f848ae`), Worktree `/Development/odata-recorder-wt-bootstrap-m3-e2e-mock`. Plan: `docs/plan/planning/done/slice-bootstrap-m3-e2e-mock.md`. Bezug: `LH-FA-12`, `LH-FA-09`, `LH-FA-03`.

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden, mit Auflagen an den Planner (Closure); keine Fix-Runde am Code.** Alle neun Stufen von `make test-integration` sind im eigenen Lauf grün, `make gates` lief im eigenen Lauf ohne Fehlerausgabe durch, der Diff liegt in den Pfaden des Plans (Marke 1100, Messung 570), `Dockerfile`, `Makefile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und `src/` sind unberührt, und die Fix-Runde deckt die Klassen F1 bis F4. Zwei Elemente der Klasse *Zuordnung durch den Header*, die der Plan nicht in seine Menge nahm, bleiben am Binary ungefangen (V1, LOW, im Prozess gefangen).

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| Diff gegen Plan | `git diff --name-only fc6d1a6b..HEAD -- . ':!docs'`: `harness/vertraege/SPEC-037-test-integration.md`, `tests/e2e/lauf_test.go`, `tests/e2e/stufe_mock_test.go`, `tests/e2e/uebergang/{main,uebergang,uebergang_test}.go`, `tools/harness/test-integration.sh` — genau die Zeilen der Tabelle in §3 plus das erzeugte `docs/plan/planning/e2e-abdeckung.md`. `git diff --stat fc6d1a6b..HEAD -- Dockerfile Makefile harness/mk harness/README.md tools/gegenproben harness/gegenproben.stempel`: leer. `src/`: leer (`produktcode: 40` im Block gilt für den Übergang unter `tests/e2e/`, kein Pfad in `src/`). |
| Größenmarke | Summe `git diff --numstat fc6d1a6b..HEAD -- tests/e2e/ tools/harness/test-integration.sh` = 570 (Plan: 570); `make groessenmarke` im Gate-Lauf: `slice-bootstrap-m3-e2e-mock.md: ok, Messung 570 von Marke 1100`. |
| `make gates` | einmal unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/vf-e2emock-7k2/gates.out` (695 Zeilen): `a-check gesamt: 0 Befund(e)`, `baseline-verify: v6.18.1 OK`, `d-check: 905 Datei(en) geprüft, 0 Befund(e)`, Stufen `test`, `build`, `race`, `handbuch`, `katalog`, `lint` ohne Fehler (race, katalog und lint kamen aus dem Cache der unveränderten Eingaben), `migration-check: ok`, `zeichen: ok, 1821 Datei(en)`. Kein `ERROR`, kein `***`, kein `FAIL` im Log. **Grenze:** der Exit-Code des `make` ist nicht aufgezeichnet (nur die Ausgabe); das Fehlen jeder Fehlerzeile und der letzte Schritt `zeichen: ok` tragen das Urteil. |
| `make test-integration` (Pflichtlauf, mit eigenem Image-Tag) | Kopie des Treibers mit `IMAGE=odata-recorder:verifier-e2emock` (nur diese Zeile geändert, Tag nach dem Lauf entfernt), unter `flock`: Stufen 1 bis 9 grün, Exit 0, darunter `Stufe 8: TestE2EMock grün` und `Stufe 9: TestE2EMockNeustart grün`. |
| `make e2e-abdeckung` | `unveraendert — docs/plan/planning/e2e-abdeckung.md (9 Stufen, 9 Deklarationen, 0 ohne Kennung)`; kein Schreiben, Baum danach sauber. Das Dokument ist damit gleich dem Erzeugnis. |
| Fehlercodes | `grep -oE '"ODR-E[0-9]{4}"' src/hexagon/model/fehler.go \| sort -u \| wc -l` = 40 im Zweig und auf `main`; `src/` unberührt. Passt zu `odr_errors_total` (40, `SPEC-230` Nr. 8). |
| Merge-Nähe | `git merge-tree --write-tree --name-only HEAD main`: Rückgabe 0, Baum `d7da61de…`, kein Konflikt. `main` liegt mit einem reinen `docs/`-Commit vor (`ada0ab5f`, Umplanung). `e2e-abdeckung.md` ist heute nicht auf `main` verändert; sobald `slice-bootstrap-m4` oder `-m6` in `main` landen, gilt die Regel des Plans (neu erzeugen, nicht von Hand lösen). |
| Überschneidung mit `slice-test-contract` und `adr-gate-laufzeit` | `slice-test-contract` berührt `Dockerfile`, `harness/README.md`, `harness/gegenproben.stempel`, `harness/mk/test-contract.mk`, `SPEC-042`, `src/bootstrap/vertrag_http_test.go`, `tests/integration/vertrag_fehlerklassen_test.go`, `tools/harness/test-contract.sh`; `adr-gate-laufzeit` allein `harness/mk/test-cache.mk`. Keine gemeinsame Datei mit diesem Zweig (weder in `tools/` noch `harness/README.md`; dieser Slice berührt `tools/harness/test-integration.sh`, jener `tools/harness/test-contract.sh`). `git merge-tree` je Paar (`HEAD` mit `slice-test-contract`, `HEAD` mit `adr-gate-laufzeit`): Rückgabe 0, kein Konflikt. |
| §3.2 / §3.18 | im Diff außerhalb von `docs/`: `nolint` 0 Treffer; `LC_ALL=C grep -nP '\\u[0-9A-Fa-f]{4}'` über die hinzugefügten Zeilen: 0 Treffer; BOM 0. Dieser Bericht ist nach dem Schreiben mit demselben Muster gegengelesen (siehe Schluss). |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Stufen `TestE2EMock` und `TestE2EMockNeustart`, je drei Zeilen im Treiber, Befehl `reset` mit Test; `make test-integration` grün | **erfüllt, Haken fehlt** | Treiber, Stufen 8 und 9 (drei Zeilen je Stufe: `echo`, `e2e_abdeckung`, `stufe`); `TestAufrufformen`, `TestReset`; voller Lauf grün (oben). |
| Abdeckungs-Dokument aus `make e2e-abdeckung`; E08/E09 mit Grund und Folge in der Spezifikation | **erfüllt, Haken fehlt** | Dokument unverändert beim Erzeugen; `SPEC-037` Nr. 10 i, j (vorhanden), k und l (neu) tragen Grund und Folge. |
| `make gates` grün; Pflichtlauf `make test-integration` | erfüllt | siehe §1. Nach einem Merge von `main` oder eines der Nachbarn einmal wiederholen. |
| Review durchgeführt, Report liegt vor | erfüllt (Haken fehlt) | `docs/reviews/2026-11-08-slice-bootstrap-m3-e2e-mock-review.md`; 0 HIGH, 0 MEDIUM, 4 LOW (F1 bis F4), 2 INFO; Fix-Runde `4b0e59ff`. |
| Doku-Update (öffentlicher Vertrag) | erfüllt | `SPEC-037` Nr. 5, 6, 7, 10 k, l und die Absätze *Festlegung* und *Stand* nachgezogen. |
| Handbuch-Kapitel entfällt | erfüllt | `make handbuch-check` im Gate-Lauf ohne Fehler; kein Anwender-Ergebnis. |
| Closure-Notiz mit Lerneintrag; Register; Paarungen | **offen, Sache des Planners** | §7 ist `Wird bei der Closure gefüllt`. |
| Ausgang je Risiko (§6) | **offen für eine Zeile** | alle Zeilen tragen *entschieden* oder *entfallen*, außer der Zeile zur Stufe 9 (Klasse `nebenläufig`): **offen**, den Ausgang setzt der Planner bei der Closure; die Raten (30 von 30, 10 von 10, 0 von 30) stehen im Plan. |

Liefer-Punkte: drei. Kein Haken gesetzt; das setzt der Planner.

## 3. Tabelle *Zusage · Fänger* gegen den Vertrag

| Zeile | Vertragsstelle | Befund |
|---|---|---|
| E01 | `SPEC-037` Nr. 5, 7; `SPEC-110` Nr. 14 c | Test deckt Start ohne `UPSTREAM_URL`/`PUBLIC_BASE_URL` (der Test bricht ab, wenn die Umgebung eine der beiden trägt), `201` mit `Location` aus `LISTEN`, `GET`, `PATCH` `204`, `DELETE` `204`, `404` `ODR-E6008`, `complete`, `501` für Aufruf und Operatorwort in `$filter` (Schleife über 2). Gleich dem Wortlaut. |
| E02 | Nr. 7, `SPEC-195` Nr. 5, 8 | Schleife 6 Formen × 5 Requests mit Zählprüfung (`n != 30`); Orakel Status und Code je Form; Zustand danach gleich der Antwort vor den Mutationen. Reset über den Übergang. Gleich dem Wortlaut; *zwei Schreibweisen* als Negativ am Binary (Nr. 10 l) ehrlich benannt, Fänger im Prozess. |
| E03 | Nr. 10 k | akzeptiertes Negativ mit Grund und Folge; der Zähler bleibt als Kontrollwert (`keinKontakt` nach jedem Schritt). Wortlaut in Nr. 7 und Plan stimmen überein. |
| E04 | Nr. 5, 7; `SPEC-139` Nr. 11 f | `SIGTERM` erst nach dem letzten Lesen (Orakel), zwei Neustarts, Exit 0 geprüft, 4 Lesezugriffe × 2 Neustarts; Bereitschaft im Modus `mock` ohne Request (Nr. 5, neu). Gleich dem Wortlaut. |
| E05 | Nr. 7, `SPEC-139` Nr. 11 f | vier Pfade, je `404` `ODR-E7001`, Zählprüfung (`n != 4`). |
| E06 | Nr. 6 | Schleife 3 Befehle × 6 Formen (18, Zählprüfung) plus 2; `TestReset` für unbekannten Workspace, Kennung einer Session, nicht bestehendes `DATA_DIR`, Datei statt Verzeichnis, nichts angelegt. Der Wortlaut von Nr. 6 ist nach F4 eindeutig geordnet (die Exit-2-Regel für `import`/`zeige` steht vor dem Absatz zu `reset`, `reset` hat seine eigene Fassung). |
| E07 | Nr. 2, 3 | selbst gefahren, siehe §4. |
| E08, E09 | Nr. 10 i, j | Negative mit Grund und Folge, unverändert. |

Die Mengenangaben in §3 des Plans (30, 20 = 18 + 2, 8, 4) stimmen mit den Zählprüfungen im Code überein.

## 4. Zeile E07 (vom Reviewer nicht gefahren)

Quelle in Wegwerfkopien unter `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/vf-e2emock-7k2/`; das Repo blieb unverändert.

| Mutation | Lauf | Ergebnis |
|---|---|---|
| Anker `TestE2EMokNeustart` in der `e2e_abdeckung`-Zeile von Stufe 9 | `bash tools/harness/e2e-abdeckung.sh` mit der Kopie als Quelle | rot: `FEHLER — Deklaration ohne Stufe: der Anker loest in der Region der Stufe, die sie deklariert, nicht woertlich auf.` mit Zeile 134 und Region 133 bis 138; Skript-Exit 1, `make e2e-abdeckung` mit derselben Quelle Exit 2 (der Plan nennt „Exit 2": das ist der `make`-Exit, das Skript endet mit 1). |
| Name der Stufe `stufe 9 TestE2EMockNeustar` | gekürzte Kopie des Treibers (Stufen 1 bis 8 ausgelassen), gebaut mit eigenem Tag `odata-recorder:verifier-e2emock`, unter `flock` | rot: `Stufe 9: Test TestE2EMockNeustar nicht gefunden`, Exit 2. |

Beide Mutationen werden aus dem benannten Grund rot (kein Aufbaufehler: die Stufe `integration` baute, die Liste lief).

## 5. Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Eigene Mutationen an Elementen, die der Befund und die Tabelle des Implementers nicht nannten. Jede lief in einer Kopie des Baums, gebaut als Stufe `integration` bzw. `test` des `Dockerfile`, Container ohne Netz, `--memory 2g`, unter `flock`, einmal je Mutation (keine Raten). Jede Mutation baut.

| Befund · Klasse | Menge (Elemente) | Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|---|
| F1 · Header der Zuordnung | 6 Formen × 5 Requests = 30, Zählprüfung im Test | **A:** der Check `einzelwert` des Mock-Dienstes entfällt für `HEAD` (`if !ok && a.Methode != "HEAD"`); `HEAD` ist nicht unter den fünf Requests | **grün** `TestE2EMock` (`CRUD` und `Zuordnung` PASS). Element außerhalb der Menge des Plans. |
| F1 · dieselbe Klasse *mehr als ein Wert* | dieselben 30 | **B:** `len(w) != 1` wird `len(w) != 1 && len(w) != 3` (drei Werte angenommen, zwei weiter abgelehnt) | **grün** am Binary. Die Form *drei Werte* steht nicht unter den sechs; im Prozess führt sie `TestZuordnung` (`src/hexagon/services/mock/zuordnung_test.go`, Zeile `drei Werte`). |
| F4 / E06 · `reset` bei fehlendem `DATA_DIR` | `TestReset`: 2 Formen (fehlt, Datei) | **C:** `|| !fi.IsDir()` entfällt in `zuruecksetzen` (`fi` durch `_` ersetzt, damit es baut) | `pruefung test` grün: **gleichwertiger Mutant** — bei einer Datei an der Stelle scheitert `LadeKonfiguration`/`MkdirAll` ohnehin mit Exit 1 und Meldung, die Beobachtung ist gleich; das Verhalten (Exit 1, nichts angelegt) bleibt von `os.Stat` allein getragen, und diese Mutation (`Stat` entfällt) nannte der Implementer und sah rot. Kein Befund. |

Die Fix-Runde selbst (F1: Form *zwei verschiedene Werte*; F2: Wortlaut Nr. 6 zu `E2E_UPSTREAM_URL`; F3: Negativ `ODR-E6004` im Raster; F4: Ordnung von Nr. 6) gelesen und gegen den Code aufgelöst:

- **F1:** die Menge steht im Test als Struktur `zuordnungsFormen` (6 Elemente) mit Zählprüfung 30; die Form *zwei verschiedene Werte* ist dabei. Die Mutation R8 des Reviews ist durch die Form gefangen (laut Plan 5 rote Elemente; die Zeile steht in der Mutationstabelle); die Klasse ist am Binary auf *zwei Werte* beschränkt, nicht auf *mehr als zwei* (V1).
- **F2:** `SPEC-037` Nr. 6 sagt jetzt, der Aufbau des Handle lese die Variable für jeden Befehl und nur das Ergebnis von `import` hänge von ihr ab. Das stimmt mit `handle` überein (liest für jeden Befehl, `upstreamVar`). Es bleibt eine Aussage ohne Fänger, aber sie behauptet nichts, was ein Test fangen müsste (kein beobachtbarer Unterschied); kein Befund.
- **F3:** das Negativ steht im Raster (Zeile *Start auf vorhandenem DATA_DIR*, Zelle *leerer Wert*) mit Grund und Folge. Wortlaut stimmt mit `SPEC-195` Nr. 3 (f vor g) überein.
- **F4:** `SPEC-037` Nr. 6 liest sich jetzt in der Reihenfolge `import`, `zeige` (mit der gemeinsamen Exit-2-Regel), dann `reset` mit eigener Regel; die Kennung einer Session als unbekannter Workspace (`ODR-E7001`) steht bei `reset`, `ODR-E7003` bei `zeige`. Die Wiederholung der Exit-2-Regel ist aufgelöst.

## 6. Befunde

### V1 · LOW

- `quelle`: `AGENTS.md` §3.17 (Fix deckt die Klasse), §3.15
- `pfad`: `tests/e2e/stufe_mock_test.go` · `zuordnungsFormen`, `requests` in `zuordnung`
- `befund`: Die Menge des Fixes (6 Formen × 5 Requests) ist aus dem Plan gebildet; `SPEC-037` Nr. 7 und `SPEC-195` Nr. 5 sagen aber „ein Request ohne …" und „mehr als ein Wert". Am Binary bleiben grün: drei Werte (Mutation B) und ein Request mit `HEAD` (Mutation A; `PUT` und `PATCH` ebenso, nicht gefahren). Im Prozess fängt `TestZuordnung` die Form *drei Werte*; ein Eingang, der Methoden verschieden behandelte (Mock-Adapter), bliebe am Binary ungefangen.
- `folge`: kein Betriebsfehler bekannt: die Zuordnung liegt im Dienst an genau einer Stelle (`einzelwert` in `Mock`) vor der Methodenwahl; eine Mutation, die sie nur für eine Methode umgeht, ist künstlich. Als akzeptiertes Negativ mit Grund und Folge in `SPEC-037` Nr. 10 (neben l) zu benennen oder die Menge um `HEAD` und die Form *drei Werte* zu erweitern; Entscheidung beim Planner.
- `verifizierbar`: ja — Mutationen A und B in Kopien, `TestE2EMock` grün.

### V2 · INFO

- `befund`: Der Plan nennt für die Mutation *Anker mit falschem Namen* „`make e2e-abdeckung`: Exit 2". Das Skript `tools/harness/e2e-abdeckung.sh` endet mit 1, `make` meldet 2. Die Aussage stimmt für den `make`-Aufruf; kein Nachzug nötig.

## 7. Auflagen an den Planner

1. Closure: §7 füllen (Lerneintrag, Register-Antwort), DoD-Haken setzen, Ausgang der Zeile zur Stufe 9 (*nebenläufig*) setzen (Raten im Plan), danach `git mv` nach `done/` (Inhalt zuerst, `AGENTS.md` §3.3).
2. V1 entscheiden: Negativ in `SPEC-037` Nr. 10 oder die Menge um `HEAD` und *drei Werte* erweitern (kleine Änderung am Test, Folgerunde nicht erforderlich).
3. `make gates` einmal nach dem Merge wiederholen: `main` liegt mit Planungs-Commits vor, `slice-test-contract` ändert `Dockerfile`, `harness/README.md` und den Stempel der Gegenproben; die Zweige überschneiden sich in keiner Datei, aber der Stempel und der Gate-Lauf gehören zusammen.
4. Merge-Reihenfolge der Nachbarn im Treiber (`slice-bootstrap-m4`, `-m6`): Stufenzahl neu setzen und `make e2e-abdeckung` neu erzeugen, kein Konflikt von Hand lösen (Plan §6).
5. Der Exit-Code des eigenen `make gates` ist nicht aufgezeichnet; wer ganz sicher gehen will, wiederholt ihn mit `echo EXIT=$?`.

## Gegenlesen (`AGENTS.md` §3.18)

Diese Datei ist nach dem Schreiben mit `LC_ALL=C grep -nP '\\u[0-9A-Fa-f]{4}|\xEF\xBB\xBF|\xC2\xA0'` gegengelesen.

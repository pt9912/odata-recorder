# Review-Report: slice-telemetry-einbau-mock-mutation (Teil 2) — 2026-11-21

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-telemetry-einbau-mock-mutation`, HEAD `12a7548a`, Messbasis `3227ffb3` (Basis des Zweigs `main` `19d0cf39`); Teil 2 des Schnitts, Zeilen M1 bis M10: Produktcode `src/hexagon/services/mock/schreiben.go` (+27/−1) und der Kommentar in `mock.go`; Tests `telemetrie_mutation_test.go` (Pakete `mock` und `bootstrap`), der Wächter in `neueUmg` (`fakes_test.go`), `Transaktion.Ende` in `fakes_schreiben_test.go`, die angehobenen Orakel in `telemetrie_ablauf_test.go` und `telemetrie_gleichzeitig_test.go`, der Helfer `baueWurzelMit` in `telemetrie_wurzel_test.go`; Plan und Roadmap.

**Skill:** `.harness/skills/reviewer.md` @ `12a7548a` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-21

**Eingangs-Kontext:**

- Plan `slice-telemetry-einbau-mock-mutation`: §1 (Ziel, vier Sendungen), §2, §3 (Tabelle M1 bis M10, Mengen, Zusatztabelle, akzeptierte Negative (1) bis (7), Absatz *Zusagen, Mutationen und Raten*), §4 (Block, Marke 1200, Eintrag `ueberschreitung: 1400`, Rückführungen), §6 (Raster, Risiken)
- `SPEC-290` (Nr. 2 b, 3, 4 c und d, 5, 6, 7, 8, 9), `SPEC-231` Nr. 4 und 7, `SPEC-139` Nr. 3, 11, 13, `SPEC-030`, `SPEC-195` Nr. 11 (Aufrufstellen S1 bis S26), `SPEC-128` Nr. 5 d und 10 g (Überschreitung)
- `AGENTS.md` §3.7, §3.9 bis §3.18
- `LH-QA-01`, `LH-FA-07`, `LH-QA-05`, `LH-QA-04`
- ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Bezug des Plans, nur genannt)
- Vorheriger Review am Modul: `slice-telemetry-einbau-mock` (Teil 1; Klassen *Zusage über Konfiguration ohne Fänger*, *Attribut je Zweig*, *Wirkungs-Fänger statt Datenrennen-Hoffnung*, *Kommentar nennt mehr, als der Fänger fängt*)

**Unverändert (geprüft mit `git diff 3227ffb3..HEAD`):** `workspace/` und `snapshot/` (kein `StartSpan`), `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/mk/`, `Dockerfile`, `tools/gegenproben/`, `harness/gegenproben.stempel`, `AGENTS.md`, `.claude/`, `.harness/skills/`. Keine Direktive `//nolint`, kein `t.Skip`; keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM und kein NBSP in den zwei neuen Testdateien (`LC_ALL=C grep -P`, 0 Treffer).

**Größenmarke (nachgemessen):** `git diff --numstat 3227ffb3..HEAD` über die Pfade des Blocks (`src/hexagon/services/mock/`, `telemetrie_wurzel_test.go` und `telemetrie_mutation_test.go` in `src/bootstrap/`): 1332 Einfügungen, 11 Löschungen, Summe 1343, wie berichtet, gegen die Marke 1200; Produktcode 27 Zeilen in `schreiben.go` und 1 in `mock.go`.

## Eigene Mutationen

Ein eigener Treiber im Scratchpad (`telemetry-einbau-mock-mutation-reviewer/`, außerhalb des Repos; Quelle des Treibers: der des Reviews von Teil 1): Container `odata-recorder:race` ohne Netz und Mount, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, Wegwerfkopie des Worktrees im Container, je Mutation bis zu drei Austausche `alt` gegen `neu` und `go test -count=1` (je Prozess ein Lauf; bei `race` mit Detektor und cgo), die Mutationen nacheinander, nie parallel zu `make gates`. `tools/harness/mutationslauf.sh` fährt `src/bootstrap` nicht (Nullprobe über 2 Minuten, bekannter Befund); die Mutationen mit Test in `src/bootstrap` laufen deshalb allein über diesen Treiber, der für das Paket `mock` dasselbe tut. `make gegenproben` ist nicht gefahren, der Stempel nicht berührt.

Zahlen: 56 Einträge ohne die fünf Nullproben (Austausch durch sich selbst; alle fünf grün), 48 verschiedene Änderungen (die Arme mit und ohne Detektor der Klasse `nebenläufig` einzeln gezählt). **51 Einträge rot** (aus dem benannten Test, `--- FAIL` mit Namen; keiner mit `PANIK` oder `ZEITLIMIT`), **4 grün**, 1 Treiber-Artefakt. Die vier grünen: (a) eine Änderung, die im Lauf nicht wirkte (`End` unter `ctx.Err() == nil` in Bestandstests, in denen der Kontextfehler vom Fake eingespeist wird und der `ctx` nie endet: gleichwertig, kein Befund); (b) und (c) die zwei Arme ohne Detektor der Änderungen *geteilte Variable ohne Port dazwischen* (akzeptiertes Negativ (5): 0 von 20 je Änderung, mit Detektor 20 von 20, bestätigt das Maß des Plans); (d) **F-1**. Das Artefakt: ein Eintrag am Dateiende einer Liste, dessen letzter Austausch durch einen Zeilenumbruch des Treibers brach; der Eintrag ist in einer weiteren Liste sauber nachgefahren (rot).

### Auflösung der Tabelle *Zusage · Fänger*

| Zeile | Eigene Mutationen (nicht die des Implementers; je Element der Klasse) | Ergebnis |
|---|---|---|
| M1 | `mutation_type` für `post` als `POST`, `delete` als `del`, `put` als `patch`; drittes Attribut `recorder.entity_set` mit dem Set des Requests; Schlüssel `recorder.workspace` statt `recorder.workspace_id`; zweites `End` nur bei `DELETE` | 6 von 6 rot (`TestMutationSpanJeOperation`, `TestMutationWorkspaceKennung`, `TestKeinRequestWertImMutationsSpan`, die Tabelle der Spans in `TestPflichtSchluesselMutation`; das zweite `End` färbt dazu rund 60 Bestandstests über den Wächter) |
| M2 | `StartSpan` hinter `ladeModell` (zwischen S5 und der Prüfung des Sets) | rot (`TestMutationSpanFolge`, `TestMutationEndeJederRueckweg`) |
| M3 | `End` nicht bei `ODR-E6009`, `ODR-E6010`, `ODR-E6003`, `ODR-E7001`, nicht bei `ODR-E6007` und `ODR-E6002` aus dem Leser, nicht bei Kontextfehler der Transaktion allein bei `PUT`; **nicht bei `ODR-E5005` aus S3**, nicht bei `ODR-E6004` (S3 und S5), nicht bei `ODR-E3001` (S4) | die ersten 7 rot (`TestKeinZaehlerImDienst`, `TestMutationEndeJedeEntscheidung`, `TestKeinZaehlerBeiKontextfehler`, `TestMutationWorkspaceKennung`; viele zusätzlich über den Wächter in Bestandstests); **`ODR-E5005` aus S3: grün (F-1)**; `ODR-E6004` und `ODR-E3001` rot, aber allein aus Bestandstests (F-1) |
| M4 | siehe M1; in der Ebene des Roots `put` als `patch` | rot (`TestWurzelMutationSpanZeile`) |
| M5 | Kennung aus `d.ID` (Kennung des Dienstes); in der Ebene des Roots Kennung in Großbuchstaben | rot (`TestMutationWorkspaceKennung` und sechs weitere; `TestWurzelMutationSpanZeile`) |
| M6 | `End` asynchron (`go span.End()`); `End` unter `ctx.Err() == nil` (20 Prozesse je Arm) | rot; Raten unten |
| M7 | abgeleiteter `ctx` an S2, S3, an `ladeModell` (S4, S5), an S19, an S21, an `StartSpan` | 6 von 6 rot; S3 und S4/S5 allein durch Bestandstests (`TestKontextMutation`, `TestAufrufzahlenUndFolge`, `TestPreferSchreiberFehler`), S2 und S19 auch durch `TestMutationCtxUnveraendert`, S21 durch `TestKontextMutation` und `TestMutationCtxUnveraendert`, `StartSpan` allein durch `TestMutationCtxUnveraendert` |
| M8 | Wert aus `If-Match` nur bei `DELETE`, `Prefer` nur bei `PUT`, `Content-Type` nur bei `POST`, `Accept` nur bei `DELETE`, `If-None-Match` nur bei `PATCH`, Literal des Schlüssels nur bei `PUT` (je ein Paar *Stelle × Methode*); `Log` im Zweig des Lesers | 7 von 7 rot (`TestKeinRequestWertImMutationsSpan`; der `Log` zusätzlich über die Zählprüfung des Quelltexts) |
| M9 | im Root: `Log` mit `err.Error()` ganz, auf 64 Bytes gekürzt, als Attribut gekürzt | 3 von 3 rot (`TestKanarienNameDerLaenge`, `TestKeinRequestWertInDerSenkeMutation`) |
| M10 | Art und Kennung der letzten Anfrage in einer Paketvariable (vor `Workspaces.Get` geschrieben); Span im Paketfeld, `End` am Ende der Funktion | rot; Raten unten |
| Orakel von Teil 1, `neueUmg` | `StartSpan` für die Paarung `ODR-E6002` vor dem Beginn; für `$metadata` mit `POST`; Beginn vor der Prüfung der Paarung | 3 von 3 rot (`TestKeinSpan`, `TestTelemetrieImQuelltext`, dazu `TestRangfolgeMutation`, `TestPaarung`) |

Die Mengen der Zusagen (`AGENTS.md` §3.17) hat der Review aus dem Wortlaut der Zusage gebildet, nicht aus den Mutationen des Implementers: *Stelle × Methode* (33 Paare), *Code × Stelle* der Rückwege (13 Stellen mit Code; die Codes der Stellen S3 bis S5), *Aufrufstelle des `ctx`* (S2 bis S5, S19, S21, `StartSpan`; die Stellen S17 bis S26 führt der Plan), *Länge des Namens* (255, 256, 257, 4096, vom Plan geführt). Alle Elemente, die der Plan nicht nannte, waren rot (S2, S3 und S4/S5 des `ctx` im Pfad der Mutation; Kennung aus dem Dienst; `If-Match` bei `DELETE`) bis auf die Menge *Codes der Stellen S3 bis S5* (F-1).

### Raten der Klasse `nebenläufig` (`-test.count=1` je Prozess, 20 Prozesse je Arm)

Messender: der Reviewer. Last des Hosts (Load Average der letzten Minute, bei 4 CPUs des Containers) von 3 bis 20; zwei weitere Linien liefen zugleich (Review von `slice-listener-tls-verdrahtung`, Architect von `slice-admin-sessions`). Die Raten sind Größenordnungen unter dieser Last, keine festen Werte.

| Mutation | mit Detektor | ohne Detektor |
|---|---|---|
| Art der letzten Anfrage in einer Paketvariable, vor dem Port geschrieben, nach ihm gelesen | 20 von 20 (`TestMutationSpanGleichzeitig`, `TestMutationAbbruchGleichzeitig`, Datenrennen) | 20 von 20 (Wirkungs-Fänger, Schranke an `Workspaces.Get`) |
| Kennung der letzten Anfrage, ebenso | 20 von 20 | 20 von 20 |
| Span im Paketfeld, `End` am Ende der Funktion (Port dazwischen) | 20 von 20 | 20 von 20 |
| Span im Paketfeld, unmittelbar vor `defer` geschrieben (kein Port dazwischen); Zähler hochgezählt nach `StartSpan` | 20 von 20 je Änderung | 0 von 20 je Änderung (akzeptiertes Negativ (5)) |
| `End` nur unter `ctx.Err() == nil` | 20 von 20 (`TestMutationAbbruchGleichzeitig`) | 20 von 20 |
| `End` asynchron (`go span.End()`) | ohne Detektor gemessen: in jedem Prozess rot durch mindestens einen deterministischen Test; die zwei Wirkungs-Fänger allein: `TestMutationAbbruchGleichzeitig` 15 von 20, `TestMutationSpanGleichzeitig` 18 von 20 | wie links |

Eine Panik oder ein Zeitlimit statt `--- FAIL` des benannten Tests trat bei keiner Mutation auf. Gesamtlauf `go test -count=1 ./src/bootstrap` (200 s) und `go test -race` des Pakets `mock` am Stand `12a7548a`: grün.

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Rückwege der Mutation S3 bis S5 führen als Fehler mit Code nur den rohen Fehler (`ODR-E9001`); die Codes, die diese drei Stellen verschieden bilden (`ODR-E5005` aus S3, `ODR-E6004` aus S3 und S5, `ODR-E3001` aus S4), führt `codeSzenarien` nur über die Query (`mitSpan`), nicht über die Mutation. Mutation „kein `End`, wenn `holeService` in `schreibe` `ODR-E5005` liefert“: alle Tests grün; dieselbe Änderung für `ODR-E6004` und `ODR-E3001` wird rot allein durch Bestandstests (`TestIfMatchVorTransaktion`, `TestRangfolgeMutation`, `TestZweigeGleichzeitigMutation`) über den Wächter. Szenario: ein Sonderweg für `ErrNichtGefunden` aus `Metadaten.Get` (bekannte Antwort `ODR-E5005`) mit eigenem Rückweg ohne `End` bleibt grün; `SPEC-290` Nr. 7 b nennt `ODR-E5005 (S3)` ausdrücklich als Span ab dem Beginn, und der Plan sagt in §3, die übrigen Codes führe `codeSzenarien`. | `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe); `SPEC-290` Nr. 7 b | `src/hexagon/services/mock/telemetrie_ablauf_test.go` · `mitSpan(model.CodeE5005, "E5005 nach dem Beginn"` und `src/hexagon/services/mock/telemetrie_mutation_test.go` · `{stelle: "S3", op: "Metadaten.Get", methode: "PATCH", id: "2"}` | ja — Mutation oben, `go test ./src/hexagon/services/mock` | Fänger führt nur eine Form der Eingabe: Codes der Stellen S3 bis S5 der Mutation |
| F-2 | LOW | Zwei Kommentare sagen, eine Schranke halte „an jedem Port vor der Transaktion“ bzw. an „den Ports vor der Transaktion“ alle 16 Aufrufe fest; die Stufen sind `Workspaces.Get` (S2), `Metadaten.Get`, `Blobs.Open`, `ParseMetadata` und der Eingang der Transaktion, `Eingabe.LeseEntity` (S17) ist keine Stufe. | `AGENTS.md` §3.11 | `src/hexagon/services/mock/telemetrie_mutation_test.go` · `Eine Schranke an jedem Port vor der Transaktion hält alle 16 Spans zugleich offen` und `den Ports vor der Transaktion und am Eingang der Transaktion` | nein — Lesen der Stufen in `fahreMutationen` | Kommentar nennt mehr, als der Fänger fängt |
| F-3 | LOW | Der Kopf von `TestKanarienNameDerLaenge` sagt, weder die ersten 8 Bytes des Namens „noch ein Teil davon“ erreichten Senke oder Attribut; der Test sucht allein `name[:8]` als Teilfolge, ein Teil unter 8 Bytes bliebe unter diesem Fänger unbemerkt (über Zeilenzahl und Attributliste der Nachbartests wäre er sichtbar, dort aber nicht benannt). | `AGENTS.md` §3.11 | `src/bootstrap/telemetrie_mutation_test.go` · `noch ein Teil davon; die Antwort ist ODR-E6007` | nein — Lesen des Tests | Kommentar nennt mehr, als der Fänger fängt |
| F-4 | LOW | Der Plan trägt für dieselbe Menge zwei Aussagen: die Zeile M9 und die Tabelle *Mengen* nennen 3 Methoden × 4 Längen gleich 12, der Absatz *Zusagen, Mutationen und Raten* und der Test 2 Modi × 3 × 4 gleich 24; die Überschrift des akzeptierten Negativs (6) sagt, `hybrid` stehe in drei der fünf Tests, der Maßsatz (3 von 5 Tests nur `mock`) und der Code sagen: `hybrid` läuft in zwei und fehlt in drei. | `AGENTS.md` §3.9 | `docs/plan/planning/in-progress/slice-telemetry-einbau-mock-mutation.md` · `| M-Canary (M9) | 3 Methoden × 4 Längen gleich 12 |` und `in drei der fünf Tests des Roots` | nein — Lesen von §3 | Plan trägt zwei Aussagen derselben Menge |
| F-5 | LOW | Die akzeptierten Negative (2) (Span des Hybrid-Fallbacks) und (4) (Dauer und Ausfall der Senke) nennen im Plan den Grund, die Folge nur als Verweis (§6, Zeilen Q21 und Q22 in Teil 1); die Folge im Betrieb steht in `SPEC-290` *Ungeprüft*, nicht in der Zeile. | `AGENTS.md` §3.15 | `docs/plan/planning/in-progress/slice-telemetry-einbau-mock-mutation.md` · `kein Nehmer, siehe §6.` | nein — Lesen des Plans | Akzeptiertes Negativ nennt die Folge nur als Verweis |
| F-6 | INFO | Größenmarke, Grund und Rückführung. Die Zahlen des Grundes stimmen mit dem Code überein (37 Fälle der Rückwege, 27 Entscheidungen, 33 Paare, 24 Läufe, 16 Goroutinen × 4 Methoden mit Schranke, Orakel von Teil 1, Wächter). Er trägt als Beschreibung; er beziffert die Zeilen je Ursache nicht, so dass nicht ablesbar ist, wie viel der 143 Zeilen über der Marke die Verdopplung der Läufe um den Modus (M9: 24 statt 12) und der Wächter tragen. §4 nennt als Rückführung `in-progress` → `next` bei „über der Marke (1200)“, und `AGENTS.md` §3.14 führt „eine Marke, die nach dem Code eine Rückführung verspricht“ als Falsch; beide Aussagen stehen zugleich im Plan, der Eintrag `ueberschreitung` mit N gleich 1400 (≥ Messung 1343) deckt die Überschreitung nach `SPEC-128` Nr. 5 d, und §6 führt die Entscheidung mit Ausgang *offen* und Adresse Planner. Der Review entscheidet nicht. | `AGENTS.md` §3.14, §3.13 | `docs/plan/planning/in-progress/slice-telemetry-einbau-mock-mutation.md` · `Die Messung liegt bei 1343 gegen die Marke 1200` und `über der Marke (1200), die der Planner beim Start gesetzt hat` | ja — `make groessenmarke` | Marke überschritten, Rückführung und Eintrag zugleich im Plan (zuständig: Planner) |
| F-7 | INFO | Die Angabe im Plan, der eine Block mit Zähler 0 sei die Zeile `len(erg.Entities) != 1` in `lesen.go`, trägt die Messung nicht: unter `go test -race -coverprofile` des Pakets `mock` (99,3 %) haben fünf Blöcke den Zähler 0, alle in unveränderten Dateien (`deuten.go` zwei, `lesen.go` einer, `schluessel.go` zwei). In `schreiben.go` und in den neuen Zeilen ist keiner; die Aussage „null Blöcke im Code dieses Slice“ gilt. | `AGENTS.md` §3.11 | `docs/plan/planning/in-progress/slice-telemetry-einbau-mock-mutation.md` · `der eine Block mit Zähler 0 ist die Zeile` | ja — Abdeckungsprofil | Messangabe im Plan zählt weniger Blöcke als die Messung |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/services/mock/schreiben.go` (Produktcode: `StartSpan` nach `holeWorkspace` und vor `holeService`, `defer span.End()`, `mutationsArt`, Verwurf des `ctx` des Spans) | geprüft, ohne Befund außer F-1; 48 eigene Änderungen, `ctx` an jeder Stelle des Pfads, Beginn und Ende, Art und Kennung |
| `src/hexagon/services/mock/mock.go` (Kommentar zu `Telemetrie`) | geprüft, ohne Befund; nennt „Query und Mutation“, wie der Code |
| `neueUmg` in `fakes_test.go` (Wächter: jeder Span endet genau einmal und trägt einen der zwei Namen) | geprüft, ohne Befund: färbt Bestandstests rot (ein zweites `End` nur bei `DELETE` rund 60 Tests; `End` nicht bei `ODR-E6003`, `ODR-E6009`, `ODR-E6010` je 11 bis 19), schwächt keinen (nur `t.Errorf` in `Cleanup`, kein Austausch von `Telemetrie` in einem Test, keine Bedingung, die einen Test überspringt); die Tests mit paniertem Fake laufen mit dem Wächter grün |
| `fakes_schreiben_test.go` (`Transaktion.Ende` nach `f`, vor Bestätigung oder Rollback) | geprüft, ohne Befund; der Eintrag liegt zwischen `f` und der Rückkehr, `End` des Dienstes danach, ein `End` in `f` stünde davor |
| `telemetrie_ablauf_test.go`, `telemetrie_gleichzeitig_test.go` (Orakel von Teil 1 gehoben: `span` 1 für die vier Mutationen, `StartSpan`, `Set`, `End` je 2, 2, 2) | geprüft, ohne Befund; die Literale folgen `SPEC-290` Nr. 6; rot gesehen durch meine drei Mutationen *Span für die Paarung* |
| `src/hexagon/services/mock/telemetrie_mutation_test.go` (M1 bis M8, M10) | geprüft, ohne Befund außer F-1, F-2; Mengen mit Zählprüfung (4, 5, 13/13/12/12, 27, 9 + 11, 33, 16 × 4) stimmen mit dem Code überein; Orakel Literale, keine Konstante des Dienstes; Fixtures nie gleich dem Wert der Mutation (Kennung `ws-gelesen` gegen den Header) |
| `src/bootstrap/telemetrie_mutation_test.go`, `telemetrie_wurzel_test.go` (M5, M7, M8, M9 durch den Root mit dem echten Adapter) | geprüft, ohne Befund außer F-3; die Zeilen der Senke stehen als Literal; kein Produktcode im Paket; keine weitere Testdatei außer den im Block genannten |
| Hybrid-Ebene (akzeptiertes Negativ (6)) | geprüft: 2 von 5 Tests führen `hybrid`, ein Mangel allein in `hybrid` ist durch die Verdrahtung des einen Adapters beider Modi (Teil 1) und die Zeilen von M1 und M9 gedeckt; Maß stimmt, nur die Überschrift im Plan nicht (F-4) |
| Akzeptiertes Negativ (5) (geteilte Stelle ohne Port dazwischen) | geprüft und nachgemessen: 20 von 20 mit Detektor, 0 von 20 ohne bei zwei Änderungen der Klasse; Maß, Grund und Folge stimmen |
| Plan: §1 bis §3, §6, Kopf, Nachzug nach `AGENTS.md` §3.9 | geprüft, ohne Befund außer F-4, F-5; §3, §6 und der Block tragen die Änderungen des Commits (Wächter, Fallzahlen, Messung, Eintrag, Risiko *Größenmarke überschritten* mit Ausgang *offen* und Adresse Planner, `AGENTS.md` §3.13) |
| Roadmap | geprüft, ohne Befund; eine Zeile *In Arbeit* für den Slice, kein Zustand mit Chronik |
| Spezifikation (`SPEC-290`, `SPEC-139` Nr. 13, `SPEC-195` Nr. 11, `SPEC-231`) | geprüft, ohne Befund; Code und Tests sagen dasselbe wie die Nummern (Beginn, Ende, Kontext, Attribute, Aufrufstellen T5 und T6, Quelltext je zweimal); der Slice schreibt keine Spezifikation |
| `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/`, `Dockerfile`, `tools/` | unverändert, ohne Befund; kein Gate gelockert |
| Geheimnis in Senke, Attribut oder `msg` (HIGH-Klasse) | geprüft, ohne Befund: 10 Stellen des Requests je Methode, Name des Ortes in vier Längen, Kennung außerhalb der Form; die Mutationen an Paaren *Stelle × Methode*, die der Plan nicht nannte, waren rot |
| Zustandsmodell, Rollback, Cursor (HIGH-Klasse) | geprüft, ohne Befund; der Slice ändert allein Telemetrie, das Ergebnis der Mutation bleibt bei Abbruch und Kontextfehler unverändert (`TestMutationAbbruchInDerTransaktion`, acht von 16 Goroutinen in `TestMutationAbbruchGleichzeitig`) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 4 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (Codes der Stellen S3 bis S5) · Kommentar nennt mehr, als der Fänger fängt · Plan trägt zwei Aussagen derselben Menge · Akzeptiertes Negativ nennt die Folge nur als Verweis · Marke überschritten, Rückführung und Eintrag zugleich im Plan · Messangabe im Plan zählt weniger Blöcke als die Messung

## Verdikt

**Merge-blockierend:** ja, durch F-1 (MEDIUM): die Menge der Codes der Stellen S3 bis S5 auf dem Pfad der Mutation hat keinen Fänger aus dem Plan. Der Produktcode (27 Zeilen) ist in jeder anderen geprüften Klasse gefangen: von 56 Einträgen sind 51 rot, die vier grünen sind eine unwirksame Änderung, zwei Arme des akzeptierten Negativs (5) und F-1. F-2 bis F-5 blockieren nicht. F-6 geht an den Planner (Entscheidung über Rückführung oder Eintrag; §6 führt sie mit Ausgang *offen*), F-7 an den Implementer.

**Übergabe:** Findings gehen an den Implementer; die **Finding-Klassen** gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ist ein **Lauf-Beleg**. Er ersetzt keine Verifikation; DoD- und Spezifikations-Konformität prüft der Verifier separat. Der Review hat die Gate-Kette nicht gefahren (`make gates` gehört dem Implementer und dem Verifier); gefahren sind allein die Läufe von `go test` im Container, die oben stehen.

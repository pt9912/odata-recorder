# Verifikations-Report: slice-bootstrap-m6-endpunkte (Endpunkte der Admin-Slices im Composition Root) — 2026-11-26

**Rolle:** Verifier (Modul 11), ein Lauf. Frage: gebaut wie in Plan und DoD zugesagt.

**Gegenstand:** Zweig `slice-bootstrap-m6-endpunkte`, HEAD `766dc2fe` (Code `5f69579d`, Review `a7dc777e`, Fix-Runde `766dc2fe`), Messbasis `ecffe236`; Zeilen A01 bis A08.

**Eingangs-Kontext:** Plan `slice-bootstrap-m6-endpunkte`, Report `2026-11-26-slice-bootstrap-m6-endpunkte-review`, `SPEC-300` (Nr. 1 bis 16 und *Ungeprüft*), `SPEC-139`, `SPEC-005`, `AGENTS.md` §3.6, §3.14, §3.15, §3.17; `LH-FA-13`, `LH-FA-18`, `LH-QA-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md). **Modell:** claude-sonnet-5-5.

---

## Verdikt

**Freigegeben mit einer Auflage.** Die Liefer-Punkte der DoD sind erfüllt, im Verhalten des Binary ist keine Abweichung gefunden. Auflage vor der Closure: V-1 (Wortlaut eines akzeptierten Negativs, kein Code). Die Closure-Punkte der DoD sind offen und liegen beim Planner.

## Findings

| ID | Kategorie | Befund | Beleg | Adresse |
|---|---|---|---|---|
| V-1 | MEDIUM | Das Negativ zur Prüfgrenze der Listen (Plan §3, *Mengen mit Achsen*) sagt als Folge, „eine Sonderbehandlung ab 10 Einträgen bliebe ungesehen“, und führt kein Maß. Ungesehen ist mehr: Länge 4 an jeder Liste und je Liste jede Länge über 3 außer der ihres Defaults. Die Prüfgrenze selbst (Länge 3) steht richtig da; der Code ist richtig. | Eigene Mutationen N5 (jede Liste mit 4 Einträgen `none`) und N6 (`MATCH_VARY_HEADERS` mit 7 Einträgen `none`): beide grün über alle `TestAdmin*`. `AGENTS.md` §3.15 (kein Fenster als Zahl, das die nächste Messung überholt) | Planner/Architect |
| V-2 | LOW | Die Handler von Exchanges und Coverage ruft kein Test in `replay`, `mock` und `hybrid` (2 Handler × 3 Modi, 0 von 6 Zellen); ihr Vertauschen bleibt grün. Der Plan (A03) beschränkt die zwei Lesewege auf `record`, führt dafür aber kein Negativ; `SPEC-300` Nr. 13 b nennt beide unter den Lesewegen. Im Code gibt es keinen Zweig je Modus für diese Handler. | Eigene Mutation N10 grün | Planner (Negativ in §3) |
| V-3 | INFO | Die Canary-Prüfung auf der Antwort der Exchanges kann nicht rot werden: die Konfiguration von `adminAufnahme` enthält keinen Wert mit `canary`. Die Zusage aus A03 ist allein an den Antworten der Diagnose geprüft. | `src/bootstrap/admin_inspektion_test.go`, `src/bootstrap/admin_imports_test.go` (gelesen) | Planner |
| V-4 | INFO | Die Zählprüfung in `TestAdminRouten` hält die Liste gegen die Summe 4+3+2+8, nicht gegen den Text von `SPEC-005`. Von Hand verglichen: gleich. | siehe *Zählprüfungen* | — |
| V-5 | INFO | Plan §3 nennt für die Bestandstests allein die Zeilen zu `Sessions()` und die Prüfung *Modus*; die geänderten Zählungen in `m3_wurzel_test.go` und `m4_wurzel_test.go` stehen nur in §6. | Diff gegen §3 | Planner |

## DoD-Konformität

| DoD-Punkt | Stand | Beleg |
|---|---|---|
| Dienste und Routen im Root (A01, A03, A06, A07, A08) | erfüllt | `src/bootstrap/wurzel_admin.go` gelesen gegen `SPEC-300` Nr. 5 a bis c, e und Nr. 6; `pruefung test` und `pruefung race` frisch grün; Mutationen N1 bis N4 und N7 bis N9 rot; die vier Zeilen zu `Sessions()` fordern „nicht nil“ |
| Ersatz der vier Wächter, `make test` grün (A02 bis A05, A07) | erfüllt | die vier Funktionen entfallen in `5f69579d`, im selben Commit stehen die Fänger in `src/bootstrap`; `pruefung test` frisch grün |
| Zwischenstand benannt (kein Admin-Listener, `ADMIN_*` abgelehnt) | erfüllt | in `src/bootstrap` ohne Tests kein `admin.New`, allein der OData-Listener wird gebunden; `nichtGelieferte` führt `ADMIN_`; `docs/user`, Katalog-Ausnahmen, Makefile, Dockerfile unverändert |
| `make gates` grün | erfüllt, nicht neu gefahren | Nachweis `.harness/state/gates-passed.head` = `766dc2fe`, `gates-passed.diffsha` gleich `working-tree-hash.sh` dieses Stands (16:07); Log des Implementers: a-check, d-check, Lint `0 issues`, test 364,8 s, race 402,1 s, contract, determinism, integration, security, zeichen. Mit dem Commit dieses Reports deckt der Nachweis den neuen HEAD nicht mehr |
| Review mit Report | erfüllt | `2026-11-26-slice-bootstrap-m6-endpunkte-review` |
| Doku-Update (dieser Teil schreibt keine Spezifikation) | erfüllt | die Nachzüge in `SPEC-300` sind Commits des Architect auf `main` |
| Handbuch/Katalog | erfüllt bis auf den Satz in §7 | `make katalog-check` Exit 0 |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Closure) | §7 ist leer. §6 führt „entschieden“ und „erledigt“; die Closure setzt je Risiko einen der drei Ausgänge. §8 verlangt den dritten Beleg zu `BEO-REPO/ausgabeformat-wechsel-zieht-tests-ausserhalb-des-blocks-mit` und die Lücke an den Architect |

## Plan-vs-Code

- Die Pfade von `5f69579d` und `766dc2fe` stehen alle im Block der Größenmarke (dazu der Plan selbst). Keine Datei außerhalb.
- Produktcode: allein `wurzel.go` und `wurzel_admin.go`. Die Fix-Runde ändert keinen Produktcode (Stat von `766dc2fe`: Plan und drei Testdateien).
- Verdrahtung wie geplant: Folge Sessions, Inspektion, Imports (nur `record`), Workspaces; `session.New` ohne Recorder mit Ersatz; `Hochfahren` und `Herunterfahren` hängen `OpenInitialSession` und `CloseAll` weiter am Modus.
- Tabelle *Zusage · Fänger* gegen den Vertrag: `SPEC-300` Nr. 5 a (A06), 5 b und e (A03), 5 c (A01, A08, A07), 6 (A03), 13 a und b (A01 bis A06, Rest V-2), `SPEC-139` Nr. 1 `Inspektion()` (A03). Nr. 5 d gehört nach Plan Teil B.

## Fix-Runde gegen die Klasse

**F-1 (Listen).** Menge: 5 Listen × {nil, leer, 1, 2, 3} = 25 Zellen; der Test führt sie als zwei Schleifen am Haken mit eigenem Anfang je Liste und zählt 25. Eigene Mutationen an Elementen, die weder Review noch Implementer nannten:

| Nr. | Mutation | Ergebnis |
|---|---|---|
| N1 | `REDACT_HEADERS`, Länge 3: letzter Eintrag fehlt | rot, `TestAdminInspektionVerdrahtet` |
| N2 | `MATCH_VARY_HEADERS`, allein nil: leerer Text | rot, ebenso |
| N3 | `VOLATILE_HEADERS`, Länge 1: Komma dahinter | rot, ebenso |
| N4 | `VOLATILE_HEADERS`, Länge 3: Reihenfolge umgekehrt | rot, ebenso |
| N5 | jede Liste, Länge 4: `none` | grün (über der Prüfgrenze): V-1 |
| N6 | `MATCH_VARY_HEADERS`, Länge 7: `none` | grün (über der Prüfgrenze): V-1 |

Innerhalb der Prüfgrenze ist die Klasse gedeckt.

**F-2 (`SQLiteBusyTimeout`).** Nach der Architect-Entscheidung umgesetzt: kein Produktcode; `SPEC-300` Nr. 6 sagt „zugesagt für die geprüfte Konfiguration“, *Ungeprüft* nennt höchstens 0 und mindestens 2147483648 ms; Plan §3 gleich; A03 führt 2147483647 ms und 999999 ns. Eigene Probe (`sqlite.Open`, `PRAGMA busy_timeout`) bestätigt den Wortlaut an beiden Grenzen:

| Konfiguration | Diagnose | Ablage |
|---|---|---|
| -1 ms | `-1ms` | 5000 |
| -500 µs | `0ms` | 5000 |
| 0 | `0ms` | 5000 |
| 1 ns; 999999 ns | `0ms` | 0 |
| 2147483647 ms; dazu 999999 ns | `2147483647ms` | 2147483647 |
| 2147483648 ms | `2147483648ms` | 0 |

**F-3, F-4** stehen als Negative in §3. **F-5:** `TestAdminSessionsAlleModi` vergleicht in `record` `session.Konfig.Record` mit `Record()`, `TestAdminBuildFehler` fängt die Panik je Fall (gelesen; Mutation nicht wiederholt).

## Weitere eigene Mutationen

| Nr. | Mutation | Ergebnis |
|---|---|---|
| N7 | Fehler von `InspektionRouten` allein in `hybrid` verschluckt (A08) | rot, `TestAdminBuildFehler` |
| N8 | `Frist` allein in `mock` aus `UpstreamTimeout` (A06) | rot, `TestAdminSessionsAlleModi`, Beobachtung (3) |
| N9 | zweiter Adapter als `Zaehler` allein in `hybrid` (A03) | rot, `TestAdminInspektionVerdrahtet` (Haken und Zähler) |
| N10 | Handler von Exchanges und Coverage außerhalb `record` vertauscht (13 b) | grün: V-2 |

Je Mutation ein Lauf `go test -count=1 -run '^TestAdmin' ./src/bootstrap/`; keine Zusage ist nebenläufig. Kein Fall war ein Baufehler.

## Entfernte Wächter (`AGENTS.md` §3.6)

Keine Lockerung. `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13 und `SPEC-298` Nr. 14 sagen heute zu, dass das Binary die Routen verdrahtet; die Wächter verboten genau das. Ersatz: Routen je Modus am `Handle` (A01), Wirkung der Handler auf Ablage und `audit.jsonl` (A02 bis A05), je Konstruktor ein Aufruf mit den Feldern des `Handle` (A07). Der zweite Gegenstand der Wächter-Kommentare (kein neuer Code, Schlüssel, Kommando) hängt an `make katalog-check`. Die übrigen Quelltext-Tests dieser vier Dateien bleiben; entfallen sind allein die vier Funktionen und ihre Importe.

## Geänderte Bestandstests

`wurzel_test.go`, `m3_wurzel_test.go`, `m4_wurzel_test.go`, `tests/integration/m4_test.go`: geändert ist allein die Aussage zu `Sessions()`. `session` fällt aus den verbotenen Konstruktoren und steht in beiden Dateien als genau 1 in der Erwartung; `pipeline`, `recorder`, `import` bleiben 0. Kein Fänger gelockert.

## Akzeptierte Negative

| Negativ | Grund | Folge | Maß |
|---|---|---|---|
| zweiter Generator an `session.Konfig.IDs` | ja | ja | 1 von 4 |
| „die eine Instanz“ am Wortlaut | ja | ja | 8 von 8 Argumenten (nachgezählt) |
| „genau einmal je `Build`“ nur je Aufrufstelle | ja | ja | 0 von 4 |
| Handler ohne Router | ja | ja | 0 von 17, 10 von 17 mit Pfadwert (nachgezählt) |
| `SQLiteBusyTimeout` als Literal | ja | ja | 1 von 4 |
| Prüfgrenze der Listen bei Länge 3 | ja | falsch | fehlt: V-1 |

Die Zahl der Mutationen des Implementers (85, 2 grün) passt zu seinen Läufen im Scratchpad (die beiden grünen sind die benannten).

## Zählprüfungen gegen die Quelle

- Routen: `SPEC-005` hat 18 Zeilen, ohne `GET /healthz` 17; die Menge ist gleich der Liste in `adminSoll` (sortierter Vergleich, kein Unterschied). 15 außerhalb `record` sind 17 ohne die zwei der Imports.
- Einstellungen: `inspection.Erlaubt()` führt 18 Schlüssel, `einstellungen` bildet 18, der Test vergleicht je Konfiguration die ganze Abbildung am Haken.
- Defaults der Listen: 5, 7, 9, 7, 7 Einträge, wie der Plan mit „5 bis 9“ sagt.

## Größenmarke

`git diff --numstat ecffe236..HEAD` über die Pfade des Blocks: 1050 Einfügungen, 138 Löschungen, **1188**; `make groessenmarke`: `Messung 1188 von Marke 1200`. Unter der Marke.

## Sensoren dieses Laufs

| Sensor | Ergebnis |
|---|---|
| `pruefung test` mit `-count=1` (Container, 4 GiB, 4 CPUs, ohne Netz) | `ok`; Last 7,3 vor, 4,2 nach; 8 min |
| `pruefung race` | `ok`; Last 2,9 vor, 3,9 nach; 10 min |
| `go test -count=1 ./...` in `tools/pruefung` | `ok` (36 s) |
| `make a-check` | 0 Befunde (Hinweis zu 97 Dateien unter `tools/` ohne Schicht) |
| `make katalog-check` | Exit 0 |
| `make groessenmarke` | Exit 0 (ohne die Sperre `flock` gefahren) |
| `make docs-check`, `make zeichen-check` | nach dem Schreiben dieses Reports |

Flakes: `TestAnfrageBodyBytes` trat in je einem Lauf von test und race nicht auf, `TestHandbuchZeitgrenzeUndWarten` in einem Lauf nicht; eine Rate ist damit nicht gemessen. Nicht gefahren: `make gates` als Ganzes, die Container-Ketten, `make lint`, `make gegenproben` (nach Auftrag). Messender: der Verifier.

## Übergabe an den Planner

V-1 vor der Closure; V-2, V-3, V-5 nach Ermessen. Offen bleiben die Closure-Punkte der DoD und der Integrationsschritt mit dem Stempel der Gegenproben.

# Review-Report: slice-bootstrap-m6-endpunkte (Endpunkte der Admin-Slices im Composition Root) — 2026-11-26

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-bootstrap-m6-endpunkte`, HEAD `c7841fa2` (Code und Plan `5f69579d`, danach der Merge von `main`), Messbasis `ecffe236`; Teil A des Schnitts von `slice-bootstrap-m6`, Zeilen A01 bis A08.

**Skill:** `.harness/skills/reviewer.md` @ `c3741274` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-26

**Eingangs-Kontext:**

- Plan `slice-bootstrap-m6-endpunkte`, alle Abschnitte.
- `SPEC-300` (Nr. 5, 6, 11, 13, 16 und *Ungeprüft*), `SPEC-139`, `SPEC-110` (Nr. 1, 2, 6), `SPEC-103` Nr. 1, `SPEC-124` (Nr. 1 und *Ungeprüft*), `SPEC-236` Nr. 7, `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13, `SPEC-298` Nr. 14, `SPEC-238` Nr. 14, `SPEC-299` Nr. 14.
- `LH-FA-13`, `LH-FA-18`, `LH-FA-09`, `LH-FA-10`, `LH-QA-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
- `AGENTS.md` §3.2, §3.6 bis §3.18.

**Größenmarke:** eigene Summe aus `git diff --numstat ecffe236..HEAD` über die Pfade des Blocks: 1022 Einfügungen, 138 Löschungen, **1160**; `make groessenmarke` Exit 0: `Messung 1160 von Marke 1200`.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Eine Liste der Diagnose mit genau einem Eintrag hat keinen Fall: die Fixtures führen 0 (nil und leer) und mehrere Einträge, die Mutation `len(eintraege) < 2` bleibt in `src/bootstrap` (ganzes Paket) und `tests/integration` grün. Unter ihr meldete die Diagnose bei `MATCH_VARY_HEADERS=Accept` (ersetzende Liste, `SPEC-110` Nr. 1) `none`; der heutige Code ist richtig. | `AGENTS.md` §3.15 (beide Seiten jeder Grenze), `SPEC-300` Nr. 6, A03 | `src/bootstrap/wurzel_admin.go` · `if len(eintraege) == 0 {`; `src/bootstrap/admin_inspektion_test.go` · `for _, leer := range [][]string{nil, {}}` | ja — `make test` mit der Mutation (heute grün) | Fänger führt nur eine Form der Eingabe |
| F-2 | MEDIUM | Das akzeptierte Negativ (Plan §3, `SPEC-300` *Ungeprüft*) nennt allein den Wert 0, und die offene Rückgabe in §6 dazu allein Werte über 2147483647 ms; gemessen weichen auch negative Werte ab (`-1ms`: Diagnose `-1ms`, Ablage 5000). Die abweichende Konfiguration von A03 erwartet `"SQLITE_BUSY_TIMEOUT": "2592000000ms"`, während ihre Ablage `busy_timeout` 0 trägt, gegen den Satz „sie trägt den wirksamen Wert“ in `SPEC-300` Nr. 6. | `AGENTS.md` §3.15 (ungefangene Klasse mit Maß), `SPEC-300` Nr. 6 | Plan §3 · „der Wert 0 an 1 der 4 Schlüssel mit Dauer“; `src/bootstrap/admin_inspektion_test.go` · `"SQLITE_BUSY_TIMEOUT": "2592000000ms"` | ja — Probe `sqlite.Open` mit `PRAGMA busy_timeout` (Tabelle unten); kein Gate | Negativ benennt die ungefangene Klasse zu eng |
| F-3 | LOW | „genau einmal je `Build`“ (A02 bis A05, `SPEC-300` Nr. 5 c) ist nur je Aufrufstelle im Quelltext gefangen: ruft `verdrahteAdmin` `adminRouten` zweimal, bleiben `src/bootstrap` und `tests/integration` grün. Ein akzeptiertes Negativ dafür führt der Plan nicht; eine beobachtbare Wirkung hat der zweite Aufruf heute nicht. | `AGENTS.md` §3.15 | `src/bootstrap/admin_konstruktoren_test.go` · `if nennungen[name] != 1 \|\| len(aufrufe[name]) != 1` | ja — `make test` mit der Mutation (heute grün) | Zusage je Lauf nur am Quelltext gefangen |
| F-4 | LOW | Das zweite und dritte akzeptierte Negativ des Implementers tragen kein Maß (n von m); das zweite zählt unter „Folge“ ungefangene Formen auf und nennt keine Folge im Betrieb, das dritte verweist für beides auf `SPEC-300`. | `AGENTS.md` §3.15 | Plan §3 · „fängt `TestAdminKonstruktorenEinmal` am Wortlaut der Argumente“ und „Die Tests rufen die Handler ohne den Router“ | nein | Akzeptiertes Negativ ohne Maß |
| F-5 | INFO | Reicht der Root im Modus `record` den Ersatz statt des Recorders an `session.Konfig`, wird allein `TestAdminBuildFehler` rot, durch eine Panik im Zusatzfall `NeuerRecorder nil`, die den Testprozess abbricht. Ohne diesen Test fängt `TestHerunterfahrenOhneAbbruchDesAusloesers` (`durchstich_test.go`, Bestand) die Verdrahtung (ein Lauf); A06 überspringt `record` (`continue`). | Reviewer-Skill (Hinweis zu `panic:`), `SPEC-139` Nr. 3 | `src/bootstrap/admin_konstruktoren_test.go` · `fall{"NeuerRecorder nil", modusRecord` | ja — `make test` mit der Mutation | Panik verdeckt den benannten Fänger |

## Schwerpunkte ohne Finding

**Ersatz der vier Wächter.** Vom Plan gedeckt (§1 *Übernimmt*, DoD zweiter Punkt, §3 A02 bis A05 und A07) und vom Vertrag verlangt: `SPEC-292` Nr. 14, `SPEC-293` Nr. 12, `SPEC-297` Nr. 13 und `SPEC-298` Nr. 14 sagen heute das Gegenteil der Wächter zu. Keine Lockerung nach `AGENTS.md` §3.6. Die Klasse „keine neuen Codes, Schlüssel, Kommandos“ hält `make katalog-check`; Routen, Modus, Ports und Einträge fängt der Ersatz am Verhalten (Mutationen unten). A07 zählt `admin.New` nicht; in `src/bootstrap` steht heute kein Aufruf, die Verdrahtung des Listeners führt `slice-bootstrap-m6-listener`.

**Geänderte Bestandstests** (`wurzel_test.go`, `m3_wurzel_test.go`, `m4_wurzel_test.go`, `tests/integration/m4_test.go`). Allein die Aussage zu `Sessions()` ändert sich; wo `session` aus der Liste der verbotenen Konstruktoren fällt, steht seine Zahl als genau 1 in der Erwartung. Kein Fänger gelockert.

**Binary und Geheimnisse.** `Build` baut zusätzlich Dienste und Routen; kein Listener, kein `admin.New`, `Hochfahren` und `Herunterfahren` hängen weiter am Modus. Die 18 Einstellungen stammen aus typisierten Feldern, keine Adresse, kein Pfad, keine URL; der Haken vergleicht die Abbildung als Ganzes, ein zusätzlicher Schlüssel fällt auf.

**Negativ Generator-Identität.** Grund, Folge und Maß stimmen: `ZufallsIDs` ist eine leere Struktur, `Systemuhr` nicht.

**Randformen-Raster.** Jede Zelle nennt einen Ort; im Code ist keine Randform ohne Ort entschieden.

**Offene Rückgabe `SQLiteBusyTimeout` (F-2), Einschätzung ohne Entscheidung.** Die Wirkung ist gering: der Lader lässt allein 1 ms bis 2147483647 ms zu (`SPEC-110`), kein Weg über das Binary erreicht die Abweichung, nur ein von Hand gebautes Literal. Dort ist sie nicht harmlos (`busy_timeout` 0 bei angezeigten 30 Tagen). Das MEDIUM gilt dem Wortlaut des Negativs und dem Fall, der die Abweichung als Soll führt, nicht einem Fehler im Binary. Wortlaut oder Verhalten entscheidet der Architect.

## Eigene Mutationen

Treiber des Reviewers im Scratchpad (außerhalb des Repos), je Mutation eine Ersetzung in einer Wegwerfkopie im Container `odata-recorder:test` ohne Netz und Mount (`--memory=4g --memory-swap=4g --cpus=4`), `go test -count=1 -trimpath`; je Mutation ein Lauf, keine Zusage ist nebenläufig.

| Zeile | Mutation | Ergebnis |
|---|---|---|
| A01 | im Modus `hybrid` fehlt `restore` | rot: `TestAdminRouten`, `TestAdminWorkspacesVerdrahtet` |
| A02 | im Modus `mock` fehlt `DELETE` der Sessions | rot: `TestAdminRouten`, `TestAdminSessionsVerdrahtet` |
| A03 | Liste mit einem Eintrag wird `none` | **grün** (ganzes Paket und `tests/integration`): F-1 |
| A04 | `GET` der Imports fehlt | rot: `TestAdminRouten`, `TestAdminImportsVerdrahtet` |
| A05 | im Modus `replay` sind die Handler von `reset` und `restore` vertauscht | rot: `TestAdminWorkspacesVerdrahtet` |
| A06 | im Modus `record` erhält `session.Konfig` den Ersatz | rot: Panik in `TestAdminBuildFehler`; ohne ihn `TestHerunterfahrenOhneAbbruchDesAusloesers`: F-5 |
| A07 | `adminRouten` zweimal je `Build` | **grün** (ganzes Paket und `tests/integration`): F-3 |
| A08 | Fehler von `baueSessions` an der Aufrufstelle in `record` verschluckt | rot: `TestAdminBuildFehler` an der Aussage zu `NeueSession scheitert/record` (danach Panik im Zusatzfall) |

Probe zu F-2 (`sqlite.Open`, dann `PRAGMA busy_timeout`; die Diagnose zeigt jeweils den Wert der linken Spalte mit `ms`):

| Konfiguration in ms | `busy_timeout` der Ablage |
|---|---|
| -1 | 5000 |
| 0 | 5000 |
| 1 | 1 |
| 2147483647 | 2147483647 |
| 2147483648 | 0 |
| 2592000000 | 0 |
| 4294967297 | 0 |

## Sensoren dieses Laufs

Container nacheinander unter `flock /tmp/odata-recorder-heavy.lock`, Messender: der Reviewer. `pruefung test` über eine Kopie des Baums mit `-count=1`: `ok` (Last 3,0 vor, 3,8 nach). `pruefung race`: `ok` nach 606 s (Last 3,2 vor). `make a-check`: 0 Befunde. Lint (`tools/harness/lint.sh` im Container `odata-recorder:lint` über die Pfade der Stufe `lint`): `0 issues`. `make groessenmarke`: Exit 0. Die Flakes `TestHandbuchZeitgrenzeUndWarten` und `TestAnfrageBodyBytes` traten in je einem Lauf nicht auf. `make docs-check` und `make zeichen-check` laufen nach dem Schreiben dieses Reports. Nicht gefahren: `make gates` als Ganzes, `make test-integration` und die übrigen Container-Ketten, `make gegenproben` (nach Auftrag).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/` (Code) | geprüft; außer F-1 ohne Befund |
| `src/bootstrap/` (Tests) | geprüft; F-1 bis F-3, F-5 |
| `src/adapters/driving/admin/` | geprüft, ohne Befund |
| `tests/integration/` | geprüft, ohne Befund |
| Plan §1, §3, §4, §6, §8 | geprüft; F-2, F-4 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 2 |
| INFO | 1 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe · Negativ benennt die ungefangene Klasse zu eng · Zusage je Lauf nur am Quelltext gefangen · Akzeptiertes Negativ ohne Maß · Panik verdeckt den benannten Fänger

## Verdikt

**Merge-blockierend:** ja — F-1 und F-2 (MEDIUM). Im Verhalten des Binary ist kein Fehler gefunden.

**Übergabe:** F-1, F-3, F-4, F-5 an den Implementer; F-2 an den Architect (offene Rückgabe in §6 des Plans). Dies ist der eine Review-Lauf; der Report ersetzt keine Verifikation.

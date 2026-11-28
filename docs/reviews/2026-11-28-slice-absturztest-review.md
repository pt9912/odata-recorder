# Review-Report: slice-absturztest (Absturz-Injektion in Mutation, Reset, Restore und Snapshot) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** `git diff dac6775a..5a67a1e1` über `src/bootstrap/absturz_test.go` und die Zeile in `spec/spezifikation.md`; der eigene Commit ist `a74c19a5` (Plan, Spec-Zeile, Testdatei, kein Produktcode). Die Änderungen aus `main` sind nicht Gegenstand.

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`5a67a1e1`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-absturztest` (§1, §3 mit T1 bis T17, N1 bis N4, Mutationen, Raten, Negative, §4, §6); `SPEC-048` samt Nr. 4; `SPEC-049` Nr. 1, 7, 9; `LH-QA-03`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md); `AGENTS.md` §3.

**Urteil: freigegeben.** Kein HIGH, kein MEDIUM. Der Test beendet den Kindprozess mit `SIGKILL`, erreicht alle 10 Haltepunkte und vergleicht den Stand nach dem Neustart gegen die Orakel.

**Vorgehen.** Wegwerfkopie (`git archive HEAD`), schreibgeschützt in einen Container aus dem Image `odata-recorder:race` eingehängt und darin kopiert (`--network none --memory=4g --memory-swap=4g --cpus=4`), unter `flock /tmp/odata-recorder-heavy.lock`, ein Container. Beim Start 14 GiB verfügbar, Last 5,4; am Ende Last 3,7. Messend: Reviewer.

| Lauf | Ergebnis |
|---|---|
| Paket `bootstrap`, `go test -race -count=1` | grün, 404,3 s (Last 5,40 vorher, 3,08 nachher); `TestAbsturz` 8,17 s, `TestAbsturzEndeOhneHaltepunkt` 1,46 s, `TestAbsturzPruefungErkennt` 5,03 s |
| `^TestAbsturz`, ohne Detektor, `-count=20` | grün, 211 s (Last 3,08 / 2,98) |
| `^TestAbsturz`, `-race -count=5` | grün, 73 s (Last 2,98 / 2,99) |
| `/tmp` und Prozesse nach diesen drei Läufen | keine Reste außer Kopie und Log des Skripts; kein Prozess übrig |

Nicht gefahren: `make gates`, `make gegenproben`, `make lint`.

**Eigene Mutationen** (je eine Wegwerfkopie, `-run '^TestAbsturz' -count=1`, ohne Detektor):

| Nr. | Mutation | Ergebnis |
|---|---|---|
| M1 | Adapter ohne Transaktion (`BEGIN IMMEDIATE` und `COMMIT` ersetzt) | rot: die drei `…-vor-commit` (Revision, Einträge, Cursors) |
| M2 | `ResetAll` der Cursors in zweiter Transaktion danach | rot: `reset-nach-commit` (Cursors) |
| M3 | `ResetAll` der Cursors in eigener Transaktion davor | rot: `reset-vor-commit` (Lage), `reset-nach-commit` |
| M4 | temporäre Datei des Writers im Verzeichnis der Blobs | rot: `snapshot-im-write` (Lage, Ablage) |
| M5 | `CreateWorkspace` zwischen `Write` und `Commit` | rot: `snapshot-nach-writer-commit`, `snapshot-nach-create`; `snapshot-im-write` grün (F-4) |
| M6 | `SIGTERM` statt `SIGKILL` | rot: 10 von 10, „ohne SIGKILL“ |
| M7 | `panic` statt Signal (Defer laufen) | rot: 10 von 10, „exit status 2 ohne SIGKILL“ |
| M8 | Prüfung des Signals in `absturzStarte` entfernt | **grün** (F-1) |
| M9 | `Startup` in `absturzPruefe` entfernt | rot: `snapshot-nach-writer-commit`, 4 Fälle von `TestAbsturzPruefungErkennt` |
| M10 | Bedingung auf das Verzeichnis `sha256` in `absturzAblage` entfernt | **grün** (F-7) |
| M11 | Zweig „der Snapshot fehlt“ abgeschaltet | **grün** (F-7) |

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | LOW | Das akzeptierte Negativ nennt für den Zweig *anderes Ende* als Grund „nicht ohne Eingriff in den Kindprozess herstellbar“. Nach dem Lesen des Codes (nicht gefahren) endet der Kindprozess bei einer unbekannten Operation, einem Argument des Tests, mit Exit 1 und liefe in diesen Zweig; ohne die Prüfung des Signals bleibt die Suite grün (M8). | `AGENTS.md` §3.10, §3.15 | Plan §3 · `nicht ohne Eingriff in den Kindprozess herstellbar` | ja, `make test` bliebe mit M8 grün | Akzeptiertes Negativ mit unzutreffendem Grund |
| F-2 | LOW | Die Ausgabe des Kindprozesses wird allein in Fehlermeldungen gezeigt, und `SIGKILL` kommt dem Exit-Code des Detektors zuvor: eine Race-Meldung des Kindprozesses macht `make test-race` nicht rot, und der Plan führt das nicht als Negativ. Dieselbe Operation läuft ohne Haltepunkt im Orakel des Elternprozesses; ungesehen bliebe ein Rennen, das nur im Kindprozess auftritt. | [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md); `AGENTS.md` §3.15 | `src/bootstrap/absturz_test.go` · `ausgabe, err := cmd.CombinedOutput()` | nein | Ausgabe des Kindprozesses auf dem Gutpfad nicht gewertet |
| F-3 | LOW | Im Raster steht in der Zelle *Neustart (`Startup`)* mal *leer* „gilt nicht“ ohne Grund. | `AGENTS.md` §3.16 | Plan §3 · `fehlender Blob: rot, Nr. 3 \| gilt nicht \|` | nein | Rasterzelle „gilt nicht“ ohne Grund |
| F-4 | INFO | `SPEC-048` Nr. 4 sagt für *`Create` vor dem `Commit` des Writers* rot am Haltepunkt *im `Write`*. Das trifft für `Create` vor `Write`; zwischen `Write` und `Commit` bleibt dieser Haltepunkt grün (M5), rot werden zwei andere. Der Plan nennt das in Zeile T9 seiner Mutationen; §6 führt keine Zeile an den Architect. Zuständig: Architect. | `SPEC-048` Nr. 4; `AGENTS.md` §3.13 | `spec/spezifikation.md` · `Fänger der Fänger` | ja, Lauf mit M5 | Vertrag nennt einen Fänger-Ort, der nur eine Form der Mutation trifft |
| F-5 | INFO | Das Paket `bootstrap` braucht unter `-race` auf dem geprüften Stand (mit dem Merge aus `main`) 404 s bei 600 s Grenze je Paket, in einem Lauf mit 4 CPUs; die Tests dieses Slice haben daran rund 15 s. §6 führt das Risiko; seine Zahl 405 s stammt von einer früheren Fassung. | Maintainability | Plan §6 · `brauchte in einem Lauf 405 s` | ja, `make test-race` | Paketlaufzeit unter `-race` nähert sich der Zeitgrenze |
| F-6 | INFO | Ein Lauf eines einzelnen Untertests (`-run '^TestAbsturz$/^mutation-vor-commit$'`) ist ohne Mutation rot („1 Haltepunkte gefahren, erwartet 10“). Ein Mutations-Beleg mit diesem Filter wäre nicht aus dem benannten Grund rot. | `SPEC-048` Nr. 2 | `src/bootstrap/absturz_test.go` · `if len(gefahren) != 10` | ja, Lauf mit Filter | Zählprüfung rot bei Filter auf einen Untertest |
| F-7 | INFO | Zwei Zweige der Prüfung haben keinen Fänger und stehen nicht unter den Negativen (M10, M11). Beide sind doppelt gedeckt: ein temporärer Name gleicht keinem Hash, und ein fehlender Snapshot `vor` fällt im Vergleich der Listen auf. | `AGENTS.md` §3.15 | `src/bootstrap/absturz_test.go` · `filepath.Dir(rel) != "sha256"`, `if !vorGesehen` | ja, `make test` bliebe grün | Zweig der Prüfung doppelt gedeckt, ohne eigenen Fänger |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Ende des Kindprozesses (`SIGKILL`, kein Defer, Marke) | geprüft, ohne Befund außer F-1: nach dem Signal läuft kein Code; ein anderes Signal und ein `panic` sind rot (M6, M7) |
| Haltepunkte (10 von 10 erreicht, Lage je Haltepunkt, Dekoratoren gegen Dienste und Adapter gelesen) | geprüft, ohne Befund: keiner, der nie trifft; die Zahl steht als Literal |
| „Stand davor oder danach“: Einträge, Revision, Cursors und Snapshot-Liste beider Workspaces, Blobs, `tmp`, `Startup` | geprüft, ohne Befund: Umfang wie `SPEC-048` Nr. 3; Orakel aus demselben Aufbau ohne Absturz |
| Prozesse, Zombies, Verzeichnisse | geprüft, ohne Befund: der Elternprozess wartet auf den Kindprozess; dieser legt kein eigenes Verzeichnis an |
| Zeitgrenzen (Frist 2 min, `-test.timeout` 100 s, `WaitDelay` 10 s) | geprüft, ohne Befund |
| Flakiness | geprüft, ohne Befund in den Läufen oben; nicht unter hoher Last des Hosts gemessen |
| `src/bootstrap/absturz_test.go` (Kommentare, `//nolint`, `t.Skip`, Schlaf, Haken im Produkt) | geprüft, ohne Befund |
| `spec/spezifikation.md` (Vermerk *Stand* von `SPEC-048`; Straten) | geprüft, ohne Befund außer F-4 am unveränderten Text von Nr. 4 |
| [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); Produktcode, Gates, `export_test.go` | geprüft, ohne Befund: im eigenen Commit nicht berührt, keine neue Abhängigkeit |
| `docs/plan/planning/` (Plan §1, §3 Tabelle und Negative N1 bis N4, §4) | geprüft, ohne Befund außer F-1, F-3, F-5 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 3 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Akzeptiertes Negativ mit unzutreffendem Grund · Ausgabe des Kindprozesses auf dem Gutpfad nicht gewertet · Rasterzelle „gilt nicht“ ohne Grund · Vertrag nennt einen Fänger-Ort, der nur eine Form der Mutation trifft · Paketlaufzeit unter `-race` nähert sich der Zeitgrenze · Zählprüfung rot bei Filter auf einen Untertest · Zweig der Prüfung doppelt gedeckt, ohne eigenen Fänger

## Verdikt

**Merge-blockierend:** nein.

**Übergabe:** F-1 bis F-3, F-6 und F-7 an den Implementer, F-4 an den Architect; die Finding-Klassen in die Slice-Closure §7. Die DoD und die Tabelle *Zusage · Fänger* gegen den Vertrag prüft der Verifier.

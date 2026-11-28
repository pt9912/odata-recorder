# Verifikation slice-absturztest (Absturz-Injektion in Mutation, Reset, Restore und Snapshot) — 2026-11-28

Verifier, frischer Kontext, genau eine Verifikation. **Modell:** Sonnet 5.5.
Stand HEAD `eaeec878`, Zweig und Worktree `slice-absturztest`, vor der Prüfung `git status` leer; Basis der Marke `dac6775a`.
Gegen: DoD des Plans (§2), Tabelle *Zusage · Fänger* (T1 bis T17, N1 bis N4), §3, §6; `SPEC-048`, `SPEC-049` Nr. 10 c; `LH-QA-03` samt `LH-QA-03.a`; `2026-11-28-slice-absturztest-review`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md).

## Urteil

**DoD-Liefer-Punkte 1 bis 3, Gates und Review: bestätigt.** Kein HIGH, kein MEDIUM, ein LOW, drei INFO. Der Teil *Fehler-Injektion* der Messmethode von `LH-QA-03` ist belegt, in den Grenzen unten. Nichts blockiert den Merge.

## Auflagen für die Closure

1. **V-1:** den ungefangenen Teil in §3 als Negativ nennen oder um einen Schadensfall schließen.
2. **V-2:** die Frage an den Architect in §6 mit Ausgang und Adresse eintragen (`AGENTS.md` §3.13; der Verifier ändert den Plan nicht).
3. **V-4:** die üblichen Closure-Punkte.

## Sensoren und Belege

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` | **nicht neu gefahren** (Auftrag): der Nachweis passt | `gates-passed.diffsha` gleich der Ausgabe von `tools/harness/working-tree-hash.sh` am sauberen HEAD vor diesem Report (`0476a9fc…`), geschrieben 06:44. `gates-passed.head` nennt den Vorgänger `d43c94db`: der Lauf lag vor dem Commit `eaeec878` über demselben Inhalt. Der Nachweis sagt nicht, welche Stufen aus dem Schicht-Cache kamen |
| `^TestAbsturz` ohne Untertest-Filter, `-race -count=3 -v` | grün, 50,7 s; 30 von 30 Haltepunkt-Untertests PASS, kein FAIL, kein SKIP, keine Race-Meldung | eigener Lauf |
| `^TestAbsturz` ohne Detektor, `-count=10` | grün, 120,4 s | eigener Lauf |
| Reste nach den Läufen | keine Testverzeichnisse in `/tmp`, kein Prozess `bootstrap.test` | eigener Lauf |
| Marke: `git diff --numstat` über den Pfad des Blocks | 745 von 3100 | nachgerechnet |
| `t.Skip`, `time.Sleep`, `nolint`, `testing.Short` in der Testdatei | kein Treffer | eigene Suche |
| `make gegenproben` | nicht gefahren (Auftrag) | — |

Lauf: ein Container aus dem Image `odata-recorder:race` (Toolchain und Module; Quellstand aus `git archive HEAD`, schreibgeschützt eingehängt), ohne Netz, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`; nach dem Warten 14 GiB verfügbar, Last 6,22, am Ende 4,70.

## Eigene Mutationen

Je eine Wegwerfkopie, ein Lauf `go test -run '^TestAbsturz' -count=1`.

| Nr. | Mutation | Ergebnis |
|---|---|---|
| m01 | `SPEC-048` Nr. 4: Transaktion des Restore geteilt | rot: `restore-nach-commit` (Revision), `restore-vor-commit` (Lage) |
| m02 | Nr. 4: Eintrag des Mock über den direkten Port außerhalb der Transaktion | rot: `mutation-vor-commit` (Eintrag) |
| m03 | Eintrag des Mock erst nach `COMMIT` | rot: `mutation-nach-commit` |
| m04 | Nr. 4: `CreateWorkspace` vor `Write` | rot: `snapshot-im-write` (`BlobsFehlend`), dazu `nach-writer-commit`, `nach-create` |
| m11 | Restore erhöht die Revision nach `COMMIT` ein zweites Mal über den direkten Port | rot: beide `restore-*` |
| m05 | Review F-1: Prüfung des Signals entfernt | rot: Exit 1, 9, 137, `SIGTERM`, `SIGUSR1`, unbekannte Operation |
| m06 | Review F-2: Datenrennen im Haltepunkt (nur der Kindprozess), `-race` | rot: 10 von 10, *Datenrennen im Kindprozess* |
| m07 | dasselbe Rennen ohne die Prüfung der Meldung, `-race` | `TestAbsturz` bleibt grün: allein die Prüfung macht m06 rot (rot sind hier nur die drei Enden *Race-Meldung*) |
| m08 | Review F-2: nur stdout gewertet | rot: *Race-Meldung auf stderr*, *bei Exit 66* |
| m09 | T11: ein Haltepunkt durch das Doppel eines anderen ersetzt | rot: 9 statt 10 |
| m10 | Fehler von `Restore` eines anderen Snapshots als `vor` verschluckt | **grün** (V-1) |

m01, m02 und m04 bestätigen DoD-Punkt 2 selbst. Die 31 Mutationen und die Raten des Implementers sind darüber hinaus nicht nachgefahren.

## DoD

| Punkt | Stand |
|---|---|
| Absturztest nach `SPEC-048` Nr. 1 bis 3, 10 Haltepunkte, in `make test` und `make test-race` | bestätigt: T1 bis T17 gegen den Vertrag gelesen; Kindprozess, Port-Dekoratoren, `SIGKILL`, Neustart mit `Startup`, Orakel aus Aufbau ohne Absturz, Zahl 10 als Literal. `pruefung test` und `pruefung race` fahren alle Pakete |
| Mutationen nach Nr. 4 je rot am benannten Haltepunkt | bestätigt (m01, m02, m04; Plan §3) |
| Vermerk *Stand* im liefernden Commit | bestätigt: `a74c19a5` |
| `make gates` | bestätigt über den Nachweis |
| Review | liegt vor, freigegeben |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Planner) |

## Fix-Runde gegen die Klassen des Reviews

| Befund | Stand |
|---|---|
| F-1 (Negativ mit unzutreffendem Grund) | gedeckt: 6 Zweige, 16 Enden nachgezählt; m05 rot |
| F-2 (Ausgabe des Kindprozesses) | gedeckt: m06 an einer Stelle, die der Befund nicht nannte, m07 als Gegenkontrolle, m08 |
| F-3 (Rasterzelle ohne Grund) | gedeckt: jede Zelle *gilt nicht* des Rasters trägt einen Grund |
| F-4 (Wortlaut Nr. 4) | an den Architect, §6 *offen*; der Wortlaut trifft für `Create` vor `Write` (m04) |
| F-5, F-6 | im Plan nachgetragen |
| F-7 (Zweige ohne Fänger) | die zwei genannten stehen als Negativ; die Klasse ist nicht vollständig (V-1) |

## Plan gegen Code

Die zwei Commits des Slice mit Code (`a74c19a5`, `eaeec878`) berühren `src/bootstrap/absturz_test.go`, den Plan und eine Zeile in `spec/spezifikation.md`: das ist §3. Kein Produktcode, kein Export, der Aufbau der Abnahme bleibt unverändert. `SPEC-049` Nr. 10 c (*ihn trägt der Absturztest*) trifft jetzt zu.

## Was `LH-QA-03` hier belegt und was nicht

Das Lastenheft sagt zu: keine fremden Zustandsänderungen oder Cursors bei parallelen Tests; abgeschlossene Mutationen und Snapshots bleiben nach **regulärem** Neustart konsistent. Messmethode: Paralleltest mit mindestens 20 Workspaces, Neustart- und Fehler-Injektionstests.

**Belegt durch diesen Slice:** ein Prozessende mit `SIGKILL` an 10 Port-Grenzen um `PATCH`, Reset, Restore und Snapshot hinterlässt Entity-State, Revision, Cursors und Snapshot-Liste ganz vorher oder ganz nachher; der zweite Workspace bleibt unverändert; die Ablage ist nach `Startup` bereinigt; jeder gelistete Snapshot lässt sich wiederherstellen und eine weitere Mutation gelingt.

**Nicht durch diesen Slice:** der Paralleltest (`TestAbnahmeIsolation`, 20 Workspaces) und der reguläre Neustart (`TestAbnahmeNeustart` im Prozess, `SPEC-049` Nr. 6 und 7). Beide sind hier nicht einzeln gefahren; sie liegen in der Gate-Kette.

**Ungeprüft bleibt:**

- Ende zwischen fsync und Rename und während `COMMIT` (N1): hier tragen Rename und die Transaktion von SQLite.
- Stromausfall und Absturz des Betriebssystems (N2): die Dauerhaftigkeit über einen Ausfall des Hosts ist nicht belegt. Der Wortlaut der Anforderung verlangt den regulären Neustart.
- Das gebaute Binary samt Aufruf der Startwartung (N4); für Snapshots über die Prozessgrenze des Binary gilt weiter `SPEC-049` Nr. 10 b.
- Anhängen eines Exchange und Import (N3).
- Der Inhalt des wiederhergestellten Standes: geprüft ist, dass `Restore` gelingt (so `SPEC-048` Nr. 3), nicht was es herstellt.
- Absturz bei gleichzeitigen Operationen: je Haltepunkt läuft genau eine.
- Die Wege in V-2.

## Findings

| ID | Kategorie | Befund | Quelle | Pfad |
|---|---|---|---|---|
| V-1 | LOW | Die Wiederherstellbarkeit des **neuen** Snapshots (*nach Create*, „wiederherstellbar“) hat im Testcode keinen eigenen Fänger: wird der Fehler jedes `Restore` außer `vor` verschluckt, bleibt alles grün (m10). §3 nennt zwei doppelt gedeckte Teile ohne diesen dritten. Die Prüfung selbst läuft, und der Plan hat sie bei T10 rot gesehen; ohne Folge für das Produkt | `SPEC-048` Nr. 2; `AGENTS.md` §3.15, §3.17 | `src/bootstrap/absturz_test.go` · `absturzWiederherstellen` |
| V-2 | INFO | `SPEC-048` Nr. 5 nennt als ungeprüft nicht: `POST`, `PUT`, `DELETE` des Mock, Anlegen und Löschen eines Workspace, das Fortschreiben eines Replay-Cursors beim Absturz. Für die drei Methoden trägt der Code: alle laufen wie `PATCH` durch dieselbe eine Transaktion (`mock/schreiben.go`, gelesen, nicht gefahren). Für die übrigen Wege ist nichts belegt. Zuständig: Architect | `LH-QA-03`; `SPEC-048` Nr. 5 | `spec/spezifikation.md` · `SPEC-048` |
| V-3 | INFO | Das Paket `bootstrap` brauchte unter `-race` laut §6 und Review 404 bis 410 s bei 600 s je Paket (hier nicht neu gemessen); `^TestAbsturz` braucht je Lauf rund 17 s. §6 führt das Risiko | Maintainability | Plan §6 |
| V-4 | INFO | Vor `done/`: DoD-Häkchen, §7, Register, je Risiko genau ein Ausgang (der Punkt F-4 steht auf *offen*, keiner der drei Werte), Paarungen, Roadmap (M7 nennt diesen Slice als Bedingung) | Plan §2, §5, §6 | `docs/plan/planning/` |

## Übergabe

An den Planner: Auflagen oben; V-2 und F-4 an den Architect.

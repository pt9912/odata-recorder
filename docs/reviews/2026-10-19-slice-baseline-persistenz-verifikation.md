# Verifikations-Report: slice-baseline-persistenz — 2026-10-19

**Art:** Verifikation (Modul 11) — gegen DoD, Closure-Trigger, Spezifikation und Entscheidungen. Nicht Review, nicht Validierung.

**Gegenstand:** [slice-baseline-persistenz](../plan/planning/done/slice-baseline-persistenz.md), Diff `b70691f..4576df8` (Umsetzung `561450a`, Behebung der Review-Befunde `4576df8`).
Bezug: [LH-FA-10](../../spec/lastenheft.md), [LH-FA-11](../../spec/lastenheft.md), [LH-FA-18](../../spec/lastenheft.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); Vertrag `SPEC-107`.
Vorgänger: [Review-Report](2026-10-19-slice-baseline-persistenz-review.md) (0 HIGH, 7 MEDIUM, 2 LOW, 3 INFO).

**Rolle:** Verifier · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

## Urteil: BESTANDEN MIT AUFLAGEN

Der Code erfüllt `SPEC-107` Kopf und Nr. 1 bis 13, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) (Treiber nur im Adapter, `make a-check` und `make lint` grün) und die Schichtregeln; `make gates` ist grün.
Die Review-Befunde F-1 bis F-4, F-8, F-9 und F-12 sind behoben, ihre Mutationen werden aus dem benannten Grund rot (selbst nachgefahren). F-5, F-6, F-10 sind in `SPEC-107` entschieden oder unter *Ungeprüft* ehrlich geführt; F-7 und F-11 bleiben offen, sind aber im Plan benannt.
Kein HIGH. Ein MEDIUM: Die mit F-5 neu in `SPEC-107` Nr. 8 aufgenommene Randform „leere Quell-Session ist `ErrNichtGefunden`“ hat keinen Fänger. Dazu zwei LOW und zwei INFO.
Der Produktcode ändert sich dafür nicht; es fehlt ein Testfall.

**Anzahl:** HIGH 0 · MEDIUM 1 · LOW 2 · INFO 2.

## Läufe (Verifier selbst)

| Lauf | Ergebnis |
|---|---|
| `make gates` (HEAD, Arbeitsbaum sauber) | Exit 0: a-check, d-check, baseline-verify, build, test (`gofmt`, `go vet`, Tests), lint, schema-check (alle Fälle `ok`), migration-check `ok`; `git status` danach sauber |
| Mutationen | 53 im Wegwerf-Klon (`git archive HEAD` unter `/tmp/claude-1000/-Development-odata-recorder/4e17703b-9e29-4182-bc79-0efc3533bb9c/scratchpad/vw`), Container `golang:1.27-alpine@sha256:8a5910f3…`, mit vorab geladenem Modul-Cache (der Container lief ohne `--network none`, die Mutationen brauchten kein Netz), `CGO_ENABLED=0`; je Mutation `gofmt -l` und `go vet` vor dem Test, Datei danach zurückgesetzt; ein Fehlbau (ungenutzter Import) verworfen und neu gefasst; Arbeitsbaum unberührt |
| `-race` | nicht verfügbar (`CGO_ENABLED=0`, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)); Gleichzeitigkeit stützt sich auf `gleichzeitige Publish mit gleichem Hash ordnen sich` und die Mutation `BEGIN IMMEDIATE` zu `BEGIN` (rot) |
| Umgebungshinweis | Im Klon mit `--user 1000` scheitert `TestRechteUnabhaengigVonUmask` (Paket `blobs`, nicht Gegenstand dieses Slice) an `mkdir … permission denied`; unter `make test` grün. Kein Befund dieses Slice |

## DoD und Closure-Trigger

| Punkt | Befund |
|---|---|
| 1 Modell und Ports für Service, Baseline-Snapshots, Import, Quellen, Coverage; Adapter auf `SPEC-004` | erfüllt: `src/hexagon/model/baseline.go`, `ports/outbound/{baseline,metadata}_repository.go`, `adapters/driven/sqlite/{baselines,imports,metadata}.go` |
| 2 Publikation in einer Transaktion, idempotent bei gleichem Anfrage-Hash | erfüllt: `Publish` in `BEGIN IMMEDIATE`, Hash-Abfrage vor dem Schreiben; Atomarität und Idempotenz je Mutation rot (siehe unten) |
| 3 Vertragstests am Port gegen SQLite einschließlich der Randformen aus §6 | erfüllt mit einer Lücke (V-1) |
| 4 `make gates` grün | erfüllt |
| Closure-Trigger (§5): Gates, eine Mutation je Zusage rot, Closure-Notiz | Gates ja; Mutation je Zusage ja, außer V-1; DoD-Häkchen, Report-Haken, Closure-Notiz (§7 „Wird bei der Closure geschrieben“), Register, drei Paarungen stehen erwartungsgemäß offen |

## Plan-vs-Code-Diff (`b70691f..4576df8`)

| Plan §3 | Code | Abweichung |
|---|---|---|
| `src/hexagon/model/` | `baseline.go`, `model_test.go` | keine |
| `src/hexagon/ports/outbound/` | `baseline_repository.go`, `metadata_repository.go` (vier Ports) | keine |
| `outboundtest/` | `baseline_vertrag.go`, `service_vertrag.go`, `vertrag.go` (zwei Hooks) | keine |
| `src/adapters/driven/sqlite/` | `baselines.go`, `imports.go`, `metadata.go`, zwei Testdateien | keine |
| `spec/spezifikation.md` | `SPEC-107` vor dem Code (`a0d3f09`), danach geschärft (`4576df8`: Nr. 3, 6, 7, 8, 13, *Ungeprüft*) | keine |
| nicht im Plan | Review-Report; Plan selbst | harmlos |

§1-Ausschlüsse gewahrt: keine Import-Logik, keine Workspaces, kein Löschen (kein Lösch-Port, `SPEC-107` *Ungeprüft*). `.golangci.yml`, `.a-check.yml`, `Dockerfile`, `go.mod`, `Makefile` unberührt; kein `//nolint` (`AGENTS.md` §3.2, §3.6 gewahrt). Adresse in §1 (`slice-workspace-state-persistenz`, DoD Zeile 60 nennt „Workspace-Snapshots“, §3.13) nimmt den Punkt an.
Marke §4/§3.14: gemessen mit `git diff --numstat b70691f..4576df8` über die vier Pfade aus §3 ohne die Spezifikation 1551 (stimmt mit dem Plan, Teilsummen 661 · 205 · 430 · 255 nachgerechnet); mit `spec/spezifikation.md`, die §3 ebenfalls nennt, 1562 (I-1).

## SPEC-107 Kopf und Nr. 1–13 gegen den Code

| Nr. | Urteil |
|---|---|
| Kopf: Blob-Format ohne Existenzprüfung, Unveränderlichkeit, Argument vor Kontext | erfüllt; Präfix-Fänger jetzt rot (`abcdefg`+64 Hex), `ctx.Err()` vor der Argumentprüfung in `SetMetadata` und `Coverage.List` rot |
| 1–2 `Create`, `Get`, `List` | erfüllt |
| 3 `SetMetadata`, Argumente vor Kennung | erfüllt; Umstellung der Reihenfolge rot (`SetMetadata ersetzt, löscht und prüft`) |
| 4 Art-Filter, Ordnung nach Zeit | erfüllt; Filter und `ORDER BY id, created_at` rot |
| 5 atomar | erfüllt |
| 6 Gültigkeit mit UTF-8 | erfüllt; `jsonGueltig` zu `json.Valid` rot (Modell und Vertrag) |
| 7 Idempotenz, Coverage unter `Snapshot.ID` | erfüllt |
| 8 Prüfreihenfolge | erfüllt bis auf „auch eine leere Kennung“ ohne Fänger (V-1) |
| 9–10 Quellen, `BaselinesOf`, `Sources`, `Get` | erfüllt |
| 11 Coverage lesen, Evidenz byte-gleich | erfüllt in `Get` und `List` |
| 12 Kontext | erfüllt |
| 13 leere Kennung | erfüllt: sechs Leser einzeln auf `ErrUngueltig` gedreht, je rot; drei Prüfungen (`Coverage.Get`, `Coverage.List`, `Snapshots.List`) entfernt, je rot |

## Zusage · Fänger: Mutationen dieses Laufs

„rot“ nennt den Teiltest, aus dem die Mutation rot wurde (Meldung geprüft); „grün“ ist eine überlebende Mutation.

| Plan-Zeile | Mutation | Ergebnis |
|---|---|---|
| Atomar | `ROLLBACK` zu `COMMIT` · Fehler von `schreibePublikation` verschluckt · `ROLLBACK` mit abgebrochenem Kontext | rot (`Publish ist atomar`; die erste auch `TestSessionVertrag/Delete nimmt Exchanges mit …`) |
| Idempotent | Hash-Abfrage unwirksam · Hash über Services · `BEGIN IMMEDIATE` zu `BEGIN` | rot (`gleicher Anfrage-Hash ist idempotent`, `gleichzeitige Publish …`, `Prüfreihenfolge`) |
| Gültigkeit | `Validate` aus `Publish` entfernt · `Import.Validate` ohne `ServiceID`, ohne `BaselineSnapshotID` · `report_json` und Evidenz ohne UTF-8-Prüfung · `utf8.ValidString` aus `jsonGueltig` | rot (`TestImportValidate`, `ungültige Publikationen …`); die 28 Verstöße nachgezählt, stimmt mit dem Plan |
| Prüfreihenfolge | Service der Quelle ungeprüft (`case dienst != s.ServiceID`) | rot (`Publish ist atomar`, `Prüfreihenfolge`) |
| Zeit als UTC | Import-Zeit ohne `UTC()` · Snapshot-Zeit ohne `UTC()` · `ORDER BY id, created_at` | rot (`TestPublishSchreibtArtZeitformatUndHaeltDieSchluessel`; `nur Baselines sind sichtbar …`) |
| Service und Metadaten | Präfix von `BlobReferenzGueltig` (`CutPrefix` auf beliebige 7 Zeichen) · Reihenfolge in `SetMetadata` umgestellt | rot (`TestBlobReferenzGueltig`; `SetMetadata ersetzt, löscht und prüft`) |
| Coverage lesen | `json()` in `Get` · `json()` in `List` · Sortierung absteigend | rot (`Coverage lesen, Evidenz byte-gleich`) |
| Kontext, Argument vor Kontext | `ctx.Err()` an den Anfang von `SetMetadata` · von `Coverage.List` · `inTx` mit `WithoutCancel` (Rest siehe Review) | die ersten zwei rot (`abgebrochener Kontext …` Service/Baseline); letztere grün, weil `ctx` an der zweiten Stelle noch meldet (kein Befund, wie im Review) |
| Leere Kennung (Nr. 13) | je Leser (`Snapshots.Get`, `Imports.Get`, `Sources`, `BaselinesOf`, `Metadata.Get`, `SetMetadata`) auf `ErrUngueltig` · die drei Prüfungen entfernt | alle rot |
| Art `baseline` | Filter in `baselineBekannt` entfernt | rot (`Baseline-Quellen …`, `Coverage lesen …`) |
| Sortierung der Leser | `Sources`, `BaselinesOf`, `Coverage.List` absteigend | rot |
| Quelle leer (Nr. 8) | `schreibeQuelle` überspringt die leere Kennung (`return nil`) | **grün** (V-1) |

## Review-Befunde: Nachweis der Behebung

| Review | Zustand |
|---|---|
| F-1 (`Coverage.List` byte-gleich) | behoben, selbst nachgefahren: `json()` in `List` und in `Get` je rot; der Test vergleicht jetzt die ganzen Einträge (`reflect.DeepEqual`) mit nicht minimaler Evidenz |
| F-2 (Präfix) | behoben, rot (`TestBlobReferenzGueltig`, Fall `abcdefg…`) |
| F-3 (Argument vor Kontext) | behoben für `SetMetadata` und `Coverage.List`, beide Mutationen rot |
| F-4 (Import-Zeit UTC) | behoben, rot (Import und Snapshot in `UTC+5`) |
| F-5 (Randformen im Code) | in `SPEC-107` Nr. 3, 7, 8 entschieden und mit Fängern versehen; die Teilform „leere Quell-Session“ ohne Fänger (V-1) |
| F-6 (a) leere Kennung | entschieden in Nr. 13, je Leser gefangen. (b) UTF-8 | entschieden in Nr. 6, gefangen. (c) NUL, nur Leerzeichen | unter *Ungeprüft* genannt (ehrlich) |
| F-7 (`complete` ohne Beweis) | nicht behoben, in `SPEC-107` *Ungeprüft* und Plan §6 als „weiter offen“ geführt; die Entscheidung gehört dem Planner/Architect, `LH-FA-11` und `LH-FA-18` (Boundary) bleiben bis dahin für die Ablage ungeschlossen, die Fachlogik des Imports ist ausdrücklich außerhalb des Slice |
| F-8 (`Import.Validate`) | behoben: `TestImportValidate`, beide Halbbedingungen rot |
| F-9 (Plan-Mutation `inTx`) | behoben: der Plan nennt die semantischen Mutationen, sie sind rot |
| F-10 (Wettlauf) | in `SPEC-107` *Ungeprüft* und Plan §6 geführt, weiter offen, mit dem Löschen von Baselines zu entscheiden |
| F-11 (Marke) | im Plan vermerkt und gemessen richtig, siehe I-1 |
| F-12 (`UnixNano`) | behoben: `rand.Uint64`; die Funktion bleibt prozessweit angemeldet (kein Befund bei sequenziellen Fällen) |

## Findings

| ID | Stufe | Befund | Quelle | Pfad | Verifizierbar |
|---|---|---|---|---|---|
| V-1 | MEDIUM | `SPEC-107` Nr. 8 sagt zu: eine unbekannte Quell-Session „(auch eine leere Kennung)“ ist `ErrNichtGefunden`. Die Zeile wurde mit der Behebung von F-5 eingefügt, hat aber weder in der Tabelle *Zusage · Fänger* noch in den Tests einen Fall (`SourceSessionIDs = []string{""}` kommt nirgends vor). Mutation `if sessionID == "" { return nil }` in `schreibeQuelle`: alle Tests grün; eine Publikation mit leerer Quelle würde dann ohne Quell-Zeile veröffentlicht (nur eine echte Quelle trägt die Provenienz). Der Code verhält sich richtig. | `AGENTS.md` §3.15, §3.10 · `SPEC-107` Nr. 8 | `src/adapters/driven/sqlite/imports.go` · `schreibeQuelle`; `src/hexagon/ports/outbound/outboundtest/baseline_vertrag.go` · `publikationFall3` | ja, Mutation (selbst gefahren) |
| V-2 | LOW | Die Kopfkommentare der Vertragsläufe sagen weniger zu als geprüft wird: `RunBaselineVertrag` nennt „Nr. 4 bis 12“, prüft aber auch Nr. 13 und den Kopf (Unveränderlichkeit, Art); `RunServiceVertrag` nennt „Nr. 1 bis 3 und Nr. 12“, prüft auch die leere Kennung (Nr. 13) und die Reihenfolge in Nr. 3. Keine Überdehnung im Sinn von §3.11, aber der Kopf führt den Leser eines zweiten Adapters in die Irre über den Umfang, den der Lauf abdeckt. | `AGENTS.md` §3.7, §3.11 | `src/hexagon/ports/outbound/outboundtest/baseline_vertrag.go` · `RunBaselineVertrag`; `service_vertrag.go` · `RunServiceVertrag` | ja |
| V-3 | LOW | Plan §6 führt drei Risiken als „weiter offen“ (Wettlauf der Leser, `complete` ohne Beweis, Form der Evidenz). Zwei davon haben ausdrücklich keine Adresse (§3.13: der Punkt geht an den Planner). Bis dahin kann der Slice nach `done/` nur, wenn der Planner die Ausgänge annimmt; ein Slice ohne Ausgang aller Risiken geht nicht nach `done/` (Modul 5). Ehrlich benannt, aber noch nicht zugewiesen. | `AGENTS.md` §3.13 · `LH-FA-11` | `docs/plan/planning/done/slice-baseline-persistenz.md` · §6 (dritter bis fünfter Punkt) | ja |
| I-1 | INFO | Die Marke in §4 nennt 1551 Zeilen „über die Pfade aus §3“; §3 nennt `spec/spezifikation.md` mit, die Summe mit ihr beträgt 1562. Die Marke ist so oder so überschritten (über 700); der Entscheid, den Slice ungeteilt zu führen, steht beim Implementer im Plan und gehört dem Planner (§3.14). | `AGENTS.md` §3.14 | `docs/plan/planning/done/slice-baseline-persistenz.md` · §4 | ja, `git diff --numstat` |
| I-2 | INFO | `SPEC-107` *Ungeprüft* benennt ehrlich, was kein Test fängt: Kennungen mit NUL oder nur Leerzeichen, Form der Evidenz, `complete` ohne Beweis, Wettlauf der Leser. Für `LH-FA-11` (nachweislich vollständig) bleibt die Durchsetzung eine offene Entscheidung (F-7); kein Befund gegen diesen Slice, aber Voraussetzung, dass `welle-record-replay` den Satz übernimmt. | `LH-FA-11`, `LH-FA-18` | `spec/spezifikation.md` · `SPEC-107` Schluss | nein |

## Offen und erwartet (kein Befund)

Häkchen der DoD, Closure-Notiz mit Steering-Loop-Lerneintrag (Klassen „Zusage ohne rot werdende Mutation“ und „Randform nicht/im Code entschieden“ treten erneut auf; das Register zählt, der Anker folgt der Welle-Closure), Beobachtungs-Register, drei Paarungen, Häkchen für den Review-Report.

## Übergabe

Der Implementer ergänzt V-1 (ein Fall in `publikationFall3` mit leerer Quell-Session, Plan-Zeile „Quelle nicht vorhanden“ nennt ihn) und nimmt V-2 mit (Köpfe auf den tatsächlichen Umfang bringen). V-3 und I-2 gehen an den Planner (Ausgang und Adresse der offenen Risiken, `AGENTS.md` §3.13), I-1 ebenfalls. Danach genügt eine erneute Verifikation der Zeile Nr. 8.

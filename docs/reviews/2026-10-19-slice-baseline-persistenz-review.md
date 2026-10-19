# Review-Report: slice-baseline-persistenz — 2026-10-19

**Review-Art:** Code — gegen Plan, Entscheidungen und Hard Rules.

**Gegenstand:** `slice-baseline-persistenz`, Diff `b70691f..561450a` (Vertrag `SPEC-107` und Plan-Nachzug davor, Umsetzung in `561450a`; 13 Dateien, 1528 Einfügungen).

**Skill:** `.harness/skills/reviewer.md` @ 424c252
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Eingangs-Kontext:**

- Plan `slice-baseline-persistenz` (inkl. Tabelle *Zusage · Fänger*, `AGENTS.md` §3.15)
- [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [LH-FA-10](../../spec/lastenheft.md), [LH-FA-11](../../spec/lastenheft.md), [LH-FA-18](../../spec/lastenheft.md)
- `SPEC-107` (Kopf und Nr. 1–12 einzeln gegen den Code), `SPEC-004`, `SPEC-043`, `SPEC-044`, `SPEC-105`, `SPEC-106`
- `AGENTS.md` (Hard Rules §3.1–§3.15)
- vorherige Reports am Nachbarmodul: `2026-10-19-slice-blob-ablage-review.md`, `2026-10-19-slice-session-recording-persistenz-review.md`

**Eigene Läufe:** `make a-check` (nennt nur die bekannten drei Dateien in `tools/gegenproben/`), `make lint` und `make docs-check` (0 Befunde) grün. Alle Mutationen und explorativen Tests liefen in Wegwerf-Kopien (`git archive HEAD`) im Scratchpad, im Container `golang:1.27-alpine` mit gebauten Modulen aus der Stufe `deps`, `--network none`, `CGO_ENABLED=0`; 105 Mutationen, jede gofmt- und compile-sicher geprüft (Fehlbauten verworfen und neu gefasst). Der Arbeitsbaum blieb unberührt. `-race` ist im Container nicht möglich (verlangt cgo).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `Coverage.List` liest die Evidenz nicht byte-gleich geprüft: Mutation `json(evidence_json)` in `List` (minifiziert) bleibt grün. `coverageFall` prüft die Bytes nur über `Get`; `List` prüft nur die Namen, und `pruefeLesen` vergleicht Evidenz, die schon minimal ist (`{"pages":3}`, `{}`). `SPEC-107` Nr. 11 sagt „Die Evidenz wird byte-gleich gelesen" für beide Leser, der Plan nennt die Mutation für beide. | `AGENTS.md` §3.10, §3.15, §3.11 · `SPEC-107` Nr. 11 | `src/adapters/driven/sqlite/baselines.go` · `"SELECT entity_set, status, evidence_json FROM coverage WHERE baseline_snapshot_id = ? ORDER BY entity_set"` | ja, Mutation grün (selbst gefahren) | Zusage ohne rot werdende Mutation |
| F-2 | MEDIUM | Das Präfix `sha256:` von `BlobReferenzGueltig` hat keinen Fänger. Mutation `strings.CutPrefix(ref, ref[:min(7, len(ref))])`: alle Tests grün; `abcdefg` + 64 Hexziffern gälte als gültige `metadata_blob`/`state_blob`. `TestBlobReferenzGueltig` ändert beim Fall `sha1:` die Länge und trifft so nur die Längenprüfung. Gleiche Lücke wie F-1 im Report zu `slice-blob-ablage`, jetzt in einer zweiten Kopie der Prüfung. | `AGENTS.md` §3.10, §3.15 · `SPEC-107` Kopf (Referenz nach `SPEC-106`) | `src/hexagon/model/baseline.go` · `strings.CutPrefix(ref, "sha256:")` | ja, Mutation grün (selbst gefahren) | Zusage ohne rot werdende Mutation |
| F-3 | MEDIUM | „Argument vor Kontext" (`SPEC-107` Kopf, gilt je Operation) hat für `SetMetadata` und `Coverage.List` keinen Fänger. Mutation: `ctx.Err()` an den Anfang der Methode gesetzt, Suite grün (`SetMetadata` mit ungültigem Paar und abgebrochenem Kontext liefert dann `context.Canceled` statt `ErrUngueltig`; ebenso `List` mit leerer Kennung). Dieselbe Mutation in `Create`, `Coverage.Get`, `Snapshots.List` wird rot. Die Plan-Zeile „Kontext (Nr. 12), Argument vor Kontext" nennt `Service und Baseline` pauschal. | `AGENTS.md` §3.15 · `SPEC-107` Kopf, Nr. 12 | `src/hexagon/ports/outbound/outboundtest/service_vertrag.go` · `erwarte(t, v.Metadata.SetMetadata(ctx, "d", "", ""), context.Canceled)` | ja, zwei Mutationen grün (selbst gefahren) | Zusage ohne Fänger |
| F-4 | MEDIUM | Die Plan-Zeile „Zeiten als UTC" nennt für den Import den Fänger `Publish schreibt alles` (Import-Zeit in `UTC+5`). Mutation `i.CreatedAt.Format(…Z07:00)` statt `zeitText(i.CreatedAt)` bleibt grün: der Test liest die Zeit zurück und vergleicht mit `Equal`, und `TestPublishSchreibt…` prüft den Text mit einer Zeit, die schon UTC ist. Ein Import mit Versatz wird so als `…+05:00` gespeichert (`SPEC-105`: UTC-Text), was jede spätere Sortierung nach Text bricht. Die Snapshot-Zeit hat einen Fänger (Mutation rot). | `AGENTS.md` §3.15 · `SPEC-105` (Zeitform), `SPEC-107` Kopf | `src/adapters/driven/sqlite/imports.go` · `zeitText(i.CreatedAt)` | ja, Mutation grün (selbst gefahren) | Fänger im Plan fängt die genannte Mutation nicht |
| F-5 | MEDIUM | Randformen, die `SPEC-107` nicht nennt, setzt der Code selbst, statt anzuhalten (`AGENTS.md` §3.12): `SetMetadata` prüft Argumente vor der Kennung, Coverage-Einträge liegen unter `Snapshot.ID`, eine leere Quell-Session ist `ErrNichtGefunden`, `Sources`/`BaselinesOf`/`Coverage.List` lesen in zwei Abfragen ohne Transaktion. Der Plan §6 gibt sie als „weiter offen" an den Architect, entschieden sind sie im Code. Die Reihenfolge in `SetMetadata` hat keinen Fänger (Mutation „Prüfung nach dem Schreiben, Kennung zuerst" grün; unbekannte Kennung mit ungültigem Paar ist ungetestet). | `AGENTS.md` §3.12 · `SPEC-107` Nr. 3, 8, 9 | `src/adapters/driven/sqlite/metadata.go` · `if err := model.MetadatenGueltig(blob, hash); err != nil` | ja, Mutation grün (selbst gefahren) | Randform im Code entschieden statt vor dem Code |
| F-6 | MEDIUM | Weitere Randformen sind weder in `SPEC-107` noch in Plan §6 genannt. (a) Leere Kennung: `Coverage.Get`/`List` und `Snapshots.List` liefern `ErrUngueltig`, dagegen `Snapshots.Get("")`, `Imports.Get("")`, `Sources("")`, `BaselinesOf("")`, `Metadata.Get("")`, `SetMetadata("", …)` `ErrNichtGefunden` (selbst gefahren). (b) `json.Valid` lässt ungültiges UTF-8 zu: Evidenz `"\xff\xfe"` und `report_json` werden gespeichert und byte-gleich gelesen, obwohl `SPEC-107` Nr. 6 „gültiges JSON" verlangt; ein späteres `json_extract` in SQLite sieht sie anders. (c) Kennungen mit NUL-Byte, ungültigem UTF-8 oder nur Leerzeichen werden angenommen (`Service`, `Snapshot`, `Import`, `entity_set`). Der Vertrag ist nicht doppeldeutig, er schweigt. | `AGENTS.md` §3.12 · `SPEC-107` Nr. 1, 3, 4, 6, 11 | `src/hexagon/model/baseline.go` · `!json.Valid([]byte(c.EvidenceJSON))` | ja, explorativer Test (selbst gefahren) | Randform nicht entschieden |
| F-7 | MEDIUM | Nichts im Slice zwingt eine `complete`-Coverage zu einem Beweis: `Validate` akzeptiert `complete` mit Evidenz `null`, `{}`, `[1, 2]`; die Tests nutzen genau diese Formen. `SPEC-107` nimmt die „Form der Evidenz" aus der Prüfung, `LH-FA-11` verlangt „nachweislich vollständig". Nicht HIGH, weil die Erzeugung von `complete` Fachlogik der Folgewelle ist und `SPEC-107` die Lücke benennt; die Ablage ist aber die einzige Stelle, an der der Satz einheitlich durchsetzbar wäre. Plan §6 führt den Punkt als „weiter offen" ohne Adresse, die ihn annimmt (`AGENTS.md` §3.13). | `LH-FA-11`, `LH-FA-18` · `AGENTS.md` §3.13 · `SPEC-044` | `docs/plan/planning/done/slice-baseline-persistenz.md` · `Die Evidenz der Coverage (\`evidence_json\`) ist frei geformt` | nein | Vollständigkeitsbeweis nicht an die Ablage gebunden |
| F-8 | LOW | Die Halbbedingungen `i.ServiceID == ""` und `i.BaselineSnapshotID == ""` in `Import.Validate` haben keinen Fänger: jeweils entfernt, bleibt die Suite grün. Über `Publication.Validate` sind sie unerreichbar (zuerst prüft `Snapshot.Validate`, dann `pruefeBezug`), und `Import.Validate` ist exportiert, aber ohne eigenen Modelltest (anders als `Service` und `BaselineSnapshot`). | `AGENTS.md` §3.10 | `src/hexagon/model/baseline.go` · `i.ID == "" \|\| i.ServiceID == "" \|\| i.BaselineSnapshotID == ""` | ja, zwei Mutationen grün (selbst gefahren) | Prüfung durch Reihenfolge verdeckt, ohne Fänger |
| F-9 | LOW | Die Plan-Mutation „`BEGIN`, `COMMIT`, `ROLLBACK` in `inTx` entfernt" wird rot, aber aus dem Aufbau: `publikationBasis` ruft zuerst `Sessions.Create` über dieselbe `inTx` und bricht mit `cannot commit - no transaction is active` bzw. `cannot start a transaction within a transaction`, bevor `Publish ist atomar` läuft. Die semantische Mutation (`ROLLBACK` durch `COMMIT` ersetzt; Fehler von `schreibePublikation` verschluckt) wird in `Publish ist atomar` rot, ist aber nicht die im Plan genannte. | `AGENTS.md` §3.10 (Test zählt nur aus dem benannten Grund) | `docs/plan/planning/done/slice-baseline-persistenz.md` · `\`BEGIN\`, \`COMMIT\`, \`ROLLBACK\` in \`inTx\` entfernt` | ja | Mutation rot aus anderem Grund |
| F-10 | INFO | `Sources`, `BaselinesOf` und `Coverage.List` prüfen die Kennung und lesen danach in einer zweiten Abfrage, ohne gemeinsame Transaktion; es gibt keinen Test Leser gegen `Publish`. Heute ohne Wirkung (der Schreiber veröffentlicht atomar, Baselines sind nicht löschbar, `SPEC-107` nimmt das Löschen aus; ein explorativer Lauf mit 40 gleichzeitigen `Publish`/`Delete`/Lesern lieferte nur typisierte Fehler, kein `database is locked`). Mit dem Löschen von Baselines würde daraus ein Wettlauf (Kennung bekannt, dann gelöscht, leere statt `ErrNichtGefunden`). | `AGENTS.md` §3.12 | `src/adapters/driven/sqlite/imports.go` · `func (r *ImportRepository) Sources` | nein | Randform nicht entschieden |
| F-11 | INFO | Marke `über 700 Zeilen (Summe aus git diff --numstat …)` überschritten: gemessen 1507 über `src/` (`b70691f..561450a`), der Plan §4 vermerkt es mit Begründung und führt den Slice ungeteilt weiter. Die Entscheidung über das Weiterführen gehört dem Planner/Architect; die Zahl im Plan ist richtig. | `AGENTS.md` §3.14 | `docs/plan/planning/done/slice-baseline-persistenz.md` · `Marke überschritten` | ja, `git diff --numstat` | Marke überschritten |
| F-12 | INFO | `neueVerbindung` registriert je Aufruf eine globale Skalarfunktion `odr_abbruch_<UnixNano>` (`MustRegisterScalarFunction`); sie bleibt für den Prozess registriert und wirft bei gleichem Namen eine Panik. Heute läuft jeder Vertragsfall nacheinander, daher kein Fehlschlag. Würde ein Fall parallel laufen (`t.Parallel`), wäre der Name nicht eindeutig garantiert. | Maintainability | `src/adapters/driven/sqlite/repositories_test.go` · `MustRegisterScalarFunction(name, 0` | nein | Globaler Zustand im Testaufbau |

## Auflösung der Tabelle *Zusage · Fänger* (Plan §3)

Jede Zeile mit eigener Mutation; „rot" nennt den Teiltest, aus dem sie rot wurde. Kürzel: PV = `TestBaselineVertrag`, SV = `TestServiceVertrag`.

| Zeile | Mutation | Ergebnis |
|---|---|---|
| Atomar (Nr. 5) | `ROLLBACK` durch `COMMIT` ersetzt | rot (PV `Publish ist atomar`) |
|  | Fehler von `schreibePublikation` verschluckt | rot (PV `Publish ist atomar`, `Prüfreihenfolge`) |
|  | `ROLLBACK` mit dem abgebrochenen Kontext | rot (PV `Publish ist atomar`) |
|  | `BEGIN`/`COMMIT`/`ROLLBACK` durch `SELECT 1` | rot, aber aus dem Aufbau (F-9) |
| Idempotent (Nr. 7) | Hash-Abfrage unwirksam | rot (`gleicher Anfrage-Hash ist idempotent`, `gleichzeitige Publish`, `Prüfreihenfolge`) |
|  | Hash über Services | rot (`gleicher Anfrage-Hash …`, `Prüfreihenfolge`) |
|  | Treffer als neu gemeldet | rot (drei Teiltests) |
|  | `BEGIN IMMEDIATE` durch `BEGIN` | rot (PV `gleichzeitige Publish …`, `TestRecordingVertrag/gleichzeitige Anhänge …`) |
| Gültigkeit (Nr. 6) | je `case` in `baseline.go` (15) durch `false` | alle rot (`ungültige Publikationen …`, `Create lehnt Ungültiges …`, `TestPublicationValidate`, `TestServiceValidate`, `TestBaselineSnapshotValidate`) |
|  | Halbbedingungen (11) | neun rot; `Import.ServiceID`, `Import.BaselineSnapshotID` grün (F-8) |
|  | je Prüfung der Liste in `Publication.Validate` entfernt (5) | rot |
|  | Duplikat-Prüfungen, „keine Quelle", `Coverage.Validate` entfernt | rot |
|  | `Validate` in `Publish` entfernt | rot |
| Prüfreihenfolge (Nr. 8) | Kennungen vor Hash; Service-Prüfung entfernt; Quelle vor Kennungen; Quellen rückwärts; Kontext vor Gültigkeit; Gültigkeit nach Hash | alle rot (`Prüfreihenfolge von Publish`) |
| Unveränderlich | Prüfung der Snapshot-Kennung / der Import-Kennung entfernt | rot |
|  | Kennungsprüfung ohne Art-Filter | rot (`Baselines sind unveränderlich`) |
|  | nur `INSERT OR REPLACE` | grün (Plan sagt „zusätzlich"; die Kennungsprüfung fängt vorher, kein Befund) |
| Quellen, `BaselinesOf`, `Sources`, `Get` (Nr. 9, 10) | Quelle nicht geschrieben; beide Sortierungen absteigend; Prüfung von Session und Baseline entfernt; Fehler von `Get` vertauscht | alle rot |
| Quelle eines anderen Service | Prüfung entfernt | rot (`Publish ist atomar`, `Prüfreihenfolge`) |
| Nur Art `baseline` (Nr. 4) | Filter in `Get`, `List`, `baselineBekannt` einzeln entfernt | je rot (`nur Baselines sind sichtbar …`) |
| Ordnung nach Zeit, UTC | Kennung vor Zeit; Snapshot-Zeit ohne UTC | rot |
|  | Import-Zeit ohne UTC | **grün** (F-4) |
| Service und Metadaten (Nr. 1–3) | `Validate`, Prüfung in `SetMetadata`, doppeltes `Create`, `List` absteigend, Fehler von `Get`/`SetMetadata`, Leere-Liste nil | alle rot |
|  | Blob-Format: Länge, Großbuchstaben | rot |
|  | Blob-Format: Präfix | **grün** (F-2) |
| Coverage lesen (Nr. 11) | `json()` in `Get` | rot |
|  | `json()` in `List` | **grün** (F-1) |
|  | Sortierung absteigend; leere Argumente (beide Hälften); Prüfung der Baseline entfernt; nil statt leerer Liste | rot |
| Kontext (Nr. 12) | Kontext ignoriert in Metadata, `Snapshots.Get`, `Snapshots.List`, `Coverage.Get`, `Imports.Get`, vollständig in `Publish` und `Coverage.List` | rot |
|  | je nur eine von zwei Stellen in `BaselinesOf`, `Coverage.List`, `kennungen`, `baselineBekannt`, `Publish` | grün, weil die zweite Stelle den Abbruch noch meldet (kein Befund; die vollständige Mutation ist rot) |
| Argument vor Kontext | Kontext zuerst in `Create`, `Coverage.Get`, `Snapshots.List` | rot |
|  | Kontext zuerst in `SetMetadata`, `Coverage.List` | **grün** (F-3) |

## Prüfung der Fragen aus dem Auftrag

- **Zusammentreffen mehrerer Verstöße:** `Publish` ist für Gültigkeit, Kontext, Hash, Service, Kennung und Quelle paarweise geprüft (alle Paar-Mutationen rot). Offen: `SetMetadata` mit unbekannter Kennung und ungültigem Paar (F-5), `Create` ohne Metadaten-Hälfte und ohne Kennung ist wegen der Einzelprüfung nicht unterscheidbar (gleicher Fehler, kein Befund).
- **Kontext vs. Argument:** F-3.
- **Zeit und Ordnung:** `ORDER BY created_at, id` auf festbreitem UTC-Text ist chronologisch; Zone `UTC-5` und Gleichstand sind getestet (rot bei Mutation). Import-Zeit F-4. Jahr 1 und 9999 über die gemeinsame `zeitGueltig`, im Vertrag getestet.
- **Leere Werte:** F-6.
- **Doppelte Quellen:** doppelte Quell-Session ist `ErrUngueltig` (Modell und Vertrag, Mutation rot); `s1`/`S1` sind verschiedene Sessions (`ErrNichtGefunden` für die unbekannte), kein Befund; doppelte `entity_set` rot.
- **Coverage-Evidenz:** F-1, F-6 (b), F-7.
- **Kommentare (§3.7, §3.11):** keine Konjunktive über Verworfenes, keine Chronik. Zusage über das Prüfbare hinaus: `Coverage` im Modell („wird byte-gleich gespeichert") und `RunBaselineVertrag` („prüft `SPEC-107` Nr. 4 bis 12") tragen die Zusage für `List` ohne Fänger (F-1).
- **Wettläufe in `Sources`/`BaselinesOf`:** F-10. Zwei gleichzeitige `Publish` mit gleichem Hash und mit gleicher Snapshot-Kennung (verschiedene Hashes) ordnen sich über `BEGIN IMMEDIATE`: genau einer schreibt, die übrigen liefern den Import bzw. `ErrBereitsVorhanden`, keine rohe SQLite-Meldung (acht Goroutinen, selbst gefahren).
- **Plan folgt der Korrektur (§3.9):** Kopf, §1, §3, §6 sind nachgezogen; die Testnamen der Tabelle lösen auf (`TestPublicationValidate`, `Prüfreihenfolge von Publish`, 26 Verstöße gezählt: stimmt); die Zeile „Zeiten als UTC" nennt einen Fänger, der die Mutation nicht fängt (F-4); die Plan-Mutation zu `inTx` ist nicht die rote (F-9). DoD-Haken stehen offen (vor der Closure üblich).
- **a-check-Schichten:** `src/adapters/driven/sqlite/{baselines,imports,metadata}.go` importieren nur `model`, `ports/outbound` und Bibliotheken; `outboundtest` hängt nur an `model` und `outbound`. `make a-check` ohne Befund außer den bekannten drei Dateien in `tools/gegenproben/`.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/baseline.go` | geprüft, Befunde F-2, F-6, F-8; Kommentare tragen Zusage und Kopplung |
| `src/hexagon/model/model_test.go` | geprüft, Befunde F-2, F-8 |
| `src/hexagon/ports/outbound/baseline_repository.go`, `metadata_repository.go` | geprüft, ohne Befund (Schnittstellen entsprechen `SPEC-107`; Kommentare nennen nur Geprüftes) |
| `src/hexagon/ports/outbound/outboundtest/baseline_vertrag.go`, `service_vertrag.go`, `vertrag.go` | geprüft, Befunde F-1, F-3, F-4 |
| `src/adapters/driven/sqlite/imports.go` | geprüft, Befunde F-4, F-5, F-10; Reihenfolge und Atomarität ohne Befund |
| `src/adapters/driven/sqlite/baselines.go` | geprüft, Befund F-1; Art-Filter und Sortierung ohne Befund |
| `src/adapters/driven/sqlite/metadata.go` | geprüft, Befund F-5 |
| `src/adapters/driven/sqlite/baselines_test.go`, `repositories_test.go` | geprüft, Befunde F-4, F-12 |
| `spec/spezifikation.md` (`SPEC-107`) | geprüft, Nr. 1–12 umgesetzt; Lücken in F-5, F-6, F-7 |
| Plan `done/slice-baseline-persistenz.md` | geprüft, Befunde F-4, F-7, F-9, F-11; Adresse in §1 (`slice-workspace-state-persistenz`) nimmt die Workspace-Snapshots an (§3.13) |
| `.golangci.yml`, `//nolint` | geprüft, ohne Befund (keine Suppression, keine Lockerung, `AGENTS.md` §3.2/§3.6) |
| `.a-check.yml`, Schichten | geprüft, ohne Befund |
| Schema `0001_init.sql` (Fremdschlüssel, `uq_import_baseline`) | geprüft, ohne Befund; `imports` hat keine Eindeutigkeit über `(service_id, request_hash)`, die Idempotenz trägt allein `BEGIN IMMEDIATE` (Mutation auf `BEGIN` rot) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 7 |
| LOW | 2 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Zusage ohne rot werdende Mutation · Zusage ohne Fänger · Fänger im Plan fängt die genannte Mutation nicht · Randform im Code entschieden statt vor dem Code · Randform nicht entschieden · Vollständigkeitsbeweis nicht an die Ablage gebunden · Prüfung durch Reihenfolge verdeckt, ohne Fänger · Mutation rot aus anderem Grund · Marke überschritten · Globaler Zustand im Testaufbau

## Verdikt

**Merge-blockierend:** ja — kein HIGH, aber sieben MEDIUM. F-1 bis F-4 sind Zusagen, deren Mutation grün bleibt (`AGENTS.md` §3.15: Zusage ohne Fänger ist ein MEDIUM); F-2 wiederholt die Lücke des Präfix-Fängers aus dem Report zu `slice-blob-ablage`, die Klasse „Zusage ohne rot werdende Mutation" tritt damit erneut auf (Zähler im Register `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`), ebenso „Randform nicht/im Code entschieden" in F-5 und F-6 (Register `BEO-REPO/randform-erst-im-review-entschieden`). F-7 ist keine HIGH-Einstufung, weil `SPEC-107` die Lücke benennt; der Architect entscheidet, ob die Ablage `complete` ohne Beweis ablehnen muss.

**Übergabe:** Findings gehen an den Implementer; F-5, F-6, F-7, F-10 zusätzlich an den Architect (Randformen vor dem Code, `AGENTS.md` §3.12; Adresse für F-7 nach §3.13), F-11 an Planner/Architect. Die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation gegen die DoD.

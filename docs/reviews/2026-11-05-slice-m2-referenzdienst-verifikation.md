# Verifikations-Report: slice-m2-referenzdienst — 2026-11-05

**Verifikations-Art:** DoD- und Entscheidungs-Konformität, Plan-vs-Code-Diff — geprüft gegen `docs/plan/planning/done/slice-m2-referenzdienst.md` (DoD §2, Tabelle *Zusage · Fänger* M1 bis M10 und M9b, Fixes F-1 bis F-10, Raster §6, Block der Größenmarke §4), `SPEC-139` Nr. 1, 3, 4, `SPEC-122` Nr. 2, `SPEC-039` Nr. 2, `SPEC-102` (Fälle 171 bis 173), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) (auf `main` Accepted), `AGENTS.md` §3.9, §3.10, §3.15, §3.17.

**Gegenstand:** `slice-m2-referenzdienst` (Welle `welle-record-replay`), Zweig `slice-m2-referenzdienst`, HEAD `a0e0dde`, Worktree `/Development/odata-recorder-wt-m2-referenzdienst`. Bezug: `LH-FA-04`, `LH-FA-05`, `LH-FA-09`, `LH-FA-14`, `LH-QA-06`.

**Rolle / Modell / Datum:** Verifier · claude-sonnet-5-5 · 2026-11-05. Frischer Kontext, einmalige Verifikation, kein Push.

## Ergebnis in einem Satz

Kein HIGH, kein MEDIUM. Ein LOW (Plan nennt [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) noch „Proposed“, auf `main` ist sie Accepted) und ein INFO (Closure-Haken offen). Die DoD-Lieferpunkte sind erfüllt, `make gates` ist selbst grün gefahren, die Größenmarke hält, drei an der Klasse neu gebildete Mutanten sind rot.

## Belege (selbst gemessen)

| Beleg | Messung |
|---|---|
| `make gates` | Exit 0 (Log `.../scratchpad/gates.log`, 41 s, Images gecacht); `a-check` (Aufruf gegen den Worktree), `migration-check: ok`, `zeichen: ok, 1155 Datei(en)` |
| Größenmarke | `groessenmarke slice-m2-referenzdienst.md: ok, Messung 980 von Marke 1700`; selbst gerechnet `git diff --numstat 1984383..HEAD` über die fünf Pfade des Blocks: **980** (m2_test 451, import_workspace_test 286, referenzdienst_test 163, wurzel 40, haken_test 25, `.a-check.yml` 8, export_test 3, Dockerfile 3, `.dockerignore` 1); 57,6 % der Marke |
| Tests im Zweig | `go test -race -count=1 ./src/bootstrap/ ./tests/...` im Image `odata-recorder:race` (netzlos, `--memory=4g --cpus=4`) auf Kopie des Worktrees: beide Pakete `ok` (65 s, 2 s) |
| Diff gegen `main` | 21 Dateien, 1319 Einfügungen, 66 Löschungen; Produktcode nur `src/bootstrap/wurzel.go` (+40), sonst Tests, Dockerfile, `.dockerignore`, `.a-check.yml`, Spezifikation, Gegenproben, Doku |
| Zweig vs. `main` | Merge-Basis `d318d25`; `main` steht auf `3c87c55` ([ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) Accepted), im Zweig nicht gemergt (Auftrag des Integrations-Planners) |

## DoD, Zeile für Zeile

| DoD-Punkt | Befund |
|---|---|
| Verdrahtung und Quellstand | erfüllt: `NeuerImport` und `NeueWorkspaces` in `bauteile` (`wurzel.go:67/68`, `99/100`), Verdrahtung im Root (Workspace in beiden Modi, Import nur in `verdrahteRecord`), `Import()`/`Workspaces()` additiv; `Dockerfile` `COPY tests ./tests`, `.dockerignore` `!tests/`, `.a-check.yml` Schicht `tests` mit drei Kanten, `harness/README.md` Zeile `make a-check` „unter `src/` und `tests/`“ mit Bindung [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md) und [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) |
| Referenzdienst und M2-Test | erfüllt: `tests/integration/referenzdienst_test.go` und `m2_test.go` (`TestDurchstichM2`), laufen grün unter `-race` |
| Zeilen und Mutationen | der Plan führt je Zeile M1 bis M10, M9b Mutation, Test und Grund (§3 Handoff-Tabelle) und Raten (30 Prozesse, `-race`); Stichprobe neu gebildet, siehe unten |
| `make gates` grün, `make gegenproben` und Stempel | Gates grün (selbst); Fälle 171 bis 173 in `tools/gegenproben/faelle.go` und `SPEC-102`, Stempel in eigenem Commit (`901a61a`) nachgezogen; `make gegenproben` nicht erneut gefahren (Stempel gleich der Eingabe, `make gates` meldet kein Veralten) |
| Review durchgeführt | Report liegt (`2026-11-01-slice-m2-referenzdienst-review.md`, `d0fbe00`), der Haken in §2 ist noch nicht gesetzt: Sache der Closure (INFO-1) |
| Doku-Update | `harness/README.md` ist geändert (Zeile `make a-check`); der Haken steht offen (Closure) |
| Closure-Notiz, Register, Ausgänge, Paarungen | offen; §7 „Wird bei der Closure gefüllt“; §6 trägt mehrere Risiken mit Ausgang *offen*, das ist Sache der Closure |

## Plan-vs-Code-Diff

| Zusage des Plans | Code | Urteil |
|---|---|---|
| Import im Modus `record` erhält `Dienste`, dieselben Regeln, `MaxCaptureBody`, `Sessions`, `Exchanges`, `Snapshots`, `Imports`, `Metadaten`, Codecs `csdl.New()`/`jsoncodec.New()`, `Blobs` Wächter, Uhr, IDs | `wurzel.go` `verdrahteRecord` | gleich |
| Import im Modus `replay` ist `nil`; Workspace-Dienst in beiden Modi | `h.importer` bleibt im Replay unbelegt; Workspace vor der Modus-Weiche | gleich |
| Workspace erhält `Workspaces`, `Transaktion`, `Snapshots`, `Blobs`, `Dienste`, `Wartung`, `Uhr`, `IDs`, `Log` | `wurzel.go:250-253` | gleich |
| `tests/` in `source`, `.dockerignore`, `.a-check.yml` mit Schicht `tests` | Diff gegen `main` | gleich |
| Marke 1700 | 980 | gleich (hält) |
| Kein Eingriff in SPEC-049, Composition-Root-Verhalten | Diff berührt `SPEC-049` nicht; `wurzel.go` nur additiv (+40/−0) | gleich |

## Klassen-Prüfung der Fixes (`AGENTS.md` §3.17)

Mutanten an Kopien des Quellstands (Image `odata-recorder:race`, netzlos, `--memory=4g`), jeweils an Elementen, die Plan und Review nicht nannten:

| Menge (Wortlaut der Zusage) | Mutation | Ergebnis |
|---|---|---|
| Felder von `workspace.Konfig` (F-3: 9 Felder) | `Uhr` des Workspace-Dienstes eine eigene `clock.NewSystemuhr()` | rot: `TestWorkspaceErhaeltWerte`, „Feld Uhr erreicht den Workspace-Dienst nicht wie erwartet“ |
| Dienste des Imports (M5, je Element) | `UpstreamURL` je Dienst um ein Zeichen verändert (der Plan nennt „nur der erste Dienst“) | rot: `TestImportErhaeltWerte` „Feld Dienste“, `TestModus`, `TestDurchstichM2` |
| Felder von `baselineimport.Konfig` (13 Felder, M5) | `Uhr` des Imports eine eigene `clock.NewSystemuhr()` | rot: `TestImportErhaeltWerte` „Feld Uhr erreicht den Import nicht“ |

Die Basis ohne Mutation ist grün. Die weiteren Zeilen (M1 bis M4, M6 bis M10) habe ich nicht neu mutiert; sie stützen sich auf die Beleglage im Plan (30 Prozesse je Mutation, `-race`) und den grünen Lauf. Ein Mutant („Import auch im Modus `replay`“) war ein Sed ohne Wirkung und zählt nicht.

## Befunde

- **LOW-1 — [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)-Status im Plan.** Der Plan nennt [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) an drei Stellen „Proposed“ (Kopf *Bezug*, §1 *Quellstand*, §6 F-2 mit dem Offen-Punkt „Annahme vor dem Merge, Adresse: User“). Auf `main` (`3c87c55`) ist sie seit 2026-11-05 Accepted, der Index dort führt `Accepted`. Der Planner zieht alle drei Stellen und den Ausgang des offenen Punktes (F-2) bei der Integration nach (`AGENTS.md` §3.9; Klasse: jede Stelle, die den Status nennt). Kein Code-Befund; der Zweig ändert die ADR nicht.
- **INFO-1 — Closure offen.** Review-Haken, Doku-Haken, Closure-Notiz, Register-Fortschreibung, Ausgänge von §6 (mehrere *offen*) und die Paarungen sind nicht gesetzt; das ist Sache der Closure, kein Befund an Code.
- **INFO-2 — akzeptierte Negative im Plan benannt.** `IDGenerator`-Instanz (zustandsloses Struct, F-10), Body-/Query-Geheimnisse, parallele Clients, Coverage `partial`, Log-Inhalt, Weg ins Netz an anderem Ziel: alle mit Grund und Folge in §6; nichts davon ist eine Zusage ohne Fänger.

## Urteil

Der Slice hält gegen die DoD-Lieferpunkte und die Entscheidungen. Offen sind nur Closure-Handlungen und der Nachzug des ADR-Status im Plan (LOW-1) durch den Planner.

# Planning — <Projektname>

Slice-Lifecycle: `open/` → `next/` → `in-progress/` → `done/`.

Reine `git mv`-Commits beim Wechsel zwischen Verzeichnissen — siehe Hard
Rule „git mv + Inhaltsänderung = zwei Commits" in
[`../../../AGENTS.md`](../../../AGENTS.md).

## Lifecycle-Bedeutungen

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand ist das Verzeichnis, kein Status-Feld.

| Verzeichnis | Bedeutung |
|---|---|
| `open/` | Geplant, noch nicht priorisiert. Keine Garantie auf Umsetzung. |
| `next/` | Als Nächstes priorisiert. Verantwortlicher zugeordnet (`Verantwortlich:`-Feld im Slice-Kopf). |
| `in-progress/` | Beansprucht: Der `git mv` hierher liegt auf dem **Hauptzweig, vor der Arbeit** — Branch/PR entsteht danach. Im Parallelbetrieb liegt er im Slice-Zweig, vor der Arbeit, und gelangt per Merge nach `main` (Übergangsregeln Nr. 3). |
| `done/` | DoD erfüllt, gemerged, Closure-Notiz vorhanden — oder Gegenstand an einen anderen Slice übergegangen oder entfallen: §7 nennt Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer (Baseline-Regelwerk `modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer übernimmt). |

## Slices vs. Wellen — zwei Ablagen, dieselbe Regel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

- **Slices** tragen ihren Zustand über das **Verzeichnis**
  (`open/` → `next/` → `in-progress/` → `done/`).
- Eine **Welle** (Bündel von Slices) ebenso: Der Zustand ist die
  Verzeichnis-Position, kein `Status:`-Feld. Der Welle-Plan (`<welle-id>.md`)
  liegt **flach** in `planning/`, solange die Welle läuft, und wandert bei
  Closure per `git mv` nach `done/` — neben seine
  `welle-<Kennung>-results.md`. Den aktiven Durchlauf `open/` → `next/` →
  `in-progress/` durchläuft er nicht; `done/` ist sein einziges
  Lifecycle-Verzeichnis. **Geplante** Wellen haben noch keine Datei — sie
  stehen in der Roadmap, die auch Sequenzierungs-Autorität bleibt
  ([`in-progress/roadmap.md`](in-progress/roadmap.md): Meilensteine, nächste
  Wellen, Zeiger auf die offenen).
- Der aktive Durchlauf `open/` → `next/` → `in-progress/` nimmt ausschließlich
  **Slices** auf; `done/` archiviert **zusätzlich** abgeschlossene
  **Nicht-Slice-Records** — Welle-Plan und Welle-Closure
  `done/welle-<Kennung>-results.md`. Aufgelöste Carveouts wandern **nicht** hierher,
  sondern in ihr eigenes `docs/plan/carveouts/done/` <!-- d-check:ignore (done/ entsteht erst bei erster Carveout-Auflösung) --> (Baseline-Regelwerk
  `modul-07-carveouts.md`).

Neben den Lifecycle-Verzeichnissen liegt **flach** in `planning/` das
**Beobachtungs-Register** (`observations/`): Es trägt den
Steering-Loop-Zähler, wird bei jeder Slice-Closure fortgeschrieben und
überlebt jede Welle (Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register).

Ein Repo, das aus einem **Brownfield-Bootstrap** kommt, trägt hier flach ein
zweites Register: `reconciliation.md` mit den offenen Funden des Rückbaus. Es
wird beim Auflösen fortgeschrieben und ist leer, sobald alle Sub-Areas
graduiert sind (Baseline-Regelwerk `modul-02-harness-bootstrap.md`
§Das Reconciliation-Register). Greenfield-Repos haben die Datei nicht.

## Aktueller Stand

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — kein Snapshot; der Stand ergibt sich aus den
Verzeichnissen.

Nicht als Snapshot hier eintragen — der Stand ergibt sich aus den
`open/`/`next/`/`in-progress/`/`done/`-Verzeichnissen (optional ein
`plan-status`-Target wie im Kurs-`lab/example`), sonst driftet die Tabelle.

## Parallelbetrieb in Worktrees (Übergangsregeln)

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine (`Verantwortlich:` trägt bei parallelem Lauf den
Zweig; WIP-Limit = 1 pro Lauf). Gilt für **jede Rolle beim Start eines Slice**, ohne
Code-Änderung am Repo, **bis `slice-worktree-parallel` ([welle-sensoren](welle-sensoren.md), Teil 0) in `done/` liegt**; danach löst er die Nr. 1 und 2 ab, die Nr. 3 bis 8 bleiben. Produktive Slices und Wellen laufen jetzt parallel; der Werkzeug-Slice ist kein Produkt-Slice und blockiert keinen.

Der Grund der Sperre: Die Gates bauen Images mit festen Tags (`odata-recorder:test`, `:race`, `:lint`, `:build`, `:gegenproben`, `:groessenmarke` in `harness/mk/*.mk`) im gemeinsamen Docker-Daemon, zwei Läufe aus zwei Arbeitsbäumen überschreiben einander; die Fristen der Gegenproben (120 s je Fall) reißen bei Last, und ein Rot unter Last ist kein Befund.

1. **Schwere Läufe unter einer Sperre.** `make gates`, `make gegenproben`, `make test-race`, `make groessenmarke`, Mutationsläufe in Containern und jeder `docker build` laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`; ein zweiter Lauf wartet. Ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, geht als Hintergrundlauf. Jeder eigene `docker run` trägt `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container je Linie zugleich; vor einem großen Lauf `free -g`. Die Sperre lässt einen schweren Lauf zugleich zu; höchstens zwei Linien warten oder fahren zugleich auf ihr (Host: 20 Kerne, 31 GB), eine dritte Linie wartet auf den Abschluss einer der beiden.
2. **Der Stempel gehört dem Integrationsschritt.** `harness/gegenproben.stempel` wird im Slice-Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Der Integrationsschritt führt nach dem Merge in `main` `make gegenproben` aus (die Eingabe der Summe zweier Zweige ist neu) und committet den Stempel allein; `src` und `tools` sind Eingabe, der Stempel kommt nach dem letzten Eingabe-Commit.
3. **Ein Worktree je Linie, ein Zweig je Slice.** Pfad `/Development/odata-recorder-wt-<kurzname>`, Zweig `slice-<Kennung>` von `main`; eine Linie führt ihre Slices nacheinander (nach dem Merge ein neuer Zweig von `main` im selben Worktree). `Verantwortlich:` im Slice-Kopf lautet `pt9912 (Zweig slice-<Kennung>)`; das WIP-Limit 1 gilt je Lauf. Der Start (`git mv` nach `in-progress/`, `AGENTS.md` §3.3) ist ein eigener Commit (reiner Move) im Slice-Zweig, vor der Arbeit; der Integrationsschritt trägt ihn mit dem Zweig per Merge nach `main`. Der Architect arbeitet vor dem Start im selben Zweig (Vertrag in `spec/` oder `harness/vertraege/`, Tabelle und Raster im Plan).
4. **Was der Slice-Zweig schreibt.** Code, Tests, Spezifikation, Verträge, den eigenen Slice-Plan samt Closure-Notiz in §7 und den `git mv` nach `done/` (eigener Commit, `AGENTS.md` §3.3). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt nur der Integrationsschritt (Planner im Hauptbaum oder im Integrations-Worktree). Ein Zweig, der dort etwas ändern müsste, nennt es in §6 seines Plans mit Ausgang *offen* und Adresse Planner (`AGENTS.md` §3.13).
5. **Merge und Gates.** Nach Verifier-Pass und Closure merged der Integrationsschritt per `git merge --no-ff slice-<Kennung>` in `main`, schreibt das Register und die Roadmap nach und fährt danach `make gates` im Hauptbaum unter der Sperre. Ist `main` seit dem Start weitergelaufen, nimmt der Zweig es vor dem Handoff per `git merge main` auf und fährt `make gates` neu. Die Merge-Reihenfolge entscheidet der erste Verifier-Pass. Solange ein Implementer im Hauptbaum arbeitet (heute `slice-groessenmarke-werkzeug`), committet der Integrationsschritt in einem eigenen Worktree auf einem Zweig `integration` und führt ihn zwischen den Commits des Implementers per `git merge --ff-only` in `main` nach; ab `slice-groessenmarke-werkzeug` in `done/` gehört der Hauptbaum dem Integrationsschritt, und Implementer arbeiten nur in Slice-Worktrees.
6. **Kennungen mit Bereich.** Eine Kennung mit eigener Datei kollidiert still, die in `spezifikation.md` laut erst beim Merge. Nächste freie Kennungen am 2026-10-30: `SPEC-129` (`SPEC-128` ist der letzte Vertrag unter `harness/vertraege/`), `ARC-022`, die ADR-Nummer 0010, `MR-002`, `CO-001`. Jede Linie vergibt nur aus ihrem Bereich; Lücken bleiben vergeben, ein erschöpfter Bereich geht als Frage an den Planner.

   | Linie | `SPEC` | `ARC` | `ADR` | `CO` |
   |---|---|---|---|---|
   | A: Rest von `welle-record-replay` | 129 bis 159 | 022 bis 029 | 0010 bis 0019 | 001 bis 019 |
   | B: Teil 1 von `welle-mock` | 160 bis 199 | 030 bis 039 | 0020 bis 0029 | 020 bis 039 |
   | C: Gate-Kette (`slice-worktree-parallel`, Slices von `welle-sensoren`, `welle-dev-container`) | 200 bis 229 | — | 0030 bis 0039 | 040 bis 059 |
   | D: Port `Telemetry` (`slice-telemetry-port`, Slices von `welle-betrieb` davor und danach) | 230 bis 259 | 040 bis 049 | 0040 bis 0049 | 060 bis 079 |

   `MR-*` vergibt nur Linie C; `LH-*` ändert keine Linie (Sache des Owners).
7. **Die Gate-Kette hat einen Besitzer zugleich.** Slices, die `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` oder den Stempel ändern (Linie C), laufen nacheinander, nie zwei zugleich; ein Slice der Linien A und B ändert sie nicht und meldet es als Rückgabe an den Planner, wenn er es müsste. `slice-worktree-parallel` startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt.
8. **Betriebsregeln.** Die Image-Falle: das Image `odata-recorder:race` trägt `/src/src` eingebacken; ein Lauf, der Quellen hineinkopiert, leert den Pfad zuvor oder nutzt die Stufe `deps`. Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, nur `Edit` und `Write`; Zeichen jenseits von ASCII und Escape-Folgen aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18). `make gates` nie parallel zu Mutationsläufen; ein Rot unter Last wird auf ruhigem Host wiederholt. Ad-hoc-Docker nur mit den Grenzen aus Nr. 1.

Die Welle-Pläne verweisen hierher; wo ein Plan in `open/` unverändert bleibt (Entscheidung des Users vom 2026-10-21), trägt die Welle-Datei den Zeiger.

## Roadmap

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

Siehe [`in-progress/roadmap.md`](in-progress/roadmap.md).

# Review-Report: slice-ci-github-actions (CI mit GitHub Actions und `make workflow-check`) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** der eigene Anteil in `7c9014bb..4afc021c` ohne die Merges von `main`: `.github/workflows/gates.yml`, `.d-check-workflows.yml`, `harness/mk/workflow.mk`, `harness/README.md`, Nachzug in `harness/vertraege/` und Plan.

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`4afc021c`) · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-ci-github-actions`; `SPEC-217`, `SPEC-211`, `SPEC-323`, `SPEC-212`; `LH-QA-06`, `LH-RB-03`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md), [ADR-0030](../plan/adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md) (`Proposed`); `AGENTS.md` §3.

**Urteil: nicht ohne Nacharbeit.** Kein HIGH, 1 MEDIUM, 1 LOW, 3 INFO. Der Workflow entspricht in der gelesenen Datei dem Vertrag (Rechte, Auslöser, Pin, kein Ausdruck `${{ }}`, kein Cache, keine Geheimnisse), und der grüne Lauf führt alle 20 Ziele. Offen ist eine Lücke des neuen Gates, die als Negativ nirgends steht.

**Vorgehen.** Allein lesend gegen GitHub: `gh run list`, `gh run view --log`, `gh api -X GET` (Läufe, Jobs, Annotationen, Einstellungen, Tag und Commit von `actions/checkout`), `git ls-remote`. Kein Rerun, kein Pull Request, keine Einstellung, kein Push. Gegenproben des Gates: `d-check` am Pin von `d-check.mk`, je Fall ein Wegwerfverzeichnis im Scratchpad, nacheinander ein Container (`--network none`, `--memory=4g --cpus=4`), je ein Lauf (Textleser, keine Nebenläufigkeit). `make gegenproben` nicht gefahren.

## Gesehen

**Läufe** (`gh`, Messender: Reviewer; jeder Lauf des Repos ist einer dieser vier, jeder mit `run_attempt` 1):

| Lauf | Commit | Gesehen |
|---|---|---|
| 46163874945 | `62fa14b3` | grün; Schritt `make gates` 31 min 37 s; Rechte `Contents: read`, `Metadata: read`; Abbild `20261122.23.1`, Docker 28.0.4, 4 Kerne, 15 GiB; `is-shallow-repository: false`; alle 20 Ziele von `GATE_CHECKS` im Log (`make -n -k gates` als Liste); `pruefung test: ok` 459,8 s, `pruefung race: ok` 954,0 s, beide nicht `CACHED`; keine Annotation |
| 46167050256 | `89c53716` | rot im Schritt `make gates`, `README.md:67 target-missing`, `docs-check` |
| 46167432658 | `ca2429c3` | grün mit `continue-on-error`, Schritt Exit 2 |
| 46167722016 | `6dfed522` | rot im Schritt *Umgebung* (`is-shallow-repository: true`), `make gates` übersprungen |

`89c53716`, `ca2429c3`, `6dfed522` sind keine Vorfahren von `main`; auf `origin` liegen allein `main` und der Slice-Zweig (bei `62fa14b3`).

**Pin:** `refs/tags/v7.0.1` von `actions/checkout` zeigt als leichtgewichtiger Tag auf `3d3c42e5aac5ba805825da76410c181273ba90b1`; Signatur des Commits verifiziert, der Commit ist Vorfahr von `main` des Repos der Aktion, `releases/latest` nennt `v7.0.1`. Der Commit `e3eb93c6` nennt Tag, Datum und Weg (`SPEC-217` Nr. 2 c).

**Gegenproben des Gates** (eigene Mutationen):

| Mutation | Ergebnis |
|---|---|
| unverändert | Exit 0 |
| `uses:` ungepinnt in den Formen: in Anführungszeichen, Flow-Mapping, Schlüssel in Anführungszeichen, Block-Skalar, Alias auf einen Anker; SHA in Großbuchstaben; leerer Kommentar; hinter `d-check:ignore` in der Zeile und im Block | je rot |
| lokale Referenz ohne Ziel; leere Datei; Tab im YAML; Datei ohne Leserecht; Workflow als Symlink | je rot |
| lokale Composite-Action `./.github/actions/a`, darin `evil/x@v4` | **grün** (F-1) |
| zweites YAML-Dokument hinter `---` mit ungepinntem `uses:` | **grün** (F-1) |
| zweite Datei `zwei.YML`, Referenz auf `@main` | **grün** (F-1) |
| `container: alpine:latest` und `services` mit `docker:dind` | **grün** (F-2) |
| `pull_request_target`, `contents: write`, `run` mit `${{ github.head_ref }}` | grün, benannt in `SPEC-217` Nr. 11 c |
| Kommentar `# irgendwas`; Konfiguration ohne `modules:` | grün, benannt in Plan §6 und `harness/README.md` |
| ungepinnter Workflow und toter Verweis unter `.wt/` | grün (`.wt/` bleibt ohne Wirkung) |

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Zugesagt ist rot bei einer Aktion ohne volle SHA; ein Workflow mit `uses: ./.github/actions/a`, deren `action.yml` `evil/x@v4` ruft, bleibt grün, und weder `SPEC-217` Nr. 11 noch „Nicht behauptet“ noch Plan §3/§6 nennen die Lücke. *Versagen:* eine ungepinnte fremde Aktion läuft auf dem Runner, während das Gate Pins meldet; ebenso grün, ohne geprüft zu haben, ob GitHub die Form ausführt: zweites YAML-Dokument, Datei mit Endung `.YML`. | `AGENTS.md` §3.15, §3.17; `SPEC-217` Nr. 8, 11 | `harness/vertraege/SPEC-217-ci.md` · „Rot wird es bei einer Aktion ohne volle SHA“; Plan §3 Zeile 8 · „jede Aktion mit voller SHA und Tag-Kommentar“ | nein — kein Gate fängt die fehlende Nennung; die Mutation selbst zeigt `make workflow-check` mit Exit 0 | Fänger schmaler als Zusage, Differenz nicht als Negativ |
| F-2 | LOW | Die Aufzählungen dessen, was ohne Fänger bleibt, nennen weder Images im Workflow (`container:`, `services:`) noch Ausdrücke `${{ }}` in `run`; ein Image mit `latest` bleibt grün, und der Rahmen des Plans verlangt gepinnte Images. Die heutige Datei trägt keines von beiden. | `AGENTS.md` §3.15; `SPEC-217` Nr. 11 c | `harness/vertraege/SPEC-217-ci.md` · „*Kein automatischer Fänger* für die Zusagen aus Nr. 2 und 3“; Plan §1 · „Actions und Images sind gepinnt“ | nein | Negativliste unvollständig |
| F-3 | INFO | Der grüne Lauf deckt `62fa14b3`. HEAD trägt danach die Übernahme von `main` (`6b0935bb`, mit `slice-batch-validator` unter `src/`) und ist nicht gepusht; `test-race` ist auf dem Runner an diesem Quellstand ungesehen. Lokal liegt ein Gate-Stempel für `4afc021c`. | `SPEC-217` Nr. 9 a (Zuständigkeit Verifier, Integrationsschritt) | Plan §3 · „Vor dem ersten Push war `make gates` am Stand `62fa14b3` lokal grün“ | ja — Lauf der CI nach dem Push | Beleg vor dem letzten Stand von main |
| F-4 | INFO | Einstellungen auf GitHub, außerhalb des Slice, zur Kenntnis des Users (öffentliches Repo): Freigabe für Pull Requests aus Forks `first_time_contributors` (nach der ersten Freigabe laufen weitere ohne sie, mit nur lesendem Token), `allowed_actions: all`, `sha_pinning_required: false`, `main` ohne Schutz. Keine Geheimnisse, Variablen, eigenen Runner, Umgebungen, Hooks oder Cache-Einträge; Vorgabe-Rechte des Tokens `read`. | Maintainability (Plan §1, `SPEC-217` Nr. 10 c, 11 e) | Plan §6 · „**Branch-Protection.**“ | ja — `gh api` lesend | Einstellung des Dienstes außerhalb des Repos |
| F-5 | INFO | Der Wegwerfzweig ist auf `origin` gelöscht, lokal besteht `slice-ci-github-actions-rot` weiter, darin der Stand mit `continue-on-error` und totem Verweis. | `SPEC-217` Nr. 9 b | Plan §3 · „Zeile 18: der Wegwerfzweig ist auf `origin` gelöscht“ | ja — `git branch` | Wegwerfzweig lokal verblieben |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `.github/workflows/gates.yml` gegen `SPEC-217` Nr. 2 und 3 (ganze Datei) | geprüft, ohne Befund: ein Job, Push auf `**` und `pull_request`, allein `contents: read`, kein `secrets`, kein `pull_request_target`, kein Ausdruck (Zweignamen gelangen in keine Shell-Zeile), `fetch-depth: 0`, `persist-credentials: false`, `ubuntu-24.04`, 90 Minuten, vier Schritte, `if:` allein in *Platz danach*, kein Cache, keine `concurrency` |
| Token im Arbeitsbaum (Log des Checkout) | geprüft, ohne Befund: Verweis und Datei unter `_temp` vor `make gates` entfernt |
| `make gates` auf dem Runner: Vollständigkeit, Variablen aus der Umgebung | geprüft, ohne Befund: 20 von 20 Zielen; `VERSION=dev` und `docker` wie lokal; keine Bezüge mit Anmeldung |
| `.d-check-workflows.yml`, `harness/mk/workflow.mk` | geprüft, ohne Befund außer F-1: netzlos, Mount `:ro`, Pin aus `d-check.mk`, Kommentare in den Klassen von `AGENTS.md` §3.7 |
| `harness/README.md` | geprüft, ohne Befund außer F-1 und F-2: Lauf-Status und Gate-Zeile stimmen mit den Läufen überein |
| `harness/vertraege/` (`SPEC-217`, `SPEC-211` Nr. 5 e, `SPEC-323`, Index) | geprüft, ohne Befund außer F-1 und F-2: Dauern, Maße und Rechte gegen die Logs nachgesehen |
| Plan (§3 Läufe und Job-Logs, §6, Raster) | geprüft, ohne Befund außer den genannten |
| `spec/`, `docs/plan/adr/`, `src/`, `tests/`, `tools/` | vom eigenen Anteil nicht berührt |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 1 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Fänger schmaler als Zusage, Differenz nicht als Negativ · Negativliste unvollständig · Beleg vor dem letzten Stand von main · Einstellung des Dienstes außerhalb des Repos · Wegwerfzweig lokal verblieben

## Verdikt

**Merge-blockierend:** ja, wegen F-1.

**Übergabe:** Findings gehen an den Implementer; die Finding-Klassen zusätzlich in die Closure §7. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation gegen die DoD.

# Slice slice-ci-github-actions: CI mit GitHub Actions — `make gates` auf einem frischen Runner bei Push und Pull Request

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — ein einzelner Slice, dessen Closure-Bedingung seine DoD ist (ein grüner und ein roter Lauf auf dem Runner); kein repo-weiter Beleg geht darüber hinaus (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Er ist eine Bedingung der Erstabnahme (M7, [Roadmap](../in-progress/roadmap.md)) und wird vor ihr umgesetzt (§4). Er ist ein **Kettenslice** (Linie C, Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)): er ändert mindestens `harness/README.md`.

**Bezug:** `LH-QA-06` (Messmethode: containerbasierter CI-Test), `LH-RB-03` (Entwicklung und Ausführung in Docker-basierten CI-Umgebungen), [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Gate-Kette), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (`test-race` ist Gate).

**Berührte Spec-Stellen:** `LH-RB-03.a` (Linux und CI) · `SPEC-004` (Migrationsregel: Satz über CI und seinen Ersatz) · `SPEC-122` (`make test-race`) · [SPEC-211](../../../../harness/vertraege/SPEC-211-container-tests-in-der-kette.md) Nr. 5 e (Abbruch durch TERM, offene Sendung mit Trigger „eine CI“) · `SPEC-323` (Neustarttest, kommt mit [slice-container-image](../done/slice-container-image.md)) · [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) (Vertrag der CI).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-ci-github-actions)

**Autor:** pt9912 (Planner, Auftrag des Users vom 2026-11-28; Umsetzung vor der Erstabnahme nach seiner Entscheidung vom selben Tag, §4). **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Bei Push und Pull Request fährt GitHub Actions `make gates` unverändert auf einem frischen, von GitHub gestellten Linux-Runner und meldet rot, wenn ein Gate rot ist; dieser Lauf ist der containerbasierte CI-Test, den die Messmethode von `LH-QA-06` verlangt und den heute ein Lauf über `make` ersetzt.

**Rahmen** (aus dem Auftrag und den harten Regeln, nicht zur Wahl):

- Docker-only (`AGENTS.md` §3.1): der Workflow installiert keine Toolchain und ruft `make`; die Prüfläufe bleiben netzlos und ohne schreibbaren Mount.
- Actions und Images sind gepinnt (Actions auf die volle Commit-SHA, Images wie heute auf den Digest).
- Keine Geheimnisse; die Rechte des Runner-Tokens sind minimal (nur lesen).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Gate-Stempel und der volle Lauf ohne Stempel (Slice (a) aus [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md)) — anderer Vorgang: die ADR ist `Proposed`, und der Slice ändert `tools/harness/working-tree-hash.sh`, wofür die Zustimmung des Users fehlt.
- Nachtlauf und Hash-Prüfung je Push (Slice (b) aus [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md)) — anderer Vorgang: diese CI ist davon unabhängig und nicht (b), sondern dessen Grundlage ([SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 1; §6 Frage 2). (b) folgt auf die Stempel (a); die Annahme der ADR liegt beim User.
- Gegenproben, Lasttest und Mutationsläufe auf dem Runner — keiner ist Teil von `make gates` ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), `SPEC-216`; [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 10 a, b; §6 Frage 6). Gegenproben in der CI wären ein eigener Vorgang nach einem gültigen grünen Lauf; die CI ersetzt ihre Eichung nicht.
- Änderungen an Gates, Schwellen, Fristen, Tests oder Produktcode, damit der Runner grün wird — eine Lockerung braucht eine ADR (`AGENTS.md` §3.6); ein Fund ist eine Rückführung (§4) oder ein Carveout, kein Beifang.
- Änderungen am Lastenheft — Sache des fachlichen Owners; nach `Accepted` nur über einen Change Request ([Lastenheft](../../../../spec/lastenheft.md) §Historie), davor ohne ihn. §6 stellt die Frage.
- Registry, Veröffentlichung, Release, Ablage von Artefakten — anderer Vorgang; sie brauchen Schreibrechte und Geheimnisse, die dieser Slice ausschließt.
- Einstellungen des Repos auf GitHub (Branch-Protection, Pflicht-Status) — liegen nicht im Repo; Entscheidung des Users (§6).
- Ersatz des lokalen Gate-Laufs vor dem Handoff — `AGENTS.md` §6 und der Stop-Hook bleiben, wie sie sind; die CI kommt hinzu.
- Der Arbeitscontainer für Entwickler in der CI — [welle-dev-container](../welle-dev-container.md); sie schließt ihrerseits eine CI aus (§6 dort).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

Die Regeln der Liefer-Punkte stehen in [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md).

- [ ] **Lauf** (`LH-QA-06`, `LH-RB-03`): der Workflow fährt `make gates` bei Push und Pull Request auf einem frischen Runner, ohne Toolchain-Installation und ohne ausgelassenes Gate; ein grüner Lauf ist belegt mit Commit, Adresse des Laufs, Runner-Label, Dauer und Messendem, darin der Neustarttest (`test-image`, `SPEC-323`) und `test-race`. Der Auslöser `pull_request` steht im Workflow; seinen Lauf belegt nur ein Pull Request, den der User bisher nicht erlaubt hat (§6). Ohne die Erlaubnis steht er als benanntes Negativ (Nr. 9 d; §3 Zeile 2).
- [ ] **Rot ist sichtbar** (`AGENTS.md` §3.10, §3.15): je Zusage und Randform des Vertrags ein Fänger oder ein benanntes Negativ mit Grund und Folge (Nr. 11; §3); mindestens ein auf dem Runner gesehener Lauf mit absichtlich gebrochenem Gate auf einem Wegwerfzweig ist rot, aus dem benannten Grund, und der gebrochene Stand gelangt nicht nach `main` (Nr. 9 b). Den Push des Wegwerfzweigs hat der User am 2026-11-28 erlaubt (§6).
- [ ] **Pins und Rechte:** jede Action auf der vollen Commit-SHA, Runner-Label ohne `latest`, Rechte des Tokens nur lesend, kein Geheimnis, kein Token im Arbeitsbaum nach dem Checkout, kein Auslöser, der fremdem Code Schreibrechte oder Geheimnisse gibt; je Zusage ein Träger (Prüfung oder benanntes Negativ mit Grund).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: jede Stelle, die „keine CI“ oder den Ersatz durch `make` trägt, zieht im liefernden Commit nach (`AGENTS.md` §3.9; Liste in §3).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.github/workflows/gates.yml` | neu | der Workflow nach [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 2 und 3 |
| `.d-check-workflows.yml` (Wurzel) | neu | Konfiguration allein für das Modul `workflows` (Nr. 8) |
| Fragment `workflow.mk` unter `harness/mk/` | neu | Ziel `workflow-check`, an `GATE_CHECKS` (Nr. 8); im selben Commit wie der Workflow |
| `harness/README.md` | update | Lauf-Status, Zeile des Gates, „Nicht behauptet“ (Nr. 11) |
| `harness/vertraege/` (`SPEC-217` Stand, `SPEC-211` Nr. 5 e, `SPEC-323` Festlegung, Index) | update | „nicht geliefert“ und „es gibt keine CI“ nachziehen (Nr. 13) |
| `Makefile`, `Dockerfile`, `.d-check.yml`, `tools/`, `src/`, `tests/`, `spec/` | unverändert | der Vertrag braucht keine Änderung daran; `SPEC-004` ist mit dem Vertrag nachgezogen |

Weitere Stellen, die „keine CI“ nennen und nur der Planner im Integrationsschritt ändert (Übergangsregeln Nr. 4): [welle-dev-container](../welle-dev-container.md) §6, [welle-betrieb](../welle-betrieb.md) §2 (historische Bedingung, falls die Welle dann noch offen ist). `docs/plan/planning/rueckverfolgung.md` ist ein datierter Auslesestand. Dateien in `done/` bleiben. [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md) ist mit dem Vertrag nachgeführt (Frage 2).
Die Pflicht, nach jedem Push auf `main` den Lauf zu lesen (Nr. 11 e), steht in Nr. 5 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln); der Slice ändert sie nicht.
Die Liste ist beim Start neu gesucht (Planner, 2026-11-28, `main` bei `dfb51e47`, durch Suchen nach `CI`): der Satz „eine CI gibt es noch nicht“ steht in `harness/README.md` unter *Aktueller Lauf-Status*; `SPEC-211` Nr. 5 e trägt noch „(es gibt keine CI)“; `SPEC-217` (*Stand*), die Zeile des Index unter `harness/vertraege/` und die Festlegung von `SPEC-323` tragen „noch nicht geliefert“ oder „ohne CI“. `SPEC-323` führt „ohne CI“ außerdem im Grund der Entscheidung unter Nr. 1; ob der Satz bleibt, entscheidet die Lieferung mit Blick auf `AGENTS.md` §3.17. Unverändert bleiben: `SPEC-004` (der Satz zum Ersatz bleibt, `SPEC-217` Nr. 13), das Lastenheft und [slice-dev-container-lauf](../open/slice-dev-container-lauf.md) (schließt eine CI aus). Der Kopf von `.claude/hooks/stop-require-gates.sh` nennt heute schon „CI auf dem Push“ als Netz; der Slice ändert den Hook nicht.

**Reihenfolge der Umsetzung:** (1) Vorprobe lokal ([SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 5 c), (2) Workflow, Konfiguration und `workflow-check` in einem Commit, Mutationen der Zeilen 8 und 10 lokal, (3) Texte, (4) die Läufe nach Nr. 9: der grüne Lauf auf dem Slice-Zweig und die Mutationen der Zeilen 12 und 5 auf dem Wegwerfzweig `slice-ci-github-actions-rot`; beide Pushes sind erlaubt (§4 Nr. 5, §6).

**Zusage · Fänger** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3; Nummern „Nr.“ im Text meinen `SPEC-217`):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Lauf bei Push auf jeden Zweig (Nr. 2 a) | Beleg: Lauf zum Push des Slice-Zweigs (Nr. 9 a) | `branches` allein `main`: kein Lauf zum Zweig | andere |
| 2 | Lauf bei Pull Request, am Merge-Stand (Nr. 2 a, 3 c) | Beleg Nr. 9 d; ohne Zustimmung akzeptiertes Negativ (kein Ablauf mit Pull Requests; Folge: der Auslöser ist ungesehen) | `pull_request` entfernt | andere |
| 3 | keine weiteren Auslöser, keine Pfadfilter (Nr. 2 a) | akzeptiertes Negativ Nr. 11 c, Review | `paths-ignore` ergänzt: bleibt grün | andere |
| 4 | ein Job, `make gates` unverändert, jedes Gate (Nr. 2 f) | Beleg: das Log des grünen Laufs führt die Rezepte aller Gates aus `make -n gates` (Verifier); danach Nr. 11 c | der Schritt ruft `make docs-check` | andere |
| 5 | rotes Gate macht den Job rot, aus dem benannten Grund (Nr. 2 f, 9 b) | Lauf auf dem Wegwerfzweig | toter Verweis; dazu `continue-on-error: true` am Schritt (der Lauf wäre grün) | andere |
| 6 | Rechte allein `contents: read` (Nr. 2 b) | Beleg Nr. 9 c; danach Nr. 11 c | `contents: write`: `workflow-check` bleibt grün (gemessen), Review | andere |
| 7 | kein Zugriff auf `secrets` (Nr. 2 b) | akzeptiertes Negativ Nr. 11 c, Review | ein `env` mit einem Geheimnis: bleibt grün | andere |
| 8 | in der Prüfmenge (Nr. 8) jede fremde Referenz mit voller SHA und Kommentar, gültiges YAML (Nr. 2 c, 8) | `make workflow-check` | `@v4`; SHA ohne Kommentar; SHA mit 39 Zeichen; kaputtes YAML | andere |
| 9 | allein `actions/checkout`, keine lokale Aktion (Nr. 2 c) | akzeptiertes Negativ Nr. 11 c, Review | zweite, gepinnte Aktion: bleibt grün | andere |
| 10 | leere Prüfmenge ist rot (Nr. 8) | `make workflow-check` | Workflow-Datei entfernt; `uses:` entfernt | andere |
| 11 | kein Token im Arbeitsbaum (Nr. 3 a) | akzeptiertes Negativ Nr. 11 c; einmal das Log des Checkout | `persist-credentials` entfernt: bleibt grün | andere |
| 12 | volle Historie (Nr. 3 a, d) | Schritt *Umgebung* in jedem Lauf | `fetch-depth` entfernt: Lauf rot im Schritt *Umgebung* | andere |
| 13 | Runner `ubuntu-24.04`, Zeitgrenze 90 Minuten (Nr. 2 d, e) | Log nennt das Abbild; danach Nr. 11 c | `ubuntu-latest`: bleibt grün | andere |
| 14 | kein Cache, keine `concurrency`, kein automatischer Neustart (Nr. 4 d, 2 g, 6 b) | akzeptiertes Negativ Nr. 11 c; einmal im Log: `test-race` nicht aus dem Cache | `actions/cache` oder `concurrency` ergänzt: bleibt grün | andere |
| 15 | Umgebung je Lauf im Log (Nr. 2 f, 4 b) | Log des grünen Laufs | Schritt entfernt: bleibt grün | andere |
| 16 | `workflow-check` in `make gates` und im Gate-Index (Nr. 8) | `make docs-check` (Zeile des Ziels); `GATE_CHECKS`: akzeptiertes Negativ Nr. 11 d, Verifier prüft `make -n gates` | Zeile im Index entfernt; Zeile `GATE_CHECKS` entfernt | andere |
| 17 | `test-race` und Neustarttest im grünen Lauf, keine Lockerung dafür (Nr. 5 d, 9 a) | Beleg Nr. 9 a; Diff ohne Änderung an Gates und Fristen (Review) | Gate per Variable ausgelassen | andere |
| 18 | der gebrochene Stand gelangt nicht nach `main` (Nr. 9 b) | Verifier: der Commit ist kein Vorfahr von `main` | Wegwerfzweig gemergt | andere |
| 19 | Texte nachgezogen (Nr. 13) | Review (`AGENTS.md` §3.9) | Satz „es gibt keine CI“ bleibt stehen | andere |
| 20 | kein `container:`, keine `services:`, kein `${{ }}` in `run` (Nr. 2 f) | akzeptiertes Negativ Nr. 11 c, Review | `container: alpine:latest` mit `services`: bleibt grün (Reviewer); `run` mit `${{ github.head_ref }}`: bleibt grün (Reviewer) | andere |
| 21 | außerhalb der Prüfmenge (lokale Aktion, zweites Dokument, `.YML`, Inhalt des Kommentars) keine ungepinnte oder falsch benannte Referenz (Nr. 8) | akzeptiertes Negativ Nr. 11 h, Review | je Weg eine: bleibt grün (Reviewer, Kommentar auch Implementer) | andere |

**Maß der Negative:** von 21 Zeilen fängt dauerhaft ein Gate die Zeilen 8 und 10 und die Hälfte von 16, ein Schritt jedes Laufs die Zeile 12; die Zeilen 1, 2, 4, 5, 15, 17 und 18 sind einmal am Beleg gesehen; die Zeilen 3, 6, 7, 9, 11, 13, 14, 19, 20 und 21 hält nach der Lieferung allein das Review.

**Lokal gesehene Mutationen** (Implementer, 2026-11-28; je Fall ein Wegwerfverzeichnis im Scratchpad mit dem echten Rezept, `d-check` am Pin von `d-check.mk`; ohne Detektor und ohne Nebenläufigkeit, je ein Lauf, das Gate ist ein Textleser). Der Exit ist der von `make`.

| Zeile | Mutation | Gesehen |
|---|---|---|
| — | unverändert | Exit 0, 0 Befunde |
| 8 | `@v7`, `@v7.0.1`, `@v4`, `@main` | rot, je `uses-pin-missing` |
| 8 | SHA mit 39, mit 41 und mit 7 Zeichen; `docker://alpine:3` | rot, je `uses-pin-missing` |
| 8 | SHA ohne Kommentar | rot, `uses-pin-untagged` |
| 8 | kaputtes YAML | rot, `workflow-unparsable` (dazu `uses-pin-missing` für die leere Prüfmenge) |
| 8 | Stelle: ungepinntes `uses:` als zweiter Schritt, in einer zweiten Datei (`.yaml`), am Job (wiederverwendeter Workflow) | rot, je `uses-pin-missing` mit Datei und Zeile |
| 10 | Datei entfernt (daneben eine `.md`), Verzeichnis `.github` entfernt, Verzeichnis leer, `uses:` entfernt | rot, je `leere Pruefmenge` |
| 8 | `workflows.dir` der Konfiguration falsch | rot, `leere Pruefmenge` |
| 12 | `run`-Block des Schritts *Umgebung*, aus der Datei gelesen, mit `bash -e` in einem flachen (`--depth 1`) und in einem vollen Klon | flach Exit 1, voll Exit 0; auf dem Runner rot gesehen (*Läufe auf dem Runner*) |
| 16 | Zeile des Ziels im Index von `harness/README.md` entfernt | `make docs-check` rot, `gate-undocumented` |
| 16 | Zeile `GATE_CHECKS += workflow-check` entfernt | `make docs-check` grün (Negativ Nr. 11 d); das Rezept verschwindet aus `make -n -k gates` (1 Treffer zu 0) |
| 3, 6, 7, 9, 11, 13, 14 | `paths-ignore`; `contents: write`; `env` mit `secrets`; zweite gepinnte Aktion mit eigener SHA; `persist-credentials` entfernt; `ubuntu-latest`; `concurrency` | `workflow-check` grün, wie im Plan |
| 1, 2, 4, 5, 12, 15 | `branches: [main]`; `pull_request` entfernt; `make docs-check` statt `make gates`; `continue-on-error`; `fetch-depth` entfernt; Schritt *Umgebung* entfernt | `workflow-check` grün; Fänger ist der Lauf bzw. das Review |

Zwei Grenzen von `workflow-check`, als Negative im Vertrag (`SPEC-217` Nr. 11 d, h): ein Kommentar hinter der SHA, der kein Tag ist (`# irgendwas`), bleibt grün; eine Konfiguration ohne die Zeile `modules:` bleibt grün. Zusätzlich meldet `d-check` zwei verschiedene Kommentare zu derselben SHA (`uses-pin-tag-conflict`); der Vertrag sagt das nicht zu.
`make -n gates` ohne `-k` endet am Pin dieses Repos mit Exit 2 beim Rezept von `schema-check` (seine Zeilen mit `make` laufen auch im Trockenlauf) und nennt die Rezepte danach nicht; für Zeile 4 und 16 ist `make -n -k gates` zu lesen.

**Pin der Aktion** (Nr. 2 c): `actions/checkout` `v7.0.1`, SHA `3d3c42e5aac5ba805825da76410c181273ba90b1`, aufgelöst am 2026-11-28 mit `git ls-remote --tags` am öffentlichen Repo `actions/checkout` (ohne Anmeldung, ohne Git-Konfiguration des Nutzers); der Tag ist leichtgewichtig und zeigt auf den Commit, `releases/latest` nennt ihn als neuestes Release.

**Vorprobe** ([SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 5 c; Implementer, 2026-11-28, ein Lauf): `go test -race -count=1 -trimpath -buildvcs=false ./...` mit cgo im Image von `RACE_IMAGE`, Quellstand `7dc8b3e7`, ein Container mit 4 CPUs, `--memory=4g --memory-swap=4g`, ohne Netz, leerer Bau-Cache, unter der Sperre der schweren Läufe, mit eigenem Image-Tag; Host mit 20 Kernen, Load Average (1 Minute) 3,7 beim Start und 4,3 am Ende, daneben liefen Container anderer Arbeiten. **Urteil: grün**, jedes Paket mit Tests `ok`, 733 s. Über 100 s je Paket: `src/bootstrap` **519,6 s**, `hexagon/model` 223,3 s, `driven/odata/csdlv2` 182,1 s, `driving/admin` 161,2 s, `driven/odata/csdl` 131,9 s, `tests/lasttest` 115,6 s. `src/bootstrap` liegt über der Schwelle von 450 s und bei 87 % der zehn Minuten je Paket; die Frage geht vor dem Push an den User (§6). Der Wert ist eine Größenordnung unter dieser Last, kein fester Wert.

**Zweite Vorprobe** (derselbe Aufruf, derselbe Messende, ein Lauf) am Stand `4dbaea61`, nach der Übernahme von `main` mit `slice-batch-parser`: Load Average (1 Minute) 9,2 beim Start und 3,3 am Ende, **grün**, 721 s. `src/bootstrap` **513,4 s**, `hexagon/model` 224,7 s, `driven/odata/csdlv2` 178,7 s, `driving/admin` 162,5 s, `driven/odata/csdl` 128,6 s, `tests/lasttest` 110,2 s; das neue Paket `driven/odata/batchcodec` 15,4 s.

**Gate-Lauf** (Implementer, 2026-11-28): `make gates` Exit 0 am Stand `d3d07f19`, nach der zweiten Übernahme von `main`. Die Stufen `build`, `test`, `race` und `lint` kamen darin aus dem Schicht-Cache, weil `src/` und `tests/` gleich dem Stand von `main` sind (`BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache`); ein neuer Lauf unter dem Detektor an diesem Quellstand ist die zweite Vorprobe darüber, außerhalb des Gates.

**Läufe auf dem Runner** ([SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 9; 2026-11-28; Messender: der Implementer, über die öffentliche API ohne Anmeldung; Zeiten UTC; jeder Auslöser `push`, Runner-Label `ubuntu-24.04`; jeder Lauf der genannten Commits ist genannt). Vor dem ersten Push war `make gates` am Stand `62fa14b3` lokal grün.

| Zeile | Commit und Zweig | Lauf | Ergebnis |
|---|---|---|---|
| 1, 4, 17 | `62fa14b3`, `slice-ci-github-actions` | [46163874945](https://github.com/pt9912/odata-recorder/actions/runs/46163874945) | **grün**; Job 31 min 47 s (16:12:15 bis 16:44:02), darin Schritt `make gates` 31 min 37 s, unter den 60 Minuten aus Nr. 2 e; ein Schritt `make gates`, Schritte *Umgebung* und *Platz danach* grün |
| 5 | `89c53716`, `slice-ci-github-actions-rot`: toter Verweis in `README.md` | [46167050256](https://github.com/pt9912/odata-recorder/actions/runs/46167050256) | **rot** im Schritt `make gates` nach 5 s, Exit 2; derselbe Stand ist lokal in `make docs-check` rot (`README.md:67`, `target-missing`) |
| 5 (Mutation) | `ca2429c3`, dazu `continue-on-error: true` | [46167432658](https://github.com/pt9912/odata-recorder/actions/runs/46167432658) | **grün**, obwohl der Schritt mit Exit 2 endet (Anmerkung des Laufs): ohne die Zeile ist der Job rot |
| 12 | `6dfed522`, beides zurück, `fetch-depth: 0` entfernt | [46167722016](https://github.com/pt9912/odata-recorder/actions/runs/46167722016) | **rot** im Schritt *Umgebung*, Exit 1; `make gates` übersprungen, *Platz danach* gelaufen |

Zeile 18: der Wegwerfzweig ist auf `origin` gelöscht; `89c53716`, `ca2429c3` und `6dfed522` sind keine Vorfahren von `main` (`git merge-base --is-ancestor`). Zeile 2: kein Lauf zu einem Pull Request, benanntes Negativ (Nr. 9 d).

**Aus den Job-Logs** (Implementer, 2026-11-28, `gh run view <Lauf> --log`, allein lesend; der User hat das Lesen mit `gh` erlaubt, übermittelt über den Orchestrator); am grünen Lauf, soweit nicht anders genannt:

- *Umgebung* (Nr. 9 a, Zeile 15): Runner `2.338.2`, Abbild `ubuntu-24.04` Version `20261122.23.1` (Ubuntu 24.04.5), Docker 28.0.4 (Client und Server, containerd `v2.4.2`, runc 1.5.3), `nproc` 4, 15 GiB Speicher und 3 GiB Swap, `df /` 145 G, davon 85 G frei vor und 79 G frei nach `make gates` (`docker system df`: 22 Images, 8,6 GB; Build-Cache 335,4 MB).
- *Rechte* (Nr. 9 c, Zeile 6): *GITHUB_TOKEN Permissions* nennt `Contents: read` und `Metadata: read`.
- *Checkout* (Zeile 11): `persist-credentials: false`; die Aktion legt das Token in einer Datei außerhalb des Arbeitsbaums ab, entfernt den Verweis darauf und den Header wieder (*Removing auth*).
- *Gates* (Zeile 4, 17): alle 20 Ziele von `GATE_CHECKS` stehen im Log (`lint-profil` ohne Echo, am Bezug seines Images zu sehen), `test-image` mit *zwei Bauten, Binary byte-gleich* und beiden Phasen grün; die Images sind ohne Anmeldung bezogen. Keine Abkündigungs-Meldung und keine Warnung des Dienstes.
- *Detektor und Cache* (Zeile 14): die Stufe `race` lief neu, `pruefung race: ok` nach 954,0 s (lokal in den Vorproben 733 s und 721 s mit 4 CPUs); die Stufe `test` 459,8 s. Eine Zeit je Paket nennt das Log nicht, der Abstand zu den zehn Minuten je Paket ist damit unbekannt.
- *Rot aus dem benannten Grund* (Zeile 5): in `46167050256` und `46167432658` meldet `d-check` `README.md:67`, 1 Befund, `make: *** [d-check.mk:21: docs-check] Error 1`; die Gates davor sind grün.
- *Flacher Klon* (Zeile 12): in `46167722016` `is-shallow-repository: true`, Exit 1 im Schritt *Umgebung*; in den anderen Läufen `false`.

**Größe (Einschätzung des Architect für §4):** ohne die Verträge rund 100 Zeilen (Workflow 40, Konfiguration und Fragment 30, `harness/README.md` und Verweise 30; nicht gemessen), weit unter der Schnittgrenze nach `SPEC-128` Nr. 1. Die Fänger sind überwiegend Läufe, Ausgaben eines Gates und Review; nach der Geltungsgrenze von `SPEC-128` Nr. 2 führt der Slice keinen Block. Der Aufwand liegt nicht in den Zeilen, sondern in den Läufen auf dem Runner und in dem, was sie zeigen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`), alle Bedingungen zugleich:

1. Der Slice wird vor der Erstabnahme umgesetzt; ein Status `Accepted` des Lastenhefts ist keine Bedingung. **Erfüllt** (Entscheidung des Users vom 2026-11-28, übermittelt über den Orchestrator: auf „Soll die CI vor Accepted kommen?“ die Antwort „ja“).
2. [slice-container-image](../done/slice-container-image.md) liegt in `done/` (Vorgänger in der Gate-Kette: erst mit ihm trägt `make gates` den Neustarttest, den die Messmethode von `LH-QA-06` verlangt).
3. Kein anderer Kettenslice liegt in `in-progress/`. Die Stellung zu den übrigen offenen Kettenslices legt der Start fest.
4. Die Fragen 1 bis 13 in §6 sind beantwortet; eine Antwort, die den Schnitt ändert, ändert den Plan vor dem Start. Vertrag, Raster und Tabelle *Zusage · Fänger* stehen vor dem ersten Code-Commit. **Erfüllt** (Architect, 2026-11-28: [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md), Raster in §6, Tabelle in §3; keine Antwort ändert den Schnitt). Der Slice-Zweig entsteht aus einem Stand, der den Vertrag trägt.
5. Der User hat zugestimmt, dass Zweige dieses Slice nach GitHub gepusht werden: ohne Push gibt es keinen Lauf und keinen Beleg. Gepusht wird in diesem Repo nur auf Anweisung. **Erfüllt** (User, 2026-11-28, übermittelt über den Orchestrator: „ja erlaubt“). Die Erlaubnis gilt allein für Zweige dieses Slice. Den Wegwerfzweig für den roten Lauf hat der User am 2026-11-28 erlaubt (§6); ein Entwurfs-Pull-Request und Einstellungen des Repos auf GitHub sind nicht erlaubt. Beides ist keine Start-Bedingung: der Wegwerfzweig bindet den roten Lauf (DoD, §5), der Pull Request allein den Beleg des Auslösers.

**Stand beim Start** (Planner, 2026-11-28, `main` bei `dfb51e47`, durch Lesen und Suchen, ohne Testlauf): die Bedingungen 1 bis 5 sind erfüllt.

- *Kette (Bedingung 3, Nr. 7 der Übergangsregeln):* in `in-progress/` liegen [slice-batch-parser](../done/slice-batch-parser.md) und [slice-betrieb-spans-nachweis](../done/slice-betrieb-spans-nachweis.md), keiner ein Kettenslice. Stellung: dieser Slice hält die Gate-Kette ab dem Start und läuft vor den übrigen offenen Kettenslices; keiner von ihnen startet, solange er in `in-progress/` liegt. Geht er blockiert nach `open/` (Rückführung unten), ist die Kette wieder frei.
- *Kollision:* keiner der beiden ändert eine Datei aus §3. Der Zweig des ersten ändert Pfade unter `src/`, das Review und seinen Plan; der zweite hat im Arbeitsbaum allein `src/bootstrap/telemetrie_betrieb_test.go` (noch nicht committet). Ihre Tests laufen nach dem Merge im CI-Lauf mit; der zweite verlängert mit einem Test in `src/bootstrap` die knappste Stelle (`test-race`, Zeit nicht gemessen). Mergt eine der Linien vor dem Handoff, nimmt der Zweig `main` auf, und der Beleg gilt am Stand danach (Nr. 5 der Übergangsregeln).
- *Gates:* `GATE_CHECKS` führt 19 Ziele (aus `make` ausgelesen, nicht gefahren), mit `workflow-check` 20; `gates` endet dazu auf `record-gates`.
- *Laufzeit:* zur Dauer von `make gates` gilt die Angabe in `SPEC-217` Nr. 2 e (1147 s und 1170 s, 20 Kerne); beim Start nicht neu gemessen.
- *Stempel der Gegenproben:* keine Datei aus §3 steht in den Eingaben von `tools/harness/gegenproben-eingabe.sh`.

**Vor dem ersten Push:** die Fragen zu *Images*, zu *Repo und Runner*, zum Wegwerfzweig und zur Paketlaufzeit nach der Vorprobe (`SPEC-217` Nr. 5 c) sind beantwortet (User, 2026-11-28, §6): gepusht wird als Probe auf dem Standard-Runner.

**Größenmarke:** keine, ohne Block (Planner, 2026-11-28). Grund ist die Geltungsgrenze von `SPEC-128` Nr. 2: die Fänger der Tabelle in §3 sind Läufe auf dem Runner, Ausgaben eines Gates, einmalige Belege und Review; keine der 19 Zeilen trägt einen Test mit Code (Go-Test, Prüflauf-Skript, Fall der Gegenprobe). Der Slice liefert keinen Messpunkt, `make groessenmarke` misst ihn nicht. Kommt Code mit Tests hinzu (etwa ein Skript nach dem Trigger in `SPEC-217` Nr. 8), setzt der Planner vor dessen erstem Commit die Marke.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): grün wird der Runner nur mit Änderungen an Produktcode, Tests, Fristen oder Gate-Rezepten über das hinaus, was der Vertrag nennt — diese Änderung ist ein eigener Slice. Ohne Marke gibt es keine Rückführung über eine Zeilenzahl.
- `in-progress` → `open` (blockiert — Carveout?): ein Gate ist nur auf dem Runner rot (Carveout mit Trigger und Folge-Slice, nie ein stilles Rot, ein übersprungenes Gate oder ein automatischer Wiederholungslauf); oder eine Entscheidung des Users steht aus (Wegwerfzweig, Kosten, Zugriff auf Images) und es bleibt keine Arbeit ohne sie.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, darin auf GitHub ein grüner Lauf am Stand des Slice-Zweigs und ein absichtlich roter auf einem Wegwerfzweig davon (`SPEC-217` Nr. 9 a, b), §7 nennt beide mit Commit und Adresse; Review und Verifikation ohne offenen Befund über LOW; Closure-Notiz mit Lerneintrag.
Den ersten Lauf auf `main` trägt der Integrationsschritt nach dem Merge und dem Push nach; ist er rot, ist das ein Befund mit Carveout oder Folge-Slice, kein Nachtrag im Stillen.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

### Fragen an den Architect (alle beantwortet, 2026-11-28)

Die Regeln stehen in [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md); hier steht je Frage die Antwort und die Nummer.

1. **ADR.** Braucht die CI eine ADR (neuer externer Dienst, zweiter Ort der Gate-Kette, Pins und Rechte als dauerhafte Setzung), oder genügt der Vertrag unter `harness/vertraege/`? — **Ausgang:** beantwortet: keine ADR, der Vertrag genügt; Dienst und Zeitpunkt hat der User entschieden, keine angenommene ADR ändert sich, keine Schwelle sinkt. Wann doch eine ADR fällig wird, steht in Nr. 1.
2. **Verhältnis zu [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md)** (`Proposed`). Sie sieht „CI mit Nachtlauf und Hash-Prüfung“ als Folge-Slice (b) nach den Gate-Stempeln (a) vor. Braucht diese CI Stempel oder Nachtlauf, oder nicht? Ist dieser Slice (b), ein Vorläufer davon, oder davon unabhängig, und muss der Entwurf angepasst werden? — **Ausgang:** beantwortet: unabhängig. Die CI braucht weder Stempel noch Nachtlauf und ist nicht (b), sondern dessen Grundlage (Nr. 1). Der Kontext von [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md) ist nachgeführt, samt der Bedingung, dass Stempel in der CI kein Gate auslassen. Die ADR bleibt `Proposed`; ihre Annahme steht beim User.
3. **Laufzeit und Aufbau des Jobs.** Ein voller Lauf dauert lokal rund 20 Minuten (Angabe des Auftrags; belegt zum Beispiel 1033 s frisch, [slice-admin-inspektion](../done/slice-admin-inspektion.md) §7, Host mit 20 Kernen); ein Runner mit weniger Kernen braucht eher länger, gemessen ist das nicht. Ein Job oder mehrere, `make` mit oder ohne `-j` (Speicher: die Container-Tests laufen dann zugleich), welche Zeitgrenze des Jobs? Bleibt bei einer Aufteilung sichtbar, dass kein Gate aus `GATE_CHECKS` fehlt? — **Ausgang:** beantwortet: ein Job, ein `make gates` ohne `-j` und `-k`, Zeitgrenze 90 Minuten; ohne Aufteilung gibt es keine zweite Liste der Gates (Nr. 2 e, f).
4. **Docker des Runners.** `test-image` startet Container im Daemon des Hosts (`SPEC-323`), der hier der Daemon des Runners ist; welche Voraussetzungen setzen die Rezepte (Plattform-Argumente, Bau-Optionen)? Das Image des Runners und sein Docker sind nicht auf einen Digest pinnbar: benanntes Negativ oder anderer Weg? — **Ausgang:** beantwortet: Voraussetzungen in Nr. 4 a (BuildKit, Speicher- und Swap-Limit, `container:`-Netz, `tmpfs`, `docker export`); Abbild und Docker bleiben ungepinnt, benanntes Negativ, das Log nennt sie (Nr. 4 b, 11 b).
5. **Netz.** Netz brauchen die Stufe `deps` (Module) und der Bezug der Images (Docker Hub und `ghcr.io`). Wie verhält sich der Lauf bei Bezugsgrenzen oder Ausfall? Sind alle gepinnten Images ohne Anmeldung beziehbar (nicht geprüft)? Wäre eine Anmeldung nötig, steht sie gegen „keine Geheimnisse“ und geht an den User. — **Ausgang:** beantwortet: keine Anmeldung; Netzfehler und Bezugsgrenze sind rot, ohne automatischen Neustart (Nr. 4 c). Die Images sind ohne Anmeldung beziehbar (User, unten; grüner Lauf ohne Anmeldung, §3).
6. **Gegenproben.** Fährt die CI `make gegenproben` (je Push, geplant, gar nicht)? Die Fälle reißen ihre Frist von 120 s unter der eigenen Last schon auf dem Entwicklungs-Host (`BEO-REPO/gegenproben-erzeugen-die-last-selbst`; [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)). — **Ausgang:** beantwortet: nein. Die CI ersetzt die Eichung nicht und setzt sie nicht voraus; Gegenproben in der CI wären ein eigener Vorgang nach einem gültigen grünen Lauf (Nr. 10 a).
7. **Paketlaufzeit unter dem Detektor.** `test-race` hat zehn Minuten je Paket; `src/bootstrap` brauchte zuletzt 415,9 s bei 4 CPUs (`BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze`, ein Lauf). Reicht ein kleiner Runner? Wenn nicht: Welcher Weg ist ohne Lockerung offen? — **Ausgang:** beantwortet, soweit ohne Runner möglich: Vorprobe mit 4 CPUs vor dem ersten Push; keine Anhebung der Zeitgrenze, kein ausgenommenes Paket; offene Wege ohne Lockerung in Nr. 5 d, jeder ein eigener Vorgang. Ob der Runner reicht, ist ungeprüft (Nr. 5 a, b).
8. **Flackern unter Last.** Tests und Stufen mit Frist werden unter Last rot und im nächsten Lauf grün (`BEO-REPO/test-fremden-pakets-flackert-unter-last`). Gilt jedes Rot als Befund, und wer darf neu starten? — **Ausgang:** beantwortet: jedes Rot ist ein Befund; von Hand neu starten dürfen der Planner im Integrationsschritt und der User; der Beleg nennt jeden Lauf (Nr. 6).
9. **Cache.** Läuft jeder Lauf ohne Cache der Docker-Schichten, oder wird er gehalten? Mit Cache kommen `test` und `test-race` bei gleichem Inhalt aus dem Cache und sind kein neuer Lauf (`BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache`). — **Ausgang:** beantwortet: kein Cache über Läufe (Nr. 4 d).
10. **Checkout und geprüfter Stand.** `make groessenmarke` braucht die Historie bis zur `basis` eines Plans in Arbeit (`SPEC-128` Nr. 7); ein flacher Klon endet dort rot. Bei Pull Requests: prüft der Lauf den Kopf des Zweigs oder den Merge mit dem Ziel, und was misst die Größenmarke dann? — **Ausgang:** beantwortet: volle Historie; bei Pull Requests der Merge-Stand, die Marke zählt dort wie lokal nach `git merge main` (Nr. 3).
11. **Auslöser, Doppel- und Abbruchläufe.** Welche Zweige bei Push; Pull Requests aus Forks; Push und Pull Request desselben Standes; Abbruch eines überholten Laufs. Übernommene Sendung (`AGENTS.md` §3.13): [SPEC-211](../../../../harness/vertraege/SPEC-211-container-tests-in-der-kette.md) Nr. 5 e lässt offen, ob Rezept oder Treiber auf TERM allein an `make` reagieren, Adresse Planner, Trigger „eine CI oder ein `timeout` um `make`“. Dieser Slice nimmt sie an: der Vertrag entscheidet, welches Signal Abbruch und Zeitgrenze des Jobs senden, und ob eine Änderung folgt oder das Negativ mit neuem Grund bleibt. — **Ausgang:** beantwortet: Push auf jeden Zweig und jeder Pull Request, Forks ohne Geheimnisse und nur lesend, Doppelläufe hingenommen, kein Abbruch überholter Läufe (Nr. 2 a, g). Keine Änderung an Rezept oder Treiber: der Runner wird nach dem Job verworfen (Nr. 7, `SPEC-211` Nr. 5 e nachgezogen); die Signale des Dienstes sind nicht geprüft.
12. **Fänger des Workflows.** Womit wird rot, wenn Pin, Rechte, Auslöser oder der Aufruf von `make gates` sich ändern? Der Workflow läuft nicht lokal; ein Beweis auf dem Runner braucht Push. d-check führt ein Modul `workflows`, das hier nicht aktiv ist (Eignung nicht geprüft). Wer hebt die Pins der Actions an, und wann? — **Ausgang:** beantwortet: `make workflow-check` (das Modul `workflows` mit eigener Konfiguration; am Pin gemessen, Nr. 8) fängt Pins und leere Prüfmenge; Rechte, Auslöser und Aufruf fängt allein das Review (Nr. 11 c). Pins hebt ein bewusster Commit, Adresse Planner (Nr. 2 c).
13. **Texte der Spezifikation.** `SPEC-004` sagt, die CI muss auch die Migrationskette auf einer Kopie bestehender Datenbanken fahren; mit einer CI entfällt der Ersatz durch `make schema-check` dem Wortlaut nach, die Kette gibt es aber erst mit einer zweiten Schemaversion. Was sagt die Spezifikation dann? — **Ausgang:** beantwortet: `SPEC-004` ist ergänzt: der CI-Lauf von `make schema-check` ist der CI-Lauf der Regel, die Migrationskette bleibt geschuldet; der Satz zum Ersatz ohne CI bleibt stehen (Nr. 13).

### Fragen an den fachlichen Owner und den User

- **Lastenheft, `LH-QA-06`.** Der Satz „Solange es keine CI gibt, erkennt der fachliche Owner einen Lauf über `make` als Ersatz für den CI-Test an“ verliert mit der CI seine Wirkung. Zu entscheiden: bleibt er stehen, wird er gestrichen, oder gilt der Lauf über `make` weiter, wenn die CI ausfällt oder nicht läuft? Eine Textänderung nach `Accepted` braucht einen Change Request; vor `Accepted` wäre sie ohne ihn möglich. Der Planner ändert nichts. *Empfehlung des Architect:* stehen lassen; der Satz ist an „solange es keine CI gibt“ gebunden und erledigt sich mit der Lieferung, ohne Textänderung und ohne Change Request. Fällt die CI aus, wartet der Beleg auf einen Lauf. — **Ausgang:** offen, Adresse fachlicher Owner.
- **Wegwerfzweig und Entwurfs-Pull-Request.** Der Push der Zweige des Slice ist erlaubt (§4 Nr. 5). Darf dazu ein Wegwerfzweig für den roten Lauf gepusht und gelöscht werden, und darf ein Entwurfs-Pull-Request angelegt werden? Die DoD verlangt den roten Lauf auf dem Wegwerfzweig und den Lauf bei Push und Pull Request: ohne Wegwerfzweig sind DoD Punkt 2 und §5 nicht erfüllbar; ohne Pull Request bleibt der Auslöser `pull_request` ungesehen, als benanntes Negativ ([SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 9 d). *Empfehlung des Architect:* beides freigeben. — **Ausgang:** Wegwerfzweig beantwortet: ja (User, 2026-11-28, übermittelt über den Orchestrator: „3. ja“). Ein Pull Request ist nicht erlaubt; der Auslöser `pull_request` bleibt ungesehen (Nr. 9 d).
- **Repo und Runner.** Ist das Repo öffentlich, welche Größe hat damit der Runner, und was kosten die Laufminuten oder welches Kontingent gilt (nicht geprüft; Annahme in `SPEC-217` Nr. 5 b)? *Empfehlung des Architect:* Standard-Runner zuerst, ein größerer nur nach gemessenem Bedarf. — **Ausgang:** beantwortet: das Repo ist öffentlich (User, 2026-11-28, übermittelt über den Orchestrator: „2. public“). Der Workflow nimmt den Standard-Runner; Größe und Kontingent bleiben die Annahme aus `SPEC-217` Nr. 5 b, bis ein Lauf sie belegt.
- **Branch-Protection.** Soll ein roter Lauf das Mergen sperren? *Empfehlung des Architect:* vorerst nein (sie sperrte den Integrationsschritt ohne Pull Requests); stattdessen liest der Integrationsschritt nach jedem Push auf `main` den Lauf (Nr. 5 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)). — **Ausgang:** offen, Adresse User.
- **Images.** Sind die Pakete `d-check`, `a-check` und `d-migrate` auf `ghcr.io` ohne Anmeldung beziehbar (nicht geprüft)? *Empfehlung des Architect:* öffentlich stellen, falls nicht; keine Anmeldung im Workflow. — **Ausgang:** beantwortet: sie sind ohne Anmeldung beziehbar (User, 2026-11-28, übermittelt über den Orchestrator: „1. ja“).

### Offene Punkte aus der Umsetzung (Implementer, 2026-11-28)

- **Paketlaufzeit unter dem Detektor vor dem Push** (`SPEC-217` Nr. 5 c): `src/bootstrap` brauchte in zwei Vorproben 519,6 s und 513,4 s (§3), über der Schwelle von 450 s und unter der Grenze von 600 s je Paket. Ein Kern des Runners ist vermutlich langsamer als einer des Hosts (`SPEC-217` Nr. 5 a, nicht gemessen); ein Rot von `test-race` allein durch die Zeitgrenze ist damit möglich. Ohne Lockerung offen (`SPEC-217` Nr. 5 d): ein größerer Runner (Kosten), die Pakete nacheinander fahren (Änderung an `SPEC-122` Nr. 1) oder die Tests des Pakets beschleunigen; die beiden letzten sind je ein eigener Vorgang. Auf die Frage „Soll ich als Probe auf dem Standard-Runner pushen?“ hat der User am 2026-11-28 „ja“ gesagt (übermittelt über den Orchestrator). Seit den Vorproben trägt `src/bootstrap` zwei Tests mehr ([slice-betrieb-spans-nachweis](../done/slice-betrieb-spans-nachweis.md): mit Detektor 0,83 bis 1,04 s und 2,13 bis 2,45 s). — **Ausgang:** die Probe auf dem Standard-Runner war grün (§3, *Läufe auf dem Runner*), keine Änderung vor dem Push. Die Stufe `race` brauchte dort 954,0 s; der Abstand zu den zehn Minuten je Paket ist nicht bekannt, weil das Log keine Zeit je Paket nennt. Ein Rot an der Zeitgrenze bleibt ein Befund nach Nr. 5 d, keine Lockerung.
- **Zwei Grenzen von `workflow-check` fehlen in [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md) Nr. 11:** der Kommentar hinter der SHA wird nicht gegen den Tag gehalten (ein beliebiger Kommentar genügt), und eine Konfiguration ohne die Zeile `modules:` lässt das Ziel grün (§3, *Lokal gesehene Mutationen*). `harness/README.md` führt beides unter *Nicht behauptet*. Folge: ein falscher Kommentar führt bei der nächsten Anhebung in die Irre; eine geleerte Konfiguration prüft nichts. — **Ausgang:** entschieden (Architect, 2026-11-28): beide sind akzeptierte Negative ohne Fänger. Kommentar: `SPEC-217` Nr. 11 h (ob ein Tag zur SHA gehört, zeigt erst das Netz; die Anhebung löst die SHA neu aus dem Tag auf). Zeile `modules:`: Nr. 11 d, Teil der Verdrahtung. Keine Änderung an Konfiguration oder Rezept; `--enable workflows` im Rezept verschöbe die Lücke nur auf `workflows.dir`.
- **`make -n gates` bricht ab** (§3): [SPEC-211](../../../../harness/vertraege/SPEC-211-container-tests-in-der-kette.md) Nr. 5 a, Zeile 4 und Nr. 11 d des Vertrags stützen eine Prüfung auf diesen Aufruf, ohne `-k` nennt er die Rezepte hinter `schema-check` nicht. Der Slice ändert `harness/mk/schema.mk` nicht. — **Ausgang:** offen, Adresse Planner; der Verifier liest `make -n -k gates`.
- **Kontext von [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md)** (`Proposed`) beginnt mit „Eine CI gibt es noch nicht“; der Satz ist mit den Läufen auf dem Runner (§3) überholt. — **Ausgang:** erledigt (Architect, 2026-11-28): Kontext und die weiteren Stellen, die die CI als fehlend führen (Annahmen, Alternative C, *Negativ*, Re-Evaluierungs-Trigger), sind nachgezogen; die Entscheidung ist unverändert, die ADR bleibt `Proposed`.

### Entscheidungen zum Review (Architect, 2026-11-28)

Report: [Review](../../../reviews/2026-11-28-slice-ci-github-actions-review.md).

- **F-1, `workflow-check` grün bei lokaler Composite-Action, zweitem YAML-Dokument und `.YML`.** — **Ausgang:** die Zusage ist eingeschränkt, kein neuer Fänger: `SPEC-217` Nr. 8 nennt die Prüfmenge, Nr. 11 h die vier grünen Wege mit Grund und Folge; Nr. 2 c schließt eine lokale Aktion ausdrücklich aus (Zeilen 8, 9, 21 in §3). *Verworfen:* ein Textsuche-Schritt im Rezept (muss jede YAML-Schreibform treffen, `AGENTS.md` §3.17, eigene Gegenproben); ein Verbot von `.github/actions` (eine lokale Aktion kann in jedem Verzeichnis liegen). Das Modul `workflows` liest laut seinem Handbuch (`d-check --manual workflows` am Pin) nicht in lokale Aktionen hinein und nimmt allein `.yml` und `.yaml`.
- **F-2, `container:`, `services:` und `${{ }}` in `run` fehlen in den Negativen.** — **Ausgang:** `SPEC-217` Nr. 2 f verbietet die drei, Nr. 11 c führt sie ohne Fänger (Zeile 20 in §3). Die gelieferte Datei erfüllt das Verbot (Review, *Negativbefunde*); der Workflow ändert sich nicht.
- **Auftrag für die Fix-Runde** (Implementer): allein Kommentare und Hilfetext auf die eingeschränkte Zusage bringen (`AGENTS.md` §3.11); `SPEC-217`, `harness/README.md`, der Index der Verträge, [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md) und §3 sind mit der Entscheidung nachgezogen. (1) Fragment `workflow.mk` unter `harness/mk/`: Hilfetext des Ziels sagt „fremde `uses:` in den Workflow-Dateien“ statt „Aktionen des Workflows“, die Zeile *Grenze* nennt lokale Aktionen, zweites Dokument und `.YML` mit Zeiger auf `SPEC-217` Nr. 11 h. (2) Kopf von `.github/workflows/gates.yml`: „Form der Pins“ gilt für die Workflow-Dateien selbst (Nr. 8, 11 h); die Aufzählung ohne Fänger nennt `container:`, `services:` und Ausdrücke in `run`. Keine Änderung an Rezept, Konfiguration oder Schritten, kein neuer Test, keine Mutation. Danach `make gates`. — **Ausgang:** offen, Adresse Implementer (Fix-Runde).

### Randformen (Raster, `AGENTS.md` §3.16)

Jede Nummer meint [SPEC-217](../../../../harness/vertraege/SPEC-217-ci.md).

| Operation | fehlend oder leer | unbekannt oder fremd | Abbruch | Wiederholung |
|---|---|---|---|---|
| Auslöser | reine Doku-Änderung läuft voll, keine Pfadfilter; Tag löst nicht aus: Nr. 2 a; gelöschter Zweig: gilt nicht, es gibt keinen Stand zu prüfen | Pull Request aus einem Fork: nur lesendes Token, keine Geheimnisse, Nr. 2 a, b; weitere Auslöser: Nr. 2 a | gilt nicht: vor dem Start gibt es keinen Lauf | Push und Pull Request desselben Standes: zwei Läufe, Nr. 2 a; schnelle Pushes: Nr. 2 g |
| Checkout | flacher Klon, Historie ohne `basis`: Nr. 3 a, d | Merge-Stand eines Pull Request: Nr. 3 c | Fehler des Checkout ist rot: Nr. 7 | gilt nicht: ein Checkout je Lauf |
| Aktion | Workflow oder `uses:` fehlt: Nr. 8 (leere Prüfmenge) | Tag, Zweig, gekürzte SHA, `docker://`, SHA ohne Kommentar, kaputtes YAML: Nr. 8; fremde, korrekt gepinnte Aktion, `container:`, `services:`: Nr. 11 c; lokale Aktion, zweites Dokument, `.YML`, Kommentar ohne Tag: Nr. 11 h | gilt nicht: Fehler beim Bezug der Aktion ist Rot vor dem ersten Schritt, wie Nr. 7 | Anhebung eines Pins: Nr. 2 c |
| Bezug von Images und Modulen | Image oder Modul nicht beziehbar, auch ohne Anmeldung: Nr. 4 c | anderer Inhalt hinter einem Pin: Images stehen auf Digest, Module prüft die Stufe `deps` gegen `go.sum` (Nr. 4 c) | Netzfehler mitten im Bezug: Nr. 4 c, 6 | Bezugsgrenze: Nr. 4 c; kein Cache: Nr. 4 d |
| Lauf `make gates` | ein Gate fehlt: Nr. 2 f, 11 c | Docker oder Runner mit anderem Verhalten: Nr. 4 a, b; zu klein oder zu langsam: Nr. 5; Platz: Nr. 2 f | Zeitgrenze oder Abbruch: Nr. 2 e, 7 | neuer Start nach Rot: Nr. 6 b, c |
| Ergebnis | kein Lauf zu einem Commit: Nr. 6 a, 11 f; Rot ungelesen: Nr. 11 e | gilt nicht: das Ergebnis ist der Status des Jobs, ein anderes liest niemand | abgebrochener oder übersprungener Lauf ist nicht grün: Nr. 6 a | ein grüner zweiter Lauf überdeckt keinen roten: Nr. 6 c |

### Risiken

- Ein Gate ist lokal grün und auf dem Runner rot (weniger Kerne, Fristen von 120 s je Stufe, Paketlaufzeit unter `-race`, Export-Namen im Neustarttest, Tests, die unter Last flackern). Am wahrscheinlichsten trifft es `test-race`. — **Ausgang:** im ersten Lauf nicht eingetreten (§3); ein Lauf sagt nichts über Flackern und Abstand zur Zeitgrenze. Offen, Adresse Planner (Integrationsschritt); tritt es ein, Carveout oder Folge-Slice.
- Der Speicherplatz des Runners reicht für Images und Zwischenstände nicht (nicht gemessen). — **Ausgang:** entfallen: im grünen Lauf waren 85 G frei vor und 79 G nach `make gates` (§3, *Aus den Job-Logs*).
- Ein roter Lauf auf `main` bleibt unbemerkt, weil kein Wächter ihn liest und nichts am Merge hängt; die Pflicht des Integrationsschritts (Nr. 5 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)) trägt kein Gate (`SPEC-217` Nr. 11 e). — **Ausgang:** offen, Adresse User (Branch-Protection) und Planner.
- Die CI hängt an einem Dienst außerhalb des Repos; fällt er aus oder ändert er das Runner-Image, ist das Ergebnis ohne Änderung am Stand anders. — **Ausgang:** offen, akzeptiert nach `SPEC-217` Nr. 11 b und f; fachlicher Owner (`LH-QA-06`).
- Die CI liegt vor der Erstabnahme (Entscheidung des Users vom 2026-11-28): die Abnahme wartet damit auf diesen Slice; sein grüner Lauf auf dem Runner ist gesehen (§3). — **Ausgang:** offen bis zur Closure, Adresse Planner und User.
- Der Slice hält die Gate-Kette (§4), solange Antworten des Users ausstehen. — **Ausgang:** entfallen: Images, Repo und Wegwerfzweig sind am 2026-11-28 beantwortet.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

Noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`).

**Vorgelagert — offene Beobachtungen sichten** (Planner beim Start, 2026-11-28, `main` bei `dfb51e47`; Zahl der Dateien unter `evidence/`):

- `BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (3): Frage 7.
- `BEO-REPO/test-fremden-pakets-flackert-unter-last` (4, geplant in [slice-test-flackern](../in-progress/slice-test-flackern.md)): Frage 8.
- `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (21): Frage 9.
- `BEO-REPO/gegenproben-erzeugen-die-last-selbst` (2): Frage 6.
- `BEO-REPO/image-tag-race-von-allen-worktrees-geteilt` (2): die Vorprobe (`SPEC-217` Nr. 5 c) läuft neben zwei Zweigen; sie fährt den eigenen Baum, nicht den zuletzt gebauten Stand des Tags.
- `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (93, verkörpert in `AGENTS.md` §3.15): die Negative stehen in §3 mit Maß.

Mit der Planung fällt kein Beleg an, kein Eintrag erreicht durch sie 3×; die Einträge ab 3 ohne Zielort liest die Welle-Closure.

**Modus-Begründungsblock:** entfällt, alle berührten Sub-Areas GF.

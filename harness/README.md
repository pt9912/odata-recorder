# Harness

---

## Purpose

Dieser Harness verbindet bestehende Spezifikationen, ADRs,
Planning-Dokumente und Gates. Er ist **kein Ersatz** für `spec/` oder
`docs/`, sondern ein **Einstiegspunkt** für Menschen und AI-Code-Agenten.

Wenn diese Datei einer kanonischen Quelle widerspricht, **gewinnt die
kanonische Quelle**, und diese Datei wird angepasst.

Strukturregeln (Verzeichniskonvention, ID-Schemata, Modus-Deklarationen
pro Sub-Area, Zusatzklassen für Sensors-Bindung) sowie Adaptionen ggü.
der adoptierten Baseline leben in [`conventions.md`](conventions.md).
Diese Datei dupliziert sie nicht.

## Source precedence

| Rang | Datei | Charakter |
|---|---|---|
| 1 | [`spec/lastenheft.md`](../spec/lastenheft.md) | vertraglich abnahmebindend |
| 2 | [`spec/spezifikation.md`](../spec/spezifikation.md) (Produkt) und [`harness/vertraege/`](vertraege/README.md) (Werkzeuge) | technisch fortschreibbar |
| 3 | [`spec/architecture.md`](../spec/architecture.md) | Komponenten/Sequenzen, meilensteinfrei |
| 4 | [`docs/plan/adr/`](../docs/plan/adr/) | Architekturentscheidungen |
| 5 | [`docs/plan/planning/in-progress/roadmap.md`](../docs/plan/planning/in-progress/roadmap.md) | Wellen-Sequenz |
| 6 | `docs/user/*` *(falls vorhanden)* | Operations, Quality, Releasing | <!-- d-check:ignore (Verzeichnis optional; entlinkt, da im frischen Repo selten vorhanden) -->
| 7 | [`README.md`](../README.md) | Projekt-Überblick |
| 8 | [`AGENTS.md`](../AGENTS.md) | Agent-Briefing |
| 9 | diese Datei | Harness-Einstieg |

> Die Ränge 1–3 sind die **drei Spec-Straten** — Vertrag, Technik, Sicht —,
> und alle drei sind obligatorisch (Baseline-Regelwerk
> `grundlagen-referenz-richtung.md` §Spec-Straten). **Adaption ist die
> Zwei-Straten-Form, nicht die Drei-Straten-Form**: Wer Rang 2 streicht,
> deklariert das als `MR-<NNN>` in [`conventions.md`](conventions.md) und
> nummeriert neu (dann acht Ränge).

## Guides (Feedforward-Quellen)



| Quelle | Inhalt |
|---|---|
| [`spec/lastenheft.md`](../spec/lastenheft.md) | Anforderungen, IDs, Akzeptanzkriterien |
| [`spec/spezifikation.md`](../spec/spezifikation.md) | technische Details, Defaults |
| [`harness/vertraege/`](vertraege/README.md) | Verträge der Harness-Werkzeuge |
| [`spec/architecture.md`](../spec/architecture.md) | Komponenten, Schichten, Constraints |
| [`docs/plan/adr/`](../docs/plan/adr/) | Architekturentscheidungen |
| [`docs/plan/planning/`](../docs/plan/planning/) | Slice-Pläne und Roadmap |
| [`AGENTS.md`](../AGENTS.md) | Hard Rules, Source Precedence, Workflow |
| [`conventions.md`](conventions.md) | repo-lokale Strukturregeln, Adaptions-Block (`MR-*`), Modus-Deklarationen |
| `.harness/skills/reviewer.md` | Reviewer-Skill: HIGH-Liste, Kategorien-Regeln, Negativbefund-Pflicht, Output-Schema (Modul 10) — nächste Rolle nach Schritt 8 des Minimal Agent Workflow, nicht Teil der Implementer-Eingabe |
| `.harness/baseline/<tag>/regelwerk/` (vendored; `README.md` = Index) | adoptiertes Betriebsregelwerk in Agenten-Kurzform — **präsente nachschlagbare Vertiefung**, pro Entscheidung abschnittsweise (siehe [`AGENTS.md`](../AGENTS.md) §1); derivativ, Stand/Tag siehe [`conventions.md`](conventions.md) §Baseline |
| `.harness/baseline/<tag>/templates/` (vendored, parallel) | Referenz-Form der Skelette, auf die das Regelwerk mit `../templates/…` als „Ziel-Form" verweist (netzlos, weil parallel zu `regelwerk/`); Vorlagen zum Kopieren-und-Ausfüllen |

## Sensors (Feedback-Gates)



| Target | Vertrag | Bindung |
|---|---|---|
| `make docs-check` | Doku-Referenzen und Spec-Straten-Richtung (d-check) | — |
| `make schema-check` | Persistenzschema: d-migrate validate und generate, DDL in SQLite, Negativtests der Constraints, `foreign_key_check`/`integrity_check` (netzlos, ohne Mount) | — |
| `make migration-check` | Eingebettete Migration des SQLite-Adapters byte-gleich dem DDL aus `schema.yaml`, keine ungeprüfte weitere Migration (netzlos, ohne Mount) | [ADR-0006](../docs/plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) |
| `make build` | Binary aus `src/` über das `Dockerfile` (Stufe `build`, reproduzierbar, netzlos außer `deps`); `pruefung build` prüft Version und `-trimpath` | [ADR-0001](../docs/plan/adr/0001-implementierungssprache-go.md) |
| `make test` | `pruefung test`: `gofmt` (nennt nicht formatierte Dateien), `go vet` und Unit-Tests über das `Dockerfile` (Stufe `test`, netzlos außer `deps`) | [ADR-0001](../docs/plan/adr/0001-implementierungssprache-go.md) |
| `make test-race` | `pruefung race`: `go test -race` mit cgo über alle Pakete der Stufe `source` (Dockerfile, Stufe `race`, Go-Image mit C-Compiler, netzlos außer `deps`); meldet Datenrennen, rote Tests, Zeitgrenze | [ADR-0007](../docs/plan/adr/0007-race-detektor-im-gate.md) |
| `make groessenmarke` | Größenmarke der Pläne mit Block unter `docs/plan/planning/in-progress/` (`SPEC-128`): Marke gleich Formel, Messung auf oder unter der Marke oder Eintrag mit Grund; grün ohne Slice in Arbeit | [ADR-0009](../docs/plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) |
| `make a-check` | Architektur-Regeln des Hexagons gemäß `.a-check.yml` (a-check, netzlos, schreibgeschützt): Import-Richtungen der Schichten; prüft heute einen leeren Baum, bis Code existiert | [ADR-0002](../docs/plan/adr/0002-architektur-gate-a-check.md) |
| `make lint-profil` | Lint-Profil `.golangci.yml` gegen das Schema von golangci-lint (netzlos, ohne Mount); prüft kein Code | [ADR-0003](../docs/plan/adr/0003-lint-profil-golangci.md) |
| `make lint` | Go-Code nach `.golangci.yml` und eigene Prüfungen (keine `//nolint`-Direktive, `# Why:` je Ausnahme, `warn-unused`); Dockerfile, Stufe `lint`, netzlos außer `deps` | [ADR-0003](../docs/plan/adr/0003-lint-profil-golangci.md) |
| `make gates` | alle inneren Gates | — |

**Werkzeuge — genannt, weil der Lauf sie braucht, aber kein Gate:**

| Target | Tut was | Bindung |
|---|---|---|
| `make schema-validate` | prüft `tools/schema/schema.yaml` mit d-migrate (netzlos, ohne Mount) | kein Gate |
| `make schema-generate` | erzeugt SQLite-DDL und Report aus dem Schema als Tar-Strom auf stdout (netzlos, ohne Mount) | kein Gate |
| `make a-check-graph` | gibt den Architektur-Graphen (Mermaid) aus `.a-check.yml` aus, prüft nichts (netzlos, schreibgeschützt) | kein Gate |
| `make gegenproben` | Gegenproben zu `a-check`, `lint.sh`, `migration-check.sh`, `pruefung` und `groessenmarke`: Fälle in Kopien oder Wegwerfverzeichnissen, bei geänderter Eingabe (Stempel `harness/gegenproben.stempel`) | [ADR-0005](../docs/plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) |

**Aktueller Lauf-Status:** lokal `make help` / `make gates`; eine CI gibt es noch nicht.
**Rote Gates:** keine; ein rotes Gate bekommt einen `CO-<NNN>` und verweist in der Bindung-Spalte darauf (Modul 7).
**Nicht behauptet:** ein Coverage-Gate, die Gegenproben zu `make docs-check`, `make schema-check`, `make lint-profil` und `make baseline-verify`, dass die Stufen `test`, `build` und `race` des `Dockerfile` `pruefung` rufen und `test-race` an `GATE_CHECKS` hängt (kein Fall der Gegenproben fängt es), ein Datenrennen, das kein Lauf auslöst, eines ohne ungeordneten Zugriff und eines in den Modulen unter `tools/` (`make test-race` meldet sie nicht, `SPEC-122` Nr. 8 und 9) ob die Marke die richtige ist, die Klasse der Zeilen, Code vor der Marke und die Verdrahtung von `make groessenmarke` (`SPEC-128` Nr. 10; die Fälle 38 bis 120 der Gegenproben fangen den Aufruf von `docker` nicht) sowie der Lasttest zu `LH-QA-01` — der Code ist ein Skelett aus Composition Root und Version; `make a-check` prüft damit erst eine Schicht. Ebenso fehlt die Migrationskette des Schemas; sie braucht eine zweite Schemaversion.

**Targets der Werkzeug-Fragmente:** [`mk/ai-harness-init.md`](mk/ai-harness-init.md) — der Teil dieses Index, den `ai-harness-init` bei jedem Lauf neu schreibt; was das Repo über eines dieser Targets entscheidet, steht in dieser Zeile.

## Traceability rules

- PRs/Commits **müssen** mindestens eine `<LH-*>` oder `ADR-*`-ID nennen.
- Neue oder geänderte Anforderungen brauchen einen Beleg: Test, Gate, Demo oder ADR.
- Neue ADRs müssen im ADR-Index ergänzt werden.
- Änderungen an Planning-Dokumenten müssen die Lifecycle-Regeln beachten (open → next → in-progress → done; reine `git mv`-Commits siehe AGENTS.md §3.3).

## Safety and scope boundaries



- Kein Anwendungscode, kein Geheimnis und kein echter Upstream-Zugang im Repo: Prüfläufe sind netzlos und laufen ohne schreibbaren Mount (`AGENTS.md` §3.1).
- Änderungen am Lastenheft folgen seiner Historie-Regel (`spec/lastenheft.md` §Historie); Tatsachenberichtigungen werden als solche gekennzeichnet.
- Die vendorte Baseline unter `.harness/baseline/` wird nicht bearbeitet; ein Stand-Wechsel ist ein eigener Commit samt `make baseline-verify`.
- Ein Gate wird nicht gelockert (`AGENTS.md` §3.6); neue Gates sind erlaubt.

## Minimal agent workflow

1. Diese Datei lesen.
2. Relevante kanonische Quelle lesen.
3. Betroffene IDs identifizieren.
4. Kleinste Änderung planen.
5. Engsten nützlichen Sensor laufen lassen.
6. Repo-weiten Gate-Lauf vor Handoff (`make gates`).
7. Doku/Indizes aktualisieren, falls ein öffentlicher Vertrag berührt.
8. Ausgeführte Sensors und verbleibende Risiken berichten.

Dieser Workflow deckt ausschließlich die Implementer-Rolle ab. Schritt 8
ist der Rollenwechsel, kein Abschluss: Bericht → Handoff an Reviewer
(`.harness/skills/reviewer.md`, siehe §Guides) → Verifier. Kein
Self-Review — anderer Kontext findet andere Findings, derselbe Kontext
dieselben blinden Flecken (Baseline-Regelwerk `modul-08-agentenrollen.md`).

## Leseordnung

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-harness-dateien.md`
§harness/README.md als Einstiegspunkt — die Menschen-Hälfte des Einstiegs:
drei bis fünf **geordnete** Zeiger, was ein neuer Mensch zuerst liest und was
bei Bedarf; eine Leseordnung, die alles nennt, ist keine.

1. [`AGENTS.md`](../AGENTS.md) §Harte Regeln
2. [`spec/lastenheft.md`](../spec/lastenheft.md) — was zugesagt ist
3. [`spec/spezifikation.md`](../spec/spezifikation.md) — wie genau, mit Defaults und Fehlercodes
4. [`spec/architecture.md`](../spec/architecture.md) — Komponenten und Sequenzen
5. bei Bedarf [`conventions.md`](conventions.md)

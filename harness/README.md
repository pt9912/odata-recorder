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
| `make docs-check` | Doku-Referenzen, Spec-Straten-Richtung und Größengrenzen der Guide-Dateien (d-check: `file`, `structure`; `SPEC-040`) | [ADR-0010](../docs/plan/adr/0010-groesse-der-guide-dateien-als-gate.md) |
| `make schema-check` | Persistenzschema: d-migrate validate und generate, DDL in SQLite, Negativtests der Constraints, `foreign_key_check`/`integrity_check` (netzlos, ohne Mount) | — |
| `make migration-check` | Eingebettete Migration des SQLite-Adapters byte-gleich dem DDL aus `schema.yaml`, keine ungeprüfte weitere Migration (netzlos, ohne Mount) | [ADR-0006](../docs/plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) |
| `make build` | Binary aus `src/` über das `Dockerfile` (Stufe `build`, reproduzierbar, netzlos außer `deps`); `pruefung build` prüft Version und `-trimpath` | [ADR-0001](../docs/plan/adr/0001-implementierungssprache-go.md) |
| `make test` | `pruefung test`: `gofmt` (nennt nicht formatierte Dateien), `go vet` und Unit-Tests über das `Dockerfile` (Stufe `test`, netzlos außer `deps`) | [ADR-0001](../docs/plan/adr/0001-implementierungssprache-go.md) |
| `make test-race` | `pruefung race`: `go test -race` mit cgo über alle Pakete der Stufe `source` (Dockerfile, Stufe `race`, Go-Image mit C-Compiler, netzlos außer `deps`); meldet Datenrennen, rote Tests, Zeitgrenze | [ADR-0007](../docs/plan/adr/0007-race-detektor-im-gate.md) |
| `make test-integration` | Binary im Container ohne Netz: Record, Upstream stoppen, Replay, Mock, Hybrid und Admin-Listener in elf Stufen (`SPEC-037`) | Reproduzierbarkeit (`LH-QA-06`), Zeitbudget `SPEC-211` Nr. 4 |
| `make test-determinism` | 100 Läufe des Replay nach Reset: gleiche normalisierte Antworten und Cursor-Stände (`SPEC-038`) | Reproduzierbarkeit (`LH-QA-06`), Zeitbudget `SPEC-211` Nr. 4 |
| `make test-contract` | 36 Fixtures `ODV4-*` über HTTP und je Fehlerklasse von `LH-QA-05` ein Negativfall (`SPEC-042`) | Reproduzierbarkeit (`LH-QA-06`), Zeitbudget `SPEC-211` Nr. 4 |
| `make test-security` | Canary-Werte (94 Stellen) in keinem Artefakt, 17 Admin-Endpunkte ohne Token abgewiesen, Listener nur Loopback oder TLS, ohne Netz (beobachtet im Container von `e2e.test`, `SPEC-041` Nr. 1) | Datensicherheit (`LH-QA-04`), Zeitbudget `SPEC-211` Nr. 4 |
| `make test-image` | Laufzeit-Image ohne Netz und Volume: Image geprüft; Recordings, Workspace und Snapshot nach Stopp und Start byte-gleich lesbar; zwei Bauten, gleiches Binary aus zwei Übersetzungen (`SPEC-323`) | Reproduzierbarkeit (`LH-QA-06`, `LH-RB-03`), Zeitbudget `SPEC-211` Nr. 4 |
| `make groessenmarke` | Größenmarke der Pläne mit Block unter `docs/plan/planning/in-progress/` (`SPEC-128`): Marke gleich Formel, Messung auf oder unter der Marke oder Eintrag mit Grund; grün ohne Slice in Arbeit | [ADR-0009](../docs/plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) |
| `make handbuch-check` | Benutzerhandbuch (`SPEC-200`): Beispiele gegen das gebaute Binary, Kopf gleich dem Binary, `make`-Ziele, keine Planung, Chronik oder Kennung im Text (Stufe `handbuch`, netzlos außer `deps`) | [ADR-0004](../docs/plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) |
| `make zeichen-check` | Unsichtbare Zeichen (`SPEC-233`): keine Textdatei außer `.harness/baseline/` trägt BOM, NBSP, Zeilentrenner, Nullbreiten- oder Steuerzeichen (253 Zeichen) oder ungültiges UTF-8 (netzlos, ohne Mount) | [ADR-0004](../docs/plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) |
| `make katalog-check` | Kataloge der Fehlercodes, Schlüssel und Kommandos (`SPEC-201`): der Quelltext steht in Spezifikation und Handbuch, das Handbuch nennt nichts Fremdes (Stufe `katalog`, netzlos außer `deps`) | [ADR-0004](../docs/plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) |
| `make a-check` | Architektur-Regeln des Hexagons gemäß `.a-check.yml` (a-check, netzlos, schreibgeschützt): Import-Richtungen der Schichten über den Code unter `src/` und `tests/` | [ADR-0002](../docs/plan/adr/0002-architektur-gate-a-check.md), [ADR-0012](../docs/plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) |
| `make lint-profil` | Lint-Profil `.golangci.yml` gegen das Schema von golangci-lint (netzlos, ohne Mount); prüft kein Code | [ADR-0003](../docs/plan/adr/0003-lint-profil-golangci.md) |
| `make lint` | Go-Code nach `.golangci.yml` und eigene Prüfungen (keine `//nolint`-Direktive, `# Why:` je Ausnahme, `warn-unused`); Dockerfile, Stufe `lint`, netzlos außer `deps` | [ADR-0003](../docs/plan/adr/0003-lint-profil-golangci.md) |
| `make workflow-check` | CI-Workflow: Aktionen auf voller SHA mit Tag-Kommentar, gültiges YAML, nicht leere Prüfmenge (d-check, Modul `workflows`, netzlos, schreibgeschützt; `SPEC-217` Nr. 8) | Reproduzierbarkeit (`LH-QA-06`) |
| `make gates` | alle inneren Gates | — |

**Werkzeuge — genannt, weil der Lauf sie braucht, aber kein Gate:**

| Target | Tut was | Bindung |
|---|---|---|
| `make schema-validate` | prüft `tools/schema/schema.yaml` mit d-migrate (netzlos, ohne Mount) | kein Gate |
| `make schema-generate` | erzeugt SQLite-DDL und Report aus dem Schema als Tar-Strom auf stdout (netzlos, ohne Mount) | kein Gate |
| `make a-check-graph` | gibt den Architektur-Graphen (Mermaid) aus `.a-check.yml` aus, prüft nichts (netzlos, schreibgeschützt) | kein Gate |
| `make gegenproben` | Gegenproben zu `a-check`, `lint.sh`, `migration-check.sh`, `pruefung`, `groessenmarke`, `zeichen`, `katalog`, `docs-check`, `lasttest.sh`, `test-image.sh`, nur bei neuer Eingabe (Stempel) | [ADR-0005](../docs/plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) |
| `make image` | Laufzeit-Image `odata-recorder:image` (Stufe `image`, `SPEC-322`), Version aus `VERSION` | kein Gate |
| `make lasttest` | Lasttest nach `LH-QA-01.a`: p95 des Gesamtspans für Replay und Mock im Container mit 2 CPU, 4 GiB ohne Netz; Profil `abnahme` rund 50 Minuten, `LASTTEST_PROFIL=rauch` als kurzer Handlauf (`SPEC-216`) | kein Gate |
| `sh tools/harness/mutationslauf.sh <Liste>` | Mutationsliste in einem Container-Lauf ohne Netz und Mount, mit `--memory`, warmer Cache des Images (`SPEC-210`); kein Laufzeitgewinn zugesagt | kein Gate, kein `make`-Ziel, keine Anweisung einer Rolle ruft es ([ADR-0005](../docs/plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) |

**Aktueller Lauf-Status:** lokal `make help` / `make gates`; die CI (Workflow `gates` unter `.github/workflows/`, `SPEC-217`) fährt `make gates` bei Push und Pull Request, auf einem Runner ist noch kein Lauf gesehen.
**Rote Gates:** keine; ein rotes Gate bekommt einen `CO-<NNN>` und verweist in der Bindung-Spalte darauf (Modul 7).
**Nicht behauptet:** ein Coverage-Gate, die Gegenproben zu `make docs-check` außer den Größengrenzen (`SPEC-040` Nr. 11) und der Ausnahme `.wt/**` (`SPEC-212`), von der nicht behauptet ist: ein getrackter Pfad unter `.wt/`, eine Kopie ohne `.git`, Symlinks, die Module `targets` und `planning` von d-check (sie lesen feste Pfade, kein Fall) und der Kontext von `docker build` (`SPEC-212` Nr. 6 a, b, d, e), `harness/README.md` und die vendorte Baseline außerhalb der Größengrenzen (`SPEC-040` Nr. 10 d), die Verdrahtung der Gegenprobe zu `make docs-check` (Kopie des Binary in die Stufe `gegenproben`, `.d-check.yml` in den Eingaben, `SPEC-040` Nr. 10 h), die Gegenproben zu `make schema-check`, `make lint-profil` und `make baseline-verify`, dass die Stufen `test`, `build` und `race` des `Dockerfile` `pruefung` rufen und `test-race` an `GATE_CHECKS` hängt (kein Fall der Gegenproben fängt es), ein Datenrennen, das kein Lauf auslöst, eines ohne ungeordneten Zugriff und eines in den Modulen unter `tools/` (`make test-race` meldet sie nicht, `SPEC-122` Nr. 8 und 9), ob die Marke die richtige ist, die Klasse der Zeilen, Code vor der Marke und die Verdrahtung von `make groessenmarke` (`SPEC-128` Nr. 10; die Fälle 38 bis 120 der Gegenproben fangen den Aufruf von `docker` nicht), die Verdrahtung von `make handbuch-check` (Stufe `handbuch`, Eintrag in `GATE_CHECKS`, `.dockerignore`, `SPEC-200` Nr. 14 h), der Aufruf von `docker` in den Beispielen, Ausblick und Chronik ohne Stichwort (`SPEC-200` Nr. 14 a, e), die Verdrahtung von `make zeichen-check` (Stufe `zeichen`, Kopie in die Stufe `lint`, Eintrag in `GATE_CHECKS`, Aufruf von `docker` und `tar`, `SPEC-233` Nr. 16 h), die Umsetzung einer Folge aus Backslash, `u` und vier Hex-Ziffern in ein sichtbares Zeichen durch das Schreibwerkzeug (`AGENTS.md` §3.18, `SPEC-233` Nr. 16 a) und Zeichen außerhalb der Menge (Variationsselektoren, Füllzeichen, Homoglyphen, `SPEC-233` Nr. 16 b) sowie die Zahlen und die Verdrahtung von `make lasttest` zu `LH-QA-01` (Stufe `lasttest`, Fragment, Container und Binary zusammen belegt allein der Handlauf, `SPEC-216` Nr. 16 a und k). Von `make test-integration`, `make test-determinism` und `make test-contract` ist nicht behauptet: die Zeile `GATE_CHECKS += <Ziel>` je Fragment und dass die Ziele in der Tabelle der Gates stehen (`SPEC-211` Nr. 5 a, b), das Zeitbudget (Nr. 5 c), Speicherlimit und Frist (Nr. 3), der Abbruch durch SIGKILL (Nr. 2 d), der Abbruch an die Prozessgruppe von `make` (nur als Handlauf ohne Mutation gesehen), TERM allein an `make` (erreicht den Treiber nicht), dass der Container nach INT oder TERM am Treiber bei `test-integration` und `test-contract` endet (nur `test-determinism` unterscheidet es), die Obergrenze der Wartezeit nach einem roten Check (Nr. 5 e), die Randformen des Baus (die Meldung fängt, nicht der Exit, Nr. 5 f), dass zwei Bauten desselben Tags denselben Inhalt liefern und dass höchstens vier Container zugleich laufen, der von `make test-security` mitgezählt (Nr. 5 g, h) sowie Text, Argument-Check und ein hängender Daemon (Nr. 5 i). Von `make test-integration` ist nicht behauptet: die Verdrahtung der Stufe `integration` und der Flags von `docker run` außer dem Netz-Fall (`SPEC-037` Nr. 10 f), Import und Workspace über die Prozessgrenze in den Stufen mit dem Übergang (`SPEC-037` Nr. 10 a; den Admin-Listener fährt allein `TestE2EAdmin`), über den Admin-Listener `LH-FA-09` Happy Path und Negative, `LH-FA-10` Boundary und Negative und `LH-FA-18` Negative (`SPEC-300` *Ungeprüft*), über den Admin-Listener der Entity-State eines Workspace ohne Änderungen nach `reset` (`LH-FA-09` Boundary, `SPEC-300` Nr. 14), die Markierung des Austauschs selbst bei vollem Speicher (`SPEC-037` Nr. 10 o), ein `SIGTERM` bei wartender Upstream-Anfrage der Stufe `TestE2EHybrid` (`SPEC-037` Nr. 10 n) und ein Fehler des Baus: Stufe 10 fängt nur Fehler der Konfigurationsprüfung (`ODR-E2001`, Exit 1), ein `Start`, der einen Fehler des Baus mit Exit 0 meldet, macht allein Stufe 4 (`TestE2EPersistenz`) rot, und für einen Fehler des Baus nur im Modus `hybrid` hat keine Stufe einen Fall. Von `make test-contract` ist nicht behauptet: die Verdrahtung (Stufe `contract`, Flags von `docker run`, Wertung des Treibers, `SPEC-042` Nr. 9 f), das gebaute Binary (Nr. 9 a) und `$metadata` mit ungültigem Dokument. Von `make test-security` ist nicht behauptet: die Verdrahtung (`--network none`, `--read-only`, `--tmpfs /tmp` und der Mount je Test-Binary: nur der Container von `e2e.test` beobachtet sie, `--tmpfs /tmp` zwei von drei; `--memory`, die Frist, der Lauf des dritten Binary, Wertung des Treibers, Eintrag in `GATE_CHECKS`, `SPEC-041` Nr. 10 h), der Root (die 17 Routen am Listener des `Handle`, die Abweisung am Prozess in jedem Modus, das Token in `audit.jsonl`, `SPEC-041` Nr. 10 a), der Start des Binary mit nicht-loopback `ADMIN_LISTEN` ohne Paar und ohne Flag (Nr. 10 b), ein Zugriff von einem anderen Host (Nr. 10 c), Orte außerhalb der Arbeitswurzel, Speicher, Swap und Core-Dumps (Nr. 10 d), ein Geheimnis im Pfad, ein Teilstück ohne das Präfix `CANARY-` und jede Kodierung, die das Präfix ändert (Nr. 10 e), und die Modi `replay`, `mock` und `hybrid` im Canary-Test (Nr. 10 f). Von `make katalog-check` ist nicht behauptet: die Verdrahtung (Stufen `katalog-bin` und `katalog`, Eintrag in `GATE_CHECKS`, `.dockerignore`, `SPEC-201` Nr. 9 f), die Bedeutung einer Zeile, die Erreichbarkeit eines Codes in den Ausnahmen (Nr. 9 c), Literale `ODR-…` außerhalb von `fehler.go` (Nr. 9 b), ein Schlüssel oder ein Kommando, den der Quelltext nicht als Literal oder Konstante nennt, und ein Name in Prosa (Nr. 9 d, 9 e). Vom Mutationslauf (`SPEC-210`) ist nicht behauptet: dass er `.wt/` auslässt (`SPEC-212` Nr. 6 c), ein Laufzeitgewinn gegenüber den Skripten der Rollen (die Messungen widersprechen sich), ein Gate für das Modul `tools/mutationslauf` (`make gates` fährt weder `gofmt` noch `go vet` noch Lint noch seine Tests), die Klasse `nebenläufig` oder `andere` der Liste (die Rolle setzt sie, das Werkzeug glaubt ihr, `SPEC-210` Nr. 12 b), `go vet` gegen die Mutationen (Nr. 12 a), ein Aufruf durch die Rollen, ein Mutations-Protokoll zwischen den Rollen, ein Cache über Läufe hinweg (der Cache lebt im Container, Nr. 8) und einzelne Fänger (etwa Anker `^` des Musters, Abbruch im Arm ohne Detektor, Signal vor `docker start`, Gesamtgrenze in der Nullprobe). Von `make test-image` ist nicht behauptet: die Zeile `GATE_CHECKS += test-image` und der Aufruf des Treibers (`SPEC-323` Nr. 9 d), ein Neustart mit eingehängtem Volume oder neuem Container (Nr. 9 c), ein erzwungenes Ende des Containers, der Modus `mock` im Image (Nr. 9 a, b), Frist, Speicherlimit und ein hängender Daemon (Nr. 9 e), die Gleichheit des Binary über Hosts, Bau-Dienste, Kernzahlen oder Plattformen, bei einem Bau aus leerem Cache und bei einer Abweichung, die zwischen den zwei Übersetzungen eines Laufs nicht erscheint, die Gleichheit der Image-Kennung und der übrigen Dateien des Image sowie die neue Übersetzung unter einer Laufzeit, die beim Kopieren den Zeitstempel neu setzt (Nr. 9 g), ein Eintrag des Image unter einem der 11 Namen, die die Laufzeit selbst anlegt, Eigentümer und Rechte der Einträge außer `/data` und `/tmp` und die Bytes der Vertrauensanker (Nr. 9 h) und die Bytes der Bodies nach dem Neustart über die zum Bestand passende Referenz hinaus (Nr. 9 i). Von der CI und `make workflow-check` ist nicht behauptet: Auslöser, Rechte des Tokens, Zugriff auf Geheimnisse, welche Aktion außer `actions/checkout` läuft, `persist-credentials`, Runner-Label, Zeitgrenze, der Aufruf `make gates`, das Verbot von `continue-on-error`, `if:`, Cache und `concurrency` und die Eindeutigkeit von Datei und Job (allein das Review, `SPEC-217` Nr. 11 c), die Zeile `GATE_CHECKS += workflow-check`, der Pfad und der Inhalt der Konfiguration (ohne die Zeile `modules:` bleibt das Ziel grün) und die Formen der Pins nach einem Wechsel des Pins von `d-check` (Nr. 11 d), dass der Kommentar hinter der SHA den Tag dieser SHA nennt, ein festes Runner-Abbild und Docker (Nr. 11 b) und dass jemand ein Rot oder einen ausbleibenden Lauf liest (Nr. 11 e, f). Ebenso fehlt die Migrationskette des Schemas; sie braucht eine zweite Schemaversion.

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

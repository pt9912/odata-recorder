# Slice slice-gate-integration-determinism: make test-integration, make test-determinism und make test-contract in der Gate-Kette

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice. Er gehört nicht zu [welle-sensoren](../welle-sensoren.md), weil deren Closure-Trigger (Backfill und Schwelle der Abdeckung) von seiner DoD verschieden ist und er keinen Slice dieser Welle voraussetzt oder blockiert; eine Zugehörigkeit würde die Closure dieser Welle an die Laufzeit zweier Tests binden (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): er ändert `harness/mk/` und `harness/README.md`; `Dockerfile`, Treiber und Gegenproben ändert er nach dem Architect nicht (§3). Er hält die Gate-Kette, damit gilt Nr. 7 (ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Gate-Kette, der Bezug ist der nächste, den es gibt), `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-RB-03` (Kriterien, die `make test-integration` am Binary belegt, [rückverfolgung](../rueckverfolgung.md)), `LH-QA-02` (Determinismus, `make test-determinism`), `LH-QA-05` und `LH-RB-01` (Referenzanfragen und Negativfälle, `make test-contract`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Muster *Ziel neben der Kette*; Aufnahme in die Kette ist eine Verschärfung, keine Lockerung, `AGENTS.md` §3.6), Register-Eintrag `BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger` (2×, Ausgang `geplant` in diesem Slice).

**Berührte Spec-Stellen:** `SPEC-211` (neu, Architect: Bindung, gleichzeitige Läufe, Zeitbudget, Randformen der Kette) · `SPEC-037`, `SPEC-038` und `SPEC-042` (Stand, Treiber, *Ungeprüft* Verdrahtung; Nachzug durch den Architect) · `harness/vertraege/README.md` (Zeile von `SPEC-211`) · `harness/README.md` §Sensors (drei Zeilen wechseln von *Werkzeuge* in die Gate-Tabelle, Spalte Vertrag höchstens 200 Zeichen) und *Nicht behauptet*. `SPEC-102` bleibt unberührt (Auftrag 4: kein Fall der Gegenproben).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner der Closure von `welle-record-replay`). **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** `make test-integration` (das gebaute Binary im Container, Record → Upstream stoppen → Replay), `make test-determinism` (100 Läufe des Replay nach Reset) und `make test-contract` (Referenzanfragen `ODV4-*` und je Fehlerklasse von `LH-QA-05` ein Negativfall über HTTP, `SPEC-042`) hängen an `GATE_CHECKS`; `make gates` fährt sie, der Nachweis `record-gates` deckt sie, und die Pflichtläufe vor der Closure einer Welle entfallen als Rat im Text (Entscheidung des Users vom 2026-11-06, Option A: beide in die Kette; Entscheidung des Users vom 2026-11-08, `make test-contract` ebenso, „bitte lösen“; neue Gates sind erlaubt, `AGENTS.md` §3.6).

**Hintergrund:** Alle drei Ziele wurden daneben gebaut (`GATE_CHECKS` unverändert); der Pflichtlauf vor der Closure hatte keinen Träger (`BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger`), der Planner fuhr ihn von Hand. Die Aufnahme ist laut `harness/mk/test-integration.mk`, `harness/mk/test-determinism.mk` und `harness/mk/test-contract.mk` eine Zeile je Ziel; die Arbeit des Slice ist, was an der Kette daran hängt. `make test-contract` baut die Stufe `contract` des `Dockerfile` und startet je Test-Binary einen Container ohne Netz; er teilt mit den beiden anderen die Frage nach dem Image-Tag und dem Speicherlimit in der parallelen Kette.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neue Stufen oder Szenarien von `make test-integration` (die Szenarien von `slice-bootstrap-m3`, `-m4` und `-m6`) und neue Fixtures oder Negativfälle von `make test-contract` — sie gehören den Slices, die ihr Verhalten liefern; dieser Slice ändert nur die Bindung.
- Eine Änderung der Fristen von 120 s je Stufe (`SPEC-037`, `SPEC-038`) — die Fristen sind Grenzen der Tests, nicht der Kette; der Architect nennt nur das Zeitbudget der Kette.
- Die Gegenproben zu den Treibern `tools/harness/test-integration.sh` und `tools/harness/test-determinism.sh` — `SPEC-037` und `SPEC-038` führen die Verdrahtung als akzeptiertes Negativ (Nr. 10 f); der Architect hat in Auftrag 4 entschieden, dass es bleibt (`SPEC-211` Nr. 5 a), ein Fall der Gegenproben ist kein Muss.
- Der Fänger der Markierung `capture_status` am Binary (`SPEC-037` Nr. 10 h) — Adresse `slice-bootstrap-m6`, dessen DoD sie führt.
- Eine Aufnahme von `make gegenproben` in die Kette — [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) hält sie draußen (Laufzeit); die Aufnahme dieses Slice ändert das nicht.
- Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills — nur als Diff-Vorschlag in der Closure; `close-welle.md` trägt den Pflichtlauf bis zu diesem Slice, sein Wegfall ist ein Diff-Vorschlag an den User.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] `GATE_CHECKS += test-integration test-determinism test-contract` (jede Zeile in der Datei ihres Ziels, `harness/mk/test-integration.mk`, `harness/mk/test-determinism.mk` und `harness/mk/test-contract.mk`), die Kopf-Kommentare der drei Dateien nennen nicht mehr „KEIN GATE“, die `make help`-Zeilen auch nicht; `make gates` fährt alle drei und ist grün.
- [ ] Die Läufe zugleich in der Kette sind nach dem Vertrag entschieden und belegt: die Treiber bauen die Stufen `integration` und `contract` des `Dockerfile`; die Checks der Kette laufen parallel (`Makefile`), der Architect entscheidet, ob die Läufe sich bei gleichem Image-Tag überschreiben können, und der Implementer weist die Lösung an der Kette nach (Auftrag 2).
- [ ] Zeitbudget der Kette gemessen und im Vertrag genannt (Auftrag 3); `harness/README.md` führt alle drei Ziele in der Tabelle der Gates mit Bindung und nennt in *Nicht behauptet* nur noch, was weiter ungeprüft ist.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `SPEC-037`, `SPEC-038`, `SPEC-042`, `harness/README.md`, `.claude/commands/close-welle.md` (Pflichtlauf) als Diff-Vorschlag.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/mk/test-integration.mk`, `harness/mk/test-determinism.mk`, `harness/mk/test-contract.mk` | update | je `GATE_CHECKS += …`, Kopf und Hilfetext |
| `harness/README.md` | update | drei Zeilen aus *Werkzeuge* in die Gate-Tabelle (Wortlaut von Bindung und Vertrag unten), *Nicht behauptet*; Zellen höchstens 200 Zeichen |
| `harness/vertraege/SPEC-211-container-tests-in-der-kette.md`, `harness/vertraege/README.md` | neu bzw. update (Architect, vor dem Code) | Bindung, gleichzeitige Läufe, Zeitbudget, Randformen der Kette; Index-Zeile |
| `harness/vertraege/SPEC-037-test-integration.md`, `harness/vertraege/SPEC-038-test-determinism.md`, `harness/vertraege/SPEC-042-test-contract.md` | update (Architect, vor dem Code) | *Stand*, Zeile zur Bindung in Nr. 2, *Ungeprüft* Verdrahtung |
| `Dockerfile`, `tools/harness/test-integration.sh`, `tools/harness/test-determinism.sh`, `tools/harness/test-contract.sh` | keine Änderung | Auftrag 2: derselbe Tag trägt denselben Inhalt (`SPEC-211` Nr. 2 a); kein fester Container-Name, `--rm` |
| `tools/gegenproben/`, `harness/mk/gegenproben.mk`, `tools/harness/gegenproben-eingabe.sh`, Stempel | keine Änderung | Auftrag 4: kein Fall der Gegenproben (`SPEC-211` Nr. 5 a) |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Aufträge 1 bis 5, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); schwere Läufe (`make gates`, `make test-integration`, `make test-determinism`) unter `flock /tmp/odata-recorder-heavy.lock`; jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; Commit-Messages per `git commit -F <datei>`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Aufträge an den Architect:**

1. **Bindung.** Ob die Aufnahme in die Kette eine ADR braucht (Linie C: ADR-Nummern 0030 bis 0039; [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) betrifft die Gegenproben, nicht diese Ziele) oder `SPEC-037` und `SPEC-038` genügen; Wortlaut der Bindung für die Spalte *Bindung* der Gate-Tabelle.
2. **Gleichzeitige Läufe in der Kette.** `test-integration` und `test-determinism` bauen die Stufe `integration`, `test-contract` die Stufe `contract`; die Checks laufen parallel. Image-Tags, Aufräumen der Container, Verhalten bei Abbruch; der Fänger, der zwei Läufe zugleich trennt, oder das benannte Negativ mit Grund und Folge (`AGENTS.md` §3.10 für Nebenläufigkeit: Wirkungs-Fänger, nicht Datenrennen).
3. **Zeitbudget und Re-Evaluierungs-Trigger.** Zuwachs von `make gates` durch beide Ziele, gemessen im Hauptbaum unter ruhiger Last (Messung der Closure von `welle-record-replay` als Ausgangswert: `make test-integration` 5,9 s und `make test-determinism` 2,1 s bei vorhandenen Images, `pruefung test` 92,6 s und `pruefung race` 231,5 s zum Vergleich, Last des Hosts 5,5 bis 5,8; ein kalter Bau der Stufe `integration` ist nicht gemessen); der Trigger „Zuwachs über 120 s“ ist der Vorschlag des Planner, der Architect setzt die Zahl mit Grund und nennt die Folge, wenn er eintritt (Gate wieder neben die Kette, Pflichtlauf mit Träger, oder Beschleunigung).
4. **Verdrahtung und Gegenproben.** Ob ein Fall der Gegenproben die Aufnahme fängt (Kette ruft alle drei Ziele) oder `SPEC-037` Nr. 10 f, `SPEC-038` Nr. 10 f und die entsprechende Stelle von `SPEC-042` als akzeptiertes Negativ mit Grund und Folge bleiben; die Gegenprobe darf das Produkt nicht mitfahren (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
5. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): Docker nicht erreichbar, Image fehlt, Stufe scheitert beim Bau, Frist von 120 s überschritten, Lauf unter dem Speicherlimit, Abbruch durch INT oder TERM mitten in der Kette; die Tabelle *Zusage · Fänger* (§3), die erwartete Produktcode-Größe für die Marke des Planners.

**Ausgang des Architect** (2026-11-17; Vertrag `SPEC-211`, Nachzug in `SPEC-037`, `SPEC-038`, `SPEC-042`):

1. **Bindung — keine ADR.** `SPEC-211` Nr. 1 trägt die Entscheidung mit Alternativen und Gründen: [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) betrifft die Gegenproben, die Aufnahme verschärft die Kette und senkt keine Schwelle; die Herausnahme eines Ziels wäre eine Lockerung und bräuchte dann eine ADR, die das ablöst. Wortlaut der Spalte *Bindung* für alle drei Zeilen der Gate-Tabelle: `Reproduzierbarkeit (LH-QA-06), Zeitbudget SPEC-211 Nr. 4`. Vorschlag der Spalte *Vertrag* (je höchstens 200 Zeichen, der Implementer zählt nach): `make test-integration`: „Binary im Container ohne Netz: Record, Upstream stoppen, Replay in neun Stufen (SPEC-037)“; `make test-determinism`: „100 Läufe des Replay nach Reset: gleiche normalisierte Antworten und Cursor-Stände (SPEC-038)“; `make test-contract`: „36 Fixtures ODV4-* über HTTP und je Fehlerklasse von LH-QA-05 ein Negativfall (SPEC-042)“.
2. **Gleichzeitige Läufe — Tag bleibt, nichts ändert sich.** `make gates` fährt die Checks nacheinander, erst `make -j gates` zugleich. Zwei Bauten von `odata-recorder:integration` aus demselben Quellstand liefern denselben Inhalt; kein fester Container-Name, `--rm` (`SPEC-211` Nr. 2 a und b). Gemessen (Architect): `make -j3 test-integration test-determinism test-contract` zweimal grün (8,7 s und 8,2 s). Abbruch und SIGKILL: `SPEC-211` Nr. 2 c und d. Der Fänger der Zusage ist der Nachweislauf von Nr. 2 e (zehn Läufe, Mutation *fester `--name`*), mit Rate; nicht in jedem Lauf rot heißt akzeptiertes Negativ mit dieser Rate.
3. **Zeitbudget — Soll mit Messvorschrift, Trigger 120 s.** Gemessen (2026-11-17, Last 3 bis 8 auf 20 Kernen): Zuwachs rund 12 s bei vorhandenen Images, rund 54 s nach geändertem Quelltext (Bau der Stufen `integration` 20,6 s und `contract` 21,8 s). Trigger: Zuwachs über 120 s bei gültiger Messung; Folge: der Planner legt dem User Beschleunigung, Anhebung mit Messung oder Herausnahme (ADR) vor (`SPEC-211` Nr. 4). Offen bleibt ein Lauf auf einem Host ohne Fremdlast; ein kalter Bau von `source` und `build` gehört zu `build`, `test` und `test-race`.
4. **Verdrahtung — kein Fall der Gegenproben.** Akzeptiertes Negativ in `SPEC-211` Nr. 5 a (dieselbe Klasse wie `test-race`, `zeichen-check`, `katalog-check`, `handbuch-check`); `make docs-check` unterscheidet nicht, ob ein Ziel an `GATE_CHECKS` hängt (Probe des Architect: die Zeile allein ändert die Befunde nicht). `make -n gates` nennt die drei Rezepte; der Verifier prüft es einmal.
5. **Randformen** stehen in `SPEC-211` Nr. 3; das Raster steht in §6.

**Erwartete Größe** (für die Marke des Planner, `SPEC-128` Nr. 1): Produktcode ohne Verträge etwa 70 Zeilen (drei Fragmente je rund 11 Zeilen Einfügungen und Löschungen, `harness/README.md` rund 8, dieser Plan rund 30 mit Nachweis und Closure); die Verträge des Architect (`SPEC-211` 44 Zeilen, drei Nachzüge je 3 bis 5, Index 1) liegen vor der Basis, wenn der Planner die Basis nach dem Commit des Architect setzt, sonst kommen rund 60 Zeilen dazu. Zeilen der Tabelle unten nach Klasse (12 Zeilen, Planner gezählt): *a* (nebenläufig) 2, *b* (andere) 10, *f* 0, *p* 70.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15). Der Fänger ist eine Mutation, die der Implementer selbst rot sieht, oder ein benanntes Negativ nach `SPEC-211`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Jedes der drei Ziele hängt an `GATE_CHECKS` und `make gates` fährt es (Menge: drei Ziele, Schleife) | kein automatischer Fall: Negativ `SPEC-211` Nr. 5 a (Grund und Folge dort); Handprobe `make -n gates` nennt je Ziel den Treiber, mit der Mutation fehlt er (der Bericht nennt die Zahl der Treffer, drei gegen zwei) | die Zeile `GATE_CHECKS += <Ziel>` fehlt, je Ziel | andere |
| 2 | Ein rotes Ziel macht `make gates` rot, und `record-gates` stempelt nicht | ein Lauf `make -k gates` mit den drei Mutationen zugleich: alle drei Ziele rot in der Ausgabe, `.harness/state/gates-passed.head` unverändert (Bericht: Zeitstempel vor und nach) | je Ziel ein `t.Fatal` im Test (`TestE2EProzess`, `TestE2EDeterminismus`, `TestVertragFehlerklassen`) | andere |
| 3 | Die Läufe zugleich überschreiben einander nicht und kollidieren nicht (`SPEC-211` Nr. 2 a, b) | `make -j3 test-integration test-determinism test-contract` zehnmal; ohne Mutation zehnmal grün, mit Mutation rot in *k* von 10 (Rate im Bericht, mit Last des Hosts und Messendem); unter 10 von 10 steht die Zusage als Negativ mit der Rate (`SPEC-211` Nr. 2 e) | fester `--name e2e-fest` an `docker run` in `laufe` von `test-integration.sh` und im Lauf von `test-determinism.sh` (dieser Treiber hat kein `laufe`) | nebenläufig |
| 4 | Der Abbruch durch INT oder TERM endet jeden Treiber und entfernt den Container; es gibt keinen Nachweis (`SPEC-211` Nr. 2 c) | Handlauf je Treiber: `sh tools/harness/test-<Ziel>.sh` im Hintergrund, `kill -TERM` an den Treiber, sobald ein Container läuft; ohne Mutation Exit 2, mit Mutation Exit 143 (Exit-Hälfte, Schleife über die drei Treiber). Die Container-Hälfte (`docker ps` leer) unterscheidet nur `test-determinism` (0,2 bis 0,3 s gegen 0,9 bis 1,3 s), für `test-integration` und `test-contract` ist sie ein akzeptiertes Negativ (Absatz *Nachweis* unten); kein automatischer Fall (`SPEC-211` Nr. 5 a, d) | `trap abbruch INT TERM` entfernt, je Treiber | nebenläufig |
| 5 | Ohne `docker` oder Daemon endet jedes Ziel mit Exit 2 und der Meldung `Aufbau der Stufe … gescheitert` (`SPEC-211` Nr. 3, Zeile 1) | Handlauf je Ziel: `DOCKER=/bin/false make <Ziel>`, erwartet Exit 2 und die Meldung; Schleife über drei Ziele; kein automatischer Fall (kein Fall prüft Docker, `SPEC-037` Nr. 10 f) | die Prüfung des Baus (`if !` um `docker build`) entfernt | andere |
| 6 | Ein Bau, der scheitert (`go test -c` übersetzt nicht), ist Exit 2 (`SPEC-211` Nr. 3, Zeile 2) | Handlauf `make test-integration`: Exit 2, Meldung `Aufbau der Stufe integration gescheitert`; `make test-determinism` ebenso; `make test-contract` mit `src/bootstrap` entsprechend | ein Syntaxfehler in `tests/e2e` (nicht committet) | andere |
| 7 | Image fehlt nach dem Bau: Exit 2 bzw. rot mit `Exit 125` (`SPEC-211` Nr. 3, Zeile 3) | akzeptiertes Negativ: Grund: nur ein `docker rmi` zwischen Bau und Lauf erzeugt die Form; Folge: Exit 1 oder 2, die Kette ist in beiden Fällen rot | — | andere |
| 8 | Die Frist von 120 s macht die Stufe rot (`SPEC-211` Nr. 3, Zeile 4) | kein Fänger (`SPEC-037` Nr. 10 f: kein Szenario braucht die Frist); Negativ in `SPEC-211` Nr. 3 | `-test.timeout` entfernt | andere |
| 9 | Das Limit von 2 GiB gilt (`SPEC-211` Nr. 3, Zeile 5) | kein Fänger (`SPEC-037` Nr. 10 f); Negativ in `SPEC-211` Nr. 3 | `--memory 2g` entfernt | andere |
| 10 | Zeitbudget: Zuwachs gemessen, Trigger 120 s (`SPEC-211` Nr. 4) | kein Prüfer (Soll mit Messvorschrift, Negativ `SPEC-211` Nr. 5 c); der Implementer misst in der Closure neu und trägt den Wert unten ein | — | andere |
| 11 | Die drei Ziele stehen in der Gate-Tabelle von `harness/README.md`, mit Bindung, Vertrag-Zelle höchstens 200 Zeichen, *Nicht behauptet* nennt nur noch den Rest | `make docs-check` (`gate-undocumented`, Größengrenze der Zellen); der Ort Gate- oder Werkzeug-Tabelle ist nur durch den Leser gefangen (Negativ `SPEC-211` Nr. 5 b) | die Zeile eines Ziels fehlt, die Zelle hat 201 Zeichen | andere |
| 12 | Kein Text sagt mehr „KEIN Gate“ oder „hängt nicht an `GATE_CHECKS`“ für die drei Ziele | Handprobe `grep -rn 'hängt nicht an .GATE_CHECKS\|KEIN Gate\|KEIN GATE' harness tools/harness` ohne Treffer zu den drei Zielen; kein automatischer Fall (Text, `AGENTS.md` §3.11) | der Satz bleibt in einer Datei | andere |

**Nachweis** (Implementer, 2026-11-18, Host mit 20 Kernen und Fremdlast, Load Average 1 Minute 9 bis 17 beim Start, Messender: der Implementer mit `date` und `docker ps`, kein Detektor; je Zahl Größenordnung, kein fester Wert). Jede Mutation einmal gefahren, danach `git checkout` der Datei; *Zusage · Mutation · roter Test*:

| Nr. | Mutation | Ergebnis (roter Test aus dem benannten Grund) |
|---|---|---|
| 1 | `GATE_CHECKS`-Zeile je Ziel entfernt | `make -n -k gates`: 3 Treffer des Treibers ohne Mutation, 2 mit Mutation je Ziel (3 von 3 Zielen); `make -n gates` allein bricht im Zweig am Rezept von `schema-check`, deshalb `-k` |
| 2 | `t.Fatal("mutation")` in `TestE2EProzess`, `TestE2EDeterminismus`, `TestVertragFehlerklassen` zugleich | ein Lauf `make -k gates`: `test-integration: Stufe 1: TestE2EProzess rot (Exit 1)`, `test-determinism: rot (Exit 1)`, `test-contract: rot (/out/integration.test: Exit 1)`, je `mutation` im Text; `record-gates` lief nicht, `.harness/state/gates-passed.head` entstand nicht (der Zweig trug vorher keine Datei). Dieselbe Mutation färbt auch `test`, `test-race` und `lint` rot (Test und unerreichbarer Code), und `docs-check` ist durch die Roadmap rot: die drei Ziele sind in der Ausgabe einzeln zu unterscheiden |
| 3 | fester `--name e2e-fest` (siehe Tabelle) | `make -j3 test-integration test-determinism test-contract` zehnmal: ohne Mutation 0 von 10 rot (Load 11,7 bis 16,9), mit Mutation **10 von 10 rot**, in allen zehn Logs `Conflict. The container name "/e2e-fest" is already in use`, Exit 2 (Load 9,6 bis 11,4) |
| 4 | `trap abbruch INT TERM` entfernt, je Treiber | Exit-Hälfte: ohne Mutation Exit 2 in 7 von 7 gewerteten Läufen (3 Treiber), mit Mutation Exit 143 in 8 von 8; Container-Hälfte nur `test-determinism` (ohne Mutation weg nach 0,2 bis 0,3 s, mit Mutation 0,9 bis 1,3 s, je 3 von 3). Läufe ohne erschienenen Container (Wettlauf mit dem Ende der kurzen Stufe) zählen nicht |
| 5 | `if ! fahre … docker build` entfernt (`if false`), `DOCKER=/bin/false make <Ziel>` | ohne Mutation Exit 2 und die Meldung `Aufbau der Stufe … gescheitert` in 3 von 3 Zielen; mit Mutation Exit weiterhin 2 (3 von 3), die Meldung fehlt in 3 von 3: der Fänger ist die Meldung, nicht der Exit |
| 6 | dieselbe Mutation plus Syntaxfehler (`x x x`) in `tests/e2e/stufe_prozess_test.go` bzw. `src/bootstrap/abnahme_etag_test.go` | ohne Mutation Exit 2 und die Meldung je Ziel (3 von 3, `integration`, `integration`, `contract`); mit Mutation Exit 0 und `grün` in 3 von 3 (der Lauf nimmt das alte Image) |
| 7 | — | akzeptiertes Negativ nach Tabelle: nur ein `docker rmi` zwischen Bau und Lauf erzeugt die Form; Folge: Exit 1 oder 2, die Kette ist rot |
| 8, 9 | `-test.timeout 120s` und `--memory 2g` in allen drei Treibern entfernt | `make test-integration test-determinism test-contract` **grün** (Exit 0): kein Fänger, wie der Vertrag sagt (`SPEC-211` Nr. 3), Negativ bestätigt |
| 10 | — | kein Prüfer; Messung unten |
| 11 | Zeile von `make test-contract` aus der Tabelle der Gates entfernt; Zelle mit 201 Zeichen | `make docs-check`: `gate-undocumented` (`harness/mk/test-contract.mk:9`); `section-cell-oversized` mit 201 Zeichen, die Zelle mit 200 Zeichen war grün (beide Seiten der Grenze); sonst die fünf Befunde der Roadmap |
| 12 | der Satz `KEIN GATE` bleibt (Stand `HEAD` vor dem Slice) | `git show HEAD:harness/mk/test-contract.mk` trägt einen Treffer; der Stand des Zweigs keinen (`grep -rn` ohne Treffer zu den drei Zielen) |

**Akzeptiertes Negativ zu Nr. 4 (Container-Hälfte von `SPEC-211` Nr. 2 c):** für `test-integration` und `test-contract` unterscheidet der Handlauf die Mutation nicht, denn die Stufen laufen nur Zehntelsekunden und der Container endet von selbst (`docker ps` leer nach 0,1 bis 0,2 s mit und ohne `trap`). *Grund:* kein Szenario hält einen Container lange genug, um zu zeigen, ob das Signal ihn erreicht. *Folge im Betrieb:* ein entferntes `trap` fällt nur durch Exit 143 statt 2 auf, und ein Container überlebt höchstens die laufende Stufe (Frist 120 s). Der Vertrag sagt „jeder Container endet per Signalweitergabe“ für alle drei; dass das nur für den Exit und für `test-determinism` gemessen ist, gibt §6 an den Architect zurück.

**Zeitbudget** (`SPEC-211` Nr. 4, Messvorschrift: einzeln nacheinander, Images vorhanden; Host mit 20 Kernen, Load 9 bis 11 beim Start, also nicht über der Kernzahl, gültig; Messender: der Implementer, `date`): zwei Läufe, `make test-integration` 7,0 s und 6,7 s, `make test-determinism` 2,3 s und 2,4 s, `make test-contract` 2,9 s und 2,9 s, zusammen 12,2 s und 12,0 s. Bau der Stufen nach geändertem Quelltext (`--no-cache-filter <Stufe>`): `integration` 20,6 s, `contract` 22,8 s; Zuwachs von `make gates` damit rund 55 s, unter dem Trigger von 120 s. Ein Host ohne Fremdlast ist nicht gemessen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Reihenfolge der Kette** (Planner, 2026-11-08, Entscheidung des Users zur Closure von `welle-mock`): dieser Slice ist der **nächste Kettenslice** (Nr. 7 der Übergangsregeln); danach `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu`, immer nur einer zugleich. `slice-handbuch-mock` und die weiteren Kettenslices von `welle-benutzerhandbuch` folgen oder gehen vor, wenn der User es entscheidet; der Planner startet keinen, solange einer die Kette hält. Es dürfen höchstens drei Slice-Worktrees zugleich laufen.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (Closure-Notiz `welle-record-replay-results`); kein anderer Slice, der `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` oder den Stempel ändert, liegt in `in-progress/` (Übergangsregeln Nr. 7). `slice-mutationslauf-werkzeug` liegt in `in-progress/`; sein Plan nimmt die Gate-Kette nicht an (Ausdrücklich NICHT, Auftrag 7 an den Architect): meldet er im Lauf, dass er sie doch ändert, wartet dieser Slice auf dessen `done/`. `slice-guard-schreibwege` berührt die Kette nur, wenn eine Gegenprobe gewählt wird; beide laufen nie zugleich.

**Größenmarke** (Planner, 2026-11-18, `SPEC-128` Nr. 1 und `AGENTS.md` §3.14, vor dem ersten Code-Commit): die Tabelle *Zusage · Fänger* in §3 hat 12 Zeilen, davon 2 der Klasse `nebenläufig` (Nr. 3, Nr. 4), *a* = 2, *b* = 10, *f* = 0, *p* = 70 (Erwartung des Architect); *S* = 85·2 + 80·10 + 200·0 + 70 = 1040, *M* = 100·⌈135·1040/10000⌉ = 1500. **Schnittfrage: nein** (*S* < 2223); es wird vor dem Code nicht geschnitten. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen, `harness/vertraege/` ausgeschlossen. Die Basis ist der Commit des Architect, der den Vertrag `SPEC-211` abschließt; die Marke `1600` der Prosa-Fassung ersetzt der Block, weil sie aus 13 geschätzten Zeilen (*b* = 11) und einer Basis vor dem Architect gerechnet war.

```groessenmarke
pfad: harness/mk/test-integration.mk
pfad: harness/mk/test-determinism.mk
pfad: harness/mk/test-contract.mk
pfad: harness/README.md
pfad: Dockerfile
pfad: tools/harness/test-integration.sh
pfad: tools/harness/test-determinism.sh
pfad: tools/harness/test-contract.sh
produktcode: 70
fakes: 0
basis: 2b88588975e1e87774b71e8c3b609bed1809c9ab
schaetzung: 1040
marke: 1500
ersetzt: 1600 | Prosa-Marke aus 13 geschätzten Tabellenzeilen (b = 11) und Basis vor dem Architect; die Tabelle hat 12 Zeilen (b = 10).
```

**Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `4a6a760b`, Stempel der Gegenproben gleich diesem Stand): `welle-record-replay` liegt in `done/`; `slice-handbuch-katalog-schluessel-kommandos` liegt in `done/`, damit hält kein Slice mehr die Gate-Kette (Nr. 7). In `in-progress/` laufen `slice-admin-authentisierung` (Pfad `src/adapters/driving/admin/`) und `slice-bootstrap-m3` (Linie B); beide ändern laut ihrem Plan `Makefile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und Stempel nicht. `slice-mutationslauf-werkzeug` liegt in `done/`, `slice-guard-schreibwege` in `open/` und startet nicht zugleich. **Kontaktstelle zu `slice-bootstrap-m3`:** dessen Plan führt die Szenarien von `make test-integration` (Treiber `tools/harness/test-integration.sh`, Stufe `integration` des `Dockerfile`, `tests/e2e`) und gibt eine Änderung des `Dockerfile` als Rückgabe an den Planner zurück; dieser Slice kann denselben Treiber und dieselbe Stufe ändern (§3, Aufträge 2 und 4). Beide Zweige fassen dort nie dieselbe Zeile an, sondern der Architect hält die Änderung am Treiber klein (Image-Tag, Aufräumen); wer als zweiter nach `main` merged, gleicht die Datei ab. Ändert dieser Slice das `Dockerfile`, meldet der Implementer es dem Planner vor dem Commit, damit `slice-bootstrap-m3` davon erfährt.

**Zurückgestellt** (`in-progress` → `open`, Planner, 2026-11-08, Grund nachgetragen): zugunsten von `welle-mock` (Vorrang des Users). Der Start-Commit `43d67170` trug keinen Code, der Architect war nicht gelaufen, Zweig und Worktree sind gelöscht; die Gate-Kette (Nr. 7 der Übergangsregeln) ist frei und geht an `slice-test-contract`. Beim erneuten Start gilt: `slice-bootstrap-m3-e2e-mock` ändert `tools/harness/test-integration.sh` (Stufen 8 und 9), `slice-test-contract` ändert `Makefile` und `harness/mk/`; dieser Slice liest beide vor dem Architect-Auftrag neu und setzt `Basis` und Marke neu (die Basis `4a6a760b` oben gilt nicht mehr). Der Start verlangt zusätzlich, dass `slice-test-contract` nicht in `in-progress/` liegt.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): die Läufe lassen sich in der Kette nicht trennen (Auftrag 2), oder das Zeitbudget liegt über dem Trigger (Auftrag 3) und der User entscheidet neu.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum mit beiden Zielen in der Kette, Closure-Notiz in §7 mit Lerneintrag (Zeitbudget mit Messung; der Pflichtlauf in `close-welle.md` entfällt als Diff-Vorschlag).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die Kette wird durch drei Container-Läufe merklich langsamer, und die Fristen der Gegenproben und von `pruefung` (für einen ruhigen Host bemessen) reißen unter Last — **Ausgang:** weiter offen: gemessen sind rund 12 s (Images vorhanden) und rund 55 s (nach geändertem Quelltext) Zuwachs (Implementer, §3 *Zeitbudget*), Trigger 120 s in `SPEC-211` Nr. 4; die Closure misst neu.
- Zwei Checks bauen dieselbe Stufe `integration` zugleich und überschreiben ein Image-Tag — **Ausgang:** entfallen: entschieden in `SPEC-211` Nr. 2 (derselbe Inhalt, keine Änderung); der Nachweislauf (Nr. 2 e) ist gefahren: ohne Mutation 0 von 10 rot, mit fester `--name` 10 von 10 rot (§3 *Nachweis*, Nr. 3).
- Größe: die Marke steht als Block in §4 (*S* = 1040, *M* = 1500, keine Schnittfrage; *a* = 2, *b* = 10, *f* = 0, *p* = 70) — **Ausgang:** weiter offen: die Messung am Ende gegen die Marke und die Auslastung (`SPEC-128` Nr. 9) trägt der Planner in der Closure ein.
- Berührung mit `slice-bootstrap-m3` und `slice-bootstrap-m3-e2e-mock` am Treiber `tools/harness/test-integration.sh` und an der Stufe `integration` des `Dockerfile` — **Ausgang:** entfallen: beide liegen in `done/` (Archiv von `welle-mock`), ihre Stufen 8 und 9 stehen im Treiber und in `SPEC-037`; dieser Slice ändert weder Treiber noch `Dockerfile` (Auftrag 2).
- Die Aufnahme verdeckt einen Rest, den die Tests nicht prüfen (`SPEC-037` Nr. 10, `SPEC-038` Nr. 10, `SPEC-042` Nr. 9): „Gate grün“ wird als „Binary fehlerfrei“ gelesen — **Ausgang:** weiter offen: *Nicht behauptet* in `harness/README.md` nennt den Rest (DoD).
- Ein Ziel, das an der Kette hängt, steht in der falschen Tabelle von `harness/README.md`, und die Verdrahtung `GATE_CHECKS` fängt kein Fall (`SPEC-211` Nr. 5 a, b) — **Ausgang:** weiter offen: akzeptiertes Negativ mit Grund und Folge im Vertrag; `make -n gates` prüft der Verifier einmal.
- `docs/plan/planning/in-progress/roadmap.md` führt fünf Befunde von `make docs-check` (Zeile 11: Ruhe-Marker *Nichts in Arbeit.* trotz Slice in `in-progress/`; Zeilen 27, 159, 168, 170: Links auf `../open/slice-gate-integration-determinism.md`), die die Roadmap als Dokument des Integrationsschritts nur der Planner schreibt; `make gates` ist im Zweig deshalb rot (Implementer, 2026-11-18, ohne diesen Slice entstanden, jede andere Zeile der Kette einzeln grün) — **Ausgang:** offen; Adresse: Planner (Integrationsschritt, Übergangsregeln Nr. 4).
- `SPEC-211` Nr. 2 c sagt zu, dass der laufende `docker run` jedes Treibers per Signalweitergabe endet und `--rm` den Container entfernt; der Handlauf (§3 *Nachweis*, Nr. 4) belegt Exit 2 für alle drei Treiber, den Container-Teil aber nur für `test-determinism` (die Stufen von `test-integration` und `test-contract` enden von selbst nach Zehntelsekunden). Das Negativ steht im Plan, nicht im Vertrag — **Ausgang:** offen; Adresse: Architect (Vertrag `SPEC-211` Nr. 2 c um das Negativ ergänzen oder die Zusage auf den Exit schmälern). Zugleich nennt Nr. 2 e die Mutation *fester `--name` in `laufe`*; `test-determinism.sh` hat kein `laufe`, die Mutation sitzt dort am `docker run` (Wortlaut, keine Entscheidung).
- `.claude/commands/close-welle.md` trägt den Pflichtlauf vor der Closure einer Welle, dessen Wegfall der Slice nicht selbst ändern darf — **Ausgang:** offen; Adresse: Planner (Diff-Vorschlag an den User in der Closure, DoD Punkt 6). Eine ADR und ein ADR-Index-Eintrag entstehen nicht (Auftrag 1); die Roadmap nennt den Slice bereits.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16). Die Operation ist der Lauf eines der drei Ziele; die Zelle nennt den Ort der Entscheidung. `SPEC-211` Nr. 3 ist die Tabelle der Randformen (Zeile = Form).

| Argumentart | `make test-integration` | `make test-determinism` | `make test-contract` |
|---|---|---|---|
| `docker` fehlt oder Daemon nicht erreichbar | `SPEC-211` Nr. 3 Zeile 1, `SPEC-037` Nr. 2 (Exit 2) | `SPEC-211` Nr. 3 Zeile 1, `SPEC-038` Nr. 2 | `SPEC-211` Nr. 3 Zeile 1, `SPEC-042` Nr. 2 |
| Stufe scheitert beim Bau | `SPEC-211` Nr. 3 Zeile 2, `SPEC-037` Nr. 2 | `SPEC-211` Nr. 3 Zeile 2, `SPEC-038` Nr. 2 | `SPEC-211` Nr. 3 Zeile 2, `SPEC-042` Nr. 2 |
| Image fehlt nach dem Bau | `SPEC-211` Nr. 3 Zeile 3 (`-test.list`, Exit 2) | `SPEC-211` Nr. 3 Zeile 3 (`Exit 125`, Exit 1) | `SPEC-211` Nr. 3 Zeile 3 (`-test.list`, Exit 2) |
| Frist von 120 s überschritten | `SPEC-211` Nr. 3 Zeile 4, `SPEC-037` Nr. 2 | `SPEC-211` Nr. 3 Zeile 4, `SPEC-038` Nr. 8 d | `SPEC-211` Nr. 3 Zeile 4, `SPEC-042` Nr. 2 |
| Limit von 2 GiB überschritten | `SPEC-211` Nr. 3 Zeile 5 | `SPEC-211` Nr. 3 Zeile 5 | `SPEC-211` Nr. 3 Zeile 5 |
| Abbruch durch INT oder TERM | `SPEC-211` Nr. 2 c, `SPEC-037` Nr. 2 | `SPEC-211` Nr. 2 c, `SPEC-038` Nr. 8 c | `SPEC-211` Nr. 2 c, `SPEC-042` Nr. 8 c |
| SIGKILL des Treibers | `SPEC-211` Nr. 2 d (Negativ) | `SPEC-211` Nr. 2 d (Negativ) | `SPEC-211` Nr. 2 d (Negativ) |
| Argument (auch leer) | `SPEC-037` Nr. 2 | `SPEC-038` Nr. 2 | `SPEC-042` Nr. 2 |
| zwei Läufe zugleich (`-j`), Wiederholung | `SPEC-211` Nr. 2 a, b | `SPEC-211` Nr. 2 a, b | `SPEC-211` Nr. 2 a, b; `SPEC-042` Nr. 8 f |
| `docker` antwortet nicht | `SPEC-211` Nr. 3 Zeile 9, `SPEC-037` Nr. 2 | `SPEC-211` Nr. 3 Zeile 9 | `SPEC-211` Nr. 3 Zeile 9 |
| Host unter Last, Frist reißt | `SPEC-211` Nr. 3 Zeile 10, Nr. 4 | `SPEC-211` Nr. 3 Zeile 10, Nr. 4 | `SPEC-211` Nr. 3 Zeile 10, Nr. 4 |
| Test-Binary fehlt oder leer (`E2E_BINARY`) | gilt nicht in der Kette: das Image trägt das Binary; Fall `SKIP` `SPEC-037` Nr. 10 g | `SPEC-038` Nr. 8 e | gilt nicht: der Treiber prüft Namen gegen `-test.list` (`SPEC-042` Nr. 2) |

Die Zeilen von `SPEC-211` Nr. 3 sind in der Reihenfolge der Tabelle gezählt: 1 `docker` fehlt, 2 Stufe scheitert, 3 Image fehlt, 4 Frist, 5 Speicher, 6 INT oder TERM, 7 Argument, 8 zwei Läufe zugleich, 9 `docker` antwortet nicht, 10 Host unter Last.

**Zusatztabelle** (die Zusage hängt vom Modus ab: seriell oder zugleich):

| Modus | Verhalten | Fänger |
|---|---|---|
| `make gates` (nacheinander) | kein gemeinsamer Zustand zwischen den Zielen außer dem Tag `odata-recorder:integration` von `test-integration` und `test-determinism` | `make gates` grün (DoD); die Gleichzeitigkeit fängt dieser Modus nicht |
| `make -j gates` oder `make -j3 <drei Ziele>` | höchstens drei Container zugleich (6 GiB), kein Namenskonflikt, dasselbe Image | Nachweislauf zehnmal mit Rate; Mutation *fester `--name`* (`SPEC-211` Nr. 2 e) |
| Abbruch im Modus zugleich | alle drei Treiber enden mit Exit 2, kein Container bleibt | Handlauf mit Mutation *`trap` entfernt* je Treiber, Exit-Hälfte für alle drei, Container-Hälfte nur für `test-determinism` (Tabelle und Negativ in §3 *Nachweis*) |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `welle-record-replay` (Entscheidung des Users vom 2026-11-06); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger` (2×, Ausgang `geplant` in diesem Slice); `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (4×, Ausgang `geplant` in `slice-coverage-gate`) berührt Auftrag 4; `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (4×, verkörpert) betrifft den Stempel, falls Gegenproben berührt werden.

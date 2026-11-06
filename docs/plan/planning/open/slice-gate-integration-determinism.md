# Slice slice-gate-integration-determinism: make test-integration und make test-determinism in der Gate-Kette

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice. Er gehört nicht zu [welle-sensoren](../welle-sensoren.md), weil deren Closure-Trigger (Backfill und Schwelle der Abdeckung) von seiner DoD verschieden ist und er keinen Slice dieser Welle voraussetzt oder blockiert; eine Zugehörigkeit würde die Closure dieser Welle an die Laufzeit zweier Tests binden (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): er ändert `harness/mk/`, `harness/README.md` und eventuell das `Dockerfile` und die Gegenproben, damit gilt Nr. 7 (ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Gate-Kette, der Bezug ist der nächste, den es gibt), `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-RB-03` (Kriterien, die `make test-integration` am Binary belegt, [rückverfolgung](../rueckverfolgung.md)), `LH-QA-02` (Determinismus, `make test-determinism`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Muster *Ziel neben der Kette*; Aufnahme in die Kette ist eine Verschärfung, keine Lockerung, `AGENTS.md` §3.6), Register-Eintrag `BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger` (2×, Ausgang `geplant` in diesem Slice).

**Berührte Spec-Stellen:** `SPEC-037` und `SPEC-038` (Bindung, Zeitbudget, *Ungeprüft* Treiber und Verdrahtung; Nachzug durch den Architect) · `SPEC-102` (Fälle der Verdrahtung, falls der Architect sie wählt) · `harness/README.md` §Sensors (zwei Zeilen wechseln von *Werkzeuge* in die Gate-Tabelle, Spalte Vertrag höchstens 200 Zeichen) und *Nicht behauptet*.
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

**Ziel:** `make test-integration` (das gebaute Binary im Container, Record → Upstream stoppen → Replay) und `make test-determinism` (100 Läufe des Replay nach Reset) hängen an `GATE_CHECKS`; `make gates` fährt sie, der Nachweis `record-gates` deckt sie, und die Pflichtläufe vor der Closure einer Welle entfallen als Rat im Text (Entscheidung des Users vom 2026-11-06, Option A: beide in die Kette).

**Hintergrund:** Beide Ziele wurden in `welle-record-replay` daneben gebaut (`GATE_CHECKS` unverändert); der Pflichtlauf vor der Closure hatte keinen Träger (`BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger`), der Planner fuhr ihn von Hand. Die Aufnahme ist laut `harness/mk/test-integration.mk` und `harness/mk/test-determinism.mk` eine Zeile je Ziel; die Arbeit des Slice ist, was an der Kette daran hängt.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neue Stufen oder Szenarien von `make test-integration` (die Szenarien von `slice-bootstrap-m3`, `-m4` und `-m6`) — sie gehören den Slices, die ihr Verhalten liefern; dieser Slice ändert nur die Bindung.
- Eine Änderung der Fristen von 120 s je Stufe (`SPEC-037`, `SPEC-038`) — die Fristen sind Grenzen der Tests, nicht der Kette; der Architect nennt nur das Zeitbudget der Kette.
- Die Gegenproben zu den Treibern `tools/harness/test-integration.sh` und `tools/harness/test-determinism.sh` — `SPEC-037` und `SPEC-038` führen die Verdrahtung als akzeptiertes Negativ (Nr. 10 f); ob der Architect das mit der Aufnahme ändert, entscheidet er in Auftrag 4, ein Fall der Gegenproben ist kein Muss.
- Der Fänger der Markierung `capture_status` am Binary (`SPEC-037` Nr. 10 h) — Adresse `slice-bootstrap-m6`, dessen DoD sie führt.
- Eine Aufnahme von `make gegenproben` in die Kette — [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) hält sie draußen (Laufzeit); die Aufnahme dieses Slice ändert das nicht.
- Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills — nur als Diff-Vorschlag in der Closure; `close-welle.md` trägt den Pflichtlauf bis zu diesem Slice, sein Wegfall ist ein Diff-Vorschlag an den User.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] `GATE_CHECKS += test-integration test-determinism` (jede Zeile in der Datei ihres Ziels, `harness/mk/test-integration.mk` und `harness/mk/test-determinism.mk`), die Kopf-Kommentare der beiden Dateien nennen nicht mehr „KEIN GATE“, die `make help`-Zeilen auch nicht; `make gates` fährt beide und ist grün.
- [ ] Die Läufe zugleich in der Kette sind nach dem Vertrag entschieden und belegt: beide Treiber bauen die Stufe `integration` des `Dockerfile`; die Checks der Kette laufen parallel (`Makefile`), der Architect entscheidet, ob die Läufe sich bei gleichem Image-Tag überschreiben können, und der Implementer weist die Lösung an der Kette nach (Auftrag 2).
- [ ] Zeitbudget der Kette gemessen und im Vertrag genannt (Auftrag 3); `harness/README.md` führt beide Ziele in der Tabelle der Gates mit Bindung und nennt in *Nicht behauptet* nur noch, was weiter ungeprüft ist.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `SPEC-037`, `SPEC-038`, `harness/README.md`, `.claude/commands/close-welle.md` (Pflichtlauf) als Diff-Vorschlag.
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
| `harness/mk/test-integration.mk`, `harness/mk/test-determinism.mk` | update | je `GATE_CHECKS += …`, Kopf und Hilfetext |
| `Dockerfile` | update (falls der Architect es verlangt) | Stufe `integration` und die Tags der Images, wenn zwei Checks zugleich bauen |
| `tools/harness/test-integration.sh`, `tools/harness/test-determinism.sh` | update (falls nötig) | Image-Tag je Lauf, Aufräumen, Umgang mit `--memory` in der Kette |
| `harness/vertraege/SPEC-037-test-integration.md`, `harness/vertraege/SPEC-038-test-determinism.md` | update (Architect) | Bindung, Zeitbudget, *Ungeprüft* der Verdrahtung |
| `harness/README.md` | update | zwei Zeilen in die Gate-Tabelle, *Nicht behauptet*; Zellen höchstens 200 Zeichen |
| `tools/gegenproben/`, `harness/mk/gegenproben.mk`, `tools/harness/gegenproben-eingabe.sh` | update (falls der Architect Fälle wählt) | Verdrahtung der Aufnahme; Stempel nach dem letzten Eingabe-Commit |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Aufträge 1 bis 5, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); schwere Läufe (`make gates`, `make test-integration`, `make test-determinism`) unter `flock /tmp/odata-recorder-heavy.lock`; jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; Commit-Messages per `git commit -F <datei>`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Aufträge an den Architect:**

1. **Bindung.** Ob die Aufnahme in die Kette eine ADR braucht (Linie C: ADR-Nummern 0030 bis 0039; [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) betrifft die Gegenproben, nicht diese Ziele) oder `SPEC-037` und `SPEC-038` genügen; Wortlaut der Bindung für die Spalte *Bindung* der Gate-Tabelle.
2. **Gleichzeitige Läufe in der Kette.** Beide Treiber bauen die Stufe `integration`; die Checks laufen parallel. Image-Tags, Aufräumen der Container, Verhalten bei Abbruch; der Fänger, der zwei Läufe zugleich trennt, oder das benannte Negativ mit Grund und Folge (`AGENTS.md` §3.10 für Nebenläufigkeit: Wirkungs-Fänger, nicht Datenrennen).
3. **Zeitbudget und Re-Evaluierungs-Trigger.** Zuwachs von `make gates` durch beide Ziele, gemessen im Hauptbaum unter ruhiger Last (Messung der Closure von `welle-record-replay` als Ausgangswert); der Trigger „Zuwachs über 120 s“ ist der Vorschlag des Planner, der Architect setzt die Zahl mit Grund und nennt die Folge, wenn er eintritt (Gate wieder neben die Kette, Pflichtlauf mit Träger, oder Beschleunigung).
4. **Verdrahtung und Gegenproben.** Ob ein Fall der Gegenproben die Aufnahme fängt (Kette ruft beide Ziele) oder `SPEC-037` Nr. 10 f und `SPEC-038` Nr. 10 f als akzeptiertes Negativ mit Grund und Folge bleiben; die Gegenprobe darf das Produkt nicht mitfahren (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
5. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): Docker nicht erreichbar, Image fehlt, Stufe scheitert beim Bau, Frist von 120 s überschritten, Lauf unter dem Speicherlimit, Abbruch durch INT oder TERM mitten in der Kette; die Tabelle *Zusage · Fänger* (§3), die erwartete Produktcode-Größe für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (Closure-Notiz `welle-record-replay-results`); kein anderer Slice, der `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` oder den Stempel ändert, liegt in `in-progress/` (Übergangsregeln Nr. 7). `slice-mutationslauf-werkzeug` liegt in `in-progress/`; sein Plan nimmt die Gate-Kette nicht an (Ausdrücklich NICHT, Auftrag 7 an den Architect): meldet er im Lauf, dass er sie doch ändert, wartet dieser Slice auf dessen `done/`. `slice-guard-schreibwege` berührt die Kette nur, wenn eine Gegenprobe gewählt wird; beide laufen nie zugleich.

**Größenmarke:** Prosa, kein Block. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect. Erwartet ist ein kleiner Slice (zwei Zeilen der Kette, Skripte und Verträge).

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): die Läufe lassen sich in der Kette nicht trennen (Auftrag 2), oder das Zeitbudget liegt über dem Trigger (Auftrag 3) und der User entscheidet neu.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum mit beiden Zielen in der Kette, Closure-Notiz in §7 mit Lerneintrag (Zeitbudget mit Messung; der Pflichtlauf in `close-welle.md` entfällt als Diff-Vorschlag).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die Kette wird durch zwei Container-Läufe merklich langsamer, und die Fristen der Gegenproben und von `pruefung` (für einen ruhigen Host bemessen) reißen unter Last — **Ausgang:** weiter offen: Zeitbudget und Trigger entscheidet der Architect (Auftrag 3).
- Zwei Checks bauen dieselbe Stufe `integration` zugleich und überschreiben ein Image-Tag — **Ausgang:** weiter offen: der Architect entscheidet (Auftrag 2), der Fänger oder das Negativ steht vor dem ersten Code-Commit.
- Die Aufnahme verdeckt einen Rest, den die Tests nicht prüfen (`SPEC-037` Nr. 10, `SPEC-038` Nr. 10): „Gate grün“ wird als „Binary fehlerfrei“ gelesen — **Ausgang:** weiter offen: *Nicht behauptet* in `harness/README.md` nennt den Rest (DoD).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `welle-record-replay` (Entscheidung des Users vom 2026-11-06); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger` (2×, Ausgang `geplant` in diesem Slice); `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (4×, Ausgang `geplant` in `slice-coverage-gate`) berührt Auftrag 4; `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (4×, verkörpert) betrifft den Stempel, falls Gegenproben berührt werden.

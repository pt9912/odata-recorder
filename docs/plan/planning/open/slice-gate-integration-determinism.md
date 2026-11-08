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
3. **Zeitbudget und Re-Evaluierungs-Trigger.** Zuwachs von `make gates` durch beide Ziele, gemessen im Hauptbaum unter ruhiger Last (Messung der Closure von `welle-record-replay` als Ausgangswert: `make test-integration` 5,9 s und `make test-determinism` 2,1 s bei vorhandenen Images, `pruefung test` 92,6 s und `pruefung race` 231,5 s zum Vergleich, Last des Hosts 5,5 bis 5,8; ein kalter Bau der Stufe `integration` ist nicht gemessen); der Trigger „Zuwachs über 120 s“ ist der Vorschlag des Planner, der Architect setzt die Zahl mit Grund und nennt die Folge, wenn er eintritt (Gate wieder neben die Kette, Pflichtlauf mit Träger, oder Beschleunigung).
4. **Verdrahtung und Gegenproben.** Ob ein Fall der Gegenproben die Aufnahme fängt (Kette ruft beide Ziele) oder `SPEC-037` Nr. 10 f und `SPEC-038` Nr. 10 f als akzeptiertes Negativ mit Grund und Folge bleiben; die Gegenprobe darf das Produkt nicht mitfahren (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
5. **Randformen** vor dem ersten Code-Commit im Vertrag (Raster in §6): Docker nicht erreichbar, Image fehlt, Stufe scheitert beim Bau, Frist von 120 s überschritten, Lauf unter dem Speicherlimit, Abbruch durch INT oder TERM mitten in der Kette; die Tabelle *Zusage · Fänger* (§3), die erwartete Produktcode-Größe für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (Closure-Notiz `welle-record-replay-results`); kein anderer Slice, der `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` oder den Stempel ändert, liegt in `in-progress/` (Übergangsregeln Nr. 7). `slice-mutationslauf-werkzeug` liegt in `in-progress/`; sein Plan nimmt die Gate-Kette nicht an (Ausdrücklich NICHT, Auftrag 7 an den Architect): meldet er im Lauf, dass er sie doch ändert, wartet dieser Slice auf dessen `done/`. `slice-guard-schreibwege` berührt die Kette nur, wenn eine Gegenprobe gewählt wird; beide laufen nie zugleich.

**Größenmarke:** Prosa, kein Block. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect. Erwartet ist ein kleiner Slice (zwei Zeilen der Kette, Skripte und Verträge).

**Vorläufige Marke** (Planner, 2026-11-08, `SPEC-128` Nr. 1; die Tabelle *Zusage · Fänger* fehlt, Schätzung aus dem Wortlaut von §3 Aufträge 2 bis 5): rund 12 Zeilen der Tabelle, davon 2 der Klasse `nebenläufig` (zwei Läufe zugleich bei gleichem Image-Tag, Abbruch mitten in der Kette), *a* = 2, *b* = 10, *f* = 0, *p* = 60 (Mk-Zeilen, Skripte, Vertragstext): *S* = 85·2 + 80·10 + 200·0 + 60 = 1030, *M* = 100·⌈135·1030/10000⌉ = 1400. **Schnittfrage: nein** (*S* < 2223, Schnittgrenze); es wird vor dem Code nicht geschnitten. Der Planner rechnet neu, sobald der Architect die Tabelle hat, und setzt die endgültige Marke vor dem ersten Code-Commit. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Plans (§3), Tests eingeschlossen. Basis der Messung: `4a6a760b` (die Start-Commits ändern allein Pläne).

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

- Die Kette wird durch zwei Container-Läufe merklich langsamer, und die Fristen der Gegenproben und von `pruefung` (für einen ruhigen Host bemessen) reißen unter Last — **Ausgang:** weiter offen: Zeitbudget und Trigger entscheidet der Architect (Auftrag 3).
- Zwei Checks bauen dieselbe Stufe `integration` zugleich und überschreiben ein Image-Tag — **Ausgang:** weiter offen: der Architect entscheidet (Auftrag 2), der Fänger oder das Negativ steht vor dem ersten Code-Commit.
- Größe: die Marke steht als Prosa in §4 (vorläufig *S* = 1030, *M* = 1400, keine Schnittfrage) — **Ausgang:** offen; Adresse: Planner (neu rechnen mit der Tabelle des Architect, endgültige Marke vor dem ersten Code-Commit).
- Berührung mit `slice-bootstrap-m3` und `slice-bootstrap-m3-e2e-mock` (dieser ändert den Treiber in Stufe 8 und 9, siehe *Zurückgestellt* in §4) am Treiber `tools/harness/test-integration.sh` und an der Stufe `integration` des `Dockerfile` (Stand der Bedingungen in §4) — **Ausgang:** offen; Adresse: Architect (Auftrag 2 hält die Änderung klein) und Planner (Meldung des Implementers vor einem Commit am `Dockerfile`).
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

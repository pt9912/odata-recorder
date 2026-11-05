# Slice slice-test-integration: `make test-integration` — das gebaute Binary im Container, Record → Upstream stoppen → Replay

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-RB-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-037` (Vertrag `make test-integration`, `harness/vertraege/SPEC-037-test-integration.md`: *Record → Upstream stoppen → Replay; Mock CRUD → Reset*, Referenz-Upstream im Container) · `LH-RB-03.a` (Nachweis: reproduzierbarer Container-Build und End-to-End-Test) · `SPEC-049` (Fixture-Rahmen, Header- und Fehlerzeilen) · `SPEC-045` (Capture-/Replay-Fähigkeitsmatrix) · `ARC-012` · `architecture.md` §6 (*End-to-End*: Test-OData-Server starten → Requests aufzeichnen → Server stoppen → Replay).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make test-integration` (`SPEC-037`) startet das gebaute Binary im Container gegen einen Referenzdienst, schickt echte HTTP-Requests im Modus `record`, beendet den Referenzdienst und spielt dieselben Requests im Modus `replay` ab. Die Szenarien sind die Akzeptanzkriterien des Lastenhefts (Happy Path, Boundary, Negative) von `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05` und `LH-FA-14`; welche Zeile welches Kriterium trägt, steht in der [Rückverfolgungs-Matrix](../rueckverfolgung.md) (Teil A, Spalte *E2E-Szenario*). Es ist der erste Lauf, der die Prozessgrenze überschreitet: Binary, Startkonfiguration aus der Umgebung, Exit-Code und Signal. Den ganzen Pfad im Prozess von `go test` fährt zuerst der M2-Test aus `slice-bootstrap-m2` (`make test`); dieser Slice prüft, was dort nicht gilt, weil das Binary nicht läuft.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Teil *Mock CRUD → Reset* von `SPEC-037` — es gibt vor `slice-bootstrap-m3` keinen Mock-Code; das Szenario ist DoD-Unterpunkt von `slice-bootstrap-m3` (dort eingetragen, `AGENTS.md` §3.13). Dasselbe gilt für das Szenario des Verbindungszählers im Modus `hybrid` (`slice-bootstrap-m4`) und für die Szenarien über den Admin-Listener (`slice-bootstrap-m6`).
- Fehler-Injektion im Speicher (Negative von `LH-FA-02` und `LH-FA-14`) — am laufenden Binary ohne Eingriff in den Speicher nicht auslösbar; die Form entscheidet der Architect (§6). Bis dahin trägt `slice-test-security` die Negative von `LH-FA-14`, die von `LH-FA-02` hat in der Matrix kein E2E-Szenario.
- Vergleichstest Direktzugriff gegen Proxy (`LH-QA-07`) — `slice-protokolltreue`; der Aufbau (Binary und Referenzdienst) wird dort wiederverwendet, falls dieser Slice vorher in `done/` liegt (Reihenfolge, kein Blocker).
- Neustart des Volumens im Image (`LH-QA-06`) — `slice-container-image`; 100 Wiederholungen — `slice-test-determinism`; Canary-Secrets in allen Artefakten — `slice-test-security`; Last — `slice-lasttest`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Das Ziel `make test-integration` und der Vertrag `SPEC-037` (der Stand „entsteht mit dem Slice, der den Code liefert“ wird zum gelieferten Umfang): Binary und Referenzdienst laufen als Container ohne Zugang zu einem Außennetz; Einbindung in die Gate-Kette oder daneben nach der Entscheidung zu §6.
- [ ] Die Szenarien Record → Upstream stoppen → Replay für die Kriterien, die die Matrix diesem Slice zuweist: `LH-FA-01` (Happy Path, Boundary, Negative), `LH-FA-02` (Happy Path, Boundary), `LH-FA-03` (alle drei, darunter der Neustart des Prozesses auf demselben Datenverzeichnis), `LH-FA-04`, `LH-FA-05` (je alle drei) und `LH-FA-14` (Happy Path, Boundary); jede Stufe nennt Anforderung und Kriterium (Form der Deklaration und Bezug zu `make e2e-abdeckung`: Architect).
- [ ] Je Szenario eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10; Beispiele: Replay kontaktiert den Referenzdienst, ein Header mit Geheimnis erreicht die Datei, die Reihenfolge wiederholter Requests kippt).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-037`, Sensors-Tabelle in `harness/README.md`).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `Makefile`, `harness/mk/` (neues Fragment) | update / neu | Ziel `make test-integration`; Bindung in der Sensors-Tabelle |
| `Dockerfile` | update | Stufe für das Binary im Laufzeit-Image und für den Referenzdienst (Architect; Abgleich mit `slice-container-image`, §6) |
| `tests/` | neu | Szenarien und Aufbau (Pfad: Architect) |
| `harness/vertraege/SPEC-037-test-integration.md` | update | Stand und Umfang |
| `harness/README.md` | update | Sensors-Tabelle: Ziel, Vertrag, Bindung |
| `tools/gegenproben/` | update | Fälle zur Verdrahtung, falls das Ziel an `GATE_CHECKS` hängt (heute ohne Fall: `harness/README.md` *Nicht behauptet*) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag unter `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-cli-adapter` liegen in `done/` (das Binary startet, hört auf Signale und bedient `record` und `replay`). Der Vertrag `SPEC-037` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile`, `harness/mk/` und `Dockerfile` und gehört damit zur Gate-Kette: Er startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7). Reihenfolge: vor `slice-protokolltreue`, damit der Aufbau (Binary und Referenzdienst) dort wiederverwendet werden kann; ein Blocker ist es nicht.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): Die Marke setzt der Planner beim Start nach `AGENTS.md` §3.14 (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Zahl steht erst mit der Tabelle *Zusage · Fänger* des Architect, die Zeilen der Szenarien (16 Kriterien der Matrix) und der Aufbau (zwei Container) gehören hinein.
- `in-progress` → `open` (blockiert): Binary und Referenzdienst lassen sich ohne Außennetz nicht verbinden (siehe §6).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Gate-Kette oder daneben: ein Ziel in der Kette verlängert jeden Lauf von `make gates`, eines daneben verrottet, wenn niemand es ruft (Vorbild der Wahl: `make gegenproben`, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); eine Fallzeit, die mit dem Produkt wächst, hat schon einmal Gegenproben gerissen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, 2×). Empfehlung des Planners: zuerst daneben, mit Pflichtlauf vor jeder Welle-Closure und bei Änderungen an `src/bootstrap` und `src/adapters/driving`, und Aufnahme in die Kette, sobald die gemessene Laufzeit unter einem Zeitbudget bleibt (neue Gates sind erlaubt, `AGENTS.md` §3.6) — **Ausgang:** offen; Adresse User, Frist vor dem Start des Slice.
- Netzlos mit zwei Containern: `AGENTS.md` §3.1 verlangt `--network none`, Binary und Referenzdienst müssen sich aber erreichen (internes Docker-Netz, gemeinsamer Netz-Namespace oder beide im selben Container) — **Ausgang:** offen; Adresse Architect, entscheidet vor dem Code.
- Überlappung mit dem M2-Test (derselbe Rundlauf im Prozess von `go test`, unter `-race`) und mit `slice-container-image` (nennt `SPEC-037` und führt einen Neustarttest gegen das Image): der Architect grenzt vor dem Code ab, was welcher Test beweist; soll `slice-container-image` ein Ziel von hier übernehmen, trägt der Planner es dort ein (`AGENTS.md` §3.13) — **Ausgang:** offen; Adresse Architect.
- Fehler-Injektion (Negative von `LH-FA-02` und `LH-FA-14`) am laufenden Binary — **Ausgang:** offen; Adresse Architect; solange ohne Entscheidung hat `LH-FA-02` Negative in der Matrix kein E2E-Szenario.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in `SPEC-037` oder „gilt nicht“ mit Grund — **Ausgang:** offen.
- Die Gate-Kette hat einen Besitzer zugleich (Übergangsregeln Nr. 7) — **Ausgang:** offen; der Start prüft, dass kein Slice der Linie C in `in-progress/` liegt.

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

*Wird bei der Closure gefüllt.*

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

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (11×, verkörpert in `AGENTS.md` §3.15; die Negative von `LH-FA-02` und `LH-FA-14` ohne E2E-Szenario sind solche Negative und brauchen Grund und Folge) und `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (2×, offen: ein Fall, der das Produkt mitfährt, wächst mit ihm; gilt für die Einbindung des Ziels, §6); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

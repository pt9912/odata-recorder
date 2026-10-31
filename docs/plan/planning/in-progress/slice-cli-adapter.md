# Slice slice-cli-adapter: CLI-Driving-Adapter — Kommandoauswahl, Signale und Exit-Codes

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-02`, `LH-RB-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md)

**Berührte Spec-Stellen:** `ARC-006` (CLI Driving Adapter: Konfiguration, Kommandoauswahl, Start-/Exit-Verhalten) · `ARC-012` (Composition Root) · `LH-FA-02.c` (Beenden des Prozesses: SIGTERM/SIGINT) · `LH-RB-03.a` (das Binary ist auch nativ unter Linux startbar) · `SPEC-110` (Startkonfiguration: es gibt keine Kommandozeilen-Optionen, die Konfiguration kommt aus der Umgebung) · `SPEC-059` (`SHUTDOWN_TIMEOUT`) · `architecture.md` §5 (Zeile *Prozess-Signal*).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-cli-adapter)

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Driving Adapter `src/adapters/driving/cli` (`ARC-006`) wählt das Kommando aus den Argumenten (heute `version` und der Start; der Bestand liegt in `bootstrap.Run`), übersetzt SIGTERM und SIGINT in den Abbruch eines Kontexts, mit dem der Composition Root den geordneten Shutdown nach `LH-FA-02.c` auslöst, und bildet das Ergebnis auf den Exit-Code des Prozesses ab; `src/bootstrap/main` ruft den Adapter. Der Prozess hat damit seinen Eingang, bevor `slice-bootstrap-m2` Listener und Dienste dahinter hängt.

**Übernimmt:** aus `slice-bootstrap-m2` die Zeile `src/bootstrap/main/` *Kommando zum Start* (Kommandoauswahl, Exit-Codes; Sendung von `slice-fehlermodell-konfiguration`, `AGENTS.md` §3.13); `slice-bootstrap-m2` führt den Punkt in §1 *Übernimmt* und im Start-Trigger nach diesem Slice (im selben Commit wie dieser Plan nachgezogen). Das Laden der Konfiguration (`LadeOderMelde`) bleibt dort.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Kommandos über Start und Version hinaus — entfallen (Entscheidung des Users, 2026-10-30, [welle-record-replay](../welle-record-replay.md) §6): `LH-FA-13` verlangt eine Verwaltungsschnittstelle, keine Kommandos, und die Spezifikation nennt keine Kommandozeilen-Optionen (`SPEC-110`).
- Listener öffnen und schließen, 503 während des Herunterfahrens — `slice-http-adapter` (zwei getrennte Schritte); dieser Slice liefert nur den Auslöser.
- Verdrahtung von Konfiguration, Diensten und `CloseAll` — `slice-bootstrap-m2`; der Adapter kennt weder Dienste noch Listener (`architecture.md` §2: Driving Adapter importieren nur Modell und Ports).
- Der Admin-Listener — `slice-admin-listener-token` ([welle-admin-api](../welle-admin-api.md)).

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

- [ ] Kommandoauswahl und Exit-Codes im Adapter `src/adapters/driving/cli`: das Bestandsverhalten von `bootstrap.Run` (ohne Argument und mit `version` schreibt es `odata-recorder <version>` und liefert 0, jedes andere Aufrufmuster liefert 2) bleibt, bis der Architect es in der Spezifikation ändert; `src/bootstrap/main` ruft den Adapter.
- [ ] Signale: SIGTERM und SIGINT brechen den Kontext des Starts ab (`LH-FA-02.c`); die Randformen (Signal vor dem Ende des Starts, zweites Signal, Signal nach einem Konfigurationsfehler) stehen als Raster in §6 und sind vor dem Code in der Spezifikation entschieden (`AGENTS.md` §3.12, §3.16).
- [ ] `make gates` grün, darunter `make a-check` (der Adapter importiert nur Modell und Ports) und `make test-race` (die Signalbehandlung ist nebenläufig, `AGENTS.md` §3.10).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (Aufrufmuster, Exit-Codes, Signalverhalten).
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
| `src/adapters/driving/cli/` | neu | Kommandoauswahl, Signal in Kontextabbruch, Exit-Code (`ARC-006`) |
| `src/bootstrap/main/` | update | ruft den Adapter statt `bootstrap.Run` |
| `src/bootstrap/bootstrap.go` | update | `Run` (Bestand: `version`, Exit-Codes) geht im Adapter auf; die Form entscheidet der Architect (§6) |
| `spec/spezifikation.md` | update | Vertrag zu Aufrufmustern, Exit-Codes und Signalverhalten (Architect, vor dem ersten Code-Commit) |
| Testdateien neben dem Adapter | neu | Happy Path, Boundary und Negative je Aufrufmuster; Signal-Tests ordnen mit Kanälen (`AGENTS.md` §3.10) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fehlermodell-konfiguration` liegt in `done/` (erfüllt: `LadeOderMelde`). Die Spezifikation trägt den Vertrag (Aufrufmuster, Exit-Codes, Signalverhalten) samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Reihenfolge: vor `slice-bootstrap-m2`, dessen Start-Trigger diesen Slice nennt; beide ändern `src/bootstrap/main/` und `src/bootstrap/bootstrap.go` und laufen deshalb nacheinander (Linie A, [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 3).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): Die Marke setzt der Planner beim Start nach `AGENTS.md` §3.14 (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht erst mit der Tabelle *Zusage · Fänger* des Architect; der Planner schätzt den Slice heute als klein, ohne Zeilenzahl.
- `in-progress` → `open` (blockiert): der Adapter erreicht den Start nicht, ohne `src/bootstrap` zu importieren (siehe §6).

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

- Wie erreicht der Adapter den Start, ohne `src/bootstrap` zu importieren? `a-check` erlaubt einem Driving Adapter nur Modell, Inbound Ports und Querschnitts-Ports; der Composition Root importiert alle. — **Ausgang:** offen; Adresse Architect, entscheidet vor dem Code (`AGENTS.md` §3.13).
- Randformen des Signals (Signal vor dem Ende des Starts, zweites Signal, Signal nach einem Konfigurationsfehler, Signal während `CloseAll`) und der Aufrufmuster (leeres Argument, Flag wie `--help`, `version` mit Zusatzargument; Bestand von `bootstrap.Run`): das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit, jede Zelle mit Ort in der Spezifikation oder „gilt nicht“ mit Grund (`AGENTS.md` §3.12, §3.16) — **Ausgang:** offen; Adresse Architect.
- Die Spezifikation hat zu Aufrufmustern und Exit-Codes keinen Vertrag; `SPEC-110` sagt nur „keine Kommandozeilen-Optionen“ und `LH-FA-02.c` das Verhalten bei SIGTERM/SIGINT (benannte Spec-Lücke) — **Ausgang:** offen; Adresse Architect, schreibt den Vertrag vor dem Code.
- Exit-Codes und Signalverhalten sind am Binary sichtbar; das Handbuch-Kapitel dazu käme aus `slice-handbuch-record-replay`, dessen Plan den Punkt nicht führt (`AGENTS.md` §3.13) — **Ausgang:** offen; Adresse Planner, trägt ihn beim Start von `slice-handbuch-record-replay` ein.
- Dateikollision mit `slice-bootstrap-m2` in `src/bootstrap/main/` und `src/bootstrap/bootstrap.go` — **Ausgang:** offen; durch die Reihenfolge im Start-Trigger gelöst, wenn beide nacheinander laufen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9) und, weil die Signalbehandlung nebenläufig ist, `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (8×, verkörpert in `AGENTS.md` §3.10); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

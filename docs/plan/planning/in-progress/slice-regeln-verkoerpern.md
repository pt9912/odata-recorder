# Slice slice-regeln-verkoerpern: Die Pflicht der Folgewellen und zwei Vorlagen der Closure von welle-mock in die Guide-Dateien

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): er ändert `AGENTS.md`, `.harness/skills/`, `.claude/commands/`, `.claude/agents/` und `.d-check.yml`, nicht die Gate-Kette; er ist trotzdem ein Kettenslice (Nr. 7), weil die Größengrenzen der Guide-Dateien ein Gate sind (`make docs-check`) und seine Verschiebung eine Änderung der Gate-Konfiguration ist.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Regeln der Rollen, der Bezug ist der nächste, den es gibt), [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Größe der Guide-Dateien als Gate; Anheben ist eine Lockerung, `AGENTS.md` §3.6), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Verträge der Werkzeuge, `SPEC-201`), Register-Einträge `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×) und `BEO-REPO/report-der-rolle-faellt-docs-check` (5×), beide über der Schwelle und in der Closure von `welle-mock` als Vorlage ohne `liegt in` geführt.

**Berührte Spec-Stellen:** `SPEC-201` Nr. 11 (Folge für Folgewellen: Pflicht, die dieser Slice in die Guide-Dateien trägt; Nachzug des Architect, falls der Wortlaut dort die Verkörperung nennen soll) · `SPEC-040` Nr. 8 b (Verschiebung der Größengrenzen, Messung beider Dateien) · `SPEC-040` Nr. 9 (Abschnitte von `AGENTS.md`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** pt9912 (Zweig slice-regeln-verkoerpern)

**Autor:** pt9912 (Planner der Closure von `welle-mock`). **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Drei Regeln, die die Closure von `welle-mock` als Vorlage ohne Träger hinterließ, stehen dort, wo jeder Lauf sie liest, und kein Gate wird gelockert: (1) die Pflicht der Folgewellen aus `SPEC-201` Nr. 11 (wer einen Fehlercode, einen Schlüssel oder ein Kommando erreichbar macht, ändert Ausnahmen und Handbuch im selben Slice) in `AGENTS.md` oder `.claude/commands/implement-slice.md`; (2) Vorlage 1 der Closure (Zweig ohne Lauf unter dem Detektor) im Reviewer-Skill `.harness/skills/reviewer.md`; (3) Vorlage 2 (Report der Rolle besteht `make docs-check`) im Reviewer-Skill (Output-Schema) und in `.claude/agents/verifier.md`. Jeder Diff, der eine Guide-Datei verlängert, kürzt im selben Diff eine Datei desselben Maßes um mindestens denselben Betrag oder verschiebt die Grenze nach `SPEC-040` Nr. 8 b; die Summe der Grenzen steigt nicht.

**Übernimmt:** die Sendung aus der Closure von `welle-mock` (Geber `welle-mock`, Abschnitt *Steering-Loop-Einträge*: Vorlage 1, Vorlage 2 und die Pflicht aus `SPEC-201` Nr. 11, Entscheidung des Users vom 2026-11-08, Punkte lösen).

**Wer legt was fest.** Der Architect legt vor dem Code fest: den Ort und den Wortlaut jeder der drei Regeln, die Kürzungen (welche Zeilen oder Bytes in `reviewer.md`, `implement-slice.md`, `AGENTS.md` und `verifier.md` entfallen oder in welchen Vertrag sie ziehen) und die Messung der Größen vor und nach dem Diff. Der Planner entscheidet nichts davon; die Texte der Vorlagen sind Eingabe, nicht Beschluss.

**Ausgangsmaße** (Messung der Closure von `welle-mock`, Grenzen in `.d-check.yml`; sie stehen auf ihren Grenzen, also verlängert jede Zeile): `.harness/skills/reviewer.md` 145 von 145 Zeilen und 8110 von 8110 Bytes; `.claude/commands/implement-slice.md` 17081 von 17081 Bytes; `AGENTS.md` 326 von 326 Zeilen und 23633 von 23633 Bytes (nach der Closure von `welle-etag`: §3.17 trägt einen Satz mehr, §1 Duplikate zu `harness/conventions.md` entfielen, die Grenzen stehen auf dem neuen Ist); `.claude/agents/architect.md` 3759 von 3759 Bytes; `.claude/agents/verifier.md` Platz bis 53 Zeilen und 3759 Bytes. Der Implementer misst vor dem ersten Edit neu.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Anhebung einer Grenze ohne gleichzeitige Senkung (Nr. 8 b von `SPEC-040`) — das wäre eine Lockerung des Gates und braucht eine ADR (`AGENTS.md` §3.6); wer die Verkörperung nicht mit Kürzung tragen kann, gibt den Punkt an den Planner zurück, statt die Grenze zu heben.
- Ein Sensor für die Pflicht aus `SPEC-201` Nr. 11 (Erreichbarkeit eines Codes gegen die Ausnahmen, `SPEC-201` Nr. 9 c) — ein Fänger am Binary ist ein anderer Vorgang; die Verkörperung ist die Gegenmaßnahme, die `SPEC-201` selbst nennt.
- Die Verkörperung weiterer Beobachtungen über der Schwelle (29 stehen in der Closure von `welle-mock`; 27 sind verkörpert) — der Auftrag nennt drei Regeln.
- Eine Änderung der Werkzeugliste der Rollen-Typen (`Edit` für den Architect) — Entscheidung des Users (Rückgabe 2 in §6 von [slice-guard-schreibwege](../done/slice-guard-schreibwege.md), Diff-Vorschlag in dessen §7; der Slice liegt in `done/` und ist keine Adresse mehr).
- Die Größenformel und der Zuschlag (`SPEC-128` Nr. 1) — Gegenstand von `slice-groessenformel-neu`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] Die drei Regeln stehen am Ort und im Wortlaut, den der Architect festlegt, jede mit dem Anker `seit welle-mock` (`grundlagen-traceability.md` §Herkunfts-Anker) und, soweit sie einen Zielort hat, dem Feld `liegt in` in der Notiz des Slice (§7); die Regel aus `SPEC-201` Nr. 11 trägt zusätzlich die Zeiger auf ihren Vertrag.
- [ ] Die Größengrenzen sind gehalten: jede berührte Guide-Datei liegt auf oder unter ihrer Grenze, keine Grenze ist ohne Senkung einer anderen im selben Maß gehoben (Messung vor und nach dem Diff in §7), `make docs-check` ist grün.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `SPEC-201` Nr. 11 (Architect), das Register: die Einträge `zweig-ohne-lauf-unter-dem-detektor` und `report-der-rolle-faellt-docs-check` bekommen den Ausgang *verkörpert* mit Zielort.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag (Einträge mit `liegt in <Zielort>`).
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.harness/skills/reviewer.md` | update | Vorlage 1 (MEDIUM-Liste) und Vorlage 2 (Output-Schema); steht auf 145/145 Zeilen und 8110/8110 Bytes, also nur mit Kürzung |
| `.claude/agents/verifier.md` | update | Vorlage 2 (Report besteht `make docs-check`); Platz bis 53 Zeilen und 3759 Bytes |
| `AGENTS.md` oder `.claude/commands/implement-slice.md` | update | Pflicht aus `SPEC-201` Nr. 11; beide stehen auf der Grenze; Abschnitt `### 3.<n>` höchstens 29 Zeilen (`SPEC-040` Nr. 9) |
| `.d-check.yml` | update (nur Verschiebung) | Grenze der verlängerten Datei hoch, Grenze einer gekürzten Datei um mindestens denselben Betrag herunter (`SPEC-040` Nr. 8 b) |
| `harness/vertraege/SPEC-201-*.md` | update (Architect) | Nr. 11 nennt den Ort der Verkörperung |
| `docs/plan/planning/observations/BEO-REPO/zweig-ohne-lauf-unter-dem-detektor/`, `report-der-rolle-faellt-docs-check/` | update | `state.md`: Ausgang *verkörpert* mit Zielort |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Texte, Orte, Kürzungen, Messung, Tabelle *Zusage · Fänger*, Raster) → Planner (Marke, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`; jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Aufträge an den Architect:**

1. **Wortlaut und Ort der drei Regeln** (Vorlagen 1 und 2 der Closure von `welle-mock` sind Eingabe, die Pflicht aus `SPEC-201` Nr. 11 ist Vertragstext): Ob die Pflicht in `AGENTS.md` (neue Hard Rule oder Zusatz zu einer vorhandenen, höchstens 29 Zeilen je Abschnitt) oder in `implement-slice.md` steht; wer sie prüft (Reviewer-Skill, Zeile der HIGH- oder MEDIUM-Liste).
2. **Kürzungen.** Welche Zeilen und Bytes in welcher Datei entfallen, damit jede berührte Datei auf oder unter ihrer Grenze bleibt, oder welche Verschiebung nach `SPEC-040` Nr. 8 b gewählt wird (Summe der Grenzen je Maß steigt nicht). Eine gekürzte Regel darf nichts zusagen, was kein Fänger mehr trägt (`AGENTS.md` §3.11); der Architect nennt je Kürzung die Regel, deren Träger entfällt, oder zeigt, dass sie doppelt stand.
3. **Messung.** Zeilen und Bytes jeder berührten Datei vor und nach dem Diff, Grenze aus `.d-check.yml`.
4. **Randformen** vor dem ersten Code-Commit (Raster in §6): eine Regel, die in zwei Dateien zugleich steht (Drift, `AGENTS.md` §1), eine Kürzung, die einen Anker `seit welle-<Kennung>` löscht (Retirement-Check, `grundlagen-traceability.md`), die Verschiebung einer Grenze in einem Commit mit den Kürzungen.
5. **Tabelle *Zusage · Fänger*** (§3), erwartete Größe für die Marke (kein Go-Code; ob die Messung zählt, entscheidet `SPEC-128`).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect): folgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Closure-Notiz `welle-mock-results`); kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln); Reihenfolge der Kette seit 2026-11-08: `slice-gate-integration-determinism`, `slice-guard-schreibwege`, dieser Slice, `slice-groessenformel-neu`. Es laufen höchstens drei Slice-Worktrees zugleich. Der Slice braucht `slice-guard-schreibwege` nicht im Code; er ändert aber dieselben Rollen-Dateien nicht, sodass die Reihenfolge der Kette eine Wahl des Planners ist, kein Blocker. Der User kann ihn vorziehen.

**Start belegt (2026-11-19, Planner):** `welle-mock-results` liegt in `done/`; `in-progress/` enthielt nur `roadmap.md`, kein anderer Slice der Kette lief (`slice-gate-integration-determinism`, `slice-guard-schreibwege` und `slice-gates-ohne-worktrees` liegen in `done/`); `git worktree list` nannte vor dem Start nur den Hauptbaum, der Worktree dieses Slice ist der erste (Verzeichnis odata-recorder-wt-regeln neben dem Hauptbaum, Zweig `slice-regeln-verkoerpern`, Basis `main`). Die Ausgangsmaße aus §1 sind am Start nachgemessen und gleich: `reviewer.md` 145 Zeilen und 8110 Bytes, `implement-slice.md` 220 Zeilen und 17081 Bytes, `AGENTS.md` 326 Zeilen und 23633 Bytes, `architect.md` 53 Zeilen und 3759 Bytes, `verifier.md` 43 Zeilen und 2639 Bytes (Grenzen und Summen in `.d-check.yml`, Kommentar zur Verschiebung vom 2026-11-06); der Implementer misst vor dem ersten Edit trotzdem neu.

**Größenmarke:** noch nicht gesetzt, Prosa, kein Block. Der Architect ist zuerst dran (Texte, Orte, Kürzungen, Messung, Tabelle *Zusage · Fänger*, erwartete Größe, §3 Aufträge 1 bis 5); danach setzt der Planner die Marke in eigenem Commit vor dem ersten Code-Commit. Die Marke setzt der Planner vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128` Nr. 1 und `AGENTS.md` §3.14; die Schätzung kommt aus der Tabelle des Architect; ob die Messung bei einem Slice ohne Go-Code zählt (Auftrag 5), entscheidet der Architect nach `SPEC-128`. Erwartet ist ein kleiner Slice (drei Regeln, Kürzungen, eine Verschiebung); die Formel wird von `slice-groessenformel-neu` geändert, die Marke gilt nach dem Stand der Formel beim Start.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): die Kürzungen lassen sich nicht ohne Verlust eines Trägers finden und der User muss über eine ADR zur Anhebung einer Grenze entscheiden.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (drei Einträge mit `liegt in <Zielort>` und Anker `seit welle-mock`, oder eine benannte Lücke, wenn eine Regel nicht ohne Verlust gekürzt werden konnte).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Eine Kürzung entfernt eine Regel, die ein Fänger oder ein Anker trägt — **Ausgang:** weiter offen: der Architect nennt je Kürzung den Träger (Auftrag 2); der Retirement-Check gilt (`grundlagen-traceability.md`).
- Die Pflicht aus `SPEC-201` Nr. 11 steht danach in zwei Dateien und driftet — **Ausgang:** weiter offen: ein Ort, ein Zeiger auf den Vertrag (Auftrag 1).
- Die Verschiebung einer Grenze wird als Lockerung gelesen — **Ausgang:** weiter offen: Messung beider Dateien in §7, der Reviewer prüft den Diff von `.d-check.yml` (`SPEC-040` Nr. 8 e).
- Raster der Randformen (Auftrag 4) ist noch leer — **Ausgang:** weiter offen: der Architect füllt es vor dem ersten Code-Commit (`AGENTS.md` §3.16).
- Größe: die Marke steht als Prosa in §4 und ist am Start nicht gesetzt (der Architect liefert zuerst die erwartete Größe) — **Ausgang:** offen; Adresse: Planner (nach dem Entwurf des Architect, vor dem ersten Code-Commit, eigener Commit).
- Der Slice-Zweig schreibt `AGENTS.md`, `.harness/skills/reviewer.md`, `.claude/agents/verifier.md` und gegebenenfalls `.claude/commands/implement-slice.md`; Nr. 4 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) behält diese Dateien dem Integrationsschritt vor und verlangt von einem Zweig, der dort etwas ändern müsste, eine Zeile in §6 mit Adresse Planner — hier ist die Änderung der Gegenstand der DoD, der Plan geht davon aus, dass der Slice-Auftrag sie erlaubt — **Ausgang:** offen; Adresse: User (Bestätigung vor dem ersten Code-Commit; ein Rollen-Lauf, dessen Rechte `.claude/` nicht schreiben lassen, gibt den Punkt zurück, statt es zu umgehen).

**Offene Aufträge an den Architect (Stand Start 2026-11-19, alle *offen*, keiner entschieden; Wortlaut je Auftrag in §3):** (1) Wortlaut und Ort der drei Regeln: Pflicht aus `SPEC-201` Nr. 11 in `AGENTS.md` oder `implement-slice.md`, wer sie prüft; Vorlage 1 und 2 aus `welle-mock-results` als Eingabe · (2) Kürzungen je Datei mit dem Träger, der entfällt (Eingabe, nicht Beschluss: alle sechs Rollen-Typen in `.claude/agents/` (`implementer`, `architect`, `validator`, `verifier`, `reviewer`, `planner`) tragen am Ende einen Kommentarblock der Vorlage mit `ANPASSEN`, gefunden mit `grep`; ob er ein Kürzungskandidat ist, liegt bei ihm) oder Verschiebung nach `SPEC-040` Nr. 8 b mit gleichbleibender Summe · (3) Messung in Zeilen und Bytes vor und nach dem Diff, Grenze aus `.d-check.yml` · (4) Raster der Randformen für §6, darunter die Regel in zwei Dateien, der gelöschte Anker `seit welle-<Kennung>`, Verschiebung und Kürzung im selben Commit · (5) Tabelle *Zusage · Fänger* und erwartete Größe für die Marke, samt der Frage, ob ein Slice ohne Go-Code unter `SPEC-128` Nr. 1 zählt · (6) nur auf Entscheidung des Users: für jede Ergänzung, die er aus der Tabelle unten wählt, die zusätzlichen Zeilen und Bytes und die Kürzung, die sie trägt.

**Mögliche Ergänzungen aus den Closures von `welle-hybrid` und `slice-guard-schreibwege`** (Planner, Start 2026-11-19; der Umfang des Slice bleibt unverändert: §1 schließt „weitere Beobachtungen über der Schwelle“ aus, die DoD nennt drei Regeln; jede Zeile ist Vorschlag, keine Aufnahme) — **Ausgang:** offen; Adresse: User (Entscheidung vor dem Auftrag an den Architect; jede Ergänzung lädt Kürzungen auf, denn die Ziele außer `verifier.md` stehen auf der Grenze, und sie änderte die DoD, die drei Regeln nennt):

| Beobachtung (Register, Zähler am Start) | Ort laut Diff-Vorschlag | Berührt dieselbe Datei wie | Vorschlag steht in |
|---|---|---|---|
| `menge-ohne-ihre-positionen` (6×, E1) | `AGENTS.md` §3.15, ein Satz vor „Eine Fixture gleicht nie dem Wert der Mutation.“; Anker erst bei der Verkörperung | die Pflicht aus `SPEC-201` Nr. 11, falls sie nach `AGENTS.md` kommt (Grenze 326 Zeilen und 23633 Bytes) | `done/welle-hybrid-results.md`, *Entscheidungen beim User* Nr. 1 |
| `verweis-auf-review-report-sperrt-die-archivierung` (3×, E2) | `implement-slice.md` Schritt 25 und die Anweisungen von Reviewer und Verifier | Vorlage 2 (`reviewer.md`, `verifier.md`) und gegebenenfalls Ort der Pflicht in `implement-slice.md` (17081 von 17081 Bytes) | dasselbe, Nr. 2 |
| `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (4×) und `laufzeitmessung-widerspricht-sich-unter-last` (3×, E4) | `AGENTS.md` §3.15 und eine Regel nach §3.10 | `AGENTS.md` | dasselbe, Nr. 4, und §7 von `done/slice-gegenproben-zeitbudget` (archiviert unter `done/welle-hybrid/`) |
| Rückgabe 2 und 3 von `slice-guard-schreibwege`: `Edit` im Typ `architect`; Satz zu den Scratchpad-Namen in sechs Rollen-Typen | `.claude/agents/*.md` (architect nimmt den Rest der Vorlage `ANPASSEN` heraus: 50 Zeilen, 3614 Bytes) | `verifier.md` (Vorlage 2) und die Kürzungen in `architect.md` | `done/slice-guard-schreibwege.md` §7, *Entscheidungen beim User* Nr. 1 und 2 |

Nicht berührt und nicht vorgeschlagen: E3 (Auftrag an den Architect zur Nullprobe), E5, E7 bis E9 (`welle-hybrid-results`), die beiden neuen Einträge unter der Schwelle von `slice-guard-schreibwege` (`zeit-zusage-an-einer-form-und-einer-locale`, `allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle`, je 1×, unter der Schwelle).

**Sendungen am Start geprüft (`AGENTS.md` §3.13, 2026-11-19):** (a) *Eingang:* die Sendung dieses Slice (Vorlage 1, Vorlage 2, Pflicht aus `SPEC-201` Nr. 11) steht in §1 *Übernimmt* und in der DoD (drei Regeln); kein Ausschluss in §1 trifft sie. Der Geber `welle-mock-results` führt die Vorlagen als „beim User offen“ ohne Nehmer (das Zeitdokument liegt in `done/` und bleibt); den Träger nennen das Drift-Log der Roadmap (2026-11-08), `welle-hybrid-results` (*Folge-Slices*) und die `state.md` der beiden Einträge, die dieser Plan am Start auf `in-progress/` nachzieht. (b) *Ausgang:* der Ausschluss „Größenformel und Zuschlag“ nennt `slice-groessenformel-neu` (`open/`, Kette, vierter); dessen §1 führt die Formel als Ziel und schließt sie nicht aus, die Adresse trägt. Der Ausschluss „Werkzeugliste der Rollen-Typen“ nannte `slice-guard-schreibwege` (`done/`, keine Adresse): im Plan auf den User umgestellt, siehe §1. Ein Sensor für `SPEC-201` Nr. 9 c (§1) hat keinen Nehmer unter den offenen Slices (geprüft: `grep` über `open/` und `next/` nach „Erreichbarkeit“ und `SPEC-201`: `slice-handbuch-mock` und `slice-handbuch-etag-zeile` nennen `SPEC-201` nur als Abgleich der Schlüssel, `slice-test-security` meint die Erreichbarkeit des Admin-Listeners; keiner nimmt einen Sensor an); er bleibt beim User. (c) *Dritte:* kein anderer Slice nennt diesen Slice als Nehmer außer der Roadmap und `welle-hybrid-results`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `welle-mock` (Entscheidung des Users vom 2026-11-08); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×) und `BEO-REPO/report-der-rolle-faellt-docs-check` (5×) sind der Gegenstand; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` und `BEO-REPO/kommentar-nennt-mehr-elemente-als-der-test-prueft` berühren Auftrag 2 (Kürzung darf nichts zusagen, was kein Fänger trägt). Register am Start gesichtet (2026-11-19, Zähler aus den Dateien unter `evidence/`): `menge-ohne-ihre-positionen` (6×), `verweis-auf-review-report-sperrt-die-archivierung` (3×), `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (4×) und `laufzeitmessung-widerspricht-sich-unter-last` (3×) liegen über der Schwelle, ohne Zielort und ohne Nehmer; sie sind nicht Gegenstand dieses Slice und stehen als mögliche Ergänzung in §6 (Adresse User). `zeit-zusage-an-einer-form-und-einer-locale` und `allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle` (je 1×) bleiben unter der Schwelle gezählt.

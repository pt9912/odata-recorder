# Slice slice-dev-container-dokument: Dokument für Entwickler zum Arbeitscontainer

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-RB-03` (Entwicklung unter Linux und in Docker-Umgebungen, Nachweis reproduzierbar), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Folgepflicht), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (die Verträge tragen die Regeln, das Dokument verweist).

**Berührte Spec-Stellen:** die Verträge der Vorgänger (nur gelesen; Abweichungen sind Befunde, keine Änderungen dieses Slice) · Rang 6 der Source Precedence in `AGENTS.md` §2 und `harness/README.md` (`docs/user/*`, falls der Architect das Verzeichnis wählt; die Rangtabelle nennt es schon, sie ändert sich nicht) · `README.md` und `harness/README.md` (Zeiger auf das Dokument).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Dokument für Entwickler führt vom leeren Host `u-boot` bis zum grünen Lauf eines eingelieferten Zweigs im Arbeitscontainer, so dass ein anderer Mensch ohne Rückfrage nachfahren kann, und nennt die Fehlerbilder mit ihrer Ursache.

**Auftrag an den Architect** (Entscheidungen vor dem Schreiben, `AGENTS.md` §3.12; der Planner entscheidet sie nicht):

1. **Ort und Rang**: wo das Dokument liegt (Kandidat: das Verzeichnis für Rang 6 der Source Precedence, `docs/user/*` in `AGENTS.md` §2; es gibt es noch nicht) und wie es in `.d-check.yml` und in die Zeiger von `README.md` und `harness/README.md` kommt, ohne die Rangtabellen zu ändern.
2. **Grenze zu den Verträgen**: das Dokument führt die Schritte und die Fehlerbilder, die Regeln (Fristen, Exit-Codes, Namensschema) stehen im Vertrag; das Dokument verweist darauf und wiederholt sie nicht (`AGENTS.md` §3.8 sinngemäß: Regeldetails gehören an eine Stelle).
3. **Prüfbare Aussagen**: welche Aussagen des Dokuments ein Sensor hält (jedes genannte `make`-Ziel steht im Makefile, `AGENTS.md` §4; die Verweise löst `make docs-check` auf) und welche nur ein Probelauf belegt; Letztere stehen in der Tabelle *Zusage · Fänger* mit dem Probelauf als Fänger oder als akzeptiertes Negativ mit Grund und Folge.
4. **Probelauf**: wer ihn fährt (Rolle ohne Vorwissen, `Verifier` oder `Validator`), mit welchem Eingang (nur das Dokument) und wie Abweichungen berichtet werden.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neues Verhalten oder Änderungen am Container, am Zugang, an der Einlieferung oder am Lauf — anderer Vorgang: eine Abweichung zwischen Dokument und Verhalten ist ein Befund an den Planner (die Vorgänger liegen in `done/`, keine Adresse, `AGENTS.md` §3.13).
- Ein Handbuch für das Produkt (Betrieb des Recorders, Konfiguration) — Bestand: `spec/spezifikation.md` und `README.md`; das Dokument handelt vom Arbeitscontainer.
- Der Ablauf der Rollen im Container (wer wann einliefert) — Welle §6; das Dokument nennt, wie ein Entwickler einliefert und fährt, nicht wie die Rollen zusammenarbeiten.
- Eine Übersetzung — das Repo schreibt Deutsch (`README.de.md` neben `README.md` ist Bestand und wird hier nicht berührt).

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

- [ ] Das Dokument führt: Voraussetzungen am Host, Start und Bereitschaft, Schlüssel, Klon und Einlieferung, Lauf von `make gates` und `make gegenproben`, Limits und Parallelbetrieb, Fehlerbilder (Bereitschaft rot, Schlüsselwarnung, Zweig abgewiesen, Limit erreicht, Frist gerissen) mit Ursache und nächstem Schritt, Aufräumen; jede Regel steht im Vertrag und wird nur verlinkt.
- [ ] Ein Probelauf einer Rolle ohne Vorwissen mit dem Dokument als einzigem Eingang endet im grünen Lauf eines eingelieferten Zweigs auf `u-boot`; jede Abweichung ist als Befund berichtet und das Dokument nachgezogen (Bericht unter `docs/reviews/`).
- [ ] `make docs-check` ist grün; jedes im Dokument genannte `make`-Ziel steht in `make help`; die Zeiger in `README.md` und `harness/README.md` lösen auf.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `README.md` und `harness/README.md` zeigen auf das Dokument (`AGENTS.md` §3.9).
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
| Dokument für Entwickler (Ort nach Entscheidung 1) | neu | der Weg vom leeren Host bis zum grünen Lauf, Fehlerbilder |
| `.d-check.yml` | update (falls nötig) | Verzeichnis des Dokuments in den Scan; Änderung nur, wenn der Architect sie verlangt, ohne Lockerung einer Prüfung (`AGENTS.md` §3.6) |
| `README.md`, `harness/README.md` | update | Zeiger auf das Dokument |
| Bericht des Probelaufs unter `docs/reviews/` | neu | Befunde der Rolle ohne Vorwissen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) entsteht beim Architect **vor** dem Übergang `next` → `in-progress`; sie ist noch leer.

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`; Befehle und Pfade im Dokument aus dem Vertrag und aus `make help` übernommen, nicht aus dem Gedächtnis (`harness/conventions.md` §Was der mitgelieferte Baum ist: maßgeblich ist allein `make help`).


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-dev-container-lauf` liegt in `done/`; die Entscheidungen 1 bis 4 und die Tabelle *Zusage · Fänger* stehen vor dem ersten Commit (Architect).

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Slice liefert Text und Zeiger, keinen Code und keine Tabelle mit mehr als wenigen Zeilen; der Planner setzt die Marke nach dem Architect mit `make groessenmarke` (`SPEC-128`), falls `slice-groessenmarke-werkzeug` dann in `done/` liegt, sonst nach der Form von `AGENTS.md` §3.14.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke; offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten).
- `in-progress` → `open` (blockiert): der Probelauf deckt eine Abweichung des Verhaltens auf, die kein Dokument behebt (Befund an den Planner; der Slice des Verhaltens liegt in `done/`, keine Adresse, `AGENTS.md` §3.13).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make docs-check` sind grün, der Probelauf ist berichtet und seine Befunde sind nachgezogen; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Das Dokument wiederholt Regeln der Verträge und driftet von ihnen (`BEO-REPO/berichtigung-erreicht-nicht-jede-stelle`, verkörpert in `AGENTS.md` §3.9) — **Ausgang:** offen; Entscheidung 2: das Dokument verweist; wer einen Vertrag berichtigt, sucht das Dokument mit (`AGENTS.md` §3.9).
- Ein Schritt des Dokuments nennt ein Ziel oder einen Pfad, den es nicht gibt — **Ausgang:** offen; Fänger: `make docs-check` für Pfade, `make help` für Ziele (Entscheidung 3), der Probelauf für den Rest.
- Das Dokument beschreibt einen Stand, der noch nicht gemerged ist, weil die Welle in einem Arbeitsbaum entsteht — **Ausgang:** offen; der Probelauf läuft auf dem gemergten Stand der Welle.
- Randformen: ein Dokument liefert keinen Vertrag im Sinne von `AGENTS.md` §3.12 — **Ausgang:** entfallen: Raster *Operation × Argumentart* gilt nicht, das Dokument hat keine Operation mit Argumenten; die Fehlerbilder stehen im Dokument aus den Verträgen der Vorgänger.

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30; keiner steht bei 3× und offen). Für diesen Slice zählen: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `verkörpert` in `AGENTS.md` §3.9) — das Dokument ist eine Stelle, die ein berichtigter Vertrag mittragen muss (Risiko oben); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — Betriebsregeln in §3.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

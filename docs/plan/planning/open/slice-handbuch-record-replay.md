# Slice slice-handbuch-record-replay: Kapitel Aufzeichnen und Wiedergeben mit belegten Beispielen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-FA-02` (Aufzeichnung), `LH-FA-04` (Replay), `LH-FA-05` (Matching), `LH-FA-14` (Schutz sensibler Daten), `LH-RB-01` (unterstützter Teilumfang), `LH-RB-04` (dokumentierte Schutzmaßnahmen), `LH-QA-05` (Fehlerdiagnose).

**Berührte Spec-Stellen:** `SPEC-110` (Startkonfiguration) · `SPEC-052` (Replay-HTTP-Identität und Lebenszyklus) · `SPEC-045` (Capture-/Replay-Fähigkeit) · `SPEC-121` (HTTP-Eingang) · `spezifikation.md §4` (Fehler-Codes) — der Architect ergänzt die Stellen, die die Kapitel tragen, nach dem Stand von M2.
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

**Ziel:** Das Handbuch beschreibt, wie ein Anwender mit dem gebauten Binary gegen einen Upstream aufzeichnet und ohne Upstream wiedergibt, was er dabei vom Schutz der Daten wissen muss (Redaction, was nicht in der Aufzeichnung steht), welchen OData-Teilumfang der Recorder trägt und welche Fehlercodes ihm dabei begegnen; jedes Beispiel liegt als Datei vor und läuft gegen das gebaute Binary und den Referenzdienst aus `slice-bootstrap-m2`.

**Auftrag an den Architect** (vor dem ersten Code-Commit; `AGENTS.md` §3.12, §3.15, §3.16):

1. **Welche Aufgaben ein Anwender nach M2 allein ausführen kann.** Der M2-Test erreicht Session beenden, Import und Workspace über den Composition Root, nicht über HTTP; die Admin-API ist Folgewelle (Welle `welle-record-replay` §6). Ein Anwender des gebauten Binary hat danach womöglich keinen Weg, eine Aufzeichnung zu beenden, zu importieren und einen Workspace anzulegen, und damit keinen vollständigen Rundlauf. Der Architect legt fest, welche Aufgaben des Standards (Abschnitt 2, 5) das Binary dem Anwender hergibt, und trennt: Aufgaben, die das Handbuch beschreibt, von Gegenständen, die erst die Admin-API ausführbar macht. Ergibt sich, dass der Rundlauf für Anwender erst mit der Admin-API existiert, ist das ein Befund an den Planner: Der Slice schrumpft auf das Aufzeichnen und die Konfiguration von Replay, oder er wartet auf die Admin-API (§6, Ausgang dort).
2. **Schnittstellen des Anwenders.** Konfiguration (Umgebungsvariablen, Dateien, `SPEC-110`), Modus-Wahl (`record`, `replay`), Zuordnung im Replay (`SPEC-052`), Ports und Pfade der HTTP-Eingänge, Ausgabe und Logs; je Angabe der Beleg (Test des Binary, Gate), sonst entfällt sie.
3. **Schutz der Daten.** Was Redaction aus Aufzeichnungen fernhält und was nicht (`LH-FA-14`, `LH-QA-04`): nur Aussagen, die ein Test mit Canary-Werten belegt; Verantwortung des Betreibers für zulässige Testdaten (`LH-RB-04`) als Hinweis im Wortlaut der Spezifikation.
4. **Unterstützter Teilumfang** (`LH-RB-01`): als Tabelle *unterstützt · nicht unterstützt* aus Spezifikation und Tests des Codec; nicht unterstützte Funktionen erscheinen als nicht unterstützt, nicht als Auslassung.
5. **Fehlerbilder.** Je Fehlercode, den das Binary im Betrieb von M2 erzeugt, Ursache und Lösung (Standard Abschnitt 8) und die Zeile in der Tabelle der Fehlercodes (`slice-handbuch-meldungskatalog`); der Abgleich bleibt grün.
6. **Beleg der Beispiele.** Ort der Beispiel-Dateien, Test (Stufe `test`) und Bindung nach der Entscheidung aus `slice-handbuch-geruest`; ein Beispiel ohne Datei ist kein Beispiel.
7. Raster *Operation × Argumentart* (§6) und Tabelle *Zusage · Fänger* (§3) füllen, Größe angeben.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Mock-Modus, Hybrid-Modus, Fallback, Admin-API, Snapshots, Telemetrie, TLS des Listeners — Folgewellen (Welle §6); ihre Kapitel entstehen dort.
- Zahlen zur Performance (`LH-QA-01`) — ungeprüft (Lasttest fehlt, `welle-record-replay` §6); das Handbuch sagt nichts zu.
- Der Probelauf durch eine Rolle ohne Vorwissen — `slice-handbuch-abnahme`; dieser Slice belegt Beispiele durch Tests, nicht die Verständlichkeit.
- Änderungen am Binary, am Referenzdienst oder am M2-Test — andere Schicht; ein Unterschied zwischen Spezifikation und Binary ist ein Befund an den Planner (die Nehmer liegen in `done/`, `AGENTS.md` §3.13).

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

- [ ] Kapitel Aufzeichnen und Wiedergeben nach Auftrag Nr. 1 bis 5: Aufgaben mit Voraussetzung, Vorgehen und Ergebnis, Schutz der Daten, Teilumfang, Fehlerbilder; jede Aussage über Verhalten hat ihren Beleg (`LH-FA-02`, `LH-FA-04`, `LH-FA-14`, `LH-RB-01`, `LH-RB-04`).
- [ ] Jedes Beispiel liegt als Datei vor und läuft in der Stufe `test` gegen das gebaute Binary; je Beispiel eine Mutation (Wert im Handbuch geändert), die den Test rot macht.
- [ ] Der Abgleich der Fehlercodes ist grün; `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Handbuch ist das Update; Kopf (Stand, Software-Version) und Änderungsverlauf nachgezogen.
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
| docs/user/benutzerhandbuch.md | update | Kapitel nach Auftrag Nr. 1 bis 5 |
| Beispiel-Dateien des Handbuchs | neu | Ort nach `slice-handbuch-geruest` |
| Test der Beispiele | neu oder update | ein Fall je Beispiel gegen das gebaute Binary und den Referenzdienst aus `slice-bootstrap-m2` |

Der Architect füllt die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und ergänzt die Zeilen oben; sie ist beim Start leer.

| Zusage | Fänger |
|---|---|
| — | — (Architect) |

**Betriebsregeln für die Rollen dieses Slice:** Dateien des Repos nur mit `Edit` und `Write`, kein `sed -i`, kein Heredoc auf Repo-Dateien (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`). Beispiele mit Backslash-u-Folge, BOM, NBSP oder Zeichen jenseits von ASCII (Bodies, Header) entstehen aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Der Handoff ist ein Commit. Docker-Läufe mit `--memory` und wenige parallel. Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/` (M2-Test grün, Binary mit Modus `record` und `replay`, Abgleich der Fehlercodes im Gate); der Architect hat die Aufträge aus §1 beantwortet, die Tabelle in §3 und das Raster in §6 gefüllt, bevor der erste Code-Commit entsteht.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect mit `make groessenmarke` (`SPEC-128`) aus der Tabelle in §3, Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; nebenläufig zählt der Test der Beispiele, wenn er Binary, Referenzdienst und Client in Prozessen startet (× 85 je Zeile).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code über 3000 nach der Formel; ein Schnitt beim Handoff nur an dem Teil, den §4 vor dem Code benennt. Vorab benannter Teil, wenn der Architect ihn bestätigt: die Fehlerbilder (Auftrag Nr. 5) in eigenem Commit.
- `in-progress` → `open` (blockiert): Nr. 1 ergibt, dass der Rundlauf für Anwender erst mit der Admin-API existiert; der Slice wartet dann auf deren Welle oder schrumpft, die Entscheidung trifft der Planner mit dem User.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, die Beispiele laufen als Teil von `make test`, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Rundlauf Aufzeichnen, Beenden, Import, Workspace, Wiedergeben ist für einen Anwender des Binary womöglich nicht ausführbar, solange die Admin-API fehlt (Auftrag Nr. 1) — **Ausgang:** offen; Adresse: der Planner und der User (Umfang dieses Slice oder Warten auf die Welle der Admin-API).
- Das Handbuch sagt mehr zu als ein Test prüft (`BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, 25×, `verkörpert` in `AGENTS.md` §3.15) — **Ausgang:** offen; Fänger: Tabelle *Zusage · Fänger* und Mutationen am Beispiel.
- Eine Berichtigung am Verhalten des Binary erreicht das Handbuch nicht (`BEO-REPO/berichtigung-erreicht-nicht-jede-stelle`, 11×, `verkörpert` in `AGENTS.md` §3.9) — **Ausgang:** offen; Folge für die Folgewellen: ihre Slices ziehen das Handbuch im selben Commit nach, die Pflicht schlägt die Closure der Welle dem User vor.
- Beispiel mit Header oder Body, die Rohbytes tragen — **Ausgang:** offen; Werte aus Bytewerten und Rohbyte-Gegenlesen (`BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`, 4×, `verkörpert` in `AGENTS.md` §3.18).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster unten als Gerüst; die Zellen füllt der Architect vor dem ersten Code-Commit, eine leere Zelle ist ein Fehler des Plans — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Aufzeichnen starten (Beispiel) | Architect | Architect | Architect | Architect | Architect |
| Wiedergeben starten (Beispiel) | Architect | Architect | Architect | Architect | Architect |
| Fehlerbild auslösen (Beispiel je Code) | Architect | Architect | Architect | Architect | Architect |
| Teilumfang prüfen (Tabelle gegen Codec) | Architect | Architect | Architect | gilt nicht: reiner Lesevorgang, Architect bestätigt | Architect |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30). Für diesen Slice zählen: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×), `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×), alle `verkörpert` in `AGENTS.md`, und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`). Der Planner liest den Stand dieser Einträge beim Start neu, denn bis dahin schreiben andere Slices das Register fort.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

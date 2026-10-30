# Slice slice-groessenmarke-werkzeug: Größenmarke als Werkzeug statt als Prosa

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md) — Einordnung vom Planner empfohlen, Entscheidung des Users (Bericht der Closure von `slice-baseline-import`); Reihenfolge nach `slice-baseline-import` und vor `slice-workspace-use-case`.

**Bezug:** `LH-QA-06` (Portabilität und Reproduzierbarkeit: das Werkzeug läuft im Container, netzlos), Register-Eintrag `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×; Ausgang `geplant` mit dieser Kennung), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Verträge der Werkzeuge in `harness/vertraege/`, `MR-001`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Gate-Kette), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Werkzeuge in Go).

**Berührte Spec-Stellen:** `harness/vertraege/` — ein neuer Vertrag (Kennung vergibt der Architect, nächste freie Nummer nach `harness/vertraege/README.md` §Kennungen) · `SPEC-102` (Fälle der Gegenproben, Stand und Grenze) · `harness/README.md` §Sensors (Zeile des neuen Ziels). Produkt-Spezifikation (`spec/spezifikation.md`) und Architektur-Sicht bleiben unberührt: der Slice liefert kein Anwendungscode.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-25.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Werkzeug rechnet aus der Tabelle *Zusage · Fänger* eines Slice-Plans die Schätzung der Größe und prüft die Marke gegen die Schätzung und die Messung (`git diff --numstat <Basis>..HEAD` über die Pfade aus §4 des Plans); es läuft vor dem ersten Code-Commit und beim Handoff, und eine Überschreitung braucht einen Eintrag in §4 des Plans mit Grund, sonst ist das Ergebnis rot. Die Faktoren der Formel stehen in `harness/vertraege/`, nicht in `AGENTS.md`; dort bleibt nur ein Zeiger auf das Werkzeug.

**Übernimmt (Register):** Den Ausgang `geplant` von `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×; Entscheidung des Users in der Closure von `slice-baseline-import`: Sensor statt Prosa). Der Eintrag stand auf `verkörpert` in `AGENTS.md` §3.14; die Prosa-Regel hat 16 Überschreitungen nicht verhindert und die Messung des dritten Punkts (`slice-baseline-import`) widerlegt ihren Leser-Zuschlag. Diese Closure setzt den Ausgang auf `verkörpert` mit dem Ziel des Werkzeugs als Zielort.

**Übernimmt (Messpunkte):** Drei gemessene Slices als Ausgangsmaterial für die Entscheidung des Architect über Formel und Faktoren (Quellen: §4 von [slice-odata-codec-csdl](../done/slice-odata-codec-csdl.md), [slice-odata-codec-json](../done/slice-odata-codec-json.md) und [slice-baseline-import](../done/slice-baseline-import.md)):

| | csdl | json | import |
|---|---|---|---|
| Zeilen der Tabelle (davon nebenläufig) | 44 (2) | 49 (2) | 70 (3) |
| Port-Fakes | 0 | 0 | 8 |
| Produktcode am Ende | 1506 | 1640 | 1802 |
| Handoff (Summe aus `git diff --numstat`, Tests eingeschlossen) | 3628 | 5612 | 7551 |
| Ende nach dem Review | 4333 | 7236 | 10027 |
| Wachstum Handoff → Ende | +19,4 % | +28,9 % | +32,8 % |
| Tests und Fakes je Zeile der Tabelle, Handoff / Ende | 48 / 64 | 81 / 114 | 82 / 117 |
| Marke, die der Slice trug | 2400 | 4500 | 8200 |

**Ausgangsmaterial, keine Entscheidung.** Die Formel auf der Platte (`AGENTS.md` §3.14: nebenläufig × 85, Leser fremder Bytes × 65, sonst × 40, plus 200 je Fake, Produktcode als fester Anteil, × 1,25) hält das Ende bei csdl (Marke 5600, Ende 77 % davon) und verfehlt es bei json (Marke 6000, Ende 120,6 %) und import (als Nicht-Leser Marke 8200, Ende 122 %). Der Vorschlag der Closure von `slice-baseline-import` (alle anderen Zeilen × 80, Wachstum × 1,35) hält alle drei (Marken 6800, 7400, 12500; Enden 63,7 %, 97,8 %, 80,2 %), bei Schätzungen gegen den Handoff von +39 %, −3 % und +22 %. Die Stichprobe hat drei Punkte, zwei davon im Regime von `AGENTS.md` §3.15 bis §3.17; der Faktor 80 ist kein Naturgesetz, und ein vierter Punkt (`slice-workspace-use-case`: kein Leser, Dienst, Persistenz) kann den Faktor zu 50 bis 60 verschieben. Ob die Formel diese Form behält, ob die Art des Slice (Leser, Dienst, Adapter) ein Faktor bleibt und welche Zahlen gelten, entscheidet der Architect im Vertrag; dieser Plan setzt keine Zahl.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Entscheidung über Formel, Faktoren, Wachstumszuschlag und Schnittgrenze — der Architect entscheidet sie im Vertrag vor dem ersten Code-Commit (`AGENTS.md` §3.12); der Planner setzt hier Ziel und Grenze, der Implementer entscheidet im Code nichts (§6).
- Eine Änderung von `AGENTS.md` — `AGENTS.md` ändert nur der User. Die Prosa von §3.14 gilt, bis der User den Zeiger auf das Werkzeug annimmt; der Slice liefert dafür in der Closure einen Diff-Vorschlag (§2, Punkt 3), schreibt ihn nicht.
- Die Generator-Pflicht (`BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`) und die Stellen-Regel (`BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`) — andere Vorgänge, andere Orte (Command `implement-slice`, Implementer-Typ, Reviewer-Skill, Verifier-Typ); sie gehen als Diff-Vorschläge an den User und haben in dieser Closure ihre eigene Adresse (Register: `geplant`, Zielort benannt).
- Das Schneiden eines Slice — das Werkzeug meldet Schätzung, Marke und Überschreitung, es schneidet nicht; ob und wo ein Slice geteilt wird, bleibt Entscheidung des Planners vor dem Code (`AGENTS.md` §3.14, Schnitt über 3000).
- Eine rückwirkende Prüfung der Pläne in `done/` — ihre Form war vor dem Werkzeug frei; sie sind Ausgangsmaterial der Faktoren (oben), nicht Gegenstand der Prüfung.
- Anwendungscode und Schichten des Hexagons — der Slice berührt `tools/`, `harness/` und `Dockerfile`, nicht `src/`; `make a-check` prüft ihn nicht.
- Die Sendung an andere Slices — keiner nimmt etwas aus diesem Slice entgegen (§3.13 geprüft: `slice-workspace-use-case` und die übrigen Pläne in `open/` bleiben unverändert, ihre Marken setzt der Planner bei deren Start nach dem Werkzeug); die Reihenfolge steht allein im Welle-Plan.

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

- [ ] Vertrag und Rechner: der Vertrag in `harness/vertraege/` (Formel, Faktoren mit Herkunft der Messpunkte, Eingabeform aus §3 und §4 eines Slice-Plans, Randformen, *Ungeprüft*) steht vor dem Code (Architect); ein Rechner der Schätzung aus der Tabelle *Zusage · Fänger* (Zeilen, Klasse nebenläufig oder andere, Port-Fakes, fester Anteil des Architect) liegt als Werkzeug in Go unter `tools/` und hat einen Test je Zusage des Vertrags.
- [ ] Sensor: das Werkzeug liest Marke, Basis und Pfade aus §4 des Plans, misst `git diff --numstat <Basis>..HEAD` über diese Pfade und prüft zwei Zeitpunkte, vor dem ersten Code-Commit (Marke gleich Schätzung × Zuschlag, aufgerundet) und beim Handoff (Messung gegen Marke); eine Überschreitung ohne Eintrag in §4 mit Grund macht es rot. Das Ziel hat seine Zeile in `harness/README.md` §Sensors mit Bindung; ob es an `GATE_CHECKS` hängt oder als Handoff-Prüfung läuft, entscheidet der Architect im Vertrag (§6).
- [ ] Gegenproben und Zeiger: jede Zusage des Vertrags hat einen Fall der Gegenproben (`SPEC-102`, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)), darunter die drei Messpunkte als Eingabe (Platte gegen Vorschlag, wie entschieden); der Diff-Vorschlag an den User für `AGENTS.md` §3.14 (Zeiger statt Formel, `AGENTS.md` §3.9: alle Stellen, die die Formel tragen, nachgezogen, darunter `welle-record-replay.md` §4 Auftrag 1) liegt vor.
- [ ] `make gates` grün, `make gegenproben` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Vertrag, `harness/README.md` §Sensors (Zeile und *Nicht behauptet*), `harness/vertraege/README.md` (Index).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; der Ausgang von `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` wechselt von `geplant` auf `verkörpert` (Zielort: das Ziel des Werkzeugs).
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

Der Architect vervollständigt diese Tabelle und fügt die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) vor dem Code ein; die Zeilen unten sind die Pfad-Kandidaten des Planners.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/SPEC-<NNN>-groessenmarke.md` | neu | Vertrag des Werkzeugs: Formel, Faktoren und ihre Herkunft, Eingabeform, Randformen aus §6, *Ungeprüft*; geschrieben vom Architect, Zeile im Index `harness/vertraege/README.md` |
| `tools/` (ein Go-Modul, Ort und Name vom Architect) | neu | Rechner der Schätzung und Sensor; Werkzeuge in Go ([ADR-0001](../../adr/0001-implementierungssprache-go.md)), Test je Zusage |
| `Dockerfile`, `harness/mk/` | update | Stufe oder Aufrufform des Werkzeugs und das Ziel `make <name>`; netzlos, soweit `git` es erlaubt (§6) |
| `harness/README.md` | update | Zeile des Ziels in §Sensors mit Bindung, im selben Commit wie das Ziel (sonst meldet `make docs-check` ein Target ohne Zeile) |
| `tools/gegenproben/` | update | Fälle zu `SPEC-102`: je Zusage ein Fall in einer Kopie oder einem Wegwerfmodul ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); Stempel `harness/gegenproben.stempel` folgt nach dem letzten Commit unter den Eingabepfaden |
| `docs/plan/planning/welle-record-replay.md` | update | Auftrag beim Start eines Slice (Formel) zeigt auf das Werkzeug, sobald der User den Zeiger in `AGENTS.md` angenommen hat (`AGENTS.md` §3.9) |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Vertrag samt Formel und Faktoren, maschinenlesbare Form von §4, Ort in der Kette, Gegenproben-Fälle, Tabelle *Zusage · Fänger*, Raster, Paare, Nachzüge, Sendungen, erwartete Produktcode-Größe) → Planner (Marke, eigener Commit, vor dem ersten Code-Commit; Wahl der Formel in §4) → Implementer (erster Code-Commit erst nach der Marke) → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur nach Diff-Vorlage an den User). Der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt, ist die `<Basis>` der Marke (§4). Hält der Implementer an einer Zelle ohne Ort an (`AGENTS.md` §3.12), geht sie an den Architect, und §6 trägt sie im selben Commit mit Ausgang *offen* (`AGENTS.md` §3.13).

**Betriebsregeln für alle Rollen** (Lehren der Läufe vom 2026-10-24: OOM durch Mutationen ohne Grenze, Zeitgrenzen der Gegenproben unter Last):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; höchstens zwei Container gleichzeitig; Mutationen nacheinander, nie parallel. Eine Mutation, die ein Limit entfernt, läuft mit `--memory=2g` und `timeout`. Vor einem großen Lauf `free -g`. Docker sonst nur über `make` (`AGENTS.md` §3.1).
- `make gates` nie parallel zu Mutationsläufen: die Zeitgrenzen der Gegenproben und von `pruefung` sind für einen ruhigen Host bemessen, ein Rot unter Last ist kein Befund, bevor es auf ruhigem Host wiederholt ist (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (Scratch außerhalb des Repos; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Das Werkzeug setzt eine Folge aus Backslash, `u` und vier Hex-Ziffern still um (`AGENTS.md` §3.18): Fixtures, Pläne und Berichte mit solchen Folgen, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen; der Bericht nennt die gegengelesene Datei.
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (`AGENTS.md` §3.17; `.claude/commands/implement-slice.md` Schritt 14 und 19), nie eine Liste der Fälle, die einfielen. Eine Hilfe, die ein Argument umhüllt (Strom, Fake, `ctx`), darf keine Eigenschaft des Originals verdecken: sie trägt jede Schnittstelle des Originals, zählt je Aufruf, und ein Fänger prüft die Hülle selbst (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13), nicht nur im Bericht. Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- `.claude/agents/`, `.claude/commands/`, `.harness/skills/` und `AGENTS.md` schreibt keine Rolle dieses Slice; was dort zu ändern ist, geht als Diff-Vorschlag an den User.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, beim Start gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der betroffenen Pläne):

1. *An diesen Slice* (Nehmer ist dieser Slice, er liegt jetzt in `in-progress/`): die drei Messpunkte und der Ausgang `geplant` des Registers aus der Closure von `slice-baseline-import` — angenommen; §1 *Übernimmt (Register)* und *Übernimmt (Messpunkte)* und DoD Punkt 1 und 3 tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie. Die Offenen Reste (2) und (3) aus `slice-baseline-import` §7 betreffen diesen Slice nur in einem Punkt: der Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Lauf über 30 s, gemessen 56,9 s) liegt bei jedem neuen Fall der Gegenproben im Weg (§6, Risiko *Laufzeit der Gegenproben*, Adresse User).
2. *Von diesem Slice an andere*: **keine Sendung an einen Slice.** Die Fassung dieses Plans nannte `slice-workspace-use-case` in §6 als Anlass der Neumessung der Faktoren; das war eine Adresse, die der Nehmer nicht annimmt (sein §1 und seine DoD kennen sie nicht, sein Start-Trigger nennt dieses Werkzeug nicht, und die Pläne in `open/` bleiben bis zu ihrem Start unverändert, Entscheidung des Users vom 2026-10-21). Berichtigt: die Neumessung ist ein **Auftrag des Planners beim Start** jedes Slice mit einer Marke (`welle-record-replay.md` §4, *Auftrag beim Start eines Slice aus `open/`*), kein Punkt eines Slice-Plans; sobald der User den Zeiger in `AGENTS.md` angenommen hat, trägt der Auftrag das Werkzeug statt der Formel. Die Marken der Nehmer mit freier Schätzung (`slice-workspace-use-case`, `slice-replay-service`, `slice-http-adapter`, `slice-bootstrap-m2`, `slice-protokolltreue`) setzt und prüft der Planner bei deren Start mit dem Werkzeug; das ist seine Handlung, keine Sendung. Der Welle-Plan nennt die Reihenfolge (`slice-workspace-use-case` nach diesem Slice) nur als Reihenfolge, nicht als Start-Trigger des Nehmers (offen, §6: ob der User den Start-Trigger von `slice-workspace-use-case` um „`slice-groessenmarke-werkzeug` liegt in `done/`“ ergänzt; der Planner trägt es dort bei dessen Start ein).
3. *Vorschläge an den User* sind keine Sendungen an Slices: `AGENTS.md` §3.14 als Zeiger, `.claude/commands/plan-welle.md` und `.claude/commands/implement-slice.md` (Aufrufstellen des Werkzeugs), Rollen-Typen und Reviewer-Skill; der Slice liefert den Diff in der Closure und schreibt ihn nicht. Der Vorschlag `slice-guide-groessen-gate` aus dem Retirement-Check (`docs/reviews/2026-10-25-retirement-check-hard-rules.md` §5) ist **kein Slice dieser Welle**; der Planner legt ihn nicht an, der Vorschlag steht im Bericht des Starts und geht an den User (§6, Risiko *Größe der Guide-Dateien*).

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), Muster ist `SPEC-102`; Kennung nach `harness/vertraege/README.md` §Kennungen; eine ADR nur, wenn eine Entscheidung sie verlangt, sonst der Vertrag, `AGENTS.md` §3.8):

1. **Vertrag des Werkzeugs** (Kennung, Titel, Zusagen, *Ungeprüft*): Eingaben (die Tabelle *Zusage · Fänger* eines Plans mit Spaltenköpfen, Zeilen und Klasse je Zeile; §4 des Plans mit Marke, Basis, Pfaden; die Ausgabe von `git diff --numstat`), Ausgaben (Schätzung nach der Formel mit ihren Bestandteilen, Marke, Messung, Überschreitungsvermerk), Meldungen, Exit-Codes (rot wegen Marke oder Messung gegen Fehler der Eingabe gegen `git` nicht erreichbar, getrennt), und welche Zeilen ein Mensch in §4 übernimmt. Die Zusagen des Auftrags in §1 sind die Decke, nicht der Wortlaut.
2. **Formel und Faktoren** mit Herkunft. Vorschlag des Planners aus der Closure von `slice-baseline-import` (Material des Planners, Diff-Vorschläge Teil 1; **ein Vorschlag, den der Architect prüft und begründet entscheidet, kein Beschluss**): alle anderen Zeilen × 80 statt × 40 oder × 65, nebenläufig × 85, plus 200 je Fake, Produktcode als fester Anteil, Wachstum Handoff → Ende × 1,35 statt × 1,25 (gemessen 19,4 bis 32,8 %; im Material „17 bis 33 %“ über die Nachmessungen), damit Marken 6800, 7400 und 12500 und Enden 63,7 %, 97,8 % und 80,2 % der Marke. Zu entscheiden und zu begründen: ob die Faktoren aus den drei Messpunkten (csdl 3628 → 4333, json 5612 → 7236, import 7551 → 10027; Tests und Fakes je Zeile der Tabelle 48/64, 81/114, 82/117 in den `done/`-Plänen) folgen oder bewusst konservativ stehen; ob die Art des Slice (Leser, Dienst, Adapter) ein Faktor bleibt; ob das Wachstum fest oder eine Spanne ist; ob der Vertrag Schätzung für den Handoff und für das Ende getrennt ausweist. Der Architect rechnet die drei Punkte aus den Plänen selbst nach (die Zahlen in §1 sind Ausgangsmaterial, vom Planner übernommen, nicht neu gemessen). **Datenlage:** drei Punkte, zwei davon im Regime von `AGENTS.md` §3.15 bis §3.17, bei zwei bis drei Parametern; die Formel hält die Punkte, aus denen sie entstand, und ist an keinem ungesehenen geprüft. Das steht unter *Ungeprüft*, ebenso der Anlass zur Neumessung (jeder Slice mit einer Marke nach dem Werkzeug trägt einen Messpunkt; wer ihn in den Vertrag einträgt und wann, entscheidet der Architect). Die Faktoren stehen im Vertrag, nicht in `AGENTS.md`, nicht in `plan-welle`, nicht in den Plänen.
3. **Maschinenlesbare Form von §4 und der Tabelle.** Feldnamen für Marke, Basis, Pfade, Schätzung, Anteil Produktcode, Zahl der Fakes, Überschreitung mit Grund (Eintrag, Gültigkeit bis zu welchem Stand, Ersatz einer älteren Marke); woran das Werkzeug die Klasse einer Zeile erkennt (nebenläufig, Leser, andere: Spalte oder Marker in der Tabelle); wie es eine Zeile erkennt, die der Autor als Allaussage mehrfach zählt. **Pfade:** §4 schließt „Text der Spezifikation und der Pläne“ aus, aber der Pfad `harness/` in §3 enthält `harness/vertraege/` (Text der Spezifikation der Werkzeuge); der Vertrag nennt die Pfadliste ausdrücklich und entscheidet, ob `harness/vertraege/` zählt. **Alte Form:** die Pläne in `open/` und `in-progress/` tragen §4 als Prosa; das Werkzeug erkennt die alte Form und meldet sie eindeutig, statt zu raten; wer sie nachzieht (der Planner bei deren Start), steht im Vertrag. Die Pläne in `done/` sind nicht Gegenstand der Prüfung, nur Fixtures.
4. **Ort in der Kette.** Entscheiden mit Gründen, ob das Werkzeug (A) ein Gate in `GATE_CHECKS` ist, (B) außerhalb der Kette läuft wie die Gegenproben ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): Pflicht zum Fahren in der Spezifikation, kein Gate erzwingt sie), oder (C) geteilt: Prüfung der Marke gegen die Formel (nur Dateien, keine Historie) in der Kette, Messung (braucht `git diff --numstat`) außerhalb. Kriterien: `AGENTS.md` §3.6 (ein Gate wird nicht gelockert; ein neues ist erlaubt); ein Gate, das an Ständen ohne Slice in Arbeit oder nach dem Review rot ist, ist ein falsches Gate; woran das Werkzeug erkennt, welcher Plan in Arbeit ist (Pläne unter `in-progress/`, Argument, Feld im Plan); ob die Bindung in `harness/README.md` §Sensors eine ADR braucht (Vorbild: [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) für die Lage außerhalb der Kette). **`git` im Container:** die Prüfläufe sind netzlos und ohne schreibbaren Mount (`AGENTS.md` §3.1); die Host-Abhängigkeiten des Repos sind `bash`, `git` und `docker` (`harness/conventions.md` §Der mitgelieferte Baum altert still), also kann der Host `git diff --numstat` rufen und die Zahlen als Eingabe in den Container geben; Alternativen sind ein schreibgeschützter Mount und `docker cp` von `.git`. Der Vertrag legt fest, was bei fehlender oder unerreichbarer Historie gilt und was das Werkzeug dann **nicht** behauptet. Zur Bindung gehört *Nicht behauptet* in `harness/README.md` §Sensors.
5. **Zeitpunkte und Zustände.** Vor dem ersten Code-Commit (Marke gleich Schätzung × Zuschlag, aufgerundet auf volle Hundert; woran das Werkzeug „vor dem ersten Code-Commit“ erkennt: leere Messung, Basis gleich Spitze, Argument), beim Handoff (Messung gegen Marke), nach dem Review und bei der Closure (Messung am Ende, Eintrag der Überschreitung mit Grund). Marke unter der Schätzung, Marke über der Schätzung, Messung gleich Marke, Marke gleich Schätzung × Zuschlag genau, Marke ersetzt eine ältere (Grund in §4 verlangt?), Rundung. Jede Grenze beidseitig (`AGENTS.md` §3.15).
6. **Raster und Paare.** Das Raster in §6 ist das Gerüst des Planners; der Architect schließt jede Zelle mit einem Ort im Vertrag oder „gilt nicht“ mit Grund und ergänzt Dimensionen, die dem Planner fehlen: Binärdateien (`-` in `numstat`), Umbenennungen (`{a => b}`), Pfade mit Leerzeichen und Sonderzeichen, Zahlen mit Tausendertrennung im Plan, Basis gleich Spitze, Basis kein Vorfahr, Merge-Commits, Leerer Diff über die Pfade, Pfad außerhalb des Baums. Die Zusatztabelle der Paare (§6) bekommt je Paar den Fänger und den Eintrag der Rangfolge der Meldungen. Zellen, die der Implementer nicht entscheidet (`AGENTS.md` §3.12), stehen alle vor dem ersten Code-Commit im Vertrag.
7. **Aufrufstellen.** Ruft das Werkzeug einen Prozess (`git`), gilt die Menge der Aufrufstellen mit Fehler, Abbruch und Zeitgrenze (`AGENTS.md` §3.15, §3.17); der Architect bestätigt oder nennt, ob das Werkzeug nebenläufig ist (Erwartung des Planners: nein) und ob es einen Fake für den Prozess-Aufruf braucht (zählt jeden Aufruf, hält jedes Argument fest, `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`).
8. **Gegenproben** (`SPEC-102`, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)): je Zusage des Vertrags ein Fall (Nummer, Kennung, Eingabe, Aufruf, erwartete Ausgabe und Exit-Code); die **drei Messpunkte** als Eingabe, jeder mit der Marke, die die Platte (`AGENTS.md` §3.14) ergäbe, und mit der Marke der neuen Formel; Fixtures mit Sonderzeichen aus Bytewerten (`AGENTS.md` §3.18); jede Menge als Schleife (Klassen der Zeilen, Formen von `numstat`); die Zeit, die die neuen Fälle zum Lauf bei geänderter Eingabe hinzufügen (Stand 56,9 s gegen den Trigger von 30 s in [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); die Eingabepfade des Stempels.
9. **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15) mit der Menge je Zeile aus dem Wortlaut der Zusage (§3.17), die Mutation je Zeile und der Fall mit zwei gleichen verglichenen Werten; je Klasse, ob sie generierbar ist (Probe-Paket des Verifiers: Pläne mit beliebig vielen Zeilen und Klassen, `numstat`-Ausgaben); jedes akzeptierte Negativ mit Grund und Folge. **Ungeprüft** nennt mindestens: das Werkzeug misst die Größe, nicht ob die Marke die richtige ist; die Marke ist so gut wie die Faktoren; die Klasse einer Zeile setzt der Autor des Plans; das Werkzeug erzwingt keine Rückführung und schneidet nicht (`AGENTS.md` §3.11).
10. **Nachzug an jeder Stelle, die dieselbe Aussage trägt** (`AGENTS.md` §3.9): Liste aller Stellen, die die Formel, die Marke oder die Schnittgrenze nennen — `AGENTS.md` §3.14 (nur als **Diff-Vorschlag an den User**: welche Sätze als Zeiger ins Werkzeug gehen, welche als Regel bleiben, etwa „Marke nennt Metrik“ und die Grenze 3000 vor dem ersten Code-Commit; der Architect schreibt `AGENTS.md` nicht), `.claude/commands/plan-welle.md` und `.claude/commands/implement-slice.md` (Aufrufstellen: der Planner setzt die Marke mit dem Werkzeug, der Implementer misst vor der Übergabe; ebenfalls nur Diff-Vorschlag), `welle-record-replay.md` §4 *Auftrag beim Start eines Slice aus `open/`* Nr. 1 und die Zeilen zur Formel (zieht der Planner in der Closure nach, nachdem der User den Zeiger angenommen hat), `harness/README.md` §Sensors, `harness/vertraege/README.md` (Index), `SPEC-102` (Fälle, Stand, Grenze), die Register-Datei `state.md`. Offene Pläne mit Formel in §4 nennt der Architect, ohne sie zu ändern.
11. **Größe** (`AGENTS.md` §3.14): die erwartete Produktcode-Größe des Werkzeugs in Zeilen Go (Rechner, Leser der Tabelle und von §4, Sensor, Aufruf von `git`; Spanne und Vergleich mit den Werkzeugen unter `tools/`), Zahl der Fakes, Zahl der Zeilen der Tabelle (davon nebenläufig). Das Werkzeug liest fremde Bytes (Markdown-Tabellen, `numstat`); ob der Leser-Zuschlag aus der Formel auf der Platte (× 65) gilt, entscheidet der Architect mit dem Vertrag (er wendet seine Formel auf sein eigenes Werkzeug an, §4). Er benennt, wenn es einen gibt, einen **Teil ohne geteilten Mechanismus**, der beim Handoff geschnitten werden dürfte (etwa die Fälle der Gegenproben gegen Rechner und Sensor), damit §4 ihn vor dem Code benennen kann. Liegt die Schätzung × 1,25 über 3000, schneidet der Planner vor dem Code.
12. **Zwei Fragen zu Nachbarn, die der Planner nicht entscheidet:** (a) `slice-guide-groessen-gate` (Größe von `AGENTS.md`, `implement-slice.md` und Reviewer-Skill durch das `file`-Modul von d-check) — gehört es zu diesem Vertrag (nein nach Einschätzung des Planners: anderer Mechanismus, `.d-check.yml` statt `tools/`, andere Grenzen, andere Rolle der Entscheidung, siehe Bericht), zu `make docs-check`, oder bleibt es unbeteiligt; der Architect schlägt vor, der Planner legt keinen Slice an. (b) `slice-plan-form-sensor` (Form der Tabelle und des Rasters): die Spaltenköpfe der Tabelle *Zusage · Fänger* aus Auftrag 3 sind die gemeinsame Kopplung; der Architect fixiert sie so, dass ein späterer Form-Sensor sie lesen kann, ohne dass dieser Slice ihn liefert. Eine Sendung an einen Folge-Slice weist er nur zu, wenn der Nehmer sie in §1 oder DoD schon trägt oder im selben Commit trägt (`AGENTS.md` §3.13); der Nehmer in `open/` ändert sich dabei nicht, der Punkt geht sonst an den Planner.

**Der Planner hat nichts davon entschieden.** Die Aufträge benennen die Entscheidungen und Mengen, die der Architect vor dem Code schließen muss; jede offene Entscheidung steht in §6 mit dem Ausgang *offen*. Gesetzt sind nur Ziel, Ausschlüsse und Reihenfolge.

**Aufträge an den Implementer** (nach der Marke): (a) der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); (b) der Absatz in §3 mit *Zusage · Mutation · roter Test* und den Raten (Zahl der Prozesse und Messender, `AGENTS.md` §3.10); (c) jeder Fix bildet die Menge aus dem Wortlaut der Zusage, schreibt den Fänger als Schleife und nennt je Befund Menge, Zahl der Elemente und eine Mutation an einem nicht genannten Element (`AGENTS.md` §3.17); (d) das Werkzeug läuft vor der Übergabe auf den Plan dieses Slice (Messung, Eintrag in §4 oder ohne Befund) und die Zahl steht im Bericht; (e) ein Fake des Prozess-Aufrufs zählt jede Operation und hält jedes Argument fest, falls der Vertrag einen verlangt; (f) die Betriebsregeln oben; (g) `AGENTS.md`, Commands, Typen und Skills werden nicht geschrieben (Diff-Vorschlag an den User).

**Aufträge an Reviewer und Verifier** (jeder in frischem Kontext): jede Zeile der Tabelle und das Raster mit je einer Mutation auflösen, Fixes gegen die **Klasse** (`AGENTS.md` §3.17), nicht gegen die genannte Mutation; der Verifier fährt zusätzlich ein **Probe-Paket aus Generatoren** über die Eingabeformen des Vertrags (Pläne mit wechselnder Zahl und Klasse der Zeilen, `numstat`-Formen, Marken auf und neben jeder Grenze), einmal gegen den Code, dann gegen die Mutationen, und legt es im Report ab (Pfad und Ergebnis); er rechnet die drei Messpunkte unabhängig nach. Zu prüfen ist auch, ob der Vertrag zusagt, was ein Fall der Gegenproben tatsächlich fängt (`AGENTS.md` §3.11). Betriebsregeln oben.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-baseline-import` liegt in `done/`. Der Vertrag samt Randformen aus §6 steht in `harness/vertraege/`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand des Starts (2026-10-30):** Der Start-Trigger ist erfüllt: `slice-baseline-import` liegt in `done/`. Vertrag, Tabelle und Raster stehen noch aus (Architect, §3); die Reihenfolge der Rollen steht in §3.

**Wahl der Formel für die Marke (vom Planner vor dem Architect festgelegt, damit der Schritt nicht frei bleibt):** Das Werkzeug misst sich selbst, existiert aber vor dem Code nicht; die Marke setzt der Planner deshalb von Hand. Er rechnet nach der Formel des **Vertrags**, wenn der Architect sie vor dem Code dort mit Faktoren und Herkunft geschrieben hat (dieser Slice liefert die Formel, die künftig gilt), und nennt in §4 zusätzlich die Zahl nach der Formel auf der Platte (`AGENTS.md` §3.14: Zeilen der Tabelle × 40 bzw. × 85, × 65 für Leser fremder Bytes, plus 200 je Fake, plus Produktcode, × 1,25, aufgerundet auf volle Hundert) als Gegenrechnung. Trägt der Vertrag keine Formel, gilt die Formel der Platte. Beide Zahlen und die Wahl mit Grund stehen im Marken-Commit. Der Lauf des Werkzeugs auf diesen Plan beim Handoff prüft die Marke nach der Formel des Vertrags; weicht sie von der von Hand gesetzten ab, ist das ein Befund zum Vertrag, nicht zur Marke.

**Größenmarke** (`AGENTS.md` §3.14): setzt der Planner vor dem ersten Code-Commit, nachdem der Architect die Tabelle *Zusage · Fänger* und die Produktcode-Erwartung in §3 eingefügt hat; die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade, die der Architect im Vertrag ausdrücklich nennt (vorläufig `tools/`, `harness/mk/`, `Dockerfile` und die Dateien des Werkzeugs; `harness/` als Ganzes enthielte `harness/vertraege/`, Text der Spezifikation, der nicht zählt, §3 Auftrag 3), Tests eingeschlossen, Text der Spezifikation und der Pläne nicht. Der Slice ist ein Werkzeug ohne Anwendungscode; sein erster Lauf ist der erste Messpunkt dieser Art, den er in §4 nennt (`AGENTS.md` §3.14, Zeilen, die keine reinen Funktionen sind).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke (sie schneidet nicht); offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten). Liegt die Schätzung nach der Tabelle über 3000, schneidet der Planner vor dem Code.
- `in-progress` → `open` (blockiert): der Vertrag lässt sich ohne Netz oder ohne Zugriff auf die Historie nicht erfüllen (§6, Risiko *git im Container*); das ist eine Entscheidung des Architect, kein Code des Implementers.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, und je Zusage des Vertrags ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Das Werkzeug hat auf den Plan dieses Slice selbst einen Lauf ohne Befund oder mit einem Eintrag in §4.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Formel und Faktoren: drei Messpunkte, zwei im Regime von `AGENTS.md` §3.15 bis §3.17; ein Faktor, der heute drei Punkte hält, kann den vierten verfehlen — **Ausgang:** offen; der Architect entscheidet Formel und Faktoren mit Herkunft im Vertrag, nennt die Grenze der Stichprobe unter *Ungeprüft* und den Anlass zur Neumessung (jeder Slice mit einer Marke nach dem Werkzeug; der erste ist `slice-workspace-use-case`, der Auftrag liegt beim Planner und im Welle-Plan, nicht im Plan des Nehmers, §3 *Prüfung der Sendungen* Nr. 2).
- `git` im Container: `git diff --numstat` braucht die Historie, die Prüfläufe laufen netzlos und ohne schreibbaren Mount (`AGENTS.md` §3.1) — **Ausgang:** offen; der Architect legt im Vertrag fest, wie die Historie in den Lauf kommt (schreibgeschützter Mount, `docker cp` des Verzeichnisses `.git`, Eingabe der Zahlen aus dem Host) und was bei fehlender Historie gilt.
- Maschinenlesbarkeit von §4: Marke, Basis und Pfade stehen heute als Prosa in §4 jedes Plans, in wechselnder Form — **Ausgang:** offen; der Architect legt eine feste Form fest (Feldnamen, Eintrag der Überschreitung mit Grund) und entscheidet, ob die Pläne in `open/` und `in-progress/` nachzuziehen sind; die Pläne in `open/` bleiben bis zu ihrem Start unverändert (Entscheidung des Users vom 2026-10-21), der Planner zieht sie bei deren Start nach.
- Ort in der Kette: ein Gate, das an jedem Stand rot ist, an dem eine Marke fehlt oder überschritten ist, blockiert `make gates` auch dort, wo kein Slice in Arbeit ist oder die Überschreitung nach dem Review erst entsteht — **Ausgang:** offen; der Architect entscheidet zwischen einem Gate in `GATE_CHECKS`, einer Handoff-Prüfung und beidem, mit Begründung gegen `AGENTS.md` §3.6 (kein Gate wird gelockert) und gegen die Bedingung, dass die Prüfung vor dem ersten Code-Commit und beim Handoff läuft.
- Der Bericht und die Gegenprobe sagen nur zu, was sie fangen (`AGENTS.md` §3.11): das Werkzeug misst die Größe, nicht ob die Marke die richtige ist — **Ausgang:** offen; der Vertrag nennt unter *Ungeprüft*, dass die Marke eines Slice nur so gut ist wie die Faktoren, und dass das Werkzeug die Rückführung nicht erzwingt.
- Henne und Ei: das Werkzeug misst sich selbst, die Marke dieses Slice entsteht vor dem Werkzeug von Hand — **Ausgang:** offen; Regel in §4 (*Wahl der Formel für die Marke*); der Lauf auf den eigenen Plan beim Handoff ist der erste Messpunkt dieser Art und prüft die Handrechnung des Planners; Abweichung ist ein Befund zum Vertrag.
- Pfade der Metrik: §4 nennt `harness/` als Pfad, `harness/vertraege/` (Text der Spezifikation) soll nicht zählen — **Ausgang:** offen; Adresse: Architect (§3 Auftrag 3), Pfadliste im Vertrag.
- Laufzeit der Gegenproben: der Lauf bei geänderter Eingabe misst 56,9 s (dritte Verifikation von `slice-baseline-import`), der Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) liegt bei 30 s und ist weder bestätigt noch ersetzt; jeder neue Fall dieses Slice verlängert den Lauf — **Ausgang:** offen; Adresse: User (Entscheidung zum Trigger, aus `slice-baseline-import` §7 *Offener Rest* (2)); der Architect nennt die Zusatzzeit der Fälle (§3 Auftrag 8).
- Start-Trigger von `slice-workspace-use-case`: der Welle-Plan nennt die Reihenfolge, der Plan des Nehmers in `open/` nennt dieses Werkzeug nicht als Bedingung — **Ausgang:** offen; Adresse: User (ob der Start-Trigger des Nehmers ergänzt wird; der Planner trägt es bei dessen Start ein, die Pläne in `open/` bleiben bis dahin unverändert, Entscheidung vom 2026-10-21).
- Größe der Guide-Dateien (Retirement-Check `docs/reviews/2026-10-25-retirement-check-hard-rules.md` §5, Vorschlag `slice-guide-groessen-gate`: `file`-Modul von d-check für `AGENTS.md`, Commands und Reviewer-Skill; weitere Vorschläge `slice-plan-form-sensor`, `slice-unsichtbare-zeichen`) — **Ausgang:** offen; Adresse: User (Vorschlag im Bericht des Starts; kein Slice angelegt, nicht Teil dieses Slice: anderer Mechanismus, anderer Pfad `.d-check.yml`, Grenzen erst nach der Verschlankung messbar; Architect nimmt Stellung, §3 Auftrag 12).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* als Gerüst; jede Zelle nennt die Frage, der Ort ist der Vertrag des Werkzeugs, den der Architect vor dem ersten Code-Commit schreibt; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Schätzung aus der Tabelle rechnen | Plan ohne Tabelle *Zusage · Fänger*, Zeile ohne Klasse: Fehler oder Schätzung 0? (Vertrag, offen) | Tabelle ohne Zeilen: Schätzung nur aus festen Anteilen? (Vertrag, offen) | Klasse außerhalb von nebenläufig und andere; Zeile mit Zusatztabelle: zählt sie? (Vertrag, offen) | gilt nicht: der Rechner liest eine Datei und schreibt eine Zahl — zu bestätigen | gleiche Eingabe, gleiche Zahl (Vertrag, offen) |
| Marke prüfen (vor dem ersten Code-Commit) | Marke fehlt in §4: rot (Vertrag, offen); Produktcode-Erwartung fehlt | Marke 0 | Marke unter Schätzung × Zuschlag: rot oder Eintrag mit Grund? Marke über der Schätzung: erlaubt? (Vertrag, offen) | gilt nicht (siehe oben) | Marke ersetzt eine ältere: Grund in §4 verlangt? (Vertrag, offen) |
| Messung prüfen (beim Handoff) | Basis fehlt oder unbekannt im Verlauf; Pfade fehlen (Vertrag, offen) | Diff über die Pfade leer: Messung 0, kein Code-Commit — grün oder rot? (Vertrag, offen) | Binärdatei (`numstat` meldet `-`), Umbenennung, Pfad außerhalb des Baums (Vertrag, offen) | `git` nicht erreichbar, Historie fehlt (Risiko oben) | Messung am selben Stand zweimal: gleiche Zahl; Messung nach Nachträgen im Commit des Plans (Pfade schließen Pläne aus) |
| Überschreitung | Eintrag in §4 ohne Grund: rot (Vertrag, offen) | Eintrag mit leerem Grund | Eintrag zu einer anderen Marke oder einem anderen Stand (Vertrag, offen) | gilt nicht | Überschreitung bei jeder weiteren Messung nach dem Eintrag: Eintrag trägt bis zu welchem Stand? (Vertrag, offen) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen am selben Plan (Fänger sind Fälle der Gegenproben, Zahl und Namen vom Architect):

| Paar | Entscheidung | Fänger |
|---|---|---|
| Marke fehlt und Basis fehlt | offen (Vertrag: welche Meldung zuerst) | offen |
| Marke unter der Schätzung und Messung über der Marke | offen (Vertrag: beide melden oder eine Rangfolge) | offen |
| Überschreitung mit Eintrag und Marke unter der Schätzung | offen (Vertrag) | offen |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, Stand 2026-10-25 nach Phase 2 der Closure von `slice-baseline-import`). Für diesen Slice zählen:
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×, Ausgang `geplant`, dieser Slice): der Gegenstand.
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, verkörpert in `AGENTS.md` §3.15) und `akzeptiertes-negativ-ohne-folge` (10×, §3.15): der Vertrag des Werkzeugs führt zu jeder Zusage einen Fänger, und was das Werkzeug nicht prüft, steht unter *Ungeprüft* mit Grund und Folge.
- `randform-erst-im-review-entschieden` (20×, §3.16) und `universalaussage-im-vertrag-ungeprueft` (9×, §3.16): das Raster in §6 steht vor dem Code, Allaussagen („jede Überschreitung braucht einen Eintrag“) nennen ihren Fänger über alle Fälle.
- `berichtigung-erreicht-nicht-jede-stelle` (11×, §3.9): die Formel steht heute an mehreren Stellen (`AGENTS.md` §3.14, `welle-record-replay.md` §4 und §3, offene Pläne); der Zeiger-Nachzug sucht alle.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, §3.18): Fixtures der Gegenproben mit Sonderzeichen werden aus Bytewerten gebaut und gegengelesen.
- `gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, offen) und `stempel-gehoert-zu-keinem-committeten-stand` (1×, offen): die Gegenproben des Werkzeugs und der Stempel nach dem letzten Commit unter den Eingabepfaden.
- `suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14, `.claude/agents/implementer.md` und `.harness/skills/reviewer.md`) und `fix-deckt-die-genannte-mutation-nicht-die-klasse` (5×, §3.17, Schritt 19): Mengen des Werkzeugs (Klassen der Zeilen, Formen von `numstat`, Grenzen der Marke) als Generatoren und Schleifen (§3 Auftrag 9, Betriebsregeln).
- `hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×, offen): trifft zu, falls der Vertrag einen Fake des Prozess-Aufrufs von `git` verlangt (§3 Auftrag 7).
- `betriebsregel-fuer-repo-dateien-verletzt` (1×, offen): Betriebsregeln in §3.
- `bericht-der-rolle-liegt-nicht-im-repo` (3×, §3.10): Absatz mit *Zusage · Mutation · roter Test* in §3 (Auftrag an den Implementer (b)).
Stand des Starts (2026-10-30): das Register trägt 19 Einträge; die Zahlen stammen aus den Dateien unter `evidence/` und den `state.md`. Unter der Schwelle ohne Bezug: die übrigen Einträge (`panik-eines-fakes-verdeckt-den-benannten-test`, `rate-ohne-detektor-schwankt-mit-der-last`, `zweig-ohne-lauf-unter-dem-detektor`; das Werkzeug ist nach Erwartung des Planners nicht nebenläufig, der Architect bestätigt es in §3 Auftrag 7).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

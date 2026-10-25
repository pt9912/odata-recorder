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

**Übernimmt (Messpunkte):** Drei gemessene Slices als Ausgangsmaterial für die Entscheidung des Architect über Formel und Faktoren (Quellen: §4 von [slice-odata-codec-csdl](../done/slice-odata-codec-csdl.md), [slice-odata-codec-json](../done/slice-odata-codec-json.md) und [slice-baseline-import](../in-progress/slice-baseline-import.md)):

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

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-baseline-import` liegt in `done/`. Der Vertrag samt Randformen aus §6 steht in `harness/vertraege/`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Größenmarke** (`AGENTS.md` §3.14): setzt der Planner vor dem ersten Code-Commit, nachdem der Architect die Tabelle *Zusage · Fänger* und die Produktcode-Erwartung in §3 eingefügt hat; die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3 (`tools/`, `harness/`, `Dockerfile`), Tests eingeschlossen, Text der Spezifikation und der Pläne nicht. Der Slice ist ein Werkzeug ohne Anwendungscode; sein erster Lauf ist der erste Messpunkt dieser Art, den er in §4 nennt (`AGENTS.md` §3.14, Zeilen, die keine reinen Funktionen sind).

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

- Formel und Faktoren: drei Messpunkte, zwei im Regime von `AGENTS.md` §3.15 bis §3.17; ein Faktor, der heute drei Punkte hält, kann den vierten verfehlen — **Ausgang:** offen; der Architect entscheidet Formel und Faktoren mit Herkunft im Vertrag, nennt die Grenze der Stichprobe unter *Ungeprüft* und den Anlass zur Neumessung (der erste Slice nach dem Werkzeug, `slice-workspace-use-case`).
- `git` im Container: `git diff --numstat` braucht die Historie, die Prüfläufe laufen netzlos und ohne schreibbaren Mount (`AGENTS.md` §3.1) — **Ausgang:** offen; der Architect legt im Vertrag fest, wie die Historie in den Lauf kommt (schreibgeschützter Mount, `docker cp` des Verzeichnisses `.git`, Eingabe der Zahlen aus dem Host) und was bei fehlender Historie gilt.
- Maschinenlesbarkeit von §4: Marke, Basis und Pfade stehen heute als Prosa in §4 jedes Plans, in wechselnder Form — **Ausgang:** offen; der Architect legt eine feste Form fest (Feldnamen, Eintrag der Überschreitung mit Grund) und entscheidet, ob die Pläne in `open/` und `in-progress/` nachzuziehen sind; die Pläne in `open/` bleiben bis zu ihrem Start unverändert (Entscheidung des Users vom 2026-10-21), der Planner zieht sie bei deren Start nach.
- Ort in der Kette: ein Gate, das an jedem Stand rot ist, an dem eine Marke fehlt oder überschritten ist, blockiert `make gates` auch dort, wo kein Slice in Arbeit ist oder die Überschreitung nach dem Review erst entsteht — **Ausgang:** offen; der Architect entscheidet zwischen einem Gate in `GATE_CHECKS`, einer Handoff-Prüfung und beidem, mit Begründung gegen `AGENTS.md` §3.6 (kein Gate wird gelockert) und gegen die Bedingung, dass die Prüfung vor dem ersten Code-Commit und beim Handoff läuft.
- Der Bericht und die Gegenprobe sagen nur zu, was sie fangen (`AGENTS.md` §3.11): das Werkzeug misst die Größe, nicht ob die Marke die richtige ist — **Ausgang:** offen; der Vertrag nennt unter *Ungeprüft*, dass die Marke eines Slice nur so gut ist wie die Faktoren, und dass das Werkzeug die Rückführung nicht erzwingt.
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
Unter der Schwelle ohne Bezug: die übrigen Einträge.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

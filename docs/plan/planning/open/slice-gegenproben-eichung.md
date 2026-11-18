# Slice slice-gegenproben-eichung: Der erste gültige Lauf von `make gegenproben` auf ruhigem Host

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): berührt `harness/vertraege/SPEC-102-gegenproben.md` und den Stempel `harness/gegenproben.stempel`, Nr. 7 (ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Laufzeit der Gegenproben, der Bezug ist der nächste, den es gibt), [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (Eichung als Folgepflicht, Re-Evaluierungs-Trigger), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Kette, Stempel), Register `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last`.

**Berührte Spec-Stellen:** `SPEC-102` (*Zeitbudget* Nr. 6 und 7: Eintrag der Messung und Absatz *Eichung* hinter den Messreihen; die Fälle 205 bis 216 und Randform 6 für `build/ohne-trimpath`); keine neue Nummer.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner der Closure von `slice-gegenproben-zeitbudget`). **Datum:** 2026-11-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Ein gültiger Lauf von `time make gegenproben` (beide Lesungen von `/proc/loadavg` höchstens gleich `nproc --all`, Eingaben vom Stempel abweichend, Images lokal) steht nach `SPEC-102` *Zeitbudget* Nr. 6 in den Messreihen und als Absatz *Eichung* nach Nr. 7; er sagt, ob das Soll von 120 s hält, wie groß der Eigenanteil des Laufs an der Lesung am Ende ist und ob die Fälle 205 bis 216 und `build/ohne-trimpath` bei gültiger Last grün sind.

**Übernimmt:** die Folgepflichten aus `slice-gegenproben-zeitbudget` (Eichung nach Nr. 7; Eigenanteil der Lesung am Ende; Dauerlast) und aus `slice-gates-ohne-worktrees` (Geber, liegt in `done/`; Messung von `time make gegenproben` mit den Fällen 205 bis 216 auf ruhigem Host, die neun Fälle `lint/*` mit `Zeit überschritten`, Ursache des Rot von `build/ohne-trimpath` bei Last 175 bis 222). Der Geber `slice-gates-ohne-worktrees` liegt in `done/` und ist keine Adresse; die Sendung ging über `slice-gegenproben-zeitbudget` §6 (Nehmer bis zur Closure) an diesen Slice.

**Voraussetzung (Start):** ein Termin, an dem alle Maschinen der Gruppe (w1, w2, w3) ihre schweren Läufe pausieren (`make gates`, `make gegenproben`, `make test-race`, `make groessenmarke`, Mutationsläufe, `docker build`); den Termin setzt der User (Bitte in §6 von `slice-gegenproben-zeitbudget`). Ohne ihn startet der Slice nicht.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Anhebung des Soll oder der Frist je Fall — Lockerung (`AGENTS.md` §3.6); liegt der gültige Lauf über 120 s, gilt der Re-Evaluierungs-Trigger von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) und eine neue ADR mit dieser Messung ist ein eigener Vorgang des Planners.
- Eine Änderung am Programm `tools/gegenproben`, am `Dockerfile` oder an `harness/mk/` — der Slice misst, er baut nichts um; ein Befund (Bruch, Fall mit `Zeit überschritten`, Rot von `build/ohne-trimpath`) geht an den Planner und wird ein eigener Slice.
- Eine Ausgabe der Fallzeiten oder eine Aufteilung der Fälle auf Stufen — beide erst nach einem Befund nach `SPEC-102` *Zeitbudget* Nr. 5 b oder c.
- Eine Entscheidung zu `Zeitbudget` Nr. 4 und 5 d gegenüber [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) und zum Sperrpfad in Nr. 1 der Übergangsregeln — Entscheidungen des Users (§6 von `slice-gegenproben-zeitbudget`), kein Gegenstand einer Messung.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] Ein gültiger Lauf von `time make gegenproben` ist gemessen (beide Lesungen höchstens `nproc --all`, Messender genannt, Wanduhr, Zahl der Fälle) und steht nach `SPEC-102` *Zeitbudget* Nr. 6 in der Reihe und als Absatz *Eichung* nach Nr. 7, samt Ausgang nach Nr. 5 a bis c; die Lesung am Ende nennt den Eigenanteil des Laufs (Vergleich mit der Last zu Beginn); liegt sie bei ruhigem Host über der Kernzahl, ist das ein Befund an den Planner (Nr. 7).
- [ ] Die Fälle 205 bis 216 und `build/ohne-trimpath` sind im selben Lauf gültig grün oder als Befund mit Text und Last berichtet (Geber `slice-gates-ohne-worktrees`); ein Rot von `build/ohne-trimpath` ohne Last ist ein Befund, kein Eintrag in die Reihe.
- [ ] Der Stempel `harness/gegenproben.stempel` ist vom Integrationsschritt nach dem grünen Lauf allein committet (Nr. 2 der Übergangsregeln); der Slice-Zweig committet ihn nicht.
- [ ] Review und Verifikation durchgeführt, Reports unter `docs/reviews/` (die Eintragungen in `SPEC-102` lesen sie nach, `SPEC-102` *Grenze*: der Eintrag trägt beide Lesungen).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; Beobachtungs-Register (`../observations/`) fortgeschrieben (`BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last`); jedes Risiko aus §6 trägt einen Ausgang; die drei Paarungen sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/SPEC-102-gegenproben.md` | update | Zeile der Messreihe (Nr. 6), Absatz *Eichung* (Nr. 7), *Stand*; geschrieben nach dem Lauf |
| `harness/gegenproben.stempel` | Integrationsschritt | nur nach grünem Lauf, allein im Commit (Nr. 2 der Übergangsregeln) |
| Code, Gates, Tests | keine Änderung | der Slice misst |

**Ablauf des Laufs:** Last (1 Minute) und `nproc --all` lesen, `flock /home/builder/work/odata-recorder/.wt/heavy.lock make gegenproben` unter `time`, Last am Ende lesen; ein ungültiger Lauf (Lesung über der Kernzahl) wird nach `SPEC-102` Nr. 4 behandelt (ein roter mit nur `Zeit überschritten` wird wiederholt, höchstens dreimal). Schwere Läufe der anderen Maschinen ruhen für die Dauer; der Planner prüft vorher mit `uptime`, dass die Last auf dem Host unter der Kernzahl liegt.

Die Tabelle *Zusage · Fänger* entfällt: der Slice liefert keinen Vertrag, er trägt einen Messwert in einen bestehenden ein; das Fehlen eines Fängers für die Vorschrift selbst steht als akzeptiertes Negativ in `slice-gegenproben-zeitbudget` §3.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `next` → `in-progress`): `slice-gegenproben-zeitbudget` liegt in `done/` und in `main`; der User hat einen Termin gesetzt, an dem w1, w2 und w3 schwere Läufe pausieren (§1); kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln).

**Größenmarke:** keine. Der Slice ändert keinen Code in den Pfaden eines Blocks; die Zeilen in `harness/vertraege/` zählen nach `SPEC-128` Nr. 4 a nicht. Ein Plan ohne Block wird von `make groessenmarke` nicht gemessen.

**Rückführungen — vorab benannt:**

- `in-progress` → `open` (blockiert): der Host wird während des Termins belastet und liefert keinen gültigen Lauf; dann wird der Termin neu gesetzt, der Slice verändert nichts.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (der gemessene Wert und sein Ausgang nach Nr. 5; bei Bruch die benannte Folge, bei Bestehen die Aussage, dass 120 s tragen).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Der Host liefert auch zum Termin keinen gültigen Lauf (Last durch fremde Prozesse, die der Slice nicht pausieren kann) — **Ausgang:** weiter offen; Adresse: User (Termin, Pausen auf w1, w2, w3).
- Die Lesung am Ende enthält die Last des Laufs selbst (neun parallele `golangci-lint`-Läufe, bis zu so viele Fälle wie Kerne) und kann auf ruhigem Host die Kernzahl erreichen — **Ausgang:** weiter offen; der Lauf misst den Eigenanteil (DoD); ein Ergebnis über der Kernzahl bei ruhigem Host setzt die Grenze am Ende neu (`SPEC-102` Nr. 7, Änderung von Nr. 3, Entscheidung des Architect).
- `build/ohne-trimpath` ist auch bei gültiger Last rot — **Ausgang:** weiter offen; Befund an den Planner, Folge-Slice mit Ursache (Geber `slice-gates-ohne-worktrees`).
- Der Lauf liegt über 120 s bei gültiger Last — **Ausgang:** weiter offen; Re-Evaluierungs-Trigger von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (Anhebung nur als neue ADR, drei gültige Läufe nach `SPEC-102` Nr. 5 b).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Closure von `slice-gegenproben-zeitbudget` (2026-11-18); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** das Register trägt `laufzeitmessung-widerspricht-sich-unter-last` (3× mit der Closure von `slice-gegenproben-zeitbudget`, Zuweisung beim User); die Eichung ist die Messung, die die Beobachtung bei gültiger Last auflöst oder bestätigt.

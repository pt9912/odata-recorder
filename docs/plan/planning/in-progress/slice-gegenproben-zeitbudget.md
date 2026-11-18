# Slice slice-gegenproben-zeitbudget: Messvorschrift für das Zeitbudget der Gegenproben

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): berührt `tools/gegenproben/` und den Stempel, Nr. 7 (ein Besitzer der Gate-Kette zugleich).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Laufzeit der Gegenproben, der Bezug ist der nächste, den es gibt), [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (Entscheidung, `Accepted`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Re-Evaluierungs-Trigger „über 30 s“, dessen Ausgang [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ist), [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Messung durch den Host).

**Berührte Spec-Stellen:** `SPEC-102` (Randform 2a und 6, Absatz *Zeitbudget* mit der Messvorschrift Nr. 1 bis 7, *Stand* und *Grenze*); keine neue Nummer.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Auftrag des Users nach `Accepted` von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)). **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Das Soll „höchstens 120 s“ für `time make gegenproben` (geänderte Eingabe, kalter Cache) hat in `SPEC-102` eine Messvorschrift: Es gilt nur für einen Lauf bei Load Average (1 Minute) höchstens der Kernzahl, zu Beginn und am Ende; ein Lauf darüber ist *ungültig*, kein Bruch, und wird, wenn er rot ist, wiederholt. Die Frist je Fall (`fallFrist`, 120 s) bleibt unverändert die einzige maschinell geprüfte Schwelle.

**Hintergrund:** [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) wählt Option B. Die Einzelheiten (Messgröße, Lastgrenze, Zahl der Wiederholungen, Ausgang bei Bruch) stehen laut ADR in `SPEC-102`: dort als *Zeitbudget* Nr. 1 bis 7, in der Form von `SPEC-211` Nr. 4. Die Entscheidungen des Architect stehen in §3.

**Beleg aus `slice-gates-ohne-worktrees`:** Der Lauf der Stufe `gegenproben` mit den Fällen 205 bis 216 färbte unter Last (Load Average 16 und 20 zu Beginn, 175 bis 222 am Ende) neun Fälle `lint/*` mit „Zeit überschritten“ (`SPEC-102` Randform 6); die Fälle 205 bis 216 selbst brauchen kein Produkt und dauern 0,4 bis 0,84 s. Das Ziel dieses Slice (ein Lauf über der Lastgrenze ist ungültig) deckt diese neun Fälle; der Beleg steht in `SPEC-102` (Messreihen). Die Messung auf ruhigem Host und die Ursache von `build/ohne-trimpath` stehen als Folgepflicht in §6.

**Umfang:**

1. `SPEC-102`: Messvorschrift Nr. 1 bis 7 im Absatz *Zeitbudget*, Randform 2a und 6 (Frist je Prozess, Bau über der Frist), *Stand* und *Grenze*; der Stand-Satz „Das Budget liegt rund 28 % über …“ entfällt, die Messreihen bleiben als Belege. Geschrieben vom Architect, vom Implementer nur bei einer Abweichung des Codes zu berichtigen.
2. `tools/gegenproben`: **kein** Programmverhalten für Last oder Fallzeiten (Entscheidung in §3). Geliefert wird allein, was der Vertrag zusagt: `bewerteBau` (ein Bau über der Frist meldet `Zeit überschritten` statt `Bau des Binary scheitert`) und Tests, die `fallFrist`, ihre Aufrufstellen und die Weitergabe des Ablaufs in `fuehreFall` an die Wertung halten.
3. Keine neuen Fälle der Gegenproben (die Fälle prüfen Gates; die Zusagen dieses Slice liegen im Programm und sind Tests des Programms), kein Stempel im Zweig (Nr. 2 der Übergangsregeln).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Anhebung der Frist je Fall oder des Soll von 120 s — Lockerung (`AGENTS.md` §3.6), von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ausgeschlossen; eine Anhebung ist eine eigene ADR mit gültigen Messwerten (Re-Evaluierungs-Trigger der ADR).
- Ein Prüfer, Gate oder Hook für das Gesamtbudget und eine Messung der Last durch das Programm — [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) benennt das Soll ausdrücklich als nicht maschinell geprüft; Gründe der Entscheidung in §3.
- Eine Ausgabe der Fallzeiten — nicht nötig, solange kein gültiger Lauf über 120 s liegt; verlangt ein Befund nach `SPEC-102` *Zeitbudget* Nr. 5 b oder c die Aufteilung der Zeit, legt der Planner sie als eigenen Slice an.
- Eine Aufteilung der Fälle auf Stufen (Option C der ADR) — eigener Slice, nur nach einer Messung, dass die Fälle und nicht die Last die Zeit bestimmen.
- Die Eichung auf ruhigem Host — Folgepflicht, kein Liefer-Punkt: sie braucht einen Zeitpunkt, an dem andere Läufe von `make` pausieren; erst messen, dann Zahl (§6, Ausgang *weiter offen*, Adresse: Planner; Form in `SPEC-102` *Zeitbudget* Nr. 7). Dazu gehört die Messung `time make gegenproben` mit den Fällen 205 bis 216 und die Wiederholung von `build/ohne-trimpath` (Geber `slice-gates-ohne-worktrees`, §6).
- Änderungen am Format der Messreihen der Fälle außer dem Kopf der Vorschrift.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [ ] `SPEC-102` trägt die Messvorschrift (Gültigkeit eines Laufs, Lastgrenze, Wiederholung, Ausgang bei Bruch, Frist je Prozess) und benennt das Soll als nicht maschinell geprüft; die Frist je Fall ist unverändert und von `TestFallFristIst120Sekunden`, `TestJederAufrufNenntDieFallFrist` und `TestFuehreFallReichtDenAblaufWeiter` gehalten.
- [ ] Jede Zeile der Tabelle *Zusage · Fänger* (§3) hat eine vom Implementer rot gesehene Mutation; `bewerteBau` und die Tests der Tabelle liegen in `tools/gegenproben`.
- [ ] Doku-Update: `SPEC-102` *Stand* und *Grenze* (geschrieben); `harness/README.md` bleibt unverändert, die Beschreibung von `make gegenproben` ändert sich nicht; `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect entscheidet (unten), der Implementer erweitert diese Liste in seinem ersten Lauf.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/SPEC-102-gegenproben.md` | update, vom Architect geschrieben | Randform 2a und 6, *Zeitbudget* Nr. 1 bis 7, *Stand*, *Grenze*; Stand-Satz entfällt |
| `tools/gegenproben/main.go` | update | `bewerteBau`, von `fuehreFall` gerufen; `fallFrist` bleibt 120 s |
| `tools/gegenproben/main_test.go` | update | `TestBewerteBau`, `TestFuehreFallBauScheitertRuftKeinGate` |
| `tools/gegenproben/frist_test.go` | neu | `TestFallFristIst120Sekunden`, `TestJederAufrufNenntDieFallFrist`, `TestFuehreFallReichtDenAblaufWeiter` |
| `harness/gegenproben.stempel` | kein Eintrag im Zweig | Nr. 2 der Übergangsregeln; der Integrationsschritt fährt `make gegenproben` und committet den Stempel allein (Eingabe `tools/gegenproben/` ändert sich) |

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Vertrag, Entscheidungen, Tabelle *Zusage · Fänger*, Raster, erwartete Produktcode-Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure).

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock /home/builder/work/odata-recorder/.wt/heavy.lock` (die Sperre der Gruppe der Maschinen; Nr. 1 der Übergangsregeln nennt noch `/tmp/odata-recorder-heavy.lock`, §6); Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Entscheidungen des Architect** (Aufträge 1 bis 5, Stand vor dem ersten Code-Commit; 6 nach dem Review):

1. **Entwurf geprüft.** Ein Entwurf der Messvorschrift liegt im Repo nicht vor (gesucht in `docs/`, `harness/` und im Scratchpad; vorhanden ist die Form von `SPEC-211` Nr. 4); die Vorschrift ist aus §1 und [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) geschrieben und gegen den Code gelesen (`tools/gegenproben/main.go`). Gegen den Ansatz in §1 geändert: (a) die Kernzahl ist `nproc --all`, nicht `nproc`: in einer Umgebung mit Quote zählt `nproc` die Kerne des Prozesses (hier 3, der Host hat 20); (b) ein Container mit `--cpus=4` ändert weder `/proc/loadavg` noch `nproc --all` (belegt am 2026-11-18), die Messung ist die des Hosts; (c) die Lesung am Ende enthält die Last des Laufs selbst und kann die Kernzahl auf ruhigem Host erreichen, der Fehler geht nur Richtung *ungültig* (Wiederholung), nie Richtung Befund, und die Eichung prüft ihn; (d) eine Lesung, die der Eintrag nicht nennt, entschuldigt nichts (der Lauf gilt als gültig); (e) die Wiederholung endet nach drei ungültigen Läufen in Folge mit einem Befund an den Planner; (f) Last entschuldigt nur den Text `Zeit überschritten`, nie einen anderen Grund.
2. **Programmverhalten: keines für Last und Fallzeiten, die Messvorschrift ist Prosa.** Gründe: (a) das Soll misst `time make gegenproben`, also auch den Bau des Images und seinen Export (7 bis 36 s), die das Programm nicht sieht; ein Urteil des Programms über die Gültigkeit stünde neben dem des Messenden und könnte ihm widersprechen; (b) jede Zusage des Programms braucht einen Fall je Zeile, einen Ersatz für `/proc/loadavg`, die Randformen *nicht lesbar*, *gleich der Kernzahl* und *Kernzahl im Container*, und macht den Host zur Eingabe des Programms; (c) der Stempel hängt am grünen Lauf, und ein grüner Lauf unter Last ist der stärkere Beleg, nicht der schwächere: ein *ungültiger grüner* Lauf bleibt grün und schreibt den Stempel wie jeder grüne Lauf (`SPEC-102` *Zeitbudget* Nr. 5 d), ein roter schreibt ihn nie; (d) die Fälle brauchen einzeln Hundertstel- bis Zehntelsekunden, die Zeit bestimmt der Host, eine Ausgabe der Fallzeiten trennt Last und Aufwand erst, wenn ein gültiger Lauf über 120 s liegt (dann Befund nach Nr. 5 b oder c, Folge-Slice des Planners). Die Rückführung `in-progress` → `open` (§4) tritt nicht ein. Was das Programm doch ändert: `bewerteBau`, weil Randform 6 „Zeit überschritten“ zusagt und der Bau nach altem Code `Bau des Binary scheitert` meldete, ohne Zeit und Fehler zu trennen (`build/ohne-trimpath`, §6).
3. **Randformen** vor dem Code im Vertrag: Raster in §6, Orte in `SPEC-102` Randform 2a und 6 und *Zeitbudget* Nr. 1 bis 7. Die Frist gilt je Prozess (Bau, Gate, Treiber, `git`), nicht je Fall; der Vertrag beschreibt es jetzt so, das Verhalten ist unverändert.
4. **Tabelle und Größe** unten; **ADR:** keine (die Entscheidung trägt [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md), die Einzelheiten stehen laut ADR in `SPEC-102`); keine ADR-Nummer aus 0030 bis 0039 und keine neue `SPEC`-Nummer vergeben.
5. **Eichung:** `SPEC-102` *Zeitbudget* Nr. 7 (Absatz *Eichung* hinter den Messreihen, Eintrag nach Nr. 6); der Slice führt sie nicht aus (§1, §6).
6. **Review F-1 bis F-6 (Fix-Runde).** (a) F-1: die Zusage „`fuehreFall` reicht den Ablauf weiter“ hat die Menge der Stellen, an denen `fuehreFall` das Ergebnis eines Prozesses wertet: 2 (Bau, Gate); ein Fänger über den Quelltext braucht keine 120 s, das Negativ entfällt, die Zeilen 8 und 9 der Tabelle treten an seine Stelle. Die übrigen 5 Aufrufstellen binden den Ablauf nicht an eine Wertung: 4 verwerfen ihn nach Randform 6 (`faelle.go` 3, `zeichen_treiber.go` 1), und `gmRepo.git` wertet `abgelaufen || code != 0`, wo ein abgelaufener Prozess ohnehin Exit -1 trägt, die Mutation also ohne Wirkung bleibt (Negativ unten). (b) F-2: die Menge der Formen, an denen die Aufrufstellen-Prüfung vorbeigeht, ist unbeschränkt (Funktionswert, überdecktes `fallFrist`, Prozess ohne `starte`: 3 genannte Formen, weitere möglich); ein Fänger je Form wäre eine Liste der Formen, die einfielen. Entschieden: benanntes Negativ mit Grund und Folge, und die Zusage nennt nur, was der Fänger fängt (Zeile 2). (c) F-3: *Zeitbudget* Nr. 1 nennt den Cache der Stufe (`SPEC-102`). (d) F-6: *Zeitbudget* Nr. 4 und 5 d haben eine Lesart: ein ungültiger grüner Lauf wird nicht wiederholt, zählt nicht als Messung der Zeit und schreibt den Stempel; drei ungültige grüne Läufe in Folge sind kein Befund, weil nichts rot ist und die Eichung (Nr. 7) den ersten gültigen Lauf ohnehin liefert; ein Befund ohne Rot wäre die Hürde, die [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) vermeidet. Nur ein ungültiger roter Lauf (Nr. 5 e) wird wiederholt, und nur drei davon in Folge gehen als Befund an den Planner. (e) F-5: Plan und Betriebsregeln nennen die Sperre `/home/builder/work/odata-recorder/.wt/heavy.lock`; die Übergangsregeln ändert der Architect nicht (§6).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt Raten und Läufe nach `AGENTS.md` §3.10 ein): jede Zeile eine Mutation, alle Klasse `andere` (keine Zusage über Nebenläufigkeit).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | `fallFrist` hat den Wert 120 s (Randform 6, Fitness Function von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)) | `TestFallFristIst120Sekunden`: Gleichheit mit `120 * time.Second` (fängt 119 s und 121 s) | `fallFrist = 121 * time.Second` | andere |
| 2 | Jeder Aufruf, dessen Funktion als Bezeichner `starte` oder `starteStroeme` geschrieben ist, nennt `fallFrist` als erstes Argument (Allaussage über die Aufrufstellen, am Stand 7: `main.go` 2, `faelle.go` 3, `groessenmarke_treiber.go` 1, `zeichen_treiber.go` 1); ein Funktionswert, ein überdecktes `fallFrist` und ein Prozess ohne `starte` gehören nicht dazu (Negativ unten) | `TestJederAufrufNenntDieFallFrist`: Schleife über die Quelldateien des Pakets ohne `_test.go`, findet mindestens 7 Aufrufstellen (eine Prüfung, die nichts findet, ist rot) | in `zeichen_treiber.go` den Aufruf mit `10 * time.Minute` statt `fallFrist` (ein Element, das ein Befund nicht nennen muss) | andere |
| 3 | Ein Bau, der die Frist reißt, ist rot mit `Zeit überschritten`, nicht mit `Bau des Binary scheitert` (Randform 6 und 2a) | `TestBewerteBau`, Zeile Ablauf (Code -1, abgelaufen) | der Zweig `abgelaufen` in `bewerteBau` entfällt | andere |
| 4 | Ein Bau mit Exit ungleich 0 ist rot mit `Bau des Binary scheitert` und den letzten Zeilen der Ausgabe (Randform 2a) | `TestBewerteBau`, Zeile Exit 1 mit Ausgabe aus mehreren Zeilen (Text und letzte Zeile genannt, erste Zeile nicht) | `letzte(out)` fällt aus dem Text | andere |
| 5 | Ein Bau mit Exit 0 ist nicht rot (beide Seiten der Grenze: Exit 0 und Exit 1) | `TestBewerteBau`, Zeile Exit 0 | `code != 0` wird `code >= 0` | andere |
| 6 | Scheitert der Bau, wird das Gate nicht gerufen (Randform 2a) | `TestFuehreFallBauScheitertRuftKeinGate`: Ersatz für `Bauen` endet mit Exit 1, der Ersatz für `Befehl` zählt jeden Aufruf; erwartet: Rot beginnt mit `Bau des Binary scheitert`, Zahl der Aufrufe 0 (gibt es schon einen Test mit diesem Gegenstand, ist er der Fänger) | `fuehreFall` kehrt nach dem Rot des Baus nicht zurück | andere |
| 7 | Ein Gate-Aufruf über der Frist ist rot mit `Zeit überschritten`; die Prozessgruppe endet (Randform 6, unverändert) | Zeile `Frist` in `TestBewerte` und der Test der Hülle in `strom_test.go` (je `getrennt` wahr und falsch) | `case abgelaufen` in `bewerte` entfällt | andere |
| 8 | `fuehreFall` reicht den Ablauf des Baus an die Wertung weiter (Randform 6): Menge der Stellen, an denen `fuehreFall` das Ergebnis eines Prozesses wertet, 2 (Bau, Gate); Element Bau | `TestFuehreFallReichtDenAblaufWeiter`, Fall Bau: Schleife über die zwei Paare (Aufruf · Wertung) im Rumpf von `fuehreFall` aus `go/parser`; je Paar bindet genau eine Zuweisung das Ergebnis von `starte` bzw. `starteStroeme` und sein letzter Name ist kein `_`, die Wertung (`bewerteBau` bzw. `bewerte`) kommt genau einmal vor, ihr letztes Argument ist ein Bezeichner mit diesem Namen, und der Name steht je Paar nur an diesen zwei Stellen; `fuehreFall` ruft `starte` und `starteStroeme` genau je einmal (eine dritte Wertung macht den Test rot, bis ihr Paar benannt ist); rot auch bei Umbau auf eine Zwischenvariable oder Hilfsfunktion, der Umbau trägt dann die Änderung des Tests | `bewerteBau(code, out, abgelaufen && false)` in `fuehreFall` | andere |
| 9 | wie 8, Element Gate | `TestFuehreFallReichtDenAblaufWeiter`, Fall Gate (dieselbe Schleife, zweites Element) | `bewerte(f, code, out, errOut, abgelaufen && false)` in `fuehreFall` | andere |

**Akzeptierte Negative** (Grund und Folge, `AGENTS.md` §3.15):

| Negativ | Grund | Folge |
|---|---|---|
| *Zeitbudget* Nr. 1 bis 7 (Gültigkeit, Wiederholung, Ausgang, Eintrag) hat keinen Fänger | der Messende liest die Last und die Wanduhr selbst; das Programm sieht Bau und Export des Images nicht (Entscheidung 2) | ein Lauf kann falsch als gültig oder ungültig geführt werden; der Eintrag trägt beide Lesungen, Reviewer und Verifier lesen sie nach, und eine Lesung, die fehlt, entschuldigt nichts (Nr. 3) |
| Aufrufe von `starte` und `starteStroeme` über einen Funktionswert, ein im Rumpf überdecktes `fallFrist` und ein Prozess, der nicht über `starte` oder `starteStroeme` startet (Zeile 2, 3 genannte Formen) | die Menge der Formen ist unbeschränkt (jede Go-Form, die einen Prozess startet oder einen Namen verdeckt); ein Fänger je Form wäre eine Liste der Formen, die einfielen, und jede verlangt eine Absicht gegen das Muster, die der Diff zeigt | ein solcher Prozess hat keine Frist von 120 s und hängt bis zum Ende des Aufrufs (Container, `make`); kein Sensor meldet es |
| Die Stellen, an denen ein Aufruf von `starte` oder `starteStroeme` den Ablauf nicht an die Wertung von `fuehreFall` weitergibt: 4 verwerfen ihn (`faelle.go` 3, `zeichen_treiber.go` 1; Randform 6), `gmRepo.git` in `groessenmarke_treiber.go` wertet `abgelaufen \|\| code != 0` | ein abgelaufener Prozess endet mit Exit -1; `git … scheitert` mit denselben letzten Zeilen kommt mit und ohne `abgelaufen \|\|`, die Mutation hat keine Wirkung; die vier Verwerfer sind die Entscheidung von Randform 6 | ein Treiber, der den Ablauf künftig meldet, ändert nur den Text des roten Falls; Last entschuldigt ihn nach *Zeitbudget* Nr. 5 f weiterhin nicht |
| Die Lesung am Ende enthält die Last des Laufs selbst | der Lauf startet bis zu so viele Fälle wie Kerne; die Größe des Eigenanteils ist nicht gemessen | ein gültiger roter Lauf kann als ungültig zählen und wird wiederholt, nie als Befund; drei ungültige rote Läufe in Folge gehen an den Planner, die Eichung prüft den Eigenanteil (Nr. 7) |
| Eine kurze Last-Spitze zwischen den Lesungen bleibt unsichtbar | das Mittel über 1 Minute am Ende deckt die letzten 60 s eines Laufs, der meist länger als 60 s dauert; mehr Lesungen verlangt kein Beleg | ein Fall mit `Zeit überschritten` kann als Befund erscheinen, den ein Lauf auf ruhigem Host entkräftet; der Planner wiederholt dort |
| Treiber-Aufruf und `git` in der Nachprüfung über der Frist melden nicht `Zeit überschritten` | der Aufruf wirft den Ablauf weg (`code, out, _ := starte(…)`); die Nachprüfung wertet den Exit -1 | ein solcher Fall zählt nach Nr. 5 f als Befund statt als durch Last entschuldigt, also Richtung Planner und nie Richtung Wiederholung |

**Nachweis** (`AGENTS.md` §3.10; Messender: der Implementer, Host mit 20 Kernen, Load Average 1 Minute je Lauf aus `/proc/loadavg` zu Beginn des Laufs; Lauf je Mutation: `go vet` und `go test -count=1 ./...` in einem Container aus dem Image `odata-recorder:race` ohne Netz, `--memory=4g --cpus=4`, Quelle per `docker cp`; alle Klasse *andere*, kein Fänger wartet auf Wahrscheinlichkeit, je Mutation ein Prozess, rot in 1 von 1; der Lauf ohne Mutation ist grün). Die Menge von Zeile 2 ist *jeder Aufruf, dessen Funktion als Bezeichner `starte` oder `starteStroeme` geschrieben ist, in den Quelldateien ohne `_test.go`*: 7 Stellen (`main.go` 2, `faelle.go` 3, `groessenmarke_treiber.go` 1, `zeichen_treiber.go` 1), aus dem Quelltext mit `go/parser` gebildet; die Weitergabe der Frist im Rumpf von `starte` zählt nicht. Die Menge der Zeilen 8 und 9 ist *die Stellen, an denen `fuehreFall` das Ergebnis eines Prozesses wertet*: 2 (Bau, Gate).

| Nr. | Zusage | Mutation | roter Test, aus dem Grund | Last |
|---|---|---|---|---|
| 1 | `fallFrist` 120 s | `fallFrist = 121 * time.Second` | `TestFallFristIst120Sekunden`: `fallFrist = 2m1s, erwartet 120s` | 13,4 |
| 2 | jeder Aufruf nennt `fallFrist` | `zeichen_treiber.go:119` mit `600000000000` statt `fallFrist` | `TestJederAufrufNenntDieFallFrist`: `zeichen_treiber.go:119:25: … nennt fallFrist nicht` | 12,7 |
| 2 (Element, das der Plan nicht nannte) | wie 2 | eine achte Stelle in `main.go`: `starte(time.Second, "", nil, "true")` | `TestJederAufrufNenntDieFallFrist`: `main.go:558:16: Aufruf von starte nennt fallFrist nicht` | 12,0 |
| 3 | Bau über der Frist: `Zeit überschritten` | Zweig `abgelaufen` in `bewerteBau` entfällt | `TestBewerteBau/Ablauf` und `Ablauf_mit_Exit_0`: nennt `Zeit überschritten` nicht | 11,4 |
| 3 (Reihenfolge) | wie 3, Ablauf geht vor Exit | Zweig `code != 0` vor `abgelaufen` | `TestBewerteBau/Ablauf`: `Bau des Binary scheitert\nz1` statt `Zeit überschritten` | 14,6 |
| 4 | Exit ungleich 0: `Bau des Binary scheitert` mit den letzten Zeilen | `+ letzte(out)` fällt aus dem Text | `TestBewerteBau` (Exit 1, Exit 2, Exit -1 ohne Ablauf), `TestFuehreFallBauScheitertRuftKeinGate`, `TestFallMitBauMeldetScheiternDesBaus` | 11,4 |
| 5 | Exit 0 nicht rot, beide Seiten | `code != 0` wird `code >= 0` | `TestBewerteBau/Exit_0` (erwartet grün), `…/nicht_startbar` (Exit -1), `TestFallMitBauBautVorDemGate` | 11,3 |
| 5 (Element, das der Plan nicht nannte) | wie 5, Exit -1 ohne Ablauf (Programm nicht startbar) | `code != 0` wird `code > 0` | `TestBewerteBau/nicht_startbar`: Grund leer | 11,5 |
| 6 | scheitert der Bau, kein Gate | `fuehreFall` kehrt nach dem Rot des Baus nicht zurück | `TestFuehreFallBauScheitertRuftKeinGate`: `Rot = ""`, `Befehl 1-mal … gerufen, erwartet 0`; `TestFallMitBauMeldetScheiternDesBaus` | 11,4 |
| 7 | Gate über der Frist: `Zeit überschritten` | `case abgelaufen` in `bewerte` entfällt | `TestBewerte/Frist`: `erwartet Exit 1, war -1` statt `Zeit überschritten` | 12,4 |
| 8 | `fuehreFall` reicht den Ablauf des Baus weiter | `bewerteBau(code, out, abgelaufen && false)` | `TestFuehreFallReichtDenAblaufWeiter/Bau`: `das letzte Argument von bewerteBau ist nicht der Bezeichner abgelaufen` | 11,4 |
| 9 | wie 8, Gate | `bewerte(f, code, out, errOut, abgelaufen && false)` | `TestFuehreFallReichtDenAblaufWeiter/Gate`: `das letzte Argument von bewerte ist nicht der Bezeichner abgelaufen` | 11,2 |
| 8 (Element, das der Review nicht nannte) | wie 8 | Zuweisung `abgelaufen = false` vor `bewerteBau` | `…/Bau` und `…/Gate`: `abgelaufen steht 5-mal in fuehreFall, erwartet 4` | 9,2 |
| 9 (Element, das der Review nicht nannte) | wie 9 | Zuweisung `abgelaufen = false` vor `bewerte` | `…/Bau` und `…/Gate`: `abgelaufen steht 5-mal in fuehreFall, erwartet 4` | 8,7 |
| 8 (Zwischenvariable) | wie 8 | `weiter := abgelaufen` und `bewerteBau(code, out, weiter)` | `…/Bau`: letztes Argument ist nicht `abgelaufen` (rot, der Umbau trägt die Änderung des Tests) | 8,6 |
| 8 (Bindung) | Bau bindet den Ablauf | `code, out, _ := starte(…)` und `abgelaufen := false` | `…/Bau`: `der Aufruf von starte bindet den Ablauf an "_"`; `…/Gate`: `abgelaufen steht 4-mal in fuehreFall, erwartet 2` | 9,8 |

Zeilen 8 und 9 (Review F-1): die Menge ist *die Stellen, an denen `fuehreFall` das Ergebnis eines Prozesses wertet*, 2 Elemente (Bau, Gate); der Fänger läuft als Schleife über beide Paare, beide Elemente sind einzeln rot gesehen, dazu zwei Mutationen an einem Element, das der Review nicht nannte (die Zuweisung vor der Wertung, auf beiden Seiten). Die Prüfgrenze der Zeile 2 (Review F-2) steht als akzeptiertes Negativ in der Tabelle oben und in `SPEC-102` *Grenze*; der Kopf von `TestJederAufrufNenntDieFallFrist` sagt nur zu, was der Test fängt. Der Lauf der Zeilen 8 und 9 wie oben (Messender: der Implementer, je 1 von 1 Prozess, Last 1 Minute aus `/proc/loadavg` zu Beginn des Laufs, Host mit 20 Kernen): rot in 1 von 1 je Mutation, die Kopie ohne Mutation grün (Last 11,9); `TestFallFristIst120Sekunden`, `TestJederAufrufNenntDieFallFrist` und die übrigen Tests des Pakets bleiben im Lauf ohne Mutation grün (`go vet` und `gofmt -l` ohne Meldung). Im Review waren die Mutationen der Zeilen 8 und 9 vor dem Fänger grün (Messender: der Reviewer, je 1 von 1 Prozess, Last 12,6 und 13,7) und ein Probe-Fänger rot.

Gegengelesen auf Rohbytes (`LC_ALL=C grep`, `$'\xc3\xbc'`): `tools/gegenproben/main.go`, `main_test.go`, `frist_test.go` tragen `ü` als zwei Bytes und keine Folge aus Backslash, `u` und Hex-Ziffern. Die Mutationen liefen auf Kopien im Scratchpad, nicht im Zweig. Die Fix-Runde (Zeilen 8 und 9) las `frist_test.go` auf Rohbytes zurück (`LC_ALL=C grep -c $'\xc3\xbc'`: 4 Zeilen mit `ü`, `ö` und `ß` als Bytes `\xc3\xb6` und `\xc3\x9f`, `ä` als `\xc3\xa4`; kein BOM, kein NBSP, keine Folge aus Backslash, `u` und Hex-Ziffern). Messung nach `SPEC-128` Nr. 4 nach der Fix-Runde (`git diff --numstat a8c40949dd557c67561f980fb4ec94878b927e48..HEAD` über die Pfade des Blocks): 271 Zeilen (`main.go` 17, `main_test.go` 79, `frist_test.go` 175; die ausgeschlossenen Pfade des Blocks 0) gegen die Marke 1400, Auslastung rund 0,19 (am Handoff vor den Zeilen 8 und 9: 167 Zeilen).

**Erwartete Größe** für die Marke des Planners (Messung nach `SPEC-128`, Pfade dieses Plans, Tests eingeschlossen): Produktcode Go rund 15 Zeilen (`main.go`: `bewerteBau` mit drei Zweigen und sein Aufruf in `fuehreFall`), Tests Go rund 100 Zeilen (`main_test.go` rund 50, `frist_test.go` rund 50), ein Fake-Paar (`Bauen`, `Befehl`) in Zeile 6 von je rund 5 Zeilen; Vertrag `SPEC-102` 23 Zeilen (16 neu, 7 ersetzt, gemessen mit `git diff --numstat` am Stand des Architect), Plan eigene Zeilen. Die Tabelle hat 9 Zeilen *andere*, 0 *nebenläufig*; erwartet ist ein kleiner Slice weit unter der Schnittgrenze von `SPEC-128` Nr. 1. Der Planner hat nach der Formel mit 7 Zeilen gerechnet: *S* 975, Marke 1400, keine Schnittfrage (Block und Rechnung in §4); die erwartete Messung liegt bei rund 125 Zeilen (15 Produktcode, 100 Tests, 10 Fakes), mit den Zeilen 8 und 9 bei rund 250, also weit unter der Marke.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln: ein Besitzer zugleich). Reihenfolge der Kette seit 2026-11-08: `slice-gate-integration-determinism`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu`, dann dieser Slice; der Start legt fest, ob er vor `slice-groessenformel-neu` rückt, solange dessen Beschluss des Users aussteht. Höchstens drei Slice-Worktrees zugleich. [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ist `Accepted`.

**Größenmarke:** Marke 1400 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nicht, `SPEC-128` Nr. 4 a), vom Planner vor dem ersten Code-Commit nach `SPEC-128` Nr. 1 gesetzt.
Rechnung: 7 Zeilen *andere*, 0 *nebenläufig* (Tabelle in §3), `fakes` 2 (das Paar `Bauen` und `Befehl` der Zeile 6, vom Architect als Fake-Paar genannt; der Faktor 200 ist die Obergrenze der Messung, die Fakes hier sind Funktionswerte von rund 5 Zeilen), `produktcode` 15 (Mitte der Erwartung „rund 15“ des Architect: `bewerteBau` mit drei Zweigen und sein Aufruf in `fuehreFall`). *S* = 80·7 + 200·2 + 15 = 975, *M* = 100·⌈135·975/10000⌉ = 1400. Schnittfrage: nein (135·975 = 131625, nicht über 300000).
Die Pfade des Plans sind die drei Dateien von `tools/gegenproben/`; dazu stehen die Pfade im Block, die dieser Plan ausschließt (§1, §3: `harness/README.md`, `Dockerfile`, `.a-check.yml`, `.d-check.yml`, `.golangci.yml`, `harness/mk/`, `tools/harness/`), damit eine Änderung dort mitzählt statt unbemerkt zu bleiben.
Der Block trägt keinen Eintrag `ersetzt` (eine ältere Marke stand nicht im Plan, nur Prosa ohne Zahl).

```groessenmarke
pfad: tools/gegenproben/main.go
pfad: tools/gegenproben/main_test.go
pfad: tools/gegenproben/frist_test.go
pfad: harness/README.md
pfad: Dockerfile
pfad: .a-check.yml
pfad: .d-check.yml
pfad: .golangci.yml
pfad: harness/mk/
pfad: tools/harness/
produktcode: 15
fakes: 2
basis: a8c40949dd557c67561f980fb4ec94878b927e48
schaetzung: 975
marke: 1400
```

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze.
- `in-progress` → `open` (blockiert): ein Slice der Gate-Kette hält Nr. 7. (Der frühere Grund, die Last des Hosts im Programm zu lesen, ist mit der Entscheidung in §3 entfallen: das Programm liest sie nicht.)

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (Messvorschrift als geschärfte Regel in `SPEC-102` mit Anker, oder benannte Spec-Lücke).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die Eichung auf ruhigem Host steht aus; ohne sie bleibt unbekannt, ob 120 s tragen und wie groß der Eigenanteil des Laufs an der Lesung am Ende ist — **Ausgang:** weiter offen; Adresse: Planner (Termin bei ruhigem Host, andere Läufe von `make` pausiert; Form `SPEC-102` *Zeitbudget* Nr. 7; liegt der gültige Lauf über 120 s, gilt der Re-Evaluierungs-Trigger von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)).
- Folgepflicht aus `slice-gates-ohne-worktrees` (Geber, Plan §6; die Eichung ist hier ausgeschlossen, §1): (a) `time make gegenproben` auf ruhigem Host mit den Fällen 205 bis 216 ist nie gemessen (der Lauf schreibt den Stempel, Übergangsregeln Nr. 2, also im Integrationsschritt) — **Ausgang:** weiter offen; Adresse: Planner (mit der Eichung der ersten Zeile dieser Liste); (b) die neun Fälle `lint/*` mit „Zeit überschritten“ unter Last 16 bis 222 sind Beleg für das Ziel dieses Slice, kein Rot der Fälle — **Ausgang:** entfallen; der Beleg steht in `SPEC-102` (Messreihen, Fälle 205 bis 216) und der Lauf ist nach Nr. 3 ungültig; (c) `build/ohne-trimpath` („Bau des Binary scheitert“) war im zweiten Lauf des Implementers bei Last 175 bis 222 rot, bei Last 10 zweimal grün; mit `bewerteBau` meldet ein Bau über der Frist künftig `Zeit überschritten`, ein Bau mit Exit ungleich 0 die letzten Zeilen von `go build`; die Ursache des alten Rot bleibt unbekannt, im Lauf der Eichung wird der Fall wiederholt, ein Rot dort ist ein Befund an den Planner — **Ausgang:** weiter offen; Adresse: Planner (Termin der Eichung); Register `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` (schreibt der Integrationsschritt, Nr. 4 der Übergangsregeln). Der Geber `slice-gates-ohne-worktrees` liegt in `done/`; der Nehmer ist der Planner, kein Slice, und trägt die Sendung hier und in §1 *Ausdrücklich NICHT*.
- Auf einem Host mit Dauerlast gibt es keinen gültigen Lauf (Negativ der ADR) — **Ausgang:** weiter offen; Folge: Meldungen nennen den Lauf als ungültig statt als Bruch, nach drei ungültigen roten Läufen in Folge geht der Befund an den Planner (`SPEC-102` *Zeitbudget* Nr. 4; ungültige grüne Läufe sind kein Befund, sie bleiben grün und unbelegt bis zur Eichung), die Closures führen die Frage nicht mehr als offen.
- Das Programm wird um eine Last-Messung erweitert, die [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) nicht verlangt, und die Gegenproben bekommen eine Host-Abhängigkeit — **Ausgang:** entfallen; der Architect hat entschieden, dass das Programm weder Last noch Fallzeiten ausgibt (§3, Entscheidung 2).
- Größe: Marke 1400 bei erwarteter Messung von rund 125 Zeilen (Auslastung rund 0,09, unter 0,60): die Formel schätzt nach 80 je Zeile der Tabelle und 200 je Fake, was ein Slice mit sieben Zeilen für wenig Code und zwei Funktionswerte als Fakes überschätzt; Messung am Handoff des Implementers: 167 Zeilen, Auslastung rund 0,12 (Nachweis in §3) — **Ausgang:** weiter offen; Adresse: Planner (Closure: Messpunkt nach `SPEC-128` Nr. 9 eintragen; liegt die Auslastung am Ende unter 0,60, meldet der Planner sie dem Architect, der die Faktoren neu rechnet; keine Rückführung). Die Fix-Runde nach dem Review fügt der Tabelle zwei Zeilen an (8 und 9, 7 → 9 Zeilen *andere*) und dem Code 104 Zeilen Test (`frist_test.go` 71 → 175; Messung am Ende: 271 Zeilen, Auslastung rund 0,19 der Marke 1400): die Formel ergäbe *S* = 80·9 + 200·2 + 15 = 1135 und die Marke 1600; die Marke 1400 gilt weiter als Grenze der Messung (erwartet rund 250), und ob der Planner den Block in §4 neu rechnet (`schaetzung` 1135, `marke` 1600, mit Grund in §4), entscheidet er; `make groessenmarke` meldet bei Messung über 0 und Tabelle über der Schätzung einen Hinweis, kein Rot (`SPEC-102` Fall 48).
- Raster der Randformen — **Ausgang:** entfallen; gefüllt, siehe unten.
- Der Slice berührt `tools/gegenproben/` und den Stempel, kollidiert mit einem anderen Kettenslice — **Ausgang:** weiter offen: Nr. 7, einer zugleich, Start laut §4.
- Die Lesung am Ende enthält die Last des Laufs selbst; auf ruhigem Host kann sie die Kernzahl erreichen, und kein gültiger Lauf entsteht — **Ausgang:** weiter offen; Adresse: Planner (Eichung, Nr. 7: eine Messung am Ende über der Kernzahl bei ruhigem Host setzt die Grenze am Ende neu, per Änderung von Nr. 3).
- Die Vorschrift hat keinen Fänger außer der Prüfung durch Reviewer und Verifier (Negativ in §3) — **Ausgang:** weiter offen; Folge: ein Lauf kann falsch geführt werden, die Einträge nennen beide Lesungen.
- Die Sperre für schwere Läufe heißt in Nr. 1 der Übergangsregeln `/tmp/odata-recorder-heavy.lock`, die Gruppe der Maschinen führt `/home/builder/work/odata-recorder/.wt/heavy.lock` (Review F-5); dieser Plan nennt die zweite — **Ausgang:** offen; Adresse: Planner (Nr. 1 der Übergangsregeln ändert den Pfad, wenn der User es entscheidet; bis dahin sperrt, wer nur die erste nimmt, gegen niemanden, der die zweite nimmt).
- Die Zeile dieses Slice in der Roadmap trägt in der Klammer weiter `open/`, obwohl der Slice in `in-progress/` liegt (Review F-4, `AGENTS.md` §3.7: Zustandsfeld) — **Ausgang:** offen; Adresse: Planner (Closure; die Roadmap schreibt der Architect nicht).

**Raster der Randformen** (`AGENTS.md` §3.16): Operation × Argumentart, jede Zelle mit Ort in `SPEC-102` (Randform *R*, *Zeitbudget* Nr. *Z*) oder „gilt nicht“ mit Grund.

| Operation | fehlend oder nicht lesbar | Grenze (gleich, darüber, darunter) | Beginn und Ende verschieden | Abbruch | Wiederholung | Cache und Images |
|---|---|---|---|---|---|---|
| Messung des Aufrufs (Wanduhr) | Z 1: Eingaben unverändert, der Aufruf misst nur 1 s | Z 5 a und b: höchstens 120 s hält, darüber ist ein Bruch (gültiger Lauf) | gilt nicht: die Wanduhr ist ein Wert | Z 5: kein Ergebnis, kein Eintrag | Z 4 | Z 1: Images lokal; die Stufe `gegenproben` baut ohne Cache (`--no-cache-filter`), der Cache der übrigen Stufen liegt, wie er liegt, und gehört nicht zur Messgröße; ein Lauf, der Images zieht, zählt nicht |
| Lesung von Last und Kernzahl | Z 3: nicht lesbar oder nicht genannt, der Lauf gilt als gültig; Z 2: Kernzahl `nproc --all` | Z 3: gleich der Kernzahl ist gültig, darüber ungültig | Z 3: je Lesung allein; Beginn gültig und Ende ungültig ist ungültig, ebenso umgekehrt | Z 5: ohne Ergebnis | Z 4: jede Wiederholung liest neu | gilt nicht: die Last hängt nicht am Cache; Z 2: Container mit `--cpus` ändert die Lesung nicht |
| Wertung des Laufs | Z 3 | Z 3 | Z 3 | Z 5 | Z 4: nach drei ungültigen roten Läufen in Folge Befund an den Planner; ein ungültiger grüner Lauf wird nicht wiederholt | gilt nicht: die Wertung kennt den Cache nicht |
| Bau des Binary und Frist im Programm | R 2a: Exit ungleich 0 mit den letzten Zeilen; R 1: Eingabe fehlt | R 6: 120 s, Bau und Gate je Prozess; R 2a und 6 trennen Exit und Ablauf | gilt nicht: das Programm misst keine Last (R 6) | R 8: Wegwerfverzeichnisse und Container | R 7: ohne Zustand zwischen den Läufen | R 6: ein kalter `GOCACHE` verlängert den Bau, die Frist ist dieselbe |

Zusatztabelle (der Ausgang hängt vom Modus des Laufs ab, `AGENTS.md` §3.16): je Lauf eine Zeile.

| Lauf | Last | Ausgang | Ort | Fänger |
|---|---|---|---|---|
| grün, höchstens 120 s | gültig | das Soll hält | Z 5 a | kein Fänger (Negativ in §3) |
| grün, über 120 s | gültig | Bruch des Soll, Befund an den Planner, Lauf grün | Z 5 b | kein Fänger (Negativ in §3) |
| ein Fall `Zeit überschritten` | gültig | Befund an den Planner, nicht wiederholen bis grün | Z 5 c | Frist: Tabelle Zeilen 7 und 9; Wertung: Negativ in §3 |
| grün | ungültig | bleibt grün, nicht wiederholt, Zeit ohne Wertung | Z 5 d, Z 4 | kein Fänger (Negativ in §3) |
| rot, jeder rote Fall `Zeit überschritten` | ungültig | wiederholen | Z 5 e, Z 4 | kein Fänger (Negativ in §3) |
| rot, ein Fall mit anderem Grund | ungültig | Befund wie sonst | Z 5 f | die Fälle der Gegenproben melden ihren Grund; die Allaussage „Last entschuldigt nur `Zeit überschritten`“ hat kein Programm (Negativ in §3) |
| drei ungültige rote Läufe in Folge | ungültig | Befund an den Planner | Z 4 | kein Fänger (Negativ in §3) |
| drei ungültige grüne Läufe in Folge | ungültig | kein Befund, bleiben grün | Z 4 | kein Fänger (Negativ in §3) |
| Bau über der Frist | beliebig | rot mit `Zeit überschritten` | R 6 | Tabelle Zeilen 3 bis 5 und 8 |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus dem Beschluss des Users zu [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (2026-11-08); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** das Register trägt `laufzeitmessung-widerspricht-sich-unter-last` (offen, Laufzeitmessung unter Last); sie betrifft das Zeitbudget unmittelbar und ist mit *Zeitbudget* Nr. 1 bis 7 und der Eichung (§6) adressiert; andere Beobachtungen betreffen das Zeitbudget der Gegenproben nicht.

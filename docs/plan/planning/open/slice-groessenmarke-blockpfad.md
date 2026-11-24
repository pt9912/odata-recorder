# Slice slice-groessenmarke-blockpfad: `make groessenmarke` meldet ein Verzeichnis im Block, das ein anderer offener Plan nennt (Sensor zu `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-sensoren](../welle-sensoren.md) — Sensor zu einer Beobachtung des Registers, niedrige Priorität; der Planner trägt den Slice in §2 und §4 der Welle nach.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Messung der Planung, der Bezug ist der nächste, den es gibt), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Tests in der Stufe des Programms), [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (die Größenmarke als Gate), `AGENTS.md` §3.14 und §3.6 (eine Verschärfung, keine Lockerung).

**Berührte Spec-Stellen:** `SPEC-128` Nr. 11 (neu) und der Nachzug von Festlegung, Nr. 3 i, 5 i, 7 b, 8 und 10 g (Architect, 2026-11-24) · `SPEC-102` (Fälle der Gegenprobe zu `make groessenmarke`, Zählung) · `harness/README.md` §Sensors (Zeile von `make groessenmarke`) und *Nicht behauptet*.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Architect; Beschluss des User vom 2026-11-24: die Beobachtung `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` wird als Sensor verkörpert, niedrige Priorität). **Datum:** 2026-11-24.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem** Abschnitt des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make groessenmarke` meldet für jeden Plan in `in-progress/` mit vollständigem Block einen Befund, wenn ein Pfad des Blocks ein Verzeichnis ist und der Block eines anderen Plans in `open/`, `next/` oder `in-progress/` denselben Pfad oder einen darunter nennt, solange der Plan den Fall nicht mit dem Eintrag `teilt: <Verzeichnis> | <Pläne> | <Grund>` erklärt (`SPEC-128` Nr. 11); die Messung zählte sonst die Zeilen des anderen Plans mit (`SPEC-128` Nr. 10 g).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Dieselbe Datei in zwei Plänen (`Dockerfile`, `harness/README.md`) — kein Befund: es ist der Fall der Gate-Kette, den die Reihenfolge der Kette löst (Übergangsregeln Nr. 7); ein Befund träfe fünf Pläne zugleich, ohne dass einer etwas falsch macht (`SPEC-128` Nr. 11 k).
- Pläne anderer Zweige und Pläne in `done/` — das Programm liest die Dateien des Stands; ein Zweig, den der Stand nicht trägt, liest kein Treiber, und ein Plan in `done/` ist abgeschlossen (`SPEC-128` Nr. 11 k). Die Kollisionen nennt der Planner beim Start weiter in §4 jedes Plans.
- Eine Änderung der Formel, der Faktoren oder der Zählung der Messung — der Sensor beobachtet die Pfade, nicht die Größe (`SPEC-128` Nr. 1, 2).
- Ein Befund, ob ein Verzeichnis „allein“ einem Slice gehört — das Werkzeug kann es nicht wissen; der Plan erklärt es mit `teilt`, und der Reviewer liest den Grund (`SPEC-128` Nr. 11 e).
- Die Zeilen von `slice-lasttest-fixtures`, `slice-lasttest-auswertung` und `slice-lasttest` — sie tragen den Eintrag `teilt` in ihren Blöcken, wenn der Planner sie startet oder dieser Slice sie trifft (DoD, Rollout); ihre Pläne schreibt der Architect nicht.
- Eine Änderung am `Dockerfile` und an `harness/mk/` außer der Zeile des Ziels — das Programm ist dasselbe, die Stufe `groessenmarke` bleibt.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Das Programm: `groessenmarke pfade` und `groessenmarke fremde` mit der Relation *unter*, dem Eintrag `teilt` und den Zuständen von *P* und *Q* (Zeilen G01 bis G13 der Tabelle in §3; `SPEC-128` Nr. 11 a bis f und h).
- [ ] Der Treiber und der Container: `groessenmarke.sh` ruft je Plan `pfade`, den Typ je Pfad (`git cat-file -t`) und `fremde` und liest die Geschwisterverzeichnisse `open` und `next`; `groessenmarke-docker.sh` führt beide Kommandos aus; die Fälle der Gegenprobe (Zeilen G14 bis G17; `SPEC-128` Nr. 11 g und j; `SPEC-102` trägt sie nach).
- [ ] Der Rollout ohne Rot: jeder Plan in `in-progress/`, den das Werkzeug trifft, trägt vor dem Merge den Eintrag `teilt` (Zeile G18); `make groessenmarke` ist am Stand des Merge grün; die Zeile des Ziels in `harness/README.md` und *Nicht behauptet* sind nachgezogen (`AGENTS.md` §3.9), und das Ziel in `harness/mk/groessenmarke.mk` trägt den Anker `seit slice-groessenmarke-blockpfad`.
- [ ] `make gates` und `make gegenproben` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag: die Beobachtung `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` bekommt ihren Zielort (`liegt in` `Makefile:groessenmarke`, `seit slice-groessenmarke-blockpfad`); Beobachtungs-Register (`../observations/`) fortgeschrieben.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo mit Wellen von der nächsten Welle-Closure.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/groessenmarke/main.go` | update | die Kommandos `pfade` und `fremde` neben `abfrage` und `pruefen` (`SPEC-128` Nr. 11 g) |
| `tools/groessenmarke/fremde.go` | neu | die Relation *unter* über Segmente, die Befunde und Hinweise, der Eintrag `teilt` (`SPEC-128` Nr. 11 a bis f) |
| `tools/groessenmarke/plan.go` | update | der Schlüssel `teilt` im Block (`SPEC-128` Nr. 3 i, 11 e); das Lesen der Pfade eines fremden Plans mit denselben Regeln wie `abfrage` |
| `tools/groessenmarke/fremde_test.go`, `mengen_fremde_test.go`, `main_test.go`, `fixtures_test.go` | neu und update | die Fänger der Zeilen G01 bis G13; die Menge der Segmentfolgen als Schleife (G03) |
| `tools/harness/groessenmarke.sh` | update | `pfade`, `git cat-file -t HEAD:<pfad>` je Pfad, `fremde` mit den Plänen aus `in-progress`, `next` und `open`; Exit das Größte (`SPEC-128` Nr. 11 g) |
| `tools/harness/groessenmarke-docker.sh` | update | die Kommandos `pfade` und `fremde` mit `docker cp` je Datei |
| `tools/gegenproben/groessenmarke.go`, `groessenmarke_treiber.go`, `groessenmarke_schleifen.go`, `faelle.go` | update | die Fälle der Gegenprobe (G14 bis G17), registriert nach dem letzten Fall von `SPEC-102`; ein Ersatz für `git`, der `cat-file` kennt und zählt |
| `harness/mk/groessenmarke.mk` | update | allein die `##`-Zeile des Ziels: der Anker `seit slice-groessenmarke-blockpfad` und der Hinweis auf die Prüfung der Verzeichnisse |
| `harness/README.md` | update | Zeile von `make groessenmarke` in der Tabelle der Gates und der Satz in *Nicht behauptet* (`SPEC-128` Nr. 11 k) |
| Pläne in `in-progress/` | update (Planner) | der Eintrag `teilt` in jedem Plan, den das Werkzeug trifft (Rollout, G18); zum Stand 2026-11-24 allein `slice-lasttest-fixtures` (`tests/lasttest/` mit den Plänen `slice-lasttest-auswertung` und `slice-lasttest`) |
| `spec/`, `harness/vertraege/SPEC-128-groessenmarke.md` | geliefert (Architect, 2026-11-24) | `SPEC-128` Nr. 11 neu; dieser Slice schreibt sie nicht; `SPEC-102` (Zählung der Fälle) zieht der Implementer nach |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen vor dem Code (`AGENTS.md` §3.12); die Randformen entscheidet der Architect in `SPEC-128` Nr. 11 h, nicht der Implementer (Architect, 2026-11-24; jede Zeile nennt die Stelle des Vertrags). Der Implementer trägt je Zeile Mutation und roten Test in einen Absatz von §3 ein (`AGENTS.md` §3.10); die Zeilen sind `andere`, keine ist nebenläufig (das Programm liest Dateien nacheinander, der Treiber ruft nacheinander).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| G01 | (Nr. 11 a) Ein Verzeichnis *p* im Block von *P* und ein *q* gleich oder unter *p* im Block von *Q* ergeben je Paar (*p*, *Q*) einen Befund mit dem ersten *q* in Blockreihenfolge, dem Text `Verzeichnis <p> auch im Block von <Q>: <q>` und der Zeile `teilt: <p> \| <Q> \| <Grund>` | `TestFremdeBefund`: *p* mit und ohne `/` × *q* gleich, eine Ebene tiefer, drei Ebenen tiefer = 6 Zellen mit dem Text als Literal; *Q* mit zwei passenden *q*: der erste in Blockreihenfolge; Exit 1 | ein Befund je *q* statt je Paar; das letzte *q* genannt; die Zeile `teilt:` fehlt in der Meldung | andere |
| G02 | (Nr. 11 a, k) Eine Datei *p* gibt nie einen Befund, auch nicht dieselbe Datei in zwei Plänen | `TestFremdeDatei`: *p* Datei und *q* gleich (1 Zelle); `Dockerfile` im Block von *P* und von 3 Plänen *Q*: kein Befund, Exit 0 | der Typ wird ignoriert (jedes *p* gilt als Verzeichnis) | andere |
| G03 | (Nr. 11 c) *q* liegt unter *p* genau dann, wenn die Segmente von *p* ein Anfang der von *q* sind; der schließende `/` fällt weg, die Schreibweise zählt | `TestUnterSchleife`: Schleife über alle Paare von Segmentfolgen der Länge 1 bis 3 aus dem Alphabet {`a`, `ab`, `B`} (39 Folgen, 1521 Paare) × mit und ohne schließenden `/` an *p* und *q* (4); Orakel: Vergleich zweier Slices ohne den Code des Lesers; dazu `src/bootstrap2/x`, `src/boot`, `SRC/bootstrap` gegen `src/bootstrap` | Vergleich der Bytes als Präfix statt der Segmente; Groß- und Kleinschreibung ignoriert; `/` am Ende zählt als Segment | andere |
| G04 | (Nr. 11 b) Der Typ von *p*: `/` am Ende ist Verzeichnis (auch bei `blob`), `tree` Verzeichnis, `blob` Datei, ohne `/` und unbekannt Datei | Gegenprobe, Ersatz für `git` mit `cat-file`: 5 Zellen (`tree`, `blob`, unbekannt ohne `/`, unbekannt mit `/`, `blob` mit `/`), je mit dem Typ in der Typen-Datei des Aufrufs `fremde` | `tree` als Datei; unbekannt als Verzeichnis; `/` ignoriert | andere |
| G05 | (Nr. 11 d) *Q* sind die Pläne in `open/`, `next/` und `in-progress/`, nicht `done/`, nicht *P* (gleicher Dateiname); ein fehlendes Geschwisterverzeichnis ist leer | Gegenprobe: Wegwerf-Baum mit je einem Plan in den vier Verzeichnissen und einer Kopie von *P* in `open/`: genau drei *Q* im Aufruf von `fremde`, in der Folge `in-progress`, `next`, `open` und je darin nach Namen; ohne `next/` kein Fehler | `done/` gelesen; *P* gegen sich selbst verglichen; Folge nicht sortiert | andere |
| G06 | (Nr. 11 d) *Q* mit Block trägt seine Pfade (auch bei ausstehender Marke); ohne Block und in der Prosa-Form still übergangen; ein unlesbarer Block ist ein Hinweis, kein Fehler von *P* | `TestFremdeZustandQ`: 3 gültige Zustände und 5 unlesbare Formen (Block mehrdeutig, nicht geschlossen, Zeile ohne Schlüssel, ungültiger Pfad, doppelter Pfad) = 8 Zellen; Hinweis `Plan <Q>: Block nicht lesbar, nicht verglichen`, Exit des Aufrufs 0 | ein unlesbarer Block ist Eingabefehler (Exit 2); die Prosa-Form gibt einen Hinweis; ausstehende Marke trägt keine Pfade | andere |
| G07 | (Nr. 11 e) Der Eintrag `teilt` hat die Form `<p> \| <Q1>, … \| <Grund>`; jede andere Form ist der Eingabefehler `Eintrag ungültig: teilt` | `TestTeiltForm`: ohne zwei `\|`, *p* kein Pfad des Blocks, leere Kennung, doppelte Kennung = 4 Zellen; Gutfall mit zwei Kennungen; zwei Einträge für zwei Verzeichnisse | *p* nicht gegen den Block geprüft; doppelte Kennung zugelassen | andere |
| G08 | (Nr. 11 e) Der Eintrag deckt genau die genannten Paare; ein nicht genannter Plan bleibt ein Befund | `TestTeiltDeckt`: Menge der Paare (2 Verzeichnisse × 3 Pläne = 6 Zellen), der Eintrag nennt zwei der drei Pläne eines Verzeichnisses: 5 gedeckt, 1 Befund | der Eintrag deckt *p* für jeden Plan; deckt nur die erste Kennung | andere |
| G09 | (Nr. 11 e) Ein leerer Grund deckt, ist aber der Befund `teilt <p> ohne Grund` | `TestTeiltOhneGrund`: Grund leer und nur Leerraum = 2 Zellen: genau dieser Befund, kein Befund zu den Paaren | der leere Grund deckt nicht; kein Befund zum Grund | andere |
| G10 | (Nr. 11 e) Der Hinweis `teilt <p>: <Qi> berührt das Verzeichnis nicht` bei einem *p*, der eine Datei ist, bei einem *Qi* ohne Treffer und bei einem *Qi*, den es nicht gibt | `TestTeiltHinweis`: 3 Zellen, je Exit 0 und die Zeile auf stdout | der Hinweis ist ein Befund; er fehlt | andere |
| G11 | (Nr. 11 f) *P* mit vollständigem Block: Befund, Exit 1; mit ausstehender Marke: derselbe Fund als Hinweis, Exit 0; ohne Block oder in der Prosa-Form: weder Befund noch Hinweis | `TestFremdeZustandP`: 4 Zellen (vollständig, ausstehend, ohne Block, Prosa-Form) mit demselben *Q* | ausstehende Marke wird Befund; ein Plan ohne Block bekommt einen Hinweis | andere |
| G12 | (Nr. 11 g) `pfade` gibt je Pfad des Blocks `pfad <p>` in Blockreihenfolge, ohne Block nichts, bei Eingabefehler des Plans Exit 2 | `TestPfade`: 3 Zellen (Block mit 3 Pfaden in nicht sortierter Folge, ohne Block, Block mehrdeutig) | die Pfade sortiert; Exit 0 bei Eingabefehler | andere |
| G13 | (Nr. 11 g) `fremde`: Aufruf (kein Argument, leeres Argument, fehlende Typen-Datei), Typen gehören zum Plan (andere Liste, fehlende Zeile, zusätzliche Zeile, andere Reihenfolge), Exit 0, 1 und 2, Meldungen auf dem Strom von Nr. 6 a | `TestFremdeAufruf`: 3 + 4 + 3 Zellen | die Typen-Datei nicht gegen die Pfade des Blocks geprüft; der Befund auf stdout | andere |
| G14 | (Nr. 11 g, h 9) Der Treiber ruft je Plan in der Reihenfolge `abfrage`, die drei Aufrufe von `git`, `pruefen`, `pfade`, `git cat-file` je Pfad, `fremde`; der Exit ist das Größte aus allen Aufrufen aller Pläne; ein Exit über 2 gilt als 2 | Gegenprobe (Ersatz für das Programm und für `git`, jeder Aufruf gezählt und mit allen Argumenten festgehalten): die Aufrufliste als Literal; Exit-Kombinationen 3 × 3 von `pruefen` und `fremde` (9 Zellen); Exit 125 von `fremde` | `fremde` vor `pruefen`; der Exit von `fremde` verworfen; Exit 125 gilt als 0 | andere |
| G15 | (Nr. 11 g) Ohne `git` oder bei nicht lesbarem `HEAD` entscheidet der Schrägstrich allein und der Treiber meldet den Hinweis; ein Plan ohne Block ruft `git` nicht | Gegenprobe: 3 Fälle (`git` nicht aufrufbar, `HEAD` nicht lesbar, Plan ohne Block mit 0 Aufrufen von `git`) | Fehler (Exit 2) statt Hinweis; `git` für den Plan ohne Block gerufen | andere |
| G16 | (Nr. 11 h) Jede der 16 Randformen ist eine Zelle des Fängers, Orakel als Literal | `TestFremdeRandformen`: 16 Zellen, die Zählprüfung der Tabelle gleich 16 und die Folge der Befunde von (9) | eine Zelle entfällt (Zählprüfung rot) | andere |
| G17 | (Nr. 8, 11 j) Das Ziel wird rot durch einen Befund und grün durch den Eintrag `teilt`; `groessenmarke-docker.sh` führt `pfade` und `fremde` mit `docker cp` je Datei aus und erhält den Exit | Gegenprobe mit dem echten Programm und Ersatz für `git`: ein Plan mit Kollision ergibt Exit 1 und die Zeile `teilt:`, derselbe Plan mit Eintrag Exit 0 (2 Fälle); Ersatz für `docker`: die Aufrufe `create`, `cp` je Datei, `start -a`, `rm -f` in dieser Folge für beide Kommandos | der Befund endet mit Exit 0; `docker cp` nur für den Plan, nicht für die Pläne *Q* | andere |
| G18 | (Rollout, Nr. 11 e) Am Stand des Merge trägt jeder Plan in `in-progress/`, den das Werkzeug trifft, den Eintrag `teilt`; die Gate-Kette bleibt grün | `make groessenmarke` am Stand vor dem Merge: die Ausgabe (kein Befund); die Zeile `teilt:` in `slice-lasttest-fixtures`, solange der Plan in `in-progress/` liegt | — (eine benannte Ausgabe des Gates, kein Test) | andere |

**Mengen, Generatoren und Zählprüfungen** (Orakel als Literal im Test; `AGENTS.md` §3.15 und §3.17): die Menge der Segmentfolgen (39) und ihre Paare (1521) mit 4 Formen des schließenden `/` (G03, eine Schleife, kein Auszug); 6 Zellen Befund (G01); 5 Typen (G04); 8 Zustände von *Q* (G06); 6 Paare der Deckung (G08); 16 Randformen (G16); 9 Exit-Kombinationen (G14). Die Allaussage „kein Befund bei zwei Dateien“ (G02) und „der Eintrag deckt allein genannte Paare“ (G08) sind Schleifen über die Mengen, nicht Listen der Fälle, die einfielen. Die Achsen der Menge der Pläne: Zustand von *P* (3), Zustand von *Q* (4: Block, ausstehend, ohne Block, unlesbar), Ort von *Q* (3: `in-progress`, `next`, `open`), Typ von *p* (2).

**Zählprüfung der Zeilen:** 18 Zeilen, alle `andere` (*a* = 0, *b* = 18), *f* = 1 (der Ersatz für `git` mit `cat-file`), *p* = 250 (`fremde.go`, die Kommandos, der Treiber, das Skript des Containers).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-worktree-parallel](slice-worktree-parallel.md) liegt in `done/` (Vorgänger in der Gate-Kette, Position 15 der Reihenfolge vom 2026-11-24, Übergangsregeln Nr. 7); die Spezifikation trägt `SPEC-128` Nr. 11 vor dem ersten Code-Commit (`AGENTS.md` §3.12; geliefert am 2026-11-24); kein anderer Slice der Gate-Kette liegt in `in-progress/`. Dieser Slice steht **vor** [slice-gegenproben-eichung](slice-gegenproben-eichung.md): die Eichung misst `time make gegenproben` mit der Zahl der Fälle, und dieser Slice fügt Fälle hinzu; eine Eichung davor wäre mit seinem Merge überholt. Die Kette wird damit Position 16 (dieser Slice) und 17 (`slice-gegenproben-eichung`); der Planner trägt die Reihenfolge und den Start-Trigger der Eichung nach.

**Warum ganz hinten** (Architect, 2026-11-24): der Sensor schützt vor einer Überzählung der Messung, die in sechs Slices einen Eintrag `ueberschreitung` kostete und nie eine Marke brach; er hat niedrige Priorität (Beschluss des User). Er hält die Kette `tools/gegenproben/`, `harness/README.md` und `harness/mk/` allein, und kein Slice mit höherer Priorität wartet auf ihn.

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Vorschlag des Architect 2026-11-24; **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 18 Zeilen der Tabelle (G01 bis G18), *a* = 0, *b* = 18, *f* = 1, *p* = 250: *S* = 85·0 + 80·18 + 200·1 + 250 = 1440 + 200 + 250 = 1890, 135·1890 = 255150, **Marke 2600** = 100 · ⌈25,515⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223; der Abstand zur Schnittgrenze beträgt 333: wächst die Tabelle um vier Zeilen, löst sie aus, und der Planner prüft sie dann neu und schneidet vor dem Code, etwa Programm (G01 bis G13) und Treiber mit Gegenprobe (G14 bis G18)). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

```groessenmarke
pfad: tools/groessenmarke/main.go
pfad: tools/groessenmarke/fremde.go
pfad: tools/groessenmarke/plan.go
pfad: tools/groessenmarke/fremde_test.go
pfad: tools/groessenmarke/mengen_fremde_test.go
pfad: tools/groessenmarke/main_test.go
pfad: tools/groessenmarke/fixtures_test.go
pfad: tools/harness/groessenmarke.sh
pfad: tools/harness/groessenmarke-docker.sh
pfad: tools/gegenproben/groessenmarke.go
pfad: tools/gegenproben/groessenmarke_treiber.go
pfad: tools/gegenproben/groessenmarke_schleifen.go
pfad: tools/gegenproben/faelle.go
pfad: harness/mk/groessenmarke.mk
pfad: harness/README.md
produktcode: 250
fakes: 1
```

Der Block nennt die Dateien einzeln, kein Verzeichnis: dieser Slice ist der Sensor gegen ein Verzeichnis im Block und führt selbst keines (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`); die Dateien des Vertrags und der Spezifikation stehen nicht im Block (`SPEC-128` Nr. 3 m). Die neuen Dateien (`fremde.go`, `fremde_test.go`, `mengen_fremde_test.go`) gibt es noch nicht (die Namen sind ein Vorschlag des Architect; die Messung zählt dort 0, bis sie entstehen); der Implementer nennt jede weitere Datei, die er ändert oder anlegt, im selben Commit im Block.

**Kollisionen** (Architect, 2026-11-24, gelesen an den Plänen und am Block jedes offenen Plans; `SPEC-128` Nr. 10 g): (1) Die Gate-Kette (Nr. 7 der Übergangsregeln): `harness/README.md`, `tools/gegenproben/` und `harness/mk/` gehören diesem Slice allein, solange er in `in-progress/` liegt; die Vorgänger der Reihenfolge liegen vor dem Start in `done/`. (2) Kein offener Plan nennt `tools/groessenmarke/`, `tools/harness/groessenmarke.sh` oder `tools/harness/groessenmarke-docker.sh` im Block; [slice-abdeckung-deklaration](slice-abdeckung-deklaration.md) nennt das Werkzeug allein im Text (Muster, kein Pfad). (3) [slice-gegenproben-eichung](slice-gegenproben-eichung.md) misst die Fälle und ändert kein Programm; er folgt nach diesem Slice (Start). (4) Die Pläne der drei Teile des Lasttests nennen `tests/lasttest/` in ihren Blöcken: der Sensor trifft sie, und der Eintrag `teilt` gehört in den Plan des Teils, der in `in-progress/` liegt (Rollout, G18).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein vorab benannter Schnitt ist der Treiber mit Container und Gegenprobe (G14 bis G18) als eigener letzter Commit, sofern das Programm (G01 bis G13) vorher fertig und gemessen ist.
- `in-progress` → `open` (blockiert — Carveout?): ein Plan in `in-progress/` trägt einen Befund, den kein Eintrag `teilt` erklärt und den der Planner nicht auflöst; oder der Vorgänger der Kette fällt zurück nach `open/`.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-24; jede Zelle nennt ihren Ort in `SPEC-128` Nr. 11 oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `pfade <Plan>` | ohne Block: keine Ausgabe (Nr. 11 g, f); fehlende Datei: `Plan fehlt`, Exit 2 (Nr. 7 a) | ein Block ohne `pfad`: Eingabefehler `Block unvollständig` (Nr. 3 k) | ein ungültiger Pfad: Eingabefehler `Pfad ungültig` (Nr. 3 m); eine unbekannte Schlüsselzeile: `Schlüssel <k> unbekannt` (Nr. 3 i) | gilt nicht: ein Lesevorgang ohne Zustand (Nr. 6 c) | dieselbe Ausgabe (Nr. 6 c) |
| `fremde <Plan> <Typen> <Q>…` | kein *Q*: kein Befund (Nr. 11 g); keine Typen-Datei: Eingabefehler `Typen gehören nicht zu diesem Plan` | eine leere Typen-Datei bei einem Plan mit Pfaden: derselbe Eingabefehler (Nr. 11 g); ein leeres Argument: `Aufruf:` (Nr. 7 a) | andere Pfadliste, andere Reihenfolge: Eingabefehler (Nr. 11 g); *Q* mit unlesbarem Block: Hinweis (Nr. 11 d) | gilt nicht: wie oben | dieselbe Ausgabe (Nr. 6 c); dieselben Pläne ergeben dieselbe Folge (Nr. 11 h 9) |
| Typ eines Pfades (`git cat-file -t`) | `HEAD` fehlt oder `git` nicht aufrufbar: der Schrägstrich allein, Hinweis (Nr. 11 g) | ein Pfad ohne Namen: schon als Pfad ungültig (Nr. 3 m) | ein Pfad, den `HEAD` nicht kennt: Datei ohne `/` (Nr. 11 b); ein Pfad mit Muster: ein Name (Nr. 11 c) | gilt nicht: ein einzelner Aufruf; Abbruch des Treibers wie Nr. 7 f | derselbe Typ bei demselben `HEAD` |
| Eintrag `teilt` im Block von *P* | gilt nicht: der Eintrag ist freiwillig; fehlt er, gilt der Befund (Nr. 11 a, e) | leerer Grund: Befund `ohne Grund`, deckt (Nr. 11 e); leere Kennung: Eingabefehler (Nr. 11 e) | *p* kein Pfad des Blocks: Eingabefehler; ein *Qi*, das nichts berührt oder fehlt: Hinweis (Nr. 11 e) | gilt nicht: eine Zeile des Blocks | derselbe Eintrag zweimal (zwei Zeilen für dasselbe *p*): zwei Einträge, beide decken ihre Pläne (Nr. 3 i); eine doppelte Kennung in einem Eintrag: Eingabefehler |
| Suche der Pläne *Q* (Geschwisterverzeichnisse) | `open` oder `next` fehlt: leer (Nr. 7 b); beide fehlen: allein `in-progress` | ein Verzeichnis ohne `slice-*.md`: keine *Q* | Dateien anderer Namen: bleiben unberücksichtigt (Nr. 7 b); der Plan *P* selbst in einem Geschwisterverzeichnis: übergangen (Nr. 11 d) | `HUP`, `INT`, `TERM`: wie Nr. 7 f, Exit 2 | dasselbe Ergebnis bei unverändertem Stand |

**Zusatztabelle** (`AGENTS.md` §3.16: die Operation meldet mehrere Verstöße, und Zusagen hängen vom Zustand ab; je Paar von Verstößen am selben Schlüssel und je Zustand eine Zeile):

| Paar oder Zustand | Ergebnis | Fänger |
|---|---|---|
| zwei Verzeichnisse von *P* mit je einem Plan *Q* | zwei Befunde, in der Folge der Pfade des Blocks (Nr. 11 a) | G01, G08 |
| dasselbe Verzeichnis in zwei Plänen *Q1* und *Q2*, einer im Eintrag | ein Befund, für den anderen (Nr. 11 e, h 12) | G08 |
| Befund der Marke und Befund eines fremden Plans am selben *P* | beide gemeldet, Exit 1; die Ausgaben stammen aus zwei Aufrufen (Nr. 5 i) | G14 |
| *P* vollständig, ausstehend, ohne Block | Befund, Hinweis, nichts (Nr. 11 f) | G11 |

**Akzeptierte Negative** (`AGENTS.md` §3.15: Grund, Folge und Maß; ausführlich in `SPEC-128` Nr. 11 k):

1. **Pläne, die der Stand nicht trägt** (andere Zweige, `done/`). Grund: das Programm liest die Dateien des Stands. Folge: ein Befund bleibt aus, die Messung zählt die Zeilen trotzdem; der Eintrag `ueberschreitung` und die Kollisionen des Plans decken es. Maß: 0 Zweige gelesen.
2. **Dieselbe Datei in zwei Plänen.** Grund und Folge: wie §1. Maß: 5 Pläne nennen `Dockerfile` im Block (gezählt am 2026-11-24 in `open/`, `next/` und `in-progress/`).
3. **Der Grund von `teilt`.** Grund: jeder nicht leere Text deckt (`SPEC-128` Nr. 10 f). Folge: ein Eintrag, der „nacheinander“ sagt, ohne dass es stimmt, bleibt grün; der Reviewer liest ihn. Maß: 1 Eintrag zum Stand (`slice-lasttest-fixtures`).
4. **Pläne ohne Block.** Grund: sie tragen keine Pfade. Folge: ein Plan ohne Block ist nie Gegenstand eines Befunds. Maß: 18 von 35 Plänen in `open/`, `next/` und `in-progress/` am 2026-11-24 (17 mit Block).
5. **Die Verdrahtung des Containers.** Grund: kein Fall fängt, dass die Stufe `groessenmarke` das neue Kommando baut und das Ziel es ruft, außer G17 mit Ersatz für `docker` (`SPEC-128` Nr. 10 h). Folge: ein Ziel, das die Prüfung nicht ruft, bliebe grün; der Implementer fährt `make groessenmarke` vor der Übergabe an einem Plan mit Kollision ohne Eintrag und nennt das Rot.

Offene Punkte und Risiken:

- Der Sensor trifft beim Start sofort `slice-lasttest-fixtures` (und später die zwei anderen Teile): ohne Eintrag wäre `make gates` rot — **Ausgang:** entschieden (Architect, 2026-11-24): das Ziel der Rollout-Zeile G18; der Planner trägt den Eintrag `teilt: tests/lasttest/ | slice-lasttest-auswertung, slice-lasttest | die drei Teile laufen nacheinander, die Basis des späteren Teils liegt hinter dem Merge des früheren` in jeden der drei Pläne ein, wenn er ihn in `in-progress/` hält oder startet; sein Plan nennt es in §4 des jeweiligen Teils.
- Die Zahl der Pläne mit Block wächst; die Gegenprobe zählt die Pläne nicht, nur das Muster — **Ausgang:** weiter offen; kein Träger nötig: der Befund zeigt sich, sobald ein Plan ein Verzeichnis nennt.
- Das Verhalten bei einem Plan, den ein Zweig ändert (Block anders im Zweig als in `main`) — **Ausgang:** entfallen: der Sensor liest den Stand des Arbeitsbaums, jeder Zweig prüft seinen Stand; der Zweig mit dem Plan trägt den Eintrag mit dem Plan.
- Größe: die Tabelle hat 18 Zeilen, *S* liegt 333 unter der Schnittgrenze — **Ausgang:** entschieden (Architect, 2026-11-24): kein Schnitt vor dem Code; ein Schnitt ist vorab benannt (§4) und nur vor dem Code zu nehmen.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks).

*(wird bei der Closure geschrieben)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** berührt wird die Sub-Area `*` (Default des Repos, Kürzel `REPO`, Modus Greenfield, `harness/conventions.md`); der Slice liegt in `tools/groessenmarke/`, `tools/harness/`, `tools/gegenproben/` und `harness/`; die Schwelle von zwei Achsen ist für eine zweite Sub-Area nicht erfüllt.

**Vorgelagert — offene Beobachtungen sichten:** das Register (`../observations/`) ist durchgegangen; Treffer: `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (sechs Belege, Schwelle 3x überschritten; Gegenstand dieses Slice, Zielort entsteht mit der Closure), `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` (betrifft die Eichung der Gegenproben, die diesem Slice folgt).

**Modus-Begründungsblock:** alle berührten Sub-Areas GF.

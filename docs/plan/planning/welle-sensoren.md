# Welle welle-sensoren: Regeln, die in AGENTS.md nicht gewirkt haben, werden Sensoren

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. **Geplante Wellen bekommen noch keine
Datei:** Sie stehen in der Roadmap unter *Nächste Wellen* und nirgends sonst —
zwei Positionen, nicht drei.

**Zielmeilenstein:** kein Meilenstein-Bezug (Werkzeug-Welle; sie liegt außerhalb des Pfads von M2, siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Jede Hard Rule in `AGENTS.md`, die ein Gate tragen kann, wird von einem Gate getragen, und jedes dieser Gates weist mit einer Gegenprobe nach, dass es rot wird, wenn die Regel bricht.
Der Anlass ist der Retirement-Check [`docs/reviews/2026-10-25-retirement-check-hard-rules.md`](../../reviews/2026-10-25-retirement-check-hard-rules.md) (Bezug `LH-QA-06`, [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)): Regeln in Prosa haben die Beobachtung, aus der sie entstanden, nicht verhindert (§3.14: 16 Überschreitungen, Rate 65 % vor und 62 % nach der Schärfung), kein Gate misst die Größe der Guide-Dateien (`AGENTS.md` 24 648 Bytes, 337 Zeilen, 56 % davon §3.9 bis §3.18), und die Regelwerk-Faustregel sagt „computational vor inferential“ (`grundlagen-klassifikation.md`).
Entscheidung des Users (2026-10-30): erst die Sensoren, dann mit Messwerten verschlanken („Weg 2“).
Die Welle liefert die Sensoren; die Verschlankung selbst ist Folgearbeit nach der Welle (§6) und bekommt in der Closure ihre Messwerte und ihren Auftrag.

Belegt wird jede Regel einzeln, in der Tabelle der Closure-Notiz (§3): Regel · Sensor · was er fängt · was Prosa bleibt · Beleg.
Die Zuordnung der fünf Gegenstände:

| Gegenstand | Regel | Träger | Stand |
|---|---|---|---|
| Größenmarke | `AGENTS.md` §3.14 | `make groessenmarke` ([SPEC-128](../../../harness/vertraege/SPEC-128-groessenmarke.md)) | `slice-groessenmarke-werkzeug`, Welle [welle-record-replay](welle-record-replay.md); sein `done/` ist der Trigger (§2) |
| Unsichtbare Zeichen | `AGENTS.md` §3.18 | Prüfung auf BOM, NBSP, U+2028, U+2029, U+200B und rohe Steuerzeichen | `slice-unsichtbare-zeichen` |
| Größe der Guide-Dateien | Guide-Datei-Wildwuchs (`grundlagen-klassifikation.md`), alle Hard Rules | d-check-Modul `file` | `slice-guide-groessen-gate` |
| Zusage und Fänger, Randformen | `AGENTS.md` §3.15, §3.16 | Abdeckungs-Deklaration über den Tests, erzeugte Tabellen, `trace` | `slice-abdeckung-deklaration` |
| Code-Abdeckung | (keine Regel; Lücke: „Nicht behauptet“ in `harness/README.md`) | Schwelle der Anweisungsabdeckung in der Stufe `test` | `slice-coverage-gate` |

Das Vorbild für die letzten beiden Gegenstände und den Anstoß dazu liefert das Nachbar-Repo `pgwire-recorder` (nur gelesen): Vertrag zur Abdeckung, Skript, Gegenprobe, Schwelle in einem Vertrag.
Übernommen wird nichts blind: Der Architect entscheidet je Slice, was für ein Go-Hexagon mit anderem Lastenheft, deutscher Sprache und dem Rahmen der Gegenproben nach [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) passt.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

Die Welle startet gestaffelt, in zwei Teilen mit je eigenem Trigger. Sie bleibt **eine** Welle, weil ihr Closure-Trigger (§3) eine Tabelle über alle Sensoren, eine Messung der Gegenproben vor dem ersten und nach dem letzten Slice und einen Diff-Vorschlag je Regel verlangt; zwei Wellen trügen zwei Teil-Tabellen und zwei Messungen, und die Verschlankung von `AGENTS.md` hätte keine Stelle, an der alle Regeln mit Gate beisammen stehen.

- **Teil 1** (`slice-unsichtbare-zeichen`, `slice-guide-groessen-gate`): `slice-groessenmarke-werkzeug` liegt in `done/`. Grund: Das Werkzeug ist der erste der Sensoren (`AGENTS.md` §3.14) und sein Handoff-Schnitt (Gegenproben, Teil G) berührt dieselben Dateien der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel); solange er in Arbeit ist, hat die Welle keine feste Basis. Mit dem Werkzeug setzt der Planner danach die Marken der Slices dieser Welle (`make groessenmarke`, [SPEC-128](../../../harness/vertraege/SPEC-128-groessenmarke.md)). Teil 1 liegt vor `slice-workspace-use-case` (Reihenfolge, kein Blocker, siehe §5).
- **Teil 2** (`slice-abdeckung-deklaration`, `slice-coverage-gate`): `welle-record-replay` liegt in `done/` (M2 erreicht) und `slice-guide-groessen-gate` liegt in `done/`. Grund: Der Backfill der Deklaration über 618 Testfunktionen in 105 Dateien `*_test.go` (Stand `0c524e2`) berührt die Dateien, die die Slices von M2 gerade schreiben, und kollidiert mit ihnen; er wäre dazu der größte Posten der Welle (Aufwand eher L, `slice-abdeckung-deklaration` §4). Die Schwelle der Anweisungsabdeckung misst den Code, den M2 liefert, und ändert die Laufzeit von `make test` für jeden Slice davor; sie wird gegen den Code gesetzt, der dann steht.
- Die Welle steht in der Roadmap unter *Nächste Wellen* (Zeile mit beiden Triggern, wichtigsten Slices und Aufwand), solange der Trigger von Teil 1 nicht eingetreten ist; mit dem Eintritt entfernt der Planner die Zeile und trägt den Zeiger unter *Offene Wellen* ein. Der Trigger von Teil 2 steht in der Roadmap und im Abhängigkeitsgraphen als eigene Kante. Diese Datei liegt vorab flach in `planning/` (Auftrag des Users vom 2026-10-30, Abweichung von der Lifecycle-Note oben, die für eine geplante Welle noch keine Datei vorsieht), damit die Pläne in `open/` auf sie zeigen; der Zeiger unter *Offene Wellen* bleibt aus, bis der Trigger von Teil 1 eintritt.

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle vier Slices der Welle liegen in `done/`; die Welle schließt erst nach Teil 2, auch wenn Teil 1 früher fertig ist (ein Closure-Trigger, eine Notiz, eine Tabelle).
- `make gates` ist grün und fährt die Ziele, die die Slices liefern; `make gegenproben` ist grün. Beide laufen im Hauptbaum auf ruhigem Host, nicht im Worktree dieser Planung.
- Das *Mehr* gegenüber den Slice-DoDs: Die Closure-Notiz `welle-sensoren-results.md` trägt eine Tabelle mit einer Zeile je Hard Rule, die ein Gate tragen kann (§3.14, §3.15, §3.16, §3.18) und je Gegenstand ohne Regel (Größe der Guide-Dateien, Code-Abdeckung): *Regel · Sensor · Fänger · was Prosa bleibt (Ungeprüft) · Beleg (Gegenprobe oder rotes Gate)*. Eine Regel, deren Sensor nur einen Teil trägt, nennt den Rest ausdrücklich; eine Regel ohne Sensor steht mit Grund da.
- Die Closure-Notiz trägt die Messwerte, die die Verschlankung braucht: die Größen von `AGENTS.md`, `.claude/commands/implement-slice.md`, `.harness/skills/reviewer.md` und der übrigen Guide-Dateien zum Stand der Closure (Bytes und Zeilen, die Grenzen des Gates daneben), die Zeichenzahl der Hard Rules, die ein Gate trägt, und die Zeilen dieser Regeln, die ein Gate ersetzt (Zeiger statt Prosa möglich).
- Die Closure-Notiz trägt die gemessene Zusatzzeit der Gegenproben je Slice und in Summe (Schritt und Gesamtlauf von `make gegenproben` bei geänderter Eingabe, vor dem ersten und nach dem letzten Slice der Welle, Zahl der Läufe und Zustand des Hosts genannt) als Eingabe für die Entscheidung des Users zum Re-Evaluierungs-Trigger von [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md); die Welle entscheidet ihn nicht.
- Die Closure-Notiz nennt als Auftrag an den User und den Planner die Verschlankung von `AGENTS.md` (Weg 2) und legt je Regel mit Gate einen Diff-Vorschlag (Zeiger statt Prosa) bei; geändert wird `AGENTS.md` von keiner Rolle dieser Welle.
- Die *Steering-Loop-Einträge* der Closure-Notiz tragen je Sensor mit dem Anker `seit welle-sensoren` das Feld `liegt in` (`Makefile:<Ziel>` oder `.d-check.yml`), soweit er verkörpert ist.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-unsichtbare-zeichen | Prüfung auf unsichtbare Zeichen in Quelltext, Werkzeugen und Dokumenten | `LH-QA-06` |
| slice-guide-groessen-gate | Größe der Guide-Dateien als Gate (d-check-Modul `file`) | `LH-QA-06` |
| slice-abdeckung-deklaration | Abdeckungs-Deklaration über den Tests, erzeugte Tabellen, Anforderung → Test | `LH-QA-06` (Gegenstand der Deklaration sind alle `LH-FA-*`, `LH-QA-*` und `LH-RB-*`) |
| slice-coverage-gate | Schwelle der Anweisungsabdeckung der Unit-Tests | `LH-QA-06`, `LH-QA-07` |

Das Lastenheft führt keine Anforderung an Gates der Harness; wie bei `slice-groessenmarke-werkzeug` trägt `LH-QA-06` (Portabilität und Reproduzierbarkeit: Prüfläufe im Container, netzlos) den Bezug. Ob das Lastenheft eine Anforderung an die Prüfbarkeit des Quellcodes braucht (das Nachbar-Repo führt sie als `LH-QA-07`), ist eine Lücke, die der Architect in `slice-coverage-gate` benennt; geändert wird das Lastenheft nicht von der Planung.

**Reihenfolge** (Pfeil heißt: danach): Teil 1 `slice-unsichtbare-zeichen` → `slice-guide-groessen-gate`; Teil 2 `slice-abdeckung-deklaration` → `slice-coverage-gate`. Teil 1 liegt zeitlich vor `slice-workspace-use-case`, Teil 2 nach `welle-record-replay` (Trigger, §2). Alle vier berühren die Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, `harness/gegenproben.stempel`) und laufen deshalb nacheinander, nie parallel; jeder Slice beginnt (Start `next` → `in-progress`) erst, wenn der vorige in `done/` liegt. Gründe der Folge:

1. `slice-unsichtbare-zeichen` zuerst: klein (S), schützt jede Datei, die die folgenden Slices schreiben (`AGENTS.md` §3.18), und braucht keine Entscheidung, die ein späterer Slice voraussetzt.
2. `slice-guide-groessen-gate` danach: klein (S), friert den Ist-Stand ein, bevor `slice-abdeckung-deklaration` und `slice-coverage-gate` ihre Zeilen in `harness/README.md` und die Tabelle der Sensoren hinzufügen; er entscheidet, ob `harness/README.md` selbst unter die Grenze fällt.
3. `slice-abdeckung-deklaration` vor `slice-coverage-gate`: die Deklaration sagt, welcher Test welche Anforderung trägt (Form), die Schwelle sagt, wie viel Code ein Test erreicht (Menge); die Form kann nicht von der Schwelle abhängen, die Schwelle nicht von der Form, aber `slice-coverage-gate` ändert die Stufe `test` und damit die Laufzeit von `make test`, und diese Last soll nicht unter die Entwicklung der Deklaration fallen. Die Folge der beiden letzten ist eine Empfehlung des Planners, kein Zwang; der Architect darf sie umkehren, wenn der Zuschnitt es zeigt (Rückgabe an den Planner mit Grund).

**Entscheidung zu `slice-plan-form-sensor`** (Vorschlag des Retirement-Checks, Abschnitt 5, M): kein fünfter Slice. Die Frage ist als Auftrag Nr. 3 in `slice-abdeckung-deklaration` aufgenommen: Trägt die Abdeckungs-Deklaration die Form der Pläne (Tabelle *Zusage · Fänger*, Raster in §6) oder nicht, und braucht es dann eine `structure`-Regel auf die Slice-Pläne? Grund: Der Gegenstand (Plan-Form) ist nur dann ein eigener Slice, wenn die Deklaration ihn nicht trägt; diese Antwort liegt beim Architect von `slice-abdeckung-deklaration`, nicht beim Planner. Kommt er zu „nicht getragen“, gibt er den Punkt mit Grund an den Planner zurück (Ausgang *offen* in §6 des Slice, Adresse: Planner), und der Planner legt dann `slice-plan-form-sensor` als Slice an und nimmt ihn in diese Welle oder in eine Folgewelle (Entscheidung des Users).

**Betriebsregeln für alle Rollen** (Lehren der Läufe vom 2026-10-24; sie stehen in jedem Plan der Welle, damit ein Lauf sie am Arbeitsort liest):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; höchstens zwei Container gleichzeitig; Mutationen nacheinander, nie parallel. Vor einem großen Lauf `free -g`. Docker sonst nur über `make` (`AGENTS.md` §3.1). Die Gates bauen Images mit festen Tags (`odata-recorder:*`) im gemeinsamen Daemon: zwei Gate-Läufe aus zwei Arbeitsbäumen stören sich.
- `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`): ein Rot unter Last ist kein Befund, bevor es auf ruhigem Host wiederholt ist.
- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Das Werkzeug setzt eine Folge aus Backslash, `u` und vier Hex-Ziffern still um (`AGENTS.md` §3.18): Fixtures, Pläne und Berichte mit solchen Folgen, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen; der Bericht nennt die gegengelesene Datei.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 des Slice mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen, der Stempel `harness/gegenproben.stempel` erst nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieser Welle; was dort zu ändern ist, geht als Diff-Vorschlag in der Closure an den User.
- Jeder Slice mit einer neuen Gegenprobe misst deren Zusatzzeit (Schritt und Gesamtlauf von `make gegenproben` bei geänderter Eingabe, vorher und nachher, gleicher Host-Zustand, Zahl der Läufe genannt) und trägt sie in §7 ein; der Slice entscheidet den Re-Evaluierungs-Trigger von [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) nicht.

**Auftrag beim Start jedes Slice** (Planner, vor dem ersten Code-Commit, nach dem Architect): die Marke nach `AGENTS.md` §3.14 mit `make groessenmarke` ([SPEC-128](../../../harness/vertraege/SPEC-128-groessenmarke.md)) setzen, den Block in §4 des Plans füllen und die Sichtung in §8 gegen den Stand des Registers erneuern. Die Marke ist in den Plänen dieser Welle bewusst nicht gesetzt: sie hängt an der Tabelle des Architect, und das Werkzeug liegt erst mit dem Trigger vor.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Wird blockiert von: Teil 1 von `slice-groessenmarke-werkzeug` (Welle [welle-record-replay](welle-record-replay.md)) wegen der Gate-Kette (§2); Teil 2 von `welle-record-replay` als Ganzer (Backfill und Schwelle gegen den Code von M2, §2) und von `slice-guide-groessen-gate` (Reihenfolge der Welle). Beide Kanten stehen im Abhängigkeitsgraphen der Roadmap.
- Reihenfolge ohne Blocker: Teil 1 liegt vor `slice-workspace-use-case` (er schützt jede Datei, die dieser und die folgenden Slices schreiben, mit `slice-unsichtbare-zeichen` und friert mit `slice-guide-groessen-gate` die Größe der Guide-Dateien ein). Es gibt keine Abhängigkeit im Code; die Slices dieser Welle ändern `Makefile`, `harness/`, `tools/` und `Dockerfile`, nicht `src/`. Startet `slice-workspace-use-case`, bevor Teil 1 fertig ist, ist das kein Verstoß; der Planner nennt dann im Start-Commit den Grund. Dasselbe gilt für `slice-handbuch-geruest` ([welle-benutzerhandbuch](welle-benutzerhandbuch.md)): Es liegt nach Teil 1 (Reihenfolge).
- Blockiert: keine Welle. Die Welle `welle-dev-container` ist inhaltlich unabhängig, berührt aber in ihrer Umsetzung dieselben Dateien der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel), weil ihr Container `make gates` und `make gegenproben` fährt. `welle-dev-container` startet nach `welle-record-replay` (Entscheidung des Users, 2026-10-30; Gate-Kette dann stabil); Teil 2 und `welle-dev-container` laufen nicht gleichzeitig. Der Planner empfiehlt Teil 2 zuerst (der Container fährt dann die Kette samt Schwelle, und die Messung der Gegenproben-Zeit liegt vor seiner Laufzeitfrage); die Reihenfolge der beiden entscheidet der User. Dagegen spricht: `welle-dev-container` trennt die Daemons und beseitigt den Tag-Konflikt der Arbeitsbäume; läuft sie zuerst, sind die Läufe dieser Welle unabhängiger voneinander.
- Unabhängig: die übrigen Slices von `welle-record-replay`. Sie ändern `src/` und Pläne; `slice-coverage-gate` misst deren Code, ändert ihn nicht. Dass Teil 2 erst nach M2 liegt, nimmt dem Slice das Risiko, dass die Schwelle einem wachsenden Code hinterherläuft.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Die Verschlankung von `AGENTS.md` (Hard Rules auf Zeiger und Kurzform, Volltext in einer Datei je Regel unter harness/rules oder in die Rollen-Dateien, Änderung der Zeiger in Commands) — Folgearbeit nach der Welle, Entscheidung „Weg 2“ des Users: erst Sensoren, dann Messwerte. Die Closure nennt den Auftrag an User und Planner; `AGENTS.md`, Commands, Rollen-Typen und Skills ändert nur der User.
- Die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Lauf über 30 s bei geänderter Eingabe; gemessen 56,9 s Schritt, 87 bis 89 s Gesamt) — liegt beim User; die Welle liefert die Messung der Zusatzzeit.
- Die Größenmarke als Werkzeug (`AGENTS.md` §3.14) — `slice-groessenmarke-werkzeug`, Welle record-replay; sie ist Trigger, nicht Inhalt.
- Sensoren für Regeln, die der Retirement-Check als nicht berechenbar oder nicht netzlos eingestuft hat: ein Mutationswerkzeug (§3.10, bräuchte Netz für `deps`), ein Test „Zahl im Plan = Zahl im Test“ (§3.9), ein Wächter für das Lesen fremder Berichte (§3.11) und ein Teil-Sensor für §3.9 („Plan im selben Commit“, Fehlalarm bei Fix-Runden) — der Retirement-Check empfiehlt sie nicht; sie bleiben Prosa, bis der User anders entscheidet.
- Die Übernahme der Gegenstände aus `slice-plan-form-sensor` (`structure`-Regel auf Slice-Pläne) — nur über den Auftrag Nr. 3 in `slice-abdeckung-deklaration` und die Entscheidung des Architect (§4).
- Die Rückwirkung auf bestehende Pläne, Tests und Dokumente, soweit ein neues Gate sie rot färbt — jeder Slice benennt die Stellen im Bestand, die er mit dem Gate berichtigt, und berichtigt sie im Slice (keine Ausnahme-Liste ohne Carveout, `AGENTS.md` §3.6); ein Backfill, der den Slice über die Marke hebt, wird vor dem Code geschnitten (`AGENTS.md` §3.14) und ist ein eigener Slice.
- Der Lasttest zu `LH-QA-01` und ein E2E-Test (`harness/README.md` *Nicht behauptet*) — Gegenstand der Produktwellen.
- Änderungen am Lastenheft und an `spec/` — Sache des Owners und des Architect; Lücken werden benannt, nicht geschlossen.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-sensoren-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).

# Review-Report: slice-gegenproben-zeitbudget, zweite Runde — 2026-11-18

**Review-Art:** Code (Nachprüfung der Fix-Runde gegen Plan, Entscheidungen und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers).

**Gegenstand:** `slice-gegenproben-zeitbudget`, Zweig `slice-gegenproben-zeitbudget`, HEAD `d1e6c27c`, Diff `1697c650..HEAD` (Architect: Plan §3 Zeilen 8 und 9, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) *Zeitbudget* Nr. 1, 3 bis 5, Randform 6, *Grenze*; Implementer: `tools/gegenproben/frist_test.go` mit `TestFuehreFallReichtDenAblaufWeiter`, `aufrufeVon`, `zuweisungenVon`, `letzterName`, neuer Kopf von `TestJederAufrufNenntDieFallFrist`; Plan §3 *Nachweis*, §4/§6). Vorgänger: [Report der ersten Runde](2026-11-18-slice-gegenproben-zeitbudget-review.md) (F-1 bis F-6).

**Skill:** `.harness/skills/reviewer.md` @ HEAD `d1e6c27c` · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-18

**Eingangs-Kontext:** Plan (§3 Tabelle Zeilen 1 bis 9, *Nachweis*, §4 Block, §6 Raster und Zusatztabelle), [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (Accepted), [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md), [SPEC-128](../../harness/vertraege/SPEC-128-groessenmarke.md), `AGENTS.md` §3.6 bis §3.18, erster Report.

**Vorgehen.** Keine Einschätzung des Implementers übernommen. Auflösung gegen die Klasse (§3.17), nicht gegen die genannten Mutationen: 31 eigene Mutationen in Kopien (`git archive HEAD`, `tools/gegenproben`, im Scratchpad `rev-zb2/`, nie im Zweig), ein Prozess je Mutation (`go vet .` und `go test -count=1 -v .`), in einem Container aus `odata-recorder:race` ohne Netz, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /home/builder/work/odata-recorder/.wt/heavy.lock`, Last bei Last unter 20 (alle Starts zwischen 7,2 und 10,8). Messender: der Reviewer; Last = erstes Feld von `/proc/loadavg` zu Beginn je Mutation, im Container gelesen.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| G-1 | INFO | `TestFuehreFallReichtDenAblaufWeiter` wird auch bei Umbauten rot, die das Verhalten nicht ändern: eine Zwischenvariable `weiter := abgelaufen` (rot, `…/Bau` und `…/Gate`, je 1 von 1), eine Hilfsfunktion, die den Ablauf getreu weiterreicht (rot mit `fuehreFall ruft bewerteBau 0-mal, erwartet 1`), und, nur aus dem Quelltext gelesen, ein Verschieben von `fuehreFall` aus `main.go` (rot mit `fuehreFall nicht gefunden`, der Test liest den Dateinamen `main.go` fest, `TestJederAufrufNenntDieFallFrist` liest das Verzeichnis). Der Plan nennt die ersten zwei Umbauten ausdrücklich („der Umbau trägt die Änderung des Tests“), das Verschieben nicht. Umgekehrt bleibt der Test grün bei Kommentar, Leerzeile, Zeilenumbruch im Aufruf und reiner Umbenennung von `abgelaufen` (siehe Tabelle, Zeilen m18 bis m24): kein Fall von Zerbrechlichkeit gegen Format. | Maintainability | `tools/gegenproben/frist_test.go` · `parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)` | ja | Test rot bei verhaltensgleichem Umbau, Fall nicht im Plan genannt |
| G-2 | INFO | [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) Entscheidung („ein Lauf außerhalb davon ist *ungültig* … und wird wiederholt“) trägt die engere Lesart von [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) *Zeitbudget* Nr. 4 und 5 d nicht mehr wörtlich: ein grüner ungültiger Lauf wird nicht wiederholt. Die ADR sagt, die Einzelheiten stünden in der Spezifikation, und ist `Accepted` (`AGENTS.md` §3.5); die Spezifikation führt im Rang. Kein Widerspruch im Verhalten, ein Satz der ADR liest sich allgemeiner. Adresse: Architect. | `AGENTS.md` §3.5, §3.8 | `docs/plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md` · „und wird wiederholt“ (Abschnitt Entscheidung) | nein | ADR-Satz allgemeiner als die Spezifikation |
| G-3 | INFO | Der neue Test führt dreimal dieselbe Durchlaufform (`ast.Inspect` über CallExpr, Bezeichner `Fun`): in `TestJederAufrufNenntDieFallFrist`, `aufrufeVon` und `zuweisungenVon`; `frist_test.go` wächst von 71 auf 175 Zeilen (neu rund 104, davon der Test rund 62). Der Umfang trägt die Zusage (Bindung, Zahl, Argument, Vorkommen); Doppelung ohne Wirkung auf die Zusage. Lint (cyclop, gocognit, gochecknoglobals) meldet nichts. | Maintainability | `tools/gegenproben/frist_test.go` · `func zuweisungenVon` | nein | Doppelte Durchlaufform im Test |
| G-4 | INFO | Zu F-4 der ersten Runde: die Roadmap-Zeile trägt weiter `open/` in der Klammer; der Plan führt den Punkt jetzt in §6 mit Ausgang *offen* und Adresse Planner (`AGENTS.md` §3.13 erfüllt). Kein neues Finding, die Aufgabe liegt beim Planner. | `AGENTS.md` §3.7, §3.13 | `docs/plan/planning/in-progress/roadmap.md` · „(`open/`, Linie C, Kettenslice, **fünfter der Kette**“ | nein | Zustandsfeld, adressiert |

Kein HIGH, kein MEDIUM, kein LOW.

---

## Auflösung von F-1 bis F-6

| Erstbefund | Stand | Beleg |
|---|---|---|
| F-1 (Weitergabe des Ablaufs, MEDIUM) | **behoben** für die Klasse | Menge: die Stellen, an denen `fuehreFall` das Ergebnis eines Prozesses wertet, 2 (Bau `starte`/`bewerteBau`, Gate `starteStroeme`/`bewerte`); `grep` über `tools/gegenproben/*.go` ohne Tests: `bewerteBau(` nur Zeile 249, `bewerte(` nur Zeile 255 im Aufruf. Der Test läuft als Schleife über beide Paare; alle Mutationen der Tabelle unten an beiden Elementen rot aus dem benannten Grund. Das Negativ ist entfallen, der Plan trägt Zeilen 8 und 9. |
| F-2 (Prüfgrenze der Aufrufstellen, MEDIUM) | **behoben** als verengte Zusage mit benanntem Negativ | Kopf von `TestJederAufrufNenntDieFallFrist`, Plan §3 Zeile 2 und Negativ, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) *Grenze* nennen dieselben drei Formen mit Grund und Folge; die zwei Formen, die ich im ersten Lauf grün sah (Funktionswert mit zweiter Stelle, überdecktes `fallFrist`), bleiben grün (m25, m26), wie das Negativ sagt. |
| F-3 (Cache in Nr. 1, LOW) | behoben | Nr. 1 nennt `--no-cache-filter gegenproben` und dass der Cache der übrigen Stufen nicht zur Messgröße gehört; die Zelle des Rasters trägt denselben Wortlaut. |
| F-4 (Roadmap, LOW) | offen, adressiert (G-4) | Plan §6. |
| F-5 (Sperre, INFO) | behoben im Plan | Betriebsregeln nennen `/home/builder/work/odata-recorder/.wt/heavy.lock`, §6 trägt den Rest der Übergangsregel Nr. 1 mit Adresse Planner. |
| F-6 (zwei Lesarten, INFO) | behoben, genau eine Lesart | Nr. 4: wiederholt wird nur ein roter ungültiger Lauf (Nr. 5 e); ein grüner ungültiger (Nr. 5 d) wird nicht wiederholt, schreibt den Stempel, zählt nicht für die drei; drei ungültige rote Läufe in Folge sind der Befund. Nr. 3, Randform 6, Plan §1, §6 Raster und Zusatztabelle tragen dieselbe Lesart (`grep wiederh` über [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md), Plan, ADR: keine Stelle sagt noch „jeder ungültige Lauf wird wiederholt“ außer G-2). |

## Mutationstabelle (eigene Messung)

Je Mutation 1 Prozess, Klasse *andere* (kein Wahrscheinlichkeits-Fänger, keine Nebenläufigkeit), Rate rot in 1 von 1; `go vet` Exit 0 und `gofmt -l` leer in allen Kopien (außer der Hilfsdatei `extra.go` von m14 bis m17, die ich selbst nicht formatiert hatte). Last zu Beginn des Prozesses in Klammern. Der Lauf ohne Mutation (m00, 9,6) ist grün.

| Kopie | Mutation (in `fuehreFall`, sofern nicht anders genannt) | Ergebnis | Meldung (Log) |
|---|---|---|---|
| m01 (10,5) | Bau: `bewerteBau(code, out, abgelaufen && false)` | rot | `frist_test.go:161: das letzte Argument von bewerteBau ist nicht der Bezeichner abgelaufen`, `…/Bau` |
| m02 (10,5) | Gate: `bewerte(…, abgelaufen && false)` | rot | wie m01 für `bewerte`, `…/Gate` |
| m03 (10,5) | Bau: Zuweisung `abgelaufen = false` vor der Wertung (Element, das der Erstbefund nicht nannte) | rot | `abgelaufen steht 5-mal in fuehreFall, erwartet 4`, beide Unterfälle |
| m04 (10,2) | Gate: `abgelaufen = false` vor der Wertung | rot | wie m03, beide Unterfälle |
| m05 (10,2) | Bau: Zwischenvariable `weiter := abgelaufen && false` | rot | letztes Argument nicht `abgelaufen`, `…/Bau` |
| m05n (10,2) | Bau: Zwischenvariable `weiter := abgelaufen` (verhaltensgleich) | rot | wie m05 (siehe G-1) |
| m06, m06n (9,4) | Gate: wie m05, m05n | rot, rot | letztes Argument nicht `abgelaufen`, `…/Gate` |
| m07 (9,4) | Bau: Argument `false`, `_ = abgelaufen` davor (Name kommt vor, Argument nicht) | rot | letztes Argument nicht `abgelaufen` |
| m08 (9,4) | Gate: wie m07 | rot | letztes Argument nicht `abgelaufen` |
| m09 (8,9) | Gate: `bewerte(f, code, errOut, out, abgelaufen)` (Vertauschung zweier Zeichenketten-Argumente; `abgelaufen` bleibt letztes) | rot, **nicht aus dem Fänger der Weitergabe** | `TestFaelleDesHandbuchsLaufenGegenPruefung`, `TestFuehreFallStartetMitDerHuelleDesFalls`; ein Tausch von `abgelaufen` mit einem anderen Argument kompiliert nicht (Typen) |
| m10 (8,9) | Bau: zweiter Aufruf `_ = bewerteBau(…)` | rot | `fuehreFall ruft bewerteBau 2-mal, erwartet 1` (und Vorkommen 5 statt 4) |
| m11 (8,9) | Gate: zweiter Aufruf `_ = bewerte(…)` | rot | `fuehreFall ruft bewerte 2-mal, erwartet 1` |
| m12 (8,3) | Bau: `code, out, _ := starte(…)` und Argument `false` | rot | `der Aufruf von starte bindet den Ablauf an "_"` |
| m13 (8,3) | Gate: `_` statt Name und Argument `false` | rot | `der Aufruf von starteStroeme bindet den Ablauf an "_"` |
| m29 (8,7) | Plan-Zeile „Bindung“: `code, out, _ := starte(…)` und `abgelaufen := false` | rot | Bau: `bindet … an "_"`; Gate: `abgelaufen steht 4-mal in fuehreFall, erwartet 2` (Plan §3 *Nachweis* stimmt) |
| m14 (8,3) | Bau: Aufruf über Hilfsfunktion `wertungBau`, die `bewerteBau(…, false)` ruft | rot | `fuehreFall ruft bewerteBau 0-mal, erwartet 1` (rot, aber durch die Zahl der Aufrufe, nicht durch das Argument; der Plan benennt das als Umbau, der den Test mitnimmt) |
| m15 (8,1) | Bau: Hilfsfunktion, die getreu `abgelaufen` weiterreicht | rot | wie m14 (verhaltensgleich, G-1) |
| m16, m17 (8,1) | Gate: Hilfsfunktion `wertungGate` mutierend / getreu | rot, rot | `fuehreFall ruft bewerte 0-mal, erwartet 1` |
| m18 bis m20 (8,1; 7,6) | verhaltensgleiche Umbenennung `abgelaufen` in `weiter` an Bindung und Argument: nur Bau, nur Gate, beide | **grün** | gewollt: der Name wird aus dem Quelltext gelesen, nicht festgelegt |
| m21 (7,6) | Nachher überschreibt `rot` unbedingt (`rot == "" &&` entfällt; ein gerissenes Gate würde grün, wenn `Nachher` leer liefert) | rot | `TestFallMitEnvVorbereitungUndNachher` (m27 ist dieselbe Mutation, rot) |
| m22 (7,2) | Bau: `rot != ""` wird `rot == ""` | rot | `TestFallMitBauMeldetScheiternDesBaus`, `TestFuehreFallBauScheitertRuftKeinGate` |
| m28 (10,8) | Bau: das Rot des Baus wird verworfen | rot | dieselben zwei Tests |
| m23, m24 (7,2) | Kommentar und Leerzeile in `fuehreFall`; Aufruf von `bewerte` auf mehrere Zeilen umbrochen | **grün** | gewollt, kein Format-Bruch |
| m25 (9,2) | F-2: `var starteFn = starte` mit zusätzlicher Stelle `starteFn(3600*time.Second, …)` | **grün** | benanntes Negativ (Funktionswert) |
| m26 (9,2) | F-2: `fallFrist := 3600 * time.Second` im Rumpf von `fuehreFall` | **grün** | benanntes Negativ (überdecktes `fallFrist`) |

Alle roten Läufe bis auf m09, m21, m22, m28 sind `--- FAIL: TestFuehreFallReichtDenAblaufWeiter` mit Meldung in `frist_test.go`, also aus dem benannten Grund: weder Parser-Fehler noch `fuehreFall nicht gefunden` noch `panic`. Die vier Ausnahmen fängt je ein anderer, schon vorhandener Test; sie gehören nicht zur Klasse „Weitergabe des Ablaufs“, nur zu „Wertung des Prozesses in `fuehreFall`“. Offene Randformen der Klasse, die ich ohne Fänger fand: keine; die Weitergabe in anderen Funktionen als `fuehreFall` nennt der Kopf des Tests als ungeprüft, und die 5 Aufrufstellen außerhalb (vier verwerfen den Ablauf, `gmRepo.git` wertet `abgelaufen || code != 0` mit gleicher Wirkung wie Exit -1) stimmen mit `grep` über die Quelldateien überein (`faelle.go` 82, 89, 95; `zeichen_treiber.go` 119; `groessenmarke_treiber.go` 136).

Nicht ausführbar als Mutation: ein Tausch des Arguments `abgelaufen` gegen ein anderes Argument in `bewerteBau` und `bewerte` (die Typen `bool` und `int`/`string` unterscheiden sich, der Compiler weist ab); der Test braucht dafür keinen Fänger.

## Antworten auf die Prüffragen

- **Rot aus dem benannten Grund:** ja, siehe Tabelle; die Meldung nennt je Fall Datei, Zeile und Zusage.
- **`go vet`, `gofmt -l`, Lint:** `go vet .` Exit 0 und `gofmt -l` leer in der Kopie ohne Mutation; `make lint` Exit 0 (cyclop, gocognit, gochecknoglobals im Profil `.golangci.yml`; keine neue Ausnahme, keine `//nolint`-Direktive im Diff).
- **Fragilität:** siehe G-1; gegen Kommentar, Leerzeile, Zeilenumbruch und Umbenennung robust, gegen Zwischenvariable, Hilfsfunktion und Verschieben der Funktion in eine andere Datei rot.
- **Umfang (175 Zeilen der Datei, rund 104 neu):** angemessen für vier Zusagen je Paar (Aufruf einmal, Bindung, Argument, Vorkommen); Doppelung siehe G-3.
- **Kommentare (§3.7, §3.11):** Kopf von `TestFuehreFallReichtDenAblaufWeiter` nennt Zusage und Ungeprüftes; jede Aussage („zwei Stellen je Paar“, „einmal“) prüft der Test. Kopf von `TestJederAufrufNenntDieFallFrist` nennt Bezeichner-Aufrufe, Zähler und drei Formen als ungeprüft; `TestFuehreFall…` führt den Zähler „1 Zuweisung“ im Test, nicht im Kopf (kein Mehr an Zusage).
- **[SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) Nr. 4 und 5 d:** genau eine Lesart; Nr. 3 („führt zu einem ungültigen Lauf, nie zu einem Befund“), Randform 6 („wiederholt wird nur ein roter, Nr. 4“), Plan §1 und §6 gleich. Randform 6 sagt weiter „jeder Aufruf von `starte` und `starteStroeme` nennt sie“ und ist damit weiter als der Test; die *Grenze* im selben Vertrag nennt die drei Formen, die der Test nicht fängt, mit Folge. Kein Kopf- oder Hilfetext sagt mehr zu als Test oder Gate prüfen.
- **Plan nachgezogen (§3.9):** Tabelle Zeilen 8 und 9 (Zusage, Fänger, Mutation, Klasse), *Nachweis* (Menge, 2 Elemente, Mutationen an einem Element, das der Erstbefund nicht nannte, Raten 1 von 1 mit Last), §1 (Ziel, Umfang, DoD), §3 Datei-Tabelle, §6 (Negativ zu Zeile 2, Negativ zu den 5 anderen Aufrufstellen, Raster und Zusatztabelle, Größe, Sperre, Roadmap-Zeile). Meine Läufe bestätigen die Plan-Zahlen der Zeilen 8 und 9 (rot in 1 von 1; `5-mal, erwartet 4` und `4-mal, erwartet 2` gesehen).
- **Größe:** `make groessenmarke`: „ok, Messung 271 von Marke 1400“ (17 + 79 + 175, gleich dem `numstat` im Plan); Hinweis „Tabelle ergibt 1135, Schätzung 975“, kein Rot. Der Plan nennt in §6 offen und ehrlich, dass er die Marke nicht neu rechnet (mit 9 Zeilen ergäbe die Formel 1135 und Marke 1600), und gibt die Entscheidung an den Planner; eine Marke über der Messung (271 gegen 1400, Auslastung rund 0,19) bleibt die Grenze. Kein Verstoß gegen §3.14: es gibt keine Rückführung und keinen Schnitt nach dem Code.
- **Sperre:** Plan nennt `/home/builder/work/odata-recorder/.wt/heavy.lock`; ich habe ausschließlich diese genommen.
- **`harness/gegenproben.stempel`:** unverändert (`git diff 1697c650..HEAD` leer für den Pfad; Übergangsregel Nr. 2).
- **Rohbytes (§3.18):** gegengelesen `tools/gegenproben/frist_test.go`, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) und der Plan mit `LC_ALL=C grep -P`: keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein NBSP, kein BOM (erste Bytes `70 61 63` bzw. `23 20 53`), Nicht-ASCII nur Umlaute in Kommentaren von `frist_test.go` (Zeilen 13, 22, 23, 24, 105, 115). `make zeichen-check`: ok. `make a-check`: Exit 0.

## Was ich nicht prüfen konnte

- Den Lauf der Fälle der Stufe `gegenproben` (`make gegenproben`, `make gates`): nicht gefahren, nicht Gegenstand der Fix-Runde (die Fälle sind unverändert; Code in `main.go` unverändert gegenüber dem ersten Report).
- Eine Rate je `-count`: kein Fänger der Klasse hängt an Wahrscheinlichkeit oder Nebenläufigkeit; je Mutation ein Prozess.
- Das Verschieben von `fuehreFall` in eine andere Datei (G-1) nur aus dem Quelltext des Tests abgelesen, nicht als Mutation gefahren.
- Die Messung nach *Zeitbudget* Nr. 1 und die Eichung (Nr. 7): nicht Gegenstand, Plan §6 offen mit Adresse Planner.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tools/gegenproben/frist_test.go` · `TestFuehreFallReichtDenAblaufWeiter` und Hilfen | geprüft, ohne Befund (G-1, G-3 als INFO) |
| `tools/gegenproben/frist_test.go` · `TestJederAufrufNenntDieFallFrist`, `TestFallFristIst120Sekunden` | geprüft, ohne Befund (Kopf sagt nur zu, was der Test fängt) |
| `tools/gegenproben/main.go` · `fuehreFall`, `bewerteBau`, `bewerte` | geprüft, ohne Befund (unverändert; Mutationen der Wertung rot) |
| [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) · Randform 6, *Zeitbudget* Nr. 1, 3 bis 5, *Grenze* | geprüft, ohne Befund (G-2 an der ADR) |
| Plan §1, §3 (Zeilen 1 bis 9, Negative, *Nachweis*), §4, §6 | geprüft, ohne Befund (Marke nicht neu gerechnet, ehrlich benannt) |
| Suppression (`//nolint`, Ausnahme in `.golangci.yml`), Gates | geprüft, ohne Befund: keine im Diff, keine Lockerung (`AGENTS.md` §3.6) |
| `harness/gegenproben.stempel`, `harness/mk/gegenproben.mk` | geprüft, ohne Befund (unverändert) |
| Rohbytes (`AGENTS.md` §3.18) | geprüft, ohne Befund |

## Summary

0 HIGH, 0 MEDIUM, 0 LOW, 4 INFO (G-1 verhaltensgleicher Umbau macht den Test rot, G-2 ADR-Satz allgemeiner als die Spezifikation, G-3 doppelte Durchlaufform, G-4 Roadmap-Klammer adressiert). F-1 und F-2 sind gegen die Klasse behoben (31 Mutationen, 1 von 1 Prozess je Mutation, Last 7,2 bis 10,8); F-3, F-5, F-6 behoben, F-4 offen mit Adresse Planner.

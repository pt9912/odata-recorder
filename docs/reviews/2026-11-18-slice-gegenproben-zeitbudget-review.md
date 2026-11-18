# Review-Report: slice-gegenproben-zeitbudget — 2026-11-18

**Review-Art:** Code (gegen Plan, Entscheidungen und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers).

**Gegenstand:** `slice-gegenproben-zeitbudget`, Zweig `slice-gegenproben-zeitbudget`, HEAD `49dc5cb6`, Diff `a9287eb9..HEAD`: Start-Move, Roadmap-Marker, Vertrag [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) (Randform 2a und 6, *Zeitbudget* Nr. 1 bis 7, *Stand*, *Grenze*), Marke 1400, `tools/gegenproben/main.go` (`bewerteBau`), `main_test.go`, `frist_test.go`, Plan §3 und §6.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `49dc5cb6` · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-18

**Eingangs-Kontext:**

- Plan `slice-gegenproben-zeitbudget` (DoD §2, §3 Tabelle *Zusage · Fänger* Nr. 1 bis 7 und Absatz *Nachweis*, §4 Block, §6 Raster)
- [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (Accepted), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md)
- [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md); `LH-QA-06`
- `AGENTS.md` §3.6 bis §3.18

**Vorgehen.** Keine Einschätzung des Implementers übernommen. Jede Zeile der Tabelle (und Elemente, die der Plan nicht nannte) mit eigener Mutation in Kopien aufgelöst (`git archive HEAD`, Scratchpad-Unterordner rev-zb, nie im Zweig), je Mutation ein Prozess in einem Container ohne Netz aus dem Image `odata-recorder:race`, `--memory=4g --memory-swap=4g --cpus=4`, `go vet` und `go test -count=1 ./...`, unter der Sperre `heavy.lock`. Messender: der Reviewer; Last = erstes Feld von `/proc/loadavg` unmittelbar vor dem Start.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Das akzeptierte Negativ „`fuehreFall` reicht `abgelaufen` weiter, ein Lauf bis zur Frist dauert 120 s“ trägt nicht: `fallFrist` ist zwar eine Konstante, aber ein Fänger über den Quelltext (go/parser, dieselbe Technik wie `TestJederAufrufNenntDieFallFrist`) prüft in `fuehreFall` das letzte Argument der Aufrufe von `bewerteBau` und `bewerte` und braucht keine 120 s. Die Mutationen `bewerteBau(code, out, abgelaufen && false)` und, auf der Gate-Seite, `bewerte(f, code, out, errOut, abgelaufen && false)` bleiben grün (je 1 von 1 Prozess, Last 12,6 und 13,7); beide färbt der Probe-Fänger rot, die unmutierte Kopie bleibt grün. Zeile 3 der Tabelle (Randform 6: ein Bau über der Frist ist rot mit `Zeit überschritten`) und Zeile 7 (Gate über der Frist) sind damit nur auf Ebene der Funktion gefangen, nicht im Aufruf, den das Programm tatsächlich macht; verlorenes Argument heißt: ein gerissener Bau meldet unverändert `Bau des Binary scheitert` (Exit -1), also genau das Verhalten vor dem Slice. Das Negativ nennt die Gate-Seite nicht, und die *Grenze* in [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) nennt sie ebenfalls nicht. | `AGENTS.md` §3.15 (Zusage ohne Fänger; ein Negativ mit nicht tragendem Grund), §3.10 („sucht der Plan einen Haltepunkt“) | `docs/plan/planning/done/slice-gegenproben-zeitbudget.md` · „reicht `abgelaufen` des Baus an `bewerteBau` weiter“ (§3, Zeile 118); `tools/gegenproben/main.go` · `bewerteBau(code, out, abgelaufen)` (Zeile 249) und `bewerte(f, code, out, errOut, abgelaufen)` (Zeile 255); `harness/vertraege/SPEC-102-gegenproben.md` · „Ungeprüft: dass `fuehreFall` den Ablauf des Baus an `bewerteBau` weiterreicht“ | ja: `make gegenproben` wird nicht rot, `go test` im Paket bleibt grün; der Probe-Fänger wird rot | Negativ mit nicht tragendem Grund (Verdrahtung ungefangen, obwohl ein Quelltext-Fänger möglich ist) |
| F-2 | MEDIUM | `TestJederAufrufNenntDieFallFrist` fängt eine schmalere Menge, als Kommentar und Vertrag zusagen („jeder Aufruf von `starte` und `starteStroeme`“): die Menge entsteht aus Aufrufen mit dem Namen als Bezeichner, und das erste Argument gilt, wenn es der Bezeichner `fallFrist` ist. Grün bleiben: ein Aufruf über einen Funktionswert (`var starteFn = starte`, Aufruf `starteFn(3600000000000, …)` bei einer zusätzlichen regulären Stelle, 1 von 1 Prozess grün, Last 12,2), ein im Rumpf von `fuehreFall` überdecktes `fallFrist := 3600 * time.Second` (grün, Last 12,2; `TestFallFristIst120Sekunden` liest die Paket-Konstante), und ein Prozess, der nicht über `starte` startet (`exec.CommandContext` an anderer Stelle, nicht gemessen, aus dem Quelltext des Tests ablesbar). Der Funktionswert allein wird nur durch den Zähler `mindestens = 7` rot („6 Aufrufstellen gefunden“), also nicht aus dem benannten Grund. Weder Plan §3 noch die *Grenze* in [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) nennen diese Prüfgrenze mit Grund und Folge. Die Menge der heute 7 Stellen selbst ist vollständig (`main.go` 2, `faelle.go` 3, `groessenmarke_treiber.go` 1, `zeichen_treiber.go` 1; keine weitere Stelle in den `*_treiber.go` oder `faelle.go`, `exec` nur in `starteStroeme` und für `LookPath`) und jedes der 7 Elemente einzeln rot gesehen. | `AGENTS.md` §3.17 (unbeschränkte Menge nennt Prüfgrenze mit Grund und Folge), §3.11 (Kommentar sagt nur zu, was ein Test prüft) | `tools/gegenproben/frist_test.go` · „jeder Aufruf von starte und starteStroeme in den Quelldateien des Pakets“ (Zeile 87); `harness/vertraege/SPEC-102-gegenproben.md` · „`TestJederAufrufNenntDieFallFrist` (jeder Aufruf von `starte` und `starteStroeme` in den Quelldateien nennt `fallFrist`)“ | ja: die zwei Mutationen in einer Kopie | Fänger schmaler als Zusage, Prüfgrenze nicht benannt |
| F-3 | LOW | *Zeitbudget* Nr. 1 nennt als Bedingung „lokal vorhandene Images“, nicht mehr den kalten Cache, den [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) und Plan §1 weiter nennen; die Zelle „Cache und Images“ des Rasters trägt „die Stufe baut immer ohne Cache“, was Nr. 1 nicht sagt (die Stufe `gegenproben` baut wegen `--no-cache-filter` kalt, die anderen Stufen nicht). Ob der Cache der übrigen Stufen zur Messgröße gehört, steht nirgends. | `AGENTS.md` §3.16 (eine Zelle nennt nur einen Ort, dessen Wortlaut die Entscheidung trägt) | `harness/vertraege/SPEC-102-gegenproben.md` · „mit lokal vorhandenen Images (ein Lauf, der ein Image zieht, misst das Netz und zählt nicht)“ (Nr. 1); Plan §6 · Raster, Spalte „Cache und Images“ | nein | Raster-Zelle ohne tragenden Wortlaut |
| F-4 | LOW | Die Zeile von `slice-gegenproben-zeitbudget` in der Roadmap trägt weiter den Zustand `open/` in der Klammer („(`open/`, Linie C, Kettenslice, **fünfter der Kette**“), obwohl ihr Link auf `in-progress/` zeigt und der Slice dort liegt; der Start-Commit hat Link und Marker „Nichts in Arbeit“ nachgezogen, den Zustand in der Klammer nicht. | `AGENTS.md` §3.9 (jede Stelle nachziehen, die die Aussage trägt), §3.7 (Zustandsfeld) | `docs/plan/planning/in-progress/roadmap.md` · „(`open/`, Linie C, Kettenslice, **fünfter der Kette**“ (Zeile 30) | nein | Zustandsfeld nicht nachgezogen |
| F-5 | INFO | Plan §3 nennt als Sperre für schwere Läufe `/tmp/odata-recorder-heavy.lock`, der Auftrag dieser Runde die Sperre unter dem Verzeichnis der Arbeitsbäume; wer sich nach dem Plan richtet, sperrt gegen niemanden, der die andere Datei nimmt. Adresse: Planner. | Maintainability | Plan §3 · „schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`“ (Zeile 91) | nein | zwei Sperrdateien für denselben Zweck |
| F-6 | INFO | *Zeitbudget* Nr. 4 sagt, ein ungültiger Lauf werde wiederholt; Nr. 5 d lässt einen grünen ungültigen Lauf grün und den Stempel schreiben und sagt nichts zur Wiederholung. Drei grüne ungültige Läufe in Folge gehen nach Nr. 4 als Befund an den Planner, obwohl nichts rot ist. Der Wortlaut lässt beide Lesarten zu; Adresse: Architect (nur eine Klarstellung, kein Befund gegen den Code). Die Entscheidung des Architect zu Nr. 5 d (Stempel hängt am grünen Lauf, nicht an der Last) ist mit der *Grenze* („belegt nur, dass ein Lauf zu dieser Prüfsumme grün war“) stimmig. | Maintainability | `harness/vertraege/SPEC-102-gegenproben.md` · „Ungültiger Lauf: (d) grün: er bleibt grün“ (Nr. 5 d) neben „Ein ungültiger Lauf … wird wiederholt“ (Nr. 4) | nein | Wortlaut mit zwei Lesarten |

---

## Auflösung der Tabelle *Zusage · Fänger* (§3.15, §3.10)

Eigene Messung, 1 Prozess je Mutation (Klasse *andere*, kein Wahrscheinlichkeits-Fänger), Last vor dem Start in Klammern; Lauf ohne Mutation grün (Last 11,2).

| Nr. | Zusage | Mutation (in Kopie) | Ergebnis | roter Test, aus dem Grund |
|---|---|---|---|---|
| 1 | `fallFrist` 120 s | 121 s (9,5); 119 s (14,5) | rot in 1 von 1 und 1 von 1 | `TestFallFristIst120Sekunden`: `fallFrist = 2m1s, erwartet 120s` (121 s mit Meldung gesehen, 119 s nur als FAIL des Pakets) |
| 2 | jeder Aufruf nennt `fallFrist` | je 1 der 7 Stellen mit `3600000000000` statt `fallFrist` (`main.go` 248 (8,8) und 254 (14,1), `faelle.go` 82 (8,3), 89 (8,0), 95 (10,2), `groessenmarke_treiber.go` 136 (11,2), `zeichen_treiber.go` 119 (11,1)) | 7 von 7 rot | `TestJederAufrufNenntDieFallFrist`: `<Datei>:<Zeile>: Aufruf von … nennt fallFrist nicht als erstes Argument` |
| 2 (Grenzen, F-2) | wie 2 | Funktionswert mit zweiter regulärer Stelle (12,2); Überdeckung `fallFrist :=` in `fuehreFall` (12,2) | **grün**, grün | Funktionswert ohne zweite Stelle: rot nur über den Zähler („6 Aufrufstellen“, 11,1) |
| 3 | Bau über Frist: `Zeit überschritten` | Zweig `abgelaufen` entfällt (14,8); `code != 0` vor `abgelaufen` (14,4) | rot, rot | `TestBewerteBau/Ablauf` (und `Ablauf_mit_Exit_0`): nennt `Zeit überschritten` nicht |
| 3 (Verdrahtung, F-1) | `fuehreFall` reicht den Ablauf des Baus weiter | `bewerteBau(code, out, abgelaufen && false)` (12,6) | **grün** | kein Test; Probe-Fänger rot |
| 4 | Exit ungleich 0: Text und letzte Zeilen | `+ letzte(out)` entfällt (10,3); `+ out` statt `letzte(out)` (11,1) | rot, rot | `TestBewerteBau/Exit_1_mit_langer_Ausgabe`, `…/Exit_2`, `…/nicht_startbar`, `TestFuehreFallBauScheitertRuftKeinGate`, `TestFallMitBauMeldetScheiternDesBaus` (zweite Mutation: nennt `z1\n` / `z2\n`) |
| 5 | Exit 0 nicht rot | `code >= 0` (11,3); `code > 0` (11,6) | rot, rot | `TestBewerteBau/Exit_0` und `…/nicht_startbar`, `TestFallMitBauBautVorDemGate`; `> 0` allein: `…/nicht_startbar` (Grund leer) |
| 6 | Bau scheitert: kein Gate | Rückkehr nach Rot des Baus entfällt (`_ = rot`) (13,7) | rot | `TestFuehreFallBauScheitertRuftKeinGate`: `Befehl 1-mal … gerufen, erwartet 0`; `TestFallMitBauMeldetScheiternDesBaus` |
| 7 | Gate über Frist: `Zeit überschritten` | `case abgelaufen` in `bewerte` wird `case false` (13,8) | rot | `TestBewerte/Frist`: `erwartet Exit 1, war -1` |
| 7 (Verdrahtung, F-1) | `fuehreFall` reicht den Ablauf des Gates weiter | `bewerte(…, abgelaufen && false)` (13,7) | **grün** | kein Test; Probe-Fänger rot |

Probe-Fänger zu F-1 (nur in der Kopie, nicht im Zweig): ein Test parst `main.go`, sucht in `fuehreFall` die Aufrufe von `bewerteBau` und `bewerte` und verlangt den Bezeichner `abgelaufen` als letztes Argument; Lauf mit den zwei Mutationen rot (Last 13,4 und 13,2, je 1 von 1), unmutiert grün (12,4).

Nicht aufgelöst: die Prozessgruppe (`Setpgid`, `Cancel`) von Zeile 7, weil die Last danach auf 44 bis 86 stieg und schwere Läufe nicht mehr zulässig waren; der Test der Hülle (`TestStarteBeendetDieProzessgruppeBeiFristablauf`, Grenze 3 s gegen `WaitDelay` 5 s) ist unverändert und nicht Gegenstand des Diffs.

## Antworten auf die Prüffragen des Auftrags

- **(a)** Negativ nicht echt, siehe F-1: Der Grund „ein Lauf dauert 120 s“ gilt für einen Verhaltenstest, nicht für einen Quelltext-Fänger; auch die Gate-Seite fehlt im Negativ.
- **(b)** Prüfgrenze nicht benannt, siehe F-2. Aufrufstellen: 7, wie im Plan; in `*_treiber.go` und `faelle.go` fehlt keine.
- **(c)** `bewerteBau`: alle drei Zweige, Reihenfolge `abgelaufen` vor `code != 0`, Exit -1 ohne Ablauf (nicht startbar, `starteStroeme` liefert -1 und `err.Error()`) und `letzte(out)` sind gefangen (Tabelle). Randform 2a und 6 passen zum Code: der Bau meldet `Zeit überschritten` oder `Bau des Binary scheitert` samt letzten Zeilen; Aufrufe von Treiber und `git` verwerfen den Ablauf (`faelle.go`, `groessenmarke_treiber.go` `scheitert`, `zeichen_treiber.go`), wie Randform 6 es sagt. Die Zusage der Verdrahtung ist nicht gefangen (F-1).
- **(d)** Die Vorschrift Nr. 1 bis 7 ist in sich stimmig bis auf F-3 und F-6. Nachgemessen: `nproc --all` 20, `nproc` 3 in dieser Umgebung (Quote); im Container mit `--cpus=4` zeigen `nproc --all` 20, `nproc` 20 und `/proc/loadavg` denselben Host-Wert wie außerhalb; die Aussage von Nr. 2 hält. Gleichheit gültig (Nr. 3) und die Eigenlast der Lesung am Ende sind als Prosa mit Fehlerrichtung „ungültig“ benannt. Die Gesamtzeit prüft kein Programm, und das steht in [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) (*Zeitbudget*, Satz 2) und in [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (Fitness Function); die Hilfezeile von `make gegenproben` und `harness/README.md` sagen nichts zur Zeit zu.
- **(e)** Nichts im Diff führt Wiederholen bis grün ein: kein Schleifenkonstrukt in `main.go`, `harness/mk/gegenproben.mk` unverändert; Nr. 5 c verbietet es für einen gültigen Lauf ausdrücklich, Nr. 5 e erlaubt die Wiederholung nur für einen ungültigen roten Lauf, dessen rote Fälle alle `Zeit überschritten` tragen, und Nr. 4 begrenzt sie auf drei.
- **(f)** Größe: `git diff --numstat a9287eb9..HEAD` über die Pfade des Blocks 17 + 79 + 71 = 167, `make groessenmarke`: „ok, Messung 167 von Marke 1400“. `harness/gegenproben.stempel` unverändert (Übergangsregel Nr. 2). Rohbytes der drei Go-Dateien, [SPEC-102](../../harness/vertraege/SPEC-102-gegenproben.md) und des Plans mit `LC_ALL=C grep -P` gegengelesen: keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein NBSP, kein BOM (erste Bytes `70 61 63`), nicht ASCII sind nur Umlaute; `make zeichen-check` ok. `make a-check`: 0 Befunde (Hinweis zu Dateien ohne Schicht unverändert). `make lint` Exit 0. `make gates` zu HEAD: Nachweis unter `.harness/state` trägt `49dc5cb6` und den Hash des Arbeitsbaums, den `working-tree-hash.sh` heute wieder liefert; ein eigener Lauf ist nicht erfolgt. `make docs-check` mit diesem Report: 0 Befunde (Stand des Aufrufs nach der letzten Fassung).

## Was ich nicht prüfen konnte

- **Lauf der Fälle der Stufe `gegenproben` (Schritt 16 des Dockerfile):** gefahren als `docker build` der Stufe mit den Build-Args aus `harness/mk/gegenproben.mk` (eigener Tag, kein Stempel, kein `make gegenproben`), Beginn bei Last 10,4. Schritt 10 (`gofmt`, `go vet`, `go test ./...`, Bau) war grün (`ok … 0.977s`). Der Lauf der 216 Fälle ist **ungültig**: ich hatte dem Build `--memory=4g` gesetzt, Compiler wurden mit `signal: killed` beendet (15 rote Fälle, davon `race/*` und `build/ohne-trimpath`), dazu stieg die Last auf 86 (neun Fälle `lint/*`, `build/basis`, `build/falsche-version` mit `Zeit überschritten`). Wiederholt habe ich nicht, weil die Last danach zeitweise 44 bis 86 war; der Befund gilt nicht für den Code. Der Verifier fährt die Stufe ohne Speichergrenze bei Last unter 20.
- Eine Messung von `time make gegenproben` und die Eichung (*Zeitbudget* Nr. 7): nicht Gegenstand, steht in Plan §6 als offen mit Adresse.
- Die Mutation an `Setpgid` und `Cancel` (siehe oben) und ein Prozess außerhalb von `starte` (F-2, nur gelesen).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tools/gegenproben/main.go` · `bewerteBau` (drei Zweige, Reihenfolge, Exit -1, `letzte`) | geprüft, ohne Befund (Verdrahtung: F-1) |
| `tools/gegenproben/main_test.go` · `TestBewerteBau`, `TestFuehreFallBauScheitertRuftKeinGate` (Fake hält Argumente fest, zählt `Befehl` und `Aufraeumen`) | geprüft, ohne Befund |
| `tools/gegenproben/frist_test.go` · `TestFallFristIst120Sekunden` (beide Seiten der Grenze) | geprüft, ohne Befund |
| `harness/vertraege/SPEC-102-gegenproben.md` · Randform 2a und 6 gegen den Code | geprüft, ohne Befund |
| `harness/vertraege/SPEC-102-gegenproben.md` · *Zeitbudget* Nr. 1 bis 7 gegen [ADR-0013](../plan/adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (keine Anhebung, keine Lockerung nach `AGENTS.md` §3.6, Soll als nicht geprüft benannt) | geprüft, ohne Befund (F-3, F-6 am Wortlaut) |
| Plan §3 Nachweis: Zahlen der Tabelle gegen eigene Läufe | geprüft, ohne Befund: jede genannte Mutation rot aus dem genannten Grund |
| Plan §4 Block und Marke (Metrik, 167 von 1400) | geprüft, ohne Befund |
| Plan §6 Raster und Zusatztabelle (keine leere Zelle, Orte lösen auf) | geprüft, ohne Befund (F-3 zu einer Zelle) |
| Kommentare in `main.go` und den Tests (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund außer dem Kopf von `TestJederAufrufNenntDieFallFrist` (F-2) |
| `harness/gegenproben.stempel`, `harness/mk/gegenproben.mk`, `harness/README.md` | geprüft, ohne Befund (unverändert, keine Zusage zur Zeit) |
| Suppression (`//nolint`, Ausnahme in `.golangci.yml`) | geprüft, ohne Befund: keine im Diff |
| Rohbytes (`AGENTS.md` §3.18) | geprüft, ohne Befund |

## Summary

0 HIGH, 2 MEDIUM (F-1 Verdrahtung des Ablaufs von Bau und Gate ungefangen, das Negativ trägt nicht; F-2 Prüfgrenze der Aufrufmenge nicht benannt), 2 LOW (F-3, F-4), 2 INFO (F-5, F-6). Alle sieben Zeilen der Tabelle sind auf Ebene der Funktion rot aus dem genannten Grund; die Befunde betreffen, was zwischen den Funktionen und außerhalb der Menge liegt.

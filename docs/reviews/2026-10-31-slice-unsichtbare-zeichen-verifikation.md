# Verifikations-Report: slice-unsichtbare-zeichen — 2026-10-31

**Rolle:** Verifier, frischer Kontext (Frage: "Bauen wir es richtig?" gegen Plan, DoD, `SPEC-233`, `SPEC-102` Fälle 131 bis 155 und `AGENTS.md` §3.9 bis §3.18). Kein Review, keine Reparatur: Code, Plan und Vertrag sind unverändert. Eingang: der [Review-Report](2026-10-31-slice-unsichtbare-zeichen-review.md) (F-1 bis F-12) und der Bericht des Implementers nur als Eingang, nicht als Beweis.

**Gegenstand:** Zweig `slice-unsichtbare-zeichen`, HEAD `5c471f2` (Inhalts-Stand des Implementers `1784a38`, Fix-Commit `5455e8e`, danach der Bestandsfix am Report von slice-baseline-import), Basis des Blocks `3cdbdca`. Plan: `docs/plan/planning/in-progress/slice-unsichtbare-zeichen.md`. Arbeitsbaum vor und nach den Läufen ohne Änderung (`git status` leer).

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Methode.** Keine Zahl und keine Mutation des Implementers oder Reviewers übernommen. Scratchpad `verif-unsichtbar/` (flüchtig, nichts im Repo außer diesem Report). `make gates` im Zweig, Exit-Code aus Datei; die Stufen `test`, `zeichen`, `lint`, `race` ein zweites Mal mit `--no-cache-filter` (je Exit 0). Ein eigenes Probe-Paket mit eigenem Orakel (nie das Programm): Z aus den Bereichen des Vertrags (Nr. 7 a, 253 Elemente von Hand gezählt), die Gültigkeit von UTF-8 aus der Tabelle der Unicode-Norm von Hand. Mutationen als Kopien von `tools/zeichen` bzw. `tools/harness/zeichen.sh` im Scratchpad, genau eine Änderung je Kopie, `go test -count=1 ./...` im Image der Stufe `zeichen` bzw. die Fälle 131 bis 155 im Image der Stufe `gegenproben`; jeder Lauf unter `flock /tmp/odata-recorder-heavy.lock`, `--network none --memory=4g --memory-swap=4g --cpus=4`, Quelle `:ro`. Das echte Gate mit echtem `docker` an zwei Klonen des Zweigs. Rohbytes (§3.18) mit `env LC_ALL=C grep -aP`; das Verfahren ist an `git show 4bbe19d^:docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md` bestätigt (Trefferzahl für die Bytes C2 A0: 1 Zeile, wie erwartet). Dieser Report nennt Zeichen nur als `U+XXXX`.

---

## 1. Verdikt

**NICHT BESTANDEN.** Kein HIGH, und das Gate verhält sich in allen geprobten Klassen nach dem Wortlaut des Vertrags: `make gates` Exit 0, ein eingepflanztes Zeichen macht es rot, mein Orakel stimmt mit dem Programm über 1 222 980 Byte-Folgen, 1 112 064 Skalarwerte und 2596 Dateien überein (Abschnitt 4). Zwei Befunde sind Fix-Gründe nach der Leitlinie (Zusage ohne Fänger, §3.15, §3.17): **V-1 (MEDIUM)**, die Zusage "der Treiber wechselt in die Wurzel" (`SPEC-233` Nr. 2 a) hat keinen Fänger, und **V-2 (MEDIUM)**, der Fix zu F-1 deckt die Klasse "keine zweite Ausnahme" nicht: 11 von 11 Ausnahmen für einen Pfad aus zwei Komponenten, einen Dateinamen oder eine Endung lassen alle 29 Tests grün. Beide sind klein (je ein Fall bzw. eine Schleife über die Menge); eine zweite Verifikation ist nicht nötig, der Planner schließt nach den Fixes, wenn die genannten Mutationen rot sind.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| V-1 | MEDIUM | **Zusage ohne Fänger: Wechsel in die Wurzel.** `SPEC-233` Nr. 2 a: "Der Treiber wechselt in die Wurzel des Git-Arbeitsbaums". Ohne die Zeile `cd "$wurzel"` bleiben alle 25 Fälle grün (die Fälle rufen den Treiber im Wurzelverzeichnis ihres Repos); aufgerufen aus einem Unterverzeichnis listet `git ls-files` nur dessen Teilbaum, und der Treiber prüft diesen allein. Mit echtem `docker` im Klon mit 7 eingepflanzten Befunden: Aufruf der mutierten Fassung aus `docs/` endet mit Exit 0 und `zeichen: ok, 443 Datei(en) geprüft`, aus der Wurzel mit 7 Befunden. Die Fassung des Zweigs aus `docs/` meldet richtig 7 Befund(e). Im Betrieb ruft `make` immer aus der Wurzel, die Wirkung ist heute null; die Zusage steht im Vertrag und in Zeile 2 der Tabelle, die Mutation steht in keiner der beiden Mutationslisten. Menge: Arbeitsverzeichnis des Aufrufs (Wurzel, Unterverzeichnis erster und tieferer Stufe). | Mutation `cd` (Zeile entfernt): 25 von 25 Fällen grün; Probe e2e mit echtem `docker`; Treiber des Zweigs aus `docs/`: 7 Befunde | Implementer: ein Fall (Fall 151 oder 155) ruft den Treiber aus einem Unterverzeichnis der Tiefe 1 und 2 und erwartet dieselbe Menge |
| V-2 | MEDIUM | **Fix zu F-1 deckt die Klasse nicht.** Die Zusage ist "es gibt keine Ausnahme außer dem Teilbaum `.harness/baseline`" (`SPEC-233` Nr. 5 a, 5 e, 5 f). `TestKeineWeitereAusnahme` führt 169 Namen als **einzelne** Komponente unter 10 Eltern-Pfaden, `TestAusschluesse` 22 Namen, `TestDateiarten` 21 Dateinamen. Folgende Ausnahmen in `sammle` bzw. `datei` lassen alle 29 Tests grün (die Fälle 131 bis 155 gegen diese Kopien nicht gefahren; die Wirkung ist an zwei Kopien, `path-reviews` und `ext-golden`, geprüft: die Datei im Pfad wird nicht gemeldet): (a) **Pfad aus zwei oder mehr Komponenten**, jede Komponente allein in der Menge der 169 Namen: `docs/reviews`, `docs/plan/planning/done`, `tools/zeichen`, `tools/gegenproben`, `harness/vertraege` (5 von 5 grün). Die Ausnahme für `docs/reviews` ist die naheliegende, denn dort fand das Gate seine zwei Befunde. (b) **Dateiname oder Endung** außerhalb der 21 Namen: `AGENTS.md` samt `README.md`, `.golden`, `.csv`, `.lock`, `.snap` samt `.pem`, `.html` samt `.svg` und `.css` (6 von 6 grün). Der Plan benennt als akzeptiertes Negativ nur "ein Name außerhalb dieser Mengen" (Zeile 22); der Pfad `docs/reviews` besteht aus Namen innerhalb der Mengen und ist trotzdem ungeprüft, Endungen nennt kein Negativ. Menge (aus dem Wortlaut, nicht aus dem Befund): jeder Verzeichnispfad des Repos (Tiefe 1 bis 5) mal Dateiart, dazu Endung und Dateiname; die Pfadmenge des Repos ist endlich (alle `git ls-files`-Verzeichnisse) und als Schleife prüfbar, die Endungen sind unbeschränkt (Prüfgrenze mit Grund und Folge nennen, §3.17). | Mutationen `path-reviews`, `path-done`, `path-tz`, `path-gp`, `path-harn`, `ext-agents`, `ext-golden`, `ext-csv`, `ext-lock`, `ext-snap`, `ext-html`: je `go test` Exit 0; Wirkungsprobe der Kopien `path-reviews` und `ext-golden` (die Datei fehlt in den Befunden) | Implementer: Fänger als Schleife über Pfade (alle Verzeichnispfade des Repos als Eltern, nicht nur 10 erfundene) und über eine Liste von Endungen und Dateinamen; Architect/Planner: Prüfgrenze für Endungen in Zeile 21/22 des Plans mit Grund und Folge |
| V-3 | LOW | **Plan folgt `5c471f2` nicht** (§3.9). Seit `5c471f2` nennt Zeile 45 des alten Reports das Zeichen als Text `<U+00A0>`; der Plan sagt weiter (§3, Zeile "Bestand der Prüfmenge"), jedes der zwei Zeichen sei "durch ein Leerzeichen ersetzt", und §6 (Zeile zur berichtigten Zeile 45) führt den Punkt mit Ausgang *weiter offen*, Adresse User, obwohl er gelöst ist. | `grep` im Plan; `cmp` Abschnitt 5 | Planner (Closure: Zeile und Ausgang nachziehen) |
| V-4 | LOW | Betriebsregel für Repo-Dateien (Plan §3: "nur `Edit`/`Write`, kein `sed -i`") und ihre Ausnahme stehen an zwei Orten: die Regel in §3, die Ausnahme (das Schreibwerkzeug setzt ein NBSP still in ein Leerzeichen um, Ausgang *eingetreten*) nur in §6. Wer §3 liest, hat keine Erlaubnis für ein Werkzeug, das Bytes schreibt. Bewertung des Mittels in Abschnitt 6. | Plan §3 und §6 | Planner (Satz in §3 nachziehen; Diff-Vorschlag für die Betriebsregel der Welle, nicht `AGENTS.md`) |
| V-5 | INFO | Von mir gefahrene Mutationen, die der Vertrag nicht entscheidet und die grün blieben: Wurzel als Symlink auf ein Verzeichnis (`Lstat` statt `Stat`: Wurzel würde abgelehnt), Name des nicht lesbaren Stammverzeichnisses (`.` oder leer), leere Umgebungsvariable `ZEICHEN_IMAGE` (leer wie ungesetzt oder leerer Image-Name), Dateigröße ab 4 MiB (`io.LimitReader`; Nr. 6 c sagt "ohne Grenze", der Fänger liest 3 MiB), `docker rm -f` ohne Umleitung (gibt den Namen auf stdout aus; der Ersatz tut es nicht). Keine ist falsches Verhalten; die Größe ist eine unbeschränkte Menge ohne benanntes Negativ. | `lstatroot`, `rootpunkt`, `imgleer`, `limit`, `noerr_rm` | Architect (entscheiden oder als Negativ eintragen) |
| V-6 | INFO | `make zeichen-check` auf dem Inhalt von `main` (`aac630e`, per `git archive` in ein frisches Repo, Treiber des Zweigs): Exit 1 mit genau den **2 Befunden** der zwei NBSP in `docs/reviews/2026-10-25-slice-baseline-import-verifikation-3.md` Zeile 45; 1003 Dateien im Index. Der Zweig berichtigt sie; nach dem Merge ist das Gate grün, solange kein anderer Zweig ein Zeichen aus Z einbringt (parallele Worktrees prüfen vor dem Merge). | `main.out` des Scratchpads | Planner (Merge-Reihenfolge) |
| V-7 | INFO | `harness/gegenproben.stempel` ist unverändert; die Eingaben haben sich geändert. Closure-Bedingung (Plan §5) ist `make gegenproben` grün im Hauptbaum, nicht gefahren (Auftrag). Der Modus von `tools/harness/zeichen.sh` ist inzwischen 100755 (F-11 erledigt). | `git diff --numstat` (Stempel 0); `git ls-files -s` | Planner (Closure) |
| V-8 | INFO | Der Kopf von `zeichen.sh` verweist für "SIGKILL und das Fenster zwischen create und trap" auf `SPEC-233` Nr. 16 g; Nr. 16 g nennt das Fenster nicht. Der Punkt ist im Plan §6 mit Ausgang *offen*, Adresse Architect, geführt (richtig nach §3.13). | `tools/harness/zeichen.sh` Kopf; Plan §6 | Architect |

## 3. DoD, Punkt für Punkt

| DoD-Punkt | Befund | Beleg |
|---|---|---|
| Gate `zeichen-check` hängt an `GATE_CHECKS`, rot bei Z oder ungültigem UTF-8; jede Zusage hat eine Zeile mit Mutation | erfüllt für das Verhalten, mit V-1 und V-2 für zwei Zusagen | `make gates` Exit 0 (893 Dateien); eingepflanzt (Klon, echtes `docker`): 7 Befunde in 7 Dateien, `make zeichen-check` Exit 2 (make), Treiber Exit 1; Tabelle hat 38 Zeilen (nachgezählt) |
| Verdrahtung `GATE_CHECKS` (akzeptiertes Negativ, Beleg `make -n -k gates`) | erfüllt (Beleg) | Kopie von Makefile und `harness/mk` ohne die Zeile `GATE_CHECKS += zeichen-check`: 0 Treffer für `--target zeichen` und `zeichen.sh`; mit der Zeile: 2 |
| Kopie von `tools/zeichen` in die Stufe `lint` | erfüllt, nicht gefangen (wie zugesagt) | `Dockerfile` Zeile `COPY tools/zeichen tools/zeichen`; der Lauf der Stufe `lint` ohne Cache zeigt den Schritt und Exit 0; das Gate sieht die Zeile nicht |
| Gegenprobe Fälle 131 bis 155, Zusatzzeit gemessen und in §7 berichtet | erfüllt (Messung in `SPEC-102` und §6), §7 folgt bei der Closure | Abschnitt 5 |
| Bestand sauber oder berichtigt (1 Datei) | erfüllt | `make zeichen-check` grün; Rohbytes der Datei: 0 Treffer für C2 A0; Zeile 45 gleich dem Original mit `<U+00A0>` an den zwei Stellen (`cmp`) |
| `make gates` grün | erfüllt | Exit 0 aus Datei; `groessenmarke ok, Messung 2884 von Marke 4900`, Hinweis "Tabelle ergibt 3720, Schätzung 3560" |
| Review durchgeführt | erfüllt | [Review-Report](2026-10-31-slice-unsichtbare-zeichen-review.md) |
| Doku-Update (`harness/README.md` Sensors, Zeile von `make gegenproben`, *Nicht behauptet*; `SPEC-102` Nachzüge; Index der Verträge) | erfüllt | Zeilen gelesen; die Zellen sind unter 200 Zeichen (`make docs-check` grün im Lauf von `make gates`); `SPEC-102` nennt "38 bis 155" an zwei Stellen, `zeichen.sh` und `tools/zeichen` stehen in `EINGABEN` |
| Closure-Notiz, Register, Ausgänge in §6, drei Paarungen | steht aus (Closure des Planners) | §7 "wird bei der Closure gefüllt"; §6: jede Zeile trägt einen Ausgang; offene Zeilen mit Adresse: Architect (21/22 Namen; Fenster zwischen `docker create` und `trap`), User (Einspruch Schnitt, Kennung, ADR, CR), Planner (Hilfe `grep`), V-3 zu aktualisieren |
| Diff-Vorschlag für `AGENTS.md` §3.18 an den User | steht aus (Closure) | Plan §5 |

## 4. Plan gegen Code, Mengen und Proben

**Plan gegen Code.** Die 38 Zeilen der Tabelle haben im Code je einen Test oder Fall; alle 25 Fälle bestehen mit dem Original (Lauf `base`, 25 von 25; 13 weitere Läufe mit je einer Treiber-Mutation). Dateien des Blocks gegen §3: Programm, Treiber, Fragment, Stufe, `.dockerignore`, `.golangci.yml`, Gegenproben, Nachzüge vorhanden; `SPEC-101` unverändert wie zugesagt; Vertrag und Plan stimmen in den Zahlen (253, 63/170/1/1/16/2, 22/65, 169/10/2, 84/51 612, 254/508, 28 Läufe des Treibers: 1, 3, 10, 8, 6). Der Plan ist gelesen und gegen die Diffstat gehalten; eine Abweichung, die nicht in §6 steht, habe ich nicht gefunden (kein Anspruch auf Vollständigkeit der Stichprobe).

**Mengen (aus dem Wortlaut des Vertrags gebildet, nicht aus den Befunden; Elementzahl, Probe, Ergebnis):**

| Menge | Elemente | Probe (eigenes Orakel) | Ergebnis |
|---|---|---|---|
| Z (Nr. 7 a) | 253 (63 + 170 + 1 + 1 + 16 + 2) | alle 1 112 064 Skalarwerte gegen die Bereiche von Hand | 0 Abweichungen |
| Zeichen mal Stellung mal Dateigröße-Grenze | 252 (ohne U+0000) mal 4 Formen (Dateianfang; Tabulator und Leerzeichen davor; Zeichen mit 2, 3, 4 Bytes, U+FFFD und kombinierendes Zeichen davor; ohne Zeilenende am Dateiende) und je 1 bis 3 Lagen über die Byte-Grenzen 4096, 8000, 12 288; U+0000 hinter Index 8000 | 2596 Dateien, 3858 Befunde, volle Ausgabe samt Summenzeile gegen das Orakel | 0 Abweichungen |
| ungültiges UTF-8 und Z in Folgen | 1 222 980 Folgen aus 1 bis 4 Bytes über 33 Randwerte (C0, C1, C2, DF, E0, E1, EC, ED, EE, EF, F0, F1, F3, F4, F5, FF, 80 bis BF-Ränder, 00, 09, 0A, 0D, 41, 7F, 85, AD) | Orakel: Tabelle der gültigen UTF-8-Folgen, Befund je Byte, Position, Z-Befunde | 0 Abweichungen |
| Reihenfolge am Schrägstrich | 15 Namen (u. a. `a-`, `a.`, `a0`, `a_`, `a~`, `a/ b`, ein Name mit Byte FF und einer mit Umlaut) | feste Erwartung von Hand nach der Bytefolge | stimmt |
| Treiber, Prüfmenge (echtes `docker`, `git`, `tar`) | getrackt, ungetrackt, getrackt mit U+2028, Name mit Leerzeichen und Umlaut, ignorierte, im Arbeitsbaum gelöschte (Index kennt sie), Gitlink (Index 160000 mit Befund im Inhalt), verschachteltes Repo, `.harness/skills`, `.harness/baseline/` | Klon des Zweigs, 7 eingepflanzte Befunde | erwartete 7 gemeldet, die vier ausgeschlossenen Klassen (Baseline, ignoriert, Gitlink-Inhalt, verschachteltes Repo) nicht, gelöschte Datei nicht |
| Treiber, Randformen | leeres Argument, defekter Index, Aufruf aus Unterverzeichnis | echtes `docker` | `Aufruf: zeichen.sh` Exit 2; `fatal` von `git` und `zeichen.sh: git ls-files scheitert` Exit 2; Unterverzeichnis richtig (Fassung des Zweigs) |

**Eigene Mutationen (33 Kopien, genau eine Änderung je Kopie).**

| Gruppe | Mutationen | rot | grün |
|---|---|---|---|
| Programm, andere (5) | Zähler *d* zählt alle Dateien; `ok` auf stderr; BOM-Spalte 2; NUL-Regel bei Länge 1; Lesepuffer 16 | 4 | 1 (Puffer 16, äquivalent, wie der Reviewer) |
| Programm, V-2 (11) | Ausnahmen für 5 Pfade und 6 Dateinamen/Endungen | 0 | 11 |
| Programm, V-5 (3) | Wurzel per `Lstat`; Name `.` des Stamms leer; 4-MiB-Grenze | 0 | 3 |
| Treiber (14) | stdout/stderr/Aufruf, `-gt 1`, `--recurse-submodules`, Exit-Abbildung, `pwd` statt `rev-parse`, `start` mit `>/dev/null`/`2>&1`/`2>/dev/null`, doppelte Liste, `${ZEICHEN_IMAGE-}`, `rm` ohne Umleitung, `create` mit `2>/dev/null`, `cd` entfernt | 10 | 4 (`cd` = V-1; `${ZEICHEN_IMAGE-}`, `rm` ohne Umleitung = V-5; `create` mit `2>/dev/null` berührt keinen Vertragssatz) |

Rate: das Programm ist sequentiell (`SPEC-233` Nr. 12 d), jede Mutation lief in einem Prozess; die Treiber-Fälle sind deterministisch (der Ersatz sendet das Signal und endet). Die Raten des Implementers (3 von 3) sind nicht nachgemessen; sie stützen kein Urteil dieses Reports.

## 5. Sensoren, Messungen

| Sensor | Ergebnis |
|---|---|
| `make gates` | Exit 0 (aus Datei gelesen); zuletzt `zeichen: ok, 893 Datei(en) geprüft, 2 Binärdatei(en) und 5 andere Einträge übergangen` |
| Stufen ohne Cache (`--no-cache-filter`) | `test` (93,0 s im Schritt), `zeichen` (8,2 s), `lint` (29,0 s), `race` (118,2 s): je Exit 0; die Stufen `lint` und `race` ließen die Basis-Schritte `CACHED`, der geprüfte Schritt lief |
| `make groessenmarke` (einzeln) | Exit 0: `ok, Messung 2884 von Marke 4900`; Hinweis "Tabelle ergibt 3720, Schätzung 3560" |
| Messung nachgerechnet | `git diff --numstat 3cdbdca..HEAD` über die zehn Pfade des Blocks: 2862 Einfügungen und 22 Löschungen = 2884, gleich dem Gate; 59 Prozent der Marke 4900. Schätzung 3560, Tabelle 38 Zeilen (S = 80 mal 38 + 200 + 480 = 3720), Marke fest nach dem ersten Code-Commit, der Hinweis ist nach `SPEC-128` Nr. 5 c erlaubt. Messpunkt Produktcode 321 gegen 380 des Architect |
| `make zeichen-check` auf dem Repo | grün (893 Dateien, im Lauf von `make gates`) |
| `make zeichen-check` mit eingepflanzten Zeichen | rot: Klon, 7 Befunde (NBSP, U+2028, CR, BOM, ZWSP in einem Namen mit Leerzeichen und Umlaut, UTF-16-BOM, NBSP unter `.harness/skills`), `make` Exit 2 |
| `make zeichen-check` auf dem Inhalt von `main` | rot, 2 Befunde (V-6) |
| Zeitbudget, Fälle 131 bis 155 | 25 Fälle in 0,14 bis 0,24 s je Lauf (14 Läufe, davon der Lauf mit dem Original 0,24 s; `--cpus=4`, je zwei Läufe parallel, Load Average 9 bis 10 zu Beginn), 28 Läufe des Treibers darin; Stufe `zeichen` ohne Cache 8,2 s im Schritt. `make gegenproben` nicht gefahren (Auftrag): die Messung des Gesamtlaufs (116,7 und 110,9 s vorher, 127,1 und 106,6 s nachher, Streuung 20,5 s) ist nicht nachgemessen; sie ist in `SPEC-102` und im Plan §6 konsistent, und die Zusatzzeit der Fälle liegt um zwei Größenordnungen unter der Streuung |
| `make docs-check` und `make zeichen-check` auf diesem Report | beide grün, Abschnitt 8 |

## 6. Bewertung der Fragen des Auftrags

1. **Fixes zu F-1 bis F-8.** F-2 (Wurzeln ohne Textdatei, 63 Teilmengen): gedeckt, durch meine Proben nicht widerlegt. F-3 (Gitlink und verschachteltes Repo, Fall 151): bestätigt mit echtem `docker` (beide ohne Inhalt, `--recurse-submodules` ist rot). F-4 (`git ls-files` scheitert; leeres Argument): bestätigt mit echtem `docker` (Exit 2, Meldungen wie im Vertrag; `-gt 1` und `[ -z ]`-Gegenform rot). F-5 (Spalte, 84 Dateien und 51 612 Prüfungen): gedeckt, mein Orakel mit anderen Vorgängern stimmt. F-6 (Sortiergrenze am Schrägstrich): gedeckt, 15 eigene Namen stimmen. **F-1 (Ausnahmen) ist nicht klassengedeckt: V-2.** F-7 und F-8 (Plan folgt, Kopf der `zeichen.sh`): im Plan nachgezogen (V-3 für den neuen Stand).
2. **Das Gate als Sensor.** Auf dem Repo grün, mit eingepflanztem Zeichen rot (Treiber und `make`); die Verdrahtung `GATE_CHECKS` als akzeptiertes Negativ belegt (2 gegen 0 Treffer); die Kopie in die Stufe `lint` ist im `Dockerfile` und im Lauf der Stufe zu sehen. Das Negativ trägt Grund, Folge und Beleg (Plan Zeile 1, `SPEC-233` Nr. 16 h).
3. **Wirkung auf das Repo.** Zweig: kein Befund. `main`: 2 Befunde (V-6), im Zweig berichtigt.
4. **Zeitbudget.** Abschnitt 5.
5. **Messung.** Abschnitt 5; die Zahlen stimmen.
6. **Offene Zeilen in §6.** Architect: 21/22 Namen (Zahl berichtigen, sie trägt nichts) und das Fenster zwischen `docker create` und `trap` (V-8). User: Zeile 45 ist inzwischen gelöst (V-3, Plan nachziehen). Beide haben Ausgang *offen* und Adresse (§3.13); der Plan-Eintrag zu Zeile 45 ist veraltet.
7. **Bestandsfix und Betriebsregel.** Das Mittel `<U+00A0>` als Text statt eines Leerzeichens ist das bessere: die Zeile sagt wieder, was die Fixture trug (ein Wert mit zwei NBSP wird nicht gelesen), und folgt der Regel des Plans "in Prosa steht `U+00A0`". Ich habe nachgemessen: `sed 's/\xC2\xA0/<U+00A0>/g'` unter `LC_ALL=C` auf Zeile 45 der Fassung vor `4bbe19d` ergibt bytegleich die heutige Zeile 45 (`cmp`), die Datei enthält danach 0 Treffer für C2 A0, `git show --stat` zeigt eine Zeile. Der Rückgriff auf `sed` statt `Edit`/`Write` ist begründet (das Schreibwerkzeug setzt das Zeichen in ein Leerzeichen um, der Beleg steht in §6) und war wegen `LC_ALL=C` bytegenau; der Fehler der ersten Berichtigung (Leerzeichen statt Text) wurde mit dem zweiten Commit richtig behoben. Mängel nur dokumentarisch: die Ausnahme steht nicht in der Betriebsregel (V-4), der Plan folgt nicht (V-3). Eine Verletzung der Regel sehe ich nicht: das Ziel der Regel (kein Edit mit Zeichen in den Argumenten) ist erreicht.

## 7. Nicht geprüft

- `make gegenproben` als Ganzes und der Stempel (Auftrag): die Fälle 131 bis 155 einzeln, 25 von 25 grün, jeder Fall nur mit dem Original und den genannten Mutanten.
- Die Raten des Implementers (3 von 3 Prozesse, 44 Mutationen je Runde): nicht nachgefahren; ersetzt durch 33 eigene Mutationen.
- `tar` des Hosts: das Gate lief mit dem GNU tar dieses Hosts; die Abweichung eines anderen `tar` fängt keine Gegenprobe (Grenze in `SPEC-102`, benannt).
- Ein Signal an den Treiber während `start` mit echtem `docker`: vom Reviewer gemessen (Exit 2, Container entfernt), nicht wiederholt.
- Der Inhalt der Dateien von `.harness/baseline/` (Kurs-Inhalt, `make baseline-verify`).

## 8. Gegenlesen und Gates auf diesem Report

Dieser Report enthält kein Zeichen aus Z und keine Folge aus Backslash, `u` und vier Hex-Ziffern; beschrieben sind Zeichen als `U+XXXX`. Rohbytes: `env LC_ALL=C grep -aP '[\x00-\x08\x0B-\x1F\x7F]|\xC2[\x80-\xA0\xAD]|\xE2\x80[\x8B-\x8F\xA8-\xAF]|\xEF\xBB\xBF'` auf dieser Datei: 0 Treffer. `make docs-check` auf dem Zweig mit diesem Report: `d-check: 470 Datei(en) geprüft, 0 Befund(e)`; `make zeichen-check`: `zeichen: ok, 894 Datei(en) geprüft, 2 Binärdatei(en) und 5 andere Einträge übergangen` (die Datei dieses Reports liegt in der Prüfmenge).

## 9. Register-Kandidaten (nur genannt)

- `fix-deckt-die-genannte-mutation-nicht-die-klasse`: V-2 (F-1 wurde mit Namen als einzelne Komponente gefangen, Pfade und Endungen blieben offen; Treffer nach den 9 des Registers).
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft`: V-1 (Wechsel in die Wurzel).
- `betriebsregel-fuer-repo-dateien-verletzt`: V-4 (Regel und Ausnahme an zwei Orten).

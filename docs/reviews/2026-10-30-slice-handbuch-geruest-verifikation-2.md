# Verifikations-Report 2: slice-handbuch-geruest — 2026-10-30

**Rolle:** Verifier, Nachverifikation (zweite Runde) nach der Fix-Runde zu [V-1 bis V-6](2026-10-30-slice-handbuch-geruest-verifikation.md). Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-200` und `AGENTS.md` §3.9 bis §3.18. Kein Review, keine Reparatur: Code, Plan und Vertrag sind von mir unverändert; der Arbeitsbaum des Zweigs war vor und nach den Läufen ohne Änderung (`git status` leer), dieser Report ist die einzige Datei dieses Commits.

**Gegenstand:** Zweig `slice-handbuch-geruest`, HEAD `3169035`, Basis des Slice `975228d`. Fix-Runde: `275e5ad` (Architect, `SPEC-200` Nr. 2 d, 9 b, 14 d), `c1a5c82` (Tests `handbuch_runde4_test.go`, Plan), `59aa774` (Lint), `7873b4a` (Messung), `3169035` (Planner, Plan-Nachzug). `git diff 959c496..HEAD --stat` über das ganze Repo: drei Dateien (Plan, `SPEC-200`, `tools/pruefung/handbuch_runde4_test.go`, 483 Zeilen); **der Produktcode von `tools/pruefung`, das `Dockerfile` und die Makefiles sind seit der Verifikation unverändert.** Die Raten von Z26 und Z49 aus der Verifikation (Abschnitt 6 dort) gelten damit weiter und sind nicht neu gemessen.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Methode.** Keine Zahl und keine Mutation des Implementers, des Reviewers oder der Vorverifikation übernommen. Scratchpad `verif-geruest-2/` (flüchtig): `make gates` im Worktree (Exit-Code aus Datei), die Stufe `handbuch` ohne Cache einzeln, `make groessenmarke`; eine Wegwerfkopie von `tools/pruefung` je Mutation (literales `sed` auf der Kopie, nie im Repo), gefahren im Go-Image des `Dockerfile` (Digest gepinnt, `--network none --memory=4g --memory-swap=4g --cpus=4`, Quelle `:ro`, `/tmp` als `tmpfs`), alle Läufe unter `flock /tmp/odata-recorder-heavy.lock` in drei Gruppen (die Wartezeit auf die Sperre betrug rund 25 Minuten, andere Rollen hielten sie); ein Probe-Paket (`probe_test.go`, Saat ab 300000) mit eigenen Orakeln aus dem Wortlaut von `SPEC-200`, dessen Mengen (die 44 Muster und ihre 104 Formen, die Zeichenfolgen von Nr. 8) aus dem Text des Vertrags geparst sind und nicht aus den Tests. Alle eigenen Container sind beendet, kein eigener Sperrwartender läuft.

---

## 1. Verdikt

**NICHT BESTANDEN.** Es gibt kein HIGH. Die Fixes zu V-1 (a) bis (g) und V-2 (a) bis (c) tragen für die **genannten** Klassen: das Probe-Paket (13 Tests, rund 983 000 Elemente, davon 980 000 Kalendertage) ist auf dem unveränderten Programm grün, `make gates` ist grün (Exit 0), Stufe `handbuch` ohne Cache grün, die Marke hält (4347 von 6900), V-3 und V-5 und die Klammer von Z53 sind im Plan nachgezogen. Der Wortlaut von `SPEC-200` Nr. 2 d, 9 b und 14 d trägt die drei Randformen von V-2, ohne dass man den Code lesen muss.

Das Urteil steht dennoch aus einem Grund, den `AGENTS.md` §3.17 vorgibt: Der Fix zählt erst, wenn er die **Klasse** deckt, nicht die genannten Mutationen. Von 34 eigenen Mutationen an Elementen, die weder Reviewer noch Implementer noch Vorverifikation nannten, bleiben **21 grün**, eine davon als benanntes Negativ (Z49, 2 s), **20 als Bruch einer Zusage des Vertrags**. Vier Befunde sind MEDIUM (W-1 bis W-4), einer LOW (W-5, Elemente von geringem Gewicht), einer LOW für den Planner (W-6). Ein Befund (W-2) betrifft den Architect: zwei Randformen stehen weiter nur im Code. Eine Re-Verifikation prüft nur W-1 bis W-5.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| W-1 | MEDIUM | **Vergleich der Software-Version und leere Fehlerausgabe ohne Fänger der Abweichungsarten** (`SPEC-200` Nr. 4 c: „der Wert im Kopf muss gleich *V* sein“, „leere Fehlerausgabe“; Z10). Z10 führt als Abweichung nur einen ganz anderen Wert (`1.0` gegen `2.0`) und als Fehlerausgabe nur `x\n`. Drei Mutationen lassen die ganze Suite grün: n04 Vergleich ohne Beachtung von Groß- und Kleinschreibung (`Dev` im Kopf, `dev` in der Ausgabe gälte als gleich), n05 Vergleich als Präfix (`1.0` im Kopf, `1.0.1` in der Ausgabe gälte als gleich), n06 eine Fehlerausgabe, die nur aus Leerraum besteht (`\n`), gälte als leer. Die Menge, die die Zusage meint: jede Art, in der zwei Werte nicht gleich sind (Schreibung, Präfix, Suffix, ein Zeichen, Leerraum am Rand) und jede nicht leere Fehlerausgabe. Das ist die Kernzusage von K2 (`SPEC-200` Nr. 1). | Abschnitt 5, Mutationen n04 bis n06 grün; `versionBefunde`, `versionAusgabe` in `handbuch_kopf.go` | Implementer |
| W-2 | MEDIUM | **Zwei Randformen ohne Ort in `SPEC-200`, im Code entschieden** (`AGENTS.md` §3.16, §3.12). (a) *Gleichstand in Nr. 4, Zeile 0*: Nr. 2 d ordnet bei gleicher Nummer, Zeile und gleichem Namen „in Nr. 4 die Schlüssel in der Reihenfolge `Software-Version`, `Gültigkeitsbereich`, `Stand`“. Seit der Fix-Runde hat `version: Ausgabe nicht lesbar` bei fehlendem, leerem oder doppeltem Schlüssel `Software-Version` ausdrücklich die Zeile 0 und steht damit gleichauf mit `Kopf: <Schlüssel> fehlt` (ebenfalls Zeile 0, ebenfalls ohne Namen). Der Vertrag sagt nicht, welcher zuerst kommt; der Code setzt `fehlt` vor `nicht lesbar` (Probe: `Kopf: Software-Version fehlt`, `Kopf: Gültigkeitsbereich fehlt`, `version: Ausgabe nicht lesbar: Exit 1`). Kein Fänger: die Hilfe `erwarteReihenfolge` in `handbuch_runde4_test.go` sortiert Befunde gleicher Zeile vor dem Vergleich um, Z52 („fehlt, nicht lesbar“) kann den Gleichstand nicht sehen; die Mutation n08 (`nicht lesbar` vor `fehlt`) lässt die Suite grün. (b) *Zeile `<n>` der Befunde von Nr. 9 b* (`make-Ziel <x> unbekannt`, `make-Aufruf ohne Ziel an erster Stelle`): Nr. 9 b nennt *n* nicht; Nr. 13 d definiert *n* nur für die Mengen (erste Zeile des Absatzes). Der Code nimmt die Zeile, in der der Span **öffnet** (Probe: Absatz in den Zeilen 8 und 9, Span in Zeile 9, Befund „Zeile 9“); das fängt `TestHandbuchMakeZiele` (die Mutation n33, Zeile des Absatzes statt der des Spans, ist rot). Es fehlt also nur der Ort. | Abschnitt 5 (n08, n33); Probe `TestPnGleichstandFehltNichtLesbar`, `TestPnSpanZeile`; `SPEC-200` Nr. 2 d, 9 b, 13 d | Architect (Ort in `SPEC-200`), danach Implementer (Fänger für (a), Zeile in Plan §6) |
| W-3 | MEDIUM | **Z51: die Zahl der Vorkommen bei `Kopf: <Schlüssel> doppelt` fehlt in der Menge.** Der Vertrag (Nr. 2 d) sagt „die Zeile des ersten Vorkommens“; die Menge ist jede Zahl von Vorkommen ab zwei. Z51 führt genau zwei. Die Mutation n07 (die Zeile wird bei jedem zweiten Vorkommen gesetzt, bei drei Vorkommen also das dritte) lässt die Suite grün. Das Probe-Paket (Saat 300000 bis 300399, Schlüssel einmal, leer, doppelt, dreifach, fehlt, in zufälliger Reihenfolge und mit Leerzeilen und Fremdzeilen dazwischen) ist auf dem unveränderten Programm grün, der Code ist also richtig, der Fänger blind. Klasse von V-2 (a) und Mutation an einem Element, das der Befund nicht nannte (§3.17). | Abschnitt 5, n07 grün; `TestPnKopfZeilen` grün | Implementer |
| W-4 | MEDIUM | **Elemente der Mengen von Nr. 5, 8, 9 und 13 b, die keine Schleife der Fix-Runde führt** (je Zusage des Vertrags, `AGENTS.md` §3.15; das Probe-Paket bestätigt jeweils, dass der Code den Wortlaut einhält). Grün bleiben: n12 Info-Zeichenkette nur am **ersten** Zaun geprüft (Nr. 5 b „jeder Zaun“; Z14 führt einen Zaun); n13 und n14 `Datei: [` an beliebiger Stelle der Zeile statt am Anfang (Nr. 5 b, 5 c „beginnt mit“; eingerückt oder mit Text davor); n17 Ziffer **hinter** dem Treffer verhindert den Treffer (Nr. 13 b „kein Buchstabe“; n16, die Ziffer davor, ist rot, die Seite dahinter nicht); n18 Zuweisung nur bei einem einzelnen Buchstaben als Namen (`make VERSION=1 build` gälte als Ziel `VERSION=1`; Z29 führt `X=1`); n19 Option nur `-j` (`make -n`, `make --help`; Z29 führt `-j`); n23 `d-check.mk` und `repo.mk` zählen als Quelle des Images (Nr. 8 nennt nur `Makefile` und `harness/mk/*.mk`; Z28 führt die erlaubten Quellen, keine verbotene); n24 der Pfad `-o /out/odata-recorder` im Makefile statt im `Dockerfile`. Die Menge ist jeweils die Zahl der Stellen (Zäune), der Arten von Zeichen hinter dem Treffer, der Optionen und Zuweisungen, der Dateien. | Abschnitt 5, Mutationen n12 bis n14, n17 bis n19, n23, n24 grün | Implementer |
| W-5 | LOW | **Elemente von geringem Gewicht, ebenfalls ohne Fänger.** n03 die Jahrhundertregel mit Teiler 1000 statt 400 (2000 gilt weiter; 1600, 2400, 2800 würden falsch; Z11 führt 2000 und die Jahre 1900, 2100); n21 Unterstrich im Namen einer Zielzeile (Nr. 9 a `[a-z0-9-]+`); n22 Datei mit `.mk` mitten im Namen (`z.mk.bak`; Nr. 9 a „Endung `.mk`“); n27 Tabulator statt Leerzeichen hinter dem Präfix der Kopfzeile des Beispiels (Nr. 6 a „genau ein Leerzeichen“); n31 Tabulator vor `## ` beendet den Kopf (Nr. 4 a „beginnt mit `## `“); n32 Schlüssel, die mit `#` beginnen, ergeben `Zeile ohne Schlüssel` statt `unbekannt` (Nr. 4 a, eine Auswahl von vier fremden Schlüsseln in Z8); n35 Tabulator an beiden Enden eines Spans wird wie ein Leerzeichen entfernt (Nr. 9 b „ein Leerzeichen (U+0020)“). Der Planner kann sie mit Grund und Folge als akzeptierte Negative führen oder je ein Element in die bestehende Schleife nehmen. | Abschnitt 5 | Implementer oder Planner (akzeptiertes Negativ) |
| W-6 | LOW | **Plan, `AGENTS.md` §3.9: Zahlen von gestern als Stand von heute.** §4 „Marke fest“ nennt „Die Tabelle ergibt jetzt 5890 (56 Zeilen) … die Messung 3586 liegt bei 6900 bei 52 %“, die „Entscheidung zur Schnittfrage“ „die Tabelle ergibt jetzt 5890“ und Nr. 1 „für die Tabelle von heute, 5890, wären es 4870“ (Zeilen 397, 401, 403). Die Tabelle hat 59 Zeilen und ergibt 6130, die Messung 4347 (63 %), wie der Absatz darunter (Zeile 399, 412) richtig sagt. Dazu stehen vier Absätze „Messung des Implementers nach …“ (Zeilen 408 bis 412) nebeneinander; jeder ist datiert, aber der Plan trägt so die Chronik der Messung statt des Stands (`AGENTS.md` §3.7). V-3 und V-5 selbst sind erledigt (Abschnitt 6). | Plan §4 | Planner |

Einordnung der Schwere. W-1, W-3 und W-4 sind Zusagen ohne Fänger (`AGENTS.md` §3.15: MEDIUM); in W-4 tragen n12 bis n14, n17 bis n19, n23 und n24 Gewicht, weil sie Elemente einer Menge sind, die der Vertrag ausdrücklich nennt („jeder Zaun“, „beginnt mit“, „kein Buchstabe“, „beginnt mit `-` oder enthält `=`“, „Makefile oder harness/mk“). W-2 ist eine im Code entschiedene Zelle (§3.16: MEDIUM), klein: je ein Satz in Nr. 2 d und 9 b.

## 3. Auflösung der Befunde der Vorverifikation gegen die Klasse (§3.17)

| Befund | Menge (aus dem Wortlaut von `SPEC-200`) | Elemente | Fänger der Fix-Runde (Schleife, Orakel von Hand) | eigene Prüfung | Stand |
|---|---|---|---|---|---|
| V-1 (a) Form des Stands, Nr. 4 d | jede der 10 Stellen von `JJJJ-MM-TT` durch ein Zeichen außerhalb der Form; Länge; Kalender | 35 + Gutfall | `TestHandbuchStandStellen` (35), Schleife über Stellen × Zeichen (`x`, `/`, `:`, `0`, `.`) | Probe: 980 000 Kalendertage (Jahre 0 bis 9999 × Monate 0 bis 13 × Tage {0, 1, 28, 29, 30, 31, 32}, Orakel `time.Date`) und 352 Zufalls- und Stellen-Zeichenketten (Saat 300001), grün; Mutationen n01 (Monat 0), n02 (30-Tage-Monat 11) rot; n03 (Teiler 400) grün | gelöst für die Stellen; Rest W-5 |
| V-1 (b) Reihenfolge der Gründe, Nr. 4 c | alle gleichzeitig möglichen Gründe | 11 | `TestHandbuchVersionGruende` | Probe: Signal, Exit 0/1/7 × Fehlerausgabe leer/`x\n` × 6 Formen der Standardausgabe = 48, Orakel „erster Grund der festen Reihenfolge“, grün | gelöst für die Reihenfolge; die Abweichungsarten in W-1 |
| V-1 (c) Überschrift, Nr. 4 a | erste nicht leere Zeile des Kopfes beginnt mit `# ` | 10 | `TestHandbuchUeberschriftPraefix` | Probe: 17 Präfixe × 3 Resttexte (51) und 13 Kopfzeilen für das Kopfende, grün; n31 (Tabulator vor `## `) grün | gelöst; Rest W-5 |
| V-1 (d) Zeichenfolgen, Nr. 8 | jede Stelle von `-t odata-recorder:build` (23) und `-o /out/odata-recorder` (22) | 71 | `TestHandbuchAufrufpraefixStellen` | Probe: beide Zeichenfolgen **aus dem Text von Nr. 8 geparst**, je Stelle ersetzt, gelöscht, groß geschrieben, Leerzeichen verdoppelt, in Makefile, mk-Datei und `Dockerfile`: 198 Elemente, grün; n23 und n24 grün | gelöst für die Stellen; Quellen in W-4 |
| V-1 (e) UTF-8 der Beispiel-Datei, Nr. 6 | ungültige Bytefolgen × Stelle | 10 | `TestHandbuchBeispielUTF8` | Probe: 30 ungültige Folgen (10 von Hand, 20 aus Saat 300002) × 4 Stellen (Standardausgabe, Fehlerausgabe, Mitte, Ausgangswert) und 5 gültige mehrbyteige Folgen × 2 Abschnitte: 130, grün; n25 (nur die erste Hälfte der Datei) rot | gelöst |
| V-1 (f) je Zeile und Zeichen einmal, Nr. 3 c | Zeichen × Zeilen | 27 | `TestHandbuchZeichenJeZeileEinmal` | Probe: alle Zeichen Cc (U+0001 bis U+001F ohne Tabulator und Zeilenumbruch, U+007F bis U+009F: 62), 45 Zeichen aus Cf, Co, Mn, Me, 9 erlaubte (Zs, Tabulator) und 150 Zufallsmengen von Zeilen (Saat 300003): 266, grün; n09 (Leerraum der Cc-Zeichen ausgenommen) und n10 (ein Befund je Zeile) rot | gelöst |
| V-1 (g) Kopfende, Nr. 4 a | Zeilen im Kopf, die nicht mit `## ` beginnen | 9 | `TestHandbuchKopfEndeZeilen` | siehe (c) | gelöst; Rest W-5 |
| V-2 (a) Zeile von `Kopf: … doppelt`, Nr. 2 d | Befundarten × Vorkommen × Stellung der Fremdzeile | 24 | `TestHandbuchKopfBefundZeile` | Probe: 400 Saaten (300000 bis 300399), Schlüssel in jedem Zustand, zufällige Reihenfolge, Leerzeilen, bis zu drei Fremdzeilen, drei Ausgaben von `version` (gleich, abweichend, Exit 1), Orakel nur aus dem Wortlaut von Nr. 2 d, grün | Fänger deckt zwei Vorkommen: **W-3**; Gleichstand: **W-2 (a)** |
| V-2 (b) Zeile von `weicht … ab` und `nicht lesbar`, Nr. 2 d | fünf Befunde × Fremdzeile davor und dahinter | 10 | `TestHandbuchVersionBefundZeile` | im selben Probe-Lauf grün | gelöst; Gleichstand W-2 (a) |
| V-2 (c) Leerzeichen am Span, Nr. 9 b, 14 d | Anfang × Ende × Inhalt | 64 | `TestHandbuchSpanLeerzeichenAnDenEnden` | Probe: 500 zufällige Absätze aus Backtick-Folgen, Leerzeichen, Zeilenumbrüchen, `make`, `nix`, Optionen (Saat 300004) gegen einen **eigenen** Span-Parser nach dem Wortlaut (Folge aus *n* Backticks, nächste Folge von genau *n*, Zeilenumbrüche als Leerzeichen, Paar entfernen): 0 Abweichungen; n35 (Tabulator als Rand) grün | gelöst; Rest W-5 |

## 4. Wortlaut von `SPEC-200` Nr. 2 d, 9 b, 14 d ohne Rückgriff auf den Code gelesen (V-2, §3.16)

- **Nr. 2 d, Zeile eines Befundes mit Schlüssel.** Die drei Fälle des Befundes stehen ausdrücklich: `doppelt` = Zeile des ersten Vorkommens; `leer`, `Stand ungültig`, `weicht … ab` = Zeile des einzigen; `version: Ausgabe nicht lesbar` = Zeile des Schlüssels `Software-Version`, wenn er genau einmal und nicht leer steht, sonst 0; `version: Zeitgrenze überschritten` Zeile 0 (Nr. 7). Das Orakel der Probe ist allein aus diesem Wortlaut gebaut und stimmt mit dem Programm in 400 von 400 Saaten überein. Offen bleibt der Gleichstand (W-2 a). Zusatz: Befunde mit `Zeile <n>:` im Text (`Zeile ohne Schlüssel`, `unbekannt`, `Überschrift fehlt`) tragen ihr *n* selbst und brauchen die Regel nicht.
- **Nr. 9 b, Leerzeichen am Span.** Der Wortlaut ist eindeutig (Paar, nicht nur aus Leerzeichen, „und nur dann“, zwei Beispiele, `  make nix  `). Offen bleibt die Zeile *n* der Befunde (W-2 b).
- **Nr. 14 d, akzeptiertes Negativ.** ` make nix` bleibt ungeprüft; Grund und Folge stehen (Zusage nach §3.15 erfüllt). Der Grund („der Text des Spans folgt CommonMark; ein Prüfen ohne diese Regel wäre eine eigene Auslegung“) ist knapp, aber ein Grund; der Satz „Folge: ein unbekanntes Ziel in einem solchen Span steht ungeprüft im Text“ steht in Nr. 14 d zweimal hintereinander (nur Form).

## 5. Eigene Mutationen und Probe-Paket

Messender: der Verifier. Je Mutation eine Kopie von `tools/pruefung` im Container, `go test -count=1` über das ganze Paket (23 s), „rot“ heißt: ein Test scheitert, und die Fehlerzeile gehört zur Zusage (bei n26 war es ein Bauabbruch, `unicode` nicht mehr benutzt: ausgeschieden). Alle Mutationen sind deterministisch, es gibt keine Rate (Zahl der Prozesse je Mutation: 1, rot in 1 von 1 bzw. grün in 1 von 1).

| Rot (13) | Test | Grün (21) |
|---|---|---|
| n01 Monat 0 gälte | `TestHandbuchStandRegeln` | n03 Jahrhundertregel Teiler 1000 (W-5) |
| n02 Monat 11 mit 31 Tagen | `TestHandbuchStand` | n04 Version ohne Beachtung der Schreibung (W-1) |
| n09 Cc-Leerraum ausgenommen | `TestHandbuchCRLF` | n05 Version als Präfix (W-1) |
| n10 ein Befund je Zeile | `TestHandbuchZeichenJeZeileEinmal`, `…ReihenfolgeGleichstaende` | n06 Fehlerausgabe aus Leerraum (W-1) |
| n11 Schluss-Zaun als Präfix | `TestHandbuchZaunOhneSchlussRand` | n07 Zeile bei drei Vorkommen (W-3) |
| n15 Zeile des Zitats +1 | `TestHandbuchMengeCImZitat` | n08 Gleichstand `nicht lesbar` vor `fehlt` (W-2 a) |
| n16 Ziffer davor verhindert Treffer | `TestHandbuchTrefferGrenze` | n12 Info nur am ersten Zaun (W-4) |
| n20 Zerlegung des Spans an jedem Leerzeichen | `TestHandbuchMakeSpanOhneZweitesWort`, `…MakeSpanRand`, `…SpanLeerzeichenAnDenEnden` | n13 `Datei: [` mit Text davor (Bindung davor) (W-4) |
| n25 `\r` nur in der ersten Hälfte der Beispiel-Datei geprüft | `TestHandbuchBeispielAbschnitte` | n14 `Datei: [` mit Text davor (Bindungszeile ohne Zaun) (W-4) |
| n29 führende Null erst ab drei Stellen | `TestHandbuchBeispielAusgangswert` | n17 Ziffer dahinter verhindert Treffer (W-4) |
| n30 Reihenfolge der Trennzeilen nur teilweise | `TestHandbuchBeispielAbschnitte` | n18 Zuweisung nur bei Name der Länge 1 (W-4) |
| n33 Zeile des Absatzes statt der des Spans | `TestHandbuchMakeZiele` | n19 Option nur `-j` (W-4) |
| n34 Zerlegung des Spans nach Abschneiden der Leerzeichen am Ende | `TestHandbuchMakeSpanOhneZweitesWort` | n21 Unterstrich im Zielnamen (W-5) |
| | | n22 `.mk` als Teil des Dateinamens (W-5) |
| | | n23 `d-check.mk` und `repo.mk` als Quelle des Images (W-4) |
| | | n24 Pfad auch im Makefile gedeckt (W-4) |
| | | n27 Tabulator hinter dem Präfix der Kopfzeile (W-5) |
| | | n28 `WaitDelay` 2 s statt 1 s: **benanntes Negativ** (Z49, Plan: 1 s bis 2,5 s bleibt grün, Grund und Folge stehen) |
| | | n31 Tabulator vor `## ` (W-5) |
| | | n32 Schlüssel mit `#` (W-5) |
| | | n35 Tabulator als Rand des Spans (W-5) |

n20 und n34 sind rot, aber nicht aus dem Grund der Zusage „zweites Wort“ bei mehreren Leerzeichen zwischen `make` und Ziel (`make  nix`): sie scheitern an den Fällen ohne zweites Wort. Die Teil-Zusage ist ohne isolierende Mutation gefahren und bleibt ungeprüft (INFO).

**Probe-Paket** (`probe_test.go`, im Scratchpad, 13 Tests, auf dem unveränderten Programm alle grün; zwei Orakelfehler von mir, die Zeile `9` statt `8` im Test der Bindungszeile und die Kennung `A1` mit Doppelpunkt statt Leerzeichen im Test der Mengen, sind vor dem gezählten Lauf behoben):

| Test | Klasse | Generator | Orakel | Elemente |
|---|---|---|---|---|
| `TestPnKopfZeilen` | Nr. 2 d, Zeile der Kopf-Befunde | 400 Saaten ab 300000 | Wortlaut von Nr. 2 d, Gleichstände als Menge | 400 |
| `TestPnStandKalender` | Nr. 4 d | Jahre 0000 bis 9999 × Monate 0 bis 13 × 7 Tage | `time.Date` | 980 000 |
| `TestPnStandStellen` | Nr. 4 d | je Stelle 14 Zeichen, Löschen, Einfügen, 200 Zufallsfälle (Saat 300001) | Zeichen für Zeichen | 352 |
| `TestPnVersionGruende` | Nr. 4 c | 4 Endearten × 2 Fehlerausgaben × 6 Ausgaben | erster Grund der Reihenfolge | 48 |
| `TestPnUeberschriftUndKopfende` | Nr. 4 a | 17 Präfixe × 3 Texte, 13 Kopfzeilen | Präfix `# ` bzw. `## ` | 64 |
| `TestPnAufrufpraefix` | Nr. 8 | Zeichenfolgen aus dem Vertragstext, je Stelle 3 Veränderungen, 4 Orte | „enthält“ | 198 |
| `TestPnBeispielUTF8` | Nr. 6 | 30 ungültige Folgen × 4 Stellen, 5 gültige × 2 | `utf8.Valid` | 130 |
| `TestPnZeichenZeilen` | Nr. 3 c | 62 Cc, 45 Zeichen aus Cf/Co/Mn/Me, 9 erlaubte, 150 Zeilenmengen (Saat 300003) | Kategorien aus Nr. 3 c, Zeile × Zeichen einmal | 266 |
| `TestPnZaunSchluss` | Nr. 5 a bis c | 14 Schlusszeilen, 1 bis 4 Zäune × Zaun mit Info, 6 Bindungspräfixe | „rechts getrimmt gleich drei Backticks“ | 30 |
| `TestPnSpanFuzz` | Nr. 9 b | 500 zufällige Absätze (Saat 300004) | eigener Span-Parser | 500 |
| `TestPnMengenAusSpec` | Nr. 10 bis 12, 13 b | 44 Muster und 104 Formen aus dem Vertragstext × 7 Umgebungen (auch Ziffer davor und dahinter) und Buchstabe davor | Muster gleich Vertrag, Zählung 16/36, 14/44, 14/24 | 796 |
| `TestPnGleichstandFehltNichtLesbar`, `TestPnSpanZeile` | beobachtet, zu W-2 | – | – | 4 Läufe |

## 6. Plan und Tabelle gegen den Vertrag

- **V-3** (Plan: Start vor dem Werkzeug-Slice, `make gates` im Zweig rot, Reihenfolge der Commits) **nachgezogen**: §4 Zeile 368 und 372, §3 Zeile 141, §6 Zeile 444 (Ausgang *entfallen*) und Zeile 456. **V-5** (Rate von Z26 mit Zahl, Prozessen, Messendem) **nachgezogen**: Zeile 261. **Klammer von Z53** (Zeile 199) stimmt mit der Tabelle des Tests überein (Anfang ein Leerzeichen oder Zeilenumbruch: Ende „kein“ nicht geprüft, Ende 1, 2 oder Zeilenumbruch geprüft; Anfang zwei: nie geprüft).
- **Weitere veraltete Stellen** (§3.9): W-6. Alle übrigen Zahlen (59 Zeilen, 6130, 4347, Z1 bis Z53, 483 Zeilen der Fix-Runde) stimmen mit `make groessenmarke` und `git diff --numstat` überein.
- **Tabelle *Zusage · Fänger***: Zählprüfungen von Z6 (27), Z8 (10), Z9 (9), Z10 (11), Z11 (35), Z21 (10), Z28 (71), Z51 (24), Z52 (10), Z53 (64) stehen im Plan und im Test mit denselben Zahlen; alle Namen von Tests im Plan existieren im Code (0 fehlende). Die Mengen der Fänger folgen dem Wortlaut des Vertrags für die genannten Klassen; Lücken innerhalb derselben Nummern sind W-1 bis W-5.
- **Akzeptierte Negative**: (i) ` make nix`, N4, Z49 (1 s bis 2,5 s), R-9 (Symbolischer Verweis), (h) der Anker `^` nennen Grund und Folge. Keines wurde seit der Vorverifikation gelockert.
- **§3.6**: kein Gate gelockert; `make handbuch-check` bleibt in `GATE_CHECKS`.
- **Rohbytes (§3.18)**: `handbuch_runde4_test.go`, `SPEC-200` und der Plan enthalten keine Folge Backslash-`u`-Hex, kein BOM, kein NBSP, kein `\r`, sind gültiges UTF-8 (`LC_ALL=C grep -P`, `iconv`); die einzigen Zeichen jenseits von ASCII sind Umlaute, `ß` und `×`.

## 7. Läufe

- **`make gates`** im Worktree: `EXIT=0` (Datei, nicht über `| tail` maskiert); darin `d-check: 385 Datei(en) geprüft, 0 Befund(e)`, `a-check` 0 Befunde, `groessenmarke` ok.
- **Stufe `handbuch` ohne Cache** (`docker build --no-cache-filter handbuch --target handbuch`): `RUN pruefung handbuch /out/odata-recorder /handbuch` lief neu, `pruefung handbuch: ok` (0,169 s), `EXIT=0`.
- **`make groessenmarke`**: `Hinweis: Tabelle ergibt 6130, Schätzung 5085`, `ok, Messung 4347 von Marke 6900` (63 %), Messpunkt `zeilen=59 nebenlaeufig=2 fakes=2 produktcode=1000 schaetzung=6130 marke=6900 messung=4347`; `EXIT=0`. Die Marke ist unverändert 6900, ein Eintrag `ueberschreitung` ist nicht nötig; die Fixes zu W-1 bis W-5 zählen zur Messung.
- **Nicht gefahren, wie beauftragt:** `make gegenproben` und der Stempel (Integrationsschritt).

## 8. DoD: was bleibt offen (nur genannt)

- **Closure-Notiz mit Lerneintrag** in §7 des Plans (*Wird bei der Closure gefüllt*); alle Häkchen in §2 sind noch leer.
- **Beobachtungs-Register** fortschreiben (Sache des Planners; Kandidaten unten) und die drei Paarungen (Anker, Folge-Slice, Register).
- **Diff-Vorschlag für `AGENTS.md`** (Hard Rule „Ist-Zustand“): liegt als Scratch-Datei beim User; §6 führt den Punkt mit Ausgang *offen*, Adresse User.
- **Integrationsschritt:** `make gates` im Hauptbaum nach dem Merge (Plan §5); `make gegenproben` bei geringer Last und der Stempel allein committet; endgültige Fallnummern 121 bis 130 in `SPEC-102`; Welle, Roadmap-Zeile, `docs/plan/planning/README.md` Nr. 6; Nachzug bei den Nehmern (`slice-handbuch-meldungskatalog`, `-abnahme`, `-record-replay`).
- **Dazu aus dieser Runde:** W-1 bis W-5 (Fänger), W-2 (Ort im Vertrag und Zeile in Plan §6 mit Ausgang *offen* und Adresse, `AGENTS.md` §3.13), W-6 (Plan). Danach `git mv` nach `done/` als eigener Commit.

## 9. Ungelaufen

- `make gegenproben`, Stempel (Auftrag); `make gates` im Hauptbaum.
- Z26 und Z49: Raten nicht neu gemessen, weil Produktcode, `Dockerfile` und Makefiles seit der Vorverifikation byte-gleich sind (`git diff 959c496..HEAD`); die Zahlen der Vorverifikation gelten.
- Zeile für Zeile jede Mutation des Implementers (Z1 bis Z50) nicht nachgefahren; gedeckt durch die Reviews, die Vorverifikation und die Mutationen und Proben dieser Runde, soweit sie die Zeilen berühren.
- Die Teil-Zusage „zweites Wort“ bei mehreren Leerzeichen zwischen `make` und Ziel ist ohne isolierende Mutation (n20, n34 scheitern an anderen Fällen).
- Die Verdrahtung gegen ein falsches `Dockerfile` (N5; der Fall „Wort aus A“ ist in der Vorverifikation gefahren, der Code ist seitdem unverändert).

## 10. Register-Kandidaten (nur genannt, nicht geschrieben; Abgleich und Zähler sind Sache des Planners)

- `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`: vierte Runde an demselben Leser; die Fixes führen die genannten Klassen als Schleife (13 von 34 neuen Mutationen rot), die Nachbarn derselben Nummern (Vergleichsarten, Zahl der Vorkommen, Orte von Quellen, Seiten einer Grenze) nicht.
- `BEO-REPO/randform-erst-im-review-entschieden` (W-2): wieder eine Zelle der Ordnung (Gleichstand) und eine der Zeile, beide durch die Fix-Runde selbst erzeugt.
- Neu: `BEO-REPO/hilfe-des-tests-sortiert-den-gleichstand-weg` (W-2 a): `erwarteReihenfolge` ordnet Befunde gleicher Zeile vor dem Vergleich; ein Fänger für die Reihenfolge von Gleichständen kann sie nicht sehen.
- `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (W-1, W-4): Vergleichsarten, Quellen, Optionen.
- `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (W-6): Zahlen mit „jetzt“ und „heute“ in §4.

## 11. Zusammenfassung

| Schwere | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 (W-1, W-2, W-3, W-4) |
| LOW | 2 (W-5, W-6) |
| INFO | V-4 und V-6 der Vorverifikation unverändert (Prosa des Handbuchs, bekannte Posten des Integrationsschritts); n20/n34 ohne isolierende Mutation; n28 benanntes Negativ |

**Übergabe an den Planner:** NICHT BESTANDEN. W-1, W-3, W-4 an den Implementer (je Element eine Zeile in der bestehenden Schleife: Abweichungsarten von *V* und Fehlerausgabe aus Leerraum; Vorkommen 2 und 3 und mehr; Zäune 1 bis *k*; Präfix und Einrückung von `Datei: [`; Ziffer hinter dem Treffer; Optionen und Zuweisungen; Quellen des Images und des Pfades); W-2 an den Architect (je ein Satz in Nr. 2 d und 9 b), danach der Implementer; W-5 an Implementer oder Planner (Element oder akzeptiertes Negativ mit Grund und Folge); W-6 an den Planner. Alles übrige hält: `make gates` Exit 0, Stufe `handbuch` ohne Cache grün, Marke 6900 fest bei Messung 4347, Produktcode seit der Vorverifikation unverändert, 13 Probe-Tests mit rund 983 000 Elementen grün. Die Re-Verifikation prüft W-1 bis W-5 gegen die Klasse, nicht gegen die hier genannten Mutationen.

# Verifikations-Report 5: slice-odata-url-pfad — 2026-10-31

**Rolle:** Verifier, frischer Kontext, fünfte Runde, gezielt auf das Delta `8d97f9d..4db72ef` (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, [SPEC-160](../../spec/spezifikation.md) bis [SPEC-163](../../spec/spezifikation.md), [SPEC-125](../../spec/spezifikation.md) Nr. 8 b, [SPEC-230](../../spec/spezifikation.md) Nr. 8, [SPEC-109](../../spec/spezifikation.md) Nr. 2, 5, 8 und [AGENTS.md](../../AGENTS.md) §3.10, §3.12, §3.15 bis §3.18). Kein Review, keine Validierung, keine Reparatur.

**Gegenstand:** Zweig `slice-odata-url-pfad`, geprüfter Stand `4db72ef` (Fix-Runde `feacb68` nach [Report 4](2026-10-30-slice-odata-url-pfad-verifikation-4.md), Merge von `main` `95eefe3`, Nachzug der Zählungen `a5aaa31`, `4db72ef`), Arbeitsbaum zu Beginn sauber. Plan: [slice-odata-url-pfad.md](../plan/planning/in-progress/slice-odata-url-pfad.md). Basis der Marke `7535bf7`, Marke 9200.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31

**Methode.** Keine Zahl, Mutation oder Äquivalenzbehauptung des Implementers ist übernommen. Scratchpad `/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-urlpfad-5/` (flüchtig; im Repo liegt nur diese Datei). Das Probe-Paket hat ein eigenes Orakel (Ausdrücke und Schleifen über `unicode.Is(L)`, `unicode.Is(Nd)`, ASCII-Regeln, nie der Parser). Jede Mutation ist eine Änderung an einer Kopie, ein Lauf des Pakets `urlparser` ohne die Probe; blieb die Suite grün, folgte die Probe, um „gleichwertig“ von „ungefangen“ zu trennen. Schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`, je Gruppe eine Sperre, `docker run --network none --memory=4g --memory-swap=4g --cpus=4`. Ein Fehler meiner Probe (Orakel der Positionen `select-mitte`, `orderby-erster`, ein Duplikat in `key3`, Pfadlänge über 2048) ist vor den Zahlen unten behoben; die Zahlen stammen vom behobenen Stand, die Probe ist am unveränderten Code grün.

**Der Produktcode des Slice ist unverändert:** `git diff --stat 8d97f9d..4db72ef -- src/adapters/driven/odata ':!*_test.go'` ist leer, ebenso `git diff --stat 95eefe3..4db72ef -- src ':!*_test.go'`. Die 13 Produktdateien, die `git diff --stat 8d97f9d..4db72ef -- src ':!*_test.go'` zeigt (Telemetrie, Workspace-Dienst, `ports/support`; 1215 Einfügungen), kommen aus dem Merge `95eefe3`, nicht aus dem Slice. Eigene Änderung seit dem Merge: 10 Dateien (Plan, `slice-telemetry-spans.md`, Spezifikation, 2 Tests im Paket `urlparser`, 5 Tests fremder Pakete).

---

## 1. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| <a id="v5-1"></a>V5-1 | **MEDIUM** | **Die Zusage „Name“ ist an den Orten gedeckt, nicht in ihrer Stellung und Form innerhalb des Orts ([AGENTS.md](../../AGENTS.md) §3.17).** `TestNamenAchsen` und `TestUnicodeKategorien` prüfen an jedem der elf Orte genau eine Form: der Name steht allein (`$select=m`, `$orderby=m`, `S(m=1)`, `/S/p.m`, `p/m`). [SPEC-161](../../spec/spezifikation.md) Nr. 4, 5 und 7 und [SPEC-162](../../spec/spezifikation.md) Nr. 6 und 7 lassen den Namen aber an jeder Stelle einer Liste, eines Pfads und vor einer Klammer gelten. 15 Mutationen, die das ändern, bleiben in der Suite grün und sind keine Äquivalenzen (die Probe wird rot): `U01` erstes Glied eines Pfads in `$select`/`$orderby` ungeprüft (`$select=1a/p` wäre `wert_nicht_unterstuetzt`, richtig `wert`), `V01b` mittleres Glied dort, `V02b` mittleres Glied eines Rest-Segments, `U22` Glieder eines Rest-Segments mit Klammer (`/S/a-b(x)`), `U13` Länge des Namens vor einer Klammer (`/<129 mal a>(1)`), `U08` Länge 128 eines `$orderby`-Namens mit Richtung, `W-orderby-name-mit-richtung` Name mit Richtung ungeprüft, `W-orderby-erstes`, `W-orderby-letztes`, `V08b` Elemente von `$orderby` an erster, letzter, ab zweiter Stelle ungeprüft, `V07b` Elemente von `$select` ab der zweiten Stelle, `W-key-erster`, `W-key-letzter`, `V05c`, `W-key-ab-zweitem` Namen des ersten, letzten, ab dem dritten, ab dem zweiten Teils eines Schlüssels. Nicht ausreichend ist die Suite auch dort, wo sie rot wird: `U05` (Namen ab dem zweiten Teil nur auf leer geprüft) fällt allein durch `TestKeyLeerraum`. Die Klasse ist (Ort, 11) × (Stellung in der Liste oder im Pfad: erste, mittlere, letzte, einzige) × (Form: mit und ohne Richtung, mit und ohne Klammer, Länge 128 und 129); der Plan nennt sie in Zeile 28 nicht, die Zählung „elf Stellen“ meint den Ort, nicht diese Achse. Der Produktcode ist konform (Abschnitt 3: 0 Abweichungen). | Abschnitt 4.1, 4.4; Probe-Orte `select-mitte`, `orderby-erster`, `key-erster`, `rest-klammer` | Implementer (Fänger als Schleife über Ort × Stellung × Form, [AGENTS.md](../../AGENTS.md) §3.17); alternativ Planner (benanntes Negativ mit Grund und Folge, [AGENTS.md](../../AGENTS.md) §3.15) |
| <a id="v5-2"></a>V5-2 | LOW | **Die Schreibweise der nicht unterstützten Optionen ist nicht gefangen.** [SPEC-162](../../spec/spezifikation.md) Nr. 3: „die Schreibweise zählt“; `$Expand` ist `option_name` (Klasse c), nicht `option_nicht_unterstuetzt`. Die Mutation `U40` (`switch asciiKlein(name)` in `nichtUnterstuetzteOption`) bleibt grün, nicht gleichwertig (`$EXPAND=1` würde 501 statt 400). `TestOptionsklassen` führt `$Top`/`$TOP` nur für die behandelten Namen (`U41`, rot). Die zehn Namen der Liste einzeln entfernt (`U42`, 10 Mutationen): alle rot. | Abschnitt 4.2 | Implementer |
| <a id="v5-3"></a>V5-3 | LOW | **Negativ (5) ist zum dritten Mal zu eng.** Der Plan nennt als ungefangen `Nl` (Buchstabe und Ziffer), `Mc` und `Me` (Ziffer) und `Sc` (Buchstabe). Gemessen in `csdl/namen.go` (23 Mutationen „Kategorie X als Buchstabe“ bzw. „als Ziffer“ in der Suite des Pakets `csdl`): als Buchstabe bleiben `Nl`, `Mc`, `Me`, `Sc`, `Pc`, `Cf`, `Co`, `Sm` grün (8 von 12), rot sind `Mn`, `No`, `Zs`, `Nd`; als Ziffer bleiben dieselben acht grün, rot sind `Mn`, `No`, `Zs`. Es fehlen im Wortlaut `Mc`, `Me` als Buchstabe, `Sc` als Ziffer und `Pc`, `Cf`, `Co`, `Sm` in beiden Rollen: 16 von 23 Mutationen grün. Die Ortsangabe (`outboundtest/metadata_codec_aufloesen.go:324`) und die Folge im Betrieb (ohne Sicherheitswirkung) bleiben richtig. | Abschnitt 5.2 | Planner (Wortlaut des Negativs: „alle Kategorien außer `Mn`, `No`, `Zs`, `Nd` am Anfang“); der Code liegt in einem Slice in `done/` |
| <a id="v5-4"></a>V5-4 | LOW | **Die Zeichenmengen sind Stichproben je Kategorie, keine Überdeckung der Codepunkte.** Sechs Punktmutationen innerhalb einer Klasse bleiben grün: `U17a` U+FDD0 und U+FDEF als ungültig, `U17b` U+0085, `V30b` U+00BF, `V30c` U+2118 und `V30d` der Bereich U+2F800 bis U+2FA1F als (Nicht-)Buchstabe, `H1` die Ziffer `7` am Anfang einer herstellerspezifischen Option. Alle nicht gleichwertig (Probe rot; `U17b` im Volllauf über alle Codepunkte 295 Abweichungen). Die 28 Mutationen „Kategorie X im Wert ungültig“ (`D18`, je Kategorie von `Lu` bis `Co`) sind dagegen alle rot. Folge im Betrieb: keine, solange `unicode.IsLetter`/`IsDigit` und `utf8.ValidString` die Tabellen der Standardbibliothek tragen; eine künftige Handtabelle würde nur an den zehn und 29 Stichproben gemessen. | Abschnitt 4.3 | Planner (akzeptiertes Negativ mit Grund und Folge) |
| <a id="v5-5"></a>V5-5 | INFO | **Zählung im Plan: zehn Stellen, elf Orte.** Plan §3 („Fix-Runde nach der Nachverifikation 4“) und Zeile 28 sagen „elf Stellen im Code“ und nennen im selben Satz neun Stellen der Regel von [SPEC-161](../../spec/spezifikation.md) Nr. 4 und `sonstigerName` als zehnte. Nachgezählt (`grep` auf `einfacherName(`, `namenAnfang(`, `typPraefix(`): `typPraefix`, `namenAnfang` (`mengeOderEntity`), `einfacherName` an sieben Aufrufen (`restForm`, `nacktesWort`, `strukturiere`, Alias in `sonstigerName`, `elementePfad`, `leseSelect`, `leseOrderBy`), dazu der ASCII-Zweig von `sonstigerName` = 10 Stellen; `elementePfad` wird von `$select` und `$orderby` gerufen = 11 Orte. Die Tests haben 10 + 1 Orte; „elf Stellen“ ist ein Wortlaut, kein Fehler im Fänger. | Abschnitt 3 | Planner |
| <a id="v5-6"></a>V5-6 | INFO | **Negativ (3) ist wahr gegen Messung; „weit“ ist eine Fensterbreite, keine Klasse.** Abschnitt 5.1: die Zeilen WA, WN, WD, WE und WRAP des Plans reproduzieren sich mit meinen Wrappern (WRAP mit einem Fenster von 100 atomaren Schritten; Kontrolle 0 von 20 in allen vier Zellen). Mit einem Fenster von 3000 Schritten wird derselbe geschützte Zustand unter dem Detektor auch auf einem Kern in 20 von 20 Prozessen gefangen (der Detektor verlangsamt das Fenster, bis die Laufzeit preemptiert). Die Aussage (c) des Plans bleibt richtig für enge Fenster; „WRAP: 0 von 20 mit `GOMAXPROCS=1`, auch unter dem Detektor“ gilt nur bis zu einer Fensterbreite, die der Plan nicht nennt. | Abschnitt 5.1 | Planner (optional: Fensterbreite nennen) |
| <a id="v5-7"></a>V5-7 | INFO | **Die Nachzüge in fremden Paketen schwächen keinen Fänger.** Abschnitt 5.3: jede Menge wächst um genau ein Element (`ODR-E6005` zwischen `ODR-E6004` und `ODR-E6101`; `ErrURLUngueltig` und `ErrURLNichtUnterstuetzt` in der Liste des Workspace-Dienstes), die Zählungen bleiben feste Literale (34, 38, 37 und 36, `19*24`, `4*(24+4+1)`). Die vier verlangten Mutationen sind rot, vier weitere Ersatz-Mutationen ebenso, kein weiteres Literal 33/22 im Repo (`grep`, vollständiger `make test` ohne Cache grün). | Abschnitt 5.3 | – |
| <a id="v5-8"></a>V5-8 | INFO | **Messung und Zahlen stimmen.** `git diff --numstat 7535bf7..HEAD` über die 14 Pfade des Blocks: Einfügungen 6041, Löschungen 9, **Summe 6050**, gleich Plan und Gate (`Messung 6050 von Marke 9200`, `zeilen=61 nebenlaeufig=2 fakes=1 produktcode=1700 schaetzung=6790`). Die fünf fremden Testdateien liegen nicht im Block (kein Pfad des Blocks deckt `adapters/driven/telemetry/` oder `services/workspace/`); der Plan nennt sie unter „Dateien außerhalb des Blocks“. Die Zahlen des Plans (62 Muster × 10 Orte = 620, 56, 29 × 11 × 4 = 1276, 40) stimmen mit den Tests, jede steht im Test als Zählprobe. | Abschnitt 2 | – |
| <a id="v5-9"></a>V5-9 | INFO | **Zwei Mutationen sind gleichwertig, belegt.** `U15` (`namenAnfang`: `zeichen > 0` statt `i > 0`): die Schleife läuft nur weiter, solange ein Zeichen angenommen ist, `i > 0` und `zeichen > 0` sind dort dasselbe; Suite und Probe grün. `U16` (`dekodiere` ohne `%`: `utf8.ValidString` ohne die NUL-Prüfung): `pruefeArgumente` weist jedes rohe Byte unter 0x20 in `pfad` und `rawQuery` vorher als `model.ErrUngueltig` ab, ein NUL erreicht `dekodiere` nur über `%00`, und das geht den anderen Weg; Suite und Probe grün. Die Befunde V4-1 bis V4-7 sind in Abschnitt 6 aufgelöst. | Abschnitt 4.1 | – |

**Offen aus Report 4, nicht Auftrag dieser Runde (§3.13):** die Planner-Punkte (g), (i), (j), (k), (l) in §6 des Plans habe ich weder geprüft noch bewertet; die Closure ist nicht Auftrag.

**Verdikt: NICHT BESTANDEN**, wegen V5-1 (MEDIUM). Es ist wieder eine Fänger-Lücke, keine Abweichung des Produktcodes: das Probe-Paket findet am unveränderten Code **0 Abweichungen** vom Wortlaut (Abschnitt 3). V4-1 bis V4-4, V4-6 und V4-7 sind im Umfang ihrer Klasse geschlossen (Abschnitt 6); offen ist die Nachbarklasse der Zusage „Name“ (Stellung und Form innerhalb des Orts), die dieser Auftrag verlangt hat zu prüfen. V5-2 bis V5-4 sind LOW und blockieren nicht. Die Entscheidung, ob V5-1 als Fix-Runde (eine Schleife über Ort × Stellung × Form) oder als benanntes Negativ mit Grund und Folge ([AGENTS.md](../../AGENTS.md) §3.15) endet, liegt beim Planner; der Produktcode ist in beiden Fällen unverändert richtig.

## 2. DoD und Gates

- **`make gates`** im Worktree auf `4db72ef`: **Exit 0** (aus der Datei `gates.rc`, nicht über `| tail` maskiert). `a-check` 0 Befunde (Hinweis: 21 Dateien unter `tools/` in keiner Schicht), `baseline-verify` 54 Dateien, `d-check` 443 Dateien (ohne diesen Report), 0 Befunde; `groessenmarke slice-odata-url-pfad.md: ok, Messung 6050 von Marke 9200`; Schema- und Migrationsprüfung grün.
- **Als `CACHED` gelaufene Stufen einzeln ohne Cache nachgefahren** (`docker build --no-cache-filter <Stufe> --target <Stufe>`, Tags `verif5-*`, danach entfernt): `test` Exit 0, `pruefung test: ok` (86,5 s); `race` Exit 0, `pruefung race: ok` (113,4 s); `lint` Exit 0, `0 issues.` (23,1 s).
- **Messung nachgerechnet:** siehe V5-8. Das Delta seit dem Merge (`95eefe3..4db72ef`) ist Plan, Spezifikation, `slice-telemetry-spans.md` und sieben Testdateien; kein Produktcode.
- **Zählstellen von 33, 34, 22, 24 im Repo** (`grep -rnE '\b(33|34|22|24|36|35|37)\b'` über `src`, `spec`, `harness`, `docs` außer den Reports): keine weitere feste Zählung der Codetabelle oder der Fehlerliste außer den in `4db72ef` geänderten (`regeln_test.go` 34, `telemetry_test.go` 34/38/37/36, `kontext_test.go` `19*24`, `reset_test.go` `4*(24+4+1)`); die übrigen Treffer (`baselineimport` „Menge 33“, `textsumme_test.go` „33 Schlüssel“, `outboundtest` „34 bekannte Namen“) zählen anderes. In `done/`, den Beobachtungs-Belegen und den Reports stehen die alten Zahlen als Stand ihres Laufs.
- **Dokumente:** [SPEC-230](../../spec/spezifikation.md) Nr. 8 steht auf 34 Reihen, die Änderungshistorie trägt die Zeile; `slice-telemetry-spans.md` Zeile 40 nennt 34 Codes und 37 Zähler mit Verweis auf `ODR-E6005`.

Alle Punkte der DoD sind `[ ]`, §7 ist „Wird bei der Closure gefüllt“ (Closure ist nicht Auftrag dieser Runde).

## 3. Probe-Paket (Orakel nie der Parser) und die Menge der Namensstellen

**Eigene Zählung der Stellen im Produktcode** (Abschnitt V5-5): 10 Stellen, 11 Orte. Anteile der Zusage: Buchstabe (`L`), Ziffer (`Nd`, nicht am Anfang), `_` (neun Stellen mit der Regel von [SPEC-161](../../spec/spezifikation.md) Nr. 4, [SPEC-125](../../spec/spezifikation.md) Nr. 8 b); ASCII-Buchstabe, Ziffer, `_`, `-`, `.` (herstellerspezifisch, [SPEC-162](../../spec/spezifikation.md) Nr. 3 e); Länge 1 bis 128 Zeichen (Zeichen, nicht Bytes; beim ASCII-Zweig Bytes gleich Zeichen); Stellung im Namen (erstes, inneres, letztes Zeichen); Stellung des Namens (allein, erster, mittlerer, letzter Teil, Glied eines Pfads); Form (mit Richtung, mit Klammer, roh oder `%XX`).

`probe_test.go` im Scratchpad (Paket `urlparser_test` einer Kopie mit Modell, Port und Parser):

| Probe | Menge | Orakel | Ergebnis am unveränderten Code |
|---|---|---|---|
| `TestProbeAlleRunen` | jeder Codepunkt von U+0001 bis U+10FFFF außer Surrogaten und den Strukturzeichen des Orts, je 5 Kandidaten (allein, `a`+x+`a`, x+`a`, `a`+x, `a`+x+`1`), an 15 Orten des Pfads und 22 Orten der Query (37 Orte): 11 120 200 Kandidaten roh, dazu je 3 724 205 mit nur den Bytes ab 0x80 als `%XX` (Großbuchstaben), als `%xx` und mit jedem Byte als `%XX`: rund 22,3 Mio Kandidaten, mehr als 0,5 Mrd Aufrufe | `unicode.Is(L)`, `unicode.Is(Nd)`, `_`; Typpräfix mit `.`; herstellerspezifisch per ASCII-Regel; Folge der Ursachen und Orte von [SPEC-161](../../spec/spezifikation.md) Nr. 4 bis 7, [SPEC-162](../../spec/spezifikation.md) Nr. 3 und 6 bis 8 | 0 Abweichungen (Volllauf 217 s) |
| `TestProbeLaengen` | 334 Muster: `a`, `ä`, U+3042, U+1D400, `_` mal 1, 2, 63, 64, 127, 128, 129, 130, 170; Ziffern von 1 bis 4 Bytes (`0`, U+0663, U+FF10, U+1D7CE), `²`, U+2167, U+0301, `-`, `.` hinter `a` in 1, 126 bis 129; gemischte Breiten 127 bis 129; ein Störzeichen an 8 Stellen einer 130-Zeichen-Folge in je 2 Füllungen; × 4 Formen, an allen Orten | wie oben | 0 Abweichungen |
| `TestProbeKombinationen` | 60 000 Zufallsfälle aus einem Ort des Pfads, 1 bis 3 Orten der Query (höchstens je ein `$select` und `$orderby`), 18 Namen (gültig, `-`, Ziffer vorn, 128, 129, mehrbytig, U+FFFD, `²`); Orakel: erster Verstoß der Klasse `ungueltig` entscheidet, sonst der erste 501-Verstoß | wie oben | 0 Abweichungen |
| `TestProbeMini` | jeder 61. Codepunkt und jede Grenze der Klassen `L`/`Nd`/sonst, 4 Formen, 4 Kodierungen | wie oben | 0 Abweichungen |

**Dekodier-Orte** (Pfad-Segment 1 und 2, Name und Wert einer Option) sind über die vier Kodierungen derselben Kandidaten gedeckt: gleiches Ergebnis roh und kodiert an allen Orten, 0 Abweichungen. **Optionsnamen außerhalb von `sonstigerName`:** Abschnitt 4.2.

Orte der Probe, die der Fänger des Repos nicht führt und die erst durch die Mutationen von V5-1 nötig wurden: `key-erster`, `select-erster`, `orderby-erster`, `orderby-erster-desc`, `select-mitte`, `rest-klammer`, `rest-erster-klammer`, `rest-mittlerer`, `select-pfad1` bis `select-pfad-mitte`, `orderby-pfad1` bis `orderby-pfad3-desc`.

## 4. Mutationen am Parser

Jede Mutation: Kopie der Basis, eine Änderung, `go test` über das Paket `urlparser` ohne die Probe. „Grün“ heißt: die Suite des Repos bleibt grün. 94 Mutationen auf Elementen, die weder Verifier noch Implementer noch Architect genannt haben, dazu 11 am Namen der herstellerspezifischen Option und am gültigen UTF-8.

**4.1 Name: Länge, Zeichen, Stellung (24 Mutationen, Gruppe 1).** Rot: 16 (`U02`, `U03`, `U04c` Glieder des Rest-Segments ohne das erste, nur das erste, ohne das letzte; `U04`, `U04b` Glieder des Pfads von `$select`/`$orderby` nur das erste, ohne das letzte; `U05` Namen ab dem zweiten Schlüsselteil nur auf leer geprüft, allein durch `TestKeyLeerraum`; `U06`, `U07`, `U11` Länge 127 bei `nacktesWort`, `$select`, Alias; `U09` Ziffer des Typpräfix nur ASCII; `U10` `Nl` im Typpräfix; `U17c` U+1F600 ungültig; `U24` Pfad mit drei Gliedern abgelehnt; `U30` U+00B7 als Buchstabe; `U31` Länge in Bytes; `U32` nur BMP). Grün und nicht gleichwertig: `U01`, `U08`, `U13`, `U22` (V5-1) und `U17a`, `U17b` (V5-4). Grün und gleichwertig: `U15`, `U16` (V5-9).

**4.2 Optionsnamen (26 Mutationen, Gruppe 2).** `U41` behandelt-Namen ohne Schreibweise: rot. `U40` nicht unterstützte Namen ohne Schreibweise: **grün** (V5-2). `U42` jeder der zehn Namen der Liste einzeln entfernt: 10 von 10 rot. `U43` jeder der sieben behandelten Namen unbekannt gemacht: 7 von 7 rot. `U44` jede der sieben Optionen darf zweimal vorkommen: 7 von 7 rot.

**4.3 Kombinationen und Punkte (8 Mutationen, Gruppe 4; 28 der Gruppe 6; 11 der Gruppe 10).** Gruppe 4: `V05c`, `V07b`, `V08b`, `V01b`, `V02b` grün (V5-1), `V30b`, `V30c`, `V30d` grün (V5-4). Gruppe 6 (`D18`, Zeichen jeder Kategorie `Lu` bis `Co` und die Zeichen U+0081 bis U+009F im Wert ungültig): 28 von 28 rot. Gruppe 10: `H2` bis `H8` (Namen der herstellerspezifischen Option: `-` und `.` verboten, Länge 127 und 129, `_` nur am Anfang, Großbuchstabe nur am Anfang, leerer Name gültig) 7 von 7 rot, `H1` (Ziffer `7` am Anfang erlaubt) grün (V5-4); `F1`, `F2`, `F3` (U+FFFD, U+FFFE/U+FFFF, U+10FFFF ungültig) 3 von 3 rot durch `TestUtf8Rand`.

**4.4 Stellung in der Liste (8 Mutationen, Gruppe 7).** Rot: `W-select-erstes` und `W-select-letztes` (durch `TestSelect`). Grün, nicht gleichwertig (Probe rot): `W-key-erster`, `W-key-letzter`, `W-key-ab-zweitem`, `W-orderby-erstes`, `W-orderby-letztes`, `W-orderby-name-mit-richtung` (V5-1). Meine Probe hatte für `key-erster`, `select-erster` und `orderby-erster` zuerst keinen Ort; die Mutationen zeigten die Lücke, sie ist geschlossen und am Code wie an den Mutationen gemessen.

Zusammen 105 Mutationen: **22 grüne, nicht gleichwertige** (15 V5-1, 1 V5-2, 6 V5-4), 2 gleichwertige (V5-9), 81 rot.

## 5. Nebenläufigkeit, Negative und Nachzüge

### 5.1 Negativ (3): Zustand hinter dem Feld `filter`

Wrapper in `WithFilter` (Zustand `last`, geschrieben vor, gelesen nach einem Fenster, der Leser bekommt `last`), je Zelle 20 Prozesse mit `-test.count=1`, `-test.run ^TestGleichzeitigMitFilter$`, ein Binary ohne und eines mit `-race`, `/proc/loadavg` 3,99 bis 7,65.

| Variante (Fenster, Schutz) | vier Kerne | `GOMAXPROCS=1` | `-race`, vier Kerne | `-race`, ein Kern |
|---|---|---|---|---|
| Kontrolle (keine Mutation) | 0 von 20 | 0 von 20 | 0 von 20 | 0 von 20 |
| WA (3000 Schritte, ungeschützt) | 20 von 20 | 0 von 20 | 20 von 20 | 20 von 20 |
| WN (kein Fenster, ungeschützt) | 0 von 20 | 0 von 20 | 20 von 20 | 20 von 20 |
| WH (100 Schritte, ungeschützt) | 20 von 20 | 0 von 20 | 20 von 20 | 20 von 20 |
| WD (WA mit `runtime.Gosched()`) | 20 von 20 | 20 von 20 | 20 von 20 | 20 von 20 |
| WE (kein Fenster, zwei Abschnitte unter Sperre) | 20 von 20 | 0 von 20 | 20 von 20 | 0 von 20 |
| WRAP (100 Schritte, zwei Abschnitte unter Sperre) | 20 von 20 | 0 von 20 | 20 von 20 | 0 von 20 |
| WRAP-weit (3000 Schritte, zwei Abschnitte unter Sperre) | 20 von 20 | 0 von 20 | 20 von 20 | **20 von 20** |

Die Zeilen WA, WN, WD, WE, WRAP des Plans reproduzieren sich; die letzte Zeile ist V5-6. Unter `-race` mit vier Kernen sind Wirkung und Detektor nicht getrennt (20 von 20); WE und WRAP sind geschützt, auf einem Kern sieht weder die Wirkung noch der Detektor etwas (0 von 20).

### 5.2 Negativ (5): `csdl/namen.go`

23 Mutationen am Ausdruck der Funktion `einfacherName` (Kategorie X als Buchstabe, als Ziffer), Lauf des Pakets `csdl` (enthält den Fänger `outboundtest`): siehe V5-3. Rot: `Mn`, `No`, `Zs` in beiden Rollen, `Nd` als Buchstabe; grün: `Nl`, `Mc`, `Me`, `Sc`, `Pc`, `Cf`, `Co`, `Sm` in beiden Rollen.

### 5.3 Nachzüge in `telemetry` und `workspace`

Alle Mutationen an Kopien, Pakete `telemetry`, `workspace` und `model`, wo nichts anderes steht.

| Mutation | roter Test |
|---|---|
| `X1` Reihe `ODR-E6005` aus der Tabelle des Modells (`fehler.go`) | 12 rote Tests, darunter `TestFehlercodeForm`, `TestReihenFest`, `TestCountErrorJeCode` |
| `X2` Zähler `ODR-E6005` im Adapter nicht angelegt | 9 rote Tests, darunter `TestReihenFest`, `TestCountErrorJeCode`, `TestSnapshotVollstaendig` |
| `X3` Zähler `ODR-E9001` (letzter Code) nicht angelegt | 9 rote Tests, dieselben |
| `X4` Fenster in `Snapshot` (Sperre freigegeben und `runtime.Gosched()`) vor `ODR-E6004`, vor `ODR-E6005`, vor `ODR-E6101`, also um `ODR-E6005` | `TestSnapshotZeitpunkt`, je 20 von 20 Prozessen bei vier Kernen und bei `GOMAXPROCS=1`, Kontrolle 0 von 20 |
| `Y1`, `Y2` Eintrag `ErrURLUngueltig` bzw. `ErrURLNichtUnterstuetzt` aus der Liste in `helfer_test.go` entfernt | je 3 rote Tests (`TestTabellenFehlerVollstaendig`, `TestFehlerAbbildungJedeStelle`, `TestFehlerInDerTransaktionDesReset`) |
| `Y3` `ODR-E6005` aus `alleCodes()` in `regeln_test.go` entfernt | `TestFehlercodeForm` |
| `Y4` Zählung in `reset_test.go` auf `4*(23+4+1)` | `TestFehlerInDerTransaktionDesReset` |

Die Mengen wachsen in der Reihenfolge der Tabelle ([SPEC-109](../../spec/spezifikation.md) Nr. 2: `ODR-E6005` hinter `ODR-E6004`); die Zählungen sind Literale, keine abgeleiteten Längen. Der Plan §6 (p) trägt den Ausgang „entschieden“, den Beleg (Mutationen M1 bis M6) und die Adresse (Planner, Integrationsschritt).

## 6. Auflösung der Befunde aus Report 4

- **V4-1** (ASCII-Anteile `_`, `-`, Länge in Zeichen an den Orten): geschlossen für die Orte. `TestNamenAchsen` 62 × 10 = 620; die Mutationen `U06`, `U07`, `U11`, `U31` an Stellen, die der Befund nicht nannte, sind rot. Offen als Nachbarklasse: V5-1.
- **V4-2** (elfte Stelle `sonstigerName`): geschlossen. `TestHerstellerNamen` 56; sieben Mutationen `H2` bis `H8` rot; `H1` (Ziffer `7` am Anfang) grün, V5-4.
- **V4-3** (Kategorien): geschlossen für die Kategorien. `TestUnicodeKategorien` 29 × 11 × 4 = 1276; die 28 Mutationen `D18` rot.
- **V4-4** (U+FFFD und gültiges UTF-8 am Rand): geschlossen. `TestUtf8Rand` 40; `F1`, `F2`, `F3` rot; Punkte abseits der zehn Stichproben V5-4.
- **V4-6** (Negativ (3)): Abschnitt 5.1, wahr; V5-6.
- **V4-7** (Negativ (5)): berichtigt und trotzdem unvollständig, V5-3.

## 7. Nicht gelaufen und Vorbehalte

- **`make gegenproben`, Stempel:** nicht gefahren (nicht Auftrag). Der Stempel `harness/gegenproben.stempel` ist nicht geprüft.
- **Die Stufe `build`** habe ich nicht ohne Cache wiederholt; der Produktcode ist unverändert.
- **Unicode-Tabelle:** das Orakel der Probe nutzt die Tabellen von Go (`unicode.Is`) selbst; eine Abweichung zwischen ihnen und dem Standard sähe es nicht (wie in Report 4).
- **Raten:** 20 Prozesse je Zelle; „0 von 20“ ist keine Obergrenze der Wahrscheinlichkeit. Die Raten in 5.1 sind unter fremder Last gemessen (`/proc/loadavg` 3,99 bis 7,65).
- **Äquivalenzen** (`U15`, `U16`) sind begründet und von Suite und Probe gestützt, nicht bewiesen.
- **Mutationen nur auf `urlparser` und `csdl/namen.go`:** die Eingänge der anderen Pakete des Blocks (`model`, `outbound`, `baselineimport`) habe ich in dieser Runde nicht neu mutiert (Delta enthält dort keine Änderung).
- **Meine Probe hat die Lücken von V5-1 erst durch die Mutationen gefunden** (Orte `key-erster`, `select-erster`, `orderby-erster`); die Menge der Stellungen ist damit aus dem Wortlaut gebildet, aber nicht aus der Zahl der Orte allein ableitbar. Eine weitere Achse, die ich nicht kenne, kann bleiben.
- **Rohbytes (§3.18):** mein Probe-Paket enthält U+0301, U+FFFD und U+00A0 als Rohbytes in Go-Quelltext (vom Schreibwerkzeug aus den Escape-Folgen umgesetzt); gegengelesen mit `od -c` (`314 201`, `357 277 275`, `302 240`), die Werte stimmen. Dieser Report ist gegengelesen: `LC_ALL=C grep -nP` auf Backslash, `u` und vier Hex-Ziffern: 0 Treffer; die ersten drei Bytes `23 20 56` (kein BOM); die Bytefolgen von U+FFFD, U+00A0 und U+0301 stehen nicht darin (`LC_ALL=C grep -cF`, dieselbe Suche trifft sie im Probe-Paket 2, 1 und 3 mal); die 61 Zeilen mit Bytes ab 0x80 tragen deutsche Umlaute und die Zeichen `„“`.

## 8. Register-Kandidaten (nur genannt, nicht geschrieben)

1. **Eine Allaussage über „Name“ ist an den Orten gedeckt und nicht in der Stellung und Form innerhalb des Orts** (V5-1): vierte Runde in Folge, in der die Menge der Klasse eine Achse auslässt (V3-1 Zeichenanteile, V4-1 ASCII und Länge, V5-1 Stellung und Form). Wiederholung des Musters von [AGENTS.md](../../AGENTS.md) §3.17: die Menge entsteht aus dem Wortlaut („an jeder Stelle einer Liste“), nicht aus der Zahl der Orte.
2. **Ein Negativ benennt die Klasse der ungefangenen Fälle zu eng** (V5-3; zuvor V4-6, V4-7): dritte Runde mit einer Allaussage im Negativ („ungefangen sind …“), die eine Messung überholt.
3. **Ein Probe-Paket des Verifiers braucht für jede Liste einen Ort je Stellung** (die Lücke meiner Probe, Abschnitt 7): die Mutationen „erstes/letztes Element ungeprüft“ gehören in die Standardmenge der Verifikation einer Namensregel.

# Verifikations-Report: slice-handbuch-geruest — 2026-10-30

**Rolle:** Verifier, frischer Kontext (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-200` und `AGENTS.md` §3.9 bis §3.18). Kein Review, keine Reparatur; Code, Plan und Vertrag sind unverändert. Eingang: [Review 1](2026-10-30-slice-handbuch-geruest-review.md) (F-1 bis F-15) und [Review 2](2026-10-30-slice-handbuch-geruest-review-2.md) (R-1 bis R-9) nur als Eingang, nicht als Beweis.

**Gegenstand:** Zweig `slice-handbuch-geruest`, HEAD `c7ceefc` (Fix-Commits `ad8b5f0`, `0d90db1`, `6fd6463`, `c7ceefc`; Planner `5867cfb`), Basis des Slice `975228d`, `main` über `c5d0cdd` aufgenommen. Plan: `docs/plan/planning/done/slice-handbuch-geruest.md`. Arbeitsbaum vor und nach den Läufen ohne Änderung (`git status` leer).

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Methode.** Keine Zahl und keine Mutation des Implementers oder des Reviewers übernommen. Scratchpad `verif-geruest/` (flüchtig, nichts im Repo außer diesem Report): `make gates` im Zweig mit echtem Exit-Code; die Stufe `handbuch` ein zweites Mal ohne Cache; eine Kopie des Baums mit einem Wort aus A im Handbuch; eine Wegwerfkopie von `tools/pruefung` je Mutation (literales `sed` auf der Kopie), gefahren im Go-Image des `Dockerfile` (Digest gepinnt, `--network none --memory=4g --memory-swap=4g --cpus=4`, Quelle `:ro`, `/tmp` als `tmpfs`), alle Läufe unter `flock /tmp/odata-recorder-heavy.lock`, jede Einzelmessung als eigener Aufruf; ein Probe-Paket (`probe_test.go`, `probe2_test.go`, Saat ab 200000, Orakel aus dem Text von `SPEC-200` und von Hand, nichts aus den Tests des Implementers) und 40 eigene Mutationen an Elementen, die weder Review noch Implementer nannten.

---

## 1. Verdikt

**NICHT BESTANDEN.** Es gibt kein HIGH. Die Lieferung erfüllt die DoD-Punkte, die der Verifier belegt (Abschnitt 3), `make gates` ist grün (Exit 0), das Programm verhält sich in allen geprobten Klassen nach dem Wortlaut von `SPEC-200` (Abschnitt 4), und die offenen Punkte R-3, R-6, R-7, R-9 des Re-Reviews sind gelöst (Abschnitt 5). Zwei Befunde stehen der Closure im Weg: **V-1 (MEDIUM)**, acht Mutationen an sieben Zusagen des Vertrags bleiben in der ganzen Suite grün, und **V-2 (MEDIUM)**, drei im Code entschiedene Randformen ohne Ort im Vertrag und ohne Fänger. Beide sind klein (je Element ein Fall); wie in den Runden davor zählt ein Fix erst, wenn er die Klasse deckt (`AGENTS.md` §3.17). Eine Re-Verifikation prüft nur V-1 und V-2.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| V-1 | MEDIUM | **Zusagen des Vertrags ohne Fänger.** Acht Mutationen an sieben Zusagen von `SPEC-200` lassen die ganze Suite (`go test` über `tools/pruefung`) grün, der Code bricht jeweils den Wortlaut: (a) Nr. 4 d, Form des Stands: ohne die Prüfung des zweiten Striches gilt `2026-10x30`; Z11 führt 28 Elemente, keines mit falschem Zeichen an Stelle 8 (Mutation s4). (b) Nr. 4 c, Reihenfolge der Gründe `Signal`, `Exit <n>`, `Fehlerausgabe nicht leer`, `Standardausgabe`: bei Exit 1 mit nicht leerer Fehlerausgabe meldet der Code `Fehlerausgabe nicht leer` statt `Exit 1`; Z10 führt die Gründe nur einzeln (s14). (c) Nr. 4 a, die Überschrift beginnt mit `# `: `#Titel` gilt (t4). (d) Nr. 8: die Zeichenfolgen `-t odata-recorder:build` und `-o /out/odata-recorder`; ohne `-t ` bzw. `-o ` bleibt Z28 grün, weil die Fixtures nur „vorhanden“ oder „fehlt ganz“ kennen (t11a, t11b). (e) Nr. 6, die Beispiel-Datei ist UTF-8: die Prüfung kann entfallen (t21). (f) Nr. 3 c, „je Zeile und Zeichen einmal“: die Menge der gesehenen Zeichen je Datei statt je Zeile meldet ein Zeichen, das in zwei Zeilen steht, nur einmal (m7). (g) Nr. 4 a, der Kopf endet an der ersten Zeile `## `: mit `##` ohne Leerzeichen beendet eine Zeile `###` oder `##x` den Kopf (m8). Klasse: Fänger führt nur eine Form der Eingabe; Zusage ohne Fänger (`AGENTS.md` §3.15, §3.17). | Abschnitt 6, Mutationen mit Name; Probe auf dem unveränderten Programm grün (Abschnitt 4) | Implementer |
| V-2 | MEDIUM | **Randformen im Code entschieden, ohne Ort in `SPEC-200` und ohne Fänger** (`AGENTS.md` §3.16). (a) Die Zeile des Befundes `Kopf: <Schlüssel> doppelt` ist die erste Zeile des Schlüssels; Nr. 2 d nennt nur Befunde „ohne Zeile“ (Mutation v4: letzte Zeile, grün). (b) Die Befunde `Software-Version … weicht von der Ausgabe von version ab` und `version: Ausgabe nicht lesbar` tragen die Zeile des Schlüssels `Software-Version` (ohne gültigen Schlüssel 0); Nr. 2 d zählt „die Befunde … der Läufe“ zur Zeile 0. Probe: ein Kopf mit `Zeile 2: Kopf: Zeile ohne Schlüssel` und abweichender Version gibt die Zeile 2 vor `weicht … ab`; mit Zeile 0 stünde es davor (Mutation x1: Zeile 0, grün). (c) Ein Inline-Code-Span mit einem Leerzeichen nur an einem Ende: `make nix ` wird geprüft, ` make nix` nicht (CommonMark entfernt ein Leerzeichen nur an beiden Enden); Nr. 9 b nennt CommonMark nur für Öffnen und Schließen (Mutation m5: Entfernen an einem Ende, grün). | Abschnitt 6; Probe `TestPvZ2Reihenfolge`, `TestPvSpanRand` | Architect (Ort in `SPEC-200` Nr. 2 d und 9 b), danach Implementer (Fänger) |
| V-3 | LOW | Der Plan nennt zwei Zustände, die nicht mehr gelten (`AGENTS.md` §3.9): §4 „Start vor `slice-groessenmarke-werkzeug` in `done/`“ sagt, der Werkzeug-Slice liege noch nicht in `done/`; §6 „Befund des Implementers …“ sagt, `make gates` sei im Zweig rot, weil der Plan des anderen Slice mitgemessen werde. Seit dem Merge `c5d0cdd` liegt er in `done/`, `make gates` ist grün, die Messung nennt nur diesen Slice. Dazu: §3 „Reihenfolge der Commits“ (zuerst Handbuch, dann Gate) und das Risiko „erster Code-Commit macht `make gates` rot, entfallen durch die Reihenfolge“ gelten für zwei Commits, geliefert ist ein Commit `1667906` mit beidem. | Plan §4, §6, §3; Verzeichnis `done/`; Abschnitt 7 | Planner |
| V-4 | INFO | Handbuch, Prosa (R-7 ist gelöst, der Rest ist lesbar, nicht prüfbar): (a) „Lösung: Rufe das Binary ohne Argument oder nur mit `version` auf“ ist eine allgemeine Aussage, abgeleitet aus zwei gebundenen Beispielen (N2); sie ist gegen `Run` gelesen wahr. (b) „Setze die Software-Version mit der Variable `VERSION` … Ohne Angabe ist die Version `dev`“: belegt durch `VERSION ?= dev` in `harness/mk/go-gates.mk` und den Kopf; `make build VERSION=1.2.3` fährt kein Gate (nur `dev`). (c) Vor der Überschrift `## Bauen` fehlt die Leerzeile (Zeile 10); das Gate fasst beide Zeilen zu einem Absatz, ohne Folge. | `docs/user/benutzerhandbuch.md` Zeilen 10, 21, 59; `src/bootstrap/bootstrap.go` | Reviewer, Abnahme (N2) |
| V-5 | INFO | Z26 (Zeitgrenze, nebenläufig): der Plan führt als Rate einen Prozess mit `-count=20` und drei Prozesse mit `-count=1` (3 von 3); für Z49 steht die Rate mit 20 Prozessen und Last, für Z26 nicht. Gemessen im Verifier: Abschnitt 6. | Plan §3 „Raten“ | Planner (Zahlen eintragen) |
| V-6 | INFO, bekannt | Nicht gedoppelt: Zeitbudget der Gegenproben ([ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)), endgültige Fallnummern 121 bis 130, `docs/plan/planning/README.md` Nr. 6 (SPEC-129), Nachzug bei den Nehmern. | Plan §6 | Planner, Integrationsschritt |

## 3. DoD, Punkt für Punkt

| DoD-Punkt | Befund | Beleg |
|---|---|---|
| Handbuch mit Kopf (Software-Version, Gültigkeitsbereich, Stand) und den vier Kapiteln (Einleitung, Bauen, Aufruf und Version, Fehlerbehebung) | erfüllt | `docs/user/benutzerhandbuch.md` gelesen; Kopf gleich der Ausgabe von `version` (Stufe `handbuch` ohne Cache: `pruefung handbuch: ok`, 0,19 s) |
| Nur Ist-Zustand: nichts Geplantes, keine Chronik, keine Kennung von Welle, Slice, ADR, LH, SPEC | erfüllt für die Menge der Wörter | eigenes `grep` über die Listen des Auftrags (Planung, Ausblick, Chronik, Kennungen): Treffer nur im Wort „Software-Version“ und im Beispiel `version x`; Rohbytes: nur UTF-8-Umlaute, kein BOM, kein NBSP, keine Escape-Folge, kein `\r` (auch in den geänderten Dateien von `tools/pruefung`, `tools/gegenproben`, `SPEC-200`); Ausblick ohne Stichwort liest der Reviewer (V-4) |
| Jedes Beispiel ist eine gebundene Datei und läuft gegen das gebaute Binary | erfüllt | vier Dateien unter docs/user/beispiele, Bindungszeilen gültig; die Stufe fährt sie; `make gates` Exit 0 |
| Anpassung des Standards (Abschnitt „Anpassung für dieses Repo“, je Abschnitt 1 bis 14 eine Entscheidung) | erfüllt | `docs/user/benutzerhandbuch-standard.md` Zeilen 363 bis 396: Regel plus 14 Zeilen; die Datei ist im Bereich `975228d..HEAD` unverändert (der Abschnitt stammt aus `6b1d0f8`); ob der Text oberhalb des Abschnitts der Quelle im Nachbar-Repo gleicht, ist nicht abgeglichen |
| Diff-Vorschlag für `AGENTS.md` liegt dem User vor | nicht belegbar | Scratch-Datei des Architect; §6 führt den Punkt mit Ausgang *offen*, Adresse User; Closure-Vorbedingung |
| Gate `make handbuch-check`: Programm, Stufe, Ziel in `GATE_CHECKS`, Allowlist, Sensors-Zeile, H1 bis H10 | erfüllt | `GATE_CHECKS := … groessenmarke handbuch-check lint-profil …` (`make -pn`); `make -n gates` nennt `--target handbuch`; `.dockerignore` trägt sechs Einträge (`docs/user`, `Makefile`, `d-check.mk`, `repo.mk`, `Dockerfile`, `harness/mk`); Zeile in `harness/README.md` §Sensors, Zellen 23, 191 und 88 Zeichen (Grenze 200); `make help` nennt das Ziel; H1 bis H10 grün gegen das gebaute `pruefung` |
| `make gates` grün | erfüllt | Exit 0 (Abschnitt 7) |
| `make docs-check` meldet das Handbuch ohne Befund | erfüllt | `d-check: 384 Datei(en) geprüft, 0 Befund(e)` im Lauf von `make gates` |
| Je Zeile der Tabelle in §3 eine Mutation rot gesehen, je Muster eine | nicht vollständig selbst nachgefahren | Abschnitt 4: Namen aller Tests des Plans existieren im Code (0 fehlende); 40 eigene Mutationen, 27 rot; V-1 und V-2 nennen die grünen |
| Review durchgeführt, Report liegt vor | erfüllt | zwei Reports unter docs/reviews |
| Doku-Update: `harness/README.md` §Sensors und *Nicht behauptet*, `SPEC-200` *Stand*, `SPEC-036`, `SPEC-102` | erfüllt | `git diff 975228d..HEAD` der drei Verträge und der README gelesen: `SPEC-036` nennt `handbuch` in Aufrufen, Randformen 1, 10, 12, *Stand* und *Grenze*; `SPEC-102` Fälle 121 bis 130, Randform 13 und *Stand*; *Nicht behauptet* nennt Verdrahtung, Aufruf von `docker`, Ausblick ohne Stichwort |
| Closure-Notiz, Register, Ausgänge in §6, drei Paarungen | offen | §7 *Wird bei der Closure gefüllt*; Closure-Vorbedingungen Abschnitt 9 |

## 4. Plan gegen Code: Z1 bis Z50, N1 bis N6, Klassen der Fixes

**Abgleich.** Jede Zeile der Tabelle verweist auf eine Nummer von `SPEC-200`, die es gibt; die Namen der Tests im Plan (alle `Test…` des Abschnitts *Bericht des Implementers* und der Nachträge) stehen im Code. Die Zählung des Vertrags habe ich aus dem Text des Vertrags geparst, nicht aus dem Test: 44 Muster, 104 Formen (A 16/36, B 14/44, C 14/24).

**Probe-Paket auf dem unveränderten Programm** (Saat ab 200000; alle grün, Lauf `probe3.log`):

| Klasse (Zusage) | Generator | Orakel | Elemente |
|---|---|---|---|
| „getrimmt“, Nr. 3 b: 19 Leerraum-Zeichen × sechs Stellen (Wert, Schlüssel, leere Zeile des Kopfes, Absatzgrenze, Schluss-Zaun, Info-Zeichenkette) | alle 19 × 6, dazu das Zeichen als alleiniger Wert (`Kopf: Stand leer`) | die 19 Codepunkte von Hand (16 Zs ohne U+0020, U+2028, U+2029, Tabulator) | 133 |
| Leerraum in *V* der Ausgabe von `version` (Nr. 4 c) | 19 × {innen, vorn, hinten} | `version: Ausgabe nicht lesbar` | 57 |
| Span `make`, Nr. 9 b, 14 d | 19 Trenner × {`make`+Z, `make `+Z, `make `+Z+Z, `make`+Z+`nix`, `xmake nix`} + `make `+Z+`nix`, dazu Backtickfolgen | Tabelle von Hand | 114 |
| Reihenfolge der Befunde, Nr. 2 d | 40 Saaten: Abschnitte (Planung, Chronik, Kennung, Make, Bindung ohne Zaun, Zeichen) in zufälliger Reihenfolge, 1 bis 5 Dateien mit Namen aus `a.txt`, `a-b.txt`, `a0.txt`, `ab.txt`, `version.txt`, `b.txt`, `a-a.txt` in zufälliger Bindungsreihenfolge, Kopf-Befund | `sort` nach (Nummer, Zeile, Name in Bytefolge), Zeilen von Hand gezählt | 40 Läufe |
| Bindungszeile und Dateiname, Nr. 5 b/5 f | 400 Namen aus 17 Bausteinen (`a`, `z`, `0`, `9`, `-`, `--`, `_`, `.`, `A`, `Z`, Leerzeichen, Umlaut, `~`, `+`, …) und 12 Endungen (`.txt`, `.TXT`, `.tat`, `.tx`, `.txtx`, `.txt.md`, …); je Name als Datei im Verzeichnis und als Pfad der Bindung; dazu 7 Reste hinter der Klammer | eigene Prüfung ohne `regexp` | 399 + 7 |
| Mengen A, B, C, Nr. 10 bis 12 | alle Formen aus dem **Text des Vertrags** × 7 Stellungen | Vertragstext | 728 |
| C4 `adrs?` | 18 Suffixe × 9 Präfixe × 7 Stellungen (`adr`, `adrs`, `adrss`, `adrx`, `adr-0001`, `adr_`, `madr`, `ädr` …) | Handorakel mit Buchstabengrenze | 1134 |
| Ausgangswert, Nr. 6 c | 34 Grenz- und Fremdformen (255/256, `00`, `0255`, 2^64, 2^64+5, 20 mal `9`, arabisch-indische Ziffer, …) + 40 zufällige Ziffernfolgen der Länge 1 bis 24 (Saat 200000) | Handorakel | 74 |
| Blöcke je Datei, Nr. 5 d | 60 Läufe: 1 bis 6 Blöcke × 7 Abweichungsarten, Zeilen der Befunde gezählt | erwartete Zeilen | 60 |
| Toter Symbolischer Verweis als Handbuch (R-9) | – | Nr. 3 a: `Eingabe: Handbuch fehlt`, Exit 2 | 1 (grün) |

**Gegenproben H1 bis H10.** Die zehn Fälle sind gegen das aus der Quelle gebaute `pruefung` grün. Stichprobe mit Mutationen: C6 entfernt (H4), Versionsvergleich entfällt (H5), Make-Ziel nicht geprüft (H8), Zitat auch von A und B getroffen (H10): jeweils `TestFaelleDesHandbuchsLaufenGegenPruefung` rot. `make gegenproben` selbst ist nicht gelaufen (Zeitbudget, V-6).

## 5. Die offenen Punkte der Re-Review

| Punkt | Stand | Beleg |
|---|---|---|
| R-3 / Z49 (Grenze 2,5 s, akzeptiertes Negativ 1 s bis 2,5 s) | gelöst | `SPEC-200` Nr. 7 e nennt „höchstens 1 s“, Raster Zeile *Lauf × offener Strom*; Plan Z49 und Nachtrag 3 (g) nennen Grund (Last 17 bis 80 auf 20 Kernen) und Folge (bis 1,5 s länger, Ausgabe unverändert). Rate selbst gemessen (Abschnitt 6): 3 s, 10 s, entfällt, `ErrWaitDelay` als Fehler rot in 20 von 20; 2 s und 1,5 s rot in 0 von 20, das Negativ ist so, wie der Plan es nennt |
| R-6 (Marke fest bei 6900, Hinweis nach `SPEC-128` Nr. 5 c) | gelöst | Block: `schaetzung: 5085`, `marke: 6900`, kein Eintrag `ersetzt`; `make groessenmarke`: `Hinweis: Tabelle ergibt 5890, Schätzung 5085`, `ok, Messung 3864 von Marke 6900` (56 %), Messpunkt `zeilen=56 nebenlaeufig=2 fakes=2 produktcode=1000 schaetzung=5890 marke=6900 messung=3864` |
| R-7 (Handbuch-Prosa ohne Verhaltenszusage) | gelöst, Rest V-4 | der Satz „Die Meldung nennt das erste Argument …“ ist ersetzt; die Prosa beschreibt, was die Beispiele zeigen |
| R-9 (Symlink-Negativ) | gelöst | Probe: toter Verweis gibt Exit 2 und `Handbuch fehlt`; das Negativ steht mit Grund und Folge im Plan (Nachtrag 3 (f)) und im Vertrag (Nr. 3 a) |
| R-1 (getrimmt), R-2 (Span), R-4 (Reihenfolge), R-5 (Name, Bindung) gegen die Klasse | gelöst | Probe des Abschnitts 4 grün; Mutationen m1 (Leerraum von *V* auf drei Zeichen), m2a/m2b/m2c (Info ohne U+205F, Schlüssel ohne U+2008, Schluss-Zaun ohne U+1680: Zeichen, die weder Review noch Implementer einzeln nannten), m4 (Backtickfolge `>=`), s25/s27 (Name ohne `z`, ohne `9`), m11 (Verzeichnis `d.txt` als Beispiel-Datei) alle rot |

## 6. Mutationen und Raten

Messender: der Verifier; Prozess je Mutation das Testbinary einmal gebaut, danach 20 Prozesse mit `-test.count=1` und `-test.run` nur auf den benannten Test (Z49: `TestHandbuchKindprozessHaeltDenStrom`; Z26: `TestHandbuchZeitgrenze`, `TestHandbuchZeitgrenzeWert`). Last: Load Average 1/5/15 Minuten vor dem Lauf, Host mit 20 Kernen, andere Läufe in der Schlange.

| Zusage | Mutation | rot in … von 20 Prozessen | Last vor dem Lauf |
|---|---|---|---|
| Z49 | unveränderter Stand | 0 | 25,4 / 30,6 / 67,7 |
| Z49 | `WaitDelay` 3 s | 20 | 11,7 / 19,9 / 53,3 |
| Z49 | `WaitDelay` 10 s | 20 | 19,5 / 16,8 / 34,9 |
| Z49 | `WaitDelay` entfällt | 20 | 12,9 / 16,9 / 29,5 |
| Z49 | `ErrWaitDelay` als Fehler | 20 | 12,3 / 14,2 / 24,3 |
| Z49 | `WaitDelay` 2 s (akzeptiertes Negativ) | 0 | 18,2 / 15,6 / 22,4 |
| Z49 | `WaitDelay` 1,5 s (akzeptiertes Negativ) | 0 | 17,8 / 13,4 / 19,4 |
| Z26 | unveränderter Stand | 0 | 10,0 / 13,9 / 17,9 |
| Z26 | Zeitgrenze entfällt (1000 Stunden dazu) | 20 | 18,6 / 15,2 / 17,3 |
| Z26 | Prozess nicht beendet (`exec.Command` statt `CommandContext`) | 20 | 10,1 / 12,3 / 15,6 |
| Z26 | Wert 10 s zu 1 s | 20 | 11,4 / 10,7 / 13,1 |

Eigene Mutationen an Elementen, die weder Review noch Implementer nannten (Suite ohne mein Probe-Paket; 40 ausgewertet, 27 rot, 11 grün mit Zusagebruch, 2 grün als benanntes Negativ):

| Rot | Grün (Zusagebruch) |
|---|---|
| m1, m2a, m2b, m2c, m4, m9, m11; s1, s2, s3 (Stand: Länge, Monat 13, Tag 0), s5, s6, s25, s27; u5, w1; Z49 (4), Z26 (3); H4, H5, H8, H10 | s4, s14, t4, t11a, t11b, t21, m7, m8 (V-1); v4, x1, m5 (V-2) |

## 7. Läufe

**`make gates`** im Zweig: Exit 0 (`EXIT=0` am Ende des Protokolls, nicht über `| tail` maskiert). Darin `docs-check` 384 Dateien, 0 Befunde; `handbuch-check` grün; `groessenmarke` `ok, Messung 3864 von Marke 6900`. In diesem Lauf war die Stufe `handbuch` `CACHED`; sie ist deshalb ein zweites Mal ohne Cache gefahren (`docker build --no-cache-filter handbuch --target handbuch`): `pruefung handbuch: ok`, Exit 0.

**N5 (Verdrahtung).** Kopie des Baums mit dem Satz „Das ist geplant.“ im Handbuch: `make handbuch-check` rot, `pruefung handbuch: Zeile 11: Planung A1: geplant`, `docker build` Exit 1, `make` Exit 2. Die Verdrahtung trägt (Stufe, `.dockerignore`, Ziel); gelaufen ist die Kopie, nicht der Zweig.

**Bedingung des Plans (§5)**: „der Lauf der Gates im Hauptbaum“. Dieser Lauf war im Worktree des Zweigs; im Hauptbaum läuft `make gates` nach dem Merge (Abschnitt 9).

## 8. Ungelaufen

- `make gegenproben` (Zeitbudget und Stempel gehören dem Integrationsschritt, V-6).
- Zeile für Zeile jede Mutation des Implementers (Z1 bis Z30, Z39 bis Z48): nicht nachgefahren; gedeckt sind sie durch die Mutationen des Reviewers, mein Probe-Paket (Abschnitt 4) und die Sweeps (Abschnitt 6), soweit sie diese Zeilen berühren. Zeilen ohne eigene Mutation oder Probe von mir: Z1 bis Z5, Z7 bis Z9, Z12 bis Z14, Z16, Z18 bis Z20, Z22 bis Z25, Z27, Z28, Z30, Z36 bis Z37, Z39 bis Z46, Z48.
- Die Verdrahtung gegen ein falsches `Dockerfile` (N5 gilt für den Fall „Wort aus A“, nicht für „Stufe fehlt“ oder „Ziel nicht in `GATE_CHECKS`“; Letzteres steht per `make -pn` gelesen, nicht gemutet).

## 9. Closure-Vorbedingungen (nur genannt)

1. V-1 und V-2 gelöst und gegen die Klasse nachverifiziert; V-3 im Plan nachgezogen.
2. §7 des Plans: Closure-Notiz mit Lerneintrag; jedes Risiko aus §6 mit Ausgang (zwei *weiter offen* mit Adresse Integrationsschritt und User).
3. `make gates` im Hauptbaum nach dem Merge (Plan §5); `make gegenproben` bei geringer Last und Stempel allein committen (Integrationsschritt, Plan §6).
4. Diff-Vorschlag für `AGENTS.md` (Hard Rule „Ist-Zustand“): der User; bis dahin bindet die Regel durch Vertrag und Gate.
5. Integrationsschritt: Welle `welle-benutzerhandbuch` (Stufe `handbuch`, Abschnitte „Aufruf und Version“ und „Fehlerbehebung“), Roadmap-Zeile, endgültige Fallnummern in `SPEC-102`, Nachzug bei den Nehmern, `docs/plan/planning/README.md` Nr. 6.
6. Messung des Endes: `make groessenmarke` meldet 3864 von 6900; ein Eintrag `ueberschreitung` ist nicht nötig. Die Fixes zu V-1 und V-2 zählen zur Messung.

**Register-Kandidaten (nicht geschrieben; Abgleich und Zähler sind Sache des Planners):**
- `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`: drei Runden (Review, Re-Review, Verifikation) fanden an demselben Leser je neue Elemente, die die Klasse des vorigen Befundes nicht umfasste (V-1 (a) bis (g)); die Mengen der Fixes folgten dem Befund, nicht dem Vertrag insgesamt. Vorschlag zur Prüfung: ein Durchlauf einfacher Mutationen (Vergleichsoperator, entfernte Bedingung) je Zeile eines Lesers vor der Übergabe.
- `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`: V-1 (acht Mutationen an sieben Zusagen).
- `BEO-REPO/randform-erst-im-review-entschieden`: V-2 (Zeile der Befunde, Rand des Spans).
- `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle`: V-3.
- `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`: V-1 (a) (Form des Stands: nur Folgetage, kein falsches Zeichen) und (b) (Gründe nur einzeln).

## 10. Zusammenfassung

| Schwere | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 (V-1, V-2) |
| LOW | 1 (V-3) |
| INFO | 3 (V-4, V-5, V-6) |

**Übergabe an den Planner:** NICHT BESTANDEN. V-1 an den Implementer (je Element ein Fall, als Schleife über die Klasse: Stellen und Zeichen des Stands, alle Paare der Gründe, Zeilen der Überschrift, Namen ohne Option, Zeichen × Zeilen, Zeilen mit `#` vor der ersten `## `), V-2 an den Architect (Ort in `SPEC-200`), danach Implementer; V-3 an den Planner. Alles übrige hält: `make gates` Exit 0, Stufe `handbuch` ohne Cache grün, Wort aus A macht das Ziel rot, Marke 6900 fest bei Messung 3864, Z49 und Z26 in 20 von 20 Prozessen gefangen (außer dem benannten Negativ).

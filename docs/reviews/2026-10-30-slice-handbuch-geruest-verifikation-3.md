# Verifikations-Report 3: slice-handbuch-geruest — 2026-10-31

**Rolle:** Verifier, Nachverifikation (dritte Runde) in frischem Kontext nach den Fix-Runden zu [W-1 bis W-6](2026-10-30-slice-handbuch-geruest-verifikation-2.md). Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-200` (Nummer für Nummer) und `AGENTS.md` §3.9 bis §3.18. Kein Review, keine Reparatur: Code, Plan und Vertrag sind von mir unverändert; der Arbeitsbaum des Zweigs war vor und nach den Läufen ohne Änderung (`git status --short` leer).

**Gegenstand:** Zweig `slice-handbuch-geruest`, HEAD `7ed5b8b`, Basis des Slice `975228d`. Gelesen: `harness/vertraege/SPEC-200-handbuch.md` (153 Zeilen), `tools/pruefung/handbuch.go`, `handbuch_kopf.go`, `handbuch_mengen.go`, `handbuch_beispiel.go` (zusammen 1128 Zeilen Produktcode), Plan §1 bis §6, die Tabelle *Zusage · Fänger* (Z1 bis Z82, N1 bis N8), `Dockerfile`, `.dockerignore`, `harness/mk/handbuch.mk`, `tools/gegenproben/handbuch.go`, Handbuch und vier Beispiel-Dateien. Die früheren Reports habe ich gelesen, mein Urteil aber gegen den Vertrag und die Klasse (§3.17) gebildet, nicht gegen ihre Mutationslisten.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31 (Dateiname nach Vorgabe des Auftrags)

**Methode.** Scratchpad `verif-geruest-3/` (flüchtig). (1) `make gates` im Worktree, Exit-Code aus Datei; danach die Stufen, die im Lauf `CACHED` standen, einzeln ohne Cache (`--no-cache-filter` für `pruefung`, `test`, `build`, `handbuch`, `race`, `lint`); `make groessenmarke` allein. (2) Ein **Probe-Paket** von 20 Tests mit eigenem Orakel (nie der Parser selbst): Generatoren über die Klasse, Orakel aus dem Wortlaut von `SPEC-200` geschrieben, die Muster und Formen der Nr. 10 bis 12 aus dem Text des Vertrags geparst und gegen den Code gehalten. (3) **59 eigene Mutationen** (Wegwerfkopie von `tools/pruefung` je Mutation im Go-Image, `sed` auf der Kopie, `go test -count=1` über das ganze Paket, netzlos, `--memory=4g --memory-swap=4g --cpus=4`); „rot“ heißt: ein Test der Zusage scheitert, kein Bauabbruch. Alle Fänger sind deterministisch und nicht nebenläufig, Rate je Mutation: rot in 1 von 1 Prozess. Schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`, eine Sperre je Gruppe. Last: Load Average 8,5 (Beginn) bis 3,5 (Zeitmessung) auf einem von mehreren Rollen geteilten Host.

---

## 1. Verdikt

**NICHT BESTANDEN.** Es gibt kein HIGH. Alle Sensoren sind grün, das Probe-Paket findet **keine** Abweichung des Codes in den Mengen von Nr. 3 bis 13 (rund 8,9 Mio. Vergleiche, Tabelle in Abschnitt 5), und die beiden vom Implementer genannten Lücken sind gedeckt (Abschnitt 3). Das Urteil steht aus vier Gründen, die `AGENTS.md` §3.15 bis §3.17 vorgeben:

1. **Eine Abweichung des Codes vom Vertrag** (X-1): Endet das Binary vor der Zeitgrenze und hält ein Kindprozess den Strom über sie hinaus, meldet die Prüfung `Zeitgrenze überschritten` und verwirft die Ausgabe; Nr. 7 e sagt, das Warten sei weder Befund noch Eingabefehler. Reproduziert in 5 von 5 Prozessen.
2. **Zusagen ohne Fänger** (X-2, X-3, X-4): Von 49 gültigen Mutationen bleiben 8 grün; 4 davon (c4 bis c7) sind die vom Planner genannte Teil-Zusage „zweites Wort bei mehreren Leerzeichen“.
3. **Zwei Randformen ohne Ort im Vertrag, im Code entschieden** (X-5), `AGENTS.md` §3.16.

Alle Befunde sind klein und lokal; mit einem Fix (Code an einer Stelle, sechs Zeilen im Test) und einer Entscheidung des Architect ist der Slice bestanden. Die Closure-Pflichten der DoD (Abschnitt 8) sind erwartungsgemäß noch offen.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| X-1 | MEDIUM | **Zeitgrenze und Warten auf offene Ströme überlagern sich falsch** (`SPEC-200` Nr. 7 c, 7 e; Z26, Z49). `fahre` entscheidet `Zeitgrenze überschritten` mit `ctx.Err() != nil` *nach* dem Warten. Endet das Binary bei 1,4 s, hält ein Kindprozess den Strom bis 2,6 s und gilt die Grenze 2 s, ist das Ergebnis `zeit=true`, `code=0`, die gelesene Ausgabe `hallo` wird verworfen; Nr. 7 e: „Die bis dahin gelesene Ausgabe und der Ausgangswert des Binary gelten; das Warten ist weder Befund noch Eingabefehler. Die Zeit kommt zur Zeitgrenze hinzu; ein Lauf dauert damit höchstens 11 s“. Im Betrieb: Binary endet in (9 s, 10 s] und ein Kind hält den Strom. Gegenprobe im selben Test: Binary endet bei 0,5 s, Kind bis 2,6 s, Exit 3: `zeit=false`, `code=3` (richtig). Z26 und Z49 prüfen die Grenze nie zusammen mit dem Warten. | Probe `TestProbeZeitgrenzeUndKind`, rot in 5 von 5 Prozessen (`-test.count=1`), Last 3,5; `tools/pruefung/handbuch_beispiel.go` `fahre`, Zeile `if ctx.Err() != nil` | Implementer (Code und Fänger an der Grenze: Binary endet knapp vor der Grenze, Kind hält darüber; je Aufrufstelle `version` und Beispiel); Architect nur, wenn die Überlagerung als akzeptiertes Negativ gelten soll |
| X-2 | MEDIUM | **Teil-Zusage „zweites Wort“ bei mehreren Leerzeichen ohne Fänger** (`SPEC-200` Nr. 9 b „nennt als zweites Wort das Ziel“; Z29, Z47, Z53). Vier Mutationen an `makeBefunde`, die nur Spans mit zwei oder mehr Leerzeichen hinter `make` ändern (c4 zweites Wort wird leer, c5 kein Aufruf, c6 drittes Wort wird Ziel, c7 wie c4 ab drei Leerzeichen): **4 von 4 grün**. Der Code folgt der Zusage (siehe Abschnitt 5, Make). Rot werden dagegen die Mutationen, die in `make `+Zeichen scheitern (c1 bis c3), also nicht isolierend. Dazu eine Lücke im Wortlaut: „Wort“ ist nicht definiert; der Code trennt an jedem Zeichen, das `unicode.IsSpace` trifft (`strings.Fields`), also auch an Tabulator und U+00A0 **hinter** dem Ziel (`make nix<TAB>foo` gilt als Ziel `nix`); Nr. 14 d kennt U+0020 als Trenner nur am Anfang des Spans. Zwischen beiden Lesarten (nur U+0020 trennt / jedes Leerraumzeichen trennt) liegen 21 617 von 126 858 strukturierten Absätzen mit Befund. | Mutationen c4 bis c7 grün (Abschnitt 6); Probe `TestProbeMakeStrukturiert`, `TestProbeMakeGerichtet` | Implementer (Fänger: Menge Abstand {2, 3 Leerzeichen} × {bekanntes Ziel, unbekanntes Ziel, Option, Zuweisung}); Architect (Definition von „Wort“ in Nr. 9 b) |
| X-3 | MEDIUM | **Z38: „Name in Bytefolge“ ohne Fänger für Namen mit Großbuchstaben** (`SPEC-200` Nr. 2 d). Die Mutation m19 (Vergleich nach `strings.ToLower`) bleibt grün: kein Test hat zwei Namen, deren Bytefolge und Kleinschreibung verschieden ordnen (`beispiele: Z.txt keine Beispiel-Datei` gegen `beispiele: a.txt ohne Bindung`; Bytefolge: `Z` vor `a`). Gefangen werden dagegen Name absteigend, Name nach Länge (m19b rot), Name ohne `.txt`. Die Menge „Zeichen, in denen Bytefolge von Kleinschreibung und Locale abweicht“ (Großbuchstaben, Nicht-ASCII) fehlt in der Schleife. | Mutation m19 grün | Implementer |
| X-4 | LOW | **Z46/Z81: der Text `Kopf: Schlüssel <k> unbekannt` nennt den getrimmten Schlüssel (Nr. 4 a „beide Seiten getrimmt“)**; die Mutation m100 (Text mit dem rohen Teil vor dem `:`, mit Leerraum) bleibt grün, weil kein Fixture einen Schlüssel mit Leerraum vor dem Doppelpunkt hat. | Mutation m100 grün | Implementer |
| X-5 | MEDIUM | **Zwei Randformen von Nr. 5 ohne Ort, im Code entschieden** (`AGENTS.md` §3.16: „eine im Code entschiedene Zelle ist ein MEDIUM“). (a) Ein Zaun mit nicht leerer Info-Zeichenkette und gültiger Bindung gilt als **gebunden** (Zitat, A und B gelten dort nicht); Nr. 5 e nennt Info nicht, aber auch nicht das Gegenteil; Mutation m83 (`gebunden = info == ""`) grün. (b) Eine Zeile `Datei: [` **im Inneren** eines geschlossenen Zaunes ist nicht `Bindungszeile ohne Zaun`; Nr. 5 c sagt wörtlich „eine Zeile, die mit `Datei: [` beginnt und auf die kein öffnender Zaun folgt“ und nimmt Zaun-Inneres nicht aus; Mutation m107 (ohne `d.im[i] == 0`) grün. | Mutationen m83, m107 grün | Architect (Ort in Nr. 5 b/5 c/5 e), danach Implementer (Fänger) |
| X-6 | INFO | **Grenze eines Treffers bei veränderlicher Länge (Muster C5 `lh-[a-z]{2}(-[a-z]+)?-\d+`).** Nr. 13 b sagt „ein Treffer endet vor einem Zeichen, das kein Buchstabe ist“. Der Code nimmt den ersten (längsten) Treffer des Ausdrucks und prüft dessen Rand: `lh-fa-01x` ist kein Treffer. Eine wörtliche Lesart („irgendein Teil, den der Ausdruck trifft“) fände `lh-fa-0` vor der Ziffer `1`. Im Differential über alle 44 Muster weicht nur C5 ab (Beispiele `lh-qa-sup-002bald`, `lh-fa-01seit version`); der Code-Treffer ist nie ohne Orakel-Treffer. | Probe `TestProbeTrefferDifferenz`: 2 640 000 Vergleiche, Code-nur-wahr 0, Orakel-nur-wahr nur C5 | Architect (Lesart festschreiben; bisher ungefangen und unentschieden) |
| X-7 | INFO | **Handbuch:** zwischen dem letzten Absatz der Einleitung und `## Bauen` fehlt die Leerzeile (`docs/user/benutzerhandbuch.md`, Zeilen 11 und 12). Markdown liest die Überschrift trotzdem; das Gate meldet nichts, weil Nr. 14 b/K6 den Aufbau der Abnahme überlässt. | `cat -A` der Zeilen 8 bis 12 | Planner (Abnahme, Autor) |
| X-8 | INFO | **Äquivalente Mutanten**, nicht als Lücke gezählt: m87 (`l.zeit && l.code == 0`; bei `zeit` ist `code` immer 0, weil `fahre` vor dem Lesen des Exit-Werts zurückkehrt) und m123 (`sort.Slice` statt `sort.SliceStable`; bei 16 Gleichständen der Menge A in umgekehrter Reihenfolge im Text bleibt die Tabellenreihenfolge, `TestProbeGleichstandGross`). Beim zweiten steht die Stabilität nur auf der Implementierung von `sort`, nicht auf einem Test; der Vertrag („Reihenfolge der Prüfung“) verlangt sie. | Mutationen m87, m123 grün | — |

Einordnung. X-1 ist die einzige Stelle, an der Code und Vertrag auseinandergehen, und eine Klasse, die kein Test streift (Grenze × Warten). X-2 und X-3 sind Zusagen ohne Fänger (§3.15: MEDIUM); bei X-2 fügt der Wortlaut eine Lücke hinzu. X-5 ist die Zellenregel des Rasters (§3.16). X-4 ist der Text eines Befundes mit getrimmtem Wert. Kein Befund betrifft die Richtigkeit einer Meldung im Betrieb des Handbuchs, das heute im Zweig liegt.

## 3. Die zwei vom Implementer genannten Lücken, aufgelöst

**(a) Z54: `version: Ausgabe nicht lesbar` zwischen die `Kopf: … fehlt`.** Menge: die Stellen, an denen die Befunde von `version` in die Folge der drei `fehlt`-Befunde (Software-Version, Gültigkeitsbereich, Stand, alle bei Zeile 0) eingefügt werden können: vor allen, nach dem ersten, nach dem zweiten (der Fall „nach allen“ ist der Code) = 3 Elemente, im Code umgesetzt als `mutZ54(…, pos)` mit pos 0, 1, 2 an der Stelle, an der `pruefeDokument` die Befunde sammelt.

| Mutation | roter Test |
|---|---|
| a1 `nicht lesbar` nach `Software-Version fehlt` | `TestHandbuchKopfGleichstandZeileNull` |
| a2 `nicht lesbar` nach `Gültigkeitsbereich fehlt` | `TestHandbuchKopfGleichstandZeileNull` |
| a3 `nicht lesbar` vor allen `fehlt` | `TestHandbuchVersionBefundZeile`, `TestHandbuchKopfGleichstandZeileNull` |

Gelöst: 3 von 3 rot, im Grund (Reihenfolge), nicht durch Bruch des Aufbaus. Dazu das Orakel-Differential des Kopfes (Abschnitt 5, 300 000 Zeilenfolgen, Reihenfolge der `fehlt`-Befunde und Zeile 0 gegen ein von mir geschriebenes Orakel) ohne Abweichung.

**(b) Z64: `harness/mk/` mit Unterverzeichnis `sub/x.mk`.** Menge: die Quellen des Aufrufpräfix und der Zielliste nach Nr. 8 und 9 a (`Makefile`, `harness/mk/*.mk` regulär, Unterverzeichnis, `d-check.mk`, `repo.mk`, `Dockerfile`); die Unterscheidung „nicht rekursiv“ gilt an **zwei** Stellen (Image in Nr. 8, Zielliste in Nr. 9 a), beide aus derselben Funktion `makeDateien(wurzel, allZiele)`.

| Mutation | roter Test |
|---|---|
| b2 `harness/mk` rekursiv gelesen, nur für das Image (`allZiele == false`) | `TestHandbuchAufrufpraefixQuellen` (Element `harness/mk/sub/x.mk`) |
| b3 `harness/mk` rekursiv gelesen, nur für die Zielliste | `TestHandbuchZielliste`, `TestHandbuchZiellisteMkDateien` |

Gelöst: 2 von 2 rot. (Die Mutation b1, beide zugleich, brach den Bau und zählt nicht.) Zusätzlich fährt `TestProbeZielliste` ein Verzeichnis `harness/mk/dir.mk/inner.mk`, einen Symbolischen Verweis `link.mk`, `ok.mk.bak`, `mk` und `x.mkx` mit: kein Ziel daraus in der Liste, wie Nr. 9 a es sagt (Symbolischer Verweis: der Vertrag schweigt, der Code ignoriert ihn).

## 4. SPEC-200 Nummer für Nummer

| Nr. | Zusage | Beleg (eigener Lauf) |
|---|---|---|
| 1 | Regel, sechs Klassen K1 bis K6 | Handbuch gegen die Regel gelesen: Kopf, drei Kommandos in gebundenen Beispielen, `make` nur in Inline-Code (`make build`, `make build VERSION=1.2.3`, `make`); kein Wort der Mengen A, B, C (Stufe `handbuch` ohne Cache: `pruefung handbuch: ok`) |
| 2 a bis e | Aufruf, Ausgabe, Eingabefehler, Reihenfolge, Wirkung | Z1/Z48/Z4/Z65 vorhanden; Verdrahtung N5 gefahren (Abschnitt 7): Wort aus A im Handbuch macht die Stufe `handbuch` rot, Exit 1, `Zeile 83: Planung A1: geplant`; Reihenfolge gegen Orakel (Kopf, Version, Zäune) ohne Abweichung; Mutationen m19b, m1 rot, m19 grün (X-3) |
| 3 a bis c | Lesen: UTF-8/NUL/BOM, Zeilen, Zeichen | `zeilenVon` gegen Orakel für **alle** 29 524 Zeichenfolgen bis Länge 9 über {`a`, `\r`, `\n`} gleich; Zeichenklassen für alle 1 112 064 Codepunkte (ohne Surrogate) gegen `unicode.Categories`: 139 774 gemeldet, je Zeile einmal, Format `U+%04X`; Mutationen m86 (kleine Hexziffern), m102 (BOM irgendwo), z1 (Zeichen nach Codepunkt statt erster Stelle) rot |
| 4 a bis e | Kopf, Schlüssel, Software-Version, Stand, Gültigkeitsbereich | Kopf-Differential 300 000 Zeilenfolgen und Version-Differential 576 Kombinationen gleich dem Orakel; Stand: 2 677 198 Zeichenketten (Jahre 0 bis 9999 × Monate 0 bis 13 × Tage 0 bis 32, dazu Zufallsketten mit U+0663, `+`, Leerzeichen) gegen `time.Parse` plus ASCII-Ziffernprüfung gleich; Mutationen m61a bis m61e (jedes der fünf Monatslängen-Elemente), m70, m71, m9, m72, m73, m80 bis m82, m8c, m90 rot |
| 5 a bis f | Zäune, Bindung, Block gleich Datei, Verwaiste | 200 000 Dokumente (`zerlege`, `absaetze`) und 200 000 für Zaun/Info/Bindung/Datei/Bindungszeile ohne Zaun gleich dem Orakel; Einträge von `beispiele/`: 20 Einträge (17 Namen, Verzeichnis, zwei Symbolische Verweise), 3 Bindungsmengen gleich; Bindungsform: 402 234 Zeichenketten gegen Orakel gleich; Mutationen m1, m3, m51, m31 rot, m83 und m107 grün (X-5) |
| 6 a bis d | Beispiel-Datei | 400 000 erzeugte Dateien (1 799 gültig) gegen Orakel, Gültigkeit und Werte gleich |
| 7 a bis e | Lauf, Umgebung, Zeitgrenze, Vergleich, offene Ströme | Z22 bis Z26, Z49 vorhanden; X-1 (Grenze × Warten) |
| 8 | Aufrufpräfix gedeckt | b2, m88 rot; Unterverzeichnis und Quellen (3 (b)) |
| 9 a, b | Zielliste, Inline-Code | Zielliste-Fixture mit 10 Zeilenformen und 11 Dateien gleich; Spans: 337 928 Absätze gleich dem Span-Orakel; `makeBefunde`: 256 070 strukturierte plus 338 286 zufällige Absätze plus 17 gerichtete Fälle gleich dem Orakel (jedes Leerraumzeichen trennt); m26, m30a, m30b, m31, m84, m85, m89 rot; c4 bis c7 grün (X-2) |
| 10 bis 12 | Mengen A, B, C | 44 Muster und 104 Formen (A 36, B 44, C 24) aus dem Text des Vertrags geparst: Ausdrücke 44 von 44 gleich dem Code; jede Form × 15 Varianten (roh, Satz, Großschreibung, Versalien, fett, kursiv, Code, Zeilenumbruch, U+00A0, Tabulator, Mehrfach-Leerzeichen, U+2003, Ziffer davor, Klammer und Punkt, Tabellenzeile) = 1 560 Elemente treffen in Prosa; im Zitat: 360 (C) treffen, 1 200 (A, B) melden nichts |
| 13 a bis d | Abgleich, Rand, Zitat, Meldung | Differential 60 000 Texte × 44 Muster = 2 640 000 Vergleiche: Code-nur-wahr 0 (X-6); m17, m18 (Zitat), m101 (Treffer in Versalien) rot |
| 14 | Negative | N1 bis N8 benennen Grund und Folge; Ergänzung: X-5, X-6 |
| 15 a bis d | Stufe, Wurzel, Verdrahtung | `Dockerfile` Stufe `handbuch` aus `build`, sechs Einträge in `.dockerignore`, `harness/mk/handbuch.mk` (`GATE_CHECKS += handbuch-check`), Zeile in `harness/README.md` §Sensors; `make -n gates` nennt `--target handbuch`; `docs-check` 0 Befunde |
| Gegenprobe | H1 bis H10 | `tools/gegenproben/handbuch.go` gelesen: die zehn Fälle entsprechen den zehn Zeilen des Vertrags (H9 ohne „Ersatz nicht gestartet“, wie Plan F-10 es streicht); **nicht gefahren** (Auftrag) |

## 5. Mengen und Probe-Paket (Elementzahlen)

Das Probe-Paket liegt flüchtig im Scratchpad (`probe/probe1_test.go` bis `probe6_test.go`); jedes Orakel ist unabhängig vom Parser (Laufliste, `time.Parse`, Tabelle `unicode.Categories`, handgeschriebene Zerlegung der Datei).

| Menge (aus dem Wortlaut) | Elemente | Orakel | Ergebnis |
|---|---|---|---|
| Muster und Formen von Nr. 10 bis 12 | 44 Muster, 104 Formen | Text von `SPEC-200`, geparst | Code gleich (44/44) |
| Formen × Stellungen | 104 × 15 = 1 560 (Prosa), 360 + 1 200 (Zitat) | Vertrag | alle grün |
| Treffer-Rand (Nr. 13 b) | 60 000 Texte × 44 Muster = 2 640 000 | Teilstück-Prüfung mit `^(?:muster)$` und Randzeichen | Code-nur-wahr 0; Orakel-nur-wahr nur C5 (X-6) |
| Zäune, Absätze (Nr. 5 a, 13 a) | 200 000 Dokumente aus 20 Zeilenarten | Zustandsautomat | grün |
| Zaun + Info + Bindung + Datei (Nr. 5 b bis e) | 200 000 Dokumente, 3 vorhandene und 1 fehlende Datei | Wortlaut | grün |
| Spans (Nr. 9 b) | 337 928 Absätze, 12 267 mit Span | Laufliste | grün |
| `make`-Befunde | 338 286 zufällige (14 mit Befund) + 256 070 strukturierte (126 858 mit Befund) + 17 gerichtete | Orakel mit `FieldsFunc(IsSpace)`; zweites Orakel „nur U+0020 trennt“ | gleich dem ersten; 21 617 strukturierte weichen vom zweiten ab (X-2) |
| Stand (Nr. 4 d) | 2 677 198 | `time.Parse` + ASCII-Ziffern | grün |
| Kopf (Nr. 4 a, b, d) | 300 000 Zeilenfolgen aus 33 Zeilenarten | Wortlaut | grün |
| Version (Nr. 4 c) | 576 (16 Ausgaben × 2 Fehlerausgaben × 3 Exit × 2 Signal × 3 Kopfwerte) | Wortlaut | grün |
| Zeilen (Nr. 3 b) | 29 524 Zeichenfolgen bis Länge 9 | Zustandsautomat | grün |
| Zeichen (Nr. 3 c) | 1 112 064 Codepunkte | `unicode.Categories` | grün |
| Bindungsform, Name (Nr. 5 b, 5 f) | 402 234 | Teilstück-Zerlegung | grün |
| Beispiel-Datei (Nr. 6) | 400 000 (1 799 gültig) | Wortlaut | grün |
| Verwaiste (Nr. 5 f) | 20 Einträge × 3 Bindungsmengen | Wortlaut | grün |
| Zielliste (Nr. 9 a) | 10 Zeilenformen im Makefile, 11 Dateien | Wortlaut | grün |

Grenze des Pakets: Die Differentiale für `make` sind bei zufälliger Erzeugung dünn (14 Befunde); die strukturierte Erzeugung (126 858 Befunde) schließt die Lücke. Die Läufe von `fahre` (Zeit, Umgebung, Ströme) prüfen die vorhandenen Tests; ich habe nur die Grenze × Warten selbst gemessen (X-1).

## 6. Eigene Mutationen (59)

Gruppen 1 bis 4, je eine Kopie, `go test -count=1 .` über das ganze Paket (26 s je Lauf in der Stufe `pruefung`).

| Ausgang | Zahl | Mutationen |
|---|---|---|
| gültig | 49 | |
| rot, im Grund | 41 | a1 a2 a3 b2 b3 c1 c2 c3 m1 m3 m8c m9 m17 m18 m26 m30a m30b m31 m51 m61a m61b m61c m61d m61e m70 m71 m72 m73 m80 m81 m82 m84 m85 m86 m88 m89 m90 m19b m101 m102 z1 |
| grün (Lücke) | 8 | c4 c5 c6 c7 (X-2), m19 (X-3), m100 (X-4), m83 und m107 (X-5) |
| ausgeschieden | 10 | Bau gebrochen: b1, m8, m91, m8b; ohne Wirkung: m92 (sed fand nichts), m112, m113, m114; äquivalent: m87, m123 (X-8) |

Mutation an einem Element, das keiner der Vorgänger im Plan nennt (Spalte *Mutation* der Tabelle): m61a bis m61e (je ein Monat), m80 bis m82 (Paare der Gründe), m84 und m85 (Option oder Zuweisung irgendwo im Span), m88 (Image nur im Makefile), m89 (Ziel mit Strich), m90 (Überschrift mit Leerraum), m101 (Treffer in Versalien), z1 (Reihenfolge der Zeichen), m72 und m73 (Trimmen nur von Leerzeichen). Rot bei den meisten dieser Elemente; die grünen stehen oben.

## 7. Läufe

- **`make gates`** im Worktree: `EXIT=0` (Datei `gates.exit`, nicht über `| tail` maskiert); darin `a-check: gesamt 0 Befund(e)`, `baseline-verify v6.18.1 OK — 54 Dateien`, `d-check: 386 Datei(en) geprüft, 0 Befund(e)`, `groessenmarke ok`, `migration-check: ok`.
- **Ohne Cache einzeln** (`make build test test-race handbuch-check lint` mit `--no-cache-filter pruefung`, `test`, `build`, `handbuch`, `race`, `lint`): `EXIT=0`; die Stufe `pruefung` lief neu (`go test ./...` ok, 26,3 s), dazu `pruefung build: ok`, `pruefung test: ok` (70 s), `pruefung race: ok` (106 s), `pruefung handbuch: ok` (0,15 s), `golangci-lint 0 issues`.
- **`make groessenmarke`** allein: `EXIT=0`; `Hinweis: Tabelle ergibt 8610, Schätzung 5085`, `ok, Messung 5503 von Marke 6900` (80 %), Messpunkt `zeilen=90 nebenlaeufig=2 fakes=2 produktcode=1000 schaetzung=8610 marke=6900 messung=5503`. Gegengerechnet: 85·2 + 80·88 + 200·2 + 1000 = 8610 (`SPEC-128` Nr. 1); `git diff --numstat 975228d..HEAD` über die sechs Pfade des Blocks gibt 5 476 Einfügungen und 27 Löschungen, Summe 5 503; die Marke 6900 = ⌈135·5085/10000⌉ auf volle Hundert gleich dem Block.
- **Tabelle gegen den Plan:** 90 Zeilen (Z1 bis Z82 lückenlos, N1 bis N8), die Summe der Mutationen in Z54 bis Z82 ist 85 (4+2+9+6+4+4+3+2+5+3+4+3+2+3+2+1+6+1+6+2+1+2+1+1+1+2+2+2+1), wie der Plan sagt; jede Zahl in §4 „Stand der Messung“ stimmt mit dem Gate überein. Die Zahl 19 von Z64 hält `TestHandbuchAufrufpraefixQuellen` (`n != 19`).
- **Verdrahtung N5, gefahren:** Kopie des Baums ohne `.git` im Scratchpad, ein Satz „Das ist geplant.“ am Ende des Handbuchs, `docker build --no-cache-filter handbuch --target handbuch`: Exit 1, `pruefung handbuch: Zeile 83: Planung A1: geplant` (`SPEC-200` Nr. 14 h).
- **Rohbytes (§3.18):** Alle 44 geänderten oder neuen Dateien des Zweigs gegen `LC_ALL=C grep -P` auf BOM, U+00A0 (`\xC2\xA0`), `\r`, die Folge Backslash-u-Hex, U+200B bis U+200F, U+2028 bis U+202E und U+00AD: kein Treffer; Handbuch und vier Beispiel-Dateien enthalten kein `\r`. Dieser Report ist nach dem Schreiben mit denselben Mustern gegengelesen (Datei `docs/reviews/2026-10-30-slice-handbuch-geruest-verifikation-3.md`).
- **Nicht gefahren, wie beauftragt:** `make gegenproben` und der Stempel.

## 8. DoD: was bleibt offen (nur genannt)

- Closure-Notiz mit Lerneintrag in §7 des Plans; alle Häkchen in §2 sind noch leer; das Beobachtungs-Register (Kandidaten unten); die drei Paarungen (Anker, Folge-Slice, Register).
- Diff-Vorschlag für `AGENTS.md` (Hard Rule „Ist-Zustand“): beim User, §6 führt ihn mit Ausgang *offen*.
- Integrationsschritt: `make gates` im Hauptbaum nach dem Merge, `make gegenproben` bei geringer Last samt Stempel, endgültige Fallnummern 121 bis 130 in `SPEC-102`, Welle, Roadmap-Zeile, Nachzug bei den Nehmern.
- Aus dieser Runde: X-1 bis X-5 (Fix und Fänger, Ort im Vertrag) sind noch in §6 des Plans als Zeilen mit Ausgang *offen* und Adresse einzutragen (`AGENTS.md` §3.13), sobald der Planner sie annimmt; ein Bericht allein ist keine Adresse.

## 9. Ungeprüft

- `make gegenproben` und der Stempel (Auftrag); die Fälle H1 bis H10 sind gelesen, nicht gefahren.
- Eine Nicht-Datei als Handbuch (`docs/user` ist eine Datei, ENOTDIR): der Code liefert wohl `Handbuch nicht lesbar` statt `Handbuch fehlt`; nicht gefahren, der Vertrag nennt den Fall nicht.
- Signale an `pruefung handbuch` und das Zurückbleiben des Arbeitsverzeichnisses (Nr. 14 g, benanntes Negativ).
- Mutationen jenseits der 59; Z1 bis Z53 des Implementers nicht Zeile für Zeile nachgefahren (gedeckt durch die Vorverifikationen und die Mutationen und Differentiale dieser Runde, soweit sie die Zeilen berühren).
- Die Verständlichkeit des Handbuchs (K6): ein Mensch liest; hier nur die Form (X-7).

## 10. Register-Kandidaten (nur genannt, nicht geschrieben; Abgleich und Zähler sind Sache des Planner)

- `BEO-REPO/zwei-zeitregeln-ueberlagern-sich-an-der-grenze`: eine Frist (Zeitgrenze) und eine Zusatzfrist (Warten) werden nacheinander ausgewertet, die Entscheidung liest die Uhr nach der zweiten (X-1); die Fänger prüfen jede Frist für sich.
- `BEO-REPO/teil-zusage-ohne-isolierende-mutation` (X-2): vierte Runde, in der eine Teil-Zusage nur durch Mutationen gefangen scheint, die an anderen Fällen scheitern.
- `BEO-REPO/ordnung-prueft-nur-die-schreibweise-der-fixture` (X-3): Fixtures mit Kleinbuchstaben ordnen jedes Vergleichsmittel gleich.
- `BEO-REPO/randform-erst-im-review-entschieden` (X-5): wieder eine Zelle (Lage im Zaun, Info und Bindung) im Code entschieden.
- `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (X-2, X-4): „Wort“ und „getrimmt“ im Text eines Befundes.

## 11. Zusammenfassung

| Schwere | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 (X-1, X-2, X-3, X-5) |
| LOW | 1 (X-4) |
| INFO | 3 (X-6, X-7, X-8) |

**Übergabe an den Planner:** NICHT BESTANDEN. X-1 an den Implementer (Code in `fahre`: das Ende des Binary entscheidet die Zeitgrenze, nicht die Uhr nach dem Warten; Fänger an der Grenze je Aufrufstelle); X-2, X-3, X-4 an den Implementer (Fänger als Zeilen in den bestehenden Schleifen: Abstand 2 und 3 hinter `make`; Namen mit Großbuchstaben und Nicht-ASCII; Schlüssel mit Leerraum vor dem Doppelpunkt); X-5, X-6 und die Definition von „Wort“ in X-2 an den Architect (Orte in Nr. 5 und 9 b, Lesart des Treffer-Rands in Nr. 13 b). Nach Fix und Entscheidung genügt eine kurze Nachverifikation gegen genau diese Mengen; die Differentiale dieser Runde liegen im Scratchpad und fahren in wenigen Sekunden.

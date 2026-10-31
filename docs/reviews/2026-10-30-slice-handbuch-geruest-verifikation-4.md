# Verifikations-Report 4: slice-handbuch-geruest — 2026-10-31

**Rolle:** Verifier, Nachverifikation (vierte Runde) in frischem Kontext nach den Fixes zu [X-1 bis X-7](2026-10-30-slice-handbuch-geruest-verifikation-3.md). Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-200` (Nummer für Nummer) und `AGENTS.md` §3.9 bis §3.18. Kein Review, keine Reparatur: Code, Plan und Vertrag sind von mir unverändert; der Arbeitsbaum des Zweigs war vor und nach den Läufen sauber (`git status --short` leer bis auf diesen Report).

**Bezug:** [LH-RB-01](../../spec/lastenheft.md), [LH-QA-05](../../spec/lastenheft.md), [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Ort der Verträge), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Größenmarke).

**Gegenstand:** Zweig `slice-handbuch-geruest`, HEAD `21954a5`, Basis des Slice `975228d`. Gelesen: `harness/vertraege/SPEC-200-handbuch.md` (153 Zeilen), `tools/pruefung/handbuch.go`, `handbuch_kopf.go`, `handbuch_mengen.go`, `handbuch_beispiel.go`, Plan (DoD, §1, §3 Tabelle Z1 bis Z89 und N1 bis N10, §4, §6), `Dockerfile`, `.dockerignore`, `harness/mk/handbuch.mk`, `harness/README.md` §Sensors, Handbuch und vier Beispiel-Dateien, der Diff `dfe72bd` (Vertrag) und `74cd084` (Fänger Z83 bis Z89). Der Report der dritten Runde ist gelesen; seine Mutationslisten habe ich nicht übernommen.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-31 (Dateiname nach Vorgabe des Auftrags) · **Last:** Load Average beim Start 8,55 / 9,90 / 10,28 (20 Kerne), während der Läufe 5,4 bis 15.

**Methode.** Scratchpad `verif-geruest-4/` (flüchtig). (1) Sensoren selbst gefahren (Abschnitt 6). (2) Ein **Probe-Paket** von 19 Testfunktionen (235 Läufe und Unterfälle, auf dem unmutierten Code alle grün) in einer Kopie der Quellen; jedes Orakel ist von Hand oder aus dem Text des Vertrags geschrieben, nie der Parser: die Tabelle der Mengen aus `SPEC-200` geparst, Normalisierung und Treffer-Rand als zweite Implementierung, erwartete Befundlisten von Hand. (3) **86 Mutationen** (85 verschiedene) in Kopien, je gegen die bestehende Suite (`go test -count=1 -failfast`) und gegen das Probe-Paket. (4) Z83 als Wirkungs-Fänger mit Rate gemessen.

---

## 1. Verdikt

**NICHT BESTANDEN.** Es gibt kein HIGH. **Der Code weicht an keiner Stelle vom Vertrag ab**, die Fixes zu X-1 bis X-7 sind vollständig (Abschnitt 3), alle Sensoren sind grün (Abschnitt 6), und Z83 ist bei der vorgefundenen Last stabil. Das Urteil beruht auf zwei **Zusagen ohne Fänger** (`AGENTS.md` §3.15: MEDIUM), die im Bestand keine Mutation rot macht, und auf einer Randform ohne Ort (LOW). Beide MEDIUM sind je eine Ergänzung einer bestehenden Schleife; der Code ist in beiden Fällen richtig.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| Y-1 | MEDIUM | **Z21: eine leere Ausgangswert-Zeile hat keinen Fänger** (`SPEC-200` Nr. 6 c „Eine Zeile mit einer Dezimalzahl von 0 bis 255“). Die Mutation „Prüfung auf die leere Zeichenkette in `ausgangswert` (`handbuch_beispiel.go` Zeile 218) entfernt“ macht eine Datei mit `--- Ausgangswert`, einer Leerzeile und sonst gültigem Aufbau gültig (Wert 0) und startet sie: **Bestand grün** (`TestHandbuchBeispielAusgangswert` führt 3 gültige und 10 ungültige Formen; das `""` darin ist eine *fehlende* Zeile, nicht eine leere). Mein Probe-Paket (Menge: 12 gültige und 28 ungültige Zeilen, darunter die Leerzeile; dazu 3 Dateiformen ohne oder mit nur leeren Zeilen) wird rot. Die Zahl-Überläufe (m78) fängt der Bestand. | Mutationen m31 = m73 | Implementer |
| Y-2 | MEDIUM | **Nr. 2 d „dann nach der Reihenfolge der Prüfung“ hat keinen Fänger für Ausgaben, die umgeordnet werden müssen** (Z38, Z54, Z55, Z73). Die Mutation `sort.SliceStable` → `sort.Slice` in `meldeBefunde` bleibt im Bestand **grün** (kein Test wird rot; Beobachtung). Die Mutation ändert nur die Reihenfolge von Gleichständen (gleiche Nummer, Zeile, Name), und nur in Ausgaben, deren Erzeugungsfolge von der Ausgabefolge abweicht: mehrere Absätze, deren Treffer je Absatz in A, B, C erzeugt und nach Nummer neu geordnet werden. Ein Absatz mit allen 44 Mustern (Erzeugungsfolge schon in Ausgabereihenfolge, `TestProbeTiesGemischt`) bleibt unter der Mutation grün; mein Zufallspaket (400 Dokumente) und `TestProbeTiesAbsaetze` (2, 3 und 6 Absätze mit je einer Form aller 44 Muster) werden rot, auf dem unmutierten Code grün. Vermutung zum Grund: Die Sortierung der Standardbibliothek ordnet kleine oder vorsortierte Eingaben stabil. **Dies widerlegt die Einstufung „äquivalent“ (m123, X-8) der Vorrunde:** deren Fänger (ein Absatz, 16 Treffer) liegt schon sortiert vor. | Mutation m01; Probe `TestProbeZufall`, `TestProbeTiesAbsaetze` | Implementer |
| Y-3 | LOW | **Pfadkomponente ist eine Datei (ENOTDIR) beim Handbuch: der Vertrag nennt den Fall nicht.** Ist `docs/user` (oder *Wurzel*) eine Datei, meldet der Code `Handbuch nicht lesbar: open …: not a directory` (Exit 2); Nr. 3 a sagt „Nur ein fehlender Eintrag ist `Handbuch fehlt`“ und entscheidet den Symbolischen Verweis ins Leere als `fehlt`. Beim Binary ist die Parallele durch den Code entschieden (`os.Stat` scheitert: `Binary fehlt`), beim Handbuch nicht; kein Fänger, und ich finde in §6 des Plans keine Zelle des Rasters dafür (`AGENTS.md` §3.16: im Code entschiedene Zelle). Der Unterschied ist nur der Text, beide Exit 2. | Probe `TestProbeEingabe/docs-user-datei` | Architect (Ort im Vertrag), dann Implementer (Fänger) |
| Y-4 | INFO | **Eine Bindungszeile ohne Zaun nennt die Datei nicht** (Nr. 5 c, 5 f, 7 a). `Datei: [beispiele/a.txt](beispiele/a.txt)` allein (kein Zaun danach) gibt `Bindungszeile ohne Zaun` und zusätzlich `beispiele: a.txt ohne Bindung`; die Datei wird nicht gefahren. Nr. 5 f spricht von „einer gültigen Bindungszeile“, Nr. 5 c nennt die Zeile ohne Zaun auch Bindungszeile. Beide Lesarten sind vertretbar; der Vertrag entscheidet nicht. | Probe `datei-ohne-zaun` (beobachtet) | Architect |
| Y-5 | INFO | **„Nicht leere Zeilen“ in Nr. 13 a ist nicht erklärt.** Nr. 3 b erklärt „leer“ für 4 a, 4 b, 5 a. Der Code behandelt eine Zeile aus Leerraum (auch NBSP) als leer, sie trennt Absätze. Eine solche Zeile zwischen `noch` und `nicht` verhindert damit den Treffer A11. Folge im Betrieb gering. | Code `absaetze` (gelesen, nicht gefahren) | Architect (Wortlaut) |
| Y-6 | INFO | **Nr. 5 b „der Befund der Info-Zeichenkette ist dann der einzige des Blocks“ stimmt nur für die Zaun-Prüfung.** Trägt der gebundene Inhalt eine Kennung (Nr. 13 c, Menge C) oder weicht das Binary ab (Nr. 7), stehen weitere Befunde am Block; Z87 prüft genau das (`spec-110` im Inhalt), der Code ist richtig. | Wortlaut | Architect |
| Y-7 | INFO | **Nr. 14 g nennt den Rest eines Signals nur als Verzeichnis.** Nach SIGTERM, SIGHUP und SIGKILL bleibt das Arbeitsverzeichnis (Exit 143, 129, 137, wie Nr. 14 g), und der Prozess des Binary läuft als Waise weiter (`sleep 20` blieb stehen). Im verworfenen Container ohne Folge. | Lauf `sig.sh` | Architect (Satz in 14 g) |
| Y-8 | INFO | **Stand `2026-10-30`** im Kopf, das Gate lief am 2026-10-31. Nr. 14 e schließt den Vergleich mit der Uhr aus; die Abnahme setzt den Tag. | Handbuch, Kopf | Planner (Abnahme) |
| Y-9 | INFO | **Äquivalente oder wirkungslose Mutanten**, nicht als Lücke gezählt: m57 (Text vor dem Treffer getrimmt: gleicher Rand), m64 und m65 (`gekillt` auch bei fehlgeschlagenem `Kill` bzw. Zusatzbedingung `err != nil`: Go ruft `cmd.Cancel` nach dem Ende des Prozesses vermutlich nicht mehr; die Zellen B2/K3 sind dafür der Beleg, beobachtet, nicht bewiesen), m68 (`<=` im Zweig `zeile != zeile`, nur bei Ungleichheit erreicht), m75 (`len(z) > 4` statt `> 3`: die Länge ist durch führende Null und Wert schon gedeckt, m78 `> 30` ist rot), m77 (`info: ""+`), m90 (`TrimSpace` am Text eines Eingabefehlers: nur ein Pfad mit Leerraum am Ende fiele auf). | Abschnitt 5 | — |

Einordnung. Y-1 und Y-2 sind Klasse-Lücken im Sinn von `AGENTS.md` §3.17: Die Menge der Zusage (jede Zeile außer einer Dezimalzahl 0 bis 255; jede Gleichstandsgruppe in einer Ausgabe, die nach der Erzeugung umgeordnet wird) ist größer als die Elemente, die der Bestand führt. Beide Fänger sind kleine Ergänzungen: für Y-1 die Leerzeile in die Liste der ungültigen Zeilen von Z21 (die Datei wird nicht gefahren, 0 Starts), für Y-2 ein Handbuch mit mindestens zwei Absätzen, die jeweils Treffer aus A, B und C tragen, mit der erwarteten Liste von Hand (Gleichstände in Tabellenreihenfolge); die Sortierung ist deterministisch, die Rate ist je Lauf 1 von 1, kein Wirkungs-Fänger nötig.

## 3. Die Fixes zu X-1 bis X-7, gegen die Klasse aufgelöst

| Befund | Menge (Wortlaut des Vertrags) | Elemente | Mein Beleg |
|---|---|---|---|
| X-1 Zeitgrenze und Warten (Nr. 7 c, 7 e, 4 c) | Ende des Binary {vor, nach der Grenze} × Ende des Kindes {kein, vor, nach der Grenze, nach Grenze + 1 s} × Aufrufstelle {`version`, Beispiel} | 2 × 4 × 2 = 16; Z83 führt 16 Zellen (anders geschnitten) | Rate unten (40 von 40 grün). Eigene Zellen, 5 Beispiel und 3 `version`, mit `TestProbeZeit`: Beispiel schläft (Grenze 1 s): 2,00 s, Befund, keine Vergleichsbefunde; Binary endet, Kind hält 4 s (Grenze 0,5 s und 0,1 s): 1,00 s, kein Befund, Ausgabe gilt; mit Exit 7: Befund `Ausgangswert`, kein `Zeitgrenze`; `version` schläft: allein `version: Zeitgrenze überschritten`; `version` endet, Kind hält: kein Befund; `version` vor `version.txt` |
| X-2 Wort in Nr. 9 b | jede Art Leerraum × vor und hinter dem Ziel × Zeichen unmittelbar hinter `make` | 24 Leerraum-Zeichen (`unicode.IsSpace` ohne `\n`), 4 Nicht-Leerraum-Zeichen; Z84 führt alle | 50 Fälle `TestProbeMake`, darunter 13 Fälle mit 10 verschiedenen Zeichen (NBSP, Tabulator, U+2028, U+2029, U+0085, U+001C, U+180E, U+200B, U+2060, U+FEFF), Paar-Entfernen, Mehrzeilig, Listen- und Tabellenzeichen; Mutation m61 (jeder Leerraum direkt hinter `make` gilt als Aufruf) rot |
| X-3 Name als Bytefolge | Namen mit `Z`, `_`, `a`, U+00C4, `-` vor `.`, Verzeichnis, Endung | 16 Einträge in einem Lauf | `TestProbeVerwaist`: Ausgabe in der Bytefolge `-a.txt`, `.gitkeep`, `.txt`, `A`, `Z.txt`, `_.txt`, `a--b.txt`, `a-b.txt`, `a.TXT`, `a.txt`, `b.txt`, `link.txt`, `sub`, `x.md`, `z.txt`, `Ä.txt`; Mutation m67 (`<=` im Namen) rot |
| X-4 Schlüssel `k` getrimmt | Leerraum vor, hinter, im Inneren; Groß- und Kleinschreibung | 5 Fälle | `TestProbeKopf`: `Schlüssel software-version unbekannt`, `Schlüssel Handbuch  Version unbekannt` (Tabulatoren außen, zwei Leerzeichen innen), NBSP vor dem Doppelpunkt |
| X-5 Info bei gültiger Bindung; `Datei: [` im Zaun | Info {`go`, getrimmt, Leerzeichen, NBSP}; Lage der Zeile {im geschlossenen Zaun, im nicht geschlossenen, vor der Öffnungszeile, nach der Schlusszeile, allein, nach Leerzeile} | 4 + 6 | `TestProbeZaun` (56 Fälle): nur der Info-Befund bei gültiger Bindung; `Datei: [` als Ausgabe eines gebundenen Beispiels ohne Befund; im nicht geschlossenen Zaun Befund, unmittelbar davor keiner; nach der Schlusszeile Befund. Mutationen m50, m51 (Teilfolge statt Präfix), m70, m87 rot |
| X-6 Treffer-Rand 13 b | Anfang × Ende × Form, Ziffern und `-` als Rand, Mehrbyte-Buchstabe | 104 Formen × 25 Varianten = 2 600 Absätze | `TestProbeSweep`: 2 054 Befunde gleich dem Orakel; Mutationen m62 (Ziffern als Buchstaben), m63 (`-` als Buchstabe) rot |
| X-7 Leerzeile im Handbuch | Zeile vor `## Bauen` | 1 | Leerzeile vorhanden; Stufe `handbuch`: `pruefung handbuch: ok` |

Die Nachbarnummern, die dieselbe Aussage tragen, sind im Vertrag konsistent (Nr. 3 b mit 9 b: „Leerraum“ gleich `unicode.IsSpace`; Nr. 3 c mit 13 a: Zs wird zum Leerzeichen, Cc nicht und ist ohnehin Befund; Nr. 7 c, 7 e und 14 e: je Lauf 10 s plus höchstens 1 s, der Ausgang an der Grenze ist offen; Nr. 14 d mit 9 b: nur das Zeichen hinter `make`; Nr. 4 c mit 7 c: `version: Zeitgrenze überschritten` allein; Nr. 5 b mit 5 e und 13 c: Info gehört nicht zur Bindung). Zwei Sätze tragen die Aussage nicht ganz: Y-5 (Nr. 13 a) und Y-6 (Nr. 5 b).

## 4. SPEC-200 Nummer für Nummer

| Nr. | Ergebnis |
|---|---|
| 1 | Handbuch gegen die Regel gelesen: Kopf, drei Kommandos in vier gebundenen Beispielen, `make` nur in Inline-Code, kein Wort der Mengen; Stufe `handbuch` ohne Cache: `ok` |
| 2 a bis e | Aufruf (Z1); beide Eingabefehler in der Reihenfolge der Prüfung; `Arbeitsverzeichnis nicht anlegbar` für das eigene Verzeichnis (TMPDIR fehlt) und für das zweite (der Fake löscht das Arbeitsverzeichnis während `version`: genau eine Zeile `Eingabe:`); kein Rest in TMPDIR nach Exit 0, 1 und 2; Umgebung leer (`/proc/$$/environ` 0 Bytes), cwd leer und je Lauf ein neues Verzeichnis, Standardeingabe leer. Reihenfolge der Nummern 3, 4, 5, 7 bis 12 in einem Handbuch mit je einem Befund gleich der Vertragsfolge (Nr. 6 nicht erzeugt). Relative Pfade für Binary und Wurzel laufen |
| 3 a bis c | 5 Gründe und Kombinationen (NUL und BOM ⇒ `NUL`, UTF-8 und NUL und BOM ⇒ `kein UTF-8`), BOM mitten im Text ⇒ Befund U+FEFF; ENOTDIR ⇒ Y-3. Zeichenmenge: je Zeile einmal, Reihenfolge der ersten Stelle bei 40 Zeichen in einer Zeile |
| 4 a bis e | 62 Fälle `TestProbeKopf` (Überschrift fehlt in 6 Formen, Schlüssel fehlt, doppelt, leer in allen Paaren, `Stand` 7 gültige und 25 ungültige, darunter 1900, 2100, 2000, 2028, 2026-02-29, Tag 0, Monat 0 und 13, fremde Ziffern), 25 Fälle `TestProbeVersion` (alle Gründe in Reihenfolge, Kürzung auf 200 Zeichen, Zeile 0 gegen Zeile des Schlüssels bei fehlendem, doppeltem, leerem Wert) |
| 5 a bis f | `TestProbeZaun` (56 Fälle: 13 + 9 ungültige Bindungsformen, 5 gültige Namen, Zaunformen, Info), `TestProbeVerwaist` (16 Einträge), Y-4 |
| 6 a bis d | 254 Argumentzeichen (jedes Byte von 0x01 bis 0xFF außer `\n`: gefahren genau für die 66 erlaubten), 40 Ausgangswerte, 3 Dateiformen, 2 Dateiformen mit `\r` und mit zweiter Leerzeile (nicht gefahren); Y-1 |
| 7 a bis e | Argumente unverändert und in der Reihenfolge; je Datei einmal bei zwei Blöcken; ungültige und nicht gebundene Dateien nicht gefahren (Starts: `version` und eine Datei); Vergleich je Strom einzeln, vertauschte Ströme ergeben beide Befunde; Kürzung auf 200 Zeichen in Zeichen, nicht Bytes (250 mal `ä`), Schnitt in einer Escape-Folge; `Signal`; X-1 |
| 8 | 8 Fälle: Image nur aus `Makefile` oder `harness/mk/*.mk`, nicht aus `d-check.mk`, `repo.mk`, Unterverzeichnis; Pfad nur aus dem `Dockerfile`; kein Eintrag, kein Befund; ein Unterverzeichnis zählt als Eintrag |
| 9 a, b | 50 Fälle `TestProbeMake`; Zielliste mit `x:=1 ## c`, `ok-1:##`, `crlf`, ohne `Bad`, `a_b`, ` a1`, `nochmal :##`, `.PHONY`, Unterverzeichnis |
| 10 bis 12 | 44 Muster und 104 Formen (A 36, B 44, C 24) aus dem Text geparst: Ausdrücke 44 von 44 gleich dem Code; `TestProbeSweep` und `TestProbeZufall` (Abschnitt 5) |
| 13 a bis d | Normalisierung (Zs, Zl, Zp, Tabulator, `*`, `_`, Backtick) und Rand im Sweep und im Zufall; Zitat: C gilt, A und B nicht; Zeile des Zitats ist die des Zaunes |
| 14 | a bis h gelesen; e (Zeit) und g (Signal) gefahren (Y-7, Y-8); h: Verdrahtung gefahren (Abschnitt 6) |
| 15 a bis d | `Dockerfile` Stufe `handbuch` aus `build` mit `RUN --network=none`, sechs Einträge in `.dockerignore`, `harness/mk/handbuch.mk` hängt an `GATE_CHECKS` (`make -n gates` nennt `--target handbuch`), Zeile in `harness/README.md` §Sensors |
| Gegenprobe | H1 bis H10 in `tools/gegenproben/handbuch.go` und `SPEC-102` Fälle 121 bis 130 vorhanden; **nicht gefahren** (Auftrag) |

## 5. Mengen, Probe-Paket, Mutationen

| Menge (aus dem Wortlaut) | Elemente | Orakel | Ergebnis |
|---|---|---|---|
| Muster, Formen (Nr. 10 bis 12) | 44, 104 | Text des Vertrags, geparst | Code gleich (44 von 44) |
| Formen × Stellungen und Normalisierung (Nr. 13) | 104 × 25 = 2 600 Absätze, 2 054 Befunde | zweite Implementierung (Treffer je Position) | 0 Abweichungen |
| Zufällige Handbücher mit Zitaten (Nr. 5, 10 bis 13) | 400 Dokumente, je 12 Blöcke | wie oben, Zeilen und Reihenfolge | 0 Abweichungen |
| Argumentzeichen (Nr. 6 a) | 254 | Wortlaut | gleich |
| Ausgangswert (Nr. 6 c) | 40 Zeilen, 3 Dateiformen | Wortlaut | gleich (Y-1 betrifft den Bestand, nicht den Code) |
| Bindungsform (Nr. 5 b) | 22 ungültige, 5 gültige | Wortlaut | gleich |
| Stand (Nr. 4 d) | 32 | Kalender von Hand | gleich |
| Leerraum und Nicht-Leerraum im Span (Nr. 9 b) | 14 Fälle | Wortlaut | gleich |
| Verwaiste Einträge (Nr. 5 f) | 16 | Bytefolge | gleich |

Grenze des Pakets: Die Zeichenmenge von Nr. 3 c und der Kalender des Stands sind in den Vorrunden über alle Codepunkte und Jahre 0 bis 9999 gegen Tabellen gefahren; ich habe sie nicht wiederholt, sondern die Fälle des Wortlauts (Kategorien, Jahrhundertregel) gezielt geprüft. Die Differentiale für Kopf, Zäune und Beispiel-Dateien der Vorrunde (200 000 bis 400 000 Erzeugungen) wurden nicht erneut gefahren.

**Mutationen: 86 Läufe, 85 verschiedene** (m31 und m73 sind dieselbe). Vor der Zählung ausgeschieden: sechs Kompilierfehler (m03, m06, m15, m27, m35, m55 im ersten Lauf, danach in geänderter Form gefahren), ein Treffer des Mutators fehlte (m12, danach gefahren), m86 traf keine Zeile.

| Ausgang | Läufe | Mutationen |
|---|---|---|
| rot im Bestand | 76 | m02 bis m30, m32 bis m56, m58 bis m63, m66, m67, m69, m70, m74, m76, m78, m80 bis m85, m87 bis m89 |
| grün, Lücke | 3 | m01 (Y-2), m31 und m73 (Y-1) |
| grün, äquivalent oder wirkungslos | 7 | m57 m64 m65 m68 m75 m77 m90 (Y-9) |

Mutationen an einem Ort, den im Plan keine Zeile der Spalte *Mutation* nennt (Auswahl): m01 (Sortierung), m31/m73 (Leerzeile), m36 (`WaitDelay` 2 s: rot in `B2/K3/Beispiel`, 3,71 s), m45 (Leerzeichen-Paar am Span), m57 bis m65 (Rand, Cancel), m78 bis m85 (Überlauf, Vergleich des Codes, Zitat), m87 bis m90 (Datei-Zeile mit Leerraum, Bindungszeile rechts getrimmt, Zeile des `version`-Befundes). „Rot“ heißt hier: der erste rote Test des Bestands ist ein Test der Zusage der Mutation (Name im Lauf-Log), nicht ein Bruch des Aufbaus; bei m36 war es die Zelle der Dauer.

## 6. Läufe

- **`make gates`** im Worktree: `EXIT=0` (Datei, nicht über `| tail` gelesen); darin `a-check`, `baseline-verify v6.18.1 OK — 54 Dateien`, `d-check: 387 Datei(en) geprüft, 0 Befund(e)`, `groessenmarke ok`, `migration-check: ok`. `make docs-check` danach mit diesem Report: `388 Datei(en) geprüft, 0 Befund(e)`.
- **Ohne Cache einzeln** (`--no-cache-filter` je Stufe): `make test`, `make build`, `make test-race`, `make lint`, `make handbuch-check` je `EXIT=0`; dazu die Stufe `pruefung` (`go test ./...` in `tools/pruefung`, 29,6 s) `ok`. In den Läufen der Ziele stand die Stufe `pruefung` weiter `CACHED`; deshalb der eigene Lauf.
- **`make groessenmarke`** (im Lauf von `make gates`): `ok, Messung 6060 von Marke 6900`. Nachgerechnet: `git diff --numstat 975228d..HEAD` über die sechs Pfade des Blocks = 6060; Tabelle in §3 = 99 Zeilen (96 andere, 3 nebenläufig: Z26, Z49, Z83); 85·3 + 80·96 + 200·2 + 1000 = 9335 gegen die Schätzung 5085 (Hinweis, kein Befund); Plan §4 nennt dieselben Zahlen.
- **Z83 als Wirkungs-Fänger:** `TestHandbuchZeitgrenzeUndWarten` (16 Zellen parallel, Dauer mit Schranke +0,6 s), `-test.count=1` je Prozess: **20 von 20 grün** nacheinander bei Load Average 6,9 / 9,0 / 10,1 (während des Laufs 5,5 bis 8,3), dazu **20 von 20 grün** mit 4 gleichzeitigen Prozessen (je 16 parallele Zellen, Container auf 4 CPUs begrenzt) bei Load 9,4 / 7,1 / 6,5. Mutation m36 rot; m64 und m65 nicht unterscheidbar (Y-9). Keine Flakiness beobachtet.
- **Verdrahtung (Nr. 14 h), gefahren:** `git archive HEAD` in den Scratchpad, ein Satz „Das ist geplant.“ am Ende des Handbuchs, `docker build --target handbuch`: Exit 1, `pruefung handbuch: Zeile 84: Planung A1: geplant`.
- **Signale (Nr. 14 g):** SIGTERM Exit 143, SIGHUP 129, SIGKILL 137; in allen drei bleibt ein Verzeichnis `pruefung-handbuch-…` in TMPDIR (Nr. 14 g), der Prozess des Binary läuft weiter (Y-7). SIGINT erreichte das Programm nicht (Hintergrundjob der Shell ignoriert ihn), der Lauf endete an der Zeitgrenze: kein Beleg.
- **Rohbytes (§3.18):** Alle 84 geänderten oder neuen Dateien des Zweigs außer `docs/reviews/` (Punktdateien eingeschlossen) gegen `LC_ALL=C grep -P` auf BOM, U+00A0 (`\xC2\xA0`), `\r`, die Folge Backslash-u-Hex, Steuerzeichen, U+200B bis U+200F, U+2028 bis U+202E, U+2060 bis U+2064 und U+00AD: kein Treffer. Dieser Report ist nach dem Schreiben auf Rohbytes gegengelesen (Ergebnis in der Übergabe).
- **Nicht gefahren, wie beauftragt:** `make gegenproben` und der Stempel.

## 7. DoD: Stand

| DoD-Zeile | Stand |
|---|---|
| Handbuch mit Kopf und Kapiteln, Beispiele als Dateien, Lauf in Stufe `handbuch` | erfüllt (4 Beispiel-Dateien, Stufe `handbuch` ok) |
| Anpassung des Standards; Diff-Vorschlag für `AGENTS.md` beim User | Anpassung erfüllt (Abschnitt „Anpassung für dieses Repo“, Entscheidungen zu den Abschnitten 1 bis 14); der Diff-Vorschlag liegt als Scratch-Datei außerhalb des Repos, **nicht prüfbar** (§6 des Plans führt ihn mit Ausgang *offen*, Adresse User) |
| Gate `make handbuch-check`: Programm, Stufe, `GATE_CHECKS`, `.dockerignore`, Zeile in §Sensors, H1 bis H10; je Zeile eine Mutation rot, je Muster eine | Verdrahtung erfüllt; Gegenproben vorhanden, nicht gefahren; Mutationen laut Plan; **Y-1 und Y-2** zeigen zwei Zusagen (Z21; Z38 und Nachbarn), deren Menge der Fänger nicht deckt |
| `make gates` grün; `make docs-check` meldet das Handbuch ohne Befund | erfüllt (Exit 0, `d-check` 0 Befunde) |
| Review durchgeführt, Report unter `docs/reviews/` | erfüllt (zwei Reviews, drei Verifikationen davor, dieser Report) |
| Doku-Update (`SPEC-200` Stand, `SPEC-102`, `harness/README.md`) | erfüllt (Zeile *Stand* nennt die gelieferten Dateien; Fälle 121 bis 130) |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen, Sache des Planner nach dem Fix dieser Runde |

## 8. Ungeprüft

- `make gegenproben`, der Stempel; die Fälle H1 bis H10 sind gelesen, nicht gefahren.
- SIGINT (siehe Abschnitt 6), das Verhalten bei vollem TMPDIR und bei fehlendem Recht zum Anlegen als Nicht-root (der Container läuft als root).
- Der Wortlaut von `Handbuch nicht lesbar: <Grund>` des Betriebssystems (gehört nicht zum Vertrag).
- Die vollen Differentiale der Vorrunden (Zeichen über alle Codepunkte, Stand über alle Jahre, Kopf und Zäune mit 200 000 Erzeugungen) wurden nicht wiederholt.
- Y-5 (leere Zeile aus NBSP) und das Verhalten von `os.Stat` bei ENOTDIR am Binary sind aus dem Code gelesen, nicht gefahren.
- Verständlichkeit und Vollständigkeit des Handbuchs (K6, Nr. 14 a, b): ein Mensch liest.

## 9. Register-Kandidaten (nur genannt, nicht geschrieben; Abgleich und Zähler sind Sache des Planner)

- `BEO-REPO/aequivalent-eingestuft-ohne-umzuordnende-eingabe`: Y-2; die Vorrunde stufte `sort.Slice` als äquivalent ein, weil ihr Fänger schon sortierte Eingaben fuhr; die Klasse „Reihenfolge-Zusage“ braucht eine Eingabe, die umgeordnet werden muss.
- `BEO-REPO/leere-zeile-ist-nicht-fehlende-zeile`: Y-1; „leer“ in der Liste der ungültigen Formen meinte die fehlende Zeile; die leere Zeile ist eine eigene Form.
- `BEO-REPO/randform-erst-im-review-entschieden`: Y-3, Y-4, Y-5 (Pfadkomponente, Bindungszeile allein, leere Zeile): wieder Zellen, die der Code entschied.

## 10. Zusammenfassung

| Schwere | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 (Y-1, Y-2) |
| LOW | 1 (Y-3) |
| INFO | 6 (Y-4 bis Y-9) |

**Übergabe an den Planner:** NICHT BESTANDEN. Y-1 und Y-2 an den Implementer (je eine Ergänzung einer bestehenden Schleife: die Leerzeile bei Z21; zwei oder mehr Absätze mit Treffern aus A, B und C und eine von Hand geschriebene Erwartung bei Z38/Z73; der Bericht nennt je Befund die Menge, die Zahl ihrer Elemente und eine Mutation, rot gesehen, an einem Element, das dieser Report nicht nannte). Y-3 bis Y-6 an den Architect (je ein Satz im Vertrag: ENOTDIR beim Handbuch wie beim Binary oder als eigener Fall; die alleinstehende Bindungszeile nennt die Datei oder nicht; „leer“ in Nr. 13 a; „einziger Befund des Zauns“ in Nr. 5 b). Nach diesen Änderungen genügt eine Teilprüfung der beiden Fänger; Code und Vertrag sind in dieser Runde nicht zu ändern.

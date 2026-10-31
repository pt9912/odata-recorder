# Review-Report: slice-handbuch-geruest, Re-Review der Fix-Runde — 2026-10-30

**Review-Art:** Code — Re-Review gegen Plan (§3 Z1 bis Z50, N1 bis N6, Nachtrag 2, §6), den Vertrag `SPEC-200`, `SPEC-102`, `SPEC-128` und die Hard Rules (`AGENTS.md` §3.7 bis §3.18). Nicht geprüft wird gegen die DoD (Aufgabe des Verifiers). Der erste Report: `2026-10-30-slice-handbuch-geruest-review.md` (F-1 bis F-15).

**Gegenstand:** `git diff 804ae07..ed58f0c`: `ad8b5f0` (`SPEC-200` nach den Entscheidungen des Architect), `0d90db1` (Code, Tests in `handbuch_review_test.go`, Handbuch, Plan, `SPEC-036`, `SPEC-102`, Kommentare), `ed58f0c` (Planner: Block und Marke). HEAD beim Review: `ed58f0c`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Methode.** Der Bericht des Implementers (Plan §3, Nachtrag 2) wurde gelesen; keine Mutation und keine Zahl ist übernommen. Werkzeuge im Scratchpad `rev2-geruest/` (nichts im Repo außer diesem Report): eine Wegwerfkopie von `tools/pruefung` je Mutation (`mk.sh`, `ed.sh`: literales Ersetzen auf der Kopie), ein Läufer im Image `odata-recorder:test` (`batch.sh`: netzlos, `--memory=4g --memory-swap=4g --cpus=4`, `go test -count=1` über das ganze Paket, unter `flock /tmp/odata-recorder-heavy.lock`; ein Baufehler zählt nicht als rot, vier solche Mutationen sind mit Hilfszeilen neu gebaut), ein Probe-Paket `prb2_test.go` (Saat 100001 und 100002) mit Orakeln aus dem Wortlaut von `SPEC-200`. Der unveränderte Stand ist im Läufer grün (`base rc=0`).

**Läufe:** `make gates` Exit 0 im Zweig (`gates.rc`: `rc=0`; darin `docs-check` „383 Datei(en) geprüft, 0 Befund(e)“, `test`, `race`, `lint`, `handbuch-check` (Stufe `handbuch`, `pruefung handbuch /out/odata-recorder /handbuch` grün), `migration-check`, `groessenmarke`: `ok, Messung 3586 von Marke 8000`, Messpunkt `zeilen=56 nebenlaeufig=2 fakes=2 produktcode=1000 schaetzung=5890 marke=8000 messung=3586`). `git status` nach dem Lauf ohne Änderung.

**Ungelaufen:** (1) die Rate zu Z49 mit 20 Prozessen und `-test.count=1` je Mutation (der Läufer `rate.sh` steht seit über 15 Minuten in der Schlange der Sperre; siehe *Raten*); (2) `make gegenproben` (Zeitbudget und Stempel gehören dem Integrationsschritt, kein Finding); (3) Mutationen an den C4-Mustern, an F-8 a bis c und am Zweig `ErrWaitDelay` mit eigenem Muster: für C4 steht die Probe (600 Texte), für F-8 a bis c das Lesen der Tests und die Probe, keine Mutation.

---

## Antworten auf die Schwerpunkte

**Mutationsstand.** 46 gültige Mutationen am Programm (vier weitere waren Baufehler und sind mit Hilfszeilen neu gebaut, alle vier rot; sie sind mitgezählt). Rot: 25. Grün: 21, davon 3 ohne Zusagebruch (b8 Anker `^` bei `HasPrefix`, äquivalent; e3 Kleinschreibung bei Namen aus Kleinbuchstaben, äquivalent; d3 Obergrenze von drei Blöcken je Datei, willkürlich, INFO) und 18 grün ohne Äquivalenz; die 18 stehen in R-1 bis R-5.

**F-10.** Die Streichung der Zusage „das Ersatz-Skript wurde nicht gestartet“ ist legitim, keine stille Lockerung: Der Fall `h9-binary-fehlt` übergibt `bin/gibt-es-nicht`, das Ersatz-Skript steht nicht auf der Kommandozeile und kann nicht gestartet werden; die Zusage hatte nie einen Fänger. „Ein Eingabefehler startet das Binary nicht“ fängt Z4 (0 Starts bei Handbuch fehlt, UTF-8, NUL, BOM, Verzeichnis, Arbeitsverzeichnis und falscher Aufruf); `SPEC-200` Gegenprobe H9, `SPEC-102` Fall 129, Plan Z40 und der Kommentar an `hbBauen` (`starts` negativ: keine Prüfung) sagen dasselbe. Die Mutation „Prüfung Binary fehlt entfällt“ bleibt über den Text rot (erster Report).

**Marke (`ed58f0c`, 6900 auf 8000, `ersetzt: 6900`).** Die Rechnung stimmt (S = 85·2 + 80·54 + 400 + 1000 = 5890, 135·5890/10000 aufgerundet = 8000; `make groessenmarke` bestätigt). Die Anhebung ist trotzdem nicht gedeckt: `SPEC-128` Nr. 5 c sagt ausdrücklich, nach dem ersten Code-Commit sei die Marke fest, die Differenz zur Tabelle sei dort ein Hinweis; `AGENTS.md` §3.14 lässt das Ersetzen einer älteren Marke zu, wenn der Planner sie vor dem ersten Code-Commit setzt; `SPEC-128` Nr. 10 e nennt „Marke und Schätzung beide erhöht, beide in der Formel stimmig“ nach dem ersten Code-Commit ausdrücklich als vom Programm ungesehen, es falle dem Reviewer im Diff auf. Es gab keinen Anlass: die Messung 3586 liegt unter der alten Marke 6900, eine Überschreitung mit Eintrag `ueberschreitung` stand nicht an. Der Eintrag `ersetzt` deckt die Anhebung formal (Nr. 5 h), legitimiert sie aber nicht. Die Begründung „keine Lockerung“ im Plan trifft nicht: die Auslastung fällt von 52 % auf 44,8 %, ohne dass die Messung es verlangte. R-6.

---

## Findings

Beleg nennt die Mutation (Name im Scratchpad `m-<Name>`) oder die Probe. Kein Finding ist HIGH.

| ID | Kategorie | Befund | Quelle | Pfad | Beleg | Zuständig | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|---|---|
| R-1 | MEDIUM | Die Menge „getrimmt“ (`SPEC-200` Nr. 3 b, jedes Zeichen, das `unicode.IsSpace` trifft) ist an den sechs Stellen mit vier Zeichen gefangen (NBSP, U+2003, Tabulator, U+3000), die Leerraum-Menge hat 19 (`TestHandbuchLeerraumMenge` führt alle 19 nur am Abgleich der Mengen). Ein Trimmen ohne Zl und Zp (U+2028, U+2029; nach Nr. 3 c zulässig) bleibt an jeder der sechs Stellen grün. | `AGENTS.md` §3.17 (Fänger als Schleife über alle Elemente der Menge), §3.15 | `tools/pruefung/handbuch_review_test.go` · `for _, c := range []string{nbsp, "\xe2\x80\x83", "\t", "\xe3\x80\x80"} {` | t1 (leere Zeile des Kopfes), t2 (Schlüssel), t3 (Wert), t4 (Info), t5 (Schluss-Zaun), t6 (Absatzgrenze): je Trimmen mit Leerzeichen, Tabulator und `unicode.Zs`, ohne Zl und Zp: alle sechs rc=0 (grün); t7 (nur fünf Zeichen) grün. Die Stellen selbst sind gedeckt | Implementer | ja — `make test` | Fänger führt nur einen Teil der Menge |
| R-2 | MEDIUM | „Ein Span, der bis auf Leerraum genau `make` ist“ (`SPEC-200` Nr. 9 b) hat einen Fänger mit drei ASCII-Spans (`make`, `make `, `make  `); Leerraum ist die Menge aus Nr. 3 b. Ein Span `make` plus NBSP, Tabulator oder eines der anderen 19 Zeichen ist ungefangen, ebenso ein Span, der mit `make` ohne Leerzeichen beginnt (`makefile x`). Das Programm verhält sich richtig (Probe: alle 19 Zeichen kein Befund), die Lücke liegt im Fänger. | `AGENTS.md` §3.17, §3.15 | `tools/pruefung/handbuch_review_test.go` · `for _, prosa := range []string{"Siehe `make` hier.\n", "Siehe `make ` hier.\n", "Siehe `make  ` hier.\n"} {` | f1 (Zerlegung nur an U+0020), f2 (Split am Leerzeichen), f3 (Präfix `make` statt `make `) grün; Probe `TestPrb2MakeSpanLeerraum` (19 Zeichen) auf dem Programm grün | Implementer | ja — `make test` | Fänger führt nur eine Form der Eingabe |
| R-3 | MEDIUM | Z49: Das Soll im Plan („hält den Lauf höchstens eine Sekunde auf“, Kommentar des Tests: „höchstens WaitDelay (1 s)“) ist ein Wert, der in keinem Vertrag steht (`SPEC-200` Nr. 7 nennt nur die Zeitgrenze von 10 s); der Fänger prüft „Dauer beider Aufrufstellen unter 5 s“. Ein `WaitDelay` von 2 s oder 2,4 s (zwei Aufrufstellen: 4 bis 4,8 s) bleibt grün, erst 3 s ist rot. Das Verhalten bei einem Kindprozess mit offenem Strom ist damit eine im Code entschiedene Randform (Raster ohne Zelle), und der Kommentar sagt mehr zu, als der Test fängt. | `AGENTS.md` §3.11, §3.12, §3.16 | `tools/pruefung/handbuch_review_test.go` · `if dauer := time.Since(t0); dauer > 5*time.Second {`; `tools/pruefung/handbuch_beispiel.go` · `cmd.WaitDelay = time.Second` | g1 (2 s) und g2 (2,4 s) grün; g3 (3 s, Dauer 6,02 s), g4 (entfällt, 12,01 s), g5 (Fall `ErrWaitDelay` entfällt, 1,01 s), g6 (10 s, 12,03 s) rot; Rate mit 20 Prozessen: ungelaufen (siehe *Raten*) | Architect (Wert und Randform in `SPEC-200` oder gestrichen), danach Implementer | ja — `make test` | Randform im Code entschieden; Kommentar sagt mehr zu als der Test |
| R-4 | MEDIUM | Die Reihenfolge (`SPEC-200` Nr. 2 d) ist durch F-5 nur gegen das Umkehren von Zeile und Name gefangen. Ungefangen und im Vertrag ohne Ort: (a) wo ein Befund ohne Zeile steht (der Code gibt `Kopf: <Schlüssel> fehlt` die Zeile 0, also vor jedem Befund mit Zeile); (b) der Name wird vor der Zeile verglichen; (c) der Name wird ohne `.txt` verglichen. | `AGENTS.md` §3.15, §3.16 (eine im Code entschiedene Zelle des Rasters ist MEDIUM) | `tools/pruefung/handbuch_kopf.go` · `kopfBefund(0, fmt.Sprintf("Kopf: %s fehlt", s))`; `tools/pruefung/handbuch.go` · `return a.name < b.name` | e4 (Zeile 2^30 statt 0), e2 (Name vor Zeile), e1 (Name ohne `.txt`) alle grün; Probe `TestPrb2OhneZeileReihenfolge`: Ausgabe `Kopf: Stand fehlt` vor `Zeile 5: Kopf: Zeile ohne Schlüssel` | Architect (Ort der Zelle), danach Implementer | ja — `make test` | Randform im Code entschieden; Zusage ohne Fänger |
| R-5 | MEDIUM | F-1 b ist nur zum Teil gedeckt: Die Bindungszeile `Datei: [P](P)` (`SPEC-200` Nr. 5 b) und der Name einer Beispiel-Datei (Nr. 5 f, `beispielName`) haben Teile ohne Fänger: (a) das Zeichen vor `txt` (`.` statt `\.`: `a_txt`, `atxt` gelten als Name); (b) der Anker `$` von `regexpName` (ein Eintrag `a.txtx` im Verzeichnis wird als „ohne Bindung“ gemeldet statt als „keine Beispiel-Datei“); (c) das Verzeichnis der Bindung (`beispiel/` statt `beispiele/`). Die vom Befund genannten Mutationen (Anker `$` der Bindungszeile, Endung `.t[a-z]t`) sind rot. | `AGENTS.md` §3.17 (Menge ist die Form, nicht die genannten Mutationen) | `tools/pruefung/handbuch.go` · `const beispielName = `; `tools/pruefung/handbuch_beispiel.go` · `return regexp.MustCompile(`^` + beispielName + `$`)`; `tools/pruefung/handbuch_beispiel.go` · `func regexpBindung() *regexp.Regexp {` | b2 (`.txt` statt `\.txt`), b9 (`$` von `regexpName` entfällt), b5 (`beispiele?/`) grün; b1 (Anker der Bindung), b3, b4, b6, b7 rot | Implementer | ja — `make test` | Fänger führt nur eine Form der Eingabe |
| R-6 | MEDIUM | Die Marke wurde nach dem ersten Code-Commit von 6900 auf 8000 angehoben (`ed58f0c`), obwohl die Messung 3586 unter 6900 liegt. `SPEC-128` Nr. 5 c: nach dem ersten Code-Commit ist die Marke fest, die Differenz zur Tabelle ein Hinweis; Nr. 10 e: Anhebung beider Werte wird vom Programm nicht gesehen und fällt dem Reviewer auf. `AGENTS.md` §3.14 kennt das Ersetzen einer Marke vor dem ersten Code-Commit. Der Eintrag `ersetzt` und die Aussage „keine Lockerung“ im Plan decken die Anhebung nicht. Failure-Szenario: Die Marke folgt jedem Zuwachs der Tabelle durch Review-Befunde und hört auf, ein Stopp zu sein. | `AGENTS.md` §3.14, `SPEC-128` Nr. 5 c, 10 e | `docs/plan/planning/done/slice-handbuch-geruest.md` · `marke: 8000`; `ersetzt: 6900 \| Die Tabelle wuchs von 46 auf 56 Zeilen` | `git show ed58f0c`; `make groessenmarke` in `make gates`: `ok, Messung 3586 von Marke 8000` | Planner (Rücknahme auf 6900 mit Hinweis, oder Entscheidung des Architect, dass die Regel „fest“ den Zuwachs der Tabelle durch Befunde nicht meint, mit Änderung des Vertrags) | ja — Lesen des Diffs | Marke nach dem Code ohne Überschreitung angehoben |
| R-7 | LOW | F-7 ist formal behoben, die Ersatzzeile ist selbst eine Verhaltenszusage in Prosa: „Die Meldung nennt das erste Argument des Aufrufs; der Ausgangswert ist 2, und die Standardausgabe bleibt leer.“ `SPEC-200` Nr. 1 Abs. 2: „Ein Verhalten des Binary steht im Handbuch nur als gebundenes Beispiel; die Prosa daneben sagt es nicht zusätzlich zu.“ Die Aussage ist wahr (gegen `Run` gelesen, beide Beispiele zeigen sie), steht aber als Prosa und gilt allgemein („das erste Argument“) statt für die zwei Beispiele. | `SPEC-200` Nr. 1, 14 b | `docs/user/benutzerhandbuch.md` · `Die Meldung nennt das erste Argument des Aufrufs; der Ausgangswert ist 2, und die Standardausgabe bleibt leer.` | Lesen gegen `src/bootstrap/bootstrap.go` und `docs/user/beispiele/fehler-*.txt` | Implementer | nein | Prosa-Aussage über Verhalten ohne Beleg |
| R-8 | LOW | Der Plan folgt dem Code an zwei Stellen nicht (`AGENTS.md` §3.9, §3.7): (a) die Zeile Z50 in §3 nennt „NBSP“ und „6 Läufe“, der Test und der Nachtrag 2 haben vier Leerraum-Zeichen, 24 Elemente und 12 Mutationen; (b) der Kommentar an `TestHandbuchAusgangswertUeberlauf` sagt „2^64 rechnete zu 0“, ein Präteritum über eine Alternative statt der geltenden Zusage. | `AGENTS.md` §3.9, §3.7 | Plan §3 · `\| Z50 \|`; `tools/pruefung/handbuch_review_test.go` · `(2^64 rechnete zu 0)` | Lesen | Implementer | nein | Berichtigung erreicht nicht jede Stelle |
| R-9 | INFO | Zu F-8 a: Die Menge „Eintrag vorhanden und nicht lesbar“ (`SPEC-200` Nr. 3 a: Verzeichnis, fehlende Berechtigung, Fehler des Mediums) hat im Fänger ein Element (das Verzeichnis; Berechtigung ist als Administrator nicht prüfbar). Ein Symlink ins Leere ergibt `Handbuch fehlt`, obwohl der Eintrag vorhanden ist. Das `Dockerfile` kopiert Dateien und erzeugt keine Symlinks; die Zelle ist offen. | `SPEC-200` Nr. 3 a | `tools/pruefung/handbuch.go` · `case errors.Is(err, fs.ErrNotExist):` | Probe `TestPrb2HandbuchSonderformen`: toter Symlink: Exit 2 `Handbuch fehlt: …`; `docs/user` als Datei: `Handbuch nicht lesbar` | Architect | ja — Probe | – |

---

## F-1 bis F-15: gelöst gegen die Klasse oder offen

| Finding | Stand | Beleg |
|---|---|---|
| F-1 a (Ausgangswert über dem Bereich) | gelöst | a1 (`len(z) > 20`), a2, a3, a4 rot; Probe: 1140 Ziffernfolgen mit Überlauf, Längen von 4 bis 60 Ziffern (Saat 100001), gelten nicht als Ausgangswert |
| F-1 b (Bindungszeile, Endung) | offen | R-5: b2, b5, b9 grün; Rest hinter der Klammer (b1, b7) rot |
| F-2 (Form der Ausgabe von `version`) | gelöst | c1 (inneres Leerzeichen), c2 (nur U+0020), c3 (`!ok2`), c4 (`v == ""`), c5, c6, c7, c8 rot; Probe `TestPrb2VersionInnen` grün |
| F-3 (Blöcke je Datei) | gelöst | d1 (nur der erste Block verglichen) rot; d3 (ab dem vierten Block nicht verglichen) grün, willkürliche Obergrenze, INFO |
| F-4 („getrimmt“ an sechs Stellen) | offen | R-1: die sechs Stellen sind gefangen, die Menge der Zeichen nicht (t1 bis t6 grün) |
| F-5 (Reihenfolge) | offen | R-4: e1, e2, e4 grün; Platzierung ohne Zeile ohne Vertragsort |
| F-6 (Wort ADR, C4 `adrs?`) | gelöst | Vertrag, Programm und Orakel des Tests stimmen (14 Muster, 24 Formen, 104; 168 Elemente); Probe `TestPrb2C4Orakel`: 600 Texte (Saat 100002) mit Zeichen-Entfernung, Zs, Kennung und Fremdwörtern (`adresse`, `madrid`, `adrenalin`) stimmen mit einem Orakel aus Handarbeit überein; keine eigene Mutation an C4 |
| F-7 (Prosa im Handbuch) | offen | R-7: die ungenaue Zeile ist ersetzt, die neue ist eine Verhaltenszusage in Prosa |
| F-8 a (nicht lesbarer Eintrag) | gelöst | Test gelesen (Zweig `err != nil`), Probe R-9; keine eigene Mutation |
| F-8 b (Verzeichnis eines Laufs) | gelöst | beide Aufrufstellen im Test; Abbildung im Code über `errors.As`; keine eigene Mutation |
| F-8 c (leerer Schlüssel) | gelöst | Test mit vier Zeilen gelesen; keine eigene Mutation; die Stelle des Trimmens: R-1 |
| F-8 d (Span `make `) | offen | R-2 |
| F-9 (`WaitDelay`) | offen | R-3 |
| F-10 (H9) | gelöst, Streichung legitim | siehe Schwerpunkte |
| F-11 a bis c | gelöst | `SPEC-036` Stand und Grenze nennen `handbuch`; Kommentar `Fall.Getrennt` „38 bis 130“; Zeile für `harness/mk/gegenproben.mk` im Plan gelesen |
| F-11 d | gelöst für die Zahlen (56 Zeilen, 2 nebenläufig); R-8 für Z50 | – |
| F-12 (Aufruf-Text) | gelöst | h1 (Reihenfolge der Argumente), h2 (Zeile `handbuch` entfällt), h3 (Zeilen `build` entfallen) rot; `SPEC-200` Nr. 2 a nennt den gemeinsamen Text |
| F-13 (Marke) | offen | R-6: vom Planner angehoben, nicht gedeckt |
| F-14 (Messpunkt) | gelöst (INFO) | Messpunkt 3586 gegen Schätzung 5890 im Plan; das Ende nach dem Review nicht neu gemessen |
| F-15 (Version im Kopf) | gelöst | `SPEC-200` Nr. 4 c nennt `make gates VERSION=<anders>` als rot, gewollt; Z10 und H5 fangen den Vergleich |

## Raten

Z49 ist die einzige nebenläufige Zeile. Messender: der Reviewer, Last 17 bis 20 auf 20 Kernen. Gemessen ist je Mutation ein Lauf mit `go test -count=1` über das Paket (kein Ratenlauf): g1 und g2 grün; g3 (3 s) rot nach 6,02 s; g4 (`WaitDelay` entfällt) rot nach 12,01 s; g5 rot nach 1,01 s; g6 (10 s) rot nach 12,03 s. **Die Rate über 20 Prozesse mit `-test.count=1` je Mutation ist ungelaufen** (`rate.sh` für den unveränderten Stand, g1, g2 und g4 stand beim Schreiben dieses Reports in der Schlange der Sperre, ohne Ergebnis); die Raten des Implementers (1 Prozess mit `-count=20`: 20 von 20 und 0 von 20) sind nicht nachgemessen. Die Rate von g1 und g2 ist nach dem Aufbau des Tests (Dauer zweier Läufe von je etwa 2 s gegen 5 s Grenze, bei einer Mutation von 4 bis 4,8 s) deterministisch unter der Grenze, solange die Last den Prozess nicht um mehr als eine Sekunde verzögert; das ist eine Erwartung, keine Messung.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tools/pruefung/handbuch*.go` (Programm) | Programm entspricht in den geprüften Klassen dem Wortlaut (Probe P1 bis P5); Befunde liegen in den Fängern, nicht im Programm |
| `tools/pruefung/handbuch_review_test.go`, `handbuch_mengen_test.go`, `handbuch_hilfe_test.go` | geprüft; R-1 bis R-5, R-8; Rohbytes: `LC_ALL=C grep -P` auf Backslash-u-Folgen, BOM, NBSP, U+2000 bis U+200F, U+2028 bis U+202F, U+3000 in allen geänderten Dateien ohne Fundstelle (die Leerraum-Zeichen entstehen als `\x…`-Bytefolgen im Quelltext) |
| `tools/gegenproben/handbuch.go`, `faelle.go` | geprüft, ohne Befund; F-10 gelöst |
| `harness/vertraege/SPEC-200`, `SPEC-102`, `SPEC-036` | geprüft; Randformen Nr. 2 a, 2 c, 3 a, 4 a, 9 b, C4 und Nr. 14 c tragen Ort und Wortlaut; R-3, R-4 (Ort fehlt) |
| `docs/user/benutzerhandbuch.md`, `docs/user/beispiele/` | geprüft; R-7; keine Planung, Chronik oder Kennung (`make handbuch-check` grün) |
| `docs/plan/planning/done/slice-handbuch-geruest.md` | geprüft; R-6, R-8 |
| `.golangci.yml`, `.d-check.yml`, `.a-check.yml`, `docs/plan/adr/` | unverändert; keine Direktive `//nolint`, keine Lockerung eines Gates |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 6 |
| LOW | 2 |
| INFO | 1 |

**Finding-Klassen dieses Laufs:** Fänger führt nur einen Teil der Menge · Fänger führt nur eine Form der Eingabe · Randform im Code entschieden · Kommentar sagt mehr zu als der Test · Marke nach dem Code ohne Überschreitung angehoben · Prosa-Aussage über Verhalten ohne Beleg · Berichtigung erreicht nicht jede Stelle

## Verdikt

**Merge-blockierend:** ja — sechs MEDIUM. Von den fünfzehn Befunden des ersten Reports sind F-1 a, F-2, F-3, F-6, F-8 a bis c, F-10, F-11, F-12, F-14 und F-15 gegen die Klasse gelöst; offen sind F-1 b, F-4, F-5, F-7, F-8 d, F-9 und F-13. Das Programm verhält sich in allen geprüften Klassen nach dem Wortlaut; die offenen Punkte sind Fänger, die die Menge nicht ganz führen (R-1, R-2, R-5), zwei Randformen ohne Vertragsort (R-3, R-4) und die Marke (R-6). `make gates` ist grün.

**Übergabe:** Implementer (R-1, R-2, R-5, R-7, R-8, nach dem Architect R-3, R-4), Architect (Ort der Zellen in R-3, R-4; R-9), Planner (R-6). Fixes zählen erst, wenn sie die Klasse decken (`AGENTS.md` §3.17): zu R-1 alle 19 Leerraum-Zeichen an allen sechs Stellen; zu R-2 alle 19 Zeichen im Span und ein Span mit Präfix `make`; zu R-5 jedes Zeichen vor `txt`, die Anker beider Ausdrücke und die Namen des Verzeichnisses.

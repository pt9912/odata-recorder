# Verifikation: slice-telemetry-log — 2026-11-07

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-231` Nr. 1 bis 3, `SPEC-232` Nr. 2, 3, 5 c, d, f, g und die Hard Rules; nicht Diff-Review (siehe `2026-11-07-slice-telemetry-log-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-telemetry-log`, HEAD `319688d` (Implementer-Stand `7e64f20`, Review `5bd05b0`, Fix-Runde `319688d`), Basis `47b9716`. Plan: `docs/plan/planning/in-progress/slice-telemetry-log.md`.

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Nicht freigegeben ohne Auflagen: zwei MEDIUM (V-1, V-2: je eine Zusage des Vertrags ohne Fänger), ein LOW nahe MEDIUM (V-3: die Fix-Runde F-3 deckt einen von drei Wegen von `Log`), vier INFO.** Der Produktcode setzt `SPEC-231` Nr. 1 bis 3 und `SPEC-232` Nr. 3, 5 um; ich habe an ihm keine Abweichung gefunden. Von 51 eigenen Mutationen an Elementen, die weder Plan noch Review nannten, blieben acht (in drei Klassen, V-1 bis V-3) im ganzen Paket grün; zwei weitere grüne sind erklärt (äquivalente Mutation, akzeptierte Prüfgrenze). `make gates` ist grün (eigener Lauf), die Größenmarke hält (1468 von 2900), das Binary blieb unverändert, die Abdeckung unter dem Detektor ist mit 98,7 Prozent bestätigt. Die beiden MEDIUM sind Test-Lücken, keine Produktfehler: sie schließt ein Test-Commit von rund 30 bis 60 Zeilen (Fix-Runde) oder ein akzeptiertes Negativ mit Grund und Folge, das der Planner (mit dem User) setzt. Eine Wiederholungsschleife der Verifikation ist nicht vorgesehen; der Planner entscheidet.

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/vf-tlog/gates.log`: `EXIT 0`; `d-check` 760 Dateien 0 Befunde, `baseline-verify` OK, `zeichen: ok`, `groessenmarke slice-telemetry-log.md: ok, Messung 1468 von Marke 2900`. Hinweis: die Stufen `test` und `race` standen im Lauf als `CACHED` (gleiche Eingaben wie ein früherer grüner Lauf des Baums); das Paket habe ich deshalb unten selbst, ohne Cache, gefahren. |
| Test, Race, Abdeckung | eigener Container (Image `odata-recorder:race`, Baum mit `docker cp` hineingeschickt, `--network none`, `--memory=4g`, `--cpus=4`, unter `flock`): `go test -race -count=1 -covermode=atomic -coverprofile` über das Paket: `ok`, **98,7 Prozent**, 35,6 s unter Host-Last 8 von 4 CPUs (ein erster Lauf ohne Kopie traf das Image eines anderen Zweigs und zählt nicht, siehe I-4). `zeile.go`: `logFolge`, `eintrag`, `zeileSchreiben`, `stufe`, `msgText`, `anfuehren` je 100,0 Prozent; `Log`, `CountError`, `Snapshot`, `verwerfen` 100,0. Nicht 100: `spanRegel` 85,7, `wertGueltig` 92,9, `Set` und `End` 0,0 — genau die vier Blöcke, die der Plan nennt (Bestand von `slice-telemetry-port` und Gegenstand von `slice-telemetry-spans`); in den Dateien dieses Slice kein Block ohne Treffer. |
| Plan-vs-Code-Diff | `git diff --name-status 47b9716..HEAD`: neun Dateien — Plan, Review-Report, in `src/adapters/driven/telemetry` `zeile.go` (neu), `telemetry.go` (Kommentar, Feld `zeilenMu`, `Log` ruft `eintrag`), `hilfen_test.go`, `sperre_test.go`, `zeile_test.go` (neu), `regeln_test.go`, `telemetry_test.go` (feste Zahlen an die Quellen). Nichts unter `src/bootstrap`, `spec/`, `harness/`, `Makefile`, `Dockerfile`, `.golangci.yml`, `.a-check.yml`; das Binary ist unverändert. Das ist die Datei-Tabelle von §3 des Plans, nicht mehr und nicht weniger. |
| Report unverändert | `git diff 5bd05b0 319688d -- docs/reviews`: eine Zeile, der Verweis auf ADR-0001 in der Eingangs-Liste wurde ein auflösender Link (nötig für die Verweisprüfung der Doku). Findings, Mutationen und Summary sind byte-gleich. Bestätigt. |
| Rohbytes (`AGENTS.md` §3.18) | Der Implementer hat die Gegenlesung in der Fix-Runde nicht wiederholt; ich habe sie gefahren: `LC_ALL=C grep -nP` über `src/adapters/driven/telemetry/*.go`, den Plan und den Review-Report auf eine Folge aus Backslash, `u` und vier Hex-Ziffern, BOM (EF BB BF), NBSP (C2 A0), U+2028 und U+2029 (E2 80 A8, E2 80 A9) und Steuerzeichen außer Tab und Zeilenende: je **0 Treffer**. Die Nicht-ASCII-Zeichen der Dateien sind ausschließlich ü, ä, ö, ß, Ä, Ü, é, §, · und ×. `make zeichen-check` im Gate-Lauf grün. Die Fix-Runde (`zeile_test.go`) bildet jedes Element aus Zahlen (`string(rune(r))`, `string([]byte{...})`). Dieser Report wurde nach dem Schreiben mit demselben `grep -P` gegengelesen (Auflage unten, Abschnitt 7 Nr. 8). |
| Größenmarke | selbst: `git diff --numstat 47b9716..HEAD -- src/adapters/driven/telemetry` = 1417 Einfügungen + 51 Löschungen = **1468** von Marke 2900 (Auslastung 0,51, unter 0,60: der Planner meldet es bei der Closure dem Architect, `SPEC-128` Nr. 9). Messpunkt-Zeile des Gates: `Messpunkt slice-telemetry-log.md: zeilen=20 nebenlaeufig=5 fakes=1 produktcode=260 schaetzung=2085 marke=2900 messung=1468`; sie steht noch nicht im Plan (Sache des Planners, Handoff und Ende). |
| Laufzeit der Stufen | Paket ohne Detektor 4,7 s (Teilmenge `TestZeilenform`, `TestMsgLaenge`, `TestMsgZeichen`, `TestMaskierungRohbytes`, `TestLogAendertNurVerwuerfe` 0,45 s); mit Detektor `TestZeilenform` 1,15 s (Plan: ca. 1,1 s, bestätigt), `TestMaskierungRohbytes` 0,37 s; Paket unter dem Detektor 31 bis 36 s bei Last 8 von 4 (Plan: 24 bis 30 s bei Last 1,4 bis 2,8; die Abweichung folgt der Last). |
| Feste Zahlen an der Quelle | `telemetry_test.go` und `regeln_test.go` halten `len(model.AlleCodes())`, `len(support.Verwuerfe())`, `len(support.Spans())`, `len(support.Eimer())`; die Meldungen sagen „Orakel-Liste nachziehen“. Die verbleibende `41*3` in `telemetry_test.go` ist die Schleife `n` von 0 bis 40 des Tests selbst, keine Zahl einer Quelle. Die Sendung `ODR-E6008` steht in `slice-mock-lesen` §1 (`open/`, „Übernimmt (Sendung von `slice-telemetry-log`, Rückgabe 2 des Architect …)“) und in dessen Tabelle. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Zeile: Felder, Folge, Zeit, Stufen, `msg`, `duration_ms`, doppelter Schlüssel | **erfüllt im Produktcode; zwei Teilzusagen ohne Fänger (V-1, V-2)** | Code gegen `SPEC-231` Nr. 1 bis 3 gelesen: Maskierung allein `"`, Backslash, unter U+0020, U+2028, U+2029; `msg` erst Ersetzen, dann Kürzen an Zeichengrenze; `FormatFloat(…, 'f', 3, 64)`; Stufen an drei Grenzen. Haken fehlt (Planner). |
| Senke und Geheimnisse: Ausfall, ein `Write`, Sperre, Reihenfolge, keine Sperre über `Now()`, Canary | **erfüllt; Fänger der Wirkung auf den Stand lückenhaft (V-3)** | L10 bis L12, L14, L17 bis L20 mit Mutationen (Abschnitt 3); Raten bestätigt. |
| `make gates` grün | erfüllt | eigener Lauf, `EXIT 0`. Nach dem Merge von `main` einmal wiederholen (Auflage 6). |
| Review, Report liegt vor | erfüllt | `docs/reviews/2026-11-07-slice-telemetry-log-review.md`; das MEDIUM (F-1) ist durch die Fix-Runde aufgelöst (Abschnitt 4). |
| Doku-Update | entfällt | Binary unverändert; die Logfelder werden erst mit dem Einbau sichtbar (`slice-telemetry-einbau-gateway`). |
| Handbuch-Kapitel | entfällt | kein Code, Schlüssel oder Kommando im Katalog (`SPEC-232` Nr. 6); `make handbuch-check` im Gate-Lauf grün. |
| Closure-Notiz, Register, Ausgänge der Risiken, drei Paarungen | offen, Sache des Planners | siehe Auflagen. |

Liefer-Punkte: zwei (unter dem Limit von drei).

## 3. Tabelle *Zusage · Fänger* gegen den Vertrag und eigene Stichproben

Ich habe jede Aussage von `SPEC-231` Nr. 1 bis 3 und `SPEC-232` Nr. 2, 3, 5 c, d, f, g gegen die Zeilen L01 bis L20 (und die Zusatzzeilen der Fix-Runde) gelegt. Gedeckt mit Zeile: ein JSON-Objekt, ein Zeilenende, Folge, fehlende Felder, jeder Aufruf eine Zeile (L01); Zeitformat und `Now()` einmal je Zeile (L02); Stufen (L03); `MinLevel`, Gleichheit, Wert statt Name, `Now()` nicht für Einträge darunter, Verwürfe zählen (L04); Kürzen und Ersetzen (L05, L06); Maskierung nach Rohbytes (L06, Fix F-1); Zahlen (L07); doppelter Schlüssel (L08); Panik-Freiheit (L09); Ausfall der Senke in allen sieben Formen (L10); Zähler warten nicht (L11); ein `Write` je Zeile und kein Behalten des Arguments (L12); Canary (L14); Sperre, Reihenfolge, `Write` im Aufruf (L17, L18); `Now()` ohne Sperre (L19, L20); akzeptierte Negative L15, L16 mit Grund und Folge. **Nicht gedeckt** sind drei Aussagen, die die Tabelle nicht als Zeile führt: die Länge von `msg` in Bytes des Textes statt der maskierten Form (V-1), `timestamp` ist das `Now()` *dieses* Aufrufs (V-2) und die Wirkung von `Log` auf den Stand auf den Wegen Senke-Ausfall und unter `MinLevel` (V-3).

**Eigene Stichproben-Mutationen** (`tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, Speicher Vorgabe 2g, unter `flock`; Listen `liste-a.json` bis `liste-d.json` und Ausgaben `lauf-a.out` bis `lauf-d.out` unter `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/vf-tlog/`; Host-Last 5 bis 8 von 4 CPUs, Messender das Werkzeug; jede Mutation am Code, nie an der Fixture). Alle Namen tragen das Präfix `vf-`; keine ist im Plan oder im Review genannt.

### 3.1 Klasse `andere` — Maskierung (Fix F-1), Element nicht genannt

Die Menge ist „jedes Zeichen der Eingabe × maskiert oder unverändert“ (63 747 Elemente). Der Befund nannte Leerzeichen, é, `/`, `a` und alle Zeichen ab U+0080; die Fix-Runde nannte zusätzlich U+2027, U+202A, 0x7E, U+10000, U+FFFD, U+2028, U+2029 und 0x1F.

| Mutation (Name) | Ergebnis |
|---|---|
| BOM U+FEFF, NBSP U+00A0, Apostroph, U+E000 (Privatbereich), U+D7FF (Grenze vor den Ersatzzeichen), U+FFFE, Bereich U+3000 bis U+303F maskiert (`vf-mask-bom`, `-nbsp`, `-apostroph`, `-e000`, `-d7ff`, `-fffe`, `-cjk-bereich`) | 7 von 7 rot, `TestMaskierungRohbytes` |
| U+1F600 maskiert (`vf-mask-1f600-vertreter`) | rot (einer der drei Vertreter oberhalb von U+FFFF) |
| U+1F601 maskiert (`vf-mask-1f601-ausserhalb`) | **grün**, auch das Paket — akzeptierte Prüfgrenze, siehe I-1 |
| Bereich U+20000 bis U+200FF maskiert (`vf-mask-20000-ausserhalb`) | `rot-paket`: nicht der benannte Test, sondern `TestMsgZeichen` (Element `f0a08080` der Quadrupel von M3) |
| 0x0B unmaskiert, 0x1E und 0x1F unmaskiert (Grenze `0x1E`), NUL unmaskiert, nur `"` maskiert (Backslash roh), nur Backslash maskiert (`vf-mask-vt-offen`, `-1e-offen`, `-nul-offen`, `-backslash-offen`, `-quote-offen`) | 5 von 5 rot, `TestMaskierungRohbytes` |

Die Fix-Runde deckt damit die Klasse „Zeichen × maskiert/unverändert“ für alle 63 488 gültigen Zeichen bis U+FFFF und alle 256 Bytewerte. Die Zählprüfung (256, 63 488, 3 Vertreter, 70 maskierte Elemente: 34 aus den Bytewerten, 36 aus dem Bereich) habe ich gegen den Code gelesen: sie stimmt (32 Steuerzeichen + `"` + Backslash = 34; plus U+2028 und U+2029 = 36). **F-1 ist aufgelöst, bis auf die Prüfgrenze oberhalb von U+FFFF (I-1).**

### 3.2 Klasse `andere` — übrige Zusagen, Element nicht genannt

| Zusage | Mutation (Name) | Ergebnis |
|---|---|---|
| L05 Kürzen | Breite des ungültigen Bytes als 1 statt 3 gezählt (`vf-msg-n-statt-runelen`); Grenze 257 (`vf-msg-257`); Schritt 1 statt Zeichenbreite (`vf-msg-schritt-1`, im ersten Anlauf Übersetzungsfehler, berichtigt und neu gefahren); ungültige Bytes ab 0xF5 als `?` (`vf-msg-zeichen-fragezeichen-nur-ab-0xf0`) | 4 von 4 rot (`TestMsgLaenge`, `TestMsgZeichen`) |
| L05 Kürzen | ungültiges Byte wird roh kopiert statt als U+FFFD (`vf-msg-roh-kopiert`) | grün, **äquivalente Mutation**: die Längenprüfung steht vor dem Anhängen mit der Breite von U+FFFD, und `anfuehren` ersetzt das rohe Byte später in der Ausgabe; die Zeile ist byte-gleich. Kein Fänger möglich. |
| L04 `MinLevel` | Vergleich in Vierer-Schritten; `level+1`; Gleichheit bei Minimum 12 und bei −8 nicht geschrieben (`vf-min-viertel`, `-plus1`, `-12-gleich`, `-minus8-gleich`) | 4 von 4 rot, `TestMinLevel` |
| L03 Stufen | Grenze `Warn-1` (`vf-stufe-warn-minus1`) | rot |
| L10 Senke | `n == 0` ohne Fehler zählt nicht (`vf-senke-n0-nil-kein-fehler`); `n < len-1` (`vf-senke-n-eins-weniger`) | 2 von 2 rot |
| L07 Zahlen | kürzeste Darstellung (`vf-float-kuerzeste`) | rot |
| L02 Zeit | auf Mikrosekunden gekürzt (`vf-zeit-mikro`) | rot |
| L01 Folge | `ereignis`/`richtung` getauscht; `mode`/`service_id` getauscht (`vf-folge-ereignis-richtung`, `vf-folge-mode-service`) | 2 von 2 rot, `TestZeilenform` |
| L12 Argument nicht behalten | nach der Rückkehr von `Write` schreibt der Adapter in das Argument (`vf-senke-argument-behalten`) | rot, `TestEinWrite` (Kopie und Verweis des Fakes) |
| **V-1** | Länge zählt die maskierte Form: `"`, Backslash und Steuerzeichen mit 6 statt 1 Byte (`vf-msg-laenge-maskiert`, Test `TestMsgLaenge`; `vf-msg-laenge-maskiert-zeichen`, Test `TestMsgZeichen`) | **grün, das ganze Paket grün** |
| **V-1** | Länge zählt U+2028 und U+2029 mit 6 statt 3 Bytes (`vf-msg-laenge-maskiert-u2028`, `TestMsgZeichen`) | **grün, das ganze Paket grün** |
| **V-2** | `timestamp` ist das `Now()` des ersten Aufrufs, spätere Aufrufe rufen `Now()` weiter, benutzen den Wert aber nicht (`vf-zeit-gemerkt2`, `TestTimestamp`) | **grün, das ganze Paket grün** |
| **V-3** | `Log` erhöht auf dem Weg unter `MinLevel` einen Fehlerzähler (`vf-f3-below-min-fehlerzaehler-minlevel`, `TestMinLevel`) | **grün, das ganze Paket grün** (zur Zufälligkeit der ersten Fassung siehe V-3) |
| **V-3** | `Log` erhöht im Weg der defekten Senke einen Fehlerzähler (`vf-f3-senkepfad-fehlerzaehler`, `TestSenkeAusfall`; `vf-f3-senkepfad-fehlerzaehler-log`, `TestLogAendertNurVerwuerfe`) | **grün, das ganze Paket grün** (2 Mutationen) |
| **V-3** | `Log` zählt im Weg der defekten Senke zusätzlich den Verwurf `zustand` (`vf-f3-senkepfad-wert`, `TestSenkeAusfall`) | **grün, das ganze Paket grün** |
| F-3 (genannter Weg) | `Log` erhöht auf dem Weg einer geschriebenen Zeile die Verteilung eines Spans (`vf-f3-geschrieben-verteilung`) | rot, `TestLogAendertNurVerwuerfe` |

Ergebnis Klasse `andere` (Abschnitt 3.1 eingeschlossen): 44 Läufe, davon 30 rot, 3 `rot-paket` (`vf-mask-20000-ausserhalb` durch `TestMsgZeichen`; `vf-f3-below-min-fehlerzaehler` durch Zufall der Schlüsselwahl, siehe V-3; `vf-zeit-gemerkt`, eine verworfene erste Fassung, siehe V-2), 1 Übersetzungsfehler (berichtigt und neu gefahren), 10 grün — 2 davon erklärt (äquivalente Mutation; Prüfgrenze I-1) und 8 Lücken (V-1: 3, V-2: 1, V-3: 4). Dazu 8 Läufe der Klasse `nebenläufig` (Abschnitt 3.3). Die vier Listen a bis d zusammen: 52 Läufe, 51 verschiedene Mutationen.

### 3.3 Klasse `nebenläufig` — Raten (je 30 Prozesse mit `-race -test.count=1` und 30 ohne Detektor, vier Prozesse zugleich, Messender `mutationslauf.sh`, Last des Hosts 7,9 bis 8,6 von 4 CPUs)

| Zusage | Mutation (Name) | mit Detektor | ohne Detektor |
|---|---|---|---|
| L11 (Zähler warten nicht) | `Log` nimmt vor dem Zählen der Verwürfe kurz die Sperre der Zeile (`vf-n-lock-unter-min`, `TestBlockierteSenke`) | 30 von 30 | 30 von 30 |
| L11 | dieselbe Sperre allein auf dem Weg unter `MinLevel` (`vf-n-lock-nur-unter-min`) | 30 von 30 | 30 von 30 |
| L17 (nie zwei `Write`) | Sperre nur für Zeilen über 160 Bytes, Haltepunkt (`vf-n-sperre-nur-lange`, `TestKeineZweiWriteHaltepunkt`) | 30 von 30 | 30 von 30 |
| L17 | dasselbe, Masse (`vf-n-sperre-nur-lange-masse`, `TestKeineZweiWriteMasse`) | 30 von 30 | 30 von 30 |
| L17 | Sperre nur für Zeilen bis 160 Bytes, Haltepunkt (`vf-n-sperre-nur-kurze`) | **0 von 30** | **0 von 30** |
| L17 | dasselbe, Masse (`vf-n-sperre-nur-kurze-masse`) | **0 von 30** | **0 von 30** |
| L18 (`Write` im Aufruf) | `Write` in einer Goroutine unter gehaltener Sperre, Haltepunkt (`vf-n-write-goroutine-haltepunkt`, `TestLogKehrtNachWriteZurueck`) | 30 von 30 | 30 von 30 |
| L20 (keine Zeilen-Sperre über `Now()`) | `Now()` unter der Sperre der Zeile im Weg von `eintrag` (`vf-n-now-unter-zeile-nur-bei-fehlerpfad`, `TestUhrRuftZeile`) | 30 von 30 | 30 von 30 |

Die Haltepunkt-Fänger sind in jedem gemessenen Lauf rot (zusammen mit den Raten des Implementers, 50 von 50, und des Reviews, 30 von 30). Die Zusage „`Write` läuft im Aufruf“ trägt allein der Haltepunkt (Implementer 100 von 100, Review 30 von 30, hier 30 von 30 und 30 von 30); die Masse `TestZeilenReihenfolgeMasse` führt der Plan als Zusatzfänger mit eigener Rate (24 von 50 mit und 49 von 50 ohne Detektor für die Form des Implementers, 19 von 30 und 30 von 30 im Review): **ehrlich geführt**; ich habe sie nicht neu gemessen, denn die Zusage hängt nicht an ihr. Der Detektor je Prozess (`-test.count=1`) und die Wirkung sind im Plan sauber getrennt: Haltepunkt-Fänger sind deterministisch in Richtung der Mutation, die Masse-Fänger nur mit Wahrscheinlichkeit.

Die zwei Zeilen „nur für kurze Zeilen“ sind eine Lücke der Form der Eingabe (siehe I-2), kein Produktfehler.

## 4. Auflösung der Befunde der Prüfrunde (Fix gegen die Klasse, `AGENTS.md` §3.17)

| Befund | Klasse (aus dem Wortlaut der Zusage) | Stand |
|---|---|---|
| F-1 MEDIUM, Maskierung | „jedes Zeichen außer den genannten steht unverändert“ (`SPEC-231` Nr. 1); 63 747 Elemente | **aufgelöst**: 15 eigene Mutationen an Elementen, die der Befund nicht nannte (Abschnitt 3.1), alle rot außer der erklärten (I-1). Prüfgrenze oberhalb U+FFFF: der Grund trägt nicht, siehe I-1. |
| F-2 INFO, Masse gegen Haltepunkt | „Zusage `Write` im Aufruf“ | **aufgelöst**: der Plan nennt den Haltepunkt als Träger, die Masse als Zusatzfänger mit Rate; Raten bestätigt. |
| F-3 INFO, Wirkung von `Log` auf den Stand | „Log ändert den Stand nicht über die Verwürfe hinaus“; Wege von `Log`: geschrieben, unter `MinLevel`, Senke defekt | **nur zu einem Drittel aufgelöst** — V-3. |
| F-4 INFO, Abdeckung | Zweig-Menge Z01 bis Z37 | **aufgelöst**: 98,7 Prozent unter dem Detektor selbst gemessen; die Blöcke ohne Treffer sind die im Plan genannten vier (Abschnitt 1). |

## 5. Befunde dieser Verifikation

### V-1 — MEDIUM — Die Länge von `msg` zählt Bytes des Textes, nicht der maskierten Form: kein Fänger

**Zusage:** `SPEC-231` Nr. 3, Absatz „Text und Zahlen“: „Die Länge zählt die Bytes dieses Textes, nicht die der maskierten Form in der Zeile.“ Der Satz steht nicht in L05 oder L06 und nicht in M2 oder M3 (Menge M2: ASCII-Füllung und ein Zeichen der Breite 1 bis 4 an der Grenze; M3: einzelne Zeichen).

**Beleg:** Mutationen an `msgText`, die `"`, Backslash und Steuerzeichen (`vf-msg-laenge-maskiert`, `-zeichen`) oder U+2028 und U+2029 (`vf-msg-laenge-maskiert-u2028`) mit der Länge ihrer Escape-Folge zählen: **grün, auch das ganze Paket**. Ein `msg` mit Anführungszeichen nahe der Grenze würde bei dieser Mutation früher gekürzt, und kein Test sieht es. Das ist `AGENTS.md` §3.15: eine Zusage des Vertrags ohne Fänger, ein MEDIUM.

**Menge (§3.17):** die Zeichen, die `anfuehren` anders schreibt (`"`, Backslash, 32 Steuerzeichen, U+2028, U+2029: 36 Zeichen) × Lage an der Grenze von 256 (Offset 250 bis 257). Ein Fänger als Schleife darüber: Füllung aus `a` bis Offset *o*, dann das Zeichen, dann `a`; Erwartung nach dem Modell (längster Anfang bis 256 Bytes des *Textes*), Rohbytes der Zeile. Rund 36 × 8 Elemente.

**Ort der Berichtigung:** Plan §3, Zeile L05 und Menge M2 (Achse Zeichen ergänzen), `zeile_test.go`. Kein Eingriff in Code oder Spezifikation.

### V-2 — MEDIUM — `timestamp` ist `Clock.Now()` des Aufrufs: kein Fänger für die Wiederverwendung des ersten Werts

**Zusage:** `SPEC-231` Nr. 2: „`timestamp` ist `Clock.Now()` zum Aufruf“. L02 (`TestTimestamp`) legt je Fall einen neuen Adapter an und prüft *eine* Zeile; die Zahl der Aufrufe von `Now()` ist gezählt, ihr Ergebnis wird nie über zwei Zeilen desselben Adapters mit verschiedenen Uhrständen verglichen. `TestZeilenform` schreibt je Aufruf zwei Zeilen, aber mit fester Uhr; `TestEinWrite` ebenso.

**Beleg:** Mutation, die den ersten gelesenen Wert festhält und für alle weiteren Zeilen verwendet, `Now()` aber weiter ruft (`vf-zeit-gemerkt2`): **grün, auch das ganze Paket**. Eine erste Fassung der Mutation (`vf-zeit-gemerkt`, Wert in einer Map abgelegt, unter der Sperre des Zustands gelesen) wurde durch Nachbartests rot, aber aus anderem Grund (Sperre über `Now()`, Zählprüfung der Gründe); sie zählt nicht. Die zweite Fassung ist sauber.

**Menge (§3.17):** Zeitwerte des Aufrufs × Zahl der Aufrufe je Adapter: mindestens zwei Zeilen desselben Adapters mit einer Uhr, die zwischen den Aufrufen weiterläuft. Fänger: Schleife über M5 (8 Zeitwerte) in *einem* Adapter, Erwartung die 8 Zeitstempel in der Folge der Aufrufe.

### V-3 — LOW (nahe MEDIUM) — Die Fix-Runde F-3 deckt einen von drei Wegen von `Log`

**Zusage (vom Plan geführt, Zusatzzeile von L13):** „`Log` ändert den Stand nicht über die Verwürfe hinaus“ (Plan §3, F-3; Wortlaut `SPEC-230` Kopf: „Telemetrie ändert nie das Ergebnis des Aufrufers“, `SPEC-232` Nr. 3). `TestLogAendertNurVerwuerfe` führt *einen* Weg aus: Stufe `ERROR`, heile Senke, alle gültigen Felder, ein ungültiges Attribut. Die Menge der Wege von `Log` hat drei Elemente: unter `MinLevel`, geschrieben, Senke defekt (Z01, Z02, Z28 bis Z33 der Zweig-Menge).

**Beleg:** Mutationen auf den beiden nicht geführten Wegen sind **grün im ganzen Paket**: `vf-f3-senkepfad-fehlerzaehler` und `-log` (Fehlerzähler), `vf-f3-senkepfad-wert` (zusätzlicher Verwurf `zustand`), `vf-f3-below-min-fehlerzaehler-minlevel` (Fehlerzähler unter `MinLevel`). L10 prüft bei defekter Senke `wert` und `senke`, nicht, dass die übrigen Gründe 0 bleiben. Zur Zufälligkeit: die erste Fassung der Mutation unter `MinLevel` erhöhte einen *zufälligen* Zähler (Iteration einer Map) und wurde einmal durch `TestBlockierteSenke` rot (`CountError 2`, weil der Test genau einen Code liest); die zweite Fassung mit demselben Text blieb grün. Ein Fänger mit einer Rate von etwa 1 zu 40 zählt nicht. Die Wirkung ist gering (die Spezifikation sagt es nicht ausdrücklich, der Plan führt die Zeile), aber die Fix-Runde nennt die Zusage und deckt ein Drittel der Wege.

**Menge:** Wege von `Log` (3) × Felder des Stands (`Errors`, `Durations`, `Dropped`, `Spans`); der Fänger als Schleife über die drei Wege mit dem vorhandenen Vergleich vor und nach dem Aufruf (`reflect.DeepEqual` auf `Errors` und `Durations`, jedes `Dropped` gleich dem Modell: 0 für alle Gründe außer dem erwarteten). Wenige Zeilen im bestehenden Test.

### Akzeptierte Negative und INFO

- **I-1 (INFO) — Prüfgrenze oberhalb U+FFFF: der Grund trägt nicht, die Folge ist benannt.** Der Plan nennt drei Vertreter (U+10000, U+10FFFF, U+1F600) und begründet: „der Code behandelt jedes Zeichen ab U+0800 gleich“. Das ist eine Aussage über den Code, kein Grund einer Prüfgrenze (`AGENTS.md` §3.17 verlangt sie für unbeschränkte Mengen; diese Menge hat 1 048 576 Elemente und ist endlich). Aufwand: `TestMaskierungRohbytes` braucht für 63 747 Elemente 0,09 s ohne und 0,37 s mit Detektor, also rund 1,4 µs und 5,8 µs je Element; die 1 048 576 Elemente oberhalb von U+FFFF kosten etwa 1,5 s ohne und 6 s mit Detektor, kein Hindernis für das Gate (die Stufe `race` liegt bei 24 bis 36 s). Die Mutation `vf-mask-1f601-ausserhalb` bleibt grün; ein Bereich ab U+20000 wird nur zufällig durch M3 gefangen. **Entscheidung beim Planner:** die Menge auf alle Zeichen bis U+10FFFF ausdehnen (empfohlen, das Ende einer Schleife) oder den Grund durch den gemessenen Aufwand ersetzen und die Folge belassen („eine Sonderbehandlung eines weiteren Zeichens oberhalb von U+FFFF bliebe ungefangen“).
- **I-2 (INFO) — Nebenläufige Fänger nur mit kurzen Zeilen.** Haltepunkt und Masse von L17 benutzen Zeilen unter 160 Bytes; eine Sperre, die nur Zeilen bis 160 Bytes schützt, bleibt in 0 von 30 Läufen rot (Abschnitt 3.3). Die Mutation ist konstruiert: `Write` hat im Code genau eine Stelle ohne Verzweigung nach der Länge. Beleg der Klasse „Form der Zeile“, kein Mangel; wer L17 stärken will, lässt den Haltepunkt eine Zeile mit allen 13 Feldern schreiben.
- **I-3 (INFO) — Akzeptierte Negative L15, L16 und das Fenster „zwischen `Write` und Freigabe, Zählung von `senke`“** tragen Grund und Folge. Die Mutation „Zählung von `senke` unter der Sperre der Zeile“ ist nicht beobachtbar (`mu` wird nie unter `zeilenMu` genommen, es entsteht kein Verklemmen), wie der Plan sagt; bestätigt durch Lesen des Codes (`zeilenMu` wird vor der Zählung freigegeben).
- **I-4 (INFO) — Das Image `odata-recorder:race` gehört mehreren Zweigen.** Mein erster Abdeckungslauf ohne Kopie des Baums traf das Image eines anderen Zweigs (Paket ohne `zeile.go`, 98,0 Prozent); der Lauf mit `docker cp` des eigenen Baums trifft den eigenen (98,7 Prozent). Der Plan nennt die Falle in den Betriebsregeln; die Zahl 98,7 stammt aus dem Lauf mit Kopie. Für künftige Berichte: bei der Abdeckung den Weg (Kopie oder Stufe) nennen.

## 6. ADRs und Hard Rules

- ADR-0001 (Go): nur Standardbibliothek im Adapter. ADR-0002 und ADR-0012 (Architektur-Gate): `make a-check` im Gate-Lauf grün, keine neue Import-Richtung. ADR-0003 (Lint): keine Unterdrückungsdirektive im Paket, `make lint` im Gate-Lauf grün. ADR-0007 (Detektor): `make test-race` im Gate-Lauf grün, eigener Lauf mit Profil grün.
- `AGENTS.md` §3.2 (kein Unterdrücken): erfüllt. §3.7 und §3.11 (Kommentare): `zeile.go` trägt Zusage und Kopplung ohne Chronik; eine Ausnahme: der Kopf von `TestLogAendertNurVerwuerfe` sagt „Log ändert den Stand nicht über die Verwürfe hinaus“, der Test führt einen Weg (V-3; §3.11 verlangt, dass der Kopf den einen Weg nennt oder die Menge deckt). §3.9 (Plan folgt jeder Korrektur): die Fix-Runde hat Tabelle, Raten und Absatz nachgezogen. §3.10: Tabelle Zusage · Mutation · Fänger und Raten im Plan, bei der Masse ehrlich. §3.14: Marke vor dem Code gesetzt, Messung 1468 von 2900. §3.15: V-1 und V-2 sind Zusagen ohne Fänger. §3.17: F-1 aufgelöst, F-3 nicht. §3.18: Gegenlesung selbst gefahren, 0 Treffer.
- Betriebsregel „nur `Edit` und `Write` auf Repo-Dateien“: im Diff kein Hinweis auf eine Verletzung; die Werkzeugaufrufe des Implementers prüft der Verifier nicht.

## 7. Auflagen an den Planner

1. **V-1, V-2, V-3 entscheiden:** Fix-Runde (Test-Commit, nur `zeile_test.go` und Plan §3, rund 30 bis 60 Zeilen: ein Fänger je Klasse als Schleife; Zeilen L05/M2, L02/M5 und die Zusatzzeile von L13 ergänzen) oder je Klasse ein akzeptiertes Negativ mit Grund und Folge im Plan. Meine Empfehlung: Fix-Runde, weil jede Klasse eine Zusage von `SPEC-231` trägt und die Fänger billig sind. Der Verifier fährt danach nicht erneut; der Reviewer prüft die Mutationen `vf-msg-laenge-maskiert`, `vf-msg-laenge-maskiert-u2028`, `vf-zeit-gemerkt2`, `vf-f3-senkepfad-fehlerzaehler`, `vf-f3-senkepfad-wert` und `vf-f3-below-min-fehlerzaehler-minlevel` (die Listen liegen im Scratchpad, Pfad in Abschnitt 3).
2. **I-1:** die Menge von F-1 auf alle Zeichen bis U+10FFFF ausdehnen oder den Grund der Prüfgrenze durch den gemessenen Aufwand ersetzen.
3. **Kopf von `TestLogAendertNurVerwuerfe`** (§3.11): den Weg nennen oder die Wege decken (V-3).
4. **Closure (§7 des Plans):** Messpunkt-Zeile für Handoff und Ende (`zeilen=20 nebenlaeufig=5 fakes=1 produktcode=260 schaetzung=2085 marke=2900 messung=1468`, Auslastung 0,51 unter 0,60: dem Architect melden, `SPEC-128` Nr. 9), Raten der Wirkungs-Fänger L11, L17, L18, L19, L20 (Plan §3 plus die 30-von-30-Läufe oben), Zusatzzeit der Stufen (Abschnitt 1: Paket 4,7 s, mit Detektor 31 bis 36 s unter Last 8 von 4), Bericht der Abdeckung (98,7 Prozent, vier Blöcke ohne Treffer), „Binary unverändert“.
5. **Register:** die Klassen „Fänger führt nur einen Weg einer Menge von Wegen“ (V-3) und „Satz des Vertrags neben der Hauptzusage ohne Zeile“ (V-1, V-2) sind Kandidaten für die Beobachtungen `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (28×) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×); die Entscheidung liegt beim Planner, dieser Zweig schreibt dort nicht. `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (7×): die Zweig-Menge ist durch die Abdeckung aufgelöst.
6. **Merge:** `main` in den Zweig ziehen und `make gates` wiederholen; Historie in `spec/spezifikation.md` §7 und Metriken-Tabelle (37 werden 38 mit `slice-mock-lesen`) löst der Integrationsschritt (Plan §6, Rückgabe 3). `slice-mock-lesen` hat die Sendung `ODR-E6008` in §1 angenommen.
7. **Offene Risiken §6 des Plans** tragen bei der Closure je einen Ausgang; „weiter offen“ mit Adresse User (L16, nie zurückkehrende Senke) bleibt im Hinweis für den User.
8. **Gegenlesung dieses Reports:** Auflage an die Rolle, die ihn committet: vor dem Commit `LC_ALL=C grep -nP` auf Backslash-u-Folge, BOM, NBSP und Zeilentrenner über diese Datei (der Report beschreibt solche Folgen in Worten und schreibt sie nicht aus).

## 8. Keine Spec-Lücke

Keine Randform ohne Ort gefunden; `SPEC-231` Nr. 1 bis 3 und `SPEC-232` Nr. 3, 5 tragen jede Entscheidung des Codes (Maskierung allein, Reihenfolge Ersetzen vor Kürzen, `FormatFloat`, Wert statt Name bei `MinLevel`, Fehlerformen der Senke, Sperre allein um `Write`). Die drei Lücken oben liegen im Test, nicht im Vertrag.

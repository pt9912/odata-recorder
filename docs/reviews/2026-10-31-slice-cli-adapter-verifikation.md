# Verifikation: slice-cli-adapter

* Datum: 2026-10-31
* Rolle: verifier (frischer Kontext; kein Code geändert)
* Gegenstand: Stand f905715 (Implementer e03912b, 63f65fb; Review 6205988; Fix-Runde c6bd5c4, f905715), Basis ba5caa1
* Plan: `docs/plan/planning/in-progress/slice-cli-adapter.md`
* Maßstab: DoD (§2), `SPEC-138`, Tabelle Zusage·Fänger (§3), Raster und Zusatztabelle (§6), `AGENTS.md` §3.9 bis §3.18, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md); Bezug `LH-FA-02`, `LH-RB-03`
* Eingang: Review `docs/reviews/2026-10-31-slice-cli-adapter-review.md` (R-1 bis R-7)
* Belege selbst gefahren: `make gates` und `make -k gates` (Worktree, Exit aus Datei), 22 Mutationen in Wegwerfkopien im Image `odata-recorder:race` (`--memory=4g --cpus=4 --network none`, Kopie nur lesend gemountet, unter `flock`), `make groessenmarke`.

## Urteil

**Nicht bestanden, solange `make gates` rot ist; ein einziger Befund, reine Doku, trivial behebbar (V-1). Kein falsches Codeverhalten, kein HIGH.** Nach dem Link-Nachzug in V-1 sind alle übrigen Gates grün (`make -k gates`: einziger Fehlschlag `docs-check`). Die Fix-Runde deckt die Klassen von R-1 bis R-3; R-4 und R-5 stehen als offene Rückgaben an den Architect in §6, der Code entscheidet sie nicht.

## 1. Gates (DoD-Punkt 3)

| Gate | Beleg |
|---|---|
| `make gates` | **Exit 2**, abgebrochen an `docs-check`: `docs/reviews/2026-10-31-slice-cli-adapter-review.md:12` meldet [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md) als `id-unlinked` (d-check, 2 Befunde in 505 Dateien) |
| `make -k gates` | alle übrigen Gates ohne Fehlschlag: `a-check` 0 Befunde, `baseline-verify` OK, `build`, `test`, `test-race` (Stufe `race`), `groessenmarke`, `handbuch-check`, `lint`; einziger Fehlschlag `docs-check` |
| `make groessenmarke` | `ok, Messung 1296 von Marke 2500` (Marke fest, nicht verändert seit `basis`); Hinweis *Tabelle ergibt 1950, Schätzung 1780* (`SPEC-128` Nr. 5 c, nur Hinweis, im Plan §4 genannt). Messung stimmt mit dem Plan überein |

## 2. Befunde

| ID | Stufe | Befund | Ort | Ausgang |
|---|---|---|---|---|
| V-1 | MEDIUM (Gate rot, kein Codefehler) | Die Kopfzeile *Eingangs-Kontext* des Review-Reports nennt die beiden ADR-Kennungen ohne Link; `make docs-check` und damit `make gates` fallen. Der Report ist ein Werkstück des Reviewers (6205988); weder Implementer-Runde noch Verifier haben es bemerkt, denn der Implementer meldete kein `make gates` nach dem Review. Behebung: beide Kennungen verlinken (`../plan/adr/0001-implementierungssprache-go.md`, `../plan/adr/0007-race-detektor-im-gate.md`), danach `make gates` erneut. | `docs/reviews/2026-10-31-slice-cli-adapter-review.md:12` | offen; Adresse Planner/Reviewer-Nachzug vor `git mv` nach `done/` |
| V-2 | LOW | `SPEC-138` *Ungeprüft* führt die beiden neuen Negative nicht: Kanalgröße genau 2 (Nr. 1) und die Vorabprüfung des vorbelegten Signals samt „genau einmal“ (Nr. 7 a, b). Sie stehen im Plan §3 (mit Grund und Folge) und §6 (Ausgang *offen*, Adresse Architect, `AGENTS.md` §3.13 erfüllt); der Plan C14 sagt „je mit Grund und Folge in *Ungeprüft* von SPEC-138“, was für diese Posten erst nach dem Nachzug des Architect stimmt. | `spec/spezifikation.md` SPEC-138 *Ungeprüft*; Plan §3 C14, §6 | offen; Adresse Architect (R-4, R-5 gleichlautend) |
| V-3 | LOW (akzeptiertes Negativ) | `TestVersionKlasse` ist an mehrbytigen Zeichen auf fünf Folgen begrenzt. Eine Mutation, die die gültige Folge U+2028 (`e2 80 a8`) aus `Version` entfernt, bleibt grün. Grund: die Menge „jedes Zeichen“ ist unbeschränkt (1,1 Mio. Codepunkte); die Byte-Klasse 0x00 bis 0xFF plus fünf Folgen fängt alle bytebasierten und generischen Mutationen (TrimX, ToValidUTF8, Kürzen, Entfernen von Steuerzeichen). Folge im Betrieb: ein künstlich auf ein einziges Zeichen zielender Eingriff bliebe unbemerkt. | `src/adapters/driving/cli/kommando_test.go` | akzeptiert, Plan sollte die Prüfgrenze (`AGENTS.md` §3.17) in §3 nennen |

## 3. DoD gegen Code, Plan und Vertrag

| DoD-Punkt | Befund |
|---|---|
| Kommandoauswahl und Exit-Codes nach `SPEC-138` Nr. 1 bis 5, 9; `bootstrap.Run` entfällt, `exitOK` und Paketkommentar bleiben; `main` ruft `cli.Run` | erfüllt: `cli.go` (121 Zeilen), `bootstrap.go` nur Paketkommentar und `exitOK`, `main.go` ruft `cli.Run`; `bootstrap_test.go` gelöscht |
| Signale nach Nr. 6, 7; Raster in §6, vor dem Code entschieden | erfüllt für Nr. 6 und 7 a bis d; Randformen *geschlossener Kanal* und *Abmeldung nil* sind nicht entschieden, sondern als offene Zeile an den Architect gegeben (R-5, siehe Abschnitt 5) |
| `make gates` grün, `a-check`, `test-race` | **nicht erfüllt** (V-1); `a-check` und `test-race` selbst grün |
| Review, Report unter `docs/reviews/` | Report liegt vor; Häkchen steht noch aus (Closure) |
| Doku-Update, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | Closure-Pflichten des Planners; alle DoD-Haken sind erwartungsgemäß noch leer. Jedes Risiko in §6 trägt einen Ausgang (entschieden / offen mit Adresse / akzeptiert) |

## 4. Tabelle Zusage·Fänger gegen SPEC-138 (nicht nur gegen den Plan)

| SPEC-138 | Fänger | Befund |
|---|---|---|
| Nr. 1 Schnittstelle: nil-Writer, `Start` nie nil, Schreibfehler | C4 (18) | gedeckt |
| Nr. 1 nur Standardbibliothek | C13 | gedeckt |
| Nr. 1 Kanalgröße 2 | C16 (Puffer 0 rot), Größe 1 vs. 2 als Negativ im Plan | gedeckt bzw. Negativ mit Grund und Folge (V-2) |
| Nr. 1, 7 d Abmeldung `signal.Stop` | C15 | gedeckt (siehe Abschnitt 5) |
| Nr. 1 „main übergibt an `os.Exit`, sonst beendet kein Code den Prozess“ | keine eigene Zeile; `os.Exit` im Paket beendete den Testprozess (rot) | Randnotiz, kein Befund |
| Nr. 2 Kommandos, bytegleicher Vergleich | C1 (8), C2 (100 plus Klasse 1558) | gedeckt |
| Nr. 3 Version | C3 (10 plus Klasse 1035) | gedeckt |
| Nr. 4 Ablehnung, Wortlaut, Zusatz, `start` bei nil | C2 | gedeckt |
| Nr. 5 Exit-Codes | C6 (10 plus Klasse 2365), C10 (130/143) | gedeckt |
| Nr. 6 Start genau einmal, Writer, Kontext | C5, C9 | gedeckt |
| Nr. 7 a anmelden vor `Start`, vorbelegt | C7, C8 (Fake vorbelegt); Vorabprüfung nicht trennbar | Negativ, offen beim Architect |
| Nr. 7 b erstes Signal, Art ändert Code nicht, wartet ohne Frist | C8 (8), C9 (6) | gedeckt; „genau einmal“ nicht beobachtbar (offen, Architect) |
| Nr. 7 c zweites Signal, 130/143, nichts schreiben | C10 (10) | gedeckt |
| Nr. 7 d Abmeldung genau einmal, jeder Pfad | C11 (9) | gedeckt |
| Nr. 7 e, Nr. 8 | `slice-bootstrap-m2` S1 bis S4 (Adresse Planner, §6) | nicht in diesem Slice |
| Nr. 9 kein Zustand | C12 (4) | gedeckt |
| Nr. 10 kein neuer Code | C5 (Adapter schreibt nichts), C10 | gedeckt |

Die Tabelle trägt C1 bis C16 (C15 und C16 aus dem Review); die Reihenfolge C15, C16 vor C14 ist ein Schönheitsfehler. Der Plan §3 „Gemessen“ trägt zu jeder nebenläufigen Zeile Zahl der Läufe und Prozesse.

## 5. Fix-Runde je Befund gegen die Klasse (`AGENTS.md` §3.17)

Eigene Mutationen im Image `odata-recorder:race`, ungemutiert grün (Basislauf `ok`). „Rot“ nennt den Test, der aus dem Grund der Mutation fiel.

| Befund | Menge (Elemente) | Fänger | Eigene Mutation an einem vom Befund nicht genannten Element | Ergebnis |
|---|---|---|---|---|
| R-1 Version unverändert | Bytes 0x00 bis 0xFF in 4 Positionen, 5 Mehrbyte-Folgen, 3 Längen, 3 Formatfolgen = 1035 (Zählprüfung im Test) | `TestVersionKlasse` | DEL (0x7f) entfernt · `ToLower` · Abschnitt ab 0x1f abgeschnitten · 0xFF entfernt (der Befund nannte `TrimRight`, NUL, `ToValidUTF8`) | 4 von 4 rot (`TestVersionKlasse`); **grün bleibt** nur U+2028 entfernt (V-3, Prüfgrenze unbeschränkter Menge) |
| R-2 Kanalpuffer | {Puffer 0, 1, 2}; Paare verschiedener Signale = 2 | `TestSignaleHintereinander` | Puffer 0 rot (`terminated/interrupt`, `interrupt/terminated`, Frist); Puffer 1: grün (benanntes Negativ im Plan §3, Grund und Folge genannt, Mutation 0 von 50) | wie im Plan; die Zusage „genau 2“ ist nicht von „1“ trennbar, Negativ nicht in SPEC-138 (V-2) |
| R-3 Abmeldung der Laufzeit | Signale {SIGTERM, SIGINT} = 2, je Kind-Prozess | `TestLaufzeitAbmeldung` | `signal.Reset(SIGTERM)` allein · `signal.Reset(os.Interrupt)` allein · `signal.Ignore(beide)` · `signal.Stop` auf fremdem Kanal (der Befund nannte `_ = kanal` und leere Abmeldung) | 4 von 4 rot (Kind endet mit 92 statt durch das Signal; je Signal geprüft) |
| R-3 Anmeldung der Laufzeit (Nachbarklasse) | {SIGTERM, SIGINT} | `TestErstesSignal` (echt) | `Notify` nur SIGINT · `Notify` nur SIGTERM | 2 von 2 rot (`echt/terminated`, `echt/interrupt`) |
| R-4 Vorabprüfung, „genau einmal“ | – | – | – | kein Code geändert; §6 trägt Zeile mit Ausgang *offen*, Adresse Architect (zweimal, Z. 222 und 223); Plan §3 trägt Grund und Folge |
| R-5 Quelle mit geschlossenem Kanal oder nil-Abmeldung | – | – | `git diff 63f65fb..HEAD -- src/adapters/driving/cli/cli.go` leer | **nicht im Code entschieden**; §6 trägt die Zeile mit Ausgang *offen*, Adresse Architect, im selben Commit wie die Fix-Runde (`c6bd5c4`) |

Hinweis zu R-3: das Kind ruft `Run` mit einem sofort zurückkehrenden `Start`; die Abmeldung auf dem Pfad „Start liefert ungleich 0“ bleibt am echten Signal unbeobachtet (die Fake-Quelle zählt sie in C11). Das ist dieselbe Funktion `abmelden` per `defer`, ein Pfad-spezifischer Eingriff wäre eine Änderung am Code, nicht an `laufzeitSignale`.

## 6. Offene Rückgaben und Adressen (`AGENTS.md` §3.13)

| Rückgabe | Ort | Adresse | Befund |
|---|---|---|---|
| „genau einmal“ (Nr. 7 b) nicht beobachtbar | §6 Z. 222 | Architect, Ausgang *offen* | korrekt adressiert |
| Vorabprüfung vorbelegtes Signal (Nr. 7 a) nicht trennbar | §6 Z. 223 | Architect, Ausgang *offen* | korrekt adressiert |
| Quelle: geschlossener Kanal, nil-Abmeldung (R-5) | §6 Z. 224 | Architect, Ausgang *offen*; Code unverändert | korrekt adressiert |
| `Konfig.Start` gleich `nil` in `main`, `start` ist Exit 2 | Plan §1, §6; `slice-bootstrap-m2` §1 *Übernimmt (Start-Befehl)* und DoD tragen die Sendung | Planner/`slice-bootstrap-m2` | getragen; der Nehmer liegt in `open/`, nicht `done/` |
| SIGTERM am gebauten Binary | §6 Z. 220, `SPEC-138` *Ungeprüft* (Verdrahtung in `main`) | akzeptiert; `slice-test-integration` nicht zugewiesen, daher keine Adresse | akzeptiertes Negativ mit Grund und Folge, konsistent |
| Handbuch-Kapitel, Folge-Pläne, Katalog | §6 Z. 214 bis 218, 221 | Planner, Ausgang *offen* | korrekt adressiert |

## 7. Weitere Prüfungen

* `AGENTS.md` §3.9: Fix-Runde zog Tabelle (C3, C15, C16), Gemessen, Negative, §4 Messung (1296) und §6 im selben Commit nach; Plan-Kopf und §1 unverändert und weiter wahr.
* `AGENTS.md` §3.18: `kommando_test.go` baut NUL, NBSP, BOM aus Bytewerten (`string([]byte{...})`); keine Escape-Folge als Klartext in den geänderten Dateien.
* `AGENTS.md` §3.7, §3.11: die Kommentare von `TestLaufzeitAbmeldung` und `TestSignaleHintereinander` sagen zu, was die Tests prüfen.
* Zeitverhalten: ungemutiert und unter `-race` grün in `make test-race`; die geforderte Rate bei `frist` 5 s ist im Plan mit 50 von 50 Läufen genannt (vom Reviewer als R-6 notiert, nicht widerlegt).

## 8. Blocker und Reste

* Blocker: nur V-1 (Doku-Link, `make gates` rot). Kein HIGH, kein falsches Codeverhalten.
* Reste mit Grund und Folge: V-2 (SPEC-Nachzug beim Architect, Folge: Spec nennt zwei Negative nicht), V-3 (Prüfgrenze mehrbytige Zeichen, Folge: gezielte Ein-Zeichen-Mutation unbemerkt), R-7 (`start` am Binary erst mit `slice-bootstrap-m2`).

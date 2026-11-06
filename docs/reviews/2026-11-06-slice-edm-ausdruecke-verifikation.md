# Verifikations-Report: slice-edm-ausdruecke — 2026-11-06

**Rolle:** Verifier, frischer Kontext (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-166`, `SPEC-167`, `SPEC-109` Nr. 5, `LH-FA-08.b`, `AGENTS.md` §3.9 bis §3.18). Kein Review, keine Reparatur: kein Produktcode, keine Roadmap, keine Welle, kein Register geschrieben.

**Gegenstand:** Zweig `slice-edm-ausdruecke`, HEAD `c66329c` (Diff `f6daf34..c66329c`; Review `4bdbd1a`, Architect `777ae26` und `1134fd2`, Fix-Runde `e474eda`, `22d215a`, `c66329c`, Merge von `main` `ec01ce3`). Plan `docs/plan/planning/done/slice-edm-ausdruecke.md`; Review-Report `docs/reviews/2026-11-06-slice-edm-ausdruecke-review.md`.

**Skripte:** die Fassung nach dem Merge von `main` im Zweig: `.harness/skills/reviewer.md` (MEDIUM-Zeilen „Fänger schmaler als Zusage, Menge unbeschränkt oder ohne Achsenliste“ und „Fixture, Hilfe oder Kommentar verdeckt, was der Test prüft“) und `.claude/commands/implement-slice.md` (Schritt 14: Menge Ziffer mal Stelle, Prüfgrenze oder Differenz im Plan; Schritt 19: Tabelle Menge · Achsen · Elementzahl · nicht genannte Mutation je Achse; Schritt 21: Handoff-Commit trägt die Tabellen in §3 nach).

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-06

**Methode.** Keine Zahl des Implementers oder Reviewers übernommen. Wegwerfkopien im Scratchpad `ver-ed/` (flüchtig, nach dem Lauf gelöscht; Reviewer- und Implementer-Ordner unberührt). Mutationen mit `perl` auf der Kopie, Lauf im vorhandenen Image `odata-recorder:test` bzw. `odata-recorder:race` (`--network none`, `--memory 2g`, ein schwerer Lauf zugleich unter `flock /tmp/odata-recorder-heavy.lock`; kein eigenes Image gebaut, nichts zu löschen). Last und Messender der Raten: Verifier, Host mit Lastmittel 4 bis 12 durch drei weitere Agenten.

---

## 1. Verdikt

**BESTANDEN mit Nachzug.** Kein HIGH, kein Fehler im Produktcode gefunden: von 36 eigenen Mutationen (25 rot, 11 grün, Abschnitt 4) und 19 Zählern für die Nebenläufigkeit (7 mit 20 von 20, 12 mit 0 von 20, Abschnitt 5), alle an Elementen, die weder der Reviewer noch der Implementer nannte, zeigen die grünen und die 0 von 20 **fünf Mengen, in denen ein Fänger schmaler ist als die Zusage** (V-1 bis V-5, alle MEDIUM, alle in Tests und Plan, je ein Fall oder ein benanntes Negativ). `make gates` ist grün, die Gate-Kette ist unberührt, die Sendung an `slice-query-interpreter` steht im Zweig, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) ist gewahrt.

**Empfehlung Closure: ja**, nachdem der Planner/Implementer V-1 bis V-5 je mit einem Fall oder einem akzeptierten Negativ (Grund, Folge) in §3 des Plans gelöst und V-6 nachgezogen hat. Eine zweite Verifikationsrunde ist nicht nötig: die Nachträge ändern keinen Vertrag und keinen Produktcode.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| V-1 | MEDIUM | **Fänger schmaler als Zusage: `TestGleichzeitig` führt nur einen Teil der Zweige.** `SPEC-166` Nr. 8 und `SPEC-167` Nr. 8 sagen für jede Funktion „gleichzeitige Aufrufe beeinflussen sich nicht“. Der Fix zu F-5 hat die Funktionen verbreitert (13), aber `PruefeFilter` läuft mit genau zwei Bäumen (ein gültiger, ein `literal_bereich`) und `Vergleiche` führt keinen Fehlerpfad. Ein unbeschütztes Paketzähler-`++` in einem nicht geführten Zweig bleibt im Detektor in 0 von 20 Prozessen unsichtbar (Nr. 5). Die Tabelle „Fix-Runde nach dem Review“ nennt die Achse *Zweig* nicht, und kein akzeptiertes Negativ nennt Grund und Folge. | Abschnitt 5; Plan §3 *Fix-Runde*, Zeile F-5 · `Funktion 13 · Klasse 6 und null · Operator 6 · Seite 2` | Implementer (Fall je Zweig) oder Planner (Negativ) |
| V-2 | MEDIUM | **Zusage ohne Fänger: „auch hier entscheidet der erste Fund der Folge“.** `SPEC-167` Nr. 7 sagt für Programmfehler (`nil`, Zeiger, unbekannter Operator, Literal mit unbekannter Art) „auch hier entscheidet der erste Fund der Folge von Nr. 3“. `TestProgrammfehler` führt jeden Programmfehler nur mit gültigem Rest, `TestReihenfolgeVerstoesse` nur die sieben Ursachen. Mutation `p21` (Operator des Vergleichs vor den Kindern geprüft), `p22` (ebenso bei der Verknüpfung), `q04` (`nil` in `Rechts` vor dem Teilbaum von `Links`), `q05` (unbekannte Art des Literals in `Rechts` vor `Links`): 0 rote Tests. Szenario: ein Baum mit einem unbekannten Operator und einer unbekannten Property im Kind meldet `ErrUngueltig` (`ODR-E9001`, ein Programmfehler) statt `400`. | Abschnitt 4, Zeilen `p21`, `p22`, `q04`, `q05`; Plan §3 Zeile 34 und 37 | Implementer |
| V-3 | MEDIUM | **Menge Zeichen mal Stelle der Guid nicht erschöpft (Fix zu Zeile 19 deckt die Klasse nicht).** `SPEC-166` Nr. 7: eine Guid sind 36 Zeichen aus Kleinbuchstaben und Ziffern. `randformen` wechselt die Schreibweise nur an Stellen, an denen der Samen einen Buchstaben trägt; der Samen `0a1b2c3d-4e5f-6789-abcd-ef0123456789` trägt an 26 der 32 Hex-Stellen eine Ziffer, `00000000-…` nur Ziffern. Mutation `p08` (Großbuchstabe `A` bis `F` an Stelle 35 angenommen), `q01` (Stelle 0), `q02` (Stelle 26): 0 rote Tests. Der Fix nannte `z19-hexgross` (an Buchstabenstellen rot). | Abschnitt 4; `src/hexagon/model/wert_vergleich_test.go` · `func randformen` (`ersatz` ohne `A` bis `F`), `gueltigeSamen` | Implementer |
| V-4 | MEDIUM | **Länge des Bruchteils eines Zeitpunkts: Menge 1 bis 12 nur an vier Elementen gefangen.** `SPEC-166` Nr. 7: „Bruchteil von 1 bis 12 Ziffern“. Gültige Längen im Test: 1, 2, 11, 12 (aus Samen und Löschen), ungültig: 13. Mutation `q09`/`r2`/`r3` (Länge 7, 3, 10 ungültig): 0 rote Tests; `r4` (Grenze 11): rot. Plan §3 nennt für Zeile 7 und 19 die Länge 13 und 12, nicht die Menge. | Abschnitt 4, Zeilen `q09`, `r2`, `r3`, `r4` | Implementer |
| V-5 | MEDIUM | **Unbeschränkte Menge ohne Prüfgrenze im Plan.** Die Zusage „Text sonst unbegrenzt“ (Raster R1) hat seit F-1 den Fänger `TestTextLaenge` bis 65537 Bytes. Mutation `r1` (Vergleich nur bis Byte 70000) bleibt grün, wie bei jeder endlichen Liste. Der Plan nennt die Grenze 65537 nicht als Prüfgrenze mit Grund und Folge im Betrieb (Schritt 14 und `.harness/skills/reviewer.md`: „Menge unbeschränkt ohne Prüfgrenze“), weder in der Fix-Tabelle noch unter „Abweichungen und akzeptierte Negative“. | Plan §3 *Fix-Runde*, Zeile F-1; Abschnitt 4, Zeile `r1` | Planner/Implementer (ein Absatz) |
| V-6 | LOW | **Plan folgt der Fix-Runde in §6 nicht (§3.9).** Die Zeilen „Randform: ein `FilterTypFehler` mit einer Ursache außerhalb der sieben“ und „Reihenfolge von `AlleFilterTypUrsachen`“ tragen weiter „Fänger offen beim Implementer“ bzw. „Kommentar offen beim Implementer“; beides ist seit `e474eda` erledigt (`TestFilterTypFehler` mit Ursache `x`, Kommentar auf „liefert die sieben Ursachen“). | Plan §6, Zeilen 290 und 292 | Planner (Closure) |
| V-7 | LOW | **Sendung an `slice-query-interpreter` steht nur im Zweig.** Der Nehmer führt die Sendung (a) bis (c) mit der Kennung dieses Slice in seinem §1 (`**Übernimmt (Sendung von slice-edm-ausdruecke …)**`), im Zweig `c66329c`; auf `main` fehlt sie bis zum Merge. §3.13 ist erfüllt (Nehmer liegt in `open/`, sein §1 Ausschluss trifft die Sendung nicht, seine DoD trägt sie), der Commit der Sendung ist der des Zweigs. Solange der Zweig nicht gemergt ist, sieht kein Leser von `main` die Sendung. | `docs/plan/planning/open/slice-query-interpreter.md` (im Zweig); `git diff main..c66329c` | Planner (Merge-Reihenfolge) |
| V-8 | INFO | `make test-race` lief im Gate-Lauf ohne das bekannte SQLite-Flake (`TestTransaktionVertrag/…ReplaceAll`); nicht Gegenstand, nur vermerkt. | `make gates` Exit 0 | — |
| V-9 | INFO | Das akzeptierte Negativ zu `r3` (geteiltes Ergebnisfeld ohne Wirkung ohne Detektor) ist mit Grund und Folge im Plan geführt und von mir nachgemessen (Abschnitt 5, Absatz `r3`); es trägt. | Plan §3 *Folge für `r3`* | — |
| V-10 | INFO | Drei Mutationen meiner Probe waren Mutations-Artefakte (`q06` ohne Wirkung, `p25` und `q10` Aufbaufehler) und eine war rot nur durch Abbruch (`q14` rot durch `panic: slice bounds out of range` statt `--- FAIL` eines benannten Tests); die drei Artefakte zählen nicht, `q14` zählt als rot. | Abschnitt 4 | — |

## 3. DoD und Plan, Punkt für Punkt

| DoD (§2) | Stand | Beleg |
|---|---|---|
| Vergleich und Sortierung je Typ, `Edm.Int64` jenseits 2^53, `Edm.Decimal` ohne Binär-Float | erfüllt | Mutationen `p01`, `p02`, `p03`, `p04`, `p11`, `p12` rot (`TestDecimalExakt`, `TestOrdneWerteGleichVergleiche`, `TestZeitpunktOrdnung`, `TestBereich`); Lücken nur V-3 und V-4 (Form, nicht Ordnung) |
| Null-Semantik, Wahrheitstafeln über alle Kombinationen | erfüllt | `p15`, `q12`, `q15` rot; `TestUndTafel`, `TestOderTafel`, `TestNichtTafel`, `TestNullwertWahrheit` vorhanden |
| Typprüfung des Filter-AST (unbekannte Property, `400`, `ODR-E6003`, Wurzel/Operanden, Literal gegen Literal, `null` gegen Typ, Typ außerhalb der acht `501`) | erfüllt, Lücke V-2 (Folge bei Programmfehlern) | `p23`, `p24`, `p26`, `p27`, `q03`, `q11` rot; Cross-Probe mit dem Leser (Abschnitt 6) grün |
| `make gates` grün | **erfüllt** | Exit 0 auf `c66329c` (Abschnitt 7) |
| Review durchgeführt, Report vorhanden | erfüllt für `04c6e0e`; die Fix-Runde (`e474eda` bis `c66329c`) ist nach Vorgabe nicht erneut reviewt, hier von mir gegen die Klasse aufgelöst | `docs/reviews/2026-11-06-slice-edm-ausdruecke-review.md` |
| Doku-Update (öffentlicher Vertrag) | erfüllt | `SPEC-166`, `SPEC-167`, Zeile in `SPEC-109` Nr. 5, Historie-Zeile; kein Ordner `docs/user` betroffen |
| Closure-Notiz, Register, Ausgang je Risiko aus §6, drei Paarungen | **offen** (Planner, bei der Closure) | §7 trägt „Wird bei der Closure gefüllt“; §6: zwei Zeilen *weiter offen* mit Adresse (Sendung an `slice-query-interpreter`, Widerspruch §3.14 gegen `SPEC-128` an den User), zwei *entschieden* mit veraltetem Zusatz (V-6) |

**Plan gegen Code-Diff.** Alle 38 Zeilen der Tabelle *Zusage · Fänger* haben ihren Test im Code (Namen per Suche je Test auf genau eine Definition im Paket des Fängers geprüft; `TestGuid` und `TestGleichzeitig` führen zusätzlich Namensvettern in anderen Paketen). Die Pfade des Blocks (11) und der Zeilen in fremden Tests (`beenden_test.go`, `helfer_test.go`, `kontext_test.go`, `reset_test.go`) stimmen mit dem Diff überein; `git diff --numstat f6daf34..c66329c` über die elf Pfade des Blocks: **3044** (Produktcode 708: 182 + 76 + 65 + 384 + 1; Tests 2336), `make groessenmarke` meldet „ok, Messung 3044 von Marke 5000“, Auslastung 0,61. Zusätzlich im Diff gegenüber `main`: nur Spec, Plan, Review, `slice-query-interpreter.md` (Sendung) und die genannten Code- und Testdateien; `Makefile`, `harness/`, `.a-check.yml`, `.golangci.yml`, `Dockerfile`, `tools/`, `AGENTS.md`, `.claude/`, `.harness/` sind gegenüber `main` unverändert (Gate-Kette unberührt).

**Mehrdateien-Änderung durch die Tabellenzeile `ErrFilterNichtBeobachtet`.** Die Zeile steht in `src/hexagon/ports/outbound/abbildung.go`; `TestTabellenzeilenVollstaendig` und `TestTabellenFehlerVollstaendig` halten die Listen gegen `EintraegeAbbildung`. Nachgezogen sind `session/beenden_test.go` (Liste), `workspace/helfer_test.go` (Liste), `kontext_test.go` (`19*26`), `reset_test.go` (`4*(26+4+1)`). Weitere Verbraucher von `EintraegeAbbildung` (`outboundtest/odata_codec_allaussagen.go`, `sqlite/fehler_test.go`) sind im Gate-Lauf grün; die Menge der Verbraucher ist damit durch `make test` und `make test-race` belegt, nicht nur durch die vier genannten Dateien.

## 4. Eigene Mutationen (Wegwerfkopie, `go test` des Pakets `src/hexagon/model`)

Jede Mutation ist ein Element, das weder der Plan noch der Review nannte. *Rot* heißt `--- FAIL` eines benannten Tests, außer wo vermerkt.

| Kennung | Achse | Mutation | Ergebnis |
|---|---|---|---|
| p01 | Zahl, Folge | Nachkommastellen zuerst nach Länge | rot: `TestDecimalExakt`, `TestOrdneWerteGleichVergleiche`, `TestGleichzeitig` |
| p02 | Zahl, Länge | ganzer Teil ohne Längenvergleich | rot: `TestDecimalExakt`, `TestOrdneWerteGleichVergleiche`, `TestZahlGleichUeberArten` |
| p03 | Zeitpunkt, Stelle | Sekunden nur bis zur Zehnerstelle | rot: `TestZeitpunktOrdnung`, `TestOrdneWerteGleichVergleiche` |
| p04 | Zeitpunkt, Länge | Bruchteil bis Stelle 11 verglichen | rot: `TestZeitpunktOrdnung`, `TestOrdneWerteGleichVergleiche`, `TestGleichzeitig` |
| p05 | Datum, Stelle | zweiter Strich nicht geprüft | rot: `TestFormverstoesse` |
| p06 | Zeitpunkt, Stelle | `T` nicht geprüft | rot: `TestFormverstoesse` |
| p07 | Zeitpunkt, Stelle | zweiter Doppelpunkt nicht geprüft | rot: `TestFormverstoesse` |
| **p08** | **Guid, Zeichen mal Stelle** | **`A` bis `F` an Stelle 35 angenommen** | **grün (V-3)** |
| p09 | Datum, Zeichen | zweite Monatsziffer nicht geprüft | rot: `TestFormverstoesse` |
| p11 | Zahl, Form | leerer ganzer Teil (`.5`) gültig | rot: `TestBereich` (mit `panic: index out of range`, rot, aber nicht als `--- FAIL` allein) |
| p12 | Zahl, Form | leerer Bruchteil (`1.`) gültig | rot: `TestBereich` (mit `panic: index out of range`) |
| p15 | Null, Seite | `x gt null` wahr (nur rechts) | rot: `TestNullVergleiche`, `TestGleichzeitig` |
| p18 | Guid, Seite | `OrdneWerte(null, guid)` gültig | rot: `TestGuid`, `TestGleichzeitig` |
| **p21** | **Programmfehler, Folge** | **Operator des Vergleichs vor den Kindern geprüft** | **grün (V-2)** |
| **p22** | **Programmfehler, Folge** | **Operator der Verknüpfung vor den Kindern geprüft** | **grün (V-2)** |
| p23 | Bereich, Typ | `Edm.Decimal` mit der Grenze von `Edm.Int64` | rot: `TestBereich` |
| p24 | Fehlertext | `unterstuetzt` statt `unterstützt` | rot: `TestFilterTypFehler` |
| p26 | Literal, Klasse | `true` als `null`-Klasse | rot: `TestLiteralGegenLiteral`, `TestVertraeglichkeit` |
| p27 | Guid, Seite | Ordnungsoperator nur bei Guid links, nicht gegen `null` | rot: `TestGuidOperatoren`, `TestNullGegenTyp` |
| q01 | Guid, Zeichen mal Stelle | `A` bis `F` an Stelle 0 | **grün (V-3)** |
| q02 | Guid, Zeichen mal Stelle | `A` bis `F` an Stelle 26 | **grün (V-3)** |
| q03 | Wahrheit, Stelle | Operand von `not` nicht geprüft | rot: `TestBooleschOperand`, `TestReihenfolgeVerstoesse` |
| q04 | Programmfehler, Folge | `nil` in `Rechts` der Verknüpfung vor `Links` | **grün (V-2)** |
| q05 | Programmfehler, Folge | unbekannte Art des Literals in `Rechts` vor `Links` | **grün (V-2)** |
| q07 | Datum, Stelle | Jahr: erste Ziffer nicht geprüft | rot: `TestFormverstoesse` |
| q08 | Zeitpunkt, Stelle | zweite Sekundenziffer nicht geprüft | rot: `TestFormverstoesse` |
| **q09** | **Zeitpunkt, Länge** | **Bruchteil von 7 Ziffern ungültig** | **grün (V-4)** |
| q11 | Zahl, Form | `-0` gültig | rot: `TestBereich`, `TestFormverstoesse` |
| q12 | Null, Klasse | `Edm.Boolean` gegen `null` mit `ne` falsch | rot: `TestNullVergleiche`, `TestGleichzeitig` |
| q13 | Null, Seite | `OrdneWerte(null, Datum)` | äquivalent (der Textvergleich liefert -1); rot nur durch `TestGleichzeitig` mit `ArtNull` und Text `x` |
| q14 | Null, Seite | `OrdneWerte(Zeitpunkt, null)` | rot durch `panic: slice bounds out of range`, nicht durch einen benannten Test (V-10) |
| q15 | Wahrheit, Zustand | `WahrheitVon(null)` falsch | rot: `TestWahrheitVonUndDurchlaesst`, `TestGleichzeitig` |
| **r1** | **Text, Länge** | **Vergleich nur bis Byte 70000** | **grün (V-5, unbeschränkte Menge)** |
| **r2** | **Zeitpunkt, Länge** | **Bruchteil von 3 Ziffern ungültig** | **grün (V-4)** |
| **r3** | **Zeitpunkt, Länge** | **Bruchteil von 10 Ziffern ungültig** | **grün (V-4)** |
| r4 | Zeitpunkt, Länge | Bruchteil höchstens 11 Ziffern | rot: `TestFormverstoesse`, `TestGleichzeitig`, `TestOrdneWerteGleichVergleiche`, `TestOrdnungTotal` |

Nicht gezählt (Aufbau-Artefakte): `q06` (ohne Wirkung), `p25` und `q10` (Aufbaufehler). Der Reviewer-Kern (Wurzel, Maske, Seite der Bereichsprüfung, Klasse der Fehler, Zeile 36) ist mit den Mutationen des Reviews und des Implementers belegt; ich habe ihn nicht dupliziert.

**Fix gegen die Klasse (§3.17, Pflicht-Menge-Tabelle).** Prüfung je Fix-Zeile gegen die Achsen Zeichen, Länge, Folge, Seite, Stelle, Zustand:

| Fix | Achsen der Tabelle | Achse, die ich zusätzlich mutierte | Ergebnis |
|---|---|---|---|
| F-1 Länge Text | Länge, Folge, Seite, Operator | Länge hinter der getesteten Menge (`r1`) | grün: Prüfgrenze fehlt im Plan (V-5) |
| F-2 Zahl, Länge mal Vorzeichen | Länge, Vorzeichen, Gestalt, Art, Aufruf | Länge des Bruchteils im Zeitpunkt (eigene Klasse, `q09`, `r2`, `r3`) | grün (V-4) |
| F-3 Typ ohne Properties | Zustand, offen, Name, Stelle | Programmfehler an der Stelle (`q04`, `q05`) | grün (V-2, gehört zu Zeile 37, nicht zu F-3) |
| F-4 Nullwert `Wahrheit` | Zustand, Funktion, Partner, Seite | `WahrheitVon(null)` (`q15`) | rot |
| F-5 alle Funktionen | Funktion, Klasse, Operator, Seite | Zweig und Fehlerpfad (Nr. 5) | 0 von 20 in 12 von 19 Zählern (V-1) |
| F-6 fremde Ursache | Ursache, Sentinel, Code und Text | Text mit Umlaut (`p24`) | rot |
| F-7 Kommentare | — | — | Kommentar `AlleFilterTypUrsachen` und Satz im Test sind gekürzt (Abschnitt 8) |

## 5. Nebenläufigkeit: Raten (Messender: Verifier)

Aufbau: je Mutation ein Binary mit `go test -race -c` im Container `odata-recorder:race` (Go 1.27.1, `--memory 2g`), danach **20 Prozesse** `-test.run '^TestGleichzeitig$' -test.count=1`; Last: Host mit Lastmittel 4 bis 12 durch drei weitere Agenten. Die Mutation ist ein unbeschütztes `zaehlerX++` (Paketvariable) an der genannten Stelle, so dass nur der Detektor sieht. Nullprobe (Paketvariable ohne Zugriff): **0 von 20**.

| Zähler an | Detektor je Prozess |
|---|---|
| `bereichGueltig`, Zweig `Edm.Int32` | 20 von 20 |
| `bereichGueltig`, vor `ParseInt` | 20 von 20 |
| `verlangeWahrheit` | 20 von 20 |
| `PruefeFilter` (Eintritt) | 20 von 20 |
| `klasseVonArt` | 20 von 20 |
| `zerlegeZahl` | 20 von 20 |
| `zeichenketteOhneAnfuehrung` | 20 von 20 |
| `knoten`, Zweig `FilterNicht` | **0 von 20** |
| `knoten`, Zweig `FilterNull` | **0 von 20** |
| `eigenschaft`, Zweig `Open` | **0 von 20** |
| `eigenschaft`, Name unbekannt | **0 von 20** |
| `eigenschaft`, nicht beobachtet | **0 von 20** |
| `vergleich`, Operator auf Guid nicht unterstützt | **0 von 20** |
| `vergleich`, Typen unverträglich | **0 von 20** |
| `literalTyp`, Zweig `LiteralGUID` | **0 von 20** |
| `FilterTypFehler.Is` | **0 von 20** |
| `FilterTypFehler.Code` | **0 von 20** |
| `pruefeWert`, `ArtRedigiert` (Fehlerpfad) | **0 von 20** |
| `Vergleiche`, unbekannter Operator (Fehlerpfad) | **0 von 20** |

Folge: ein Datenrennen im Code eines nicht geführten Zweiges von `PruefeFilter` oder der Fehlerpfade von `Vergleiche` meldet `make test-race` nicht (V-1). Das ist die Klasse von F-5 (Zeile „Fänger schmaler als Zusage“), nur auf der Achse *Zweig*.

**`r3` (geteiltes Ergebnisfeld, zugewiesen und zurückgegeben; Paketvariable `Wahrheit` in `Vergleiche` geschrieben und gelesen):** Detektor je Prozess mit `-test.count=1`: **20 von 20** (`DATA RACE`); ohne Detektor, `-test.count=1`: **0 von 20** Prozessen rot; ohne Detektor, `-test.count=20`: **0 von 20** Prozessen rot (je 20 Prozesse, Messender Verifier, Last wie oben). Die Raten des Plans (20 von 20, 0 von 20, 0 von 20) stimmen. Das akzeptierte Negativ („ein Wirkungs-Fänger fehlt, das Fenster ist zu kurz; die Zusage gilt für ein Datenrennen als gefangen, für eine Wirkung ohne Datenrennen durch `r4`“) trägt Grund und Folge und bleibt; das Fehlen der Sperre fängt der Detektor in jedem Lauf.

## 6. Randformen der Typen, Gegenprobe mit dem Leser

Probe `TestProbeLeser` (Wegwerfkopie, 51 Eingaben): `X eq <Zahl, Datum, Zeitpunkt, Guid, Zeichenkette, Wahrheitswert>` über `urlparser.NewFilterLeser().LeseFilter`, dann `WertAusLiteral`, `Vergleiche(eq, w, w)` und `PruefeFilter` gegen alle acht Property-Typen. Ergebnis: jedes Literal, das der Leser liefert, ist für `WertAusLiteral` und `Vergleiche` gültig und wahr gegen sich selbst, und `PruefeFilter` meldet genau die Ursache der Klassenrelation. Der Leser normiert, was das Modell als kanonisch verlangt: `1.0`, `1.000`, `00`, `007`, `-0`, `-0.0`, `0.0` ergeben `1`, `1`, `0`, `7`, `0`, `0`, `0`; `-1.50` ergibt `-1.5`; `...00.100Z` ergibt `...00.1Z`; Zeitpunkte mit Versatz sind UTC (`...10:00:00+02:00` zu `...08:00:00Z`); die Guid `...000A` wird kleingeschrieben. Der Leser weist ab (`ODR-E6005` oder `ODR-E6001`): `1.`, `.5`, `+1`, `1e3`, `1d`, `NaN`, `INF`, `0000-01-01`, `2020-02-30`, Stunde `24`, Sekunde `60`, Bruchteil von 13 Ziffern. Damit ist die Zusage „das Modell prüft die Form, der Leser liefert nur Gültiges“ (`SPEC-166` *Ungeprüft*) an der Grenze Leser/Modell belegt; ein Literal, das das Modell als `ErrUngueltig` ablehnte, kommt aus dem Leser nicht.

`Edm.Decimal`, `Edm.DateTimeOffset`, `Edm.Date`, `Edm.Guid`: Randformen der Eingabe sind im Test über den Generator `randformen` geführt (7261 ungültige, 862 gültige Randformen); meine Lücken sind nur V-3 (Stelle mal Großbuchstabe) und V-4 (Länge des Bruchteils).

## 7. Gate-Lauf und Sensoren

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` | **Exit 0** (Zweig `c66329c`, im Worktree) | Lauf unter `flock`, Ausgabe im Scratchpad |
| `make a-check` (Teil von `gates`) | 0 Befunde (Hinweis zu 55 Dateien unter `tools/` ohne Schicht, nicht aus diesem Diff) | `gesamt: 0 Befund(e)` |
| `make docs-check` (Teil von `gates`) | 653 Dateien, 0 Befunde (vor diesem Report) | `d-check: 653 Datei(en) geprüft, 0 Befund(e)` |
| `make groessenmarke` | ok, Messung 3044 von Marke 5000 | `groessenmarke slice-edm-ausdruecke.md: ok, Messung 3044 von Marke 5000` |
| `make zeichen-check` | ok, 1312 Dateien (vor diesem Report) | `zeichen: ok` |
| `make test-race` | ohne Befund im Gate-Lauf (das SQLite-Flake trat nicht auf) | Exit 0 |
| `make test`, `make lint`, `make build`, `make schema-check`, `make migration-check`, `make handbuch-check`, `make lint-profil` | grün | Exit 0 von `make gates` |

**[ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md)**: die vier neuen Dateien des Pakets `model` importieren nur `errors`, `fmt`, `strconv`, `strings` — kein V2- und kein V4-Adapterpaket; die Prüfung `src/adapters/driven/odata/versionstrennung_test.go` betrifft `src/adapters/driven/odata/` und ist unberührt. Der Test unter `tests/integration/filter_instant_test.go` nennt den V4-Leser `urlparser`; das ist nach [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) erlaubt (Schicht `tests` darf Modell und Driven Adapter nennen) und kein V2-Paket. **[ADR-0001](../plan/adr/0001-implementierungssprache-go.md)**: reine Go-Funktionen, kein neuer Importpfad. **[ADR-0003](../plan/adr/0003-lint-profil-golangci.md)**: keine `nolint`-Direktive, `d-check:ignore` oder `t.Skip` im Diff (Suche im Diff, 0 Treffer); `make lint` grün. **[ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)**: `TestGleichzeitig` läuft unter `make test-race`; die Reichweite zeigt Abschnitt 5.

## 8. Kommentare (§3.7, §3.11) und Rohbytes (§3.18)

- **Kommentare im Diff** (Code und Tests, durchgesehen): kein Konjunktiv über verworfene Alternativen, keine Chronik, keine Slice-Kennung als Begründung. Die beiden Kommentare des Reviews (F-7) sind gekürzt (`AlleFilterTypUrsachen liefert die sieben Ursachen`; der Satz „Bereich vor Operator“ steht nicht mehr im Test; Suche im Paket: kein Treffer). **LOW-Hinweis zu §3.11:** der Kopf von `TestGleichzeitig` sagt „rufen je mindestens 1000 mal alle Funktionen von SPEC-166 und PruefeFilter auf“; das gilt je Funktion, nicht je Zweig (V-1).
- **Rohbytes:** gegengelesen mit `LC_ALL=C grep -P` auf die dreizehn Dateien des Diffs (Code, Tests, Plan, Spec): kein BOM (`\xef\xbb\xbf`), kein NBSP (`\xc2\xa0`), kein Nullbreiten-Zeichen (`\xe2\x80[\x80-\x8f]`), kein U+2028, kein U+FFFD, keine Doppelkodierung (`\xc3\x83`), kein NUL; `ü` mit der Byte-Folge `\xc3\xbc` per Festtext-Suche in `filter_typ_fehler.go` (4) und in der Spezifikation vorhanden. Die Escape-Folge aus Backslash und vier Hex-Ziffern kommt in den Dateien des Diffs nicht vor (die Treffer der Suche liegen in älteren Abschnitten der Spezifikation, nicht im Diff). Die Fixtures mit Sonderbytes entstehen aus `[]byte{…}` oder Hilfen (`nbsp()`, `bom()`). Dieser Report wurde auf Rohbytes gegengelesen (siehe Abschnitt 9).

## 9. Nicht aufgelöst

- Die Fix-Runde selbst ist **nicht erneut reviewt** (Vorgabe: je einmal); die Aussagen zu den Mutationen des Implementers (18 plus 13 Detektor-Läufe, Plan §3) habe ich nicht wiederholt, nur gegen die Klasse mit eigenen Elementen aufgelöst.
- Die Raten der Zähler in Abschnitt 5 sind je Zähler **20 Prozesse**; sie belegen 0 von 20 und 20 von 20, nicht die Rate einer Mutation mit Wahrscheinlichkeit dazwischen (keine solche gefunden).
- Nicht gefahren: `make gegenproben` (Stempel), die Closure-Notiz, das Register und die drei Paarungen (Sache des Planners bei der Closure); `Edm.Date` mit Jahr außerhalb 0001 bis 9999 und ungültiges UTF-8 (beides in `SPEC-166` *Ungeprüft*).
- V-3 und V-4 sind die Menge „Zeichen mal Stelle der Form“ und „Länge des Bruchteils“; ich habe sie an je einem bis drei Elementen gefunden und die Menge nicht erschöpfend gemutet (Guid: 32 Hex-Stellen mal 6 Großbuchstaben, 192 Elemente; Bruchteil: 12 Längen). Der Fix muss beide Mengen als Schleife führen, nicht die genannten Elemente.

## 10. Übergabe

Befunde V-1 bis V-5 und V-6 gehen an den Planner zur Entscheidung: Fall oder akzeptiertes Negativ (Grund, Folge) je Befund; V-7 an den Planner (Merge-Reihenfolge). Die Finding-Klassen dieses Laufs für das Lerneintrags-Feld der Closure: **Fänger führt eine Teilmenge der Zweige und Stellen** (V-1, V-3, V-4) · **Zusage zur Folge bei Programmfehlern ohne Fänger** (V-2) · **Prüfgrenze einer unbeschränkten Menge nicht benannt** (V-5).

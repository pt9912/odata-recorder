# Verifikationsbericht: slice-expand-parser — 2026-11-18

**Art:** Verifikation (Bauen wir es richtig? — DoD, Plan, Spezifikation, ADR), Modul 11. Nicht das Review, nicht die Validierung.

**Gegenstand:** Zweig `slice-expand-parser`, `HEAD` `3e1ea109`. Basis des Slice `2ac8a8d3`, Messbasis des Plans `d158068d`; Code `1a9efb6c`, `cb5f0ee0`; Review-Report `docs/reviews/2026-11-18-slice-expand-parser-1.md`; Befundrunde `f97de2f0`, `3e1ea109`.

**Bezug:** `LH-FA-16` (Unterpunkt a), `LH-RB-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md) (Schichten).

**Verträge, gegen die geprüft wurde:** Plan `slice-expand-parser` (§2 DoD, §3 Tabelle *Zusage · Fänger* Nr. 1 bis 30, Mengen, Fix-Runde, §4, §6), `SPEC-262`, `SPEC-263` (nur Abgrenzung), `SPEC-160` Nr. 2 bis 9, `SPEC-162` Nr. 2 und 3, `AGENTS.md` §3.9 bis §3.18.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

## Urteil

**Nicht bestanden.** Ein MEDIUM (V-1), zwei LOW, zwei INFO. Der Code weicht an keiner geprüften Stelle von `SPEC-262` ab; das MEDIUM ist eine Zusage der Spezifikation, deren Menge der Fix von F-1 nicht deckt und für die der Plan keine Prüfgrenze nennt (§3.17).

## 1. DoD-Zeilen gegen Beleg

Die Häkchen setzt der Planner; die Spalte nennt den Beleg und den Stand.

| DoD-Zeile (§2) | Beleg | Stand |
|---|---|---|
| Grammatik, AST, Einhängestelle | `expand.go` (90 Zeilen), `query.go`, `parser.go`, `url_ast.go`, `url_ast_expand.go`; `WithExpand` setzt ein Merkmal, `New` ohne die Option liest `$expand` nicht (`idx == optKeine && nichtUnterstuetzteOption`); `Expand` ist letztes Feld von `Optionen`; Tabelle Nr. 1, 2, 3, 27 | belegt |
| Randformen des Optionstextes, je Zelle Ort und Fänger | Raster §6 (jede Zelle `SPEC-262` Nr. oder „gilt nicht“ mit Grund), Tabelle Nr. 4 bis 24; Generator mit Orakel `exOrakel` (eigene Tabelle) | belegt, bis auf V-1 (Zahl der Teile eines qualifizierten Namens) |
| Parser-Contract-Test | `TestExpandContract` (27 Zeilen der Beispiele, 4 Ressourcenarten); `make a-check` im Gate-Lauf grün | belegt |
| `make gates` grün | selbst gefahren am `HEAD` `3e1ea109`: Exit 0, `zeichen`, `groessenmarke`, a-check, lint, docs-check, test, test-race im Lauf | belegt |
| Review, Report liegt vor | `docs/reviews/2026-11-18-slice-expand-parser-1.md`, 0 HIGH, 3 MEDIUM, 2 LOW, 2 INFO; Fix-Runde `f97de2f0` | belegt |
| Doku-Update | `SPEC-160` (Nr. 2, 3, 6, 8, 9), `SPEC-162` (Nr. 2, 3 a, 3 b), `SPEC-262`, `SPEC-263` im Diff; Stellen gegengelesen | belegt |
| Handbuch/Katalog oder Notiz in §7 | Code bringt keinen Code, Schlüssel, Kommando (`AlleURLUrsachen` bleibt, `src/bootstrap` nur Test); die Notiz „Binary unverändert“ in §7 fehlt | offen (Closure) |
| Closure-Notiz, Register, Paarungen | §7 trägt „Wird bei der Closure gefüllt“ | offen (Closure) |
| Jedes Risiko aus §6 trägt einen Ausgang | alle Zeilen tragen einen; die Zeilen 1 und 2 nennen „offen“, die DoD kennt „weiter offen“ | Wortlaut an die DoD anpassen (V-5) |

## 2. Tabelle *Zusage · Fänger* gegen den Vertrag

Jede Nummer von `SPEC-262` (Nr. 1 bis 9 samt Unterpunkten a bis e, Beispiele, *Ungeprüft*) hat eine Zeile: Nr. 1 → Zeilen 1, 3, 28; Nr. 2 → 5 (`option_name` in 26); Nr. 3 a bis d → 6, 13, 14, 16; Nr. 4 a bis e → 7 bis 12, 15; Nr. 5 → 2, 3, 20, 27; Nr. 6 → 14, 22, 29; Nr. 7 → 3, 26; Nr. 8 → 17, 18, 21, 23; Nr. 9 → 26; *Ungeprüft* → 24, 29, 30. Die akzeptierten Negative (24, 29, 30) nennen Grund und Folge. Die Zusatztabelle (Paare im Wert, über Optionen) hat je Paar einen Fänger (Zeilen 16, 17). Gezählte Mengen nachgerechnet: 359 × 24 × 4 = 34 464; 16 105 × 4 = 64 420; C(33,3) = 5456 Paare, 16 368 und 10 912; 40 × 7 = 280. Befund: die Zahl der Teile eines qualifizierten Namens (Nr. 4 c: „mindestens zwei Namen“) ist unbeschränkt, die Achse führt 1 bis 3 Teile, und keine Prüfgrenze steht im Plan (V-1).

## 3. Nachverifikation der Befundrunde gegen die Klasse

Eigene Mutationen an Elementen, die weder der Review noch der Implementer (`n…`, `f<N>-n…`) nannte. Liste in der Verifier-Arbeit (`v-…`), gefahren mit `sh tools/harness/mutationslauf.sh` im Image `odata-recorder:race`, ohne Netz und Mount, unter dem gemeinsamen Schloss, Klasse `andere`; zwei Läufe mit gleichem Ergebnis. **Last des Hosts 20 bis 32 von 4 CPUs** (andere Läufe, Messender der Lauf des Verifiers): die Ergebnisse der Klasse `andere` sind deterministisch und von der Last nicht abhängig. Der Baum blieb unverändert (`git status` sauber).

| Befund | Menge (aus dem Wortlaut) | Mutation (Verifier) | Ergebnis | roter Test |
|---|---|---|---|---|
| F-1 | jeder Teil jedes Segments eines Pfads, Zahl der Teile eines qualifizierten Namens | `v-f1-teile3`: höchstens drei Teile (`SplitN 3`) | `rot-paket` | nicht `TestExpandSegmentLaenge`; rot im Paket |
| F-1 | dieselbe Menge | `v-f1-teile5`: höchstens fünf Teile (`SplitN 5`), also ein Name mit sechs Teilen `a.b.c.d.e.f` wird `wert` statt `501` | **grün** | keiner im ganzen Paket (V-1) |
| F-2 | Abstand der Duplikate bis 31 | `v-f2-fenster16` (Fenster 16); `v-f2-fenster2` (Fenster 2) | rot, rot | `TestExpandDoppeltesElementAbstand` |
| F-3 | Folgen der Einhängestellen | `v-f3-expand-nur-ohne-filter`: `expand` nur, wenn zum Zeitpunkt von `WithExpand` kein Leser hängt | rot | `TestExpandOptionenFolgen` |
| F-4 | Paare von Verstößen an jeder Stelle 0 bis 5 | `v-f4-stern-i4` (`*` oder Pfad am fünften Element beendet das Lesen); `v-f4-ungueltig-i4` (ungültiges fünftes Element wird `501`) | rot, rot | `TestExpandRangImWert` |
| F-5 | Aufrufe von `ParseURL` auf einem Parser | `v-f5-schreibt-klammer` (ein Aufruf mit `(` schreibt `filter`); `v-f5-schreibt-len13` (ein Aufruf mit genau 13 gültigen Namen schreibt `expand`) | rot; `rot-paket` | `TestExpandParserUnveraendert`; die zweite aus dem Paket, nicht aus dem benannten Test (V-2) |
| F-6, F-7 | Plan-Zeilen | gelesen: Zeile 28 nennt beide Fänger; §4 trägt den Messpunkt | stimmig | — |

F-2, F-3 und F-4 sind gegen die Klasse gedeckt. F-1 ist es für die Zeichenmenge und die Länge je Teil (die Mutationen des Reviews und alle `n1` bis `n6` des Implementers bleiben außen vor), nicht für die Zahl der Teile (V-1). F-5 ist gedeckt, soweit der Plan es als Prüfgrenze nennt (V-2).

## 4. Plan gegen Code, ADR, Adressen, Abgrenzung

- **Plan gegen Code.** `elementKlasse`, `expandPfad`, `expandSegment`, `leseExpand` folgen der Folge von `SPEC-262` Nr. 3 (Klammer, Zahl, Elemente, `ungueltig` vor `nicht_unterstuetzt`, erstes der Klasse); doppelte Elemente bytegleich; Ort `query/<n>` über `fehler(…, ort)`; kein Teil der Eingabe im Fehler. Produktcode von Hand nachgemessen: 127 Zeilen (`expand.go`, `parser.go`, `query.go`, `url_ast.go`, `url_ast_expand.go`), wie §4 sagt.
- **Entscheidungen.** [ADR-0001](../plan/adr/0001-implementierungssprache-go.md): Go, kein neuer Fremdcode, Docker-only (alle Läufe über `make` und das Mutationswerkzeug). [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md): der Parser importiert nur `model`; `make a-check` grün; kein Dienst, kein Modell des Dienstes.
- **Abgrenzung.** Kein Validator im Diff (`SPEC-263` ist Vertrag in der Spezifikation, Code liegt bei `slice-expand-validator`); `src/bootstrap` ändert nur eine Test-Datei, `wurzel.go` bleibt, das Binary liest `$expand` weiter nicht; `ODV4-E-03` c und `abnahme_matrix_test.go` unverändert; `$expand` im V2-Parser unberührt.
- **Adressen (`AGENTS.md` §3.13).** `slice-expand-ausfuehrung` führt die Verdrahtung, den Aufruf des Validators, die Änderung von `ODV4-E-03` c und `TestSatzVersion4OhneExpand` in §1 und DoD mit dem Geber `slice-expand-parser`; `slice-expand-validator` führt den Schnitt in §1 mit dem Geber. Beide Nehmer liegen in `open/`, keiner schließt die Sendung aus. Die Planner-Adresse (vier gegen fünf Slices der Welle) steht in §6 mit Ausgang *offen*.
- **Kommentare (§3.7, §3.11).** Die Köpfe der Tests sagen, was die Schleifen führen (Stichprobe: Mengen von `TestExpandSegmentZeichen`, `…Struktur`, `…Abstand`); die Achse „Zahl der Teile 1 bis 3“ sagt nicht, dass die Zahl der Teile nur begrenzt geprüft ist (V-1).
- **Rohbytes (§3.18).** Gegengelesen mit `LC_ALL=C grep -nP` auf BOM, NBSP und Backslash-u-Folgen: `expand*.go` und der Plan: keine Fundstelle. Dieser Bericht wurde nach dem Schreiben mit demselben Muster gegengelesen.

## 5. Größenmarke (§4)

`make groessenmarke` im Gate-Lauf: `ok, Messung 1723 von Marke 3500`; `Messpunkt … messung=1723` gleich dem Eintrag in §4. Von Hand: `git diff --numstat d158068d..HEAD` über die Pfade des Blocks gibt 1723. Marke, Formel (`schaetzung=2555`, `marke=3500`) und Basis stehen vor dem Code im Block. Auslastung 0,49 unter 0,60, das meldet der Planner in der Closure dem Architect (`SPEC-128` Nr. 9). Die Messung liegt in der Spanne der Erwartung (1500 bis 2400); der Wert 1298 des Reviews war die Messung vor der Fix-Runde.

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Klasse |
|---|---|---|---|---|---|
| V-1 | MEDIUM | Die Zahl der Teile eines qualifizierten Namens im Segment eines Pfads ist unbeschränkt (`SPEC-262` Nr. 4 c: „mindestens zwei Namen“), geführt sind 1 bis 3 Teile; ein Name mit sechs Teilen (`SplitN` mit 5, `v-f1-teile5`) bleibt im ganzen Paket grün, ein Name mit vier oder fünf Teilen (`v-f1-teile3`) wird nur aus einem anderen Test als dem benannten rot. Der Plan nennt keine Prüfgrenze der Zahl der Teile (§3, *Prüfgrenzen mit Folge im Betrieb*, Zeile 12). Folge: ein Umbau, der Teile jenseits der geführten Zahl anders behandelt, liefert `400` statt `501` oder umgekehrt. | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 4 c | `expand_segment_test.go`, `expand_grenzen_test.go`; Plan §3 Zeile 12 und Fix-Runde F-1 | Menge unbeschränkt, Prüfgrenze nicht als akzeptiertes Negativ benannt |
| V-2 | LOW | `TestExpandParserUnveraendert` führt keine gültige Liste von 4 bis 31 Namen; ein Schreiben, das erst bei einer solchen Liste greift (`v-f5-schreibt-len13`), ist `rot-paket`, nicht aus dem benannten Test. Der Plan nennt die Grenze („außerhalb von `korpus()`, `exQueries`, den Folgen bis Länge 3 und den vier langen Eingaben“), aber nicht, dass gültige Listen mittlerer Länge zu den Außenstehenden gehören; die Listen der Länge 1 bis 32 aus `TestExpandListe` wären eine billige zusätzliche Eingabe. | `AGENTS.md` §3.17 | `expand_test.go` · `TestExpandParserUnveraendert`; Plan Fix-Runde F-5 | Prüfgrenze benannt, aber ohne die gültigen Listen |
| V-3 | LOW | Die Raten der Zeile 23 (Last „7,1 bis 9,6“ und „4,4 bis 4,9“ von 4 CPUs im Plan) wurden hier nicht nachgemessen; die Läufe des Verifiers fanden 20 bis 32 von 4 CPUs (mehrere Läufe auf dem Host). Die Klasse `andere` ist von der Last unabhängig. | `AGENTS.md` §3.10 | Plan §3 *Raten der Nebenläufigkeit* | nicht nachgemessen |
| V-4 | INFO | §7 ist leer („Wird bei der Closure gefüllt“): die DoD-Zeilen zu Katalog-Notiz, Closure-Notiz, Register und Paarungen sind offen, das ist der Stand vor der Closure. | Plan §2, §7 | Plan §7 | — |
| V-5 | INFO | §6 Zeilen 1 und 2 tragen „Ausgang: offen“; die DoD verlangt *eingetreten*, *entfallen* oder *weiter offen*. | Plan §2, §6 | Plan §6 | — |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 2 |
| INFO | 2 |

## Übergabe an den Planner

Merge-blockierend ist V-1: die Zahl der Teile des qualifizierten Namens bekommt entweder eine Schleife (Teile 1 bis 8 oder mehr, je an jeder Stelle des Pfads, mit Orakel) oder eine Prüfgrenze mit Grund und Folge im Plan; der Bericht des Fixes nennt dann eine Mutation an einem Element, das dieser Bericht nicht nannte (etwa die Zahl der Teile nach dem Punkt hinter einem Typ-Cast). V-2 bis V-5 sind ohne Auswirkung auf das Urteil und gehen in die Closure. Die Finding-Klasse dieses Laufs: *Menge unbeschränkt, Prüfgrenze nicht benannt (Zahl der Teile)*.

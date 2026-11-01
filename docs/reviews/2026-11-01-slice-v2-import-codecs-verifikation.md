# Verifikationsbericht: slice-v2-import-codecs — 2026-11-01

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan und DoD, nicht gegen den Diff (Reviewer) und nicht gegen den Bedarf (Validator).

**Gegenstand:** `slice-v2-import-codecs` (Welle `welle-odata-v2`), Zweig `slice-v2-import-codecs`, geprüfter Stand **HEAD `e1a1de0fe8a611dad71fbcbb9e2199f0012cc1a8`**, Basis `b5ac4a5`. Bezug: `LH-FA-17`, `LH-FA-18`, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md).

**Geprüft gegen:** DoD und §3 (Tabelle Z1 bis Z9, Absatz *Mutationen und Raten*, Fix-Runde), §4 (Block, Marke), §6 (Raster, Zusatztabelle, Risiken) des Plans; `SPEC-127` Nr. 1 (nach L1), `SPEC-175` Nr. 3, 4, 6; [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md); `AGENTS.md` §3.10, §3.12, §3.15 bis §3.17; Review `docs/reviews/2026-11-01-slice-v2-import-codecs-review.md` (F-1 bis F-7).

**Methode.** Der Beleg des Implementers (Gates grün, Messung 758, Raten) ist nicht übernommen. Selbst gefahren: `make gates` (alle Gates, ein Lauf), `go test -count=20` über das Paket, 30 Mutanten auf Kopien im Scratchpad (`.../scratchpad/ver/`, Quelle `git archive HEAD`; im Repo liegt nichts außer diesem Bericht), Z7-Raten mit eigenen Bauarten (30 bzw. 20 Prozesse je Zelle, `-test.count=1`). Docker nur mit `--network none --memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, jeweils ein Container. Die Mutanten sind aus der Klasse der Zusage gebildet (Aufrufstellen mal Dienste mal Argumente, Ausgänge der Rangfolge), nicht aus den Mutationen des Plans oder des Reviews; die Mutanten des Reviews (F-1, F-2) sind mit anderen Elementen der Klasse gegengeprüft.

---

## Zahlen (selbst gemessen)

| Größe | Wert |
|---|---|
| `make gates` auf HEAD `e1a1de0` | exit 0 (a-check `gesamt: 0 Befund(e)`, `pruefung test`, `pruefung race`, lint, groessenmarke, docs-check, zeichen-check, schema-/migration-check) |
| `make groessenmarke` | `slice-v2-import-codecs.md: ok, Messung 758 von Marke 1500`; Messpunkt `zeilen=9 nebenlaeufig=1 fakes=1 produktcode=120 schaetzung=1045 marke=1500 messung=758` |
| `git diff --numstat b5ac4a5..HEAD` über die Pfade des Blocks | 16 + 43 + 7 + 692 = **758** (`codec_satz.go` 16/0, `importer.go` 34/9, `metadaten.go` 4/3, `satz_test.go` 692/0) |
| `git diff --numstat b5ac4a5..HEAD -- '*_test.go'` | allein `satz_test.go` (692 hinzu, 0 entfernt); `helfer_test.go` und jede vorhandene Testdatei unberührt |
| `go test -count=20 ./src/hexagon/services/baselineimport/` | ok (55,8 s) |
| Mutanten dieser Prüfung | 30 von 30 rot (Tabelle unten), 0 grün |
| Z7 ungemutiert | 0 von 30 rot mit `-race`, 0 von 30 ohne |
| Arbeitsbaum nach allen Läufen | sauber (`git status --short` leer vor diesem Bericht) |

## Plan-vs-Code-Diff

| Plan sagt | Code sagt | Urteil |
|---|---|---|
| `outbound.CodecSatz{Version, Metadaten, Daten, NurLesend}` im Port-Paket, `URL` und `Schreiber` nicht in diesem Slice | `codec_satz.go`: genau diese vier Felder, keine Importe | konform |
| `Dienst.Codecs *outbound.CodecSatz`; Prüfung in `New` (`Metadaten`/`Daten` `nil` bei `Codecs` ungleich `nil` ist `ErrUngueltig`); `nil` ist gültig | `importer.go`: Feld, Prüfung in `normiereDienste`, `nil` unberührt | konform |
| Wahl je Aufrufstelle (`ParseMetadata`, `ParseCollection`, `ParseEntity`) aus dem Dienst, `Importer` bleibt bei einem Feld, kein Zwischenspeicher | `metadatenCodec`/`datenCodec` lesen `dienstVon(serviceID).Codecs` je Aufruf; `Importer{k Konfig}` unverändert | konform |
| Wahl hinter Nr. 4 b (Risiko 1) | `Import` prüft die `ServiceID` vor `wiederhole`; die Wahl steht in `liesDokument`/`liesAntwort` | konform |
| `normiereDienste` nimmt `Codecs` mit und gibt eine Kopie zurück (Z6, Risiko 3) | `out = append(out, Dienst{..., Codecs: d.Codecs})`, `out` ist eine neue Liste | konform |
| Der Dienst liest allein `Metadaten` und `Daten` | kein Zugriff auf `Codecs.Version` oder `Codecs.NurLesend` im Produktcode (`grep` über `src/`, 0 Treffer) | konform |
| Kommentar der Konfig nachziehen (Risiko 7) | Kopf von `Dienst` und `Konfig` ergänzt | konform |
| V4-Tests textgleich, neue Tests in neuer Datei | siehe Zahlen | konform |
| Hexagon: `CodecSatz` im Port-Paket, kein neuer Import in `baselineimport` | `make a-check`: 0 Befunde; `git diff` zeigt keinen neuen Import | konform |
| Spec-Nachzug L1 in `SPEC-127` Nr. 1 vor dem Code | `d8157e7`, Historienzeile 2026-11-01; Code folgt dem Wortlaut (nur `Metadaten`/`Daten`, geteilter Satz zulässig, Codecs der Konfig bleiben Pflicht) | konform |
| Test-Helfer heißt `fakeSatz` (§3 *Fakes (Architect)*, Risiko 5) | Datei und Plan-Absatz *Mutationen und Raten* nennen `satzFake` | Abweichung im Namen, siehe F-V1 |

## Tabelle Z1 bis Z9 gegen den Vertrag (jede Zeile eigene Mutation, gegen die Klasse)

Die Menge ist aus dem Wortlaut der Zusage gebildet. „rot“ heißt: der Fänger der Zeile wurde aus dem Grund der Zeile rot (Meldung gelesen, nicht nur der Name des Tests); wo das Paket bei einer Mutation panikt und so die neuen Tests nicht erreicht, ist mit `-run` auf die neuen Tests allein wiederholt.

| Zeile | Mutation (Element der Klasse, vom Plan oder Review nicht genannt, soweit nicht anders vermerkt) | Ergebnis |
|---|---|---|
| Z1 (Argument mal Aufrufstelle mal Dienst) | `grenze` von `ParseCollection` plus 1 bei Dienst mit Satz · `grenze` von `ParseMetadata` minus 1 · `entitySet` von `ParseEntity` um `x` verlängert bei Dienst mit Satz · `ctx` von `ParseEntity` `Background()` bei Dienst mit Satz · Nullmodell an `ParseCollection` bei Dienst mit Satz · Strom von `ParseEntity` mit `io.MultiReader` gehüllt bei Dienst mit Satz · Daten-Codec kommt vom letzten Dienst der Liste · Metadatenleser kommt von `Dienste[1]` · `ParseEntity` allein nimmt den Codec der Konfig (9 Mutanten) | 9 von 9 rot, `TestSatzWahlJeAufrufstelle` (bei `Dienste[1]` mit `-run` auf die neuen Tests: das Paket panikt sonst im Bestand) |
| Z2 (Fehler an jeder Aufrufstelle, danach kein Port) | nach dem Fehler von `ParseCollection` ruft der Dienst `ParseEntity` der Konfig · nach dem Fehler von `ParseEntity` ruft er `ParseCollection` von D1 · nach dem Fehler von `ParseMetadata` ruft er `ParseMetadata` von D2 (3 Mutanten; Klasse der Zähler aller neun Aufrufstellen, F-1) | 3 von 3 rot, `TestSatzFehlerJeAufrufstelle` mit der Meldung des fremden Zählers (`ParseMetadata von D2 1-mal, erwartet 0`, `ParseCollection von D1 1-mal, erwartet 0`) |
| Z3 (Satz teil-`nil` an jeder Stelle) | Prüfung nur für die ersten zwei Dienste · Prüfung lässt den mittleren Dienst aus · Prüfung nur für den letzten | 3 von 3 rot, `TestNewSatzTeilNil` |
| Z4 (Codecs der Konfig bleiben Pflicht) | `DatenCodec` der Konfig entfällt, sobald irgendein Dienst einen Satz trägt (gemischt) · `MetadatenCodec` der Konfig entfällt, sobald jeder Dienst einen Satz trägt | 2 von 2 rot, `TestNewKonfigCodecsPflicht` (der erste allein dieser Test) |
| Z5 (übrige Felder ohne Wirkung) | `New` weist `Version` `2.0` ab · `New` weist `NurLesend` ab · Daten-Codec fällt bei `Version` `x` auf die Konfig · Metadatenleser fällt auf die Konfig, wenn `NurLesend` falsch ist | 4 von 4 rot, `TestSatzUebrigeFelder` |
| Z6 (Normierung behält den Satz, doppelte `ID`) | `normiereDienste` setzt `Codecs` auf `nil` · die doppelte `ID` wird nur bei gleichem Satz erkannt | 2 von 2 rot, `TestNormierungBehaeltSatz` (der erste mit weiteren) |
| Z7 (gleichzeitige Imports) | Raten unten | konform; Negativ siehe unten |
| Z8 (kein Codec vor dem Lesen, Wahl hinter der Dienstprüfung; F-2) | `waehleDokument`-Fehler ruft `ParseCollection` des Dienstes · `waehleDokument`-Fehler ruft den Metadatenleser der Konfig · Fehler `Session eines anderen Service` ruft `ParseMetadata` · Fehler von `Imports.Get` ruft `ParseMetadata` · ungültige Anfrage ruft den Metadatenleser der Konfig · `Get` ruft einen Codec · die Wahl des Satzes steht vor der Dienstprüfung (Panik) | 7 von 7 rot, `TestSatzKeinAufrufVorDemLesen`, jede mit dem Zähler des eigenen Ausgangs (`$metadata fehlt D0: ParseCollection von D0 1-mal gerufen`, `Session eines anderen Service D0: ParseMetadata von D0 1-mal gerufen`) |
| Z9 (`Codecs` `nil` wie vor dem Slice) | keine eigene Mutation: die Suite des Bestands läuft unverändert in `make gates` und in `-count=20` grün; die Mutationen (a) und (b) der Zeile nennt der Plan, der Reviewer hat sie rot gesehen (100 Tests bzw. Panik) | konform, Mutation nicht von mir wiederholt |

Gesamt 30 von 30 Mutanten rot, 0 grün. Die zwei grünen des Reviews (F-1, F-2) sind durch den Fix-Commit `3bef041` gedeckt, und zwar über die Klasse: die Zähler laufen über alle neun Aufrufstellen (3 mal 3) in 18 Fällen, die Ausgänge über alle 13 Ausgänge mal drei Dienste plus `Get` (42 Fälle). Meine Mutanten treffen Elemente, die weder der Befund noch der Plan nennen (`ParseCollection`-Verwurf mit `ParseEntity` der Konfig, `Session eines anderen Service`, `Imports.Get`-Fehler, ungültige Anfrage, `Get`).

## Z7, Raten selbst gemessen (§3.10)

Messender: der Verifier. Kompiliertes Testprogramm, `-test.run '^TestSatzGleichzeitig$' -test.count=1`, je Zelle 30 Prozesse (20 bei a und c), nacheinander in einem Container mit 4 CPUs ohne Netz; mit `-race` Image `odata-recorder:race` (cgo), ohne `-race` Image `odata-recorder:test` (cgo aus). Die Mutationen sind eigene Bauten nach der Beschreibung des Plans (Buchstaben wie im Plan): (a) Wahl des Antwort-Codecs vor `ParseMetadata` in einer Paketvariablen, danach gelesen; (c) Modell nach `ParseMetadata` in einer Paketvariablen; (d) Paketvariable zu Beginn von `liesAntwort` geschrieben, nach `Blobs.Open` gelesen; (e) wie (d) mit `atomic.Value`; (h) `atomic.Value` unmittelbar vor dem Lesen geschrieben, kein Portaufruf dazwischen; (i) dasselbe mit einer Paketvariablen.

| Mutation | mit `-race`: rot (davon `DATA RACE`) | ohne `-race`: rot | Plan (Implementer), mit / ohne |
|---|---|---|---|
| ungemutiert | 0 von 30 | 0 von 30 | 0 von 30 beide |
| (a) | 20 von 20 (20) | 20 von 20 | 30 (30) / 30 |
| (c) | 20 von 20 (20) | 20 von 20 | 30 (30) / 30 |
| (d) | 30 von 30 (28) | 30 von 30 | 30 (14) / 30 |
| (e) | 30 von 30 (0) | 30 von 30 | 30 (0) / 30 |
| (h) | 3 von 30 (0) | **0 von 30** | 2 (0) / 0 |
| (i) | 29 von 30 (29) | **0 von 30** | 10 (10) / 0 |

Die Wirkungs-Fänger (a), (c), (d), (e) sind in jedem Lauf rot, mit und ohne Detektor, wie §3.10 für Nebenläufigkeit verlangt. Die zwei Fenster ohne Aufruf eines Ports mit Haltepunkt ((h), (i)) sind ohne `-race` nie rot, wie als akzeptiertes Negativ mit Grund und Folge im Plan steht (Grund: kein Test kann zwischen zwei Anweisungen des Dienstes anhalten; Folge: ein atomarer Zwischenspeicher der Wahl in einem solchen Fenster bleibt unentdeckt, und der Plan benennt, dass allein der Review der Wahl in `metadatenCodec` und `datenCodec` ihn fängt). Das Negativ ist bekannt und akzeptiert; die Messung bestätigt es, ohne es zu erweitern. Abweichung der Zahl, nicht der Aussage: siehe F-V2.

## Raster §6 (jede Zelle mit Ort, nichts im Code entschieden)

Alle 24 Zellen der drei Zeilen (`New`, `Import`, `Get`) nennen eine Nummer von `SPEC-127`/`SPEC-175`/`SPEC-171` oder „gilt nicht“ mit Grund; die genannten Nummern habe ich am Wortlaut der Spezifikation gegengelesen (Nr. 1, 2, 4 b, 4 d, 5, 6, 7 f, 9, 16, 17, 18). Die Zusatztabelle (Paar teil-`nil` und leere `ID`; Dienst mit Satz neben Dienst ohne; Satz-Codec gleich Konfig-Codec; Satz nach `New` geändert) trägt Entscheidung und Fänger oder ein akzeptiertes Negativ. Im Code entschieden ist keine Zelle: die Reihenfolge der Prüfungen in `normiereDienste` (Satz vor URL) ist dort unter „welcher Verstoß gemeldet wird, ist nicht zugesagt“ gedeckt (Nr. 3, Nr. 1).

## Fixes des Reviews gegen die Klasse (F-1 bis F-7)

| Befund | Klasse (Wortlaut der Zusage) | Elemente im Fänger | Mutation an einem nicht genannten Element (eigene) | Urteil |
|---|---|---|---|---|
| F-1 (MEDIUM, Z2) | „danach kein weiterer Port“: Aufrufstelle des Fehlers mal Dienst mal Fehlerart, Zähler aller 9 Aufrufstellen | 18 Fälle mit je 9 Zählern (Schleife über `satzOps()`, drei Wahlen, zwei Arten) | `ParseEntity` der Konfig nach Fehler von `ParseCollection`; `ParseCollection` von D1 nach Fehler von `ParseEntity`; `ParseMetadata` von D2 nach Fehler von `ParseMetadata` | geschlossen, 3 von 3 rot |
| F-2 (MEDIUM, Z8) | „vor dem Lesen endet“: jeder Ausgang der Rangfolge vor Nr. 7 f mal jeder Dienst, dazu `Get` | 13 Ausgänge mal 3 Dienste plus 3 Fälle `Get` (Menge 42, im Test gezählt) | `waehleDokument` ruft `ParseCollection`; `Session eines anderen Service`; `Imports.Get`-Fehler; ungültige Anfrage; `Get` | geschlossen, 7 von 7 rot |
| F-3 (MEDIUM, Z7) | „in jeder Verschränkung“: jedes Fenster zwischen zwei Aufrufen von Ports mit Haltepunkt | 9 Schranken im Gleichschritt, 3 Codec-Schranken; Fenster ohne Aufruf als benanntes Negativ mit Grund und Folge | (d), (e) selbst gemessen, siehe Z7 | geschlossen (Klasse gedeckt, Rest als Negativ benannt) |
| F-4 (LOW) | Kopf von `TestSatzFehlerJeAufrufstelle` | Kopf nennt „die drei Aufrufstellen jedes der drei Dienste“; der Test zählt `satzOps()` mal alle Dienste | — | geschlossen |
| F-5 (LOW) | Parameter von `satzDienste` | Kopf nennt `ersatz[k]`, Parameter heißt `ersatz` | — | geschlossen |
| F-6 (INFO) | „Satz-Zweig“ im Plan | `grep` über `*.md` und `*.go`: nur die Zeile F-6 des Plans und der Review selbst | — | geschlossen (Selbstbezug, siehe F-V3) |
| F-7 (INFO) | Messpunkt in §4 | `Messpunkt`-Zeile steht in §4, Zahl 758 gleich meiner Messung | — | geschlossen |

## DoD

| Punkt | Zustand |
|---|---|
| 1. `Dienst.Codecs`, Prüfung, Wahl, V4-Tests textgleich grün | erfüllt (siehe Plan-vs-Code und Zahlen) |
| 2. Fänger: Fake-Satz mit zwei Diensten, Schleife über Aufrufstellen mal Dienste, `nil`, Teil-`nil` | erfüllt (drei Dienste, D0 `nil`; Mengen im Test gezählt) |
| 3. Handoff nach §3.9, Tabelle *Zusage · Fänger*, Absatz der Mutationen | erfüllt (`d3600a3` trägt Code und Plan; Fix-Runde `3bef041` trägt Test und Plan) |
| 4. `make gates` grün | erfüllt (eigener Lauf, exit 0) |
| 5. Review durchgeführt | erfüllt (`docs/reviews/2026-11-01-slice-v2-import-codecs-review.md`) |
| 6 bis 10 (Doku-Update, Closure-Notiz, Register, Ausgänge in §6, drei Paarungen) | offen, Sache der Closure durch den Planner; §6 trägt die Ausgänge der Risiken 1 bis 7 noch als *offen (bei Closure)*, §7 ist leer — kein Verstoß vor der Closure |

## Findings

| ID | Kategorie | Befund | Pfad |
|---|---|---|---|
| F-V1 | LOW | Der Plan nennt den Test-Helfer zweimal `fakeSatz` (§3 *Fakes (Architect)*, §6 Risiko 5) und einmal `satzFake` (§3 *Mutationen und Raten*); der Code heißt `satzFake`. `AGENTS.md` §3.9 verlangt, dass der Plan der Korrektur folgt. Keine Wirkung auf Gate oder Test. | `slice-v2-import-codecs.md` §3 *Fakes (Architect)*, §6 Risiko 5; `satz_test.go` · `type satzFake struct` |
| F-V2 | INFO | Vier Raten im Plan lassen sich mit der Bauart, die ich aus der Beschreibung gebildet habe, nicht zahlengleich reproduzieren: (a) und (c) mit `-race` melden bei mir in 20 von 20 Prozessen `DATA RACE` (Plan: 30 von 30), (d) 28 von 30 (Plan: 14), (i) 29 von 30 (Plan: 10). Die Aussage des Plans bleibt richtig (der Detektor meldet nur mit Wahrscheinlichkeit, ohne `-race` nie); die Zahlen hängen von der Bauart der Mutation ab, nicht vom Code. Keine Aktion; die Wirkungs-Fänger sind davon nicht berührt (in jedem Lauf rot). | Plan §3 *Raten von Z7* |
| F-V3 | INFO | Die Zeile F-6 der Tabelle *Fix-Runde* sagt, `Satz-Zweig` komme nicht mehr vor; die Zeile enthält den Ausdruck selbst. Ein Selbstbezug ohne Wirkung. | Plan §3 *Fix-Runde*, Zeile F-6 |
| F-V4 | INFO | Die Schätzung (1045) liegt über der Messung (758 im Stand `e1a1de0`); bekannt und akzeptiert, hier nicht erneut bewertet. Marke 1500 gehalten (758). | Plan §4 *Messung* |

0 HIGH, 0 MEDIUM, 1 LOW, 3 INFO.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Plan-vs-Code-Diff (Pfade `ports/outbound`, `baselineimport`) | geprüft, ohne Befund außer F-V1 (Name im Plan) |
| Tabelle Z1 bis Z9 gegen den Vertrag | geprüft, ohne Befund: 30 von 30 Mutanten an Elementen außerhalb der genannten rot, aus dem Grund der Zeile |
| Raster §6 und Zusatztabelle | geprüft, ohne Befund: keine leere Zelle, keine im Code entschiedene |
| Fixes F-1 bis F-7 gegen die Klasse | geprüft, ohne Befund: jede Menge als Schleife, je eine Mutation an einem nicht genannten Element rot |
| V4-Tests textgleich (Z9) | geprüft, ohne Befund: `--numstat` nennt nur `satz_test.go`, Suite grün in `make gates` und `-count=20` |
| Hexagon (`make a-check`, Richtung `ports/outbound` ← `baselineimport`) | geprüft, ohne Befund: 0 Befunde |
| `make gates` | geprüft, ohne Befund: exit 0 |
| Größenmarke | geprüft, ohne Befund: 758 von 1500, `numstat` und Gate gleich |
| Rohbytes (§3.18) | geprüft, ohne Befund: `LC_ALL=C grep -P` über die hinzugefügten Zeilen des Diffs `b5ac4a5..HEAD` (Folge Backslash-`u`-vier-Hex, BOM, NBSP, Nullbreiten- und Steuerzeichen): 0 Treffer; die 3 Treffer im Bestand der beiden Pakete liegen in Zeilen, die der Slice nicht berührt; dieser Bericht ist nach dem Schreiben auf Rohbytes zurückgelesen (0 Treffer) |
| Z7-Raten (stichprobenartig selbst) | geprüft, ohne Befund: Wirkungs-Fänger 100 %, Negativ bestätigt; Abweichung der Zahlen nur als F-V2 |
| [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) und `SPEC-175` Nr. 3, 4, 6 | geprüft, ohne Befund: der Dienst verzweigt nicht nach der Version, der Satz ist je Dienst, kein Zustand im Importer |
| Suppression (§3.2) | geprüft, ohne Befund: keine `//nolint`-Direktive im Diff |

## Ergebnis

Konform. Der Slice liefert die Zusage der DoD-Punkte 1 bis 5; jede Zeile Z1 bis Z9 hat einen Fänger, der aus dem Grund der Zeile rot wird, und die Fixes des Reviews decken ihre Klassen. Offen bleibt allein, was die Closure schuldet (DoD 6 bis 10, §6 Ausgänge, §7). An den Planner: F-V1 (Name `fakeSatz` im Plan) im Zuge der Closure nachziehen.

# Review-Report: slice-v2-import-codecs — 2026-11-01

**Review-Art:** Code — geprüft gegen Plan (`slice-v2-import-codecs` §1, §3 mit der Tabelle *Zusage · Fänger* Z1 bis Z9 und dem Absatz *Mutationen und Raten*, §4 mit Block, Marke 1500 und Basis `b5ac4a5`, §6 mit Raster und Zusatztabelle), gegen `SPEC-127` Nr. 1 (nach dem Nachzug L1) und die Nachbarn (`SPEC-175` Nr. 3, 4 und 6, `SPEC-171` Nr. 2), gegen [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) und die Hard Rules (`AGENTS.md` §3.1 bis §3.18). Nicht geprüft wird gegen die DoD; das ist die Frage des Verifiers.

**Gegenstand:** `slice-v2-import-codecs` (Welle `welle-odata-v2`), `git diff b5ac4a5..8f5e466` (HEAD `8f5e466`). Pfade: `src/hexagon/ports/outbound/codec_satz.go`, `src/hexagon/services/baselineimport/` (`importer.go`, `metadaten.go`, `satz_test.go`), daneben Plan und `spec/spezifikation.md` (Klarstellung L1 des Planners).

**Skill:** `.harness/skills/reviewer.md` @ HEAD `8f5e466` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-01

**Methode.** Der Bericht des Implementers lag nicht vor; die Prüfung rekonstruiert aus Plan, Spezifikation, ADR und Diff. Keine Mutation, Rate oder Einschätzung des Plans ist übernommen. Alle Mutanten laufen auf Kopien im Scratchpad (`.../scratchpad/rvm/`, Quelle `rv-ic` = Kopie von HEAD `8f5e466`; im Repo liegt nichts außer diesem Report), `--network none --memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, je Mutant ein Lauf `go test -count=1 -v ./src/hexagon/services/baselineimport/` gegen **die letzte Testfassung** (`satz_test.go` in HEAD). Mutanten, die nicht übersetzen (5, danach korrigiert neu gefahren), zählen nicht. Die Mengen sind aus dem Wortlaut der Zusage gebildet (Aufrufstellen mal Dienste mal Argumente), nicht aus dem Plan; wo die Suite eine Achse nur mit einer Ausprägung führt, steht ein Mutant an einer anderen.

**Eingangs-Kontext:**

- Slice-Plan `slice-v2-import-codecs` in `8f5e466` (§1, §3, §4, §6)
- `SPEC-127` Nr. 1 bis 4, 7 f, 9, 16 bis 18; `SPEC-175` Nr. 3, 4, 6; `SPEC-171` Nr. 2; `SPEC-128`
- [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md)
- [LH-FA-17, LH-FA-18](../../spec/lastenheft.md)
- `AGENTS.md` (Hard Rules); Reviewer-Skill mit den Zeilen *Fänger führt nur eine Form der Eingabe*, *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*, *Akzeptiertes Negativ ohne Grund und Folge*

---

## Antworten auf die Fragen des Auftrags

**(1) Z1 bis Z9 und Raster, gegen die Klasse (§3.17).** 84 gültige Produktcode-Mutanten: **81 rot, 3 grün** (F-1, F-2). Zeile für Zeile; „rot“ heißt: der Fänger der Zeile wurde aus dem Grund der Zeile rot.

| Zeile | Eigene Mutanten (Menge nach Wortlaut der Zusage) | Ergebnis |
|---|---|---|
| Z1 | 24: Argument {`ctx`, `grenze`, Modell, `entitySet`, Strom} an jeder der drei Aufrufstellen verändert (`Background()`, `grenze` ±1 und 0, Nullmodell, `entitySet` verlängert und leer, `io.MultiReader` als Hülle; je Stelle, soweit das Argument existiert); Codec der Konfig an jeder Aufrufstelle; Satz von `Dienste[0]`, vom letzten Dienst, D1 statt D2, D0 statt D1 (je Metadaten und Daten); Entity-Spur allein, Collection-Spur allein nimmt den Codec der Konfig | 24 von 24 rot (`TestSatzWahlJeAufrufstelle`, bei vielen auch der Bestand) |
| Z2 | 10: Fehler an jeder Aufrufstelle verworfen (3), Kontextfehler wird `ErrUngueltig` an jeder Aufrufstelle (3), nur `DeadlineExceeded` wird `ErrUngueltig` (Entity), `Close` doppelt, `Close` bei Fehler ausgelassen, fremder Fehler statt des Verstoßes (Collection) | 10 von 10 rot; **aber** F-1: ein Aufruf eines weiteren Codecs nach dem Fehler bleibt grün |
| Z3 | 10: Prüfung nur `Daten`, nur `Metadaten`, `&&`, nur erster, nur letzter, nur mittlerer, alle außer erster, außer letzter, außer mittlerer, ganz entfernt | 10 von 10 rot (`TestNewSatzTeilNil`) |
| Z4 | 6: Konfig-Codec entfällt bei allen Diensten mit Satz, bei gemischt, je allein für `MetadatenCodec` oder `DatenCodec` (4 Formen) | 6 von 6 rot (`TestNewKonfigCodecsPflicht`) |
| Z5 | 12: `Version` leer, `x`, `2.0`, `4.0` je abgewiesen oder vergleichend gegen `Modell.Version()`; `NurLesend` in `New` abgewiesen, in `Import` abgewiesen, wählt die Codecs der Konfig; `Version` `2.0` wählt Konfig (Metadaten), `Version` leer wählt Konfig (Daten); geteilter Satz abgewiesen, geteilter Satz im zweiten Dienst ohne Satz | 12 von 12 rot (`TestSatzUebrigeFelder`) |
| Z6 | 4: `Codecs` verloren, `in` statt Kopie, `==` am ganzen `Dienst`, Doppelte nur bei gleichem Satz | 4 von 4 rot (`TestNormierungBehaeltSatz`) |
| Z7 | 15 Varianten, je 30 Prozesse, siehe (2) | Mutationen der Zeile rot; Fenster ohne Haltepunkt nicht (F-3) |
| Z8 | 8 (Wahl vor Anfrageprüfung, vor Dienstprüfung für Metadaten und Daten, Codec vor Wiederholung (Konfig und Satz), vor Session (Metadaten und Daten), `Get` ruft Codec) und 2 (Kennung/Zeit vor Wiederholung, Bestand) | 10 von 10 rot; die drei Wahl-Mutanten vor der Dienstprüfung auch isoliert (`-run` allein auf die neuen Tests) durch `TestSatzKeinAufrufVorDemLesen` (Panic); **aber** F-2: Codec im Fehlerzweig von `waehleDokument` bleibt grün |
| Z9 | `Codecs` `nil` wird `ErrUngueltig` | rot, 100 Tests |

Raster (§6): Zeile `New` und `Import` je Zelle einer Mutation zugeordnet (leer = Z3, übrige Felder = Z5, geteilter Satz = Z5 und Z6, unbekannter Wert = Z8, Abbruch = Z1 und Z2, Fehler = Z2); jede Zelle, die „gilt nicht“ sagt, trägt ihren Grund. Die Zelle *Wiederholung* nennt „keine Kennung, keine Zeit (Z8)“: Z8 zählt nur Codecs; Kennung und Zeit trägt der Bestand (Aufruf von `IDs.SnapshotID()` und `Uhr.Now()` vor der Wiederholung, 2 von 2 rot). Zusatztabelle: Paar *Satz teil-nil und leere ID* hat seinen Fall; die drei Dienste in einer Konfig führt Z1.

**(2) Z7, selbst gemessen.** Binary: `go test -c` mit `-race` (Image `odata-recorder:race`, cgo, 4 CPUs, ohne Netz) und ohne `-race` (cgo aus), je Prozess `-test.run '^TestSatzGleichzeitig$' -test.count=1`, **30 Prozesse je Zeile**, Messender der Reviewer. Ungemutiert: 30 von 30 grün, in beiden Bauarten.

| Mutation (Paketvariable, außer b) | mit `-race`: rot (davon `DATA RACE`) | ohne `-race`: rot | Lesart |
|---|---|---|---|
| (a) Wahl des Antwort-Codecs vor `ParseMetadata`, gelesen bei Collection und Entity | 30 (4) | 30 | Wirkungs-Fänger fängt |
| Restmutation: gelesen nur bei `ParseCollection` | 30 (4) | 30 | fängt |
| Restmutation: gelesen nur bei `ParseEntity` | 30 (8) | 30 | fängt |
| (c) Modell nach `ParseMetadata` gemerkt, gelesen bei Collection und Entity | 30 (30) | 30 | fängt |
| (c) gelesen nur bei `ParseCollection` | 30 (30) | 30 | fängt |
| (c) gelesen nur bei `ParseEntity` | 30 (30) | 30 | fängt |
| (b) `sync.Mutex` um den ganzen Import | 30 (0) | nicht gemessen | Frist der Schranke; kein Datenrennen im Spiel |
| (b) Mutex um `liesAntwort`; um `liesDokument` | 30 (0); 30 (0) | nicht gemessen | Frist der Schranke |
| Fenster: Schreiben und Lesen der Paketvariable unmittelbar hintereinander in `liesAntwort` | 30 (30) | **0** | **nur der Detektor** (F-3) |
| Fenster: Schreiben in `liesAntwort`, Lesen bei `ParseEntity` | 30 (30) | **1** | **nur der Detektor** (F-3) |
| Zähler auf Paketebene, ohne Wirkung auf die Wahl | 29 (29) | **0** | Detektor 29 von 30, Wirkung nie |
| Fenster mit `atomic.Value` (kein Datenrennen), Schreiben und Lesen hintereinander | **2** (0) | **0** | **nichts fängt** (F-3) |
| Fenster mit `atomic.Value`, Lesen bei `ParseEntity` | **10** (0) | **2** | **nichts fängt zuverlässig** (F-3) |

Die drei Mutationen des Plans (a), (b), (c) und die Restmutation „liest die Paketvariable nur bei `ParseCollection`/`ParseEntity`“ sind in jedem Lauf rot, wie der Plan sagt: 30 von 30 je Variante, mit und ohne Detektor (bei (b) nur mit). Die zweite `schranke` trägt die Zusage dort. Außerhalb der Haltepunkte (Zugriffe ohne Haltepunkt dazwischen) trägt sie nicht.

**(3) V4-Tests textgleich.** `git diff --numstat b5ac4a5..8f5e466 -- '*_test.go'` nennt allein `satz_test.go` (629 hinzu, 0 entfernt); `helfer_test.go` und jede vorhandene `*_test.go` sind unberührt. Die Suite des Pakets ist grün: 112 `--- PASS`, `-count=20` über die neuen Tests grün, `go test -race -count=1` über Paket und `ports/...` grün (Reviewer, HEAD).

**(4) Aufrufstellen-Argumente.** `pruefeAufruf` prüft je Aufrufstelle `ctx` (dasselbe Objekt, Wertschlüssel je Dienst), `grenze` (`satzGrenze`, ein Wert, den keine Vorgabe der Welt trägt), Modell (`reflect.DeepEqual` gegen das Modell, das `ParseMetadata` desselben Dienstes lieferte; die drei Modelle sind verschieden, das prüft `satzWelt`), `entitySet` (drei verschiedene) und Strom (Identität gegen den Strom von `Open`, ungehüllt); die Fakes der Sätze prüfen Hülle, verbrauchten Strom und fehlende Vorgabe in `t.Cleanup`. Jede Veränderung eines Arguments an einer der drei Aufrufstellen (Z1, 13 Mutanten) ist rot.

**(5) Hexagon.** `make a-check`: 0 Befunde. `CodecSatz` liegt im Port-Paket `ports/outbound` und importiert nichts; `baselineimport` bekommt keinen Import hinzu. `gofmt -l` und `go vet` über beide Pakete: sauber.

**(6) Größenmarke.** `git diff --numstat b5ac4a5..8f5e466` über die Pfade des Blocks: 16 + 43 + 7 + 629 = **695** (Marke 1500, Schätzung 1045); die Messung des Implementers in §4 stimmt.

**(7) Rohbytes (§3.18).** Gegengelesen mit `LC_ALL=C grep -P` die sechs geänderten Dateien: keine Folge Backslash-`u`-vier-Hex in den hinzugefügten Zeilen, kein BOM, NBSP, Nullbreiten- oder Steuerzeichen; Zeichen jenseits von ASCII in `src/` sind ausschließlich `ä ö ü ß` in Kommentaren und Bezeichnern von Meldungen. Kein Fixture des Diffs trägt einen Wert dieser Klasse.

**(8) Kommentar-Regeln (§3.7, §3.11).** Die Kommentare in `importer.go`, `metadaten.go` und `codec_satz.go` tragen Zusage, Kopplung oder Rang-Zeiger und beschreiben, was da ist; der Kommentar von `Konfig` ist nachgezogen (Risiko 7 des Plans). Zwei Ausnahmen: F-4, F-5.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Zusage Z2 („danach kein weiterer Port außer `Close`“) hat einen Fänger, der nur die drei Zähler des Ziel-Satzes sowie `Blobs.Begin` und `Imports.Publish` liest. Ein Aufruf des `MetadatenCodec` der Konfig oder des Metadatenlesers eines anderen Satzes nach dem Fehler von `ParseMetadata` des Satzes lässt die ganze Suite grün (0 rote Tests, beide Mutanten). Die entsprechenden Mutanten für `ParseCollection`/`ParseEntity` sind rot nur, weil der Fake der Konfig für fremde Referenzen keine Vorgabe hat (Zähler `fehlt`), nicht durch einen Zähler der anderen Codecs. Failure-Szenario: ein späterer Rückfall „bei Fehler den Leser der Konfig versuchen“ liest das `$metadata` einer Version mit dem Leser der anderen, ohne dass ein Test es merkt. | `AGENTS.md` §3.15; Plan Z2 | `satz_test.go` · `for _, o := range []string{"Blobs.Begin", "Imports.Publish"}` und `ziel.p.anzahl(o) - vor[ziel][o]`; Plan §3 · Z2 „danach kein weiterer Port“ | ja: Mutanten `ZG-meta-failover-konfig-nur-bei-Satz`, `ZG-meta-failover-anderer-satz` (je grün) | Fehlerpfad zählt nur die Ports des Ziels, nicht die der anderen Dienste |
| F-2 | MEDIUM | Z8 führt 7 von 10 Ausgängen der Rangfolge `SPEC-127` Nr. 4 vor dem Lesen: es fehlen `$metadata` fehlt, `$metadata` mehrdeutig und Hash der Anfrage ungleich der Referenz (Nr. 7 a bis e). Ein Aufruf des Metadatenlesers des Satzes im Fehlerzweig von `waehleDokument` lässt die ganze Suite grün (0 rote Tests); der Bestand sieht den Satz nicht, Z8 hat den Fall nicht. Failure-Szenario: eine Vorab-Prüfung der Version (`SPEC-175` Nr. 4 d) wird an diese Stelle gesetzt und ruft den Leser eines Satzes, bevor die Wahl des Dokuments entschieden ist. | `AGENTS.md` §3.15, §3.17; Plan Z8 | `satz_test.go` · `faelle := []fall{` (sieben Ausgänge); Plan §3 · Z8 „Schleife über Ausgang“ | ja: Mutant `ZH-waehleDokument-fehler-ruft-satzcodec` (grün); `ZH-ladeSession-...` und `ZH-liesDokument-...` rot | Fänger führt nur eine Teilmenge der Ausgänge vor dem Lesen |
| F-3 | MEDIUM | Z7 fängt Paketvariablen, die zwischen den Haltepunkten geschrieben und gelesen werden (30 von 30), nicht aber ein Fenster ohne Haltepunkt dazwischen: Schreiben und Lesen hintereinander in `liesAntwort` ist ohne Detektor 0 von 30 rot, mit Detektor 30 von 30; mit `atomic.Value` (kein Datenrennen) 0 von 30 ohne und 2 von 30 mit `-race` (ohne `DATA RACE`); Schreiben in `liesAntwort`, Lesen bei `ParseEntity`: 2 von 30 (atomar, ohne `-race`), 10 von 30 (atomar, mit `-race`). Der Plan nennt allein, `make test-race` fange (a) „nicht zugesagt“; er nennt weder dieses Fenster noch Grund und Folge als akzeptiertes Negativ. Failure-Szenario: ein Zwischenspeicher der Wahl mit atomarem Zugriff liefert gelegentlich den Satz des anderen Dienstes (V4-Leser auf dem Dokument von V2), und `TestSatzGleichzeitig` bleibt in 28 bis 30 von 30 Läufen grün. | `AGENTS.md` §3.10; Reviewer-Skill (*Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*, *Akzeptiertes Negativ ohne Grund und Folge*) | Plan §3 · Z7 „in jeder Verschränkung“ und Absatz *Nebenläufigkeit (§3.10)*; `satz_test.go` · `setzeHaken("ParseCollection"` | ja: Mutanten `a-fenster-schreiben+lesen-in-liesAntwort`, `a-atomic-fenster`, `a-atomic-schreiben-liesAntwort-lesen-Ent`, `r-nur-schreiben-ohne-wirkung` im Scratchpad (`rvm/tool/rate.go`) | Zusage über Nebenläufigkeit mit Fenster ohne Haltepunkt, weder gefangen noch als Negativ benannt |
| F-4 | LOW | Der Kopf von `TestSatzFehlerJeAufrufstelle` sagt, danach rufe der Dienst „weder `Blobs.Begin` noch `Imports.Publish` noch einen weiteren Codec“; der Test zählt von den Codecs allein die Aufrufstellen des Ziel-Satzes (F-1). | `AGENTS.md` §3.11 | `satz_test.go` · `danach ruft der Dienst weder Blobs.Begin noch Imports.Publish noch einen weiteren Codec` | ja: die Mutanten von F-1 | Kommentar sagt mehr zu als der Test fängt |
| F-5 | LOW | Der Kopf von `satzDienste` nennt `fehler[k]`; der Parameter heißt `ersatz`. | `AGENTS.md` §3.7 | `satz_test.go` · `fehler[k] ersetzt den Satz des Dienstes k` | nein | Kommentar nennt einen Namen, den es nicht gibt |
| F-6 | INFO | Die Mutationen der Zeilen Z2 (a) bis (c) und Z1 (d) bis (f) sind im Plan als „im Satz-Zweig“ beschrieben; der Code hat keinen eigenen Satz-Zweig, `liesAntwort`, `liesDokument` und `oeffne` sind für Satz und Konfig dieselben, und die Mutanten treffen auch den Bestand (bis 100 rote Tests). Die Wahl allein steht in `metadatenCodec` und `datenCodec`. Keine Aktion; für Reviewer und Verifier: die satzeigenen Mutanten sind die der Auswahl (Z1 c, Z5, Z8, F-1, F-2). | Maintainability | Plan §3 · Z2 „(c) der Satz-Zweig von `ParseMetadata` schließt den Strom nicht“ | nein | Planwortlaut nennt einen Zweig, den der Code nicht hat |
| F-7 | INFO | Messung 695 bei Marke 1500 und Schätzung 1045: die Formel (`AGENTS.md` §3.14) überschätzt diesen Slice um rund die Hälfte (Zeilen mit Fakes und Tabellen kosten hier weniger als × 40/85). Kein Verstoß; Hinweis an den Planner. | `AGENTS.md` §3.14 | Plan §4 · `Messung beim Handoff` | ja: `make groessenmarke` (Verifier) | Formel überschätzt |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/outbound/codec_satz.go` | geprüft, ohne Befund (Struktur im Port-Paket, `SPEC-175` Nr. 3; Kopf nennt nur, was Z5 fängt: `Version`, `NurLesend` ohne Wirkung, Satz teilbar) |
| `src/hexagon/services/baselineimport/importer.go` (`Dienst.Codecs`, `normiereDienste`, `metadatenCodec`, `datenCodec`, `liesAntwort`) | geprüft, ohne Befund; Wahl je Aufruf aus `Dienste`, hinter Nr. 4 b, `Importer` bleibt bei einem Feld, Paketkopf „hält keinen Zustand“ gilt |
| `src/hexagon/services/baselineimport/metadaten.go` (`liesDokument`) | geprüft, ohne Befund |
| `src/hexagon/services/baselineimport/satz_test.go`, Zeilen Z1, Z3, Z4, Z5, Z6, Z9 | geprüft, ohne Befund (57 von 57 Mutanten rot) |
| Raster in §6 und Zusatztabelle | geprüft, ohne Befund außer den Zellen unter F-1 und F-2 (dort ist der Fänger zu schmal, nicht der Ort leer); keine im Code entschiedene Zelle gefunden |
| akzeptierte Negative (Schnittstelle mit nil-Zeiger, `NurLesend` ohne Leser, Satz-Codec gleich Konfig-Codec, Satz nach `New` geändert) | geprüft, ohne Befund: Grund und Folge stehen im Plan; `SPEC-127` Nr. 1 nennt den letzten als *Ungeprüft* |
| `spec/spezifikation.md` (Klarstellung L1, `SPEC-127` Nr. 1, Historienzeile) | geprüft, ohne Befund: keine Änderung einer Entscheidung, Code und Test folgen dem Wortlaut (nur `Metadaten` und `Daten`; geteilter Satz; Konfig-Codecs bleiben Pflicht) |
| Plan §1, §3, §4, §6, Kopf (`AGENTS.md` §3.9, §3.13) | geprüft, ohne Befund: Tabelle, Absatz der Mutationen, Messung und Berichtigung der DoD stehen im Handoff-Commit; keine Sendung an einen anderen Slice |
| Hexagon (`make a-check`, `gofmt`, `go vet`) | geprüft, ohne Befund |
| Suppression (§3.2) | geprüft, ohne Befund: keine `//nolint`-Direktive im Diff |
| Rohbytes (§3.18) | geprüft, ohne Befund |

## Summary

0 HIGH, 3 MEDIUM, 2 LOW, 2 INFO. Der Produktcode trägt die Zusagen: 81 von 84 Mutanten rot, die drei grünen sitzen in zwei Fängern (Z2 und Z8), nicht im Code; Z7 ist für die Mutationen des Plans und die Restmutation in 30 von 30 Prozessen rot, mit und ohne Detektor. Die Lücken sind Fänger, die schmaler sind als ihre Zusage: der Fehlerpfad zählt die anderen Codecs nicht (F-1), Z8 führt nicht alle Ausgänge vor dem Lesen (F-2), Z7 hat für das Fenster ohne Haltepunkt weder Fänger noch benanntes Negativ (F-3). Den Fix prüfen Reviewer und Verifier gegen die Klasse (§3.17): F-1 über alle Aufrufstellen mal alle fremden Codecs, F-2 über alle Ausgänge der Rangfolge vor Nr. 7 f, F-3 über Schreiben und Lesen an jeder Stelle zwischen den Haltepunkten oder als akzeptiertes Negativ mit Grund und Folge. Die Läufe und Mutanten liegen im Scratchpad (`rvm/tool/main.go`, `rvm/tool/rate.go`, Logs `muts.log`, `zf.log`, `zg.log`, `zh.log`, `z7-race.log`, `z7-norace.log`, `z7-atomic.log`).

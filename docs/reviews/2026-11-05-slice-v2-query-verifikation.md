# Verifikationsbericht: slice-v2-query (Teil A) — 2026-11-05

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan und DoD, nicht gegen den Diff (Reviewer) und nicht gegen den Bedarf (Validator).

**Gegenstand:** `slice-v2-query` (Welle `welle-odata-v2`), Zweig `slice-v2-query`, geprüfter Stand **HEAD `02e14ff`** (nach der Fix-Runde), Basis des Blocks `ab7b3083`. Bezug: `LH-FA-17`, `LH-QA-05`, `LH-RB-01`, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md).

**Geprüft gegen:** DoD, §3 (Tabelle *Zusage · Fänger* mit 49 Zeilen, *Mutationen und Raten*, *Fix-Runde nach dem Review*, Betriebsregeln), §4 (Block, Marke), §6 (Raster, Zusatztabelle, Risiken, Sendungen) des Plans; `SPEC-160` bis `SPEC-162`, `SPEC-173`, `SPEC-174`; [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md); `AGENTS.md` §3.10, §3.15 bis §3.18; Review `docs/reviews/2026-11-05-slice-v2-query-review.md` (F-1 bis F-5).

**Methode.** Der Beleg des Implementers ist nicht übernommen. Selbst gefahren: `make gates` (ein Lauf, Exit 0), Rohbyte-Gegenlesen aller 36 geänderten Dateien, 24 eigene Mutanten und zwei Probe-Pakete auf Wegwerfkopien (Scratchpad `.../scratchpad/mut/`, Quelle der Kopien der Arbeitsbaum von HEAD; Image `odata-recorder:test`, `--memory=2g`, ohne Netz, Paket-Verzeichnis `:ro` eingehängt). Im Repo liegt nichts außer diesem Bericht (`git status` vor dem Bericht leer).

---

## Zahlen (selbst gemessen)

| Größe | Wert |
|---|---|
| `make gates` auf `02e14ff` | Exit 0 (a-check, d-check `587 Datei(en), 0 Befund(e)`, build, test, race, groessenmarke, handbuch, lint, zeichen `1144 Datei(en)`, migration-check); `git status` danach sauber |
| `make groessenmarke` | `slice-v2-query.md: ok, Messung 3781 von Marke 7000` (Review maß 3721 im Stand `f4651c9`; die Fix-Runde fügt 63 Zeilen in Tests hinzu) |
| Fix-Runde `02e14ff` | vier Dateien: Plan, `literale_test.go`, `query_test.go`, `rang_test.go`; **Produktcode unverändert** (`git diff f4651c9 02e14ff -- src` nennt nur `*_test.go`) |
| Probe-Paket Schlüssel (eigen) | 5460 Fälle (1 bis 6 Teile × 3 Klassen × 5 Rotationen): am HEAD-Code 0 Abweichungen vom Orakel; am Mutanten F1a 1220 Meldungen |
| Probe-Paket Rangfolge (eigen) | `$select` mit 257 und `$orderby` mit 33 Elementen, je eine von 6 Formen an jeder 7. Stelle: am HEAD-Code 0 Abweichungen |

## Plan-vs-Code-Diff

| Plan sagt | Code sagt | Urteil |
|---|---|---|
| Parser in `src/adapters/driven/odata/urlparserv2/`, `New(opts ...Option)`, `WithFilter`, `FilterLeser`; Kopie von V4, kein Import | Importe des Pakets: nur `model`, `ports/outbound` und das eigene Paket (Tests); kein Paket unter `odata/` ohne `v2`; `make a-check` 0 Befunde; `TestKeinV4Import` grün | konform ([ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md)) |
| `ODR-E6006`, Status 501, Ziel `mock`, Text `Unsupported OData V2 feature`, hinter `CodeE6005` | `fehler.go`: `CodeE6006` direkt hinter `CodeE6005`, Eintrag `{CodeE6006, "ODR-E6006", []int{501}, ZielMock, "Unsupported OData V2 feature"}` | konform |
| Abbildung `ErrKompatibilitaet` → `ODR-E6006` | `abbildung.go`: Zeile `ErrKompatibilitaet` → `CodeE6006`, 1 Zeile | konform |
| Feld `URL URLCodec` in `CodecSatz` | `codec_satz.go`: `URL URLCodec` mit Kommentar | konform |
| Query-AST unverändert, `ParseURL(pfad, rawQuery)` ohne Kopfzeilen, ohne Kontext | Signatur wie im Plan; in `model` nur `KompatFehler` und der Code | konform |
| Tests der Telemetrie, `klassen_test.go`, Tests der Dienste: Zahlen 35/36 nachgezogen | `numstat`: je 1 bis 16 Zeilen, nur Zahlen und die Liste der Fehler der Abbildung | konform |
| Marke 7000 (Metrik: `numstat` über die Pfade des Blocks) | Gate: 3781 | konform |
| Keine Suppression (§3.2) | kein `nolint` im Paket; `make lint` grün | konform |
| Fix-Runde: F-3 Zahlen der Zusatztabelle 38 und 18 | `grep` über den Plan: „41 Paare“ und „19 Paare“ kommen nicht mehr vor; Zeile 342 nennt 38 und 1600, Zeile 344 nennt 18 | konform |
| Fix-Runde: F-1 `TestSchluesselLiteralPaar` Menge 1600, F-2 `TestRangElementzahl` Menge 323 | beide Tests liegen; Zählprüfung 1600 und 323 im Test; grün | konform, Klasse nicht gedeckt, siehe V-1 und V-2 |

## DoD

| Punkt | Zustand |
|---|---|
| 1. Parser, AST, `ODR-E6005`, `WithFilter` mit `FilterLeser`, ohne Leser Kompatibilität `filter` | erfüllt (Zeilen 19 bis 21, 45, 49 der Tabelle; eigene Mutanten und Probe-Pakete) |
| 2. `*model.KompatFehler` mit `ODR-E6006`, 36 Codes, Abbildung, Feld `URL` | erfüllt (Plan-vs-Code oben; `TestKompatFehlerModell`, `TestAbbildungKompat`, `TestCodeTabelle` laufen im Gate grün) |
| 3. Randformen des Lesers: je Zelle Ort und Fänger; Generator über die Klasse; echte V2-URLs | **erfüllt mit Einschränkung**: das Raster hat keine leere Zelle, 49 Zeilen tragen Fänger; zwei Klassen haben Lücken im Fänger (V-1, V-2); Generator Zeile 48 läuft (2000 Läufe) |
| 4. `make gates` grün | erfüllt (eigener Lauf, Exit 0) |
| 5. Review durchgeführt, Report liegt vor | erfüllt (`docs/reviews/2026-11-05-slice-v2-query-review.md`) |
| 6 bis 11 (Doku-Update, Handbuch entfällt mit Grund, Closure-Notiz, Register, Ausgänge in §6, Paarungen) | offen, Sache der Closure; §7 ist leer; §6 trägt Ausgänge, davon zwei *weiter offen* mit Adresse (Anhang A an den Planner; Kopfzeilen-Fänger an `slice-v2-bootstrap`) — kein Verstoß vor der Closure |

## Fixes des Reviews gegen die Klasse (AGENTS.md §3.17)

| Befund | Klasse (Wortlaut der Zusage) | Elemente im Fänger | Eigene Mutanten an Elementen, die der Befund nicht nannte | Urteil |
|---|---|---|---|---|
| F-1 (MEDIUM) | `SPEC-173` Nr. 1, 6: „ein Verstoß der Klasse `ungueltig` geht jedem Kompatibilitätsfehler vor, **gleich wo er steht**“ — im Schlüssel: Paar aus `literal_typ` und ungültigem Literal an **jeder Stelle und in jeder Teilezahl** (1 bis 16 Teile) | 1600 Paare, **alle in Schlüsseln mit genau zwei Teilen** (`A=`, `B=`) | F1a Rückkehr nach `literal_typ` ab dem zweiten Teil: **grün**; F1b Rückkehr, wenn der Schlüssel mehr als zwei Teile hat: **grün**; F1d Rückkehr am dritten Teil: **grün**; F1c Rückkehr am ersten Teil: rot (`TestSchluesselLiteralPaar`) | **nicht geschlossen** (V-1) |
| F-2 (MEDIUM) | `SPEC-162` Nr. 6: „mehr als 256 Elemente: `zu_viele` (**vor der Prüfung der Elemente**)“, Nr. 7 mit `SPEC-173` Nr. 4: 32 Elemente — Menge: jede Form eines Elements (Kompatibilität `option`; `wert`: leer, doppelt, Leerraum, `(`, Name über 128 Zeichen, bei `$orderby` falsche Richtung) × jede Stelle | 257 Stellen × `/` (`$select`), 33 Stellen × `(` und `/` (`$orderby`); **keine `wert`-Form** | F2c `$select`: leeres Element vor der Zählung als `wert`: **grün**; F2d `$select`: doppeltes Element vor der Zählung: **grün**; F2e `$orderby`: leeres Element vor der Zählung: **grün**; F2f `$orderby`: Richtung `up` vor der Zählung: **grün**; F2b `$orderby`: Zählung hinter der Elementschleife: **grün**; F2a `$select`: Zählung hinter Pfadtest und Schleife: rot; F2g, F2h `/` nur am ersten oder letzten Element: rot (`TestSelect`); F2i `(` nur am ersten Element: rot | **nur für `option` gedeckt**, `wert` offen (V-2) |
| F-3 (LOW) | Zahlen der Zusatztabelle an jeder Stelle des Plans | Plan §3 Zeile 7 und 9, §6 Bullet, §6 Zusatztabelle | `grep` über alle Zahl-Stellen („41 Paare“, „19 Paare“, „(41“, „(19“) | geschlossen |
| F-4, F-5 (INFO) | Adressen und Merge | — | — | siehe V-4 und V-5 |

## Mutanten dieser Prüfung

24 Mutanten auf Wegwerfkopien, je ein Lauf `go test` des Pakets, jeder deterministisch; die Grundlinie (ungemutiert) ist grün. Ein weiterer Entwurf (Prüfung des Folge-Literals im Schlüssel) ist nicht gewertet, sein Lauf kam nicht zu Ende.

| Gruppe | Mutanten | rot | grün |
|---|---|---|---|
| Grenzen (die 9 Konstanten in `parser.go` je um 1 über die Grenze, dazu `maxBruchteil` in `zeit.go`) | 10 | 10 | 0 |
| `query.go`: `asciiKlein(wert)` an allen Stellen entfernt (Schreibweise von `$format` und `$inlinecount`) | 1 | 1 (`TestFormat`) | 0 |
| F-1, Klasse Teilezahl und Stelle (F1a, F1b, F1c, F1d) | 4 | 1 (F1c) | **3** |
| F-2, Rang der Zählung gegen `option` (F2a, F2g, F2h, F2i) | 4 | 4 | 0 |
| F-2, Rang der Zählung gegen `wert` (F2b, F2c, F2d, F2e, F2f) | 5 | 0 | **5** |
| Summe | 24 | 16 | 8 |

## Findings

| ID | Kategorie | Befund | Quelle | Pfad |
|---|---|---|---|---|
| V-1 | MEDIUM | Der Fix von F-1 deckt die Klasse nicht. `TestSchluesselLiteralPaar` führt nur Schlüssel mit zwei Teilen (`A=…,B=…`). Die Zusage („gleich wo er steht“) gilt für jede Stelle und jede Teilezahl bis 16. Drei Mutanten, die der Befund nicht nannte und die zur selben Klasse gehören — Rückkehr nach `literal_typ` ab dem zweiten Teil, bei mehr als zwei Teilen, am dritten Teil —, bleiben grün; etwa müsste `/S(A=1,B=1d,C=null)` `ODR-E6005` sein und wäre `ODR-E6006`. Der **Code ist richtig**: ein Probe-Paket über 5460 Schlüssel mit 1 bis 6 Teilen (Klassen gültig, `literal_typ`, ungültig) stimmt in jedem Fall mit dem Orakel überein; es fehlt allein der Fänger (`AGENTS.md` §3.15, §3.17). Die Mutation an einem Element, die der Fix-Bericht nennt (`zu_lang`, 1 von 32 Formen), liegt in derselben Dimension wie der Befund (die Form des zweiten Literals), nicht in der Dimension Stelle und Teilezahl. | `SPEC-173` Nr. 1, 6; `AGENTS.md` §3.17 | `src/adapters/driven/odata/urlparserv2/rang_test.go` · `TestSchluesselLiteralPaar` |
| V-2 | MEDIUM | Der Fix von F-2 deckt die Klasse nur für die Kompatibilitätsformen. `SPEC-162` Nr. 6 sagt ausdrücklich „`zu_viele` (vor der Prüfung der Elemente)“; die Menge der Elementformen schließt die Formen der Ursache `wert` ein (leeres Element, doppeltes Element, Leerraum, `(`, Name über 128 Zeichen, bei `$orderby` falsche Richtung). Fünf Mutanten (F2b bis F2f) bleiben grün: eine Prüfung eines leeren, doppelten oder falsch gerichteten Elements vor der Zählung ändert für 257 bzw. 33 Elemente das Ergebnis von `zu_viele` zu `wert`, und kein Test führt 257 oder 33 Elemente mit einem solchen Element. Der **Code ist richtig** (Probe: 6 Formen an je 37 bzw. 5 Stellen je Option, 0 Abweichungen). Für `$orderby` sagt `SPEC-162` Nr. 7 die Rangfolge nur durch die Folge der Sätze, nicht mit dem Wort „vor“; der Ort ist damit schwächer als für `$select`, die Folge gilt aber in beiden Optionen im Code und in der Zeile 43 des Plans. | `SPEC-162` Nr. 6, 7; `AGENTS.md` §3.17 | `src/adapters/driven/odata/urlparserv2/query_test.go` · `TestRangElementzahl` |
| V-3 | LOW | In 75 der 1600 Paare von `TestSchluesselLiteralPaar` wird nur ein Literal gelesen: die ungültigen Formen `'a`, `'a''` und `guid'` tragen eine ungerade Zahl von Anführungszeichen; steht die Form im ersten Teil (`A='a,B=1d`), schluckt `zerlegeTeile` den Rest als Text, und das Paar entartet zu einem einzelnen Verstoß. Der Test bleibt grün und ist für diese 75 Paare kein Paar-Fänger; die Zahl 1600 im Plan zählt sie mit. | `AGENTS.md` §3.15 (die Menge zählt, was prüft) | `src/adapters/driven/odata/urlparserv2/rang_test.go` · `TestSchluesselLiteralPaar` |
| V-4 | INFO | Zweig und `main` stehen im Konflikt in **einer** Datei: `spec/spezifikation.md` (`git merge-tree --write-tree HEAD main`; 52 Commits auf `main` seit der Merge-Basis `b41380b`). `outbound/codec_satz.go`, `abbildung.go`, `fehler.go` und der Plan sind ohne Konflikt. `main` trägt außerdem `slice-v2-query-filter.md` unter `open/` und eine geänderte `welle-odata-v2.md`; die Sendungen aus §6 des Plans sind gegen diese Pläne nicht geprüft. Zur Rolle des Integrations-Planners; kein Befund des Slice. | Merge | `spec/spezifikation.md` |
| V-5 | INFO | Die offenen Adressen aus dem Review (F-4) sind unverändert: Anhang A (Planner) und der Fänger „Versions-Kopfzeilen ändern nichts“ (`slice-v2-bootstrap`) stehen in §6 mit Ausgang *offen*. Ihre Annahme ist am Zweig nicht prüfbar. | `AGENTS.md` §3.13 | Plan §6 |
| V-6 | INFO | Die Erwartung des Plans (rund 3800, Spanne 3400 bis 4400) trifft: gemessen 3781 von Marke 7000. | `SPEC-128` | Plan §4 |

0 HIGH, 2 MEDIUM, 1 LOW, 3 INFO.

## Rohbytes der geänderten Dateien (AGENTS.md §3.18)

Anlass: Der Implementer meldet `sed -i` und Heredoc auf Repo-Dateien (Betriebsregel im Plan §3). Gelesen wurden alle **36** Dateien des Diffs `ab7b3083..HEAD` (34 Quellen unter `src/`, der Plan und der Review-Report), mit `LC_ALL=C grep -P`:

| Prüfung | Treffer |
|---|---|
| Folge Backslash, kleines oder großes `u`, vier bzw. acht Hex-Ziffern | 0 |
| BOM (`EF BB BF`) | 0 |
| NBSP (`C2 A0`), U+0085, U+2000 bis U+200F, U+2028 bis U+202F, U+2060 bis U+206F, U+3000, U+FEFF | 0 |
| Steuerzeichen außer Tabulator und Zeilenvorschub, `\r` | 0 |
| Datei endet ohne Zeilenvorschub; ungültiges UTF-8 (`iconv`); `git diff --check` | 0, 0, 0 |
| Nicht-ASCII-Zeichen in den Dateien | nur `ä ö ü ß Ä Ö Ü × § · ÷ ±`; ein `é` in `src/hexagon/services/session/beenden_test.go` Zeile 902, im Stand `ab7b3083` ebenfalls vorhanden (Bestand, vom Slice nicht berührt) |
| `make zeichen-check` (253 Zeichen, alle Textdateien) | grün (`1144 Datei(en)`) |
| Diff der Dateien von `model`, `outbound`, Telemetrie, `httpodata`, Dienste | nur die im Plan genannten Zahlen und Zeilen; kein verfälschter Wert |

**Ergebnis: kein Inhalt verfälscht.** Die Testwerte jenseits von ASCII entstehen in `helfer_test.go` aus Codepunktzahlen (`zeichen`) und `%XX`-Folgen (`pct`); in keiner `*_test.go`-Datei steht U+0085, U+00A0, U+2003, U+3000 oder U+FEFF als Rohzeichen, und keine Folge Backslash-`u`-vier-Hex ist zu einem Zeichen geworden. Gegengelesen wurde auch dieser Bericht nach dem Schreiben (Datei `docs/reviews/2026-11-05-slice-v2-query-verifikation.md`, `LC_ALL=C grep -P`, 0 Treffer in den Klassen oben).

## Raster §6 und Zusatztabelle

Alle fünf Zeilen des Rasters tragen in jeder der sechs Zellen eine Nummer von `SPEC-160` bis `SPEC-162`, `SPEC-173` oder `SPEC-174` oder „gilt nicht“ mit Grund; die Nummern von `SPEC-162` Nr. 6, 7 und `SPEC-173` Nr. 4 habe ich am Wortlaut gegengelesen. Die Zusatztabelle hat für jedes Paar von Verstößen einen Fänger; die Zeilen mit Paaren im Schlüssel und mit Elementzahl gegen Form sind die Stellen von V-1 und V-2. Keine im Code entschiedene Zelle gefunden.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `make gates` | geprüft, ohne Befund: Exit 0, Arbeitsbaum danach sauber |
| Plan-vs-Code (Fehlercode, Abbildung, Feld `URL`, Signatur) | geprüft, ohne Befund |
| [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) (Trennung der Versionen) | geprüft, ohne Befund: kein V4-Import, `TestKeinV4Import` und `make a-check` grün; der AST bleibt unverändert |
| Grenzen (`SPEC-160` Nr. 6) | geprüft, ohne Befund: 10 von 10 Mutanten über die Grenze rot |
| Code gegen Klasse (Probe-Pakete zu V-1, V-2) | geprüft, ohne Befund am **Code**; die Lücken liegen in den Fängern |
| Größenmarke | geprüft, ohne Befund: 3781 von 7000 |
| Rohbytes | geprüft, ohne Befund (Abschnitt oben) |
| Suppression (§3.2) | geprüft, ohne Befund |
| Nebenläufigkeit (Zeilen 17, 18) | nicht erneut gemessen: die Fix-Runde berührt keinen Produktcode und keinen Nebenläufigkeitstest; `make test-race` ist im Gate-Lauf grün |

## Ergebnis

**Konform bis auf die Klasse der beiden Fixes.** DoD 1 bis 5 sind erfüllt, der Plan stimmt mit dem Code überein, die Gates sind grün, die Rohbytes sind unverfälscht. F-3 ist geschlossen. F-1 und F-2 sind **nicht gegen die Klasse** geschlossen (V-1 und V-2, je MEDIUM): jede Zusage gilt für eine Menge, die größer ist als die Elemente, die der Fix führt; Mutanten an den nicht genannten Elementen (Teilezahl und Stelle im Schlüssel; die `wert`-Formen gegen die Elementzahl) bleiben grün, obwohl der Code die Zusage hält. Die Lücken sind reine Test-Lücken und klein zu schließen (eine Schleife über die Teilezahl 3 bis 16 mit der Stelle des `literal_typ`; 257 bzw. 33 Elemente mit je einer `wert`-Form an jeder Stelle); sie berühren weder Produktcode noch Spezifikation.

An den Planner: V-1 und V-2 vor der Closure an den Implementer zurückgeben, oder als akzeptiertes Negativ mit Grund und Folge in §3 führen (ein Negativ ohne Grund wäre eine Zusage ohne Fänger, `AGENTS.md` §3.15); V-3 mit V-1 zusammen; V-4 und V-5 im Integrationsschritt. Diese Prüfung wird nicht wiederholt; die Mutanten stehen im Scratchpad zur Nachprüfung bereit.

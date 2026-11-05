# Verifikationsbericht: slice-v2-query-filter (Teil B) — 2026-11-05

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan und DoD, nicht gegen den Diff (Reviewer) und nicht gegen den Bedarf (Validator).

**Gegenstand:** Zweig `slice-v2-query-filter`, HEAD `35c7aa8` (Implementer `8a316f0`, Review `ca054db`, Architect `e320a6d`, Fix-Runde `35c7aa8`), Basis des Blocks `80d7910c`. Bezug: `LH-FA-17`, `LH-QA-05`, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md).

**Geprüft gegen:** DoD, §3 (Tabelle *Zusage · Fänger*, *Fix-Runde zum Review*), §4 (Block, Marke), §6 (Raster, Zusatztabelle) des Plans; `SPEC-173` Nr. 5 (mit Schärfung und Berichtigung vom 2026-11-05), `SPEC-164` Nr. 1 bis 4 und 7 bis 9 (Nr. 6 mit der Zuordnung nach `SPEC-173`), [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md); `AGENTS.md` §3.10, §3.15 bis §3.18; Review `docs/reviews/2026-11-05-slice-v2-query-filter-review.md` (F-1 bis F-8).

**Methode.** Belege des Implementers nicht übernommen. Selbst gefahren: `make gates` (ein Lauf), 67 Mutanten (ohne Grundlinie) und 3 Probe-Pakete auf Wegwerfkopien (Scratchpad `scratchpad/ver-filter/`, Image `odata-recorder:test`, `--network none --memory=4g --cpus=4`, `-count=1`, `-run TestFl`; Z18-Gemeinsam mit dem ganzen Paket), 90 Prozessläufe je Zeile 32 und 33 im Image `odata-recorder:race`. Die Mutanten fahren je Element, das weder Review noch Implementer nannten.

---

## Zahlen (selbst gemessen)

| Größe | Wert |
|---|---|
| `make gates` auf `35c7aa8` | Exit 0; a-check `gesamt: 0 Befund(e)`; d-check `607 Datei(en), 0 Befund(e)`; baseline-verify OK; migration-check ok; zeichen `1198 Datei(en)`; `git status` danach sauber |
| `make groessenmarke` | `slice-v2-query-filter.md: ok, Messung 2546 von Marke 4600` |
| `git diff --numstat 80d7910c..HEAD` über `urlparserv2/` | 2506 + 40 = 2546, gleich Plan §4; Produktcode 469 Zeilen (356 + 113) |
| Grundlinie (Wegwerfkopie, `TestFl`) | grün |
| Probe-Pakete (eigen, am unveränderten Code) | 3 von 3 grün, 0 Abweichungen: Aufruf (jeder Lauf aus 1 bis 3 von 135 Zeichen vor `(1)`, rund 2,48 Mio. Läufe, Orakel nach `SPEC-164` Nr. 4 c 1); V4-Form (8 Vorlagen × jede Stelle × 23 Zeichen, Paare von Stellen bei Datum und Uhrzeit mit 13 Zeichen, Guid-Löschung und -Einfügung; Orakel nach Nr. 4 b); `INF`/`NaN` (alle Folgen bis 4 Zeichen aus 13, je links und rechts) |

## Plan-vs-Code-Diff

| Plan sagt | Code sagt | Urteil |
|---|---|---|
| Kein V4-Import ([ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md)) | `filter.go` importiert `strings`, `unicode`, `unicode/utf8`, `model`; `filter_token.go` nur `strings`; a-check 0 Befunde; kein `urlparser`, `csdl`, `jsoncodec` | konform |
| Tabelle *Zusage · Fänger* 37 Zeilen | Zeilen 1 bis 37, Nummern lückenlos | konform |
| F-5: `SPEC-173` Nr. 5 ohne Widerspruch | Satz „lesbar ist“ ersetzt durch „der Form nach beginnt … die Form entscheidet, nicht die Lesbarkeit“; Historie-Zeile 2026-11-05; kein weiteres „lesbar“ zur V4-Form in `SPEC-173`; Code folgt der Form | konform |
| F-8: Kopf nennt `SPEC-164` Nr. 1 bis 4 und 7 bis 9, Nr. 6 mit Zuordnung | gleiche Menge wie `SPEC-173` Nr. 5 *Der Leser* | konform |
| `SPEC-164` Nr. 1 bis 4, 7 bis 9 | Code gegen den Wortlaut gelesen: Kodierung vor Grammatik, Leerzeichen als einziger Leerraum, Rangfolge und Linksassoziativität, Schlüsselwörter ASCII, Wortlauf-Folge (Aufruf, `$`, `@`, Pfad, qualifiziert, `true`/`false`/`null`, `INF`/`NaN`), Tiefe 32 `zu_viele`, Literal 512 `zu_lang`, kein Zustand in Instanz und Paket | konform |
| `SPEC-173` Nr. 5 (Pfad im Ausdruck `filter`, Form statt Lesbarkeit, erster Verstoß entscheidet) | `glieder` liefert für `/` und `.` `filter`; `laufLiteral` prüft `formVonV4` nach der Zahl und `literal_typ`; Probe-Pakete grün | konform |
| Marke 4600 (Metrik `numstat`) | Gate 2546 | konform |
| Keine Suppression (`AGENTS.md` §3.2) | 0 `nolint` in den neuen Dateien | konform |

## DoD

| Punkt | Zustand |
|---|---|
| 1. Leser mit Grammatik, Fehlern, Zuordnung nach `SPEC-173` Nr. 1 | erfüllt (Code und Gates) |
| 2. Randformen je Zelle mit Ort und Fänger, Generator über die Klasse | **erfüllt mit Einschränkung**: das Raster hat keine leere Zelle, die Fänger decken die genannten Elemente; drei Klassen sind nicht gedeckt (V-1 bis V-3) |
| 3. Nebenläufigkeit (Zeilen 32, 33) | erfüllt, 30 von 30 (unten) |
| 4. `make gates` grün | erfüllt (eigener Lauf) |
| 5. Review durchgeführt | erfüllt |
| 6 bis 11 (Doku, Handbuch entfällt mit Grund, Closure-Notiz, Register, Ausgänge, Paarungen) | offen, Sache der Closure; §7 ist leer; kein Verstoß vor der Closure |

## Fixes des Reviews gegen die Klasse (`AGENTS.md` §3.17)

| Befund | Klasse (Wortlaut der Zusage) | Eigene Mutanten an Elementen, die weder Review noch Implementer nannten | Urteil |
|---|---|---|---|
| F-1 | `SPEC-164` Nr. 4 c 1: **jeder** Wortlauf unmittelbar vor `(` (und nicht `not`) ist ein Aufruf, vor jeder Namensprüfung | 14 Mutanten: grün (6) Aufruf nicht für `null`, nicht für `false`, nicht für `INF`/`NaN`, nicht für `and`/`or`/`eq`, nicht für Läufe mit `_` am Anfang, nicht für Läufe mit Nicht-ASCII-Buchstaben am Anfang; rot (8): einstelliger Name, Großbuchstabe am Anfang, Aufruf hinter `/`, hinter `.`, hinter dem Schalter `true`/`false`/`null`, `$` und `@` einzeln ausgenommen, Läufe mit `/` am Ende | **nicht geschlossen** (V-1) |
| F-2 | `SPEC-173` Nr. 5 c / `SPEC-164` Nr. 4 b: die Form von V4 Position für Position, Länge, Trenner, Ziffernart | 28 Mutanten: grün (15) Ziffer an Stelle 0 des Datums, Ziffer an Stelle 0 der Uhrzeit (beide nur mit `+`/`.`/`-` am Anfang sichtbar), Hex-Prüfung nimmt `g`, `G`, `:`, `/`, `@` oder `g` bis `w` an, Trenner `-` der Guid nimmt `_` an, Trenner des Datums an Stelle 4 oder 7 nimmt `_` bzw. `.` an, Trenner `:` der Uhrzeit nimmt `.` an, Ziffernprüfung nimmt `:`, `/` oder Unicode-Ziffern (`U+0663`) an; rot (13): die übrigen Einzelstellen des Datums und der Uhrzeit (Stellen 1 bis 3, 5, 6, 8, 9; 1; 3, 4), Hex-Prüfung an Stelle 9, 12, 14, 35, Länge 35 bis 36 und 36 bis 37 | **nicht geschlossen** (V-2) |
| F-3 | `SPEC-164` Nr. 4 c 7, Nr. 4 b: `INF`, `NaN`, `-INF` genau so; sonst Name bzw. `literal`/`filter` | 9 Mutanten: grün (4) `-iNf` als `-INF` (Literal-Lauf und Fall `literal_typ`, je einzeln), `+INF` als `literal_typ`, `-NaN` als `literal_typ`; rot (5) `NAn`, `iNF`, Anfang `NaN`, `N?N`, `EqualFold("-INF")` | **nicht geschlossen** (V-3) |
| F-4 | `SPEC-164` Nr. 3: Schlüsselwörter ohne Beachtung der ASCII-Schreibweise; Zeichen jenseits von ASCII mit Klein-/Großschreibungs- oder Faltungsbild in ASCII | 8 Mutanten: rot (6) Faltung nur `U+017F`, nur `U+0130`, nur `U+0131`, Klein nach Groß am Operator, `EqualFold` im Wortlauf, Klein nach Groß im Wortlauf; grün (2) Vollbreiten-Buchstaben (`U+FF21` bis `U+FF5A`) als ASCII gefaltet; `not` mit Klein-nach-Groß (äquivalent: `not` hat kein `i`, `s`, `k`) | **geschlossen** für die Menge des Plans (alle Zeichen mit einfachem Klein-/Großschreibungs- oder Faltungsbild in ASCII, 464 Elemente); die Menge der Kompatibilitätsfaltung (Vollbreite, Ligaturen, hochgestellt) ist unbeschränkt und steht nicht als Prüfgrenze im Plan (V-4) |
| F-5 | Satz in `SPEC-173` Nr. 5 | `grep` über Spezifikation und Plan | geschlossen |
| F-6, F-7, F-8 | Plan §3 Zeilen 26, 27; Kommentar; Kopf | gelesen | geschlossen |

Mutanten ohne Wirkung am Code oder gewertet als äquivalent: `not` mit Klein-nach-Groß; unter Zeile 18 die Mutanten „Datum als Zeitpunkt“ (im V2-Umfang gibt es kein `LiteralDatum`, `grep` im Produktcode 0 Treffer) und „Nullen am Ende der Zahl abgeschnitten“ (die Literal-Funktion normalisiert vorher).

## Zeile 18 (Gleichheitsprobe) nachgemessen

Filter-eigene Abweichungen am Knoten, je Art der Literale: Zeichenkette (`''` zu `'`) rot (5 Tests), Wahrheitswert groß rot, negative Zahl ohne Minus rot, Zeitpunkt ohne `Z` rot; alle in `TestFlLiteraleWieSchluessel`. Abweichung in der gemeinsamen Funktion (Wahrheitswert groß in `literal.go`): `TestFlLiteraleWieSchluessel` bleibt grün (beide Seiten gleich), der Plan sagt das; rot sind `TestKoerper` und `TestSchluesselLiterale` (Teil A) bzw. mit `-run TestFl` `TestFlLiteralEnde`. Die Aussage des Plans ist belegt.

## Zeilen 32 und 33 (Nebenläufigkeit), 30 Prozesse je Zeile und Modus

Messender: der Verifier. Test-Binary einmal übersetzt (`-race` bzw. ohne), `-test.run '^TestFlGleichzeitig$' -test.count=1` je Prozess, Image `odata-recorder:race`, `--cpus=4`, Maschine unter fremder Last.

| Lauf | mit Detektor | ohne Detektor |
|---|---|---|
| unmutiert | 30 von 30 grün, 0 `DATA RACE` | 30 von 30 grün |
| Zeile 32: Tiefenzähler als Feld der Instanz, je Aufruf auf 0 | 30 von 30 rot, 30 `DATA RACE` | 30 von 30 rot (Wirkungs-Fänger) |
| Zeile 33: `map`-Cache als Variable des Pakets, bei Fehlen geschrieben | 30 von 30 rot, 30 `DATA RACE` | 30 von 30 rot |

Bestätigt die 30 von 30 des Plans (§3, Zeilen 32 und 33).

## Rohbytes (`AGENTS.md` §3.18)

Gelesen mit `LC_ALL=C grep -P`: alle 18 Dateien von `80d7910c..HEAD` (Plan, Roadmap, `welle-odata-v2.md`, Review, `spec/spezifikation.md`, `filter.go`, `filter_token.go`, 9 `fl_*_test.go`, `literale_test.go`). Folge Backslash-`u`-vier-Hex: 0 im Diff (die 5 Treffer in `spec/spezifikation.md` sind Bestand), BOM 0, NBSP 0, Nullbreite, Zeilentrenner und U+0085 0, Steuerzeichen und `\r` 0, ungültiges UTF-8 0, jede Datei endet mit Zeilenvorschub, `git diff --check` leer; `make zeichen-check` grün. Nicht-ASCII nur Umlaute und `§ ² µ · × ÷ é`. Dieser Bericht ist nach dem Schreiben gegengelesen (Datei `docs/reviews/2026-11-05-slice-v2-query-filter-verifikation.md`).

## Findings

| ID | Kategorie | Befund | Quelle | Pfad |
|---|---|---|---|---|
| V-1 | MEDIUM | Der Fix von F-1 deckt die Klasse „jeder Name vor `(`“ nicht. `TestFlAufrufWortlaeufe` führt Zeichen in der Mitte, die Namen der Liste und `true`; sechs Mutanten bleiben grün, weil Aufruf-Namen `null`, `false`, `INF`/`NaN`, `and`/`or`/`eq`, Namen mit `_` am Anfang und Namen mit Nicht-ASCII-Buchstaben am Anfang nicht vorkommen. Szenario: `null(1)`, `false(1)`, `INF(1)`, `and(1)`, `_x(1)`, `ärger(1)` müssen `filter` sein (`SPEC-164` Nr. 4 c 1 geht Nr. 4 c 7 und 4 d voraus); die Mutanten liefern `wert`. **Der Code ist richtig** (Probe über 2,48 Mio. Läufe: 0 Abweichungen); es fehlt der Fänger. | `SPEC-164` Nr. 4 c 1; `AGENTS.md` §3.17 | `src/adapters/driven/odata/urlparserv2/fl_menge_test.go` · `TestFlAufrufWortlaeufe`, `TestFlAufruf` |
| V-2 | MEDIUM | Der Fix von F-2 deckt die Klasse „Form von V4 Position für Position“ nur für das Zeichen `x`. 15 Mutanten bleiben grün: (a) die Ziffer an Stelle 0 des Datums und der Uhrzeit (`x` am Anfang macht den Lauf zum Wortlauf, die Form wird nie geprüft; sichtbar nur mit `+`, `.` oder `-` am Anfang: `+020-01-01`, `.2:30` müssen `literal` sein), (b) die Nachbarn der Zeichenklassen (`g`, `G`, `:`, `/`, `@` als Hex-Ziffer; `:`, `/`, `U+0663` als Ziffer), (c) der Trenner anderer Art (`_` statt `-`, `.` statt `:`). Der Fänger wählt als Ersatz überall `x`. **Der Code ist richtig** (Probe: 8 Vorlagen × jede Stelle × 23 Zeichen, dazu Paare von Stellen: 0 Abweichungen). | `SPEC-164` Nr. 4 b; `SPEC-173` Nr. 5 c; `AGENTS.md` §3.15, §3.17 | `src/adapters/driven/odata/urlparserv2/fl_menge_test.go` · `TestFlV4FormPositionen` |
| V-3 | MEDIUM | Der Fix von F-3 deckt „`INF`, `-INF`, `NaN` genau so“ nur für `inf`/`nan` ohne Vorzeichen. Vier Mutanten bleiben grün: `-iNf` (eine andere Schreibweise von `-INF`) als `-INF` gelesen (an beiden Stellen des Codes), `+INF` als `literal_typ`, `-NaN` als `literal_typ`. Szenario: `Id eq -iNf` muss `filter` (unäres Minus) sein, `+INF` muss `literal`, `-NaN` `filter` sein; die Mutanten liefern `literal_typ`/`literal`. **Der Code ist richtig** (Probe: alle Folgen bis 4 Zeichen). | `SPEC-164` Nr. 4 b, 4 c 7; `AGENTS.md` §3.17 | `src/adapters/driven/odata/urlparserv2/fl_menge_test.go` · `TestFlInfNanSchreibweise`; `fl_lexik_test.go` · `TestFlMinusLaeufe` |
| V-4 | LOW | Die Menge zu F-4 ist als „Zeichen mit einfachem Klein-/Großschreibungs- oder Faltungsbild in ASCII“ gefangen (464 Elemente, alle Mutanten rot). Die Kompatibilitätsfaltung (Vollbreite, Ligaturen, hochgestellte Buchstaben) bleibt ungeprüft (Mutant „Vollbreite als ASCII“ grün); der Plan nennt dafür keine Prüfgrenze. Folge im Betrieb: keine, wenn der Code kein Normalisieren kennt (er kennt keines); es fehlt das benannte akzeptierte Negativ (`AGENTS.md` §3.17, Menge unbeschränkt). | `AGENTS.md` §3.17, §3.15 | `docs/plan/planning/done/slice-v2-query-filter.md` · Zeile 6 |
| V-5 | INFO | Zeilen 32 und 33 gefangen: 30 von 30 Prozessen je Modus und Zeile, unmutiert 30 von 30 grün. | `AGENTS.md` §3.10 | `TestFlGleichzeitig` |
| V-6 | INFO | Der Zweig liegt hinter `main` (`git diff main..HEAD` zeigt Dateien, die `main` neu trägt); Merge-Verträglichkeit nicht geprüft, Sache des Integrationsschritts. | Merge | — |

0 HIGH, 3 MEDIUM, 1 LOW, 2 INFO. Alle drei MEDIUM sind reine Fänger-Lücken: Produktcode und Spezifikation sind in den Probe-Paketen über die Klassen ohne Abweichung.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `make gates` | geprüft, ohne Befund |
| [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) (kein V4-Import) | geprüft, ohne Befund |
| `SPEC-173` Nr. 5 mit Schärfung und Berichtigung, `SPEC-164` Nr. 1 bis 4, 7 bis 9 | Code gegen Wortlaut und Probe-Pakete: ohne Befund |
| F-5, F-6, F-7, F-8 des Reviews | geschlossen |
| Zeile 18, Zeilen 32 und 33 | nachgemessen, ohne Befund |
| Größenmarke 2546 von 4600 | ohne Befund |
| Rohbytes, Suppression | ohne Befund |

## Ergebnis

**Konform bis auf die Klasse dreier Fixes.** DoD 1, 3, 4, 5 sind erfüllt, DoD 2 mit Einschränkung; Plan und Code stimmen überein, Gates grün, Rohbytes unverfälscht, die Nebenläufigkeit ist 30 von 30 gefangen. F-4 ist für die Menge des Plans geschlossen, F-1 bis F-3 nicht gegen die Klasse (V-1 bis V-3): je eine Schleife über die fehlenden Elemente (Namen `null`, `false`, `INF`, `NaN`, `and`, `_x`, Nicht-ASCII-Anfang vor `(`; Stelle 0 mit `+`/`.`/`-` und Nachbarzeichen der Zeichenklassen; Schreibweisen von `-INF`, `+INF`, `-NaN`) schließt sie, ohne Produktcode oder Spezifikation zu berühren. An den Planner: V-1 bis V-3 vor der Closure zurück an den Implementer oder als akzeptiertes Negativ mit Grund und Folge in §3; V-4 als Prüfgrenze in Zeile 6 benennen. Diese Prüfung wird nicht wiederholt; Mutanten und Probe-Paket liegen in `scratchpad/ver-filter/`.

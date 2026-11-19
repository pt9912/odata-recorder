# Nachverifikationsbericht 2: slice-expand-parser — 2026-11-19

**Art:** Nachverifikation der Fix-Runde nach der Nachverifikation (Bauen wir es richtig? — gegen die Klasse der Zusage, `AGENTS.md` §3.17), Modul 11, frischer Kontext. Nicht das Review, nicht die Validierung. Der Bericht `2026-11-18-slice-expand-parser-nachverifikation.md` bleibt unverändert; dieser löst nur dessen Fix auf.

**Gegenstand:** Zweig `slice-expand-parser`, `HEAD` `d1fd5cea` (Merge von `main`). Fix in `7327ba0f` (fünf neue Testdateien `expand_stelle_test.go`, `expand_dublette_test.go`, `expand_folge_test.go`, `expand_groesse_test.go`, `expand_escape_stellung_test.go`) und `4ec1aba2` (Plan §3 *Fix-Runde nach der Nachverifikation*, §4 Messpunkt). Ausgangslage: vier MEDIUM N-1 bis N-4.

**Bezug:** `LH-FA-16` (Unterpunkt a), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Verträge, gegen die geprüft wurde:** `SPEC-262` Nr. 3 bis 6 und 8, `SPEC-263` (nur Abgrenzung: der Slice berührt `PruefeExpand` im Fix nicht), `SPEC-160` Nr. 5 und 6, `SPEC-210`, `SPEC-128`, `AGENTS.md` §3.10, §3.15, §3.17, §3.18, Plan `slice-expand-parser` §3 (Fix-Runde), §4.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-19

## Urteil

**Nicht bestanden.** 0 HIGH, 2 MEDIUM (R-1, R-2), 0 LOW, 4 INFO. N-1 und N-3 sind gegen die Klasse gedeckt. N-2 und N-4 sind es teilweise: die genannten elf Mutationen sind rot, aber an Elementen der Menge, die weder der Befund noch der Plan nennen, bleiben fünf von sechzehn eigenen Mutationen im ganzen Paket grün (`v05`, `v07`, `v08`, `v11`, `v12`). Der Code weicht an keiner Stelle von `SPEC-262` ab, soweit die Tests ihn führen; ob er an den ungefangenen Stellen abweicht, ist nicht probiert (Abschnitt *Ungeprüft*). Die Lücken sind klein: zwei Schleifen mehr oder zwei benannte Prüfgrenzen mit Grund und Folge.

## 1. Belege der Sensoren

- `make gates` einmal vollständig selbst gefahren, am Stand `d1fd5cea`: **Exit 0** (Ausgabe endet mit `zeichen: ok, 1966 Datei(en) geprüft`; die Stufen `test`, `race`, `lint`, `a-check`, `docs-check` und alle übrigen liefen durch). Die Behauptung des Implementers ist damit belegt, nicht übernommen.
- `make groessenmarke`: `Messpunkt slice-expand-parser.md: zeilen=30 nebenlaeufig=1 fakes=0 produktcode=150 schaetzung=2555 marke=3500 messung=2268`, ok. Die Zahl **2004** gilt für den Stand vor der Fix-Runde (`ef038aa2`); der Plan §4 nennt beide Zeilen und die Messung nach der Fix-Runde, **2268 von 3500**, stimmt mit meiner überein. Auslastung 2268 / 3500 = 0,65, über 0,60: der Plan sagt, dass der Planner es in der Closure dem Architect meldet (`SPEC-128` Nr. 9). Die Messung liegt unter der Marke; der Produktcode des Slice (`expand.go` u. a.) ist in der Fix-Runde nicht gewachsen.
- Nachgerechnet gegen den Plan: N-1 `4 × (2+6+42+132+600+992) = 7096`, plus `81 × 6 = 486`, zusammen 7582; N-2 `92 + 36 + 9 = 137`; N-3 `121 + 1331 = 1452`; N-4 1968 und 88 und N-1 Dekodierung 420 (die Tests zählen selbst gegen diese Zahlen, `exZaehlt`). Jede Zahl der Tabelle des Plans steht als Konstante im Test.

## 2. Der Merge von `main` verändert den Slice-Code nicht

`git diff 4ec1aba2 HEAD -- src` nennt nur `src/bootstrap/*` (Konfiguration des Modus hybrid, anderer Slice); unter `src/adapters/driven/odata/` und `src/hexagon/` gibt es **keine** Änderung. `spec/spezifikation.md` ist im Merge geändert (57 Zeilen), aber keine Zeile von `SPEC-262` oder `SPEC-263` (`git diff … | grep SPEC-26[23]` meldet 0 Zeilen mit Änderung). Der Plan trägt im Merge zwei Zeilen: die Übergabezeile *Stand der Übergabe* (Nr. 356). Der Merge ist für den Slice neutral.

## 3. Plan-vs-Code-Diff und DoD

- Die Fix-Runde ändert nur Tests und den Plan (`git diff 795996ba 4ec1aba2 --name-only | grep -v _test` nennt allein den Plan). Kein Produktcode, wie der Auftrag verlangt. Die fünf neuen Tests entsprechen den fünf Zeilen der Tabelle in Plan §3 (Fänger, Mengen, Zahlen).
- Die Köpfe der Tests nennen ihre Mengen und ihre offenen Grenzen (`expand_escape_stellung_test.go`: *Offen bleibt: ein anderes Escape als diese zehn*; `expand_groesse_test.go`: *Offen bleibt: jede Gliederzahl zwischen den genannten Stufen*). Das stimmt mit dem Plan überein (`AGENTS.md` §3.11); für die beiden Lücken R-1 und R-2 sagt weder Kopf noch Plan etwas.
- Rohbytes (§3.18): in den fünf neuen Dateien kein BOM, kein NBSP und keine Folge Backslash-`u`-vier-Hex (`LC_ALL=C grep -nP`); gefunden sind nur Umlaute in Kommentaren. Dieser Bericht ist nach dem Schreiben mit demselben Befehl gegengelesen.
- DoD (`slice-expand-parser.md` Zeilen 69 bis 81): alle Häkchen offen. Das ist der erwartete Stand vor der Closure, die dieser Lauf nicht schreibt; die Übergabezeile nennt als offen: diese Nachverifikation, Closure, `make slice-mv`, Integration. `make gates` (Zeile 72) ist belegt (Abschnitt 1).

## 4. Konformität mit SPEC-262

Das Orakel `exOrakel` (`expand_hilfe_test.go`) gibt `SPEC-262` Nr. 3 und 4 wieder: leer, `(`, Zahl der `,` ab 32, Elementfolge mit Rang `wert` vor `501`, Dublette bytegleich vor Klassifikation. Es nutzt keinen Code des Lesers und ist in Rangfolge und Klassen deckungsgleich mit dem Text (Nr. 3 a bis d, 4 a bis e). Die Zeichenklassen des Orakels sind eine endliche Tabelle (Buchstaben a bis z, A bis Z, `_`, drei Nicht-ASCII-Buchstaben, Ziffern 0 bis 9 und eine Nd-Ziffer); das trägt `SPEC-262` Nr. 4 a, ohne jede Unicode-Kategorie zu führen. Das ist eine Prüfgrenze der früheren Runde und kein neuer Befund.

## 5. Der Fix gegen die Klasse

Mengen aus dem Wortlaut von `SPEC-262`, nicht aus dem Befund: Stelle des Elements (Nr. 3 d „gleich wo er steht“, Nr. 6 bis 32 Elemente) × Art; Dublette (Nr. 4 e „bytegleich“) × Form × Länge × Stelle des Paares; Folge der Segmentarten (Nr. 4 c „alle Segmente“); Größe (Teile, Segmente, Bytes) × Stelle × Länge des Werts (Nr. 6 „allein durch `rawQuery`“, 8192 Bytes) × Zahl der Elemente.

Eigene Mutationen an Elementen, die weder der Nachverifikationsbericht noch die Fix-Runde (`n1-…` bis `n4-…`) nannte, gefahren mit `sh tools/harness/mutationslauf.sh` unter `flock /tmp/odata-recorder-heavy.lock`, Image `odata-recorder:race`, ohne Netz und Mount, `--memory 2g`, Klasse `andere` (deterministisch, keine Zusage über Nebenläufigkeit). Ein Lauf, 16 Mutationen, Dauer 77,8 s, Host-Last 4,3 bis 6,0 von 4 CPUs (Messender dieser Verifier, eine Messung, Größenordnung). Liste: `v01` bis `v16` in `expand.go`.

| ID | Menge | Mutation | Ergebnis | roter Test |
|---|---|---|---|---|
| v01 | Dublette × Stelle | Dublette an Stelle 4 nicht erkannt | rot-paket | andere `TestExpand*` des Pakets (nicht `TestExpandDublette`) |
| v02 | Dublette × Stelle | Dublette nur am letzten Element erkannt | rot-paket | dto. |
| v03 | Dublette × Stelle | Dublette nur gegen das erste Element | rot-paket | dto. |
| v04 | Dublette × Stelle | Dublette nur vor Stelle 25 erkannt | rot-paket | dto. |
| v05 | Dublette × Form | Dublette nicht erkannt, wenn das Element `$count` enthält | **gruen** | — |
| v06 | Dublette × Form | Dublette nicht erkannt bei Nicht-ASCII | rot-paket | dto. |
| v07 | Dublette × Länge | Dublette nur bei Elementen unter 150 Bytes | **gruen** | — |
| v08 | Dublette × Länge | Dublette nur bei Elementen unter 1000 Bytes | **gruen** | — |
| v09 | Folge × Länge 4 | Pfad `$value/x/y/$value` ungültig | rot-paket | dto. (Segmentzahl) |
| v10 | Folge × Stelle im Pfad | `$value` an Segment 7 ungültig | rot-paket | dto. |
| v11 | Zahl × Länge des Werts | Elemente nur in den ersten 5000 Bytes gezählt | **gruen** | — |
| v12 | Zahl × Länge des Werts | Elemente nur in den ersten 7000 Bytes gezählt | **gruen** | — |
| v13 | Stelle × Art | `*` ab Stelle 20 beendet das Lesen | rot | `TestExpandStelleImWert` |
| v14 | Stelle × Art | `*` ab Stelle 27 übergangen | rot | `TestExpandStelleImWert` |
| v15 | Teile × Segment | Name mit sieben oder mehr Teilen unter 300 Bytes ungültig | rot-paket | andere Tests (Teile) |
| v16 | Länge des Werts | `(` nur in den ersten 7990 Bytes gesucht | rot | `TestExpandLaengeVorKlammerUndZahl` |

Summe: 3 rot, 8 rot-paket, 5 gruen. `rot-paket` zählt als rot (`SPEC-210` Nr. 6 c); es zeigt aber, dass bei v01 bis v04, v06 und v09 bis v10 der neue, zuständige Test **nicht** die Mutation fängt, sondern ein anderer, älterer Test des Pakets; der Fix hält die Menge nur über diese. Das ist kein Befund (die Zusage ist gefangen), aber die neuen Dubletten- und Folge-Tests tragen die Stelle des Paares und die Länge 4 nicht selbst.

### Je Befund

| Befund | Urteil | Begründung |
|---|---|---|
| N-1 Stelle × Art | **gedeckt** | `TestExpandStelleImWert` (7582): v13 und v14 rot am zuständigen Test; jede Stelle der Listenlängen 2, 3, 7, 12, 25, 32; die Listenlängen dazwischen sind im Plan als akzeptiertes Negativ mit Grund und Folge benannt. |
| N-2 Dublette × Form × Länge | **teilweise**, R-1 | Gleichheit mit `$ref`, `*`, Namen bis 128 und Pfaden bis 40 Segmente ist geführt (v01 bis v04 und v06 rot). Ungefangen: `$count`/`$value` im Dublettenpaar (v05) und gleiche Elemente über 130 Bytes (v07, v08). |
| N-3 Folge der Segmentarten | **gedeckt** | `TestExpandSegmentFolge` (1452: alle Paare und Tripel der 11 Arten); Folgen der Länge 4 und die Stelle im Pfad fängt das übrige Paket (v09, v10, v15 rot-paket). |
| N-4 Größe × Stelle × Länge des Werts | **teilweise**, R-2 | `(` am Ende bis 8000 Bytes (v16 rot) und Elementgröße an jeder Stelle gedeckt; die Zählung der Elemente (`zu_viele`) ist nur bis rund 4400 Bytes geführt (v11, v12 gruen). |
| N-1 Dekodierung | gedeckt | `TestExpandEscapeInJederStellung` (420); der Bericht fand keine grüne Mutation, die Fix-Runde fügt keine neue an. Offen bleibt ein anderes Escape als die zehn (benannt). |

## 6. Findings

| ID | Kategorie | Befund | Quelle | Pfad | Klasse |
|---|---|---|---|---|---|
| R-1 | MEDIUM | Die Dublette ist für drei Dimensionen der Form nicht geführt: Pfade mit `$count` oder `$value` (`a/$count,a/$count`, Mutation `v05` grün: liefert `501` statt `wert`) und gleiche Elemente über 130 Bytes (`v07`, `v08` grün: ein Pfad mit 100 Segmenten zweimal ist `501` statt `wert`). `SPEC-262` Nr. 4 e sagt „bytegleich“, jede Form und Länge; Nr. 4 c nennt `$ref`, `$count`, `$value` gleichrangig. Der Plan führt Dubletten bis 128 Zeichen und 40 Segmente und nur `$ref`; die Länge darüber (verschiedene Pfade bis 100 Segmente sind geführt, gleiche nicht) und `$count`/`$value` sind keine benannte Prüfgrenze. Folge im Betrieb: ein Umbau der Gleichheitsprüfung (Hash, Normalisierung, Längenschwelle) kann gleiche lange Pfade als verschieden lesen und liefert dann `501` statt `400`, ohne dass ein Test rot wird. | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 4 c, 4 e | `expand_dublette_test.go` · `TestExpandDublette` (`formen`); Plan §3 Zeile N-2 | Menge Form (alle Segmentarten) × Länge (bis 8192) der gleichen Elemente; Schleife über die 11 `exSegmentArten` und Längen 130 bis 4000 deckt sie, unter einer Zehntelsekunde |
| R-2 | MEDIUM | Die Zählung der Elemente (`zu_viele`, `SPEC-262` Nr. 3 c) ist nur für Werte bis rund 4400 Bytes geführt (31 bis 34 Namen zu höchstens 128 Zeichen). Eine Zählung, die nur die ersten 5000 oder 7000 Bytes liest (`v11`, `v12`), bleibt im ganzen Paket grün: 33 Pfade zu je 160 Bytes (zusammen 5300 Bytes, unter der Grenze von 8192) sind dann `501` statt `zu_viele`. `SPEC-262` Nr. 6: „die Länge des Werts ist allein durch `rawQuery` begrenzt“. Der Plan nennt für die Länge des Werts nur `(`; die Zählung ist weder geführt noch eine benannte Prüfgrenze (derselbe Befund wie `nv13` des Berichts, bei einer anderen Länge). | `AGENTS.md` §3.15, §3.17; `SPEC-262` Nr. 3 c, 6 | `expand_groesse_test.go` · `TestExpandLaengeVorKlammerUndZahl` (Elemente bis 128 Zeichen); Plan §3 Zeile N-4 | Menge Zahl der Elemente (31 bis 34) × Länge des Werts (bis 8192) × Art der Elemente (Pfad statt Name) |
| I-1 | INFO | Die neuen Dubletten- und Folge-Tests fangen die Mutationen `v01` bis `v04`, `v06`, `v09`, `v10` nicht selbst; ein älteres Paar von Tests tut es (`rot-paket`). Die Zusage ist gefangen, die Zuordnung Zeile → Fänger der Plantabelle nicht in jeder Zeile eindeutig. | `AGENTS.md` §3.15 | Plan §3 Fix-Runde | — |
| I-2 | INFO | Auslastung der Marke 0,65 über 0,60 (Messung 2268 von 3500); der Plan hält die Meldung an den Architect für die Closure fest (`SPEC-128` Nr. 9). Die Marke gilt als eingehalten. | `SPEC-128` Nr. 9 | Plan §4 | — |
| I-3 | INFO | Das Orakel führt eine endliche Zeichentabelle statt jeder Unicode-Kategorie; für `SPEC-262` Nr. 4 a als Prüfgrenze der früheren Runde bekannt. | `AGENTS.md` §3.17 | `expand_hilfe_test.go` | akzeptiertes Negativ |
| I-4 | INFO | Der Plan nennt als Messender der Laufzeit der Tests den Implementer (Host-Last 17 von 4 CPUs, rund 0,15 s ohne Detektor); eine eigene Laufzeit wurde nicht neu gemessen, `make test-race` lief im Gate grün durch (Zeitgrenze von 10 Minuten je Paket, `SPEC-122` Nr. 5, nicht gefährdet). | `SPEC-122` Nr. 5 | `make test-race` | — |

## 7. Ungeprüft

- **Die Eingaben der ungefangenen Mutationen gegen den unveränderten Code.** `v05`, `v07`, `v08`, `v11`, `v12` zeigen nur, dass kein Test sie fängt. Ich habe keine Probe auf dem Stand `d1fd5cea` gefahren, die `a/$count,a/$count`, zwei gleiche Pfade mit 100 Segmenten und 33 Pfade zu 160 Bytes dem Orakel gegenüberstellt; die Lesung des Codes (`slices.Contains`, `strings.Count`) lässt keine Abweichung erwarten.
- Listenlängen zwischen den geführten (4 bis 6, 8 bis 11, 13 bis 24, 26 bis 31) bei `*` und Dublette: im Plan als akzeptiertes Negativ benannt, nicht als Befund gewertet.
- Zahlen von Teilen und Segmenten zwischen den Stufen über 300 (Probe P7, 170 550 Eingaben): benannte Prüfgrenze des Plans.
- Eine Laufzeitmessung des neuen Testpakets mit Detektor; die Raten der Zeile 23 sind laut Plan nicht neu gemessen (kein Code der Nebenläufigkeit berührt), und ich habe sie nicht nachgemessen.
- Dekodierung jenseits der zehn Escapes; die Unicode-Kategorien jenseits der Tabelle des Orakels.

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 0 |
| INFO | 4 |

## Übergabe an den Planner

Merge-blockierend nach der Regel der früheren Runden sind R-1 und R-2: je eine Schleife über die Menge (R-1: die 11 Arten von `exSegmentArten` als Pfad mit Längen von 130 bis 4000 Bytes, je gleich und je verschieden, an den Stellen 0 und letzte; R-2: 33 Elemente aus Pfaden und Teilen mit Gesamtlängen 4500 bis 8190 Bytes, mit und ohne `(` am Ende) oder je ein akzeptiertes Negativ mit Grund und Folge im Plan. Beides kostet unter einer Zehntelsekunde. Der Bericht des Fixes nennt dann je Befund die Menge, die Zahl ihrer Elemente und eine Mutation an einem Element, das dieser Bericht nicht nannte (etwa Dublette nur bei Elementen mit `$value` ungeprüft oder Zählung nur bis 6500 Bytes). Die Finding-Klasse dieses Laufs: *die Menge ist je Achse geführt, aber die Achse „Länge des Werts bzw. des Elements“ endet an der Größe, die die Fixtures der Fix-Runde hatten, nicht an der Grenze von `rawQuery`*. N-1 und N-3 sind abgeschlossen; DoD-Häkchen und Closure stehen aus und gehören nicht zu diesem Lauf.

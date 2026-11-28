# Verifikation slice-betrieb-spans-nachweis (Record, Replay und Mock mit Spans und Zählern) — 2026-11-28

**Rolle:** Verifier (genau ein Lauf). **Stand:** Zweig `slice-betrieb-spans-nachweis`, HEAD `e065f5e3` (Fix-Runde `ca5e191e` nach dem Review `2026-11-28-slice-betrieb-spans-nachweis-review`). **Modell:** Sonnet 5.5.
**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-FA-01`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `SPEC-026` bis `SPEC-030`, `SPEC-230`, `SPEC-290`, `SPEC-294` bis `SPEC-296`; `welle-betrieb` §3.
**Gegenstand:** `src/bootstrap/telemetrie_betrieb_test.go` (nur Test, kein Produktcode im Zweig) und der Plan mit DoD, BS1 bis BS11, §3 und §6.

## Urteil

**Bestanden.** Die DoD ist in ihren Liefer-Punkten, im Gate-Lauf und im Review-Punkt erfüllt; offen sind allein die Punkte der Closure.
Das Closure-Kriterium von `welle-betrieb` §3 zu Record, Replay und Mock ist im Wortlaut durch `TestBetriebSpansUndZaehler` belegt.
Die Fix-Runde deckt die Klassen F-1, F-2 und F-3. Kein Verhalten widersprach den Spec-Stellen.

**Auflage** (an die Closure, nicht blockierend für Merge und Verifikation):

1. **Wortlaut eines Negativs (Planner).** §3, *Nicht jeder Weg trägt jeden Ort*, sagt als Folge „die Stelle fängt allein ein anderer Weg“. Für 2 der 53 gefahrenen Stellen gibt es keinen anderen Weg: `mock/lesen.go:230` und `mock/schreiben.go:275` erreicht allein ein Weg mit dem festen Schlüssel `1`; eine Ausgabe des Pfads dort bleibt grün (selbst gesehen). Im selben Absatz zählt die 24 den Weg `existiert` mit, dessen Schlüssel allein im Body steht (POST auf eine Menge); Literale im Pfad sind es 23.

**Hinweise:**

- Wer das Kriterium der Welle belegt, nennt `TestBetriebSpansUndZaehler`. `TestBetriebFehlerwegeOhneWerte` löst weitere Codes aus und prüft keine Zähler; er ist nicht Teil dieses Belegs. Eine Probe in der Wegwerfkopie (Zählervergleich in diesen Test eingesetzt) blieb grün: in 9 Prozessen stehen die Zähler gleich den erhaltenen Fehlerobjekten, 22 verschiedene Codes. Das ist ein Befund dieses Laufs, kein Fänger im Repo.
- Closure offen: die Häkchen in §2, die Ausgänge in §6 (zwei Zeilen stehen auf *offen*: Raster und Abweichung F-3), §7, Register, die Zeile der Tabelle in `welle-betrieb` (dort noch „nicht erfüllt“). Die übrigen offenen Kriterien der Welle (etwa `make gegenproben`) berührt dieser Slice nicht.

## DoD

| Punkt | Ergebnis | Beleg |
|---|---|---|
| Ein Test findet die fünf Spans mit jedem Pflicht-Attribut; Canary als Query-Literal und in beiden Bodies in keiner Zeile der Senke | erfüllt | gelesen: Vergleich der ganzen Attributmenge je Anfrage, Eltern über `parent_id`, Menge der fünf Namen, Ankunft der Werte je Modus; Mutationen D1 bis D5, D7, D8 rot |
| Derselbe Test vergleicht je Modus die Zähler mit den Codes der Fehlerobjekte; mindestens einer je Modus | erfüllt | gelesen: Gleichheit der Mengen mit Anzahl gegen erhaltene und erwartete Codes; `record` 2, `replay` 3, `mock` 3 Fehlerobjekte; Mutation D6 rot |
| `make gates` grün | erfüllt | Stempel, siehe unten |
| Review durchgeführt, Report liegt vor | erfüllt | `2026-11-28-slice-betrieb-spans-nachweis-review` |
| Closure-Notiz, Register, Ausgänge in §6, Paarungen | offen | Aufgabe der Closure |

## Tabelle Zusage · Fänger gegen den Vertrag

BS1 bis BS11 stimmen mit `SPEC-026` bis `SPEC-030`, `SPEC-230` Nr. 8, `SPEC-290` Nr. 4, `SPEC-294` Nr. 4 und 5, `SPEC-295` Nr. 4 und `SPEC-296` Nr. 4 überein; die erwarteten Werte des Tests (etwa `incomplete` ohne Antwort des Upstream, `complete` beim Einzel-Lookup ohne Eintrag, kein Dienst-Span vor der Annahme) entsprechen diesen Stellen.
Die Attributwerte jenseits davon führt der Test nicht; das Negativ dazu nennt sie.

## Das Kriterium der Welle im Wortlaut

| Teil | Beleg in `TestBetriebSpansUndZaehler` |
|---|---|
| „Ein Test durchläuft Record, Replay und Mock“ | eine Testfunktion, Composition Root, OData-Listener, drei Modi nacheinander |
| „findet die Spans `SPEC-026` bis `SPEC-030` mit ihren Pflicht-Attributen“ | Gesamtspan je Anfrage, Span des Dienstes darunter, am Ende genau die fünf Namen |
| „ohne Query-Literale und Bodies als Attribut“ | genaue Attributmenge und Suche der elf Werte in der Senke je Modus |
| „die Zähler je Fehlercode steigen für jeden Code, den der Test auslöst“ | `Snapshot().Errors` (je Modus ein neuer Adapter) gleich den Codes der Fehlerobjekte |

## Fix-Runde gegen die Klassen

**F-1 (Werte auf den Fehlerwegen): gedeckt.** Die Menge selbst gebildet: Zeilen der drei Dienste, die einen Fehler mit Code bilden. Sie ergibt 107, davon fahren die Fehlerwege 53 (`record` 9 von 12, `replay` 16 von 24, `mock` 28 von 71); die Zahlen aus §3 stimmen.
Probe in der Wegwerfkopie: an jeder erreichten Stelle einzeln einen Ort der Anfrage als Eintrag ausgegeben, 159 Läufe von `TestBetriebFehlerwegeOhneWerte`.

| Ausgegebener Ort | rot („die Senke trägt“) | grün |
|---|---|---|
| Query | 53 von 53 | keine |
| Body des Requests | 53 von 53 | keine |
| Pfad (Literal des Schlüssels) | 47 von 53 | 6: `mock/lesen.go:141`, `:230`, `:282`, `mock/schreiben.go:89`, `:251`, `:275` |

Von den 6 grünen haben 4 im Pfad keinen Schlüssel (Anfrage auf eine Menge); 2 sind die Stellen der Auflage 1.
Die 54 ungefahrenen Stellen sind nicht geprüft; das akzeptierte Negativ nennt sie.

**F-2 (Negativ zu den Attributwerten): gedeckt.** Nachgezählt: 4 + 5 + 6 + 4 + 1 + 4 + 3 = 27 Werte (`recorder.mode` ohne `hybrid`), geprüft 14, ungeprüft 13; die vier zuvor fehlenden Werte stehen im Text.

**F-3 (Abweichung ohne Zeile in §6): gedeckt.** §6 führt die Zeile mit Ausgang *offen* und Adresse Planner; die Spalte *Mutation* nennt HTTP-Adapter und Dienste.

## Eigene Mutationen

Alle am Produktcode in der Wegwerfkopie, je 3 Läufe beider Tests, ohne Panic und ohne Baufehler.

| ID | Mutation | Ergebnis |
|---|---|---|
| D1 | `mutation_type` `put` wird `patch` | rot 3 von 3, Spans unter dem Gesamtspan |
| D2 | `coverage_status` `partial` wird `complete` | rot 3 von 3 |
| D3 | Mutations-Span ohne Eltern | rot 3 von 3 |
| D4 | `capture_status` `incomplete` wird `failed` | rot 3 von 3 |
| D5 | `match_result` `kein_treffer` wird `nicht_gewaehlt` | rot 3 von 3 |
| D6 | `ODR-E7003` wird nicht gezählt | rot 3 von 3, Stand der Zähler |
| D7 | Status 204 im Span als 200 | rot 3 von 3, Gesamtspan |
| D8 | Methode `DELETE` im Span als `POST` | rot 3 von 3 |
| C1 | HTTP-Adapter gibt die Query bei jedem Fehlerobjekt aus | rot 3 von 3, beide Tests |
| C2 | HTTP-Adapter gibt den Pfad bei jedem Fehlerobjekt aus | rot 3 von 3 |
| C3 | HTTP-Adapter gibt den Pfad allein bei `ODR-E6010` aus | rot 3 von 3 |
| E1 | Mock-Dienst gibt den gespeicherten Wert bei `ODR-E6010` aus | rot 3 von 3 |
| B1 | Replay-Dienst setzt das Query-Literal auf Fehlerwegen als `match_result` | rot 3 von 3, „die Senke trägt“ |
| B2 | Record-Dienst setzt das Schlüssel-Literal als `session_id` nach dem Ende des Spans | rot 3 von 3 allein am Verwurf `zustand`; die Mutante gibt nichts aus und zählt nicht als Beleg für BS6 |

## Sensoren

| Sensor | Ergebnis |
|---|---|
| `make gates` | nicht neu gefahren: `.harness/state/gates-passed.head` gleich `e065f5e3`, `gates-passed.diffsha` gleich dem selbst gerechneten Hash des Arbeitsbaums, Stempel von 15:43 nach dem Commit von 15:19 |
| Beide Tests unverändert, `go test -count=20` | grün, 39,0 s |
| Größenmarke, `git diff --numstat a819be8e..HEAD` über den Pfad des Blocks | 842 von 1200 |
| `make gegenproben` | nicht gefahren (Auftrag) |

Ein Container (`--network none`, `--memory=4g --memory-swap=4g --cpus=4`, ohne Mount, Quellen per `git archive` und `docker cp`) unter `flock /tmp/odata-recorder-heavy.lock`; Last des Hosts 4,5 beim Start, 3,4 bis 6,6 am Ende; kein OOM; der Container ist entfernt. Ein erster Start scheiterte vor dem Anlegen des Containers an einem inzwischen ersetzten Image.

## Nicht geprüft

Die 40 und die 13 Mutationen aus §3 sind nicht einzeln wiederholt. Die Raten mit Detektor und die Laufzeit des Pakets unter `race` (422,4 s) sind nicht nachgemessen; dafür steht allein der eine Lauf von `test-race` im Gate-Nachweis.

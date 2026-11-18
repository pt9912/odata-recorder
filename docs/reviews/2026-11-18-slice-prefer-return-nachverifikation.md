# Nachverifikation slice-prefer-return

**Rolle:** Verifier (Nachverifikation nach Fix-Runden 2 bis 4). **Datum:** 2026-11-18. **Stand:** Zweig `slice-prefer-return`, HEAD `9e1d1126`, Basis der ersten Verifikation `9a6ee7dc`. **Bezug:** `LH-FA-07` (`LH-FA-07.b`), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), `SPEC-285` bis `SPEC-289`.
**Plan:** [slice-prefer-return](../plan/planning/done/slice-prefer-return.md). **Vorberichte:** [Verifikation](2026-11-18-slice-prefer-return-verifikation.md), [Review](2026-11-18-slice-prefer-return-review.md), [Runde 2](2026-11-18-slice-prefer-return-review-2.md), [Runde 3](2026-11-18-slice-prefer-return-review-3.md), [Runde 4](2026-11-18-slice-prefer-return-review-4.md). **Modell:** claude-sonnet-5-5.
Die Reviews sind nicht als Beleg gelesen, nur als Liste der offenen Punkte. `make gates` ist nicht neu gefahren (Auftrag: grün auf `9e1d1126`, Ausgang 0); gefahren sind `make docs-check`, `make groessenmarke`, eigene Mutationen und eigene Zeitmessungen. Host geteilt mit anderen Läufen (Load average 10 bis 23, 20 Kerne).

## Urteil

**PASS, ohne Auflage.** Ein LOW (Zeitreserve von `TestLiesRueckgabeErschoepfend` unter `-race` bei knapper CPU), zwei INFO.
Der Produktcode ist seit `9a6ee7dc` verhaltensgleich, V-1 bis V-3 sind behoben, die offenen Punkte aus Review Runde 4 (MEDIUM, zwei LOW) sind im Code und im Plan behoben, und neun eigene Mutationen an Backslash- und Paar-Stellen, die weder Review noch Implementer nannten, ergeben sieben rot im benannten Test und zwei grün aus nachgewiesenem Grund (gleichwertig).

## Findings

| Nr. | Schwere | Ort | Befund |
|---|---|---|---|
| N-V1 | LOW | `src/hexagon/model/rueckgabe_test.go` · `TestLiesRueckgabeErschoepfend` gegen die Zeitgrenze von 10 Minuten je Paket | Die Laufzeit des Tests und des Pakets unter `-race` ist gemessen (Tabelle unten): bei 4 CPUs 96 bis 111 s (Test) und 133 s (Paket), bei **1 CPU 352 s (Paket)**, die Grenze ist 600 s. Reserve 4,5-fach bei 4 CPUs, 1,7-fach bei 1 CPU. Die Gate-Stufe `race` fährt `go test -race -count=1` (`tools/pruefung/main.go`, ohne `-timeout`, es gilt die Vorgabe von 10 Minuten je Paket); ein Host, der den Container auf weniger als etwa 0,6 CPU drosselt, bräche die Grenze. Kein Fehler im Code, ein Betriebsrisiko, das der Plan (§3, Prüfgrenze) mit dem Satz „länger als die Zeitgrenze von zehn Minuten“ als Grund der Mengengröße nennt, aber nicht als Folge für `make test-race` bei Last. **Adresse:** Planner, bei der Closure als Risiko-Ausgang in §6 eintragen (weiter beobachten, Zahl 133 s und 352 s als Messpunkt); kein Blocker. |
| N-V2 | INFO | Mutationen N1 und N9 grün | Gleichwertig, nicht ungefangen: N1 (Backslash im Paar schützt nicht vor `;`) ändert nichts, weil `;` im Paar kein Sonderzeichen ist, nur `"` beendet das Paar. N9 (Token endet am Backslash) ändert nichts, weil `istTokenByte` den Backslash nicht enthält (`src/hexagon/model/redaction.go`, Menge `!#$%&'*+-.^_`|~` und alphanumerisch). |
| N-V3 | INFO | Plan §4, Absatz Größenmarke | Der Absatz nennt zuerst die Rechnung mit 26 Zeilen (Stand vor dem Code, so gewollt) und danach den Stand „nach der Fix-Runde 4“ mit 32 Zeilen, *S* = 3065, Marke 3500; die Marke ist nach dem ersten Code-Commit fest. Gelesen, stimmig. |

## (1) Produktcode seit `9a6ee7dc`

`git diff 9a6ee7dc..HEAD -- src ':!*_test.go'` liefert genau eine Datei: `src/hexagon/model/rueckgabe.go`, +2 −1, nur der Kopfkommentar von `LiesRueckgabe` („an 20 Stellen des langen Laufs der Grammatik … je mit einem Baustein; andere Läufe an diesen Stellen bleiben ungeprüft“). Kein Code, keine Signatur, keine Konstante geändert; Verhalten gleich. Die Zusammenführung zu `istTokenByte` (F-13) steht nicht in diesem Diff, sie liegt vor `9a6ee7dc` (beide Aufrufe in `rueckgabe.go` Zeile 41 und 74 nutzen `istTokenByte` aus `redaction.go`).
Der Kommentar sagt nur zu, was ein Test prüft (`AGENTS.md` §3.11): `TestLiesRueckgabeZeit` führt 20 Formen und prüft `n != 20` (selbst gelesen, `zeitFormen`); `SPEC-285` *Ungeprüft* zählt dieselben 20 Stellen (Zeilen, Elemente, Kommas, Leerraum an vier Stellen, Name, Wert als Token, Wort ohne Backslash, mit Backslash, offen, Rest hinter dem Wert, fremder Wert, fremder Rest, fremdes Paar ohne Backslash, mit Backslash, mit Folge von Backslashes, offen, Backslash außerhalb) gleich der Liste im Test.
Test-Datei `src/hexagon/services/mock/prefer_test.go`: nur ein Kommentar umformuliert (Indikativ über die Schleife statt Konjunktiv, `AGENTS.md` §3.7); die Schleife selbst unverändert.

## (2) DoD und V-1 bis V-3

| Punkt | Ergebnis | Beleg |
|---|---|---|
| DoD 1 bis 4 (Antworten, Randformen, Zustand, `make gates`) | erfüllt, `[x]` im Plan | Produktcode gleich; `make gates` nach Auftrag grün auf `9e1d1126`; nicht neu gefahren |
| DoD Review, Doku-Update, Closure-Notiz, Register, Ausgänge, Paarungen | offen, noch nicht fällig | Stand vor der Closure; §6 trägt Ausgänge *offen* (Mengen, Zustand, Fakes, Raster, Body) |
| V-1 (`welle-etag` Zeile 47 und 64) | behoben | `git diff 9a6ee7dc..HEAD -- docs/plan/planning/welle-etag.md`: Kriterium nennt Schlüssel `ETAG_ENABLED` und Fehlercode für `412`, ETag, `If-Match` und `Prefer` als HTTP-Verhalten tragen `SPEC-280` bis `SPEC-289`; *Pflicht Handbuch* gilt für Slices mit Schlüssel oder Fehlercode, `slice-prefer-return` ausgenommen. Plan §6 trägt die Zeile mit Ausgang *entschieden* |
| V-2 (Sendung in §6) | behoben | §6 trägt die Zeile „Zeile `ETAG_ENABLED` …“ mit Ausgang *eingetreten: `slice-handbuch-etag-zeile`*; Nehmer [slice-handbuch-etag-zeile](../plan/planning/open/slice-handbuch-etag-zeile.md) führt die Sendung in §1 (*Übernimmt*, Geber `slice-prefer-return`) und der DoD, Kopf in `welle-benutzerhandbuch` nachgezogen (`AGENTS.md` §3.13) |
| V-3 (Zahl der Zeilen in §4) | behoben | §4 nennt 32 Zeilen; die Tabelle in §3 hat 32 (1 bis 28 mit 6a, 6b, 6c, 9a: 13 + 19) |
| Messpunkt gegen `make groessenmarke` | gleich | Ausgabe: `Messpunkt slice-prefer-return.md: zeilen=32 nebenlaeufig=1 fakes=1 produktcode=300 schaetzung=3065 marke=3500 messung=2399`, „ok, Messung 2399 von Marke 3500“; der Plan nennt dieselbe Zeile, Stand „nach der Fix-Runde 4“ |
| `make docs-check` | grün | `d-check: 878 Datei(en) geprüft, 0 Befund(e)` |

## (3) Tabelle *Zusage · Fänger*

32 Zeilen gezählt. Deckung gegen den Vertrag unverändert zum Vorbericht (`SPEC-285` Nr. 1 bis 4 durch 1 bis 9a; `SPEC-286` durch 10 bis 13; `SPEC-287` durch 14 bis 17, 21, 24, 27; `SPEC-288` durch 18 bis 22; `SPEC-289` durch 23, 28; Adapter 25). Neu seit dem Vorbericht: Zeile 6c (Backslash an jeder Stelle) mit zwei Mengen (`TestLiesRueckgabeErschoepfend`, 1.948.717 und 2.396.745 Zeilen, dazu die Strukturmenge mit 2.396.745; `TestLiesRueckgabeBackslashStellen`, 71 × 6 × 8 × 7 = 23856) und Orakel ohne Code des Lesers. Die Zeile 9a benennt jetzt die 20 Stellen. Zeile 26 „gilt nicht“ nennt Grund (`SPEC-200` Nr. 1: kein Schlüssel, kein Kommando) und Folge (Verhalten tragen `SPEC-285` bis `SPEC-289` und die Tests). Prüfgrenze der Mengen (Plan §3, nach Zeile 6c): Grund (jede Länge mehr multipliziert mit 8 oder 11, Zeitgrenze) und Folge im Betrieb (Formen erst jenseits sechs Symbolen, sieben Zeichen) genannt, jede „gilt nicht“ und jedes Negativ hat Grund und Folge. Zu §3.11: der Kopf von `TestLiesRueckgabeBackslashStellen` sagt „kein Anspruch auf jede Stelle“ und nennt zwei Ungeprüftes, Kopf von `TestLiesRueckgabeErschoepfend` nennt Mengen, Längen und das Ungeprüfte (länger als sechs, sieben, sieben; mehr als eine Zeile; anderer Buchstabe als `a`; ein Element `return` vor dem festen in der Zeichenmenge).

## (4) Offene Punkte aus Review Runde 4

**MEDIUM (Fänger schmaler als die Zusage):** behoben. Die Klasse wird jetzt erschöpfend über das Alphabet der Grammatik gebildet statt aus Beispielen: drei Mengen, je Zeile gleich dem Orakel (`orakelZeilen`, reguläre Ausdrücke, kein Code mit `LiesRueckgabe` geteilt, selbst gelesen). Das Zeichen vor der Folge und der Kontext des Paares sind darin Achsen (jede Folge bis sieben Zeichen). Eigene Mutationen (unten) bestätigen es.
**LOW §3.11:** behoben (Köpfe, siehe (3)). **LOW §3.9 (Zeile 6c fehlte, Plan):** behoben, Zeile 6c steht in der Tabelle, Kennungen `S1`/`S2` statt der doppelten `B1`/`B2`. **LOW Beschriftung des Messpunkts:** behoben („Stand nach der Fix-Runde 4“, Wert 2399 gleich der Messung).

### Eigene Mutationen (`sh tools/harness/mutationslauf.sh`, Klasse `andere`, unter `flock` auf `.wt/heavy.lock`)

Messender: der Verifier; 9 Mutationen in `src/hexagon/model/rueckgabe.go`, Prozess je Stufe (Stufe 1 der benannte Test `TestLiesRueckgabeErschoepfend`, Stufe 2 das Paket), ohne Detektor; Last des Hosts 15,6 bis 23,2 von 4 CPUs. Die Liste liegt nicht im Repo. Die Mutationen stammen aus Positionen, die weder Review (Leerzeichen, `;`, Tab, `=`, Elementanfang, hinter `;`, hinter Paar, zweites Paar, Zeilenende, hinter `,`) noch Implementer (EA2, EA8, S1, S2, C1 bis C5) nannten.

| Nr. | Mutation | Ergebnis |
|---|---|---|
| N1 | Backslash im Paar schützt nicht, wenn `;` folgt | grün, gleichwertig (`;` ist im Paar kein Sonderzeichen) |
| N2 | Backslash im Paar schützt nicht, wenn ein Backslash folgt (Folge `\\`) | rot im benannten Test |
| N3 | Backslash unmittelbar hinter dem Öffnungszeichen schützt nicht (`"\"`) | rot |
| N4 | Ein Paar hinter Tabulator ist kein Paar | rot |
| N5 | Ein Paar unmittelbar hinter `=` ist kein Paar | rot |
| N6 | `\\` im Wort in Anführungszeichen fällt weg (`"min\\imal"` wird `minimal`) | rot |
| N7 | Backslash am Ende des offenen Wortes ergibt ein gültiges Wort (`return="minimal\`) | rot |
| N8 | Ein Backslash hinter dem Wert beendet den Wert (`return="minimal"\`) | rot |
| N9 | Der Token des Wertes endet am Backslash | grün, gleichwertig (`istTokenByte` enthält keinen Backslash) |

Zusammenfassung des Werkzeugs: `mutationen=9 rot=7 rot-paket=0 gruen=2 flatternd=0 uebersetzungsfehler=0 abbruch=0`, Dauer 101,5 s. Sieben Fänger rot im benannten Test in jedem Lauf (ein Prozess je Mutation, deterministisch, kein Zufallsfänger), zwei gleichwertige grün aus Grund. Gegenlesen: die Liste ist ASCII ohne Escape-Folge aus Backslash, `u` und Hexziffern; geschrieben mit `Write` im Scratchpad, nicht im Repo; `git status` im Repo sauber vor dem Bericht.

### Laufzeit von `TestLiesRueckgabeErschoepfend` unter `-race` gegen die Grenze von 10 Minuten je Paket

Messender: der Verifier; Image `odata-recorder:race`, Kopie des Baums per `docker cp`, `CGO_ENABLED=1 go test -race -count=1 -trimpath -buildvcs=false`, netzlos, `--memory=2g`, unter `flock`; Last des Hosts 10 bis 23 (20 Kerne). Je ein Prozess.

| Lauf | CPUs des Containers | Dauer |
|---|---|---|
| `TestLiesRueckgabeErschoepfend` allein, Lauf 1 | 4 | 111,1 s |
| dasselbe, Lauf 2 | 4 | 95,4 s |
| Paket `src/hexagon/model` ganz | 4 | 132,7 s |
| Paket `src/hexagon/model` ganz | 1 | 352,3 s |

Der Plan nennt 93 s für den Test allein (Last 11 bis 12); das stimmt in der Größenordnung mit 95 bis 111 s überein. Bewertung: bei 4 CPUs liegt das Paket bei 22 Prozent der Grenze; bei einer CPU bei 59 Prozent. Der Test ist CPU-gebunden und läuft in 11 Goroutinen, er skaliert mit der freien CPU. Ein Risiko für `make test-race` bei Last entsteht erst, wenn dem Container dauerhaft weniger als gut eine halbe CPU zur Verfügung steht oder der Detektor-Aufwand (etwa 40-fach gegenüber 2,5 s ohne Detektor) zunimmt; die Mengen um ein Symbol zu erweitern (Faktor 8 bis 11) wäre über der Grenze, wie der Plan sagt. Das ist N-V1 (LOW). Die Gate-Stufe `race` (nach Auftrag etwa 335 s für alle Pakete; nicht von mir gemessen) ist nach dieser Messung durch dieses Paket allein nicht gefährdet.

## (5) Hard Rules im Delta

| Regel | Befund |
|---|---|
| §3.7 | Kommentare im Delta in Indikativ über den Zustand (`rueckgabe.go` Kopf, Kopf der Tests, `prefer_test.go`); die Zustandsangabe „Stand nach der Fix-Runde 4“ stimmt mit dem Wert überein |
| §3.9 | Die Fix-Commits `ad8e4f01` und `9e1d1126` tragen je Test und Plan (§3, §4, §6) im selben Commit; Spec (`SPEC-285` *Ungeprüft*) und Kopfkommentar nennen dieselben 20 Stellen wie der Test; `welle-etag` und `welle-benutzerhandbuch` nachgezogen |
| §3.10 | Je Zusage Mutation und roter Test; Zufallsfänger `TestLiesRueckgabeZeit` mit Raten im Plan; die neuen Fänger sind deterministisch (ein Prozess, rot in jedem Lauf, oben) |
| §3.11 | Köpfe nennen nur, was die Prüfung fängt, und was offen bleibt |
| §3.13 | Sendung `ETAG_ENABLED`-Zeile: Nehmer `slice-handbuch-etag-zeile` in `open/` führt sie in §1 und DoD; kein Nehmer in `done/` |
| §3.17 | Menge aus dem Wortlaut (jedes Zeichen des Alphabets der Grammatik bis zur Länge) statt aus dem Befund; Prüfgrenze mit Grund und Folge; eigene Mutationen an Elementen, die der Befund nicht nannte (N2 bis N8), rot |

## (6) Validator: Wert für den Endnutzer

Unverändert vorhanden. Die Fix-Runden 2 bis 4 ändern keine Zeile des Produktcodes außer einem Kommentar; `Prefer: return=minimal` und `return=representation` für POST, PATCH und PUT liefern weiterhin Status, Body, `Location`, `OData-EntityId` und `Preference-Applied` nach `SPEC-286` (Verhalten gleich, Beleg: leerer Produktcode-Diff, `make gates` nach Auftrag grün). Das Handbuch (Weg B) bleibt ohne Kapitel zu `Prefer`; der Anwender-Nutzen am HTTP-Verhalten liegt in Spezifikation und Tests, wie in der ersten Verifikation festgehalten.

## Nicht gefahren

`make gates` (Auftrag: grün auf `9e1d1126`), `make gegenproben`, `make test-integration`, `make test-contract`: der Delta-Diff ändert weder Gates noch Gegenproben-Eingaben noch Produktcode.

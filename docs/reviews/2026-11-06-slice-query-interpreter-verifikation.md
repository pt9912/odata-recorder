# Verifikationsbericht: slice-query-interpreter — 2026-11-06

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan, DoD und Vertrag; nicht „das Richtige" (Validator) und nicht „der Diff" (Reviewer).

**Gegenstand:** `slice-query-interpreter`, Zweig `slice-query-interpreter`, HEAD `5e1ea25`, Diff `055c295..5e1ea25` (Architect `055c295` und `a37fb08`, Review `5f06269`, Fix-Runde `ccc87b9` bis `5e1ea25`). Bezug: `LH-FA-08`, `LH-FA-11`, `LH-QA-02`, `LH-QA-05`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md).

**Geprüft gegen:** Plan `slice-query-interpreter` (§1, §2 DoD, §3 Tabelle *Zusage · Fänger* mit 39 Zeilen, Mengen, Raten, Fix-Runde, §4 Marke und Messung, §6 Raster und Ausgänge), Vertrag `SPEC-190` und `SPEC-191` (jede Nummer), `SPEC-166` Nr. 4, `SPEC-128` Nr. 1 und 5, `AGENTS.md` §3.9, §3.10, §3.13 bis §3.18, der Review-Report `2026-11-06-slice-query-interpreter-review`.

**Methode.** Einmal gefahren, keine Wiederholungsschleife. Der Bericht des Implementers ist nicht übernommen. Gemessen hat der Verifier. Docker nur unter `flock /tmp/odata-recorder-heavy.lock`, ein schwerer Lauf zugleich; Mutationen mit `tools/harness/mutationslauf.sh` (Image `odata-recorder:race`, Container `--memory 2g --cpus=4 --network none`, ohne Mount), Last des Hosts 4 bis 9 bei 20 CPUs (die Ergebniszeilen nennen sie je Mutation, `last=`), Messender: der Verifier. Die Listen liegen im Scratchpad des Laufs (`ver-qi/liste-a.json`, `liste-b.json`, `liste-z.json`, `liste-p.json`), nicht im Repo. Der Produktcode wurde nicht verändert; das Probe-Paket (eine Testdatei) lag nur in einer Wegwerfkopie, die geräumt ist. Rohbytes (§3.18): in den zehn neuen Dateien des Diffs 0 Treffer für NBSP, BOM, Nullbreitenzeichen, Zeilentrenner und die Folge Backslash, `u`, vier Hex-Ziffern (`LC_ALL=C grep -cP`); diese Datei ist ebenso gegengelesen (Ergebnis am Ende).

**Ein Fehlanlauf, offen genannt.** Mein erster `make gates` schrieb in eine Logdatei, die ein anderer Lauf desselben Scratchpads mitbeschrieb (das Log zeigte ein `d-check` auf `/Development/odata-recorder` statt auf den Worktree); ich habe ihn verworfen und mit eigener Logdatei wiederholt. Nur der zweite Lauf zählt (unten).

---

## Ergebnis in einem Satz

**Closure ja, nach Nachzug.** Kein HIGH, kein Befund am Produktcode: jede Nummer von `SPEC-190` und `SPEC-191` ist im Code getragen, `make gates` ist grün, ein Differenzlauf über 6000 zufällige Abfragen gegen eine unabhängige Referenz weicht in keinem Fall ab, und 18 eigene Zähler in nicht genannten Zweigen sind je 20 von 20 gefangen. Die Fix-Runde deckt die Klasse **nicht ganz**: die Klasse „Position in einer Liste" des Befunds M-7 bleibt an vier Orten offen, fünf Mutationen bleiben grün (V-1, MEDIUM). Dazu kommt eine Sendung an `slice-mock-lesen`, die noch nicht beim Nehmer steht (V-2, MEDIUM, Planner).

---

## DoD-Konformität (Plan §2)

| Punkt der DoD | Beleg | Urteil |
|---|---|---|
| Filter und Count: Zählung nach dem Filter und vor dem Paging, `@odata.count` und `/$count` mit derselben Zahl | `FuehreAus` liefert `Anzahl` aus der gefilterten, geordneten Menge vor `seite`; Fänger `TestAnzahl` (60 Fälle); Differenzlauf: `Anzahl` gleich der Referenz in 6000 von 6000 Fällen. Die Gleichheit von `@odata.count` und `/$count` ist eine Pflicht des Aufrufers (V-2) | erfüllt |
| Sortierung und Paging: stabil, Gleichstand nach Entity-Key, `$top=0` leer, `$skip` über das Ende leer | `TestPaging` (147 Fälle), `TestPermutation` (720 Permutationen), `TestGleichstandSchluessel`; Differenzlauf: Reihenfolge, Skip, Top und eine zufällig gemischte Eingabe ergeben dieselbe Folge (6000 von 6000) | erfüllt |
| Projektion mit Property-Maske: `$select` einer nicht beobachteten Property ergibt `ODR-E6003`; `null` und fehlend unterscheidbar | `pruefeName` meldet `nicht_beobachtet` (Klasse `ErrFilterNichtBeobachtet`, `ODR-E6003`); Abbildung über `Abbilden` in `abbildung_test.go` (4 Ursachen, roh, verpackt, verbunden); `TestNullUndFehlend`; Differenzlauf prüft die Properties jeder Entity je Abfrage | erfüllt |
| `make gates` grün | eigener Lauf, `EXIT 0`: `a-check` 0 Befund, `d-check` 674 Dateien 0 Befund, `groessenmarke` ok 3070 von 4500, `zeichen` ok; die Stufen `test`, `race`, `build` und `lint` kamen aus dem Build-Cache (gleiche Eingabe wie ein früherer Lauf, in dem sie liefen); die Pakettests habe ich zusätzlich in den Mutationsläufen selbst gefahren (Nullprobe je Liste grün, auch mit Detektor) | erfüllt |
| Review durchgeführt, Report liegt vor | `docs/reviews/2026-11-06-slice-query-interpreter-review.md` (8 MEDIUM, 1 LOW, 2 INFO) | erfüllt |
| Doku-Update bei öffentlichem Vertrag | `spec/spezifikation.md` trägt `SPEC-190` und `SPEC-191` (Architect, vor dem ersten Code-Commit); kein neuer Fehlercode im Katalog des Binary, Handbuch unberührt | erfüllt |
| Closure-Notiz mit Lerneintrag, Register fortgeschrieben, Ausgang je Risiko, drei Paarungen | §7 steht noch auf „Wird bei der Closure gefüllt"; das Register ist nicht fortgeschrieben; §6 trägt je Risiko einen Ausgang (ein *offen* mit Adresse, ein *weiter offen* mit Trigger) | Aufgabe der Closure (Planner), noch nicht erledigt |

---

## Spec-Konformität: jede Nummer von `SPEC-190` und `SPEC-191` gegen den Code

Gelesen: `abfrage.go`, `abfrage_filter.go`, `abfrage_ordnung.go`, `abfrage_fehler.go` (zusammen 491 Zeilen). Gegen Zusage und Code gehalten, nicht gegen den Bericht des Implementers.

| Nummer | Code | Urteil |
|---|---|---|
| 190 Kopf, Signaturen | `FuehreAus(typ, entities, opt, beobachtet)`, `Ergebnis`, `AbfrageFehler` mit `Code`, `Error`, `Is`, `AlleAbfrageUrsachen` (je Aufruf eine neue Liste) | getragen |
| 190 Nr. 1 a bis d | `Count` und `Sonstige` werden nirgends gelesen; `nil` und leere Liste gleich (`len`); die Maske nur über `pruefeName` und `PruefeFilter` | getragen |
| 190 Nr. 2 | `projiziere` legt `make([]Entity, 0, …)` an (nie `nil`); `Anzahl` aus der gefilterten Menge; `ordne` ordnet eine Kopie, `projiziere` kopiert `Properties` (Mutationen VE1 bis VE3 rot); jeder Fehler gibt `Ergebnis{}` zurück | getragen |
| 190 Nr. 3 a bis f | `pruefeArgumente` und `pruefeEntities` laufen vor allem anderen, alle sechs Fälle als `formFehler` (`ErrUngueltig`, kein Fehlerwert) | getragen |
| 190 Nr. 3 g | Fehler von `filterWert`, `WahrheitVon`, `Vergleiche`, `pruefeWert` und `baueSpalte` kommen unverändert zurück; kein Fall wird zu „falsch" | getragen (Lücke des Fängers bei einer Entity: V-3) |
| 190 Nr. 4 | Tabelle der Ursachen, Klassen, Codes; Folge `eigenschaft_unbekannt`, `typ_nicht_unterstuetzt`, `ordnung_nicht_unterstuetzt`, `nicht_beobachtet` in `pruefeName`; `Error()` nach Muster; kein Name im Text; Ursache außerhalb der vier hat die Klasse `ungueltig` | getragen |
| 190 Nr. 5 | Stufenfolge in `FuehreAus`: `pruefeArgumente`, `PruefeFilter`, `$orderby`, `$select`, `pruefeVorhandene` (Filter vor `$orderby`), Auswertung; jeder Fehler beendet sofort | getragen (110 Paare: Mutationen VJ1 bis VJ3 rot) |
| 190 Nr. 6 | keine Paketvariable, kein Cache, keine geteilte Liste; 18 eigene Zähler 20 von 20 (unten) | getragen |
| 190 Nr. 7 | keine Obergrenze im Code (gelesen); Prüfgrenze 10.000 und 10.001 an acht Stellen (VB1 bis VB4 rot) | getragen |
| 191 Nr. 1 | Stufenfolge; Ordnung entfällt nie | getragen |
| 191 Nr. 2 a | `PruefeFilter(typ, Ausdruck, beobachtet)` vor jeder Auswertung, ohne Rücksicht auf die Eingabe; Fehler unverändert | getragen |
| 191 Nr. 2 b | `filterNamen` sammelt die Namen über alle Knotenarten ohne Kurzschluss, `verlangeProperties` prüft jede Entity | getragen (Lücke des Fängers: `not not`, V-4) |
| 191 Nr. 2 c | beide Operanden immer ausgewertet; Wahrheit als `boolean` und unbekannt als `null`; nur `Durchlaesst()` lässt durch | getragen; der Differenzlauf bestätigt die Null-Semantik von `SPEC-166` Nr. 4 (ein Vergleich mit `null` ist wahr oder falsch, nie unbekannt) |
| 191 Nr. 3 a bis c | Spalten in der Folge von `OrderBy`, `desc` kehrt das Vorzeichen, `null` zuerst bzw. zuletzt; Property in jeder Entity | getragen |
| 191 Nr. 3 d | Key-Properties in der Folge von `typ.Keys`, aufsteigend; `Edm.Guid` nach dem Text der kanonischen Schreibweise (Spalte mit `guid`) | getragen |
| 191 Nr. 3 e | letzte Stufe `strings.Compare(Schluessel)`, bytegleich, kein Programmfehler | getragen (Lücke des Fängers: lange Texte, V-5) |
| 191 Nr. 4 | `seite` rechnet `von = min(skip, n)`, `bis = von + top` nur wenn `top < n - von`: keine Summe der beiden Werte, kein Überlauf bei 2147483647 | getragen |
| 191 Nr. 5 a bis c | `projiziere`: ohne `Select` oder mit `Alle` alle, sonst Namen aus einer Menge in der Folge der Eingabe; `Typ` und `Schluessel` bleiben, `Uebersprungen` nil, `Properties` kopiert | getragen (VAF rot) |
| 191 Nr. 6 | Filter und Ordnung `422`, `$select` lässt aus | getragen |

---

## Plan gegen Code

| Plan | Code | Urteil |
|---|---|---|
| §3 Dateien: `abfrage.go`, `abfrage_filter.go`, `abfrage_ordnung.go`, `abfrage_fehler.go` neu; `abbildung_test.go` update; `abfrage_url_test.go` neu; kein neues Paket | der Diff nennt genau diese Dateien und ihre `_test.go`-Paare; `abbildung.go` unberührt; `spec/spezifikation.md` mit drei Zeilen (Nr. 3 e) | gleich |
| §3 Tabelle: 39 Zeilen, 2 `nebenläufig` (32, 39) | 39 Zeilen; die Fänger sind benannt und vorhanden (`func Test…` zu allen Namen der Spalte, gegengelesen) | gleich |
| §4 Gate-Kette: Slice berührt keine Datei der Kette, `.a-check.yml` unverändert | `git diff 055c295..5e1ea25` über `Makefile`, `harness`, `.a-check.yml`, `Dockerfile`, `tools`, `.golangci.yml`: leer; `//nolint` in `src` und `tests`: 0 Treffer; `make a-check` 0 Befund | gleich |
| §4 Produktcode 491 gegen 550 geschätzt | 491 (gemessen, vier Dateien) | gleich |
| §3 Mengen, gemessen: „TestNamen 2 × 39 × 3 = 234" | `TestNamen` zählt 43 Arten und `2 × 43 × 3 = 258` (der Test verlangt es selbst); die Fix-Runde nennt 43 und 258 | Abweichung im Plan (V-6) |
| §4 Hinweis: „nach `SPEC-128` Nr. 1 ergäbe 3680 die Marke 4600" | 100 · ⌈135 · 3680 / 10000⌉ = 5000 | Abweichung im Plan (V-6) |

---

## Größenmarke: `SPEC-128` Nr. 1 und 5 d

`make groessenmarke` (im `make gates`-Lauf): `slice-query-interpreter.md: ok, Messung 3070 von Marke 4500`, dazu der Hinweis `Tabelle ergibt 3680, Schätzung 3275`.

- **Kein `ueberschreitung`-Eintrag.** Nr. 5 d: eine Überschreitung ist die Messung *über* der Marke. 3070 liegt unter 4500; ein Eintrag wäre nach Nr. 5 f ein Hinweis „Eintrag ohne Überschreitung". Erst eine Messung über 4500 verlangt `ueberschreitung: <aufgerundet> | <Grund>`. Spielraum: 1430 Zeilen.
- **Keine Neumarke.** Nr. 5 b prüft `marke` gegen die Formel angewandt auf `schaetzung` (3275, Marke 4500 stimmt); Nr. 5 c macht den Unterschied zur Tabelle nach dem ersten Code-Commit zum Hinweis, nicht zum Befund, und `AGENTS.md` §3.14 hält die Marke nach dem ersten Code-Commit fest. Eine Neumarke aus 3680 wäre eine andere Formel-Eingabe (`schaetzung` ändern) und ist nicht verlangt.
- **Der Satz im Plan ist falsch gerechnet:** 3680 ergäbe 5000, nicht 4600 (V-6). Er ändert das Ergebnis nicht, soll aber vor der Closure berichtigt werden, denn ein Leser nähme 4600 als Formelwert.

---

## Die Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Der Review nannte 8 MEDIUM (M-1 bis M-8). Ich habe nicht die genannten Mutationen wiederholt, sondern **42 eigene Mutationen an Elementen gesetzt, die weder der Review noch die Fix-Runde nannte** (Listen A und B, Klasse `andere`, ein Prozess je Mutation), dazu 18 Zähler (Klasse `nebenläufig`). Ergebnis der 42: **29 rot, 13 grün**. Die 13 grünen sind unten eingeordnet.

### Rot gesehen (Elemente, die der Befund nicht nannte)

| Klasse (Befund) | Mutation am nicht genannten Element | Ergebnis |
|---|---|---|
| M-1 Programmfehler der Ordnung | Klassenprüfung nur nicht im Schlüssel (VA1); Klasse `Zahl` ausgenommen (VA5); Fehler der zweiten und folgenden Key-Spalte übergangen (VA6) | rot (`TestOrdnungProgrammfehler`) |
| M-2 Prüfgrenze an jeder Stelle | Zählung im Filter bei 10.000 gekappt (VB1); Index der Ordnung gekappt (VB2); `bis` in `seite` gekappt (VB3); Stufe 5 nur für `$orderby` bei 10.000 gekappt (VB4) | rot (`TestGrosseMengeStellen`) |
| M-4 und M-5 Schlüssel | Null der zweiten Key-Property in Stufe 1 übergangen (VC6) | rot (`TestStufe1Schluessel`) |
| M-6 Filter und `$orderby` | Element 11 von `OrderBy` in `ordne` übersprungen (VC4); Property ab dem 33. Eintrag der Property-Liste nicht gefunden (VC9) | rot (`TestOrderByLaenge`) |
| M-7 Listen | fünfter Name von `$select` nicht in die Menge der Projektion (VC7); letzte Navigation nicht erkannt (VC8) | rot (`TestProjektionNamen`, `TestNamen`) |
| M-8 Ergebnis gehört dem Aufrufer | Alias der fünften Entity ohne `$select` (VE1); Alias der zweiten Entity mit `$select` (VE2); `Typ` der Entity fehlt (VE3); `Uebersprungen` leer statt nil (VAF) | rot (`TestErgebnisUnabhaengig`) |
| Paging, Rangfolge, Fehler | `top < n - von - 1` (VD1); Reihenfolge der Namen von `$select` und `$orderby` umgekehrt (VJ1, VJ2); Ordnung vor Maske vertauscht (VJ3); `Select{Alle}` mit doppeltem Namen angenommen (VI1); `Skip` -1 angenommen (VI2); Richtung zählt bei doppeltem `$orderby`-Namen (VG2); Klasse der Ordnungs-Ursache (VK1); Text der Klasse `501` (VK2); `Skip` auf `n - 1` begrenzt (VL1, rot, wahrscheinlich über einen Abbruch bei `n` gleich 0, nicht geprüft); doppelter Schlüssel nicht erkannt, wenn schon fünf gemerkt sind (VP1) | rot |
| Filter, Baum | Entity 2 von `verlangeProperties` übersprungen (VF1) | rot (`TestFilterEigenschaftFehlt`) |
| offener Typ | bekannte Property im offenen Typ als `typ_nicht_unterstuetzt` (VU1) | `rot-paket`: `TestNamen` bleibt grün, ein anderer Test des Pakets wird rot (V-7, INFO) |

### Grün geblieben (13), eingeordnet

| Mutation | Was sie ändert | Einordnung |
|---|---|---|
| VC2 | die Namen von `$select` ab dem siebten werden nicht geprüft | Lücke: V-1 |
| VC3 | die Namen von `$orderby` ab dem 17. werden nicht geprüft | Lücke: V-1 |
| VC5 | Stufe 5 prüft die 21. und folgende `$orderby`-Property nicht gegen die Entities | Lücke: V-1 |
| VC1 | ein doppelter Name wird nur erkannt, wenn sein zweites Vorkommen an Stelle 1 bis 4 steht | Lücke: V-1 |
| VN1 | ein doppelter Name wird nicht erkannt, wenn sein zweites Vorkommen an Stelle 4 steht | Lücke: V-1 |
| VA7, VA8, VA4 | bei **einer** Entity bleibt ein Wert ohne Leser (`redigiert`, unbekannte Art) oder eine Guid unter Zahlen in der Ordnung ohne Fehler | Lücke: V-3 (Produktcode ist richtig, nur der Fänger führt keine Eingabe mit einer Entity) |
| VA2, VA3 | der Fehler von `pruefeWert` wird bei der letzten bzw. dritten Entity übergangen | **äquivalent** bei mehr als einer Entity: der Rest ergibt `Klassen nicht vergleichbar` in derselben Klasse; gefangen erst bei einer Entity (gleiche Lücke wie VA7) |
| VF2 | `not not P`: die Property unter dem doppelten `not` wird nicht gesammelt | Lücke: V-4 |
| VF3 | die sechste Entity wird in Stufe 5 nicht geprüft | **akzeptiertes Negativ vorzuschlagen** (I-1): unbeschränkte Positionsmenge, Prüfgrenze 5 und der letzte von 10.000/10.001 |
| VH1 | der Rest der Ordnung vergleicht nur die ersten 8 Bytes von `Schluessel` | Lücke: V-5 |

### Zeile 32 und die Zähler (`AGENTS.md` §3.10)

Ich habe **18 Zähler** gesetzt: je eine Zuweisung `ErrKodierung = ErrKompatibilitaet` (ein echtes Schreiben auf eine Paketvariable, nicht `x = x`, das der Compiler streicht), jeweils am Anfang eines Zweigs, den weder der Review (48 Zähler) noch der Plan als Zähler nannte oder nur mit anderer Zeile: die Fehler `Filter ohne Ausdruck`, `doppelter Name in $orderby`, `Select ohne Namen`, `Key-Property fehlt`, das Merken der Schlüssel, `verlangeProperties` im Verstoß, `bis` in `seite`, `von` in `seite`, die Menge der Projektion, `wertVon` im Fehlschlag, `FilterNull`, das Ende von `filterNamen`, `treffer` im Filter, `false` in `wahrheitAlsWert`, die Guid-Spalte, der Guid-Vergleich, die absteigende Richtung, `Code()` für die Klasse `501`. Klasse `nebenläufig`, `TestAbfrageGleichzeitig`, Detektor (`-race`, `-test.count=1`), **20 Prozesse je Zähler**, ohne Arm ohne Detektor (das Fehlen der Sperre fängt der Detektor, `AGENTS.md` §3.10): **18 von 18 Zählern je 20 von 20 rot, kein grüner, kein flatternder**; unmutierter Code: Nullprobe grün. Last des Hosts 4,0 bis 4,8 bei 20 CPUs, Container 4 CPUs und 2g, Messender: der Verifier. Damit bestätigt sich das Ergebnis des Reviews unabhängig (48 Zähler) und erweitert sich um 18 weitere Zweige.

### Differenzlauf (Probe-Paket aus Generatoren, `AGENTS.md` §3.17)

Ein Test in der Wegwerfkopie erzeugte mit festem Samen 6000 Fälle: 0 bis 13 Entities mit unabhängigen Werten für `Name` (null, leer, `a`, `B`, `b`, `ab`, ein Zeichen jenseits von ASCII), `Big` (`Edm.Int64`, auch 2^53 und 2^53 + 1) und `Id`, zufällige Filterbäume der Tiefe 3 (alle sechs Operatoren, beide Seiten, `eq`/`ne null`, `and`, `or`, `not`), 0 bis 3 Sortierschlüssel in zufälliger Richtung, `Skip` und `Top` von 0 bis n + 2, zufälliges `Select`, gelegentlich eine Entity ohne Property (erwartet `422` mit der richtigen Option). Die Referenz (Kleene-Logik, Byte-Ordnung, `big.Int`, stabiler Gleichstand nach `Id`) ist unabhängig vom Code geschrieben; dieselbe Eingabe in zufälliger Reihenfolge ergibt dasselbe Ergebnis. **0 Abweichungen in 6000 Fällen.** Meine erste Fassung der Referenz irrte sich in der Null-Semantik (sie nahm „unbekannt" für einen Ordnungsvergleich mit `null`); der Fehler lag bei mir, `SPEC-166` Nr. 4 und der Code stimmen überein, und die Referenz ist berichtigt. Kontrollen (drei Mutationen gegen das Probe-Paket): `desc` nicht gekehrt und `Top` um eins zu kurz sind rot, `unbekannt` lässt durch ist `rot-paket` (die Referenz kennt dort kein unbekannt, die Lücke ist die der Referenz, nicht des Codes).

---

## Sendungen an `slice-mock-lesen`

Der Plan weist drei Pflichten des Aufrufers zu: die Maske aus den bekannten Properties der Baseline (`SPEC-190` Nr. 1 d), `/$count` nur mit `Filter` (Nr. 2), Entities mit den Key-Properties (Nr. 3 f). Ich habe die Datei `slice-mock-lesen` (heute unter `open/` auf `main`) gegengelesen:

- Ihr §1 *Ausdrücklich NICHT* schließt die Sendung nicht aus (Mutationen, Fallback, Spans, Service-Dokument); ihre DoD und ihr Ziel nennen sie **nicht**. Der Nachtrag steht auf `main` aus (so führt es auch §6 des Plans). Nach `AGENTS.md` §3.13 gehört er in den Commit, der die Sendung nennt; der Slice-Zweig darf den Plan des Nehmers nicht schreiben, also ist es ein Nachzug beim Integrieren (V-2).
- **Die Sendung ist unvollständig.** `SPEC-190` Nr. 3 g macht einen Wert der Art `redigiert` in einer Property, die `$filter` oder `$orderby` nennt, zum Programmfehler (`ODR-E9001`, 500). Liefert der Workspace redigierte Werte (der Marker), die ein Client filtern kann, muss der Aufrufer sie vor dem Aufruf behandeln. Dasselbe gilt für `entities` des Typs des Sets (Typ abgeleiteter Entities: *Ungeprüft*) und `Schluessel` eindeutig und nichtleer (Nr. 3 e). Beides fehlt in der Sendung.

---

## Gate-Kette, `.a-check.yml`, `//nolint`, Rohbytes

- Gate-Kette: `Makefile`, `harness/`, `Dockerfile`, `tools/`, `.a-check.yml`, `.golangci.yml`, `harness/README.md` stehen nicht im Diff (`git diff --stat 055c295..5e1ea25 -- <Pfade>` leer). `GATE_CHECKS` unverändert.
- `.a-check.yml`: unverändert; `make a-check`: `gesamt: 0 Befund(e)`.
- `//nolint`: `grep -rn nolint src tests`: 0 Treffer.
- Rohbytes: 0 Treffer in den zehn neuen Dateien (siehe Methode).

---

## Befunde nach Schwere

**HIGH: keine.**

### V-1 (MEDIUM): Die Fix-Klasse M-7 „Position in einer Liste" ist an vier Orten offen

- `quelle`: `AGENTS.md` §3.17 (der Fix deckt die Klasse, die Menge aus dem Wortlaut der Zusage), §3.15 (ein Fall je Element einer Menge); `SPEC-190` Nr. 3 c, d (jeder doppelte Name), Nr. 4 (*„Die Ursachen gelten für die Namen in `$orderby` und `$select`"*), Nr. 5 Stufen 3 bis 5, Nr. 7 (keine Obergrenze).
- `befund`: Der Review nannte als Klasse „Position in einer Liste nicht über die Länge variiert". Die Fix-Runde variiert die Länge für die **Ordnung** (Zeile 17: 71 Läufe) und die **Projektion** (Zeile 27: 27 Fälle), nicht für die anderen Orte, an denen der Code eine Liste von Namen durchläuft. `TestNamen` führt 3 Positionen (2 × 43 × 3); die Doppelten (`TestAbfrageProgrammfehler`) werden an erster und letzter Stelle von drei geprüft. Fünf Mutationen an Elementen, die der Befund nicht nannte, bleiben grün: `VC2` (`$select`-Namen ab dem 7. ungeprüft), `VC3` (`$orderby`-Namen ab dem 17. ungeprüft), `VC5` (Stufe 5 für die 21. `$orderby`-Property ungeprüft), `VC1` und `VN1` (Doppelte erst ab Stelle 5 bzw. nicht an Stelle 4).
- `klasse`: Menge „Orte, an denen der Code eine Namenliste durchläuft" (Doppelte in Stufe 1, Namen von `$orderby`, Namen von `$select`, Stufe 5, dazu Ordnung und Projektion) × Länge × Stelle des Verstoßes; die Fix-Runde führte nur die Ordnung und die Projektion.
- `verifizierbar`: ja (VC1, VC2, VC3, VC5, VN1).
- Auswirkung: der Code ist richtig (gelesen: keine Schranke); der Fänger würde eine eingebaute Schranke nicht sehen.
- Adresse: Implementer. Fänger als Schleife über Länge (1, 2, 3, 4, 7, 17, 21, 32, 33) × Ort × Stelle des Verstoßes (erste, letzte, jede) mit Zählprüfung, wie `TestOrderByLaenge`.

### V-2 (MEDIUM): Die Sendung an `slice-mock-lesen` steht nicht beim Nehmer und ist unvollständig

- `quelle`: `AGENTS.md` §3.13 (der Geber trägt die Sendung im selben Commit in §1 oder DoD des Nehmers ein); Plan §6, Punkt *Pflichten des Aufrufers* (Ausgang *offen*, Nachtrag „steht auf `main` aus").
- `befund`: siehe Abschnitt *Sendungen*. Die Datei des Nehmers nennt weder Maske, `/$count` noch Key-Properties; zusätzlich fehlen die Pflichten aus `SPEC-190` Nr. 3 e und 3 g (Eindeutigkeit des `Schluessel`, kein Wert `redigiert` in einer gefilterten oder geordneten Property).
- `klasse`: Rückgabe mit Adresse, Nachtrag ausstehend.
- Adresse: Planner, beim Integrieren auf `main`: ein DoD-Punkt in `slice-mock-lesen` mit allen fünf Pflichten und der Kennung des Gebers (`slice-query-interpreter`), im selben Commit, der den Slice nach `done/` bringt oder davor. Bis dahin steht das Risiko *offen* (zulässig), und die Closure darf es nicht als *eingetreten* oder *entfallen* schließen.

### V-3 (LOW): Programmfehler der Ordnung bei **einer** Entity ohne Fänger

- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-190` Nr. 3 g.
- `befund`: Die Menge von Zeile 37 (5 Formen × 4 Spalten × 5 Positionen) hat nie weniger als zwei Entities. Bei einer Entity bleibt ein Wert ohne Leser (VA7) und eine Spalte aus einer Guid unter einer Zahlenklasse (VA4, VA8) ohne Fehler. Der Code gibt den Fehler zurück (gelesen), nur die Mutation „Fehler bei genau einer Entity übergangen" bleibt grün.
- `klasse`: Achse *Zahl der Entities* (0, 1, mehr) nicht in der Menge von Zeile 37.
- `verifizierbar`: ja (VA7, VA8, VA4). Äquivalent bleiben VA2 und VA3 für mehr als eine Entity.
- Adresse: Implementer (eine Achse mehr in `TestOrdnungProgrammfehler`).

### V-4 (LOW): `not not P` in Stufe 5 ohne Fänger

- `quelle`: `AGENTS.md` §3.15; `SPEC-191` Nr. 2 b (*„Jede `FilterEigenschaft` des Baums"*).
- `befund`: Die Stellen von Zeile 14 sind Wurzel, Operand von `not`, Seiten von Vergleich, `and`, `or`; ein `not` unter einem `not` ist keine eigene Stelle. `VF2` (die Property unter dem doppelten `not` wird nicht gesammelt) bleibt grün. Der Code ist richtig (`filterNamen` rekursiv).
- `klasse`: Knotenart unter derselben Knotenart (`not` in `not`, `and` in `and`, `or` in `or`) nicht in der Menge der Stellen.
- Adresse: Implementer.

### V-5 (LOW): Rest der Ordnung: lange Texte

- `quelle`: `AGENTS.md` §3.15; `SPEC-191` Nr. 3 e (*„bytegleich aufsteigend"*).
- `befund`: Die 13 Texte von Zeile 35 sind kurz. `VH1` (nur die ersten 8 Bytes verglichen) bleibt grün; die benannten Mutationen „zwei Bytes" und „Länge zuerst" sind rot. Der Fall tritt im Betrieb nicht auf (kein Leser liefert `Keys` leer; *Ungeprüft*).
- `klasse`: Länge des Texts als Achse, nur bis 2 Bytes fangend.
- Adresse: Implementer (ein Textpaar, das sich erst nach dem 16. Byte unterscheidet).

### V-6 (LOW): Zwei Zahlen im Plan stimmen nicht

- `quelle`: `AGENTS.md` §3.9 (der Plan folgt jeder Korrektur), `SPEC-128` Nr. 1.
- `befund`: (a) §4: „nach `SPEC-128` Nr. 1 ergäbe 3680 die Marke 4600"; die Formel gibt 5000. (b) §3 *Mengen, gemessen*, Zeile *Namen der Optionen*: „`TestNamen` 2 × 39 × 3 = 234"; der Test zählt 43 Arten (258) und verlangt es selbst, die Fix-Runde nennt 258.
- `klasse`: Stand des Plans hinter der Korrektur.
- Adresse: Implementer (oder Planner bei der Closure; beides sind Sätze im Plan).

### V-7 (INFO): `VU1` wird vom benannten Test nicht gefangen

Eine bekannte Property im offenen Typ als „nicht unterstützt" zu melden (`VU1`) macht `TestNamen` nicht rot, aber ein anderer Test des Pakets (`rot-paket`). Die Zusage hat also einen Fänger, nur nicht den, den die Tabelle in Zeile 21 nennt. Kein Befund am Produkt; beim Nachzug von V-1 kann der Fall in `TestNamen` aufgenommen werden (ein bekannter Name im offenen Typ).

### I-1 (INFO): Positionen jenseits der geführten Menge

`VF3` (die sechste Entity in Stufe 5 nicht geprüft) ist eine Position, die keine Menge des Plans führt (5 Positionen und der letzte Eintrag von 10.000 und 10.001). Ich schlage vor, sie als akzeptiertes Negativ zu führen: *Grund:* die Positionen sind unbeschränkt, und der Code hat keine positionsabhängige Logik (gelesen); *Folge:* eine eingebaute Schranke bei einem Index zwischen 6 und 9.999 bliebe bei grünen Gates unbemerkt.

### I-2 (INFO): Größenmarke

Siehe Abschnitt *Größenmarke*: weder `ueberschreitung` noch Neumarke, der Satz „4600" ist zu berichtigen (V-6).

---

## Nicht aufgelöst (ehrlich)

- **Zeilen 1 bis 31 des Plans mit den Mutationen des Implementers nicht nachgefahren.** Ich habe eigene Elemente gemutet (42 plus 18) und einen Differenzlauf gefahren, nicht die Listen `m1.json` und `m2.json` des Implementers; die Aussage „alle benannten Mutationen rot" aus §3 ist damit nicht von mir bestätigt, nur nicht widerlegt. Der Review hat 48 Zähler und eigene Mutationen unabhängig gefahren (sein Bericht).
- **Die Rate ohne Detektor der Zeile 32 (Wirkung)** habe ich nicht neu gemessen; die Zähler laufen nur mit Detektor (richtig für ein Datenrennen, `AGENTS.md` §3.10). Die Wirkungs-Mutationen `S1` bis `S4` des Plans (Rate ohne Detektor) sind nicht von mir gemessen.
- **Die drei Zweige, die kein Aufruf erreicht** (`Property der Entity fehlt`, `Knoten des Filters`, `Operator der Verknüpfung`): benanntes Negativ im Plan, von mir nicht gezählt.
- **`make gates`: die Stufen `test`, `race`, `build` und `lint` kamen aus dem Build-Cache** (gleiche Eingabe wie ein früherer Lauf in diesem Zweig). Die Pakettests des Modells habe ich in den Nullproben aller vier Listen selbst gefahren (grün, bei den Zählern mit Detektor); die Pakete außerhalb des Modells nicht einzeln.
- **Mein Differenzlauf kennt keinen Wert „unbekannt"**: `SPEC-166` Nr. 4 lässt Vergleiche nie unbekannt werden, ein unbekannter Wahrheitswert entsteht nur aus einer `Edm.Boolean`-Property mit `null` (Fänger `TestFilterDurchlass`, `TestWahrheitAlsOperand`); den habe ich nicht gemutet.
- **Round-trip über den Schreiber:** der Zweig enthält den Schreiber nicht (er liegt auf `main` in `slice-odata-json-schreiben`); den Weg Parser, Interpreter bis zum Teilbestand deckt `TestAbfrageAusURL` (12 Texte mit Ergebnis, 2 im Parser), den Weg bis in die Serialisierung trägt `slice-mock-lesen`.
- **Die Last des Hosts** schwankte (Tabellen oben); für die Zähler ist das ohne Bedeutung (Detektor, 20 von 20).

---

## Empfehlung: Closure **ja**, nach Nachzug

Der Produktcode trägt `SPEC-190` und `SPEC-191` vollständig; die DoD-Punkte der Lieferung sind erfüllt. Die Closure darf nach den Nachzügen folgen.

**Nachzug Implementer (vor der Closure, im Zweig, ein Commit mit Plan nach §3.9):**
1. V-1: Fänger als Schleife über die vier offenen Orte der Namenliste (Doppelte, `$orderby`-Namen, `$select`-Namen, Stufe 5) × Länge × Stelle; Menge und eine Mutation an einem nicht genannten Element im Plan (nach `AGENTS.md` §3.17).
2. V-3, V-4, V-5: je ein Fall (eine Entity; `not not`; langer Text), wenn der Aufwand klein ist; sonst als akzeptiertes Negativ mit Grund und Folge.
3. V-6: die zwei Zahlen im Plan (Marke 5000 statt 4600; 258 statt 234).

**Nachzug Planner (auf `main`, beim Integrieren):**
1. V-2: DoD-Punkt in `slice-mock-lesen` mit den fünf Pflichten des Aufrufers (Maske, `/$count`, Key-Properties, eindeutiger `Schluessel`, kein Wert `redigiert` in gefilterten oder geordneten Properties) und der Kennung `slice-query-interpreter`.
2. Closure-Notiz mit Lerneintrag; die Klassen dieses Laufs: *Fix deckt Ordnung und Projektion, nicht die vier anderen Orte derselben Menge* (zweiter Beleg zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`), *Achse Zahl der Entities fehlt*; Register fortschreiben.
3. Größenmarke: keine Aktion (3070 von 4500, Messung neu nach dem Nachzug des Implementers; sie liegt weit unter der Marke).

*Gegengelesen auf Rohbytes:* diese Datei, `LC_ALL=C grep -cP` auf NBSP, BOM, Nullbreitenzeichen, Zeilentrenner und Backslash-`u`-Hex.

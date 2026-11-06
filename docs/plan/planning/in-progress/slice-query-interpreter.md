# Slice slice-query-interpreter: Query-Interpreter mit Filter, Sortierung, Paging und Projektion

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-08`, `LH-FA-11`, `LH-QA-02`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-08.a` (Pipeline) · `LH-FA-08.b` · `LH-QA-02.a` (stabile Sortierung) · `SPEC-190`, `SPEC-191` (Vertrag der Ausführung) · `SPEC-044` (Property-Maske) · `SPEC-160`, `SPEC-162` (Optionen, `$top`, `$skip`) · `SPEC-166`, `SPEC-167` (Werte, Typprüfung) · `SPEC-109` (Abbildung) · `SPEC-022`, `SPEC-024` · `ARC-001`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-query-interpreter, Worktree /Development/odata-recorder-wt-query-interpreter)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Interpreter führt einen geprüften Query-AST auf einer Menge typisierter Entities aus: Filter, stabile Sortierung (bei Gleichstand nach Entity-Key, ohne `$orderby` nach Key), Zählung vor dem Paging, `$skip` und `$top` (`$top=0` ergibt eine leere Liste) und Projektion mit Property-Maske, und liefert Elemente und Count für die Serialisierung. Der Vertrag steht in `SPEC-190` (Eingabe, Ergebnis, Fehler, Rangfolge) und `SPEC-191` (Stufen); die Funktion ist `model.FuehreAus` im Paket `src/hexagon/model/`, ohne neues Paket.

**Übernimmt (Hinweis, keine Pflicht; Sendung von `slice-bootstrap-m2`, `AGENTS.md` §3.13):** Längen-Differenz zwischen Filter und Schlüssel (`SPEC-161`, `SPEC-164`; 1641 Abweichungen in 167634 Läufen): alle sind `literal` im Filter gegen `zu_lang` im Schlüssel, erklärt durch die Regel „der Rest hinter dem schließenden `'` zählt nicht zur Länge“ (`SPEC-164` Nr. 4 a); entschieden und vertraglich getragen, kein Fehler. **Beschluss (Architect): keiner, die Regel bleibt.** Der Interpreter liest weder die Länge eines Literals noch die eines Schlüssels (er arbeitet auf Werten des Modells, `SPEC-191`), sein Lauf verwendet den Pfad des Schlüssels nicht; eine Angleichung hätte hier keinen Prüfstand und wäre eine Änderung des Vertrags des Lesers.

**Übernimmt (Sendung von `slice-edm-ausdruecke`, `AGENTS.md` §3.13):** (a) der Interpreter ruft `model.PruefeFilter` mit der Property-Maske vor der Auswertung des Filters und bildet den Fehler über die Abbildung ab (`SPEC-167` Nr. 6) — im Vertrag: `SPEC-191` Nr. 2 a, Fänger Zeile 8; (b) `$orderby` nach einer Property, für die `model.HatOrdnung` falsch ist (`Edm.Guid`), ist `501` mit `ODR-E6001` (`SPEC-166` Nr. 3 f) — `SPEC-190` Nr. 4, Fänger Zeile 20; (c) `desc` kehrt die Ordnung von `model.OrdneWerte` um, `null` steht dann zuletzt (`SPEC-166` Nr. 6) — `SPEC-191` Nr. 3 b, Fänger Zeile 15. Die Auswertung der Wahrheitswerte, die Vergleiche und die Ordnung selbst liefert `slice-edm-ausdruecke`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Coverage-Prüfung vor der Ausführung — `slice-mock-lesen`; der Interpreter bekommt einen Bestand und fragt nicht, wie vollständig er ist.
- Lesen aus dem Repository und Serialisierung — `slice-mock-lesen` und `slice-odata-json-schreiben`; der Interpreter hat keinen Port.
- `$expand` und Navigationen — `welle-expand-batch`.
- Die Entscheidung über negative und nicht ganzzahlige Werte von `$top` und `$skip` — sie fällt dem Parser zu und ist geliefert (`SPEC-162` Nr. 4: `wert`, `400`, `ODR-E6005`; `slice-odata-url-pfad` in `done/`); der Interpreter sieht nur nichtnegative `int` und behandelt einen negativen als Programmfehler (`SPEC-190` Nr. 3 a). Keine Änderung am Parser, keine Sendung.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Filter und Count: die Zählung nach dem Filter und vor dem Paging, `@odata.count` und `/$count` mit derselben Zahl.
- [ ] Sortierung und Paging: stabil, Gleichstand nach Entity-Key, `$top=0` leer, `$skip` über das Ende leer (`LH-FA-08`, Boundary).
- [ ] Projektion mit Property-Maske: `$select` einer nicht beobachteten Property ergibt `ODR-E6003`; `null` und fehlend bleiben unterscheidbar (`LH-FA-18`, Teilprojektion).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/model/abfrage.go` | neu | `FuehreAus`, `Ergebnis`: Prüfung der Argumente (Programmfehler), Rangfolge der Stufen, Zählung, `$skip` und `$top`, Projektion, Kopie (`SPEC-190`, `SPEC-191` Nr. 1, 4, 5) |
| `src/hexagon/model/abfrage_filter.go` | neu | Aufruf von `PruefeFilter`, statische Prüfung der Properties je Entity, Auswertung des Baums je Entity (`SPEC-191` Nr. 2) |
| `src/hexagon/model/abfrage_ordnung.go` | neu | Namen von `$orderby` und `$select` gegen den Typ, Ordnung über die Schlüssel von `OrderBy`, Gleichstand nach dem Schlüssel (`SPEC-190` Nr. 4, `SPEC-191` Nr. 3) |
| `src/hexagon/model/abfrage_fehler.go` | neu | `AbfrageFehler`, `AbfrageOption`, `AbfrageUrsache`, `AlleAbfrageUrsachen` (`SPEC-190` Nr. 4); nutzt die vorhandenen Sentinel, keine neue Zeile der Abbildung |
| `src/hexagon/ports/outbound/abbildung_test.go` | update | ein Fänger: der `AbfrageFehler` aller vier Ursachen ergibt `ODR-E6005`, `ODR-E6001`, `ODR-E6003` über `Abbilden`, auch verpackt (Zeile 24); `abbildung.go` bleibt unberührt |
| `tests/integration/abfrage_url_test.go` | neu | Query-Text → `urlparser` mit `NewFilterLeser` → `FuehreAus`: der Happy Path von `LH-FA-08` und die Grenzfälle (`tests/` darf Modell und Driven Adapter nennen, [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)) |
| `spec/spezifikation.md` | update | Vertrag mit Randformen: `SPEC-190`, `SPEC-191`, Zeiger in `LH-FA-08.a` (Architect) |
| Tests | neu | je Datei der Produktcode `…_test.go` (Paare im Block von §4); Fänger der Tabelle unten |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht hier, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16); beide entstehen mit dem Vertrag in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12), die Randformen entscheidet der Architect, nicht der Implementer. Die Namen der Tests in der Spalte *Fänger* sind Vorschläge; eine Zeile bleibt gleich, wenn der Test anders heißt. Eine Mutation ist eine einzelne Änderung am Code, die der Implementer vor der Übergabe selbst rot sieht (`AGENTS.md` §3.10); eine Schleife läuft über alle Elemente der Menge, die der Wortlaut der Zusage meint (`AGENTS.md` §3.17), und vergleicht gegen eine Erwartung, die der Test selbst trägt (eine Tabelle, eine unabhängig geschriebene Referenz), nicht gegen den Code. Eine Fixture gleicht nie dem Wert der Mutation. Der Test hält am Argument `beobachtet` (eine Funktion, die jeden Namen festhält und jeden Aufruf zählt) und an der Eingabe (vorher und nachher gleich, `AGENTS.md` §3.15).

- Reine Funktionen auf einer Liste im Speicher; wie der Interpreter bei 10.000 Entities pro Workspace bleibt (`LH-QA-01.a`), ist eine Randform des Vertrags: Prüfgrenze 10.000 und 10.001 Entities (Zeile 33), Latenz nicht zugesagt (`SPEC-190` Nr. 7 und *Ungeprüft*).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Die Stufen laufen in der Folge Filter, Zählung, Ordnung, `$skip`, `$top`, Projektion (`SPEC-191` Nr. 1; `LH-FA-08` Happy Path) | `TestStufen`: 8 Entities, eine Abfrage mit Filter, `$orderby` (zwei Schlüssel, einer absteigend), `$skip=1`, `$top=2`, `$select` gegen die ausgeschriebene Erwartung; dazu je eine Abfrage, in der die Folge zweier Nachbarstufen die Menge ändert (Filter gegen Paging, Ordnung gegen Paging, Paging gegen Projektion) | `$top` vor `$skip`; `$skip` auf der ungefilterten Menge; Ordnung nach dem Paging; Filter nach dem Paging | andere |
| 2 | `Anzahl` ist die Zahl nach dem Filter und vor dem Paging, immer, gleich für `@odata.count` und `/$count` (`SPEC-190` Nr. 2; `LH-FA-08` Boundary) | `TestAnzahl`: Schleife über `Count` (nicht gesetzt, falsch, wahr) × `Top` (nicht gesetzt, 0, 1, n, n+1) × `Skip` (nicht gesetzt, 0, 1, n+1) = 60 Fälle gegen die Zahl der Entities, die der Filter durchlässt (4 von 8); dazu derselbe Aufruf mit nur dem Filter (der Fall `/$count`) liefert dieselbe Zahl | `Anzahl` nach dem Paging; `Anzahl` nur bei `Count` wahr; `Anzahl` ohne Filter | andere |
| 3 | `Entities` ist nie `nil` (`SPEC-190` Nr. 2) | `TestEntitiesNieNil`: leere Eingabe, `Top` 0, `Skip` über das Ende, ein Filter, den keine Entity besteht: `Entities != nil` und `len == 0`; bei einem Fehler `Entities == nil` und `Anzahl == 0` | `nil` statt leerer Liste; ein Teilergebnis neben dem Fehler | andere |
| 4 | `$skip` zuerst, dann `$top`; `$skip` ab dem Ende und `$top` 0 ergeben die leere Liste; `$top` über dem Rest liefert den Rest (`SPEC-191` Nr. 4) | `TestPaging`: Schleife über `Top` und `Skip` je aus {nicht gesetzt, 0, 1, n−1, n, n+1, 2147483647} (49 Paare) × n ∈ {0, 1, 5} = 147 Fälle gegen den Ausschnitt der geordneten Liste (Referenz im Test) | `$top` vor `$skip`; `Skip == n` als Fehler; `Skip > n` nicht abgefangen; Summe `Skip + Top` in `int32` gerechnet (Überlauf bei 2147483647 + 2147483647) | andere |
| 5 | Mit `Gesetzt` falsch wird `Wert` nicht gelesen (`SPEC-190` Nr. 3 a, `SPEC-191` Nr. 4) | `TestGesetztFalsch`: `Top{false, 5}` und `Skip{false, −3}` und `Skip{false, 2147483647}` verhalten sich wie nicht gesetzt, kein Fehler | `Wert` ohne `Gesetzt` gelesen (`Top{false, 0}` leert) | andere |
| 6 | Programmfehler der Stufe 1 sind `ErrUngueltig` ohne Fehlerwert (`SPEC-190` Nr. 3 a bis f) | `TestAbfrageProgrammfehler` (48 Fälle) und `TestStufe1Schluessel` (jede der 4 Key-Properties × fehlt, `null`, `redigiert` × 5 Positionen = 60; `Schluessel` leer an 8 Positionen; 28 Paare gleicher `Schluessel` unter 8 Entities); Schleife über 15 Fälle: `Top` und `Skip` je −1 und `MinInt32`; `Filter` mit `Ausdruck` `nil`; `Select{}`; ein Name doppelt in `Select.Namen` (erster und letzter von drei); ein Name doppelt in `OrderBy`; `Schluessel` leer; `Schluessel` doppelt (erste und letzte Entity); eine Key-Property fehlt, ist `null`, ist `redigiert`; je mit gültigem Rest und je mit leerer und mit nichtleerer Eingabe: `errors.Is(ErrUngueltig)`, kein `*AbfrageFehler`, kein `*FilterTypFehler`, `Ergebnis{}` | `Top` −1 als 0 gelesen; `Ausdruck` `nil` als „kein Filter“; doppelten Namen übergangen; leeren Schlüssel angenommen; `null` als Schlüssel angenommen; Prüfung nur der ersten Entity | andere |
| 7 | Ein Programmfehler beim Auswerten kommt unverändert zurück, nie als falsch (`SPEC-190` Nr. 3 g) | `TestAuswertungProgrammfehler`: Schleife über die Filter-Property mit dem Wert `redigiert`, eine Entity mit `int32` für eine `Edm.String`-Property (Klassen nicht vergleichbar), unbekannte Art, Form verfehlt (`True`, `01`): `ErrUngueltig`, kein Ergebnis; in `Links` und in `Rechts` eines Vergleichs | Fehler als `WahrheitFalsch` zurückgegeben; Entity still ausgelassen | andere |
| 8 | `PruefeFilter` wird vor der Auswertung mit Typ, Baum und Maske gerufen, auch bei leerer Eingabe; sein Fehler kommt unverändert (Sendung a; `SPEC-191` Nr. 2 a) | `TestPruefeFilterAufruf`: Schleife über je eine Abfrage der sieben Ursachen (`Foo eq 1`, `not Price gt 5`, `Id eq 'x'`, `Id eq 1.5`, Navigation, Guid mit `gt`, Maske falsch) × Eingabe (leer, 1, 5 Entities): derselbe `*FilterTypFehler` (`errors.As`, Ursache gleich, `Code()`); ohne Filter kein Aufruf der Maske | `PruefeFilter` nicht gerufen (`Id eq 'x'` ergibt die leere Liste statt `400`); nur bei nichtleerer Eingabe gerufen; `nil` statt der Maske übergeben; Fehler in `AbfrageFehler` verpackt | andere |
| 9 | Die Maske wird nur für Namen unterstützter Properties gerufen, einmal je Vorkommen, nach dem ersten Verstoß nicht mehr (`SPEC-190` Nr. 5; `SPEC-167` Nr. 3 a) | `TestMaskeAufrufe`: der Fake hält jeden Namen und zählt; Schleife über die Stelle (Filter, `$orderby`, `$select`) × Namensart (unterstützt, unbekannt, Navigation, Name im offenen Typ, Typ außerhalb der acht, `Edm.Guid` in `$orderby`) × Position (erster, letzter von drei): gerufen genau für die unterstützten, je Vorkommen einmal; nach einem Verstoß ist der Zähler stehen geblieben | Maske auch für unbekannte Namen; zweimal je Name; nach dem Verstoß weiter gerufen; Maske vor der Prüfung des Typs | andere |
| 10 | Nur `wahr` lässt eine Entity durch; `unbekannt` nicht, auch nicht die Wurzel `null` (`SPEC-191` Nr. 2 c; `LH-FA-08.b`) | `TestFilterDurchlass`: Wurzeln aus Vergleich wahr und falsch, `Edm.Boolean`-Property mit den Werten `true`, `false`, `null`, Literal `true` und `false`, `null`, `not null`, `null and true`, `null or true`, `null or false` gegen die Erwartung je Entity | `unbekannt` lässt durch; Wurzel `null` lässt alle durch; `null or true` als unbekannt | andere |
| 11 | Die Null-Semantik der sechs Operatoren gilt auf den Werten der Entities (`SPEC-191` Nr. 2 c; `SPEC-166` Nr. 4) | `TestFilterNull`: Entities, deren Property `null`, der leere Text, `0`, `false` ist; Schleife über sechs Operatoren und beide Seiten von `null` (Property links, `null` links) gegen die Tabelle von `LH-FA-08.b`; `P eq null` und `null eq P` | `'' eq null` wahr; `null ne null` wahr; `ge` bei zwei `null` falsch; `null` links anders als rechts behandelt | andere |
| 12 | Die sechs Operatoren vergleichen jeden unterstützten Typ richtig, mit der Property links und rechts (`SPEC-191` Nr. 2 c) | `TestFilterOperatoren`: Schleife über 7 Typen × 6 Operatoren und `Edm.Guid` × `eq`/`ne` (nur diese zwei lässt `SPEC-167` Nr. 5 zu) × 2 Seiten (Property links, Literal links) = 88 Gruppen × 3 Lagen (Wert kleiner, gleich, größer als das Literal) = 264 Läufe; `TestFilterZahlKlassen`: `int32`, `int64`, `decimal` gegeneinander als Properties (6 geordnete Paare × 6 Operatoren × 3 Lagen = 108; ein Literal wie `1.0` ist kein kanonischer Text, `SPEC-167` weist es als `literal_bereich` ab) | `gt` wie `ge`; Operanden vertauscht (`lt` an der falschen Seite); `Edm.Int64` über `float64` | andere |
| 13 | Ein Wahrheitswert als Operand eines Vergleichs ist `boolean`, `unbekannt` ist `null` (`SPEC-191` Nr. 2 c) | `TestWahrheitAlsOperand`: `(P eq 1) eq true`, `(P eq 1) eq false`, `(not null) eq null`, `(not null) eq true`, `(not null) ne null`, `(null and true) eq false` gegen die Erwartung, je mit einer Entity, die `P` 1, 2 und `null` hat | `unbekannt` als `false` in den Vergleich gereicht; `unbekannt` als `true` | andere |
| 14 | Jede Property des Filters muss in jeder Entity stehen, sonst `422`, statisch und ohne Kurzschluss (`SPEC-191` Nr. 2 b) | `TestFilterEigenschaftFehlt`: Schleife über die sieben Stellen der Property (Wurzel, Operand von `not`, `Links` und `Rechts` von Vergleich, von `and`, von `or` bei wahrem Gegenüber) × Position der Entity ohne die Property (erste, mittlere, letzte von 5) × Verknüpfung (`Id eq 1 or P eq 'x'` mit `Id eq 1` für jede Entity wahr): `AbfrageFehler{$filter, nicht_beobachtet}`, `ErrFilterNichtBeobachtet`, `Ergebnis{}`; derselbe Filter auf lauter Entities mit `P` (Wert `null` eingeschlossen) gilt | fehlende Property als `null` gedeutet; Entity ausgelassen; nur die erste Entity geprüft; nur die Wurzel geprüft; Kurzschluss bei `or`/`and` | andere |
| 15 | Aufsteigend steht `null` zuerst, `desc` kehrt um und `null` steht zuletzt; gleiche Werte bleiben gleich (Sendung c; `SPEC-191` Nr. 3 b) | `TestOrdnungTypen`: Schleife über die 7 Typen mit Ordnung (`string`, `boolean`, `int32`, `int64`, `decimal`, `date`, `datetimeoffset`): je 5 Werte plus `null` plus ein doppelter Wert, `asc` und `desc`, Eingabe in 3 Permutationen, gegen die ausgeschriebene Liste; `desc` ist genau die Umkehr der Werte mit `null` am Ende | `null` aufsteigend zuletzt; `desc` kehrt die Werte, lässt `null` zuerst; Textvergleich für `decimal`; Datum nach Monat vor Jahr | andere |
| 16 | Gemischte Schlüssel: jedes Element hat Typ und Richtung, das erste verschiedene entscheidet, `null` in jedem (`SPEC-191` Nr. 3 a, b) | `TestGemischteSchluessel`: 3 Properties verschiedener Typen (`string`, `decimal`, `datetimeoffset`) mit `null` und Gleichständen in jeder, 12 Entities; Schleife über alle 6 Folgen der drei Schlüssel × 8 Richtungen = 48 Abfragen gegen eine lexikographische Referenz im Test | letzter Schlüssel zuerst gelesen; die Richtung des ersten gilt für alle; der zweite Schlüssel nur bei `null` im ersten; Abbruch nach dem ersten Schlüssel | andere |
| 17 | Die Zahl der Schlüssel ist nicht begrenzt (`SPEC-190` Nr. 7) | `TestOrderByLaenge`: ein Typ mit 33 Properties; `OrderBy` mit 1, 2, 3, 32 und 33 Elementen, wobei jedes Element einmal allein die Gleichstände auflöst (1 + 2 + 3 + 32 + 33 = 71 Läufe); `TestNamen` führt fünf Navigation-Properties | nur die ersten 32 gelesen; nur die ersten zwei gelesen; ein Element mitten in der Liste (Index 3, 16, 31) überlesen; nur die erste oder die ersten vier Navigations ausgewertet | andere |
| 18 | Gleichstand: der Schlüssel, aufsteigend nach den Werten, typisiert, in der Folge von `Keys`; ohne `$orderby` nur der Schlüssel (`SPEC-191` Nr. 3 d) | `TestGleichstandSchluessel`: Schlüssel `1`, `2`, `10` (nicht `1`, `10`, `2`); zusammengesetzt `Id`, `Region` mit `Keys` in anderer Folge als `Properties`; `Edm.Guid` nach Text; `Edm.String` mit `B` vor `a`; `$orderby` absteigend, dessen Gleichstand aufsteigend ausfällt; ohne `$orderby` | Text von `Schluessel` statt Werte; Folge der Properties statt `Keys`; Gleichstand folgt der Richtung des letzten Elements; Gleichstand folgt der Eingabe | andere |
| 19 | Das Ergebnis hängt nur von der Menge ab, nicht von der Reihenfolge der Eingabe; gleiche Eingabe, gleiches Ergebnis (`SPEC-190` Nr. 1 b, 6; `LH-QA-02`) | `TestPermutation`: 6 Entities (720 Permutationen) × 3 Abfragen (Gleichstände in `$orderby`, Filter mit Paging, ohne Optionen) gegen die Folge der ersten Permutation; 100 Wiederholungen derselben Eingabe, tief gleich | Ordnung entfällt (Eingabefolge bleibt); Ordnung nur nach `OrderBy`, Gleichstand nach der Eingabe | andere |
| 20 | `$orderby` nach `Edm.Guid` ist `501`, `ODR-E6001`, `ordnung_nicht_unterstuetzt`, bei jeder Eingabe (Sendung b; `SPEC-190` Nr. 4, 5) | `TestOrderByGuid`: `OrderBy` mit der Guid-Property an erster, mittlerer und letzter Stelle von drei × Eingabe (leer, 5) × `Top` (nicht gesetzt, 0) × Richtung; die Guid-Property im Filter (`eq`, `ne`) gilt; `HatOrdnung` wahr für die sieben anderen | `HatOrdnung` nicht gefragt; nur bei nichtleerer Eingabe geprüft; nur das erste Element geprüft; Guid nach Text geordnet | andere |
| 21 | Die Namen in `$orderby` und `$select` haben die Ursachen und Klassen der Tabelle (`SPEC-190` Nr. 4) | `TestNamen`: Schleife über Stelle (`$orderby`, `$select`) × Namensart (`Foo`, `id`, `ID`, ` Id`, `Id `, leer, Navigation, offener Typ mit unbekanntem Namen, geschlossener Typ mit demselben, 29 Typnamen außerhalb der acht, Maske falsch) × Position (erster, letzter von drei): Ursache, Option, Klasse; `*` mit unbekanntem Namen daneben; Name `Id` gültig | Suche ohne Schreibweise; `Open` ignoriert; Navigation als unbekannt; nur der erste Name geprüft; Maske vor dem Typ (`422` statt `501`); Namen neben `*` nicht geprüft | andere |
| 22 | Jede Property von `$orderby` muss in jeder Entity stehen, sonst `422` (`SPEC-191` Nr. 3 c) | `TestOrderByEigenschaftFehlt`: Position der Entity ohne die Property (jede der 5) × Element (alle drei) × Richtung = 30; dazu mit Filter (erfüllt für jede Entity, für keine, `not`) × Element × Position = 45; Property mit Wert `null` gilt; Option im Fehler ist `$orderby` | fehlende Property als `null` einsortiert; nur die erste Entity geprüft; Stufe 5 prüft `$orderby` nicht, sobald ein Filter gesetzt ist oder dieser ein `not` ist | andere |
| 23 | Form und Klasse des `AbfrageFehler`: Ursache, Code, Text, kein Teil der Eingabe (`SPEC-190` Nr. 4) | `TestAbfrageFehler`: Schleife über die vier Ursachen × drei Optionen: `errors.Is` trifft genau die Sentinel der Klasse (keine der beiden anderen), `Code()`, `Error()` nach Muster; der Name `GEHEIM` aus der Eingabe steht nicht im Text; eine Ursache `x` hat die Klasse `ungueltig` | `Is` auch für die falsche Klasse; Name im Text; `ordnung_nicht_unterstuetzt` als `ODR-E6005` | andere |
| 24 | Die drei Klassen werden über die Abbildung zu `ODR-E6005`, `ODR-E6001`, `ODR-E6003` (`SPEC-109` Nr. 5, `SPEC-190` Nr. 4) | Test in `abbildung_test.go`: Schleife über die vier Ursachen, roh, verpackt (`fmt.Errorf` mit `%w`) und in `errors.Join` mit einem fremden Fehler: `Abbilden` liefert den Code der Klasse | Sentinel nicht getroffen (Fall `ODR-E9001`); `ordnung_nicht_unterstuetzt` auf `ODR-E6005` abgebildet | andere |
| 25 | Der erste Verstoß der Stufen entscheidet, gleich welcher Klasse (`SPEC-190` Nr. 5) | `TestRangfolge`: alle 110 geordneten Paare aus den 11 Verstößen (Programmfehler der Stufe 1; Fehler des Filters; `$orderby`: unbekannt, nicht unterstützt, Guid, nicht beobachtet; `$select`: unbekannt, nicht unterstützt, nicht beobachtet; Entity ohne Filter-Property; Entity ohne `$orderby`-Property) gegen die Stufenfolge; innerhalb von `$orderby` und `$select` das erste Element; derselbe Name als Guid und mit falscher Maske | Stufen vertauscht (`$select` vor `$orderby`); `400` vor `501` nach Klasse; Entity-Prüfung vor der Prüfung der Namen; Programmfehler nach dem Filter | andere |
| 26 | Die Verstöße der Stufen 1 bis 4 hängen nicht von den Entities ab (`SPEC-190` Nr. 5, Folge) | `TestStatischOhneEntities`: die neun Verstöße der Stufen 1 bis 4 aus Zeile 25 × Eingabe (`nil`, leer, 1, 5 Entities): derselbe Fehler | leere Eingabe beendet früh mit der leeren Liste, ohne zu prüfen | andere |
| 27 | Die Projektion: ohne `$select` und mit `*` alle bekannten Properties, mit Namen nur diese in der Folge der Eingabe (`SPEC-191` Nr. 5 a, b) | `TestProjektion`: Schleife über `Select` (`nil`, `Alle`, ein Name, alle Namen, ein Teil, `*` mit Namen, Namen in umgekehrter Folge) × den Zustand der Property in der Entity (bekannt, `null`, fehlend) × 4 Properties; `B,A` und `A,B` ergeben dasselbe; `TestProjektionNamen`: 1 bis 9 Namen als Anfang, als Ende in umgekehrter Folge und jeder Name allein (27 Fälle) | Folge der URL statt der Eingabe; `Alle` ignoriert; Key-Property ergänzt; Namen neben `*` ignoriert; nur die ersten 4 oder 8 Namen gelesen | andere |
| 28 | `null` und fehlend bleiben unterscheidbar (`SPEC-191` Nr. 5 b; `SPEC-044`; `LH-FA-18`) | `TestNullUndFehlend`: Entity A mit `P` gleich `null`, B ohne `P`, C mit `P` gleich 5; `$select=P`: A trägt `P` mit Art `null`, B trägt `P` nicht, C trägt 5; ebenso ohne `$select`; Länge von `Properties` je Entity | fehlende Property als `null` ergänzt; `null` entfernt | andere |
| 29 | Eine Property, die die Maske nicht kennt, ist `422`; eine, die nur eine Entity nicht kennt, wird in `$select` ausgelassen (`SPEC-191` Nr. 6; Teilprojektion) | `TestSelectMaskeUndTeilprojektion`: Maske falsch für `P`, jede Entity trägt `P`: `nicht_beobachtet` (`$select`); Maske wahr oder `nil`, drei von fünf Entities ohne `P`: kein Fehler, `P` fehlt in genau diesen | Maske bei `$select` ignoriert; Auslassung als `422`; Auslassung als `null` | andere |
| 30 | Das Ergebnis trägt `Typ` und `Schluessel`, `Uebersprungen` ist `nil`, und es teilt nichts mit der Eingabe (`SPEC-191` Nr. 5 c, `SPEC-190` Nr. 2) | `TestErgebnisUnabhaengig` (Zeile 39 führt die Menge der Ergebnisse): `Typ`, `Schluessel` gleich; `Uebersprungen == nil` bei Eingabe mit zwei Einträgen; jedes Element von `Properties` jeder der 5 Entities des Ergebnisses ändern, dann die Eingabe und ein zweites Ergebnis prüfen; die Eingabe ändern, dann das Ergebnis prüfen; ohne `$select` (der Fall `Alle`) und mit Namen | `Properties` der Eingabe weitergereicht (Alias bei `Alle`); `Uebersprungen` kopiert | andere |
| 31 | Die Eingabe bleibt unverändert: Reihenfolge der Liste, Entities, `Optionen`, Baum, auch bei einem Fehler (`SPEC-190` Nr. 2, 6) | `TestEingabeUnveraendert`: `reflect.DeepEqual` der Eingabe vor und nach dem Aufruf, über die Abfragen aus Zeile 1, 16, 19 und über je einen Fehler jeder Stufe; Wirkungs-Fänger zur Nebenläufigkeit (rot in jedem Lauf, nicht erst unter dem Detektor) | die Eingabeliste in place sortieren (`copy(entities, geordnet)`); `Select.Namen` vertauschen; die `FilterOption` des Baums ändern (die Knoten sind Werttypen, nur der Zeiger auf `FilterOption` ist veränderbar) | andere |
| 32 | Gleichzeitige Aufrufe auf derselben Eingabe beeinflussen sich nicht (`SPEC-190` Nr. 6) | `TestAbfrageGleichzeitig`: 8 Goroutinen mit je 1000 Aufrufen über 40 Abfragen (statt 12: der Abdeckungslauf des Tests allein verlangte mehr) auf geteilter Liste, geteilten `Optionen` und geteiltem Baum, je eine für jeden Zweig (Filter, Ordnung ohne und mit `OrderBy`, `$skip`, `$top`, `$select` mit `Alle` und mit Namen, jede der vier Ursachen, jeder Programmfehler der Stufen 1 und 6, Schlüssel mit Guid und mit Gleichstand); jedes Ergebnis gleich dem sequentiellen Lauf, Text, Code und Klasse der Fehler gelesen, die Zahl der Aufrufe je Abfrage gezählt; unter `make test-race` | in place sortieren der geteilten Liste; Cache in einer `map` des Pakets; geteilter Puffer für die Schlüssel; `Select.Namen` in place ordnen | nebenläufig |
| 33 | Die Prüfgrenze 10.000 Entities und 33 Schlüssel: der Code kennt keine Obergrenze (`SPEC-190` Nr. 7; `LH-QA-01.a`) | `TestGrosseMenge`: 10.000 und 10.001 Entities (Schlüssel in fester Mischung), Filter `Id gt 100`, `$orderby` absteigend, `$skip=5`, `$top=50`, gegen die Referenz; `Anzahl` 9.900 und 9.901; Gleichstände auf 1.000 Werten; Zeile 38 führt dieselben Zahlen durch jede Stelle der Pipeline (ohne Filter und ohne Paging) | eingebaute Grenze 10.000 (Abschneiden) in Filter, Ordnung, Seite, Anzahl, Projektion, Stufe 1 und Stufe 5 | andere |
| 34 | Ende zu Ende: der Query-Text des Clients führt über Parser und Interpreter zum erwarteten Teilbestand (`LH-FA-08` Happy Path, Boundary; `SPEC-162` Nr. 4) | `TestAbfrageAusURL` unter `tests/integration/`: 14 Query-Texte über `urlparser.New(WithFilter(NewFilterLeser()))`: kombinierte Abfrage mit fünf Datensätzen, `$top=0`, `$skip` über das Ende, `$count=true`, `Name desc`, `$select=*,Name`, `$filter` mit `null`; `$top=-1` und `$top=1.5` enden im Parser mit `wert` (der Interpreter wird nicht erreicht) | `Absteigend` des Parsers nicht gelesen; `Select.Alle` ignoriert; `Top` 0 als „nicht gesetzt“ | andere |
| 35 | Der Rest der Ordnung ist der Text von `Schluessel`, bytegleich aufsteigend, bei `typ` ohne `Keys` und bei gleichen Key-Werten, gleich in welcher Richtung das letzte Element von `OrderBy` ordnet (`SPEC-191` Nr. 3 e) | `TestSchluesselRest`: 13 Texte (`10` vor `2`, Länge, Groß und Klein, führende Null, Leerraum, ein Zeichen jenseits von ASCII; aus Bytewerten gebaut) × 4 Formen (Typ ohne Keys; gleiche Werte; gleiche Werte mit `N` aufsteigend und absteigend) × 3 Reihenfolgen gegen `sort.Strings` | Länge zuerst; Kleinschreibung gefaltet; führende `0` abgeschnitten; absteigend bei absteigendem `OrderBy`; nur die ersten 2 Bytes | andere |
| 36 | Der Schlüssel ordnet nach den Werten seines Typs, in der Folge von `Keys`, je Key-Property eine Spalte (`SPEC-191` Nr. 3 d) | `TestSchluesselTypen`: die 8 Typen mit Schlüsselordnung (`Edm.Int32` negativ, `Edm.Int64` über 2^53, `Edm.Decimal`, `Edm.DateTimeOffset` mit Bruchteil, `Edm.String` mit `B` vor `a`, `Edm.Boolean`, `Edm.Date`, `Edm.Guid` mit Hex-Buchstaben und letzter Gruppe gegen die Ordnung) × 3 Reihenfolgen, `Schluessel` in umgekehrter Folge der Werte; `TestSchluesselLaenge`: 1, 2, 3 und 4 Key-Properties, jede einmal allein entscheidend (10 Fälle) × 3 Reihenfolgen | Text statt Wert je Typ (`Edm.Int64`, `Edm.Decimal`, `Edm.DateTimeOffset`, `Edm.Int32`); Guid nur über die letzte Gruppe; Hex-Buchstaben vor Ziffern; `Edm.Boolean` umgekehrt; nur 2 oder 3 Keys gelesen | andere |
| 37 | Ein Wert, den kein Leser liefert, in einer Spalte der Ordnung ist `ErrUngueltig` und kein Ergebnis (`SPEC-190` Nr. 3 g, Nr. 5 Stufe 6) | `TestOrdnungProgrammfehler`: 5 Formen des Werts (`redigiert`, unbekannte Art, Form `01`, andere Klasse, Guid unter Zahlen) × 4 Spalten (drei Elemente von `$orderby`, zweite Key-Property) × 5 Positionen der Entity = 100, dazu eine Spalte nur aus Guids | Fehler von `pruefeWert`, „Klassen nicht vergleichbar“ und „Ordnung auf Guid“ verschluckt; Fehler einer Key-Spalte oder einer `$orderby`-Spalte übergangen; jeder Fehler von `ordne` verschluckt; Klasse Text ausgenommen | andere |
| 38 | Die Prüfgrenze 10.000 und 10.001 gilt an jeder Stelle der Pipeline (`SPEC-190` Nr. 7) | `TestGrosseMengeStellen`: 2 Zahlen × 8 Stellen: Stufe 1 (doppelter Schlüssel und Key `null` an der letzten Entity), Stufe 5 (letzte Entity ohne die Property von `$filter`, ohne die von `$orderby`), Filter, Ordnung, Seite, Projektion; `Anzahl` und `Entities` in jeder Abfrage | Ordnung auf 10.000 abgeschnitten; `seite` und `Anzahl` auf 10.000 begrenzt; Stufe 1 und Stufe 5 prüfen nur 10.000 Entities; Projektion abgeschnitten | andere |
| 39 | Das Ergebnis gehört dem Aufrufer, auch ein leeres, und `AlleAbfrageUrsachen` liefert je Aufruf eine eigene Liste (`SPEC-190` Nr. 2, 6) | `TestErgebnisUnabhaengig`: 3 Formen von `Select` × 5 Entities × 3 Richtungen (in das Ergebnis schreiben, in ein zweites Ergebnis sehen, in die Eingabe schreiben), 4 Arten leerer Ergebnisse, `AlleAbfrageUrsachen`; `TestAbfrageGleichzeitig` schreibt in jedes Ergebnis (jede Property, Anhängen) | Alias der Properties ab der zweiten oder der vierten Entity; ein leeres Ergebnis aus einer Paketvariablen; `AlleAbfrageUrsachen` aus einer Paketvariablen; ein gemeinsamer Puffer der Projektion | andere, nebenläufig |

**Akzeptierte Negative** (Grund und Folge im Vertrag, `SPEC-190` *Ungeprüft* bzw. `SPEC-191` *Ungeprüft*; keine Zeile der Tabelle, weil kein Fänger möglich oder sinnvoll ist):

| Punkt | Zusage ohne Fänger | Folge, wenn der Code regrediert |
|---|---|---|
| Laufzeit bei 10.000 Entities | keine Zusage über Latenz | eine quadratische Ordnung bliebe bei grünen Gates unbemerkt, solange `TestGrosseMenge` unter einer Sekunde bleibt; den Lasttest zu `LH-QA-01.a` trägt `welle-betrieb` |
| Entities abgeleiteter Typen | `typ` ist der Typ des Sets | ein Filter auf eine Property eines abgeleiteten Typs ist `400`; die Folge ist beschrieben, ein Fänger würde sie nur festhalten |
| Passt `Schluessel` zu den Werten der Key-Properties | die Ordnung folgt den Werten | bei abweichender Eingabe ordnet der Code nach den Werten, der `Schluessel` dient der Eindeutigkeit |
| Doppelter Name in `Properties` einer Entity | der erste Eintrag gilt | der Leser liefert eine Property je Name (`SPEC-126` Nr. 9) |
| `Filter.Ausdruck == nil` in Stufe 1 (`SPEC-190` Nr. 3 b) | die Prüfung allein hat keinen Fänger | äquivalente Mutation (`M-AusdruckNilNurStufe1`): ohne sie liefert `PruefeFilter` für `nil` ebenfalls `ErrUngueltig`, in derselben Klasse; gefangen ist die Zusage (kein Filter ohne Ausdruck, `TestAbfrageProgrammfehler`) erst, wenn auch `PruefeFilter` und `filtere` den Fall überspringen (`M-AusdruckNil`, rot); die Prüfung in Stufe 1 bleibt als ausdrückliche Zusage der Rangfolge stehen |
| Key-Property fehlt oder ist `redigiert` (`SPEC-190` Nr. 3 f) | Stufe 1 ohne Fänger, der von Stufe 6 unterscheidet | äquivalente Mutationen (`R06h`, `R06j`): `ordne` liest die Key-Property und endet mit `ErrUngueltig` in derselben Klasse; gefangen ist `null` (`R06i`) und die Position der Entity (`R06k`, `R06l`); Folge: fällt die Prüfung in Stufe 1, ändert sich nur, welche Stufe den Programmfehler meldet |
| `WertAusLiteral` liefert einen Fehler, den der Filter verschluckt (`R07j`) | ein Literal ohne `'` außen | äquivalente Mutation: der verschluckte Fehler lässt `Wert{}` zurück, `Vergleiche` meldet `ErrUngueltig` (Art des Wertes) in derselben Klasse |
| Paar (Entity ohne `$orderby`-Property, `$orderby` nicht beobachtet) in Zeile 25 | beide Verstöße tragen `$orderby` und `nicht_beobachtet`, die Kennung unterscheidet sie nicht | beide Reihenfolgen liefern denselben Fehlerwert; ein Vertauschen der Stufen 3 und 5 für genau dieses Paar bliebe ungesehen, während `M-PaarVertauscht` (Stufe 5 vor Stufe 2) und `R25a` bis `R25d` für jedes andere Paar rot werden |
| Ein Datenrennen, das kein Lauf auslöst | `make test-race` | `AGENTS.md` §3.10: das Fehlen der Sperre fängt der Detektor, in place sortieren fängt zusätzlich Zeile 31 in jedem Lauf |

**Mengen der Zusagen** (Menge · Achsen · Elementzahl je Achse · eine Mutation an einem Element, das keine Zeile oben nennt; `AGENTS.md` §3.17):

| Menge | Achsen und Elementzahl | Mutation an einem nicht genannten Element |
|---|---|---|
| Typen mit Ordnung (Zeilen 12, 15, 16) | Typ 7 (Guid nur `eq`/`ne`: 8 mit Gleichheit), Operator 6, Lage 3, Seite 2, Richtung 2, Wert `null`/Wert 2 | `Edm.Date` in `desc` falsch umgekehrt; `boolean` mit `null` in `desc` zuerst |
| Namen der Optionen (Zeilen 9, 21) | Stelle 3, Namensart 11 (darunter 29 Typnamen), Position 2, Option 2 | der 29. Typname (`Edm.GeometryCollection`) als unterstützt; Name im Mittelfeld einer Liste von drei nicht geprüft |
| Paging (Zeile 4) | `Top` 7, `Skip` 7, n 3 | `Top == n−1` mit `Skip == 1` liefert ein Element zu viel |
| Entities der Eingabe (Zeilen 14, 19, 22, 33) | Zahl 0, 1, 5, 6, 10.000, 10.001; Permutationen 720; Position der Entity ohne Property 3 | die vorletzte Entity ohne Property wird nicht geprüft |
| Schlüssel von `$orderby` (Zeilen 16, 17) | Länge 1, 2, 3, 32, 33; Folge 6; Richtung 8 | Gleichstand erst im 33. Schlüssel nicht aufgelöst |
| Verstöße (Zeilen 25, 26) | 11 Arten, 110 geordnete Paare, Eingabe 4 | Paar (Entity ohne `$orderby`-Property, Fehler des Filters) vertauscht |
| Stellen der Property im Filter (Zeile 14) | Stelle 7, Position 3, Verknüpfung 2 | Property nur in `Rechts` von `or` nicht geprüft |
| Zustand einer Property in der Entity (Zeilen 11, 27, 28, 29) | bekannt, `null`, fehlend; Typ 8 | `Edm.Boolean` mit Wert `null` im Filter als `false` gelesen |
| Aufrufstellen der Maske (Zeile 9) | `PruefeFilter`, `$orderby`, `$select`: 3 | Maske in `$select` nicht gerufen, wenn `Alle` gesetzt ist |

**Beleg des Implementers: Zusage · Mutation · roter Test** (`AGENTS.md` §3.10). Jede Mutation ist eine einzelne Änderung an `src/hexagon/model/abfrage*.go`, auf einer Wegwerfkopie angewandt (Werkzeug im Scratchpad des Implementers, nicht im Repo); der Test war aus dem benannten Grund rot, nicht weil der Aufbau brach (Mutationen, die nicht übersetzten, sind nicht gezählt). Gemessen vom Implementer: 147 Läufe, die übersetzten und liefen; 4 Mutationen blieben grün und sind äquivalent (`R06h`, `R06j`, `R07j`, `M-AusdruckNilNurStufe1`; Tabelle der Negative); 3 weitere überlebten im ersten Lauf (`R07c`, `R07e`, `R07k`: Fehler in einem verschachtelten Operanden) und wurden durch die verschachtelten Fälle von `TestAuswertungProgrammfehler` rot.

| Zeile | Mutation (Kennung der Läufe) | roter Test |
|---|---|---|
| 1 | Filter nach Paging (`R01a`); Ordnung nach Paging (`R01b`); `$top` ohne `$skip` (`R01c`) | `TestStufen` |
| 2 | `Anzahl` nach Paging (`R02a`), nur bei `Count` (`R02b`), ohne Filter (`R02c`) | `TestAnzahl` |
| 3 | `nil` statt leerer Liste (`R03a`); Teilergebnis neben dem Fehler (`R03b`, `R03c`) | `TestEntitiesNieNil` |
| 4 | `$skip` ohne Grenze (`R04a`); `Skip == n` verschoben (`R04b`); Summe in `int32` (`R04c`); `Top` ohne Skip +1 (`M-TopOhneSkip`) | `TestPaging` |
| 5 | `Wert` ohne `Gesetzt` gelesen (`R05a` bis `R05d`) | `TestGesetztFalsch` |
| 6 | Select ohne Namen, doppelter Name, doppelter `OrderBy`, leerer und doppelter Schlüssel, Key `null`, nur erste oder letzte Entity geprüft (`R06c` bis `R06l`) | `TestAbfrageProgrammfehler` |
| 7 | Fehler als falsch, Entity ausgelassen, Fehler in `not`, Links, Rechts, `and`, `or` verschluckt (`R07a` bis `R07l`) | `TestAuswertungProgrammfehler` |
| 8 | `PruefeFilter` nicht gerufen (`R08a`), nur bei nichtleerer Eingabe (`R08b`), `nil` statt Maske (`R08c`), Fehler verpackt (`R08d`) | `TestPruefeFilterAufruf` |
| 9 | Maske für unbekannte Namen (`R09a`), doppelt (`R09b`), nach dem Verstoß (`R09c`), vor dem Typ (`R09d`), bei `Alle` nicht (`R09e`) | `TestAbfrageMaskeAufrufe` |
| 10 | `unbekannt` lässt durch (`R10a`); Wurzel `null` (`R10b`); `null or true` unbekannt (`R10c`, `R10d`) | `TestFilterDurchlass` |
| 11 | Literal `null` als leerer Text (`R11a`); Operanden mit `null` gleichgesetzt (`R11b`) | `TestFilterNull` |
| 12 | Operanden vertauscht (`R12a`); `Edm.Int64` auf 15 Stellen gekürzt (`R12b`) | `TestFilterOperatoren` |
| 13 | `unbekannt` als `false` (`R13a`) und als `true` (`R13b`) | `TestWahrheitAlsOperand` |
| 14 | fehlende Property als `null` (`R14a`); nur erste, letzte, vorletzte Entity (`R14b`, `R14c`, `M-Vorletzte`); je Stelle im Baum (`R14e` bis `R14j`) | `TestFilterEigenschaftFehlt` |
| 15 | `desc` lässt `null` zuerst (`R15a`); `null` aufsteigend zuletzt (`R15b`); Text statt Dezimal (`R15c`); `desc` gekehrt fehlt (`R15d`); Monat vor Jahr (`R15e`) | `TestOrdnungTypen` |
| 16 | letzter Schlüssel zuerst (`R16a`); Richtung des ersten für alle (`R16b`); zweiter nur bei `null` (`R16c`); Abbruch nach dem ersten (`R16d`) | `TestGemischteSchluessel` |
| 17 | nur 32 (`R17a`), nur 2 Elemente (`R17b`) | `TestOrderByLaenge` |
| 18 | Text statt Werte (`R18a`); Folge der Keys umgekehrt (`R18b`); Richtung des letzten (`R18c`); Eingabefolge (`R18d`); Guid nach Schlüssel (`R18e`) | `TestGleichstandSchluessel` |
| 19 | Ordnung entfällt (`R19a`) | `TestPermutation` |
| 20 | `HatOrdnung` nicht gefragt (`R20a`); nur bei nichtleerer Eingabe (`R20b`); nur das erste Element (`R20c`) | `TestOrderByGuid` |
| 21 | Schreibweise gefaltet (`R21a`); `Open` ignoriert (`R21b`); Navigation als unbekannt (`R21c`); nur der erste Name (`R21d`); Typ ungeprüft (`R21e`) | `TestNamen` |
| 22 | `$orderby`-Property nicht geprüft (`R22a`); falsche Option (`R22b`) | `TestOrderByEigenschaftFehlt` |
| 23 | `Is` für die falsche Klasse (`R23a`); Klasse der Ursache (`R23b`, `R23c`); Code (`R23d`, `R23g`); Text (`R23e`, `R23f`) | `TestAbfrageFehler` |
| 24 | `Is` immer falsch (`R24a`); Ursache auf die falsche Klasse (`R24b`, `R24c`) | `TestAbbildungAbfrage` |
| 25 | `$select` vor `$orderby` (`R25a`); Entity-Prüfung vor den Namen (`R25b`); Stufe 1 nach dem Filter (`R25c`); `$orderby` vor `$filter` in Stufe 5 (`R25d`); Fehler des Filters umgedeutet (`R25e`); Stufe 5 vor Stufe 2 (`M-PaarVertauscht`) | `TestRangfolge` |
| 26 | leere Eingabe beendet früh (`R26a`) | `TestStatischOhneEntities` |
| 27 | Folge der URL (`R27a`); `Alle` ignoriert (`R27b`); Key ergänzt (`R27c`) | `TestProjektion` |
| 28 | fehlende Property als `null` (`R28a`); `null` entfernt (`R28b`) | `TestNullUndFehlend` |
| 29 | Maske bei `$select` ignoriert (`R29a`); Auslassung als 422 (`R29b`) | `TestSelectMaskeUndTeilprojektion` |
| 30 | Alias der Properties (`R30a`); `Uebersprungen` kopiert (`R30b`); `Schluessel` fehlt (`R30c`) | `TestErgebnisUnabhaengig` |
| 31 | Liste in place (`R31a`); `Select.Namen` vertauscht (`R31b`); `FilterOption` geändert (`R31c`) | `TestEingabeUnveraendert` |
| 32 | siehe Raten unten | `TestAbfrageGleichzeitig` |
| 33 | Eingabe auf 10.000 abgeschnitten (`R33a`) | `TestGrosseMenge` |
| 34 | `Absteigend` nicht gelesen (`R34a`); `Alle` ignoriert (`R34b`); `Top` 0 als nicht gesetzt (`R34c`) | `TestAbfrageAusURL` |

**Mengen der Zusagen, gemessen** (Menge · gezählte Elemente · eine Mutation an einem Element, das keine Zeile oben nennt, rot gesehen):

| Menge | Achsen, im Test gezählt | Mutation, rot |
|---|---|---|
| Typen mit Ordnung | Zeile 12: 88 Gruppen, 264 Läufe; Zeile 15: 7 Typen × 2 Richtungen × 3 Reihenfolgen = 42; Zeile 16: 6 Folgen × 8 Richtungen = 48 | `Edm.Date` nach Monat vor Jahr (`R15e`): `TestOrdnungTypen` (Day), `Edm.Int64` gekürzt (`R12b`) |
| Namen der Optionen | `TestNamen` 2 × 39 × 3 = 234; Zeile 9: 7 Arten × 4 Stellen × 3 Positionen = 84 | der 29. Typname `Ns.Color` als unterstützt (`M-Typ29`): `TestNamen`; Name in der Mitte nicht geprüft (`M-NameMitte`): `TestNamen`, `TestAbfrageMaskeAufrufe` |
| Paging | 7 × 7 × 3 = 147 | `Top == n−1` ohne `Skip` liefert ein Element zu viel (`M-TopOhneSkip`): `TestPaging` |
| Entities der Eingabe | Zahl 0, 1, 5, 6, 10.000, 10.001; 720 Permutationen; Position der Entity ohne Property 5 von 5 (Zeile 14: 200 Fälle, Zeile 22: 30) | die vorletzte Entity wird nicht geprüft (`M-Vorletzte`, `M-Vorletzte22`): `TestFilterEigenschaftFehlt`, `TestOrderByEigenschaftFehlt` |
| Schlüssel von `$orderby` | Länge 1, 2, 3, 32, 33; Folge 6; Richtung 8 | nur 32 gelesen (`R17a`): `TestOrderByLaenge` (das 33. Element löst den Gleichstand auf) |
| Verstöße | 11 allein, 110 geordnete Paare, Eingabe 4 | Stufe 5 vor Stufe 2 (`M-PaarVertauscht`): `TestRangfolge` |
| Stellen der Property im Filter | 8 Stellen × 5 Verknüpfungen × 5 Positionen = 200 | Property nur in `Rechts` einer Verknüpfung nicht geprüft (`R14f`): `TestFilterEigenschaftFehlt` |
| Zustand einer Property | bekannt, `null`, fehlend; Typ 8 | `Active` mit Wert `null` im Filter als `false` (`M-BoolNull`): `TestFilterDurchlass`, `TestFilterNull` |
| Aufrufstellen der Maske | `PruefeFilter`, `$orderby`, `$select` mit und ohne `*`: 4 Stellen | Maske in `$select` bei `Alle` nicht gerufen (`R09e`): `TestAbfrageMaskeAufrufe` |

**Raten der Zeile 32** (Messender: der Implementer; 20 Prozesse je Mutation, jeder Prozess ein eigenes Testprogramm mit `-test.run ^TestAbfrageGleichzeitig$ -test.count=1`; Image `odata-recorder:race` mit `-race` (cgo) für den Detektor, `odata-recorder:deps` ohne `-race` für die Wirkung; Docker mit `--cpus 4 --memory 2g`, Last des Rechners 5 bis 16 bei 20 CPUs wegen zwei weiterer Agenten; Unmutierter Code: 0 von 20 in beiden Läufen):

| Mutation | Detektor (`-race`) | Wirkung (ohne `-race`) |
|---|---|---|
| Liste in place ordnen (`S1inplace`) | 20 von 20 (20 mit `DATA RACE`) | 20 von 20 |
| Cache als `map` im Paket (`S2cache`) | 20 von 20 (20 mit `DATA RACE`) | 20 von 20 |
| geteilter Puffer für die Schlüssel (`S3puffer`) | 20 von 20 (20 mit `DATA RACE`) | 20 von 20 |
| `Select.Namen` vertauscht (`S4namen`) | 20 von 20 (20 mit `DATA RACE`) | 5 von 20; das geteilte Ergebnisfeld fängt nur der Detektor (benanntes Negativ: ohne `-race` bleibt das Vertauschen in 15 von 20 Prozessen ungesehen, denn die Projektion hängt nicht von der Folge der Namen ab) |

Jede Funktion von `abfrage*.go`, die der Test erreicht, trägt einen Zähler: eine Schreibung auf eine Paketvariable am Anfang der Funktion wird vom Detektor in 5 von 5 Prozessen gemeldet, für 20 Funktionen (`FuehreAus`, `pruefeNamen`, `pruefeVorhandene`, `verlangeProperties`, `wertVon`, `pruefeArgumente`, `hatDoppelte`, `pruefeEntities`, `seite`, `projiziere`, `filterNamen`, `filtere`, `filterWert`, `operanden2`, `verknuepfeWerte`, `wahrheitAlsWert`, `pruefeName`, `ordne`, `baueSpalte`, `vergleiche`). `AlleAbfrageUrsachen`, `klasse`, `Code`, `Error`, `Is` und `abfrageVerstoss` liest der Test in den Goroutinen (kein gemeinsamer Zustand). Der Abdeckungslauf von `TestAbfrageGleichzeitig` allein (`-covermode=atomic`) betritt jede Funktion von `abfrage*.go`; offen bleiben nur die Anweisungen, die kein Aufruf erreicht: die Fehler `Property der Entity fehlt` und `Knoten des Filters` (Stufe 5 und `PruefeFilter` gehen voraus), `Operator der Verknüpfung` (`PruefeFilter` geht voraus). Ein Datenrennen im Code dieser Zweige fängt der Detektor nicht (akzeptiertes Negativ, Folge: ein Zugriff ohne Lauf bleibt ungeprüft).

**Fix-Runde nach dem Review (M-1 bis M-8, L-1, I-1)** (`AGENTS.md` §3.15, §3.17): Je Befund die Menge aus dem Wortlaut von `SPEC-190` und `SPEC-191`, der Fänger als Schleife über alle ihre Elemente mit Zählprüfung, und mindestens eine Mutation an einem Element, das das Review nicht nannte (`N-…`), mit `tools/harness/mutationslauf.sh` gefahren (Image `odata-recorder:race`, `--memory 2g --cpus=4`, unter `flock`; Last des Hosts 3,2 bis 5,7; Messender: der Implementer). Die vom Review genannten Mutationen (`V…`, `D…`, `W…`, `S…`) sind erneut gefahren und rot. Produktcode unverändert: kein Test zeigte einen Fehler.

| Befund | Menge · Elemente | Mutation, nicht im Review genannt, rot |
|---|---|---|
| M-1 Programmfehler der Ordnung (Zeile 37) | Spalten der Ordnung × Form des Werts × Position der Entity: 5 × 4 × 5 = 100, dazu eine Spalte nur aus Guids | andere Klasse mit Text ausgenommen (`N-M1-Klasse-Text`); `W1`, `W2`, `W3`, `W5`, `W6`, `W7` rot |
| M-2 Prüfgrenze an jeder Stelle (Zeile 38) | 2 Zahlen × 8 Stellen (Stufe 1 zweifach, Stufe 5 zweifach, Filter, Ordnung, Seite, Projektion); `Anzahl` in jeder | Projektion abgeschnitten (`N-M2a-Projektion`); Stufe 5 prüft nur 10.000 (`N-M2b-Stufe5`); `V07` bis `V10` rot |
| M-3 Rest der Ordnung (Zeile 35) | 13 Texte × 4 Formen × 3 Reihenfolgen = 12 Läufe | nur die ersten 2 Bytes (`N-M3-zwei-Bytes`); `V03`, `V04`, `D3`, `D4` rot |
| M-4 Schlüssel als Spalte (Zeile 36) | 8 Typen × 3 Reihenfolgen = 24 Läufe; 1, 2, 3, 4 Key-Properties × jede entscheidend = 10 Fälle × 3 = 30 Läufe | `Edm.Boolean` umgekehrt (`N-M4a-Boolean`); nur 3 Keys gelesen (`N-M4b-drei-Keys`); `V01`, `V02`, `V17` bis `V20`, `D1` rot |
| M-5 Stufe 1 (Zeile 6) | 4 Key-Properties × 3 Zustände × 5 Positionen = 60; Schlüssel leer an 8 Positionen; 28 Paare unter 8 Entities | nur 3 Keys in Stufe 1 (`N-M4c-drei-Keys`); nur 5 Schlüssel gemerkt (`N-M5-fuenf`); `V22`, `V23`, `D2` rot |
| M-6 Filter und `$orderby`-Property (Zeile 22) | 3 Filter × 3 Elemente × 5 Positionen = 45 | Stufe 5 prüft `$orderby` nicht bei einem `not` (`N-M6-FilterNicht`); `V15` rot |
| M-7 Positionen in Listen (Zeilen 17, 21, 27) | 71 Läufe (Länge × entscheidendes Element); 43 Namensarten × 2 Stellen × 3 Positionen = 258; 27 Fälle mit 1 bis 9 Namen | Element 32 überlesen (`N-M7a-Index31`); nur 4 Navigations (`N-M7b-Nav4`); neunter Name ignoriert (`N-M7c-Name9`); `V05`, `V06`, `V13`, `V21` rot |
| M-8 Ergebnis gehört dem Aufrufer (Zeile 39) | 3 Formen von `Select` × 5 Entities × 3 Richtungen = 9 Fälle, 4 leere Ergebnisse, `AlleAbfrageUrsachen` | Alias nur an der vierten Entity (`N-M8-Entity4`); `V12`, `S1-unit`, `S2-unit` rot |

Raten der Zeile 32 nach der Fix-Runde (`TestAbfrageGleichzeitig` schreibt jetzt in jedes Ergebnis, hängt an, setzt den ersten Eintrag von `AlleAbfrageUrsachen`, lässt den Zeitplaner laufen und liest zurück; 20 Prozesse je Arm, `-test.count=1`, Detektor und Wirkung): leeres Ergebnis aus einer Paketvariablen (`S1-conc`) 20 von 20 mit Detektor, 20 von 20 ohne; `AlleAbfrageUrsachen` aus einer Paketvariablen (`S2-conc`) 20 von 20 und 20 von 20; gemeinsamer Puffer der Projektion (`S3-conc`) 20 von 20 und 20 von 20; Cache als `map` (`S4-Cache-Map`) 20 von 20 und 20 von 20. Vor dem Zurücklesen waren `S1` und `S3` ohne Detektor in 0 von 20 rot und `S2` in keinem Arm.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-edm-ausdruecke` liegt in `done/` (eingetreten, Beleg [slice-edm-ausdruecke](../done/slice-edm-ausdruecke.md)); der Start als Teil 1 von `welle-mock` folgt dem Trigger aus deren §2 (Linie B). Zusammenhang mit der Regel „höchstens eine Linie von Teil 1 zugleich“ (`welle-mock` §4): `slice-odata-json-schreiben` aus Teil 1 läuft bereits in einem eigenen Worktree; die Regel bleibt im Wortlaut der Welle stehen und gilt hier als Hinweis, denn der User hat höchstens drei Worktrees zugleich entschieden. Beide Slices brauchen nur das Modell (der Interpreter zusätzlich `slice-edm-ausdruecke`, in `done/`) und teilen keinen Code. Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-190`, `SPEC-191`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke aus dem Block unten (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): eine Randform des Rasters in §6 ohne Ort in der Spezifikation, die der Implementer findet (`AGENTS.md` §3.12); er hält an und gibt sie an den Architect zurück. Der Ort der Entscheidung über negative und nicht ganzzahlige Werte von `$top` und `$skip` ist entschieden (Parser, `SPEC-162` Nr. 4), kein Grund mehr.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): der Block unten. Der Architect liefert `pfad`, `produktcode` (550: `abfrage.go` rund 190, `abfrage_filter.go` rund 120, `abfrage_ordnung.go` rund 130, `abfrage_fehler.go` rund 110), `fakes` (0: die Maske ist eine Funktion, kein Port) und `schnitt-grund`; `basis` (`055c295`, der Commit des Architect, der Vertrag, Tabelle und Raster abschließt), `schaetzung` (3275) und `marke` (4500) hat der Planner in eigenem Commit vor dem ersten Code-Commit im Block gesetzt (im Slice-Zweig, damit `make groessenmarke` auf `main` bis zum Merge grün bleibt); eine ältere Marke gibt es nicht, ein Eintrag `ersetzt` entfällt. Die Tabelle in §3 hat 34 Zeilen, davon 1 nebenläufig (Zeile 32) und 33 andere, kein Port-Fake; der Interpreter ist kein Leser fremder Bytes (er arbeitet auf dem Modell). Schätzung nach `SPEC-128` Nr. 1: *S* = 85 · 1 + 80 · 33 + 200 · 0 + 550 = 3275, Marke = 100 · ⌈135 · 3275 / 10000⌉ = 4500; die Schnittfrage stellt sich (*S* ≥ 2223) und ist im Block beantwortet; der Planner bestätigt „kein Schnitt“: der größte abtrennbare Teil hat *S* = 580, der Rest 2695 liegt weiter über 2223, und die Vergleichswerte enden unter ihrer Schätzung. Die Formel steht allein in `SPEC-128` (Entscheidung des Users, 2026-11-06; Beleg [welle-record-replay-results](../done/welle-record-replay-results.md)). Vergleichswerte: `slice-edm-ausdruecke` Schätzung 3645, Messung 3229 (0,89); `slice-odata-url-filter` Schätzung 3460, Messung 2388 (0,69). **Messung nach der Fix-Runde** (`make groessenmarke`, Summe aus Einfügungen und Löschungen von `git diff --numstat 055c295..HEAD` über die Pfade des Blocks): 3070 von Marke 4500 (Schätzung 3275, Verhältnis 0,94); Produktcode 491 Zeilen gegen 550 geschätzt, Tests 2579; die Tabelle in §3 hat nach dem Review 39 Zeilen statt der 34, mit denen der Planner die Marke setzte, die Marke bleibt. Beim ersten Handoff: 2588. Keine `ueberschreitung`. Die Anzahl der Abfragen von `TestAbfrageGleichzeitig` (40 statt 12) und die vollständigen Positionsachsen der Zeilen 14 und 22 haben die Tests über die Formel hinaus vergrößert, ohne die Marke zu berühren.

```groessenmarke
pfad: src/hexagon/model/abfrage.go
pfad: src/hexagon/model/abfrage_test.go
pfad: src/hexagon/model/abfrage_filter.go
pfad: src/hexagon/model/abfrage_filter_test.go
pfad: src/hexagon/model/abfrage_ordnung.go
pfad: src/hexagon/model/abfrage_ordnung_test.go
pfad: src/hexagon/model/abfrage_fehler.go
pfad: src/hexagon/model/abfrage_fehler_test.go
pfad: src/hexagon/ports/outbound/abbildung_test.go
pfad: tests/integration/abfrage_url_test.go
produktcode: 550
fakes: 0
schnitt-grund: Kein Schnitt vor dem Code. Der Vertrag liefert eine Funktion, deren Stufen (Filter und Zählung, Ordnung und Paging, Projektion) denselben Fehlerwert, dieselbe Rangfolge über alle Optionen (Zeile 25: 110 Paare über Filter, $orderby, $select und Entities), dieselbe Eingabeprüfung und denselben Zugriff auf die Werte der Entity teilen; der Lieferwert (Count vor Paging, Gleichstand nach dem Schlüssel) entsteht erst im Zusammenspiel. Der größte Teil ohne geteilten Mechanismus, $orderby mit Guid, Namen und gemischten Schlüsseln (Zeilen 15 bis 17, 20, 22: 5 Zeilen und rund 180 Zeilen Produktcode, S = 5 · 80 + 180 = 580), liegt weit unter der Schwelle: der Rest hätte S = 2695 und stünde weiter über 2223. Ein zweiter Slice verdoppelte Review und Verifikation ohne neuen Lieferwert. Vergleichswerte: slice-edm-ausdruecke (S = 3645) endete bei 3229, slice-odata-url-filter (S = 3460) bei 2388, beide ohne Schnitt.
basis: 055c295
schaetzung: 3275
marke: 4500
```

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7): Geprüft vom Architect: der Slice berührt keine Datei der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel) und braucht keine Änderung von `.a-check.yml`: `FuehreAus` liegt im vorhandenen Paket `src/hexagon/model/`, und `tests/integration/` darf Modell und Driven Adapter nennen (Kanten `tests` → `model`, `tests` → `adapters-driven`, [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)). Muss der Implementer eine ändern, hält er an und gibt es als Rückgabe an den Planner zurück (§6, Ausgang *offen*, `AGENTS.md` §3.13); der Slice ändert sie nicht selbst.

**Kennungen** (Teilung der Linie B): dieser Slice vergibt aus dem hinteren Teil des Bereichs: `SPEC-190` bis `SPEC-199` (vergeben: `SPEC-190`, `SPEC-191`), `ARC-035` bis `ARC-039` (keine nötig: kein neues Paket, keine neue Komponente), ADR `0025` bis `0029` (keine nötig: handgeschrieben über dem Modell, keine Bibliothek, keine Abwägung mit Folgen über den Slice hinaus), `CO-030` bis `CO-039` (keine). `slice-odata-json-schreiben` behält `SPEC-176` bis `SPEC-189`, `ARC-030` bis `ARC-034`, ADR `0020` bis `0024`, `CO-020` bis `CO-029`. Ein erschöpfter Teil geht als Frage an den Planner.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Negative oder nicht ganzzahlige Werte von `$top` und `$skip` (`400`): der Ort der Entscheidung — **Ausgang:** entfallen als Lücke: entschieden im Parser (`SPEC-162` Nr. 4, geliefert; `SPEC-160` Nr. 2: `Wert` nie negativ); der Interpreter behandelt einen negativen `Wert` mit `Gesetzt` wahr als Programmfehler (`SPEC-190` Nr. 3 a, Fänger Zeilen 5, 6, 34). Keine Rückgabe an den Planner: `slice-odata-url-pfad` liegt in `done/` und hat die Entscheidung schon getragen.
- Ordnung von `null` im `$orderby` — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-166` Nr. 6 (`null` zuerst aufsteigend, zuletzt absteigend); Fänger Zeile 15. Gemischte Sortierschlüssel — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-191` Nr. 3 a, b (Typ und Richtung je Element, das erste verschiedene entscheidet; Schlüssel als aufsteigender Gleichstand nach den Werten, Nr. 3 d); Fänger Zeilen 16, 17, 18.
- Laufzeit bei 10.000 Entities pro Workspace (`LH-QA-01.a`): dieser Slice sagt Korrektheit zu, nicht Latenz — **Ausgang:** weiter offen; die Messung trägt der Closure-Trigger von `welle-betrieb` (§3 dort, Lasttest); die Prüfgrenze 10.000 und 10.001 steht als benannte Grenze in `SPEC-190` Nr. 7 und als Zeile 33.
- Teilprojektionen: `$select` einer Property, die nur teilweise beobachtet ist (`SPEC-044`) — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-191` Nr. 6 (Filter und Ordnung `422`, `$select` lässt aus); Fänger Zeilen 14, 22, 27 bis 29. **Folge im Betrieb:** ein Eintrag mit einer Teilprojektion lässt jeden Filter auf die fehlende Property mit `422` enden; ob die Collection vollständig ist, entscheidet die Coverage vor dem Aufruf.
- Pflichten des Aufrufers aus dem Vertrag: die Maske aus den bekannten Properties der Baseline (`SPEC-190` Nr. 1 d), `/$count` nur mit `Filter` (Nr. 2), Entities mit den Key-Properties in `Properties` (Nr. 3 f). **Ausgang:** offen; Adresse: `slice-mock-lesen` (Planner geprüft, 2026-11-06): dessen §1 *Ausdrücklich NICHT* (Mutationen, Fallback, Spans, Service-Dokument) schließt die Sendung nicht aus, seine DoD nennt sie nicht. Der Nachtrag (ein DoD-Punkt „Aufruf des Interpreters nach `SPEC-190` Nr. 1 d, 2, 3 f: Maske aus den bekannten Properties der Baseline, `/$count` nur mit `Filter`, Entities mit den Key-Properties; Geber `slice-query-interpreter`“) steht auf `main` aus; ihn trägt der Integrationsschritt dieses Slice oder der Start des Nehmers ein, denn der Slice-Zweig schreibt nur den eigenen Plan.
- Längen-Differenz zwischen Filter und Schlüssel (`SPEC-164` Nr. 4 a): **Ausgang:** entfallen: kein Beschluss, die Regel bleibt (Grund in §1).
- Gate-Kette: **Ausgang:** entfallen als Lücke: keine Änderung nötig (Prüfung in §4).
- Randform ohne Ort (`AGENTS.md` §3.12, §3.13): `typ.Keys` leer bei nichtleerer Eingabe, oder zwei Entities mit gleichen Werten der Key-Properties und verschiedenem `Schluessel`. **Ausgang:** entschieden: `SPEC-191` Nr. 3 e (Text von `Schluessel` bytegleich als letzte Stufe, kein Programmfehler; der Code ordnet so, keine Auflage an den Implementer); Fänger `TestSchluesselRest`, Zeile 35 der Tabelle *Zusage · Fänger* (§3); der Ist-Zustand ist Zusage, es bleibt nichts offen und nichts geht an den Planner zurück.
- Formel der Größenmarke (`AGENTS.md` §3.14 gegen `SPEC-128` Nr. 1 und 2) — **Ausgang:** entschieden: `SPEC-128` gilt, `AGENTS.md` §3.14 trägt keine Zahlen (Entscheidung des Users, 2026-11-06; Beleg [welle-record-replay-results](../done/welle-record-replay-results.md)). Dieser Slice rechnet 3275 und Marke 4500 (§4).

**Zusatztabelle (mehrere Verstöße und Modi; Fänger):**

| Paar oder Modus | Entscheidung | Ort | Fänger |
|---|---|---|---|
| alle 110 geordneten Paare aus den 11 Verstößen (Stufe 1; Fehler des Filters; `$orderby` unbekannt, nicht unterstützt, Guid, nicht beobachtet; `$select` unbekannt, nicht unterstützt, nicht beobachtet; Entity ohne Filter-Property; Entity ohne `$orderby`-Property) | der erste der Stufenfolge, gleich welcher Klasse | `SPEC-190` Nr. 5 | Zeile 25 |
| zwei Verstöße in `$orderby` oder `$select` (zwei Namen) | das erste Element | `SPEC-190` Nr. 5 Stufen 3 und 4 | Zeile 25 |
| derselbe Name: `Edm.Guid` mit Maske falsch; Typ außerhalb der acht mit Maske falsch | `ordnung_nicht_unterstuetzt` bzw. `typ_nicht_unterstuetzt` (`501`), nicht `422` | `SPEC-190` Nr. 5, `SPEC-167` Nr. 5 | Zeilen 21, 25 |
| dieselbe Property fehlt in einer Entity und steht in `$filter`, `$orderby` und `$select` | `$filter` zuerst (`422`, Option `$filter`) | `SPEC-190` Nr. 5 Stufe 5 | Zeile 25 |
| Modus `Count`: nicht gesetzt, falsch, wahr | `Anzahl` gleich | `SPEC-190` Nr. 2 | Zeile 2 |
| Modus `Select`: `nil`, `Alle`, Namen, `*` mit Namen | alle bzw. die Namen | `SPEC-191` Nr. 5 | Zeile 27 |
| Modus Eingabe: `nil`, leer, nichtleer | Stufen 1 bis 4 unabhängig davon | `SPEC-190` Nr. 5 | Zeile 26 |
| Modus Filter: `nil` gegen gesetzt | `nil` ruft die Maske nicht | `SPEC-191` Nr. 2 a | Zeile 8 |
| Modus Richtung je Schlüssel: aufsteigend, absteigend | umgekehrt mit `null` zuletzt | `SPEC-191` Nr. 3 b | Zeilen 15, 16 |

Die Allaussagen „der erste Verstoß entscheidet“ (Zeile 25), „kein Teil der Eingabe im Fehler“ (Zeile 23), „die Eingabe bleibt unverändert“ (Zeile 31) und „jede Property des Filters in jeder Entity“ (Zeile 14) haben ihren Fänger über alle Fälle.

**Raster *Operation × Argumentart*** (Ort = Nummer in `SPEC-190` bzw. `SPEC-191`; „gilt nicht“ mit Grund). Die Operation ist `FuehreAus`; das Argument ist das, was sie deutet.

| Argument | fehlend oder nicht gesetzt | leer | unbekannt oder fremd | Zustand, zu groß oder zu viele | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| `typ` | Nullwert: jeder Name ist `eigenschaft_unbekannt`, `SPEC-190` Nr. 4 | ohne Properties wie fehlend, Nr. 4 | Typ eines anderen Sets oder abgeleitet: nicht prüfbar, `SPEC-190` *Ungeprüft*; Navigation und offener Typ: `typ_nicht_unterstuetzt`, Nr. 4 | `Abstract` wird nicht gelesen (gilt nicht: der Typ des Sets genügt, `SPEC-190` Nr. 1 a) | gilt nicht: kein `ctx`, die Funktion hat keine Wartestelle (`SPEC-190` Kopf) | Nr. 6 |
| `entities` | `nil` gleich leer, Nr. 1 b | leere Liste: `Ergebnis{Entities: [], Anzahl: 0}`, Nr. 2; Stufen 1 bis 4 laufen, Nr. 5 | Property, die `typ` nicht kennt: bleibt bei `Alle` erhalten, `SPEC-191` Nr. 5 a; `Schluessel` leer oder doppelt: Nr. 3 e | Property `null`, fehlend, `redigiert`: Nr. 1 b, 3 f, 3 g; `SPEC-191` Nr. 2 b, 3 c, 5 b; 10.000 und 10.001: Nr. 7 | gilt nicht: wie `typ` | Reihenfolge ohne Bedeutung, Nr. 1 b, 6 |
| `opt.Filter` | `nil`: keine Auswahl, `SPEC-191` Nr. 2 a | `Ausdruck` `nil`: Programmfehler, `SPEC-190` Nr. 3 b | Name, Typ, Literal, Operator: `PruefeFilter`, `SPEC-191` Nr. 2 a, `SPEC-167`; fehlend in einer Entity: Nr. 2 b | Wert `null`: `SPEC-166` Nr. 4; Tiefe: `SPEC-167` Nr. 8 | gilt nicht: wie `typ` | Nr. 6 |
| `opt.OrderBy` | `nil`: nur der Schlüssel, `SPEC-191` Nr. 3 d | Liste ohne Element wie `nil`, Nr. 3 d | `SPEC-190` Nr. 4; doppelter Name: Nr. 3 d | Wert `null`: `SPEC-191` Nr. 3 b; 32 und 33 Elemente: `SPEC-190` Nr. 7 | gilt nicht: wie `typ` | Nr. 6 |
| `opt.Top`, `opt.Skip` | `Gesetzt` falsch: `Wert` ungelesen, `SPEC-190` Nr. 3 a, `SPEC-191` Nr. 4 | `Wert` 0: leere Liste bzw. nichts übersprungen, `SPEC-191` Nr. 4 | negativ mit `Gesetzt`: Programmfehler, `SPEC-190` Nr. 3 a (Entscheidung im Parser, `SPEC-162` Nr. 4) | `Skip` ab n: leer; `Top` über dem Rest; 2147483647: `SPEC-191` Nr. 4 | gilt nicht: wie `typ` | Nr. 6 |
| `opt.Select` | `nil`: alle Properties, `SPEC-191` Nr. 5 a | `Alle` falsch ohne Namen: Programmfehler, `SPEC-190` Nr. 3 c | `SPEC-190` Nr. 4; doppelter Name: Nr. 3 c | Property `null`, fehlend: `SPEC-191` Nr. 5 b, 6 | gilt nicht: wie `typ` | Nr. 6 |
| `opt.Count`, `opt.Sonstige` | gelten nicht: werden nicht gelesen, `SPEC-190` Nr. 1 c | gilt nicht: wie fehlend | gilt nicht: wie fehlend | gilt nicht: wie fehlend | gilt nicht: wie `typ` | gilt nicht: ohne Wirkung |
| `beobachtet` | `nil`: alle beobachtet, `SPEC-190` Nr. 1 d, `SPEC-167` Nr. 1 | falsch für jeden Namen: `nicht_beobachtet`, `SPEC-190` Nr. 4 | Name, den der Typ nicht trägt oder der nicht unterstützt ist: wird nicht gerufen, `SPEC-190` Nr. 5, `SPEC-167` Nr. 3 a | gilt nicht: eine Funktion eines Namens hat keinen Zustand | gilt nicht: wie `typ` | einmal je Vorkommen, `SPEC-190` Nr. 5 |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start am 2026-11-06 erneut; der Zähler folgt aus den Dateien unter `../observations/BEO-REPO/<slug>/evidence/`, hier steht keine Zahl. Reine Funktionen mit Allaussagen: `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16). Beim Start zusätzlich einschlägig für Fänger und Raster (verkörpert in `AGENTS.md` §3.17 bzw. §3.14): `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (10.000 Entities, Mengen von Sortierschlüsseln), `BEO-REPO/zustand-des-arguments-als-dimension` (fehlend, `null`, leer) und `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`. Offen: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (Stand: geplant in `slice-guard-schreibwege`); die Betriebsregel „nur `Edit` und `Write`“ gilt für jede Rolle dieses Slice (Übergangsregeln Nr. 8).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

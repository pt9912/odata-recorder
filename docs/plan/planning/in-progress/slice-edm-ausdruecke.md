# Slice slice-edm-ausdruecke: Typregeln der EDM-Typen und Null-Semantik als reine Funktionen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-08`, `LH-QA-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-08.b` (Typtabelle, Null-Semantik) · `LH-FA-08.a` · `LH-QA-02.a` · `SPEC-049` · `SPEC-024` · `SPEC-164` Nr. 10 · `SPEC-166` · `SPEC-167` · `SPEC-109` Nr. 5 (Zeile `ErrFilterNichtBeobachtet`) · `ARC-001`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-edm-ausdruecke, Worktree /Development/odata-recorder-wt-edm-ausdruecke)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Typregeln der acht EDM-Typen und die dreiwertige Null-Semantik liegen als reine Funktionen im Domain-Modell: Vergleich und Sortierung je Typ (ordinal, dezimal ohne Binär-Float, Instant für `Edm.DateTimeOffset`, `SPEC-166`), `and`, `or`, `not` mit `null` als unbekannt, und die Typprüfung des Filter-AST gegen das Modell des Dienstes (Property vorhanden, Operanden verträglich, nicht beobachtete Property ergibt `ODR-E6003`, `SPEC-167`).

**Übernimmt (Sendung von `slice-odata-url-filter`, `SPEC-164` Nr. 10; der Leser prüft sie nicht):** (a) Wurzel des Filters nicht boolesch sowie `not`, `and`, `or` mit nicht booleschen Operanden ergibt `400` (`not Price gt 5` ist ein Fall); (b) Literal gegen Literal und `null` gegen einen Typ (Nr. 10 b); (c) Property eines Typs außerhalb der acht im Filter ergibt `501` mit `ODR-E6001` (Nr. 10 d). Die Einzelregeln stehen in `SPEC-167` Nr. 3 bis 5; der Ausschluss unten zu den Typen außerhalb der acht trifft (c) nicht, er stützt es.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Pipeline aus Filter, Sortierung, Paging und Projektion — `slice-query-interpreter`; hier entstehen nur die Funktionen, die er ruft. Die Sendungen an ihn (Aufruf von `PruefeFilter` mit der Maske, `$orderby` nach Guid, Umkehr der Ordnung bei `desc`) stehen in dessen §1 und in §6 unten.
- Typen außerhalb der acht (`Edm.Binary`, `Edm.Duration`, `Edm.TimeOfDay`, `Edm.Geography*`) — nicht im Muss-Umfang; sie ergeben `501` mit `ODR-E6001`.
- Jeder Zugriff auf Ports oder Dateisystem — das Paket bleibt reine Funktion (`architecture.md` §2, Schicht Domain).

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

- [ ] Vergleich und Sortierung je Typ nach der Tabelle in `LH-FA-08.b`, einschließlich `Edm.Int64` jenseits von 2^53 und `Edm.Decimal` ohne Binär-Float.
- [ ] Null-Semantik: Wahrheitstafeln für `eq`, `ne`, `gt`, `ge`, `lt`, `le`, `and`, `or`, `not` über allen Kombinationen mit `null`.
- [ ] Typprüfung des Filter-AST: unbekannte Property, unverträgliche Operanden `400`, nicht beobachtete Property `ODR-E6003`; Wurzel und Operanden von `not`/`and`/`or` nur boolesch (`400`), Literal gegen Literal und `null` gegen Typ, Property eines Typs außerhalb der acht `501` (Sendung von `slice-odata-url-filter`, `SPEC-164` Nr. 10).
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
| `src/hexagon/model/wert_vergleich.go` | neu | `WertAusLiteral`, `Vergleiche`, `OrdneWerte`, `HatOrdnung`: Klassen, Ordnung je Klasse, Form der kanonischen Texte (`SPEC-166` Nr. 1 bis 4, 6, 7) |
| `src/hexagon/model/wahrheit.go` | neu | `Wahrheit`, `Und`, `Oder`, `Nicht`, `WahrheitVon`, `Durchlaesst` (`SPEC-166` Nr. 5) |
| `src/hexagon/model/filter_typ.go`, `filter_typ_fehler.go` | neu | `PruefeFilter` mit Wurzel, Operanden von `not`/`and`/`or`, Literal gegen Literal, `null` gegen Typ, Typ außerhalb der acht; `FilterTypFehler` und `ErrFilterNichtBeobachtet` (`SPEC-167`, `SPEC-164` Nr. 10) |
| `src/hexagon/ports/outbound/abbildung.go` | update | eine Zeile der Tabelle: `ErrFilterNichtBeobachtet` ergibt `ODR-E6003` (`SPEC-109` Nr. 5) |
| `tests/integration/filter_instant_test.go` | neu | Zeitpunkte mit verschiedenem Versatz aus dem Leser des Filters vergleichen gleich (Zeile 8); `tests/` darf Modell und Driven Adapter nennen ([ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)) |
| `spec/spezifikation.md` | update | Vertrag mit Randformen: `SPEC-166`, `SPEC-167`, Zeile in `SPEC-109`, Zeiger in `LH-FA-08.b` und `SPEC-164` Nr. 10 (Architect) |
| Tests | neu | je Datei der Produktcode `…_test.go` (Paare im Block von §4); Fänger der Tabelle unten |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht hier, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16); beide entstehen mit dem Vertrag in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12), die Randformen entscheidet der Architect, nicht der Implementer. Die Namen der Tests in der Spalte *Fänger* sind Vorschläge; eine Zeile bleibt gleich, wenn der Test anders heißt. Eine Mutation ist eine einzelne Änderung am Code, die der Implementer vor der Übergabe selbst rot sieht (`AGENTS.md` §3.10); eine Schleife läuft über alle Elemente der Menge, die der Wortlaut der Zusage meint (`AGENTS.md` §3.17), und vergleicht gegen eine Erwartung, die der Test selbst trägt (eine Tabelle, `math/big` als Referenz), nicht gegen den Code. Eine Fixture gleicht nie dem Wert der Mutation.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | `Edm.Int64` jenseits von 2^53 ist exakt (`SPEC-166` Nr. 3 c) | `TestZahlExakt`: Schleife über alle Paare von `2^53-1`, `2^53`, `2^53+1`, `2^53+2`, `-(2^53)-1`, `MaxInt64`, `MinInt64` (beide Seiten von 2^53) und alle sechs Operatoren gegen `math/big` | Wert über `float64` lesen: `…993 eq …992` wird wahr | andere |
| 2 | Gleiche Werte vergleichen gleich über `int32`, `int64`, `decimal` (`SPEC-166` Nr. 3 c) | `TestZahlGleichUeberArten`: Schleife über `0`, `1`, `-1`, `5`, `2147483647`, `-2147483648` in allen neun Paaren der drei Arten, je sechs Operatoren | Gleichheit nur bei gleicher Art | andere |
| 3 | `Edm.Decimal` exakt, ohne Binär-Float und ohne Facetten, bis 1024 Zeichen (`SPEC-166` Nr. 3 c, 7) | `TestDecimalExakt`: sortierte Liste von 14 Dezimalzahlen (`0.1`, `0.10000000000000001`, `-1.5`, `-1.25`, `-0.5`, `0`, `9`, `10`, 40 Nachkommastellen, 1024 Zeichen, …), alle Paare gegen `math/big`; 1024 Zeichen gültig, 1025 `ErrUngueltig` | Textvergleich statt Zahlvergleich; `float64` | andere |
| 4 | Text ordinal, ohne Locale, Normalisierung und Faltung (`SPEC-166` Nr. 3 a) | `TestTextOrdinal`: Schleife über `''`, `a`, `ab`, `abc`, `B`, `a` gegen `B`, `é` zusammengesetzt und zerlegt, U+0000, U+FFFF und U+10000 (Codepunkt gegen UTF-16-Reihenfolge), gleiche Texte, alle Paare, sechs Operatoren | `strings.EqualFold`; Reihenfolge nach UTF-16 | andere |
| 5 | `false` vor `true` (`SPEC-166` Nr. 3 b) | `TestWahrheitOrdnung`: alle vier Paare, sechs Operatoren | `true` vor `false` | andere |
| 6 | `Edm.Date` ordnet wie der Kalender, gleiche Daten sind gleich (`SPEC-166` Nr. 3 d) | `TestDatum`: `0001-01-01`, `2024-02-28`, `2024-02-29`, `2024-03-01`, `2025-01-01`, `9999-12-31`, alle Paare, sechs Operatoren | Vergleich nach Monat vor Jahr | andere |
| 7 | `Edm.DateTimeOffset` ordnet nach Instant, Sekunde und Bruchteil (`SPEC-166` Nr. 3 e) | `TestZeitpunktOrdnung`: sortierte Liste mit `…:00Z`, `…:00.000000000001Z`, `…:00.5Z`, `…:00.51Z`, `…:01Z`, Tageswechsel, Jahr `0001` und `9999`, alle Paare gegen eine Referenz mit `math/big` | Textvergleich (`Z` hinter `.`); Bruchteil als ganze Zahl lesen | andere |
| 8 | Zeitpunkte mit verschiedenem Versatz desselben Instants sind gleich (`SPEC-166` Nr. 3 e; Normierung `SPEC-161` Nr. 6 f) | `TestInstantVerschiedenerVersatz` unter `tests/integration/`: `$filter` mit `…T10:00:00+02:00`, `…T08:00:00Z`, `…T07:00:00-01:00`, dazu `2020-01-01T00:30:00+01:00` gegen `2019-12-31T23:30:00Z`, alle Paare über `urlparser.NewFilterLeser`, `WertAusLiteral`, `Vergleiche` | Versatz ignorieren (Text bis `T` vergleichen) | andere |
| 9 | Werte verschiedener Klassen sind nicht vergleichbar, `ArtNull` ausgenommen (`SPEC-166` Nr. 2, 7) | `TestKlassenUnvergleichbar`: Schleife über alle 30 geordneten Paare verschiedener Klassen, sechs Operatoren: `ErrUngueltig`; alle Paare derselben Klasse: kein Fehler; `Date` gegen `DateTimeOffset` ausdrücklich | `Date` und `DateTimeOffset` als eine Klasse; `string` und `guid` als eine | andere |
| 10 | Guid: nur Gleichheit; Ordnungsoperatoren und `OrdneWerte` sind `ErrUngueltig`; `HatOrdnung` falsch (`SPEC-166` Nr. 3 f, 4) | `TestGuid`: gleiche Guids `eq`/`ne`, verschiedene, vier Ordnungsoperatoren (auch gegen `ArtNull`), `OrdneWerte`; `HatOrdnung` über die acht Typen (nur `Edm.Guid` falsch) und `edm.string`, `Edm.Binary`, leer | `gt` auf Guid nach Text ordnen; `HatOrdnung` immer wahr | andere |
| 11 | `null` in den sechs Vergleichen nach der Tabelle von `LH-FA-08.b` (`SPEC-166` Nr. 4) | `TestNullVergleiche`: Schleife über sechs Operatoren, die Paare (null, null), (null, x), (x, null) und acht Vertreter (`''`, `0`, `false`, ein Datum, ein Zeitpunkt, eine Zahl, eine Guid nur für `eq`/`ne`, ein Text); `null eq null` (gleiche verglichene Werte) | `null eq null` falsch; `ge` bei zwei `null` falsch; `'' eq null` wahr | andere |
| 12 | Wahrheitstafel `and` über alle Kombinationen mit `null` (`SPEC-166` Nr. 5) | `TestUndTafel`: Schleife über alle neun Paare aus falsch, wahr, unbekannt in beiden Reihenfolgen und zwei Werte außerhalb der Konstanten, gegen die ausgeschriebene Tafel | `null and true` falsch; `null and false` unbekannt | andere |
| 13 | Wahrheitstafel `or` über alle Kombinationen mit `null` (`SPEC-166` Nr. 5) | `TestOderTafel`: wie Zeile 12 | `null or false` falsch; `null or true` unbekannt | andere |
| 14 | Wahrheitstafel `not`, `not null` ist unbekannt (`SPEC-166` Nr. 5) | `TestNichtTafel`: alle drei Werte und zwei außerhalb der Konstanten | `not null` wahr | andere |
| 15 | `WahrheitVon` und `Durchlaesst`: nur wahr lässt durch (`SPEC-166` Nr. 5) | `TestWahrheitVonUndDurchlaesst`: Schleife über alle zehn Arten, `true`, `false` und `null`; `Durchlaesst` über die drei Konstanten und einen Wert außerhalb | `Durchlaesst` auch für unbekannt | andere |
| 16 | `OrdneWerte` stimmt mit `Vergleiche` überein, `null` kommt zuerst (`SPEC-166` Nr. 6) | `TestOrdneWerteGleichVergleiche`: je Klasse eine aufsteigende Liste; alle Paare: Vorzeichen gleich `lt`/`eq`/`gt`; `null` gegen jeden Vertreter -1, jeder gegen `null` 1, `null` gegen `null` 0 | `null` zuletzt; `null` gegen `null` ungleich | andere |
| 17 | Die Ordnung je Klasse ist total (`SPEC-166` Nr. 6) | `TestOrdnungTotal`: Schleife über alle Tripel der Listen aus Zeile 16 und 7: Antisymmetrie und Transitivität | Zeitpunkte als Text ordnen (nicht transitiv zu den Bruchteilen) | andere |
| 18 | `WertAusLiteral` bildet jede der sechs Arten ab (`SPEC-166` Nr. 1) | `TestWertAusLiteral`: eine Zeile je Art; Zeichenketten `'a'`, `''`, `''''`, `'O''Hara'`, `'a''''b'`; Verstöße `Literal{}`, Art 7, `abc`, `'a'b'`, `'` ohne Gegenstück | Anführungszeichen bleiben im Text; `''` bleibt doppelt | andere |
| 19 | Eine Form, die kein Leser liefert, ist `ErrUngueltig`, nie ein Ergebnis (`SPEC-166` Nr. 7) | `TestFormverstoesse`: Schleife über `boolean`, `int32`, `int64`, `decimal`, `guid`, `date`, `datetimeoffset` mit je mindestens drei verfehlten Formen (`True`, `1.0`, `01`, `-0`, `+1`, `1e3`, leer, Großbuchstaben, ohne `-`, `2020-1-1`, `…t…z`, Bruchteil mit Null am Ende, 13 Ziffern), `ArtRedigiert`, unbekannte Art, Operator `''`, `EQ`, `like`; `Error()` enthält den Wert nicht | Fehler als `WahrheitFalsch` zurückgeben; Form nicht prüfen | andere |
| 20 | Gleiche Eingabe, gleiche Ausgabe; gleichzeitige Aufrufe beeinflussen sich nicht (`SPEC-166` Nr. 8, `SPEC-167` Nr. 8) | `TestGleichzeitig`: 8 Goroutinen mit je 1000 Aufrufen von `Vergleiche`, `OrdneWerte`, `Und`, `Oder`, `Nicht` und `PruefeFilter` auf einem geteilten Typ und Baum, jedes Ergebnis gleich dem Ergebnis des sequentiellen Laufs; Lauf unter `make test-race` | Cache in einer `map` des Pakets; geteilter Puffer für den Zahlvergleich | nebenläufig |
| 21 | Die Wurzel ist boolesch oder `null`, sonst `400` (`SPEC-167` Nr. 3 d) | `TestWurzel`: Schleife über 7 gültige (Vergleich, `and`, `or`, `not`, `Edm.Boolean`-Property, `true`, `null`) und alle ungültigen Wurzeln (die sieben anderen Property-Typen, fünf andere Literalarten): `nicht_boolesch` | Wurzel nicht geprüft; `null` als Wurzel abgewiesen | andere |
| 22 | `not`, `and`, `or` haben boolesche Operanden an jeder Stelle (`SPEC-167` Nr. 3 b) | `TestBooleschOperand`: Schleife über die fünf Stellen (Operand von `not`, `Links` und `Rechts` von `and` und `or`) und alle zwölf nicht booleschen Operanden (sieben Property-Typen, fünf Literalarten); `not Price gt 5` als Baum aus dem Leser | `Rechts` nicht geprüft; `not` ohne Prüfung | andere |
| 23 | Boolesche Operanden und `null` sind gültig (`SPEC-167` Nr. 3 b, 4 a) | `TestBooleschGueltig`: dieselben fünf Stellen mit `Edm.Boolean`-Property, `true`, `false`, Vergleich, `and`, `or`, `not`, `null` | `null` als Operand abgewiesen | andere |
| 24 | Eine unbekannte Property ist `400`, bytegleich gesucht (`SPEC-167` Nr. 3 a) | `TestEigenschaftUnbekannt`: Namen `Foo`, `id` gegen `Id`, `ID`, ` Id`, `Id `, leer, an jeder der sieben Stellen (Wurzel, Operand von `not`, `Links`/`Rechts` von Vergleich und von `and`/`or` mit gültigem Gegenüber) | Suche ohne Beachtung der Schreibweise | andere |
| 25 | Navigation-Property und jeder Name eines offenen Typs sind `501`, im geschlossenen Typ ein unbekannter Name `400` (`SPEC-167` Nr. 3 a, 5) | `TestNavigationUndOffen`: Navigation-Name an den sieben Stellen; offener Typ mit unbekanntem Namen; geschlossener Typ mit demselben Namen | `Open` ignoriert; Navigation als unbekannt | andere |
| 26 | Eine Property außerhalb der acht ist `501`, an jeder Stelle (`SPEC-167` Nr. 5) | `TestTypAusserhalb`: Schleife über 29 Typnamen (`Edm.Binary`, `Duration`, `TimeOfDay`, `Stream`, `Double`, `Single`, `Int16`, `Byte`, `SByte`, `Untyped`, die 16 räumlichen, `Collection(Edm.String)`, ein komplexer Typ, ein Aufzählungstyp) an den sieben Stellen, dazu `P eq null` und eine unbeobachtete Property dieser Typen (bleibt `501`) | Nur `Edm.Binary` geprüft; Maske vor dem Typ | andere |
| 27 | Verträglichkeit der Operanden nach der Klassenrelation (`SPEC-167` Nr. 2, 4 c) | `TestVertraeglichkeit`: Schleife über alle 225 geordneten Paare aus 8 Property-Typen, 6 Literalarten und `null` (Operator `eq`) gegen eine Tabelle im Test; unverträglich ergibt `typ_unvertraeglich`, auch `Id eq 'x'` und Datum gegen Zeitpunkt | `Date` mit `DateTimeOffset` verträglich; Zeichenkette mit Guid verträglich | andere |
| 28 | `null` gegen jeden Typ, auch gegen nicht `Nullable` (`SPEC-167` Nr. 4 a) | `TestNullGegenTyp`: `null` gegen die acht Property-Typen, beide Seiten, sechs Operatoren (Guid mit Ordnungsoperator: `operator_nicht_unterstuetzt`); `null eq null`; eine Property mit `Nullable` falsch | `null` nur gegen `Nullable` gültig | andere |
| 29 | Literal gegen Literal ist ein Vergleich mit denselben Regeln, ohne Bereichsprüfung (`SPEC-167` Nr. 4 b) | `TestLiteralGegenLiteral`: Schleife über 6 x 6 Literalarten (gültig genau bei gleicher Klasse), `1 eq 1`, `99999999999999999999 eq 1`, GUID mit `gt`: `operator_nicht_unterstuetzt` | Bereichsprüfung auch für Literale | andere |
| 30 | Bereich von `Edm.Int32` und `Edm.Int64` bei Literalen, beide Seiten der Grenze (`SPEC-167` Nr. 3 c) | `TestBereich`: je Typ `Max`, `Max+1`, `Min`, `Min-1`, `1.5`, `1.0`, `0`, `1e3`, leerer Text, 25 Ziffern; die Literale auf beiden Seiten des Vergleichs, sechs Operatoren; `Edm.Decimal`-Property ohne Grenze; `Edm.Int32` gegen `Edm.Int64`-Property ohne Literal gültig | Grenze `<=` statt `<`; nur `Rechts` geprüft; `Edm.Int64` mit der Grenze von `Edm.Int32` | andere |
| 31 | `gt`, `ge`, `lt`, `le` auf Guid sind `501`, `eq` und `ne` gültig (`SPEC-167` Nr. 3 c) | `TestGuidOperatoren`: vier Ordnungsoperatoren mal Property und GUID-Literal mal beide Seiten mal Gegenüber Guid und `null`; `eq`, `ne` gültig; `Guid gt 5` ist `typ_unvertraeglich` | Prüfung nur für Property, nicht für Literal | andere |
| 32 | Eine nicht beobachtete Property ist `ODR-E6003` an jeder Stelle, `nil` heißt alle beobachtet (`SPEC-167` Nr. 3 a) | `TestNichtBeobachtet`: sieben Stellen mal Maske `nil`, leer, nur diese Property; `P eq null` mit unbeobachteter `P` ist `nicht_beobachtet`, nicht wahr oder falsch; Literal und `null` fragen die Maske nicht | Maske nur an der Wurzel; `nil` als „nichts beobachtet“ | andere |
| 33 | Die Maske wird nur für Properties mit unterstütztem Typ gerufen, einmal je Vorkommen, nach dem ersten Verstoß nicht mehr (`SPEC-167` Nr. 3 a) | `TestMaskeAufrufe`: Fake hält jeden Namen und zählt die Aufrufe; Bäume mit Namen, Navigation, unbekanntem Namen, nicht unterstütztem Typ, doppeltem Vorkommen, Verstoß in der Mitte | Maske auch für unbekannte Namen; mehrfach je Vorkommen | andere |
| 34 | Der erste Verstoß in der Lesefolge entscheidet, gleich welcher Klasse (`SPEC-167` Nr. 3) | `TestReihenfolgeVerstoesse`: alle 49 geordneten Paare der sieben Ursachen als `Links` und `Rechts` einer Verknüpfung; Kind vor Regel des Knotens; Anforderung an `Links` vor dem Teilbaum von `Rechts`; Wurzel nach dem Baum; `Guid gt 5` (zwei Verstöße am selben Knoten) | `Rechts` vor `Links`; `400` vor `501`; Regel des Knotens vor den Kindern | andere |
| 35 | Ursache, Klasse, Code und Text des Fehlers; kein Teil der Eingabe im Text (`SPEC-167` Nr. 6) | `TestFilterTypFehler`: Schleife über sieben Ursachen: `errors.Is` trifft genau die Sentinel der Klasse (und keine der beiden anderen), `Code()`, `Error()`; Name `GEHEIM` und Literal `'GEHEIM'` stehen nicht im Text | Name im Text; `Is` auch für die falsche Klasse | andere |
| 36 | `ErrFilterNichtBeobachtet` ergibt `ODR-E6003`, die übrigen Ursachen `ODR-E6005` und `ODR-E6001` über die Tabelle (`SPEC-109` Nr. 5, `SPEC-167` Nr. 6) | Test der Abbildung in `abbildung_test.go`: Schleife über die sieben Ursachen, auch verpackt und in `errors.Join`; die Namen der Tabelle gegen die `Err…` des Pakets (`EintraegeAbbildung`) | Zeile fehlt (Fall `ODR-E9001`) | andere |
| 37 | Ein Programmfehler ist `ErrUngueltig`, nie `FilterTypFehler` (`SPEC-167` Nr. 7) | `TestProgrammfehler`: Schleife über `nil`, Kind `nil` an jeder Stelle (`Links`, `Rechts`, `Operand`), Knoten als Zeiger, fremder Typ des Knotens, Operator `''`, `EQ`, `Literal{}` und Art 7; je mit gültigem Rest | Zeiger als Knoten angenommen; `nil` als gültig | andere |
| 38 | Jede Tiefe wird bewältigt, der Baum bleibt unverändert (`SPEC-167` Nr. 1, 8) | `TestTiefeUndUnveraendert`: Kette `a or a or …` von 2000 Ebenen, 32 verschachtelte Gruppen, 32 `not`; `reflect.DeepEqual` des Baums vor und nach dem Aufruf, auch bei einem Verstoß | Tiefe auf 32 begrenzt; Kind im Baum ändern | andere |

**Raten der nebenläufigen Zeile 20:** der Implementer trägt hier je Mutation Rate, Zahl der Prozesse und Messenden ein (`AGENTS.md` §3.10). Das Fehlen der Sperre oder eines geteilten Puffers fängt der Detektor von `make test-race` je Prozess; eine Mutation, die nur das Ergebnis verfälscht (ein Cache, der den ersten Wert hält), fängt der Vergleich mit dem sequentiellen Lauf.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-odata-url-filter` liegt in `done/` (der Filter-AST ist die Eingabe). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `SPEC-128` Nr. 1 mit `make groessenmarke` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks unten, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): eine Randform des Rasters in §6 ohne Ort in der Spezifikation, die der Implementer findet (`AGENTS.md` §3.12); er hält an und gibt sie an den Architect zurück.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): der Block unten. Der Architect liefert `pfad`, `produktcode` (600: `wert_vergleich.go` rund 230, `wahrheit.go` rund 60, `filter_typ.go` rund 230, `filter_typ_fehler.go` rund 80), `fakes` (0: die Maske ist eine Funktion, kein Port) und `schnitt-grund`; `basis`, `schaetzung`, `marke` setzt der Planner mit `make groessenmarke` in eigenem Commit vor dem ersten Code-Commit (im Slice-Zweig, damit `make groessenmarke` auf `main` bis zum Merge grün bleibt). Die Tabelle in §3 hat 38 Zeilen, davon 1 nebenläufig (Zeile 20) und 37 andere, kein Port-Fake, kein Leser fremder Bytes. Schätzung nach `SPEC-128` Nr. 1: *S* = 85 · 1 + 80 · 37 + 200 · 0 + 600 = 3645, Marke = 100 · ⌈135 · 3645 / 10000⌉ = 5000; die Schnittfrage stellt sich (*S* ≥ 2223) und ist im Block beantwortet. Nach dem Wortlaut von `AGENTS.md` §3.14 (40 je andere Zeile, 85 je nebenläufige, kein fester Anteil für einen Nicht-Leser): 85 + 40 · 37 = 1565, × 1,25 = 1957, Marke 2000 und keine Schnittfrage (mit dem Produktcode als festem Anteil: 2165, × 1,25 = 2707, Marke 2800, ebenfalls unter 3000). Beide Wortlaute widersprechen sich; das Gate rechnet nach `SPEC-128` (Rang 2 vor Rang 8), der Block gilt danach. Die Zeile in §6 trägt den Widerspruch an den User.

```groessenmarke
pfad: src/hexagon/model/wert_vergleich.go
pfad: src/hexagon/model/wert_vergleich_test.go
pfad: src/hexagon/model/wahrheit.go
pfad: src/hexagon/model/wahrheit_test.go
pfad: src/hexagon/model/filter_typ.go
pfad: src/hexagon/model/filter_typ_test.go
pfad: src/hexagon/model/filter_typ_fehler.go
pfad: src/hexagon/model/filter_typ_fehler_test.go
pfad: src/hexagon/ports/outbound/abbildung.go
pfad: src/hexagon/ports/outbound/abbildung_test.go
pfad: tests/integration/filter_instant_test.go
produktcode: 600
fakes: 0
schnitt-grund: Kein Schnitt vor dem Code. Der Vertrag zerfällt in zwei Teile ohne Datenfluss (Werte und Logik, SPEC-166, Zeilen 1 bis 20 mit S = 85 + 80 · 19 + 300 = 1905; Typprüfung, SPEC-167, Zeilen 21 bis 38 mit S = 80 · 18 + 300 = 1740), aber beide teilen die Klassenrelation der Arten und der EDM-Typen (SPEC-166 Nr. 2 trägt SPEC-167 Nr. 2) und die Grammatik der kanonischen Zahl (Wert und Bereichsprüfung des Literals); ein Schnitt verdoppelte sie oder ordnete die Teile. Beide Teile sind Domain ohne Port, ohne Fake, ohne Änderung des Binary, der Interpreter braucht beide, und ein zweiter Slice verdoppelte Review und Verifikation für keinen sichtbaren Lieferwert. Der Vergleichswert slice-odata-url-filter (S = 3460, Marke 4700) endete bei 2388, Auslastung 0,51, ohne Schnitt.
```

**Messpunkte** (`SPEC-128` Nr. 9): beim Handoff und bei der Closure trägt der Planner hier die Messpunkt-Zeile des Programms ein.

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

- Scale und Precision von `Edm.Decimal`: `LH-FA-08.b` sagt „validieren“ ohne Wortlaut der Prüfregel — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-166` Nr. 3 c (keine Facetten, Form der Zahl und 1024 Zeichen; Alternative und Grund dort); Fänger Zeilen 3, 19.
- Gleichheit von `Edm.DateTimeOffset` mit verschiedenen Offsets desselben Instants und Vergleich von `Edm.Date` — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-166` Nr. 2, 3 d, 3 e (Date und DateTimeOffset nicht vergleichbar; Instant, Bruchteil als Dezimalbruch); Fänger Zeilen 6 bis 9 mit gleichen verglichenen Werten.
- Die Allaussage der Null-Semantik („alle Kombinationen“) — **Ausgang:** entfallen als Lücke: Wahrheitstafeln als Schleife über alle Kombinationen, eine Zeile je Tafel (`SPEC-166` Nr. 4, 5); Fänger Zeilen 11 bis 15.
- Bereichsgrenzen von `Edm.Int32` und `Edm.Int64` bei Literalen außerhalb des Bereichs — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-167` Nr. 3 c (`literal_bereich`, `400`; `SPEC-164` Nr. 10 b); Fänger Zeile 30 mit beiden Seiten jeder Grenze.
- Randformen der Sendung von `slice-odata-url-filter` (`SPEC-164` Nr. 10): nicht boolesche Wurzel, `not`/`and`/`or` mit nicht booleschem Operanden (je Stelle des Operanden), Literal gegen Literal, `null` gegen Typ, Property eines Typs außerhalb der acht — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-167` Nr. 3 b, 3 d, 4, 5; Fänger Zeilen 21 bis 31.
- Weitere Entscheidungen des Architect (ohne eigenes Risiko), jede mit Alternative und Grund in der Spezifikation: `null` zuerst in der Sortierung (`SPEC-166` Nr. 6); Guid ohne Ordnung (Nr. 3 f); `null` ist mit jedem Typ verträglich, auch als Wurzel (`SPEC-167` Nr. 4 a); Literal gegen Literal ohne Bereichsprüfung (Nr. 4 b); ein Literal fremder Klasse ist `400`, kein stilles `false` (Nr. 4 c); der erste Verstoß in der Lesefolge entscheidet, kein Fehler trägt Name oder Stelle (Nr. 3, 6); Navigation-Property und jeder Name eines offenen Typs sind `501` (Nr. 5); eine nicht unterstützte, nicht beobachtete Property ist `501`, nicht `422` (Nr. 5).
- **Sendung an `slice-query-interpreter`** (Geber: dieser Slice, `AGENTS.md` §3.13). Gelesen: §1 *Ausdrücklich NICHT* des Nehmers (Coverage-Prüfung, Lesen und Serialisierung, `$expand`) trifft die Sendung nicht, seine DoD („Sortierung und Paging“, „Projektion mit Property-Maske“) trägt sie. Zuzuweisen sind: (a) der Interpreter ruft `model.PruefeFilter` mit der Property-Maske vor der Auswertung und bildet den Fehler über die Abbildung ab (`SPEC-167` Nr. 6); (b) `$orderby` nach einer Property, für die `model.HatOrdnung` falsch ist, ist `501` mit `ODR-E6001` (`SPEC-166` Nr. 3 f); (c) `desc` kehrt die Ordnung von `OrdneWerte` um, `null` steht dann zuletzt (`SPEC-166` Nr. 6). Der Nehmer führt die Sendung in seinem §1 mit der Kennung dieses Slice (im selben Commit eingetragen); sein Risiko „Ordnung von `null` im `$orderby`“ ist durch `SPEC-166` Nr. 6 entschieden. **Ausgang:** weiter offen (Nehmer `slice-query-interpreter`).
- **Wortlaut von `AGENTS.md` §3.14 widerspricht `SPEC-128` Nr. 1 und 2** (40 gegen 80 je andere Zeile, 1,25 gegen 1,35, fester Anteil nur für Leser fremder Bytes gegen immer). Für diesen Slice gilt `SPEC-128`, weil das Gate danach rechnet; die Zahlen beider Wortlaute stehen in §4 (3645 und Marke 5000 gegen 1565 und Marke 2000). Der Architect setzt nichts um. **Ausgang:** offen; Adresse: User, danach Planner (Angleichung von `AGENTS.md` §3.14 an `SPEC-128` oder umgekehrt).

**Zusatztabelle (mehrere Verstöße; Fänger):** die Paare am selben Baum, die zugleich vorliegen können.

| Paar | Entscheidung | Ort | Fänger |
|---|---|---|---|
| zwei Verstöße in `Links` und `Rechts` derselben Verknüpfung oder desselben Vergleichs (alle 49 geordneten Paare der sieben Ursachen) | der erste in der Lesefolge, gleich welcher Klasse | `SPEC-167` Nr. 3 | Zeile 34 |
| Verstoß in einem Kind gegen die Regel des Knotens (`nicht_boolesch` des Operanden, `typ_unvertraeglich`) | das Kind zuerst | `SPEC-167` Nr. 3 | Zeile 34 |
| Anforderung an `Links` (`nicht_boolesch`) gegen Verstoß im Teilbaum von `Rechts` | `Links` zuerst | `SPEC-167` Nr. 3 b | Zeile 34 |
| `typ_unvertraeglich` gegen `operator_nicht_unterstuetzt` am selben Vergleich (`Guid gt 5`) | `typ_unvertraeglich` | `SPEC-167` Nr. 3 c | Zeile 34 |
| nicht beobachtet gegen nicht unterstützter Typ an derselben Property | `typ_nicht_unterstuetzt` (`501`) | `SPEC-167` Nr. 5 | Zeile 26 |
| nicht boolesche Wurzel gegen Verstoß im Baum | der Baum zuerst | `SPEC-167` Nr. 3 d | Zeile 34 |
| Modus der Maske: `nil` gegen gesetzt, an jeder Stelle | `nil` heißt alle beobachtet | `SPEC-167` Nr. 1, 3 a | Zeile 32 |

Die Allaussagen „der erste Verstoß entscheidet“ (Zeile 34) und „kein Teil der Eingabe im Fehler“ (Zeile 35) haben ihren Fänger über alle Fälle.

**Raster *Operation × Argumentart*** (Ort = Nummer in `SPEC-166` bzw. `SPEC-167`; „gilt nicht“ mit Grund). Operationen sind die Stellen, an denen eine Funktion etwas deutet.

| Operation | fehlend | leer | unbekannt oder fremd | zu lang, zu groß oder zu tief | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| R1 `Vergleiche(op, links, rechts)` | `Wert{}` hat eine leere Art: unbekannte Art, `SPEC-166` Nr. 7 | leerer Text: `ArtString` gültig (Nr. 3 a), jede andere Art Formverstoß (Nr. 7), `ArtNull` liest ihn nicht (Nr. 7) | andere Klasse Nr. 2 und 7; unbekannter Operator, `ArtRedigiert`, unbekannte Art Nr. 7; Ordnungsoperator auf Guid Nr. 4 | Zahl über 1024 Zeichen Nr. 3 c und 7; Text sonst unbegrenzt (Aufwand linear, Nr. 7) | gilt nicht: kein `ctx`, reine Funktion ohne Wartestelle, Nr. 7 | Nr. 8 |
| R2 `OrdneWerte(links, rechts)` | wie R1 | wie R1 | wie R1; Guid Nr. 3 f und 7; `ArtNull` Nr. 6 | wie R1 | wie R1 | Nr. 8 |
| R3 `WertAusLiteral(l)` | `Literal{}` hat die Art 0: Nr. 1 (`ErrUngueltig`) | `''` ist der leere Text, Nr. 1; Text einer anderen Art bleibt unverändert, die Form prüft erst `Vergleiche`, Nr. 1 und 7 | Art außerhalb der sechs, Zeichenkette ohne die äußeren `'` oder mit einzelnem `'`: Nr. 1 | gilt nicht: der Leser begrenzt das Literal auf 512 Bytes (`SPEC-164` Nr. 7), die Funktion liest den Text einmal | gilt nicht: kein `ctx`, Nr. 7 | Nr. 8 |
| R4 `WahrheitVon`, `Und`, `Oder`, `Nicht`, `Durchlaesst` | Nullwert von `Wahrheit` ist unbekannt, Nr. 5 | gilt nicht: kein Text; `WahrheitVon` einer leeren Art siehe R1 | Zahlenwert außerhalb der drei Konstanten gilt als unbekannt, Nr. 5; `WahrheitVon` einer anderen Art als `boolean` und `null`: `ErrUngueltig`, Nr. 5 | gilt nicht: feste Größe | gilt nicht: wie R1 | Nr. 8 |
| R5 `PruefeFilter`, Wurzel und Knoten | `ausdruck` `nil`, Kind `nil`: `ErrUngueltig`, `SPEC-167` Nr. 7 | Nullwert eines Knotens: `FilterEigenschaft{}` ergibt `eigenschaft_unbekannt` (Nr. 3 a), `FilterVergleich{}` hat Kinder `nil`: Nr. 7 | Knoten als Zeiger oder fremder Typ, unbekannter Operator: Nr. 7; nicht boolesche Wurzel und Operanden: Nr. 3 b und 3 d | Tiefe: Nr. 8 (keine Grenze, Ketten bis rund 1700 Ebenen) | gilt nicht: kein `ctx`, Nr. 8 | Nr. 8 |
| R6 `PruefeFilter`, `FilterEigenschaft` und Maske | `typ` ohne Properties: jeder Name ist unbekannt (Nr. 3 a), `beobachtet` `nil`: alle beobachtet (Nr. 1) | Name leer: `eigenschaft_unbekannt` (Nr. 3 a) | unbekannter Name, Schreibweise, Navigation, offener Typ, Typ außerhalb der acht, nicht beobachtet: Nr. 3 a und 5 | gilt nicht: der Leser begrenzt den Namen auf 128 Zeichen (`SPEC-164` Nr. 7), die Suche ist bytegleich ohne Grenze (Nr. 1) | gilt nicht: wie R5 | Nr. 8 |
| R7 `PruefeFilter`, Literal und `null` im Vergleich | `FilterLiteral{}` hat die Art 0: Nr. 7 | `''` ist ein Text (Nr. 4 c, Klasse Text); Zahl-Literal mit leerem Text ist `literal_bereich` bei `Edm.Int32` und `Edm.Int64` (Nr. 3 c), sonst ohne Wirkung | Literal fremder Klasse: Nr. 4 c; `null` gegen alles: Nr. 4 a; Guid mit Ordnungsoperator: Nr. 3 c | Bereich von `Edm.Int32` und `Edm.Int64`: Nr. 3 c; zu `Edm.Decimal` und Literal gegen Literal keine Grenze: Nr. 3 c und 4 b | gilt nicht: wie R5 | Nr. 8 |
| R8 `HatOrdnung(edmType)` | gilt nicht: ein Go-String hat keinen fehlenden Wert | leer: falsch, `SPEC-166` Kopf | Name außerhalb der acht, falsche Schreibweise (`edm.string`): falsch | gilt nicht: eine Zeichenkette ohne Grenze, die Prüfung ist ein Vergleich mit acht Namen | gilt nicht: kein `ctx` | Nr. 8 |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Reine Funktionen mit Allaussagen: `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, verkörpert in `AGENTS.md` §3.16) gelten für die Wahrheitstafeln und das Raster. Offen: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); er erreicht mit diesem Slice nicht 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

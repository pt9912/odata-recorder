# Slice slice-v2-query-filter: Leser von `$filter` des V2-Umfangs (Grammatik, Kompatibilitätsfehler im Ausdruck)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-QA-05`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-173` Nr. 3 und 5 (Literale, Grammatik von `$filter`; Nr. 5 trägt seit 2026-11-05 die Absätze *Der Leser*, *Rangfolge im Ausdruck*, *Weitere Konstrukte, die V2 nicht kennt* und *Literal-Lauf ohne Präfix*) · `SPEC-174` (Kompatibilitätsfehler, Ursachen `filter` und `literal_typ`) · `SPEC-160` Nr. 8 (Einhängestelle `WithFilter`, in `slice-v2-query` geliefert) · `SPEC-164` (V4-Grammatik, die V2 mit Abweichung kopiert; Nr. 1 bis 4 und 7 bis 9 gelten wortgleich, Nr. 6 gilt mit der Zuordnung der Fehler von `SPEC-173` Nr. 5 *Der Leser*) · `SPEC-165` (Query-AST) · `SPEC-170` Nr. 3 · `LH-FA-17.a` (versionierter Adapter, keine Vermischung von V2 und V4) · `LH-FA-17` Negative (nicht unterstützte V2-Eigenschaft: expliziter Kompatibilitätsfehler) · `ARC-001`, `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** — bis zur Priorisierung.

**Autor:** pt9912. **Datum:** 2026-11-01.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Leser von `$filter` für V2 (`urlparserv2.NewFilterLeser()`, `SPEC-173` Nr. 5) liest den dekodierten Wert der Option in den Query-AST des Domain-Modells (`SPEC-165`), mit den V2-Literalen aus `slice-v2-query` (`SPEC-173` Nr. 3), und hängt sich über `WithFilter` in den Parser ein; ein Operator, eine Funktion oder ein Pfad außerhalb des Umfangs ist der Kompatibilitätsfehler `filter`, ein Literal außerhalb `literal_typ` (`ODR-E6006`, `SPEC-174`), ein ungültiger Ausdruck `ODR-E6005`. Der Slice liest Text und kennt das Modell nicht; er schreibt keine Antwort.

**Übernimmt (Sendung von `slice-v2-query`, Architect-Ausgang zum Schnitt, `AGENTS.md` §3.13):** die Grammatik von `$filter` (`SPEC-173` Nr. 5) samt Lexer, Tiefe und Rangfolge der Operatoren; die Zuordnung der Fehler des Lesers (`SPEC-173` Nr. 1 nach der Berichtigung L7: `*URLFehler` der Klasse `ungueltig` und `*KompatFehler` mit `filter` oder `literal_typ` gelten mit ihrer Ursache, Ort `query/<n>`); die Kompatibilitätsfehler `filter` und `literal_typ` im Ausdruck; die Berichtigung L6 in `SPEC-173` Nr. 5 (`has`, `in`, `divby`, V4-Literale ohne Präfix, `INF` und `NaN`, Pfad im Ausdruck); die Literal-Funktion und den Lauf aus `slice-v2-query` (Code-Abhängigkeit, §4). Der Slice schließt `$filter` nicht aus; er ist die Adresse dafür.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Pfad, Schlüsselliterale, Systemoptionen, Kompatibilitätsfehlerwert `KompatFehler`, Code `ODR-E6006`, Abbildung in `outbound`, Feld `URL` in `CodecSatz` und die Einhängestelle `WithFilter` samt `FilterLeser` — `slice-v2-query`; dieser Slice nutzt sie und ändert keine von ihnen (Teile ohne geteilten Mechanismus mit der Grammatik).
- Ausführung des Ausdrucks und Typprüfung gegen das Modell — der Query-Interpreter aus `slice-query-interpreter` (`welle-mock`); der Parser kennt das Modell nicht (`SPEC-171` Nr. 5).
- Verdrahtung `CodecSatz{URL: urlparserv2.New(urlparserv2.WithFilter(urlparserv2.NewFilterLeser()))}`, Rundlauf und Handbuch — `slice-v2-bootstrap`, das diesen Slice als Voraussetzung nennt.
- Methode (`methode` erzeugt der Mock, `SPEC-174` Nr. 3) — `slice-v2-bootstrap`; der Parser hat keine Methode.
- Änderungen am V4-Parser `urlparser` — [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md): kein V2-Paket importiert ein V4-Paket; die Grammatik ist eine Kopie mit Abweichung (Produktcode rund 400 Zeilen).
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`.

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

- [ ] Leser: `urlparserv2.NewFilterLeser()` liest die Grammatik von `$filter` des V2-Umfangs (Vergleiche, Verknüpfungen, `not`, Klammern, die Literale von `SPEC-173` Nr. 3) in den Query-AST; ungültig `ODR-E6005`, ein Operator, eine Funktion oder ein Pfad außerhalb `filter`, ein Literal außerhalb `literal_typ` (`ODR-E6006`), nie eine scheinbar gültige Antwort; `New(WithFilter(NewFilterLeser()))` ordnet die Fehler des Lesers nach `SPEC-173` Nr. 1 (L7) zu.
- [ ] Randformen des Lesers fremder Bytes (Zeichen außerhalb der Grammatik, Minus-Läufe, Leerraum, Tiefe, Länge, Wiederholung, V4-Formen ohne Präfix): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der Ausdrücke, Allaussage über Operatoren und Funktionen außerhalb des Umfangs.
- [ ] Nebenläufigkeit: gleichzeitige Aufrufe des Lesers und Wächter der Knotenzahl und von `TotalAlloc` (Zeilen der Tabelle in §3).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (die Berichtigung L6 in `SPEC-173` Nr. 5 steht vor dem ersten Code-Commit).
- [ ] Der Handbuch-Eintrag entfällt: der Slice bringt keinen Code und kein Kommando in den Katalog des Binary (`ODR-E6006` kommt mit `slice-v2-query`; das Binary ändert sich erst mit `slice-v2-bootstrap`).
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
| `src/adapters/driven/odata/urlparserv2/filter.go`, `filter_token.go` | neu | Leser `NewFilterLeser()` und Lexer (`SPEC-173` Nr. 5); Kopie von `urlparser` mit Abweichung (Operatorwörter, Literale mit Präfix, keine Zeit mit Zone, Literal-Lauf ohne Präfix nach der Form), Produktcode rund 400 |
| `src/adapters/driven/odata/urlparserv2/` (Tests) | neu | Tabelle *Zusage · Fänger* unten, Zähl-Fake und Generator; eigene Test-Hilfen, keine geteilten mit `urlparser` (Black-Box-Tests anderer Pakete sind nicht importierbar); Tabellen der Literale aus `literale_test.go` darf das Testpaket `urlparserv2_test` mitbenutzen |
| `src/adapters/driven/odata/urlparserv2/` (Produktcode aus `slice-v2-query`) | update | nur, was die Grammatik an Literal-Funktion und Lauf braucht; die Signaturen aus `slice-v2-query` bleiben (Stand Handoff: der Produktcode von `slice-v2-query` ist unverändert; die Grammatik ruft `leseLiteral`, `asciiKlein`, `einfacherName`, `nurZiffern`, `hexWert`, `gueltigerText`, `fehler` und `maxLiteral` unverändert) |
| `src/adapters/driven/odata/urlparserv2/literale_test.go` | update | die Listen der Formen (35 Schlüsselliterale, 17 + 3 Guid-Randformen, 21 + 21 Zeitpunkte) stehen als Funktionen `schluesselLiterale`, `guidAbweichend`, `guidRoh`, `datetimeGut`, `datetimeSchlecht`, damit Zeile 18 über dieselben Listen läuft; die Tests von `slice-v2-query` prüfen unverändert dieselben Fälle |
| `spec/spezifikation.md` (`SPEC-173` Nr. 5, *Ungeprüft*, Historie) | update | Architect, 2026-11-05, vor dem ersten Code-Commit (`AGENTS.md` §3.12): Signatur und Fehlerform des Lesers, die wortgleich geltenden Nummern von `SPEC-164`, erster Verstoß im Ausdruck, Minus, `$it`, Alias und qualifizierter Name als `filter`, Literal-Lauf ohne Präfix; nach dem Review berichtigt (Architect, 2026-11-05, F-5): der Satz zur Lesbarkeit eines V4-förmigen Laufs folgt dem Absatz *Literal-Lauf ohne Präfix* c (die Form entscheidet: `2020-13-45` und `12:30x` sind `wert`), Code und Tabelle blieben |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten (37 Zeilen, gefüllt vom Architect am 2026-11-05); das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) steht vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen hat der Architect entschieden, nicht der Implementer; ihm fehlt eine Zelle, hält er an (`AGENTS.md` §3.12).

- Der Leser liest fremde Bytes (den Text des Ausdrucks): die Produktcode-Erwartung des Architect (400) ist ein fester Anteil der Schätzung (`SPEC-128`); Eingaben mit Escape-Folgen oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).
- Das Paket `urlparserv2` importiert kein V4-Paket ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)); `TestKeinV4Import` aus `slice-v2-query` deckt die neuen Dateien mit.
- Die Tests dieses Slice tragen das Namenspräfix `TestFl…` (`TestLiteralLaenge`, `TestNamen` und weitere Namen liegen schon in `slice-v2-query`).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; 37 Zeilen: 2 `nebenläufig` (32, 33), 35 `andere`). Mengen aus dem Wortlaut der Zusage mit ihrer Elementzahl; ein Fänger ist eine Schleife oder ein Generator über die Menge, nie eine Auswahl (`AGENTS.md` §3.17). „Nr. 5“ meint `SPEC-173` Nr. 5 mit seinen Absätzen, „164“ die wortgleich geltende Nummer von `SPEC-164`; Eingaben mit Escape-Folge, Rohbytes, U+FEFF, U+00A0 und Steuerzeichen entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Jede Mutation fährt der Implementer vor der Übergabe selbst und trägt Zusage, Mutation und roten Test in den Bericht ein.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Vergleiche (164 Nr. 3): `eq ne gt ge lt le` mit jedem Paar von Operandenarten. Menge 54 = 6 Operatoren × 9 Paare (Eigenschaft, Literal, `null` je links und rechts) | `TestFlVergleiche`: Schleife, erwarteter Baum aus Knoten | `ge` als `gt`; Operanden vertauscht | andere |
| 2 | `and`, `or`, `not` bauen `FilterVerknuepfung` bzw. `FilterNicht` (164 Nr. 3). Menge 6: `and` und `or` über (Eigenschaft, Eigenschaft) und (Vergleich, Vergleich), `not` über Eigenschaft und geklammerten Vergleich | `TestFlVerknuepfungen`: Tabelle | `and` als `or`; `not` ohne Knoten | andere |
| 3 | Rangfolge (164 Nr. 3): Klammer, `not`, `gt ge lt le`, `eq ne`, `and`, `or`. Menge 58 = 2 Schreibfolgen × (21 + 8): jedes Paar der fünf Stufen mit jedem Operator der Stufen (21 Paare zweier binärer Stufen, 8 mit `not`); Baum gleich dem mit Klammer um die stärkere Stufe, darunter `not Price gt 5` ist `(not Price) gt 5` | `TestFlRangfolge`: Paar-Generator, Vergleich mit der geklammerten Fassung | `and` und `or` gleiche Stufe; `not` schwächer als Vergleiche; `gt` schwächer als `eq` | andere |
| 4 | Assoziativität links (164 Nr. 3). Menge 18 = 8 binäre Operatoren × Ketten von 3 und 4 Operanden (16) + 2 gemischte Ketten derselben Stufe (`a eq 1 ne 2`, `a gt 1 le 2`) | `TestFlAssoziativitaet`: Schleife, Vergleich mit linksgeklammerter Fassung | rechts gebaut | andere |
| 5 | Klammern (164 Nr. 3 und 9): überschreiben die Rangfolge, sind kein Knoten, `(a and b) and c` gleich `a and b and c`, `a and (b and c)` verschieden. Menge 67 = die 29 Operatorpaare von Zeile 3 mit Klammer um die schwächere Seite (50) + 3 Stellen (Operand, Wurzel, Operand von `not`) × Tiefe 1 bis 5 (15) + 2 Assoziativ-Paare | `TestFlKlammern`: Schleife | Klammer als Knoten; Klammer liest ihren Inhalt ab `and` | andere |
| 6 | Schlüsselwörter ohne Beachtung der ASCII-Schreibweise (164 Nr. 3); die Operatorwörter von V2 sind `filter` (Nr. 5, L6). Menge 196 = 12 Wörter (`eq ne gt ge lt le and or not true false null`, alle Schreibweisen der Bitmasken: 108) ergeben denselben Baum wie die Kleinschreibung + 8 Operatorwörter (`has in add sub mul div divby mod`, 84 Schreibweisen) sind alle `filter` + 4 Namen (`id` gegen `Id`, `Nullx`, `ANDx`, `Truex`) | `TestFlSchluesselwoerter`: Bitmasken-Schleife | nur Kleinschreibung erkannt; `divby` nicht in der Liste; `has` als `wert` | andere |
| 7 | Ein unbekanntes Token am Platz eines Operators ist `wert`, nie `filter` (164 Nr. 4, letzter Absatz). Menge 12: `Id foo 1`, `Id equals 1`, `Id = 1`, `Id == 1`, `Id != 1`, `Id > 1`, `Id <= 1`, `Id && 1`, `Id eqq 1`, `Id e 1`, `Id 1`, `Id 'x'` | `TestFlOperatorUnbekannt`: Tabelle mit Ursache | unbekanntes Wort als `filter` gemeldet | andere |
| 8 | Am Platz eines Operanden sind Schlüsselwörter Namen, `not` ist Operator (164 Nr. 4 d). Menge 19: 16 Wörter (`eq ne gt ge lt le and or has in add sub mul div divby mod`) als `W eq 1` ergeben `FilterEigenschaft`; `not eq 1`, `not` allein, `not not` sind `wert` | `TestFlOperatorwortAlsName`: Schleife | Operatorwort am Anfang als Operator gelesen | andere |
| 9 | Namen (164 Nr. 4 c): 1 bis 128 Zeichen gültig, 129 `wert`; Klasse Buchstabe und Ziffer (Kategorien `L`, `Nd`; nicht `Nl`, `No`, `M`) am ersten Zeichen und danach, ohne Normalisierung, die Schreibweise zählt. Menge 66 = 29 Zeichen (6 `L`, 4 `Nd`, 3 `Nl`, 4 `No`, 3 `M`, 9 sonstige) × 2 Stellen (58) + Länge 1, 128, 129 (3) + zusammengesetzt und zerlegt (2) + `Id`/`id` (2) + Ziffer am Anfang (1, `literal`). Prädikat `unicode.Is` ist Orakel (**akzeptiertes Negativ**, Raster *Namen* unten) | `TestFlNamen`: Schleife, Orakel `unicode.Is` | `Nl` als Buchstabe; 129 angenommen; das zerlegte Paar der Fixture zum zusammengesetzten Zeichen zusammengesetzt (Stellvertreter der Normalisierung, die Standardbibliothek hat keine); Name klein geliefert | andere |
| 10 | Zeichen außerhalb der Grammatik sind `wert` (164 Nr. 2, 4 c 4, 4 e): jedes druckbare ASCII-Zeichen außer Buchstabe, Ziffer, `_` (31) und jedes Steuerzeichen U+0001 bis U+001F, U+007F (32) am Anfang eines Laufs, in der Mitte und am Ende eines Namens. Menge 189 = 63 × 3 Stellen; die Ausnahmen (`.`, `/`, `$`, `@`, `+`, `-`, `'`, `(`, `)`, `,`) stehen im Test als Tabelle der erwarteten Ursache (`filter`, `literal`, `wert`) | `TestFlZeichenAussenGrammatik`: Schleife über Zeichen × Stelle, Orakel = Tabelle der Ausnahmen | ein Zeichen zugelassen (`#` als Name); Ausnahme `.` als `wert` | andere |
| 11 | Es gibt keine Kommentare (164 Nr. 2): `//`, `/*`, `#`, `;` und der Rückstrich beginnen nichts und lassen nichts aus; am Platz eines Operators sind sie `wert`, am Platz eines Operanden `wert`, `--` ist ein unäres Minus und `filter` (Nr. 5, *Weitere Konstrukte*). Menge 12 = 6 Formen (`-- c`, `// c`, `/* c */`, `# c`, `; c`, ein Rückstrich) × 2 Plätze | `TestFlKommentarformen`: Tabelle | Kommentar bis Zeilenende ignoriert; `--` als `wert` | andere |
| 12 | Minus-Läufe (164 Nr. 4 b; Nr. 5, *Literal-Lauf ohne Präfix* und *Weitere Konstrukte*): `-` mit Ziffer oder `.` ist ein Literal-Lauf (Zahl, `literal_typ`, `literal`), genau `-INF` ist `literal_typ`, jeder andere Lauf mit `-` ist ein unäres Minus, `filter`. Menge 35 = 5 Folgen nach `-` (Ziffer, `.`, Buchstabe, `-`, `(`) × 7 Fortsetzungen (Ende, `1`, `5L`, `5M`, `1e3`, `INF`, `INFO`); Orakel = Tabelle der Regel im Test. **Prüfgrenze:** die Klasse „Läufe mit `-`“ ist unbeschränkt; die Auswahl ist ein akzeptiertes Negativ (Raster *Minus-Läufe* unten) | `TestFlMinusLaeufe`: Schleife | `-Id` als Literal; `-INF` als `filter`; `-5M` als `literal` | andere |
| 13 | Leerzeichen (164 Nr. 2): 1, 2, 5 zwischen Token; um `(` und `)` an 4 Stellen mit 0, 1, 3; Pflicht zwischen Läufen (`Id eq1`, `Ideq 1`, `Id eq 1and Name eq 2`, `Id eq 1 andName eq 2` sind Fehler); am Rand `wert`. Menge 24 = 3 + 12 + 4 + 5 (Anfang, Ende, beides, nur 1, nur 3); der Test führt dazu den Zusatzfall `Id eq1 1` (Menge 25), damit ein angenommenes `eq1` sichtbar wird | `TestFlLeerzeichen`: Schleife | Rand getrimmt; `Id eq1` angenommen | andere |
| 14 | Anderer Leerraum und unsichtbare Zeichen trennen nichts (164 Nr. 2): U+0009, U+000A, U+000D, U+000B, U+000C, U+0085, U+00A0, U+2003, U+2028, U+3000, U+200B, U+FEFF (12) an 5 Stellen (Anfang, zwischen Operanden, nach dem Operator, nach `(`, Ende) sind Fehler mit der Ursache der Form des Laufs (60); im Inhalt einer Zeichenkette bleiben sie Text (12). Menge 72 | `TestFlAndererLeerraum`: Schleife; Zeichen aus Bytewerten (`AGENTS.md` §3.18) | `unicode.IsSpace` als Trenner; im String entfernt | andere |
| 15 | Länge geht vor Form (164 Nr. 4 a, 7; Nr. 5): 512 Bytes mit Anführungszeichen gültig, 513 `zu_lang`. Menge 18 = 6 Formen (Zeichenkette, ungeschlossenes `'`, `X'…'`, `guid'…'`, `datetime'…'`, Ziffernlauf) × 2 Seiten (12) + Lauf in V4-Form `2020-01-01` + Füllung (2, `wert` und `zu_lang`) + Suffix `d` mit Füllung (2, `literal_typ` und `zu_lang`) + Exponent mit Füllung (2) | `TestFlLiteralLaenge`: gebaute Eingaben genau auf und über der Grenze | Grenze 511 oder `>=`; Form vor Länge (`literal_typ` statt `zu_lang`) | andere |
| 16 | Der Rest hinter dem schließenden `'` zählt nicht zur Länge (164 Nr. 4 a): `'a'` plus 600 Bytes Rest ist `literal`, nicht `zu_lang`; 512 Bytes plus kurzem Rest ist `literal`, 513 plus 1 Rest `zu_lang`. Menge 3 | `TestFlRestZaehltNicht` | Länge des ganzen Laufs statt der Zeichenkette gemessen | andere |
| 17 | Ende eines Literals (164 Nr. 4 a): bis U+0020, `)`, `,` oder Textende. Menge 27 = 6 Arten (Zeichenkette, Wahrheitswert, Zahl, `guid'…'`, `datetime'…'`, `X'…'`) × 3 Ender (18) + 6 Arten mit `,` danach (`wert`; hinter `X'…'` ist es `literal_typ`, denn der erste Verstoß in der Lesefolge entscheidet) + `'a'b`, `'a'1`, `'a''` (`literal`) | `TestFlLiteralEnde`: Schleife | Literal bis zum ersten `)` im String; `,` als Ender nicht geprüft | andere |
| 18 | Literale gleich dem Schlüssel (Nr. 5 mit Nr. 3): jede Form aus `slice-v2-query` (Tabelle Zeilen 28 bis 31) ergibt als `Id eq <Literal>` dieselbe Art und denselben Text bzw. dieselbe Ursache wie `/Set(<Literal>)` (`literal_typ` als `*KompatFehler`, nicht `*URLFehler`). Menge 122 = 35 (Z28) + 20 Randformen (Z29, `guid'…'`) + 42 (Z30, `datetime'…'`) + 25 Formen (Z31, `literal_typ`); 6 Ausnahmen als Liste mit erwartetem Ergebnis und Grund im Test: `null` (Knoten), nacktes Wort (Name), Guid und Datum ohne Präfix (`wert`, Zeile 19), `INF`, `NaN` (`literal_typ`) | `TestFlLiteraleWieSchluessel`: Schleife über die Formenlisten von `literale_test.go` | Guid nicht kleingeschrieben; `L` oder `M` im Text; Filter liest `datetime'…'` mit Zone; eigene Kopie der Literal-Funktion weicht ab | andere |
| 19 | Literal-Lauf ohne Präfix, in der Folge Zahl · `literal_typ` · `wert` (Form von V4) · `literal` (Nr. 5, *Literal-Lauf ohne Präfix*). Menge 24 = 15 `wert` (Guid-Form in 4 Schreibweisen, auch mit Buchstaben am Anfang; Datum 3: `2020-01-01`, `2020-13-45`, `2020-01-01x`; Zeitpunkt 4: `…T00:00:00Z`, mit `+02:00`, ohne Zone, mit Bruchteil; Uhrzeit 4: `12:30`, `12:30:00`, `12:30:00.5`, `12:30x`) + 3 `literal_typ` (`INF`, `NaN`, `-INF`) + 6 `literal` (`2020-1-1`, `+1`, `.5`, `5.`, `1e`, `0x10`) | `TestFlV4FormenOhnePraefix`: Tabelle mit Ursache, Typ `*URLFehler` bzw. `*KompatFehler` | Guid ohne Präfix als Literal gelesen; `INF` als Name; `2020-1-1` als `wert`; Bereich geprüft (`2020-13-45` als `literal`) | andere |
| 20 | `null` (164 Nr. 4 c 7): `FilterNull`, 16 Schreibweisen rechts, links, `null eq null`; `nullx`, `xnull`, `null1`, `_null` sind Namen. Menge 37 = 16 × 2 + 1 + 4 | `TestFlNull`: Schleife | `null` als Name; nur `null` kleingeschrieben | andere |
| 21 | Die Wurzel wird nicht typgeprüft (164 Nr. 10, *Ungeprüft*): jeder Operand ist als Wurzel lesbar. Menge 9: `Id`, `5`, `'x'`, `null`, `true`, `Id eq 'x'`, `1 eq 'a'`, `'a' eq 1`, `not 5` | `TestFlWurzelOhneTyp`: Tabelle mit Baum | Parser lehnt Nicht-Boolesches ab | andere |
| 22 | Ein Pfad ist `filter` (Nr. 5, L1, L6), kein `wert_nicht_unterstuetzt`, nie `*URLFehler`; ein fehlerhafter Pfad ist `wert` (164 Nr. 4 c 5). Menge 10: `a/b`, `a/b/c`, `a/b/c/d`, `a/b eq 1`, `1 eq a/b` (`filter`); `a/`, `/a`, `a//b`, `a/1`, `a/b.c` (`wert`); über `ParseURL` bleibt die Ursache `filter`, nicht `option` (`SPEC-173` Nr. 1) | `TestFlPfad`: Tabelle, Typ und Ursache | Pfad als Name mit `/`; Pfad als `option`; `a/1` als Pfad | andere |
| 23 | Jeder Name unmittelbar vor `(` ist ein Aufruf, `filter`, auch ein unbekannter; mit Leerraum davor ist es keiner (Nr. 5, 164 Nr. 4 c 1 und 5 a). Menge 168 = 33 Namen × 5 Formen + 3: 22 Funktionen aus `SPEC-170` Nr. 3, 5 Namen von V4 (`contains`, `matchesPattern`, `in`, `any`, `all`), `geo.distance`, 5 erfundene (`foo`, `Ns.foo`, `a/any`, `true`, `x1`); Formen `N()`, `N(Id)`, `N(Id,1)`, `N(f(Id))`, `N(Id` ohne `)`; dazu `Name (Id)` (`wert`), `not(Id)` und `not (Id)` (Operator) | `TestFlAufruf`: Schleife über Namen × Formen | nur Namen der Liste; Leerraum vor `(` als Aufruf | andere |
| 24 | Operatoren und weitere Konstrukte, die V2 nicht kennt, sind `filter` (Nr. 5, L6 und *Weitere Konstrukte*). Menge 49 = 24 (8 Operatorwörter `add sub mul div mod has in divby` × 3 Einbettungen: `1 W 2 eq 3`, `Id W 1`, `(Id W 1)`) + 25: Minus (`-Id`, `- 5`, `-(5)`, `--5`, `-INFO`), `$it`, `$root`, `$this`, `$it/Id`, `$this/Id`, `$root/Id` (`$x` ist `wert`), Alias `@p` (`@`, `@1` sind `wert`), qualifizierter Name (`Ns.Typ`, `a.b`, `a.b.c`; `a.`, `a..b` sind `wert`, `.a` ist `literal`), `Id in (1,2)`, `Id in [1]`, `Items/any(x: x eq 1)`, `isof('Ns.T')` | `TestFlNichtUnterstuetzt`: Tabelle mit Ursache und Typ | `-Id` als Literal; `$it` als Name; qualifizierter Name als `wert`; `mod` nicht in der Liste | andere |
| 25 | Tiefe (164 Nr. 7): 32 offene `(` und `not` zusammen gültig, die 33. `zu_viele`; die Tiefe wird beim Schließen abgebaut. Menge 77 = n = 1 bis 32 für `(` und für `not` (64) + 33, 34, 100, 4000 für beide (8) + gemischt 16 + 16 gültig, 17 + 16 ungültig, je in zwei Schreibweisen (4) + 40 flache Gruppen und 40 flache `not` nacheinander (1) | `TestFlTiefe`: Schleife, Generator der Eingabe | Grenze 31 oder 33; `not` zählt nicht; Zähler nicht zurückgesetzt | andere |
| 26 | Der erste Verstoß in der Lesefolge entscheidet, gleich welcher Klasse (Nr. 5, *Rangfolge im Ausdruck*, 164 Nr. 6). Menge 49 = Paare (A vor B) über 7 Verstoßformen (`wert`, `literal`, `zu_lang`, `zu_viele`, `filter` als Aufruf, `filter` als Pfad, `literal_typ`): 42 verschiedene Paare + 7 gleiche | `TestFlRangImAusdruck`: Paar-Generator (Ausdruck mit A, dann B) gegen Ursache und Typ von A | die Klasse `ungueltig` hinter einem `filter` gemeldet; zweiter Verstoß gemeldet (Aufruf und Pfad vertauscht ist äquivalent, siehe *Akzeptierte Negative*) | andere |
| 27 | Fehlerwert (Nr. 5, *Der Leser*, `SPEC-174` Nr. 1): `*URLFehler` mit Code und Sentinel der Klasse `ungueltig` (`errors.Is` genau eine) für `kodierung` (nur unmittelbar), `wert`, `literal`, `zu_lang`, `zu_viele`; `*KompatFehler` mit `Code()` `ODR-E6006` und `errors.Is(err, model.ErrKompatibilitaet)` für `filter` und `literal_typ`, nie beides; Ergebnis `nil`, `Ort` leer; über `ParseURL` ist der Ort `query/<n>`; kein Teil der Eingabe (kein Name, Literal, keine Position) steht in `Error()` oder `Ort`; die Kodierung geht der Grammatik vor (`kodierung` auch hinter `foo(`). Menge 52 = 7 Ursachen unmittelbar + 3 `kodierung` hinter einem Aufruf (U+0000, die Bytes 0xFF, die Bytes 0xC3 0x28) + 7 Ursachen × `n` = 1, 2, 3 (21) + 7 Ursachen × 3 Marker-Stellen (Name davor, Literal davor, im Verstoß-Token selbst; 21) | `TestFlFehlerwert`: Schleife; Marker in der Eingabe, Suche im Text | Position im Fehler; Ort des Lesers gesetzt; `filter` als `*URLFehler`; `Code()` vertauscht; `kodierung` nach der Grammatik | andere |
| 28 | Einhängung mit echtem Leser (`SPEC-173` Nr. 1 c und d, `SPEC-160` Nr. 8): `Filter.Text` ist der dekodierte Wert, `Filter.Ausdruck` die Wurzel, die Reihenfolge der Optionen ändert nichts; `New()` und `WithFilter(nil)` bleiben `filter`; die Klasse `ungueltig` des Lesers und die einer anderen Option ordnet `SPEC-173` Nr. 6; `$filter` doppelt ist `option_doppelt`, der Leser wird für das zweite nicht gerufen. Menge 15 = 6 Bäume + `$filter` an 3 Stellen der Query + 2 ohne Leser + 4 Rangfolgen (Leser-`filter` vor späterem `wert` einer Option: `wert`; Leser-`wert` vor einer Kompatibilitäts-Option: `wert`; frühere Kompatibilitäts-Option vor Leser-`filter`: die erste; doppeltes `$filter`) | `TestFlEinhaengung`: Tabelle über `ParseURL` mit dem Zähl-Fake aus `slice-v2-query` für den Aufruf je `ParseURL` | `Filter.Text` leer; Rang der Klassen ignoriert; Leser mit rohem Wert gerufen; Leser bei doppeltem `$filter` gerufen | andere |
| 29 | Dekodierung (`SPEC-162` Nr. 2 mit Nr. 5, *Der Leser*): der Leser sieht den einmal dekodierten Text. Menge 22 = 6 Zeichen (`'`, `(`, `)`, U+0020, `,`, `/`) × 3 Schreibweisen (roh, `%XX`, `%xx`) gleicher Baum (18) + `%2527` ergibt in der Zeichenkette den Text `%27` (1) + `+` bleibt `+` (`Id+eq+1` ist `wert`, `'a+b'` hat `Text` `'a+b'`, `+` roh gleich `%2B`) (3) | `TestFlDekodierung`: Schleife über `ParseURL` | doppelt dekodiert; `+` als Leerzeichen | andere |
| 30 | Struktur (164 Nr. 3 und 6): fehlender oder überzähliger Operand, Operator oder Klammer, `,` sind `wert`. Menge 24: `()`, `(`, `)`, `(Id`, `Id)`, `((Id)`, `(Id))`, `Id eq`, `Id eq 1 and`, `Id eq 1 or`, `not`, `Id eq 1 not`, `Id 1`, `Id Id`, `Id eq 1 2`, `1 2`, `(Id) (Id)`, `Id,`, `,Id`, `Id eq 1,`, `(Id,Id)`, `Id eq ()`, `not ()`, `Id eq not` | `TestFlStruktur`: Tabelle | überzähliges Token ignoriert; `()` als wahr | andere |
| 31 | Wiederholung und Nullwert (164 Nr. 8): gleicher Text, gleicher Baum (tiefe Gleichheit) und gleicher Fehler, auch auf `NewFilterLeser()` und `FilterGrammatik{}`; der Baum teilt nichts mit einem anderen (Ändern eines Wertes am ersten Ergebnis verändert das zweite nicht). Menge 64 × 2 Aufrufe × 2 Instanzen: 40 Gutfälle (aus den Zeilen 1 bis 5, 20 und 21, je Zeile mindestens 5) und die 24 Fehlerfälle der Zeile 30 | `TestFlWiederholung`: Schleife | letzter Baum zwischengespeichert; Zustand der Instanz zwischen Aufrufen | andere |
| 32 | Gleichzeitige Aufrufe liefern je das Ergebnis des einzelnen Aufrufs (164 Nr. 8). Menge: 16 Goroutinen × 64 (Korpus der Zeile 31) × 10 Runden, dazu 2 verschränkte Ausdrücke (Tiefe 32 gültig, Tiefe 33 `zu_viele`): ein Tiefenzähler im Feld der Instanz liefert dem einen das Ergebnis des anderen (Wirkungs-Fänger, unabhängig vom Detektor) | `TestFlGleichzeitig` | Tiefenzähler und Puffer als Feld der Instanz; Rate je Lauf von `go test` gemessen | nebenläufig |
| 33 | Kein Zustand: weder die Instanz noch eine Variable des Pakets wird von einem Aufruf geschrieben (164 Nr. 8) | `make test-race` über `TestFlGleichzeitig` (jeder Prozess `-test.count=1`), Rate je Prozess gemessen | Cache (`map`) für gelesene Literale in einer Variablen des Pakets | nebenläufig |
| 34 | Wächter (164 Nr. 7 a bis c): Knoten ≤ Länge, Zuteilung ≤ 64 · n + 65536, auch für `KompatFehler`. Menge 40 = 8 Familien (`(`, `not `, `'`, `''`, `a and `, `a eq 1 or `, `-`, `x'`) × n = 8192 und 4096 (16) + 24 Eingaben aus rohen Bytes und Escapes; die Knoten werden an den 40 Gutfällen der Zeile 31 gezählt | `TestFlWaechter`: `TotalAlloc` und Knotenzählung | Verkettung `s += …` im Lauf; Kopie des Slices je Token | andere |
| 35 | Allaussage: jeder Operator und jede Funktion außerhalb des Umfangs ist `filter` (`SPEC-170` Nr. 3, Nr. 5). Menge 99 = 33 Namen (5 Operatoren `add sub mul div mod`; 28 Funktionen: die 22 von `SPEC-170` Nr. 3, `contains`, `matchesPattern`, `in`, `any`, `all`, `geo.distance`) × 3 Einbettungen (Wurzel, Operand eines Vergleichs, hinter `and`); die Listen stehen im Test aus dem Wortlaut der Spezifikation, nicht aus dem Code | `TestFlAusserhalbAllaussage`: Schleife über beide Listen × Einbettungen | Aufruf nur für Namen einer Liste ohne `trim`; Operator `mod` fehlt | andere |
| 36 | Umkehrung und Klasse: zufälliger Ausdruck (alle Knotenarten, alle Literalarten aus Zeile 18, Tiefe bis 6, 1 bis 3 Leerzeichen, Schreibweise der Schlüsselwörter), vom Test mit minimaler Klammerung zu Text gerendert, gleicher Baum; mit einem Einsprengsel aus den Klassen außerhalb (Operator, Funktion, Pfad, Literal fremden Typs, V4-Form) an zufälliger Stelle: Ursache des ersten Verstoßes nach dem Renderer. Menge: 2000 Läufe, Saat 1 bis 2000; Orakel ist der Renderer samt Verdikt-Regel des Tests, nie der Leser. **Prüfgrenze** (Raster *Menge der Ausdrücke*) | `TestFlGenerator`: Orakel = der Renderer des Tests; Zählprüfung der Läufe je Verstoßform | Rangfolge, Assoziativität oder kanonischer Text verändert (je eine Mutation, Rate je Lauf); Einsprengsel ungezählt | andere |
| 37 | Echte V2-Filter: Menge 17 = die 5 `filter`-Anfragen der Erfassung aus `TestEchteV2URLs` (`Age gt 30`, `Age gt 999` je mit und ohne `$format`, `Age bogus 30`, `/Employees/$count` mit `Age gt 30`: mit Leser vier Bäume und `Age bogus 30` ist `wert`) + 7 (`substringof('a',Name) eq true` `filter`, `Date gt datetime'2020-01-01T00:00:00'`, `Id eq guid'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'`, `Price gt 1.5M`, `Big eq 5L`, `not (Age gt 3)`, `Name eq null`) + 5 weitere (`Name eq 'O''Hara'`, `Id eq 1 and Name ne 'x'`, `startswith(Name,'a') eq true` `filter`, `Price add 1 gt 5` `filter`, `Name eq 'x' or Id eq 1L`) | `TestFlEchteV2Filter`: Tabelle über `ParseURL` mit dem Leser | Literal `datetime'…'` mit Zone; `5L` als `literal`; `substringof` als Name | andere |

**Akzeptierte Negative** (`AGENTS.md` §3.15, Grund und Folge):

- *Laufzeit ohne Zuteilung* (`SPEC-164` *Ungeprüft*, in `SPEC-173` Nr. 5 übernommen). Grund: der Text ist auf 8192 Bytes begrenzt, das schlechteste quadratische Verhalten wären rund 7 · 10^7 Schritte. Folge: eine Regression ohne Zuteilung bleibt bei grünen Gates unbemerkt, das Review des Diffs fängt sie. Fänger: keiner (Review).
- *Prüfgrenze der Zeilen 9, 12 und 36* (Klassen ohne Obergrenze: Zeichen, Läufe mit `-`, Ausdrücke): die Menge ist eine Auswahl, Zeile 36 mischt sie mit 2000 Saaten. Folge im Betrieb: eine Form, die weder Tabelle noch Saat trifft, kann falsch eingeordnet sein; beide Ursachen derselben Seite (`wert`/`literal`) ergeben denselben HTTP-Status und Code.
- *Ein Host mit einem einzigen Kern* (Zeilen 32 und 33): kleinere Rate der Wirkungs-Fänger, kein falsches Grün.
- *Aufruf und Pfad vertauscht* (Zeile 26, äquivalenter Mutant). Grund: ein Aufruf und ein Pfad sind beide `KompatFehler` mit der Ursache `filter` (`SPEC-173` Nr. 5, `SPEC-174` Nr. 2); vertauscht der Code, welcher der beiden Verstöße zuerst gemeldet wird, ändern sich weder Ursache noch Typ noch Ort, und der Vertrag unterscheidet sie nicht. Folge: keine im Betrieb (gleiche Antwort `501` mit `ODR-E6006` und gleicher Ursache); nicht zu verwechseln mit der Reihenfolge der Prüfungen im selben Lauf (Aufruf vor Namens- und Pfadprüfung, `SPEC-164` Nr. 4 c 1), die ändert die Ursache und hat ihren Fänger in Zeile 23. Fänger: keiner, nicht nötig.

**Zusage · Mutation · roter Test** (Implementer, Stand Handoff; `AGENTS.md` §3.10): Jede Mutation lief auf einer Wegwerfkopie von `src/` in einem Container (`--network none --memory=4g --cpus=4`), gegen `go test -count=1` über das Paket `urlparserv2` (die Namen der Tests tragen das Präfix `TestFl`); die Spalte nennt die roten Tests, die der Zeile gehören, weitere rote Tests derselben Mutation stehen im Bericht. Mutationen ohne Zufall (feste Saaten, Tabellen, Schleifen) sind in jedem Lauf rot, solange der Quelltext der Mutation gleich bleibt: ein Lauf je Mutation, 0 grüne. Die Wirkungs-Mutation der Zeile 32 und die Detektor-Mutation der Zeile 33 sind in je 30 Prozessen gemessen (jeder Prozess `-test.count=1`, `TestFlGleichzeitig`; der Detektor ist `go test -race` mit cgo im Image `race`, Messender: der Implementer).

| Zeile | Mutation (Wegwerfkopie) | roter Test |
|---|---|---|
| 1 | `ge` als `gt` gebaut; Operanden des Vergleichs vertauscht | `TestFlVergleiche` |
| 2 | `and` baut `FilterVerknuepfung` mit `Oder`; `not` liefert den Operanden ohne Knoten | `TestFlVerknuepfungen` |
| 3 | `and` auf der Stufe von `or`; `not` liest eine Stufe `gt` statt `unaer`; Stufen `eq` und `gt` vertauscht | `TestFlRangfolge` |
| 4 | rechts gebaut (`stufe(s)` statt `stufe(s+1)`) | `TestFlAssoziativitaet` |
| 5 | Klammer liefert einen Knoten; Inhalt der Klammer wird nur bis `and` gelesen | `TestFlKlammern` |
| 6 | `t.klein` ohne Kleinschreibung; `divby` nicht in der Liste; `has` nicht in der Liste | `TestFlSchluesselwoerter` |
| 7 | unbekanntes Token am Platz eines Operators als `filter` | `TestFlOperatorUnbekannt` |
| 8 | Operatorwort `eq` am Platz eines Operanden als Operator | `TestFlOperatorwortAlsName` |
| 9 | `Nl` als Buchstabe; 129 Zeichen angenommen | `TestFlNamen` |
| 9 | der Name wird in Kleinschreibung geliefert (Ersatz für „Normalisierung“, siehe §6) | `TestFlNamen` und weitere (jede Zeile mit einem Namen in Großschreibung) |
| 9 | das zerlegte Paar der Fixture (Grundzeichen und kombinierendes Zeichen) wird zum zusammengesetzten Zeichen zusammengesetzt (Stellvertreter der Normalisierung; Architect, 2026-11-05) | `TestFlNamen` erwartet; der Implementer fährt sie in der Fix-Runde, bleibt sie grün, ist die Zeile ohne Fänger |
| 10 | `#` als Zeichen eines Namens; `.` im Namen nicht als qualifizierter Name | `TestFlZeichenAussenGrammatik` |
| 11 | Text ab ` --` abgeschnitten; `--` am Platz eines Operanden als `wert` | `TestFlKommentarformen` |
| 12 | Lauf `-` als Literal-Lauf für jede Folge; `-INF` nicht Literal-Lauf; `5M` nicht als Zahl | `TestFlMinusLaeufe` |
| 13 | Text am Rand getrimmt; `eq1` als Operator angenommen | `TestFlLeerzeichen` |
| 14 | jedes Byte unter U+0020 als Trenner; Tabulator im Inhalt einer Zeichenkette entfernt | `TestFlAndererLeerraum` |
| 15 | Grenze 511 (`>=`); Grenze 513; Länge der Zahl nicht vor der Form geprüft | `TestFlLiteralLaenge` |
| 16 | Zeichenkette mit dem Rest dahinter gemessen | `TestFlRestZaehltNicht` |
| 17 | `,` nicht als Ende der Zeichenkette | `TestFlLiteralEnde` |
| 17 | Ende der Zeichenkette am ersten `)` im Inhalt | `TestFlGenerator`, `TestFlLiteraleWieSchluessel` |
| 18 | nur im Filter: eigene Kopie der Literal-Funktion, die die Schreibweise ändert; `L` und `M` bleiben im Text; `datetime'…'` mit Zone gelesen | `TestFlLiteraleWieSchluessel` |
| 18 | in der gemeinsamen Literal-Funktion (Schlüssel und Filter): Guid nicht klein; `L` und `M` im Text; `datetime'…'` mit Zone | `TestGuid`, `TestSchluesselLiterale`, `TestDatetime` (Schlüssel) und `TestFlGenerator`; `TestFlLiteraleWieSchluessel` bleibt grün, weil beide Seiten gleich abweichen (der Test deckt die Abweichung zwischen den Seiten, nicht die der Funktion) |
| 19 | Guid ohne Präfix als Literal gelesen; `INF` als Name; `2020-1-1` als `wert`; Bereich geprüft (`2020-13-45` als `literal`) | `TestFlV4FormenOhnePraefix` |
| 20 | `null` als Name; nur `null` in Kleinschreibung | `TestFlNull` |
| 21 | Parser lehnt eine Wurzel ohne Boolesches ab | `TestFlWurzelOhneTyp` |
| 22 | Pfad als Name mit `/`; Glied vor dem letzten ungeprüft; Glied `1` am Ende als Pfad | `TestFlPfad` |
| 23 | Aufruf nur für zwei Namen; Leerraum vor `(` als Aufruf | `TestFlAufruf` |
| 24 | `-Id` als Literal; `$root` nicht in der Liste; qualifizierter Name als `wert`; `mod` nicht in der Liste | `TestFlNichtUnterstuetzt` |
| 25 | Grenze 31; Grenze 33; `not` zählt nicht; Zähler nach der Klammer nicht abgebaut | `TestFlTiefe` |
| 26 | Kompatibilitätsfehler wird gemerkt und weitergelesen, ein späterer Verstoß der Klasse `ungueltig` gemeldet | `TestFlRangImAusdruck` |
| 27 | Ort im Kompatibilitätsfehler; `filter` als `*URLFehler`; `literal_typ` als `wert`; Ort des Lesers gesetzt; `kodierung` nach der Grammatik; `Code()` des Kompatibilitätsfehlers vertauscht (`ODR-E6005` statt `ODR-E6006`, Wegwerfkopie von `src/hexagon/model`, vom Reviewer gefahren 2026-11-05) | `TestFlFehlerwert` und weitere (`TestFlAufruf`, `TestFlAusserhalbAllaussage`, `TestFlDekodierung`, `TestFlGenerator`, `TestFlKommentarformen`) |
| 28 | `Filter.Text` leer; Leser mit rohem Wert gerufen; Leser bei doppeltem `$filter` gerufen; Kompatibilitätsfehler des Lesers beendet das Lesen | `TestFlEinhaengung` |
| 29 | Wert doppelt dekodiert; `+` als Leerzeichen | `TestFlDekodierung` |
| 30 | überzähliges Token ignoriert; `()` als Wahrheitswert | `TestFlStruktur` |
| 31 | Instanz hält den letzten Fehler | `TestFlWiederholung` |
| 32 | Tiefenzähler als Feld der Instanz (je Aufruf auf 0 gesetzt) | `TestFlGleichzeitig`: rot in 30 von 30 Prozessen ohne Detektor, in 30 von 30 mit Detektor (30 von 30 melden `DATA RACE`) |
| 33 | Cache (`map`) in einer Variablen des Pakets | `TestFlGleichzeitig` (Texte ohne frühere Lesung): mit Detektor in 30 von 30 Prozessen `DATA RACE`; ohne Detektor in 30 von 30 Prozessen rot durch den Laufzeitfehler `concurrent map writes` |
| 34 | Verkettung `s += …` je Lauf; Kopie des Textes je Token | `TestFlWaechter` |
| 35 | Aufruf nur für Namen ohne `trim` | `TestFlAusserhalbAllaussage` |
| 36 | Rangfolge, Assoziativität, kanonischer Text (Zeilen 3, 4 und 18); Einsprengsel `divby` und `mod` aus der Liste | `TestFlGenerator` (feste Saaten 1 bis 2000, in jedem Lauf rot) |
| 37 | Zeitpunkt ohne `Z` im Text; `5L` als `literal`; `1.5M` nicht als Zahl; Aufruf nur für `contains` und `substringof` (`startswith` als Name) | `TestFlEchteV2Filter` |

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `next` → `in-progress`): `slice-v2-query` liegt in `done/` (Code-Abhängigkeit, unten); [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` (erfüllt). Die Berichtigung L6 steht in `SPEC-173` Nr. 5 (seit 2026-11-01, ergänzt 2026-11-05), und die Tabelle *Zusage · Fänger* ist gefüllt (Architect, 2026-11-05; `AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-v2-query` (Literal-Funktion, Lauf, `KompatFehler`, `ODR-E6006`, `FilterLeser` und `WithFilter` liegen dort; B erzeugt ohne sie keinen Fehlerwert und hat keine Einhängestelle), `slice-odata-url-filter` und das Query-AST in `done/` (`SPEC-165`).
- **Was der Slice nicht braucht — und warum:** `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben` (der Parser kennt das Modell nicht), `welle-mock` ganz (der Test über beide Seiten liegt im Rundlauf von `slice-v2-bootstrap`).
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): vorläufig aus der Grobschätzung des Architect, mit der Tabelle neu zu rechnen: 37 Zeilen, davon 2 nebenläufig, mal 85 bzw. 80, keine Fakes, Produktcode 400 (`filter.go` 311 und `filter_token.go` 113, ohne die Guid-Form des Wortlaufs, mit der Zuordnung der Kompatibilitätsfehler): *S* = 85·2 + 80·35 + 400 = 3370, Marke *M* = 100 · ⌈135·3370 / 10000⌉ = **4600**. Der Planner hat `basis`, `schaetzung` und `marke` beim Start (nach `slice-v2-query` in `done/`, vor dem ersten Code-Commit) mit der gefüllten Tabelle gesetzt; die vorläufigen Zahlen bestätigt die Rechnung. **Stand (Planner, 2026-11-05):** Die Tabelle *Zusage · Fänger* ist gefüllt (37 Zeilen: *a* = 2, *b* = 35), *p* = 400 und *f* = 0 bestätigt der Architect; der Block trägt `basis: 80d7910c1bc7e9c5fe1537f7ea66476ef7e35642` (der Stand beim Start, die Messung über den Pfad ist dort 0), `schaetzung: 3370`, `marke: 4600` und `schnitt-grund` (Ausgang des Architect: kein Schnitt). Beide Rechnungen mit den Maßen der Übersicht (2 nebenläufige, 35 andere Zeilen, *p* = 400, *f* = 0): `SPEC-128` Nr. 1: *S* = 85·2 + 80·35 + 400 = 3370, Marke 4600, die Schnittfrage gilt (*S* ≥ 2223), `schnitt-grund` ist Pflicht im Block (`SPEC-128` Nr. 5 g); `AGENTS.md` §3.14 (Leser fremder Bytes, 65 je andere Zeile, Produktcode als fester Anteil, × 1,25): 85·2 + 65·35 + 400 = 2845, × 1,25 = 3556, Marke 3600, über 3000 und damit ebenfalls vor dem ersten Code-Commit zu schneiden oder mit Grund in §4 zu belassen. Der Formelkonflikt ist eine offene Entscheidung des Users; der Block trägt die Fassung von `SPEC-128`. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/adapters/driven/odata/urlparserv2/`, Tests eingeschlossen; die Messung beginnt bei der Basis nach `slice-v2-query` (derselbe Pfad). Erwartung der Messung (Messreihe Ende ÷ *S* = 0,69 bis 0,78): rund 2300 (2100 bis 2600). Der Block steht unten (vom Planner am 2026-11-05 vor dem ersten Code-Commit gesetzt, Messung 0).

```groessenmarke
pfad: src/adapters/driven/odata/urlparserv2/
produktcode: 400
fakes: 0
basis: 80d7910c1bc7e9c5fe1537f7ea66476ef7e35642
schaetzung: 3370
marke: 4600
schnitt-grund: Kein Teil ist ohne geteilten Mechanismus abtrennbar. Der Kern (Zeilen 1 bis 18, 20, 21, 25 bis 34, 36, 37) hätte allein S = 2460 (SPEC-128: 2910) und stellte dieselbe Schnittfrage; die Erkennung der Konstrukte, die V2 nicht kennt (Zeilen 19, 22, 23, 24, 35; S = 385), teilt den Lexer (die Klassen der Läufe entscheiden Name, Pfad, Aufruf, Alias und Literal zugleich) und die Rangfolge der Verstöße mit dem Kern, und ohne sie lieferte der Kern 400 statt 501 für V2-Konstrukte (LH-FA-17 Negative), also keinen Lieferwert. Ein Schnitt zwischen Produktcode und Tests ließe Zusagen ohne Fänger. Der Slice ist eine Kopie der Grammatik von slice-odata-url-filter mit Abweichung (Produktcode rund 400); dessen Ende lag bei 0,50 der Marke, die Erwartung hier liegt bei rund 2300.
```

- **Messung beim Handoff (Implementer, `git diff --numstat <Basis>` über `src/adapters/driven/odata/urlparserv2/`, Tests eingeschlossen):** 2351 Zeilen (2311 Einfügungen, 40 Löschungen), das sind 0,51 der Marke 4600 und innerhalb der Erwartung (2100 bis 2600). Der Produktcode hat 469 Zeilen (`filter.go` 356, `filter_token.go` 113) statt der erwarteten 400; die Tests 1882 Zeilen. Die Löschungen sind die Listen von `literale_test.go`, die als Funktionen stehen (§3). Ein Schnitt war nicht nötig.
- **Schnittfrage — Ausgang des Architect (2026-11-05): kein Schnitt, `schnitt-grund` im Block oben (`SPEC-128` Nr. 5 g).** Die Tabelle hat die 37 Zeilen der Übersicht (2 `nebenläufig`, 35 `andere`), *p* bleibt 400 (die V4-Dateien haben 311 + 113 = 424 Zeilen; die Form-Entscheidung für den Literal-Lauf ohne Präfix, `SPEC-173` Nr. 5, spart eine zweite Bereichsprüfung), *f* = 0; beide Formeln liefern damit die Zahlen des Planner (3370 und Marke 4600 nach `SPEC-128`; 2845, × 1,25 = 3556 nach `AGENTS.md` §3.14). Geprüft wurde der Schnitt entlang der Zeilen: (A) Kern (Zeilen 1 bis 18, 20 bis 21, 25 bis 34, 36, 37: 32 Zeilen, 2 nebenläufig, *S*(A) ≈ 85·2 + 65·30 + 340 = 2460 nach `AGENTS.md` §3.14, × 1,25 = 3075; nach `SPEC-128` 85·2 + 80·30 + 340 = 2910) und (B) Erkennung dessen, was V2 nicht kennt (Zeilen 19, 22, 23, 24, 35: *S*(B) ≈ 65·5 + 60 = 385). (A) allein bliebe über beiden Schwellen (3000 bzw. 2223) und stellte dieselbe Frage; (B) teilt den Lexer (die Klassen der Läufe von `SPEC-164` Nr. 4 c entscheiden Name, Pfad, Aufruf, Alias und Literal zugleich) und die Rangfolge der Verstöße (Zeilen 26 und 28) mit dem Kern; ohne (B) lieferte der Kern `wert` (`400`) für `substringof(…)`, wo `LH-FA-17` Negative `501` verlangt, also keinen Lieferwert. Ein Schnitt nach Tests (Produktcode in einem Slice, Fänger im zweiten) ließe die Zusagen des ersten ohne Fänger (`AGENTS.md` §3.15). Das Vorbild `slice-odata-url-filter` (dieselbe Grammatik, 35 Zeilen, *S* = 3460, *p* = 650) schloss ohne Schnitt mit 2334 Zeilen (Auslastung 0,50 der Marke 4700); die Erwartung hier (rund 2300, Spanne 2100 bis 2600) liegt unter 3000. Der **Formelkonflikt** (`SPEC-128` Marke 4600 gegen `AGENTS.md` §3.14 Marke 3600) ändert den Ausgang nicht, beide verlangen `schnitt-grund`; entscheidet der User ihn, gilt nach der Source Precedence der höhere Rang (`harness/vertraege/` vor `AGENTS.md`), und `AGENTS.md` §3.14 wird angepasst — das ist seine Entscheidung, nicht die des Architect.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 4600 des Blocks (`SPEC-128`; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`, oder `slice-v2-query` verlässt `done/` nicht.

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

- Die Menge der Ausdrücke ist unbeschränkt (*Menge der Ausdrücke*) — **Ausgang:** offen; Generator über die Klasse (Zeile 36: 2000 Läufe, Saat 1 bis 2000) mit Orakel in eigener Datei und Zählprüfung, Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (§3, `AGENTS.md` §3.17).
- Operatoren und Funktionen außerhalb des Umfangs sind eine Allaussage — **Ausgang:** offen; Schleife über die Listen der Spezifikation (Zeile 35: 5 Operatoren, 28 Funktionen, je 3 Einbettungen), nicht über die Fälle, die einfielen (`AGENTS.md` §3.15).
- Die Kopie der Grammatik von `urlparser` und die Literale aus `slice-v2-query` laufen auseinander — **Ausgang:** offen; akzeptiertes Negativ mit Fänger: Zeile 18 (Literale gleich dem Schlüssel) und der Rundlauf in `slice-v2-bootstrap`.
- Nebenläufigkeit (gleichzeitige Aufrufe des Lesers, Wächter): eine Zusage über Nebenläufigkeit gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird; Rate mit Zahl der Prozesse im Plan (`AGENTS.md` §3.10; Zeilen 32 und 33) — **Ausgang:** offen.
- Fakes: der Zähl-Fake des `FilterLeser` aus `slice-v2-query` bleibt dort; dieser Slice braucht keinen Port-Fake (Block `fakes: 0`) und nutzt ihn in Zeile 28 — **Stand Handoff:** `TestFlEinhaengung` ruft den Zähl-Fake mit dem echten Leser als `liefert`-Funktion und hält Argumente und Aufrufzahl fest; kein neuer Fake — **Ausgang:** offen bis zur Closure (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`).
- Die Berichtigung L6 in `SPEC-173` Nr. 5 muss vor dem ersten Code-Commit stehen — **Ausgang:** entfallen; sie steht seit 2026-11-01 in `SPEC-173` Nr. 5 und ist am 2026-11-05 um die Absätze *Der Leser*, *Rangfolge im Ausdruck*, *Weitere Konstrukte, die V2 nicht kennt* und *Literal-Lauf ohne Präfix* geschärft (Historie in `spec/spezifikation.md`).
- Größenmarke: der Block steht in §4 (Marke 4600 aus *S* = 3370 nach `SPEC-128`, Basis `80d7910c`, `schnitt-grund` gesetzt; kein Schnitt nach dem Ausgang des Architect) — **Ausgang:** akzeptiert für den Block (Planner); offen bleibt allein der Formelkonflikt `SPEC-128` (Marke 4600) gegen `AGENTS.md` §3.14 (Marke 3600, über 3000 mit `schnitt-grund` in §4 belassen); Adresse: User. Die Messung gegen die Marke folgt beim Handoff (`make groessenmarke`).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster steht unten, jede Zelle nennt ihren Ort oder „gilt nicht“ mit Grund; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect — **Stand Handoff:** keine Zelle ohne Ort gefunden; die Randformen sind je ein Fall der Tabelle (§3) — **Ausgang:** offen bis zur Closure; Adresse: Architect.
- Abweichungen der Tabelle *Zusage · Fänger* von der Spezifikation oder von sich selbst, vom Implementer beim Handoff gefunden (`AGENTS.md` §3.13) — **Ausgang:** entschieden (Architect, 2026-11-05; Entscheidung am Ende der Liste); Stand der Rückgabe des Implementers:
  - Zeile 17: hinter `X'…'` ergibt ein folgendes `,` `literal_typ`, nicht `wert`, denn der erste Verstoß in der Lesefolge entscheidet (`SPEC-173` Nr. 5, `SPEC-164` Nr. 6); der Test folgt der Spezifikation, die Zeile nennt den Fall.
  - Zeile 27: die dritte Stelle der Marker ist der Verstoß selbst statt „Operator“; am Platz eines Operators löst ein Marker nur die Ursache `wert` aus.
  - Zeile 28: bei doppeltem `$filter` wird der Leser für das erste gerufen, für das zweite nicht (wie `TestEinhaengungAufruf`); die Zeile sagt „für das zweite nicht“.
  - Zeile 13: ein Zusatzfall `Id eq1 1` (Menge 25); ohne rechten Operanden wäre ein angenommenes `eq1` nicht sichtbar.
  - Zeile 9: die Mutation „Normalisierung (NFC)“ ist nicht mit einer Bibliothek fahrbar, die Standardbibliothek hat keine; stattdessen „der Name wird klein geliefert“ (ergänzt, siehe Entscheidung unten).
  - Zeilen 26 und 27: „Aufruf und Pfad vertauscht“ (beide `filter`, nicht unterscheidbar) und „`Code()` vertauscht“ (der Code ist ein Wert des Modells, nicht dieses Slice) hat der Implementer nicht gefahren (entschieden unten).
  - Zeile 33: ein Fänger, der nur Texte liest, die ein früherer Aufruf schon gelesen hat, sieht einen Cache nicht (nur lesende Zugriffe); `TestFlGleichzeitig` liest deshalb zusätzlich Texte ohne frühere Lesung (Bezug aus dem Test).
- Ein V4-förmig beginnender, sonst ungültiger Lauf meldet `wert` statt `literal` (`SPEC-173` Nr. 5, Entscheidung des Architect 2026-11-05) — **Ausgang:** weiter offen als akzeptiertes Negativ: Grund: der Leser kennt V4 nicht und liest die Formen nicht ein zweites Mal; Folge: dieselbe Antwort `400` mit `ODR-E6005`, nur die Ursache im Log und im Fehlerwert unterscheidet sich; Fänger: Zeile 19.
- **Entscheidung des Architect zu den Abweichungen der Tabelle (2026-11-05):** Zeilen 17, 27 (Stelle „im Verstoß-Token selbst“), 28 und 13 sind bestätigt, die Tabelle trägt die Fassung des Implementers, die Spezifikation bleibt. Zeile 17: der erste Verstoß in der Lesefolge entscheidet, und hinter `X'…'` ist das Literal selbst der erste (`literal_typ`, `SPEC-173` Nr. 5 *Rangfolge im Ausdruck*); Zeile 27: am Platz eines Operators löst ein Marker nur `wert` aus, die Stelle im Verstoß trägt den Vertrag der Zeile; Zeile 28: ein zweites Vorkommen von `$filter` ist `option_doppelt` (`SPEC-162` Nr. 3 a), der Leser wird höchstens einmal gerufen (`SPEC-160` Nr. 8 a), also für das erste und nicht für das zweite; Zeile 13: der Zusatzfall `Id eq1 1` gilt (Menge 25). Zeile 9 ist berichtigt: die Mutation „NFC“ ist als Bibliotheksaufruf nicht fahrbar, aber ein Stellvertreter ohne Bibliothek (das zerlegte Paar der Fixture zum zusammengesetzten Zeichen zusammengesetzt) ist es; Zeile 9 der Tabelle nennt ihn, der Implementer fährt ihn in der Fix-Runde (grün wäre ein Befund). Zeile 26 (F-6): „Aufruf und Pfad vertauscht“ ist ein äquivalenter Mutant, akzeptiertes Negativ mit Grund und Folge in §3. Zeile 27 (F-6): „`Code()` vertauscht“ ist fahrbar und rot (Reviewer), die Zeile in §3 trägt sie. Folge für den Code: keine.

**Raster *Operation × Argumentart*** (Zellen aus dem Architect-Ausgang zu `slice-v2-query`, verfeinert vom Architect am 2026-11-05; „Nr. 5“ meint `SPEC-173` Nr. 5 mit dem genannten Absatz, „164“ die wortgleich geltende Nummer von `SPEC-164`):

| Operation | Ausdruck fehlt oder leer | unbekannter Operator, Funktion oder Pfad (außerhalb des Umfangs) | fremde Syntax (V4 am V2-Leser) | Literale in V2-Schreibweise, Typfehler | Wiederholung (doppeltes `$filter`), über Grenze (Länge, Tiefe) | Abbruch (Kontext) |
|---|---|---|---|---|---|---|
| Lesen des Ausdrucks (`NewFilterLeser`) | `$filter=` leer über `ParseURL`: `wert` (`SPEC-162` Nr. 8 mit `SPEC-173` Nr. 4), der Leser wird nicht gerufen (`SPEC-160` Nr. 8 a); unmittelbar gerufen: leerer Text `wert`, Leerraum am Rand `wert` (164 Nr. 1 und 2, Nr. 5 *Der Leser*) | Operator, Funktion (auch unbekannte), Pfad mit `/`: `filter` (Nr. 5, L1, L6; 164 Nr. 4 c 1, 5 a); unbekanntes Token am Platz eines Operators: `wert` (164 Nr. 4, letzter Absatz) | Literal ohne Präfix: `wert` (Guid, Datum, Zeitpunkt, Uhrzeit) oder `literal_typ` (`INF`, `-INF`, `NaN`) (Nr. 5 *Literal-Lauf ohne Präfix* b und c); `has in divby`, Minus, `$it`, Alias, qualifizierter Name: `filter` (Nr. 5, L6 und *Weitere Konstrukte*); JSON-Literale, doppelte Anführungszeichen: `wert` (164 *Ungeprüft*, Nr. 5 *Ungeprüft*); Kommentarformen: 164 Nr. 2 | `literal`, `literal_typ`, `zu_lang` nach Nr. 3 und Nr. 5 *Literal-Lauf ohne Präfix*; Typfehler gegen das Modell: gilt nicht, der Leser kennt das Modell nicht (164 Nr. 10 und *Ungeprüft*, Nr. 5 *Ungeprüft*) | doppeltes `$filter`: `option_doppelt` (`SPEC-162` Nr. 3 a, `SPEC-173` Nr. 4); Tiefe 32 `zu_viele`, Literal 512 Bytes `zu_lang`, Name 128 Zeichen `wert`, Text höchstens 8192 Bytes (164 Nr. 7, Nr. 5 *Der Leser*; `SPEC-160` Nr. 6); gleicher Text: 164 Nr. 8 | gilt nicht: der Leser hat keinen `ctx` (164 Nr. 1, Nr. 5 *Der Leser*) |
| Einhängung mit echtem Leser (`WithFilter`) | kein Leser: `filter` (`SPEC-160` Nr. 8); `$filter=` leer: `wert` ohne Aufruf (`SPEC-162` Nr. 8, `SPEC-160` Nr. 8 a) | Leser liefert `*KompatFehler` mit `filter`: Ursache und Ort `query/<n>` gelten (`SPEC-173` Nr. 1 c) | Fehler anderer Art oder Klasse (`*URLFehler` `nicht_unterstuetzt`, `*KompatFehler` mit fremder Ursache, beide Typen, Zeiger `nil`, weder Ausdruck noch Fehler): Fehler des Aufrufers `model.ErrUngueltig` (`SPEC-173` Nr. 1 d) | `literal_typ` des Lesers gilt mit Ort (`SPEC-173` Nr. 1 c); Fehler der Klasse `ungueltig` gilt mit Ort | höchstens ein Aufruf je `ParseURL`, nicht mehr nach einem Verstoß der Klasse `ungueltig` davor (`SPEC-160` Nr. 8 a) | gilt nicht: `ParseURL` hat keinen `ctx` (`SPEC-160` Nr. 1) |
| Kompatibilitätsfehler im Ausdruck | gilt nicht: er hat immer eine Ursache (`SPEC-174` Nr. 2) | die Menge: `SPEC-170` Nr. 3 und Nr. 5 *Weitere Konstrukte* | `SPEC-173` Nr. 5 (*Literal-Lauf ohne Präfix*, *Weitere Konstrukte*) | `literal_typ` (`SPEC-174` Nr. 2; Nr. 3 und Nr. 5) | Rangfolge im Ausdruck: der erste Verstoß in der Lesefolge, gleich welcher Klasse (Nr. 5 *Rangfolge im Ausdruck*); zwischen den Optionen `SPEC-174` Nr. 4 und `SPEC-173` Nr. 6 | gilt nicht (wie oben) |

**Raster *Menge der Ausdrücke*, *Namen*, *Minus-Läufe*** (Prüfgrenzen als akzeptierte Negative, Zeilen 9, 12, 36):

| Menge | Entscheidung (Ort) | Prüfgrenze und Folge |
|---|---|---|
| Namen (Zeichen der Kategorien `L` und `Nd`) | 164 Nr. 4 c 7 mit `SPEC-125` Nr. 8 b | Prädikat `unicode.Is` ist Orakel und Code zugleich; ein Zeichen, das die Bibliothek anders einordnet als die Spezifikation, wird gleich eingeordnet (Folge: `wert` oder Name an einer Stelle, die die Spezifikation anders sähe) |
| Läufe mit `-` | 164 Nr. 4 b; Nr. 5 *Literal-Lauf ohne Präfix* und *Weitere Konstrukte* | die Tabelle nennt 35 Formen, Zeile 36 mischt weitere; Folge: eine Form hinter `-`, die keine der beiden trifft, kann falsch eingeordnet sein (`filter` oder `literal`, beide Ausgänge sind Fehler) |
| Ausdrücke (Zeile 36) | 164 Nr. 3 und 4; Nr. 5 | 2000 Läufe, Saat 1 bis 2000; Folge: eine Kombination, die keine Saat trifft, bleibt ungeprüft |

**Zusatztabelle** (je Paar von Verstößen am selben Ausdruck und je Modus; `AGENTS.md` §3.16):

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| Kompatibilität und `400` im selben Ausdruck | der erste Verstoß in der Lesefolge gilt (Nr. 5 *Rangfolge im Ausdruck*, 164 Nr. 6); `SPEC-173` Nr. 1 und 6 gelten zwischen den Optionen | Zeile 26 (Paare über 7 Verstoßformen) |
| zwei Kompatibilitätsfehler im Ausdruck (`filter` und `literal_typ`, Aufruf und Pfad) | der erste in der Lesefolge (164 Nr. 6) | Zeile 26 |
| `wert` und `literal` am selben V4-förmigen Lauf | Form von V4 geht vor: `wert` (Nr. 5 *Literal-Lauf ohne Präfix* c, d) | Zeile 19 |
| Länge und Form desselben Literals | Länge zuerst, `zu_lang` (164 Nr. 4 a, 6; Nr. 5 *Literal-Lauf ohne Präfix*) | Zeile 15 |
| `kodierung` und Grammatik (nur bei unmittelbarem Aufruf) | die Kodierung geht der Grammatik vor (164 Nr. 1, Nr. 5 *Der Leser*) | Zeile 27 |
| Fehler des Lesers gegen eine spätere Option der Query | `SPEC-160` Nr. 5 mit `SPEC-173` Nr. 6 | Zeile 28 |
| mit Leser gegen ohne Leser (Modus der Einhängestelle) | `SPEC-160` Nr. 8 | Fänger in `slice-v2-query` (ohne Leser), hier Zeile 28 (mit Leser) |
| Pfad im Filter gegen Pfad in `$select` und `$orderby` | im Filter `filter`, in den beiden anderen `option` (`SPEC-173` Nr. 1 und 4) | Zeile 22 (über `ParseURL` mit `$select=a/b`) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-01); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (Schwelle erreicht, Verkörperung offen) gilt für Fakes; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (Schwelle erreicht, Zustand offen) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nebenläufig in zwei Zeilen der Tabelle (§3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.

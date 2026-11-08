# Verifikation: slice-odata-url-filter — 2026-10-31

**Rolle:** Verifier (Modul 11) — "Bauen wir es richtig?" gegen Plan, DoD, `SPEC-164`, `SPEC-165` und `AGENTS.md`. Kein Review, keine Reparatur.

**Gegenstand:** `slice-odata-url-filter` (Welle `welle-mock`), HEAD `7ca6071`; Implementer `84b24a5`, Review `a28412e` (F-1 bis F-13), Spec- und Plan-Nachzug `3ce36c0` und `6ead1a6`, Fix-Runde `7ca6071`. Pfade: `src/adapters/driven/odata/urlparser/`, `src/hexagon/model/filter_ast.go`, `spec/spezifikation.md` (`SPEC-164`, `SPEC-165`), Plan `docs/plan/planning/done/slice-odata-url-filter.md`.

**Modell:** claude-sonnet-5-5. **Messverfahren:** Docker, `--memory=4g --memory-swap=4g --cpus=4 --network none`, unter `flock /tmp/odata-recorder-heavy.lock`; Exit-Codes aus Datei; Mutationen und Probe-Pakete in Kopien unter `/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verify-urlfilter/` (`mm/<Kennung>`, `new`, `old`, `m29`, `m30`), Worktree-Code unverändert (`git status` leer vor und nach den Läufen).

## Urteil

**Bestanden, ohne Blocker.** Kein HIGH, kein falsches Codeverhalten, keine Abweichung vom Wortlaut von `SPEC-164` und `SPEC-165` gefunden. Die Fix-Runde deckt die Klassen von F-2 bis F-5, F-7 bis F-10 und F-13: an je mindestens einem Element, das weder Befund noch Implementer nannten, ist die Mutation rot. Zwei Reste bleiben als Fänger-Lücken ohne Codefolge (MEDIUM, LOW), dazu eine Plan-Unstimmigkeit (LOW) und Hinweise an den Planner.

## Stufe (g): `make gates`

Im Worktree `make gates` gelaufen: **Exit 0** (Datei `gates.exit` im Scratchpad, Inhalt `0`). Darin `d-check: 532 Datei(en) geprüft, 0 Befund(e)`, `migration-check: ok`, `zeichen: ok`, `groessenmarke slice-odata-url-filter.md: ok, Messung 2334 von Marke 4700`, die Stufe `race` wurde gebaut. Der Worktree blieb sauber (kein Stempel).

## Stufe (f): Größenmarke

Marke **4700** ist fest: Block `marke: 4700`, `schaetzung: 3460`, `basis: 65f75c8`; der Plan nennt in §4 die Rechnung S = 85·2 + 80·33 + 650 = 3460, Marke = 100·⌈135·3460/10000⌉ = 4700 und führt den Nachzug der 8 Zeilen ausdrücklich ohne neue Marke (`SPEC-128` Nr. 5 c). Messung nachgerechnet: `git diff --numstat 65f75c8..HEAD` über die drei Pfade des Blocks = **2334 Einfügungen, 0 Löschungen**, Summe 2334 (Produktcode 497 = 311 + 113 + 73, Tests 1837); gleich dem Plan und gleich der Ausgabe des Gates. 2334 liegt unter 4700, ein Eintrag `ueberschreitung` entfällt.

## Stufe (a): Tabelle gegen den Vertrag

43 Zeilen gezählt (35 beim Start plus 11a, 11b, 14a, 14b, 15a bis 15d); Klasse `nebenläufig` allein bei 29 und 30, 41 `andere`. Mengen nachgezählt: Z4 (21 + 8 je zwei Schreibfolgen = 58), Z5 (16 + 2), Z7 (108 + 84 + 4: Bitmasken der zwölf Wörter ergeben 4+4+16+8+4+8+16+32+16 = 108, die acht Wörter von Nr. 5 b 84), Z10 (58 + 8), Z11 (189 + 6), Z12 (24), Z13 (60 + 12), Z27 (24 Texte einzeln gezählt) stimmen mit der Rechnung; die Tests brechen bei anderer Zahl ab (`n != 245`, `n != 77`).

Jede Zusage von `SPEC-164` Nr. 1 bis 9 und `SPEC-165` Nr. 1 bis 3 hat eine Zeile: Nr. 1 (Z1, Z24), Nr. 2 (Z11 bis Z13, Z11a, Z16), Nr. 3 (Z2 bis Z7), Nr. 4 a (Z14, Z14a, Z14b, Z16), 4 b (Z15, Z15a bis Z15d, Z11b), 4 c (Z10, Z17, Z19 bis Z21), 4 d (Z9), 4 e und Operatorplatz (Z8, Z27), Nr. 5 (Z19 bis Z21), Nr. 6 (Z23, Z24, Z35), Nr. 7 (Z10, Z14, Z22, Z31, Z32 als Negativ), Nr. 8 (Z28 bis Z30), Nr. 9 (Z6, Z7, Z12, Z26), `SPEC-165` (Z33, Fänger `TestFilterNurWerte`, `TestFilterBaumNurWerte`). Nr. 10 liest der Leser nicht (Z18). **Zwei Lücken:**

- **MEDIUM — Zusage ohne Fänger (`AGENTS.md` §3.15):** der Schlusssatz von `SPEC-164` Nr. 4, "ein Operatorwort unmittelbar vor `'` ist ein Literal mit Präfix und deshalb am Platz eines Operators `wert`" (Raster O5 nennt ihn), hat keine Zeile und keinen Test. Mutation MV (der Zerleger behandelt `and` vor `'` nicht als Präfix; `Id eq 1 and'x'` wird gelesen) und MV2 (dasselbe für alle 16 Operatorwörter von Nr. 4) bleiben **grün**. Das Verhalten des Codes ist richtig (13 Eingaben von Hand gelesen: `Id eq'x'`, `Id and'x'`, `Id eq 1 AND'x'`, `Id has'x'` sind `wert`; am Platz eines Operanden `not'x'`, `and'x'` sind `literal_typ`). Folge: eine Regression im Zerleger bleibt bei grünen Gates unbemerkt.
- **LOW — Zeile 11b:** die Menge "jeder Lauf mit `-` am Anfang, der kein Literal-Lauf ist" ist unbeschränkt, die Tabelle führt 8 Läufe ohne Prüfgrenze mit Grund und Folge (`AGENTS.md` §3.17). Mutationen an Elementen außerhalb der 8 bleiben grün: MT (`-` und danach ein Byte jenseits von ASCII wird `wert`), MT3 (Lauf `-$…` wird `wert`), MT4 (Läufe ab 6 Bytes werden `wert`). Das Verhalten ist richtig (`laufOperand` kennt nur `c == '-'`); der Fänger ist eine Auswahl. Folge: eine zeichenabhängige Sonderbehandlung nach `-` bliebe unbemerkt.

Plan-Unstimmigkeit (LOW, Doku): Zeile 22 der Tabelle nennt "Menge 75 = … (64 + 8 + 4 + 1)" und danach "Menge 77"; der Test sagt 77. Der Satz trägt die alte Zahl.

## Stufe (b): Fix-Runde je Befund gegen die Klasse (`AGENTS.md` §3.17)

Mutationen an Kopien, `go test -count=1` über `urlparser` und `model`, "rot" nur mit benanntem Test, kein Aufbaufehler (Spalte Aufbau 0).

| Befund | Klasse (Menge) | Mutation an einem Element, das weder Befund noch Implementer nannten | roter Test |
|---|---|---|---|
| F-2 Paare am selben Token | Paare (Aufruf, Fehler im Lauf) über Anfänge `$`/`@`, Pfad, qualifiziert | MA: Aufruf nicht vor dem Sonderfall `$`/`@`; MB: Aufruf nicht für Läufe mit `/`; MC: nicht für Läufe mit `.` | MA `TestAufrufAnfaenge`; MB und MC `TestAufruf`, `TestAufrufAnfaenge`, `TestAufrufVorUngueltigemZeichen` |
| F-3 Lauf-Anfänge vor `(` | 7 Anfänge × 7 Reste × 5 Formen = 245 | wie MA bis MC | `TestAufrufAnfaenge` |
| F-4 Tiefe nach der Gruppe | Art außen × 32 Tiefen × 4 Kombinationen × 2 Seiten = 512 | MF: nach `)` ist die Tiefe `klammern + 1`; MF2: nach `)` ist die Tiefe `klammern` (äußeres `not` vergessen) | `TestTiefeNachGruppe` (beide) |
| F-5 Gestalt der Knoten | 6 Typen, 40 Bäume | MI: Slice in `FilterVerknuepfung`; MJ: Zeiger im gemeinsamen Typ `Literal`; MK1 bis MK4: Zeiger statt Wert für `and`, `null`, Eigenschaft, `not` | `TestFilterNurWerte`, `TestFilterBaumNurWerte` (alle sechs) |
| F-7 `nurNamenzeichen` entfernt | je Glied und je Name | ML (Glieder von `.` ungeprüft), ML2 (letztes Glied ungeprüft), ML3 (Glieder von `/` ungeprüft, außer dem letzten), ML4 (Alias ohne Namensprüfung), ML5 (Name ungeprüft) | `TestNichtUnterstuetzt`, `TestZeichenAussenGrammatik`, `TestPfad`, `TestPfadGlieder`, `TestNamen` u. a. |
| F-9 Pfad-Glieder | 2 bis 64, 4096, 6 Stellen | MN: ab 16 Gliedern `wert`; MN2: genau bei 40 Gliedern `wert` | `TestPfadGlieder` (beide) |
| F-13 Zustand im Paket | 24 Vorläufe × 64 Texte = 1536 | MP: Tiefe als Paketvariable, im Fehlerpfad nicht zurückgesetzt | `TestFilterReihenfolge` (daneben 16 weitere) |
| F-8 Parallelität | siehe Stufe (d) | – | – |
| F-10 Kommentar | Text auf das Geprüfte gekürzt | – | gelesen: `filter_token.go` nennt den Fänger `TestFilterWaechter` und nennt die Laufzeit ohne Zuteilung als ungeprüft (Zeile 32) |

Dazu für F-1 (Länge vor Form) vier Mutationen an Elementen außerhalb der Befund-Formen: MQ (ungeschlossen vor Länge, bis 600 Bytes), MQ2 (Rest hinter `'` vor Länge), MQ3 (Länge nur für geschlossene), MQ4 (Länge nur bei gültigem Ende) — alle rot in `TestLaengeVorForm` bzw. `TestRestZaehltNicht`.

## Stufe (c): Guid-Form, Länge vor Form, `--` nach dem Spec-Wortlaut

Probe-Paket aus Generatoren (`zz_probe_test.go` im Scratchpad, nicht im Worktree):

- **Gegen den Schlüssel:** 167634 Literal-Läufe der Klasse von `SPEC-164` Nr. 4 a und b (mutierte Guids mit allen 22 Anfangszeichen, Zahlen, Daten, Zeitpunkte, Präfix-Zeichenketten der Längen 0 bis 600 mit und ohne Rest) als `Id eq <Lauf>` gegen `ParseURL("/Set(Id=<Lauf>)")`: **1641 Abweichungen, alle `literal` im Filter gegen `zu_lang` im Schlüssel**, und alle erklärt durch die Regel "der Rest hinter dem schließenden `'` zählt nicht zur Länge" (`SPEC-164` Nr. 4 a, vom Test mit eigenem Prädikat geprüft: geschlossen, Länge bis zum `'` höchstens 512, Rest dahinter). Sonst gleiche Art, gleicher kanonischer Text, gleiche Ursache.
- **Tabelle aus dem Wortlaut** (50 Eingaben): `--` am Operandenplatz `filter` an allen Stellungen, am Operatorplatz `wert`; `//`, `/*`, `#`, `;`, Rückstrich an beiden Plätzen `wert`; `---`, `-x`, `-(5)`, `--5` `filter`; `-5`, `-.5`, `-INF` Literal-Läufe; Guid mit Buchstaben am Anfang klein und groß `LiteralGUID`, 35 Zeichen mit Buchstaben `wert`, mit Ziffer `literal`, falscher Trenner `wert`, `g` im Lauf `wert`, 32 Hex ohne `-` ein Name, vor `(` `wert`, vor `'` `literal`, ohne Leerraum nach `eq` `wert`; `a#(1)` `filter`, `#(1)` `wert`. Alle erfüllt (eine Erwartung von mir, `-.5` als gültige Zahl, war falsch; `-.5` ist ein Literal-Lauf mit der Ursache des Schlüssels, wie Z11b es führt).
- **Zeile 34 mit eigener Saat:** `TestGeneratorFilter` mit den Saaten 100000 bis 101999 (2000 Läufe, Orakel der Renderer des Tests): grün.

## Stufe (d): Raten von Zeile 29 und 30 mit Bedingung

Testprogramm einmal übersetzt, je Zelle 20 Prozesse nacheinander, `-test.run '^TestFilterGleichzeitig$' -test.count=1`, Container mit `--cpus=4`; M29 (Tiefenzähler als Feld der Instanz, selbst gebaut), M30 (Cache `map` als Paketvariable, selbst gebaut), Detektor mit cgo im Image der Stufe `race`:

| Fassung | ohne Detektor: Standard · `-test.cpu=1` · `-test.cpu=2` | mit Detektor: Standard · `-test.cpu=1` · `-test.cpu=2` |
|---|---|---|
| unverändert | 0 · 0 · 0 von 20 | 0 · 0 · 0 von 20 |
| M29 | 20 · 20 · 20 von 20 | 20 · 20 · 20 von 20 |
| M30 | 20 · 20 · 20 von 20 | 20 · 20 · 20 von 20 |

Das stimmt mit dem Plan überein (20 von 20, unverändert 0 von 10). Die Bedingung (Anhebung von `GOMAXPROCS` auf mindestens 4 im Test, `-test.cpu=1` ergibt trotzdem 20 von 20) ist im Plan genannt; ungemessen bleibt ein Host mit einem Kern, wie der Plan sagt.

## Stufe (e): Entfernung von `nurNamenzeichen`

Verhalten unverändert. Der Beleg ist ein Differenzialtest: die Fassung von `HEAD` mit wiederhergestelltem Fall `!nurNamenzeichen(s)` gegen `HEAD`, 400000 Eingaben aus 100 Atomen (Operatorwörter, Namen, Pfade, qualifizierte Namen, `$`, `@`, `#`, NBSP, BOM, Steuerzeichen, Guids, Literale, rohe Bytes; Saat 100000), Ausgabe je Eingabe Ursache oder `%#v` des Baums: **`cmp` identisch**, 400000 Zeilen (rund 12000 gelesene Bäume, sonst `wert` 245000, `filter` 49000, `kodierung` 41000, `literal_typ` 23000, `literal` 23000, `wert_nicht_unterstuetzt` 7000).

## Stufe (h): DoD, `New()`, §6

- Die vier Haken der DoD (Grammatik und Rangfolge, Operanden und Baum, Fehlerklassen, `make gates`) sind wahr (Stufen a bis g). Die übrigen Punkte stehen offen, wie sie sollen (Closure). **Hinweis an den Planner:** "Review durchgeführt" ist offen, obwohl der Report `docs/reviews/2026-10-31-slice-odata-url-filter-review.md` vorliegt; der Haken gehört in die Closure.
- `New()` bleibt ohne Leser: `NewFilterLeser` hat in `src/` nur die Definition, keinen Aufruf außerhalb der Tests, `src/bootstrap` ruft `WithFilter` nicht.
- §6: die Zeilen mit Ausgang *offen* (Sendung an `slice-bootstrap-m3`, Sendung an `slice-edm-ausdruecke`, Wortlaut von `AGENTS.md` §3.14 gegen `SPEC-128`) tragen Adresse und Nehmer. **Hinweis an den Planner:** die Einträge in den Nehmern stehen bereits in `done/slice-bootstrap-m3.md` (Übernimmt-Absatz, Verdrahtung in der DoD) und `done/slice-edm-ausdruecke.md` (Absatz, DoD, Randformen), vermutlich über den Merge von `main`; der Ausgang der beiden Zeilen ist bei der Closure zu setzen. Die Diff-Vorlage an den User zu `AGENTS.md` §3.14 steht noch aus. Die Zeilen von F-1, F-6, F-12 und die Guid-Rückgabe tragen ihren Ausgang *entschieden* mit Ort in der Spezifikation.

## Für den Planner

1. MEDIUM: Zeile für den Schlusssatz von `SPEC-164` Nr. 4 (Operatorwort vor `'`) mit Fänger und Mutation, oder benanntes Negativ mit Grund und Folge; kein Codefehler.
2. LOW: Zeile 11b mit Prüfgrenze für die Menge hinter dem `-` (Grund, Folge) oder Generator über das zweite Zeichen.
3. LOW: "Menge 75" in Zeile 22 auf 77 berichtigen.
4. Closure: Haken "Review durchgeführt", Ausgang der beiden Sendungen, Diff-Vorlage zu `AGENTS.md` §3.14.

LH-Bezug: [LH-FA-08, LH-QA-05](../../spec/lastenheft.md).

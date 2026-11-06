# Review-Bericht: slice-query-interpreter — 2026-11-06

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-query-interpreter`, Zweig `slice-query-interpreter`, HEAD `a37fb08`, Diff `055c295..a37fb08` (12 Dateien, 2669 Einfügungen, 6 Löschungen; Code und Tests des Implementers bis `1772e28`, Vertrag des Architect `a37fb08`). Bezug: `LH-FA-08`, `LH-FA-11`, `LH-QA-02`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md).

**Geprüft gegen:** Plan `slice-query-interpreter` (§3 Tabelle mit 34 Zeilen, Mengen-Tabelle mit 9 Mengen, Raten der Zeile 32, §4 Messung 2588 gegen Marke 4500, §6 mit Raster), Vertrag `SPEC-190` (Eingabe, Ergebnis, Programmfehler, Fehler der Anfrage, Rangfolge, Wiederholung, Grenzen, *Ungeprüft*) und `SPEC-191` (Stufen, Filter, Ordnung Nr. 3 d und e, Paging, Projektion, Teilprojektion), `SPEC-166`, `SPEC-167`, `SPEC-125` Nr. 10 und 11, `AGENTS.md` §3.7, §3.9, §3.10, §3.11, §3.13 bis §3.18; Skill `.harness/skills/reviewer.md` in der Fassung des Zweigs (nach Merge von `main`).

**Methode.** Einmal gefahren (Durchstich vor Härtung, keine Wiederholungsschleife). Der Bericht des Implementers ist nicht übernommen. Gemessen mit `tools/harness/mutationslauf.sh` (SPEC-210) auf einer Wegwerfkopie des Zweigs im Scratchpad (`rev-qi/baum`, ohne `.git`; dazu eine Hilfsdatei mit einer Zählervariable im Paket `model`, nicht im Repo), Image `odata-recorder:race`, unter `flock /tmp/odata-recorder-heavy.lock`, Docker `--memory 2g --cpus=4` (Vorgabe des Werkzeugs), Last des Hosts 4 bis 11 bei 20 CPUs, ein schwerer Lauf zugleich. Alle Mutationen sind eigene Elemente; die Listen `m1.json` und `m2.json` des Implementers sind nicht gefahren, nur gelesen. Rohbytes (§3.18): in allen neuen Dateien des Diffs 0 Treffer für NBSP, BOM, Nullbreitenzeichen, Zeilentrenner und die Folge Backslash-`u`-Hex (`LC_ALL=C grep -cP`).

## Befund zuerst: was die Zeile 32 hält

Die Zähler in `abfrage*.go` sind lückenlos gefangen. Ich habe **48 eigene Zähler** gesetzt (Schreiben auf eine Paketvariable, Klasse `nebenläufig`, Detektor `-race`, `-test.count=1`, **20 Prozesse je Zähler**, Last 4 bis 11, Container 4 CPUs und 2g): in jeder der 20 genannten Funktionen (mindestens ein Zähler je Funktion) und in Zweigen, die der Plan nicht nennt (Schleifenrümpfe von `pruefeNamen` für `$select` und `$orderby`, `verlangeProperties` im Fehlerzweig, `wertVon` gefunden und nicht gefunden, jeder Programmfehlerzweig von `pruefeArgumente`, `hatDoppelte` beide Rückgaben, `pruefeEntities` beide Fehlerzweige, `seite` in `$top`, `projiziere` in Map und Anhängen, Ende von `FuehreAus`, `filterNamen` für `not` und Ende, `filtere` ohne Filter und Anhängen, `filterWert` für Literal und `null`, `verknuepfeWerte` für `or`, `wahrheitAlsWert` für `false`, `operanden2` Ende, `pruefeName` in drei Zweigen, `ordne` in beiden Schleifen, im Rest und beim Index, `baueSpalte` für `null`, Klassenfehler, Guid und Ordnung auf Guid, `vergleiche` für Guid und `desc`, `klasse()` in drei Zweigen, `Code()` in zwei, `Error()` in zwei, `Is`, `abfrageVerstoss`, `AlleAbfrageUrsachen`): **jeder 20 von 20 rot** (`mit-detektor=20/20`), keiner grün, keiner flatternd. Die Funktionen `klasse`, `Code`, `Error`, `Is`, `abfrageVerstoss` und `AlleAbfrageUrsachen`, die der Plan als „kein gemeinsamer Zustand“ ohne Zähler führt, sind damit mitgemessen.

Messhinweis für spätere Läufe: Die Zuweisung einer Variablen an sich selbst (`x = x`) streicht der Compiler; mein erster Versuch lag deshalb bei 0 von 20 (2 Zähler, grün) und ist verworfen. Gezählt ist erst der echte Schreibzugriff (`zaehler++`).

Nicht aufgelöst bei Zeile 32: die drei Zweige, die kein Aufruf erreicht (`Property der Entity fehlt`, `Knoten des Filters`, `Operator der Verknüpfung`); sie stehen im Plan als akzeptiertes Negativ mit Folge, ich habe sie nicht gezählt.

## Eigene Mutationen (Wegwerfkopie, `abfrage*.go`)

`rot` heißt: der benannte Test endet mit Exit-Code ungleich 0; `rot-paket`: ein anderer Test des Pakets. `grün`: das ganze Paket besteht. Kontrollen (`V11`, `V24`, `W4`) sind rot, das Werkzeug misst also.

| Nr. | Mutation (einzelne Änderung) | Ergebnis |
|---|---|---|
| V01 | Guid als Schlüssel: nur die letzte Gruppe (`Text[24:]`) verglichen | grün |
| V02 | Guid als Schlüssel: Hex-Buchstaben `a` bis `f` vor den Ziffern | grün |
| V03 | Rest der Ordnung: erst die Länge von `Schluessel`, dann der Text | grün |
| V04 | Rest der Ordnung: `strings.ToLower` auf beiden Seiten | grün |
| D3 | Rest der Ordnung: führende `0` abgeschnitten | grün |
| D4 | Rest der Ordnung: absteigend, wenn das letzte Element von `OrderBy` absteigend ist | grün |
| W4 | Rest der Ordnung entfällt (`return 0`) — Kontrolle | rot (`TestSchluesselRest`) |
| V05 | `ordne` überliest das Element mit Index 3 von `OrderBy` | grün |
| V06 | `ordne` überliest das Element mit Index 16 von `OrderBy` | grün |
| V07 | `ordne` schneidet `treffer` auf 10.000 ab | grün |
| V08 | `seite` rechnet mit `min(n, 10000)` | grün |
| V09 | `Anzahl` ist `min(len(geordnet), 10000)` | grün |
| V10 | `pruefeEntities` prüft nur die ersten 10.000 Entities | grün |
| V11 | `filtere` liest nur die ersten 10.000 Entities — Kontrolle | rot (`TestGrosseMenge`) |
| V12 | `projiziere` reicht ab der zweiten Entity `Properties` der Eingabe durch (ohne `$select`) | grün |
| V13 | `pruefeName` kennt nur die erste Navigation-Property | grün |
| V14 | `HatOrdnung` auch für `$select` gefragt | rot-paket (`TestAbfrageMaskeAufrufe`; `TestNamen` bleibt grün) |
| V15 | Stufe 5 prüft `$orderby` nicht mehr, wenn ein Filter gesetzt ist | grün |
| V16 | `baueSpalte` prüft die Klasse nicht für den Index 7 | grün |
| V17 | Key-Spalte vom Typ `Edm.Int64` nach dem Text geordnet | grün |
| V18 | Key-Spalte vom Typ `Edm.DateTimeOffset` nach dem Text geordnet | grün |
| V19 | Key-Spalte vom Typ `Edm.Decimal` nach dem Text geordnet | grün |
| V20 | Key-Spalte `Edm.Int32` mit negativem Wert nach dem Text geordnet | grün |
| V21 | `projiziere` liest nur die ersten vier Namen von `Select.Namen` | grün |
| V22 | Stufe 1 prüft nur die erste Key-Property | grün |
| D2 | Stufe 1 prüft nur die ersten zwei Key-Properties | grün |
| D1 | `ordne` liest nur die ersten zwei Key-Properties | grün |
| V23 | Stufe 1 merkt nur die ersten drei Schlüssel (Doppelte später unentdeckt) | grün |
| V24 | leeres Ergebnis mit `Entities` gleich `nil` — Kontrolle | rot (`TestEntitiesNieNil`) |
| W1 | `baueSpalte` verschluckt den Fehler von `pruefeWert` (`redigiert`, unbekannte Art in `$orderby`) | grün |
| W2 | `baueSpalte` verschluckt „Klassen nicht vergleichbar“ | grün |
| W3 | `baueSpalte` verschluckt „Ordnung auf Guid“ | grün |
| W5 | `ordne` überliest den Fehler einer Key-Spalte | grün |
| W6 | `ordne` überliest den Fehler einer `$orderby`-Spalte | grün |
| W7 | `FuehreAus` verschluckt jeden Fehler von `ordne` | grün |
| S1 | leeres Ergebnis liefert eine Paketvariable (gemeinsame Liste, Kapazität 4) für alle Aufrufe | grün; Detektor 0 von 20, Wirkung 0 von 20 |
| S2 | `AlleAbfrageUrsachen` liefert eine Paketvariable | grün; Detektor 0 von 20, Wirkung 0 von 20 |
| S3 | `projiziere` hängt in einen gemeinsamen Puffer an (Kontrolle, entspricht `S3puffer`) | flatternd: Detektor 20 von 20, Wirkung 0 von 20 |

---

## Findings nach Schwere

**HIGH: keine.** Kein Verstoß gegen eine ADR: keine Datei der Gate-Kette berührt (`Makefile`, `harness/mk/`, `Dockerfile`, `.a-check.yml`, `harness/README.md` stehen nicht im Diff, die Kante `tests` → `model` ([ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)) unverändert), `grep nolint` über `src/` und `tests/` ohne Treffer, keine Importe außer `slices` und `strings` im Modell. Kein Geheimnis in einer Senke: der `AbfrageFehler` trägt nie Name, Entity oder Wert (Test mit `GEHEIM`, `TestAbfrageFehler`). Kein Spec-Straten-Widerspruch gefunden zwischen `SPEC-190`, `SPEC-191` und dem Lastenheft; `SPEC-191` Nr. 3 d und e sind miteinander und mit `SPEC-190` *Ungeprüft* stimmig.

#### M-1 (MEDIUM): Programmfehler der Ordnung (`SPEC-190` Nr. 3 g) haben keinen Fänger

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.16; `SPEC-190` Nr. 3 g (Programmfehler beim Auswerten: `redigiert`, nicht vergleichbare Klassen, Form, die kein Leser liefert) und Nr. 5 Stufe 6
- `pfad`: `src/hexagon/model/abfrage_ordnung.go` · `return spalte{}, formFehler("Klassen nicht vergleichbar")` und `return spalte{}, formFehler("Ordnung auf Guid")` und `k, err := pruefeWert(w)`; Test `src/hexagon/model/abfrage_fehler_test.go` · `func TestAuswertungProgrammfehler`
- `befund`: `TestAuswertungProgrammfehler` führt nur Filter-Bäume. Die Ordnungsfälle (`Ordnung Klasse`, `Ordnung Guid`, `Ordnung redigiert`, `Key Klasse`) stehen nur in `abAbfragen`, und dort vergleicht `TestAbfrageGleichzeitig` das Ergebnis mit dem sequentiellen Lauf desselben Codes, `TestEingabeUnveraendert` verwirft den Fehler (`_, _ =`). Sechs Mutationen bleiben grün, darunter `W7`, bei der `FuehreAus` **jeden** Fehler von `ordne` verschluckt und eine Eingabe mit `redigiert` in `$orderby` still geordnet liefert (`W1` bis `W3`, `W5` bis `W7`).
- `verifizierbar`: ja (W1, W2, W3, W5, W6, W7)
- `klasse`: Programmfehlerpfad einer Stufe ohne Assertion (nur der Filterpfad gefangen)

#### M-2 (MEDIUM): Die Prüfgrenze 10.000 und 10.001 gilt nur vor dem Filter

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (beide Seiten jeder Grenze), §3.17 (unbeschränkte Menge: Prüfgrenze mit Grund und Folge); `SPEC-190` Nr. 7 (*„die Tests führen 10.000 Entities … und 10.001 aus (beide Seiten der Zahl, damit eine eingebaute Grenze rot wird)“*)
- `pfad`: `src/hexagon/model/abfrage_test.go` · `r := abOk(t, "große Menge", typ, es, model.Optionen{Filter: abF(abIDGt("100")), OrderBy: ..., Skip: abZ(5), Top: abZ(50)}, nil)`
- `befund`: Der Filter `Id gt 100` lässt 9.900 bzw. 9.901 Entities durch, `$top=50` begrenzt die Seite. Ordnung, Paging, Projektion und `Anzahl` sehen also nie 10.000 oder 10.001. Eine eingebaute Grenze in `ordne` (`V07`), in `seite` (`V08`), in `Anzahl` (`V09`) und in der Stufe-1-Prüfung (`V10`) bleibt grün; nur die in `filtere` (`V11`) wird rot. Ein Lauf ohne `$top` und ohne Filter über 10.001 Entities (die Antwort der Größe `LH-QA-01.a`) kommt im Test nicht vor.
- `verifizierbar`: ja (V07 bis V10)
- `klasse`: Prüfgrenze nur an einer Stufe der Pipeline gemessen

#### M-3 (MEDIUM): Die letzte Stufe der Ordnung (Text von `Schluessel`, `SPEC-191` Nr. 3 e) hat nur den Fall `a`/`b` und `1`/`2`

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fixture gleicht nie dem Wert der Mutation; Zusage ohne Zeile ist ein MEDIUM); `SPEC-191` Nr. 3 e (*„bytegleich aufsteigend“*, Folge: *„bei `Keys` leer steht `10` vor `2` (Text, nicht Wert)“*)
- `pfad`: `src/hexagon/model/abfrage_ordnung_test.go` · `func TestSchluesselRest` und `a.Schluessel, b.Schluessel = "b", "a"`; Plan `docs/plan/planning/done/slice-query-interpreter.md` · `Er hat keine Zeile der Tabelle`
- `befund`: Alle vier Eingaben des Fängers haben gleich lange, einheitlich kleingeschriebene Schlüssel. Länge zuerst (`V03`), Kleinschreibung gefaltet (`V04`), führende Nullen weg (`D3`) und absteigend bei absteigendem letztem `OrderBy` (`D4`) bleiben grün; die im Vertrag benannte Folge `10` vor `2` und „bytegleich“ sind ungefangen. Die Zusage hat im Plan keine Zeile der Tabelle *Zusage · Fänger* (§6 benennt das selbst).
- `verifizierbar`: ja (V03, V04, D3, D4)
- `klasse`: Fixture gleicht dem Wert der Mutation, Zusage ohne Zeile

#### M-4 (MEDIUM): Der Schlüssel als Spalte der Ordnung: Typen und Länge ohne Achsenliste

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (ein Fall je Element einer Menge), §3.17 (Menge mit Achsen); `SPEC-191` Nr. 3 d (*„Typen mit Ordnung nach `OrdneWerte`“*, Folge von `typ.Keys`)
- `pfad`: `src/hexagon/model/abfrage_ordnung.go` · `for _, k := range typ.Keys {` und `if klasse == klasseGUID {`; Test `src/hexagon/model/abfrage_ordnung_test.go` · `func TestGleichstandSchluessel`
- `befund`: Der Fänger führt als Schlüssel `Edm.Int32` (nur positiv), `Edm.String`, `Edm.Guid` und eine Folge aus zwei Properties. Nach dem Text geordnet bleiben grün: `Edm.Int64` (`V17`), `Edm.Decimal` (`V19`), `Edm.DateTimeOffset` (`V18`, der Text von `…:00Z` und `…:00.5Z` ist falsch geordnet), negative `Edm.Int32` (`V20`); Guid-Vergleich nur über die letzte Gruppe (`V01`) und mit Hex-Buchstaben vor Ziffern (`V02`) bleibt grün, denn die Fixtures sind Ziffern, die sich in der letzten Gruppe unterscheiden. Ein dritter Key wird weder geordnet (`D1`) noch in Stufe 1 geprüft (`D2`). Die Mengen-Tabelle des Plans führt keine Menge „Typ des Schlüssels“ und keine Achsenliste dafür.
- `verifizierbar`: ja (V01, V02, V17 bis V20, D1, D2)
- `klasse`: Menge ohne Achsenliste, Fänger führt nur drei Formen des Schlüssels

#### M-5 (MEDIUM): Stufe 1 prüft Key-Properties und Schlüssel nur an den ersten Positionen und Zuständen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.16; `SPEC-190` Nr. 3 e und f (leerer oder doppelter `Schluessel`; jede Key-Property fehlt, `null`, `redigiert`)
- `pfad`: `src/hexagon/model/abfrage.go` · `for _, k := range typ.Keys {` und `gesehen[e.Schluessel] = struct{}{}`; Test `src/hexagon/model/abfrage_fehler_test.go` · `zwei.Keys = []string{"Id", "Name"}`
- `befund`: Der Test für zwei Key-Properties kennt nur den Zustand *fehlend* der zweiten (`abE(1)` ohne `Name`); fehlend fängt aber auch die Ordnung (`pruefeWert` auf dem leeren Wert, anderer Pfad, gleiche Klasse). `null` und `redigiert` an der zweiten Key-Property sind ungefangen: Mutation `V22` (nur die erste Key-Property geprüft) bleibt grün, und eine zweite Key-Property `null` liefert ein Ergebnis statt `ErrUngueltig`. Der doppelte Schlüssel wird nur an drei Entities geprüft: `V23` (nur die ersten drei Schlüssel gemerkt) bleibt grün.
- `verifizierbar`: ja (V22, V23, D2)
- `klasse`: Zustandsachse der zweiten Key-Property und Positionsachse des Schlüssels nur teilweise geführt

#### M-6 (MEDIUM): Filter gesetzt zusammen mit einer `$orderby`-Property, die eine Entity nicht nennt, hat keinen Fall

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.16 (Zusatztabelle: je Modus eine Zeile); `SPEC-190` Nr. 5 Stufe 5 (zuerst die des Filters, dann die von `$orderby`)
- `pfad`: `src/hexagon/model/abfrage.go` · `if err := verlangeProperties(entities, filterNamen(opt.Filter.Ausdruck, nil), AbfrageFilter); err != nil {`; Test `src/hexagon/model/abfrage_ordnung_test.go` · `func TestOrderByEigenschaftFehlt`
- `befund`: `TestOrderByEigenschaftFehlt` hat nie einen Filter; in `TestRangfolge` tragen beide Verstöße der Stufe 5 den Fehler des Filters zuerst. Der Fall „Filter erfüllt, `$orderby`-Property fehlt in einer Entity“ ist ungeführt: Mutation `V15` (Stufe 5 prüft `$orderby` nicht, sobald ein Filter gesetzt ist) bleibt grün; die Antwort wäre eine Ordnung, die die fehlende Property still als `null` einsortiert (`SPEC-191` Nr. 3 c verbietet das).
- `verifizierbar`: ja (V15)
- `klasse`: Modus der Zusatztabelle (Filter gesetzt gegen nil) nicht an der Stufe 5 für `$orderby` geführt

#### M-7 (MEDIUM): Positionsachsen von Listen: Elemente mitten in `$orderby`, Navigations- und Select-Namen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe, die der Code verschieden behandelt), §3.17 (Menge mit Achsen); `SPEC-190` Nr. 7 (keine Obergrenze für Elemente von `$orderby`), Nr. 4 (Name einer Navigation-Property)
- `pfad`: `src/hexagon/model/abfrage_ordnung.go` · `for _, s := range order {` und `slices.ContainsFunc(typ.NavigationProperties, func(n NavigationProperty) bool { return n.Name == name })`; `src/hexagon/model/abfrage.go` · `for _, n := range sel.Namen {`
- `befund`: `TestOrderByLaenge` lässt in jedem Lauf nur das letzte Element variieren, alle mittleren sind in allen Entities gleich; ein überlesenes Element mitten in der Liste (`V05` Index 3, `V06` Index 16) ist nicht zu sehen. Der Typ der Tests hat genau eine Navigation-Property, eine Prüfung nur der ersten (`V13`) bleibt grün. `$select` mit mehr als vier Namen kommt nie vor (`TestProjektion` ist auf vier Properties gebaut, `V21` bleibt grün).
- `verifizierbar`: ja (V05, V06, V13, V21)
- `klasse`: Position in einer Liste nicht über die Länge variiert

#### M-8 (MEDIUM): „Das Ergebnis gehört dem Aufrufer“ ist weder zwischen zwei Aufrufen noch über die erste Entity hinaus gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.10 (Zusage über Nebenläufigkeit: Wirkungs-Fänger oder Negativ), §3.15; `SPEC-190` Nr. 2 (*„Das Ergebnis gehört dem Aufrufer … teilt keinen Slice mit ihr“*) und Nr. 6 (*„keinen geteilten Puffer“*)
- `pfad`: `src/hexagon/model/abfrage_test.go` · `func TestErgebnisUnabhaengig` und `r, err := model.FuehreAus(abTypFuer(qs[i].name), ess[i], opts[i], qs[i].beob)`; `src/hexagon/model/abfrage.go` · `return Ergebnis{Entities: projiziere(geordnet[von:bis], opt.Select), Anzahl: len(geordnet)}, nil`
- `befund`: Kein Test schreibt in ein zurückgegebenes Ergebnis, während andere Aufrufe laufen, und kein Test hält zwei Ergebnisse zugleich und ändert eines. `S1` (leeres Ergebnis ist eine gemeinsame Paketliste mit Kapazität) bleibt grün, Detektor 0 von 20 und Wirkung 0 von 20; `S2` (gemeinsame Liste in `AlleAbfrageUrsachen`) ebenso. `TestErgebnisUnabhaengig` prüft eine Entity: `V12` (Alias auf `Properties` der Eingabe ab der zweiten Entity, ohne `$select`) bleibt grün. Die Zeile 30 der Tabelle nennt „teilt nichts mit der Eingabe“; zwei Ergebnisse untereinander und Entities ab Index 1 sind nicht geführt.
- `verifizierbar`: ja (S1, S2, V12)
- `klasse`: Wirkungs-Fänger zur Nebenläufigkeit schreibt nicht in das Ergebnis

#### L-1 (LOW): Der Ausgang des Risikos „Rest der Ordnung“ widerspricht seinem Text

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.13 (eine offene Entscheidung steht mit Ausgang *offen* und Adresse in §6)
- `pfad`: `docs/plan/planning/done/slice-query-interpreter.md` · `Ausgang:** entschieden: `SPEC-191` Nr. 3 e` und `die Zeile ist eine Rückgabe an den Planner`
- `befund`: Derselbe Absatz nennt den Ausgang *entschieden* und die fehlende Tabellenzeile eine Rückgabe an den Planner; die Rückgabe hat weder Ausgang *offen* noch eine Adresse. Der Fänger bleibt damit (siehe M-3) ohne Zeile und ohne Empfänger.
- `verifizierbar`: nein
- `klasse`: Rückgabe ohne Adresse im Plan

#### I-1 (INFO): Kommentar im Konjunktiv über die Ursache

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.7
- `pfad`: `src/hexagon/model/abfrage_fehler_test.go` · `auch wenn das zweite die frühere Ursache hätte`
- `befund`: Beschreibt einen Fall des Tests, keine verworfene Alternative; nach §3.7 trägt er die Klasse *Abgrenzung*. Kein Befund, nur der einzige Konjunktiv in den neuen Dateien.
- `verifizierbar`: nein
- `klasse`: —

#### I-2 (INFO): Rate des Detektors für `S3` und Folge für Zeile 32

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.10
- `pfad`: `docs/plan/planning/done/slice-query-interpreter.md` · `Raten der Zeile 32`
- `befund`: Der gemeinsame Puffer in `projiziere` (`S3`) wird vom Detektor in 20 von 20 Prozessen gefangen, ohne Detektor in 0 von 20 (flatternd, wie `S4namen` im Plan); die Aussage des Plans, solche Zugriffe seien nur dem Detektor zugänglich, stimmt. Die Aussage des Plans, dass jede Funktion von `abfrage*.go` einen Zähler fängt (5 von 5), ist mit 20 von 20 je Zähler an 48 Stellen bestätigt (siehe oben).
- `verifizierbar`: ja
- `klasse`: —

---

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/abfrage.go` (Rangfolge, Programmfehler Stufe 1, Paging, Projektion) | geprüft: gelesen gegen `SPEC-190` Nr. 3 und 5 und `SPEC-191` Nr. 4 und 5; Befunde M-2, M-5, M-8; sonst ohne Befund |
| `src/hexagon/model/abfrage_filter.go` (Filter, statische Prüfung) | geprüft, ohne Befund: eigene Zähler 20 von 20; Mutationen des Plans (R14, R07) nicht wiederholt, `V11` rot |
| `src/hexagon/model/abfrage_ordnung.go` | Befunde M-1, M-3, M-4, M-6, M-7; Namenstabelle und Reihenfolge der Ursachen gelesen, `V14` rot-paket |
| `src/hexagon/model/abfrage_fehler.go` | geprüft, ohne Befund: Klassen, Codes, Texte gegen `SPEC-190` Nr. 4; Zähler 20 von 20 |
| `src/hexagon/ports/outbound/abbildung_test.go` (Zeile 24) | gelesen, nicht mutiert: 4 Ursachen × 3 Optionen × 4 Formen × 8 Gegenstände, ohne Befund; `abbildung.go` unberührt |
| `tests/integration/abfrage_url_test.go` (Zeile 34) | gelesen, nicht mutiert: 12 Query-Texte mit Ergebnis, 2 im Parser, ohne Befund; Test hängt nur an `urlparser` und `model` (Kante `tests` → `model` und `tests` → Driven Adapter der Architekturregeln) |
| Sendung (a) bis (c) von `slice-edm-ausdruecke` | geprüft: (a) Zeile 8, (b) Zeile 20, (c) Zeile 15 haben Fänger; an (b) nur `V14` (Gegenrichtung: `HatOrdnung` auch für `$select`, rot-paket); sonst ohne Befund |
| Gate-Kette, `.a-check.yml`, `//nolint`, V2 | geprüft, ohne Befund: keine Datei der Kette im Diff; kein `nolint`; der Interpreter bleibt versionslos, wie es [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) für den Kern festlegt |
| Größenmarke (`AGENTS.md` §3.14) | geprüft, ohne Befund: `git diff --numstat 055c295..a37fb08` über die zehn Pfade ergibt 2588 (Produktcode 491, Tests 2097), Marke 4500 |
| Rohbytes (`AGENTS.md` §3.18) | geprüft, ohne Befund: 0 Treffer in den elf neuen Dateien |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft: Produktcode ohne Konjunktiv und ohne abwesenden Text; Köpfe der Tests nennen Fälle, die sie führen (`I-1`) |

**Nicht aufgelöst (ehrlich).** Die Zeilen 1 bis 14, 21 bis 23, 25 bis 29, 31 und 34 habe ich nicht mit den Mutationen des Implementers nachgefahren; eigene Mutationen an den Zeilen 4, 6, 9, 12 bis 22, 25, 27, 30, 32 und 33 stehen oben. `make gates` und `make test-race` nicht gefahren (Aufgabe des Verifiers); der SQLite-Flake in `make test-race` wird deshalb nur vermerkt, nicht gesehen. Der Rest des Plans (§1, §2) nicht gegen die DoD gelesen. Die Zeilen der Mengen-Tabelle „Zustand einer Property“ (`Edm.Boolean` mit `null` und die übrigen sieben Typen) und „Stellen im Filter“ habe ich nicht mit eigenen Elementen gemutet; `abfrage_filter.go` ist mit Zählern und nur einer Mutation (`V11`) belegt.

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 8 |
| LOW | 1 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Programmfehlerpfad einer Stufe ohne Assertion · Prüfgrenze nur an einer Stufe der Pipeline gemessen · Fixture gleicht dem Wert der Mutation, Zusage ohne Zeile · Menge ohne Achsenliste, Fänger führt nur wenige Formen des Schlüssels · Zustandsachse der zweiten Key-Property und Positionsachse des Schlüssels nur teilweise geführt · Modus der Zusatztabelle nicht an einer Stufe geführt · Position in einer Liste nicht über die Länge variiert · Wirkungs-Fänger zur Nebenläufigkeit schreibt nicht in das Ergebnis · Rückgabe ohne Adresse im Plan

## Verdikt

**Merge-blockierend:** ja — die MEDIUM-Findings M-1 bis M-8 blockieren typischerweise; eine Abweichung davon begründet dieser Bericht nicht. Am Produktcode selbst fand ich keinen Fehler; die grünen Mutationen liegen an Formen der Eingabe, die kein Fänger führt, und die Kontrollen (`V11`, `V24`, `W4`) wurden rot.

**Übergabe:** Findings gehen an den Implementer; die Finding-Klassen gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation.

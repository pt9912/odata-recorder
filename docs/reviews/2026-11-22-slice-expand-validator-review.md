# Review-Report: slice-expand-validator — 2026-11-22

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-expand-validator`, HEAD `a86ab781` (Messbasis `a8d8f720`, Basis des Zweigs `main` `c70252f2`); Zeilen 1 bis 21 der Tabelle *Zusage · Fänger*. Produktcode `src/hexagon/model/expand_typ.go` (`PruefeExpand`, 56 Zeilen) und `src/hexagon/model/expand_typ_fehler.go` (`ExpandFehler`, `ExpandUrsache`, 46 Zeilen); Tests `expand_typ_fixture_test.go`, `expand_typ_test.go`, `expand_typ_fehler_test.go`, `expand_typ_eigenschaft_test.go`, `expand_typ_contract_test.go`; Plan und Roadmap.

**Skill:** `.harness/skills/reviewer.md` @ `a86ab781` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-22

**Eingangs-Kontext:**

- Plan `slice-expand-validator`: §1 (Ziel, Abgrenzung, Sendungen), §2, §3 (Tabelle Zeilen 1 bis 21, Abweichungen vom Entwurf, Mengen, Mutationen, Raten, akzeptierte Negative und Prüfgrenzen), §4 (Block, Marke 2500, Messung), §6 (Risiken, Entscheidungen V1 bis V7, Raster, Zusatztabelle)
- `SPEC-263` (Kopf, Nr. 1 bis 6, *Ungeprüft*), `SPEC-109` Nr. 5, `SPEC-262` Nr. 4 und 6, `SPEC-167` (Muster), `SPEC-210` Nr. 5 c, `SPEC-122` Nr. 8 und 9
- `LH-FA-16` (der Welle-Plan `welle-expand-batch` nennt `LH-FA-16.a`), `LH-RB-01`, `LH-QA-05`
- ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Bezug des Plans, nur genannt)
- `AGENTS.md` §3.7, §3.9 bis §3.18
- Vorherige Reviews am Modul `model`: `slice-expand-parser`, `slice-telemetry-einbau-mock-mutation` (Klassen *Fänger schmaler als Zusage*, *Wirkungs-Fänger statt Datenrennen-Hoffnung*, *Kommentar nennt mehr, als der Fänger fängt*)

**Unverändert (geprüft mit `git diff a8d8f720..HEAD --stat`):** außer Plan, Roadmap und den sieben neuen Dateien unter `src/hexagon/model/` ist nichts berührt; insbesondere nicht `src/hexagon/ports/outbound/abbildung.go`, `spec/`, `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/mk/`, `Dockerfile`, `tools/gegenproben/`, `harness/gegenproben.stempel`, `AGENTS.md`, `.claude/`, `.harness/skills/`. Keine Direktive `//nolint`, kein `t.Skip`; in den sieben Dateien keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP, kein Nullbreitenzeichen (`LC_ALL=C grep -P`, 0 Treffer je Datei); Zeichen jenseits von ASCII stehen nur in Kommentaren und als Bytewerte im Fixture.

**Größenmarke (nachgemessen):** `git diff --numstat a8d8f720..HEAD` über `src/hexagon/model/`: 56 + 46 + 104 + 163 + 175 + 351 + 508 = 1403 Zeilen (Produktcode 102, Tests 1301), wie berichtet, gegen die Marke 2500.

**Zählprüfungen (nachgemessen):** `go test -v` der Expand-Tests nennt 206, 1036, 84, 1014, 2062, 136, 1200, 8192, 750, 39 680, 336, 18, 264 und 12 Fälle, genau die Zahlen der Tabelle *Mengen*; die neun Generatoren samt den drei Verstößen der Zeile 12 ergeben 6827 Eingaben. Abdeckung der zwei Dateien (ohne Detektor, nur die Expand-Tests): jede Funktion 100,0 %.

## Eigene Mutationen

Ein eigener Treiber im Scratchpad (`expand-validator-reviewer/`, außerhalb des Repos): Container `odata-recorder:race` ohne Netz und Mount, `--memory=4g --memory-swap=4g --cpus=4`, Wegwerfkopie des Worktrees im Container, je Mutation ein Austausch mit `perl` und `go test -count=1 -run 'PruefeExpand|ExpandFehler' ./src/hexagon/model/` (je Prozess ein Lauf; die Klasse `andere` ohne Detektor), nacheinander unter `flock /tmp/odata-recorder-heavy.lock`, nie parallel zu `make gates`. `tools/harness/mutationslauf.sh` ist für die Klasse `nebenläufig` im Paket `model` nicht gefahren (Nullprobe über 2 Minuten, Befund des Plans, bestätigt durch den Lauf des ganzen Pakets unter dem Detektor: 123 s). `make gegenproben` ist nicht gefahren, der Stempel nicht berührt.

Die Mengen der Zusagen hat der Review aus dem Wortlaut von `SPEC-263` gebildet (Achsen: Art der Liste, Merkmale des Typs *abstrakt, offen, abgeleitet*, Zeichenmenge des Namens, Länge der Liste, Paar von Verstößen, Inhalt des Fehlerwerts), nicht aus den Mutationen des Implementers.

**Ergebnis:** 31 verschiedene Mutationen, **16 rot**, **15 grün**. Von den 15 grünen sind 14 die Grundlage der Findings F-1 bis F-5 und INFO-1; eine (R15) fällt unter das akzeptierte Negativ der Zeile 8 (Regel erst über 32 Elementen).

| Nr. | Mutation (Wegwerfkopie) | Ergebnis | Finding |
|---|---|---|---|
| R01 | Navigation-Suche nur über die ersten 20 Navigationen | rot (`TestPruefeExpandGueltig`) | — |
| R02 | Property-Suche nur über die ersten 3 Properties | rot (`TestPruefeExpandProperty`) | — |
| R03 | `case typ.Open && !typ.Abstract` | **grün** | F-1 |
| R04 | `case typ.Open && !(typ.Abstract && typ.BaseType != "")` | **grün** | F-1 |
| R05 | Liste über 32 Elementen: `formFehler` (Programmfehler) | **grün** | F-2 |
| R22 | Liste über 33 Elementen: `formFehler` | **grün** | F-2 |
| R06 | Prüfschleife nur über die ersten 32 Elemente (der Rest bleibt ungeprüft, Ergebnis `nil`) | **grün** | F-2 |
| R15 | Prüfung auf doppelte Namen aus bei mehr als 32 Elementen | grün | akzeptiertes Negativ Zeile 8 |
| R11 | Feld `Name string` im `ExpandFehler` (ungenutzt in `Error()`) | **grün** | F-3 |
| R07 | `case "*", "**", "*x":` (mehr Namen als Programmfehler) | **grün** | F-4 |
| R08 | Programmfehler für jeden Namen mit Präfix `*` | **grün** | F-4 |
| R19 | `case "*", "a*":` | **grün** | F-4 |
| R09 | Programmfehler für jeden Namen, der nur aus Leerraum besteht | **grün** | F-4 |
| R20 | `case "", " ":` | **grün** | F-4 |
| R10 | Schlüssel der Prüfung auf doppelte Namen: `strings.TrimSpace(e.Name)` | **grün** | F-4 |
| R21 | Schlüssel der Prüfung auf doppelte Namen: `strings.ToLower(e.Name)` | rot (`TestPruefeExpandContract`, `TestPruefeExpandProgrammfehler`) | — |
| R16 | Vergleich der Properties nach `NFD` zu `NFC` des Namens | **grün** | F-5 |
| R17 | Vergleich der Navigationen nach `NFD` zu `NFC` des Namens | rot (`TestPruefeExpandSchreibweise`) | — |
| R18 | Name in `typ.Keys` ist `keine_navigation` | grün | INFO-1 |
| R12 | Navigation gilt nur mit nichtleerem `Target` | rot (`TestPruefeExpandOhneBinding`) | — |
| R13 | `*` wird übersprungen statt Programmfehler | rot (`TestPruefeExpandContract` und weitere) | — |
| R14 | Property mit nicht unterstütztem Typ im offenen Typ ist keine Property | rot (`TestPruefeExpandProperty`) | — |
| C1 bis C8 | Zähler einer Paketvariable je Zweig: `o` nil, leere Liste, leerer Name, `*`, doppelter Name, Navigation gefunden, Name im offenen Typ, Name im nicht offenen Typ | 8 von 8 rot **nur mit Detektor** (siehe unten) | INFO-2 |
| C9 | Paketweite `map` als Zwischenspeicher, vom Aufruf beschrieben | rot mit und ohne Detektor (6 von 6 je Arm) | — |

Die Mutationen R01, R02, R12 bis R14, R17 und R21 sind eigene Austausche, keine Wiederholung der Namen des Implementers (dessen 58 rote stehen im Plan). Eine Panik oder ein Zeitlimit statt `--- FAIL` des benannten Tests trat bei keiner roten Mutation auf, ausgenommen C9 (Absturz des Prozesses durch gleichzeitigen Schreibzugriff, rot ohne Detektor in jedem Lauf).

### Raten der Zeile 16 (Detektor gegen Wirkung)

Messender: der Reviewer, mit Prozessen aus zwei Test-Binaries (mit `-race` und cgo, ohne), je `-test.run '^TestPruefeExpandGleichzeitig$' -test.count=1`, nacheinander. Last des Hosts (Load Average der letzten Minute, 4 CPUs des Containers) zu den Messzeitpunkten 3,3 bis 8,8; zwei weitere Linien liefen zugleich (Mutationslauf und Review anderer Slices). Die Zahlen sind Größenordnungen unter dieser Last, kein fester Wert (`AGENTS.md` §3.10).

| Mutation | mit Detektor | ohne Detektor |
|---|---|---|
| unverändert, mit zusätzlicher unbenutzter Paketvariable | 0 von 2 | 0 von 2 |
| C1 bis C8: Zähler `zz++` je Zweig des Codes (acht Zweige) | je Zweig 6 von 6 | je Zweig 0 von 6 |
| C9: paketweite `map` beschrieben | 6 von 6 | 6 von 6 |
| C6 im **ganzen Paket `model` unter dem Detektor** (Form des Gates `make test-race`, 123 s) | 1 von 1 (2 Meldungen `DATA RACE`, `TestPruefeExpandGleichzeitig` rot) | — |

Ein weiterer Arm (nur die Expand-Tests, fünf Prozesse) zählte rot, weil mein Aufruf des Test-Binary außerhalb des Paketverzeichnisses lief und `TestExpandFehler` seine Quelldatei nicht fand; er ist verworfen und zählt nicht.

### Auflösung der Tabelle *Zusage · Fänger*

| Zeile | Eigene Mutationen (nicht die des Implementers; je Element der Klasse) | Ergebnis |
|---|---|---|
| 1 | R01 (nur die ersten 20 Navigationen) | rot |
| 2 | R17 (Normalform bei Navigationen), R21 (Falten bei Doppelten); R16 (Normalform bei Properties) | R17 rot, **R16 grün (F-5)** |
| 3 | R02 (nur die ersten 3 Properties), R14 (nicht unterstützter Typ im offenen Typ) | rot |
| 4 | R03 (offen und nicht abstrakt), R04 (offen und nicht abgeleitet-abstrakt) | **grün (F-1)** |
| 5 | wie 4 (abstrakter Typ im Verbund mit `Open`) | **grün (F-1)**; abstrakt und nicht offen ist geführt (Mutation des Implementers `z05a`) |
| 6 | keine eigene; die Mutationen des Implementers `z06a`, `z06b` sind im Code der Funktion formulierbar | nicht aufgelöst |
| 7 | Reihenfolge und Länge: R05, R06, R22 (über 32 Elemente) | **grün (F-2)** |
| 8 | R07 bis R10, R19, R20 (Namen, die kein Programmfehler sind), R13 | R13 rot, **R07 bis R10, R19, R20 grün (F-4)** |
| 9 | R13 in der Schleife über die Namen | rot |
| 10 | R11 (Feld im Fehlerwert) | **grün (F-3)** |
| 11, 12 | wie 10 (Zusage über den Inhalt des Fehlerwerts) | **grün (F-3)** |
| 13 | R12 | rot |
| 14 | keine eigene; `reflect.DeepEqual` über Typ und Knoten deckt jede Änderung (Schleife über 6827 Eingaben) | ohne Befund |
| 15 | R05, R22 (Grenze über 32) | **grün (F-2)** |
| 16 | C1 bis C9 (je Zweig, ein Zustand im Paket) | rot (Detektor 8 von 8 Zweigen; Wirkung nur C9) |
| 17 | Import von `ports/outbound` und von `adapters/driven/sqlite` in `expand_typ.go` (Wegwerfkopie, `a-check` im Image des Gates, netzlos, `:ro`) | je 1 Befund `core-impurity`; ohne Import 0 Befunde |
| 18 | Signatur über `reflect`: durch Lesen des Tests; Fälle 2 a bis d, 3 und 4 und die Zahl 31 nachgezählt | ohne Befund |
| 19 bis 21 | akzeptierte Negative; Grund, Folge und Maß stehen im Plan; Zeile 21 durch C1 bis C8 und 100,0 % Abdeckung bestätigt (0 von 9 Zweigen ohne Lauf) | ohne Befund |

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Das Fixture-Modell führt keinen abstrakten offenen Typ und keinen abstrakten abgeleiteten Typ. Die Mutationen `case typ.Open && !typ.Abstract` (R03) und `case typ.Open && !(typ.Abstract && typ.BaseType != "")` (R04) bleiben in allen Tests grün: ein abstrakter offener Typ lieferte für einen unbekannten Namen `navigation_unbekannt` (400) statt `navigation_nicht_unterstuetzt` (501), gegen `SPEC-263` Nr. 2 c und d („ein abstrakter Typ verhält sich wie jeder andere“). Zeile 4 führt offene Typen nur ohne `Abstract`, Zeile 5 abstrakte nur ohne `Open`; weder *Mengen* noch ein Negativ nennen das Produkt der Merkmale. | `AGENTS.md` §3.15, §3.17 (Menge ohne Achsen, Fänger führt eine Form) | `src/hexagon/model/expand_typ_fixture_test.go` · `{name: "Ns.Abstrakt", abstrakt: true,` und `{name: "Ns.Offen", offen: true,` | ja (`go test` der Expand-Tests mit R03 oder R04) | Achsenprodukt der Typmerkmale ungeführt |
| F-2 | MEDIUM | `SPEC-263` Nr. 6 sagt „keine eigene Grenze“ zu, Zeile 15 trägt die Zusage; der Fänger führt Listen bis 32 Elemente, nie 33 oder mehr. Die Mutationen „über 32 Elementen `formFehler`“ (R05), „über 33“ (R22) und „Prüfschleife nur über die ersten 32 Elemente“ (R06, ein Verstoß ab Stelle 33 bleibt ungeprüft, Ergebnis `nil`) bleiben grün; die Prüfgrenze „über 32“ steht im Plan allein für den doppelten Namen (Zeile 8), nicht für die Zusage der Zeile 15 und nicht für die Zeilen 1 bis 7. Szenario: hebt `SPEC-262` Nr. 6 die Grenze des Parsers an, gilt ein Name hinter Stelle 32 ungeprüft (`LH-RB-01`: nie eine Antwort ohne die verlangte Expansion). | `AGENTS.md` §3.15 (beide Seiten jeder Grenze), §3.17 (Prüfgrenze mit Grund und Folge) | `src/hexagon/model/expand_typ_eigenschaft_test.go` · `for n := 1; n <= len(tt.navs); n++ {` | ja (R05, R06, R22) | Zusage ohne Grenze, Prüfgrenze nicht benannt |
| F-3 | MEDIUM | `SPEC-263` Nr. 3 sagt zu: „Kein Teil der Eingabe steht im Fehler“, der Wert ist `ExpandFehler{Ursache}`. Der Fänger der Zeile 12 prüft nur drei Textformen (`Error()`, `%v`, `%+v`); ein zusätzliches Feld `Name string` im `ExpandFehler`, vom Code mit dem Namen belegt und nie in `Error()` genannt, lässt jeden Test grün (R11; `errors.As` reicht den Wert an jeden Aufrufer, `%#v` und eine Serialisierung des Werts zeigen ihn). Das Paket führt die Form der Prüfung bereits (`url_ast_expand_test.go`: `typ.NumField() != len(felder)`). | `AGENTS.md` §3.15 (Zusage ohne Fänger), §3.17 | `src/hexagon/model/expand_typ_fehler_test.go` · `for _, s := range []string{err.Error(), fmt.Sprintf("%v", err), fmt.Sprintf("%+v", err)} {` | ja (R11) | Zusage über Inhalt des Fehlerwerts prüft nur Textformen |
| F-4 | LOW | Die Namen der Fälle sind durchweg Bezeichner (`exvNamen`: Member, Typnamen, `Unbekannt`, `Fremd`); kein Fall führt einen Namen, der nur *fast* ein Programmfehler ist. Die Mutationen, die `*x`, `a*`, `**`, Präfix `*` oder ` ` (nur Leerraum) zum Programmfehler machen (R07, R08, R19, R09, R20) oder `a` und ` a` als Doppelte zählen (R10), bleiben grün. `SPEC-263` Nr. 4 nennt `""` und `*` abschließend, Nr. 1 verlangt bytegleiche Suche; Folge einer solchen Regel: `500` statt `400`. Erreichbar nur, wenn der Parser mehr als Buchstaben und Ziffern liefert (`SPEC-262` Nr. 4 a). | `AGENTS.md` §3.15 (Negativliste statt Positivliste), §3.17 | `src/hexagon/model/expand_typ_fixture_test.go` · `namen := []string{"Unbekannt", "Fremd"}` | ja (R07 bis R10, R19, R20) | Negativliste der Programmfehler ohne Gegenfälle |
| F-5 | LOW | Das einzige Zeichen jenseits von ASCII (`é`, NFC und NFD) steht in genau einer Navigation (`Café`), in keiner Property. Eine Normalisierung `NFD` zu `NFC` nur beim Vergleich der Properties (R16) bleibt grün, beim Vergleich der Navigationen (R17) rot: die Achse *Liste × Zeichenmenge* ist ungeführt. Die Prüfgrenze der Zeile 2 nennt fünf Abwandlungen, nicht die Liste, in der das Zeichen steht. | `AGENTS.md` §3.17 (Zeichen- und Strukturmenge getrennt, Achsen) | `src/hexagon/model/expand_typ_fixture_test.go` · `exvN("Caf"+exvEAkut(), "Ns.Lieferant", true)` | ja (R16) | Zeichenmenge nur in einer Liste des Fixtures |
| F-6 | LOW | Der Kopf des Fixture-Modells nennt „zwei offenen Typen“; das Fixture führt vier (`Ns.Offen`, `Ns.OffenKind`, `Ns.LeerOffen`, `Ns.VieleOffen`), die Tabelle *Mengen* des Plans ebenfalls vier. | `AGENTS.md` §3.11 (Kommentar sagt nur zu, was ein Test prüft), §3.7 | `src/hexagon/model/expand_typ_fixture_test.go` · `abstrakten, zwei offenen Typen, Typen ohne Navigation` | nein (kein Gate liest Zahlen im Kommentar) | Kommentar nennt andere Zahl als das Fixture |
| F-7 | LOW | §6 des Plans gibt die Grenze der Nullprobe an den Planner als neue Klasse „Mutationswerkzeug scheitert an der Laufzeit des Pakets unter dem Detektor“ mit „erster Evidenz dieser Slice“. Das Register führt dieselbe Ursache bereits als `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` mit 10 Dateien unter `evidence/`; die Beobachtung wird nicht zitiert, der Zähler würde zwei Namen getrennt führen. | Slice-Vorlage (vorhandene `BEO-<KUERZEL>/<slug>` zitieren statt neu formulieren), `AGENTS.md` §3.13 (Adresse) | `docs/plan/planning/in-progress/slice-expand-validator.md` · `erste Evidenz dieser Slice` | nein | Beobachtung unter neuem Namen statt Zitat |
| F-8 | LOW | Die Zellen *Fänger* und *Mutation* der Zeilen 1, 5, 6, 9 und 18 tragen weiter den Entwurf (etwa Zeile 9: „Schleife über die fünf Arten des Programmfehlers“, der Test führt drei; Zeile 6: „Die Prüfung sucht auch in den Subtypen“, im Code nicht formulierbar); der Ist-Stand steht nur im Absatz *Abweichungen* darunter. Wer die Tabelle allein liest, findet Zusagen, die der Test nicht trägt. Die Abweichung ist ausgewiesen und ändert keine Zusage. | `AGENTS.md` §3.9 (jede Stelle, die eine berichtigte Aussage trägt, im selben Commit nachziehen), §3.11 | `docs/plan/planning/in-progress/slice-expand-validator.md` · `Schleife über die fünf Arten des Programmfehlers` | nein | Tabellenzelle trägt Entwurf neben Ist-Stand |
| INFO-1 | INFO | `exvModell` setzt `Keys: []string{"Schluessel"}`; der Name ist keine Property des Fixtures und steht in keiner Menge der Fälle. Eine Regel, die `typ.Keys` liest (R18), bleibt grün. `SPEC-263` sagt über `Keys` nichts zu; Hinweis an den Implementer, falls eine spätere Fassung die Keys liest. | Maintainability | `src/hexagon/model/expand_typ_fixture_test.go` · `Keys: []string{"Schluessel"}` | ja (R18) | — |
| INFO-2 | INFO | Zeile 16: Die Mutationen `z16a` und `z16c` fängt allein der Detektor (Plan: 50 von 50 mit, 0 von 50 ohne). Das ist nach `AGENTS.md` §3.10 gefangen, kein akzeptiertes Negativ: ein Datenrennen zählt der Detektor je Prozess. Eigene Messung bestätigt es für acht Zweige des Codes (je 6 von 6 mit Detektor, 0 von 6 ohne) und für das ganze Paket unter dem Detektor (1 von 1, 123 s); die Aufteilung der Prüfgrenze in §3 (zwei Mutationen mit Wirkung gefangen, zwei nur vom Detektor) stimmt. Ohne Detektor-Lauf im Gate (ein Gate ohne cgo) gälte die Zusage nicht. | `AGENTS.md` §3.10 | `docs/plan/planning/in-progress/slice-expand-validator.md` · `der Übersetzer reicht den geschriebenen Wert weiter` | ja (`make test-race`) | — |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/expand_typ.go` (Reihenfolge Nr. 2 a bis d, Programmfehler vor jedem Verstoß, bytegleicher Vergleich, kein `ctx`, kein Port, keine Paketvariable) | geprüft, ohne Befund; 16 eigene Mutationen rot, die grünen führen auf Fänger, nicht auf den Code |
| `src/hexagon/model/expand_typ_fehler.go` (`Code`, `Error`, `Is`, Klasse der fremden Ursache) | geprüft, ohne Befund; Code stimmt mit Tabelle `SPEC-263` Nr. 3 überein, 100,0 % Abdeckung |
| Rang (Zeile 7): Paar 400 vor 501 am selben Schlüssel, beide Folgen, offener und nicht offener Typ, Listen bis 32, jedes Paar von Stellen | geprüft, ohne Befund (Zusatztabelle in §6: alle fünf Paare tragen einen benannten Fänger) |
| Programmfehler (Zeilen 8 und 9): `nil`, leere Liste, leerer Name, `*`, Doppelte mit jedem Abstand 0 bis 30, Verstoß davor und danach | geprüft, ohne Befund bis auf die Gegenfälle in F-4 |
| Schreibweise (Zeile 2): ASCII, Unicode, NFC und NFD, Leerzeichen, Falten | geprüft, ohne Befund bis auf F-5 |
| Subtyp, Basis-Kette, geerbte Member, Navigation ohne Binding, abstrakter Typ (Zeilen 1, 3, 5, 6, 13) | geprüft, ohne Befund bis auf F-1 |
| Wiederholung und Änderungsfreiheit (Zeilen 14, 15) | geprüft, ohne Befund bis auf F-2 |
| Nebenläufigkeit (Zeile 16) | geprüft, ohne Befund (INFO-2) |
| Architektur (Zeile 17): Regeln von `make a-check` gegen Import von `ports/outbound` und eines Adapters | geprüft, ohne Befund; je 1 Befund `core-impurity`, ohne Import 0 |
| Validator-Contract-Test (Zeile 18): 31 Fälle von `SPEC-263` Nr. 2 a bis d, 3, 4; Signatur über `reflect` | geprüft, ohne Befund |
| Kommentare im Produktcode (§3.7): drei Funktionen, vier Methoden, drei Typen und Konstanten | geprüft, ohne Befund (Zusage, Kopplung, Rang-Zeiger; keine verworfene Alternative, kein abwesender Text) |
| Plan-Nachzug (§3.9): §1, §2, §3 (Dateitabelle 102 Zeilen, Messung 1403, Mutationszahl 59 = 58 rot + 1 grün nachgezählt), §4, §6, Roadmap; Sendungen an `slice-expand-ausfuehrung` (dessen §1 nennt beide Sendungen und den Geber) | geprüft, ohne Befund bis auf F-7 und F-8 |
| Akzeptierte Negative (Zeilen 19 bis 21 und die fünf Prüfgrenzen): Grund, Folge, Maß | geprüft, ohne Befund bis auf die Lücke in F-2 |
| Roadmap, `spec/`, `abbildung.go`, Gates, `.claude/`, `.harness/skills/` | geprüft: unverändert bzw. nur die Zeile *In Arbeit* in der Roadmap, ohne Befund |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 5 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Achsenprodukt der Typmerkmale ungeführt · Zusage ohne Grenze, Prüfgrenze nicht benannt · Zusage über Inhalt des Fehlerwerts prüft nur Textformen · Negativliste der Programmfehler ohne Gegenfälle · Zeichenmenge nur in einer Liste des Fixtures · Kommentar nennt andere Zahl als das Fixture · Beobachtung unter neuem Namen statt Zitat · Tabellenzelle trägt Entwurf neben Ist-Stand

## Verdikt

**Merge-blockierend:** ja nach der Regel des Skills (F-1 bis F-3, MEDIUM, Fänger schmaler als die Zusage); der Produktcode selbst ist ohne Befund, und jede Mutation, die die Wirkung der Funktion ändert, ist rot: die drei MEDIUM-Findings betreffen Fänger und Plan, nicht den Code.

**Übergabe:** Die Findings gehen an den Implementer; die Finding-Klassen gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. F-7 geht an den Planner. Dieser Report ersetzt keine Verifikation: DoD- und Spec-Konformität prüft der Verifier.

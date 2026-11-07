# Verifikation: slice-odata-json-eingabe — 2026-11-07

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-180`, `SPEC-181`, ADRs und Plan; nicht Code-Review (das ist `docs/reviews/2026-11-06-slice-odata-json-eingabe-review.md`).

**Gegenstand:** Zweig `slice-odata-json-eingabe`, HEAD `0f3b125`, Basis `d4ca2e2`. Plan: `docs/plan/planning/done/slice-odata-json-eingabe.md`. Die Fix-Runde nach dem Review (F-1 bis F-15) hat nur Tests, Fixture und Plan geändert; der Produktcode ist gegenüber dem Review-Stand unverändert (`git diff 7ef049a..HEAD` über `eingabe*.go` ohne Tests ist leer, geprüft).

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Konform mit Auflagen: der Produktcode erfüllt `SPEC-180`/`SPEC-181`, die Gates sind grün, die Größenmarke hält. Zwei Fänger-Lücken (V-1, V-2, je MEDIUM nach `AGENTS.md` §3.15) zeigen, dass zwei Fixes der Fix-Runde die Klasse nicht decken (`AGENTS.md` §3.17). Schließbar nach: V-1 und V-2 durch den Planner entschieden (Fix oder akzeptiertes Negativ mit Grund und Folge), Closure-Notiz, Register.** Eine Spec-Abweichung des Codes habe ich nicht gefunden.

## 1. Sensor-Belege (selbst gefahren)

| Sensor | Beleg |
|---|---|
| `make gates` | einmal selbst, unter `flock /tmp/odata-recorder-heavy.lock`, eigenes Log im Scratchpad: Exit 0. `a-check`: 0 Befunde; `d-check`: 695 Dateien, 0 Befunde; `groessenmarke`: ok, Messung 3597 von Marke 5900; `migration-check` ok; `zeichen`: ok, 1420 Dateien; `baseline-verify` ok. Die Stufen `test` und `race` des `Dockerfile` standen im Lauf als `CACHED` (gleiche Schichten wie ein früherer grüner Lauf, ein Cache-Treffer setzt einen bestandenen `RUN` voraus); zusätzlich lief die Nullprobe des Mutationslaufs das Paket `jsoncodec` ungemutiert mit und ohne Detektor grün. |
| Größenmarke | 3597 (Summe aus Einfügungen und Löschungen über die Pfade des Blocks) gegen 5900: 61 %. Der Block (`basis d4ca2e2`, `schaetzung 4360`, `marke 5900`, `schnitt-grund`) steht vor dem ersten Code-Commit (`b833471`). Plan §4 nennt die Metrik. |
| Gate-Kette | `git diff d4ca2e2..HEAD` über `Makefile`, `harness/mk`, `Dockerfile`, `.a-check.yml`, `.golangci.yml`, `tools/`: leer. Bedingung aus §4 („nicht berührt“) stimmt. |
| Gemeinsamer Code | `leser.go`, `wert.go`, `scan.go`, `zahl.go`, `baum.go`, `deuten.go`: unverändert (Plan §6 und Zeile 43 stimmen). |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -nP '\\u[0-9A-Fa-f]{4}'` über alle `eingabe*.go`, `*eingabe*_test.go` unter `jsoncodec`, `model`, `outbound` und den Plan: kein Treffer. Kommentare der neuen Dateien auf Chronik durchsucht (`Plan nennt`, `Rückgabe`, `früher`, `Review`): kein Treffer (F-1 behoben). |
| Code `ODR-E6007`, 37 Codes | `model/fehler.go`: `CodeE6007`, 400, `ZielMock`, `Invalid request body`; `abbildung.go`: `ErrEingabeUngueltig` → `E6007`, `ErrEingabeNichtUnterstuetzt` → `E6002`; `fehler_test.go` prüft 37; `SPEC-109` Nr. 2 und 5, `SPEC-230` Nr. 8 (37), `SPEC-180` §4 und `architecture.md` tragen den Code. Stimmig. |

## 2. DoD

| DoD-Punkt (§2) | Befund |
|---|---|
| Typisierte Eingabe, acht Typen, `null` und fehlend unterscheidbar, unbekannte Property `400` | erfüllt (Zeilen 23, 24, 20; `TestEingabeZustaende` 192 Zellen, `TestEingabeWerte` 423) |
| `501` mit `ODR-E6002`, fehlende erforderliche Property bei `POST`/`PUT` `400` | erfüllt im Port und in der Abbildung (Zeilen 20, 30, 44); die Wirkung am Client entsteht erst mit den Nehmern (Zeile 45, akzeptiertes Negativ, Sendung angenommen) |
| Randformen je Zelle Ort und Fänger | Raster in §6 lückenlos (jede Zelle trägt Nummer oder „gilt nicht“); Fänger vorhanden, Ausnahmen siehe V-1, V-2 |
| `make gates` grün | erfüllt (Abschnitt 1) |
| Review durchgeführt, Report liegt vor | erfüllt (`2026-11-06-…-review.md`, F-1 bis F-15; Fix-Runde nachgezogen) |
| Doku-Update | erfüllt (`SPEC-180`, `SPEC-181`, Verweise, `architecture.md`, Historie); Handbuch bewusst nicht (Planner-Entscheidung §6, Code erreicht keinen Client) |
| Closure-Notiz mit Lerneintrag | **offen** (§7 „Wird bei der Closure gefüllt“) — Aufgabe des Planners |
| Beobachtungs-Register fortgeschrieben | **offen** — V-1 ist ein weiterer Treffer für `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (Planner zählt, kein Zähler wird hier gesetzt) |
| Jedes Risiko aus §6 trägt einen Ausgang | erfüllt mit drei bewusst *offen* gebliebenen: F-15 (Länge des Ortes, Adresse Architect und Planner, im selben Commit wie der Bericht, `AGENTS.md` §3.13 gehalten), `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (Lese-Schritt der Closure), Historie der Spezifikation beim Merge. Mit V-1 und V-2 kommen zwei weitere hinzu. |
| Drei Paarungen (Anker · Folge-Slice · Register) | Folge-Slice-Sendungen angenommen (§6); Anker/Register mit der Closure |

## 3. Plan-vs-Code-Diff

- Tabelle *Zusage · Fänger*: 45 Zeilen, jede Zeile hat einen benannten Test, der im Paket existiert (Namensliste gegen `grep '^func Test'` geglichen: alle 45 Fänger-Namen und die Generatoren vorhanden). Zeile 45 ist das akzeptierte Negativ mit Grund und Folge, die Nehmer tragen den Fänger.
- Abweichungen des Berichts (Zeilen 3, 6, 12, 33, 40, 43): vom Architect entschieden (§6), im Plan nachgezogen.
- Zählungen im Test gegen Plan: Zeile 29 im Test 35 (ohne Detektor) und 27 (mit Detektor) mit Zählprüfung; Zeile 39 im Test 16 Zweige, im Plan 16 — gleich.
- Der Produktcode (`eingabe.go` 102, `eingabe_contenttype.go` 46, `eingabe_deuten.go` 153, Modell 118) liest sich gegen `SPEC-181` Nr. 1 bis 8 ohne Abweichung: Rangfolge Argumente → Kontext → `Content-Type` → Lesen → Kontext → Deuten; `@odata.type` vor den Mitgliedern; Fehler als Nullwert; `charset` jedes Vorkommen.

## 4. Befunde

| Nr. | Schwere | Befund |
|---|---|---|
| V-1 | MEDIUM | **Zweige unter dem Detektor: die Fix-Runde zu F-7 deckt 16 Zweige, nicht die Klasse „jede Rückgabe von `LeseEntity`“.** Der Kopf von `eGleichzeitigFaelle` (`eingabe_vertrag_test.go:158`) sagt zu: „jeder Zweig, den der Leser ausführen kann, hat einen Aufruf“, und das akzeptierte Negativ der Zeile 39 begründet mit „die Zweige der Spezifikation haben je einen Aufruf, weitere gibt es nicht“ (`AGENTS.md` §3.11). Eigene Mutationen der Klasse `nebenläufig` (Schreibzugriff auf `model.ErrEingabeUngueltig` in genau einem Zweig, `mutationslauf.sh`, `--memory 2g`, unter `flock`, Last des Hosts 2,7 bis 4,3 bei 4 CPUs, Messender: ich, 10 Prozesse mit Detektor und 2 ohne, `-test.count=1` je Prozess) bleiben in `TestEingabeGleichzeitig` **grün, 0 von 10**: `q01` beendeter Kontext vor dem `Content-Type`, `q02` `null` an einer Schlüssel-Property (`null`), `q03` ungültiger Term (`struktur` einer Annotation), `q04` `unbekannte_property`, `q05` Wurzel kein Objekt (`struktur`), `q06` abstrakter Typ des Sets ohne `@odata.type`, `q07` Kontext nach dem Lesen. **Kontrolle (Methode gültig):** `q08` im Zweig „Fehler von `src`“, den der Test führt, meldet der Detektor in 10 von 10 Prozessen (`flatternd` nur, weil der Arm ohne Detektor wie erwartet 0 von 2 ist). Die Menge aus dem Wortlaut sind die Rückgaben des Lesers: neunzehn Ursachen (`SPEC-180` Nr. 3), beendeter Kontext (zwei Stellen), Programmfehler; der Test führt davon `fehlend`, `wert`, `doppelter_name`, `bind`, `navigation`, `property_typ`, `offen`, `odata_typ`, `schluessel`, `content_type`, `src` und die Programmfehler; nicht `null`, `struktur`, `unbekannte_property`, `typ`/`zahl_zu_lang` als eigener Zweig, Kontext. Folge: ein Zugriff auf eine Variable des Pakets in einem dieser Zweige bleibt auch unter `make test-race` unbemerkt, genau die Klasse des Beobachtungseintrags `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`. Behebung entweder als Schleife über alle Ursachen (Menge 19 plus Kontext), oder als benanntes Negativ mit Grund und Folge, nicht mit dem jetzigen Satz. |
| V-2 | MEDIUM | **Mitglieder × Typ der Entity: F-6 deckte Pflicht und Schlüssel am abgeleiteten Typ, nicht die übrigen Mitglieder-Klassen.** Das Modell der Hierarchie (`Mitte`, `Blatt`, `Blatt2`, `Pflicht`, `Pflicht2`) trägt in den abgeleiteten Typen nur `Edm.Int32`/`Edm.String`-Properties, keine Navigation, keinen offenen Typ, keine nicht unterstützte Property. Meine Mutationen, die das Verhalten nur für einen Typ mit `BaseType` verändern, bleiben **grün in 2 von 2 Prozessen** (`TestEingabeMitglieder` bzw. `TestEingabeSchluesselNul`): `v01` Navigation (`navigation` würde `unbekannte_property`), `v02` offener Typ (`offen`), `v03` nicht unterstützte Property (`property_typ`), `v04` U+0000 in einer Schlüssel-Property (`schluessel`). Gegenprobe: `v05` (`Schluessel` des abgeleiteten Typs leer) ist rot-paket, also gefangen. `SPEC-181` Nr. 5 sagt „wirksame Property des Typs (die geerbten eingeschlossen)“ und Nr. 3 lässt den Typ des Bodys den abgeleiteten sein; die Zusage hat die Achse *Typ des Sets / abgeleitet* für alle Mitglieder-Klassen. Folge im Betrieb: ein Fehler beim Aufbau der Sicht eines abgeleiteten Typs (geerbte Navigation, offener abgeleiteter Typ) lässt einen `501` als `400` oder eine Entity als gültig durch. |
| V-3 | LOW | Unbeschränkte Klassen am Rand ohne Zeile in der Tabelle der Pflicht-Mengen (§3): (a) Rand des Medientyps und des `charset`-Teils: Fixture führt Leerzeichen, Tab, CR, LF, FF, VT, NBSP; `Trim` um U+0000 (`v10`) oder U+0085 (`v11`) bleibt grün (2 von 2). Realistischer Bruch (`TrimSpace`, `unicode.IsSpace`) ist gefangen (NBSP, `c02`), die Mutationen sind künstlich. (b) Nullbare Schlüssel-Property mit `null`: Fixture führt `Rs.Id`, `KN.A` (Int32) und `KN.B` (String); der Ausschluss von `Edm.Guid` (`v06`) bleibt grün. Die Prüfung ist im Code typunabhängig. (c) Property-Name mit angehängtem U+0000 (`v07`) bleibt grün. Folge: gering, weil keine realistische Fehlform; der Plan nennt die Prüfgrenze für diese Achsen nicht. Gefangen waren dagegen `v08`, `v09` (Steuerzeichen im Ort: Tab, U+007F), `v12` bis `v15` (Präfix und Suffix des `charset`-Wertes, des Namens und des Medientyps): rot, `TestEingabeCharset`/`TestEingabeMedientyp`, 1 von 1, bzw. rot-paket durch `TestEingabeEscapeRohbytes`. |
| V-4 | LOW | **Stufen über 1 MiB laufen unter `make test-race` nicht — als akzeptiertes Negativ: Grund ja, Folge für den Detektor-Lauf nur implizit.** Aufteilung ist ehrlich: `eingabe_detektor_an_test.go` (`//go:build race`, `eUnterDetektor = true`) und `eingabe_detektor_aus_test.go` (`//go:build !race`, `false`) tragen die Konstante, `TestEingabeZuteilung` kürzt nur die `1,`-Folge und zählt die Menge (35 ohne, 27 mit Detektor, `t.Errorf` bei falscher Zahl); `make test` (`go test ./...` ohne `-race`) fährt alle 35, der Kopf nennt die Aufteilung, der Plan (Absatz „Zuteilung“, §3) nennt Grund (2 GiB, Lauf beendet) und die Stelle. Es fehlt (a) die Folge der Aufteilung selbst (was der Detektor bei Eingaben über 1 MiB nicht sieht: ein Rennen in einem Pfad, der nur dort läuft — praktisch derselbe Code wie bei 1 MiB, ein Wachstumsschritt von `kinder`), denn die genannte „Folge im Betrieb“ gilt der Grenze von 14 MiB; (b) Zeile 29 der Tabelle nennt weiter nur „Menge 35“ ohne die 27 unter dem Detektor. |
| V-5 | INFO | Sägezahn der Zuteilung: die Stufen über 1 MiB fährt nur die `1,`-Folge; `[0,`-Folge (gleiche Form mit anderer Ziffer), `Mitglieder` und `Annotationen` stehen bei 1 MiB (54,5 und 44,7 je Byte nach dem Plan, Abstand zu 256 mehr als Faktor 4). Ich habe diese Formen über den Stufen nicht gemessen. Der Plan sagt „die Formen mit vielen Knoten (`[0,`- und `1,`-Folge)“, der Test deckt eine. |
| V-6 | INFO | Die Fix-Runde ist in §3 mit Menge, Elementzahl und Mutationen je Befund dokumentiert und die dort genannten Mutationen sind plausibel; ich habe sie nicht wiederholt, sondern gegen nicht genannte Elemente derselben Klassen gefahren (Abschnitt 5). Gut gehalten: F-2 (11 Größen, 224,3 Maximum), F-3 (3888 Zellen), F-4 (288), F-8 (24 Schreibweisen), F-9. Mutationen `v12` bis `v15` an Elementen, die der Befund F-3 nicht nannte, sind rot. |

## 5. Eigene Mutationen (Klasse, nicht genannte Elemente)

Treiber `sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `--memory 2g`, unter `flock`, ohne Netz, im Scratchpad (`…/scratchpad/verif-oje/liste-a.json`, `liste-b2.json`; `liste-b.json` war wirkungslos: `x = x` wird vom Übersetzer nicht instrumentiert, die Kontrolle `q08` blieb grün — Fehler der Mutation, berichtigt in `liste-b2.json`). Klasse `andere`: 15 Mutationen, Ergebnis 4 rot, 3 rot-paket, 8 grün (V-2: `v01` bis `v04`; V-3: `v06`, `v07`, `v10`, `v11`). Klasse `nebenläufig`: 8 Mutationen, 10 Prozesse mit und 2 ohne Detektor, `-test.count=1`: `q01` bis `q07` 0 von 10 (V-1), `q08` 10 von 10 (Kontrolle). Last des Hosts im Lauf 2,7 bis 4,5 bei `--cpus=4`; die Zahlen sind Größenordnung (`AGENTS.md` §3.10).

Nicht gefahren: Mutationen im gemeinsamen Code (`scan.go`, `zeit.go`; Review F-12: dort fängt `TestODataCodecVertrag`), Zuteilung über die Stufen für Mitglieder und Annotationen (V-5).

## 6. Akzeptierte Negative (Grund und Folge)

| Negativ | Grund | Folge | Urteil |
|---|---|---|---|
| Zeile 45, kein Aufrufer | `MockUseCase` und `$batch` liegen in Folge-Slices, Sendung angenommen | bis dahin ungesehen, Nehmer tragen den Fänger | trägt |
| Zeile 39, Zweig ohne Aufruf | „Zweige der Spezifikation haben je einen Aufruf“ | Rennen fiele erst mit eigenem Fänger auf | **trägt nicht**, Grund falsch (V-1) |
| Exponent `1e999999999` nicht mutierbar | mehr als 1 GiB | Klasse durch `1e9999999` belegt | trägt |
| Größen über 14 MiB nicht geprüft | Speicher des Containers | Betreiber begrenzt `MAX_CAPTURE_BODY` | trägt |
| Stufen über 1 MiB nicht unter dem Detektor | Speicher | `make test` fährt sie; Folge für den Detektor implizit | trägt mit LOW (V-4) |
| Prüfgrenzen der unbeschränkten Mengen (§3, fünf Zeilen) | Grammatik durch gemeinsamen Code bewiesen | Folge je Zeile genannt | trägt; Ränder des `Content-Type` fehlen als Zeile (V-3) |
| Schlüssel nicht unterstützten Typs (Spec *Ungeprüft*) | im Betrieb kein solches Modell (`SPEC-125` Nr. 10) | `fehlend` bleibt aus | trägt |

## 7. Offen mit Adresse

- V-1, V-2: Planner (Fix durch den Implementer als Schleife über die Klasse, oder Negativ mit Grund und Folge); in §6 des Plans mit Ausgang *offen* einzutragen (`AGENTS.md` §3.13).
- F-15 (Länge des Ortes im `Error()`): Architect und Planner, steht bereits in §6.
- Register: V-1 als weiterer Treffer von `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`; Planner.

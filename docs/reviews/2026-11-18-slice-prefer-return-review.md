# Review-Report: slice-prefer-return — 2026-11-18

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-prefer-return`, Zweig `slice-prefer-return`, `5a98ef2a..6ecf2279`; Pfade `src/hexagon/model/rueckgabe.go`, `src/hexagon/services/mock/schreiben_antwort.go`, `src/hexagon/services/mock/schreiben.go` samt Tests (`rueckgabe_test.go`, `prefer_test.go` in `mock`, `hybrid` und `httpodata`, `prefer_ende_zu_ende_test.go`), Plan-Abschnitte §3 (Tabelle *Zusage · Fänger* Zeilen 1 bis 26, Mengen, Mutationen und Raten), §4 (Marke) und §6 (Raster, Zusatztabelle, offene Rückfragen).

**Skill:** `.harness/skills/reviewer.md` @ `6ecf2279` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- Slice-Plan `slice-prefer-return` (Marke 3500, Messung 1826 nach `git diff --numstat` über die Pfade des Blocks)
- `SPEC-285` bis `SPEC-289` (Vertrag), `SPEC-194`, `SPEC-195` Nr. 11, `SPEC-192`, `SPEC-200` Nr. 1, 5, 6
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-07.b`
- `AGENTS.md` §3.7, §3.9 bis §3.18

**Vorgehen und Belege.** Gelesen: der gesamte Produktcode des Diffs (rund 260 Zeilen) und alle neuen Tests. Gefahren (Image `odata-recorder:race`, `--network none`, `--memory=4g`, `--cpus=4`, unter `flock` auf `.wt/heavy.lock`, Mutationen in einer Kopie im Container):

1. Eigene Mutationslisten über `tools/harness/mutationslauf.sh` (Klasse `andere`, ein Prozess je Stufe; 96 Mutationen in sieben Läufen, Last des Hosts 5 bis 9 von 4 CPUs, gemessen vom Reviewer): 85 rot oder rot-paket, 11 grün (Aufstellung unten). Keine Mutation der Liste des Implementers wurde übernommen; die Namen der Kennungen (`R`, `D`, `W`, `G`, `H`, `A`, `B`, `T`, `Q`) sind die des Reviewers.
2. `src/bootstrap` läuft im Mutationslauf nicht (`Ausgangsstand rot: src/bootstrap`, `SPEC-210` Nr. 5 e); die Mutationen gegen `TestPreferEntityGleichGet` und `TestPreferMinimalEntityId` fuhr der Reviewer in einem Container mit dem ganzen Baum (`go test ./src/bootstrap`, netzlos, `perl -0pi` auf einer Kopie): E1 und E2 rot, E3 und E4 grün.
3. Zeile 23 (nebenläufig): Mutation D21 (ETag der Version vor der Mutation) 20 von 20 Prozessen rot ohne Detektor (`prozesse=20`, `-test.count=1`); die Wirkung ist Rechnung, kein Rennen. Die Mutation W04 (Sortierung der Namen entfernt, Reihenfolge der Map) 20 von 20 rot.
4. `make groessenmarke`, `make docs-check`, `make zeichen-check`, `make a-check`: grün. Marke 3500, Messung 1826.
5. Rohbytes gegengelesen (`LC_ALL=C grep -cP '\\u[0-9A-Fa-f]{4}'`): alle elf Dateien des Diffs, kein Treffer; keine Direktive `//nolint`, kein `d-check:ignore` im Diff.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Zusage „ein Backslash im Wort in Anführungszeichen ersetzt sich durch das folgende Zeichen, der Wert gilt nach dem Auflösen“ (`SPEC-285` Nr. 2 und 3) hat keinen Fänger. Die Mutationen R15 (Zweig abgeschaltet), B1 (Zeichen gestrichen) und B2 (Backslash bleibt im Wert) bleiben in der ganzen Suite des Pakets `model` grün; `TestLiesRueckgabeAnfuehrung` führt den Backslash nur im Element (`elementEnde`), `TestLiesRueckgabeBeispiele` und `TestLiesRueckgabeSchreibweise` führen kein Wort mit Backslash, und das Orakel in `TestLiesRueckgabeFolgen` endet bei Länge 5, `return="\minimal"` hat sechs Symbole. Folge: `return="\minimal"` liefert Keine statt Minimal, ohne dass etwas rot wird. | `AGENTS.md` §3.15 (ein Fall je Form der Eingabe, die der Code verschieden behandelt), §3.17 | `rueckgabe.go` · `case z[j] == '\\' && j+1 < len(z):`; Tests `rueckgabe_test.go` · `TestLiesRueckgabeAnfuehrung`, `TestLiesRueckgabeFolgen` | ja — Mutationen R15, B1, B2 (Liste 1 und 2), Paket grün | Fänger führt nur eine Position der Menge (Backslash im Element, nicht im Wort) |
| F-2 | MEDIUM | Zeile 24 (`Prefer` geht unverändert in den Fallback) führt nur Anfragen ohne Body: `anfrage(methode)` setzt weder `Body` noch `Laenge`, damit läuft der Zweig mit gepuffertem Body von `vorbereiten` (`m, r = a, a`) nie. Die Mutationen H04 (Header der Anfrage an den Fallback leer) und H05 (Header der Anfrage an den Mock leer) bleiben grün; sie träfen genau die Mutationen `POST`, `PATCH` und `PUT` mit `Prefer`. Der Kopf des Tests sagt „Methode 5“ und nennt den Fall mit Body nicht. | `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe), §3.11 | `src/hexagon/services/hybrid/puffer.go` · `m, r = a, a`; Test `prefer_test.go` · `a := anfrage(methode)` | ja — Mutationen H04 und H05, Paket grün; H03 (Zweig ohne Body) ist rot | Fänger führt nur eine Form der Eingabe (ohne Body) |
| F-3 | MEDIUM | Die Zusage „ein leerer Strom ist der Body der `204`, nie `nil`“ (`SPEC-286` Nr. 1, `SPEC-194` Nr. 3) hat für die neue Antwort `204` mit `return=minimal` keinen Fänger. Die Mutation D27 (`Body` ist `nil`, nur bei Minimal) bleibt im Paket `mock` grün, E3 im Test mit echtem Schreiber ebenso: `TestPreferAntwortTabelle` liest den Body nur, wenn er nicht `nil` ist, und wertet `nil` wie leer. Folge: der HTTP-Adapter beantwortet `a.Body == nil` mit `ODR-E9001` und Status 500, ein Client mit `return=minimal` erhielte an jedem `POST`, `PATCH` und `PUT` einen Fehler. | `AGENTS.md` §3.15 (Zusage ohne Fänger) | `schreiben_antwort.go` · `return inbound.Antwort{Status: 204, Header: h, Body: io.NopCloser(bytes.NewReader(nil))}, nil`; Test `prefer_test.go` · `if err == nil && ant.Body != nil` | ja — Mutationen D27, E3 | Zusage der Tabelle ohne Fänger (Body nie `nil`) |
| F-4 | MEDIUM | Die Zusage „die Zeit wächst linear mit der Gesamtlänge“ (`SPEC-285` Nr. 1) hat weder Fänger noch ein Negativ, das sie nennt (`SPEC-285` *Ungeprüft* nennt nur die Länge über der Grenze des Adapters). Die Mutation Q1 (je Element eine Kopie der ganzen Zeile, quadratisch) bleibt grün; der Kommentar von `LiesRueckgabe` sagt „braucht Zeit linear in der Gesamtlänge“ zu. | `AGENTS.md` §3.15, §3.11 (ein Kommentar sagt nur zu, was ein Test prüft) | `rueckgabe.go` · `LiesRueckgabe liest die Wahl aus den Zeilen` | ja — Mutation Q1, Paket grün | Zusage ohne Fänger und ohne Negativ (Laufzeitklasse) |
| F-5 | LOW | Das Raster in §6 hat keine Spalten *Abbruch* und *Wiederholung*, und die Tabelle in §3 hat keine Zeile für `SPEC-287` Nr. 7 (dieselbe Mutation zweimal) und `SPEC-289` Nr. 1 (gleiche Anfrage, gleiche Bytes). Verhalten und Struktur sind gedeckt: ein Zustand zwischen Aufrufen braucht eine Paketvariable, und die Mutation D28c wird von `TestPaketOhneZustand` rot (nicht von einem Test des Vertrags); mit `Dienst.Etag` fangen ältere ETag-Tests D28; E4 (veralteter Eintrag in der Antwort, echter Schreiber) bleibt grün. | `AGENTS.md` §3.16 (Raster *Operation × Argumentart* mit Abbruch und Wiederholung), §3.15 | `docs/plan/planning/in-progress/slice-prefer-return.md` · `**Raster *Operation × Argumentart***` | nein — Form des Plans; Mutationen D28 und D28c nur strukturell oder durch ältere Tests | Raster ohne Spalte (Abbruch, Wiederholung) |
| F-6 | LOW | Der Kommentar von `kopf204` nennt die Funktion „der Header einer 204 ohne Wahl“; sie ist die Grundlage jeder `204`, auch mit `return=minimal` (`antwortOhneBody` ergänzt auf ihrem Ergebnis). | `AGENTS.md` §3.7 (ein Kommentar beschreibt, was da ist) | `schreiben_antwort.go` · `kopf204 ist der Header einer 204 ohne Wahl` | nein — Lektüre | Kommentar schmaler als die Verwendung |
| F-7 | LOW | Der Kopf von `TestPreferSchreiberFehler` sagt „Stelle 13 × Fehler 5“; der Test fährt 14 Stellen (`len(stellen) != 14`, `n != 14*5`) und der Plan nennt 14. | `AGENTS.md` §3.11 | `src/hexagon/services/mock/prefer_test.go` · `Stelle 13 × Fehler 5` | nein — Lektüre | Zahl im Kommentar nicht nachgezogen |
| F-8 | LOW | Die Prüfgrenze in §3 nennt als gleichwertig nur die Mutation „`tchar` ab 0x80“. Gleichwertig und grün sind alle Änderungen der Menge der `tchar`, die keinen Trenner und keinen Buchstaben von `return`, `minimal`, `representation` betreffen: T1 (`!` entfernt) und T2 (`Z` entfernt) bleiben grün, weil ein Zeichen außerhalb des Tokens und ein anderes Token das Element in beiden Fällen unlesbar oder ohne Wirkung machen. Grund und Folge des Negativs stehen im Plan für eine kleinere Menge als die tatsächlich ungefangene. | `AGENTS.md` §3.15 (akzeptiertes Negativ nennt Grund und Folge), §3.17 (Menge aus dem Wortlaut) | `docs/plan/planning/in-progress/slice-prefer-return.md` · `die Mutation „`tchar` ab 0x80“ ist deshalb gleichwertig` | ja — Mutationen T1 und T2 | Negativ schmaler als die ungefangene Menge |
| F-9 | LOW | Der Eintrag `welle-etag` der Roadmap nennt `slice-prefer-return` weiter „als Nächstes startbar“, obwohl der Slice in `in-progress/` liegt; die Zelle sagt den Zustand nicht. | `AGENTS.md` §3.7 (Zustandsfelder nennen den Zustand), §3.9 | `docs/plan/planning/in-progress/roadmap.md` · `als Nächstes startbar [slice-prefer-return]` | nein — Lektüre | Zustandsfeld nicht nachgezogen |
| F-10 | INFO | Zeile 26 (Handbuch-Beispiel nicht bindbar): Adressierung geprüft und getragen. §6 trägt die Zeile mit Ausgang *offen* und den Adressen Architect (Vertrag `SPEC-200`) und Planner (Punkt der DoD), im selben Commit wie der Plan, wie `AGENTS.md` §3.13 es für Rückgaben an eine Rolle verlangt; der Nehmer ist keine Slice-Datei, ein Ausschluss in §1 eines Nehmers entfällt. Die Aussage stimmt gegen `SPEC-200`: `K1` verlangt ein Beispiel mit Argumenten aus `A-Z a-z 0-9 . _ = -` gegen das Binary (Nr. 5, 6 a). Für die Entscheidung des Architect: das Handbuch beschreibt `ETAG_ENABLED` schon heute mit HTTP-Verhalten (Kopfzeile, `@odata.etag`, `412`) in einer Zeile der Schlüsseltabelle ohne `K1`-Beispiel; ein `Prefer`-Schlüssel gibt es nicht, ein Anker wie dort fehlt. Der DoD-Punkt „Handbuch-Kapitel“ bleibt offen, der Slice schließt damit erst nach der Antwort. | `AGENTS.md` §3.13; `SPEC-200` Nr. 1 | `slice-prefer-return.md` · `Handbuch-Kapitel der DoD` | nein — Entscheidung von Architect und Planner | — |
| F-11 | INFO | Zeile 21: die Mutationen G-S24, G-S25, G-S26 (Gegenstand `Workspace` statt `Sonstiges` je Stelle) bleiben in `TestPreferSchreiberFehler` grün und werden nur von älteren Tests des Pakets rot (`S24 bereits vorhanden`, rot-paket). Die Zeile des Plans nennt den Gegenstand nicht als Zusage der Prefer-Tests; `SPEC-288` Nr. 3 nennt ihn je Stelle, und `AGENTS.md` §3.15 nennt den Gegenstand als Menge je Aufrufstelle. Gedeckt, aber nicht vom benannten Fänger. | `AGENTS.md` §3.15 | `prefer_test.go` · `func TestPreferSchreiberFehler` | ja — Mutationen G-S24 bis G-S26 | Fänger der Tabelle schmaler als der Vertrag, gedeckt durch ältere Tests |
| F-12 | INFO | Der Kopf von `TestPreferAntwortUnveraendert` sagt „byte-gleich“ und prüft die Namen in der kanonischen Schreibweise des Clients (`Odata-Entityid`); `setzeHeader` schreibt mit `Header.Add`, die Namen gehen kanonisiert aus (Bestand, nicht dieser Diff). Die Werte sind byte-gleich, die Schreibweise des Namens ist nicht zugesagt und nicht geprüft. | `AGENTS.md` §3.11 | `src/adapters/driving/httpodata/prefer_test.go` · `byte-gleich aus` | nein | Zusage im Kopf breiter als die Prüfung (Namen) |
| F-13 | INFO | `tokenZeichen` in `rueckgabe.go` wiederholt `istTokenByte` aus `redaction.go` (gleiche Menge, selbes Paket); beide Funktionen sind gleichwertig. Keine Wirkung, ein Ort mehr bei einer Änderung der Token-Menge. | Maintainability | `rueckgabe.go` · `func tokenZeichen(b byte) bool` | nein | Doppelte Hilfsfunktion im Paket |

---

## Auflösung der Tabelle (Zeilen 1 bis 26) und des Rasters mit je eigener Mutation

Eigene Mutationen (die Kennungen sind die des Reviewers, nicht die des Implementers); „rot“ heißt `--- FAIL` des benannten Tests, „rot-paket“ ein anderer Test des Pakets bei grünem benanntem Test, „grün“ beide grün.

| Zeile | Eigene Mutation | Ergebnis |
|---|---|---|
| 1 | R16 Wort schließt nie; R20 Rest hinter dem Wort ohne Wirkung | rot `TestLiesRueckgabeBeispiele` (beide) |
| 2 | R01 Name byte-genau; R02 Wert byte-genau | rot `TestLiesRueckgabeSchreibweise` |
| 3 | R03 Tabulator nicht Leerraum hinter dem Namen | rot `TestLiesRueckgabeLeerraum` |
| 4 | R08 spätere Instanz zählt; R09 Zeilen rückwärts; R10 nur die erste Zeile; R22 Zeilen ohne Trenner verbunden | rot (R09 rot-paket, die anderen rot) |
| 5 | W01 Namen absteigend; W02 nur `Prefer`; W03 nur der erste Wert je Name; W04 Sortierung entfernt (Rate 20/20 ohne Detektor); D26 nur die erste Zeile gelesen | rot `TestPreferNamenUndReihenfolge` (alle) |
| 6 | R13 Paar im Element nicht verfolgt; R14 Backslash im Element ohne Wirkung; R21 Zeilen vor dem Lesen verbunden | rot `TestLiesRueckgabeAnfuehrung`; R15, B1, B2 (Backslash im Wort) **grün (F-1)** |
| 7 | R11 `;` nach dem Wort nicht zugelassen; R12 nur `;` nach dem Wort | rot `TestLiesRueckgabeRest` (R12 rot-paket) |
| 8 | R04, R05, R06 Komma, Semikolon, Gleichheitszeichen am Ende des Namens; R07 Ende des Namens immer; R18 Name als Präfix lesbar; R19 `return` als Präfix | rot `TestLiesRueckgabeUnlesbar` und `TestLiesRueckgabeFolgen`; T1 und T2 (tchar) **grün, gleichwertig (F-8)** |
| 9 | R17 Index ohne Grenze; B3 Backslash am Ende ohne Grenze | rot `TestLiesRueckgabeFolgen` (Panik); Q1 quadratisch **grün (F-4)** |
| 10 | D01 `POST`·Entity `200`; D02 `PATCH`·Entity `201`; D03 `204` als `200`; D04 und D05 `Vary`; D06 `Content-Type` an der `204`; D07 `ETag` bei aus; D08 `Preference-Applied` ohne Wahl; D09 Minimal-Zweig immer; D10a bis D10c `OData-EntityId` und `Location` je Zelle; D22 nur `PATCH`·Entity; D23 nur `PUT`·Entity; D24 `PUT`·Minimal ohne `OData-EntityId` | rot `TestPreferAntwortTabelle` (alle); D25 (`ETag` an `DELETE`) rot-paket; D27 (`Body` `nil` bei Minimal) **grün (F-3)** |
| 11 | D11 `OData-EntityId` ungleich `Location`; D11b Wurzel leer; D11c Typname statt Set; D29 Wurzel ohne Dienstroot; E2 dasselbe im Test mit echtem Schreiber | rot `TestPreferAdresseNieAusHost`, `TestPreferMinimalEntityId` |
| 12 | D12 falscher Wert bei Minimal; D13 Schreibweise des Wertes bei Entity | rot `TestPreferApplied` |
| 13 | D14 `DELETE` liest `Prefer` | rot `TestPreferAppliedNurBeiErfolg` |
| 14 | D15 Wert ohne Wahl ist ein Fehler (vor Schritt (b)) | rot `TestPreferFehlerGleich` |
| 15 | D16 zweites `Entities.Get` nach `Put`; D16b `Advance` ausgelassen bei `Prefer` | rot `TestPreferZustand` |
| 16 | G01 `GET` und `HEAD` mit `return=minimal` ein Fehler; D14 | rot `TestPreferLesendOhneWirkung` |
| 17 | D15 | rot `TestPreferFehlerGleich` |
| 18 | D21 ETag der Version davor; D30 Entity ohne letzte Property (echter Schreiber); E1 dasselbe | rot (`TestPreferGleichzeitig` 20 von 20; `TestPreferEntityGleichGet`) |
| 19 | D17a `SchreibeAnzahl` bei `POST`·Entity; D17b `Adresse` bei Entity | rot `TestPreferSchreiberAufrufe` |
| 20 | D18a `Accept` nicht weitergereicht; D18b `Auswahl` gesetzt; D18c Wurzel verändert; D18d und D18e `context.Background()` an S24 und S26; D11c Set an S25 | rot `TestPreferSchreiberArgumente`, `TestPreferAdresseNieAusHost` |
| 21 | D19a, D19b, D19c Fehler jeder Stelle (auch der Kontextfehler) als `ODR-E9001`; G-S24, G-S25, G-S26 Gegenstand | rot `TestPreferSchreiberFehler`; Gegenstand **rot-paket (F-11)** |
| 22 | D20a ohne Prüfung des Wertes; D20b ohne Prüfung des `ContentType`; D20c ohne Prüfung des Namens | rot `TestPreferDarstellungLeer` |
| 23 | D21 (Wirkung, 20 von 20 Prozessen ohne Detektor) | rot `TestPreferGleichzeitig`; der Plan nennt den Zustand innerhalb der Transaktion als Negativ mit Grund und Folge |
| 24 | H01 `Preference-Applied` gestrichen; H02 angefügt; H03 Header der Fallback-Anfrage leer (ohne Body) | rot `TestPreferGehtUnveraendertInDenFallback`; H04 und H05 (mit Body) **grün (F-2)** |
| 25 | A01 nur die erste Kopfzeile; A02 Zeilen umgekehrt; A03 `OData-EntityId` gefiltert; A04 `200` als `201`; A05 `204` als `200`; A06 `Preference-Applied` gefiltert; A07 `Location` gefiltert | rot `TestPreferHeaderUnveraendert`, `TestPreferAntwortUnveraendert` |
| 26 | nicht mutierbar: kein Beispiel, kein Fänger (offen, §6) | siehe F-10 |

**Raster (Zeile je Operation, Zelle je Argumentart).** Je Operation und Zelle eine der Mutationen oben, deren Wirkung genau diese Zelle trifft; kein Ort der Zelle liegt im Code:

| Operation | fehlt | minimal | representation | leer, unbekannt | mehrere, widersprüchlich | fehlgeschlagen |
|---|---|---|---|---|---|---|
| POST | D08 rot | D10b, D09 rot | D01, D10a rot | R08, D26 rot | R08, R09, W01, W03 rot | D15 rot |
| PATCH | D08 rot | D10c, D24 rot | D02, D22 rot | R08 rot | wie POST | D15 rot |
| PUT | D08 rot | D24 rot | D23 rot | R08 rot | wie POST | D15 rot |
| DELETE | D14 rot | D14 rot | D14 rot | D14 rot | D14 rot | D15 rot (`TestPreferAppliedNurBeiErfolg`; `TestPreferFehlerGleich` übergeht `DELETE`) |
| GET, HEAD | G01 rot | G01 rot | kein Zweig, nicht mutierbar | G01 rot | G01 rot | gilt nicht (Plan) |

Die Zusatztabelle (Modi): `mock` mit und ohne `ETAG_ENABLED` D07, D21, D25 rot; `hybrid` H01 bis H03 rot, H04 und H05 grün (F-2); `replay` und `record`: der Slice ändert den Pfad nicht, Negativ im Plan mit Grund (geprüft: kein Diff in den Paketen).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/rueckgabe.go` (Grammatik gegen `SPEC-285` Nr. 1 bis 3, gelesen und gegen die Tabelle der Beispiele gehalten) | geprüft, Befunde F-1, F-4, F-8, F-13 |
| `src/hexagon/services/mock/schreiben_antwort.go`, `schreiben.go` (Antwort je Zelle, Aufrufe des Schreibers, Lesen erst in Schritt (l), `DELETE`) | geprüft, Befunde F-3, F-6; Antwort je Zelle von `SPEC-286` Nr. 1 und Zahl der Aufrufe von `SPEC-288` Nr. 3 ohne Abweichung |
| `src/hexagon/services/hybrid/` (nur Tests im Diff) | geprüft, Befund F-2 |
| `src/adapters/driving/httpodata/` (nur Test im Diff) | geprüft, Befund F-12 |
| `src/bootstrap/prefer_ende_zu_ende_test.go` | geprüft, ohne Befund außer F-3 (E3 grün) |
| `schreiben_antwort_test.go` (Eingrenzung von `TestPreferIgnoriert`) | geprüft, ohne Befund |
| Plan `slice-prefer-return` (Kopf, §1 bis §6, Marke, Raster) | geprüft, Befunde F-5, F-8, F-10 |
| `docs/plan/planning/in-progress/roadmap.md` | geprüft, Befund F-9 |
| Hard Rules §3.1 (Docker-only), §3.2 (keine Suppression), §3.18 (Rohbytes) | geprüft, ohne Befund |
| Sicherheit und Geheimnisse (kein Header der Anfrage in Fehlertext, Adresse nie aus `Host`) | geprüft, ohne Befund (`TestPreferFehlerGleich`, `TestPreferAdresseNieAusHost`, Mutationen D15, D11 bis D11c rot) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 |
| LOW | 5 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Position oder Form der Eingabe · Zusage der Tabelle ohne Fänger · Zusage ohne Fänger und ohne Negativ (Laufzeitklasse) · Negativ schmaler als die ungefangene Menge · Raster ohne Spalte (Abbruch, Wiederholung) · Kommentar schmaler als die Verwendung · Zustandsfeld nicht nachgezogen

## Verdikt

**Merge-blockierend:** ja — vier MEDIUM (F-1 bis F-4), alle vom Implementer mit je einer Mutation an der benannten Stelle rot zu machen oder als akzeptiertes Negativ mit Grund und Folge zu führen; die Klasse bei F-1 ist „Backslash im Wort“ (Wort vor und nach dem Auflösen, jede Stelle des Wortes), bei F-2 die Formen von `vorbereiten` (ohne Body, mit Body, über der Grenze) je Stufe. Zeile 26 bleibt offen und wartet auf Architect und Planner (F-10); sie blockiert den Handoff an den Verifier nicht, den Schluss des Slice schon.

**Übergabe:** Findings gehen an den Implementer; die **Finding-Klassen** gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ist ein **Lauf-Beleg**; er ersetzt keine Verifikation (DoD und Spec prüft der Verifier separat).

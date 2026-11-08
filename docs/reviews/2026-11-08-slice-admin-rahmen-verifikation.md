# Verifikation: slice-admin-rahmen — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-238` Nr. 1 bis 3 a, 5, 6, 14, `SPEC-240` und die Hard Rules; nicht Diff-Review (siehe `2026-11-07-slice-admin-rahmen-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-admin-rahmen`, HEAD `0d49e4e0` (Implementer-Stand `0c13cfa4`/`f93fd0ab`, Review `a69a4236`, Fix-Runde `0d49e4e0`), Basis `0504b729`. Plan: `docs/plan/planning/in-progress/slice-admin-rahmen.md`. Angrenzende ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Rolle/Modell:** verifier · claude-sonnet-5-5. `make docs-check` über diesen Report: siehe Abschnitt 1.

## Urteil

**Bestanden mit Auflagen an den Planner; kein HIGH, keine DoD-Verletzung im Code.** `make gates` ist grün (eigener Lauf), die Größenmarke hält (1785 von 2400), der Produktcode setzt `SPEC-238` Nr. 1, 2, 3 a, 5 (404, 405) und `SPEC-240` richtig um, das Binary bleibt unverändert. Die drei Befunde des Reviews sind für die genannten Klassen gedeckt: Z3 vollständig (7 von 7 fremden Mutationen rot). Z4 und Z8 sind für die genannten Mutationen gedeckt, aber die Klasse ist enger gezogen, als die Zusage reicht: sieben fremde Mutationen an Z4 und vier an Z8 bleiben grün (drei MEDIUM, ein LOW, siehe Abschnitt 4). Keine ist ein Produktfehler heute; sie sind Lücken im Fänger. Empfehlung: kein zweiter Fix-Lauf nötig, wenn der Planner die Lücken als akzeptierte Negative mit Grund und Folge in §3 einträgt oder an den Folge-Slice `slice-admin-authentisierung` adressiert (Auflage A1). Fix-Runde nur, wenn der User die Fänger geschlossen haben will.

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/gates.log`: `EXIT=0`; darin `baseline-verify` OK (54 Dateien), `d-check: 849 Datei(en) geprüft, 0 Befund(e)`, `migration-check: ok`, `zeichen: ok, 1715 Datei(en)`, Tests und Detektor ohne Rot. `TestWettlaufNebenlaeufig` (Teil 1) lief in diesem Lauf grün; der Flake (Plan §6, Fund Teil 1) ist nicht Gegenstand. |
| `make groessenmarke` | im selben Lauf: `slice-admin-rahmen.md: ok, Messung 1785 von Marke 2400`; eigene Rechnung `git diff --numstat 0504b729..HEAD` über die acht Pfade des Blocks: Summe 1785 (0,74 der Marke, 1,02 der Schätzung 1750). |
| Plan-vs-Code-Diff | `git diff --stat main...HEAD`: 19 Dateien. Produktcode: `admin.go` (176), `route.go` (163), `fehler.go` (+2); Katalog `harness/katalog-ausnahmen.txt` (+1); `spec/spezifikation.md` (4 Zeilen, Architect, vor dem Code); der Rest Tests, Review-Report und Plan. Alle Pfade stehen im Block. `src/bootstrap` und `tests/` unverändert (`git diff --stat 0504b729..HEAD -- src/bootstrap` leer; N-3 gehalten). |
| Zweige unter dem Detektor | eigenes `go test -coverprofile` im Container (`odata-recorder:test`, netzlos, `--memory=4g`): Paket 98,4 %, außerhalb von `tokendatei.go` genau ein Block mit Zähler 0: `route.go:128.4,129.1` (Z-5, N-4). Das ist der im Plan benannte eine Ausreißer. |
| Merge-Nähe | `git merge-tree --write-tree HEAD main`: Rückgabe 0, kein Konflikt. `main` trägt 39 Codes (`ODR-E6009` der 39.), `ODR-E2102` ist hier der 40.; die Zählstellen der fünf Test-Dateien folgen `model.AlleCodes()` (keine Literalzahl). Der Zweig steht 36 Commits hinter `main`. |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P` auf BOM, NBSP, Nullbreitenzeichen und Backslash-u-Folge über `src/adapters/driving/admin/*.go` (außer `tokendatei*`, Teil 1): 0 Treffer. `make zeichen-check` ist in `make gates` grün. |
| Betriebsregeln | Docker ad hoc nur mit `--memory=4g --memory-swap=4g`, unter `flock`, ein Container zur Zeit, nie parallel zu `make gates`; Mutationen in einer Wegwerfkopie im Container (Quellen schreibgeschützt eingehängt); keine Repo-Datei außer diesem Report geschrieben. Hinweis: das Image-Tag `odata-recorder:test` ist über alle Worktrees geteilt und kann den Stand eines anderen Zweigs tragen; deshalb hängen die Läufe oben die Quellen dieses Zweigs ein, statt das Image zu lesen. |
| `make docs-check` über diesen Report | Zeile im Commit-Bericht des Verifiers (nach dem Commit gefahren). Kennungen von ADRs stehen als auflösender Link. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Konfiguration, Routen und Zuordnung nach `SPEC-238` Nr. 1, 2, Ablauf 3 a, `Anfrage`, 404, 405, `bindung_test.go` | **im Code erfüllt, Haken fehlt** | `admin.go`, `route.go` gelesen und gegen den Wortlaut gehalten (Abschnitt 3); Bindungs-Test vorhanden (`bindung_test.go`, 20 Zeilen). |
| Code `ODR-E2102` samt Katalog-Ausnahme, Zählstellen im selben Commit | **erfüllt, Haken fehlt** | `fehler.go` (Zeile 400, 404, 405, `admin`, `Admin request invalid`, hinter `ODR-E2101`), `fehler_test.go`, `regeln_test.go`, `audit_test.go`, `lesen_fehler_test.go`, `httpodata/fehler_test.go`, `harness/katalog-ausnahmen.txt`. `make katalog-check` ist grün (im `make gates`). Der Commit `2ae5d0da` trägt Code und Ausnahme gemeinsam. |
| `make gates` grün | erfüllt (eigener Lauf, Abschnitt 1) | `EXIT=0`. |
| Review, Report unter `docs/reviews/` | erfüllt (Haken fehlt) | `2026-11-07-slice-admin-rahmen-review.md`, 0 HIGH, 3 MEDIUM, 2 LOW; MEDIUM 1 bis 3 und LOW 5 in `0d49e4e0` bearbeitet; LOW 4 (Zahlen) im Plan berichtigt. |
| Doku-Update öffentlicher Vertrag; Handbuch | erfüllt | Der Vertrag liegt in `spec/` (Architect). Das Handbuch bleibt unverändert, `ODR-E2102` steht in der Ausnahmedatei, `make handbuch-check` und `make katalog-check` sind grün. Das Binary verdrahtet den Listener nicht (N-3). |
| Closure-Notiz, Register, Ausgänge der Risiken, drei Paarungen | offen, Sache des Planners | §7 trägt „Wird bei der Closure gefüllt“; §6 hat Zeilen mit Ausgang *offen* (Abschnitt 5). |

Die Liefer-Punkte (drei) halten das Limit.

## 3. Tabelle *Zusage · Fänger* gegen den Vertrag

Jede Zeile des Plans gegen `SPEC-238`; das Orakel der Tests ist ein Literal oder Modell des Tests, nicht der Code.

| Zeile | Vertrag | Befund |
|---|---|---|
| K1 bis K3 | Nr. 1: nil-Felder, Token ab 32 Bytes, `New` ruft nichts, öffnet nichts, startet nichts | Code stimmt: `New` prüft vier Felder und Länge, ruft nur `baueRouten`, `slices.Clone`. Fänger vorhanden (`konfig_test.go`); N-7 (kurzlebige Goroutine in `baueRouten`) ist benanntes Negativ mit Grund und Folge. |
| Z1 | Nr. 2 a bis c | Code `gueltigeMethode`, `parseSegment`, `parseMuster` stimmen mit dem Wortlaut (Zeichenmengen, Namen `[a-z][a-z_]*`, keine doppelten Namen, kein abschließender `/`). Fänger: 786 Fälle, Orakel Literale; der Reviewer hat 8 Elemente fremd gefahren (rot). |
| Z3 | Nr. 2 d | Code `ueberlappt` stimmt. Fänger 14520 Paarfälle + 4 mit Modell je Segment; eigene Klassenprobe: Abschnitt 4, 7 von 7 rot. |
| Z4 | Nr. 1, 2 Zuordnung | Code `zerlegePfad` (`EscapedPath`, je Segment `PathUnescape`), `passt` stimmen. Fänger deckt die genannten Formen; Lücken in Abschnitt 4. |
| Z5 | Nr. 2 kein Bereinigen, führender `/` | Code stimmt; Fänger 16 Fälle. |
| Z6 | Nr. 2, 5 (404, 405, `Allow`, Fehlerobjekt) | Code stimmt (`Allow` nach Bytefolge, `, `, Methode bytegleich, eingebaute Route zählt mit). Fänger 32 + 7, Literale. |
| Z7 | Nr. 2 `Anfrage`, 3 a | Code stimmt: `KorrelationsID()` ist die erste Anweisung von `ServeHTTP`, `Pfadwerte` nie nil, `Laenge` aus `ContentLength` (-1). Fänger vorhanden. Bekannt und im Plan benannt: N-6 (kein `X-Correlation-Id` in diesem Teil), `Konfig.Vergleich` gehört zu Teil 3. |
| Z8 | Nr. 2 `Header` ohne `Authorization`, Kopie | Code stimmt für einzelne Werte; Lücken in Abschnitt 4. |
| Y2 bis Y4 | `SPEC-240`, Nr. 14, `SPEC-230` Nr. 8 | Code, Tabelle und Katalog stimmen; die Zählstellen rechnen gegen `model.AlleCodes()` (keine feste 39 oder 40 mehr im Test). |

Abweichungen vom Plan im Code: keine. Der Plan ist im Fix-Commit nachgezogen (§3 Zeilen Z3, Z4, Z8, Absatz *Fix-Runde nach dem Review*, §6); die Zählungen stimmen (13 Mutationen = 5 + 3 + 5; 19 Werte = 7 + 12; 48 Nachbarn = 14 + 13 + 13 + 8).

## 4. Fix gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

Eigene Mutationen in einer Wegwerfkopie im Container (`odata-recorder:test` mit den Quellen des Zweigs schreibgeschützt eingehängt, netzlos, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock`), ein Prozess je Mutation, `go test -count=1 -run <Tests des Pakets ohne Teil 1>`. Messender: der Verifier. Last des Hosts: Load Average 4 bis 7 bei 4 CPUs je Container. Alle Fänger sind deterministisch (kein Zeit- oder Nebenläufigkeitsanteil): eine Einzelzahl je Mutation, nur Größenordnung. Basislauf grün. Das Skript liegt unter `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/mut.sh`. Die Mutationen sind Elemente, die weder der Befund noch die Fix-Runde nannte.

| Klasse (aus dem Wortlaut der Zusage) | Menge | Mutation (fremd) | Ergebnis |
|---|---|---|---|
| Z3 „an jeder Stelle“: Stelle × Länge | Stellen 1 bis 4 bei Längen 1 bis 4, 120 Muster | Mitte bei Länge 3 übersprungen; Stelle 2 bei Länge 2; Stelle 3 bei Länge 4; Stelle 1 bei Länge 3; Stelle 1 bei Länge 4; Stelle 4 bei Länge 4; alle ungeraden Stellen | **7 von 7 rot** (`TestRoutenUeberlappung`). Z3 ist als Klasse gedeckt. |
| Z4 Dekodierung je Segment | 19 Werte × 3 Parameterstellen | Wert nach dem Dekodieren ein zweites Mal dekodiert (Fehler ignoriert), also `%2541` ergäbe `A` | **grün**. Kein Wert der Menge hat nach dem ersten Dekodieren die Form `%XX` (kein `%25` gefolgt von zwei Hex-Ziffern). |
| Z4 „ein Parameter auf jedes nichtleere Segment“ | unbeschränkt (Länge, Zeichen) | Wert in ASCII-Kleinbuchstaben gewandelt (`X` ergäbe `x`) | **grün**: kein Wert der Menge trägt einen ASCII-Großbuchstaben. |
| ebenso | | Parameter nur bis 8 Bytes Länge; nur bis 64 Bytes | **grün**, beide: der längste Wert hat 5 Bytes. Weder Prüfgrenze noch Negativ im Plan (Register `unbeschraenkte-menge-ohne-benannte-pruefgrenze`). |
| Z4 „ein Literal passt bytegleich auf das dekodierte Segment“ | Segmente am Literal: gleich, Schreibweise, Prozentfolge, Segmentzahl ±1 | Vergleich durch `HasPrefix`; durch `HasSuffix`; durch `Contains` | **grün**, alle drei: kein Fall setzt am Literal ein längeres Segment ein, das mit dem Literal beginnt, endet oder es enthält (`ab` gegen `a`). Das Literal der Muster ist ein Zeichen (`a`, `c`), der Fall „Rest des Literals“ fehlt auch deshalb. |
| ebenso | | Vergleich nur des ersten Bytes | rot, aber nur in `TestZuordnungGrenze` (die Literale `r0` bis `r63` teilen das erste Byte), nicht in `TestZuordnung` |
| Z8 „`Authorization` erreicht keine Route“ | Authorization in jeder Form: Schreibweise, Zahl der Zeilen | `Authorization` mit zwei Zeilen (Slice der Länge 2) wird nicht gefiltert | **grün**. Der Fänger sendet `Authorization` nur mit einem Wert. Das Token erreichte die Route. |
| Z8 „jede andere Zeile bleibt“ | Werte der Nachbarn | Header mit leerem erstem Wert verworfen; Header mit leerer Liste verworfen; Namen ab 20 Zeichen verworfen | **grün**, alle drei: Nachbarn tragen immer den Wert `1`, der längste Name hat 19 Zeichen. |

Befunde daraus:

### 1. MEDIUM — Z8: `Authorization` mit mehreren Zeilen erreicht in einer Mutation die Route, der Fänger sendet nur eine

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 („jede Form der Eingabe, die der Code verschieden behandelt“), §3.17; `SPEC-238` Nr. 2 (`Header` ohne `Authorization`) und Nr. 13 (das Token erreicht keinen Ausgabeweg)
- `pfad`: `src/adapters/driving/admin/anfrage_test.go` · `func z8Faelle() []z8Fall` (`bearer := []string{"Bearer " + canary()}`); Code `admin.go` · `func anfrage`
- `befund`: Die Menge „Authorization in allen Schreibweisen“ (8192) trägt je Schlüssel genau einen Wert. Ein Filter, der nur einzeilige `Authorization` entfernt, bleibt grün; bei `Authorization` mit zwei Zeilen (auch mit gleichlautenden) liefe das Token an die Route. In diesem Teil prüft der Adapter `Authorization` noch nicht; sobald Teil 3 die mehrzeilige Form mit 401 abweist (`SPEC-238` Nr. 4 a), erreicht sie keine Route mehr, die Zusage Z8 trägt aber die Gegenprobe selbst.
- `verifizierbar`: ja (die Mutation oben, `go test` bleibt grün)
- `klasse`: Fänger einer Filter-Zusage deckt die Formen des Wertes (eine Zeile, mehrere) nicht

### 2. MEDIUM — Z4: das Literal wird nur gegen gleichlange Segmente geprüft; Prefix, Suffix, Teilstring bleiben grün

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 („beide Seiten jeder Grenze“), §3.17; `SPEC-238` Nr. 2 (Literal „bytegleich“)
- `pfad`: `src/adapters/driving/admin/zuordnung_test.go` · `func z4Generator() []z4Fall`, Muster `/a`, `/a/{id}`, `/{p}/c/{q}`; Code `route.go` · `func (r route) passt`
- `befund`: Am Literal prüft der Generator gleich, andere Schreibweise und Prozentfolge, aber kein Segment, das das Literal enthält (`ab`, `xa`, `xay`). `HasPrefix`, `HasSuffix` und `Contains` bleiben grün; `GET /ab` träfe die Route `/a`. Die Zusage „bytegleich“ ist die Allaussage dahinter.
- `verifizierbar`: ja
- `klasse`: Fänger der Gleichheit ohne die Nachbarn der Grenze (länger, kürzer, enthalten)

### 3. MEDIUM — Z4: Dekodieren genau einmal ist nicht gefangen; doppeltes Dekodieren bleibt grün

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.17 (die Klasse der Fix-Runde „Wertformen, in denen Dekodierer sich unterscheiden“), `SPEC-238` Nr. 2 („jedes Segment einzeln dekodiert“)
- `pfad`: wie 2; Code `route.go` · `func zerlegePfad`
- `befund`: Die zwölf neuen Formen decken Pfad- gegen Query-Dekodierung, aber keine Form, in der ein erstes Dekodieren wieder eine Prozentfolge ergibt (`%252F`, `%2541`). Zweimaliges Dekodieren (Pfadwert `%252e%252e` ergäbe `..`) bleibt grün. Das ist die Klasse der Fix-Runde („Dekodierung“) an einem Element, das sie nicht nannte, und sicherheitsrelevant für die Endpunkte, die den Wert als Namen nutzen.
- `verifizierbar`: ja
- `klasse`: Wertformen der Dekodierung ohne die Form „dekodiert ergibt wieder eine Prozentfolge“

### 4. LOW — Z4/Z8: unbeschränkte Mengen (Länge und Zeichen eines Parameterwertes, Länge und Wert eines Headers) ohne Prüfgrenze im Plan

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.17 („Ist die Menge unbeschränkt, nennt der Bericht die Prüfgrenze mit Grund und Folge“); Register `unbeschraenkte-menge-ohne-benannte-pruefgrenze`
- `pfad`: Plan §3, Absatz *Unbeschränkte Mengen*
- `befund`: Der Plan nennt als Prüfgrenze nur die Zahl der Routen (64) und der Parameter (4). Länge und Zeichen eines Parameterwertes (Mutationen: Länge 8 und 64, ASCII-Großbuchstaben) und Wert und Name eines Headers (leerer Wert, leere Liste, Name ab 20 Zeichen) sind ebenfalls unbeschränkt, ohne Grenze und ohne Folge im Betrieb. Der Code hat keine solche Grenze; eine künftige Änderung bliebe grün.
- `verifizierbar`: nein (Text des Plans)
- `klasse`: unbeschränkte Menge ohne benannte Prüfgrenze

Ein kleiner Fänger je Befund (je ein Fall: Parameterwert `X`, `%2541`, 200-Byte-Wert; Segment `ab`, `xa`; `Authorization` mit zwei Zeilen; leerer Header-Wert) schließt 1 bis 3; für 4 reichen Grund und Folge im Plan.

## 5. Risiken und Adressen (§6)

- Die drei Abweichungen des Implementers (Z6 sieben Fälle, Z4 achter Wert, Z3 neun Muster) stehen mit Ausgang *offen*, Adresse Architect. Z3 und Z4 sind seither erweitert (Fix-Runde), der Text in §6 nennt den Stand. Der Architect hat noch nicht bestätigt; er sollte sie mit Abschnitt 4 zusammen sehen.
- Zwei Punkte zur Zahl der Codes (Merge-Nähe, `slice-snapshot-use-case`, `slice-sqlite-interrupt-transaktion`): Ausgang *offen*, Adresse Planner (Integrationsschritt). Probe-Merge sauber (Abschnitt 1).
- Fund Teil 1 (`TestWettlaufNebenlaeufig`): offen, Adresse Architect, danach Fix-Slice. In meinem Gate-Lauf grün; ein rotes `make gates` aus diesem Grund bliebe kein Fehler dieses Slice (Reviewer INFO 2).
- Bekannte, im Plan benannte Lücken dieses Teils, die spätere Teile schließen: `X-Correlation-Id` (N-6), `Konfig.Vergleich`, Kopie des Tokens (K4), `Antwort.Status` außerhalb 200 bis 299 (`WriteHeader`-Panik bei `Antwort{}`, `slice-admin-fehlerabbildung`). Ich habe sie nicht als Befund gezählt.

## 6. Auflagen an den Planner

- **A1:** Befunde 1 bis 4 aus Abschnitt 4 entweder an einen Folge-Fix richten (Fänger je ein Fall, kleine Menge) oder als akzeptierte Negative mit Grund und Folge in §3 eintragen; Befund 1 (`Authorization` mit mehreren Zeilen) bietet sich für `slice-admin-authentisierung` an (A1, E4, Z8), sofern dessen §1 oder DoD ihn ausdrücklich annimmt (`AGENTS.md` §3.13, vorher dessen §1 *Ausdrücklich NICHT* und DoD lesen).
- **A2:** Die DoD-Haken (Konfiguration/Routen, Code `ODR-E2102`, `make gates`, Review, Handbuch) sind im Code erfüllt, aber nicht abgehakt; §7 und Register sind offen.
- **A3:** Nach dem Merge von `main` einmal `make gates` wiederholen; die Zahl der Codes folgt `model.AlleCodes()`.
- **A4:** Die Fix-Runde nennt „Last des Hosts: nicht gemessen“; das ist ehrlich, aber unter der Regel „eine Rate steht mit Last des Hosts“ eine Lücke; die Fänger sind deterministisch, ein Einzellauf genügt als Größenordnung (§3.10).

## Zusammenfassung

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 (Z8 mehrzeilige Authorization; Z4 Literal-Nachbarn; Z4 doppeltes Dekodieren) |
| LOW | 1 (Prüfgrenzen der unbeschränkten Mengen) |
| Eigene Mutationen | 19: 8 rot (Z3 7, erstes Byte 1), 11 grün (Z4 7, Z8 4) |

Geprüft, ohne Befund: `make gates` (grün), Größenmarke (1785 von 2400), Zweige unter dem Detektor (ein Ausreißer, benannt), Katalog und Handbuch, Binary unverändert, Probe-Merge mit `main`, Rohbytes, Zählstellen der Codes.

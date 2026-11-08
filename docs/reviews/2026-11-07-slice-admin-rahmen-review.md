# Review-Report: slice-admin-rahmen — 2026-11-07

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-admin-rahmen`, HEAD `f93fd0ab` (Commits `2ae5d0da`, `0c13cfa4`, `f93fd0ab`), Basis `0504b729`; Paket `src/adapters/driving/admin` (`admin.go`, `route.go`, Tests `konfig_`, `route_`, `zuordnung_`, `anfrage_`, `bindung_`, `fakes_`, `helfer_test.go`), `src/hexagon/model/fehler.go` und `fehler_test.go`, die Zählstellen in `audit_test.go`, `regeln_test.go`, `lesen_fehler_test.go`, `httpodata/fehler_test.go`, `harness/katalog-ausnahmen.txt`; Plan §3, §4, §6.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-07

**Eingangs-Kontext:**

- `docs/plan/planning/done/slice-admin-rahmen.md` (im Zweig): Tabelle *Zusage · Fänger* K1 bis K3, Z1, Z3 bis Z8, Y2 bis Y4, Mengen, Zweig-Tabelle, Absatz *Mutationen und Raten*, N-1 bis N-7, Raster und Abweichungen in §6
- `SPEC-238` Nr. 1 und 2 (Stand mit den Nachzügen des Architect), angrenzend `SPEC-239` Nr. 1 a und 3 d, `SPEC-240`
- `AGENTS.md` §3.2, §3.7, §3.10, §3.11, §3.13, §3.15 bis §3.18
- `LH-FA-13`, `LH-QA-04`
- ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) (nur genannt)

**Nicht Gegenstand, nicht gelesen als Befund:** DoD, `make gates` (Aufgabe des Verifiers). Nicht gelaufen: `go test -race -coverprofile` (die Zählung der Zweige und der eine Ausreißer `route.go:128` sind vom Implementer berichtet und hier nicht nachgemessen).

## Geprüft ohne Befund

- Keine Direktive `//nolint` im Paket (`grep`), keine Backslash-u-Folge, kein BOM, kein NBSP, kein Nullbreitenzeichen in `src/adapters/driving/admin/*.go` und `harness/katalog-ausnahmen.txt` (`LC_ALL=C grep -P`, 0 Treffer). Die Werte jenseits von ASCII in `z4Werte` entstehen aus `[]byte{...}`.
- Produktcode: keine feste Zahl der Codes. `model.CodeE2102` steht hinter `CodeE2101`, Status 400, 404, 405, Ziel `admin`, Text `Admin request invalid`; Tests in `audit_test.go` und `fehler_test.go` rechnen gegen `model.AlleCodes()` (`n := len(alle)`); die Suche nach `39` und `40` in den fünf Zählstellen findet keine Zahl der Codes mehr (die Treffer sind andere Größen). Die Zahl `+ 2` in `httpodata/fehler_test.go` ist die Zahl der Stati von `ODR-E2102` und gehört dorthin.
- Mengen und Zählprüfungen gelesen: M-Route 786 (`5 + 12 + 3 * 256 + 1`), M-Überlappung 90 + 4, M-Zuordnung 42 (`6 * 3 + 3 * 8`) und 16, M-Grenze 64 gegen die Länge der Liste, jeweils als Summe der Achsen aus Listen des Tests.
- Kein Aufruf von `admin` aus `src/bootstrap`; `git diff --stat main...HEAD -- src/bootstrap tests` leer (N-3 gehalten).
- `admin.go`: keine `go`-Anweisung (N-7 gelesen), kein `Listen*`, keine Umgebung, keine Datei; `New` kopiert `Token` und `Routen`; `ServeHTTP` ruft `ids.KorrelationsID()` einmal als erste Anweisung.
- Kommentare in `admin.go` und `route.go` tragen Zusage, Kopplung oder Grenze (§3.7); kein Konjunktiv über verworfene Alternativen gefunden.
- Fehlerobjekt wird an einer Stelle gebildet (`schreibeFehler`); Fehlertext von `New` nennt weder Methode, Muster noch Token (`ungueltig` mit festen Texten).

## Eigene Mutationen

Wegwerfkopie von `go.mod`, `src/`, `tests/` in einem Container (`odata-recorder:test`, ohne Netz, `--memory 3g --memory-swap 3g --cpus 4`, unter `flock`), `go test -count=1 ./src/adapters/driving/admin/` über das ganze Paket, ein Prozess je Mutation, Messender: der Reviewer, Last des Hosts 4 bis 10 von 4 CPUs. Alle Fänger sind deterministisch; die Zahlen sind eine Größenordnung (1 Prozess je Mutation), keine Rate. Klasse `andere`; die Klasse `nebenläufig` kommt in diesem Teil nicht vor (Plan §3, `SPEC-238` Nr. 12 liegt bei `slice-admin-abweisen-audit`). Das Paket war ohne Mutation grün (10,7 s).

32 Mutationen gefahren (eine weitere Sed-Fassung blieb wirkungslos und wurde neu gefahren). 23 rot, 9 grün: 5 davon Befunde 1 bis 3, 2 gleichwertig (`Methode` in Großschreibung, `if true` beim leeren Body), 2 bei Teil 3 und N-7 (`Token` ohne Kopie, Goroutine in `baueRouten`). Die Zeilen Z3 (Segmentzahl, Methode) und Y2 sind in der Tabelle als nicht neu gefahren gekennzeichnet.

| Zeile | Mutation (an einem Element, das der Plan nicht nannte, wo möglich) | Ergebnis |
|---|---|---|
| Z1 | `PUT` aus der Menge der Methoden | rot (`TestRoutenGutfaelle`, `TestRoutenUeberlappung`, `TestMethodeUndPfad`) |
| Z1 | Literal: `0`, `9`, `A`, `z` je ausgeschlossen (4) | je rot (`TestRoutenGrammatik`) |
| Z1 | Name: weiteres Zeichen `z`, erstes Zeichen `a` ausgeschlossen (2) | je rot (`TestRoutenGrammatik`) |
| Z1 | Prüfung des Suffixes `}` entfernt | rot (`TestRoutenGrammatik`, mit Panik `slice bounds out of range [1:0]`: rot, aber nicht nur als `--- FAIL`) |
| Z3 | Segmentzahl, Methode nicht verglichen | rot (siehe Plan, nicht neu gefahren) |
| Z3 | Parameter nur auf einer Seite (`||` statt `&&`) | rot (`TestRoutenUeberlappung`, `TestEingebauteRouteUeberlappt`) |
| Z3 | Vergleich nur der ersten zwei Segmente | **grün** (Befund 1) |
| Z3 | Vergleich ohne die mittleren Segmente | **grün** (Befund 1) |
| Z4 | `url.QueryUnescape` statt `PathUnescape` | **grün** (Befund 2) |
| Z4 | Segmentzahl `>` statt `!=` | rot (`TestZuordnung`, Panik) |
| Z4 | `Pfadwerte` nil statt leerer Map | rot (`TestAnfrageFelder`, Panik) |
| Z4 | `r.URL.Path` statt `EscapedPath` in der Zuordnung | rot (`TestAnfrageFelder`, `TestZuordnung`) |
| Z4 | `Body` der `Anfrage` nil | rot (`TestAnfrageFelder`, Panik) |
| Z4 | `if true` vor dem Schreiben eines leeren Body | gleichwertig (`Write` mit 0 Bytes) |
| Z5 | Prüfung des führenden `/` ersetzt durch `kodiert == "*"` | rot (`TestKeinBereinigen`, Panik) |
| Z6 | `Allow` auch bei 404 (leerer Wert gesetzt) | rot (fünf Tests) |
| Z6 | `Allow` mit doppelten Einträgen | rot (`TestMethodeUndPfad`, `TestMethodeSchreibweise`) |
| Z6 | `Content-Type` mit Zusatz `; charset=utf-8` | rot (fünf Tests) |
| Z7 | `ctx` des Aufrufs `context.Background()` | rot (`TestCtxDerRoute`) |
| Z7 | `Laenge` auf mindestens 0 | rot (`TestAnfrageFelder`) |
| Z7 | `KorrelationsID()` nur bei Treffer einer Route | rot (`TestKorrelationsID`) |
| Z7 | `Methode` der `Anfrage` in Großschreibung | gleichwertig (jede gültige Methode ist schon groß) |
| Z8 | Filter: jeder Name, der auf `authorization` endet (trifft `Proxy-Authorization`) | **grün** (Befund 3) |
| Z8 | Filter: jeder Name mit Präfix `auth` | **grün** (Befund 3) |
| K1 | Prüfung `Listener` entfernt | rot (`TestNewNilFelder`) |
| K2 | `tokenMinBytes` 33 statt 32 | rot (22 Tests, davon `TestNewTokenLaenge`; die Konstante gehört zu Teil 1) |
| K3 | `go func() {}()` in `baueRouten` | **grün**, N-7 wie im Plan (Folge dort benannt) |
| Y2 | Status der Fehlerzeile `ODR-E2102` | vom Implementer rot gesehen (`TestCodeTabelle`), hier nicht neu gefahren; die Schleife über Status 0 bis 999 im Test gelesen |
| Fehler der Route | `E9001` mit Status 502 statt 500 | rot (`TestAntwortDerRoute`) |
| `Token` | `token: k.Token` ohne Kopie | **grün**, gehört nach Plan zu K4 in `slice-admin-authentisierung` (INFO 1) |

## Findings

### 1. MEDIUM — Z3: der Fänger vergleicht nur Muster mit höchstens zwei Segmenten gegeneinander

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe), §3.17 (Fix deckt die Klasse)
- `pfad`: `src/adapters/driving/admin/route_test.go` · `func z3Muster() []string` (Anker `"/a/b/c"`); Code `route.go` · `func ueberlappt(a, b route) bool`
- `befund`: Von den neun Mustern hat genau eines drei Segmente (`/a/b/c`), und es steht nur mit sich selbst gleicher Länge zusammen. Die Mutationen „nur die ersten zwei Segmente verglichen“ und „mittlere Segmente übersprungen“ bleiben damit grün, obwohl `SPEC-238` Nr. 2 d „an jeder Stelle“ sagt; ein Paar `/a/b/c` gegen `/a/x/c` (gültig) oder `/a/{p}/d` (gültig) wäre mit der Mutation fälschlich überlappend oder gültig, je nach Richtung.
- `verifizierbar`: ja (die zwei Mutationen oben; `go test` über das Paket bleibt grün)
- `klasse`: Fänger der Allaussage „an jeder Stelle“ führt keine Muster mit mehr als zwei Segmenten gegeneinander

Der Implementer hat die Abweichung in §6 (neun statt acht Muster) wegen der Mutation „nur das letzte Segment“ eingeführt und die Mutation `/b/c` gedeckt: der Fix deckt das genannte Element, nicht die Klasse „jede Stelle bei jeder Segmentzahl“ (§3.17). Dasselbe Muster, das der Fix schon einmal fand, wiederholt sich eine Stelle weiter.

### 2. MEDIUM — Z4: die Werte der Zuordnung enthalten kein `+`; Dekodieren als Query-Wert bleibt grün

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-238` Nr. 2 (jedes Segment einzeln dekodiert; `EscapedPath`)
- `pfad`: `src/adapters/driving/admin/zuordnung_test.go` · `roh = []string{"x", ".", "..", "a%2Fb", "%C3%A4", "%FF", "%00"}`; Code `route.go` · `url.PathUnescape(t)`
- `befund`: Mit `url.QueryUnescape` statt `url.PathUnescape` bleibt das Paket grün; die einzige Form, in der beide verschieden dekodieren, ist `+` im Segment (`/a/x+y` ergäbe den Parameterwert `x y` und passte auf kein Literal `x+y`). Der Plan nennt die Werte als Menge „jedes nichtleere Segment“, die Prozentfolgen sind abgedeckt, das Pluszeichen nicht.
- `verifizierbar`: ja (die Mutation oben)
- `klasse`: Menge der Wertformen eines Parameters ohne die Form, in der zwei Dekodierer sich unterscheiden

### 3. MEDIUM — Z8: „jede andere Header-Zeile bleibt“ hat als Gegenprobe nur `X-A` und `X-Multi`

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-238` Nr. 2 (`Header` eine Kopie ohne `Authorization`)
- `pfad`: `src/adapters/driving/admin/anfrage_test.go` · `func z8Faelle() []z8Fall`; Code `admin.go` · `strings.EqualFold(k, "Authorization")`
- `befund`: Ein Filter, der jeden Header mit Suffix `authorization` oder Präfix `auth` entfernt, bleibt grün. `Proxy-Authorization` und ein Name mit Präfix `Auth` (etwa `Auth-Info`) gehören nach dem Wortlaut zu den Zeilen, die bleiben; kein Fall führt einen Namen, der `Authorization` enthält, ohne gleich zu sein.
- `verifizierbar`: ja (beide Mutationen oben)
- `klasse`: Gegenprobe einer Filter-Zusage ohne Nachbarname

Die Befunde 1 bis 3 sind Lücken von je einer Zeile im Fänger; der Code ist an diesen Stellen richtig (gelesen und gegen den Wortlaut der Spezifikation gehalten). Kein Failure-Szenario im Produkt heute, daher MEDIUM nach der Liste des Skills („Fänger führt nur eine Form der Eingabe“), nicht höher.

### 4. LOW — Plan: die Zahlen der Mutationsläufe passen nicht zusammen

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.9 (der Plan folgt jeder Korrektur), §3.10 (Zahlen)
- `pfad`: `docs/plan/planning/done/slice-admin-rahmen.md` · `108 Läufe von Mutationen, 106 rot; vier blieben grün`
- `befund`: 108 Läufe mit 106 rot lässt zwei grüne zu, der Satz nennt vier; die beiden gleichen Läufe derselben Mutation und die danach rot gesehenen sind nicht aus der Zählung herausgerechnet.
- `verifizierbar`: nein (Zahl im Text)
- `klasse`: Zählung im Bericht widerspricht ihrer Aufzählung

### 5. LOW — `treffer.gefund` wird gesetzt und nie gelesen

- `kategorie`: LOW
- `quelle`: Maintainability (§3.7: Zustand, den kein Leser braucht)
- `pfad`: `src/adapters/driving/admin/route.go` · `gefund bool`
- `befund`: `ServeHTTP` entscheidet nach `t.route != nil` und `len(t.allow)`; das Feld `gefund` und sein Kommentar („gefund falsch, allow leer“) sagen dasselbe zweimal, und der Fänger kann es nicht unterscheiden.
- `verifizierbar`: nein
- `klasse`: ungenutztes Zustandsfeld in einer Ergebnisstruktur

## Bewertung der Abweichungen, Negative und Adressen

- **Abweichungen 1 bis 3** (Z6 sieben Zusatzfälle, Z4 achter Wert, Z3 neun Muster): alle drei stärken den Fänger, stehen im selben Commit in der Tabelle, in §6 mit Ausgang *offen* und Adresse Architect (§3.13 erfüllt), und widersprechen `SPEC-238` nicht. Z6 und Z4 sind im Gewicht in Ordnung; Z3 deckt die Klasse nicht (Befund 1). Kein Anlass, den Plan zurückzuweisen.
- **Abweichung 4 / Fund Teil 1** (`TestWettlaufNebenlaeufig` in `tokendatei_test.go`): gehört nicht in diesen Slice. Ich habe die Ursache an `tokendatei.go` nachgelesen: `istFehlend` ruft `os.Lstat` nach einem `Stat` mit `ErrNotExist`, `statFehler` ruft es ein zweites Mal, und beide Stellen bilden einen zwischenzeitlich entstandenen Eintrag auf `grundKeineDatei` ab; das Fenster ist so, wie der Implementer es beschreibt. Die Adressierung stimmt (§6, Ausgang *offen*, Architect, danach Fix-Slice des Planners); der Implementer ändert den Code von Teil 1 nicht. Der Test lief in meinem Container beim Mutationslauf nicht rot (Paket grün, 11 s). Ich dupliziere den Fund nicht; für den Verifier: ein rotes `make gates` aus diesem Grund ist kein Fehler dieses Slice (INFO 2).
- **Negative N-1 bis N-7:** Grund und Folge stehen bei jedem. N-4 (Dekodieren scheitert nie): stimmig, denn `EscapedPath` liefert nur gültige Kodierungen; der Zweig ist der eine erlaubte Ausreißer. N-7: die Folge („eine kurzlebige Goroutine in `baueRouten` bleibt grün“) bestätigt meine Mutation; die Datei `admin.go` und `route.go` tragen keine `go`-Anweisung. N-2 (zählt und auditiert nicht): `K3` fängt `Audit` und `Telemetrie` nur in `New`; ein vorzeitiges `CountError` in `ServeHTTP` bliebe grün, im Plan so benannt.
- **Raten (§3.10):** Prozesszahl (1 je Mutation), Messender (Implementer), Last (55 bis 69 bei 4 CPUs) und der Hinweis „Größenordnung“ stehen; alle Fänger sind deterministisch, ein Nebenläufigkeitsanteil fehlt in diesem Teil. Genügt.
- **§3.11:** Der Kopf von `TestNewStartetKeineGoroutine` nennt seine Grenze („fängt allein TestNewRuftNichts, und nur solange sie läuft“). Der Kommentar von `z3Muster` („`/b/c` unterscheidet sich von `/a/c` allein im ersten Segment“) sagt zu, was der Test prüft, und nicht mehr.

## INFO

1. `Konfig` führt das Feld `Vergleich` aus `SPEC-238` Nr. 1 noch nicht, `New` klont `Token` schon; beides gehört nach Plan zu `slice-admin-authentisierung` (K4, U4, U5). Die Mutation „`Token` ohne Kopie“ bleibt grün, bis K4 ihren Fänger bekommt; die Lücke ist im Plan benannt (S1-Tabelle).
2. `TestWettlaufNebenlaeufig` liegt in der Kette von `make test` und `make gates`; solange der Fund aus Teil 1 offen ist, kann die DoD-Zeile „`make gates` grün“ dieses Slice zufällig rot werden. Das ist die Frage des Verifiers und liegt nicht bei diesem Diff.
3. `rufe` gibt `Antwort.Status` ohne Prüfung an `WriteHeader`; ein Handler mit `Antwort{}` (Status 0) lässt `net/http` panikieren. Der Vertrag (`SPEC-238` Nr. 6 e: Status 200 bis 299) und die Abbildung liegen bei `slice-admin-fehlerabbildung`; hier nur zur Kenntnis.
4. Die Zählung der Zweige unter dem Detektor (`go test -race -coverprofile`) habe ich nicht neu gemessen.

## Zusammenfassung

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 (Z3 Segmentzahl, Z4 Pluszeichen, Z8 Nachbarnamen) |
| LOW | 2 |
| INFO | 4 |

Geprüft, ohne Befund: `src/hexagon/model` (Code `ODR-E2102`, Status, Reihenfolge); `src/adapters/driven/audit`, `src/adapters/driven/telemetry`, `src/hexagon/services/mock`, `src/adapters/driving/httpodata` (nur Zählstellen, gegen `model.AlleCodes()`); `harness/katalog-ausnahmen.txt`; `src/bootstrap` (unverändert); `spec/` (Änderungen sind die des Architect vor dem Code, nicht Gegenstand).

Kein Konflikt-Pfad nötig: kein HIGH, kein Rollen-Widerspruch.
